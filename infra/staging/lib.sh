# shellcheck shell=bash
# Operator-side library of infra/staging/staging.sh (docs/foundation/STAGING.md). Sourced, never executed.
#
# staging.conf (private, 0600, outside Git; example: infra/staging/examples/staging.conf.example) holds non-secret
# identifiers only, as KEY=VALUE lines; it is parsed, never sourced. Secret values live only in Secret Manager
# (<env>-<NAME>) and reach only the VM (remote/lib.sh). The operator's own credentials: gcloud and the Terraform
# providers use the owner's Application Default Credentials impersonating the deployer service account
# (DEPLOYER_SA); Terraform's Cloudflare provider reads CLOUDFLARE_API_TOKEN from the operator's environment (passed to
# the container by name, never on a command line). Operator tools: bash, git, gcloud, docker, jq, tar, sha256sum.

CONF_REQUIRED=(STAGING_ENV GCP_PROJECT GCP_REGION GCP_ZONE DEPLOYER_SA AR_REPOSITORY BASE_DOMAIN CLOUDFLARE_ZONE_ID ACME_EMAIL
  EDGE_ACME_CA EDGE_DATA_VOLUME PRODUCT_NAME DATABASE_NAME OBJECT_STORE_BUCKET TERRAFORM_IMAGE EVIDENCE_DIR OWNER)
# DEFAULT_TENANT_SLUG and PLATFORM_ROOT_PAGE (G071, DOMAINS.md): optional non-secret core settings, passed through
# remote.env and the rendered env as given; empty or absent = unset (core's defaults apply, none is chosen here).
# MCP_CONNECT_ISSUER and MCP_ALLOWED_ORIGINS (MCP.md "Configuration (core)"): the same kind of setting for MCP. Empty or absent
# keeps MCP disabled (no route); the issuer is the exact AuthKit Connect issuer of this environment's WorkOS (a public value),
# and the origins need it (core refuses a partial configuration).
# SESSION_MAX_SECONDS: the session length a provision (first allocation) or start without --duration selects.
# GCLOUD_IMPERSONATE_SA (G072 Plan §10.9.2, additive): the identity gcloud impersonates (GFLAGS), exactly
# <env>-deployer, <env>-ci-stop or <env>-ci-read @<GCP_PROJECT>.iam.gserviceaccount.com; absent or empty =
# DEPLOYER_SA (today's behaviour). Terraform and push keep DEPLOYER_SA.
CONF_OPTIONAL=(PRIVATE_DIR GOOGLE_ADC_FILE DEFAULT_TENANT_SLUG PLATFORM_ROOT_PAGE MCP_CONNECT_ISSUER MCP_ALLOWED_ORIGINS
  SESSION_MAX_SECONDS GCLOUD_IMPERSONATE_SA)
# Session length (G070 Plan r1.9): the native max_run_duration with action STOP is the only timer. One default (the
# Terraform variable defaults, the examples and STAGING.md state the same value; check.sh STAGING-GREP pins them) and
# the policy range, enforced here before any command and again by the module's variable validation. No clamping.
SESSION_DEFAULT_SECONDS=7200
SESSION_FLOOR=600
SESSION_CEILING=28800
# An existing VM's length is written by start (gcloud set-scheduling while TERMINATED, then a refresh-only apply; G070
# A2 fallback): hashicorp/google replaces the VM for this field, so the module sets it only at creation and ignores it
# afterwards. refresh_scope_check: the VM in the root, the one attribute whose drift must be the length change, and the
# status-like computed attributes a stopped VM's refresh may also report (any other drift is a HOLD).
SESSION_ADDRESS=module.vm.google_compute_instance.vm
SESSION_PATH='scheduling[0].max_run_duration[0].seconds'
SESSION_DRIFT_STATUS=(current_status cpu_platform)
# set-scheduling flags besides the length: every other scheduling attribute stated as the module provisions it (never
# --clear-*, never --termination-time); scheduling_check reads all of them back.
SESSION_SCHED_FLAGS=(--instance-termination-action=STOP --maintenance-policy=MIGRATE --restart-on-failure --provisioning-model=STANDARD)
SESSION_SCHED_FMT='value(status,scheduling.maxRunDuration.seconds,scheduling.instanceTerminationAction,scheduling.provisioningModel,scheduling.automaticRestart,scheduling.onHostMaintenance,scheduling.preemptible)'
LE_PRODUCTION_CA=https://acme-v02.api.letsencrypt.org/directory
LE_STAGING_CA=https://acme-staging-v02.api.letsencrypt.org/directory
# Services built from the candidate (Compose stanzas of local.yml/migrate.yml, plus the edge).
BUILT_SERVICES=(core web intelligence minio minio-init migrate edge)

DRY_RUN=0
REPO=""
CONF_FILE=""
COMMAND=""
ENV_NAME=""
VM=""
GFLAGS=()
declare -A CONF=()

ts() { date -u +%Y-%m-%dT%H:%M:%SZ; }
log() { printf 'staging: %s\n' "$*" >&2; }
die() { local code=$1; shift; printf 'staging: %s\n' "$*" >&2; exit "$code"; }

# forbid ARGS...: the verbs this tooling never runs: volume-removing down, migrate down/reapply, prune, create, and
# instance/disk/address create or delete outside Terraform.
forbid() {
  local line=" $* " re
  # `--profile migrate` names a Compose profile, not the migrate command (see remote/lib.sh forbid).
  line=${line// --profile / --profile=}
  for re in ' down( [^ ]+)* (-v|--volumes)( |$)' ' migrate (down|reapply)( |$)' \
    ' (system|volume|image|container|network|builder) prune( |$)' ' create( |$)' \
    ' (instances|disks|addresses|networks|firewall-rules) delete( |$)'; do
    [[ ! $line =~ $re ]] || die 2 "refused a forbidden command (${BASH_REMATCH[0]# })"
  done
}

# show CMD...: print one command line as it would run (--dry-run: on stdout, the reviewable plan; otherwise on stderr,
# so command output that is parsed stays clean).
show() {
  if ((DRY_RUN)); then
    printf '+'
    printf ' %q' "$@"
    printf '\n'
  else
    { printf '+'; printf ' %q' "$@"; printf '\n'; } >&2
  fi
}

# run CMD...: refuse forbidden verbs; print the command; execute it unless --dry-run.
run() {
  forbid "$@"
  show "$@"
  ((DRY_RUN)) || "$@"
}

# mcp_settings_valid ISSUER ORIGINS: the optional MCP settings (MCP.md), both empty or well formed; sets MCP_SETTINGS_ERR to the
# refusal text (a key name, never a value). The issuer is a bare https host (no path, query or trailing slash); the origins are
# comma-separated exact https origins, no spaces, and need an issuer (core refuses MCP configured partially). Core's own
# validation stays authoritative; this only refuses a value before it reaches a VM.
MCP_HOST_RE='https://[a-z0-9]([a-z0-9.-]*[a-z0-9])?'
mcp_settings_valid() {
  local issuer=$1 origins=$2
  MCP_SETTINGS_ERR=""
  if [[ -n $issuer && ! $issuer =~ ^$MCP_HOST_RE$ ]]; then
    MCP_SETTINGS_ERR="MCP_CONNECT_ISSUER must be empty or an exact https issuer (https://<host>, no path or trailing slash)"
  elif [[ -n $origins && ! $origins =~ ^$MCP_HOST_RE(:[0-9]{1,5})?(,$MCP_HOST_RE(:[0-9]{1,5})?)*$ ]]; then
    MCP_SETTINGS_ERR="MCP_ALLOWED_ORIGINS must be empty or comma-separated exact https origins without spaces"
  elif [[ -n $origins && -z $issuer ]]; then
    MCP_SETTINGS_ERR="MCP_ALLOWED_ORIGINS needs MCP_CONNECT_ISSUER"
  fi
  [[ -z $MCP_SETTINGS_ERR ]]
}

# load_conf FILE: parse and validate staging.conf into CONF. Messages name keys, never values of other files.
load_conf() {
  local file=$1 line k v n=0
  [[ -f $file ]] || die 2 "no config file $file"
  CONF_FILE="$(cd "$(dirname "$file")" && pwd -P)/$(basename "$file")"
  while IFS= read -r line || [[ -n $line ]]; do
    n=$((n + 1))
    [[ $line =~ ^[[:space:]]*(#|$) ]] && continue
    [[ $line =~ ^([A-Z][A-Z0-9_]*)=(.*)$ ]] || die 2 "staging.conf line $n is not KEY=VALUE"
    k=${BASH_REMATCH[1]} v=${BASH_REMATCH[2]}
    if [[ $v =~ ^\"(.*)\"$ ]]; then v=${BASH_REMATCH[1]}; fi
    [[ " ${CONF_REQUIRED[*]} ${CONF_OPTIONAL[*]} " == *" $k "* ]] || die 2 "staging.conf: unknown key $k"
    [[ -z ${CONF[$k]+x} ]] || die 2 "staging.conf: $k is set twice"
    [[ $v != *[\'\"\$\`\\]* ]] || die 2 "staging.conf: $k holds a quote, \$, backquote or backslash"
    CONF[$k]=$v
  done <"$file"
  for k in "${CONF_REQUIRED[@]}"; do [[ -n ${CONF[$k]:-} ]] || die 2 "staging.conf: $k is required"; done
  [[ -n ${CONF[PRIVATE_DIR]:-} ]] || CONF[PRIVATE_DIR]="$(dirname "$CONF_FILE")"
  [[ -n ${CONF[GOOGLE_ADC_FILE]:-} ]] || CONF[GOOGLE_ADC_FILE]="${HOME:-/root}/.config/gcloud/application_default_credentials.json"
  [[ -n ${CONF[SESSION_MAX_SECONDS]:-} ]] || CONF[SESSION_MAX_SECONDS]=$SESSION_DEFAULT_SECONDS
  validate_conf
  ENV_NAME=${CONF[STAGING_ENV]}
  VM=$ENV_NAME-vm
  GFLAGS=(--project "${CONF[GCP_PROJECT]}" --impersonate-service-account "${CONF[GCLOUD_IMPERSONATE_SA]:-${CONF[DEPLOYER_SA]}}")
}

validate_conf() {
  local e=${CONF[STAGING_ENV]} p=${CONF[GCP_PROJECT]} r=${CONF[GCP_REGION]} k s=${CONF[SESSION_MAX_SECONDS]}
  [[ $e =~ ^[a-z][a-z0-9-]{1,20}$ ]] || die 2 "staging.conf: STAGING_ENV must match ^[a-z][a-z0-9-]{1,20}\$"
  [[ $p =~ ^[a-z][a-z0-9-]{4,28}[a-z0-9]$ ]] || die 2 "staging.conf: GCP_PROJECT is not a project ID"
  [[ $r =~ ^[a-z]+-[a-z]+[0-9]+$ ]] || die 2 "staging.conf: GCP_REGION is not a region"
  [[ ${CONF[GCP_ZONE]} =~ ^$r-[a-z]$ ]] || die 2 "staging.conf: GCP_ZONE is not a zone of GCP_REGION"
  [[ ${CONF[DEPLOYER_SA]} == "$e-deployer@$p.iam.gserviceaccount.com" ]] || die 2 "staging.conf: DEPLOYER_SA must be $e-deployer@<GCP_PROJECT>.iam.gserviceaccount.com"
  [[ ${CONF[AR_REPOSITORY]} =~ ^[a-z0-9-]+-docker\.pkg\.dev/$p/[a-z][a-z0-9-]*$ ]] || die 2 "staging.conf: AR_REPOSITORY must be <location>-docker.pkg.dev/<GCP_PROJECT>/<repository>"
  valid_hostname "${CONF[BASE_DOMAIN]}" || die 2 "staging.conf: BASE_DOMAIN must be a lowercase hostname of two or more labels"
  [[ ${CONF[CLOUDFLARE_ZONE_ID]} =~ ^[0-9a-f]{32}$ ]] || die 2 "staging.conf: CLOUDFLARE_ZONE_ID must be 32 hex digits"
  [[ ${CONF[ACME_EMAIL]} =~ ^[^@[:space:]]+@[a-z0-9.-]+\.[a-z]+$ ]] || die 2 "staging.conf: ACME_EMAIL is not an address"
  case "${CONF[EDGE_ACME_CA]} ${CONF[EDGE_DATA_VOLUME]}" in
    "$LE_PRODUCTION_CA edge-data" | "$LE_STAGING_CA edge-data-rehearsal") ;;
    *) die 2 "staging.conf: EDGE_ACME_CA/EDGE_DATA_VOLUME must pair the production directory with edge-data or the staging directory with edge-data-rehearsal" ;;
  esac
  [[ -n ${CONF[PRODUCT_NAME]} && ${CONF[PRODUCT_NAME]} != *[[:cntrl:]]* ]] || die 2 "staging.conf: PRODUCT_NAME"
  [[ ${CONF[DATABASE_NAME]} =~ ^[a-z_][a-z0-9_]{0,62}$ ]] || die 2 "staging.conf: DATABASE_NAME is not a database name"
  [[ ${CONF[OBJECT_STORE_BUCKET]} =~ ^[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]$ ]] || die 2 "staging.conf: OBJECT_STORE_BUCKET is not a bucket name"
  [[ ${CONF[TERRAFORM_IMAGE]} =~ ^hashicorp/terraform:[0-9]+\.[0-9]+\.[0-9]+@sha256:[0-9a-f]{64}$ ]] || die 2 "staging.conf: TERRAFORM_IMAGE must be hashicorp/terraform:<x.y.z>@sha256:<digest>"
  for k in EVIDENCE_DIR PRIVATE_DIR GOOGLE_ADC_FILE; do
    [[ ${CONF[$k]} == /* && ${CONF[$k]} != *[[:space:]]* ]] || die 2 "staging.conf: $k must be an absolute path without spaces"
  done
  [[ ${CONF[OWNER]} != *[[:cntrl:]]* ]] || die 2 "staging.conf: OWNER"
  [[ -z ${CONF[DEFAULT_TENANT_SLUG]:-} || ${CONF[DEFAULT_TENANT_SLUG]} =~ ^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$ ]] ||
    die 2 "staging.conf: DEFAULT_TENANT_SLUG must be empty or a slug (^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?\$)"
  [[ ${CONF[PLATFORM_ROOT_PAGE]:-} =~ ^(signup_redirect|landing)?$ ]] ||
    die 2 "staging.conf: PLATFORM_ROOT_PAGE must be empty, signup_redirect or landing"
  mcp_settings_valid "${CONF[MCP_CONNECT_ISSUER]:-}" "${CONF[MCP_ALLOWED_ORIGINS]:-}" || die 2 "staging.conf: $MCP_SETTINGS_ERR"
  [[ $s =~ ^[1-9][0-9]{0,5}$ ]] && ((s >= SESSION_FLOOR && s <= SESSION_CEILING)) ||
    die 2 "staging.conf: SESSION_MAX_SECONDS must be whole seconds from $SESSION_FLOOR to $SESSION_CEILING"
  case ${CONF[GCLOUD_IMPERSONATE_SA]:-} in
    "" | "$e-deployer@$p.iam.gserviceaccount.com" | "$e-ci-stop@$p.iam.gserviceaccount.com" | "$e-ci-read@$p.iam.gserviceaccount.com") ;;
    *) die 2 "staging.conf: GCLOUD_IMPERSONATE_SA must be empty or $e-deployer, $e-ci-stop or $e-ci-read @<GCP_PROJECT>.iam.gserviceaccount.com" ;;
  esac
}

# duration_seconds D: --duration D (<n>m, <n>h or whole seconds) as whole seconds on stdout; status 1 for any other
# form or a value outside SESSION_FLOOR..SESSION_CEILING (refused, never clamped).
duration_seconds() {
  local s
  [[ $1 =~ ^([1-9][0-9]{0,5})([mh]?)$ ]] || return 1
  s=${BASH_REMATCH[1]}
  case ${BASH_REMATCH[2]} in
    m) s=$((s * 60)) ;;
    h) s=$((s * 3600)) ;;
  esac
  ((s >= SESSION_FLOOR && s <= SESSION_CEILING)) || return 1
  printf '%s\n' "$s"
}

valid_hostname() {
  local h=$1 l labels
  ((${#h} <= 253)) || return 1
  IFS=. read -ra labels <<<"$h"
  ((${#labels[@]} >= 2)) || return 1
  for l in "${labels[@]}"; do [[ $l =~ ^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$ ]] || return 1; done
}

# candidate SHA: a full commit ID present in this repository.
check_candidate() {
  [[ $1 =~ ^[0-9a-f]{40}$ ]] || die 2 "--candidate must be a full 40-digit commit ID"
  git -C "$REPO" cat-file -e "$1^{commit}" 2>/dev/null || die 2 "--candidate $1 is not a commit of this repository"
}

# private_dirs: the operator's private directories (0700), created on first use.
private_dirs() {
  ((DRY_RUN)) && return 0
  install -d -m 0700 "${CONF[EVIDENCE_DIR]}" "${CONF[EVIDENCE_DIR]}/deploy" "${CONF[EVIDENCE_DIR]}/logs" \
    "${CONF[EVIDENCE_DIR]}/receipts" "${CONF[EVIDENCE_DIR]}/tmp" "${CONF[PRIVATE_DIR]}" "${CONF[PRIVATE_DIR]}/plans"
}

# inventory CATEGORY ID ACTION LIFECYCLE: one row of the environment's resource inventory, written when the resource is
# created or changed (EVIDENCE_DIR/inventory.tsv).
inventory() {
  local f=${CONF[EVIDENCE_DIR]}/inventory.tsv row
  row="$(printf '%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s' "$(ts)" "$ENV_NAME" "$COMMAND" "$1" "$2" "$3" "${CONF[OWNER]}" "$4")"
  if ((DRY_RUN)); then printf '[inventory] %s\n' "$row"; return 0; fi
  [[ -s $f ]] || printf 'time\tenvironment\tcommand\tcategory\tid\taction\towner\tlifecycle\n' >"$f"
  printf '%s\n' "$row" >>"$f"
}

manifest_path() { printf '%s/deploy/%s.tsv' "${CONF[EVIDENCE_DIR]}" "$1"; }

# manifest_ref MANIFEST SERVICE: the registry digest reference of SERVICE ("" or "-" when not pushed).
manifest_ref() { awk -F'\t' -v s="$2" '$1 == "image" && $2 == s {print $6}' "$1"; }

# require_pushed MANIFEST CANDIDATE: every built service has a registry digest from CANDIDATE under AR_REPOSITORY.
require_pushed() {
  local m=$1 c=$2 svc ref
  if ((DRY_RUN)) && [[ ! -f $m ]]; then log "dry-run: no manifest $m yet (build and push write it)"; return 0; fi
  [[ -f $m ]] || die 2 "no deploy manifest for $c: run build and push first"
  [[ "$(awk -F'\t' '$1 == "candidate" {print $2}' "$m")" == "$c" ]] || die 2 "the manifest is not for $c"
  for svc in "${BUILT_SERVICES[@]}"; do
    ref="$(manifest_ref "$m" "$svc")"
    [[ $ref == "${CONF[AR_REPOSITORY]}/$svc@sha256:"* && ${ref##*@sha256:} =~ ^[0-9a-f]{64}$ ]] || die 2 "the manifest has no registry digest for $svc: run push"
    [[ "$(awk -F'\t' -v s="$svc" '$1 == "image" && $2 == s {print $3}' "$m")" == "$c" ]] || die 2 "manifest: $svc comes from another candidate"
  done
}

# tf ARGS...: Terraform in the pinned image, on the committed example root (infra/terraform/environments/example,
# read-only) with the private backend.hcl, terraform.tfvars, data directory and saved plans under PRIVATE_DIR.
tf() {
  ((DRY_RUN)) || [[ -f ${CONF[GOOGLE_ADC_FILE]} ]] || die 2 "no Application Default Credentials file (gcloud auth application-default login)"
  run docker run --rm --user "$(id -u):$(id -g)" \
    --volume "$REPO/infra/terraform:/src:ro" --volume "${CONF[PRIVATE_DIR]}:/private" \
    --volume "${CONF[GOOGLE_ADC_FILE]}:/adc.json:ro" \
    --env HOME=/tmp --env TF_DATA_DIR=/private/tf-data --env TF_IN_AUTOMATION=1 --env TF_INPUT=0 \
    --env GOOGLE_APPLICATION_CREDENTIALS=/adc.json --env "GOOGLE_IMPERSONATE_SERVICE_ACCOUNT=${CONF[DEPLOYER_SA]}" \
    --env CLOUDFLARE_API_TOKEN --label "staging.environment=$ENV_NAME" --label staging.tool=terraform \
    --workdir /src/environments/example "${CONF[TERRAFORM_IMAGE]}" "$@"
}

tf_init() { tf init -input=false -lockfile=readonly -backend-config=/private/backend.hcl; }

need_cloudflare_token() {
  ((DRY_RUN)) && return 0
  [[ -n ${CLOUDFLARE_API_TOKEN:-} ]] || die 2 "set CLOUDFLARE_API_TOKEN (Terraform's zone-scoped token) in this shell first"
  export CLOUDFLARE_API_TOKEN
}

# vm_read: VMSTATE = the instance's status (RUNNING, TERMINATED, ...), or ABSENT only on positive confirmation: the
# describe failed and the zone's instance list, filtered on exact name equality, succeeded and is empty. Status 1
# (VMSTATE=UNREADABLE) otherwise: a failed list, the name listed although its describe failed, or no status reported.
vm_read() {
  local listed
  if VMSTATE="$(gcloud compute instances describe "$VM" --zone "${CONF[GCP_ZONE]}" "${GFLAGS[@]}" --format 'value(status)' 2>/dev/null)"; then
    [[ -z $VMSTATE ]] || return 0
  elif listed="$(gcloud compute instances list --zones "${CONF[GCP_ZONE]}" "${GFLAGS[@]}" --filter "name=$VM" --format 'value(name)')" &&
    [[ -z $listed ]]; then
    VMSTATE=ABSENT
    return 0
  fi
  VMSTATE=UNREADABLE
  return 1
}

# vm_status: VMSTATE as vm_read reads it; dry-run assumes RUNNING. An unreadable VM is a HOLD (exit 3) for every
# caller: it is never taken for absent, so nothing is planned, applied, deployed, started or stopped from a guess and no
# default session length is inferred.
VMSTATE=""
vm_status() {
  if ((DRY_RUN)); then
    show gcloud compute instances describe "$VM" --zone "${CONF[GCP_ZONE]}" "${GFLAGS[@]}" --format 'value(status)'
    VMSTATE=RUNNING
    return 0
  fi
  vm_read || die 3 "HOLD: the state of $VM could not be read (no status, and the exact-name instance list did not confirm it absent); nothing planned, applied, deployed, started or stopped"
}

# vm_session: the instance's session as Compute Engine reports it: SESSION_STATE (status), SESSION_HAVE (the
# max_run_duration seconds) and SESSION_STOP (resourceStatus.scheduling.terminationTimestamp, armed at each start and
# empty while stopped). Status 1 when the instance cannot be read. A dry run reads nothing (placeholders). The first
# output line is split on each tab by parameter expansion, not `IFS=$'\t' read` (tab is IFS whitespace there, so an
# empty column collapses and shifts the next one): a missing column stays empty, never takes its neighbour's value.
SESSION_STATE="" SESSION_HAVE="" SESSION_STOP=""
vm_session() {
  local fmt='value(status,scheduling.maxRunDuration.seconds,resourceStatus.scheduling.terminationTimestamp)' out
  if ((DRY_RUN)); then
    show gcloud compute instances describe "$VM" --zone "${CONF[GCP_ZONE]}" "${GFLAGS[@]}" --format "$fmt"
    SESSION_STATE=RUNNING SESSION_HAVE='<observed>' SESSION_STOP='<terminationTimestamp>'
    return 0
  fi
  out="$(gcloud compute instances describe "$VM" --zone "${CONF[GCP_ZONE]}" "${GFLAGS[@]}" --format "$fmt")" || return 1
  out=${out%%$'\n'*}
  SESSION_STATE=${out%%$'\t'*} SESSION_HAVE="" SESSION_STOP=""
  [[ $out == *$'\t'* ]] || return 0
  out=${out#*$'\t'}
  SESSION_HAVE=${out%%$'\t'*}
  [[ $out != *$'\t'* ]] || SESSION_STOP=${out#*$'\t'}
}

# session_pair: SESSION_HAVE from the instance and SESSION_TF from Terraform's max_run_duration_seconds output, both
# whole positive seconds. Either unreadable, missing or malformed is a HOLD (exit 3): nothing is applied or started and
# no default is inferred. A dry run reads nothing.
SESSION_TF=""
session_pair() {
  if ((DRY_RUN)); then vm_session; tf output -raw max_run_duration_seconds; return 0; fi
  vm_session || die 3 "HOLD: the session length of $VM could not be read; nothing applied or started"
  [[ $SESSION_HAVE =~ ^[1-9][0-9]*$ ]] || die 3 "HOLD: $VM reports no max run duration; nothing applied or started"
  SESSION_TF="$(tf output -raw max_run_duration_seconds)" ||
    die 3 "HOLD: Terraform's max_run_duration_seconds output could not be read; nothing applied or started"
  [[ $SESSION_TF =~ ^[1-9][0-9]*$ ]] ||
    die 3 "HOLD: Terraform's max_run_duration_seconds output is not whole seconds ([$SESSION_TF]); nothing applied or started"
}

# session_mismatch: the HOLD (exit 3) for an instance and Terraform that disagree on the length (drift, or a partial
# session change), naming both values and the exact manual reconcile steps (docs/foundation/STAGING.md "Session
# length"). start itself recovers only the one partial state it can prove: a stopped VM that already reports the
# requested length while Terraform differs; every other mismatch (a running VM included) ends here.
session_mismatch() {
  local how
  # GFLAGS (the configured --project and --impersonate-service-account) stay in the printed command, as in the one
  # set_session runs: followed under another active gcloud project or identity it must not address another VM.
  printf -v how 'gcloud compute instances set-scheduling %s --zone %s %s --max-run-duration=%ss %s' "$VM" "${CONF[GCP_ZONE]}" \
    "${GFLAGS[*]}" "$SESSION_TF" "${SESSION_SCHED_FLAGS[*]}"
  die 3 "HOLD: $VM has a ${SESSION_HAVE}s session length but Terraform records ${SESSION_TF}s; nothing applied or started: reconcile first (docs/foundation/STAGING.md \"Session length\"), with the VM stopped (staging.sh stop): to keep the instance's ${SESSION_HAVE}s run start --duration ${SESSION_HAVE} (it refreshes Terraform, verifies both and starts); to return to Terraform's ${SESSION_TF}s run $how, then start again"
}

# session_hold WHY WANT: the HOLD (exit 3) of a session change that stopped after it may have changed the instance: the
# VM is not started; prints the partial state as read now (best effort) and the recovery.
session_hold() {
  local why=$1 want=$2 inst=unreadable rec
  if vm_session 2>/dev/null && [[ -n $SESSION_HAVE ]]; then inst="${SESSION_HAVE}s (${SESSION_STATE:-status unknown})"; fi
  if rec="$(tf output -raw max_run_duration_seconds 2>/dev/null)" && [[ -n $rec ]]; then rec=${rec}s; else rec=unreadable; fi
  die 3 "HOLD: $why; $VM not started; partial state: the instance reports $inst, Terraform records $rec; recovery: with the VM stopped run start --duration $want again: once the instance reports ${want}s it only refreshes Terraform, verifies both and starts, and when the instance still matches Terraform it repeats the change; any other state HOLDs with the manual reconcile steps (docs/foundation/STAGING.md \"Session length\")"
}

# vm_fresh_terminated: the VM's state read just now must be TERMINATED before a session change (a RUNNING or
# unreadable VM never gets one): HOLD (exit 3) otherwise, nothing changed.
vm_fresh_terminated() {
  if ((DRY_RUN)); then
    show gcloud compute instances describe "$VM" --zone "${CONF[GCP_ZONE]}" "${GFLAGS[@]}" --format 'value(status)'
    return 0
  fi
  vm_read || die 3 "HOLD: the state of $VM could not be read just before the session change; nothing changed or started"
  [[ $VMSTATE == TERMINATED ]] || die 3 "HOLD: $VM is $VMSTATE, not TERMINATED, just before the session change; nothing changed or started"
}

# tsv_fields LINE: FIELDS = LINE split on every tab; an empty column stays an empty field (see vm_session).
FIELDS=()
tsv_fields() {
  local rest=$1
  FIELDS=()
  while [[ $rest == *$'\t'* ]]; do
    FIELDS+=("${rest%%$'\t'*}")
    rest=${rest#*$'\t'}
  done
  FIELDS+=("$rest")
}

# scheduling_check WANT: the instance's whole scheduling read back after set-scheduling: TERMINATED, max run duration
# WANT, action STOP, provisioning model STANDARD, automatic restart true, host maintenance MIGRATE, not preemptible. A
# failed read, a missing column or any other value is a HOLD (session_hold): the VM is not started.
scheduling_check() {
  local want=$1 out bad=""
  if ((DRY_RUN)); then
    show gcloud compute instances describe "$VM" --zone "${CONF[GCP_ZONE]}" "${GFLAGS[@]}" --format "$SESSION_SCHED_FMT"
    printf '+ (HOLD unless TERMINATED with maxRunDuration %s s, STOP, STANDARD, automaticRestart true, MIGRATE, preemptible false)\n' "$want"
    return 0
  fi
  out="$(gcloud compute instances describe "$VM" --zone "${CONF[GCP_ZONE]}" "${GFLAGS[@]}" --format "$SESSION_SCHED_FMT")" ||
    session_hold "the scheduling of $VM could not be read back" "$want"
  tsv_fields "${out%%$'\n'*}"
  ((${#FIELDS[@]} == 7)) || session_hold "the scheduling readback of $VM has ${#FIELDS[@]} columns, not 7" "$want"
  [[ ${FIELDS[0]} == TERMINATED ]] || bad+="status [${FIELDS[0]}], not TERMINATED; "
  [[ ${FIELDS[1]} == "$want" ]] || bad+="maxRunDuration.seconds [${FIELDS[1]}], not $want; "
  [[ ${FIELDS[2]} == STOP ]] || bad+="instanceTerminationAction [${FIELDS[2]}], not STOP; "
  [[ ${FIELDS[3]} == STANDARD ]] || bad+="provisioningModel [${FIELDS[3]}], not STANDARD; "
  [[ ${FIELDS[4],,} == true ]] || bad+="automaticRestart [${FIELDS[4]}], not true; "
  [[ ${FIELDS[5]} == MIGRATE ]] || bad+="onHostMaintenance [${FIELDS[5]}], not MIGRATE; "
  [[ ${FIELDS[6],,} == false ]] || bad+="preemptible [${FIELDS[6]}], not false; "
  [[ -z $bad ]] || session_hold "the scheduling read back from $VM differs (${bad%; })" "$want"
}

# session_var: SESSION_VAR = -var session_max_seconds=<the instance's observed length> for the Terraform calls of status
# and retire, so they never plan a scheduling change of their own; the configured default only when no instance exists
# (vm_status confirmed it). An existing instance whose session cannot be read, or that reports a missing or malformed
# length (not whole positive seconds), is a HOLD (exit 3) before any Terraform plan, never the default.
SESSION_VAR=()
session_var() {
  local s=${CONF[SESSION_MAX_SECONDS]}
  if [[ $VMSTATE != ABSENT ]]; then
    vm_session || die 3 "HOLD: the session length of $VM could not be read; nothing planned"
    ((DRY_RUN)) || [[ $SESSION_HAVE =~ ^[1-9][0-9]*$ ]] ||
      die 3 "HOLD: the session length of $VM is missing or malformed (max run duration [$SESSION_HAVE]); nothing planned"
    s=$SESSION_HAVE
  fi
  # shellcheck disable=SC2034 # read by staging.sh (status, retire)
  SESSION_VAR=(-var "session_max_seconds=$s")
}

# session_detail: one line for status and the receipt from the last vm_session: configured seconds, VM state and the
# scheduled stop (the armed terminationTimestamp while RUNNING).
session_detail() {
  local stop="not armed (stopped)" have=${SESSION_HAVE:-unknown}
  [[ ! $have =~ ^[0-9]+$ ]] || have=${have}s
  [[ $SESSION_STATE != RUNNING ]] || stop=${SESSION_STOP:-not armed}
  printf 'configured %s; %s; scheduled stop %s' "$have" "${SESSION_STATE:-unknown}" "$stop"
}

# wait_status WANT MAX_SECONDS: poll the instance until it reports WANT.
wait_status() {
  local want=$1 max=$2 i
  ((DRY_RUN)) && { printf '+ (poll the instance status until %s, at most %ss)\n' "$want" "$max"; return 0; }
  for ((i = 0; i < max; i += 5)); do
    # A read that fails mid-poll (after a start or stop was issued) keeps polling; only WANT itself ends the wait.
    vm_read || true
    [[ $VMSTATE == "$want" ]] && return 0
    sleep 5
  done
  die 4 "$VM did not reach $want within ${max}s (last: $VMSTATE)"
}

# vm_ssh COMMAND: one command on the VM over OS Login + IAP (no external SSH, no project keys).
vm_ssh() { run gcloud compute ssh "$VM" --zone "${CONF[GCP_ZONE]}" --tunnel-through-iap "${GFLAGS[@]}" --quiet --command "$1"; }

# wait_vm_ready: IAP/OS Login reachable and the startup script's Docker Engine with the Compose plugin installed.
wait_vm_ready() {
  local i
  ((DRY_RUN)) && { vm_ssh 'sudo docker compose version'; return 0; }
  for ((i = 0; i < 180; i++)); do
    if gcloud compute ssh "$VM" --zone "${CONF[GCP_ZONE]}" --tunnel-through-iap "${GFLAGS[@]}" --quiet \
      --command 'sudo docker compose version' >/dev/null 2>&1; then
      return 0
    fi
    sleep 5
  done
  die 4 "$VM is not reachable over IAP with Docker ready after 15 minutes"
}

# remote SCRIPT ARGS...: run a VM-side script of the current release as root; its key<TAB>value output is kept in
# EVIDENCE_DIR/logs. In --dry-run the same script prints its VM commands locally instead.
remote() {
  local script=$1 logf rc=0 cmd=""
  shift
  (($# == 0)) || printf -v cmd ' %q' "$@"
  if ((DRY_RUN)); then
    vm_ssh "sudo bash /srv/$ENV_NAME/current/infra/staging/remote/$script --env $ENV_NAME$cmd"
    # The VM script's own dry-run status is the result: a refusal there must fail the dry run, not vanish.
    bash "$REPO/infra/staging/remote/$script" --env "$ENV_NAME" "$@" --dry-run || rc=$?
    return "$rc"
  fi
  logf="${CONF[EVIDENCE_DIR]}/logs/$(date -u +%Y%m%dT%H%M%SZ)-$COMMAND-${script%.sh}.log"
  vm_ssh "sudo bash /srv/$ENV_NAME/current/infra/staging/remote/$script --env $ENV_NAME$cmd" | tee "$logf" || rc=$?
  log "VM output kept in $logf"
  return "$rc"
}

# env_object_names [KIND]: the exact names infra/terraform/modules/gcp-staging-vm gives this environment's Compute
# Engine objects of KIND (instances, disks, addresses, firewall-rules, networks, subnets), or of every kind. Exact names,
# never an <env>- prefix: environment names may extend one another (staging and staging-eu in one project), and the
# prefix of one would match the other's objects.
env_object_names() {
  local e=$ENV_NAME
  case ${1:-all} in
    instances) echo "$e-vm" ;;
    disks) echo "$e-vm $e-data" ;; # the boot disk takes the instance's name
    addresses) echo "$e-ip" ;;
    firewall-rules) echo "$e-edge $e-iap" ;;
    networks) echo "$e-vpc" ;;
    subnets) echo "$e-subnet" ;;
    all) echo "$e-vm $e-data $e-ip $e-edge $e-iap $e-vpc $e-subnet" ;;
    *) die 2 "env_object_names: unknown kind $1" ;;
  esac
}

# env_name_filter KIND: those names as one anchored expression, ^(a|b)$, for gcloud --filter (RE2, through
# name_match_filter) and grep -E (an environment name holds only [a-z0-9-], so nothing needs escaping).
env_name_filter() { local n; n="$(env_object_names "$1")"; printf '^(%s)$' "${n// /|}"; }

# env_secret_filter: this environment's Secret Manager IDs, <env>-<NAME> with an upper-case NAME, anchored: another
# environment's secret never matches (staging-eu-FOO for staging leaves eu-FOO, which is not upper case).
env_secret_filter() { printf '^%s-[A-Z][A-Z0-9_]*$' "$ENV_NAME"; }

# name_match_filter RE: the gcloud --filter argument name ~ 'RE'. The expression is quoted in the filter language
# itself: an unquoted literal ends at a parenthesis there, so the shell's quoting alone would not keep ^(a|b)$ whole.
# Every name/regex filter goes through here; an expression never holds a single quote (names are [a-z0-9-]).
name_match_filter() {
  [[ $1 != *\'* ]] || die 2 "name_match_filter: a single quote in the expression"
  printf "name ~ '%s'" "$1"
}

# env_secrets: the IDs of this environment's Secret Manager secrets (gcloud matches the full resource name,
# projects/<n>/secrets/<id>, so the expression is anchored at the ID's end and after /secrets/).
env_secrets() {
  local re cmd
  re="$(env_secret_filter)"
  cmd=(gcloud secrets list --project "${CONF[GCP_PROJECT]}" --filter "$(name_match_filter "/secrets/${re#^}")" --format 'value(name.basename())')
  if ((DRY_RUN)); then show "${cmd[@]}"; return 0; fi
  "${cmd[@]}"
}

# env_owns LABEL NAME INSTANCE_NAME: a resource belongs to this environment (its environment label, or one of its exact
# object names as the resource's name or instance name).
env_owns() {
  local n names
  [[ $1 != "$ENV_NAME" ]] || return 0
  read -ra names <<<"$(env_object_names)"
  for n in "${names[@]}"; do
    [[ $2 != "$n" && $3 != "$n" ]] || return 0
  done
  return 1
}

# plan_changes PLANFILE: one line per resource the saved plan creates, changes or destroys: address, type, environment
# label, name and instance name, tab-separated ("-" for an absent or non-string value).
plan_changes() {
  tf show -json "$1" | jq -r '
    def str: if type == "string" and . != "" then . else "-" end;
    .resource_changes[]?
    | select(any(.change.actions[]; . != "no-op" and . != "read"))
    | (.change.after // .change.before // {}) as $v
    | [.address, .type, ($v.labels.environment? // $v.terraform_labels.environment? | str), ($v.name? | str),
       ($v.instance_name? | str)]
    | @tsv'
}

# label_check PLANFILE: every resource the saved plan creates, changes or destroys belongs to this environment
# (env_owns, or a DNS record under BASE_DOMAIN). Refuses otherwise.
label_check() {
  local plan=$1 rows addr type label name inst base=${CONF[BASE_DOMAIN]} bad=""
  if ((DRY_RUN)); then tf show -json "$plan"; printf '+ (refuse unless every changed resource belongs to %s)\n' "$ENV_NAME"; return 0; fi
  rows="$(plan_changes "$plan")" || die 2 "the saved plan could not be read"
  while IFS=$'\t' read -r addr type label name inst; do
    [[ -n $addr ]] || continue
    # An instance IAM binding names its instance by full path; only this environment's own project and zone reduce to the name.
    [[ $inst != "projects/${CONF[GCP_PROJECT]}/zones/${CONF[GCP_ZONE]}/instances/"* ]] || inst=${inst##*/}
    if env_owns "$label" "$name" "$inst"; then continue; fi
    if [[ $type == cloudflare_* && ($name == "$base" || $name == *".$base") ]]; then continue; fi
    bad+="$addr "
  done <<<"$rows"
  [[ -z $bad ]] || die 2 "the plan touches resources outside $ENV_NAME: ${bad% }"
}

# refresh_scope_rows PLANFILE: the saved refresh-only plan for refresh_scope_check, tab-separated: a "change ADDRESS
# ACTIONS" row per resource_changes entry with an action other than no-op (actions comma-joined; a refresh-only plan
# proposes none), a "drift ADDRESS ACTIONS" row per resource_drift entry, then a "diff ADDRESS PATH BEFORE AFTER" row
# for every leaf value that drift alters (PATH as scheduling[0].max_run_duration[0].seconds; BEFORE and AFTER as JSON,
# AFTER "unknown" when it is known only after apply). The caller compares values; it never prints another attribute's
# value. after_unknown contributes only its true leaves: Terraform keeps every known nested block there as {}, which
# as a path would compare the whole before/after block.
refresh_scope_rows() {
  tf show -json "$1" | jq -r '
    def leaves: [paths((type != "object" and type != "array") or length == 0)];
    def pstr: reduce .[] as $k (""; if ($k | type) == "number" then "\(.)[\($k)]" elif . == "" then $k else "\(.).\($k)" end);
    def diffs($a):
      (.change.before // {}) as $b | (.change.after // {}) as $f | (.change.after_unknown // {}) as $u
      | ($b | leaves) + ($f | leaves) + ($u | [paths(. == true)]) | unique | .[]
      | . as $p
      | (try ($b | getpath($p)) catch null) as $bv
      | (try ($f | getpath($p)) catch null) as $av
      | ((try ($u | getpath($p)) catch null) == true) as $unk
      | select($unk or $bv != $av)
      | ["diff", $a, ($p | pstr), ($bv | tojson), (if $unk then "unknown" else ($av | tojson) end)] | @tsv;
    (.resource_changes[]? | select(any(.change.actions[]; . != "no-op")) | ["change", .address, (.change.actions | join(","))] | @tsv),
    (.resource_drift[]? | .address as $a | (["drift", $a, (.change.actions | join(","))] | @tsv), diffs($a))'
}

# benign_drift ADDRESS PATH BEFORE AFTER: succeeds for a recorded-state lag leaf, never an unknown value. Terraform's
# recorded state of a VM that was never refreshed after its first start lags the provider's read of it: the attachment
# of this environment's VM (users, last_attach_timestamp; the exact configured project, zone and VM name, so the same
# name elsewhere is refused, as is any missing context), defaults the provider now reports as empty collections, and an
# IAM etag. Exactly those leaves on exactly those addresses are accepted, so the first custom-duration start after a
# provision passes without a separate state write; any other leaf or value still HOLDs.
benign_drift() {
  local a=$1 p=$2 y=$3 z=$4 inst re
  [[ $z != unknown && -n ${CONF[GCP_PROJECT]:-} && -n ${CONF[GCP_ZONE]:-} && -n $VM ]] || return 1 # no context, no match
  inst="https://www.googleapis.com/compute/v1/projects/${CONF[GCP_PROJECT]}/zones/${CONF[GCP_ZONE]}/instances/$VM"
  case $a in
    module.vm.google_compute_address.ip | module.vm.google_compute_disk.data)
      case $p in
        users) [[ $y == '[]' || $y == null ]] && [[ $z == "[\"$inst\"]" ]] ;;
        'users[0]') [[ $y == null ]] && [[ $z == "\"$inst\"" ]] ;;
        last_attach_timestamp) [[ $a == module.vm.google_compute_disk.data && ($y == null || $y == '""') ]] &&
          re='^"[0-9T:.+-]+"$' && [[ $z =~ $re ]] ;;
        *) return 1 ;;
      esac ;;
    module.vm.google_compute_firewall.edge | 'module.vm.google_compute_firewall.iap[0]')
      case $p in
        source_service_accounts | source_tags | target_service_accounts) [[ $y == null || $y == '[]' ]] && [[ $z == '[]' ]] ;;
        *) return 1 ;;
      esac ;;
    "$SESSION_ADDRESS")
      case $p in
        resource_policies | 'boot_disk[0].initialize_params[0].replica_zones') [[ $y == null || $y == '[]' ]] && [[ $z == '[]' ]] ;;
        'boot_disk[0].initialize_params[0].resource_manager_tags') [[ $y == null || $y == '{}' ]] && [[ $z == '{}' ]] ;;
        *) return 1 ;;
      esac ;;
    'module.vm.google_compute_instance_iam_member.os_login['*)
      [[ $p == etag ]] && re='^"[A-Za-z0-9+/=_-]+"$' && [[ $z =~ $re ]] ;;
    *) return 1 ;;
  esac
}

# refresh_scope_check PLANFILE FROM WANT: the saved refresh-only plan proposes no resource change, and its drift is
# SESSION_ADDRESS, updated (never deleted), altering SESSION_PATH from FROM (Terraform's recorded length) to WANT exactly
# once plus, at most, the known status-like values SESSION_DRIFT_STATUS and the recorded-state lag leaves of benign_drift
# (on the VM itself, the address, the data disk, both firewall rules and the OS Login IAM member, nowhere else). Another
# resource, a replacement or any other proposed change, another or an unknown attribute, a delete or a different value
# is a HOLD (session_hold): nothing is applied, the VM is not started and the plan is kept for review.
refresh_scope_check() {
  local plan=$1 from=$2 want=$3 rows kind addr x y z diffs=0 bad=""
  if ((DRY_RUN)); then
    tf show -json "$plan"
    printf '+ (HOLD unless the refresh-only plan proposes no change and only %s drifts: %s from %s to %s, plus %s)\n' \
      "$SESSION_ADDRESS" "$SESSION_PATH" "$from" "$want" "${SESSION_DRIFT_STATUS[*]}"
    return 0
  fi
  rows="$(refresh_scope_rows "$plan")" || session_hold "the saved refresh-only plan ${plan##*/} could not be read; nothing applied" "$want"
  while IFS=$'\t' read -r kind addr x y z; do
    case $kind in
      "") ;;
      change) bad+="it proposes $x of $addr; " ;;
      drift)
        case $addr in
          "$SESSION_ADDRESS" | module.vm.google_compute_address.ip | module.vm.google_compute_disk.data | \
            module.vm.google_compute_firewall.edge | 'module.vm.google_compute_firewall.iap[0]' | \
            'module.vm.google_compute_instance_iam_member.os_login['*) ;;
          *) bad+="$addr drifts too; " ;;
        esac
        [[ $x == update ]] || bad+="$addr drifts by $x, not update; "
        ;;
      diff)
        if [[ $addr != "$SESSION_ADDRESS" ]]; then # another address drifts only by recorded-state lag leaves
          benign_drift "$addr" "$x" "$y" "$z" || bad+="$addr also changes $x; "
          continue
        fi
        if [[ $x == "$SESSION_PATH" ]]; then
          diffs=$((diffs + 1))
          [[ $y == "$from" ]] || bad+="$SESSION_PATH is $y before, not Terraform's $from; "
          [[ $z == "$want" ]] || bad+="$SESSION_PATH becomes $z, not $want; "
        elif [[ " ${SESSION_DRIFT_STATUS[*]} " == *" $x "* ]]; then
          [[ $z != unknown ]] || bad+="$x becomes unknown; "
        elif benign_drift "$addr" "$x" "$y" "$z"; then
          :
        else
          bad+="it also changes $x; "
        fi
        ;;
      *) bad+="unreadable plan row; " ;;
    esac
  done <<<"$rows"
  ((diffs == 1)) || bad+="$diffs changes of $SESSION_PATH, not 1; "
  [[ -z $bad ]] || session_hold "the refresh-only plan is not the session-length drift alone (${bad%; }); nothing applied; plan kept as ${CONF[PRIVATE_DIR]}/plans/${plan##*/}" "$want"
}

# state_names: the name of every resource in the Terraform state (all modules).
state_names() {
  tf show -json | jq -r '.values.root_module // {} | [recurse(.child_modules[]?) | .resources[]? | .values.name? // empty] | .[]'
}

# live_names: the environment's Compute Engine objects by their exact environment label (instances, disks, addresses:
# a stray labelled object is found too) and by their exact names (firewall rules, networks, subnets, which carry no
# labels; env_name_filter, so another environment's objects are never listed).
live_names() {
  local kind
  for kind in instances disks addresses; do
    gcloud compute "$kind" list "${GFLAGS[@]}" --filter "labels.environment=$ENV_NAME" --format 'value(name)' || return 1
  done
  for kind in firewall-rules networks; do
    gcloud compute "$kind" list "${GFLAGS[@]}" --filter "$(name_match_filter "$(env_name_filter "$kind")")" --format 'value(name)' || return 1
  done
  gcloud compute networks subnets list "${GFLAGS[@]}" --filter "$(name_match_filter "$(env_name_filter subnets)")" --format 'value(name)' || return 1
}

# reconcile: label/name query of the environment's Compute Engine objects versus the Terraform state; prints both sets
# and refuses (exit 2) while an object exists that the state does not hold (orphan: import it, never a second one).
reconcile() {
  local names live kind orphans=""
  if ((DRY_RUN)); then
    tf show -json
    for kind in instances disks addresses; do
      show gcloud compute "$kind" list "${GFLAGS[@]}" --filter "labels.environment=$ENV_NAME" --format 'value(name)'
    done
    for kind in firewall-rules networks; do
      show gcloud compute "$kind" list "${GFLAGS[@]}" --filter "$(name_match_filter "$(env_name_filter "$kind")")" --format 'value(name)'
    done
    show gcloud compute networks subnets list "${GFLAGS[@]}" --filter "$(name_match_filter "$(env_name_filter subnets)")" --format 'value(name)'
    return 0
  fi
  names="$(state_names)" || die 2 "the Terraform state could not be read"
  live="$(live_names)" || die 2 "the Compute Engine label query failed"
  printf 'reconcile\tstate\t%s\n' "$(sort -u <<<"$names" | tr '\n' ' ')"
  printf 'reconcile\tlive\t%s\n' "$(sort -u <<<"$live" | tr '\n' ' ')"
  while IFS= read -r kind; do
    [[ -z $kind ]] || grep -qxF -- "$kind" <<<"$names" || orphans+="$kind "
  done <<<"$live"
  [[ -z $orphans ]] || die 2 "objects of $ENV_NAME outside the Terraform state: ${orphans}(terraform import them; never create a second one)"
}

# remote_env: the non-secret values the VM renders into app.env (keys = remote/lib.sh REMOTE_ENV_KEYS).
remote_env() {
  local b=${CONF[BASE_DOMAIN]}
  cat <<EOF
STAGING_ENV=$ENV_NAME
GCP_PROJECT=${CONF[GCP_PROJECT]}
REGISTRY_HOST=${CONF[AR_REPOSITORY]%%/*}
COMPOSE_PROJECT_NAME=$ENV_NAME
PRODUCT_NAME=${CONF[PRODUCT_NAME]}
PLATFORM_BASE_DOMAIN=$b
CUSTOM_DOMAIN_CNAME_TARGET=custom.$b
PLATFORM_AUTH_HOST=auth.$b
DATABASE_NAME=${CONF[DATABASE_NAME]}
APP_ENV=development
BIND_ADDRESS=127.0.0.1
WEB_PORT=3100
CORE_PORT=8180
INTELLIGENCE_PORT=8100
POSTGRES_PORT=5532
REDIS_PORT=6479
MINIO_API_PORT=9100
MINIO_CONSOLE_PORT=9101
CORE_INTERNAL_URL=http://core:8180
INTELLIGENCE_URL=http://intelligence:8000
REDIS_HOST=redis
REDIS_PORT_INTERNAL=6379
OBJECT_STORE_ENDPOINT=http://minio:9000
OBJECT_STORE_REGION=us-east-1
OBJECT_STORE_BUCKET=${CONF[OBJECT_STORE_BUCKET]}
OBJECT_STORE_FORCE_PATH_STYLE=true
OBJECT_STORE_ACCESS_KEY=core-service
MINIO_ROOT_USER=minio-root
EDGE_BASE_DOMAIN=$b
ACME_EMAIL=${CONF[ACME_EMAIL]}
EDGE_ACME_CA=${CONF[EDGE_ACME_CA]}
EDGE_DATA_VOLUME=${CONF[EDGE_DATA_VOLUME]}
DEFAULT_TENANT_SLUG=${CONF[DEFAULT_TENANT_SLUG]:-}
PLATFORM_ROOT_PAGE=${CONF[PLATFORM_ROOT_PAGE]:-}
MCP_CONNECT_ISSUER=${CONF[MCP_CONNECT_ISSUER]:-}
MCP_ALLOWED_ORIGINS=${CONF[MCP_ALLOWED_ORIGINS]:-}
EOF
}
