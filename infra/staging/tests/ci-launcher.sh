#!/usr/bin/env bash
# CI-LAUNCHER (G072 Plan r1 §10.7, §10.10, §10.13 AC1/AC3; the infra/staging/check.sh row of the same name): offline
# cases for the staging CI launcher (infra/ci/staging-action.sh) and the publication functions
# (infra/staging/ci-publish.sh). No Docker, network, cloud, GitLab or real credential. Each launcher case runs the real
# launcher under sh in a scratch GitLab job layout (a build-directory parent holding a local Git clone as the project
# and the file-variable directory beside it), with a fake `docker` first on PATH whose `run` executes a fixture adapter
# in place of `bash infra/staging/ci.sh ACTION`, with exactly the forwarded -e environment. The ID and job tokens are
# generated per run and never printed or written by the harness; their absence from every file of a case is asserted.
#
#   bash infra/staging/tests/ci-launcher.sh --work DIR --report DIR [--only NAME[,NAME]...]
#   bash infra/staging/tests/ci-launcher.sh --list
#
# Cases: g-* the approved-context gate (16, no Docker call, no private directory, no token file); i-* input refusals
# (2, before the token file; selection only by arguments); p-* approved runs (the -e allow-list, the docker client's
# environment, the token file's mode and location, the profile per action, cleanup); c-* failures, cancellation and the
# after-script; pub-* publication (the allowlist pinned against the G070 printf/say keys, a fake `terraform show` plan
# carrying G072PLANCANARY, the canary classes, the summary); s-* static no-fallback; m-* mutation controls: the named
# case, run against a copy of the launcher or ci-publish.sh with one protected behaviour removed, must FAIL.
# Report: REPORT/cases.tsv (case, result, detail) and REPORT/summary.tsv (selected, passed, failed). Exit 0 when every
# selected case passes, 1 on any failure, 64 usage, 65 refused (an unsafe path).
set -uo pipefail

REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd -P)"
LAUNCHER="$REPO/infra/ci/staging-action.sh"
PUBLISH="$REPO/infra/staging/ci-publish.sh"
SUBJ_L="$LAUNCHER" SUBJ_P="$PUBLISH"

usage() { echo "usage: ci-launcher.sh --work DIR --report DIR [--only NAME[,NAME]...] | --list" >&2; exit 64; }
die() { echo "ci-launcher.sh: $*" >&2; exit 65; }

CASES=(g-not-ci g-usage g-mr-iid g-unprotected g-other-branch g-source-push g-source-schedule g-source-trigger
  g-source-mr g-source-parent g-wrong-job g-env-mismatch g-env-prefix g-project-id g-project-path g-pin-source g-tag
  g-no-token g-profile-missing g-prepare-token g-prepare-environment g-build-profile
  i-stop-candidate i-stop-duration i-env-none i-env-unregistered i-env-case i-candidate i-duration i-reuse
  i-no-env-fallback i-probes
  p-deploy p-check p-stop p-runner-url c-unsafe-url p-prepare p-build p-exit-class p-no-summary
  c-check-dotenv c-token-leak c-jobtoken-file c-mount c-cleanup-fail c-after-script c-signal
  pub-plan pub-inline pub-token pub-kinds pub-summary pub-canary s-no-fallback s-sanitizer-sync)
# Mutation controls: NAME|TARGET CASE|SUBJECT (L launcher, P ci-publish.sh)|sed program (separated by ;; for several).
MUTATIONS=(
  'm-sanitizer-sync|s-sanitizer-sync|L|s/remotes = 0;/remotes = 1;/'
  'm-url-sanitize|p-runner-url|L|/sanitize_git_config [|][|]/d'
  'm-gate-mr|g-mr-iid|L|/WHY="merge request"/d'
  'm-gate-protected|g-unprotected|L|/WHY="protected ref"/d'
  'm-gate-branch|g-other-branch|L|/WHY="branch"/d'
  'm-gate-source|g-source-push|L|/WHY="pipeline source"/d'
  'm-gate-job|g-wrong-job|L|/WHY="job name"/d'
  'm-gate-environment|g-env-mismatch|L|/WHY="environment";/d'
  'm-gate-project|g-project-id|L|/WHY="project";/d'
  'm-gate-credential-free|g-prepare-token|L|/WHY="ID token in a credential-free job"/d'
  'm-stop-inputs|i-stop-candidate|L|/stop takes no candidate and no duration/d'
  'm-registry|i-env-unregistered|L|s/grep -qxF -- "\$ENV_SEL" "\$REG" ||/true ||/'
  'm-env-fallback|i-no-env-fallback|L|s/^ENV_SEL="" CAND=""/ENV_SEL="${STAGING_CI_ENV:-}" CAND=""/;;s/^\[ "\$S_ENV" = 1 \] || BAD=1$/:/'
  'm-lease-unset|p-deploy|L|/^for v in \$(env | sed/d;;s/ STAGING_LEASE_HELD//'
  'm-token-mode|p-deploy|L|s/(umask 077 && printf/(umask 022 \&\& printf/;;/chmod 600 "\$P\/id-token"/d;;/the token file is not a mode-600 file/d'
  'm-check-profile|p-check|L|s/if \[ "\$ACTION" != check \]; then$/if true; then/'
  'm-allow-list|p-deploy|L|s/-e GITLAB_CI -e CI_PIPELINE_SOURCE/-e GITLAB_CI -e CI_JOB_TOKEN -e CI_PIPELINE_SOURCE/'
  'm-tmpfs|p-deploy|L|s/ --mount "type=tmpfs,destination=\$CI_PROJECT_DIR.tmp,tmpfs-size=1m"//'
  'm-cleanup|p-deploy|L|s/elif \[ -e "\$P" \]; then rm -rf -- "\$P" 2>\/dev\/null; fi/fi/'
  'm-cleanup-verify|c-cleanup-fail|L|/cleanup: the private directory could not be removed"; return 1; fi/d'
  'm-mount|c-mount|L|s/^  say "refused: the private directory is not on the build directory.s filesystem"; exit 16$/  :/'
  'm-jobtoken-scan|c-jobtoken-file|L|s/if scan_tree "\$NS" "\$PF"; then :; else/if true; then :; else/'
  'm-leak-scan|c-token-leak|L|s/^if \[ "\$CRED" = 1 \] && \[ -s "\$P\/id-token" \]; then$/if false; then/'
  'm-filter-allow-all|pub-kinds|P|s/keep = 0; kind = ""/keep = 1; kind = ""/'
  'm-secret-values|pub-kinds|P|/if (keep && kind == "secret"/d'
  'm-scan-literal|pub-token|P|s/if \[\[ -n "\$lit" && -s "\$lit" \]\] && LC_ALL=C grep -aqF/if false \&\& LC_ALL=C grep -aqF/'
  'm-scan-plan-canary|pub-inline|P|/"plan-canary" .G072PLANCANARY./d'
  'm-scan-access-token|pub-canary|P|/"access-token" .ya29/d'
  'm-summary-unknown|pub-summary|P|/unknown key \$k/d'
)
for m in "${MUTATIONS[@]}"; do CASES+=("${m%%|*}"); done

WORK="" REPORT="" ONLY="" LIST=0
while [[ $# -gt 0 ]]; do
  case "$1" in
    --work) [[ $# -ge 2 ]] || usage; WORK="$2"; shift 2 ;;
    --report) [[ $# -ge 2 ]] || usage; REPORT="$2"; shift 2 ;;
    --only) [[ $# -ge 2 ]] || usage; ONLY="$2"; shift 2 ;;
    --list) LIST=1; shift ;;
    *) usage ;;
  esac
done
if ((LIST)); then printf '%s\n' "${CASES[@]}"; exit 0; fi
[[ -n "$WORK" && -n "$REPORT" ]] || usage
if [[ -n "$ONLY" ]]; then
  IFS=, read -ra sel <<<"$ONLY"
  for r in "${sel[@]}"; do [[ " ${CASES[*]} " == *" $r "* ]] || { echo "--only: unknown case $r" >&2; exit 64; }; done
fi
# A path inside this checkout must be Git-ignored; the launcher accepts only [A-Za-z0-9._/-] in the build path.
safe_dir() {
  mkdir -p "$1"; local p; p="$(cd "$1" && pwd -P)"
  if [[ "$p/" == "$REPO/"* ]] && ! git -C "$REPO" check-ignore -q "${p#"$REPO"/}/"; then die "$1 is inside the checkout and not ignored"; fi
  [[ "$p" =~ ^[A-Za-z0-9._/-]+$ ]] || die "$1 resolves to a path with characters the launcher refuses"
  printf '%s\n' "$p"
}
WORK="$(safe_dir "$WORK")" || exit 65
REPORT="$(safe_dir "$REPORT")" || exit 65
[[ ! -e "$REPORT/cases.tsv" ]] || die "$REPORT/cases.tsv exists; use a new report directory per run"
random_hex() { head -c "$1" /dev/urandom | od -An -tx1 | tr -d ' \n'; }
random_alnum() { head -c 200 /dev/urandom | base64 | tr -dc 'A-Za-z0-9' | head -c "$1"; }
RUN_ID="$(date -u +%Y%m%dt%H%M%Sz)-$(random_hex 4)"
S="$WORK/ci-launcher-$RUN_ID"; (umask 077; mkdir -p "$S")
trap 'chmod -R u+rwx "$S" 2>/dev/null; rm -rf "$S"' EXIT
TOKEN="fixture-id-token-$(random_hex 16)" JOBTOK="fixture-job-token-$(random_hex 16)"
PIN_ID=4242 PIN_PATH=fixture/group/project JOB_ID=4711
PIN_SHA="$(printf '%s' "$PIN_PATH" | sha256sum | cut -d' ' -f1)"
REGISTRY_SRC="$REPO/infra/staging/ci-environments.txt" REGISTRY_NOTE="registry: infra/staging/ci-environments.txt"
if [[ ! -f "$REGISTRY_SRC" ]]; then REGISTRY_SRC="$S/fixture-registry.txt" REGISTRY_NOTE="registry: fixture (staging, preview-1; the committed file is absent)"; printf 'staging\npreview-1\n' >"$REGISTRY_SRC"; fi

# ---- fixtures ------------------------------------------------------------------------------------------------------------
mkdir -p "$S/bin" "$S/tools" "$S/template"
cat >"$S/bin/docker" <<'EOF'
#!/usr/bin/env bash
# CI-LAUNCHER fake docker: records argv, runs the fixture adapter for `run` with exactly the forwarded environment.
d="$FAKE_DIR"
printf '%s\n' "$*" >>"$d/docker-calls"
case "${1:-}" in
  version) exit 0 ;;
  build) head -n 1 >"$d/build-input"; cat >/dev/null; exit 0 ;;
  image) exit 1 ;;
  container) if [[ -s "$d/ctx.pid" ]] && kill -0 "$(cat "$d/ctx.pid")" 2>/dev/null; then exit 0; fi; exit 1 ;;
  stop) if [[ -s "$d/ctx.pid" ]]; then kill -TERM "$(cat "$d/ctx.pid")" 2>/dev/null; fi; exit 0 ;;
  run) shift ;;
  *) exit 0 ;;
esac
env | sed 's/=.*//' | LC_ALL=C sort >"$d/docker-client-env-names"
envs=() wd=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    --rm|--init) shift ;;
    --name|--label|--network|--security-opt|-v|--mount) shift 2 ;;
    -w) wd="$2"; shift 2 ;;
    -e) if [[ "$2" == *=* ]]; then envs+=("$2"); elif [[ -n "${!2+x}" ]]; then envs+=("$2=${!2}"); fi; shift 2 ;;
    -*) echo "fake docker: unexpected option $1" >&2; exit 125 ;;
    *) break ;;
  esac
done
shift
printf '%s\n' "$@" >"$d/ctx-command"
printf '%s\n' "${envs[@]}" | sed 's/=.*//' | LC_ALL=C sort >"$d/ctx-env-names"
printf '%s\n' "${envs[@]}" >"$d/ctx-env"
cd "$wd" || exit 125
echo $$ >"$d/ctx.pid"
exec env -i "${envs[@]}" PATH="$FAKE_TOOLS:/usr/bin:/bin" FAKE_DIR="$d" FAKE_MODE="${FAKE_MODE:-ok}" bash "$FAKE_ADAPTER" "${@:3}"
EOF
cat >"$S/tools/terraform" <<'EOF'
#!/usr/bin/env bash
# CI-LAUNCHER fake terraform: `show -no-color PLAN` prints a session-change plan whose attributes carry the plan canary.
[[ "${1:-}" == show ]] || exit 0
printf '%s\n' 'Terraform will perform the following actions:' \
  '  # module.vm.google_compute_instance.vm will be updated in-place' \
  '  ~ resource "google_compute_instance" "vm" {' \
  '      ~ scheduling { max_run_duration { seconds = 7200 -> 600 } }' \
  '        metadata   = { "startup-marker" = "G072PLANCANARY" }' \
  '    }' 'Plan: 0 to add, 1 to change, 0 to destroy.'
EOF
cat >"$S/adapter.sh" <<'EOF'
#!/usr/bin/env bash
# CI-LAUNCHER fixture adapter (stands in for infra/staging/ci.sh): records what it received (never a value of the token),
# then follows the publication contract of Plan §10.7 for deploy/stop: raw/ private, filter, scan, publish.
set -u
a="${1:-}" d="$FAKE_DIR"
env | sed 's/=.*//' | LC_ALL=C sort >"$d/adapter-env-names"
if [[ -n "${STAGING_CI_ID_TOKEN_FILE:-}" ]]; then
  printf '%s %s %s\n' "$(stat -c %a "$STAGING_CI_ID_TOKEN_FILE")" "$(sha256sum <"$STAGING_CI_ID_TOKEN_FILE" | cut -d' ' -f1)" \
    "$STAGING_CI_ID_TOKEN_FILE" >"$d/token-facts"
fi
if [[ -n "${STAGING_CI_PROFILE_DIR:-}" ]]; then (cd "$STAGING_CI_PROFILE_DIR" && LC_ALL=C ls -A) | tr '\n' ' ' >"$d/profile-files"; fi
: >"$d/adapter-started"
case "$FAKE_MODE" in
  exit3) exit 3 ;;
  sleep) exec sleep 30 ;;
  leak) cat "$STAGING_CI_ID_TOKEN_FILE" >"$STAGING_CI_OUT/oops.txt"; exit 0 ;;
  check-bad) printf 'STAGING_REUSE=yes\nDOCKER_HOST=tcp://elsewhere:2376\n' >"$STAGING_CI_OUT/reuse.env"; exit 0 ;;
  no-summary) exit 0 ;;
esac
case "$a" in
  check) printf 'STAGING_REUSE=no\n' >"$STAGING_CI_OUT/reuse.env"; exit 0 ;;
  deploy|stop) ;;
  *) exit 0 ;;
esac
# shellcheck source=/dev/null
source infra/staging/ci-publish.sh
raw="$STAGING_CI_PRIVATE/raw"; mkdir -m 700 "$raw"
{
  printf 'staging: VM TERMINATED; session length 600s differs from 7200s: refresh-only plan\n'
  printf '+ terraform plan -refresh-only -out=/private/plans/session.tfplan\n'
  terraform show -no-color /private/plans/session.tfplan
  if [[ "$FAKE_MODE" == inline ]]; then printf 'staging: plan summary G072PLANCANARY\n'; fi
  if [[ "$FAKE_MODE" == token ]]; then printf 'debug: authorization bearer %s\n' "$(cat "$STAGING_CI_ID_TOKEN_FILE")"; fi
  printf 'session\t600s\tscheduled stop 2026-10-28T12:00:00Z\n'
  printf 'address\t203.0.113.10\n'
  printf 'step\t2026-10-28T10:00:00Z\tready\tall services healthy\n'
  printf 'result\tREADY\t%s\n' "${CI_COMMIT_SHA:-}"
  printf 'secret\tIDENTITY_WORKOS_API_KEY\t3\n'
  printf 'secret\tCF_API_TOKEN\tnot-a-version-G072PLANCANARY\n'
  printf 'vm\tstaging-vm\tRUNNING\n'
  printf 'staging-vm: VM-side message\n'
} >"$raw/$a.log" 2>&1
publish_filter "$raw/$a.log" "$STAGING_CI_PRIVATE/$a.pub" || exit 4
if ! canary_scan "$raw/$a.log" "$STAGING_CI_PRIVATE/$a.pub" >"$STAGING_CI_PRIVATE/canary.txt"; then cat "$STAGING_CI_PRIVATE/canary.txt"; exit 6; fi
mkdir -p "$d/gcs/logs"; cp "$STAGING_CI_PRIVATE/$a.pub" "$d/gcs/logs/$CI_PIPELINE_ID-$CI_JOB_ID.log"
cat "$STAGING_CI_PRIVATE/$a.pub"
summary_render "$STAGING_CI_OUT/summary.txt" action="$a" environment="$STAGING_CI_ENV" candidate="${CI_COMMIT_SHA:-}" \
  selection="${STAGING_CI_CANDIDATE_INPUT:-}" actor="${GITLAB_USER_LOGIN:-}" pipeline="$CI_PIPELINE_ID" job="$CI_JOB_ID" \
  session=600s readiness=READY exit_class=0
EOF
chmod 755 "$S/bin/docker" "$S/tools/terraform" "$S/adapter.sh"

# The template job layout: a bare origin (staging, feature/x and a merge-request ref) and a single-branch clone.
T="$S/template"
git init -q --bare "$T/origin.git"
git init -q -b staging "$T/seed"
printf 'seed\n' >"$T/seed/README"
git -C "$T/seed" add README
git -C "$T/seed" -c user.name=fixture -c user.email=fixture@example.invalid commit -q -m seed
git -C "$T/seed" branch feature/x
git -C "$T/seed" push -q "$T/origin.git" staging feature/x staging:refs/merge-requests/1/head
mkdir -p "$T/ns"
git clone -q --single-branch -b staging "$T/origin.git" "$T/ns/project"
SHA="$(git -C "$T/ns/project" rev-parse HEAD)"
mkdir -p "$T/ns/project.tmp" "$T/home"
for v in STAGING_CI_CONF STAGING_CI_BACKEND STAGING_CI_TFVARS STAGING_CI_SETTINGS; do printf 'FIXTURE_%s=1\n' "$v" >"$T/ns/project.tmp/$v"; done

setup_case() { # NAME: a fresh copy of the template with the current subject files
  C="$S/cases/$1"; mkdir -p "$S/cases"; cp -a "$T" "$C"
  local p="$C/ns/project"
  mkdir -p "$p/infra/ci" "$p/infra/staging" "$p/infra/docker" "$C/bin"
  cp "$SUBJ_L" "$p/infra/ci/staging-action.sh"
  cp "$SUBJ_P" "$p/infra/staging/ci-publish.sh"
  cp "$REGISTRY_SRC" "$p/infra/staging/ci-environments.txt"
  cp "$REPO/infra/docker/ci-context.Dockerfile" "$REPO/infra/docker/staging-ci.Dockerfile" "$p/infra/docker/"
  sed -e "s/^STG_PROJECT_ID=.*/STG_PROJECT_ID=$PIN_ID/" -e "s/^STG_PROJECT_PATH_SHA=.*/STG_PROJECT_PATH_SHA=$PIN_SHA/" \
    "$REPO/infra/ci/hosted-context.sh" >"$p/infra/ci/hosted-context.sh"
  PROB=""
  [[ "$(grep -c -e "^STG_PROJECT_ID=$PIN_ID\$" -e "^STG_PROJECT_PATH_SHA=$PIN_SHA\$" "$p/infra/ci/hosted-context.sh")" == 2 ]] ||
    prob "the hosted-context.sh pins were not substituted (fixture)"
}
prob() { PROB+="$*; "; }
base_env() { # ACTION [ENV]: the approved job environment of staging:ACTION
  local a="$1" e="${2:-staging}" t="$C/ns/project.tmp"
  ENVV=(PATH="$C/bin:$S/bin:/usr/bin:/bin" HOME="$C/home" LC_ALL=C GITLAB_CI=true CI_JOB_ID="$JOB_ID"
    CI_PROJECT_DIR="$C/ns/project" DOCKER_HOST=tcp://docker:2376 DOCKER_TLS_VERIFY=1 DOCKER_CERT_PATH=/certs/client
    DOCKER_TLS_CERTDIR=/certs CI_PROJECT_ID="$PIN_ID" CI_PROJECT_PATH="$PIN_PATH" CI_COMMIT_BRANCH=staging
    CI_COMMIT_REF_NAME=staging CI_COMMIT_REF_PROTECTED=true CI_PIPELINE_SOURCE=web CI_JOB_NAME="staging:$a"
    CI_PIPELINE_ID=900 CI_COMMIT_SHA="$SHA" GITLAB_USER_LOGIN=fixture-user CI_JOB_TOKEN="$JOBTOK"
    STAGING_LEASE_HELD="ci:1/2@5" FAKE_DIR="$C" FAKE_ADAPTER="$S/adapter.sh" FAKE_TOOLS="$S/tools" FAKE_MODE=ok)
  case "$a" in
    check|deploy|stop) ENVV+=(CI_ENVIRONMENT_NAME="$e" STAGING_CI_ID_TOKEN="$TOKEN" STAGING_CI_CONF="$t/STAGING_CI_CONF"
      STAGING_CI_BACKEND="$t/STAGING_CI_BACKEND" STAGING_CI_TFVARS="$t/STAGING_CI_TFVARS" STAGING_CI_SETTINGS="$t/STAGING_CI_SETTINGS") ;;
  esac
  case "$a" in build|deploy) ENVV+=(STAGING_REUSE=no) ;; esac
}
setenv() { local kv n i; for kv in "$@"; do n="${kv%%=*}"; for i in "${!ENVV[@]}"; do [[ "${ENVV[i]%%=*}" == "$n" ]] && unset 'ENVV[i]'; done; ENVV+=("$kv"); done; ENVV=("${ENVV[@]}"); }
unsetenv() { local n i; for n in "$@"; do for i in "${!ENVV[@]}"; do [[ "${ENVV[i]%%=*}" == "$n" ]] && unset 'ENVV[i]'; done; done; ENVV=("${ENVV[@]}"); }
launch() { RC=0; env -i "${ENVV[@]}" sh "$C/ns/project/infra/ci/staging-action.sh" "$@" >"$C/joblog" 2>&1 || RC=$?; }
P_DIR() { printf '%s\n' "$C/ns/.staging-ci-$JOB_ID"; }
expect_rc() { [[ "$RC" == "$1" ]] || prob "exit $RC, expected $1 ($(grep -m1 'refused\|usage' "$C/joblog" | cut -c1-160))"; }
no_private() { [[ ! -e "$(P_DIR)" && ! -L "$(P_DIR)" ]] || prob "the private directory remains"; }
no_docker() { [[ ! -e "$C/docker-calls" ]] || prob "a Docker call was made"; }
no_token_file() { if find "$C" -name id-token 2>/dev/null | grep -q .; then prob "a token file exists"; fi; }
no_token_anywhere() { # the generated ID token in no file of the case and not in the job log
  if grep -rlF -e "$TOKEN" "$C" >/dev/null 2>&1; then prob "the ID token value is in a file of the case"; fi
}
refused16() { expect_rc 16; no_docker; no_private; no_token_file; no_token_anywhere; }
refused2() { expect_rc 2; no_docker; no_private; no_token_file; no_token_anywhere; }
gate() { # ACTION ARGS-AFTER-ACTION... with ENVV already adjusted: refused 16 before any token or Docker call
  launch "$@"; refused16
}
DARGS=(--environment staging --candidate "" --duration "")   # the job script's arguments for staging, empty inputs

# ---- gate cases ------------------------------------------------------------------------------------------------------------
case_g_not_ci() { base_env deploy; unsetenv GITLAB_CI; gate deploy --environment staging; }
case_g_usage() {
  base_env deploy; launch reset --environment staging; refused16
  base_env deploy; launch deploy --candidate ""; refused16
  base_env deploy; launch deploy --environment staging --environment staging; refused16
}
case_g_mr_iid() { base_env deploy; setenv CI_MERGE_REQUEST_IID=7; gate deploy --environment staging; }
case_g_unprotected() { base_env deploy; setenv CI_COMMIT_REF_PROTECTED=false; gate deploy --environment staging; }
case_g_other_branch() { base_env deploy; setenv CI_COMMIT_BRANCH=main CI_COMMIT_REF_NAME=main; gate deploy --environment staging; }
case_g_source_push() { base_env deploy; setenv CI_PIPELINE_SOURCE=push; gate deploy --environment staging; }
case_g_source_schedule() { base_env stop; setenv CI_PIPELINE_SOURCE=schedule; gate stop --environment staging; }
case_g_source_trigger() { base_env deploy; setenv CI_PIPELINE_SOURCE=trigger; gate deploy --environment staging; }
case_g_source_mr() { base_env deploy; setenv CI_PIPELINE_SOURCE=merge_request_event; gate deploy --environment staging; }
case_g_source_parent() { base_env deploy; setenv CI_PIPELINE_SOURCE=parent_pipeline; gate deploy --environment staging; }
case_g_wrong_job() { base_env deploy; setenv CI_JOB_NAME=staging:stop; gate deploy --environment staging; }
case_g_env_mismatch() { base_env deploy preview-1; gate deploy --environment staging; }
case_g_env_prefix() { base_env deploy preview-10; gate deploy --environment preview-1; }
case_g_project_id() { base_env deploy; setenv CI_PROJECT_ID=4243; gate deploy --environment staging; }
case_g_project_path() { base_env deploy; setenv CI_PROJECT_PATH=fixture/group/project-fork; gate deploy --environment staging; }
case_g_pin_source() { # the launcher reads the pin from hosted-context.sh: a different pin there refuses the same job
  sed -i "s/^STG_PROJECT_ID=.*/STG_PROJECT_ID=4243/" "$C/ns/project/infra/ci/hosted-context.sh"
  base_env deploy; gate deploy --environment staging
}
case_g_tag() { base_env deploy; setenv CI_COMMIT_TAG=0.2.0-rc.20; gate deploy --environment staging; }
case_g_no_token() { base_env deploy; unsetenv STAGING_CI_ID_TOKEN; gate deploy --environment staging; }
case_g_profile_missing() { base_env deploy; unsetenv STAGING_CI_TFVARS; gate deploy --environment staging; }
case_g_prepare_token() { base_env prepare; setenv STAGING_CI_ID_TOKEN="$TOKEN"; gate prepare --environment staging; }
case_g_prepare_environment() { base_env prepare; setenv CI_ENVIRONMENT_NAME=staging; gate prepare --environment staging; }
case_g_build_profile() { base_env build; setenv STAGING_CI_CONF="$C/ns/project.tmp/STAGING_CI_CONF"; gate build --environment staging; }

# ---- input cases -----------------------------------------------------------------------------------------------------------
case_i_stop_candidate() { base_env stop; launch stop --environment staging --candidate "$SHA" --duration ""; refused2; }
case_i_stop_duration() { base_env stop; launch stop --environment staging --candidate "" --duration 2h; refused2; }
case_i_env_none() { base_env deploy none; launch deploy --environment none; refused2; }
case_i_env_unregistered() { base_env deploy preview-10; launch deploy --environment preview-10; refused2; }
case_i_env_case() { base_env deploy Preview-1; launch deploy --environment Preview-1; refused2; }
case_i_candidate() {
  local c
  for c in 'a;b' '-x' "$(printf 'abc\nfeature')" "$(printf 'a%.0s' {1..201})" 'feature x'; do
    base_env deploy; launch deploy --environment staging --candidate "$c"; refused2
  done
}
case_i_duration() { local d; for d in 0 1234567 5s 2d 07; do base_env deploy; launch deploy --environment staging --duration "$d"; refused2; done; }
case_i_reuse() {
  base_env deploy; setenv STAGING_REUSE=maybe; launch deploy --environment staging; refused2
  base_env build; unsetenv STAGING_REUSE; launch build --environment staging; refused2
}
case_i_no_env_fallback() { # only the arguments select: inherited STAGING_CI_* values are neither a fallback nor forwarded
  base_env deploy; setenv STAGING_CI_ENV=staging; launch deploy --candidate ""; refused16
  rm -f "$C/docker-calls"
  base_env deploy preview-1; setenv STAGING_CI_ENV=staging STAGING_CI_CANDIDATE_INPUT=feature/x STAGING_CI_DURATION_INPUT=8h
  launch deploy --environment preview-1; expect_rc 0
  grep -qxF STAGING_CI_ENV=preview-1 "$C/ctx-env" 2>/dev/null || prob "the context did not get the argument's environment"
  grep -qxF STAGING_CI_CANDIDATE_INPUT= "$C/ctx-env" 2>/dev/null || prob "an inherited candidate reached the context"
  grep -qxF STAGING_CI_DURATION_INPUT= "$C/ctx-env" 2>/dev/null || prob "an inherited duration reached the context"
}
case_i_probes() { base_env deploy; setenv 'CI_DENIAL_PROBES=foo;rm -rf /'; launch deploy --environment staging; refused2; }

# ---- approved runs ---------------------------------------------------------------------------------------------------------
ALLOWED="HOME TMPDIR DOCKER_HOST DOCKER_TLS_VERIFY DOCKER_CERT_PATH STAGING_CI_ACTION STAGING_CI_ENV STAGING_CI_CANDIDATE_INPUT
  STAGING_CI_DURATION_INPUT STAGING_CI_PRIVATE STAGING_CI_OUT STAGING_CI_ID_TOKEN_FILE STAGING_CI_PROFILE_DIR STAGING_REUSE
  CI_DENIAL_PROBES GITLAB_CI CI_PIPELINE_SOURCE CI_COMMIT_SHA CI_COMMIT_BRANCH CI_COMMIT_REF_NAME CI_COMMIT_REF_PROTECTED
  CI_DEFAULT_BRANCH CI_PIPELINE_ID CI_PIPELINE_IID CI_PIPELINE_URL CI_JOB_ID CI_JOB_URL CI_JOB_NAME CI_JOB_STARTED_AT
  CI_JOB_TIMEOUT CI_ENVIRONMENT_NAME GITLAB_USER_LOGIN CI_PROJECT_DIR CI_PROJECT_ID CI_SERVER_URL CI_SERVER_HOST CI_RUNNER_ID
  CI_RUNNER_DESCRIPTION CI_RUNNER_TAGS CI_SERVER_VERSION RUNNER_SCRIPT_TIMEOUT"
approved_run() { # ACTION OUTDIR CONTEXT_DOCKERFILE REQUIRED_NAMES...: the common assertions of an approved run
  local a="$1" out="$2" df="$3" n run; shift 3
  [[ "$(grep -c '^run ' "$C/docker-calls" 2>/dev/null)" == 1 ]] || prob "not exactly one docker run"
  run="$(grep '^run ' "$C/docker-calls" 2>/dev/null)"
  [[ "$run" == *" bash infra/staging/ci.sh $a" ]] || prob "the context command is not bash infra/staging/ci.sh $a"
  [[ "$run" == *" --security-opt no-new-privileges "* && "$run" == *" --network host "* && "$run" != *--privileged* ]] || prob "context isolation flags differ"
  [[ "$run" == *" -v $C/ns:$C/ns "* ]] || prob "the build directory parent is not mounted at its own path"
  [[ "$run" == *" --mount type=tmpfs,destination=$C/ns/project.tmp,tmpfs-size=1m "* ]] || prob "the file-variable directory is not hidden"
  [[ "$(cat "$C/build-input" 2>/dev/null)" == "$(head -n 1 "$REPO/$df")" ]] || prob "the context image is not $df"
  while IFS= read -r n; do [[ " $ALLOWED " == *[[:space:]]"$n"[[:space:]]* ]] || prob "context variable $n is outside the allow-list"; done <"$C/ctx-env-names"
  for n in STAGING_CI_ACTION STAGING_CI_ENV STAGING_CI_PRIVATE STAGING_CI_OUT CI_JOB_ID CI_PROJECT_DIR "$@"; do
    grep -qxF "$n" "$C/ctx-env-names" 2>/dev/null || prob "context variable $n missing"
  done
  for n in STAGING_CI_ID_TOKEN STAGING_LEASE_HELD STAGING_CI_CONF STAGING_CI_BACKEND STAGING_CI_TFVARS STAGING_CI_SETTINGS STAGING_REUSE; do
    if grep -qxF "$n" "$C/docker-client-env-names" 2>/dev/null; then prob "$n is still set in the job shell (docker client environment)"; fi
  done
  grep -qxF "STAGING_CI_OUT=$C/ns/project/$out" "$C/ctx-env" 2>/dev/null || prob "STAGING_CI_OUT is not $out"
  grep -qxF "STAGING_CI_PRIVATE=$(P_DIR)" "$C/ctx-env" 2>/dev/null || prob "STAGING_CI_PRIVATE is not the private directory"
  no_private; no_token_anywhere
  if grep -qF -e "$JOBTOK" "$C/docker-calls" "$C/ctx-env" "$C/joblog" 2>/dev/null; then prob "the job token reached the context or the log"; fi
}
token_ok() { # the token file was a mode-600 file holding exactly the token, in the private directory outside the project
  local mode sum path
  read -r mode sum path <"$C/token-facts" 2>/dev/null || { prob "no token file reached the adapter"; return; }
  [[ "$mode" == 600 ]] || prob "token file mode $mode"
  [[ "$sum" == "$(printf '%s' "$TOKEN" | sha256sum | cut -d' ' -f1)" ]] || prob "the token file does not hold exactly the token"
  [[ "$path" == "$(P_DIR)/id-token" && "$path" != "$C/ns/project/"* ]] || prob "token file location $path"
}
summary_ok() { # the fixed-field summary in ci-publish.sh's key order
  local keys got
  keys="$(sed -n 's/^SUMMARY_KEYS="\(.*\)"$/\1/p' "$SUBJ_P")"
  got="$(sed 's/: .*//' "$C/ns/project/staging-summary/summary.txt" 2>/dev/null | tr '\n' ' ')"
  [[ -n "$keys" && "$got" == "$keys " ]] || prob "the summary keys differ: [$got]"
  grep -q '^staging-action: summary' "$C/joblog" || prob "the job log carries no summary"
  local k
  for k in $keys; do grep -q "^$k: " "$C/joblog" || prob "the job log carries no $k summary line"; done
}
case_p_deploy() {
  base_env deploy; launch deploy "${DARGS[@]}"
  expect_rc 0; approved_run deploy staging-summary infra/docker/staging-ci.Dockerfile STAGING_CI_ID_TOKEN_FILE STAGING_CI_PROFILE_DIR STAGING_REUSE
  token_ok; summary_ok
  [[ "$(cat "$C/profile-files" 2>/dev/null)" == "backend.hcl ci-settings staging.conf terraform.tfvars " ]] || prob "deploy profile: $(cat "$C/profile-files" 2>/dev/null)"
  grep -qxF STAGING_REUSE=no "$C/ctx-env" || prob "STAGING_REUSE not forwarded"
}
case_p_check() {
  base_env check; launch check "${DARGS[@]}"
  expect_rc 0; approved_run check staging-check infra/docker/staging-ci.Dockerfile STAGING_CI_ID_TOKEN_FILE STAGING_CI_PROFILE_DIR
  token_ok
  [[ "$(cat "$C/profile-files" 2>/dev/null)" == "ci-settings staging.conf " ]] || prob "check profile: $(cat "$C/profile-files" 2>/dev/null)"
  if grep -qxF STAGING_REUSE "$C/ctx-env-names"; then prob "check received STAGING_REUSE"; fi
  [[ "$(cat "$C/ns/project/staging-check/reuse.env" 2>/dev/null)" == STAGING_REUSE=no ]] || prob "reuse.env missing"
}
case_p_stop() {
  base_env stop; launch stop "${DARGS[@]}"
  expect_rc 0; approved_run stop staging-summary infra/docker/staging-ci.Dockerfile STAGING_CI_ID_TOKEN_FILE STAGING_CI_PROFILE_DIR
  token_ok; summary_ok
  [[ "$(cat "$C/profile-files" 2>/dev/null)" == "backend.hcl ci-settings staging.conf terraform.tfvars " ]] || prob "stop profile"
  grep -qxF STAGING_CI_CANDIDATE_INPUT= "$C/ctx-env" || prob "stop forwarded a candidate"
}
case_s_sanitizer_sync() {
  local actual expected
  actual="$(sed -n "/^SAN_AWK='/,/^}'$/p" "$SUBJ_L")"
  expected="$(sed -n "/^SAN_AWK='/,/^}'$/p" "$REPO/infra/ci/hosted-context.sh")"
  [[ -n "$actual" && "$actual" == "$expected" ]] || prob "the deployment and validation Git-config sanitizers differ"
}
case_p_runner_url() {
  git -C "$C/ns/project" config remote.origin.url "https://gitlab-ci-token:$JOBTOK@gitlab.example.invalid/project.git"
  git -C "$C/ns/project" config remote.origin.pushurl "https://gitlab-ci-token:$JOBTOK@gitlab.example.invalid/push.git"
  case_p_stop
  [[ "$(git -C "$C/ns/project" config remote.origin.url)" == https://gitlab-ci-token@gitlab.example.invalid/project.git ]] || prob "remote URL not sanitized"
  [[ "$(git -C "$C/ns/project" config remote.origin.pushurl)" == https://gitlab-ci-token@gitlab.example.invalid/push.git ]] || prob "push URL not sanitized"
}
case_c_unsafe_url() {
  git -C "$C/ns/project" config 'url.https://user:password@gitlab.example.invalid/.insteadOf' https://gitlab.example.invalid/
  base_env stop; launch stop "${DARGS[@]}"
  expect_rc 16; no_private
  if grep -q '^run ' "$C/docker-calls" 2>/dev/null; then prob "a context started"; fi
}
case_p_prepare() {
  base_env prepare; launch prepare "${DARGS[@]}"
  expect_rc 0; approved_run prepare staging-prepare infra/docker/ci-context.Dockerfile
  [[ ! -e "$C/token-facts" && ! -e "$C/profile-files" ]] || prob "prepare received a token file or profile"
  git -C "$C/ns/project" rev-parse -q --verify refs/remotes/origin/feature/x >/dev/null || prob "prepare did not fetch the branches"
  if git -C "$C/ns/project" for-each-ref | grep -q merge-requests; then prob "a merge-request ref was fetched"; fi
}
case_p_build() {
  base_env build; setenv STAGING_REUSE=yes; launch build "${DARGS[@]}"
  expect_rc 0; approved_run build staging-build infra/docker/ci-context.Dockerfile STAGING_REUSE
  grep -qxF STAGING_REUSE=yes "$C/ctx-env" || prob "STAGING_REUSE not forwarded"
  if grep -qE '^STAGING_CI_(ID_TOKEN_FILE|PROFILE_DIR)$' "$C/ctx-env-names"; then prob "build received credential paths"; fi
}
case_p_exit_class() {
  base_env deploy; setenv FAKE_MODE=exit3; launch deploy "${DARGS[@]}"
  expect_rc 3; no_private
  grep -qx 'exit_class: 3' "$C/ns/project/staging-summary/summary.txt" 2>/dev/null || prob "the fallback summary does not carry exit class 3"
}
case_p_no_summary() {
  base_env stop; setenv FAKE_MODE=no-summary; launch stop "${DARGS[@]}"
  expect_rc 0; summary_ok
  grep -qx 'exit_class: 0' "$C/ns/project/staging-summary/summary.txt" 2>/dev/null || prob "the fallback summary does not carry exit class 0"
}

# ---- failures, cancellation, after-script --------------------------------------------------------------------------------------
case_c_check_dotenv() {
  base_env check; setenv FAKE_MODE=check-bad; launch check "${DARGS[@]}"
  expect_rc 2; no_private
  [[ ! -e "$C/ns/project/staging-check/reuse.env" ]] || prob "a malformed dotenv was kept"
}
case_c_token_leak() {
  base_env deploy; setenv FAKE_MODE=leak; launch deploy "${DARGS[@]}"
  expect_rc 6; no_private; no_token_anywhere
  [[ ! -e "$C/ns/project/staging-summary/oops.txt" ]] || prob "a leaking output was kept"
  grep -qx 'exit_class: 6' "$C/ns/project/staging-summary/summary.txt" 2>/dev/null || prob "no summary with exit class 6"
}
case_c_jobtoken_file() {
  printf 'leftover %s\n' "$JOBTOK" >"$C/ns/project/stray.txt"
  base_env deploy; launch deploy "${DARGS[@]}"
  expect_rc 16; no_private; no_token_anywhere
  if grep -q '^run ' "$C/docker-calls" 2>/dev/null; then prob "a context started"; fi
}
case_c_mount() { # the private directory on another filesystem than the build directory
  cat >"$C/bin/stat" <<'EOF'
#!/usr/bin/env bash
if [[ "$*" == "-c %d "*/.staging-ci-[0-9]* && "$*" != */.staging-ci-*/* ]]; then echo 99999; exit 0; fi
exec /usr/bin/stat "$@"
EOF
  chmod 755 "$C/bin/stat"
  base_env deploy; launch deploy "${DARGS[@]}"
  expect_rc 16; no_private; no_token_anywhere
  grep -q "not on the build directory's filesystem" "$C/joblog" || prob "not refused for the filesystem"
}
case_c_cleanup_fail() { # removal fails: an otherwise successful run exits 4 and says so
  cat >"$C/bin/rm" <<'EOF'
#!/usr/bin/env bash
for a in "$@"; do case "$a" in */.staging-ci-*) exit 1 ;; esac; done
exec /bin/rm "$@"
EOF
  chmod 755 "$C/bin/rm"
  base_env stop; launch stop "${DARGS[@]}"
  expect_rc 4
  grep -q 'the private directory could not be removed' "$C/joblog" || prob "the failed removal is not reported"
  /bin/rm -rf "$(P_DIR)"
}
case_c_after_script() { # a job ended before its cleanup: the after-script removes the private directory and writes the summary
  mkdir -m 700 "$(P_DIR)"; printf 'stale\n' >"$(P_DIR)/id-token"
  base_env deploy; launch --after-script deploy --environment staging
  expect_rc 0; no_private
  grep -q '^container inspect platform-staging-ci-4711' "$C/docker-calls" 2>/dev/null || prob "the after-script did not look for the context"
  grep -q '^exit_class: unknown' "$C/ns/project/staging-summary/summary.txt" 2>/dev/null || prob "no after-script summary"
  base_env deploy; unsetenv GITLAB_CI; launch --after-script deploy --environment staging; expect_rc 0
}
case_c_signal() { # TERM during the context: the context is stopped, the private directory removed, the job fails
  local lp i
  base_env deploy; setenv FAKE_MODE=sleep
  env -i "${ENVV[@]}" sh "$C/ns/project/infra/ci/staging-action.sh" deploy "${DARGS[@]}" >"$C/joblog" 2>&1 &
  lp=$!
  for ((i = 0; i < 100; i++)); do [[ -e "$C/adapter-started" ]] && break; sleep 0.1; done
  [[ -e "$C/adapter-started" ]] || prob "the context did not start"
  kill -TERM "$lp" 2>/dev/null
  RC=0; wait "$lp" || RC=$?
  [[ "$RC" != 0 ]] || prob "a canceled job exited 0"
  grep -q '^stop -t 60 platform-staging-ci-4711' "$C/docker-calls" 2>/dev/null || prob "the context was not stopped"
  no_private; no_token_anywhere
  if [[ -s "$C/ctx.pid" ]] && kill -0 "$(cat "$C/ctx.pid")" 2>/dev/null; then prob "the context outlived the job"; kill "$(cat "$C/ctx.pid")"; fi
}

# ---- publication -----------------------------------------------------------------------------------------------------------
case_pub_plan() { # a session change prints a Terraform plan with the canary: published nowhere, raw/ removed
  base_env deploy; launch deploy "${DARGS[@]}"
  expect_rc 0; no_private
  if grep -rlF G072PLANCANARY "$C/joblog" "$C/ns/project/staging-summary" "$C/gcs" >/dev/null 2>&1; then prob "the plan canary was published"; fi
  local up="$C/gcs/logs/900-4711.log"
  [[ -f "$up" ]] || { prob "no filtered log was uploaded"; return; }
  grep -qx $'session\t600s\tscheduled stop 2026-10-28T12:00:00Z' "$up" && grep -qx $'result\tREADY\t'"$SHA" "$up" &&
    grep -qx $'secret\tIDENTITY_WORKOS_API_KEY\t3' "$up" || prob "allowlisted lines missing from the upload"
  if grep -qE $'^(vm\t|\\+ |Terraform|staging-vm:|secret\tCF_API_TOKEN)' "$up"; then prob "a non-allowlisted line was uploaded"; fi
  [[ "$(tail -n 1 "$up")" =~ ^dropped-lines:\ ([0-9]+)$ && "${BASH_REMATCH[1]}" -ge 10 ]] || prob "no dropped-line count"
}
case_pub_inline() { # the canary inside an allowlisted line: exit 6, nothing published
  base_env deploy; setenv FAKE_MODE=inline; launch deploy "${DARGS[@]}"
  expect_rc 6; no_private
  [[ ! -e "$C/gcs" ]] || prob "something was uploaded"
  if grep -rlF G072PLANCANARY "$C/joblog" "$C/ns/project/staging-summary" >/dev/null 2>&1; then prob "the plan canary was published"; fi
  grep -q 'plan-canary' "$C/joblog" || prob "the canary class is not reported"
  if grep -q $'^session\t' "$C/joblog"; then prob "filtered lines reached the job log"; fi
}
case_pub_token() { # the ID token in a raw line: exit 6, nothing published, the value nowhere
  base_env deploy; setenv FAKE_MODE=token; launch deploy "${DARGS[@]}"
  expect_rc 6; no_private; no_token_anywhere
  [[ ! -e "$C/gcs" ]] || prob "something was uploaded"
  grep -q ' literal$' "$C/joblog" || prob "the literal class is not reported"
}
case_pub_kinds() { # the allowlist against the printf/say keys of staging.sh, lib.sh and remote/*.sh
  local derived k in="$C/kinds.in" out="$C/kinds.out" want="$C/kinds.want" kinds n=0
  derived="$(grep -ohE "printf '[a-z][a-z_-]*\\\\t" "$REPO/infra/staging/staging.sh" "$REPO/infra/staging/lib.sh" "$REPO"/infra/staging/remote/*.sh |
    sed -E "s/^printf '//; s/\\\\t\$//" | LC_ALL=C sort -u)"
  grep -q "^say() { printf 'step\\\\t" "$REPO/infra/staging/remote/lib.sh" || prob "remote/lib.sh say no longer prints step lines"
  grep -q "^log() { printf 'staging: %s" "$REPO/infra/staging/lib.sh" || prob "lib.sh log no longer prints staging: lines"
  kinds="$(bash -c 'source "$1"; printf "%s" "$PUBLISH_KINDS"' _ "$SUBJ_P")"
  [[ "$kinds" == "step result session address reconcile receipt digest cert marker service secret" ]] || prob "PUBLISH_KINDS differs from Plan §10.7: $kinds"
  for k in $kinds; do grep -qxF "$k" <<<"$derived" || prob "allowlisted kind $k is not printed by the G070 scripts"; done
  : >"$in"; : >"$want"
  while IFS= read -r k; do
    if [[ $k == secret ]]; then printf 'secret\tCF_API_TOKEN\t3\n' | tee -a "$want" >>"$in"; printf 'secret\tCF_API_TOKEN\tabcdef\n' >>"$in"; n=$((n + 1))
    elif [[ " $kinds " == *" $k "* ]]; then printf '%s\tvalue\n' "$k" | tee -a "$want" >>"$in"
    else printf '%s\tvalue\n' "$k" >>"$in"; n=$((n + 1)); fi
  done <<<"$derived"
  printf 'staging: a message\n' | tee -a "$want" >>"$in"
  printf 'result\tREADY\r\n' >>"$in"; printf 'result\tREADY\n' >>"$want"
  printf '%s\n' '+ terraform apply' '[vm] (write state)' 'staging-vm: message' '  # module.vm will be updated' 'Staging: x' 'staging:x' >>"$in"; n=$((n + 6))
  printf 'step\t\303\251\n' >>"$in"; printf 'step\t%s\n' "$(printf 'x%.0s' {1..2000})" >>"$in"; n=$((n + 2))
  printf 'dropped-lines: %d\n' "$n" >>"$want"
  bash -c 'source "$1"; publish_filter "$2" "$3"' _ "$SUBJ_P" "$in" "$out" || prob "publish_filter failed"
  cmp -s "$want" "$out" || prob "publish_filter output differs from the allowlist ($(diff "$want" "$out" | head -n 4 | tr '\n' ' ' | cut -c1-200))"
  bash -c 'source "$1"; publish_filter "$2"' _ "$SUBJ_P" "$in" 2>/dev/null && prob "publish_filter accepted a missing argument"
  bash -c 'source "$1"; publish_filter "$2" "$3"' _ "$SUBJ_P" "$C/missing" "$C/o2" 2>/dev/null && prob "publish_filter accepted a missing file"
}
case_pub_summary() {
  local f="$C/summary.txt" keys
  keys="action environment candidate selection branch actor pipeline job reuse digests session scheduled_stop readiness receipt_id receipt_sha256 fence_open exit_class"
  sum() { bash -c 'source "$1"; shift; summary_render "$@"' _ "$SUBJ_P" "$@"; }
  sum "$f" action=deploy environment=staging $'actor=a\tb\nc' exit_class=0 || prob "summary_render failed"
  [[ "$(sed 's/: .*//' "$f" | tr '\n' ' ')" == "$keys " ]] || prob "summary keys or order differ"
  grep -qx 'actor: a b c' "$f" && grep -qx 'receipt_id: -' "$f" && grep -qx 'exit_class: 0' "$f" || prob "summary values differ"
  rm -f "$f"; sum "$f" action=deploy token=x 2>/dev/null && prob "an unknown key was accepted"
  [[ ! -e "$f" ]] || prob "a refused summary was written"
  sum "$f" action=deploy action=stop 2>/dev/null && prob "a repeated key was accepted"
  sum "$f" action 2>/dev/null && prob "a malformed argument was accepted"
  sum "$f" "digests=$(printf 'd%.0s' {1..1500})" || prob "a long value was refused"
  [[ "$(grep '^digests: ' "$f" | wc -c)" -le 1011 ]] || prob "a long value was not bounded"
}
case_pub_canary() {
  local d="$C/canary" f out pk cf jwt cls
  mkdir -p "$d/raw"
  scan() { bash -c 'source "$1"; shift; canary_scan "$@"' _ "$SUBJ_P" "$@"; }
  # Shapes are assembled at run time so this source carries none of them.
  pk="-----BEGIN ""PRIVATE KEY-----" cf="cf""ut_$(random_alnum 48)"
  jwt="$(printf '{"alg":"RS256"}' | base64 | tr -d '=\n').$(printf '{"sub":"fixture-subject"}' | base64 | tr -d '=\n').sig"
  printf 'ya''29.%s\n' "$(random_alnum 20)" >"$d/access-token"
  printf '%s\n' "$jwt" >"$d/jwt"
  printf '%s\n' "$pk" >"$d/private-key"
  printf 'x %s y\n' "$cf" >"$d/cloudflare-token"
  printf 'cf''at_%s\n' "$(random_alnum 48)" >"$d/cloudflare-token-2"
  printf 'G072PLANCANARY\n' >"$d/plan-canary"
  for f in access-token jwt private-key cloudflare-token cloudflare-token-2 plan-canary; do
    cls="${f%-2}"
    out="$(scan "$d/$f")"; [[ $? == 6 && "$out" == "canary: $d/$f $cls" ]] || prob "class $cls not detected exactly once ($out)"
  done
  if scan "$d/jwt" "$d/cloudflare-token" | grep -qF -e "$jwt" -e "$cf"; then prob "a matched value was printed"; fi
  printf 'G072PLANCANARY in a saved plan\n' >"$d/raw/plan.log"
  scan "$d/raw/plan.log" >/dev/null || prob "the plan canary in raw/ was refused (raw holds plan text by design)"
  cp "$d/private-key" "$d/raw/key.log"; scan "$d/raw/key.log" >/dev/null && prob "a credential in raw/ was not refused"
  printf 'clean line\n' >"$d/clean"; scan "$d/clean" >/dev/null || prob "a clean file was refused"
  scan "$d/missing" >/dev/null && prob "an unreadable file was accepted"
  printf '%s' "fixture-literal-$(random_hex 8)" >"$d/tokenfile"; printf 'a %s b\n' "$(cat "$d/tokenfile")" >"$d/lit"
  out="$(env STAGING_CI_ID_TOKEN_FILE="$d/tokenfile" TMPDIR="$d" bash -c 'source "$1"; shift; canary_scan "$@"' _ "$SUBJ_P" "$d/lit" "$d/clean")"
  [[ $? == 6 && "$out" == "canary: $d/lit literal" ]] || prob "the ID token literal was not detected ($out)"
  printf 'other-literal-%s\n' "$(random_hex 8)" >"$d/literals"; printf 'z %s\n' "$(cat "$d/literals")" >"$d/lit2"
  out="$(env STAGING_CI_CANARY_LITERALS="$d/literals" TMPDIR="$d" bash -c 'source "$1"; shift; canary_scan "$@"' _ "$SUBJ_P" "$d/lit2")"
  [[ $? == 6 && "$out" == "canary: $d/lit2 literal" ]] || prob "a listed literal was not detected ($out)"
  if find "$d" -name 'canary-literals.*' | grep -q .; then prob "the literal pattern file was left behind"; fi
}

# ---- static ----------------------------------------------------------------------------------------------------------------
case_s_no_fallback() {
  if grep -qE '\$\{?STAGING_' "$REPO/.gitlab-ci.yml"; then prob ".gitlab-ci.yml references a STAGING_* variable"; fi
  if grep -q '^variables:' "$REPO/.gitlab-ci.yml"; then prob ".gitlab-ci.yml has top-level variables"; fi
  if grep -qE 'inputs\.images|^ +images:' "$REPO/.gitlab-ci.yml"; then prob "an images input exists"; fi
  if grep -qE '\$\{?STAGING_CI_(ENV|CANDIDATE_INPUT|DURATION_INPUT|ACTION|IMAGES|PRIVATE|OUT)\b' "$SUBJ_L"; then prob "the launcher reads an inherited selection variable"; fi
  if grep -qE '\$\{?STAGING_LEASE_HELD' "$SUBJ_L"; then prob "the launcher reads STAGING_LEASE_HELD"; fi
}

# ---- mutation controls -----------------------------------------------------------------------------------------------------
run_mutation() { # NAME: the target case against a mutated subject must FAIL
  local spec="" m name target subj prog cp e mp
  for m in "${MUTATIONS[@]}"; do [[ "${m%%|*}" == "$1" ]] && spec="$m"; done
  IFS='|' read -r name target subj prog <<<"$spec"
  mp="$S/mut/$name"; mkdir -p "$mp"
  if [[ $subj == L ]]; then cp="$mp/staging-action.sh"; cp "$LAUNCHER" "$cp"; else cp="$mp/ci-publish.sh"; cp "$PUBLISH" "$cp"; fi
  while [[ -n "$prog" ]]; do e="${prog%%;;*}"; sed -i -e "$e" "$cp"; [[ "$prog" == *";;"* ]] && prog="${prog#*;;}" || prog=""; done
  if cmp -s "$cp" "$([[ $subj == L ]] && echo "$LAUNCHER" || echo "$PUBLISH")"; then MPROB="the mutation does not apply"; return; fi
  if [[ $subj == L ]]; then SUBJ_L="$cp"; else SUBJ_P="$cp"; fi
  setup_case "$name-$target"; "case_${target//-/_}"
  SUBJ_L="$LAUNCHER" SUBJ_P="$PUBLISH"
  if [[ -z "$PROB" ]]; then MPROB="$target still passes with the mutation"; else MPROB=""; MDETAIL="$target refused it: ${PROB%%; *}"; fi
}

# ---- run -------------------------------------------------------------------------------------------------------------------
printf 'case\tresult\tdetail\n' >"$REPORT/cases.tsv"
NSEL=0 NPASS=0 NFAIL=0
for name in "${CASES[@]}"; do
  [[ -z "$ONLY" || ",$ONLY," == *",$name,"* ]] || continue
  NSEL=$((NSEL + 1))
  if [[ $name == m-* ]]; then
    MPROB="" MDETAIL=""; run_mutation "$name"
    if [[ -z "$MPROB" ]]; then res=PASS det="$MDETAIL"; else res=FAIL det="$MPROB"; fi
  else
    setup_case "$name"; "case_${name//-/_}"
    if [[ -z "$PROB" ]]; then res=PASS det="ok"; else res=FAIL det="${PROB%; }"; fi
  fi
  [[ $name == p-* || $name == pub-* ]] && det="$det; $REGISTRY_NOTE"
  det="${det//$'\t'/ }"; det="${det//$'\n'/ }"
  printf '%s\t%s\t%s\n' "$name" "$res" "$det" >>"$REPORT/cases.tsv"
  printf '%s %s %s\n' "$name" "$res" "$det"
  if [[ $res == PASS ]]; then NPASS=$((NPASS + 1)); else NFAIL=$((NFAIL + 1)); fi
done
printf 'selected\tpassed\tfailed\n%s\t%s\t%s\n' "$NSEL" "$NPASS" "$NFAIL" >"$REPORT/summary.tsv"
printf 'ci-launcher: %s selected, %s passed, %s failed\n' "$NSEL" "$NPASS" "$NFAIL"
((NFAIL == 0))
