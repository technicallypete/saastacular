#!/usr/bin/env bash
# Baseline CI entry point (docs/foundation/OPERATIONS.md). It runs the repository's containerized
# validation harness under a parent-owned supervisor and then always finalizes, once: postcheck,
# positive-ownership cleanup, evidence scans, completeness, all-or-nothing staging, a pre-publication
# summary with a supplemental JUnit report (G050; infra/ci/junit.sh), a candidate scan, atomic
# publication (CI only), the job-log report and a scrub. The exit status aggregates every recorded step; a
# success never masks a recorded failure.
#
#   bash infra/ci/validate.sh [--profile fast|gate|staging|affected|integration|full] [--base REV | --compare OLD]
#     [--evidence-dir DIR] [--publish-dir ci-evidence]
#
# Profiles (G003 CI revision; infra/ci/lib.sh): in GitLab CI --profile is fast (the fast job) or gate (the
# gate job; infra/ci/select-profile.sh decides integration, full or none from the pipeline context and the
# changed paths); a local run takes fast, affected, integration or full. Without --profile the run is full. The
# identity-staging job (G004 CI-1) runs --profile staging: the real WorkOS staging contract, only with the
# credential file infra/ci/hosted-context.sh forwards in the approved context (IDENTITY_STAGING_ENV_FILE).
# G051: the selector also names the pipeline and executed suites, the areas and a selection digest; each job
# executes its partition (the fast job: fast or affected; the gate: the remainder of an integration or full
# pipeline, or nothing) and validate.sh re-derives that partition before anything runs. --base REV and
# --compare OLD are local options of --profile affected, passed to the selector (refused in CI).
#
# Exit: 0 pass; the harness status when it is nonzero (1 or 2 from run.sh, 124 at the internal deadline,
# 128+n on signal n, 143 when the finalizer stopped it); 128+n on a signal with no harness running;
# 13 postcheck drift; 14 evidence incomplete or malformed; 15 unsafe evidence (withheld); 16 location,
# interrupted-run recovery (local; a launch blocked by an unrecovered earlier run), runner preflight,
# attribution, precheck, budget or state import; 17 scanner failure (nothing
# published); 18 cleanup incomplete, refused or unproven; 19 staging, summary, publication, report or
# scrub failure, an orchestration error, or any other recorded nonzero status.
# The authoritative result is this exit status and the final "validate: final status" line. The
# published summary.json and phase-status.tsv are pre-publication metadata.
#
# Requires Bash >= 4.4, GNU coreutils/findutils/grep, util-linux setsid, git and Docker. State shared
# between steps lives in parent-shell variables; isolated steps hand values back only through a
# validated state file (state_import). Every check returns explicitly; no statement is negated with `!`.
set -euo pipefail

if (( BASH_VERSINFO[0] < 4 || (BASH_VERSINFO[0] == 4 && BASH_VERSINFO[1] < 4) )); then
  echo "validate: bash >= 4.4 is required" >&2
  exit 16
fi
shopt -s inherit_errexit

SELF_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)"
REPO="$(dirname -- "$(dirname -- "$SELF_DIR")")"
if [[ ! -f "$SELF_DIR/lib.sh" ]]; then echo "validate: infra/ci/lib.sh is missing" >&2; exit 16; fi
# shellcheck source=lib.sh
source "$SELF_DIR/lib.sh"
if [[ ! -f "$SELF_DIR/junit.sh" ]]; then echo "validate: infra/ci/junit.sh is missing" >&2; exit 16; fi
# shellcheck source=junit.sh
source "$SELF_DIR/junit.sh"

# Git must not write repository metadata (no optional locks, no fsmonitor) and must not be redirected
# by the caller's environment.
unset GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE GIT_OBJECT_DIRECTORY GIT_ALTERNATE_OBJECT_DIRECTORIES \
  GIT_COMMON_DIR GIT_NAMESPACE GIT_CEILING_DIRECTORIES
export GIT_OPTIONAL_LOCKS=0
# G051: the harness reads the bootstrap record only from validate.sh's own hosted-context preflight, never from
# the caller's environment.
unset PLATFORM_LEDGER_BOOTSTRAP_FILE

# --- arguments (a usage error is recorded as a location failure, after the traps are installed) -------
EVARG="" PUBARG="" PROFILE_ARG="" BASE_ARG="" COMPARE_ARG="" ARG_ERR=0 SEEN_EV=0 SEEN_PUB=0 SEEN_PROF=0 SEEN_BASE=0 SEEN_CMP=0
while (( $# > 0 )); do
  case "$1" in
    --profile)
      if (( $# < 2 || SEEN_PROF )) || [[ -z "$2" ]]; then ARG_ERR=1; break; fi
      PROFILE_ARG="$2"; SEEN_PROF=1; shift 2 ;;
    --base)
      if (( $# < 2 || SEEN_BASE )) || [[ -z "$2" ]]; then ARG_ERR=1; break; fi
      BASE_ARG="$2"; SEEN_BASE=1; shift 2 ;;
    --compare)
      if (( $# < 2 || SEEN_CMP )) || [[ -z "$2" ]]; then ARG_ERR=1; break; fi
      COMPARE_ARG="$2"; SEEN_CMP=1; shift 2 ;;
    --evidence-dir)
      if (( $# < 2 || SEEN_EV )) || [[ -z "$2" ]]; then ARG_ERR=1; break; fi
      EVARG="$2"; SEEN_EV=1; shift 2 ;;
    --publish-dir)
      if (( $# < 2 || SEEN_PUB )) || [[ -z "$2" ]]; then ARG_ERR=1; break; fi
      PUBARG="$2"; SEEN_PUB=1; shift 2 ;;
    *) ARG_ERR=1; break ;;
  esac
done

# --- parent-held state -----------------------------------------------------------------------------------
declare -A ST=() PHASE_SECONDS=()
# The single phase list: the finalizer records into it and the aggregator iterates over it.
PHASES=(location recovery preflight state harness signal postcheck cleanup raw_scan complete stage summary junit
  candidate_scan summary_withheld junit_withheld rescan publish report scrub internal)
LOC_OK=0 MAIN_DONE=0 FINALIZED=0 DEADLINE=0 STRAGGLERS=0 CI_MODE=0
EV_CREATED=0 EV_PROVEN=0 REG_CREATED=0 REG_PROVEN=0
HPID="" WPID="" EV="" REG="" REPO_R="" PUBLISH_TARGET=""
SIG_NAME="" HARNESS_STOP=none REGSTATE=none PUBLISHED=none FINAL=19 HARNESS_DONE=0
HEAD_SHA="" TREE_SHA="" STATUS_DIGEST="" BUDGET_S="" TESTED_CONTENT=""
# G051 cleanup amendment: the registry's job, the local interrupted-run marker directory and this run's marker,
# and where a local run kept its scanned storage report.
REG_JOB="" PEND_DIR="" PEND_FILE="" STORAGE_KEPT="" MARKER_BASE="" STATE_ROOT="" LEGACY_PEND_DIR=""
JOB_KIND="" PIPELINE_PROFILE="" EXECUTED_PROFILE="" SELECTION_REASON="" COMPARISON_BASE=""
AREAS="" PIPELINE_SUITES="" EXECUTED_SUITES="" COVERAGE="" SELECTION_DIGEST="" PARTITION=""
STARTED_UTC="$(ci_utc)"
# The selection keys (G051; infra/ci/select-profile.sh writes them, validate.sh itself for staging and the CI
# default full) and their fixed expressions; staging is a validate.sh-written profile only.
RE_PIPE='^(fast|affected|integration|full|staging)$'
RE_EXEC='^(fast|affected|integration|full|staging|none)$'
RE_AREAS='^(none|(web|core|intel):[a-z][a-z0-9-]{0,31}(,(web|core|intel):[a-z][a-z0-9-]{0,31})*)$'
RE_PSUITES='^[a-z][a-z-]*( [a-z][a-z-]*)*$'
RE_ESUITES='^(none|[a-z][a-z-]*( [a-z][a-z-]*)*)$'
SEL_SPECS=("PIPELINE_PROFILE=$RE_PIPE" "EXECUTED_PROFILE=$RE_EXEC" 'SELECTION_REASON=^[a-z][a-z-]{1,31}$'
  'COMPARISON_BASE=^([0-9a-f]{40}|none)$' "AREAS=$RE_AREAS" "PIPELINE_SUITES=$RE_PSUITES" "EXECUTED_SUITES=$RE_ESUITES"
  'COVERAGE=^(complete|partial|none)$' 'SELECTION_DIGEST=^[0-9a-f]{64}$')
# The suites an affected pipeline may name (Plan §3.3 rule 8: the fast suites and the per-kind area suites).
AFFECTED_SUITES=" lint lint-base lint-web lint-core provenance contracts unit unit-web unit-core unit-intel feature-web feature-core "
# The hosted adapter's context image (infra/ci/hosted-context.sh; G051 bootstrap record).
CTX_DOCKERFILE="infra/docker/ci-context.Dockerfile"
STOP_GRACE_DS=1500   # 150 s, in tenths of a second
RECOVERY_WAIT_DS=50  # 5 s: how long recovery lets an interrupted run's watchdog end and its KILLed group be reaped

MARKER_WITHHELD="Evidence withheld: a scan found unsafe content in the evidence. The authoritative result is the job status and the final job-log line."
MARKER_STAGING="Evidence not published: staging failed. The authoritative result is the job status and the final job-log line."

# Allow-listed evidence (copied by name; anything else, including harness-console.log, deps-codes.txt
# and cleanup.log, is never staged under those names). The scanned console is published as harness.log. Browser JSON is staged with every attachments entry removed.
STAGE_REPORT=(results.tsv images.txt
  browser-root-tests.tsv browser-off-tests.tsv browser-rename-tests.tsv
  browser-root-results.json browser-off-results.json browser-rename-results.json
  deps-controls.tsv wrapper-controls.log ci-controls.log
  contracts-results-go.json contracts-results-python.json contracts-agree.tsv
  schema-negative-go.json schema-negative-python.json
  error-output-core.tsv error-output-web.tsv error-output-intelligence.tsv
  plan.out repeat-render.out catalog-p.log migrate-it.log deps-server.txt inventory-leftover.txt resources.tsv timings.tsv
  builds.tsv ledger-summary.tsv ledger-bootstrap.tsv ci-controls-timings.tsv
  browser-feature-results.json browser-feature-tests.tsv
  go-tests.tsv go-tests-expected.tsv go-tests-junit.xml db-lifecycle.tsv)
# G064 (Plan r2 §5.3-§5.4): the per-test Go evidence of the migrations and ops suites above, plus the go-tests/
# directory of redacted, bounded per-invocation logs (staged like diagnostics/). Supplemental to results.tsv and the
# authoritative junit.xml (D-5); both scans gate them like every other staged file.
STAGE_RAW=(attribution.json runner-preflight.tsv cleanup-inventory.tsv selection.tsv storage.tsv recovery.tsv)
ATTR_VARS=(GITLAB_CI CI_PIPELINE_SOURCE CI_COMMIT_SHA CI_COMMIT_BRANCH CI_COMMIT_REF_NAME CI_COMMIT_TAG
  CI_DEFAULT_BRANCH CI_MERGE_REQUEST_IID CI_MERGE_REQUEST_EVENT_TYPE CI_MERGE_REQUEST_SOURCE_BRANCH_NAME
  CI_MERGE_REQUEST_TARGET_BRANCH_NAME CI_MERGE_REQUEST_SOURCE_BRANCH_SHA CI_MERGE_REQUEST_TARGET_BRANCH_SHA
  CI_MERGE_REQUEST_DIFF_BASE_SHA CI_MERGE_REQUEST_APPROVED CI_PIPELINE_ID CI_PIPELINE_IID CI_PIPELINE_URL
  CI_JOB_ID CI_JOB_URL GITLAB_USER_LOGIN CI_RUNNER_ID CI_RUNNER_DESCRIPTION CI_RUNNER_TAGS
  CI_SERVER_VERSION CI_JOB_TIMEOUT CI_COMMIT_BEFORE_SHA CI_JOB_STARTED_AT CI_PROJECT_ID)
# ONEOFF RC qualification: the canonical release-candidate tag name (MAJOR.MINOR.PATCH-rc.N; no leading zeroes, N >= 1,
# lowercase rc, no v prefix or build suffix). The promotion merge request's source branch is release/<that name>; the
# workflow's admission rule uses the same expression, as a route only (never as proof).
RC_BODY='(0|[1-9][0-9]*)[.](0|[1-9][0-9]*)[.](0|[1-9][0-9]*)-rc[.]([1-9][0-9]*)'
RC_TAG_RE="^${RC_BODY}\$"
# The verified promotion identity of an mr-rc-promotion run ("none" otherwise), imported from the preflight state:
# the RC tag, its annotated tag object and this job's own observation of refs/remotes/origin/staging (staging may
# advance between the two jobs; each observation is checked independently, never required to be equal).
RELEASE_TAG="" RELEASE_TAG_OBJECT="" RELEASE_STAGING_SHA=""
# Every accepted tested_content value (attribution, the preflight state and the summary use this one expression).
TC_RE='^(local|branch-push|branch-schedule|branch-web|branch-api|mr-source|mr-merged-result|mr-rc-promotion)$'

say() { printf 'validate: %s\n' "$*"; }
inside_repo() { ci_path_within "$1" "$REPO_R"; }
g() { git -C "$REPO_R" -c core.fsmonitor=false --no-optional-locks "$@"; }

# phase NAME FUNC [ARGS...]: FUNC runs isolated, `set +e; ( set -e; FUNC ); rc=$?`. Only its status
# reaches the parent (ST[NAME]); the caller's errexit setting is restored afterwards.
phase() {
  local name="$1" rc had_e=0 started=$SECONDS
  shift
  if [[ $- == *e* ]]; then had_e=1; fi
  set +e
  ( set -e; "$@" )
  rc=$?
  if (( had_e )); then set -e; fi
  ST[$name]=$rc
  PHASE_SECONDS[$name]=$((SECONDS - started))
}

# --- L0 location (parent; containment is proven before anything is created) ---------------------------
make_registry() {
  local token job
  if mkdir -m 700 -- "$REG" 2>/dev/null; then REG_CREATED=1; else say "location: cannot create the registry"; return 16; fi
  REG="$(realpath -e -- "$REG" 2>/dev/null)" || return 16
  if inside_repo "$REG"; then return 16; fi
  REG_PROVEN=1
  mkdir -m 700 -- "$REG/cid" || return 16
  token="$(ci_random_hex 16)"
  if [[ ! "$token" =~ ^[0-9a-f]{32}$ ]]; then return 16; fi
  if (( CI_MODE )); then job="$CI_JOB_ID"; else job="local-$(ci_random_hex 6)"; fi
  if [[ ! "$job" =~ ^([1-9][0-9]{0,18}|local-[0-9a-f]{12})$ ]]; then return 16; fi
  REG_JOB="$job"
  (umask 077 && printf '%s\n' "$token" >"$REG/token") || return 16
  (umask 077 && printf '%s\njob\t%s\ntoken\t%s\nrepo\t%s\n' "$CI_REGISTRY_HEADER" "$job" "$token" "$REPO_R" \
    >"$REG/owned.tsv") || return 16
  return 0
}

# volatile_fs PATH: 0 when PATH, or its nearest existing ancestor, lies on tmpfs or ramfs, or its type is unreadable.
volatile_fs() {
  local p="$1" t
  while [[ ! -e "$p" && ! -L "$p" && "$p" != / ]]; do p="$(dirname -- "$p")"; done
  t="$(stat -f -c %T -- "$p" 2>/dev/null)" || return 0
  [[ "$t" == tmpfs || "$t" == ramfs ]]
}

# state_root_prepare (local runs; G064 Plan A corr2 §D3): STATE_ROOT = G003_STATE_ROOT, else
# ${XDG_STATE_HOME:-$HOME/.local/state}/platform-evidence/g003-local. Absolute, outside the repository, not on tmpfs/ramfs
# (checked before anything is created and again after), never a symlink, created mode 700 when absent (new parents private),
# owned by this user and mode 700 (an existing root is never chmod-ed). It holds the run markers (g003-pending/) and the
# default evidence directory with its registry. Exported as G003_STATE_ROOT so that the harness reads the same marker
# (lib.sh it_runner_precondition). A failure is location 16.
state_root_prepare() {
  local r="${G003_STATE_ROOT:-}"
  if [[ -z "$r" ]]; then
    if [[ -n "${XDG_STATE_HOME:-}" ]]; then r="$XDG_STATE_HOME"; elif [[ -n "${HOME:-}" ]]; then r="$HOME/.local/state"; else
      say "location: no state root (G003_STATE_ROOT, XDG_STATE_HOME and HOME are unset)"; return 16
    fi
    r="$r/platform-evidence/g003-local"
  fi
  if [[ "$r" != /* ]]; then say "location: the state root $r is not an absolute path"; return 16; fi
  if [[ -L "$r" ]]; then say "location: the state root $r is a symbolic link"; return 16; fi
  r="$(realpath -m -- "$r" 2>/dev/null)" || { say "location: the state root does not resolve"; return 16; }
  if inside_repo "$r"; then say "location: the state root $r lies inside the repository"; return 16; fi
  if volatile_fs "$r"; then say "location: the state root $r lies on tmpfs/ramfs (recovery state must survive a restart)"; return 16; fi
  if [[ ! -e "$r" ]]; then
    if (umask 077 && mkdir -p -- "$(dirname -- "$r")") 2>/dev/null && mkdir -m 700 -- "$r" 2>/dev/null; then :; else
      say "location: cannot create the state root $r"; return 16
    fi
  fi
  if [[ -L "$r" || ! -d "$r" || ! -O "$r" || "$(stat -c %a -- "$r" 2>/dev/null)" != 700 ]]; then
    say "location: the state root $r is not a private directory of this user (mode 700, owned, not a symbolic link)"; return 16
  fi
  if volatile_fs "$r"; then say "location: the state root $r lies on tmpfs/ramfs (recovery state must survive a restart)"; return 16; fi
  STATE_ROOT="$r"
  export G003_STATE_ROOT="$r"
}

locate() {
  local cand tmpd regc="" cpd
  REPO_R="$(realpath -e -- "$REPO" 2>/dev/null)" || return 16
  if (( ARG_ERR )); then say "usage: validate.sh [--profile fast|gate|staging|affected|integration|full] [--base REV | --compare OLD] [--evidence-dir DIR] [--publish-dir ci-evidence]"; return 16; fi
  if [[ -n "$PUBARG" && "$PUBARG" != ci-evidence ]]; then say "location: --publish-dir accepts only ci-evidence"; return 16; fi
  tmpd="$(realpath -m -- "${TMPDIR:-/tmp}" 2>/dev/null)" || return 16
  if inside_repo "$tmpd"; then say "location: TMPDIR lies inside the repository"; return 16; fi
  if [[ "${GITLAB_CI:-}" == true ]]; then
    CI_MODE=1
    if [[ ! "${CI_JOB_ID:-}" =~ ^[1-9][0-9]{0,18}$ ]]; then say "location: CI_JOB_ID is missing or malformed"; return 16; fi
    if [[ -z "${CI_PROJECT_DIR:-}" ]]; then say "location: CI_PROJECT_DIR is missing"; return 16; fi
    cpd="$(realpath -e -- "$CI_PROJECT_DIR" 2>/dev/null)" || { say "location: CI_PROJECT_DIR does not resolve"; return 16; }
    if [[ "$cpd" != "$REPO_R" ]]; then say "location: CI_PROJECT_DIR is not this checkout"; return 16; fi
    regc="$(ci_registry_dir "$cpd" "$CI_JOB_ID")" || return 16
    if inside_repo "$regc"; then say "location: the CI registry would lie inside the repository"; return 16; fi
    if [[ -e "$regc" || -L "$regc" ]]; then say "location: the CI registry already exists"; return 16; fi
    if [[ -n "$PUBARG" ]]; then PUBLISH_TARGET="$cpd/ci-evidence"; fi
  fi
  # Profile: CI accepts fast, gate or staging (the default is the full profile); a local run accepts fast,
  # integration or full. Staging needs the hosted adapter context and its forwarded credential file.
  if (( CI_MODE )); then
    case "$PROFILE_ARG" in
      fast|gate) JOB_KIND="$PROFILE_ARG" ;;
      staging)
        JOB_KIND=staging
        # The identity-staging job's own timeout floor holds first, before any credential-bearing step
        # (runner-preflight.sh, which knows only fast/gate floors, then checks the fast profile's capacity).
        if ci_time_ok staging "${CI_JOB_TIMEOUT:-}"; then :; else say "location: the identity-staging job timeout is missing or below its ${CI_JOBFLOOR_staging}s floor"; return 16; fi
        if ci_hosted_context; then :; else say "location: the staging profile runs only in the hosted adapter context"; return 16; fi
        if [[ "${IDENTITY_STAGING_ENV_FILE:-}" != /run/platform-identity-staging.env || -L "$IDENTITY_STAGING_ENV_FILE" || ! -f "$IDENTITY_STAGING_ENV_FILE" ]]; then
          say "location: the staging profile has no forwarded identity staging file"; return 16
        fi ;;
      "") JOB_KIND=full ;;
      *) say "location: --profile in CI accepts fast, gate or staging"; return 16 ;;
    esac
    if [[ -n "$BASE_ARG" || -n "$COMPARE_ARG" ]]; then say "location: --base and --compare are local options"; return 16; fi
  else
    case "$PROFILE_ARG" in
      fast|affected|integration|full) JOB_KIND=local ;;
      "") JOB_KIND=local PROFILE_ARG=full ;;
      *) say "location: --profile locally accepts fast, affected, integration or full"; return 16 ;;
    esac
    # G051 local selection (Plan §3.3 rule 9): one of --base or --compare, only with --profile affected; the
    # selector resolves the value (a revision, never an option).
    if [[ -n "$BASE_ARG" && -n "$COMPARE_ARG" ]]; then say "location: --base and --compare are mutually exclusive"; return 16; fi
    if [[ ( -n "$BASE_ARG" || -n "$COMPARE_ARG" ) && "$PROFILE_ARG" != affected ]]; then say "location: --base and --compare need --profile affected"; return 16; fi
    if [[ -n "$BASE_ARG$COMPARE_ARG" && ! "$BASE_ARG$COMPARE_ARG" =~ ^[^-[:space:][:cntrl:]][^[:space:][:cntrl:]]{0,255}$ ]]; then
      say "location: the --base or --compare revision is malformed"; return 16
    fi
  fi
  # G064 D3: local recovery state (markers, the default evidence directory and its registry) lives on persistent
  # private storage, never in TMPDIR; CI writes no marker and keeps its job registry.
  if (( CI_MODE == 0 )); then state_root_prepare || return 16; fi
  if [[ -n "$EVARG" ]]; then
    cand="$(realpath -m -- "$EVARG" 2>/dev/null)" || return 16
    if inside_repo "$cand"; then say "location: --evidence-dir lies inside the repository"; return 16; fi
    if [[ -e "$cand" || -L "$cand" ]]; then say "location: --evidence-dir already exists"; return 16; fi
    if (( CI_MODE == 0 )) && volatile_fs "$(dirname -- "$cand")"; then say "location: --evidence-dir lies on tmpfs/ramfs (local evidence and its registry must survive a restart)"; return 16; fi
    if mkdir -m 700 -- "$cand" 2>/dev/null; then EV_CREATED=1; EV="$cand"; else say "location: cannot create --evidence-dir"; return 16; fi
  elif (( CI_MODE )); then
    if [[ ! -d "$tmpd" ]]; then say "location: TMPDIR is not a directory"; return 16; fi
    cand="$(mktemp -d "$tmpd/g003-validate-XXXXXX" 2>/dev/null)" || { say "location: mktemp failed"; return 16; }
    EV_CREATED=1; EV="$cand"
  else
    cand="$(mktemp -d "$STATE_ROOT/g003-validate-XXXXXX" 2>/dev/null)" || { say "location: mktemp failed in $STATE_ROOT"; return 16; }
    EV_CREATED=1; EV="$cand"
  fi
  EV="$(realpath -e -- "$EV" 2>/dev/null)" || return 16
  if inside_repo "$EV"; then return 16; fi
  EV_PROVEN=1
  chmod 700 -- "$EV" || return 16
  if (( CI_MODE )); then REG="$regc"; else REG="$EV/owned"; PEND_DIR="$STATE_ROOT/g003-pending"; LEGACY_PEND_DIR="$tmpd/g003-pending"; fi
  make_registry || return 16
  mkdir -m 700 -- "$EV/raw" "$EV/state" "$EV/final" "$EV/scan" || return 16
  return 0
}

# --- P1 preflight (isolated; values reach the parent only through state/p1.env) ------------------------
preflight_rows_ok() { # FILE: RP1-RP13 exactly once each, well formed, none FAIL
  local f="$1" line n id st re=$'^(RP[0-9]{1,2})\t(PASS|FAIL|INFO)(\t[^\t]*)?$'
  local -A seen=()
  if [[ -L "$f" || ! -f "$f" || ! -s "$f" ]]; then return 1; fi
  while IFS= read -r line || [[ -n "$line" ]]; do
    if [[ ! "$line" =~ $re ]]; then return 1; fi
    id="${BASH_REMATCH[1]}" st="${BASH_REMATCH[2]}"
    if [[ ! "$id" =~ ^RP([1-9]|1[0-3])$ ]]; then return 1; fi
    if [[ -n "${seen[$id]+x}" ]]; then return 1; fi
    seen[$id]=1
    if [[ "$st" == FAIL ]]; then return 1; fi
  done <"$f"
  for ((n = 1; n <= 13; n++)); do
    if [[ -z "${seen[RP$n]+x}" ]]; then return 1; fi
  done
  return 0
}

# classify_promotion HEAD: a merge request pipeline whose source branch is release/... (ONEOFF RC qualification; the
# coordinator's tag-pinned promotion merge request to the default branch). Only the verified RC promotion is accepted;
# any failed check refuses, with no mr-source fallback, so a release/ branch prefix alone never qualifies. The
# pipeline must be detached: there GitLab leaves CI_COMMIT_BRANCH, CI_COMMIT_TAG and the merged-result
# CI_MERGE_REQUEST_SOURCE_BRANCH_SHA/TARGET_BRANCH_SHA empty, so a value in any of them is contamination (the source
# head is proven offline from the merge request API by release-verify.sh qualified, never from these variables). The
# refs come only from the runner's own credentialed fetch before credential removal (GIT_FETCH_EXTRA_FLAGS):
# refs/tags/TAG annotated and peeled to CI_COMMIT_SHA (= HEAD), and refs/remotes/origin/staging containing that commit.
# There is no fallback fetch; a missing ref is refused. Sets A_TC=mr-rc-promotion with A_TAG, A_TAGOBJ and A_STAGING,
# or A_REASON (fixed text).
classify_promotion() {
  local head="$1" src="${CI_MERGE_REQUEST_SOURCE_BRANCH_NAME:-}" tag x obj typ peel stg
  tag="${src#release/}"
  if [[ "${CI_MERGE_REQUEST_EVENT_TYPE:-}" != detached ]]; then A_REASON="RC promotion outside a detached merge request pipeline"; return 0; fi
  # (classify has already refused any CI_COMMIT_TAG and CI_COMMIT_BRANCH; the merged-result SHAs are checked here.)
  for x in CI_MERGE_REQUEST_SOURCE_BRANCH_SHA CI_MERGE_REQUEST_TARGET_BRANCH_SHA; do
    if [[ -n "${!x:-}" ]]; then A_REASON="RC promotion pipeline carries $x"; return 0; fi
  done
  if [[ -z "${CI_DEFAULT_BRANCH:-}" || "${CI_MERGE_REQUEST_TARGET_BRANCH_NAME:-}" != "$CI_DEFAULT_BRANCH" ]]; then
    A_REASON="RC promotion target is not the default branch"; return 0
  fi
  if [[ ! "$tag" =~ $RC_TAG_RE ]]; then A_REASON="RC promotion source branch is not release/ and a canonical RC name"; return 0; fi
  if [[ "${CI_COMMIT_REF_NAME:-}" != "$src" ]]; then A_REASON="CI_COMMIT_REF_NAME differs from the merge request source branch"; return 0; fi
  obj="$(g rev-parse --verify -q "refs/tags/$tag" 2>/dev/null)" || obj=""
  if [[ ! "$obj" =~ ^[0-9a-f]{40}$ ]]; then A_REASON="RC refs not prepared: refs/tags of the RC"; return 0; fi
  typ="$(g cat-file -t "$obj" 2>/dev/null)" || typ=""
  if [[ "$typ" != tag ]]; then A_REASON="RC tag is not annotated"; return 0; fi
  peel="$(g rev-parse --verify -q "refs/tags/$tag^{commit}" 2>/dev/null)" || peel=""
  # classify has already required HEAD = CI_COMMIT_SHA, so peel = HEAD is peel = CI_COMMIT_SHA too.
  if [[ "$peel" != "$head" ]]; then A_REASON="RC tag peel differs from the merge request head"; return 0; fi
  stg="$(g rev-parse --verify -q 'refs/remotes/origin/staging^{commit}' 2>/dev/null)" || stg=""
  if [[ ! "$stg" =~ ^[0-9a-f]{40}$ ]]; then A_REASON="RC refs not prepared: refs/remotes/origin/staging"; return 0; fi
  if g merge-base --is-ancestor "$peel" "$stg" 2>/dev/null; then :; else A_REASON="RC commit is not in origin/staging"; return 0; fi
  A_TC=mr-rc-promotion A_TAG="$tag" A_TAGOBJ="$obj" A_STAGING="$stg"
  return 0
}

# classify HEAD: checkout attribution from documented GitLab variables and the commit object only.
# Sets A_TC (tested_content) or A_REASON (a fixed reason; never a variable's value).
classify() {
  local head="$1" x b src tgt parents want
  A_TC="" A_REASON="" A_TAG=none A_TAGOBJ=none A_STAGING=none
  if [[ -z "${GITLAB_CI:-}" ]]; then
    for x in CI_PIPELINE_SOURCE CI_COMMIT_SHA CI_JOB_ID CI_MERGE_REQUEST_IID CI_MERGE_REQUEST_EVENT_TYPE CI_COMMIT_BRANCH; do
      if [[ -n "${!x:-}" ]]; then A_REASON="local run carries $x"; return 0; fi
    done
    A_TC=local; return 0
  fi
  if [[ "$GITLAB_CI" != true ]]; then A_REASON="GITLAB_CI is not true"; return 0; fi
  if [[ ! "${CI_COMMIT_SHA:-}" =~ ^[0-9a-f]{40}$ ]]; then A_REASON="CI_COMMIT_SHA missing or malformed"; return 0; fi
  if [[ "$CI_COMMIT_SHA" != "$head" ]]; then A_REASON="checked-out HEAD differs from CI_COMMIT_SHA"; return 0; fi
  # Tag pipelines never run (the workflow admits no tag, RC or final); a tag variable is refused outright.
  if [[ -n "${CI_COMMIT_TAG:-}" ]]; then A_REASON="tag pipelines are not a supported profile"; return 0; fi
  case "${CI_PIPELINE_SOURCE:-}" in
    push)
      for x in CI_MERGE_REQUEST_IID CI_MERGE_REQUEST_EVENT_TYPE CI_MERGE_REQUEST_SOURCE_BRANCH_SHA CI_MERGE_REQUEST_TARGET_BRANCH_SHA; do
        if [[ -n "${!x:-}" ]]; then A_REASON="push pipeline carries $x"; return 0; fi
      done
      b="${CI_COMMIT_BRANCH:-}"
      if [[ -z "$b" || -z "${CI_DEFAULT_BRANCH:-}" ]]; then A_REASON="push pipeline without CI_COMMIT_BRANCH or CI_DEFAULT_BRANCH"; return 0; fi
      if [[ "${CI_COMMIT_REF_NAME:-}" != "$b" ]]; then A_REASON="CI_COMMIT_REF_NAME differs from CI_COMMIT_BRANCH"; return 0; fi
      A_TC=branch-push ;;
    schedule)
      for x in CI_MERGE_REQUEST_IID CI_MERGE_REQUEST_EVENT_TYPE CI_MERGE_REQUEST_SOURCE_BRANCH_SHA CI_MERGE_REQUEST_TARGET_BRANCH_SHA; do
        if [[ -n "${!x:-}" ]]; then A_REASON="schedule pipeline carries $x"; return 0; fi
      done
      b="${CI_COMMIT_BRANCH:-}"
      # G004 E-1: the protected identity-qualification branch is the one other accepted schedule.
      if [[ -n "$b" && "$b" == identity-qualification && "${CI_COMMIT_REF_PROTECTED:-}" == true ]]; then :
      elif [[ -z "$b" || -z "${CI_DEFAULT_BRANCH:-}" || "$b" != "$CI_DEFAULT_BRANCH" ]]; then A_REASON="schedule outside the default branch"; return 0; fi
      if [[ "${CI_COMMIT_REF_NAME:-}" != "$b" ]]; then A_REASON="CI_COMMIT_REF_NAME differs from CI_COMMIT_BRANCH"; return 0; fi
      A_TC=branch-schedule ;;
    web)
      for x in CI_MERGE_REQUEST_IID CI_MERGE_REQUEST_EVENT_TYPE CI_MERGE_REQUEST_SOURCE_BRANCH_SHA CI_MERGE_REQUEST_TARGET_BRANCH_SHA; do
        if [[ -n "${!x:-}" ]]; then A_REASON="web pipeline carries $x"; return 0; fi
      done
      b="${CI_COMMIT_BRANCH:-}"
      if [[ -z "$b" ]]; then A_REASON="web pipeline without CI_COMMIT_BRANCH"; return 0; fi
      if [[ "${CI_COMMIT_REF_NAME:-}" != "$b" ]]; then A_REASON="CI_COMMIT_REF_NAME differs from CI_COMMIT_BRANCH"; return 0; fi
      A_TC=branch-web ;;
    api)
      # A branch pipeline started through the pipelines API (G048 supplement): attributed exactly like a web
      # pipeline, recorded as its own source and tested content.
      for x in CI_MERGE_REQUEST_IID CI_MERGE_REQUEST_EVENT_TYPE CI_MERGE_REQUEST_SOURCE_BRANCH_SHA CI_MERGE_REQUEST_TARGET_BRANCH_SHA; do
        if [[ -n "${!x:-}" ]]; then A_REASON="api pipeline carries $x"; return 0; fi
      done
      b="${CI_COMMIT_BRANCH:-}"
      if [[ -z "$b" ]]; then A_REASON="api pipeline without CI_COMMIT_BRANCH"; return 0; fi
      if [[ "${CI_COMMIT_REF_NAME:-}" != "$b" ]]; then A_REASON="CI_COMMIT_REF_NAME differs from CI_COMMIT_BRANCH"; return 0; fi
      A_TC=branch-api ;;
    merge_request_event)
      if [[ ! "${CI_MERGE_REQUEST_IID:-}" =~ ^[1-9][0-9]{0,18}$ ]]; then A_REASON="merge request IID missing"; return 0; fi
      if [[ -z "${CI_MERGE_REQUEST_SOURCE_BRANCH_NAME:-}" || -z "${CI_MERGE_REQUEST_TARGET_BRANCH_NAME:-}" ]]; then
        A_REASON="merge request branch names missing"; return 0
      fi
      if [[ -n "${CI_COMMIT_BRANCH:-}" ]]; then A_REASON="CI_COMMIT_BRANCH set in a merge request pipeline"; return 0; fi
      # ONEOFF RC qualification: every release/ source branch is a promotion candidate, verified or refused.
      if [[ "$CI_MERGE_REQUEST_SOURCE_BRANCH_NAME" == release/* ]]; then classify_promotion "$head"; return 0; fi
      case "${CI_MERGE_REQUEST_EVENT_TYPE:-}" in
        detached)
          if [[ -n "${CI_MERGE_REQUEST_SOURCE_BRANCH_SHA:-}" || -n "${CI_MERGE_REQUEST_TARGET_BRANCH_SHA:-}" ]]; then
            A_REASON="detached pipeline carries merged-result SHAs"; return 0
          fi
          A_TC=mr-source ;;
        merged_result)
          src="${CI_MERGE_REQUEST_SOURCE_BRANCH_SHA:-}" tgt="${CI_MERGE_REQUEST_TARGET_BRANCH_SHA:-}"
          if [[ ! "$src" =~ ^[0-9a-f]{40}$ || ! "$tgt" =~ ^[0-9a-f]{40}$ ]]; then
            A_REASON="merged-result pipeline without source/target SHAs"; return 0
          fi
          if [[ "$head" == "$src" || "$head" == "$tgt" ]]; then A_REASON="merged-result HEAD equals a branch head"; return 0; fi
          # The raw commit object lists both parents even at the shallow boundary; only header lines count.
          parents="$(g cat-file commit "$head" | sed -n '1,/^$/ s/^parent //p' | LC_ALL=C sort | tr '\n' ' ')" || {
            A_REASON="cannot read the HEAD commit object"; return 0; }
          want="$(printf '%s\n' "$src" "$tgt" | LC_ALL=C sort | tr '\n' ' ')"
          if [[ "$parents" != "$want" ]]; then A_REASON="merged-result parents are not exactly the source and target SHAs"; return 0; fi
          A_TC=mr-merged-result ;;
        merge_train) A_REASON="merge_train pipelines are not a supported profile"; return 0 ;;
        *) A_REASON="merge request event type missing or unknown"; return 0 ;;
      esac ;;
    *) A_REASON="unsupported pipeline source"; return 0 ;;
  esac
  return 0
}

json_or_null() { if [[ -n "$1" ]]; then ci_json_str "$1"; else printf 'null'; fi; }

# attribute: writes raw/attribution.json (allow-listed metadata only, never an environment dump) and
# prints one line, "TESTED_CONTENT RC_TAG TAG_OBJECT STAGING_SHA" (the last three "none" outside an RC promotion);
# returns 16 when the run is not an accepted profile.
attribute() {
  local head tree="" commit="" dv cv x sep=""
  head="$(g rev-parse --verify 'HEAD^{commit}' 2>/dev/null)" || head=""
  if [[ -n "$head" ]]; then
    classify "$head"
    tree="$(g rev-parse --verify 'HEAD^{tree}' 2>/dev/null)" || tree=""
    commit="$(g log -1 --format='%H %an <%ae> %cI' 2>/dev/null)" || commit=""
  else
    A_TC="" A_REASON="no checked-out HEAD"
  fi
  dv="$(docker version --format '{{.Server.Version}}' 2>/dev/null)" || dv=""
  cv="$(docker compose version --short 2>/dev/null)" || cv=""
  {
    printf '{\n  "format": "platform-validation-attribution/1",\n'
    if [[ -n "$A_TC" ]]; then
      printf '  "result": "accepted",\n  "tested_content": "%s",\n  "reason": null,\n' "$A_TC"
    else
      printf '  "result": "rejected",\n  "tested_content": null,\n  "reason": %s,\n' "$(ci_json_str "$A_REASON")"
    fi
    printf '  "head": %s,\n  "tree": %s,\n  "commit": %s,\n' "$(json_or_null "$head")" "$(json_or_null "$tree")" "$(json_or_null "$commit")"
    if [[ "$A_TC" == mr-rc-promotion ]]; then
      printf '  "release": {"tag": "%s", "tag_object": "%s", "peeled": "%s", "staging_sha": "%s"},\n' "$A_TAG" "$A_TAGOBJ" "$head" "$A_STAGING"
    else
      printf '  "release": null,\n'
    fi
    printf '  "ci": {'
    for x in "${ATTR_VARS[@]}"; do
      if [[ -n "${!x:-}" ]]; then printf '%s\n    "%s": %s' "$sep" "$x" "$(ci_json_str "${!x}")"; sep=","; fi
    done
    printf '\n  },\n  "docker_version": %s,\n  "compose_version": %s,\n  "started_utc": "%s"\n}\n' \
      "$(json_or_null "$dv")" "$(json_or_null "$cv")" "$STARTED_UTC"
  } >"$EV/raw/attribution.json" || return 16
  if [[ -z "$A_TC" ]]; then printf 'validate: attribution rejected: %s\n' "$A_REASON" >&2; return 16; fi
  printf '%s %s %s %s\n' "$A_TC" "$A_TAG" "$A_TAGOBJ" "$A_STAGING"
}

# selection_digest BASE AREAS PIPELINE SUITES: the G051 selection digest of the checked-out candidate, SHA-256
# over head, tree, the map, selector and registry blob IDs at HEAD, the base, the areas, the pipeline profile
# and its suites, each line newline-terminated (the selector computes the same lines).
selection_digest() {
  local head tree map sel reg d
  head="$(g rev-parse --verify 'HEAD^{commit}' 2>/dev/null)" || return 1
  tree="$(g rev-parse --verify 'HEAD^{tree}' 2>/dev/null)" || return 1
  map="$(g rev-parse --verify 'HEAD:infra/ci/profile-paths.tsv' 2>/dev/null)" || return 1
  sel="$(g rev-parse --verify 'HEAD:infra/ci/select-profile.sh' 2>/dev/null)" || return 1
  reg="$(g rev-parse --verify 'HEAD:infra/validation/required-checks.tsv' 2>/dev/null)" || return 1
  d="$(printf 'head=%s\ntree=%s\nmap=%s\nselector=%s\nregistry=%s\nbase=%s\nareas=%s\npipeline=%s\nsuites=%s\n' \
    "$head" "$tree" "$map" "$sel" "$reg" "$1" "$2" "$3" "$4" | sha256sum | cut -c1-64)" || return 1
  if [[ ! "$d" =~ ^[0-9a-f]{64}$ ]]; then return 1; fi
  printf '%s\n' "$d"
}

# select_profile TESTED_CONTENT: writes state/selection.env (the nine G051 selection keys) and
# raw/selection.tsv. The CI default (no --profile) is full; it and the staging job are written here (no areas,
# the profile's suites, complete), every other selection by infra/ci/select-profile.sh.
select_profile() {
  local tc="$1" sel="$EV/state/selection.env" p r suites digest
  local -a opt=()
  if [[ "$JOB_KIND" == full || "$JOB_KIND" == staging ]]; then
    : >"$EV/raw/selection.tsv" || return 16
    if [[ "$JOB_KIND" == full ]]; then p=full r=default-full; else p=staging r=identity-staging-job; fi
    suites="$(ci_profile_suites "$p")" || return 16
    if [[ "$p" == full ]] && ! ci_controls_required "$REPO_R" none HEAD; then
      suites="$(ci_suites_minus "$suites" ci-controls)" || return 16
    fi
    digest="$(selection_digest none none "$p" "$suites")" || { say "preflight: the selection digest cannot be computed"; return 16; }
    (set -C; printf 'PIPELINE_PROFILE=%s\nEXECUTED_PROFILE=%s\nSELECTION_REASON=%s\nCOMPARISON_BASE=none\nAREAS=none\nPIPELINE_SUITES=%s\nEXECUTED_SUITES=%s\nCOVERAGE=complete\nSELECTION_DIGEST=%s\n' \
      "$p" "$p" "$r" "$suites" "$suites" "$digest" >"$sel") || return 16
    return 0
  fi
  if [[ "$JOB_KIND" == local ]]; then
    opt=(--requested "$PROFILE_ARG")
    if [[ -n "$BASE_ARG" ]]; then opt+=(--base "$BASE_ARG"); fi
    if [[ -n "$COMPARE_ARG" ]]; then opt+=(--compare "$COMPARE_ARG"); fi
  fi
  bash "$REPO_R/infra/ci/select-profile.sh" --job "$JOB_KIND" "${opt[@]}" --tested-content "$tc" \
    --map "$REPO_R/infra/ci/profile-paths.tsv" --out "$sel" --tsv "$EV/raw/selection.tsv" >>"$EV/raw/selection.log" 2>&1 || {
    say "preflight: profile selection refused (see selection.log)"; return 16; }
  return 0
}

# suite_list_ok LIST: every suite is a known run.sh suite (CI_SUITE_ORDER), once, in canonical order.
suite_list_ok() {
  local s i last=-1
  local -a order=() names=()
  local -A pos=()
  if [[ ! "$1" =~ $RE_PSUITES ]]; then return 1; fi
  read -r -a order <<<"$CI_SUITE_ORDER"
  for i in "${!order[@]}"; do pos[${order[$i]}]="$i"; done
  read -r -a names <<<"$1"
  for s in "${names[@]}"; do
    if [[ -z "${pos[$s]+x}" ]]; then return 1; fi
    if (( pos[$s] <= last )); then return 1; fi
    last="${pos[$s]}"
  done
  return 0
}

# suites_minus LIST MINUS: LIST without the suites of MINUS, order kept (empty when nothing remains).
suites_minus() {
  local s out="" sep=""
  local -a names=()
  read -r -a names <<<"$1"
  for s in "${names[@]}"; do
    if [[ " $2 " == *" $s "* ]]; then continue; fi
    out+="$sep$s"; sep=" "
  done
  printf '%s\n' "$out"
}

# partition_ok: the imported selection is one consistent partition (G051 L1, Plan §3.7): suites known and in
# canonical order, the pipeline suites those of the pipeline profile, areas sorted and unique, and the executed
# profile, suites and coverage re-derived from PIPELINE_SUITES and JOB_KIND. Sets PARTITION (fast, remainder,
# complete or none); 1 on any inconsistency. TESTED_CONTENT (the argument) forces the CI meta-tests for an RC promotion.
partition_ok() {
  local tc="${1:-}" s p want_exec want_suites want_cov part profile_suites
  local -a names=()
  PARTITION=""
  if suite_list_ok "$PIPELINE_SUITES"; then :; else return 1; fi
  if [[ "$EXECUTED_SUITES" != none ]]; then if suite_list_ok "$EXECUTED_SUITES"; then :; else return 1; fi; fi
  if [[ "$AREAS" != none ]]; then
    if printf '%s\n' "${AREAS//,/$'\n'}" | LC_ALL=C sort -C -u; then :; else return 1; fi
  fi
  p="$PIPELINE_PROFILE"
  case "$p" in
    fast|staging) if [[ "$PIPELINE_SUITES" != "$(ci_profile_suites "$p")" ]]; then return 1; fi ;;
    integration|full)
      profile_suites="$(ci_profile_suites "$p")" || return 1
      if [[ "$p" == full ]] && ! ci_controls_required "$REPO_R" "$COMPARISON_BASE" HEAD "$tc"; then
        profile_suites="$(suites_minus "$profile_suites" ci-controls)" || return 1
      fi
      if [[ "$(suites_minus "$PIPELINE_SUITES" "feature-web feature-core")" != "$profile_suites" ]]; then return 1; fi
      # Every registered web area present at HEAD has its feature suite (and only then).
      if [[ " $PIPELINE_SUITES " == *" feature-web "* && ",$AREAS" != *",web:"* ]]; then return 1; fi
      if [[ " $PIPELINE_SUITES " != *" feature-web "* && ",$AREAS" == *",web:"* ]]; then return 1; fi ;;
    affected)
      read -r -a names <<<"$PIPELINE_SUITES"
      for s in "${names[@]}"; do if [[ "$AFFECTED_SUITES" != *" $s "* ]]; then return 1; fi; done
      if [[ " $PIPELINE_SUITES " != *" lint-base "* || " $PIPELINE_SUITES " != *" provenance "* ]]; then return 1; fi ;;
    *) return 1 ;;
  esac
  if [[ " $PIPELINE_SUITES " == *" feature-web "* && ",$AREAS" != *",web:"* ]]; then return 1; fi
  if [[ " $PIPELINE_SUITES " == *" feature-core "* && ",$AREAS" != *",core:"* ]]; then return 1; fi
  # The staging profile belongs to the identity-staging job only; the CI default is full.
  if [[ "$JOB_KIND" == staging && "$p" != staging ]] || [[ "$JOB_KIND" != staging && "$p" == staging ]]; then return 1; fi
  if [[ "$JOB_KIND" == full && "$p" != full ]]; then return 1; fi
  case "$JOB_KIND" in
    local|full|staging) want_exec="$p" want_suites="$PIPELINE_SUITES" want_cov=complete part=complete ;;
    fast)
      case "$p" in
        affected) want_exec=affected want_suites="$PIPELINE_SUITES" want_cov=complete part=complete ;;
        fast) want_exec=fast want_suites="$CI_SUITES_fast" want_cov=complete part=complete ;;
        *) want_exec=fast want_suites="$CI_SUITES_fast" want_cov=partial part=fast ;;
      esac ;;
    gate)
      case "$p" in
        fast|affected) want_exec=none want_suites=none want_cov=none part=none ;;
        *) want_exec="$p" want_suites="$(suites_minus "$PIPELINE_SUITES" "$CI_SUITES_fast")" want_cov=partial part=remainder ;;
      esac ;;
    *) return 1 ;;
  esac
  if [[ "$EXECUTED_PROFILE" != "$want_exec" || "$EXECUTED_SUITES" != "$want_suites" || "$COVERAGE" != "$want_cov" ]]; then return 1; fi
  PARTITION="$part"
  return 0
}

# bootstrap_pre (hosted adapter context only; G051 Plan §3.8.3, Codex Q1): the identity of the context image the
# adapter built before this job's harness, in state/bootstrap.tsv (a header line and one row: ref, image_id,
# dockerfile, dockerfile_sha256, bases, platform). bases are the Dockerfile's two digest-pinned FROM references,
# sorted and comma-joined; platform is the daemon's, which the image's must equal. A field that cannot be
# established is "unknown" (the ledger then counts the bootstrap row as unknown: AC5 UNMET), never omitted.
# The adapter builds with -q, so no step count is claimed. Returns 16 only when the record cannot be written.
bootstrap_pre() {
  local ref="platform-ci-context:$CI_JOB_ID" f="$EV/state/bootstrap.tsv" id sha line plat iplat bases=unknown
  local -a from=()
  id="$(docker image inspect --format '{{.Id}}' "$ref" 2>/dev/null)" || id=""
  if [[ ! "$id" =~ ^sha256:[0-9a-f]{64}$ ]]; then id=unknown; fi
  sha="$(sha256sum <"$REPO_R/$CTX_DOCKERFILE" 2>/dev/null | cut -c1-64)" || sha=""
  if [[ ! "$sha" =~ ^[0-9a-f]{64}$ ]]; then sha=unknown; fi
  if [[ -f "$REPO_R/$CTX_DOCKERFILE" && ! -L "$REPO_R/$CTX_DOCKERFILE" ]]; then
    while IFS= read -r line || [[ -n "$line" ]]; do
      if [[ "$line" =~ ^[[:space:]]*[Ff][Rr][Oo][Mm][[:space:]] ]]; then
        if [[ "$line" =~ ^FROM[[:space:]]+([a-z0-9][a-z0-9._/-]*:[A-Za-z0-9._-]+@sha256:[0-9a-f]{64})([[:space:]]+AS[[:space:]]+[a-z][a-z0-9_-]*)?[[:space:]]*$ ]]; then
          from+=("${BASH_REMATCH[1]}")
        else
          from+=(unknown)
        fi
      fi
    done <"$REPO_R/$CTX_DOCKERFILE"
  fi
  if (( ${#from[@]} == 2 )) && [[ " ${from[*]} " != *" unknown "* ]]; then
    bases="$(printf '%s\n' "${from[@]}" | LC_ALL=C sort | tr '\n' ',')" || bases=""
    bases="${bases%,}"
    if [[ ! "$bases" =~ ^[^,[:space:]]+,[^,[:space:]]+$ ]]; then bases=unknown; fi
  fi
  plat="$(docker version --format '{{.Server.Os}}/{{.Server.Arch}}' 2>/dev/null)" || plat=""
  iplat="$(docker image inspect --format '{{.Os}}/{{.Architecture}}{{with .Variant}}/{{.}}{{end}}' "$ref" 2>/dev/null)" || iplat=""
  if [[ ! "$plat" =~ ^[a-z0-9]+/[a-z0-9_]+$ || "$iplat" != "$plat" ]]; then plat=unknown; fi
  (umask 077 && set -C && printf 'ref\timage_id\tdockerfile\tdockerfile_sha256\tbases\tplatform\n%s\t%s\t%s\t%s\t%s\t%s\n' \
    "$ref" "$id" "$CTX_DOCKERFILE" "$sha" "$bases" "$plat" >"$f") || return 16
  say "preflight: bootstrap context image $id recorded (dockerfile $sha, platform $plat)"
  return 0
}

do_preflight() {
  local out="$EV/raw/runner-preflight.tsv" rc=0 att tc rtag robj rstg head tree digest budget prof min sdig
  local PIPELINE_PROFILE="" EXECUTED_PROFILE="" SELECTION_REASON="" COMPARISON_BASE=""
  local AREAS="" PIPELINE_SUITES="" EXECUTED_SUITES="" COVERAGE="" SELECTION_DIGEST="" PARTITION=""
  # The start inventory comes first, so every preflight container is provably new.
  ci_docker_start_inventory "$REG/start-inventory.ids" || { say "preflight: Docker start inventory failed"; return 16; }
  att="$(attribute)" || return 16
  read -r tc rtag robj rstg <<<"$att"
  select_profile "$tc" || return 16
  state_import "$EV/state/selection.env" "${SEL_SPECS[@]}" || { say "preflight: the profile selection is malformed"; return 16; }
  # The gate executes nothing only when the pipeline profile is fast or affected (validate.sh re-asserts it).
  if [[ "$EXECUTED_PROFILE" == none && ( "$PIPELINE_PROFILE" =~ ^(integration|full|staging)$ || "$JOB_KIND" != gate ) ]]; then
    say "preflight: an empty gate outside a fast or affected pipeline"; return 16
  fi
  if partition_ok "$tc"; then :; else say "preflight: the selection is not one consistent partition"; return 16; fi
  prof="$EXECUTED_PROFILE"
  if [[ "$prof" == none ]]; then prof=fast; fi
  if [[ "$prof" == staging ]]; then
    # runner-preflight.sh takes fast, affected, integration or full: the staging profile's capacity equals fast's
    # (infra/ci/lib.sh); the staging job's own timeout floor was checked at location.
    bash "$REPO_R/infra/ci/runner-preflight.sh" --registry "$REG" --out "$out" --profile fast --job fast >"$EV/raw/runner-preflight.log" 2>&1 || rc=$?
  else
    bash "$REPO_R/infra/ci/runner-preflight.sh" --registry "$REG" --out "$out" --profile "$prof" --job "$JOB_KIND" >"$EV/raw/runner-preflight.log" 2>&1 || rc=$?
  fi
  if preflight_rows_ok "$out"; then :; else say "preflight: runner preflight failed or incomplete (see runner-preflight.tsv)"; return 16; fi
  if (( rc != 0 )); then say "preflight: runner preflight exited $rc"; return 16; fi
  docker pull "$CI_GITLEAKS_IMAGE" >>"$EV/raw/setup.log" 2>&1 || { say "preflight: cannot pull the scanner image"; return 16; }
  docker pull "$CI_YQ_IMAGE" >>"$EV/raw/setup.log" 2>&1 || { say "preflight: cannot pull the yq image"; return 16; }
  if ci_hosted_context; then bootstrap_pre || { say "preflight: the bootstrap record cannot be written"; return 16; }; fi
  g status --porcelain=v1 -z --untracked-files=all --ignored >"$EV/state/status.z" || return 16
  if [[ -s "$EV/state/status.z" ]]; then say "preflight: the checkout is not clean (status with ignored files is not empty)"; return 16; fi
  digest="$(sha256sum <"$EV/state/status.z" | cut -c1-64)" || return 16
  head="$(g rev-parse --verify 'HEAD^{commit}')" || return 16
  tree="$(g rev-parse --verify 'HEAD^{tree}')" || return 16
  # The selection belongs to this clean candidate: its digest is recomputed from HEAD (G051 §3.7).
  sdig="$(selection_digest "$COMPARISON_BASE" "$AREAS" "$PIPELINE_PROFILE" "$PIPELINE_SUITES")" || sdig=""
  if [[ "$sdig" != "$SELECTION_DIGEST" ]]; then say "preflight: the selection digest does not match the candidate"; return 16; fi
  budget="$(ci_profile_budget_seconds "$prof")" || {
    say "preflight: CI_JOB_TIMEOUT, RUNNER_SCRIPT_TIMEOUT or CI_JOB_STARTED_AT is missing or malformed"; return 16; }
  min="CI_MINBUDGET_$prof"
  if [[ ! "$budget" =~ ^-?[0-9]+$ ]] || (( budget < ${!min} )); then say "preflight: harness budget ${budget}s is below ${!min}s ($prof)"; return 16; fi
  (set -C; printf 'HEAD_SHA=%s\nTREE_SHA=%s\nSTATUS_DIGEST=%s\nBUDGET_S=%s\nTESTED_CONTENT=%s\nPIPELINE_PROFILE=%s\nEXECUTED_PROFILE=%s\nSELECTION_REASON=%s\nCOMPARISON_BASE=%s\nAREAS=%s\nPIPELINE_SUITES=%s\nEXECUTED_SUITES=%s\nCOVERAGE=%s\nSELECTION_DIGEST=%s\nRELEASE_TAG=%s\nRELEASE_TAG_OBJECT=%s\nRELEASE_STAGING_SHA=%s\n' \
    "$head" "$tree" "$digest" "$budget" "$tc" "$PIPELINE_PROFILE" "$EXECUTED_PROFILE" "$SELECTION_REASON" "$COMPARISON_BASE" \
    "$AREAS" "$PIPELINE_SUITES" "$EXECUTED_SUITES" "$COVERAGE" "$SELECTION_DIGEST" "$rtag" "$robj" "$rstg" >"$EV/state/p1.env") || return 16
  return 0
}

# --- P2 harness: launched, supervised and stopped by the parent ----------------------------------------
group_ready() { # PGID: the setsid child becomes a group leader asynchronously; wait (bounded) for it
  local i
  for ((i = 0; i < 50; i++)); do
    if kill -0 -- "-$1" 2>/dev/null; then return 0; fi
    if kill -0 "$1" 2>/dev/null; then :; else return 1; fi
    sleep 0.1
  done
  return 1
}

stop_harness_group() { # TERM to the whole group, a bounded grace, then KILL; reap; report whether it is gone
  local i
  if [[ -z "$HPID" ]]; then return 0; fi
  if group_ready "$HPID"; then :; else kill -TERM "$HPID" 2>/dev/null || true; fi
  kill -TERM -- "-$HPID" 2>/dev/null || true
  for ((i = 0; i < STOP_GRACE_DS; i++)); do
    if kill -0 -- "-$HPID" 2>/dev/null; then sleep 0.1; else break; fi
  done
  if kill -0 -- "-$HPID" 2>/dev/null; then
    say "harness: the process group outlived the grace period; sending KILL"
    kill -KILL -- "-$HPID" 2>/dev/null || true
  fi
  wait "$HPID" 2>/dev/null || true
  for ((i = 0; i < 50; i++)); do
    if kill -0 -- "-$HPID" 2>/dev/null; then sleep 0.1; else break; fi
  done
  if kill -0 -- "-$HPID" 2>/dev/null; then say "harness: process group still present after KILL"; else say "harness: process group gone"; fi
  return 0
}

stop_watchdog() {
  if [[ -z "$WPID" ]]; then return 0; fi
  if group_ready "$WPID"; then kill -TERM -- "-$WPID" 2>/dev/null || true; else kill -TERM "$WPID" 2>/dev/null || true; fi
  wait "$WPID" 2>/dev/null || true
  WPID=""
}

harness_group_lingers() { # after a normal exit: give the group a moment (tee flushing), then report
  local i
  for ((i = 0; i < 30; i++)); do
    if kill -0 -- "-$HPID" 2>/dev/null; then sleep 0.1; else return 1; fi
  done
  return 0
}

# The deadline watchdog (G051 cycle 4), run as `setsid bash -c "$WATCHDOG_SH" watchdog BUDGET_S PID PID_START`.
# It first closes every inherited descriptor above stdio, so it never holds a caller's lock or pipe. Its only child
# is the timer `sleep BUDGET_S`; once a second it checks that its parent is still the same live process (PID and
# start time) and ends, stopping the timer, when it is not. When the timer ends (the deadline) it sends USR1 only
# to that verified parent. A KILLed validate.sh therefore never leaves it behind for longer than about a second, and
# a reused PID is never signalled.
# shellcheck disable=SC2016  # expanded by the watchdog's own shell
WATCHDOG_SH='
for f in /proc/$$/fd/*; do n="${f##*/}"; if [[ "$n" =~ ^[0-9]+$ ]] && (( n > 2 )); then eval "exec $n<&-"; fi; done
alive() { local s a; read -r s 2>/dev/null <"/proc/$1/stat" || return 1; read -r -a a <<<"${s##*) }"; [[ "${a[0]}" != Z && "${a[19]}" == "$2" ]]; }
exec {t}< <(exec sleep "$1")
tp=$!
while alive "$2" "$3"; do
  read -r -t 1 -u "$t" _
  if (( $? > 128 )); then continue; fi
  if alive "$2" "$3"; then kill -USR1 "$2"; fi
  exit 0
done
kill "$tp" 2>/dev/null
exit 0
'

# G066: print "N/TOTAL ICON STATUS ID" (ICON U+2705 for PASS, U+274C for FAIL/UNMET; no "validate:" prefix, fixed
# printf format, to avoid mobile wrapping) for each new complete
# ID<TAB>PASS|FAIL|UNMET line of the progress file whose ID is a required ID (PG_REQ) not yet seen, so each required
# check counts once and N <= TOTAL; only that matched required ID is printed, never a message, and INFO, non-required
# IDs, repeats and anything else are ignored. An unterminated tail is carried to the next call.
declare -A PG_REQ=() PG_SEEN=()
PG_FD="" PG_CARRY="" PG_N=0 PG_TOTAL="?" PG_RE=$'^([A-Z][A-Z0-9-]*)\t(PASS|FAIL|UNMET)$'
progress_drain() {
  local line="" icon=""
  if [[ -z "$PG_FD" ]]; then return 0; fi
  while IFS= read -r -u "$PG_FD" line; do
    line="$PG_CARRY$line"; PG_CARRY=""
    if [[ "$line" =~ $PG_RE && -n "${PG_REQ[${BASH_REMATCH[1]}]+x}" && -z "${PG_SEEN[${BASH_REMATCH[1]}]+x}" ]]; then
      icon=$'\xe2\x9d\x8c'; if [[ "${BASH_REMATCH[2]}" == PASS ]]; then icon=$'\xe2\x9c\x85'; fi
      PG_SEEN[${BASH_REMATCH[1]}]=1; PG_N=$((PG_N + 1))
      printf '%s/%s %s %s %s\n' "$PG_N" "$PG_TOTAL" "$icon" "${BASH_REMATCH[2]}" "${BASH_REMATCH[1]}"
    fi
  done
  PG_CARRY+="$line"
}

run_harness() {
  local rc suites=() pst hs ws started pg_ids pg_id
  cd -- "$REPO_R"
  # The watchdog verifies its parent by start time; without it there is no safe deadline, so no harness.
  if pst="$(proc_start_time "$$")"; then :; else
    ST[internal]=19; say "harness: not started (the start time of validate.sh is unreadable, so the watchdog cannot verify it)"; return 0
  fi
  # A crash from here on leaves a marker that names this stage; recovery then refuses to guess (local runs).
  if marker_stage launching; then :; else
    ST[internal]=19; say "harness: not started (the run marker could not record the launch)"; return 0
  fi
  # G051: the harness runs exactly the executed suites of this job's partition, with the selected areas.
  read -r -a suites <<<"$EXECUTED_SUITES"
  if [[ "$EXECUTED_PROFILE" == staging ]]; then
    # Only the API key is a strict secret; the client ID and issuer are public metadata (Plan r5 §8.3).
    sed -n 's/^IDENTITY_STAGING_WORKOS_API_KEY=//p' "$IDENTITY_STAGING_ENV_FILE" | awk 'length($0) >= 8' >>"$REG/secrets"
  fi
  # The hosted context's bootstrap record joins the harness's build ledger (G051 §3.8.3); nowhere else.
  if (( CI_MODE )) && ci_hosted_context && [[ -f "$EV/state/bootstrap.tsv" && ! -L "$EV/state/bootstrap.tsv" ]]; then
    export PLATFORM_LEDGER_BOOTSTRAP_FILE="$EV/state/bootstrap.tsv"
  fi
  say "harness: starting infra/validation/run.sh ${suites[*]} (profile $EXECUTED_PROFILE; deadline ${BUDGET_S}s)"
  # G066: record() appends each ID<TAB>STATUS to this private file (never scanned or staged; do_scrub removes it).
  # TOTAL is the required checks of the executed suites, as completeness counts them (loaded once into PG_REQ).
  export PLATFORM_PROGRESS_FILE="$EV/state/progress.tsv"
  : >"$PLATFORM_PROGRESS_FILE" || PLATFORM_PROGRESS_FILE=""
  if pg_ids="$(required_ids "$EXECUTED_SUITES")"; then
    while IFS= read -r pg_id; do if [[ -n "$pg_id" ]]; then PG_REQ[$pg_id]=1; fi; done <<<"$pg_ids"
    PG_TOTAL=${#PG_REQ[@]}
  fi
  started=$SECONDS
  setsid bash infra/validation/run.sh --base HEAD --allow-file infra/ci/unchanged.allow --owned-registry "$REG/owned.tsv" --secrets-file "$REG/secrets" --report-dir "$EV/raw/report" "${suites[@]}" --areas "$AREAS" >"$EV/raw/harness-console.log" 2>&1 &
  HPID=$!
  group_ready "$HPID" || true
  printf '%s\n' "$HPID" >"$EV/state/harness.pgid"
  setsid bash -c "$WATCHDOG_SH" watchdog "$BUDGET_S" "$$" "$pst" </dev/null >/dev/null 2>&1 &
  WPID=$!
  group_ready "$WPID" || true
  hs="$(proc_start_time "$HPID")" || hs=none
  ws="$(proc_start_time "$WPID")" || ws=none
  if marker_stage launched harness_pgid "$HPID" harness_start "$hs" watchdog_pid "$WPID" watchdog_start "$ws"; then :; else
    say "recovery: the run marker could not record the harness and watchdog; a crash now leaves it at the launching stage"
  fi
  rc=0
  # G066: poll every 0.5 s while the harness lives instead of blocking, printing progress; then the unchanged wait.
  if [[ -n "$PLATFORM_PROGRESS_FILE" ]]; then { exec {PG_FD}<"$PLATFORM_PROGRESS_FILE"; } 2>/dev/null || PG_FD=""; fi
  while (( DEADLINE == 0 )) && kill -0 "$HPID" 2>/dev/null; do progress_drain; sleep 0.5 || :; done
  if (( DEADLINE == 0 )); then
    set +e
    wait "$HPID"
    rc=$?
    set -e
  fi
  if (( DEADLINE )); then
    say "harness: internal deadline reached; stopping the harness process group"
    stop_harness_group
    rc=124; HARNESS_STOP=deadline
  fi
  progress_drain
  if [[ -n "$PG_FD" ]]; then exec {PG_FD}<&-; PG_FD=""; fi
  stop_watchdog
  PHASE_SECONDS[harness]=$((SECONDS - started))
  ST[harness]=$rc; HARNESS_DONE=1
  if harness_group_lingers; then
    STRAGGLERS=1
    say "info stragglers=1 (processes left in the harness group are stopped)"
    stop_harness_group
  fi
  HPID=""
  say "harness: exit status $rc"
  # Parent-side record of the harness result, written after the harness has finished.
  printf 'status=%s\nstop=%s\n' "$rc" "$HARNESS_STOP" >"$EV/state/harness.result"
}

on_signal() { # NAME NUMBER: an external interruption, recorded in the parent as such
  SIG_NAME="$1"
  ST[signal]=$((128 + $2))
  say "signal $1 received"
  stop_watchdog
  if [[ -n "$HPID" ]]; then
    stop_harness_group
    # A signal during the straggler check keeps the harness's own recorded status.
    if (( ! HARNESS_DONE )); then ST[harness]=$((128 + $2)); HARNESS_STOP=signal; fi
    HPID=""
  fi
  exit $((128 + $2))
}

# --- finalizer step bodies (isolated) -----------------------------------------------------------------
do_postcheck_repo_only() {
  local st n
  if [[ -z "$REPO_R" ]]; then return 13; fi
  st="$(g status --porcelain=v1 -z --untracked-files=all --ignored | tr '\0' '\n')" || return 13
  if [[ -n "$st" ]]; then
    n="$(printf '%s\n' "$st" | grep -c .)" || n="?"
    say "postcheck: $n paths differ from the checked-out commit"
    return 13
  fi
  return 0
}

do_postcheck() {
  local digest
  g status --porcelain=v1 -z --untracked-files=all --ignored >"$EV/state/post-status.z" || return 13
  if [[ -s "$EV/state/post-status.z" ]]; then
    say "postcheck: drift: $(tr -cd '\0' <"$EV/state/post-status.z" | wc -c) status entries differ from the checked-out commit"
    return 13
  fi
  if [[ -n "$HEAD_SHA" ]]; then
    if [[ "$(g rev-parse --verify 'HEAD^{commit}' 2>/dev/null)" != "$HEAD_SHA" ]]; then say "postcheck: HEAD changed"; return 13; fi
    if [[ "$(g rev-parse --verify 'HEAD^{tree}' 2>/dev/null)" != "$TREE_SHA" ]]; then say "postcheck: tree changed"; return 13; fi
    digest="$(sha256sum <"$EV/state/post-status.z" | cut -c1-64)" || return 13
    if [[ "$digest" != "$STATUS_DIGEST" ]]; then say "postcheck: status digest changed"; return 13; fi
  fi
  return 0
}

do_cleanup() {
  local rc=0 inv="$EV/raw/cleanup-inventory.tsv" n
  bash "$REPO_R/infra/ci/cleanup-owned.sh" --registry "$REG" --mode finalize --out "$inv" >"$EV/raw/cleanup.log" 2>&1 || rc=$?
  if [[ -L "$inv" || ! -f "$inv" ]]; then say "cleanup: no inventory was written (exit $rc)"; return 18; fi
  cp -- "$inv" "$REG/cleanup-inventory.tsv" || { say "cleanup: cannot copy the inventory into the registry"; return 18; }
  n="$(grep -c . "$inv")" || n=0
  if (( rc != 0 )); then say "cleanup: incomplete, refused or unproven (exit $rc; $n inventory rows)"; return 18; fi
  if awk -F'\t' 'NF != 4 || $3 != "removed" { found = 1 } END { exit found ? 0 : 1 }' "$inv"; then
    say "cleanup: the inventory has rows that are not removed"
    return 18
  fi
  say "cleanup: verified ($n resources removed)"
  return 0
}

# --- interrupted-run recovery and the storage report (G051 cleanup amendment; local runs) --------------
# A local run writes a marker, STATE_ROOT/g003-pending/<job>.tsv (registry, evidence, pid, pid_start, started, boot_id,
# repo and its launch stage; once launched, the harness group's and the watchdog's identities: G051 cycle 4; G064 D3
# moved it from $TMPDIR to persistent private storage and added boot_id and repo), after its recovery gate and before
# anything touches Docker. It holds no resource rows: the registry stays the only ownership record.
# finalize removes the marker when cleanup verified and the registry was removed; a kept registry (cleanup
# incomplete, refused or unproven) keeps it with a "finalized" row; a crash or KILL leaves it unfinalized.
# Markers of earlier versions in $TMPDIR/g003-pending are still read (same-boot rules), never written.
# CI jobs write no marker (each job has a fresh daemon and the after_script backstop).
proc_start_time() { # PID: the kernel start time of a live process (field 22 of /proc/PID/stat)
  local s f=()
  s="$(cat -- "/proc/$1/stat" 2>/dev/null)" || return 1
  read -r -a f <<<"${s##*) }"
  if [[ "${f[0]:-}" == Z || ! "${f[19]:-}" =~ ^[0-9]+$ ]]; then return 1; fi
  printf '%s' "${f[19]}"
}

priv_dir_ok() { # DIR: an existing directory, never a symlink, owned by this user, mode 700
  [[ ! -L "$1" && -d "$1" && -O "$1" && "$(stat -c %a -- "$1" 2>/dev/null)" == 700 ]]
}

pend_dir_ok() { # the marker directory: created mode 700 when absent; never a symlink; owned by this user; not tmpfs/ramfs
  if [[ -z "$PEND_DIR" ]]; then return 1; fi
  if [[ -e "$PEND_DIR" || -L "$PEND_DIR" ]]; then :; else mkdir -m 700 -- "$PEND_DIR" 2>/dev/null || return 1; fi
  if priv_dir_ok "$PEND_DIR"; then :; else return 1; fi
  if volatile_fs "$PEND_DIR"; then return 1; fi
  return 0
}

boot_id_now() { # this boot's identity: /proc/sys/kernel/random/boot_id as 32 lowercase hex (dashes removed); 1 when unreadable
  local b
  b="$(cat /proc/sys/kernel/random/boot_id 2>/dev/null)" || return 1
  b="${b//-/}"
  if [[ ! "$b" =~ ^[0-9a-f]{32}$ ]]; then return 1; fi
  printf '%s' "$b"
}

marker_write() { # parent, local runs, after the recovery gate passed
  local st bid
  st="$(proc_start_time "$$")" || return 1
  bid="$(boot_id_now)" || { say "recovery: the boot identity (/proc/sys/kernel/random/boot_id) is unreadable"; return 1; }
  if [[ ! "$REG_JOB" =~ ^local-[0-9a-f]{12}$ ]]; then return 1; fi
  PEND_FILE="$PEND_DIR/$REG_JOB.tsv"
  MARKER_BASE="$(printf 'registry\t%s\nevidence\t%s\npid\t%s\npid_start\t%s\nstarted\t%s\nboot_id\t%s\nrepo\t%s' "$REG" "$EV" "$$" "$st" "$STARTED_UTC" "$bid" "$REPO_R")"
  (umask 077 && set -C && printf '%s\nstage\tprelaunch\n' "$MARKER_BASE" >"$PEND_FILE") || { PEND_FILE=""; return 1; }
  return 0
}

# marker_stage STAGE [KEY VALUE]...: the run's launch stage (prelaunch, launching, launched) and, once launched, the
# harness group's and the watchdog's identities (PID or PGID and start time, or "none" when already ended), so that
# recovery can tell whether any of the run's processes survived. Rewritten by an atomic rename; a no-op without a
# marker (CI jobs).
marker_stage() {
  local t
  if [[ -z "$PEND_FILE" ]]; then return 0; fi
  t="$PEND_DIR/.$REG_JOB.tsv.new"
  rm -f -- "$t" || return 1
  (umask 077 && set -C && printf '%s\nstage\t%s\n' "$MARKER_BASE" "$1" >"$t") || return 1
  shift
  while (( $# >= 2 )); do printf '%s\t%s\n' "$1" "$2" >>"$t" || return 1; shift 2; done
  mv -f -- "$t" "$PEND_FILE" || return 1
}

marker_finish() { # parent, end of finalize
  if [[ -z "$PEND_FILE" || -L "$PEND_FILE" || ! -f "$PEND_FILE" ]]; then return 0; fi
  if [[ "$REGSTATE" == removed ]]; then
    rm -f -- "$PEND_FILE" || say "recovery: the run marker $PEND_FILE could not be removed"
    rmdir -- "$PEND_DIR" 2>/dev/null || true
  else
    printf 'finalized\tcleanup=%s\n' "${ST[cleanup]-none}" >>"$PEND_FILE" || true
    say "recovery: the registry was kept; the next local run must recover it before it starts (marker $PEND_FILE)"
  fi
  return 0
}

marker_get() { # FILE KEY: the value of the single KEY row
  awk -F'\t' -v k="$2" '$1 == k { n++; v = $2 } END { if (n == 1) print v; else exit 1 }' "$1"
}

proc_is_start() { # PID START: PID is the live process that had START ("none" never matches)
  [[ "$2" != none && "$(proc_start_time "$1" 2>/dev/null)" == "$2" ]]
}

pgrp_live() { # PGID: some live (not zombie) process is in process group PGID; read from /proc, nothing signalled
  local f s a
  for f in /proc/[0-9]*/stat; do
    read -r s 2>/dev/null <"$f" || continue
    read -r -a a <<<"${s##*) }"
    if [[ "${a[0]}" != Z && "${a[2]}" == "$1" ]]; then return 0; fi
  done
  return 1
}

# recovery_survivors MARKER JOB: 0 when nothing of the interrupted run's harness group or watchdog is alive. The
# watchdog ends within about a second of its parent, and a KILLed group needs a moment to be reaped, so this waits
# (bounded, RECOVERY_WAIT_DS) before anything counts. A live leader with the recorded start time is that run's
# process. Any live member of a recorded group whose leader is gone (or unrecorded) cannot be attributed, because
# its number may have been reused, so it blocks as ambiguous. A reused PID outside that group is ignored.
# Nothing is ever signalled: the operator resolves.
recovery_survivors() {
  local m="$1" job="$2" hp hs wp ws i
  hp="$(marker_get "$m" harness_pgid)" || hp=""; hs="$(marker_get "$m" harness_start)" || hs=""
  wp="$(marker_get "$m" watchdog_pid)" || wp=""; ws="$(marker_get "$m" watchdog_start)" || ws=""
  if [[ ! "$hp" =~ ^[1-9][0-9]{0,9}$ || ! "$wp" =~ ^[1-9][0-9]{0,9}$ || ! "$hs" =~ ^([0-9]+|none)$ || ! "$ws" =~ ^([0-9]+|none)$ ]]; then
    say "recovery: the harness and watchdog identities in the marker of run $job are malformed; operator resolution required (marker kept)"; return 1
  fi
  for ((i = 0; i < RECOVERY_WAIT_DS; i++)); do
    if proc_is_start "$hp" "$hs" || pgrp_live "$hp" || proc_is_start "$wp" "$ws" || pgrp_live "$wp"; then sleep 0.1; else return 0; fi
  done
  if proc_is_start "$hp" "$hs"; then
    say "recovery: the harness of run $job (process group $hp) is still running; wait for it or stop it (marker kept)"
  elif pgrp_live "$hp"; then
    say "recovery: process group $hp, run $job's harness group, still has members but its leader has ended; ownership cannot be verified; operator resolution required (marker kept)"
  elif proc_is_start "$wp" "$ws"; then
    say "recovery: the deadline watchdog of run $job (pid $wp) is still running; operator resolution required (marker kept)"
  else
    say "recovery: process group $wp, run $job's watchdog group, still has members but its leader has ended; ownership cannot be verified; operator resolution required (marker kept)"
  fi
  return 1
}

# do_recovery (isolated; local runs): every marker of an earlier local run is resolved before this run starts, or
# the launch is blocked (16) and the marker kept for an explicit operator resolution (OPERATIONS.md). A live run
# (same PID and start time), a run that ended while launching its harness, a surviving or unattributable process of
# its harness group or watchdog (recovery_survivors), unavailable Docker, a missing or invalid registry, or a
# recovery cleanup that is not fully verified blocks. Recovery is the existing cleanup-owned.sh finalize on that
# run's registry; names and labels never authorize a removal. A verified recovery records a delayed-recovery row for the original run (in its
# registry and in this run's raw/recovery.tsv), removes the old run's generated-secret list and its marker.
#
# G064 D3: markers live in STATE_ROOT/g003-pending (persistent, private); a marker of an earlier version in
# $TMPDIR/g003-pending is still read under the same-boot rules. A current marker must name a boot_id and this checkout
# (repo): an unreadable or malformed marker, or a missing or malformed boot_id, blocks (16, marker kept); another
# checkout's marker blocks ("recover it from that checkout"). A boot_id other than this boot's means the run's processes
# ended with that boot: the PID, survivor and launching checks are skipped, and the unchanged Docker/registry/cleanup path
# decides (verified: delayed-recovery row, marker removed; anything else 16). Never "clean" without that verification.
do_recovery() {
  local m job reg pid pst stage rc inv n docker_ok=0 now rows="$EV/raw/recovery.tsv" dir legacy bid mbid mrepo prev
  local hdr=$'original_job\tregistry\toriginal_started\trecovered_utc\trecovering_job\tcleanup_exit\tinventory_rows\tverdict'
  local -a dirs=()
  if pend_dir_ok; then :; else say "recovery: the marker directory $PEND_DIR is not a private directory of this user on persistent storage"; return 16; fi
  dirs=("$PEND_DIR")
  if [[ -n "$LEGACY_PEND_DIR" && "$LEGACY_PEND_DIR" != "$PEND_DIR" ]] && [[ -e "$LEGACY_PEND_DIR" || -L "$LEGACY_PEND_DIR" ]]; then
    if priv_dir_ok "$LEGACY_PEND_DIR"; then dirs+=("$LEGACY_PEND_DIR"); else
      say "recovery: the legacy marker directory $LEGACY_PEND_DIR is not a private directory of this user"; return 16
    fi
  fi
  bid="$(boot_id_now)" || { say "recovery: the boot identity (/proc/sys/kernel/random/boot_id) is unreadable; markers cannot be judged"; return 16; }
  for dir in "${dirs[@]}"; do
  legacy=0
  if [[ "$dir" != "$PEND_DIR" ]]; then legacy=1; fi
  for m in "$dir"/*; do
    if [[ ! -e "$m" && ! -L "$m" ]]; then continue; fi
    job="${m##*/}"; job="${job%.tsv}"
    if [[ -L "$m" || ! -f "$m" || ! "${m##*/}" =~ ^local-[0-9a-f]{12}\.tsv$ ]]; then
      say "recovery: $m is not a run marker; operator resolution required"; return 16
    fi
    if { : <"$m"; } 2>/dev/null; then :; else
      say "recovery: the marker of run $job ($m) is unreadable; operator resolution required (marker kept)"; return 16
    fi
    reg="$(marker_get "$m" registry)" || reg=""
    pid="$(marker_get "$m" pid)" || pid=""
    pst="$(marker_get "$m" pid_start)" || pst=""
    if [[ "$reg" != /* || ! "$pid" =~ ^[1-9][0-9]{0,9}$ || ! "$pst" =~ ^[0-9]+$ ]]; then
      say "recovery: the marker of run $job is malformed; operator resolution required"; return 16
    fi
    prev=0 mbid=""
    if (( legacy == 0 )); then
      mbid="$(marker_get "$m" boot_id)" || mbid=""
      mrepo="$(marker_get "$m" repo)" || mrepo=""
      if [[ ! "$mbid" =~ ^[0-9a-f]{32}$ ]]; then
        say "recovery: the marker of run $job has a missing or malformed boot_id; operator resolution required (marker kept)"; return 16
      fi
      if [[ "$mrepo" != /* ]]; then say "recovery: the marker of run $job has a missing or malformed repo; operator resolution required (marker kept)"; return 16; fi
      if [[ "$mrepo" != "$REPO_R" ]]; then
        say "recovery: run $job belongs to checkout $mrepo; recover it (or wait for it) from that checkout (marker kept)"; return 16
      fi
      if [[ "$mbid" != "$bid" ]]; then prev=1; fi
    fi
    stage="$(marker_get "$m" stage)" || stage=""
    case "$stage" in
      prelaunch|launching|launched) ;;
      *) say "recovery: the marker of run $job is malformed; operator resolution required"; return 16 ;;
    esac
    if (( prev )); then
      say "recovery: run $job was started in an earlier boot ($mbid); its processes ended with that boot, so the process checks are skipped"
    else
      if [[ "$(proc_start_time "$pid" 2>/dev/null)" == "$pst" ]]; then
        say "recovery: local validation run $job (pid $pid) is still running; wait for it or stop it"; return 16
      fi
      case "$stage" in
        prelaunch) ;;
        launched) if recovery_survivors "$m" "$job"; then :; else return 16; fi ;;
        launching) say "recovery: run $job ended while starting its harness, so its harness and watchdog were not recorded; operator resolution required (marker kept)"; return 16 ;;
      esac
    fi
    if (( docker_ok == 0 )); then
      if timeout 30 docker version --format '{{.Server.Version}}' >/dev/null 2>&1; then docker_ok=1; else
        say "recovery: Docker is unavailable; the unfinalized run $job cannot be recovered (marker kept)"; return 16
      fi
    fi
    if [[ -L "$reg" || ! -d "$reg" || ! -f "$reg/owned.tsv" ]]; then
      say "recovery: the ownership registry of run $job ($reg) is missing; operator resolution required (marker kept)"; return 16
    fi
    if [[ "$(awk -F'\t' '$1 == "job" { print $2 }' "$reg/owned.tsv" 2>/dev/null)" != "$job" ]]; then
      say "recovery: the registry $reg does not belong to run $job; operator resolution required (marker kept)"; return 16
    fi
    inv="$reg/recovery-inventory.tsv"; rc=0
    rm -f -- "$inv"
    bash "$REPO_R/infra/ci/cleanup-owned.sh" --registry "$reg" --mode finalize --out "$inv" >"$reg/recovery.log" 2>&1 || rc=$?
    n=unavailable
    if [[ -f "$inv" && ! -L "$inv" ]]; then n="$(grep -c . "$inv")" || n=0; fi
    now="$(ci_utc)"
    if [[ ! -e "$rows" ]]; then printf '%s\n' "$hdr" >"$rows" || return 16; fi
    if [[ "$rc" == 0 && "$n" != unavailable ]] && awk -F'\t' 'NF != 4 || $3 != "removed" { f = 1 } END { exit f ? 1 : 0 }' "$inv"; then
      printf '%s\n%s\t%s\t%s\t%s\t%s\t%s\t%s\tdelayed-recovery-verified\n' "$hdr" "$job" "$reg" "$(marker_get "$m" started || echo unknown)" \
        "$now" "$REG_JOB" "$rc" "$n" >"$reg/recovery.tsv" || return 16
      tail -n 1 "$reg/recovery.tsv" >>"$rows" || return 16
      rm -f -- "$reg/secrets" || return 16
      rm -f -- "$m" || return 16
      say "recovery: run $job did not finalize; delayed recovery verified ($n resources removed; report $reg/recovery.tsv)"
    else
      printf '%s\t%s\t%s\t%s\t%s\t%s\t%s\tblocked\n' "$job" "$reg" "$(marker_get "$m" started || echo unknown)" \
        "$now" "$REG_JOB" "$rc" "$n" >>"$rows" || true
      say "recovery: run $job left owned resources that cleanup did not verify removed (exit $rc; see $inv); operator resolution required (marker kept)"
      return 16
    fi
  done
  done
  return 0
}

# storage_observe PHASE (pre | post): read-only storage observations appended to raw/storage.tsv. Scope: the whole
# Docker daemon (its images) and the daemon's built-in builder store (its build cache); both are shared with every
# other user of the daemon, so they are never project usage, and differences are never bytes freed by this run.
# Sizes are the Docker CLI's own figures (decimal units). A failed or bounded-out command, a missing row or a
# malformed figure reads "unavailable", never zero. post adds this run's cleanup status and owned resources left.
# Not a phase: it never changes the run's status. No builder bootstrap, container, bind mount or prune.
STORAGE_HDR=$'observed_utc\tphase\tjob\tcandidate\tscope\tmetric\tvalue\tunit\tstatus'
STORAGE_SIZE_UNIT='Docker CLI size (decimal: kB=1000 B)'
storage_row() { # PHASE NOW SCOPE METRIC VALUE UNIT STATUS (the candidate is storage_observe's)
  printf '%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n' "$2" "$1" "${REG_JOB:-unknown}" "$cand" "$3" "$4" "$5" "$6" "$7" \
    >>"$EV/raw/storage.tsv"
}
storage_observe() {
  local ph="$1" now df="" rc=0 t key scope row c s r why n inv cand attempt
  local re_size='^[0-9]+(\.[0-9]+)?(B|kB|MB|GB|TB|PB)$'
  if [[ -z "$EV" || ! -d "$EV/raw" ]]; then return 0; fi
  if [[ ! -e "$EV/raw/storage.tsv" ]]; then printf '%s\n' "$STORAGE_HDR" >"$EV/raw/storage.tsv" || return 0; fi
  now="$(ci_utc)"
  cand="$HEAD_SHA"
  if [[ ! "$cand" =~ ^[0-9a-f]{40}$ ]]; then cand="$(g rev-parse --verify -q 'HEAD^{commit}' 2>/dev/null)" || cand=unknown; fi
  # The daemon is shared with other clients while this read-only query runs. G063 repair: VS3 in the background
  # ci-controls of a full run read unavailable; the unproven hypothesis is a transient failure while the
  # foreground harness removed an image. A failed query is retried twice, 2 s apart; a bounded-out (124) or killed
  # query is not retried, its budget is spent. The last attempt's exit is the reported status. VS5 covers this with
  # an injected synthetic transient failure, not a reproduced daemon race; VS4 covers a persistent failure.
  for attempt in 1 2 3; do
    rc=0
    df="$(timeout 60 docker system df --format $'{{.Type}}\t{{.TotalCount}}\t{{.Size}}\t{{.Reclaimable}}' 2>/dev/null)" || rc=$?
    if (( rc == 0 || rc >= 124 )); then break; fi
    if (( attempt < 3 )); then sleep 2; fi
  done
  for t in Images 'Build Cache'; do
    if [[ "$t" == Images ]]; then key=images scope='daemon-wide: every image on this Docker daemon (shared)'
    else key=build_cache scope="daemon builder: the daemon's built-in build cache (shared; other buildx builders not observed)"; fi
    c=unavailable s=unavailable r=unavailable why=""
    if (( rc != 0 )); then why="unavailable:docker-system-df-exit-$rc"; else
      row="$(awk -F'\t' -v t="$t" '$1 == t { print; exit }' <<<"$df")"
      if [[ -z "$row" ]]; then why="unavailable:no-${key//_/-}-row"; else
        IFS=$'\t' read -r _ c s r <<<"$row"
        r="${r%% (*}"
        if [[ ! "$c" =~ ^[0-9]+$ || ! "$s" =~ $re_size || ! "$r" =~ $re_size ]]; then c=unavailable s=unavailable r=unavailable why="unavailable:malformed"; fi
      fi
    fi
    if [[ "$key" == images ]]; then storage_row "$ph" "$now" "$scope" images_count "$c" count "${why:-ok}"
    else storage_row "$ph" "$now" "$scope" build_cache_records "$c" count "${why:-ok}"; fi
    storage_row "$ph" "$now" "$scope" "${key}_size" "$s" "$STORAGE_SIZE_UNIT" "${why:-ok}"
    storage_row "$ph" "$now" "$scope" "${key}_reclaimable" "$r" "$STORAGE_SIZE_UNIT" "${why:-ok}"
  done
  if [[ "$ph" == post ]]; then
    scope="this run's ownership registry"
    storage_row "$ph" "$now" "$scope" cleanup_status "${ST[cleanup]-not-run}" 'exit status' ok
    inv="$EV/raw/cleanup-inventory.tsv"
    if [[ -f "$inv" && ! -L "$inv" ]]; then
      n="$(awk -F'\t' 'NF != 4 || $3 != "removed" { n++ } END { print n + 0 }' "$inv")" || n=unavailable
      storage_row "$ph" "$now" "$scope" owned_remaining "$n" count ok
    else storage_row "$ph" "$now" "$scope" owned_remaining unavailable count unavailable:no-cleanup-inventory; fi
  fi
  return 0
}

storage_keep() { # parent, before the report: a local --evidence-dir run keeps the scanned candidate's reports
  local n
  if (( CI_MODE )) || [[ -z "$EVARG" || ! -d "$EV/candidate" ]]; then return 0; fi
  if [[ "${ST[candidate_scan]-}" == 0 || "${ST[rescan]-}" == 0 ]]; then :; else return 0; fi
  for n in storage.tsv recovery.tsv; do
    if [[ -f "$EV/candidate/$n" && ! -L "$EV/candidate/$n" ]]; then cp -- "$EV/candidate/$n" "$EV/$n" || return 0; fi
  done
  if [[ -f "$EV/storage.tsv" ]]; then STORAGE_KEPT="$EV/storage.tsv"; fi
  return 0
}

storage_summary() { # job log (do_report, after clean scans): one validated line per phase, then where the report is
  local f="$EV/raw/storage.tsv" ph line
  if [[ -L "$f" || ! -f "$f" ]]; then say "storage: no report"; return 0; fi
  for ph in pre post; do
    # shellcheck disable=SC2016  # an awk program
    line="$(awk -F'\t' -v p="$ph" -v re='^([0-9]+([.][0-9]+)?([kMGTP]?B)?|unavailable|not-run|n/a)$' \
      'NR > 1 && $2 == p && $6 ~ /^[a-z_]+$/ && $7 ~ re { printf "%s%s=%s", (n++ ? " " : ""), $6, $7 } END { if (!n) exit 1 }' "$f")" || continue
    say "storage $ph: $line"
  done
  if (( CI_MODE )) && [[ -n "$PUBLISH_TARGET" ]]; then
    say "storage: daemon-wide and builder-wide observations, not project usage; report ci-evidence/storage.tsv when published"
  elif [[ -n "$STORAGE_KEPT" ]]; then
    say "storage: daemon-wide and builder-wide observations, not project usage; report $STORAGE_KEPT"
  else
    say "storage: daemon-wide and builder-wide observations, not project usage; not kept (a local run keeps it with --evidence-dir)"
  fi
}

# bootstrap_post (finalize, after cleanup and before the raw scan; G051 §3.8.3, Codex Q1): writes
# raw/report/ledger-bootstrap.tsv, a header line and one row: ref, pre_id, post_id, dockerfile_sha256, bases,
# platform, verdict. In the hosted adapter context it re-inspects the context image after cleanup. Equal, valid
# identities before and after (with the same Dockerfile, bases and platform) support only the structural
# inference "structural-inference:reuse" (the after-script finds the image and builds nothing); it is never
# step, cache or build-count evidence. A missing identity is "absent" and a different image or Dockerfile
# "changed"; both leave AC5 UNMET. Outside that context the row is "not-applicable". Written only after a
# harness run, never through a symlink; it is not a phase and never gates the job.
bootstrap_post() {
  local d="$EV/raw/report" f b ref=none pre=none post=none sha=none bases=none plat=none verdict=not-applicable now
  local -a lines=() v=()
  if [[ ! "${ST[harness]-}" =~ ^[0-9]+$ ]]; then return 0; fi
  if [[ -L "$d" ]]; then say "bootstrap: the report directory is a symlink"; return 1; fi
  if [[ ! -d "$d" ]]; then mkdir -m 700 -- "$d" 2>/dev/null || { say "bootstrap: cannot create the report directory"; return 1; }; fi
  f="$d/ledger-bootstrap.tsv"
  rm -f -- "$f" 2>/dev/null || true
  if [[ -e "$f" || -L "$f" ]]; then say "bootstrap: an existing ledger-bootstrap.tsv cannot be replaced"; return 1; fi
  if (( CI_MODE )) && ci_hosted_context; then
    ref="platform-ci-context:$CI_JOB_ID" pre=unknown sha=unknown bases=unknown plat=unknown
    b="$EV/state/bootstrap.tsv"
    if [[ -f "$b" && ! -L "$b" ]]; then mapfile -t lines <"$b" || lines=(); fi
    if (( ${#lines[@]} == 2 )) && [[ "${lines[0]}" == $'ref\timage_id\tdockerfile\tdockerfile_sha256\tbases\tplatform' ]]; then
      IFS=$'\t' read -r -a v <<<"${lines[1]}"
      if (( ${#v[@]} == 6 )) && [[ "${v[0]}" == "$ref" && "${v[2]}" == "$CTX_DOCKERFILE" ]]; then
        if [[ "${v[1]}" =~ ^sha256:[0-9a-f]{64}$ ]]; then pre="${v[1]}"; fi
        if [[ "${v[3]}" =~ ^[0-9a-f]{64}$ ]]; then sha="${v[3]}"; fi
        if [[ "${v[4]}" =~ ^[^,[:space:]]+,[^,[:space:]]+$ ]]; then bases="${v[4]}"; fi
        if [[ "${v[5]}" =~ ^[a-z0-9]+/[a-z0-9_]+$ ]]; then plat="${v[5]}"; fi
      fi
    fi
    post="$(docker image inspect --format '{{.Id}}' "$ref" 2>/dev/null)" || post=absent
    if [[ "$post" != absent && ! "$post" =~ ^sha256:[0-9a-f]{64}$ ]]; then post=unknown; fi
    now="$(sha256sum <"$REPO_R/$CTX_DOCKERFILE" 2>/dev/null | cut -c1-64)" || now=""
    if [[ "$pre" == unknown || ! "$post" =~ ^sha256: ]]; then verdict=absent
    elif [[ "$pre" != "$post" ]]; then verdict=changed
    elif [[ "$sha" == unknown || "$bases" == unknown || "$plat" == unknown ]]; then verdict=absent
    elif [[ "$now" != "$sha" ]]; then verdict=changed
    else verdict=structural-inference:reuse
    fi
  fi
  (umask 077 && set -C && printf 'ref\tpre_id\tpost_id\tdockerfile_sha256\tbases\tplatform\tverdict\n%s\t%s\t%s\t%s\t%s\t%s\t%s\n' \
    "$ref" "$pre" "$post" "$sha" "$bases" "$plat" "$verdict" >"$f") || { rm -f -- "$f" 2>/dev/null || true; say "bootstrap: cannot write ledger-bootstrap.tsv"; return 1; }
  if [[ "$verdict" == not-applicable ]]; then say "bootstrap: not-applicable (not the hosted adapter context)"
  else say "bootstrap: $verdict (a structural inference from the context image identity before and after cleanup; not step evidence)"; fi
  return 0
}

# scan_set LABEL DIR...: 0 clean, 15 finding, 17 scanner failure. (0) every entry is a readable regular
# file or a readable, searchable directory; (a) literal scan of the generated secrets; (b) gitleaks;
# (c) forbidden file names. gitleaks reads a tar stream on stdin instead of bind mounts: Docker Desktop
# on WSL never releases a bind mount's host path, so per-run directories exhaust its mount table
# (docs/foundation/OPERATIONS.md, "Docker Desktop bind-mount leak").
scan_set() {
  local label="$1" d rc n i=0 out="$EV/scan/$1" src="$EV/scan/$1.src" content st=() names=()
  shift
  for d in "$@"; do
    if [[ -L "$d" || ! -d "$d" ]]; then say "scan: $label: $d is not a directory"; return 17; fi
    n="$(find "$d" \( ! -type f ! -type d \) -print | wc -l)" || { say "scan: $label: cannot list the evidence"; return 17; }
    if (( n != 0 )); then say "scan: $label: $n entries are neither regular files nor directories"; return 17; fi
    n="$(find "$d" \( ! -readable -o \( -type d ! -executable \) \) -print | wc -l)" || { say "scan: $label: cannot list the evidence"; return 17; }
    if (( n != 0 )); then say "scan: $label: $n entries are unreadable"; return 17; fi
  done
  if [[ "${ST[harness]-}" =~ ^[0-9]+$ ]]; then
    if [[ -L "$REG/secrets" || ! -f "$REG/secrets" ]]; then say "scan: $label: the generated-secret list is missing"; return 17; fi
    n="$(grep -c -v '^$' "$REG/secrets")" || n=0
    if (( n == 0 )); then say "scan: $label: the generated-secret list is empty"; return 17; fi
  fi
  if [[ -f "$REG/secrets" && ! -L "$REG/secrets" ]]; then
    n="$(grep -c -v '^$' "$REG/secrets")" || n=0
    if (( n > 0 )); then
      rc=0
      grep -rqF -f <(grep -v '^$' "$REG/secrets") -- "$@" || rc=$?
      case "$rc" in
        0) say "scan: $label: a generated secret was found"; return 15 ;;
        1) ;;
        *) say "scan: $label: literal scan failed (grep exit $rc)"; return 17 ;;
      esac
    fi
  fi
  rm -rf -- "$out" "$src"
  mkdir -m 700 -- "$out" "$src" "$src/scan" "$src/cfg" || return 17
  : >"$out.log" || return 17
  # The stream holds scan/s<i> (each DIR) and cfg/gitleaks.toml. The links are top-level only: tar -h
  # follows them, and step (0) proved the trees below contain no other links. --mode=u+rwX gives the owner write
  # and search access in the stream only: the scanner's busybox tar runs as this user and could not unpack the
  # contents of a directory without them (an evidence directory may legitimately be read-only).
  for d in "$@"; do ln -s -- "$d" "$src/scan/s$i" || return 17; names+=("scan/s$i"); i=$((i + 1)); done
  ln -s -- "$REPO_R/infra/security/gitleaks.toml" "$src/cfg/gitleaks.toml" || return 17
  # The container unpacks into /tmp/w, scans /tmp/w/scan and returns the report as a tar on stdout; its
  # exit status is gitleaks's (122 when the evidence could not be unpacked). st = (pack, docker, unpack).
  tar -C "$src" -chf - --mode=u+rwX "${names[@]}" cfg/gitleaks.toml 2>>"$out.log" \
    | wrapper_run --rm -i --pull never --network none --user "$(id -u):$(id -g)" --entrypoint sh \
      "$CI_GITLEAKS_IMAGE" -c 'mkdir -p /tmp/w/out && tar -xf - -C /tmp/w || exit 122
        gitleaks dir /tmp/w/scan --config /tmp/w/cfg/gitleaks.toml --redact --no-banner --exit-code 23 \
          --report-format json --report-path /tmp/w/out/report.json >&2
        rc=$?; tar -C /tmp/w/out -cf - report.json; exit "$rc"' 2>>"$out.log" \
    | tar -xf - -C "$out" --no-same-owner 2>>"$out.log" \
    && st=("${PIPESTATUS[@]}") || st=("${PIPESTATUS[@]}")
  rm -rf -- "$src"
  rc="${st[1]:-17}"
  if [[ -L "$out/report.json" || ! -f "$out/report.json" ]]; then say "scan: $label: gitleaks wrote no report (exit $rc)"; return 17; fi
  content="$(tr -d ' \t\r\n' <"$out/report.json")" || return 17
  case "$rc" in
    0) if [[ "$content" != "[]" ]]; then say "scan: $label: gitleaks report is not an empty list"; return 17; fi
       if [[ "${st[0]}" != 0 || "${st[2]}" != 0 ]]; then
         say "scan: $label: the evidence stream failed (pack ${st[0]}, unpack ${st[2]})"; return 17
       fi ;;
    23) if [[ "$content" != "["*"]" ]]; then say "scan: $label: gitleaks report is unparseable"; return 17; fi
        say "scan: $label: gitleaks reported a finding"; return 15 ;;
    *) say "scan: $label: gitleaks failed (exit $rc)"; return 17 ;;
  esac
  n="$(find "$@" \( -name '.env' -o -name '*.env' -o -name 'secrets*' -o -name 'exec-token' \) -print | wc -l)" || {
    say "scan: $label: cannot list the evidence"; return 17; }
  if (( n != 0 )); then say "scan: $label: a forbidden file name is present"; return 15; fi
  return 0
}

# wrapper_run ARGS...: `docker run ARGS...` for validate.sh's own post-harness evidence helpers (G064 Plan A corr1 P5,
# corr2 D2). A registry cidfile, cid/wrapper-<random>.cid, binds the container; after the call it must be verified gone
# (inspect fails "No such object" within WRAPPER_GONE_DS tenths of a second; one `rm -f -v` of the exact ID and the same
# bound otherwise). Verified: the cidfile goes. Not verified, or a status 0 without a cidfile ID: a line in
# state/wrapper-unverified, which makes the cleanup status 18 (registry and marker kept). A run latch (hold.d/run) is
# logged and the helper still runs (serial, --network none; teardown and evidence are never gated). Returns the helper's
# own status; writes nothing to stdout itself (callers read it).
WRAPPER_GONE_DS=50
wrapper_gone() { # ID: 0 when inspect reports "No such" within WRAPPER_GONE_DS
  local i out
  for ((i = 0; i < WRAPPER_GONE_DS; i++)); do
    if out="$(docker container inspect --format '{{.Id}}' "$1" 2>&1)"; then :; elif [[ "$out" == *"No such"* ]]; then return 0; fi
    sleep 0.1
  done
  return 1
}
wrapper_run() {
  local cf rc=0 id="" why=""
  if [[ -e "$REG/hold.d/run" ]]; then printf 'validate: wrapper: the harness run is held (hold.d/run); this evidence helper still runs\n' >&2; fi
  if [[ -L "$REG/cid" || ! -d "$REG/cid" ]]; then
    printf 'wrapper\t-\tno registry cid directory\n' >>"$EV/state/wrapper-unverified"
    docker run "$@" || rc=$?
    return "$rc"
  fi
  cf="$REG/cid/wrapper-$(ci_random_hex 6).cid"
  docker run --cidfile "$cf" "$@" || rc=$?
  if [[ -f "$cf" && ! -L "$cf" ]]; then id="$(tr -d ' \n' <"$cf")"; fi
  if [[ -z "$id" ]]; then
    if (( rc == 0 )); then why="status 0 without a cidfile ID"; else rm -f -- "$cf"; fi
  elif [[ ! "$id" =~ ^[0-9a-f]{64}$ ]]; then why="malformed cidfile ID"
  elif wrapper_gone "$id"; then rm -f -- "$cf"
  else
    docker rm -f -v "$id" >/dev/null 2>&1 || true
    if wrapper_gone "$id"; then rm -f -- "$cf"; else why="container not verified gone"; fi
  fi
  if [[ -n "$why" ]]; then printf 'wrapper\t%s\t%s\n' "${id:--}" "$why" >>"$EV/state/wrapper-unverified"; fi
  return "$rc"
}

do_raw_scan() { scan_set raw "$EV/raw"; }
do_candidate_scan() { scan_set candidate "$EV/candidate" "$EV/final"; }

yq_json() { # FILE EXPRESSION: evaluate EXPRESSION on a JSON file with the pinned yq image (no network)
  wrapper_run --rm --pull never --network none -i "$CI_YQ_IMAGE" -p json -o tsv "$2" <"$1" 2>>"$EV/scan/yq.log"
}

# suite_set SUITES: " suite suite ... final " (run.sh suites plus the final checks; G051: the executed or
# pipeline suites, never a profile name).
suite_set() {
  if [[ ! "$1" =~ $RE_PSUITES || "$1" == none ]]; then return 1; fi
  printf ' %s final ' "$1"
}

# required_ids SUITES: the sorted required IDs of SUITES plus final from the checked-out required-checks.tsv.
required_ids() {
  local set id suite
  set="$(suite_set "$1")" || return 1
  while IFS=$'\t' read -r id suite; do
    if [[ -z "$id" || "$id" == '#'* ]]; then continue; fi
    if [[ "$set" == *" $suite "* ]]; then printf '%s\n' "$id"; fi
  done <"$REPO_R/infra/validation/required-checks.tsv" | LC_ALL=C sort
}

# pipeline_suite_list: the suites of E_pipeline (G051 §3.7): PIPELINE_SUITES, plus the fast suites when the
# fast job ran the fast partition of this pipeline.
pipeline_suite_list() {
  if [[ "$PARTITION" == fast || "$PARTITION" == remainder ]]; then printf '%s %s\n' "$PIPELINE_SUITES" "$CI_SUITES_fast"; return 0; fi
  printf '%s\n' "$PIPELINE_SUITES"
}

# go_tests_complete REP SET (G064 Plan r2 §5.4): when the executed set holds migrations or ops, the per-test Go
# evidence is complete: (i) go-tests.tsv, go-tests-expected.tsv, go-tests-junit.xml and db-lifecycle.tsv are regular,
# non-empty and well-formed (their exact headers; 15 columns per go-tests row; the JUnit fragment grammar the harness
# writes, one suite per invocation row); (ii) every invocation the executed suites declared has exactly one invocation
# row with status pass or fail and a numeric rc: interrupted, unfinished, duplicate, missing or undeclared is a
# finding; (iii) each invocation row's counts agree with its package and test rows, and every log_file is a regular
# file under go-tests/. Every finding is said (at most 20, then the number left) and the status is 14. It only adds
# findings: the results.tsv and required-ID rules (iv) are unchanged and still decide PASS (D-5). Bash and awk only.
# The header is GO_TESTS_HEADER of infra/validation/lib.sh (= Header of infra/validation/tools/internal/gotest).
GO_TESTS_HDR=$'kind\tinvocation\tpackage\ttest\tstatus\trc\texpected\tpassed\tfailed\tunfinished\telapsed_s\tlog_file\tline_from\tline_to\tdetail'
# shellcheck disable=SC2016 # an awk program: its $ fields are awk's
GO_TESTS_AWK='
function bad(m) { print "F " m }
function okname(s) { return s ~ /^[A-Za-z0-9_.-]+$/ && length(s) <= 120 }
function num(s) { return s ~ /^[0-9]+$/ }
BEGIN { FS = "\t"; n = split(suites, a, " "); for (i = 1; i <= n; i++) want[a[i]] = 1 }
FILENAME == fe {
  if (FNR == 1) { if ($0 != "suite\tinvocation") bad("go-tests-expected.tsv does not start with its header"); next }
  if (NF != 2 || $1 !~ /^[a-z][a-z-]*$/ || !okname($2)) { bad("go-tests-expected.tsv line " FNR " is malformed"); next }
  if ($2 in decl) { bad("go-tests-expected.tsv declares invocation " $2 " twice"); next }
  decl[$2] = $1
  if ($1 in want) { nd[$1]++; dorder[++ndo] = $2 }
  next
}
FILENAME == ft {
  if (FNR == 1) { if ($0 != hdr) bad("go-tests.tsv does not start with its header"); next }
  if (NF != 15 || !okname($2) || ($1 != "invocation" && $1 != "package" && $1 != "test")) { bad("go-tests.tsv line " FNR " is malformed"); next }
  if ($12 !~ /^go-tests\/[A-Za-z0-9_.-]+$/ || length($12) > 129) { bad("go-tests.tsv line " FNR " names no go-tests/ log"); next }
  v = $2; logs[$12] = 1; seen[v] = 1
  if ($1 == "invocation") {
    if (++ninv[v] == 1) iorder[++nio] = v
    st[v] = $5; rc[v] = $6; ie[v] = $7; ip[v] = $8; ifl[v] = $9; iu[v] = $10
    if (!num($7) || !num($8) || !num($9) || !num($10)) { bad("invocation " v " has malformed counts"); badc[v] = 1 }
  } else if ($1 == "package") {
    if (num($7)) pe[v] += $7; else { bad("go-tests.tsv line " FNR " has a malformed expected count"); badc[v] = 1 }
  } else {
    if ($15 == "parent=") { if ($5 == "pass") tp[v]++; else if ($5 == "fail") tf[v]++ }
    if ($5 == "unfinished") tu[v]++
  }
  next
}
FILENAME == fj {
  jn = FNR
  if (FNR == 1) { if ($0 != "<?xml version=\"1.0\" encoding=\"UTF-8\"?>") jbad = 1; next }
  if (FNR == 2) { if ($0 != "<testsuites>") jbad = 1; next }
  if (jend) { jbad = 1; next }
  if ($0 == "</testsuites>") { if (jin) jbad = 1; jend = 1; next }
  if ($0 ~ /^<testsuite name="[A-Za-z0-9_.-]+" tests="[0-9]+" failures="[0-9]+" errors="[0-9]+" skipped="[0-9]+">$/) {
    if (jin) jbad = 1
    jin = 1; s = $0; sub(/^<testsuite name="/, "", s); sub(/".*$/, "", s); js[s]++; next
  }
  if ($0 == "</testsuite>") { if (!jin) jbad = 1; jin = 0; next }
  if ($0 ~ /^<testcase [^<>]*\/>$/ || $0 ~ /^<testcase [^<>]*><(failure|error|skipped) [^<>]*\/><\/testcase>$/) { if (!jin) jbad = 1; next }
  jbad = 1
}
FILENAME == fd {
  if (FNR == 1) { if ($0 != "ts_utc\tkind\tsubject\tfield\tvalue") bad("db-lifecycle.tsv does not start with its header"); next }
  if (NF < 5 || $1 !~ /^[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]T[0-9][0-9]:[0-9][0-9]:[0-9][0-9][.][0-9][0-9][0-9]Z$/ || $2 == "") bad("db-lifecycle.tsv line " FNR " is malformed")
  next
}
END {
  for (s in want) if (!(s in nd)) bad("go-tests-expected.tsv declares no invocation of the executed suite " s)
  for (i = 1; i <= ndo; i++) {
    v = dorder[i]
    if (!(v in ninv)) { bad("declared invocation " v " (" decl[v] ") has no invocation row"); continue }
    if (ninv[v] > 1) { bad("invocation " v " has " ninv[v] " invocation rows"); continue }
    if (st[v] != "pass" && st[v] != "fail") bad("invocation " v " is " st[v] " (rc " rc[v] ")")
    if (!num(rc[v])) bad("invocation " v " records no exit status")
  }
  for (i = 1; i <= nio; i++) {
    v = iorder[i]
    if (!(v in decl)) bad("invocation " v " is not declared in go-tests-expected.tsv")
    else if (!(decl[v] in want)) bad("invocation " v " belongs to " decl[v] ", which did not execute")
    if (ninv[v] != 1 || (v in badc)) continue
    if (tp[v] + 0 != ip[v] || tf[v] + 0 != ifl[v] || tu[v] + 0 != iu[v] || pe[v] + 0 != ie[v])
      bad("invocation " v " counts expected/passed/failed/unfinished " ie[v] "/" ip[v] "/" ifl[v] "/" iu[v] " disagree with its rows " pe[v] + 0 "/" tp[v] + 0 "/" tf[v] + 0 "/" tu[v] + 0)
  }
  for (v in seen) if (!(v in ninv)) bad("rows of " v " have no invocation row")
  if (jbad || !jend || jn < 3) bad("go-tests-junit.xml is malformed")
  else {
    for (v in js) if (js[v] != 1 || !(v in ninv)) bad("go-tests-junit.xml suite " v " does not match one invocation row")
    for (i = 1; i <= nio; i++) if (!(iorder[i] in js)) bad("go-tests-junit.xml has no suite for invocation " iorder[i])
  }
  for (l in logs) print "L " l
}'
go_tests_complete() {
  local rep="$1" set="$2" s gsuites="" f bad=0 out line kind val nf=0
  for s in migrations ops; do if [[ "$set" == *" $s "* ]]; then gsuites+=" $s"; fi; done
  if [[ -z "$gsuites" ]]; then return 0; fi
  for f in go-tests.tsv go-tests-expected.tsv go-tests-junit.xml db-lifecycle.tsv; do
    if [[ -L "$rep/$f" || ! -f "$rep/$f" || ! -s "$rep/$f" ]]; then say "complete: go-tests: $f is missing"; bad=1; fi
  done
  if (( bad )); then return 14; fi
  out="$(LC_ALL=C awk -v suites="$gsuites" -v hdr="$GO_TESTS_HDR" -v fe="$rep/go-tests-expected.tsv" -v ft="$rep/go-tests.tsv" \
    -v fj="$rep/go-tests-junit.xml" -v fd="$rep/db-lifecycle.tsv" "$GO_TESTS_AWK" "$rep/go-tests-expected.tsv" "$rep/go-tests.tsv" \
    "$rep/go-tests-junit.xml" "$rep/db-lifecycle.tsv")" || { say "complete: go-tests: the reports cannot be read"; return 14; }
  while IFS= read -r line; do
    kind="${line%% *}" val="${line#* }"
    case "$kind" in
      F) nf=$((nf + 1)); if (( nf <= 20 )); then say "complete: go-tests: $val"; fi ;;
      L) if [[ -L "$rep/go-tests" || -L "$rep/$val" || ! -f "$rep/$val" ]]; then
           nf=$((nf + 1)); if (( nf <= 20 )); then say "complete: go-tests: $val is missing"; fi
         fi ;;
    esac
  done <<<"$out"
  if (( nf > 20 )); then say "complete: go-tests: $((nf - 20)) further findings"; fi
  if (( nf )); then return 14; fi
  return 0
}

do_complete() {
  local rep="$EV/raw/report" f line id st n out d name malformed=0 failrow=0 nreq=0 set label gt=0
  local re=$'^([A-Z][A-Z0-9-]*)\t(PASS|FAIL|UNMET|INFO)\t[^\t]*$'
  local rq=$'^([A-Z][A-Z0-9-]*)\t(lint|lint-base|lint-web|lint-core|controls|ci-controls|provenance|contracts|unit|unit-web|unit-core|unit-intel|root|migrations|ops|rename|off|feature-web|feature-core|staging|final)$'
  local sre=$'^[1-9][0-9]*\t0\t0\t0$'
  local nre=$'^!!map\t(!!null\tnull|!!int\t0)\t([1-9][0-9]*)\t([0-9]+)$' cvec tok key evec rvec cn cset
  local -a v=()
  local -A seen=() pass=() required=() listed=()
  complete_diag "$rep"
  f="$rep/results.tsv"
  : >"$EV/final/status-columns.tsv" || return 14
  if [[ -L "$f" || ! -f "$f" || ! -s "$f" ]]; then say "complete: results.tsv is missing"; return 14; fi
  while IFS= read -r line || [[ -n "$line" ]]; do
    if [[ "$line" =~ $re ]]; then
      id="${BASH_REMATCH[1]}" st="${BASH_REMATCH[2]}"
      printf '%s\t%s\n' "$id" "$st" >&3
      seen[$id]=$(( ${seen[$id]:-0} + 1 ))
      if [[ "$st" == PASS ]]; then pass[$id]=1; fi
      if [[ "$st" == FAIL || "$st" == UNMET ]]; then failrow=1; fi
    else
      malformed=1
    fi
  done <"$f" 3>"$EV/final/status-columns.tsv"
  if (( malformed )); then say "complete: results.tsv has a malformed line"; return 14; fi
  f="$REPO_R/infra/validation/required-checks.tsv"
  if [[ -L "$f" || ! -f "$f" ]]; then say "complete: required-checks.tsv is missing"; return 14; fi
  # The job's own obligations (G051): the executed suites of its partition plus final.
  set="$(suite_set "$EXECUTED_SUITES")" || { say "complete: unknown executed suites"; return 14; }
  label="$EXECUTED_PROFILE"
  if [[ "$PARTITION" == remainder ]]; then label=remainder; fi
  # G064: the per-test Go evidence is reported first and its status kept; every rule below still applies unchanged.
  go_tests_complete "$rep" "$set" || gt=$?
  while IFS= read -r line || [[ -n "$line" ]]; do
    if [[ -z "$line" || "$line" == '#'* ]]; then continue; fi
    if [[ ! "$line" =~ $rq ]]; then say "complete: required-checks.tsv has a malformed line"; return 14; fi
    id="${BASH_REMATCH[1]}"
    if [[ -n "${listed[$id]+x}" ]]; then say "complete: required-checks.tsv lists $id twice"; return 14; fi
    listed[$id]="${BASH_REMATCH[2]}"
    if [[ "$set" == *" ${BASH_REMATCH[2]} "* ]]; then required[$id]=1; nreq=$((nreq + 1)); fi
  done <"$f"
  if (( nreq == 0 )); then say "complete: required-checks.tsv lists no check"; return 14; fi
  for id in "${!required[@]}"; do
    if [[ "${seen[$id]:-0}" != 1 ]]; then say "complete: required check $id occurs ${seen[$id]:-0} times"; return 14; fi
    if [[ -z "${pass[$id]+x}" ]]; then say "complete: required check $id is not PASS"; return 14; fi
  done
  if (( failrow )); then say "complete: results.tsv has a FAIL or UNMET row"; return 14; fi
  # No row may come from a required check outside the profile (a narrower result cannot pose as a wider one,
  # and a wider one cannot hide inside a narrower profile).
  for id in "${!seen[@]}"; do
    if [[ -n "${listed[$id]+x}" && -z "${required[$id]+x}" ]]; then
      say "complete: results.tsv has a row for $id, outside the $label profile"; return 14
    fi
  done
  # Browser reports for each executed browser suite (G051: feature-web writes browser-feature-results.json).
  for d in root off rename feature-web; do
    if [[ "$set" != *" $d "* ]]; then continue; fi
    name="${d%-web}"
    f="$rep/browser-$name-results.json"
    if [[ -L "$f" || ! -f "$f" || ! -s "$f" ]]; then say "complete: browser-$name-results.json is missing"; return 14; fi
    out="$(yq_json "$f" '[.stats.expected, .stats.skipped, .stats.flaky, .stats.unexpected]')" || {
      say "complete: browser-$name-results.json is malformed"; return 14; }
    if [[ ! "$out" =~ $sre ]]; then say "complete: browser-$name stats are not (expected>0, 0 skipped, 0 flaky, 0 unexpected)"; return 14; fi
  done
  if [[ "$EXECUTED_PROFILE" == staging ]]; then
    # The identity-staging profile (G004 CI-1) is the staging suite and the final checks only: no contract or
    # browser report belongs to it.
    f="$rep/inventory-leftover.txt"
    if [[ -L "$f" || ! -f "$f" ]]; then say "complete: inventory-leftover.txt is missing"; return 14; fi
    n="$(awk '$0 != "" { c++ } END { print c + 0 }' "$f")" || return 14
    if (( n != 0 )); then say "complete: the harness left $n resources behind"; return 14; fi
    complete_resources_present "$rep" || return 14
    return "$gt"
  fi
  # Contract reports are required if and only if the contracts suite executed (G051).
  if [[ "$set" == *" contracts "* ]]; then
    # Normal contract reports: each must hold exactly the unique, complete (id, valid, class) case set of
    # the declared corpus (contracts/conformance/intelligence/expectations.json in this checkout), with
    # typed fields, counts consistent with its cases, nothing skipped and no mismatch. A case key is
    # "id valid class"; ids and classes are plain tokens, so once every case is typed and every id is
    # unique the sorted, joined key list is an exact encoding of the set. Corpus vector: root tag, cases
    # tag, cases, typed cases, unique ids, key list.
    tok='select(tag == "!!map") | select((.id | tag) == "!!str") | select((.valid | tag) == "!!bool") | select((.class | tag) == "!!str") | select(.id | test("^[A-Za-z0-9._-]{1,64}$")) | select(.class | test("^[a-z][a-z0-9-]{0,63}$"))'
    key='(.id | to_string) + " " + (.valid | to_string) + " " + (.class | to_string)'
    evec="[tag, (.cases | tag), (.cases | length), ([.cases[] | $tok] | length), ([.cases[] | .id] | unique | length), \"set:\" + ([.cases[] | $key] | sort | join(\",\"))]"
    # Report vector (23 fields): root tag, results tag, cases, typed cases, unique ids, valid, invalid;
    # counts tag, total tag and value, valid tag and value, invalid tag and value; by_class as declared and
    # as counted; skipped tag and value; mismatches tag and length; match keys, true match keys; key list.
    rvec="[tag, (.results | tag), (.results | length), ([.results[] | $tok] | length), ([.results[] | .id] | unique | length), ([.results[] | select(.valid == true)] | length), ([.results[] | select(.valid == false)] | length), (.counts | tag), (.counts.total | tag), (.counts.total | to_string), (.counts.valid | tag), (.counts.valid | to_string), (.counts.invalid | tag), (.counts.invalid | to_string), \"class:\" + ([.counts.by_class | to_entries | .[] | .key + \"=\" + (.value | to_string)] | sort | join(\",\")), \"class:\" + ([.results | group_by(.class) | .[] | (.[0].class | to_string) + \"=\" + (length | to_string)] | sort | join(\",\")), (.skipped | tag), (.skipped | to_string), (.mismatches | tag), (.mismatches | length), ([.. | select(tag == \"!!map\" and has(\"match\"))] | length), ([.. | select(tag == \"!!map\" and has(\"match\")) | .match | select(tag == \"!!bool\") | select(. == true)] | length), \"set:\" + ([.results[] | $key] | sort | join(\",\"))]"
    f="$REPO_R/contracts/conformance/intelligence/expectations.json"
    if [[ -L "$f" || ! -f "$f" || ! -s "$f" ]]; then say "complete: the declared corpus expectations are missing"; return 14; fi
    out="$(yq_json "$f" "$evec")" || { say "complete: the declared corpus expectations are malformed"; return 14; }
    IFS=$'\t' read -r -a v < <(printf '%s\n' "$out")
    if (( ${#v[@]} != 6 )) || [[ "${v[0]}" != '!!map' || "${v[1]}" != '!!seq' || ! "${v[2]}" =~ ^[1-9][0-9]*$ || "${v[3]}" != "${v[2]}" || "${v[4]}" != "${v[2]}" ]]; then
      say "complete: the declared corpus expectations are malformed, empty or repeat a case ID"; return 14
    fi
    cn="${v[2]}" cset="${v[5]}"
    # Schema-negative reports keep their own vector: root tag, skipped tag and value, number of match
    # keys, number of match keys that are the boolean true.
    cvec='[tag, (.skipped | tag), (.skipped), ([.. | select(tag == "!!map" and has("match"))] | length), ([.. | select(tag == "!!map" and has("match")) | .match | select(tag == "!!bool") | select(. == true)] | length)]'
    for d in go python; do
      f="$rep/contracts-results-$d.json"
      if [[ -L "$f" || ! -f "$f" || ! -s "$f" ]]; then say "complete: contracts-results-$d.json is missing"; return 14; fi
      out="$(yq_json "$f" "$rvec")" || { say "complete: contracts-results-$d.json is malformed"; return 14; }
      IFS=$'\t' read -r -a v < <(printf '%s\n' "$out")
      if (( ${#v[@]} != 23 )) || [[ "${v[0]}" != '!!map' || "${v[1]}" != '!!seq' || ! "${v[2]}" =~ ^[0-9]+$ ]]; then
        say "complete: contracts-results-$d.json is malformed"; return 14
      fi
      if [[ "${v[3]}" != "${v[2]}" ]]; then say "complete: contracts-results-$d.json has a mistyped or malformed case"; return 14; fi
      if [[ "${v[4]}" != "${v[2]}" ]]; then say "complete: contracts-results-$d.json repeats a case ID"; return 14; fi
      if [[ "${v[7]}" != '!!map' || "${v[8]}" != '!!int' || "${v[10]}" != '!!int' || "${v[12]}" != '!!int' ]] ||
        [[ "${v[9]}" != "${v[2]}" || "${v[11]}" != "${v[5]}" || "${v[13]}" != "${v[6]}" || "${v[14]}" != "${v[15]}" ]]; then
        say "complete: contracts-results-$d.json counts are inconsistent with its cases"; return 14
      fi
      if [[ "${v[16]}" != '!!int' || "${v[17]}" != 0 || "${v[20]}" != "${v[21]}" ]] ||
        [[ "${v[18]}" != '!!null' && ( "${v[18]}" != '!!seq' || "${v[19]}" != 0 ) ]]; then
        say "complete: contracts-results-$d.json skipped a case or reports a mismatch"; return 14
      fi
      if [[ "${v[2]}" != "$cn" || "${v[22]}" != "$cset" ]]; then
        say "complete: contracts-results-$d.json is not the declared case set (a case is missing, extra or has another verdict)"; return 14
      fi
      f="$rep/schema-negative-$d.json"
      if [[ -L "$f" || ! -f "$f" || ! -s "$f" ]]; then say "complete: schema-negative-$d.json is missing"; return 14; fi
      out="$(yq_json "$f" "$cvec")" || { say "complete: schema-negative-$d.json is malformed"; return 14; }
      if [[ ! "$out" =~ $nre || "${BASH_REMATCH[2]}" != "${BASH_REMATCH[3]}" ]]; then
        say "complete: schema-negative-$d.json skipped a case or has a mismatch"; return 14
      fi
    done
  fi
  f="$rep/inventory-leftover.txt"
  if [[ -L "$f" || ! -f "$f" ]]; then say "complete: inventory-leftover.txt is missing"; return 14; fi
  n="$(awk '$0 != "" { c++ } END { print c + 0 }' "$f")" || return 14
  if (( n != 0 )); then say "complete: the harness left $n resources behind"; return 14; fi
  complete_resources_present "$rep" || return 14
  return "$gt"
}

# complete_resources_present REP (G064 corr1 C1): resources.tsv must exist, since validate.sh always runs the harness with
# a registry; its FAIL rows reach results.tsv as RESOURCES FAIL rows (any FAIL row is 14 above).
complete_resources_present() {
  local f="$1/resources.tsv"
  if [[ -L "$f" || ! -f "$f" || ! -s "$f" ]]; then say "complete: resources.tsv is missing (the harness ran with a registry)"; return 1; fi
  return 0
}

# complete_diag REP (G064 corr1 C6 Q6): the leftover count and the accounting findings, printed before any rule of
# do_complete can return, so that cleanup and leftover findings appear beside an earlier failure. A diagnostic only:
# it never changes a status (the rules of do_complete decide, unchanged).
complete_diag() {
  local rep="$1" f n line
  f="$rep/inventory-leftover.txt"
  if [[ -L "$f" || ! -f "$f" ]]; then say "complete: diagnostic: inventory-leftover.txt is missing"
  else
    n="$(awk '$0 != "" { c++ } END { print c + 0 }' "$f" 2>/dev/null)" || n=unreadable
    say "complete: diagnostic: leftover resources (inventory-leftover.txt): $n"
  fi
  f="$rep/resources.tsv"
  if [[ -L "$f" || ! -f "$f" ]]; then say "complete: diagnostic: resources.tsv is missing"
  else
    n="$(awk -F'\t' '($1 == "total" && $9 ~ /^FAIL/) || $1 == "finding" { c++ } END { print c + 0 }' "$f" 2>/dev/null)" || n=unreadable
    say "complete: diagnostic: resources.tsv accounting findings: $n"
    while IFS= read -r line; do say "complete: diagnostic:   $line"; done < <(awk -F'\t' '
      $1 == "total" && $9 ~ /^FAIL/ { print $2 " " $3 ": " $9 " (creations " $4 ", final_leftover " $7 ", unaccounted " $8 ")" }
      $1 == "finding" { print $2 " " $3 " " substr($4, 1, 12) ": " $5 }' "$f" 2>/dev/null | head -n 10)
  fi
  f="$rep/results.tsv"
  if [[ -f "$f" && ! -L "$f" ]]; then
    n="$(grep -c $'^RESOURCES\tFAIL\t' "$f" 2>/dev/null)" || n=0
    if (( n > 0 )); then
      say "complete: diagnostic: RESOURCES FAIL rows: $n"
      while IFS= read -r line; do say "complete: diagnostic:   ${line:0:300}"; done < <(grep $'^RESOURCES\tFAIL\t' "$f" | cut -f3 | head -n 10)
    fi
  fi
  return 0
}

# stage_one SRC NAME: copy one allow-listed file into staged.tmp (browser JSON without attachments).
stage_one() {
  local src="$1" name="$2" dst="$EV/staged.tmp/$2" rc
  if [[ ! -e "$src" && ! -L "$src" ]]; then return 0; fi
  if [[ "$name" == browser-*-results.json ]]; then
    if [[ -L "$src" || ! -f "$src" ]]; then return 1; fi
    : >"$dst" || return 1
    rc=0
    wrapper_run --rm --pull never --network none -i "$CI_YQ_IMAGE" -p json -o json \
      'del(.. | select(tag == "!!map") | .attachments)' <"$src" >"$dst" 2>>"$EV/scan/yq.log" || rc=$?
    case "$rc" in
      0) return 0 ;;
      1) rm -f -- "$dst" || return 1; say "stage: $name cannot be sanitized (not JSON); it is not published"; return 0 ;;
      *) return 1 ;;
    esac
  fi
  cp -- "$src" "$dst"
}

do_stage() {
  local n f d
  rm -rf -- "$EV/staged.tmp" "$EV/staged"
  mkdir -m 700 -- "$EV/staged.tmp" || return 19
  for n in "${STAGE_REPORT[@]}"; do
    if stage_one "$EV/raw/report/$n" "$n"; then :; else rm -rf -- "$EV/staged.tmp"; say "stage: copying $n failed; nothing is staged"; return 19; fi
  done
  for f in "$EV/raw/report"/obs-*.tsv; do
    if [[ ! -e "$f" && ! -L "$f" ]]; then continue; fi
    n="${f##*/}"
    if [[ ! "$n" =~ ^obs-[a-z0-9-]{1,64}\.tsv$ ]]; then continue; fi
    if stage_one "$f" "$n"; then :; else rm -rf -- "$EV/staged.tmp"; say "stage: copying $n failed; nothing is staged"; return 19; fi
  done
  # Both inputs already passed raw_scan. Candidate scanning gates their final publication again.
  stage_one "$EV/raw/harness-console.log" harness.log || return 19
  # Failure diagnostics and (G064) the per-invocation Go test logs: regular files with safe names only.
  for d in diagnostics go-tests; do
    if [[ -d "$EV/raw/report/$d" && ! -L "$EV/raw/report/$d" ]]; then
      mkdir -p "$EV/staged.tmp/$d" || return 19
      for f in "$EV/raw/report/$d"/*; do
        [[ -e "$f" || -L "$f" ]] || continue
        n="${f##*/}"
        [[ "$n" =~ ^[A-Za-z0-9_.-]{1,120}$ && -f "$f" && ! -L "$f" ]] || return 19
        cp -- "$f" "$EV/staged.tmp/$d/$n" || return 19
      done
    fi
  done
  for n in "${STAGE_RAW[@]}"; do
    if stage_one "$EV/raw/$n" "$n"; then :; else rm -rf -- "$EV/staged.tmp"; say "stage: copying $n failed; nothing is staged"; return 19; fi
  done
  if [[ "${ST[harness]-}" =~ ^[0-9]+$ ]]; then
    if stage_one "$REPO_R/infra/validation/required-checks.tsv" required-checks.tsv; then :; else
      rm -rf -- "$EV/staged.tmp"; say "stage: copying required-checks.tsv failed; nothing is staged"; return 19
    fi
  fi
  mv -T -- "$EV/staged.tmp" "$EV/staged" || { rm -rf -- "$EV/staged.tmp"; return 19; }
  return 0
}

int_or_null() { # NAME VALUE: a validated integer, or null when unset; fails on a malformed value
  if [[ -z "$2" ]]; then printf 'null'; return 0; fi
  if [[ "$2" =~ ^(0|[1-9][0-9]{0,17})$ ]]; then printf '%s' "$2"; return 0; fi
  say "summary: $1 is malformed" >&2
  return 1
}
token_or_null() { # NAME VALUE REGEX
  if [[ -z "$2" ]]; then printf 'null'; return 0; fi
  if [[ "$2" =~ $3 ]]; then printf '"%s"' "$2"; return 0; fi
  say "summary: $1 is malformed" >&2
  return 1
}

# do_summary KIND MARKER REPORT: builds a fresh candidate (full: staged files; minimal: the marker) plus
# summary.json and phase-status.tsv, both pre-publication and made of validated fields only. REPORT is the
# JUnit phase that follows it (junit, or junit_withheld for the withheld rebuild).
do_summary() {
  local kind="$1" marker="$2" rp="$3" d="$EV/candidate" p v phases="" sep="" ci
  local pid piid jid rid mriid jto src evt sha ver rset rcount rsha pset psha psuites esuites cov part areas sdig projid rel
  pid="$(int_or_null CI_PIPELINE_ID "${CI_PIPELINE_ID:-}")" || return 19
  projid="$(int_or_null CI_PROJECT_ID "${CI_PROJECT_ID:-}")" || return 19
  piid="$(int_or_null CI_PIPELINE_IID "${CI_PIPELINE_IID:-}")" || return 19
  jid="$(int_or_null CI_JOB_ID "${CI_JOB_ID:-}")" || return 19
  rid="$(int_or_null CI_RUNNER_ID "${CI_RUNNER_ID:-}")" || return 19
  mriid="$(int_or_null CI_MERGE_REQUEST_IID "${CI_MERGE_REQUEST_IID:-}")" || return 19
  jto="$(int_or_null CI_JOB_TIMEOUT "${CI_JOB_TIMEOUT:-}")" || return 19
  src="$(token_or_null CI_PIPELINE_SOURCE "${CI_PIPELINE_SOURCE:-}" '^[a-z_]{1,64}$')" || return 19
  evt="$(token_or_null CI_MERGE_REQUEST_EVENT_TYPE "${CI_MERGE_REQUEST_EVENT_TYPE:-}" '^[a-z_]{1,32}$')" || return 19
  sha="$(token_or_null CI_COMMIT_SHA "${CI_COMMIT_SHA:-}" '^[0-9a-f]{40}$')" || return 19
  ver="$(token_or_null CI_SERVER_VERSION "${CI_SERVER_VERSION:-}" '^[0-9][0-9A-Za-z.+-]{0,63}$')" || return 19
  if [[ ! "$kind" =~ ^(full|minimal)$ ]]; then return 19; fi
  if [[ "$kind" == minimal && ! "$marker" =~ ^(WITHHELD|STAGING-FAILED)$ ]]; then return 19; fi
  if [[ ! "$rp" =~ ^(junit|junit_withheld)$ ]]; then return 19; fi
  if [[ -n "$HEAD_SHA" && ! "$HEAD_SHA" =~ ^[0-9a-f]{40}$ ]]; then return 19; fi
  if [[ -n "$TREE_SHA" && ! "$TREE_SHA" =~ ^[0-9a-f]{40}$ ]]; then return 19; fi
  if [[ -n "$STATUS_DIGEST" && ! "$STATUS_DIGEST" =~ ^[0-9a-f]{64}$ ]]; then return 19; fi
  if [[ -n "$BUDGET_S" && ! "$BUDGET_S" =~ ^[1-9][0-9]{0,5}$ ]]; then return 19; fi
  if [[ -n "$TESTED_CONTENT" && ! "$TESTED_CONTENT" =~ $TC_RE ]]; then return 19; fi
  # ONEOFF RC qualification: the verified promotion identity (null outside an RC promotion). The source and target are
  # the verified merge request branches (release/<tag> and the default branch); the peeled commit is the candidate HEAD.
  rel=null
  if [[ "$TESTED_CONTENT" == mr-rc-promotion ]]; then
    if [[ ! "$RELEASE_TAG" =~ $RC_TAG_RE || ! "$RELEASE_TAG_OBJECT" =~ ^[0-9a-f]{40}$ || ! "$RELEASE_STAGING_SHA" =~ ^[0-9a-f]{40}$ ||
      ! "$HEAD_SHA" =~ ^[0-9a-f]{40}$ || "$mriid" == null || "${CI_MERGE_REQUEST_SOURCE_BRANCH_NAME:-}" != "release/$RELEASE_TAG" ||
      ! "${CI_MERGE_REQUEST_TARGET_BRANCH_NAME:-}" =~ ^[A-Za-z0-9][A-Za-z0-9._/-]{0,254}$ ]]; then
      # Only a state import that already failed (state 16) can leave it inconsistent; then the summary records null.
      if [[ "${ST[state]-}" == 0 ]]; then say "summary: the RC promotion identity is malformed"; return 19; fi
    else rel="{\"tag\": \"$RELEASE_TAG\", \"tag_object\": \"$RELEASE_TAG_OBJECT\", \"peeled\": \"$HEAD_SHA\", \"staging_sha\": \"$RELEASE_STAGING_SHA\", \"merge_request_iid\": $mriid, \"source_branch\": \"release/$RELEASE_TAG\", \"target_branch\": \"$CI_MERGE_REQUEST_TARGET_BRANCH_NAME\"}"
    fi
  fi
  if [[ -n "$PROFILE_ARG" && ! "$PROFILE_ARG" =~ ^(fast|gate|staging|affected|integration|full)$ ]]; then return 19; fi
  if [[ -n "$PIPELINE_PROFILE" && ! "$PIPELINE_PROFILE" =~ $RE_PIPE ]]; then return 19; fi
  if [[ -n "$EXECUTED_PROFILE" && ! "$EXECUTED_PROFILE" =~ $RE_EXEC ]]; then return 19; fi
  if [[ -n "$SELECTION_REASON" && ! "$SELECTION_REASON" =~ ^[a-z][a-z-]{1,31}$ ]]; then return 19; fi
  if [[ -n "$COMPARISON_BASE" && ! "$COMPARISON_BASE" =~ ^([0-9a-f]{40}|none)$ ]]; then return 19; fi
  # G051 selection fields: validated tokens, or null before the selection was imported.
  psuites="$(token_or_null PIPELINE_SUITES "$PIPELINE_SUITES" "$RE_PSUITES")" || return 19
  esuites="$(token_or_null EXECUTED_SUITES "$EXECUTED_SUITES" "$RE_ESUITES")" || return 19
  cov="$(token_or_null COVERAGE "$COVERAGE" '^(complete|partial|none)$')" || return 19
  part="$(token_or_null PARTITION "$PARTITION" '^(fast|remainder|complete|none)$')" || return 19
  areas="$(token_or_null AREAS "$AREAS" "$RE_AREAS")" || return 19
  sdig="$(token_or_null SELECTION_DIGEST "$SELECTION_DIGEST" '^[0-9a-f]{64}$')" || return 19
  # required_count and required_set_sha256 cover E_job (the executed suites plus final); pipeline_set_sha256
  # covers E_pipeline (the pipeline suites, plus the fast suites when the fast job ran the fast partition).
  rcount=null rsha=null psha=null
  if [[ -n "$EXECUTED_SUITES" && "$EXECUTED_SUITES" != none ]]; then
    rset="$(required_ids "$EXECUTED_SUITES")" || return 19
    rcount="$(grep -c . <<<"$rset")" || rcount=0
    rsha="\"$(printf '%s\n' "$rset" | sha256sum | cut -c1-64)\""
  fi
  if [[ -n "$PIPELINE_SUITES" && -n "$PARTITION" ]]; then
    pset="$(required_ids "$(pipeline_suite_list)")" || return 19
    psha="\"$(printf '%s\n' "$pset" | sha256sum | cut -c1-64)\""
  fi
  if [[ ! "$HARNESS_STOP" =~ ^(none|deadline|signal|finalizer)$ ]]; then return 19; fi
  if [[ ! "$STARTED_UTC" =~ ^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z$ ]]; then return 19; fi
  for p in location recovery preflight state harness signal postcheck cleanup raw_scan complete stage internal; do
    if [[ -z "${ST[$p]+x}" ]]; then continue; fi
    v="${ST[$p]}"
    if [[ "$v" =~ ^[0-9]{1,3}$ ]]; then phases+="$sep\"$p\": $v"
    elif [[ "$v" =~ ^(skip|n/a|not-run|not-required)$ ]]; then phases+="$sep\"$p\": \"$v\""
    else say "summary: phase $p has a malformed status"; return 19; fi
    sep=", "
  done
  ci="{\"project_id\": $projid, \"pipeline_id\": $pid, \"pipeline_iid\": $piid, \"job_id\": $jid, \"runner_id\": $rid, \"merge_request_iid\": $mriid, \"job_timeout\": $jto, \"pipeline_source\": $src, \"merge_request_event_type\": $evt, \"commit_sha\": $sha, \"server_version\": $ver}"
  rm -rf -- "$d"
  mkdir -m 700 -- "$d" || return 19
  if [[ "$kind" == full ]]; then
    cp -R -- "$EV/staged/." "$d/" || return 19
  elif [[ "$marker" == WITHHELD ]]; then
    printf '%s\n' "$MARKER_WITHHELD" >"$d/WITHHELD" || return 19
  else
    printf '%s\n' "$MARKER_STAGING" >"$d/STAGING-FAILED" || return 19
  fi
  {
    printf '{\n  "format": "platform-validation-summary/1",\n  "status_kind": "pre-publication",\n'
    printf '  "authoritative": "the job exit status and the final job-log line",\n'
    printf '  "candidate": "%s",\n  "marker": %s,\n' "$kind" "$([[ "$kind" == minimal ]] && printf '"%s"' "$marker" || printf 'null')"
    printf '  "tested_content": %s,\n' "$([[ -n "$TESTED_CONTENT" ]] && printf '"%s"' "$TESTED_CONTENT" || printf 'null')"
    printf '  "release": %s,\n' "$rel"
    printf '  "profile": %s,\n  "pipeline_profile": %s,\n  "executed_profile": %s,\n' \
      "$([[ -n "$PROFILE_ARG" ]] && printf '"%s"' "$PROFILE_ARG" || printf '"full"')" \
      "$([[ -n "$PIPELINE_PROFILE" ]] && printf '"%s"' "$PIPELINE_PROFILE" || printf 'null')" \
      "$([[ -n "$EXECUTED_PROFILE" ]] && printf '"%s"' "$EXECUTED_PROFILE" || printf 'null')"
    printf '  "selection_reason": %s,\n  "comparison_base": %s,\n  "required_count": %s,\n  "required_set_sha256": %s,\n' \
      "$([[ -n "$SELECTION_REASON" ]] && printf '"%s"' "$SELECTION_REASON" || printf 'null')" \
      "$([[ -n "$COMPARISON_BASE" ]] && printf '"%s"' "$COMPARISON_BASE" || printf 'null')" "$rcount" "$rsha"
    printf '  "pipeline_suites": %s,\n  "pipeline_set_sha256": %s,\n  "executed_suites": %s,\n  "coverage": %s,\n' \
      "$psuites" "$psha" "$esuites" "$cov"
    printf '  "partition": %s,\n  "areas": %s,\n  "selection_digest": %s,\n' "$part" "$areas" "$sdig"
    printf '  "head_sha": %s,\n  "tree_sha": %s,\n  "status_digest": %s,\n' \
      "$([[ -n "$HEAD_SHA" ]] && printf '"%s"' "$HEAD_SHA" || printf 'null')" \
      "$([[ -n "$TREE_SHA" ]] && printf '"%s"' "$TREE_SHA" || printf 'null')" \
      "$([[ -n "$STATUS_DIGEST" ]] && printf '"%s"' "$STATUS_DIGEST" || printf 'null')"
    printf '  "budget_s": %s,\n  "harness_stop": "%s",\n  "stragglers": %s,\n' "${BUDGET_S:-null}" "$HARNESS_STOP" "$STRAGGLERS"
    printf '  "started_utc": "%s",\n  "summary_utc": "%s",\n' "$STARTED_UTC" "$(ci_utc)"
    printf '  "phases": {%s},\n' "$phases"
    printf '  "pending": ["summary", "%s", "candidate_scan", "publish", "report", "scrub"],\n' "$rp"
    printf '  "ci": %s\n}\n' "$ci"
  } >"$d/summary.json" || return 19
  {
    printf 'status_kind\tpre-publication\ncandidate\t%s\n' "$kind"
    for p in location recovery preflight state harness signal postcheck cleanup raw_scan complete stage internal; do
      if [[ -n "${ST[$p]+x}" ]]; then printf '%s\t%s\n' "$p" "${ST[$p]}"; fi
    done
    for p in summary "$rp" candidate_scan publish report scrub; do printf '%s\tpending\n' "$p"; done
  } >"$d/phase-status.tsv" || return 19
  return 0
}

# junit_check FILE TESTS FAILURES SUITE: the written report is exactly the generator's fixed shape (an XML
# declaration line, then one line of testsuites/testsuite/testcase/failure elements with only the allowed
# attributes, classes and fixed messages: no DOCTYPE, comment, other element, attribute or text), and it parses
# with the pinned yq (no network, no external resolution) as exactly one suite named SUITE whose count
# attributes and parsed testcase and failure elements equal the emitted TESTS and FAILURES.
JUNIT_SHAPE='^<testsuites tests="[0-9]+" failures="[0-9]+" errors="0" skipped="0"><testsuite name="[a-z]+:[a-z]+" tests="[0-9]+" failures="[0-9]+" errors="0" skipped="0">(<testcase classname="validate[.][a-z]+[.]timing" name="timing:validation-elapsed" time="[0-9]{1,5}"/>|<testcase classname="validate[.][a-z]+[.](check|phase)" name="[A-Za-z0-9:_-]+"(/>|><failure type="(FAIL|UNMET|phase-error)" message="(check failed|required prerequisite unmet|phase [a-z_]+ exited [0-9]+)"/></testcase>))*</testsuite></testsuites>$'
JUNIT_YQ_CHECK='[(keys | join(",")), (.testsuites."+@tests"), (.testsuites."+@failures"), (.testsuites."+@errors"),
  (.testsuites."+@skipped"), ([.testsuites.testsuite] | flatten(1) | length), (.testsuites.testsuite."+@name"),
  (.testsuites.testsuite."+@tests"), (.testsuites.testsuite."+@failures"), (.testsuites.testsuite."+@errors"),
  (.testsuites.testsuite."+@skipped"),
  ([.testsuites.testsuite.testcase] | flatten(1) | map(select(tag == "!!map")) | length),
  ([.testsuites.testsuite.testcase] | flatten(1) | map(select(tag == "!!map" and has("failure"))) | length),
  ([.testsuites.testsuite.testcase] | flatten(1) | map(select(tag == "!!map" and (."+@classname" | test("^validate[.](fast|gate|staging|full|local)[.](check|phase|timing)$")))) | length),
  ([.. | select(tag == "!!map") | keys | .[] | select(. == "+content" or . == "+directive")] | length)]'
junit_check() {
  local got want n=0 line
  while IFS= read -r line || [[ -n "$line" ]]; do
    n=$((n + 1))
    if (( n == 1 )) && [[ "$line" == '<?xml version="1.0" encoding="UTF-8"?>' ]]; then continue; fi
    if (( n == 2 )) && [[ "$line" =~ $JUNIT_SHAPE ]]; then continue; fi
    return 1
  done <"$1" || return 1
  if (( n != 2 )); then return 1; fi
  # Non-additive timing: at most one case carries a time (the measured validation elapsed seconds).
  local doc rest
  doc="$(<"$1")" || return 1
  rest="${doc//' time="'/}"
  if (( (${#doc} - ${#rest}) / 7 > 1 )); then return 1; fi
  got="$(wrapper_run --rm --pull never --network none -i "$CI_YQ_IMAGE" -p xml -o tsv "$JUNIT_YQ_CHECK" <"$1" 2>>"$EV/scan/yq.log")" || return 1
  want="$(printf '+p_xml,testsuites\t%s\t%s\t0\t0\t1\t%s\t%s\t%s\t0\t0\t%s\t%s\t%s\t0' "$2" "$3" "$4" "$2" "$3" "$2" "$3" "$2")"
  [[ "$got" == "$want" ]]
}

# do_junit KIND MARKER (G050): the supplemental candidate/junit.xml from validated fields only (infra/ci/junit.sh).
# The suite is JOB:PROFILE, PROFILE the executed profile, or remainder for a gate's partition (G051).
# Check cases come from the sanitized status columns of a full or minimal STAGING-FAILED candidate whose
# completeness step ran; a WITHHELD candidate gets phase cases only. A converter error or a failed self-check
# leaves no junit.xml and returns 19; the rest of the candidate is unchanged.
do_junit() {
  local kind="$1" marker="$2" out="$EV/candidate/junit.xml" cols="" prof="${EXECUTED_PROFILE:-none}" counts p
  local -a ph=()
  if [[ -L "$EV/candidate" || ! -d "$EV/candidate" ]]; then return 19; fi
  # G051: a gate's partial coverage is reported as gate:remainder, never as the pipeline profile.
  if [[ "$PARTITION" == remainder ]]; then prof=remainder; fi
  if [[ "$kind" == full || "$marker" == STAGING-FAILED ]] && [[ "${ST[complete]-}" =~ ^[0-9]+$ ]]; then
    cols="$EV/final/status-columns.tsv"
  fi
  for p in "${PHASES[@]}"; do
    if [[ -n "${ST[$p]+x}" ]]; then ph+=("$p=${ST[$p]}"); fi
  done
  local duration_file="$EV/final/junit-durations.tsv" phase_name
  : >"$duration_file" || return 19
  if [[ "$kind" == full && -f "$EV/raw/report/timings.tsv" && ! -L "$EV/raw/report/timings.tsv" ]]; then
    awk -F'\t' -v suites=" $EXECUTED_SUITES final " '
      $1=="suite" && NF==4 && $2 ~ /^[a-z][a-z0-9-]*$/ && index(suites," " $2 " ") &&
      $3 ~ /^[0-9]+$/ && length($3)<=12 && $4 ~ /^[0-9]+$/ && length($4)<=12 && $4 >= $3 && $4-$3 <= 86400 {
        if(!seen[$2]++) print "suite\t" $2 "\t" $4-$3
      }' "$EV/raw/report/timings.tsv" >>"$duration_file" || return 19
  fi
  for phase_name in "${PHASES[@]}"; do
    if [[ "${PHASE_SECONDS[$phase_name]:-}" =~ ^[0-9]{1,5}$ ]]; then
      printf 'phase\t%s\t%s\n' "$phase_name" "${PHASE_SECONDS[$phase_name]}" >>"$duration_file" || return 19
    fi
  done
  # The one GitLab-total time (the timing case): this run's measured elapsed seconds so far, one wall-clock
  # interval, never a sum of the overlapping suite/phase rows above (diagnostic only). It replaces the former
  # per-suite timing cases, so it is reported exactly when executed-suite timing rows exist.
  if grep -q '^suite' "$duration_file" && [[ "$SECONDS" =~ ^[0-9]{1,5}$ ]]; then
    printf 'elapsed\tvalidation\t%s\n' "$SECONDS" >>"$duration_file" || return 19
  fi
  counts="$(JUNIT_DURATIONS_FILE="$duration_file" junit_report "$out" "$JOB_KIND" "$prof" "$cols" "${PHASES[*]}" "${ph[@]}")" || {
    rm -f -- "$out" "$out.tmp"; say "junit: the report inputs are invalid; no junit.xml"; return 19; }
  if [[ ! "$counts" =~ ^[0-9]+\ [0-9]+$ ]]; then rm -f -- "$out"; say "junit: the report counts are malformed; no junit.xml"; return 19; fi
  if junit_check "$out" "${counts% *}" "${counts#* }" "$JOB_KIND:$prof"; then :; else
    rm -f -- "$out"; say "junit: the written report failed its self-check; no junit.xml"; return 19
  fi
  say "junit: junit.xml with ${counts% *} cases, ${counts#* } failing (supplementary)"
  return 0
}

# junit_discard PHASE (parent, in finalize): after a failed report phase no junit.xml or junit.xml.tmp may stay
# in the candidate; when one cannot be removed, nothing is published (finalize's kind becomes none).
junit_discard() {
  local d="$EV/candidate"
  if [[ "${ST[$1]-}" == 0 ]]; then return 0; fi
  rm -f -- "$d/junit.xml" "$d/junit.xml.tmp" 2>/dev/null || true
  if [[ -e "$d/junit.xml" || -L "$d/junit.xml" || -e "$d/junit.xml.tmp" || -L "$d/junit.xml.tmp" ]]; then
    say "junit: a failed report could not be removed; nothing is published"
    kind=none
  fi
  return 0
}

do_publish() { # CI only: copy the scanned candidate to ci-evidence.tmp, then rename it into place
  local dst="$PUBLISH_TARGET" tmp="$PUBLISH_TARGET.tmp"
  if [[ -z "$dst" ]]; then return 19; fi
  if [[ -e "$dst" || -L "$dst" || -e "$tmp" || -L "$tmp" ]]; then say "publish: the publication path already exists; nothing is published"; return 19; fi
  if cp -R -- "$EV/candidate" "$tmp" 2>/dev/null && mv -T -- "$tmp" "$dst" 2>/dev/null; then return 0; fi
  if [[ -e "$tmp" ]]; then rm -rf -- "$tmp" 2>/dev/null || true; fi
  say "publish: failed; nothing is published"
  return 19
}

keep_local_candidate() { # retain only the already scanned artifact set, never raw logs or secrets
  (( CI_MODE == 0 )) && [[ -n "$EVARG" && -d "$EV/candidate" && ! -L "$EV/candidate" ]] &&
    [[ "${ST[candidate_scan]-}" == 0 || "${ST[rescan]-}" == 0 ]]
}

do_report() { # job log: phase lines, then the validated check columns only if both scans were clean
  local p id st
  if keep_local_candidate; then
    say "evidence: scanned local report retained at $EV/candidate (logs, check ledger and supplemental JUnit when available)"
  fi
  for p in "${PHASES[@]}"; do
    if [[ -n "${ST[$p]+x}" ]]; then say "phase=$p status=${ST[$p]}"; fi
  done
  if [[ "${ST[raw_scan]-}" == 0 && "${ST[candidate_scan]-}" == 0 && -f "$EV/final/status-columns.tsv" ]]; then
    if [[ "$PUBLISHED" == full && -f "$PUBLISH_TARGET/harness.log" ]]; then
      say "diagnostics: scanned harness.log is available in job artifacts; failing runs also retain bounded diagnostics/ logs"
    fi
    while IFS=$'\t' read -r id st; do
      if [[ "$id" =~ ^[A-Z][A-Z0-9-]*$ && "$st" =~ ^(PASS|FAIL|UNMET|INFO)$ ]]; then say "check $id $st"; fi
    done <"$EV/final/status-columns.tsv"
  fi
  # G051 cleanup amendment: the storage summary, only from scanned evidence.
  if [[ "${ST[raw_scan]-}" == 0 ]] && [[ "${ST[candidate_scan]-n/a}" =~ ^(0|n/a)$ || "${ST[rescan]-}" == 0 ]]; then storage_summary
  else say "storage: report withheld or not scanned (see the scan phases)"; fi
  return 0
}

do_scrub() { # the secret list always goes; the registry stays whenever cleanup did not verify
  local rc=0
  if [[ -e "$REG/secrets" || -L "$REG/secrets" ]]; then rm -f -- "$REG/secrets" || return 19; fi
  if [[ -e "$REG/secrets" || -L "$REG/secrets" ]]; then say "scrub: the generated-secret list survived"; return 19; fi
  # Local --evidence-dir is an explicit retention request. Keep the same scanned candidate that CI
  # publishes, including failure diagnostics; scrubbing it used to leave only storage/recovery reports.
  # A withheld/minimal candidate contains only its safe summary, never the unsafe raw evidence.
  if keep_local_candidate; then :; else rm -rf -- "$EV/candidate" 2>/dev/null || rc=19; fi
  rm -rf -- "$EV/raw" "$EV/staged" "$EV/staged.tmp" "$EV/final" "$EV/state" "$EV/scan" 2>/dev/null || rc=19
  if [[ "${ST[cleanup]-}" == 0 ]]; then
    rm -rf -- "$REG" 2>/dev/null || rc=19
  fi
  if [[ -z "$EVARG" ]]; then
    if [[ "$REG" == "$EV"/* && ( -e "$REG" || -L "$REG" ) ]]; then
      say "scrub: keeping $EV, reduced to the retained registry"
    else
      rm -rf -- "$EV" 2>/dev/null || rc=19
    fi
  fi
  if (( rc != 0 )); then say "scrub: some evidence could not be deleted"; fi
  return "$rc"
}

# --- aggregation and the authoritative final line (parent; reads ST, sets FINAL, never mutates ST) -----
aggregate() {
  local p v any=0
  FINAL=0
  for p in "${PHASES[@]}"; do
    v="${ST[$p]-}"
    if [[ "$v" =~ ^[0-9]+$ ]] && (( v != 0 )); then any=1; fi
  done
  for p in location recovery preflight state; do
    v="${ST[$p]-}"
    if [[ "$v" =~ ^[0-9]+$ ]] && (( v != 0 )); then FINAL=16; break; fi
  done
  v="${ST[harness]-}"
  if (( FINAL == 0 )) && [[ "$v" =~ ^[0-9]+$ ]] && (( v != 0 )); then FINAL="$v"; fi
  v="${ST[signal]-}"
  if (( FINAL == 0 )) && [[ "$v" =~ ^[0-9]+$ ]] && (( v != 0 )); then FINAL="$v"; fi
  if (( FINAL == 0 )); then
    if [[ "${ST[raw_scan]-}" == 17 || "${ST[candidate_scan]-}" == 17 || "${ST[rescan]-}" == 17 ]]; then FINAL=17
    elif [[ "${ST[raw_scan]-}" == 15 || "${ST[candidate_scan]-}" == 15 || "${ST[rescan]-}" == 15 ]]; then FINAL=15
    elif [[ "${ST[complete]-0}" =~ ^[0-9]+$ && "${ST[complete]-0}" != 0 ]]; then FINAL=14
    elif [[ "${ST[postcheck]-0}" =~ ^[0-9]+$ && "${ST[postcheck]-0}" != 0 ]]; then FINAL=13
    elif [[ "${ST[cleanup]-0}" =~ ^[0-9]+$ && "${ST[cleanup]-0}" != 0 ]]; then FINAL=18
    fi
  fi
  # No-masking invariant: a recorded nonzero status, or an unfinished main body, never yields 0.
  if (( FINAL == 0 && any )); then FINAL=19; fi
  if (( FINAL == 0 && MAIN_DONE != 1 )); then FINAL=19; fi
  return 0
}

final_line() {
  local p list=""
  for p in "${PHASES[@]}"; do
    if [[ -n "${ST[$p]+x}" ]]; then list+=" $p=${ST[$p]}"; fi
  done
  printf 'validate: final status %s; published=%s; registry=%s; harness_stop=%s; phases:%s\n' \
    "$FINAL" "$PUBLISHED" "$REGSTATE" "$HARNESS_STOP" "$list"
}

finalize_prelocation() { # L0 failed: no Docker operation and no evidence write has happened
  ST[location]="${ST[location]:-16}"
  phase postcheck do_postcheck_repo_only
  ST[cleanup]="n/a"
  if (( REG_CREATED && REG_PROVEN )) && [[ -n "$REG" && "$REG" != "$EV"/* ]]; then rm -rf -- "$REG" 2>/dev/null || true; fi
  if (( EV_CREATED && EV_PROVEN )) && [[ -n "$EV" ]]; then rm -rf -- "$EV" 2>/dev/null || true; fi
  REGSTATE=none PUBLISHED=none
  aggregate
  final_line
  exit "$FINAL"
}

finalize() {
  local kind reason
  if (( FINALIZED )); then return 0; fi
  FINALIZED=1
  set +eu
  trap '' TERM INT HUP USR1
  stop_watchdog
  if [[ -n "$HPID" ]]; then
    say "harness: still running at finalization; stopping its process group"
    stop_harness_group
    ST[harness]=143; HARNESS_STOP=finalizer; HPID=""
  fi
  # Provenance is recorded here, in the parent, before anything is aggregated or printed.
  if (( LOC_OK )) && [[ -z "$SIG_NAME" ]] && (( MAIN_DONE == 0 )); then ST[internal]=19; fi
  if (( LOC_OK == 0 )); then finalize_prelocation; fi
  cd -- "$REPO_R" || true
  if [[ -z "${ST[harness]+x}" ]]; then ST[harness]=not-run; fi
  phase postcheck do_postcheck
  phase cleanup do_cleanup
  # G051 cleanup amendment: storage after cleanup, whatever the run's status (not a phase; never gates).
  storage_observe post
  # The G051 bootstrap record (not a phase: it never gates; a missing or non-reuse record leaves AC5 UNMET).
  bootstrap_post || say "bootstrap: no ledger-bootstrap.tsv was written"
  phase raw_scan do_raw_scan
  if [[ "${ST[harness]}" =~ ^[0-9]+$ ]]; then phase complete do_complete; else ST[complete]="n/a"; fi
  if [[ "${ST[raw_scan]}" == 0 ]]; then phase stage do_stage; else ST[stage]=skip; fi
  # Candidate selection (parent): full | minimal | none.
  kind=none reason=""
  if [[ "${ST[raw_scan]}" == 17 ]]; then kind=none
  elif [[ "${ST[raw_scan]}" == 0 && "${ST[stage]}" == 0 ]]; then kind=full
  elif [[ "${ST[raw_scan]}" == 15 ]]; then kind=minimal reason=WITHHELD
  elif [[ "${ST[raw_scan]}" == 0 ]]; then kind=minimal reason=STAGING-FAILED
  fi
  if [[ "$kind" != none ]]; then
    phase summary do_summary "$kind" "$reason" junit
    if [[ "${ST[summary]}" != 0 ]]; then kind=none; rm -rf -- "$EV/candidate" 2>/dev/null || true; fi
  fi
  # The supplemental JUnit report joins the candidate before it is scanned; a report defect is 19, never a
  # reason to withhold the rest of the candidate.
  if [[ "$kind" != none ]]; then phase junit do_junit "$kind" "$reason"; junit_discard junit; fi
  if [[ "$kind" != none ]]; then
    phase candidate_scan do_candidate_scan
    case "${ST[candidate_scan]}" in
      0) ;;
      15)
        if [[ "$kind" == full ]]; then
          kind=minimal
          phase summary_withheld do_summary minimal WITHHELD junit_withheld
          if [[ "${ST[summary_withheld]}" == 0 ]]; then
            phase junit_withheld do_junit minimal WITHHELD; junit_discard junit_withheld
            phase rescan do_candidate_scan
            if [[ "${ST[rescan]}" != 0 ]]; then kind=none; fi
          else
            kind=none
          fi
        else
          kind=none
        fi ;;
      *) kind=none ;;
    esac
  fi
  if [[ "$kind" == none ]]; then rm -rf -- "$EV/candidate" 2>/dev/null || true; fi
  if (( CI_MODE )) && [[ -n "$PUBLISH_TARGET" && "$kind" != none ]]; then
    phase publish do_publish
    if [[ "${ST[publish]}" == 0 ]]; then PUBLISHED="$kind"; fi
  fi
  # G064 P5: an evidence-helper container that was not verified gone is a cleanup failure (18; registry and marker kept).
  if [[ -s "$EV/state/wrapper-unverified" ]]; then
    say "cleanup: $(grep -c . "$EV/state/wrapper-unverified") evidence-helper container(s) not verified removed: $(cut -f2,3 "$EV/state/wrapper-unverified" | tr '\t\n' ' ;' | cut -c1-300)"
    cp -- "$EV/state/wrapper-unverified" "$REG/wrapper-unverified.tsv" 2>/dev/null || true
    if [[ "${ST[cleanup]-}" == 0 ]]; then ST[cleanup]=18; fi
  fi
  storage_keep
  phase report do_report
  phase scrub do_scrub
  if [[ -e "$REG" || -L "$REG" ]]; then REGSTATE="kept:$REG"; else REGSTATE=removed; fi
  marker_finish
  aggregate
  final_line
  exit "$FINAL"
}

# --- run ---------------------------------------------------------------------------------------------------
trap finalize EXIT
trap 'on_signal TERM 15' TERM
trap 'on_signal INT 2' INT
trap 'on_signal HUP 1' HUP
trap 'DEADLINE=1' USR1

if locate; then
  LOC_OK=1; ST[location]=0
else
  ST[location]=16
  exit 16
fi
say "location: evidence $EV; registry $REG"
# G051 cleanup amendment: a local run first resolves every unfinalized earlier local run, then marks itself.
if (( CI_MODE == 0 )); then
  phase recovery do_recovery
  if [[ "${ST[recovery]}" == 0 ]]; then
    if marker_write; then :; else ST[recovery]=16; say "recovery: the run marker cannot be written in $PEND_DIR"; fi
  fi
fi
storage_observe pre
if [[ "${ST[recovery]-0}" == 0 ]]; then phase preflight do_preflight; fi
if [[ "${ST[preflight]-}" == 0 ]]; then
  if state_import "$EV/state/p1.env" 'HEAD_SHA=^[0-9a-f]{40}$' 'TREE_SHA=^[0-9a-f]{40}$' \
    'STATUS_DIGEST=^[0-9a-f]{64}$' 'BUDGET_S=^[1-9][0-9]{0,5}$' \
    "TESTED_CONTENT=$TC_RE" "RELEASE_TAG=^(none|$RC_BODY)\$" 'RELEASE_TAG_OBJECT=^(none|[0-9a-f]{40})$' \
    'RELEASE_STAGING_SHA=^(none|[0-9a-f]{40})$' "${SEL_SPECS[@]}"; then
    ST[state]=0
  else
    ST[state]=16
  fi
  # The RC identity exists exactly for an RC promotion (all three values), never otherwise.
  if [[ "${ST[state]}" == 0 ]]; then
    if [[ "$TESTED_CONTENT" == mr-rc-promotion ]]; then
      if [[ "$RELEASE_TAG" == none || "$RELEASE_TAG_OBJECT" == none || "$RELEASE_STAGING_SHA" == none ]]; then ST[state]=16; fi
    elif [[ "$RELEASE_TAG/$RELEASE_TAG_OBJECT/$RELEASE_STAGING_SHA" != none/none/none ]]; then ST[state]=16; fi
    if [[ "${ST[state]}" != 0 ]]; then say "state: the RC identity does not match the tested content"; fi
  fi
  # Re-assert the empty-gate invariant and the partition (G051) in the parent.
  if [[ "${ST[state]}" == 0 && "$EXECUTED_PROFILE" == none ]] && [[ ! "$PIPELINE_PROFILE" =~ ^(fast|affected)$ || "$JOB_KIND" != gate ]]; then ST[state]=16; fi
  if [[ "${ST[state]}" == 0 ]]; then if partition_ok "$TESTED_CONTENT"; then :; else ST[state]=16; say "state: the selection is not one consistent partition"; fi; fi
  if [[ "${ST[state]}" == 0 ]]; then
    say "tested_content=$TESTED_CONTENT head=$HEAD_SHA"
    if [[ "$TESTED_CONTENT" == mr-rc-promotion ]]; then
      say "release tag=$RELEASE_TAG tag_object=$RELEASE_TAG_OBJECT peeled=$HEAD_SHA staging=$RELEASE_STAGING_SHA merge_request=${CI_MERGE_REQUEST_IID:-}"
    fi
    say "profile pipeline=$PIPELINE_PROFILE executed=$EXECUTED_PROFILE reason=$SELECTION_REASON base=$COMPARISON_BASE"
    say "suites pipeline=[$PIPELINE_SUITES] executed=[$EXECUTED_SUITES] coverage=$COVERAGE partition=$PARTITION areas=$AREAS"
    if [[ "$EXECUTED_PROFILE" == none ]]; then
      ST[harness]=not-required
      say "harness: not required (the pipeline profile is $PIPELINE_PROFILE; the fast job ran it)"
    else
      run_harness
    fi
  fi
fi
MAIN_DONE=1
