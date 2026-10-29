# shellcheck shell=bash
# Environment metadata, lease and schema fence of one staging environment (G072 Plan §10.3-§10.5). Sourced after lib.sh
# by staging.sh and ci-lib.sh; never executed. Uses CONF, GFLAGS, DRY_RUN, REPO and log/die/show/forbid from lib.sh.
#
# The metadata bucket gs://<GCP_PROJECT>-<STAGING_ENV>-ci-meta (created by infra/terraform/bootstrap/ci-federation.sh,
# never by this tooling) holds only this environment's non-secret records: current.tsv, lease, lineage.tsv, fence.tsv,
# schema/<sha>.tsv, deploy/<sha>.tsv, stop-requests/<pipeline>, runs/, receipts/, logs/ (sanitized only), lease-breaks/.
# Every write is `gcloud storage cp ... --if-generation-match=<0|observed>` and every delete `gcloud storage rm ...
# --if-generation-match=<observed>`: no create verb (lib.sh forbid), no unconditional write.
#
# Presence decides whether the environment is CI-enabled: present -> the lease is mandatory for every covered G070
# mutator (provision, deploy, start, stop, reset, retire) and deploys append a schema fence entry; positively absent
# (404) -> today's behaviour, one log line; any other answer (403, an error, a foreign name, an owner project number
# other than GCP_PROJECT's or none) -> refused, exit 2.
#
# Lease (TSV key<TAB>value): holder (local:<user>@<host>:<pid> or ci:<pipeline>/<job>), command, candidate,
# acquired_epoch, expires_epoch (`none` for a local lease), nonce. A CI lease past its expiry belongs to an ended job and
# may be broken by the next CI job (lease-breaks/ record first); a local lease is never broken except by the owner's
# break-glass (staging.sh lease --break --confirm ENV --reason TEXT). Re-entrancy: STAGING_LEASE_HELD=<holder>@<gen>
# lets a command run under a lease its caller holds (exact holder and generation, read back; no acquire or release).

META_STATE=""   # present | absent | dry-run ("" = not checked)
META_BUCKET=""  # bucket name, no gs:// prefix
META_TMP=""     # private scratch of this process (removed by meta_cleanup)
META_ERR=""     # stderr of the last metadata call (classification; never holds an object's content)
META_OUT=""     # stdout of the last metadata call
META_GEN=0      # generation read by meta_read (0 = absent)
LEASE_OWN=0     # 1 when this process created the lease it holds (released at exit)
LEASE_HOLDER="" # the holder this process acts as
LEASE_GEN=""    # the generation of that lease
declare -A LEASE_F=() # the fields of the last lease read

meta_bucket() { printf '%s-%s-ci-meta' "${CONF[GCP_PROJECT]}" "${CONF[STAGING_ENV]}"; }

meta_init() {
  META_BUCKET="$(meta_bucket)"
  ((${#META_BUCKET} <= 63)) || die 2 "the metadata bucket name $META_BUCKET is longer than 63 characters"
  if [[ -z $META_TMP ]] && ((!DRY_RUN)); then META_TMP="$(mktemp -d "${TMPDIR:-/tmp}/staging-meta.XXXXXX")"; fi
}

meta_cleanup() { [[ -z $META_TMP ]] || rm -rf "$META_TMP"; META_TMP=""; }

# meta_call ARGS...: gcloud storage ARGS with GFLAGS (the configured project and impersonated identity). META_OUT and
# META_ERR receive stdout and stderr; returns gcloud's status. The command line is printed (stderr) as every other call.
meta_call() {
  local rc=0
  forbid gcloud storage "$@"
  show gcloud storage "$@" "${GFLAGS[@]}"
  META_OUT="$(gcloud storage "$@" "${GFLAGS[@]}" 2>"$META_TMP/err")" || rc=$?
  META_ERR="$(cat "$META_TMP/err")"
  return "$rc"
}

# meta_class: the failure class of the last call from its stderr: precondition (412), denied (403), notfound (404,
# "matched no objects") or error. 412 and 403 are tested before 404 so a denial is never taken for an absence. The
# gs:// URLs are removed first and a status must stand alone (not inside a word): an object name, a commit SHA or a
# pipeline ID in the message (deploy/<sha>.tsv, stop-requests/<pipeline>) never decides the class.
meta_class() {
  local e
  e="$(sed -E 's#gs://[^[:space:]]*#<url>#g' <<<"$META_ERR")"
  if [[ $e =~ (^|[^0-9A-Za-z])412([^0-9A-Za-z]|$)|PreconditionFailed|[Pp]recondition ]]; then echo precondition
  elif [[ $e =~ (^|[^0-9A-Za-z])403([^0-9A-Za-z]|$)|AccessDenied|[Ff]orbidden|does\ not\ have ]]; then echo denied
  elif [[ $e =~ (^|[^0-9A-Za-z])404([^0-9A-Za-z]|$)|[Nn]ot\ ?[Ff]ound|matched\ no\ objects ]]; then echo notfound
  else echo error; fi
}

meta_why() { local l=${META_ERR%%$'\n'*}; printf '%s' "${l:0:200}"; }

# meta_project_number: META_PNUM = GCP_PROJECT's project number, the owner the metadata bucket must have (Plan §10.5).
# Source: STAGING_META_PROJECT_NUMBER when set (the CI adapter exports the number of its profile's CI_WIF_PROVIDER,
# projects/<number>/..., created in the environment's own project by ci-federation.sh against its pinned number; the
# CI identities cannot read resourcemanager), otherwise - a local operator - gcloud projects describe GCP_PROJECT as the
# configured identity (the deployer). Unreadable, missing or malformed: refused, exit 2 (fail closed).
META_PNUM=""
meta_project_number() {
  local n=${STAGING_META_PROJECT_NUMBER:-} rc=0
  if [[ -z $n ]]; then
    show gcloud projects describe "${CONF[GCP_PROJECT]}" --format 'value(projectNumber)' "${GFLAGS[@]}"
    n="$(gcloud projects describe "${CONF[GCP_PROJECT]}" --format 'value(projectNumber)' "${GFLAGS[@]}" 2>"$META_TMP/err")" || rc=$?
    ((rc == 0)) || die 2 "the project number of ${CONF[GCP_PROJECT]} could not be read to check the owner of gs://$META_BUCKET ($(head -c 200 "$META_TMP/err" | head -n 1)); refused (fail closed; STAGING_META_PROJECT_NUMBER=<number> supplies it)"
    n=${n%%$'\n'*}
  fi
  [[ $n =~ ^[1-9][0-9]{5,19}$ ]] || die 2 "the project number of ${CONF[GCP_PROJECT]} is missing or malformed; refused (fail closed)"
  META_PNUM=$n
}

# meta_presence: META_STATE from the bucket describe (see the header): present only when the answer names exactly this
# bucket and its owner's project number is GCP_PROJECT's (a foreign or unreadable owner is refused, exit 2). A dry run
# reads nothing.
meta_presence() {
  local name num
  meta_init
  if ((DRY_RUN)); then
    # The owner check (meta_project_number) is described in the header; only the presence command line is printed.
    show gcloud storage buckets describe "gs://$META_BUCKET" --raw --format 'value(name,projectNumber)' "${GFLAGS[@]}"
    META_STATE=dry-run
    return 0
  fi
  if meta_call buckets describe "gs://$META_BUCKET" --raw --format 'value(name,projectNumber)'; then
    name=${META_OUT%%$'\n'*} num=""
    [[ $name != *$'\t'* ]] || { num=${name#*$'\t'}; name=${name%%$'\t'*}; }
    [[ $name == "$META_BUCKET" ]] ||
      die 2 "the metadata bucket check answered another name than $META_BUCKET; refused (fail closed)"
    [[ $num =~ ^[1-9][0-9]{0,19}$ ]] ||
      die 2 "gs://$META_BUCKET reports no readable owner project number; refused (fail closed)"
    meta_project_number
    [[ $num == "$META_PNUM" ]] ||
      die 2 "gs://$META_BUCKET belongs to project number $num, not ${CONF[GCP_PROJECT]} ($META_PNUM); refused (fail closed)"
    META_STATE=present
  elif [[ "$(meta_class)" == notfound ]]; then
    META_STATE=absent
  else
    die 2 "the metadata bucket gs://$META_BUCKET could not be checked ($(meta_class): $(meta_why)); refused (fail closed)"
  fi
}

# meta_gen KEY: GEN = the live generation of KEY (0 when absent). Status 2 on any other failure.
GEN=0
meta_gen() {
  GEN=0
  if meta_call objects describe "gs://$META_BUCKET/$1" --format 'value(generation)'; then
    GEN=${META_OUT%%$'\n'*}
    [[ $GEN =~ ^[1-9][0-9]*$ ]] || return 2
    return 0
  fi
  [[ "$(meta_class)" == notfound ]] && return 0
  return 2
}

# meta_read KEY FILE: FILE = KEY's live content (empty when absent), META_GEN = its generation (0 when absent); the
# content is read at exactly that generation. Status 2 on any failure (a denial is never taken for an absence).
meta_read() {
  local key=$1 file=$2
  : >"$file"
  META_GEN=0
  meta_gen "$key" || return 2
  META_GEN=$GEN
  ((META_GEN == 0)) && return 0
  meta_call cat "gs://$META_BUCKET/$key#$META_GEN" || return 2
  [[ -z $META_OUT ]] || printf '%s\n' "$META_OUT" >"$file"
}

# meta_put FILE KEY GEN: write FILE to KEY only if its live generation is GEN (0 = only if absent). Status 0 written,
# 1 precondition failed (412: KEY changed or exists), 2 any other failure (403 included).
meta_put() {
  meta_call cp "$1" "gs://$META_BUCKET/$2" "--if-generation-match=$3" && return 0
  [[ "$(meta_class)" == precondition ]] && return 1
  return 2
}

# meta_rm KEY GEN: delete KEY only at generation GEN. Status as meta_put.
meta_rm() {
  meta_call rm "gs://$META_BUCKET/$1" "--if-generation-match=$2" && return 0
  [[ "$(meta_class)" == precondition ]] && return 1
  return 2
}

# meta_list PREFIX: KEYS = the object names under PREFIX (an empty listing is no failure). Status 2 on failure.
KEYS=()
meta_list() {
  local l
  KEYS=()
  if ! meta_call ls "gs://$META_BUCKET/$1"; then
    [[ "$(meta_class)" == notfound ]] && return 0
    return 2
  fi
  while IFS= read -r l; do [[ $l != "gs://$META_BUCKET/"* ]] || KEYS+=("${l#"gs://$META_BUCKET/"}"); done <<<"$META_OUT"
}

# ---------------------------------------------------------------------------------------------------------------- lease
meta_now() { date -u +%s; }
meta_iso() { date -u -d "@$1" +%Y-%m-%dT%H:%M:%SZ; }

local_holder() {
  local u h
  u="$(id -un 2>/dev/null || echo unknown)" h="$(uname -n 2>/dev/null || echo unknown)"
  printf 'local:%s@%s:%s' "${u//[^A-Za-z0-9._-]/_}" "${h//[^A-Za-z0-9._-]/_}" "$$"
}

valid_holder() { [[ $1 =~ ^(ci:[0-9]{1,20}/[0-9]{1,20}|local:[A-Za-z0-9._-]+@[A-Za-z0-9._-]+(:[0-9]+)?)$ ]]; }

# lease_parse FILE: LEASE_F from a lease record; status 1 when it is not one.
lease_parse() {
  local k v
  LEASE_F=()
  while IFS=$'\t' read -r k v; do [[ -z $k ]] || LEASE_F[$k]=$v; done <"$1"
  valid_holder "${LEASE_F[holder]:-}" && [[ ${LEASE_F[expires_epoch]:-} =~ ^([0-9]+|none)$ ]]
}

lease_describe() {
  local e=${LEASE_F[expires_epoch]:-?} a=${LEASE_F[acquired_epoch]:-}
  [[ $e == none || ! $e =~ ^[0-9]+$ ]] || e="$(meta_iso "$e")"
  [[ ! $a =~ ^[0-9]+$ ]] || a="$(meta_iso "$a")"
  printf 'held by %s (command %s, candidate %s, acquired %s, expires %s)' "${LEASE_F[holder]:-?}" \
    "${LEASE_F[command]:-?}" "${LEASE_F[candidate]:--}" "${a:-?}" "$e"
}

# lease_expired_ci: the lease last parsed is a CI lease past its expiry (only such a lease may be broken by CI).
lease_expired_ci() {
  local e=${LEASE_F[expires_epoch]:-}
  [[ ${LEASE_F[holder]:-} == ci:* && $e =~ ^[0-9]+$ ]] && (("$(meta_now)" >= e))
}

# lease_read: the live lease into LEASE_F with its generation in META_GEN (0 = free). Dies 2 when it cannot be read or
# is not a lease record.
lease_read() {
  meta_read lease "$META_TMP/lease.cur" || die 2 "the environment lease could not be read ($(meta_class): $(meta_why)); refused"
  LEASE_F=()
  ((META_GEN == 0)) && return 0
  lease_parse "$META_TMP/lease.cur" || die 2 "gs://$META_BUCKET/lease (generation $META_GEN) is not a lease record; refused (break-glass: staging.sh lease --break)"
}

# lease_acquire HOLDER COMMAND CANDIDATE TTL: create the lease (generation 0 only); TTL seconds, or empty for a local
# lease (no expiry). A held lease refuses (exit 2) naming its holder and times, before any other call.
lease_acquire() {
  local holder=$1 cmd=$2 cand=${3:--} ttl=${4:-} now exp=none f rc
  valid_holder "$holder" || die 2 "lease holder $holder is not ci:<pipeline>/<job> or local:<user>@<host>[:<pid>]"
  if ((DRY_RUN)); then
    [[ $holder != local:* ]] || holder='local:<user>@<host>:<pid>' # stable dry-run text
    printf '+ (gs://%s present: gcloud storage cp <lease: %s> gs://%s/lease --if-generation-match=0, released at exit; absent (404): no lease; any other answer: refused)\n' \
      "$META_BUCKET" "$holder" "$META_BUCKET"
    return 0
  fi
  now="$(meta_now)"
  [[ -z $ttl ]] || exp=$((now + ttl))
  f=$META_TMP/lease.new
  printf 'holder\t%s\ncommand\t%s\ncandidate\t%s\nacquired_epoch\t%s\nexpires_epoch\t%s\nnonce\t%s\n' "$holder" "$cmd" "$cand" \
    "$now" "$exp" "$(od -An -N8 -tx1 /dev/urandom | tr -d ' \n')" >"$f"
  rc=0
  meta_put "$f" lease 0 || rc=$?
  case $rc in
    0) ;;
    1)
      lease_read
      ((META_GEN == 0)) && die 2 "the environment lease changed while it was acquired; refused (retry)"
      die 2 "the environment lease is $(lease_describe); nothing changed"
      ;;
    *) die 2 "the environment lease could not be created ($(meta_class): $(meta_why)); nothing changed" ;;
  esac
  # Read back: the live lease must be exactly the record written (its nonce), and its generation is kept for release.
  meta_read lease "$META_TMP/lease.cur" || die 2 "the new lease could not be read back; refused"
  cmp -s <(grep . "$f") <(grep . "$META_TMP/lease.cur") || die 2 "the lease read back is not the one written; refused"
  LEASE_OWN=1 LEASE_HOLDER=$holder LEASE_GEN=$META_GEN
  log "lease: acquired gs://$META_BUCKET/lease as $holder (generation $LEASE_GEN)"
}

# lease_check_held SPEC: SPEC = <holder>@<generation> (STAGING_LEASE_HELD): the live lease must have exactly that holder
# and generation. Nothing is acquired or released.
lease_check_held() {
  local spec=$1 holder gen
  gen=${spec##*@} holder=${spec%@*}
  [[ $spec == *@* && $gen =~ ^[1-9][0-9]*$ ]] && valid_holder "$holder" ||
    die 2 "STAGING_LEASE_HELD must be <holder>@<generation>"
  if ((DRY_RUN)); then
    printf '+ (gs://%s present: the live lease must be %s at generation %s; nothing acquired or released)\n' "$META_BUCKET" "$holder" "$gen"
    return 0
  fi
  lease_read
  ((META_GEN == gen)) && [[ ${LEASE_F[holder]} == "$holder" ]] ||
    die 2 "STAGING_LEASE_HELD names $holder at generation $gen, but the live lease is $( ((META_GEN == 0)) && echo free || { lease_describe; printf ' at generation %s' "$META_GEN"; }); nothing changed"
  LEASE_OWN=0 LEASE_HOLDER=$holder LEASE_GEN=$gen
  log "lease: running under $holder (generation $gen)"
}

# lease_release_own: delete the lease this process created, at its generation. Never dies (called from the EXIT trap).
lease_release_own() {
  ((LEASE_OWN)) || return 0
  LEASE_OWN=0
  if meta_rm lease "$LEASE_GEN"; then
    log "lease: released gs://$META_BUCKET/lease (generation $LEASE_GEN)"
  else
    log "lease: gs://$META_BUCKET/lease (generation $LEASE_GEN) could not be released ($(meta_class)); release it with staging.sh lease --release"
    return 1
  fi
}

# lease_break OPERATOR REASON: the create-only lease-breaks/<utc>-<holder>.tsv record (operator, reason, the broken
# lease's content and generation), then the delete of exactly that generation. The caller read the lease (LEASE_F,
# META_GEN, META_TMP/lease.cur) and decided it may be broken.
lease_break() {
  local op=$1 reason=$2 gen=$META_GEN key f rc=0
  ((gen > 0)) || die 2 "the lease is free; nothing to break"
  key="lease-breaks/$(date -u +%Y%m%dT%H%M%SZ)-${LEASE_F[holder]//[^A-Za-z0-9._-]/_}.tsv"
  f=$META_TMP/lease.break
  { printf 'operator\t%s\nreason\t%s\ntime\t%s\nbroken_generation\t%s\n' "$op" "$reason" "$(ts)" "$gen"
    sed 's/^/lease./' "$META_TMP/lease.cur"; } >"$f"
  meta_put "$f" "$key" 0 || die 2 "the lease-break record $key could not be written ($(meta_class)); the lease is kept"
  meta_rm lease "$gen" || rc=$?
  ((rc == 0)) || die 2 "the lease changed or could not be deleted after the break record $key ($(meta_class)); nothing else changed"
  log "lease: broke the lease of ${LEASE_F[holder]} (generation $gen); record gs://$META_BUCKET/$key"
}

# lease_guard COMMAND CANDIDATE: staging.sh main, before a covered mutator: the presence check, then the lease (or the
# STAGING_LEASE_HELD check) when the bucket exists; one log line and nothing else when it is positively absent.
lease_guard() {
  meta_presence
  case $META_STATE in
    absent)
      [[ -z ${STAGING_LEASE_HELD:-} ]] || die 2 "STAGING_LEASE_HELD is set but gs://$META_BUCKET does not exist; refused"
      log "ci-meta: gs://$META_BUCKET does not exist (404): $ENV_NAME is not CI-enabled; no lease, no schema fence"
      return 0
      ;;
  esac
  if [[ -n ${STAGING_LEASE_HELD:-} ]]; then
    lease_check_held "$STAGING_LEASE_HELD"
  else
    lease_acquire "$(local_holder)" "$1" "${2:--}" ""
  fi
}

# --------------------------------------------------------------------------------------------------------------- schema
# schema_of DIR SHA: the candidate's core/migrations version -> blob map: "candidate<TAB>SHA", then one
# "<version><TAB><blob ID>" row per NNNNN_*.sql file, by version (deterministic: byte-equal for every writer).
schema_of() {
  local dir=$1 sha=$2 out meta path f
  out="$(git -C "$dir" ls-tree "$sha" core/migrations/)" || return 1
  printf 'candidate\t%s\n' "$sha"
  while IFS=$'\t' read -r meta path; do
    f=${path##*/}
    [[ $meta == *" blob "* && $f =~ ^([0-9]+)_[^/]*\.sql$ ]] || continue
    printf '%s\t%s\n' "$((10#${BASH_REMATCH[1]}))" "${meta##* }"
  done <<<"$out" | sort -n -k1,1
}

# schema_publish FILE SHA: schema/<SHA>.tsv is create-only; an existing object must be byte-equal to FILE.
schema_publish() {
  local f=$1 sha=$2 rc=0
  meta_read "schema/$sha.tsv" "$META_TMP/schema.cur" || return 2
  if ((META_GEN == 0)); then
    meta_put "$f" "schema/$sha.tsv" 0 || rc=$?
    ((rc == 1)) || return "$rc"
    meta_read "schema/$sha.tsv" "$META_TMP/schema.cur" || return 2
  fi
  cmp -s "$f" "$META_TMP/schema.cur" || return 3
}

# fence_append SHA: before the release upload of a deploy (staging.sh deploy_candidate), under the lease: publish the
# candidate's schema and append "<time> <holder> <candidate> <schema sha256>" to fence.tsv at its observed generation.
# Dies before any upload when it cannot (exit 2, or 4 inside provision after an apply).
fence_append() {
  local sha=$1 code=2 sf rc=0
  [[ $COMMAND != provision ]] || code=4
  if ((DRY_RUN)); then
    printf '+ (gs://%s present: create-only gs://%s/schema/%s.tsv, then an entry appended to gs://%s/fence.tsv at its observed generation, before the upload)\n' \
      "$META_BUCKET" "$META_BUCKET" "$sha" "$META_BUCKET"
    return 0
  fi
  [[ $META_STATE == present ]] || return 0
  [[ -n $LEASE_HOLDER ]] || die "$code" "schema fence: no lease held; nothing uploaded"
  sf=$META_TMP/schema.new
  schema_of "$REPO" "$sha" >"$sf" || die "$code" "schema fence: the migrations of $sha could not be listed; nothing uploaded"
  schema_publish "$sf" "$sha" || rc=$?
  case $rc in
    0) ;;
    3) die "$code" "schema fence: gs://$META_BUCKET/schema/$sha.tsv exists and differs from $sha's migrations; nothing uploaded" ;;
    *) die "$code" "schema fence: gs://$META_BUCKET/schema/$sha.tsv could not be written ($(meta_class)); nothing uploaded" ;;
  esac
  meta_read fence.tsv "$META_TMP/fence.cur" || die "$code" "schema fence: fence.tsv could not be read ($(meta_class)); nothing uploaded"
  grep . "$META_TMP/fence.cur" >"$META_TMP/fence.new" || true
  printf 'fence\t%s\t%s\t%s\t%s\n' "$(ts)" "$LEASE_HOLDER" "$sha" "$(sha256sum "$sf" | cut -d' ' -f1)" >>"$META_TMP/fence.new"
  meta_put "$META_TMP/fence.new" fence.tsv "$META_GEN" ||
    die "$code" "schema fence: fence.tsv could not be written at generation $META_GEN ($(meta_class)); nothing uploaded"
  log "fence: $sha appended to gs://$META_BUCKET/fence.tsv"
}
