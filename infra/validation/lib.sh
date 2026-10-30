#!/usr/bin/env bash
# Shared helpers for infra/validation/run.sh. Host requirements: bash, GNU
# coreutils/findutils/grep (stat -c, find -printf, grep -P), git and Docker
# (Compose >= 2.24). No host Go, Python, Node or jq is used. Host probes use
# bash /dev/tcp; host curl is used only as a fallback when /dev/tcp is absent.
#
# Originally adapted from the archetype Skill's validation helpers (Skill 0.1.0). These helpers are project-owned.
# shellcheck disable=SC2034  # variables are consumed by run.sh

YQ_IMAGE="mikefarah/yq:4.48.1"
BUSYBOX_IMAGE="busybox:1.37.0"
SHELLCHECK_IMAGE="koalaman/shellcheck:v0.11.0"
HADOLINT_IMAGE="hadolint/hadolint:v2.15.1"

# Runtimes the project must never use from the host.
SHADOWED_RUNTIMES=(go gofmt python python3 pip pip3 pytest uv uvicorn node npm npx corepack pnpm yarn bun deno psql postgres pg_ctl redis-cli redis-server mc minio aws)

log() { printf '[%s] %s\n' "$(date +%H:%M:%S)" "$*"; }

# --- results -------------------------------------------------------------------
# record ID STATUS MESSAGE   (STATUS: PASS | FAIL | UNMET | INFO)
# A background child (bg_start) sets its own RESULTS and TIMINGS (G051).
record() {
  local id="$1" status="$2" msg="${3:-}"
  printf '%s\t%s\t%s\n' "$id" "$status" "$msg" >>"$RESULTS"
  printf 'check\t%s\t%s\n' "$id" "$(date +%s)" >>"${TIMINGS:-$WORK/timings.tsv}"
  log "$status $id ${msg:+- $msg}"
  # G066: ID<TAB>STATUS (never the message) for validate.sh's live N/TOTAL progress line; an append error changes nothing.
  if [[ -n "${PLATFORM_PROGRESS_FILE:-}" ]]; then { printf '%s\t%s\n' "$id" "$status" >>"$PLATFORM_PROGRESS_FILE"; } 2>/dev/null || :; fi
}

# Keep bounded failed-run diagnostics before owned work directories are removed. Only generated-secret-redacted
# text logs are exported; the CI raw and candidate scanners still gate publication. No env/dump/receipt files.
export_diagnostics() {
  [[ -n "${REPORT_DIR:-}" ]] || return 0
  if [[ "${RUN_STATUS:-0}" == 0 ]]; then grep -qE $'\t(FAIL|UNMET)\t' "$RESULTS" || return 0; fi
  local f name out="$REPORT_DIR/diagnostics"
  mkdir -p "$out" || return 1
  for f in "$WORK"/*.log "$WORK"/*.log.cmds "$WORK"/recovery/logs/*.log "$WORK"/recovery-steps.tsv; do
    [[ -f "$f" && ! -L "$f" ]] || continue
    name="${f#"$WORK"/}"; name="${name//\//_}"
    [[ "$name" =~ ^[A-Za-z0-9_.-]{1,120}$ ]] || continue
    redact_log "$f" "$out/$name" || return 1
  done
  log "failure diagnostics exported (bounded, generated secrets redacted): diagnostics/"
}

# redact_log SRC DST: SRC with every registered generated secret replaced by [REDACTED], bounded to 1 MiB (first and
# final 512 KiB), written to DST through a temporary file in DST's directory (G064 Plan r2 §5.3: factored from
# export_diagnostics, behaviour unchanged; also used for the per-invocation go-tests/<NAME>.log).
redact_log() {
  local f="$1" dst="$2" clean
  clean="$(dirname -- "$dst")/.redacting"
  awk 'FILENAME != "-" { if (length($0)) secret[++n]=$0; next }
      { line=$0; for(i=1;i<=n;i++) { replacement=""; rest=line;
          while((pos=index(rest,secret[i]))>0) { replacement=replacement substr(rest,1,pos-1) "[REDACTED]"; rest=substr(rest,pos+length(secret[i])); }
          line=replacement rest; }
        print line }' "${SECRETS_FILE:-/dev/null}" - <"$f" >"$clean" || return 1
  if (( $(stat -c %s "$clean") > 1048576 )); then
    { head -c 524288 "$clean"; printf '\n[diagnostic log truncated; final 512 KiB follows]\n'; tail -c 524288 "$clean"; } >"$dst"
  else mv -- "$clean" "$dst"; fi
  rm -f -- "$clean"
}

# --- G064 per-test Go evidence (Plan r2 §5.1-§5.3) ---------------------------------------
# Columns of $WORK/go-tests.tsv and REPORT_DIR/go-tests.tsv (the converter's Header, infra/validation/tools/internal/
# gotest/report.go). Invocations register in $WORK/go-tests/started.tsv before they run (go_it, run.sh); a converted
# and exported invocation leaves $WORK/go-tests/NAME.done.
GO_TESTS_HEADER=$'kind\tinvocation\tpackage\ttest\tstatus\trc\texpected\tpassed\tfailed\tunfinished\telapsed_s\tlog_file\tline_from\tline_to\tdetail'

# go_tests_export NAME: append NAME's rows to $WORK/go-tests.tsv, regenerate $WORK/go-tests-junit.xml from the converted
# invocations in start order, and publish both, db-lifecycle.tsv and the redacted, bounded go-tests/NAME.log to
# REPORT_DIR (temporary file + mv). Raw .events files never leave $WORK.
go_tests_export() {
  local name="$1" d="$WORK/go-tests" n tmp
  [[ -s "$d/$name.tsv" && -f "$d/$name.log" ]] || return 1
  [[ -s "$WORK/go-tests.tsv" ]] || printf '%s\n' "$GO_TESTS_HEADER" >"$WORK/go-tests.tsv"
  cat -- "$d/$name.tsv" >>"$WORK/go-tests.tsv"
  { printf '<?xml version="1.0" encoding="UTF-8"?>\n<testsuites>\n'
    while IFS=$'\t' read -r n _; do
      if [[ -f "$d/$n.done" || "$n" == "$name" ]] && [[ -f "$d/$n.junit.xml" ]]; then cat -- "$d/$n.junit.xml"; fi
    done <"$d/started.tsv"
    printf '</testsuites>\n'; } >"$WORK/go-tests-junit.xml"
  [[ -n "${REPORT_DIR:-}" ]] || return 0
  mkdir -p "$REPORT_DIR/go-tests" || return 1
  redact_log "$d/$name.log" "$REPORT_DIR/go-tests/$name.log" || return 1
  for n in go-tests.tsv go-tests-junit.xml; do
    tmp="$REPORT_DIR/.$n.$BASHPID"
    { cp -- "$WORK/$n" "$tmp" && mv -f -- "$tmp" "$REPORT_DIR/$n"; } || return 1
  done
  db_lifecycle_export
}

# go_tests_declare SUITE NAME...: the suite's declared invocation inventory (REPORT_DIR/go-tests-expected.tsv rows
# `suite<TAB>invocation`), written before any invocation of the suite; validate.sh's complete step requires one
# pass/fail invocation row in go-tests.tsv for each.
go_tests_declare() {
  local suite="$1" n tmp; shift
  [[ -s "$WORK/go-tests-expected.tsv" ]] || printf 'suite\tinvocation\n' >"$WORK/go-tests-expected.tsv"
  for n in "$@"; do printf '%s\t%s\n' "$suite" "$n" >>"$WORK/go-tests-expected.tsv"; done
  [[ -n "${REPORT_DIR:-}" ]] || return 0
  tmp="$REPORT_DIR/.go-tests-expected.tsv.$BASHPID"
  cp -- "$WORK/go-tests-expected.tsv" "$tmp" && mv -f -- "$tmp" "$REPORT_DIR/go-tests-expected.tsv"
}

# go_tests_finalize (EXIT trap, before export_diagnostics and cleanup): every registered invocation that was not
# converted (its container was stopped by a signal or the group stop) is converted from the events observed so far with
# status interrupted (terminal rows kept, started tests unfinished) and exported. An interrupted run is never green:
# RUN_STATUS is non-zero and validate.sh's complete step refuses an interrupted invocation.
go_tests_finalize() {
  local reg="${WORK:-}/go-tests/started.tsv" name logf
  [[ -n "${WORK:-}" && -f "$reg" ]] || return 0
  while IFS=$'\t' read -r name logf; do
    [[ -n "$name" && ! -f "$WORK/go-tests/$name.done" ]] || continue
    log "G064: invocation $name did not end; finalized as interrupted"
    go_it_convert "$name" "$logf" 255 interrupted || true
  done <"$reg"
}

# obj_fail_codes LOG N: the closed failed-subtest codes of a B-OBJ-IT failure (G014 object diagnostic publication),
# for the existing results.tsv row only. It reads nothing but the top-level and first-level `--- FAIL:`/`--- PASS:`
# header lines of the go test -v log of TestObjects and TestObjectsFootprint, maps each failed subtest's exact name
# to one reviewed code, and prints exactly one token list: the codes in canonical order (then one `unknown` for any
# failed subtest outside the table), or `unknown`, or `none`. N is it_count's TestObjects value. `none` needs N == 10
# and an explicit top-level PASS of both tests; a missing, skipped, truncated or panic-ended result is `unknown`.
# A missing, unreadable, empty, NUL-bearing or oversize log, or any helper error, is `unknown`. It always exits 0 and
# never echoes input, so under set -e it cannot abort the FAIL record or change the status decided before it.
obj_fail_codes() {
  local log="${1:-}" n="${2:-}" size nuls hdr rc line name unknown=0 pfail=0 ppass=0 ffail=0 fpass=0 out="" c
  local safe='^[A-Za-z0-9_,:;-]+$'
  local -a order=(OBJ-ATTEST OBJ-READ-WRITE OBJ-RLS OBJ-RT OBJ-SCOPE OBJ-SCHEMA OBJ-DEL OBJ-SIGN OBJ-ADMIN OBJ-LOG)
  local -A table=(
    [OBJ-ATTEST_runtime_login_and_scoped_credential]=OBJ-ATTEST
    [OBJ-READ_and_OBJ-WRITE:_the_oracle_service_matrix]=OBJ-READ-WRITE
    [OBJ-RLS:_the_oracle_RLS_matrix,_guard_transitions,_FORCE_owner,_poisoned_context]=OBJ-RLS
    [OBJ-RT_round_trips]=OBJ-RT
    [OBJ-SCOPE_scope_discovery_is_constant_in_the_binding_count]=OBJ-SCOPE
    [OBJ-SCHEMA_constraints,_generated_key_and_forged_binding]=OBJ-SCHEMA
    [OBJ-DEL_physical_deletion,_tombstone_and_idempotent_repeat]=OBJ-DEL
    [OBJ-SIGN_real_fetch,_expiry,_alteration_and_loopback_binding]=OBJ-SIGN
    [OBJ-ADMIN_the_scoped_credential_administers_nothing;_anonymous_access_is_denied]=OBJ-ADMIN
    [OBJ-LOG_records_catalogued,_no_canary_or_secret]=OBJ-LOG
  )
  local -A seen=()
  if [[ -z "$log" || -L "$log" || ! -f "$log" || ! -r "$log" || ! -s "$log" ]]; then printf 'unknown\n'; return 0; fi
  size="$(stat -c %s -- "$log" 2>/dev/null)" || { printf 'unknown\n'; return 0; }
  if [[ ! "$size" =~ ^[0-9]+$ ]] || (( size > 67108864 )); then printf 'unknown\n'; return 0; fi
  nuls="$(tr -d '\000' <"$log" 2>/dev/null | wc -c)" || { printf 'unknown\n'; return 0; }
  if [[ "$nuls" != "$size" ]]; then printf 'unknown\n'; return 0; fi
  rc=0
  hdr="$(LC_ALL=C grep -a -E -- '^(    --- FAIL: TestObjects/[^ ]+|--- (FAIL|PASS): TestObjects(Footprint)?) \([0-9]+\.[0-9]+s\)$' "$log" 2>/dev/null)" || rc=$?
  if (( rc > 1 )); then printf 'unknown\n'; return 0; fi
  while IFS= read -r line; do
    case "$line" in
      '    --- FAIL: TestObjects/'*)
        name="${line#    --- FAIL: TestObjects/}"; name="${name% (*}"
        # Only a name of the table's own characters is looked up (never `@`, `*` or a bracket as a subscript).
        if [[ "$name" =~ $safe && -n "${table[$name]+x}" ]]; then seen[${table[$name]}]=1; else unknown=1; fi ;;
      '--- FAIL: TestObjects ('*) pfail=1 ;;
      '--- PASS: TestObjects ('*) ppass=1 ;;
      '--- FAIL: TestObjectsFootprint ('*) ffail=1 ;;
      '--- PASS: TestObjectsFootprint ('*) fpass=1 ;;
    esac
  done <<<"$hdr"
  for c in "${order[@]}"; do if [[ -n "${seen[$c]+x}" ]]; then out+="${out:+,}$c"; fi; done
  if (( pfail )) && [[ -z "$out" ]] && (( unknown == 0 )); then out+="${out:+,}PARENT"; fi
  if (( ffail )); then out+="${out:+,}FOOTPRINT"; fi
  if (( unknown )); then out+="${out:+,}unknown"; fi
  if [[ -n "$out" ]]; then printf '%s\n' "$out"; return 0; fi
  if [[ "$n" == 10 ]] && (( ppass && fpass )); then printf 'none\n'; return 0; fi
  printf 'unknown\n'
  return 0
}

# check ID MESSAGE COMMAND...: records PASS/FAIL and always returns 0, so one
# failing check never aborts the remaining checks.
check() {
  local id="$1" msg="$2"; shift 2
  if "$@"; then record "$id" PASS "$msg"; else record "$id" FAIL "$msg"; fi
  return 0
}

# --- PATH hygiene and runtime shadows (B-DOCKER-ONLY) ---------------------------
setup_path_hygiene() {
  local cleaned="" entry r name parts=()
  IFS=':' read -r -a parts <<<"$PATH"
  for entry in "${parts[@]}"; do
    case "$entry" in /mnt/*|"") continue ;; esac
    cleaned="${cleaned:+$cleaned:}$entry"
  done
  SHADOW_DIR="$WORK/shadow-bin"
  SHADOW_LOG="$WORK/shadow.log"
  mkdir -p "$SHADOW_DIR"; : >"$SHADOW_LOG"
  for r in "${SHADOWED_RUNTIMES[@]}"; do
    for name in "$r" "$r.exe"; do
      cat >"$SHADOW_DIR/$name" <<EOF
#!/usr/bin/env bash
printf '%s %s\n' "$name" "\$*" >>"$SHADOW_LOG"
echo "forbidden host runtime invoked: $name" >&2
exit 127
EOF
      chmod +x "$SHADOW_DIR/$name"
    done
  done
  export PATH="$SHADOW_DIR:$cleaned"
}

# --- files, hashes, copies ----------------------------------------------------------
random_hex() { head -c "${1:-20}" /dev/urandom | od -An -tx1 | tr -d ' \n'; }

sha_of() { sha256sum "$1" | awk '{print $1}'; }

# tree_state DIR: content hash and mtime of every file outside .git
# (zero-write evidence).
tree_state() {
  (cd "$1" && find . -path ./.git -prune -o -type f -printf '%p %T@ ' -exec sha256sum {} \; | LC_ALL=C sort)
}

# impl_files: the implementation as git sees it (tracked plus untracked,
# excluding ignored files such as .env), NUL-separated, relative to the repo.
impl_files() { git -C "$REPO" ls-files -z --cached --others --exclude-standard -- . ":(top,exclude).agents/skills/restruct-sdlc" ":(top,exclude).claude/skills/restruct-sdlc" | sort -z -u; }

# copy_repo DEST: a disposable copy of the current implementation.
copy_repo() {
  mkdir -p "$1"
  (cd "$REPO" && impl_files | while IFS= read -r -d '' f; do [[ -e "$f" ]] && printf '%s\0' "$f"; done | tar --null --no-recursion -cf - -T -) | tar -xf - -C "$1"
}

# impl_tree_hash DIR: one hash over every implementation path and content.
impl_tree_hash() {
  (cd "$REPO" && impl_files | while IFS= read -r -d '' f; do [[ -e "$f" ]] && printf '%s\0' "$f"; done) |
    (cd "$1" && xargs -0 sha256sum) | sha256sum | awk '{print $1}'
}

# --- environment files ---------------------------------------------------------------
env_get() { sed -n "s/^$2=//p" "$1/.env" | tail -1 | sed 's/^"\(.*\)"$/\1/'; }

env_set() { # env_set DEST KEY VALUE
  local tmp; tmp="$(mktemp "$1/.env.tmp.XXXXXX")"
  awk -v k="$2" -v v="$3" 'BEGIN{FS=OFS="="} $1==k{print k "=" v; next} {print}' "$1/.env" >"$tmp" && chmod 600 "$tmp" && mv "$tmp" "$1/.env"
}

# remember_secrets DEST: every non-empty secret in DEST/.env, for the canary scan.
remember_secrets() {
  grep -E '^[A-Z0-9_]*(PASSWORD|SECRET|TOKEN|API_KEY)[A-Z0-9_]*=.+' "$1/.env" | cut -d= -f2- >>"$SECRETS_FILE" || true
}

# configure DEST PROJECT PORTS...: .env through the documented helper, then a
# unique Compose project name and the given host ports
# (WEB CORE INTELLIGENCE|- POSTGRES REDIS MINIO_API MINIO_CONSOLE).
configure() {
  local dest="$1" project="$2"; shift 2
  local keys=(WEB_PORT CORE_PORT INTELLIGENCE_PORT POSTGRES_PORT REDIS_PORT MINIO_API_PORT MINIO_CONSOLE_PORT) ports=("$@") i
  (cd "$dest" && infra/local/configure-env.sh) >"$WORK/configure-$project.log" 2>&1 || return 1
  env_set "$dest" COMPOSE_PROJECT_NAME "$project"
  for i in "${!keys[@]}"; do
    if [[ -n "${ports[$i]:-}" && "${ports[$i]}" != "-" ]]; then env_set "$dest" "${keys[$i]}" "${ports[$i]}"; fi
  done
  register_project "$project"
  remember_secrets "$dest"
}

# capture_after DEST ARGS...: owned_capture after Compose commands that create or start resources, and
# build records after an explicit Compose build (build or --build).
capture_after() {
  local dest="$1" a cap=0 built=0; shift
  [[ -n "$OWNED_REG" ]] || return 0
  for a in "$@"; do
    case "$a" in up|run|start|create|build|restart) cap=1 ;; esac
    case "$a" in build|--build) built=1 ;; esac
  done
  (( cap )) || return 0
  owned_capture "$dest"
  if (( built )); then record_project_images "$dest"; fi
  return 0
}

compose() { # compose DEST args...  (the documented invocation)
  local dest="$1" rc=0; shift
  compose_call "$dest" infra/compose/local.yml -- "$@" || rc=$?
  capture_after "$dest" "$@"
  return "$rc"
}

compose_migrate() { # compose_migrate DEST args...  (local.yml plus the migrate overlay)
  local dest="$1" rc=0; shift
  compose_call "$dest" infra/compose/local.yml infra/compose/migrate.yml -- "$@" || rc=$?
  capture_after "$dest" "$@"
  return "$rc"
}

# compose_call DEST FILE... -- ARGS...: `docker compose --env-file .env -f FILE... ARGS...` from DEST (G051). A call
# that can build (up, run or create without --no-build; build) is accounted in the build ledger, its stdout and
# stderr captured separately and returned unchanged; every other call (and every call before ledger_init) runs
# exactly as before, with no extra Docker invocation.
# G064 D2: up|run|create|start|restart are latch-checked before (refused: 125, nothing run) and recounted after under the
# scope locks (a crossing latches; the call's own status is returned unchanged). Teardown subcommands are never gated.
compose_call() {
  local dest="$1" files=() fa=() rc=0 sub=""; shift
  while (($#)) && [[ "$1" != -- ]]; do files+=("$1"); fa+=(-f "$1"); shift; done
  (($#)) && shift
  res_compose_begin "$@" || return 125
  sub="$RES_CSUB"
  if [[ -z "$LEDGER_DIR" ]] || ! _lg_buildable "$@"; then
    (cd "$dest" && docker compose --env-file .env "${fa[@]}" "$@") || rc=$?
  else
    _lg_compose "$dest" "${#files[@]}" "${files[@]}" "$@" || rc=$?
  fi
  RES_CSUB="$sub" res_compose_end
  return "$rc"
}
_res_compose_allocates() { _lg_parse_compose "$@"; case "$_LC_SUB" in up|run|create|start|restart) return 0 ;; esac; return 1; }
# res_compose_begin ARGS... / res_compose_end: the gate and recount of compose_call for Compose calls made elsewhere (their
# own env file). begin: RES_CSUB = the allocating subcommand (empty otherwise); 1 = refused (nothing may run).
res_compose_begin() {
  RES_CSUB=""
  _res_compose_allocates "$@" || return 0
  RES_CSUB="$_LC_SUB"
  res_gate "compose $RES_CSUB" && return 0
  printf 'G064 compose %s refused (ran nothing): %s\n' "$RES_CSUB" "$RES_WHY" >&2; return 1
}
res_compose_end() { [[ -z "$RES_CSUB" ]] || res_recount "compose $RES_CSUB" || true; }
# _res_block_allocates CMDS: a line allocates (Compose up|run|create|start|restart; docker run|create|start|restart, network
# or volume create) or cannot be parsed (as _lg_block_buildable). Pure: no Docker invocation.
_res_block_allocates() {
  local line t=()
  while IFS= read -r line; do
    line="${line#"${line%%[![:space:]]*}"}"
    [[ -n "$line" && "$line" != \#* && "$line" == *docker* ]] || continue
    [[ ! "$line" =~ $LG_SHELL_RE ]] || return 0
    read -r -a t <<<"$line"
    [[ "${t[0]}" == docker ]] || return 0
    case "${t[1]:-}" in
      compose) _res_compose_allocates "${t[@]:2}" && return 0 ;;
      run|create|start|restart) return 0 ;;
      container) case "${t[2]:-}" in run|create|start|restart) return 0 ;; esac ;;
      network|volume) [[ "${t[2]:-}" == create ]] && return 0 ;;
    esac
  done <<<"$1"
  return 1
}

register_project() { printf '%s\n' "$1" >>"$WORK/projects"; registry_row project "$1"; }

# --- owned-resource registry (run.sh --owned-registry; G003 Plan §11.4) ------------------------
# Rows are tab-separated and append-only. Without --owned-registry every helper is a no-op except
# docker_run_owned, which then runs the container exactly as a plain `docker run`.
OWNED_REG="${OWNED_REG:-}"
registry_row() { # registry_row KIND VALUE...
  [[ -n "$OWNED_REG" ]] || return 0
  local IFS=$'\t'; printf '%s\n' "$*" >>"$OWNED_REG"
}

# docker_inventory_ids: full IDs present at the start (capture rows never name these).
docker_inventory_ids() {
  { docker ps -aq --no-trunc | sed 's/^/container /'; docker network ls -q --no-trunc | sed 's/^/network /'
    docker volume ls -q | sed 's/^/volume /'; docker image ls -q --no-trunc | sed 's/^/image /'; } | LC_ALL=C sort -u
}
id_at_start() { grep -qxF "$1 $2" "$WORK/inventory-start.ids" 2>/dev/null; }

# owned_capture DEST: positive ownership evidence for DEST's Compose project, written while it is
# provable: every container of the project whose working directory lies under this repository or
# under $WORK and that was absent at the start, plus the networks it is attached to, the named
# volumes it mounts and its image (tag and ID).
owned_capture() {
  [[ -n "$OWNED_REG" ]] || return 0
  local dest="$1" p cid wd x img imgid
  p="$(sed -n 's/^COMPOSE_PROJECT_NAME=//p' "$dest/.env" 2>/dev/null | tail -1)"
  [[ -n "$p" ]] || return 0
  for cid in $(docker ps -aq --no-trunc --filter "label=com.docker.compose.project=$p" 2>/dev/null); do
    id_at_start container "$cid" && continue
    wd="$(docker inspect -f '{{index .Config.Labels "com.docker.compose.project.working_dir"}}' "$cid" 2>/dev/null)" || continue
    [[ "$wd" == "$REPO" || "$wd" == "$REPO"/* || "$wd" == "$WORK"/* ]] || continue
    registry_row container "$cid"
    for x in $(docker inspect -f '{{range $k, $v := .NetworkSettings.Networks}}{{$v.NetworkID}} {{end}}' "$cid" 2>/dev/null); do
      id_at_start network "$x" || registry_row network "$x"
    done
    for x in $(docker inspect -f '{{range .Mounts}}{{if eq .Type "volume"}}{{.Name}} {{end}}{{end}}' "$cid" 2>/dev/null); do
      id_at_start volume "$x" || registry_row volume "$x"
    done
    img="$(docker inspect -f '{{.Config.Image}}' "$cid" 2>/dev/null)" || continue
    imgid="$(docker inspect -f '{{.Image}}' "$cid" 2>/dev/null)" || continue
    [[ -n "$img" && -n "$imgid" ]] && ! id_at_start image "$imgid" && registry_row image "$img" "$imgid"
  done
  res_project_networks "$p"
  return 0
}

# res_project_networks PROJECT [remove] (G064 accounting reconciliation; corr1 C1 exact accounting): `compose run` creates
# PROJECT's network, and `run --rm` leaves it with no container through which owned_capture could attach it, so it would
# end as an `unaccounted` labelled object (or hide behind a scope baseline). Called right after the creating call, for a
# project this run registered (its name carries RUN_ID): every network whose com.docker.compose.project label is exactly
# PROJECT, absent from the start inventory and without a network row, gets a network row by full ID (positive capture,
# never by name). remove (the project is finished): each of them is removed and verified gone ("No such network" within
# RES_RM_S); an unverified removal latches the run (C3 unverified teardown). A failed query latches the run (never zero).
# Always returns 0: its callers run it as a plain statement under set -e, and a failure is the latch, never an abort.
res_project_networks() {
  [[ -n "$OWNED_REG" ]] || return 0
  local p="$1" ids id i out
  [[ -n "$p" ]] && grep -qxF -- "$p" "$WORK/projects" 2>/dev/null || return 0
  ids="$(docker network ls -q --no-trunc --filter "label=com.docker.compose.project=$p" 2>&1)" || {
    res_latch_set run "accounting: network query of project $p failed: $(head -1 <<<"$ids")"; return 0; }
  for id in $ids; do
    if [[ ! "$id" =~ ^[0-9a-f]{64}$ ]]; then res_latch_set run "accounting: project $p network query returned a malformed ID"; return 0; fi
    id_at_start network "$id" && continue
    awk -F'\t' -v i="$id" '$1 == "network" && $2 == i { f = 1 } END { exit !f }' "$OWNED_REG" 2>/dev/null || registry_row network "$id"
    [[ "${2:-}" == remove ]] || continue
    docker network rm "$id" >/dev/null 2>&1 || true
    for ((i = 0; i <= RES_RM_S * 2; i++)); do
      out="$(docker network inspect -f '{{.Id}}' "$id" 2>&1)" || { grep -qi 'no such\|not found' <<<"$out" && continue 2; }
      sleep 0.5
    done
    res_latch_set run "unverified teardown: network ${id:0:12} of project $p not verified removed"
  done
  return 0
}

# record_built_image TAG: the run's build record for an image it just built with -t.
record_built_image() {
  local id
  printf '%s\n' "$1" >>"$WORK/images-built"
  [[ -n "$OWNED_REG" ]] || return 0
  id="$(docker image inspect -f '{{.Id}}' "$1" 2>/dev/null)" || return 0
  id_at_start image "$id" || registry_row image "$1" "$id"
}

# record_project_images DEST: build records for exactly the services with a build section in DEST's
# Compose files (tag = their image:, else Compose's <project>-<service>:latest), read right after an
# explicit build. No name pattern is used, and an ID present at the start is never recorded.
record_project_images() {
  [[ -n "$OWNED_REG" ]] || return 0
  local project f svc tag id
  project="$(sed -n 's/^COMPOSE_PROJECT_NAME=//p' "$1/.env" 2>/dev/null | tail -1)"
  [[ -n "$project" ]] || return 0
  for f in "$1"/infra/compose/*.yml; do
    while IFS=$'\t' read -r svc tag; do
      [[ -n "$svc" ]] || continue
      [[ -n "$tag" ]] || tag="$project-$svc:latest"
      id="$(docker image inspect -f '{{.Id}}' "$tag" 2>/dev/null)" || continue
      id_at_start image "$id" || registry_row image "$tag" "$id"
    done < <(yq_stdin -p yaml -o tsv '.services // {} | to_entries | .[] | select(.value.build) | [.key, (.value.image // "")]' <"$f" 2>/dev/null || true)
  done
}

# --- G064 resource latch, reservations and enforced ceilings (Plan A corr1 §C1 P1, §C2; corr2 §D2) ----------------------
# Callers run inside $(...) and nested shells, where a variable would not persist, so the hold is files under RESD (the
# registry directory; $WORK/res without a registry, which makes no accounting claim):
#   hold.d/<scope|run>  latch, created once with noclobber (O_EXCL); a later reason appends a line. Existence = held; an
#                       unreadable hold.d is held. Only the main shell moves hold.d/<scope> to held.d/ (res_scope_end),
#                       after a verified zero foreground container count (a nested scope: zero beyond the enclosing
#                       scopes' containers recorded when it began); otherwise it writes hold.d/run (never lifted).
#   rsv.d/<scope>/<pid>-<hex>  reservation (owner PID, start ticks, cidfile, counts), written under the scope locks.
#   res-<scope>.lock    flock of every scope on the stack, outer first (nested scope locks); the kernel drops it on death.
#   base-<scope>.ids    live foreground IDs when the outermost scope began (not counted against it); for a nested scope
#                       the containers live when it began (its latched end counts beyond them).
# Scopes: RES_SCOPES, the stack (outermost first) set by the main shell and inherited by every child; "run" when empty.
# Ceilings (containers/networks/volumes, C2/D2) are finite; RES_CEILING_LOWER="<scope>:<n> <scope>:<class>:<n>" may only
# lower one (proofs); an override that would raise one is refused (latched).
# G064 corr7: S-OFF is 8, not the A5 note's 7 (5 + browser + transient): `docker ps -a` counts the exited one-shot
# minio-init, so the off stack is 6, and browser_run's `run --build` capture (record_project_images' yq) is admitted while
# the retained browser one-off still exists: 6 + 1 + 1 = 8. S-REN (8) builds first and runs without --build (7 + 1).
RES_SCOPES="${RES_SCOPES:-}" RES_WHY="" RES_RSV="" RES_LOCK_S="${RES_LOCK_S:-30}" RES_RM_S="${RES_RM_S:-5}"
declare -gA RES_CEIL=([B-MIG]=5/1/2 [PROV]=9/1/7 [S-REC]=10/2/10 [S-ROOT]=11/3/3 [S-OPS]=8/3/3 [S-IDN]=15/3/4 \
  [S-REN]=8/3/3 [S-OFF]=8/2/3 [instrumentation]=1/0/0 [run]=63/5/12)
_res_d() {
  if [[ -n "$OWNED_REG" ]]; then RESD="${OWNED_REG%/*}"; else RESD="$WORK/res"; fi
  [[ -d "$RESD/rsv.d" ]] || (umask 077; mkdir -p "$RESD/hold.d" "$RESD/held.d" "$RESD/rsv.d" "$RESD/cid") 2>/dev/null
  RES_IN="${RES_SCOPES##* }"; RES_IN="${RES_IN:-run}"; RES_OUT="${RES_SCOPES%% *}"
}
_res_name_scope() { # NAME -> scope: the B-MIG batch "mig", a scope name, else the innermost scope
  case "$1" in mig) RES_NS=B-MIG ;; *) if [[ -n "$1" && -n "${RES_CEIL[$1]:-}" ]]; then RES_NS="$1"; else RES_NS="$RES_IN"; fi ;; esac
}
# _res_ceiling SCOPE CLASS -> RES_MAX (a scope without its own C2 row takes the run-wide row).
_res_ceiling() {
  local c="${RES_CEIL[$1]:-${RES_CEIL[run]}}" i=0 o k s a b n
  case "$2" in network) i=1 ;; volume) i=2 ;; esac
  IFS=/ read -r -a o <<<"$c"; RES_MAX="${o[$i]}"
  for k in ${RES_CEILING_LOWER:-}; do
    IFS=: read -r s a b <<<"$k"
    [[ "$s" == "$1" ]] || continue
    if [[ -n "$b" ]]; then [[ "$a" == "$2" ]] || continue; n="$b"; else [[ "$2" == container ]] || continue; n="$a"; fi
    if [[ ! "$n" =~ ^[0-9]+$ ]] || (( n > RES_MAX )); then RES_WHY="ceiling override refused ($k raises or is invalid)"; return 1; fi
    RES_MAX="$n"
  done
}
# res_latch_held [NAME]: 0 (held, RES_WHY) when hold.d is unreadable or hold.d/run, any scope on the stack or NAME's
# scope is latched. Reads files only.
res_latch_held() {
  _res_d; local s
  if [[ ! -r "$RESD/hold.d" || ! -x "$RESD/hold.d" ]]; then RES_WHY="held: hold.d unreadable"; return 0; fi
  _res_name_scope "${1:-}"
  for s in run $RES_SCOPES $RES_NS; do
    [[ -e "$RESD/hold.d/$s" ]] || continue
    RES_WHY="held: $s ($(sed -n 's/^reason\t//p' "$RESD/hold.d/$s" 2>/dev/null | head -1))"; return 0
  done
  return 1
}
# res_latch_set NAME WHY: latch NAME's scope ("run" for the run). The creator records one RESOURCES FAIL row (non-required;
# any FAIL row fails the run); a later reason appends one line. Never lifted here.
res_latch_set() {
  _res_d; _res_name_scope "$1"
  local f="$RESD/hold.d/$RES_NS"
  if (umask 077; set -C; printf 'scope\t%s\nreason\t%s\npid\t%s\nts\t%s\n' "$RES_NS" "$2" "$BASHPID" "$(date -u +%FT%TZ)" >"$f") 2>/dev/null; then
    if [[ -n "${RESULTS:-}" ]]; then record RESOURCES FAIL "$RES_NS: $2" >&2; fi
  else
    printf 'reason\t%s\n' "$2" >>"$f" 2>/dev/null || true
    # D1: the latch may come from a child that records no row (the witness parser); the first later reason with RESULTS
    # records it (res_report also records every latch still without a row).
    if [[ -n "${RESULTS:-}" ]] && ! grep -qF -- $'RESOURCES\tFAIL\t'"$RES_NS: " "$RESULTS" 2>/dev/null; then record RESOURCES FAIL "$RES_NS: $2" >&2; fi
  fi
  return 0
}
# _res_live CLASS -> RES_IDS: live foreground IDs of CLASS (label g001.run=$RUN_ID, a registered Compose project, or a
# registry row of that kind), instrumentation IDs (exclude.ids) omitted; one Docker query. A failed query fails.
# Also RES_FGN (their count) and RES_ALLN, the run-wide simultaneous sample (C1 observation points): the union of the
# foreground, every object carrying an sdlc.ci.run label and the instrumentation (W), from the same query (G064 corr8:
# networks and volumes too, no longer their foreground count only).
_res_live() {
  local out both
  case "$1" in
    container) out="$(docker ps -a --no-trunc --format '{{.ID}}\t{{.Label "com.docker.compose.project"}}\t{{.Label "g001.run"}}\t{{.Label "sdlc.ci.run"}}\t{{.Label "g064.bg"}}' 2>&1)" ;;
    network) out="$(docker network ls --no-trunc --format '{{.ID}}\t{{.Label "com.docker.compose.project"}}\t{{.Label "g001.run"}}\t{{.Label "sdlc.ci.run"}}' 2>&1)" ;;
    *) out="$(docker volume ls --format '{{.Name}}\t{{.Label "com.docker.compose.project"}}\t{{.Label "g001.run"}}\t{{.Label "sdlc.ci.run"}}' 2>&1)" ;;
  esac || { RES_WHY="accounting: $1 query failed: $(head -1 <<<"$out")"; return 1; }
  both="$( { sed 's/^/P\t/' "$WORK/projects" 2>/dev/null
    # G064 corr3 (Duo): a mirrored CI-control container (`standalone <id> ci-<control>`, validate-controls.sh gate_record)
    # runs in the background (run.sh starts ci-controls before the foreground scopes): it stays registry-owned (cleanup,
    # recovery) and in the run-wide sample (B), but it is not a foreground object of any scope, ceiling or zero check.
    # G064 corr8: so are that child's networks and volumes. gate_record mirrors them as bare `network|volume <id>` rows,
    # which alone made them foreground (RC17: a control network alive during B-MIG's `compose up` counted 2/1 and latched
    # it). Every control object carries the sdlc.ci.run label (ctl_labels) and no harness network or volume does, so a
    # network or volume carrying it is background (g) over any row, run label or project, and stays in the sample (A).
    # G064 corr4: so is a container a supervised background child of this run created itself (`standalone <id>
    # bg-<child>-<purpose>`, _dro): it carries the run label, so the row decides, never the run label.
    # G064 corr5: that row is appended only after an attached `docker run` returns, so while the child's container is
    # alive it has no row yet; _dro also gives it the creation-time label g064.bg=<child> (atomic with the create), and
    # a run-labelled container carrying it is background from its first listing. A row other than bg-*/ci-* (a dead
    # owner's `reconciled` leftover) overrides the label: it is foreground again and still latches.
    # G064 corr7-B: without a registry (the VC/VB control drivers) the operand must be /dev/null, never "": gawk skips an
    # empty operand and reads STDIN, which here is the caller's pipe (yq_stdin's input in _lg_config: the compose config
    # was swallowed and yq got 0 bytes). The group reads no stdin at all.
    awk -F'\t' -v k="$1" '$1 == "standalone" && k == "container" && $3 ~ /^(ci|bg)-/ { print "B\t" $2; next }
      $1 == k || ($1 == "standalone" && k == "container") { print "R\t" $2 }' "${OWNED_REG:-/dev/null}" 2>/dev/null
    sed 's/^/X\t/' "$RESD/exclude.ids" 2>/dev/null; printf -- '---\n%s\n' "$out"; } </dev/null |
    awk -F'\t' -v run="$RUN_ID" -v k="$1" 'm == 0 { if ($0 == "---") { m = 1; next } if ($1 == "P") p[$2] = 1; else if ($1 == "R") r[$2] = 1; else if ($1 == "B") b[$2] = 1; else if ($1 == "X") x[$2] = 1; next }
      $1 == "" { next } { g = ($1 in b) || (run != "" && $3 == run && $5 != "" && !($1 in r)) || (k != "container" && $4 != "")
        f = ((run != "" && $3 == run) || ($2 != "" && ($2 in p)) || ($1 in r)) && !g }
      f && !($1 in x) { print "F\t" $1 } f || $4 != "" || ($1 in x) || g { print "A\t" $1 }')"
  RES_IDS="$(sed -n 's/^F\t//p' <<<"$both" | LC_ALL=C sort -u)"
  RES_FGN="$(grep -c . <<<"$RES_IDS" || true)"
  RES_ALLN="$(sed -n 's/^A\t//p' <<<"$both" | LC_ALL=C sort -u | grep -c . || true)"
}
# _res_sample CLASS -> RES_IDS minus the outermost scope's baseline, RES_LIVE_N (their count), RES_ALLN. 1: query failed.
_res_sample() {
  _res_live "$1" || return 1
  if [[ -n "$RES_OUT" && -s "$RESD/base-$RES_OUT.ids" ]]; then
    RES_IDS="$(awk -F'\t' -v k="$1" 'NR == FNR { if ($1 == k) b[$2] = 1; next } !($1 in b)' "$RESD/base-$RES_OUT.ids" - <<<"$RES_IDS")"
  fi
  RES_LIVE_N="$(grep -c . <<<"$RES_IDS" || true)"
}
# _res_obs POINT CLASS: one sampled observation row (C1: M, a lower bound of the true peak) for the innermost scope
# ("unscoped" outside any scope) plus the run-wide row; only with a registry. Single short appends (O_APPEND).
_res_obs() {
  [[ -n "$OWNED_REG" ]] || return 0
  local sc="$RES_IN" ts; [[ -n "$RES_SCOPES" ]] || sc=unscoped
  ts="$(date -u +%FT%TZ)"
  printf 'obs\t%s\t%s\t%s\t%s\t%s\n' "$ts" "$sc" "$2" "$RES_LIVE_N" "$1" >>"$WORK/res-obs.tsv"
  printf 'obs\t%s\trun\t%s\t%s\t%s\n' "$ts" "$2" "$RES_ALLN" "$1" >>"$WORK/res-obs.tsv"
}
# res_observe POINT CLASS...: an observation point outside the scope locks (allocation-helper returns, scope boundaries).
# A failed query is an accounting failure (latched), never zero.
res_observe() {
  [[ -n "$OWNED_REG" ]] || return 0
  _res_d
  local p="$1" cls; shift
  for cls in "$@"; do
    if _res_sample "$cls"; then _res_obs "$p" "$cls"; else res_latch_set "$RES_IN" "$RES_WHY (observation: $p)"; fi
  done
  return 0
}
_res_kv() { sed -n "s/^$2\t//p" "$1" 2>/dev/null | head -1; } # FILE KEY
# _res_count SCOPE CLASS -> RES_N: live foreground objects of CLASS not in the outermost scope's baseline, plus live-owner
# reservations (their cidfile ID once written, deduplicated; otherwise the reserved count). A dead owner's reservation is
# reconciled here (under the caller's lock): dropped, logged to rsv-reconciled.tsv, and a still-existing container it
# created is counted live and given a registry row, never counted free or left unowned.
_res_count() {
  local scope="$1" cls="$2" f pid st cf id n out
  _res_sample "$cls" || return 1
  local extra=0
  for f in "$RESD/rsv.d/$scope"/*; do
    [[ -f "$f" ]] || continue
    pid="$(_res_kv "$f" pid)" st="$(_res_kv "$f" start)" cf="$(_res_kv "$f" cidfile)" id=""
    [[ -s "$cf" ]] && id="$(head -c 64 "$cf")"
    if ! _bg_live "$pid" "$st"; then
      out=gone
      if [[ -z "$id" ]]; then out=no-id
      elif docker inspect -f '{{.Id}}' "$id" >/dev/null 2>&1; then
        out=live
        [[ -z "$OWNED_REG" ]] || grep -qF "$id" "$OWNED_REG" || { registry_row standalone "$id" reconciled; out='live-registered'; }
        [[ "$cls" == container ]] && ! grep -qxF "$id" <<<"$RES_IDS" && extra=$((extra + 1))
      fi
      printf 'reconciled\t%s\t%s\towner=%s\tid=%s\toutcome=%s\n' "$(date -u +%FT%TZ)" "${f##*/}" "$pid" "${id:-none}" "$out" >>"$RESD/rsv-reconciled.tsv"
      log "G064-RES reservation ${f##*/} of dead owner $pid reconciled: id=${id:0:12} outcome=$out" >&2
      rm -f -- "$f"; continue
    fi
    n="$(_res_kv "$f" "$cls")"; [[ "$n" =~ ^[0-9]+$ ]] || n=0
    if [[ "$cls" == container && -n "$id" ]]; then grep -qxF "$id" <<<"$RES_IDS" || extra=$((extra + 1))
    elif [[ "$cls" != container || -z "$id" ]]; then extra=$((extra + n)); fi
  done
  RES_N=$(( $(grep -c . <<<"$RES_IDS" || true) + extra ))
  return 0
}
# _res_locked CMD...: CMD in a subshell holding every scope lock of the stack, outer first. A lock not obtained within
# RES_LOCK_S refuses and latches that scope.
_res_locked() {
  ( local s fd
    for s in ${RES_SCOPES:-run}; do
      exec {fd}>>"$RESD/res-$s.lock" && flock -w "$RES_LOCK_S" "$fd" && continue
      res_latch_set "$s" "accounting: scope lock not obtained in ${RES_LOCK_S}s"; printf 'held: %s lock timeout' "$s"; exit 1
    done
    "$@" )
}
_res_refused() { printf 'refused\t%s\t%s\n' "$(date -u +%FT%TZ)" "$1" >>"$RESD/held.log" 2>/dev/null || true; } # WHAT
# _res_admit WHAT CIDFILE NVOL FILE OWNER (under the locks): latch check; count + reservation against the innermost scope's
# ceiling for the container and NVOL volumes; a query failure or a crossing refuses and latches; else the reservation.
_res_admit() {
  local cls need st
  if res_latch_held; then _res_refused "$1"; printf '%s' "$RES_WHY"; return 1; fi
  for cls in container volume; do
    need=1; [[ "$cls" == volume ]] && need="$3"
    (( need > 0 )) || continue
    if ! _res_ceiling "$RES_IN" "$cls"; then res_latch_set run "$RES_WHY"; printf 'held: %s' "$RES_WHY"; return 1; fi
    if ! _res_count "$RES_IN" "$cls"; then res_latch_set "$RES_IN" "$RES_WHY"; printf 'held: %s' "$RES_WHY"; return 1; fi
    if (( RES_N + need > RES_MAX )); then
      res_latch_set "$RES_IN" "ceiling $cls $((RES_N + need))/$RES_MAX ($1)"; _res_refused "$1"
      printf 'held: resource ceiling %s %s/%s for %s' "$cls" "$((RES_N + need))" "$RES_MAX" "$RES_IN"; return 1
    fi
  done
  _bg_stat "$5" || { printf 'reservation owner %s not found' "$5"; return 1; }
  st="$_BG_ST"
  if ! (umask 077; mkdir -p "${4%/*}" && set -C && printf 'pid\t%s\nstart\t%s\ncidfile\t%s\ncontainer\t1\nvolume\t%s\nwhat\t%s\n' "$5" "$st" "$2" "$3" "$1" >"$4") 2>/dev/null; then
    printf 'reservation %s not written' "${4##*/}"; return 1
  fi
}
# res_reserve WHAT CIDFILE NVOL: serialized check-and-reserve for one container (+NVOL anonymous volumes) in the innermost
# scope. 0: RES_RSV = the reservation (the caller drops it after the registry row). 1: refused, RES_WHY.
res_reserve() {
  _res_d
  # The owner is this shell (it creates and drops the reservation), never the $(...) that takes the locks.
  local me="$BASHPID" f out rc=0
  f="$RESD/rsv.d/$RES_IN/$me-$(random_hex 4)"
  RES_RSV=""
  out="$(_res_locked _res_admit "$1" "$2" "${3:-0}" "$f" "$me")" || rc=$?
  if (( rc )); then RES_WHY="${out:-held: reservation refused}"; return 1; fi
  RES_RSV="$f"
}
# res_gate WHAT: latch check before a creator that cannot reserve (Compose, documented blocks, product code). 1 = refused.
res_gate() {
  res_latch_held || return 0
  _res_refused "$1"; return 1
}
# _res_recount_locked WHAT / res_recount WHAT: after such a creator, every class recounted under the scope locks against
# the innermost scope's ceiling; a crossing or a failed query latches (C2: the caller's status is unchanged).
_res_recount_locked() {
  local cls bad=""
  for cls in container network volume; do
    if ! _res_ceiling "$RES_IN" "$cls"; then res_latch_set run "$RES_WHY"; bad+="$RES_WHY; "; continue; fi
    if ! _res_count "$RES_IN" "$cls"; then res_latch_set "$RES_IN" "$RES_WHY"; bad+="$RES_WHY; "; continue; fi
    _res_obs "$1" "$cls"
    (( RES_N <= RES_MAX )) || { res_latch_set "$RES_IN" "ceiling $cls $RES_N/$RES_MAX ($1)"; bad+="ceiling $cls $RES_N/$RES_MAX; "; }
  done
  [[ -z "$bad" ]] || { printf '%s' "${bad%; }"; return 1; }
}
res_recount() {
  _res_d
  local out
  # D1: a witnessed creator (Compose call, documented block, recovery tool) gets its marker first, so its attributed rows
  # and verified anonymous-volume removals precede the recount; a Go invocation boundary is an observation point only.
  [[ "$1" == go_it* ]] || res_witness_mark "$1"
  out="$(_res_locked _res_recount_locked "$1")" && return 0
  RES_WHY="$out"; log "G064-RES $RES_IN after $1: $out" >&2; return 1
}
# res_scope_begin SCOPE / res_scope_end SCOPE (main shell only). The outermost scope records its baseline. At the end of the
# outermost scope, or of a latched scope, the foreground container count must be a verified zero: then a latch of that
# scope moves to held.d/; otherwise hold.d/run is written (every later allocation refused). Networks and volumes left
# beyond the baseline at an outermost end are reported (log) and stay with the run's final cleanup.
res_scope_begin() {
  _res_d
  local cls
  res_witness_mark "scope begin $1"
  if [[ -z "$RES_SCOPES" ]]; then
    : >"$RESD/base-$1.ids"
    for cls in container network volume; do
      _res_live "$cls" || { res_latch_set "$1" "$RES_WHY (scope begin)"; break; }
      [[ -z "$RES_IDS" ]] || sed "s/^/$cls\t/" <<<"$RES_IDS" >>"$RESD/base-$1.ids"
    done
  else
    # A nested scope (PROV, S-REC, S-IDN; the driver's B-MIG under S-F1STACK) begins while the enclosing scope's
    # containers live (the migrations stack): they are its baseline, and its latched end (res_scope_end) requires a
    # verified zero beyond them. A failed query records no baseline, which keeps that end unverified.
    rm -f -- "$RESD/base-$1.ids"
    if _res_live container; then
      { [[ -z "$RES_IDS" ]] || sed 's/^/container\t/' <<<"$RES_IDS"; } >"$RESD/base-$1.ids"
    else res_latch_set "$1" "$RES_WHY (scope begin)"; fi
  fi
  RES_SCOPES="${RES_SCOPES:+$RES_SCOPES }$1"
  _res_scope_row begin "$1"
  res_observe "scope begin $1" container network volume
  if res_latch_held "$1"; then log "G064-RES scope $1 begins held: $RES_WHY"; fi
  return 0
}
# _res_scope_row begin|end SCOPE: the scope boundary as a registry line offset in $WORK/res-scopes.tsv (C1).
_res_scope_row() {
  [[ -n "$OWNED_REG" ]] || return 0
  local n; n="$(wc -l <"$OWNED_REG" 2>/dev/null)" || n=""
  printf '%s\t%s\t%s\t%s\n' "$1" "$2" "${n:-?}" "$(date -u +%FT%TZ)" >>"$WORK/res-scopes.tsv"
}
res_scope_end() {
  _res_d
  local s="$1" rest n cls left="" beyond="" base
  if [[ "$RES_IN" != "$s" ]]; then res_latch_set run "scope end $s is not the innermost scope ($RES_IN)"; return 1; fi
  res_witness_mark "scope end $s"
  res_observe "scope end $s" container network volume
  _res_scope_row end "$s"
  rest="${RES_SCOPES% *}"; [[ "$rest" == "$RES_SCOPES" ]] && rest=""
  if [[ -z "$rest" || -e "$RESD/hold.d/$s" ]]; then
    base="$RESD/base-$s.ids"
    if ! _res_live container; then n="unverified (${RES_WHY})"
    elif [[ -z "$rest" ]]; then n="$(grep -c . <<<"$RES_IDS" || true)"
    elif [[ ! -f "$base" || ! -r "$base" ]]; then n="unverified (no baseline recorded for $s)"
    else
      # a latched nested scope: the containers live when it began belong to the enclosing scopes (its baseline)
      [[ ! -s "$base" ]] || RES_IDS="$(awk -F'\t' 'NR == FNR { if ($1 == "container") b[$2] = 1; next } !($1 in b)' "$base" - <<<"$RES_IDS")"
      n="$(grep -c . <<<"$RES_IDS" || true)"; beyond=" beyond its baseline"
    fi
    if [[ "$n" == 0 ]]; then
      [[ ! -e "$RESD/hold.d/$s" ]] || mv -- "$RESD/hold.d/$s" "$RESD/held.d/$s.$(date -u +%Y%m%dT%H%M%S).$BASHPID"
    else
      res_latch_set run "scope $s ended with foreground containers $n$beyond (verified zero required): $(tr '\n' ' ' <<<"${RES_IDS:-}" | cut -c1-200)"
    fi
    if [[ -z "$rest" ]]; then
      for cls in network volume; do
        _res_count "$s" "$cls" 2>/dev/null && (( RES_N > 0 )) && left+="$cls=$RES_N "
      done
      [[ -z "$left" ]] || log "G064-RES scope $s ended with ${left}beyond its baseline (reported; the run's final cleanup removes them)"
    fi
  fi
  RES_SCOPES="$rest"
  return 0
}

# docker_run_owned PURPOSE ARGS... / docker_create_owned PURPOSE ARGS...: every harness `docker run`/`create` (C1 P1).
# The Docker CLI writes the full ID into a cidfile at creation (inside the run's mode-700 registry; $WORK/res/cid without
# one) and a standalone row follows (G003 r4 §11.4). Capacity is reserved first (res_reserve; refused: rc 125, nothing
# created). Detached and create: mounts are inspected at once and anonymous volumes become volume rows. Attached --rm keeps
# --rm unless the image declares Config.Volumes (one cached image inspect per image; unknown counts as declaring) or the
# arguments hold an anonymous mount; then create -> mounts -> start -a [-i] -> rm -v -> verified gone, returning the
# container's status. DRO_TIMEOUT=N bounds the run (or start) with timeout(1).
docker_run_owned() { _dro run "$@"; }
docker_create_owned() { _dro create "$@"; }
_dro_parse() { # ARGS... -> DRO_DETACH DRO_RM DRO_STDIN DRO_IMAGE DRO_ANON DRO_ARGS (without --rm) DRO_PARSED
  DRO_DETACH=0 DRO_RM=0 DRO_STDIN=0 DRO_IMAGE="" DRO_ANON=0 DRO_ARGS=() DRO_PARSED=1
  local a v
  while (($#)); do
    a="$1"; shift
    case "$a" in
      --rm) DRO_RM=1; continue ;;
      -d|--detach) DRO_DETACH=1 ;;
      -i|--interactive) DRO_STDIN=1 ;;
      --init|-t|--tty|--read-only|--privileged|-q|--quiet) ;;
      -v|--volume|--mount)
        v="${1-}"; shift
        if [[ "$a" == --mount ]]; then [[ "$v" == *type=volume* && "$v" != *src=* && "$v" != *source=* ]] && DRO_ANON=$((DRO_ANON + 1))
        elif [[ "$v" != *:* ]]; then DRO_ANON=$((DRO_ANON + 1)); fi
        DRO_ARGS+=("$a" "$v"); continue ;;
      --cidfile|--name|--label|-l|--network|--net|-e|--env|--env-file|-w|--workdir|--entrypoint|-u|--user|--shm-size|--pull|\
      --tmpfs|-p|--publish|--add-host|--hostname|-h|--platform|-m|--memory|--cpus|--ulimit|--security-opt|--cap-add|--cap-drop|\
      --stop-signal|--stop-timeout|--restart|--log-driver|--log-opt|--ipc|--pid|--device|--group-add|--dns)
        DRO_ARGS+=("$a" "${1-}"); shift; continue ;;
      --*=*) ;;
      -*) DRO_PARSED=0 ;;
      *) DRO_IMAGE="$a"; DRO_ARGS+=("$a" "$@"); return 0 ;;
    esac
    DRO_ARGS+=("$a")
  done
}
_dro_imgvols() { # IMAGE -> DRO_IV: its Config.Volumes count, -1 when unknown; cached per image under RESD/imgvol
  local f
  DRO_IV=-1
  [[ -n "$1" ]] || return 0
  f="$RESD/imgvol/$(printf '%s' "$1" | sha256sum | cut -c1-32)"
  if [[ -s "$f" ]]; then DRO_IV="$(cat "$f")"; return 0; fi
  DRO_IV="$(docker image inspect -f '{{len .Config.Volumes}}' "$1" 2>/dev/null)" || DRO_IV=-1
  [[ "$DRO_IV" =~ ^[0-9]+$ ]] || { DRO_IV=-1; return 0; }
  (umask 077; mkdir -p "$RESD/imgvol" && printf '%s' "$DRO_IV" >"$f") 2>/dev/null || true
}
_dro_mounts() { # ID -> DRO_NV anonymous (volume-type) mounts, each a volume row; 1 on an inspect failure other than gone
  local out x
  DRO_NV=0
  out="$(docker inspect -f '{{range .Mounts}}{{if eq .Type "volume"}}{{.Name}} {{end}}{{end}}' "$1" 2>&1)" || {
    grep -qi 'no such' <<<"$out" && return 0; RES_WHY="accounting: mount inspect failed: $(head -1 <<<"$out")"; return 1; }
  for x in $out; do DRO_NV=$((DRO_NV + 1)); id_at_start volume "$x" || registry_row volume "$x"; done
}
_dro_remove() { # ID: rm -f -v, then verified gone within RES_RM_S; unverified latches (C3)
  local out i
  docker rm -f -v "$1" >/dev/null 2>&1 || true
  for ((i = 0; i <= RES_RM_S * 2; i++)); do
    out="$(docker inspect -f '{{.Id}}' "$1" 2>&1)" || { grep -qi 'no such' <<<"$out" && return 0; }
    sleep 0.5
  done
  res_latch_set "$RES_IN" "teardown unverified: container ${1:0:12}"; return 1
}
_dro() {
  local mode="$1" purpose="$2" nv rc=0 id split=0 sa=(-a) tmo=() bgl=(); shift 2
  # G064 corr4: a creation by a supervised background child (bg_start sets BG_CHILD; the controls suite's yq containers
  # run concurrently with the foreground scopes) is registered as that child's background object (`bg-<child>-<purpose>`,
  # _res_live B): owned, cleaned and sampled, never a foreground object of the scope open in the main shell.
  # G064 corr5: the row follows the call, and an attached run returns only after the container is gone, so the container
  # also carries the creation-time label g064.bg=<child> (_res_live reads it): background from the create itself.
  [[ -z "${BG_CHILD:-}" ]] || { purpose="bg-$BG_CHILD-$purpose"; bgl=(--label "g064.bg=$BG_CHILD"); }
  DRO_NV=0; _res_d; _dro_parse "$@"; _dro_imgvols "$DRO_IMAGE"
  (( ${#bgl[@]} == 0 )) || DRO_ARGS=("${bgl[@]}" "${DRO_ARGS[@]}")
  nv=$DRO_ANON; (( DRO_PARSED == 0 || DRO_IV < 0 )) || nv=$((nv + DRO_IV))
  [[ "$mode" == run ]] && (( DRO_RM && !DRO_DETACH )) && (( DRO_ANON > 0 || DRO_IV != 0 || DRO_PARSED == 0 )) && split=1
  # Unique without parent-shell state: callers run this inside $(...), where a counter would not persist.
  local cf; cf="$RESD/cid/$RUN_ID-$purpose-$BASHPID-$(random_hex 4).cid"
  if ! res_reserve "$mode $purpose" "$cf" "$nv"; then
    printf 'G064 %s %s refused (created nothing): %s\n' "$mode" "$purpose" "$RES_WHY" >&2; return 125
  fi
  [[ -z "${DRO_TIMEOUT:-}" ]] || tmo=(timeout "$DRO_TIMEOUT")
  if (( split )); then
    docker create --cidfile "$cf" "${DRO_ARGS[@]}" >/dev/null || rc=$?
  else
    "${tmo[@]}" docker "$mode" --cidfile "$cf" "${bgl[@]}" "$@" || rc=$?
  fi
  id="$(head -c 64 "$cf" 2>/dev/null)" || id=""
  if [[ -n "$id" ]]; then
    registry_row standalone "$id" "$purpose"
    if (( split || DRO_DETACH )) || [[ "$mode" == create ]]; then
      if ! _dro_mounts "$id"; then res_latch_set "$RES_IN" "$RES_WHY ($purpose)"; (( split )) || _dro_remove "$id" || true; rc=125
      elif (( DRO_NV > nv )) && ! res_recount "$mode $purpose"; then
        # More anonymous volumes than reserved (image unknown before the pull) and the scope crossed: latched; remove it.
        printf 'G064 %s %s crossed a ceiling: %s\n' "$mode" "$purpose" "$RES_WHY" >&2; _dro_remove "$id" || true; split=0; rc=125
      fi
    fi
    if (( split )); then
      if (( rc == 0 )); then (( DRO_STDIN )) && sa+=(-i); "${tmo[@]}" docker start "${sa[@]}" "$id" || rc=$?; fi
      _dro_remove "$id" || true
    fi
  fi
  rm -f -- "$RES_RSV"
  # C1 observation point at the helper's return (volumes too after a creation with anonymous volumes).
  if [[ -n "$id" ]]; then if (( nv > 0 || DRO_NV > 0 )); then res_observe "$mode $purpose" container volume; else res_observe "$mode $purpose" container; fi; fi
  return "$rc"
}

# --- G064 run-scoped creation witness and resources.tsv (Plan A corr1 §C1 P2-P4; corr2 §D1) ---------------------------
# With a registry, ONE witness container W (busybox, --network none; created through docker_run_owned under D2's latch and
# lock in the "instrumentation" scope, 1/0/0) and ONE `docker events` stream covering container AND volume events, with no
# daemon-side label filter, read by one parser in a supervised child (bg_start res-witness). Files under $WORK/res-witness:
#   w.id, events.pid (PID and start time), events.tsv (journal), win-<n>.tsv (window n's events), ack (windows closed,
#   written by atomic rename after the window is processed), attr.tsv, bound.tsv, gone.tsv, named.tsv, stop, dead, stopped.
# Attribution is positive and live: a container from its own create event (a registered project, working_dir under REPO
# or WORK, absent at the start), an anonymous volume only from its own mount event naming an attributed container (never
# from a container's labels). Both get a capture row in the existing registry at once (no other registry). After every
# witnessed call (res_recount) and at each scope boundary the caller takes marker.lock, runs `docker top W` (a documented
# event on W that creates nothing) and waits for a larger ack. Each failure (marker timeout, stream EOF, W lost, an
# unverified removal, unbound-volume) latches (D2): never "unknown".
RES_WD_FMT='{{.Type}}{{"\t"}}{{.Action}}{{"\t"}}{{.Actor.ID}}{{"\t"}}{{index .Actor.Attributes "com.docker.compose.project"}}{{"\t"}}{{index .Actor.Attributes "com.docker.compose.project.working_dir"}}{{"\t"}}{{index .Actor.Attributes "container"}}{{"\t"}}{{index .Actor.Attributes "destination"}}{{"\t"}}{{index .Actor.Attributes "image"}}'
RES_ACK_S="${RES_ACK_S:-10}" RES_WD_SINCE=""
# res_witness_start (main shell, before the suite loop; never inside $(...)).
res_witness_start() {
  local wd="$WORK/res-witness" w="" i cls RES_SCOPES=instrumentation
  if [[ -z "$OWNED_REG" ]]; then log "G064-RES accounting=unavailable (no registry)"; return 0; fi
  _res_d
  if ! { mkdir -p "$wd" && : >"$wd/events.tsv" && printf '0\n' >"$wd/ack"; } 2>/dev/null; then res_latch_set run "accounting: witness files not writable"; return 0; fi
  _dro_imgvols "$BUSYBOX_IMAGE"
  if [[ "$DRO_IV" != 0 ]]; then res_latch_set run "accounting: witness image $BUSYBOX_IMAGE declares volumes or cannot be inspected ($DRO_IV); not started"; return 0; fi
  RES_WD_SINCE="$(date +%s)"   # the stream replays from here, so W's own create event cannot precede its subscription
  bg_start res-witness _res_witness_child
  : >"$RESD/base-instrumentation.ids"
  for cls in container network volume; do
    if _res_live "$cls"; then [[ -z "$RES_IDS" ]] || sed "s/^/$cls\t/" <<<"$RES_IDS" >>"$RESD/base-instrumentation.ids"
    else res_latch_set run "$RES_WHY (witness start)"; return 0; fi
  done
  w="$(docker_run_owned witness -d --pull never --network none --label "g001.run=$RUN_ID" "$BUSYBOX_IMAGE" sleep 2147483647)" || w=""
  if [[ ! "$w" =~ ^[0-9a-f]{64}$ ]]; then res_latch_set run "accounting: witness container not started"; return 0; fi
  printf '%s\n' "$w" >>"$RESD/exclude.ids"
  printf '%s\n' "$w" >"$wd/w.id.tmp" && mv -f -- "$wd/w.id.tmp" "$wd/w.id"
  for ((i = 0; i < RES_ACK_S * 20; i++)); do
    if grep -qF -- $'container\tcreate\t'"$w" "$wd/events.tsv" 2>/dev/null; then log "G064-RES witness ${w:0:12} live (its own create event seen)"; return 0; fi
    [[ ! -e "$wd/dead" ]] || break
    sleep 0.05
  done
  res_latch_set run "accounting: witness stream not live (no create event of ${w:0:12} within ${RES_ACK_S}s)"; return 0
}
_res_witness_child() {
  set +e
  local wd="$WORK/res-witness" ev
  RESULTS="" RES_SCOPES="" # a latch here writes no row; res_report records every latch as a RESOURCES FAIL row
  exec {ev}< <(exec docker events --since "$RES_WD_SINCE" --format "$RES_WD_FMT" --filter type=container --filter type=volume \
    --filter event=create --filter event=destroy --filter event=mount --filter event=top 2>>"$wd/events.err")
  _bg_stat "$!" && printf '%s %s\n' "$!" "$_BG_ST" >"$wd/events.pid"
  _res_witness_parse <&"$ev"
}
# _res_imgdests IMAGE -> RES_DESTS: the image's declared Config.Volumes destinations (cached); 1 when not inspectable.
_res_imgdests() {
  local f; f="$RESD/imgdest/$(printf '%s' "$1" | sha256sum | cut -c1-32)"
  if [[ -f "$f" ]]; then RES_DESTS="$(cat "$f")"; return 0; fi
  RES_DESTS="$(docker image inspect -f '{{range $k, $v := .Config.Volumes}}{{$k}} {{end}}' "$1" 2>/dev/null)" || return 1
  (umask 077; mkdir -p "$RESD/imgdest" && printf '%s' "$RES_DESTS" >"$f") 2>/dev/null || true
}
# _res_vol_gone V: `docker volume inspect V` fails "No such volume" within RES_RM_S (P5's bound).
_res_vol_gone() {
  local i out
  for ((i = 0; i <= RES_RM_S * 2; i++)); do
    out="$(docker volume inspect -f '{{.Name}}' "$1" 2>&1)" || { grep -qi 'no such' <<<"$out" && return 0; }
    sleep 0.5
  done
  return 1
}
_res_witness_parse() {
  local wd="$WORK/res-witness" line raw w="" n=1 typ act id c p d
  local -a f
  local -A attr=() cgone=() cmnt=() vc=() vd=() vcre=() vgone=() pend=() chk=()
  _res_d
  : >"$wd/win-1.tsv"
  while IFS= read -r line; do
    raw="$line"; f=()
    while [[ "$line" == *$'\t'* ]]; do f+=("${line%%$'\t'*}"); line="${line#*$'\t'}"; done; f+=("$line")
    typ="${f[0]}" act="${f[1]:-}" id="${f[2]:-}"
    printf '%s\n' "$raw" >>"$wd/events.tsv"; printf '%s\n' "$raw" >>"$wd/win-$n.tsv"
    [[ -n "$w" || ! -s "$wd/w.id" ]] || w="$(head -c 64 "$wd/w.id")"
    case "$typ/$act" in
      container/create)
        [[ -n "$id" && "$id" != "$w" && -z "${attr[$id]:-}" ]] || continue
        p="${f[3]:-}" d="${f[4]:-}"
        [[ -n "$p" ]] && grep -qxF -- "$p" "$WORK/projects" 2>/dev/null || continue
        [[ "$d" == "$REPO" || "$d" == "$REPO"/* || "$d" == "$WORK"/* ]] || continue
        id_at_start container "$id" && continue
        attr[$id]="${f[7]:--}"
        registry_row container "$id"
        printf 'container\t%s\tproject=%s\twindow=%s\n' "$id" "$p" "$n" >>"$wd/attr.tsv" ;;
      container/destroy)
        if [[ -n "$w" && "$id" == "$w" && ! -e "$wd/stop" ]]; then res_latch_set run "accounting: witness ${w:0:12} destroyed during the run"; fi
        [[ -z "${attr[$id]:-}" ]] || cgone[$id]=1 ;;
      container/top)
        [[ -n "$w" && "$id" == "$w" ]] || continue
        _res_witness_window
        printf '%s\n' "$n" >"$wd/ack.tmp" && mv -f -- "$wd/ack.tmp" "$wd/ack"
        n=$((n + 1)); : >"$wd/win-$n.tsv" ;;
      volume/create) vcre[$id]="$n"; unset "vgone[$id]" ;; # a witnessed re-create (compose down -v, up) voids the destroy
      volume/destroy) vgone[$id]=1 ;;
      volume/mount)
        c="${f[5]:-}" d="${f[6]:-}"
        [[ -n "$c" && -n "${attr[$c]:-}" ]] || continue # bound only by its own mount event's container attribute
        cmnt[$c]+=" $d"
        [[ -z "${vc[$id]:-}" ]] || continue
        id_at_start volume "$id" && continue
        vc[$id]="$c" vd[$id]="$d" pend[$id]=1
        registry_row volume "$id"
        printf 'bound\t%s\tmount:%s:%s\n' "$id" "$c" "$d" >>"$wd/bound.tsv" ;;
    esac
  done
  [[ ! -e "$wd/stop" ]] || return 0
  : >"$wd/dead"
  res_latch_set run "accounting: witness stream ended (events EOF${w:+, W ${w:0:12}}): $(tail -1 "$wd/events.err" 2>/dev/null)"
  return 1
}
# _res_witness_window (in the parser, at W's top event): each bound volume whose container is gone is verified removed
# (its destroy seen, or `docker volume rm`), unless it is Compose-named or pre-existing; then the unbound-volume check.
_res_witness_window() {
  local v c out rc by al
  for v in "${!pend[@]}"; do
    c="${vc[$v]}"
    if [[ -z "${cgone[$c]:-}" ]]; then
      out="$(docker inspect -f '{{.Id}}' "$c" 2>&1)" && continue # still exists: a later window
      grep -qi 'no such' <<<"$out" || { res_latch_set run "accounting: witness cannot inspect ${c:0:12}: $(head -1 <<<"$out")"; continue; }
    fi
    unset "pend[$v]"
    if [[ -z "${vcre[$v]:-}" ]]; then printf 'named\t%s\tcreate not witnessed (pre-existing)\n' "$v" >>"$wd/named.tsv"; continue; fi
    by=compose al=unread
    if [[ -z "${vgone[$v]:-}" ]]; then
      # Its own labels decide: a Compose project label = named (project teardown); com.docker.volume.anonymous is recorded only.
      out="$(docker volume inspect -f '{{index .Labels "com.docker.compose.project"}}|{{index .Labels "com.docker.volume.anonymous"}}' "$v" 2>&1)" && rc=0 || rc=$?
      if (( rc == 0 )); then
        al="${out#*|}" out="${out%%|*}"; al="${al:-none}"
        if [[ -n "$out" ]]; then printf 'named\t%s\tcompose project %s (project teardown)\n' "$v" "$out" >>"$wd/named.tsv"; continue; fi
        docker volume rm "$v" >/dev/null 2>&1; by=harness
      elif ! grep -qi 'no such' <<<"$out"; then
        res_latch_set run "accounting: anonymous volume ${v:0:12} of ${c:0:12} not inspectable: $(head -1 <<<"$out")"; continue
      fi
    fi
    if _res_vol_gone "$v"; then printf 'gone\t%s\tcall-end removed_by=%s\tmount:%s:%s\tanonymous_label=%s\n' "$v" "$by" "$c" "${vd[$v]}" "$al" >>"$wd/gone.tsv"
    else res_latch_set run "accounting: anonymous volume ${v:0:12} of ${c:0:12} not verified removed"; fi
  done
  for c in "${!cgone[@]}"; do
    [[ -z "${chk[$c]:-}" ]] || continue
    chk[$c]=1
    if ! _res_imgdests "${attr[$c]}"; then res_latch_set run "accounting: image ${attr[$c]} of ${c:0:12} not inspectable (declared volumes unknown)"; continue; fi
    for v in $RES_DESTS; do
      [[ " ${cmnt[$c]:-} " == *" $v "* ]] || res_latch_set run "accounting: unbound-volume: ${c:0:12} (image ${attr[$c]}) declares $v; no mount event named it before its destroy"
    done
  done
}
# res_witness_mark WHAT: the D1 marker. Works from $(...) and nested shells (files and flock only). Always returns 0;
# a failure latches the run.
res_witness_mark() {
  local wd="$WORK/res-witness"
  [[ -s "$wd/w.id" && ! -e "$wd/stopped" ]] || return 0
  ( set +e
    w="$(head -c 64 "$wd/w.id")"
    [[ ! -e "$wd/dead" ]] || { res_latch_set run "accounting: witness stream down at marker ($1)"; exit 0; }
    flock -w "$RES_LOCK_S" 9 || { res_latch_set run "accounting: witness marker lock not obtained in ${RES_LOCK_S}s ($1)"; exit 0; }
    a0=0; { read -r a0 <"$wd/ack"; } 2>/dev/null; [[ "$a0" =~ ^[0-9]+$ ]] || a0=0
    out="$(docker top "$w" 2>&1)" || { res_latch_set run "accounting: witness lost at marker ($1): $(head -1 <<<"$out")"; exit 0; }
    for ((i = 0; i < RES_ACK_S * 20; i++)); do
      a=""; { read -r a <"$wd/ack"; } 2>/dev/null
      if [[ "$a" =~ ^[0-9]+$ ]] && (( a > a0 )); then exit 0; fi
      [[ ! -e "$wd/dead" ]] || break
      sleep 0.05
    done
    res_latch_set run "accounting: witness marker not acknowledged within ${RES_ACK_S}s ($1)"; exit 0
  ) 9>>"$wd/marker.lock" || res_latch_set run "accounting: witness marker lock file unavailable ($1)"
  return 0
}
# res_witness_stop (main shell, after the background suites are joined): a final marker, the stream ended, the parser
# joined, W removed and verified gone (teardown is never gated). A parser that does not stop latches.
res_witness_stop() {
  local wd="$WORK/res-witness" pid="" st="" i w
  [[ -n "${BG_PID[res-witness]:-}" ]] || return 0
  _res_d
  res_witness_mark "witness stop"
  : >"$wd/stop"
  { read -r pid st <"$wd/events.pid"; } 2>/dev/null || pid=""
  if [[ -n "$pid" ]] && _bg_live "$pid" "$st"; then kill -TERM "$pid" 2>/dev/null || true; fi
  for ((i = 0; i < RES_ACK_S * 20; i++)); do
    _bg_live "${BG_PID[res-witness]}" "${BG_PSTART[res-witness]}" || break
    sleep 0.05
  done
  if _bg_live "${BG_PID[res-witness]}" "${BG_PSTART[res-witness]}"; then res_latch_set run "accounting: witness parser did not stop within ${RES_ACK_S}s"
  else bg_join res-witness res-witness; fi
  if [[ -s "$wd/w.id" ]]; then w="$(head -c 64 "$wd/w.id")"; _dro_remove "$w" || true; fi
  : >"$wd/stopped"
}
# res_witness_abort (G065 K3; the EXIT trap, after bg_stop_all and before cleanup): a run that ends before res_witness_stop
# still removes its recorded W by exact ID (_dro_remove: rm -f -v, verified gone, latched when unverified) and marks the
# witness stopped. Not res_witness_stop: bg_stop_all has already stopped the parser, so a marker would time out and the
# join would find it not running. A no-op without a recorded W or once stopped. 1: W not verified gone.
res_witness_abort() {
  local wd="$WORK/res-witness" w rc=0
  [[ -s "$wd/w.id" && ! -e "$wd/stopped" ]] || return 0
  _res_d
  w="$(head -c 64 "$wd/w.id")"
  _dro_remove "$w" || rc=1
  : >"$wd/stopped"
  return "$rc"
}
# res_report: resources.tsv (C1) from the registry, $WORK/res-scopes.tsv (scope boundaries as registry line offsets),
# $WORK/res-obs.tsv and one final query per class. One row per scope and class (creations: unique IDs by first registry
# row, M exact; observed_max: M sampled; ceiling: D; final_leftover, unaccounted: M), detail rows (created, gone, pulled,
# finding). W is excluded by exact ID from the other scopes (instrumentation) and included in the run totals. Every FAIL
# row and every latch (hold.d, held.d) is a RESOURCES FAIL row. Without a registry: no accounting claim.
res_report() {
  if [[ -z "$OWNED_REG" ]]; then log "G064-RES accounting=unavailable (no registry)"; return 0; fi
  _res_d
  local wd="$WORK/res-witness" out="$WORK/resources.tsv" live="$WORK/res-live.tsv" ceil="$WORK/res-ceil.tsv" w="" q s c f
  local latched="$WORK/res-latched.tsv" fails=() line sc why
  [[ ! -s "$wd/w.id" ]] || w="$(head -c 64 "$wd/w.id")"
  : >"$live"; : >"$ceil"; : >"$latched"
  mkdir -p "$wd" 2>/dev/null; touch "$WORK/res-scopes.tsv" "$WORK/res-obs.tsv" "$WORK/projects" "$WORK/inventory-start.ids" "$wd/bound.tsv" "$wd/gone.tsv" 2>/dev/null
  if q="$(docker ps -a --no-trunc --format '{{.ID}}\t{{.Label "com.docker.compose.project"}}\t{{.Label "g001.run"}}' 2>&1)"; then sed '/^$/d; s/^/container\t/' <<<"$q" >>"$live"; else fails+=("run: accounting: final container query failed"); fi
  if q="$(docker network ls --no-trunc --format '{{.ID}}\t{{.Label "com.docker.compose.project"}}\t{{.Label "g001.run"}}' 2>&1)"; then sed '/^$/d; s/^/network\t/' <<<"$q" >>"$live"; else fails+=("run: accounting: final network query failed"); fi
  if q="$(docker volume ls --format '{{.Name}}\t{{.Label "com.docker.compose.project"}}\t{{.Label "g001.run"}}' 2>&1)"; then sed '/^$/d; s/^/volume\t/' <<<"$q" >>"$live"; else fails+=("run: accounting: final volume query failed"); fi
  if q="$(docker image ls -q --no-trunc 2>&1)"; then sed '/^$/d; s/^/image\t/' <<<"$q" | LC_ALL=C sort -u >>"$live"; else fails+=("run: accounting: final image query failed"); fi
  for s in run unscoped instrumentation $(cut -f2 "$WORK/res-scopes.tsv" 2>/dev/null | LC_ALL=C sort -u); do
    f=""; for c in container network volume; do if _res_ceiling "$s" "$c"; then f+="$RES_MAX/"; else f+="?/"; fi; done
    printf '%s\t%s\n' "$s" "${f%/}" >>"$ceil"
  done
  for f in "$RESD"/hold.d/* "$RESD"/held.d/*; do
    [[ -f "$f" ]] || continue
    printf '%s\t%s\n' "$(_res_kv "$f" scope)" "$(_res_kv "$f" reason)" >>"$latched"
  done
  { awk -F'\t' -v OFS='\t' -v W="$w" -v RUN="$RUN_ID" \
    -v CEIL="$ceil" -v SC="$WORK/res-scopes.tsv" -v REG="$OWNED_REG" -v BOUND="$wd/bound.tsv" -v GONE="$wd/gone.tsv" \
    -v OBS="$WORK/res-obs.tsv" -v PROJ="$WORK/projects" -v START="$WORK/inventory-start.ids" -v LIVE="$live" -v LAT="$latched" '
    FILENAME == CEIL { ceil[$1] = $2; known[$1] = 1; next }
    FILENAME == SC { ne++; ek[ne] = $1; es[ne] = $2; eo[ne] = $3 + 0; known[$2] = 1; next }
    FILENAME == BOUND { bdet[$2] = $3; next }
    FILENAME == GONE { ng++; gv[ng] = $2; gd[ng] = $3 " " $4 " " $5; next }
    FILENAME == OBS { k = $3 SUBSEP $4; if (!(k in om) || $5 + 0 > om[k]) om[k] = $5 + 0; next }
    FILENAME == PROJ { proj[$1] = 1; next }
    FILENAME == START { split($0, a, " "); start[a[1], a[2]] = 1; next }
    FILENAME == LIVE { live[$1, $2] = 1; if ($1 != "image" && ((RUN != "" && $4 == RUN) || ($3 != "" && ($3 in proj)))) lab[$1, $2] = 1; next }
    FILENAME == LAT { if ($1 != "") { held[$1] = 1; if ($1 == "run") heldrun = 1 } next }
    FILENAME == REG {
      L = FNR
      while (ep < ne && eo[ep + 1] < L) { ep++; if (ek[ep] == "begin") st[++sp] = es[ep]; else if (sp > 0) sp-- }
      k = $1; id = $2
      if (k == "standalone") k = "container"; else if (k == "image") id = $3; else if (k != "container" && k != "network" && k != "volume") next
      if (id == "" || ((k, id) in own)) next
      sc = (id == W && k == "container") ? "instrumentation" : (sp > 0 ? st[sp] : "unscoped")
      own[k, id] = sc; cre[sc, k]++; cre["run", k]++; used[sc] = 1
      det = (k == "image") ? $2 : (k == "volume" && (id in bdet) ? bdet[id] : $3)
      created[++nc] = "created" OFS sc OFS k OFS id OFS (det == "" ? "-" : det)
      next }
    END {
      for (key in own) if (key in live) { split(key, a, SUBSEP); left[own[key], a[1]]++; left["run", a[1]]++ }
      for (key in lab) if (!(key in own)) { split(key, a, SUBSEP); un[a[1]]++; finding[++nf] = "finding" OFS "run" OFS a[1] OFS a[2] OFS "unaccounted: labelled object without a registry row" }
      for (key in live) { split(key, a, SUBSEP); if (a[1] == "image" && !(key in start) && !(key in own)) pulled[++np] = "pulled" OFS "run" OFS "image" OFS a[2] OFS "retained (shared cache)" }
      print "# G064 resources.tsv (Plan A corr1 C1, corr2 D1). creations, final_leftover, unaccounted: M (exact IDs, deduplicated); observed_max: M sampled (a lower bound of the true peak; - = no sample); ceiling: D; historical peaks: U (not here)."
      print "row", "scope", "class", "creations", "observed_max", "ceiling", "final_leftover", "unaccounted", "status"
      n = 0; order[++n] = "run"; order[++n] = "instrumentation"
      for (s in known) if (s != "run" && s != "instrumentation") order[++n] = s
      for (i = 1; i <= n; i++) {
        s = order[i]; if (s == "unscoped" && !(s in used) && !(("unscoped" SUBSEP "container") in om)) continue
        split(ceil[s] != "" ? ceil[s] : ceil["run"], cv, "/")
        for (j = 1; j <= 4; j++) {
          k = (j == 1 ? "container" : j == 2 ? "network" : j == 3 ? "volume" : "image")
          cr = cre[s, k] + 0; lf = left[s, k] + 0; ce = (j < 4 ? cv[j] : "-"); ob = ((s, k) in om) ? om[s, k] : "-"
          ua = (s == "run" && j < 4) ? un[k] + 0 : "-"; why = ""
          if ((s in held) && j == 1) why = why "latched;"
          if (j < 4 && ob != "-" && ce ~ /^[0-9]+$/ && ob + 0 > ce + 0) why = why "observed_max>" ce ";"
          if (j == 1 && lf > 0) why = why "final_leftover;"
          if (ua != "-" && ua > 0) why = why "unaccounted;"
          print "total", s, k, cr, ob, ce, lf, ua, (why == "" ? (j == 4 ? "reported" : "PASS") : "FAIL " why)
        }
      }
      for (i = 1; i <= nc; i++) print created[i]
      for (i = 1; i <= ng; i++) print "gone", ((("volume", gv[i]) in own) ? own["volume", gv[i]] : "-"), "volume", gv[i], gd[i]
      for (i = 1; i <= np; i++) print pulled[i]
      for (i = 1; i <= nf; i++) print finding[i]
    }' "$ceil" "$WORK/res-scopes.tsv" "$wd/bound.tsv" "$wd/gone.tsv" "$WORK/res-obs.tsv" "$WORK/projects" \
       "$WORK/inventory-start.ids" "$live" "$latched" "$OWNED_REG" 2>/dev/null; } >"$out.tmp" || fails+=("run: accounting: resources.tsv not written")
  mv -f -- "$out.tmp" "$out" 2>/dev/null || fails+=("run: accounting: resources.tsv not written")
  while IFS=$'\t' read -r _ sc c _ _ _ _ _ why; do
    if [[ "$why" == FAIL* ]]; then fails+=("$sc $c: ${why#FAIL }"); fi
  done < <(grep '^total' "$out" 2>/dev/null || true)
  # Every latch is a RESOURCES FAIL row, including one set in a child shell that recorded no row here (the witness).
  while IFS=$'\t' read -r sc why; do
    [[ -n "$sc" ]] || continue
    if ! grep -qF -- $'RESOURCES\tFAIL\t'"$sc: " "$RESULTS" 2>/dev/null; then fails+=("$sc: $why (latch recorded at report)"); fi
  done <"$latched"
  for line in "${fails[@]}"; do record RESOURCES FAIL "$line"; done
  if (( ${#fails[@]} == 0 )); then log "G064-RES resources.tsv: no accounting finding"; fi
  return 0
}

# --- G064 batch runner (Plan addendum A §A2-A4; correction 1 §C3, §C4) ------------------------------------------------
# One long-lived runner container per batch (B-MIG) receives one `docker exec` per eligible Go invocation; db_fresh still
# precedes each. State lives in files under $WORK/it-runner/ because callers run inside $(...):
#   BATCH.gen (generation N), BATCH.id, BATCH.base (started, sleep_pid, image, gocache, teldir), BATCH.gen<N>.gone (only
#   after verified teardown), BATCH.faults, BATCH.stopped (terminal: named batch ends and EXIT), BATCH.held (terminal,
#   with its reason). A failed or unverified step is never treated as absence; Docker unavailability holds the batch.
ITR_WHY="" ITR_FINDING=""
ITR_SLEEP=2147483647
g003_state_root() { printf '%s' "${G003_STATE_ROOT:-${XDG_STATE_HOME:-$HOME/.local/state}/platform-evidence/g003-local}"; }
# it_runner_precondition: a valid run registry plus a durable recovery registration (§C4). CI: the job registry of this
# CI_JOB_ID (fresh daemon and the after_script backstop). Local: the persistent marker STATE_ROOT/g003-pending/<job>.tsv
# naming this registry with stage launching|launched (§D3). Otherwise ITR_WHY=no-registry|no-durable-recovery, return 1.
it_runner_precondition() {
  local reg d job tok repo m st
  ITR_WHY=no-registry
  reg="${OWNED_REG:-}"
  [[ -n "$reg" && -f "$reg" && ! -L "$reg" ]] || return 1
  d="$(dirname -- "$reg")"
  [[ -d "$d" && ! -L "$d" && "$(stat -c %a -- "$d" 2>/dev/null)" == 700 && -O "$d" && -d "$d/cid" ]] || return 1
  job="$(awk -F'\t' '$1 == "job" { n++; v = $2 } END { if (n == 1) print v }' "$reg")"
  tok="$(awk -F'\t' '$1 == "token" { n++; v = $2 } END { if (n == 1) print v }' "$reg")"
  repo="$(awk -F'\t' '$1 == "repo" { n++; v = $2 } END { if (n == 1) print v }' "$reg")"
  [[ "$job" =~ ^([1-9][0-9]{0,18}|local-[0-9a-f]{12})$ && "$tok" =~ ^[0-9a-f]{32}$ ]] || return 1
  [[ -f "$d/token" && "$(cat -- "$d/token" 2>/dev/null)" == "$tok" ]] || return 1
  [[ "$repo" == "$(realpath -e -- "$REPO" 2>/dev/null)" ]] || return 1
  ITR_WHY=no-durable-recovery
  if [[ "${GITLAB_CI:-}" == true ]]; then [[ "$job" == "${CI_JOB_ID:-}" ]] || return 1; ITR_WHY=""; return 0; fi
  [[ "$job" == local-* ]] || return 1
  m="$(g003_state_root)/g003-pending/$job.tsv"
  [[ -f "$m" && ! -L "$m" ]] || return 1
  [[ "$(awk -F'\t' '$1 == "registry" { n++; v = $2 } END { if (n == 1) print v }' "$m" 2>/dev/null)" == "$d" ]] || return 1
  st="$(awk -F'\t' '$1 == "stage" { n++; v = $2 } END { if (n == 1) print v }' "$m" 2>/dev/null)"
  [[ "$st" == launching || "$st" == launched ]] || return 1
  ITR_WHY=""
}
_itr_get() { awk -F'\t' -v k="$2" '$1 == k { print $2; exit }' "$WORK/it-runner/$1.base" 2>/dev/null; } # BATCH KEY
_itr_diag() { printf '[%s] %s: %s\n' "$(date -u +%H:%M:%S)" "$1" "$2" >>"$WORK/it-runner-$1.log"; }      # BATCH LINE
# docker_top_args ID: the container's process command lines, one per line. The daemon refuses a ps format without a PID
# column ("Couldn't find PID field in ps output", so never `-eo args`): `docker top ID -eo pid,args` prints a header
# ("PID COMMAND") and space-padded "<pid> <args>" rows; only rows whose first field is numeric are kept, PID stripped.
# Returns 1 (prints nothing) when top fails.
docker_top_args() {
  local out
  out="$(docker top "$1" -eo pid,args 2>/dev/null)" || return 1
  awk '$1 ~ /^[0-9]+$/ { sub(/^[ \t]*[0-9]+[ \t]*/, ""); sub(/[ \t]+$/, ""); print }' <<<"$out"
}
# _itr_top ID -> ITR_TOP_SLEEP (host PID of the one sleep child) when the process set is exactly the idle baseline:
# PID 1 = /sbin/docker-init -- sleep N, and one live, non-zombie child of it = sleep N. Otherwise ITR_FINDING.
_itr_top() {
  local out n=0 ipid="" spid="" pid ppid stat args
  ITR_TOP_SLEEP=""
  out="$(docker top "$1" -eo pid,ppid,stat,args 2>&1)" || { ITR_FINDING="runner-fault: top: $(head -1 <<<"$out")"; return 1; }
  while read -r pid ppid stat args; do
    [[ "$pid" =~ ^[0-9]+$ ]] || continue
    n=$((n + 1))
    if [[ "$args" == "/sbin/docker-init -- sleep $ITR_SLEEP" ]]; then ipid="$pid"
    elif [[ "$args" == "sleep $ITR_SLEEP" && "$stat" != Z* ]]; then spid="$pid:$ppid"; fi
  done <<<"$out"
  if (( n != 2 )) || [[ -z "$ipid" || "${spid#*:}" != "$ipid" ]]; then
    ITR_FINDING="runner-residue: process ($n processes: $(awk 'NR > 1 { $1 = $2 = ""; print }' <<<"$out" | tr -s ' \n' ' ' | head -c 200))"; return 1
  fi
  ITR_TOP_SLEEP="${spid%%:*}"
}
# _itr_diff BATCH ID: every A/D path of `docker diff` lies under an allowed root (GOCACHE, GOTELEMETRYDIR, /tmp); C only on
# those roots and their ancestors. Otherwise ITR_FINDING=runner-residue: file <path>. A line exactly equal to a line of
# the generation's start baseline BATCH.diff0 (it_runner_start, before the first exec: e.g. a fresh --init container's
# "A /sbin" and "A /sbin/docker-init", or "C /sbin" when the image has /sbin) is not residue; any other line is checked.
_itr_diff() {
  local out k p r ok roots=() base
  for k in gocache teldir; do r="$(_itr_get "$1" "$k")"; [[ "$r" == /?* ]] && roots+=("${r%/}"); done
  roots+=(/tmp)
  base="$(cat "$WORK/it-runner/$1.diff0" 2>/dev/null)" || base=""
  out="$(docker diff "$2" 2>&1)" || { ITR_FINDING="runner-fault: diff: $(head -1 <<<"$out")"; return 1; }
  while read -r k p; do
    [[ -n "$k" ]] || continue
    if [[ -n "$base" ]] && grep -qxF -- "$k $p" <<<"$base"; then continue; fi
    ok=0
    for r in "${roots[@]}"; do
      if [[ "$p" == "$r" || "$p" == "$r"/* ]]; then ok=1; break; fi
      if [[ "$k" == C && ( "$p" == / || "$r" == "$p"/* ) ]]; then ok=1; break; fi
    done
    (( ok )) || { ITR_FINDING="runner-residue: file $p"; return 1; }
  done <<<"$out"
}
# it_runner_start BATCH IMAGE NETWORK WANT_IMAGE_ID: allocate generation N+1 (§C3) when there is no stopped/held state,
# generation N's teardown was verified, faults < 2, the resource latch allows it and the durable registration holds.
# Prints nothing; sets ITR_WHY on refusal (return 1).
it_runner_start() {
  local b="$1" img="$2" net="$3" want="$4" d="$WORK/it-runner" n f id out started image gc tel
  mkdir -p "$d"
  ITR_WHY=""
  [[ -e "$d/$b.stopped" ]] && { ITR_WHY="stopped"; return 1; }
  [[ -e "$d/$b.held" ]] && { ITR_WHY="held: $(head -1 "$d/$b.held")"; return 1; }
  n="$(cat "$d/$b.gen" 2>/dev/null || echo 0)"; f="$(cat "$d/$b.faults" 2>/dev/null || echo 0)"
  if (( n > 0 )) && [[ ! -e "$d/$b.gen$n.gone" ]]; then ITR_WHY="generation $n teardown unverified"; return 1; fi
  (( f < 2 )) || { ITR_WHY="repeated runner fault"; return 1; }
  if declare -F res_latch_held >/dev/null && res_latch_held "$b"; then ITR_WHY="held: resource latch"; return 1; fi
  it_runner_precondition || return 1
  n=$((n + 1))
  printf '%s\n' "$n" >"$d/$b.gen"; rm -f -- "$d/$b.id" "$d/$b.base" "$d/$b.diff0"
  id="$(docker_run_owned it-runner -d --init --pull never --network "$net" --label "g001.run=$RUN_ID" \
    --label "g064.batch=$b-$RUN_ID" --name "g064-itr-$RUN_ID-$b-$n" --entrypoint sleep "$img" "$ITR_SLEEP" 2>&1)" || {
    ITR_WHY="runner create failed: $(tail -1 <<<"$id")"; _itr_diag "$b" "$ITR_WHY"
    it_runner_teardown "$b" "create failed" "$(docker ps -aq --no-trunc --filter "label=g064.batch=$b-$RUN_ID" 2>/dev/null | head -1)" || true
    return 1; }
  id="$(tail -1 <<<"$id")"
  [[ "$id" =~ ^[0-9a-f]{64}$ ]] || { ITR_WHY="runner ID not recorded at creation"; it_runner_teardown "$b" "$ITR_WHY" "" || true; return 1; }
  printf '%s\n' "$id" >"$d/$b.id"
  out="$(docker inspect -f '{{.State.StartedAt}}|{{.Image}}|{{len .Mounts}}' "$id" 2>&1)" || out="!$out"
  IFS='|' read -r started image f <<<"$out"
  if [[ "$f" != 0 || "$image" != "$want" || "$started" == '!'* ]]; then
    ITR_WHY="runner refused at start (mounts=${f:-?} image=${image:0:19} want=${want:0:19})"; _itr_diag "$b" "$ITR_WHY"
    it_runner_teardown "$b" "$ITR_WHY" "$id" || true; return 1
  fi
  ITR_FINDING=""
  if ! _itr_top "$id"; then ITR_WHY="runner baseline: $ITR_FINDING"; _itr_diag "$b" "$ITR_WHY"; it_runner_teardown "$b" "$ITR_WHY" "$id" || true; return 1; fi
  # The filesystem baseline before any exec (what the runtime adds itself, e.g. --init's /sbin/docker-init): _itr_diff
  # skips exactly these lines. Taken before `go env`, the first exec; a failed diff refuses the runner.
  out="$(docker diff "$id" 2>&1)" || { ITR_WHY="runner baseline: diff: $(head -1 <<<"$out")"; _itr_diag "$b" "$ITR_WHY"; it_runner_teardown "$b" "$ITR_WHY" "$id" || true; return 1; }
  { [[ -z "$out" ]] || printf '%s\n' "$out"; } >"$d/$b.diff0"
  out="$(docker exec "$id" go env GOCACHE GOTELEMETRYDIR 2>&1)" || { ITR_WHY="runner go env: $(head -1 <<<"$out")"; it_runner_teardown "$b" "$ITR_WHY" "$id" || true; return 1; }
  gc="$(sed -n 1p <<<"$out")"; tel="$(sed -n 2p <<<"$out")"
  printf 'id\t%s\nstarted\t%s\nsleep_pid\t%s\nimage\t%s\ngocache\t%s\nteldir\t%s\n' "$id" "$started" "$ITR_TOP_SLEEP" "$image" "$gc" "$tel" >"$d/$b.base"
  _itr_diag "$b" "generation $n started: ${id:0:12} sleep_pid=$ITR_TOP_SLEEP gocache=$gc teldir=$tel"
}
# it_runner_check BATCH pre|post: §C3 per-exec ordering. pre: identity, process set, diff, /tmp empty. post: (a) identity,
# (b) process set, (c) diff, each before any reset; only then (d) /tmp reset and verified empty. ITR_FINDING on failure.
it_runner_check() {
  local b="$1" ph="$2" id out
  ITR_FINDING=""
  id="$(cat "$WORK/it-runner/$b.id" 2>/dev/null)" || { ITR_FINDING="runner-fault: no runner"; return 1; }
  out="$(docker inspect -f '{{.Id}}|{{.State.Running}}|{{.State.StartedAt}}' "$id" 2>&1)" || { ITR_FINDING="runner-fault: inspect: $(docker_msg_line "$out")"; return 1; }
  [[ "$out" == "$id|true|$(_itr_get "$b" started)" ]] || { ITR_FINDING="runner-fault: identity ($out)"; return 1; }
  if ! _itr_top "$id"; then # post: the file residue is reported too (T-9), still before any reset
    if [[ "$ph" == post ]]; then out="$ITR_FINDING"; _itr_diff "$b" "$id" || out+="; $ITR_FINDING"; ITR_FINDING="$out"; fi
    return 1
  fi
  [[ "$ITR_TOP_SLEEP" == "$(_itr_get "$b" sleep_pid)" ]] || { ITR_FINDING="runner-fault: sleep pid changed"; return 1; }
  _itr_diff "$b" "$id" || return 1
  if [[ "$ph" == post ]]; then
    out="$(docker exec "$id" find /tmp -mindepth 1 -delete 2>&1)" || { ITR_FINDING="runner-residue: tmp reset failed: $(head -1 <<<"$out")"; return 1; }
  fi
  out="$(docker exec "$id" find /tmp -mindepth 1 -print -quit 2>&1)" || { ITR_FINDING="runner-fault: tmp check: $(head -1 <<<"$out")"; return 1; }
  [[ -z "$out" ]] || { ITR_FINDING="runner-residue: tmp $out"; return 1; }
}
# docker_msg_line TEXT: the first meaningful line of a Docker CLI output: empty lines and the "[]" that an inspect of a
# missing object prints on stdout (before its "No such ..." error on stderr) are skipped.
docker_msg_line() { awk 'NF && $0 != "[]" { print; exit }' <<<"$1"; }
# docker_container_gone ID: 0 only when Docker positively reports the container absent: `docker inspect --type container
# ID` fails AND the exact-ID listing `docker ps -aq --no-trunc --filter id=ID` succeeds with empty output. (A removed
# ID's inspect prints "[]" and "Error: No such container|object: ID" with rc 1, so its text is not the proof; a daemon
# that cannot be reached fails the listing too and is never "gone".) Otherwise 1, DCG_WHY = still present | still listed
# | the inspect and listing errors (first meaningful lines).
docker_container_gone() {
  local out q
  DCG_WHY=""
  if out="$(docker inspect --type container "$1" 2>&1)"; then DCG_WHY="still present"; return 1; fi
  if q="$(docker ps -aq --no-trunc --filter "id=$1" 2>&1)"; then
    [[ -n "$q" ]] || return 0
    DCG_WHY="still listed"; return 1
  fi
  DCG_WHY="$(docker_msg_line "$out"); listing: $(docker_msg_line "$q")"; return 1
}
# docker_rm_pending TEXT: a `docker rm` refused because the daemon is already removing that container (real text
# "Error response from daemon: removal of container ID is already in progress", rc 1): a concurrent `rm -f` of the same
# ID, e.g. T-10's or the F1 rkill's own removal racing the exec path's fault teardown. Neither a failure nor a proof of
# absence: the caller waits for docker_container_gone (docker_container_wait_gone).
docker_rm_pending() { grep -qiE 'removal of container [^ ]+ is already in progress' <<<"$1"; }
# docker_container_wait_gone ID SECONDS: docker_container_gone polled every 0.5 s until it holds (0) or SECONDS pass (1,
# DCG_WHY: still present/listed, or the daemon unreachable: never gone).
docker_container_wait_gone() {
  local i n=$(( ${2:-0} * 2 ))
  for ((i = 0; ; i++)); do
    docker_container_gone "$1" && return 0
    (( i < n )) || return 1
    sleep 0.5
  done
}
# it_runner_teardown BATCH REASON [ID]: verified teardown (§C3): rm -f the exact ID (a removal already in progress is
# waited for, bounded); then docker_container_gone (inspect fails and the exact-ID listing succeeds empty); the g064.batch label query must succeed and be empty now and after 6 s
# (a late hit is removed, recorded and still holds). Verified: BATCH.gen<N>.gone. Unverified: BATCH.held, a RESOURCES
# FAIL row and diagnostics; return 1.
it_runner_teardown() {
  local b="$1" why="$2" id="${3-}" d="$WORK/it-runner" n out step="" late x
  n="$(cat "$d/$b.gen" 2>/dev/null || echo 0)"
  [[ $# -ge 3 ]] || id="$(cat "$d/$b.id" 2>/dev/null)" || id=""
  if [[ -n "$id" ]]; then
    # A runner already gone (killed mid-exec, T-10) is not a removal failure; the inspect proof below still decides. A
    # removal already in progress (T-10/rkill: the caller's `docker rm -f` of the runner kills the exec, whose fault
    # teardown then reaches this rm while that removal still runs) is pending: docker_container_gone is polled for up
    # to ITR_RM_S (default 30) seconds; not gone by then (or the daemon unreachable) is unverified.
    out="$(docker rm -f "$id" 2>&1)" || docker_rm_pending "$out" || grep -qi 'no such container' <<<"$out" || step="rm: $(head -1 <<<"$out")"
    if [[ -z "$step" ]] && docker_rm_pending "$out" && ! docker_container_wait_gone "$id" "${ITR_RM_S:-30}"; then
      step="rm: removal already in progress, not verified gone within ${ITR_RM_S:-30}s ($DCG_WHY)"
    fi
    if [[ -z "$step" ]] && ! docker_container_gone "$id"; then step="inspect: $DCG_WHY"; fi
  fi
  for late in 0 6; do
    [[ -n "$step" ]] && break
    (( late == 0 )) || sleep "${ITR_LATE_S:-6}"
    if out="$(docker ps -aq --no-trunc --filter "label=g064.batch=$b-$RUN_ID" 2>&1)"; then
      for x in $out; do docker rm -f "$x" >/dev/null 2>&1 || true; _itr_diag "$b" "late recreation ${x:0:12} removed"; step="late recreation ${x:0:12}"; done
    else step="query: $(head -1 <<<"$out")"; fi
  done
  if [[ -z "$step" ]]; then
    : >"$d/$b.gen$n.gone"; rm -f -- "$d/$b.id"
    _itr_diag "$b" "generation $n removed and verified ($why)"; return 0
  fi
  printf 'runner teardown unverified (%s)\nreason\t%s\n' "$step" "$why" >>"$d/$b.held"
  _itr_diag "$b" "generation $n teardown unverified: $step ($why)"
  record RESOURCES FAIL "$b: runner teardown unverified ($step)"
  if declare -F res_latch_set >/dev/null; then res_latch_set "$b" "runner teardown unverified ($step)" || true; fi
  return 1
}
# it_runner_fault BATCH FINDING: a faulted or residue-bearing runner is removed at once (verified); the second fault in a
# batch holds it ("repeated runner fault"). Returns 1 when the batch is now held.
it_runner_fault() {
  local b="$1" d="$WORK/it-runner" f
  f="$(cat "$d/$b.faults" 2>/dev/null || echo 0)"; f=$((f + 1)); printf '%s\n' "$f" >"$d/$b.faults"
  _itr_diag "$b" "fault $f: $2"
  it_runner_teardown "$b" "$2" || return 1
  if (( f >= 2 )); then printf 'repeated runner fault\nreason\t%s\n' "$2" >>"$d/$b.held"; return 1; fi
  return 0
}
# it_runner_stop BATCH: terminal stop at a named batch end (and on EXIT): a live generation is torn down (verified),
# then BATCH.stopped. Returns 1 when the teardown was unverified (BATCH.held is then set too).
it_runner_stop() {
  local b="$1" d="$WORK/it-runner" rc=0 n
  [[ -e "$d/$b.gen" ]] || { mkdir -p "$d"; : >"$d/$b.stopped"; return 0; }
  n="$(cat "$d/$b.gen")"
  if [[ ! -e "$d/$b.gen$n.gone" ]]; then it_runner_teardown "$b" "batch end" || rc=1; fi
  : >"$d/$b.stopped"
  return "$rc"
}
it_runner_stop_all() { # EXIT path: every batch with a generation that is not yet stopped
  local f b
  for f in "${WORK:-/nonexistent}"/it-runner/*.gen; do
    [[ -f "$f" ]] || continue
    b="${f##*/}"; b="${b%.gen}"
    [[ -e "${f%.gen}.stopped" ]] || it_runner_stop "$b" || log "G064 runner $b: teardown unverified at exit"
  done
  return 0
}
# it_runner_ready BATCH IMAGE NETWORK WANT_IMAGE_ID: the batch's live runner (ITR_ID), starting the next generation lazily
# when none is live (§C3: a faulted runner is replaced by the next eligible invocation). On failure (return 1) ITR_WHY says
# why and ITR_FALLBACK=1 only for the §C4 precondition refusals (no-registry|no-durable-recovery), which take today's
# per-invocation container path; every other refusal FAILs the invocation (held/stopped batch; a failed start counts as a
# runner fault, so a second one holds the batch). No silent fallback.
it_runner_ready() {
  local b="$1" d="$WORK/it-runner" why
  ITR_ID="" ITR_FALLBACK=0
  if [[ -e "$d/$b.held" ]]; then ITR_WHY="batch held: $(head -1 "$d/$b.held")"; return 1; fi
  if [[ -e "$d/$b.stopped" ]]; then ITR_WHY="batch stopped"; return 1; fi
  # D2: a latched scope (or run) holds dispatch even with a live runner; the runner stays until its batch end.
  if declare -F res_latch_held >/dev/null && res_latch_held "$b"; then ITR_WHY="batch held: resource latch (${RES_WHY#held: })"; return 1; fi
  if [[ -s "$d/$b.id" ]]; then ITR_ID="$(cat "$d/$b.id")"; return 0; fi
  if it_runner_start "$b" "$2" "$3" "$4"; then ITR_ID="$(cat "$d/$b.id")"; return 0; fi
  case "$ITR_WHY" in
    no-registry|no-durable-recovery) ITR_FALLBACK=1; return 1 ;;
    held:*|stopped) ITR_WHY="batch $ITR_WHY"; return 1 ;;
  esac
  why="runner-fault: start: $ITR_WHY"
  it_runner_fault "$b" "$why" || true
  ITR_WHY="$why"; return 1
}
# it_runner_exec BATCH OUT CMD...: one exec in the batch's live runner in the §C3 order: pre-check, then
# `docker exec "${GO_IT_XARGS[@]}" ID CMD... >OUT 2>&1` (GO_IT_XARGS: the caller's -e K=V / --env-file F arguments, the
# exec's only environment), its status captured first, then the post-check. ITR_RC = the exec status; 125 when it did not
# run (failed pre-check) or when it exited 0 but the post-check failed. ITR_FINDING = the runner-fault/runner-residue
# finding (empty when clean); a finding removes the runner at once (it_runner_fault: verified teardown, second fault
# holds). Returns 0 only when the runner was clean before and after the exec, whatever ITR_RC.
it_runner_exec() {
  local b="$1" out="$2" id f; shift 2
  ITR_RC=125 ITR_FINDING=""
  id="$(cat "$WORK/it-runner/$b.id" 2>/dev/null)"
  if ! it_runner_check "$b" pre; then
    f="$ITR_FINDING"; printf 'G064 runner %s: pre-exec check failed (%s); the command did not run\n' "${id:0:12}" "$f" >"$out"
    it_runner_fault "$b" "$f" || true; ITR_FINDING="$f"; return 1
  fi
  ITR_RC=0
  docker exec "${GO_IT_XARGS[@]}" "$id" "$@" >"$out" 2>&1 || ITR_RC=$?
  it_runner_check "$b" post && return 0
  f="$ITR_FINDING"; _itr_diag "$b" "exec rc=$ITR_RC post-check: $f"
  it_runner_fault "$b" "$f" || true
  ITR_FINDING="$f"; ((ITR_RC != 0)) || ITR_RC=125
  return 1
}

# --- CI-control resource owners and the startup guard (G064 correction 11 item 4; containment only) -----------------
# Every Docker object infra/ci/validate-controls.sh creates carries sdlc.ci.owner=<KEY> from its creation. KEY names
# the process responsible for removing the object (the controlling runner; see validate-controls.sh for nested runs) and
# binds it to one boot and one PID namespace, so that neither a reboot nor a reused PID can pass as that owner, and to
# the canonical checkout and control run:
#   v1.<boot id, 32 hex>.<PID namespace inode>.<PID>.<start time: /proc/PID/stat field 22, clock ticks>.<checkout id:
#   first 16 hex of sha256(canonical checkout root path)>.<control run id, 12 hex>
# Nothing in it is secret. These helpers only read /proc and Docker: they never remove, stop or signal anything, and a
# label, a PID or an age alone never authorizes anything. Durable registry/marker recovery is not part of them.
G003_OWNER_RE='^v1\.([0-9a-f]{32})\.([1-9][0-9]{0,19})\.([1-9][0-9]{0,9})\.(0|[1-9][0-9]{0,19})\.([0-9a-f]{16})\.([0-9a-f]{12})$'
G003_KEYS="sdlc.ci.owner sdlc.ci.run sdlc.ci.control sdlc.ci.outer" # an object carrying any of them is a control's
G003_GUARD_MAX=20 G3_STATE="" G3_PPID="" G3_START=""
g003_boot_id() { local b; b="$(cat /proc/sys/kernel/random/boot_id 2>/dev/null)" || return 1; b="${b//-/}"; [[ "$b" =~ ^[0-9a-f]{32}$ ]] || return 1; printf '%s' "$b"; }
g003_pidns() { local l; l="$(readlink "/proc/$1/ns/pid" 2>/dev/null)" || return 1; [[ "$l" =~ ^pid:\[([1-9][0-9]{0,19})\]$ ]] || return 1; printf '%s' "${BASH_REMATCH[1]}"; }
g003_stat() { # PID -> G3_STATE, G3_PPID, G3_START (/proc/PID/stat fields 3, 4, 22); fails when unreadable or malformed
  local s f
  s="$(cat "/proc/$1/stat" 2>/dev/null)" || return 1
  s="${s##*) }"
  read -r -a f <<<"$s"
  [[ "${f[0]:-}" =~ ^[A-Za-z]$ && "${f[1]:-}" =~ ^[0-9]+$ && "${f[19]:-}" =~ ^[0-9]+$ ]] || return 1
  G3_STATE="${f[0]}" G3_PPID="${f[1]}" G3_START="${f[19]}"
}
g003_checkout_id() { local r h; r="$(realpath -e -- "$1" 2>/dev/null)" || return 1; h="$(printf '%s' "$r" | sha256sum)" || return 1; printf '%s' "${h:0:16}"; }
g003_owner_key() { # PID RUN CHECKOUT -> the owner key of that live process for control run RUN of CHECKOUT
  local b ns c
  [[ "$1" =~ ^[1-9][0-9]{0,9}$ && "$2" =~ ^[0-9a-f]{12}$ ]] || return 1
  b="$(g003_boot_id)" || return 1
  ns="$(g003_pidns "$1")" || return 1
  c="$(g003_checkout_id "$3")" || return 1
  g003_stat "$1" || return 1
  [[ "$G3_STATE" != [ZX] ]] || return 1
  printf 'v1.%s.%s.%s.%s.%s.%s' "$b" "$ns" "$1" "$G3_START" "$c" "$2"
}
# /proc mounted with hidepid=2 (invisible) hides other users' processes: an absent PID is then no proof that it ended.
g003_proc_hides() { awk '$2 == "/proc" && $3 == "proc" && $4 ~ /(^|,)hidepid=(2|invisible)(,|$)/ { f = 1 } END { exit !f }' /proc/mounts 2>/dev/null; }
# g003_owner_class VALUE MYKEY -> own (VALUE is MYKEY: the caller's controlling run) | live (another live owner: another
# run, another checkout) | dead:<why> (verifiably ended: boot, pid-absent, pid-reuse, exited) | unknown:<why> (missing,
# malformed, own-identity, pid-namespace, proc-hidden, proc-unreadable). Same boot, same PID namespace, the PID present
# with the recorded start time and not a zombie: live. Missing or unreadable proof is unknown, never dead or clean.
g003_owner_class() {
  local v="$1" me="$2" mb mns b ns p st
  if [[ "$me" =~ $G003_OWNER_RE ]]; then mb="${BASH_REMATCH[1]}" mns="${BASH_REMATCH[2]}"; else printf 'unknown:own-identity'; return 0; fi
  if [[ -z "$v" || "$v" == "<no value>" ]]; then printf 'unknown:missing'; return 0; fi
  if [[ "$v" =~ $G003_OWNER_RE ]]; then :; else printf 'unknown:malformed'; return 0; fi
  b="${BASH_REMATCH[1]}" ns="${BASH_REMATCH[2]}" p="${BASH_REMATCH[3]}" st="${BASH_REMATCH[4]}"
  if [[ "$b" != "$mb" ]]; then printf 'dead:boot'; return 0; fi
  if [[ "$ns" != "$mns" ]]; then printf 'unknown:pid-namespace'; return 0; fi
  if g003_stat "$p"; then :
  elif [[ -e "/proc/$p" ]]; then printf 'unknown:proc-unreadable'; return 0
  elif g003_proc_hides; then printf 'unknown:proc-hidden'; return 0
  else printf 'dead:pid-absent'; return 0; fi
  if [[ "$G3_START" != "$st" ]]; then printf 'dead:pid-reuse'; return 0; fi
  if [[ "$G3_STATE" == [ZX] ]]; then printf 'dead:exited'; return 0; fi
  if [[ "$v" == "$me" ]]; then printf 'own'; else printf 'live'; fi
}
# g003_owner_ancestor KEY PID CHECKOUT: KEY is the owner key of a live ancestor of PID for CHECKOUT (same boot, PID
# namespace, start time and checkout id). A nested controls run uses its controlling runner's key only after this.
g003_owner_ancestor() {
  local k="$1" p="$2" kp kr n
  if [[ "$k" =~ $G003_OWNER_RE ]]; then kp="${BASH_REMATCH[3]}" kr="${BASH_REMATCH[6]}"; else return 1; fi
  [[ "$(g003_owner_key "$kp" "$kr" "$3")" == "$k" ]] || return 1
  for ((n = 0; n < 64; n++)); do
    g003_stat "$p" || return 1
    p="$G3_PPID"
    if [[ "$p" == "$kp" ]]; then [[ "$(g003_owner_key "$kp" "$kr" "$3")" == "$k" ]]; return; fi
    (( p > 1 )) || return 1
  done
  return 1
}
g003_ls() { # DOCKER KIND [FILTER]: the full IDs (volumes: names) of every object of KIND, or of those matching FILTER
  local f=()
  if [[ -n "${3:-}" ]]; then f=(--filter "$3"); fi
  case "$2" in
    container) timeout 60 "$1" ps -aq --no-trunc "${f[@]}" ;;
    network) timeout 60 "$1" network ls -q --no-trunc "${f[@]}" ;;
    volume) timeout 60 "$1" volume ls -q "${f[@]}" ;;
    image) timeout 60 "$1" image ls -aq --no-trunc "${f[@]}" ;;
  esac 2>/dev/null
}
g003_describe() { # KIND ID CLASS VALUE MYKEY -> one bounded, nonsecret line for a held object
  local k="$1" id="${2#sha256:}" c="$3" v="$4" me="$5" who="" p co run
  if [[ "$k" == volume ]]; then id="$(tr -c 'A-Za-z0-9_.\n-' '?' <<<"${id:0:64}")"; else id="${id:0:12}"; fi
  if [[ "$v" =~ $G003_OWNER_RE ]]; then
    p="${BASH_REMATCH[3]}" co="${BASH_REMATCH[5]}" run="${BASH_REMATCH[6]}"
    if [[ "$me" =~ $G003_OWNER_RE && "$co" == "${BASH_REMATCH[5]}" ]]; then co="this checkout"; else co="checkout id $co"; fi
    who="; pid $p, run $run, $co"
  fi
  if [[ "$c" == dead:* ]]; then printf 'startup guard: %s %s: owner gone (%s%s)\n' "$k" "$id" "${c#*:}" "$who"
  else printf 'startup guard: %s %s: owner cannot be attributed (%s%s)\n' "$k" "$id" "${c#*:}" "$who"; fi
}
# g003_owner_guard DOCKER MYKEY: the CI controls' startup guard. Every container, network, volume and image carrying a
# CI-control key is classified (g003_owner_class). Objects of the caller's own controlling run and of other live owners
# are ignored and untouched. Any whose owner is gone or cannot be attributed, an object that is listed but cannot be
# read, or a listing Docker refuses (unavailability is not absence) holds: exit 16, with at most G003_GUARD_MAX bounded
# IDs and the operator's recovery steps on stdout. Otherwise one "clear" line, exit 0.
g003_owner_guard() {
  local d="$1" me="$2" k key x v c all fmt n=0 nown=0 nlive=0 lines=() ids objs=()
  for k in container network volume image; do
    if [[ "$k" == container || "$k" == image ]]; then fmt='{{index .Config.Labels "sdlc.ci.owner"}}'; else fmt='{{index .Labels "sdlc.ci.owner"}}'; fi
    ids=""
    for key in $G003_KEYS; do
      if x="$(g003_ls "$d" "$k" "label=$key")"; then ids+="$x"$'\n'; else
        printf 'startup guard: hold: Docker could not list the %ss carrying %s; their owners are unverified (unavailability is not absence); nothing was allocated, removed or stopped\n' "$k" "$key"
        return 16
      fi
    done
    mapfile -t objs < <(awk 'NF && !s[$0]++' <<<"$ids")
    for x in "${objs[@]}"; do
      if v="$(timeout 60 "$d" inspect --type "$k" --format "$fmt" "$x" 2>/dev/null </dev/null)"; then c="$(g003_owner_class "$v" "$me")"
      elif all="$(g003_ls "$d" "$k" </dev/null)"; then
        if grep -qxF -- "$x" <<<"$all"; then c=unknown:unreadable v=""; else continue; fi # removed meanwhile: not a survivor
      else
        printf 'startup guard: hold: Docker could not list the %ss; their owners are unverified (unavailability is not absence); nothing was allocated, removed or stopped\n' "$k"
        return 16
      fi
      case "$c" in
        own) nown=$((nown + 1)) ;;
        live) nlive=$((nlive + 1)) ;;
        *) n=$((n + 1)); if (( n <= G003_GUARD_MAX )); then lines+=("$(g003_describe "$k" "$x" "$c" "$v" "$me")"); fi ;;
      esac
    done
  done
  if (( n == 0 )); then
    printf 'startup guard: clear: no control resource of an ended or unattributable owner (%s of the controlling run and %s of other live owners ignored and untouched)\n' "$nown" "$nlive"
    return 0
  fi
  printf 'startup guard: hold: %s control resource(s) survive whose owner is gone or cannot be attributed; nothing was allocated, removed or stopped\n' "$n"
  printf '%s\n' "${lines[@]}"
  if (( n > G003_GUARD_MAX )); then printf 'startup guard: ... and %s more\n' "$((n - G003_GUARD_MAX))"; fi
  printf 'startup guard: recovery (operator): confirm that no controls run owning these objects is still active; verify each object (docker inspect <id>: its sdlc.ci.* labels name the run and control) and remove only objects positively established as abandoned control resources, recording each removal under the continuous-cleanup policy; then start again. This guard is containment only: it removes, stops and reclassifies nothing, and a held start is not a clean resource receipt.\n'
  return 16
}

# --- documented command blocks ---------------------------------------------------------
# docs_block FILE MARKER -> commands of the fenced sh block after <!-- MARKER -->
docs_block() {
  awk -v m="<!-- $2 -->" '$0==m{f=1; next} f && /^```sh/{c=1; next} c && /^```/{exit} c{print}' "$1"
}
# run_docs_block DIR DOC MARKER LOG: runs the block verbatim from DIR. A block that can build (for example
# `run --rm --build`, whose container is gone before any capture) also gets build records for DIR's images.
run_docs_block() {
  local dir="$1" doc="$2" marker="$3" out="$4" cmds rc=0
  cmds="$(docs_block "$dir/$doc" "$marker")"
  [[ -n "$cmds" ]] || { log "no block $marker in $doc"; return 1; }
  printf '%s\n' "$cmds" >"$out.cmds"
  ledger_run_block "$dir" "$cmds" "$out" || rc=$?
  if _lg_block_buildable "$cmds"; then capture_after "$dir" up build; else capture_after "$dir" up; fi
  if (( rc != 0 )); then log "documented block $marker failed (exit $rc); see diagnostics/${out##*/}"; fi
  return "$rc"
}

# ledger_run_block DIR CMDS LOG: runs CMDS verbatim from DIR (bash -euo pipefail) with both streams in LOG; a block
# with a command that can build is accounted in the build ledger from LOG (G051 §3.8.3).
# G064 D2: a block that allocates or cannot be parsed is latch-checked before (refused: 125, the reason in LOG) and
# recounted after under the scope locks; the block's status is returned unchanged.
ledger_run_block() {
  local dir="$1" cmds="$2" out="$3" rc=0 alloc=0
  if _res_block_allocates "$cmds"; then
    alloc=1
    res_gate "documented block" || { printf 'G064 documented block refused (ran nothing): %s\n' "$RES_WHY" >"$out"; return 125; }
  fi
  if [[ -z "$LEDGER_DIR" ]] || ! _lg_block_buildable "$cmds"; then
    (cd "$dir" && bash -euo pipefail -c "$cmds") >"$out" 2>&1 || rc=$?
  else
    _lg_begin "$dir" "?"
    local id="$LG_CALL"
    _lg_builder "$LEDGER_DIR/progress/$id.bld1"
    (cd "$dir" && bash -euo pipefail -c "$cmds") >"$out" 2>&1 || rc=$?
    _lg_block_end "$id" "$dir" "$rc" "" "$out" "$cmds"
  fi
  (( alloc == 0 )) || res_recount "documented block" || true
  return "$rc"
}

# --- host probes ----------------------------------------------------------------------------
PROBE_MODE=""
detect_probe() {
  if (exec 3<>/dev/tcp/127.0.0.1/1) 2>&1 | grep -q 'No such file or directory\|not supported'; then
    PROBE_MODE="curl"
  else
    PROBE_MODE="devtcp"
  fi
  if [[ "$PROBE_MODE" == curl ]] && ! command -v curl >/dev/null 2>&1; then PROBE_MODE="none"; fi
}

# http_get HOST PORT PATH -> prints "STATUS<TAB>BODY" (body on one line).
http_get() {
  local host="$1" port="$2" path="$3" out status
  if [[ "$PROBE_MODE" == devtcp ]]; then
    out="$(timeout 10 bash -c 'exec 3<>/dev/tcp/'"$host/$port"' || exit 9; printf "GET %s HTTP/1.0\r\nHost: %s\r\nConnection: close\r\n\r\n" "'"$path"'" "'"$host"'" >&3; cat <&3' 2>/dev/null)" || { printf '000\t\n'; return 0; }
    status="$(printf '%s' "$out" | head -1 | awk '{print $2}')"
    printf '%s\t%s\n' "${status:-000}" "$(printf '%s' "$out" | awk 'BEGIN{b=0} b{print} /^\r?$/{b=1}' | tr -d '\r\n')"
  elif [[ "$PROBE_MODE" == curl ]]; then
    out="$(curl -s -m 10 -w '\n%{http_code}' "http://$host:$port$path" 2>/dev/null)" || { printf '000\t\n'; return 0; }
    printf '%s\t%s\n' "$(printf '%s' "$out" | tail -1)" "$(printf '%s' "$out" | sed '$d' | tr -d '\r\n')"
  else
    printf 'NOPROBE\t\n'
  fi
}

tcp_redis_ping() {
  local host="$1" port="$2"
  if [[ "$PROBE_MODE" == devtcp ]]; then
    timeout 5 bash -c 'exec 3<>/dev/tcp/'"$host/$port"' || exit 9; printf "PING\r\n" >&3; head -c 7 <&3' 2>/dev/null
  elif [[ "$PROBE_MODE" == curl ]]; then
    printf 'PING\r\n' | timeout 5 curl -s -m 3 "telnet://$host:$port" 2>/dev/null | head -c 7
  fi
}
tcp_pg_sslrequest() {
  local host="$1" port="$2"
  if [[ "$PROBE_MODE" == devtcp ]]; then
    timeout 5 bash -c 'exec 3<>/dev/tcp/'"$host/$port"' || exit 9; printf "\x00\x00\x00\x08\x04\xd2\x16\x2f" >&3; head -c 1 <&3' 2>/dev/null
  elif [[ "$PROBE_MODE" == curl ]]; then
    printf '\x00\x00\x00\x08\x04\xd2\x16\x2f' | timeout 5 curl -s -m 3 "telnet://$host:$port" 2>/dev/null | head -c 1
  fi
}
tcp_open() { # tcp_open HOST PORT -> success if a TCP connection opens
  if [[ "$PROBE_MODE" == devtcp ]]; then
    timeout 5 bash -c 'exec 3<>/dev/tcp/'"$1/$2" 2>/dev/null
  else
    timeout 5 curl -s -m 3 "telnet://$1:$2" </dev/null >/dev/null 2>&1; [[ $? -ne 7 && $? -ne 28 ]]
  fi
}
ports_busy() { local p; for p in "$@"; do [[ "$p" == "-" ]] && continue; tcp_open 127.0.0.1 "$p" && return 0; done; return 1; }

host_nonloopback_ip() {
  local ip=""
  if command -v hostname >/dev/null 2>&1; then ip="$(hostname -I 2>/dev/null | awk '{print $1}')"; fi
  if [[ -z "$ip" ]] && command -v ip >/dev/null 2>&1; then ip="$(ip -4 -o addr show scope global 2>/dev/null | awk '{print $4}' | cut -d/ -f1 | head -1)"; fi
  printf '%s' "$ip"
}

# http_get_headers HOST PORT PATH -> the status code on the first line, then one
# "name:value" line per response header (lower-case name).
http_get_headers() {
  local host="$1" port="$2" path="$3" out
  if [[ "$PROBE_MODE" == devtcp ]]; then
    out="$(timeout 10 bash -c 'exec 3<>/dev/tcp/'"$host/$port"' || exit 9; printf "GET %s HTTP/1.0\r\nHost: %s\r\nConnection: close\r\n\r\n" "'"$path"'" "'"$host"'" >&3; cat <&3' 2>/dev/null)" || { echo 000; return 0; }
  elif [[ "$PROBE_MODE" == curl ]]; then
    out="$(curl -s -m 10 -D - -o /dev/null "http://$host:$port$path" 2>/dev/null)" || { echo 000; return 0; }
  else
    echo NOPROBE; return 0
  fi
  printf '%s\n' "$out" | tr -d '\r' | awk 'NR==1{print $2; next} /^$/{exit} {i=index($0, ":"); print tolower(substr($0, 1, i - 1)) ":" substr($0, i + 1)}'
}

# no_store HOST PORT PATH STATUS: the response has STATUS and Cache-Control: no-store.
no_store() {
  local h; h="$(http_get_headers "$1" "$2" "$3")"
  [[ "$(head -1 <<<"$h")" == "$4" ]] && grep -qE '^cache-control:.*\bno-store\b' <<<"$h"
}

# wait_http HOST PORT PATH STATUS SECONDS
wait_http() {
  local i
  for ((i = 0; i < $5; i++)); do
    [[ "$(http_get "$1" "$2" "$3" | cut -f1)" == "$4" ]] && return 0
    sleep 1
  done
  return 1
}

# --- in-network probes (busybox sharing a container's network namespace) -------------
cid_of() { compose "$1" ps -q "$2" | head -1; }
ns_tcp() { # ns_tcp CONTAINER_ID HOST PORT -> success if reachable
  docker_run_owned ns-tcp --rm --label g001.run=$RUN_ID --network "container:$1" "$BUSYBOX_IMAGE" nc -z -w 3 "$2" "$3" >/dev/null 2>&1
}
ns_resolves() { # ns_resolves CONTAINER_ID NAME
  docker_run_owned ns-resolves --rm --label g001.run=$RUN_ID --network "container:$1" "$BUSYBOX_IMAGE" nslookup "$2" >/dev/null 2>&1
}
container_ips() { # container_ips CONTAINER_ID -> all IPv4 addresses
  docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}} {{end}}' "$1"
}

yq_stdin() { docker_run_owned yq --rm --label g001.run=$RUN_ID -i "$YQ_IMAGE" "$@"; }

# --- web dependency policy (B-DEPS; G002 plan §5 and §15) ------------------------------------
# Approved direct dependencies of web, as category|name|exact version.
DEPS_EXISTING=(
  "dependencies|next|16.4.1" "dependencies|react|19.3.0" "dependencies|react-dom|19.3.0" "dependencies|server-only|0.0.1"
  "devDependencies|@tailwindcss/postcss|4.3.3" "devDependencies|@types/node|24.19.1" "devDependencies|@types/react|19.3.0"
  "devDependencies|@types/react-dom|19.3.0" "devDependencies|eslint|9.39.5" "devDependencies|eslint-config-next|16.4.1"
  "devDependencies|tailwindcss|4.3.3" "devDependencies|typescript|5.9.3")
DEPS_ADDITIONS=(
  "dependencies|@base-ui/react|1.9.0"
  "dependencies|class-variance-authority|0.7.1"
  "dependencies|lucide-react|1.52.0" "dependencies|tw-animate-css|1.4.0"
  # G069 (Plan r1 S2): the runtime packages the unaltered kit source imports (cn 0.4.0 removed: the kit's own
  # lib/utils.ts replaces it), and shadcn as a dev-only package for the kit's @import "shadcn/tailwind.css" (no CLI run).
  "dependencies|cmdk|1.1.1" "dependencies|clsx|2.1.1" "dependencies|tailwind-merge|3.7.0"
  "dependencies|next-themes|0.4.6" "dependencies|sonner|2.0.8" "dependencies|react-day-picker|9.14.0"
  "dependencies|embla-carousel-react|8.6.0" "dependencies|input-otp|1.5.0"
  "dependencies|react-resizable-panels|4.15.0" "dependencies|recharts|3.10.1"
  "devDependencies|@playwright/test|1.63.0" "devDependencies|@axe-core/playwright|4.13.0"
  "devDependencies|shadcn|4.22.1")
# G069 (Plan r1 S3, ADR-001 G069 amendment): the only @radix-ui packages allowed, at any lockfile depth, are cmdk's
# transitive closure (cmdk -> @radix-ui/react-dialog and its utilities), as exact name@version entries. No @radix-ui
# package is ever a direct dependency, and cmdk must be direct while this list is non-empty. WEB-SHELL.md records the
# reason per package.
# G069: exact list taken from the resolved lockfile (P0.1, cmdk 1.1.1; the only dependents of each entry are cmdk and
# these entries themselves): 16 packages, each present because cmdk's module graph imports @radix-ui/react-dialog.
DEPS_RADIX_VIA_CMDK=(
  "@radix-ui/primitive@1.2.0" "@radix-ui/react-compose-refs@1.1.5" "@radix-ui/react-context@1.3.0"
  "@radix-ui/react-dialog@1.3.1" "@radix-ui/react-dismissable-layer@1.2.1" "@radix-ui/react-focus-guards@1.2.0"
  "@radix-ui/react-focus-scope@1.3.1" "@radix-ui/react-id@1.1.4" "@radix-ui/react-portal@1.2.0"
  "@radix-ui/react-presence@1.1.11" "@radix-ui/react-primitive@2.2.0" "@radix-ui/react-slot@1.5.0"
  "@radix-ui/react-use-callback-ref@1.1.5" "@radix-ui/react-use-controllable-state@1.3.0"
  "@radix-ui/react-use-effect-event@0.0.5" "@radix-ui/react-use-layout-effect@1.1.4")
# G069 end of the replaceable list.
# Never a direct dependency: the Radix umbrella package (G069: shadcn is a reviewed dev-only addition above).
DEPS_FORBIDDEN=(radix-ui)

# deps_pair_check PKG LOCK: prints the sorted unique set of "CODE name" violations of a
# package.json/package-lock.json pair and fails when the set is non-empty. Every rule is
# evaluated independently; no rule suppresses another. A lock entry is dev-only when npm marks
# it "dev". npm marks a devDependency "devOptional" instead when a non-dev package also declares
# it as an optional peer (for example @playwright/test for next, @types/react for @base-ui/react);
# that marking is accepted only when such a declaration exists in the same lockfile.
deps_pair_check() {
  local pkg="$1" lock="$2" kind a b c d ent cat name ver other key spec want
  local -A LDEV=() LOPT=()
  local -A P=() R=() LV=() LD=() OPTPEER=() approved=() RX=() radix=()
  local out=()
  while IFS=$'\t' read -r cat name spec; do
    [[ -n "$cat" ]] && P["$cat|$name"]="$spec"
  done < <(yq_stdin -p json -o tsv '(.dependencies // {} | to_entries | .[] | ["dependencies", .key, .value]), (.devDependencies // {} | to_entries | .[] | ["devDependencies", .key, .value])' <"$pkg")
  while IFS=$'\t' read -r kind a b c d; do
    case "$kind" in
      root) R["$a|$b"]="$c" ;;
      lock) LV["$a"]="$b"; LDEV["$a"]="$c"; LOPT["$a"]="$d" ;;
      optpeer) OPTPEER["$a"]=1 ;;
      radix) RX["$a"]="$b" ;;
    esac
  done < <(yq_stdin -p json -o tsv '(.packages[""].dependencies // {} | to_entries | .[] | ["root", "dependencies", .key, .value]), (.packages[""].devDependencies // {} | to_entries | .[] | ["root", "devDependencies", .key, .value]), (.packages // {} | to_entries | .[] | select(.key | test("^node_modules/(@[^/]+/)?[^/]+$")) | ["lock", (.key | sub("^node_modules/"; "")), (.value.version // "-"), (.value.dev // false), (.value.devOptional // false)]), (.packages // {} | to_entries | .[] | select((.value.dev // false) == false and (.value.devOptional // false) == false) | (.value.peerDependenciesMeta // {}) | to_entries | .[] | select(.value.optional == true) | ["optpeer", .key]), (.packages // {} | to_entries | .[] | select(.key | test("(^|/)node_modules/@radix-ui/[^/]+$")) | ["radix", .key, (.value.version // "-")])' <"$lock")
  # G069 (Plan r1 §3.1): no package.json key under @radix-ui/; every lockfile packages key node_modules/@radix-ui/<x>,
  # at any depth, is an exact name@version entry of the reviewed cmdk closure; cmdk is direct while that list is non-empty.
  for ent in "${DEPS_RADIX_VIA_CMDK[@]}"; do radix["$ent"]=1; done
  for key in "${!P[@]}"; do name="${key#*|}"; [[ "$name" == "@radix-ui/"* ]] && out+=("DEPS_RADIX_DIRECT $name"); done
  for key in "${!RX[@]}"; do
    name="${key##*node_modules/}"
    [[ -n "${radix[$name@${RX[$key]}]:-}" ]] || out+=("DEPS_RADIX_UNEXPECTED $name")
  done
  [[ ${#DEPS_RADIX_VIA_CMDK[@]} -eq 0 || -n "${P[dependencies|cmdk]+x}" ]] || out+=("DEPS_RADIX_WITHOUT_CMDK cmdk")
  for name in "${!LV[@]}"; do
    LD["$name"]=runtime
    if [[ "${LDEV[$name]}" == true ]] || { [[ "${LOPT[$name]}" == true ]] && [[ -n "${OPTPEER[$name]:-}" ]]; }; then LD["$name"]=dev; fi
  done
  for ent in "${DEPS_EXISTING[@]}"; do
    IFS='|' read -r cat name ver <<<"$ent"; other="$(deps_other "$cat")"; approved["$name"]=1
    [[ -z "${P[$cat|$name]+x}" && -z "${P[$other|$name]+x}" ]] && out+=("DEPS_EXISTING_MISSING $name")
    [[ -n "${P[$other|$name]+x}" ]] && out+=("DEPS_EXISTING_RECATEGORIZED $name")
    [[ -n "${P[$cat|$name]+x}" && "${P[$cat|$name]}" != "$ver" ]] && out+=("DEPS_EXISTING_CHANGED $name")
  done
  for ent in "${DEPS_ADDITIONS[@]}"; do
    IFS='|' read -r cat name ver <<<"$ent"; other="$(deps_other "$cat")"; approved["$name"]=1
    [[ -z "${P[$cat|$name]+x}" || -n "${P[$other|$name]+x}" || "${P[$cat|$name]:-}" != "$ver" ]] && out+=("DEPS_ADDITION_WRONG $name")
  done
  for key in "${!P[@]}"; do
    cat="${key%%|*}"; name="${key#*|}"; spec="${P[$key]}"
    [[ -z "${approved[$name]:-}" ]] && out+=("DEPS_EXTRA_DIRECT $name")
    for ent in "${DEPS_FORBIDDEN[@]}"; do [[ "$name" == "$ent" ]] && out+=("DEPS_FORBIDDEN $name"); done
    if [[ "$spec" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
      [[ "${LV[$name]:-}" == "$spec" ]] || out+=("DEPS_LOCK_MISMATCH $name") # (b) locked version
    else
      out+=("DEPS_NOT_EXACT $name")
    fi
    want="$([[ "$cat" == devDependencies ]] && echo dev || echo runtime)"
    [[ -z "${LD[$name]:-}" || "${LD[$name]}" == "$want" ]] || out+=("DEPS_LOCK_MISMATCH $name") # (c) dev flag
  done
  for key in "${!P[@]}" "${!R[@]}"; do # (a) lockfile root specifier and category
    [[ "${P[$key]-<absent>}" == "${R[$key]-<absent>}" ]] || out+=("DEPS_LOCK_MISMATCH ${key#*|}")
  done
  [[ ${#out[@]} -eq 0 ]] && return 0
  printf '%s\n' "${out[@]}" | LC_ALL=C sort -u
  return 1
}
deps_other() { if [[ "$1" == dependencies ]]; then echo devDependencies; else echo dependencies; fi; }

# deps_repo_check REPO: deps_pair_check on web's real pair, plus the repository-level rules.
deps_repo_check() {
  local repo="$1" out pw
  # OS001 distributions cutover: the render-customization registration rule (DEPS_NOT_REGISTERED) is retired with the
  # generator; the pair, forbidden-import and image rules are unchanged.
  out="$( {
    deps_pair_check "$repo/web/package.json" "$repo/web/package-lock.json" || true
    if grep -rqF --exclude-dir=node_modules --exclude-dir=.next 'next/font' "$repo/web"; then echo "DEPS_FORBIDDEN next/font"; fi
    pw="$(yq_stdin -p json -o tsv '.devDependencies["@playwright/test"] // ""' <"$repo/web/package.json")"
    if ! grep -qE "^FROM mcr\.microsoft\.com/playwright:v${pw//./\\.}-noble@sha256:[0-9a-f]{64}( |$)" "$repo/infra/docker/web-browser.Dockerfile" 2>/dev/null || [[ -z "$pw" ]]; then
      echo DEPS_IMAGE_MISMATCH
    fi
  } | LC_ALL=C sort -u)"
  [[ -z "$out" ]] && return 0
  printf '%s\n' "$out"
  return 1
}

# --- adopted components and design inputs (B-WEB-COMPONENTS, B-WEB-DESIGN; G048 Plan §8.2, §9.1, §11.5) ---------
# G069 (Plan r1 S1, §3.2): web/src is exactly the 69 kit/src files of the adopted skill's LOCAL.sha256 (paths kit/src/<p>
# -> web/src/<p>), each with its manifest SHA-256; the Spec's source-manifest digest over them (files sorted by relative
# path, lines "<sha256>  <path>\n", hashed) is pinned here and bound both to those manifest rows and to web/src.
UI_KIT_FILES=69
UI_KIT_DIGEST=597f2aac251987be26050108dc6affe3a39a5594e7bfc7e2744de2fb5829a6ee
# data-slot values app-owned code (web/{app,components,lib}) may write besides the kit's own: the ButtonLink <a>.
UI_APP_SLOTS=(button)
# web/public: exactly these files, each with its intake SHA-256 (WEB-SHELL.md records the upstream blob SHA-1s;
# G067: Geist and Geist Mono from google/fonts 9e25e2ba265e5298f70f6182dd4e8a3ebf1b9123, Roboto removed).
WEB_PUBLIC_PINS=(
  "web/public/fonts/oswald/oswald-wght.ttf 5b38c246e255a12f5712d640d56bcced0472466fc68983d2d0410ec0457c2817"
  "web/public/fonts/oswald/OFL.txt 0fd731a904b729a4e02eaf5e8ebd06783edd9abe400e8882760160230675b652"
  "web/public/fonts/barlow-condensed/barlow-condensed-400.ttf 583cec5da3b84bc4dc7c9c72e2a565c94d34e431518b19d7e250b7830ad5f996"
  "web/public/fonts/barlow-condensed/barlow-condensed-600.ttf 7b619d14bc2327509a9ef32b0890f709626f7ecc9ff61191c2a4314c5499d2d9"
  "web/public/fonts/barlow-condensed/OFL.txt 186d750eb496a4c17a76385f82be6aea2ac1cf2de074a811d63786cf374ea73f"
  "web/public/fonts/geist/geist-wght.ttf 73894e0448cae90a92b6c2f8732b7bb9acb7b94c418bff559dad4a18e1de9659"
  "web/public/fonts/geist/OFL.txt 1781d2806a07d91c4edf4740b88449fab7d0eadad53f7c351b94cd4d4eb8c00f"
  "web/public/fonts/geist-mono/geist-mono-wght.ttf d00e590b8eb3a59acc329b2d044fd143ae935090b7da33199ebee27cc7de8196"
  "web/public/fonts/geist-mono/OFL.txt 1781d2806a07d91c4edf4740b88449fab7d0eadad53f7c351b94cd4d4eb8c00f"
  "web/public/blueprint/overview.svg 942a2a242594773f6be7e638d6e07f9f723a0ff6b6c6e8334427702e52f98c65")
# The url() targets of web CSS, sorted (the five bundled faces and the drawing).
WEB_CSS_URLS="/blueprint/overview.svg /fonts/barlow-condensed/barlow-condensed-400.ttf /fonts/barlow-condensed/barlow-condensed-600.ttf /fonts/geist-mono/geist-mono-wght.ttf /fonts/geist/geist-wght.ttf /fonts/oswald/oswald-wght.ttf "
# The drawing's allowed vocabulary (§8.1).
SVG_ELEMENTS="svg g path line polyline polygon rect circle"
SVG_ATTRIBUTES="xmlns viewBox fill stroke stroke-width stroke-opacity stroke-dasharray stroke-linecap stroke-linejoin id d x1 y1 x2 y2 points x y width height cx cy r transform"

# --- browser suite (G002 plan §11) ----------------------------------------------------------
# compose_browser DEST args...  (local.yml plus the browser overlay, as README platform:browser)
compose_browser() {
  local dest="$1" rc=0; shift
  compose_call "$dest" infra/compose/local.yml infra/compose/browser.yml -- "$@" || rc=$?
  capture_after "$dest" "$@"
  return "$rc"
}

# browser_collect_report NAME REPORT_DIR LOG: collect only this run's newly created browser one-off by immutable ID.
# A failed Compose run/name collision must not copy from or remove an unrelated container.
browser_collect_report() {
  local name="$1" rep="$2" out="$3" cid owner service rc=0
  read -r cid owner service < <(docker inspect -f '{{.Id}} {{index .Config.Labels "g001.run"}} {{index .Config.Labels "com.docker.compose.service"}}' "$name" 2>>"$out") || return 1
  if [[ "$owner" != "$RUN_ID" || "$service" != web-browser ]] || id_at_start container "$cid"; then
    log "browser report container ownership not established; leaving $name untouched"; return 1
  fi
  docker cp "$cid:/tmp/report/." "$rep" >>"$out" 2>&1 || { log "browser report copy failed"; rc=1; }
  docker rm -f -v "$cid" >>"$out" 2>&1 || { log "browser container removal failed"; rc=1; }
  return "$rc"
}

# browser_run DEST REPORT_DIR LOG [playwright args...]: retain the named one-off just long enough to copy its
# internal reports out with host ownership, then remove it. Keep the image's non-root user and worker count.
browser_run() {
  local dest="$1" rep="$2" out="$3" rc=0 runc; shift 3
  runc="g001-browser-$RUN_ID-$(random_hex 4)"
  mkdir -p "$rep"
  compose_browser "$dest" --profile browser run --name "$runc" --label "g001.run=$RUN_ID" --build --no-deps \
    -e PLAYWRIGHT_JSON_OUTPUT_FILE=/tmp/report/results.json -e PLAYWRIGHT_OUTPUT_DIR=/tmp/report/out \
    -e PLAYWRIGHT_WORKERS="${HARNESS_PLAYWRIGHT_WORKERS:-2}" web-browser "$@" >"$out" 2>&1 || rc=$?
  browser_collect_report "$runc" "$rep" "$out" || rc=1
  return "$rc"
}

# Host ports of the disposable feature-stack project (G051 §3.4; web core intelligence postgres redis minio-api minio-console).
PORTS_FEAT=(3700 8780 8700 6132 7079 9700 9701)

# Required scenarios per project (§11.2), as "projects|scenario IDs". S-ERRORS is carried by every
# test of a project: each must pass and carry the fixture's S-ERRORS annotation.
BROWSER_MATRIX_FULL=(
  "shell-light shell-dark|S-IDENTITY S-LANDMARKS S-NAV S-SKIP S-TAB S-OVERLAY S-MOTION S-ERRORS S-RESPONSIVE S-A11Y S-CONTRAST S-BOUNDARY S-DESIGN S-ANGLE S-SCREENSHOTS S-PALETTE S-TEXT-SPACING S-FIDELITY S-FONTS S-BLUEPRINT S-LONG-IDENTITY S-ACCOUNT-ANON S-ACCOUNT-UNAVAILABLE S-ACCOUNT-LINK S-ACCOUNT-A11Y S-ACCOUNT-MCP-UNAVAILABLE"
  "shell-light|S-HEAD-ORDER S-HEALTH S-NOFIXTURE S-FAILCLOSED"
  "tokens-light tokens-dark|S-ERRORS S-A11Y S-CONTRAST S-TOKENS S-PORTAL S-RUNTIME S-ANGLE S-SCREENSHOTS S-STATES S-TEXT-SPACING S-PORTAL-INIT S-FONTS S-BLUEPRINT"
  "appearance|S-ERRORS S-APPEARANCE-DEFAULT S-APPEARANCE-LIVE S-APPEARANCE-OVERRIDE S-APPEARANCE-RETURN S-APPEARANCE-PERSIST S-APPEARANCE-INVALID S-APPEARANCE-NOSTORAGE S-FIRSTPAINT S-FIRSTPAINT-DELAYED S-HTML-ATTRS"
  "dev-hydration|S-HYDRATION-DEV")
BROWSER_MATRIX_IDENTITY=("shell-light shell-dark|S-IDENTITY")
# G004 (B-IDENTITY-E2E): the L2 identity scenarios in both identity projects of the disposable HTTPS environment, and
# (G014) the MCP connection setup scenarios in both mcp projects there, and (G018) the onboarding scenarios with the
# member-page landmark and navigation checks in both onboarding projects, (G019) the membership scenarios in both
# membership projects and (G058) the enrollment scenarios in both enrollment projects. G060 (Plan r0 §5.2 T11; r1 C6;
# correction02 C02-4): E-ONB-8 (read-only onboarding for Organization staff), E-MEM-8 (the owner-only Group-
# administration setting), E-MEM-9 (override administration and its revocation) and, by the owner's invitation-message
# amendment, E-MEM-10 (one generic invalid-invitation presentation; the intended recipient still accepts). G062 (Plan
# r5 SUP-15, §11): the root-host tenant signup scenarios E-SUP-1..6 in both signup projects (base https://localhost:543).
# G020 (Plan r2 §5.2 T12, D-10): the label scenarios E-LAB-1..6 with the Organization labels page's landmark and
# navigation checks in both labels projects. G069 r3 (owner finding): S-SELECT-KIT, the members page's role dropdown as
# the kit Select, in both membership projects (the two appearances).
BROWSER_MATRIX_IDENTITY_E2E=(
  "identity-light identity-dark|E-LOGIN A-1 E-ORIGINS E-ATTR E-FIX E-CB-1 E-CB-2 E-CB-3 E-CB-4 E-CB-5 E-CB-6 E-CB-7 A-7 E-CB-8 E-CB-10 E-PKCE E-CT E-CSRF E-LO-1 E-LO-2 E-LO-3 E-LO-4 E-LO-6 E-LO-7 E-LO-8 E-CACHE E-A11Y A-3 E-NET-2 E-PRIVACY E-HANDOFF S-ERRORS"
  "mcp-light mcp-dark|M-NOAUTH M-NONMEMBER M-SETUP M-LIST M-A11Y M-HANDOFF M-RENAME M-ACCOUNT S-ERRORS"
  "onboarding-light onboarding-dark|E-ONB-1 E-ONB-2 E-ONB-3 E-ONB-4 E-ONB-5 E-ONB-6 E-ONB-7 E-ONB-8 S-LANDMARKS S-NAV S-ERRORS"
  "membership-light membership-dark|E-MEM-1 E-MEM-2 E-MEM-3 E-MEM-4 E-MEM-5 E-MEM-6 E-MEM-7 E-MEM-8 E-MEM-9 E-MEM-10 S-SELECT-KIT S-LANDMARKS S-NAV S-ERRORS"
  "enrollment-light enrollment-dark|E-ENR-1 E-ENR-2 E-ENR-3 E-ENR-4 E-ENR-5 S-ERRORS"
  "signup-light signup-dark|E-SUP-1 E-SUP-2 E-SUP-3 E-SUP-4 E-SUP-5 E-SUP-6 S-ERRORS"
  "labels-light labels-dark|E-LAB-1 E-LAB-2 E-LAB-3 E-LAB-4 E-LAB-5 E-LAB-6 S-LANDMARKS S-NAV S-ERRORS")

# browser_matrix JSON ROWS_OUT MATRIX...: every required scenario passes in each required project;
# no test is skipped, flaky or unexpected. Writes project/status/title/annotations rows.
browser_matrix() {
  local json="$1" rows="$2" ok=1 entry projects ids pr id n bad stats; shift 2
  [[ -s "$json" ]] || { log "no JSON report at $json"; return 1; }
  stats="$(yq_stdin -p json -o tsv '[.stats.expected, .stats.skipped, .stats.flaky, .stats.unexpected] | join(" ")' <"$json")"
  [[ "$stats" =~ ^[1-9][0-9]*\ 0\ 0\ 0$ ]] || { ok=0; log "browser stats (expected skipped flaky unexpected): $stats"; }
  yq_stdin -p json -o tsv '.. | select(tag == "!!map" and has("specs")) | .specs[] | .title as $t | .tests[] | [.projectName, .status, $t, ([.annotations[]?.type, .results[]?.annotations[]?.type] | unique | join(","))]' <"$json" >"$rows"
  bad="$(awk -F'\t' '$2 != "expected"' "$rows")"
  [[ -z "$bad" ]] || { ok=0; log "tests not passing: ${bad//$'\n'/; }"; }
  for entry in "$@"; do
    projects="${entry%%|*}"; ids="${entry#*|}"
    for pr in $projects; do
      for id in $ids; do
        if [[ "$id" == S-ERRORS ]]; then
          n="$(awk -F'\t' -v p="$pr" '$1==p' "$rows" | wc -l)"
          bad="$(awk -F'\t' -v p="$pr" '$1==p && ($2!="expected" || ("," $4 ",") !~ /,S-ERRORS,/)' "$rows" | wc -l)"
          [[ "$n" -gt 0 && "$bad" == 0 ]] || { ok=0; log "S-ERRORS in $pr: $n tests, $bad without a passing S-ERRORS annotation"; }
        else
          n="$(awk -F'\t' -v p="$pr" -v t="[$id]" '$1==p && $2=="expected" && index($3, t)' "$rows" | wc -l)"
          [[ "$n" -gt 0 ]] || { ok=0; log "$id did not pass in $pr"; }
        fi
      done
    done
  done
  return $((1 - ok))
}

# init_hash_documented JSON DOC: every S-HEAD-ORDER hash attachment (the served appearance-init
# script's SHA-256) equals the hash documented in WEB-SHELL.md.
init_hash_documented() {
  local doc seen
  doc="$(sed -nE 's/^.*`appearance-init` SHA-256: `([0-9a-f]{64})`.*$/\1/p' "$2" | head -1)"
  seen="$(yq_stdin -p json -o tsv '.. | select(tag == "!!map" and has("name") and .name == "appearance-init.sha256") | .body' <"$1" | while IFS= read -r b; do base64 -d <<<"$b"; echo; done | LC_ALL=C sort -u)"
  [[ -n "$doc" && "$seen" == "$doc" ]] || { log "appearance-init hash: served {${seen//$'\n'/ }}, documented '${doc}'"; return 1; }
}

# chromium_version DEST: the browser version inside the pinned runner image.
chromium_version() {
  compose_browser "$1" --profile browser run --rm --no-deps web-browser \
    node -e 'require("playwright").chromium.launch().then(async (b) => { console.log(b.version()); await b.close(); })' 2>/dev/null | tail -1
}

# psql_super DEST SQL: superuser catalog query over the PostgreSQL
# container's local socket (validation only; no password on any command line).
psql_super() {
  compose "$1" exec -T postgres psql -X -q -tA -v ON_ERROR_STOP=1 -U postgres -d "$(env_get "$1" DATABASE_NAME)" -c "$2"
}

# --- G064 test-database lifecycle (Plan r2 §3.1-§3.3) --------------------------------------------------
# A run-private, write-closed template <db>_tpl is copied once from the verified canonical database; before every
# independent Go invocation db_fresh drops and recreates <db> from it under the canonical name, so role derivation,
# current_database() in migration SQL and baked role literals stay production-equivalent. Template and copies never
# leave the run's own PostgreSQL container/volume. Every create/drop/refusal is a row of $WORK/db-lifecycle.tsv,
# exported to REPORT_DIR after each row. Lifecycle statements run only on the maintenance database "postgres";
# identifiers are psql :"var"-quoted after a name check. No session is ever terminated: a client session on <db> is
# a refusal (plain DROP DATABASE, never FORCE). Callers test the status (helper || ...); DB_LC_ERR names the cause.
DB_LC_NAME_RE='^[a-z_][a-z0-9_]{0,47}$'        # the runner's database-name rule (core/internal/migrate/runner.go)
DB_LC_DERIVED_RE='^[a-z_][a-z0-9_]{0,62}$'     # derived names (<db>_tpl, control targets): the 63-byte identifier limit
DB_TPL_REQUIRE_ROWS="${DB_TPL_REQUIRE_ROWS:-}" # migrations lifecycle only: results.tsv IDs that must already be PASS
DB_LC_ERR="" DB_LC_CALLER="${DB_LC_CALLER:-}"

# _psql_on DEST DBNAME SQL [VAR=VALUE...]: superuser psql on DBNAME over the container's local socket. SQL goes on
# stdin (statement by statement, so CREATE/DROP DATABASE never run inside a transaction block); values are psql
# variables (:'v' literal, :"v" identifier). No password on any command line.
_psql_on() {
  local dest="$1" dbn="$2" sql="$3" a vars=(); shift 3
  for a in "$@"; do vars+=(-v "$a"); done
  printf '%s\n' "$sql" | compose "$dest" exec -T postgres psql -X -q -tA -v ON_ERROR_STOP=1 -U postgres -d "$dbn" "${vars[@]}" -f -
}
# psql_maint DEST SQL [VAR=VALUE...]: psql_super on the instance's maintenance database "postgres", the only connection
# used for CREATE/DROP/ALTER DATABASE and the shared-catalog queries.
psql_maint() { local dest="$1" sql="$2"; shift 2; _psql_on "$dest" postgres "$sql" "$@"; }

_db_ms() { date +%s%3N; }
_db_rec() { awk -F'\t' -v k="$2" '$1 == k { sub(/^[^\t]*\t/, ""); print }' "$1"; } # RECORD FIELD -> value line(s)
_db_ledger_head() { [[ -s "$WORK/db-lifecycle.tsv" ]] || printf 'ts_utc\tkind\tsubject\tfield\tvalue\n' >"$WORK/db-lifecycle.tsv"; }
# _db_ledger KIND SUBJECT FIELD VALUE: one ledger row (VALUE is the rest of the line and may hold tabs), then export.
_db_ledger() {
  _db_ledger_head
  printf '%s\t%s\t%s\t%s\t%s\n' "$(date -u +%Y-%m-%dT%H:%M:%S.%3NZ)" "$1" "$2" "$3" "${4:-}" >>"$WORK/db-lifecycle.tsv"
  db_lifecycle_export
}
# _db_ledger_file KIND SUBJECT FIELD FILE: one row per line of FILE (verbatim), one export.
_db_ledger_file() {
  local ts line
  _db_ledger_head; ts="$(date -u +%Y-%m-%dT%H:%M:%S.%3NZ)"
  while IFS= read -r line || [[ -n "$line" ]]; do printf '%s\t%s\t%s\t%s\t%s\n' "$ts" "$1" "$2" "$3" "$line"; done <"$4" >>"$WORK/db-lifecycle.tsv"
  db_lifecycle_export
}
# db_lifecycle_export: copy the ledger to REPORT_DIR atomically after every row, so an interrupted run keeps it.
db_lifecycle_export() {
  [[ -n "${REPORT_DIR:-}" && -f "$WORK/db-lifecycle.tsv" ]] || return 0
  local tmp="$REPORT_DIR/.db-lifecycle.tsv.$BASHPID"
  cp -- "$WORK/db-lifecycle.tsv" "$tmp" && mv -f -- "$tmp" "$REPORT_DIR/db-lifecycle.tsv"
}
# _db_refuse HELPER SUBJECT REASON: set DB_LC_ERR and record a refuse row (status 0; the caller returns 1).
_db_refuse() {
  DB_LC_ERR="$1: $3"
  _db_ledger refuse "$2" "$1" "${DB_LC_CALLER:+caller=$DB_LC_CALLER }$3" || true
}
_db_oneline() { tr '\n' ' ' <<<"$1" | sed 's/ *$//'; }
# _db_delta A B: "-line" for lines only in A, "+line" for lines only in B (C collation, as sorted).
_db_delta() {
  LC_ALL=C comm -23 <(printf '%s\n' "$1" | LC_ALL=C sort) <(printf '%s\n' "$2" | LC_ALL=C sort) | sed 's/^/-/'
  LC_ALL=C comm -13 <(printf '%s\n' "$1" | LC_ALL=C sort) <(printf '%s\n' "$2" | LC_ALL=C sort) | sed 's/^/+/'
}

# _db_bound TREE: the bound migration set of the copied tree (derived, never pinned): DB_BOUND_FILES (sorted file
# names), DB_BOUND (their versions, space-separated) and DB_MIG_TOP (the highest).
_db_bound() {
  local f b n prev=-1
  DB_BOUND_FILES=() DB_BOUND="" DB_MIG_TOP=""
  for f in "$1"/core/migrations/*.sql; do
    [[ -f "$f" ]] || continue
    b="${f##*/}"
    [[ "$b" =~ ^([0-9]{5})_[A-Za-z0-9_]+[.]sql$ ]] || { DB_LC_ERR="unexpected migration file name $b"; return 1; }
    n=$((10#${BASH_REMATCH[1]}))
    ((n > prev)) || { DB_LC_ERR="duplicate migration version $n"; return 1; }
    DB_BOUND_FILES+=("$b"); DB_BOUND+="${DB_BOUND:+ }$n"; DB_MIG_TOP=$n; prev=$n
  done
  [[ -n "$DB_MIG_TOP" ]] || { DB_LC_ERR="no migration under $1/core/migrations"; return 1; }
}
# _db_source TREE: DB_SRC_MIG (sha256 over the sorted migration files and names), DB_SRC_ROLES / DB_SRC_PROV (the
# provisioning scripts) and DB_SRC_TREE (impl_tree_hash of the copied tree the images were built from).
_db_source() {
  DB_SRC_MIG="$(cd "$1/core/migrations" && sha256sum -- *.sql | sha256sum | awk '{print $1}')" || { DB_LC_ERR="migration digest"; return 1; }
  DB_SRC_ROLES="$(sha_of "$1/infra/postgres/init/10-roles.sh")" || { DB_LC_ERR="10-roles.sh digest"; return 1; }
  DB_SRC_PROV="$(sha_of "$1/infra/postgres/provision-roles.sql")" || { DB_LC_ERR="provision-roles.sql digest"; return 1; }
  # impl_tree_hash exits non-zero for the tracked directory symlinks (sha256sum skips them, as for B-SOURCE-HASH); the
  # binding is its deterministic digest, so only a well-formed digest is required.
  DB_SRC_TREE="$(impl_tree_hash "$1" 2>/dev/null)" || true
  [[ "$DB_SRC_TREE" =~ ^[0-9a-f]{64}$ ]] || { DB_LC_ERR="tree hash"; return 1; }
}

# _db_goose_check DEST DBNAME SCHEMA TABLE: goose's own history against the bound set (run _db_bound first). The
# applied set is the version_id of the latest row (highest id) per version whose is_applied is true; it must equal
# {0} ∪ bound (goose inserts the version 0 row at table creation; a superseded down/reapply row, deleted or marked
# is_applied=false, never counts), so its maximum is DB_MIG_TOP. No row is counted or pinned.
_db_goose_check() {
  local got n lc_fails=()
  local -A have=()
  got="$(_psql_on "$1" "$2" 'SELECT coalesce(string_agg(version_id::text, '"' '"' ORDER BY version_id), '"''"') FROM (SELECT DISTINCT ON (version_id) version_id, is_applied FROM :"s".:"t" ORDER BY version_id, id DESC) l WHERE is_applied;' s="$3" t="$4" 2>&1)" ||
    { DB_LC_ERR="goose history unreadable ($3.$4): $(_db_oneline "$got")"; return 1; }
  DB_GOOSE_APPLIED="$got"
  for n in $got; do have[$n]=1; done
  [[ -n "${have[0]:-}" ]] || lc_fails+=("version 0 row missing or not applied")
  unset 'have[0]'
  for n in $DB_BOUND; do
    if [[ -n "${have[$n]:-}" ]]; then unset "have[$n]"; else lc_fails+=("bound version $n not applied (latest row)"); fi
  done
  for n in "${!have[@]}"; do lc_fails+=("unexpected applied version $n"); done
  ((${#lc_fails[@]} == 0)) || { DB_LC_ERR="goose history ($3.$4): $(IFS=';'; printf '%s' "${lc_fails[*]}")"; return 1; }
}
# _db_history DEST DBNAME: the raw goose history (id, version_id, is_applied, tstamp by id), recorded verbatim.
_db_history() { _psql_on "$1" "$2" "SELECT format(E'%s\\t%s\\t%s\\t%s', id, version_id, is_applied, tstamp) FROM app.goose_db_version ORDER BY id;"; }
# _db_baseline DEST DBNAME SCHEMAS EXCLUDE: "schema.relation<TAB>rows" for every row-bearing relation (relkind r, p,
# m: ordinary tables, partitioned parents, materialized views) in the comma-separated SCHEMAS except EXCLUDE (or -);
# sequences, indexes, views, TOAST and composite relations are not row-bearing evidence. Superuser: RLS bypassed.
_db_baseline() {
  _psql_on "$1" "$2" "SELECT format(E'%s.%s\\t%s', n.nspname, c.relname,
  (xpath('/row/c/text()', query_to_xml(format('SELECT count(*) AS c FROM %I.%I', n.nspname, c.relname), false, true, '')))[1]::text)
FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
WHERE n.nspname = ANY (string_to_array(:'schemas', ',')) AND c.relkind IN ('r', 'p', 'm')
  AND format('%s.%s', n.nspname, c.relname) <> :'exclude';" schemas="$3" exclude="$4" | LC_ALL=C sort
}
# _db_fingerprint DEST DBNAME TREE: "catalog<TAB>md5" and "roles<TAB>md5" (infra/recovery/sql/fingerprint-catalog.sql).
_db_fingerprint() { _psql_on "$1" "$2" "$(cat "$3/infra/recovery/sql/fingerprint-catalog.sql")" db="$2"; }
# _db_sessions DEST DBNAME: "count<TAB>inventory" of client sessions on DBNAME (autovacuum is not a session).
_db_sessions() {
  psql_maint "$1" "SELECT format(E'%s\\t%s', count(*), coalesce(string_agg(format('pid=%s user=%s app=%s state=%s', pid, usename, application_name, state), '; ' ORDER BY pid), '-'))
FROM pg_stat_activity WHERE datname = :'d' AND backend_type = 'client backend' AND pid <> pg_backend_pid();" d="$2"
}
# _db_dbacl DEST DBNAME: "owner<TAB>non-owner ACL items<TAB>database-qualified pg_db_role_setting rows".
_db_dbacl() {
  psql_maint "$1" "SELECT format(E'%s\\t%s\\t%s', pg_get_userbyid(d.datdba),
  coalesce((SELECT string_agg(v, ',' ORDER BY v COLLATE \"C\") FROM (SELECT format('%s=%s/%s', CASE a.grantee WHEN 0 THEN 'public' ELSE pg_get_userbyid(a.grantee)::text END, a.privilege_type, a.is_grantable) AS v
    FROM aclexplode(coalesce(d.datacl, acldefault('d', d.datdba))) a WHERE a.grantee <> d.datdba) q), '-'),
  (SELECT count(*) FROM pg_db_role_setting s WHERE s.setdatabase = d.oid))
FROM pg_database d WHERE d.datname = :'d';" d="$2"
}
_db_oid() { psql_maint "$1" "SELECT oid FROM pg_database WHERE datname = :'d';" d="$2"; }

# db_shared_catalog DEST [TEMPLATE]: the cluster-global and database-OID-bound state that cloning never copies, as one
# sorted C-collated line set: every role of the <db>_ family and postgres with its attributes; every membership whose
# member OR role is in the family (counterparties outside it by name); every pg_db_role_setting row whose role is in
# the family or whose database is <db> or the template (a row whose database no longer exists prints ?); the
# owner/flags/ACL of <db> and the template.
db_shared_catalog() {
  local db; db="$(env_get "$1" DATABASE_NAME)"
  psql_maint "$1" "WITH fam AS (SELECT oid FROM pg_roles WHERE left(rolname::text, length(:'db') + 1) = :'db' || '_'),
dbs AS (SELECT oid FROM pg_database WHERE datname IN (:'db', :'tpl'))
SELECT line FROM (
  SELECT format('role:%s:%s,%s,%s,%s,%s,%s,%s,%s,%s', r.rolname, r.rolcanlogin, r.rolinherit, r.rolsuper, r.rolbypassrls,
         r.rolcreaterole, r.rolcreatedb, r.rolreplication, r.rolconnlimit, coalesce(r.rolvaliduntil::text, '-')) AS line
    FROM pg_roles r WHERE r.oid IN (SELECT oid FROM fam) OR r.rolname = 'postgres'
  UNION ALL
  SELECT format('member:%s>%s:%s,%s,%s', pg_get_userbyid(m.member), pg_get_userbyid(m.roleid), m.inherit_option, m.set_option, m.admin_option)
    FROM pg_auth_members m WHERE m.member IN (SELECT oid FROM fam) OR m.roleid IN (SELECT oid FROM fam)
  UNION ALL
  SELECT format('setting:%s:%s:%s',
         CASE s.setrole WHEN 0 THEN '-' ELSE coalesce((SELECT rolname::text FROM pg_roles WHERE oid = s.setrole), '?') END,
         CASE s.setdatabase WHEN 0 THEN '-' ELSE coalesce((SELECT datname::text FROM pg_database WHERE oid = s.setdatabase), '?') END,
         array_to_string(s.setconfig, ','))
    FROM pg_db_role_setting s
   WHERE s.setrole IN (SELECT oid FROM fam) OR s.setdatabase IN (SELECT oid FROM dbs)
      OR (s.setdatabase <> 0 AND NOT EXISTS (SELECT 1 FROM pg_database d WHERE d.oid = s.setdatabase))
  UNION ALL
  SELECT format('db:%s:%s:%s,%s:%s', d.datname, pg_get_userbyid(d.datdba), d.datallowconn, d.datistemplate,
         coalesce((SELECT string_agg(v, ',' ORDER BY v COLLATE \"C\") FROM (SELECT format('%s=%s/%s/%s',
           CASE a.grantee WHEN 0 THEN 'public' ELSE pg_get_userbyid(a.grantee)::text END, a.privilege_type, pg_get_userbyid(a.grantor), a.is_grantable) AS v
           FROM aclexplode(coalesce(d.datacl, acldefault('d', d.datdba))) a) q), '-'))
    FROM pg_database d WHERE d.oid IN (SELECT oid FROM dbs)
) x;" db="$db" tpl="${2:-${db}_tpl}" | LC_ALL=C sort
}

# db_pristine_verify DEST TREE STATUS_LOG: the source-bound pristine check every template caller runs itself (Plan r2
# §3.1). STATUS_LOG is the caller's own `migrate status` run of the bound migrate image. Checks, all reported:
# (a) goose history: one "migration<TAB><file><TAB>applied<TAB><time>" line per bound file in order, no other
# migration line, nothing pending, current_version = MIG_TOP; in SQL the latest-row applied set = {0} ∪ bound; the raw
# history is recorded, never counted; (b) catalog_p / catalog_roles (run.sh) with MIG_RELATIONS; (c) the catalog
# fingerprint (recorded); (d) every row-bearing app/audit relation except app.goose_db_version empty; (e) no
# <db>-qualified pg_db_role_setting row and the database ACL of 10-roles.sh (owner postgres, PUBLIC revoked, CONNECT
# for <db>_migrator, <db>_runtime, <db>_retention); (f) no client session on <db>. Nothing is created or dropped.
# Output: a pristine row and $WORK/db-lifecycle/pristine-<db>.facts for db_template_create.
db_pristine_verify() {
  local dest="$1" tree="$2" slog="$3" db got want facts hist fp base acl sess lc_fails=() caller
  DB_LC_ERR=""
  db="$(env_get "$dest" DATABASE_NAME)"; caller="${DB_LC_CALLER:-${dest##*/}}"
  [[ "$db" =~ $DB_LC_NAME_RE ]] || { _db_refuse db_pristine_verify "${db:-?}" "invalid database name"; return 1; }
  _db_bound "$tree" || { _db_refuse db_pristine_verify "$db" "$DB_LC_ERR"; return 1; }
  [[ -s "$slog" ]] || { _db_refuse db_pristine_verify "$db" "status listing $slog missing or empty"; return 1; }
  [[ "${MIG_RELATIONS:-}" =~ ^[0-9]+$ ]] || { _db_refuse db_pristine_verify "$db" "MIG_RELATIONS unset"; return 1; }
  mkdir -p "$WORK/db-lifecycle"; facts="$WORK/db-lifecycle/pristine-$db.facts"; : >"$facts"
  # (a) the caller's own status listing, then the SQL counterpart.
  got="$(awk -F'\t' -v files="${DB_BOUND_FILES[*]}" -v top="$DB_MIG_TOP" '
    BEGIN { n = split(files, want, " ") }
    $1 == "migration" { m++
      if (m > n) lc_bad = lc_bad "; extra migration line " $2
      else if ($2 != want[m] || $3 != "applied" || $4 == "") lc_bad = lc_bad "; line " m ": " $2 " " $3 " (want " want[m] " applied with a time)" }
    $1 == "current_version" { cv++; if ($2 != top) lc_bad = lc_bad "; current_version " $2 " (want " top ")" }
    /pending/ { lc_bad = lc_bad "; pending line" }
    END { if (m < n) lc_bad = lc_bad "; " m + 0 " migration lines for " n " bound files"
          if (cv != 1) lc_bad = lc_bad "; " cv + 0 " current_version lines"
          print substr(lc_bad, 3) }' "$slog")"
  [[ -z "$got" ]] || lc_fails+=("status listing: $got")
  awk -F'\t' '$1 == "migration" || $1 == "current_version" { print "status\t" $0 }' "$slog" >>"$facts"
  _db_goose_check "$dest" "$db" app goose_db_version || lc_fails+=("$DB_LC_ERR")
  if hist="$(_db_history "$dest" "$db" 2>&1)"; then
    sed 's/^/history\t/' <<<"$hist" >>"$facts"
    printf 'history_digest\t%s\n' "$(sha256sum <<<"$hist" | awk '{print $1}')" >>"$facts"
  else lc_fails+=("goose history listing: $(_db_oneline "$hist")"); fi
  printf 'bound_digest\t%s\n' "$(printf '%s\n' "${DB_BOUND_FILES[@]}" | sha256sum | awk '{print $1}')" >>"$facts"
  # (b) metadata protection and role facts (the existing run.sh readers and the hoisted relation pin).
  if declare -F catalog_p >/dev/null && declare -F catalog_roles >/dev/null; then
    got="$(catalog_p "$dest" 2>&1)" || true
    [[ "$got" == "objects=2 owners=${db}_owner extra_acl=0 table_privs=0 seq_privs=0" ]] || lc_fails+=("catalog_p: $got")
    got="$(catalog_roles "$dest" 2>&1)" || true
    want="app_owner=${db}_owner runtime_super=f runtime_bypassrls=f runtime_member_owner=f runtime_member_migrator=f runtime_owned=0 migrator_grant=inherit=f,set=t app_objects_beyond_metadata=$MIG_RELATIONS"
    [[ "$got" == "$want" ]] || lc_fails+=("catalog_roles: $got")
  else lc_fails+=("catalog_p/catalog_roles unavailable"); fi
  # (c) catalog fingerprint, recorded for the template record and the db_fresh postcondition.
  if fp="$(_db_fingerprint "$dest" "$db" "$tree" 2>&1)" && [[ "$fp" == *catalog$'\t'* && "$fp" == *roles$'\t'* ]]; then
    awk -F'\t' '$1 == "catalog" { print "fingerprint_catalog\t" $2 } $1 == "roles" { print "fingerprint_roles\t" $2 }' <<<"$fp" >>"$facts"
  else lc_fails+=("catalog fingerprint: $(_db_oneline "$fp")"); fi
  # (d) empty fixture baseline over the precise row-bearing table set.
  if base="$(_db_baseline "$dest" "$db" app,audit app.goose_db_version 2>&1)" && [[ -n "$base" ]]; then
    sed 's/^/baseline\t/' <<<"$base" >>"$facts"
    printf 'baseline_digest\t%s\n' "$(sha256sum <<<"$base" | awk '{print $1}')" >>"$facts"
    got="$(awk -F'\t' '$2 != "0" { printf "%s%s=%s", sep, $1, $2; sep = "," }' <<<"$base")"
    [[ -z "$got" ]] || lc_fails+=("baseline rows not empty: $got")
  else lc_fails+=("baseline enumeration: $(_db_oneline "$base")"); fi
  # (e) database-level state.
  acl="$(_db_dbacl "$dest" "$db" 2>&1)" || true
  want="postgres"$'\t'"$(printf '%s\n' "${db}_migrator=CONNECT/f" "${db}_runtime=CONNECT/f" "${db}_retention=CONNECT/f" | LC_ALL=C sort | paste -sd,)"$'\t0'
  [[ "$acl" == "$want" ]] || lc_fails+=("database owner/ACL/settings: $(_db_oneline "$acl") (want $(_db_oneline "$want"))")
  # (f) sessions.
  sess="$(_db_sessions "$dest" "$db" 2>&1)" || true
  [[ "${sess%%$'\t'*}" == 0 ]] || lc_fails+=("client sessions on $db: $(_db_oneline "$sess")")
  printf 'mig_top\t%s\nbound\t%s\n' "$DB_MIG_TOP" "$DB_BOUND" >>"$facts"
  if ((${#lc_fails[@]})); then
    DB_LC_ERR="db_pristine_verify: $(IFS='|'; printf '%s' "${lc_fails[*]}")"
    _db_ledger pristine "$db" FAIL "caller=$caller mig_top=$DB_MIG_TOP ${DB_LC_ERR#db_pristine_verify: }" || true
    return 1
  fi
  _db_ledger pristine "$db" PASS "caller=$caller mig_top=$DB_MIG_TOP bound=${#DB_BOUND_FILES[@]} $(awk -F'\t' '$1 ~ /^(bound_digest|history_digest|fingerprint_catalog|fingerprint_roles|baseline_digest)$/ { printf "%s=%s ", $1, $2 }' "$facts")tables=$(grep -c . <<<"$base")"
}

# db_template_create DEST PROJECT TREE STATUS_LOG [NAME]: record the source binding, require db_pristine_verify (and,
# in the migrations lifecycle only, the DB_TPL_REQUIRE_ROWS PASS rows), then copy <db> once into the write-closed
# template NAME (default <db>_tpl: IS_TEMPLATE true, ALLOW_CONNECTIONS false, CONNECT revoked) and write its record
# $WORK/db-lifecycle/NAME.rec (also as template-record ledger rows). Any failure: no template.
db_template_create() {
  local dest="$1" project="$2" tree="$3" slog="$4" db name t0 t1 head state cimg mimg out id rec snap size facts why porcelain
  DB_LC_ERR=""
  db="$(env_get "$dest" DATABASE_NAME)"
  [[ "$db" =~ $DB_LC_NAME_RE ]] || { _db_refuse db_template_create "${db:-?}" "invalid database name"; return 1; }
  name="${5:-${db}_tpl}"
  [[ "$name" =~ $DB_LC_DERIVED_RE && "$name" != "$db" && "$name" != postgres ]] || { _db_refuse db_template_create "$name" "invalid template name"; return 1; }
  t0="$(_db_ms)"
  # Source binding, recorded first: derived migration set, digests, tested tree, HEAD and images (never --base).
  { _db_bound "$tree" && _db_source "$tree"; } || { _db_refuse db_template_create "$name" "source binding: $DB_LC_ERR"; return 1; }
  head="$(git -C "$REPO" rev-parse HEAD 2>/dev/null)" || head=-
  if porcelain="$(git -C "$REPO" status --porcelain -- . ":(top,exclude).agents/skills/restruct-sdlc" ":(top,exclude).claude/skills/restruct-sdlc" 2>/dev/null)"; then [[ -z "$porcelain" ]] && state="head" || state="dirty"; else state="unknown"; fi
  cimg="$(docker image inspect -f '{{.Id}}' "$project-core-check" 2>/dev/null)" || cimg=-
  mimg="$(docker image inspect -f '{{.Id}}' "$project-migrate" 2>/dev/null)" || mimg=-
  for id in $DB_TPL_REQUIRE_ROWS; do
    awk -F'\t' -v i="$id" '$1 == i && $2 == "PASS" { f = 1 } END { exit !f }' "$RESULTS" 2>/dev/null ||
      { _db_refuse db_template_create "$name" "$id is not recorded PASS in this run"; return 1; }
  done
  if ! DB_LC_CALLER="${DB_LC_CALLER:-$project}" db_pristine_verify "$dest" "$tree" "$slog"; then
    why="$DB_LC_ERR"; _db_refuse db_template_create "$name" "pristine check failed: ${why#db_pristine_verify: }"; return 1
  fi
  facts="$WORK/db-lifecycle/pristine-$db.facts"
  out="$(psql_maint "$dest" "SELECT count(*) FROM pg_database WHERE datname = :'n';" n="$name" 2>&1)" || out="?"
  [[ "$out" == 0 ]] || { _db_refuse db_template_create "$name" "a database named $name already exists ($out)"; return 1; }
  if ! out="$(psql_maint "$dest" 'CREATE DATABASE :"n" TEMPLATE :"d";
ALTER DATABASE :"n" WITH IS_TEMPLATE true ALLOW_CONNECTIONS false;
REVOKE ALL ON DATABASE :"n" FROM PUBLIC, :"m", :"r", :"x";' n="$name" d="$db" m="${db}_migrator" r="${db}_runtime" x="${db}_retention" 2>&1)"; then
    why="$(_db_oneline "$out")"
    # Only the database this call just created is removed (never one it did not create).
    if [[ "$why" != *"already exists"* ]]; then psql_maint "$dest" 'ALTER DATABASE :"n" WITH IS_TEMPLATE false; DROP DATABASE IF EXISTS :"n";' n="$name" >/dev/null 2>&1 || true; fi
    _db_refuse db_template_create "$name" "create failed: $why"; return 1
  fi
  out="$(psql_maint "$dest" "SELECT format('%s,%s', datistemplate, datallowconn) FROM pg_database WHERE datname = :'n';" n="$name" 2>&1)" || true
  [[ "$out" == "t,f" ]] || { _db_refuse db_template_create "$name" "template flags $(_db_oneline "$out") (want istemplate t, allowconn f)"; return 1; }
  snap="$(db_shared_catalog "$dest" "$name" 2>&1)" || { _db_refuse db_template_create "$name" "shared catalog: $(_db_oneline "$snap")"; return 1; }
  size="$(psql_maint "$dest" "SELECT pg_database_size(:'n');" n="$name" 2>&1)" || size=-
  t1="$(_db_ms)"
  rec="$WORK/db-lifecycle/$name.rec"
  {
    printf 'name\t%s\ndatabase\t%s\nproject\t%s\nflags\tdatistemplate=t datallowconn=f connect=revoked\n' "$name" "$db" "$project"
    printf 'mig_top\t%s\nbound\t%s\nmig_digest\t%s\nroles_sh_digest\t%s\nprovision_sql_digest\t%s\n' "$DB_MIG_TOP" "$DB_BOUND" "$DB_SRC_MIG" "$DB_SRC_ROLES" "$DB_SRC_PROV"
    printf 'tree_path\t%s\ntree_hash\t%s\nhead\t%s\ntree_state\t%s\nbase\t%s\n' "$tree" "$DB_SRC_TREE" "$head" "$state" "${BASE:--}"
    printf 'core_check_image\t%s\nmigrate_image\t%s\n' "$cimg" "$mimg"
    grep -E $'^(bound_digest|history_digest|fingerprint_catalog|fingerprint_roles|baseline_digest|status|history|baseline)\t' "$facts"
    sed 's/^/shared\t/' <<<"$snap"
    printf 'size_bytes\t%s\nduration_ms\t%s\n' "$size" "$((t1 - t0))"
  } >"$rec"
  _db_ledger template "$name" create "status=PASS source=$db mig_top=$DB_MIG_TOP tree=$DB_SRC_TREE head=$head state=$state size_bytes=$size duration_ms=$((t1 - t0))"
  _db_ledger_file template-record "$name" record "$rec"
}

# db_fresh DEST PROJECT PURPOSE [RECORD]: drop <db> and recreate it from the template under the canonical name, then
# restore the database-level grants cloning does not copy (exactly 10-roles.sh). Preconditions (any failure: nothing
# dropped): the template exists closed; the record's migration/provisioning digests, migration set and tested tree
# hash equal the bound tree now (stale/source mismatch); the shared catalog equals the snapshot (a role, membership,
# setting or template delta refuses; a <db>-level ACL/setting delta is residue removed by the recreation, recorded);
# no client session on <db>. Postconditions: the goose-history predicate and history listing, the catalog
# fingerprint, the shared catalog = snapshot and an empty baseline over the recorded table set. One fresh row with
# PG-010 durations. RECORD defaults to the <db>_tpl record (T-1b passes an altered copy).
db_fresh() {
  local dest="$1" project="$2" purpose="$3" db rec name tree f v want got snap cur line t0 t1 t2 t3 t4 sess out fp
  local residue=() lc_bad=() lc_post=()
  DB_LC_ERR=""
  db="$(env_get "$dest" DATABASE_NAME)"
  [[ "$db" =~ $DB_LC_NAME_RE ]] || { _db_refuse db_fresh "${db:-?}" "invalid database name"; return 1; }
  rec="${4:-$WORK/db-lifecycle/${db}_tpl.rec}"
  [[ -s "$rec" ]] || { _db_refuse db_fresh "$db" "purpose=$purpose no template record $rec"; return 1; }
  name="$(_db_rec "$rec" name)"; tree="$(_db_rec "$rec" tree_path)"
  [[ "$name" =~ $DB_LC_DERIVED_RE && "$(_db_rec "$rec" database)" == "$db" && -d "$tree" ]] ||
    { _db_refuse db_fresh "$db" "purpose=$purpose template record does not belong to $db"; return 1; }
  t0="$(_db_ms)"
  out="$(psql_maint "$dest" "SELECT format('%s,%s', datistemplate, datallowconn) FROM pg_database WHERE datname = :'n';" n="$name" 2>&1)" || true
  [[ "$out" == "t,f" ]] || { _db_refuse db_fresh "$db" "purpose=$purpose template $name missing or not closed ($(_db_oneline "$out"))"; return 1; }
  { _db_bound "$tree" && _db_source "$tree"; } || { _db_refuse db_fresh "$db" "purpose=$purpose source binding: $DB_LC_ERR"; return 1; }
  for f in mig_top:DB_MIG_TOP bound:DB_BOUND mig_digest:DB_SRC_MIG roles_sh_digest:DB_SRC_ROLES provision_sql_digest:DB_SRC_PROV tree_hash:DB_SRC_TREE; do
    v="${f#*:}"; want="$(_db_rec "$rec" "${f%%:*}")"
    [[ "$want" == "${!v}" ]] || lc_bad+=("${f%%:*} record=${want:-none} now=${!v}")
  done
  ((${#lc_bad[@]} == 0)) || { _db_refuse db_fresh "$db" "purpose=$purpose stale or source-mismatched template record: ${lc_bad[*]}"; return 1; }
  snap="$(_db_rec "$rec" shared)"
  cur="$(db_shared_catalog "$dest" "$name" 2>&1)" || { _db_refuse db_fresh "$db" "purpose=$purpose shared catalog unreadable: $(_db_oneline "$cur")"; return 1; }
  while IFS= read -r line; do
    [[ -n "$line" ]] || continue
    case "${line:1}" in
      "db:$db:"* | "setting:-:$db:"*) residue+=("$line") ;;
      *) lc_bad+=("$line") ;;
    esac
  done < <(_db_delta "$snap" "$cur")
  ((${#lc_bad[@]} == 0)) || { _db_refuse db_fresh "$db" "purpose=$purpose shared catalog delta (roles/memberships/settings/template): ${lc_bad[*]}"; return 1; }
  sess="$(_db_sessions "$dest" "$db" 2>&1)" || sess="?"$'\t'"$(_db_oneline "$sess")"
  [[ "${sess%%$'\t'*}" == 0 ]] || { _db_refuse db_fresh "$db" "purpose=$purpose client session(s) on $db, nothing dropped or terminated: $(_db_oneline "$sess")"; return 1; }
  t1="$(_db_ms)"
  out="$(psql_maint "$dest" 'DROP DATABASE :"d";' d="$db" 2>&1)" || { _db_refuse db_fresh "$db" "purpose=$purpose drop refused: $(_db_oneline "$out")"; return 1; }
  t2="$(_db_ms)"
  out="$(psql_maint "$dest" 'CREATE DATABASE :"d" TEMPLATE :"n";
REVOKE ALL ON DATABASE :"d" FROM PUBLIC;
GRANT CONNECT ON DATABASE :"d" TO :"m", :"r", :"x";' d="$db" n="$name" m="${db}_migrator" r="${db}_runtime" x="${db}_retention" 2>&1)" ||
    { _db_refuse db_fresh "$db" "purpose=$purpose recreate after drop failed: $(_db_oneline "$out")"; return 1; }
  t3="$(_db_ms)"
  for line in "${residue[@]}"; do _db_ledger residue "$db" "residue-removed-by-recreate" "purpose=$purpose $line"; done
  _db_goose_check "$dest" "$db" app goose_db_version || lc_post+=("$DB_LC_ERR")
  got="$(_db_history "$dest" "$db" 2>&1)" || true
  [[ "$got" == "$(_db_rec "$rec" history)" ]] || lc_post+=("goose history listing differs from the template record")
  fp="$(_db_fingerprint "$dest" "$db" "$tree" 2>&1)" || true
  [[ "$(awk -F'\t' '$1 == "catalog" { c = $2 } $1 == "roles" { r = $2 } END { print c " " r }' <<<"$fp")" == "$(_db_rec "$rec" fingerprint_catalog) $(_db_rec "$rec" fingerprint_roles)" ]] ||
    lc_post+=("catalog fingerprint differs from the template record")
  cur="$(db_shared_catalog "$dest" "$name" 2>&1)" || true
  [[ "$cur" == "$snap" ]] || lc_post+=("shared catalog after recreation: $(_db_oneline "$(_db_delta "$snap" "$cur")")")
  got="$(_db_baseline "$dest" "$db" app,audit app.goose_db_version 2>&1)" || true
  [[ "$(cut -f1 <<<"$got")" == "$(_db_rec "$rec" baseline | cut -f1)" ]] || lc_post+=("baseline table set differs from the template record")
  v="$(awk -F'\t' '$2 != "0" { printf "%s%s=%s", sep, $1, $2; sep = "," }' <<<"$got")"
  [[ -z "$v" ]] || lc_post+=("baseline rows not empty: $v")
  t4="$(_db_ms)"
  if ((${#lc_post[@]})); then
    DB_LC_ERR="db_fresh: purpose=$purpose postcondition failed: $(IFS='|'; printf '%s' "${lc_post[*]}")"
    _db_ledger fresh "$db" FAIL "${DB_LC_CALLER:+caller=$DB_LC_CALLER }purpose=$purpose template=$name ${DB_LC_ERR#db_fresh: } total_ms=$((t4 - t0))" || true
    return 1
  fi
  _db_ledger fresh "$db" PASS "${DB_LC_CALLER:+caller=$DB_LC_CALLER }purpose=$purpose template=$name check_ms=$((t1 - t0)) drop_ms=$((t2 - t1)) create_ms=$((t3 - t2)) post_ms=$((t4 - t3)) total_ms=$((t4 - t0)) residue_removed=${#residue[@]}"
}

# --- background suites (G051 Plan §3.6, C3) ------------------------------------------------------
# A child is a subshell in the harness process group with private RESULTS, TIMINGS and log under
# $WORK/bg/NAME/; only bg_join merges its rows. bg_stop_all (EXIT trap, before cleanup) stops every
# unjoined child and its descendants, found by a bounded /proc PPID walk and verified by start time;
# their rows are never merged. Nothing is matched by process name or pattern.
BG_GRACE_S="${BG_GRACE_S:-60}" BG_FORCE_FAIL=0 BG_NAMES=()
declare -gA BG_PID=() BG_PSTART=() BG_T0=() BG_DONE=() _BG_K=() _BG_R=()

# _bg_stat PID: _BG_S (state), _BG_PP (parent) and _BG_ST (start time, /proc/PID/stat field 22); fails when gone.
_bg_stat() {
  local s f
  { read -r s <"/proc/$1/stat"; } 2>/dev/null || return 1
  s="${s##*) }"
  read -r -a f <<<"$s"
  [[ -n "${f[19]:-}" ]] || return 1
  _BG_S="${f[0]}" _BG_PP="${f[1]}" _BG_ST="${f[19]}"
}
_bg_live() { _bg_stat "$1" && [[ "$_BG_S" != Z && "$_BG_ST" == "$2" ]]; } # PID START: still that live process
_bg_own() { [[ -n "$2" ]] && _bg_stat "$1" && [[ "$_BG_ST" == "$2" ]]; } # PID START: still that process (or its zombie)

# G064 correction 11: bash forgets an ended child's status once a later child of this shell gets the same PID (its wait
# then returns 127, or the newer child's status). Each child therefore records its own exit status in rc (_bg_rcw, its
# EXIT trap: rc.tmp, then mv), and _bg_status waits for a child's PID only while that PID names no other process.
# _bg_rcw STATUS: records FUNC's return status or a non-zero exit (exit N, errexit). bash also runs this trap, with a
# stale $? (often 0), when TERM, HUP or INT ends the child, so a 0 short of FUNC's return is not recorded; KILL records
# nothing. An unrecorded status is never read as success.
_bg_rcw() {
  local rc="$1"
  if [[ -n "${_BG_END:-}" ]]; then rc="$_BG_END"; elif [[ "$rc" == 0 ]]; then return 0; fi
  { printf '%s\n' "$rc" >"$_BG_RCDIR/rc.tmp" && mv -f -- "$_BG_RCDIR/rc.tmp" "$_BG_RCDIR/rc"; } 2>/dev/null || :
}

# _bg_status NAME: NAME's exit status in _BG_RC; fails with the reason in _BG_WHY (an invalid child state). NAME's PID is
# waited for while it is still the process bg_start started (same start time) or gone (no process holds it, so the wait
# cannot block); a PID that names another process is never waited for. When bash cannot attribute a status (127) the
# child's record decides; exit 0 counts only when the record says 0 too. Any other status that differs from an existing
# record (the saved status of a later child that reused the PID) is an invalid child state naming both values.
_bg_status() {
  local name="$1" pid="${BG_PID[$1]}" f="$WORK/bg/$1/rc" how w=0 rec="" re=$'^(0|[1-9][0-9]{0,2})\n$'
  _BG_RC="" _BG_WHY=""
  if _bg_own "$pid" "${BG_PSTART[$name]:-}"; then
    wait "$pid" 2>/dev/null || w=$?; how="wait for its running process (pid $pid) returned 127"
  elif _bg_stat "$pid"; then
    w=127 how="pid $pid now names another process"
  else
    wait "$pid" 2>/dev/null || w=$?; how="pid $pid is no longer a child of this shell"
  fi
  if [[ -e "$f" || -L "$f" ]]; then
    if [[ -f "$f" && ! -L "$f" ]]; then { IFS= read -r -N 5 rec <"$f"; } 2>/dev/null || :; fi
    if [[ "$rec" =~ $re ]] && (( ${rec%$'\n'} <= 255 )); then rec="${rec%$'\n'}"; else _BG_WHY="$name: invalid recorded exit status"; return 1; fi
  fi
  if (( w == 127 )); then
    if [[ -z "$rec" ]]; then _BG_WHY="$name: exit status unknown ($how) and none recorded"; return 1; fi
    w="$rec"
  elif (( w == 0 )) && [[ "$rec" != 0 ]]; then
    _BG_WHY="$name: wait reported exit 0 but the child recorded ${rec:+exit $rec}${rec:-no exit status}"; return 1
  elif [[ -n "$rec" ]] && (( w != rec )); then
    _BG_WHY="$name: wait reported exit $w but the child recorded exit $rec"; return 1
  fi
  _BG_RC="$w"
}

# _bg_walk: adds to _BG_K (PID -> start time) every live descendant of a live member whose start time is
# unchanged (one /proc pass, then to a fixpoint), with its child's name in _BG_R. The harness shell is never added.
_bg_walk() {
  local d p ch=1 n=0
  local -A bgPP=() bgST=()
  for d in /proc/[0-9]*; do
    p="${d#/proc/}"
    _bg_stat "$p" || continue
    [[ "$_BG_S" != Z ]] || continue
    bgPP[$p]="$_BG_PP" bgST[$p]="$_BG_ST"
  done
  while (( ch && n < 64 )); do
    ch=0; n=$((n + 1))
    for p in "${!bgPP[@]}"; do
      [[ -z "${_BG_K[$p]:-}" && "$p" != "$$" && "$p" != "$BASHPID" ]] || continue
      [[ -n "${_BG_K[${bgPP[$p]}]:-}" && "${bgST[${bgPP[$p]}]:-}" == "${_BG_K[${bgPP[$p]}]}" ]] || continue
      _BG_K[$p]="${bgST[$p]}" _BG_R[$p]="${_BG_R[${bgPP[$p]}]:-}"; ch=1
    done
  done
}
_bg_signal() { local p; for p in "${!_BG_K[@]}"; do _bg_live "$p" "${_BG_K[$p]}" && kill -s "$1" "$p" 2>/dev/null; done; return 0; }
_bg_survivors() { local p out=""; for p in "${!_BG_K[@]}"; do _bg_live "$p" "${_BG_K[$p]}" && out+="$p "; done; printf '%s' "${out% }"; }

# bg_start NAME FUNC [ARGS...]: FUNC ARGS... as a supervised background child. A child that cannot start is a
# BACKGROUND FAIL row (returns 0, like check). FUNC succeeds by returning 0: an `exit 0` inside it leaves no exit
# record (_bg_rcw cannot tell it from a TERM), so its join fails.
bg_start() {
  local name="$1" func="$2" dir; shift 2
  if [[ ! "$name" =~ ^[a-z][a-z0-9-]{0,31}$ || -n "${BG_PID[$name]:-}" ]] || ! declare -F "$func" >/dev/null; then
    record BACKGROUND FAIL "cannot start background child '$name' ($func)"; return 0
  fi
  dir="$WORK/bg/$name"
  mkdir -p "$dir" && rm -f -- "$dir/rc" "$dir/rc.tmp" && : >"$dir/results.tsv" && : >"$dir/timings.tsv" && : >"$dir/log" ||
    { record BACKGROUND FAIL "$name: no private outputs"; return 0; }
  (
    _BG_END="" _BG_RCDIR="$dir"
    trap '_bg_rcw "$?"' EXIT
    RESULTS="$dir/results.tsv" TIMINGS="$dir/timings.tsv" BG_CHILD="$name"
    exec >"$dir/log" 2>&1
    "$func" "$@"
    _BG_END=$?
    exit "$_BG_END"
  ) &
  BG_PID[$name]=$!
  BG_PSTART[$name]=""
  if _bg_stat "${BG_PID[$name]}"; then BG_PSTART[$name]="$_BG_ST"; fi
  BG_T0[$name]="$(date +%s)"
  BG_NAMES+=("$name")
  log "background $name started (pid ${BG_PID[$name]})"
}

# bg_join NAME SUITE: waits for NAME (_bg_status), then merges its well-formed rows and its timings (verbatim) and log.
# BACKGROUND INFO when it exited 0 and every row is a required ID of SUITE, each once; otherwise BACKGROUND FAIL (an
# invalid child state: the _bg_status reason).
bg_join() {
  local name="$1" suite="$2" dir rc=0 bad=0 line id s re why=""
  local -A bgreq=() bgseen=()
  if [[ -z "${BG_PID[$name]:-}" || -n "${BG_DONE[$name]:-}" ]]; then record BACKGROUND FAIL "$name was not running"; return 0; fi
  dir="$WORK/bg/$name"
  if _bg_status "$name"; then rc="$_BG_RC"; else why="$_BG_WHY"; fi
  BG_DONE[$name]=1
  while IFS=$'\t' read -r id s; do
    [[ -n "$id" && "$id" != \#* && "$s" == "$suite" ]] && bgreq[$id]=1
  done <"${BG_REQUIRED_FILE:-$REPO/infra/validation/required-checks.tsv}"
  re=$'^([A-Z][A-Z0-9-]*)\t(PASS|FAIL|UNMET|INFO)\t'
  while IFS= read -r line || [[ -n "$line" ]]; do
    if [[ "$line" =~ $re ]]; then
      id="${BASH_REMATCH[1]}"
      if [[ -z "${bgreq[$id]:-}" || -n "${bgseen[$id]:-}" ]]; then bad=1; fi
      bgseen[$id]=1
      printf '%s\n' "$line" >>"$RESULTS"
    else
      bad=1
    fi
  done <"$dir/results.tsv"
  cat -- "$dir/timings.tsv" >>"${TIMINGS:-$WORK/timings.tsv}"
  log "== background $name log"
  cat -- "$dir/log"
  if [[ -n "$why" ]]; then
    if (( bad )); then record BACKGROUND FAIL "$why; foreign or malformed rows"; else record BACKGROUND FAIL "$why"; fi
  elif (( rc != 0 && bad )); then
    record BACKGROUND FAIL "$name exited $rc; foreign or malformed rows"
  elif (( rc != 0 )); then
    record BACKGROUND FAIL "$name exited $rc"
  elif (( bad )); then
    record BACKGROUND FAIL "$name foreign or malformed rows"
  else
    record BACKGROUND INFO "$name joined (exit 0, $(( $(date +%s) - BG_T0[$name] )) s)"
  fi
  return 0
}

# bg_wait_optional NAME: waits for a prebuild child (_bg_status); a failure is only noted (the authoritative build
# follows), an invalid child state is BACKGROUND FAIL. Its rows are never merged.
bg_wait_optional() {
  local name="$1" rc=0 why=""
  if [[ -z "${BG_PID[$name]:-}" || -n "${BG_DONE[$name]:-}" ]]; then return 0; fi
  if _bg_status "$name"; then rc="$_BG_RC"; else why="$_BG_WHY"; fi
  BG_DONE[$name]=1
  log "== background $name log"
  cat -- "$WORK/bg/$name/log" 2>/dev/null
  if [[ -n "$why" ]]; then record BACKGROUND FAIL "$why"
  elif (( rc != 0 )); then record PREBUILD INFO "$name exit $rc; authoritative build follows"; fi
  return 0
}

# bg_stop_all: every unjoined child together: TERM it and its verified descendants, poll up to BG_GRACE_S (re-walking
# for newcomers, at most 3 times), KILL survivors (at most 3 rounds, re-walking each), reap the children and verify
# that none is alive. Records BACKGROUND FAIL per child; never merges rows; always returns 0 (it runs before cleanup).
bg_stop_all() {
  local pending=$? name bgnames=() end n p left
  local -
  set +e
  _BG_K=() _BG_R=()
  for name in "${BG_NAMES[@]}"; do
    [[ -z "${BG_DONE[$name]:-}" ]] || continue
    BG_DONE[$name]=1; bgnames+=("$name")
    if [[ -n "${BG_PSTART[$name]}" ]] && _bg_live "${BG_PID[$name]}" "${BG_PSTART[$name]}"; then
      _BG_K[${BG_PID[$name]}]="${BG_PSTART[$name]}" _BG_R[${BG_PID[$name]}]="$name"
    fi
  done
  (( ${#bgnames[@]} )) || return 0
  if (( ${#_BG_K[@]} )); then
    _bg_walk; _bg_signal TERM
    end=$((SECONDS + BG_GRACE_S)); n=0
    while [[ -n "$(_bg_survivors)" ]] && (( SECONDS < end )); do
      sleep 0.5
      if (( n < 3 )); then _bg_walk; _bg_signal TERM; n=$((n + 1)); fi
    done
    for n in 1 2 3; do
      [[ -n "$(_bg_survivors)" ]] || break
      _bg_walk; _bg_signal KILL; sleep 0.2
    done
  fi
  # Reap only a PID that is still the child's: a reused PID may name another (even a live) child of this shell.
  for name in "${bgnames[@]}"; do if _bg_own "${BG_PID[$name]}" "${BG_PSTART[$name]}"; then wait "${BG_PID[$name]}" 2>/dev/null; fi; done
  for name in "${bgnames[@]}"; do
    left=""
    for p in $(_bg_survivors); do [[ "${_BG_R[$p]:-}" != "$name" ]] || left+="$p "; done
    record BACKGROUND FAIL "$name stopped before join"
    [[ -z "$left" ]] || record BACKGROUND FAIL "$name survived KILL (pids ${left% })"
  done
  if (( pending == 0 )); then BG_FORCE_FAIL=1; fi
  return 0
}

# --- build ledger (G051 Plan §3.8, C4) -------------------------------------------------------------
# One builds.tsv row per build invocation in scope (explicit harness builds, cache-only prebuilds, Compose
# calls and command blocks that can build, the hosted bootstrap), built only from BuildKit plain progress,
# the Dockerfile text, `docker compose config`, image IDs and the build inputs. The hooks print nothing:
# wrapped calls return their stdout, stderr and status unchanged. Anything not determinable is `unknown`
# and stays in the denominator (never an exclusion); unknown rows make AC5 UNMET.
LEDGER_DIR="" LEDGER_DAEMON=unknown LEDGER_PLATFORM=unknown
LEDGER_SUITE="${LEDGER_SUITE:-}" LEDGER_KIND="${LEDGER_KIND:-authoritative}"
# The reviewed named exceptions (§3.8.3): a row of kind named:<ID> may repeat or nearly repeat another key.
LEDGER_EXCEPTIONS="prebuild-fallback B-BUILD-CONTEXT B-WEB-ERROR-OUTPUT web-browser-fixture B-TOGGLE-OFF"
LEDGER_HEADER=$'seq\tjob\tsuite\tkind\tdockerfile\tdockerfile_sha256\ttarget\tfrontend\tbases\tplatform\tignore_file\tignore_sha256\tcontext_inventory\tnamed_contexts\targs_digest\tdaemon\tprogress\tsteps_executed\tsteps_cached\tresult\timage_id'
LEDGER_SUMMARY_KEYS="expected rows unknown duplicates near_duplicates exceptions bootstrap ac5_status"
LG_CALL="" LG_JOB=unknown LG_SUITEV=unknown LG_KINDV=unknown LG_LINES=0 LG_U=() LG_F=() LG_SNAPBAD=0 LG_WHY=""
# Shell syntax a command-block line must not contain to be split into words safely.
LG_SHELL_RE='["'"'"'`$\;&|<>(){}]'
declare -gA _LG_INVC=() LG_DEF=() LG_DEFBAD=() LG_IMG=() LG_LE=() LG_LW=() LG_PRED=() LG_PROJS=()
LG_BLDNAMES="" LG_BLDUSE="" LG_BDRV="" LG_BWHY=""

# Shared awk: references as BuildKit prints them (docker.io/library/ and :latest explicit), whitespace squashing.
LEDGER_AWK_LIB='
function lg_normref(r,   name, dig, p, first, last) {
  dig = ""; p = index(r, "@")
  if (p > 0) { dig = substr(r, p); name = substr(r, 1, p - 1) } else name = r
  p = index(name, "/")
  if (p == 0) name = "docker.io/library/" name
  else { first = substr(name, 1, p - 1); if (first !~ /[.:]/ && first != "localhost") name = "docker.io/" name }
  if (name ~ /^docker\.io\/[^\/]+$/) name = "docker.io/library/" substr(name, 11)
  last = name; sub(/^.*\//, "", last)
  if (index(last, ":") == 0 && dig == "") name = name ":latest"
  return name dig
}
function lg_squash(s) { gsub(/[ \t]+/, " ", s); sub(/^ /, "", s); sub(/ $/, "", s); return s }
function lg_nodig(r) { sub(/@sha256:[0-9a-f]+$/, "", r); return r }
function lg_digof(r,   d) { if (match(r, /@sha256:[0-9a-f]+$/)) { d = substr(r, RSTART + 8); if (length(d) == 64) return d } return "" }
'

# Dockerfile model (-v target, -v named=" NAME... "): "syntax REF", then for the stages reachable from the
# target "step STAGE I N KIND TEXT" (FROM of an external image, RUN, COPY, ADD, WORKDIR, numbered as BuildKit
# numbers them), "base STAGE REF from|copy|mount" and "platform STAGE VALUE"; "error REASON" for anything
# this model does not cover (heredocs, ONBUILD, variables in FROM, other escape characters, ...).
LEDGER_DF_AWK='
BEGIN { OFS = "\t"; dirs = 1; ns = 0; buf = ""; cont = 0; err = ""; syntax = "" }
function fail(m) { if (err == "") err = m }
function stage_of(x, lim,   k, lx) {
  lx = tolower(x)
  for (k = lim; k >= 1; k--) if (sname[k] != "" && sname[k] == lx) return k
  if (x ~ /^[0-9]+$/ && x + 1 <= lim) return x + 1
  return 0
}
function ref_dep(x, how,   d) {
  if (x ~ /\$/) { fail("variable in a --from reference"); return }
  d = stage_of(x, ns - 1)
  if (d > 0) { deps[ns] = deps[ns] " " d; return }
  if (index(named, " " x " ") > 0) return
  nb[ns]++; bref[ns, nb[ns]] = lg_normref(x); bhow[ns, nb[ns]] = how
}
function instr(s,   kw, args, n, t, i, ref, nm, pl, x, m, j, kv) {
  s = lg_squash(s); if (s == "") return
  kw = s; sub(/ .*$/, "", kw); args = substr(s, length(kw) + 2); kw = toupper(kw)
  if (kw == "FROM") {
    n = split(args, t, " "); i = 1; pl = ""
    while (i <= n && substr(t[i], 1, 2) == "--") { if (substr(t[i], 1, 11) == "--platform=") pl = substr(t[i], 12); else fail("FROM option " t[i]); i++ }
    ref = t[i]; nm = ""
    if (i + 2 == n && toupper(t[i + 1]) == "AS") nm = tolower(t[i + 2])
    else if (i != n) fail("FROM syntax")
    x = stage_of(ref, ns)
    ns++; sname[ns] = nm; nst[ns] = 0; sbase[ns] = ""; nb[ns] = 0
    if (pl != "") splat[ns] = pl
    if (ref == "") { fail("FROM without an image"); return }
    if (x > 0) { deps[ns] = deps[ns] " " x; return }
    if (tolower(ref) == "scratch") return
    if (ref ~ /\$/) { fail("variable in FROM"); return }
    sbase[ns] = lg_normref(ref); nst[ns]++; stk[ns, nst[ns]] = "FROM"; stx[ns, nst[ns]] = sbase[ns]
    return
  }
  if (kw ~ /^(ENV|ARG|LABEL|EXPOSE|USER|CMD|ENTRYPOINT|HEALTHCHECK|SHELL|STOPSIGNAL|VOLUME|MAINTAINER)$/) { if (ns == 0 && kw != "ARG") fail(kw " before FROM"); return }
  if (kw !~ /^(RUN|COPY|ADD|WORKDIR)$/) { fail("instruction " kw); return }
  if (ns == 0) { fail(kw " before FROM"); return }
  if (index(args, "<<") > 0) fail("heredoc")
  n = split(args, t, " ")
  for (i = 1; i <= n && substr(t[i], 1, 2) == "--"; i++) {
    if ((kw == "COPY" || kw == "ADD") && substr(t[i], 1, 7) == "--from=") ref_dep(substr(t[i], 8), "copy")
    if (kw == "RUN" && substr(t[i], 1, 8) == "--mount=") { m = split(substr(t[i], 9), kv, ","); for (j = 1; j <= m; j++) if (substr(kv[j], 1, 5) == "from=") ref_dep(substr(kv[j], 6), "mount") }
  }
  nst[ns]++; stk[ns, nst[ns]] = kw; stx[ns, nst[ns]] = kw " " args
}
{
  line = $0; sub(/\r$/, "", line)
  if (dirs) {
    if (line ~ /^#[ \t]*[A-Za-z]+[ \t]*=/) {
      k = line; sub(/^#[ \t]*/, "", k); v = k; sub(/[ \t]*=.*$/, "", k); sub(/^[^=]*=[ \t]*/, "", v); sub(/[ \t]+$/, "", v)
      k = tolower(k)
      if (k == "syntax") syntax = v; else if (k == "escape" && v != "\\") fail("escape directive")
      next
    }
    dirs = 0
  }
  if (line ~ /^[ \t]*#/ || line ~ /^[ \t]*$/) next
  if (line ~ /\\[ \t]*$/) { sub(/\\[ \t]*$/, "", line); buf = buf line; cont = 1; next }
  buf = buf line; cont = 0; instr(buf); buf = ""
}
END {
  if (cont && buf != "") instr(buf)
  if (ns == 0) fail("no stage")
  t = 0
  if (target == "") t = ns
  else { for (k = 1; k <= ns; k++) if (sname[k] == tolower(target)) t = k; if (t == 0) fail("target " target " not found") }
  if (syntax != "") print "syntax", lg_normref(syntax)
  if (err != "") { print "error", err; exit }
  reach[t] = 1; q[1] = t; qh = 1; qt = 1
  while (qh <= qt) { k = q[qh++]; n = split(deps[k], d, " "); for (i = 1; i <= n; i++) if (!(d[i] in reach)) { reach[d[i]] = 1; q[++qt] = d[i] } }
  for (k = 1; k <= ns; k++) {
    if (!(k in reach)) continue
    lab = (sname[k] != "") ? sname[k] : ((ns > 1) ? "stage-" (k - 1) : "stage")
    for (j = 1; j <= nst[k]; j++) print "step", lab, j, nst[k], stk[k, j], stx[k, j]
    if (sbase[k] != "") print "base", lab, sbase[k], "from"
    for (j = 1; j <= nb[k]; j++) print "base", lab, bref[k, j], bhow[k, j]
    if (k in splat) print "platform", lab, splat[k]
  }
}'

# Progress table (-v mode=explicit|compose, -v soff=N) of one plain-progress stream: "S SESSION BUILDER DRIVER
# header" per "#0 building with" header; per print block of a vertex "V SESSION N SERVICE LABELTYPE STAGE I N TEXT";
# per terminal line "T SESSION N DONE|CACHED|ERROR|CANCELED"; per image name "N SESSION N NAME"; and
# "X SESSION REASON" for anything outside the format (session 0: the whole stream).
# G051 C3, Compose Bake progress (captured from the pinned hosted toolchain, Compose v5.5.1 / buildx v0.37.2): each
# invocation's stream starts with the vertex "[internal] load local bake definitions" and prints no "#0 building
# with" header. In compose mode that vertex opens a session "S SESSION - unknown bake" (its builder comes only from
# the call's verified builder evidence, never from the stream) unless it is the first vertex after a header; either
# way the session gets "Q SESSION". Bake prefixes vertices with the service only when it builds more than one
# target, so a Bake session's vertices are printed when it closes: prefixed when its build definitions read
# "[SERVICE internal] load build definition ...", unprefixed (no service; "K SESSION unprefixed") when its one build
# definition reads "[internal] load build definition ...". Mixed prefixes or several unprefixed build definitions
# are "X" (outside the supported format). G051 C5: a re-printed vertex may differ only by its service prefix in a
# prefixed session (a target sharing it); its blocks keep their own service. G051 C6 (-v targets=" SVC... "): a
# lone service label of a frontend vertex is such an alias only between bound build targets of the call (bakealias);
# "D SESSION N LINE" rows carry the bounded, reduced context of each vertex an X row names (dflush).
LEDGER_TABLE_AWK='
BEGIN { OFS = "\t"; sess = soff + 0; cur = ""; insess = 0; vl = ""; fresh = 0; bake = 0; nbv = 0 }
function endblock() { if (cur != "") last[sess, cur] = (bterm > 0) ? "T" : (bcont ? "C" : "O"); cur = "" }
function closesess(   k, n, a) {
  endblock()
  if (!insess) return
  if (bake) bakeflush()
  n = split(vl, a, " ")
  for (k = 1; k <= n; k++) if (last[sess, a[k]] != "T") xv(a[k], "no terminal line for #" a[k])
  dflush()
  vl = ""; bake = 0
}
function term(s) { if (bterm > 0) xv(cur, "duplicate terminal line for #" cur); bterm++; tdone[sess, cur] = 1; print "T", sess, cur, s }
# G051 C6: the bounded context of every vertex an X row names (xv). Each "#N" line of a session is kept reduced (dadd):
# vertex names, terminal, continuation, naming and unknown status lines verbatim, the BuildKit status templates of
# dred verbatim, log lines, layer progress and anything else as a class token; consecutive repeats counted; 200
# characters and 200 entries per vertex. At the session end (dflush) "D SESSION N LINE" prints the first and last 10
# entries of at most 6 named vertices. The raw stream itself is never kept.
function xv(v, why) { print "X", sess, why; if (!((sess, v) in dfl)) { dfl[sess, v] = 1; dvl = dvl " " v } }
function dred(r) {
  if (r ~ /^[0-9]+\.[0-9]+( |$)/) return "<log>"
  if (r ~ /^sha256:[0-9a-f]+ /) return "<layer progress>"
  if (r ~ /^(resolve image config for |docker-image:\/\/|resolve [^ ]+( [0-9.]+s)? done$|load (build definition from|metadata for|\.dockerignore|build context|local bake definitions)|transferring (dockerfile|context|config)|exporting (to image|layers|manifest|config|attestation)|writing image sha256:|unpacking |extracting sha256:|resolving provenance|preparing |sending tarball|importing to docker|exporting to docker image format)/) return r
  return "<other>"
}
function dadd(v, t,   k) {
  t = substr("#" v " " t, 1, 200); k = dn[sess, v] + 0
  if (k > 0 && dl[sess, v, k] == t) { dc[sess, v, k]++; return }
  if (k >= 200) { dover[sess, v]++; return }
  dn[sess, v] = ++k; dl[sess, v, k] = t; dc[sess, v, k] = 1
}
function dflush(   a, n, j, k, m, v) {
  n = split(dvl, a, " ")
  for (j = 1; j <= n && j <= 6; j++) {
    v = a[j]; m = dn[sess, v] + 0
    for (k = 1; k <= m; k++) {
      if (m > 20 && k == 11) { print "D", sess, v, "#" v " <" (m - 20) " entries omitted>"; k = m - 10; continue }
      print "D", sess, v, dl[sess, v, k] ((dc[sess, v, k] > 1) ? " x" dc[sess, v, k] : "")
    }
    if (dover[sess, v] > 0) print "D", sess, v, "#" v " <" dover[sess, v] " further lines not kept>"
  }
  if (n > 6) print "D", sess, "-", "<" (n - 6) " further named vertices not shown>"
  dvl = ""
}
function vrow(v, rest, pfx,   lab, text, n, tk, s0, svc, ltype, stage, si, sn, k, ij) {
  lab = ""; text = rest
  if (substr(rest, 1, 1) == "[" && index(rest, "] ") > 0) { lab = substr(rest, 2, index(rest, "] ") - 2); text = substr(rest, index(rest, "] ") + 2) }
  n = split(lab, tk, " "); s0 = 1; svc = ""
  # Compose prefixes the service (pfx); a lone [internal] or [auth] (e.g. bake definitions) belongs to the session.
  if (pfx && n >= 1 && !(n == 1 && (tk[1] == "internal" || tk[1] == "auth"))) { svc = tk[1]; s0 = 2 }
  ltype = "none"; stage = "-"; si = "-"; sn = "-"
  if (lab == "") ltype = "none"
  else if (s0 > n) ltype = "plain"
  else if (tk[n] ~ /^[0-9]+\/[0-9]+$/) {
    ltype = "step"; split(tk[n], ij, "/"); si = ij[1]; sn = ij[2]; stage = ""
    for (k = s0; k < n; k++) stage = stage ((stage == "") ? "" : " ") tk[k]
    if (stage == "") stage = "-"
  }
  else if (tk[s0] == "internal") ltype = "internal"
  else if (tk[s0] == "context") ltype = "context"
  else ltype = "other"
  print "V", sess, v, svc, ltype, stage, si, sn, lg_squash(text)
}
# bakealias A B: two names of one re-printed vertex that differ only in their first label token, the service of a
# prefixed Bake session: "[S1 rest] text" and "[S2 rest] text" with S1 != S2, the same remaining label tokens (stage,
# index or internal) and the same operation text. G051 C6: a lone service label ("[S] text", the frontend vertices of
# a prefixed session, d7) is an alias only for the frontend operations "resolve image config for docker-image://..."
# and "docker-image://...", with the same text, and when both services are bound (bakebound). Anything else (no
# label, a reserved internal or auth token, a token that is not a bound target, a lone token against a longer label,
# another stage, index or operation) is not an alias.
function bakealias(a, b,   ia, ib, la, lb, na, nb, ta, tb, k) {
  ia = index(a, "] "); ib = index(b, "] ")
  if (substr(a, 1, 1) != "[" || substr(b, 1, 1) != "[" || ia == 0 || ib == 0) return 0
  if (substr(a, ia) != substr(b, ib)) return 0
  la = substr(a, 2, ia - 2); lb = substr(b, 2, ib - 2)
  na = split(la, ta, " "); nb = split(lb, tb, " ")
  if (na < 1 || na != nb || ta[1] == tb[1] || ta[1] == "internal" || tb[1] == "internal") return 0
  if (na == 1) {
    if (ta[1] == "auth" || tb[1] == "auth" || !bakebound(ta[1]) || !bakebound(tb[1])) return 0
    return (index(substr(a, ia + 2), "resolve image config for docker-image://") == 1 || index(substr(a, ia + 2), "docker-image://") == 1)
  }
  for (k = 2; k <= na; k++) if (ta[k] != tb[k]) return 0
  return 1
}
# bakebound S: S is a build target of the call (-v targets, from its Compose plan) that prints its own export
# ("[S] exporting to image") in this session.
function bakebound(s) { return (index(targets, " " s " ") > 0 && (s in bexp)) }
function bakeflush(   k, r, lab, text, n, tk, p, u, v, vr) {
  # G051 C4: BuildKit re-prints a vertex name at each of its print blocks; a vertex counts once (by session and
  # vertex). G051 C5: in a prefixed (multi-target) session BuildKit re-prints a vertex shared by several targets under
  # the service prefix of another target; such an alias (bakealias) keeps its own service-qualified V row. Any other
  # different name of one vertex, or any different name in an unprefixed session, is outside the format.
  p = 0; u = 0; split("", bexp)
  for (k = 1; k <= nbv; k++) {
    v = bv[k]; r = brest[k]
    if (v in vr) continue
    vr[v] = r
    lab = ""; text = r
    if (substr(r, 1, 1) == "[" && index(r, "] ") > 0) { lab = substr(r, 2, index(r, "] ") - 2); text = substr(r, index(r, "] ") + 2) }
    if (lab != "" && index(lab, " ") == 0 && text ~ /^exporting to (image|docker image format|oci image format)$/) bexp[lab] = 1
    if (index(text, "load build definition from ") != 1) continue
    n = split(lab, tk, " ")
    if (n == 1 && tk[1] == "internal") u++; else if (n == 2 && tk[2] == "internal") p++
  }
  for (k = 1; k <= nbv; k++) {
    v = bv[k]; r = brest[k]
    if (r == vr[v]) continue
    if (!(p > 0 && u == 0 && bakealias(vr[v], r))) xv(v, "Bake progress prints #" v " under different names")
  }
  if (p > 0 && u > 0) print "X", sess, "Bake progress mixes prefixed and unprefixed build definitions"
  else if (u > 1) print "X", sess, "Bake progress with more than one unprefixed build definition"
  for (k = 1; k <= nbv; k++) vrow(bv[k], brest[k], (p > 0))
  if (p == 0 && u == 1) print "K", sess, "unprefixed"
  nbv = 0
}
function header(v, rest) {
  if (rest ~ /^(DONE [0-9.]+s|CACHED|ERROR|ERROR: .*|CANCELED|\.\.\.)$/) { xv(v, "status line without a vertex name for #" v); return }
  if (!((sess, v) in seen)) { seen[sess, v] = 1; vl = vl " " v }
  if (bake) { nbv++; bv[nbv] = v; brest[nbv] = rest; return }
  vrow(v, rest, (mode == "compose"))
}
{
  line = $0; sub(/\r$/, "", line)
  if (line ~ /^ *=> / || line ~ /^\[\+\] Building [0-9]/) { print "X", 0, "TTY-format progress"; next }
  if (substr(line, 1, 18) == "#0 building with \"") {
    closesess(); sess++; insess = 1; fresh = 1
    nm = substr(line, 19); dr = nm; sub(/".*$/, "", nm)
    if (!sub(/^[^"]*" instance using /, "", dr)) dr = "unknown"
    sub(/ driver$/, "", dr)
    print "S", sess, ((nm == "") ? "-" : nm), ((dr == "") ? "unknown" : dr), "header"; next
  }
  if (line !~ /^#[0-9]+ /) { endblock(); next }
  v = substr(line, 2); sub(/ .*$/, "", v); rest = substr(line, length(v) + 3)
  # A new invocation numbers its vertices from #1 again: its Bake definitions vertex opens a session unless it
  # directly follows a header or re-prints a vertex of this session that has no terminal line yet.
  if (mode == "compose" && rest == "[internal] load local bake definitions") {
    if (insess && fresh) { if (!bake) { bake = 1; print "Q", sess } }
    else if (!insess || !((sess, v) in seen) || ((sess, v) in tdone)) {
      closesess(); sess++; insess = 1; bake = 1; print "S", sess, "-", "unknown", "bake"; print "Q", sess
    }
  }
  if (!insess) { print "X", 0, "vertex line before a session header"; next }
  dt = rest
  if (rest ~ /^[0-9]+\.[0-9]+( |$)/ || rest ~ /^sha256:[0-9a-f]+ /) dt = dred(rest)
  else if (v == cur && rest != "..." && rest !~ /^(DONE [0-9]+(\.[0-9]+)?s|CACHED|CANCELED|ERROR|ERROR: .*)$/ && rest !~ /^[A-Z][A-Z]+([: ]|$)/ && substr(rest, 1, 10) != "naming to ") dt = dred(rest)
  dadd(v, dt)
  if (v != cur) { fresh = 0; endblock(); cur = v; bterm = 0; bcont = 0; header(v, rest); next }
  if (rest == "...") { bcont = 1; next }
  if (rest ~ /^DONE [0-9]+(\.[0-9]+)?s$/) { term("DONE"); next }
  if (rest == "CACHED") { term("CACHED"); next }
  if (rest == "ERROR" || substr(rest, 1, 7) == "ERROR: ") { term("ERROR"); next }
  if (rest == "CANCELED") { term("CANCELED"); next }
  if (rest ~ /^[A-Z][A-Z]+([: ]|$)/) { xv(v, "unknown status token for #" v); next }
  if (substr(rest, 1, 10) == "naming to ") { nm = substr(rest, 11); sub(/ .*$/, "", nm); print "N", sess, v, nm }
}
END { closesess() }'

# One row from a progress table (files: model, image snapshot, table; -v sess svc mode rc prebuild tagged).
# Prints "progress executed cached frontend bases result driver reason". parsed requires: the session header and
# this build's "load build definition" vertex; every expected step of the reachable stages printed, under this
# build's prefix or, when BuildKit deduplicated the identical vertex, under another label with the same index and
# text (a FROM by its resolved reference); no step vertex of this build outside the model; every external base
# resolved by digest (or, for a base from the local store, by the image ID taken just before the call); a resolved
# frontend when the Dockerfile names one; a completed export (none for a cache-only prebuild) agreeing with the exit
# status, and an ERROR vertex for a nonzero exit. Executed counts DONE/ERROR RUN, COPY and ADD vertices this build
# printed first; a vertex reused from the session (deduplicated) or CACHED counts as cached.
# G051 C3 (-v bdrv bwhy same): a headerless Bake session ("S ... bake") takes its driver only from the call's
# verified builder evidence (bdrv; empty: unknown with bwhy); evidence that contradicts a header's driver is
# unknown. In a Bake session ("Q") Bake prints one build definition vertex for services whose build definitions are
# identical, so a service without its own uses the one vertex of a service in "same" (the services whose Compose
# context and Dockerfile equal this one's) when exactly one exists; its own completed export is still required.
# Services with identical build definitions in one Bake session share every vertex under arbitrary prefixes; their
# split execution is not reassigned, so such a pair stays a duplicate (UNMET).
LEDGER_ROW_AWK='
BEGIN { FS = OFS = "\t"; why_s = ""; nst = 0; nb = 0; nv = 0; nS = 0; sbake = 0; qbake = 0 }
function why(m) { if (why_s == "") why_s = m; else if (length(why_s) < 400 && index(why_s, m) == 0) why_s = why_s "; " m }
function lastword(t,   r) { r = t; sub(/^.* /, "", r); return r }
function utext(t,   w) { w = t; sub(/ .*$/, "", w); return toupper(w) substr(t, length(w) + 1) }
FILENAME == ARGV[1] {
  if ($1 == "syntax") syn = $2
  else if ($1 == "error") why("Dockerfile: " $2)
  else if ($1 == "step") { nst++; sl[nst] = $2; si[nst] = $3; sn[nst] = $4; sk[nst] = $5; st[nst] = $6 }
  else if ($1 == "base") { nb++; bl[nb] = $2; br[nb] = $3; bh[nb] = $4 }
  next
}
FILENAME == ARGV[2] { if (!($1 in loc)) loc[$1] = $2; next }
$1 == "S" { if ($2 == sess) { hasS = 1; drv = $4; sbake = ($5 == "bake") } nS++; next }
$1 == "Q" { if ($2 == sess) qbake = 1; next }
$1 == "K" { next }
$1 == "X" { if ($2 == sess || $2 == 0) why($3); next }
$2 != sess { next }
$1 == "V" { nv++; vv[nv] = $3; vs[nv] = $4; vt[nv] = $5; vi[nv] = $7; vx[nv] = $9; if (!($3 in own)) own[$3] = $4; if ($4 == svc) mine[$3] = 1; next }
$1 == "T" { if ($4 == "ERROR") { verr[$3] = 1; anyerr = 1 } else if ($4 == "DONE") vdone[$3] = 1; else if ($4 == "CACHED") vcached[$3] = 1; next }
$1 == "N" { nmd[$3] = 1; next }
END {
  if (!hasS) why("no progress session")
  if (mode == "explicit" && nS > 1) why("more than one progress session")
  if (sbake) { if (bdrv == "") why((bwhy != "") ? bwhy : "Bake progress without verified builder evidence"); else drv = bdrv }
  else if (qbake && bdrv != "" && bdrv != drv) why("builder evidence (" bdrv ") contradicts the progress header (" drv ")")
  for (k = 1; k <= nv; k++) if (vs[k] == svc && vt[k] == "internal" && index(vx[k], "load build definition from ") == 1) hdr = 1
  if (!hdr && qbake) {
    c = 0
    for (k = 1; k <= nv; k++) if (vs[k] != svc && vt[k] == "internal" && index(vx[k], "load build definition from ") == 1 && index(same, " " vs[k] " ") > 0 && !(vv[k] in twin)) { twin[vv[k]] = 1; c++ }
    if (c == 1) hdr = 1
  }
  if (!hdr) why("no build definition vertex")
  for (k = 1; k <= nv; k++) if (index(vx[k], "FROM ") == 1) {
    r = lg_normref(lastword(vx[k])); n = lg_nodig(r); d = lg_digof(r)
    if (d != "") { if ((n in fres) && fres[n] != d) conflict[n] = 1; fres[n] = d }
    if (vt[k] == "step" && vi[k] == 1 && (!(n in fv) || vs[k] == svc)) fv[n] = vv[k]
  }
  for (j = 1; j <= nst; j++) {
    v = ""
    if (sk[j] == "FROM") { n = lg_nodig(st[j]); if (n in fv) v = fv[n] }
    else for (k = 1; k <= nv; k++) if (vt[k] == "step" && vi[k] == si[j] && utext(vx[k]) == st[j]) { if (vs[k] == svc) { v = vv[k]; break } if (v == "") v = vv[k] }
    if (v == "") { why("missing step " sl[j] " " si[j] "/" sn[j]); continue }
    if (sk[j] == "RUN" || sk[j] == "COPY" || sk[j] == "ADD") cnt[v] = 1
  }
  for (k = 1; k <= nv; k++) if (vs[k] == svc && vt[k] == "step") {
    ok = 0
    if (index(vx[k], "FROM ") == 1) { n = lg_nodig(lg_normref(lastword(vx[k]))); for (j = 1; j <= nst; j++) if (sk[j] == "FROM" && lg_nodig(st[j]) == n) ok = 1 }
    else for (j = 1; j <= nst; j++) if (sk[j] != "FROM" && si[j] == vi[k] && st[j] == utext(vx[k])) ok = 1
    if (!ok) why("unexpected step vertex #" vv[k])
  }
  ex = 0; ca = 0
  for (v in cnt) { if ((v in verr) || (v in vdone)) { if (own[v] == svc) ex++; else ca++ } else if (v in vcached) ca++ }
  m = 0
  for (b = 1; b <= nb; b++) {
    n = lg_nodig(br[b]); pin = lg_digof(br[b]); d = (n in fres) ? fres[n] : ""
    if (n in conflict) { why("conflicting resolution of " n); continue }
    if (d == "" && (br[b] in loc)) { d = loc[br[b]]; sub(/^sha256:/, "", d) }
    if (length(d) != 64) { why("unresolved base " n); continue }
    if (pin != "" && pin != d) { why("base " n " resolved to another digest than its pin"); continue }
    e[++m] = bl[b] ((bh[b] == "from") ? "=" : "+") n "@sha256:" d
  }
  for (i = 2; i <= m; i++) { x = e[i]; for (j = i - 1; j >= 1 && e[j] > x; j--) e[j + 1] = e[j]; e[j + 1] = x }
  bases = ""
  for (i = 1; i <= m; i++) bases = bases ((i > 1) ? "," : "") e[i]
  if (nb == 0) bases = "none"; else if (m < nb) bases = "unknown"
  fe = "builtin"
  if (syn != "") {
    fe = "unknown"
    for (k = 1; k <= nv; k++) if (index(vx[k], "docker-image://") == 1) {
      r = substr(vx[k], 16)
      if (lg_nodig(r) == lg_nodig(syn) && lg_digof(r) != "" && (lg_digof(syn) == "" || lg_digof(syn) == lg_digof(r))) fe = lg_nodig(syn) "@sha256:" lg_digof(r)
    }
    if (fe == "unknown") why("frontend " syn " unresolved")
  }
  expd = 0; expv = 0; ev = ""
  for (k = 1; k <= nv; k++) if (vs[k] == svc && vx[k] ~ /^exporting to (image|docker image format|oci image format)$/) { expv = 1; if ((vv[k] in vdone) && !(vv[k] in verr)) { expd = 1; ev = vv[k] } }
  res = "-"
  if (mode == "explicit") {
    if (prebuild && expv) why("cache-only prebuild exported an image")
    if (rc == 0) {
      if (anyerr) why("exit 0 with an ERROR vertex")
      if (!prebuild && !expd) why("no completed export")
      if (!prebuild && tagged && !(ev in nmd)) why("no naming line")
      res = 0
    } else { if (!anyerr) why("nonzero exit without an ERROR vertex"); res = rc }
  } else {
    ownerr = 0; for (v in verr) if (v in mine) ownerr = 1
    if (expd && !ownerr) res = 0; else if (ownerr) res = 1; else why("no completed export or ERROR for " svc)
    if (rc == "0" && anyerr) why("exit 0 with an ERROR vertex")
  }
  prog = (why_s == "") ? "parsed" : "unknown"
  if (prog == "unknown") { ex = "-"; ca = "-" }
  print prog, ex, ca, fe, bases, res, ((drv == "") ? "unknown" : drv), ((why_s == "") ? "-" : why_s)
}'

# raw.tsv ("B CALL NPRED JOB SUITE KIND", "R CALL <20 row fields> NOTE", "F CALL") -> builds.tsv rows from seq 1;
# a call without its F line (interrupted) leaves "missing" rows. Notes go to -v notes=FILE as "SEQ NOTE".
LEDGER_ASSEMBLE_AWK='
BEGIN { FS = OFS = "\t"; n = 0; seq = 0 }
function ph(c, w,   k, line) {
  seq++; line = seq OFS bj[c] OFS bs[c] OFS bk[c]
  for (k = 1; k <= 12; k++) line = line OFS "unknown"
  print line, "missing", "-", "-", "-", "-"; print seq, w > notes
}
$1 == "B" { if (!($2 in np)) ord[++n] = $2; np[$2] = $3; bj[$2] = $4; bs[$2] = $5; bk[$2] = $6; next }
$1 == "R" { nr[$2]++; row[$2, nr[$2]] = $0; next }
$1 == "F" { fin[$2] = 1; next }
END {
  for (i = 1; i <= n; i++) {
    c = ord[i]
    for (j = 1; j <= nr[c]; j++) {
      if (split(row[c, j], f, "\t") != 23) { ph(c, "malformed ledger row"); continue }
      seq++; line = seq; for (x = 3; x <= 22; x++) line = line OFS f[x]
      print line
      if (f[23] != "-" && f[23] != "") print seq, f[23] > notes
    }
    if (!(c in fin)) { m = ((np[c] ~ /^[0-9]+$/) ? np[c] + 0 : 1) - nr[c]; if (m < 1 && nr[c] == 0) m = 1; for (j = 1; j <= m; j++) ph(c, "build interrupted before its row") }
  }
}'

# §3.8.3 measures over builds.tsv (files: images list, notes, builds.tsv; -v hdr exc imgmissing bootok bootstate).
# Prints the scalar summary lines, "duplicate SEQS" per unexcepted duplicated key, "ND A B excepted|unexcepted INV-A
# INV-B NAMED-A NAMED-B" per near-duplicate pair and "unknown_row SEQ REASON"; exit 2 when builds.tsv is malformed.
LEDGER_CHECK_AWK='
BEGIN {
  FS = OFS = "\t"; ne = split(exc, xa, " "); for (i = 1; i <= ne; i++) okx["named:" xa[i]] = 1
  split("dockerfile dockerfile_sha256 target frontend bases platform ignore_file ignore_sha256 context_inventory named_contexts args_digest", fnm, " ")
}
function add(u, m) { return (u == "") ? m : u "; " m }
FILENAME == ARGV[1] { if ($1 == "" || substr($1, 1, 1) == "#") next; r = lg_normref($1); img[lg_nodig(r)] = 1; if (lg_digof(r) != "") pin[lg_nodig(r)] = lg_digof(r); next }
FILENAME == ARGV[2] { note[$1] = $2; next }
FNR == 1 { if ($0 != hdr) { bad = "header"; exit } next }
{ if (NF != 21 || $1 !~ /^[0-9]+$/) { bad = "line " FNR; exit } nr++; for (k = 1; k <= 21; k++) F[nr, k] = $k }
END {
  if (bad != "") { print "malformed builds.tsv (" bad ")" > "/dev/stderr"; exit 2 }
  nunk = 0; rows = 0; nboot = 0
  for (r = 1; r <= nr; r++) if (F[r, 4] == "bootstrap") nboot++
  for (r = 1; r <= nr; r++) {
    u = ""; kind = F[r, 4]; prog = F[r, 17]
    if (prog != "missing") rows++
    if (prog == "missing") u = "no row"
    else if (prog == "not_observed") { if (kind != "bootstrap") u = "progress not observed" }
    else if (prog != "parsed") u = "progress " prog
    if (kind !~ /^(authoritative|prebuild|bootstrap|named:[A-Za-z0-9][A-Za-z0-9-]*)$/) u = add(u, "kind " kind)
    for (k = 5; k <= 15; k++) if (F[r, k] == "" || F[r, k] ~ /(^|[;,=:])unknown($|[;,])/) u = add(u, fnm[k - 4] " unknown")
    if (prog == "parsed" && (F[r, 18] !~ /^[0-9]+$/ || F[r, 19] !~ /^[0-9]+$/)) u = add(u, "step counts")
    if (kind == "bootstrap" && (!bootok || nboot > 1)) u = add(u, "bootstrap " bootstate)
    if (F[r, 9] != "none" && F[r, 9] != "unknown") {
      nbs = split(F[r, 9], be, ",")
      for (i = 1; i <= nbs; i++) {
        x = be[i]; p = index(x, "="); q = index(x, "+"); if (p == 0 || (q > 0 && q < p)) p = q
        ref = substr(x, p + 1); n = lg_nodig(ref); d = lg_digof(ref)
        if (p < 2 || d == "") u = add(u, "base entry " x)
        else if (imgmissing) u = add(u, "no images.txt for base " n)
        else if (!(n in img)) u = add(u, "base " n " not in images.txt")
        else if ((n in pin) && pin[n] != d) u = add(u, "base " n " differs from its images.txt pin")
      }
    }
    if (u != "") { nunk++; print "unknown_row", F[r, 1] " " ((note[F[r, 1]] != "") ? note[F[r, 1]] : u); continue }
    exe[r] = (kind == "bootstrap" || F[r, 18] > 0)
  }
  if (bootstate != "not-applicable" && nboot == 0) bootok = 0
  if (bootstate == "not-applicable" && nboot > 0) bootok = 0
  ndup = 0; nnd = 0; nune = 0; nexc = 0
  for (r = 1; r <= nr; r++) if (exe[r]) {
    key = ""; for (k = 5; k <= 15; k++) key = key SUBSEP F[r, k]
    kn[key]++; kr[key] = kr[key] " " r; if (F[r, 4] in okx) kx[key]++
  }
  for (key in kn) if (kn[key] > 1) {
    m = split(kr[key], rr, " "); s = ""
    for (i = 1; i <= m; i++) { s = s ((s == "") ? "" : " ") F[rr[i], 1]; if (F[rr[i], 4] in okx) used[rr[i]] = 1 }
    if (kn[key] - 1 - kx[key] > 0) { ndup++; print "duplicate", s }
  }
  for (r = 1; r <= nr; r++) if (exe[r]) for (s = r + 1; s <= nr; s++) if (exe[s]) {
    same = 1; for (k = 5; k <= 12; k++) if (F[r, k] != F[s, k]) same = 0
    if (!same || F[r, 15] != F[s, 15] || (F[r, 13] == F[s, 13] && F[r, 14] == F[s, 14])) continue
    nnd++; x = (F[r, 4] in okx) || (F[s, 4] in okx)
    if (x) { if (F[r, 4] in okx) used[r] = 1; if (F[s, 4] in okx) used[s] = 1 } else nune++
    print "ND", F[r, 1], F[s, 1], (x ? "excepted" : "unexcepted"), F[r, 13], F[s, 13], F[r, 14], F[s, 14]
  }
  for (r in used) nexc++
  met = (nunk == 0 && ndup == 0 && nune == 0 && rows == nr && bootok)
  print "expected", nr; print "rows", rows; print "unknown", nunk; print "duplicates", ndup
  print "near_duplicates", nnd; print "exceptions", nexc; print "bootstrap", bootstate; print "ac5_status", (met ? "MET" : "UNMET")
}'

# yq expression: one TSV row per Compose service (all profiles): name, has-build, image, context, dockerfile,
# target, platforms, platform, base64 build args, base64 additional contexts, profiles, depends_on, inline
# Dockerfile, project.
LEDGER_YQ_CFG='.name as $p | (.services // {}) | to_entries | .[] | [.key, ((.value | has("build")) and .value.build != null), (.value.image // "-"), (.value.build.context // "-"), (.value.build.dockerfile // "Dockerfile"), (.value.build.target // "-"), ((.value.build.platforms // []) | join(",")), (.value.platform // "-"), ((.value.build.args // {}) | to_entries | map(.key + "=" + ((.value // "") | tostring)) | join("\n") | @base64), ((.value.build.additional_contexts // {}) | to_entries | map(.key + "=" + .value) | join("\n") | @base64), ((.value.profiles // []) | join(",")), ((.value.depends_on // {}) | keys | join(",")), ((.value.build.dockerfile_inline // "") != ""), $p]'

# ledger_init: once per run, after WORK exists. Exports BUILDKIT_PROGRESS=plain (Compose and documented
# commands keep their text) and records the daemon version and default platform.
ledger_init() {
  [[ -z "$LEDGER_DIR" ]] || return 0
  local v
  mkdir -p "$WORK/ledger/inv" "$WORK/ledger/cfg" "$WORK/ledger/model" "$WORK/ledger/progress" "$WORK/ledger/watch" && : >"$WORK/ledger/raw.tsv" || return 1
  LEDGER_DIR="$WORK/ledger"
  export BUILDKIT_PROGRESS=plain
  v="$(docker version --format '{{.Server.Version}} {{.Server.Os}}/{{.Server.Arch}}' 2>/dev/null)" || v=""
  if [[ "${v% *}" =~ ^[0-9A-Za-z][0-9A-Za-z.+~-]*$ ]]; then LEDGER_DAEMON="${v% *}"; fi
  if [[ "${v#* }" =~ ^[a-z0-9]+/[a-z0-9_]+$ ]]; then LEDGER_PLATFORM="${v#* }"; fi
  printf 'daemon\t%s\nplatform\t%s\n' "$LEDGER_DAEMON" "$LEDGER_PLATFORM" >"$LEDGER_DIR/daemon.tsv"
  log "build ledger: daemon $LEDGER_DAEMON, default platform $LEDGER_PLATFORM, BUILDKIT_PROGRESS=plain"
}

_lg_sha() { local h; h="$(sha256sum -- "$1" 2>/dev/null)" || return 1; printf '%s' "${h:0:64}"; }
# _lg_split LINE: LG_F = the tab-separated fields of LINE, empty fields kept.
_lg_split() { local IFS=$'\x1f'; read -r -a LG_F <<<"${1//$'\t'/$'\x1f'}"; }
_lg_normref() { # REF -> LG_REF (as lg_normref)
  local r="$1" name dig="" first
  if [[ "$r" == *@* ]]; then dig="@${r#*@}"; name="${r%%@*}"; else name="$r"; fi
  if [[ "$name" != */* ]]; then name="docker.io/library/$name"
  else first="${name%%/*}"; if [[ "$first" != *[.:]* && "$first" != localhost ]]; then name="docker.io/$name"; fi; fi
  if [[ "$name" =~ ^docker\.io/[^/]+$ ]]; then name="docker.io/library/${name#docker.io/}"; fi
  if [[ "${name##*/}" != *:* && -z "$dig" ]]; then name="$name:latest"; fi
  LG_REF="$name$dig"
}
_lg_callinfo() { # KIND [SUITE]: LG_JOB LG_SUITEV LG_KINDV
  LG_JOB="${LEDGER_JOB:-${CI_JOB_NAME:-local}}"
  [[ "$LG_JOB" =~ ^[A-Za-z0-9][A-Za-z0-9:_.-]{0,63}$ ]] || LG_JOB=unknown
  LG_SUITEV="${2:-${LEDGER_SUITE:-}}"
  [[ "$LG_SUITEV" =~ ^[a-z][a-z0-9-]{0,31}$ ]] || LG_SUITEV=unknown
  LG_KINDV="$1"
  [[ "$LG_KINDV" =~ ^(authoritative|prebuild|named:[A-Za-z0-9][A-Za-z0-9-]{0,63})$ ]] || LG_KINDV=unknown
}
# _lg_snapshot FILE: "reference<TAB>image ID" of every tagged local image (one docker call).
_lg_snapshot() {
  local -
  set -o pipefail
  docker image ls --no-trunc --format '{{.Repository}}:{{.Tag}} {{.ID}}' 2>/dev/null |
    awk "$LEDGER_AWK_LIB"'$1 !~ /<none>/ && $2 ~ /^sha256:[0-9a-f]+$/ { print lg_normref($1) "\t" $2 }' >"$1"
}
_lg_snapid() { awk -F'\t' -v r="$2" '$1 == r { print $2; exit }' "$1" 2>/dev/null; } # SNAP REF -> image ID
# _lg_begin DEST NPRED [KIND [SUITE]]: a new call (LG_CALL), its image snapshot and its B line.
_lg_begin() {
  _lg_callinfo "${3:-$LEDGER_KIND}" "${4:-}"
  LG_CALL="$(date +%s%N).$BASHPID.$RANDOM"
  LG_SNAPBAD=0
  _lg_snapshot "$LEDGER_DIR/progress/$LG_CALL.snap1" || { LG_SNAPBAD=1; : >"$LEDGER_DIR/progress/$LG_CALL.snap1"; }
  printf 'B\t%s\t%s\t%s\t%s\t%s\n' "$LG_CALL" "$2" "$LG_JOB" "$LG_SUITEV" "$LG_KINDV" >>"$LEDGER_DIR/raw.tsv"
}
# _lg_emit CALL FIELD... NOTE: one R line (20 row fields after seq; empty or multi-line fields become unknown).
_lg_emit() {
  local out="R"$'\t'"$1" f; shift
  for f in "$@"; do
    if [[ -z "$f" || "$f" == *$'\t'* || "$f" == *$'\n'* ]]; then f=unknown; fi
    out+=$'\t'"$f"
  done
  printf '%s\n' "$out" >>"$LEDGER_DIR/raw.tsv"
}
_lg_unknown_row() { # CALL NOTE: a row with nothing determinable
  _lg_emit "$1" "$LG_JOB" "$LG_SUITEV" "$LG_KINDV" unknown unknown unknown unknown unknown unknown unknown unknown unknown unknown unknown unknown missing - - - - "$2"
}

# _lg_inv DIR: LG_INVD = SHA-256 over DIR's complete entry inventory (every entry below DIR: type, mode, numeric
# owner, relative path, symlink target, regular-file SHA-256), enumerated NUL-separated and encoded as one
# NUL-terminated item per entry with length-prefixed path and target, sorted bytewise; kept as inv/<digest>.
# Fails on unreadable entries or types other than file, directory and symlink (the field is then unknown).
_lg_inv() {
  local LC_ALL=C dir="$1" t e p y m o l h
  local -
  set -o pipefail
  local -A lgH=()
  if [[ -n "${_LG_INVC[$dir]:-}" ]]; then LG_INVD="${_LG_INVC[$dir]}"; return 0; fi
  [[ -d "$dir" ]] || return 1
  t="$LEDGER_DIR/inv/.tmp.$BASHPID.$RANDOM"
  if ! (cd -- "$dir" && find . -mindepth 1 -type f -print0 | xargs -0 -r sha256sum -z --) >"$t.h" 2>/dev/null ||
     ! (cd -- "$dir" && find . -mindepth 1 -printf '%P\0%y\0%m\0%U:%G\0%l\0') >"$t.m" 2>/dev/null; then
    rm -f -- "$t".*; return 1
  fi
  while IFS= read -r -d '' e; do
    [[ ${#e} -gt 66 && "${e:64:2}" == "  " ]] || { rm -f -- "$t".*; return 1; }
    p="${e:66}"; lgH["${p#./}"]="${e:0:64}"
  done <"$t.h"
  while IFS= read -r -d '' p && IFS= read -r -d '' y && IFS= read -r -d '' m && IFS= read -r -d '' o && IFS= read -r -d '' l; do
    case "$y" in
      f) h="${lgH[$p]:-}"; [[ -n "$h" ]] || { rm -f -- "$t".*; return 1; } ;;
      d|l) h=- ;;
      *) rm -f -- "$t".*; return 1 ;;
    esac
    printf '%s %s %s %s %s:%s %s:%s\0' "$y" "$m" "$o" "$h" "${#p}" "$p" "${#l}" "$l"
  done <"$t.m" >"$t.u"
  LC_ALL=C sort -z -- "$t.u" >"$t.s" || { rm -f -- "$t".*; return 1; }
  h="$(sha256sum <"$t.s")" || { rm -f -- "$t".*; return 1; }
  LG_INVD="${h:0:64}"
  mv -f -- "$t.s" "$LEDGER_DIR/inv/$LG_INVD"; rm -f -- "$t".*
  _LG_INVC[$dir]="$LG_INVD"
}

# _lg_model DF TARGET NAMES: LG_MODEL = the cached Dockerfile model (LEDGER_DF_AWK) of DF for TARGET.
_lg_model() {
  local k tmp
  k="$( { _lg_sha "$1"; printf '\n%s\n%s\n' "$2" "$3"; } | sha256sum)"
  LG_MODEL="$LEDGER_DIR/model/${k:0:64}.tsv"
  [[ -s "$LG_MODEL" ]] && return 0
  tmp="$LG_MODEL.$BASHPID.$RANDOM"
  awk -v target="$2" -v named=" $3 " "$LEDGER_AWK_LIB$LEDGER_DF_AWK" "$1" >"$tmp" 2>/dev/null || printf 'error\tDockerfile unreadable\n' >"$tmp"
  mv -f -- "$tmp" "$LG_MODEL"
}

# _lg_key CTX DF TARGET PLATFORM NAMED ARGS: the §3.8.1 key fields LG_DF LG_DFSHA LG_TGT LG_FROMPLAT LG_IGN
# LG_IGNSHA LG_INV LG_NAMED LG_ARGS LG_PLAT and LG_MODEL. NAMED: "name=path" lines; ARGS: "KEY=VALUE" lines.
# Paths are relative to the context (stable across disposable copies).
_lg_key() {
  local ctx="$1" df="$2" plat="$4" named="$5" args="$6" a b n p lgnames="" lgents=() line
  LG_DF=unknown LG_DFSHA=unknown LG_TGT="${3:-final}" LG_IGN=unknown LG_IGNSHA=unknown LG_INV=unknown LG_NAMED=none LG_ARGS=none
  LG_PLAT="${plat:-$LEDGER_PLATFORM}"
  [[ "$LG_PLAT" =~ ^[a-z0-9]+/[a-z0-9_]+(/[a-z0-9]+)?$ ]] || LG_PLAT=unknown
  if [[ -f "$df" ]]; then LG_DF="$(realpath -m --relative-to="$ctx" -- "$df")"; LG_DFSHA="$(_lg_sha "$df")" || LG_DFSHA=unknown; fi
  if [[ -d "$ctx" ]]; then
    a="$df.dockerignore" b="$ctx/.dockerignore"
    if [[ -f "$a" ]]; then
      LG_IGN="$(realpath -m --relative-to="$ctx" -- "$a")"; LG_IGNSHA="$(_lg_sha "$a")" || LG_IGNSHA=unknown
      if [[ -f "$b" ]]; then LG_IGN+=";also:.dockerignore"; LG_IGNSHA+=";also:$(_lg_sha "$b" || printf unknown)"; fi
    elif [[ -f "$b" ]]; then LG_IGN=.dockerignore; LG_IGNSHA="$(_lg_sha "$b")" || LG_IGNSHA=unknown
    else LG_IGN=none LG_IGNSHA=none; fi
    if _lg_inv "$ctx"; then LG_INV="$LG_INVD"; fi
  fi
  if [[ -n "$named" ]]; then
    while IFS= read -r line; do
      [[ -n "$line" ]] || continue
      n="${line%%=*}" p="${line#*=}"
      lgnames+="$n "
      if [[ "$line" == *=* && "$n" =~ ^[A-Za-z0-9][A-Za-z0-9_.-]*$ && "$p" != *://* && "$p" != *:* && -d "$p" ]] && _lg_inv "$(realpath -m -- "$p")"; then
        lgents+=("$n=$LG_INVD")
      else
        lgents+=("$n=unknown")
      fi
    done <<<"$named"
    LG_NAMED="$(printf '%s\n' "${lgents[@]}" | LC_ALL=C sort | paste -sd, -)"
  fi
  if [[ -n "${args//$'\n'/}" ]]; then
    LG_ARGS="$(printf '%s\n' "$args" | grep -v '^$' | LC_ALL=C sort | sha256sum)"; LG_ARGS="${LG_ARGS:0:64}"
  fi
  if [[ -f "$df" ]]; then _lg_model "$df" "${3:-}" "$lgnames"; else LG_MODEL="$LEDGER_DIR/model/missing.tsv"; printf 'error\tDockerfile missing\n' >"$LG_MODEL"; fi
  while IFS=$'\t' read -r a b p; do
    if [[ "$a" == platform ]]; then LG_PLAT+=";from:$b=$p"; fi
  done <"$LG_MODEL"
  return 0
}

# _lg_eval TABLE SNAP SESSION SVC MODE RC PREBUILD TAGGED [BDRV BWHY SAME]: LEDGER_ROW_AWK over LG_MODEL into LG_PROG
# LG_EXEC LG_CACHED LG_FRONT LG_BASES LG_RES LG_DRV LG_WHY (BDRV/BWHY: the call's builder evidence verdict; SAME: the
# services with this one's Compose context and Dockerfile).
_lg_eval() {
  local o
  o="$(awk -v sess="$3" -v svc="$4" -v mode="$5" -v rc="$6" -v prebuild="$7" -v tagged="$8" -v bdrv="${9:-}" -v bwhy="${10:-}" -v same=" ${11:-} " \
    "$LEDGER_AWK_LIB$LEDGER_ROW_AWK" "$LG_MODEL" "$2" "$1" 2>/dev/null)" || o=""
  IFS=$'\t' read -r LG_PROG LG_EXEC LG_CACHED LG_FRONT LG_BASES LG_RES LG_DRV LG_WHY <<<"$o"
  if [[ "$LG_PROG" != parsed && "$LG_PROG" != unknown ]]; then
    LG_PROG=unknown LG_EXEC=- LG_CACHED=- LG_FRONT=unknown LG_BASES=unknown LG_RES=- LG_DRV=unknown LG_WHY="progress evaluation failed"
  fi
}

# _lg_platform_check ID: an image output must match the expected platform (LG_PLAT before ";"), else LG_PLAT=unknown.
_lg_platform_check() {
  local got want="${LG_PLAT%%;*}"
  got="$(docker image inspect --format '{{.Os}}/{{.Architecture}}{{with .Variant}}/{{.}}{{end}}' "$1" 2>/dev/null)" || got=""
  [[ "$got" == */arm64/v8 && "$want" != */v8 ]] && got="${got%/v8}"
  if [[ -z "$got" || "$got" != "$want" ]]; then LG_PLAT=unknown; _lg_why "image platform '${got:-none}' is not $want"; fi
}
_lg_why() { if [[ -z "${LG_WHY:-}" || "$LG_WHY" == - ]]; then LG_WHY="$1"; else LG_WHY+="; $1"; fi; } # MESSAGE: add a reason

# _lg_context CALL TABLE: the table's "D" rows (G051 C6: bounded context of the vertices its X rows name) for the
# unknown_context lines of ledger-summary.tsv; the published file passes the run's literal, gitleaks and scrub gates.
_lg_context() { awk -F'\t' -v c="$1" '$1 == "D" { print c "\t" $2 "\t" $4 }' "$2" >>"$LEDGER_DIR/context.tsv" 2>/dev/null || true; }

# _lg_table OUT MODE FILE...: the progress table of the FILEs, sessions numbered across them (LG_TBLTGT: the call's
# build targets, which may own lone-label aliases; empty: none).
_lg_table() {
  local tabf="$1" mode="$2" f n; shift 2
  : >"$tabf" || return 1
  for f in "$@"; do
    [[ -f "$f" ]] || continue
    n="$(awk -F'\t' '$1 == "S"' "$tabf" | wc -l)"
    awk -v mode="$mode" -v soff="$n" -v targets=" ${LG_TBLTGT:-} " "$LEDGER_AWK_LIB$LEDGER_TABLE_AWK" "$f" >>"$tabf" || return 1
  done
}

# ledger_build SUITE KIND -- DOCKER-BUILD-ARGS...: replaces `docker build -q ...` (G051 §3.8.2): the same build with
# --progress=plain and an --iidfile, its progress kept under the ledger, one row; prints nothing; returns the
# build's status. KIND: authoritative | prebuild (cache-only: no tag, --output type=cacheonly added) | named:<ID>.
ledger_build() {
  local suite="$1" kind="$2"; shift 2
  [[ "${1:-}" == -- ]] && shift
  if [[ -z "$LEDGER_DIR" ]]; then docker build "$@" >/dev/null 2>&1; return; fi
  local -
  set +e
  local a p pb args=() df="" ctx="" target="" plat="" named="" bargs="" tagged=0 outp="" bad="" rc=0 id plog iid img cmd=(docker build)
  while (($#)); do
    a="$1"; shift
    case "$a" in
      -q|--quiet|--progress=*|--iidfile=*) ;;
      --progress|--iidfile) shift ;;
      -f|--file) df="${1-}"; args+=("$a" "${1-}"); shift ;;
      --file=*) df="${a#*=}"; args+=("$a") ;;
      --target) target="${1-}"; args+=("$a" "${1-}"); shift ;;
      --target=*) target="${a#*=}"; args+=("$a") ;;
      --platform) plat="${1-}"; args+=("$a" "${1-}"); shift ;;
      --platform=*) plat="${a#*=}"; args+=("$a") ;;
      --build-context) named+="${1-}"$'\n'; args+=("$a" "${1-}"); shift ;;
      --build-context=*) named+="${a#*=}"$'\n'; args+=("$a") ;;
      --build-arg|--build-arg=*)
        if [[ "$a" == --build-arg ]]; then p="${1-}"; args+=("$a" "${1-}"); shift; else p="${a#*=}"; args+=("$a"); fi
        if [[ "$p" == *=* ]]; then bargs+="$p"$'\n'
        elif [[ ! "$p" =~ ^[A-Za-z_][A-Za-z0-9_]*$ ]]; then bad="build argument '$p'"
        elif [[ -n "${!p+x}" ]]; then bargs+="$p=${!p}"$'\n'; fi ;;
      -t|--tag) tagged=1; args+=("$a" "${1-}"); shift ;;
      --tag=*) tagged=1; args+=("$a") ;;
      -o|--output) outp="${1-}"; args+=("$a" "${1-}"); shift ;;
      --output=*) outp="${a#*=}"; args+=("$a") ;;
      --secret|--ssh|--secret=*|--ssh=*) bad="build secret or ssh input (not in the key)"; args+=("$a"); [[ "$a" == *=* ]] || { args+=("${1-}"); shift; } ;;
      --label|--network|--add-host|--cache-from|--cache-to|--shm-size|--ulimit|--annotation|--attest|--allow|--metadata-file|--provenance|--sbom|--no-cache-filter|--cgroup-parent|--builder) args+=("$a" "${1-}"); shift ;;
      --load|--push) outp="${outp:-image}"; args+=("$a") ;;
      --pull|--no-cache|--rm|--force-rm) args+=("$a") ;;
      --*=*) args+=("$a") ;;
      -*) bad="option $a"; args+=("$a") ;;
      *) [[ -z "$ctx" ]] || bad="more than one context"; ctx="$a"; args+=("$a") ;;
    esac
  done
  if [[ -z "$ctx" || "$ctx" == - || "$ctx" == *://* ]]; then bad="${bad:-context '${ctx:-none}' unsupported}"; else ctx="$(realpath -m -- "$ctx")"; fi
  [[ "$df" != - ]] || bad="${bad:-Dockerfile on stdin unsupported}"
  df="$(realpath -m -- "${df:-$ctx/Dockerfile}")"
  if [[ "$kind" == prebuild ]]; then
    cmd=(docker buildx build)
    if [[ "${_LG_PREBUILD_TAG_OK:-0}" == 1 ]]; then :
    elif [[ "$tagged" == 0 && -z "$outp" ]]; then args+=(--output type=cacheonly)
    elif [[ "$tagged" == 1 || "$outp" != type=cacheonly ]]; then
      _lg_begin "$ctx" 1 prebuild "$suite"
      _lg_unknown_row "$LG_CALL" "a prebuild must be cache-only (no tag or image output); not built"
      printf 'F\t%s\n' "$LG_CALL" >>"$LEDGER_DIR/raw.tsv"
      return 2
    fi
  fi
  _lg_begin "$ctx" 1 "$kind" "$suite"
  id="$LG_CALL" plog="$LEDGER_DIR/progress/$LG_CALL.log" iid="$LEDGER_DIR/progress/$LG_CALL.iid"
  "${cmd[@]}" --progress=plain --iidfile "$iid" "${args[@]}" >"$plog" 2>&1 || rc=$?
  LG_PREBUILD_UNSUPPORTED=0
  if [[ "$kind" == prebuild && $rc -ne 0 ]] && ! grep -q '^#0 building with "' "$plog" && grep -qi cacheonly "$plog"; then LG_PREBUILD_UNSUPPORTED=1; fi
  _LG_INVC=()
  _lg_key "$ctx" "$df" "$target" "$plat" "$named" "$bargs"
  _lg_table "$plog.tab" explicit "$plog" || : >"$plog.tab"
  _lg_context "$id" "$plog.tab"
  pb=0; [[ "$kind" == prebuild && "$outp" != image && $tagged == 0 ]] && pb=1
  _lg_eval "$plog.tab" "$LEDGER_DIR/progress/$id.snap1" 1 "" explicit "$rc" "$pb" "$tagged"
  [[ -z "$bad" ]] || { LG_PROG=unknown LG_EXEC=- LG_CACHED=-; _lg_why "$bad"; }
  img=-
  if (( pb )); then img=cacheonly
  elif (( rc == 0 )); then
    img="$(cat -- "$iid" 2>/dev/null)"
    if [[ "$img" =~ ^sha256:[0-9a-f]{64}$ ]]; then _lg_platform_check "$img"; else img=unknown; LG_PROG=unknown; _lg_why "no image ID"; fi
  fi
  if (( LG_SNAPBAD )) && [[ "$LG_BASES" == unknown ]]; then _lg_why "image inventory unavailable"; fi
  _lg_emit "$id" "$LG_JOB" "$LG_SUITEV" "$LG_KINDV" "$LG_DF" "$LG_DFSHA" "$LG_TGT" "$LG_FRONT" "$LG_BASES" "$LG_PLAT" "$LG_IGN" "$LG_IGNSHA" \
    "$LG_INV" "$LG_NAMED" "$LG_ARGS" "$LEDGER_DAEMON/$LG_DRV" "$LG_PROG" "$LG_EXEC" "$LG_CACHED" "$LG_RES" "$img" "${LG_WHY:--}"
  printf 'F\t%s\n' "$id" >>"$LEDGER_DIR/raw.tsv"
  return "$rc"
}

# ledger_prebuild NAME -- DOCKER-BUILD-ARGS...: a cache-only prebuild (G051 L4) meant to run in bg_start: the build
# through `docker buildx build --output type=cacheonly --progress=plain` with one row of kind prebuild. Only when the
# builder rejects the cache-only output before building does it fall back to a run-labelled tag recorded for
# ownership (that image output stays visible in its row).
ledger_prebuild() {
  local name="$1" rc=0 tag; shift
  [[ "${1:-}" == -- ]] && shift
  [[ "$name" =~ ^[a-z][a-z0-9-]{0,31}$ ]] || return 2
  ledger_build "${LEDGER_SUITE:-prebuild}" prebuild -- "$@" || rc=$?
  if (( rc != 0 )) && [[ "${LG_PREBUILD_UNSUPPORTED:-0}" == 1 ]]; then
    tag="g001-prebuild-$RUN_ID-$name"
    rc=0
    _LG_PREBUILD_TAG_OK=1 ledger_build "${LEDGER_SUITE:-prebuild}" prebuild -- --label "g001.run=$RUN_ID" -t "$tag" "$@" || rc=$?
    record_built_image "$tag"
  fi
  return "$rc"
}

# _lg_parse_compose ARGS...: the arguments after `docker compose` into _LC_SUB, _LC_SERVICES, _LC_PROFILES,
# _LC_FILES, _LC_ENVFILE and the flags _LC_BUILD _LC_NOBUILD _LC_NODEPS _LC_DEPS; _LC_BAD names anything not
# understood (an option whose arity is unknown, a project-changing global option).
_lg_parse_compose() {
  local a
  _LC_SUB="" _LC_BUILD=0 _LC_NOBUILD=0 _LC_NODEPS=0 _LC_DEPS=0 _LC_BAD="" _LC_ENVFILE="" _LC_BUILDER=""
  _LC_PROFILES=() _LC_SERVICES=() _LC_FILES=()
  while (($#)); do
    a="$1"; shift
    if [[ -z "$_LC_SUB" ]]; then
      case "$a" in
        --profile) _LC_PROFILES+=("${1-}"); shift ;;
        --profile=*) _LC_PROFILES+=("${a#*=}") ;;
        -f|--file) _LC_FILES+=("${1-}"); shift ;;
        --file=*) _LC_FILES+=("${a#*=}") ;;
        --env-file) _LC_ENVFILE="${1-}"; shift ;;
        --env-file=*) _LC_ENVFILE="${a#*=}" ;;
        --compatibility|--dry-run|--all-resources) ;;
        -*) _LC_BAD="global option $a" ;;
        *) _LC_SUB="$a" ;;
      esac
      continue
    fi
    case "$_LC_SUB" in up|create|run|build) ;; *) return 0 ;; esac
    if [[ "$_LC_SUB" == run && ${#_LC_SERVICES[@]} -eq 1 ]]; then break; fi
    case "$a" in
      --build) _LC_BUILD=1 ;;
      --no-build) _LC_NOBUILD=1 ;;
      --no-deps) _LC_NODEPS=1 ;;
      --with-dependencies) _LC_DEPS=1 ;;
      --builder) _LC_BUILDER="${1-}"; shift ;;
      --builder=*) _LC_BUILDER="${a#*=}" ;;
      --*=*) ;;
      -e|--env|--env-from-file|-v|--volume|--name|--entrypoint|-w|--workdir|-u|--user|-l|--label|-p|--publish|--cap-add|--cap-drop|--pull|--scale|-t|--timeout|--wait-timeout|--exit-code-from|--attach|--no-attach|--build-arg|-m|--memory|--ssh) shift ;;
      -d|--detach|--rm|-T|--no-TTY|-i|--interactive|--service-ports|--use-aliases|--remove-orphans|--quiet-pull|-q|--quiet|--wait|--force-recreate|--no-recreate|-V|--renew-anon-volumes|--abort-on-container-exit|--abort-on-container-failure|--always-recreate-deps|--no-color|--no-log-prefix|--no-start|--timestamps|--menu|--watch|-y|--yes|--no-cache|--push|--check|--print) ;;
      -*) _LC_BAD="option $a of $_LC_SUB" ;;
      *) _LC_SERVICES+=("$a") ;;
    esac
  done
}
_lg_buildable() { # ARGS...: the Compose call can build (pure argument parse, no Docker invocation)
  _lg_parse_compose "$@"
  case "$_LC_SUB" in build) return 0 ;; up|create|run) [[ "$_LC_NOBUILD" == 0 ]] ;; *) return 1 ;; esac
}

# _lg_config DEST ENVFILE FILE...: prints the cached LEDGER_YQ_CFG table of `docker compose config --format json`
# (all profiles) for these inputs; only build fields are kept (the configuration's environment never reaches disk).
_lg_config() {
  local dest="$1" envf="$2" f k out ef tmp; shift 2
  local -
  set -o pipefail
  local fa=()
  [[ "$envf" == /* ]] && ef="$envf" || ef="$dest/$envf"
  k="$( { printf '%s\0' "$dest" "$envf" "$@"; cat -- "$ef"; for f in "$@"; do if [[ "$f" == /* ]]; then cat -- "$f"; else cat -- "$dest/$f"; fi; done; } 2>/dev/null | sha256sum)"
  out="$LEDGER_DIR/cfg/${k:0:64}.tsv"
  if [[ ! -s "$out" ]]; then
    for f in "$@"; do fa+=(-f "$f"); done
    tmp="$out.$BASHPID.$RANDOM"
    (cd -- "$dest" && docker compose --env-file "$envf" "${fa[@]}" --profile '*' config --format json 2>/dev/null) |
      yq_stdin -p json -o tsv "$LEDGER_YQ_CFG" >"$tmp" 2>/dev/null && [[ -s "$tmp" ]] || { rm -f -- "$tmp"; return 1; }
    mv -f -- "$tmp" "$out"
  fi
  printf '%s' "$out"
}

_lg_plan_reset() { LG_LINES=0; LG_U=(); LG_DEF=(); LG_DEFBAD=(); LG_IMG=(); LG_LE=(); LG_LW=(); LG_PRED=(); LG_PROJS=(); LG_BLDNAMES=""; }

# _lg_plan_argv DEST ARGV...: one `docker compose` command (ARGV after "docker compose") into the plan: for line k,
# LG_LE[k] = services it builds even when present (--build, build) and LG_LW[k] = targeted build services incl.
# the dependencies it starts; the build definitions (LG_DEF, a conflict in LG_DEFBAD) and image names (LG_IMG).
# Anything unresolvable adds a reason to LG_U (an unknown row).
_lg_plan_argv() {
  local dest="$1"; shift
  local envf cfg line svc s d p act=" " e w q
  local -A lgBLD=() lgPR=() lgDE=() lgsel=() lgclo=()
  _lg_parse_compose "$@"
  case "$_LC_SUB" in build) ;; up|create|run) [[ "$_LC_NOBUILD" == 0 ]] || return 0 ;; *) return 0 ;; esac
  if [[ -n "$_LC_BAD" ]]; then LG_U+=("compose command not understood ($_LC_BAD)"); return 0; fi
  if [[ -n "$_LC_BUILDER" ]]; then LG_BLDNAMES+="$_LC_BUILDER "; fi
  ((${#_LC_FILES[@]})) || { LG_U+=("compose command without -f"); return 0; }
  envf="${_LC_ENVFILE:-.env}"
  cfg="$(_lg_config "$dest" "$envf" "${_LC_FILES[@]}")" || { LG_U+=("compose config unavailable for $_LC_SUB"); return 0; }
  while IFS= read -r line; do
    _lg_split "$line"
    svc="${LG_F[0]:-}"; [[ -n "$svc" ]] || continue
    lgBLD[$svc]="${LG_F[1]:-}" lgPR[$svc]="${LG_F[10]:-}" lgDE[$svc]="${LG_F[11]:-}"
    [[ "${LG_F[1]:-}" == true ]] || continue
    [[ -z "${LG_F[13]:-}" ]] || LG_PROJS[${LG_F[13]}]=1
    if [[ "${LG_F[12]:-}" != false ]]; then LG_DEFBAD[$svc]="inline Dockerfile"; fi
    if [[ "${LG_F[2]:-}" != - ]]; then _lg_normref "${LG_F[2]}"; else _lg_normref "${LG_F[13]:-}-$svc"; fi
    d="${LG_F[3]:-}"$'\x1f'"${LG_F[4]:-}"$'\x1f'"${LG_F[5]:-}"$'\x1f'"${LG_F[6]:-}"$'\x1f'"${LG_F[7]:-}"$'\x1f'"${LG_F[8]:-}"$'\x1f'"${LG_F[9]:-}"$'\x1f'"$LG_REF"
    if [[ -n "${LG_DEF[$svc]:-}" && "${LG_DEF[$svc]}" != "$d" ]]; then LG_DEFBAD[$svc]="conflicting build definitions"; fi
    LG_DEF[$svc]="$d" LG_IMG[$svc]="$LG_REF"
  done <"$cfg"
  [[ "$envf" == /* ]] && e="$envf" || e="$dest/$envf"
  p="${COMPOSE_PROFILES:-},$(sed -n 's/^COMPOSE_PROFILES=//p' "$e" 2>/dev/null | tail -1)"
  for q in "${_LC_PROFILES[@]}" ${p//,/ }; do act+="$q "; done
  if ((${#_LC_SERVICES[@]})); then
    for s in "${_LC_SERVICES[@]}"; do [[ -n "${lgBLD[$s]+x}" ]] || { LG_U+=("unknown service $s"); return 0; }; lgsel[$s]=1; done
  else
    for s in "${!lgBLD[@]}"; do
      [[ -z "${lgPR[$s]}" || "$act" == *" * "* ]] && { lgsel[$s]=1; continue; }
      for q in ${lgPR[$s]//,/ }; do [[ "$act" == *" $q "* ]] && lgsel[$s]=1; done
    done
  fi
  for s in "${!lgsel[@]}"; do lgclo[$s]=1; done
  if [[ ( "$_LC_SUB" != build && "$_LC_NODEPS" == 0 ) || ( "$_LC_SUB" == build && "$_LC_DEPS" == 1 ) ]]; then
    w=1
    while (( w )); do
      w=0
      for s in "${!lgclo[@]}"; do
        for d in ${lgDE[$s]//,/ }; do
          [[ -n "${lgBLD[$d]+x}" ]] || { LG_U+=("unknown dependency $d"); continue; }
          [[ -n "${lgclo[$d]:-}" ]] || { lgclo[$d]=1; w=1; }
        done
      done
    done
  fi
  LG_LINES=$((LG_LINES + 1)); LG_LE[$LG_LINES]="" LG_LW[$LG_LINES]=""
  for s in "${!lgclo[@]}"; do
    [[ "${lgBLD[$s]}" == true ]] || continue
    LG_LW[$LG_LINES]+="$s "
    if [[ "$_LC_SUB" == build || ( "$_LC_BUILD" == 1 && ( "$_LC_SUB" != run || -n "${lgsel[$s]:-}" ) ) ]]; then LG_LE[$LG_LINES]+="$s "; fi
  done
}

# _lg_plan_line DEST LINE: one line of a command block into the plan (docker compose commands; any other build
# command or a line this parser cannot split safely is an unknown row).
_lg_plan_line() {
  local dest="$1" line="$2" lgtok=()
  line="${line#"${line%%[![:space:]]*}"}"
  [[ -n "$line" && "$line" != \#* && "$line" == *docker* ]] || return 0
  if [[ "$line" =~ $LG_SHELL_RE ]]; then LG_U+=("command line not parsed: ${line:0:100}"); return 0; fi
  read -r -a lgtok <<<"$line"
  if [[ "${lgtok[0]}" == docker && "${lgtok[1]:-}" == compose ]]; then _lg_plan_argv "$dest" "${lgtok[@]:2}"
  elif [[ "${lgtok[0]}" == docker && ( "${lgtok[1]:-}" == build || ( "${lgtok[1]:-}" == buildx && "${lgtok[2]:-}" == build ) ) ]]; then LG_U+=("docker build in a command block")
  elif [[ "${lgtok[0]}" != docker ]]; then LG_U+=("command line not parsed: ${line:0:100}"); fi
}
_lg_block_buildable() { # CMDS: some line can build or cannot be parsed (pure, no Docker invocation)
  local line lgtok=()
  while IFS= read -r line; do
    line="${line#"${line%%[![:space:]]*}"}"
    [[ -n "$line" && "$line" != \#* && "$line" == *docker* ]] || continue
    [[ ! "$line" =~ $LG_SHELL_RE ]] || return 0
    read -r -a lgtok <<<"$line"
    [[ "${lgtok[0]}" == docker ]] || return 0
    if [[ "${lgtok[1]:-}" == compose ]]; then _lg_buildable "${lgtok[@]:2}" && return 0
    elif [[ "${lgtok[1]:-}" == build || "${lgtok[1]:-}" == buildx ]]; then return 0; fi
  done <<<"$1"
  return 1
}

# _lg_predict SNAP: LG_PRED[svc] = builds predicted by the plan's lines in order (LG_LE always; LG_LW when the
# service image is absent before that line).
_lg_predict() {
  local k s r
  local -A lgpresent=() lgset=()
  while IFS=$'\t' read -r r s; do [[ -n "$r" ]] && lgpresent[$r]=1; done <"$1"
  for ((k = 1; k <= LG_LINES; k++)); do
    lgset=()
    for s in ${LG_LE[$k]}; do lgset[$s]=1; done
    for s in ${LG_LW[$k]}; do [[ -n "${lgpresent[${LG_IMG[$s]:-?}]:-}" ]] || lgset[$s]=1; done
    for s in "${!lgset[@]}"; do LG_PRED[$s]=$(( ${LG_PRED[$s]:-0} + 1 )); lgpresent[${LG_IMG[$s]:-?}]=1; done
  done
}

# _lg_service_row CALL SVC SESSION TABLE RC [MISSING-REASON]: the row of one Compose service build.
_lg_service_row() {
  local id="$1" svc="$2" sess="$3" tab="$4" rc="$5" miss="${6:-}" ctx dfl tgt pls pl a64 c64 ref df plat named args img same t tc td
  if [[ -z "${LG_DEF[$svc]:-}" || -n "${LG_DEFBAD[$svc]:-}" ]]; then
    _lg_unknown_row "$id" "service $svc: ${LG_DEFBAD[$svc]:-no build definition in the Compose configuration}${miss:+; $miss}"; return
  fi
  IFS=$'\x1f' read -r ctx dfl tgt pls pl a64 c64 ref <<<"${LG_DEF[$svc]}"
  [[ "$dfl" == /* ]] && df="$dfl" || df="$ctx/$dfl"
  plat=""
  if [[ -n "$pls" ]]; then plat="$pls"; [[ "$pls" != *,* ]] || plat=unknown; elif [[ "$pl" != - ]]; then plat="$pl"; fi
  named="$(base64 -d <<<"$c64" 2>/dev/null)"; args="$(base64 -d <<<"$a64" 2>/dev/null)"
  _lg_key "$ctx" "$(realpath -m -- "$df")" "${tgt#-}" "$plat" "$named" "$args"
  if [[ -n "$miss" ]]; then
    _lg_emit "$id" "$LG_JOB" "$LG_SUITEV" "$LG_KINDV" "$LG_DF" "$LG_DFSHA" "$LG_TGT" unknown unknown "$LG_PLAT" "$LG_IGN" "$LG_IGNSHA" \
      "$LG_INV" "$LG_NAMED" "$LG_ARGS" "$LEDGER_DAEMON/unknown" missing - - - - "service $svc: $miss"
    return
  fi
  same=""
  for t in "${!LG_DEF[@]}"; do
    [[ "$t" != "$svc" && -z "${LG_DEFBAD[$t]:-}" ]] || continue
    IFS=$'\x1f' read -r tc td _ <<<"${LG_DEF[$t]}"
    if [[ "$tc" == "$ctx" && "$td" == "$dfl" ]]; then same+="$t "; fi
  done
  _lg_eval "$tab" "$LEDGER_DIR/progress/$id.snap1" "$sess" "$svc" compose "$rc" 0 0 "$LG_BDRV" "$LG_BWHY" "$same"
  img="$(_lg_snapid "$LEDGER_DIR/progress/$id.snap2" "$ref")"
  if [[ "$LG_RES" == 0 ]]; then
    if [[ "$img" =~ ^sha256:[0-9a-f]{64}$ ]]; then _lg_platform_check "$img"; else img=unknown; LG_PROG=unknown; _lg_why "no image for $ref"; fi
  fi
  _lg_emit "$id" "$LG_JOB" "$LG_SUITEV" "$LG_KINDV" "$LG_DF" "$LG_DFSHA" "$LG_TGT" "$LG_FRONT" "$LG_BASES" "$LG_PLAT" "$LG_IGN" "$LG_IGNSHA" \
    "$LG_INV" "$LG_NAMED" "$LG_ARGS" "$LEDGER_DAEMON/$LG_DRV" "$LG_PROG" "$LG_EXEC" "$LG_CACHED" "$LG_RES" "${img:--}" "service $svc: ${LG_WHY:--}"
}

# _lg_builder OUT [NAME]: the builder a Compose invocation uses, read-only: `docker buildx inspect [NAME]` resolves
# the --builder name, BUILDX_BUILDER or the context's selected builder exactly as Bake does, and never creates,
# selects or starts one. Writes "NAME<TAB>DRIVER" when the builder section has exactly one of each, else "-<TAB>-".
_lg_builder() {
  local o n d
  o="$(docker buildx inspect ${2:+"$2"} 2>/dev/null)" || o=""
  n="$(awk '/^Nodes:/ {exit} /^Name:/ {c++; v = $2} END {if (c == 1) print v}' <<<"$o")"
  d="$(awk '/^Nodes:/ {exit} /^Driver:/ {c++; v = $2} END {if (c == 1) print v}' <<<"$o")"
  if [[ "$n" =~ ^[A-Za-z0-9][A-Za-z0-9_.-]*$ && "$d" =~ ^[a-z][a-z0-9-]*$ ]]; then printf '%s\t%s\n' "$n" "$d" >"$1"; else printf -- '-\t-\n' >"$1"; fi
}
# _lg_bld_verdict CALL MODE: LG_BDRV = the driver of the builder read before (CALL.bld1) and after (CALL.bld2) the
# call or block when both readings are complete and identical, else empty with the reason in LG_BWHY. A --builder
# option inside a command block is not attributable to one invocation (no verdict).
_lg_bld_verdict() {
  local b1 b2 nb
  LG_BDRV="" LG_BWHY=""
  read -r -a nb <<<"$LG_BLDNAMES"
  if [[ "$2" == block && ${#nb[@]} -gt 0 ]]; then LG_BWHY="a --builder option in a command block"; return 0; fi
  _lg_builder "$LEDGER_DIR/progress/$1.bld2" "$([[ "$2" == call ]] && printf '%s' "$LG_BLDUSE")"
  b1="$(cat -- "$LEDGER_DIR/progress/$1.bld1" 2>/dev/null)" || b1=""
  b2="$(cat -- "$LEDGER_DIR/progress/$1.bld2" 2>/dev/null)" || b2=""
  if [[ -z "$b1" || "$b1" == -$'\t'- || -z "$b2" || "$b2" == -$'\t'- ]]; then LG_BWHY="builder evidence unavailable"; return 0; fi
  if [[ "$b1" != "$b2" ]]; then LG_BWHY="the selected builder changed during the call ($b1 / $b2)"; return 0; fi
  LG_BDRV="${b1#*$'\t'}"
}
# _lg_bake_map TABLE: an unprefixed Bake session ("K") is assigned the one service this call or block builds (the
# plan's build services) whose image its one completed export names on its one "naming to" line; otherwise it keeps
# no service and an unknown row says why (its predicted builds stay missing).
_lg_bake_map() {
  local tab="$1" s nex nn nm t k
  local -A lgbs=()
  local ca=()
  for ((k = 1; k <= LG_LINES; k++)); do for t in ${LG_LW[$k]}; do lgbs[$t]=1; done; done
  while IFS=$'\t' read -r s nex nn nm; do
    ca=()
    if [[ "$nex" == 1 && "$nn" == 1 ]]; then
      _lg_normref "$nm"
      for t in "${!lgbs[@]}"; do if [[ "${LG_IMG[$t]:-}" == "$LG_REF" && -z "${LG_DEFBAD[$t]:-}" ]]; then ca+=("$t"); fi; done
    fi
    if (( ${#ca[@]} == 1 )); then
      awk -F'\t' -v OFS='\t' -v s="$s" -v svc="${ca[0]}" '$1 == "V" && $2 == s && $4 == "" { $4 = svc } { print }' "$tab" >"$tab.m" &&
        mv -f -- "$tab.m" "$tab" || LG_U+=("Bake progress session $s could not be assigned")
    else
      LG_U+=("Bake progress session $s: exports $nex, names $nn (${nm:-none}); not the image of exactly one service built here")
    fi
  done < <(awk -F'\t' '
    $1 == "K" { k[$2] = 1 }
    $1 == "V" && $9 ~ /^exporting to (image|docker image format|oci image format)$/ { if (!(($2, $3) in ex)) { ex[$2, $3] = 1; nex[$2]++ } }
    $1 == "N" { if (!(($2, $3, $4) in nk)) { nk[$2, $3, $4] = 1; nr++; ns[nr] = $2; nv[nr] = $3; nn[nr] = $4 } }
    END {
      for (r = 1; r <= nr; r++) if ((ns[r], nv[r]) in ex) { c[ns[r]]++; name[ns[r]] = nn[r] }
      for (s in k) print s "\t" (nex[s] + 0) "\t" (c[s] + 0) "\t" name[s]
    }' "$tab")
}

# _lg_finish CALL DEST RC PROJECT MODE LOG...: rows for one Compose call (MODE call) or command block (MODE block)
# from the BuildKit sessions in the LOGs: one row per (session, service) printed; a missing row for every predicted
# build without a section; a missing row for every watched or <PROJECT>-<svc> image that appeared or changed ID
# without a section (blocks, or calls that predicted or printed builds); an unknown row per LG_U reason; then F.
_lg_finish() {
  local id="$1" dest="$2" rc="$3" proj="$4" mode="$5"; shift 5
  local tab="$LEDGER_DIR/progress/$id.tab" s sess n p r i1 i2 k LG_TBLTGT
  local -A lgOBS=() lgSEEN=()
  _LG_INVC=()
  _lg_bld_verdict "$id" "$mode"
  LG_TBLTGT="$(printf '%s ' "${!LG_PRED[@]}")"
  _lg_table "$tab" compose "$@" || { : >"$tab"; LG_U+=("progress table not built"); }
  _lg_context "$id" "$tab"
  _lg_bake_map "$tab"
  while IFS=$'\t' read -r s sess; do lgOBS[$s]+="$sess "; done < <(awk -F'\t' '$1 == "V" && $4 != "" && !(($4, $2) in u) { u[$4, $2] = 1; print $4 "\t" $2 }' "$tab")
  if [[ "$mode" == block || ${#LG_PRED[@]} -gt 0 || ${#lgOBS[@]} -gt 0 ]]; then
    _lg_snapshot "$LEDGER_DIR/progress/$id.snap2" || { : >"$LEDGER_DIR/progress/$id.snap2"; LG_U+=("image inventory after the call unavailable"); }
    for s in $(printf '%s\n' "${!LG_PRED[@]}" "${!lgOBS[@]}" | LC_ALL=C sort -u); do
      n=0
      for sess in ${lgOBS[$s]:-}; do _lg_service_row "$id" "$s" "$sess" "$tab" "$rc"; n=$((n + 1)); done
      while (( n < ${LG_PRED[$s]:-0} )); do _lg_service_row "$id" "$s" 0 "$tab" "$rc" "no build progress for a predicted build"; n=$((n + 1)); done
      lgSEEN[$s]=1
    done
    for ((k = 1; k <= LG_LINES; k++)); do
      for s in ${LG_LW[$k]}; do
        [[ -z "${lgSEEN[$s]:-}" ]] || continue
        i1="$(_lg_snapid "$LEDGER_DIR/progress/$id.snap1" "${LG_IMG[$s]}")"; i2="$(_lg_snapid "$LEDGER_DIR/progress/$id.snap2" "${LG_IMG[$s]}")"
        if [[ -n "$i2" && "$i1" != "$i2" ]]; then lgSEEN[$s]=1; _lg_service_row "$id" "$s" 0 "$tab" "$rc" "image appeared or changed without build progress"; fi
      done
    done
    for p in "$proj" "${!LG_PROJS[@]}"; do
      [[ "$p" =~ ^[a-z0-9][a-z0-9_-]*$ ]] || continue
      while IFS=$'\t' read -r r i2; do
        [[ "$r" == "docker.io/library/$p-"*":latest" ]] || continue
        s="${r#docker.io/library/"$p"-}"; s="${s%:latest}"
        [[ -z "${lgSEEN[$s]:-}" ]] || continue
        i1="$(_lg_snapid "$LEDGER_DIR/progress/$id.snap1" "$r")"
        if [[ "$i1" != "$i2" ]]; then lgSEEN[$s]=1; _lg_service_row "$id" "$s" 0 "$tab" "$rc" "image $p-$s appeared or changed without build progress"; fi
      done <"$LEDGER_DIR/progress/$id.snap2"
    done
  fi
  for r in "${LG_U[@]}"; do _lg_unknown_row "$id" "$r"; done
  printf 'F\t%s\n' "$id" >>"$LEDGER_DIR/raw.tsv"
}

# _lg_compose DEST NFILES FILE... ARGS...: compose_call for a call that can build.
_lg_compose() {
  local dest="$1" nf="$2" f fa=() rc=0 of ef id; shift 2
  local files=("${@:1:nf}"); shift "$nf"
  local -
  set +e
  for f in "${files[@]}"; do fa+=(-f "$f"); done
  _lg_plan_reset
  _lg_plan_argv "$dest" --env-file .env "${fa[@]}" "$@"
  if [[ -z "${LG_LW[1]:-}" && ${#LG_U[@]} -eq 0 ]]; then
    (cd "$dest" && docker compose --env-file .env "${fa[@]}" "$@")
    return
  fi
  _lg_begin "$dest" "?"
  id="$LG_CALL" of="$LEDGER_DIR/progress/$LG_CALL.out" ef="$LEDGER_DIR/progress/$LG_CALL.err"
  LG_BLDUSE="${LG_BLDNAMES%% *}"
  _lg_builder "$LEDGER_DIR/progress/$id.bld1" "$LG_BLDUSE"
  _lg_predict "$LEDGER_DIR/progress/$id.snap1"
  (cd "$dest" && docker compose --env-file .env "${fa[@]}" "$@") >"$of" 2>"$ef" || rc=$?
  cat -- "$of"
  cat -- "$ef" >&2
  _lg_finish "$id" "$dest" "$rc" "" call "$of" "$ef"
  rm -f -- "$of" "$ef"
  return "$rc"
}

# _lg_block_end CALL DEST RC PROJECT LOG CMDS: rows for a command block (run_docs_block, ledger_watch_end).
_lg_block_end() {
  local id="$1" dest="$2" rc="$3" proj="$4" log="$5" cmds="$6" line
  local -
  set +e
  _lg_plan_reset
  while IFS= read -r line; do _lg_plan_line "$dest" "$line"; done <<<"$cmds"
  _lg_predict "$LEDGER_DIR/progress/$id.snap1"
  _lg_finish "$id" "$dest" "$rc" "$proj" block "$log"
  return 0
}

# ledger_watch_begin DEST PROJECT TOKEN / ledger_watch_end DEST PROJECT TOKEN LOGFILE CMDTEXT: account builds of
# commands the harness runs itself (CMDTEXT: the exact docker compose lines run from DEST, unquoted; LOGFILE: their
# merged output). Begin snapshots the image IDs; end parses LOGFILE per service and adds a row for every printed
# build section, every predicted build (--build/build lines: all targeted build services; others: absent images)
# and every <PROJECT>-<svc> or targeted image that appeared or changed ID. Both always return 0.
ledger_watch_begin() {
  [[ -n "$LEDGER_DIR" ]] || return 0
  local -
  set +e
  _lg_begin "$1" "?"
  _lg_builder "$LEDGER_DIR/progress/$LG_CALL.bld1"
  if [[ "${3:-}" =~ ^[A-Za-z0-9_.-]{1,64}$ ]]; then
    printf '%s\t%s\t%s\t%s\n' "$LG_CALL" "$LG_JOB" "$LG_SUITEV" "$LG_KINDV" >"$LEDGER_DIR/watch/$3"
  else
    _lg_unknown_row "$LG_CALL" "ledger_watch_begin: invalid token"; printf 'F\t%s\n' "$LG_CALL" >>"$LEDGER_DIR/raw.tsv"
  fi
  return 0
}
ledger_watch_end() {
  [[ -n "$LEDGER_DIR" ]] || return 0
  local -
  set +e
  local id="" t="$LEDGER_DIR/watch/${3:-?}"
  if [[ "${3:-}" =~ ^[A-Za-z0-9_.-]{1,64}$ && -f "$t" ]]; then
    IFS=$'\t' read -r id LG_JOB LG_SUITEV LG_KINDV <"$t"; rm -f -- "$t"
  fi
  if [[ -z "$id" ]]; then
    _lg_begin "$1" 1; _lg_unknown_row "$LG_CALL" "ledger_watch_end without its ledger_watch_begin"; printf 'F\t%s\n' "$LG_CALL" >>"$LEDGER_DIR/raw.tsv"; return 0
  fi
  _lg_block_end "$id" "$1" - "${2:-}" "$4" "$5"
  return 0
}

# _lg_images: the run's image list, computed exactly as final_checks writes images.txt.
_lg_images() {
  { grep -h '^FROM ' "$REPO"/infra/docker/*.Dockerfile | awk '{print $2}' | grep -E '/|:'
    grep -h 'image: ' "$REPO"/infra/compose/*.yml | awk '{print $2}'
    printf '%s\n' "$YQ_IMAGE" "$BUSYBOX_IMAGE" "$SHELLCHECK_IMAGE" "$HADOLINT_IMAGE" golang:1.27.3-alpine; } | LC_ALL=C sort -u
}

# _lg_boot_row: LG_BOOT and, when PLATFORM_LEDGER_BOOTSTRAP_FILE names the hosted adapter's preflight record, the
# seq-0 bootstrap row: kind bootstrap, Dockerfile infra/docker/ci-context.Dockerfile (stdin, no context), its two
# pinned bases, platform and image ID; steps not_observed (counted as one executing build, never a hit). The file
# is read by column name (ref, pre_id or image_id, dockerfile_sha256, bases, platform). Locally: not-applicable;
# a hosted job (GITLAB_CI or CI_JOB_ID) without the file: missing.
_lg_boot_row() {
  local f="${PLATFORM_LEDGER_BOOTSTRAP_FILE:-}" df="$REPO/infra/docker/ci-context.Dockerfile" hdr row i ref="" pre="" dsha="" bases="" plat="" want refs got sha
  local -A lgcol=()
  LG_BOOT=not-applicable
  if [[ -z "$f" ]]; then [[ -z "${GITLAB_CI:-}${CI_JOB_ID:-}" ]] || LG_BOOT=missing; return 0; fi
  LG_BOOT=invalid
  if { IFS= read -r hdr && IFS= read -r row; } 2>/dev/null <"$f"; then
    _lg_split "$hdr"; for i in "${!LG_F[@]}"; do lgcol[${LG_F[$i]}]="$i"; done
    _lg_split "$row"
    [[ -n "${lgcol[ref]+x}" ]] && ref="${LG_F[${lgcol[ref]}]:-}"
    if [[ -n "${lgcol[pre_id]+x}" ]]; then pre="${LG_F[${lgcol[pre_id]}]:-}"; elif [[ -n "${lgcol[image_id]+x}" ]]; then pre="${LG_F[${lgcol[image_id]}]:-}"; fi
    [[ -n "${lgcol[dockerfile_sha256]+x}" ]] && dsha="${LG_F[${lgcol[dockerfile_sha256]}]:-}"
    [[ -n "${lgcol[bases]+x}" ]] && bases="${LG_F[${lgcol[bases]}]:-}"
    [[ -n "${lgcol[platform]+x}" ]] && plat="${LG_F[${lgcol[platform]}]:-}"
  fi
  sha="$(_lg_sha "$df")" || sha=""
  [[ -n "$sha" && "$dsha" == "$sha" ]] || dsha=unknown
  [[ "$pre" =~ ^sha256:[0-9a-f]{64}$ ]] || pre=unknown
  [[ "$plat" =~ ^[a-z0-9]+/[a-z0-9_]+(/[a-z0-9]+)?$ ]] || plat=unknown
  [[ "$ref" =~ ^platform-ci-context:[A-Za-z0-9_.-]+$ ]] || pre=unknown
  _lg_model "$df" "" ""
  want="$(awk -F'\t' '$1 == "base" && $4 == "from" { print $2 "=" $3 }' "$LG_MODEL" | LC_ALL=C sort | paste -sd, -)"
  refs="$(awk -F'\t' '$1 == "base" { print $3 }' "$LG_MODEL" | LC_ALL=C sort -u | paste -sd' ' -)"
  got="$(_lg_refset "$bases")"
  if [[ -z "$want" || -z "$got" || "$got" != "$refs" ]] || grep -qE '^(error|syntax)' "$LG_MODEL" || grep -qvE '@sha256:[0-9a-f]{64}$' <<<"${refs// /$'\n'}"; then want=unknown; fi
  if [[ "$dsha$pre$plat$want" != *unknown* ]]; then LG_BOOT=present; fi
  _lg_callinfo authoritative
  printf '0\t%s\tpreflight\tbootstrap\tinfra/docker/ci-context.Dockerfile\t%s\tfinal\tbuiltin\t%s\t%s\tnone\tnone\tnone\tnone\tnone\t%s/not_observed\tnot_observed\tnot_observed\tnot_observed\t0\t%s\n' \
    "$LG_JOB" "$dsha" "$want" "$plat" "$LEDGER_DAEMON" "$pre"
}
_lg_refset() { # TEXT: the sorted, space-separated set of NAME@sha256:<64 hex> references in TEXT, normalized
  local r
  grep -oE '[^[:space:],;|=+]+@sha256:[0-9a-f]{64}' <<<"$1" | while IFS= read -r r; do _lg_normref "$r"; printf '%s\n' "$LG_REF"; done | LC_ALL=C sort -u | paste -sd' ' -
}

# _lg_compute BUILDS IMAGES NOTES BOOTSTATE PUBLISHED-SUMMARY: the ledger-summary.tsv lines; near-duplicate pairs
# carry their differing paths (inventories of this run, else the published summary's list).
_lg_compute() {
  local builds="$1" imgs="$2" notes="$3" boot="$4" pub="${5:-}" bootok=0 imgmiss=0 o line a b x st ia ib na nb paths
  case "$boot" in not-applicable|present) bootok=1 ;; esac
  [[ -f "$imgs" ]] || { imgmiss=1; imgs=/dev/null; }
  o="$(awk -v hdr="$LEDGER_HEADER" -v exc="$LEDGER_EXCEPTIONS" -v imgmissing="$imgmiss" -v bootok="$bootok" -v bootstate="$boot" \
    "$LEDGER_AWK_LIB$LEDGER_CHECK_AWK" "$imgs" "$notes" "$builds")" || return 2
  printf '%s\n' "$o" | awk -F'\t' -v keys=" $LEDGER_SUMMARY_KEYS " 'index(keys, " " $1 " ")'
  while IFS= read -r line; do
    case "$line" in
      duplicate$'\t'*|unknown_row$'\t'*) printf '%s\n' "$line" ;;
      ND$'\t'*)
        IFS=$'\t' read -r x a b st ia ib na nb <<<"$line"
        paths=""
        if [[ -n "$LEDGER_DIR" && -f "$LEDGER_DIR/inv/$ia" && -f "$LEDGER_DIR/inv/$ib" ]]; then
          paths="$(_lg_inv_diff "$ia" "$ib" "")$(_lg_named_diff "$na" "$nb")"
        elif [[ -n "$pub" && -f "$pub" ]]; then
          paths="$(awk -F'\t' -v k="$a $b " '$1 == "near_duplicate" && index($2, k) == 1 { s = $2; sub(/^[^=]*paths=/, "", s); print s; exit }' "$pub")"
        fi
        paths="${paths# }"; printf 'near_duplicate\t%s %s %s paths=%s\n' "$a" "$b" "$st" "${paths:-unavailable}" ;;
    esac
  done <<<"$o"
}
# _lg_inv_diff A B PREFIX: up to 20 paths (shell-quoted, PREFIX added) whose inventory entries differ.
_lg_inv_diff() {
  local LC_ALL=C it rest lp p n=0 more=0
  local -A lgA=() lgB=() lgD=()
  while IFS= read -r -d '' it; do rest="${it#* * * * }"; lp="${rest%%:*}"; p="${rest:${#lp}+1:lp}"; lgA["$p"]="$it"; done <"$LEDGER_DIR/inv/$1"
  while IFS= read -r -d '' it; do rest="${it#* * * * }"; lp="${rest%%:*}"; p="${rest:${#lp}+1:lp}"; lgB["$p"]="$it"; done <"$LEDGER_DIR/inv/$2"
  for p in "${!lgA[@]}" "${!lgB[@]}"; do [[ "${lgA[$p]-}" == "${lgB[$p]-}" ]] || lgD["$p"]=1; done
  while IFS= read -r -d '' p; do
    n=$((n + 1)); if (( n > 20 )); then more=$((more + 1)); continue; fi
    printf ' %q' "$3$p"
  done < <(for p in "${!lgD[@]}"; do printf '%s\0' "$p"; done | LC_ALL=C sort -z)
  (( more == 0 )) || printf ' +%s more' "$more"
}
_lg_named_diff() { # NAMED-A NAMED-B: differing paths of named contexts (name:path)
  local e n x y
  local -A lgA=() lgB=()
  for e in ${1//,/ }; do [[ "$e" == *=* ]] && lgA[${e%%=*}]="${e#*=}"; done
  for e in ${2//,/ }; do [[ "$e" == *=* ]] && lgB[${e%%=*}]="${e#*=}"; done
  for n in $(printf '%s\n' "${!lgA[@]}" "${!lgB[@]}" | LC_ALL=C sort -u); do
    x="${lgA[$n]:-}" y="${lgB[$n]:-}"
    [[ "$x" != "$y" ]] || continue
    if [[ -f "$LEDGER_DIR/inv/$x" && -f "$LEDGER_DIR/inv/$y" ]]; then _lg_inv_diff "$x" "$y" "$n:"; else printf ' %q' "$n:*"; fi
  done
}

# ledger_check: writes $WORK/builds.tsv and $WORK/ledger-summary.tsv from this run's rows (with the bootstrap row) and
# records BUILD-LEDGER INFO with the summary, or BUILD-LEDGER FAIL when the ledger cannot be computed (§3.8.4).
ledger_check() {
  local msg rc=0
  msg="$(_lg_check_run 2>&1)" || rc=$?
  if (( rc == 0 )) && [[ -n "$msg" ]]; then record BUILD-LEDGER INFO "$msg"; else record BUILD-LEDGER FAIL "build ledger not computed: ${msg:-no output}"; fi
  return 0
}
_lg_check_run() {
  local -
  set +e
  [[ -n "$LEDGER_DIR" && -f "$LEDGER_DIR/raw.tsv" ]] || { printf 'ledger_init was not run'; return 1; }
  local out="$WORK/builds.tsv" sum="$WORK/ledger-summary.tsv" notes="$LEDGER_DIR/notes.tsv" imgs="${LEDGER_IMAGES_FILE:-}" k v s=""
  : >"$notes" || return 1
  { printf '%s\n' "$LEDGER_HEADER"; _lg_boot_row; awk -v notes="$notes" "$LEDGER_ASSEMBLE_AWK" "$LEDGER_DIR/raw.tsv"; } >"$out.tmp" || return 1
  mv -f -- "$out.tmp" "$out" || return 1
  if [[ -z "$imgs" ]]; then imgs="$LEDGER_DIR/images.txt"; _lg_images >"$imgs" || return 1; fi
  _lg_compute "$out" "$imgs" "$notes" "$LG_BOOT" "" >"$sum.tmp" || { printf 'malformed builds.tsv'; return 1; }
  # G051 C6: the unknown calls' bounded progress context (_lg_context), at most 120 lines, after the summary lines.
  if [[ -s "$LEDGER_DIR/context.tsv" ]]; then
    awk -F'\t' 'NR <= 120 { print "unknown_context\tcall " $1 " session " $2 ": " $3 } END { if (NR > 120) print "unknown_context\t" (NR - 120) " further context lines not kept" }' \
      "$LEDGER_DIR/context.tsv" >>"$sum.tmp" || return 1
  fi
  mv -f -- "$sum.tmp" "$sum" || return 1
  while IFS=$'\t' read -r k v; do
    case " $LEDGER_SUMMARY_KEYS " in *" $k "*) s+="${s:+, }$k $v" ;; esac
  done <"$sum"
  printf '%s (daemon %s, %s; builds.tsv, ledger-summary.tsv)' "$s" "$LEDGER_DAEMON" "$LEDGER_PLATFORM"
}

# ledger_check_file DIR: the same computation over a published job directory (builds.tsv, images.txt,
# ledger-bootstrap.tsv; ledger-summary.tsv only for notes and near-duplicate paths). Prints the recomputed
# ledger-summary lines (AC5 in ac5_status); returns 0 when computed, 1 when a ledger-summary.tsv in DIR disagrees
# with the recomputation, 2 when it cannot be computed. The caller compares and gates (pipeline-union.sh --ledger).
ledger_check_file() {
  local dir="$1" t rc=0 boot nb ba pre post dsha bases plat verdict row k v w hdr line
  local -
  set +e
  [[ -f "$dir/builds.tsv" ]] || { echo "ledger_check_file: no builds.tsv in $dir" >&2; return 2; }
  t="$(mktemp -d)" || return 2
  awk -F'\t' '$1 == "unknown_row" { s = $2; i = index(s, " "); if (i) print substr(s, 1, i - 1) "\t" substr(s, i + 1) }' "$dir/ledger-summary.tsv" >"$t/notes" 2>/dev/null
  nb="$(awk -F'\t' 'NR > 1 && $4 == "bootstrap"' "$dir/builds.tsv" | wc -l)"
  row="$(awk -F'\t' 'NR > 1 && $4 == "bootstrap" { print $6 "\t" $9 "\t" $10 "\t" $21; exit }' "$dir/builds.tsv")"
  boot=not-applicable
  if [[ -f "$dir/ledger-bootstrap.tsv" ]]; then
    { IFS= read -r hdr; IFS= read -r line; } <"$dir/ledger-bootstrap.tsv"
    _lg_split "${line:-}"
    pre="${LG_F[1]:-}" post="${LG_F[2]:-}" dsha="${LG_F[3]:-}" bases="${LG_F[4]:-}" plat="${LG_F[5]:-}" verdict="${LG_F[6]:-}"
    if [[ "$hdr" != $'ref\tpre_id\tpost_id\tdockerfile_sha256\tbases\tplatform\tverdict' ]]; then boot=invalid
    elif (( nb == 0 )); then [[ "$verdict" == not-applicable ]] || boot=missing
    elif (( nb > 1 )); then boot=invalid
    else
      _lg_split "$row"
      k="${LG_F[0]:-}" ba="${LG_F[1]:-}" w="${LG_F[2]:-}" v="${LG_F[3]:-}"
      boot="${verdict:-invalid}"
      if [[ "$verdict" == structural-inference:reuse ]]; then
        # Verified: the in-run state "present" stands (a structural inference of reuse, never step evidence).
        boot=present
        [[ "$pre" == "$v" && "$post" == "$pre" && "$dsha" == "$k" && "$plat" == "$w" ]] || boot=mismatch
        [[ -n "$bases" && "$(_lg_refset "$bases")" == "$(_lg_refset "$ba")" ]] || boot=mismatch
      fi
    fi
  elif (( nb > 0 )); then boot=missing
  fi
  _lg_compute "$dir/builds.tsv" "$dir/images.txt" "$t/notes" "$boot" "$dir/ledger-summary.tsv" >"$t/sum" || { rm -rf -- "$t"; echo "ledger_check_file: malformed builds.tsv" >&2; return 2; }
  cat -- "$t/sum"
  if [[ -f "$dir/ledger-summary.tsv" ]]; then
    for k in $LEDGER_SUMMARY_KEYS; do
      v="$(awk -F'\t' -v k="$k" '$1 == k { print $2; exit }' "$t/sum")"
      w="$(awk -F'\t' -v k="$k" '$1 == k { print $2; exit }' "$dir/ledger-summary.tsv")"
      [[ "$v" == "$w" ]] || { echo "ledger_check_file: $k recomputed '$v', published '${w:-absent}'" >&2; rc=1; }
    done
  fi
  rm -rf -- "$t"
  return "$rc"
}

# --- Docker inventory and cleanup ------------------------------------------------------
docker_inventory() {
  { docker ps -a --format 'container {{.Names}}'; docker volume ls --format 'volume {{.Name}}'
    docker network ls --format 'network {{.Name}}'; docker image ls --format 'image {{.Repository}}:{{.Tag}}'; } | LC_ALL=C sort
}

cleanup() {
  local p img
  # Positive ownership evidence first, while the run's resources still exist (G003 §11.4).
  if [[ -n "${OWNED_REG:-}" && -f "$WORK/projects" ]]; then
    while IFS= read -r p; do
      [[ -n "$p" ]] || continue
      for img in "$REPO" "$WORK"/*; do
        [[ -f "$img/.env" && "$(sed -n 's/^COMPOSE_PROJECT_NAME=//p' "$img/.env" | tail -1)" == "$p" ]] && owned_capture "$img"
      done
    done < <(LC_ALL=C sort -u "$WORK/projects")
  fi
  # G051 attempt 5 (a): owned teardown first (every registered Compose project, its one-off containers included; a
  # standalone --rm container is removed by the daemon once it ends), then a bounded wait (about 15 s) until no
  # container still uses an image this run built. Only then are those images removed, never with -f. A timeout, an
  # inspection failure or a failed removal of an image not shown to be gone keeps the images, is logged and fails the
  # run unless an earlier failure already set its status. Images: those built with -t (g003-*, g001-*) and those
  # built under a registered project's name, in both cases only when absent before the run.
  local -a imgs=()
  local -A seen_img=()
  local busy out i cfail=0
  if [[ -f "$WORK/projects" ]]; then
    while IFS= read -r p; do
      [[ -n "$p" ]] || continue
      docker compose -p "$p" down -v --remove-orphans >/dev/null 2>&1 || true
    done < <(LC_ALL=C sort -u "$WORK/projects")
  fi
  if [[ -f "$WORK/images-built" ]]; then
    while IFS= read -r img; do
      [[ -n "$img" ]] || continue
      grep -qxF "image $img" "$WORK/inventory-start.txt" || { [[ -n "${seen_img[$img]:-}" ]] || { seen_img[$img]=1; imgs+=("$img"); }; }
    done < <(LC_ALL=C sort -u "$WORK/images-built")
  fi
  if [[ -f "$WORK/projects" ]]; then
    while IFS= read -r p; do
      [[ -n "$p" ]] || continue
      for img in $(docker image ls --format '{{.Repository}}:{{.Tag}}' | grep -E "^${p}-[a-z-]+:" || true); do
        # A unit suite records its project image too: each image is listed once.
        grep -qxF "image $img" "$WORK/inventory-start.txt" || { [[ -n "${seen_img[$img]:-}" ]] || { seen_img[$img]=1; imgs+=("$img"); }; }
      done
    done < <(LC_ALL=C sort -u "$WORK/projects")
  fi
  busy=0
  for ((i = 0; i < 150 && ${#imgs[@]} > 0; i++)); do
    busy=0
    for img in "${imgs[@]}"; do
      if out="$(docker ps -aq --filter "ancestor=$img" 2>/dev/null)"; then [[ -z "$out" ]] || busy=1; else busy=2; break; fi
    done
    (( busy == 1 )) || break
    sleep 0.1
  done
  if (( busy == 0 )); then
    for img in "${imgs[@]}"; do
      docker image rm "$img" >/dev/null 2>&1 && continue
      # G051 C1: a failed removal is accepted only when Docker positively shows the image gone (an answered query
      # without a match; it may already be gone). An image still present, or a query that fails, fails the run.
      if ! out="$(docker image ls -q --filter "reference=$img" 2>/dev/null)"; then
        log "cleanup: image $img could not be removed and Docker could not confirm it is gone (unverified)"; cfail=1
      elif [[ -n "$out" ]]; then
        log "cleanup: image $img could not be removed (kept)"; cfail=1
      fi
    done
  elif (( busy == 1 )); then
    log "cleanup: containers still use images this run built after 15 s; the images are kept: ${imgs[*]}"; cfail=1
  else
    log "cleanup: Docker could not be inspected; the images this run built are kept: ${imgs[*]}"; cfail=1
  fi
  if [[ -n "${ROOT_ENV_SHA:-}" && -f "$REPO/.env" && "$(sha_of "$REPO/.env")" == "$ROOT_ENV_SHA" ]]; then
    rm -f "$REPO/.env"
  fi
  if [[ -f "$WORK/inventory-start.txt" ]]; then
    docker_inventory >"$WORK/inventory-end.txt"
    comm -13 "$WORK/inventory-start.txt" "$WORK/inventory-end.txt" | grep -v '^image ' >"$WORK/inventory-leftover.txt" || true
    [[ -s "$WORK/inventory-leftover.txt" ]] && log "resources created and not removed: $(tr '\n' ' ' <"$WORK/inventory-leftover.txt")"
    [[ -n "${REPORT_DIR:-}" ]] && cp "$WORK/inventory-leftover.txt" "$REPORT_DIR/" 2>/dev/null
  fi
  if [[ "${KEEP:-0}" == 1 ]]; then
    log "keeping $WORK"
  else
    rm -rf "$WORK" 2>/dev/null || docker run --rm --label g001.run=$RUN_ID -v "$(dirname "$WORK"):/w" "$BUSYBOX_IMAGE" rm -rf "/w/$(basename "$WORK")" || true
  fi
  # G051 §3.6: a background child that bg_stop_all had to stop on an otherwise successful exit fails the run.
  if [[ "${BG_FORCE_FAIL:-0}" == 1 ]]; then exit 1; fi
  # G051 attempt 5 (a): an unverified image teardown fails an otherwise successful run (RUN_STATUS: the trap's entry
  # status); an earlier failure keeps its own status.
  if (( cfail )) && [[ "${RUN_STATUS:-0}" == 0 ]]; then exit 1; fi
}
