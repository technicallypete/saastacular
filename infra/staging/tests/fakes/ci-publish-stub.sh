# shellcheck shell=bash
# Stub of W3's infra/staging/ci-publish.sh (G072 shared contract 3, as W3 fixed it) for W2's offline fixtures until
# integration: publish_filter RAW OUT, canary_scan FILE..., summary_render OUT KEY=VALUE..., SUMMARY_KEYS. Bash
# functions only, no top-level side effects. The fixtures use the real ci-publish.sh once it is in the tree (or
# FX_PUBLISH_FILE).

PUBLISH_KINDS="step result session address reconcile receipt digest cert marker service secret"
SUMMARY_KEYS="action environment candidate selection branch actor pipeline job reuse digests session scheduled_stop readiness receipt_id receipt_sha256 fence_open exit_class"

# publish_filter RAW OUT: only allowlisted line kinds, then "dropped-lines: <n>".
publish_filter() {
  local raw=$1 out=$2 line n=0 kind
  : >"$out"
  while IFS= read -r line || [[ -n $line ]]; do
    kind=${line%%$'\t'*}
    if [[ $line == 'staging: '* ]] || [[ $line == *$'\t'* && " $PUBLISH_KINDS " == *" $kind "* ]]; then
      printf '%s\n' "$line" >>"$out"
    else
      n=$((n + 1))
    fi
  done <"$raw"
  printf 'dropped-lines: %s\n' "$n" >>"$out"
}

# canary_scan FILE...: 0 clean, 6 on any hit (file name and class only). Literals: STAGING_CI_ID_TOKEN_FILE and the lines
# of STAGING_CI_CANARY_LITERALS; a file below raw/ is not scanned for the plan canary.
canary_scan() {
  local f hit=0 cls re v
  for f in "$@"; do
    [[ -f $f ]] || { printf 'canary: %s unreadable\n' "$f"; hit=1; continue; }
    for v in "${STAGING_CI_ID_TOKEN_FILE:-}" "${STAGING_CI_CANARY_LITERALS:-}"; do
      [[ -n $v && -s $v ]] && grep -aqF -f "$v" -- "$f" && { printf 'canary: %s literal\n' "$f"; hit=1; }
    done
    for cls in access-token jwt private-key cloudflare-token plan-canary; do
      case $cls in
        access-token) re='ya29\.' ;;
        jwt) re='eyJ[A-Za-z0-9_-]{6,}\.[A-Za-z0-9_-]{6,}' ;;
        private-key) re='BEGIN [A-Z0-9 ]*PRIVATE KEY' ;;
        cloudflare-token) re='cf[au]t_[A-Za-z0-9]{16,}' ;;
        plan-canary) [[ "/$f" != */raw/* ]] || continue; re='G072PLANCANARY' ;;
      esac
      if grep -aqE -- "$re" "$f"; then printf 'canary: %s %s\n' "$f" "$cls"; hit=1; fi
    done
  done
  ((hit == 0)) || return 6
}

# summary_render OUT KEY=VALUE...: SUMMARY_KEYS in order, "-" for absent; an unknown key is rc 2.
summary_render() {
  local out=$1 kv k
  local -A val=()
  shift
  for kv in "$@"; do
    k=${kv%%=*}
    [[ " $SUMMARY_KEYS " == *" $k "* ]] || return 2
    val[$k]=${kv#*=}
  done
  : >"$out"
  for k in $SUMMARY_KEYS; do printf '%s: %s\n' "$k" "${val[$k]:--}" >>"$out"; done
}
