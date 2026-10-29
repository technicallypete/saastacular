#!/usr/bin/env bash
# Release-candidate verification (ONEOFF RC qualification; docs/agents/GITLAB-BRANCH-WORKFLOW.md "Release candidates").
# Offline and fail-closed: it reads local Git objects of this checkout and files the operator saved with the documented
# git/glab commands (git ls-remote listings, merge request/pipeline/jobs API JSON, the two jobs' ci-evidence artifacts).
# Those files are data only: JSON is read through the pinned yq image (no network), listings by fixed expressions;
# nothing in them is executed or sourced. It makes no network call and never creates, moves or pushes a ref.
#
#   release-verify.sh name TAG [LSREMOTE RECEIPT]        canonical RC name; with a complete pre-cut `git ls-remote origin`
#                                                       listing: the tag and release/TAG are absent (remote and local),
#                                                       and a new receipt records the staging commit at cut
#   release-verify.sh candidate TAG LSREMOTE RECEIPT     after the tag push: exactly one remote tag object and peel line,
#                                                       equal to the local annotated tag (named TAG), peel in the listed
#                                                       staging; first run: peel = the staging commit recorded at cut,
#                                                       then records object/peel/tree; later runs compare them
#   release-verify.sh qualified TAG RECEIPT PROJECT_ID MR_JSON PIPELINE_JSON JOBS_JSON FAST_DIR GATE_DIR
#                                                       the promotion merge request's latest pipeline is a complete,
#                                                       successful, correctly attributed full qualification of the RC
#   release-verify.sh promotable TAG RECEIPT MAIN_SHA    conflict-free merge of the RC into MAIN_SHA whose tree is the RC
#                                                       tree; records MAIN_SHA as the expected main
#   release-verify.sh recheck TAG RECEIPT LSREMOTE MR_JSON JOBS_JSON
#                                                       immediately before the merge: remote tag object/peel,
#                                                       release/TAG = peel, main = expected main; the merge request head,
#                                                       latest pipeline (freshly successful), its qualified jobs and
#                                                       settings unchanged
#   release-verify.sh promoted TAG RECEIPT MERGE_SHA LSREMOTE
#                                                       the actual main merge: parents exactly (expected main, RC), tree =
#                                                       RC tree, and on the freshly listed remote main; records MERGE_SHA
#   release-verify.sh merged TAG RECEIPT PROJECT_ID MR_JSON PIPELINE_JSON JOBS_JSON FAST_DIR GATE_DIR MERGE_SHA LSREMOTE
#                                                       evidence only, for a promotion MR that has already merged (in
#                                                       place of qualified, promotable, recheck and promoted, on a
#                                                       candidate receipt): the qualified checks with the MR merged as
#                                                       MERGE_SHA; parents exactly (P1, RC), tree = RC tree, the merge of
#                                                       the RC into P1 recreated conflict-free with the RC tree; on the
#                                                       listed remote main; tag object/peel unchanged; release/TAG
#                                                       deleted. Records the proof once, with VERIFICATION_MODE=
#                                                       merged-evidence; it infers no merge actor, mode or authority
#   release-verify.sh final pre|post RC_TAG FINAL_TAG RECEIPT LSREMOTE
#                                                       pre: FINAL_TAG = the RC's X.Y.Z, absent (remote and local), the
#                                                       verified merge on remote main; post: one annotated tag peeled to
#                                                       that merge (still on remote main) whose tree is the RC tree
#
# The receipt is a KEY=VALUE file this helper creates (name) and extends (one strict key set; each step requires the
# previous step's keys, so the order cut -> candidate -> qualified -> promotable -> recheck -> promoted -> final is
# enforced; for an already merged promotion MR, cut -> candidate -> merged -> final). Eligibility is not authority: a
# PASS of qualified makes the RC eligible only; merging and the final tag each need the owner's explicit instruction.
# GitLab binds the merge to the source SHA, not the target SHA; recheck and promoted detect main movement, they do not
# make it atomic.
#
# Output: exactly one line, "release-verify: PASS <command> ..." or "release-verify: FAIL <command>: <reason>"
# ("release-verify: usage ..."). Exit: 0 pass; 1 fail; 2 usage. Every check returns explicitly; no statement is negated
# with `!`.
set -euo pipefail
if (( BASH_VERSINFO[0] < 4 || (BASH_VERSINFO[0] == 4 && BASH_VERSINFO[1] < 4) )); then
  echo "release-verify: FAIL setup: bash >= 4.4 is required"
  exit 1
fi
shopt -s inherit_errexit
export LC_ALL=C
unset GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE GIT_OBJECT_DIRECTORY GIT_ALTERNATE_OBJECT_DIRECTORIES GIT_COMMON_DIR GIT_NAMESPACE

SELF_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)"
REPO_R="$(dirname -- "$(dirname -- "$SELF_DIR")")"
CMD="${1:-}"
usage() { echo "release-verify: usage: release-verify.sh name|candidate|qualified|promotable|recheck|promoted|merged|final ... (see the header)"; exit 2; }
fail() { echo "release-verify: FAIL $CMD: $*"; exit 1; }
pass() { echo "release-verify: PASS $CMD $*"; exit 0; }
if [[ -L "$SELF_DIR/lib.sh" || ! -f "$SELF_DIR/lib.sh" ]]; then fail "infra/ci/lib.sh is missing"; fi
# shellcheck source=lib.sh
source "$SELF_DIR/lib.sh"
if [[ -z "${CI_YQ_IMAGE:-}" ]]; then fail "infra/ci/lib.sh lacks the yq image"; fi

# The promotion target (the default branch) and the canonical names.
DEFAULT_BRANCH=main
RC_RE='^((0|[1-9][0-9]*)[.](0|[1-9][0-9]*)[.](0|[1-9][0-9]*))-rc[.]([1-9][0-9]*)$'
FINAL_RE='^(0|[1-9][0-9]*)[.](0|[1-9][0-9]*)[.](0|[1-9][0-9]*)$'
SHA_RE='^[0-9a-f]{40}$'
INT_RE='^[1-9][0-9]{0,18}$'

SCR=""
cleanup() { if [[ -n "$SCR" ]]; then rm -rf -- "$SCR"; fi; }
trap cleanup EXIT
scratch() { if [[ -z "$SCR" ]]; then SCR="$(mktemp -d "${TMPDIR:-/tmp}/release-verify.XXXXXX")" || fail "mktemp failed"; chmod 700 "$SCR"; fi; }

g() { git -C "$REPO_R" -c core.fsmonitor=false --no-optional-locks "$@"; }
commit_of() { g rev-parse --verify -q "$1^{commit}" 2>/dev/null; }
tree_of() { g rev-parse --verify -q "$1^{tree}" 2>/dev/null; }
file_sha() { sha256sum <"$1" | cut -c1-64; }
data_file() { if [[ -L "$1" || ! -f "$1" ]]; then fail "$2 is not a regular file"; fi; }

# --- receipt ---------------------------------------------------------------------------------------------------
RKEYS=(FORMAT TAG CUT_STAGING CUT_LSREMOTE_SHA256 TAG_OBJECT PEELED TREE PROJECT_ID MR_IID PIPELINE_ID FAST_JOB_ID
  GATE_JOB_ID EXPECTED_MAIN MAIN_MERGE_SHA VERIFICATION_MODE FINAL_TAG FINAL_PRE_LSREMOTE_SHA256 FINAL_TAG_OBJECT)
declare -A RRE=(
  [FORMAT]='^release-receipt/1$' [TAG]="${RC_RE}" [CUT_STAGING]="$SHA_RE" [CUT_LSREMOTE_SHA256]='^[0-9a-f]{64}$'
  [TAG_OBJECT]="$SHA_RE" [PEELED]="$SHA_RE" [TREE]="$SHA_RE" [PROJECT_ID]="$INT_RE" [MR_IID]="$INT_RE"
  [PIPELINE_ID]="$INT_RE" [FAST_JOB_ID]="$INT_RE" [GATE_JOB_ID]="$INT_RE" [EXPECTED_MAIN]="$SHA_RE"
  [MAIN_MERGE_SHA]="$SHA_RE" [VERIFICATION_MODE]='^merged-evidence$' [FINAL_TAG]="$FINAL_RE"
  [FINAL_PRE_LSREMOTE_SHA256]='^[0-9a-f]{64}$' [FINAL_TAG_OBJECT]="$SHA_RE"
)
declare -A R=()
receipt_load() { # FILE: R[KEY] from a regular file of known KEY=VALUE lines, each at most once and well formed
  local line k v
  data_file "$1" receipt
  while IFS= read -r line || [[ -n "$line" ]]; do
    if [[ ! "$line" =~ ^([A-Z][A-Z0-9_]*)=(.*)$ ]]; then fail "the receipt has a malformed line"; fi
    k="${BASH_REMATCH[1]}" v="${BASH_REMATCH[2]}"
    if [[ -z "${RRE[$k]+x}" ]]; then fail "the receipt has an unknown key $k"; fi
    if [[ -n "${R[$k]+x}" ]]; then fail "the receipt repeats $k"; fi
    if [[ ! "$v" =~ ${RRE[$k]} ]]; then fail "the receipt $k is malformed"; fi
    R[$k]="$v"
  done <"$1"
  if [[ "${R[FORMAT]-}" != release-receipt/1 ]]; then fail "the receipt has no release-receipt/1 FORMAT"; fi
}
receipt_need() { local k; for k in "$@"; do if [[ -z "${R[$k]+x}" ]]; then fail "the receipt lacks $k (an earlier step has not passed)"; fi; done; }
receipt_save() { # FILE: R in key order, written to a private temporary file and renamed over FILE
  local k
  (
    umask 077
    for k in "${RKEYS[@]}"; do if [[ -n "${R[$k]+x}" ]]; then printf '%s=%s\n' "$k" "${R[$k]}"; fi; done >"$1.tmp"
  ) || fail "cannot write the receipt"
  mv -f -- "$1.tmp" "$1" || fail "cannot replace the receipt"
}
receipt_tag() { # TAG: the receipt belongs to TAG
  if [[ "${R[TAG]-}" != "$1" ]]; then fail "the receipt belongs to another tag"; fi
}
receipt_manual() { # the manual steps (qualified, promotable, recheck, promoted) never extend a merged-evidence receipt
  if [[ -n "${R[VERIFICATION_MODE]+x}" ]]; then fail "the receipt holds merged evidence (VERIFICATION_MODE=${R[VERIFICATION_MODE]}): manual and merged evidence never mix"; fi
}

# --- git ls-remote listings --------------------------------------------------------------------------------------
declare -A LS=() LSN=()
ls_load() { # FILE: LS[ref]=sha and LSN[ref]=count; every line "<40 hex><TAB><ref>"
  local line
  data_file "$1" "the ls-remote listing"
  if [[ ! -s "$1" ]]; then fail "the ls-remote listing is empty"; fi
  while IFS= read -r line || [[ -n "$line" ]]; do
    if [[ ! "$line" =~ ^([0-9a-f]{40})$'\t'(refs/[^[:space:]]+|HEAD)$ ]]; then fail "the ls-remote listing has a malformed line"; fi
    LSN[${BASH_REMATCH[2]}]=$(( ${LSN[${BASH_REMATCH[2]}]:-0} + 1 ))
    LS[${BASH_REMATCH[2]}]="${BASH_REMATCH[1]}"
  done <"$1"
  # A complete listing of the expected remote: both long-lived branches, each once.
  ls_one refs/heads/staging; ls_one "refs/heads/$DEFAULT_BRANCH"
}
ls_one() { if [[ "${LSN[$1]:-0}" != 1 ]]; then fail "the listing has ${LSN[$1]:-0} lines for $1, want exactly 1"; fi; }
ls_none() { if [[ "${LSN[$1]:-0}" != 0 ]]; then fail "the listing has $1 (it must be absent)"; fi; }

# tag_identity TAG LOCAL_OBJECT_VAR: the remote tag object and peel lines (exactly one each) equal the local annotated
# tag named TAG; sets T_OBJ and T_PEEL.
T_OBJ="" T_PEEL=""
tag_identity() {
  local t="$1" typ hdr pk
  ls_one "refs/tags/$t"; ls_one "refs/tags/$t^{}"
  T_OBJ="$(g rev-parse --verify -q "refs/tags/$t" 2>/dev/null)" || T_OBJ=""
  if [[ ! "$T_OBJ" =~ $SHA_RE ]]; then fail "the tag is not present locally (fetch the tags)"; fi
  typ="$(g cat-file -t "$T_OBJ" 2>/dev/null)" || typ=""
  if [[ "$typ" != tag ]]; then fail "the tag is not annotated"; fi
  if [[ "$T_OBJ" != "${LS[refs/tags/$t]}" ]]; then fail "the local tag object differs from the remote one"; fi
  hdr="$(g cat-file tag "$T_OBJ" | sed -n '1,/^$/p')" || fail "cannot read the tag object"
  if [[ "$hdr" != *$'\n'"type commit"$'\n'* || "$hdr" != *$'\n'"tag $t"$'\n'* ]]; then fail "the tag object does not name $t on a commit"; fi
  T_PEEL="$(commit_of "refs/tags/$t")" || T_PEEL=""
  pk="refs/tags/$t^{}"
  if [[ ! "$T_PEEL" =~ $SHA_RE || "$T_PEEL" != "${LS[$pk]}" ]]; then fail "the local peeled commit differs from the remote one"; fi
}

# --- JSON (pinned yq image, no network) ----------------------------------------------------------------------------
# json_vec FILE EXPRESSION...: V[] = the tab-separated values of one yq array expression over FILE.
V=()
json_vec() {
  local f="$1" out
  shift
  data_file "$f" "$(basename -- "$f")"
  out="$(docker run --rm --network none -i "$CI_YQ_IMAGE" -p json -o tsv "$1" <"$f" 2>/dev/null)" || fail "$(basename -- "$f") is not JSON of the expected shape"
  IFS=$'\t' read -r -a V <<<"$out"
}
# typed EXPR...: a yq array of (tag, value) pairs for each path
typed() { local e="" s="" p; for p in "$@"; do e+="$s(.$p | tag), (.$p | to_string)"; s=", "; done; printf '[%s]' "$e"; }
# want N TAG REGEX NAME: pair N of V is TAG with a value matching REGEX; W = the value (fails in this shell)
W=""
want() {
  local i="$1"
  if [[ "${V[2 * i]-}" != "$2" || ! "${V[2 * i + 1]-}" =~ $3 ]]; then fail "$4 is missing or malformed"; fi
  W="${V[2 * i + 1]}"
}
# eq N TAG REGEX NAME EXPECTED MESSAGE: want, and the value is EXPECTED (otherwise fail with MESSAGE)
eq() { want "$1" "$2" "$3" "$4"; if [[ "$W" != "$5" ]]; then fail "$6"; fi; }
WORD='^[^[:space:]]+$' BOOL='^(true|false)$' TOKEN='^[a-z_]+$'

# mr_check MR_JSON PROJECT_ID TAG PEELED [PIPELINE_ID [MERGE_SHA]]: the promotion merge request (same project as
# source and target, no fork; release/TAG to the default branch; open; squash off; no queued auto-merge; head = the RC
# peel; its head (latest) pipeline is PIPELINE_ID when given, for that head). With MERGE_SHA (merged evidence only):
# merged, with merge_commit_sha = MERGE_SHA, instead of open with no queued auto-merge (the queue flag of a merged MR
# proves neither the merge mode nor its actor; it is not read). Sets MR_IID and MR_PIPE.
MR_IID="" MR_PIPE=""
mr_check() {
  local f="$1" pid="$2" t="$3" peel="$4" x ms="${6:-}" fields=26 extra=""
  if [[ -n "$ms" ]]; then fields=28 extra=merge_commit_sha; fi
  # shellcheck disable=SC2086 # extra is empty or one fixed field name
  json_vec "$f" "$(typed iid project_id source_project_id target_project_id source_branch target_branch sha state squash \
    merge_when_pipeline_succeeds head_pipeline.id head_pipeline.sha head_pipeline.status $extra)"
  if (( ${#V[@]} != fields )); then fail "the merge request JSON is malformed"; fi
  want 0 '!!int' "$INT_RE" "merge request iid"; MR_IID="$W"
  for x in 1 2 3; do eq "$x" '!!int' "$INT_RE" "merge request project" "$pid" "the merge request belongs to another project or a fork"; done
  eq 4 '!!str' "$WORD" source_branch "release/$t" "the merge request source branch is not release/$t"
  eq 5 '!!str' "$WORD" target_branch "$DEFAULT_BRANCH" "the merge request target is not $DEFAULT_BRANCH"
  eq 6 '!!str' "$SHA_RE" "merge request sha" "$peel" "the merge request head is not the RC peeled commit"
  if [[ -n "$ms" ]]; then
    eq 7 '!!str' "$TOKEN" state merged "the merge request is not merged"
    eq 13 '!!str' "$SHA_RE" merge_commit_sha "$ms" "the merge request's merge commit is not MERGE_SHA $ms"
  else
    eq 7 '!!str' "$TOKEN" state opened "the merge request is not open"
  fi
  eq 8 '!!bool' "$BOOL" squash false "the merge request squashes"
  if [[ -z "$ms" ]]; then eq 9 '!!bool' "$BOOL" merge_when_pipeline_succeeds false "the merge request has a queued auto-merge"; fi
  want 10 '!!int' "$INT_RE" head_pipeline.id; MR_PIPE="$W"
  eq 11 '!!str' "$SHA_RE" head_pipeline.sha "$peel" "the merge request's latest pipeline is not for the RC head"
  if [[ -n "${5:-}" && "$MR_PIPE" != "$5" ]]; then fail "the qualified pipeline is not the merge request's latest pipeline"; fi
  # Freshly successful: a retried job can leave the same pipeline ID and SHA running or failed.
  want 12 '!!str' "$TOKEN" head_pipeline.status
  if [[ "$W" != success ]]; then fail "the merge request's latest pipeline is $W, not success"; fi
}

# jobs_check JOBS_JSON PIPELINE PROJECT PEELED: exactly validate:checks and validate:integration, each successful and
# required, of this pipeline, project and commit. Sets FAST_JOB and GATE_JOB.
FAST_JOB="" GATE_JOB=""
jobs_check() {
  local jj="$1" pipe="$2" pid="$3" peel="$4" n i name id st af ppid pproj cid
  local -A seen=()
  FAST_JOB="" GATE_JOB=""
  # One tuple per job (map, then flatten: a bare ".[] | (a, b)" streams field by field across jobs in yq).
  json_vec "$jj" '[(. | tag), length] + (map([(.name | to_string), (.id | tag), (.id | to_string), (.status | to_string), (.allow_failure | tag), (.allow_failure | to_string), (.pipeline.id | to_string), (.pipeline.project_id | to_string), (.commit.id | to_string)]) | flatten)'
  if [[ "${V[0]-}" != '!!seq' || ! "${V[1]-}" =~ ^[0-9]+$ ]]; then fail "the jobs JSON is not a list"; fi
  n="${V[1]}"
  if (( n != 2 || ${#V[@]} != 2 + 9 * n )); then fail "the pipeline has $n jobs, want exactly validate:checks and validate:integration"; fi
  for ((i = 0; i < n; i++)); do
    name="${V[2 + 9 * i]}" id="${V[4 + 9 * i]}" st="${V[5 + 9 * i]}" af="${V[7 + 9 * i]}" ppid="${V[8 + 9 * i]}" pproj="${V[9 + 9 * i]}" cid="${V[10 + 9 * i]}"
    if [[ "$name" != validate:checks && "$name" != validate:integration ]]; then fail "unexpected job in the pipeline"; fi
    if [[ -n "${seen[$name]+x}" ]]; then fail "$name occurs twice (fetch the jobs without retried jobs)"; fi
    seen[$name]=1
    if [[ "${V[3 + 9 * i]}" != '!!int' || ! "$id" =~ $INT_RE ]]; then fail "$name has no job ID"; fi
    if [[ "$st" != success ]]; then fail "$name is $st, not success"; fi
    if [[ "${V[6 + 9 * i]}" != '!!bool' || "$af" != false ]]; then fail "$name is optional (allow_failure)"; fi
    if [[ "$ppid" != "$pipe" || "$pproj" != "$pid" ]]; then fail "$name belongs to another pipeline or project"; fi
    if [[ "$cid" != "$peel" ]]; then fail "$name tested another commit"; fi
    if [[ "$name" == validate:checks ]]; then FAST_JOB="$id"; else GATE_JOB="$id"; fi
  done
  if [[ -z "$FAST_JOB" || -z "$GATE_JOB" ]]; then fail "validate:checks or validate:integration is missing"; fi
}

# main_contains LSREMOTE_LOADED COMMIT: the freshly listed remote main is present locally and contains COMMIT (the
# promotion merge is published on main; main may have advanced since). Sets MAIN_TIP.
MAIN_TIP=""
main_contains() {
  MAIN_TIP="${LS[refs/heads/$DEFAULT_BRANCH]}"
  if [[ "$(commit_of "$MAIN_TIP")" != "$MAIN_TIP" ]]; then fail "the listed main $MAIN_TIP is not present locally (fetch main)"; fi
  if g merge-base --is-ancestor "$1" "$MAIN_TIP" 2>/dev/null; then :; else
    fail "the merge $1 is not on the remote main $MAIN_TIP (unpublished or replaced): no release claim or final tag"
  fi
}

# --- name ----------------------------------------------------------------------------------------------------------
cmd_name() {
  local t="${1:-}"
  if (( $# != 1 && $# != 3 )); then usage; fi
  if [[ ! "$t" =~ $RC_RE ]]; then fail "not a canonical RC name (MAJOR.MINOR.PATCH-rc.N, no leading zeroes, N >= 1, no v prefix or suffix)"; fi
  if (( $# == 1 )); then pass "$t: canonical RC name"; fi
  ls_load "$2"
  ls_none "refs/tags/$t"; ls_none "refs/tags/$t^{}"; ls_none "refs/heads/release/$t"
  if g rev-parse --verify -q "refs/tags/$t" >/dev/null 2>&1; then fail "a local tag $t already exists"; fi
  if [[ -e "$3" || -L "$3" ]]; then fail "the receipt already exists (an RC name is never reused)"; fi
  R=([FORMAT]=release-receipt/1 [TAG]="$t" [CUT_STAGING]="${LS[refs/heads/staging]}" [CUT_LSREMOTE_SHA256]="$(file_sha "$2")")
  receipt_save "$3"
  pass "$t: canonical; absent from the remote listing (tag, peel, release/$t) and locally; cut from staging ${LS[refs/heads/staging]}"
}

# --- candidate -----------------------------------------------------------------------------------------------------
cmd_candidate() {
  local t="${1:-}" stg tree
  if (( $# != 3 )); then usage; fi
  if [[ ! "$t" =~ $RC_RE ]]; then fail "not a canonical RC name"; fi
  receipt_load "$3"; receipt_tag "$t"; receipt_need CUT_STAGING CUT_LSREMOTE_SHA256
  ls_load "$2"
  tag_identity "$t"
  stg="${LS[refs/heads/staging]}"
  if [[ -z "$(commit_of "$stg")" ]]; then fail "the listed staging commit is not present locally (fetch staging)"; fi
  if g merge-base --is-ancestor "$T_PEEL" "$stg" 2>/dev/null; then :; else fail "the RC commit is not in the listed staging"; fi
  tree="$(tree_of "$T_PEEL")" || fail "cannot read the RC tree"
  if [[ -n "${R[TAG_OBJECT]+x}" ]]; then
    if [[ "${R[TAG_OBJECT]}" != "$T_OBJ" ]]; then fail "the tag object differs from the cut receipt (substituted or recreated tag)"; fi
    if [[ "${R[PEELED]}" != "$T_PEEL" || "${R[TREE]}" != "$tree" ]]; then fail "the peeled commit or tree differs from the cut receipt (moved tag)"; fi
    pass "$t: unchanged tag object $T_OBJ, peeled $T_PEEL, tree $tree; in staging $stg"
  fi
  # First binding: the tag must peel to exactly the reviewed staging commit recorded at cut (not an older ancestor,
  # not a newer commit staging reached later). Afterwards only the recorded identity and staging ancestry count.
  if [[ "$T_PEEL" != "${R[CUT_STAGING]}" ]]; then fail "the tag peels to $T_PEEL, not the staging commit recorded at cut ${R[CUT_STAGING]}"; fi
  R[TAG_OBJECT]="$T_OBJ" R[PEELED]="$T_PEEL" R[TREE]="$tree"
  receipt_save "$3"
  pass "$t: annotated tag object $T_OBJ, peeled $T_PEEL, tree $tree; in staging $stg (recorded)"
}

# --- qualified -----------------------------------------------------------------------------------------------------
# ids_of REGISTRY SUITES: the sorted registry IDs of SUITES plus final (as validate.sh and pipeline-union.sh compute).
ids_of() {
  local id suite set=" $2 final "
  while IFS=$'\t' read -r id suite; do
    if [[ -z "$id" || "$id" == '#'* ]]; then continue; fi
    if [[ "$set" == *" $suite "* ]]; then printf '%s\n' "$id"; fi
  done <"$1" | sort
}
set_sha() { printf '%s\n' "$1" | sha256sum | cut -c1-64; }
count_of() { if [[ -z "$1" ]]; then printf '0'; else printf '%s\n' "$1" | grep -c .; fi; }
suites_minus() { local s o="" sp=""; for s in $1; do if [[ " $2 " == *" $s "* ]]; then continue; fi; o+="$sp$s"; sp=" "; done; printf '%s' "$o"; }

# artifact_check KIND DIR PROJECT PIPELINE JOB TAG PEELED TREE IID: one job's published summary (data only).
# Sets A_PSUITES, A_AREAS, A_DIGEST, A_PSET, A_RSET and A_RCOUNT.
A_PSUITES="" A_AREAS="" A_DIGEST="" A_PSET="" A_RSET="" A_RCOUNT=""
artifact_check() {
  local kind="$1" d="$2" stg part n=25
  if [[ -L "$d" || ! -d "$d" ]]; then fail "the $kind artifact directory is missing"; fi
  json_vec "$d/summary.json" "$(typed format status_kind candidate tested_content pipeline_profile partition head_sha tree_sha \
    ci.project_id ci.pipeline_id ci.job_id ci.commit_sha ci.pipeline_source ci.merge_request_iid ci.merge_request_event_type \
    release.tag release.tag_object release.peeled release.staging_sha release.merge_request_iid release.source_branch \
    release.target_branch pipeline_suites areas selection_digest)"
  if (( ${#V[@]} != 2 * n )); then fail "$kind: summary.json is malformed"; fi
  eq 0 '!!str' "$WORD" format platform-validation-summary/1 "$kind: not a platform-validation-summary/1"
  eq 1 '!!str' "$WORD" status_kind pre-publication "$kind: not a pre-publication summary"
  eq 2 '!!str' "$WORD" candidate full "$kind: not a full candidate (minimal or withheld evidence)"
  eq 3 '!!str' "$WORD" tested_content mr-rc-promotion "$kind: tested_content is not mr-rc-promotion"
  eq 4 '!!str' "$WORD" pipeline_profile full "$kind: the pipeline profile is not full"
  want 5 '!!str' "$WORD" partition; part="$W"
  if [[ "$kind/$part" != fast/fast && "$kind/$part" != gate/remainder ]]; then fail "$kind: partition $part"; fi
  eq 6 '!!str' "$SHA_RE" head_sha "$7" "$kind: the candidate commit is not the RC peeled commit"
  eq 7 '!!str' "$SHA_RE" tree_sha "$8" "$kind: the candidate tree is not the RC tree"
  eq 8 '!!int' "$INT_RE" ci.project_id "$3" "$kind: evidence of another project"
  eq 9 '!!int' "$INT_RE" ci.pipeline_id "$4" "$kind: evidence of another pipeline"
  eq 10 '!!int' "$INT_RE" ci.job_id "$5" "$kind: evidence of another job"
  eq 11 '!!str' "$SHA_RE" ci.commit_sha "$7" "$kind: CI_COMMIT_SHA is not the RC peeled commit"
  eq 12 '!!str' "$TOKEN" ci.pipeline_source merge_request_event "$kind: not a merge request pipeline"
  eq 13 '!!int' "$INT_RE" ci.merge_request_iid "$9" "$kind: evidence of another merge request"
  eq 14 '!!str' "$TOKEN" ci.merge_request_event_type detached "$kind: not a detached merge request pipeline"
  eq 15 '!!str' "$RC_RE" release.tag "$6" "$kind: the release tag differs"
  eq 16 '!!str' "$SHA_RE" release.tag_object "${R[TAG_OBJECT]}" "$kind: the release tag object is not the cut tag object"
  eq 17 '!!str' "$SHA_RE" release.peeled "$7" "$kind: the release peeled commit differs"
  want 18 '!!str' "$SHA_RE" release.staging_sha; stg="$W"
  eq 19 '!!int' "$INT_RE" release.merge_request_iid "$9" "$kind: the release merge request differs"
  eq 20 '!!str' "$WORD" release.source_branch "release/$6" "$kind: the release source branch differs"
  eq 21 '!!str' "$WORD" release.target_branch "$DEFAULT_BRANCH" "$kind: the release target differs"
  # This job's own observation of staging (independent of the other job's): present locally and containing the RC.
  if [[ -z "$(commit_of "$stg")" ]]; then fail "$kind: the observed staging commit is not present locally (fetch staging)"; fi
  if g merge-base --is-ancestor "$7" "$stg" 2>/dev/null; then :; else fail "$kind: the RC commit is not in the job's observed staging"; fi
  want 22 '!!str' '^[a-z][a-z -]*$' pipeline_suites; A_PSUITES="$W"
  want 23 '!!str' '^[a-z0-9:,-]+$' areas; A_AREAS="$W"
  want 24 '!!str' '^[0-9a-f]{64}$' selection_digest; A_DIGEST="$W"
  json_vec "$d/summary.json" "$(typed pipeline_set_sha256 required_set_sha256 required_count)"
  if (( ${#V[@]} != 6 )); then fail "$kind: summary.json is malformed"; fi
  want 0 '!!str' '^[0-9a-f]{64}$' pipeline_set_sha256; A_PSET="$W"
  want 1 '!!str' '^[0-9a-f]{64}$' required_set_sha256; A_RSET="$W"
  want 2 '!!int' "$INT_RE" required_count; A_RCOUNT="$W"
}

# qualification TAG PROJECT_ID MR_JSON PIPELINE_JSON JOBS_JSON FAST_DIR GATE_DIR [MERGE_SHA]: the merge request's
# latest pipeline is a complete, successful, correctly attributed full qualification of the receipt's RC (shared by
# qualified and merged; MERGE_SHA selects mr_check's merged evidence). It neither saves the receipt nor passes: on
# success it sets Q_PIPE, Q_FJOB, Q_GJOB (MR_IID from mr_check) and Q_COV (the coverage summary); any refusal fails.
Q_PIPE="" Q_FJOB="" Q_GJOB="" Q_COV=""
qualification() {
  local t="$1" pid="$2" mrj="$3" pj="$4" jj="$5" fd="$6" gd="$7" ms="${8:-}"
  local pipe fjob="" gjob="" k u fast_suites ps fs rs us rc=0 sel wt line
  # The local tag is still the recorded one (the merge request binds to the tag through its head).
  if [[ "$(g rev-parse --verify -q "refs/tags/$t" 2>/dev/null)" != "${R[TAG_OBJECT]}" || "$(commit_of "refs/tags/$t")" != "${R[PEELED]}" ]]; then
    fail "the local tag differs from the receipt"
  fi
  # Pipeline: this project, a merge request pipeline of the merge request head ref, the RC commit, successful, no tag.
  json_vec "$pj" "$(typed id project_id sha ref status source tag)"
  if (( ${#V[@]} != 14 )); then fail "the pipeline JSON is malformed"; fi
  # Tag-shaped proof is refused first and explicitly: tag pipelines never qualify an RC promotion.
  eq 6 '!!bool' "$BOOL" "pipeline tag" false "a tag pipeline is not a promotion qualification"
  want 0 '!!int' "$INT_RE" "pipeline id"; pipe="$W"
  eq 1 '!!int' "$INT_RE" "pipeline project_id" "$pid" "the pipeline belongs to another project"
  eq 2 '!!str' "$SHA_RE" "pipeline sha" "${R[PEELED]}" "the pipeline did not test the RC peeled commit"
  want 3 '!!str' "$WORD" "pipeline ref"; line="$W"
  eq 4 '!!str' "$TOKEN" "pipeline status" success "the pipeline did not succeed"
  eq 5 '!!str' "$TOKEN" "pipeline source" merge_request_event "not a merge request pipeline"
  # Merge request: the promotion, its head the RC, this pipeline its latest; the pipeline runs on its head ref.
  mr_check "$mrj" "$pid" "$t" "${R[PEELED]}" "$pipe" "$ms"
  if [[ "$line" != "refs/merge-requests/$MR_IID/head" ]]; then fail "the pipeline is not the detached pipeline of merge request $MR_IID"; fi
  # Jobs: exactly validate:checks and validate:integration, each successful and required, of this pipeline and commit.
  jobs_check "$jj" "$pipe" "$pid" "${R[PEELED]}"
  fjob="$FAST_JOB" gjob="$GATE_JOB"
  # Artifacts: each job's own summary, attributed to this project, pipeline, job, merge request, RC and tree.
  artifact_check fast "$fd" "$pid" "$pipe" "$fjob" "$t" "${R[PEELED]}" "${R[TREE]}" "$MR_IID"
  local f_ps="$A_PSUITES" f_ar="$A_AREAS" f_dg="$A_DIGEST" f_pset="$A_PSET" f_rset="$A_RSET" f_rc="$A_RCOUNT"
  artifact_check gate "$gd" "$pid" "$pipe" "$gjob" "$t" "${R[PEELED]}" "${R[TREE]}" "$MR_IID"
  # The union of the two partitions (pipeline-union.sh of this checkout).
  line="$(bash "$SELF_DIR/pipeline-union.sh" "$fd" "$gd" 2>&1)" || rc=$?
  if (( rc != 0 )); then fail "the pipeline union fails: ${line#pipeline-union: }"; fi
  # Complete coverage, recomputed from the candidate itself: its own selector, map, suite lists and registry, run in a
  # private detached clone of the RC commit with CI_CONTROLS_FORCE=0 (a self-consistent reduced set cannot pass).
  scratch
  wt="$SCR/candidate"
  g clone -q --shared --no-checkout -- "$REPO_R" "$wt" >/dev/null 2>&1 || fail "cannot clone the candidate"
  git -C "$wt" -c advice.detachedHead=false checkout -q --detach "${R[PEELED]}" >/dev/null 2>&1 || fail "cannot check out the candidate"
  for k in infra/ci/select-profile.sh infra/ci/lib.sh infra/ci/profile-paths.tsv infra/validation/required-checks.tsv; do
    if [[ -L "$wt/$k" || ! -f "$wt/$k" ]]; then fail "the candidate lacks $k"; fi
  done
  mkdir -m 700 "$SCR/tmp" || fail "mktemp failed"
  sel="$SCR/selection.env"
  rc=0
  env -i PATH="$PATH" HOME="${HOME:-/nonexistent}" LC_ALL=C TMPDIR="$SCR/tmp" GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 CI_CONTROLS_FORCE=0 \
    bash "$wt/infra/ci/select-profile.sh" --job gate --tested-content mr-rc-promotion --map "$wt/infra/ci/profile-paths.tsv" \
    --out "$sel" --tsv "$SCR/selection.tsv" >"$SCR/selection.log" 2>&1 || rc=$?
  if (( rc != 0 )); then fail "the candidate's selector refused the RC promotion context"; fi
  declare -A S=()
  while IFS= read -r line || [[ -n "$line" ]]; do
    if [[ "$line" =~ ^([A-Z_]+)=(.*)$ ]]; then S[${BASH_REMATCH[1]}]="${BASH_REMATCH[2]}"; fi
  done <"$sel"
  if [[ "${S[PIPELINE_PROFILE]-}/${S[SELECTION_REASON]-}" != full/rc-promotion ]]; then fail "the candidate's selector does not select full (rc-promotion)"; fi
  ps="${S[PIPELINE_SUITES]-}"
  if [[ " $ps " != *" ci-controls "* ]]; then fail "the recomputed full profile lacks the CI meta-tests"; fi
  for k in fast gate; do
    if [[ "$k" == fast ]]; then u="$f_ps/$f_ar/$f_dg"; else u="$A_PSUITES/$A_AREAS/$A_DIGEST"; fi
    if [[ "$u" != "$ps/${S[AREAS]-}/${S[SELECTION_DIGEST]-}" ]]; then fail "$k: the published pipeline suites, areas or selection digest differ from the candidate's recomputation"; fi
  done
  for k in "$fd" "$gd"; do
    if cmp -s -- "$wt/infra/validation/required-checks.tsv" "$k/required-checks.tsv"; then :; else fail "a published registry is not the candidate's registry"; fi
  done
  fast_suites="$(sed -n 's/^CI_SUITES_fast="\([a-z -]*\)"$/\1/p' "$wt/infra/ci/lib.sh")"
  if [[ -z "$fast_suites" ]]; then fail "the candidate's fast suites cannot be read"; fi
  fs="$(ids_of "$wt/infra/validation/required-checks.tsv" "$fast_suites")"
  rs="$(ids_of "$wt/infra/validation/required-checks.tsv" "$(suites_minus "$ps" "$fast_suites")")"
  us="$(ids_of "$wt/infra/validation/required-checks.tsv" "$ps $fast_suites")"
  if [[ "$f_rset/$f_rc" != "$(set_sha "$fs")/$(count_of "$fs")" ]]; then fail "fast: the required set is not the candidate's fast set"; fi
  if [[ "$A_RSET/$A_RCOUNT" != "$(set_sha "$rs")/$(count_of "$rs")" ]]; then fail "gate: the required set is not the candidate's full remainder"; fi
  if [[ "$f_pset" != "$(set_sha "$us")" || "$A_PSET" != "$(set_sha "$us")" ]]; then fail "the pipeline set is not the candidate's complete full set"; fi
  Q_PIPE="$pipe" Q_FJOB="$fjob" Q_GJOB="$gjob"
  Q_COV="full fast $(count_of "$fs") + remainder $(count_of "$rs") -> $(count_of "$us") required IDs incl. ci-controls, areas ${S[AREAS]-}"
}

cmd_qualified() {
  local t="${1:-}" rf="${2:-}" pid="${3:-}"
  if (( $# != 8 )); then usage; fi
  if [[ ! "$t" =~ $RC_RE ]]; then fail "not a canonical RC name"; fi
  if [[ ! "$pid" =~ $INT_RE ]]; then fail "PROJECT_ID is not a project ID"; fi
  receipt_load "$rf"; receipt_tag "$t"; receipt_manual; receipt_need TAG_OBJECT PEELED TREE
  qualification "$t" "$pid" "$4" "$5" "$6" "$7" "$8"
  R[PROJECT_ID]="$pid" R[MR_IID]="$MR_IID" R[PIPELINE_ID]="$Q_PIPE" R[FAST_JOB_ID]="$Q_FJOB" R[GATE_JOB_ID]="$Q_GJOB"
  receipt_save "$rf"
  pass "$t: eligible (owner promotion instruction still required): MR !$MR_IID pipeline $Q_PIPE (jobs $Q_FJOB, $Q_GJOB) project $pid; commit ${R[PEELED]} tree ${R[TREE]}; $Q_COV"
}

# --- promotable ----------------------------------------------------------------------------------------------------
cmd_promotable() {
  local t="${1:-}" rf="${2:-}" m="${3:-}" out rc=0 mt
  if (( $# != 3 )); then usage; fi
  receipt_load "$rf"; receipt_tag "$t"; receipt_manual; receipt_need TAG_OBJECT PEELED TREE PIPELINE_ID
  if [[ ! "$m" =~ $SHA_RE || "$(commit_of "$m")" != "$m" ]]; then fail "MAIN_SHA is not a local commit (fetch main)"; fi
  out="$(g merge-tree --write-tree --no-messages "$m" "${R[PEELED]}" 2>/dev/null)" || rc=$?
  if (( rc == 1 )); then fail "merging the RC into main $m conflicts: integrate on staging and cut a new RC"; fi
  if (( rc != 0 )); then fail "git merge-tree failed (exit $rc)"; fi
  mt="${out%%$'\n'*}"
  if [[ "$mt" != "${R[TREE]}" ]]; then fail "the prospective merge tree $mt differs from the RC tree ${R[TREE]}: main has content the RC lacks; integrate on staging and cut a new RC"; fi
  R[EXPECTED_MAIN]="$m"
  receipt_save "$rf"
  pass "$t: merging into main $m is conflict-free with tree ${R[TREE]} = the RC tree (expected main recorded)"
}

# --- recheck -------------------------------------------------------------------------------------------------------
cmd_recheck() {
  local t="${1:-}" rf="${2:-}" pk
  if (( $# != 5 )); then usage; fi
  receipt_load "$rf"; receipt_tag "$t"; receipt_manual; receipt_need TAG_OBJECT PEELED TREE PROJECT_ID MR_IID PIPELINE_ID FAST_JOB_ID GATE_JOB_ID EXPECTED_MAIN
  ls_load "$3"
  ls_one "refs/tags/$t"; ls_one "refs/tags/$t^{}"; ls_one "refs/heads/release/$t"
  if [[ "${LS[refs/tags/$t]}" != "${R[TAG_OBJECT]}" ]]; then fail "the remote tag object changed (substituted or recreated tag)"; fi
  pk="refs/tags/$t^{}"
  if [[ "${LS[$pk]}" != "${R[PEELED]}" ]]; then fail "the remote tag peel changed (moved tag)"; fi
  if [[ "${LS[refs/heads/release/$t]}" != "${R[PEELED]}" ]]; then fail "release/$t is not the RC peeled commit"; fi
  if [[ "${LS[refs/heads/$DEFAULT_BRANCH]}" != "${R[EXPECTED_MAIN]}" ]]; then
    fail "main moved to ${LS[refs/heads/$DEFAULT_BRANCH]}: run promotable on it (equal tree, no conflict), then recheck again"
  fi
  mr_check "$4" "${R[PROJECT_ID]}" "$t" "${R[PEELED]}" "${R[PIPELINE_ID]}"
  if [[ "$MR_IID" != "${R[MR_IID]}" ]]; then fail "another merge request"; fi
  # The pipeline's current jobs are still the qualified ones; a retried job (new identity) needs a refreshed qualified.
  jobs_check "$5" "${R[PIPELINE_ID]}" "${R[PROJECT_ID]}" "${R[PEELED]}"
  if [[ "$FAST_JOB/$GATE_JOB" != "${R[FAST_JOB_ID]}/${R[GATE_JOB_ID]}" ]]; then
    fail "the pipeline's jobs are now $FAST_JOB/$GATE_JOB, not the qualified ${R[FAST_JOB_ID]}/${R[GATE_JOB_ID]}: re-run qualified with their artifacts"
  fi
  pass "$t: tag object, peel, release/$t, main ${R[EXPECTED_MAIN]} and MR !$MR_IID (head, latest pipeline ${R[PIPELINE_ID]} success with the qualified jobs, squash off, no auto-merge) unchanged; merge now with sha=${R[PEELED]} squash=false (GitLab binds the source SHA only; the target is not compared atomically)"
}

# --- promoted ------------------------------------------------------------------------------------------------------
cmd_promoted() {
  local t="${1:-}" rf="${2:-}" m="${3:-}" parents
  if (( $# != 4 )); then usage; fi
  receipt_load "$rf"; receipt_tag "$t"; receipt_manual; receipt_need TAG_OBJECT PEELED TREE PIPELINE_ID EXPECTED_MAIN
  ls_load "$4"
  if [[ ! "$m" =~ $SHA_RE || "$(commit_of "$m")" != "$m" ]]; then fail "MERGE_SHA is not a local commit (fetch main)"; fi
  parents="$(g cat-file commit "$m" | sed -n '1,/^$/ s/^parent //p' | tr '\n' ' ')" || fail "cannot read the merge commit"
  if [[ "$parents" != "${R[EXPECTED_MAIN]} ${R[PEELED]} " ]]; then
    fail "the merge's parents are not exactly (expected main ${R[EXPECTED_MAIN]}, RC ${R[PEELED]}): stop, no release claim or final tag; escalate"
  fi
  if [[ "$(tree_of "$m")" != "${R[TREE]}" ]]; then fail "the main merge tree differs from the RC tree: stop, no release claim or final tag; escalate"; fi
  # Publication: the fresh listing's remote main contains the merge (equal, or main advanced past it).
  main_contains "$m"
  R[MAIN_MERGE_SHA]="$m"
  receipt_save "$rf"
  pass "$t: main merge $m has parents (${R[EXPECTED_MAIN]}, ${R[PEELED]}) and the RC tree ${R[TREE]}, published on remote main $MAIN_TIP (recorded)"
}

# --- merged (evidence only) -----------------------------------------------------------------------------------------
# For a promotion MR that has already merged (for example through an owner-selected auto-merge), this verifies the same
# qualification and the actual main merge from saved evidence, on a candidate receipt, in place of qualified,
# promotable, recheck and promoted. It verifies; it authorizes nothing and infers neither the merge mode, its actor nor
# the owner's authority (a merged MR and its current queue flag prove none of them). The receipt is written once, only
# after every proof passed; any refusal leaves it byte for byte unchanged.
cmd_merged() {
  local t="${1:-}" rf="${2:-}" pid="${3:-}" m="${9:-}" k pk out rc=0 mt p1
  local -a par=()
  if (( $# != 10 )); then usage; fi
  if [[ ! "$t" =~ $RC_RE ]]; then fail "not a canonical RC name"; fi
  if [[ ! "$pid" =~ $INT_RE ]]; then fail "PROJECT_ID is not a project ID"; fi
  receipt_load "$rf"; receipt_tag "$t"; receipt_need TAG_OBJECT PEELED TREE
  # Only a candidate receipt: manual and merged evidence never mix in one receipt.
  for k in PROJECT_ID MR_IID PIPELINE_ID FAST_JOB_ID GATE_JOB_ID EXPECTED_MAIN MAIN_MERGE_SHA VERIFICATION_MODE FINAL_TAG \
    FINAL_PRE_LSREMOTE_SHA256 FINAL_TAG_OBJECT; do
    if [[ -n "${R[$k]+x}" ]]; then fail "the receipt already holds $k (merged evidence needs a candidate receipt; manual and merged evidence never mix)"; fi
  done
  ls_load "${10}"
  if [[ ! "$m" =~ $SHA_RE || "$(commit_of "$m")" != "$m" ]]; then fail "MERGE_SHA is not a local commit (fetch main)"; fi
  # Qualification, with the merge request merged as MERGE_SHA.
  qualification "$t" "$pid" "$4" "$5" "$6" "$7" "$8" "$m"
  # The actual merge: exactly two ordered parents (P1, RC) and the RC tree.
  out="$(g cat-file commit "$m" | sed -n '1,/^$/ s/^parent //p' | tr '\n' ' ')" || fail "cannot read the merge commit"
  read -r -a par <<<"$out"
  if (( ${#par[@]} != 2 )) || [[ "${par[1]}" != "${R[PEELED]}" || "${par[0]}" == "${R[PEELED]}" ]]; then
    fail "the merge's parents are not exactly (main before the merge, RC ${R[PEELED]}): stop, no release claim or final tag; escalate"
  fi
  p1="${par[0]}"
  if [[ "$(tree_of "$m")" != "${R[TREE]}" ]]; then fail "the main merge tree differs from the RC tree: stop, no release claim or final tag; escalate"; fi
  # promotable, retroactively: merging the RC into P1 is conflict-free and yields exactly the RC tree.
  out="$(g merge-tree --write-tree --no-messages "$p1" "${R[PEELED]}" 2>/dev/null)" || rc=$?
  if (( rc == 1 )); then fail "merging the RC into P1 $p1 conflicts: stop, no release claim or final tag; escalate"; fi
  if (( rc != 0 )); then fail "git merge-tree failed (exit $rc)"; fi
  mt="${out%%$'\n'*}"
  if [[ "$mt" != "${R[TREE]}" ]]; then fail "the merge of the RC into P1 $p1 recreates tree $mt, not the RC tree ${R[TREE]}: stop, no release claim or final tag; escalate"; fi
  # Publication and the remote state after the merge: on the listed main, the tag unchanged, release/TAG deleted.
  main_contains "$m"
  ls_one "refs/tags/$t"; ls_one "refs/tags/$t^{}"
  if [[ "${LS[refs/tags/$t]}" != "${R[TAG_OBJECT]}" ]]; then fail "the remote tag object changed (substituted or recreated tag)"; fi
  pk="refs/tags/$t^{}"
  if [[ "${LS[$pk]}" != "${R[PEELED]}" ]]; then fail "the remote tag peel changed (moved tag)"; fi
  if [[ "${LSN[refs/heads/release/$t]:-0}" != 0 ]]; then fail "release/$t is still on the remote (verify its tip, delete only that branch, list again)"; fi
  R[PROJECT_ID]="$pid" R[MR_IID]="$MR_IID" R[PIPELINE_ID]="$Q_PIPE" R[FAST_JOB_ID]="$Q_FJOB" R[GATE_JOB_ID]="$Q_GJOB"
  R[EXPECTED_MAIN]="$p1" R[MAIN_MERGE_SHA]="$m" R[VERIFICATION_MODE]=merged-evidence
  receipt_save "$rf"
  pass "$t: merged evidence (verification only; no merge mode, actor or authority inferred): MR !$MR_IID merged as $m with parents ($p1, ${R[PEELED]}) and the RC tree ${R[TREE]}, recreated conflict-free from $p1; on remote main $MAIN_TIP; tag object ${R[TAG_OBJECT]} unchanged; release/$t deleted; pipeline $Q_PIPE (jobs $Q_FJOB, $Q_GJOB) project $pid; $Q_COV (recorded)"
}

# --- final ---------------------------------------------------------------------------------------------------------
cmd_final() {
  local phase="${1:-}" t="${2:-}" ft="${3:-}" rf="${4:-}"
  if (( $# != 5 )) || [[ "$phase" != pre && "$phase" != post ]]; then usage; fi
  CMD="final $phase"
  if [[ ! "$t" =~ $RC_RE ]]; then fail "not a canonical RC name"; fi
  if [[ ! "$ft" =~ $FINAL_RE || "$ft" != "${t%%-rc.*}" ]]; then fail "the final tag must be exactly the RC's MAJOR.MINOR.PATCH (${t%%-rc.*})"; fi
  receipt_load "$rf"; receipt_tag "$t"; receipt_need TAG_OBJECT PEELED TREE MAIN_MERGE_SHA
  if [[ "$(tree_of "${R[MAIN_MERGE_SHA]}")" != "${R[TREE]}" ]]; then fail "the verified main merge does not carry the RC tree"; fi
  ls_load "$5"
  main_contains "${R[MAIN_MERGE_SHA]}"
  if [[ "$phase" == pre ]]; then
    if [[ -n "${R[FINAL_TAG_OBJECT]+x}" ]]; then fail "the final tag is already recorded"; fi
    ls_none "refs/tags/$ft"; ls_none "refs/tags/$ft^{}"
    if g rev-parse --verify -q "refs/tags/$ft" >/dev/null 2>&1; then fail "a local tag $ft already exists"; fi
    R[FINAL_TAG]="$ft" R[FINAL_PRE_LSREMOTE_SHA256]="$(file_sha "$5")"
    receipt_save "$rf"
    pass "$ft: absent from the remote listing and locally; tag the verified main merge ${R[MAIN_MERGE_SHA]} (annotated), push it, then run final post"
  fi
  receipt_need FINAL_TAG FINAL_PRE_LSREMOTE_SHA256
  if [[ "${R[FINAL_TAG]}" != "$ft" ]]; then fail "the pre-creation absence proof is for another final tag"; fi
  tag_identity "$ft"
  if [[ "$T_PEEL" != "${R[MAIN_MERGE_SHA]}" ]]; then fail "the final tag does not point at the verified main merge ${R[MAIN_MERGE_SHA]}"; fi
  if [[ -n "${R[FINAL_TAG_OBJECT]+x}" && "${R[FINAL_TAG_OBJECT]}" != "$T_OBJ" ]]; then fail "the final tag object differs from the recorded one"; fi
  R[FINAL_TAG_OBJECT]="$T_OBJ"
  receipt_save "$rf"
  pass "$ft: annotated tag object $T_OBJ on the verified main merge ${R[MAIN_MERGE_SHA]} (tree ${R[TREE]} = RC $t)"
}

if (( $# == 0 )); then usage; fi
shift
case "$CMD" in
  name) cmd_name "$@" ;;
  candidate) cmd_candidate "$@" ;;
  qualified) cmd_qualified "$@" ;;
  promotable) cmd_promotable "$@" ;;
  recheck) cmd_recheck "$@" ;;
  promoted) cmd_promoted "$@" ;;
  merged) cmd_merged "$@" ;;
  final) cmd_final "$@" ;;
  *) usage ;;
esac
usage
