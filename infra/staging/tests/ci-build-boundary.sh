#!/usr/bin/env bash
# G072 EP-1: the credential-free build boundary of the manual staging CI (Plan r0 §1.3 EP-1, r1 §10.12; Spec R10).
# Coordinator-run in the single build slot, unregistered (not a harness suite or check ID); its results.tsv lines are
# recorded in the G072 Result with command, exit code and commit.
#
#   bash infra/staging/tests/ci-build-boundary.sh --work DIR --report DIR [--only CASE[,CASE]...] [--allow-dirty]
#        [--max-artifact-mb N] [--prepare-timeout MIN] [--build-timeout MIN] [--env NAME=VALUE]...
#   bash infra/staging/tests/ci-build-boundary.sh --list
#   bash infra/staging/tests/ci-build-boundary.sh --dry-run [--work DIR] [--only CASE[,CASE]...]
#
# It drives the REAL prepare and build phases of the committed control adapter, `bash infra/staging/ci.sh
# prepare|build`, from a disposable clone of HEAD, with the launcher/adapter environment of Plan §10.10 (STAGING_CI_*,
# the public CI metadata, STAGING_REUSE=no for build) and a job environment that carries canaries: CI_JOB_TOKEN (also
# in CI_REGISTRY_PASSWORD, CI_REPOSITORY_URL and the control checkout's .git/config, as a runner leaves it), a fake
# STAGING_CI_ID_TOKEN and CLOUDFLARE_API_TOKEN, plus a planted canary in untracked project files and the job's Docker
# client configuration. Prepare's STAGING_CI_OUT is handed to build in place, as GitLab restores artifacts at the same
# path. The candidates come from a synthetic same-repository origin (a bare repository under DIR holding `staging` =
# HEAD and three fixture branches built from infra/staging/tests/fixtures/ci-build-boundary/), selected for preview-1 by
# branch name; the baseline is the default (empty) candidate for staging, i.e. the frozen CI_COMMIT_SHA.
#
# Cases (in run order; Docker cases imply fixture):
#   fixture           (git only) the control clone, origin and branches; each branch differs from HEAD in exactly its
#                     one fixture path; the Compose path is a compatibility (S6) path, the Dockerfile path is not
#   scan-controls     (no Docker) mutation controls of the canary scanner: raw, base64 at each alignment, inside gzip,
#                     inside a tar and in a nested tree are each reported; a clean file is not; no hit line holds a value
#   compose-refused   the hostile Compose variant (build args from the job's credentials, /builds, /certs and HOME as
#                     additional contexts) is refused by prepare with exit 2 naming the incompatible infra/compose path,
#                     before any build: no image is created, no build manifest or archive exists; build run anyway on
#                     what prepare left exits 2 and creates no image; no canary in any output
#   baseline-measure  prepare + build of the baseline (staging, empty candidate): the G070-format manifest names the
#                     seven services and the pinned SHA; no canary anywhere (as dockerfile-clean); A5 measurement: every
#                     artifact's size and the total against --max-artifact-mb; image-ID preservation: this tooling's
#                     tags removed, the archives loaded (timed), every manifest ID present again, the restored tags and
#                     an ID-preserving retag resolve to the manifest IDs (measure.tsv)
#   dockerfile-clean  prepare + build of the hostile Dockerfile variant: the manifest records the hostile Dockerfile's
#                     hash; its dumps exist in the image and saw no build argument; no canary (raw or base64) in the
#                     `docker export` of every image, their inspect/history metadata, prepare/build stdout and stderr,
#                     STAGING_CI_OUT (archives decompressed), STAGING_CI_PRIVATE, the job home or new project files
#   mutant-detected   mutation control: the same variant with the CI_JOB_TOKEN canary baked in on purpose; the scan must
#                     report it in the migrate image's export and history (hits kept apart from the clean cases)
#
# Stop criterion (Plan §10.12, Codex A5 decision): any canary hit in a clean case or a total artifact size above the
# bound records STOP, skips the remaining Docker cases and exits 1; return to Codex for a reviewed Plan change. Without
# infra/staging/ci.sh at HEAD a run selecting a Docker case refuses (exit 65, "adapter not present").
#
# --list and --dry-run run no Docker and create nothing: --list names the cases, --dry-run prints the plan.
# Report: REPORT/results.tsv (case, result, exit, commit, detail), measure.tsv, resources.tsv, hits.tsv (label, canary
# name, encoding, count: never a value), logs/ (adapter output of clean cases, copied only after a clean scan; mutant
# output is never copied). Canary values live only in a 0700 scratch directory under DIR, removed at exit. Containers
# carry g072.task=G072 and g072.boundary.run=<run>; image tags this run's builds and loads created are removed by exact
# reference and recorded ID, and their absence verified.
# Exit: 0 every selected case PASS; 1 any FAIL, STOP or SKIP; 64 usage; 65 refused (adapter absent, unsafe path, tree
# differs from HEAD).
set -euo pipefail

REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd -P)"
FIX="$REPO/infra/staging/tests/fixtures/ci-build-boundary"
CASES=(fixture scan-controls compose-refused baseline-measure dockerfile-clean mutant-detected)
DOCKER_CASES=(compose-refused baseline-measure dockerfile-clean mutant-detected)
# The deployment-infrastructure paths a candidate must share with control (Plan S6).
S6_PATHS=(infra/compose infra/postgres infra/minio infra/edge infra/staging infra/docker/edge.Dockerfile infra/docker/minio.Dockerfile)
SERVICES=(core web intelligence minio minio-init migrate edge)
BR_COMPOSE=g072-boundary-hostile-compose BR_DOCKERFILE=g072-boundary-hostile-dockerfile BR_MUTANT=g072-boundary-mutant-leak
CANARY_NAMES=(CI_JOB_TOKEN STAGING_CI_ID_TOKEN CLOUDFLARE_API_TOKEN PLANTED_FILE)

usage() { echo "usage: ci-build-boundary.sh --work DIR --report DIR [--only CASE[,CASE]...] [--allow-dirty] [--max-artifact-mb N] [--prepare-timeout MIN] [--build-timeout MIN] [--env NAME=VALUE]... | --list | --dry-run [--work DIR] [--only CASE[,CASE]...]" >&2; exit 64; }
die() { echo "ci-build-boundary.sh: $*" >&2; exit 65; }
log() { printf '[boundary %s] %s\n' "$(date -u +%H:%M:%S)" "$*" >&2; }

WORK="" REPORT="" ONLY="" LIST=0 DRY=0 ALLOW_DIRTY=0 MAX_MB=1024 PREP_MIN=10 BUILD_MIN=45
EXTRA_ENV=()
while [[ $# -gt 0 ]]; do
  case "$1" in
    --work) [[ $# -ge 2 ]] || usage; WORK="$2"; shift 2 ;;
    --report) [[ $# -ge 2 ]] || usage; REPORT="$2"; shift 2 ;;
    --only) [[ $# -ge 2 ]] || usage; ONLY="$2"; shift 2 ;;
    --list) LIST=1; shift ;;
    --dry-run) DRY=1; shift ;;
    --allow-dirty) ALLOW_DIRTY=1; shift ;;
    --max-artifact-mb) [[ $# -ge 2 && "$2" =~ ^[1-9][0-9]{0,5}$ ]] || usage; MAX_MB="$2"; shift 2 ;;
    --prepare-timeout) [[ $# -ge 2 && "$2" =~ ^[1-9][0-9]{0,2}$ ]] || usage; PREP_MIN="$2"; shift 2 ;;
    --build-timeout) [[ $# -ge 2 && "$2" =~ ^[1-9][0-9]{0,2}$ ]] || usage; BUILD_MIN="$2"; shift 2 ;;
    # Interface reconciliation after integration: one more adapter variable (never a canary carrier).
    --env) [[ $# -ge 2 && "$2" =~ ^[A-Z][A-Z0-9_]*=[A-Za-z0-9._:/@+-]*$ && ! "${2%%=*}" =~ ^(CI_JOB_TOKEN|CI_REGISTRY_PASSWORD|CI_REPOSITORY_URL|STAGING_CI_ID_TOKEN|CLOUDFLARE_API_TOKEN|HOME|PATH|DOCKER_CONFIG)$ ]] || usage
      EXTRA_ENV+=("$2"); shift 2 ;;
    *) usage ;;
  esac
done

SEL=()
if [[ -n "$ONLY" ]]; then
  IFS=, read -ra want <<<"$ONLY"
  for c in "${want[@]}"; do [[ " ${CASES[*]} " == *" $c "* ]] || { echo "--only: unknown case $c" >&2; exit 64; }; done
  for c in "${CASES[@]}"; do [[ ",$ONLY," == *",$c,"* ]] && SEL+=("$c"); done
else SEL=("${CASES[@]}"); fi
selected() { [[ " ${SEL[*]} " == *" $1 "* ]]; }
NEED_DOCKER=0
for c in "${DOCKER_CASES[@]}"; do selected "$c" && NEED_DOCKER=1; done
if (( NEED_DOCKER )) && ! selected fixture; then SEL=(fixture "${SEL[@]}"); fi

describe() {
  case "$1" in
    fixture) echo "git only: control clone of HEAD, bare origin (staging + 3 fixture branches), one-path diffs, S6 membership" ;;
    scan-controls) echo "no Docker: canary scanner mutation controls (raw, base64 x3, gzip, tar, nested tree, clean file, no value in hits)" ;;
    compose-refused) echo "hostile Compose variant: prepare exit 2 (incompatible infra/compose), no image/manifest/archive; build on its leftovers exit 2" ;;
    baseline-measure) echo "baseline staging tip: prepare+build, no canary, artifact sizes vs --max-artifact-mb, ID preservation across save/load/retag" ;;
    dockerfile-clean) echo "hostile Dockerfile variant: dumps present, no build argument, no canary in exports, metadata, logs, artifacts" ;;
    mutant-detected) echo "mutation control: canary baked into the migrate image must be reported in its export and history" ;;
  esac
}
if (( LIST )); then
  for c in "${CASES[@]}"; do printf '%-17s %s\n' "$c" "$(describe "$c")"; done
  exit 0
fi

COMMIT="$(git -C "$REPO" rev-parse HEAD)"
ADAPTER_AT_HEAD=0
git -C "$REPO" cat-file -e "HEAD:infra/staging/ci.sh" 2>/dev/null && ADAPTER_AT_HEAD=1

# ---- dry run: the plan, nothing executed -----------------------------------------------------------------------------
if (( DRY )); then
  w="${WORK:-<WORK>}/boundary-<run>" ctl="${WORK:-<WORK>}/boundary-<run>/control"
  jobenv="env: only PATH HOME DOCKER_CONFIG DOCKER_HOST LANG CI=true GITLAB_CI=true CI_PROJECT_DIR=$ctl CI_COMMIT_SHA=$COMMIT CI_COMMIT_BRANCH=staging CI_COMMIT_REF_PROTECTED=true CI_PIPELINE_SOURCE=web CI_PIPELINE_ID CI_JOB_ID CI_JOB_NAME=staging:<action> GITLAB_USER_LOGIN STAGING_CI_ACTION STAGING_CI_ENV STAGING_CI_CANDIDATE_INPUT STAGING_CI_DURATION_INPUT= STAGING_CI_PRIVATE STAGING_CI_OUT [STAGING_REUSE=no] CI_JOB_TOKEN=<canary> CI_REGISTRY_PASSWORD=<canary> CI_REPOSITORY_URL=<url with canary> STAGING_CI_ID_TOKEN=<canary> CLOUDFLARE_API_TOKEN=<canary>${EXTRA_ENV[*]:+ ${EXTRA_ENV[*]}}"
  echo "# ci-build-boundary.sh dry run at $COMMIT: no Docker, nothing created"
  if (( ADAPTER_AT_HEAD )); then echo "# adapter infra/staging/ci.sh: present at HEAD"
  else echo "# adapter infra/staging/ci.sh: NOT present at HEAD; a run selecting a Docker case refuses (exit 65, adapter not present)"; fi
  echo "# bounds: total artifacts <= ${MAX_MB} MiB, prepare ${PREP_MIN} min, build ${BUILD_MIN} min"
  for c in "${SEL[@]}"; do
    echo "[$c] $(describe "$c")"
    case "$c" in
      fixture)
        echo "  + git clone --quiet --no-tags --single-branch $REPO $ctl; git -C $ctl checkout --detach $COMMIT; drop clone refs and remote"
        echo "  + git init --bare $w/origin.git; push $COMMIT:refs/heads/staging"
        echo "  + plant canaries: http.extraHeader in $ctl/.git/config, untracked $ctl/.g072-boundary-planted and $ctl/core/.g072-boundary-planted, job home .docker/config.json"
        echo "  + commit-tree branches: $BR_DOCKERFILE (infra/docker/migrate.Dockerfile = migrate.hostile.Dockerfile), $BR_MUTANT (+ migrate.mutant.fragment with the CI_JOB_TOKEN canary), $BR_COMPOSE (infra/compose/local.yml + local.yml.core-build.fragment after core's dockerfile line); push each"
        echo "  + assert git diff --name-only $COMMIT <branch> is the one fixture path; S6 membership: compose yes, migrate.Dockerfile no" ;;
      scan-controls) echo "  + write raw/base64/gzip/tar/nested/clean samples under the scratch dir; scan each; expect hits exactly in the mutated samples and no value in the hit lines" ;;
      compose-refused)
        echo "  + docker image ls -a -q --no-trunc (before)"
        echo "  + (cd $ctl; $jobenv timeout ${PREP_MIN}m bash infra/staging/ci.sh prepare)  STAGING_CI_ENV=preview-1 STAGING_CI_CANDIDATE_INPUT=$BR_COMPOSE; expect exit 2, 'incompatible candidate' and infra/compose"
        echo "  + (cd $ctl; ... timeout ${BUILD_MIN}m bash infra/staging/ci.sh build) on prepare's leftovers; expect exit 2"
        echo "  + docker image ls -a -q --no-trunc (after) == before; no manifest or archive in STAGING_CI_OUT; canary scan of every output" ;;
      baseline-measure|dockerfile-clean|mutant-detected)
        case "$c" in baseline-measure) e=staging cand="" ;; dockerfile-clean) e=preview-1 cand=$BR_DOCKERFILE ;; *) e=preview-1 cand=$BR_MUTANT ;; esac
        echo "  + (cd $ctl; $jobenv timeout ${PREP_MIN}m bash infra/staging/ci.sh prepare)  STAGING_CI_ENV=$e STAGING_CI_CANDIDATE_INPUT='$cand'; expect exit 0"
        echo "  + (cd $ctl; ... timeout ${BUILD_MIN}m bash infra/staging/ci.sh build)  STAGING_REUSE=no; expect exit 0"
        echo "  + find the G070-format manifest in STAGING_CI_OUT (seven services, candidate = pinned SHA)"
        echo "  + per manifest image: docker create --network none --label g072.task=G072 --label g072.boundary.run=<run> <id>; docker export | canary scan; docker rm"
        echo "  + per manifest image: docker image inspect <id> | canary scan; docker history --no-trunc <id> | canary scan"
        echo "  + canary scan: prepare/build stdout+stderr, STAGING_CI_OUT (gzip -dc of archives), STAGING_CI_PRIVATE, job home, new project files"
        case "$c" in
          baseline-measure) echo "  + sizes of STAGING_CI_OUT files; total <= ${MAX_MB} MiB else STOP; docker image rm <this tooling's tags>; docker load -i <archive> (timed); IDs present; tags and docker tag <id> g072-boundary-retag.invalid/<svc>:<tag> resolve to the manifest IDs" ;;
          dockerfile-clean) echo "  + manifest migrate Dockerfile sha256 == migrate.hostile.Dockerfile; export lists leak/env.txt leak/proc-environ.txt leak/mounts.txt leak/build-args.txt; build-args.txt == 'job=unset id=unset cf=unset repo=unset'" ;;
          *) echo "  + expect hits in the migrate export and history (mutant-hits.tsv); output never copied" ;;
        esac ;;
    esac
  done
  echo "# cleanup: containers by label g072.boundary.run=<run>; recorded image tags by exact reference and ID; the scratch directory; absence verified (resources.tsv)"
  exit 0
fi

# ---- run: setup ------------------------------------------------------------------------------------------------------
[[ -n "$WORK" && -n "$REPORT" ]] || usage
(( ! NEED_DOCKER || ADAPTER_AT_HEAD )) ||
  die "adapter not present: infra/staging/ci.sh is missing at HEAD $COMMIT; EP-1 drives the real prepare/build phases and runs after the adapter is integrated"
for f in migrate.hostile.Dockerfile migrate.mutant.fragment local.yml.core-build.fragment; do [[ -f "$FIX/$f" ]] || die "fixture $FIX/$f is missing"; done
if (( NEED_DOCKER )) && [[ -n "$(git -C "$REPO" status --porcelain --untracked-files=all -- infra)" ]]; then
  (( ALLOW_DIRTY )) || die "infra/ differs from HEAD (the adapter runs from a clone of HEAD); commit first or pass --allow-dirty"
  COMMIT_NOTE="$COMMIT+dirty-fixtures"
else COMMIT_NOTE="$COMMIT"; fi
# A path inside this checkout must be Git-ignored (the main checkout's .local/); anything else would change the tree.
safe_dir() {
  mkdir -p "$1"; local p; p="$(cd "$1" && pwd -P)"
  if [[ "$p/" == "$REPO/"* ]] && ! git -C "$REPO" check-ignore -q "${p#"$REPO"/}/"; then die "$1 is inside the checkout and not ignored"; fi
  printf '%s\n' "$p"
}
WORK="$(safe_dir "$WORK")" REPORT="$(safe_dir "$REPORT")"
[[ ! -e "$REPORT/results.tsv" ]] || die "$REPORT/results.tsv exists; use a new report directory per run"
random_hex() { head -c "$1" /dev/urandom | od -An -tx1 | tr -d ' \n'; }
RUN_ID="$(date -u +%Y%m%dt%H%M%Sz)-$(random_hex 4)"
RESULTS="$REPORT/results.tsv" MEASURE="$REPORT/measure.tsv" RES="$REPORT/resources.tsv" HITS="$REPORT/hits.tsv"
printf 'case\tresult\texit\tcommit\tdetail\n' >"$RESULTS"
printf 'case\tmetric\tvalue\n' >"$MEASURE"
printf 'kind\tname\tcreated_utc\tremoved_utc\tstatus\n' >"$RES"
printf 'label\tcanary\tencoding\tcount\n' >"$HITS"
mkdir -p "$REPORT/logs"
S="$WORK/boundary-$RUN_ID"; (umask 077; mkdir -p "$S")
printf 'scratch\t%s\t%s\t\tallocated\n' "$S" "$(date -u +%FT%TZ)" >>"$RES"
REFS="$S/refs.tsv"; : >"$REFS"
HITS_FILE="$HITS" STOPPED="" FAILED=0 NEED_CLEANUP_DOCKER=$NEED_DOCKER
CONTROL="$S/control" ORIGIN="$S/origin.git"
declare -A BR_SHA=()

cleanup() {
  local rc=$? ids n ref id have
  if (( NEED_CLEANUP_DOCKER )); then
    ids="$(docker ps -aq --filter "label=g072.boundary.run=$RUN_ID" 2>/dev/null || true)"
    n="$(grep -c . <<<"$ids" || true)"
    [[ -z "$ids" ]] || xargs -r docker rm -f <<<"$ids" >/dev/null 2>&1 || true
    if [[ -n "$(docker ps -aq --filter "label=g072.boundary.run=$RUN_ID" 2>/dev/null || true)" ]]; then
      printf 'container\tlabel g072.boundary.run=%s\t\t\tLEFTOVER (removal failed)\n' "$RUN_ID" >>"$RES"; rc=1
    else printf 'container\tlabel g072.boundary.run=%s\t\t%s\tabsent (swept %s)\n' "$RUN_ID" "$(date -u +%FT%TZ)" "$n" >>"$RES"; fi
    # Image tags: only references this run recorded, and only while they still name the recorded image.
    while IFS=$'\t' read -r ref id; do
      [[ -n "$ref" ]] || continue
      have="$(docker image inspect --format '{{.Id}}' "$ref" 2>/dev/null || true)"
      if [[ -z "$have" ]]; then printf 'image-ref\t%s\t\t%s\tabsent\n' "$ref" "$(date -u +%FT%TZ)" >>"$RES"
      elif [[ "$have" != "$id" ]]; then printf 'image-ref\t%s\t\t\tleft (now names another image; not this run'\''s)\n' "$ref" >>"$RES"
      elif docker image rm "$ref" >/dev/null 2>&1 && [[ -z "$(docker image inspect --format '{{.Id}}' "$ref" 2>/dev/null || true)" ]]; then
        printf 'image-ref\t%s\t\t%s\tremoved\n' "$ref" "$(date -u +%FT%TZ)" >>"$RES"
      else printf 'image-ref\t%s\t\t\tLEFTOVER (removal failed)\n' "$ref" >>"$RES"; rc=1; fi
    done < <(sort -u "$REFS" 2>/dev/null || true)
  fi
  rm -rf "$S" 2>/dev/null || true
  if [[ -e "$S" ]]; then printf 'scratch\t%s\t\t\tLEFTOVER\n' "$S" >>"$RES"; rc=1
  else printf 'scratch\t%s\t\t%s\tremoved\n' "$S" "$(date -u +%FT%TZ)" >>"$RES"; fi
  exit "$rc"
}
trap cleanup EXIT
trap 'exit 130' INT TERM HUP

result() { # CASE RESULT EXIT DETAIL
  printf '%s\t%s\t%s\t%s\t%s\n' "$1" "$2" "$3" "$COMMIT_NOTE" "$4" >>"$RESULTS"
  [[ "$2" == PASS ]] || FAILED=1
  log "$1: $2 ($4)"
}
measure() { printf '%s\t%s\t%s\n' "$1" "$2" "$3" >>"$MEASURE"; }

# ---- canaries and the scanner ----------------------------------------------------------------------------------------
# Each canary: G072BND + carrier tag + 32 hex characters, fresh per run, never written outside $S. Patterns: the raw
# value and, for each of the three base64 alignments, the characters that depend on the value alone.
declare -A CANARY=()
CANARY[CI_JOB_TOKEN]="G072BNDJOB$(random_hex 16)"
CANARY[STAGING_CI_ID_TOKEN]="G072BNDOIDC$(random_hex 16)"
CANARY[CLOUDFLARE_API_TOKEN]="G072BNDCF$(random_hex 16)"
CANARY[PLANTED_FILE]="G072BNDFILE$(random_hex 16)"
b64core() { # PREFIX_LEN VALUE
  local p=$1 v=$2 s g0 g1
  s="$(printf '%s%s' "${FILLER:0:p}" "$v" | base64 -w0)"
  g0=$(( p > 0 ? 1 : 0 )); g1=$(( (p + ${#v}) / 3 ))
  printf '%s' "${s:$((4 * g0)):$((4 * (g1 - g0)))}"
}
FILLER="xx" PATTERNS="$S/patterns" PATTERN_NAMES="$S/pattern-names"
( umask 077; : >"$PATTERNS"; : >"$PATTERN_NAMES"
  i=0
  for name in "${CANARY_NAMES[@]}"; do
    for enc in raw b64-0 b64-1 b64-2; do
      i=$((i + 1))
      case "$enc" in raw) v="${CANARY[$name]}" ;; *) v="$(b64core "${enc#b64-}" "${CANARY[$name]}")" ;; esac
      printf '%s\n' "$v" >>"$PATTERNS"; printf '%s\n' "$v" >"$S/pat.$i"; printf '%s\t%s\n' "$name" "$enc" >>"$PATTERN_NAMES"
    done
  done )
# An empty or short pattern would match everything (grep -F) or too much.
! awk 'length($0) < 32 {f = 1} END {exit !f}' "$PATTERNS" || die "a canary pattern is shorter than 32 characters"
capped() { (( $1 > 255 )) && return 255; return "$1"; } # a count as an exit status that never wraps to 0

# scan_stream LABEL CMD...: scans CMD's stdout without keeping it; 0 clean, 1 hit (one hits line per canary and
# encoding: label, name, encoding, count; never the value), 2 unscannable (the command failed; counted as a failure).
scan_stream() {
  local label=$1 i=0 n name enc st
  shift
  set +e; "$@" 2>/dev/null | grep -a -c -F -f "$PATTERNS" >"$S/scan.count"; st=("${PIPESTATUS[@]}"); set -e
  n="$(cat "$S/scan.count")"
  if (( st[0] != 0 )) || [[ ! "$n" =~ ^[0-9]+$ ]]; then printf '%s\t-\tunscannable\t-\n' "$label" >>"$HITS_FILE"; return 2; fi
  (( n > 0 )) || return 0
  while IFS=$'\t' read -r name enc; do
    i=$((i + 1))
    set +e; "$@" 2>/dev/null | grep -a -c -F -f "$S/pat.$i" >"$S/scan.count"; set -e
    n="$(cat "$S/scan.count")"
    [[ "$n" =~ ^[0-9]+$ ]] && (( n == 0 )) || printf '%s\t%s\t%s\t%s\n' "$label" "$name" "$enc" "${n:-?}" >>"$HITS_FILE"
  done <"$PATTERN_NAMES"
  return 1
}
# scan_tree LABEL DIR [EXCLUDED_RELATIVE_PATH]...: every regular file (symlinks not followed), and gzip content of *.gz
# and *.tgz; returns the number of dirty or unscannable files (0 clean).
scan_tree() {
  local label=$1 dir=$2 bad=0 f rel
  shift 2
  [[ -d "$dir" ]] || return 0
  while IFS= read -r -d '' f; do
    rel="${f#"$dir"/}"
    [[ " $* " == *" $rel "* ]] && continue
    scan_stream "$label $rel" cat -- "$f" || bad=$((bad + 1))
    case "$f" in *.gz|*.tgz) scan_stream "$label $rel (gunzip)" gzip -dc -- "$f" || bad=$((bad + 1)) ;; esac
  done < <(find "$dir" -type f -print0 | sort -z)
  capped "$bad"
}

# ---- case: scan-controls (no Docker) ---------------------------------------------------------------------------------
case_scan_controls() {
  local d="$S/scan-controls" why="" ok=1 p name save=$HITS_FILE
  mkdir -p "$d/art" "$d/tarsrc/leak" "$d/tree/a/b"
  HITS_FILE="$d/hits.tsv"; : >"$HITS_FILE"
  printf 'clean line\nG072BND without a value\n' >"$d/clean.log"
  { cat "$d/clean.log"; printf 'before%safter\n' "${CANARY[CI_JOB_TOKEN]}"; } >"$d/raw.log"
  # base64 at each alignment, and after the 16-byte prefix a Docker client credential carries (alignment 1).
  for p in 0 1 2; do printf 'auth: %s tail\n' "$(printf '%s%s' "${FILLER:0:p}" "${CANARY[STAGING_CI_ID_TOKEN]}" | base64 -w0)" >"$d/b64-$p.log"; done
  printf '"auth":"%s"\n' "$(printf 'gitlab-ci-token:%s' "${CANARY[STAGING_CI_ID_TOKEN]}" | base64 -w0)" >"$d/b64-credential.log"
  printf '%s\n' "${CANARY[CLOUDFLARE_API_TOKEN]}" | gzip >"$d/art/leak.txt.gz"
  printf 'K=%s\n' "${CANARY[PLANTED_FILE]}" >"$d/tarsrc/leak/env.txt"; tar -cf "$d/export.tar" -C "$d/tarsrc" .
  printf '%s\n' "${CANARY[CI_JOB_TOKEN]}" >"$d/tree/a/b/raw.log"
  printf '%s\n' "${CANARY[CLOUDFLARE_API_TOKEN]}" | gzip >"$d/tree/images.tar.gz"
  expect() { # WANT(clean|hit|unscannable) NAME_OR_- LABEL CMD...
    local verdict=$1 nm=$2 label=$3 rc=0 wrc; shift 3
    case "$verdict" in clean) wrc=0 ;; hit) wrc=1 ;; *) wrc=2 ;; esac
    scan_stream "$label" "$@" || rc=$?
    (( rc == wrc )) || { ok=0; why+="$label: rc $rc, want $wrc ($verdict); "; return 0; }
    [[ $verdict != hit ]] || awk -F'\t' -v l="$label" -v n="$nm" '$1 == l && $2 == n {f = 1} END {exit !f}' "$HITS_FILE" ||
      { ok=0; why+="$label: hit not attributed to $nm; "; }
  }
  expect clean - clean.log cat -- "$d/clean.log"
  expect hit CI_JOB_TOKEN raw.log cat -- "$d/raw.log"
  for p in 0 1 2; do expect hit STAGING_CI_ID_TOKEN "b64-$p.log" cat -- "$d/b64-$p.log"; done
  expect hit STAGING_CI_ID_TOKEN b64-credential.log cat -- "$d/b64-credential.log"
  expect clean - "leak.txt.gz (compressed bytes)" cat -- "$d/art/leak.txt.gz"
  expect hit CLOUDFLARE_API_TOKEN "leak.txt.gz (gunzip)" gzip -dc -- "$d/art/leak.txt.gz"
  expect hit PLANTED_FILE export.tar cat -- "$d/export.tar"
  expect unscannable - "missing file" cat -- "$d/absent" # an unreadable input is never clean
  local tb=0; scan_tree tree "$d/tree" || tb=$?
  (( tb == 2 )) || { ok=0; why+="nested tree: $tb dirty files, want 2 (a/b/raw.log, images.tar.gz content); "; }
  tb=0; scan_tree excluded "$d/tree" a/b/raw.log images.tar.gz || tb=$?
  (( tb == 0 )) || { ok=0; why+="nested tree with the file excluded: $tb dirty files, want 0; "; }
  for name in "${CANARY_NAMES[@]}"; do
    if grep -qF -- "${CANARY[$name]}" "$HITS_FILE"; then ok=0; why+="a hit line holds the $name value; "; fi
  done
  HITS_FILE=$save
  if (( ok )); then result scan-controls PASS - "clean file clean; raw, base64 at alignments 0/1/2 and after a credential prefix, gzip content (not its compressed bytes), tar member, nested-tree file and nested gzip archive canaries each reported and attributed; excluded paths skipped; an unreadable input unscannable (rc 2, never clean); no value in $(grep -c . "$d/hits.tsv") hit lines"
  else result scan-controls FAIL - "$why"; fi
}

# ---- fixture (git only) ----------------------------------------------------------------------------------------------
fixture_commit() { # TREE MESSAGE -> commit SHA (fixed fixture identity and date)
  env GIT_AUTHOR_NAME="G072 boundary fixture" GIT_AUTHOR_EMAIL=fixture@example.invalid GIT_AUTHOR_DATE="2026-10-28T00:00:00Z" \
    GIT_COMMITTER_NAME="G072 boundary fixture" GIT_COMMITTER_EMAIL=fixture@example.invalid GIT_COMMITTER_DATE="2026-10-28T00:00:00Z" \
    git -C "$CONTROL" commit-tree "$1" -p "$COMMIT" -m "$2"
}
fixture_branch() { # BRANCH PATH CONTENT_FILE: HEAD with PATH replaced (mode kept), pushed to origin
  local br=$1 path=$2 file=$3 idx="$S/index.$1" mode blob tree
  mode="$(git -C "$CONTROL" ls-tree "$COMMIT" -- "$path" | awk '{print $1}')"
  [[ -n "$mode" ]] || return 1
  GIT_INDEX_FILE="$idx" git -C "$CONTROL" read-tree "$COMMIT"
  blob="$(git -C "$CONTROL" hash-object -w -- "$file")"
  GIT_INDEX_FILE="$idx" git -C "$CONTROL" update-index --cacheinfo "$mode,$blob,$path"
  tree="$(GIT_INDEX_FILE="$idx" git -C "$CONTROL" write-tree)"
  BR_SHA[$br]="$(fixture_commit "$tree" "G072 EP-1 fixture: $br")"
  git -C "$CONTROL" push --quiet origin "${BR_SHA[$br]}:refs/heads/$br"
  rm -f "$idx"
}
in_s6() { local p; for p in "${S6_PATHS[@]}"; do [[ "$1" == "$p" || "$1" == "$p/"* ]] && return 0; done; return 1; }
case_fixture() {
  local why="" ok=1 r br want got n0 n1
  log "fixture: clone HEAD $COMMIT"
  git clone --quiet --no-tags --single-branch "$REPO" "$CONTROL"
  git -C "$CONTROL" checkout --quiet --detach "$COMMIT"
  while IFS= read -r r; do git -C "$CONTROL" update-ref -d "$r"; done < <(git -C "$CONTROL" for-each-ref --format='%(refname)' refs/heads refs/remotes)
  git -C "$CONTROL" remote remove origin
  git init --quiet --bare "$ORIGIN"
  # The origin borrows the clone's objects, so the pushes below transfer refs, not the history.
  printf '%s\n' "$CONTROL/.git/objects" >"$ORIGIN/objects/info/alternates"
  git -C "$CONTROL" remote add origin "$ORIGIN"
  git -C "$CONTROL" push --quiet origin "$COMMIT:refs/heads/staging"
  # Planted canaries: the runner's credential in .git/config and untracked project files (the GitLab project directory
  # also holds restored artifacts); the job home's Docker client configuration is written per job (job_home).
  git -C "$CONTROL" config http.extraHeader "JOB-TOKEN: ${CANARY[CI_JOB_TOKEN]}"
  printf '%s\n' "${CANARY[PLANTED_FILE]}" >"$CONTROL/.g072-boundary-planted"
  printf '%s\n' "${CANARY[PLANTED_FILE]}" >"$CONTROL/core/.g072-boundary-planted"
  # Branches.
  cp "$FIX/migrate.hostile.Dockerfile" "$S/hostile.Dockerfile"
  fixture_branch "$BR_DOCKERFILE" infra/docker/migrate.Dockerfile "$S/hostile.Dockerfile" || { ok=0; why+="infra/docker/migrate.Dockerfile not at HEAD; "; }
  awk -v frag="$FIX/migrate.mutant.fragment" -v leak="${CANARY[CI_JOB_TOKEN]}" '
    $0 == "USER nobody" && !done { while ((getline l < frag) > 0) { gsub(/@G072_MUTANT_LEAK@/, leak, l); print l }; done = 1 }
    {print}' "$S/hostile.Dockerfile" >"$S/mutant.Dockerfile"
  [[ "$(grep -cF -- "${CANARY[CI_JOB_TOKEN]}" "$S/mutant.Dockerfile")" == 1 ]] || { ok=0; why+="mutant fragment not inserted exactly once; "; }
  fixture_branch "$BR_MUTANT" infra/docker/migrate.Dockerfile "$S/mutant.Dockerfile" || { ok=0; why+="mutant branch; "; }
  git -C "$CONTROL" show "$COMMIT:infra/compose/local.yml" >"$S/local.yml.orig"
  awk -v frag="$FIX/local.yml.core-build.fragment" '
    {print}
    $0 == "      dockerfile: ../infra/docker/core.Dockerfile" && !done { while ((getline l < frag) > 0) print l; done = 1 }' \
    "$S/local.yml.orig" >"$S/local.yml.hostile"
  n0="$(grep -c 'additional_contexts:' "$S/local.yml.orig" || true)" n1="$(grep -c 'additional_contexts:' "$S/local.yml.hostile" || true)"
  { (( n1 == n0 + 1 )) && grep -q 'CI_JOB_TOKEN: \${CI_JOB_TOKEN:-}' "$S/local.yml.hostile"; } || { ok=0; why+="compose fragment not inserted after core's dockerfile line; "; }
  fixture_branch "$BR_COMPOSE" infra/compose/local.yml "$S/local.yml.hostile" || { ok=0; why+="compose branch; "; }
  for br in "$BR_DOCKERFILE" "$BR_MUTANT" "$BR_COMPOSE"; do
    [[ -n "${BR_SHA[$br]:-}" ]] || continue
    case "$br" in "$BR_COMPOSE") want=infra/compose/local.yml ;; *) want=infra/docker/migrate.Dockerfile ;; esac
    got="$(git -C "$CONTROL" diff --name-only "$COMMIT" "${BR_SHA[$br]}" | tr '\n' ' ')"
    [[ "$got" == "$want " ]] || { ok=0; why+="$br differs in [$got], want [$want]; "; }
    [[ "$(git -C "$ORIGIN" rev-parse "refs/heads/$br")" == "${BR_SHA[$br]}" ]] || { ok=0; why+="$br not in origin; "; }
  done
  in_s6 infra/compose/local.yml || { ok=0; why+="infra/compose/local.yml is not an S6 path; "; }
  ! in_s6 infra/docker/migrate.Dockerfile || { ok=0; why+="infra/docker/migrate.Dockerfile is an S6 path (the Dockerfile variant would be refused, not built); "; }
  if (( ok )); then result fixture PASS - "control ${COMMIT:0:12} cloned; origin: staging=HEAD, $BR_DOCKERFILE=${BR_SHA[$BR_DOCKERFILE]:0:12}, $BR_MUTANT=${BR_SHA[$BR_MUTANT]:0:12}, $BR_COMPOSE=${BR_SHA[$BR_COMPOSE]:0:12}; each differs in its one fixture path; compose path in S6, migrate.Dockerfile not; canaries planted"
  else result fixture FAIL - "$why"; return 1; fi
}

# ---- Docker helpers --------------------------------------------------------------------------------------------------
PIPELINE_ID=$(( 900000 + RANDOM )) JOB_SEQ=$(( 7000000 + RANDOM )) DOCKER_HOST_VALUE=""
job_home() { # DIR: the job's HOME and Docker client configuration (a planted canary credential; CLI plugins linked)
  local h=$1 plugins="${DOCKER_CONFIG:-$HOME/.docker}/cli-plugins"
  (umask 077; mkdir -p "$h/.docker"
    printf '{"auths":{"registry.invalid":{"auth":"%s"}}}\n' "$(printf 'gitlab-ci-token:%s' "${CANARY[PLANTED_FILE]}" | base64 -w0)" >"$h/.docker/config.json")
  [[ ! -d "$plugins" ]] || ln -s "$plugins" "$h/.docker/cli-plugins"
}
# adapter ACTION CASE_DIR ENV CANDIDATE TIMEOUT_MIN: runs control `bash infra/staging/ci.sh ACTION` in the control clone
# with exactly the job environment (everything else unset; values exported, never on a command line); stdout and
# stderr go to CASE_DIR/ACTION.{out,err}; returns its exit code.
adapter() {
  local action=$1 cd=$2 env=$3 cand=$4 tmo=$5 rc=0 v
  local -a e
  JOB_SEQ=$((JOB_SEQ + 1))
  (umask 077; mkdir -p "$cd/private-$action" "$cd/out")
  e=(PATH="$PATH" HOME="$cd/home" DOCKER_CONFIG="$cd/home/.docker" LANG=C.UTF-8 CI=true GITLAB_CI=true
    CI_PROJECT_DIR="$CONTROL" CI_COMMIT_SHA="$COMMIT" CI_COMMIT_BRANCH=staging CI_COMMIT_REF_NAME=staging
    CI_COMMIT_REF_PROTECTED=true CI_PIPELINE_SOURCE=web CI_PIPELINE_ID="$PIPELINE_ID" CI_JOB_ID="$JOB_SEQ"
    CI_JOB_NAME="staging:$action" GITLAB_USER_LOGIN=g072-boundary
    STAGING_CI_ACTION="$action" STAGING_CI_ENV="$env" STAGING_CI_CANDIDATE_INPUT="$cand" STAGING_CI_DURATION_INPUT=
    STAGING_CI_PRIVATE="$cd/private-$action" STAGING_CI_OUT="$cd/out"
    CI_JOB_TOKEN="${CANARY[CI_JOB_TOKEN]}" CI_REGISTRY_PASSWORD="${CANARY[CI_JOB_TOKEN]}"
    CI_REPOSITORY_URL="https://gitlab-ci-token:${CANARY[CI_JOB_TOKEN]}@gitlab.invalid/g072/boundary.git"
    STAGING_CI_ID_TOKEN="${CANARY[STAGING_CI_ID_TOKEN]}" CLOUDFLARE_API_TOKEN="${CANARY[CLOUDFLARE_API_TOKEN]}")
  [[ $action != build ]] || e+=(STAGING_REUSE=no)
  [[ -z "$DOCKER_HOST_VALUE" ]] || e+=(DOCKER_HOST="$DOCKER_HOST_VALUE")
  e+=("${EXTRA_ENV[@]}")
  log "$(basename "$cd"): ci.sh $action (STAGING_CI_ENV=$env, candidate '${cand}', timeout ${tmo}m)"
  set +e
  (
    cd "$CONTROL" || exit 65
    while IFS= read -r v; do unset "$v" 2>/dev/null || true; done < <(compgen -e)
    for v in "${e[@]}"; do export "${v?}"; done
    exec timeout --kill-after=30s "${tmo}m" bash "$CONTROL/infra/staging/ci.sh" "$action"
  ) >"$cd/$action.out" 2>"$cd/$action.err" </dev/null
  rc=$?
  set -e
  return "$rc"
}
images_snapshot() { docker image ls -a -q --no-trunc | sort -u; }
# find_manifest OUT: the G070-format build manifest (image rows: service, candidate, Dockerfile sha256, image ID).
find_manifest() {
  local f
  while IFS= read -r -d '' f; do
    awk -F'\t' '$1 == "image" && NF >= 6 && $5 ~ /^sha256:[0-9a-f]{64}$/ {n++} END {exit !(n >= 7)}' "$f" && { printf '%s\n' "$f"; return 0; }
  done < <(find "$1" -type f -name '*.tsv' -print0 2>/dev/null | sort -z)
  return 1
}
# manifest_check MANIFEST CANDIDATE: exactly the seven services, each once, with this candidate and an image ID.
manifest_check() {
  local got
  got="$(awk -F'\t' -v c="$2" '$1 == "image" && $3 == c && $5 ~ /^sha256:[0-9a-f]{64}$/ {print $2}' "$1" | sort | tr '\n' ' ')"
  [[ "$got" == "$(printf '%s\n' "${SERVICES[@]}" | sort | tr '\n' ' ')" ]]
}
manifest_id() { awk -F'\t' -v s="$2" '$1 == "image" && $2 == s {print $5; exit}' "$1"; }
# owned_refs MANIFEST CANDIDATE: tags naming a manifest image that this tooling creates (<registry>/<svc>:<candidate>
# and Compose's <env>-build-<svc>), as "ref<TAB>id"; appended to the cleanup record.
owned_refs() {
  local m=$1 c=$2 svc id ref
  declare -A want=()
  for svc in "${SERVICES[@]}"; do id="$(manifest_id "$m" "$svc")"; [[ -z "$id" ]] || want[$id]=$svc; done
  { docker image ls --no-trunc --format "{{.Repository}}:{{.Tag}}"$'\t'"{{.ID}}" || true; } | while IFS=$'\t' read -r ref id; do
    svc="${want[$id]:-}"
    [[ -n "$svc" && "$ref" != *"<none>"* ]] || continue
    if [[ "$ref" == */"$svc:$c" || "$ref" =~ ^[a-z][a-z0-9-]*-build-${svc}:latest$ ]]; then printf '%s\t%s\n' "$ref" "$id"; fi
  done | sort -u | tee -a "$REFS"
}
# scan_job CASE_DIR ACTION: adapter stdout/stderr, STAGING_CI_PRIVATE, the job home (planted file excluded) and new
# untracked project files (planted files excluded); returns the number of dirty files.
scan_job() {
  local cd=$1 a=$2 bad=0 tag f
  tag="$(basename "$cd") $a"
  scan_stream "$tag stdout" cat -- "$cd/$a.out" || bad=$((bad + 1))
  scan_stream "$tag stderr" cat -- "$cd/$a.err" || bad=$((bad + 1))
  scan_tree "$tag private" "$cd/private-$a" || bad=$((bad + $?))
  scan_tree "$tag home" "$cd/home" .docker/config.json || bad=$((bad + $?))
  while IFS= read -r -d '' f; do
    f="${f:3}"
    [[ "$f" == .g072-boundary-planted || "$f" == core/.g072-boundary-planted || -d "$CONTROL/$f" ]] && continue
    scan_stream "$tag project $f" cat -- "$CONTROL/$f" || bad=$((bad + 1))
  done < <(git -C "$CONTROL" status --porcelain=v1 -z --untracked-files=all --ignored)
  capped "$bad"
}
# scan_images CASE MANIFEST: docker export of a container of every manifest image, its inspect JSON and history; the
# migrate export's file list is kept for the positive controls. Returns the number of dirty streams.
scan_images() {
  local cs=$1 m=$2 bad=0 svc id cid
  for svc in "${SERVICES[@]}"; do
    id="$(manifest_id "$m" "$svc")"
    cid="$(docker create --label g072.task=G072 --label "g072.boundary.run=$RUN_ID" --network none "$id" /g072-boundary-noop)" ||
      { printf '%s export %s\t-\tunscannable (docker create failed)\t-\n' "$cs" "$svc" >>"$HITS_FILE"; bad=$((bad + 1)); continue; }
    printf 'container\t%s (%s %s)\t%s\t\tcreated\n' "$cid" "$cs" "$svc" "$(date -u +%FT%TZ)" >>"$RES"
    scan_stream "$cs export $svc" docker export "$cid" || bad=$((bad + 1))
    if [[ $svc == migrate ]]; then
      set +e; docker export "$cid" | tar -tf - >"$S/$cs.migrate.list" 2>/dev/null
      docker export "$cid" | tar -xOf - leak/build-args.txt >"$S/$cs.migrate.args" 2>/dev/null; set -e
    fi
    docker rm "$cid" >/dev/null 2>&1 || true
    if [[ -z "$(docker ps -aq --filter "id=$cid" 2>/dev/null || true)" ]]; then printf 'container\t%s\t\t%s\tremoved\n' "$cid" "$(date -u +%FT%TZ)" >>"$RES"
    else printf 'container\t%s\t\t\tLEFTOVER (swept at exit)\n' "$cid" >>"$RES"; fi
    scan_stream "$cs inspect $svc" docker image inspect "$id" || bad=$((bad + 1))
    scan_stream "$cs history $svc" docker history --no-trunc --format '{{.CreatedBy}}' "$id" || bad=$((bad + 1))
  done
  capped "$bad"
}
copy_logs() { # CASE_DIR ACTION...: adapter output of a clean case into the report
  local cd=$1 a f; shift
  for a in "$@"; do for f in out err; do [[ ! -f "$cd/$a.$f" ]] || cp "$cd/$a.$f" "$REPORT/logs/$(basename "$cd")-$a.$f"; done; done
}
stop_run() { STOPPED="$1"; log "STOP: $1 (Plan §10.12 / A5: stop and return for a reviewed Plan change)"; }

# build_case CASE ENV CANDIDATE_INPUT EXPECTED_SHA: prepare + build + manifest; sets CASE_DIR, MANIFEST, CASE_RC, WHY.
build_case() {
  local cs=$1 env=$2 cand=$3 sha=$4 prc=0 brc=0 t0
  CASE_DIR="$S/$cs" MANIFEST="" WHY="" CASE_RC=""
  mkdir -p "$CASE_DIR"; job_home "$CASE_DIR/home"
  t0=$SECONDS; adapter prepare "$CASE_DIR" "$env" "$cand" "$PREP_MIN" || prc=$?
  measure "$cs" prepare_seconds "$((SECONDS - t0))"
  CASE_RC="prepare=$prc"
  (( prc == 0 )) || { WHY="prepare exit $prc; "; return 1; }
  t0=$SECONDS; adapter build "$CASE_DIR" "$env" "$cand" "$BUILD_MIN" || brc=$?
  measure "$cs" build_seconds "$((SECONDS - t0))"
  CASE_RC+=" build=$brc"
  (( brc == 0 )) || { WHY="build exit $brc; "; return 1; }
  MANIFEST="$(find_manifest "$CASE_DIR/out")" || { WHY="no G070-format build manifest in STAGING_CI_OUT; "; return 1; }
  # Record this build's tags for cleanup by the candidate the manifest itself names, before judging it.
  owned_refs "$MANIFEST" "$(awk -F'\t' '$1 == "image" {print $3; exit}' "$MANIFEST")" >/dev/null || true
  manifest_check "$MANIFEST" "$sha" || { WHY="manifest ${MANIFEST#"$CASE_DIR"/} does not name exactly the seven services with candidate ${sha:0:12}; "; return 1; }
  return 0
}
# clean_scan CASE: every output of a clean case. Returns the number of dirty or unscannable streams/files.
clean_scan() {
  local cs=$1 bad=0
  scan_job "$CASE_DIR" prepare || bad=$((bad + $?))
  [[ ! -f "$CASE_DIR/build.out" ]] || { scan_job "$CASE_DIR" build || bad=$((bad + $?)); }
  scan_tree "$cs out" "$CASE_DIR/out" || bad=$((bad + $?))
  [[ -z "$MANIFEST" ]] || { scan_images "$cs" "$MANIFEST" || bad=$((bad + $?)); }
  capped "$bad"
}

# ---- case: compose-refused -------------------------------------------------------------------------------------------
case_compose_refused() {
  local cd="$S/compose-refused" prc=0 brc=0 why="" ok=1 before after bad=0
  mkdir -p "$cd"; job_home "$cd/home"
  before="$(images_snapshot)"
  adapter prepare "$cd" preview-1 "$BR_COMPOSE" "$PREP_MIN" || prc=$?
  (( prc == 2 )) || { ok=0; why+="prepare exit $prc, want 2; "; }
  { cat "$cd/prepare.out" "$cd/prepare.err"; find "$cd/out" -type f -exec cat {} + 2>/dev/null; } >"$S/compose-refusal.txt" || true
  grep -q 'incompatible candidate' "$S/compose-refusal.txt" && grep -q 'infra/compose' "$S/compose-refusal.txt" ||
    { ok=0; why+="no 'incompatible candidate ... infra/compose' refusal in prepare output or STAGING_CI_OUT; "; }
  if find_manifest "$cd/out" >/dev/null || [[ -n "$(find "$cd/out" -type f \( -name '*.gz' -o -name '*.tgz' -o -name '*.tar' \) 2>/dev/null)" ]]; then
    ok=0; why+="a build manifest or image archive exists after the refusal; "
  fi
  adapter build "$cd" preview-1 "$BR_COMPOSE" "$BUILD_MIN" || brc=$?
  (( brc == 2 )) || { ok=0; why+="build on prepare's leftovers exit $brc, want 2; "; }
  after="$(images_snapshot)"
  [[ "$before" == "$after" ]] || { ok=0; why+="$(comm -13 <(printf '%s\n' "$before") <(printf '%s\n' "$after") | grep -c .) image(s) created; "; }
  CASE_DIR="$cd" MANIFEST=""
  clean_scan compose-refused || bad=$?
  (( bad == 0 )) || { ok=0; why+="$bad dirty or unscannable outputs (hits.tsv); "; stop_run "canary in compose-refused outputs"; }
  (( bad > 0 )) || copy_logs "$cd" prepare build
  if (( ok )); then result compose-refused PASS "prepare=$prc build=$brc" "refused by prepare (exit 2, incompatible infra/compose) before any build: no image created (image set unchanged), no manifest or archive; build on the leftovers exit 2; no canary in any output"
  else result compose-refused FAIL "prepare=$prc build=$brc" "$why"; fi
}

# ---- case: baseline-measure ------------------------------------------------------------------------------------------
case_baseline_measure() {
  local cs=baseline-measure ok=1 why="" bad=0 total=0 sz f rel svc id refs ref t0 lrc=0 have tag indep=0
  build_case "$cs" staging "" "$COMMIT" || { ok=0; why+="$WHY"; }
  clean_scan "$cs" || bad=$?
  (( bad == 0 )) || { ok=0; why+="$bad dirty or unscannable outputs (hits.tsv); "; stop_run "canary in $cs outputs"; }
  if [[ -n "$MANIFEST" ]]; then
    while IFS= read -r -d '' f; do
      rel="${f#"$CASE_DIR/out/"}"; sz="$(stat -c %s -- "$f")"; total=$((total + sz))
      measure "$cs" "artifact_bytes:$rel" "$sz"
    done < <(find "$CASE_DIR/out" -type f -print0 | sort -z)
    measure "$cs" artifact_bytes_total "$total"
    measure "$cs" artifact_bound_bytes "$((MAX_MB * 1048576))"
    for svc in "${SERVICES[@]}"; do measure "$cs" "image_size_bytes:$svc" "$(docker image inspect --format '{{.Size}}' "$(manifest_id "$MANIFEST" "$svc")")"; done
    if (( total > MAX_MB * 1048576 )); then ok=0; why+="artifacts $total bytes > bound ${MAX_MB} MiB; "; stop_run "artifact size $total bytes above ${MAX_MB} MiB"; fi
    # Image-ID preservation: remove this tooling's tags, load the archives, then the IDs, the restored tags and an
    # ID-preserving retag must all be the manifest's.
    refs="$(owned_refs "$MANIFEST" "$COMMIT" || true)"
    while IFS=$'\t' read -r ref id; do [[ -z "$ref" ]] || docker image rm "$ref" >/dev/null 2>&1 || true; done <<<"$refs"
    for svc in "${SERVICES[@]}"; do
      id="$(manifest_id "$MANIFEST" "$svc")"
      if docker image inspect "$id" >/dev/null 2>&1; then measure "$cs" "id_removed_before_load:$svc" "no (other references)"
      else measure "$cs" "id_removed_before_load:$svc" yes; indep=$((indep + 1)); fi
    done
    t0=$SECONDS
    while IFS= read -r -d '' f; do
      docker load -q -i "$f" >>"$S/$cs.load" 2>&1 || lrc=$?
    done < <(find "$CASE_DIR/out" -type f \( -name '*.tar.gz' -o -name '*.tgz' -o -name '*.tar' \) -print0 | sort -z)
    measure "$cs" load_seconds "$((SECONDS - t0))"
    (( lrc == 0 )) || { ok=0; why+="docker load exit $lrc; "; }
    [[ -s "$S/$cs.load" ]] || { ok=0; why+="no image archive (*.tar.gz, *.tgz, *.tar) in STAGING_CI_OUT to load; "; }
    owned_refs "$MANIFEST" "$COMMIT" >/dev/null || true
    for svc in "${SERVICES[@]}"; do
      id="$(manifest_id "$MANIFEST" "$svc")"
      have="$(docker image inspect --format '{{.Id}}' "$id" 2>/dev/null || true)"
      [[ "$have" == "$id" ]] || { ok=0; why+="$svc: manifest ID absent after load; "; measure "$cs" "id_after_load:$svc" absent; continue; }
      measure "$cs" "id_after_load:$svc" match
      tag="g072-boundary-retag.invalid/$svc:${COMMIT:0:12}-${RUN_ID##*-}"
      printf '%s\t%s\n' "$tag" "$id" >>"$REFS"
      docker tag "$id" "$tag" || { ok=0; why+="$svc: docker tag failed; "; continue; }
      have="$(docker image inspect --format '{{.Id}}' "$tag" 2>/dev/null || true)"
      if [[ "$have" == "$id" ]]; then measure "$cs" "id_after_retag:$svc" match
      else ok=0; why+="$svc: retag resolves to another ID; "; measure "$cs" "id_after_retag:$svc" mismatch; fi
    done
    while IFS=$'\t' read -r ref id; do
      [[ -z "$ref" ]] && continue
      have="$(docker image inspect --format '{{.Id}}' "$ref" 2>/dev/null || true)"
      [[ -z "$have" || "$have" == "$id" ]] || { ok=0; why+="restored tag $ref names another image; "; }
    done <<<"$refs"
    measure "$cs" id_preservation_independent "$indep of ${#SERVICES[@]} IDs absent before load"
  fi
  (( bad > 0 )) || copy_logs "$CASE_DIR" prepare build
  if (( ok )); then result "$cs" PASS "$CASE_RC" "manifest of ${COMMIT:0:12} with seven services; no canary; artifacts $total bytes <= ${MAX_MB} MiB; IDs preserved across save/load/retag ($indep of 7 removed first); measure.tsv"
  else result "$cs" FAIL "$CASE_RC" "$why"; fi
}

# ---- case: dockerfile-clean ------------------------------------------------------------------------------------------
case_dockerfile_clean() {
  local cs=dockerfile-clean ok=1 why="" bad=0 want got l logs=no
  build_case "$cs" preview-1 "$BR_DOCKERFILE" "${BR_SHA[$BR_DOCKERFILE]}" || { ok=0; why+="$WHY"; }
  clean_scan "$cs" || bad=$?
  (( bad == 0 )) || { ok=0; why+="$bad dirty or unscannable outputs (hits.tsv); "; stop_run "canary in $cs outputs"; }
  if [[ -n "$MANIFEST" ]]; then
    want="$(sha256sum <"$FIX/migrate.hostile.Dockerfile" | cut -d' ' -f1)"
    got="$(awk -F'\t' '$1 == "image" && $2 == "migrate" {print $4; exit}' "$MANIFEST")"
    [[ "$got" == "$want" ]] || { ok=0; why+="manifest migrate Dockerfile hash is not the hostile Dockerfile's; "; }
    for l in leak/env.txt leak/proc-environ.txt leak/mounts.txt leak/build-args.txt leak/context-files.txt; do
      grep -qxE "(\./)?$l" "$S/$cs.migrate.list" 2>/dev/null || { ok=0; why+="hostile dump $l missing from the migrate image (the variant did not run); "; }
    done
    [[ "$(cat "$S/$cs.migrate.args" 2>/dev/null)" == "job=unset id=unset cf=unset repo=unset" ]] ||
      { ok=0; why+="the candidate build saw a build argument (leak/build-args.txt is not all unset); "; }
    cat "$CASE_DIR"/build.out "$CASE_DIR"/build.err 2>/dev/null | grep -q '/leak/env.txt' && logs=yes
    [[ $logs == yes ]] || find "$CASE_DIR/private-build" "$CASE_DIR/out" -type f -exec grep -lq '/leak/env.txt' {} + 2>/dev/null && logs=yes
    measure "$cs" build_log_shows_hostile_steps "$logs"
  fi
  (( bad > 0 )) || copy_logs "$CASE_DIR" prepare build
  if (( ok )); then result "$cs" PASS "$CASE_RC" "hostile migrate.Dockerfile built (manifest hash matches); dumps present, no build argument; no canary (raw or base64) in 7 image exports, their inspect/history, adapter output, STAGING_CI_OUT (gunzipped), STAGING_CI_PRIVATE, job home or new project files; build log carries the hostile steps: $logs"
  else result "$cs" FAIL "$CASE_RC" "$why"; fi
}

# ---- case: mutant-detected -------------------------------------------------------------------------------------------
case_mutant_detected() {
  local cs=mutant-detected ok=1 why="" save=$HITS_FILE ex=0 hi=0
  build_case "$cs" preview-1 "$BR_MUTANT" "${BR_SHA[$BR_MUTANT]}" || { ok=0; why+="$WHY"; }
  # The adapter's own publication scan (exit 6, outputs withheld) catching the leak first is also a detection.
  if [[ "$CASE_RC" == *"build=6" ]]; then
    result "$cs" PASS "$CASE_RC" "intentional leak refused by the adapter's own canary scan (exit 6, outputs withheld); image scan not reached"
    return 0
  fi
  if [[ -n "$MANIFEST" ]]; then
    HITS_FILE="$REPORT/mutant-hits.tsv"; printf 'label\tcanary\tencoding\tcount\n' >"$HITS_FILE"
    scan_images "$cs" "$MANIFEST" || true
    scan_job "$CASE_DIR" build || true
    HITS_FILE=$save
    awk -F'\t' '$1 == "mutant-detected export migrate" && $2 == "CI_JOB_TOKEN" {f = 1} END {exit !f}' "$REPORT/mutant-hits.tsv" && ex=1
    awk -F'\t' '$1 == "mutant-detected history migrate" && $2 == "CI_JOB_TOKEN" {f = 1} END {exit !f}' "$REPORT/mutant-hits.tsv" && hi=1
    (( ex )) || { ok=0; why+="the baked-in canary was not reported in the migrate export; "; }
    (( hi )) || { ok=0; why+="the baked-in canary was not reported in the migrate history; "; }
    measure "$cs" build_log_hit "$(grep -c '^mutant-detected build std' "$REPORT/mutant-hits.tsv" || true)"
  fi
  if (( ok )); then result "$cs" PASS "$CASE_RC" "intentional leak reported: migrate export and history (mutant-hits.tsv; output not copied)"
  else result "$cs" FAIL "$CASE_RC" "$why"; fi
}

# ---- main ------------------------------------------------------------------------------------------------------------
if (( NEED_DOCKER )); then
  for t in docker timeout gzip base64 tar; do command -v "$t" >/dev/null || die "$t is required"; done
  docker info >/dev/null 2>&1 || die "no Docker daemon"
  DOCKER_HOST_VALUE="${DOCKER_HOST:-$(docker context inspect --format '{{.Endpoints.docker.Host}}' 2>/dev/null || true)}"
fi
log "run $RUN_ID at $COMMIT_NOTE: ${SEL[*]}"
FIXTURE_OK=1
for c in "${SEL[@]}"; do
  case "$c" in
    fixture) case_fixture || FIXTURE_OK=0 ;;
    scan-controls) case_scan_controls ;;
    *)
      if (( ! FIXTURE_OK )); then result "$c" SKIP - "fixture failed"; continue; fi
      if [[ -n "$STOPPED" ]]; then result "$c" SKIP - "stopped: $STOPPED"; continue; fi
      case "$c" in
        compose-refused) case_compose_refused ;;
        baseline-measure) case_baseline_measure ;;
        dockerfile-clean) case_dockerfile_clean ;;
        mutant-detected) case_mutant_detected ;;
      esac ;;
  esac
done
[[ -z "$STOPPED" ]] || printf 'stop\tSTOP\t-\t%s\t%s (Plan §10.12 / A5: return for a reviewed Plan change)\n' "$COMMIT_NOTE" "$STOPPED" >>"$RESULTS"
log "results: $RESULTS"
(( FAILED == 0 && ${#STOPPED} == 0 )) || exit 1
exit 0
