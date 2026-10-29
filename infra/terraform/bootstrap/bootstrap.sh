#!/usr/bin/env bash
# One-time bootstrap of a staging project (G070; infra/terraform/README.md). Run by the environment owner with
# their own short-lived ADC/gcloud login, never by an agent before the live release.
#
#   bootstrap.sh --project ID --region REGION --env NAME --operator MEMBER
#                [--repository NAME] [--state-bucket NAME] [--dry-run]
#   bootstrap.sh --teardown --confirm ID --project ID --region REGION --env NAME
#                [--repository NAME] [--state-bucket NAME] [--dry-run]
#
# Idempotent: every item is created only when absent, bindings are added (adding an existing binding is a no-op),
# and one receipt line per item says whether it already existed or was created. It enables the APIs; creates the
# Terraform state bucket (versioned, uniform access, public-access prevention, 90-day noncurrent-version expiry),
# the VM runtime and deployer service accounts with the bindings of the README's IAM table, one Docker repository
# in Artifact Registry, and the Secret Manager secrets <env>-<NAME>. The eight application secrets get one
# generated version (48 hexadecimal characters, as infra/local/configure-env.sh) only while they have no version
# at all; the value goes from /dev/urandom through a pipe to gcloud and is never printed, logged or passed as an
# argument. CF_API_TOKEN and OPENROUTER_API_KEY are entered by the owner; the IDENTITY_WORKOS_* secrets stay empty
# for their identity owner. Terraform never touches any of these items; they are the retained bootstrap set.
#
# --teardown removes this environment's bootstrap items (secrets, service accounts with their project bindings,
# the repository with its images) and the state bucket when no live object is left in it. It refuses while
# this environment's state object or any labelled VM, disk or address exists. --dry-run prints the commands
# without calling gcloud and treats every item as absent.
set -euo pipefail

GENERATED_SECRETS=(POSTGRES_SUPERUSER_PASSWORD DB_MIGRATOR_PASSWORD DB_RUNTIME_PASSWORD DB_RETENTION_PASSWORD
  REDIS_PASSWORD MINIO_ROOT_PASSWORD OBJECT_STORE_SECRET_KEY INTELLIGENCE_SERVICE_TOKEN)
OWNER_SECRETS=(CF_API_TOKEN OPENROUTER_API_KEY)
IDENTITY_SECRETS=(IDENTITY_WORKOS_CLIENT_ID IDENTITY_WORKOS_ISSUER IDENTITY_WORKOS_API_KEY)
APIS=(compute.googleapis.com iam.googleapis.com iamcredentials.googleapis.com iap.googleapis.com
  oslogin.googleapis.com storage.googleapis.com secretmanager.googleapis.com artifactregistry.googleapis.com
  cloudresourcemanager.googleapis.com)
DEPLOYER_PROJECT_ROLES=(roles/compute.instanceAdmin.v1 roles/compute.networkAdmin roles/compute.securityAdmin
  roles/iap.tunnelResourceAccessor roles/compute.osAdminLogin)

usage() { sed -n '5,8p' "$0" | sed 's/^# \{0,1\}//' >&2; exit 2; }
die() { printf 'bootstrap: ERROR: %s\n' "$*" >&2; exit 1; }
receipt() { printf 'bootstrap\t%s\t%s\t%s\n' "$1" "$2" "$3"; }

project="" region="" env_name="" operator="" repository="" bucket="" confirm="" teardown=0 dry_run=0
while (($#)); do
  case "$1" in
    --project) project="${2:?}"; shift ;;
    --region) region="${2:?}"; shift ;;
    --env) env_name="${2:?}"; shift ;;
    --operator) operator="${2:?}"; shift ;;
    --repository) repository="${2:?}"; shift ;;
    --state-bucket) bucket="${2:?}"; shift ;;
    --confirm) confirm="${2:?}"; shift ;;
    --teardown) teardown=1 ;;
    --dry-run) dry_run=1 ;;
    -h|--help) usage ;;
    *) die "unknown argument $1" ;;
  esac
  shift
done

[[ "$project" =~ ^[a-z][a-z0-9-]{4,28}[a-z0-9]$ ]] || die "--project must be a GCP project ID"
[[ "$region" =~ ^[a-z][a-z0-9-]+[0-9]$ ]] || die "--region must be a GCP region"
# Service-account IDs need 6-30 characters: <env>-vm and <env>-deployer fit for 3-21 character names.
[[ "$env_name" =~ ^[a-z][a-z0-9-]{2,20}$ ]] || die "--env must match ^[a-z][a-z0-9-]{2,20}\$"
repository="${repository:-$env_name-images}"
bucket="${bucket:-$project-tfstate}"
vm_sa="$env_name-vm@$project.iam.gserviceaccount.com"
deployer_sa="$env_name-deployer@$project.iam.gserviceaccount.com"
labels="environment=$env_name,managed-by=bootstrap"
if ((teardown)); then
  [[ "$confirm" == "$project" ]] || die "--teardown requires --confirm $project"
else
  [[ "$operator" =~ ^(user|group):[^[:space:]]+@[^[:space:]]+$ ]] || die "--operator must be a user: or group: member (granted impersonation of the deployer)"
fi

# Mutating call (`g ...` is gcloud on this project): printed in dry-run mode. Arguments never contain a secret.
run() {
  [[ "$1" != g ]] || { shift; set -- gcloud --project "$project" --quiet "$@"; }
  if ((dry_run)); then printf '+ %s\n' "$*"; return 0; fi
  "$@"
}
# Read-only existence check; dry-run treats the item as absent.
exists() {
  ((dry_run)) && return 1
  "$@" >/dev/null 2>&1
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
g() { gcloud --project "$project" --quiet "$@"; }
tmp_dir=""
trap '[[ -z "$tmp_dir" ]] || rm -rf "$tmp_dir"' EXIT

preflight() {
  ((dry_run)) && return 0
  command -v gcloud >/dev/null 2>&1 || die "gcloud not found"
  [[ -n "$(gcloud config get-value account 2>/dev/null)" ]] || die "no active gcloud account; run gcloud auth login"
  [[ "$(gcloud projects describe "$project" --format='value(projectId)' 2>/dev/null)" == "$project" ]] \
    || die "project $project is not accessible to the active account"
}

enable_apis() {
  local enabled="" api
  ((dry_run)) || enabled="$(g services list --enabled --format='value(config.name)')"
  for api in "${APIS[@]}"; do
    if grep -qxF "$api" <<<"$enabled"; then receipt api "$api" existing
    else run g services enable "$api" --format=none; receipt api "$api" enabled; fi
  done
}

state_bucket() {
  local lifecycle
  tmp_dir="$(mktemp -d)"
  lifecycle="$tmp_dir/lifecycle.json"
  if exists g storage buckets describe "gs://$bucket"; then
    receipt state-bucket "gs://$bucket" existing
  else
    run g storage buckets create "gs://$bucket" --location="$region" --uniform-bucket-level-access --public-access-prevention
    receipt state-bucket "gs://$bucket" created
  fi
  # Noncurrent state versions expire after 90 days; the live state object is never touched by this rule.
  printf '%s\n' '{"rule":[{"action":{"type":"Delete"},"condition":{"isLive":false,"daysSinceNoncurrentTime":90}}]}' >"$lifecycle"
  run g storage buckets update "gs://$bucket" --versioning --lifecycle-file="$lifecycle" --update-labels="managed-by=bootstrap"
  receipt state-bucket-settings "gs://$bucket" "versioning,uniform-access,public-access-prevention,noncurrent-90d"
}

service_account() {
  local email="$1" id="${1%%@*}" name="$2"
  if exists g iam service-accounts describe "$email"; then receipt service-account "$email" existing
  else run g iam service-accounts create "$id" --display-name="$name" --format=none; receipt service-account "$email" created; fi
}

artifact_repository() {
  if exists g artifacts repositories describe "$repository" --location="$region"; then
    receipt repository "$region-docker.pkg.dev/$project/$repository" existing
  else
    run g artifacts repositories create "$repository" --location="$region" --repository-format=docker \
      --description="Staging images for $env_name (digest-pinned deploys)" --labels="$labels" --format=none
    receipt repository "$region-docker.pkg.dev/$project/$repository" created
  fi
}

secret() {
  local name="$1" kind="$2" id
  id="$env_name-$name"
  if exists g secrets describe "$id"; then receipt secret "$id" existing
  else
    run g secrets create "$id" --replication-policy=user-managed --locations="$region" --labels="$labels" --format=none
    receipt secret "$id" created
  fi
  [[ "$kind" == generated ]] || { receipt secret-version "$id" "awaiting-$kind"; return 0; }
  if ((dry_run == 0)) && [[ -n "$(g secrets versions list "$id" --limit=1 --format='value(name)')" ]]; then
    receipt secret-version "$id" present
    return 0
  fi
  if ((dry_run)); then
    printf '+ <48 generated hexadecimal characters on stdin> | gcloud --project %s --quiet secrets versions add %s --data-file=-\n' "$project" "$id"
  else
    head -c 24 /dev/urandom | od -An -tx1 | tr -d ' \n' | g secrets versions add "$id" --data-file=- --format=none
  fi
  receipt secret-version "$id" generated
}

bind_all() {
  local role name
  for role in "${DEPLOYER_PROJECT_ROLES[@]}"; do
    retry g projects add-iam-policy-binding "$project" --member="serviceAccount:$deployer_sa" --role="$role" --condition=None --format=none
    receipt binding "$deployer_sa" "$role@project"
  done
  retry g iam service-accounts add-iam-policy-binding "$vm_sa" --member="serviceAccount:$deployer_sa" --role=roles/iam.serviceAccountUser --format=none
  receipt binding "$deployer_sa" "roles/iam.serviceAccountUser@$vm_sa"
  retry g storage buckets add-iam-policy-binding "gs://$bucket" --member="serviceAccount:$deployer_sa" --role=roles/storage.objectAdmin --format=none
  receipt binding "$deployer_sa" "roles/storage.objectAdmin@gs://$bucket"
  retry g artifacts repositories add-iam-policy-binding "$repository" --location="$region" --member="serviceAccount:$deployer_sa" --role=roles/artifactregistry.writer --format=none
  receipt binding "$deployer_sa" "roles/artifactregistry.writer@$repository"
  retry g artifacts repositories add-iam-policy-binding "$repository" --location="$region" --member="serviceAccount:$vm_sa" --role=roles/artifactregistry.reader --format=none
  receipt binding "$vm_sa" "roles/artifactregistry.reader@$repository"
  for name in "${GENERATED_SECRETS[@]}" "${OWNER_SECRETS[@]}" "${IDENTITY_SECRETS[@]}"; do
    retry g secrets add-iam-policy-binding "$env_name-$name" --member="serviceAccount:$vm_sa" --role=roles/secretmanager.secretAccessor --format=none
    receipt binding "$vm_sa" "roles/secretmanager.secretAccessor@$env_name-$name"
  done
  retry g iam service-accounts add-iam-policy-binding "$deployer_sa" --member="$operator" --role=roles/iam.serviceAccountTokenCreator --format=none
  receipt binding "$operator" "roles/iam.serviceAccountTokenCreator@$deployer_sa"
}

bootstrap() {
  local name
  preflight
  enable_apis
  state_bucket
  service_account "$vm_sa" "Staging $env_name VM runtime"
  service_account "$deployer_sa" "Staging $env_name deployer"
  artifact_repository
  for name in "${GENERATED_SECRETS[@]}"; do secret "$name" generated; done
  secret CF_API_TOKEN owner
  secret OPENROUTER_API_KEY owner-optional
  for name in "${IDENTITY_SECRETS[@]}"; do secret "$name" identity-owner; done
  bind_all
  receipt "done" "$env_name" "project=$project region=$region state=gs://$bucket/staging/$env_name repository=$region-docker.pkg.dev/$project/$repository"
}

teardown() {
  local name role kind left
  preflight
  if ((dry_run == 0)); then
    left="$(g storage ls "gs://$bucket/staging/$env_name/" 2>/dev/null || true)"
    [[ -z "$left" ]] || die "state objects remain under gs://$bucket/staging/$env_name/; retire the environment and remove its state object first"
    for kind in instances disks addresses; do
      left="$(g compute "$kind" list --filter="labels.environment=$env_name" --format='value(name)')"
      [[ -z "$left" ]] || die "labelled $kind remain for $env_name: ${left//$'\n'/ }; retire the environment first"
    done
  fi
  for name in "${GENERATED_SECRETS[@]}" "${OWNER_SECRETS[@]}" "${IDENTITY_SECRETS[@]}"; do
    if ((dry_run)) || exists g secrets describe "$env_name-$name"; then
      run g secrets delete "$env_name-$name"; receipt secret "$env_name-$name" removed
    else receipt secret "$env_name-$name" absent; fi
  done
  for role in "${DEPLOYER_PROJECT_ROLES[@]}"; do
    if ((dry_run)) || [[ -n "$(g projects get-iam-policy "$project" --flatten='bindings[].members' \
      --filter="bindings.role=$role AND bindings.members=serviceAccount:$deployer_sa" --format='value(bindings.role)')" ]]; then
      run g projects remove-iam-policy-binding "$project" --member="serviceAccount:$deployer_sa" --role="$role" --condition=None --format=none
      receipt binding "$deployer_sa" "$role@project removed"
    fi
  done
  for name in "$deployer_sa" "$vm_sa"; do
    if ((dry_run)) || exists g iam service-accounts describe "$name"; then
      run g iam service-accounts delete "$name"; receipt service-account "$name" removed
    else receipt service-account "$name" absent; fi
  done
  if ((dry_run)) || exists g artifacts repositories describe "$repository" --location="$region"; then
    run g artifacts repositories delete "$repository" --location="$region"; receipt repository "$repository" removed
  else receipt repository "$repository" absent; fi
  if ((dry_run == 0)) && ! exists g storage buckets describe "gs://$bucket"; then
    receipt state-bucket "gs://$bucket" absent
  elif ((dry_run == 0)) && [[ -n "$(g storage ls "gs://$bucket/**" 2>/dev/null || true)" ]]; then
    receipt state-bucket "gs://$bucket" "retained (live objects of other environments)"
  else
    run g storage rm --recursive --all-versions "gs://$bucket"; receipt state-bucket "gs://$bucket" removed
  fi
}

if ((teardown)); then teardown; else bootstrap; fi
