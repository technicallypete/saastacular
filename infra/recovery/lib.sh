# shellcheck shell=bash
# Shared helpers of infra/recovery/recovery.sh (G013; docs/foundation/RECOVERY.md). Sourced, never executed.
#
# Value handling: secret values are read from the operator's env files and passed to containers only through
# --env-file process substitutions (never argv, never output, never a file outside the operator's private
# directories). Containers are located by their Compose labels (project and service). Every container this tool
# creates is labelled g013.recovery=<command> plus RECOVERY_LABEL when the caller sets it, and a detached one also
# writes a distinct, initially absent cidfile under RECOVERY_CID_DIR when that is set (the harness ownership registry).

RECOVERY_REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)"
RECOVERY_SQL="$RECOVERY_REPO/infra/recovery/sql"
# shellcheck disable=SC2034  # the format names are consumed by recovery.sh
MANIFEST_FORMAT="g013-backup/1"
LEDGER_FORMAT="g013-ledger/1"
# shellcheck disable=SC2034
FENCE_FORMAT="g013-fence/1"
UUID_RE='^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'
STEP_TIMEOUT="${RECOVERY_STEP_TIMEOUT:-300}"

# --- output and failure ------------------------------------------------------------------------------------------------

log() { printf '%s recovery: %s\n' "$(date -u +%FT%TZ)" "$*" >&2; }

# die CODE REASON MESSAGE: one closed reason code on stderr, then exit CODE (2 input, 3 refused, 4 failed).
die() {
  local code="$1" reason="$2"; shift 2
  printf 'recovery: refused reason=%s: %s\n' "$reason" "$*" >&2
  exit "$code"
}

now_utc() { date -u +%Y-%m-%dT%H:%M:%SZ; }
new_uuid() { local h; h="$(head -c 16 /dev/urandom | od -An -tx1 | tr -d ' \n')"
  printf '%s-%s-4%s-%x%s-%s' "${h:0:8}" "${h:8:4}" "${h:13:3}" $(( (16#${h:16:1} & 3) | 8 )) "${h:17:3}" "${h:20:12}"; }
sha_file() { sha256sum -- "$1" | cut -d' ' -f1; }
sha_text() { printf '%s' "$1" | sha256sum | cut -d' ' -f1; }

# private_dir DIR: create (0700) or verify an operator-private directory outside the repository.
private_dir() {
  local d="$1"
  [[ "$d" == /* ]] || die 2 invalid_input "directories must be absolute"
  case "$(realpath -m -- "$d")/" in "$RECOVERY_REPO"/*) die 2 invalid_input "private material must stay outside the repository" ;; esac
  if [[ ! -e "$d" ]]; then
    mkdir -p -- "$d" 2>/dev/null && chmod 700 -- "$d" 2>/dev/null || die 3 destination_unavailable "cannot create a private directory"
  fi
  # An existing directory is never re-permissioned: it must already be a writable, owner-only directory.
  [[ -d "$d" && -w "$d" && -x "$d" ]] || die 3 destination_unavailable "a private directory is not writable"
  [[ "$(stat -c %a -- "$d")" =~ ^[0-7]00$ ]] || die 3 destination_unavailable "a private directory is readable by others (mode 0700 expected)"
}

# --- env files -----------------------------------------------------------------------------------------------------------

# envv FILE KEY: the value of KEY (last assignment; surrounding double quotes removed).
envv() { sed -n "s/^$2=//p" "$1" | tail -1 | sed 's/^"\(.*\)"$/\1/'; }

check_env_file() { # check_env_file FILE KEY...: readable, and every KEY non-empty
  local f="$1" k; shift
  [[ -f "$f" && -r "$f" ]] || die 2 invalid_input "env file not readable"
  for k in "$@"; do [[ -n "$(envv "$f" "$k")" ]] || die 2 invalid_input "env file lacks $k"; done
}

# --- Docker resolution ---------------------------------------------------------------------------------------------------

LABELS=()
recovery_labels() { # recovery_labels COMMAND: the label arguments for containers this command creates
  LABELS=(--label "g013.recovery=$1")
  [[ -n "${RECOVERY_LABEL:-}" ]] && LABELS+=(--label "$RECOVERY_LABEL")
  return 0
}

# cid_of PROJECT SERVICE [any]: the container of a Compose service (running only unless "any").
cid_of() {
  local flt=(--filter "label=com.docker.compose.project=$1" --filter "label=com.docker.compose.service=$2")
  if [[ "${3:-}" == any ]]; then docker ps -aq --no-trunc "${flt[@]}" | head -1; else docker ps -q --no-trunc "${flt[@]}" | head -1; fi
}
volume_of() { docker inspect -f '{{range .Mounts}}{{if eq .Type "volume"}}{{.Name}}{{end}}{{end}}' "$1" 2>/dev/null; }
network_of() { docker inspect -f '{{range $k, $v := .NetworkSettings.Networks}}{{$k}}{{"\n"}}{{end}}' "$1" 2>/dev/null | grep -m1 . ; }
image_of() { docker inspect -f '{{.Image}}' "$1" 2>/dev/null; }

# resolve SIDE ENVFILE: SIDE_PROJECT, SIDE_DB, SIDE_PG, SIDE_PGVOL, SIDE_NET, SIDE_PGIMG, SIDE_MINIO, SIDE_MINIOVOL,
# SIDE_MCIMG, SIDE_BUCKET and SIDE_MINIO_UP (exported as plain globals: SRC_* or TGT_*). PostgreSQL must be running;
# MinIO must be running unless the third argument is "minio-may-stop" (restore, replay and tombstone purge then record
# the provider as unavailable for their object work instead of refusing; its stopped container still identifies it).
resolve() {
  local side="$1" f="$2" p db pg mi mcimg up=1
  check_env_file "$f" COMPOSE_PROJECT_NAME DATABASE_NAME OBJECT_STORE_BUCKET
  p="$(envv "$f" COMPOSE_PROJECT_NAME)"; db="$(envv "$f" DATABASE_NAME)"
  pg="$(cid_of "$p" postgres)"; [[ -n "$pg" ]] || die 3 source_unavailable "$side PostgreSQL is not running"
  mi="$(cid_of "$p" minio)"
  if [[ -z "$mi" && "${3:-}" == minio-may-stop ]]; then mi="$(cid_of "$p" minio any)"; up=0; fi
  [[ -n "$mi" ]] || die 3 source_unavailable "$side MinIO is not running"
  mcimg=""
  if docker image inspect "$p-minio-init" >/dev/null 2>&1; then mcimg="$p-minio-init"
  else
    local ic; ic="$(cid_of "$p" minio-init any)"
    [[ -n "$ic" ]] && mcimg="$(image_of "$ic")"
  fi
  [[ -n "$mcimg" ]] || die 3 tool_unavailable "$side has no built minio-init image"
  printf -v "${side}_PROJECT" '%s' "$p"; printf -v "${side}_DB" '%s' "$db"; printf -v "${side}_PG" '%s' "$pg"
  printf -v "${side}_PGVOL" '%s' "$(volume_of "$pg")"; printf -v "${side}_NET" '%s' "$(network_of "$pg")"
  printf -v "${side}_PGIMG" '%s' "$(image_of "$pg")"; printf -v "${side}_MINIO" '%s' "$mi"
  printf -v "${side}_MINIOVOL" '%s' "$(volume_of "$mi")"; printf -v "${side}_MCIMG" '%s' "$mcimg"
  printf -v "${side}_BUCKET" '%s' "$(envv "$f" OBJECT_STORE_BUCKET)"
  printf -v "${side}_ENV" '%s' "$f"; printf -v "${side}_MINIO_UP" '%s' "$up"
}

# --- PostgreSQL (superuser over the container's local socket; values via psql variables) ---------------------------------

# psql_c CID DB [-v k=v]...: psql reading SQL on stdin (tuples only, unaligned, stop on error, bounded).
psql_c() {
  local cid="$1" db="$2"; shift 2
  timeout "$STEP_TIMEOUT" docker exec -i "$cid" psql -X -q -tA -v ON_ERROR_STOP=1 -U postgres -d "$db" "$@"
}
# sql1 CID DB SQL: one statement, its single result.
sql1() { printf '%s\n' "$3" | psql_c "$1" "$2"; }
# sqlf CID DB FILE [-v k=v]...: run a reviewed SQL file from infra/recovery/sql.
sqlf() { local cid="$1" db="$2" f="$3"; shift 3; psql_c "$cid" "$db" "$@" <"$RECOVERY_SQL/$f"; }

# pg_client NET ENVFILE USERKEY|super [docker args] -- CMD...: a one-shot PostgreSQL client container on NET with the
# password of the given env key in PGPASSWORD (POSTGRES_SUPERUSER_PASSWORD for super).
pg_client() {
  local net="$1" f="$2" who="$3" img="$4"; shift 4
  local key="$who" user="postgres"
  if [[ "$who" == super ]]; then key=POSTGRES_SUPERUSER_PASSWORD; fi
  if [[ "$who" == retention ]]; then key=DB_RETENTION_PASSWORD; user="$(envv "$f" DATABASE_NAME)_retention"; fi
  docker run --rm -i "${LABELS[@]}" --network "$net" --env-file <(printf 'PGPASSWORD=%s\nPGUSER=%s\nPGHOST=postgres\nPGDATABASE=%s\n' \
    "$(envv "$f" "$key")" "$user" "$(envv "$f" DATABASE_NAME)") "$img" "$@"
}

pg_sysid() { sql1 "$1" "$2" "SELECT system_identifier FROM pg_control_system()"; }
pg_major() { sql1 "$1" "$2" "SELECT current_setting('server_version_num')::int / 10000"; }
schema_version() { sql1 "$1" "$2" "SELECT coalesce((SELECT max(version_id) FROM app.goose_db_version WHERE is_applied), 0)" 2>/dev/null || echo 0; }

# --- MinIO (mc in a one-shot container of the project's minio-init image) ------------------------------------------------

# mcx NET IMG ENVFILE admin|svc CMD: mc with alias x for the fixture administrator or the scoped service user.
mcx() {
  local net="$1" img="$2" f="$3" who="$4"; shift 4
  local u p
  if [[ "$who" == admin ]]; then u="$(envv "$f" MINIO_ROOT_USER)"; p="$(envv "$f" MINIO_ROOT_PASSWORD)"
  else u="$(envv "$f" OBJECT_STORE_ACCESS_KEY)"; p="$(envv "$f" OBJECT_STORE_SECRET_KEY)"; fi
  timeout "$STEP_TIMEOUT" docker run --rm -i "${LABELS[@]}" --network "$net" -e MC_CONFIG_DIR=/tmp/mc \
    --env-file <(printf 'MC_HOST_x=http://%s:%s@minio:9000\n' "$u" "$p") "$img" sh -c "$*"
}

# minio_ready CID: wait (bounded) until the provider answers.
minio_ready() {
  local i
  for ((i = 0; i < 120; i++)); do docker exec "$1" /usr/bin/mc ready local >/dev/null 2>&1 && return 0; sleep 1; done
  return 1
}

# objects_list NET IMG ENVFILE BUCKET: "name<TAB>size<TAB>etag" per object (sorted), as the fixture administrator.
objects_list() {
  local out
  out="$(mcx "$1" "$2" "$3" admin "mc ls --json --recursive 'x/$4/'")" || return 1
  printf '%s\n' "$out" | sed -n 's/.*"size":\([0-9]*\),"key":"\([^"]*\)","etag":"\([^"]*\)".*/\2\t\1\t\3/p' | LC_ALL=C sort
}

# --- fence records -------------------------------------------------------------------------------------------------------

fence_put() { # fence_put FILE KEY VALUE...: append one progress line and flush it to disk
  local f="$1"; shift
  local IFS=$'\t'
  printf '%s\n' "$*" >>"$f" && sync -- "$f"
}
fence_get() { [[ -f "$1" ]] || return 0; awk -F'\t' -v k="$2" '$1 == k { $1 = ""; sub(/^\t/, ""); v = $0 } END { if (v != "") print v }' OFS='\t' "$1"; }
fence_has() { awk -F'\t' -v k="$2" '$1 == k { f = 1 } END { exit !f }' "$1"; }

# --- strict KEY=VALUE policy files ---------------------------------------------------------------------------------------

declare -A POL=()
# read_policy FILE ALLOWED REQUIRED: unknown, duplicate or missing keys refuse (exit 2); values go to POL.
read_policy() {
  local f="$1" allowed=" $2 " required="$3" line k v
  POL=()
  [[ -f "$f" && -r "$f" ]] || die 2 policy_invalid "policy file not readable"
  while IFS= read -r line || [[ -n "$line" ]]; do
    [[ -z "$line" || "$line" == \#* ]] && continue
    [[ "$line" =~ ^([a-z_]+)=(.*)$ ]] || die 2 policy_invalid "malformed line"
    k="${BASH_REMATCH[1]}" v="${BASH_REMATCH[2]}"
    [[ "$allowed" == *" $k "* ]] || die 2 policy_invalid "unknown key $k"
    [[ -z "${POL[$k]+x}" ]] || die 2 policy_invalid "duplicate key $k"
    POL[$k]="$v"
  done <"$f"
  for k in $required; do [[ -n "${POL[$k]+x}" ]] || die 2 policy_invalid "missing key $k"; done
}
int_in() { [[ "$1" =~ ^[0-9]+$ ]] && (( 10#$1 >= $2 && 10#$1 <= $3 )); }

# check_holds VALUE: comma-separated canonical UUIDs (at most 256), or empty.
check_holds() {
  local h="$1" n=0 u
  [[ -z "$h" ]] && return 0
  local IFS=,
  for u in $h; do [[ "$u" =~ $UUID_RE ]] || return 1; n=$((n + 1)); done
  (( n <= 256 ))
}

# --- the restore-input ledger --------------------------------------------------------------------------------------------

# ledger_check FILE: header and hash chain valid; prints "UUID<TAB>DB<TAB>HEADSEQ<TAB>HEADDIGEST". Non-zero on failure.
ledger_check() {
  local f="$1" header prev line seq=0 fields
  [[ -f "$f" && -r "$f" ]] || return 1
  header="$(head -1 "$f")"
  IFS=$'\t' read -r -a fields <<<"$header"
  [[ "${fields[0]:-}" == format && "${fields[1]:-}" == "$LEDGER_FORMAT" && "${fields[2]:-}" =~ $UUID_RE && -n "${fields[3]:-}" ]] || return 1
  prev="$(sha_text "$header")"
  while IFS= read -r line; do
    local e; IFS=$'\t' read -r -a e <<<"$line"
    seq=$((seq + 1))
    [[ "${e[0]:-}" == "$seq" && "${e[1]:-}" == "$prev" && ${#e[@]} -eq 7 ]] || return 1
    case "${e[3]}" in object.delete|user.disable|policy.disable|membership.remove|membership.downgrade|profile.unlink) ;; *) return 1 ;; esac
    [[ "${e[5]}" == sync || "${e[5]}" == operator ]] || return 1
    prev="$(sha_text "$line")"
  done < <(tail -n +2 "$f")
  printf '%s\t%s\t%s\t%s\n' "${fields[2]}" "${fields[3]}" "$seq" "$prev"
}

# ledger_append FILE KIND REFS ORIGIN BYREF: one chained entry (seq, previous line digest, UTC time, kind, refs, origin, ref).
ledger_append() {
  local f="$1" kind="$2" refs="$3" origin="$4" by="$5" head seq prev
  head="$(ledger_check "$f")" || die 3 ledger_invalid "ledger header or chain invalid"
  seq="$(cut -f3 <<<"$head")"; prev="$(cut -f4 <<<"$head")"
  local IFS=$'\t'
  printf '%s\n' "$((seq + 1))${IFS}${prev}${IFS}$(now_utc)${IFS}${kind}${IFS}${refs}${IFS}${origin}${IFS}${by}" >>"$f"
  sync -- "$f"
}

# ref_get REFS KEY: one key of an entry's refs (k=v,k=v).
ref_get() { tr ',' '\n' <<<"$1" | sed -n "s/^$2=//p" | head -1; }

# refs_valid KIND REFS: exactly the kind's keys, each once; canonical lowercase UUIDs; a closed-set scope type and roles
# consistent with it; an Organization-scope binding's scope is its Organization; a downgrade lowers the role within its
# scope (owner > admin > member). Checked at append and again at replay (anything else replays as a conflict).
role_rank() { case "${1#*_}" in owner) echo 3 ;; admin) echo 2 ;; member) echo 1 ;; *) echo 0 ;; esac; }
refs_valid() {
  local kind="$1" refs="$2" want pair k v seen=" " n=0 nw st pre r
  case "$kind" in
    object.delete) want="organization object deleter" ;;
    user.disable) want="user at" ;;
    policy.disable) want="organization" ;;
    membership.remove) want="organization subject scope_type scope_id role" ;;
    membership.downgrade) want="organization subject scope_type scope_id from_role to_role" ;;
    profile.unlink) want="organization group profile" ;;
    *) return 1 ;;
  esac
  [[ "$refs" =~ ^[a-z_]+=[0-9A-Za-z_.:-]+(,[a-z_]+=[0-9A-Za-z_.:-]+)*$ ]] || return 1
  for pair in ${refs//,/ }; do
    k="${pair%%=*}"; v="${pair#*=}"
    [[ " $want " == *" $k "* && "$seen" != *" $k "* ]] || return 1
    seen+="$k "; n=$((n + 1))
    case "$k" in
      at) [[ "$v" =~ ^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}(\.[0-9]{1,6})?Z$ ]] || return 1 ;;
      scope_type) [[ "$v" == organization || "$v" == group ]] || return 1 ;;
      role|from_role|to_role) [[ "$v" =~ ^(org|group)_(owner|admin|member)$ ]] || return 1 ;;
      *) [[ "$v" =~ $UUID_RE ]] || return 1 ;;
    esac
  done
  nw=0; for k in $want; do nw=$((nw + 1)); done
  (( n == nw )) || return 1
  [[ "$kind" == membership.* ]] || return 0
  st="$(ref_get "$refs" scope_type)"; pre=org; [[ "$st" == group ]] && pre=group
  for r in role from_role to_role; do v="$(ref_get "$refs" "$r")"; [[ -z "$v" || "$v" == "${pre}_"* ]] || return 1; done
  [[ "$st" == group || "$(ref_get "$refs" scope_id)" == "$(ref_get "$refs" organization)" ]] || return 1
  [[ "$kind" == membership.remove ]] || (( $(role_rank "$(ref_get "$refs" from_role)") > $(role_rank "$(ref_get "$refs" to_role)") ))
}
