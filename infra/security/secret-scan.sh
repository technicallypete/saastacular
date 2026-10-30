#!/usr/bin/env bash
# Submitted-source secret scan (G012 Plan r1 §4.2; README platform:check).
#
#   bash infra/security/secret-scan.sh --env-file FILE [--root DIR] secret-scan
#
# Scans exactly the submitted set of the Git work tree DIR (default: this repository): tracked files plus untracked,
# non-ignored files (git ls-files --cached --others --exclude-standard), never the ambient directory. The set is
# streamed into the pinned gitleaks container on stdin (no bind mount, no network) by the Compose overlay
# infra/compose/secret-scan.yml, which checks the regular-file and link counts before scanning.
#
# - The independent development submodule is excluded by path without inspection; other gitlinks are refused.
# - A symbolic link is kept as a link and never followed; its target text is scanned (scan/link-targets.txt).
# - An untracked nested repository (one directory entry) is skipped and counted, never descended.
# - Any other entry type, a missing path or a name with a tab or newline is refused.
# - A submitted path matching a local-only exemption of the candidate gitleaks.toml (the [[allowlists]] block
#   "local-only files that are never committed") is refused before scanning, so the exemption can only hide
#   untracked, ignored files.
#
# Exit: 0 clean; 1 finding; 2 scanner, stream, listing or accounting error (never a pass); 3 a submitted path matches
# a local-only exemption; 64 usage. Values are never printed (gitleaks --redact).
set -euo pipefail

usage() { echo "usage: secret-scan.sh --env-file FILE [--root DIR] secret-scan" >&2; exit 64; }
fail() { echo "secret-scan: $2" >&2; exit "$1"; }

REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)"
ENV_FILE="" ROOT="$REPO" SERVICE=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    --env-file) [[ $# -ge 2 ]] || usage; ENV_FILE="$2"; shift 2 ;;
    --root) [[ $# -ge 2 ]] || usage; ROOT="$2"; shift 2 ;;
    secret-scan) [[ -z "$SERVICE" ]] || usage; SERVICE="$1"; shift ;;
    *) usage ;;
  esac
done
[[ -n "$ENV_FILE" && "$SERVICE" == secret-scan ]] || usage
[[ -f "$ENV_FILE" ]] || fail 64 "--env-file: not a file"
[[ -d "$ROOT" ]] || fail 2 "--root: not a directory"
ROOT="$(cd "$ROOT" && pwd -P)"
top="$(git -C "$ROOT" rev-parse --show-toplevel 2>/dev/null)" || fail 2 "$ROOT is not a Git work tree"
[[ "$(cd "$top" && pwd -P)" == "$ROOT" ]] || fail 2 "$ROOT is not the top level of its Git work tree"
CONFIG="$ROOT/infra/security/gitleaks.toml"
[[ -f "$CONFIG" && ! -L "$CONFIG" ]] || fail 2 "infra/security/gitleaks.toml is missing"

TMP="$(mktemp -d "${TMPDIR:-/tmp}/secret-scan-XXXXXX")"
trap 'rm -rf -- "$TMP"' EXIT
mkdir -p "$TMP/w/scan/src" "$TMP/w/meta"

# 1. Gitlinks first (index mode 160000), so a submodule checkout is never mistaken for a skippable directory.
declare -A GITLINK=() SEEN=()
git -C "$ROOT" ls-files -z --stage >"$TMP/stage.z" || fail 2 "git ls-files --stage failed"
while IFS= read -r -d '' rec; do
  if [[ "${rec%% *}" == 160000 ]]; then GITLINK["${rec#*$'\t'}"]=1; fi
done <"$TMP/stage.z"
git -C "$ROOT" ls-files -z --cached --others --exclude-standard >"$TMP/list.z" || fail 2 "git ls-files failed"

: >"$TMP/files.z"; : >"$TMP/w/scan/link-targets.txt"
nreg=0 nlink=0 nskip=0 nsub=0
while IFS= read -r -d '' e; do
  p="${e%/}"
  if [[ -n "${SEEN[$p]+x}" ]]; then continue; fi
  SEEN["$p"]=1
  [[ "$p" != *$'\t'* && "$p" != *$'\n'* ]] || fail 2 "a submitted path contains a tab or newline"
  if [[ -n "${GITLINK[$p]+x}" ]]; then
    if [[ "$p" == .agents/skills/restruct-sdlc ]]; then
      nsub=$((nsub + 1)); echo "secret-scan: excluded independent development submodule: $p" >&2
      continue
    fi
    fail 2 "gitlink (submodule) entry is not supported: $p"
  elif [[ -L "$ROOT/$p" ]]; then
    nlink=$((nlink + 1)); printf '%s\0' "$p" >>"$TMP/files.z"
    printf '%s\t%s\n' "$p" "$(readlink -- "$ROOT/$p")" >>"$TMP/w/scan/link-targets.txt"
  elif [[ -f "$ROOT/$p" ]]; then
    nreg=$((nreg + 1)); printf '%s\0' "$p" >>"$TMP/files.z"
  elif [[ -d "$ROOT/$p" && "$e" == */ ]]; then
    nskip=$((nskip + 1)); echo "secret-scan: skipped untracked nested repository: $p" >&2
  else
    fail 2 "unsupported or missing submitted entry: $p"
  fi
done <"$TMP/list.z"

# 2. Exemption guard: the candidate's local-only path patterns must match no submitted path.
awk '
  /^\[\[allowlists\]\]/ { inblk = 0; inpaths = 0; next }
  /^description *= *"local-only files that are never committed"/ { inblk = 1; found = 1; next }
  inblk && /^paths *= *\[/ { inpaths = 1; next }
  inpaths && /^\]/ { inpaths = 0; inblk = 0; next }
  inpaths { print }
  END { if (!found) exit 1 }
' "$CONFIG" >"$TMP/exempt-lines.txt" || fail 2 "the local-only allowlist block is missing from gitleaks.toml"
# Each pattern line is a TOML literal string: '''REGEX''' with an optional trailing comma.
Q="'''"
sed -n "s/^[[:space:]]*$Q\(.*\)$Q,\{0,1\}[[:space:]]*\$/\1/p" "$TMP/exempt-lines.txt" >"$TMP/exempt.txt"
[[ -s "$TMP/exempt.txt" ]] || fail 2 "the local-only allowlist block has no path patterns"
# gitleaks reads these patterns as RE2, the guard as POSIX ERE: syntax only RE2 understands (flags, non-capturing groups,
# Perl classes, lazy quantifiers) is refused, so the guard can never silently evaluate a different pattern.
if grep -qE '\(\?|\\[dDsSwWbB]|[*+?}]\?' "$TMP/exempt.txt"; then
  fail 2 "a local-only exemption pattern uses RE2-only syntax the guard cannot evaluate"
fi
if [[ -s "$TMP/files.z" ]]; then
  grc=0
  grep -zE -f "$TMP/exempt.txt" "$TMP/files.z" >"$TMP/exempt-hits.z" 2>"$TMP/exempt-grep.err" || grc=$?
  case "$grc" in
    0) while IFS= read -r -d '' p; do echo "secret-scan: submitted path matches a local-only exemption: $p" >&2; done <"$TMP/exempt-hits.z"
       exit 3 ;;
    1) ;;
    *) fail 2 "the local-only exemption patterns cannot be evaluated (grep exit $grc)" ;;
  esac
fi

# 3. Stream the submitted set (links stored as links, never dereferenced) and its accounting.
(( nreg > 0 )) || fail 2 "no regular file in the submitted set"
tar -C "$ROOT" --null --no-recursion -T "$TMP/files.z" -cf - | tar -C "$TMP/w/scan/src" -xf - || fail 2 "cannot stage the submitted set"
printf '%s\n' "$nreg" >"$TMP/w/meta/count-regular"
printf '%s\n' "$nlink" >"$TMP/w/meta/count-links"
cp -- "$CONFIG" "$TMP/w/meta/gitleaks.toml"
echo "secret-scan: submitted set of $ROOT: $nreg regular files, $nlink links, $nskip skipped nested repositories; $nsub independent development submodules" >&2

rc=0
tar -C "$TMP/w" -cf - scan meta | docker compose --env-file "$ENV_FILE" -f "$REPO/infra/compose/local.yml" \
  -f "$REPO/infra/compose/secret-scan.yml" --profile checks run --rm -T secret-scan >"$TMP/scan.log" 2>&1 || rc=$?
cat -- "$TMP/scan.log"
# Exit 1 only when the overlay itself reported findings; any Compose, image, stream or other failure is exit 2.
case "$rc" in
  0) exit 0 ;;
  1) if grep -q '^secret-scan: findings (redacted)$' "$TMP/scan.log"; then exit 1; fi
     echo "secret-scan: exit 1 without the scanner's findings report (treated as an error)" >&2; exit 2 ;;
  *) echo "secret-scan: scanner, stream or accounting error (exit $rc)" >&2; exit 2 ;;
esac
