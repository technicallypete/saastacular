#!/usr/bin/env bash
# Offline checks of the on-demand staging tooling (G070 Plan S16 as corrected by r1.2/r1.3, D-C3; Codex r1 review
# clarification 1 and the shared-file fallback). Coordinator-run and unregistered: not a harness suite or check ID;
# its ledger lines are recorded in the G070 Result with command, exit code and input digest.
#
#   bash infra/staging/check.sh prepare --work DIR [--edge-tag g070-edge:TAG] [--platform OS_ARCH]
#        [--netns-tag g070-netns:TAG [--debian-image debian:12@sha256:DIGEST]]
#   bash infra/staging/check.sh run --work DIR --report DIR --edge-image REF [--only ROW[,ROW]...] [--env-name NAME]
#        [--netns-image REF] [--allow-dirty] [--cases ROW=CASE[,CASE]...]...
#
# prepare  the only networked mode, run once per input change: pulls the pinned tool images (Terraform, jq by digest,
#          and the harness's yq, ShellCheck and hadolint pins), mirrors the Terraform
#          providers that infra/terraform/environments/example requires into DIR/tf-mirror (hold the coordinator's
#          cache-preparation lock) and, with --edge-tag, builds infra/docker/edge.Dockerfile under that task-owned tag
#          (hold the single build slot). With --netns-tag it pulls --debian-image (default debian:12; pass the P0.5
#          rehearsal's digest pin) and builds, under that tag, the same image plus iptables, iproute2 and busybox from
#          Debian's archive for METADATA-GUARD's packet proof (build slot; one build after the other). Writes
#          DIR/prepare.tsv (image, id, repo digest).
# run      every container runs with --network none and --pull never; nothing is built, pulled or published. Rows:
#   TF-FMT            terraform fmt -check -recursive on a disposable writable copy of infra/terraform
#   TF-VALIDATE       init -backend=false -lockfile=readonly + validate in environments/example, against the mirror
#   TF-TEST           per module (modules/*): init + validate + terraform test -json, isolated TF_DATA_DIR; a module
#                     passes only with failed=errored=skipped=0 and passed>=1 (a zero-test exit is a FAIL)
#   TF-GREP           no owner/editor/IAM-admin/key roles, service-account keys, DELETE termination, provisioners,
#                     metadata_startup_script, Secret Manager resources or secret-shaped variables/outputs; exactly one
#                     lifecycle ignore_changes in infra/terraform, ignore_changes = [scheduling[0].max_run_duration] in
#                     gcp-staging-vm's google_compute_instance.vm, whose max_run_duration still takes var.session_max_seconds
#   EDGE-ADAPT        caddy list-modules has dns.providers.cloudflare; caddy adapt --validate of infra/edge/Caddyfile;
#                     the adapted JSON: admin off, ask URL, on-demand only without subjects, DNS-01 token as an env
#                     reference, filtered access loggers, X-Forwarded-* removed upstream, MCP paths, no /internal
#                     matcher, HSTS without includeSubDomains/preload, no token value
#   EDGE-FILTER       the site log blocks in a plain-HTTP loopback copy of the Caddyfile; G070SENTINEL requests on the
#                     callback/signup paths (plus another key, another path and Cookie/Authorization headers) leave
#                     zero sentinel or sensitive-key hits, while every request is logged with host, path and status
#   LIFECYCLE-DRYRUN  staging.sh <command> --dry-run for every subcommand with the committed example config; no
#                     tool is executed (shims record any call); forbidden verbs absent; reset removes exactly the two
#                     data volumes. staging.conf's optional DEFAULT_TENANT_SLUG/PLATFORM_ROOT_PAGE: one accepted and
#                     two refused configs (exact refusal text); the accepted values and the empty defaults reach
#                     remote.env and, through remote/lib.sh load_meta + render_env, the private env (KEY= when empty).
#                     Two environments (staging, staging-eu) in one project, staging.sh sourced with a filter-applying
#                     fake gcloud: reconcile, verify_absent, env_secrets and label_check take only staging's exact
#                     names and label; a stray labelled staging-extra is still reported, an unlabelled one refused.
#                     The fake reads the filter language: name ~ needs one single-quoted value (an unquoted one fails
#                     the call), and the grouped expressions arrive as name ~ '^(a|b)$'; preview-1 beside preview-10
#                     takes only its own names. Session length (Plan r1.9, A2 fallback of round 3): start --duration
#                     90m/2h/7200 dry runs show, in order, the fresh status read, gcloud set-scheduling with the parsed
#                     seconds, the scheduling readback, a saved refresh-only plan, label check, refresh scope check,
#                     apply, re-read and start, and no normal plan; malformed, out-of-range, repeated or misplaced
#                     --duration and --override session_max_seconds exit 2 before any command; deploy runs no
#                     Terraform, VM start/stop or set-scheduling. staging.sh sourced with the instance, Terraform and the
#                     VM side faked (the fake terraform fails a normal plan that would replace the VM): running
#                     same/different/drift, stopped same/different (600)/default (7200) with the exact set-scheduling
#                     argv and order, missing token, drift HOLD, recovery of instance == requested != Terraform, and
#                     each failure (set-scheduling, readback length/attribute/unreadable, refresh plan exit 1 or 0, scope
#                     refusal, apply, output old or unreadable) HOLDs without a VM start and prints the partial state
#                     and recovery; provision keeps an existing VM's observed length and refuses --duration there. A VM
#                     is absent only when the exact-name list confirms it; an unreadable state or session length HOLDs
#                     in provision, start, stop, status and retire before any write or default. refresh_scope_check on
#                     refresh_scope_rows TSV fixtures, and the real refresh_scope_rows jq program (pinned JQ_IMAGE,
#                     --network none) on hand-written refresh-only `terraform show -json` documents, accept only the
#                     length drift of module.vm.google_compute_instance.vm (plus current_status/cpu_platform and the
#                     recorded-state lag leaves of benign_drift: attachment records, empty defaults, an IAM etag)
#   STAGING-GREP      infra/staging scripts: docker inspect only with --format and never .Config.Env/.Config;
#                     compose config only on a dry-run line; no config.v2.json, /proc/*/environ or test overlays. The
#                     session default 7200 pinned in lib.sh, both Terraform variable defaults, the examples, STAGING.md
#                     and the Terraform README; the 600..28800 bounds in lib.sh and the module validation; STAGING.md
#                     documents --duration; no Cloud Scheduler, cron or at job in infra/staging or infra/terraform;
#                     staging.sh plans normally (no -refresh-only/-destroy) only in cmd_provision and applies only in
#                     cmd_provision, session_refresh and cmd_retire, with no saved session plan or duration_scope_check
#   COMPOSE-STATIC    from an archive of infra/{compose,postgres,minio,edge,staging} only (P0.1): local.yml +
#                     migrate.yml + staging.yml config; only the edge publishes (80/443); no build key; every built
#                     service and the edge digest-pinned from ${IMAGE_*}; networks; no test service or variable;
#                     APP_ENV not test; https cookie profile on web and core; per-secret service matrix;
#                     DEFAULT_TENANT_SLUG and PLATFORM_ROOT_PAGE reach core only, as given
#   START-ORDER       remote/start.sh --dry-run: app and edge are stopped by label before the first secret fetch;
#                     start.sh sourced with only compose_env failing: the label stop ran first, then HOLD (exit 3),
#                     and no store, app or edge start. remote/stop.sh sourced with docker replaced: a failed Compose
#                     stop falls back to the label stop of every service (STOPPED); both failing exit 4 without
#                     STOPPED and without removing the env file first. remote/deploy.sh sourced with docker replaced:
#                     a fresh volume is seeded, a dataset with its marker never; one without its marker is seeded
#                     once only when its application-data count is 0, and HOLDs on rows, a failed or odd count
#   HTTPS-CERT        remote/lib.sh https_ready on synthetic cert_lines (curl, sleep and cert_lines replaced; no
#                     network): distinct apex and wildcard serials pass; a label SAN without exactly *.B, an apex SAN
#                     without B, an expiry within 7 days on either name, an unverified label chain or, in the
#                     rehearsal, an issuer without (STAGING) fail with their reason
#   CF-TOKEN-CANARY   remote/lib.sh compose_env with only the secret-fetch seams (sm_token, sm_access) and recorders
#                     for docker/run replaced: two malformed CF_API_TOKEN canaries and an unsupported-prefix (cfxx_)
#                     canary are refused with the fixed message in deploy and pinned mode, with no docker/compose call
#                     and no env file; the canaries appear in no output, scratch file or report; well-formed tokens
#                     generated at run time (40 characters and cfut_) pass (unprinted)
#   METADATA-PROBE    remote/lib.sh metadata_denied with docker replaced by a fake daemon holding a container table (no
#                     Docker). Classification: statuses 1, 124 and 143 with no answer are denied; 139, 137 and 125 with
#                     no answer are UNPROVEN; an answer (also with a failure status) is REACHABLE; a failed tool control
#                     is UNPROVEN and stops that network's probes; the answer is never printed; no wget -T. Identity:
#                     every probe container is named, labelled, has a --cidfile, --init (timeout is never PID 1, where it
#                     would not fire) and no --rm; the inventory
#                     (state/probe-containers.tsv) has allocate, created (64-hex container ID, image ID) and removed
#                     rows, removal uses the recorded ID. Reconciliation: an owned stale container is reaped by
#                     identity, a stale record is closed, and an unowned container or one whose image ID, container ID
#                     or run label differs from its record is held (exit 1) and untouched, as is an unrelated
#                     container; a failed removal, a missing container ID and an interrupted run (cleanup) are covered
#   METADATA-GUARD    static: startup.sh's guard unit (oneshot, Before=docker.service), the docker.service drop-in
#                     (Requires=, After=), raw PREROUTING DROP rules for 169.254.169.254 and fd20:ce::254, the guard
#                     installed before Docker in both modes; metadata_denied first in readiness and before app_up in
#                     deploy/start/reset; no host network or metadata address in the composed files. Packet level
#                     (--netns-image; NET_ADMIN, SYS_ADMIN, ip_forward in a --network none container): a fake container
#                     namespace routed through the container's own namespace (the host) to a metadata namespace
#                     reaches the listener before and not after the real `startup.sh --metadata-guard-rules`; the host
#                     still does, port 53 stays open, a second run adds no rule; a missing prerequisite is a FAIL
#   LIFECYCLE-DRYRUN  (continued) G072 (Plan §10.5, §10.9): each covered mutator (provision, deploy, start, stop, reset, retire)
#                     prints the metadata-bucket presence check and the lease create as its first two commands, build,
#                     push, status and receipt print neither; deploy and provision print the schema-fence entry before
#                     the release upload. Additive interfaces: build --source DIR takes the Compose files, contexts and
#                     git checks from DIR (the default still from the checkout) and refuses a relative, missing or
#                     absent value and push --source; GCLOUD_IMPERSONATE_SA absent or empty gives the default stop's
#                     exact gcloud lines, ci-stop the same lines with only the identity replaced, push keeps the
#                     deployer, and four near-miss values are refused with the exact text; lease --show/--release/
#                     --break dry runs and their four refusals; STAGING_LEASE_HELD checks the named lease instead of
#                     acquiring, and a malformed value is refused
#   STAGING-GREP      (continued) G072 routing rules over infra/staging/ci.sh and ci-lib.sh (Plan §10.7, §10.9):
#                     every staging.sh subprocess writes stdout and stderr only to $RAW/<step>.log (the local
#                     `lease` passthrough excepted) and runs under STAGING_LEASE_HELD unless its subcommand takes no
#                     lease; nothing reads raw/ to the job's stdout/stderr/fd 3 or uploads from raw/; the logs/ upload is
#                     the publish_filter output of its own function; staging.sh takes the lease for exactly the six
#                     mutators before cmd_<command>; meta-lib.sh's gcloud storage verbs are only describe/cat/cp/rm/ls
#                     and lib.sh's forbid is the G070 baseline's; the job refuses an inherited STAGING_LEASE_HELD and
#                     the launcher unsets it. Mutation controls: each rule must flag a deliberately broken copy.
#                     infra/staging/tests/fixtures/ is excluded from every STAGING-GREP scan (EP-1's hostile fixture)
#   CI-FEDERATION     bash infra/staging/tests/ci-federation.sh (fake gcloud): PASS only with exit 0 and a final
#                     "RESULT PASS passed=N failed=0" (N >= 1)
#   CI-ADAPTER        bash infra/staging/tests/ci-adapter.sh (fakes only): exit 0 and "RESULT PASS cases=N failed=0"
#   ENV-LEASE         bash infra/staging/tests/env-lease.sh (fakes only): exit 0 and "RESULT PASS cases=N failed=0"
#   CI-LAUNCHER       bash infra/staging/tests/ci-launcher.sh (fake docker): exit 0 and its final line
#                     "ci-launcher: N selected, N passed, 0 failed" (N >= 1; it prints no RESULT line)
#                     The four fixture rows write their reports under REPORT/<row in lower case>/ and their stdout to
#                     REPORT/<row>.out; --cases ROW=SEL passes --only SEL to CI-ADAPTER, ENV-LEASE or CI-LAUNCHER (a
#                     targeted rerun, recorded as a subset in the ledger)
#   SHELLCHECK-DIRECT shellcheck (severity warning, as B-SHELL) on the staging tooling: exactly B-SHELL's
#                     infra/staging, infra/terraform and infra/ci/staging-action.sh entries (the row fails when its
#                     list and run.sh's differ, or when a tracked infra/staging or infra/terraform script outside
#                     tests/fakes is in neither)
#   HADOLINT-DIRECT   hadolint (failure threshold warning, as B-HADOLINT) on infra/docker/edge.Dockerfile
#
# Ledger: REPORT/ledger.tsv (row, result, exit, commit, inputs_sha256, command, detail); detail never holds a value
# from the generated environment. The generated secrets and the fake edge token live only in a 0700 scratch
# directory under DIR that is removed at exit; copied logs are scrubbed of them first. Every container carries the
# labels g070.task=G070 and g070.check.run=<run>; leftovers are swept by label and their absence verified.
# Exit: 0 every selected row PASS; 1 any FAIL; 64 usage; 65 refused (tree differs from HEAD, unsafe path).
set -euo pipefail

REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)"
TERRAFORM_IMAGE="hashicorp/terraform:1.16.7"
# jq for LIFECYCLE-DRYRUN's real plan_scope_rows fixtures (the operator's jq; neither the host nor the harness has one).
JQ_IMAGE="ghcr.io/jqlang/jq:1.7.1@sha256:096b83865ad59b5b02841f103f83f45c51318394331bf1995e187ea3be937432"
# The harness's own tool pins (single source: infra/validation/lib.sh, read, never sourced).
pin_of() { sed -n -E "s/^$1=\"([^\"]+)\"$/\1/p" "$REPO/infra/validation/lib.sh"; }
YQ_IMAGE="$(pin_of YQ_IMAGE)" SHELLCHECK_IMAGE="$(pin_of SHELLCHECK_IMAGE)" HADOLINT_IMAGE="$(pin_of HADOLINT_IMAGE)"
ROWS=(TF-FMT TF-VALIDATE TF-TEST TF-GREP EDGE-ADAPT EDGE-FILTER LIFECYCLE-DRYRUN STAGING-GREP COMPOSE-STATIC
  START-ORDER HTTPS-CERT CF-TOKEN-CANARY METADATA-PROBE METADATA-GUARD CI-FEDERATION CI-ADAPTER ENV-LEASE CI-LAUNCHER
  SHELLCHECK-DIRECT HADOLINT-DIRECT)
BASE_DOMAIN="staging.example.net"

usage() { echo "usage: check.sh prepare --work DIR [--edge-tag g070-edge:TAG] [--platform OS_ARCH] [--netns-tag g070-netns:TAG [--debian-image debian:12@sha256:DIGEST]] | run --work DIR --report DIR --edge-image REF [--only ROW[,ROW]...] [--env-name NAME] [--netns-image REF] [--allow-dirty] [--cases ROW=CASE[,CASE]...]..." >&2; exit 64; }
die() { echo "check.sh: $*" >&2; exit 65; }
log() { printf '[check %s] %s\n' "$(date -u +%H:%M:%S)" "$*" >&2; }

MODE="${1:-}"; [[ "$MODE" == prepare || "$MODE" == run ]] || usage; shift
WORK="" REPORT="" EDGE_IMAGE="" EDGE_TAG="" ONLY="" ENV_NAME="" ALLOW_DIRTY=0 PLATFORM=linux_amd64
declare -A FIXTURE_CASES=()
NETNS_TAG="" NETNS_IMAGE="" DEBIAN_IMAGE="debian:12"
while [[ $# -gt 0 ]]; do
  case "$1" in
    --work) [[ $# -ge 2 ]] || usage; WORK="$2"; shift 2 ;;
    --report) [[ $# -ge 2 && "$MODE" == run ]] || usage; REPORT="$2"; shift 2 ;;
    --edge-image) [[ $# -ge 2 && "$MODE" == run ]] || usage; EDGE_IMAGE="$2"; shift 2 ;;
    --only) [[ $# -ge 2 && "$MODE" == run ]] || usage; ONLY="$2"; shift 2 ;;
    --env-name) [[ $# -ge 2 && "$MODE" == run && "$2" =~ ^[a-z][a-z0-9-]{1,20}$ ]] || usage; ENV_NAME="$2"; shift 2 ;;
    --allow-dirty) [[ "$MODE" == run ]] || usage; ALLOW_DIRTY=1; shift ;;
    # G072: a case selection passed as --only to one fixture row (a targeted rerun; the ledger records the subset).
    --cases) [[ $# -ge 2 && "$MODE" == run && "$2" =~ ^(CI-ADAPTER|ENV-LEASE|CI-LAUNCHER)=[a-z0-9][a-z0-9,-]*$ ]] || usage
      FIXTURE_CASES[${2%%=*}]="${2#*=}"; shift 2 ;;
    --edge-tag) [[ $# -ge 2 && "$MODE" == prepare && "$2" =~ ^g070-edge:[a-z0-9][a-z0-9._-]{0,63}$ ]] || usage; EDGE_TAG="$2"; shift 2 ;;
    --platform) [[ $# -ge 2 && "$MODE" == prepare && "$2" =~ ^[a-z0-9]+_[a-z0-9]+$ ]] || usage; PLATFORM="$2"; shift 2 ;;
    --netns-tag) [[ $# -ge 2 && "$MODE" == prepare && "$2" =~ ^g070-netns:[a-z0-9][a-z0-9._-]{0,63}$ ]] || usage; NETNS_TAG="$2"; shift 2 ;;
    --debian-image) [[ $# -ge 2 && "$MODE" == prepare && "$2" =~ ^debian:12(@sha256:[0-9a-f]{64})?$ ]] || usage; DEBIAN_IMAGE="$2"; shift 2 ;;
    --netns-image) [[ $# -ge 2 && "$MODE" == run ]] || usage; NETNS_IMAGE="$2"; shift 2 ;;
    *) usage ;;
  esac
done
[[ -n "$WORK" ]] || usage
[[ -n "$YQ_IMAGE" && -n "$SHELLCHECK_IMAGE" && -n "$HADOLINT_IMAGE" ]] || die "tool pins not found in infra/validation/lib.sh"
if [[ "$MODE" == run ]]; then
  [[ -n "$REPORT" && -n "$EDGE_IMAGE" ]] || usage
  if [[ -n "$ONLY" ]]; then
    IFS=, read -ra sel <<<"$ONLY"
    for r in "${sel[@]}"; do [[ " ${ROWS[*]} " == *" $r "* ]] || { echo "--only: unknown row $r" >&2; exit 64; }; done
  fi
fi

# A path inside this checkout must be Git-ignored (the main checkout's .local/); anything else would change the tree.
safe_dir() {
  mkdir -p "$1"; local p; p="$(cd "$1" && pwd -P)"
  if [[ "$p/" == "$REPO/"* ]] && ! git -C "$REPO" check-ignore -q "${p#"$REPO"/}/"; then die "$1 is inside the checkout and not ignored"; fi
  printf '%s\n' "$p"
}
WORK="$(safe_dir "$WORK")"
RUN_ID="$(date -u +%Y%m%dt%H%M%Sz)-$(head -c 4 /dev/urandom | od -An -tx1 | tr -d ' \n')"
LABELS=(--label g070.task=G070 --label "g070.check.run=$RUN_ID")
UIDGID="$(id -u):$(id -g)"
COMMIT="$(git -C "$REPO" rev-parse HEAD)"
random_hex() { head -c "$1" /dev/urandom | od -An -tx1 | tr -d ' \n'; }
q() { local out="" a; for a in "$@"; do out+="$(printf '%q' "$a") "; done; printf '%s' "${out% }"; }
# The tracked files under the given paths, copied from the working tree (equal to HEAD unless --allow-dirty).
copy_tracked() { # DEST PATH...
  local dest="$1"; shift; mkdir -p "$dest"
  (cd "$REPO" && git ls-files -z -- "$@" | tar --null -T - -cf -) | tar -xf - -C "$dest"
}

# ---- prepare ---------------------------------------------------------------------------------------------------------
if [[ "$MODE" == prepare ]]; then
  out="$WORK/prepare.tsv"; printf 'image\tid\trepo_digest\n' >"$out"
  for img in "$TERRAFORM_IMAGE" "$JQ_IMAGE" "$YQ_IMAGE" "$SHELLCHECK_IMAGE" "$HADOLINT_IMAGE"; do
    log "pull $img"; docker pull -q "$img" >/dev/null
    printf '%s\t%s\t%s\n' "$img" "$(docker image inspect --format '{{.Id}}' "$img")" \
      "$(docker image inspect --format '{{join .RepoDigests " "}}' "$img")" >>"$out"
  done
  pw="$WORK/prepare-$RUN_ID"; mkdir -p "$pw/home" "$WORK/tf-mirror"
  trap 'rm -rf "$pw"' EXIT
  copy_tracked "$pw" infra/terraform
  [[ -d "$pw/infra/terraform/environments/example" ]] || die "infra/terraform/environments/example is missing"
  log "terraform providers mirror -platform=$PLATFORM (network)"
  # shellcheck disable=SC2016 # expanded by the container's sh
  docker run --rm "${LABELS[@]}" --user "$UIDGID" --entrypoint /bin/sh -v "$pw:/work" -v "$WORK/tf-mirror:/mirror" \
    -e HOME=/work/home -e CHECKPOINT_DISABLE=1 -e TF_IN_AUTOMATION=1 -e "PLATFORM=$PLATFORM" "$TERRAFORM_IMAGE" \
    -c 'cd /work/infra/terraform/environments/example && terraform init -backend=false -input=false >/dev/null && terraform providers mirror -platform="$PLATFORM" /mirror'
  if [[ -n "$EDGE_TAG" ]]; then
    log "docker build $EDGE_TAG (build slot)"
    docker build -q "${LABELS[@]}" -f "$REPO/infra/docker/edge.Dockerfile" -t "$EDGE_TAG" "$REPO/infra/docker" >/dev/null
    printf '%s\t%s\t%s\n' "$EDGE_TAG" "$(docker image inspect --format '{{.Id}}' "$EDGE_TAG")" "local build at $COMMIT" >>"$out"
  fi
  if [[ -n "$NETNS_TAG" ]]; then
    log "pull $DEBIAN_IMAGE; docker build $NETNS_TAG (build slot)"
    docker pull -q "$DEBIAN_IMAGE" >/dev/null
    printf '%s\t%s\t%s\n' "$DEBIAN_IMAGE" "$(docker image inspect --format '{{.Id}}' "$DEBIAN_IMAGE")" \
      "$(docker image inspect --format '{{join .RepoDigests " "}}' "$DEBIAN_IMAGE")" >>"$out"
    # shellcheck disable=SC2016 # ${BASE} is a Dockerfile build argument
    printf '%s\n' 'ARG BASE' 'FROM ${BASE}' \
      'RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends iptables iproute2 busybox && rm -rf /var/lib/apt/lists/*' |
      docker build -q "${LABELS[@]}" --build-arg "BASE=$DEBIAN_IMAGE" -t "$NETNS_TAG" - >/dev/null
    printf '%s\t%s\t%s\n' "$NETNS_TAG" "$(docker image inspect --format '{{.Id}}' "$NETNS_TAG")" "local build from $DEBIAN_IMAGE (+ iptables iproute2 busybox)" >>"$out"
  fi
  cat "$out"
  exit 0
fi

# ---- run: setup ------------------------------------------------------------------------------------------------------
if [[ -n "$(git -C "$REPO" status --porcelain --untracked-files=all -- infra docs/foundation README.md .env.example)" ]]; then
  [[ $ALLOW_DIRTY == 1 ]] || die "infra/, docs/foundation, README.md or .env.example differ from HEAD; commit first or pass --allow-dirty"
  COMMIT="$COMMIT+dirty"
fi
REPORT="$(safe_dir "$REPORT")"
[[ ! -e "$REPORT/ledger.tsv" ]] || die "$REPORT/ledger.tsv exists; use a new report directory per run"
LEDGER="$REPORT/ledger.tsv" RES="$REPORT/resources.tsv"
printf 'row\tresult\texit\tcommit\tinputs_sha256\tcommand\tdetail\n' >"$LEDGER"
printf 'kind\tname\tcreated_utc\tremoved_utc\tstatus\n' >"$RES"
S="$WORK/check-$RUN_ID"; (umask 077; mkdir -p "$S")
printf 'scratch\t%s\t%s\t\tallocated\n' "$S" "$(date -u +%FT%TZ)" >>"$RES"
FILTER_NAME="g070-check-$RUN_ID-edge-filter" FILTER_STARTED=0
SCRUB=() # generated values never copied into the report

cleanup() {
  local rc=$? ids n
  if (( FILTER_STARTED )); then docker rm -f "$FILTER_NAME" >/dev/null 2>&1 || true; fi
  ids="$(docker ps -aq --filter "label=g070.check.run=$RUN_ID" 2>/dev/null || true)"
  n="$(grep -c . <<<"$ids" || true)"
  [[ -z "$ids" ]] || xargs -r docker rm -f <<<"$ids" >/dev/null 2>&1 || true
  if [[ -n "$(docker ps -aq --filter "label=g070.check.run=$RUN_ID" 2>/dev/null || true)" ]]; then
    printf 'container\tlabel g070.check.run=%s\t\t\tLEFTOVER (removal failed)\n' "$RUN_ID" >>"$RES"; rc=1
  else
    printf 'container\tlabel g070.check.run=%s\t\t%s\tabsent (swept %s)\n' "$RUN_ID" "$(date -u +%FT%TZ)" "$n" >>"$RES"
  fi
  rm -rf "$S" 2>/dev/null || true
  if [[ -e "$S" ]]; then printf 'scratch\t%s\t\t\tLEFTOVER\n' "$S" >>"$RES"; rc=1
  else printf 'scratch\t%s\t\t%s\tremoved\n' "$S" "$(date -u +%FT%TZ)" >>"$RES"; fi
  exit "$rc"
}
trap cleanup EXIT

selected() { [[ -z "$ONLY" || ",$ONLY," == *",$1,"* ]]; }
FAILS=0
ledger() { # ROW RESULT EXIT INPUTS_SHA COMMAND DETAIL
  local d="${6//$'\t'/ }"; d="${d//$'\n'/ }"; local c="${5//$'\t'/ }"; c="${c//$'\n'/ ; }"
  printf '%s\t%s\t%s\t%s\t%s\t%s\t%s\n' "$1" "$2" "$3" "$COMMIT" "$4" "$c" "$d" >>"$LEDGER"
  printf '%s %s (exit %s) %s\n' "$1" "$2" "$3" "$d"
  [[ "$2" != FAIL ]] || FAILS=$((FAILS + 1)) # NOT-SELECTED rows are reported, never counted as passes or failures
}
inputs_sha() { (cd "$REPO" && git ls-files -z -- "$@" | LC_ALL=C sort -z | xargs -0 -r sha256sum | sha256sum | cut -d' ' -f1); }
scrub_copy() { # SRC DEST: copy with every generated value replaced
  local v; cp "$1" "$2"
  for v in "${SCRUB[@]}"; do [[ -z "$v" ]] || sed -i "s/$v/G070-REDACTED/g" "$2"; done
}
image_id() { docker image inspect --format '{{.Id}}' "$1" 2>/dev/null || true; }
yq_json() { docker run --rm -i --pull never --network none "${LABELS[@]}" "$YQ_IMAGE" -p json -o tsv "$1"; }
has_line() { grep -qxF -- "$1" <<<"$2"; }

# ---- Terraform rows (r1.2; clarification 1: TF_DATA_DIR is exported inside each shell command) -----------------------
TF_INPUTS="$(inputs_sha infra/terraform)"
if selected TF-FMT || selected TF-VALIDATE || selected TF-TEST; then
  T="$S/tf"; mkdir -p "$T/data" "$T/out" "$T/home"
  copy_tracked "$T" infra/terraform
  TF_PRE=""
  if [[ -d "$T/infra/terraform" ]]; then mv "$T/infra/terraform" "$T/src"; rmdir "$T/infra"
  else mkdir -p "$T/src"; TF_PRE="no tracked infra/terraform files"; fi
  printf '%s\n' 'provider_installation {' '  filesystem_mirror { path = "/mirror" }' '  direct { exclude = ["*/*"] }' '}' >"$T/cli.tfrc"
  TF_RUN=(docker run --rm --pull never --network none "${LABELS[@]}" --user "$UIDGID" --entrypoint /bin/sh
    -v "$T:/work" -v "$WORK/tf-mirror:/mirror:ro" -e TF_CLI_CONFIG_FILE=/work/cli.tfrc -e TF_IN_AUTOMATION=1
    -e TF_INPUT=0 -e CHECKPOINT_DISABLE=1 -e HOME=/work/home "$TERRAFORM_IMAGE")
  [[ -n "$TF_PRE" || -n "$(image_id "$TERRAFORM_IMAGE")" ]] || TF_PRE="terraform image $TERRAFORM_IMAGE not prepared"
  [[ -n "$TF_PRE" ]] || compgen -G "$WORK/tf-mirror/*" >/dev/null || TF_PRE="provider mirror $WORK/tf-mirror missing or empty (run prepare)"
  TF_ID="image $(image_id "$TERRAFORM_IMAGE")"
  mkdir -p "$REPORT/terraform"
fi

if selected TF-FMT; then
  c='terraform fmt -check -recursive -diff /work/src'
  if [[ -n "$TF_PRE" ]]; then ledger TF-FMT FAIL - "$TF_INPUTS" "$c" "$TF_PRE"
  else
    rc=0; "${TF_RUN[@]}" -c "$c" >"$T/out/fmt.log" 2>&1 || rc=$?
    cp "$T/out/fmt.log" "$REPORT/terraform/fmt.log"
    [[ $rc == 0 ]] && ledger TF-FMT PASS 0 "$TF_INPUTS" "$(q "${TF_RUN[@]}") -c '$c'" "$TF_ID; every file formatted" \
      || ledger TF-FMT FAIL "$rc" "$TF_INPUTS" "$(q "${TF_RUN[@]}") -c '$c'" "$TF_ID; unformatted files (terraform/fmt.log)"
  fi
fi

if selected TF-VALIDATE; then
  c='export TF_DATA_DIR=/work/data/root; cd /work/src/environments/example && terraform init -backend=false -input=false -lockfile=readonly >/work/out/root-init.log 2>&1 && terraform validate -json >/work/out/root-validate.json'
  if [[ -n "$TF_PRE" ]]; then ledger TF-VALIDATE FAIL - "$TF_INPUTS" "$c" "$TF_PRE"
  elif [[ ! -d "$T/src/environments/example" ]]; then ledger TF-VALIDATE FAIL - "$TF_INPUTS" "$c" "infra/terraform/environments/example missing"
  else
    rc=0; "${TF_RUN[@]}" -c "$c" || rc=$?
    for f in root-init.log root-validate.json; do [[ ! -f "$T/out/$f" ]] || cp "$T/out/$f" "$REPORT/terraform/$f"; done
    if [[ $rc == 0 ]] && grep -qE '"valid":[[:space:]]*true' "$T/out/root-validate.json"; then
      ledger TF-VALIDATE PASS 0 "$TF_INPUTS" "$(q "${TF_RUN[@]}") -c '$c'" "$TF_ID; committed lock matches the mirror; root valid"
    else
      ledger TF-VALIDATE FAIL "$rc" "$TF_INPUTS" "$(q "${TF_RUN[@]}") -c '$c'" "$TF_ID; init or validate failed (terraform/root-init.log, root-validate.json)"
    fi
  fi
fi

if selected TF-TEST; then
  cmds="" detail="" total=0 rc_all=0 ok=1 nmod=0
  if [[ -n "$TF_PRE" ]]; then ok=0 detail="$TF_PRE" rc_all=-
  else
    for d in "$T/src/modules"/*/; do
      [[ -d "$d" ]] || continue
      m="$(basename "$d")"
      [[ "$m" =~ ^[a-z0-9][a-z0-9-]*$ ]] || { ok=0; detail+="unexpected module name $m; "; continue; }
      nmod=$((nmod + 1))
      if ! compgen -G "$d/tests/*.tftest.hcl" >/dev/null; then ok=0; detail+="$m: no tests/*.tftest.hcl; "; continue; fi
      c="export TF_DATA_DIR=/work/data/$m; cd /work/src/modules/$m && terraform init -backend=false -input=false >/work/out/$m-init.log 2>&1 && terraform validate -json >/work/out/$m-validate.json && terraform test -json -no-color >/work/out/$m-test.json"
      cmds+="$(q "${TF_RUN[@]}") -c '$c' ; "
      rc=0; "${TF_RUN[@]}" -c "$c" || rc=$?
      [[ $rc == 0 ]] || rc_all=$rc
      for f in "$m-init.log" "$m-validate.json" "$m-test.json"; do [[ ! -f "$T/out/$f" ]] || cp "$T/out/$f" "$REPORT/terraform/$f"; done
      sum="$(grep -E '"type":[[:space:]]*"test_summary"' "$T/out/$m-test.json" 2>/dev/null || true)"
      if [[ "$(grep -c . <<<"$sum" || true)" != 1 ]]; then ok=0; detail+="$m: exit $rc, no single test_summary (FAIL, never an offline pass); "; continue; fi
      sum="$(sed -n -E 's/.*"test_summary":[[:space:]]*\{([^}]*)\}.*/\1/p' <<<"$sum")"
      num() { sed -n -E "s/.*\"$1\":[[:space:]]*([0-9]+).*/\1/p" <<<"$sum"; }
      st="$(sed -n -E 's/.*"status":[[:space:]]*"([a-z]+)".*/\1/p' <<<"$sum")"
      p="$(num passed)" f="$(num failed)" e="$(num errored)" s="$(num skipped)"
      if [[ $rc == 0 && "$st" == pass && "${p:-0}" -ge 1 && "$f" == 0 && "$e" == 0 && "$s" == 0 ]]; then
        total=$((total + p)); detail+="$m: passed=$p; "
      else ok=0; detail+="$m: exit $rc status=${st:-?} passed=${p:-?} failed=${f:-?} errored=${e:-?} skipped=${s:-?}; "; fi
    done
    (( nmod >= 2 )) || { ok=0; detail+="modules=$nmod (the Plan requires the VM and DNS module tests); "; }
  fi
  [[ $ok == 1 ]] && ledger TF-TEST PASS 0 "$TF_INPUTS" "${cmds% ; }" "$TF_ID; modules=$nmod runs=$total passed=$total; ${detail% ; }" \
    || ledger TF-TEST FAIL "$rc_all" "$TF_INPUTS" "${cmds% ; }" "${detail% ; }"
fi

if selected TF-GREP; then
  hits="" files=()
  while IFS= read -r f; do
    case "$f" in */tests/*|*.tftest.hcl|*.terraform.lock.hcl) continue ;; *.tf|*.hcl|*.sh) files+=("$f") ;; esac
  done < <(cd "$REPO" && git ls-files -- infra/terraform)
  scan() { # LABEL ERE [tf-only]
    local f
    for f in "${files[@]}"; do
      [[ "${3:-}" != tf || "$f" == *.tf ]] || continue
      hits+="$(grep -nEi -- "$2" "$REPO/$f" | grep -vE '^[0-9]+:[[:space:]]*(#|//)' | sed "s|^|$f:|; s|\$| [$1]|" || true)"$'\n'
    done
  }
  scan role 'roles/(owner|editor|resourcemanager\.projectIamAdmin|iam\.serviceAccountKeyAdmin|iam\.securityAdmin)([^A-Za-z0-9.]|$)'
  scan sa-key 'google_service_account_key|serviceAccountKey|service-accounts[[:space:]]+keys[[:space:]]+create'
  scan delete-action '=[[:space:]]*"DELETE"'
  scan provisioner '(^|[^A-Za-z_])provisioner([^A-Za-z_]|$)|local-exec|remote-exec|metadata_startup_script'
  scan secret-manager 'google_secret_manager_' tf
  scan secret-variable '^[[:space:]]*variable[[:space:]]+"[^"]*(secret|token|password|api_key|private_key)[^"]*"' tf
  scan secret-output '^[[:space:]]*output[[:space:]]+"[^"]*(secret|token|password|key)[^"]*"' tf
  # Session length (A2 fallback): the only lifecycle ignore_changes in infra/terraform is gcp-staging-vm's
  # google_compute_instance.vm ignoring exactly scheduling[0].max_run_duration (nothing broader); creation still sets
  # max_run_duration from var.session_max_seconds.
  vmtf=infra/terraform/modules/gcp-staging-vm/instance.tf ig=""
  for f in "${files[@]}"; do
    [[ "$f" == *.tf ]] || continue
    ig+="$(grep -nE 'ignore_changes' "$REPO/$f" | grep -vE '^[0-9]+:[[:space:]]*(#|//)' | sed "s|^|$f:|" || true)"$'\n'
  done
  ig="$(grep . <<<"$ig" || true)"
  [[ "$(grep -c . <<<"$ig" || true)" == 1 && "$ig" =~ ^$vmtf:[0-9]+:[[:space:]]*ignore_changes[[:space:]]*=[[:space:]]*\[scheduling\[0\]\.max_run_duration\][[:space:]]*$ ]] ||
    hits+="${ig//$'\n'/ | } [ignore_changes must be exactly one line, ignore_changes = [scheduling[0].max_run_duration] in $vmtf]"$'\n'
  inres="$(awk '/^resource "google_compute_instance" "vm"/ {r = 1} r && /^[[:space:]]*lifecycle[[:space:]]*\{/ {l = 1}
    r && l && /^[[:space:]]*ignore_changes[[:space:]]*=/ {print "ignore"} r && /^[[:space:]]*seconds[[:space:]]*=[[:space:]]*var\.session_max_seconds[[:space:]]*$/ {print "seconds"}
    r && /^}/ {r = 0; l = 0}' "$REPO/$vmtf" | sort | tr '\n' ' ')"
  [[ "$inres" == "ignore seconds " ]] ||
    hits+="$vmtf: google_compute_instance.vm lacks its lifecycle ignore_changes or max_run_duration seconds = var.session_max_seconds (found: $inres) [session length writer]"$'\n'
  hits="$(grep . <<<"$hits" || true)"
  c="grep -nEi <guards> over ${#files[@]} non-test infra/terraform files (comment lines excluded); ignore_changes scope"
  if (( ${#files[@]} == 0 )); then ledger TF-GREP FAIL - "$TF_INPUTS" "$c" "no infra/terraform source files"
  elif [[ -z "$hits" ]]; then ledger TF-GREP PASS 0 "$TF_INPUTS" "$c" "no forbidden role, key, DELETE action, provisioner, metadata_startup_script, Secret Manager resource or secret-shaped variable/output; the only ignore_changes is [scheduling[0].max_run_duration] on google_compute_instance.vm, which still sets seconds = var.session_max_seconds"
  else printf '%s\n' "$hits" >"$REPORT/tf-grep-hits.txt"; ledger TF-GREP FAIL 1 "$TF_INPUTS" "$c" "$(grep -c . <<<"$hits") hits (tf-grep-hits.txt)"; fi
fi

# ---- Edge rows (P0.2; A3, A4) ----------------------------------------------------------------------------------------
EDGE_INPUTS="$(inputs_sha infra/edge infra/docker/edge.Dockerfile)"
if selected EDGE-ADAPT || selected EDGE-FILTER; then
  CANARY="g070canary$(random_hex 16)"; SCRUB+=("$CANARY")
  printf '%s\n' "EDGE_BASE_DOMAIN=$BASE_DOMAIN" "ACME_EMAIL=ops@example.net" \
    "EDGE_ACME_CA=https://acme-staging-v02.api.letsencrypt.org/directory" "CF_API_TOKEN=$CANARY" >"$S/edge.env"
  EDGE_ID="$(image_id "$EDGE_IMAGE")"
  EDGE_RUN=(docker run --rm --pull never --network none "${LABELS[@]}")
  mkdir -p "$REPORT/edge"
fi

if selected EDGE-ADAPT; then
  ok=1 why=""; fail() { ok=0; why+="$1; "; }
  c="$(q "${EDGE_RUN[@]}") $EDGE_IMAGE caddy list-modules ; $(q "${EDGE_RUN[@]}" -v "$REPO/infra/edge:/edge:ro" --env-file "<scratch>/edge.env") $EDGE_IMAGE caddy adapt --config /edge/Caddyfile --adapter caddyfile --validate --pretty"
  rc=0
  if [[ -z "$EDGE_ID" ]]; then fail "edge image $EDGE_IMAGE not present locally (prepare --edge-tag)"; rc=1
  elif [[ ! -f "$REPO/infra/edge/Caddyfile" ]]; then fail "infra/edge/Caddyfile missing"; rc=1
  else
    mods="$("${EDGE_RUN[@]}" "$EDGE_IMAGE" caddy list-modules 2>&1 || true)"
    ver="$("${EDGE_RUN[@]}" "$EDGE_IMAGE" caddy version 2>&1 | head -1 || true)"
    printf '%s\n' "$ver" "$mods" >"$REPORT/edge/list-modules.txt"
    grep -qxE '[[:space:]]*dns\.providers\.cloudflare[[:space:]]*' <<<"$mods" || fail "dns.providers.cloudflare not in caddy list-modules"
    "${EDGE_RUN[@]}" -v "$REPO/infra/edge:/edge:ro" --env-file "$S/edge.env" "$EDGE_IMAGE" \
      caddy adapt --config /edge/Caddyfile --adapter caddyfile --validate --pretty >"$S/adapt.json" 2>"$S/adapt.err" || rc=$?
    scrub_copy "$S/adapt.err" "$REPORT/edge/adapt.stderr.txt"
    if [[ $rc != 0 ]]; then fail "caddy adapt --validate exit $rc under network none (A3 finding; edge/adapt.stderr.txt)"
    else
      grep -qF -- "$CANARY" "$S/adapt.json" "$S/adapt.err" && fail "the CF_API_TOKEN value reached the adapted config or its output (must stay {env.CF_API_TOKEN})"
      grep -qF -- 'replace-me' "$S/adapt.json" && fail "literal replace-me in the adapted config"
      scrub_copy "$S/adapt.json" "$REPORT/edge/edge-adapt.json"
      sed -n '/^{$/,/^}$/p' "$S/adapt.json" >"$S/adapt.doc.json"
      J="$S/adapt.doc.json"
      [[ "$(yq_json '.admin.disabled // false' <"$J")" == true ]] || fail "admin API not disabled (admin off)"
      grep -qF 'http://core:8180/internal/v1/domains/tls-ask' "$J" || fail "on-demand ask URL http://core:8180/internal/v1/domains/tls-ask absent"
      pol="$(yq_json '.apps.tls.automation.policies // [] | .[] | [(.on_demand // false), ((.subjects // []) | join(" "))]' <"$J")"
      grep -qE '^true' <<<"$pol" || fail "no on-demand policy (custom hosts)"
      grep -qE $'^true\t.+' <<<"$pol" && fail "on_demand enabled on a policy with subjects (only the catch-all may use it)"
      wild="$(yq_json ".apps.tls.automation.policies // [] | .[] | select((.subjects // []) | contains([\"$BASE_DOMAIN\", \"*.$BASE_DOMAIN\"])) | (.issuers // []) | .. | select(tag == \"!!map\") | select(has(\"provider\")) | (.provider.name // \"\")" <"$J")"
      has_line cloudflare "$wild" || fail "no DNS-01 cloudflare provider on the policy for $BASE_DOMAIN and *.$BASE_DOMAIN"
      toks="$(yq_json '.. | select(tag == "!!map") | select(has("api_token")) | .api_token' <"$J")"
      [[ -n "$toks" ]] || fail "no DNS provider api_token reference"
      grep -qvE '^\{env\.[A-Z0-9_]+\}$' <<<"$toks" && fail "a DNS provider token is not an {env.*} reference"
      logs="$(yq_json '.logging.logs // {} | to_entries | .[] | [.key, (.value.writer.output // "stderr"), (.value.encoder.format // "-"), ((.value.include // []) | map(select(test("^http[.]log[.]access"))) | length), ((.value.encoder.fields // {}) | keys | join(","))]' <"$J")"
      nacc=0 ndisc=0
      while IFS=$'\t' read -r name out fmt inc fields; do
        [[ "${inc:-0}" -ge 1 ]] || continue
        nacc=$((nacc + 1))
        if [[ "$out" == discard ]]; then ndisc=$((ndisc + 1)); continue; fi
        [[ "$fmt" == filter ]] || { fail "access logger $name is not a filter encoder"; continue; }
        for k in 'request>uri' 'request>headers>Cookie' 'request>headers>Authorization' 'resp_headers>Set-Cookie' 'resp_headers>Location'; do
          [[ ",$fields," == *",$k,"* ]] || fail "access logger $name has no $k filter"
        done
      done <<<"$logs"
      (( nacc >= 1 )) || fail "no access logger (site log directive)"
      rp="$(yq_json '.. | select(tag == "!!map") | select(.handler == "reverse_proxy") | ("del:" + ((.headers.request.delete // []) | join(",")))' <"$J")"
      nrp="$(grep -c '^del:' <<<"$rp" || true)"
      (( nrp >= 2 )) || fail "expected reverse_proxy handlers for web and core, found $nrp"
      while IFS= read -r del; do
        [[ "$del" == del:* ]] || continue
        del="${del#del:}"
        [[ ",$del," == *",X-Forwarded-*,"* || ( ",$del," == *",X-Forwarded-For,"* && ",$del," == *",X-Forwarded-Proto,"* && ",$del," == *",X-Forwarded-Host,"* ) ]] \
          || fail "a reverse_proxy does not delete X-Forwarded-For/-Proto/-Host upstream ($del)"
      done <<<"$rp"
      dials="$(yq_json '.. | select(tag == "!!map") | select(.handler == "reverse_proxy") | (.upstreams // []) | .[] | .dial' <"$J")"
      has_line core:8180 "$dials" || fail "no upstream core:8180"
      has_line web:3000 "$dials" || fail "no upstream web:3000"
      paths="$(yq_json '.. | select(tag == "!!map") | select(has("path")) | .path | select(tag == "!!seq") | .[]' <"$J")"
      has_line '/mcp/c/*' "$paths" || fail "no /mcp/c/* path matcher"
      has_line '/.well-known/oauth-protected-resource/mcp/c/*' "$paths" || fail "no /.well-known/oauth-protected-resource/mcp/c/* path matcher"
      grep -qE '^/internal' <<<"$paths" && fail "an /internal path matcher exists (internal routes must stay unreachable)"
      hsts="$(yq_json '.. | select(tag == "!!map") | select(has("Strict-Transport-Security")) | .["Strict-Transport-Security"] | .[]' <"$J")"
      [[ -n "$hsts" ]] || fail "no Strict-Transport-Security header"
      grep -qiE 'includesubdomains|preload' <<<"$hsts" && fail "HSTS carries includeSubDomains or preload"
      why+="access loggers=$nacc (discard $ndisc); reverse_proxy handlers=$nrp; HSTS ${hsts//$'\n'/,}; "
    fi
  fi
  [[ $ok == 1 ]] && ledger EDGE-ADAPT PASS 0 "$EDGE_INPUTS" "$c" "image $EDGE_ID; ${ver:-}; network none; ${why% ; }" \
    || ledger EDGE-ADAPT FAIL "$rc" "$EDGE_INPUTS" "$c" "image ${EDGE_ID:-absent}; ${why% ; }"
fi

if selected EDGE-FILTER; then
  ok=1 why="" rc=0; fail() { ok=0; why+="$1; "; }
  F="$S/filter"; mkdir -p "$F"
  c="$(q docker run -d --pull never --network none "${LABELS[@]}" --name "$FILTER_NAME" --env-file "<scratch>/edge.env" -v "<scratch>/filter:/filter:ro") $EDGE_IMAGE caddy run --config /filter/Caddyfile.filter-test --adapter caddyfile ; docker exec $FILTER_NAME wget <sentinel requests> ; docker logs $FILTER_NAME | grep"
  if [[ -z "$EDGE_ID" ]]; then fail "edge image $EDGE_IMAGE not present locally"; rc=1
  elif [[ ! -f "$REPO/infra/edge/Caddyfile" ]]; then fail "infra/edge/Caddyfile missing"; rc=1
  else
    cp -R "$REPO/infra/edge/." "$F/"
    # The plain-HTTP loopback copy: global options and snippets verbatim; every site block re-addressed to
    # http://:<port> with its site-level tls directive removed (an HTTP site cannot carry TLS). Routes and log blocks
    # are unchanged; upstreams are unreachable with network none, so requests end in 502 and are still logged.
    awk -v meta="$S/filter.meta" -v open_re='(^|[ \t])[{]$' -v tls_re='^tls([ \t]|$)' -v log_re='^log([ \t]|$)' '
      function strip(s) { sub(/^[ \t]+/, "", s); if (s ~ /^#/) return ""; sub(/[ \t]+#.*$/, "", s); sub(/[ \t]+$/, "", s); return s }
      function opens(s) { return s ~ open_re }
      BEGIN { depth = 0; port = 8080; n = 0; skip = 0; logs = 0; discards = 0; logdepth = 0; pending = "" }
      {
        s = strip($0)
        if (depth == 0) {
          if (pending != "") s = pending " " s
          if (s == "") { print; next }
          if (s ~ /,$/) { pending = s; next }
          pending = ""
          if (s == "{") { kind = "global"; depth = 1; print; next }
          if (opens(s)) {
            if (s ~ /^\(/) { kind = "snippet"; print } else { kind = "site"; printf "http://:%d {\n", port; ports = ports " " port; port++; n++ }
            depth = 1; next
          }
          print; next
        }
        if (skip > 0) { if (opens(s)) skip++; else if (s == "}") skip--; next }
        if (depth == 1 && kind != "global" && s ~ tls_re) { if (opens(s)) skip = 1; next }
        if (kind != "global" && s ~ log_re) { logs++; if (opens(s)) { logdepth = depth + 1; cur = 0 } }
        if (logdepth > 0 && s ~ /^output[ \t]+discard$/) cur = 1
        if (s == "}") {
          if (logdepth > 0 && depth == logdepth) { if (cur) discards++; logdepth = 0 }
          depth--; print; next
        }
        if (opens(s)) depth++
        print
      }
      END { bad = (depth != 0) + (skip != 0) + (pending != ""); printf "%d %d %d%s\n", bad, logs, discards, ports > meta }
    ' "$REPO/infra/edge/Caddyfile" >"$F/Caddyfile.filter-test"
    read -r bad nlogs ndisc fports <"$S/filter.meta"
    read -ra ports <<<"${fports:-}"
    cp "$F/Caddyfile.filter-test" "$REPORT/edge/Caddyfile.filter-test"
    if [[ "$bad" != 0 ]]; then fail "unbalanced Caddyfile blocks; cannot derive the filter-test copy"; rc=1
    elif (( ${#ports[@]} == 0 )); then fail "no site block"; rc=1
    elif (( nlogs == 0 )); then fail "no site log directive: declare log { output discard } (A4 fallback) or a filtered access log"; rc=1
    else
      FILTER_STARTED=1
      printf 'container\t%s\t%s\t\tstarted\n' "$FILTER_NAME" "$(date -u +%FT%TZ)" >>"$RES"
      docker run -d --pull never --network none "${LABELS[@]}" --name "$FILTER_NAME" --env-file "$S/edge.env" \
        -v "$F:/filter:ro" "$EDGE_IMAGE" caddy run --config /filter/Caddyfile.filter-test --adapter caddyfile >/dev/null || rc=$?
      up=0
      for p in "${ports[@]}"; do
        for _ in $(seq 1 40); do
          o="$(docker exec "$FILTER_NAME" wget -q -S -O /dev/null "http://127.0.0.1:$p/" 2>&1 || true)" # any HTTP status = serving
          if grep -q 'HTTP/' <<<"$o"; then up=$((up + 1)); break; fi
          [[ "$(docker inspect --format '{{.State.Running}}' "$FILTER_NAME" 2>/dev/null)" == true ]] || break
          sleep 0.5
        done
      done
      reqs=('/auth/callback?code=G070SENTINEL-1&state=G070SENTINEL-2&error=G070SENTINEL-3&error_description=G070SENTINEL-4&other=G070SENTINEL-8'
        '/signup/continue?ref=G070SENTINEL-5' '/signup/prove?n=G070SENTINEL-6' '/signup/complete?p=G070SENTINEL-7'
        '/any/path?token=G070SENTINEL-9')
      if (( up != ${#ports[@]} )); then fail "caddy did not serve every test site ($up of ${#ports[@]}; docker run exit $rc; edge/filter.log)"
      else
        for p in "${ports[@]}"; do
          for r in "${reqs[@]}"; do
            docker exec "$FILTER_NAME" wget -q -O /dev/null --header 'Cookie: g070=G070SENTINEL-C' \
              --header 'Authorization: Bearer G070SENTINEL-A' "http://127.0.0.1:$p$r" >/dev/null 2>&1 || true
          done
        done
        sleep 1
      fi
      docker logs "$FILTER_NAME" >"$S/filter.log" 2>&1 || true
      docker rm -f "$FILTER_NAME" >/dev/null 2>&1 || true
      if [[ -z "$(docker ps -aq --filter "name=^$FILTER_NAME\$")" ]]; then
        printf 'container\t%s\t\t%s\tremoved\n' "$FILTER_NAME" "$(date -u +%FT%TZ)" >>"$RES"; FILTER_STARTED=0
      fi
      scrub_copy "$S/filter.log" "$REPORT/edge/filter.log"
      grep -qF -- "$CANARY" "$S/filter.log" && fail "the CF_API_TOKEN value appears in the log"
      sen="$(grep -c 'G070SENTINEL' "$S/filter.log" || true)"
      key="$(grep -cE '[?&](code|state|error|error_description|ref|n|p)=' "$S/filter.log" || true)"
      (( sen == 0 )) || fail "$sen log lines carry G070SENTINEL (A4 finding; edge/filter.log)"
      (( key == 0 )) || fail "$key log lines carry a sensitive query key"
      if (( ndisc == nlogs )); then why+="mode=discard (every site log discards; no hop lines by design); "
      elif (( up == ${#ports[@]} )); then
        miss=0
        for p in "${ports[@]}"; do
          hl="$(grep -E "\"host\":[[:space:]]*\"127\\.0\\.0\\.1:$p\"" "$S/filter.log" || true)"
          for path in /auth/callback /signup/continue /signup/prove /signup/complete /any/path; do
            pl="$(grep -F -- "$path" <<<"$hl" || true)"
            grep -qE '"status":[[:space:]]*[0-9]' <<<"$pl" || miss=$((miss + 1))
          done
        done
        (( miss == 0 )) || fail "$miss requests have no access log line with host, path and status"
        why+="mode=filtered; ${#ports[@]} sites x ${#reqs[@]} requests logged with host/path/status; "
      fi
      why+="sentinel hits=$sen sensitive-key hits=$key; "
    fi
  fi
  [[ $ok == 1 ]] && ledger EDGE-FILTER PASS 0 "$EDGE_INPUTS" "$c" "image $EDGE_ID; network none; ${why% ; }" \
    || ledger EDGE-FILTER FAIL "${rc:-1}" "$EDGE_INPUTS" "$c" "image ${EDGE_ID:-absent}; ${why% ; }"
fi

# ---- Lifecycle dry-run (P0.4) and static script guards (r1.3) --------------------------------------------------------
STAGING_INPUTS="$(inputs_sha infra/staging infra/compose/staging.yml)"
# make_shims: a dry run or a sourced script must execute none of these tools; each call is recorded in
# <scratch>/shim.log (exit 97) and fails the row that observes it.
make_shims() {
  local t
  mkdir -p "$S/shim" "$S/home"; : >"$S/shim.log"
  for t in gcloud gsutil terraform docker ssh scp curl sudo; do
    printf '#!/bin/sh\nprintf "%%s\\n" "%s $*" >>"%s"\nexit 97\n' "$t" "$S/shim.log" >"$S/shim/$t"; chmod 755 "$S/shim/$t"
  done
}
if selected LIFECYCLE-DRYRUN; then
  ok=1 why="" rc_all=0; fail() { ok=0; why+="$1; "; }
  conf="infra/staging/examples/staging.conf.example"
  [[ -n "$ENV_NAME" || ! -f "$REPO/$conf" ]] || ENV_NAME="$(sed -n -E "s/^STAGING_ENV=[\"']?([a-z][a-z0-9-]{1,20})[\"']?[[:space:]]*\$/\1/p" "$REPO/$conf" | head -1)"
  c="PATH=<shims>:\$PATH bash infra/staging/staging.sh <command> --config $conf --dry-run"
  if [[ ! -f "$REPO/infra/staging/staging.sh" || ! -f "$REPO/$conf" ]]; then fail "infra/staging/staging.sh or $conf missing"
  elif [[ -z "$ENV_NAME" ]]; then fail "no STAGING_ENV in $conf (or pass --env-name)"
  else
    make_shims; mkdir -p "$REPORT/lifecycle"
    # The --duration forms (Plan r1.9): a dry run with --duration shows start's stopped-VM change and provision's first
    # allocation; outputs lifecycle/<command>-duration-<D>.txt.
    cmds=("provision --candidate $COMMIT" "deploy --candidate $COMMIT" "start" "stop" "reset --confirm $ENV_NAME"
      "retire --yes" "status" "receipt" "build --candidate $COMMIT" "push --candidate $COMMIT"
      "start --duration 90m" "start --duration 2h" "start --duration 7200" "provision --candidate $COMMIT --duration 2h")
    [[ "$COMMIT" != *+dirty ]] || cmds=("${cmds[@]//+dirty/}")
    vol_re="^${ENV_NAME}_(postgres|minio)-data$"
    # G072 (Plan §10.5, §10.9): the environment's metadata bucket, the covered mutators and the printed lease lines.
    g72_proj="$(sed -n -E 's/^GCP_PROJECT=([a-z][a-z0-9-]+)$/\1/p' "$REPO/$conf" | head -1)"
    g72_sa="$(sed -n -E 's/^DEPLOYER_SA=([^[:space:]]+)$/\1/p' "$REPO/$conf" | head -1)"
    g72_mb="$g72_proj-$ENV_NAME-ci-meta" g72_cand="${COMMIT%+dirty}"
    g72_acquire="+ (gs://$g72_mb present: gcloud storage cp <lease: local:<user>@<host>:<pid>> gs://$g72_mb/lease --if-generation-match=0, released at exit; absent (404): no lease; any other answer: refused)"
    g72_fence="+ (gs://$g72_mb present: create-only gs://$g72_mb/schema/$g72_cand.tsv, then an entry appended to gs://$g72_mb/fence.tsv at its observed generation, before the upload)"
    g72_covered=" provision deploy start stop reset retire "
    for cmd in "${cmds[@]}"; do
      sub="${cmd%% *}" rc=0 name="${cmd%% *}"
      [[ "$cmd" != *" --duration "* ]] || name="$sub-duration-${cmd##* }"
      out="$REPORT/lifecycle/$name.txt"
      read -ra args <<<"$cmd"
      (cd "$REPO" && env PATH="$S/shim:$PATH" HOME="$S/home" CLOUDSDK_CONFIG="$S/home/gcloud" \
        bash infra/staging/staging.sh "${args[@]}" --config "$conf" --dry-run) >"$out" 2>&1 || rc=$?
      [[ $rc == 0 ]] || { rc_all=$rc; fail "$sub: exit $rc"; }
      grep -q . "$out" || fail "$sub: printed nothing"
      if grep -q . "$S/shim.log"; then fail "$sub: executed $(cut -d' ' -f1 "$S/shim.log" | sort -u | tr '\n' ' ')"; : >"$S/shim.log"; fi
      bad="$(grep -nE 'gcloud[^|;&]*[[:space:]](instances|disks|addresses|networks|firewall-rules)[[:space:]]+(create|delete)([[:space:]]|$)' "$out" || true)"
      bad+="$(grep -nE 'compose[^|;&]*[[:space:]]down([[:space:]][^|;&]*)?[[:space:]](-v|--volumes)([[:space:]=]|$)' "$out" || true)"
      bad+="$(grep -nE '(^|[[:space:]])(run|exec)[[:space:]][^|;&]*[[:space:]]migrate[[:space:]]+(down|reapply)([[:space:]]|$)|(^|[^A-Za-z_-])reapply([^A-Za-z_-]|$)' "$out" || true)"
      bad+="$(grep -nE '(volume|system|image|container|network)[[:space:]]+prune' "$out" || true)"
      bad+="$(grep -nE 'identity\.yml|browser\.yml|secret-scan\.yml' "$out" || true)"
      [[ -z "$bad" ]] || fail "$sub: forbidden command printed (${bad//$'\n'/ | })"
      [[ "$sub" == retire ]] || ! grep -qE 'terraform[^|;&]*[[:space:]](destroy|-destroy)([[:space:]]|$)' "$out" || fail "$sub: terraform destroy outside retire"
      # apply: provision's saved plan; in retire only the saved destroy plan its own `plan -destroy -out=FILE` wrote
      # (Plan §2.3: plan -destroy -> show -> destroy; applying that saved plan destroys exactly what label_check read);
      # in start --duration only the saved refresh-only plan its own `plan -refresh-only ... -out=FILE` wrote (no
      # normal plan at all: the provider would replace the VM for a length change).
      applies="$(grep -E 'terraform[^|;&]*[[:space:]]apply([[:space:]]|$)' "$out" || true)"
      if [[ "$sub" == retire || "$name" == start-duration-* ]]; then
        if [[ "$sub" == retire ]]; then
          dplans="$(grep -E 'terraform[^|;&]*[[:space:]]plan([[:space:]][^|;&]*)?[[:space:]]-destroy([[:space:]]|$)' "$out" | grep -oE -- '[[:space:]]-out=[^[:space:]]+' | sed -E 's/^[[:space:]]-out=//' || true)"
        else
          dplans="$(grep -E 'terraform[^|;&]*[[:space:]]plan[[:space:]]-refresh-only[[:space:]][^|;&]*[[:space:]]session_max_seconds=[0-9]+[[:space:]]' "$out" | grep -oE -- '[[:space:]]-out=[^[:space:]]+' | sed -E 's/^[[:space:]]-out=//' || true)"
          [[ -n "$applies" ]] || fail "$name: no terraform apply of the refresh-only plan"
          ! grep -E 'terraform[^|;&]*[[:space:]]plan[[:space:]]' "$out" | grep -qvE '[[:space:]]plan[[:space:]]-refresh-only[[:space:]]' ||
            fail "$name: a terraform plan other than the refresh-only plan"
        fi
        while IFS= read -r l; do
          [[ -n "$l" ]] || continue
          [[ -n "$dplans" ]] && has_line "${l##* }" "$dplans" || fail "$name: terraform apply of anything but its own saved plan (${l##* })"
        done <<<"$applies"
      elif [[ "$sub" != provision && -n "$applies" ]]; then fail "$name: terraform apply outside provision"; fi
      [[ "$sub" == reset ]] || ! grep -qE 'compose[^|;&]*[[:space:]]down([[:space:]]|$)' "$out" || fail "$sub: compose down outside reset"
      vols="$(grep -oE 'volume[[:space:]]+rm[[:space:]]+[^|;&]*' "$out" | sed -E 's/^volume[[:space:]]+rm[[:space:]]+//' | tr -d "\"'\\\\" | tr ' ' '\n' | grep -vE '^(-f|--force)?$' | sort -u || true)"
      if [[ "$sub" == reset ]]; then
        [[ "$(grep -cE "$vol_re" <<<"$vols" || true)" == 2 && -z "$(grep -vE "$vol_re" <<<"$vols" || true)" ]] \
          || fail "reset: volume rm must name exactly ${ENV_NAME}_postgres-data and ${ENV_NAME}_minio-data (found: ${vols//$'\n'/ })"
      elif [[ -n "$vols" ]]; then fail "$sub: removes volumes (${vols//$'\n'/ })"; fi
      # G072 lease lines: a covered mutator's first two commands are the presence check and the lease create; the
      # others print nothing about the metadata bucket; deploy and provision print the fence entry before the upload.
      if [[ "$g72_covered" == *" $sub "* ]]; then
        mapfile -t g72_first < <(grep -m2 '^+ ' "$out" || true)
        [[ "${g72_first[0]:-}" == "+ gcloud storage buckets describe gs://$g72_mb --raw --format value\(name\,projectNumber\) --project $g72_proj --impersonate-service-account $g72_sa" &&
          "${g72_first[1]:-}" == "$g72_acquire" ]] || fail "$name: the first two commands are not the metadata-bucket presence check and the lease create"
      elif grep -qF "gs://$g72_mb" "$out"; then fail "$name: an uncovered command prints a metadata-bucket or lease command"; fi
      if [[ "$sub" == deploy || "$sub" == provision ]]; then
        g72_f="$(grep -nxF -- "$g72_fence" "$out" | head -1 | cut -d: -f1)"
        g72_u="$(grep -nE '^\+ git -C [^ ]+ archive --format=tar ' "$out" | head -1 | cut -d: -f1)"
        [[ -n "$g72_f" && -n "$g72_u" ]] && (( g72_f < g72_u )) || fail "$name: no schema-fence entry before the release upload"
      elif grep -qF 'fence.tsv' "$out"; then fail "$name: a schema-fence entry outside deploy/provision"; fi
    done
    why+="${#cmds[@]} subcommands with --dry-run; env $ENV_NAME; outputs in lifecycle/; G072 lease lines first in the six covered mutators and absent elsewhere, fence entry before the upload in deploy and provision only; "
    # G072 additive interfaces (Plan §10.9), dry runs only under the same shims: build --source, GCLOUD_IMPERSONATE_SA,
    # the lease subcommand and STAGING_LEASE_HELD. Outputs lifecycle/g072-<name>.txt.
    gi="$S/g072-iface"; mkdir -p "$gi/src"
    g72_dry() { # NAME CONF ARGS...: g72_rc; STAGING_LEASE_HELD only from the caller's G72_HELD
      local name=$1 cf=$2; shift 2
      g72_rc=0
      (cd "$REPO" && env -u STAGING_LEASE_HELD PATH="$S/shim:$PATH" HOME="$S/home" ${G72_HELD:+"STAGING_LEASE_HELD=$G72_HELD"} \
        bash infra/staging/staging.sh "$@" --config "$cf" --dry-run) >"$REPORT/lifecycle/g072-$name.txt" 2>&1 || g72_rc=$?
      if grep -q . "$S/shim.log"; then fail "g072 $name: executed $(cut -d' ' -f1 "$S/shim.log" | sort -u | tr '\n' ' ')"; : >"$S/shim.log"; fi
    }
    g72_refused() { # NAME TEXT CONF ARGS...: exit 2 with exactly the line "staging: TEXT"
      local name=$1 text=$2; shift 2
      g72_dry "$name" "$@"
      [[ $g72_rc == 2 ]] && grep -qxF -- "staging: $text" "$REPORT/lifecycle/g072-$name.txt" || fail "g072 $name: exit $g72_rc, want 2 with 'staging: $text'"
    }
    g72_has() { grep -qF -- "$2" "$REPORT/lifecycle/g072-$1.txt"; } # NAME TEXT
    g72_cmds() { grep '^+ ' "$REPORT/lifecycle/g072-$1.txt" || true; } # NAME: the printed commands
    # 1. build --source DIR: Compose files, build contexts, edge context and git checks from DIR; the default unchanged.
    g72_dry build-source "$conf" build --candidate "$g72_cand" --source "$gi/src"
    [[ $g72_rc == 0 ]] && g72_has build-source "git -C $gi/src rev-parse HEAD" && g72_has build-source "-f $gi/src/infra/compose/local.yml" &&
      g72_has build-source "-f $gi/src/infra/compose/migrate.yml" && g72_has build-source "--file $gi/src/infra/docker/edge.Dockerfile" &&
      g72_has build-source "$gi/src/infra/docker" && ! g72_has build-source "$REPO/infra/compose/" ||
      fail "g072 build --source: exit $g72_rc; the build does not take its Compose files, contexts and git checks from DIR only"
    grep -qF -- "-f $REPO/infra/compose/local.yml" "$REPORT/lifecycle/build.txt" && grep -qF -- "git -C $REPO rev-parse HEAD" "$REPORT/lifecycle/build.txt" ||
      fail "g072 build without --source no longer builds from the checkout"
    g72_refused source-relative "--source takes an existing absolute directory without spaces" "$conf" build --candidate "$g72_cand" --source relative/dir
    g72_refused source-absent "--source takes an existing absolute directory without spaces" "$conf" build --candidate "$g72_cand" --source "$gi/absent"
    g72_refused source-push "push does not take --source" "$conf" push --candidate "$g72_cand" --source "$gi/src"
    g72_refused source-space "--source takes an existing absolute directory without spaces" "$conf" build --candidate "$g72_cand" --source "$gi/src dir"
    # 2. GCLOUD_IMPERSONATE_SA: absent or empty = the deployer (the default stop's exact commands); ci-stop replaces only
    # the identity; push keeps the deployer; near misses refused.
    # All in one directory (the default PRIVATE_DIR is the config's): the example as is, then with the key.
    sa_conf() { { cat "$REPO/$conf"; [[ $2 == - ]] || printf 'GCLOUD_IMPERSONATE_SA=%s\n' "$2"; } >"$gi/$1.conf"; } # NAME VALUE|-
    sa_text="staging.conf: GCLOUD_IMPERSONATE_SA must be empty or $ENV_NAME-deployer, $ENV_NAME-ci-stop or $ENV_NAME-ci-read @<GCP_PROJECT>.iam.gserviceaccount.com"
    g72_stop_sa="$ENV_NAME-ci-stop@$g72_proj.iam.gserviceaccount.com" g72_read_sa="$ENV_NAME-ci-read@$g72_proj.iam.gserviceaccount.com"
    sa_conf sa-default -; sa_conf sa-empty ""; sa_conf sa-stop "$g72_stop_sa"; sa_conf sa-read "$g72_read_sa"
    g72_dry sa-default "$gi/sa-default.conf" stop; g72_dry sa-empty "$gi/sa-empty.conf" stop; g72_dry sa-stop "$gi/sa-stop.conf" stop
    g72_dry sa-read-push "$gi/sa-read.conf" push --candidate "$g72_cand"
    sa_ids() { grep -oE -- '--impersonate-service-account [^ ]+' "$REPORT/lifecycle/g072-$1.txt" | sort -u | tr '\n' ' '; } # NAME
    [[ -n "$(g72_cmds sa-default)" && "$(g72_cmds sa-default)" == "$(g72_cmds sa-empty)" && "$(sa_ids sa-default)" == "--impersonate-service-account $g72_sa " ]] ||
      fail "g072 GCLOUD_IMPERSONATE_SA absent/empty: stop's commands differ, or impersonate another identity than DEPLOYER_SA"
    # Only gcloud's identity changes; Terraform (GOOGLE_IMPERSONATE_SERVICE_ACCOUNT) and push keep DEPLOYER_SA.
    [[ "$(sa_ids sa-stop)" == "--impersonate-service-account $g72_stop_sa " &&
      "$(g72_cmds sa-stop)" == "$(g72_cmds sa-default | sed "s|--impersonate-service-account $g72_sa|--impersonate-service-account $g72_stop_sa|g")" ]] ||
      fail "g072 GCLOUD_IMPERSONATE_SA=ci-stop: stop's commands are not the default's with only gcloud's identity replaced"
    g72_has sa-read-push "print-access-token --impersonate-service-account $g72_sa" || fail "g072 GCLOUD_IMPERSONATE_SA=ci-read: push no longer uses DEPLOYER_SA"
    for v in "$ENV_NAME-vm@$g72_proj.iam.gserviceaccount.com" "$ENV_NAME-ci-stop@other-project.iam.gserviceaccount.com" \
      "${ENV_NAME}0-ci-stop@$g72_proj.iam.gserviceaccount.com" "$g72_stop_sa.example.net"; do
      sa_conf sa-bad "$v"; g72_refused "sa-bad-${v%%@*}-${v#*@}" "$sa_text" "$gi/sa-bad.conf" stop
    done
    # 3. The lease subcommand (no lease taken itself) and its refusals.
    g72_dry lease-show "$conf" lease --show
    [[ $g72_rc == 0 ]] && g72_has lease-show "gcloud storage objects describe gs://$g72_mb/lease" && ! g72_has lease-show "<lease:" ||
      fail "g072 lease --show: exit $g72_rc; want the lease read only"
    g72_dry lease-release "$conf" lease --release
    [[ $g72_rc == 0 ]] && g72_has lease-release "+ (gcloud storage rm gs://$g72_mb/lease --if-generation-match=<observed>)" && ! g72_has lease-release "<lease:" ||
      fail "g072 lease --release: exit $g72_rc; want the conditional delete only"
    g72_dry lease-break "$conf" lease --break --confirm "$ENV_NAME" --reason "g072 check"
    [[ $g72_rc == 0 ]] && g72_has lease-break "gs://$g72_mb/lease-breaks/<utc>-<holder>.tsv --if-generation-match=0, then gcloud storage rm gs://$g72_mb/lease --if-generation-match=<observed>" ||
      fail "g072 lease --break: exit $g72_rc; want the create-only break record, then the conditional delete"
    g72_refused lease-two "lease takes exactly one of --show, --release and --break" "$conf" lease --show --release
    g72_refused lease-none "lease takes exactly one of --show, --release and --break" "$conf" lease
    g72_refused lease-noreason "lease --break needs --reason TEXT (3 to 200 characters, no control character)" "$conf" lease --break --confirm "$ENV_NAME"
    g72_refused lease-confirm "lease --break needs --confirm $ENV_NAME" "$conf" lease --break --confirm other-env --reason "g072 check"
    g72_refused lease-stray "--confirm and --reason belong to lease --break" "$conf" lease --show --confirm "$ENV_NAME"
    # 4. STAGING_LEASE_HELD: the named lease is checked, nothing acquired; malformed refused; uncovered commands ignore it.
    G72_HELD="ci:123/456@7" g72_dry held-stop "$conf" stop
    [[ $g72_rc == 0 ]] && g72_has held-stop "+ (gs://$g72_mb present: the live lease must be ci:123/456 at generation 7; nothing acquired or released)" &&
      ! g72_has held-stop "<lease:" || fail "g072 STAGING_LEASE_HELD: exit $g72_rc; want the held-lease check instead of an acquire"
    G72_HELD="bogus" g72_refused held-bad "STAGING_LEASE_HELD must be <holder>@<generation>" "$conf" stop
    G72_HELD="ci:123/456@7" g72_dry held-build "$conf" build --candidate "$g72_cand"
    [[ $g72_rc == 0 ]] && ! g72_has held-build "gs://$g72_mb" || fail "g072 STAGING_LEASE_HELD on build: exit $g72_rc; build takes no lease"
    if grep -q '^STAGING_SSH_KEY_EXPIRE_AFTER\|STAGING_SSH_KEY_EXPIRE_AFTER' "$REPO/infra/staging/lib.sh"; then g72_a6="present (not covered here)"; else g72_a6="absent (A6 not needed)"; fi
    why+="G072 interfaces: build --source from DIR (default unchanged; 4 refusals), GCLOUD_IMPERSONATE_SA absent/empty = today's stop commands, ci-stop replaces only the identity, push keeps the deployer, 4 near misses refused, lease --show/--release/--break and 5 refusals, STAGING_LEASE_HELD checked not acquired (malformed refused, build ignores it); STAGING_SSH_KEY_EXPIRE_AFTER $g72_a6; "
    # staging.conf's optional root settings (G071 S11(iv)) through the real parsers: operator load_conf + remote_env,
    # then the VM's load_meta + render_env on that remote.env and a synthetic manifest (no secret: all versions none).
    rs="$S/root-settings" cand="${COMMIT%+dirty}" zero="$(printf '0%.0s' {1..64})"
    ar="$(sed -n -E 's/^AR_REPOSITORY=(.*)$/\1/p' "$REPO/$conf")"
    rs_env="$(sed -n -E 's/^STAGING_ENV=([a-z][a-z0-9-]{1,20})$/\1/p' "$REPO/$conf" | head -1)"
    mkdir -p "$rs"
    root_case() { # NAME WANT_RC WANT_STDERR WANT_LINES CONF_LINE...: WANT_LINES = remote.env@app.env pairs, |-separated
      local name=$1 want_rc=$2 want_err=$3 want_env=$4 rc=0 d="$rs/$1" l svc rs_lines=()
      shift 4
      mkdir -p "$d/meta"
      { cat "$REPO/$conf"; printf '%s\n' "$@"; } >"$d/staging.conf"
      (cd "$REPO" && env PATH="$S/shim:$PATH" HOME="$S/home" bash -c 'source infra/staging/lib.sh; load_conf "$1"; remote_env' _ \
        "$d/staging.conf") >"$d/meta/remote.env" 2>"$d/stderr.txt" || rc=$?
      cp "$d/stderr.txt" "$REPORT/lifecycle/root-settings-$name.stderr.txt"
      [[ $rc == "$want_rc" ]] || { fail "root settings $name: load_conf exit $rc, want $want_rc"; return 0; }
      if [[ $want_rc != 0 ]]; then
        [[ "$(cat "$d/stderr.txt")" == "staging: $want_err" ]] || fail "root settings $name: refusal text differs (lifecycle/root-settings-$name.stderr.txt)"
        return 0
      fi
      printf 'candidate\t%s\n' "$cand" >"$d/meta/manifest.tsv"
      for svc in core web intelligence minio minio-init migrate edge; do
        printf 'image\t%s\t%s\t-\t-\t%s/%s@sha256:%s\n' "$svc" "$cand" "$ar" "$svc" "$zero" >>"$d/meta/manifest.tsv"
      done
      rc=0
      # shellcheck disable=SC2016 # expanded by the inner bash
      (cd "$REPO" && env PATH="$S/shim:$PATH" HOME="$S/home" bash -c 'set -u; source infra/staging/remote/lib.sh; ENV_NAME=$1 ENV_FILE=$2/app.env
        load_meta "$2"; for n in "${SECRETS_REQUIRED[@]}" "${SECRETS_OPTIONAL[@]}"; do SECRET_VERSION[$n]=none; done; render_env' _ \
        "$rs_env" "$d/meta") >"$d/vm.txt" 2>&1 || rc=$?
      [[ $rc == 0 && -f "$d/meta/app.env" ]] || { fail "root settings $name: load_meta/render_env exit $rc ($(head -c 200 "$d/vm.txt"))"; return 0; }
      IFS='|' read -ra rs_lines <<<"$want_env"
      for l in "${rs_lines[@]}"; do
        grep -qxF -- "${l%%@*}" "$d/meta/remote.env" || fail "root settings $name: remote.env lacks the line ${l%%@*}"
        grep -qxF -- "${l#*@}" "$d/meta/app.env" || fail "root settings $name: the rendered env lacks the line ${l#*@}"
      done
      cp "$d/meta/remote.env" "$REPORT/lifecycle/root-settings-$name.remote.env"
      grep -E '^(DEFAULT_TENANT_SLUG|PLATFORM_ROOT_PAGE|MCP_CONNECT_ISSUER|MCP_ALLOWED_ORIGINS)=' "$d/meta/app.env" >"$REPORT/lifecycle/root-settings-$name.app-env.txt" || true
    }
    root_case defaults 0 "" "DEFAULT_TENANT_SLUG=@DEFAULT_TENANT_SLUG=|PLATFORM_ROOT_PAGE=@PLATFORM_ROOT_PAGE=|MCP_CONNECT_ISSUER=@MCP_CONNECT_ISSUER=|MCP_ALLOWED_ORIGINS=@MCP_ALLOWED_ORIGINS="
    root_case accepted 0 "" "DEFAULT_TENANT_SLUG=alpha-agency@DEFAULT_TENANT_SLUG='alpha-agency'|PLATFORM_ROOT_PAGE=landing@PLATFORM_ROOT_PAGE='landing'" \
      DEFAULT_TENANT_SLUG=alpha-agency PLATFORM_ROOT_PAGE=landing
    root_case bad-slug 2 'staging.conf: DEFAULT_TENANT_SLUG must be empty or a slug (^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$)' "" \
      DEFAULT_TENANT_SLUG=Alpha_Agency
    root_case bad-page 2 'staging.conf: PLATFORM_ROOT_PAGE must be empty, signup_redirect or landing' "" PLATFORM_ROOT_PAGE=redirect
    # MCP settings (MCP.md): empty keeps MCP off; an exact https issuer (and origins with it) reaches remote.env and the rendered env.
    root_case mcp-accepted 0 "" "MCP_CONNECT_ISSUER=https://example-env.authkit.app@MCP_CONNECT_ISSUER='https://example-env.authkit.app'|MCP_ALLOWED_ORIGINS=https://client.example,https://other.example:8443@MCP_ALLOWED_ORIGINS='https://client.example,https://other.example:8443'" \
      MCP_CONNECT_ISSUER=https://example-env.authkit.app MCP_ALLOWED_ORIGINS=https://client.example,https://other.example:8443
    root_case mcp-issuer-only 0 "" "MCP_CONNECT_ISSUER=https://example-env.authkit.app@MCP_CONNECT_ISSUER='https://example-env.authkit.app'|MCP_ALLOWED_ORIGINS=@MCP_ALLOWED_ORIGINS=" \
      MCP_CONNECT_ISSUER=https://example-env.authkit.app
    root_case mcp-issuer-http 2 'staging.conf: MCP_CONNECT_ISSUER must be empty or an exact https issuer (https://<host>, no path or trailing slash)' "" \
      MCP_CONNECT_ISSUER=http://example-env.authkit.app
    root_case mcp-issuer-slash 2 'staging.conf: MCP_CONNECT_ISSUER must be empty or an exact https issuer (https://<host>, no path or trailing slash)' "" \
      MCP_CONNECT_ISSUER=https://example-env.authkit.app/
    root_case mcp-origin-http 2 'staging.conf: MCP_ALLOWED_ORIGINS must be empty or comma-separated exact https origins without spaces' "" \
      MCP_CONNECT_ISSUER=https://example-env.authkit.app MCP_ALLOWED_ORIGINS=http://client.example
    root_case mcp-origin-space 2 'staging.conf: MCP_ALLOWED_ORIGINS must be empty or comma-separated exact https origins without spaces' "" \
      MCP_CONNECT_ISSUER=https://example-env.authkit.app "MCP_ALLOWED_ORIGINS=https://a.example, https://b.example"
    root_case mcp-origins-alone 2 'staging.conf: MCP_ALLOWED_ORIGINS needs MCP_CONNECT_ISSUER' "" MCP_ALLOWED_ORIGINS=https://client.example
    if grep -q . "$S/shim.log"; then fail "root settings: executed $(cut -d' ' -f1 "$S/shim.log" | sort -u | tr '\n' ' ')"; fi
    why+="root settings: defaults and accepted values (root host and MCP settings) reach remote.env and the rendered env (KEY= when empty), a bad slug, a bad page and six malformed MCP settings (http or slashed issuer, http or spaced origins, origins without an issuer) are refused with their text; "
    # Two environments in one project, staging and staging-eu: for staging, the real reconcile, verify_absent,
    # env_secrets and label_check match only staging's exact names and label. staging.sh sourced; gcloud replaced by a
    # fake that applies each --filter it receives (labels.environment=X exactly, name ~ RE as a regular expression on
    # the name, or on projects/<n>/secrets/<id>) to an object table; state_names, cf_records and plan_changes replaced
    # by fixtures (no Terraform, jq or network).
    me="$S/multi-env"; mkdir -p "$me"
    printf '%s\t%s\t%s\n' instances staging-vm staging instances staging-eu-vm staging-eu disks staging-vm staging \
      disks staging-data staging disks staging-eu-vm staging-eu disks staging-eu-data staging-eu addresses staging-ip staging \
      addresses staging-eu-ip staging-eu firewall-rules staging-edge - firewall-rules staging-iap - firewall-rules staging-eu-edge - \
      firewall-rules staging-eu-iap - networks staging-vpc - networks staging-eu-vpc - subnets staging-subnet - \
      subnets staging-eu-subnet - secrets staging-CF_API_TOKEN - secrets staging-REDIS_PASSWORD - secrets staging-eu-FOO - \
      secrets staging-eu-CF_API_TOKEN - >"$me/both.tsv"
    grep -vE $'\tstaging-(vm|data|ip|edge|iap|vpc|subnet|CF_API_TOKEN|REDIS_PASSWORD)\t' "$me/both.tsv" >"$me/retired.tsv" || true
    { cat "$me/both.tsv"; printf 'disks\tstaging-extra\tstaging\n'; } >"$me/both-stray.tsv"
    { cat "$me/retired.tsv"; printf 'disks\tstaging-extra\tstaging\n'; } >"$me/retired-stray.tsv"
    printf '%s\n' staging-vm staging-data staging-ip staging-edge staging-iap staging-vpc staging-subnet staging.example.net >"$me/state-live.txt"
    : >"$me/state-empty.txt"
    # address, type, environment label, name, instance name ("-" = none), as plan_changes prints them
    printf '%s\t%s\t%s\t%s\t%s\n' module.vm.google_compute_network.vpc google_compute_network - staging-vpc - \
      module.vm.google_compute_subnetwork.subnet google_compute_subnetwork - staging-subnet - \
      module.vm.google_compute_firewall.edge google_compute_firewall - staging-edge - \
      module.vm.google_compute_firewall.iap google_compute_firewall - staging-iap - \
      module.vm.google_compute_address.ip google_compute_address staging staging-ip - \
      module.vm.google_compute_disk.data google_compute_disk staging staging-data - \
      module.vm.google_compute_instance.vm google_compute_instance staging staging-vm - \
      'module.vm.google_compute_instance_iam_member.os_login["user:op@example.net"]' google_compute_instance_iam_member - - staging-vm \
      module.dns.cloudflare_dns_record.base cloudflare_dns_record - staging.example.net - \
      module.dns.cloudflare_dns_record.wildcard cloudflare_dns_record - '*.staging.example.net' - >"$me/plan-own.tsv"
    printf '%s\t%s\t%s\t%s\t%s\n' module.vm.google_compute_subnetwork.subnet google_compute_subnetwork - staging-subnet - \
      module.eu.google_compute_network.vpc google_compute_network - staging-eu-vpc - \
      module.eu.google_compute_firewall.edge google_compute_firewall - staging-eu-edge - \
      module.eu.google_compute_instance.vm google_compute_instance staging-eu staging-eu-vm - \
      module.eu.google_compute_instance_iam_member.os_login google_compute_instance_iam_member - - staging-eu-vm >"$me/plan-other.tsv"
    printf '%s\t%s\t%s\t%s\t%s\n' module.vm.google_compute_firewall.edge google_compute_firewall - staging-edge - \
      module.x.google_compute_firewall.extra google_compute_firewall - staging-extra - >"$me/plan-stray.tsv"
    # Terraform names an instance IAM binding's instance by full path: own project and zone, then a foreign project, zone and instance.
    ipath=projects/g070-check-project/zones/us-east1-b/instances
    printf '%s\t%s\t%s\t%s\t%s\n' 'module.vm.google_compute_instance_iam_member.os_login["user:op@example.net"]' \
      google_compute_instance_iam_member - - "$ipath/staging-vm" >"$me/plan-path-own.tsv"
    printf '%s\t%s\t%s\t%s\t%s\n' module.a.google_compute_instance_iam_member.os_login google_compute_instance_iam_member - - \
      projects/other-project/zones/us-east1-b/instances/staging-vm \
      module.b.google_compute_instance_iam_member.os_login google_compute_instance_iam_member - - \
      projects/g070-check-project/zones/other-zone/instances/staging-vm \
      module.c.google_compute_instance_iam_member.os_login google_compute_instance_iam_member - - "$ipath/staging-eu-vm" >"$me/plan-path-foreign.tsv"
    # Two environments whose names extend one another, preview-1 and preview-10: preview-1's exact names and label only.
    printf '%s\t%s\t%s\n' instances preview-1-vm preview-1 instances preview-10-vm preview-10 disks preview-1-vm preview-1 \
      disks preview-1-data preview-1 disks preview-10-vm preview-10 disks preview-10-data preview-10 addresses preview-1-ip preview-1 \
      addresses preview-10-ip preview-10 firewall-rules preview-1-edge - firewall-rules preview-1-iap - firewall-rules preview-10-edge - \
      firewall-rules preview-10-iap - networks preview-1-vpc - networks preview-10-vpc - subnets preview-1-subnet - \
      subnets preview-10-subnet - secrets preview-1-CF_API_TOKEN - secrets preview-10-CF_API_TOKEN - >"$me/preview.tsv"
    grep -vE $'\tpreview-1-(vm|data|ip|edge|iap|vpc|subnet|CF_API_TOKEN)\t' "$me/preview.tsv" >"$me/preview-retired.tsv" || true
    printf '%s\n' preview-1-vm preview-1-data preview-1-ip preview-1-edge preview-1-iap preview-1-vpc preview-1-subnet >"$me/state-preview.txt"
    # The fake gcloud reads --filter in the filter language: name ~ takes one single-quoted value (an unquoted literal
    # ends at a parenthesis there), anything else fails the call with "fake gcloud: unquoted filter value".
    # shellcheck disable=SC2016 # expanded by the inner bash
    me_script='source infra/staging/staging.sh
      ENV_NAME=${MENV:-staging} DRY_RUN=0 COMMAND=check
      CONF[GCP_PROJECT]=g070-check-project CONF[GCP_ZONE]=us-east1-b CONF[BASE_DOMAIN]=staging.example.net
      GFLAGS=(--project g070-check-project)
      gcloud() {
        local a prev="" filter="" kind k n l f re sq=$SQ
        printf "%s\n" "$*" >>"$MED/gcloud.log"
        for a in "$@"; do [[ $prev != --filter ]] || filter=$a; prev=$a; done
        case "$1 $2 $3" in
          "compute networks subnets") kind=subnets ;;
          "secrets list "*) kind=secrets ;;
          compute\ *) kind=$2 ;;
          *) return 3 ;;
        esac
        if [[ $filter == "name ~ "* ]]; then
          re=${filter#"name ~ "}
          [[ ${#sq} == 1 && ${#re} -ge 2 && $re == "$sq"*"$sq" && ${re:1:${#re}-2} != *"$sq"* ]] ||
            { printf "fake gcloud: unquoted filter value: %s\n" "$filter" >&2; return 4; }
          re=${re:1:${#re}-2}
        fi
        while read -r k n l; do # tab-separated, no field empty or with a space
          [[ $k == "$kind" ]] || continue
          case $filter in
            labels.environment=*) [[ $l == "${filter#labels.environment=}" ]] || continue ;;
            "name ~ "*)
              f=$n
              [[ $kind != secrets ]] || f=projects/123456789/secrets/$n
              [[ $f =~ $re ]] || continue ;;
            *) return 3 ;;
          esac
          printf "%s\n" "$n"
        done <"$MED/$TABLE"
      }
      state_names() { cat "$MED/$STATE"; }
      cf_records() { echo 0; }
      plan_changes() { cat "$MED/$PLAN"; }
      case $1 in
        reconcile) reconcile ;;
        absent) verify_absent ;;
        secrets) env_secrets ;;
        plan) label_check /private/plans/fixture.tfplan; printf "plan\taccepted\n" ;;
      esac'
    me_case() { # NAME MODE TABLE STATE PLAN [ENV]: one run; sets me_rc; outputs in <scratch>/multi-env/NAME.{out,err}
      me_rc=0
      rm -f "$me/gcloud.log"
      (cd "$REPO" && env LC_ALL=C PATH="$S/shim:$PATH" HOME="$S/home" MED="$me" TABLE="$3" STATE="$4" PLAN="$5" MENV="${6:-staging}" \
        SQ="'" bash -c "$me_script" _ "$2") >"$me/$1.out" 2>"$me/$1.err" || me_rc=$?
      cp "$me/$1.out" "$REPORT/lifecycle/multi-env-$1.out"; cp "$me/$1.err" "$REPORT/lifecycle/multi-env-$1.err"
      [[ ! -f "$me/gcloud.log" ]] || cp "$me/gcloud.log" "$REPORT/lifecycle/multi-env-$1.gcloud.log"
    }
    me_case reconcile-two-envs reconcile both.tsv state-live.txt -
    [[ $me_rc == 0 ]] || fail "two environments: reconcile exit $me_rc, want 0 (staging-eu objects taken for orphans of staging, or an unquoted filter value?)"
    live="$(awk -F'\t' '$1 == "reconcile" && $2 == "live" {print $3}' "$me/reconcile-two-envs.out")"
    [[ "$live" == "staging-data staging-edge staging-iap staging-ip staging-subnet staging-vm staging-vpc " ]] ||
      fail "two environments: reconcile lists [$live], want exactly staging's seven objects"
    # The grouped name expressions reach gcloud quoted in the filter language (name ~ '^(a|b)$'), not as a raw literal.
    for l in "compute firewall-rules list --project g070-check-project --filter name ~ '^(staging-edge|staging-iap)\$' --format value(name)" \
      "compute networks list --project g070-check-project --filter name ~ '^(staging-vpc)\$' --format value(name)" \
      "compute networks subnets list --project g070-check-project --filter name ~ '^(staging-subnet)\$' --format value(name)"; do
      grep -qxF -- "$l" "$me/gcloud.log" || fail "two environments: gcloud did not receive the quoted filter [$l]"
    done
    # The reconcile dry run prints the same quoted argument (status runs reconcile).
    grep -qF -- "--filter $(printf '%q' "name ~ '^(staging-edge|staging-iap)\$'")" "$REPORT/lifecycle/status.txt" ||
      fail "the reconcile dry run does not print the quoted name filter"
    me_case reconcile-preview reconcile preview.tsv state-preview.txt - preview-1
    live="$(awk -F'\t' '$1 == "reconcile" && $2 == "live" {print $3}' "$me/reconcile-preview.out")"
    [[ $me_rc == 0 && "$live" == "preview-1-data preview-1-edge preview-1-iap preview-1-ip preview-1-subnet preview-1-vm preview-1-vpc " ]] ||
      fail "preview-1 and preview-10: reconcile exit $me_rc lists [$live], want exactly preview-1's seven objects"
    me_case absent-preview absent preview-retired.tsv state-empty.txt - preview-1
    [[ $me_rc == 0 ]] && grep -q $'^absent\t' "$me/absent-preview.out" ||
      fail "preview-1 and preview-10: verify_absent exit $me_rc after preview-1's retirement, want 0 (preview-10 objects counted?)"
    me_case secrets-preview secrets preview.tsv - - preview-1
    [[ $me_rc == 0 && "$(tr '\n' ' ' <"$me/secrets-preview.out")" == "preview-1-CF_API_TOKEN " ]] ||
      fail "preview-1 and preview-10: env_secrets lists [$(tr '\n' ' ' <"$me/secrets-preview.out")] (exit $me_rc), want preview-1-CF_API_TOKEN"
    me_case reconcile-stray reconcile both-stray.tsv state-live.txt -
    [[ $me_rc == 2 ]] && grep -qF 'objects of staging outside the Terraform state: staging-extra (' "$me/reconcile-stray.err" ||
      fail "two environments: a stray labelled staging-extra is not reported as an orphan (exit $me_rc)"
    me_case absent-two-envs absent retired.tsv state-empty.txt -
    [[ $me_rc == 0 ]] && grep -q $'^absent\t' "$me/absent-two-envs.out" ||
      fail "two environments: verify_absent exit $me_rc after staging's retirement, want 0 (staging-eu objects counted?)"
    me_case absent-stray absent retired-stray.tsv state-empty.txt -
    [[ $me_rc == 4 ]] && grep -qF 'retirement left objects: live: staging-extra;' "$me/absent-stray.err" ||
      fail "two environments: verify_absent does not report a stray labelled staging-extra (exit $me_rc)"
    me_case secrets-two-envs secrets both.tsv - -
    [[ $me_rc == 0 && "$(LC_ALL=C sort "$me/secrets-two-envs.out" | tr '\n' ' ')" == "staging-CF_API_TOKEN staging-REDIS_PASSWORD " ]] ||
      fail "two environments: env_secrets lists [$(tr '\n' ' ' <"$me/secrets-two-envs.out")] (exit $me_rc), want staging-CF_API_TOKEN staging-REDIS_PASSWORD"
    grep -qxF -- "secrets list --project g070-check-project --filter name ~ '/secrets/staging-[A-Z][A-Z0-9_]*\$' --format value(name.basename())" "$me/gcloud.log" ||
      fail "two environments: the Secret Manager filter is not the quoted, anchored upper-case ID expression"
    me_case plan-own plan both.tsv - plan-own.tsv
    [[ $me_rc == 0 ]] && grep -q $'^plan\taccepted$' "$me/plan-own.out" || fail "two environments: label_check refuses staging's own plan (exit $me_rc)"
    me_case plan-other plan both.tsv - plan-other.tsv
    if [[ $me_rc != 2 ]]; then fail "two environments: label_check accepts staging-eu's resources for staging (exit $me_rc)"
    else
      for a in module.eu.google_compute_network.vpc module.eu.google_compute_firewall.edge module.eu.google_compute_instance.vm \
        module.eu.google_compute_instance_iam_member.os_login; do
        grep -qF -- " $a" "$me/plan-other.err" || fail "two environments: label_check does not name $a"
      done
      ! grep -qF 'module.vm.google_compute_subnetwork.subnet' "$me/plan-other.err" || fail "two environments: label_check refuses staging's own subnet"
    fi
    me_case plan-stray plan both.tsv - plan-stray.tsv
    [[ $me_rc == 2 ]] && grep -qF 'outside staging: module.x.google_compute_firewall.extra' "$me/plan-stray.err" ||
      fail "two environments: label_check accepts an unlabelled staging-extra (exit $me_rc)"
    me_case plan-path-own plan both.tsv - plan-path-own.tsv
    [[ $me_rc == 0 ]] && grep -q $'^plan\taccepted$' "$me/plan-path-own.out" ||
      fail "two environments: label_check refuses the IAM binding of staging's own instance named by full path (exit $me_rc)"
    me_case plan-path-foreign plan both.tsv - plan-path-foreign.tsv
    if [[ $me_rc != 2 ]]; then fail "two environments: label_check accepts a foreign project, zone or instance path (exit $me_rc)"
    else
      for a in module.a.google_compute_instance_iam_member.os_login module.b.google_compute_instance_iam_member.os_login \
        module.c.google_compute_instance_iam_member.os_login; do
        grep -qF -- " $a" "$me/plan-path-foreign.err" || fail "two environments: label_check does not refuse $a"
      done
    fi
    # the prefix forms must not come back (reconcile/live_names, verify_absent and the receipt's secret listing)
    pre="$(grep -nE 'name ~ \^\$ENV_NAME-|/secrets/\$ENV_NAME-|startswith\(\$env' "$REPO/infra/staging/lib.sh" "$REPO/infra/staging/staging.sh" || true)"
    [[ -z "$pre" ]] || fail "an <env>- prefix match remains (${pre//$'\n'/ | })"
    grep -qF 'secrets="$(env_secrets 2>/dev/null)"' "$REPO/infra/staging/staging.sh" || fail "write_receipt does not list secrets through env_secrets"
    if grep -q . "$S/shim.log"; then fail "two environments: executed $(cut -d' ' -f1 "$S/shim.log" | sort -u | tr '\n' ' ')"; fi
    why+="two environments (staging, staging-eu): reconcile, verify_absent, env_secrets and label_check take only staging's exact names and label; a stray labelled staging-extra is still an orphan and blocks the absence check, an unlabelled one is refused in a plan; an IAM binding's full instance path counts as its instance only in this environment's project and zone; name filters reach gcloud quoted (name ~ '^(a|b)\$'), an unquoted value fails the fake; preview-1 beside preview-10 takes only its own names; "

    # ---- Session length (Plan r1.9; Codex runtime review clarifications 1-4) ----------------------------------------
    # (a) The real CLI with --dry-run: parser forms, refusals before any command, the dry-run command sets.
    lc="$REPORT/lifecycle" sl="$S/session"; mkdir -p "$sl"
    line_of() { { grep -nE -- "$1" "$2" || true; } | awk -F: -v a="${3:-0}" '$1 > a {print $1; exit}'; } # ERE FILE [AFTER_LINE]
    tf_plan_re='terraform[^|;&]*[[:space:]]plan[[:space:]]'
    # set-scheduling sets the length (every other scheduling attribute stated as provisioned, never --clear-* or
    # --termination-time), then the refresh-only plan carries the same seconds
    sched_re='gcloud compute instances set-scheduling staging-vm --zone [^|;&]* --max-run-duration=SECONDSs --instance-termination-action=STOP --maintenance-policy=MIGRATE --restart-on-failure --provisioning-model=STANDARD$'
    for d in 90m:5400 2h:7200 7200:7200; do
      f="$lc/start-duration-${d%%:*}.txt"
      grep -qE "${sched_re/SECONDS/${d#*:}}" "$f" || fail "start --duration ${d%%:*}: no set-scheduling with --max-run-duration=${d#*:}s and the provisioned attributes"
      grep -qE "${tf_plan_re}-refresh-only[[:space:]][^|;&]*[[:space:]]session_max_seconds=${d#*:}[[:space:]]" "$f" ||
        fail "start --duration ${d%%:*}: no saved refresh-only plan with -var session_max_seconds=${d#*:}"
      ! grep -qE -- '--clear-|--termination-time' "$f" || fail "start --duration ${d%%:*}: a --clear-* or --termination-time flag"
    done
    f="$lc/start-duration-90m.txt" n=0
    for p in 'output -raw max_run_duration_seconds' 'instances describe [^|;&]*--format value\\\(status\\\)$' "${sched_re/SECONDS/5400}" \
      'instances describe [^|;&]*instanceTerminationAction' "${tf_plan_re}-refresh-only[[:space:]][^|;&]*session_max_seconds=5400 " \
      '\(refuse unless every changed resource belongs' \
      '\(HOLD unless the refresh-only plan proposes no change and only module\.vm\.google_compute_instance\.vm drifts: scheduling\[0\]\.max_run_duration\[0\]\.seconds' \
      'terraform[^|;&]*[[:space:]]apply[[:space:]]' 'instances describe [^|;&]*maxRunDuration' 'output -raw max_run_duration_seconds' \
      'gcloud compute instances start '; do
      n="$(line_of "$p" "$f" "$n")"
      [[ -n "$n" ]] || { fail "start --duration 90m: dry run lacks, or misorders, [$p] (session read, fresh status, set-scheduling, readback, refresh-only plan, label check, scope check, apply, re-read, start)"; break; }
    done
    grep -qF 'output -raw max_run_duration_seconds' "$lc/start.txt" || fail "start: the dry run does not read Terraform's max_run_duration_seconds"
    ! grep -qE "${tf_plan_re}|terraform[^|;&]*[[:space:]]apply[[:space:]]" "$lc/start.txt" || fail "start without --duration: the dry run plans or applies"
    grep -qE "${tf_plan_re}[^|;&]*session_max_seconds=\\\\<observed\\\\>" "$lc/provision.txt" ||
      fail "provision on an existing VM: the plan does not pass the observed session length"
    grep -qE "${tf_plan_re}[^|;&]*session_max_seconds=7200[[:space:]]" "$lc/provision-duration-2h.txt" ||
      fail "provision --duration 2h (first allocation): the plan does not pass session_max_seconds=7200"
    ! grep -qE 'terraform|instances[[:space:]]+(start|stop)[[:space:]]|set-scheduling' "$lc/deploy.txt" ||
      fail "deploy: the dry run runs Terraform, a VM start/stop or set-scheduling"
    refuse_case() { # NAME WANT_STDERR COMMAND [ARG...]: exit 2 before any command, stderr exactly "staging: WANT_STDERR"
      local f="$sl/refuse-$1.txt" want="staging: $2" rc=0 cmd=$3
      shift 3
      (cd "$REPO" && env PATH="$S/shim:$PATH" HOME="$S/home" CLOUDSDK_CONFIG="$S/home/gcloud" \
        bash infra/staging/staging.sh "$cmd" --config "$conf" --dry-run "$@") >"$f" 2>&1 || rc=$?
      cp "$f" "$lc/session-refuse-$1.txt"
      [[ $rc == 2 && "$(cat "$f")" == "$want" ]] || fail "session refusal $1: exit $rc, output differs from [$want] (lifecycle/session-refuse-$1.txt)"
      if grep -q . "$S/shim.log"; then fail "session refusal $1: executed $(cut -d' ' -f1 "$S/shim.log" | sort -u | tr '\n' ' ')"; : >"$S/shim.log"; fi
    }
    range="whole seconds from 600 (10m) to 28800 (8h)"
    for d in 599 28801 9h 1.5h 0 090m 2d abc 10M -5m ''; do
      refuse_case "duration-${d:-empty}" "--duration takes <n>m, <n>h or $range, not $d" start --duration "$d"
    done
    refuse_case duration-missing "--duration needs a value" start --duration
    refuse_case duration-twice "--duration is given twice" start --duration 2h --duration 3h
    refuse_case deploy "deploy does not take --duration" deploy --duration 2h --candidate "${COMMIT%+dirty}"
    for cmd in stop status receipt retire reset build push; do refuse_case "$cmd" "$cmd does not take --duration" "$cmd" --duration 2h; done
    refuse_case override "--override session_max_seconds is not accepted: the session length is --duration (provision on first allocation, otherwise start --duration)" \
      provision --candidate "${COMMIT%+dirty}" --override session_max_seconds=600
    { cat "$REPO/$conf"; echo SESSION_MAX_SECONDS=3600; } >"$sl/conf-3600"
    rc=0; (cd "$REPO" && env PATH="$S/shim:$PATH" HOME="$S/home" bash infra/staging/staging.sh start --config "$sl/conf-3600" --dry-run) >"$sl/conf-3600.txt" 2>&1 || rc=$?
    [[ $rc == 0 ]] && grep -qF 'already runs its 3600s session' "$sl/conf-3600.txt" || fail "SESSION_MAX_SECONDS=3600: start does not select it (exit $rc)"
    { cat "$REPO/$conf"; echo SESSION_MAX_SECONDS=28801; } >"$sl/conf-bad"
    rc=0; (cd "$REPO" && env PATH="$S/shim:$PATH" HOME="$S/home" bash infra/staging/staging.sh start --config "$sl/conf-bad" --dry-run) >"$sl/conf-bad.txt" 2>&1 || rc=$?
    [[ $rc == 2 && "$(cat "$sl/conf-bad.txt")" == "staging: staging.conf: SESSION_MAX_SECONDS must be whole seconds from 600 to 28800" ]] ||
      fail "SESSION_MAX_SECONDS=28801 is not refused with its text (exit $rc)"
    if grep -q . "$S/shim.log"; then fail "session dry runs: executed $(cut -d' ' -f1 "$S/shim.log" | sort -u | tr '\n' ' ')"; : >"$S/shim.log"; fi

    # (b) staging.sh sourced, DRY_RUN=0, with the instance, Terraform and the VM side replaced by fakes over a state
    # directory (state, have = the instance's seconds, tfout = Terraform's output, stop = terminationTimestamp, sched =
    # the other scheduling columns); every call is appended to calls.log. plan_changes (label_check's reader) and
    # refresh_scope_rows (refresh_scope_check's jq reader) return fixtures: no Terraform, jq, gcloud or network.
    # shellcheck disable=SC2016 # expanded by the inner bash
    ses_script='source infra/staging/staging.sh
      load_conf infra/staging/examples/staging.conf.example
      DRY_RUN=0 COMMAND=$MODE
      OPT[candidate]=0000000000000000000000000000000000000000
      [[ -z ${DUR:-} ]] || OPT[duration]=$DUR
      [[ -z ${YES:-} ]] || OPT[yes]=1
      ev() { printf "%s\n" "$*" >>"$SD/calls.log"; }
      private_dirs() { :; }
      reconcile() { ev reconcile; }
      record_state() { ev record_state; }
      inventory() { ev "inventory $*"; }
      check_address() { ev check_address; }
      wait_vm_ready() { ev wait_vm_ready; }
      remote() { ev "remote $*"; }
      deploy_candidate() { ev "deploy $*"; }
      sleep() { :; }
      # DESC=fail fails every instance describe, DESC=empty makes the status describe succeed with no output,
      # DESC=session fails only the session describe; LIST=fail fails the exact-name list. The list answers only the form
      # vm_read sends (--zones, --filter name=staging-vm): empty when the instance is ABSENT, otherwise the name; any
      # other list form (a regex, a prefix, no zone) fails, so it can never confirm an absence. A raw file, when
      # present, is the session describe output verbatim (the parse cases: lines without tabs, an empty output).
      # SETSCHED=fail fails set-scheduling with no change, =nochange succeeds without changing the length, =other
      # sets the length but turns automatic restart off; DESC=sched fails only the full scheduling readback.
      gcloud() {
        local a
        ev "gcloud $*"
        case "$*" in
          "compute instances list --zones us-central1-a "*" --filter name=staging-vm --format value(name)")
            [[ ${LIST:-ok} != fail ]] || return 1
            [[ $(<"$SD/state") == ABSENT ]] || echo staging-vm
            return 0 ;;
          "compute instances describe "*) [[ ${DESC:-ok} != fail ]] || return 1 ;;
        esac
        [[ $(<"$SD/state") != ABSENT || $1 != compute ]] || return 1
        case "$*" in
          "compute instances set-scheduling "*)
            [[ $(<"$SD/state") == TERMINATED ]] || { ev "gcloud-fake: set-scheduling on a $(<"$SD/state") VM"; return 1; }
            case ${SETSCHED:-ok} in
              fail) return 1 ;;
              nochange) return 0 ;;
              other) printf "STOP\tSTANDARD\tFalse\tMIGRATE\tFalse\n" >"$SD/sched" ;;
            esac
            for a in "$@"; do [[ $a != --max-run-duration=* ]] || { a=${a#--max-run-duration=}; echo "${a%s}" >"$SD/have"; }; done ;;
          "compute instances describe "*instanceTerminationAction*)
            [[ ${DESC:-ok} != sched ]] || return 1
            printf "%s\t%s\t%s\n" "$(<"$SD/state")" "$(<"$SD/have")" "$(cat "$SD/sched" 2>/dev/null || printf "STOP\tSTANDARD\tTrue\tMIGRATE\tFalse")" ;;
          "compute instances describe "*maxRunDuration*)
            [[ ${DESC:-ok} != session ]] || return 1
            if [[ -f $SD/raw ]]; then cat "$SD/raw"; else printf "%s\t%s\t%s\n" "$(<"$SD/state")" "$(<"$SD/have")" "$(<"$SD/stop")"; fi ;;
          "compute instances describe "*"value(status)"*) [[ ${DESC:-ok} == empty ]] || cat "$SD/state" ;;
          "compute instances describe "*) ;;
          "compute instances start "*) echo RUNNING >"$SD/state"; echo 2026-10-27T14:00:00.000-07:00 >"$SD/stop" ;;
          *) return 3 ;;
        esac
      }
      # Each saved plan gets a kind by its -out file: refresh (-refresh-only), destroy, create (no VM), same (the
      # length Terraform records) or replace: a normal plan with another length on an existing VM, which hashicorp/google
      # 8.8.1 plans as a replacement; applying one logs "terraform-fake: REPLACEMENT" and fails, and the row fails on it.
      # The refresh apply records the instance length (APPLY=fail fails it, =noop leaves the Terraform output old,
      # =unreadable makes the output fail).
      tf() {
        local a w="" out="" kind=same
        ev "terraform $*"
        case $1 in
          init) ;;
          show) [[ -z ${REALJQ:-} || $2 != -json ]] || cat "$SD/plan.json" ;; # the hand-written terraform show -json document
          output) [[ $(<"$SD/tfout") != none ]] || return 1; printf "%s" "$(<"$SD/tfout")" ;;
          plan)
            for a in "$@"; do
              case $a in
                session_max_seconds=*) w=${a#session_max_seconds=} ;;
                -out=*) out=${a#-out=} ;;
                -refresh-only) kind=refresh ;;
                -destroy) kind=destroy ;;
              esac
            done
            if [[ $kind == same ]]; then
              if [[ $(<"$SD/state") == ABSENT ]]; then kind=create; elif [[ $w != "$(<"$SD/tfout")" ]]; then kind=replace; fi
            fi
            [[ -z $out ]] || printf "%s\n" "$kind" >"$SD/kind-${out##*/}"
            printf "%s\n" "$w" >"$SD/planned"
            return "${PLANRC:-2}" ;;
          apply)
            a=${!#}
            kind="$(cat "$SD/kind-${a##*/}" 2>/dev/null || true)"
            case $kind in
              replace) ev "terraform-fake: REPLACEMENT applied ($*)"; return 1 ;;
              refresh)
                case ${APPLY:-ok} in
                  ok) cp "$SD/have" "$SD/tfout" ;;
                  fail) return 1 ;;
                  noop) ;;
                  unreadable) echo none >"$SD/tfout" ;;
                esac ;;
              create) cp "$SD/planned" "$SD/have"; cp "$SD/planned" "$SD/tfout"; echo RUNNING >"$SD/state" ;;
              same | destroy) ;;
              *) ev "terraform-fake: apply of a plan no plan call saved ($*)"; return 1 ;;
            esac ;;
          *) return 3 ;;
        esac
      }
      plan_changes() { ev label_check; printf "%s\t%s\t%s\t%s\t%s\n" "$SESSION_ADDRESS" google_compute_instance staging staging-vm -; }
      # LABELFAIL=1 makes the label check refuse the plan the way the real one does (die 2)
      if [[ -n ${LABELFAIL:-} ]]; then label_check() { ev label_check; die 2 "label check refused (fake)"; }; fi
      # REALJQ=1 keeps the real refresh_scope_rows (its jq program, run by the pinned jq image through the jq shim)
      if [[ -z ${REALJQ:-} ]]; then refresh_scope_rows() { ev scope_check; cat "$SD/scope.tsv"; }; fi
      case $MODE in
        start) cmd_start ;;
        provision) cmd_provision ;;
        stop) cmd_stop ;;
        status) cmd_status ;;
        retire) cmd_retire ;;
        scope) refresh_scope_check /private/plans/fixture-refresh.tfplan "$SHAVE" "$SWANT"; printf "scope\taccepted\n" ;;
        parse) vm_session; printf "%s|%s|%s\n" "$SESSION_STATE" "$SESSION_HAVE" "$SESSION_STOP" ;;
        replace-control) # a normal plan with another length on the existing VM, then its apply
          tf plan -input=false -var session_max_seconds=600 -out=/private/plans/control.tfplan || true
          tf apply -input=false /private/plans/control.tfplan || true ;;
      esac'
    vmaddr=module.vm.google_compute_instance.vm spath='scheduling[0].max_run_duration[0].seconds'
    ses_case() { # NAME MODE STATE HAVE TFOUT STOP [VAR=VALUE]...: STATE "-" keeps NAME's state directory; sets ses_rc
      local d="$sl/$1" mode=$2
      ses_rc=0
      mkdir -p "$d"
      if [[ $3 != - ]]; then
        printf '%s\n' "$3" >"$d/state"; printf '%s\n' "$4" >"$d/have"; printf '%s\n' "$5" >"$d/tfout"; printf '%s\n' "$6" >"$d/stop"
        rm -f "$d/sched"
      fi
      : >"$d/calls.log"; [[ -f "$d/scope.tsv" ]] || : >"$d/scope.tsv"
      shift 6
      (cd "$REPO" && env -u CLOUDFLARE_API_TOKEN LC_ALL=C PATH="$S/shim:$PATH" HOME="$S/home" SD="$d" MODE="$mode" "$@" \
        bash -c "$ses_script") >"$d/out" 2>"$d/err" || ses_rc=$?
      printf '%s\n' "$ses_rc" >"$d/rc"
      for f in out err calls.log; do cp "$d/$f" "$lc/session-$(basename "$d")-$f.txt"; done
    }
    tfw() { grep -E '^terraform (plan|apply) ' "$1" || true; }                              # Terraform writes
    vmw() { grep -E '^gcloud compute instances (start|stop|set-scheduling) ' "$1" || true; } # VM power/scheduling
    in_order() { # LOG ERE...: each expression matches a line after the previous one's
      local log=$1 n=0 p; shift
      for p in "$@"; do n="$(line_of "$p" "$log" "$n")"; [[ -n "$n" ]] || return 1; done
    }
    tok=CLOUDFLARE_API_TOKEN=g070-check-placeholder
    ses_case running-same start RUNNING 7200 7200 2026-10-27T11:00:00Z
    d="$sl/running-same"
    [[ $ses_rc == 0 && -z "$(tfw "$d/calls.log")$(vmw "$d/calls.log")" ]] && grep -q '^remote start.sh' "$d/calls.log" &&
      grep -qF 'scheduled stop 2026-10-27T11:00:00Z' "$d/err" && grep -qx $'session\t7200s\tscheduled stop 2026-10-27T11:00:00Z' "$d/out" ||
      fail "session running-same: exit $ses_rc; want the resume checks only (no Terraform write, no start) with the scheduled stop shown"
    ses_case running-different start RUNNING 7200 7200 2026-10-27T11:00:00Z DUR=600
    d="$sl/running-different"
    [[ $ses_rc == 2 && -z "$(tfw "$d/calls.log")$(vmw "$d/calls.log")" ]] && ! grep -q '^remote ' "$d/calls.log" &&
      grep -qF 'stop it first, then start --duration' "$d/err" ||
      fail "session running-different: exit $ses_rc; want a refusal (exit 2) with no Terraform write, VM call or VM-side step"
    ses_case stopped-same start TERMINATED 7200 7200 ''
    d="$sl/stopped-same"
    [[ $ses_rc == 0 && -z "$(tfw "$d/calls.log")" && "$(vmw "$d/calls.log" | grep -c .)" == 1 ]] &&
      grep -qx $'session\t7200s\tscheduled stop 2026-10-27T14:00:00.000-07:00' "$d/out" &&
      grep -qF 'session 7200s, scheduled stop 2026-10-27T14:00:00.000-07:00' "$d/calls.log" ||
      fail "session stopped-same (no token): exit $ses_rc; want a start with no Terraform write and the armed stop recorded"
    # Length changes (A2 fallback, round 3): a fresh TERMINATED read, gcloud set-scheduling with the provisioned
    # attributes, the full readback, a saved refresh-only plan, label check, refresh scope check, its apply, both
    # sources re-read, then the start; never a normal plan (the fake terraform logs and fails the apply of a normal plan
    # that would replace the VM; every calls.log is scanned for it below).
    sched_line() { # SECONDS: the exact set-scheduling argv (STOP kept, no --clear-*, no --termination-time)
      printf 'gcloud compute instances set-scheduling staging-vm --zone us-central1-a --project example-project --impersonate-service-account staging-deployer@example-project.iam.gserviceaccount.com --max-run-duration=%ss --instance-termination-action=STOP --maintenance-policy=MIGRATE --restart-on-failure --provisioning-model=STANDARD' "$1"
    }
    seq_ok() { # LOG SECONDS: the whole ordered sequence of a length change
      in_order "$1" '^terraform output ' '^gcloud compute instances describe .*--format value\(status\)$' \
        "^$(sched_line "$2" | sed 's/[.[\*^$()+?{|]/\\&/g')\$" '^gcloud compute instances describe .*instanceTerminationAction' \
        "^terraform plan -refresh-only .*session_max_seconds=$2 -out=" '^label_check$' '^scope_check$' '^terraform apply -input=false /private/plans/[^ ]*-refresh\.tfplan$' \
        '^gcloud compute instances describe .*maxRunDuration' '^terraform output ' '^gcloud compute instances start ' '^remote start.sh'
    }
    normal_plan() { grep -E '^terraform plan ' "$1" | grep -vE '^terraform plan -refresh-only ' || true; } # a normal plan in start
    held() { # NAME TEXT: that case's own exit 3 (read from its rc file), no VM start or VM-side step, the partial state and the recovery printed, TEXT in stderr
      local d="$sl/$1"
      [[ $(<"$d/rc") == 3 ]] && ! grep -qE '^(gcloud compute instances start |remote )' "$d/calls.log" &&
        grep -qF 'staging-vm not started; partial state: the instance reports ' "$d/err" &&
        grep -qF 'recovery: with the VM stopped run start --duration ' "$d/err" && grep -qF -- "$2" "$d/err"
    }
    drift_ok() { # FROM TO: the refresh scope rows of the legitimate drift (the length plus the status value)
      printf 'drift\t%s\tupdate\ndiff\t%s\t%s\t%s\t%s\ndiff\t%s\tcurrent_status\t"RUNNING"\t"TERMINATED"\n' \
        "$vmaddr" "$vmaddr" "$spath" "$1" "$2" "$vmaddr"
    }
    for c2 in 600:7200 7200:600; do # stopped-different: 7200 -> 600 and 600 -> 7200 (both through set-scheduling)
      from=${c2%%:*} to=${c2#*:}
      mkdir -p "$sl/stopped-$to"; drift_ok "$from" "$to" >"$sl/stopped-$to/scope.tsv"
      ses_case "stopped-$to" start TERMINATED "$from" "$from" '' DUR="$to" "$tok"
      d="$sl/stopped-$to"
      [[ $ses_rc == 0 && "$(cat "$d/have")" == "$to" && "$(cat "$d/tfout")" == "$to" ]] && seq_ok "$d/calls.log" "$to" &&
        [[ "$(grep -c '^gcloud compute instances set-scheduling ' "$d/calls.log")" == 1 && -z "$(normal_plan "$d/calls.log")" ]] &&
        grep -qxF "$(sched_line "$to")" "$d/calls.log" && grep -qx $'session\t'"$to"$'s\tscheduled stop 2026-10-27T14:00:00.000-07:00' "$d/out" ||
        fail "session stopped-$to: exit $ses_rc; want fresh read, set-scheduling (exact argv), readback, refresh-only plan, label check, scope check, apply, re-read, start in that order, instance and Terraform at $to, no normal plan"
    done
    mkdir -p "$sl/stopped-default"; drift_ok 600 7200 >"$sl/stopped-default/scope.tsv"
    ses_case stopped-default start TERMINATED 600 600 '' "$tok"
    d="$sl/stopped-default"
    [[ $ses_rc == 0 ]] && grep -qxF "$(sched_line 7200)" "$d/calls.log" && seq_ok "$d/calls.log" 7200 && [[ -z "$(normal_plan "$d/calls.log")" ]] ||
      fail "session stopped-default: exit $ses_rc; a start without --duration must set the configured 7200 by the same path, not keep the previous 600"
    ses_case no-token start TERMINATED 7200 7200 '' DUR=600
    d="$sl/no-token"
    [[ $ses_rc == 2 && -z "$(tfw "$d/calls.log")$(vmw "$d/calls.log")" ]] && grep -qF 'set CLOUDFLARE_API_TOKEN' "$d/err" ||
      fail "session no-token: exit $ses_rc; a length-changing start must refuse before set-scheduling, any plan or start without the token"
    ses_case stopped-not-terminated start STOPPED 7200 7200 '' DUR=600 "$tok"
    d="$sl/stopped-not-terminated"
    [[ $ses_rc == 3 && -z "$(tfw "$d/calls.log")$(vmw "$d/calls.log")" ]] && grep -qF 'not TERMINATED' "$d/err" ||
      fail "session stopped-not-terminated: exit $ses_rc; a length change on a VM that is not TERMINATED must HOLD before set-scheduling"
    # Instance and Terraform disagree: HOLD with the manual steps, except the recoverable partial state below.
    ses_case drift start TERMINATED 7200 600 '' DUR=3600 "$tok"
    d="$sl/drift"
    [[ $ses_rc == 3 && -z "$(tfw "$d/calls.log")$(vmw "$d/calls.log")" ]] && grep -qF 'session length but Terraform records 600s' "$d/err" &&
      grep -qF 'run start --duration 7200 (it refreshes Terraform' "$d/err" && grep -qF 'set-scheduling staging-vm --zone us-central1-a --project example-project --impersonate-service-account staging-deployer@example-project.iam.gserviceaccount.com --max-run-duration=600s --instance-termination-action=STOP' "$d/err" ||
      fail "session drift: exit $ses_rc; instance 7200, Terraform 600, requested 3600 must HOLD (exit 3) with the manual reconcile steps and nothing changed"
    ses_case running-drift start RUNNING 7200 600 2026-10-27T11:00:00Z "$tok"
    d="$sl/running-drift"
    [[ $ses_rc == 3 && -z "$(tfw "$d/calls.log")$(vmw "$d/calls.log")" ]] && ! grep -q '^remote ' "$d/calls.log" &&
      grep -qF 'session length but Terraform records 600s' "$d/err" ||
      fail "session running-drift: exit $ses_rc; a running VM whose instance and Terraform disagree must HOLD with nothing changed"
    # Recovery: the stopped instance already reports the requested length, Terraform differs: refresh, verify, start.
    mkdir -p "$sl/recovery"; drift_ok 600 7200 >"$sl/recovery/scope.tsv"
    ses_case recovery start TERMINATED 7200 600 '' "$tok"
    d="$sl/recovery"
    [[ $ses_rc == 0 && "$(cat "$d/tfout")" == 7200 && -z "$(grep '^gcloud compute instances set-scheduling ' "$d/calls.log")" ]] &&
      in_order "$d/calls.log" '^gcloud compute instances describe .*instanceTerminationAction' '^terraform plan -refresh-only .*session_max_seconds=7200 ' \
        '^label_check$' '^scope_check$' '^terraform apply ' '^terraform output ' '^gcloud compute instances start ' && grep -qF 'recovery: staging-vm already reports 7200s' "$d/err" ||
      fail "session recovery: exit $ses_rc; instance 7200 == requested, Terraform 600 must refresh, verify and start without set-scheduling"
    # Failure injection: each HOLDs (exit 3) with the partial state and recovery printed and never starts the VM.
    ses_case setsched-fails start TERMINATED 7200 7200 '' DUR=600 "$tok" SETSCHED=fail
    d="$sl/setsched-fails"
    held setsched-fails 'gcloud compute instances set-scheduling failed' && [[ -z "$(tfw "$d/calls.log")" ]] ||
      fail "session setsched-fails: exit $ses_rc; a failed set-scheduling must HOLD before any plan, with the partial state and recovery"
    ses_case readback-length start TERMINATED 7200 7200 '' DUR=600 "$tok" SETSCHED=nochange
    d="$sl/readback-length"
    held readback-length 'maxRunDuration.seconds [7200], not 600' && [[ -z "$(tfw "$d/calls.log")" ]] ||
      fail "session readback-length: exit $ses_rc; a readback without the new length must HOLD before any plan"
    ses_case readback-attribute start TERMINATED 7200 7200 '' DUR=600 "$tok" SETSCHED=other
    d="$sl/readback-attribute"
    held readback-attribute 'automaticRestart [False], not true' && [[ -z "$(tfw "$d/calls.log")" ]] ||
      fail "session readback-attribute: exit $ses_rc; another scheduling attribute changed must HOLD before any plan"
    ses_case readback-unreadable start TERMINATED 7200 7200 '' DUR=600 "$tok" DESC=sched
    d="$sl/readback-unreadable"
    held readback-unreadable 'the scheduling of staging-vm could not be read back' && [[ -z "$(tfw "$d/calls.log")" ]] ||
      fail "session readback-unreadable: exit $ses_rc; an unreadable readback must HOLD before any plan"
    ses_case refresh-plan-fails start TERMINATED 7200 7200 '' DUR=600 "$tok" PLANRC=1
    d="$sl/refresh-plan-fails"
    held refresh-plan-fails 'terraform plan -refresh-only failed (exit 1)' && ! grep -qE '^terraform apply ' "$d/calls.log" ||
      fail "session refresh-plan-fails: exit $ses_rc; a failed refresh-only plan must HOLD before any apply"
    ses_case refresh-plan-nodrift start TERMINATED 7200 7200 '' DUR=600 "$tok" PLANRC=0
    d="$sl/refresh-plan-nodrift"
    held refresh-plan-nodrift 'the refresh-only plan shows no drift' && ! grep -qE '^terraform apply ' "$d/calls.log" ||
      fail "session refresh-plan-nodrift: exit $ses_rc; a refresh-only plan with exit 0 must HOLD before any apply"
    ses_case label-hold start TERMINATED 7200 7200 '' DUR=600 "$tok" LABELFAIL=1
    d="$sl/label-hold"
    held label-hold 'the label check refused the refresh-only plan' && ! grep -qE '^(terraform apply |scope_check)' "$d/calls.log" ||
      fail "session label-hold: exit $ses_rc; a label check that refuses the refresh-only plan must HOLD with the partial state and recovery, before the scope check and apply"
    mkdir -p "$sl/scope-hold"; { drift_ok 7200 600; printf 'change\t%s\tdelete,create\n' "$vmaddr"; } >"$sl/scope-hold/scope.tsv"
    ses_case scope-hold start TERMINATED 7200 7200 '' DUR=600 "$tok"
    d="$sl/scope-hold"
    held scope-hold "it proposes delete,create of $vmaddr" && grep -q '^scope_check$' "$d/calls.log" && ! grep -qE '^terraform apply ' "$d/calls.log" ||
      fail "session scope-hold: exit $ses_rc; a refresh-only plan outside the length-drift scope must HOLD before apply and start"
    for m in fail noop unreadable; do
      mkdir -p "$sl/apply-$m"; drift_ok 7200 600 >"$sl/apply-$m/scope.tsv"
      ses_case "apply-$m" start TERMINATED 7200 7200 '' DUR=600 "$tok" APPLY="$m"
    done
    held apply-fail 'terraform apply of the refresh-only plan failed' ||
      fail "session apply-fail: exit $ses_rc; a failed refresh apply must HOLD before the start with the partial state"
    held apply-noop 'after the refresh the instance reports [600]s and Terraform [7200]s, not 600s' ||
      fail "session apply-noop: Terraform's output still old after the apply must HOLD before the start"
    held apply-unreadable "Terraform's max_run_duration_seconds output could not be read after the refresh" ||
      fail "session apply-unreadable: an unreadable output after the apply must HOLD before the start"
    for f in out err calls.log; do cp "$sl/apply-fail/$f" "$lc/session-apply-fail-first-$f.txt"; done
    # The recovery of that partial state (instance 600, Terraform 7200): the same start --duration 600 refreshes and
    # starts without another set-scheduling; a start with another length (the default 7200) HOLDs with the manual steps.
    cp -r "$sl/apply-fail" "$sl/apply-fail-default"
    ses_case apply-fail start - - - - DUR=600 "$tok"
    d="$sl/apply-fail"
    [[ $ses_rc == 0 && "$(cat "$d/tfout")" == 600 && -z "$(grep '^gcloud compute instances set-scheduling ' "$d/calls.log")" ]] &&
      grep -q '^gcloud compute instances start ' "$d/calls.log" ||
      fail "session apply-fail rerun: exit $ses_rc; start --duration 600 on the partial state must refresh, verify and start without set-scheduling"
    ses_case apply-fail-default start - - - - "$tok"
    d="$sl/apply-fail-default"
    [[ $ses_rc == 3 && -z "$(tfw "$d/calls.log")$(vmw "$d/calls.log")" ]] && grep -qF 'run start --duration 600 (it refreshes Terraform' "$d/err" ||
      fail "session apply-fail-default: exit $ses_rc; the partial state with another requested length must HOLD with the manual steps, nothing changed"
    # Control: the fake terraform catches the apply of a normal plan that would replace the VM.
    ses_case fake-replace-control replace-control TERMINATED 7200 7200 ''
    grep -q '^terraform-fake: REPLACEMENT applied' "$sl/fake-replace-control/calls.log" ||
      fail "session fake-replace-control: the fake terraform does not flag a replacing normal plan"
    ses_case provision-existing provision TERMINATED 3600 3600 '' "$tok" YES=1
    d="$sl/provision-existing"
    [[ $ses_rc == 0 && -z "$(vmw "$d/calls.log")" ]] && grep -qE '^terraform plan .*session_max_seconds=3600( |$)' "$d/calls.log" &&
      ! grep -q '^deploy ' "$d/calls.log" || fail "session provision-existing: exit $ses_rc; want the observed 3600 on the plan, no start, no deploy"
    ses_case provision-existing-duration provision TERMINATED 3600 3600 '' "$tok" DUR=600 YES=1
    d="$sl/provision-existing-duration"
    [[ $ses_rc == 2 && -z "$(tfw "$d/calls.log")" ]] && grep -qF 'use start --duration' "$d/err" ||
      fail "session provision-existing-duration: exit $ses_rc; --duration on an existing VM must be refused before any plan"
    ses_case provision-first provision ABSENT none none '' "$tok" DUR=600 YES=1
    d="$sl/provision-first"
    [[ $ses_rc == 0 ]] && grep -qE '^terraform plan .*session_max_seconds=600( |$)' "$d/calls.log" && grep -q '^deploy ' "$d/calls.log" ||
      fail "session provision-first: exit $ses_rc; the first allocation must plan --duration 600 and deploy"
    ses_case provision-first-default provision ABSENT none none '' "$tok" YES=1
    [[ $ses_rc == 0 ]] && grep -qE '^terraform plan .*session_max_seconds=7200( |$)' "$sl/provision-first-default/calls.log" ||
      fail "session provision-first-default: exit $ses_rc; the first allocation without --duration must plan the configured 7200"
    # ABSENT came from the positive path: the failed describe was followed by the exact-name list (catches restoring
    # `|| VMSTATE=ABSENT`, which skips the list; the fake answers no other list form, so a regex/prefix filter fails above).
    grep -qE '^gcloud compute instances list --zones us-central1-a .* --filter name=staging-vm --format value\(name\)$' "$d/calls.log" ||
      fail "session provision-first: absence was not confirmed by the exact-name instance list"

    # Unreadable VM (a source review finding): an existing VM (custom length 3600) whose state cannot be read HOLDs
    # (exit 3) in every command, before any Terraform plan/apply, deploy, VM-side step or VM start/stop, and no length
    # (the default 7200 included) is selected. Each case varies one read; the comment names the mutation it catches.
    unread_case() { # NAME MODE STATE MUTATION [VAR=VALUE]...
      local name=$1 mode=$2 st=$3 mut=$4 d="$sl/$1"
      shift 4
      ses_case "$name" "$mode" "$st" 3600 3600 '' "$tok" "$@"
      [[ $ses_rc == 3 && -z "$(tfw "$d/calls.log")$(vmw "$d/calls.log")" ]] &&
        ! grep -qE '^(deploy|remote) |session_max_seconds=' "$d/calls.log" &&
        grep -qF 'HOLD: the state of staging-vm could not be read' "$d/err" ||
        fail "session $name: exit $ses_rc; an unreadable VM must HOLD (exit 3) before any write or length selection ($mut)"
    }
    # (a) describe and list both fail: catches restoring `|| VMSTATE=ABSENT` and a failed list taken for an empty one
    unread_case unread-provision-running provision RUNNING "describe/list failure taken for absence" DESC=fail LIST=fail
    unread_case unread-provision-running-yes provision RUNNING "describe/list failure taken for absence" DESC=fail LIST=fail YES=1
    unread_case unread-provision-stopped provision TERMINATED "describe/list failure taken for absence" DESC=fail LIST=fail
    unread_case unread-provision-stopped-yes provision TERMINATED "describe/list failure taken for absence" DESC=fail LIST=fail YES=1
    # (b) a transient describe failure while the list still shows the VM: catches a non-empty list taken for absence
    unread_case listed-provision-running provision RUNNING "a listed VM taken for absence" DESC=fail
    unread_case listed-provision-stopped-yes provision TERMINATED "a listed VM taken for absence" DESC=fail YES=1
    # (c) the describe succeeds with no status: catches restoring the empty-status default `: "${VMSTATE:=ABSENT}"`
    unread_case nostatus-provision-stopped-yes provision TERMINATED "an empty status taken for absence" DESC=empty YES=1
    # (e) the other VM commands: catches any of them turning an unreadable VM into ABSENT (old: start/stop exit 2,
    # status/retire a default-length plan)
    unread_case unread-start start TERMINATED "start on an unreadable VM" DESC=fail LIST=fail
    unread_case unread-stop stop RUNNING "stop on an unreadable VM" DESC=fail LIST=fail
    unread_case unread-status status TERMINATED "status on an unreadable VM" DESC=fail LIST=fail
    unread_case unread-retire retire TERMINATED "retire on an unreadable VM" DESC=fail LIST=fail YES=1
    # status and retire on a readable VM whose session describe fails: catches session_var falling back to the
    # configured default for an existing instance (old: a refresh-only/destroy plan with session_max_seconds=7200)
    for m in status retire; do
      ses_case "session-unread-$m" "$m" TERMINATED 3600 3600 '' "$tok" DESC=session YES=1
      d="$sl/session-unread-$m"
      [[ $ses_rc == 3 && -z "$(tfw "$d/calls.log")" ]] && grep -qF 'HOLD: the session length of staging-vm could not be read' "$d/err" ||
        fail "session session-unread-$m: exit $ses_rc; an unreadable session length must HOLD before any plan, not take the default"
    done
    # status and retire on a readable VM whose session describe succeeds with a bad length (a source review
    # finding): HOLD (exit 3) naming it missing or malformed, before any Terraform plan/apply (refresh-only, destroy),
    # VM call, write or VM-side step, and no length (the default 7200 included) reaches the transcript. retire runs on
    # the stopped VM (a running one is stopped first). The comment names the mutation each catches.
    bad_case() { # NAME MODE STATE HAVE STOP MUTATION: the HOLD names the reported length exactly ([] when missing)
      local name=$1 mode=$2 d="$sl/$1"
      ses_case "$name" "$mode" "$3" "$4" 3600 "$5" "$tok" YES=1
      [[ $ses_rc == 3 && -z "$(tfw "$d/calls.log")$(vmw "$d/calls.log")" ]] &&
        ! grep -qE '^(deploy|remote|inventory|record_state) |session_max_seconds=' "$d/calls.log" &&
        ! grep -qF 7200 "$d/out" "$d/err" "$d/calls.log" &&
        grep -qF "HOLD: the session length of staging-vm is missing or malformed (max run duration [$4])" "$d/err" ||
        fail "session $name: exit $ses_rc; a missing or malformed session length must HOLD (exit 3) naming [$4] before any plan or write ($6)"
    }
    tstamp=2026-10-27T11:00:00Z
    for m in status retire; do
      bst=TERMINATED; [[ $m == retire ]] || bst=RUNNING
      # (i) empty length and timestamp: catches restoring the silent default for a VM that reports no length
      bad_case "session-empty-$m" "$m" "$bst" '' '' "the configured default kept for an empty length"
      # (ii) no length before a termination timestamp: catches restoring `IFS=$'\t' read` (the empty column collapses,
      # the timestamp becomes the length and the HOLD names it instead of []) as well as the silent default
      bad_case "session-gap-$m" "$m" "$bst" '' "$tstamp" "the timestamp shifted into the length"
      # (iii) a non-numeric length: catches accepting any non-empty value, or a value with a unit suffix
      bad_case "session-abc-$m" "$m" "$bst" abc '' "a non-numeric length accepted"
      bad_case "session-suffix-$m" "$m" "$bst" 3600s '' "a length with a unit suffix accepted"
    done
    # (iii) zero seconds: catches a length check of ^[0-9]+$
    bad_case session-zero-status status TERMINATED 0 '' "a zero length accepted"
    # The real vm_session's field split on verbatim describe output: each column lands in its own variable, a missing
    # one stays empty (catches restoring `IFS=$'\t' read`, which collapses an empty middle column and shifts the next)
    parse_case() { # NAME RAW WANT: RAW is printf'd into the fake's raw output
      local d="$sl/parse-$1"
      mkdir -p "$d"
      # shellcheck disable=SC2059 # RAW is the format: its \t and \n escapes are the describe output's separators
      printf "$2" >"$d/raw"
      ses_case "parse-$1" parse RUNNING - - -
      [[ $ses_rc == 0 && "$(cat "$d/out")" == "$3" ]] ||
        fail "session parse $1: exit $ses_rc, vm_session read [$(cat "$d/out")], want [$3]"
    }
    parse_case all 'RUNNING\t3600\t2026-10-27T11:00:00Z\n' 'RUNNING|3600|2026-10-27T11:00:00Z'
    parse_case no-length 'RUNNING\t\t2026-10-27T11:00:00Z\n' 'RUNNING||2026-10-27T11:00:00Z'
    parse_case stopped 'TERMINATED\t3600\t\n' 'TERMINATED|3600|'
    parse_case stopped-no-length 'TERMINATED\t\t\n' 'TERMINATED||'
    parse_case status-only 'RUNNING\n' 'RUNNING||'
    parse_case empty '' '||'
    # retire on a positively absent VM still plans with the configured default (catches a HOLD on confirmed absence)
    ses_case retire-absent retire ABSENT none none '' "$tok" PLANRC=0
    d="$sl/retire-absent"
    [[ $ses_rc == 0 ]] && grep -qE '^terraform plan -destroy .*session_max_seconds=7200( |$)' "$d/calls.log" ||
      fail "session retire-absent: exit $ses_rc; a confirmed-absent VM must plan the destroy with the configured 7200"

    # The fake terraform never applied a replacing normal plan and gcloud never got set-scheduling on a VM that was not
    # TERMINATED, in any case above (the control case aside, which proves the fake flags it).
    bad="$(grep -lE '^(terraform|gcloud)-fake:' "$sl"/*/calls.log 2>/dev/null | grep -v '/fake-replace-control/' || true)"
    [[ -z "$bad" ]] || fail "session fakes: a replacing normal plan applied or set-scheduling on a VM not TERMINATED (${bad//$'\n'/ })"

    # (c) refresh_scope_check on refresh_scope_rows fixtures (Terraform 7200, requested 600): only the length drift of
    # module.vm.google_compute_instance.vm (7200 -> 600), with at most current_status/cpu_platform beside it, and no
    # proposed resource change, is accepted.
    scope_case() { # NAME WANT_REASON ROWS: WANT_REASON "" = accepted
      local d="$sl/scope-$1"
      mkdir -p "$d"; printf '%s' "$3" >"$d/scope.tsv"
      ses_case "scope-$1" scope TERMINATED 600 7200 '' SHAVE=7200 SWANT=600
      if [[ -z $2 ]]; then
        [[ $ses_rc == 0 ]] && grep -qx $'scope\taccepted' "$d/out" || fail "scope $1: exit $ses_rc, want accepted"
      else
        [[ $ses_rc == 3 ]] && grep -qF 'HOLD: the refresh-only plan is not the session-length drift alone (' "$d/err" && grep -qF -- "$2" "$d/err" &&
          ! grep -q 'accepted' "$d/out" || fail "scope $1: exit $ses_rc, want HOLD naming [$2] (lifecycle/session-scope-$1-err.txt)"
      fi
    }
    ok_rows="$(drift_ok 7200 600)"$'\n'
    vmd() { printf 'diff\t%s\t%s\t%s\t%s\n' "$vmaddr" "$@"; } # PATH BEFORE AFTER: one more VM drift leaf
    scope_case accept "" "$ok_rows"
    scope_case accept-length-only "" "$(printf 'drift\t%s\tupdate\n' "$vmaddr")"$'\n'"$(vmd "$spath" 7200 600)"$'\n'
    scope_case accept-cpu-platform "" "$ok_rows$(vmd cpu_platform '"Intel Broadwell"' '""')"$'\n'
    # Recorded-state lag (benign_drift): the leaves a never-refreshed VM's first custom-duration start reports beside the length.
    # The attachment must name this configuration's exact project, zone and VM (example-project, us-central1-a, staging-vm).
    lagu='"https://www.googleapis.com/compute/v1/projects/example-project/zones/us-central1-a/instances/staging-vm"'
    lagd() { lagd_a=$1; printf 'drift\t%s\tupdate\n' "$1"; shift; while (($#)); do printf 'diff\t%s\t%s\t%s\t%s\n' "$lagd_a" "$1" "$2" "$3"; shift 3; done; }
    lag_rows() { # one drift block per lagging address, then the VM's own empty-default leaves
      lagd_a=module.vm.google_compute_address.ip
      lagd "$lagd_a" users '[]' "[$lagu]" 'users[0]' null "$lagu"
      lagd_a=module.vm.google_compute_disk.data
      lagd "$lagd_a" last_attach_timestamp null '"2026-10-28T11:46:34.448-07:00"' users '[]' "[$lagu]" 'users[0]' null "$lagu"
      for lagd_a in module.vm.google_compute_firewall.edge 'module.vm.google_compute_firewall.iap[0]'; do
        lagd "$lagd_a" source_service_accounts null '[]' source_tags null '[]' target_service_accounts null '[]'
      done
      lagd_a='module.vm.google_compute_instance_iam_member.os_login["user:op@example.net"]'
      lagd "$lagd_a" etag '"BwZe6vxqeSo="' '"BwZe6v6dTY4="'
    }
    vm_lag="$(printf 'diff\t%s\tresource_policies\tnull\t[]\ndiff\t%s\tboot_disk[0].initialize_params[0].replica_zones\tnull\t[]\ndiff\t%s\tboot_disk[0].initialize_params[0].resource_manager_tags\tnull\t{}' "$vmaddr" "$vmaddr" "$vmaddr")"
    scope_case accept-lag-leaves "" "$ok_rows$vm_lag"$'\n'"$(lag_rows)"$'\n'
    scope_case lag-attached-to-other-vm "module.vm.google_compute_address.ip also changes users" \
      "$ok_rows$(lagd module.vm.google_compute_address.ip users '[]' '["https://www.googleapis.com/compute/v1/projects/example-project/zones/us-central1-a/instances/other-vm"]')"$'\n'
    lagw='https://www.googleapis.com/compute/v1/projects'
    scope_case lag-same-vm-wrong-project "module.vm.google_compute_address.ip also changes users" \
      "$ok_rows$(lagd module.vm.google_compute_address.ip users '[]' "[\"$lagw/unrelated-project/zones/us-central1-a/instances/staging-vm\"]")"$'\n'
    scope_case lag-same-vm-wrong-zone "module.vm.google_compute_address.ip also changes users" \
      "$ok_rows$(lagd module.vm.google_compute_address.ip users '[]' "[\"$lagw/example-project/zones/europe-west1-b/instances/staging-vm\"]")"$'\n'
    scope_case lag-disk-item-wrong-project "module.vm.google_compute_disk.data also changes users[0]" \
      "$ok_rows$(lagd module.vm.google_compute_disk.data 'users[0]' null "\"$lagw/unrelated-project/zones/us-central1-a/instances/staging-vm\"")"$'\n'
    scope_case lag-disk-item-wrong-zone "module.vm.google_compute_disk.data also changes users[0]" \
      "$ok_rows$(lagd module.vm.google_compute_disk.data 'users[0]' null "\"$lagw/example-project/zones/europe-west1-b/instances/staging-vm\"")"$'\n'
    scope_case lag-nonempty-service-account "module.vm.google_compute_firewall.edge also changes target_service_accounts" \
      "$ok_rows$(lagd module.vm.google_compute_firewall.edge target_service_accounts null '["x@y.iam.gserviceaccount.com"]')"$'\n'
    scope_case lag-firewall-rule-changed "module.vm.google_compute_firewall.iap[0] also changes source_ranges" \
      "$ok_rows$(lagd 'module.vm.google_compute_firewall.iap[0]' source_ranges '["35.235.240.0/20"]' '["0.0.0.0/0"]')"$'\n'
    scope_case lag-etag-unknown "also changes etag" \
      "$ok_rows$(lagd 'module.vm.google_compute_instance_iam_member.os_login["user:op@example.net"]' etag '"a"' unknown)"$'\n'
    scope_case lag-vm-empty-default-nonempty "it also changes resource_policies" \
      "$ok_rows$(vmd resource_policies null '["projects/p/regions/r/resourcePolicies/x"]')"$'\n'
    scope_case replace "it proposes delete,create of $vmaddr" "$ok_rows$(printf 'change\t%s\tdelete,create' "$vmaddr")"$'\n'
    scope_case proposed-update "it proposes update of module.vm.google_compute_firewall.edge" \
      "$ok_rows$(printf 'change\tmodule.vm.google_compute_firewall.edge\tupdate')"$'\n'
    scope_case second-drift 'module.dns.cloudflare_dns_record.this["base"] drifts too' \
      "$ok_rows$(printf 'drift\tmodule.dns.cloudflare_dns_record.this["base"]\tupdate\ndiff\tmodule.dns.cloudflare_dns_record.this["base"]\tcontent\t"192.0.2.1"\t"192.0.2.2"')"$'\n'
    scope_case delete-drift "$vmaddr drifts by delete, not update" "$(printf 'drift\t%s\tdelete\n' "$vmaddr")"$'\n'"$(vmd "$spath" 7200 null)"$'\n'
    scope_case unqualified-address "google_compute_instance.vm drifts too" \
      "$(printf 'drift\tgoogle_compute_instance.vm\tupdate\ndiff\tgoogle_compute_instance.vm\t%s\t7200\t600' "$spath")"$'\n'
    scope_case other-attribute "it also changes machine_type" "$ok_rows$(vmd machine_type '"e2-standard-2"' '"e2-standard-4"')"$'\n'
    scope_case scheduling-other "it also changes scheduling[0].automatic_restart" "$ok_rows$(vmd 'scheduling[0].automatic_restart' true false)"$'\n'
    scope_case unknown-attribute "it also changes metadata_fingerprint" "$ok_rows$(vmd metadata_fingerprint '"abc"' unknown)"$'\n'
    scope_case status-unknown "current_status becomes unknown" "$(printf 'drift\t%s\tupdate\n' "$vmaddr")"$'\n'"$(vmd "$spath" 7200 600)"$'\n'"$(vmd current_status '"RUNNING"' unknown)"$'\n'
    scope_case length-unknown "$spath becomes unknown, not 600" "$(printf 'drift\t%s\tupdate\n' "$vmaddr")"$'\n'"$(vmd "$spath" 7200 unknown)"$'\n'
    scope_case before-not-terraform "$spath is 3600 before, not Terraform's 7200" "$(drift_ok 3600 600)"$'\n'
    scope_case after-not-want "$spath becomes 900, not 600" "$(drift_ok 7200 900)"$'\n'
    scope_case no-drift "0 changes of $spath, not 1" ""
    scope_case unreadable-row "unreadable plan row" "$ok_rows$(printf 'bogus\t%s' "$vmaddr")"$'\n'

    # (d) The real refresh_scope_rows -> refresh_scope_check on hand-written refresh-only `terraform show -json`
    # documents. (c) above gives refresh_scope_check ready-made rows and so covers only its bash classification; this
    # layer runs the jq program itself (no-op filter, drift entries, leaf paths, path strings, after_unknown,
    # before/after values) on Terraform's plan structure: resource_changes[] (one unchanged no-op resource first in every
    # plan) and resource_drift[] with change.actions/before/after/after_unknown, after_unknown as Terraform writes it
    # (known leaves omitted or false, nested blocks kept as {} and []). jq is the pinned JQ_IMAGE through a jq shim on
    # the fixture PATH: docker run -i --rm --pull never --network none with the run's labels and a per-call name (swept
    # with every labelled container at exit). Each refused plan differs from the accepted one in one condition; the
    # comment names the mutation it catches.
    if [[ -z "$(image_id "$JQ_IMAGE")" ]]; then
      fail "jq image $JQ_IMAGE not prepared (run check.sh prepare): the real refresh_scope_rows fixtures did not run"
    else
      mkdir -p "$S/jqshim"
      printf '#!/bin/sh\nexport DOCKER_CONFIG=%s\nexec %s run -i --rm --pull never --network none %s --name "%s-$$" %s "$@"\n' \
        "$(q "${DOCKER_CONFIG:-$HOME/.docker}")" "$(q "$(command -v docker)")" "$(q "${LABELS[@]}")" "g070-check-$RUN_ID-jq" \
        "$(q "$JQ_IMAGE")" >"$S/jqshim/jq"
      chmod 755 "$S/jqshim/jq"
      # jq_vm SECONDS STATUS [PURPOSE [RESTART [METADATA]]]: google_compute_instance values as a plan shows them;
      # SECONDS "" leaves max_run_duration[0] without seconds, STATUS "" omits current_status (unknown after apply)
      jq_vm() {
        local st="" sec=""
        [[ -z $2 ]] || st="\"current_status\":\"$2\","
        [[ -z $1 ]] || sec=",\"seconds\":$1"
        printf '{"name":"staging-vm","machine_type":"e2-standard-2",%s"cpu_platform":"Intel Broadwell","labels":{"environment":"staging","purpose":"%s"},"metadata":{%s},"scheduling":[{"automatic_restart":%s,"instance_termination_action":"STOP","max_run_duration":[{"nanos":0%s}],"node_affinities":[],"on_host_maintenance":"MIGRATE","preemptible":false,"provisioning_model":"STANDARD"}]}' \
          "$st" "${3:-staging}" "${5:-\"enable-oslogin\":\"TRUE\"}" "${4:-true}" "$sec"
      }
      # jq_unk MRD METADATA [EXTRA]: an after_unknown; an unknown leaf is true; EXTRA adds top-level members
      jq_unk() { printf '{"labels":{},"metadata":{%s},"scheduling":[{"max_run_duration":[{%s}],"node_affinities":[]}]%s}' "$2" "$1" "${3:-}"; }
      # jq_rc ADDRESS TYPE ACTIONS BEFORE AFTER AFTER_UNKNOWN: one resource_changes[] or resource_drift[] entry
      jq_rc() {
        printf '{"address":"%s","mode":"managed","type":"%s","name":"%s","change":{"actions":%s,"before":%s,"after":%s,"after_unknown":%s}}' \
          "$1" "$2" "${1##*.}" "$3" "$4" "$5" "$6"
      }
      upd='["update"]' fw='{"name":"staging-iap","source_ranges":["35.235.240.0/20"]}' gci=google_compute_instance unk0="$(jq_unk '' '')"
      noop="$(jq_rc module.vm.google_compute_firewall.iap google_compute_firewall '["no-op"]' "$fw" "$fw" '{"source_ranges":[false]}')"
      run7200="$(jq_vm 7200 RUNNING)" term600="$(jq_vm 600 TERMINATED)"
      drift_vm="$(jq_rc "$vmaddr" "$gci" "$upd" "$run7200" "$term600" "$unk0")" # the legitimate drift, 7200 -> 600
      jq_case() { # NAME FROM WANT WANT_REASON CHANGES DRIFTS: WANT_REASON "" = accepted; CHANGES follow the no-op entry
        local name=$1 from=$2 want=$3 reason=$4 d="$sl/jq-$1"
        mkdir -p "$d"
        printf '{"format_version":"1.2","terraform_version":"1.16.7","resource_changes":[%s],"resource_drift":[%s]}\n' \
          "$noop${5:+,$5}" "$6" >"$d/plan.json"
        cp "$d/plan.json" "$lc/session-jq-$name-plan.json"
        ses_case "jq-$name" scope TERMINATED "$want" "$from" '' SHAVE="$from" SWANT="$want" REALJQ=1 PATH="$S/jqshim:$S/shim:$PATH"
        if [[ -z $reason ]]; then
          [[ $ses_rc == 0 ]] && grep -qx $'scope\taccepted' "$d/out" ||
            fail "jq $name: exit $ses_rc, want accepted (lifecycle/session-jq-$name-err.txt)"
        else
          [[ $ses_rc == 3 ]] && grep -qF -- "$reason" "$d/err" && ! grep -q accepted "$d/out" ||
            fail "jq $name: exit $ses_rc, want HOLD naming [$reason] (lifecycle/session-jq-$name-err.txt)"
        fi
      }
      # Accepted: catches dropping the no-op filter (the firewall becomes a proposed change), resource_drift not read,
      # a container path taken for a changed leaf (after_unknown's [{}] under max_run_duration), a broken path string,
      # swapped before/after, current_status not tolerated
      jq_case accept-7200-600 7200 600 "" "" "$drift_vm"
      jq_case accept-600-28800 600 28800 "" "" "$(jq_rc "$vmaddr" "$gci" "$upd" "$(jq_vm 600 TERMINATED)" "$(jq_vm 28800 TERMINATED)" "$unk0")"
      # A replacement proposed beside the drift, either order, or an ordinary update: catches resource_changes ignored
      jq_case replace-delete-create 7200 600 "it proposes delete,create of $vmaddr" \
        "$(jq_rc "$vmaddr" "$gci" '["delete","create"]' "$run7200" "$term600" "$unk0")" "$drift_vm"
      jq_case replace-create-delete 7200 600 "it proposes create,delete of $vmaddr" \
        "$(jq_rc "$vmaddr" "$gci" '["create","delete"]' "$run7200" "$term600" "$unk0")" "$drift_vm"
      jq_case proposed-update 7200 600 "it proposes update of module.vm.google_compute_firewall.iap" \
        "$(jq_rc module.vm.google_compute_firewall.iap google_compute_firewall "$upd" "$fw" '{"name":"staging-iap","source_ranges":["0.0.0.0/0"]}' '{}')" "$drift_vm"
      # The VM drifted away (deleted outside Terraform): catches drift actions not reaching the classification
      jq_case delete-drift 7200 600 "$vmaddr drifts by delete, not update" "" "$(jq_rc "$vmaddr" "$gci" '["delete"]' "$run7200" null '{}')"
      # A second resource drifted: catches a select that keeps only the VM's drift
      jq_case second-drift 7200 600 "module.dns.cloudflare_dns_record.base drifts too" "" "$drift_vm,$(jq_rc module.dns.cloudflare_dns_record.base \
        cloudflare_dns_record "$upd" '{"name":"staging.example.net","type":"A","content":"192.0.2.1"}' '{"name":"staging.example.net","type":"A","content":"192.0.2.2"}' '{}')"
      # Another known attribute on the VM, one inside scheduling, an unknown new key: catches diff rows limited to the
      # scheduling path or prefix, and dropping the after_unknown leaves or `$unk or`
      jq_case vm-label 7200 600 "it also changes labels.purpose" "" "$(jq_rc "$vmaddr" "$gci" "$upd" "$run7200" "$(jq_vm 600 TERMINATED qa)" "$unk0")"
      jq_case scheduling-other 7200 600 "it also changes scheduling[0].automatic_restart" "" \
        "$(jq_rc "$vmaddr" "$gci" "$upd" "$run7200" "$(jq_vm 600 TERMINATED staging false)" "$unk0")"
      jq_case metadata-unknown 7200 600 "it also changes metadata.probe" "" "$(jq_rc "$vmaddr" "$gci" "$upd" "$run7200" "$term600" "$(jq_unk '' '"probe":true')")"
      # A status value or the length known only after apply: catches AFTER reported as null instead of "unknown"
      jq_case status-unknown 7200 600 "current_status becomes unknown" "" \
        "$(jq_rc "$vmaddr" "$gci" "$upd" "$run7200" "$(jq_vm 600 '')" "$(jq_unk '' '' ',"current_status":true')")"
      jq_case length-unknown 7200 600 "$spath becomes unknown, not 600" "" \
        "$(jq_rc "$vmaddr" "$gci" "$upd" "$run7200" "$(jq_vm '' TERMINATED)" "$(jq_unk '"seconds":true' '')")"
      # Before is not Terraform's length / after is not the requested one: catches BEFORE or AFTER from the wrong side
      jq_case before-not-terraform 7200 600 "$spath is 1800 before, not Terraform's 7200" "" \
        "$(jq_rc "$vmaddr" "$gci" "$upd" "$(jq_vm 1800 RUNNING)" "$term600" "$unk0")"
      jq_case after-not-want 7200 600 "$spath becomes 5400, not 600" "" "$(jq_rc "$vmaddr" "$gci" "$upd" "$run7200" "$(jq_vm 5400 TERMINATED)" "$unk0")"
      # The same drift at an unqualified address: catches an address match on the resource type/name suffix
      jq_case unqualified-address 7200 600 "google_compute_instance.vm drifts too" "" \
        "$(jq_rc google_compute_instance.vm "$gci" "$upd" "$run7200" "$term600" "$unk0")"
      # No drift at all: catches accepting a refresh that saw nothing
      jq_case no-drift 7200 600 "0 changes of $spath, not 1" "" ""
    fi
    if grep -q . "$S/shim.log"; then fail "session fakes: executed $(cut -d' ' -f1 "$S/shim.log" | sort -u | tr '\n' ' ')"; fi
    why+="session length: --duration 90m/2h/7200 set 5400/7200/7200 s by set-scheduling (exact flags, no --clear-*/--termination-time) and a refresh-only plan, never a normal plan (session read, fresh status, set-scheduling, readback, refresh-only plan, label check, scope check, apply, re-read, start in order), malformed, out-of-range, repeated or misplaced --duration and --override session_max_seconds refused before any command, deploy without Terraform/start/stop/set-scheduling, SESSION_MAX_SECONDS selected and bounded; fakes: running same/different/drift, stopped same (no token), 7200->600 and 600->7200 with the exact argv and order, default 7200, missing token, not TERMINATED, drift HOLD with the manual steps, recovery (instance == requested != Terraform: refresh, verify, start), HOLDs without a VM start (partial state and recovery printed) on set-scheduling failure, readback length/attribute/unreadable, refresh plan exit 1 and 0, scope refusal, apply failure, output old, output unreadable; the failed apply recovered by the same start --duration, another length HOLDs; the fake terraform flags an applied replacing normal plan (control) and none was applied; provision existing (observed length; --duration refused) and first allocation (absence confirmed by the exact-name list); unreadable VM (describe and list failing, describe failing while listed, empty status) HOLDs provision with and without --yes, start, stop, status and retire before any write or default, an unreadable session length HOLDs status and retire, a confirmed-absent retire plans the default; refresh scope check: 4 accepted and 23 refused TSV row sets (the added accepted set is the decoded recorded-state lag of a never-refreshed VM; the added refusals are an attachment to another VM, to the same VM name in another project or zone (address and disk, list and item), a non-empty service account, a changed firewall range, an unknown etag and a non-empty resource_policies), and the real refresh_scope_rows jq program ($JQ_IMAGE, --network none) on refresh-only terraform show -json documents: 2 accepted, 14 refused (replacement both orders, proposed update, deleted VM, second drift, VM label, scheduling attribute, unknown metadata key, unknown status, unknown length, before/after mismatch, unqualified address, no drift); "
  fi
  [[ $ok == 1 ]] && ledger LIFECYCLE-DRYRUN PASS 0 "$STAGING_INPUTS" "$c" "${why}no shimmed tool executed (jq only in the pinned image); no create/delete, down -v, migrate down/reapply, prune or test overlay; destroy only in retire, apply only in provision, of retire's own saved destroy plan or of start --duration's own saved refresh-only plan, compose down only in reset; reset removes exactly the two data volumes" \
    || ledger LIFECYCLE-DRYRUN FAIL "${rc_all:-1}" "$STAGING_INPUTS" "$c" "${why% ; }"
fi

if selected STAGING-GREP; then
  hits="" files=()
  # G072: EP-1's hostile fixture (infra/staging/tests/fixtures/) dumps environments on purpose and is never scanned.
  while IFS= read -r f; do
    case "$f" in infra/staging/check.sh | infra/staging/tests/fixtures/*) ;; *) files+=("$f") ;; esac
  done < <(cd "$REPO" && git ls-files -- 'infra/staging/*.sh')
  for f in "${files[@]}"; do
    body="$(grep -nvE '^[[:space:]]*#' "$REPO/$f" || true)"
    # A fixture's grep pattern naming a fake's recorded call ("container inspect ..." in a call log) is not a command.
    hits+="$(grep -E '(docker|container)[[:space:]]+inspect' <<<"$body" | grep -vE '(--format|[[:space:]]-f[[:space:]=])' |
      grep -vE "grep[[:space:]]+(-[A-Za-z]+[[:space:]]+)*'[^']*(docker|container)[[:space:]]+inspect[^']*'" | sed "s|^|$f:|; s|\$| [inspect without --format]|" || true)"$'\n'
    hits+="$(grep -E '(docker|container)[[:space:]]+inspect' <<<"$body" | grep -E '\.Config\.Env|\.Config[[:space:]]*\}\}|json[[:space:]]+\.Config[[:space:]]*\}' | sed "s|^|$f:|; s|\$| [inspect of the container environment]|" || true)"$'\n'
    hits+="$(grep -E 'compose([^|;&]*)[[:space:]]config([[:space:]]|$|\))' <<<"$body" | grep -viE 'dry' | sed "s|^|$f:|; s|\$| [compose config outside a dry-run line]|" || true)"$'\n'
    hits+="$(grep -E 'config\.v2\.json|/proc/[^[:space:]]*/environ' <<<"$body" | sed "s|^|$f:|; s|\$| [container environment file]|" || true)"$'\n'
    hits+="$(grep -E 'identity\.yml|browser\.yml|secret-scan\.yml' <<<"$body" | sed "s|^|$f:|; s|\$| [test overlay]|" || true)"$'\n'
  done
  # Session length (Plan r1.9): one default everywhere it is stated, the policy bounds in the CLI and the module, the
  # documented interface, and the native max_run_duration as the only timer (no scheduler, cron or at job).
  SESSION_PIN=7200
  pin() { # FILE VALUE WHAT: VALUE must be the pin
    [[ "$2" == "$SESSION_PIN" ]] || hits+="$1: $3 is '${2//$'\n'/ }', not $SESSION_PIN [session default pin]"$'\n'
  }
  tfdef() { awk '/^variable "session_max_seconds"/ {b = 1} b && /^[[:space:]]*default[[:space:]]*=/ {gsub(/[^0-9]/, ""); print; exit} b && /^}/ {exit}' "$REPO/$1"; }
  pin infra/staging/lib.sh "$(sed -n -E 's/^SESSION_DEFAULT_SECONDS=([0-9]+)$/\1/p' "$REPO/infra/staging/lib.sh")" SESSION_DEFAULT_SECONDS
  for f in infra/terraform/modules/gcp-staging-vm/variables.tf infra/terraform/environments/example/variables.tf; do
    pin "$f" "$(tfdef "$f")" 'the session_max_seconds default'
  done
  pin infra/terraform/environments/example/terraform.tfvars.example \
    "$(sed -n -E 's/^session_max_seconds[[:space:]]*=[[:space:]]*([0-9]+)[[:space:]]*$/\1/p' "$REPO/infra/terraform/environments/example/terraform.tfvars.example")" session_max_seconds
  pin infra/staging/examples/staging.conf.example \
    "$(sed -n -E 's/^#?[[:space:]]*SESSION_MAX_SECONDS=([0-9]+)$/\1/p' "$REPO/infra/staging/examples/staging.conf.example")" SESSION_MAX_SECONDS
  pin docs/foundation/STAGING.md "$(sed -n -E 's/^\| Session length [^|]*\| ([0-9]+) \(2 h\) \|.*/\1/p' "$REPO/docs/foundation/STAGING.md")" 'the Session length row'
  v="$({ grep -oE '\(default [0-9]+ s' "$REPO/infra/terraform/README.md" || true; } | grep -oE '[0-9]+' | sort -u || true)"
  pin infra/terraform/README.md "$v" 'the session_max_seconds default'
  grep -qx 'SESSION_FLOOR=600' "$REPO/infra/staging/lib.sh" && grep -qx 'SESSION_CEILING=28800' "$REPO/infra/staging/lib.sh" ||
    hits+="infra/staging/lib.sh: SESSION_FLOOR/SESSION_CEILING are not 600/28800 [session bounds]"$'\n'
  grep -qF 'var.session_max_seconds >= 600 && var.session_max_seconds <= 28800 && floor(var.session_max_seconds) == var.session_max_seconds' \
    "$REPO/infra/terraform/modules/gcp-staging-vm/variables.tf" || hits+="gcp-staging-vm/variables.tf: the validation is not 600..28800 whole seconds [session bounds]"$'\n'
  for t in '--duration' 'SESSION_MAX_SECONDS' '600' '28800' 'terminationTimestamp' 'CLOUDFLARE_API_TOKEN' 'module.vm.google_compute_instance.vm'; do
    grep -qF -- "$t" "$REPO/docs/foundation/STAGING.md" || hits+="docs/foundation/STAGING.md: does not document $t [session interface]"$'\n'
  done
  # No length change through a normal saved plan (A2 fallback): staging.sh plans normally (neither -refresh-only nor
  # -destroy) only in cmd_provision (first allocation, or the observed length on an existing VM, which the module ignores)
  # and applies only in cmd_provision, session_refresh (its own refresh-only plan) and cmd_retire; the saved session
  # plan and its duration-only check are gone.
  fns() { # PLAN|APPLY: the staging.sh functions with a non-comment tf plan (normal) or tf apply line
    awk -v kind="$1" '/^[a-z_]+\(\) \{/ {fn = $1; sub(/\(\).*/, "", fn)} /^[[:space:]]*#/ {next}
      kind == "PLAN" && /(^|[^a-z_])tf plan / && !/ -refresh-only / && !/ -destroy / {print fn}
      kind == "APPLY" && /(^|[^a-z_])tf apply / {print fn}' "$REPO/infra/staging/staging.sh" | sort -u | tr '\n' ' '
  }
  v="$(fns PLAN)"; [[ "$v" == "cmd_provision " ]] || hits+="infra/staging/staging.sh: normal terraform plans in [$v], want only cmd_provision [no length change through a normal plan]"$'\n'
  v="$(fns APPLY)"; [[ "$v" == "cmd_provision cmd_retire session_refresh " ]] ||
    hits+="infra/staging/staging.sh: terraform applies in [$v], want cmd_provision cmd_retire session_refresh [no length change through a normal plan]"$'\n'
  grep -q 'tf plan -refresh-only -input=false -detailed-exitcode -var-file=/private/terraform.tfvars -var "session_max_seconds=$want"' "$REPO/infra/staging/staging.sh" ||
    hits+="infra/staging/staging.sh: session_refresh does not save a refresh-only plan with the requested length [no length change through a normal plan]"$'\n'
  v="$(grep -nE 'duration_scope_check|plan_scope_rows|-session\.tfplan' "$REPO/infra/staging/staging.sh" "$REPO/infra/staging/lib.sh" || true)"
  [[ -z "$v" ]] || hits+="${v//$'\n'/ | } [the saved session plan path remains]"$'\n'
  timers=()
  while IFS= read -r f; do
    case "$f" in infra/staging/check.sh|*.terraform.lock.hcl|infra/staging/tests/fixtures/*) continue ;; *.sh|*.tf|*.hcl|*.example) timers+=("$f") ;; esac
  done < <(cd "$REPO" && git ls-files -- infra/staging infra/terraform)
  for f in "${timers[@]}"; do
    hits+="$(grep -nvE '^[[:space:]]*(#|//)' "$REPO/$f" | grep -iE 'google_cloud_scheduler|cloudscheduler|gcloud[^|;&]*[[:space:]]scheduler[[:space:]]|crontab|/etc/cron|cron\.(d|daily|hourly)|systemd-run[^|;&]*--on-|OnCalendar=|(^|[;&|(:[:space:]])at[[:space:]]+(-f|-t|now|midnight|noon|teatime)([[:space:]]|$)' |
      sed "s|^|$f:|; s|\$| [scheduler, cron or at job]|" || true)"$'\n'
  done
  # G072 raw-log and lease routing (Plan §10.7, §10.9). Adapter rules over ci.sh and ci-lib.sh with physical lines
  # joined at a trailing backslash and comment lines dropped: (a) every staging.sh subprocess writes stdout and stderr
  # only to $RAW/<step>.log (the local `lease` passthrough prints to the operator); (b) it runs under STAGING_LEASE_HELD
  # unless its subcommand takes no lease; (c) nothing uploads from raw/; (d) nothing reads raw/ to the job's
  # stdout/stderr/fd 3 (a $(...) capture or a redirection to a file is not output); (e) Cloud Storage only through
  # meta-lib's meta_call. Every hit line: FILE:LINE [rule].
  g72_route() { # FILE...
    awk '/^[[:space:]]*#/ {next}
      { l = $0; if (sub(/\\[[:space:]]*$/, "", l)) { buf = buf l " "; if (!st) st = FNR; next }
        printf "%s:%d:%s%s\n", FILENAME, (st ? st : FNR), buf, l; buf = ""; st = 0 }' "$@" |
    awk -v nolease='^(build|push|status|receipt|lease)$' '
      { i = index($0, ":"); f = substr($0, 1, i - 1); r = substr($0, i + 1); j = index(r, ":")
        loc = f ":" substr(r, 1, j - 1); t = substr(r, j + 1) }
      t ~ /bash[[:space:]]+"?\$[{]?STAGING_SH[}]?"?([[:space:]]|$)/ || t ~ /bash[[:space:]]+[^[:space:]|;&]*staging\.sh"?([[:space:]]|$)/ {
        s = t; sub(/.*(STAGING_SH[}]?"?|staging\.sh"?)[[:space:]]+/, "", s); split(s, w, /[[:space:]]+/)
        if (w[1] != "lease" && t !~ />>?[[:space:]]*"?\$[{]?RAW[}]?\/[^[:space:]]+\.log"?[[:space:]]+2>&1/) print loc " [staging.sh output not only to $RAW/<step>.log]"
        if (t !~ /STAGING_LEASE_HELD=/ && w[1] !~ nolease) print loc " [staging.sh " w[1] " without STAGING_LEASE_HELD]"
      }
      t ~ /(meta_put|put_new|gcloud[[:space:]]+storage[[:space:]]+cp|curl)[^|;&]*\$[{]?RAW[}]?/ { print loc " [upload from raw/]" }
      t ~ /(^|[^A-Za-z_])(cat|tail|head|tee|less|more|sed|awk|grep)[[:space:]][^|;&]*\$[{]?RAW[}]?"?\// && t !~ /\$\(/ &&
        (t ~ />&[123]/ || t !~ /(^|[^0-9&>])>>?[[:space:]]*[^&>[:space:]]/) { print loc " [raw/ read to the job output]" }
      t ~ /(^|[^A-Za-z_])gcloud[[:space:]]+storage[[:space:]]/ && t !~ /^[[:space:]]*(show|forbid|printf)[[:space:]]/ &&
        t !~ /META_OUT="\$\(gcloud storage "\$@"/ { print loc " [Cloud Storage outside meta_call]" }'
  }
  # (f) the logs/ upload is the publish_filter output of its own function (Plan §10.7); at least one exists.
  g72_logs() { # FILE...
    awk '/^[a-z_]+\(\) *\{/ {fn = $1; sub(/\(\).*/, "", fn); pf = ""}
      /^[[:space:]]*#/ {next}
      match($0, /publish_filter[[:space:]]+[^[:space:]]+[[:space:]]+[^[:space:];|&]+/) { split(substr($0, RSTART, RLENGTH), a, /[[:space:]]+/); pf = a[3] }
      match($0, /(meta_put|put_new)[[:space:]]+[^[:space:]]+[[:space:]]+"?logs\//) { split(substr($0, RSTART, RLENGTH), b, /[[:space:]]+/); n++
        if (pf == "" || b[2] != pf) printf "%s:%d [logs/ upload of %s is not the publish_filter output of %s]\n", FILENAME, FNR, b[2], fn }
      END { if (!n) print "adapter [no logs/ publication]" }' "$@"
  }
  # (g) staging.sh takes the lease for exactly the six covered mutators, before cmd_<command>; (h) meta-lib's Cloud
  # Storage verbs are describe/cat/cp/rm/ls only (no create verb: lib.sh forbid stays the G070 baseline's, (i));
  # (j) the job refuses an inherited STAGING_LEASE_HELD and the launcher unsets it.
  # The G070 baseline lib.sh is a byte-exact historical copy vendored as a fixture (no Git history needed), pinned by its
  # content SHA-256; a missing, non-regular or altered copy yields no baseline, so the forbid rule fails closed.
  BASELINE_LIB="$REPO/infra/staging/tests/fixtures/baseline/lib.sh"
  BASELINE_LIB_SHA=9d11fe404de15c362137b7457caa779debf53196a40bbd5b968082f6983cb06b
  baseline_forbid() {
    [[ -f $BASELINE_LIB && ! -L $BASELINE_LIB && "$(sha256sum <"$BASELINE_LIB" | cut -d' ' -f1)" == "$BASELINE_LIB_SHA" ]] || return 0
    awk '/^forbid\(\) \{/,/^}/' "$BASELINE_LIB"
  }
  g72_lease() { # STAGING_SH META_LIB LIB_SH CI_LIB LAUNCHER
    local s v base
    [[ "$(grep -cxF 'LEASE_COMMANDS="provision deploy start stop reset retire"' "$1")" == 1 ]] || echo "$1 [LEASE_COMMANDS is not exactly the six covered mutators]"
    awk '/^main\(\) \{/ {m = 1} m && /^}/ {m = 0} m && /lease_guard "\$COMMAND"/ && /\$LEASE_COMMANDS/ && !g {g = FNR}
      m && /"cmd_\$COMMAND"/ {c = FNR} END {if (!g || !c || g > c) print FILENAME " [main does not take the lease for LEASE_COMMANDS before cmd_<command>]"}' "$1"
    while IFS= read -r v; do
      case "$v" in "buckets describe" | "objects describe" | cat | cp | rm | ls) ;; *) echo "$2 [gcloud storage $v through meta_call]" ;; esac
    done < <(grep -vE '^[[:space:]]*#' "$2" | grep -oE 'meta_call [a-z]+( describe)?' | sed 's/^meta_call //' | sort -u)
    base="$(baseline_forbid)"
    [[ -n "$base" && "$base" == "$(awk '/^forbid\(\) \{/,/^}/' "$3")" ]] || echo "$3 [forbid differs from the G070 baseline (or the baseline fixture is missing or altered)]"
    grep -qE '^[[:space:]]*\[\[ -z \$\{STAGING_LEASE_HELD:-\} \]\] \|\| die 2 ' "$4" || echo "$4 [the job does not refuse an inherited STAGING_LEASE_HELD]"
    grep -qE '^[[:space:]]*unset[[:space:]][^#]*STAGING_LEASE_HELD' "$5" || echo "$5 [the launcher does not unset STAGING_LEASE_HELD]"
  }
  g72_all() { # STAGING_SH META_LIB LIB_SH CI_SH CI_LIB LAUNCHER
    g72_route "$4" "$5" "$2"; g72_logs "$4" "$5"; g72_lease "$1" "$2" "$3" "$5" "$6"
  }
  g72_hits="$(cd "$REPO" && g72_all infra/staging/staging.sh infra/staging/meta-lib.sh infra/staging/lib.sh infra/staging/ci.sh \
    infra/staging/ci-lib.sh infra/ci/staging-action.sh)"
  [[ -z "$g72_hits" ]] || hits+="$g72_hits"$'\n'
  # Mutation controls: each rule flags a deliberately broken copy (FAIL when it does not).
  gm="$S/g072-grep"; mkdir -p "$gm"
  g72_mut() { # NAME FILE(of the six) SED RULE_TEXT
    local name=$1 file=$2 prog=$3 want=$4 k a=() out
    rm -rf "${gm:?}/$name"; mkdir -p "$gm/$name"
    for k in infra/staging/staging.sh infra/staging/meta-lib.sh infra/staging/lib.sh infra/staging/ci.sh infra/staging/ci-lib.sh infra/ci/staging-action.sh; do
      cp "$REPO/$k" "$gm/$name/${k//\//_}"; a+=("$gm/$name/${k//\//_}")
    done
    sed -i -e "$prog" "$gm/$name/${file//\//_}"
    if cmp -s "$REPO/$file" "$gm/$name/${file//\//_}"; then hits+="g072 mutation $name: does not apply [STAGING-GREP mutation control]"$'\n'; return; fi
    out="$(g72_all "${a[@]}")"
    grep -qF -- "$want" <<<"$out" || hits+="g072 mutation $name: not flagged ($want) [STAGING-GREP mutation control]"$'\n'
  }
  g72_mut raw-to-stdout infra/staging/ci-lib.sh 's|--config "\$CI_CONF" >>"\$RAW/\$step.log" 2>&1 </dev/null &|--config "$CI_CONF" </dev/null \&|' "[staging.sh output not only to"
  g72_mut no-held infra/staging/ci-lib.sh 's|^    STAGING_LEASE_HELD="\$LEASE_HOLDER@\$LEASE_GEN" bash "\$STAGING_SH"|    bash "$STAGING_SH"|' "without STAGING_LEASE_HELD]"
  g72_mut upload-raw infra/staging/ci-lib.sh 's|put_new "\$out" "logs/|put_new "$RAW/deploy.log" "logs/|' "[upload from raw/]"
  g72_mut raw-cat infra/staging/ci-lib.sh 's|^publish_logs() {|publish_logs() {\n  cat "$RAW/deploy.log" >\&3|' "[raw/ read to the job output]"
  g72_mut logs-unfiltered infra/staging/ci-lib.sh 's|put_new "\$out" "logs/|put_new "$all" "logs/|' "is not the publish_filter output"
  g72_mut storage-direct infra/staging/ci-lib.sh 's|^publish_logs() {|publish_logs() {\n  gcloud storage cp x gs://b/y|' "[Cloud Storage outside meta_call]"
  g72_mut lease-commands infra/staging/staging.sh 's|^LEASE_COMMANDS="provision deploy start stop reset retire"|LEASE_COMMANDS="provision deploy start stop reset"|' "[LEASE_COMMANDS is not exactly"
  g72_mut lease-after-cmd infra/staging/staging.sh '/lease_guard "\$COMMAND"/d' "[main does not take the lease"
  g72_mut storage-verb infra/staging/meta-lib.sh 's|meta_call cp "\$1"|meta_call objects update "$1"|' "[gcloud storage objects through meta_call]"
  g72_mut forbid-changed infra/staging/lib.sh '/^forbid() {/a\  return 0' "[forbid differs"
  g72_mut held-inherited infra/staging/ci-lib.sh '/\[\[ -z \${STAGING_LEASE_HELD:-} \]\] || die 2 /d' "[the job does not refuse an inherited"
  g72_mut launcher-keeps infra/ci/staging-action.sh 's/ STAGING_LEASE_HELD / /' "[the launcher does not unset"
  # Baseline fixture negatives: the real six files against a missing and a tampered baseline copy (one appended comment
  # line, forbid() itself unchanged, so only the content pin catches it); each must be flagged.
  for g72_neg in baseline-missing baseline-tampered; do
    rm -rf "${gm:?}/$g72_neg"; mkdir -p "$gm/$g72_neg"
    if [[ $g72_neg == baseline-tampered ]]; then cp "$BASELINE_LIB" "$gm/$g72_neg/lib.sh" 2>/dev/null || :; printf '#\n' >>"$gm/$g72_neg/lib.sh"; fi
    g72_negout="$(cd "$REPO" && BASELINE_LIB="$gm/$g72_neg/lib.sh" g72_all infra/staging/staging.sh infra/staging/meta-lib.sh \
      infra/staging/lib.sh infra/staging/ci.sh infra/staging/ci-lib.sh infra/ci/staging-action.sh)"
    grep -qF -- "[forbid differs from the G070 baseline (or the baseline fixture is missing or altered)]" <<<"$g72_negout" ||
      hits+="g072 baseline negative $g72_neg: not flagged [STAGING-GREP mutation control]"$'\n'
  done
  hits="$(grep . <<<"$hits" || true)"
  SG_INPUTS="$(inputs_sha infra/staging infra/compose/staging.yml infra/terraform docs/foundation/STAGING.md)"
  c="grep -nE <guards> over ${#files[@]} infra/staging scripts (check.sh and comment lines excluded); session default/bounds pins; scheduler/cron/at grep over ${#timers[@]} infra/staging and infra/terraform files (tests/fixtures excluded); G072 routing rules over ci.sh, ci-lib.sh, meta-lib.sh, staging.sh, lib.sh and the launcher, with 12 mutation controls"
  if (( ${#files[@]} == 0 )); then ledger STAGING-GREP FAIL - "$SG_INPUTS" "$c" "no infra/staging scripts"
  elif [[ -z "$hits" ]]; then ledger STAGING-GREP PASS 0 "$SG_INPUTS" "$c" "every docker inspect formatted without .Config/.Config.Env; no compose config outside dry-run lines; no config.v2.json or /proc/*/environ; no identity/browser/secret-scan overlay; session default $SESSION_PIN in lib.sh, both variable defaults, tfvars and staging.conf examples, STAGING.md and the Terraform README; bounds 600..28800 in lib.sh and the module validation; STAGING.md documents the interface; no scheduler, cron or at job; staging.sh plans normally only in cmd_provision and applies only in cmd_provision, session_refresh and cmd_retire (no saved session plan); G072: staging.sh subprocesses write only to \$RAW/<step>.log under STAGING_LEASE_HELD (lease passthrough and lease-free subcommands excepted), no raw/ upload or raw/ read to the job output, Cloud Storage only through meta_call, logs/ is the publish_filter output, the lease taken for exactly the six mutators before cmd_<command>, describe/cat/cp/rm/ls only, forbid as in G070, inherited STAGING_LEASE_HELD refused and unset by the launcher; 12 mutation controls flagged"
  else printf '%s\n' "$hits" >"$REPORT/staging-grep-hits.txt"; ledger STAGING-GREP FAIL 1 "$SG_INPUTS" "$c" "$(grep -c . <<<"$hits") hits (staging-grep-hits.txt)"; fi
fi

# ---- COMPOSE-STATIC: the overlay invariants of Plan §3.1 carried here while G069 owns run.sh (Codex fallback) ----------
if selected COMPOSE-STATIC; then
  ok=1 why="" rc=0; fail() { ok=0; why+="$1; "; }
  CS_INPUTS="$(inputs_sha infra/compose infra/postgres infra/minio infra/edge infra/staging .env.example)"
  A="$S/archive"; copy_tracked "$A" infra/compose infra/postgres infra/minio infra/edge infra/staging
  env_f="$S/static.env" cfg="$S/compose-config.json"
  c="cd <archive of infra/{compose,postgres,minio,edge,staging}> && env -i docker compose --env-file <scratch>/static.env -f infra/compose/local.yml -f infra/compose/migrate.yml -f infra/compose/staging.yml --profile migrate config --format json ; yq assertions"
  if [[ ! -f "$A/infra/compose/staging.yml" ]]; then fail "infra/compose/staging.yml missing"; rc=1
  else
    # The static environment: .env.example with every empty value generated (as B-COMPOSE's make_env_file), then the
    # staging contract (Plan §2.8, §3.1) with placeholder digests and generated tokens.
    declare -A over=([COMPOSE_PROJECT_NAME]="g070-static-${RUN_ID//[^a-z0-9]/}" [PLATFORM_BASE_DOMAIN]="$BASE_DOMAIN"
      [CUSTOM_DOMAIN_CNAME_TARGET]="custom.$BASE_DOMAIN" [PLATFORM_AUTH_HOST]="auth.$BASE_DOMAIN"
      [EDGE_BASE_DOMAIN]="$BASE_DOMAIN" [ACME_EMAIL]="ops@example.net"
      [EDGE_ACME_CA]="https://acme-v02.api.letsencrypt.org/directory" [CF_API_TOKEN]="$(random_hex 24)"
      [IDENTITY_WORKOS_API_KEY]="$(random_hex 24)" [DEFAULT_TENANT_SLUG]="g070-static-tenant" [PLATFORM_ROOT_PAGE]="landing"
      [MCP_CONNECT_ISSUER]="https://example-env.authkit.app" [MCP_ALLOWED_ORIGINS]="https://client.example")
    mapfile -t image_vars < <(grep -oE '\$\{IMAGE_[A-Z0-9_]+' "$A/infra/compose/staging.yml" | sed 's/^\${//' | sort -u)
    declare -A digests=()
    for v in "${image_vars[@]}"; do
      digests[$v]="registry.example.invalid/$(tr 'A-Z_' 'a-z-' <<<"${v#IMAGE_}")@sha256:$(random_hex 32)"; over[$v]="${digests[$v]}"
    done
    while IFS= read -r line || [[ -n "$line" ]]; do
      if [[ "$line" =~ ^([A-Z0-9_]+)= ]]; then
        k="${BASH_REMATCH[1]}"
        [[ -z "${over[$k]+x}" ]] || continue
        if [[ "$line" == "$k=" && "$k" != OPENROUTER_API_KEY ]]; then printf '%s=%s\n' "$k" "$(random_hex 24)"; continue; fi
      fi
      printf '%s\n' "$line"
    done <"$REPO/.env.example" >"$env_f"
    for v in "${!over[@]}"; do printf '%s=%s\n' "$v" "${over[$v]}" >>"$env_f"; done
    chmod 600 "$env_f"
    while IFS= read -r line; do
      [[ "$line" =~ ^[A-Z0-9_]+=(.+)$ ]] || continue
      val="${BASH_REMATCH[1]}"
      if [[ ${#val} -ge 32 && "$val" =~ ^[0-9a-f]+$ ]]; then SCRUB+=("$val"); fi
    done <"$env_f"
    # Hermetic interpolation: the operator's shell variables would otherwise override the env file.
    cenv=(PATH="$PATH" HOME="${HOME:-/nonexistent}")
    for v in DOCKER_HOST DOCKER_CONTEXT DOCKER_CONFIG; do [[ -z "${!v:-}" ]] || cenv+=("$v=${!v}"); done
    (cd "$A" && env -i "${cenv[@]}" docker compose --env-file "$env_f" -f infra/compose/local.yml -f infra/compose/migrate.yml \
      -f infra/compose/staging.yml --profile migrate config --format json) >"$cfg" 2>"$S/compose.err" || rc=$?
    scrub_copy "$S/compose.err" "$REPORT/compose-static.stderr.txt"
    if [[ $rc != 0 ]]; then fail "compose config exit $rc from the archive (A5 or contract finding; compose-static.stderr.txt)"
    else
      has_line IMAGE_EDGE "$(printf '%s\n' "${image_vars[@]}")" || fail "staging.yml does not use IMAGE_EDGE"
      bad_img="$(grep -nE '^[[:space:]]*image:' "$A/infra/compose/staging.yml" | grep -vE '^[0-9]+:[[:space:]]*image:[[:space:]]*"?\$\{IMAGE_[A-Z0-9_]+(:\?[^}]*)?\}"?[[:space:]]*$' || true)"
      [[ -z "$bad_img" ]] || fail "staging.yml image lines other than \${IMAGE_*} (registry literal?): ${bad_img//$'\n'/ | }"
      grep -vE '^[[:space:]]*#' "$A/infra/compose/staging.yml" | grep -qE 'identity\.yml|browser\.yml|secret-scan\.yml|extends:' \
        && fail "staging.yml references a test overlay or extends another file (comment lines excluded)"
      svcs="$(yq_json '.services | keys | .[]' <"$cfg")"
      for s in edge core web intelligence postgres redis minio minio-init migrate; do has_line "$s" "$svcs" || fail "service $s missing"; done
      grep -qxE 'emulator|proxy|identity-setup|core-replica|web-browser|secret-scan|core-check|web-check|intelligence-check' <<<"$svcs" && fail "a test or check service is composed"
      while IFS=$'\t' read -r s n; do
        [[ "$s" == edge || "$n" == 0 ]] || fail "$s publishes $n port(s)"
      done < <(yq_json '.services | to_entries | .[] | [.key, ((.value.ports // []) | length)]' <"$cfg")
      eports="$(yq_json '.services.edge.ports // [] | .[] | [(.published // ""), .target, (.protocol // "tcp")]' <"$cfg" | LC_ALL=C sort)"
      [[ "$eports" == $'443\t443\ttcp\n80\t80\ttcp' ]] || fail "edge must publish exactly 80/tcp and 443/tcp (found ${eports//$'\n'/, })"
      built="$(yq_json '.services | to_entries | .[] | select(.value | has("build")) | .key' <"$cfg")"
      [[ -z "$built" ]] || fail "services with a build key: ${built//$'\n'/ }"
      imgs="$(yq_json '.services | to_entries | .[] | [.key, (.value.image // "")]' <"$cfg")"
      for s in core web intelligence minio minio-init migrate edge; do
        img="$(awk -F'\t' -v s="$s" '$1 == s {print $2}' <<<"$imgs")"
        [[ "$img" =~ @sha256:[0-9a-f]{64}$ ]] || { fail "$s image is not digest-pinned"; continue; }
        printf '%s\n' "${digests[@]}" | grep -qxF -- "$img" || fail "$s image does not come from an IMAGE_* variable"
      done
      nets="$(yq_json '.services | to_entries | .[] | [.key, ((.value.networks // {}) | keys | sort | join(" "))]' <"$cfg")"
      for want in $'edge\tedge' $'web\tedge' $'core\tdata edge inference' $'intelligence\tinference' $'migrate\tdata'; do
        has_line "$want" "$nets" || fail "networks of ${want%%$'\t'*} must be exactly ${want#*$'\t'}"
      done
      yq_json '.services | to_entries | .[] | .key as $s | (.value.environment // {}) | to_entries | .[] | [$s, .key, .value]' <"$cfg" >"$S/service-env.tsv"
      tk="$(awk -F'\t' '$2 ~ /^(IDENTITY_TEST_|CUSTOM_DOMAIN_DNS_FIXTURE_FILE|DOMAIN_TLS_FIXTURE_FILE|MCP_CONNECT_JWKS_FILE|PLATFORM_HOST_PORT|IDENTITY_WORKOS_API_BASE_URL)/ {print $1 ":" $2}' "$S/service-env.tsv")"
      [[ -z "$tk" ]] || fail "test-only variables composed: ${tk//$'\n'/ }"
      ae="$(awk -F'\t' '$2 == "APP_ENV" && $3 == "test" {print $1}' "$S/service-env.tsv")"
      [[ -z "$ae" ]] || fail "APP_ENV=test in ${ae//$'\n'/ }"
      for s in web core; do
        [[ "$(awk -F'\t' -v s="$s" '$1 == s && $2 == "IDENTITY_COOKIE_PROFILE" {print $3}' "$S/service-env.tsv")" == https ]] || fail "$s IDENTITY_COOKIE_PROFILE is not https"
      done
      for k in EDGE_BASE_DOMAIN ACME_EMAIL EDGE_ACME_CA CF_API_TOKEN; do
        awk -F'\t' -v k="$k" '$1 == "edge" && $2 == k {f = 1} END {exit !f}' "$S/service-env.tsv" || fail "edge lacks $k"
      done
      # G071 S11(iv): the optional root host settings reach core, as given, and no other service; so do the MCP settings.
      for k in DEFAULT_TENANT_SLUG PLATFORM_ROOT_PAGE MCP_CONNECT_ISSUER MCP_ALLOWED_ORIGINS; do
        got="$(awk -F'\t' -v k="$k" '$2 == k {print $1 "=" $3}' "$S/service-env.tsv" | tr '\n' ' ')"
        [[ "$got" == "core=${over[$k]} " ]] || fail "$k must reach core only, as ${over[$k]} (found: ${got:-none})"
      done
      declare -A allowed=(
        [POSTGRES_SUPERUSER_PASSWORD]="postgres" [DB_MIGRATOR_PASSWORD]="postgres migrate"
        [DB_RUNTIME_PASSWORD]="postgres core" [DB_RETENTION_PASSWORD]="postgres"
        [REDIS_PASSWORD]="redis core" [MINIO_ROOT_PASSWORD]="minio minio-init"
        [OBJECT_STORE_SECRET_KEY]="minio-init core" [INTELLIGENCE_SERVICE_TOKEN]="core intelligence"
        [CF_API_TOKEN]="edge" [IDENTITY_WORKOS_API_KEY]="core")
      for sk in "${!allowed[@]}"; do
        sv="$(sed -n "s/^$sk=//p" "$env_f" | tail -1)"; [[ -n "$sv" ]] || { fail "no generated value for $sk"; continue; }
        while IFS=$'\t' read -r s k v; do
          [[ "$v" != *"$sv"* || " ${allowed[$sk]} " == *" $s "* ]] || fail "$sk reaches service $s ($k)"
        done <"$S/service-env.tsv"
      done
      awk -F'\t' '$1 == "intelligence" {print $2}' "$S/service-env.tsv" | grep -qE '^(DB_|REDIS_|OBJECT_STORE_|MINIO_|POSTGRES_)' && fail "intelligence holds data-store settings"
    fi
    rm -f "$cfg" "$S/service-env.tsv" "$env_f" # the generated environment's values
  fi
  [[ $ok == 1 ]] && ledger COMPOSE-STATIC PASS 0 "$CS_INPUTS" "$c" "archive-only config valid (A5); only edge publishes 80/tcp+443/tcp; no build key; core/web/intelligence/minio/minio-init/migrate/edge digest-pinned from \${IMAGE_*} (${#image_vars[@]} variables), staging.yml image lines are variables only; networks edge=[edge] web=[edge] core=[data edge inference] intelligence=[inference] migrate=[data]; no test/check service or test variable; APP_ENV not test; https cookie profile on web and core; edge has its four inputs; DEFAULT_TENANT_SLUG, PLATFORM_ROOT_PAGE, MCP_CONNECT_ISSUER and MCP_ALLOWED_ORIGINS reach core only, as given; every secret only in its owning services (+CF_API_TOKEN edge, +IDENTITY_WORKOS_API_KEY core); intelligence holds no data-store settings" \
    || ledger COMPOSE-STATIC FAIL "$rc" "$CS_INPUTS" "$c" "${why% ; }"
fi

# ---- Source-review corrections: start order, per-name certificates, token format, metadata guard ----------------
REMOTE_INPUTS="$(inputs_sha infra/staging/remote)"
if selected START-ORDER; then
  ok=1 why="" rc=0; fail() { ok=0; why+="$1; "; }
  make_shims; so="$S/start-order"; mkdir -p "$so" "$REPORT/start-order"
  so_env="${ENV_NAME:-staging}"
  c="PATH=<shims>:\$PATH bash infra/staging/remote/start.sh --env $so_env --dry-run ; PATH=<shims>:\$PATH bash -c 'source infra/staging/remote/start.sh; compose_env() { <fails>; }; main --env $so_env --dry-run' ; bash -c 'source infra/staging/remote/stop.sh; <VM seams, compose_env and docker replaced>; main --env $so_env' per stop case ; bash -c 'source infra/staging/remote/deploy.sh; <VM seams, deploy steps, readiness and docker replaced>; main --env $so_env --incoming DIR' per deploy case"
  # (a) The dry-run transcript: every label stop of app and edge precedes the first secret fetch.
  rca=0
  (cd "$REPO" && env PATH="$S/shim:$PATH" HOME="$S/home" bash infra/staging/remote/start.sh --env "$so_env" --dry-run) \
    >"$so/dry-run.txt" 2>&1 || rca=$?
  cp "$so/dry-run.txt" "$REPORT/start-order/dry-run.txt"
  [[ $rca == 0 ]] || fail "dry-run: start.sh exit $rca"
  first_fetch="$(grep -n -m1 -F '(fetch secrets/' "$so/dry-run.txt" | cut -d: -f1 || true)"
  [[ -n "$first_fetch" ]] || fail "dry-run: no secret fetch line"
  last_stop=0
  for svc in edge web intelligence core; do
    n="$(grep -n -m1 -F "docker stop --time 60 <each listed $svc container>" "$so/dry-run.txt" | cut -d: -f1 || true)"
    if [[ -z "$n" ]]; then fail "dry-run: no label stop of $svc"; continue; fi
    (( n < last_stop )) || last_stop=$n
    [[ -z "$first_fetch" ]] || (( n < first_fetch )) || fail "dry-run: the label stop of $svc (line $n) follows the first secret fetch (line $first_fetch)"
  done
  grep -q $'^result\tREADY' "$so/dry-run.txt" || fail "dry-run: no READY result"
  # (b) Failure ordering: the real start.sh, only compose_env replaced by a failing fetch.
  rcb=0
  # shellcheck disable=SC2016 # expanded by the inner bash
  (cd "$REPO" && env PATH="$S/shim:$PATH" HOME="$S/home" bash -c 'source infra/staging/remote/start.sh
    compose_env() { printf "seam\tcompose_env\tfailed\n"; return 2; }
    main --env "$1" --dry-run' _ "$so_env") >"$so/fetch-fails.txt" 2>&1 || rcb=$?
  cp "$so/fetch-fails.txt" "$REPORT/start-order/fetch-fails.txt"
  [[ $rcb == 3 ]] || fail "failed fetch: exit $rcb, want 3 (HOLD)"
  seam="$(grep -n -m1 $'^seam\tcompose_env\tfailed$' "$so/fetch-fails.txt" | cut -d: -f1 || true)"
  if [[ -z "$seam" ]]; then fail "failed fetch: compose_env was not reached"
  else
    for svc in edge web intelligence core; do
      n="$(grep -n -m1 -F "docker stop --time 60 <each listed $svc container>" "$so/fetch-fails.txt" | cut -d: -f1 || true)"
      [[ -n "$n" ]] && (( n < seam )) || fail "failed fetch: the label stop of $svc did not run before compose_env"
    done
    hl="$(grep -n -m1 $'\thold\t' "$so/fetch-fails.txt" | cut -d: -f1 || true)"
    [[ -n "$hl" ]] && (( hl > seam )) || fail "failed fetch: no HOLD step after compose_env"
  fi
  grep -q $'^result\tHOLD' "$so/fetch-fails.txt" || fail "failed fetch: no HOLD result"
  grep -qE 'up -d --wait (postgres|core)' "$so/fetch-fails.txt" && fail "failed fetch: stores or app and edge were started"
  grep -qF '(fetch secrets/' "$so/fetch-fails.txt" && fail "failed fetch: the real secret fetch ran"
  # (c) stop.sh: a failed Compose stop falls back to the label stop of every service; when that fails too, exit 4 with
  # no STOPPED result and the env file left for the EXIT trap. The real main; the VM seams (remote_init without the
  # trap, load_meta, vm_preflight, use_release), compose_env and docker replaced.
  # shellcheck disable=SC2016 # expanded by the inner bash
  st_script='source infra/staging/remote/stop.sh
    remote_init() { ENV_NAME=$1 SRV=$SOD/srv RUN_DIR=$SOD/run ENV_FILE=$SOD/run/app.env; }
    load_meta() { :; }
    vm_preflight() { :; }
    use_release() { REL=$PWD; }
    compose_env() { [[ $FX_ENV == ok ]] || return 2; printf "RENDERED=1\n" >"$ENV_FILE"; }
    docker() {
      local a
      printf "%s\n" "$*" >>"$SOD/docker.log"
      case $1 in
        compose) [[ " $* " != *" stop --timeout 60 "* ]] || return "$FX_COMPOSE_RC" ;;
        ps) for a in "$@"; do [[ $a != label=com.docker.compose.service=* ]] || echo "cid-${a#label=com.docker.compose.service=}"; done ;;
        stop) return "$FX_LABEL_RC" ;;
      esac
      return 0
    }
    main --env "$1"'
  st_case() { # NAME RENDER(ok|fail) COMPOSE_STOP_RC LABEL_STOP_RC: one stop.sh run; sets st_rc
    st_rc=0
    rm -rf "$so/stop"; mkdir -p "$so/stop/run"
    (cd "$REPO" && env PATH="$S/shim:$PATH" HOME="$S/home" SOD="$so/stop" FX_ENV="$2" FX_COMPOSE_RC="$3" FX_LABEL_RC="$4" \
      bash -c "$st_script" _ "$so_env") >"$so/stop-$1.out" 2>"$so/stop-$1.err" || st_rc=$?
    cp "$so/stop-$1.out" "$REPORT/start-order/stop-$1.out"; cp "$so/stop-$1.err" "$REPORT/start-order/stop-$1.err"
    cp "$so/stop/docker.log" "$REPORT/start-order/stop-$1.docker.log" 2>/dev/null || true
  }
  st_labels() { grep -cE '^stop --time 60 cid-(edge|web|intelligence|core|minio|redis|postgres)$' "$so/stop/docker.log" 2>/dev/null || true; }
  st_case compose-ok ok 0 0
  [[ $st_rc == 0 ]] && grep -q $'^result\tSTOPPED$' "$so/stop-compose-ok.out" && grep -q $'\tstop\tevery service stopped (compose stop)$' "$so/stop-compose-ok.out" ||
    fail "stop, compose stop succeeds: exit $st_rc, want 0, the compose-stop step and STOPPED"
  [[ "$(st_labels)" == 0 ]] || fail "stop, compose stop succeeds: containers were also stopped by label"
  [[ ! -e "$so/stop/run/app.env" ]] || fail "stop, compose stop succeeds: the env file was not removed"
  st_case compose-fails-label-ok ok 1 0
  [[ $st_rc == 0 ]] && grep -q $'^result\tSTOPPED$' "$so/stop-compose-fails-label-ok.out" ||
    fail "stop, compose stop fails: exit $st_rc, want 0 and STOPPED after the label fallback"
  grep -q $'\tstop\tcompose stop failed; stopping by Compose labels$' "$so/stop-compose-fails-label-ok.out" || fail "stop, compose stop fails: no fallback step"
  [[ "$(st_labels)" == 7 ]] || fail "stop, compose stop fails: $(st_labels) of 7 services (postgres included) stopped by label"
  [[ ! -e "$so/stop/run/app.env" ]] || fail "stop, compose stop fails: the env file was not removed after the fallback stop"
  st_case both-fail ok 1 1
  [[ $st_rc == 4 ]] || fail "stop, compose and label stop fail: exit $st_rc, want 4"
  ! grep -q $'^result\tSTOPPED' "$so/stop-both-fail.out" || fail "stop, compose and label stop fail: STOPPED reported"
  grep -qF 'compose stop failed and a container could not be listed or stopped by label; check docker ps' "$so/stop-both-fail.err" || fail "stop, compose and label stop fail: no exit-4 reason"
  [[ -e "$so/stop/run/app.env" ]] || fail "stop, compose and label stop fail: the env file was removed before the failure"
  st_case env-unrendered fail 0 0
  [[ $st_rc == 0 ]] && grep -q $'^result\tSTOPPED$' "$so/stop-env-unrendered.out" && [[ "$(st_labels)" == 7 ]] ||
    fail "stop, env not rendered: exit $st_rc, want 0, STOPPED and 7 label stops"
  ! grep -q '^compose ' "$so/stop/docker.log" 2>/dev/null || fail "stop, env not rendered: Compose was invoked"
  # (d) deploy.sh's seeding rule: a fresh volume is seeded; an existing one with its marker never; one without its
  # marker only while the application-data count is exactly 0 (an interrupted first deploy), else HOLD (rows, a failed
  # count or an unexpected answer). The real main, marker_label, dataset_empty, seed_dataset and hold; the VM seams,
  # the secret, image, store, migration, metadata and app steps, readiness (marker only) and docker replaced.
  # shellcheck disable=SC2016 # expanded by the inner bash
  dp_script='source infra/staging/remote/deploy.sh
    remote_init() { ENV_NAME=$1 SRV=$DPD/srv RUN_DIR=$DPD/run ENV_FILE=$DPD/run/app.env; }
    load_meta() { RCONF[DATABASE_NAME]=app RCONF[CANDIDATE]=0123456789abcdef0123456789abcdef01234567; }
    vm_preflight() { :; }
    compose_env() { printf "RENDERED=1\n" >"$ENV_FILE"; }
    promote() { :; }
    write_secret_versions() { :; }
    use_release() { REL=$PWD; }
    pull_images() { :; }
    stores_up() { :; }
    migrate_current() { :; }
    metadata_denied() { :; }
    app_up() { :; }
    readiness() { marker_label; printf "marker\t%s\n" "$MARKER"; [[ -n $MARKER ]] || return 1; printf "result\tREADY\n"; }
    docker() {
      printf "%s\n" "$*" >>"$DPD/docker.log"
      case $1 in
        volume) [[ $FX_VOLUME == 0 ]] || echo "${ENV_NAME}_postgres-data" ;;
        ps) echo pgcid ;;
        exec)
          case "$*" in
            *platform_label*) [[ $FX_MARKER == 0 && ! -e $DPD/seeded ]] || echo 0123456789abcdef0123456789abcdef ;;
            *query_to_xml*) if [[ $FX_ROWS == fail ]]; then echo "ERROR: fixture" >&2; return 2; fi; echo "$FX_ROWS" ;;
            *) return 97 ;;
          esac ;;
        compose) if [[ " $* " == *" exec -T postgres psql "* ]]; then cat >/dev/null; echo seed >>"$DPD/seeds.log"; : >"$DPD/seeded"; fi ;;
      esac
      return 0
    }
    main --env "$1" --incoming "$DPD/incoming"'
  dp_case() { # NAME VOLUME(0|1) MARKER(0|1) ROWS(number|fail|text): one deploy.sh run; sets dp_rc and dp_seeds
    dp_rc=0
    rm -rf "$so/deploy"; mkdir -p "$so/deploy/run"
    (cd "$REPO" && env PATH="$S/shim:$PATH" HOME="$S/home" DPD="$so/deploy" FX_VOLUME="$2" FX_MARKER="$3" FX_ROWS="$4" \
      bash -c "$dp_script" _ "$so_env") >"$so/deploy-$1.out" 2>"$so/deploy-$1.err" || dp_rc=$?
    cp "$so/deploy-$1.out" "$REPORT/start-order/deploy-$1.out"; cp "$so/deploy-$1.err" "$REPORT/start-order/deploy-$1.err"
    dp_seeds="$(grep -c . "$so/deploy/seeds.log" 2>/dev/null || true)"; dp_seeds=${dp_seeds:-0}
    dp_counts="$(grep -c 'query_to_xml' "$so/deploy/docker.log" 2>/dev/null || true)"; dp_counts=${dp_counts:-0}
  }
  dp_ready() { grep -q $'^result\tREADY$' "$so/deploy-$1.out" && grep -qE $'^marker\t[0-9a-f]{32}$' "$so/deploy-$1.out"; }
  dp_hold() { grep -q $'^result\tHOLD\t'"$2" "$so/deploy-$1.out" && [[ -n "$(ls "$so/deploy/srv/state"/hold-*.tsv 2>/dev/null)" ]]; }
  dp_case fresh-volume 0 0 0
  [[ $dp_rc == 0 && $dp_seeds == 1 && $dp_counts == 0 ]] && dp_ready fresh-volume ||
    fail "deploy, fresh volume: exit $dp_rc, seeds $dp_seeds, counts $dp_counts; want 0, one seed, no count, READY with a marker"
  dp_case existing-with-marker 1 1 0
  [[ $dp_rc == 0 && $dp_seeds == 0 && $dp_counts == 0 ]] && dp_ready existing-with-marker ||
    fail "deploy, existing dataset with its marker: exit $dp_rc, seeds $dp_seeds, counts $dp_counts; want 0, no seed, no count, READY"
  dp_case interrupted-first-deploy 1 0 0
  [[ $dp_rc == 0 && $dp_seeds == 1 && $dp_counts == 1 ]] && dp_ready interrupted-first-deploy ||
    fail "deploy, existing volume without marker and without application data: exit $dp_rc, seeds $dp_seeds, counts $dp_counts; want 0, one count, one seed, READY with the new marker"
  grep -q $'\tdataset\texisting postgres-data without its marker and without application data (interrupted first deploy): retrying the seed$' \
    "$so/deploy-interrupted-first-deploy.out" || fail "deploy, interrupted first deploy: no seed-retry step"
  cl="$(grep -n -m1 'query_to_xml' "$so/deploy/docker.log" | cut -d: -f1 || true)"
  sl="$(grep -n -m1 ' exec -T postgres psql ' "$so/deploy/docker.log" | cut -d: -f1 || true)"
  [[ -n "$cl" && -n "$sl" ]] && (( cl < sl )) || fail "deploy, interrupted first deploy: the seed did not follow the application-data count"
  dp_case rows-present 1 0 3
  [[ $dp_rc == 3 && $dp_seeds == 0 ]] && dp_hold rows-present 'existing dataset without its marker (no reseed' ||
    fail "deploy, existing volume without marker with application data: exit $dp_rc, seeds $dp_seeds; want HOLD (3), no seed"
  dp_case count-fails 1 0 fail
  [[ $dp_rc == 3 && $dp_seeds == 0 ]] && dp_hold count-fails 'existing dataset without its marker whose application data could not be counted' ||
    fail "deploy, application-data count fails: exit $dp_rc, seeds $dp_seeds; want HOLD (3), no seed"
  dp_case count-unexpected 1 0 "0 rows"
  [[ $dp_rc == 3 && $dp_seeds == 0 ]] && dp_hold count-unexpected 'existing dataset without its marker whose application data could not be counted' ||
    fail "deploy, unexpected application-data answer: exit $dp_rc, seeds $dp_seeds; want HOLD (3), no seed"
  if grep -q . "$S/shim.log"; then fail "executed $(cut -d' ' -f1 "$S/shim.log" | sort -u | tr '\n' ' ')"; fi
  [[ $ok == 1 ]] && ledger START-ORDER PASS 0 "$REMOTE_INPUTS" "$c" "dry-run: app/edge label stops (last line $last_stop) before the first secret fetch (line $first_fetch), READY; compose_env failing: the label stops ran first, then HOLD (exit 3); no store, app or edge start; stop.sh: compose stop ok -> STOPPED without label stops; compose stop failing -> the label stop of all 7 services, STOPPED; both failing -> exit 4, no STOPPED, env file not removed first; env not rendered -> label stops, STOPPED; deploy.sh: fresh volume -> one seed; marker present -> no seed; no marker and an application-data count of 0 -> the count, then one seed and the marker (READY); no marker with rows, a failed count or an unexpected answer -> HOLD (3), no seed; no tool executed; transcripts in start-order/" \
    || ledger START-ORDER FAIL "${rcb:-1}" "$REMOTE_INPUTS" "$c" "${why% ; }"
fi

if selected HTTPS-CERT; then
  ok=1 why="" rc=0; fail() { ok=0; why+="$1; "; }
  make_shims; hc="$S/https-cert"; mkdir -p "$hc" "$REPORT/https-cert"
  c="bash -c 'source infra/staging/remote/lib.sh; <curl, sleep and cert_lines replaced by fixtures>; https_ready' per case (no network, no Docker)"
  # Fixture: B=staging.example.net, '@' in a SAN stands for B; the apex and label certificates have distinct serials.
  # shellcheck disable=SC2016 # expanded by the inner bash
  hc_script='source infra/staging/remote/lib.sh
    B=staging.example.net
    RCONF[PLATFORM_BASE_DOMAIN]=$B
    if [[ $FX_CA == rehearsal ]]; then RCONF[EDGE_ACME_CA]=$LE_STAGING_CA; else RCONF[EDGE_ACME_CA]=https://acme-v02.api.letsencrypt.org/directory; fi
    sleep() { :; }
    curl() { local a; for a in "$@"; do [[ $a != https://* || $a == "https://$B/"* || $FX_LABEL_CURL == ok ]] || return 60; done; }
    cert_lines() {
      local k=LABEL serial san v7 iss
      [[ $1 != "$B" ]] || k=APEX
      serial=FX_${k}_SERIAL san=FX_${k}_SAN v7=FX_${k}_7D iss=FX_${k}_ISSUER
      printf "cert\t%s\tserial\t%s\ncert\t%s\tissuer\t%s\ncert\t%s\tsan\t%s\ncert\t%s\tvalid_7d\t%s\n" \
        "$1" "${!serial}" "$1" "${!iss}" "$1" "${!san//@/$B}" "$1" "${!v7}"
    }
    https_ready'
  # NAME|pass or fail|reason (fail)|ca|apex SAN|label SAN|apex 7d|label 7d|label chain|apex issuer|label issuer
  hc_cases=(
    "distinct-serials|pass||production|DNS:@|DNS:*.@|yes|yes|ok|CN=E5|CN=E6"
    "one-certificate-both-names|pass||production|DNS:@,DNS:*.@|DNS:@,DNS:*.@|yes|yes|ok|CN=E5|CN=E5"
    "label-san-without-wildcard|fail|does not list *.staging.example.net|production|DNS:@|DNS:@|yes|yes|ok|CN=E5|CN=E6"
    "label-san-wildcard-not-exact|fail|does not list *.staging.example.net|production|DNS:@|DNS:*.@.evil.example|yes|yes|ok|CN=E5|CN=E6"
    "apex-san-without-base|fail|served for staging.example.net does not list staging.example.net|production|DNS:*.@|DNS:*.@|yes|yes|ok|CN=E5|CN=E6"
    "apex-expires-within-7d|fail|served for staging.example.net expires within 7 days|production|DNS:@|DNS:*.@|no|yes|ok|CN=E5|CN=E6"
    "label-expires-within-7d|fail|expires within 7 days|production|DNS:@|DNS:*.@|yes|no|ok|CN=E5|CN=E6"
    "label-chain-unverified|fail|no verified TLS answer for a random label host|production|DNS:@|DNS:*.@|yes|yes|bad|CN=E5|CN=E6"
    "rehearsal-staging-issuers|pass||rehearsal|DNS:@|DNS:*.@|yes|yes|ok|CN=(STAGING) Ersatz E5|CN=(STAGING) Ersatz E6"
    "rehearsal-issuer-not-staging|fail|is not from the staging directory|rehearsal|DNS:@|DNS:*.@|yes|yes|ok|CN=(STAGING) Ersatz E5|CN=E6"
  )
  for spec in "${hc_cases[@]}"; do
    IFS='|' read -r name want reason ca asan lsan a7 l7 lchain aiss liss <<<"$spec"
    rc=0
    (cd "$REPO" && env PATH="$S/shim:$PATH" HOME="$S/home" FX_CA="$ca" FX_APEX_SERIAL=0A0A FX_LABEL_SERIAL=0B0B \
      FX_APEX_SAN="$asan" FX_LABEL_SAN="$lsan" FX_APEX_7D="$a7" FX_LABEL_7D="$l7" FX_LABEL_CURL="$lchain" \
      FX_APEX_ISSUER="$aiss" FX_LABEL_ISSUER="$liss" bash -c "$hc_script") >"$hc/$name.txt" 2>&1 || rc=$?
    cp "$hc/$name.txt" "$REPORT/https-cert/$name.txt"
    if [[ $want == pass ]]; then [[ $rc == 0 ]] || fail "$name: exit $rc, want PASS (https-cert/$name.txt)"
    else
      [[ $rc != 0 ]] || fail "$name: passed, want FAIL"
      grep -qF -- "$reason" "$hc/$name.txt" || fail "$name: no reason '$reason' (https-cert/$name.txt)"
    fi
  done
  if grep -q . "$S/shim.log"; then fail "executed $(cut -d' ' -f1 "$S/shim.log" | sort -u | tr '\n' ' ')"; fi
  [[ $ok == 1 ]] && ledger HTTPS-CERT PASS 0 "$(inputs_sha infra/staging/remote/lib.sh)" "$c" "${#hc_cases[@]} cases: distinct apex/label serials and one certificate for both names pass; label SAN without exactly *.B, apex SAN without B, either expiry within 7 days, an unverified label chain and a non-(STAGING) rehearsal issuer fail with their reason; outputs in https-cert/" \
    || ledger HTTPS-CERT FAIL 1 "$(inputs_sha infra/staging/remote/lib.sh)" "$c" "${why% ; }"
fi

if selected CF-TOKEN-CANARY; then
  ok=1 why="" rc=0; fail() { ok=0; why+="$1; "; }
  make_shims; cf="$S/cf-token"; mkdir -p "$cf/run" "$cf/srv/state" "$REPORT/cf-token-canary"
  # Two malformed canaries (a space and the wrong length; a valid alphabet but 41 characters), an unsupported-prefix
  # canary and two well-formed tokens (40 characters; cfut_), all generated now and passed in the environment, never
  # on a command line or in the report.
  canary_a="G070CANARY$(random_hex 8)-bad token" canary_b="G070CANARY$(random_hex 15)x"
  good_token="$(head -c 30 /dev/urandom | base64 | tr '+/' '-_')"
  # The current prefixed form: cfut_ plus 48 alphanumerics (53 characters), also generated now.
  good_token_pfx="cfut_$(head -c 90 /dev/urandom | base64 | tr -dc 'A-Za-z0-9' | head -c 48)"
  # An unsupported prefix of the same shape (cfxx_ plus 48 alphanumerics) must be refused: the pinned DNS plugin accepts only cfut_ and cfat_.
  canary_c="cfxx_$(head -c 90 /dev/urandom | base64 | tr -dc 'A-Za-z0-9' | head -c 48)"
  SCRUB+=("$canary_a" "$canary_b" "$canary_c" "$good_token" "$good_token_pfx")
  msg='staging-vm: CF_API_TOKEN does not have the Cloudflare API-token format; refused before any env file or container'
  c="bash -c 'source infra/staging/remote/lib.sh; <sm_token, sm_access replaced; docker and run record and fail>; compose_env deploy|pinned' with CF_API_TOKEN = 2 malformed canaries and 1 unsupported-prefix (cfxx_) canary, then 2 generated well-formed tokens (40-character and cfut_) (environment only)"
  # shellcheck disable=SC2016 # expanded by the inner bash
  cf_script='set -u
    source infra/staging/remote/lib.sh
    ENV_NAME=staging SRV=$CFD/srv RUN_DIR=$CFD/run ENV_FILE=$CFD/run/app.env DRY_RUN=0
    for k in "${REMOTE_ENV_KEYS[@]}"; do RCONF[$k]=g070-value; done
    RCONF[DEFAULT_TENANT_SLUG]="" RCONF[PLATFORM_ROOT_PAGE]="" RCONF[MCP_CONNECT_ISSUER]="" RCONF[MCP_ALLOWED_ORIGINS]=""
    for svc in "${IMAGE_SERVICES[@]}"; do var=${svc^^}; IMAGES[IMAGE_${var//-/_}]=registry.example.invalid/$svc@sha256:0; done
    docker() { printf "docker %s\n" "$1" >>"$CFD/calls.log"; return 97; }
    run() { printf "run %s\n" "$1" >>"$CFD/calls.log"; return 97; }
    sm_token() { SM_TOKEN=$PH; }
    sm_access() {
      SM_VERSION="" SM_VALUE=""
      ! is_optional "$1" || return 1
      SM_VERSION=7
      if [[ $1 == CF_API_TOKEN ]]; then SM_VALUE=$CF_VALUE; else SM_VALUE=g070-dummy-$1; fi
    }
    { for n in "${SECRETS_REQUIRED[@]}"; do printf "%s\t7\n" "$n"; done
      for n in "${SECRETS_OPTIONAL[@]}"; do printf "%s\tnone\n" "$n"; done; } >"$SRV/state/secret-versions.tsv"
    rc=0
    compose_env "$1" || rc=$?
    printf "compose_env\t%s\texit\t%s\n" "$1" "$rc"
    exit "$rc"'
  cf_run() { # MODE CASE VALUE: one compose_env run; outputs in <scratch>/cf-token/MODE-CASE.{out,err}
    local r=0
    rm -f "$cf/calls.log"
    (cd "$REPO" && env PATH="$S/shim:$PATH" HOME="$S/home" CFD="$cf" PH=placeholder CF_VALUE="$3" bash -c "$cf_script" _ "$1") \
      >"$cf/$1-$2.out" 2>"$cf/$1-$2.err" || r=$?
    return "$r"
  }
  for mode in deploy pinned; do
    for cn in a b c; do
      case $cn in a) val=$canary_a ;; b) val=$canary_b ;; c) val=$canary_c ;; esac
      rc=0; cf_run "$mode" "canary-$cn" "$val" || rc=$?
      [[ $rc != 0 ]] || fail "$mode, canary $cn: accepted"
      grep -qxF -- "$msg" "$cf/$mode-canary-$cn.err" || fail "$mode, canary $cn: not refused with the fixed message"
      [[ ! -s "$cf/calls.log" ]] || fail "$mode, canary $cn: $(grep -c . "$cf/calls.log") docker/compose call(s) after the refusal"
      [[ ! -e "$cf/run/app.env" ]] || fail "$mode, canary $cn: the env file was written"
    done
    for gcase in good good-pfx; do
      if [[ $gcase == good ]]; then val=$good_token; else val=$good_token_pfx; fi
      rc=0; cf_run "$mode" "$gcase" "$val" || rc=$?
      [[ $rc == 0 ]] || fail "$mode, well-formed token ($gcase): exit $rc"
      [[ "$(grep -c '^CF_API_TOKEN=' "$cf/run/app.env" 2>/dev/null || true)" == 1 ]] || fail "$mode, well-formed token ($gcase): not rendered"
      [[ ! -s "$cf/calls.log" ]] || fail "$mode, well-formed token ($gcase): docker/compose called"
      rm -f "$cf/run/app.env"
    done
  done
  val=""
  for f in "$cf"/*.out "$cf"/*.err; do scrub_copy "$f" "$REPORT/cf-token-canary/${f##*/}"; done
  for v in "$canary_a" "$canary_b" "$canary_c"; do
    grep -rqF -- "$v" "$cf" && fail "a canary appears in an output or scratch file"
    grep -rqF -- "$v" "$REPORT" && fail "a canary appears in the report directory"
  done
  grep -qF -- "$good_token" "$cf"/*.out "$cf"/*.err && fail "the well-formed token was printed"
  grep -qF -- "$good_token_pfx" "$cf"/*.out "$cf"/*.err && fail "the prefixed well-formed token was printed"
  if grep -q . "$S/shim.log"; then fail "executed $(cut -d' ' -f1 "$S/shim.log" | sort -u | tr '\n' ' ')"; fi
  [[ $ok == 1 ]] && ledger CF-TOKEN-CANARY PASS 0 "$(inputs_sha infra/staging/remote/lib.sh)" "$c" "deploy and pinned: the two malformed canaries and an unsupported-prefix canary (cfxx_) refused with the fixed message, no docker/compose call, no env file; canaries absent from outputs, scratch and report; the generated well-formed 40-character and cfut_ tokens rendered once and never printed; outputs in cf-token-canary/" \
    || ledger CF-TOKEN-CANARY FAIL 1 "$(inputs_sha infra/staging/remote/lib.sh)" "$c" "${why% ; }"
fi

if selected METADATA-PROBE; then
  ok=1 why="" rc=0; fail() { ok=0; why+="$1; "; }
  mp="$S/metadata-probe"; mkdir -p "$mp"
  canary="G070CANARY$(random_hex 12)"
  SCRUB+=("$canary")
  c="bash -c 'source infra/staging/remote/lib.sh; <docker replaced by a fake daemon with a container table>; metadata_denied' for the classification, ownership, inventory and reconciliation cases"
  # The fake daemon: one line per container in \$MPD/containers.tsv (ID, name, run label, env label, image ID, task label,
  # probe-kind label); run writes the --cidfile and the container (and never removes it: no --rm), rm removes it unless
  # STUCK=1, ps filters by the task/probe/env labels, id and name, inspect answers .Image, image inspect answers .Id.
  cat >"$mp/fake-docker.sh" <<'FAKE'
fdb="$MPD/containers.tsv"; touch "$fdb"
docker() {
  printf '%s\n' "$*" >>"$MPD/calls.log"
  local a name="" cidf="" url="" run="" env="" id nm rn ev img tk kd labtask="" labkind="" labenv="" idf="" namef="" quiet=0 prev=""
  case "$1" in
    network) for a in ${NETS:-net1}; do echo "$a"; done ;;
    image) echo "sha256:img-edge" ;;
    inspect) awk -F'\t' -v i="$4" 'index($1, i) == 1 {print $5; exit}' "$fdb" ;;
    rm)
      shift 2
      for id in "$@"; do
        # STUCK=1: no removal works; STUCK_SEQ=N: only the container whose name ends in -N cannot be removed
        nm="$(awk -F'\t' -v i="$id" 'index($1, i) == 1 {print $2; exit}' "$fdb")"
        [[ ${STUCK:-0} == 1 || ( -n ${STUCK_SEQ:-} && $nm == *"-$STUCK_SEQ" ) ]] ||
          { grep -v "^$id" "$fdb" >"$fdb.new" || true; mv "$fdb.new" "$fdb"; }
      done ;;
    ps)
      [[ $2 == -aq ]] && quiet=1
      for a in "$@"; do
        case "$a" in
          label=g070.task=*) labtask=${a#label=g070.task=} ;;
          label=g070.probe=*) labkind=${a#label=g070.probe=} ;;
          label=g070.env=*) labenv=${a#label=g070.env=} ;;
          id=*) idf=${a#id=} ;;
          name=*) namef=${a#name=^}; namef=${namef%\$} ;;
        esac
      done
      while IFS=$'\t' read -r id nm rn ev img tk kd; do
        [[ -n $id ]] || continue
        [[ -z $labtask || $tk == "$labtask" ]] || continue
        [[ -z $labkind || $kd == "$labkind" ]] || continue
        [[ -z $labenv || $ev == "$labenv" ]] || continue
        [[ -z $idf || $id == "$idf"* ]] || continue
        [[ -z $namef || $nm == "$namef" ]] || continue
        if ((quiet)); then echo "${id:0:12}"; else printf '%s\t%s\t%s\n' "${id:0:12}" "$nm" "$rn"; fi
      done <"$fdb" ;;
    run)
      for a in "$@"; do
        case "$prev" in
          --name) name=$a ;;
          --cidfile) cidf=$a ;;
          --label) case "$a" in g070.probe.run=*) run=${a#g070.probe.run=} ;; g070.env=*) env=${a#g070.env=} ;; esac ;;
        esac
        prev=$a
        url=$a
      done
      if [[ ${NOCREATE:-0} != 1 ]]; then
        id="$(printf '%s' "$name" | sha256sum | cut -c1-64)"
        mkdir -p "$(dirname "$cidf")"
        printf '%s' "$id" >"$cidf"
        printf '%s\t%s\t%s\t%s\tsha256:img-edge\tG070\tmetadata-denied\n' "$id" "$name" "$run" "$env" >>"$fdb"
      fi
      # BREAKINV=1: the inventory becomes unwritable (a directory) right after the container exists
      if [[ ${BREAKINV:-0} == 1 ]]; then rm -f "$MPD/srv/state/probe-containers.tsv"; mkdir "$MPD/srv/state/probe-containers.tsv"; fi
      if [[ $url == http://127.0.0.1:1/ ]]; then return "$CTL_RC"; fi
      printf '%s' "$FAKE_OUT"
      return "$FAKE_RC" ;;
  esac
}
FAKE
  # shellcheck disable=SC2016 # expanded by the inner bash
  mp_script='set -u
    source infra/staging/remote/lib.sh
    ENV_NAME=staging DRY_RUN=0 SRV=$MPD/srv
    IMAGES[IMAGE_EDGE]=registry.example.invalid/edge@sha256:0
    source "$MPD/fake-docker.sh"
    rc=0
    if [[ $MODE == cancel ]]; then
      # an interrupted run: the EXIT trap path (cleanup) must reap the owned probe containers by identity
      ENV_FILE=$MPD/none.env RUN_DIR=$MPD/run DOCKER_CFG="" PROBE_RUN=20261027000000-1
      cleanup
      printf "cancel\tcleanup-done\n"
      exit 0
    fi
    metadata_denied || rc=$?
    printf "exit\t%s\n" "$rc"'
  inv="$mp/srv/state/probe-containers.tsv"
  seed_name=g070-md-20260101000000-1-1
  seed_id="$(printf '%s' "$seed_name" | sha256sum | cut -c1-64)"
  seed_row() { # RUN IMAGE: the allocate record of the seeded container
    printf '20261027T000000Z\t%s\t%s\tallocate\tnet1\t%s\t-\t\n' "$1" "$seed_name" "$2" >>"$inv"
  }
  seed_ctr() { # RUN ENV IMAGE [CIDFILE_ID]: the seeded container and its cidfile
    printf '%s\t%s\t%s\t%s\t%s\tG070\tmetadata-denied\n' "$seed_id" "$seed_name" "$1" "$2" "$3" >>"$mp/containers.tsv"
    printf '%s' "${4:-$seed_id}" >"$mp/srv/state/probe-cids/$seed_name"
  }
  seed_owned() { seed_row 20260101000000-1 sha256:img-edge; seed_ctr 20260101000000-1 staging sha256:img-edge; }
  seed_unowned() { seed_ctr 20260101000000-1 staging sha256:img-edge; }
  seed_mismatch_image() { seed_row 20260101000000-1 sha256:img-old; seed_ctr 20260101000000-1 staging sha256:img-edge; }
  seed_mismatch_cid() { seed_row 20260101000000-1 sha256:img-edge; seed_ctr 20260101000000-1 staging sha256:img-edge "$(printf '%064d' 7)"; }
  seed_mismatch_run() { seed_row 20250505000000-9 sha256:img-edge; seed_ctr 20260101000000-1 staging sha256:img-edge; }
  seed_stale_record() { seed_row 20260101000000-1 sha256:img-edge; printf '%s' "$seed_id" >"$mp/srv/state/probe-cids/$seed_name"; }
  seed_inv_dir() { mkdir -p "$inv"; }
  # a stale record whose cidfile cannot be unlinked (a non-empty directory: rm -f fails for any user)
  seed_stale_unlink() { seed_row 20260101000000-1 sha256:img-edge; mkdir -p "$mp/srv/state/probe-cids/$seed_name"; : >"$mp/srv/state/probe-cids/$seed_name/x"; }
  seed_unrelated() { seed_ctr 20260101000000-1 other-env sha256:img-edge; }
  mp_case() { # NAME CTL_RC PROBE_RC PROBE_ANSWER EXPECT(denied|UNPROVEN|REACHABLE|held|cancel); SEED=fn STUCK=1 NOCREATE=1 MODE=cancel optional
    local name="$1" got r=0
    rm -f "$mp/calls.log"
    if [[ "${KEEP:-0}" != 1 ]]; then # KEEP=1: continue on the state the previous case left (a retry)
      rm -rf "$mp/srv" "$mp/containers.tsv"
      mkdir -p "$mp/srv/state/probe-cids"; : >"$mp/containers.tsv"
    fi
    [[ -z "${SEED:-}" ]] || "$SEED"
    (cd "$REPO" && env PATH="$S/shim:$PATH" HOME="$S/home" MPD="$mp" CTL_RC="$2" FAKE_RC="$3" FAKE_OUT="$4" \
      STUCK="${STUCK:-0}" STUCK_SEQ="${STUCK_SEQ:-}" NETS="${NETS:-net1}" NOCREATE="${NOCREATE:-0}" BREAKINV="${BREAKINV:-0}" MODE="${MODE:-probe}" bash -c "$mp_script") \
      >"$mp/$name.out" 2>"$mp/$name.err" || r=$?
    got="$(awk -F'\t' '$1 == "metadata-denied" {print $3}' "$mp/$name.out" | sort -u | paste -sd,)"
    case "$5" in
      denied) [[ $r == 0 && -z "$got" ]] && grep -q $'^exit\t0$' "$mp/$name.out" || fail "$name: expected denied (exit 0, no failure line), got exit $r lines [$got]" ;;
      held) grep -q $'^exit\t1$' "$mp/$name.out" && grep -q $'\tprobe-reap\tHELD ' "$mp/$name.out" &&
        grep -q $'\tmetadata-denied\tunresolved probe containers or records of staging; no new probe container is started' "$mp/$name.out" &&
        [[ "$(grep -c '^run ' "$mp/calls.log" 2>/dev/null || true)" == 0 ]] || fail "$name: expected exit 1, a HELD step, no new probe and zero docker run calls, got exit $r" ;;
      blocked) grep -q $'^exit\t1$' "$mp/$name.out" &&
        grep -q $'\tmetadata-denied\tunresolved probe containers or records of staging; no new probe container is started' "$mp/$name.out" &&
        [[ "$(grep -c '^run ' "$mp/calls.log" 2>/dev/null || true)" == 0 ]] || fail "$name: expected exit 1, an unresolved-reconciliation stop and zero docker run calls, got exit $r" ;;
      cancel) grep -q $'^cancel\tcleanup-done$' "$mp/$name.out" || fail "$name: cleanup did not complete (exit $r)" ;;
      *) [[ "$got" == "$5" ]] && grep -q $'^exit\t1$' "$mp/$name.out" || fail "$name: expected $5 and exit 1, got exit $r lines [$got]" ;;
    esac
    ! grep -qF -- "$canary" "$mp/$name.out" "$mp/$name.err" || fail "$name: the injected answer was printed"
  }
  # status classification (fail closed)
  mp_case denied-status-1 1 1 "" denied
  mp_case denied-timeout-124 1 124 "" denied
  mp_case denied-timeout-143 1 143 "" denied
  mp_case crash-139-empty 1 139 "" UNPROVEN
  mp_case killed-137 1 137 "" UNPROVEN
  mp_case docker-125 1 125 "" UNPROVEN
  mp_case reachable-answer 1 0 "$canary" REACHABLE
  mp_case reachable-body-with-failure-status 1 1 "$canary" REACHABLE
  mp_case control-crash-139 139 1 "" UNPROVEN
  mp_case control-success 0 1 "" UNPROVEN
  # a failed control must stop before any metadata probe of that network
  ! grep -qE '169\.254\.169\.254|fd20:ce|metadata\.google' "$mp/calls.log" || fail "a probe ran although its tool control failed"
  # the production probe never passes wget -T (BusyBox 1.35 segfaults with it, exit 139)
  ! sed -n '/^metadata_denied() {/,/^}/p;/^probe_cmd() {/,/^}/p' "$REPO/infra/staging/remote/lib.sh" | grep -vE '^[[:space:]]*#' | grep -qE 'wget[^|]* -T( |$)' || fail "the probe passes wget -T"
  # ownership and inventory of a normal run: named, labelled, cidfile, no --rm; allocate, created (container ID, image
  # ID) and removed rows for every probe, removal by the recorded ID, an empty container table and the receipt steps
  mp_case inventory-normal 1 1 "" denied
  ! { grep '^run ' "$mp/calls.log" | grep -vqE -- '^run --pull never --init --name g070-md-[0-9]+-[0-9]+-[0-9]+ --cidfile [^ ]+/state/probe-cids/g070-md-[0-9]+-[0-9]+-[0-9]+ --label g070\.task=G070 --label g070\.env=staging --label g070\.probe=metadata-denied --label g070\.probe\.run=[0-9]+-[0-9]+ '; } ||
    fail "a probe container lacks its unique name, cidfile or ownership labels"
  ! grep '^run ' "$mp/calls.log" | grep -qE -- ' --rm( |$)' || fail "a probe container runs with --rm (its ID would not be recorded before removal)"
  # BusyBox `timeout` as a container's PID 1 never fires (no signal without a handler reaches a namespace's init): the
  # probe runs under docker-init, so every probe has --init with the timeout entrypoint as its child
  [[ "$(grep '^run ' "$mp/calls.log" | grep -cE -- ' --init .* --entrypoint timeout ')" == 4 ]] ||
    fail "a probe container lacks --init (timeout would be PID 1 and never fire; a dropped connect lasts about 130 s)"
  [[ "$(grep -c '^run ' "$mp/calls.log")" == 4 ]] || fail "expected 4 probe containers (control + 3 endpoints), got $(grep -c '^run ' "$mp/calls.log")"
  for ev in allocate created removed; do
    [[ "$(awk -F'\t' -v e="$ev" '$4 == e' "$inv" | wc -l)" == 4 ]] || fail "inventory: expected 4 $ev rows"
  done
  awk -F'\t' '$4 == "created" && $6 == "sha256:img-edge" && $7 ~ /^[0-9a-f]{64}$/ {n++} END {exit n == 4 ? 0 : 1}' "$inv" || fail "inventory: created rows lack the 64-hex container ID or the image ID"
  n=0
  while read -r cid; do grep -q -- "^rm -f $cid\$" "$mp/calls.log" && n=$((n + 1)); done < <(awk -F'\t' '$4 == "created" {print $7}' "$inv" | sort)
  [[ $n == 4 ]] || fail "removal did not use the recorded container IDs ($n of 4)"
  [[ ! -s "$mp/containers.tsv" ]] || fail "a probe container remains after the run"
  [[ "$(grep -c $'\tprobe-container\tallocated ' "$mp/inventory-normal.out")" == 4 && "$(grep -c $'\tprobe-container\tcreated ' "$mp/inventory-normal.out")" == 4 ]] ||
    fail "allocation and creation are not both recorded as receipt steps"
  grep -q $'\tprobe-reap\tno probe container of staging remains' "$mp/inventory-normal.out" || fail "no probe-reap receipt step"
  [[ -z "$(ls -A "$mp/srv/state/probe-cids")" ]] || fail "cidfiles remain after a normal run (a cidfile goes only after the verified removal)"
  cf="$mp/srv/state/probe-cids/$seed_name"
  # reconciliation: an owned leftover or stale record is closed by identity (and its cidfile goes only then); anything
  # else is held and untouched, its cidfile kept as evidence, and no new probe container is created
  SEED=seed_owned mp_case stale-owned-container 1 1 "" denied
  grep -q "^rm -f $seed_id\$" "$mp/calls.log" && awk -F'\t' -v n="$seed_name" '$3 == n && $4 == "reaped"' "$inv" | grep -q . || fail "an owned stale container was not removed by its recorded identity (reaped row)"
  [[ ! -e $cf ]] || fail "the cidfile of a reaped stale container remains"
  SEED=seed_unowned mp_case unowned-container 1 1 "" held
  ! grep -q "^rm -f $seed_id" "$mp/calls.log" || fail "an unowned container was removed"
  awk -F'\t' -v n="$seed_name" '$3 == n && $4 == "held"' "$inv" | grep -q . || fail "no held row for the unowned container"
  [[ -e $cf ]] || fail "the cidfile of a held container was deleted"
  SEED=seed_mismatch_image mp_case mismatch-image 1 1 "" held
  ! grep -q "^rm -f $seed_id" "$mp/calls.log" || fail "a container with a different image ID was removed"
  SEED=seed_mismatch_cid mp_case mismatch-container-id 1 1 "" held
  ! grep -q "^rm -f $seed_id" "$mp/calls.log" || fail "a container whose ID differs from its cidfile was removed"
  SEED=seed_mismatch_run mp_case mismatch-run 1 1 "" held
  ! grep -q "^rm -f $seed_id" "$mp/calls.log" || fail "a container whose run label differs from its record was removed"
  SEED=seed_stale_record mp_case stale-record 1 1 "" denied
  awk -F'\t' -v n="$seed_name" '$3 == n && $4 == "stale-closed"' "$inv" | grep -q . || fail "a record without a container was not closed as stale"
  [[ ! -e $cf ]] || fail "the cidfile of a closed stale record remains"
  # a stale record whose cidfile cannot be unlinked stays open: nonzero, no terminal close, no probe container; the
  # retry after the unlink works closes it exactly once
  SEED=seed_stale_unlink mp_case stale-unlink-fails 1 1 "" blocked
  ! awk -F'\t' -v n="$seed_name" '$3 == n && $4 == "stale-closed"' "$inv" | grep -q . || fail "a stale record was closed although its cidfile could not be removed"
  grep -q $'\tprobe-reap\tthe cidfile of the stale record '"$seed_name"' could not be removed; the record stays open' "$mp/stale-unlink-fails.out" || fail "no receipt step for the failed cidfile removal"
  [[ -e $cf ]] || fail "the undeletable cidfile disappeared"
  rm -rf "$cf"; printf '%s' "$seed_id" >"$cf"
  KEEP=1 mp_case stale-unlink-retry 1 1 "" denied
  [[ "$(awk -F'\t' -v n="$seed_name" '$3 == n && $4 == "stale-closed"' "$inv" | wc -l)" == 1 && ! -e $cf ]] || fail "the retry did not close the stale record exactly once and remove its cidfile"
  SEED=seed_unrelated mp_case unrelated-container 1 1 "" denied
  ! grep -q "^rm -f $seed_id" "$mp/calls.log" || fail "an unrelated container (other environment label) was touched"
  [[ -s "$mp/containers.tsv" ]] || fail "the unrelated container disappeared"
  # failed and interrupted runs
  STUCK=1 mp_case removal-stuck 1 1 "" UNPROVEN
  awk -F'\t' '$4 == "remove-failed"' "$inv" | grep -q . || fail "a failed removal is not recorded (remove-failed)"
  [[ "$(grep -c '^run ' "$mp/calls.log")" == 1 ]] || fail "a later probe container was started after a failed removal (single network)"
  # an unresolved lifecycle failure stops the check at once: with two networks, a probe container that cannot be removed
  # (control, then a later endpoint) means no later endpoint and no second network is ever probed, the exact-owner
  # reconciliation is tried again, and the check fails
  NETS="net1 net2" STUCK_SEQ=2 mp_case endpoint-removal-stuck 1 1 "" UNPROVEN
  [[ "$(grep -c '^run ' "$mp/calls.log")" == 2 ]] || fail "docker run was invoked $(grep -c '^run ' "$mp/calls.log") times; expected the control and the first endpoint only"
  ! grep -q -- '--network net2' "$mp/calls.log" || fail "the second network was probed after an unresolved removal"
  grep -q $'\tmetadata-denied\ta probe container on net1 could not be proven created, recorded or removed; no further probe container is started' "$mp/endpoint-removal-stuck.out" || fail "no stop step for the unresolved endpoint removal"
  stuck_id="$(printf '%s' "$(grep '^run ' "$mp/calls.log" | sed -n 2p | sed -E 's/.*--name ([^ ]+) .*/\1/')" | sha256sum | cut -c1-64)"
  [[ "$(grep -c "^rm -f $stuck_id\$" "$mp/calls.log")" -ge 2 ]] || fail "the exact-owner reconciliation did not retry the stuck container"
  grep -q "$stuck_id" "$mp/containers.tsv" || fail "the stuck container vanished from the fixture table (the removal was supposed to fail)"
  awk -F'\t' '$4 == "remove-failed"' "$inv" | grep -q . || fail "the stuck endpoint removal is not recorded (remove-failed)"
  NETS="net1 net2" STUCK_SEQ=1 mp_case control-removal-stuck 1 1 "" UNPROVEN
  [[ "$(grep -c '^run ' "$mp/calls.log")" == 1 ]] || fail "a later probe container was started after the control could not be removed"
  ! grep -q -- '--network net2' "$mp/calls.log" || fail "the second network was probed after the control could not be removed"
  NOCREATE=1 mp_case no-container-id 1 1 "" UNPROVEN
  SEED=seed_owned MODE=cancel mp_case interrupted-cleanup 1 1 "" cancel
  grep -q "^rm -f $seed_id\$" "$mp/calls.log" || fail "cleanup (EXIT trap) did not reap the owned probe container of an interrupted run"
  [[ ! -e $cf ]] || fail "the cidfile of a container reaped by the interrupted run's cleanup remains"
  # unresolved reconciliation or an unwritable inventory: no new docker run
  SEED=seed_owned STUCK=1 mp_case preflight-stuck 1 1 "" blocked
  awk -F'\t' -v n="$seed_name" '$3 == n && $4 == "remove-failed"' "$inv" | grep -q . || fail "a stuck stale removal is not recorded (remove-failed)"
  [[ -e $cf ]] || fail "the cidfile of a container that could not be removed was deleted"
  SEED=seed_inv_dir mp_case inventory-unwritable 1 1 "" UNPROVEN
  [[ "$(grep -c '^run ' "$mp/calls.log" 2>/dev/null || true)" == 0 ]] || fail "a probe container was started although its allocate record could not be written"
  grep -q $'\tprobe-container\t.*: the allocate record could not be written; no container was started' "$mp/inventory-unwritable.out" || fail "no receipt step for the failed allocate record"
  NETS="net1 net2" BREAKINV=1 mp_case record-fails-after-create 1 1 "" UNPROVEN
  ! grep -q -- '--network net2' "$mp/calls.log" || fail "the second network was probed after a record failure"
  rf_name="$(grep '^run ' "$mp/calls.log" | head -n 1 | sed -E 's/.*--name ([^ ]+) .*/\1/')"
  rf_id="$(printf '%s' "$rf_name" | sha256sum | cut -c1-64)"
  [[ "$(grep -c '^run ' "$mp/calls.log")" == 1 ]] || fail "more than one probe container started after the inventory failed"
  grep -q "^rm -f $rf_id\$" "$mp/calls.log" && [[ ! -s "$mp/containers.tsv" && ! -e "$mp/srv/state/probe-cids/$rf_name" ]] ||
    fail "a probe container whose created record failed was not removed by its exact ID (with its cidfile)"
  [[ $ok == 1 ]] && ledger METADATA-PROBE PASS 0 "$(inputs_sha infra/staging/remote/lib.sh)" "$c" "fail closed: statuses 1, 124 and 143 with no answer are denied; 139, 137 and 125 with no answer are UNPROVEN; an answer (also with a failure status) is REACHABLE; a failed tool control is UNPROVEN and stops that network's probes; no wget -T. Identity: every probe container is named, labelled, has a cidfile and no --rm; allocate, created (64-hex container ID and image ID) and removed rows, removal by the recorded ID, receipt steps. Reconciliation: an owned stale container is reaped by identity, a stale record is closed, and an unowned container or one with a different image ID, container ID or run label is held (exit 1) and left untouched, as is an unrelated container; a failed removal, a missing container ID and an interrupted run are covered. Fail closed: an unresolved reconciliation (held, or a stuck removal) starts no probe container, an unwritable inventory starts none, a failed created record removes the exact recorded container, and a cidfile is deleted only after the container's absence is verified; an unresolved lifecycle failure (stuck removal of the control or of a later endpoint, with two networks; a record failure) stops the check at once: no later endpoint or network gets a docker run, and the exact-owner reconciliation is retried; a stale record is closed only after its cidfile is verified gone (a failed unlink keeps it open, unresolved, retryable)" \
    || ledger METADATA-PROBE FAIL 1 "$(inputs_sha infra/staging/remote/lib.sh)" "$c" "${why% ; }"
fi

if selected METADATA-GUARD; then
  ok=1 why="" rc=0; fail() { ok=0; why+="$1; "; }
  MG_INPUTS="$(inputs_sha infra/terraform/modules/gcp-staging-vm/startup.sh infra/staging/remote infra/compose)"
  mg_st="$REPO/infra/terraform/modules/gcp-staging-vm/startup.sh" rl="$REPO/infra/staging/remote"
  mg_run=(docker run --rm --pull never --network none "${LABELS[@]}" --cap-add NET_ADMIN --cap-add SYS_ADMIN
    --sysctl net.ipv4.ip_forward=1 -v "$REPO/infra/terraform/modules/gcp-staging-vm:/vm:ro" -v "$S/netns.sh:/netns.sh:ro"
    --entrypoint bash)
  c="grep/awk over startup.sh, remote/{lib,deploy,start,reset}.sh and infra/compose/{local,migrate,staging}.yml ; $(q "${mg_run[@]}") ${NETNS_IMAGE:-<--netns-image>} /netns.sh"
  # (a) Static: the unit, the drop-in, the rules in one place, the install order, the on-VM gate and the compose files.
  if [[ ! -f "$mg_st" ]]; then fail "startup.sh missing"
  else
    for want in 'GUARD_UNIT=staging-metadata-guard.service' "'Before=docker.service'" \
      "'Type=oneshot' 'RemainAfterExit=yes' \"ExecStart=\$GUARD_BIN\"" \
      '"Requires=$GUARD_UNIT" "After=$GUARD_UNIT" >/etc/systemd/system/docker.service.d/' \
      '-w -t raw -I PREROUTING -d "$2" -p "$proto" ! --dport 53 -m comment --comment staging-metadata-guard -j DROP' \
      'guard iptables 169.254.169.254/32' 'guard ip6tables fd20:ce::254/128' 'bash -c "$(metadata_guard_script)"' \
      'metadata_guard_script >"$GUARD_BIN.new"' 'systemctl enable --now "$GUARD_UNIT"'; do
      grep -qF -- "$want" "$mg_st" || fail "startup.sh lacks: $want"
    done
    order="$(grep -oE '^[[:space:]]*(install_metadata_guard|install_docker) </dev/null' "$mg_st" | tr -d ' ' | sed 's#</dev/null##' | tr '\n' ' ')"
    [[ "$order" == "install_metadata_guard install_docker install_metadata_guard install_docker " ]] ||
      fail "startup.sh must install the guard before Docker in --no-disk and disk mode (found: ${order:-none})"
  fi
  first="$(awk '$0 == "readiness() {" {getline; print; exit}' "$rl/lib.sh")"
  [[ "$first" == *"metadata_denied || hold "* ]] || fail "readiness does not start with metadata_denied || hold"
  for f in deploy start reset; do
    m="$(grep -n -m1 -E '^[[:space:]]*metadata_denied \|\| hold' "$rl/$f.sh" | cut -d: -f1 || true)"
    apl="$(grep -n -m1 -E '^[[:space:]]*app_up \|\|' "$rl/$f.sh" | cut -d: -f1 || true)"
    [[ -n "$m" && -n "$apl" ]] && (( m < apl )) || fail "remote/$f.sh: metadata_denied || hold does not precede app_up"
  done
  for f in local.yml migrate.yml staging.yml; do
    h="$(grep -vE '^[[:space:]]*#' "$REPO/infra/compose/$f" | grep -nE 'network_mode:[[:space:]]*"?host|169\.254\.|fd20:ce|metadata\.google' || true)"
    [[ -z "$h" ]] || fail "infra/compose/$f: host network or metadata address (${h//$'\n'/ | })"
  done
  # (b) Packet level, in an isolated network namespace, with the real `startup.sh --metadata-guard-rules`.
  if [[ -z "$NETNS_IMAGE" ]]; then fail "packet proof: missing prerequisite --netns-image (prepare --netns-tag g070-netns:TAG)"
  elif [[ -z "$(image_id "$NETNS_IMAGE")" ]]; then fail "packet proof: missing prerequisite: image $NETNS_IMAGE not present locally (prepare --netns-tag)"
  else
    cat >"$S/netns.sh" <<'NETNS'
#!/bin/bash
# METADATA-GUARD packet proof (infra/staging/check.sh) inside a --network none container. This container's namespace is
# the host (forwarding on); a fake container (veth c0/c1, 10.99.1.0/24) and the metadata server (veth m0/m1,
# 10.99.2.0/24, holding 169.254.169.254) are unshare -n children, so the host's own requests leave through OUTPUT and
# the container's arrive through PREROUTING, as on Compute Engine. Prints probe lines and one RESULT line; the
# listener serves a fixed marker, never a credential.
set -u
pids=()
trap 'for p in "${pids[@]}"; do kill "$p" 2>/dev/null; done' EXIT
fail() { echo "RESULT FAIL $*"; exit 1; }
must() { "$@" >/dev/null 2>&1 || fail "setup step failed: $*"; }
for t in ip iptables ip6tables unshare nsenter timeout busybox; do command -v "$t" >/dev/null 2>&1 || fail "missing prerequisite: $t"; done
for a in wget httpd; do busybox --list | grep -qx "$a" || fail "missing prerequisite: busybox $a applet"; done
[[ "$(cat /proc/sys/net/ipv4/ip_forward 2>/dev/null)" == 1 ]] || fail "missing prerequisite: net.ipv4.ip_forward=1 (docker run --sysctl)"
iptables -w -t raw -S >/dev/null 2>&1 || fail "missing prerequisite: the iptables raw table (CAP_NET_ADMIN, kernel netfilter)"
if iptables -w -t raw -S PREROUTING | grep -q staging-metadata-guard; then fail "guard rules present before the control"; fi
newns() {
  unshare -n sleep 600 &
  NSPID=$!
  pids+=("$NSPID")
  sleep 0.3
  [[ -n "$(readlink "/proc/$NSPID/ns/net" 2>/dev/null)" && "$(readlink "/proc/$NSPID/ns/net")" != "$(readlink /proc/self/ns/net)" ]] ||
    fail "missing prerequisite: unshare -n (CAP_SYS_ADMIN; seccomp)"
}
ns_do() { local p=$1; shift; nsenter -t "$p" -n "$@"; }
newns; C=$NSPID
newns; M=$NSPID
must ip link set lo up
must ip link add c0 type veth peer name c1
must ip link set c1 netns "$C"
must ip link add m0 type veth peer name m1
must ip link set m1 netns "$M"
must ip addr add 10.99.1.1/24 dev c0
must ip link set c0 up
must ip addr add 10.99.2.1/24 dev m0
must ip link set m0 up
must ip route add 169.254.169.254/32 via 10.99.2.2
must ns_do "$C" ip link set lo up
must ns_do "$C" ip addr add 10.99.1.2/24 dev c1
must ns_do "$C" ip link set c1 up
must ns_do "$C" ip route add default via 10.99.1.1
must ns_do "$M" ip link set lo up
must ns_do "$M" ip addr add 10.99.2.2/24 dev m1
must ns_do "$M" ip addr add 169.254.169.254/32 dev m1
must ns_do "$M" ip link set m1 up
must ns_do "$M" ip route add default via 10.99.2.1
www=/tmp/www/computeMetadata/v1/instance/service-accounts/default
mkdir -p "$www" && printf 'G070-METADATA-MARKER\n' >"$www/token" || fail "setup step failed: listener files"
must ns_do "$M" busybox httpd -p 169.254.169.254:80 -h /tmp/www
must ns_do "$M" busybox httpd -p 169.254.169.254:53 -h /tmp/www
sleep 0.5
probe() { # LABEL host|PID PORT reach|drop
  local out rc=0 url="http://169.254.169.254:$3/computeMetadata/v1/instance/service-accounts/default/token"
  # no wget -T: Debian's busybox 1.35 wget segfaults with it; timeout bounds a dropped connection instead
  if [[ $2 == host ]]; then out="$(timeout 5 busybox wget -q -O- "$url" 2>/dev/null)" || rc=$?
  else out="$(ns_do "$2" timeout 5 busybox wget -q -O- "$url" 2>/dev/null)" || rc=$?; fi
  if [[ $rc == 0 && $out == G070-METADATA-MARKER ]]; then printf 'probe\t%s\treached\n' "$1"; [[ $4 == reach ]]
  else printf 'probe\t%s\tno answer (wget exit %s)\n' "$1" "$rc"; [[ $4 == drop && -z $out ]]; fi
}
probe control-container "$C" 80 reach || fail "control: the fake container does not reach the listener without the guard (topology)"
probe control-host host 80 reach || fail "control: the host does not reach the listener (topology)"
out="$(bash /vm/startup.sh --metadata-guard-rules 2>&1)" || fail "startup.sh --metadata-guard-rules: exit $? ($out)"
printf 'guard\t%s\n' "${out//$'\n'/ }"
bash /vm/startup.sh --metadata-guard-rules >/dev/null 2>&1 || fail "a second startup.sh --metadata-guard-rules failed (not idempotent)"
n4="$(iptables -w -t raw -S PREROUTING | grep -c staging-metadata-guard)"
[[ $n4 == 2 ]] || fail "IPv4 guard rules after two runs: $n4, want 2 (tcp, udp)"
printf 'rules\tipv4\t%s\n' "$n4"
if ip6tables -w -t raw -S >/dev/null 2>&1; then
  n6="$(ip6tables -w -t raw -S PREROUTING | grep -c staging-metadata-guard)"
  [[ $n6 == 2 ]] || fail "IPv6 guard rules after two runs: $n6, want 2 (tcp, udp)"
  printf 'rules\tipv6\t%s\n' "$n6"
else
  printf 'rules\tipv6\tno IPv6 raw table (skipped by design)\n'
fi
probe guarded-container "$C" 80 drop || fail "the fake container still reaches the metadata listener on port 80"
probe guarded-host host 80 reach || fail "the host no longer reaches the metadata listener"
probe guarded-container-port-53 "$C" 53 reach || fail "port 53 (DNS) to the metadata address is blocked"
echo "RESULT PASS"
NETNS
    prc=0
    "${mg_run[@]}" "$NETNS_IMAGE" /netns.sh >"$S/netns.out" 2>&1 || prc=$?
    cp "$S/netns.out" "$REPORT/metadata-guard-netns.txt"
    if [[ $prc == 0 ]] && grep -qx 'RESULT PASS' "$S/netns.out"; then
      why+="packet proof: $(grep -E '^(probe|rules)' "$S/netns.out" | tr '\t\n' ': ' | sed 's/ $//'); "
    else
      fail "packet proof: $(grep '^RESULT' "$S/netns.out" | tail -1 || true) (docker run exit $prc; metadata-guard-netns.txt)"
    fi
  fi
  [[ $ok == 1 ]] && ledger METADATA-GUARD PASS 0 "$MG_INPUTS" "$c" "static: guard unit Before=docker.service, docker drop-in Requires=/After=, raw PREROUTING DROP for 169.254.169.254 and fd20:ce::254 (tcp/udp except 53) in one script, installed before Docker in both modes; metadata_denied first in readiness and before app_up in deploy/start/reset; no host network or metadata address in the compose files; ${why% ; }" \
    || ledger METADATA-GUARD FAIL 1 "$MG_INPUTS" "$c" "${why% ; }"
fi

# ---- G072 fixture rows (Plan §10.13): one offline fixture script each, fake tools only (no Docker, network or cloud) --
# fixture_row ROW SCRIPT RESULT_ERE INPUT_PATH...: PASS only when the script exits 0 and its last line matching
# RESULT_ERE reports at least one case and no failure (group 1 = cases, group 2 = passed or cases, group 3 = failed).
fixture_row() {
  local row=$1 script=$2 re=$3 low w r rc=0 last n p f c det sel=() bad
  shift 3
  low="$(tr '[:upper:]' '[:lower:]' <<<"$row")" w="$S/$low" r="$REPORT/$low"
  [[ -z "${FIXTURE_CASES[$row]:-}" ]] || sel=(--only "${FIXTURE_CASES[$row]}")
  c="bash $script --work <scratch>/$low --report $r${sel[*]:+ ${sel[*]}}"
  if [[ ! -f "$REPO/$script" ]]; then ledger "$row" FAIL - - "$c" "$script missing"; return 0; fi
  (cd "$REPO" && bash "$script" --work "$w" --report "$r" "${sel[@]}") >"$REPORT/$low.out" 2>&1 </dev/null || rc=$?
  last="$(grep -E "$re" "$REPORT/$low.out" | tail -n 1 || true)"
  n="" p="" f=""
  if [[ "$last" =~ $re ]]; then n=${BASH_REMATCH[1]} p=${BASH_REMATCH[2]} f=${BASH_REMATCH[3]}; fi
  # The detail shows the script's own last result line, PASS or FAIL.
  det="$(grep -E '^(RESULT |ci-launcher: )' "$REPORT/$low.out" | tail -n 1 || true)"
  det="${det:-no result line}; ${sel[*]:+subset ${sel[*]}; }stdout $low.out, report $low/"
  if [[ $rc == 0 && -n "$n" && "$n" -ge 1 && "$p" == "$n" && "$f" == 0 ]]; then
    ledger "$row" PASS 0 "$(inputs_sha "$@")" "$c" "$det"
  else
    bad="$(cat "$r/summary.tsv" "$r/cases.tsv" 2>/dev/null | awk -F'\t' '$2 == "FAIL" {print $1}' | sort -u | head -12 | tr '\n' ' ' || true)"
    ledger "$row" FAIL "$rc" "$(inputs_sha "$@")" "$c" "$det; exit $rc${bad:+; failed: $bad}"
  fi
}
# ci-adapter.sh and env-lease.sh: "RESULT PASS|FAIL cases=N failed=M"; ci-federation.sh: "RESULT PASS|FAIL passed=N
# failed=M" (its passed count stands for both); ci-launcher.sh prints no RESULT line, only "ci-launcher: N selected, P
# passed, F failed".
if selected CI-FEDERATION; then
  fixture_row CI-FEDERATION infra/staging/tests/ci-federation.sh '^RESULT PASS passed=(([0-9]+)) failed=([0-9]+)$' \
    infra/staging/tests/ci-federation.sh infra/terraform/bootstrap infra/ci/hosted-context.sh infra/staging/ci-environments.txt
fi
if selected CI-ADAPTER; then
  fixture_row CI-ADAPTER infra/staging/tests/ci-adapter.sh '^RESULT PASS cases=(([0-9]+)) failed=([0-9]+)$' \
    infra/staging infra/terraform/bootstrap/ci-meta-permissions.tsv
fi
if selected ENV-LEASE; then
  fixture_row ENV-LEASE infra/staging/tests/env-lease.sh '^RESULT PASS cases=(([0-9]+)) failed=([0-9]+)$' \
    infra/staging infra/terraform/bootstrap/ci-meta-permissions.tsv
fi
if selected CI-LAUNCHER; then
  fixture_row CI-LAUNCHER infra/staging/tests/ci-launcher.sh '^ci-launcher: ([0-9]+) selected, ([0-9]+) passed, ([0-9]+) failed$' \
    infra/ci/staging-action.sh infra/ci/hosted-context.sh infra/staging
fi

# ---- Focused lint, directly on G070's files (Codex fallback while G069 owns run.sh) ------------------------------------
if selected SHELLCHECK-DIRECT; then
  # G072: exactly B-SHELL's staging entries (infra/validation/run.sh lint_shellcheck: infra/staging, infra/terraform and
  # the launcher infra/ci/staging-action.sh); a difference between the two lists, or a tracked infra/staging or
  # infra/terraform script in neither (tests/fakes and tests/fixtures aside), fails the row before any lint.
  sh_files=(infra/staging/staging.sh infra/staging/lib.sh infra/staging/check.sh infra/staging/remote/deploy.sh
    infra/staging/remote/lib.sh infra/staging/remote/reset.sh infra/staging/remote/start.sh infra/staging/remote/status.sh
    infra/staging/remote/stop.sh infra/staging/ci.sh infra/staging/ci-lib.sh infra/staging/meta-lib.sh infra/staging/ci-publish.sh
    infra/staging/tests/ci-adapter.sh infra/staging/tests/env-lease.sh infra/staging/tests/ci-federation.sh
    infra/staging/tests/ci-launcher.sh infra/staging/tests/ci-build-boundary.sh infra/terraform/bootstrap/bootstrap.sh
    infra/terraform/bootstrap/ci-federation.sh infra/terraform/modules/gcp-staging-vm/startup.sh infra/ci/staging-action.sh)
  SC_INPUTS="$(inputs_sha "${sh_files[@]}" infra/validation/run.sh)"
  sc=(docker run --rm --pull never --network none "${LABELS[@]}" -v "$REPO:/repo:ro" -w /repo "$SHELLCHECK_IMAGE" -S warning -x -P SCRIPTDIR)
  sc_mine="$(printf '%s\n' "${sh_files[@]}" | sort -u)"
  sc_bshell="$(awk '/^lint_shellcheck\(\) \{/,/^}/' "$REPO/infra/validation/run.sh" |
    grep -oE 'infra/(staging|terraform)/[A-Za-z0-9_./-]+\.sh|infra/ci/staging-action\.sh' | sort -u || true)"
  sc_diff="$(diff <(printf '%s\n' "$sc_mine") <(printf '%s\n' "$sc_bshell") | grep -E '^[<>]' | sed 's/^</only here:/; s/^>/only in B-SHELL:/' | tr '\n' ' ' || true)"
  sc_missing="$(cd "$REPO" && git ls-files -- 'infra/staging/*.sh' 'infra/terraform/*.sh' |
    grep -vE '^infra/staging/tests/(fakes|fixtures)/' | grep -vxF -f <(printf '%s\n' "$sc_mine") | tr '\n' ' ' || true)"
  if [[ -n "$sc_diff" ]]; then ledger SHELLCHECK-DIRECT FAIL - "$SC_INPUTS" "$(q "${sc[@]}")" "list differs from B-SHELL (run.sh lint_shellcheck): $sc_diff"
  elif [[ -n "$sc_missing" ]]; then ledger SHELLCHECK-DIRECT FAIL - "$SC_INPUTS" "$(q "${sc[@]}")" "tracked scripts linted by neither list: $sc_missing"
  elif (( ${#sh_files[@]} == 0 )); then ledger SHELLCHECK-DIRECT FAIL - "$SC_INPUTS" "$(q "${sc[@]}")" "no G070 scripts"
  elif [[ -z "$(image_id "$SHELLCHECK_IMAGE")" ]]; then ledger SHELLCHECK-DIRECT FAIL - "$SC_INPUTS" "$(q "${sc[@]}")" "$SHELLCHECK_IMAGE not prepared"
  else
    rc=0; "${sc[@]}" "${sh_files[@]}" >"$REPORT/shellcheck.log" 2>&1 || rc=$?
    [[ $rc == 0 ]] && ledger SHELLCHECK-DIRECT PASS 0 "$SC_INPUTS" "$(q "${sc[@]}" "${sh_files[@]}")" "$SHELLCHECK_IMAGE; ${#sh_files[@]} files: ${sh_files[*]}" \
      || ledger SHELLCHECK-DIRECT FAIL "$rc" "$SC_INPUTS" "$(q "${sc[@]}" "${sh_files[@]}")" "$SHELLCHECK_IMAGE; findings in shellcheck.log; files: ${sh_files[*]}"
  fi
fi

if selected HADOLINT-DIRECT; then
  df=infra/docker/edge.Dockerfile
  hd=(docker run --rm -i --pull never --network none "${LABELS[@]}" "$HADOLINT_IMAGE" hadolint --failure-threshold warning -)
  if [[ ! -f "$REPO/$df" ]]; then ledger HADOLINT-DIRECT FAIL - "-" "$(q "${hd[@]}") <$df" "$df missing"
  elif [[ -z "$(image_id "$HADOLINT_IMAGE")" ]]; then ledger HADOLINT-DIRECT FAIL - "$(inputs_sha "$df")" "$(q "${hd[@]}") <$df" "$HADOLINT_IMAGE not prepared"
  else
    rc=0; "${hd[@]}" <"$REPO/$df" >"$REPORT/hadolint.log" 2>&1 || rc=$?
    [[ $rc == 0 ]] && ledger HADOLINT-DIRECT PASS 0 "$(inputs_sha "$df")" "$(q "${hd[@]}") <$df" "$HADOLINT_IMAGE; $df" \
      || ledger HADOLINT-DIRECT FAIL "$rc" "$(inputs_sha "$df")" "$(q "${hd[@]}") <$df" "$HADOLINT_IMAGE; findings in hadolint.log"
  fi
fi

n_sel=0
for r in "${ROWS[@]}"; do
  if selected "$r"; then n_sel=$((n_sel + 1)); else ledger "$r" NOT-SELECTED - - - "not run in this invocation (--only $ONLY)"; fi
done
log "run $RUN_ID: $n_sel of ${#ROWS[@]} rows selected, $FAILS FAIL; ledger $LEDGER"
(( FAILS == 0 ))
