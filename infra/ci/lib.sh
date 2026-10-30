#!/usr/bin/env bash
# Shared helpers for infra/ci/validate.sh, infra/ci/cleanup-owned.sh and infra/ci/runner-preflight.sh
# (docs/foundation/OPERATIONS.md). Sourced, never executed. Bash >= 4.4, GNU coreutils, git, Docker.
# Every check returns explicitly (`if ...; then return N; fi` or `|| return N`); no statement is negated
# with `!`, because errexit ignores such statements. No here-document or here-string is used, so no
# temporary file is ever created through TMPDIR by this file.
# shellcheck disable=SC2034  # constants are consumed by the scripts that source this file

CI_GITLEAKS_IMAGE="zricethezav/gitleaks:v8.30.1"
CI_YQ_IMAGE="mikefarah/yq:4.48.1"
CI_BUSYBOX_IMAGE="busybox:1.37.0"
CI_REGISTRY_HEADER="#platform-validation-owned v1"
CI_LOCAL_BUDGET_S=14400
CI_BUDGET_RESERVE_S=1200
CI_MIN_BUDGET_S=3600

# Validation profiles (G003 CI revision Plan r2 §2.1, §3.2, §3.3; Spec §12). Suites are run.sh suites; "all"
# in run.sh equals the full list. Budgets are seconds; capacity is CPUs, GiB of memory and GiB free on the
# temporary directory and on daemon storage.
# The staging profile (G004 CI-1) runs only in the protected identity-staging job: the real WorkOS staging
# contract (run.sh staging) with the credential infra/ci/hosted-context.sh forwards there.
# G051 (Plan r3 §3.1, §3.3 rule 8): the static suites are split (lint, lint-base, lint-web, lint-core) and the
# affected profile has no static list; infra/ci/select-profile.sh computes its suites from the selected areas.
# CI_SUITE_ORDER is the single canonical suite order ("final" always runs and is never listed).
CI_SUITE_ORDER="lint lint-base lint-web lint-core controls ci-controls provenance contracts unit unit-web unit-core unit-intel root migrations ops rename off feature-web feature-core staging"
CI_PROFILES="fast affected integration full staging"
CI_SUITES_fast="lint lint-base lint-web lint-core provenance contracts unit"
CI_SUITES_integration="lint lint-base lint-web lint-core provenance contracts unit root ops rename off"
CI_SUITES_full="lint lint-base lint-web lint-core controls ci-controls provenance contracts root migrations ops rename off"
CI_SUITES_staging="staging"
# G051 (Plan r3 §3.9): affected capacity and budget equal integration's; within the fast job the cap is 1500 s
# and the minimum 600 s.
CI_RESERVE_fast=300 CI_RESERVE_affected=300 CI_RESERVE_integration=1200 CI_RESERVE_full=1200 CI_RESERVE_staging=300
CI_CAP_fast=1500 CI_CAP_affected=1500 CI_CAP_integration=5400 CI_CAP_full=0 CI_CAP_staging=1800
CI_MINBUDGET_fast=600 CI_MINBUDGET_affected=600 CI_MINBUDGET_integration=1800 CI_MINBUDGET_full=3600 CI_MINBUDGET_staging=900
CI_JOBFLOOR_fast=1800 CI_JOBFLOOR_gate=10800 CI_JOBFLOOR_staging=2700
CI_CPUS_fast=2 CI_CPUS_affected=4 CI_CPUS_integration=4 CI_CPUS_full=4 CI_CPUS_staging=2
CI_MEMGIB_fast=4 CI_MEMGIB_affected=8 CI_MEMGIB_integration=8 CI_MEMGIB_full=8 CI_MEMGIB_staging=4
# The staging floor is provisional: re-derived from the first authorized run's measured peak with
# max(15, ceil(1.5 x peak + 4)) (Plan r5 §8.5); any change is reviewed.
CI_DISKGIB_fast=15 CI_DISKGIB_affected=40 CI_DISKGIB_integration=40 CI_DISKGIB_full=40 CI_DISKGIB_staging=15
# Provisional start floor for affected (G051), integration and full on the qualified hosted adapter context
# only (Spec §12; not the fast or staging profiles).
CI_DISKGIB_HOSTED=35
# The hosted adapter context marker (infra/ci/hosted-context.sh sets it; ci_hosted_context verifies it).
CI_HOSTED_MARKER="hosted-context/1"

# ci_profile_known NAME: NAME is a validation profile.
ci_profile_known() { [[ "$1" =~ ^(fast|affected|integration|full|staging)$ ]]; }

# ci_profile_suites PROFILE: the run.sh suites of PROFILE. affected has no static list (G051), so it fails.
ci_profile_suites() {
  if ci_profile_known "$1"; then :; else return 1; fi
  if [[ "$1" == affected ]]; then return 1; fi
  local v="CI_SUITES_$1"
  printf '%s\n' "${!v}"
}

# ci_suites_canonical LIST: the suites of LIST (space-separated, may be empty) once each, in CI_SUITE_ORDER.
# An unknown or malformed name fails (G051).
ci_suites_canonical() {
  local list="$1" s out="" IFS=$' \t\n'
  if [[ -n "$list" && ! "$list" =~ ^[a-z][a-z-]*( [a-z][a-z-]*)*$ ]]; then return 1; fi
  for s in $list; do
    if [[ " $CI_SUITE_ORDER " == *" $s "* ]]; then :; else return 1; fi
  done
  for s in $CI_SUITE_ORDER; do
    if [[ " $list " == *" $s "* ]]; then out+="${out:+ }$s"; fi
  done
  printf '%s\n' "$out"
}

# ci_suites_minus LIST REMOVE: the canonical suites of LIST that are not in REMOVE (G051 remainder partition).
ci_suites_minus() {
  local a b s out="" IFS=$' \t\n'
  a="$(ci_suites_canonical "$1")" || return 1
  b="$(ci_suites_canonical "$2")" || return 1
  for s in $a; do
    if [[ " $b " == *" $s "* ]]; then :; else out+="${out:+ }$s"; fi
  done
  printf '%s\n' "$out"
}

# ci_job_partition JOB PIPELINE_PROFILE PIPELINE_SUITES: the job's share of the pipeline (G051 Plan r3 §3.3
# rule 8, §3.7) as one line "EXECUTED_PROFILE<TAB>EXECUTED_SUITES<TAB>COVERAGE<TAB>PARTITION". The fast job runs
# affected (the whole pipeline) or fast; the gate runs nothing for fast or affected and the remainder
# (PIPELINE_SUITES minus CI_SUITES_fast) for integration or full; a local run is complete. "none" stands for
# an empty suite list. Anything else fails.
ci_job_partition() {
  local job="$1" pipe="$2" suites rem
  suites="$(ci_suites_canonical "$3")" || return 1
  if [[ ! "$job" =~ ^(fast|gate|local)$ || ! "$pipe" =~ ^(fast|affected|integration|full)$ || -z "$suites" ]]; then return 1; fi
  case "$job" in
    local) printf '%s\t%s\tcomplete\tcomplete\n' "$pipe" "$suites" ;;
    fast)
      if [[ "$pipe" == affected ]]; then printf 'affected\t%s\tcomplete\tcomplete\n' "$suites"
      elif [[ "$pipe" == fast ]]; then printf 'fast\t%s\tcomplete\tcomplete\n' "$CI_SUITES_fast"
      else printf 'fast\t%s\tpartial\tfast\n' "$CI_SUITES_fast"; fi ;;
    gate)
      if [[ "$pipe" =~ ^(fast|affected)$ ]]; then printf 'none\tnone\tnone\tnone\n'; return 0; fi
      rem="$(ci_suites_minus "$suites" "$CI_SUITES_fast")" || return 1
      if [[ -z "$rem" ]]; then return 1; fi
      printf '%s\t%s\tpartial\tremainder\n' "$pipe" "$rem" ;;
  esac
  return 0
}

# ci_hosted_context: the run executes inside the reviewed hosted adapter context, not merely with a caller
# flag: GitLab CI with a numeric job, the adapter marker, the job-private TLS daemon on the shared network
# namespace (tcp://localhost:2376) and a readable client certificate at /certs/client.
ci_hosted_context() {
  if [[ "${GITLAB_CI:-}" == true && "${CI_JOB_ID:-}" =~ ^[1-9][0-9]{0,18}$ ]]; then :; else return 1; fi
  if [[ "${PLATFORM_CI_CONTEXT:-}" == "$CI_HOSTED_MARKER" ]]; then :; else return 1; fi
  if [[ "${DOCKER_HOST:-}" == tcp://localhost:2376 && "${DOCKER_TLS_VERIFY:-}" == 1 && "${DOCKER_CERT_PATH:-}" == /certs/client ]]; then :; else return 1; fi
  if [[ -f /certs/client/cert.pem && -r /certs/client/cert.pem && -f /.dockerenv ]]; then :; else return 1; fi
  return 0
}

# ci_disk_floor_gib PROFILE: the free-space floor that applies to this run.
ci_disk_floor_gib() {
  if ci_profile_known "$1"; then :; else return 1; fi
  local v="CI_DISKGIB_$1"
  if [[ "$1" =~ ^(affected|integration|full)$ ]] && ci_hosted_context; then printf '%s\n' "$CI_DISKGIB_HOSTED"; return 0; fi
  printf '%s\n' "${!v}"
}

# ci_capacity_ok PROFILE CPUS MEM_KIB TMP_FREE_KIB DAEMON_FREE_KIB: 0 when every measured value meets the
# profile's requirement. A non-integer measurement fails. DAEMON_FREE_KIB may be "same" when daemon storage
# is the temporary directory's filesystem.
ci_capacity_ok() {
  local p="$1" cpus="$2" memk="$3" tmpk="$4" dk="$5" need_cpu need_mem floor
  if ci_profile_known "$p"; then :; else return 1; fi
  if [[ ! "$cpus" =~ ^[0-9]{1,6}$ || ! "$memk" =~ ^[0-9]{1,15}$ || ! "$tmpk" =~ ^[0-9]{1,15}$ ]]; then return 1; fi
  if [[ "$dk" != same && ! "$dk" =~ ^[0-9]{1,15}$ ]]; then return 1; fi
  need_cpu="CI_CPUS_$p" need_mem="CI_MEMGIB_$p"
  floor="$(ci_disk_floor_gib "$p")" || return 1
  if (( 10#$cpus < ${!need_cpu} )); then return 1; fi
  if (( 10#$memk < ${!need_mem} * 1024 * 1024 )); then return 1; fi
  if (( 10#$tmpk < floor * 1024 * 1024 )); then return 1; fi
  if [[ "$dk" != same ]] && (( 10#$dk < floor * 1024 * 1024 )); then return 1; fi
  return 0
}

# ci_time_ok JOB TIMEOUT_S: CI_JOB_TIMEOUT meets the floor of job "fast", "gate" or "staging".
ci_time_ok() {
  local v="CI_JOBFLOOR_$1"
  if [[ ! "$1" =~ ^(fast|gate|staging)$ || ! "$2" =~ ^[0-9]{1,9}$ ]]; then return 1; fi
  (( 10#$2 >= ${!v} ))
}

# ci_started_epoch: CI_JOB_STARTED_AT (strict UTC "YYYY-MM-DDTHH:MM:SSZ") as epoch seconds.
ci_started_epoch() {
  local s="${CI_JOB_STARTED_AT:-}" e
  if [[ ! "$s" =~ ^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z$ ]]; then return 1; fi
  e="$(date -u -d "${s%Z} UTC" +%s 2>/dev/null)" || return 1
  if [[ ! "$e" =~ ^[0-9]{9,11}$ ]]; then return 1; fi
  printf '%s\n' "$e"
}

# ci_profile_budget_seconds PROFILE: min(CI_JOB_TIMEOUT, RUNNER_SCRIPT_TIMEOUT) - elapsed - reserve, capped
# per profile. In CI the elapsed time is measured from a valid CI_JOB_STARTED_AT (missing or malformed
# fails); a local run keeps CI_LOCAL_BUDGET_S. The caller enforces the profile minimum.
ci_profile_budget_seconds() {
  local p="$1" job="${CI_JOB_TIMEOUT:-}" script="${RUNNER_SCRIPT_TIMEOUT:-}" base rs start now el r c b
  if ci_profile_known "$p"; then :; else return 1; fi
  if [[ "${GITLAB_CI:-}" != true ]]; then ci_budget_seconds; return $?; fi
  if [[ ! "$job" =~ ^[0-9]{1,9}$ ]]; then return 1; fi
  base=$((10#$job))
  if [[ -n "$script" ]]; then
    rs="$(ci_go_duration_seconds "$script")" || return 1
    if (( rs < base )); then base="$rs"; fi
  fi
  start="$(ci_started_epoch)" || return 1
  now="$(date -u +%s)" || return 1
  el=$((now - start))
  if (( el < 0 )); then el=0; fi
  r="CI_RESERVE_$p" c="CI_CAP_$p"
  b=$((base - el - ${!r}))
  if (( ${!c} > 0 && b > ${!c} )); then b="${!c}"; fi
  printf '%s\n' "$b"
}

# ci_random_hex BYTES: lowercase hex from /dev/urandom.
ci_random_hex() { head -c "$1" /dev/urandom | od -An -tx1 | tr -d ' \n'; }

# ci_utc: the current UTC time, second precision.
ci_utc() { date -u +%Y-%m-%dT%H:%M:%SZ; }

# ci_path_within PATH ROOT: PATH is ROOT or lies below it (both already resolved).
ci_path_within() { [[ "$1" == "$2" || "$1" == "$2"/* ]]; }

# ci_json_str VALUE: VALUE as a JSON string literal (at most 512 characters; control characters other
# than tab, newline and carriage return are dropped).
ci_json_str() {
  local s="${1:0:512}"
  s="${s//\\/\\\\}"
  s="${s//\"/\\\"}"
  s="${s//$'\t'/\\t}"
  s="${s//$'\n'/\\n}"
  s="${s//$'\r'/\\r}"
  s="$(printf '%s' "$s" | tr -d '\000-\010\013\014\016-\037\177')"
  printf '"%s"' "$s"
}

# ci_go_duration_seconds DURATION: whole seconds of a Go duration made of integer h, m and s parts
# (for example 2h45m, 165m, 9900s, 2h45m0s). Anything else fails.
ci_go_duration_seconds() {
  local d="$1" h=0 m=0 s=0
  if [[ -z "$d" ]]; then return 1; fi
  if [[ "$d" =~ ^(([0-9]{1,6})h)?(([0-9]{1,6})m)?(([0-9]{1,9})s)?$ ]]; then
    if [[ -n "${BASH_REMATCH[2]}" ]]; then h=$((10#${BASH_REMATCH[2]})); fi
    if [[ -n "${BASH_REMATCH[4]}" ]]; then m=$((10#${BASH_REMATCH[4]})); fi
    if [[ -n "${BASH_REMATCH[6]}" ]]; then s=$((10#${BASH_REMATCH[6]})); fi
    printf '%s\n' "$((h * 3600 + m * 60 + s))"
    return 0
  fi
  return 1
}

# ci_budget_seconds: the harness deadline in seconds, min(CI_JOB_TIMEOUT, RUNNER_SCRIPT_TIMEOUT) minus the
# finalization reserve; 14400 when neither variable is set. Fails on a malformed value. The caller
# enforces the 3600-second minimum.
ci_budget_seconds() {
  local job="${CI_JOB_TIMEOUT:-}" script="${RUNNER_SCRIPT_TIMEOUT:-}" min="" rs
  if [[ -n "$job" ]]; then
    if [[ ! "$job" =~ ^[0-9]{1,9}$ ]]; then return 1; fi
    min=$((10#$job))
  fi
  if [[ -n "$script" ]]; then
    rs="$(ci_go_duration_seconds "$script")" || return 1
    if [[ -z "$min" ]] || (( rs < min )); then min="$rs"; fi
  fi
  if [[ -z "$min" ]]; then printf '%s\n' "$CI_LOCAL_BUDGET_S"; return 0; fi
  printf '%s\n' "$((min - CI_BUDGET_RESERVE_S))"
}

# state_import FILE KEY=REGEX...: the only path from an isolated phase to parent-shell state. FILE must
# be a regular, non-symlink file of KEY=VALUE lines; every declared key exactly once, no other key,
# each value matching its regular expression. Values are assigned with printf -v only after the whole
# file has been validated (never source or eval). Any violation returns 16 and assigns nothing.
state_import() {
  local f="$1" line k v spec
  shift
  local -A want=() got=()
  for spec in "$@"; do
    if [[ ! "${spec%%=*}" =~ ^[A-Z][A-Z0-9_]*$ ]]; then return 16; fi
    want[${spec%%=*}]="${spec#*=}"
  done
  if [[ -L "$f" || ! -f "$f" ]]; then return 16; fi
  while IFS= read -r line || [[ -n "$line" ]]; do
    if [[ ! "$line" =~ ^([A-Z][A-Z0-9_]*)=(.*)$ ]]; then return 16; fi
    k="${BASH_REMATCH[1]}" v="${BASH_REMATCH[2]}"
    if [[ -z "${want[$k]+x}" || -n "${got[$k]+x}" ]]; then return 16; fi
    if [[ ! "$v" =~ ${want[$k]} ]]; then return 16; fi
    got[$k]="$v"
  done <"$f"
  for k in "${!want[@]}"; do
    if [[ -z "${got[$k]+x}" ]]; then return 16; fi
  done
  for k in "${!got[@]}"; do printf -v "$k" '%s' "${got[$k]}"; done
  return 0
}

# ci_docker_start_inventory FILE: every container, network, volume and image ID present now, as
# "KIND ID" lines (full IDs; volume names), written atomically. A resource listed here is never removed.
ci_docker_start_inventory() {
  local out="$1" c n v i x
  c="$(docker ps -aq --no-trunc)" || return 1
  n="$(docker network ls -q --no-trunc)" || return 1
  v="$(docker volume ls -q)" || return 1
  i="$(docker image ls -aq --no-trunc)" || return 1
  {
    for x in $c; do printf 'container %s\n' "$x"; done
    for x in $n; do printf 'network %s\n' "$x"; done
    for x in $v; do printf 'volume %s\n' "$x"; done
    for x in $i; do printf 'image %s\n' "$x"; done
  } | LC_ALL=C sort -u >"$out.tmp" || return 1
  mv -f -- "$out.tmp" "$out" || return 1
  return 0
}

# ci_registry_dir CI_PROJECT_DIR CI_JOB_ID: the CI registry location, a sibling of the checkout. The
# parent directory is resolved; the last component is not, so a symlinked registry stays detectable.
ci_registry_dir() {
  local p parent
  p="$(realpath -m -- "$1" 2>/dev/null)" || return 1
  parent="$(dirname -- "$p")"
  if [[ "$parent" == / ]]; then parent=""; fi
  printf '%s/.platform-validation-owned-%s\n' "$parent" "$2"
}

# Owner direction 2026-10-14: the expensive CI meta-tests have their own suite.
# Return 0 when required, 1 only for an attributable unrelated delta. Unknowns run.
# The shared root Vestive skill (2026-10-18) keeps the selection of its former web/.claude/ location.
# ci_controls_required REPO BASE [TARGET] [TESTED_CONTENT]: a verified RC-promotion merge request (TESTED_CONTENT
# mr-rc-promotion, which only validate.sh's attribution establishes and passes on) always requires them (ONEOFF RC
# qualification): the promotion pipeline is the complete full release qualification, whatever CI_CONTROLS_FORCE says.
ci_controls_required() {
  local repo="$1" base="$2" target="${3:-HEAD}" p resolved fallback
  local -a ps=()
  if [[ "${4:-}" == mr-rc-promotion ]]; then return 0; fi
  if [[ "${CI_CONTROLS_FORCE:-0}" != 0 || "${CI_PIPELINE_SOURCE:-}" == schedule ]]; then return 0; fi
  resolved="$(git --no-optional-locks -C "$repo" rev-parse --verify "$target^{commit}" 2>/dev/null)" || return 0
  if [[ "$base" == none || "$base" == "$resolved" || "$base" == HEAD ]]; then
    fallback="$(git --no-optional-locks -C "$repo" merge-base "$resolved" origin/staging 2>/dev/null)" || fallback=""
    if [[ -n "$fallback" && "$fallback" != "$resolved" ]]; then base="$fallback"
    else base="$(git --no-optional-locks -C "$repo" rev-parse --verify "$resolved^" 2>/dev/null)" || return 0; fi
  fi
  git --no-optional-locks -C "$repo" rev-parse --verify "$base^{commit}" >/dev/null 2>&1 || return 0
  git --no-optional-locks -C "$repo" merge-base --is-ancestor "$base" "$resolved" 2>/dev/null || return 0
  if (
    git --no-optional-locks -C "$repo" diff --name-only -z --no-renames "$base" "$resolved" -- . ":(top,exclude).agents/skills/restruct-sdlc" ":(top,exclude).claude/skills/restruct-sdlc" || exit 2
    git --no-optional-locks -C "$repo" diff --name-only -z --no-renames "$resolved" -- . ":(top,exclude).agents/skills/restruct-sdlc" ":(top,exclude).claude/skills/restruct-sdlc" || exit 2
    git --no-optional-locks -C "$repo" ls-files -z --others --exclude-standard -- . ":(top,exclude).agents/skills/restruct-sdlc" ":(top,exclude).claude/skills/restruct-sdlc" || exit 2
  ) | (
    while IFS= read -r -d '' p; do
      case "$p" in
        infra/*|contracts/*|distributions/skills/*|.gitlab-ci.yml|.env.example|.gitignore|.dockerignore|README.md|Dockerfile|*/Dockerfile|*/.dockerignore|*.Dockerfile|core/go.mod|core/go.sum|web/package.json|web/package-lock.json|intelligence/pyproject.toml|intelligence/uv.lock|web/.claude/*|web/.agents/*|.claude/skills/vestive-brand|.agents/skills/vestive-brand/*) exit 0 ;;
        core/*|web/*|intelligence/*|AGENTS.md|CLAUDE.md|.gitmodules|LICENSE|docs/*.md|management/*.md|.agents/*|.claude/*) ;;
        *) exit 0 ;;
      esac
    done
    exit 1
  ); then return 0
  else
    ps=("${PIPESTATUS[@]}")
    if [[ "${ps[0]:-2}" == 0 && "${ps[1]:-2}" == 1 ]]; then return 1; fi
    return 0
  fi
}
