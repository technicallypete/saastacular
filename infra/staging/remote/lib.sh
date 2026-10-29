# shellcheck shell=bash
# VM-side library of infra/staging/remote/{deploy,start,stop,reset,status}.sh (docs/foundation/STAGING.md). Sourced,
# never executed. The scripts run as root on the staging VM, invoked by infra/staging/staging.sh over OS Login + IAP;
# with --dry-run they print the VM commands ("[vm] + ...") anywhere, without Docker, metadata or Secret Manager.
#
# Layout on the VM (one environment <env> per VM):
#   /srv/<env>/releases/<sha12>/  the candidate's infra/{compose,postgres,minio,edge,staging} subtree (git archive)
#   /srv/<env>/current            symlink to the release of the latest deploy
#   /srv/<env>/meta/              remote.env (non-secret values from staging.conf) and manifest.tsv (image digests)
#   /srv/<env>/state/             secret-versions.tsv (secret names and pinned version numbers, never values), holds
#   /run/<env>/                   tmpfs only (refused otherwise): app.env (0600) while one script run lasts
#
# Secrets. Every Compose invocation needs the full interpolation, so each script run renders /run/<env>/app.env from
# remote.env, the manifest's IMAGE_* digests and the Secret Manager versions <env>-<NAME> (deploy: latest unless
# pinned; every other script: exactly the versions the latest deploy pinned in state/secret-versions.tsv), fetched with
# the VM service account's metadata token, and deletes it when the run ends. Docker nevertheless keeps each container's
# injected environment in the container configuration under /var/lib/docker, on the retained data disk (encrypted at
# rest by Compute Engine), for as long as the container exists, stopped containers included: deleting app.env revokes
# nothing. A value leaves the VM only with its container (deploy after a rotation, reset) or the disk (retire).
# Disabling a Secret Manager version stops new fetches only. CF_API_TOKEN must have the Cloudflare API-token format
# (40 characters of A-Z a-z 0-9 _ -, or the cfut_ or cfat_ prefix plus 48 of A-Z a-z 0-9, the only prefixes the pinned
# caddy-dns/cloudflare accepts) before anything is rendered: the edge's DNS plugin repeats a malformed token in
# its provisioning error, which no access-log filter redacts. Evidence never includes container configuration: every
# inspection carries a --format naming non-environment fields; no rendered Compose model, no /proc/*/environ.
#
# Access boundary. The file, the running processes' environment and the container configuration are readable by root
# and the docker group on this VM. Interactive access is OS Login over IAP (no SSH keys, no external port 22) for the
# configured operator principals; project and platform administrators can also reach the data (they can grant
# themselves that access, or snapshot/attach the disk) - the boundary is the staging project's administrators, not
# only the Secret Manager readers. The metadata token is a property of the VM, not of a process: the startup
# script's staging-metadata-guard drops every container network's traffic to the metadata server (host traffic is
# untouched), and metadata_denied proves that from each of this project's networks before app and edge are reported
# ready; a reachable metadata server HOLDs.
#
# Readiness. HTTPS is proven per name: B and a random label host each answer through the edge with a verified chain,
# B's certificate lists B and the label's lists *.B among its subjectAltNames, each valid for 7 more days (the two
# may be distinct certificates).

SECRETS_REQUIRED=(POSTGRES_SUPERUSER_PASSWORD DB_MIGRATOR_PASSWORD DB_RUNTIME_PASSWORD DB_RETENTION_PASSWORD REDIS_PASSWORD
  MINIO_ROOT_PASSWORD OBJECT_STORE_SECRET_KEY INTELLIGENCE_SERVICE_TOKEN CF_API_TOKEN)
SECRETS_OPTIONAL=(OPENROUTER_API_KEY IDENTITY_WORKOS_CLIENT_ID IDENTITY_WORKOS_ISSUER IDENTITY_WORKOS_API_KEY)
IDENTITY_SECRETS=(IDENTITY_WORKOS_CLIENT_ID IDENTITY_WORKOS_ISSUER IDENTITY_WORKOS_API_KEY)
# remote.env: every key below, nothing else. STAGING_ENV, GCP_PROJECT and REGISTRY_HOST are metadata of this library;
# the rest is the non-secret part of the Compose variable contract.
REMOTE_ENV_KEYS=(STAGING_ENV GCP_PROJECT REGISTRY_HOST COMPOSE_PROJECT_NAME PRODUCT_NAME PLATFORM_BASE_DOMAIN
  CUSTOM_DOMAIN_CNAME_TARGET PLATFORM_AUTH_HOST DATABASE_NAME APP_ENV BIND_ADDRESS WEB_PORT CORE_PORT INTELLIGENCE_PORT
  POSTGRES_PORT REDIS_PORT MINIO_API_PORT MINIO_CONSOLE_PORT CORE_INTERNAL_URL INTELLIGENCE_URL REDIS_HOST
  REDIS_PORT_INTERNAL OBJECT_STORE_ENDPOINT OBJECT_STORE_REGION OBJECT_STORE_BUCKET OBJECT_STORE_FORCE_PATH_STYLE
  OBJECT_STORE_ACCESS_KEY MINIO_ROOT_USER EDGE_BASE_DOMAIN ACME_EMAIL EDGE_ACME_CA EDGE_DATA_VOLUME
  DEFAULT_TENANT_SLUG PLATFORM_ROOT_PAGE MCP_CONNECT_ISSUER MCP_ALLOWED_ORIGINS)
META_KEYS=" STAGING_ENV GCP_PROJECT REGISTRY_HOST "
# Optional non-secret settings (G071, DOMAINS.md "Default-tenant root alias and root presentation"): empty or absent
# means unset; core applies its own defaults. Every other key is required and non-empty.
OPTIONAL_KEYS=" DEFAULT_TENANT_SLUG PLATFORM_ROOT_PAGE MCP_CONNECT_ISSUER MCP_ALLOWED_ORIGINS "
IMAGE_SERVICES=(core web intelligence minio minio-init migrate edge)
# The dataset marker Organization of infra/staging/seed.sql.
MARKER_ID=5ea70000-0000-4000-8000-000000000001
LE_STAGING_CA=https://acme-staging-v02.api.letsencrypt.org/directory

DRY_RUN=0
ENV_NAME=""
SRV=""
RUN_DIR=""
ENV_FILE=""
REL=""
SM_TOKEN=""
SM_VERSION=""
SM_VALUE=""
DOCKER_CFG=""
PROBE_RUN="" PROBE_SEQ=0 # metadata_denied: this run's probe-container identity and counter (probe_reap removes by label)
PENDING=0
MARKER=""
CID=""
REHEARSAL=0
declare -A RCONF=() IMAGES=() SECRET_VERSION=() SECRET_VALUE=() PIN=()

ts() { date -u +%Y-%m-%dT%H:%M:%SZ; }
# say STEP DETAIL: one line of the observed sequence (the operator keeps this output as the run's log).
say() { printf 'step\t%s\t%s\t%s\n' "$(ts)" "$1" "${2:-}"; }
die() { local code=$1; shift; printf 'staging-vm: %s\n' "$*" >&2; exit "$code"; }

# forbid ARGS...: the verbs this tooling never runs - volume-removing down, migrate down/reapply, prune, create - and
# volume removal outside this environment's two dataset volumes.
forbid() {
  local line=" $* " re vols v
  # `--profile migrate` names a Compose profile, not the migrate command: `--profile migrate down` is a plain `down`.
  line=${line// --profile / --profile=}
  for re in ' down( [^ ]+)* (-v|--volumes)( |$)' ' migrate (down|reapply)( |$)' \
    ' (system|volume|image|container|network|builder) prune( |$)' ' create( |$)'; do
    [[ ! $line =~ $re ]] || die 2 "refused a forbidden command (${BASH_REMATCH[0]# })"
  done
  re=' volume rm (.*)$'
  if [[ $line =~ $re ]]; then
    read -ra vols <<<"${BASH_REMATCH[1]}"
    for v in "${vols[@]}"; do
      [[ $v == "${ENV_NAME}_postgres-data" || $v == "${ENV_NAME}_minio-data" ]] || die 2 "refused: volume removal outside the dataset volumes"
    done
  fi
}

# run CMD...: refuse forbidden verbs; print in dry-run, execute otherwise.
run() {
  forbid "$@"
  if ((DRY_RUN)); then
    printf '[vm] +'
    printf ' %q' "$@"
    printf '\n'
    return 0
  fi
  "$@"
}

# capture VAR CMD...: VAR = the output of CMD (dry-run: print CMD, VAR empty).
capture() {
  local __var=$1 __out
  shift
  if ((DRY_RUN)); then
    # A wrapper such as dc prints its own expanded command through run; anything else is printed as given.
    if [[ $(type -t "$1") == function ]]; then "$@"; else run "$@"; fi
    printf -v "$__var" '%s' ""
    return 0
  fi
  forbid "$@"
  __out="$("$@")" || return $?
  printf -v "$__var" '%s' "$__out"
}

# dc ARGS...: one Compose invocation of this environment with the run's private env file.
dc() {
  run docker compose --env-file "$ENV_FILE" -f "$REL/infra/compose/local.yml" -f "$REL/infra/compose/migrate.yml" \
    -f "$REL/infra/compose/staging.yml" --profile migrate "$@"
}

cleanup() {
  ((DRY_RUN)) && return 0
  [[ -z $PROBE_RUN ]] || probe_reap || true
  rm -f "$ENV_FILE" 2>/dev/null || true
  rm -f "$RUN_DIR"/sm.* 2>/dev/null || true
  [[ -z $DOCKER_CFG ]] || rm -rf "$DOCKER_CFG" 2>/dev/null || true
}

# remote_init ENV: validate the environment name, set its paths and install the cleanup trap.
remote_init() {
  [[ $1 =~ ^[a-z][a-z0-9-]{1,20}$ ]] || die 2 "invalid --env"
  ENV_NAME=$1
  SRV=/srv/$1
  RUN_DIR=/run/$1
  ENV_FILE=/run/$1/app.env
  trap cleanup EXIT
  trap 'exit 130' INT TERM HUP
}

md() { curl -fsS --max-time 5 -H 'Metadata-Flavor: Google' "http://metadata.google.internal/computeMetadata/v1/$1"; }

# vm_preflight: root, this VM is <env>-vm of the configured project, /run/<env> is tmpfs (Plan r1.2).
vm_preflight() {
  if ((DRY_RUN)); then
    printf '[vm] (preflight: root; metadata project-id = GCP_PROJECT and instance/name = %s-vm; findmnt -no FSTYPE --target %s = tmpfs)\n' "$ENV_NAME" "$RUN_DIR"
    return 0
  fi
  [[ $EUID == 0 ]] || die 2 "run as root (sudo)"
  local project name fstype
  project="$(md project/project-id)" || die 2 "metadata server unavailable"
  name="$(md instance/name)" || die 2 "metadata server unavailable"
  [[ $project == "${RCONF[GCP_PROJECT]}" && $name == "$ENV_NAME-vm" ]] || die 2 "this VM ($name in $project) is not $ENV_NAME-vm in ${RCONF[GCP_PROJECT]}"
  install -d -m 0700 "$RUN_DIR"
  fstype="$(findmnt -no FSTYPE --target "$RUN_DIR")"
  [[ $fstype == tmpfs ]] || die 2 "$RUN_DIR is on $fstype, not tmpfs: refusing to write the private env file"
  say preflight "vm=$name project=$project run_dir_fstype=$fstype"
}

# mcp_settings_valid ISSUER ORIGINS (identical to infra/staging/lib.sh; this library runs on the VM without it): the optional MCP settings (MCP.md), both empty or well formed; sets MCP_SETTINGS_ERR to the
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

# load_meta DIR: DIR/remote.env into RCONF and DIR/manifest.tsv into IMAGES, strictly (dry-run: placeholders).
load_meta() {
  local dir=$1 line k v n=0 kind svc cand="" c ref var
  if ((DRY_RUN)); then
    for k in "${REMOTE_ENV_KEYS[@]}"; do RCONF[$k]="<$k>"; done
    RCONF[DEFAULT_TENANT_SLUG]="" RCONF[PLATFORM_ROOT_PAGE]="" RCONF[MCP_CONNECT_ISSUER]="" RCONF[MCP_ALLOWED_ORIGINS]=""
    RCONF[STAGING_ENV]=$ENV_NAME RCONF[COMPOSE_PROJECT_NAME]=$ENV_NAME RCONF[APP_ENV]=development RCONF[CANDIDATE]="<candidate>"
    for svc in "${IMAGE_SERVICES[@]}"; do var=${svc^^}; IMAGES[IMAGE_${var//-/_}]="<registry>/$svc@sha256:<digest>"; done
    printf '[vm] (reads %s/remote.env and %s/manifest.tsv)\n' "$dir" "$dir"
    return 0
  fi
  [[ -f $dir/remote.env && -f $dir/manifest.tsv ]] || die 2 "$dir lacks remote.env or manifest.tsv: run staging.sh deploy"
  while IFS= read -r line || [[ -n $line ]]; do
    n=$((n + 1))
    [[ -z $line || $line == \#* ]] && continue
    [[ $line =~ ^([A-Z][A-Z0-9_]*)=(.*)$ ]] || die 2 "remote.env line $n is not KEY=VALUE"
    k=${BASH_REMATCH[1]} v=${BASH_REMATCH[2]}
    [[ " ${REMOTE_ENV_KEYS[*]} " == *" $k "* ]] || die 2 "remote.env: unknown key $k"
    [[ (-n $v || $OPTIONAL_KEYS == *" $k "*) && $v != *"'"* ]] || die 2 "remote.env: $k is empty or holds a single quote"
    RCONF[$k]=$v
  done <"$dir/remote.env"
  for k in "${REMOTE_ENV_KEYS[@]}"; do
    if [[ $OPTIONAL_KEYS == *" $k "* ]]; then RCONF[$k]=${RCONF[$k]:-}; continue; fi
    [[ -n ${RCONF[$k]:-} ]] || die 2 "remote.env: $k is missing"
  done
  [[ -z ${RCONF[DEFAULT_TENANT_SLUG]} || ${RCONF[DEFAULT_TENANT_SLUG]} =~ ^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$ ]] ||
    die 2 "remote.env: DEFAULT_TENANT_SLUG is not a slug"
  [[ ${RCONF[PLATFORM_ROOT_PAGE]} =~ ^(signup_redirect|landing)?$ ]] || die 2 "remote.env: PLATFORM_ROOT_PAGE must be empty, signup_redirect or landing"
  mcp_settings_valid "${RCONF[MCP_CONNECT_ISSUER]}" "${RCONF[MCP_ALLOWED_ORIGINS]}" || die 2 "remote.env: $MCP_SETTINGS_ERR"
  [[ ${RCONF[STAGING_ENV]} == "$ENV_NAME" && ${RCONF[COMPOSE_PROJECT_NAME]} == "$ENV_NAME" ]] || die 2 "remote.env names another environment"
  [[ ${RCONF[APP_ENV]} == development ]] || die 2 "remote.env: APP_ENV must be development (synthetic staging profile)"
  while IFS=$'\t' read -r kind svc c _ _ ref; do
    case $kind in
      candidate) cand=$svc ;;
      image)
        [[ " ${IMAGE_SERVICES[*]} " == *" $svc "* ]] || die 2 "manifest: unknown service $svc"
        [[ -n $cand && $c == "$cand" ]] || die 2 "manifest: $svc was built from another candidate than the manifest's"
        [[ $ref == "${RCONF[REGISTRY_HOST]}/"* && ${ref#"${RCONF[REGISTRY_HOST]}/"} =~ ^[a-z0-9._/-]+@sha256:[0-9a-f]{64}$ ]] ||
          die 2 "manifest: $svc has no registry digest under ${RCONF[REGISTRY_HOST]}"
        var=${svc^^}
        IMAGES[IMAGE_${var//-/_}]=$ref
        ;;
    esac
  done <"$dir/manifest.tsv"
  [[ $cand =~ ^[0-9a-f]{40}$ ]] || die 2 "manifest: no candidate"
  for svc in "${IMAGE_SERVICES[@]}"; do
    var=${svc^^}
    [[ -n ${IMAGES[IMAGE_${var//-/_}]:-} ]] || die 2 "manifest: no image for $svc"
  done
  RCONF[CANDIDATE]=$cand
}

# sm_access NAME VERSION: fetch one Secret Manager version into SM_VERSION/SM_VALUE. Returns 1 when it does not exist
# or has no enabled version (404, 400 FAILED_PRECONDITION), 2 on any other failure. Values and the token never reach
# argv, a log or the disk (the response file lives on tmpfs and is deleted at once).
sm_access() {
  local name=$1 ver=$2 out code
  SM_VERSION="" SM_VALUE=""
  out="$(mktemp "$RUN_DIR/sm.XXXXXX")" || return 2
  code="$(curl -sS --max-time 20 --output "$out" --write-out '%{http_code}' \
    --config <(printf 'header = "Authorization: Bearer %s"\n' "$SM_TOKEN") \
    "https://secretmanager.googleapis.com/v1/projects/${RCONF[GCP_PROJECT]}/secrets/$ENV_NAME-$name/versions/$ver:access")" || code=000
  case $code in
    200)
      SM_VERSION="$(jq -r '.name' "$out")" || { rm -f "$out"; return 2; }
      SM_VALUE="$(jq -r '.payload.data' "$out" | base64 -d)" || { rm -f "$out"; return 2; }
      rm -f "$out"
      SM_VERSION=${SM_VERSION##*/}
      [[ $SM_VERSION =~ ^[0-9]+$ ]] || return 2
      ;;
    404 | 400) rm -f "$out"; return 1 ;;
    *) rm -f "$out"; printf 'staging-vm: Secret Manager answered %s for %s-%s\n' "$code" "$ENV_NAME" "$name" >&2; return 2 ;;
  esac
}

# sm_token: SM_TOKEN = the VM service account's access token from the metadata server (never printed or stored).
sm_token() {
  SM_TOKEN="$(md instance/service-accounts/default/token | jq -r '.access_token')" || return 2
  [[ -n $SM_TOKEN && $SM_TOKEN != null ]] || return 2
}

is_optional() { [[ " ${SECRETS_OPTIONAL[*]} " == *" $1 "* ]]; }

identity_configured() { [[ ${SECRET_VERSION[IDENTITY_WORKOS_CLIENT_ID]:-none} != none ]]; }

# resolve_secrets deploy|pinned: fill SECRET_VERSION and SECRET_VALUE. deploy: PIN[NAME] (a number, latest or none)
# else latest; an optional secret without an enabled version is unset. pinned: exactly state/secret-versions.tsv.
# Returns non-zero with a message naming the secret (never a value) when a version is unavailable, and when
# CF_API_TOKEN lacks the Cloudflare API-token format (a fixed message: no value, length or prefix), in every mode and
# before render_env writes anything.
resolve_secrets() {
  local mode=$1 name ver rc n k
  if ((DRY_RUN)); then
    for name in "${SECRETS_REQUIRED[@]}" "${SECRETS_OPTIONAL[@]}"; do
      printf '[vm] (fetch secrets/%s-%s/versions/%s:access with the metadata token)\n' "$ENV_NAME" "$name" "${PIN[$name]:-$([[ $mode == deploy ]] && echo latest || echo pinned)}"
      SECRET_VERSION[$name]=none
    done
    return 0
  fi
  sm_token || { echo "staging-vm: no metadata token" >&2; return 2; }
  if [[ $mode == pinned ]]; then
    [[ -f $SRV/state/secret-versions.tsv ]] || { echo "staging-vm: no pinned secret versions: run staging.sh deploy" >&2; return 2; }
    while IFS=$'\t' read -r k ver; do
      [[ -n $k ]] && SECRET_VERSION[$k]=$ver
    done <"$SRV/state/secret-versions.tsv"
  fi
  for name in "${SECRETS_REQUIRED[@]}" "${SECRETS_OPTIONAL[@]}"; do
    if [[ $mode == deploy ]]; then ver=${PIN[$name]:-latest}; else ver=${SECRET_VERSION[$name]:-}; fi
    [[ $ver =~ ^([0-9]+|latest|none)$ ]] || { echo "staging-vm: no valid version recorded for $name" >&2; return 2; }
    if [[ $ver == none ]]; then
      is_optional "$name" || { echo "staging-vm: $name is required and cannot be none" >&2; return 2; }
      SECRET_VERSION[$name]=none
      continue
    fi
    rc=0
    sm_access "$name" "$ver" || rc=$?
    if ((rc == 0)); then
      case $SM_VALUE in
        "" | *"'"* | *$'\n'* | *$'\r'*) echo "staging-vm: $ENV_NAME-$name version $SM_VERSION is empty or holds a quote or line break" >&2; return 2 ;;
      esac
      if [[ $name == CF_API_TOKEN && ! $SM_VALUE =~ ^([A-Za-z0-9_-]{40}|cf(ut|at)_[A-Za-z0-9]{48})$ ]]; then
        SM_VALUE=""
        echo "staging-vm: CF_API_TOKEN does not have the Cloudflare API-token format; refused before any env file or container" >&2
        return 2
      fi
      SECRET_VERSION[$name]=$SM_VERSION
      SECRET_VALUE[$name]=$SM_VALUE
    elif ((rc == 1)) && is_optional "$name" && [[ $ver == latest ]]; then
      SECRET_VERSION[$name]=none
    else
      echo "staging-vm: $ENV_NAME-$name version $ver is unavailable" >&2
      return 2
    fi
  done
  n=0
  for name in "${IDENTITY_SECRETS[@]}"; do [[ ${SECRET_VERSION[$name]} == none ]] || n=$((n + 1)); done
  ((n == 0 || n == 3)) || { echo "staging-vm: the identity secrets must all have a version or none (found $n of 3)" >&2; return 2; }
  for name in "${SECRETS_REQUIRED[@]}" "${SECRETS_OPTIONAL[@]}"; do say secret "$name=${SECRET_VERSION[$name]}"; done
}

# write_secret_versions: record the resolved version numbers (names and numbers only) for the later scripts.
write_secret_versions() {
  local name
  if ((DRY_RUN)); then printf '[vm] (write %s/state/secret-versions.tsv: names and version numbers)\n' "$SRV"; return 0; fi
  install -d -m 0750 "$SRV/state"
  for name in "${SECRETS_REQUIRED[@]}" "${SECRETS_OPTIONAL[@]}"; do printf '%s\t%s\n' "$name" "${SECRET_VERSION[$name]}"; done >"$SRV/state/secret-versions.tsv.new"
  mv -f "$SRV/state/secret-versions.tsv.new" "$SRV/state/secret-versions.tsv"
}

# render_env: the run's private env file (0600, tmpfs), deleted by the EXIT trap. An empty optional setting is written
# as KEY= (unset for core: no default is chosen here).
render_env() {
  local k name svc var
  if ((DRY_RUN)); then printf '[vm] (render %s: remote.env values, IMAGE_* digests, the resolved secret versions)\n' "$ENV_FILE"; return 0; fi
  {
    for k in "${REMOTE_ENV_KEYS[@]}"; do
      [[ $META_KEYS != *" $k "* ]] || continue
      if [[ -z ${RCONF[$k]} ]]; then printf '%s=\n' "$k"; else printf "%s='%s'\n" "$k" "${RCONF[$k]}"; fi
    done
    for svc in "${IMAGE_SERVICES[@]}"; do var=${svc^^}; var=IMAGE_${var//-/_}; printf "%s='%s'\n" "$var" "${IMAGES[$var]}"; done
    for name in "${SECRETS_REQUIRED[@]}" "${SECRETS_OPTIONAL[@]}"; do
      [[ ${SECRET_VERSION[$name]} == none ]] || printf "%s='%s'\n" "$name" "${SECRET_VALUE[$name]}"
    done
    if identity_configured; then printf "IDENTITY_PUBLIC_ORIGINS='https://%s'\n" "${RCONF[PLATFORM_AUTH_HOST]}"; fi
  } >"$ENV_FILE"
  chmod 0600 "$ENV_FILE"
}

# compose_env MODE: resolve secrets and render the env for this run (MODE as resolve_secrets).
compose_env() {
  resolve_secrets "$1" || return $?
  render_env
}

# use_release DIR: the release every Compose file path of this run names (a concrete directory, so a later release
# never changes what an existing container mounts).
use_release() {
  if ((DRY_RUN)); then REL="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd -P)"; return 0; fi
  REL="$(readlink -f "$1")"
  [[ -f $REL/infra/compose/staging.yml ]] || die 2 "no release at $1: run staging.sh deploy"
}

volume_exists() { # volume_exists NAME
  local out
  capture out docker volume ls --quiet --filter "name=^$1\$"
  ((DRY_RUN)) || [[ $out == "$1" ]]
}

# pull_images: pull the manifest digests with a short-lived registry login kept on tmpfs (never in /root/.docker).
pull_images() {
  if ((DRY_RUN)); then
    printf '[vm] (docker login to %s with the metadata token on stdin, config on tmpfs)\n' "${RCONF[REGISTRY_HOST]}"
    dc pull --quiet
    return 0
  fi
  DOCKER_CFG="$(mktemp -d "$RUN_DIR/docker.XXXXXX")"
  printf '%s' "$SM_TOKEN" | docker --config "$DOCKER_CFG" login --username oauth2accesstoken --password-stdin "https://${RCONF[REGISTRY_HOST]}" >/dev/null
  export DOCKER_CONFIG="$DOCKER_CFG"
  dc pull --quiet
  unset DOCKER_CONFIG
  rm -rf "$DOCKER_CFG"
  DOCKER_CFG=""
  say pull "manifest digests present"
}

stores_up() {
  dc up -d --wait postgres redis minio minio-init || return
  say stores "postgres redis minio healthy; minio-init done"
}

# migration_status: PENDING = the number of pending migrations; non-zero when the status cannot be read consistently.
migration_status() {
  local out
  capture out dc run --rm -T migrate status || return 1
  if ((DRY_RUN)); then PENDING=1; return 0; fi # dry-run shows the pending path
  grep -q $'^current_version\t[0-9]' <<<"$out" || return 1
  PENDING="$(awk -F'\t' '$1 == "migration" && $3 == "pending" {n++} END {print n + 0}' <<<"$out")"
  say migrate-status "pending=$PENDING $(grep $'^current_version\t' <<<"$out" | tr '\t' '=')"
}

migrate_up() {
  dc run --rm -T migrate up || return
  say migrate-up "pending migrations applied"
}

# migrate_current: migrate status; when a migration or an image change is pending, stop app and edge first (the
# quiescent window of docs/foundation/MIGRATIONS.md), then apply only the pending migrations. Returns 1 when the status
# cannot be read consistently, 4 when stopping or migrating failed.
migrate_current() {
  migration_status || return 1
  if ((PENDING > 0)) || ! runs_manifest_images >/dev/null; then stop_app_edge || return 4; fi
  ((PENDING == 0)) || migrate_up || return 4
}

seed_dataset() {
  dc exec -T postgres psql --no-psqlrc --set=ON_ERROR_STOP=1 --username postgres --dbname "${RCONF[DATABASE_NAME]}" \
    <"$REL/infra/staging/seed.sql" || return
  ((!DRY_RUN)) || printf '[vm]   (stdin: infra/staging/seed.sql)\n'
  say seed "infra/staging/seed.sql on the fresh dataset"
}

# cid_of SERVICE: CID = the running container of SERVICE in this project ("" when none). Docker labels only: no Compose
# invocation, so no env file is needed.
cid_of() {
  local out
  CID=""
  capture out docker ps --quiet --filter "label=com.docker.compose.project=$ENV_NAME" --filter "label=com.docker.compose.service=$1" || out=""
  CID=${out%%$'\n'*}
}

# marker_label: MARKER = the dataset marker's platform label ("" when absent or unreadable).
marker_label() {
  MARKER=""
  cid_of postgres
  if ((DRY_RUN)) || [[ -n $CID ]]; then
    capture MARKER docker exec "$CID" psql --no-psqlrc --tuples-only --no-align --username postgres \
      --dbname "${RCONF[DATABASE_NAME]}" --command "SELECT platform_label FROM app.organizations WHERE id = '$MARKER_ID'" || MARKER=""
  fi
  [[ $MARKER =~ ^[0-9a-f]{32}$ ]] || MARKER=""
  ((!DRY_RUN)) || MARKER=dry-run
  return 0
}

# The number of tables holding at least one row in the database's own schemas (every schema except pg_catalog,
# information_schema and pg_*), goose's app.goose_db_version excluded; one LIMIT 1 probe per table.
DATASET_ROWS_SQL="SELECT count(*) FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
  WHERE c.relkind IN ('r', 'p') AND n.nspname NOT IN ('pg_catalog', 'information_schema') AND n.nspname !~ '^pg_'
    AND NOT (n.nspname = 'app' AND c.relname = 'goose_db_version')
    AND (xpath('/row/n/text()', query_to_xml(format('SELECT count(*) AS n FROM (SELECT 1 FROM %I.%I LIMIT 1) s',
      n.nspname, c.relname), false, true, '')))[1]::text::int > 0"

# dataset_empty: the database holds no application data (DATASET_ROWS_SQL answers exactly 0), read like marker_label
# as the superuser (row-level security does not apply). Returns 0 empty, 1 when a table holds a row, 2 when the count
# cannot be read (no PostgreSQL container, a query error or any answer but a number): never "empty" on doubt.
dataset_empty() {
  local out=""
  cid_of postgres
  if ((DRY_RUN)) || [[ -n $CID ]]; then
    capture out docker exec "$CID" psql --no-psqlrc --set=ON_ERROR_STOP=1 --tuples-only --no-align --username postgres \
      --dbname "${RCONF[DATABASE_NAME]}" --command "$DATASET_ROWS_SQL" || return 2
  fi
  ((!DRY_RUN)) || return 0
  [[ $out =~ ^[0-9]+$ ]] || return 2
  [[ $out == 0 ]] || return 1
}

# stop_by_label SERVICE...: docker stop of every container of each SERVICE in this project, by Docker labels only (no
# Compose, so no env file and no secret is needed). Non-zero when a listing or a stop failed: a container that cannot
# be listed is never taken as stopped.
stop_by_label() {
  local svc ids id rc=0
  for svc in "$@"; do
    ids=""
    capture ids docker ps --quiet --filter "label=com.docker.compose.project=$ENV_NAME" --filter "label=com.docker.compose.service=$svc" || { rc=1; continue; }
    if ((DRY_RUN)); then printf '[vm] + docker stop --time 60 <each listed %s container>\n' "$svc"; continue; fi
    while IFS= read -r id; do
      [[ -z $id ]] || run docker stop --time 60 "$id" >/dev/null || rc=1
    done <<<"$ids"
  done
  return "$rc"
}

# stop_app_edge: stop edge, web, intelligence and core (Compose when this run rendered its env, else by label).
stop_app_edge() {
  if ((DRY_RUN)) || [[ -s $ENV_FILE ]]; then
    dc stop edge web intelligence core || return
  else
    stop_by_label edge web intelligence core || return
  fi
  say app-edge-stopped "edge web intelligence core"
}

# hold REASON: leave app and edge stopped, write a HOLD receipt, exit 3. Nothing is reseeded or reset.
hold() {
  say hold "$1"
  stop_app_edge || say hold "app/edge stop reported an error; check docker ps"
  if ((!DRY_RUN)); then
    install -d -m 0750 "$SRV/state"
    printf 'time\t%s\nenvironment\t%s\nreason\t%s\naction\tescalate; app and edge stay stopped; no reseed (reset only by reset --confirm)\n' \
      "$(ts)" "$ENV_NAME" "$1" >"$SRV/state/hold-$(date -u +%Y%m%dT%H%M%SZ).tsv"
  fi
  printf 'result\tHOLD\t%s\n' "$1"
  exit 3
}

app_up() {
  dc up -d --wait core web intelligence edge || return
  say app-edge "core web intelligence edge running"
}

# runs_manifest_images: every app/edge container runs exactly the manifest's digest. Prints one line per service.
runs_manifest_images() {
  local svc var img digests ok=0
  for svc in core web intelligence edge; do
    var=${svc^^}
    var=IMAGE_${var//-/_}
    img="" digests=""
    cid_of "$svc"
    if ((DRY_RUN)) || [[ -n $CID ]]; then
      capture img docker inspect --format '{{.Image}}' "$CID" || img=""
      capture digests docker image inspect --format '{{range .RepoDigests}}{{println .}}{{end}}' "$img" || digests=""
    fi
    if ((DRY_RUN)) || { [[ -n $digests ]] && grep -qxF "${IMAGES[$var]}" <<<"$digests"; }; then
      printf 'digest\t%s\t%s\tmatch\n' "$svc" "${IMAGES[$var]}"
    else
      printf 'digest\t%s\t%s\tMISMATCH\n' "$svc" "${IMAGES[$var]}"
      ok=1
    fi
  done
  return "$ok"
}

# cert_lines HOST: the certificate the edge serves for HOST (serial, not-after, issuer, san = the comma-separated
# subjectAltName entries such as DNS:B,DNS:*.B, valid_7d) as cert\t lines.
cert_lines() {
  local host=$1 pem
  if ((DRY_RUN)); then run openssl s_client -connect 127.0.0.1:443 -servername "$host"; return 0; fi
  pem="$({ openssl s_client -connect 127.0.0.1:443 -servername "$host" </dev/null 2>/dev/null || true; } | openssl x509 2>/dev/null)" || return 1
  [[ -n $pem ]] || return 1
  printf 'cert\t%s\tserial\t%s\n' "$host" "$(openssl x509 -noout -serial <<<"$pem" | cut -d= -f2)"
  printf 'cert\t%s\tnot_after\t%s\n' "$host" "$(openssl x509 -noout -enddate <<<"$pem" | cut -d= -f2)"
  printf 'cert\t%s\tissuer\t%s\n' "$host" "$(openssl x509 -noout -issuer <<<"$pem" | sed 's/^issuer=//')"
  printf 'cert\t%s\tsan\t%s\n' "$host" "$(openssl x509 -noout -ext subjectAltName <<<"$pem" | tail -n +2 | tr -d ' ')"
  if openssl x509 -noout -checkend 604800 <<<"$pem" >/dev/null; then
    printf 'cert\t%s\tvalid_7d\tyes\n' "$host"
  else
    printf 'cert\t%s\tvalid_7d\tno\n' "$host"
  fi
}

# san_has LINES HOST NAME: HOST's certificate (the cert\t lines of cert_lines) lists NAME exactly as a DNS
# subjectAltName.
san_has() {
  local san entry entries
  san="$(awk -F'\t' -v h="$2" '$2 == h && $3 == "san" {print $4}' <<<"$1")"
  IFS=, read -ra entries <<<"$san"
  for entry in "${entries[@]}"; do [[ $entry != "DNS:$3" ]] || return 0; done
  return 1
}

# cert_proof LINES HOST NAME: HOST's certificate lists NAME, is valid for 7 more days and, in the rehearsal, comes from
# the Let's Encrypt staging directory (a (STAGING) issuer). Says why and returns 1 otherwise.
cert_proof() {
  san_has "$1" "$2" "$3" || { say https-ready "the certificate served for $2 does not list $3"; return 1; }
  awk -F'\t' -v h="$2" '$2 == h && $3 == "valid_7d" && $4 == "yes" {f = 1} END {exit !f}' <<<"$1" ||
    { say https-ready "the certificate served for $2 expires within 7 days"; return 1; }
  if ((REHEARSAL)) && ! awk -F'\t' -v h="$2" '$2 == h && $3 == "issuer" && index($4, "(STAGING)") {f = 1} END {exit !f}' <<<"$1"; then
    say https-ready "rehearsal certificate for $2 is not from the staging directory"
    return 1
  fi
}

# https_ready: per name, through the edge on this VM: the platform root B answers /api/ready and a random label host
# completes a request (any HTTP status), each with a verified chain and host name (curl; the Let's Encrypt staging
# directory's untrusted chain is accepted only in the rehearsal and only with a (STAGING) issuer); B's certificate
# lists B and the label's lists *.B, each valid for at least 7 more days. Caddy's {B, *.B} order may be served as one
# certificate or as distinct ones for the exact and the wildcard name: serials are printed, never compared.
# Non-zero when not ready.
https_ready() {
  local b=${RCONF[PLATFORM_BASE_DOMAIN]} i tls=() probe base wild
  REHEARSAL=0
  [[ ${RCONF[EDGE_ACME_CA]} != "$LE_STAGING_CA" ]] || { tls=(--insecure); REHEARSAL=1; }
  for ((i = 0; i < 60; i++)); do
    if run curl -fs --max-time 10 --resolve "$b:443:127.0.0.1" "${tls[@]}" -o /dev/null "https://$b/api/ready"; then break; fi
    ((i < 59)) || { say https-ready "no answer from https://$b/api/ready after 5 minutes"; return 1; }
    sleep 5
  done
  probe="$(head -c 6 /dev/urandom | od -An -tx1 | tr -d ' \n').$b"
  run curl -s --max-time 10 --resolve "$probe:443:127.0.0.1" "${tls[@]}" -o /dev/null "https://$probe/" ||
    { say https-ready "no verified TLS answer for a random label host"; return 1; }
  base="$(cert_lines "$b")" || { say https-ready "no certificate for $b"; return 1; }
  wild="$(cert_lines "$probe")" || { say https-ready "no certificate for a random label host"; return 1; }
  printf '%s\n%s\n' "$base" "$wild"
  ((DRY_RUN)) && return 0
  cert_proof "$base" "$b" "$b" || return 1
  cert_proof "$wild" "$probe" "*.$b" || return 1
  if ((REHEARSAL)); then
    say https-ready "rehearsal: staging-directory chains for $b and a label host (untrusted by design)"
  else
    say https-ready "verified chains: $b lists $b, a label host lists *.$b; both valid for 7 more days"
  fi
}

# Probe containers of metadata_denied: identity, inventory, reconciliation.
# Every probe container has a unique name, explicit ownership labels (the G070 task, the environment, the probe kind and
# this run's identity) and a --cidfile, and runs WITHOUT --rm: Docker writes the container ID to the cidfile when it
# creates the container, so the exact ID and the immutable image ID are in the inventory before anything removes the
# container, and removal is an explicit, verified step. The inventory is $SRV/state/probe-containers.tsv (append-only
# TSV, the existing state directory that survives a reboot): ts, run, name, event, network, image ID, container ID,
# detail; events allocate (before docker run; names the image ID), created (the container ID), removed, reaped
# (removed by probe_reap), remove-failed, held (a candidate that does not match its record, never removed) and
# stale-closed (a record whose container no longer exists). The probe-container and probe-reap steps repeat the
# receipt. A container found by label is removed only when its name has an allocate record whose run, image ID and
# container ID (cidfile) all equal what the daemon reports; anything else is held and fails the check. Fail closed
# throughout: the check never starts a probe container while reconciliation is unresolved or before its allocate
# record is written (every probe_record result is checked: callers run inside `|| hold`, where errexit is suspended),
# and a cidfile is deleted only after the container's absence is verified.
probe_cmd() {
  PROBE_SEQ=$((PROBE_SEQ + 1))
  PROBE_NAME="g070-md-$PROBE_RUN-$PROBE_SEQ"
  PROBE_CIDFILE="$SRV/state/probe-cids/$PROBE_NAME"
  # --init: `timeout` must not be the container's PID 1. The kernel never delivers a signal without a handler to a PID
  # namespace's init, so BusyBox `timeout` as the entrypoint never fires: a guard-dropped connect then ran for the
  # kernel's SYN retries (about 130 s per probe, measured on the VM) instead of 5 s. Under docker-init it is an
  # ordinary child and exits 143 after 5 s; the status classification is unchanged.
  PCMD=(docker run --pull never --init --name "$PROBE_NAME" --cidfile "$PROBE_CIDFILE" --label g070.task=G070 --label "g070.env=$ENV_NAME"
    --label g070.probe=metadata-denied --label "g070.probe.run=$PROBE_RUN" --network "$1" --entrypoint timeout
    "${IMAGES[IMAGE_EDGE]}" 5 wget -q -O- --header 'Metadata-Flavor: Google' "$2")
}

# probe_record EVENT NAME NETWORK IMAGE_ID CONTAINER_ID [DETAIL]: one inventory row; non-zero when it cannot be written.
probe_record() {
  ((!DRY_RUN)) || return 0
  install -d -m 0750 "$SRV/state" "$SRV/state/probe-cids" || return 1
  printf '%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n' "$(ts)" "${PROBE_RUN:--}" "$2" "$1" "$3" "$4" "$5" "${6:-}" >>"$SRV/state/probe-containers.tsv" || return 1
}

# probe_remove NAME CONTAINER_ID: remove exactly that container and verify its absence; the cidfile goes only then.
probe_remove() {
  local left=""
  docker rm -f "$2" >/dev/null 2>&1 || true
  capture left docker ps -aq --filter "id=$2" || left="unknown"
  [[ -z $left ]] || return 1
  rm -f "$SRV/state/probe-cids/$1"
}

# probe_exec NETWORK URL DETAIL: one probe container, recorded and removed by identity; sets PROBE_RC and PROBE_OUT (the
# answer: only tested for emptiness by the caller, never printed). PROBE_RC 125 = the probe could not be proven to have
# run (no image ID, no allocate record, no container ID, an inventory write failed, removal not verified); the creator
# is never invoked unless the image ID is known and the allocate record is written.
probe_exec() {
  local img="" cid=""
  PROBE_OUT="" PROBE_RC=0
  probe_cmd "$1" "$2"
  forbid "${PCMD[@]}"
  capture img docker image inspect --format '{{.Id}}' "${IMAGES[IMAGE_EDGE]}" || img=""
  [[ -n $img ]] || { PROBE_RC=125; return 0; }
  if ! probe_record allocate "$PROBE_NAME" "$1" "$img" - "$3"; then
    say probe-container "$PROBE_NAME: the allocate record could not be written; no container was started"
    PROBE_RC=125
    return 0
  fi
  say probe-container "allocated $PROBE_NAME on $1 image $img${3:+ ($3)}"
  PROBE_OUT="$("${PCMD[@]}" 2>/dev/null)" || PROBE_RC=$?
  [[ ! -r $PROBE_CIDFILE ]] || cid="$(<"$PROBE_CIDFILE")"
  if [[ ! $cid =~ ^[0-9a-f]{64}$ ]]; then
    say probe-container "$PROBE_NAME has no container ID (not created?); the probe is unproven"
    PROBE_OUT="" PROBE_RC=125
    return 0
  fi
  if ! probe_record created "$PROBE_NAME" "$1" "$img" "$cid"; then
    # the identity is known (cidfile) but cannot be recorded: remove exactly that container and fail closed
    say probe-container "$PROBE_NAME id ${cid:0:12}: the created record could not be written; removing it by its ID"
    probe_remove "$PROBE_NAME" "$cid" || say probe-container "$PROBE_NAME id ${cid:0:12} could not be removed"
    PROBE_OUT="" PROBE_RC=125
    return 0
  fi
  say probe-container "created $PROBE_NAME id ${cid:0:12}"
  if probe_remove "$PROBE_NAME" "$cid"; then
    probe_record removed "$PROBE_NAME" "$1" "$img" "$cid" || { say probe-container "$PROBE_NAME: the removal could not be recorded"; PROBE_OUT="" PROBE_RC=125; }
  else
    probe_record remove-failed "$PROBE_NAME" "$1" "$img" "$cid" || true
    say probe-container "$PROBE_NAME id ${cid:0:12} could not be removed"
    PROBE_OUT="" PROBE_RC=125
  fi
}

# probe_reap: reconcile this environment's G070 probe containers and records, before the check starts (stale ones of
# an interrupted or rebooted earlier run), when it ends and from cleanup (the EXIT trap). Candidates are found by the
# task/probe/environment labels, never by name or image alone; each is verified against its allocate record (run label,
# image ID from docker inspect .Image, container ID from the cidfile) and removed, or held and reported. Records whose
# container is gone are closed as stale. A cidfile is deleted only after the container's absence is verified (reaped
# containers and closed stale records); a held container keeps its cidfile as evidence. A stale record is closed only
# after its cidfile is verified gone (a failed unlink keeps the record open and the check unresolved, retryable).
# Returns 1 when a candidate is held, still present, a cidfile cannot be removed or a record cannot be written.
probe_reap() {
  local cands="" left="" id name run rrow rrun rimg cid img held=0 inv="$SRV/state/probe-containers.tsv" nm
  ((!DRY_RUN)) || return 0
  capture cands docker ps -a --filter label=g070.task=G070 --filter label=g070.probe=metadata-denied \
    --filter "label=g070.env=$ENV_NAME" --format '{{.ID}}'$'\t''{{.Names}}'$'\t''{{.Label "g070.probe.run"}}' || { say probe-reap "cannot list the probe containers"; return 1; }
  while IFS=$'\t' read -r id name run; do
    [[ -n $id ]] || continue
    rrow="$(awk -F'\t' -v n="$name" '$3 == n && $4 == "allocate" {r = $0} END {print r}' "$inv" 2>/dev/null || true)"
    rrun="$(cut -f2 <<<"$rrow")" rimg="$(cut -f6 <<<"$rrow")" cid=""
    [[ ! -r $SRV/state/probe-cids/$name ]] || cid="$(<"$SRV/state/probe-cids/$name")"
    img=""
    capture img docker inspect --format '{{.Image}}' "$id" || img=""
    if [[ -n $rrow && $rrun == "$run" && -n $rimg && $img == "$rimg" && $cid =~ ^[0-9a-f]{64}$ && $cid == "$id"* ]]; then
      if probe_remove "$name" "$cid"; then
        if probe_record reaped "$name" - "$rimg" "$cid" "run $run"; then
          say probe-reap "removed $name id $id (recorded run $run, image $rimg)"
        else
          say probe-reap "removed $name id $id but the inventory row could not be written"
          held=1
        fi
      else
        probe_record remove-failed "$name" - "$rimg" "$cid" "run $run" || true
        say probe-reap "$name id $id is recorded for $ENV_NAME but could not be removed"
        held=1
      fi
    else
      probe_record held "$name" - "${img:--}" "$id" "no matching allocate record, run label, image ID or container ID" || true
      say probe-reap "HELD $name id $id: not recorded for $ENV_NAME or its identity differs; left untouched"
      held=1
    fi
  done <<<"$cands"
  # records that never reached removal, whose container no longer exists
  if [[ -r $inv && ! -d $inv ]]; then
    while IFS= read -r nm; do
      [[ -n $nm ]] || continue
      left=""
      capture left docker ps -aq --filter "name=^${nm}\$" || left="unknown"
      [[ -z $left ]] || continue
      # the cidfile goes first and must be verified gone: only then is the record closed, so a failed unlink leaves the
      # record open and the recovery retryable (and unresolved now)
      if ! rm -f "$SRV/state/probe-cids/$nm" 2>/dev/null || [[ -e $SRV/state/probe-cids/$nm ]]; then
        say probe-reap "the cidfile of the stale record $nm could not be removed; the record stays open"
        held=1
      elif probe_record stale-closed "$nm" - - - "record without a container"; then
        say probe-reap "closed the stale record of $nm (no container exists)"
      else
        say probe-reap "the stale record of $nm could not be closed"
        held=1
      fi
    done < <(awk -F'\t' '{last[$3] = $4} END {for (n in last) if (last[n] == "allocate" || last[n] == "created" || last[n] == "remove-failed") print n}' "$inv" | LC_ALL=C sort)
  fi
  left=""
  capture left docker ps -a --filter label=g070.task=G070 --filter label=g070.probe=metadata-denied \
    --filter "label=g070.env=$ENV_NAME" --format '{{.ID}}' || left="unknown"
  if [[ -n $left ]]; then
    say probe-reap "probe containers of $ENV_NAME remain (held or not removable)"
    return 1
  fi
  ((held == 0)) || return 1
  say probe-reap "no probe container of $ENV_NAME remains"
}

# probe_stop NETWORK: an unresolved lifecycle or accounting failure of a probe container (PROBE_RC 125: no image or
# container ID, an inventory row not written, a removal not verified, or Docker itself failing): no further probe
# container may be created while one might still exist or be unaccounted for. Report it, reconcile through the exact-owner
# path (probe_reap; a container that cannot be verified absent stays and is reported), and let the caller return 1 so
# readiness HOLDs.
probe_stop() {
  printf 'metadata-denied\t%s\tUNPROVEN\n' "$1"
  say metadata-denied "a probe container on $1 could not be proven created, recorded or removed; no further probe container is started"
  probe_reap || say metadata-denied "reconciliation of the probe containers of $ENV_NAME is unresolved; HOLD"
}

# metadata_denied: no container network of this project reaches the Compute Engine metadata server (the startup
# script's staging-metadata-guard drops it; this proves it from the real networks). For every Compose network of the
# project, throwaway containers of the manifest's edge image (busybox, never pulled) first run a tool control, the same
# wget against a closed loopback port, which must fail with exactly status 1 (a crash such as 139, a missing applet or
# any other status means the probe tool cannot be trusted); then ask 169.254.169.254, [fd20:ce::254] and
# metadata.google.internal for the service account's token under busybox `timeout 5` (never wget -T: BusyBox 1.35's
# wget segfaults with it, exit 139, and an empty body). Fail closed: a probe counts as denied only when its answer is
# empty AND its status is wget's failure status 1 (refused, unreachable, bad address) or the timeout's 124/143 (the
# guard drops silently); status 0 or any answer is REACHABLE; every other status, 139, 137 and Docker's 125-127
# included, is UNPROVEN. The answer stays in a local variable that is only tested for emptiness and then cleared: never
# printed, logged or stored. Stale probe containers and records of earlier runs are reconciled first (probe_reap); an
# unresolved reconciliation stops the check before any probe container is created, and an unresolved lifecycle or
# accounting failure of a probe (probe_stop) stops it at once: no later endpoint or network is probed.
# Returns 1 after a metadata-denied<TAB><network><TAB>REACHABLE|UNPROVEN line or a held container; callers HOLD.
metadata_denied() {
  local nets net url bad=0 path=/computeMetadata/v1/instance/service-accounts/default/token
  nets=""
  PROBE_RUN="$(date -u +%Y%m%d%H%M%S)-$$" PROBE_SEQ=0
  if ! probe_reap; then
    say metadata-denied "unresolved probe containers or records of $ENV_NAME; no new probe container is started"
    return 1
  fi
  capture nets docker network ls --filter "label=com.docker.compose.project=$ENV_NAME" --format '{{.Name}}' || nets=""
  ((!DRY_RUN)) || nets="<network>"
  [[ -n $nets ]] || { say metadata-denied "no network of $ENV_NAME to probe"; return 1; }
  while IFS= read -r net; do
    [[ -n $net ]] || continue
    if ((DRY_RUN)); then
      probe_cmd "$net" http://127.0.0.1:1/
      run "${PCMD[@]}"
    else
      probe_exec "$net" http://127.0.0.1:1/ "tool control"
      if ((PROBE_RC == 125)); then probe_stop "$net"; return 1; fi
      if ((PROBE_RC != 1)) || [[ -n $PROBE_OUT ]]; then
        printf 'metadata-denied\t%s\tUNPROVEN\n' "$net"
        bad=1
        PROBE_OUT=""
        continue
      fi
    fi
    for url in "http://169.254.169.254$path" "http://[fd20:ce::254]$path" "http://metadata.google.internal$path"; do
      if ((DRY_RUN)); then
        probe_cmd "$net" "$url"
        run "${PCMD[@]}"
        continue
      fi
      probe_exec "$net" "$url" ""
      if ((PROBE_RC == 125)); then probe_stop "$net"; return 1; fi
      if ((PROBE_RC == 0)) || [[ -n $PROBE_OUT ]]; then
        printf 'metadata-denied\t%s\tREACHABLE\n' "$net"
        bad=1
      elif ((PROBE_RC != 1 && PROBE_RC != 124 && PROBE_RC != 143)); then
        printf 'metadata-denied\t%s\tUNPROVEN\n' "$net"
        bad=1
      fi
      PROBE_OUT=""
    done
  done <<<"$nets"
  probe_reap || bad=1
  ((bad == 0)) || return 1
  say metadata-denied "every network of $ENV_NAME: 169.254.169.254, fd20:ce::254 and metadata.google.internal unreachable"
}

# readiness: the metadata boundary from every network (HOLD otherwise: app and edge stop), HTTPS readiness, running
# digests and the marker; prints "result\tREADY" or returns non-zero.
readiness() {
  metadata_denied || hold "a container network reaches the metadata server or could not be probed (staging-metadata-guard)"
  https_ready || return 1
  runs_manifest_images || { say readiness "a running image differs from the manifest"; return 1; }
  marker_label
  [[ -n $MARKER ]] || { say readiness "dataset marker absent"; return 1; }
  printf 'marker\t%s\n' "$MARKER"
  printf 'result\tREADY\t%s\n' "${RCONF[CANDIDATE]:-dry-run}"
}
