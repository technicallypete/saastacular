#!/bin/sh
# Hosted CI adapter (G003 CI revision Plan r2 §3.1; Spec §12; docs/foundation/OPERATIONS.md). POSIX sh for the
# job's docker:cli shell, used only inside a GitLab CI job beside its job-private TLS Docker daemon (service
# alias "docker"). It removes credentials from the job-created Git configuration and proves no literal job
# credential remains in the build directory BEFORE any nested container mounts it, then runs the unchanged
# entry point (infra/ci/validate.sh) in an unprivileged context container that shares the daemon's network
# namespace and sees the build directory at the identical path. A disk guard samples the shared filesystem
# at most every 30 s and fails the job below 4 GiB available, on a failed measurement or when the guard dies,
# independently of the entry point's own status.
#
#   sh infra/ci/hosted-context.sh -- VALIDATE_ARGS...     (job script)
#   sh infra/ci/hosted-context.sh --after-script          (job after_script; always exits 0)
#
# Identity staging (G004 CI-1; Plan r5 §8.5): the WorkOS staging API key, client ID and issuer reach the context
# only for --profile staging in the exactly approved context (the approved project, the protected
# identity-qualification branch, a web or scheduled pipeline without a merge request, job and environment
# identity-staging), as one mode-600 file outside the build directory mounted read-only. The file reaches the job
# daemon through a job-scoped volume filled over a helper container's stdin, never as a bind-mount source: the CLI
# talks to the job-private daemon over TCP, so a -v source path would name the daemon's filesystem, not this job's.
# The volume is removed when the job ends and by the after_script backstop. Only the API key joins the credential
# patterns. A staging profile anywhere else is refused, and any IDENTITY_STAGING_WORKOS_* variable
# visible to another profile fails closed before a context starts.
#
# Exit: the entry point's status; 16 outside a GitLab CI job with the job-private TLS daemon, on an unsafe
# temporary path, on a credential that cannot be proven absent, on a failed context build, on a staging profile
# outside the approved context or on a staging variable visible to another profile; 20 when the disk guard fails
# (below 4 GiB, measurement failure or guard death) and the entry point reported success.
# No secret value is ever printed, passed as an argument or written below the build directory.
set -u
set +x
umask 077

PFX="hosted-context"
say() { printf '%s: %s\n' "$PFX" "$*"; }
CTX_IMAGE_FILE="infra/docker/ci-context.Dockerfile"
GUARD_IMAGE="busybox:1.37.0"
FLOOR_K=4194304          # 4 GiB in KiB: the runtime disk reserve
HOSTED_FLOOR_GIB=35      # provisional hosted start floor (infra/ci/lib.sh CI_DISKGIB_HOSTED)
INTERVAL="${PLATFORM_CI_GUARD_INTERVAL:-30}"   # guard sampling interval, seconds (1-30; the controls shorten it)
case "$INTERVAL" in ''|*[!0-9]*) INTERVAL=30 ;; esac
if [ "$INTERVAL" -lt 1 ] || [ "$INTERVAL" -gt 30 ]; then INTERVAL=30; fi
STOP_GRACE=150           # context stop grace, seconds (as validate.sh's harness grace)

MODE=main
case "${1:-}" in
  --after-script) MODE=after; shift ;;
  --) shift ;;
  *) say "usage: hosted-context.sh -- VALIDATE_ARGS... | --after-script"; exit 16 ;;
esac

# --- G1: only in a GitLab CI job with the job-private TLS daemon (no Docker call otherwise) -------------
ci_ok() {
  [ "${GITLAB_CI:-}" = true ] || return 1
  case "${CI_JOB_ID:-}" in ''|*[!0-9]*) return 1 ;; esac
  case "${CI_PROJECT_DIR:-}" in /*) ;; *) return 1 ;; esac
  [ -d "$CI_PROJECT_DIR" ] || return 1
  [ "${DOCKER_HOST:-}" = tcp://docker:2376 ] || return 1
  [ "${DOCKER_TLS_VERIFY:-}" = 1 ] || return 1
  [ "${DOCKER_CERT_PATH:-}" = /certs/client ] || return 1
  [ "${DOCKER_TLS_CERTDIR:-}" = /certs ] || return 1
  return 0
}
if ci_ok; then :; else
  say "refused: not a GitLab CI job with the job-private TLS daemon (no Docker call made)"
  if [ "$MODE" = after ]; then exit 0; fi
  exit 16
fi
NS="$(dirname "$CI_PROJECT_DIR")"
case "$NS" in /|/*/) say "refused: unsafe build directory parent"; [ "$MODE" = after ] && exit 0; exit 16 ;; esac
CTX_NAME="platform-ci-$CI_JOB_ID"
CTX_IMAGE="platform-ci-context:$CI_JOB_ID"

# --- identity staging forwarding gate (G004 CI-1) ------------------------------------------------------------
# The approved project: its numeric ID and the SHA-256 of its path (the path is not written here).
# Template placeholders (synthetic example project 999999999 at example-group/example-project/platform): every real
# project is refused until an adopter re-pins both lines to their own GitLab project, keeping each line exactly in this
# form (staging-action.sh, infra/terraform/bootstrap/ci-federation.sh and the tests read ^STG_PROJECT_ID=[1-9][0-9]*$
# and ^STG_PROJECT_PATH_SHA=[0-9a-f]{64}$):
#   STG_PROJECT_ID=<numeric project ID>; STG_PROJECT_PATH_SHA=$(printf '%s' '<full/project/path>' | sha256sum)
# Re-pin the identity-staging controls in infra/ci/validate-controls.sh with them (g_stg_path, g_stg_env) and see
# docs/foundation/STAGING.md "Pinning your GitLab project".
STG_PROJECT_ID=999999999
STG_PROJECT_PATH_SHA=c2c0591cd119ad1562af3d324064601386019214296d759fbc576f606b2b6fe3
STG_MOUNT=/run/platform-identity-staging.env
STG_FILE=""
STG_VOL=""
requested_profile() { # the --profile value among the validate.sh arguments ("" when absent)
  rp="" prev=""
  for a in "$@"; do if [ "$prev" = --profile ]; then rp="$a"; fi; prev="$a"; done
  printf '%s\n' "$rp"
}
stg_visible() { env | grep -q '^IDENTITY_STAGING_WORKOS_[A-Z_]*='; }
stg_approved() {
  [ "${CI_PROJECT_ID:-}" = "$STG_PROJECT_ID" ] || return 1
  [ "$(printf '%s' "${CI_PROJECT_PATH:-}" | sha256sum | cut -d' ' -f1)" = "$STG_PROJECT_PATH_SHA" ] || return 1
  [ "${CI_COMMIT_BRANCH:-}" = identity-qualification ] || return 1
  [ "${CI_COMMIT_REF_PROTECTED:-}" = true ] || return 1
  case "${CI_PIPELINE_SOURCE:-}" in web|schedule) ;; *) return 1 ;; esac
  [ -z "${CI_MERGE_REQUEST_IID:-}" ] || return 1
  [ "${CI_JOB_NAME:-}" = identity-staging ] || return 1
  [ "${CI_ENVIRONMENT_NAME:-}" = identity-staging ] || return 1
  return 0
}
stg_forward() { # writes the present staging values (names fixed, values never printed) to a mode-600 file in /tmp
  STG_FILE="$(mktemp /tmp/hosted-context-staging.XXXXXX)" || return 1
  chmod 600 "$STG_FILE" || return 1
  case "$STG_FILE" in "$NS"/*) return 1 ;; esac
  n=0
  for k in IDENTITY_STAGING_WORKOS_API_KEY IDENTITY_STAGING_WORKOS_CLIENT_ID IDENTITY_STAGING_WORKOS_ISSUER; do
    eval "v=\"\${$k:-}\""
    if [ -n "$v" ]; then printf '%s=%s\n' "$k" "$v" >>"$STG_FILE"; n=$((n + 1)); fi
  done
  say "identity staging: approved context; $n of 3 inputs forwarded as one read-only file (values not shown)"
  return 0
}
stg_deliver() { # the file to the job daemon: a labelled volume filled through a helper's stdin (no argv, no env)
  STG_VOL="platform-ci-stg-$CI_JOB_ID"
  docker volume create --label "platform.ci.staging=$CI_JOB_ID" "$STG_VOL" >/dev/null 2>&1 || return 1
  docker run --rm -i --network none --name "$CTX_NAME-stg" --label "platform.ci.staging=$CI_JOB_ID" -v "$STG_VOL:/d" \
    "$CTX_IMAGE" sh -c 'umask 077 && cat >/d/env && test -s /d/env' <"$STG_FILE" >/dev/null 2>&1
}
stg_remove() { if [ -n "$STG_VOL" ]; then docker volume rm -f "$STG_VOL" >/dev/null 2>&1 || true; fi; }

# --- G5: the context's temporary directory: new, mode 700, a real directory directly under $NS ------------
make_tmp() { # NAME -> prints the path
  t="$NS/$1"
  if [ -e "$t" ] || [ -L "$t" ]; then return 1; fi
  mkdir -m 700 "$t" 2>/dev/null || return 1
  if [ -L "$t" ] || [ ! -d "$t" ]; then return 1; fi
  if [ "$(cd "$t" 2>/dev/null && pwd -P)" != "$(cd "$NS" && pwd -P)/$1" ]; then return 1; fi
  printf '%s\n' "$t"
}

# --- credentials: job-local secret patterns, never mounted, never in argv or logs -------------------------
PF=""
cleanup_pf() { if [ -n "$PF" ]; then rm -f "$PF" "$PF.inc" "$PF.hits" "$PF.err"; fi; PF=""; }
make_patterns() { # writes the non-empty job credentials (at least 8 bytes) to a mode-600 file in /tmp
  PF="$(mktemp /tmp/hosted-context-patterns.XXXXXX)" || return 1
  chmod 600 "$PF" || return 1
  case "$PF" in "$NS"/*) return 1 ;; esac
  n=0
  for v in "${CI_JOB_TOKEN:-}" "${CI_REGISTRY_PASSWORD:-}" "${CI_DEPENDENCY_PROXY_PASSWORD:-}" "${IDENTITY_STAGING_WORKOS_API_KEY:-}"; do
    if [ "${#v}" -ge 8 ]; then printf '%s\n' "$v" >>"$PF"; n=$((n + 1)); fi
  done
  # Without the job token there is nothing to prove absent: fail closed.
  jt="${CI_JOB_TOKEN:-}"
  [ "${#jt}" -ge 8 ] || return 1
  [ "$n" -ge 1 ] || return 1
  return 0
}

# sanitize_config FILE: remove the password from every remote url/pushurl of FILE (Git config syntax),
# keeping the scheme, user name, host and path. Unparseable or unsafe entries fail closed (nonzero, no write).
# Prints "remotes=N sanitized=M includes=K" with the include paths on later lines.
SAN_AWK='
function trim(s) { sub(/^[ \t]+/, "", s); sub(/[ \t\r]+$/, "", s); return s }
BEGIN { sec = ""; remotes = 0; sanitized = 0; includes = 0; bad = 0 }
{
  line = $0
  if (line ~ /^[ \t]*\[/) {
    h = line; sub(/^[ \t]*\[/, "", h); sub(/\][ \t]*$/, "", h)
    sec = tolower(h); sub(/[ \t"].*$/, "", sec)
    # A credential-bearing URL in a section name (for example url "https://user:pass@host/".insteadOf).
    if (h ~ /[A-Za-z][A-Za-z0-9+.-]*:\/\/[^\/@ \t"]*:[^\/@ \t"]*@/) { bad = 1 }
    out[++n] = line; next
  }
  if (line ~ /^[ \t]*([#;].*)?$/) { out[++n] = line; next }
  k = line; sub(/=.*$/, "", k); k = tolower(trim(k))
  v = line; if (sub(/^[^=]*=/, "", v) == 0) { v = "" } ; v = trim(v)
  if (sec == "remote" && (k == "url" || k == "pushurl")) {
    remotes++
    if (v ~ /["\\]/) { bad = 1; out[++n] = line; next }
    if (match(v, /^[A-Za-z][A-Za-z0-9+.-]*:\/\/[^\/@:]*:[^\/@]*@/)) {
      pre = v; sub(/:\/\/.*$/, "://", pre)
      rest = substr(v, length(pre) + 1)
      user = rest; sub(/:.*$/, "", user)
      host = rest; sub(/^[^@]*@/, "", host)
      ind = line; sub(/[^ \t].*$/, "", ind)
      out[++n] = ind k " = " pre user "@" host; sanitized++; next
    }
    out[++n] = line; next
  }
  if ((sec == "include" || sec == "includeif") && k == "path") {
    if (v ~ /["\\]/ || v ~ /^~/) { bad = 1 }
    includes++; inc[includes] = v
  }
  out[++n] = line
}
END {
  if (bad) { exit 3 }
  for (i = 1; i <= n; i++) print out[i] > target
  printf "remotes=%d sanitized=%d includes=%d\n", remotes, sanitized, includes
  for (i = 1; i <= includes; i++) print inc[i]
}'
PASSWORD_URL_RE='[A-Za-z][A-Za-z0-9+.-]*://[^/@[:space:]"]*:[^/@[:space:]"]*@'

# prove_clean [SKIP_DIR]: sanitize .git/config, classify its includes, then prove that no password-bearing URL
# is left in the mounted configuration and that no job credential is stored in any regular file below $NS
# (SKIP_DIR, the job's own context temporary directory in after_script, is not read: the context never held a
# credential). 0 = proven clean.
prove_clean() {
  cfg="$CI_PROJECT_DIR/.git/config"
  NS_R="$(cd "$NS" 2>/dev/null && pwd -P)"
  if [ -z "$NS_R" ]; then say "credentials: the build directory parent cannot be resolved"; return 1; fi
  if [ -L "$cfg" ] || [ ! -f "$cfg" ] || [ ! -r "$cfg" ]; then say "credentials: .git/config missing or unreadable"; return 1; fi
  make_patterns || { say "credentials: no job credential to check against"; return 1; }
  new="$CI_PROJECT_DIR/.git/config.hosted-context.$$"
  if [ -e "$new" ] || [ -L "$new" ]; then say "credentials: scratch config path exists"; return 1; fi
  meta="$(awk -v target="$new" "$SAN_AWK" "$cfg")"
  rc=$?
  if [ "$rc" -ne 0 ]; then rm -f "$new"; say "credentials: unparseable or unsafe Git configuration (fail closed)"; return 1; fi
  if [ -L "$new" ] || [ ! -f "$new" ]; then say "credentials: sanitized configuration not written"; return 1; fi
  cat "$new" >"$cfg" || { rm -f "$new"; say "credentials: cannot rewrite .git/config"; return 1; }
  rm -f "$new"
  counts="$(printf '%s\n' "$meta" | head -n 1)"
  inside=0 outside=0 absent=0
  printf '%s\n' "$meta" | tail -n +2 >"$PF.inc" || return 1
  while IFS= read -r p; do
    [ -n "$p" ] || continue
    case "$p" in /*) f="$p" ;; *) f="$CI_PROJECT_DIR/.git/$p" ;; esac
    if [ ! -e "$f" ] && [ ! -L "$f" ]; then absent=$((absent + 1)); continue; fi
    rd="$(cd "$(dirname "$f")" 2>/dev/null && pwd -P)"
    if [ -z "$rd" ]; then rm -f "$PF.inc"; say "credentials: an include cannot be resolved"; return 1; fi
    rf="$rd/$(basename "$f")"
    case "$rf" in
      "$NS_R"/*)
        inside=$((inside + 1))
        if [ ! -r "$rf" ]; then rm -f "$PF.inc"; say "credentials: an included configuration is unreadable"; return 1; fi
        if grep -qiE '^[[:space:]]*\[?[[:space:]]*include' "$rf"; then rm -f "$PF.inc"; say "credentials: nested include (fail closed)"; return 1; fi
        if grep -qE "$PASSWORD_URL_RE" "$rf"; then rm -f "$PF.inc"; say "credentials: an included configuration holds a password-bearing URL"; return 1; fi ;;
      *) outside=$((outside + 1)) ;;
    esac
  done <"$PF.inc"
  rm -f "$PF.inc"
  if grep -qE "$PASSWORD_URL_RE" "$cfg"; then say "credentials: a password-bearing URL remains in .git/config"; return 1; fi
  # Every regular file below $NS, hidden and binary included (FIFOs, sockets and devices are never opened). Any
  # error output (an unreadable file or directory, a failed find or grep) means absence is not proven: busybox
  # grep -r reports an unreadable directory with exit 1, so the exit status alone is not enough.
  : >"$PF.err"
  if [ -n "${1:-}" ]; then find "$NS" -path "$1" -prune -o -type f -print0 2>>"$PF.err"
  else find "$NS" -type f -print0 2>>"$PF.err"; fi | xargs -0 -r grep -l -F -a -f "$PF" -- >"$PF.hits" 2>>"$PF.err"
  xrc=$?
  if [ -s "$PF.hits" ]; then say "credentials: a job credential is stored in $(grep -c . "$PF.hits") file(s) under the build directory (fail closed)"; return 1; fi
  # xargs: 0 when every grep matched nothing to report and exited 0, 123 when a grep found no match.
  if [ -s "$PF.err" ] || { [ "$xrc" -ne 0 ] && [ "$xrc" -ne 123 ]; }; then
    say "credentials: the build-directory scan failed or met unreadable entries (fail closed)"; return 1
  fi
  say "credentials: $counts include_inside=$inside include_outside=$outside include_absent=$absent password_urls=0 token_files=0 scan=clean"
  return 0
}

# --- the context ---------------------------------------------------------------------------------------
wait_daemon() {
  i=0
  until docker version >/dev/null 2>&1; do
    i=$((i + 1)); [ "$i" -ge 90 ] && return 1
    sleep 2
  done
  return 0
}
build_context() {
  if docker image inspect "$CTX_IMAGE" >/dev/null 2>&1; then return 0; fi
  docker build -q -t "$CTX_IMAGE" - <"$CI_PROJECT_DIR/$CTX_IMAGE_FILE" >/dev/null 2>&1
}
run_context() { # TMPDIR ARGS... : the context, in the foreground of the caller
  # The -e NAME list is public pipeline metadata only (ONEOFF RC qualification adds CI_PROJECT_ID, the project
  # attribution of the evidence); credentials never travel this way.
  ct="$1"; shift
  stg=""
  if [ -n "$STG_VOL" ]; then
    stg="--mount type=volume,src=$STG_VOL,dst=$STG_MOUNT,volume-subpath=env,readonly -e IDENTITY_STAGING_ENV_FILE=$STG_MOUNT"
  fi
  # shellcheck disable=SC2086 # $stg is empty or the staging mount and its path variable (no spaces)
  docker run --rm --init --name "$CTX_NAME" --network host --security-opt no-new-privileges $stg \
    -v "$NS:$NS" --mount "type=tmpfs,destination=$CI_PROJECT_DIR.tmp,tmpfs-size=1m" -v /certs/client:/certs/client:ro -w "$CI_PROJECT_DIR" \
    -e TMPDIR="$ct" -e DOCKER_HOST=tcp://localhost:2376 -e DOCKER_TLS_VERIFY=1 -e DOCKER_CERT_PATH=/certs/client \
    -e PLATFORM_CI_CONTEXT=hosted-context/1 \
    -e GITLAB_CI -e CI_PIPELINE_SOURCE -e CI_COMMIT_SHA -e CI_COMMIT_BRANCH -e CI_COMMIT_REF_NAME -e CI_COMMIT_TAG \
    -e CI_DEFAULT_BRANCH -e CI_MERGE_REQUEST_IID -e CI_MERGE_REQUEST_EVENT_TYPE -e CI_MERGE_REQUEST_SOURCE_BRANCH_NAME \
    -e CI_MERGE_REQUEST_TARGET_BRANCH_NAME -e CI_MERGE_REQUEST_SOURCE_BRANCH_SHA -e CI_MERGE_REQUEST_TARGET_BRANCH_SHA \
    -e CI_MERGE_REQUEST_DIFF_BASE_SHA -e CI_MERGE_REQUEST_APPROVED -e CI_PIPELINE_ID -e CI_PIPELINE_IID -e CI_PIPELINE_URL \
    -e CI_JOB_ID -e CI_JOB_URL -e GITLAB_USER_LOGIN -e CI_RUNNER_ID -e CI_RUNNER_DESCRIPTION -e CI_RUNNER_TAGS \
    -e CI_SERVER_VERSION -e CI_JOB_TIMEOUT -e CI_PROJECT_DIR -e CI_JOB_STARTED_AT -e CI_COMMIT_BEFORE_SHA \
    -e RUNNER_SCRIPT_TIMEOUT -e CI_COMMIT_REF_PROTECTED -e CI_PROJECT_ID \
    "$CTX_IMAGE" "$@"
}
stop_context() { docker stop -t "$STOP_GRACE" "$CTX_NAME" >/dev/null 2>&1 || true; }

# --- the disk guard ------------------------------------------------------------------------------------
GD=""          # job-local guard directory (/tmp, never mounted)
GFAIL=""       # the first guard failure, also kept in this shell (independent of a write to $GD/fail)
# Each measurement is time-limited; a guard whose samples stop (alive but stalled) fails too. The stall window is
# 420 s at the default interval: a hosted runner was observed to pause the guard for over 180 s during a heavy step (no
# low-space reading; the true length is unknown), so the former 180 s window stopped a healthy context. The 4 GiB
# floor and the per-measurement timeout are unchanged, and a guard that stays silent past the window still stops the
# context. The heartbeat is also written atomically and the supervisor ignores an empty read: a 420 s window once
# stopped a release-candidate job although the guard had sampled throughout (65 samples), because the
# supervisor read the file between the redirect's truncation and the write.
MEAS_T=$((INTERVAL * 2)); if [ "$MEAS_T" -lt 2 ]; then MEAS_T=2; fi
STALL_S=$((9 * INTERVAL + 2 * MEAS_T + 30))
# Daemon storage on a filesystem distinct from the build directory is sampled through one long-lived sampler
# container started before the workload (so it is in every start inventory and never a harness leftover).
GCTR="platform-ci-guard-$CI_JOB_ID" GCTR_ON=0
measure_ns() { timeout "$MEAS_T" df -Pk "$NS" 2>/dev/null | awk 'NR==2 && $2 ~ /^[0-9]+$/ && $4 ~ /^[0-9]+$/ {print $2, $4}'; }
measure_daemon() { # [SECONDS]
  if [ "$GCTR_ON" = 1 ]; then timeout "${1:-$MEAS_T}" docker exec "$GCTR" df -Pk / 2>/dev/null
  else timeout "${1:-$MEAS_T}" docker run --rm --network none --label "platform.ci.guard=$CI_JOB_ID" "$GUARD_IMAGE" df -Pk / 2>/dev/null; fi |
    awk 'NR==2 && $2 ~ /^[0-9]+$/ && $4 ~ /^[0-9]+$/ {print $2, $4}'
}
stop_sampler() { if [ "$GCTR_ON" = 1 ]; then docker rm -f "$GCTR" >/dev/null 2>&1 || true; GCTR_ON=0; fi; }
guard_fail() { # REASON: record the first failure (in this shell and in $GD/fail), then stop the context
  if [ -z "$GFAIL" ]; then GFAIL="$1"; fi
  if [ ! -s "$GD/fail" ]; then printf '%s\n' "$1" >"$GD/fail" 2>/dev/null || true; fi
  say "disk guard: FAIL ($1); stopping the context"
  stop_context
}
guard_note() { # the guard subshell failed (exit 3): carry its reason into this shell
  if [ -z "$GFAIL" ]; then GFAIL="$(head -n 1 "$GD/fail" 2>/dev/null)"; fi
  if [ -z "$GFAIL" ]; then GFAIL="guard failure"; fi
}
sample() { # appends "epoch ns_avail daemon_avail" or fails the guard
  a="$(measure_ns)"
  if [ -z "$a" ]; then guard_fail "measurement failed"; return 1; fi
  na="${a#* }"
  if [ "$(cat "$GD/shared")" = yes ]; then da="$na"; else
    b="$(measure_daemon)"
    if [ -z "$b" ]; then guard_fail "measurement failed"; return 1; fi
    da="${b#* }"
  fi
  printf '%s %s %s\n' "$(date +%s)" "$na" "$da" >>"$GD/samples"
  if [ "$na" -lt "$FLOOR_K" ] || [ "$da" -lt "$FLOOR_K" ]; then guard_fail "available space below 4 GiB"; return 1; fi
  return 0
}
guard_loop() { # runs in the background until the context ends: exit 0 only on the stop marker, 3 on a failure
  trap - EXIT TERM INT HUP
  while :; do
    sleep "$INTERVAL"
    if [ -e "$GD/stop" ]; then exit 0; fi
    sample || exit 3
    beat_now="$(date +%s)"
    printf '%s\n' "$beat_now" >"$GD/beat.tmp" 2>/dev/null && mv -f "$GD/beat.tmp" "$GD/beat" 2>/dev/null || true
  done
}
report_capacity() { # per filesystem: baseline, minimum, incremental peak and the qualification floor
  for col in 2 3; do
    if [ "$col" = 2 ]; then fs=build-directory; else fs=daemon-storage; fi
    if [ "$col" = 3 ] && [ "$(cat "$GD/shared")" = yes ]; then
      say "capacity: fs=daemon-storage shared_with=build-directory (same filesystem; not counted twice)"; continue
    fi
    base="$(awk -v c="$col" 'NR==1 {print $c}' "$GD/samples")"
    min="$(awk -v c="$col" 'NR==1 || $c < m {m = $c} END {print m}' "$GD/samples")"
    n="$(grep -c . "$GD/samples")"
    peak=$((base - min)); [ "$peak" -lt 0 ] && peak=0
    # durable floor = max(35 GiB, ceil(1.5 x incremental peak + 4 GiB)), in GiB
    need=$(( (3 * peak + 1) / 2 + FLOOR_K ))
    fg=$(( (need + 1048575) / 1048576 ))
    [ "$fg" -lt "$HOSTED_FLOOR_GIB" ] && fg="$HOSTED_FLOOR_GIB"
    meets=no; [ "$base" -ge $((fg * 1048576)) ] && meets=yes
    say "capacity: fs=$fs baseline_avail_bytes=$((base * 1024)) min_avail_bytes=$((min * 1024)) incremental_peak_bytes=$((peak * 1024)) samples=$n durable_floor_gib=$fg baseline_meets_durable_floor=$meets"
  done
}

# --- after_script: stop a remaining context, then the backstop cleanup in a fresh context (always 0) -------
if [ "$MODE" = after ]; then
  trap 'cleanup_pf' EXIT
  if wait_daemon; then :; else say "after-script: the job daemon is unreachable; nothing to do"; exit 0; fi
  if docker container inspect "$CTX_NAME" >/dev/null 2>&1; then say "after-script: stopping the remaining context"; stop_context; fi
  if docker container inspect "$GCTR" >/dev/null 2>&1; then GCTR_ON=1; stop_sampler; fi
  if [ "${CI_JOB_NAME:-}" = identity-staging ]; then STG_VOL="platform-ci-stg-$CI_JOB_ID"; stg_remove; STG_VOL=""; fi
  if prove_clean "$NS/.platform-tmp-$CI_JOB_ID"; then :; else cleanup_pf; say "after-script: credentials not proven absent; no context started"; exit 0; fi
  cleanup_pf
  if build_context; then :; else say "after-script: cannot build the context image"; exit 0; fi
  T="$(make_tmp ".platform-tmp-$CI_JOB_ID-after")" || { say "after-script: unsafe temporary directory"; exit 0; }
  CTX_NAME="platform-ci-$CI_JOB_ID-after"
  run_context "$T" bash infra/ci/cleanup-owned.sh --after-script || true
  rm -rf "$T" 2>/dev/null || true
  exit 0
fi

# --- job script --------------------------------------------------------------------------------------------
GD="$(mktemp -d /tmp/hosted-context-guard.XXXXXX)" || { say "cannot create the guard directory"; exit 16; }
SPID="" CPID="" SIG=""
finish() { cleanup_pf; stop_sampler; stg_remove; if [ -n "$GD" ]; then rm -rf "$GD"; fi; if [ -n "$STG_FILE" ]; then rm -f "$STG_FILE"; fi; }
on_signal() { # forwarded termination: stop the context (bounded), keep the signal status
  SIG="$1"
  say "signal $1 received; stopping the context"
  stop_context
}
trap 'finish' EXIT
trap 'on_signal TERM' TERM
trap 'on_signal INT' INT
trap 'on_signal HUP' HUP

PROF="$(requested_profile "$@")"
if [ "$PROF" = staging ]; then
  if stg_approved; then :; else say "refused: the staging profile outside the approved identity-staging context (nothing forwarded)"; exit 16; fi
  stg_forward || { say "refused: cannot write the identity staging file"; exit 16; }
elif stg_visible; then
  say "refused: an IDENTITY_STAGING_WORKOS_* variable is visible to the ${PROF:-default} profile (scope misconfiguration; nothing forwarded)"; exit 16
fi
if wait_daemon; then :; else say "the job daemon is unreachable"; exit 16; fi
if prove_clean; then :; else cleanup_pf; say "refused: credentials not proven absent; no context started"; exit 16; fi
cleanup_pf
T="$(make_tmp ".platform-tmp-$CI_JOB_ID")" || { say "refused: unsafe or pre-existing temporary directory"; exit 16; }
if build_context; then :; else say "cannot build the context image"; exit 16; fi
if [ -n "$STG_FILE" ]; then
  if stg_deliver; then :; else say "refused: cannot deliver the identity staging file to the job daemon; no context started"; exit 16; fi
fi

# Baseline after setup and before the workload; shared filesystem confirmed, not assumed.
a="$(measure_ns)"; b="$(measure_daemon 300)"
if [ -z "$a" ] || [ -z "$b" ]; then say "disk guard: FAIL (baseline measurement failed); no context started"; exit 20; fi
ns_size="${a% *}" ns_av="${a#* }" d_size="${b% *}" d_av="${b#* }"
diff=$((ns_av - d_av)); [ "$diff" -lt 0 ] && diff=$((0 - diff))
if [ "$ns_size" = "$d_size" ] && [ "$diff" -le 524288 ]; then echo yes >"$GD/shared"; else echo no >"$GD/shared"; fi
printf '%s %s %s\n' "$(date +%s)" "$ns_av" "$d_av" >"$GD/samples"
say "disk guard: baseline build-directory avail_bytes=$((ns_av * 1024)) size_bytes=$((ns_size * 1024)); daemon-storage avail_bytes=$((d_av * 1024)) size_bytes=$((d_size * 1024)); shared=$(cat "$GD/shared"); reserve 4 GiB; interval ${INTERVAL}s; stall ${STALL_S}s"
if [ "$ns_av" -lt "$FLOOR_K" ] || [ "$d_av" -lt "$FLOOR_K" ]; then say "disk guard: FAIL (available space below 4 GiB before the workload); no context started"; exit 20; fi
if [ "$(cat "$GD/shared")" = no ]; then
  if docker run -d --rm --name "$GCTR" --network none --label "platform.ci.guard=$CI_JOB_ID" "$GUARD_IMAGE" sleep 86400 >/dev/null 2>&1; then GCTR_ON=1
  else say "disk guard: FAIL (cannot start the daemon-storage sampler); no context started"; exit 20; fi
fi

if [ -n "$SIG" ]; then say "signal $SIG before the workload; no context started"; exit 143; fi
date +%s >"$GD/beat"
guard_loop &
SPID=$!
run_context "$T" bash infra/ci/validate.sh "$@" &
CPID=$!
# Supervise: the guard must stay alive and keep sampling while the context runs.
SDONE=0
last="$(date +%s)"
while kill -0 "$CPID" 2>/dev/null; do
  if [ "$SDONE" = 0 ]; then
    if kill -0 "$SPID" 2>/dev/null; then
      # An empty or partial read (the heartbeat file is mid-rewrite) is not a stall: keep the last numeric beat.
      cur="$(cat "$GD/beat" 2>/dev/null)"
      case "$cur" in ''|*[!0-9]*) ;; *) last="$cur" ;; esac
      if [ $(( $(date +%s) - last )) -gt "$STALL_S" ]; then
        kill -9 "$SPID" 2>/dev/null || true; wait "$SPID" 2>/dev/null || true; SDONE=1
        guard_fail "guard stalled"
      fi
    else
      SDONE=1; GRC=0; wait "$SPID" || GRC=$?
      if [ "$GRC" = 3 ]; then guard_note; else guard_fail "guard died"; fi
    fi
  fi
  sleep 2
done
CRC=0
wait "$CPID" || CRC=$?
: >"$GD/stop" 2>/dev/null || true
if [ "$SDONE" = 0 ]; then
  if kill -0 "$SPID" 2>/dev/null; then kill "$SPID" 2>/dev/null || true; wait "$SPID" 2>/dev/null || true
  else
    GRC=0; wait "$SPID" || GRC=$?
    case "$GRC" in 0) ;; 3) guard_note ;; *) guard_fail "guard died" ;; esac
  fi
fi
sample >/dev/null 2>&1 || true
stop_sampler
report_capacity
rm -rf "$T" 2>/dev/null || true
if [ -z "$GFAIL" ] && [ -s "$GD/fail" ]; then guard_note; fi
if [ -n "$GFAIL" ]; then
  say "final: context=$CRC guard=FAIL ($GFAIL)"
  if [ "$CRC" -ne 0 ]; then exit "$CRC"; fi
  exit 20
fi
if [ -n "$SIG" ] && [ "$CRC" -eq 0 ]; then say "final: context=0 after signal $SIG"; exit 143; fi
say "final: context=$CRC guard=ok"
exit "$CRC"
