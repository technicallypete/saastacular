#!/usr/bin/env bash
# Positive-ownership cleanup for one validation run (docs/foundation/OPERATIONS.md).
#
#   bash infra/ci/cleanup-owned.sh --registry DIR --mode finalize --out FILE
#       Gating; called by infra/ci/validate.sh (F2). Exit 0 only if every owned resource was removed and
#       verified and nothing was refused, unproven or failed; otherwise 18.
#   bash infra/ci/cleanup-owned.sh --after-script
#       GitLab after_script backstop (non-gating). It finds the registry from CI_PROJECT_DIR and
#       CI_JOB_ID, applies the same rules, prints the inventory, always removes the generated-secret
#       list, deletes the registry only when everything verified, and always exits 0.
#
# The registry (mode 700, never a symlink) holds owned.tsv (tab-separated, append-only rows: job, token,
# repo, work_root, label, project, rootenv, image <tag> <id>, standalone <id> <purpose>, and capture rows
# container <id>, network <id>, volume <name>), token, start-inventory.ids and cid/*.cid (container IDs
# written by the Docker CLI at creation).
#
# Ownership (removal needs every listed condition; anything else is kept and reported):
#   Compose container  absent from start-inventory.ids; its com.docker.compose.project label is a
#                      registered project; its com.docker.compose.project.working_dir equals repo, or lies
#                      under repo/, or under a work_root/ (Compose sets it to the project directory, by
#                      default the directory of the first -f file, for example <checkout>/infra/compose).
#                      A registered project with another working_dir is "refused".
#   standalone         absent from the start inventory; named by a standalone row or a cid/*.cid file.
#   network / volume   absent from the start inventory; named by a capture row, or attached to (mounted
#                      by) an owned container now.
#   image              absent from the start inventory; an image row matches both the tag and the ID that
#                      tag resolves to now. Removal is by that verified tag (never -f), so Docker deletes
#                      the ID only when no other reference remains.
#   root .env          its SHA-256 equals a rootenv row; otherwise "refused".
# Registered labels and <project>- image names only find candidates for the report ("unproven"); they
# never authorize a removal. A running container that carries a registered label but is not bound gets a
# bounded 60 s wait (ephemeral `docker run --rm` tools); if it remains it is "unproven".
# Never: prune, removal by label or name alone, any resource listed in the start inventory.
# Inventory rows: kind<TAB>id<TAB>action<TAB>reason, action in removed|refused|unproven|remaining|failed.
# A refused registry (symlink, wrong owner or mode, malformed row, token or job mismatch, missing start
# inventory) removes nothing.
set -euo pipefail

if (( BASH_VERSINFO[0] < 4 || (BASH_VERSINFO[0] == 4 && BASH_VERSINFO[1] < 4) )); then
  echo "cleanup: bash >= 4.4 is required" >&2
  exit 18
fi
shopt -s inherit_errexit nullglob

SELF_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)"
REPO="$(dirname -- "$(dirname -- "$SELF_DIR")")"
# shellcheck source=lib.sh
source "$SELF_DIR/lib.sh"
REPO_R="$(realpath -e -- "$REPO")"

EPHEMERAL_WAIT_S=60
MODE="" REG="" OUT="" AFTER=0 USAGE_ERR=0
while (( $# > 0 )); do
  case "$1" in
    --registry) if (( $# < 2 )); then USAGE_ERR=1; break; fi; REG="$2"; shift 2 ;;
    --mode) if (( $# < 2 )); then USAGE_ERR=1; break; fi; MODE="$2"; shift 2 ;;
    --out) if (( $# < 2 )); then USAGE_ERR=1; break; fi; OUT="$2"; shift 2 ;;
    --after-script) AFTER=1; shift ;;
    *) USAGE_ERR=1; break ;;
  esac
done
if (( AFTER )); then
  if (( USAGE_ERR )) || [[ -n "$REG$MODE$OUT" ]]; then
    echo "cleanup: usage: cleanup-owned.sh --after-script (non-gating; nothing done)"
    exit 0
  fi
elif (( USAGE_ERR )) || [[ "$MODE" != finalize || -z "$REG" || -z "$OUT" ]]; then
  echo "usage: cleanup-owned.sh --registry DIR --mode finalize --out FILE | --after-script" >&2
  exit 18
fi

# --- inventory rows ---------------------------------------------------------------------------------------
ROW_KEYS=()
declare -A ROW_ACTION=() ROW_REASON=()
row() { # KIND ID ACTION REASON (a later action for the same resource replaces the earlier one)
  local key="$1"$'\t'"$2"
  if [[ -z "${ROW_ACTION[$key]+x}" ]]; then ROW_KEYS+=("$key"); fi
  ROW_ACTION[$key]="$3"
  ROW_REASON[$key]="$4"
}
row_once() { # as row, but never replaces an existing row
  local key="$1"$'\t'"$2"
  if [[ -n "${ROW_ACTION[$key]+x}" ]]; then return 0; fi
  row "$@"
}

# --- Docker access that records failures instead of hiding them -------------------------------------------
QOUT=""
query() { # DESCRIPTION DOCKER-ARGS...: output in QOUT; a failure is recorded as a failed row
  local desc="$1"
  shift
  if QOUT="$(docker "$@" 2>/dev/null)"; then return 0; fi
  QOUT=""
  row query "$desc" failed "docker query failed"
  return 1
}
INS=""
inspect_obj() { # TYPE ID TEMPLATE: 0 found (output in INS), 1 absent, 2 error (recorded)
  local out
  if out="$(docker inspect --type "$1" -f "$3" "$2" 2>&1)"; then INS="$out"; return 0; fi
  INS=""
  case "$out" in
    *"No such"*|*"no such"*|*"not found"*) return 1 ;;
  esac
  row "$1" "$2" failed "docker inspect failed"
  return 2
}

# --- registry -------------------------------------------------------------------------------------------
declare -A START=() PROJ=() LABELS=() ROOTENV=() IMGROW=() SA=() CAPC=() CAPN=() CAPV=()
WORKROOTS=()
REFUSE=""
RESOURCE_ROWS=0

load_registry() {
  local line k a b rest n_job=0 n_tok=0 n_repo=0 tok_file="" tok_row="" job="" repo="" first=1 cf cid
  if [[ -L "$REG" ]]; then REFUSE="registry is a symlink"; return 1; fi
  if [[ ! -d "$REG" ]]; then REFUSE="registry is not a directory"; return 1; fi
  if [[ "$(stat -c %u -- "$REG")" != "$(id -u)" ]]; then REFUSE="registry is not owned by this user"; return 1; fi
  if [[ "$(stat -c %a -- "$REG")" != 700 ]]; then REFUSE="registry mode is not 700"; return 1; fi
  if [[ -L "$REG/owned.tsv" || ! -f "$REG/owned.tsv" ]]; then REFUSE="owned.tsv is missing"; return 1; fi
  if [[ -L "$REG/token" || ! -f "$REG/token" ]]; then REFUSE="token file is missing"; return 1; fi
  tok_file="$(cat -- "$REG/token")" || { REFUSE="token file is unreadable"; return 1; }
  if [[ ! "$tok_file" =~ ^[0-9a-f]{32}$ ]]; then REFUSE="token file is malformed"; return 1; fi
  while IFS= read -r line || [[ -n "$line" ]]; do
    if (( first )); then
      first=0
      if [[ "$line" != "$CI_REGISTRY_HEADER" ]]; then REFUSE="owned.tsv header is missing"; return 1; fi
      continue
    fi
    if [[ -z "$line" ]]; then continue; fi
    IFS=$'\t' read -r k a b rest < <(printf '%s\n' "$line")
    if [[ -n "${rest:-}" ]]; then REFUSE="owned.tsv has a row with too many fields"; return 1; fi
    case "$k" in
      job) if [[ -n "$b" || ! "$a" =~ ^([1-9][0-9]{0,18}|local-[0-9a-f]{12})$ ]]; then REFUSE="malformed job row"; return 1; fi
        job="$a"; n_job=$((n_job + 1)) ;;
      token) if [[ -n "$b" || ! "$a" =~ ^[0-9a-f]{32}$ ]]; then REFUSE="malformed token row"; return 1; fi
        tok_row="$a"; n_tok=$((n_tok + 1)) ;;
      repo) if [[ -n "$b" || "$a" != /* ]]; then REFUSE="malformed repo row"; return 1; fi
        repo="$a"; n_repo=$((n_repo + 1)) ;;
      work_root) if [[ -n "$b" || "$a" != /* || "$a" == / || "/$a/" == */../* ]]; then REFUSE="malformed work_root row"; return 1; fi
        WORKROOTS+=("${a%/}") ;;
      label) if [[ -n "$b" || ! "$a" =~ ^[A-Za-z0-9._-]{1,128}=[A-Za-z0-9._-]{1,128}$ ]]; then REFUSE="malformed label row"; return 1; fi
        LABELS[$a]=1 ;;
      project) if [[ -n "$b" || ! "$a" =~ ^[a-z0-9][a-z0-9_-]{0,62}$ ]]; then REFUSE="malformed project row"; return 1; fi
        PROJ[$a]=1; RESOURCE_ROWS=1 ;;
      rootenv) if [[ -n "$b" || ! "$a" =~ ^[0-9a-f]{64}$ ]]; then REFUSE="malformed rootenv row"; return 1; fi
        ROOTENV[$a]=1; RESOURCE_ROWS=1 ;;
      image) if [[ ! "$a" =~ ^[A-Za-z0-9][A-Za-z0-9._/:@+-]{0,254}$ || ! "$b" =~ ^sha256:[0-9a-f]{64}$ ]]; then REFUSE="malformed image row"; return 1; fi
        IMGROW["$(norm_tag "$a")|$b"]=1; RESOURCE_ROWS=1 ;;
      standalone) if [[ ! "$a" =~ ^[0-9a-f]{64}$ || ! "$b" =~ ^[A-Za-z0-9._-]{1,64}$ ]]; then REFUSE="malformed standalone row"; return 1; fi
        SA[$a]="$b"; RESOURCE_ROWS=1 ;;
      container) if [[ -n "$b" || ! "$a" =~ ^[0-9a-f]{64}$ ]]; then REFUSE="malformed container row"; return 1; fi
        CAPC[$a]=1; RESOURCE_ROWS=1 ;;
      network) if [[ -n "$b" || ! "$a" =~ ^[0-9a-f]{64}$ ]]; then REFUSE="malformed network row"; return 1; fi
        CAPN[$a]=1; RESOURCE_ROWS=1 ;;
      volume) if [[ -n "$b" || ! "$a" =~ ^[A-Za-z0-9][A-Za-z0-9_.-]{0,254}$ ]]; then REFUSE="malformed volume row"; return 1; fi
        CAPV[$a]=1; RESOURCE_ROWS=1 ;;
      *) REFUSE="owned.tsv has an unknown row kind"; return 1 ;;
    esac
  done <"$REG/owned.tsv"
  if (( first )); then REFUSE="owned.tsv is empty"; return 1; fi
  if (( n_job != 1 || n_tok != 1 || n_repo != 1 )); then REFUSE="owned.tsv needs exactly one job, token and repo row"; return 1; fi
  if [[ "$tok_row" != "$tok_file" ]]; then REFUSE="token mismatch"; return 1; fi
  if (( AFTER )) || [[ "${GITLAB_CI:-}" == true ]]; then
    if [[ "$job" != "${CI_JOB_ID:-}" ]]; then REFUSE="job mismatch"; return 1; fi
  elif [[ ! "$job" =~ ^local-[0-9a-f]{12}$ ]]; then
    REFUSE="job mismatch"; return 1
  fi
  if [[ "$repo" != "$REPO_R" ]]; then REFUSE="repo mismatch"; return 1; fi
  if [[ -L "$REG/cid" || ! -d "$REG/cid" ]]; then REFUSE="cid directory is missing"; return 1; fi
  for cf in "$REG"/cid/*.cid; do
    if [[ -L "$cf" || ! -f "$cf" ]]; then REFUSE="a cidfile is not a regular file"; return 1; fi
    cid="$(tr -d ' \n' <"$cf")" || { REFUSE="a cidfile is unreadable"; return 1; }
    if [[ -z "$cid" ]]; then continue; fi   # killed before the Docker CLI wrote the ID: binds nothing
    if [[ ! "$cid" =~ ^[0-9a-f]{64}$ ]]; then REFUSE="a cidfile is malformed"; return 1; fi
    if [[ -z "${SA[$cid]+x}" ]]; then SA[$cid]="cidfile"; fi
    RESOURCE_ROWS=1
  done
  if (( ${#LABELS[@]} > 0 )); then RESOURCE_ROWS=1; fi
  if [[ -e "$REG/start-inventory.ids" || -L "$REG/start-inventory.ids" ]]; then
    if [[ -L "$REG/start-inventory.ids" || ! -f "$REG/start-inventory.ids" ]]; then REFUSE="start inventory is not a regular file"; return 1; fi
    while IFS= read -r line || [[ -n "$line" ]]; do
      if [[ -z "$line" ]]; then continue; fi
      if [[ ! "$line" =~ ^(container|network|volume|image)\ [^\ ]+$ ]]; then REFUSE="start inventory is malformed"; return 1; fi
      START[$line]=1
    done <"$REG/start-inventory.ids"
  elif (( RESOURCE_ROWS )); then
    REFUSE="start inventory is missing"; return 1
  fi
  return 0
}

norm_tag() { # image reference with an implicit :latest made explicit
  local t="$1" last
  if [[ "$t" == *@* ]]; then printf '%s' "$t"; return 0; fi
  last="${t##*/}"
  if [[ "$last" != *:* ]]; then t="$t:latest"; fi
  printf '%s' "$t"
}

capture_row() { # KIND VALUE...: append a capture row (owned_capture inside cleanup), once
  local IFS=$'\t' line
  line="$*"
  if grep -qxF -- "$line" "$REG/owned.tsv"; then return 0; fi
  if printf '%s\n' "$line" >>"$REG/owned.tsv"; then return 0; fi
  row registry "$REG" failed "capture row not written"
}

# --- classification -----------------------------------------------------------------------------------
CLS="" CREASON="" CCOMPOSE=0
classify_container() { # ID -> CLS in owned|refused|unproven|start|absent|error
  local id="$1" p wd w rc
  CLS="" CREASON="" CCOMPOSE=0
  rc=0; inspect_obj container "$id" '{{index .Config.Labels "com.docker.compose.project"}}' || rc=$?
  if (( rc == 1 )); then CLS=absent; return 0; fi
  if (( rc == 2 )); then CLS=error; return 0; fi
  p="$INS"
  if [[ -n "${START["container $id"]+x}" ]]; then CLS=start; return 0; fi
  if [[ -n "$p" && -n "${PROJ[$p]+x}" ]]; then
    rc=0; inspect_obj container "$id" '{{index .Config.Labels "com.docker.compose.project.working_dir"}}' || rc=$?
    if (( rc == 1 )); then CLS=absent; return 0; fi
    if (( rc == 2 )); then CLS=error; return 0; fi
    wd="$INS"
    if [[ -n "$wd" && ( "$wd" == "$REPO_R" || "$wd" == "$REPO_R"/* ) ]]; then
      CLS=owned CREASON="registered Compose project, working_dir under the checkout" CCOMPOSE=1; return 0
    fi
    for w in "${WORKROOTS[@]}"; do
      if [[ -n "$wd" && "$wd" == "$w"/* ]]; then CLS=owned CREASON="registered Compose project, working_dir under work_root" CCOMPOSE=1; return 0; fi
    done
    if [[ -n "${SA[$id]+x}" ]]; then CLS=owned CREASON="bound at creation (${SA[$id]})"; return 0; fi
    CLS=refused CREASON="registered Compose project with a foreign working_dir"
    return 0
  fi
  if [[ -n "${SA[$id]+x}" ]]; then CLS=owned CREASON="bound at creation (${SA[$id]})"; return 0; fi
  if [[ -n "${CAPC[$id]+x}" ]]; then CLS=unproven CREASON="capture row without a registered project label now"; return 0; fi
  CLS=unproven CREASON="registered label only; a label is not deletion authority"
  return 0
}

container_candidates() { # -> CANDS (container IDs from projects, labels, bindings and capture rows)
  local p l id
  local -A seen=()
  CANDS=()
  for p in "${!PROJ[@]}"; do
    if query "ps project" ps -aq --no-trunc --filter "label=com.docker.compose.project=$p"; then
      for id in $QOUT; do if [[ -z "${seen[$id]+x}" ]]; then seen[$id]=1; CANDS+=("$id"); fi; done
    fi
  done
  for l in "${!LABELS[@]}"; do
    if query "ps label" ps -aq --no-trunc --filter "label=$l"; then
      for id in $QOUT; do if [[ -z "${seen[$id]+x}" ]]; then seen[$id]=1; CANDS+=("$id"); fi; done
    fi
  done
  for id in "${!SA[@]}" "${!CAPC[@]}"; do
    if [[ -z "${seen[$id]+x}" ]]; then seen[$id]=1; CANDS+=("$id"); fi
  done
}

# Running, unbound containers that carry a registered label (ephemeral `docker run --rm` tools) get a
# bounded wait to disappear before anything is decided.
ephemeral_wait() {
  local l id deadline rc auto
  local -a waitlist=() next=()
  if (( ${#LABELS[@]} == 0 )); then return 0; fi
  for l in "${!LABELS[@]}"; do
    if query "ps running label" ps -q --no-trunc --filter "label=$l"; then
      for id in $QOUT; do
        classify_container "$id"
        if [[ "$CLS" == unproven ]]; then waitlist+=("$id"); fi
      done
    fi
  done
  if (( ${#waitlist[@]} == 0 )); then return 0; fi
  echo "cleanup: waiting up to ${EPHEMERAL_WAIT_S}s for ${#waitlist[@]} running label-only container(s) to exit"
  deadline=$((SECONDS + EPHEMERAL_WAIT_S))
  while (( ${#waitlist[@]} > 0 && SECONDS < deadline )); do
    sleep 1
    next=()
    for id in "${waitlist[@]}"; do
      rc=0; inspect_obj container "$id" '{{.State.Running}} {{.HostConfig.AutoRemove}}' || rc=$?
      if (( rc != 0 )); then continue; fi
      auto="${INS#* }"
      if [[ "${INS%% *}" == true || "$auto" == true ]]; then next+=("$id"); fi
    done
    waitlist=("${next[@]}")
  done
}

declare -A OWN_C=() OWN_N=() OWN_V=() OWN_I=() ATT_N=() ATT_V=()
OWN_C_ORDER=()
OWN_ENV=0

# evaluate PASS: decide (classify, capture, report refused/unproven) or verify (anything still owned
# is "remaining").
evaluate() {
  local pass="$1" id p x tag iid cur key rc sha
  local -A ncand=() vcand=()
  container_candidates
  for id in "${CANDS[@]}"; do
    classify_container "$id"
    case "$CLS" in
      owned)
        if [[ "$pass" == verify ]]; then row container "$id" remaining "still present after removal"; continue; fi
        if [[ -z "${OWN_C[$id]+x}" ]]; then OWN_C[$id]="$CREASON"; OWN_C_ORDER+=("$id"); fi
        rc=0; inspect_obj container "$id" '{{range $k, $v := .NetworkSettings.Networks}}{{$v.NetworkID}} {{end}}' || rc=$?
        if (( rc == 0 )); then for x in $INS; do if [[ -z "${START["network $x"]+x}" ]]; then ATT_N[$x]=1; fi; done; fi
        rc=0; inspect_obj container "$id" '{{range .Mounts}}{{if eq .Type "volume"}}{{.Name}} {{end}}{{end}}' || rc=$?
        if (( rc == 0 )); then for x in $INS; do if [[ -z "${START["volume $x"]+x}" ]]; then ATT_V[$x]=1; fi; done; fi
        if (( CCOMPOSE )); then
          capture_row container "$id"
          for x in "${!ATT_N[@]}"; do capture_row network "$x"; done
          for x in "${!ATT_V[@]}"; do capture_row volume "$x"; done
          rc=0; inspect_obj container "$id" '{{.Config.Image}} {{.Image}}' || rc=$?
          if (( rc == 0 )); then
            tag="${INS%% *}" iid="${INS##* }"
            if [[ -n "$tag" && "$iid" =~ ^sha256:[0-9a-f]{64}$ && -z "${START["image $iid"]+x}" ]]; then
              IMGROW["$(norm_tag "$tag")|$iid"]=1
              capture_row image "$tag" "$iid"
            fi
          fi
        fi ;;
      refused) row_once container "$id" refused "$CREASON" ;;
      unproven) row_once container "$id" unproven "$CREASON" ;;
      *) ;;
    esac
  done
  # Networks and volumes: capture rows, current attachments of owned containers, registered project labels.
  for x in "${!CAPN[@]}" "${!ATT_N[@]}"; do ncand[$x]=1; done
  for x in "${!CAPV[@]}" "${!ATT_V[@]}"; do vcand[$x]=1; done
  for p in "${!PROJ[@]}"; do
    if query "network ls project" network ls -q --no-trunc --filter "label=com.docker.compose.project=$p"; then
      for x in $QOUT; do ncand[$x]=1; done
    fi
    if query "volume ls project" volume ls -q --filter "label=com.docker.compose.project=$p"; then
      for x in $QOUT; do vcand[$x]=1; done
    fi
  done
  for x in "${!ncand[@]}"; do
    rc=0; inspect_obj network "$x" '{{.Id}}' || rc=$?
    if (( rc != 0 )); then continue; fi
    id="$INS"
    if [[ -n "${START["network $id"]+x}" ]]; then continue; fi
    if [[ -n "${CAPN[$id]+x}" || -n "${ATT_N[$id]+x}" ]]; then
      if [[ "$pass" == verify ]]; then row network "$id" remaining "still present after removal"; else OWN_N[$id]=1; fi
    else
      row_once network "$id" unproven "registered project label without a capture row or an owned attachment"
    fi
  done
  for x in "${!vcand[@]}"; do
    rc=0; inspect_obj volume "$x" '{{.Name}}' || rc=$?
    if (( rc != 0 )); then continue; fi
    id="$INS"
    if [[ -n "${START["volume $id"]+x}" ]]; then continue; fi
    if [[ -n "${CAPV[$id]+x}" || -n "${ATT_V[$id]+x}" ]]; then
      if [[ "$pass" == verify ]]; then row volume "$id" remaining "still present after removal"; else OWN_V[$id]=1; fi
    else
      row_once volume "$id" unproven "registered project label without a capture row or an owned mount"
    fi
  done
  # Images: a build or capture row must match the tag and the ID it resolves to now.
  local -A owned_tags=()
  for key in "${!IMGROW[@]}"; do
    tag="${key%|*}" iid="${key##*|}"
    rc=0; inspect_obj image "$tag" '{{.Id}}' || rc=$?
    if (( rc != 0 )); then continue; fi
    cur="$INS"
    if [[ "$cur" != "$iid" ]]; then continue; fi
    if [[ -n "${START["image $cur"]+x}" ]]; then continue; fi
    owned_tags[$tag]=1
    if [[ "$pass" == verify ]]; then row image "$tag|$iid" remaining "still present after removal"; else OWN_I[$tag]="$iid"; fi
  done
  if (( ${#PROJ[@]} > 0 )); then
    if query "image ls" image ls --no-trunc --format '{{.Repository}}:{{.Tag}} {{.ID}}'; then
      while read -r tag iid; do
        if [[ -z "$tag" || "$tag" == *"<none>"* ]]; then continue; fi
        for p in "${!PROJ[@]}"; do
          if [[ "$tag" == "$p"-* && -z "${owned_tags[$tag]+x}" && -z "${START["image $iid"]+x}" ]]; then
            row_once image "$tag|$iid" unproven "registered project name without a build or capture row"
          fi
        done
      done < <(printf '%s\n' "$QOUT")
    fi
  fi
  # Root .env, bound by the hash the harness recorded.
  if (( ${#ROOTENV[@]} > 0 )) && [[ -e "$REPO_R/.env" || -L "$REPO_R/.env" ]]; then
    if [[ -L "$REPO_R/.env" || ! -f "$REPO_R/.env" ]]; then
      row_once env "$REPO_R/.env" refused "root .env is not a regular file"
    else
      sha="$(sha256sum <"$REPO_R/.env" | cut -c1-64)" || sha=""
      if [[ -n "$sha" && -n "${ROOTENV[$sha]+x}" ]]; then
        if [[ "$pass" == verify ]]; then row env "$REPO_R/.env" remaining "still present after removal"; else OWN_ENV=1; fi
      else
        row_once env "$REPO_R/.env" refused "root .env hash differs from the rootenv row"
      fi
    fi
  fi
}

remove_owned() {
  local id rc tag cur
  for id in "${OWN_C_ORDER[@]}"; do
    rc=0; docker rm -f -v "$id" >/dev/null 2>&1 || rc=$?
    if (( rc == 0 )); then row container "$id" removed "${OWN_C[$id]}"; else row container "$id" failed "docker rm exited $rc"; fi
  done
  for id in "${!OWN_N[@]}"; do
    rc=0; docker network rm "$id" >/dev/null 2>&1 || rc=$?
    if (( rc == 0 )); then row network "$id" removed "capture row or owned attachment"; else row network "$id" failed "docker network rm exited $rc"; fi
  done
  for id in "${!OWN_V[@]}"; do
    rc=0; inspect_obj volume "$id" '{{.Name}}' || rc=$?
    if (( rc == 1 )); then row volume "$id" removed "removed with its owned container"; continue; fi
    rc=0; docker volume rm "$id" >/dev/null 2>&1 || rc=$?
    if (( rc == 0 )); then row volume "$id" removed "capture row or owned mount"; else row volume "$id" failed "docker volume rm exited $rc"; fi
  done
  for tag in "${!OWN_I[@]}"; do
    rc=0; inspect_obj image "$tag" '{{.Id}}' || rc=$?
    if (( rc == 1 )); then continue; fi
    cur="$INS"
    if (( rc != 0 )) || [[ "$cur" != "${OWN_I[$tag]}" ]]; then
      row_once image "$tag|${OWN_I[$tag]}" unproven "the tag no longer resolves to the recorded ID"
      continue
    fi
    rc=0; docker image rm "$tag" >/dev/null 2>&1 || rc=$?
    if (( rc == 0 )); then row image "$tag|$cur" removed "build or capture row matching tag and ID"; else row image "$tag|$cur" failed "docker image rm exited $rc"; fi
  done
  if (( OWN_ENV )); then
    if rm -f -- "$REPO_R/.env"; then row env "$REPO_R/.env" removed "hash equals the rootenv row"; else row env "$REPO_R/.env" failed "rm failed"; fi
  fi
}

# --- output -------------------------------------------------------------------------------------------
write_inventory() { # FILE
  local key
  : >"$1" || return 1
  for key in "${ROW_KEYS[@]}"; do
    printf '%s\t%s\t%s\n' "$key" "${ROW_ACTION[$key]}" "${ROW_REASON[$key]}" >>"$1" || return 1
  done
}
summary_counts() { # -> COUNT_<action>
  local key a
  C_removed=0 C_refused=0 C_unproven=0 C_remaining=0 C_failed=0
  for key in "${ROW_KEYS[@]}"; do
    a="${ROW_ACTION[$key]}"
    case "$a" in
      removed) C_removed=$((C_removed + 1)) ;;
      refused) C_refused=$((C_refused + 1)) ;;
      unproven) C_unproven=$((C_unproven + 1)) ;;
      remaining) C_remaining=$((C_remaining + 1)) ;;
      *) C_failed=$((C_failed + 1)) ;;
    esac
  done
}

# --- main ---------------------------------------------------------------------------------------------
if (( AFTER )); then
  if [[ ! "${CI_JOB_ID:-}" =~ ^[1-9][0-9]{0,18}$ || -z "${CI_PROJECT_DIR:-}" ]]; then
    echo "cleanup: CI_PROJECT_DIR or CI_JOB_ID is unavailable; nothing done (non-gating)"
    exit 0
  fi
  REG="$(ci_registry_dir "$CI_PROJECT_DIR" "$CI_JOB_ID")" || { echo "cleanup: cannot derive the registry path (non-gating)"; exit 0; }
  if [[ ! -e "$REG" && ! -L "$REG" ]]; then
    echo "cleanup: no registry (finalization completed)"
    exit 0
  fi
  if [[ "$(realpath -m -- "$CI_PROJECT_DIR")" != "$REPO_R" ]]; then
    echo "cleanup: CI_PROJECT_DIR is not this checkout; refused (non-gating)"
    exit 0
  fi
else
  REG="${REG%/}"
  if [[ -L "$REG" ]]; then :; else REG="$(realpath -m -- "$REG")"; fi
fi

if load_registry; then
  if (( RESOURCE_ROWS )); then
    if docker version --format '{{.Server.Version}}' >/dev/null 2>&1; then
      ephemeral_wait
      evaluate decide
      remove_owned
      OWN_C=() OWN_N=() OWN_V=() OWN_I=() OWN_ENV=0
      evaluate verify
    else
      row docker daemon failed "Docker is unreachable; nothing could be verified"
    fi
  fi
else
  row registry "$REG" refused "$REFUSE; nothing was removed"
fi
summary_counts
STATUS=0
if (( C_refused + C_unproven + C_remaining + C_failed > 0 )); then STATUS=18; fi
SUMMARY="cleanup: removed=$C_removed refused=$C_refused unproven=$C_unproven remaining=$C_remaining failed=$C_failed"

if (( AFTER == 0 )); then
  if write_inventory "$OUT"; then :; else echo "cleanup: cannot write the inventory" >&2; exit 18; fi
  echo "$SUMMARY"
  exit "$STATUS"
fi

# after_script: print the inventory (IDs, names and fixed reasons only), always drop the secret list, and
# delete the registry only when everything verified. Never gating.
for key in "${ROW_KEYS[@]}"; do
  printf 'cleanup: %s\t%s\t%s\n' "$key" "${ROW_ACTION[$key]}" "${ROW_REASON[$key]}"
done
echo "$SUMMARY"
if [[ -n "$REFUSE" ]]; then
  echo "cleanup: registry refused ($REFUSE); it is left untouched for the owner"
  exit 0
fi
write_inventory "$REG/cleanup-inventory.tsv" || echo "cleanup: cannot write cleanup-inventory.tsv into the registry"
rm -f -- "$REG/secrets" || echo "cleanup: cannot remove the generated-secret list"
if (( STATUS == 0 )); then
  if rm -rf -- "$REG"; then echo "cleanup: registry removed"; else echo "cleanup: registry could not be removed: $REG"; fi
else
  echo "cleanup: registry kept for the owner: $REG"
fi
exit 0
