# shellcheck shell=bash
# The G072 CI adapter's library (infra/staging/ci.sh; Plan §10.2-§10.8, r0 §2.3, S5-S16). Sourced after lib.sh,
# meta-lib.sh and ci-publish.sh; never executed. Trusted control code from the protected staging pipeline revision: it
# drives control staging.sh against the pinned candidate and never runs candidate code outside the credential-free
# build's Docker builds.
#
# Order of every credential-bearing job (Plan §10.6): phase A local checks only (no cloud call, no token exchange);
# phase B authenticated reads and gates; phase C writes, only after every gate passed. Every refusal before phase C
# exits 2 with zero writes. Every staging.sh subprocess writes to STAGING_CI_PRIVATE/raw/<step>.log (0600) and nothing
# of it reaches this process's stdout/stderr: publication is only ci-publish.sh publish_filter output after
# canary_scan. The job log receives "ci: " lines (fd 3) and the summary.
#
# Exit classes (S14): 0 observed success; 2 refused before any change; 3 HOLD; 4 failed after a change; 5 session
# expired during the action; 6 privacy canary or isolation finding (outputs withheld); 16 is the launcher's.

# Library state is consumed by ci.sh (sourced); shellcheck cannot see across the source.
# shellcheck disable=SC2034
CI_ACTIONS="prepare check build deploy stop"
# Deployment-infrastructure paths that must be byte-identical between control and candidate (S6).
CI_COMPAT_PATHS=(infra/compose infra/postgres infra/minio infra/edge infra/staging infra/docker/edge.Dockerfile infra/docker/minio.Dockerfile)
CI_IDENTITY_NAMES=(IDENTITY_WORKOS_CLIENT_ID IDENTITY_WORKOS_ISSUER IDENTITY_WORKOS_API_KEY)
# The credential-free build tags images under this placeholder registry; deploy retags them (IDs preserved).
CI_PLACEHOLDER_PROJECT=ci-build-placeholder
CI_PLACEHOLDER_AR=us-central1-docker.pkg.dev/ci-build-placeholder/images
# Observed lineage (Plan §10.4, A13): the applied goose versions of the protected store app.goose_db_version
# (core/internal/migrate MetadataTable; columns version_id, is_applied), read as the superuser like remote/lib.sh
# marker_label; tests/ci-adapter.sh pins the exact text.
CI_LINEAGE_SQL='SELECT DISTINCT version_id FROM app.goose_db_version WHERE version_id <> 0 AND is_applied ORDER BY version_id'
CI_SESSION_SECONDS=7200 # CI profiles must keep SESSION_MAX_SECONDS=7200 (Plan §10.2)
CI_MIN_REMAINING=300    # seconds required before a running session's scheduled stop (S13)
declare -A CI_LEASE_TTL=([deploy]=3900 [stop]=2100) # job timeout + 5 min (Plan §10.5)
declare -A CI_KEY_TTL=([deploy]=65m [stop]=35m)

ACTION="" ENVSEL="" PRIV="" OUT="" RAW="" PROF="" CI_CONF="" STAGING_SH="" CTL="" CAND="" SELECTION="" BRANCH=""
HOLDER="" REUSE="" MAN_OK=0 MAN_WHY="" MAN_FILE="" MAN_SHA="" MAN_GEN=0 WRITES=0 CHANGED=0 KEY_FILE="" WORKTREE=""
CANARY_FILE="" CF_TOKEN="" WIF_PROVIDER="" PINS_ARGS=() DENIAL_PROBES="" DEADLINE="" WANT="" NEED_TOKEN=0 BREAK=0
LOGS_DONE=0 RUN_DONE=0 RECEIPT_ID="" RECEIPT_SHA="" RECEIPT_FILE=""
declare -A SUM=() OBS=()

say() { printf 'ci: %s\n' "$*" >&3; }
# die/log from lib.sh, redirected: in a job a refusal reaches the job log as one "ci: " line (fd 3) and the private
# log; a local subcommand prints it once.
die() {
  local code=$1
  shift
  printf 'ci: %s\n' "$*" >&3
  [[ -z $RAW ]] || printf 'staging: %s\n' "$*" >&2
  SUM[detail]="$*"
  exit "$code"
}
log() { printf 'staging: %s\n' "$*" >&2; }

is_sha() { [[ $1 =~ ^[0-9a-f]{40}$ ]]; }

# ------------------------------------------------------------------------------------------------------- phase A: setup
# job_setup ACTION: the launcher contract (Plan §10.10) checked before anything else; private layout under
# STAGING_CI_PRIVATE; this process's own diagnostics (stderr) move to raw/ci.log.
job_setup() {
  local a=$1 v
  ACTION=$a
  SUM[action]=$a
  [[ -z ${STAGING_CI_ACTION:-} || $STAGING_CI_ACTION == "$a" ]] || die 2 "STAGING_CI_ACTION is ${STAGING_CI_ACTION}, not $a"
  [[ -z ${STAGING_LEASE_HELD:-} ]] || die 2 "STAGING_LEASE_HELD is adapter-internal and must not be in the job environment"
  # Never inherited: the Terraform Cloudflare token exists only in the one start that fetched it; the raw ID token
  # only as the launcher's file. Dropped, not trusted (the launcher's allow-list never passes them).
  unset CLOUDFLARE_API_TOKEN STAGING_CI_ID_TOKEN
  PRIV=${STAGING_CI_PRIVATE:-} OUT=${STAGING_CI_OUT:-}
  [[ $PRIV == /* && $PRIV != *[[:space:]]* && -d $PRIV && ! -L $PRIV ]] || die 2 "STAGING_CI_PRIVATE must be an existing absolute directory without spaces"
  [[ "$(stat -c '%a %u' "$PRIV")" == "700 $(id -u)" ]] || die 2 "STAGING_CI_PRIVATE must be mode 0700 and owned by this user"
  [[ $OUT == /* && $OUT != *[[:space:]]* && $OUT != "$PRIV"* && ! -L $OUT ]] || die 2 "STAGING_CI_OUT must be an absolute directory outside STAGING_CI_PRIVATE"
  mkdir -p "$OUT"
  RAW=$PRIV/raw
  install -d -m 0700 "$RAW" "$PRIV/evidence" "$PRIV/private" "$PRIV/home"
  exec 2>>"$RAW/ci.log"
  ENVSEL=${STAGING_CI_ENV:-}
  SUM[environment]=$ENVSEL
  grep -qxF -- "$ENVSEL" "$CI_DIR/ci-environments.txt" 2>/dev/null && [[ -n $ENVSEL ]] ||
    die 2 "environment [$ENVSEL] is not in infra/staging/ci-environments.txt"
  case $a in
    check | deploy | stop) [[ ${CI_ENVIRONMENT_NAME:-} == "$ENVSEL" ]] || die 2 "CI_ENVIRONMENT_NAME [${CI_ENVIRONMENT_NAME:-}] is not the selected environment $ENVSEL" ;;
    *) [[ -z ${CI_ENVIRONMENT_NAME:-} || ${CI_ENVIRONMENT_NAME} == "$ENVSEL" ]] || die 2 "CI_ENVIRONMENT_NAME [${CI_ENVIRONMENT_NAME}] is not $ENVSEL" ;;
  esac
  for v in CI_PIPELINE_ID CI_JOB_ID; do [[ ${!v:-} =~ ^[1-9][0-9]{0,19}$ ]] || die 2 "$v is not a pipeline/job ID"; done
  is_sha "${CI_COMMIT_SHA:-}" || die 2 "CI_COMMIT_SHA is not a full commit ID"
  [[ ${GITLAB_USER_LOGIN:-} =~ ^[A-Za-z0-9._-]{1,255}$ ]] || die 2 "GITLAB_USER_LOGIN is missing or malformed"
  CTL=$CI_COMMIT_SHA HOLDER="ci:$CI_PIPELINE_ID/$CI_JOB_ID"
  SUM[actor]=$GITLAB_USER_LOGIN SUM[pipeline]=$CI_PIPELINE_ID SUM[job]=$CI_JOB_ID
  SELECTION=${STAGING_CI_CANDIDATE_INPUT:-}
  [[ $SELECTION =~ ^([0-9a-f]{40}|[A-Za-z0-9][A-Za-z0-9._/-]{0,199})?$ ]] || die 2 "candidate input is malformed"
  [[ ${STAGING_CI_DURATION_INPUT:-} =~ ^([1-9][0-9]{0,5}[mh]?)?$ ]] || die 2 "duration input is malformed"
  WANT=""
  if [[ -n ${STAGING_CI_DURATION_INPUT:-} ]]; then
    WANT="$(duration_seconds "$STAGING_CI_DURATION_INPUT")" ||
      die 2 "duration takes <n>m, <n>h or whole seconds from $SESSION_FLOOR (10m) to $SESSION_CEILING (8h)"
  fi
  if [[ $a == stop ]]; then
    [[ -z $SELECTION && -z ${STAGING_CI_DURATION_INPUT:-} ]] || die 2 "stop takes no candidate and no duration"
  fi
  # ci-publish.sh canary_scan reads its literals from STAGING_CI_ID_TOKEN_FILE and this file (a fetched token).
  CANARY_FILE=$PRIV/canary-values
  : >"$CANARY_FILE"
  chmod 0600 "$CANARY_FILE"
  export STAGING_CI_CANARY_LITERALS=$CANARY_FILE
}

# Artifact directories (launcher contract): GitLab restores each earlier job's artifacts at its own path beside this
# job's STAGING_CI_OUT: staging-prepare/, staging-check/ (reuse.env), staging-build/, staging-summary/ (deploy, stop).
# Prepare's directory may also be this job's own (artifacts handed on in place).
prep_dir() { if [[ -f $OUT/candidate.tsv ]]; then printf '%s' "$OUT"; else printf '%s/staging-prepare' "$(dirname "$OUT")"; fi; }
build_dir() { if [[ $ACTION == build ]]; then printf '%s' "$OUT"; else printf '%s/staging-build' "$(dirname "$OUT")"; fi; }

require_credential_free() {
  [[ -z ${STAGING_CI_ID_TOKEN_FILE:-}${STAGING_CI_PROFILE_DIR:-} ]] || die 2 "the credential-free $ACTION job received an ID token or profile"
}

# profile_load: the trusted profile (four protected, environment-scoped file variables copied by the launcher) into the
# G070 staging.conf this job runs with, plus the CI-only settings. Everything local; nothing from the pipeline inputs.
profile_load() {
  local k line sa="" f
  PROF=${STAGING_CI_PROFILE_DIR:-}
  [[ $PROF == /* && -d $PROF ]] || die 2 "STAGING_CI_PROFILE_DIR is missing"
  for f in staging.conf ci-settings; do [[ -f $PROF/$f ]] || die 2 "the profile has no $f"; done
  case $ACTION in
    deploy) for f in backend.hcl terraform.tfvars; do [[ -f $PROF/$f ]] || die 2 "the deploy profile has no $f"; done ;;
    check) for f in backend.hcl terraform.tfvars; do [[ ! -e $PROF/$f ]] || die 2 "the check job must not receive $f"; done ;;
  esac
  while IFS= read -r line || [[ -n $line ]]; do
    [[ $line =~ ^[[:space:]]*(#|$) ]] && continue
    k=${line%%=*}
    case $k in
      EVIDENCE_DIR | PRIVATE_DIR | GOOGLE_ADC_FILE | GCLOUD_IMPERSONATE_SA) die 2 "the CI profile must not set $k (the adapter sets it)" ;;
    esac
  done <"$PROF/staging.conf"
  grep -qx 'SESSION_MAX_SECONDS=7200' "$PROF/staging.conf" ||
    die 2 "the CI profile must set SESSION_MAX_SECONDS=7200 (the CI default session is 2 h)"
  case $ACTION in
    stop) sa="GCLOUD_IMPERSONATE_SA=$ENVSEL-ci-stop@" ;;
    check) sa="GCLOUD_IMPERSONATE_SA=$ENVSEL-ci-read@" ;;
  esac
  CI_CONF=$PRIV/private/staging.conf
  install -d -m 0700 "$PRIV/private/ci" "$PRIV/private/plans"
  {
    grep -v '^[[:space:]]*#' "$PROF/staging.conf"
    printf 'EVIDENCE_DIR=%s\nPRIVATE_DIR=%s\nGOOGLE_ADC_FILE=%s\n' "$PRIV/evidence" "$PRIV/private" "$PRIV/private/ci/tf-adc.json"
  } >"$CI_CONF"
  if [[ -n $sa ]]; then
    k="$(sed -n -E 's/^GCP_PROJECT=([a-z][a-z0-9-]{4,28}[a-z0-9])$/\1/p' "$PROF/staging.conf")"
    printf '%s%s.iam.gserviceaccount.com\n' "$sa" "$k" >>"$CI_CONF"
  fi
  chmod 0600 "$CI_CONF"
  load_conf "$CI_CONF"
  [[ ${CONF[STAGING_ENV]} == "$ENVSEL" ]] || die 2 "the profile is for ${CONF[STAGING_ENV]}, not $ENVSEL"
  [[ ${CONF[SESSION_MAX_SECONDS]} == "$CI_SESSION_SECONDS" ]] || die 2 "the CI profile must keep SESSION_MAX_SECONDS=$CI_SESSION_SECONDS"
  if [[ $ACTION == deploy ]]; then
    install -m 0600 "$PROF/backend.hcl" "$PRIV/private/backend.hcl"
    install -m 0600 "$PROF/terraform.tfvars" "$PRIV/private/terraform.tfvars"
  fi
  settings_load "$PROF/ci-settings"
}

# settings_load FILE: CI_WIF_PROVIDER (required), CI_IDENTITY_PINS (required for deploy), CI_DENIAL_PROBES (optional).
settings_load() {
  local line k v n=0 seen=" " name ver numeric=0 none=0
  WIF_PROVIDER="" PINS_ARGS=() DENIAL_PROBES=""
  local pins=""
  while IFS= read -r line || [[ -n $line ]]; do
    n=$((n + 1))
    [[ $line =~ ^[[:space:]]*(#|$) ]] && continue
    [[ $line =~ ^([A-Z][A-Z0-9_]*)=(.*)$ ]] || die 2 "ci-settings line $n is not KEY=VALUE"
    k=${BASH_REMATCH[1]} v=${BASH_REMATCH[2]}
    [[ $seen != *" $k "* ]] || die 2 "ci-settings: $k is set twice"
    seen+="$k "
    case $k in
      CI_WIF_PROVIDER) WIF_PROVIDER=$v ;;
      CI_IDENTITY_PINS) pins=$v ;;
      CI_DENIAL_PROBES) DENIAL_PROBES=$v ;;
      *) die 2 "ci-settings: unknown key $k" ;;
    esac
  done <"$1"
  [[ $WIF_PROVIDER =~ ^projects/[0-9]{1,20}/locations/global/workloadIdentityPools/[a-z0-9-]{4,32}/providers/[a-z0-9-]{4,32}$ ]] ||
    die 2 "ci-settings: CI_WIF_PROVIDER must be projects/<number>/locations/global/workloadIdentityPools/<pool>/providers/<provider>"
  # The provider lives in the environment's own project (ci-federation.sh, pinned --project-number): its number is the
  # owner the metadata bucket must have (meta-lib.sh meta_presence), here and in every staging.sh subprocess.
  STAGING_META_PROJECT_NUMBER=${WIF_PROVIDER#projects/}
  export STAGING_META_PROJECT_NUMBER=${STAGING_META_PROJECT_NUMBER%%/*}
  [[ $STAGING_META_PROJECT_NUMBER =~ ^[1-9][0-9]{5,19}$ ]] || die 2 "ci-settings: CI_WIF_PROVIDER has no valid project number"
  # The launcher passes the owner's CI_DENIAL_PROBES (job variable, character-checked) through; the profile may list more.
  DENIAL_PROBES="$DENIAL_PROBES ${CI_DENIAL_PROBES:-}"
  DENIAL_PROBES="$(printf '%s' "${DENIAL_PROBES//,/ }" | xargs)"
  [[ -z $DENIAL_PROBES || $DENIAL_PROBES =~ ^(bucket:[a-z0-9][a-z0-9._-]{1,61}[a-z0-9])(\ bucket:[a-z0-9][a-z0-9._-]{1,61}[a-z0-9])*$ ]] ||
    die 2 "CI_DENIAL_PROBES must be bucket:<name> entries separated by spaces or commas"
  [[ $ACTION == deploy ]] || return 0
  # S16 (G057 hazard rule): every deploy pins all three identity secrets, each a version number or none (all-or-none,
  # as remote/lib.sh requires); latest, a missing or an extra name refuses.
  [[ -n $pins ]] || die 2 "ci-settings: CI_IDENTITY_PINS is required (IDENTITY_WORKOS_<NAME>=<version|none> for all three)"
  seen=" "
  for v in ${pins//,/ }; do
    [[ $v =~ ^(IDENTITY_WORKOS_[A-Z_]+)=([0-9]+|none)$ ]] || die 2 "ci-settings: CI_IDENTITY_PINS entry [${v%%=*}=...] must be NAME=<version number|none> (never latest)"
    name=${BASH_REMATCH[1]} ver=${BASH_REMATCH[2]}
    [[ " ${CI_IDENTITY_NAMES[*]} " == *" $name "* ]] || die 2 "ci-settings: CI_IDENTITY_PINS names unknown secret $name"
    [[ $seen != *" $name "* ]] || die 2 "ci-settings: CI_IDENTITY_PINS pins $name twice"
    seen+="$name "
    if [[ $ver == none ]]; then none=$((none + 1)); else numeric=$((numeric + 1)); fi
    PINS_ARGS+=(--pin-secret "$name=$ver")
  done
  for name in "${CI_IDENTITY_NAMES[@]}"; do [[ $seen == *" $name "* ]] || die 2 "ci-settings: CI_IDENTITY_PINS lacks $name"; done
  ((numeric == 0 || none == 0)) || die 2 "ci-settings: CI_IDENTITY_PINS must pin all three identity secrets to versions or all three to none"
}

# ---------------------------------------------------------------------------------------------------- candidate (S5/S6)
git_q() { git -C "$REPO" "$@" 2>>"$RAW/git.log"; }

# on_first_parent SHA: SHA is on the control's first-parent history. The list is read whole first: a pipe into grep -q
# would end rev-list with SIGPIPE on a match and pipefail would report that as a failure.
on_first_parent() { local h; h="$(git_q rev-list --first-parent "$CTL")" || return 1; grep -qxF -- "$1" <<<"$h"; }

# resolve_candidate: CAND (and BRANCH) from the selection (Plan §10.2): empty or "staging" = the frozen pipeline commit
# for both environments; staging otherwise only a full SHA on first-parent staging history; a preview also a full SHA
# reachable from a same-repository branch (refs/heads/* only, fetched as refs/remotes/origin/*) or a branch name,
# resolved once here and pinned (the bundle carries it to later jobs).
resolve_candidate() {
  local s=$SELECTION
  BRANCH=""
  if [[ -z $s || $s == staging ]]; then
    CAND=$CTL
  elif [[ $ENVSEL == staging ]]; then
    is_sha "$s" || die 2 "staging accepts only an empty candidate, staging or a full SHA on first-parent staging history"
    on_first_parent "$s" ||
      die 2 "candidate $s is not on first-parent staging history (an unmerged or intermediate commit is refused for staging)"
    CAND=$s
  elif is_sha "$s"; then
    git_q cat-file -e "$s^{commit}" || die 2 "candidate $s is not a commit of this repository"
    [[ -n "$(git_q for-each-ref --contains "$s" --format '%(refname)' refs/remotes/origin/)" ]] ||
      die 2 "candidate $s is not reachable from a same-repository branch (merge-request and fork refs are never accepted)"
    CAND=$s
  else
    git_q check-ref-format --branch "$s" >/dev/null || die 2 "candidate [$s] is not a branch name"
    CAND="$(git_q rev-parse --verify -q "refs/remotes/origin/$s^{commit}")" || die 2 "branch $s does not exist in this repository"
    BRANCH=$s
  fi
  is_sha "$CAND" || die 2 "candidate resolution failed"
  SUM[candidate]=$CAND SUM[selection]=${SELECTION:-default} SUM[branch]=${BRANCH:--}
}

# compat_gate SHA (S6): the candidate's deployment infrastructure is byte-identical to the trusted control's.
compat_gate() {
  local diff
  git_q cat-file -e "$1^{commit}" || die 2 "candidate $1 is not available locally"
  diff="$(git_q diff --name-only "$CTL" "$1" -- "${CI_COMPAT_PATHS[@]}")" || die 2 "the compatibility check failed"
  [[ -z $diff ]] ||
    die 2 "incompatible candidate: deployment infrastructure differs from trusted control: ${diff//$'\n'/ }; integrate it to staging first"
}

# verify_prepare: the prepare job's artifacts against this pipeline (phase A of check, build and deploy): candidate.tsv
# fields, the bundle's single head = the recorded candidate, the selection policy re-checked without re-resolving a
# branch, S6 again; the candidate's schema recomputed from the bundle (never trusted from the artifact).
declare -A CT=()
verify_prepare() {
  local prep k v heads
  prep="$(prep_dir)"
  [[ -f $prep/candidate.tsv && -f $prep/candidate.bundle ]] || die 2 "the prepare job's artifacts are missing"
  CT=()
  while IFS=$'\t' read -r k v; do [[ -z $k ]] || CT[$k]=$v; done <"$prep/candidate.tsv"
  [[ ${CT[environment]:-} == "$ENVSEL" && ${CT[pipeline]:-} == "$CI_PIPELINE_ID" && ${CT[control]:-} == "$CTL" ]] ||
    die 2 "candidate.tsv is not from this pipeline, environment and control commit"
  [[ ${CT[selection]-x} == "$SELECTION" ]] || die 2 "candidate.tsv records another selection than this pipeline's input"
  CAND=${CT[candidate]:-} BRANCH=${CT[branch]:--}
  [[ $BRANCH != - ]] || BRANCH=""
  is_sha "$CAND" || die 2 "candidate.tsv has no candidate"
  git_q bundle verify "$prep/candidate.bundle" >/dev/null || die 2 "the candidate bundle does not verify"
  heads="$(git_q bundle list-heads "$prep/candidate.bundle")" || die 2 "the candidate bundle cannot be read"
  [[ $heads == "$CAND refs/g072/candidate" ]] || die 2 "the candidate bundle's head is not the recorded candidate"
  git_q fetch --no-tags "$prep/candidate.bundle" "+refs/g072/candidate:refs/g072/candidate" >/dev/null || die 2 "the candidate bundle could not be fetched"
  if [[ -z $SELECTION || $SELECTION == staging ]]; then
    [[ $CAND == "$CTL" ]] || die 2 "the default candidate must be this pipeline's commit"
  elif [[ $ENVSEL == staging ]]; then
    on_first_parent "$CAND" || die 2 "candidate $CAND is not on first-parent staging history"
  elif is_sha "$SELECTION"; then
    [[ $CAND == "$SELECTION" && -z $BRANCH ]] || die 2 "candidate.tsv does not match the selected SHA"
  else
    [[ $BRANCH == "$SELECTION" ]] || die 2 "candidate.tsv does not match the selected branch"
  fi
  compat_gate "$CAND"
  schema_of "$REPO" "$CAND" >"$PRIV/candidate-schema.tsv" || die 2 "the candidate's migrations could not be listed"
  SUM[candidate]=$CAND SUM[selection]=${SELECTION:-default} SUM[branch]=${BRANCH:--}
}

# verify_build: the build job's artifacts when it built (checksums, candidate); BUILT=1, or 0 when it skipped.
BUILT=0
verify_build() {
  local b
  b="$(build_dir)"
  BUILT=0
  [[ -d $b && ! -f $b/skipped ]] || return 0
  [[ -f $b/SHA256SUMS && -f $b/images.tar.gz && -f $b/build-manifest.tsv ]] || die 2 "the build job's artifacts are incomplete"
  (cd "$b" && sha256sum --quiet -c SHA256SUMS) >/dev/null 2>>"$RAW/ci.log" || die 2 "the build artifacts do not match their checksums"
  [[ "$(awk -F'\t' '$1 == "candidate" {print $2}' "$b/build-manifest.tsv")" == "$CAND" ]] || die 2 "the build manifest is for another candidate"
  BUILT=1
}

# ------------------------------------------------------------------------------------------------- phase B: credentials
# cred_setup: runner-local federation files (Plan §2.5): gcloud's external_account file reads the ID token file in
# place; Terraform's (deploy) reads its copy inside PRIVATE_DIR, which tf() mounts at /private. No service-account key,
# no impersonation URL: the single impersonation hop stays GFLAGS / GOOGLE_IMPERSONATE_SERVICE_ACCOUNT (A3).
cred_setup() {
  local tok=${STAGING_CI_ID_TOKEN_FILE:-}
  [[ $tok == /* && -f $tok && ! -L $tok && -s $tok ]] || die 2 "STAGING_CI_ID_TOKEN_FILE is missing or empty"
  [[ "$(stat -c '%a' "$tok")" == 600 ]] || die 2 "the ID token file must be mode 0600"
  [[ "$(grep -c . "$tok")" == 1 ]] || die 2 "the ID token file must hold one line"
  export HOME=$PRIV/home CLOUDSDK_CONFIG=$PRIV/home/gcloud CLOUDSDK_CORE_DISABLE_PROMPTS=1
  install -d -m 0700 "$CLOUDSDK_CONFIG"
  ext_account "$tok" >"$PRIV/gcloud-adc.json"
  chmod 0600 "$PRIV/gcloud-adc.json"
  if [[ $ACTION == deploy ]]; then
    install -m 0600 "$tok" "$PRIV/private/ci/id-token"
    ext_account /private/ci/id-token >"$PRIV/private/ci/tf-adc.json"
    chmod 0600 "$PRIV/private/ci/tf-adc.json"
  fi
  show gcloud auth login --cred-file "$PRIV/gcloud-adc.json" --quiet
  gcloud auth login --cred-file="$PRIV/gcloud-adc.json" --quiet >/dev/null 2>&1 || die 2 "gcloud could not load the federation credential file"
}

ext_account() {
  printf '{\n  "type": "external_account",\n  "audience": "//iam.googleapis.com/%s",\n' "$WIF_PROVIDER"
  printf '  "subject_token_type": "urn:ietf:params:oauth:token-type:jwt",\n  "token_url": "https://sts.googleapis.com/v1/token",\n'
  printf '  "credential_source": {"file": "%s"}\n}\n' "$1"
}

meta_require() {
  meta_presence
  [[ $META_STATE == present ]] || die 2 "gs://$META_BUCKET does not exist: $ENVSEL is not CI-enabled (infra/terraform/bootstrap/ci-federation.sh)"
}

# manifest_check: MAN_OK=1 when deploy/<CAND>.tsv is a valid published manifest for this profile whose seven digests
# exist in the environment's registry (Plan §10.8); MAN_WHY otherwise. MAN_FILE holds the download.
manifest_check() {
  local svc ref rc
  MAN_OK=0 MAN_WHY="" MAN_FILE=$PRIV/manifest.tsv MAN_SHA="" MAN_GEN=0
  meta_read "deploy/$CAND.tsv" "$MAN_FILE" || die 2 "deploy/$CAND.tsv could not be read ($(meta_class))"
  MAN_GEN=$META_GEN
  ((MAN_GEN > 0)) || { MAN_WHY="no published manifest for $CAND"; return 0; }
  MAN_SHA="$(sha256sum "$MAN_FILE" | cut -d' ' -f1)"
  manifest_valid "$MAN_FILE" "$CAND" || return 0
  for svc in "${BUILT_SERVICES[@]}"; do
    ref="$(manifest_ref "$MAN_FILE" "$svc")"
    rc=0
    show gcloud artifacts docker images describe "$ref" --format 'value(image_summary.digest)' "${GFLAGS[@]}"
    gcloud artifacts docker images describe "$ref" --format 'value(image_summary.digest)' "${GFLAGS[@]}" >/dev/null 2>"$PRIV/ar.err" || rc=$?
    if ((rc)); then
      # The reference (its sha256 digest may hold "404") is removed first; no pipe into grep -q under pipefail.
      grep -qE 'NOT_FOUND|(^|[^0-9A-Za-z])404([^0-9A-Za-z]|$)' <<<"$(sed -E 's%[^[:space:]]*@sha256:[0-9a-f]+%<ref>%g' "$PRIV/ar.err")" ||
        die 2 "the registry could not be read for $svc"
      MAN_WHY="digest of $svc is not in the registry"
      return 0
    fi
  done
  MAN_OK=1
}

# manifest_valid FILE SHA: candidate line = SHA; exactly one image row per built service, from SHA, with an image ID
# and the digest reference <AR_REPOSITORY>/<service>@sha256:<64 hex> (repository = profile, digest shape).
manifest_valid() {
  local f=$1 c=$2 svc rows
  [[ "$(awk -F'\t' '$1 == "candidate" {print $2}' "$f")" == "$c" ]] || { MAN_WHY="the manifest's candidate is not $c"; return 1; }
  for svc in "${BUILT_SERVICES[@]}"; do
    rows="$(awk -F'\t' -v s="$svc" '$1 == "image" && $2 == s' "$f")"
    [[ -n $rows && $rows != *$'\n'* ]] || { MAN_WHY="the manifest has no single image row for $svc"; return 1; }
    tsv_fields "$rows"
    ((${#FIELDS[@]} == 6)) || { MAN_WHY="the $svc row has ${#FIELDS[@]} fields"; return 1; }
    [[ ${FIELDS[2]} == "$c" ]] || { MAN_WHY="$svc comes from another candidate"; return 1; }
    [[ ${FIELDS[4]} =~ ^sha256:[0-9a-f]{64}$ ]] || { MAN_WHY="$svc has no image ID"; return 1; }
    [[ ${FIELDS[5]} == "${CONF[AR_REPOSITORY]}/$svc@sha256:"* && ${FIELDS[5]##*@sha256:} =~ ^[0-9a-f]{64}$ ]] ||
      { MAN_WHY="$svc has no digest in the profile's repository"; return 1; }
  done
}

# ------------------------------------------------------------------------------------------------ lineage (Plan §10.4)
# members_load DIR: the possible-lineage set into DIR/member-<n>.tsv: lineage.tsv (when present) and the schema of every
# fence entry, each verified against the sha256 the entry recorded (a mismatch HOLDs). MEMBERS = the count;
# LINEAGE_STATE = verified | assumed | "" (none); FENCE_N = open entries.
MEMBERS=0 LINEAGE_STATE="" FENCE_N=0 LIN_GEN=0 FENCE_GEN=0
members_load() {
  local d=$1 kind t holder sha sum
  rm -rf "$d"; mkdir -p "$d"
  MEMBERS=0 LINEAGE_STATE="" FENCE_N=0
  meta_read lineage.tsv "$d/lineage.tsv" || die 2 "lineage.tsv could not be read ($(meta_class))"
  LIN_GEN=$META_GEN
  if ((LIN_GEN > 0)); then
    LINEAGE_STATE="$(awk -F'\t' '$1 == "state" {print $2}' "$d/lineage.tsv")"
    [[ $LINEAGE_STATE == verified || $LINEAGE_STATE == assumed ]] || die 3 "HOLD: lineage.tsv has no valid state; reconcile (STAGING.md)"
    MEMBERS=$((MEMBERS + 1)); cp "$d/lineage.tsv" "$d/member-$MEMBERS.tsv"
  fi
  meta_read fence.tsv "$d/fence.tsv" || die 2 "fence.tsv could not be read ($(meta_class))"
  FENCE_GEN=$META_GEN
  while IFS=$'\t' read -r kind t holder sha sum; do
    [[ -n $kind ]] || continue
    [[ $kind == fence ]] && is_sha "$sha" && [[ $sum =~ ^[0-9a-f]{64}$ ]] || die 3 "HOLD: fence.tsv holds an unreadable entry"
    meta_read "schema/$sha.tsv" "$d/schema.tmp" || die 2 "schema/$sha.tsv could not be read ($(meta_class))"
    ((META_GEN > 0)) && [[ "$(sha256sum "$d/schema.tmp" | cut -d' ' -f1)" == "$sum" ]] ||
      die 3 "HOLD: schema/$sha.tsv does not match the fence entry of $holder at $t"
    MEMBERS=$((MEMBERS + 1)); FENCE_N=$((FENCE_N + 1)); mv "$d/schema.tmp" "$d/member-$MEMBERS.tsv"
    printf '%s\t%s\t%s\n' "$sha" "$holder" "$t" >>"$d/fence-entries"
  done <"$d/fence.tsv"
}

# s7_ok CANDIDATE MEMBER: r0's S7 rule against one member: the same blob for every version up to the member's maximum
# (present in both or absent from both), and the candidate's maximum >= the member's.
s7_ok() {
  awk -F'\t' '
    FNR == 1 { file++ }
    $1 ~ /^[0-9]+$/ { if (file == 1) { c[$1] = $2; if ($1 + 0 > cm) cm = $1 + 0 } else { m[$1] = $2; if ($1 + 0 > mm) mm = $1 + 0 } }
    END {
      if (cm < mm) exit 1
      for (v in m) if (v + 0 <= mm && c[v] != m[v]) exit 1
      for (v in c) if (v + 0 <= mm && c[v] != m[v]) exit 1
      exit 0
    }' "$1" "$2"
}

# lineage_gate: the candidate passes only against every member; a refusal names the fence entry or lineage.
lineage_gate() {
  local d=$PRIV/lineage i e
  members_load "$d"
  ((LIN_GEN > 0)) || die 2 "no recorded schema lineage for $ENVSEL: run ci.sh lineage init (STAGING.md) before the first CI deploy"
  for ((i = 1; i <= MEMBERS; i++)); do
    s7_ok "$PRIV/candidate-schema.tsv" "$d/member-$i.tsv" && continue
    if ((i == 1)); then die 2 "retained schema incompatible: candidate $CAND is older than or diverges from the $LINEAGE_STATE lineage; an explicit, separately authorized staging.sh reset --confirm $ENVSEL is required (not available in CI)"; fi
    e="$(sed -n "$((i - 1))p" "$d/fence-entries")"
    die 2 "schema fence unreconciled since ${e#*$'\t'} deployed ${e%%$'\t'*}; deploy that candidate or a successor, or reconcile (STAGING.md)"
  done
  SUM[fence_open]="$FENCE_N open"
}

lineage_query() {
  printf "cid=\$(sudo docker ps --quiet --filter label=com.docker.compose.project=%s --filter label=com.docker.compose.service=postgres | head -n 1) && test -n \"\$cid\" && sudo docker exec \"\$cid\" psql --no-psqlrc --tuples-only --no-align --username postgres --dbname %s --command '%s'" \
    "$ENV_NAME" "${CONF[DATABASE_NAME]}" "$CI_LINEAGE_SQL"
}

# observed_versions FILE: the applied versions read from the running VM (one per line, ascending). Status 1 when the
# read fails or answers anything but version numbers.
observed_versions() {
  local out
  out="$(vm_ssh "$(lineage_query)")" || return 1
  out="$(grep . <<<"$out" || true)"
  [[ -z $out ]] || ! grep -qvE '^[0-9]+$' <<<"$out" || return 1
  : >"$1"
  [[ -z $out ]] || sort -n <<<"$out" >"$1"
}

# lineage_write STATE SOURCE ROWS_FILE GEN: lineage.tsv := STATE from ROWS_FILE (<version><TAB><blob>) at generation GEN.
lineage_write() {
  local f=$PRIV/lineage.new
  { printf 'state\t%s\nsource\t%s\nwriter\t%s\ntime\t%s\n' "$1" "$2" "${LEASE_HOLDER:-$HOLDER}" "$(ts)"; grep -E '^[0-9]+'$'\t' "$3" || true; } >"$f"
  meta_put "$f" lineage.tsv "$4"
}

# ----------------------------------------------------------------------------------------------- staging.sh subprocesses
# run_staging STEP ARGS...: control staging.sh under the lease this job holds; stdout and stderr only into raw/STEP.log.
# The Terraform Cloudflare token reaches only the start that needs it (process environment, never an argument).
# The subprocess runs in the background and is waited for, so a cancellation's TERM (docker stop of the context) reaches
# ci_signal at once: the step is stopped (staging.sh exits through its own trap) and the EXIT trap releases the lease.
STAGING_RC=0 STAGING_PID=""
run_staging() {
  local step=$1
  shift
  STAGING_RC=0
  if [[ -n $CF_TOKEN && $step == start ]]; then
    CLOUDFLARE_API_TOKEN=$CF_TOKEN STAGING_LEASE_HELD="$LEASE_HOLDER@$LEASE_GEN" bash "$STAGING_SH" "$@" --config "$CI_CONF" \
      >>"$RAW/$step.log" 2>&1 </dev/null &
  else
    STAGING_LEASE_HELD="$LEASE_HOLDER@$LEASE_GEN" bash "$STAGING_SH" "$@" --config "$CI_CONF" >>"$RAW/$step.log" 2>&1 </dev/null &
  fi
  STAGING_PID=$!
  wait "$STAGING_PID" || STAGING_RC=$?
  STAGING_PID=""
  chmod 0600 "$RAW/$step.log"
}

ci_signal() {
  if [[ -n $STAGING_PID ]]; then kill -TERM "$STAGING_PID" 2>/dev/null || true; wait "$STAGING_PID" 2>/dev/null || true; fi
  STAGING_PID=""
  SUM[detail]="canceled (signal)"
  exit 143
}

iso_epoch() { date -u -d "$1" +%s 2>/dev/null; }

# fail_class STEP: the exit class of a failed staging.sh step (S14): its HOLD stays 3; a session that ended at its armed
# deadline during the action is 5 (never success); a refusal before this job changed anything stays 2; else 4.
fail_class() {
  local step=$1 rc=$STAGING_RC
  ((rc == 3)) && die 3 "HOLD: staging.sh $step reported a HOLD (see the published log)"
  if [[ $DEADLINE =~ ^[0-9]+$ ]] && vm_read 2>/dev/null && [[ $VMSTATE == TERMINATED ]] && (("$(meta_now)" >= DEADLINE)); then
    die 5 "the session expired during staging.sh $step (exit $rc): $VM is TERMINATED after its scheduled stop $(meta_iso "$DEADLINE"); not a success"
  fi
  ((rc == 2 && CHANGED == 0)) && die 2 "staging.sh $step refused before any change (see the published log)"
  die 4 "staging.sh $step failed (exit $rc) after this job changed the environment (see the published log)"
}

# ----------------------------------------------------------------------------------------------------- denial probes
# probe_tiam BUCKET: Cloud Storage testIamPermissions as this job's identity; the access token must be a non-empty,
# well-formed value (kept in the private canary list; on a pipe to curl, never an argument). PROBE_WHY explains a non-zero
# status. Status 0 = denied, affirmatively only: HTTP 403 or 404, or HTTP 200 with a testIamPermissionsResponse body that
# grants nothing; 1 = a permission is granted (finding); 2 = not proven (no or malformed token, transport error, 401,
# 5xx, any other status, an unparsable body): the caller fails closed.
PROBE_WHY=""
probe_tiam() {
  local tok code body=$PRIV/probe.body rc=0
  PROBE_WHY=""
  tok="$(gcloud auth print-access-token "${GFLAGS[@]}" 2>>"$RAW/ci.log")" || { PROBE_WHY="no access token for the probe"; return 2; }
  tok=${tok%%$'\n'*}
  [[ $tok =~ ^[A-Za-z0-9._~+/=-]{16,4096}$ ]] || { tok=""; PROBE_WHY="an empty or malformed access token for the probe"; return 2; }
  printf '%s\n' "$tok" >>"$CANARY_FILE"
  : >"$body"
  code="$(curl -sS --max-time 20 -o "$body" -w '%{http_code}' --config <(printf 'header = "Authorization: Bearer %s"\n' "$tok") \
    "https://storage.googleapis.com/storage/v1/b/$1/iam/testPermissions?permissions=storage.objects.get&permissions=storage.objects.list&permissions=storage.objects.create&permissions=storage.objects.delete" 2>>"$RAW/ci.log")" || rc=$?
  tok=""
  ((rc == 0)) || { PROBE_WHY="transport error (curl exit $rc)"; return 2; }
  case $code in
    403 | 404) return 0 ;;
    200) ;;
    401) PROBE_WHY="authentication refused (HTTP 401)"; return 2 ;;
    *) PROBE_WHY="HTTP ${code:-none}"; return 2 ;;
  esac
  case "$(probe_verdict "$body")" in
    denied) return 0 ;;
    granted) PROBE_WHY="granted"; return 1 ;;
    *) PROBE_WHY="unparsable or unexpected testIamPermissions answer"; return 2 ;;
  esac
}

# The 200 answer, parsed by jq (pinned 1.7.1, /usr/local/bin/jq in infra/docker/staging-ci.Dockerfile): exactly one JSON
# object of kind storage#testIamPermissionsResponse whose permissions key is absent or an array of strings. "denied"
# when that array is absent or empty, "granted" when it holds any string; anything else (invalid or truncated JSON, an
# empty body, several values, another kind or shape, a non-array permissions, a non-string element) is "invalid".
PROBE_JQ='if type == "object" and .kind == "storage#testIamPermissionsResponse" and ((has("permissions") | not) or ((.permissions | type) == "array" and all(.permissions[]; type == "string"))) then (if ((.permissions // []) | length) == 0 then "denied" else "granted" end) else "invalid" end'
probe_verdict() {
  local out rc=0
  out="$(jq -r "$PROBE_JQ" <"$1" 2>>"$RAW/ci.log")" || rc=$?
  if ((rc == 0)) && [[ $out == denied || $out == granted ]]; then printf '%s' "$out"; else printf 'invalid'; fi
}

# probe_or_die BUCKET LABEL: exit 6 (isolation finding, or isolation not proven) unless probe_tiam proves the denial.
probe_or_die() {
  local rc=0
  probe_tiam "$1" || rc=$?
  ((rc != 1)) || die 6 "isolation finding: the $ACTION identity holds object permissions on $2"
  ((rc == 0)) || die 6 "isolation not proven on $2: $PROBE_WHY (fail closed)"
}

# denial_probes: non-mutating isolation probes (Plan §10.3): the stop/read identity can neither impersonate the deployer
# (refused with a permission denial, not merely failed) nor touch the state bucket; no identity reaches an owner-listed
# sibling bucket. A success, or a probe that proves nothing, is exit 6.
denial_probes() {
  local b
  if [[ $ACTION == stop || $ACTION == check ]]; then
    if gcloud auth print-access-token --impersonate-service-account "${CONF[DEPLOYER_SA]}" >/dev/null 2>"$PRIV/imp.err"; then
      die 6 "isolation finding: the $ACTION identity could impersonate ${CONF[DEPLOYER_SA]}"
    fi
    grep -qE 'PERMISSION_DENIED|[Pp]ermission .*denied' "$PRIV/imp.err" ||
      die 6 "isolation not proven: impersonating ${CONF[DEPLOYER_SA]} failed without a permission denial (fail closed)"
    b=""
    [[ ! -f $PROF/backend.hcl ]] || b="$(sed -n -E 's/^[[:space:]]*bucket[[:space:]]*=[[:space:]]*"([a-z0-9._-]+)".*/\1/p' "$PROF/backend.hcl")"
    b=${b%%$'\n'*}
    probe_or_die "${b:-${CONF[GCP_PROJECT]}-tfstate}" "the Terraform state bucket"
  fi
  for b in $DENIAL_PROBES; do probe_or_die "${b#bucket:}" "${b#bucket:}"; done
  return 0
}

# --------------------------------------------------------------------------------------------------- lease (CI holder)
# ci_lease_gate: the lease observed in phase B: free, this job's own, or an expired CI lease (broken in phase C);
# a local lease, expired or not, and a live CI lease refuse.
ci_lease_gate() {
  lease_read
  BREAK=0
  ((META_GEN == 0)) && return 0
  [[ ${LEASE_F[holder]} == "$HOLDER" ]] && return 0
  if lease_expired_ci; then BREAK=1; return 0; fi
  die 2 "the environment lease is $(lease_describe); nothing changed"
}

ci_lease_take() {
  local cmd=$1
  if ((BREAK)); then
    lease_read
    lease_expired_ci || die 2 "the environment lease changed: $(lease_describe); nothing changed"
    lease_break "$HOLDER" "expired CI lease (past job timeout + 5 min)"
  fi
  lease_acquire "$HOLDER" "$cmd" "${CAND:--}" "${CI_LEASE_TTL[$cmd]}"
}

# oslogin_key: a fresh key registered with OS Login for this job's TTL (removed at exit). Status 1 when it could not be
# generated or registered: deploy treats that as a failure (exit 4); stop continues without it (Spec R5: the guest-side
# stop is best effort and the provider stop still runs).
oslogin_key() {
  KEY_FILE=$HOME/.ssh/google_compute_engine
  install -d -m 0700 "$HOME/.ssh"
  ssh-keygen -t ed25519 -N '' -q -C "$HOLDER" -f "$KEY_FILE" >/dev/null 2>&1 </dev/null || { log "ssh-keygen failed"; KEY_FILE=""; return 1; }
  run gcloud compute os-login ssh-keys add --key-file "$KEY_FILE.pub" --ttl "${CI_KEY_TTL[$ACTION]}" "${GFLAGS[@]}" >/dev/null ||
    { log "the OS Login key could not be registered"; KEY_FILE=""; return 1; }
}

# --------------------------------------------------------------------------------------------------------- metadata
# put_new FILE KEY: a create-only metadata object (runs/, receipts/, logs/, stop-requests/, deploy/).
put_new() { meta_put "$1" "$2" 0; }

cur_write() { # KEY=VALUE... over the previous current.tsv, at its observed generation
  local f=$PRIV/current.new kv k
  declare -A CUR=()
  meta_read current.tsv "$PRIV/current.old" || die 4 "current.tsv could not be read ($(meta_class))"
  while IFS=$'\t' read -r k kv; do [[ -z $k ]] || CUR[$k]=$kv; done <"$PRIV/current.old"
  for kv in "$@"; do CUR[${kv%%=*}]=${kv#*=}; done
  : >"$f"
  for k in environment candidate selection branch actor pipeline job manifest_sha256 manifest_generation schema status_log \
    session_seconds scheduled_stop started_at stopped_at state receipt last_action last_pipeline last_job last_actor; do
    printf '%s\t%s\n' "$k" "${CUR[$k]:--}" >>"$f"
  done
  meta_put "$f" current.tsv "$META_GEN" || die 4 "current.tsv could not be written at generation $META_GEN ($(meta_class): $(meta_why))"
}

run_record() { # OUTCOME: runs/<pipeline>-<job>.tsv (create-only), once
  local f=$PRIV/run.tsv k
  ((RUN_DONE == 0)) || return 0
  RUN_DONE=1
  {
    printf 'outcome\t%s\n' "$1"
    for k in action environment candidate selection branch actor pipeline job reuse digests session scheduled_stop readiness fence_open receipt_id receipt_sha256 detail; do
      printf '%s\t%s\n' "$k" "${SUM[$k]:--}"
    done
    printf 'lease_generation\t%s\n' "${LEASE_GEN:--}"
  } >"$f"
  put_new "$f" "runs/$CI_PIPELINE_ID-$CI_JOB_ID.tsv" || log "the run record could not be written ($(meta_class))"
}

# publish_logs: the sanitized log (publish_filter over every raw step log) to logs/<pipeline>-<job>.log, after
# canary_scan of the raw and filtered files; a hit publishes nothing and is exit 6.
publish_logs() {
  local all=$RAW/all.raw out=$PRIV/publish.log f
  ((LOGS_DONE == 0)) || return 0
  LOGS_DONE=1
  : >"$all"
  for f in "$RAW"/*.log; do [[ -f $f ]] && cat "$f" >>"$all"; done
  publish_filter "$all" "$out"
  scan_all "$RAW"/*.log "$all" "$out" || die 6 "privacy canary in the job's output: nothing published"
  put_new "$out" "logs/$CI_PIPELINE_ID-$CI_JOB_ID.log" || log "the sanitized log could not be written ($(meta_class))"
  SUM[status_log]="logs/$CI_PIPELINE_ID-$CI_JOB_ID.log"
}

# scan_all FILE...: ci-publish.sh canary_scan (credential shapes, the plan canary outside raw/, and the exact values this
# job holds: STAGING_CI_ID_TOKEN_FILE and STAGING_CI_CANARY_LITERALS). Its report lines go to the private log only.
scan_all() { canary_scan "$@" >>"$RAW/canary.log" 2>&1; }

# ------------------------------------------------------------------------------------------------------- deploy steps
# deploy_gates: phase B's reads and gates (Plan §10.6), re-run after the lease is taken: supersession, the lease, the
# lineage set, a valid manifest or this pipeline's build artifacts, the VM state, duration and remaining time.
deploy_gates() {
  local max=0 k id now
  meta_list stop-requests/ || die 2 "stop-requests/ could not be listed ($(meta_class))"
  for k in "${KEYS[@]}"; do
    id=${k#stop-requests/}
    [[ $id =~ ^[0-9]{1,19}$ ]] && ((id > max)) && max=$id
  done
  ((max <= CI_PIPELINE_ID)) ||
    die 2 "superseded: stop pipeline $max was requested after this pipeline ($CI_PIPELINE_ID); nothing changed"
  ci_lease_gate
  lineage_gate
  manifest_check
  if ((MAN_OK)); then REUSE=yes
  elif ((BUILT)); then REUSE=no
  else die 2 "no valid published manifest for $CAND (${MAN_WHY}) and no build artifacts in this pipeline; nothing changed"; fi
  vm_read || die 3 "HOLD: the state of $VM could not be read; nothing changed"
  [[ $VMSTATE != ABSENT ]] || die 2 "$VM does not exist (provision it with the G070 local path); nothing changed"
  vm_session || die 3 "HOLD: the session of $VM could not be read; nothing changed"
  [[ $SESSION_HAVE =~ ^[1-9][0-9]*$ ]] || die 3 "HOLD: $VM reports no max run duration; nothing changed"
  NEED_TOKEN=0
  case $VMSTATE in
    RUNNING)
      [[ -z $WANT || $WANT == "$SESSION_HAVE" ]] ||
        die 2 "$VM runs a ${SESSION_HAVE}s session; a different duration needs the stopped VM (run stop first); nothing changed"
      DEADLINE="$(iso_epoch "$SESSION_STOP")" || DEADLINE=""
      [[ $DEADLINE =~ ^[0-9]+$ ]] || die 2 "$VM runs without a readable scheduled stop; nothing changed"
      now="$(meta_now)"
      ((DEADLINE - now >= CI_MIN_REMAINING)) ||
        die 2 "only $((DEADLINE - now))s remain before the scheduled stop (at least ${CI_MIN_REMAINING}s required); nothing changed"
      ;;
    TERMINATED) [[ ${WANT:-$CI_SESSION_SECONDS} == "$SESSION_HAVE" ]] || NEED_TOKEN=1 ;;
    *) die 2 "$VM is $VMSTATE; nothing changed" ;;
  esac
}

# remaining_guard: after a start, before the deploy: the scheduled stop read back and at least CI_MIN_REMAINING left.
remaining_guard() {
  local now
  vm_session || die 4 "the session of $VM could not be read after the start"
  DEADLINE="$(iso_epoch "$SESSION_STOP")" || DEADLINE=""
  [[ $DEADLINE =~ ^[0-9]+$ ]] || die 4 "$VM reports no scheduled stop"
  now="$(meta_now)"
  ((DEADLINE - now >= CI_MIN_REMAINING)) || die 4 "only $((DEADLINE - now))s remain before the scheduled stop"
  SUM[session]="${SESSION_HAVE}s" SUM[scheduled_stop]=$SESSION_STOP
}

# deploy_images: reuse the verified published manifest (no load, no push), or load this pipeline's build artifacts,
# verify every image ID against the build manifest, retag ID-preserving to the profile's repository (G070 R-1/R-1b),
# staging.sh push (re-verifies the IDs, records digests) and publish deploy/<sha>.tsv create-only.
deploy_images() {
  local m=$PRIV/evidence/deploy/$CAND.tsv b svc want have ph
  install -d -m 0700 "$PRIV/evidence/deploy"
  if [[ $REUSE == yes ]]; then
    install -m 0600 "$MAN_FILE" "$m"
    SUM[reuse]="yes (published manifest, generation $MAN_GEN)"
    return 0
  fi
  b="$(build_dir)"
  run docker load --input "$b/images.tar.gz" >>"$RAW/images.log" || die 2 "the build artifacts could not be loaded; nothing changed"
  for svc in "${BUILT_SERVICES[@]}"; do
    want="$(awk -F'\t' -v s="$svc" '$1 == "image" && $2 == s {print $5}' "$b/build-manifest.tsv")"
    ph=$CI_PLACEHOLDER_AR/$svc:$CAND
    have="$(docker image inspect --format '{{.Id}}' "$ph" 2>>"$RAW/images.log")" || have=""
    [[ $want =~ ^sha256:[0-9a-f]{64}$ && $have == "$want" ]] || die 2 "$svc: the loaded image is not the one the build recorded; refused before push"
    run docker tag "$ph" "${CONF[AR_REPOSITORY]}/$svc:$CAND" >>"$RAW/images.log" || die 2 "$svc could not be tagged; refused before push"
    have="$(docker image inspect --format '{{.Id}}' "${CONF[AR_REPOSITORY]}/$svc:$CAND" 2>>"$RAW/images.log")" || have=""
    [[ $have == "$want" ]] || die 2 "$svc: the retag changed the image ID; refused before push"
  done
  install -m 0600 "$b/build-manifest.tsv" "$m"
  CHANGED=1
  run_staging push push --candidate "$CAND"
  ((STAGING_RC == 0)) || die 4 "staging.sh push failed (exit $STAGING_RC)"
  manifest_valid "$m" "$CAND" || die 4 "the pushed manifest is invalid: $MAN_WHY"
  put_new "$m" "deploy/$CAND.tsv" || die 4 "deploy/$CAND.tsv could not be published ($(meta_class))"
  meta_gen "deploy/$CAND.tsv" || true
  MAN_GEN=$GEN MAN_SHA="$(sha256sum "$m" | cut -d' ' -f1)"
  SUM[reuse]="no (built in this pipeline, pushed, published)"
}

# fetch_cf_token: the Terraform Cloudflare token (G070 r4 secret <env>-TERRAFORM_CLOUDFLARE_API_TOKEN), read by the
# federated start principal itself (no impersonation, A4), only for a session-length change or the start's own request
# for it (start_if_needed); kept in this process for the one start that needs it, and in the private canary list.
fetch_cf_token() {
  local s="$ENV_NAME-TERRAFORM_CLOUDFLARE_API_TOKEN"
  show gcloud secrets versions access latest --secret "$s" --project "${CONF[GCP_PROJECT]}"
  CF_TOKEN="$(gcloud secrets versions access latest --secret "$s" --project "${CONF[GCP_PROJECT]}" 2>>"$RAW/ci.log")" ||
    die 4 "the Terraform Cloudflare token could not be read (needed for the session-length change)"
  [[ -n $CF_TOKEN && $CF_TOKEN != *[[:space:]]* ]] || die 4 "the Terraform Cloudflare token is empty or malformed"
  printf '%s\n' "$CF_TOKEN" >>"$CANARY_FILE"
}

# start_once: one staging.sh start (with --duration when one was asked for), under this job's lease.
start_once() {
  if [[ -n $WANT ]]; then run_staging start start --duration "$WANT"; else run_staging start start; fi
}

# start_if_needed: a stopped VM that already reports the requested length while Terraform records another (an earlier
# set-scheduling whose refresh did not complete) is recovered by start through a refresh that needs the token, which this
# job cannot know beforehand (it reads the instance, not Terraform's record). start asks for the token before any write
# (exit 2, nothing changed), so exactly that refusal is retried once with the token read then; any other result, and a
# second refusal, fails as before.
start_if_needed() {
  [[ $VMSTATE == TERMINATED ]] || return 0
  ((NEED_TOKEN == 0)) || fetch_cf_token
  CHANGED=1
  start_once
  if ((STAGING_RC == 2)) && [[ -z $CF_TOKEN ]] && grep -qF "set CLOUDFLARE_API_TOKEN (Terraform's zone-scoped token) in this shell first" "$RAW/start.log" >/dev/null; then
    fetch_cf_token
    start_once
  fi
  CF_TOKEN=""
  ((STAGING_RC == 0)) || fail_class start
}

# readiness_check: success only when the deploy's VM-side lines report READY for the candidate and every running
# app/edge digest equals the manifest's.
readiness_check() {
  local m=$PRIV/evidence/deploy/$CAND.tsv l=$RAW/deploy.log svc ref d=""
  grep -qxF "$(printf 'result\tREADY\t%s' "$CAND")" "$l" || die 4 "the deploy did not report READY for $CAND"
  for svc in core web intelligence edge; do
    ref="$(manifest_ref "$m" "$svc")"
    grep -qxF "$(printf 'digest\t%s\t%s\tmatch' "$svc" "$ref")" "$l" || die 4 "the running $svc image is not the manifest's digest"
  done
  for svc in "${BUILT_SERVICES[@]}"; do ref="$(manifest_ref "$m" "$svc")"; ref=${ref##*@sha256:}; d+="$svc=sha256:${ref:0:12} "; done
  SUM[readiness]="READY $CAND" SUM[digests]="${d% }"
}

# verify_observed_lineage (clearer 1 of Plan §10.4): the applied versions read from the VM must be exactly the
# candidate's; then lineage.tsv := verified and the fence is deleted, both at their observed generations.
verify_observed_lineage() {
  local obs=$PRIV/observed.txt want=$PRIV/want.txt rows=$PRIV/lineage.rows
  observed_versions "$obs" || die 4 "the observed schema lineage could not be read; the fence is kept (reconcile: STAGING.md)"
  awk -F'\t' '$1 ~ /^[0-9]+$/ {print $1}' "$PRIV/candidate-schema.tsv" | sort -n >"$want"
  cmp -s "$obs" "$want" ||
    die 4 "the observed schema lineage is not exactly $CAND's migrations; the fence is kept (reconcile: STAGING.md)"
  awk -F'\t' '$1 ~ /^[0-9]+$/' "$PRIV/candidate-schema.tsv" >"$rows"
  meta_gen lineage.tsv || die 4 "lineage.tsv could not be read"
  lineage_write verified "$CAND" "$rows" "$GEN" || die 4 "lineage.tsv could not be written ($(meta_class))"
  meta_gen fence.tsv || die 4 "fence.tsv could not be read"
  if ((GEN > 0)); then meta_rm fence.tsv "$GEN" || die 4 "fence.tsv could not be cleared ($(meta_class))"; fi
  SUM[fence_open]="0 open (lineage verified at $CAND)"
}

# receipt_key ID FILE: receipts/<id>-<sha256 prefix>.tsv: G070 receipt IDs count per evidence directory, so the content
# digest keeps every job's and operator's receipt distinct (an identical receipt maps to the same key).
receipt_key() { printf 'receipts/%s-%s.tsv' "$1" "$(sha256sum "$2" | cut -c1-12)"; }

# receipt_publish LOG: the G070 receipt line of a staging.sh step's log; the receipt file is verified, scanned and
# published create-only.
receipt_publish() {
  local line rc=0
  line="$(grep -E $'^receipt\t[a-z0-9-]+\t/[^\t]+\tsha256 [0-9a-f]{64}$' "$1" | tail -n 1)" || die 4 "no receipt in the staging.sh output"
  tsv_fields "$line"
  RECEIPT_ID=${FIELDS[1]} RECEIPT_FILE=${FIELDS[2]} RECEIPT_SHA=${FIELDS[3]#sha256 }
  [[ $RECEIPT_FILE == "$PRIV/evidence/receipts/"* && -f $RECEIPT_FILE ]] || die 4 "the receipt is not in this job's evidence directory"
  [[ "$(sha256sum "$RECEIPT_FILE" | cut -d' ' -f1)" == "$RECEIPT_SHA" ]] || die 4 "the receipt does not match its sha256"
  scan_all "$RECEIPT_FILE" || die 6 "privacy canary in the receipt; nothing published"
  put_new "$RECEIPT_FILE" "$(receipt_key "$RECEIPT_ID" "$RECEIPT_FILE")" || rc=$?
  ((rc == 0 || rc == 1)) || die 4 "the receipt could not be published ($(meta_class))"
  SUM[receipt_id]="$RECEIPT_ID ($(receipt_key "$RECEIPT_ID" "$RECEIPT_FILE"))" SUM[receipt_sha256]=$RECEIPT_SHA
}

# ---------------------------------------------------------------------------------------------------------- stop steps
# stop_request: stop-requests/<pipeline> (create-only intent record; a retry of this pipeline finds it present). Written
# before the lease, so an older queued start can never restart the environment after this stop request (S12).
stop_request() {
  local f=$PRIV/stop-request.tsv rc=0
  printf 'pipeline\t%s\njob\t%s\nactor\t%s\ntime\t%s\n' "$CI_PIPELINE_ID" "$CI_JOB_ID" "$GITLAB_USER_LOGIN" "$(ts)" >"$f"
  put_new "$f" "stop-requests/$CI_PIPELINE_ID" || rc=$?
  ((rc == 0 || rc == 1)) || die 4 "the stop request could not be written ($(meta_class): $(meta_why))"
}

# runs_overwrite_probe: the stop identity must not overwrite an object outside current.tsv and lease: its own run record,
# at the observed generation, must be denied (403). A success, or any other answer, is an isolation finding (exit 6).
runs_overwrite_probe() {
  local k="runs/$CI_PIPELINE_ID-$CI_JOB_ID.tsv" rc=0
  meta_gen "$k" || die 6 "the run record could not be read for the overwrite probe"
  ((GEN > 0)) || die 6 "no run record to probe"
  meta_put "$PRIV/run.tsv" "$k" "$GEN" || rc=$?
  ((rc != 0)) || die 6 "isolation finding: the stop identity overwrote $k"
  [[ "$(meta_class)" == denied ]] || die 6 "the overwrite probe of $k was not denied ($(meta_class))"
}

# ---------------------------------------------------------------------------------------------------------- build step
# build_images: control staging.sh build --source <candidate worktree> against a synthetic placeholder config, under
# env -i (PATH, HOME and DOCKER_* only); the seven images saved, the build manifest and checksums as artifacts.
build_images() {
  local syn=$PRIV/build v envs=() svc tags=() rc=0 m
  WORKTREE=$PRIV/src
  # The job's umask 077 would check the sources out as 0600/0700, and a Dockerfile COPY keeps those modes: the image's
  # non-root user could not read /srv/app/src. The tree sits in the 0700 private directory, so 022 exposes nothing.
  (umask 022 && git_q worktree add --detach "$WORKTREE" "$CAND" >/dev/null) || die 2 "the candidate worktree could not be created"
  install -d -m 0700 "$syn" "$syn/evidence" "$syn/private" "$syn/home"
  printf '%s\n' STAGING_ENV=cibuild "GCP_PROJECT=$CI_PLACEHOLDER_PROJECT" GCP_REGION=us-central1 GCP_ZONE=us-central1-a \
    "DEPLOYER_SA=cibuild-deployer@$CI_PLACEHOLDER_PROJECT.iam.gserviceaccount.com" "AR_REPOSITORY=$CI_PLACEHOLDER_AR" \
    BASE_DOMAIN=build.example.net CLOUDFLARE_ZONE_ID=00000000000000000000000000000000 ACME_EMAIL=build@example.net \
    "EDGE_ACME_CA=$LE_STAGING_CA" EDGE_DATA_VOLUME=edge-data-rehearsal PRODUCT_NAME=build-placeholder \
    DATABASE_NAME=build_placeholder OBJECT_STORE_BUCKET=build-placeholder \
    "TERRAFORM_IMAGE=hashicorp/terraform:0.0.0@sha256:0000000000000000000000000000000000000000000000000000000000000000" \
    "EVIDENCE_DIR=$syn/evidence" "PRIVATE_DIR=$syn/private" "GOOGLE_ADC_FILE=$syn/private/none.json" "OWNER=ci build" \
    >"$syn/staging.conf"
  for v in $(compgen -e); do [[ $v != DOCKER_* ]] || envs+=("$v=${!v}"); done # Plan S8: PATH, HOME, DOCKER_* only
  env -i PATH="$PATH" HOME="$syn/home" "${envs[@]}" bash "$STAGING_SH" build --config "$syn/staging.conf" --candidate "$CAND" \
    --source "$WORKTREE" >"$RAW/build.log" 2>&1 </dev/null || rc=$?
  chmod 0600 "$RAW/build.log"
  ((rc == 0)) || die "$([[ $rc == 2 ]] && echo 2 || echo 4)" "staging.sh build failed (exit $rc; see the job's raw log)"
  m=$syn/evidence/deploy/$CAND.tsv
  [[ -f $m ]] || die 4 "the build wrote no manifest"
  for svc in "${BUILT_SERVICES[@]}"; do tags+=("$CI_PLACEHOLDER_AR/$svc:$CAND"); done
  { run docker save "${tags[@]}" | gzip -n >"$OUT/images.tar.gz"; } 2>>"$RAW/images.log" || die 4 "docker save failed"
  install -m 0644 "$m" "$OUT/build-manifest.tsv"
  (cd "$OUT" && sha256sum images.tar.gz build-manifest.tsv >SHA256SUMS)
  SUM[reuse]="no (built $CAND)"
}

# ------------------------------------------------------------------------------------------------------- exit and summary
# summary_publish RC: deploy and stop write the fixed-field summary STAGING_CI_OUT/summary.txt (ci-publish.sh
# summary_render, SUMMARY_KEYS); the launcher prints it. A canary in it withholds every field but the exit class.
summary_publish() { # RC
  local rc=$1 s=$OUT/summary.txt args=() k
  [[ $ACTION == deploy || $ACTION == stop ]] || return 0
  SUM[exit_class]=$rc
  for k in $SUMMARY_KEYS; do
    [[ -z ${SUM[$k]:-} ]] || args+=("$k=${SUM[$k]//$'\n'/ }")
  done
  summary_render "$s.tmp" "${args[@]}" || { printf 'ci: the summary could not be rendered\n' >&3; return 6; }
  if ! scan_all "$s.tmp"; then
    rm -f "$s.tmp"
    summary_render "$s" "action=$ACTION" "environment=$ENVSEL" "exit_class=6" || true
    return 6
  fi
  mv "$s.tmp" "$s"
}

# ci_exit: the EXIT trap of every job: the lease this job holds is released (a failure is exit 4 at best), the OS Login
# key removed, the candidate worktree and every credential file removed; failures after phase C publish their sanitized
# log and run record; the summary is rendered and scanned; raw/ is removed and its absence verified.
ci_exit() {
  local rc=$? r
  trap - EXIT
  set +e
  if ((WRITES)) && [[ $META_STATE == present ]]; then
    if ((LOGS_DONE == 0)); then (publish_logs) || rc=6; LOGS_DONE=1; fi
    ((RUN_DONE)) || run_record "exit $rc"
  fi
  if ((LEASE_OWN)); then lease_release_own || { ((rc != 0)) || rc=4; }; fi
  if [[ -n $KEY_FILE && -f $KEY_FILE.pub ]]; then
    run gcloud compute os-login ssh-keys remove --key-file "$KEY_FILE.pub" "${GFLAGS[@]}" >/dev/null 2>&1 || log "the OS Login key could not be removed (TTL expires it)"
  fi
  [[ -z $WORKTREE ]] || git -C "$REPO" worktree remove --force "$WORKTREE" >/dev/null 2>&1 || rm -rf "$WORKTREE"
  CF_TOKEN=""
  unset CLOUDFLARE_API_TOKEN
  if [[ -n $OUT && -d $OUT ]]; then summary_publish "$rc" || rc=6; fi
  if [[ -n $PRIV && -d $PRIV ]]; then
    rm -rf "$RAW" "$PRIV/private/ci" "$PRIV/gcloud-adc.json" "$CANARY_FILE" "$PRIV/home/.ssh"
    for r in "$RAW" "$PRIV/private/ci" "$PRIV/gcloud-adc.json"; do [[ ! -e $r ]] || { printf 'ci: %s could not be removed\n' "$r" >&3; rc=4; }; done
  fi
  meta_cleanup
  exit "$rc"
}
