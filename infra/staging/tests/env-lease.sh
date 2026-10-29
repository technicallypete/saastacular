#!/usr/bin/env bash
# ENV-LEASE (G072 Plan §10.5 proofs (a)-(g), §10.3 IAM-model cycles): the environment lease that excludes concurrent
# G070 mutators, local or CI, over fake tools only (tests/fakes: gcloud with a fake Cloud Storage enforcing generation
# preconditions and the permission model of infra/terraform/bootstrap/ci-meta-permissions.tsv, docker/Terraform, jq,
# curl, date). No network, Docker or cloud; everything under --work.
#
#   bash infra/staging/tests/env-lease.sh --work DIR --report DIR [--only CASE[,CASE]...]
#
# Cases: baseline-fixture (the vendored G070 baseline copies under tests/fixtures/baseline/ exist and match their pinned
# SHA-256; otherwise the run fails closed before any other case), parity-<command>-<free|absent> (f: the G070 baseline
# staging.sh/lib.sh and the current ones give
# identical stdout, exit codes and non-metadata call logs, with a free lease and with an absent bucket), dryrun-parity,
# ci-holds-<command> (b), concurrent-local (a local mutator holds while another is refused), release-then-proceed and
# break-glass (c), lease-held (e), presence-403/presence-error (g), local-blocks-ci-deploy/-stop and ci-lease-blocks-ci
# (a), ci-break-expired (d: an expired CI lease broken with a record; an expired local lease never broken by CI),
# iam-cycle (start/stop x3 and a repeated stop under the IAM model) and iam-no-delete (the r0 defect reproduced: without
# the exact-key delete grant the first stop fails at the current.tsv overwrite) - these six through tests/ci-adapter.sh
# and imported - and the control-* mutation cases (each protected assertion fails against a deliberately broken copy;
# control-baseline-missing and control-baseline-tampered prove the fixture check fails closed).
# Ledger REPORT/summary.tsv; last line RESULT PASS|FAIL cases=N failed=M; exit 0 all pass, 1 any failure, 64 usage.
set -uo pipefail
# shellcheck source=fakes/fixture.sh
source "$(dirname "${BASH_SOURCE[0]}")/fakes/fixture.sh"
# The G070 baseline staging.sh/lib.sh: byte-exact historical copies vendored as fixtures (no Git history needed), each
# pinned by its content SHA-256.
BASELINE_DIR=$FX_SRC/tests/fixtures/baseline
BASELINE_PINS=("staging.sh 7c319382c48e8b3f089bee07ec33746c364567bac61131cf9edb775f47393ee7"
  "lib.sh 9d11fe404de15c362137b7457caa779debf53196a40bbd5b968082f6983cb06b")
# baseline_check DIR: prints why DIR's baseline copies are unusable (missing, not a regular file or altered); empty = ok.
baseline_check() {
  local e f
  for e in "${BASELINE_PINS[@]}"; do
    f=$1/${e%% *}
    if [[ ! -f $f || -L $f ]]; then printf '%s missing; ' "${e%% *}"
    elif [[ "$(sha256sum <"$f" | cut -d' ' -f1)" != "${e##* }" ]]; then printf '%s altered (SHA-256 differs from its pin); ' "${e%% *}"; fi
  done
}

fx_args "$@"
FX_PERMS="$(fx_perms_source)"
FAKE_ROOT=$W/fake
R=$W/repo
fx_repo "$R"
FAKE_REPO=$R
C=$FX_BASE
CONF=$W/priv/staging.conf
fx_conf "$CONF" "$W/evidence" "$W/priv"
mkdir -p "$W/base" "$W/new"
BASELINE_WHY="$(baseline_check "$BASELINE_DIR")"
if [[ -n $BASELINE_WHY ]]; then
  record baseline-fixture 0 "tests/fixtures/baseline: ${BASELINE_WHY}fail closed, no case run"
  fx_finish; exit 1
fi
record baseline-fixture 1 "tests/fixtures/baseline: staging.sh and lib.sh present and equal to their pinned SHA-256"
cp "$BASELINE_DIR/staging.sh" "$W/base/staging.sh"
cp "$BASELINE_DIR/lib.sh" "$W/base/lib.sh"
# Fixture negatives: a missing and a tampered copy are each refused by the same check.
if selected control-baseline-missing; then
  mkdir -p "$W/bl-missing"; cp "$BASELINE_DIR/staging.sh" "$W/bl-missing/"
  control control-baseline-missing "lib.sh removed from a copy of the fixture directory" "$(baseline_check "$W/bl-missing")"
fi
if selected control-baseline-tampered; then
  mkdir -p "$W/bl-tampered"; cp "$BASELINE_DIR/staging.sh" "$BASELINE_DIR/lib.sh" "$W/bl-tampered/"
  printf '#\n' >>"$W/bl-tampered/lib.sh"
  control control-baseline-tampered "one comment line appended to lib.sh (forbid() unchanged)" "$(baseline_check "$W/bl-tampered")"
fi
cp "$FX_SRC/staging.sh" "$FX_SRC/lib.sh" "$FX_SRC/meta-lib.sh" "$W/new/"
COMMANDS=(provision deploy start stop reset retire)
declare -A ARGS=([provision]="--candidate $C --yes" [deploy]="--candidate $C" [start]="" [stop]="" [reset]="--confirm $FX_ENV"
  [retire]="--yes")

# use_scripts DIR: the staging.sh/lib.sh/meta-lib.sh copy under test, installed as the fixture's control code.
use_scripts() { cp "$1"/*.sh "$R/infra/staging/"; }

# world [absent]: a fresh fake world plus a fresh evidence/private state with C's pushed manifest.
world() {
  fx_state "${1:-}"
  rm -rf "$W/evidence" "$W/priv/plans" "$W/priv/tf-data"
  mkdir -p "$W/evidence/deploy"
  fx_manifest "$W/evidence/deploy/$C.tsv" "$C"
}

# st NAME COMMAND [VAR=VALUE...]: staging.sh COMMAND (its fixture arguments) with the given extra environment.
st() {
  local name=$1 cmd=$2 a
  shift 2
  read -ra a <<<"${ARGS[$cmd]:-}"
  fx_run "$name" "CLOUDFLARE_API_TOKEN=$FX_CF_TOKEN" "$@" -- bash "$R/infra/staging/staging.sh" "$cmd" "${a[@]}" --config "$CONF"
}

nonmeta() { calls_nonmeta | sed "s#$R#<repo>#g"; }

# ---- (f) legacy parity ----------------------------------------------------------------------------------------------
parity() {
  local cmd=$1 mode=$2
  local c="parity-$cmd-$mode" ok=1 why="" brc nrc bout nout
  world "$( [[ $mode == absent ]] && echo absent)"
  use_scripts "$W/base"; st "$c-base" "$cmd"; brc=$RC; nonmeta >"$W/runs/$c-base.calls"; bout=$OUT
  keep "$c" base
  world "$( [[ $mode == absent ]] && echo absent)"
  use_scripts "$W/new"; st "$c-new" "$cmd"; nrc=$RC; nonmeta >"$W/runs/$c-new.calls"; nout=$OUT
  keep "$c" new
  [[ $brc == 0 ]] || { ok=0; why+="baseline exit $brc (fixture world broken); "; }
  [[ $brc == "$nrc" ]] || { ok=0; why+="exit $brc vs $nrc; "; }
  cmp -s "$bout" "$nout" || { ok=0; why+="stdout differs ($(diff "$bout" "$nout" | head -3 | tr '\n' ' ')); "; }
  cmp -s "$W/runs/$c-base.calls" "$W/runs/$c-new.calls" ||
    { ok=0; why+="non-metadata calls differ ($(diff "$W/runs/$c-base.calls" "$W/runs/$c-new.calls" | head -3 | tr '\n' ' ')); "; }
  if [[ $mode == absent ]]; then
    [[ "$(calls_all | grep -c $'^gcloud\t.\tstorage ')" == 1 ]] || { ok=0; why+="absent bucket: not exactly one metadata call; "; }
    [[ "$(grep -c '^staging: ci-meta:' "$nout" "$ERR" | awk -F: '{s += $NF} END {print s}')" == 1 ]] || { ok=0; why+="absent bucket: not exactly one ci-meta log line; "; }
  else
    grep -q $'^gcloud\tW\tstorage cp .* gs://'"$FX_BUCKET"'/lease --if-generation-match=0' "$FAKE_ROOT/calls.log" || { ok=0; why+="no lease create; "; }
    [[ "$(fx_gen lease)" == 0 ]] || { ok=0; why+="lease still held after exit; "; }
    if [[ $cmd == deploy || $cmd == provision ]]; then
      [[ "$(fx_cat fence.tsv | awk -F'\t' -v c="$C" '$1 == "fence" && $4 == c' | nlines)" == 1 ]] || { ok=0; why+="no fence entry for $C; "; }
      fx_cat "schema/$C.tsv" | grep -q $'^24\t' || { ok=0; why+="no schema/$C.tsv; "; }
      # the fence entry is written before the release upload
      awk -F'\t' '/fence.tsv --if-generation-match/ {f = NR} /--no-same-owner -xf -/ {u = NR} END {exit !(f && u && f < u)}' "$FAKE_ROOT/calls.log" ||
        { ok=0; why+="fence not written before the upload; "; }
    fi
  fi
  record "$c" "$ok" "${why:-baseline and current: exit $nrc, identical stdout and non-metadata calls ($(nlines <"$W/runs/$c-new.calls") calls)}"
}

# dryrun-parity: --dry-run of every command runs no tool (every fake tool replaced by a recording shim); uncovered
# commands print exactly the baseline; covered ones add only the lease/fence lines.
dryrun_parity() {
  local ok=1 why="" cmd args a extra
  mkdir -p "$W/shim" "$W/runs" "$W/tmp" "$W/home"
  for a in gcloud gsutil terraform docker ssh scp curl sudo jq; do
    printf '#!/bin/sh\nprintf "%%s\\n" "%s $*" >>"%s/shim.log"\nexit 97\n' "$a" "$W" >"$W/shim/$a"; chmod 755 "$W/shim/$a"
  done
  : >"$W/shim.log"
  for cmd in provision deploy start stop reset retire status receipt build push lease; do
    case $cmd in
      build | push) args="--candidate $C" ;;
      lease) args="--show" ;;
      *) args=${ARGS[$cmd]:-} ;;
    esac
    read -ra a <<<"$args"
    for v in base new; do
      [[ $v == new || $cmd != lease ]] || continue
      use_scripts "$W/$v"
      RC=0
      env -i PATH="$W/shim:/usr/bin:/bin" HOME="$W/home" TMPDIR="$W/tmp" LC_ALL=C bash "$R/infra/staging/staging.sh" "$cmd" "${a[@]}" \
        --config "$CONF" --dry-run >"$W/runs/dry-$cmd-$v.raw" 2>&1 || RC=$?
      [[ $RC == 0 ]] || { ok=0; why+="$cmd ($v) dry-run exit $RC; "; }
      # status: probe-$RANDOM; inventory and plan names carry the wall-clock second, which the two runs may straddle
      sed -E 's/probe-[0-9]+\./probe-N./; s/[0-9]{8}T[0-9]{6}Z/TS/g; s/[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z/TS/g' \
        "$W/runs/dry-$cmd-$v.raw" >"$W/runs/dry-$cmd-$v.out"
    done
    [[ $cmd == lease ]] && continue
    extra="$(diff "$W/runs/dry-$cmd-base.out" "$W/runs/dry-$cmd-new.out" | grep -v '^[0-9]' || true)"
    if [[ " ${COMMANDS[*]} " == *" $cmd "* ]]; then
      grep -v -E '^> \+ (gcloud storage |\(gs://)' <<<"$extra" | grep -q . && { ok=0; why+="$cmd: dry-run differs beyond lease/fence lines; "; }
      grep -q -E '^> \+ gcloud storage buckets describe' <<<"$extra" || { ok=0; why+="$cmd: no presence line; "; }
    else
      [[ -z $extra ]] || { ok=0; why+="$cmd: uncovered command's dry-run differs; "; }
    fi
  done
  grep -q . "$W/shim.log" && { ok=0; why+="executed: $(cut -d' ' -f1 "$W/shim.log" | sort -u | tr '\n' ' '); "; }
  mkdir -p "$REPORT/dryrun-parity"; cp "$W"/runs/dry-*.out "$REPORT/dryrun-parity/" 2>/dev/null || true
  record dryrun-parity "$ok" "${why:-no tool executed; uncovered commands identical to baseline; covered commands add only lease/fence lines}"
}

# ---- (b) a CI lease blocks every covered local command ---------------------------------------------------------------
# held_refused NAME: the last run was refused with exit 2 naming HOLDER, before any non-metadata call, lease unchanged.
held_refused() { # NAME HOLDER GEN
  local why=""
  [[ $RC == 2 ]] || why+="exit $RC, not 2; "
  [[ -z "$(calls_nonmeta)" ]] || why+="non-lease calls: $(calls_nonmeta | cut -f3 | cut -c1-60 | tr '\n' ' '); "
  [[ "$(fx_gen lease)" == "$3" ]] || why+="lease generation changed; "
  grep -qF "$2" "$ERR" || why+="refusal does not name $2; "
  printf '%s' "$why"
}

ci_holds() {
  local cmd=$1 why g
  world; use_scripts "$W/new"
  fx_lease "ci:100/200" "$((FX_NOW + 600))"; g="$(fx_gen lease)"
  st "ci-holds-$cmd" "$cmd"; keep "ci-holds-$cmd"
  why="$(held_refused x "ci:100/200" "$g")"
  record "ci-holds-$cmd" "$([[ -z $why ]] && echo 1 || echo 0)" "${why:-refused exit 2 naming ci:100/200 before any non-lease call; lease unchanged}"
}

ci_holds_expired() {
  local why g
  world; use_scripts "$W/new"
  fx_lease "ci:100/200" "$((FX_NOW - 10))"; g="$(fx_gen lease)"
  st ci-holds-expired stop; keep ci-holds-expired
  why="$(held_refused x "ci:100/200" "$g")"
  record ci-holds-expired "$([[ -z $why ]] && echo 1 || echo 0)" "${why:-an expired CI lease still refuses a local command (only CI breaks CI leases)}"
}

control_b() {
  local why g
  mkdir -p "$W/mut-b"; cp "$W/new"/*.sh "$W/mut-b/"
  sed -i '/lease_guard "\$COMMAND"/d' "$W/mut-b/staging.sh"
  world; use_scripts "$W/mut-b"
  fx_lease "ci:100/200" "$((FX_NOW + 600))"; g="$(fx_gen lease)"
  st control-b stop
  why="$(held_refused x "ci:100/200" "$g")"
  control control-b "main without lease_guard (ci-holds assertion)" "$why"
}

# ---- concurrency: a running local mutator holds the lease; another local mutator is refused --------------------------
concurrent_local() {
  local ok=1 why="" n0 i bg
  world; use_scripts "$W/new"
  : >"$FAKE_ROOT/vm/block-start"
  ( st concurrent-start start ) &
  bg=$!
  for ((i = 0; i < 300; i++)); do [[ -f $FAKE_ROOT/vm/start-blocked ]] && break; sleep 0.1; done
  [[ -f $FAKE_ROOT/vm/start-blocked ]] || { ok=0; why+="start never reached the VM side; "; }
  fx_cat lease | grep -q $'^holder\tlocal:' || { ok=0; why+="no local lease while start runs; "; }
  n0="$(wc -l <"$FAKE_ROOT/calls.log")"
  st concurrent-stop stop
  [[ $RC == 2 ]] || { ok=0; why+="stop exit $RC while start holds, not 2; "; }
  tail -n +"$((n0 + 1))" "$FAKE_ROOT/calls.log" | awk -F'\t' '!($1 == "gcloud" && $3 ~ /^(storage|projects describe) /)' | grep -q . &&
    { ok=0; why+="stop made non-lease calls while start held the lease; "; }
  rm -f "$FAKE_ROOT/vm/block-start"
  wait "$bg" || true
  [[ "$(cat "$W/runs/concurrent-start.rc")" == 0 ]] || { ok=0; why+="start exit $(cat "$W/runs/concurrent-start.rc"); "; }
  [[ "$(fx_gen lease)" == 0 ]] || { ok=0; why+="lease not released after start; "; }
  st concurrent-stop-after stop
  [[ $RC == 0 ]] || { ok=0; why+="stop after the release exit $RC; "; }
  record concurrent-local "$ok" "${why:-local start held the lease; a concurrent local stop was refused (exit 2, lease calls only); after start released it, stop ran}"
}

# ---- (c) release by the holder, or break-glass with its record, lets the waiting mutator proceed ---------------------
release_then_proceed() {
  local ok=1 why="" me
  world; use_scripts "$W/new"
  me="local:$(id -un)@$(uname -n):99999"
  fx_lease "$me" none
  st rp-stop-refused stop; [[ $RC == 2 ]] || { ok=0; why+="stop not refused under the local lease ($RC); "; }
  fx_run rp-release -- bash "$R/infra/staging/staging.sh" lease --release --config "$CONF"
  [[ $RC == 0 && "$(fx_gen lease)" == 0 ]] || { ok=0; why+="lease --release exit $RC; "; }
  st rp-stop stop; [[ $RC == 0 ]] || { ok=0; why+="stop after release exit $RC; "; }
  record release-then-proceed "$ok" "${why:-lease --release by the holder freed the lease; the waiting stop then ran}"
}

break_glass() {
  local ok=1 why="" rec
  world; use_scripts "$W/new"
  fx_lease "local:someone@elsewhere:1" none
  fx_run bg-release -- bash "$R/infra/staging/staging.sh" lease --release --config "$CONF"
  [[ $RC == 2 && "$(fx_gen lease)" != 0 ]] || { ok=0; why+="another operator's lease released ($RC); "; }
  fx_run bg-noconfirm -- bash "$R/infra/staging/staging.sh" lease --break --reason "owner decision" --config "$CONF"
  [[ $RC == 2 ]] || { ok=0; why+="--break without --confirm exit $RC; "; }
  fx_run bg-wrongconfirm -- bash "$R/infra/staging/staging.sh" lease --break --confirm preview-1 --reason "owner decision" --config "$CONF"
  [[ $RC == 2 ]] || { ok=0; why+="--break with another environment's --confirm exit $RC; "; }
  fx_run bg-noreason -- bash "$R/infra/staging/staging.sh" lease --break --confirm "$FX_ENV" --config "$CONF"
  [[ $RC == 2 ]] || { ok=0; why+="--break without --reason exit $RC; "; }
  [[ "$(calls_writes | nlines)" == 0 ]] || { ok=0; why+="a refused lease command wrote; "; }
  fx_run bg-break -- bash "$R/infra/staging/staging.sh" lease --break --confirm "$FX_ENV" --reason "owner decision" --config "$CONF"
  [[ $RC == 0 && "$(fx_gen lease)" == 0 ]] || { ok=0; why+="--break exit $RC; "; }
  rec="$(fx_keys | grep '^lease-breaks/' || true)"
  [[ -n $rec ]] && fx_cat "$rec" | grep -q $'^reason\towner decision' && fx_cat "$rec" | grep -q $'^lease.holder\tlocal:someone@elsewhere:1' ||
    { ok=0; why+="no lease-breaks record with reason and broken holder; "; }
  awk -F'\t' '/lease-breaks\// && $2 == "W" {b = NR} /storage rm .*\/lease / {r = NR} END {exit !(b && r && b < r)}' "$FAKE_ROOT/calls.log" ||
    { ok=0; why+="record not written before the delete; "; }
  st bg-stop stop; [[ $RC == 0 ]] || { ok=0; why+="stop after break exit $RC; "; }
  keep break-glass
  record break-glass "$ok" "${why:-release refused for another operator; --break needs --confirm $FX_ENV and --reason; the record precedes the delete; then stop ran}"
}

# ---- (e) STAGING_LEASE_HELD re-entrancy ------------------------------------------------------------------------------
lease_held_case() { # NAME SPEC EXPECT_OK
  local g why=""
  world; use_scripts "${4:-$W/new}"
  fx_lease "ci:7/8" "$((FX_NOW + 600))"; g="$(fx_gen lease)"
  st "$1" stop "STAGING_LEASE_HELD=${2//GEN/$g}"
  if (($3)); then
    [[ $RC == 0 ]] || why+="exit $RC; "
    [[ "$(fx_gen lease)" == "$g" ]] || why+="the caller's lease was released or changed; "
    grep -q $'^gcloud\tW\tstorage cp .*/lease ' "$FAKE_ROOT/calls.log" && why+="an inner acquire was attempted; "
  else
    why="$(held_refused x "" "$g")"
  fi
  printf '%s' "$why"
}

lease_held() {
  local ok=1 why="" w
  w="$(lease_held_case lh-ok "ci:7/8@GEN" 1)"; [[ -z $w ]] || { ok=0; why+="exact holder and generation: $w"; }
  w="$(lease_held_case lh-gen "ci:7/8@1" 0)"; [[ -z $w ]] || { ok=0; why+="wrong generation: $w"; }
  w="$(lease_held_case lh-holder "ci:7/9@GEN" 0)"; [[ -z $w ]] || { ok=0; why+="wrong holder: $w"; }
  w="$(lease_held_case lh-bad "garbage" 0)"; [[ -z $w ]] || { ok=0; why+="malformed: $w"; }
  record lease-held "$ok" "${why:-STAGING_LEASE_HELD with the exact holder and generation runs without acquire or release; a wrong generation, holder or form is refused exit 2 before any non-lease call}"
}

control_e() {
  local w
  mkdir -p "$W/mut-e"; cp "$W/new"/*.sh "$W/mut-e/"
  sed -i 's/((META_GEN == gen)) \&\& \[\[/true \&\& [[/' "$W/mut-e/meta-lib.sh"
  w="$(lease_held_case control-e "ci:7/8@1" 0 "$W/mut-e")"
  control control-e "STAGING_LEASE_HELD generation not compared (wrong-generation assertion)" "$w"
}

# ---- (g) an unreadable presence check refuses ------------------------------------------------------------------------
presence_case() { # NAME MODE [SCRIPTS]
  local why=""
  world; use_scripts "${3:-$W/new}"
  st "$1" stop "FAKE_BUCKET_MODE=$2"
  [[ $RC == 2 ]] || why+="exit $RC, not 2; "
  [[ -z "$(calls_nonmeta)" ]] || why+="non-metadata calls after the failed check; "
  printf '%s' "$why"
}

presence() {
  local w
  w="$(presence_case presence-403 403)"
  record presence-403 "$([[ -z $w ]] && echo 1 || echo 0)" "${w:-a 403 on the bucket check refuses exit 2 before any other call (fail closed)}"
  w="$(presence_case presence-error 500)"
  record presence-error "$([[ -z $w ]] && echo 1 || echo 0)" "${w:-a 503 on the bucket check refuses exit 2 before any other call (fail closed)}"
}

control_g() {
  local w
  mkdir -p "$W/mut-g"; cp "$W/new"/*.sh "$W/mut-g/"
  sed -i 's/elif \[\[ "\$(meta_class)" == notfound \]\]; then/elif true; then/' "$W/mut-g/meta-lib.sh"
  w="$(presence_case control-g 403 "$W/mut-g")"
  control control-g "any failed bucket check taken as absent (presence-403 assertion)" "$w"
}

# ---- owner of the metadata bucket, local operator path (Codex review finding 3) ----------------------------------------
# Without STAGING_META_PROJECT_NUMBER a local command reads GCP_PROJECT's number with gcloud projects describe (as the
# deployer); a foreign owner, or a number that cannot be read, refuses exit 2 before anything but the presence check.
owner_local() {
  local why=""
  world; use_scripts "$W/new"; printf '999999999999\n' >"$FAKE_ROOT/gcs/$FX_BUCKET/.owner"
  st owner-local-foreign stop
  [[ $RC == 2 && -z "$(calls_nonmeta)" ]] && grep -q "belongs to project number 999999999999" "$ERR" || why+="foreign owner: exit $RC; "
  world; use_scripts "$W/new"; : >"$FAKE_ROOT/projects-describe-fail"
  st owner-local-unreadable stop
  [[ $RC == 2 && -z "$(calls_nonmeta)" ]] && grep -q "STAGING_META_PROJECT_NUMBER" "$ERR" || why+="unreadable number: exit $RC; "
  st owner-local-env stop STAGING_META_PROJECT_NUMBER=123456789012
  [[ $RC == 0 ]] || why+="STAGING_META_PROJECT_NUMBER supplied: exit $RC; "
  record owner-local "$([[ -z $why ]] && echo 1 || echo 0)" "${why:-local: a foreign owner and an unreadable project number refuse exit 2 before any other call; STAGING_META_PROJECT_NUMBER supplies the number}"
}

# control-f: a lease line printed on stdout would break parity: the parity assertion must catch it.
control_f() {
  local ok
  mkdir -p "$W/mut-f"; cp "$W/new"/*.sh "$W/mut-f/"
  sed -i 's/^  LEASE_OWN=1 LEASE_HOLDER=\$holder LEASE_GEN=\$META_GEN$/  LEASE_OWN=1 LEASE_HOLDER=$holder LEASE_GEN=$META_GEN; printf "lease\\t%s\\n" "$holder"/' "$W/mut-f/meta-lib.sh"
  world; use_scripts "$W/base"; st control-f-base stop
  cp "$OUT" "$W/runs/control-f-base.keep"
  world; use_scripts "$W/mut-f"; st control-f-mut stop
  cmp -s "$W/runs/control-f-base.keep" "$OUT" && ok=0 || ok=1
  record control-f "$ok" "mutation: the lease holder printed on stdout; the parity stdout comparison $( ((ok)) && echo catches it || echo DID NOT catch it)"
}

for cmd in "${COMMANDS[@]}"; do
  for mode in free absent; do selected "parity-$cmd-$mode" && parity "$cmd" "$mode"; done
done
selected dryrun-parity && dryrun_parity
selected control-f && control_f
for cmd in "${COMMANDS[@]}"; do selected "ci-holds-$cmd" && ci_holds "$cmd"; done
selected ci-holds-expired && ci_holds_expired
selected control-b && control_b
selected concurrent-local && concurrent_local
selected release-then-proceed && release_then_proceed
selected break-glass && break_glass
selected lease-held && lease_held
selected control-e && control_e
selected presence && presence
selected control-g && control_g
selected owner-local && owner_local
# ---- CI side of the lease (a), (d) and the IAM-model cycles: driven through ci.sh by tests/ci-adapter.sh (the same
# fakes and permission model); its case results are imported here under the ENV-LEASE names.
ci_cases() {
  local sub=$REPORT/ci-adapter rc=0 name src r d
  bash "$FX_SRC/tests/ci-adapter.sh" --work "$W/ci-adapter" --report "$sub" --only order,life,control >"$W/ci-adapter.out" 2>&1 || rc=$?
  for name in local-blocks-ci-deploy:order-lease-local local-blocks-ci-stop:life-stop-local-lease \
    ci-lease-blocks-ci:order-lease-ci-live ci-break-expired:life-ci-break-expired iam-cycle:life-cycles \
    iam-no-delete:control-iam; do
    src=${name#*:}
    r="$(awk -F'\t' -v c="$src" '$1 == c {print $2}' "$sub/summary.tsv" 2>/dev/null)"
    d="$(awk -F'\t' -v c="$src" '$1 == c {print $3}' "$sub/summary.tsv" 2>/dev/null)"
    record "${name%%:*}" "$([[ $r == PASS ]] && echo 1 || echo 0)" "ci-adapter.sh $src: ${r:-MISSING (ci-adapter exit $rc)} ${d}"
  done
}
if [[ -f $FX_SRC/ci.sh ]] && { selected local || selected ci || selected iam; }; then ci_cases; fi
fx_finish
