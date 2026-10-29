#!/usr/bin/env bash
# Controls W1-W9 for infra/validation/run-approved.sh (G002 Plan §15, B-WRAPPER-CONTROLS).
# Runs a copy of the wrapper against a stub harness in a scratch tree outside the repository,
# with input names containing spaces. Writes nothing inside the repository.
# Exit: 0 when every control behaves as expected; 1 otherwise.
set -euo pipefail
shopt -s inherit_errexit

REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)"
WRAPPER="$REPO/infra/validation/run-approved.sh"
T="$(mktemp -d "${TMPDIR:-/tmp}/run-approved-controls-XXXXXX")"
T="$(cd "$T" && pwd -P)"
case "$T/" in "$REPO"/*) echo "scratch tree must be outside the repository" >&2; exit 1 ;; esac
trap 'rm -rf -- "$T"' EXIT

failures=0
pass() { printf 'PASS %-34s %s\n' "$1" "$2"; }
bad()  { printf 'FAIL %-34s %s\n' "$1" "$2"; failures=$((failures + 1)); }

# A fresh scratch tree per case: wrapper copy, stub harness, three inputs, snapshot.
setup() {
  rm -rf -- "$T/tree"; mkdir -p "$T/tree/infra/validation" "$T/tree/in/sub dir"
  cp -- "$WRAPPER" "$T/tree/infra/validation/run-approved.sh"
  cat >"$T/tree/infra/validation/run.sh" <<'STUB'
#!/usr/bin/env bash
# Stub harness: records its invocation, optionally mutates an input or the snapshot, exits STUB_RC.
printf 'invoked %s\n' "$*" >>"$STUB_MARKER"
case "${STUB_MUTATE:-}" in
  input) printf 'drift\n' >>"$STUB_ROOT/in/sub dir/second input.txt" ;;
  snapshot) printf '\n' >>"$STUB_SNAPSHOT" ;;
esac
exit "${STUB_RC:-0}"
STUB
  chmod +x "$T/tree/infra/validation/run.sh" "$T/tree/infra/validation/run-approved.sh"
  printf 'plan\n' >"$T/tree/in/plan.md"
  printf 'first\n' >"$T/tree/in/first input.txt"
  printf 'second\n' >"$T/tree/in/sub dir/second input.txt"
  (cd "$T/tree" && sha256sum -- "in/plan.md" "in/first input.txt" "in/sub dir/second input.txt") >"$T/snap"
  : >"$T/marker"
}
digest() { sha256sum -- "$1" | cut -d' ' -f1; }

# wrap NAME SNAPSHOT DIGEST [STUB_RC] [STUB_MUTATE] -> sets rc and runs (stub invocation count)
wrap() {
  local snap="$2" sha="$3"
  rc=0
  STUB_MARKER="$T/marker" STUB_ROOT="$T/tree" STUB_SNAPSHOT="$snap" STUB_RC="${4:-0}" STUB_MUTATE="${5:-}" \
    bash "$T/tree/infra/validation/run-approved.sh" --snapshot "$snap" --snapshot-sha256 "$sha" --count 3 \
    --path-re '^in/.+$' --plan in/plan.md -- --stub-args >"$T/$1.out" 2>&1 || rc=$?
  runs="$(grep -c '^invoked ' "$T/marker" || true)"
}
expect() { # expect NAME RC RUNS [PATTERN-IN-OUTPUT]
  if [[ "$rc" == "$2" && "$runs" == "$3" ]] && { [[ -z "${4:-}" ]] || grep -q -- "$4" "$T/$1.out"; }; then
    pass "$1" "exit $rc, harness runs $runs"
  else bad "$1" "exit $rc (want $2), harness runs $runs (want $3): $(tail -2 "$T/$1.out" | tr '\n' ' ')"; fi
}

setup; wrap W1-all-valid "$T/snap" "$(digest "$T/snap")"; expect W1-all-valid 0 1 'postcheck passed'
setup; wrap W2-snapshot-digest "$T/snap" "$(printf '0%.0s' {1..64})"; expect W2-snapshot-digest 10 0
setup; printf 'changed\n' >"$T/tree/in/first input.txt"; wrap W3-input-mismatch "$T/snap" "$(digest "$T/snap")"; expect W3-input-mismatch 11 0
setup; sed -i '$d' "$T/snap"; wrap W4-count "$T/snap" "$(digest "$T/snap")"; expect W4-count 12 0

w5() { # w5 NAME SED-EXPRESSION
  setup; sed -i "$2" "$T/snap"; wrap "$1" "$T/snap" "$(digest "$T/snap")"; expect "$1" 12 0
}
setup; l1="$(sed -n 1p "$T/snap")"
w5 W5a-duplicate "2s#.*#${l1//#/\\#}#"
w5 W5b-absolute '1s#  in/#  /in/#'
w5 W5c-dotdot '1s#  in/#  in/../in/#'
w5 W5d-escaped '1s#^#\\#'
w5 W5e-outside-path-re '1s#  in/plan.md#  other/plan.md#'

setup; wrap W6-harness-fails-clean-post "$T/snap" "$(digest "$T/snap")" 1; expect W6-harness-fails-clean-post 1 1 'postcheck passed'
setup; wrap W7-post-drift-input "$T/snap" "$(digest "$T/snap")" 0 input; expect W7-post-drift-input 13 1 'an input changed'
setup; wrap W8-post-drift-snapshot "$T/snap" "$(digest "$T/snap")" 0 snapshot; expect W8-post-drift-snapshot 13 1 'snapshot digest changed'
setup; wrap W9-harness-fails-and-drift "$T/snap" "$(digest "$T/snap")" 1 input; expect W9-harness-fails-and-drift 1 1 'an input changed'

if [[ $failures -gt 0 ]]; then echo "run-approved controls: $failures FAILED"; exit 1; fi
echo "run-approved controls: all passed"
