#!/usr/bin/env bash
# Operator lifecycle of one on-demand SaaStacular staging environment on a GCP VM (docs/foundation/STAGING.md).
#
#   staging.sh build     --config FILE --candidate SHA [--source DIR] [--dry-run]
#   staging.sh push      --config FILE --candidate SHA [--dry-run]
#   staging.sh provision --config FILE --candidate SHA [--duration D] [--override VAR=VALUE]... [--yes] [--dry-run]
#   staging.sh deploy    --config FILE --candidate SHA [--pin-secret NAME=VERSION|latest|none]... [--dry-run]
#   staging.sh start     --config FILE [--duration D] [--dry-run]
#   staging.sh stop      --config FILE [--dry-run]
#   staging.sh reset     --config FILE --confirm ENV [--dry-run]
#   staging.sh retire    --config FILE [--yes] [--purge-images] [--dry-run]
#   staging.sh status    --config FILE [--dry-run]
#   staging.sh receipt   --config FILE [--candidate SHA] [--dry-run]
#   staging.sh lease     --config FILE --show|--release|--break --confirm ENV --reason TEXT [--dry-run]
#
# build/push: images of the exact candidate (a clean detached checkout), built with the existing Compose stanzas and
#   the edge Dockerfile one at a time, tagged <AR_REPOSITORY>/<service>:<sha>; push adds their registry digests to the
#   deploy manifest EVIDENCE_DIR/deploy/<sha>.tsv (no secret: candidate, Dockerfile hashes, image IDs, digests).
#   build --source DIR (G072, additive): the Compose files, build contexts, edge context and git checks come from the
#   clean detached checkout DIR at the candidate instead of this checkout (the CI build runs this trusted control copy
#   of staging.sh against a candidate worktree); without it, this checkout as before.
# Environment lease (G072 Plan §10.5, additive; meta-lib.sh): when gs://<GCP_PROJECT>-<STAGING_ENV>-ci-meta exists,
#   provision, deploy, start, stop, reset and retire first create its lease object (refused, exit 2, while another
#   holder has it) and release it at exit; deploy (and provision's deploy) then appends the candidate's schema-fence
#   entry before the release upload. The bucket positively absent (404): exactly the behaviour above. Any other answer
#   to that check, or a bucket whose owner project number is not GCP_PROJECT's: refused, exit 2. The expected number is
#   STAGING_META_PROJECT_NUMBER when set (G072 CI: from the profile's WIF provider), else gcloud projects describe as
#   the deployer; when it cannot be read the command is refused (set STAGING_META_PROJECT_NUMBER to proceed). STAGING_LEASE_HELD=<holder>@<generation> runs a command under a lease its caller
#   holds (G072 CI). lease --show prints the lease; --release deletes this operator's own local lease; --break, the
#   environment owner's break-glass, writes a lease-breaks/ record (operator, reason, the broken lease) and deletes it.
# provision: terraform init/plan (saved privately, always shown, refused when it touches anything outside the
#   environment), apply only with --yes, then deploy when the VM runs. Re-running converges; a failed apply keeps its
#   partial state and the next provision completes it. --override passes one root variable (-var), visible here and in
#   the saved plan (e.g. the deliberate first-allocation failure proof); never session_max_seconds (--duration).
#   Session length: on first allocation --duration D or the configured default; an existing VM keeps the length it
#   reports (every plan passes it; Terraform ignores it there in any case) and refuses --duration (use start
#   --duration). The VM counts as absent only when its describe fails and an exact-name instance list positively
#   returns nothing; provision, deploy, start, stop, reset, retire and status HOLD (exit 3) on a VM whose state
#   cannot be read, before any plan, apply, deploy, start or stop, and the receipt records it as unreadable, never absent. An existing VM whose session length is unreadable,
#   missing or malformed also HOLDs (exit 3); only a confirmed-absent VM takes the configured default.
# deploy: uploads the candidate's infra subtree and the manifest, then remote/deploy.sh (secrets, pull, stores,
#   migrations, seed only on a fresh dataset, app and edge, readiness). --pin-secret selects Secret Manager versions.
#   Never Terraform, a VM start or stop, or a scheduling change: a running session's timer is untouched.
# start/stop: gcloud start/stop of the VM around remote/start.sh (holds app and edge until the recovery checks pass)
#   and remote/stop.sh (compose stop). No instance is made or deleted outside Terraform; no volume-removing down; no migrate down or reapply.
#   start selects the session length (--duration D, or the configured default: never a previous session's); a stopped
#   (TERMINATED) VM whose length differs gets it with gcloud compute instances set-scheduling (every other scheduling
#   attribute stated as provisioned and read back), then one saved refresh-only Terraform plan whose only drift must be
#   that length, its apply, and both sources re-read; only then does it start. Terraform ignores an existing VM's length
#   (the provider would replace the VM), so -var session_max_seconds never changes one. A stopped VM that already
#   reports the requested length while Terraform differs (a change whose refresh failed) gets only the refresh steps.
#   D = <n>m, <n>h or whole seconds, from 600 (10m) to 28800 (8h); only provision and start take it.
# reset: the only data-discarding path (exact dataset volumes of this VM's project; edge-data kept).
# retire: stop, terraform plan -destroy (shown, saved, checked), destroy with --yes, absence verification, receipt.
# status/receipt: read-only state, reconciliation and the all-category receipt (EVIDENCE_DIR/receipts/).
# --dry-run prints every command (and the VM-side commands) without running anything.
#
# Exit: 0 success; 2 refused before any change; 3 HOLD (VM side: app and edge stopped; operator side: the VM's state or
# session length cannot be read or is missing or malformed, the instance and Terraform disagree on the session length,
# or a session change failed its readback, refresh, scope check, apply or verification: the VM is not started and the
# partial state and recovery are printed); 4 failed after a change started.
set -euo pipefail
umask 077
# shellcheck source=lib.sh
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"
# shellcheck source=meta-lib.sh
source "$(dirname "${BASH_SOURCE[0]}")/meta-lib.sh"
REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)"

declare -A OPT=()
# The Dockerfile (infra/docker/<name>.Dockerfile) of each built service.
declare -A DOCKERFILE=([core]=core [web]=web [intelligence]=intelligence [minio]=minio [minio-init]=minio [migrate]=migrate
  [edge]=edge)
OVERRIDES=()
PINS=()
TMP=""

# The lease this process acquired is released here (INT/TERM/HUP exit through it); a failed release is logged.
cleanup() { lease_release_own || true; meta_cleanup; [[ -z $TMP ]] || rm -rf "$TMP"; }
trap cleanup EXIT
trap 'exit 130' INT TERM HUP

parse() {
  local allowed=$1 ov
  shift
  while (($#)); do
    [[ " $allowed " == *" ${1%%=*} "* || $1 == --dry-run ]] || die 2 "$COMMAND does not take $1"
    case $1 in
      --dry-run) DRY_RUN=1 ;;
      --yes | --purge-images | --show | --release | --break) OPT[${1#--}]=1 ;;
      --config | --candidate | --confirm | --reason)
        (($# >= 2)) || die 2 "$1 needs a value"
        OPT[${1#--}]=$2
        shift
        ;;
      --source)
        (($# >= 2)) || die 2 "--source needs a directory"
        [[ $2 == /* && $2 != *[[:space:]]* && -d $2 ]] || die 2 "--source takes an existing absolute directory without spaces"
        OPT[source]=$2
        shift
        ;;
      --duration)
        (($# >= 2)) || die 2 "--duration needs a value"
        [[ -z ${OPT[duration]:-} ]] || die 2 "--duration is given twice"
        OPT[duration]="$(duration_seconds "$2")" ||
          die 2 "--duration takes <n>m, <n>h or whole seconds from $SESSION_FLOOR (10m) to $SESSION_CEILING (8h), not $2"
        shift
        ;;
      --override)
        (($# >= 2)) || die 2 "--override needs VAR=VALUE"
        [[ $2 =~ ^([a-z][a-z0-9_]*)=([A-Za-z0-9._:/@-]*)$ ]] || die 2 "--override takes VAR=VALUE (lowercase variable, plain value)"
        ov=${BASH_REMATCH[1]} # a second =~ resets BASH_REMATCH
        [[ ! $ov =~ (token|secret|password|credential|key) ]] || die 2 "--override never carries a secret-shaped variable"
        [[ $ov != session_max_seconds ]] ||
          die 2 "--override session_max_seconds is not accepted: the session length is --duration (provision on first allocation, otherwise start --duration)"
        OVERRIDES+=(-var "$2")
        shift
        ;;
      --pin-secret)
        (($# >= 2)) || die 2 "--pin-secret needs NAME=VERSION"
        [[ $2 =~ ^[A-Z][A-Z0-9_]*=([0-9]+|latest|none)$ ]] || die 2 "--pin-secret takes NAME=VERSION|latest|none"
        PINS+=(--pin-secret "$2")
        shift
        ;;
      *) die 2 "unknown argument $1" ;;
    esac
    shift
  done
  [[ -n ${OPT[config]:-} ]] || die 2 "--config FILE is required"
  load_conf "${OPT[config]}"
}

need_candidate() { [[ -n ${OPT[candidate]:-} ]] || die 2 "--candidate SHA is required"; check_candidate "${OPT[candidate]}"; }

# ---------------------------------------------------------------------------------------------------------------- build
cmd_build() {
  local c=${OPT[candidate]} ar=${CONF[AR_REPOSITORY]} src=${OPT[source]:-$REPO} m dummy svc id df
  if ((DRY_RUN)); then
    show git -C "$src" rev-parse HEAD
    show git -C "$src" status --porcelain --untracked-files=no
  else
    [[ "$(git -C "$src" rev-parse HEAD)" == "$c" ]] || die 2 "check out the candidate first (detached HEAD at $c)"
    [[ -z "$(git -C "$src" status --porcelain --untracked-files=no)" ]] || die 2 "the checkout has local changes"
  fi
  private_dirs
  m="$(manifest_path "$c")"
  dummy="${CONF[EVIDENCE_DIR]}/tmp/build.env"
  if ((!DRY_RUN)); then
    TMP="$(mktemp -d "${CONF[EVIDENCE_DIR]}/tmp/build.XXXXXX")"
    dummy=$TMP/build.env
    # Compose needs every required variable to parse; builds use none of them. Placeholders, not secrets.
    { git -C "$src" show "$c:.env.example" | grep -v '^COMPOSE_PROJECT_NAME=' | sed -E 's/^([A-Z0-9_]+)=$/\1=build-placeholder/'
      printf 'COMPOSE_PROJECT_NAME=%s-build\n' "$ENV_NAME"; } >"$dummy"
  fi
  run env COMPOSE_PARALLEL_LIMIT=1 docker compose --env-file "$dummy" -f "$src/infra/compose/local.yml" \
    -f "$src/infra/compose/migrate.yml" --profile migrate build core web intelligence minio minio-init migrate
  for svc in core web intelligence minio minio-init migrate; do
    run docker tag "$ENV_NAME-build-$svc" "$ar/$svc:$c"
  done
  run docker build --file "$src/infra/docker/edge.Dockerfile" --tag "$ar/edge:$c" "$src/infra/docker"
  if ((DRY_RUN)); then printf '+ (write %s: candidate, Dockerfile sha256, image IDs)\n' "$m"; return 0; fi
  {
    printf 'candidate\t%s\n' "$c"
    for svc in "${BUILT_SERVICES[@]}"; do
      df=${DOCKERFILE[$svc]}
      id="$(docker image inspect --format '{{.Id}}' "$ar/$svc:$c")"
      printf 'image\t%s\t%s\t%s\t%s\t-\n' "$svc" "$c" "$(git -C "$src" show "$c:infra/docker/$df.Dockerfile" | sha256sum | cut -d' ' -f1)" "$id"
      inventory docker-image "$ar/$svc:$c $id" built "task-owned; removed at closeout"
    done
    for df in infra/compose/local.yml infra/compose/migrate.yml infra/compose/staging.yml infra/edge/Caddyfile infra/staging/seed.sql; do
      printf 'file\t%s\t%s\n' "$df" "$(git -C "$src" show "$c:$df" | sha256sum | cut -d' ' -f1)"
    done
  } >"$m"
  log "manifest $m (local image IDs; push adds the registry digests)"
}

# ----------------------------------------------------------------------------------------------------------------- push
cmd_push() {
  local c=${OPT[candidate]} ar=${CONF[AR_REPOSITORY]} m svc want have ref
  m="$(manifest_path "$c")"
  ((DRY_RUN)) || [[ -f $m ]] || die 2 "no manifest for $c: run build first"
  private_dirs
  if ((DRY_RUN)); then
    printf '+ gcloud auth print-access-token --impersonate-service-account %q | docker --config <private tmp> login --username oauth2accesstoken --password-stdin %q\n' \
      "${CONF[DEPLOYER_SA]}" "https://${ar%%/*}"
    for svc in "${BUILT_SERVICES[@]}"; do show docker --config '<private tmp>' push "$ar/$svc:$c"; done
    printf '+ (record each RepoDigest in %s)\n' "$m"
    return 0
  fi
  TMP="$(mktemp -d "${CONF[EVIDENCE_DIR]}/tmp/push.XXXXXX")"
  show gcloud auth print-access-token --impersonate-service-account "${CONF[DEPLOYER_SA]}"
  gcloud auth print-access-token --impersonate-service-account "${CONF[DEPLOYER_SA]}" |
    docker --config "$TMP" login --username oauth2accesstoken --password-stdin "https://${ar%%/*}" >/dev/null
  cp "$m" "$TMP/manifest.tsv"
  for svc in "${BUILT_SERVICES[@]}"; do
    want="$(awk -F'\t' -v s="$svc" '$1 == "image" && $2 == s {print $5}' "$m")"
    have="$(docker image inspect --format '{{.Id}}' "$ar/$svc:$c")"
    [[ -n $want && $want == "$have" ]] || die 2 "$ar/$svc:$c is not the image build recorded"
    run docker --config "$TMP" push "$ar/$svc:$c"
    ref="$(docker image inspect --format '{{range .RepoDigests}}{{println .}}{{end}}' "$ar/$svc:$c" | grep -m1 "^$ar/$svc@sha256:")" ||
      die 4 "no registry digest for $svc after the push"
    awk -F'\t' -v OFS='\t' -v s="$svc" -v r="$ref" '$1 == "image" && $2 == s {$6 = r} {print}' "$TMP/manifest.tsv" >"$TMP/manifest.new"
    mv "$TMP/manifest.new" "$TMP/manifest.tsv"
    inventory registry-image "$ref" pushed "retained in the repository until retire --purge-images or bootstrap teardown"
  done
  install -m 0600 "$TMP/manifest.tsv" "$m"
  log "manifest $m now holds the registry digests"
}

# --------------------------------------------------------------------------------------------------- provision / deploy
record_state() {
  local n
  ((DRY_RUN)) && return 0
  while IFS= read -r n; do [[ -z $n ]] || inventory terraform-resource "$n" applied "retained baseline until retire"; done < <(state_names || true)
}

cmd_provision() {
  local c=${OPT[candidate]} plan rc want
  need_cloudflare_token
  private_dirs
  tf_init
  reconcile
  vm_status
  # A dry run reads nothing: with --duration it shows the first allocation, otherwise the existing-VM path.
  ((!DRY_RUN)) || [[ -z ${OPT[duration]:-} ]] || VMSTATE=ABSENT
  if [[ $VMSTATE == ABSENT ]]; then
    want=${OPT[duration]:-${CONF[SESSION_MAX_SECONDS]}} # first allocation: the VM is created running, armed with want
  else
    # Infrastructure maintenance (partial-allocation repair included) keeps the length the instance reports.
    [[ -z ${OPT[duration]:-} ]] ||
      die 2 "$VM exists: provision keeps its session length; use start --duration (stop the VM first if it runs)"
    vm_session || die 3 "HOLD: the session length of $VM could not be read; nothing planned"
    ((DRY_RUN)) || [[ $SESSION_HAVE =~ ^[1-9][0-9]*$ ]] || die 3 "HOLD: $VM reports no max run duration; nothing planned"
    want=$SESSION_HAVE
  fi
  plan="/private/plans/$(date -u +%Y%m%dT%H%M%SZ)-apply.tfplan"
  rc=0
  tf plan -input=false -detailed-exitcode -var-file=/private/terraform.tfvars -var "session_max_seconds=$want" "${OVERRIDES[@]}" \
    -out="$plan" || rc=$?
  ((!DRY_RUN)) || rc=2
  case $rc in
    0) log "terraform: no changes" ;;
    2)
      tf show -no-color "$plan"
      label_check "$plan"
      if [[ -z ${OPT[yes]:-} ]]; then
        log "plan saved as ${CONF[PRIVATE_DIR]}/plans/${plan##*/}; re-run with --yes to apply"
        ((DRY_RUN)) || return 0
      fi
      rc=0
      tf apply -input=false "$plan" || rc=$?
      record_state
      if ((rc)); then
        (reconcile) || true
        die 4 "terraform apply failed; the partial state is recorded (re-run provision to converge)"
      fi
      ;;
    *) die 2 "terraform plan failed" ;;
  esac
  vm_status
  case $VMSTATE in
    RUNNING)
      wait_vm_ready
      deploy_candidate "$c"
      ;;
    TERMINATED) log "$VM is stopped: infrastructure applied; run start" ;;
    *) die 4 "$VM is $VMSTATE after apply" ;;
  esac
}

# deploy_candidate SHA: upload release + meta, then remote/deploy.sh from the uploaded release.
deploy_candidate() {
  local c=$1 m sha12=${1:0:12} incoming rc=0 logf
  incoming=/srv/$ENV_NAME/incoming/$sha12
  m="$(manifest_path "$c")"
  require_pushed "$m" "$c"
  vm_status
  [[ $VMSTATE == RUNNING ]] || die 2 "$VM is $VMSTATE: start it first"
  # G072 schema fence (Plan §10.4): before the upload (the VM's promote and migrate follow it), under the lease, when
  # the environment's metadata bucket exists; a failed fence write uploads nothing.
  [[ $META_STATE != present && $META_STATE != dry-run ]] || fence_append "$c"
  if ((DRY_RUN)); then
    printf '+ git -C %q archive --format=tar %s infra/compose infra/postgres infra/minio infra/edge infra/staging | tar -x -C <private tmp>/release\n' "$REPO" "$c"
    printf '+ (write <private tmp>/meta/remote.env from staging.conf; copy %s to <private tmp>/meta/manifest.tsv)\n' "$m"
    printf '+ tar -C <private tmp> -cf - release meta   (piped into the next command)\n'
    vm_ssh "sudo rm -rf $incoming && sudo install -d -m 0750 /srv/$ENV_NAME /srv/$ENV_NAME/incoming $incoming && sudo tar -C $incoming --no-same-owner -xf -"
    vm_ssh "sudo bash $incoming/release/infra/staging/remote/deploy.sh --env $ENV_NAME --incoming $incoming ${PINS[*]}"
    bash "$REPO/infra/staging/remote/deploy.sh" --env "$ENV_NAME" --incoming "$incoming" "${PINS[@]}" --dry-run
    inventory vm-release "$incoming" uploaded "replaced by the next deploy"
    return 0
  fi
  TMP="$(mktemp -d "${CONF[EVIDENCE_DIR]}/tmp/upload.XXXXXX")"
  mkdir "$TMP/release" "$TMP/meta"
  # The release holds only repository files: extract with a normal umask, not the private 077, so a bind-mounted file
  # (minio's init.sh) stays readable by a non-root container user; /srv/<env> itself is 0750 root.
  git -C "$REPO" archive --format=tar "$c" infra/compose infra/postgres infra/minio infra/edge infra/staging | (umask 022 && tar -x -C "$TMP/release")
  cp "$m" "$TMP/meta/manifest.tsv"
  remote_env >"$TMP/meta/remote.env"
  tar -C "$TMP" -cf - release meta |
    vm_ssh "sudo rm -rf $incoming && sudo install -d -m 0750 /srv/$ENV_NAME /srv/$ENV_NAME/incoming $incoming && sudo tar -C $incoming --no-same-owner -xf -"
  rm -rf "$TMP"
  TMP=""
  inventory vm-release "/srv/$ENV_NAME/releases/$sha12" uploaded "pruned after the next deploy; destroyed with the VM"
  logf="${CONF[EVIDENCE_DIR]}/logs/$(date -u +%Y%m%dT%H%M%SZ)-deploy-$sha12.log"
  vm_ssh "sudo bash $incoming/release/infra/staging/remote/deploy.sh --env $ENV_NAME --incoming $incoming ${PINS[*]}" | tee "$logf" || rc=$?
  log "VM output kept in $logf"
  inventory compose-project "$ENV_NAME" "deployed $c (exit $rc)" "retained baseline until retire"
  finish "$rc"
}

cmd_deploy() { private_dirs; deploy_candidate "${OPT[candidate]}"; }

# finish RC: report a VM-side outcome and exit with it.
finish() {
  case $1 in
    0) return 0 ;;
    3) die 3 "HOLD on the VM: app and edge are stopped; nothing was reseeded (see the log; escalate)" ;;
    *) die "$1" "the VM-side step failed (exit $1; see the log)" ;;
  esac
}

# --------------------------------------------------------------------------------------------------------- start / stop
check_address() {
  local reserved attached
  if ((DRY_RUN)); then
    show gcloud compute addresses describe "$ENV_NAME-ip" --region "${CONF[GCP_REGION]}" "${GFLAGS[@]}" --format 'value(address)'
    show gcloud compute instances describe "$VM" --zone "${CONF[GCP_ZONE]}" "${GFLAGS[@]}" --format 'value(networkInterfaces[0].accessConfigs[0].natIP)'
    return 0
  fi
  reserved="$(gcloud compute addresses describe "$ENV_NAME-ip" --region "${CONF[GCP_REGION]}" "${GFLAGS[@]}" --format 'value(address)')"
  attached="$(gcloud compute instances describe "$VM" --zone "${CONF[GCP_ZONE]}" "${GFLAGS[@]}" --format 'value(networkInterfaces[0].accessConfigs[0].natIP)')"
  [[ -n $reserved && $reserved == "$attached" ]] || die 4 "the VM's external address ($attached) is not the reserved $ENV_NAME-ip ($reserved)"
  printf 'address\t%s\n' "$reserved"
}

expiry() { # the instance's scheduling as Compute Engine reports it (max run duration, termination time when armed)
  run gcloud compute instances describe "$VM" --zone "${CONF[GCP_ZONE]}" "${GFLAGS[@]}" \
    --format 'yaml(status,scheduling,resourceStatus.scheduling)'
}

# set_session WANT: the stopped VM's session length from SESSION_HAVE (which Terraform records too) to WANT (G070 A2
# fallback: hashicorp/google replaces the VM for this field, so the module ignores it after creation). Preconditions:
# the Cloudflare token (the refresh plans the whole root) and TERMINATED read just now. gcloud set-scheduling sets the
# length and states every other provisioned scheduling attribute again (SESSION_SCHED_FLAGS); then session_refresh.
# Never runs on a running VM; never starts the VM after a failure.
set_session() {
  local want=$1 from=$SESSION_HAVE
  need_cloudflare_token
  vm_fresh_terminated
  run gcloud compute instances set-scheduling "$VM" --zone "${CONF[GCP_ZONE]}" "${GFLAGS[@]}" "--max-run-duration=${want}s" \
    "${SESSION_SCHED_FLAGS[@]}" || session_hold "gcloud compute instances set-scheduling failed" "$want"
  inventory vm "$VM" "session length ${from}s -> ${want}s set with gcloud set-scheduling while stopped" "retained baseline until retire"
  session_refresh "$want" "$from"
}

# session_refresh WANT FROM: after the instance's length became WANT (set_session, or found so by start's recovery):
# the full scheduling readback, then one saved refresh-only plan of the whole root (exit 2 expected: the length drift;
# 0 or any other exit HOLDs), label_check, refresh_scope_check (only the VM's length drifts, FROM -> WANT), the apply
# of exactly that plan, and both sources re-read: the instance and Terraform must report WANT. Every failure is a
# HOLD (session_hold) with the VM stopped. The only Terraform write of an existing VM's length; never a normal plan.
session_refresh() {
  local want=$1 from=$2 plan rc=0 rec
  scheduling_check "$want"
  plan="/private/plans/$(date -u +%Y%m%dT%H%M%SZ)-refresh.tfplan"
  tf plan -refresh-only -input=false -detailed-exitcode -var-file=/private/terraform.tfvars -var "session_max_seconds=$want" \
    -out="$plan" || rc=$?
  ((!DRY_RUN)) || rc=2
  case $rc in
    2) ;;
    0) session_hold "the refresh-only plan shows no drift although $VM reports ${want}s and Terraform ${from}s" "$want" ;;
    *) session_hold "terraform plan -refresh-only failed (exit $rc)" "$want" ;;
  esac
  tf show -no-color "$plan" || session_hold "terraform show of the refresh-only plan failed" "$want"
  (label_check "$plan") || session_hold "the label check refused the refresh-only plan" "$want"
  refresh_scope_check "$plan" "$from" "$want"
  rc=0
  tf apply -input=false "$plan" || rc=$?
  record_state
  ((rc == 0)) || session_hold "terraform apply of the refresh-only plan failed (exit $rc)" "$want"
  # Both sources again, after the apply: the instance and Terraform's output must report WANT.
  vm_session || session_hold "$VM could not be read again after the refresh" "$want"
  if ((DRY_RUN)); then
    tf output -raw max_run_duration_seconds
  else
    rec="$(tf output -raw max_run_duration_seconds)" ||
      session_hold "Terraform's max_run_duration_seconds output could not be read after the refresh" "$want"
    [[ $SESSION_HAVE == "$want" && $rec == "$want" ]] ||
      session_hold "after the refresh the instance reports [${SESSION_HAVE}]s and Terraform [${rec}]s, not ${want}s" "$want"
  fi
  inventory vm "$VM" "session length ${want}s recorded by a refresh-only apply; instance and Terraform agree" "retained baseline until retire"
}

cmd_start() {
  local rc=0 want=${OPT[duration]:-${CONF[SESSION_MAX_SECONDS]}} recover=0
  private_dirs
  tf_init
  reconcile
  vm_status
  case $VMSTATE in
    TERMINATED | STOPPED | RUNNING) ;;
    *) die 2 "$VM is $VMSTATE" ;;
  esac
  session_pair
  if ((DRY_RUN)); then
    # Nothing was read: with --duration a dry run shows the stopped-VM change, otherwise the resume path.
    if [[ -n ${OPT[duration]:-} ]]; then VMSTATE=TERMINATED SESSION_HAVE='<observed>'; else SESSION_HAVE=$want; fi
    SESSION_TF=$SESSION_HAVE
  fi
  # Instance and Terraform disagree: the one recoverable partial state is a stopped VM that already reports the
  # requested length (an earlier set-scheduling whose refresh did not complete); anything else HOLDs.
  if [[ $SESSION_HAVE != "$SESSION_TF" ]]; then
    [[ $VMSTATE == TERMINATED && $SESSION_HAVE == "$want" ]] || session_mismatch
    recover=1
  fi
  case $VMSTATE in
    RUNNING)
      [[ $want == "$SESSION_HAVE" ]] ||
        die 2 "$VM runs a ${SESSION_HAVE}s session; nothing changed (no stop is issued here): stop it first, then start --duration; --duration ${SESSION_HAVE} runs the resume checks on this session"
      log "$VM already runs its ${SESSION_HAVE}s session (scheduled stop ${SESSION_STOP:-not armed}; restart policy may have resumed containers): running the explicit resume checks"
      ;;
    *)
      if ((recover)); then
        log "recovery: $VM already reports ${want}s while Terraform records ${SESSION_TF}s: refreshing Terraform, then verifying both before the start"
        need_cloudflare_token
        vm_fresh_terminated
        session_refresh "$want" "$SESSION_TF"
      elif [[ $want != "$SESSION_HAVE" ]]; then
        [[ $VMSTATE == TERMINATED ]] ||
          die 3 "HOLD: $VM is $VMSTATE, not TERMINATED: its session length is changed only while TERMINATED; nothing changed or started"
        set_session "$want"
      fi
      run gcloud compute instances start "$VM" --zone "${CONF[GCP_ZONE]}" "${GFLAGS[@]}"
      wait_status RUNNING 300
      ;;
  esac
  check_address
  wait_vm_ready
  remote start.sh || rc=$?
  expiry
  vm_session || SESSION_STOP=""
  ((DRY_RUN)) || printf 'session\t%ss\tscheduled stop %s\n' "$want" "${SESSION_STOP:-not reported}"
  inventory vm "$VM" "started (VM-side exit $rc); session ${want}s, scheduled stop ${SESSION_STOP:-not reported}" "retained baseline until retire"
  finish "$rc"
}

# stop_vm: VM-side compose stop (best effort: the VM stop proceeds either way), then the VM stop.
stop_vm() {
  remote status.sh || log "VM status could not be read before the stop"
  remote stop.sh || log "the VM-side stop failed; stopping the VM anyway (start holds app and edge until its checks pass)"
  run gcloud compute instances stop "$VM" --zone "${CONF[GCP_ZONE]}" "${GFLAGS[@]}"
  wait_status TERMINATED 600
  inventory vm "$VM" stopped "retained baseline until retire"
}

cmd_stop() {
  private_dirs
  vm_status
  case $VMSTATE in
    RUNNING) stop_vm ;;
    TERMINATED | STOPPED) log "$VM is already stopped" ;;
    *) die 2 "$VM is $VMSTATE" ;;
  esac
  write_receipt stop
}

cmd_reset() {
  local rc=0
  [[ ${OPT[confirm]:-} == "$ENV_NAME" ]] || die 2 "reset discards the dataset: pass --confirm $ENV_NAME"
  private_dirs
  vm_status
  [[ $VMSTATE == RUNNING ]] || die 2 "$VM is $VMSTATE: reset needs the running VM"
  remote reset.sh --confirm "$ENV_NAME" || rc=$?
  inventory docker-volume "${ENV_NAME}_postgres-data ${ENV_NAME}_minio-data" "removed, then initialized again (exit $rc)" "retained baseline until retire"
  finish "$rc"
}

# ------------------------------------------------------------------------------------------------------------- retire
# cf_records NAME: the number of Cloudflare DNS records named NAME in the zone (token from the environment, via a
# curl config on a pipe: never on a command line).
cf_records() {
  curl -fsS --max-time 20 --config <(printf 'header = "Authorization: Bearer %s"\n' "$CLOUDFLARE_API_TOKEN") \
    "https://api.cloudflare.com/client/v4/zones/${CONF[CLOUDFLARE_ZONE_ID]}/dns_records?name=${1/\*/%2A}" | jq -r '.result | length'
}

verify_absent() {
  local b=${CONF[BASE_DOMAIN]} n names live count bad=""
  if ((DRY_RUN)); then
    tf show -json
    printf '+ (the label/name queries of reconcile return nothing)\n'
    for n in "$b" "*.$b" "custom.$b"; do
      printf '+ curl --config <token header on a pipe> %q | jq -r .result|length\n' "https://api.cloudflare.com/client/v4/zones/${CONF[CLOUDFLARE_ZONE_ID]}/dns_records?name=${n/\*/%2A}"
    done
    return 0
  fi
  names="$(state_names)" || die 4 "the Terraform state could not be read"
  [[ -z $names ]] || bad+="state: ${names//$'\n'/ }; "
  live="$(live_names)" || die 4 "the Compute Engine label query failed"
  [[ -z $live ]] || bad+="live: ${live//$'\n'/ }; "
  for n in "$b" "*.$b" "custom.$b"; do
    count="$(cf_records "$n")" || die 4 "the Cloudflare read failed"
    [[ $count == 0 ]] || bad+="dns: $n ($count); "
  done
  [[ -z $bad ]] || die 4 "retirement left objects: $bad"
  printf 'absent\tstate=empty compute=empty dns=%s,*.%s,custom.%s\n' "$b" "$b" "$b"
}

purge_images() {
  local f ref
  for f in "${CONF[EVIDENCE_DIR]}"/deploy/*.tsv; do
    [[ -f $f ]] || continue
    while IFS= read -r ref; do
      [[ $ref == "${CONF[AR_REPOSITORY]}/"*@sha256:* ]] || continue
      run gcloud artifacts docker images delete "$ref" --delete-tags --quiet "${GFLAGS[@]}"
      inventory registry-image "$ref" deleted "absent"
    done < <(awk -F'\t' '$1 == "image" {print $6}' "$f")
  done
}

cmd_retire() {
  local plan rc=0
  need_cloudflare_token
  private_dirs
  vm_status
  [[ $VMSTATE != RUNNING ]] || stop_vm
  tf_init
  session_var
  plan="/private/plans/$(date -u +%Y%m%dT%H%M%SZ)-destroy.tfplan"
  tf plan -destroy -input=false -var-file=/private/terraform.tfvars "${SESSION_VAR[@]}" -out="$plan"
  tf show -no-color "$plan"
  label_check "$plan"
  if [[ -z ${OPT[yes]:-} ]]; then
    log "destroy plan saved as ${CONF[PRIVATE_DIR]}/plans/${plan##*/}; re-run with --yes to retire"
    ((DRY_RUN)) || return 0
  fi
  tf apply -input=false "$plan" || rc=$?
  ((rc == 0)) || { (reconcile) || true; die 4 "terraform destroy failed; re-run retire to converge"; }
  inventory environment "$ENV_NAME" "retired (VM, disks, address, firewall, network, DNS records destroyed)" absent
  verify_absent
  [[ -z ${OPT[purge-images]:-} ]] || purge_images
  write_receipt retire
}

# ---------------------------------------------------------------------------------------------------- status / receipt
dns_answers() {
  local b=${CONF[BASE_DOMAIN]} n
  for n in "$b" "probe-$RANDOM.$b" "custom.$b" "auth.$b"; do
    if ((DRY_RUN)); then show getent ahostsv4 "$n"; continue; fi
    printf 'dns\t%s\t%s\n' "$n" "$(getent ahostsv4 "$n" | awk '{print $1}' | sort -u | tr '\n' ' ')"
  done
}

cmd_status() {
  local rc=0
  need_cloudflare_token
  private_dirs
  tf_init
  vm_status
  session_var
  tf plan -refresh-only -input=false -lock=false -detailed-exitcode -var-file=/private/terraform.tfvars "${SESSION_VAR[@]}" || rc=$?
  if ((!DRY_RUN)); then
    case $rc in
      0) printf 'terraform\tno drift\n' ;;
      2) printf 'terraform\tdrift (refresh-only changes above)\n' ;;
      *) die 2 "terraform plan -refresh-only failed" ;;
    esac
  fi
  reconcile
  printf 'vm\t%s\t%s\n' "$VM" "$VMSTATE"
  expiry
  if ((!DRY_RUN)); then
    if [[ $VMSTATE != ABSENT ]] && vm_session; then printf 'session\t%s\n' "$(session_detail)"; else printf 'session\tunavailable\n'; fi
  fi
  dns_answers
  if [[ $VMSTATE == RUNNING ]]; then
    check_address
    remote status.sh || log "VM status could not be read"
    printf 'cost\trunning: VM, 2 disks, in-use address (plus the stopped items)\n'
  else
    printf 'cost\tstopped: 2 disks, idle reserved address, secrets, registry storage, state bucket\n'
  fi
}

# rline CATEGORY ITEM DETAIL OWNER REASON RELEASE: one receipt row.
rline() { printf '%s\t%s\t%s\t%s\t%s\t%s\n' "$@" >>"$RECEIPT"; }
RECEIPT=""

# write_receipt KIND: the all-category receipt (stop | retire | manual): EVIDENCE_DIR/receipts/<env>-<sha12>-<n>.tsv,
# its ID and sha256 printed. Read-only queries; Terraform-owned objects through the deployer, bootstrap-owned ones
# (secrets, service accounts) with the operator's own gcloud login (a refusal is recorded as unavailable).
write_receipt() {
  local kind=$1 c=${OPT[candidate]:-} last n id owner=${CONF[OWNER]} retained base row f ref secrets
  # No logs directory yet (first run, or --dry-run, which creates nothing) means no status log, not a failure.
  last="$({ find "${CONF[EVIDENCE_DIR]}/logs" -name '*-status.log' 2>/dev/null || true; } | LC_ALL=C sort | tail -n 1)"
  [[ -n $c || -z $last ]] || c="$(awk -F'\t' '$1 == "candidate" {print $2}' "$last")"
  [[ $c =~ ^[0-9a-f]{40}$ ]] || c=000000000000
  if ((DRY_RUN)); then
    printf '+ (write %s/receipts/%s-%s-<n>.tsv: VM, disks, address, DNS records, edge-data/edge-config, Terraform state,\n' "${CONF[EVIDENCE_DIR]}" "$ENV_NAME" "${c:0:12}"
    printf '+  secrets with pinned versions, registry digests, service accounts, local Docker objects, worktrees, private paths)\n'
    show gcloud compute instances describe "$VM" --zone "${CONF[GCP_ZONE]}" "${GFLAGS[@]}" --format 'value(name,status,machineType.basename())'
    vm_session
    show gcloud compute disks list "${GFLAGS[@]}" --filter "labels.environment=$ENV_NAME" --format 'value(name,sizeGb,type.basename())'
    show gcloud compute addresses list "${GFLAGS[@]}" --filter "labels.environment=$ENV_NAME" --format 'value(name,address,status)'
    tf show -json
    env_secrets
    show docker image ls --filter "reference=${CONF[AR_REPOSITORY]}/*" --format '{{.Repository}}:{{.Tag}} {{.ID}}'
    show git -C "$REPO" worktree list --porcelain
    return 0
  fi
  n=1
  while [[ -e ${CONF[EVIDENCE_DIR]}/receipts/$ENV_NAME-${c:0:12}-$n.tsv ]]; do n=$((n + 1)); done
  id=$ENV_NAME-${c:0:12}-$n
  RECEIPT=${CONF[EVIDENCE_DIR]}/receipts/$id.tsv
  retained="reusable stopped baseline (repeat sessions; dependent qualification)"
  base="observed retire"
  [[ $kind != retire ]] || { retained="absent after retire"; base="-"; }
  printf 'receipt\t%s\tkind\t%s\ttime\t%s\tcandidate\t%s\n' "$id" "$kind" "$(ts)" "$c" >"$RECEIPT"
  printf 'category\titem\tdetail\towner\treason\trelease condition\n' >>"$RECEIPT"
  # absent only when vm_read positively confirms it; an unreadable VM is recorded as such, never as absent
  if ! row="$(gcloud compute instances describe "$VM" --zone "${CONF[GCP_ZONE]}" "${GFLAGS[@]}" --format 'value(name,status,machineType.basename())' 2>/dev/null)"; then
    row="unreadable (not confirmed absent)"
    ! vm_read || [[ $VMSTATE != ABSENT ]] || row=absent
  fi
  rline vm "$VM" "${row//$'\t'/ }" "$owner" "$retained" "$base"
  [[ $row == absent ]] || row=unavailable
  if vm_session 2>/dev/null; then row="$(session_detail)"; fi
  rline vm-session "$VM" "$row" operator "native expiry bound" "stop/expiry/retire"
  while IFS=$'\t' read -r f row; do [[ -z $f ]] || rline disk "$f" "$row" "$owner" "$retained" "$base"; done < <(
    gcloud compute disks list "${GFLAGS[@]}" --filter "labels.environment=$ENV_NAME" --format 'value(name,sizeGb,type.basename())' 2>/dev/null | sed 's/\t/ GB /2')
  while IFS=$'\t' read -r f row; do [[ -z $f ]] || rline address "$f" "$row" "$owner" "$retained" "$base"; done < <(
    gcloud compute addresses list "${GFLAGS[@]}" --filter "labels.environment=$ENV_NAME" --format 'value(name,address,status)' 2>/dev/null)
  while IFS=$'\t' read -r f row; do [[ -z $f ]] || rline dns-record "$f" "id $row" "$owner" "$retained" "$base"; done < <(
    tf show -json 2>/dev/null | jq -r '.values.root_module // {} | [recurse(.child_modules[]?) | .resources[]? | select(.type | startswith("cloudflare_")) | [.values.name, .values.id] | @tsv] | .[]' || true)
  if [[ -n $last ]]; then
    rline edge-data "${ENV_NAME}_${CONF[EDGE_DATA_VOLUME]}" "$(awk -F'\t' '$1 == "cert" && $3 ~ /serial|not_after|issuer/ {printf "%s=%s; ", $3, $4}' "$last")" \
      "$owner" "ACME account and certificates (reset never removes them)" "$base"
    while IFS=$'\t' read -r _ f row; do rline vm-volume "$f" "${row:-}" "$owner" "$retained" "$base"; done < <(awk -F'\t' '$1 == "volume"' "$last")
    while IFS=$'\t' read -r _ f row; do rline secret "$ENV_NAME-$f" "pinned version $row" "$owner (bootstrap)" "runtime delivery" "bootstrap teardown"; done < <(awk -F'\t' '$1 == "secret"' "$last")
    while IFS=$'\t' read -r _ f row _; do rline running-digest "$f" "$row" "$owner" "$retained" "$base"; done < <(awk -F'\t' '$1 == "digest"' "$last")
    rline vm-status-source "$last" "the VM-side status this receipt cites" "$owner" "evidence" "closeout"
  fi
  secrets="$(env_secrets 2>/dev/null)" || secrets=""
  [[ -n $secrets ]] || rline secret-versions "-" "unavailable (no read permission for this login, or no secrets)" "$owner (bootstrap)" "runtime delivery" "bootstrap teardown"
  while IFS= read -r f; do
    [[ -n $f ]] || continue
    row="$(gcloud secrets versions list "$f" --project "${CONF[GCP_PROJECT]}" --filter state=ENABLED --format 'value(name)' 2>/dev/null | wc -l)" || row=unavailable
    rline secret-versions "$f" "enabled versions: $row" "$owner (bootstrap)" "runtime delivery" "bootstrap teardown"
  done <<<"$secrets"
  for f in "$ENV_NAME-vm" "$ENV_NAME-deployer"; do
    rline service-account "$f@${CONF[GCP_PROJECT]}.iam.gserviceaccount.com" "bootstrap-owned" "$owner (bootstrap)" "runtime/deployer identity" "bootstrap teardown"
  done
  rline terraform-state "$(grep -E '^[[:space:]]*(bucket|prefix)[[:space:]]*=' "${CONF[PRIVATE_DIR]}/backend.hcl" 2>/dev/null | tr -d ' "' | tr '\n' ' ')" \
    "GCS, versioned" "$owner (bootstrap)" "state history" "no state objects remain, then bootstrap teardown"
  for f in "${CONF[EVIDENCE_DIR]}"/deploy/*.tsv; do
    [[ -f $f ]] || continue
    while IFS=$'\t' read -r _ row _ _ _ ref; do
      [[ $ref == *@sha256:* ]] || continue
      rline registry-image "$ref" "$row" "$owner (bootstrap repository)" "deployable candidates" "retire --purge-images or bootstrap teardown"
    done < <(awk -F'\t' '$1 == "image"' "$f")
  done
  while IFS= read -r row; do [[ -z $row ]] || rline local-image "$row" "operator machine" "$owner" "build/push" "closeout cleanup"; done < <(
    docker image ls --filter "reference=${CONF[AR_REPOSITORY]}/*" --filter "reference=$ENV_NAME-build-*" --format '{{.Repository}}:{{.Tag}} {{.ID}}' 2>/dev/null || true)
  rline worktrees "$(git -C "$REPO" worktree list --porcelain | grep -c '^worktree ')" "git worktree list" "$owner" "operator checkouts" "closeout cleanup"
  for f in "${CONF[PRIVATE_DIR]}" "${CONF[EVIDENCE_DIR]}"; do
    rline private-path "$f" "$(du -sh "$f" 2>/dev/null | cut -f1)" "$owner" "config, plans, manifests, logs, receipts (no secret values)" "closeout"
  done
  chmod 0600 "$RECEIPT"
  printf 'receipt\t%s\t%s\tsha256 %s\n' "$id" "$RECEIPT" "$(sha256sum "$RECEIPT" | cut -d' ' -f1)"
}

cmd_receipt() { private_dirs; write_receipt manual; }

# ---------------------------------------------------------------------------------------------------------------- lease
# cmd_lease: --show | --release (this operator's own local lease: local:<user>@<host>, any process) | --break (the
# environment owner's break-glass: --confirm ENV --reason TEXT, lease-breaks/ record first). Takes no lease itself.
cmd_lease() {
  local k n=0 me f
  for k in show release break; do [[ -z ${OPT[$k]:-} ]] || n=$((n + 1)); done
  ((n == 1)) || die 2 "lease takes exactly one of --show, --release and --break"
  if [[ -n ${OPT[break]:-} ]]; then
    [[ ${OPT[confirm]:-} == "$ENV_NAME" ]] || die 2 "lease --break needs --confirm $ENV_NAME"
    [[ ${OPT[reason]:-} =~ ^[^[:cntrl:]]{3,200}$ ]] || die 2 "lease --break needs --reason TEXT (3 to 200 characters, no control character)"
  else
    [[ -z ${OPT[confirm]:-}${OPT[reason]:-} ]] || die 2 "--confirm and --reason belong to lease --break"
  fi
  meta_presence
  if ((DRY_RUN)); then
    show gcloud storage objects describe "gs://$META_BUCKET/lease" --format 'value(generation)' "${GFLAGS[@]}"
    [[ -z ${OPT[release]:-}${OPT[break]:-} ]] ||
      printf '+ (%sgcloud storage rm gs://%s/lease --if-generation-match=<observed>)\n' \
        "$([[ -n ${OPT[break]:-} ]] && printf 'gcloud storage cp <record> gs://%s/lease-breaks/<utc>-<holder>.tsv --if-generation-match=0, then ' "$META_BUCKET")" "$META_BUCKET"
    return 0
  fi
  [[ $META_STATE == present ]] || die 2 "gs://$META_BUCKET does not exist: $ENV_NAME is not CI-enabled and has no lease"
  lease_read
  if [[ -n ${OPT[show]:-} ]]; then
    if ((META_GEN == 0)); then printf 'lease\tfree\n'; return 0; fi
    printf 'lease\tgeneration\t%s\n' "$META_GEN"
    for f in holder command candidate; do printf 'lease\t%s\t%s\n' "$f" "${LEASE_F[$f]:--}"; done
    printf 'lease\tacquired\t%s\n' "$(meta_iso "${LEASE_F[acquired_epoch]:-0}")"
    if [[ ${LEASE_F[expires_epoch]} == none ]]; then printf 'lease\texpires\tnone (local)\n'
    else printf 'lease\texpires\t%s\n' "$(meta_iso "${LEASE_F[expires_epoch]}")"; fi
    return 0
  fi
  if ((META_GEN == 0)); then log "lease: gs://$META_BUCKET/lease is free; nothing to do"; return 0; fi
  if [[ -n ${OPT[break]:-} ]]; then
    lease_break "$(local_holder)" "${OPT[reason]}"
    return 0
  fi
  me="$(local_holder)"
  me=${me%:*}
  [[ ${LEASE_F[holder]} == "$me" || ${LEASE_F[holder]} == "$me:"* ]] ||
    die 2 "the environment lease is $(lease_describe), not this operator's ($me); the environment owner breaks it with lease --break"
  n=0
  meta_rm lease "$META_GEN" || n=$?
  ((n == 0)) || die 2 "the lease changed or could not be deleted ($(meta_class)); nothing changed"
  log "lease: released gs://$META_BUCKET/lease (generation $META_GEN) of ${LEASE_F[holder]}"
}

# ----------------------------------------------------------------------------------------------------------------- main
# The G070 mutators that take the environment lease (G072 Plan §10.5); build, push, status, receipt and lease do not.
LEASE_COMMANDS="provision deploy start stop reset retire"

main() {
  (($#)) || die 2 "usage: staging.sh <build|push|provision|deploy|start|stop|reset|retire|status|receipt|lease> --config FILE ... (see the header)"
  COMMAND=$1
  shift
  case $COMMAND in
    build) parse "--config --candidate --source" "$@"; need_candidate ;;
    push) parse "--config --candidate" "$@"; need_candidate ;;
    provision) parse "--config --candidate --duration --override --yes" "$@"; need_candidate ;;
    deploy) parse "--config --candidate --pin-secret" "$@"; need_candidate ;;
    start) parse "--config --duration" "$@" ;;
    stop | status) parse "--config" "$@" ;;
    reset) parse "--config --confirm" "$@" ;;
    retire) parse "--config --yes --purge-images" "$@" ;;
    receipt) parse "--config --candidate" "$@"; [[ -z ${OPT[candidate]:-} ]] || check_candidate "${OPT[candidate]}" ;;
    lease) parse "--config --show --release --break --confirm --reason" "$@" ;;
    *) die 2 "unknown command $COMMAND" ;;
  esac
  ((!DRY_RUN)) || log "dry-run: printing commands only; nothing runs"
  [[ " $LEASE_COMMANDS " != *" $COMMAND "* ]] || lease_guard "$COMMAND" "${OPT[candidate]:-}"
  "cmd_$COMMAND"
}

# Sourcing this file (a test) defines the commands without running main.
if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then main "$@"; fi
