#!/usr/bin/env bash
# Runs the project validation harness bound to a supervisor-approved input snapshot,
# failing closed (G002 Plan §17).
#
#   run-approved.sh --snapshot FILE --snapshot-sha256 HEX64 --count N --path-re ERE --plan PATH
#                   [--check-only] -- <run.sh arguments...>
#
# The snapshot (sha256sum format) is read once; the SHA-256 of those bytes must equal the pinned
# digest, and only those bytes are parsed. Every entry is verified before the harness runs, the
# harness status is captured explicitly, and a postcheck always runs afterwards.
#
# Exit: 0 everything passed; 2 usage; 10 snapshot digest mismatch; 11 an input differs from its
#       entry; 12 malformed snapshot (format, path, duplicate, count, Plan missing); 13 postcheck
#       drift after a passing harness; otherwise the harness's own nonzero status, which a passing
#       postcheck can never replace.
set -euo pipefail
shopt -s inherit_errexit

REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)"

usage() { echo "usage: run-approved.sh --snapshot FILE --snapshot-sha256 HEX64 --count N --path-re ERE --plan PATH [--check-only] -- <run.sh arguments...>" >&2; exit 2; }
stop() { printf 'run-approved: FAIL %s: %s\n' "$1" "$2" >&2; exit "$1"; }

SNAP="" SNAP_SHA="" COUNT="" PATH_RE="" PLAN="" CHECK_ONLY=0 passthrough=()
while [[ $# -gt 0 ]]; do
  case "$1" in
    --snapshot) SNAP="${2:-}"; shift 2 || usage ;;
    --snapshot-sha256) SNAP_SHA="${2:-}"; shift 2 || usage ;;
    --count) COUNT="${2:-}"; shift 2 || usage ;;
    --path-re) PATH_RE="${2:-}"; shift 2 || usage ;;
    --plan) PLAN="${2:-}"; shift 2 || usage ;;
    --check-only) CHECK_ONLY=1; shift ;;
    --) shift; passthrough=("$@"); break ;;
    *) usage ;;
  esac
done
[[ -n "$SNAP" && "$SNAP_SHA" =~ ^[0-9a-f]{64}$ && "$COUNT" =~ ^[1-9][0-9]*$ && -n "$PATH_RE" && -n "$PLAN" ]] || usage
[[ $CHECK_ONLY == 1 || ${#passthrough[@]} -gt 0 ]] || usage

digest_of_file() { local b; b="$(cat -- "$1"; printf x)" || return 1; printf '%s' "${b%x}" | sha256sum | cut -d' ' -f1; }

SNAP="$(realpath -e -- "$SNAP" 2>/dev/null)" || stop 10 "snapshot not found"

# 1. Read once; verify the digest of exactly those bytes.
bytes="$(cat -- "$SNAP"; printf x)" || stop 10 "cannot read the snapshot"
bytes="${bytes%x}"
[[ "$(printf '%s' "$bytes" | sha256sum | cut -d' ' -f1)" == "$SNAP_SHA" ]] || stop 10 "snapshot digest mismatch"

# 2. Parse only those bytes.
LINE_RE='^([0-9a-f]{64})  ([^\].*)$'
[[ -n "$bytes" && "$bytes" == *$'\n' ]] || stop 12 "snapshot is empty or lacks a final newline"
declare -A want=()
paths=()
while IFS= read -r line; do
  [[ "$line" =~ $LINE_RE ]] || stop 12 "malformed entry: $line"
  sha="${BASH_REMATCH[1]}" path="${BASH_REMATCH[2]}"
  [[ "$path" != /* && "/$path/" != */./* && "/$path/" != */../* ]] || stop 12 "non-relative path: $path"
  [[ "$path" =~ $PATH_RE ]] || stop 12 "path outside the approved set: $path"
  [[ -z "${want[$path]+x}" ]] || stop 12 "duplicate path: $path"
  want[$path]="$sha"; paths+=("$path")
done <<<"${bytes%$'\n'}"
[[ ${#paths[@]} -eq "$COUNT" ]] || stop 12 "expected $COUNT entries, found ${#paths[@]}"
[[ -n "${want[$PLAN]+x}" ]] || stop 12 "the Plan is not an entry: $PLAN"

cd -- "$REPO"

# verify_entries: 0 when every entry matches; prints each mismatch.
verify_entries() {
  local p bad=0
  for p in "${paths[@]}"; do
    if [[ ! -f "$p" || -L "$p" ]]; then printf 'run-approved: missing or non-regular input: %s\n' "$p" >&2; bad=1
    elif [[ "$(sha256sum -- "$p" | cut -d' ' -f1)" != "${want[$p]}" ]]; then printf 'run-approved: hash mismatch: %s\n' "$p" >&2; bad=1; fi
  done
  return "$bad"
}

# 3. Precheck: every input matches before anything runs.
verify_entries || stop 11 "an input differs from the approved snapshot"
printf 'run-approved: precheck passed (%d inputs, snapshot %s)\n' "${#paths[@]}" "$SNAP_SHA"
[[ $CHECK_ONLY == 0 ]] || exit 0

# 4. Arguments from the parsed entries only.
args=()
for p in "${paths[@]}"; do
  args+=(--expect-hash "$p=${want[$p]}")
  [[ "$p" == "$PLAN" ]] || args+=(--exception "$p")
done

# 5. Harness, status captured explicitly.
set +e
"$REPO/infra/validation/run.sh" "${passthrough[@]}" "${args[@]}"
harness_rc=$?
set -e

# 6. Postcheck, always.
post_rc=0
if [[ "$(digest_of_file "$SNAP" 2>/dev/null || true)" != "$SNAP_SHA" ]]; then
  echo "run-approved: postcheck: snapshot digest changed during the run" >&2; post_rc=13
fi
verify_entries || { echo "run-approved: postcheck: an input changed during the run" >&2; post_rc=13; }
printf 'run-approved: harness exit %d; postcheck %s\n' "$harness_rc" "$([[ $post_rc == 0 ]] && echo passed || echo FAILED)"

# 7. Exit: a failing harness always wins; otherwise drift fails the run.
if [[ $harness_rc -ne 0 ]]; then exit "$harness_rc"; fi
exit "$post_rc"
