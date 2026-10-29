# shellcheck shell=bash
# Shared fixture of tests/env-lease.sh and tests/ci-adapter.sh (G072): a throwaway Git repository holding the trusted
# control copy of infra/staging plus synthetic deployment files and migrations, a staging.conf for it, the fake tool
# state (fakelib.sh) and the case ledger. Sourced; no network, Docker or cloud; everything lives under WORK.
#
# Ledger: REPORT/summary.tsv (case, result, detail); the last stdout line is "RESULT PASS|FAIL cases=N failed=M"
# (the form check.sh rows grep for). Mutation controls are cases of their own named control-*: each runs the protected
# assertion against a deliberately broken copy and passes only when the assertion catches the break.

FX_SRC="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)" # infra/staging of the checkout under test
FX_PROJECT="$(cd "$FX_SRC/../.." && pwd -P)"
FAKES="$FX_SRC/tests/fakes"
FX_ENV=staging
FX_PROJECT_ID=example-project
FX_AR=us-central1-docker.pkg.dev/example-project/staging
FX_BUCKET=example-project-staging-ci-meta
FX_NOW=1800000000
FX_CF_TOKEN="cfut_G072FAKECFTOKENCANARY000000000000000000000000000000"
FX_SERVICES=(core web intelligence minio minio-init migrate edge)
CASES=0 FAILED=0 ONLY=""
# Fixed commit dates: the fixture's commit SHAs depend only on the tree under test, so a run is repeatable (the SHAs
# once decided an intermittent failure; see ci-adapter.sh class-meta).
export GIT_AUTHOR_DATE="2026-01-01T00:00:00Z" GIT_COMMITTER_DATE="2026-01-01T00:00:00Z"

fx_usage() { echo "usage: $0 --work DIR --report DIR [--only NAME[,NAME]...]" >&2; exit 64; }

fx_args() {
  WORK="" REPORT=""
  while (($#)); do
    case $1 in
      --work) (($# >= 2)) || fx_usage; WORK=$2; shift ;;
      --report) (($# >= 2)) || fx_usage; REPORT=$2; shift ;;
      --only) (($# >= 2)) || fx_usage; ONLY=$2; shift ;;
      *) fx_usage ;;
    esac
    shift
  done
  [[ -n $WORK && -n $REPORT ]] || fx_usage
  mkdir -p "$WORK" "$REPORT"
  WORK="$(cd "$WORK" && pwd -P)" REPORT="$(cd "$REPORT" && pwd -P)"
  [[ ! -e $REPORT/summary.tsv ]] || { echo "$REPORT/summary.tsv exists; use a new report directory" >&2; exit 64; }
  printf 'case\tresult\tdetail\n' >"$REPORT/summary.tsv"
  W="$(mktemp -d "$WORK/run.XXXXXX")"
  trap 'rm -rf "$W"' EXIT
}

# selected CASE: --only names cases or groups (the part before the first "-": --only fence runs every fence-* case).
selected() { [[ -z $ONLY || ",$ONLY," == *",$1,"* || ",$ONLY," == *",${1%%-*},"* ]]; }

# record CASE OK DETAIL
record() {
  CASES=$((CASES + 1))
  local r=PASS
  ((${2})) || { r=FAIL; FAILED=$((FAILED + 1)); }
  printf '%s\t%s\t%s\n' "$1" "$r" "${3//$'\n'/ }" >>"$REPORT/summary.tsv"
  printf '%s %s %s\n' "$r" "$1" "${3//$'\n'/ }"
}

# control NAME MUTATION WHY: a mutation control passes only when the protected assertion reported a failure (WHY).
control() {
  if [[ -n $3 ]]; then record "$1" 1 "mutation: $2; the assertion catches it (${3:0:160})"
  else record "$1" 0 "mutation: $2; the assertion DID NOT catch it"; fi
}

fx_finish() {
  local r=PASS
  ((FAILED == 0 && CASES > 0)) || r=FAIL
  printf 'RESULT %s cases=%s failed=%s\n' "$r" "$CASES" "$FAILED"
  [[ $r == PASS ]]
}

# fx_repo DIR: the fixture repository at DIR: synthetic deployment files, migrations 1..24 and the control copy of
# infra/staging (the scripts under test), committed on branch staging. FX_BASE = that commit.
fx_repo() {
  local d=$1 n f
  mkdir -p "$d"
  git -C "$d" init -q -b staging
  git -C "$d" config user.email fixture@example.net
  git -C "$d" config user.name fixture
  git -C "$d" config commit.gpgsign false
  mkdir -p "$d/infra/compose" "$d/infra/postgres" "$d/infra/minio" "$d/infra/edge" "$d/infra/docker" "$d/infra/staging/remote" "$d/core/migrations"
  printf 'COMPOSE_PROJECT_NAME=x\nDATABASE_NAME=\n' >"$d/.env.example"
  for f in local migrate staging; do printf 'services: {}\n' >"$d/infra/compose/$f.yml"; done
  printf 'init\n' >"$d/infra/postgres/init.sql"; printf 'init\n' >"$d/infra/minio/init.sh"; printf ':80\n' >"$d/infra/edge/Caddyfile"
  for f in core web intelligence minio migrate edge; do printf 'FROM scratch\n' >"$d/infra/docker/$f.Dockerfile"; done
  for f in deploy start stop status reset; do printf '#!/bin/sh\n' >"$d/infra/staging/remote/$f.sh"; done
  printf -- '-- seed\n' >"$d/infra/staging/seed.sql"
  for ((n = 1; n <= 24; n++)); do printf -- '-- migration %s\n' "$n" >"$d/core/migrations/$(printf '%05d' "$n")_m$n.sql"; done
  printf 'package migrations\n' >"$d/core/migrations/embed.go"
  printf 'app\n' >"$d/core/app.txt"
  fx_install_scripts "$d"
  git -C "$d" add -A
  git -C "$d" commit -q -m base
  FX_BASE="$(git -C "$d" rev-parse HEAD)"
}

# fx_install_scripts DIR: copy the G072 scripts under test into DIR/infra/staging (control code).
fx_install_scripts() {
  local f
  for f in staging.sh lib.sh meta-lib.sh ci.sh ci-lib.sh ci-environments.txt; do
    [[ ! -f $FX_SRC/$f ]] || cp "$FX_SRC/$f" "$1/infra/staging/$f"
  done
  # W3's publication functions (shared contract 3): the real file once integrated, the stub until then.
  if [[ -n ${FX_PUBLISH_FILE:-} ]]; then cp "$FX_PUBLISH_FILE" "$1/infra/staging/ci-publish.sh"
  elif [[ -f $FX_SRC/ci-publish.sh ]]; then cp "$FX_SRC/ci-publish.sh" "$1/infra/staging/ci-publish.sh"
  else cp "$FAKES/ci-publish-stub.sh" "$1/infra/staging/ci-publish.sh"; fi
}

# fx_conf FILE EVIDENCE PRIVATE [KEY=VALUE...]: a staging.conf for the fixture environment.
fx_conf() {
  local f=$1 ev=$2 pr=$3
  shift 3
  mkdir -p "$pr" "$ev"
  : >"$pr/adc.json"
  printf 'bucket = "%s-tfstate"\nprefix = "staging/%s"\n' "$FX_PROJECT_ID" "$FX_ENV" >"$pr/backend.hcl"
  printf 'project_id = "%s"\n' "$FX_PROJECT_ID" >"$pr/terraform.tfvars"
  {
    printf '%s\n' "STAGING_ENV=$FX_ENV" "GCP_PROJECT=$FX_PROJECT_ID" GCP_REGION=us-central1 GCP_ZONE=us-central1-a \
      "DEPLOYER_SA=$FX_ENV-deployer@$FX_PROJECT_ID.iam.gserviceaccount.com" "AR_REPOSITORY=$FX_AR" \
      BASE_DOMAIN=staging.example.net CLOUDFLARE_ZONE_ID=00000000000000000000000000000000 ACME_EMAIL=ops@example.net \
      EDGE_ACME_CA=https://acme-v02.api.letsencrypt.org/directory EDGE_DATA_VOLUME=edge-data "PRODUCT_NAME=Example Product" \
      DATABASE_NAME=example_app OBJECT_STORE_BUCKET=example-objects \
      "TERRAFORM_IMAGE=hashicorp/terraform:1.16.7@sha256:$(printf '0%.0s' {1..64})" "EVIDENCE_DIR=$ev" "PRIVATE_DIR=$pr" \
      "GOOGLE_ADC_FILE=$pr/adc.json" "OWNER=fixture owner"
    (($# == 0)) || printf '%s\n' "$@"
  } >"$f"
}

# fx_manifest FILE SHA: a pushed G070 deploy manifest of SHA (digests registered in the fake registry).
fx_manifest() {
  local f=$1 c=$2 svc id
  mkdir -p "$(dirname "$f")" "$FAKE_ROOT/ar"
  {
    printf 'candidate\t%s\n' "$c"
    for svc in "${FX_SERVICES[@]}"; do
      id="sha256:$(printf '%s-%s' "$svc" "$c" | sha256sum | cut -d' ' -f1)"
      printf 'image\t%s\t%s\t%s\t%s\t%s/%s@%s\n' "$svc" "$c" "$(printf '%s' "$svc" | sha256sum | cut -d' ' -f1)" "$id" "$FX_AR" "$svc" \
        "sha256:$(printf '%s' "$id" | sha256sum | cut -d' ' -f1)"
    done
    printf 'file\tinfra/compose/local.yml\t%s\n' "$(printf 'x' | sha256sum | cut -d' ' -f1)"
  } >"$f"
  awk -F'\t' '$1 == "image" {print $6}' "$f" >>"$FAKE_ROOT/ar/digests"
}

# fx_state [absent]: a fresh fake world: the metadata bucket (unless "absent"), a RUNNING VM with a 7200 s session and
# an hour left, Terraform agreeing, the permission model, the Cloudflare canary.
fx_state() {
  rm -rf "$FAKE_ROOT"
  mkdir -p "$FAKE_ROOT/vm" "$FAKE_ROOT/tf" "$FAKE_ROOT/ar" "$FAKE_ROOT/gcs"
  : >"$FAKE_ROOT/calls.log"
  [[ ${1:-} == absent ]] || mkdir -p "$FAKE_ROOT/gcs/$FX_BUCKET"
  cp "$FX_PERMS" "$FAKE_ROOT/perms.tsv"
  printf '%s\n' RUNNING >"$FAKE_ROOT/vm/status"
  printf '7200\n' >"$FAKE_ROOT/vm/max_run"
  printf '%s\n' "$((FX_NOW + 3600))" >"$FAKE_ROOT/vm/term"
  printf '7200\n' >"$FAKE_ROOT/tf/session"
  printf '%s\n' staging-vm staging-data staging-ip staging-edge staging-iap staging-vpc staging-subnet >"$FAKE_ROOT/tf/state_names"
  printf '2\n' >"$FAKE_ROOT/tf/plan_rc"
  printf '%s\n' "$FX_CF_TOKEN" >"$FAKE_ROOT/cf-token"
  printf '%s\n' "$FX_NOW" >"$FAKE_ROOT/now"
}

# The IAM model: W1's single source when it is in the tree, the contract-format fixture copy until then.
# FX_PERMS_FILE (environment) names another copy of W1's file, e.g. extracted from its branch before integration.
fx_perms_source() {
  if [[ -n ${FX_PERMS_FILE:-} ]]; then
    printf '%s' "$FX_PERMS_FILE"
  elif [[ -f $FX_PROJECT/infra/terraform/bootstrap/ci-meta-permissions.tsv ]]; then
    printf '%s' "$FX_PROJECT/infra/terraform/bootstrap/ci-meta-permissions.tsv"
  else
    printf '%s' "$FAKES/ci-meta-permissions.tsv"
  fi
}

# fx_put KEY FILE [PRINCIPAL]: place an object in the fake metadata bucket directly (test setup, not a logged call).
fx_put() {
  local d n
  d="$FAKE_ROOT/gcs/$FX_BUCKET/obj/${1//\//%2F}"
  mkdir -p "$d"
  n=$(($(cat "$FAKE_ROOT/gcs/.counter" 2>/dev/null || echo 1000) + 1))
  printf '%s\n' "$n" >"$FAKE_ROOT/gcs/.counter"
  cp "$2" "$d/v$n"
  printf '%s\n' "$n" >"$d/gen"
}
fx_gen() { local d="$FAKE_ROOT/gcs/$FX_BUCKET/obj/${1//\//%2F}"; [[ -f $d/gen ]] && cat "$d/gen" || echo 0; }
fx_cat() { local d="$FAKE_ROOT/gcs/$FX_BUCKET/obj/${1//\//%2F}" g; g="$(fx_gen "$1")"; [[ $g != 0 ]] && cat "$d/v$g"; }
has_key() { local k; k="$(fx_keys)"; grep -qE -- "$1" <<<"$k"; } # no pipe into grep -q (pipefail and SIGPIPE)
fx_keys() { local d; for d in "$FAKE_ROOT/gcs/$FX_BUCKET/obj"/*; do [[ -d $d && "$(cat "$d/gen")" != 0 ]] && { d=${d##*/}; printf '%s\n' "${d//%2F//}"; }; done; return 0; }

# fx_lease HOLDER EXPIRES_EPOCH|none: seed a lease object.
fx_lease() {
  printf 'holder\t%s\ncommand\tseeded\ncandidate\t-\nacquired_epoch\t%s\nexpires_epoch\t%s\nnonce\tseed\n' "$1" "$((FX_NOW - 60))" "$2" >"$W/lease.seed"
  fx_put lease "$W/lease.seed"
}

# fx_run NAME VAR=VALUE... -- CMD...: run CMD with only the fakes and the given variables in its environment;
# NAME.out, NAME.err, NAME.rc under W/runs. RC = the exit code.
fx_run() {
  local name=$1 vars=()
  shift
  while (($#)) && [[ $1 != -- ]]; do vars+=("$1"); shift; done
  shift
  mkdir -p "$W/runs" "$W/tmp" "$W/home"
  RC=0
  env -i PATH="$FAKES:/usr/local/bin:/usr/bin:/bin" HOME="$W/home" TMPDIR="$W/tmp" LC_ALL=C FAKE_ROOT="$FAKE_ROOT" DOCKER_FAKE_ROOT="$FAKE_ROOT" \
    FAKE_REPO="${FAKE_REPO:-}" CLOUDSDK_CONFIG="$W/home/gcloud" "${vars[@]}" "$@" \
    >"$W/runs/$name.out" 2>"$W/runs/$name.err" </dev/null || RC=$?
  printf '%s\n' "$RC" >"$W/runs/$name.rc"
  OUT="$W/runs/$name.out" ERR="$W/runs/$name.err"
}

# Call-log views: every call, writes (class W), and calls other than the metadata (gcloud storage) ones.
calls_all() { cat "$FAKE_ROOT/calls.log"; }
calls_writes() { awk -F'\t' '$2 == "W"' "$FAKE_ROOT/calls.log"; }
calls_cloud() { awk -F'\t' '$1 == "gcloud" || $1 == "curl"' "$FAKE_ROOT/calls.log"; }
calls_nonmeta() { awk -F'\t' '!($1 == "gcloud" && $3 ~ /^(storage|projects describe) /)' "$FAKE_ROOT/calls.log"; }
nlines() { grep -c . || true; }
keep() { mkdir -p "$REPORT/$1"; cp "$OUT" "$REPORT/$1/${2:-run}.out"; cp "$ERR" "$REPORT/$1/${2:-run}.err"; cp "$FAKE_ROOT/calls.log" "$REPORT/$1/${2:-run}.calls"; }
