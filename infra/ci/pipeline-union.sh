#!/usr/bin/env bash
# Pipeline union evidence check (G051 C2; Plan §3.7, §3.8.3-§3.8.4; docs/foundation/OPERATIONS.md). Run outside CI
# jobs on published ci-evidence/ directories, which are read as data only. Deterministic: summary.json through the
# pinned yq image (no network), the TSV files by fixed expressions; never host python or jq.
#
#   bash infra/ci/pipeline-union.sh FAST_DIR GATE_DIR   the fast and gate partitions of one integration or full pipeline
#   bash infra/ci/pipeline-union.sh --single JOB_DIR    one complete single-job pipeline (a fast or affected fast job)
#   bash infra/ci/pipeline-union.sh --ledger JOB_DIR    AC5 recomputed from builds.tsv, ledger-summary.tsv and
#                                                       ledger-bootstrap.tsv (infra/validation/lib.sh ledger_check_file)
#
# Union rules (any one fails): a partition is missing or not successful (a minimal candidate, or a recorded phase
# that is not 0); the candidate commit or tree differs; selection_digest, pipeline_suites or pipeline_set_sha256
# differs; a job's executed suites differ from its expected partition (fast: the fast suites; gate: the pipeline
# suites minus the fast suites); the union of the PASS rows is not E_pipeline (a gap or an extra ID); the jobs
# overlap in anything but exactly the final IDs; a required ID is not PASS exactly once; the tested content, CI project,
# pipeline or commit differs; for an RC promotion (mr-rc-promotion, always the full profile) the release identity is
# missing, malformed or differs (tag, tag object, peeled commit = candidate, merge request, release/<tag> source and
# target), while each job's observed staging tip is validated per job only (staging may advance between the jobs).
# The ledger mode needs ac5_status MET as recomputed by ledger_check_file and equal to the published summary, and
# a bootstrap verdict of structural-inference:reuse for a CI job (not-applicable for a local run). That verdict is
# a structural inference from equal context image IDs, never step, cache or build-count evidence.
#
# Output: exactly one line, "pipeline-union: PASS ..." or "pipeline-union: FAIL ..." (usage: "pipeline-union: usage
# ..."). Exit: 0 pass; 1 fail; 2 usage. Every check returns explicitly; no statement is negated with `!`.
set -euo pipefail
if (( BASH_VERSINFO[0] < 4 || (BASH_VERSINFO[0] == 4 && BASH_VERSINFO[1] < 4) )); then
  echo "pipeline-union: FAIL bash >= 4.4 is required"
  exit 1
fi
shopt -s inherit_errexit
export LC_ALL=C

SELF_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)"
REPO_R="$(dirname -- "$(dirname -- "$SELF_DIR")")"

result() { printf 'pipeline-union: %s\n' "$*"; }
fail() { result "FAIL $*"; exit 1; }
usage() { result "usage: pipeline-union.sh FAST_DIR GATE_DIR | --single JOB_DIR | --ledger JOB_DIR"; exit 2; }

MODE="" A="" B=""
case "${1:-}" in
  --single|--ledger)
    if (( $# != 2 )) || [[ -z "$2" || "$2" == -* ]]; then usage; fi
    MODE="${1#--}" A="$2" ;;
  ""|-*) usage ;;
  *)
    if (( $# != 2 )) || [[ -z "$2" || "$2" == -* ]]; then usage; fi
    MODE=union A="$1" B="$2" ;;
esac

if [[ -L "$SELF_DIR/lib.sh" || ! -f "$SELF_DIR/lib.sh" ]]; then fail "infra/ci/lib.sh is missing"; fi
# shellcheck source=lib.sh
source "$SELF_DIR/lib.sh"
if [[ -z "${CI_SUITE_ORDER:-}" || -z "${CI_SUITES_fast:-}" || -z "${CI_YQ_IMAGE:-}" ]]; then fail "infra/ci/lib.sh lacks the G051 suite lists"; fi

SCR=""
cleanup() { if [[ -n "$SCR" ]]; then rm -rf -- "$SCR"; fi; }
trap cleanup EXIT

WHY=""
declare -A RSUITE=() J=()
RQ=$'^([A-Z][A-Z0-9-]*)\t(lint|lint-base|lint-web|lint-core|controls|ci-controls|provenance|contracts|unit|unit-web|unit-core|unit-intel|root|migrations|ops|rename|off|feature-web|feature-core|staging|final)$'
RE_SUITES='^[a-z][a-z-]*( [a-z][a-z-]*)*$'
RE_AREAS='^(none|(web|core|intel):[a-z][a-z0-9-]{0,31}(,(web|core|intel):[a-z][a-z0-9-]{0,31})*)$'

# --- sets ------------------------------------------------------------------------------------------------------
# load_registry FILE: RSUITE[ID]=suite from a published required-checks.tsv (fixed shape, each ID once).
load_registry() {
  local line id
  RSUITE=()
  if [[ -L "$1" || ! -f "$1" ]]; then return 1; fi
  while IFS= read -r line || [[ -n "$line" ]]; do
    if [[ -z "$line" || "$line" == '#'* ]]; then continue; fi
    if [[ ! "$line" =~ $RQ ]]; then return 1; fi
    id="${BASH_REMATCH[1]}"
    if [[ -n "${RSUITE[$id]+x}" ]]; then return 1; fi
    RSUITE[$id]="${BASH_REMATCH[2]}"
  done <"$1"
  if (( ${#RSUITE[@]} == 0 )); then return 1; fi
  return 0
}

# ids_of SUITES: the sorted registry IDs of SUITES plus final, one per line.
ids_of() {
  local id set=" $1 final "
  for id in "${!RSUITE[@]}"; do
    if [[ "$set" == *" ${RSUITE[$id]} "* ]]; then printf '%s\n' "$id"; fi
  done | sort
}

# set_sha LIST: SHA-256 of the newline-joined sorted list, as validate.sh computes required_set_sha256.
set_sha() { printf '%s\n' "$1" | sha256sum | cut -c1-64; }
count_of() { if [[ -z "$1" ]]; then printf '0\n'; else printf '%s\n' "$1" | grep -c .; fi; }

# set_minus A B / set_and A B (newline lists): the sorted lines of A not in B / also in B.
set_pick() {
  local x
  local -A inb=()
  while IFS= read -r x; do if [[ -n "$x" ]]; then inb[$x]=1; fi; done <<<"$3"
  while IFS= read -r x; do
    if [[ -z "$x" ]]; then continue; fi
    if [[ "$1" == minus && -z "${inb[$x]+x}" ]] || [[ "$1" == and && -n "${inb[$x]+x}" ]]; then printf '%s\n' "$x"; fi
  done <<<"$2" | sort
}
set_minus() { set_pick minus "$1" "$2"; }
set_and() { set_pick and "$1" "$2"; }
first_of() { printf '%s\n' "${1%%$'\n'*}"; }

# suite_list_ok LIST: known run.sh suites (CI_SUITE_ORDER), each once, in canonical order.
suite_list_ok() {
  local s i last=-1
  local -a order=() list=()
  local -A pos=()
  if [[ ! "$1" =~ $RE_SUITES ]]; then return 1; fi
  read -r -a order <<<"$CI_SUITE_ORDER"
  for i in "${!order[@]}"; do pos[${order[$i]}]="$i"; done
  read -r -a list <<<"$1"
  for s in "${list[@]}"; do
    if [[ -z "${pos[$s]+x}" ]]; then return 1; fi
    if (( pos[$s] <= last )); then return 1; fi
    last="${pos[$s]}"
  done
  return 0
}

# suites_minus LIST MINUS: LIST without the suites of MINUS, order kept.
suites_minus() {
  local s out="" sep=""
  local -a list=()
  read -r -a list <<<"$1"
  for s in "${list[@]}"; do
    if [[ " $2 " == *" $s "* ]]; then continue; fi
    out+="$sep$s"; sep=" "
  done
  printf '%s\n' "$out"
}

# --- published job evidence ------------------------------------------------------------------------------------
yq_vec() { docker run --rm --network none -i "$CI_YQ_IMAGE" -p json -o tsv "$2" <"$1" 2>/dev/null; }

SFIELDS=(format status_kind candidate marker tested_content pipeline_profile executed_profile head_sha tree_sha
  selection_digest pipeline_suites pipeline_set_sha256 executed_suites coverage partition areas required_count
  required_set_sha256)
declare -A SRE=(
  [format]='^platform-validation-summary/1$' [status_kind]='^pre-publication$' [candidate]='^full$' [marker]='^null$'
  [tested_content]='^(local|branch-push|branch-schedule|branch-web|branch-api|mr-source|mr-merged-result|mr-rc-promotion)$'
  [pipeline_profile]='^(fast|affected|integration|full|staging)$' [executed_profile]='^(fast|affected|integration|full|staging)$'
  [head_sha]='^[0-9a-f]{40}$' [tree_sha]='^[0-9a-f]{40}$' [selection_digest]='^[0-9a-f]{64}$'
  [pipeline_suites]="$RE_SUITES" [pipeline_set_sha256]='^[0-9a-f]{64}$' [executed_suites]="$RE_SUITES"
  [coverage]='^(complete|partial)$' [partition]='^(fast|remainder|complete)$' [areas]="$RE_AREAS"
  [required_count]='^[1-9][0-9]{0,5}$' [required_set_sha256]='^[0-9a-f]{64}$'
)

# load_summary TAG DIR: J[TAG.field] from a full pre-publication summary.json whose every field is typed and
# matches its expression (a minimal, withheld or nothing-executed candidate never qualifies).
load_summary() {
  local tag="$1" f="$2/summary.json" expr="" sep="" out k i want
  local -a v=()
  if [[ -L "$f" || ! -f "$f" ]]; then WHY="summary.json is missing"; return 1; fi
  for k in "${SFIELDS[@]}"; do expr+="$sep(.$k | tag), (.$k | to_string)"; sep=", "; done
  out="$(yq_vec "$f" "[$expr]")" || { WHY="summary.json is not JSON"; return 1; }
  IFS=$'\t' read -r -a v <<<"$out"
  if (( ${#v[@]} != 2 * ${#SFIELDS[@]} )); then WHY="summary.json is malformed"; return 1; fi
  for i in "${!SFIELDS[@]}"; do
    k="${SFIELDS[$i]}" want='!!str'
    if [[ "$k" == marker ]]; then want='!!null'; fi
    if [[ "$k" == required_count ]]; then want='!!int'; fi
    if [[ "${v[2 * i]}" != "$want" || ! "${v[2 * i + 1]}" =~ ${SRE[$k]} ]]; then WHY="summary.json $k is missing or malformed"; return 1; fi
    J[$tag.$k]="${v[2 * i + 1]}"
  done
  return 0
}

# load_ident TAG DIR (ONEOFF RC qualification): J[TAG.x.*] for the CI identity (project, pipeline, commit) and the RC
# promotion identity, each typed: an integer or null, a token or null. release is null, or an object of exactly the
# validate.sh fields; an RC promotion needs it, any other tested content needs null.
XFIELDS=(ci.project_id ci.pipeline_id ci.commit_sha release.tag release.tag_object release.peeled release.staging_sha
  release.merge_request_iid release.source_branch release.target_branch)
declare -A XRE=(
  [ci.project_id]='^[1-9][0-9]{0,18}$' [ci.pipeline_id]='^[1-9][0-9]{0,18}$' [ci.commit_sha]='^[0-9a-f]{40}$'
  [release.tag]='^(0|[1-9][0-9]*)[.](0|[1-9][0-9]*)[.](0|[1-9][0-9]*)-rc[.]([1-9][0-9]*)$'
  [release.tag_object]='^[0-9a-f]{40}$' [release.peeled]='^[0-9a-f]{40}$' [release.staging_sha]='^[0-9a-f]{40}$'
  [release.merge_request_iid]='^[1-9][0-9]{0,18}$' [release.source_branch]='^release/[0-9a-z.-]{1,64}$'
  [release.target_branch]='^[A-Za-z0-9][A-Za-z0-9._/-]{0,254}$'
)
load_ident() {
  local tag="$1" f="$2/summary.json" expr='(.release | tag), ((.release // {"none": 1}) | keys | sort | join(","))' out k i want
  local -a v=()
  for k in "${XFIELDS[@]}"; do expr+=", (.$k | tag), (.$k | to_string)"; done
  out="$(yq_vec "$f" "[$expr]")" || { WHY="summary.json is not JSON"; return 1; }
  IFS=$'\t' read -r -a v <<<"$out"
  if (( ${#v[@]} != 2 + 2 * ${#XFIELDS[@]} )); then WHY="summary.json identity fields are malformed"; return 1; fi
  J[$tag.x.release]="${v[0]}"
  if [[ "${v[0]}" == '!!map' ]]; then
    if [[ "${v[1]}" != merge_request_iid,peeled,source_branch,staging_sha,tag,tag_object,target_branch ]]; then
      WHY="summary.json release has other fields"; return 1
    fi
  elif [[ "${v[0]}" != '!!null' ]]; then WHY="summary.json release is neither an object nor null"; return 1; fi
  for i in "${!XFIELDS[@]}"; do
    k="${XFIELDS[$i]}" want='!!str'
    if [[ "$k" =~ (_id|_iid)$ ]]; then want='!!int'; fi
    if [[ "${v[2 + 2 * i]}" == '!!null' ]]; then J[$tag.x.$k]=null; continue; fi
    if [[ "${v[2 + 2 * i]}" != "$want" || ! "${v[3 + 2 * i]}" =~ ${XRE[$k]} ]]; then WHY="summary.json $k is malformed"; return 1; fi
    J[$tag.x.$k]="${v[3 + 2 * i]}"
  done
  if [[ "${J[$tag.tested_content]}" == mr-rc-promotion ]]; then
    if [[ "${J[$tag.x.release]}" != '!!map' ]]; then WHY="an RC promotion without its release identity"; return 1; fi
    for k in "${XFIELDS[@]}"; do if [[ "${J[$tag.x.$k]}" == null ]]; then WHY="an RC promotion without $k"; return 1; fi; done
    if [[ "${J[$tag.x.release.peeled]}" != "${J[$tag.head_sha]}" || "${J[$tag.x.ci.commit_sha]}" != "${J[$tag.head_sha]}" ]]; then
      WHY="the RC peeled commit is not the candidate commit"; return 1
    fi
    if [[ "${J[$tag.x.release.source_branch]}" != "release/${J[$tag.x.release.tag]}" ]]; then WHY="the RC source branch is not release/<tag>"; return 1; fi
  elif [[ "${J[$tag.x.release]}" != '!!null' ]]; then WHY="a release identity outside an RC promotion"; return 1; fi
  return 0
}

# load_phases DIR: phase-status.tsv of a successful full candidate: every recorded phase 0, the rest pending.
load_phases() {
  local f="$1/phase-status.tsv" line k re=$'^([a-z_]{1,32})\t([a-z0-9/_-]{1,16})$'
  local -A seen=()
  if [[ -L "$f" || ! -f "$f" ]]; then WHY="phase-status.tsv is missing"; return 1; fi
  while IFS= read -r line || [[ -n "$line" ]]; do
    if [[ ! "$line" =~ $re ]]; then WHY="phase-status.tsv is malformed"; return 1; fi
    k="${BASH_REMATCH[1]}"
    if [[ -n "${seen[$k]+x}" ]]; then WHY="phase-status.tsv repeats $k"; return 1; fi
    seen[$k]="${BASH_REMATCH[2]}"
  done <"$f"
  if [[ "${seen[status_kind]-}" != pre-publication || "${seen[candidate]-}" != full ]]; then WHY="the candidate is not a full pre-publication candidate"; return 1; fi
  for k in "${!seen[@]}"; do
    case "$k" in
      status_kind|candidate) ;;
      location|preflight|state|harness|signal|postcheck|cleanup|raw_scan|complete|stage|internal)
        if [[ "${seen[$k]}" != 0 ]]; then WHY="phase $k is ${seen[$k]}"; return 1; fi ;;
      summary|junit|candidate_scan|publish|report|scrub)
        if [[ "${seen[$k]}" != pending ]]; then WHY="phase $k is ${seen[$k]}"; return 1; fi ;;
      *) WHY="phase-status.tsv has an unknown phase $k"; return 1 ;;
    esac
  done
  for k in location preflight state harness postcheck cleanup raw_scan complete stage summary junit candidate_scan publish report scrub; do
    if [[ -z "${seen[$k]+x}" ]]; then WHY="phase $k is not recorded"; return 1; fi
  done
  return 0
}

# load_results TAG DIR: J[TAG.rows], the sorted registry IDs with a row; every line well formed, no FAIL or UNMET
# row, each registry ID at most once and PASS (INFO rows of other IDs are evidence only).
load_results() {
  local tag="$1" f="$2/results.tsv" line id st re=$'^([A-Z][A-Z0-9-]*)\t(PASS|FAIL|UNMET|INFO)\t[^\t]*$'
  local -A cnt=() ok=()
  if [[ -L "$f" || ! -f "$f" || ! -s "$f" ]]; then WHY="results.tsv is missing"; return 1; fi
  while IFS= read -r line || [[ -n "$line" ]]; do
    if [[ ! "$line" =~ $re ]]; then WHY="results.tsv has a malformed line"; return 1; fi
    id="${BASH_REMATCH[1]}" st="${BASH_REMATCH[2]}"
    if [[ "$st" == FAIL || "$st" == UNMET ]]; then WHY="$id is $st"; return 1; fi
    if [[ -z "${RSUITE[$id]+x}" ]]; then continue; fi
    cnt[$id]=$(( ${cnt[$id]:-0} + 1 ))
    if [[ "$st" == PASS ]]; then ok[$id]=1; fi
  done <"$f"
  for id in "${!cnt[@]}"; do
    if (( cnt[$id] != 1 )); then WHY="required ID $id occurs ${cnt[$id]} times"; return 1; fi
    if [[ -z "${ok[$id]+x}" ]]; then WHY="required ID $id is not PASS"; return 1; fi
  done
  J[$tag.rows]="$(for id in "${!cnt[@]}"; do printf '%s\n' "$id"; done | sort)"
  return 0
}

# job_set TAG: E_job of TAG (its executed suites plus final) must be what its summary declares.
job_set() {
  local tag="$1" set
  set="$(ids_of "${J[$tag.executed_suites]}")"
  if [[ "$(set_sha "$set")" != "${J[$tag.required_set_sha256]}" || "$(count_of "$set")" != "${J[$tag.required_count]}" ]]; then
    WHY="required_count/required_set_sha256 are not the registry set of its executed suites"; return 1
  fi
  J[$tag.eset]="$set"
  return 0
}

# rows_exact TAG: the job's rows are exactly its E_job.
rows_exact() {
  local tag="$1" miss extra
  miss="$(set_minus "${J[$tag.eset]}" "${J[$tag.rows]}")"
  extra="$(set_minus "${J[$tag.rows]}" "${J[$tag.eset]}")"
  if [[ -n "$miss" ]]; then WHY="required ID $(first_of "$miss") has no row ($(count_of "$miss") missing)"; return 1; fi
  if [[ -n "$extra" ]]; then WHY="a row for $(first_of "$extra") lies outside its partition ($(count_of "$extra") extra)"; return 1; fi
  return 0
}

job_dir() { # TAG DIR
  if [[ -L "$2" || ! -d "$2" ]]; then fail "the $1 partition is missing ($2 is not a directory)"; fi
}
job_load() { # TAG DIR: summary and phases
  if load_summary "$1" "$2"; then :; else fail "$1: $WHY"; fi
  if load_ident "$1" "$2"; then :; else fail "$1: $WHY"; fi
  if load_phases "$2"; then :; else fail "$1: not successful: $WHY"; fi
  if suite_list_ok "${J[$1.pipeline_suites]}" && suite_list_ok "${J[$1.executed_suites]}"; then :; else
    fail "$1: pipeline or executed suites are not known suites in canonical order"
  fi
}

# --- modes -----------------------------------------------------------------------------------------------------
union_main() {
  local k rem pset union gap extra overlap finals nf ng nu
  job_dir fast "$A"; job_dir gate "$B"
  if load_registry "$A/required-checks.tsv"; then :; else fail "fast: required-checks.tsv is missing or malformed"; fi
  if [[ -L "$B/required-checks.tsv" || ! -f "$B/required-checks.tsv" ]]; then fail "gate: required-checks.tsv is missing"; fi
  if cmp -s -- "$A/required-checks.tsv" "$B/required-checks.tsv"; then :; else fail "the partitions carry different registries"; fi
  job_load fast "$A"; job_load gate "$B"
  if [[ "${J[fast.head_sha]}" != "${J[gate.head_sha]}" ]]; then fail "the candidate commit differs"; fi
  if [[ "${J[fast.tree_sha]}" != "${J[gate.tree_sha]}" ]]; then fail "the candidate tree differs"; fi
  for k in selection_digest pipeline_profile pipeline_suites pipeline_set_sha256 areas tested_content; do
    if [[ "${J[fast.$k]}" != "${J[gate.$k]}" ]]; then fail "$k differs between the partitions"; fi
  done
  # ONEOFF RC qualification: one project, pipeline and commit, and one RC identity (tag, tag object, peeled commit,
  # merge request and branches). Each job's observed staging tip is its own (staging may advance between the jobs);
  # it is checked per job, never required to be equal here.
  for k in ci.project_id ci.pipeline_id ci.commit_sha release.tag release.tag_object release.peeled release.merge_request_iid \
    release.source_branch release.target_branch; do
    if [[ "${J[fast.x.$k]}" != "${J[gate.x.$k]}" ]]; then fail "$k differs between the partitions"; fi
  done
  if [[ "${J[fast.tested_content]}" == mr-rc-promotion && "${J[fast.pipeline_profile]}" != full ]]; then
    fail "an RC promotion pipeline is not the full profile"
  fi
  if [[ ! "${J[fast.pipeline_profile]}" =~ ^(integration|full)$ ]]; then fail "the pipeline profile ${J[fast.pipeline_profile]} has no gate partition"; fi
  if [[ "${J[fast.partition]}" != fast || "${J[fast.executed_profile]}" != fast || "${J[fast.coverage]}" != partial ||
    "${J[fast.executed_suites]}" != "$CI_SUITES_fast" ]]; then
    fail "fast: the executed partition (${J[fast.partition]}, ${J[fast.executed_profile]}: ${J[fast.executed_suites]}) is not the fast partition"
  fi
  rem="$(suites_minus "${J[fast.pipeline_suites]}" "$CI_SUITES_fast")"
  if [[ -z "$rem" || "${J[gate.partition]}" != remainder || "${J[gate.executed_profile]}" != "${J[gate.pipeline_profile]}" ||
    "${J[gate.coverage]}" != partial || "${J[gate.executed_suites]}" != "$rem" ]]; then
    fail "gate: the executed partition (${J[gate.partition]}, ${J[gate.executed_profile]}: ${J[gate.executed_suites]}) is not the remainder [$rem]"
  fi
  # E_pipeline: the pipeline suites plus the fast job's suites, plus final.
  pset="$(ids_of "${J[fast.pipeline_suites]} $CI_SUITES_fast")"
  if [[ "$(set_sha "$pset")" != "${J[fast.pipeline_set_sha256]}" ]]; then fail "pipeline_set_sha256 is not the registry set of the pipeline suites"; fi
  for k in fast gate; do if job_set "$k"; then :; else fail "$k: $WHY"; fi; done
  for k in fast gate; do if load_results "$k" "${J[$k.dir]}"; then :; else fail "$k: $WHY"; fi; done
  union="$(printf '%s\n%s\n' "${J[fast.rows]}" "${J[gate.rows]}" | sed '/^$/d' | sort -u)"
  gap="$(set_minus "$pset" "$union")"
  if [[ -n "$gap" ]]; then fail "union gap: $(count_of "$gap") IDs of the pipeline set have no PASS row (first $(first_of "$gap"))"; fi
  extra="$(set_minus "$union" "$pset")"
  if [[ -n "$extra" ]]; then fail "the union has $(count_of "$extra") IDs outside the pipeline set (first $(first_of "$extra"))"; fi
  overlap="$(set_and "${J[fast.rows]}" "${J[gate.rows]}")"
  finals="$(ids_of "")"
  if [[ "$overlap" != "$finals" ]]; then
    k="$(set_minus "$overlap" "$finals")"
    if [[ -n "$k" ]]; then fail "unexpected overlap: $(first_of "$k") is in both partitions"; fi
    fail "the overlap is not exactly the final IDs (first missing $(first_of "$(set_minus "$finals" "$overlap")"))"
  fi
  for k in fast gate; do if rows_exact "$k"; then :; else fail "$k: $WHY"; fi; done
  nf="$(count_of "${J[fast.rows]}")" ng="$(count_of "${J[gate.rows]}")" nu="$(count_of "$union")"
  result "PASS ${J[fast.pipeline_profile]} pipeline ${J[fast.head_sha]}: fast $nf + gate $ng -> $nu distinct (overlap: the $(count_of "$finals") final IDs)"
  exit 0
}

single_main() {
  local pset
  job_dir job "$A"
  if load_registry "$A/required-checks.tsv"; then :; else fail "job: required-checks.tsv is missing or malformed"; fi
  job_load job "$A"
  if [[ "${J[job.partition]}" != complete || "${J[job.coverage]}" != complete ||
    "${J[job.executed_profile]}" != "${J[job.pipeline_profile]}" || "${J[job.executed_suites]}" != "${J[job.pipeline_suites]}" ]]; then
    fail "job: not a complete single-job pipeline (${J[job.partition]}, ${J[job.coverage]}, ${J[job.executed_profile]}: ${J[job.executed_suites]})"
  fi
  pset="$(ids_of "${J[job.pipeline_suites]}")"
  if [[ "$(set_sha "$pset")" != "${J[job.pipeline_set_sha256]}" ]]; then fail "pipeline_set_sha256 is not the registry set of the pipeline suites"; fi
  if job_set job; then :; else fail "job: $WHY"; fi
  if load_results job "$A"; then :; else fail "job: $WHY"; fi
  if rows_exact job; then :; else fail "job: $WHY"; fi
  result "PASS single ${J[job.pipeline_profile]} pipeline ${J[job.head_sha]}: $(count_of "${J[job.rows]}") distinct (complete)"
  exit 0
}

# parse_kv FILE PREFIX: J[PREFIX.key]=value for the scalar ledger-summary keys, each exactly once.
LKEYS=(expected rows unknown duplicates near_duplicates exceptions bootstrap ac5_status)
parse_kv() {
  local line k re=$'^([a-z0-9_]{1,32})\t([^\t]{0,256})$'
  local -A n=()
  while IFS= read -r line || [[ -n "$line" ]]; do
    if [[ ! "$line" =~ $re ]]; then continue; fi
    k="${BASH_REMATCH[1]}"
    if [[ " ${LKEYS[*]} " != *" $k "* ]]; then continue; fi
    n[$k]=$(( ${n[$k]:-0} + 1 ))
    J[$2.$k]="${BASH_REMATCH[2]}"
  done <"$1"
  for k in "${LKEYS[@]}"; do
    if [[ "${n[$k]:-0}" != 1 ]]; then WHY="$k occurs ${n[$k]:-0} times"; return 1; fi
  done
  return 0
}

ledger_main() {
  local f out rc hosted=0 jobid line i ki kk kd nb=0 rec=""
  local -a v=() lines=() hdr=()
  job_dir job "$A"
  for f in summary.json builds.tsv ledger-summary.tsv ledger-bootstrap.tsv; do
    if [[ -L "$A/$f" || ! -f "$A/$f" ]]; then fail "ledger: $f is missing"; fi
  done
  out="$(yq_vec "$A/summary.json" '[(.format | tag), (.format | to_string), (.tested_content | tag), (.tested_content | to_string), (.ci.job_id | tag), (.ci.job_id | to_string)]')" || fail "ledger: summary.json is not JSON"
  IFS=$'\t' read -r -a v <<<"$out"
  if (( ${#v[@]} != 6 )) || [[ "${v[1]}" != platform-validation-summary/1 || "${v[2]}" != '!!str' ]] ||
    [[ ! "${v[3]}" =~ ^(local|branch-push|branch-schedule|branch-web|branch-api|mr-source|mr-merged-result|mr-rc-promotion)$ ]]; then
    fail "ledger: summary.json is malformed"
  fi
  jobid=""
  if [[ "${v[3]}" != local ]]; then
    hosted=1
    if [[ "${v[4]}" != '!!int' || ! "${v[5]}" =~ ^[1-9][0-9]{0,18}$ ]]; then fail "ledger: summary.json has no CI job ID"; fi
    jobid="${v[5]}"
  fi
  # The bootstrap record (validate.sh): a CI job needs the structural-inference reuse verdict, a local run none.
  mapfile -t lines <"$A/ledger-bootstrap.tsv"
  if (( ${#lines[@]} != 2 )) || [[ "${lines[0]}" != $'ref\tpre_id\tpost_id\tdockerfile_sha256\tbases\tplatform\tverdict' ]]; then
    fail "ledger: ledger-bootstrap.tsv is malformed"
  fi
  IFS=$'\t' read -r -a v <<<"${lines[1]}"
  if (( ${#v[@]} != 7 )); then fail "ledger: ledger-bootstrap.tsv is malformed"; fi
  if (( hosted )); then
    if [[ "${v[6]}" != structural-inference:reuse ]]; then fail "ledger: bootstrap verdict ${v[6]} (AC5 needs structural-inference:reuse)"; fi
    if [[ "${v[0]}" != "platform-ci-context:$jobid" || ! "${v[1]}" =~ ^sha256:[0-9a-f]{64}$ || "${v[2]}" != "${v[1]}" ]] ||
      [[ ! "${v[3]}" =~ ^[0-9a-f]{64}$ || ! "${v[4]}" =~ ^[^,[:space:]]+,[^,[:space:]]+$ || ! "${v[5]}" =~ ^[a-z0-9]+/[a-z0-9_]+$ ]]; then
      fail "ledger: the bootstrap reuse record is inconsistent"
    fi
  elif [[ "${lines[1]}" != $'none\tnone\tnone\tnone\tnone\tnone\tnot-applicable' ]]; then
    fail "ledger: a local run's bootstrap record is not not-applicable"
  fi
  # builds.tsv carries exactly the one bootstrap row of that image in a CI job, none locally.
  mapfile -t lines <"$A/builds.tsv"
  if (( ${#lines[@]} == 0 )); then fail "ledger: builds.tsv is empty"; fi
  IFS=$'\t' read -r -a hdr <<<"${lines[0]}"
  ki="" kk="" kd=""
  for i in "${!hdr[@]}"; do
    case "${hdr[$i]}" in kind) kk="$i" ;; image_id) ki="$i" ;; dockerfile_sha256) kd="$i" ;; esac
  done
  if [[ -z "$ki" || -z "$kk" || -z "$kd" ]]; then fail "ledger: builds.tsv has no kind, image_id or dockerfile_sha256 column"; fi
  for line in "${lines[@]:1}"; do
    IFS=$'\t' read -r -a hdr <<<"$line"
    if [[ "${hdr[$kk]:-}" != bootstrap ]]; then continue; fi
    nb=$((nb + 1))
    if (( hosted == 0 )); then fail "ledger: a local run's builds.tsv has a bootstrap row"; fi
    if [[ "${hdr[$ki]:-}" != "${v[1]}" || "${hdr[$kd]:-}" != "${v[3]}" ]]; then fail "ledger: the builds.tsv bootstrap row is not the recorded context image"; fi
  done
  if (( hosted && nb != 1 )) || (( hosted == 0 && nb != 0 )); then fail "ledger: builds.tsv has $nb bootstrap rows"; fi
  # Recompute over a copy (the published files are never written); the published summary must agree.
  if [[ -L "$REPO_R/infra/validation/lib.sh" || ! -f "$REPO_R/infra/validation/lib.sh" ]]; then fail "ledger: infra/validation/lib.sh is missing"; fi
  SCR="$(mktemp -d "${TMPDIR:-/tmp}/pipeline-union.XXXXXX")" || fail "ledger: mktemp failed"
  mkdir -m 700 -- "$SCR/job" "$SCR/work" || fail "ledger: cannot create the scratch directory"
  for f in "$A"/*; do
    if [[ -L "$f" || ! -f "$f" || "${f##*/}" == ledger-summary.tsv ]]; then continue; fi
    cp -- "$f" "$SCR/job/" || fail "ledger: cannot copy the evidence"
  done
  rc=0
  (
    cd -- "$SCR/work"
    WORK="$SCR/work" RESULTS="$SCR/work/results.tsv" TIMINGS="$SCR/work/timings.tsv" REPO="$REPO_R"
    export WORK RESULTS TIMINGS REPO
    # shellcheck source=../validation/lib.sh
    source "$REPO_R/infra/validation/lib.sh"
    if declare -F ledger_check_file >/dev/null; then :; else exit 3; fi
    ledger_check_file "$SCR/job"
  ) >"$SCR/out" 2>"$SCR/err" || rc=$?
  if (( rc == 3 )); then fail "ledger: infra/validation/lib.sh has no ledger_check_file"; fi
  if (( rc != 0 )); then fail "ledger: ledger_check_file exited $rc"; fi
  if grep -q $'^ac5_status\t' "$SCR/out"; then rec="$SCR/out"
  elif [[ -f "$SCR/job/ledger-summary.tsv" && ! -L "$SCR/job/ledger-summary.tsv" ]]; then rec="$SCR/job/ledger-summary.tsv"
  else fail "ledger: the recomputation produced no summary"; fi
  if parse_kv "$rec" rec; then :; else fail "ledger: the recomputed summary is malformed ($WHY)"; fi
  if parse_kv "$A/ledger-summary.tsv" pub; then :; else fail "ledger: the published ledger-summary.tsv is malformed ($WHY)"; fi
  for f in "${LKEYS[@]}"; do
    if [[ "${J[rec.$f]}" != "${J[pub.$f]}" ]]; then fail "ledger: the published ledger-summary.tsv differs from the recomputation ($f)"; fi
  done
  if [[ "${J[rec.ac5_status]}" != MET ]]; then
    fail "ledger: ac5_status ${J[rec.ac5_status]} (expected=${J[rec.expected]} rows=${J[rec.rows]} unknown=${J[rec.unknown]} duplicates=${J[rec.duplicates]} near_duplicates=${J[rec.near_duplicates]})"
  fi
  if (( hosted )); then line="bootstrap structural-inference:reuse (a structural inference, not step evidence)"; else line="bootstrap not-applicable (local)"; fi
  result "PASS ledger: ac5_status=MET (expected=${J[rec.expected]} rows=${J[rec.rows]} unknown=${J[rec.unknown]} duplicates=${J[rec.duplicates]}); $line"
  exit 0
}

case "$MODE" in
  union) J[fast.dir]="$A" J[gate.dir]="$B"; union_main ;;
  single) single_main ;;
  ledger) ledger_main ;;
esac
usage
