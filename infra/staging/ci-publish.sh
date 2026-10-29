#!/usr/bin/env bash
# G072 publication of staging CI output (Plan r1 §10.7; docs/foundation/STAGING.md). Sourced by infra/staging/ci.sh
# (and the CI-LAUNCHER fixture); it defines functions and two constant lists only, with no top-level side effects.
# Every staging.sh subprocess of a CI job writes to STAGING_CI_PRIVATE/raw/<step>.log (0600, removed with the private job
# directory, never an artifact, never uploaded): saved Terraform plans, init/plan/apply output and `+ ` command echoes
# stay there. Only publish_filter output and the summary reach the job log, the summary artifact and the durable logs/.
#
#   publish_filter RAW_FILE OUT_FILE     keep the allowlisted line kinds, then append "dropped-lines: <n>"; rc 0
#                                        (2 on a usage or file error, OUT_FILE then removed)
#   canary_scan FILE...                  rc 0 clean, 6 on any hit or unreadable file; prints "canary: <file> <class>" per
#                                        hit, never the matched value (2 on a usage error)
#   summary_render OUT_FILE KEY=VALUE... the fixed-order "key: value" lines of SUMMARY_KEYS, "-" for an absent or empty
#                                        value; rc 0 (2 on an unknown, malformed or repeated key, OUT_FILE not written)
#
# Allowlisted kinds: "staging: " messages (lib.sh log/die) and the tab-separated G070 kinds below, exactly as staging.sh,
# lib.sh and remote/*.sh print them (a CI-LAUNCHER case derives their printf/say keys and pins this list against them);
# secret lines only as name and version number. A line with a character outside printable ASCII and tab, or longer than
# 2000 bytes, is dropped. Canary classes: the job's ID token and any listed literal (STAGING_CI_ID_TOKEN_FILE, and one
# literal per line of STAGING_CI_CANARY_LITERALS, e.g. a fetched Cloudflare token; values of 8 bytes or more), Google
# access tokens (ya29.), JWTs, PEM private keys, Cloudflare cfut_/cfat_ tokens and the plan canary G072PLANCANARY. A file
# below a raw/ directory holds plan text by design and is scanned for the credential classes only; the plan canary marks
# plan text that must never be published.

PUBLISH_KINDS="step result session address reconcile receipt digest cert marker service secret"
SUMMARY_KEYS="action environment candidate selection branch actor pipeline job reuse digests session scheduled_stop readiness receipt_id receipt_sha256 fence_open exit_class"

publish_filter() {
  [[ $# -eq 2 && -n "$1" && -n "$2" ]] || { echo "publish_filter: usage: publish_filter RAW_FILE OUT_FILE" >&2; return 2; }
  local raw="$1" out="$2"
  [[ -f "$raw" && -r "$raw" ]] || { echo "publish_filter: $raw is not a readable file" >&2; return 2; }
  if ! (umask 077 && : >"$out") 2>/dev/null; then echo "publish_filter: cannot write $out" >&2; return 2; fi
  # shellcheck disable=SC2016 # awk program
  if ! LC_ALL=C awk -v kinds="$PUBLISH_KINDS" -v out="$out" '
    BEGIN { n = split(kinds, k, " "); for (i = 1; i <= n; i++) ok[k[i]] = 1; dropped = 0 }
    {
      line = $0; sub(/\r$/, "", line); keep = 0; kind = ""
      if (length(line) <= 2000 && line !~ /[^\t -~]/) {
        if (index(line, "staging: ") == 1) keep = 1
        else {
          t = index(line, "\t")
          if (t > 1) { kind = substr(line, 1, t - 1); if (kind in ok) keep = 1 }
        }
        if (keep && kind == "secret" && line !~ /^secret\t[A-Z][A-Z0-9_]*\t([0-9]*|none)$/) keep = 0
      }
      if (keep) print line > out; else dropped++
    }
    END { printf "dropped-lines: %d\n", dropped > out }' "$raw"; then
    rm -f "$out"; echo "publish_filter: filtering $raw failed" >&2; return 2
  fi
  return 0
}

canary_scan() {
  [[ $# -ge 1 ]] || { echo "canary_scan: usage: canary_scan FILE..." >&2; return 2; }
  local f c re hit=0 lit="" v
  local -a classes=(
    "access-token" 'ya29\.'
    "jwt" 'eyJ[A-Za-z0-9_-]{6,}\.[A-Za-z0-9_-]{6,}'
    "private-key" 'BEGIN [A-Z0-9 ]*PRIVATE KEY'
    "cloudflare-token" 'cf[au]t_[A-Za-z0-9]{16,}'
    "plan-canary" 'G072PLANCANARY'
  )
  # The literal values go to a private pattern file (removed below), never to an argument or the output.
  if [[ -n "${STAGING_CI_ID_TOKEN_FILE:-}" || -n "${STAGING_CI_CANARY_LITERALS:-}" ]]; then
    lit="$(umask 077 && mktemp "${STAGING_CI_PRIVATE:-${TMPDIR:-/tmp}}/canary-literals.XXXXXX")" || { echo "canary: cannot create the literal pattern file" >&2; return 6; }
    if [[ -n "${STAGING_CI_ID_TOKEN_FILE:-}" ]]; then
      [[ -f "$STAGING_CI_ID_TOKEN_FILE" && -r "$STAGING_CI_ID_TOKEN_FILE" ]] || { rm -f "$lit"; echo "canary: the ID token file is unreadable (fail closed)" >&2; return 6; }
      v="$(head -c 65536 "$STAGING_CI_ID_TOKEN_FILE")"
      [[ ${#v} -lt 8 ]] || printf '%s\n' "$v" >>"$lit"
    fi
    if [[ -n "${STAGING_CI_CANARY_LITERALS:-}" && -f "$STAGING_CI_CANARY_LITERALS" ]]; then
      while IFS= read -r v || [[ -n "$v" ]]; do [[ ${#v} -lt 8 ]] || printf '%s\n' "$v" >>"$lit"; done <"$STAGING_CI_CANARY_LITERALS"
    fi
    v=""
  fi
  for f in "$@"; do
    if [[ ! -f "$f" || ! -r "$f" ]]; then printf 'canary: %s unreadable\n' "$f"; hit=1; continue; fi
    if [[ -n "$lit" && -s "$lit" ]] && LC_ALL=C grep -aqF -f "$lit" -- "$f"; then printf 'canary: %s literal\n' "$f"; hit=1; fi
    for ((c = 0; c < ${#classes[@]}; c += 2)); do
      re="${classes[c + 1]}"
      if [[ "${classes[c]}" == plan-canary && "/$f" == */raw/* ]]; then continue; fi
      if LC_ALL=C grep -aqE -- "$re" "$f"; then printf 'canary: %s %s\n' "$f" "${classes[c]}"; hit=1; fi
    done
  done
  [[ -z "$lit" ]] || rm -f "$lit"
  if ((hit)); then return 6; fi
  return 0
}

summary_render() {
  [[ $# -ge 1 && -n "$1" ]] || { echo "summary_render: usage: summary_render OUT_FILE KEY=VALUE..." >&2; return 2; }
  local out="$1" kv k v; shift
  local -A val=()
  for kv in "$@"; do
    [[ "$kv" == *=* ]] || { echo "summary_render: malformed argument (KEY=VALUE expected)" >&2; return 2; }
    k="${kv%%=*}" v="${kv#*=}"
    [[ " $SUMMARY_KEYS " == *" $k "* ]] || { echo "summary_render: unknown key $k" >&2; return 2; }
    [[ -z "${val[$k]+x}" ]] || { echo "summary_render: repeated key $k" >&2; return 2; }
    # One line per key: control characters become spaces, non-ASCII is dropped, at most 1000 bytes.
    v="$(printf '%s' "$v" | LC_ALL=C tr '\000-\037\177' ' ' | LC_ALL=C tr -cd ' -~')"
    if ((${#v} > 1000)); then v="${v:0:997}..."; fi
    val[$k]="$v"
  done
  {
    for k in $SUMMARY_KEYS; do
      v="${val[$k]:-}"
      printf '%s: %s\n' "$k" "${v:--}"
    done
  } >"$out" || { echo "summary_render: cannot write $out" >&2; return 2; }
  return 0
}
