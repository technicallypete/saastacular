#!/usr/bin/env bash
# CI federation for one environment project (G072 Plan r1 §10.3, §10.8, §10.10; r0 §2.4; infra/terraform/README.md
# "CI federation"). Run by the environment owner with their own short-lived gcloud login, after bootstrap.sh, never by
# an agent before the live release.
#
#   ci-federation.sh --project ID --project-number N --region REGION --zone ZONE --env NAME
#                    --gitlab-project-path PATH --gitlab-namespace-id N [--repository NAME] [--state-bucket NAME] [--dry-run]
#   ci-federation.sh --verify  (the same arguments) [--dry-run]
#   ci-federation.sh --teardown --confirm ID --project ID --project-number N --region REGION --zone ZONE --env NAME
#                    [--confirm-meta NAME] [--repository NAME] [--state-bucket NAME] [--dry-run]
#
# Setup is idempotent: every item is created only when absent (a soft-deleted pool, provider or custom role is
# undeleted), the provider, both custom roles and the bucket settings are re-applied so a rerun converges, bindings are
# added (adding an existing binding is a no-op), and one receipt line per item says what happened. Per environment
# project it enables sts.googleapis.com; creates the workload identity pool <env>-ci with the OIDC provider "gitlab"
# (issuer https://gitlab.com, audiences staging-ci:<env>:start|stop|read, google.subject = assertion.sub,
# attribute.action = start|stop|read from the audience, else none, and the attribute condition below); the service
# accounts <env>-ci-stop and <env>-ci-read (<env>-deployer and <env>-vm are bootstrap.sh's, unchanged); the custom
# roles ciStop<env> (hyphens become underscores) and ciMetaMutable; the metadata bucket
# gs://<project>-<env>-ci-meta (versioned, uniform access, public-access prevention, 90-day noncurrent expiry, label
# managed-by=ci-federation) whose bindings come only from ci-meta-permissions.tsv; and the principal bindings of the
# README's CI table. It never grants Owner/Editor, project IAM administration, service-account keys or any Secret
# Manager role other than the start principal's accessor on <env>-TERRAFORM_CLOUDFLARE_API_TOKEN, and it binds nothing
# on the Terraform state bucket. The GitLab project ID is read from infra/ci/hosted-context.sh (STG_PROJECT_ID) and
# --gitlab-project-path must hash to its STG_PROJECT_PATH_SHA; the path itself is never written to the repository.
#
# --verify reads back (describe, get-iam-policy and list calls only) and compares every item with the rendered set,
# including the absence of any binding for the CI identities on the state bucket, on the project beyond the stop
# identity's two roles and on any other secret, and the absence of user-managed keys; exit 1 on any mismatch. A read
# that fails is never taken as absence: its item is UNREADABLE and verification fails.
# --teardown removes exactly the set setup creates (the API stays enabled, as bootstrap.sh never disables one) and
# refuses, before any change, while the metadata bucket holds a live object unless --confirm-meta <env> is given, or
# when its live objects cannot be listed.
# The metadata bucket name is global: setup and teardown use it only when its describe names this bucket with the
# pinned --project-number as owner (checked before setup's first change, before teardown's first change, after its
# creation and again right before every write naming it: settings, each IAM grant, the deletion; --verify compares
# it). A create that fails because the name was taken meanwhile is refused, not adopted. A foreign, missing or malformed owner and any failed describe other than
# a confirmed 404 (absent) are refused. Elsewhere an existence read is absent only on a not-found answer; any other
# failure stops the run.
# --dry-run prints mutating commands ("+ ") and verify reads ("? ") without calling gcloud; setup treats the created
# items as absent and the bootstrap prerequisites (and <env>-vm, the Cloudflare secret) as present, teardown treats
# every item as present.
set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)"
REPO="$(cd "$HERE/../../.." && pwd -P)"
PERMISSIONS_FILE="$HERE/ci-meta-permissions.tsv"
REGISTRY_FILE="$REPO/infra/staging/ci-environments.txt"
PIN_FILE="$REPO/infra/ci/hosted-context.sh"
ISSUER="https://gitlab.com"
CONFIG_HOST="gitlab.com"
PROVIDER="gitlab"
ACTIONS=(start stop read)
APIS=(sts.googleapis.com)
STOP_PERMISSIONS=(compute.instances.get compute.instances.list compute.instances.stop compute.disks.list
  compute.addresses.list compute.zoneOperations.get compute.projects.get)
MUTABLE_PERMISSIONS=(storage.objects.delete)
MUTABLE_ROLE_ID="ciMetaMutable"
CONDITION_TITLE="ci-meta-exact-keys"
IAP_CONDITION_TITLE="ci-stop-iap-instance"
SEP=$'\x1f'

usage() { sed -n '6,10p' "$0" | sed 's/^# \{0,1\}//' >&2; exit 2; }
die() { printf 'ci-federation: ERROR: %s\n' "$*" >&2; exit 1; }
receipt() { printf 'ci-federation\t%s\t%s\t%s\n' "$1" "$2" "$3"; }

project="" project_number="" region="" zone="" env_name="" gl_path="" gl_namespace="" repository="" bucket=""
confirm="" confirm_meta="" teardown=0 verify=0 dry_run=0
while (($#)); do
  case "$1" in
    --project) project="${2:?}"; shift ;;
    --project-number) project_number="${2:?}"; shift ;;
    --region) region="${2:?}"; shift ;;
    --zone) zone="${2:?}"; shift ;;
    --env) env_name="${2:?}"; shift ;;
    --gitlab-project-path) gl_path="${2:?}"; shift ;;
    --gitlab-namespace-id) gl_namespace="${2:?}"; shift ;;
    --repository) repository="${2:?}"; shift ;;
    --state-bucket) bucket="${2:?}"; shift ;;
    --confirm) confirm="${2:?}"; shift ;;
    --confirm-meta) confirm_meta="${2:?}"; shift ;;
    --teardown) teardown=1 ;;
    --verify) verify=1 ;;
    --dry-run) dry_run=1 ;;
    -h|--help) usage ;;
    *) die "unknown argument $1" ;;
  esac
  shift
done

# ---- input validation (before any gcloud call) -------------------------------------------------------------------------
((teardown + verify <= 1)) || die "--teardown and --verify are exclusive"
[[ "$project" =~ ^[a-z][a-z0-9-]{4,28}[a-z0-9]$ ]] || die "--project must be a GCP project ID"
[[ "$project_number" =~ ^[1-9][0-9]{5,19}$ ]] || die "--project-number must be the project's number (the pin checked against the project)"
[[ "$region" =~ ^[a-z][a-z0-9-]+[0-9]$ ]] || die "--region must be a GCP region"
[[ "$zone" =~ ^$region-[a-z]$ ]] || die "--zone must be a zone of --region"
# Service-account IDs need 6-30 characters: <env>-deployer (bootstrap.sh) fits for 3-21 character names.
[[ "$env_name" =~ ^[a-z][a-z0-9-]{2,20}$ ]] || die "--env must match ^[a-z][a-z0-9-]{2,20}\$"
[[ -f "$REGISTRY_FILE" ]] || die "environment registry infra/staging/ci-environments.txt is missing"
grep -qxF -- "$env_name" "$REGISTRY_FILE" || die "--env $env_name is not listed in infra/staging/ci-environments.txt"
if ((teardown)); then
  [[ "$confirm" == "$project" ]] || die "--teardown requires --confirm $project"
  [[ -z "$confirm_meta" || "$confirm_meta" == "$env_name" ]] || die "--confirm-meta must name this environment ($env_name)"
else
  [[ -z "$confirm$confirm_meta" ]] || die "--confirm and --confirm-meta belong to --teardown"
  pin() { sed -n -E "s/^$1=([0-9a-f]+)$/\1/p" "$PIN_FILE"; }
  gl_project_id="$(pin STG_PROJECT_ID 2>/dev/null || true)" path_sha="$(pin STG_PROJECT_PATH_SHA 2>/dev/null || true)"
  [[ "$gl_project_id" =~ ^[1-9][0-9]*$ && "$path_sha" =~ ^[0-9a-f]{64}$ ]] \
    || die "the GitLab project pins (STG_PROJECT_ID, STG_PROJECT_PATH_SHA) are not readable in infra/ci/hosted-context.sh"
  [[ "$gl_path" =~ ^[A-Za-z0-9][A-Za-z0-9._-]*(/[A-Za-z0-9][A-Za-z0-9._-]*)+$ ]] || die "--gitlab-project-path must be a GitLab project path (group/.../project)"
  [[ "$(printf '%s' "$gl_path" | sha256sum | cut -d' ' -f1)" == "$path_sha" ]] \
    || die "--gitlab-project-path does not match the pinned project (STG_PROJECT_PATH_SHA in infra/ci/hosted-context.sh)"
  [[ "$gl_namespace" =~ ^[1-9][0-9]{0,19}$ ]] || die "--gitlab-namespace-id must be the numeric ID of the project's namespace"
fi

repository="${repository:-$env_name-images}"
state_bucket="${bucket:-$project-tfstate}"
[[ "$repository" =~ ^[a-z][a-z0-9-]{0,62}$ ]] || die "--repository must be an Artifact Registry repository ID"
[[ "$state_bucket" =~ ^[a-z0-9][a-z0-9._-]{1,61}[a-z0-9]$ ]] || die "--state-bucket must be a bucket name"
meta_bucket="$project-$env_name-ci-meta"
((${#meta_bucket} <= 60)) || die "metadata bucket name $meta_bucket exceeds 60 characters"
[[ "$meta_bucket" != "$state_bucket" ]] || die "the metadata bucket must differ from the state bucket"
pool="$env_name-ci"
deployer_sa="$env_name-deployer@$project.iam.gserviceaccount.com"
vm_sa="$env_name-vm@$project.iam.gserviceaccount.com"
stop_sa="$env_name-ci-stop@$project.iam.gserviceaccount.com"
read_sa="$env_name-ci-read@$project.iam.gserviceaccount.com"
instance="$env_name-vm"
cf_secret="$env_name-TERRAFORM_CLOUDFLARE_API_TOKEN"
# Custom role IDs allow only [A-Za-z0-9_.]: the Plan's ciStop<env> with every hyphen as an underscore.
stop_role_id="ciStop${env_name//-/_}"
stop_role="projects/$project/roles/$stop_role_id"
mutable_role="projects/$project/roles/$MUTABLE_ROLE_ID"
pool_path="projects/$project_number/locations/global/workloadIdentityPools/$pool"
principal() { printf 'principalSet://iam.googleapis.com/%s/attribute.action/%s' "$pool_path" "$1"; }
join() { local IFS="$1"; shift; printf '%s' "$*"; }

# ---- the permissions file: the single source of the metadata bucket's bindings ---------------------------------------
META_ROWS=() # principal SEP grant SEP keys, validated
load_permissions() {
  local line n=0 who grant keys re=$'^([a-z-]+)\t([A-Za-z./]+)\t([^\t ]+)$'
  [[ -f "$PERMISSIONS_FILE" ]] || die "missing $PERMISSIONS_FILE"
  while IFS= read -r line || [[ -n "$line" ]]; do
    n=$((n + 1))
    [[ "$line" =~ $re ]] || die "ci-meta-permissions.tsv line $n is not principal<TAB>grant<TAB>keys"
    who=${BASH_REMATCH[1]} grant=${BASH_REMATCH[2]} keys=${BASH_REMATCH[3]}
    case "$who" in deployer|ci-stop|ci-read) ;; *) die "ci-meta-permissions.tsv line $n: unknown principal $who" ;; esac
    case "$grant" in
      roles/storage.objectAdmin|roles/storage.objectViewer|roles/storage.objectCreator|roles/storage.legacyBucketReader)
        [[ "$keys" == "*" ]] || die "ci-meta-permissions.tsv line $n: a predefined role is bucket-wide (keys *)" ;;
      "$MUTABLE_ROLE_ID")
        [[ "$keys" =~ ^[a-z0-9][a-z0-9._-]*(,[a-z0-9][a-z0-9._-]*)*$ ]] \
          || die "ci-meta-permissions.tsv line $n: $MUTABLE_ROLE_ID needs exact object keys, never *" ;;
      *) die "ci-meta-permissions.tsv line $n: grant $grant is not allowed" ;;
    esac
    # Least-privilege invariants: the read identity never writes, the stop identity never administers objects.
    case "$who:$grant" in
      ci-read:roles/storage.objectViewer|ci-read:roles/storage.legacyBucketReader|ci-stop:roles/storage.objectViewer| \
        ci-stop:roles/storage.objectCreator|ci-stop:roles/storage.legacyBucketReader|ci-stop:"$MUTABLE_ROLE_ID"|deployer:*) ;;
      *) die "ci-meta-permissions.tsv line $n: $grant is not allowed for $who" ;;
    esac
    [[ " ${META_ROWS[*]/#/|} " != *"|$who$SEP$grant$SEP"* ]] || die "ci-meta-permissions.tsv line $n: $who $grant repeated"
    META_ROWS+=("$who$SEP$grant$SEP$keys")
  done <"$PERMISSIONS_FILE"
  ((${#META_ROWS[@]})) || die "ci-meta-permissions.tsv has no rows"
}
load_permissions

member_of() {
  case "$1" in
    deployer) printf 'serviceAccount:%s' "$deployer_sa" ;;
    ci-stop) printf 'serviceAccount:%s' "$stop_sa" ;;
    ci-read) printf 'serviceAccount:%s' "$read_sa" ;;
  esac
}
key_condition() { # KEYS (comma-separated): the exact-object-name IAM condition
  local k out="" IFS=,
  for k in $1; do out+="${out:+ || }resource.name == \"projects/_/buckets/$meta_bucket/objects/$k\""; done
  printf '%s' "$out"
}
# IAP TCP forwarding checks roles/iap.tunnelResourceAccessor on the tunnel instance resource; a project binding
# conditioned on that exact resource name confines the stop identity's tunnel access to <env>-vm (Plan §2.4: the
# instance). gcloud has no per-instance IAP-TCP binding verb this script relies on.
iap_condition() { printf 'resource.name == "projects/%s/iap_tunnel/zones/%s/instances/%s"' "$project_number" "$zone" "$instance"; }
cond_title() { case "$1" in project) printf '%s' "$IAP_CONDITION_TITLE" ;; *) printf '%s' "$CONDITION_TITLE" ;; esac; }

# kind SEP resource SEP member SEP role SEP condition, one line per binding this script owns.
desired_bindings() {
  local row who grant keys role cond
  printf 'sa%s%s%s%s%sroles/iam.workloadIdentityUser%s\n' "$SEP" "$deployer_sa" "$SEP" "$(principal start)" "$SEP" "$SEP"
  printf 'secret%s%s%s%s%sroles/secretmanager.secretAccessor%s\n' "$SEP" "$cf_secret" "$SEP" "$(principal start)" "$SEP" "$SEP"
  printf 'sa%s%s%s%s%sroles/iam.workloadIdentityUser%s\n' "$SEP" "$stop_sa" "$SEP" "$(principal stop)" "$SEP" "$SEP"
  printf 'sa%s%s%s%s%sroles/iam.workloadIdentityUser%s\n' "$SEP" "$read_sa" "$SEP" "$(principal read)" "$SEP" "$SEP"
  printf 'project%s%s%sserviceAccount:%s%s%s%s\n' "$SEP" "$project" "$SEP" "$stop_sa" "$SEP" "$stop_role" "$SEP"
  printf 'project%s%s%sserviceAccount:%s%sroles/iap.tunnelResourceAccessor%s%s\n' "$SEP" "$project" "$SEP" "$stop_sa" "$SEP" "$SEP" "$(iap_condition)"
  printf 'instance%s%s%sserviceAccount:%s%sroles/compute.osAdminLogin%s\n' "$SEP" "$instance" "$SEP" "$stop_sa" "$SEP" "$SEP"
  printf 'sa%s%s%sserviceAccount:%s%sroles/iam.serviceAccountUser%s\n' "$SEP" "$vm_sa" "$SEP" "$stop_sa" "$SEP" "$SEP"
  printf 'repo%s%s%sserviceAccount:%s%sroles/artifactregistry.reader%s\n' "$SEP" "$repository" "$SEP" "$read_sa" "$SEP" "$SEP"
  for row in "${META_ROWS[@]}"; do
    IFS="$SEP" read -r who grant keys <<<"$row"
    role=$grant cond=""
    if [[ "$grant" == "$MUTABLE_ROLE_ID" ]]; then role=$mutable_role cond="$(key_condition "$keys")"; fi
    printf 'bucket%s%s%s%s%s%s%s%s\n' "$SEP" "$meta_bucket" "$SEP" "$(member_of "$who")" "$SEP" "$role" "$SEP" "$cond"
  done
}

# ---- command discipline -----------------------------------------------------------------------------------------------
ROLE_ALLOW=(roles/iam.workloadIdentityUser roles/secretmanager.secretAccessor roles/compute.osAdminLogin
  roles/iap.tunnelResourceAccessor roles/iam.serviceAccountUser roles/artifactregistry.reader "$stop_role" "$mutable_role")
for row in "${META_ROWS[@]}"; do IFS="$SEP" read -r _ grant _ <<<"$row"; [[ "$grant" == "$MUTABLE_ROLE_ID" ]] || ROLE_ALLOW+=("$grant"); done
# forbid ARGS...: the verbs and grants this script never issues: any service-account key operation, deletion of a
# bootstrap service account, secret/repository/compute create or delete, an API disable, any change on the state
# bucket, a basic role or a role outside ROLE_ALLOW, Secret Manager beyond the start principal's one accessor, and
# workloadIdentityUser for anything but a principal set, and iap.tunnelResourceAccessor without the instance condition.
forbid() {
  local line=" $* " re role
  for re in ' service-accounts keys ' ' keys (create|upload)( |$)' ' (secrets|repositories|instances|disks|addresses|networks|firewall-rules) (create|delete|update|start|stop|reset)( |$)' \
    ' versions( |$)' ' services disable( |$)' ' --role=roles/(owner|editor|viewer)( |$)'; do
    [[ ! $line =~ $re ]] || die "refused a forbidden command (${BASH_REMATCH[0]# })"
  done
  [[ $line != *" gs://$state_bucket "* && $line != *" gs://$state_bucket/"* ]] || die "refused a change on the state bucket gs://$state_bucket"
  if [[ $line =~ \ service-accounts\ delete\ ([^ ]+) ]]; then
    [[ ${BASH_REMATCH[1]} == "$stop_sa" || ${BASH_REMATCH[1]} == "$read_sa" ]] || die "refused deleting service account ${BASH_REMATCH[1]}"
  fi
  if [[ $line =~ \ --role=([^ ]+) ]]; then
    role=${BASH_REMATCH[1]}
    [[ " ${ROLE_ALLOW[*]} " == *" $role "* ]] || die "refused role $role"
    case "$role" in
      roles/secretmanager.*) [[ $line == *" secrets "*"-iam-policy-binding $cf_secret --member=$(principal start) "* ]] \
        || die "refused $role outside the start principal's accessor on $cf_secret" ;;
      roles/iam.workloadIdentityUser) [[ $line == *" --member=principalSet://iam.googleapis.com/$pool_path/attribute.action/"* ]] \
        || die "refused $role for a member other than this pool's principal sets" ;;
      roles/iap.tunnelResourceAccessor) [[ $line == *" --condition=expression=$(iap_condition),title=$IAP_CONDITION_TITLE "* ]] \
        || die "refused $role without the instance condition (never project-wide)" ;;
    esac
  fi
}

# meta_guard ARGV: a fresh owner check (meta_owner) right before every write naming gs://$meta_bucket other than its
# creation, so no cached state is trusted: a deleted bucket's global name can be reused by another project within
# seconds, and a settings update, IAM grant or deletion must never reach that replacement. Gone or foreign: exit 1.
meta_guard() {
  local line=" $* "
  [[ $line == *" gs://$meta_bucket "* || $line == *" gs://$meta_bucket/"* ]] || return 0
  [[ $line != *" storage buckets create gs://$meta_bucket "* ]] || return 0
  meta_owner present
  [[ $META_STATE == present ]] || die "gs://$meta_bucket no longer exists before a change to it; refused (fail closed)"
}
# Mutating call (`g ...` is gcloud on this project): refused by forbid, printed in dry-run mode, owner-checked by
# meta_guard. No secret in arguments.
run() {
  [[ "$1" != g ]] || { shift; set -- gcloud --project "$project" --quiet "$@"; }
  forbid "$@"
  if ((dry_run)); then printf '+ %s\n' "$*"; return 0; fi
  meta_guard "$@"
  "$@"
}
g() { gcloud --project "$project" --quiet "$@"; }
first_line() { local l=${1%%$'\n'*}; printf '%s' "${l:0:200}"; }
# gerr_class ERR: denied | error | notfound | nomatch, decided by gcloud's status marker only. Every URL, resource path
# and name this script passes is removed first (as meta-lib.sh's meta_class strips gs:// URLs), so a legal name that
# contains 404, NOT_FOUND or "not found" (project demo-404-x, its bucket, accounts, pool) never decides the class. A
# denial (it can also say "or it may not exist") and any other status code or status token win over not-found;
# not-found needs an explicit marker (HTTPError/HTTP/status/code 404, [404], (404), "not found: 404", NOT_FOUND,
# compute's "The resource '...' was not found"); nomatch is gcloud's "One or more URLs matched no objects" alone.
ERR_NAMES=("$deployer_sa" "$vm_sa" "$stop_sa" "$read_sa" "$meta_bucket" "$state_bucket" "$pool_path" "$stop_role" "$mutable_role"
  "$cf_secret" "$repository" "$instance" "$stop_role_id" "$pool" "$project" "$project_number" "$env_name")
RE_DENIED='PERMISSION_DENIED|AccessDenied|[Ff]orbidden|does not have|[Pp]ermission.*denied'
RE_CODE='(HTTPError|HTTP|[Ss]tatus|[Cc]ode)[ :=]+[0-9]{3}([^0-9]|$)|\[[0-9]{3}\]|\([0-9]{3}\)|: [0-9]{3}\.?$'
RE_TOKEN='(^|[^A-Z_])(UNAVAILABLE|INTERNAL|DEADLINE_EXCEEDED|RESOURCE_EXHAUSTED|UNAUTHENTICATED|FAILED_PRECONDITION|ABORTED|UNKNOWN|ALREADY_EXISTS|INVALID_ARGUMENT|CANCELLED|OUT_OF_RANGE|UNIMPLEMENTED|DATA_LOSS)([^A-Z_]|$)'
RE_NF="(HTTPError|HTTP|[Ss]tatus|[Cc]ode)[ :=]+404([^0-9]|\$)|\\[404\\]|\\(404\\)|not found: 404([^0-9]|\$)|(^|[^A-Z_])NOT_FOUND([^A-Z_]|\$)|The resource '[^']*' was not found"
gerr_class() {
  local e n code
  e="$(sed -E "s#(gs|https?)://[^[:space:]'\"]*#<url>#g; s#projects/[^[:space:]'\"]*#<resource>#g" <<<"$1")"
  for n in "${ERR_NAMES[@]}"; do e=${e//"$n"/<name>}; done
  if [[ $e =~ $RE_DENIED ]]; then echo denied; return 0; fi
  while read -r code; do
    [[ -z $code || $code == 404 ]] && continue
    if [[ $code == 403 ]]; then echo denied; else echo error; fi
    return 0
  done < <(grep -oE "$RE_CODE" <<<"$e" | grep -oE '[0-9]{3}' || true)
  if [[ $e =~ $RE_TOKEN ]]; then echo error
  elif [[ $e =~ $RE_NF ]]; then echo notfound
  elif [[ $e == *"One or more URLs matched no objects"* ]]; then echo nomatch
  else echo error; fi
}
not_found() { [[ "$(gerr_class "$1")" == notfound ]]; }
# gcall ARGS...: one gcloud read in this shell; G_OUT = stdout, G_ERR = stderr, the status is gcloud's.
G_OUT="" G_ERR=""
gcall() {
  local rc=0
  g "$@" >"$tmp_dir/gcall.out" 2>"$tmp_dir/gcall.err" || rc=$?
  G_OUT="$(cat "$tmp_dir/gcall.out")" G_ERR="$(cat "$tmp_dir/gcall.err")"
  return "$rc"
}
# probe ARGS...: 0 found (G_OUT), 1 a confirmed not-found; any other failure stops the run (exit 1, also from an if).
probe() {
  gcall "$@" && return 0
  not_found "$G_ERR" && return 1
  die "gcloud $1 ${2:-} ${3:-} failed and is not a not-found answer ($(first_line "$G_ERR")); refused (fail closed)"
}
# Existence of an item this script creates; dry-run treats it as absent.
exists() { ((dry_run)) && return 1; probe "$@"; }
# Existence of a prerequisite (bootstrap item, the VM, the owner's secret); dry-run treats it as present.
present() { ((dry_run)) && return 0; probe "$@"; }
# value_of ARGS...: VALUE = the read's text, empty on a not-found answer and in dry-run; any other failure stops the run.
# A global, not a command substitution, so the stop is never confined to a subshell.
VALUE=""
value_of() { VALUE=""; ((dry_run)) && return 0; if probe "$@"; then VALUE=$G_OUT; fi; return 0; }
# Verify read: printed in dry-run ("? ", on the original stdout, fd 4, past any capture), never a mutation. It runs in
# command substitutions, so a failure is recorded in RD_FAIL (and on stderr) for check/read_ok, never passed off as
# empty output that could match an expected absence.
exec 4>&1
RD_FAIL=""
rd() {
  local rc=0
  if ((dry_run)); then printf '? gcloud --project %s --quiet %s\n' "$project" "$*" >&4; return 0; fi
  g "$@" 2>"$tmp_dir/rd.err" || rc=$?
  if ((rc)); then
    printf 'gcloud %s: exit %s\n' "$*" "$rc" >>"$RD_FAIL"
    printf 'ci-federation: verify read failed (exit %s): gcloud %s %s %s: %s\n' "$rc" "$1" "$2" "${3:-}" "$(first_line "$(cat "$tmp_dir/rd.err")")" >&2
  fi
  return 0
}
# IAM changes on a just-created service account are eventually consistent; retry a binding a few times.
retry() {
  local attempt=1
  until run "$@"; do
    ((attempt >= 6)) && return 1
    attempt=$((attempt + 1))
    sleep 10
  done
}
tmp_dir=""
trap '[[ -z "$tmp_dir" ]] || rm -rf "$tmp_dir"' EXIT

POLICY_CMD=()
policy_cmd() { # KIND RESOURCE: the policy read, one "role<TAB>member<TAB>condition" line per member
  case "$1" in
    project) POLICY_CMD=(projects get-iam-policy "$2") ;;
    sa) POLICY_CMD=(iam service-accounts get-iam-policy "$2") ;;
    secret) POLICY_CMD=(secrets get-iam-policy "$2") ;;
    repo) POLICY_CMD=(artifacts repositories get-iam-policy "$2" --location="$region") ;;
    instance) POLICY_CMD=(compute instances get-iam-policy "$2" --zone="$zone") ;;
    bucket) POLICY_CMD=(storage buckets get-iam-policy "gs://$2") ;;
  esac
  POLICY_CMD+=('--flatten=bindings[].members' '--format=value(bindings.role,bindings.members,bindings.condition.expression)')
}
BINDING_CMD=()
binding_cmd() { # add|remove KIND RESOURCE MEMBER ROLE CONDITION
  local verb="$1-iam-policy-binding" kind=$2 res=$3 member=$4 role=$5 cond=$6 extra=()
  case "$kind" in
    project) BINDING_CMD=(projects "$verb" "$res") ;;
    sa) BINDING_CMD=(iam service-accounts "$verb" "$res") ;;
    secret) BINDING_CMD=(secrets "$verb" "$res") ;;
    repo) BINDING_CMD=(artifacts repositories "$verb" "$res" --location="$region") ;;
    instance) BINDING_CMD=(compute instances "$verb" "$res" --zone="$zone") ;;
    bucket) BINDING_CMD=(storage buckets "$verb" "gs://$res") ;;
  esac
  case "$kind" in
    project|bucket)
      if [[ -n "$cond" ]]; then extra=("--condition=expression=$cond,title=$(cond_title "$kind")"); else extra=(--condition=None); fi ;;
  esac
  BINDING_CMD+=("--member=$member" "--role=$role" "${extra[@]}" --format=none)
}
# The member lines of one policy that belong to this script's principals (the deployer only on the metadata bucket).
ours() { # KIND RESOURCE <policy lines: role TAB member TAB condition>
  local kind=$1 res=$2 role member cond
  while IFS=$'\t' read -r role member cond; do
    [[ -n "$role" ]] || continue
    case "$member" in
      "serviceAccount:$stop_sa"|"serviceAccount:$read_sa"|*"/workloadIdentityPools/$pool/"*) ;;
      "serviceAccount:$deployer_sa") [[ "$kind:$res" == "bucket:$meta_bucket" ]] || continue ;;
      *) continue ;;
    esac
    printf '%s%s%s%s%s%s%s%s%s\n' "$kind" "$SEP" "$res" "$SEP" "$member" "$SEP" "$role" "$SEP" "${cond%"${cond##*[![:space:]]}"}"
  done
}
has_binding() { # KIND RESOURCE MEMBER ROLE: teardown's presence check; dry-run treats it as present
  local lines
  ((dry_run)) && return 0
  policy_cmd "$1" "$2"
  # A not-found resource has no binding; a failed policy read stops the run (value_of), never reads as "absent".
  value_of "${POLICY_CMD[@]}"
  lines="$(ours "$1" "$2" <<<"$VALUE")"
  grep -qF -- "$1$SEP$2$SEP$3$SEP$4$SEP" <<<"$lines"
}

# meta_owner DRY-STATE: META_STATE = present when the describe names exactly gs://$meta_bucket and its owner's project
# number is the pinned --project-number, absent on a confirmed not-found (404). A foreign, missing or malformed owner,
# another name and any other failure (a denial, a backend error) are refused, exit 1. Dry-run reads nothing.
META_STATE=""
meta_owner() {
  local name num
  if ((dry_run)); then META_STATE=$1; return 0; fi
  if gcall storage buckets describe "gs://$meta_bucket" --raw --format='value(name,projectNumber)'; then
    name=${G_OUT%%$'\n'*} num=""
    [[ $name != *$'\t'* ]] || { num=${name#*$'\t'}; name=${name%%$'\t'*}; }
    [[ $name == "$meta_bucket" ]] || die "the describe of gs://$meta_bucket answered another name; refused (fail closed)"
    [[ $num =~ ^[1-9][0-9]{0,19}$ ]] || die "gs://$meta_bucket reports no readable owner project number; refused (fail closed)"
    [[ $num == "$project_number" ]] \
      || die "gs://$meta_bucket belongs to project number $num, not $project ($project_number); refused (fail closed)"
    META_STATE=present
  elif not_found "$G_ERR"; then META_STATE=absent
  else die "gs://$meta_bucket could not be checked ($(first_line "$G_ERR")); refused (fail closed)"; fi
}
# meta_live: LIVE = the live objects in gs://$meta_bucket (present per meta_owner). gcloud answers a listing that
# matched nothing with "One or more URLs matched no objects" (gerr_class nomatch: no status marker, the names
# removed): confirmed empty. Any other failure (a 404 for the bucket just described, a denial, a backend error) is
# refused, exit 1, never counted as empty.
LIVE=0
meta_live() {
  LIVE=0
  if gcall storage ls "gs://$meta_bucket/**"; then LIVE="$(grep -c . <<<"$G_OUT" || true)"; return 0; fi
  [[ "$(gerr_class "$G_ERR")" != nomatch ]] || return 0
  die "the live objects of gs://$meta_bucket could not be listed ($(first_line "$G_ERR")); refused before any change (fail closed)"
}

preflight() {
  local name
  ((dry_run)) && return 0
  command -v gcloud >/dev/null 2>&1 || die "gcloud not found"
  [[ -n "$(gcloud config get-value account 2>/dev/null)" ]] || die "no active gcloud account; run gcloud auth login"
  [[ "$(g projects describe "$project" --format='value(projectId)' 2>/dev/null)" == "$project" ]] \
    || die "project $project is not accessible to the active account"
  [[ "$(g projects describe "$project" --format='value(projectNumber)' 2>/dev/null)" == "$project_number" ]] \
    || die "project $project does not have the pinned number $project_number"
  ((teardown)) && return 0
  for name in "$deployer_sa" "$vm_sa"; do
    present iam service-accounts describe "$name" || die "$name is missing; run bootstrap.sh for $env_name first"
  done
  present artifacts repositories describe "$repository" --location="$region" || die "repository $repository is missing; run bootstrap.sh first"
}

# ---- setup ------------------------------------------------------------------------------------------------------------
enable_apis() {
  local enabled="" api
  ((dry_run)) || enabled="$(g services list --enabled --format='value(config.name)')"
  for api in "${APIS[@]}"; do
    if grep -qxF "$api" <<<"$enabled"; then receipt api "$api" existing
    else run g services enable "$api" --format=none; receipt api "$api" enabled; fi
  done
}

audiences() { local a out=(); for a in "${ACTIONS[@]}"; do out+=("staging-ci:$env_name:$a"); done; join , "${out[@]}"; }
# attribute.action = the action of an exact audience, else "none" (no commas: gcloud splits the mapping on them).
action_mapping() {
  local a expr=""
  for a in "${ACTIONS[@]}"; do expr+="assertion.aud == \"staging-ci:$env_name:$a\" ? \"$a\" : "; done
  printf 'google.subject=assertion.sub,attribute.action=%s"none"' "$expr"
}
attribute_condition() {
  local a auds="" clauses c expr=""
  for a in "${ACTIONS[@]}"; do auds+="${auds:+, }\"staging-ci:$env_name:$a\""; done
  clauses=(
    "assertion.project_id == \"$gl_project_id\""
    "assertion.namespace_id == \"$gl_namespace\""
    "assertion.ref_type == \"branch\""
    "assertion.ref == \"staging\""
    "assertion.ref_protected == \"true\""
    "assertion.ci_config_ref_uri == \"$CONFIG_HOST/$gl_path//.gitlab-ci.yml@refs/heads/staging\""
    "assertion.pipeline_source in [\"web\", \"api\"]"
    "assertion.environment == \"$env_name\""
    "assertion.environment_protected == \"true\""
    "assertion.runner_environment == \"gitlab-hosted\""
    "assertion.aud in [$auds]"
  )
  for c in "${clauses[@]}"; do expr+="${expr:+ && }$c"; done
  printf '%s' "$expr"
}

pool_and_provider() {
  local state flags
  value_of iam workload-identity-pools describe "$pool" --location=global --format='value(state)'; state=$VALUE
  case "$state" in
    ACTIVE) receipt pool "$pool" existing ;;
    DELETED) run g iam workload-identity-pools undelete "$pool" --location=global --format=none; receipt pool "$pool" undeleted ;;
    *) run g iam workload-identity-pools create "$pool" --location=global --display-name="CI $env_name" \
         --description="GitLab CI federation for $env_name (G072)" --format=none
       receipt pool "$pool" created ;;
  esac
  flags=("--workload-identity-pool=$pool" --location=global "--issuer-uri=$ISSUER" "--allowed-audiences=$(audiences)"
    "--attribute-mapping=$(action_mapping)" "--attribute-condition=$(attribute_condition)" --format=none)
  value_of iam workload-identity-pools providers describe "$PROVIDER" --workload-identity-pool="$pool" --location=global --format='value(state)'; state=$VALUE
  case "$state" in
    ACTIVE) run g iam workload-identity-pools providers update-oidc "$PROVIDER" "${flags[@]}"; receipt provider "$pool/$PROVIDER" "existing (re-applied)" ;;
    DELETED) run g iam workload-identity-pools providers undelete "$PROVIDER" --workload-identity-pool="$pool" --location=global --format=none
      run g iam workload-identity-pools providers update-oidc "$PROVIDER" "${flags[@]}"; receipt provider "$pool/$PROVIDER" "undeleted (re-applied)" ;;
    *) run g iam workload-identity-pools providers create-oidc "$PROVIDER" "${flags[@]}"; receipt provider "$pool/$PROVIDER" created ;;
  esac
}

service_account() {
  local email="$1" id="${1%%@*}" name="$2"
  if exists iam service-accounts describe "$email"; then receipt service-account "$email" existing
  else run g iam service-accounts create "$id" --display-name="$name" --format=none; receipt service-account "$email" created; fi
}

custom_role() { # ID TITLE PERMISSIONS...
  local id=$1 title=$2; shift 2
  if exists iam roles describe "$id" --format='value(deleted)'; then
    if [[ "$G_OUT" == True ]]; then run g iam roles undelete "$id" --format=none; fi
    run g iam roles update "$id" --permissions="$(join , "$@")" --stage=GA --format=none
    receipt custom-role "projects/$project/roles/$id" "existing (permissions re-applied)"
  else
    run g iam roles create "$id" --title="$title" --permissions="$(join , "$@")" --stage=GA --format=none
    receipt custom-role "projects/$project/roles/$id" created
  fi
}

meta_bucket_create() { # after meta_owner (setup checks the owner before its first change; run rechecks every write)
  local lifecycle="$tmp_dir/lifecycle.json"
  if [[ $META_STATE == present ]]; then receipt meta-bucket "gs://$meta_bucket" existing
  else
    # A failed create (the global name taken since the owner check) is refused, never adopted; a created bucket is
    # owner-checked before anything else touches it.
    run g storage buckets create "gs://$meta_bucket" --location="$region" --uniform-bucket-level-access --public-access-prevention \
      || die "gs://$meta_bucket could not be created (its global name may have been taken since the owner check); refused, not adopted"
    meta_owner present; [[ $META_STATE == present ]] || die "gs://$meta_bucket is not readable after its creation; refused (fail closed)"
    receipt meta-bucket "gs://$meta_bucket" created
  fi
  # Noncurrent metadata versions expire after 90 days; live objects are never touched by this rule.
  printf '%s\n' '{"rule":[{"action":{"type":"Delete"},"condition":{"isLive":false,"daysSinceNoncurrentTime":90}}]}' >"$lifecycle"
  run g storage buckets update "gs://$meta_bucket" --versioning --uniform-bucket-level-access --public-access-prevention \
    --lifecycle-file="$lifecycle" --update-labels="managed-by=ci-federation"
  receipt meta-bucket-settings "gs://$meta_bucket" "versioning,uniform-access,public-access-prevention,noncurrent-90d,managed-by=ci-federation"
}

bind_all() {
  local kind res member role cond
  while IFS="$SEP" read -r -u 3 kind res member role cond; do
    if [[ "$kind" == instance ]] && ! present compute instances describe "$res" --zone="$zone"; then
      receipt binding "$member" "$role@$res awaiting-provision (rerun after staging.sh provision)"; continue
    fi
    if [[ "$kind" == secret ]] && ! present secrets describe "$res"; then
      receipt binding "$member" "$role@$res awaiting-owner (the owner creates the secret, then reruns)"; continue
    fi
    binding_cmd add "$kind" "$res" "$member" "$role" "$cond"
    retry g "${BINDING_CMD[@]}"
    receipt binding "$member" "$role@$res${cond:+ (exact keys)}"
  done 3< <(desired_bindings)
}

setup() {
  preflight
  meta_owner absent
  enable_apis
  pool_and_provider
  service_account "$stop_sa" "CI stop $env_name"
  service_account "$read_sa" "CI read $env_name"
  custom_role "$stop_role_id" "CI stop $env_name" "${STOP_PERMISSIONS[@]}"
  custom_role "$MUTABLE_ROLE_ID" "CI metadata mutable objects" "${MUTABLE_PERMISSIONS[@]}"
  meta_bucket_create
  bind_all
  receipt "done" "$env_name" "project=$project pool=$pool_path provider=$PROVIDER audiences=$(audiences) meta=gs://$meta_bucket"
}

# ---- verify (read-only) -----------------------------------------------------------------------------------------------
VERIFY_BAD=0 RD_SEEN=0
read_ok() { # ITEM: false, with ITEM UNREADABLE and verification failed, when a read since the last item failed
  local n
  n="$(grep -c . "$RD_FAIL" || true)"
  ((n > RD_SEEN)) || return 0
  RD_SEEN=$n
  receipt verify "$1" UNREADABLE; VERIFY_BAD=1
  return 1
}
check() { # ITEM WANT GOT (GOT's reads ran before; a failed one makes ITEM UNREADABLE, never ok)
  ((dry_run)) && return 0
  read_ok "$1" || return 0
  if [[ "$2" == "$3" ]]; then receipt verify "$1" ok
  else receipt verify "$1" MISMATCH; printf '  want: %s\n  got:  %s\n' "${2//$'\n'/ | }" "${3//$'\n'/ | }" >&2; VERIFY_BAD=1; fi
}
norm_list() { tr ';,' '\n\n' | sed '/^$/d' | LC_ALL=C sort; }

verify_all() {
  local want got kind res secrets_list name
  preflight
  # "|| true" covers only grep's no-match (sts absent: a MISMATCH); a failed list is recorded by rd (UNREADABLE).
  got="$(rd services list --enabled --format='value(config.name)' | grep -xF sts.googleapis.com || true)"
  check api sts.googleapis.com "$got"
  check pool ACTIVE "$(rd iam workload-identity-pools describe "$pool" --location=global --format='value(state)')"
  for name in state oidc.issuerUri attributeCondition; do
    case "$name" in state) want=ACTIVE ;; oidc.issuerUri) want=$ISSUER ;; attributeCondition) want="$(attribute_condition)" ;; esac
    check "provider.$name" "$want" "$(rd iam workload-identity-pools providers describe "$PROVIDER" --workload-identity-pool="$pool" --location=global --format="value($name)")"
  done
  check provider.oidc.allowedAudiences "$(audiences | norm_list)" \
    "$(rd iam workload-identity-pools providers describe "$PROVIDER" --workload-identity-pool="$pool" --location=global --format='value(oidc.allowedAudiences)' | norm_list)"
  check provider.attributeMapping "$(action_mapping | norm_list)" \
    "$(rd iam workload-identity-pools providers describe "$PROVIDER" --workload-identity-pool="$pool" --location=global --format='value(attributeMapping)' | norm_list)"
  for name in "$deployer_sa" "$vm_sa" "$stop_sa" "$read_sa"; do
    check "service-account $name" "$name" "$(rd iam service-accounts describe "$name" --format='value(email)')"
  done
  for name in "$stop_sa" "$read_sa"; do
    check "user-managed keys $name" "" "$(rd iam service-accounts keys list --iam-account="$name" --managed-by=user --format='value(name)')"
  done
  check "custom-role $stop_role_id" "$(join , "${STOP_PERMISSIONS[@]}" | norm_list)" "$(rd iam roles describe "$stop_role_id" --format='value(includedPermissions)' | norm_list)"
  check "custom-role $stop_role_id deleted" "" "$(rd iam roles describe "$stop_role_id" --format='value(deleted)' | sed '/^False$/d')"
  check "custom-role $MUTABLE_ROLE_ID" "$(join , "${MUTABLE_PERMISSIONS[@]}" | norm_list)" "$(rd iam roles describe "$MUTABLE_ROLE_ID" --format='value(includedPermissions)' | norm_list)"
  check "custom-role $MUTABLE_ROLE_ID deleted" "" "$(rd iam roles describe "$MUTABLE_ROLE_ID" --format='value(deleted)' | sed '/^False$/d')"
  check "meta-bucket owner" "$meta_bucket $project_number" \
    "$(rd storage buckets describe "gs://$meta_bucket" --raw --format='value(name,projectNumber)' | tr '\t' ' ')"
  for name in versioning_enabled uniform_bucket_level_access public_access_prevention; do
    case "$name" in public_access_prevention) want=enforced ;; *) want=True ;; esac
    check "meta-bucket $name" "$want" "$(rd storage buckets describe "gs://$meta_bucket" --format="value($name)")"
  done
  got="$(rd storage buckets describe "gs://$meta_bucket" --format='value(labels)')"
  check "meta-bucket label managed-by" yes "$([[ "$got" == *"managed-by=ci-federation"* ]] && echo yes || echo no)"
  got="$(rd storage buckets describe "gs://$meta_bucket" --format='json(lifecycle_config)' | tr -d ' \n')"
  check "meta-bucket noncurrent-90d" yes "$([[ "$got" == *'"isLive":false'* && "$got" == *'"daysSinceNoncurrentTime":90'* ]] && echo yes || echo no)"
  # Every policy this script's principals could appear in: exactly the rendered bindings, nothing else. The state
  # bucket's expected set is empty; so is every secret's except the start principal's accessor on the Cloudflare one.
  # A failed secrets list would silently drop the unexpected-secret checks: it fails verification on its own item.
  secrets_list="$(rd secrets list --format='value(name.basename())')"
  read_ok "secrets list" || true
  while IFS="$SEP" read -r -u 3 kind res; do
    want="$(desired_bindings | awk -F"$SEP" -v k="$kind" -v r="$res" '$1 == k && $2 == r' | LC_ALL=C sort)"
    policy_cmd "$kind" "$res"
    got="$(rd "${POLICY_CMD[@]}" | ours "$kind" "$res" | LC_ALL=C sort)"
    check "bindings $kind $res" "$want" "$got"
  done 3< <(
    printf 'project%s%s\n' "$SEP" "$project"
    for name in "$deployer_sa" "$vm_sa" "$stop_sa" "$read_sa"; do printf 'sa%s%s\n' "$SEP" "$name"; done
    printf 'repo%s%s\ninstance%s%s\nbucket%s%s\nbucket%s%s\n' "$SEP" "$repository" "$SEP" "$instance" "$SEP" "$meta_bucket" "$SEP" "$state_bucket"
    { printf '%s\n' "$cf_secret"; [[ -z "$secrets_list" ]] || printf '%s\n' "$secrets_list"; } | LC_ALL=C sort -u | sed "s/^/secret$SEP/"
  )
  if ((dry_run)); then receipt verify rendered dry-run; return 0; fi
  read_ok "reads" || true
  if ((VERIFY_BAD)) || [[ -s "$RD_FAIL" ]]; then receipt verify "$env_name" FAILED; exit 1; fi
  receipt verify "$env_name" ok
}

# ---- teardown ---------------------------------------------------------------------------------------------------------
teardown_all() {
  local kind res member role cond n first_meta
  preflight
  # Before any change: the bucket is ours (or a confirmed 404) and its live objects were listed (or confirmed none).
  meta_owner present
  first_meta=$META_STATE
  if ((dry_run == 0)) && [[ $META_STATE == present ]]; then
    meta_live
    n=$LIVE
    if ((n > 0)) && [[ "$confirm_meta" != "$env_name" ]]; then
      die "gs://$meta_bucket holds $n live metadata object(s); retire the environment first, then pass --confirm-meta $env_name"
    fi
  fi
  while IFS="$SEP" read -r -u 3 kind res member role cond; do
    [[ "$kind" != bucket ]] || continue # removed with the bucket below
    if ((dry_run == 0)); then
      case "$kind" in
        instance) present compute instances describe "$res" --zone="$zone" || { receipt binding "$member" "$role@$res absent (no instance)"; continue; } ;;
        secret) present secrets describe "$res" || { receipt binding "$member" "$role@$res absent (no secret)"; continue; } ;;
        sa) present iam service-accounts describe "$res" || { receipt binding "$member" "$role@$res absent (no account)"; continue; } ;;
      esac
    fi
    if has_binding "$kind" "$res" "$member" "$role"; then
      binding_cmd remove "$kind" "$res" "$member" "$role" "$cond"
      run g "${BINDING_CMD[@]}"; receipt binding "$member" "$role@$res removed"
    else receipt binding "$member" "$role@$res absent"; fi
  done 3< <(desired_bindings)
  for res in "$stop_sa" "$read_sa"; do
    if ((dry_run)) || exists iam service-accounts describe "$res"; then
      run g iam service-accounts delete "$res"; receipt service-account "$res" removed
    else receipt service-account "$res" absent; fi
  done
  for res in "$stop_role_id" "$MUTABLE_ROLE_ID"; do
    if ((dry_run)) || { exists iam roles describe "$res" --format='value(deleted)' && [[ "$G_OUT" != True ]]; }; then
      run g iam roles delete "$res" --format=none; receipt custom-role "projects/$project/roles/$res" removed
    else receipt custom-role "projects/$project/roles/$res" absent; fi
  done
  value_of iam workload-identity-pools providers describe "$PROVIDER" --workload-identity-pool="$pool" --location=global --format='value(state)'
  if ((dry_run)) || [[ "$VALUE" == ACTIVE ]]; then
    run g iam workload-identity-pools providers delete "$PROVIDER" --workload-identity-pool="$pool" --location=global --format=none
    receipt provider "$pool/$PROVIDER" removed
  else receipt provider "$pool/$PROVIDER" absent; fi
  value_of iam workload-identity-pools describe "$pool" --location=global --format='value(state)'
  if ((dry_run)) || [[ "$VALUE" == ACTIVE ]]; then
    run g iam workload-identity-pools delete "$pool" --location=global --format=none; receipt pool "$pool" removed
  else receipt pool "$pool" absent; fi
  # The owner again, right before the recursive deletion.
  meta_owner present
  # Absent at the start but present now: its live objects were never listed or confirmed, so it is never deleted.
  [[ $META_STATE != present || $first_meta == present ]] \
    || die "gs://$meta_bucket was absent when teardown started and is present now; its live objects were not checked; refused (rerun teardown)"
  if [[ $META_STATE == present ]]; then
    run g storage rm --recursive --all-versions "gs://$meta_bucket"; receipt meta-bucket "gs://$meta_bucket" removed
  else receipt meta-bucket "gs://$meta_bucket" absent; fi
  receipt api sts.googleapis.com "retained (APIs are never disabled)"
}

tmp_dir="$(mktemp -d)"
RD_FAIL="$tmp_dir/read-failures"
: >"$RD_FAIL"
if ((teardown)); then teardown_all
elif ((verify)); then verify_all
else setup; fi
