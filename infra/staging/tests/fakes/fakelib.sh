# shellcheck shell=bash
# Shared state of the G072 offline fakes (tests/env-lease.sh, tests/ci-adapter.sh). Sourced by the fake tools; no
# network, no Docker, no cloud. State lives under FAKE_ROOT:
#   calls.log            one line per tool call: tool<TAB>class<TAB>argv (class R read, W write, A auth, T terraform,
#                        X refused/unknown); the fixtures assert on it (zero-write negatives, parity)
#   gcs/<bucket>/        a bucket (directory present = exists); obj/<key with / as %2F>/{gen,v<gen>}; noncurrent versions
#                        are kept; gcs/.counter is the global generation counter
#   perms.tsv            the IAM model (copied from infra/terraform/bootstrap/ci-meta-permissions.tsv or the fixture)
#   vm/                  the fake instance (status, max_run, term, natip, ip, candidate, db_versions, deploy_mode ...)
#   tf/                  the fake Terraform state (session, state_names, plan_rc, refresh_want, refresh_from)
#   ar/digests           registry references that exist
#   docker/images.tsv    tag<TAB>id of the fake local image store
set -u
# The credential-free build runs staging.sh under env -i with only PATH, HOME and DOCKER_*: DOCKER_FAKE_ROOT carries the
# state directory through it (fixture-only name).
FAKE_ROOT=${FAKE_ROOT:-${DOCKER_FAKE_ROOT:-}}
: "${FAKE_ROOT:?FAKE_ROOT is not set}"

fake_log() { local tool=$1 class=$2; shift 2; printf '%s\t%s\t%s\n' "$tool" "$class" "$*" >>"$FAKE_ROOT/calls.log"; }
fake_get() { [[ -f $FAKE_ROOT/$1 ]] && cat "$FAKE_ROOT/$1" || printf '%s' "${2:-}"; }
fake_set() { mkdir -p "$(dirname "$FAKE_ROOT/$1")"; printf '%s\n' "$2" >"$FAKE_ROOT/$1"; }
fake_now() { if [[ -f $FAKE_ROOT/now ]]; then cat "$FAKE_ROOT/now"; else printf '%s\n' "${FAKE_NOW:-1800000000}"; fi; }
fake_iso() { "$(fake_real date)" -u -d "@$1" +%Y-%m-%dT%H:%M:%SZ; }
# fake_real TOOL: the real tool, never one of the fakes.
fake_real() {
  local d
  for d in /usr/bin /bin /usr/local/bin; do [[ -x $d/$1 ]] && { printf '%s/%s' "$d" "$1"; return 0; }; done
  return 1
}
