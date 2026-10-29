#!/bin/sh
# Staging CI launcher (G072 Plan r1 §10.10, r0 §2.2/§2.5; docs/foundation/STAGING.md). POSIX sh for the job's docker:cli
# shell, used only by the staging:* jobs of .gitlab-ci.yml beside their job-private TLS Docker daemon (service alias
# "docker"), after the pattern of infra/ci/hosted-context.sh. That file is not edited: its pinned project ID and path
# SHA-256 are read from it here as the single source.
#
#   sh infra/ci/staging-action.sh ACTION --environment ENV [--candidate C] [--duration D]   (job script)
#   sh infra/ci/staging-action.sh --after-script ACTION --environment ENV                   (job after_script; exits 0)
#
# ACTION is prepare, check, build, deploy or stop, and the job must be staging:ACTION. In order, each step before the next:
#  1. a GitLab CI job with the job-private TLS daemon (G1), else 16 with no Docker call;
#  2. the approved context, else 16, before any token is touched: the pinned project, branch staging, a protected ref, a
#     web or API pipeline without merge request or tag, the exact job name; check/deploy/stop: CI_ENVIRONMENT_NAME equals
#     the selected environment, an ID token and the environment's profile files are present; prepare/build: no
#     environment, no ID token and no profile variable (a scope misconfiguration fails closed);
#  3. the inputs, else 2, before the token file exists: the environment is exactly one line of
#     infra/staging/ci-environments.txt; candidate and duration match the pipeline-input regexes; stop takes neither;
#     build/deploy need STAGING_REUSE (check's dotenv) yes or no. Only the arguments select: no inherited STAGING_*
#     variable is read as a fallback or forwarded;
#  4. the private directory "$NS/.staging-ci-$CI_JOB_ID" (new, mode 700, on the build directory's filesystem, which the
#     job daemon shares, and outside the project directory, so never an artifact). check/deploy/stop move the ID token
#     into id-token (0600) and copy the profile files into profile/ (check: staging.conf and ci-settings only); then
#     every STAGING_* variable is unset (STAGING_LEASE_HELD included);
#  5. no job credential in any file below "$NS" and no ID token below the project directory, else 16. prepare first
#     fetches refs/heads/* through the runner's own credential helper (merge-request refs are never fetched);
#  6. the context image (ci-context.Dockerfile for prepare/build, staging-ci.Dockerfile for check/deploy/stop) runs
#     `bash infra/staging/ci.sh ACTION` with the build directory at its own path, the job's file-variable directory
#     hidden, host network to the job daemon and an explicit -e allow-list (the STAGING_CI_* contract values and
#     public CI metadata);
#  7. on exit, also after TERM/INT/HUP: the context is stopped, the ID token is searched for below the project
#     directory (a hit withholds the job's outputs, 6), deploy/stop get the fixed-field summary if the adapter wrote
#     none, and the private directory is removed with its absence verified.
# Contract with the adapter: STAGING_CI_ACTION, STAGING_CI_ENV, STAGING_CI_CANDIDATE_INPUT, STAGING_CI_DURATION_INPUT,
# STAGING_CI_PRIVATE, STAGING_CI_OUT (staging-prepare/, staging-check/ with the dotenv reuse.env, staging-build/,
# staging-summary/ with summary.txt), STAGING_CI_ID_TOKEN_FILE and STAGING_CI_PROFILE_DIR (check/deploy/stop),
# STAGING_REUSE (build/deploy). refs/remotes/origin/* is fetched here for prepare; the adapter does not fetch.
#
# Exit: the adapter's class (Plan S14: 0 success, 2 refused before any change, 3 HOLD, 4 failed after a change, 5 session
# expired, 6 privacy canary); 2 refused input; 16 not an approved CI context (usage, gate, unsafe path, profile not
# delivered, credential not proven absent, failed context build); 4 when an otherwise successful run cannot prove the
# private directory removed. No secret value is printed, passed as an argument or written below the project directory.
set -u
set +x
umask 077

PFX="staging-action"
say() { printf '%s: %s\n' "$PFX" "$*"; }
STOP_GRACE=60            # context stop grace, seconds: the adapter's own trap releases its lease within it

MODE=main
case "${1:-}" in --after-script) MODE=after; shift ;; esac
ACTION="${1:-}"
[ $# -gt 0 ] && shift
ENV_SEL="" CAND="" DUR="" S_ENV=0 S_CAND=0 S_DUR=0 BAD=0
while [ $# -gt 0 ]; do
  case "$1" in
    --environment) if [ $# -lt 2 ] || [ "$S_ENV" = 1 ]; then BAD=1; break; fi; ENV_SEL="$2"; S_ENV=1; shift 2 ;;
    --candidate) if [ $# -lt 2 ] || [ "$S_CAND" = 1 ] || [ "$MODE" = after ]; then BAD=1; break; fi; CAND="$2"; S_CAND=1; shift 2 ;;
    --duration) if [ $# -lt 2 ] || [ "$S_DUR" = 1 ] || [ "$MODE" = after ]; then BAD=1; break; fi; DUR="$2"; S_DUR=1; shift 2 ;;
    *) BAD=1; break ;;
  esac
done
case "$ACTION" in prepare|check|build|deploy|stop) ;; *) BAD=1 ;; esac
[ "$S_ENV" = 1 ] || BAD=1
if [ "$BAD" = 1 ]; then
  say "usage: staging-action.sh ACTION --environment ENV [--candidate C] [--duration D] | --after-script ACTION --environment ENV (ACTION: prepare, check, build, deploy, stop)"
  if [ "$MODE" = after ]; then exit 0; fi
  exit 16
fi
case "$ACTION" in check|deploy|stop) CRED=1 ;; *) CRED=0 ;; esac
case "$ACTION" in
  prepare) OUTN=staging-prepare; CTX_FILE=infra/docker/ci-context.Dockerfile ;;
  check) OUTN=staging-check; CTX_FILE=infra/docker/staging-ci.Dockerfile ;;
  build) OUTN=staging-build; CTX_FILE=infra/docker/ci-context.Dockerfile ;;
  *) OUTN=staging-summary; CTX_FILE=infra/docker/staging-ci.Dockerfile ;;
esac

# --- G1: only in a GitLab CI job with the job-private TLS daemon (no Docker call otherwise) -------------------------
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
case "$NS" in *[!A-Za-z0-9._/-]*) say "refused: unexpected characters in the build directory path"; [ "$MODE" = after ] && exit 0; exit 16 ;; esac
P="$NS/.staging-ci-$CI_JOB_ID"
OUT="$CI_PROJECT_DIR/$OUTN"
CTX_NAME="platform-staging-ci-$CI_JOB_ID"
CTX_IMAGE="platform-staging-ci:$CI_JOB_ID"
PF=""
NL="$(printf '\nx')"; NL="${NL%x}"

wait_daemon() { # TRIES
  i=0
  until docker version >/dev/null 2>&1; do
    i=$((i + 1)); [ "$i" -ge "$1" ] && return 1
    sleep 2
  done
  return 0
}
stop_context() { if docker container inspect "$CTX_NAME" >/dev/null 2>&1; then docker stop -t "$STOP_GRACE" "$CTX_NAME" >/dev/null 2>&1 || true; fi; }
clean_word() { # VALUE: printable ASCII without spaces, else "-"
  case "$1" in ''|*[!A-Za-z0-9._:/@+-]*) printf '%s\n' "-" ;; *) printf '%s\n' "$1" ;; esac
}
# The fixed-field summary (deploy/stop) when the adapter wrote none; the key order is ci-publish.sh's SUMMARY_KEYS.
fallback_summary() { # EXIT_CLASS
  case "$ACTION" in deploy|stop) ;; *) return 0 ;; esac
  [ -L "$OUT" ] && return 0
  if [ -f "$OUT/summary.txt" ]; then return 0; fi
  keys="$(sed -n 's/^SUMMARY_KEYS="\([a-z0-9_ ]*\)"$/\1/p' "$CI_PROJECT_DIR/infra/staging/ci-publish.sh" 2>/dev/null)"
  [ -n "$keys" ] || keys="action environment pipeline job actor exit_class"
  mkdir -p "$OUT" 2>/dev/null || return 0
  for k in $keys; do
    case "$k" in
      action) v="$ACTION" ;;
      environment) v="$(clean_word "$ENV_SEL")" ;;
      pipeline) v="$(clean_word "${CI_PIPELINE_ID:-}")" ;;
      job) v="$(clean_word "${CI_JOB_ID:-}")" ;;
      actor) v="$(clean_word "${GITLAB_USER_LOGIN:-}")" ;;
      exit_class) v="$1" ;;
      *) v="-" ;;
    esac
    printf '%s: %s\n' "$k" "$v"
  done >"$OUT/summary.txt" 2>/dev/null || true
}
show_summary() {
  case "$ACTION" in deploy|stop) ;; *) return 0 ;; esac
  if [ -f "$OUT/summary.txt" ] && [ ! -L "$OUT/summary.txt" ]; then
    say "summary ($OUTN/summary.txt):"
    grep -E '^[a-z0-9_]+: [ -~]*$' "$OUT/summary.txt" | head -n 40
  fi
}
remove_private() { # 0 when the private directory is proven absent
  if [ -L "$P" ]; then rm -f "$P" 2>/dev/null; elif [ -e "$P" ]; then rm -rf -- "$P" 2>/dev/null; fi
  if [ -e "$P" ] || [ -L "$P" ]; then say "cleanup: the private directory could not be removed"; return 1; fi
  return 0
}

# --- after_script: stop a remaining context, the summary backstop, remove the private directory (always 0) ---------
if [ "$MODE" = after ]; then
  if wait_daemon 30; then stop_context; else say "after-script: the job daemon is unreachable"; fi
  fallback_summary "unknown (the job ended before the adapter wrote its summary)"
  if [ -e "$P" ] || [ -L "$P" ]; then
    if remove_private; then say "after-script: removed the private directory (verified)"; fi
  fi
  exit 0
fi

# --- the approved context (before any token is touched) ---------------------------------------------------------------
HC="$CI_PROJECT_DIR/infra/ci/hosted-context.sh"
PIN_ID="$(sed -n 's/^STG_PROJECT_ID=\([0-9][0-9]*\)$/\1/p' "$HC" 2>/dev/null)"
PIN_PATH_SHA="$(sed -n 's/^STG_PROJECT_PATH_SHA=\([0-9a-f]\{64\}\)$/\1/p' "$HC" 2>/dev/null)"
WHY=""
has_token() { case "${STAGING_CI_ID_TOKEN:-}" in ????????*) return 0 ;; *) return 1 ;; esac; }
profile_file() { # VALUE: a non-empty regular file outside the project directory
  case "$1" in ''|"$CI_PROJECT_DIR"/*) return 1 ;; esac
  [ -f "$1" ] && [ ! -L "$1" ] && [ -s "$1" ] && [ -r "$1" ]
}
approved() {
  if [ -z "$PIN_ID" ] || [ -z "$PIN_PATH_SHA" ] || [ "$PIN_ID" != "${PIN_ID%"$NL"*}" ] || [ "$PIN_PATH_SHA" != "${PIN_PATH_SHA%"$NL"*}" ]; then
    WHY="the pinned project values are not readable from infra/ci/hosted-context.sh"; return 1
  fi
  [ "${CI_PROJECT_ID:-}" = "$PIN_ID" ] || { WHY="project"; return 1; }
  [ "$(printf '%s' "${CI_PROJECT_PATH:-}" | sha256sum | cut -d' ' -f1)" = "$PIN_PATH_SHA" ] || { WHY="project path"; return 1; }
  [ "${CI_COMMIT_BRANCH:-}" = staging ] && [ "${CI_COMMIT_REF_NAME:-}" = staging ] || { WHY="branch"; return 1; }
  [ "${CI_COMMIT_REF_PROTECTED:-}" = true ] || { WHY="protected ref"; return 1; }
  case "${CI_PIPELINE_SOURCE:-}" in web|api) ;; *) WHY="pipeline source"; return 1 ;; esac
  [ -z "${CI_MERGE_REQUEST_IID:-}" ] || { WHY="merge request"; return 1; }
  [ -z "${CI_COMMIT_TAG:-}" ] || { WHY="tag"; return 1; }
  [ "${CI_JOB_NAME:-}" = "staging:$ACTION" ] || { WHY="job name"; return 1; }
  if [ "$CRED" = 1 ]; then
    [ -n "$ENV_SEL" ] && [ "${CI_ENVIRONMENT_NAME:-}" = "$ENV_SEL" ] || { WHY="environment"; return 1; }
    has_token || { WHY="no ID token"; return 1; }
    profile_file "${STAGING_CI_CONF:-}" && profile_file "${STAGING_CI_SETTINGS:-}" || { WHY="profile not delivered"; return 1; }
    if [ "$ACTION" != check ]; then
      profile_file "${STAGING_CI_BACKEND:-}" && profile_file "${STAGING_CI_TFVARS:-}" || { WHY="profile not delivered"; return 1; }
    fi
  else
    [ -z "${CI_ENVIRONMENT_NAME:-}" ] || { WHY="environment in a credential-free job"; return 1; }
    [ -z "${STAGING_CI_ID_TOKEN+x}" ] || { WHY="ID token in a credential-free job"; return 1; }
    for v in STAGING_CI_CONF STAGING_CI_BACKEND STAGING_CI_TFVARS STAGING_CI_SETTINGS; do
      eval "[ -z \"\${$v+x}\" ]" || { WHY="profile variable in a credential-free job"; return 1; }
    done
  fi
  return 0
}
if approved; then :; else say "refused: not the approved staging CI context ($WHY); no token touched"; exit 16; fi

# --- inputs (before the token file exists) ------------------------------------------------------------------------------
REG="$CI_PROJECT_DIR/infra/staging/ci-environments.txt"
refuse() { say "refused: $*"; exit 2; }
if [ ! -f "$REG" ] || [ -L "$REG" ] || [ ! -s "$REG" ] || grep -qvxE '[a-z][a-z0-9-]{0,30}' "$REG"; then refuse "the environment registry is missing or malformed"; fi
case "$ENV_SEL$CAND$DUR" in *"$NL"*) refuse "an input holds a line break" ;; esac
grep -qxF -- "$ENV_SEL" "$REG" || refuse "environment '$(clean_word "$ENV_SEL")' is not in infra/staging/ci-environments.txt"
printf '%s\n' "$CAND" | grep -qxE '([0-9a-f]{40}|[A-Za-z0-9][A-Za-z0-9._/-]{0,199})?' || refuse "malformed candidate"
printf '%s\n' "$DUR" | grep -qxE '([1-9][0-9]{0,5}[mh]?)?' || refuse "malformed duration"
if [ "$ACTION" = stop ] && { [ -n "$CAND" ] || [ -n "$DUR" ]; }; then refuse "stop takes no candidate and no duration"; fi
REUSE=""
case "$ACTION" in
  build|deploy) case "${STAGING_REUSE:-}" in yes|no) REUSE="$STAGING_REUSE" ;; *) refuse "STAGING_REUSE from staging:check is not yes or no" ;; esac ;;
esac
PROBES=""
if [ "$CRED" = 1 ] && [ -n "${CI_DENIAL_PROBES:-}" ]; then
  case "$CI_DENIAL_PROBES" in *[!A-Za-z0-9\ ._:/,-]*) refuse "CI_DENIAL_PROBES holds an unexpected character" ;; esac
  PROBES="$CI_DENIAL_PROBES"
fi

# --- the private directory, the token file and the profile ---------------------------------------------------------------
CLEANED=0
finish() { # the backstop of every exit path after this point (the normal path cleans explicitly)
  if [ "$CLEANED" = 1 ]; then return 0; fi
  stop_context
  if [ -n "$PF" ]; then rm -f "$PF" "$PF.hits" "$PF.err"; fi
  remove_private >/dev/null 2>&1 || say "cleanup: the private directory could not be removed"
}
on_signal() { SIG="$1"; say "signal $1 received; stopping the context"; stop_context; }
SIG=""
trap 'finish' EXIT
trap 'on_signal TERM' TERM
trap 'on_signal INT' INT
trap 'on_signal HUP' HUP

if [ -e "$P" ] || [ -L "$P" ]; then say "refused: the private directory already exists"; exit 16; fi
mkdir -m 700 "$P" 2>/dev/null || { say "refused: cannot create the private directory"; exit 16; }
if [ -L "$P" ] || [ ! -d "$P" ] || [ "$(cd "$P" 2>/dev/null && pwd -P)" != "$(cd "$NS" && pwd -P)/.staging-ci-$CI_JOB_ID" ]; then
  say "refused: the private directory is not a plain directory under the build directory parent"; exit 16
fi
# On the build directory's filesystem (the job daemon shares it, so the context sees the same path), never inside it.
if [ "$(stat -c %d "$P" 2>/dev/null)" != "$(stat -c %d "$CI_PROJECT_DIR" 2>/dev/null)" ]; then
  say "refused: the private directory is not on the build directory's filesystem"; exit 16
fi
mkdir -m 700 "$P/home" "$P/tmp" 2>/dev/null || { say "refused: cannot prepare the private directory"; exit 16; }
if [ "$CRED" = 1 ]; then
  (umask 077 && printf '%s' "$STAGING_CI_ID_TOKEN" >"$P/id-token") 2>/dev/null || { say "refused: cannot write the token file"; exit 16; }
  chmod 600 "$P/id-token" 2>/dev/null
  if [ "$(stat -c %a "$P/id-token" 2>/dev/null)" != 600 ] || [ ! -s "$P/id-token" ]; then say "refused: the token file is not a mode-600 file"; exit 16; fi
  mkdir -m 700 "$P/profile" || { say "refused: cannot prepare the profile directory"; exit 16; }
  cp "$STAGING_CI_CONF" "$P/profile/staging.conf" && cp "$STAGING_CI_SETTINGS" "$P/profile/ci-settings" ||
    { say "refused: cannot copy the profile"; exit 16; }
  if [ "$ACTION" != check ]; then
    cp "$STAGING_CI_BACKEND" "$P/profile/backend.hcl" && cp "$STAGING_CI_TFVARS" "$P/profile/terraform.tfvars" ||
      { say "refused: cannot copy the profile"; exit 16; }
  fi
  chmod 600 "$P/profile/"* 2>/dev/null
fi
for v in $(env | sed -n 's/^\(STAGING_[A-Za-z0-9_]*\)=.*/\1/p'); do unset "$v"; done
unset STAGING_CI_ID_TOKEN STAGING_CI_CONF STAGING_CI_BACKEND STAGING_CI_TFVARS STAGING_CI_SETTINGS STAGING_LEASE_HELD STAGING_REUSE \
  STAGING_CI_ACTION STAGING_CI_ENV STAGING_CI_CANDIDATE_INPUT STAGING_CI_DURATION_INPUT STAGING_CI_IMAGES STAGING_CI_PRIVATE \
  STAGING_CI_OUT STAGING_CI_ID_TOKEN_FILE STAGING_CI_PROFILE_DIR STAGING_CI_CANARY_LITERALS
say "context approved: staging:$ACTION for $ENV_SEL; private directory ready$([ "$CRED" = 1 ] && printf '%s' '; ID token moved to a mode-600 file (value not shown)')"

# --- prepare: same-repository branches through the runner's credential helper (never merge-request refs) -----------
if [ "$ACTION" = prepare ]; then
  git -C "$CI_PROJECT_DIR" fetch --no-tags --prune --quiet origin '+refs/heads/*:refs/remotes/origin/*' ||
    { say "refused: cannot fetch the repository's branches"; exit 16; }
fi

# --- no job credential below $NS; no ID token below the project directory ------------------------------------------
# Match the validation launcher: runner checkout may still persist a remote password even
# with FF_GIT_URLS_WITHOUT_TOKENS. Remove remote URL passwords before the unchanged
# literal-credential scan; unsafe syntax and credentials anywhere else still refuse.
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
sanitize_git_config() {
  cfg="$CI_PROJECT_DIR/.git/config"
  if [ -L "$cfg" ] || [ ! -f "$cfg" ] || [ ! -r "$cfg" ]; then return 1; fi
  new="$CI_PROJECT_DIR/.git/config.staging-action.$$"
  if [ -e "$new" ] || [ -L "$new" ]; then return 1; fi
  if awk -v target="$new" "$SAN_AWK" "$cfg" >/dev/null && [ ! -L "$new" ] && [ -f "$new" ]; then
    cat "$new" >"$cfg" || { rm -f "$new"; return 1; }
    rm -f "$new"
  else
    rm -f "$new"; return 1
  fi
}
PASSWORD_URL_RE='[A-Za-z][A-Za-z0-9+.-]*://[^/@[:space:]"]*:[^/@[:space:]"]*@'
scan_tree() { # DIR PATTERN_FILE: 0 when no regular file below DIR holds a pattern and the scan itself was clean
  : >"$PF.err"
  find "$1" -type f -print0 2>>"$PF.err" | xargs -0 -r grep -l -F -a -f "$2" -- >"$PF.hits" 2>>"$PF.err"
  xrc=$?
  if [ -s "$PF.hits" ]; then return 1; fi
  if [ -s "$PF.err" ] || { [ "$xrc" -ne 0 ] && [ "$xrc" -ne 123 ]; }; then return 1; fi
  return 0
}
prove_clean() {
  cfg="$CI_PROJECT_DIR/.git/config"
  if [ -L "$cfg" ] || [ ! -f "$cfg" ] || [ ! -r "$cfg" ]; then say "credentials: .git/config missing or unreadable"; return 1; fi
  sanitize_git_config || { say "credentials: unsafe Git configuration (fail closed)"; return 1; }
  if grep -qE "$PASSWORD_URL_RE" "$cfg"; then say "credentials: a password-bearing URL is in .git/config (fail closed)"; return 1; fi
  PF="$(mktemp /tmp/staging-action-patterns.XXXXXX)" || return 1
  case "$PF" in "$NS"/*) say "credentials: the pattern file is below the build directory parent"; return 1 ;; esac
  for v in "${CI_JOB_TOKEN:-}" "${CI_REGISTRY_PASSWORD:-}" "${CI_DEPENDENCY_PROXY_PASSWORD:-}"; do
    if [ "${#v}" -ge 8 ]; then printf '%s\n' "$v" >>"$PF"; fi
  done
  jt="${CI_JOB_TOKEN:-}"
  if [ "${#jt}" -lt 8 ]; then say "credentials: no job credential to check against (fail closed)"; return 1; fi
  jt=""
  if scan_tree "$NS" "$PF"; then :; else say "credentials: a job credential is stored below the build directory parent, or the scan failed (fail closed)"; return 1; fi
  if [ "$CRED" = 1 ]; then
    if scan_tree "$CI_PROJECT_DIR" "$P/id-token"; then :; else say "credentials: the ID token is stored below the project directory, or the scan failed (fail closed)"; return 1; fi
  fi
  rm -f "$PF" "$PF.hits" "$PF.err"; PF=""
  say "credentials: password_urls=0 job_credential_files=0$([ "$CRED" = 1 ] && printf '%s' ' id_token_in_project=0') scan=clean"
  return 0
}
if prove_clean; then :; else say "refused: credentials not proven absent; no context started"; exit 16; fi

# --- the context -------------------------------------------------------------------------------------------------------
if [ -L "$OUT" ] || { [ -e "$OUT" ] && [ ! -d "$OUT" ]; }; then say "refused: $OUTN is not a plain directory"; exit 16; fi
mkdir -p "$OUT" || { say "refused: cannot create $OUTN"; exit 16; }
if wait_daemon 90; then :; else say "the job daemon is unreachable"; exit 16; fi
if docker image inspect "$CTX_IMAGE" >/dev/null 2>&1; then :
elif docker build -q --label "platform.ci.staging=$CI_JOB_ID" -t "$CTX_IMAGE" - <"$CI_PROJECT_DIR/$CTX_FILE" >/dev/null 2>&1; then :
else say "cannot build the context image ($CTX_FILE)"; exit 16; fi
if [ -n "$SIG" ]; then say "signal $SIG before the context; nothing started"; exit 143; fi

set -- --rm --init --name "$CTX_NAME" --label "platform.ci.staging=$CI_JOB_ID" --network host --security-opt no-new-privileges \
  -v "$NS:$NS" --mount "type=tmpfs,destination=$CI_PROJECT_DIR.tmp,tmpfs-size=1m" -v /certs/client:/certs/client:ro \
  -w "$CI_PROJECT_DIR" -e "HOME=$P/home" -e "TMPDIR=$P/tmp" \
  -e DOCKER_HOST=tcp://localhost:2376 -e DOCKER_TLS_VERIFY=1 -e DOCKER_CERT_PATH=/certs/client \
  -e "STAGING_CI_ACTION=$ACTION" -e "STAGING_CI_ENV=$ENV_SEL" -e "STAGING_CI_CANDIDATE_INPUT=$CAND" \
  -e "STAGING_CI_DURATION_INPUT=$DUR" -e "STAGING_CI_PRIVATE=$P" -e "STAGING_CI_OUT=$OUT"
if [ "$CRED" = 1 ]; then set -- "$@" -e "STAGING_CI_ID_TOKEN_FILE=$P/id-token" -e "STAGING_CI_PROFILE_DIR=$P/profile"; fi
if [ -n "$REUSE" ]; then set -- "$@" -e "STAGING_REUSE=$REUSE"; fi
if [ -n "$PROBES" ]; then set -- "$@" -e "CI_DENIAL_PROBES=$PROBES"; fi
# Public pipeline metadata only (by name); credentials never travel this way.
set -- "$@" -e GITLAB_CI -e CI_PIPELINE_SOURCE -e CI_COMMIT_SHA -e CI_COMMIT_BRANCH -e CI_COMMIT_REF_NAME \
  -e CI_COMMIT_REF_PROTECTED -e CI_DEFAULT_BRANCH -e CI_PIPELINE_ID -e CI_PIPELINE_IID -e CI_PIPELINE_URL -e CI_JOB_ID \
  -e CI_JOB_URL -e CI_JOB_NAME -e CI_JOB_STARTED_AT -e CI_JOB_TIMEOUT -e CI_ENVIRONMENT_NAME -e GITLAB_USER_LOGIN \
  -e CI_PROJECT_DIR -e CI_PROJECT_ID -e CI_SERVER_URL -e CI_SERVER_HOST -e CI_RUNNER_ID -e CI_RUNNER_DESCRIPTION \
  -e CI_RUNNER_TAGS -e CI_SERVER_VERSION -e RUNNER_SCRIPT_TIMEOUT
docker run "$@" "$CTX_IMAGE" bash infra/staging/ci.sh "$ACTION" &
CPID=$!
RC=0
wait "$CPID" || RC=$?
while kill -0 "$CPID" 2>/dev/null; do RC=0; wait "$CPID" || RC=$?; done
say "adapter: staging:$ACTION exit $RC"

# --- outputs and cleanup -----------------------------------------------------------------------------------------------
stop_context
if [ "$CRED" = 1 ] && [ -s "$P/id-token" ]; then
  PF="$(mktemp /tmp/staging-action-patterns.XXXXXX)" || PF="/tmp/staging-action-patterns.$$"
  if scan_tree "$CI_PROJECT_DIR" "$P/id-token"; then :; else
    say "canary: the ID token (or an unreadable file) is below the project directory; outputs withheld"
    rm -rf -- "$OUT"
    while IFS= read -r h; do rm -f -- "$h"; done <"$PF.hits"
    RC=6
  fi
  rm -f "$PF" "$PF.hits" "$PF.err"; PF=""
fi
if [ "$ACTION" = check ] && [ "$RC" = 0 ]; then
  if [ "$(cat "$OUT/reuse.env" 2>/dev/null)" != STAGING_REUSE=yes ] && [ "$(cat "$OUT/reuse.env" 2>/dev/null)" != STAGING_REUSE=no ]; then
    say "refused: staging:check produced no valid reuse decision (reuse.env must be exactly STAGING_REUSE=yes or no)"
    rm -f "$OUT/reuse.env"; RC=2
  fi
fi
fallback_summary "$RC"
show_summary
CLEANED=1
if remove_private; then say "cleanup: private directory removed (verified)"
else [ "$RC" = 0 ] && RC=4; fi
if [ -n "$SIG" ] && [ "$RC" = 0 ]; then RC=143; fi
exit "$RC"
