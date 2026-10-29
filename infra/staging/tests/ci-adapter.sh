#!/usr/bin/env bash
# CI-ADAPTER (G072 Plan §10.4, §10.6, §10.8, §10.13 AC2/AC4/AC5): infra/staging/ci.sh against fake tools only
# (tests/fakes): a fixture origin repository with a staging history and same-repository branches, job clones, a fake
# Cloud Storage with generation preconditions and the ci-meta permission model, a fake VM, registry, Terraform and
# Docker. Every call is classified read/write; refusals are asserted to write nothing. No network, Docker or cloud.
#
#   bash infra/staging/tests/ci-adapter.sh --work DIR --report DIR [--only CASE|GROUP[,...]]
#
# Groups (--only GROUP runs every GROUP-* case): fence (partial-deploy lineage: fence kept, older candidate refused
# with zero writes, cancellation, successor verifies, reconcile verified/refused/HOLD, a local divergent deploy),
# order (each phase-A and phase-B refusal exits 2 with zero writes; phase A with zero cloud calls), cand (candidate
# policy: defaults, first-parent, intermediate/unmerged/fork refused, branch pinned through a deleted branch, S6),
# reuse (automatic reuse with no previous pipeline artifacts, check/deploy disagreement, build path, image-ID
# mismatch), life (stopped/running starts, duration and token use, stop without artifacts, after a failed or canceled
# start, three cycles, supersession, rc 255 after the deadline = exit 5, idempotent stop, stop under a local lease),
# cred (credential-file shapes, cleanup, GCLOUD_IMPERSONATE_SA validation, privacy canary), control (mutations: each
# protected assertion fails against a broken copy).
# Ledger REPORT/summary.tsv; last line RESULT PASS|FAIL cases=N failed=M; exit 0 all pass, 1 any failure, 64 usage.
# Case helpers share state through dynamic scope with begin/chk/done_case (their reads are not visible to shellcheck);
# the optional arguments of pipeline/stop_pipeline are forwarded to job.
# shellcheck disable=SC2034,SC2120
set -uo pipefail
# shellcheck source=fakes/fixture.sh
source "$(dirname "${BASH_SOURCE[0]}")/fakes/fixture.sh"

fx_args "$@"
FX_PERMS="$(fx_perms_source)"
FAKE_ROOT=$W/fake
FX_JWT="eyJG072FAKEHEADER0.eyJG072FAKEPAYLOAD0.c2lnbmF0dXJlZmFrZQ"
PROVIDER=projects/123456789012/locations/global/workloadIdentityPools/ci-pool/providers/gitlab
PINS="IDENTITY_WORKOS_CLIENT_ID=none,IDENTITY_WORKOS_ISSUER=none,IDENTITY_WORKOS_API_KEY=none"

# ---- fixture origin: staging history with a merged feature (first parent M1 -> C0), branches from M1 ----------------
O=$W/origin
fx_repo "$O"
C0=$FX_BASE
gitc() { git -C "$O" "$@" >/dev/null 2>&1; }
commit_file() { # BRANCH FROM PATH CONTENT MESSAGE -> HEAD
  gitc switch -q -C "$1" "$2"
  mkdir -p "$(dirname "$O/$3")"
  printf '%s\n' "$4" >"$O/$3"
  gitc add -A
  gitc commit -q -m "$5"
  git -C "$O" rev-parse HEAD
}
A1="$(commit_file feat-a "$C0" core/a.txt a feature-a)"
gitc switch -q staging
gitc merge -q --no-ff feat-a -m merge-feature-a
M1="$(git -C "$O" rev-parse HEAD)"
gitc branch -D feat-a
X="$(commit_file x "$M1" core/migrations/00025_m25.sql '-- migration 25' x-25)"
Y="$(commit_file y "$X" core/migrations/00026_m26.sql '-- migration 26' y-26)"
Z="$(commit_file z "$M1" core/migrations/00024_m24.sql '-- migration 24 diverged' z-divergent-24)"
P1="$(commit_file preview-feature "$M1" core/preview.txt preview preview-change)"
I1="$(commit_file infra-change "$M1" infra/compose/local.yml 'services: {changed: true}' infra-change)"
U1="$(commit_file unmerged "$M1" core/u.txt u unmerged)"
F1="$(commit_file fork-tmp "$M1" core/fork.txt fork fork-only)"
gitc update-ref refs/merge-requests/7/head "$F1"
gitc switch -q staging
gitc branch -D fork-tmp
CTL=$M1
JOB=$W/job
git clone -q --branch staging "$O" "$JOB" 2>/dev/null
git -C "$JOB" checkout -q --detach "$CTL"
FAKE_REPO=$JOB
PIPE=100 # pipeline IDs grow with each simulated pipeline

# ---- profiles, worlds and jobs ---------------------------------------------------------------------------------------
set_env() { # ENV: the fixture environment of the following jobs (staging or preview-1)
  FX_ENV=$1 FX_BUCKET=$FX_PROJECT_ID-$1-ci-meta FX_AR=us-central1-docker.pkg.dev/$FX_PROJECT_ID/$1
  JENV=$1 JCAND="" JDUR=""
  local p=$W/profile-$1 f
  rm -rf "$p" "$p-check"; mkdir -p "$p" "$p-check"
  fx_conf "$p/full.conf" "$W/unused-evidence" "$p"
  grep -v -E '^(EVIDENCE_DIR|PRIVATE_DIR|GOOGLE_ADC_FILE)=' "$p/full.conf" >"$p/staging.conf"
  printf 'SESSION_MAX_SECONDS=7200\n' >>"$p/staging.conf"
  rm -f "$p/full.conf" "$p/adc.json"
  printf 'CI_WIF_PROVIDER=%s\nCI_IDENTITY_PINS=%s\n' "$PROVIDER" "$PINS" >"$p/ci-settings"
  for f in staging.conf ci-settings; do cp "$p/$f" "$p-check/$f"; done
  PROFILE=$p
  printf "cid=\$(sudo docker ps --quiet --filter label=com.docker.compose.project=%s --filter label=com.docker.compose.service=postgres | head -n 1) && test -n \"\$cid\" && sudo docker exec \"\$cid\" psql --no-psqlrc --tuples-only --no-align --username postgres --dbname example_app --command 'SELECT DISTINCT version_id FROM app.goose_db_version WHERE version_id <> 0 AND is_applied ORDER BY version_id'" \
    "$1" >"$W/lineage-query.$1"
}

schema_rows() { git -C "$JOB" ls-tree "$1" core/migrations/ | sed -nE 's#^[0-9]+ blob ([0-9a-f]+)\tcore/migrations/0*([0-9]+)_[^/]*\.sql$#\2\t\1#p' | sort -n; }

# world [TERMINATED]: fresh fake world; lineage.tsv verified at the control's 24 migrations (L0); VM running the
# control with an hour left (or stopped).
world() {
  fx_state
  cp "$W/lineage-query.$FX_ENV" "$FAKE_ROOT/lineage-query"
  { printf 'state\tverified\nsource\t%s\nwriter\tfixture\ntime\t-\n' "$CTL"; schema_rows "$CTL"; } >"$W/l0.tsv"
  fx_put lineage.tsv "$W/l0.tsv"
  seq 1 24 >"$FAKE_ROOT/vm/db_versions"
  printf '%s\n' "$CTL" >"$FAKE_ROOT/vm/candidate"
  if [[ ${1:-} == TERMINATED ]]; then printf 'TERMINATED\n' >"$FAKE_ROOT/vm/status"; : >"$FAKE_ROOT/vm/term"; fi
}

# publish_manifest SHA: a valid pushed manifest of SHA in the bucket (deploy/<sha>.tsv) and its digests in the registry.
publish_manifest() { fx_manifest "$W/m-$1.tsv" "$1"; fx_put "deploy/$1.tsv" "$W/m-$1.tsv"; }

# job ACTION [VAR=VALUE...]: ci.sh ACTION as a job of pipeline PIPE (artifact siblings under W/ns/p<PIPE>/), with
# the launcher contract; check, deploy and stop get an ID token file and their profile and federate as their action.
job() {
  local action=$1 jid priv out envs=() p
  shift
  case $action in prepare) jid=1 ;; check) jid=2 ;; build) jid=3 ;; deploy) jid=4 ;; *) jid=5 ;; esac
  jid=$((PIPE * 10 + jid))
  # The launcher's artifact directories (staging-action.sh): staging-prepare/-check/-build, staging-summary for deploy/stop.
  case $action in deploy | stop) out=$W/ns/p$PIPE/staging-summary ;; *) out=$W/ns/p$PIPE/staging-$action ;; esac
  priv=$W/ns/.staging-ci-$jid
  rm -rf "$priv"; mkdir -p "$priv" "$W/ns/p$PIPE"; chmod 0700 "$priv"
  envs=("STAGING_CI_ACTION=$action" "STAGING_CI_ENV=$JENV" "STAGING_CI_CANDIDATE_INPUT=$JCAND" "STAGING_CI_DURATION_INPUT=$JDUR"
    "STAGING_CI_PRIVATE=$priv" "STAGING_CI_OUT=$out" "CI_PIPELINE_ID=$PIPE" "CI_JOB_ID=$jid" "CI_COMMIT_SHA=$CTL"
    "GITLAB_USER_LOGIN=fixture-user")
  case $action in
    check | deploy | stop)
      printf '%s\n' "$FX_JWT" >"$priv/id-token"; chmod 0600 "$priv/id-token"
      p=$PROFILE; [[ $action != check ]] || p=$PROFILE-check
      envs+=("CI_ENVIRONMENT_NAME=$JENV" "STAGING_CI_ID_TOKEN_FILE=$priv/id-token" "STAGING_CI_PROFILE_DIR=$p")
      case $action in check) envs+=(FAKE_BASE_PRINCIPAL=federated-read) ;; deploy) envs+=(FAKE_BASE_PRINCIPAL=federated-start) ;; stop) envs+=(FAKE_BASE_PRINCIPAL=federated-stop) ;; esac
      ;;
  esac
  if [[ $action == build || $action == deploy ]] && [[ -f $W/ns/p$PIPE/staging-check/reuse.env ]]; then
    envs+=("$(grep '^STAGING_REUSE=' "$W/ns/p$PIPE/staging-check/reuse.env")")
  fi
  MARK="$(wc -l <"$FAKE_ROOT/calls.log")"
  fx_run "p$PIPE-$action" "${envs[@]}" "$@" -- bash "${JOBDIR:-$JOB}/infra/staging/ci.sh" "$action"
  JPRIV=$priv JOUT=$out
  # Every failing job keeps its stdout, stderr and call-log slice (refusals included; the scratch W is removed at exit).
  if ((RC != 0)); then
    JOBSEQ=$((JOBSEQ + 1))
    mkdir -p "$REPORT/jobs"
    cp "$OUT" "$REPORT/jobs/$JOBSEQ.$BASHPID-p$PIPE-$action.out"
    cp "$ERR" "$REPORT/jobs/$JOBSEQ.$BASHPID-p$PIPE-$action.err"
    since >"$REPORT/jobs/$JOBSEQ.$BASHPID-p$PIPE-$action.calls"
    printf '%s\t%s\tp%s-%s\texit %s\n' "$JOBSEQ.$BASHPID" "${CASE_HINT:-}" "$PIPE" "$action" "$RC" >>"$REPORT/jobs/index.tsv"
  fi
}
JOBSEQ=0 CASE_HINT=""

since() { tail -n +"$((MARK + 1))" "$FAKE_ROOT/calls.log"; }
since_writes() { since | awk -F'\t' '$2 == "W"'; }
since_cloud() { since | awk -F'\t' '$1 == "gcloud" || $1 == "curl"'; }

# pipeline: prepare, check, build, deploy of a new pipeline (stops at the first failing job; RC = its exit).
pipeline() {
  local a
  PIPE=$((PIPE + 1))
  for a in prepare check build deploy; do
    job "$a" "$@"
    ((RC == 0)) || return 0
  done
}
stop_pipeline() { PIPE=$((PIPE + 1)); local c=$JCAND d=$JDUR; JCAND="" JDUR=""; job stop "$@"; JCAND=$c JDUR=$d; }

expect() { # CASE OK DETAIL-ON-PASS FAIL-REASONS
  if [[ $2 == 1 ]]; then record "$1" 1 "$3"; else record "$1" 0 "$4"; fi
}
why=""
chk() { eval "$1" || why+="$2; "; } # chk 'CONDITION' 'REASON'
begin() { why=""; }
done_case() { # CASE PASS-DETAIL
  keep "$1"
  if [[ -z $why ]]; then record "$1" 1 "$2"; else record "$1" 0 "$why (last: $(tail -n 2 "$ERR" | tr '\n' ' ' | cut -c1-200); failing jobs: jobs/index.tsv)"; fi
  CASE_HINT="after $1"
}
lastjob_rc() { cat "$W/runs/p$PIPE-$1.rc"; }
zero_writes() { [[ -z "$(since_writes)" ]]; }
fence_has() { fx_cat fence.tsv | awk -F'\t' -v c="$1" '$1 == "fence" && $4 == c' | grep -q .; }
lineage_source() { fx_cat lineage.tsv | awk -F'\t' '$1 == "source" {print $2}'; }
lineage_state() { fx_cat lineage.tsv | awk -F'\t' '$1 == "state" {print $2}'; }

# ======================================================================================== fence (Plan §10.4)
fence_cases() {
  local g0 g1
  set_env preview-1
  # fence-partial: X (+25) deploys; the VM applies 25, readiness fails: exit 4, fence [X] kept, lineage still L0.
  begin; world
  printf 'fail-readiness\n' >"$FAKE_ROOT/vm/deploy_mode"
  g0="$(fx_gen lineage.tsv)"
  JCAND=$X; pipeline
  chk '[[ $RC == 4 ]]' "deploy exit $RC, not 4"
  chk 'fence_has $X' "no fence entry for X"
  chk '[[ "$(fx_gen lineage.tsv)" == "$g0" && "$(lineage_source)" == "$CTL" ]]' "lineage changed"
  chk '[[ "$(tail -n 1 "$FAKE_ROOT/vm/db_versions")" == 25 ]]' "the fake VM did not apply 25"
  chk '[[ "$(fx_gen lease)" == 0 ]]' "lease left after the failed deploy"
  done_case fence-partial "readiness failed after 25 applied: exit 4, fence [X] kept, lineage L0 unchanged, lease released"
  # fence-older-refused: W (= the control, 1..24) is refused with zero writes (no lease, key, push, compute, ssh, upload).
  begin
  rm -f "$FAKE_ROOT/vm/deploy_mode"
  JCAND=""; PIPE=$((PIPE + 1)); job prepare; job check; job build; job deploy
  chk '[[ $RC == 2 ]]' "deploy exit $RC, not 2"
  chk 'zero_writes' "writes: $(since_writes | cut -f3 | cut -c1-50 | tr '\n' ' ')"
  chk 'grep -q "schema fence unreconciled" "$ERR" "$W/runs/p$PIPE-deploy.out"' "refusal does not name the fence"
  done_case fence-older-refused "an older candidate after a partial deploy is refused exit 2 with zero write calls"
  # fence-cancel: the job is killed between the fence write and readiness (process group TERM): the fence stays.
  begin; world
  : >"$FAKE_ROOT/vm/block-deploy"
  JCAND=$X; PIPE=$((PIPE + 1)); job prepare; job check; job build
  cancel_job deploy deploy-blocked block-deploy
  chk '[[ "$(cat "$W/runs/p$PIPE-deploy.rc")" == 143 ]]' "canceled deploy exit $(cat "$W/runs/p$PIPE-deploy.rc" 2>/dev/null), not 143"
  chk '[[ "$(fx_gen lease)" == 0 ]]' "the canceled job left its lease"
  chk 'fence_has $X' "no fence entry after the cancellation"
  chk '[[ "$(lineage_source)" == "$CTL" ]]' "lineage changed by a canceled deploy"
  JCAND=""; PIPE=$((PIPE + 1)); job prepare; job check; job build; job deploy
  chk '[[ $RC == 2 ]] && zero_writes' "older candidate after the cancellation: exit $RC, writes $(since_writes | nlines)"
  done_case fence-cancel "a deploy canceled after its fence write keeps the fence; the older candidate is then refused with zero writes"
  # fence-successor: Y (= X + 26) is accepted; success verifies Y's lineage and deletes the fence.
  begin
  JCAND=$Y; pipeline
  chk '[[ $RC == 0 ]]' "deploy of Y exit $RC"
  chk '[[ "$(lineage_state)" == verified && "$(lineage_source)" == "$Y" ]]' "lineage not verified at Y"
  chk '[[ "$(fx_gen fence.tsv)" == 0 ]]' "fence not deleted"
  chk 'fx_cat lineage.tsv | grep -q "^26	"' "lineage lacks 26"
  done_case fence-successor "a successor (X+26) is accepted; the observed lineage verifies Y and the fence is deleted"
  # fence-reconcile-l0: fence [X] whose VM-side run applied nothing (observed 1..24): reconcile -> verified L0; W accepted.
  begin; world
  printf 'fail-before\n' >"$FAKE_ROOT/vm/deploy_mode"
  JCAND=$X; pipeline
  chk '[[ $RC == 4 ]] && fence_has $X' "failed X deploy: exit $RC"
  rm -f "$FAKE_ROOT/vm/deploy_mode"
  operator_reconcile
  chk '[[ $RC == 0 ]]' "reconcile exit $RC"
  chk '[[ "$(lineage_state)" == verified && "$(fx_gen fence.tsv)" == 0 ]]' "reconcile did not verify and clear"
  JCAND=""; pipeline
  chk '[[ $RC == 0 ]]' "W after reconcile: exit $RC"
  done_case fence-reconcile-l0 "observed 1..24 reconciles to verified L0 (fence cleared); the older candidate is then accepted"
  # fence-reconcile-25: observed 1..25 -> verified 1..25 (X's blob); W still refused.
  begin; world
  printf 'fail-readiness\n' >"$FAKE_ROOT/vm/deploy_mode"
  JCAND=$X; pipeline
  rm -f "$FAKE_ROOT/vm/deploy_mode"
  operator_reconcile
  chk '[[ $RC == 0 ]] && fx_cat lineage.tsv | grep -q "^25	"' "reconcile with 25 applied: exit $RC"
  JCAND=""; pipeline
  chk '[[ $RC == 2 ]] && zero_writes' "W after reconcile at 25: exit $RC"
  done_case fence-reconcile-25 "observed 1..25 reconciles to verified 1..25; the older candidate stays refused"
  # fence-reconcile-unknown: an applied version in no member (27) is HOLD 3; nothing written.
  begin; world
  printf 'fail-readiness\n' >"$FAKE_ROOT/vm/deploy_mode"
  JCAND=$X; pipeline
  { seq 1 25; echo 27; } >"$FAKE_ROOT/vm/observed"
  g1="$(fx_gen fence.tsv)"
  operator_reconcile
  chk '[[ $RC == 3 ]]' "reconcile exit $RC, not 3"
  chk '[[ "$(fx_gen fence.tsv)" == "$g1" && "$(lineage_source)" == "$CTL" ]]' "HOLD changed lineage or fence"
  done_case fence-reconcile-unknown "an applied version absent from every member HOLDs (exit 3); lineage and fence unchanged"
  # fence-reconcile-divergent: fence [Z] (24 diverges from L0), observed 1..24 -> two blobs for 24: HOLD 3.
  begin; world
  printf 'fail-before\n' >"$FAKE_ROOT/vm/deploy_mode"
  local_deploy "$Z"
  chk 'fence_has $Z' "the local deploy of Z added no fence entry"
  operator_reconcile
  chk '[[ $RC == 3 ]]' "reconcile exit $RC, not 3"
  done_case fence-reconcile-divergent "an applied version with different blobs across members HOLDs (exit 3)"
  # fence-local-divergent: a local staging.sh deploy of Z appends a fence entry; CI then refuses Y and W.
  begin; world
  rm -f "$FAKE_ROOT/vm/deploy_mode"
  local_deploy "$Z"
  chk '[[ $RC == 0 ]] && fence_has $Z' "local deploy of Z: exit $RC"
  JCAND=$Y; pipeline
  chk '[[ $RC == 2 ]] && zero_writes' "Y after the local Z deploy: exit $RC"
  JCAND=""; pipeline
  chk '[[ $RC == 2 ]] && zero_writes' "W after the local Z deploy: exit $RC"
  done_case fence-local-divergent "a local deploy of divergent Z fences it; CI refuses incompatible candidates with zero writes"
}


# operator_reconcile / local_deploy: the local operator (own ADC impersonating the deployer).
OPCONF=""
op_conf() {
  OPCONF=$W/operator/$FX_ENV.conf
  rm -rf "$W/operator/ev-$FX_ENV"
  fx_conf "$OPCONF" "$W/operator/ev-$FX_ENV" "$W/operator/pr-$FX_ENV"
}
operator_reconcile() {
  op_conf
  MARK="$(wc -l <"$FAKE_ROOT/calls.log")"
  fx_run "op-reconcile-$PIPE" FAKE_BASE_PRINCIPAL=operator -- bash "$JOB/infra/staging/ci.sh" lineage reconcile --config "$OPCONF"
}
local_deploy() {
  op_conf
  fx_manifest "$W/operator/ev-$FX_ENV/deploy/$1.tsv" "$1"
  MARK="$(wc -l <"$FAKE_ROOT/calls.log")"
  fx_run "op-deploy-$1" FAKE_BASE_PRINCIPAL=operator -- bash "$JOB/infra/staging/staging.sh" deploy --candidate "$1" --config "$OPCONF"
}

# ================================================================================ order (Plan §10.6 zero writes)
# refusal CASE PHASE REASON: the last job exited 2 naming REASON, with zero writes; phase A also with zero cloud calls
# (no token exchange).
refusal() {
  begin
  chk '[[ $RC == 2 ]]' "exit $RC, not 2"
  REASON=$3
  chk 'grep -qF -- "$REASON" "$ERR"' "the refusal does not say [$3]"
  chk 'zero_writes' "writes: $(since_writes | cut -f3 | cut -c1-60 | tr '\n' ' ')"
  [[ $2 != A ]] || chk '[[ -z "$(since_cloud)" ]]' "phase-A refusal made cloud calls: $(since_cloud | cut -f3 | cut -c1-40 | tr '\n' ' ')"
  done_case "$1" "exit 2 ($3) with zero writes$([[ $2 == A ]] && echo ' and zero cloud calls (phase A)')"
}

order_cases() {
  local p
  set_env staging
  world
  # phase A
  JCAND="-bad"; PIPE=$((PIPE + 1)); job prepare; refusal order-candidate-malformed A "candidate input is malformed"
  JCAND=""; JDUR="5x"; PIPE=$((PIPE + 1)); job prepare; refusal order-duration-malformed A "duration input is malformed"
  JDUR="300"; PIPE=$((PIPE + 1)); job prepare; job deploy; refusal order-duration-range A "duration takes"
  JDUR=""
  JENV=preview-10; PIPE=$((PIPE + 1)); job prepare; refusal order-selector-unknown A "ci-environments.txt"
  JENV=Preview-1; PIPE=$((PIPE + 1)); job prepare; refusal order-selector-case A "ci-environments.txt"
  JENV=staging
  # a good prepare/check/build for the deploy-side phase-A cases
  PIPE=$((PIPE + 1)); job prepare; job check; job build; p=$PIPE
  job deploy CI_ENVIRONMENT_NAME=preview-1; refusal order-env-name-mismatch A "CI_ENVIRONMENT_NAME"
  sed -i "s/^STAGING_ENV=staging$/STAGING_ENV=preview-1/; s/staging-deployer@/preview-1-deployer@/; s|/staging$|/preview-1|" "$PROFILE/staging.conf"
  job deploy; refusal order-profile-mismatch A "the profile is for"
  set_env staging; world
  PIPE=$p
  sed -i 's/^CI_IDENTITY_PINS=.*/CI_IDENTITY_PINS=IDENTITY_WORKOS_CLIENT_ID=none,IDENTITY_WORKOS_ISSUER=none/' "$PROFILE/ci-settings"
  job deploy; refusal order-pin-missing A "lacks IDENTITY_WORKOS_API_KEY"
  sed -i 's/^CI_IDENTITY_PINS=.*/CI_IDENTITY_PINS=IDENTITY_WORKOS_CLIENT_ID=latest,IDENTITY_WORKOS_ISSUER=none,IDENTITY_WORKOS_API_KEY=none/' "$PROFILE/ci-settings"
  job deploy; refusal order-pin-latest A "never latest"
  sed -i 's/^CI_IDENTITY_PINS=.*/CI_IDENTITY_PINS=IDENTITY_WORKOS_CLIENT_ID=3,IDENTITY_WORKOS_ISSUER=none,IDENTITY_WORKOS_API_KEY=4/' "$PROFILE/ci-settings"
  job deploy; refusal order-pin-mixed A "or all three to none"
  set_env staging
  sed -i '/^SESSION_MAX_SECONDS=/d' "$PROFILE/staging.conf"; printf 'SESSION_MAX_SECONDS=3600\n' >>"$PROFILE/staging.conf"
  job deploy; refusal order-session-default A "SESSION_MAX_SECONDS=7200"
  set_env staging
  printf 'EVIDENCE_DIR=/tmp/x\n' >>"$PROFILE/staging.conf"
  job deploy; refusal order-profile-path A "must not set EVIDENCE_DIR"
  set_env staging
  # bundle/selection mismatch: candidate.tsv edited to another selection
  cp "$W/ns/p$p/staging-prepare/candidate.tsv" "$W/ct.keep"
  sed -i "s/^selection\t.*/selection\t$Y/" "$W/ns/p$p/staging-prepare/candidate.tsv"
  job deploy; refusal order-selection-mismatch A "another selection"
  sed -i "s/^candidate\t.*/candidate\t$Y/; s/^selection\t.*/selection\t/" "$W/ns/p$p/staging-prepare/candidate.tsv"
  job deploy; refusal order-bundle-mismatch A "head is not the recorded candidate"
  cp "$W/ct.keep" "$W/ns/p$p/staging-prepare/candidate.tsv"
  # S6 at deploy time: hand-made prepare artifacts for the infrastructure-changing commit (bypassing prepare's gate)
  JENV=preview-1; set_env preview-1; world; JCAND=$I1
  PIPE=$((PIPE + 1)); mkdir -p "$W/ns/p$PIPE/staging-prepare"
  git -C "$JOB" update-ref refs/g072/candidate "$I1"
  git -C "$JOB" bundle create "$W/ns/p$PIPE/staging-prepare/candidate.bundle" refs/g072/candidate 2>/dev/null
  printf 'candidate\t%s\nselection\t%s\nbranch\t-\nactor\tx\npipeline\t%s\ncontrol\t%s\nenvironment\tpreview-1\n' "$I1" "$I1" "$PIPE" "$CTL" \
    >"$W/ns/p$PIPE/staging-prepare/candidate.tsv"
  printf 'STAGING_REUSE=no\n' >"$W/ns/reuse-no.env"; mkdir -p "$W/ns/p$PIPE/staging-check"; cp "$W/ns/reuse-no.env" "$W/ns/p$PIPE/staging-check/reuse.env"
  job deploy; refusal order-s6-deploy A "incompatible candidate"
  set_env staging; world
  # phase B
  PIPE=$((PIPE + 1)); job prepare; job check; job build; p=$PIPE
  printf 'pipeline\t%s\n' "$((p + 50))" >"$W/sr"; fx_put "stop-requests/$((p + 50))" "$W/sr"
  job deploy; refusal order-superseded B "superseded"
  world; fx_lease "local:op@host:1" none
  job deploy; refusal order-lease-local B "held by local:op@host:1"
  world; fx_lease "ci:1/2" "$((FX_NOW + 600))"
  job deploy; refusal order-lease-ci-live B "held by ci:1/2"
  world; JDUR=1h
  PIPE=$((PIPE + 1)); job prepare; job check; job build; job deploy; refusal order-running-duration B "runs a 7200s session"
  JDUR=""; world; printf '%s\n' "$((FX_NOW + 200))" >"$FAKE_ROOT/vm/term"
  PIPE=$((PIPE + 1)); job prepare; job check; job build; job deploy; refusal order-remaining B "remain before the scheduled stop"
  world
  PIPE=$((PIPE + 1)); job prepare; job check
  printf 'STAGING_REUSE=yes\n' >"$W/ns/p$PIPE/staging-check/reuse.env"; job build
  job deploy; refusal order-no-manifest-no-build B "no valid published manifest"
  set_env preview-1; world; JCAND=$Z
  PIPE=$((PIPE + 1)); job prepare; job check; job build; job deploy; refusal order-lineage B "retained schema incompatible"
  JCAND=""
}

# ===================================================================================== cand (S5/§10.2, S6)
cand_case() { # CASE ENV SELECTION EXPECT_RC [EXPECT_SHA]
  begin
  set_env "$2"; world; JCAND=$3
  PIPE=$((PIPE + 1)); job prepare
  EXP_RC=$4 EXP_SHA=${5:-}
  chk '[[ $RC == "$EXP_RC" ]]' "prepare exit $RC, not $4"
  if [[ $4 == 0 ]]; then
    chk 'grep -qx "candidate	$EXP_SHA" "$JOUT/candidate.tsv"' "pinned $(awk -F'\t' '$1 == "candidate" {print $2}' "$JOUT/candidate.tsv" 2>/dev/null), not $5"
  else
    chk '[[ -z "$(since_cloud)" && ! -e $JOUT/candidate.bundle ]]' "a refused prepare left calls or artifacts"
  fi
  done_case "$1" "$6"
}

cand_cases() {
  cand_case cand-default-staging staging "" 0 "$CTL" "empty candidate = the frozen pipeline commit (staging)"
  cand_case cand-default-preview preview-1 "" 0 "$CTL" "empty candidate = the frozen pipeline commit (preview-1)"
  cand_case cand-staging-word preview-1 staging 0 "$CTL" "candidate staging = the pipeline commit"
  cand_case cand-staging-first-parent staging "$C0" 0 "$C0" "a full SHA on first-parent staging history is accepted for staging"
  cand_case cand-staging-intermediate staging "$A1" 2 "" "an intermediate (merged feature) commit is refused for staging"
  cand_case cand-staging-unmerged staging "$U1" 2 "" "an unmerged branch SHA is refused for staging"
  cand_case cand-staging-branch staging preview-feature 2 "" "a branch name is refused for staging"
  cand_case cand-preview-sha preview-1 "$U1" 0 "$U1" "a same-repository branch SHA is accepted for preview-1"
  cand_case cand-fork-refused preview-1 "$F1" 2 "" "a merge-request/fork-only commit is refused"
  cand_case cand-missing-branch preview-1 no-such-branch 2 "" "a missing branch is refused"
  cand_case cand-s6 preview-1 infra-change 2 "" "an infrastructure-changing candidate is refused by S6 (incompatible candidate)"
  # cand-branch-pinned: the branch resolves once in prepare; deleted before the later jobs, which run in a fresh clone
  # without that commit: the bundle carries it and the deploy installs exactly the pinned SHA.
  begin
  set_env preview-1; world; JCAND=preview-feature
  PIPE=$((PIPE + 1)); job prepare
  chk '[[ $RC == 0 ]] && grep -qx "branch	preview-feature" "$JOUT/candidate.tsv" && grep -qx "candidate	$P1" "$JOUT/candidate.tsv"' "branch not resolved and recorded"
  git -C "$O" branch -q -D preview-feature
  git clone -q --no-local --branch staging "$O" "$W/job2" 2>/dev/null
  git -C "$W/job2" checkout -q --detach "$CTL"
  chk '! git -C "$W/job2" cat-file -e "$P1^{commit}" 2>/dev/null' "the fresh clone already has the commit"
  JOBDIR=$W/job2 FAKE_REPO=$W/job2
  job check; job build; job deploy
  chk '[[ $RC == 0 ]]' "deploy from the bundle: exit $RC"
  chk '[[ "$(cat "$FAKE_ROOT/vm/candidate")" == "$P1" ]]' "the VM runs $(cat "$FAKE_ROOT/vm/candidate"), not the pinned $P1"
  JOBDIR="" FAKE_REPO=$JOB
  git -C "$O" branch -q preview-feature "$P1"
  done_case cand-branch-pinned "a branch resolved once is pinned; after its deletion the later jobs deploy the bundled SHA"
}

# ==================================================================================== reuse (Plan §10.8)
reuse_cases() {
  set_env staging
  # reuse-no-previous-artifacts: a fresh pipeline, a valid published manifest, no earlier artifacts: check yes, build
  # skipped, deploy reuses with zero load/push and reports the manifest's digests.
  begin; world; publish_manifest "$CTL"; rm -rf "$W/ns"
  JCAND=""; pipeline
  chk '[[ $RC == 0 ]]' "pipeline exit $RC"
  chk 'grep -qx "STAGING_REUSE=yes" "$W/ns/p$PIPE/staging-check/reuse.env"' "check did not decide yes"
  chk '[[ -f $W/ns/p$PIPE/staging-build/skipped && ! -e $W/ns/p$PIPE/staging-build/images.tar.gz ]]' "build not skipped"
  chk '! grep -qE "^docker	(L|W)	" "$FAKE_ROOT/calls.log"' "docker load/tag/push/build ran"
  chk 'grep -q "^digests: core=sha256:" "$JOUT/summary.txt"' "summary lacks the digests"
  done_case reuse-no-previous-artifacts "fresh pipeline, no earlier artifacts: check yes, build skipped, deploy reused with zero load/push"
  # reuse-check-no: invalid manifests make check decide no
  local mode
  for mode in repository candidate shape missing-digest; do
    begin; world; fx_manifest "$W/m.tsv" "$CTL"
    case $mode in
      repository) sed -i "s#$FX_AR/#us-central1-docker.pkg.dev/$FX_PROJECT_ID/other/#" "$W/m.tsv" ;;
      candidate) sed -i "s/^candidate\t.*/candidate\t$Y/" "$W/m.tsv" ;;
      shape) sed -i 's/@sha256:\([0-9a-f]\{10\}\)[0-9a-f]*$/@sha256:\1/' "$W/m.tsv" ;;
      missing-digest) sed -i "\#/core@#d" "$FAKE_ROOT/ar/digests" ;;
    esac
    fx_put "deploy/$CTL.tsv" "$W/m.tsv"
    PIPE=$((PIPE + 1)); job prepare; job check
    chk '[[ $RC == 0 ]] && grep -qx "STAGING_REUSE=no" "$JOUT/reuse.env"' "check did not decide no"
    chk 'zero_writes' "check wrote"
    done_case "reuse-check-no-$mode" "an invalid published manifest ($mode) makes check decide no; check writes nothing"
  done
  # reuse-check-yes-deploy-invalid: the manifest disappears between check and deploy (no build artifacts): exit 2, zero writes.
  begin; world; publish_manifest "$CTL"
  PIPE=$((PIPE + 1)); job prepare; job check; job build
  printf 'candidate\t%s\n' "$Y" >"$W/bad.tsv"; fx_put "deploy/$CTL.tsv" "$W/bad.tsv"
  job deploy
  chk '[[ $RC == 2 ]] && zero_writes' "deploy exit $RC, writes $(since_writes | nlines)"
  done_case reuse-check-yes-deploy-invalid "check yes but invalid at deploy without build artifacts: exit 2, zero writes"
  # reuse-check-no-deploy-valid: check no, build ran, a valid manifest appears before deploy: deploy reuses (no push).
  begin; world
  PIPE=$((PIPE + 1)); job prepare; job check; job build
  chk 'grep -qx "STAGING_REUSE=no" "$W/ns/p$PIPE/staging-check/reuse.env" && [[ -f $W/ns/p$PIPE/staging-build/images.tar.gz ]]' "check/build did not take the build path"
  publish_manifest "$CTL"
  job deploy
  chk '[[ $RC == 0 ]]' "deploy exit $RC"
  chk '[[ -z "$(since | grep -E "^docker	(L|W)	")" ]]' "deploy loaded or pushed although a valid manifest appeared"
  done_case reuse-check-no-deploy-valid "check no but valid at deploy: the deploy reuses; the build is discarded"
  # reuse-build-path: no manifest: build, load, verify, retag, push, publish; the published manifest is valid.
  begin; world
  JCAND=""; pipeline
  chk '[[ $RC == 0 ]]' "pipeline exit $RC"
  chk '[[ "$(since | grep -c "^docker	W	.*push ")" == 7 ]]' "not seven pushes"
  chk 'fx_cat "deploy/$CTL.tsv" | grep -c "@sha256:" | grep -qx 7' "published manifest lacks digests"
  chk 'grep -E "^docker	L	compose " "$FAKE_ROOT/calls.log" | grep -q "/src/infra/compose/local.yml"' "the build did not use the candidate worktree"
  chk '! grep -E "^docker	L	(compose|build) " "$FAKE_ROOT/calls.log" | grep -qF "$JOB/infra/"' "the build used the control checkout"
  chk 'grep -qE "^docker	R	srcmode [0-7][0-7][4-7]$" "$FAKE_ROOT/calls.log"' "the candidate worktree is not world-readable (the image user could not read /srv/app/src)"
  done_case reuse-build-path "no manifest: built from the candidate worktree, IDs verified, retagged, pushed, manifest published"
  # reuse-image-mismatch: a build manifest whose image ID differs from the loaded image: refused before push (exit 2).
  begin; world
  PIPE=$((PIPE + 1)); job prepare; job check; job build
  sed -i -E '0,/^image\tcore\t/s/(^image\tcore\t[^\t]*\t[^\t]*\t)sha256:[0-9a-f]{64}/\1sha256:'"$(printf '%064d' 0)"'/' "$W/ns/p$PIPE/staging-build/build-manifest.tsv"
  (cd "$W/ns/p$PIPE/staging-build" && sha256sum images.tar.gz build-manifest.tsv >SHA256SUMS)
  job deploy
  chk '[[ $RC == 2 ]]' "deploy exit $RC, not 2"
  chk '[[ -z "$(since | grep "^docker	W	")" ]]' "pushed despite the mismatch"
  done_case reuse-image-mismatch "a loaded image whose ID differs from the build manifest is refused before any push"
}

# ======================================================================================= life (AC4/AC5)
life_cases() {
  local g i ok
  set_env staging
  # life-stopped-default: TERMINATED 7200 s, empty duration: no token read, start without --duration, deploy, 2 h session.
  begin; world TERMINATED; publish_manifest "$CTL"; : >"$FAKE_ROOT/tf-cf.log"
  JCAND=""; pipeline
  chk '[[ $RC == 0 ]]' "exit $RC"
  chk '! grep -q "secrets versions access" "$FAKE_ROOT/calls.log"' "the Cloudflare token was read without a length change"
  chk 'grep -q "^gcloud	W	compute instances start" "$FAKE_ROOT/calls.log"' "no start"
  chk '! grep -q "set-scheduling" "$FAKE_ROOT/calls.log"' "set-scheduling without a length change"
  chk 'grep -qx "session: 7200s" "$JOUT/summary.txt" && grep -q "^scheduled_stop: 20" "$JOUT/summary.txt"' "summary lacks the 2 h session or scheduled stop"
  done_case life-stopped-default "stopped VM, empty duration: started with the 7200 s default, no token, deployed; scheduled stop reported"
  # life-stopped-duration: 90m: the token is read once (federated start principal) and reaches only the start's Terraform.
  begin; world TERMINATED; publish_manifest "$CTL"; : >"$FAKE_ROOT/tf-cf.log"
  JCAND=""; JDUR=90m; pipeline; JDUR=""
  chk '[[ $RC == 0 ]]' "exit $RC"
  chk '[[ "$(grep -c "secrets versions access" "$FAKE_ROOT/calls.log")" == 1 ]]' "token reads: $(grep -c "secrets versions access" "$FAKE_ROOT/calls.log")"
  chk 'grep -q -- "--max-run-duration=5400s" "$FAKE_ROOT/calls.log"' "no set-scheduling to 5400 s"
  chk 'grep -qx set "$FAKE_ROOT/tf-cf.log"' "the start's Terraform did not receive the token"
  chk '[[ "$(tail -n 1 "$FAKE_ROOT/tf-cf.log")" == unset ]]' "a later Terraform call (receipt) still had the token"
  chk 'grep -qx "session: 5400s" "$JOUT/summary.txt"' "summary session is not 5400s"
  done_case life-stopped-duration "stopped VM, 90m: one token read, set-scheduling 5400 s, token only in the start's process"
  # life-stopped-recovery: 90m, the VM already reports 5400 s while Terraform records 7200 s (an earlier set-scheduling whose
  # refresh did not complete): the job reads no token beforehand, start refuses before any write, and the job retries it
  # once with the token read then; no second set-scheduling.
  begin; world TERMINATED; publish_manifest "$CTL"; : >"$FAKE_ROOT/tf-cf.log"
  printf '5400\n' >"$FAKE_ROOT/vm/max_run"
  JCAND=""; JDUR=90m; pipeline; JDUR=""
  chk '[[ $RC == 0 ]]' "exit $RC"
  chk '[[ "$(grep -c "secrets versions access" "$FAKE_ROOT/calls.log")" == 1 ]]' "token reads: $(grep -c "secrets versions access" "$FAKE_ROOT/calls.log")"
  chk '[[ "$(grep -c "^gcloud	W	compute instances start" "$FAKE_ROOT/calls.log")" == 1 ]]' "the VM was not started exactly once"
  chk '! grep -q "set-scheduling" "$FAKE_ROOT/calls.log"' "set-scheduling although the VM already reported 5400 s"
  chk 'grep -qx set "$FAKE_ROOT/tf-cf.log" && [[ "$(tail -n 1 "$FAKE_ROOT/tf-cf.log")" == unset ]]' "the token did not reach the start's Terraform only"
  chk 'grep -qx "session: 5400s" "$JOUT/summary.txt"' "summary session is not 5400s"
  done_case life-stopped-recovery "stopped VM already at 5400 s, Terraform at 7200 s: start's own token request retried once with the token, one start, no set-scheduling"
  # life-running-keep: RUNNING, empty duration: no start, no token, no scheduling change; deploy only.
  begin; world; publish_manifest "$CTL"
  JCAND=""; pipeline
  chk '[[ $RC == 0 ]]' "exit $RC"
  chk '! grep -qE "compute instances (start|stop|set-scheduling)|secrets versions access" "$FAKE_ROOT/calls.log"' "running VM: start/stop/scheduling/token call"
  done_case life-running-keep "running VM, empty duration: no start, no token, no timer change; the existing scheduled stop is reported"
  # life-stop-no-artifacts: a fresh stop pipeline with no candidate, build, artifact or earlier summary.
  begin; world; publish_manifest "$CTL"; JCAND=""; pipeline; rm -rf "$W/ns"
  g="$(fx_gen current.tsv)"
  stop_pipeline
  chk '[[ $RC == 0 ]]' "stop exit $RC"
  chk '[[ "$(cat "$FAKE_ROOT/vm/status")" == TERMINATED ]]' "VM not stopped"
  chk '[[ "$(fx_gen current.tsv)" -gt "$g" ]] && fx_cat current.tsv | grep -qx "state	TERMINATED"' "current.tsv not overwritten"
  chk '[[ "$(fx_gen lease)" == 0 ]]' "lease left"
  chk 'has_key "^stop-requests/$PIPE$"' "no stop request"
  chk 'has_key "^receipts/"' "no receipt published"
  done_case life-stop-no-artifacts "stop needs no candidate, build, artifact or earlier summary; current.tsv overwritten, lease released"
  # life-stop-after-failed-start: the start fails (VM-side checks): deploy exit 4; stop then succeeds.
  begin; world TERMINATED; publish_manifest "$CTL"; printf 'fail\n' >"$FAKE_ROOT/vm/start_mode"
  JCAND=""; pipeline
  chk '[[ $RC == 4 ]]' "failed start: exit $RC, not 4"
  rm -f "$FAKE_ROOT/vm/start_mode"
  stop_pipeline
  chk '[[ $RC == 0 && "$(fx_gen lease)" == 0 ]]' "stop after the failed start: exit $RC"
  done_case life-stop-after-failed-start "a failed start leaves stop available: stop exit 0, lease released"
  # life-stop-after-canceled-start: the deploy job is killed while its start is on the VM side; stop then succeeds.
  begin; world TERMINATED; publish_manifest "$CTL"; : >"$FAKE_ROOT/vm/block-start"
  PIPE=$((PIPE + 1)); job prepare; job check; job build
  cancel_job deploy start-blocked block-start
  stop_pipeline
  chk '[[ $RC == 0 && "$(fx_gen lease)" == 0 ]]' "stop after the canceled start: exit $RC, lease gen $(fx_gen lease)"
  done_case life-stop-after-canceled-start "a canceled start leaves stop available: stop exit 0, lease released"
  # life-cycles: start/stop three times as deployer and ci-stop under the IAM model: every step permitted, lease absent
  # after each job, current.tsv's generation advancing at each stop.
  begin; world TERMINATED; publish_manifest "$CTL"
  ok=1
  for i in 1 2 3; do
    JCAND=""; pipeline
    [[ $RC == 0 && "$(fx_gen lease)" == 0 ]] || { ok=0; why+="cycle $i deploy exit $RC; "; }
    g="$(fx_gen current.tsv)"
    stop_pipeline
    [[ $RC == 0 && "$(fx_gen lease)" == 0 ]] || { ok=0; why+="cycle $i stop exit $RC; "; }
    [[ "$(fx_gen current.tsv)" -gt "$g" ]] || { ok=0; why+="cycle $i: current.tsv generation did not advance; "; }
  done
  stop_pipeline
  [[ $RC == 0 ]] || why+="a second consecutive stop exit $RC; "
  done_case life-cycles "three start/stop cycles and a repeated stop: every step permitted, lease absent after each job, current.tsv advanced at each stop"
  # life-supersession: a later stop pipeline refuses an earlier queued start (zero writes); a later start proceeds.
  begin; world; publish_manifest "$CTL"
  PIPE=$((PIPE + 1)); local early=$PIPE; job prepare; job check; job build
  PIPE=$((PIPE + 10)); stop_pipeline
  chk '[[ $RC == 0 ]]' "stop exit $RC"
  PIPE=$early; job deploy
  chk '[[ $RC == 2 ]] && zero_writes' "earlier start after the later stop: exit $RC"
  chk 'grep -q superseded "$ERR" "$W/runs/p$PIPE-deploy.out"' "refusal does not say superseded"
  PIPE=$((early + 20)); pipeline
  chk '[[ $RC == 0 ]]' "a later start: exit $RC"
  done_case life-supersession "an earlier queued start is refused after a later stop (exit 2, zero writes); a later start proceeds"
  # life-expired: rc 255 with the instance TERMINATED after its armed deadline is exit 5, never success.
  begin; world; publish_manifest "$CTL"; printf '%s\n' "$((FX_NOW + 400))" >"$FAKE_ROOT/vm/term"; printf 'expire\n' >"$FAKE_ROOT/vm/deploy_mode"
  JCAND=""; pipeline
  chk '[[ $RC == 5 ]]' "exit $RC, not 5"
  chk 'grep -qx "exit_class: 5" "$JOUT/summary.txt"' "summary exit class is not 5"
  done_case life-expired "the session expired during the deploy (rc 255, TERMINATED after the deadline): exit 5"
  # life-stop-idempotent: an already stopped VM: exit 0 after the describe confirms it; no compute stop call.
  begin; world TERMINATED; publish_manifest "$CTL"; JCAND=""; pipeline; stop_pipeline
  stop_pipeline
  chk '[[ $RC == 0 ]]' "exit $RC"
  chk '[[ -z "$(since | grep "compute instances stop")" ]]' "a stopped VM was stopped again"
  done_case life-stop-idempotent "a stopped VM: stop is an idempotent success, no compute stop issued"
  # life-stop-local-lease: a live local lease: stop writes only its stop request, then refuses (exit 2).
  begin; world; fx_lease "local:op@host:1" none; g="$(fx_gen lease)"
  stop_pipeline
  chk '[[ $RC == 2 ]]' "exit $RC, not 2"
  chk '[[ "$(since_writes | nlines)" == 1 ]] && since_writes | grep -q "stop-requests/$PIPE "' "writes: $(since_writes | cut -f3 | cut -c1-60 | tr '\n' ' ')"
  chk '[[ "$(fx_gen lease)" == "$g" ]]' "the local lease changed"
  done_case life-stop-local-lease "under a live local lease stop writes only its stop request and refuses (exit 2); the lease is untouched"
  # life-ci-break-expired: an expired CI lease is broken by the next CI job with a lease-breaks record; an expired
  # local lease refuses CI.
  begin; world; fx_lease "ci:9/99" "$((FX_NOW - 5))"
  stop_pipeline
  chk '[[ $RC == 0 ]]' "stop over an expired CI lease: exit $RC"
  chk 'has_key "^lease-breaks/.*ci_9_99"' "no lease-breaks record"
  world; fx_lease "local:op@host:1" "$((FX_NOW - 5))"
  stop_pipeline
  chk '[[ $RC == 2 ]]' "stop over an expired local lease: exit $RC"
  done_case life-ci-break-expired "an expired CI lease is broken with a record; an expired local lease still refuses CI"
}

# cancel_job ACTION BLOCKED_FLAG BLOCK_FILE: the job in its own process group, TERM to the group once the fake VM side
# holds it.
# As GitLab's cancellation reaches the context (docker stop: TERM to the adapter process only, through --init), TERM goes
# to the ci.sh process alone; the fake VM side is released shortly after so the interrupted step can return.
cancel_job() {
  local i pid cpid=""
  declare -p PIPE JENV JCAND JDUR PROFILE JOB W FAKE_ROOT FAKE_REPO FAKES FX_JWT FX_ENV CTL REPORT JOBSEQ CASE_HINT >"$W/job-env.sh"
  declare -f job fx_run since >>"$W/job-env.sh"
  ( setsid bash -c 'echo $$ >"$1"; shift; source "$1"; shift; job "$@"' _ "$W/cancel.pgid" "$W/job-env.sh" "$1" ) &
  pid=$!
  for ((i = 0; i < 300; i++)); do [[ -f $FAKE_ROOT/vm/$2 ]] && break; sleep 0.1; done
  [[ -f $W/cancel.pgid ]] && cpid="$(pgrep -g "$(cat "$W/cancel.pgid")" -f "infra/staging/ci.sh $1\$" | head -n 1)"
  [[ -z $cpid ]] || kill -TERM "$cpid" 2>/dev/null
  sleep 0.5
  rm -f "$FAKE_ROOT/vm/$3"
  wait "$pid" 2>/dev/null || true
  rm -f "$W/cancel.pgid" "$FAKE_ROOT/vm/$2"
  [[ -n $cpid ]] || why+="cancel: no ci.sh process found; "
}

# ======================================================================================== cred (AC3)
cred_cases() {
  local f
  set_env staging
  begin; world TERMINATED; publish_manifest "$CTL"
  JCAND=""; JDUR=90m; pipeline; JDUR=""
  chk '[[ $RC == 0 ]]' "deploy exit $RC"
  f=$FAKE_ROOT/auth-cred-copy.json
  chk 'grep -q "\"type\": \"external_account\"" "$f" && grep -q "\"audience\": \"//iam.googleapis.com/$PROVIDER\"" "$f"' "gcloud credential file type/audience"
  chk 'grep -q "\"credential_source\": {\"file\": \"$JPRIV/id-token\"}" "$f"' "gcloud credential file does not read the token file in place"
  chk 'grep -q "\"credential_source\": {\"file\": \"/private/ci/id-token\"}" "$FAKE_ROOT/tf-adc-copy.json"' "Terraform credential file does not read /private/ci/id-token"
  chk '! grep -qE "service_account_impersonation_url|private_key|$FX_JWT" "$f" "$FAKE_ROOT/tf-adc-copy.json"' "a credential file holds an impersonation URL, a key or the token"
  chk '[[ ! -e $JPRIV/raw && ! -e $JPRIV/private/ci && ! -e $JPRIV/gcloud-adc.json ]]' "raw/, private/ci or gcloud-adc.json left after the job"
  chk '! grep -rqF "$FX_JWT" "$JOUT" "$FAKE_ROOT/gcs"' "the ID token reached an artifact or upload"
  chk '! grep -rqF "$FX_CF_TOKEN" "$JOUT" "$FAKE_ROOT/gcs" "$FAKE_ROOT/calls.log"' "the Cloudflare token reached an artifact, upload or argument"
  chk '! grep -rq G072PLANCANARY "$JOUT" "$FAKE_ROOT/gcs" "$W/runs/p$PIPE-deploy.out" "$W/runs/p$PIPE-deploy.err"' "the Terraform plan text reached an output"
  done_case cred-shapes "external_account files (token read in place; Terraform via /private/ci/id-token), no key or impersonation URL; credentials and raw/ removed; no token or plan text in outputs"
  # cred-impersonate-sa: lib.sh GCLOUD_IMPERSONATE_SA validation (Plan §10.9.2).
  begin
  impersonate_case "" 0 "staging-deployer@$FX_PROJECT_ID.iam.gserviceaccount.com"
  impersonate_case "staging-ci-stop@$FX_PROJECT_ID.iam.gserviceaccount.com" 0 "staging-ci-stop@$FX_PROJECT_ID.iam.gserviceaccount.com"
  impersonate_case "staging-ci-read@$FX_PROJECT_ID.iam.gserviceaccount.com" 0 "staging-ci-read@$FX_PROJECT_ID.iam.gserviceaccount.com"
  impersonate_case "preview-1-ci-stop@$FX_PROJECT_ID.iam.gserviceaccount.com" 2
  impersonate_case "staging-ci-stop@other-project.iam.gserviceaccount.com" 2
  impersonate_case "staging-ci-stop@$FX_PROJECT_ID.iam.gserviceaccount.com.evil" 2
  impersonate_case "staging-owner@$FX_PROJECT_ID.iam.gserviceaccount.com" 2
  done_case cred-impersonate-sa "absent = DEPLOYER_SA; exactly <env>-deployer/-ci-stop/-ci-read @<project> accepted; another environment, project, suffix or role refused"
  # cred-canary: a canary inside an allowlisted VM-side line: exit 6, no sanitized log published, summary withheld.
  begin; world; publish_manifest "$CTL"; printf 'canary\n' >"$FAKE_ROOT/vm/deploy_mode"
  JCAND=""; pipeline
  chk '[[ $RC == 6 ]]' "exit $RC, not 6"
  chk '! has_key "^logs/"' "a log was published"
  chk '! grep -rq G072PLANCANARY "$JOUT" "$FAKE_ROOT/gcs" "$W/runs/p$PIPE-deploy.out" "$W/runs/p$PIPE-deploy.err"' "the canary reached an output"
  chk '[[ "$(fx_gen lease)" == 0 ]]' "lease left"
  done_case cred-canary "a canary in an allowlisted line: exit 6, nothing published, lease released"
}

impersonate_case() { # VALUE EXPECT_RC [EXPECT_SA]
  local c=$W/imp.conf
  fx_conf "$c" "$W/imp-ev" "$W/imp-pr" ${1:+"GCLOUD_IMPERSONATE_SA=$1"}
  fx_run imp -- bash -c 'source "$1"; load_conf "$2"; printf "%s\n" "${GFLAGS[@]}"' _ "$JOB/infra/staging/lib.sh" "$c"
  [[ $RC == "$2" ]] || why+="[$1] exit $RC, not $2; "
  [[ $2 != 0 ]] || grep -qx -- "$3" "$OUT" || why+="[$1] GFLAGS impersonate $(sed -n 4p "$OUT"), not $3; "
}

# ================================================================================ control (mutations)
# with_mutation NAME SED_FILE SED_EXPR: the fixture job clone runs a mutated copy of one control script.
with_mutation() { cp "$JOB/infra/staging/$1" "$W/orig-$1"; sed -i "$2" "$JOB/infra/staging/$1"; }
without_mutation() { cp "$W/orig-$1" "$JOB/infra/staging/$1"; }

control_cases() {
  set_env staging
  # control-fence-gate: s7 always passing: the older candidate after a partial deploy is no longer refused.
  begin; world; printf 'fail-readiness\n' >"$FAKE_ROOT/vm/deploy_mode"; JCAND=$X; pipeline; rm -f "$FAKE_ROOT/vm/deploy_mode"
  with_mutation ci-lib.sh 's/^      if (cm < mm) exit 1$/      exit 0/'
  JCAND=""; pipeline
  without_mutation ci-lib.sh
  [[ $RC == 2 ]] && zero_writes || why="older candidate exit $RC with $(since_writes | nlines) writes"
  control control-fence-gate "S7 gate always passes (fence-older-refused assertion)" "$why"
  # control-fence-write: staging.sh without its fence call: the partial deploy leaves no fence, W is accepted.
  begin; world; printf 'fail-readiness\n' >"$FAKE_ROOT/vm/deploy_mode"
  with_mutation staging.sh '/|| fence_append "\$c"$/d'
  JCAND=$X; pipeline; rm -f "$FAKE_ROOT/vm/deploy_mode"
  fence_has "$X" || why+="no fence entry for X; "
  JCAND=""; pipeline
  without_mutation staging.sh
  [[ $RC == 2 ]] && zero_writes || why+="older candidate exit $RC"
  control control-fence-write "deploy_candidate without fence_append (fence-partial/older-refused assertions)" "$why"
  # control-order: the lease taken before the phase-B gates: a superseded deploy now writes.
  begin; world
  with_mutation ci.sh '0,/^  deploy_gates$/{/^  deploy_gates$/d}'
  PIPE=$((PIPE + 1)); job prepare; job check; job build
  printf 'x\n' >"$W/sr"; fx_put "stop-requests/$((PIPE + 50))" "$W/sr"
  job deploy
  without_mutation ci.sh
  [[ $RC == 2 ]] && zero_writes || why="superseded deploy exit $RC with writes $(since_writes | cut -f3 | cut -c1-40 | tr '\n' ' ')"
  control control-order "lease taken before the gates (order-superseded zero-write assertion)" "$why"
  # control-phase-a: the profile identity check removed: a profile for another environment now reaches phase B (cloud calls).
  begin; world
  PIPE=$((PIPE + 1)); job prepare; job check; job build
  with_mutation ci-lib.sh '/die 2 "the profile is for /d'
  sed -i "s/^STAGING_ENV=staging$/STAGING_ENV=preview-1/; s/staging-deployer@/preview-1-deployer@/; s|/staging$|/preview-1|" "$PROFILE/staging.conf"
  job deploy
  without_mutation ci-lib.sh; set_env staging
  [[ $RC == 2 && -z "$(since_cloud)" ]] || why="profile mismatch exit $RC with cloud calls $(since_cloud | nlines)"
  control control-phase-a "profile identity check removed (order-profile-mismatch phase-A zero-cloud-call assertion)" "$why"
  # control-canary: scan_all disabled: the canary case publishes.
  begin; world; publish_manifest "$CTL"; printf 'canary\n' >"$FAKE_ROOT/vm/deploy_mode"
  with_mutation ci-lib.sh 's/^scan_all() { canary_scan .*/scan_all() { true; }/'
  JCAND=""; pipeline
  without_mutation ci-lib.sh
  [[ $RC == 6 ]] && ! has_key '^logs/' || why="canary run exit $RC, logs $(fx_keys | grep -c '^logs/')"
  control control-canary "canary scan disabled (cred-canary assertion)" "$why"
  # control-iam: the exact-key delete grant removed from the model: the first stop fails at the current.tsv overwrite
  # (the r0 defect reproduced).
  begin; world; publish_manifest "$CTL"; JCAND=""; pipeline
  grep -v $'^ci-stop\tciMetaMutable\t' "$FX_PERMS" >"$FAKE_ROOT/perms.tsv"
  stop_pipeline
  if [[ $RC == 0 ]]; then why=""
  elif grep -q "current.tsv could not be written" "$ERR" "$W/runs/p$PIPE-stop.out"; then why="first stop exit $RC at the current.tsv overwrite"
  else why="first stop exit $RC elsewhere: $(tail -n 1 "$ERR")"; fi
  control control-iam "ciMetaMutable (exact-key delete) removed: r0 defect, the first stop fails at the current.tsv overwrite (life-cycles assertion)" "$why"
  # control-impersonate: lib.sh accepting any identity: the foreign-environment value is no longer refused.
  begin
  with_mutation lib.sh 's/^    \*) die 2 "staging.conf: GCLOUD_IMPERSONATE_SA must be/    *) : "/'
  impersonate_case "preview-1-ci-stop@$FX_PROJECT_ID.iam.gserviceaccount.com" 2
  without_mutation lib.sh
  control control-impersonate "GCLOUD_IMPERSONATE_SA validation removed (cred-impersonate-sa assertion)" "$why"
}

# =========================================================== probe (Codex review finding 1: affirmative denial proofs)
# probe_run MODE TOKEN_MODE EXPECT_RC: a check job (its read identity runs the impersonation and state-bucket probes)
# against the given curl answer and token behaviour; why += each failed expectation.
probe_run() {
  world; publish_manifest "$CTL"
  [[ -z $1 ]] || printf '%s\n' "$1" >"$FAKE_ROOT/probe-mode"
  [[ -z $2 ]] || printf '%s\n' "$2" >"$FAKE_ROOT/token-mode"
  JCAND=""; PIPE=$((PIPE + 1)); job prepare; job check
  EXP_RC=$3
  chk '[[ $RC == "$EXP_RC" ]]' "check exit $RC, not $3"
  chk 'zero_writes' "the check wrote"
  if [[ $3 == 6 ]]; then chk '[[ ! -e $JOUT/reuse.env ]]' "a failed probe still produced a reuse decision"; fi
  if [[ $3 == 0 ]]; then chk 'grep -q "testPermissions" "$FAKE_ROOT/calls.log"' "no testIamPermissions probe ran"; fi
}

probe_cases() {
  local m
  set_env staging
  for m in deny403:0 deny404:0 empty200:0 emptyarray200:0 granted200:6 transport:6 5xx:6 401:6 garbage200:6 truncated200:6 \
    wrongkind200:6 permstring200:6 permnonstring200:6 emptybody200:6 twovalues200:6; do
    begin; probe_run "${m%%:*}" "" "${m#*:}"
    done_case "probe-${m%%:*}" "testIamPermissions answer ${m%%:*}: check exit ${m#*:}$([[ ${m#*:} == 6 ]] && echo ' (no reuse decision, zero writes)')"
  done
  begin; probe_run "" empty 6; done_case probe-token-empty "an empty access token proves nothing: exit 6, zero writes"
  begin; probe_run "" fail 6; done_case probe-token-fail "a failed token command proves nothing: exit 6, zero writes"
  # control-probe-failopen: the old behaviour (a curl failure counts as a denial) restored: the transport case passes.
  begin
  with_mutation ci-lib.sh 's/^  ((rc == 0)) || { PROBE_WHY="transport error (curl exit \$rc)"; return 2; }$/  ((rc == 0)) || return 0/'
  probe_run transport "" 6
  without_mutation ci-lib.sh
  control control-probe-failopen "a transport failure counted as a denial (probe-transport assertion)" "$why"
  # control-probe-grep: the grep-only body check restored (kind found, no permissions string = denied): the truncated
  # body passes as an empty permission set.
  begin
  cp "$JOB/infra/staging/ci-lib.sh" "$W/orig-ci-lib.sh"
  cat >>"$JOB/infra/staging/ci-lib.sh" <<'EOF'
probe_verdict() { grep -qE '"kind"[[:space:]]*:[[:space:]]*"storage#testIamPermissionsResponse"' "$1" || { printf invalid; return; }; if grep -qE '"permissions"[[:space:]]*:[[:space:]]*\[[[:space:]]*"' "$1"; then printf granted; else printf denied; fi; }
EOF
  probe_run truncated200 "" 6
  without_mutation ci-lib.sh
  control control-probe-grep "the grep-only 200-body check restored (probe-truncated200 assertion)" "$why"
}

# ======================================================= stopfb (Codex review finding 2: provider-level stop fallback)
# stopfb_run MODES EXPECT_RC: a running, deployed environment; the failure files MODES (FAKE_ROOT-relative) are set,
# then a fresh stop pipeline.
stopfb_run() {
  local f
  world; publish_manifest "$CTL"; JCAND=""; pipeline
  for f in $1; do : >"$FAKE_ROOT/$f"; done
  [[ " $1 " != *" vm/ssh_mode "* ]] || printf 'fail\n' >"$FAKE_ROOT/vm/ssh_mode"
  stop_pipeline
  EXP_RC=$2
  chk '[[ $RC == "$EXP_RC" ]]' "stop exit $RC, not $2"
  chk 'grep -q "^gcloud	W	compute instances stop" "$FAKE_ROOT/calls.log"' "the provider stop was not attempted"
  chk '[[ "$(fx_gen lease)" == 0 ]]' "lease left"
  if [[ $2 == 0 ]]; then
    chk '[[ "$(cat "$FAKE_ROOT/vm/status")" == TERMINATED ]] && fx_cat current.tsv | grep -qx "state	TERMINATED"' "not observed and recorded TERMINATED"
  else
    chk '! fx_cat current.tsv | grep -qx "state	TERMINATED"' "current.tsv claims TERMINATED after a failed provider stop"
    chk 'grep -qx "exit_class: $EXP_RC" "$JOUT/summary.txt"' "summary does not carry exit class $2"
  fi
}

stopfb_cases() {
  set_env staging
  begin; stopfb_run "keygen-fail vm/ssh_mode" 0
  done_case stopfb-keygen "key generation fails: guest stop skipped, provider stop issued, VM observed TERMINATED, exit 0"
  begin; stopfb_run "oslogin-fail vm/ssh_mode" 0
  done_case stopfb-oslogin "key registration fails: guest stop skipped, provider stop issued, VM observed TERMINATED, exit 0"
  begin; stopfb_run "vm/ssh_mode" 0
  done_case stopfb-guest "the guest-side status/stop fail: provider stop issued, VM observed TERMINATED, exit 0"
  begin; stopfb_run "vm/stop_fail" 4
  done_case stopfb-provider-fail "the provider stop fails: exit 4, never success; current.tsv not marked stopped"
  # control-stop-strict: the key required again for stop: a key-generation failure blocks the provider stop.
  begin
  with_mutation ci.sh 's/^  if \[\[ \$VMSTATE == RUNNING \]\] \&\& ! oslogin_key; then say .*/  [[ $VMSTATE != RUNNING ]] || oslogin_key || die 4 "strict key"/'
  stopfb_run "keygen-fail vm/ssh_mode" 0
  without_mutation ci.sh
  control control-stop-strict "the OS Login key strict for stop (stopfb-keygen assertion)" "$why"
}

# ========================================================== owner (Codex review finding 3: metadata bucket ownership)
# owner_run OWNER EXPECT_RC REASON: the bucket's owner project number (gcs/<B>/.owner; "" = this project, "absent" = no
# bucket) against the profile's WIF provider number, through a stop job (presence is its first cloud read).
owner_run() {
  world
  case $1 in
    "") ;;
    absent) rm -rf "${FAKE_ROOT:?}/gcs/$FX_BUCKET" ;;
    *) printf '%s\n' "$1" >"$FAKE_ROOT/gcs/$FX_BUCKET/.owner" ;;
  esac
  stop_pipeline
  EXP_RC=$2 REASON=$3
  chk '[[ $RC == "$EXP_RC" ]]' "stop exit $RC, not $2"
  [[ -z $3 ]] || chk 'grep -qF -- "$REASON" "$ERR"' "the refusal does not say [$3]"
  [[ $2 == 0 ]] || chk 'zero_writes' "writes: $(since_writes | cut -f3 | cut -c1-50 | tr '\n' ' ')"
  chk '! since | grep -q "projects describe"' "the CI job read resourcemanager instead of the WIF provider number"
}

owner_cases() {
  set_env staging
  begin; owner_run "" 0 ""; done_case owner-match "owner number = the WIF provider's: the stop proceeds (no resourcemanager read)"
  begin; owner_run 999999999999 2 "belongs to project number 999999999999"; done_case owner-foreign "a foreign owner is refused exit 2 with zero writes"
  begin; owner_run none 2 "no readable owner project number"; done_case owner-missing "no owner number in the describe: refused exit 2, zero writes"
  begin; owner_run 12ab 2 "no readable owner project number"; done_case owner-malformed "a malformed owner number: refused exit 2, zero writes"
  begin; owner_run absent 2 "does not exist"
  chk '[[ "$(since | grep -c "storage buckets describe")" == 1 ]]' "absent bucket: not exactly one describe"
  done_case owner-absent "a confirmed 404 stays absent (the CI job then refuses: not CI-enabled), zero writes"
  begin; world
  sed -i 's#^CI_WIF_PROVIDER=projects/[0-9]*/#CI_WIF_PROVIDER=projects/12/#' "$PROFILE/ci-settings"
  stop_pipeline
  set_env staging
  chk '[[ $RC == 2 ]] && [[ -z "$(since_cloud)" ]] && grep -qF "no valid project number" "$ERR"' "short WIF number: exit $RC, cloud calls $(since_cloud | nlines)"
  done_case owner-wif-number "a WIF provider without a valid project number is refused in phase A (zero cloud calls)"
  # control-owner: the owner comparison removed: the foreign bucket is accepted.
  begin
  with_mutation meta-lib.sh 's/\[\[ \$num == "\$META_PNUM" \]\] ||/true ||/'
  owner_run 999999999999 2 "belongs to project number"
  without_mutation meta-lib.sh
  control control-owner "the owner number comparison removed (owner-foreign assertion)" "$why"
}

# ================================================ class (meta-lib.sh meta_class: an object name never decides the class)
# The intermittent cand-branch-pinned failure: a commit SHA holding "412" (434d3492...2f412c...) in the 404 message for
# deploy/<sha>.tsv was classified as a precondition failure, so the check job died. Fixed SHAs and pipeline IDs in
# every status class; class_run SCRIPT_DIR prints the mismatches.
class_run() {
  fx_run class -- bash -c '
    source "$1/lib.sh"; source "$1/meta-lib.sh"; bad=""
    while IFS="|" read -r want msg; do
      META_ERR=$msg; got="$(meta_class)"; [[ $got == "$want" ]] || bad+="[$msg] -> $got, not $want; "
    done <<EOF
notfound|ERROR: (gcloud.storage.objects.describe) gs://b-ci-meta/deploy/434d34920bcb2f412c6aff7574ffbe02d08af7ba.tsv not found: 404.
notfound|ERROR: (gcloud.storage.objects.describe) gs://b-ci-meta/deploy/a403b404c412d0000000000000000000000000000.tsv not found: 404.
notfound|ERROR: (gcloud.storage.objects.describe) gs://b-ci-meta/stop-requests/403 not found: 404.
notfound|ERROR: (gcloud.storage.ls) One or more URLs matched no objects.
precondition|ERROR: (gcloud.storage.cp) HTTPError 412: At least one of the pre-conditions you specified did not hold. (gs://b-ci-meta/runs/404-4041.tsv)
denied|ERROR: (gcloud.storage.cp) HTTPError 403: ci-stop does not have storage.objects.create access to gs://b-ci-meta/deploy/f412c404.tsv.
error|ERROR: (gcloud.storage.cat) HTTPError 503: backend error (gs://b-ci-meta/deploy/f404c403a412.tsv)
EOF
    printf "%s" "$bad"' _ "$1"
}

class_cases() {
  set_env staging; world
  begin
  class_run "$JOB/infra/staging"
  [[ $RC == 0 && ! -s $OUT ]] || why="$(cat "$OUT") (exit $RC)"
  done_case class-meta "precondition/denied/notfound/error decided by the status alone; SHAs, object names and pipeline IDs in the message never decide it"
  # control-class: the old classifier (a bare 412/403/404 anywhere in the message, URLs included) restored.
  begin
  with_mutation meta-lib.sh 's/^  e="\$(sed -E .*/  e=$META_ERR/; s/\[^0-9A-Za-z\]/[^0-9]/g'
  class_run "$JOB/infra/staging"
  without_mutation meta-lib.sh
  [[ $RC == 0 && ! -s $OUT ]] && why="" || why="$(cut -c1-120 "$OUT")"
  control control-class "the old status classifier restored (class-meta assertion)" "$why"
}

selected_group() { [[ -z $ONLY || ",$ONLY," == *",$1,"* ]] || grep -q ",$1-" <<<",$ONLY,"; }
selected_group class && class_cases
selected_group probe && probe_cases
selected_group stopfb && stopfb_cases
selected_group owner && owner_cases
selected_group fence && fence_cases
selected_group order && order_cases
selected_group cand && cand_cases
selected_group reuse && reuse_cases
selected_group life && life_cases
selected_group cred && cred_cases
selected_group control && control_cases
fx_finish
