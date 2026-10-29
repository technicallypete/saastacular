#!/usr/bin/env bash
# Dependency vulnerability checks (G012 Plan r1 §5; D0 dispositions D1/D3 in
# management/v0/results/G012-D0-codex-disposition.md). B-DEPSCAN runs it; it can also run alone.
#
#   bash infra/security/dependency-scan.sh --out DIR --advisory-volume NAME --gocache-volume NAME
#        [--label KEY=VALUE]... [--cid-dir DIR] [--date YYYY-MM-DD] scan
#
# Steps, each bounded in time and failing closed (never a pass):
#   1. fetch    the four fixed public advisory objects by plain HTTPS GET into the owned advisory volume (the only
#               advisory network step; no project data, query string, cookie or authentication is sent), with HTTP,
#               size, Content-Length, published MD5, zip, Last-Modified (72 h) and Go-database (30 d) checks;
#   2. phase A  go mod download + verify of the core module and go install of the pinned govulncheck (Go module
#               proxy and checksum database, as every core build already does), into the owned module-cache volume;
#   3. tools    vulnpolicy built from infra/validation/tools (network none);
#   4. fixtures fixed offline fixtures (infra/security/testdata): govulncheck against a synthetic Go database
#               (called, imported, compile failure), osv-scanner --offline against synthetic per-ecosystem
#               databases (High, Low, unrated, clean, malformed lockfile), vulnpolicy over those real outputs with
#               the none, valid and expired exception files;
#   5. current  osv-scanner --offline over core/go.mod, web/package-lock.json and intelligence/uv.lock, govulncheck
#               over the core module, vulnpolicy with infra/security/dependency-exceptions.tsv.
# Every scan and the evaluator run with --network none; inputs are streamed on stdin (no bind mount); containers
# have no environment secret, no Docker socket, all capabilities dropped and no-new-privileges. Both volumes and all
# containers are removed at exit. Results: DIR/results.tsv (step<TAB>PASS|FAIL<TAB>detail) and the raw outputs.
# Exit: 0 every step PASS; 1 any step FAIL; 64 usage.
set -uo pipefail

# ---- pins (single source; the harness reports them) ----
OSV_IMAGE="ghcr.io/google/osv-scanner:v2.6.0@sha256:afd838850ac1a0fcc15ff4a041dc9ba11123c3f0d2666217a5f0fcf9222b55fa"
GO_IMAGE="golang:1.27.3-alpine@sha256:f2c872bcee949b7be9c490efd02b48a8a2fbf947cdf533f1c58525ec9d8170fd"
GOVULNCHECK_MODULE="golang.org/x/vuln"
GOVULNCHECK_VERSION="v1.8.0"
GOVULNCHECK_H1="h1:clG4qBU6zH5VKjti8n5j8BBuYzoSha392xXMkXS351U="
GO_STDLIB_VERSION="1.27.3"
ADVISORY_OBJECTS=(
  "Go https://osv-vulnerabilities.storage.googleapis.com/Go/all.zip 67108864"
  "npm https://osv-vulnerabilities.storage.googleapis.com/npm/all.zip 536870912"
  "PyPI https://osv-vulnerabilities.storage.googleapis.com/PyPI/all.zip 134217728"
  "govulndb https://vuln.go.dev/vulndb.zip 33554432"
)
# osv-scanner v2.6.0 reads its offline databases from <cache>/osv-scalibr/<ecosystem>/all.zip (D1).
OSV_DB_SUBDIR="osv-scalibr"

usage() { echo "usage: dependency-scan.sh --out DIR --advisory-volume NAME --gocache-volume NAME [--label K=V]... [--cid-dir DIR] [--date YYYY-MM-DD] scan" >&2; exit 64; }
REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)"
OUT="" ADV="" GOC="" CID_DIR="" DATE="" MODE="" LABELS=()
while [[ $# -gt 0 ]]; do
  case "$1" in
    --out) [[ $# -ge 2 ]] || usage; OUT="$2"; shift 2 ;;
    --advisory-volume) [[ $# -ge 2 ]] || usage; ADV="$2"; shift 2 ;;
    --gocache-volume) [[ $# -ge 2 ]] || usage; GOC="$2"; shift 2 ;;
    --label) [[ $# -ge 2 && "$2" =~ ^[a-z0-9.]+=[A-Za-z0-9._-]+$ ]] || usage; LABELS+=(--label "$2"); shift 2 ;;
    --cid-dir) [[ $# -ge 2 ]] || usage; CID_DIR="$2"; shift 2 ;;
    --date) [[ $# -ge 2 && "$2" =~ ^[0-9]{4}-[0-9]{2}-[0-9]{2}$ ]] || usage; DATE="$2"; shift 2 ;;
    scan) MODE=scan; shift ;;
    *) usage ;;
  esac
done
[[ -n "$OUT" && -n "$ADV" && -n "$GOC" && "$MODE" == scan ]] || usage
[[ "$ADV" =~ ^[a-z0-9][a-z0-9_.-]{0,127}$ && "$GOC" =~ ^[a-z0-9][a-z0-9_.-]{0,127}$ && "$ADV" != "$GOC" ]] || usage
[[ -n "$DATE" ]] || DATE="$(date -u +%F)"
mkdir -p "$OUT" && OUT="$(cd "$OUT" && pwd -P)" || exit 64
[[ "$OUT" != "$REPO" && "$OUT" != "$REPO"/* ]] || { echo "dependency-scan: --out must be outside the repository" >&2; exit 64; }
[[ -z "$(ls -A -- "$OUT")" ]] || { echo "dependency-scan: --out must be empty (no stale input)" >&2; exit 64; }
# Both volumes must be new: an existing volume is never reused or removed by this script.
for v in "$ADV" "$GOC"; do
  if docker volume inspect "$v" >/dev/null 2>&1; then echo "dependency-scan: volume $v already exists" >&2; exit 64; fi
done
RES="$OUT/results.tsv"; : >"$RES"
TAG="$(od -An -N4 -tx1 /dev/urandom | tr -d ' \n')"
HARD=(--cap-drop ALL --security-opt no-new-privileges "${LABELS[@]}")
NOENV=(-e GOTOOLCHAIN=local)
FAILED=0
log() { echo "[$(date -u +%H:%M:%S)] dependency-scan: $*"; }
result() { printf '%s\t%s\t%s\n' "$1" "$2" "$3" >>"$RES"; log "$2 $1 - $3"; [[ "$2" == PASS ]] || FAILED=1; }

cleanup() {
  local n
  for n in $(docker ps -aq --filter "label=g012.depscan=$TAG" 2>/dev/null); do docker rm -f "$n" >/dev/null 2>&1; done
  docker volume rm "$ADV" "$GOC" >/dev/null 2>&1
}
trap cleanup EXIT

# drun NAME SECONDS docker-run-args...: one bounded, labelled container; on expiry it is killed (exit 124).
drun() {
  local name="$1" secs="$2" rc=0 cid=() sfx cf
  shift 2
  if [[ -n "$CID_DIR" ]]; then
    # One distinct, initially absent cidfile per actual invocation (callers run inside $(...) and pipelines, where a
    # counter would not persist): this subshell's PID plus a random suffix, as lib.sh docker_run_owned does. The earlier
    # cidfiles stay in the registry directory for the owned-resource cleanup.
    sfx="$(od -An -N4 -tx1 /dev/urandom 2>/dev/null | tr -d ' \n')"
    cf="$CID_DIR/g012-$name-$TAG-$BASHPID-$sfx.cid"
    if [[ ! "$sfx" =~ ^[0-9a-f]{8}$ || -e "$cf" ]]; then log "cannot allocate a new cidfile for $name"; return 125; fi
    cid=(--cidfile "$cf")
  fi
  timeout -k 10 "$secs" docker run --name "g012-$name-$TAG" --label "g012.depscan=$TAG" "${cid[@]}" "${HARD[@]}" "$@" || rc=$?
  if (( rc == 124 || rc == 137 )); then docker kill "g012-$name-$TAG" >/dev/null 2>&1; log "TIMEOUT $name after ${secs}s"; rc=124; fi
  docker rm -f "g012-$name-$TAG" >/dev/null 2>&1
  return "$rc"
}
# stream PATH...: the tracked bytes of the given repository paths as one tar on stdout (no bind mount).
stream() {
  (cd "$REPO" && git ls-files -z --cached --others --exclude-standard -- "$@" |
    while IFS= read -r -d '' f; do if [[ -f "$f" && ! -L "$f" ]]; then printf '%s\0' "$f"; fi; done |
    tar --null --no-recursion -T - -cf -)
}

log "pins: osv-scanner $OSV_IMAGE; Go $GO_IMAGE; govulncheck $GOVULNCHECK_MODULE@$GOVULNCHECK_VERSION ($GOVULNCHECK_H1); stdlib $GO_STDLIB_VERSION; scan date $DATE"
docker volume create "${LABELS[@]}" --label "g012.depscan=$TAG" "$ADV" >/dev/null || { result setup FAIL "cannot create the advisory volume"; exit 1; }
docker volume create "${LABELS[@]}" --label "g012.depscan=$TAG" "$GOC" >/dev/null || { result setup FAIL "cannot create the module-cache volume"; exit 1; }

# ---- 1. advisory fetch ----
{
  echo 'set -u; now=$(date -u +%s); fail=0'
  echo "mkdir -p /db/$OSV_DB_SUBDIR/Go /db/$OSV_DB_SUBDIR/npm /db/$OSV_DB_SUBDIR/PyPI /db/govulndb"
  cat <<'INNER'
fetch() { # ECO URL CAP OUT
  eco=$1 url=$2 cap=$3 out=$4
  if ! wget -S -T 120 -O "$out" "$url" 2>"$out.hdr"; then echo "OBJECT $eco url=$url FAIL=download"; fail=1; return; fi
  status=$(grep -m1 'HTTP/' "$out.hdr" | awk '{print $2}')
  size=$(wc -c <"$out" | tr -d ' ')
  clen=$(grep -i 'content-length:' "$out.hdr" | tail -1 | tr -d '\r' | awk '{print $2}')
  lm=$(grep -i 'last-modified:' "$out.hdr" | tail -1 | tr -d '\r' | sed 's/^ *[Ll]ast-[Mm]odified: *//')
  b64=$(grep -io 'md5=[A-Za-z0-9+/=]*' "$out.hdr" | tail -1 | cut -d= -f2-)
  pub=""; [ -n "$b64" ] && pub=$(printf '%s' "$b64" | base64 -d | od -An -tx1 | tr -d ' \n')
  md5=$(md5sum "$out" | cut -d' ' -f1)
  if unzip -tq "$out" >/dev/null 2>&1; then zt=ok; else zt=failed; fi
  entries=$(unzip -l "$out" 2>/dev/null | tail -1 | awk '{print $2}')
  lme=$(date -u -D '%a, %d %b %Y %H:%M:%S GMT' -d "$lm" +%s 2>/dev/null || echo 0)
  age=$(( now - lme ))
  rm -f "$out.hdr"
  echo "OBJECT $eco url=$url status=$status size=$size content_length=$clen md5=$md5 published_md5=${pub:-absent} zip=$zt entries=${entries:-0} last_modified=\"$lm\" age_s=$age"
  [ "$status" = 200 ] || { echo "CHECK $eco HTTP status"; fail=1; }
  [ "$size" -le "$cap" ] || { echo "CHECK $eco size cap $cap"; fail=1; }
  [ "$size" = "$clen" ] || { echo "CHECK $eco size differs from Content-Length"; fail=1; }
  if [ -n "$pub" ]; then [ "$pub" = "$md5" ] || { echo "CHECK $eco MD5 differs from the published value"; fail=1; }; else echo "NOTE $eco published MD5 absent"; fi
  [ "$zt" = ok ] && [ "${entries:-0}" -ge 1 ] || { echo "CHECK $eco zip test or entries"; fail=1; }
  [ "$lme" -gt 0 ] && [ "$age" -le 259200 ] || { echo "CHECK $eco Last-Modified older than 72 h or unreadable"; fail=1; }
}
INNER
  for o in "${ADVISORY_OBJECTS[@]}"; do
    read -r eco url cap <<<"$o"
    if [[ "$eco" == govulndb ]]; then echo "fetch $eco $url $cap /db/vulndb.zip"; else echo "fetch $eco $url $cap /db/$OSV_DB_SUBDIR/$eco/all.zip"; fi
  done
  cat <<'INNER'
if unzip -q -o /db/vulndb.zip -d /db/govulndb >/dev/null 2>&1 && [ -f /db/govulndb/index/db.json ]; then
  mod=$(sed -n 's/.*"modified" *: *"\([^"]*\)".*/\1/p' /db/govulndb/index/db.json | head -1)
  me=$(date -u -D '%Y-%m-%dT%H:%M:%S' -d "$(echo "$mod" | cut -c1-19)" +%s 2>/dev/null || echo 0)
  echo "GODB index/db.json modified=$mod age_s=$(( now - me ))"
  [ "$me" -gt 0 ] && [ $(( now - me )) -le 2592000 ] || { echo "CHECK Go database older than 30 days or unreadable"; fail=1; }
else echo "CHECK vulndb.zip does not unpack to index/db.json"; fail=1; fi
rm -f /db/vulndb.zip
chmod -R a+rX /db
du -sk /db | awk '{print "STORED_KB " $1}'
exit $fail
INNER
} >"$OUT/fetch.sh"
t0=$(date +%s); rc=0
drun fetch 300 -i --rm -v "$ADV:/db" "$GO_IMAGE" sh -s <"$OUT/fetch.sh" >"$OUT/fetch.log" 2>&1 || rc=$?
grep -E '^(OBJECT|GODB|CHECK|NOTE|STORED_KB)' "$OUT/fetch.log" | sed 's/^/  /'
if (( rc == 0 )); then result fetch PASS "4 advisory objects fetched and checked in $(( $(date +%s) - t0 ))s ($(sed -n 's/^STORED_KB //p' "$OUT/fetch.log") KiB stored)"
else result fetch FAIL "advisory fetch exit $rc"; exit 1; fi

# ---- 2. phase A: module download/verify and the pinned govulncheck (network: Go module proxy, checksum database) ----
# The owned cache starts empty: allow ten minutes for module downloads, verification and a cold tool compile.
# This remains a hard failure on expiry; scanner versions, hash verification and offline scan limits do not change.
t0=$(date +%s); rc=0
stream core | drun gova 600 -i --rm -v "$GOC:/gocache" -e GOPATH=/gocache/gopath -e GOMODCACHE=/gocache/mod \
  -e GOCACHE=/gocache/build "${NOENV[@]}" "$GO_IMAGE" sh -ec "
    go telemetry off
    mkdir -p /src && tar -x -C /src && cd /src/core && GOFLAGS=-mod=readonly go mod download && go mod verify
    cd / && go install $GOVULNCHECK_MODULE/cmd/govulncheck@$GOVULNCHECK_VERSION
    h=\$(cat /gocache/mod/cache/download/$GOVULNCHECK_MODULE/@v/$GOVULNCHECK_VERSION.ziphash)
    [ \"\$h\" = '$GOVULNCHECK_H1' ] || { echo \"govulncheck module hash \$h differs from the pin\"; exit 3; }
    go version" >"$OUT/phase-a.log" 2>&1 || rc=$?
tail -3 "$OUT/phase-a.log" | sed 's/^/  /'
if (( rc == 0 )); then result phase-a PASS "core modules downloaded and verified; $GOVULNCHECK_MODULE@$GOVULNCHECK_VERSION installed, module hash equals the pin ($(( $(date +%s) - t0 ))s)"
else result phase-a FAIL "module download, verification or govulncheck install exit $rc"; exit 1; fi

# Network-free Go runner: the module cache read-write (owned volume), a fresh build cache, no proxy.
GORUN=(-v "$GOC:/gocache" -e GOPATH=/gocache/gopath -e GOMODCACHE=/gocache/mod -e GOCACHE=/tmp/gobuild -e GOPROXY=off
  -e GOFLAGS=-mod=readonly "${NOENV[@]}" --network none)

# ---- 3. vulnpolicy (network none) ----
rc=0
stream infra/validation/tools | drun tools 120 -i --rm "${GORUN[@]}" "$GO_IMAGE" sh -ec \
  'mkdir -p /src && tar -x -C /src && cd /src/infra/validation/tools && go build -o /gocache/bin/vulnpolicy ./cmd/vulnpolicy' >"$OUT/tools.log" 2>&1 || rc=$?
if (( rc == 0 )); then result vulnpolicy-build PASS "vulnpolicy built without network"; else result vulnpolicy-build FAIL "exit $rc: $(tail -2 "$OUT/tools.log" | tr '\n' ' ')"; exit 1; fi

# evaluate NAME SECONDS ARGS...: vulnpolicy evaluate over files staged in $OUT/eval-NAME (streamed in, network none).
evaluate() {
  local name="$1" rc=0
  shift
  tar -C "$OUT/eval-$name" -cf - . | drun "eval-$name" 60 -i --rm "${GORUN[@]}" "$GO_IMAGE" sh -ec \
    'mkdir -p /e && tar -x -C /e && cd /e && rc=0 && { /gocache/bin/vulnpolicy evaluate --report /e/report.json "$@" || rc=$?; } && echo "VULNPOLICY_EXIT $rc" && echo REPORT_JSON && cat /e/report.json' \
    sh "$@" >"$OUT/eval-$name.log" 2>&1 || rc=$?
  (( rc == 0 )) || { echo "VULNPOLICY_EXIT runner-$rc" >>"$OUT/eval-$name.log"; }
  sed -n 's/^VULNPOLICY_EXIT //p' "$OUT/eval-$name.log" | tail -1
}

# ---- 4. fixed offline fixtures ----
FX="$REPO/infra/security/testdata"
# 4a. govulncheck: synthetic local Go database, called and imported findings, compile failure.
rc=0
stream infra/security/testdata/govulncheck | drun fx-govuln 240 -i --rm "${GORUN[@]}" "$GO_IMAGE" sh -ec \
  'mkdir -p /s && tar -x -C /s && cd /s/infra/security/testdata/govulncheck/module && /gocache/gopath/bin/govulncheck -db file:///s/infra/security/testdata/govulncheck/db -format json -test ./...' \
  >"$OUT/fx-govulncheck.json" 2>"$OUT/fx-govulncheck.stderr" || rc=$?
brc=0
stream infra/security/testdata/govulncheck | drun fx-govuln-broken 240 -i --rm "${GORUN[@]}" "$GO_IMAGE" sh -ec \
  'mkdir -p /s && tar -x -C /s && cd /s/infra/security/testdata/govulncheck/broken && /gocache/gopath/bin/govulncheck -db file:///s/infra/security/testdata/govulncheck/db -format json -test ./...' \
  >"$OUT/fx-govulncheck-broken.json" 2>"$OUT/fx-govulncheck-broken.stderr" || brc=$?
fx_go=""
(( rc == 0 )) || fx_go+="fixture module exit $rc; "
(( brc != 0 )) || fx_go+="the non-compiling module returned exit 0; "
mkdir -p "$OUT/eval-fx-go"
cp "$OUT/fx-govulncheck.json" "$OUT/eval-fx-go/govulncheck.json"
cp "$FX/govulncheck/module/go.mod" "$OUT/eval-fx-go/go.mod"
cp "$FX/exceptions/none.tsv" "$FX/exceptions/go-valid.tsv" "$FX/exceptions/go-expired.tsv" "$OUT/eval-fx-go/"
for e in none:1 go-valid:0 go-expired:1; do
  got="$(evaluate fx-go --date "$DATE" --exceptions "${e%%:*}.tsv" --govulncheck govulncheck.json --gomod go.mod)"
  [[ "$got" == "${e##*:}" ]] || fx_go+="exceptions ${e%%:*}: vulnpolicy exit $got, want ${e##*:}; "
  cp "$OUT/eval-fx-go.log" "$OUT/eval-fx-go-${e%%:*}.log"
done
grep -q 'FINDING BLOCKING Go:stdlib@1.27.3 ids=GO-2099-0001 .*class=Unclassified reachability=called' "$OUT/eval-fx-go-none.log" || fx_go+="GO-2099-0001 is not a blocking called finding; "
grep -q 'FINDING BLOCKING Go:stdlib@1.27.3 ids=GO-2099-0002 .*class=Unclassified reachability=imported' "$OUT/eval-fx-go-none.log" || fx_go+="GO-2099-0002 is not a blocking imported finding; "
grep -q 'FINDING excepted Go:stdlib@1.27.3 ids=GO-2099-0002' "$OUT/eval-fx-go-go-valid.log" || fx_go+="the valid exception did not mark GO-2099-0002 excepted; "
grep -q 'PROBLEM .*expired on 2026-01-01 (stale)' "$OUT/eval-fx-go-go-expired.log" || fx_go+="the expired row was not reported stale; "
if [[ -z "$fx_go" ]]; then result fixtures-govulncheck PASS "synthetic Go database: GO-2099-0001 called and GO-2099-0002 imported, both blocking (Unclassified); non-compiling module fails (exit $brc); exceptions none/valid/expired -> 1/0/1"
else result fixtures-govulncheck FAIL "$fx_go"; fi

# 4b. osv-scanner: synthetic per-ecosystem databases zipped at run time, tiny lockfiles, a malformed lockfile.
fx_osv=""
mkdir -p "$OUT/fx-osv/db/$OSV_DB_SUBDIR/npm" "$OUT/fx-osv/db/$OSV_DB_SUBDIR/PyPI" "$OUT/fx-osv/in"
for eco in npm PyPI; do
  rc=0
  stream infra/security/testdata/osv/records/$eco | drun "fx-zip-$eco" 60 -i --rm "${GORUN[@]}" "$GO_IMAGE" sh -ec \
    'mkdir -p /s && tar -x -C /s && /gocache/bin/vulnpolicy fixture-db --out /tmp/all.zip /s/infra/security/testdata/osv/records/*/*.json >&2 && cat /tmp/all.zip' \
    >"$OUT/fx-osv/db/$OSV_DB_SUBDIR/$eco/all.zip" 2>"$OUT/fx-zip-$eco.log" || rc=$?
  (( rc == 0 )) || fx_osv+="fixture database $eco exit $rc; "
done
cp "$FX/osv/lock/package-lock.json" "$FX/osv/lock/uv.lock" "$OUT/fx-osv/in/"
mkdir -p "$OUT/fx-osv/in/broken"; cp "$FX/osv/broken/package-lock.json" "$OUT/fx-osv/in/broken/"
osv_run() { # osv_run NAME DBDIR OUTFILE LOCKFILE...: osv-scanner --offline over the streamed tar (unpacked at /tmp/s)
  local name="$1" db="$2" outf="$3" rc=0 args=() f
  shift 3
  for f in "$@"; do args+=(-L "/tmp/s/$f"); done
  drun "$name" 180 -i --rm --network none --user 65534:65534 -w /tmp -v "$ADV:/db:ro" -e OSV_SCANNER_LOCAL_DB_CACHE_DIRECTORY="$db" \
    --entrypoint /bin/sh "$OSV_IMAGE" -c 'mkdir -p /tmp/s && tar -x -C /tmp/s && exec /osv-scanner scan source --offline --no-call-analysis=go --format json --all-packages "$@"' \
    sh "${args[@]}" >"$outf" 2>"${outf%.json}.stderr" || rc=$?
  return "$rc"
}
rc=0; tar -C "$OUT/fx-osv" -cf - db in | osv_run fx-osv /tmp/s/db "$OUT/fx-osv.json" in/package-lock.json in/uv.lock || rc=$?
[[ $rc == 1 ]] || fx_osv+="fixture scan exit $rc, want 1 (findings); "
brc=0; tar -C "$OUT/fx-osv" -cf - db in | osv_run fx-osv-broken /tmp/s/db "$OUT/fx-osv-broken.json" in/broken/package-lock.json || brc=$?
[[ $brc != 0 && $brc != 1 ]] || fx_osv+="the malformed lockfile returned exit $brc; "
mkdir -p "$OUT/eval-fx-osv"
cp "$OUT/fx-osv.json" "$OUT/eval-fx-osv/osv.json"; cp "$FX/osv/lock/package-lock.json" "$FX/osv/lock/uv.lock" "$OUT/eval-fx-osv/"
cp "$FX/exceptions/none.tsv" "$FX/exceptions/osv-valid.tsv" "$FX/exceptions/osv-expired.tsv" "$OUT/eval-fx-osv/"
for e in none:1 osv-valid:0 osv-expired:1; do
  got="$(evaluate fx-osv --date "$DATE" --exceptions "${e%%:*}.tsv" --osv osv.json --npm-lock package-lock.json --uv-lock uv.lock)"
  [[ "$got" == "${e##*:}" ]] || fx_osv+="exceptions ${e%%:*}: vulnpolicy exit $got, want ${e##*:}; "
  cp "$OUT/eval-fx-osv.log" "$OUT/eval-fx-osv-${e%%:*}.log"
done
grep -q 'FINDING BLOCKING npm:g012-fixture-high@1.0.0 ids=OSV-2099-1001 .*class=High' "$OUT/eval-fx-osv-none.log" || fx_osv+="the High npm finding is not blocking; "
grep -q 'FINDING reported PyPI:g012-fixture-low@2.0.0 ids=OSV-2099-2001 .*class=Low' "$OUT/eval-fx-osv-none.log" || fx_osv+="the Low PyPI finding is not reported non-blocking; "
grep -q 'FINDING BLOCKING PyPI:g012-fixture-unrated@3.0.0 ids=OSV-2099-2002 .*class=Unclassified' "$OUT/eval-fx-osv-none.log" || fx_osv+="the unrated PyPI finding is not blocking Unclassified; "
grep -q 'GRAPH npm identities=2 ' "$OUT/eval-fx-osv-none.log" && grep -q 'GRAPH PyPI identities=2 ' "$OUT/eval-fx-osv-none.log" || fx_osv+="fixture identities do not reconcile; "
grep -q 'FINDING excepted npm:g012-fixture-high@1.0.0' "$OUT/eval-fx-osv-osv-valid.log" || fx_osv+="the valid exception did not mark the High finding excepted; "
if [[ -z "$fx_osv" ]]; then result fixtures-osv PASS "synthetic databases: High npm blocking, Low PyPI reported, unrated PyPI blocking (Unclassified), clean package none; malformed lockfile fails (exit $brc); identities reconcile; exceptions none/valid/expired -> 1/0/1"
else result fixtures-osv FAIL "$fx_osv"; fi

# ---- 5. current tree ----
cur=""
t0=$(date +%s); rc=0
stream core/go.mod web/package-lock.json intelligence/uv.lock | osv_run osv /db "$OUT/osv.json" core/go.mod web/package-lock.json intelligence/uv.lock || rc=$?
grep -E '^Loaded .* local db' "$OUT/osv.stderr" | sed 's/^/  /'
[[ $rc == 0 || $rc == 1 ]] || cur+="osv-scanner exit $rc; "
osv_s=$(( $(date +%s) - t0 ))
t0=$(date +%s); grc=0
stream core | drun govb 240 -i --rm "${GORUN[@]}" -v "$ADV:/db:ro" "$GO_IMAGE" sh -ec \
  'go telemetry off; mkdir -p /src && tar -x -C /src && cd /src/core && exec /gocache/gopath/bin/govulncheck -db file:///db/govulndb -format json -test ./...' \
  >"$OUT/govulncheck.json" 2>"$OUT/govulncheck.stderr" || grc=$?
(( grc == 0 )) || cur+="govulncheck exit $grc; "
gv_s=$(( $(date +%s) - t0 ))
mkdir -p "$OUT/eval-current"
cp "$OUT/osv.json" "$OUT/eval-current/osv.json"; cp "$OUT/govulncheck.json" "$OUT/eval-current/govulncheck.json"
cp "$REPO/core/go.mod" "$REPO/web/package-lock.json" "$REPO/intelligence/uv.lock" "$REPO/infra/security/dependency-exceptions.tsv" "$OUT/eval-current/"
got="$(evaluate current --date "$DATE" --exceptions dependency-exceptions.tsv --osv osv.json --govulncheck govulncheck.json \
  --gomod go.mod --npm-lock package-lock.json --uv-lock uv.lock)"
sed -n '/^REPORT_JSON$/,$p' "$OUT/eval-current.log" | sed 1d >"$OUT/current-report.json"
grep -E '^(GRAPH|FINDING|INCOMPLETE|PROBLEM|RESULT)' "$OUT/eval-current.log" | sed 's/^/  /'
[[ "$got" == 0 ]] || cur+="vulnpolicy exit $got; "
if [[ -z "$cur" ]]; then result current PASS "osv-scanner ${osv_s}s (exit $rc), govulncheck ${gv_s}s, vulnpolicy pass: $(grep -c '^FINDING' "$OUT/eval-current.log") findings, $(grep -c '^FINDING BLOCKING' "$OUT/eval-current.log") blocking unexcepted"
else result current FAIL "$cur"; fi

exit "$FAILED"
