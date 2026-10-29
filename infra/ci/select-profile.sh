#!/usr/bin/env bash
# Deterministic CI profile selection (G003 CI revision Plan r2 §2.4-§2.5; G051 Plan r3 §3.3; docs/foundation/OPERATIONS.md).
# It decides the pipeline profile from the pipeline context and, where the context requires it, from the
# changed paths against a validated comparison base, using the reviewed map infra/ci/profile-paths.tsv.
# G051: registered feature areas select the affected profile through a fail-closed file allow-list and an
# over-approximating dependency closure over the comparison base (old edges) and HEAD (new edges).
# Bash and git only; no Docker, no network. It writes KEY=VALUE lines (imported by validate.sh through
# state_import with fixed expressions) and a selection table (status, class, escaped path, area, via).
#
#   bash infra/ci/select-profile.sh --job fast|gate|local [--requested fast|affected|integration|full]
#     [--base REV | --compare OLD] --tested-content VALUE --map FILE --out FILE --tsv FILE
#
# --base and --compare are local options of --requested affected only (G051 rule 9).
# Exit: 0 with a selection; 16 on a usage error, a malformed map, or a refused job/context combination.
set -euo pipefail
shopt -s inherit_errexit

JOB="" REQ="" TC="" MAP="" OUT="" TSV="" XBASE="" XCMP="" NBASE=0 NCMP=0
usage() { echo "usage: select-profile.sh --job fast|gate|local [--requested PROFILE] [--base REV | --compare OLD] --tested-content VALUE --map FILE --out FILE --tsv FILE" >&2; exit 16; }
while (( $# > 0 )); do
  case "$1" in
    --job) JOB="${2:-}"; shift 2 || usage ;;
    --requested) REQ="${2:-}"; shift 2 || usage ;;
    --base) XBASE="${2:-}" NBASE=$((NBASE + 1)); shift 2 || usage ;;
    --compare) XCMP="${2:-}" NCMP=$((NCMP + 1)); shift 2 || usage ;;
    --tested-content) TC="${2:-}"; shift 2 || usage ;;
    --map) MAP="${2:-}"; shift 2 || usage ;;
    --out) OUT="${2:-}"; shift 2 || usage ;;
    --tsv) TSV="${2:-}"; shift 2 || usage ;;
    *) usage ;;
  esac
done
if [[ ! "$JOB" =~ ^(fast|gate|local)$ || -z "$TC" || -z "$MAP" || -z "$OUT" || -z "$TSV" ]]; then usage; fi
if [[ "$JOB" == local && ! "$REQ" =~ ^(fast|affected|integration|full)$ ]]; then usage; fi
if [[ "$JOB" != local && -n "$REQ" ]]; then usage; fi
# G051 rule 9: --base/--compare only for a local affected request, at most one of them, a plain revision.
if (( NBASE + NCMP > 0 )); then
  if [[ "$JOB" != local || "$REQ" != affected ]] || (( NBASE + NCMP > 1 )); then
    echo "select-profile: --base and --compare are exclusive options of a local affected run" >&2; exit 16
  fi
  if [[ ! "$XBASE$XCMP" =~ ^[A-Za-z0-9_][A-Za-z0-9._/@{}~^-]{0,255}$ ]]; then echo "select-profile: malformed revision" >&2; exit 16; fi
fi

SELF_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)"
REPO="$(dirname -- "$(dirname -- "$SELF_DIR")")"
if [[ ! -f "$SELF_DIR/lib.sh" ]]; then echo "select-profile: infra/ci/lib.sh is missing" >&2; exit 16; fi
# shellcheck source=lib.sh
source "$SELF_DIR/lib.sh"
g() { git -C "$REPO" -c core.fsmonitor=false --no-optional-locks --literal-pathspecs "$@"; }
gg() { g -c grep.lineNumber=false -c grep.column=false -c grep.fullName=false -c grep.patternType=basic \
  -c grep.extendedRegexp=false -c color.grep=never grep --no-color --no-recurse-submodules "$@"; }

# Ranks (G051): fast 1 < affected 2 < integration 3 < full 4; any doubt resolves to the higher rank.
rank() { case "$1" in fast) echo 1 ;; affected) echo 2 ;; integration) echo 3 ;; full) echo 4 ;; *) echo 0 ;; esac; }
class_profile() { case "$1" in docs|agent) echo fast ;; runtime) echo integration ;; foundation) echo full ;; *) echo full ;; esac; }
esc() { # escaped path for the table: bytes outside [A-Za-z0-9._/+-] become %XX
  local s="$1" out="" i c
  for ((i = 0; i < ${#s}; i++)); do
    c="${s:i:1}"
    if [[ "$c" =~ [A-Za-z0-9._/+-] ]]; then out+="$c"; else out+="$(printf '%%%02X' "'$c")"; fi
  done
  printf '%s' "$out"
}
unsafe_path() { if [[ -z "$1" || "$1" =~ [[:cntrl:]] || "$1" == *..* || "$1" == /* ]]; then return 0; fi; return 1; }
is_link() { if [[ "$1" == 120000 || "$1" == 160000 ]]; then return 0; fi; return 1; }

# --- the reviewed map ------------------------------------------------------------------------------------
# G051 area rows: area-web|area-core|area-intel, kind dir, an allowed root; the area id is KIND:NAME.
declare -A FILES=() DIRS=() DIR_AREA=() AREAS_REG=() ROOT_AREA=()
ROOT_LIST=()
area_root_id() { # CLASS PATH: prints the area id of an allowed area root, or fails
  case "$1" in
    area-web) if [[ "$2" =~ ^web/(app/\(features\)|tests/browser/features|tests/fixtures/features)/([a-z][a-z0-9-]{0,31})/$ ]]; then printf 'web:%s' "${BASH_REMATCH[2]}"; return 0; fi ;;
    area-core) if [[ "$2" =~ ^core/internal/([a-z][a-z0-9]{0,31})/$ ]]; then printf 'core:%s' "${BASH_REMATCH[1]}"; return 0; fi ;;
    area-intel) if [[ "$2" =~ ^intelligence/(src|tests)/([a-z][a-z0-9]{0,31})/$ ]]; then printf 'intel:%s' "${BASH_REMATCH[2]}"; return 0; fi ;;
  esac
  return 1
}
if [[ -L "$MAP" || ! -f "$MAP" ]]; then echo "select-profile: the path map is missing" >&2; exit 16; fi
while IFS= read -r line || [[ -n "$line" ]]; do
  if [[ -z "$line" || "$line" == '#'* ]]; then continue; fi
  if [[ ! "$line" =~ ^(docs|agent|runtime|foundation|area-web|area-core|area-intel)$'\t'(file|dir)$'\t'([^[:space:]]+)$ ]]; then
    echo "select-profile: malformed map row" >&2; exit 16
  fi
  cls="${BASH_REMATCH[1]}" kind="${BASH_REMATCH[2]}" p="${BASH_REMATCH[3]}"
  if [[ "$p" == /* || "$p" == *..* || "$p" == *//* || "$p" == ./* ]]; then echo "select-profile: unsafe map path" >&2; exit 16; fi
  if [[ "$kind" == dir && "$p" != */ ]] || [[ "$kind" == file && "$p" == */ ]]; then echo "select-profile: map kind/path mismatch" >&2; exit 16; fi
  if [[ "$cls" == area-* ]]; then
    if [[ "$kind" != dir ]]; then echo "select-profile: an area row must be a directory" >&2; exit 16; fi
    id="$(area_root_id "$cls" "$p")" || { echo "select-profile: area row outside the allowed roots or with a bad name" >&2; exit 16; }
    DIR_AREA[$p]="$id" AREAS_REG[$id]=1 ROOT_AREA[$p]="$id"
    ROOT_LIST+=("$p")
  fi
  if [[ "$kind" == file ]]; then
    if [[ -n "${FILES[$p]+x}" ]]; then echo "select-profile: duplicate map row" >&2; exit 16; fi
    FILES[$p]="$cls"
  else
    if [[ -n "${DIRS[$p]+x}" ]]; then echo "select-profile: duplicate map row" >&2; exit 16; fi
    DIRS[$p]="$cls"
  fi
done <"$MAP"
if (( ${#FILES[@]} + ${#DIRS[@]} == 0 )); then echo "select-profile: the path map is empty" >&2; exit 16; fi
# G051: no other row may lie inside an area root (registration never narrows another row).
for a in "${ROOT_LIST[@]}"; do
  for p in "${!FILES[@]}" "${!DIRS[@]}"; do
    if [[ "$p" != "$a" && "$p" == "$a"* ]]; then echo "select-profile: a map row overlaps an area root" >&2; exit 16; fi
  done
done

CP_CLASS="" CP_AREA="" CP_DIR=""
classify() { # PATH: CP_CLASS (docs|agent|runtime|foundation|area-*|unknown), CP_AREA (area id or "") and CP_DIR
  local p="$1" best="" d
  CP_AREA="" CP_DIR=""
  if [[ -n "${FILES[$p]+x}" ]]; then CP_CLASS="${FILES[$p]}"; return 0; fi
  # G049: a project-subtree Claude skill (anywhere below a project directory, including docs/ and management/, but
  # not the root agent directories) is foundation, ahead of the Markdown documentation bypass below.
  if [[ "$p" != .claude/* && "$p" != .agents/* && "$p" == */.claude/skills/* ]]; then CP_CLASS=foundation; return 0; fi
  # Owner-approved documentation bypass, shared with workflow admission.
  # Planning-input JSON and exact Goal scope allow-lists are supervisor metadata; other controls stay classified.
  if [[ "$p" == docs/*.md || "$p" == management/*.md || ( "$p" != */* && "$p" == *.md ) ||
        ( "$p" == management/v0/plans/G[0-9][0-9][0-9]-planning-inputs*.json && "${p#management/v0/plans/}" != */* ) ||
        "$p" == management/v0/results/G[0-9][0-9][0-9]-scope-allowlist.txt ]]; then
    CP_CLASS=docs; return 0
  fi
  for d in "${!DIRS[@]}"; do
    if [[ "$p" == "$d"* && ${#d} -gt ${#best} ]]; then best="$d"; fi
  done
  if [[ -z "$best" ]]; then CP_CLASS=unknown; return 0; fi
  CP_CLASS="${DIRS[$best]}" CP_DIR="$best" CP_AREA="${DIR_AREA[$best]:-}"
  return 0
}

AFK=""
area_file_kind() { # ID ROOT PATH LINK: AFK = ordinary|route|unclassified (G051 rule 2, the per-kind allow-list)
  local kind="${1%%:*}" rel="/${3#"$2"}" b="${3##*/}"
  AFK=unclassified
  if [[ "$4" == 1 ]]; then return 0; fi
  # Scripts, build, dependency, environment and tool configuration (and any dot-file) are never area-ordinary.
  case "$b" in
    .*|Dockerfile*|Makefile*|requirements*|*.config.*|package*.json|tsconfig*.json|components.json|go.mod|go.sum|go.work*|pyproject.toml|*.sh|*.sql) return 0 ;;
  esac
  case "$kind" in
    web)
      case "$b" in
        route.ts|route.js|route.tsx) AFK=route ;;
        *.ts|*.tsx|*.css|*.json|*.svg|*.png|*.webp|*.jpg|*.md|*.txt) AFK=ordinary ;;
      esac ;;
    core)
      if [[ "$rel" == */testdata/* ]]; then AFK=ordinary; return 0; fi
      case "$b" in *.go|*.md) AFK=ordinary ;; esac ;;
    intel)
      case "$b" in
        conftest.py) if [[ "$3" != intelligence/src/* ]]; then AFK=ordinary; fi ;;
        *.py|*.md) AFK=ordinary ;;
        *.json|*.yaml|*.txt) if [[ "$rel" == */testdata/* || "$rel" == */fixtures/* ]]; then AFK=ordinary; fi ;;
      esac ;;
  esac
  return 0
}
in_area() { # PATH ID: PATH lies in one of the roots of area ID
  local r
  for r in "${ROOT_LIST[@]}"; do
    if [[ "${ROOT_AREA[$r]}" == "$2" && "$1" == "$r"* ]]; then return 0; fi
  done
  return 1
}
lands_in() { # TARGET ID: the resolved TARGET is a root of area ID or lies below it
  local r
  for r in "${ROOT_LIST[@]}"; do
    if [[ "${ROOT_AREA[$r]}" == "$2" && "$1/" == "$r"* ]]; then return 0; fi
  done
  return 1
}
area_roots() { # ID: AR = the roots of area ID
  local r
  AR=()
  for r in "${ROOT_LIST[@]}"; do if [[ "${ROOT_AREA[$r]}" == "$1" ]]; then AR+=("$r"); fi; done
}
NJ=""
norm_join() { # DIR SPEC: NJ = DIR/SPEC with "." and ".." resolved; 1 when it escapes the repository root
  local rest="$1/$2" seg
  local -a out=()
  while [[ -n "$rest" ]]; do
    seg="${rest%%/*}"
    if [[ "$rest" == */* ]]; then rest="${rest#*/}"; else rest=""; fi
    case "$seg" in
      ''|.) ;;
      ..) if (( ${#out[@]} == 0 )); then return 1; fi; unset 'out[-1]' ;;
      *) out+=("$seg") ;;
    esac
  done
  local IFS=/
  NJ="${out[*]}"
  return 0
}
dir_of() { if [[ "$1" == */* ]]; then printf '%s' "${1%/*}"; fi; }

# --- context first ------------------------------------------------------------------------------------------
PIPE="" REASON="" BASE=none MODE=diff cand=""
tmpd="$(mktemp -d "${TMPDIR:-/tmp}/select-profile.XXXXXX")" || { echo "select-profile: mktemp failed" >&2; exit 16; }
trap 'rm -rf -- "$tmpd"' EXIT
: >"$TSV"
case "$TC" in
  local)
    if [[ "$JOB" != local ]]; then echo "select-profile: a local run must use --job local" >&2; exit 16; fi
    if [[ "$REQ" != affected ]]; then PIPE="$REQ" REASON=local-explicit
    elif (( NBASE == 1 )); then cand="$(g rev-parse --verify -q "$XBASE^{commit}" 2>/dev/null)" || cand=""
    elif (( NCMP == 1 )); then MODE=compare; cand="$(g rev-parse --verify -q "$XCMP^{commit}" 2>/dev/null)" || cand=""
    else PIPE=full REASON=missing-base; fi ;;
  branch-push)
    b="${CI_COMMIT_BRANCH:-}"
    if [[ "$b" != staging && "$b" != "${CI_DEFAULT_BRANCH:-}" ]]; then PIPE=fast REASON=feature-push; fi ;;
  branch-schedule) PIPE=full REASON=schedule ;;
  branch-web) PIPE=full REASON=web-explicit ;;
  branch-api) PIPE=full REASON=api-explicit ;;
  mr-merged-result) PIPE=full REASON=merged-result ;;
  # ONEOFF RC qualification: a verified RC-promotion merge request (validate.sh's attribution) is always the complete
  # full profile: no diff (documentation-only candidate content included), every present registered area, and the CI
  # meta-tests even with CI_CONTROLS_FORCE=0.
  mr-rc-promotion) PIPE=full REASON=rc-promotion ;;
  mr-source) ;;
  *) echo "select-profile: unsupported tested content" >&2; exit 16 ;;
esac
if [[ "$JOB" == local && "$TC" != local ]]; then echo "select-profile: --job local outside a local run" >&2; exit 16; fi
target="$(g rev-parse --verify 'HEAD^{commit}' 2>/dev/null)" || { echo "select-profile: no HEAD commit" >&2; exit 16; }

# --- area presence (cached per area and revision) ---------------------------------------------------------------
declare -A PRES=()
AS=0
area_state() { # ID REV: AS = 1 when the area has a tracked path at REV, else 0; 1 on a read failure
  local key="$1@$2"
  if [[ -z "${PRES[$key]+x}" ]]; then
    area_roots "$1"
    if g ls-tree -r --name-only -z "$2" -- "${AR[@]}" >"$tmpd/ls" 2>/dev/null; then :; else return 1; fi
    if [[ -s "$tmpd/ls" ]]; then PRES[$key]=1; else PRES[$key]=0; fi
  fi
  AS="${PRES[$key]}"
  return 0
}

# --- classify by paths (detached MR; staging or default-branch push; local affected) ---------------------------
base_check() { # CANDIDATE TARGET: prints the failure reason, or nothing when the base is usable
  if [[ ! "$1" =~ ^[0-9a-f]{40}$ || "$1" =~ ^0{40}$ ]]; then echo missing-base; return 0; fi
  if g cat-file -e "$1^{commit}" 2>/dev/null; then :; else echo missing-base; return 0; fi
  if [[ "$MODE" == compare ]]; then return 0; fi
  if [[ "$1" == "$2" ]]; then echo degenerate-base; return 0; fi
  if g merge-base --is-ancestor "$1" "$2" 2>/dev/null; then :; else echo degenerate-base; return 0; fi
  return 0
}
best=0 FASTS=0 CLOSURE_FAIL=0
declare -A TOUCHED=() SEL=() REMOVED=() CONS=() CONS_LINK=()
REC_ST=() REC_P=() REC_LINK=() REC_PART=()
emit() { # STATUS CLASS PATH AREA VIA
  printf '%s\t%s\t%s\t%s\t%s\n' "$1" "$2" "$(esc "$3")" "$4" "$5" >>"$TSV"
}
raise() { # PROFILE REASON: the first path reaching the highest rank gives the reason
  if (( $(rank "$1") > best )); then best="$(rank "$1")" PIPE="$1" REASON="$2"; fi
}
partner_ordinary() { # IDX: a rename/copy partner is an area-ordinary path (no partner: true)
  local j="${REC_PART[$1]}" q
  if (( j < 0 )); then return 0; fi
  q="${REC_P[j]}"
  if unsafe_path "$q"; then return 1; fi
  classify "$q"
  if [[ -z "$CP_AREA" ]]; then return 1; fi
  area_file_kind "$CP_AREA" "$CP_DIR" "$q" "${REC_LINK[j]}"
  if [[ "$AFK" == ordinary ]]; then return 0; fi
  return 1
}
if [[ -z "$PIPE" ]]; then
  if [[ "$TC" == mr-source ]]; then cand="${CI_MERGE_REQUEST_DIFF_BASE_SHA:-}"
  elif [[ "$TC" == branch-push ]]; then cand="${CI_COMMIT_BEFORE_SHA:-}"; fi
  why="$(base_check "$cand" "$target")"
  if [[ -n "$why" ]]; then
    PIPE=full REASON="$why"
  else
    BASE="$cand" PIPE=fast REASON=no-change
    parts=() sorted=()
    # The diff and the sort go through files so their exit statuses are checked (a process substitution's
    # status is lost); any failure selects full ("diff-failed") without a table. G051: --raw adds the modes
    # (symbolic links and submodules are never area-ordinary).
    if g diff -z --raw --no-abbrev -M --no-ext-diff --no-textconv "$cand" "$target" >"$tmpd/diff" 2>/dev/null; then
      mapfile -d '' -t parts <"$tmpd/diff"
    else
      BASE=none PIPE=full REASON=diff-failed best=4
    fi
    # Records: a header ":OLDMODE NEWMODE OLDID NEWID STATUS", then one path (two for renames and copies).
    i=0 n=0
    while (( i < ${#parts[@]} )); do
      hdr="${parts[i]}"; i=$((i + 1))
      if [[ -z "$hdr" ]]; then continue; fi
      if [[ ! "$hdr" =~ ^:([0-7]{6})\ ([0-7]{6})\ [0-9a-f]+\ [0-9a-f]+\ ([A-Z][0-9]*)$ ]]; then
        BASE=none PIPE=full REASON=diff-failed best=4 n=0; break
      fi
      om="${BASH_REMATCH[1]}" nm="${BASH_REMATCH[2]}" st="${BASH_REMATCH[3]}"
      case "${st:0:1}" in
        R|C)
          REC_ST[n]="$st" REC_P[n]="${parts[i]:-}" REC_LINK[n]=0 REC_PART[n]=$((n + 1))
          if is_link "$om"; then REC_LINK[n]=1; fi
          REC_ST[n+1]="$st" REC_P[n+1]="${parts[i+1]:-}" REC_LINK[n+1]=0 REC_PART[n+1]="$n"
          if is_link "$nm"; then REC_LINK[n+1]=1; fi
          n=$((n + 2)) i=$((i + 2)) ;;
        *)
          REC_ST[n]="$st" REC_P[n]="${parts[i]:-}" REC_LINK[n]=0 REC_PART[n]=-1
          if is_link "$om" || is_link "$nm"; then REC_LINK[n]=1; fi
          n=$((n + 1)) i=$((i + 1)) ;;
      esac
    done
    # Byte order of the path, so the reason is the first path that reaches the maximum.
    if (( n > 0 )); then
      if for ((k = 0; k < n; k++)); do printf '%s\x1f%s\0' "$k" "${REC_P[k]}"; done | LC_ALL=C sort -z -t $'\x1f' -k2 -s >"$tmpd/sorted"; then
        mapfile -d '' -t sorted <"$tmpd/sorted"
      else
        BASE=none PIPE=full REASON=diff-failed best=4
      fi
    fi
    for r in "${sorted[@]}"; do
      if [[ -z "$r" ]]; then continue; fi
      k="${r%%$'\x1f'*}"
      st="${REC_ST[k]}" p="${REC_P[k]}" area=-
      if unsafe_path "$p"; then cls=unsafe prof=full why=unsafe-path
      else
        classify "$p"
        case "$CP_CLASS" in
          # Adding, deleting or renaming documentation must not promote a mixed commit to full. A move across
          # this boundary still has its non-documentation/control path inspected independently.
          docs) cls=docs prof=fast why=docs-only FASTS=1 ;;
          # G049: root agent paths (.claude/, .agents/) select fast for additions, deletions and renames alike; a
          # move out of them has its destination classified independently, as for documentation.
          agent) cls=agent prof=fast why=agent-only FASTS=1 ;;
          area-*)
            # G051 rules 2-3: an area-ordinary A/M selects its area; a route handler selects integration; D/R/C
            # selects the area only when both sides are area-ordinary; anything else in an area root is full.
            cls="$CP_CLASS" area="$CP_AREA" TOUCHED[$CP_AREA]=1
            area_file_kind "$CP_AREA" "$CP_DIR" "$p" "${REC_LINK[k]}"
            if [[ "$AFK" == unclassified ]]; then prof=full why=area-unclassified
            elif [[ "${st:0:1}" =~ ^[AM]$ ]]; then
              if [[ "$AFK" == route ]]; then prof=integration why=area-route-handler; else prof=affected why=area; fi
            elif [[ "${st:0:1}" =~ ^[DRC]$ && "$AFK" == ordinary ]] && partner_ordinary "$k"; then
              prof=affected why=area
              if area_state "$area" "$target"; then if [[ "$AS" == 0 ]]; then why="area-removed"; fi; else CLOSURE_FAIL=1; fi
            else cls=change prof=full why=deletion-or-rename; fi ;;
          *)
            case "${st:0:1}" in
              A|M)
                cls="$CP_CLASS"
                if [[ "$cls" == unknown ]]; then prof=full why=unknown-path; else prof="$(class_profile "$cls")" why="$cls"; fi ;;
              *) cls=change prof=full why=deletion-or-rename ;;
            esac ;;
        esac
      fi
      emit "${st:0:1}" "$cls" "$p" "$area" -
      raise "$prof" "$why"
    done
    if (( CLOSURE_FAIL == 1 )); then raise full closure-failed; fi
  fi
fi

# --- G051 rule 4-7: the dependency closure over BASE (old edges) and HEAD (new edges) --------------------------
QUEUE=()
grep_l() { # OUTVAR REV SCOPE ARGS...: append to array OUTVAR the paths at REV below SCOPE whose content matches
  local -n gl_out="$1"
  local rev="$2" scope="$3" e rc=0
  local -a gl_raw=()
  shift 3
  gg -l -z "$@" "$rev" -- "$scope" >"$tmpd/grep" 2>"$tmpd/grep.err" || rc=$?
  # git grep reports an unreadable object on stderr with status 1, so any diagnostic is a read failure.
  if (( rc > 1 )) || [[ -s "$tmpd/grep.err" ]]; then return 1; fi
  mapfile -d '' -t gl_raw <"$tmpd/grep"
  for e in "${gl_raw[@]}"; do
    if [[ "$e" != "$rev:"* ]]; then return 1; fi
    gl_out+=("${e#"$rev":}")
  done
  return 0
}
grep_o() { # REV SCOPE ARGS...: GO_PATH/GO_MATCH = each matching part of the text files at REV below SCOPE
  local rev="$1" scope="$2" rc=0 cur t j
  local -a tok=()
  shift 2
  GO_PATH=() GO_MATCH=()
  gg -o -z -I "$@" "$rev" -- "$scope" >"$tmpd/grep" 2>"$tmpd/grep.err" || rc=$?
  if (( rc > 1 )) || [[ -s "$tmpd/grep.err" ]]; then return 1; fi
  mapfile -d '' -t tok <"$tmpd/grep"
  if (( ${#tok[@]} == 0 )); then return 0; fi
  # "REV:path NUL match LF" per match: the text after the first LF of a token is the next path.
  cur="${tok[0]}"
  for ((j = 1; j < ${#tok[@]}; j++)); do
    t="${tok[j]}"
    if [[ "$t" != *$'\n'* || "$cur" != "$rev:"* ]]; then return 1; fi
    GO_PATH+=("${cur#"$rev":}") GO_MATCH+=("${t%%$'\n'*}")
    cur="${t#*$'\n'}"
  done
  if [[ -n "$cur" ]]; then return 1; fi
  return 0
}
path_link() { # REV PATH: PL = 1 when PATH is a symbolic link or submodule at REV
  local -a e=()
  PL=0
  if g ls-tree -z "$1" -- "$2" >"$tmpd/lt" 2>/dev/null; then :; else return 1; fi
  mapfile -d '' -t e <"$tmpd/lt"
  if (( ${#e[@]} > 0 )) && is_link "${e[0]%% *}"; then PL=1; fi
  return 0
}
add_consumer() { # PATH VIA REV [LINK]: record a referencing file of area VIA (outside it); its area joins
  local p="$1" via="$2" rev="$3" link="${4:-}" key
  if [[ "$p" == infra/ci/profile-paths.tsv ]] || in_area "$p" "$via"; then return 0; fi
  key="$p"$'\x1f'"$via"
  CONS[$key]=1
  if unsafe_path "$p"; then return 0; fi
  classify "$p"
  if [[ -z "$CP_AREA" ]]; then return 0; fi
  if [[ -z "$link" ]]; then path_link "$rev" "$p" || return 1; link="$PL"; fi
  if [[ "$link" == 1 ]]; then CONS_LINK[$key]=1; fi
  if [[ -z "${SEL[$CP_AREA]+x}" ]]; then SEL[$CP_AREA]=1; QUEUE+=("$CP_AREA"); fi
  return 0
}
# Per-revision scans (once each): resolved relative references (web specifiers, Python relative imports,
# symbolic links) and the ambiguous forms of Plan r3 §3.3 rule 4.
declare -A SCANNED=()
PY_FROM_RE='^from[[:space:]]+(\.+)([A-Za-z_][A-Za-z0-9_.]*)?[[:space:]]*import([[:space:](].*)?$'
RL_REV=() RL_FILE=() RL_TGT=() RL_LINK=() AMB_FILE=()
rel_add() { RL_REV+=("$1") RL_FILE+=("$2") RL_TGT+=("$3") RL_LINK+=("$4"); }
web_root_below() { # DIR: a web area root lies strictly below DIR
  local r
  for r in "${ROOT_LIST[@]}"; do if [[ "$r" == web/* && "$r" == "$1"/* ]]; then return 0; fi; done
  return 1
}
scan_rev() { # REV
  local rev="$1" j f m s fd pre tpl dots mod rest ir nb item nm e sha tgt
  local -a hits=() ents=() bases=() names=()
  if [[ -n "${SCANNED[$rev]+x}" ]]; then return 0; fi
  SCANNED[$rev]=1
  # Web: every quoted relative string (a superset of import, export-from, import(), require(), CSS @import,
  # new URL()) and unquoted CSS url(); resolved against the file's directory. Escaping web/ is ambiguous.
  grep_o "$rev" web/ -E -e "'\\.\\.?(/[^']*)?'" -e '"\.\.?(/[^"]*)?"' -e $'\x60\\.\\.?(/[^\x60]*)?\x60' \
    -e 'url\([[:space:]]*\.\.?/[^)'"'"'"[:space:]]*' || return 1
  for ((j = 0; j < ${#GO_PATH[@]}; j++)); do
    f="${GO_PATH[j]}" m="${GO_MATCH[j]}" tpl=0
    if [[ "$m" == url\(* ]]; then s="${m#url(}"; s="${s#"${s%%[![:space:]]*}"}"; else s="${m:1:${#m}-2}"; fi
    if [[ "$s" == *'${'* ]]; then
      pre="${s%%\$\{*}" tpl=1
      if [[ "$pre" == */* ]]; then s="${pre%/*}"; else s=.; fi
    fi
    fd="$(dir_of "$f")"
    if norm_join "$fd" "$s" && [[ "$NJ" == web || "$NJ" == web/* ]]; then
      if (( tpl == 1 )) && web_root_below "$NJ"; then AMB_FILE+=("$f"); fi
      rel_add "$rev" "$f" "$NJ" ""
    else AMB_FILE+=("$f"); fi
  done
  # Web ambiguity: a non-literal or template specifier in import()/require(); an unresolvable (features) reference.
  grep_l hits "$rev" web/ -I -E -e '(^|[^A-Za-z0-9_$])(import|require)[[:space:]]*\([[:space:]]*($|[^'"'"'"`[:space:]])' \
    -e '(^|[^A-Za-z0-9_$])(import|require)[[:space:]]*\([[:space:]]*`[^`]*\$\{' || return 1
  grep_o "$rev" web/ -E -e '\(features\)[^(]{0,34}' || return 1
  for ((j = 0; j < ${#GO_PATH[@]}; j++)); do
    if [[ ! "${GO_MATCH[j]}" =~ ^\(features\)/[a-z][a-z0-9-]{0,31}([^a-z0-9-]|$) ]]; then hits+=("${GO_PATH[j]}"); fi
  done
  # Python ambiguity: dynamic imports, exec, sys.path manipulation.
  grep_l hits "$rev" intelligence/ -I -E -e 'import_module[[:space:]]*\(' -e '__import__[[:space:]]*\(' \
    -e '(^|[^A-Za-z0-9_])exec[[:space:]]*\(' -e 'sys\.path' || return 1
  AMB_FILE+=("${hits[@]}")
  # Python relative imports, resolved against the file's package path within its import root.
  grep_o "$rev" intelligence/ -E -e '(^|[;:])[[:space:]]*from[[:space:]]+\.[^;#]*' || return 1
  for ((j = 0; j < ${#GO_PATH[@]}; j++)); do
    f="${GO_PATH[j]}" m="${GO_MATCH[j]}"
    m="${m#[;:]}"; m="${m#"${m%%[![:space:]]*}"}"; m="${m%"${m##*[![:space:]]}"}"
    if [[ ! "$m" =~ $PY_FROM_RE ]]; then AMB_FILE+=("$f"); continue; fi
    dots="${BASH_REMATCH[1]}" mod="${BASH_REMATCH[2]}" rest="${BASH_REMATCH[3]}"
    rest="${rest#"${rest%%[![:space:]]*}"}"
    case "$f" in intelligence/src/*) ir=intelligence/src ;; intelligence/tests/*) ir=intelligence/tests ;; *) ir=intelligence ;; esac
    fd="$(dir_of "$f")" s=.
    for ((nb = 1; nb < ${#dots}; nb++)); do s+=/..; done
    if norm_join "$fd" "$s" && [[ "$NJ" == "$ir" || "$NJ" == "$ir"/* ]]; then :; else AMB_FILE+=("$f"); continue; fi
    if [[ -n "$mod" ]]; then rel_add "$rev" "$f" "$NJ/${mod//./\/}" ""; continue; fi
    names=() bases=("$NJ")
    if [[ "$rest" == '*' ]]; then rel_add "$rev" "$f" "$NJ" ""; continue; fi
    if [[ "$rest" == \(*\) ]]; then rest="${rest:1:${#rest}-2}"; fi
    nm=ok
    if [[ -z "$rest" || "$rest" == *[\(\)\\]* ]]; then nm=unknown; fi
    while [[ "$nm" == ok && -n "$rest" ]]; do
      item="${rest%%,*}"
      if [[ "$rest" == *,* ]]; then rest="${rest#*,}"; else rest=""; fi
      item="${item#"${item%%[![:space:]]*}"}"; item="${item%"${item##*[![:space:]]}"}"
      if [[ -z "$item" ]]; then continue; fi
      if [[ "$item" =~ ^([A-Za-z_][A-Za-z0-9_]*)([[:space:]]+as[[:space:]]+[A-Za-z_][A-Za-z0-9_]*)?$ ]]; then names+=("${BASH_REMATCH[1]}"); else nm=unknown; fi
    done
    if [[ "$nm" == unknown ]]; then
      # Unknown imported names can be any area directly below the import root.
      if [[ "${bases[0]}" == "$ir" ]]; then AMB_FILE+=("$f"); else rel_add "$rev" "$f" "${bases[0]}" ""; fi
      continue
    fi
    for e in "${names[@]}"; do rel_add "$rev" "$f" "${bases[0]}/$e" ""; done
  done
  # Symbolic links anywhere: the resolved target (absolute or escaping targets are not repository references).
  if g ls-tree -r -z --full-tree "$rev" >"$tmpd/tree" 2>/dev/null; then :; else return 1; fi
  mapfile -d '' -t ents <"$tmpd/tree"
  for e in "${ents[@]}"; do
    if [[ "$e" != 120000\ blob\ * ]]; then continue; fi
    sha="${e#120000 blob }" sha="${sha%%$'\t'*}" f="${e#*$'\t'}"
    tgt="$(g cat-file blob "$sha" 2>/dev/null)" || return 1
    if [[ "$tgt" == /* ]]; then continue; fi
    if norm_join "$(dir_of "$f")" "$tgt"; then rel_add "$rev" "$f" "$NJ" 1; fi
  done
  return 0
}
find_refs() { # ID REV: the referencing files of area ID at REV (Plan r3 §3.3 rule 4)
  local id="$1" rev="$2" kind="${1%%:*}" n="${1#*:}" e j
  local -a hits=()
  case "$kind" in
    # Go: the import path (any form) or any text naming the package directory, subpackages included.
    core) grep_l hits "$rev" . -E -e "core/internal/$n([^a-z0-9]|\$)" || return 1 ;;
    # Web alias: (features)/<n> or features/<n> in any file.
    web) grep_l hits "$rev" . -E -e "\\(features\\)/$n([^a-z0-9-]|\$)" -e "features/$n([^a-z0-9-]|\$)" || return 1 ;;
    # Python: import <n>[.sub], from <n>[.sub] import, in intelligence/; and any text naming the area roots.
    intel)
      grep_l hits "$rev" intelligence/ -E \
        -e "(^|[;:])[[:space:]]*(from[[:space:]]+$n([.[:space:]]|\$)|import[[:space:]]+([^#]*[,[:space:]])?$n([.[:space:],]|\$))" || return 1
      grep_l hits "$rev" . -E -e "intelligence/(src|tests)/$n([^A-Za-z0-9_]|\$)" || return 1 ;;
  esac
  for e in "${hits[@]}"; do add_consumer "$e" "$id" "$rev" || return 1; done
  scan_rev "$rev" || return 1
  for ((j = 0; j < ${#RL_REV[@]}; j++)); do
    if [[ "${RL_REV[j]}" == "$rev" ]] && lands_in "${RL_TGT[j]}" "$id"; then
      add_consumer "${RL_FILE[j]}" "$id" "$rev" "${RL_LINK[j]}" || return 1
    fi
  done
  return 0
}
stack_refs() { # ID: rule 6, a touched core or intelligence area adds every registered web area with a *.stack.spec.ts
  local id="$1" w rev f
  local -a fl=()
  for w in "${AREA_IDS[@]}"; do
    if [[ "$w" != web:* ]]; then continue; fi
    area_roots "$w"
    for rev in "${TREES[@]}"; do
      if g ls-tree -r --name-only -z "$rev" -- "${AR[@]}" >"$tmpd/ls" 2>/dev/null; then :; else return 1; fi
      mapfile -d '' -t fl <"$tmpd/ls"
      for f in "${fl[@]}"; do
        if [[ "$f" == *.stack.spec.ts ]]; then add_consumer "$f" "$id" "$rev" || return 1; fi
      done
    done
  done
  return 0
}
closure() { # the fixpoint over the selected areas; 1 on any read failure
  local qi=0 id rev
  QUEUE=("${SELIDS[@]}")
  while (( qi < ${#QUEUE[@]} )); do
    id="${QUEUE[qi]}"; qi=$((qi + 1))
    for rev in "${TREES[@]}"; do find_refs "$id" "$rev" || return 1; done
    if [[ "$id" == core:* || "$id" == intel:* ]]; then stack_refs "$id" || return 1; fi
  done
  return 0
}
sort_z() { # ITEMS...: SZ = the items in byte order (they must not contain NUL)
  SZ=()
  if (( $# == 0 )); then return 0; fi
  printf '%s\0' "$@" | LC_ALL=C sort -z >"$tmpd/sz" || return 1
  mapfile -d '' -t SZ <"$tmpd/sz"
  return 0
}
SZ=() AREA_IDS=() SELIDS=()
sort_z "${!AREAS_REG[@]}" || { echo "select-profile: sort failed" >&2; exit 16; }
AREA_IDS=("${SZ[@]}")
TREES=("$target")
if [[ "$BASE" != none ]]; then TREES=("$BASE" "$target"); fi

if (( ${#TOUCHED[@]} > 0 )); then
  for id in "${!TOUCHED[@]}"; do SEL[$id]=1; done
  # Rule 7: an area with files at BASE and none at HEAD is removed (its feature obligation becomes its consumers').
  for id in "${!TOUCHED[@]}"; do
    if area_state "$id" "$BASE"; then pb="$AS"; else raise full closure-failed; continue; fi
    if area_state "$id" "$target"; then :; else raise full closure-failed; continue; fi
    if [[ "$pb" == 1 && "$AS" == 0 ]]; then REMOVED[$id]=1; fi
  done
fi
if (( ${#TOUCHED[@]} > 0 && best < 4 )); then
  sort_z "${!SEL[@]}" || { echo "select-profile: sort failed" >&2; exit 16; }
  SELIDS=("${SZ[@]}")
  if closure; then
    # Ambiguous forms outside the selected areas give full when any web or intelligence area is selected.
    amb=0
    for id in "${!SEL[@]}"; do if [[ "$id" == web:* || "$id" == intel:* ]]; then amb=1; fi; done
    if (( amb == 1 )); then
      for f in "${AMB_FILE[@]}"; do
        inside=0
        for id in "${!SEL[@]}"; do if in_area "$f" "$id"; then inside=1; fi; done
        if (( inside == 0 )); then CONS["$f"$'\x1f'ambiguous]=1; fi
      done
    fi
    sort_z "${!CONS[@]}" || { echo "select-profile: sort failed" >&2; exit 16; }
    for key in "${SZ[@]}"; do
      p="${key%$'\x1f'*}" via="${key##*$'\x1f'}" area=-
      if [[ -n "${REMOVED[$via]+x}" ]]; then why="area-removed"; else why="area-imported"; fi
      if unsafe_path "$p"; then emit + unsafe "$p" - "$via"; raise full unsafe-path; continue; fi
      classify "$p"
      cls="$CP_CLASS" prof=none
      if [[ -n "$CP_AREA" ]]; then area="$CP_AREA"; fi
      if [[ "$via" == ambiguous ]]; then prof=full why=area-ambiguous
      elif [[ "$p" == core/cmd/* ]]; then prof=full
      else
        case "$cls" in
          docs) ;;
          agent) FASTS=1 ;;
          runtime) prof=integration ;;
          area-*)
            area_file_kind "$CP_AREA" "$CP_DIR" "$p" "${CONS_LINK[$key]:-0}"
            case "$AFK" in ordinary) prof=affected ;; route) prof=integration ;; *) prof=full ;; esac ;;
          *) prof=full ;;
        esac
      fi
      emit + "$cls" "$p" "$area" "$via"
      if [[ "$prof" != none ]]; then raise "$prof" "$why"; fi
    done
  else
    raise full closure-failed
  fi
fi

# --- suites and areas (G051 rule 8; INTERFACES §1) ---------------------------------------------------------------
has_it() { # ID: HI = 1 when the core area has a //go:build integration file at HEAD
  local -a hl=()
  local f
  HI=0
  area_roots "$1"
  for f in "${AR[@]}"; do
    grep_l hl "$target" "$f" -E -e '^//go:build[[:space:]](.*[^!A-Za-z0-9_.])?integration([^A-Za-z0-9_.]|$)' || return 1
  done
  for f in "${hl[@]}"; do if [[ "$f" == *.go ]]; then HI=1; fi; done
  return 0
}
LIST="" AREAS_OUT=""
compute_suites() { # PIPE: LIST (suites before canonical ordering) and AREAS_OUT; 1 on a read failure
  local id
  local -a ids=()
  LIST=""
  case "$1" in
    fast) LIST="$CI_SUITES_fast" ;;
    affected)
      LIST="lint-base provenance"
      sort_z "${!SEL[@]}" || return 1
      ids=("${SZ[@]}")
      for id in "${ids[@]}"; do
        area_state "$id" "$target" || return 1
        case "$id" in
          web:*) LIST+=" lint-web unit-web"; if [[ "$AS" == 1 ]]; then LIST+=" feature-web"; fi ;;
          core:*)
            LIST+=" lint-core unit-core"
            if [[ "$AS" == 1 ]]; then has_it "$id" || return 1; if [[ "$HI" == 1 ]]; then LIST+=" feature-core"; fi; fi ;;
          intel:*) LIST+=" contracts unit-intel" ;;
        esac
      done
      if (( FASTS == 1 )); then LIST+=" $CI_SUITES_fast"; fi ;;
    integration|full)
      if [[ "$1" == integration ]]; then LIST="$CI_SUITES_integration"; else LIST="$CI_SUITES_full"; fi
      for id in "${AREA_IDS[@]}"; do
        area_state "$id" "$target" || return 1
        if [[ "$AS" == 0 ]]; then continue; fi
        ids+=("$id")
        if [[ "$id" == web:* ]]; then LIST+=" feature-web"; fi
        if [[ "$id" == core:* ]]; then has_it "$id" || return 1; if [[ "$HI" == 1 ]]; then LIST+=" feature-core"; fi; fi
      done ;;
  esac
  AREAS_OUT=none
  if (( ${#ids[@]} > 0 )); then local IFS=,; AREAS_OUT="${ids[*]}"; fi
  return 0
}
if compute_suites "$PIPE"; then :
elif [[ "$PIPE" == affected ]]; then
  PIPE=full REASON=closure-failed
  compute_suites full || { echo "select-profile: cannot read the areas at HEAD" >&2; exit 16; }
else echo "select-profile: cannot read the areas at HEAD" >&2; exit 16; fi
if [[ "$PIPE" == full ]] && ! ci_controls_required "$REPO" "$BASE" "$target" "$TC"; then
  LIST="$(ci_suites_minus "$LIST" ci-controls)" || exit 16
  echo "select-profile: INFO CI meta-tests not selected (no impacted CI/harness/control dependency)"
fi
PSUITES="$(ci_suites_canonical "$LIST")" || { echo "select-profile: unknown suite" >&2; exit 16; }

# Rule 7 INFO: a registered area with no tracked files at BASE and HEAD adds no obligation.
for id in "${AREA_IDS[@]}"; do
  if area_state "$id" "$target" && [[ "$AS" == 0 ]]; then
    if [[ "$BASE" == none ]] || { area_state "$id" "$BASE" && [[ "$AS" == 0 ]]; }; then
      echo "select-profile: INFO registered-absent $id (no tracked files at the comparison base or HEAD; no obligation)"
    fi
  fi
done
sort_z "${!REMOVED[@]}" || { echo "select-profile: sort failed" >&2; exit 16; }
for id in "${SZ[@]}"; do echo "select-profile: INFO area-removed $id (consumer obligations replace its feature suite)"; done

# --- per-job outcome -----------------------------------------------------------------------------------------
if [[ "$JOB" == gate ]] && [[ "$TC" == local || "$REASON" == feature-push ]]; then
  echo "select-profile: a gate job is not part of this pipeline context" >&2; exit 16
fi
part="$(ci_job_partition "$JOB" "$PIPE" "$PSUITES")" || { echo "select-profile: no partition" >&2; exit 16; }
EXEC="${part%%$'\t'*}" part="${part#*$'\t'}"
ESUITES="${part%%$'\t'*}" part="${part#*$'\t'}"
COVERAGE="${part%%$'\t'*}"
blob_at() { # PATH: the blob id of PATH at HEAD, or "none"
  local -a e=()
  if g ls-tree -z "$target" -- "$1" >"$tmpd/lt" 2>/dev/null; then :; else return 1; fi
  mapfile -d '' -t e <"$tmpd/lt"
  if (( ${#e[@]} == 0 )); then printf none; return 0; fi
  if [[ "${e[0]}" =~ ^[0-7]{6}\ blob\ ([0-9a-f]{40})$'\t' ]]; then printf '%s' "${BASH_REMATCH[1]}"; return 0; fi
  return 1
}
tree="$(g rev-parse --verify "$target^{tree}" 2>/dev/null)" || { echo "select-profile: no HEAD tree" >&2; exit 16; }
bmap="$(blob_at infra/ci/profile-paths.tsv)" && bsel="$(blob_at infra/ci/select-profile.sh)" &&
  breg="$(blob_at infra/validation/required-checks.tsv)" || { echo "select-profile: cannot read HEAD" >&2; exit 16; }
digest="$(printf 'head=%s\ntree=%s\nmap=%s\nselector=%s\nregistry=%s\nbase=%s\nareas=%s\npipeline=%s\nsuites=%s\n' \
  "$target" "$tree" "$bmap" "$bsel" "$breg" "$BASE" "$AREAS_OUT" "$PIPE" "$PSUITES" | sha256sum)" || { echo "select-profile: digest failed" >&2; exit 16; }
digest="${digest%% *}"
# Fixed-shape outputs (INTERFACES §3); anything else is an internal error.
if [[ ! "$PIPE" =~ ^(fast|affected|integration|full)$ || ! "$EXEC" =~ ^(fast|affected|integration|full|none)$ ||
      ! "$REASON" =~ ^[a-z][a-z-]{1,31}$ || ! "$BASE" =~ ^([0-9a-f]{40}|none)$ ||
      ! "$AREAS_OUT" =~ ^(none|(web|core|intel):[a-z][a-z0-9-]{0,31}(,(web|core|intel):[a-z][a-z0-9-]{0,31})*)$ ||
      ! "$PSUITES" =~ ^[a-z][a-z-]*(\ [a-z][a-z-]*)*$ || ! "$ESUITES" =~ ^(none|[a-z][a-z-]*(\ [a-z][a-z-]*)*)$ ||
      ! "$COVERAGE" =~ ^(complete|partial|none)$ || ! "$digest" =~ ^[0-9a-f]{64}$ ]]; then
  echo "select-profile: internal error: malformed selection" >&2; exit 16
fi
(set -C; printf 'PIPELINE_PROFILE=%s\nEXECUTED_PROFILE=%s\nSELECTION_REASON=%s\nCOMPARISON_BASE=%s\nAREAS=%s\nPIPELINE_SUITES=%s\nEXECUTED_SUITES=%s\nCOVERAGE=%s\nSELECTION_DIGEST=%s\n' \
  "$PIPE" "$EXEC" "$REASON" "$BASE" "$AREAS_OUT" "$PSUITES" "$ESUITES" "$COVERAGE" "$digest" >"$OUT") || { echo "select-profile: cannot write the selection" >&2; exit 16; }
exit 0
