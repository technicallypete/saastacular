#!/usr/bin/env bash
# Runner preflight (docs/foundation/OPERATIONS.md): every runner requirement the harness relies on, as an
# executable check. It writes one row per check, "ID<TAB>STATUS<TAB>detail", with STATUS PASS, FAIL or
# INFO; the detail holds versions, counts and fixed text only. RP1-RP13 each appear exactly once.
# Preflight containers are bound to the run's owned-resource registry at creation (a cidfile in
# REGISTRY/cid written by the Docker CLI, then a standalone row) and removed before exit; the preflight
# label is never used as removal authority.
#
#   bash infra/ci/runner-preflight.sh --registry DIR --out FILE [--profile fast|affected|integration|full] [--job fast|gate|full|local]
#
# --profile selects the capacity and budget thresholds (infra/ci/lib.sh; default full; G051 affected has its
# own constants there, integration's capacity with the fast job's cap); --job selects the
# CI_JOB_TIMEOUT floor in GitLab CI (fast: the fast job; gate and full: the gate job; default full, or local
# outside CI).
#
# Exit: 0 when no row is FAIL; 1 when any row is FAIL; 2 on a usage or registry error (no rows written).
set -euo pipefail

REG="" OUT="" PROFILE=full JOB=""
pf_usage() { echo "usage: runner-preflight.sh --registry DIR --out FILE [--profile fast|affected|integration|full] [--job fast|gate|full|local]" >&2; exit 2; }
while (( $# > 0 )); do
  case "$1" in
    --registry) REG="${2:-}"; shift 2 || pf_usage ;;
    --out) OUT="${2:-}"; shift 2 || pf_usage ;;
    --profile) PROFILE="${2:-}"; shift 2 || pf_usage ;;
    --job) JOB="${2:-}"; shift 2 || pf_usage ;;
    *) pf_usage ;;
  esac
done
if [[ -z "$REG" || -z "$OUT" || ! "$PROFILE" =~ ^(fast|affected|integration|full)$ ]]; then pf_usage; fi
if [[ -z "$JOB" ]]; then if [[ "${GITLAB_CI:-}" == true ]]; then JOB=full; else JOB=local; fi; fi
if [[ ! "$JOB" =~ ^(fast|gate|full|local)$ ]]; then pf_usage; fi
if [[ -L "$REG" || ! -d "$REG" || -L "$REG/owned.tsv" || ! -f "$REG/owned.tsv" || ! -d "$REG/cid" || -L "$REG/cid" || ! -f "$REG/token" ]]; then
  echo "runner-preflight: --registry is not an owned-resource registry" >&2
  exit 2
fi
: >"$OUT" || exit 2

row() { # ID STATUS DETAIL
  local d="${3//$'\n'/ }"
  d="${d//$'\t'/ }"
  d="$(printf '%s' "$d" | tr -c 'A-Za-z0-9 ._:/=,+@()<>%-' '?')"
  printf '%s\t%s\t%s\n' "$1" "$2" "${d:0:200}" >>"$OUT"
  if [[ "$2" == FAIL ]]; then FAILED=1; fi
}
FAILED=0

# RP1 must not depend on Bash >= 4.4 features.
if (( BASH_VERSINFO[0] < 4 || (BASH_VERSINFO[0] == 4 && BASH_VERSINFO[1] < 4) )); then
  row RP1 FAIL "bash ${BASH_VERSION} (>= 4.4 required)"
  exit 1
fi
shopt -s inherit_errexit
row RP1 PASS "bash ${BASH_VERSION%%(*}"

SELF_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)"
REPO="$(dirname -- "$(dirname -- "$SELF_DIR")")"
# shellcheck source=lib.sh
source "$SELF_DIR/lib.sh"
REPO_R="$(realpath -e -- "$REPO")"
TOKEN="$(tr -d ' \n' <"$REG/token")"
if [[ ! "$TOKEN" =~ ^[0-9a-f]{32}$ ]]; then echo "runner-preflight: malformed registry token" >&2; exit 2; fi
LABEL="platform.validation.preflight=$TOKEN"
CID_SEQ=0

registry_row() { local IFS=$'\t'; printf '%s\n' "$*" >>"$REG/owned.tsv"; }
next_cidfile() { CID_SEQ=$((CID_SEQ + 1)); CIDFILE="$REG/cid/preflight-$$-$CID_SEQ.cid"; }
bind_cidfile() { # PURPOSE: the standalone row for the container the Docker CLI recorded in CIDFILE
  local id
  if [[ -s "$CIDFILE" ]]; then
    id="$(tr -d ' \n' <"$CIDFILE")"
    if [[ "$id" =~ ^[0-9a-f]{64}$ ]]; then registry_row standalone "$id" "$1"; printf '%s' "$id"; return 0; fi
  fi
  return 1
}
remove_container() { # ID: remove and verify
  docker rm -f -v "$1" >/dev/null 2>&1 || true
  if docker inspect --type container "$1" >/dev/null 2>&1; then return 1; fi
  return 0
}
version_ge() { # A.B C.D
  local a1="${1%%.*}" a2="${1#*.}" b1="${2%%.*}" b2="${2#*.}"
  a2="${a2%%.*}"; b2="${b2%%.*}"
  if [[ ! "$a1$a2" =~ ^[0-9]+$ ]]; then return 1; fi
  (( 10#$a1 > 10#$b1 || (10#$a1 == 10#$b1 && 10#$a2 >= 10#$b2) ))
}

registry_row label "$LABEL"

# RP2: host tools.
missing=()
for t in sha256sum realpath stat mktemp sort find grep sed awk tar timeout git setsid ps od head tr cut wc date id df nproc; do
  if command -v "$t" >/dev/null 2>&1; then :; else missing+=("$t"); fi
done
gnu=1
for t in find sed tar timeout stat; do
  out="$("$t" --version 2>/dev/null)" || out=""
  if [[ "${out%%$'\n'*}" == *GNU* ]]; then :; else gnu=0; fi
done
if printf 'ab\n' | grep -qP '^a(?=b)' 2>/dev/null; then :; else gnu=0; fi
if ps -o pgid= -p "$$" >/dev/null 2>&1; then :; else missing+=("ps -o pgid"); fi
out="$(setsid --version 2>/dev/null)" || out=""
if [[ "$out" == *util-linux* ]]; then :; else missing+=("setsid (util-linux)"); fi
if (( ${#missing[@]} == 0 && gnu == 1 )); then row RP2 PASS "GNU coreutils, findutils, grep -P, sed, awk, tar, timeout, git, setsid, ps present"
else row RP2 FAIL "missing or not GNU: ${missing[*]:-} gnu=$gnu"; fi

# RP3: Docker daemon.
DOCKER_OK=0
dv="$(docker version --format '{{.Server.Version}} {{.Server.Os}}/{{.Server.Arch}}' 2>/dev/null)" || dv=""
di="$(docker info --format 'driver={{.Driver}} cgroup={{.CgroupDriver}}/{{.CgroupVersion}}' 2>/dev/null)" || di=""
if [[ -n "$dv" ]] && version_ge "${dv%% *}" 24.0 && [[ "${dv#* }" == linux/amd64 ]]; then row RP3 PASS "docker ${dv} ${di}"; DOCKER_OK=1
else row RP3 FAIL "docker '${dv:-unreachable}' (>= 24.0 on linux/amd64 required)"; fi

# RP4: Compose v2.
cv="$(docker compose version --short 2>/dev/null)" || cv=""
cv="${cv#v}"
if [[ -n "$cv" ]] && version_ge "$cv" 2.24; then row RP4 PASS "compose $cv"; else row RP4 FAIL "compose '${cv:-missing}' (>= 2.24 required)"; fi

# RP5: BuildKit through buildx.
if bv="$(docker buildx version 2>/dev/null)"; then row RP5 PASS "$(printf '%s' "$bv" | awk '{print $1, $2}')"; else row RP5 FAIL "docker buildx version failed"; fi

# RP6: loopback parity. A published 127.0.0.1 port of a detached container is reachable from this shell.
if (( DOCKER_OK )); then
  next_cidfile
  rp6="FAIL" detail="httpd container did not start"
  if docker run -d --cidfile "$CIDFILE" --label "$LABEL" -p 127.0.0.1::80 "$CI_BUSYBOX_IMAGE" httpd -f -p 80 >/dev/null 2>&1; then :; fi
  if id="$(bind_cidfile preflight-httpd)"; then
    port=""
    out="$(docker port "$id" 80/tcp 2>/dev/null)" || out=""
    for a in $out; do if [[ "$a" =~ ^127\.0\.0\.1:([0-9]{1,5})$ ]]; then port="${BASH_REMATCH[1]}"; break; fi; done
    detail="no 127.0.0.1 port mapping"
    if [[ "$port" =~ ^[0-9]{1,5}$ ]]; then
      detail="127.0.0.1:$port not reachable from the job shell"
      for _ in 1 2 3 4 5 6 7 8 9 10; do
        resp="$(timeout 3 bash -c 'exec 3<>"/dev/tcp/127.0.0.1/$1" || exit 9; printf "GET / HTTP/1.0\r\n\r\n" >&3; head -c 9 <&3' rp6 "$port" 2>/dev/null)" || resp=""
        if [[ "$resp" == HTTP/1.* ]]; then rp6=PASS detail="published 127.0.0.1 port reachable from the job shell"; break; fi
        sleep 1
      done
    fi
    if remove_container "$id"; then :; else rp6=FAIL detail="$detail; preflight container not removed"; fi
  fi
  row RP6 "$rp6" "$detail"
else
  row RP6 FAIL "Docker unavailable"
fi

# RP7: bind-mount path parity for a job temporary directory and a tracked file of the checkout.
if (( DOCKER_OK )); then
  tmpd="$(mktemp -d "${TMPDIR:-/tmp}/platform-preflight-XXXXXX")"
  ci_random_hex 16 >"$tmpd/marker"
  tracked="$REPO_R/infra/ci/runner-preflight.sh"
  want="$(sha256sum "$tmpd/marker" "$tracked" | awk '{print $1}' | tr '\n' ' ')"
  next_cidfile
  got="$(docker run --rm --cidfile "$CIDFILE" --label "$LABEL" --network none -v "$tmpd:$tmpd:ro" -v "$REPO_R:$REPO_R:ro" \
    "$CI_BUSYBOX_IMAGE" sha256sum "$tmpd/marker" "$tracked" 2>/dev/null | awk '{print $1}' | tr '\n' ' ')" || got=""
  bind_cidfile preflight-bind >/dev/null || true
  rm -rf -- "$tmpd"
  if [[ -n "$got" && "$got" == "$want" ]]; then row RP7 PASS "temporary directory and checkout readable at identical paths"
  else row RP7 FAIL "bind mounts do not show the same files at the same paths"; fi
else
  row RP7 FAIL "Docker unavailable"
fi

# RP8: a non-loopback IPv4 address (the harness's loopback-only proof needs one).
ip4="" out=""
if command -v hostname >/dev/null 2>&1; then out="$(hostname -I 2>/dev/null)" || out=""; fi
if [[ -z "$out" ]] && command -v ip >/dev/null 2>&1; then
  out="$(ip -4 -o addr show scope global 2>/dev/null | awk '{sub(/\/.*/, "", $4); print $4}')" || out=""
fi
for a in $out; do
  if [[ "$a" =~ ^[0-9]+(\.[0-9]+){3}$ && "$a" != 127.* ]]; then ip4="$a"; break; fi
done
if [[ -n "$ip4" ]]; then row RP8 PASS "non-loopback IPv4 present"; else row RP8 FAIL "no non-loopback IPv4 address"; fi

# RP9: the default and variant host port sets are free on 127.0.0.1.
ports=()
for k in WEB_PORT CORE_PORT INTELLIGENCE_PORT POSTGRES_PORT REDIS_PORT MINIO_API_PORT MINIO_CONSOLE_PORT; do
  v="$(sed -n "s/^$k=\([0-9]\{1,5\}\)$/\1/p" "$REPO_R/.env.example" | tail -1)" || v=""
  if [[ -n "$v" ]]; then ports+=("$v"); fi
done
while IFS= read -r v; do
  for p in $v; do if [[ "$p" =~ ^[0-9]{1,5}$ ]]; then ports+=("$p"); fi; done
done < <(sed -n 's/^PORTS_[A-Z]*=(\(.*\))$/\1/p' "$REPO_R/infra/validation/run.sh")
busy=()
for p in "${ports[@]}"; do
  if timeout 2 bash -c 'exec 3<>"/dev/tcp/127.0.0.1/$1"' rp9 "$p" 2>/dev/null; then busy+=("$p"); fi
done
if (( ${#ports[@]} < 14 )); then row RP9 FAIL "could not read the port sets (${#ports[@]} ports)"
elif (( ${#busy[@]} == 0 )); then row RP9 PASS "${#ports[@]} default and variant ports free on 127.0.0.1"
else row RP9 FAIL "busy on 127.0.0.1: ${busy[*]}"; fi

# RP10: no existing resources of the root Compose project.
proj="$(sed -n 's/^COMPOSE_PROJECT_NAME=//p' "$REPO_R/.env.example" | tail -1)" || proj=""
if [[ ! "$proj" =~ ^[a-z0-9][a-z0-9_-]{0,62}$ ]]; then row RP10 FAIL "root Compose project name not found in .env.example"
elif (( DOCKER_OK == 0 )); then row RP10 FAIL "Docker unavailable"
else
  nc="$(docker ps -aq --filter "label=com.docker.compose.project=$proj" | wc -l)" || nc="?"
  nn="$(docker network ls -q --filter "label=com.docker.compose.project=$proj" | wc -l)" || nn="?"
  nv="$(docker volume ls -q --filter "label=com.docker.compose.project=$proj" | wc -l)" || nv="?"
  if [[ "$nc$nn$nv" == 000 ]]; then row RP10 PASS "no containers, networks or volumes of the root Compose project"
  else row RP10 FAIL "root Compose project resources exist: containers=$nc networks=$nn volumes=$nv"; fi
fi

# RP11: capacity for the profile (infra/ci/lib.sh ci_capacity_ok). Daemon storage is measured inside the
# daemon by an owned, cidfile-bound container (DockerRootDir is not readable from a nested context).
cpus="$(nproc 2>/dev/null)" || cpus=0
memk="$(awk '/^MemTotal:/ {print $2}' /proc/meminfo 2>/dev/null)" || memk=0
tmpk="$(df -Pk "${TMPDIR:-/tmp}" 2>/dev/null | awk 'NR==2 {print $4}')" || tmpk=0
dockk=""
if (( DOCKER_OK )); then
  next_cidfile
  out="$(docker run --rm --cidfile "$CIDFILE" --label "$LABEL" --network none "$CI_BUSYBOX_IMAGE" df -Pk / 2>/dev/null)" || out=""
  bind_cidfile preflight-df >/dev/null || true
  dockk="$(awk 'NR==2 {print $4}' <<<"$out")" || dockk=""
fi
if [[ ! "$cpus" =~ ^[0-9]+$ ]]; then cpus=0; fi
if [[ ! "$memk" =~ ^[0-9]+$ ]]; then memk=0; fi
if [[ ! "$tmpk" =~ ^[0-9]+$ ]]; then tmpk=0; fi
if [[ ! "$dockk" =~ ^[0-9]+$ ]]; then dockk=0; fi
floor="$(ci_disk_floor_gib "$PROFILE")" || floor="?"
need="profile $PROFILE needs >= $(v="CI_CPUS_$PROFILE"; printf '%s' "${!v}") cpus, >= $(v="CI_MEMGIB_$PROFILE"; printf '%s' "${!v}") GiB, >= $floor GiB free"
if ci_hosted_context; then need+=" (hosted context)"; fi
if ci_capacity_ok "$PROFILE" "$cpus" "$memk" "$tmpk" "$dockk"; then rp11=PASS; else rp11=FAIL; fi
row RP11 "$rp11" "cpus=$cpus mem=$((memk / 1048576)) GiB tmp free $((tmpk / 1048576)) GiB, daemon storage free $((dockk / 1048576)) GiB ($need)"

# RP12: time budget for the profile (CI: the job's CI_JOB_TIMEOUT floor and the remaining budget measured
# from CI_JOB_STARTED_AT; local: the local budget).
budget="$(ci_profile_budget_seconds "$PROFILE")" || budget=""
minv="CI_MINBUDGET_$PROFILE"
jf="$JOB"
if [[ "$jf" == full ]]; then jf=gate; fi
if [[ "${GITLAB_CI:-}" == true ]]; then
  if [[ "$jf" == local ]] || ci_time_ok "$jf" "${CI_JOB_TIMEOUT:-}"; then
    if [[ ! "$budget" =~ ^-?[0-9]+$ ]] || (( budget < ${!minv} )); then
      row RP12 FAIL "harness budget '${budget:-malformed}' s (>= ${!minv} s required for $PROFILE)"
    else
      row RP12 PASS "CI_JOB_TIMEOUT=${CI_JOB_TIMEOUT} s, harness budget ${budget} s ($PROFILE)"
    fi
  else
    jfv="CI_JOBFLOOR_$jf"
    row RP12 FAIL "CI_JOB_TIMEOUT '${CI_JOB_TIMEOUT:-unset}' (>= ${!jfv} s required for the $jf job)"
  fi
elif [[ ! "$budget" =~ ^-?[0-9]+$ ]] || (( budget < ${!minv} )); then
  row RP12 FAIL "harness budget '${budget:-malformed}' s (>= ${!minv} s required for $PROFILE)"
else
  row RP12 PASS "local run, harness budget ${budget} s ($PROFILE)"
fi

# RP13: information only.
running="$(docker ps -q 2>/dev/null | wc -l)" || running="?"
up="$(cut -d. -f1 /proc/uptime 2>/dev/null)" || up="?"
boot="$(sha256sum /proc/sys/kernel/random/boot_id 2>/dev/null | cut -c1-16)" || boot="?"
row RP13 INFO "runner=${CI_RUNNER_ID:-none} description=${CI_RUNNER_DESCRIPTION:-none} tags=${CI_RUNNER_TAGS:-none} gitlab=${CI_SERVER_VERSION:-none} running_containers=$running uptime_s=$up boot_id=$boot"

if (( FAILED )); then exit 1; fi
exit 0
