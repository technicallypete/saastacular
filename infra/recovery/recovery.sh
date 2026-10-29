#!/usr/bin/env bash
# G013 recovery, retention and incident operator commands (docs/foundation/RECOVERY.md; Plan r3 §3-§7).
#
#   recovery.sh fence    --env-file FILE --backup-root DIR --admin-window REF
#   recovery.sh unfence  --env-file FILE --backup-root DIR [--fence-id ID]
#   recovery.sh backup   --env-file FILE --backup-root DIR --ledger FILE --admin-window REF [--use-fence] [--backup-id ID]
#                        [--timeout S]
#   recovery.sh verify   --backup DIR/ID
#   recovery.sh restore  --backup DIR/ID --target-env-file FILE --receipts DIR --backup-policy FILE [--source-env-file FILE]
#                        [--steps all|database|objects] [--timeout S]
#   recovery.sh invalidate-auth --target-env-file FILE --receipts DIR
#   recovery.sh ledger init|append|sync|verify|attest ...   (see RECOVERY.md)
#   recovery.sh replay   --backup DIR/ID --ledger FILE --target-env-file FILE --receipts DIR
#   recovery.sh reopen   --backup DIR/ID --ledger FILE --receipts DIR --sign-off REF --runtime-verification pass
#   recovery.sh retention audit|objects --policy FILE --env-file FILE --preview FILE [--apply --receipts DIR] [--ledger FILE]
#   recovery.sh status   --backup-root DIR --backup-policy FILE --receipts DIR [--env-file FILE] [--ledger FILE] [--sink FILE]
#   recovery.sh quarantine --backup DIR/ID --reason CODE
#   recovery.sh expire   --backup-root DIR --backup-policy FILE [--apply]
#
# Privileged local tool: the cluster superuser (over the PostgreSQL container's local socket) for capture, restore,
# invalidation, replay and ledger sync; the fixture's MinIO administrator for the capture fence and the capture; the
# scoped object service user for restore uploads, replay deletes and purge checks; the retention login (00009) for
# retention. Nothing is given to core, web or intelligence. Raw sets, ledgers, receipts and env files are private
# material (0700/0600, outside the repository); only sanitized lines go to stdout. Archives travel on stdin/stdout only;
# nothing is bind-mounted. Exit: 0 success; 1 unhealthy (status) or a held/incomplete outcome; 2 invalid input; 3 a
# refusal before any change of the target; 4 a failure after a change started (the receipt says what happened).
set -euo pipefail
umask 077
# shellcheck source=lib.sh
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

declare -A OPT=()
ARGS=()
parse_opts() { # parse_opts ARGS...: --key value pairs into OPT; --use-fence and --apply are flags
  while (($#)); do
    case "$1" in
      --use-fence|--apply) OPT[${1#--}]=1; shift ;;
      --*) [[ $# -ge 2 && "$2" != --* ]] || die 2 invalid_input "option $1 needs a value"; OPT[${1#--}]="$2"; shift 2 ;;
      *) ARGS+=("$1"); shift ;;
    esac
  done
}
need() { local k; for k in "$@"; do [[ -n "${OPT[$k]:-}" ]] || die 2 invalid_input "missing --$k"; done; }

# Exit guard (Plan r3 §3.2; S-review F5): while a fence is being established or a capture runs, any exit that did not
# go through fence_abort or capture_fail (a set -e failure, a failed record write) still releases what this command
# established and marks the set incomplete. GUARD is fence, capture-integrated or capture-external; empty otherwise.
GUARD=""
guard_exit() {
  local rc=$?
  [[ -n "$GUARD" ]] || return 0
  # Every step below is independent (Codex review C2): errexit is off, so a failed marker or log write never skips the
  # release, and the outcome stays non-zero whatever the cleanup does.
  set +e
  local g="$GUARD"; GUARD=""
  trap - TERM INT HUP
  log "unexpected exit ($rc) during the ${g%%-*}; releasing what this command established"
  if [[ -n "${SET_DIR:-}" && -d "${SET_DIR:-}" && ! -e "$SET_DIR/COMPLETE" ]]; then
    { printf 'status\tincomplete\nreason\tunexpected_exit\n' >"$SET_DIR/INCOMPLETE"; } 2>/dev/null || log "the INCOMPLETE marker could not be written"
  fi
  if [[ "$g" != capture-external && -n "${FENCE_FILE:-}" && -f "${FENCE_FILE:-}" ]]; then
    fence_release "$FENCE_FILE" "${FENCE_ENV:-}" >/dev/null 2>&1
    # The actual fence state decides what is reported, not the release's return.
    if fence_has "$FENCE_FILE" released 2>/dev/null; then log "fence $(fence_get "$FENCE_FILE" fence_id) released (verified)"
    else log "fence $(fence_get "$FENCE_FILE" fence_id 2>/dev/null) still held: run unfence"; fi
  fi
  (( rc != 0 )) || rc=4
  exit "$rc"
}
trap guard_exit EXIT

# lock_root ROOT: one fence or backup per backup root at a time (S-review F14), held until this command exits.
ROOT_LOCKED=""
lock_root() {
  [[ "$ROOT_LOCKED" == "$1" ]] && return 0
  mkdir -p "$1/fences"; chmod 700 "$1/fences"
  exec 9>>"$1/fences/.lock"
  flock -n 9 || die 3 fence_busy "another fence or backup of this backup root is running"
  ROOT_LOCKED="$1"
}

# with_deadline SECONDS: re-run this command under one overall bound (TERM at the bound; the traps release what this
# command established).
with_deadline() {
  [[ -n "${RECOVERY_DEADLINE:-}" ]] && return 0
  int_in "$1" 1 7200 || die 2 invalid_input "--timeout must be 1-7200 seconds"
  RECOVERY_DEADLINE="$1" exec timeout --signal=TERM --kill-after=30 "$1" bash "${BASH_SOURCE[0]}" "${ORIG_ARGV[@]}"
}

# ======================================================================================================================
# Fence (Plan r3 §3.2): writer exclusion and drain held for the whole capture, recorded before any change.
# ======================================================================================================================

FENCE_FILE="" FENCE_ID="" FENCE_HOLDER=""

state_of() { awk -F'\t' -v k="$2" '$1 == k { print $2; exit }' <<<"$1"; }
fence_state() { sqlf "$SRC_PG" "$SRC_DB" fence-state.sql -v holder="$1"; }

user_status() { # user_status ENVFILE: enabled|disabled of the scoped service user (fixture administrator view)
  local out
  out="$(mcx "$SRC_NET" "$SRC_MCIMG" "$1" admin "mc admin user info --json x '$(envv "$1" OBJECT_STORE_ACCESS_KEY)'")" || return 1
  sed -n 's/.*"userStatus":"\([a-z]*\)".*/\1/p' <<<"$out" | head -1
}
svc_refused() { ! mcx "$SRC_NET" "$SRC_MCIMG" "$1" svc "mc ls 'x/$SRC_BUCKET/'" >/dev/null 2>&1; }

# fence_establish ENVFILE ROOT ADMINREF MAXDUR: X0-X5; sets FENCE_FILE/FENCE_ID. Any failure releases what was done
# and exits non-zero with fence_not_established (or the specific refusal).
fence_establish() {
  local envf="$1" root="$2" admin="$3" maxdur="$4" st other grantees status t0 t1 cid n i
  [[ -n "$admin" ]] || die 3 admin_window_unavailable "no trusted administrator-exclusive window recorded (--admin-window)"
  [[ "$admin" =~ ^[A-Za-z0-9._:/-]{1,128}$ ]] || die 2 invalid_input "--admin-window must be a short non-confidential reference"
  int_in "$maxdur" 1 7200 || die 2 invalid_input "--max-duration must be 1-7200 seconds"
  lock_root "$root"
  FENCE_ID="$(new_uuid)"; FENCE_HOLDER="g013-fence-$FENCE_ID"
  st="$(fence_state "$FENCE_HOLDER")" || die 3 source_unavailable "source PostgreSQL unreadable"
  other="$(state_of "$st" other_superusers)"
  [[ "$other" == 0 ]] || die 3 window_changed "another superuser session is open at X0: the administrator window is contradicted"
  # PUBLIC CONNECT would survive the per-grantee revoke (S-review F6): refuse before any change.
  [[ "$(state_of "$st" public_connect)" == 0 ]] || die 3 fence_not_established "PUBLIC holds CONNECT on the database"
  grantees="$(state_of "$st" grantees)"
  status="$(user_status "$envf")" || die 3 source_unavailable "source MinIO administration unreachable"
  [[ "$status" == enabled || "$status" == disabled ]] || die 3 source_unavailable "service user status unknown"
  mkdir -p "$root/fences"; chmod 700 "$root/fences"
  # An unreleased fence of this source must be released first: a new record would take the fenced state as "prior".
  local of
  for of in "$root"/fences/*.tsv; do
    [[ -f "$of" ]] || continue
    fence_has "$of" released && continue
    [[ "$(fence_get "$of" project)" == "$SRC_PROJECT" ]] && die 3 fence_held "an unreleased fence of this source exists: run unfence first"
  done
  FENCE_FILE="$root/fences/$FENCE_ID.tsv"
  [[ ! -e "$FENCE_FILE" ]] || die 3 fence_not_established "fence record exists"
  {
    printf 'format\t%s\n' "$FENCE_FORMAT"; printf 'fence_id\t%s\n' "$FENCE_ID"; printf 'created\t%s\n' "$(now_utc)"
    printf 'project\t%s\n' "$SRC_PROJECT"; printf 'pg_container\t%s\n' "$SRC_PG"; printf 'pg_volume\t%s\n' "$SRC_PGVOL"
    printf 'sysid\t%s\n' "$(state_of "$st" sysid)"; printf 'minio_container\t%s\n' "$SRC_MINIO"; printf 'minio_volume\t%s\n' "$SRC_MINIOVOL"
    printf 'network\t%s\n' "$SRC_NET"; printf 'prior_datacl\t%s\n' "$(state_of "$st" datacl)"
    printf 'prior_datacl_norm\t%s\n' "$(state_of "$st" datacl_norm)"; printf 'prior_grantees\t%s\n' "$grantees"
    printf 'prior_user_status\t%s\n' "$status"; printf 'admin_window\t%s\n' "$admin"; printf 'admin_operator\t%s\n' "$(id -un)"
    printf 'admin_window_start\t%s\n' "$(now_utc)"; printf 'holder_app\t%s\n' "$FENCE_HOLDER"; printf 'max_duration_s\t%s\n' "$maxdur"
    printf 'steps\tapp_stopped,db_revoked,locks_held,user_disabled,provider_restarted,window_open\n'
    printf 'trace\tnot_used (optional diagnostics, r3)\n'
  } >"$FENCE_FILE.tmp"
  sync -- "$FENCE_FILE.tmp"; mv -- "$FENCE_FILE.tmp" "$FENCE_FILE"; GUARD=fence; sync -- "$root/fences"
  [[ "$(fence_get "$FENCE_FILE" fence_id)" == "$FENCE_ID" && "$(fence_get "$FENCE_FILE" prior_user_status)" == "$status" ]] ||
    die 4 fence_not_established "fence record read-back failed"
  log "fence $FENCE_ID recorded (window $admin)"
  trap 'fence_abort interrupted' TERM INT HUP
  # X1 drain the application.
  local stopped=""
  for svc in core web intelligence; do
    cid="$(cid_of "$SRC_PROJECT" "$svc")"
    [[ -n "$cid" ]] || continue
    timeout 125 docker stop -t 110 "$cid" >/dev/null 2>&1 || fence_abort writers_not_stopped
    [[ "$(docker inspect -f '{{.State.Status}}' "$cid")" == exited ]] || fence_abort writers_not_stopped
    stopped+="$svc=$(docker inspect -f '{{.State.ExitCode}}' "$cid") "
  done
  fence_put "$FENCE_FILE" app_stopped "${stopped:-none}"
  # X2 database fence.
  sqlf "$SRC_PG" "$SRC_DB" fence-revoke.sql -v grantees="$grantees" >/dev/null || fence_abort fence_not_established
  for ((i = 0; i < 30; i++)); do
    n="$(state_of "$(fence_state "$FENCE_HOLDER")" nonsuper_backends)"
    [[ "$n" == 0 ]] && break; sleep 1
  done
  [[ "$n" == 0 ]] || fence_abort fence_not_established
  # No project login may still hold CONNECT, directly, through PUBLIC or through a role it inherits (S-review F6).
  [[ "$(state_of "$(fence_state "$FENCE_HOLDER")" connectable_logins)" == 0 ]] || fence_abort fence_not_established
  fence_put "$FENCE_FILE" db_revoked "$grantees"
  # Drill seam (inert unless set; never used in operation): a bounded pause between X2 and X3, so the validation drill
  # can start an administrator write after the X0 window check and exercise the X3 lock-wait refusal (Plan r3 RD-B(1)).
  if int_in "${RECOVERY_DRILL_PAUSE_AFTER_X2:-0}" 1 120; then sleep "$RECOVERY_DRILL_PAUSE_AFTER_X2"; fi
  # X3 database drain and write lock: a superuser holder takes SHARE on every table of app and audit.
  local tables ntab cidarg=() hpw
  st="$(fence_state "$FENCE_HOLDER")"; tables="$(state_of "$st" table_list)"; ntab="$(state_of "$st" tables)"
  [[ -n "${RECOVERY_CID_DIR:-}" ]] && cidarg=(--cidfile "$RECOVERY_CID_DIR/$FENCE_ID-holder-$(new_uuid).cid")
  hpw="$(envv "$envf" POSTGRES_SUPERUSER_PASSWORD)"
  [[ -n "$hpw" ]] || fence_abort fence_not_established
  docker run -d "${LABELS[@]}" --label "g013.fence=$FENCE_ID" --label g013.role=holder "${cidarg[@]}" --network "$SRC_NET" \
    --env-file <(printf 'PGPASSWORD=%s\nPGAPPNAME=%s\n' "$hpw" "$FENCE_HOLDER") "$SRC_PGIMG" \
    psql -X -q -v ON_ERROR_STOP=1 -h postgres -U postgres -d "$SRC_DB" \
    -c "BEGIN; SET LOCAL lock_timeout = '30s'; LOCK TABLE $tables IN SHARE MODE; SELECT pg_sleep($maxdur); COMMIT;" >/dev/null ||
    fence_abort fence_not_established
  local hc; hc="$(docker ps -aq --filter "label=g013.fence=$FENCE_ID" --filter label=g013.role=holder | head -1)"
  for ((i = 0; i < 70; i++)); do
    st="$(fence_state "$FENCE_HOLDER")"
    [[ "$(state_of "$st" holder_locks)" == "$ntab" ]] && break
    if [[ "$(docker inspect -f '{{.State.Running}}' "$hc" 2>/dev/null)" != true ]]; then
      docker logs "$hc" 2>&1 | grep -q 'lock timeout' && fence_abort admin_write_in_progress
      fence_abort fence_not_established
    fi
    sleep 0.5
  done
  [[ "$(state_of "$st" holder_locks)" == "$ntab" ]] || fence_abort admin_write_in_progress
  fence_put "$FENCE_FILE" locks_held "$(state_of "$st" holder)" "$hc"
  # X4 object fence.
  if [[ "$status" == enabled ]]; then
    mcx "$SRC_NET" "$SRC_MCIMG" "$envf" admin "mc admin user disable x '$(envv "$envf" OBJECT_STORE_ACCESS_KEY)'" >/dev/null 2>&1 ||
      fence_abort fence_not_established
  fi
  [[ "$(user_status "$envf")" == disabled ]] && svc_refused "$envf" || fence_abort fence_not_established
  fence_put "$FENCE_FILE" user_disabled "prior=$status"
  # X5 provider drain: graceful stop until exited, start, ready.
  t0="$(docker inspect -f '{{.State.StartedAt}}' "$SRC_MINIO")"
  timeout 125 docker stop -t 110 "$SRC_MINIO" >/dev/null 2>&1 || fence_abort fence_not_established
  [[ "$(docker inspect -f '{{.State.Status}}' "$SRC_MINIO")" == exited ]] || fence_abort fence_not_established
  timeout 60 docker start "$SRC_MINIO" >/dev/null 2>&1 || fence_abort fence_not_established
  minio_ready "$SRC_MINIO" || fence_abort fence_not_established
  t1="$(docker inspect -f '{{.State.StartedAt}}' "$SRC_MINIO")"
  [[ "$t0" != "$t1" ]] || fence_abort fence_not_established
  svc_refused "$envf" || fence_abort fence_not_established
  fence_put "$FENCE_FILE" provider_restarted "$t0" "$t1"
  fence_put "$FENCE_FILE" window_open "$(now_utc)"
  trap - TERM INT HUP
  GUARD=""
  log "fence $FENCE_ID established"
}

fence_abort() { # fence_abort REASON: release only a fence this command established, mark the set incomplete, exit non-zero
  trap - TERM INT HUP
  # Ownership is read before anything is cleared (MR !29 D1): a signal, timeout or refusal during a capture under an
  # operator-established fence (--use-fence: GUARD capture-external, fence_mode external) never releases that fence; it
  # stays held until an explicit unfence, exactly as guard_exit and capture_fail leave it. A fence this command
  # established (fence, or an integrated capture) is released as before. The steps below are independent (errexit
  # off): a failed release, log or marker write never skips the rest, and the outcome stays non-zero.
  set +e
  local g="$GUARD" owned=1
  GUARD=""
  [[ "$g" == capture-external || "${fence_mode:-}" == external ]] && owned=0
  if [[ $owned == 1 ]]; then
    log "fence ${FENCE_ID:-?} aborted: $1; releasing"
    [[ -n "$FENCE_FILE" && -f "$FENCE_FILE" ]] && fence_release "$FENCE_FILE" "${FENCE_ENV:-}" >/dev/null 2>&1
    if [[ -n "$FENCE_FILE" ]] && fence_has "$FENCE_FILE" released 2>/dev/null; then log "fence $FENCE_ID released (verified)"
    elif [[ -n "$FENCE_FILE" ]]; then log "fence $FENCE_ID still held: run unfence"; fi
  else
    log "capture aborted: $1; the external fence ${FENCE_ID:-?} was not established by this command and stays held (release: deferred to unfence)"
  fi
  if [[ -n "${SET_DIR:-}" && -d "${SET_DIR:-}" && ! -e "$SET_DIR/COMPLETE" ]]; then
    { printf 'status\tincomplete\nreason\t%s\n' "$1" >"$SET_DIR/INCOMPLETE"; } 2>/dev/null || log "the INCOMPLETE marker could not be written"
  fi
  die 4 "$1" "fence not established or capture stopped"
}

# fence_release FILE ENVFILE: converge every element to the recorded prior state (idempotent; safe after any interruption,
# including of an earlier release), verify, append released. Non-zero when an element cannot be restored.
fence_release() {
  local f="$1" envf="$2" id grantees prior status st norm ok=1
  id="$(fence_get "$f" fence_id)"; grantees="$(fence_get "$f" prior_grantees)"; prior="$(fence_get "$f" prior_user_status)"
  [[ -n "$id" && -n "$prior" ]] || return 1
  [[ "$(fence_get "$f" project)" == "$SRC_PROJECT" && "$(fence_get "$f" pg_container)" == "$SRC_PG" ]] || { log "source identity differs from the fence record"; return 1; }
  [[ "$(sql1 "$SRC_PG" "$SRC_DB" "SELECT system_identifier FROM pg_control_system()")" == "$(fence_get "$f" sysid)" ]] || return 1
  # Holder: end its transaction (terminate the backend; removing the container alone does not, b0), remove its containers.
  sqlf "$SRC_PG" "$SRC_DB" fence-restore.sql -v holder="g013-fence-$id" -v grantees="-" >/dev/null || ok=0
  docker ps -aq --filter "label=g013.fence=$id" | xargs -r docker rm -f -v >/dev/null 2>&1 || true
  [[ "$(docker ps -aq --filter "label=g013.fence=$id" | wc -l)" == 0 ]] && fence_put "$f" holder_released "$(now_utc)" || ok=0
  # Service user: exactly its recorded prior status (a user disabled before the fence stays disabled).
  status="$(user_status "$envf" 2>/dev/null || true)"
  if [[ "$prior" == enabled && "$status" != enabled ]]; then
    mcx "$SRC_NET" "$SRC_MCIMG" "$envf" admin "mc admin user enable x '$(envv "$envf" OBJECT_STORE_ACCESS_KEY)'" >/dev/null 2>&1 || ok=0
    status="$(user_status "$envf" 2>/dev/null || true)"
  fi
  [[ "$status" == "$prior" ]] && fence_put "$f" user_restored "$status" || ok=0
  # Database ACL: re-grant only the recorded prior grantees; verify the normalized ACL equals the record.
  sqlf "$SRC_PG" "$SRC_DB" fence-restore.sql -v holder="g013-fence-$id" -v grantees="$grantees" >/dev/null || ok=0
  st="$(sqlf "$SRC_PG" "$SRC_DB" fence-state.sql -v holder="g013-fence-$id")" || ok=0
  norm="$(state_of "$st" datacl_norm)"
  [[ "$norm" == "$(fence_get "$f" prior_datacl_norm)" && "$(state_of "$st" holder)" == - ]] && fence_put "$f" acl_restored "$norm" || ok=0
  [[ $ok == 1 ]] || return 1
  fence_put "$f" admin_window_end "$(now_utc)"
  fence_put "$f" released "$(now_utc)"
}

open_fence() { # open_fence ROOT: the record of an established, unreleased fence of this source (one), or empty
  local f
  for f in "$1"/fences/*.tsv; do
    [[ -f "$f" ]] || continue
    fence_has "$f" released && continue
    fence_has "$f" window_open || continue
    [[ "$(fence_get "$f" pg_container)" == "$SRC_PG" ]] && { printf '%s' "$f"; return 0; }
  done
}

# window_checks FILE DATAFINGERPRINT: check (a), the PostgreSQL safeguard, plus no other superuser backend.
window_checks() {
  local f="$1" st
  st="$(sqlf "$SRC_PG" "$SRC_DB" fence-state.sql -v holder="$(fence_get "$f" holder_app)")" || return 1
  [[ "$(state_of "$st" holder)" == "$(fence_get "$f" locks_held | cut -f1)" ]] || { log "holder backend or transaction changed"; return 1; }
  [[ "$(state_of "$st" holder_locks)" == "$(state_of "$st" tables)" ]] || { log "SHARE locks no longer all granted"; return 1; }
  [[ "$(state_of "$st" other_superusers)" == 0 ]] || { log "another superuser session appeared"; return 1; }
}

cmd_fence() {
  need env-file backup-root
  resolve SRC "${OPT[env-file]}"; check_env_file "${OPT[env-file]}" POSTGRES_SUPERUSER_PASSWORD MINIO_ROOT_USER MINIO_ROOT_PASSWORD OBJECT_STORE_ACCESS_KEY
  private_dir "${OPT[backup-root]}"
  recovery_labels fence; FENCE_ENV="${OPT[env-file]}"
  fence_establish "${OPT[env-file]}" "${OPT[backup-root]}" "${OPT[admin-window]:-}" "${OPT[max-duration]:-1800}"
  printf 'fence\t%s\testablished\n' "$FENCE_ID"
}

cmd_unfence() {
  need env-file backup-root
  resolve SRC "${OPT[env-file]}"; recovery_labels unfence
  local f rc=0 found=0
  for f in "${OPT[backup-root]}"/fences/*.tsv; do
    [[ -f "$f" ]] || continue
    [[ -z "${OPT[fence-id]:-}" || "$(fence_get "$f" fence_id)" == "${OPT[fence-id]}" ]] || continue
    fence_has "$f" released && continue
    [[ "$(fence_get "$f" project)" == "$SRC_PROJECT" ]] || continue
    found=1
    if fence_release "$f" "${OPT[env-file]}"; then printf 'fence\t%s\treleased\n' "$(fence_get "$f" fence_id)"
    else printf 'fence\t%s\trelease_failed\n' "$(fence_get "$f" fence_id)"; rc=4; fi
  done
  [[ $found == 1 ]] || printf 'fence\tnone\tnothing to release\n'
  [[ $rc == 0 ]] || die 4 unfence_failed "a fence element could not be restored; rerun unfence"
}

# ======================================================================================================================
# Backup (Plan r3 §3.2 C1-C7, §3.3)
# ======================================================================================================================

fingerprint() { # fingerprint SIDE OUT: data and catalog fingerprints of SIDE's database
  local pg db
  case "$1" in SRC) pg="$SRC_PG" db="$SRC_DB" ;; TGT) pg="$TGT_PG" db="$TGT_DB" ;; *) return 1 ;; esac
  { sqlf "$pg" "$db" fingerprint-data.sql; sqlf "$pg" "$db" fingerprint-catalog.sql -v db="$db"; } >"$2"
}

cmd_backup() {
  need env-file backup-root ledger
  local envf="${OPT[env-file]}" root="${OPT[backup-root]}" ledger="${OPT[ledger]}" id t_start fence_mode release=deferred
  with_deadline "${OPT[timeout]:-1800}"
  check_env_file "$envf" POSTGRES_SUPERUSER_PASSWORD MINIO_ROOT_USER MINIO_ROOT_PASSWORD OBJECT_STORE_ACCESS_KEY
  [[ -n "${OPT[use-fence]:-}" || -n "${OPT[admin-window]:-}" ]] ||
    die 3 admin_window_unavailable "no trusted administrator-exclusive window recorded (--admin-window)"
  resolve SRC "$envf"; recovery_labels backup; FENCE_ENV="$envf"
  ledger_check "$ledger" >/dev/null || die 3 ledger_invalid "ledger missing or invalid (ledger init first)"
  private_dir "$root"
  lock_root "$root"
  id="${OPT[backup-id]:-$(new_uuid)}"; [[ "$id" =~ $UUID_RE ]] || die 2 invalid_input "--backup-id must be a UUID"
  SET_DIR="$root/$id"
  mkdir -- "$SET_DIR" 2>/dev/null || { SET_DIR=""; die 3 destination_unavailable "set directory exists or cannot be created"; }
  chmod 700 "$SET_DIR"
  trap 'fence_abort interrupted' TERM INT HUP
  t_start="$(now_utc)"
  if [[ -n "${OPT[use-fence]:-}" ]]; then
    GUARD=capture-external
    FENCE_FILE="$(open_fence "$root")"
    [[ -n "$FENCE_FILE" ]] || { GUARD=""; printf 'status\tincomplete\nreason\tfence_not_established\n' >"$SET_DIR/INCOMPLETE"; die 4 fence_not_established "--use-fence without an established fence"; }
    FENCE_ID="$(fence_get "$FENCE_FILE" fence_id)"; fence_mode=external
    [[ -n "$(fence_get "$FENCE_FILE" admin_window)" ]] || { FENCE_FILE=""; capture_fail admin_window_unavailable; }
    window_checks "$FENCE_FILE" || { FENCE_FILE=""; capture_fail window_changed; }
  else
    fence_mode=integrated
    fence_establish "$envf" "$root" "${OPT[admin-window]}" "${RECOVERY_DEADLINE:-1800}"
    GUARD=capture-integrated
    trap 'fence_abort interrupted' TERM INT HUP
  fi
  # Drill seam (inert unless set; never used in operation): an unexpected failure during the capture whose INCOMPLETE
  # marker cannot be written (its path is a directory), so the validation drill shows the exit guard still releases the
  # fence and the command stays non-zero (Codex review C2).
  if [[ "${RECOVERY_DRILL_FAIL_CAPTURE:-}" == marker ]]; then mkdir -p "$SET_DIR/INCOMPLETE"; false; fi
  # C1 ledger sync and mark.
  ledger_sync "$ledger" "$envf" "$root" >/dev/null || capture_fail ledger_invalid
  local mark; mark="$(ledger_check "$ledger")"
  # C2 fingerprint P1, C3 listing L1.
  local w_db0; w_db0="$(sql1 "$SRC_PG" "$SRC_DB" "SELECT to_char(clock_timestamp() AT TIME ZONE 'UTC', 'YYYY-MM-DD\"T\"HH24:MI:SS.US\"Z\"')")"
  fingerprint SRC "$SET_DIR/p1.tsv" || capture_fail source_unavailable
  objects_list "$SRC_NET" "$SRC_MCIMG" "$envf" "$SRC_BUCKET" >"$SET_DIR/l1.tsv" || capture_fail source_unavailable
  # C4 database capture: exactly DATABASE_NAME, custom format, streamed on stdout.
  timeout "$STEP_TIMEOUT" docker exec "$SRC_PG" pg_dump --format=custom -U postgres -d "$SRC_DB" >"$SET_DIR/database.dump" 2>"$SET_DIR/pg_dump.err" ||
    capture_fail capture_failed
  [[ -s "$SET_DIR/database.dump" && ! -s "$SET_DIR/pg_dump.err" ]] || capture_fail capture_failed
  timeout "$STEP_TIMEOUT" docker run --rm -i "${LABELS[@]}" --network none "$SRC_PGIMG" sh -c 'cat >/tmp/a && pg_restore -l /tmp/a' \
    <"$SET_DIR/database.dump" >"$SET_DIR/toc.txt" 2>/dev/null || capture_fail capture_failed
  toc_check "$SET_DIR/toc.txt" "$SET_DIR/p1.tsv" || capture_fail capture_failed
  # C5 object capture as the fixture administrator (the service user is fenced): bytes tar and binding metadata.
  if [[ -s "$SET_DIR/l1.tsv" ]]; then
    mcx "$SRC_NET" "$SRC_MCIMG" "$envf" admin "mkdir -p /tmp/o && mc cp -q --recursive 'x/$SRC_BUCKET/' /tmp/o/ >/dev/null && tar -C /tmp/o -cf - ." \
      >"$SET_DIR/objects.tar" || capture_fail capture_failed
    mcx "$SRC_NET" "$SRC_MCIMG" "$envf" admin "mc stat --json --recursive 'x/$SRC_BUCKET/'" >"$SET_DIR/stat.jsonl" || capture_fail capture_failed
  else
    tar -cf "$SET_DIR/objects.tar" -T /dev/null; : >"$SET_DIR/stat.jsonl"
  fi
  meta_from_stat "$SET_DIR/stat.jsonl" >"$SET_DIR/objects.meta"; rm -f -- "$SET_DIR/stat.jsonl"
  # C6 window checks, then the supplementary re-check.
  window_checks "$FENCE_FILE" || capture_fail window_changed
  fingerprint SRC "$SET_DIR/p2.tsv" || capture_fail source_unavailable
  objects_list "$SRC_NET" "$SRC_MCIMG" "$envf" "$SRC_BUCKET" >"$SET_DIR/l2.tsv" || capture_fail source_unavailable
  cmp -s "$SET_DIR/p1.tsv" "$SET_DIR/p2.tsv" && cmp -s "$SET_DIR/l1.tsv" "$SET_DIR/l2.tsv" || capture_fail window_changed
  local w_db1; w_db1="$(sql1 "$SRC_PG" "$SRC_DB" "SELECT to_char(clock_timestamp() AT TIME ZONE 'UTC', 'YYYY-MM-DD\"T\"HH24:MI:SS.US\"Z\"')")"
  rm -f -- "$SET_DIR/p2.tsv" "$SET_DIR/l2.tsv"
  # C7 integrity: per-object size and SHA-256 from the tar against L1 and the metadata.
  local name size sum ct oid org grp objrows=""
  while IFS=$'\t' read -r name size _; do
    [[ -n "$name" ]] || continue
    sum="$(tar -xOf "$SET_DIR/objects.tar" "./$name" | sha256sum | cut -d' ' -f1)" || capture_fail integrity_failed
    [[ "$(tar -xOf "$SET_DIR/objects.tar" "./$name" | wc -c)" == "$size" ]] || capture_fail integrity_failed
    IFS=$'\t' read -r ct oid org grp < <(awk -F'\t' -v n="$name" '$1 == n { print $2 "\t" $3 "\t" $4 "\t" $5 }' "$SET_DIR/objects.meta")
    [[ -n "${ct:-}" && "${oid:-}" =~ $UUID_RE ]] || capture_fail integrity_failed
    objrows+="object"$'\t'"$name"$'\t'"$size"$'\t'"$sum"$'\t'"$ct"$'\t'"$oid"$'\t'"$org"$'\t'"${grp:--}"$'\n'
  done <"$SET_DIR/l1.tsv"
  # Manifest, then COMPLETE (capture completeness only; release is separate).
  {
    printf 'format\t%s\n' "$MANIFEST_FORMAT"; printf 'backup_id\t%s\n' "$id"; printf 'created\t%s\n' "$t_start"
    printf 'source_project\t%s\n' "$SRC_PROJECT"; printf 'source_pg_container\t%s\n' "$SRC_PG"; printf 'source_pg_volume\t%s\n' "$SRC_PGVOL"
    printf 'source_minio_container\t%s\n' "$SRC_MINIO"; printf 'source_minio_volume\t%s\n' "$SRC_MINIOVOL"
    printf 'source_sysid\t%s\n' "$(pg_sysid "$SRC_PG" "$SRC_DB")"; printf 'database\t%s\n' "$SRC_DB"
    printf 'server_major\t%s\n' "$(pg_major "$SRC_PG" "$SRC_DB")"; printf 'server_version\t%s\n' "$(sql1 "$SRC_PG" "$SRC_DB" 'SHOW server_version')"
    printf 'bucket\t%s\n' "$SRC_BUCKET"; printf 'schema_version\t%s\n' "$(schema_version "$SRC_PG" "$SRC_DB")"
    printf 'pg_image\t%s\n' "$SRC_PGIMG"; printf 'mc_image\t%s\n' "$(docker image inspect -f '{{.Id}}' "$SRC_MCIMG")"
    printf 'pg_dump_version\t%s\n' "$(docker exec "$SRC_PG" pg_dump --version)"
    printf 'window_db\t%s\t%s\n' "$w_db0" "$w_db1"; printf 'window_host\t%s\t%s\n' "$t_start" "$(now_utc)"
    printf 'fence_id\t%s\t%s\n' "$FENCE_ID" "$fence_mode"
    printf 'admin_window\t%s\t%s\t%s\n' "$(fence_get "$FENCE_FILE" admin_window)" "$(fence_get "$FENCE_FILE" admin_operator)" "$(fence_get "$FENCE_FILE" admin_window_start)"
    printf 'holder\t%s\n' "$(fence_get "$FENCE_FILE" locks_held)"; printf 'provider_restart\t%s\n' "$(fence_get "$FENCE_FILE" provider_restarted)"
    printf 'trace\tnot_used\n'; printf 'supplementary\tp2_equals_p1 l2_equals_l1 no_other_superuser\n'
    printf 'ledger\t%s\t%s\t%s\n' "$(cut -f1 <<<"$mark")" "$(cut -f3 <<<"$mark")" "$(cut -f4 <<<"$mark")"
    printf 'component\tdatabase.dump\t%s\t%s\n' "$(stat -c %s "$SET_DIR/database.dump")" "$(sha_file "$SET_DIR/database.dump")"
    printf 'component\tobjects.tar\t%s\t%s\n' "$(stat -c %s "$SET_DIR/objects.tar")" "$(sha_file "$SET_DIR/objects.tar")"
    printf 'component\tobjects.meta\t%s\t%s\n' "$(stat -c %s "$SET_DIR/objects.meta")" "$(sha_file "$SET_DIR/objects.meta")"
    cat "$SET_DIR/p1.tsv"
    printf '%s' "$objrows"
  } >"$SET_DIR/manifest.tsv"
  rm -f -- "$SET_DIR/p1.tsv" "$SET_DIR/l1.tsv" "$SET_DIR/pg_dump.err"
  sync -- "$SET_DIR"/*
  sha_file "$SET_DIR/manifest.tsv" >"$SET_DIR/COMPLETE"; sync -- "$SET_DIR/COMPLETE"
  # Release (integrated fence): recorded and reported separately; a failed release keeps the set complete. The signal
  # trap ends with the capture in both modes: after COMPLETE nothing is aborted, and an external fence stays held.
  local rc=0
  trap - TERM INT HUP
  GUARD=""
  if [[ "$fence_mode" == integrated ]]; then
    if fence_release "$FENCE_FILE" "$envf"; then release=verified; else release=failed; rc=4; fi
  fi
  printf 'backup_id\t%s\nstatus\tcomplete\nrelease\t%s\nfence_id\t%s\nfinished\t%s\n' "$id" "$release" "$FENCE_ID" "$(now_utc)" >"$SET_DIR/backup-receipt.tsv"
  printf 'backup\t%s\tcomplete\trelease=%s\tobjects=%s\tdump_bytes=%s\n' "$id" "$release" "$(grep -c '^object' "$SET_DIR/manifest.tsv" || true)" "$(stat -c %s "$SET_DIR/database.dump")"
  [[ $rc == 0 ]] || die 4 unfence_failed "set complete; the fence release failed: rerun unfence"
}

capture_fail() { # capture_fail REASON: incomplete set, release an integrated fence, exit non-zero
  local reason="$1"
  trap - TERM INT HUP
  GUARD=""
  # Release first: a failed marker write must not skip it (Codex review C2).
  if [[ "${fence_mode:-}" == integrated && -n "$FENCE_FILE" ]]; then fence_release "$FENCE_FILE" "$FENCE_ENV" >/dev/null 2>&1 || log "fence release failed: rerun unfence"; fi
  if [[ -n "${SET_DIR:-}" && -d "$SET_DIR" && ! -e "$SET_DIR/COMPLETE" ]]; then
    { printf 'status\tincomplete\nreason\t%s\n' "$reason" >"$SET_DIR/INCOMPLETE"; } 2>/dev/null || log "the INCOMPLETE marker could not be written"
  fi
  printf 'backup\t%s\tincomplete\treason=%s\n' "${id:-?}" "$reason"
  die 4 "$reason" "capture stopped; no COMPLETE written"
}

toc_check() { # toc_check TOC P1: only schemas app and audit; the TABLE entries are exactly P1's tables
  local toc="$1" p1="$2" ns
  ns="$(grep '^[0-9]' "$toc" | awk '{print $5}' | LC_ALL=C sort -u | tr '\n' ' ')"
  [[ "$ns" == "- ACL CONSTRAINT DATA SECURITY SET app audit " || "$ns" == "- ACL CONSTRAINT DATA SECURITY app audit " ]] || { log "TOC namespaces: $ns"; return 1; }
  [[ "$(grep -cE '^[0-9]+; [0-9]+ [0-9]+ SCHEMA - (app|audit) ' "$toc")" == 2 ]] || return 1
  diff <(grep -E '^[0-9]+; [0-9]+ [0-9]+ TABLE (app|audit) ' "$toc" | awk '{print $5 "." $6}' | LC_ALL=C sort) \
       <(awk -F'\t' '$1 == "table" { print $2 }' "$p1" | LC_ALL=C sort) >/dev/null
}

meta_from_stat() { # meta_from_stat STAT.JSONL: "name<TAB>content-type<TAB>object-id<TAB>organization-id<TAB>group-id"
  local line
  while IFS= read -r line; do
    [[ "$line" == *'"type":"file"'* ]] || continue
    printf '%s\t%s\t%s\t%s\t%s\n' \
      "$(sed -n 's/.*"name":"\([^"]*\)".*/\1/p' <<<"$line")" \
      "$(sed -n 's/.*"Content-Type":"\([^"]*\)".*/\1/p' <<<"$line")" \
      "$(sed -n 's/.*"X-Amz-Meta-Object-Id":"\([^"]*\)".*/\1/p' <<<"$line")" \
      "$(sed -n 's/.*"X-Amz-Meta-Organization-Id":"\([^"]*\)".*/\1/p' <<<"$line")" \
      "$(sed -n 's/.*"X-Amz-Meta-Group-Id":"\([^"]*\)".*/\1/p' <<<"$line")"
  done <"$1" | LC_ALL=C sort
}

# ======================================================================================================================
# Verify (integrity of a set)
# ======================================================================================================================

verify_set() { # verify_set SET: prints the failed component (or nothing); 0 when intact
  local s="$1" name size sum
  [[ -f "$s/COMPLETE" && -f "$s/manifest.tsv" ]] || { printf 'incomplete'; return 1; }
  [[ "$(cat "$s/COMPLETE")" == "$(sha_file "$s/manifest.tsv")" ]] || { printf 'manifest'; return 1; }
  [[ "$(fence_get "$s/manifest.tsv" format)" == "$MANIFEST_FORMAT" ]] || { printf 'format'; return 1; }
  while IFS=$'\t' read -r _ name size sum; do
    [[ -f "$s/$name" && "$(stat -c %s "$s/$name")" == "$size" && "$(sha_file "$s/$name")" == "$sum" ]] || { printf '%s' "$name"; return 1; }
  done < <(awk -F'\t' '$1 == "component"' "$s/manifest.tsv")
  diff <(tar -tf "$s/objects.tar" 2>/dev/null | grep -v '/$' | sed 's|^\./||' | LC_ALL=C sort) \
       <(awk -F'\t' '$1 == "object" { print $2 }' "$s/manifest.tsv" | LC_ALL=C sort) >/dev/null || { printf 'objects.tar'; return 1; }
  while IFS=$'\t' read -r _ name size sum _; do
    [[ "$(tar -xOf "$s/objects.tar" "./$name" | sha256sum | cut -d' ' -f1)" == "$sum" ]] || { printf 'objects.tar'; return 1; }
  done < <(awk -F'\t' '$1 == "object"' "$s/manifest.tsv")
}

cmd_verify() {
  need backup
  local bad
  if bad="$(verify_set "${OPT[backup]}")"; then printf 'verify\t%s\tok\n' "$(basename "${OPT[backup]}")"
  else printf 'verify\t%s\tfailed\tcomponent=%s\n' "$(basename "${OPT[backup]}")" "$bad"; die 4 integrity_failed "component $bad"; fi
}

# ======================================================================================================================
# Restore (Plan r3 §4)
# ======================================================================================================================

backup_policy() { read_policy "$1" "policy_id backup_max_age_s backup_retain_for_s backup_holds pending_review_after_s" \
  "policy_id backup_max_age_s backup_retain_for_s backup_holds pending_review_after_s"
  int_in "${POL[backup_max_age_s]}" 1 31536000 && int_in "${POL[backup_retain_for_s]}" 1 315360000 && int_in "${POL[pending_review_after_s]}" 1 31536000 ||
    die 2 policy_invalid "backup policy bounds"
  check_holds "${POL[backup_holds]}" || die 2 policy_invalid "backup_holds"
}

set_age_s() { echo $(( $(date -u +%s) - $(date -u -d "$(fence_get "$1/manifest.tsv" created)" +%s) )); }
set_expirable() { # set_expirable SET: older than the retention, not held, not quarantined (needs POL)
  local s="$1" id; id="$(basename "$s")"
  [[ ! -e "$s/QUARANTINED" ]] || return 1
  [[ ",${POL[backup_holds]}," != *",$id,"* ]] || return 1
  (( $(set_age_s "$s") > 10#${POL[backup_retain_for_s]} ))
}

rcpt_get() { [[ -f "$1" ]] || return 0; awk -F'\t' -v k="$2" '$1 == k { print $2; exit }' "$1"; }

cmd_restore() {
  need backup target-env-file receipts backup-policy
  local s="${OPT[backup]}" tenv="${OPT[target-env-file]}" rdir="${OPT[receipts]}" steps="${OPT[steps]:-all}" m bad
  with_deadline "${OPT[timeout]:-1800}"
  [[ "$steps" == all || "$steps" == database || "$steps" == objects ]] || die 2 invalid_input "--steps all|database|objects"
  backup_policy "${OPT[backup-policy]}"
  private_dir "$rdir"
  recovery_labels restore
  # Set: complete, intact, not expired, not quarantined, known format.
  [[ -e "$s/EXPIRED" || -e "$s/EXPIRED-PARTIAL" ]] && die 3 set_expired "the set is expired"
  [[ -e "$s/QUARANTINED" ]] && die 3 set_quarantined "the set is quarantined"
  bad="$(verify_set "$s")" || die 3 set_invalid "component $bad"
  m="$s/manifest.tsv"
  set_expirable "$s" && die 3 set_expired "the set is past backup_retain_for_s and not held"
  # A first restore needs a receipts directory without later-stage receipts: stale invalidation, replay or reopen
  # receipts of an earlier target must never satisfy the gate of this one (S-review F1).
  if [[ "$steps" != objects ]]; then
    for v in invalidate.tsv replay.tsv reopen.tsv .replay.prev; do
      [[ ! -e "$rdir/$v" ]] || die 3 receipts_reused "the receipts directory holds receipts of an earlier restore; use a fresh directory"
    done
  fi
  # Target identity and compatibility.
  check_env_file "$tenv" POSTGRES_SUPERUSER_PASSWORD OBJECT_STORE_ACCESS_KEY OBJECT_STORE_SECRET_KEY
  resolve TGT "$tenv"
  [[ "$TGT_DB" == "$(fence_get "$m" database)" ]] || die 3 incompatible "target DATABASE_NAME differs"
  [[ "$(pg_major "$TGT_PG" "$TGT_DB")" == "$(fence_get "$m" server_major)" ]] || die 3 incompatible "server major version differs"
  local maxmig; maxmig="$(find "$RECOVERY_REPO/core/migrations" -maxdepth 1 -name '[0-9][0-9][0-9][0-9][0-9]_*.sql' | sed 's|.*/||' | sort | tail -1 | cut -c1-5)"
  (( 10#$(fence_get "$m" schema_version) <= 10#${maxmig:-0} )) || die 3 incompatible "set schema version above this checkout"
  local tsys; tsys="$(pg_sysid "$TGT_PG" "$TGT_DB")"
  [[ "$TGT_PROJECT" != "$(fence_get "$m" source_project)" ]] || die 3 target_alias "target project equals the source"
  for v in "$TGT_PG" "$TGT_PGVOL" "$TGT_MINIO" "$TGT_MINIOVOL"; do
    for k in source_pg_container source_pg_volume source_minio_container source_minio_volume; do
      [[ "$v" != "$(fence_get "$m" "$k")" ]] || die 3 target_alias "target container or volume equals the source"
    done
  done
  [[ "$tsys" != "$(fence_get "$m" source_sysid)" ]] || die 3 target_alias "target system identifier equals the source"
  if [[ -n "${OPT[source-env-file]:-}" ]]; then
    resolve SRC "${OPT[source-env-file]}"
    for v in "$TGT_PG" "$TGT_PGVOL" "$TGT_MINIO" "$TGT_MINIOVOL"; do
      [[ "$v" != "$SRC_PG" && "$v" != "$SRC_PGVOL" && "$v" != "$SRC_MINIO" && "$v" != "$SRC_MINIOVOL" ]] || die 3 target_alias "target equals the live source"
    done
    [[ "$tsys" != "$(pg_sysid "$SRC_PG" "$SRC_DB")" ]] || die 3 target_alias "target system identifier equals the live source"
  fi
  for svc in core web intelligence; do [[ -z "$(cid_of "$TGT_PROJECT" "$svc")" ]] || die 3 target_open "target $svc is running"; done
  local rc="$rdir/restore.tsv" nrel nobj
  nrel="$(sql1 "$TGT_PG" "$TGT_DB" "SELECT count(*) FROM pg_class WHERE relnamespace IN ('app'::regnamespace, 'audit'::regnamespace)")"
  objects_list "$TGT_NET" "$TGT_MCIMG" "$tenv" "$TGT_BUCKET" >"$rdir/.target-listing" 2>/dev/null || { rm -f "$rdir/.target-listing"; die 3 target_unavailable "target MinIO unreachable"; }
  nobj="$(grep -c . "$rdir/.target-listing" || true)"
  if [[ "$steps" == objects ]]; then
    # Continuation: only on the exact receipt-bound target and checkpoint, database step done and still equal.
    [[ -f "$rc" && "$(rcpt_get "$rc" backup_id)" == "$(fence_get "$m" backup_id)" && "$(rcpt_get "$rc" manifest_sha256)" == "$(cat "$s/COMPLETE")" &&
       "$(rcpt_get "$rc" target_project)" == "$TGT_PROJECT" && "$(rcpt_get "$rc" target_pg)" == "$TGT_PG" && "$(rcpt_get "$rc" target_pg_volume)" == "$TGT_PGVOL" &&
       "$(rcpt_get "$rc" target_minio)" == "$TGT_MINIO" && "$(rcpt_get "$rc" target_minio_volume)" == "$TGT_MINIOVOL" &&
       "$(rcpt_get "$rc" target_sysid)" == "$tsys" && "$(rcpt_get "$rc" database)" == "done" ]] ||
      { rm -f "$rdir/.target-listing"; die 3 continuation_mismatch "no receipt binding this backup, target and a verified database step"; }
    fingerprint TGT "$rdir/.target-fp" || die 3 target_unavailable "target unreadable"
    diff <(awk -F'\t' '$1 == "table" || $1 == "catalog" || $1 == "roles"' "$m") "$rdir/.target-fp" >/dev/null ||
      { rm -f "$rdir/.target-fp" "$rdir/.target-listing"; die 3 continuation_mismatch "target data or catalog differs from the set"; }
    rm -f "$rdir/.target-fp"
    # Every object already on the target must be the set's object: same name and size, and the same bytes and binding
    # (a mismatch refuses; the continuation never overwrites it: S-review F9). fd 3: docker clients read stdin.
    local name size row
    while IFS=$'\t' read -r -u 3 name size _; do
      [[ -n "$name" ]] || continue
      row="$(awk -F'\t' -v n="$name" -v z="$size" '$1 == "object" && $2 == n && $3 == z { print $4 "\t" $6 "\t" $7 "\t" $8; exit }' "$m")"
      [[ -n "$row" ]] || { rm -f "$rdir/.target-listing"; die 3 continuation_mismatch "target bucket holds an object outside the set"; }
      object_matches "$tenv" "$name" "$(cut -f1 <<<"$row")" "$(cut -f2 <<<"$row")" "$(cut -f3 <<<"$row")" "$(cut -f4 <<<"$row")" ||
        { rm -f "$rdir/.target-listing"; die 3 continuation_mismatch "a target object differs from the set's bytes or binding"; }
    done 3<"$rdir/.target-listing"
  else
    [[ "$nrel" == 0 && "$nobj" == 0 ]] || { rm -f "$rdir/.target-listing"; die 3 target_not_fresh "target is not fresh"; }
    {
      printf 'format\tg013-restore/1\nbackup_id\t%s\nmanifest_sha256\t%s\n' "$(fence_get "$m" backup_id)" "$(cat "$s/COMPLETE")"
      printf 'source_sysid\t%s\nsource_project\t%s\n' "$(fence_get "$m" source_sysid)" "$(fence_get "$m" source_project)"
      printf 'target_project\t%s\ntarget_pg\t%s\ntarget_pg_volume\t%s\ntarget_minio\t%s\ntarget_minio_volume\t%s\ntarget_sysid\t%s\n' \
        "$TGT_PROJECT" "$TGT_PG" "$TGT_PGVOL" "$TGT_MINIO" "$TGT_MINIOVOL" "$tsys"
      printf 'started\t%s\n' "$(now_utc)"
    } >"$rc"
  fi
  rm -f "$rdir/.target-listing"
  # Roles from provisioning (fresh values): attributes and memberships equal the set's role section.
  local t0 tfp; tfp="$rdir/.target-fp"
  if [[ "$steps" != objects ]]; then
    t0=$SECONDS
    fingerprint TGT "$tfp" || { rcpt_put "$rc" database failed; die 4 restore_failed "target unreadable"; }
    [[ "$(awk -F'\t' '$1 == "roles"' "$tfp")" == "$(awk -F'\t' '$1 == "roles"' "$m")" ]] ||
      { rcpt_put "$rc" database failed; rcpt_put "$rc" status failed; rm -f "$tfp"; die 4 restore_failed "target roles differ from the set (provisioning)"; }
    # Database: TOC minus the two provisioned schema entries, one transaction, stop on the first error. The provisioned
    # default sequence grant of the owner is withdrawn first in the same transaction, so restored sequences (the
    # migration ledger's identity sequence) do not gain it; the set's own DEFAULT ACL entries then restore the source's
    # exact default privileges (b3: without this the restored ledger sequence gained runtime SELECT and USAGE).
    if ! docker run --rm -i "${LABELS[@]}" --network "$TGT_NET" --env-file <(printf 'PGPASSWORD=%s\n' "$(envv "$tenv" POSTGRES_SUPERUSER_PASSWORD)") \
         -e RESTORE_PRE="ALTER DEFAULT PRIVILEGES FOR ROLE \"${TGT_DB}_owner\" IN SCHEMA app REVOKE ALL ON SEQUENCES FROM \"${TGT_DB}_runtime\";" \
         "$TGT_PGIMG" sh -ec 'cat >/tmp/a; pg_restore -l /tmp/a | grep -vE "^[0-9]+; [0-9]+ [0-9]+ SCHEMA - (app|audit) " >/tmp/l
           printf "%s\n" "$RESTORE_PRE" >/tmp/s.sql; pg_restore -L /tmp/l -f /tmp/r.sql /tmp/a; cat /tmp/r.sql >>/tmp/s.sql
           psql -X -q -v ON_ERROR_STOP=1 --single-transaction --no-password -h postgres -U postgres -d "$0" -f /tmp/s.sql >/dev/null' "$TGT_DB" \
         <"$s/database.dump" >"$rdir/.pg_restore.log" 2>&1; then
      rcpt_put "$rc" database failed; rcpt_put "$rc" status failed; rm -f "$tfp" "$rdir/.pg_restore.log"; die 4 restore_failed "pg_restore failed (rolled back)"
    fi
    rm -f "$rdir/.pg_restore.log"
    fingerprint TGT "$tfp" || { rcpt_put "$rc" database failed; die 4 restore_failed "target unreadable"; }
    if diff <(awk -F'\t' '$1 == "table" || $1 == "catalog" || $1 == "roles"' "$m") "$tfp" >/dev/null; then
      rcpt_put "$rc" database "done"; rcpt_put "$rc" database_seconds $((SECONDS - t0))
    else
      rcpt_put "$rc" database failed; rcpt_put "$rc" status failed; rm -f "$tfp"; die 4 restore_failed "restored data or catalog differs from the set"
    fi
    rm -f "$tfp"
  fi
  if [[ "$steps" != database ]]; then
    t0=$SECONDS
    local name size sum ct oid org grp fail=0 have attrs
    [[ "$TGT_MINIO_UP" == 1 ]] || fail=1
    [[ $fail == 1 ]] || objects_list "$TGT_NET" "$TGT_MCIMG" "$tenv" "$TGT_BUCKET" >"$rdir/.target-listing" 2>/dev/null || fail=1
    # fd 3: the docker clients in the body inherit stdin and must not consume the manifest lines.
    while IFS=$'\t' read -r -u 3 _ name size sum ct oid org grp; do
      [[ $fail == 0 ]] || break
      have="$(awk -F'\t' -v n="$name" '$1 == n { print $2 }' "$rdir/.target-listing" 2>/dev/null)"
      if [[ "$have" == "$size" ]] && object_matches "$tenv" "$name" "$sum" "$oid" "$org" "$grp"; then continue; fi
      attrs="Content-Type=$ct;Content-Disposition=attachment;X-Amz-Meta-Object-Id=$oid;X-Amz-Meta-Organization-Id=$org"
      [[ "$grp" != - ]] && attrs+=";X-Amz-Meta-Group-Id=$grp"
      tar -xOf "$s/objects.tar" "./$name" | mcx "$TGT_NET" "$TGT_MCIMG" "$tenv" svc "cat >/tmp/f && mc cp -q --attr '$attrs' /tmp/f 'x/$TGT_BUCKET/$name' >/dev/null" ||
        { fail=1; break; }
      object_matches "$tenv" "$name" "$sum" "$oid" "$org" "$grp" || { fail=1; break; }
    done 3< <(awk -F'\t' '$1 == "object"' "$m")
    rm -f "$rdir/.target-listing"
    if [[ $fail == 0 ]] && diff <(objects_list "$TGT_NET" "$TGT_MCIMG" "$tenv" "$TGT_BUCKET" | cut -f1,2) \
         <(awk -F'\t' '$1 == "object" { print $2 "\t" $3 }' "$m" | LC_ALL=C sort) >/dev/null; then
      rcpt_put "$rc" objects "done"; rcpt_put "$rc" objects_seconds $((SECONDS - t0))
    else
      rcpt_put "$rc" objects failed; rcpt_put "$rc" status partial
      printf 'restore\t%s\tpartial\tdatabase=%s\tobjects=failed\n' "$(fence_get "$m" backup_id)" "$(rcpt_get "$rc" database)"
      die 4 restore_partial "object step failed; the target stays closed (continue with --steps objects)"
    fi
  fi
  # Redis: fresh and empty (or not running).
  local rcid redis=not_running
  rcid="$(cid_of "$TGT_PROJECT" redis)"
  if [[ -n "$rcid" ]]; then
    redis="$(docker exec -i --env-file <(printf 'REDISCLI_AUTH=%s\n' "$(envv "$tenv" REDIS_PASSWORD)") "$rcid" redis-cli DBSIZE 2>/dev/null | tr -d '\r')"
    [[ "$redis" == 0 ]] && redis=empty
  fi
  rcpt_put "$rc" redis "$redis"
  if [[ "$(rcpt_get "$rc" database)" == "done" && "$(rcpt_get "$rc" objects)" == "done" && ( "$redis" == empty || "$redis" == not_running ) ]]; then
    rcpt_put "$rc" status restored-closed; rcpt_put "$rc" finished "$(now_utc)"
    printf 'restore\t%s\trestored-closed\n' "$(fence_get "$m" backup_id)"
  else
    rcpt_put "$rc" status partial
    printf 'restore\t%s\tpartial\tdatabase=%s\tobjects=%s\tredis=%s\n' "$(fence_get "$m" backup_id)" "$(rcpt_get "$rc" database)" "$(rcpt_get "$rc" objects)" "$redis"
    [[ "$steps" == database ]] && return 0
    die 4 restore_partial "restore incomplete; the target stays closed"
  fi
}

rcpt_put() { # rcpt_put FILE KEY VALUE: replace (or add) one receipt key
  local f="$1" tmp; tmp="$(mktemp "$f.XXXXXX")"
  awk -F'\t' -v k="$2" '$1 != k' "$f" >"$tmp"; printf '%s\t%s\n' "$2" "$3" >>"$tmp"; mv -- "$tmp" "$f"; sync -- "$f"
}

object_matches() { # object_matches ENVFILE NAME SHA OBJ ORG GRP: target bytes and binding equal the set's
  local stat
  stat="$(mcx "$TGT_NET" "$TGT_MCIMG" "$1" svc "mc stat --json 'x/$TGT_BUCKET/$2'" 2>/dev/null)" || return 1
  [[ "$stat" == *"\"X-Amz-Meta-Object-Id\":\"$4\""* && "$stat" == *"\"X-Amz-Meta-Organization-Id\":\"$5\""* ]] || return 1
  if [[ "$6" != - ]]; then [[ "$stat" == *"\"X-Amz-Meta-Group-Id\":\"$6\""* ]] || return 1; fi
  [[ "$(mcx "$TGT_NET" "$TGT_MCIMG" "$1" svc "mc cat 'x/$TGT_BUCKET/$2'" | sha256sum | cut -d' ' -f1)" == "$3" ]]
}

# ======================================================================================================================
# Restored authentication (Plan r3 §5.1)
# ======================================================================================================================

# restore_binding RECEIPTS: the digest of the restore receipt's identity lines, fixed when the first restore step wrote
# it (set, manifest digest, target project, PostgreSQL container and volume, system identifier, start time). Invalidation
# and replay receipts carry it, and replay and reopen accept only receipts bound to the current restore (S-review F1;
# the whole receipt's digest would change when a staged object step completes after invalidation).
restore_binding() {
  local rc="$1/restore.tsv" k v=""
  [[ -f "$rc" ]] || return 1
  for k in backup_id manifest_sha256 target_project target_pg target_pg_volume target_sysid started; do v+="$k=$(rcpt_get "$rc" "$k");"; done
  sha_text "$v"
}
bound_to_restore() { # bound_to_restore RECEIPTFILE RECEIPTS: the receipt exists and carries the current restore binding
  [[ -f "$1" && -n "$(rcpt_get "$1" restore_binding)" && "$(rcpt_get "$1" restore_binding)" == "$(restore_binding "$2")" ]]
}

target_bound() { # target_bound RECEIPTS: the live target is the restore receipt's target and not its source
  local rc="$1/restore.tsv"
  [[ -f "$rc" ]] || die 3 receipt_missing "no restore receipt"
  [[ "$(rcpt_get "$rc" target_project)" == "$TGT_PROJECT" && "$(rcpt_get "$rc" target_pg)" == "$TGT_PG" &&
     "$(rcpt_get "$rc" target_sysid)" == "$(pg_sysid "$TGT_PG" "$TGT_DB")" ]] || die 3 target_mismatch "target is not the receipt's target"
  [[ "$(rcpt_get "$rc" source_sysid)" != "$(pg_sysid "$TGT_PG" "$TGT_DB")" ]] || die 3 target_alias "target is the source"
}

cmd_invalidate() {
  need target-env-file receipts
  resolve TGT "${OPT[target-env-file]}"; recovery_labels invalidate
  target_bound "${OPT[receipts]}"
  [[ "$(rcpt_get "${OPT[receipts]}/restore.tsv" database)" == "done" ]] || die 3 receipt_missing "database step not done"
  for svc in core web intelligence; do [[ -z "$(cid_of "$TGT_PROJECT" "$svc")" ]] || die 3 target_open "target $svc is running"; done
  local out
  out="$(sqlf "$TGT_PG" "$TGT_DB" invalidate.sql)" || die 4 invalidation_failed "invalidation rolled back"
  { printf 'format\tg013-invalidate/1\nstatus\tok\nfinished\t%s\n' "$(now_utc)"
    printf 'restore_binding\t%s\ntarget_project\t%s\ntarget_pg\t%s\ntarget_sysid\t%s\n' "$(restore_binding "${OPT[receipts]}")" \
      "$TGT_PROJECT" "$TGT_PG" "$(pg_sysid "$TGT_PG" "$TGT_DB")"
    printf '%s\n' "$out"; } >"${OPT[receipts]}/invalidate.tsv"
  printf 'invalidate\tok\t%s\n' "$(tr '\n\t' ' =' <<<"$out")"
}

# ======================================================================================================================
# Restore-input ledger (Plan r3 §5.2, §5.3)
# ======================================================================================================================

latest_mark() { # latest_mark ROOT LEDGERUUID: the mark sequence of the latest complete, unquarantined set of this ledger
  local s best="" bestt="" t
  for s in "$1"/*/; do
    s="${s%/}"
    [[ -f "$s/COMPLETE" && ! -e "$s/QUARANTINED" && ! -e "$s/EXPIRED" ]] || continue
    [[ "$(fence_get "$s/manifest.tsv" ledger | cut -f1)" == "$2" ]] || continue
    t="$(fence_get "$s/manifest.tsv" created)"
    [[ -z "$bestt" || "$t" > "$bestt" ]] && { bestt="$t"; best="$(fence_get "$s/manifest.tsv" ledger | cut -f2)"; }
  done
  printf '%s' "${best:-0}"
}

ledger_sync() { # ledger_sync LEDGER ENVFILE ROOT: append entries for currently restrictive states (Plan r3 §5.2 rule)
  local f="$1" envf="$2" root="$3" head luuid mark kind refs obj n=0
  head="$(ledger_check "$f")" || return 1
  luuid="$(cut -f1 <<<"$head")"
  [[ "$(cut -f2 <<<"$head")" == "$SRC_DB" ]] || return 1
  mark="$(latest_mark "$root" "$luuid")"
  while IFS=$'\t' read -r kind refs; do
    [[ -n "$kind" ]] || continue
    if [[ "$kind" == object.delete ]]; then
      obj="$(ref_get "$refs" object)"
      awk -F'\t' -v o="object=$obj" 'NR > 1 && $4 == "object.delete" && index("," $5 ",", "," o ",") { f = 1 } END { exit !f }' "$f" && continue
    else
      awk -F'\t' -v k="$kind" -v r="$refs" -v m="$mark" 'NR > 1 && $1 > m && $4 == k && $5 == r { f = 1 } END { exit !f }' "$f" && continue
    fi
    ledger_append "$f" "$kind" "$refs" sync ledger-sync
    n=$((n + 1))
  done < <(sqlf "$SRC_PG" "$SRC_DB" ledger-sync.sql)
  printf '%s' "$n"
}

ledger_verify_mark() { # ledger_verify_mark LEDGER SET: chain valid, same UUID, mark present with the manifest digest
  local f="$1" m="$2/manifest.tsv" head mrow seq dig line
  head="$(ledger_check "$f")" || { printf 'chain'; return 1; }
  mrow="$(fence_get "$m" ledger)"
  [[ "$(cut -f1 <<<"$head")" == "$(cut -f1 <<<"$mrow")" ]] || { printf 'uuid'; return 1; }
  seq="$(cut -f2 <<<"$mrow")"; dig="$(cut -f3 <<<"$mrow")"
  (( 10#$(cut -f3 <<<"$head") >= 10#$seq )) || { printf 'stale'; return 1; }
  if (( seq == 0 )); then line="$(head -1 "$f")"; else line="$(sed -n "$((seq + 1))p" "$f")"; fi
  [[ "$(sha_text "$line")" == "$dig" ]] || { printf 'mark'; return 1; }
}

cmd_ledger() {
  local sub="${ARGS[0]:-}"
  case "$sub" in
    init)
      need ledger env-file
      [[ ! -e "${OPT[ledger]}" ]] || die 3 ledger_exists "ledger exists"
      check_env_file "${OPT[env-file]}" DATABASE_NAME
      private_dir "$(dirname "${OPT[ledger]}")"
      printf 'format\t%s\t%s\t%s\t%s\n' "$LEDGER_FORMAT" "$(new_uuid)" "$(envv "${OPT[env-file]}" DATABASE_NAME)" "$(now_utc)" >"${OPT[ledger]}"
      sync -- "${OPT[ledger]}"; printf 'ledger\tinit\t%s\n' "$(ledger_check "${OPT[ledger]}" | cut -f1)" ;;
    append)
      need ledger kind refs by
      case "${OPT[kind]}" in membership.remove|membership.downgrade|profile.unlink) ;; *) die 2 invalid_input "operator kinds: membership.remove, membership.downgrade, profile.unlink" ;; esac
      refs_valid "${OPT[kind]}" "${OPT[refs]}" ||
        die 2 invalid_input "refs must be exactly the kind's keys with canonical UUIDs, closed-set roles and (downgrade) a lower role"
      [[ "${OPT[by]}" =~ ^[A-Za-z0-9._:/-]{1,128}$ ]] || die 2 invalid_input "--by must be a short reference"
      ledger_append "${OPT[ledger]}" "${OPT[kind]}" "${OPT[refs]}" operator "${OPT[by]}"
      printf 'ledger\tappend\t%s\n' "$(ledger_check "${OPT[ledger]}" | cut -f3)" ;;
    sync)
      need ledger env-file backup-root
      resolve SRC "${OPT[env-file]}"
      local n; n="$(ledger_sync "${OPT[ledger]}" "${OPT[env-file]}" "${OPT[backup-root]}")" || die 3 ledger_invalid "sync refused (ledger invalid or another database)"
      printf 'ledger\tsync\tappended=%s\n' "$n" ;;
    verify)
      need ledger
      local why
      if [[ -n "${OPT[backup]:-}" ]]; then
        why="$(ledger_verify_mark "${OPT[ledger]}" "${OPT[backup]}")" || { printf 'ledger\tverify\tfailed\t%s\n' "$why"; die 3 ledger_invalid "$why"; }
      else ledger_check "${OPT[ledger]}" >/dev/null || { printf 'ledger\tverify\tfailed\tchain\n'; die 3 ledger_invalid chain; }; fi
      printf 'ledger\tverify\tok\n' ;;
    attest)
      need ledger through-seq basis by
      local head; head="$(ledger_check "${OPT[ledger]}")" || die 3 ledger_invalid chain
      [[ "${OPT[basis]}" == final-sync || "${OPT[basis]}" == last-sync ]] || die 2 invalid_input "--basis final-sync|last-sync"
      [[ "${OPT[through-seq]}" == "$(cut -f3 <<<"$head")" ]] || die 3 ledger_invalid "--through-seq must be the ledger head"
      printf 'format\tg013-attest/1\nledger\t%s\nthrough_seq\t%s\nhead\t%s\nbasis\t%s\nby\t%s\nowner_risk\t%s\nattested\t%s\n' \
        "$(cut -f1 <<<"$head")" "${OPT[through-seq]}" "$(cut -f4 <<<"$head")" "${OPT[basis]}" "${OPT[by]}" "${OPT[owner-risk]:--}" "$(now_utc)" \
        >"${OPT[ledger]}.attest"
      printf 'ledger\tattest\t%s\t%s\n' "${OPT[through-seq]}" "${OPT[basis]}" ;;
    *) die 2 invalid_input "ledger init|append|sync|verify|attest" ;;
  esac
}

# ======================================================================================================================
# Replay (Plan r3 §5.2) and reopen (§5.4)
# ======================================================================================================================

kv() { awk -F'\t' -v k="$1" '$1 == k { print $2; exit }' <<<"$2"; }

cmd_replay() {
  need backup ledger target-env-file receipts
  local s="${OPT[backup]}" f="${OPT[ledger]}" tenv="${OPT[target-env-file]}" rdir="${OPT[receipts]}" why mark rr prev lhead hseq
  resolve TGT "$tenv" minio-may-stop; recovery_labels replay
  check_env_file "$tenv" OBJECT_STORE_ACCESS_KEY OBJECT_STORE_SECRET_KEY
  target_bound "$rdir"
  [[ "$(rcpt_get "$rdir/restore.tsv" status)" == restored-closed && "$(rcpt_get "$rdir/restore.tsv" backup_id)" == "$(basename "$s")" ]] ||
    die 3 receipt_missing "restore not restored-closed for this set"
  [[ "$(rcpt_get "$rdir/invalidate.tsv" status)" == ok ]] || die 3 receipt_missing "invalidation not done"
  bound_to_restore "$rdir/invalidate.tsv" "$rdir" || die 3 receipt_unbound "the invalidation receipt belongs to another restore"
  why="$(ledger_verify_mark "$f" "$s")" || die 3 ledger_invalid "$why"
  for svc in core web intelligence; do [[ -z "$(cid_of "$TGT_PROJECT" "$svc")" ]] || die 3 target_open "target $svc is running"; done
  mark="$(fence_get "$s/manifest.tsv" ledger | cut -f2)"
  # The replay covers exactly the validated ledger head read here (Codex review C1): entries appended later are not
  # applied by this run, and a head that changed before the end makes the run incomplete.
  lhead="$(ledger_check "$f")" || die 3 ledger_invalid chain
  hseq="$(cut -f3 <<<"$lhead")"
  # Drill seam (inert unless set; never used in operation): a bounded pause after the head is read, so the validation
  # drill can append to the ledger during a replay (Codex review C1, ledger drift).
  if int_in "${RECOVERY_DRILL_PAUSE_IN_REPLAY:-0}" 1 120; then sleep "$RECOVERY_DRILL_PAUSE_IN_REPLAY"; fi
  rr="$rdir/replay.tsv"; prev="$rdir/.replay.prev"
  # Earlier gaps carry forward only from a replay receipt of this same restore (S-review F1).
  if bound_to_restore "$rr" "$rdir"; then cp -- "$rr" "$prev"; else : >"$prev"; fi
  { printf 'format\tg013-replay/1\nbackup_id\t%s\nmark\t%s\nstarted\t%s\n' "$(basename "$s")" "$mark" "$(now_utc)"
    printf 'restore_binding\t%s\ntarget_project\t%s\ntarget_pg\t%s\ntarget_sysid\t%s\nexecuted_by\t%s\n' "$(restore_binding "$rdir")" \
      "$TGT_PROJECT" "$TGT_PG" "$(pg_sysid "$TGT_PG" "$TGT_DB")" "$(id -un)"
    printf 'ledger_uuid\t%s\nledger_through_seq\t%s\nledger_head\t%s\n' "$(cut -f1 <<<"$lhead")" "$hseq" "$(cut -f4 <<<"$lhead")"
    awk -F'\t' '$1 == "gap"' "$prev"; } >"$rr.new"
  local seq _p _t kind ck refs _o _b out st n=0 bad=0 org obj del u at gap ind
  # fd 3: the docker clients in the body inherit stdin and must not consume the ledger lines.
  while IFS=$'\t' read -r -u 3 seq _p _t kind refs _o _b; do
    (( seq > mark && seq <= hseq )) || continue
    n=$((n + 1)); st=failed; gap=""
    # Strict references (S-review F2/F3): anything else is a conflict, never already_applied.
    ck="$kind"; refs_valid "$kind" "$refs" || ck=invalid_refs
    case "$ck" in
      object.delete)
        org="$(ref_get "$refs" organization)"; obj="$(ref_get "$refs" object)"; del="$(ref_get "$refs" deleter)"
        if out="$(sqlf "$TGT_PG" "$TGT_DB" replay-object-start.sql -v org="$org" -v obj="$obj" -v deleter="$del" 2>/dev/null)"; then
          case "$(kv prior "$out")" in
            deleted) st=already_applied ;;
            absent) st=conflict ;;
            ready|pending|deleting)
              if [[ "$(kv prior "$out")" != deleting && "$(kv updated "$out")" != 1 ]]; then st=conflict
              elif [[ "$TGT_MINIO_UP" == 1 ]] && mcx "$TGT_NET" "$TGT_MCIMG" "$tenv" svc "mc rm -q 'x/$TGT_BUCKET/objects/$org/$obj' >/dev/null 2>&1; mc stat 'x/$TGT_BUCKET/objects/$org/$obj' 2>&1 | grep -q 'Object does not exist'" >/dev/null 2>&1 &&
                   out="$(sqlf "$TGT_PG" "$TGT_DB" replay-object-finish.sql -v org="$org" -v obj="$obj" -v deleter="$del" 2>/dev/null)" &&
                   [[ "$(kv state "$out")" == deleted ]]; then st=applied
              else st=partial; fi ;;
          esac
        fi ;;
      user.disable)
        u="$(ref_get "$refs" user)"; at="$(ref_get "$refs" at)"
        if out="$(sqlf "$TGT_PG" "$TGT_DB" replay-user-disable.sql -v user="$u" -v at="$at" 2>/dev/null)"; then
          if [[ "$(kv updated "$out")" == 1 ]]; then
            st=applied
            if [[ "$(kv gap "$out")" != - ]]; then
              if sqlf "$TGT_PG" "$TGT_DB" replay-unaudited.sql -v user="$u" >/dev/null 2>&1; then ind=recorded; else ind=failed; fi
              gap="gap"$'\t'"$seq"$'\t'"user.disable"$'\t'"independent=$ind"
            elif [[ "$(kv event "$out")" != 1 ]]; then gap="gap"$'\t'"$seq"$'\t'"user.disable"$'\t'"independent=not_attempted"; fi
          elif [[ "$(kv state "$out")" == disabled ]]; then st=already_applied
          else st=conflict; fi
        else
          # A failed or unknown COMMIT: the rerun reads the actual state.
          gap="gap"$'\t'"$seq"$'\t'"user.disable"$'\t'"independent=not_attempted"
        fi ;;
      policy.disable|membership.remove|membership.downgrade|profile.unlink)
        local v=(-v kind="$kind" -v org="$(ref_get "$refs" organization)" -v subject="$(ref_get "$refs" subject)" -v scope_type="$(ref_get "$refs" scope_type)"
                 -v scope_id="$(ref_get "$refs" scope_id)" -v role="$(ref_get "$refs" role)" -v from_role="$(ref_get "$refs" from_role)"
                 -v to_role="$(ref_get "$refs" to_role)" -v grp="$(ref_get "$refs" group)" -v profile="$(ref_get "$refs" profile)")
        local i; for ((i = 1; i < ${#v[@]}; i += 2)); do [[ "${v[$i]}" == *= ]] && v[$i]="${v[$i]}-"; done
        if out="$(sqlf "$TGT_PG" "$TGT_DB" replay-restrict.sql "${v[@]}" 2>/dev/null)"; then
          if [[ "$(kv updated "$out")" == 1 ]]; then st=applied
          else
            case "$kind:$(kv state "$out")" in
              policy.disable:disabled|membership.remove:none|profile.unlink:unlinked) st=already_applied ;;
              membership.downgrade:"$(ref_get "$refs" to_role)") st=already_applied ;;
              *) st=conflict ;;
            esac
          fi
        fi ;;
      *) st=conflict ;;
    esac
    [[ "$st" == applied || "$st" == already_applied ]] || bad=1
    printf 'entry\t%s\t%s\t%s\n' "$seq" "$kind" "$st" >>"$rr.new"
    [[ -n "$gap" ]] && ! grep -qF "$gap" "$rr.new" && printf '%s\n' "$gap" >>"$rr.new"
  done 3< <(tail -n +2 "$f")
  local status=complete drift=0
  [[ $bad == 0 ]] || status=incomplete
  [[ "$(ledger_check "$f" 2>/dev/null)" == "$lhead" ]] || { drift=1; status=incomplete; }
  grep -q $'^gap\t' "$rr.new" && [[ "$status" == complete ]] && status=gap
  [[ $drift == 0 ]] || printf 'ledger_drift\tthe ledger changed during the replay\n' >>"$rr.new"
  printf 'entries\t%s\nstatus\t%s\nfinished\t%s\n' "$n" "$status" "$(now_utc)" >>"$rr.new"
  mv -- "$rr.new" "$rr"; rm -f -- "$prev"; sync -- "$rr"
  printf 'replay\t%s\tentries=%s\t%s\n' "$status" "$n" "$(awk -F'\t' '$1 == "entry" { c[$4]++ } END { for (k in c) printf "%s=%d ", k, c[k] }' "$rr")"
  [[ $drift == 0 ]] || die 4 ledger_drift "the ledger changed during the replay; replay again (the target stays closed)"
  [[ $bad == 0 ]] || die 4 replay_incomplete "an entry was not applied; the target stays closed"
}

cmd_reopen() {
  need backup ledger receipts sign-off runtime-verification
  local s="${OPT[backup]}" f="${OPT[ledger]}" rdir="${OPT[receipts]}" why att head mark rr gaps
  [[ "${OPT[sign-off]}" =~ ^[A-Za-z0-9._:/-]{1,128}$ ]] || die 2 invalid_input "--sign-off must be a short reference"
  [[ "$(rcpt_get "$rdir/restore.tsv" status)" == restored-closed ]] || die 3 reopen_blocked "restore not restored-closed"
  [[ "$(rcpt_get "$rdir/invalidate.tsv" status)" == ok ]] || die 3 reopen_blocked "invalidation missing"
  bound_to_restore "$rdir/invalidate.tsv" "$rdir" || die 3 receipt_unbound "the invalidation receipt belongs to another restore"
  [[ "$(rcpt_get "$rdir/restore.tsv" backup_id)" == "$(basename "$s")" ]] || die 3 reopen_blocked "the restore receipt is for another set"
  [[ -f "$f" ]] || die 3 reopen_blocked "ledger missing"
  why="$(ledger_verify_mark "$f" "$s")" || die 3 reopen_blocked "ledger: $why"
  att="$f.attest"; [[ -f "$att" ]] || die 3 reopen_blocked "attestation missing"
  head="$(ledger_check "$f")"
  [[ "$(rcpt_get "$att" head)" == "$(cut -f4 <<<"$head")" && "$(rcpt_get "$att" through_seq)" == "$(cut -f3 <<<"$head")" ]] ||
    die 3 reopen_blocked "attestation does not cover the ledger head"
  [[ "$(rcpt_get "$att" basis)" == final-sync || "$(rcpt_get "$att" owner_risk)" != - ]] ||
    die 3 reopen_blocked "last-sync attestation without an owner risk-acceptance reference"
  mark="$(fence_get "$s/manifest.tsv" ledger | cut -f2)"
  rr="$rdir/replay.tsv"
  if (( 10#$(cut -f3 <<<"$head") > 10#$mark )); then
    [[ -f "$rr" && "$(rcpt_get "$rr" backup_id)" == "$(basename "$s")" ]] || die 3 reopen_blocked "replay not done"
    bound_to_restore "$rr" "$rdir" || die 3 receipt_unbound "the replay receipt belongs to another restore"
    [[ "$(rcpt_get "$rr" status)" == complete || "$(rcpt_get "$rr" status)" == gap ]] || die 3 reopen_blocked "replay incomplete"
    # The replay must cover the attested head itself: a new head needs a new replay, not only a new attestation (C1).
    [[ "$(rcpt_get "$rr" ledger_uuid)" == "$(cut -f1 <<<"$head")" && "$(rcpt_get "$rr" ledger_through_seq)" == "$(cut -f3 <<<"$head")" &&
       "$(rcpt_get "$rr" ledger_head)" == "$(cut -f4 <<<"$head")" ]] || die 3 replay_stale "the replay does not cover the current ledger head; replay again"
  elif [[ -f "$rr" ]] && ! bound_to_restore "$rr" "$rdir"; then
    die 3 receipt_unbound "the replay receipt belongs to another restore"
  fi
  [[ "${OPT[runtime-verification]}" == pass ]] || die 3 reopen_blocked "runtime verification did not pass"
  gaps="$(awk -F'\t' '$1 == "gap" { printf "%s:%s:%s ", $2, $3, $4 }' "$rr" 2>/dev/null || true)"
  { printf 'format\tg013-reopen/1\nbackup_id\t%s\nsign_off\t%s\nreopened\t%s\n' "$(basename "$s")" "${OPT[sign-off]}" "$(now_utc)"
    printf 'restore_receipt\t%s\ninvalidate_receipt\t%s\nreplay_receipt\t%s\nattestation\t%s\n' "$(sha_file "$rdir/restore.tsv")" \
      "$(sha_file "$rdir/invalidate.tsv")" "$( [[ -f "$rr" ]] && sha_file "$rr" || echo -)" "$(sha_file "$att")"
    printf 'acknowledged_gaps\t%s\n' "${gaps:--}"; printf 'runtime_verification\tpass\n'; } >"$rdir/reopen.tsv"
  printf 'reopen\tsigned_off\tgaps=%s\n' "${gaps:-none}"
  printf 'start: docker compose --env-file <target .env> -f infra/compose/local.yml up -d --wait core web intelligence\n'
}

# ======================================================================================================================
# Retention (Plan r3 §6.2, §6.3, §6.5)
# ======================================================================================================================

# cutoff_past CUTOFF: 0 when the UTC cutoff is strictly before now (nanosecond precision), 1 when it is equal or later,
# 2 when it or the clock cannot be read. Each instant is read once by GNU date as epoch seconds and a nine-digit
# fraction, both validated (canonical seconds within the years 0000-9999) and compared separately in base 10: seconds
# first, fractions only on equal seconds. No concatenated nanosecond value, so four-digit years cannot overflow.
cutoff_past() {
  local c n cs cf ns nf re='^(-?[1-9][0-9]{0,11}|0)\.([0-9]{9})$'
  c="$(date -u -d "$1" +%s.%N 2>/dev/null)" || return 2
  n="$(date -u +%s.%N 2>/dev/null)" || return 2
  [[ "$c" =~ $re ]] || return 2; cs="${BASH_REMATCH[1]}" cf="${BASH_REMATCH[2]}"
  [[ "$n" =~ $re ]] || return 2; ns="${BASH_REMATCH[1]}" nf="${BASH_REMATCH[2]}"
  (( cs >= -62167219200 && cs <= 253402300799 && ns >= -62167219200 && ns <= 253402300799 )) || return 2
  if (( cs != ns )); then (( cs < ns )) && return 0; return 1; fi
  (( 10#$cf < 10#$nf )) && return 0
  return 1
}

retention_policy() {
  read_policy "$1" "policy_id policy_label scope global_authorization audit_cutoff audit_retain_for_s object_cutoff object_retain_for_s holds batch_size max_batches statement_timeout_ms lock_timeout_ms run_timeout_s" \
    "policy_id policy_label scope holds batch_size max_batches statement_timeout_ms lock_timeout_ms run_timeout_s"
  [[ "${POL[scope]}" == global || "${POL[scope]}" =~ ^organization:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$ ]] || die 2 policy_invalid "scope"
  if [[ "${POL[scope]}" == global ]]; then [[ -n "${POL[global_authorization]:-}" ]] || die 2 policy_invalid "global scope requires global_authorization"
  else [[ -z "${POL[global_authorization]+x}" ]] || die 2 policy_invalid "global_authorization only with scope=global"; fi
  check_holds "${POL[holds]}" || die 2 policy_invalid "holds"
  int_in "${POL[batch_size]}" 1 500 && int_in "${POL[max_batches]}" 1 1000 && int_in "${POL[statement_timeout_ms]}" 100 30000 &&
    int_in "${POL[lock_timeout_ms]}" 100 10000 && int_in "${POL[run_timeout_s]}" 1 3600 || die 2 policy_invalid "bounds"
  local k rc
  for k in audit object; do
    if [[ -n "${POL[${k}_cutoff]+x}" && -n "${POL[${k}_retain_for_s]+x}" ]]; then die 2 policy_invalid "both ${k}_cutoff and ${k}_retain_for_s"; fi
    if [[ -n "${POL[${k}_cutoff]+x}" ]]; then
      [[ "${POL[${k}_cutoff]}" =~ ^[0-9]{4}-(0[1-9]|1[0-2])-(0[1-9]|[12][0-9]|3[01])T([01][0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9](\.[0-9]{1,6})?Z$ ]] || die 2 policy_invalid "${k}_cutoff"
      rc=0; cutoff_past "${POL[${k}_cutoff]}" || rc=$?
      case $rc in 0) ;; 1) die 2 policy_invalid "${k}_cutoff is not in the past" ;; *) die 2 policy_invalid "${k}_cutoff" ;; esac
    fi
    if [[ -n "${POL[${k}_retain_for_s]+x}" ]]; then int_in "${POL[${k}_retain_for_s]}" 1 315360000 || die 2 policy_invalid "${k}_retain_for_s"; fi
  done
}

# ret_sql FILE VARS...: one statement file as the retention login (TCP inside the project network).
ret_sql() { local f="$1"; shift; pg_client "$SRC_NET" "$SRC_ENV" retention "$SRC_PGIMG" psql -X -q -tA -v ON_ERROR_STOP=1 "$@" <"$RECOVERY_SQL/$f"; }

cmd_retention() {
  local kind="${ARGS[0]:-}"
  [[ "$kind" == audit || "$kind" == objects ]] || die 2 invalid_input "retention audit|objects"
  need policy env-file preview
  [[ "${OPT[preview]}" =~ ^/.+/[A-Za-z0-9._-]+$ ]] || die 2 invalid_input "--preview must be an absolute file path"
  private_dir "$(dirname "${OPT[preview]}")"
  retention_policy "${OPT[policy]}"
  local pk=audit; [[ "$kind" == objects ]] && pk=object
  [[ -n "${POL[${pk}_cutoff]+x}" || -n "${POL[${pk}_retain_for_s]+x}" ]] || die 2 policy_invalid "exactly one of ${pk}_cutoff and ${pk}_retain_for_s is required"
  [[ "$kind" == audit || "${POL[scope]}" != global ]] || die 2 policy_invalid "tombstone purge needs an Organization scope"
  check_env_file "${OPT[env-file]}" DB_RETENTION_PASSWORD
  resolve SRC "${OPT[env-file]}" minio-may-stop; recovery_labels retention
  local psha holds scope org cutoff global=false
  psha="$(sha_file "${OPT[policy]}")"; holds="${POL[holds]:-}"; [[ -n "$holds" ]] || holds=none
  scope="${POL[scope]}"; org="${scope#organization:}"; [[ "$scope" == global ]] && { org=-; global=true; }
  local common=(-v scope="$scope" -v holds="$holds" -v org="$org" -v stmt_ms="${POL[statement_timeout_ms]}" -v lock_ms="${POL[lock_timeout_ms]}")
  if [[ -z "${OPT[apply]:-}" ]]; then
    # Dry-run: resolve the cutoff once from the database clock; count; write the preview. Nothing changes.
    if [[ -n "${POL[${pk}_cutoff]+x}" ]]; then cutoff="${POL[${pk}_cutoff]}"
    else cutoff="$(printf "SELECT to_char((clock_timestamp() - make_interval(secs => %d)) AT TIME ZONE 'UTC', 'YYYY-MM-DD\"T\"HH24:MI:SS.US\"Z\"');\n" "${POL[${pk}_retain_for_s]}" |
      pg_client "$SRC_NET" "$SRC_ENV" retention "$SRC_PGIMG" psql -X -q -tA -v ON_ERROR_STOP=1)" || die 3 source_unavailable "database clock unreadable"; fi
    local n
    if [[ "$kind" == audit ]]; then n="$(ret_sql retention-audit.sql "${common[@]}" -v cutoff="$cutoff" -v count=true -v global="$global" -v batch=1 | awk -F'\t' '$1 == "eligible" { print $2 }')"
    else n="$(ret_sql retention-objects.sql "${common[@]}" -v cutoff="$cutoff" -v mode=count -v batch=1 -v ids=- | awk -F'\t' '$1 == "eligible" { print $2 }')"; fi
    [[ "$n" =~ ^[0-9]+$ ]] || die 3 retention_refused "eligibility count failed"
    printf 'format\tg013-preview/1\nkind\t%s\npolicy_sha256\t%s\nscope\t%s\ncutoff\t%s\neligible\t%s\ncreated\t%s\n' "$kind" "$psha" "$scope" "$cutoff" "$n" "$(now_utc)" >"${OPT[preview]}"
    printf 'retention\t%s\tdry-run\teligible=%s\n' "$kind" "$n"
    return 0
  fi
  need receipts
  private_dir "${OPT[receipts]}"
  [[ -f "${OPT[preview]}" && "$(rcpt_get "${OPT[preview]}" kind)" == "$kind" ]] || die 3 preview_missing "apply requires the preview record"
  [[ "$(rcpt_get "${OPT[preview]}" policy_sha256)" == "$psha" ]] || die 3 policy_changed "the policy changed since the preview"
  cutoff="$(rcpt_get "${OPT[preview]}" cutoff)"
  local pcount now_n; pcount="$(rcpt_get "${OPT[preview]}" eligible)"
  if [[ "$kind" == audit ]]; then now_n="$(ret_sql retention-audit.sql "${common[@]}" -v cutoff="$cutoff" -v count=true -v global="$global" -v batch=1 | awk -F'\t' '$1 == "eligible" { print $2 }')"
  else now_n="$(ret_sql retention-objects.sql "${common[@]}" -v cutoff="$cutoff" -v mode=count -v batch=1 -v ids=- | awk -F'\t' '$1 == "eligible" { print $2 }')"; fi
  [[ "$now_n" =~ ^[0-9]+$ ]] || die 3 retention_refused "eligibility count failed"
  (( now_n <= pcount )) || die 3 eligible_grew "more rows are eligible than the preview counted"
  local stop=0 b=0 deleted=0 got outcome=complete t_end=$((SECONDS + 10#${POL[run_timeout_s]})) skipped="" failure=""
  trap 'stop=1' TERM INT
  local rcpt; rcpt="${OPT[receipts]}/maintenance-$(date -u +%Y%m%dT%H%M%S)-$kind.tsv"
  while (( b < 10#${POL[max_batches]} )); do
    (( stop == 0 && SECONDS < t_end )) || { outcome=partial; break; }
    b=$((b + 1))
    if [[ "$kind" == audit ]]; then
      got="$(ret_sql retention-audit.sql "${common[@]}" -v cutoff="$cutoff" -v count=false -v global="$global" -v batch="${POL[batch_size]}" 2>/dev/null |
             awk -F'\t' '$1 == "deleted" { print $2 }')" || got=""
      [[ "$got" =~ ^[0-9]+$ ]] || { failure=batch_failed; outcome=partial; break; }
      deleted=$((deleted + got))
      (( got < 10#${POL[batch_size]} )) && break
    else
      local cands confirmed="" c ledger="${OPT[ledger]:-}" npres=0 nbytes=0 nprov=0
      cands="$(ret_sql retention-objects.sql "${common[@]}" -v cutoff="$cutoff" -v mode=candidates -v batch="${POL[batch_size]}" -v ids=- 2>/dev/null |
               awk -F'\t' '$1 == "candidate" { print $2 }')" || { failure=batch_failed; outcome=partial; break; }
      [[ -n "$cands" ]] || break
      for c in $cands; do
        if [[ -z "$ledger" ]] || ! ledger_check "$ledger" >/dev/null ||
           ! awk -F'\t' -v o="object=$c" 'NR > 1 && $4 == "object.delete" && index("," $5 ",", "," o ",") { f = 1 } END { exit !f }' "$ledger"; then
          npres=$((npres + 1)); continue
        fi
        local pr=unavailable; [[ "$SRC_MINIO_UP" == 1 ]] && pr="$(mcx "$SRC_NET" "$SRC_MCIMG" "${OPT[env-file]}" svc "mc stat 'x/$SRC_BUCKET/objects/$org/$c' >/dev/null 2>/tmp/e && echo present || { grep -q 'Object does not exist' /tmp/e && echo absent || echo unavailable; }" 2>/dev/null || echo unavailable)"
        case "$pr" in absent) confirmed+="${confirmed:+,}$c" ;; present) nbytes=$((nbytes + 1)) ;; *) nprov=$((nprov + 1)) ;; esac
      done
      skipped="not_preserved=$npres bytes_present=$nbytes provider_unavailable=$nprov"
      if [[ -n "$confirmed" ]]; then
        got="$(ret_sql retention-objects.sql "${common[@]}" -v cutoff="$cutoff" -v mode=delete -v batch=1 -v ids="$confirmed" 2>/dev/null | grep -c $'^deleted\t')" || got=""
        [[ "$got" =~ ^[0-9]+$ ]] || { failure=batch_failed; outcome=partial; break; }
        deleted=$((deleted + got))
      else got=0; fi
      (( npres + nbytes + nprov > 0 )) && outcome=partial
      (( got == 0 )) && break
    fi
  done
  trap - TERM INT
  local remaining
  if [[ "$kind" == audit ]]; then remaining="$(ret_sql retention-audit.sql "${common[@]}" -v cutoff="$cutoff" -v count=true -v global="$global" -v batch=1 2>/dev/null | awk -F'\t' '$1 == "eligible" { print $2 }')"
  else remaining="$(ret_sql retention-objects.sql "${common[@]}" -v cutoff="$cutoff" -v mode=count -v batch=1 -v ids=- 2>/dev/null | awk -F'\t' '$1 == "eligible" { print $2 }')"; fi
  # Eligible rows left (max_batches, the run timeout or a signal) are never reported complete (S-review F8).
  [[ "${remaining:-1}" != 0 && "$outcome" == complete ]] && outcome=partial
  [[ -n "$failure" && $deleted == 0 ]] && outcome=failed
  printf 'format\tg013-maintenance/1\nkind\t%s\npolicy_id\t%s\npolicy_sha256\t%s\nscope\t%s\ncutoff\t%s\nbatches\t%s\ndeleted\t%s\nremaining\t%s\nskipped\t%s\nfailure\t%s\noutcome\t%s\nfinished\t%s\n' \
    "$kind" "${POL[policy_id]}" "$psha" "$scope" "$cutoff" "$b" "$deleted" "${remaining:-unknown}" "${skipped:--}" "${failure:--}" "$outcome" "$(now_utc)" >"$rcpt"
  printf 'retention\t%s\t%s\tbatches=%s\tdeleted=%s\tremaining=%s\t%s\n' "$kind" "$outcome" "$b" "$deleted" "${remaining:-unknown}" "${skipped:--}"
  [[ "$outcome" == complete ]] || exit 1
}

# ======================================================================================================================
# Backup expiry and quarantine (Plan r3 §6.6)
# ======================================================================================================================

cmd_quarantine() {
  need backup reason
  [[ "${OPT[reason]}" =~ ^[a-z_]{1,40}$ ]] || die 2 invalid_input "--reason must be a closed code"
  [[ -d "${OPT[backup]}" ]] || die 2 invalid_input "no such set"
  printf 'reason\t%s\nquarantined\t%s\n' "${OPT[reason]}" "$(now_utc)" >"${OPT[backup]}/QUARANTINED"
  printf 'quarantine\t%s\t%s\n' "$(basename "${OPT[backup]}")" "${OPT[reason]}"
}

cmd_expire() {
  need backup-root backup-policy
  backup_policy "${OPT[backup-policy]}"
  local s id rc=0 item
  for s in "${OPT[backup-root]}"/*/; do
    s="${s%/}"; id="$(basename "$s")"
    [[ "$id" =~ $UUID_RE && -f "$s/manifest.tsv" ]] || continue
    [[ -e "$s/EXPIRED" && ! -e "$s/EXPIRED-PARTIAL" ]] && continue
    if [[ ! -e "$s/EXPIRED-PARTIAL" ]] && ! set_expirable "$s"; then printf 'expire\t%s\tkept\n' "$id"; continue; fi
    if [[ -z "${OPT[apply]:-}" ]]; then printf 'expire\t%s\teligible\n' "$id"; continue; fi
    printf 'expired\t%s\n' "$(now_utc)" >"$s/EXPIRED"
    local left=""
    for item in database.dump objects.tar objects.meta toc.txt backup-receipt.tsv; do
      [[ -e "$s/$item" ]] || continue
      rm -f -- "$s/$item" 2>/dev/null || true
      [[ -e "$s/$item" ]] && left+="$item "
    done
    if [[ -n "$left" ]]; then
      printf 'remaining\t%s\n' "$left" >"$s/EXPIRED-PARTIAL"; rc=1
      printf 'expire\t%s\texpired-partial\tremaining=%s\tother_copies=unknown\n' "$id" "${left% }"
    else
      # Nothing of the set remains but its EXPIRED record (Plan r3 §6.6; S-review F11).
      rm -f -- "$s/EXPIRED-PARTIAL" "$s/manifest.tsv" "$s/COMPLETE" "$s/INCOMPLETE" "$s/QUARANTINED"
      printf 'expire\t%s\texpired\tonly the local copy is managed; provider, replicated or immutable copies are unknown\n' "$id"
    fi
  done
  return $rc
}

# ======================================================================================================================
# Status hook (Plan r3 §7.1): bounded, content-free JSON lines; exit 0 only when every check is ok or none.
# ======================================================================================================================

ST_BAD=0 ST_SINK=""
st_line() { # st_line CHECK STATUS REASON [BACKUP]
  local l
  l="$(printf '{"time":"%s","check":"%s","status":"%s","reason":"%s"%s}' "$(now_utc)" "$1" "$2" "$3" "${4:+,\"backup\":\"$4\"}")"
  printf '%s\n' "$l"
  [[ -n "$ST_SINK" ]] && printf '%s\n' "$l" >>"$ST_SINK"
  [[ "$2" == ok || "$2" == none ]] || ST_BAD=1
}

cmd_status() {
  need backup-root backup-policy receipts
  local root="${OPT[backup-root]}" rdir="${OPT[receipts]}" s latest="" lt="" t pol=ok
  STEP_TIMEOUT=30
  if [[ -n "${OPT[sink]:-}" ]]; then
    [[ "${OPT[sink]}" =~ ^/.+/[A-Za-z0-9._-]+$ ]] || die 2 invalid_input "--sink must be an absolute file path"
    private_dir "$(dirname "${OPT[sink]}")"; ST_SINK="${OPT[sink]}"; ( umask 077; : >>"$ST_SINK" ); chmod 600 "$ST_SINK"; fi
  if ( backup_policy "${OPT[backup-policy]}" ) >/dev/null 2>&1; then backup_policy "${OPT[backup-policy]}"; else pol=unknown; fi
  st_line policy "$pol" "$([[ $pol == ok ]] && echo valid || echo missing_or_invalid)"
  for s in "$root"/*/; do
    s="${s%/}"
    [[ "$(basename "$s")" =~ $UUID_RE && -f "$s/COMPLETE" && ! -e "$s/QUARANTINED" && ! -e "$s/EXPIRED" ]] || continue
    t="$(fence_get "$s/manifest.tsv" created)"
    [[ -z "$lt" || "$t" > "$lt" ]] && { lt="$t"; latest="$s"; }
  done
  if [[ -z "$latest" ]]; then st_line backup.latest missing no_complete_set; st_line backup.integrity unknown no_complete_set
  else
    if [[ $pol == ok ]] && (( $(set_age_s "$latest") > 10#${POL[backup_max_age_s]} )); then st_line backup.latest stale older_than_max_age "$(basename "$latest")"
    elif [[ $pol == ok ]]; then st_line backup.latest ok fresh "$(basename "$latest")"
    else st_line backup.latest unknown policy_unknown "$(basename "$latest")"; fi
    if verify_set "$latest" >/dev/null 2>&1; then st_line backup.integrity ok verified "$(basename "$latest")"
    else st_line backup.integrity failed component_mismatch "$(basename "$latest")"; fi
  fi
  local held=0 f
  for f in "$root"/fences/*.tsv; do [[ -f "$f" ]] && ! fence_has "$f" released && held=1; done
  if [[ $held == 1 ]]; then st_line backup.fence held unreleased_fence; else st_line backup.fence ok no_unreleased_fence; fi
  local rst; rst="$(rcpt_get "$rdir/restore.tsv" status)"
  case "$rst" in restored-closed) st_line restore.last ok restored_closed ;; "") st_line restore.last none never_run ;;
    partial) st_line restore.last partial restore_partial ;; *) st_line restore.last failed restore_failed ;; esac
  # replay.last: an audit gap not yet covered by a reopen sign-off is reported first (it is never healthy), then an entry
  # not applied.
  local rp; rp="$(rcpt_get "$rdir/replay.tsv" status)"
  if [[ -z "$rp" ]]; then st_line replay.last none never_run
  elif grep -q $'^gap\t' "$rdir/replay.tsv" && ! [[ -f "$rdir/reopen.tsv" && "$(rcpt_get "$rdir/reopen.tsv" replay_receipt)" == "$(sha_file "$rdir/replay.tsv")" ]]; then
    st_line replay.last gap audit_gap_unsigned
  elif [[ "$rp" == complete || "$rp" == gap ]]; then st_line replay.last ok "$([[ "$rp" == gap ]] && echo gap_signed_off || echo complete)"
  else st_line replay.last partial entries_not_applied; fi
  local mr; mr="$(find "$rdir" -maxdepth 1 -name 'maintenance-*.tsv' 2>/dev/null | LC_ALL=C sort | tail -1 || true)"
  if [[ -z "$mr" ]]; then st_line maintenance.last none never_run
  else case "$(rcpt_get "$mr" outcome)" in complete) st_line maintenance.last ok complete ;; partial) st_line maintenance.last partial bound_or_skip ;;
         *) st_line maintenance.last failed failed ;; esac; fi
  if [[ -n "${OPT[env-file]:-}" ]]; then
    # One subshell: a refused resolve (die) ends only the subshell, and status reports the source as unreadable.
    if out="$(resolve SRC "${OPT[env-file]}" minio-may-stop 2>/dev/null &&
              sqlf "$SRC_PG" "$SRC_DB" status-counts.sql -v pending_after="${POL[pending_review_after_s]:-86400}" 2>/dev/null)"; then
      local dl po nt=0 id
      dl="$(kv deleting "$out")"; po="$(kv pending_old "$out")"
      while IFS=$'\t' read -r _ id; do
        [[ -n "$id" ]] || continue
        if [[ -z "${OPT[ledger]:-}" ]] || ! awk -F'\t' -v o="object=$id" 'NR > 1 && $4 == "object.delete" && index("," $5 ",", "," o ",") { f = 1 } END { exit !f }' "${OPT[ledger]}"; then nt=$((nt + 1)); fi
      done < <(awk -F'\t' '$1 == "tombstone"' <<<"$out")
      if [[ "$dl" == 0 && "$po" == 0 && "$nt" == 0 ]]; then st_line maintenance.pending_review ok none_pending
      else st_line maintenance.pending_review pending "deleting=$dl,pending_old=$po,tombstones_unpreserved=$nt"; fi
    else st_line maintenance.pending_review unknown database_unreadable; fi
  fi
  local summary=ok; [[ $ST_BAD == 0 ]] || summary=unhealthy
  st_line summary "$summary" "$([[ $ST_BAD == 0 ]] && echo all_ok_or_none || echo see_checks)"
  [[ $ST_BAD == 0 ]] || exit 1
}

# ======================================================================================================================

ORIG_ARGV=("$@")
cmd="${1:-}"; shift || true
parse_opts "$@"
case "$cmd" in
  fence) cmd_fence ;; unfence) cmd_unfence ;; backup) cmd_backup ;; verify) cmd_verify ;; restore) cmd_restore ;;
  invalidate-auth) cmd_invalidate ;; ledger) cmd_ledger ;; replay) cmd_replay ;; reopen) cmd_reopen ;;
  retention) cmd_retention ;; status) cmd_status ;; quarantine) cmd_quarantine ;; expire) cmd_expire ;;
  *) sed -n '2,22p' "${BASH_SOURCE[0]}" >&2; exit 2 ;;
esac
