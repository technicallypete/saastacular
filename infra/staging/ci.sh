#!/usr/bin/env bash
# G072 CI adapter: manual staging/preview start/deploy and stop from protected staging pipelines (docs/foundation/
# STAGING.md "CI start/deploy and stop"; Plan §10). Trusted control code: it runs only from the pipeline's protected
# staging revision, through infra/ci/staging-action.sh, and drives control staging.sh (G070) against a pinned candidate.
#
#   bash infra/staging/ci.sh prepare|check|build|deploy|stop          GitLab jobs (launcher contract, Plan §10.10)
#   bash infra/staging/ci.sh lease --config FILE --acquire            local handoff: hold the environment lease and
#                                                                     print STAGING_LEASE_HELD=<holder>@<generation>
#   bash infra/staging/ci.sh lease --config FILE --show|--release|--break --confirm ENV --reason TEXT   (staging.sh lease)
#   bash infra/staging/ci.sh meta pull|push --config FILE             durable metadata <-> EVIDENCE_DIR (push: lease held)
#   bash infra/staging/ci.sh lineage init --config FILE --candidate SHA     setup only: assumed lineage of the VM's release
#   bash infra/staging/ci.sh lineage reconcile --config FILE                observed lineage -> verified, fence cleared
#
# Jobs (environment selected by STAGING_CI_ENV, one line of infra/staging/ci-environments.txt):
#   prepare  credential-free (the launcher fetched refs/heads/*): pin the candidate (empty or "staging" = the pipeline commit for every
#            environment; staging: a full SHA on first-parent staging history; preview: also a same-repository branch
#            or a full SHA on one), the infrastructure-compatibility gate (S6), candidate.tsv, schema.tsv and a git
#            bundle of the pinned commit as artifacts.
#   check    read-only identity <env>-ci-read: STAGING_REUSE=yes when a valid published manifest of the candidate with
#            all seven digests in the registry exists, else no (dotenv reuse.env). Nothing is written.
#   build    credential-free: STAGING_REUSE=yes writes a skipped marker; no builds the candidate with control
#            staging.sh build --source <candidate worktree> under env -i and saves the images, manifest, checksums.
#   deploy   start identity (impersonates <env>-deployer): phase A local checks, phase B authenticated reads and gates
#            (supersession, lease, schema lineage set, manifest or build artifacts, VM state, duration, >= 300 s left),
#            phase C writes (lease, OS Login key, images, start when stopped, deploy with the schema fence, readiness,
#            observed lineage, receipt, metadata, lease release).
#   stop     stop identity <env>-ci-stop: stop request, lease, control staging.sh stop, receipt, current.tsv, release.
#            Needs no candidate, build, artifact or earlier pipeline.
# Every staging.sh subprocess writes only to STAGING_CI_PRIVATE/raw/<step>.log; the job log gets "ci: " lines and the
# fixed-field summary (STAGING_CI_OUT/summary.txt); durable logs are publish_filter output only (ci-publish.sh).
# Exit (S14): 0 success; 2 refused before any change; 3 HOLD; 4 failed after a change; 5 the session expired during the
# action; 6 privacy canary or isolation finding (outputs withheld).
set -euo pipefail
umask 077
CI_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)"
# shellcheck source=lib.sh
source "$CI_DIR/lib.sh"
# shellcheck source=meta-lib.sh
source "$CI_DIR/meta-lib.sh"
# shellcheck source=ci-publish.sh
source "$CI_DIR/ci-publish.sh"
# shellcheck source=ci-lib.sh
source "$CI_DIR/ci-lib.sh"
REPO="$(cd "$CI_DIR/../.." && pwd -P)"
STAGING_SH=$CI_DIR/staging.sh
exec 3>&2

# ------------------------------------------------------------------------------------------------------------- jobs
ci_prepare() {
  job_setup prepare
  require_credential_free
  resolve_candidate # the launcher fetched refs/heads/* as refs/remotes/origin/*; the adapter does not fetch
  compat_gate "$CAND"
  git_q update-ref refs/g072/candidate "$CAND" || die 2 "the candidate ref could not be set"
  git_q bundle create "$OUT/candidate.bundle" refs/g072/candidate >/dev/null || die 2 "the candidate bundle could not be created"
  schema_of "$REPO" "$CAND" >"$OUT/schema.tsv" || die 2 "the candidate's migrations could not be listed"
  printf 'candidate\t%s\nselection\t%s\nbranch\t%s\nactor\t%s\npipeline\t%s\ncontrol\t%s\nenvironment\t%s\n' "$CAND" "$SELECTION" \
    "${BRANCH:--}" "$GITLAB_USER_LOGIN" "$CI_PIPELINE_ID" "$CTL" "$ENVSEL" >"$OUT/candidate.tsv"
  say "candidate $CAND pinned for $ENVSEL (selection ${SELECTION:-default}${BRANCH:+, branch $BRANCH}, actor $GITLAB_USER_LOGIN)"
}

ci_check() {
  job_setup check
  profile_load
  verify_prepare
  cred_setup
  meta_require
  denial_probes
  manifest_check
  if ((MAN_OK)); then REUSE=yes; else REUSE=no; fi
  printf 'STAGING_REUSE=%s\n' "$REUSE" >"$OUT/reuse.env"
  SUM[reuse]="$REUSE${MAN_WHY:+ ($MAN_WHY)}"
  say "reuse decision for $CAND: $REUSE${MAN_WHY:+ ($MAN_WHY)}"
}

ci_build() {
  job_setup build
  require_credential_free
  case ${STAGING_REUSE:-} in
    yes)
      printf 'skipped\ta valid published manifest exists (staging:check); nothing built\n' >"$OUT/skipped"
      SUM[reuse]="yes (build skipped)"
      say "build skipped: staging:check found a valid published manifest"
      return 0
      ;;
    no) ;;
    *) die 2 "STAGING_REUSE must be yes or no (the staging:check dotenv)" ;;
  esac
  verify_prepare
  build_images
  say "built the seven images of $CAND (credential-free; artifacts with checksums)"
}

ci_deploy() {
  # Phase A: local only.
  job_setup deploy
  profile_load
  verify_prepare
  [[ ${STAGING_REUSE:-} == yes || ${STAGING_REUSE:-} == no ]] || die 2 "STAGING_REUSE must be yes or no (the staging:check dotenv)"
  verify_build
  # Phase B: authenticated, read-only.
  cred_setup
  meta_require
  denial_probes
  deploy_gates
  # Phase C: writes, only after every gate; the gates run again against what the lease now protects.
  ci_lease_take deploy
  deploy_gates
  WRITES=1 # from here a failure also publishes its sanitized log and run record
  oslogin_key || die 4 "the OS Login key could not be generated or registered"
  deploy_images
  start_if_needed
  remaining_guard
  CHANGED=1
  run_staging deploy deploy --candidate "$CAND" "${PINS_ARGS[@]}"
  ((STAGING_RC == 0)) || fail_class deploy
  readiness_check
  verify_observed_lineage
  vm_session || true
  SUM[session]="${SESSION_HAVE:-unknown}s" SUM[scheduled_stop]=${SESSION_STOP:-unreadable}
  run_staging receipt receipt --candidate "$CAND"
  ((STAGING_RC == 0)) || die 4 "staging.sh receipt failed (exit $STAGING_RC)"
  receipt_publish "$RAW/receipt.log"
  publish_logs
  cur_write "environment=$ENVSEL" "candidate=$CAND" "selection=${SELECTION:-default}" "branch=${BRANCH:--}" \
    "actor=$GITLAB_USER_LOGIN" "pipeline=$CI_PIPELINE_ID" "job=$CI_JOB_ID" "manifest_sha256=${MAN_SHA:--}" \
    "manifest_generation=$MAN_GEN" "schema=schema/$CAND.tsv" "status_log=${SUM[status_log]:--}" \
    "session_seconds=${SESSION_HAVE:--}" "scheduled_stop=${SESSION_STOP:--}" "started_at=$(ts)" "stopped_at=-" "state=RUNNING" \
    "receipt=${SUM[receipt_id]:--}" "last_action=deploy" "last_pipeline=$CI_PIPELINE_ID" "last_job=$CI_JOB_ID" \
    "last_actor=$GITLAB_USER_LOGIN"
  run_record "exit 0"
  lease_release_own || die 4 "the environment lease could not be released (staging.sh lease --break after review)"
  say "deployed $CAND to $ENVSEL: READY, digests = manifest, scheduled stop ${SESSION_STOP:-unreadable}"
}

ci_stop() {
  # Phase A: local only (no candidate, no duration, no artifact).
  job_setup stop
  profile_load
  # Phase B: authenticated, read-only.
  cred_setup
  meta_require
  denial_probes
  lease_read
  vm_read || die 3 "HOLD: the state of $VM could not be read; nothing changed"
  # Writes: the stop request first (also when a lease then refuses), then the lease.
  stop_request
  ci_lease_gate
  ci_lease_take stop
  WRITES=1
  # Best effort (Spec R5): without a key the guest-side stop fails and staging.sh still issues the provider stop.
  if [[ $VMSTATE == RUNNING ]] && ! oslogin_key; then say "no OS Login key: the guest-side stop is skipped; the provider stop follows"; fi
  run_staging stop stop
  case $STAGING_RC in
    0) ;;
    3) die 3 "HOLD: staging.sh stop reported a HOLD (see the published log)" ;;
    2) die 2 "staging.sh stop refused (see the published log)" ;;
    *) die 4 "staging.sh stop failed (exit $STAGING_RC)" ;;
  esac
  CHANGED=1
  vm_read && [[ $VMSTATE == TERMINATED || $VMSTATE == STOPPED ]] || die 4 "$VM is not observed stopped after the stop"
  SUM[readiness]="stopped ($VMSTATE observed)"
  receipt_publish "$RAW/stop.log"
  publish_logs
  cur_write "state=$VMSTATE" "stopped_at=$(ts)" "receipt=${SUM[receipt_id]:--}" "last_action=stop" "last_pipeline=$CI_PIPELINE_ID" \
    "last_job=$CI_JOB_ID" "last_actor=$GITLAB_USER_LOGIN"
  run_record "exit 0"
  runs_overwrite_probe
  lease_release_own || die 4 "the environment lease could not be released (staging.sh lease --break after review)"
  say "stopped $ENVSEL: $VM observed $VMSTATE; receipt ${SUM[receipt_id]:-unavailable}"
}

# ------------------------------------------------------------------------------------------------------ local commands
# local_conf ARGS...: --config FILE (and the given extra options) for the operator's local subcommands; their own
# ADC impersonating the deployer, as every G070 local command. OPTS holds the extras.
declare -A OPTS=()
local_conf() {
  local allowed=$1
  shift
  while (($#)); do
    [[ " $allowed " == *" $1 "* ]] || die 2 "this subcommand does not take $1"
    case $1 in
      --acquire | --show | --release | --break) OPTS[${1#--}]=1 ;;
      --config | --candidate | --confirm | --reason)
        (($# >= 2)) || die 2 "$1 needs a value"
        OPTS[${1#--}]=$2
        shift
        ;;
    esac
    shift
  done
  [[ -n ${OPTS[config]:-} ]] || die 2 "--config FILE is required"
  load_conf "${OPTS[config]}"
  meta_require
  PRIV=$META_TMP
  trap 'lease_release_own || true; meta_cleanup' EXIT
}

local_lease() { # COMMAND: the lease for one local subcommand (STAGING_LEASE_HELD when the operator holds it)
  if [[ -n ${STAGING_LEASE_HELD:-} ]]; then lease_check_held "$STAGING_LEASE_HELD"; else lease_acquire "$(local_holder)" "$1" - ""; fi
}

ci_lease_local() {
  local h
  if [[ " $* " != *" --acquire "* ]]; then exec bash "$STAGING_SH" lease "$@"; fi
  local_conf "--config --acquire" "$@"
  h="$(local_holder)"
  lease_acquire "${h%:*}" handoff - ""
  LEASE_OWN=0 # held across commands until staging.sh lease --release
  printf 'STAGING_LEASE_HELD=%s@%s\n' "$LEASE_HOLDER" "$LEASE_GEN"
}

ci_meta() {
  local sub=${1:-} k d f c sha key rc conflicts=0
  shift || true
  case $sub in
    pull)
      local_conf "--config" "$@"
      d=${CONF[EVIDENCE_DIR]}/ci-meta
      install -d -m 0700 "$d" "${CONF[EVIDENCE_DIR]}/deploy"
      for k in current.tsv lineage.tsv fence.tsv lease; do
        meta_read "$k" "$d/$k" || die 2 "$k could not be read ($(meta_class))"
        printf 'meta\tpulled\t%s\tgeneration %s\n' "$k" "$META_GEN"
      done
      c="$(awk -F'\t' '$1 == "candidate" {print $2}' "$d/current.tsv")"
      if is_sha "$c"; then
        meta_read "deploy/$c.tsv" "$d/manifest.tsv" || die 2 "deploy/$c.tsv could not be read"
        f=${CONF[EVIDENCE_DIR]}/deploy/$c.tsv
        if ((META_GEN > 0)); then
          if [[ -f $f ]]; then cmp -s "$f" "$d/manifest.tsv" || die 2 "the local manifest of $c differs from the published one"
          else install -m 0600 "$d/manifest.tsv" "$f"; fi
          printf 'meta\tmanifest\t%s\n' "$f"
        fi
      fi
      ;;
    push)
      local_conf "--config" "$@"
      [[ -n ${STAGING_LEASE_HELD:-} ]] || die 2 "meta push needs the environment lease: ci.sh lease --acquire, then export STAGING_LEASE_HELD"
      lease_check_held "$STAGING_LEASE_HELD"
      for f in "${CONF[EVIDENCE_DIR]}"/deploy/*.tsv; do
        [[ -f $f ]] || continue
        sha=${f##*/}; sha=${sha%.tsv}
        is_sha "$sha" && manifest_valid "$f" "$sha" || { printf 'meta\tskipped\t%s\t%s\n' "${f##*/}" "not a pushed manifest"; continue; }
        meta_read "deploy/$sha.tsv" "$PRIV/remote.tsv" || die 2 "deploy/$sha.tsv could not be read"
        if ((META_GEN == 0)); then
          put_new "$f" "deploy/$sha.tsv" || die 2 "deploy/$sha.tsv could not be written ($(meta_class))"
          printf 'meta\tpushed\tdeploy/%s.tsv\n' "$sha"
        elif cmp -s "$f" "$PRIV/remote.tsv"; then printf 'meta\tpresent\tdeploy/%s.tsv\n' "$sha"
        else printf 'meta\tconflict\tdeploy/%s.tsv\n' "$sha"; conflicts=$((conflicts + 1)); fi
      done
      for f in "${CONF[EVIDENCE_DIR]}"/receipts/*.tsv; do
        [[ -f $f ]] || continue
        key="$(receipt_key "$(basename "$f" .tsv)" "$f")"
        rc=0
        put_new "$f" "$key" || rc=$?
        ((rc <= 1)) || die 2 "$key could not be written ($(meta_class))"
        printf 'meta\t%s\t%s\n' "$([[ $rc == 0 ]] && echo pushed || echo present)" "$key"
      done
      ((conflicts == 0)) || die 2 "$conflicts published manifest(s) differ from the local ones; nothing overwritten"
      ;;
    *) die 2 "usage: ci.sh meta pull|push --config FILE" ;;
  esac
}

ci_lineage() {
  local sub=${1:-} c last obs rows v blobs n
  shift || true
  case $sub in
    init)
      local_conf "--config --candidate" "$@"
      c=${OPTS[candidate]:-}
      check_candidate "$c"
      local_lease lineage-init
      meta_gen lineage.tsv || die 2 "lineage.tsv could not be read ($(meta_class))"
      ((GEN == 0)) || die 2 "lineage.tsv exists (generation $GEN): init is setup only; use lineage reconcile"
      last="$({ find "${CONF[EVIDENCE_DIR]}/logs" -name '*-status.log' 2>/dev/null || true; } | LC_ALL=C sort | tail -n 1)"
      if [[ -n $last ]]; then
        [[ "$(awk -F'\t' '$1 == "candidate" {print $2}' "$last")" == "$c" ]] ||
          die 2 "the last VM status log ($last) names another candidate than $c"
      else
        log "no local VM status log: $c is recorded as given"
      fi
      rows=$PRIV/rows
      schema_of "$REPO" "$c" | awk -F'\t' '$1 ~ /^[0-9]+$/' >"$rows"
      lineage_write assumed "$c" "$rows" 0 || die 2 "lineage.tsv could not be written ($(meta_class))"
      printf 'lineage\tassumed\t%s\t%s versions\n' "$c" "$(grep -c . "$rows")"
      ;;
    reconcile)
      local_conf "--config" "$@"
      local_lease lineage-reconcile
      vm_read && [[ $VMSTATE == RUNNING ]] || die 2 "lineage reconcile needs the running VM ($VM is ${VMSTATE:-unreadable})"
      members_load "$PRIV/lineage"
      obs=$PRIV/observed rows=$PRIV/rows
      observed_versions "$obs" || die 3 "HOLD: the observed lineage could not be read"
      : >"$rows"
      while IFS= read -r v; do
        blobs="$(for ((n = 1; n <= MEMBERS; n++)); do awk -F'\t' -v v="$v" '$1 == v {print $2}' "$PRIV/lineage/member-$n.tsv"; done | sort -u)"
        [[ -n $blobs ]] || die 3 "HOLD: applied version $v is in no recorded lineage or fence entry: an owner decision (data is discarded only by a separately authorized staging.sh reset --confirm)"
        [[ $blobs != *$'\n'* ]] || die 3 "HOLD: applied version $v has different migration files across the recorded candidates: an owner decision"
        printf '%s\t%s\n' "$v" "$blobs" >>"$rows"
      done <"$obs"
      lineage_write verified reconcile "$rows" "$LIN_GEN" || die 2 "lineage.tsv changed or could not be written ($(meta_class))"
      if ((FENCE_GEN > 0)); then meta_rm fence.tsv "$FENCE_GEN" || die 2 "fence.tsv changed or could not be deleted ($(meta_class))"; fi
      printf 'lineage\tverified\t%s versions\tfence cleared (%s entries)\n' "$(grep -c . "$rows" || true)" "$FENCE_N"
      ;;
    *) die 2 "usage: ci.sh lineage init --config FILE --candidate SHA | lineage reconcile --config FILE" ;;
  esac
}

main() {
  local a=${1:-}
  (($#)) && shift
  case $a in
    prepare | check | build | deploy | stop)
      (($# == 0)) || die 2 "ci.sh $a takes no arguments (the launcher passes the contract in the environment)"
      trap ci_exit EXIT
      trap ci_signal INT TERM HUP
      "ci_$a"
      ;;
    lease) ci_lease_local "$@" ;;
    meta) ci_meta "$@" ;;
    lineage) ci_lineage "$@" ;;
    *) die 2 "usage: ci.sh prepare|check|build|deploy|stop | lease|meta|lineage ... (see the header)" ;;
  esac
}

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then main "$@"; fi
