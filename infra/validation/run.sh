#!/usr/bin/env bash
# Project validation harness: validates this repository and disposable copies
# of it against the bootstrap acceptance criteria, in containers only.
#
#   infra/validation/run.sh --base REV --allow-file FILE
#       [--expect-hash PATH=SHA256]... [--exception PATH]...
#       [--keep] [--report-dir DIR] [--owned-registry FILE] [--secrets-file FILE] [--areas LIST] SUITE...
#
#   --base         commit that protected files and the change scope are compared with
#   --allow-file   extended regular expressions, one per line, for paths a change may touch
#   --expect-hash  a file that legitimately differed from --base at the start and must
#                  still match this sha256 (for example an approved plan)
#   --exception    an exact path changed by its owner outside this change (reported)
#   --owned-registry  append owned-resource rows (work_root, label, project, rootenv, image,
#                  capture and standalone rows) to FILE for infra/ci/cleanup-owned.sh (G003 §11.4);
#                  standalone cidfiles go to "$(dirname FILE)/cid"
#   --secrets-file this run's generated secrets and canaries are appended to FILE (outside $WORK)
#   --areas        the selected feature areas for feature-web/feature-core: none (default) or a sorted list of
#                  KIND:NAME (KIND web, core or intel), as infra/ci/select-profile.sh writes AREAS (G051)
#   SUITE          lint | lint-base | lint-web | lint-core | controls | ci-controls | provenance | contracts | unit | unit-web |
#                  unit-core | unit-intel | root | migrations | ops | rename | off | feature-web | feature-core |
#                  staging, or an alias: static | variants | all (all = the full profile; never the feature suites)
#                  (staging: the real WorkOS staging contract, only with IDENTITY_STAGING_ENV_FILE; G004 CI-1)
#   HARNESS_PLAYWRIGHT_WORKERS  Playwright workers for the harness's own browser runs (1-4, default 2; G051 L6)
#
# Safety: work happens outside the repository; destructive tests use disposable
# copies and uniquely named Compose projects; an existing root .env or existing
# Compose resources of the root project are never touched; only resources this
# run created are removed. Host Go/Python/Node/psql/redis-cli/mc are shadowed.
# Exit status: 0 when every recorded check is PASS (or INFO), 1 otherwise.
set -euo pipefail

REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)"
# shellcheck source=lib.sh
source "$REPO/infra/validation/lib.sh"
# shellcheck source=../ci/lib.sh
source "$REPO/infra/ci/lib.sh"
# G066 live progress: validate.sh exports PLATFORM_PROGRESS_FILE for this run's own record() rows. It stays a shell
# variable of this process and of its forked children (bg_start subshells), so record() keeps appending, but it is no
# longer exported: the processes this harness launches (validate-controls.sh and its fixtures, nested run.sh runs,
# drivers) never inherit it, so their rows cannot appear as this run's live progress (RC9: VB7's B-PERSIST FAIL
# fixture and C_SEC_9's nested staging run were printed as false failures and hid the real rows).
if [[ -n "${PLATFORM_PROGRESS_FILE:-}" ]]; then export -n PLATFORM_PROGRESS_FILE; fi

# The full profile (B-CI-CONFIG: = CI_SUITES_full of infra/ci/lib.sh); G051 splits lint into lint (the infrastructure
# checks), lint-base, lint-web and lint-core. The feature suites need --areas and are never part of all.
SUITES_ALL=(lint lint-base lint-web lint-core controls ci-controls provenance contracts root migrations ops rename off)
# Every suite this harness runs (G051 adds lint-base/web/core, unit-web/core/intel and feature-web/core).
SUITES_KNOWN=(lint lint-base lint-web lint-core controls ci-controls provenance contracts unit unit-web unit-core unit-intel root
  migrations ops rename off feature-web feature-core staging)
AREAS_RE='^(none|(web|core|intel):[a-z][a-z0-9-]{0,31}(,(web|core|intel):[a-z][a-z0-9-]{0,31})*)$'

BASE="" ALLOW_FILE="" KEEP=0 REPORT_DIR="" SUITES=() AUTO_CI_CONTROLS=0 EXPECT=() EXCEPTIONS=() SECRETS_ARG="" AREAS=none
while [[ $# -gt 0 ]]; do
  case "$1" in
    --base) BASE="$2"; shift 2 ;;
    --allow-file) ALLOW_FILE="$(cd "$(dirname "$2")" && pwd -P)/$(basename "$2")"; shift 2 ;;
    --expect-hash) EXPECT+=("$2"); shift 2 ;;
    --exception) EXCEPTIONS+=("$2"); shift 2 ;;
    --keep) KEEP=1; shift ;;
    --report-dir) mkdir -p "$2"; REPORT_DIR="$(cd "$2" && pwd -P)"; shift 2 ;;
    --owned-registry) OWNED_REG="$2"; shift 2 ;;
    --secrets-file) SECRETS_ARG="$2"; shift 2 ;;
    --areas)
      [[ $# -ge 2 && "$2" =~ $AREAS_RE ]] || { echo "--areas: expected none or KIND:NAME[,KIND:NAME]... (KIND web, core or intel)" >&2; exit 2; }
      AREAS="$2"; shift 2 ;;
    lint|lint-base|lint-web|lint-core|controls|ci-controls|provenance|contracts|unit|unit-web|unit-core|unit-intel|root|migrations|ops|rename|off|feature-web|feature-core|staging) SUITES+=("$1"); shift ;;
    # Aliases (G003 CI revision; G051): static = the four lint suites and controls; variants = ops rename off;
    # all = the full profile.
    static) SUITES+=(lint lint-base lint-web lint-core controls ci-controls); shift ;;
    variants) SUITES+=(ops rename off); shift ;;
    all) AUTO_CI_CONTROLS=1; SUITES+=("${SUITES_ALL[@]}"); shift ;;
    *) echo "unknown argument: $1" >&2; exit 2 ;;
  esac
done
if [[ ${#SUITES[@]} -eq 0 || -z "$BASE" || -z "$ALLOW_FILE" ]]; then
  echo "usage: run.sh --base REV --allow-file FILE [--expect-hash PATH=SHA256]... [--exception PATH]... [--keep] [--report-dir DIR] [--owned-registry FILE] [--secrets-file FILE] [--areas LIST] lint|lint-base|lint-web|lint-core|controls|ci-controls|provenance|contracts|unit|unit-web|unit-core|unit-intel|root|migrations|ops|rename|off|feature-web|feature-core|staging|static|variants|all..." >&2
  exit 2
fi
# G051 L6: the Playwright worker count of the harness's own browser runs (web/playwright.config.ts accepts 1-4).
PW_WORKERS="${HARNESS_PLAYWRIGHT_WORKERS:-2}"
[[ "$PW_WORKERS" =~ ^[1-4]$ ]] || { echo "HARNESS_PLAYWRIGHT_WORKERS: expected an integer from 1 to 4" >&2; exit 2; }
# Each suite once, in the order given.
declare -A SEEN_SUITE=(); DEDUP=()
for s in "${SUITES[@]}"; do if [[ -z "${SEEN_SUITE[$s]+x}" ]]; then SEEN_SUITE[$s]=1; DEDUP+=("$s"); fi; done
SUITES=("${DEDUP[@]}")
git -C "$REPO" rev-parse --verify --quiet "$BASE^{commit}" >/dev/null || { echo "--base: unknown commit $BASE" >&2; exit 2; }
[[ -f "$ALLOW_FILE" ]] || { echo "--allow-file: not found" >&2; exit 2; }
if (( AUTO_CI_CONTROLS )) && ! ci_controls_required "$REPO" "$BASE" HEAD; then
  DEDUP=()
  for s in "${SUITES[@]}"; do [[ "$s" == ci-controls ]] || DEDUP+=("$s"); done
  SUITES=("${DEDUP[@]}")
  echo "CI meta-tests not selected: no changed CI/harness/control dependency (set CI_CONTROLS_FORCE=1 to force)."
fi
if [[ -n "$OWNED_REG" ]]; then
  [[ -f "$OWNED_REG" && ! -L "$OWNED_REG" && -d "$(dirname "$OWNED_REG")/cid" ]] || { echo "--owned-registry: not a registry file" >&2; exit 2; }
  OWNED_REG="$(cd "$(dirname "$OWNED_REG")" && pwd -P)/$(basename "$OWNED_REG")"
fi

RUN_ID="$(random_hex 3)"
WORK="$(mktemp -d "${TMPDIR:-/tmp}/project-validation-XXXXXX")"
if [[ "$WORK" == "$REPO"/* ]]; then echo "work directory must be outside the repository" >&2; exit 2; fi
RESULTS="$WORK/results.tsv"; : >"$RESULTS"
if [[ -n "$SECRETS_ARG" ]]; then
  # An external secrets list (validate.sh's registry) is outside $WORK, so the scans need no exclusion.
  [[ "$(realpath -m -- "$SECRETS_ARG")" != "$REPO"/* ]] || { echo "--secrets-file: inside the repository" >&2; exit 2; }
  (umask 077; : >>"$SECRETS_ARG"); SECRETS_FILE="$(realpath -e -- "$SECRETS_ARG")"
else
  SECRETS_FILE="$WORK/secrets"; : >"$SECRETS_FILE"
fi
HARNESS_LOG="$WORK/harness.log"
docker_inventory >"$WORK/inventory-start.txt"
docker_inventory_ids >"$WORK/inventory-start.ids"
# G051 (Plan r3 §3.6): every exit path stops unjoined background children before ownership cleanup. G064 (addendum A
# §A4): then every batch runner is torn down (verified) before go_tests_finalize, because an exec'd go test outlives its
# killed `docker exec` client until the runner goes. G065 K3: on an abort before res_witness_stop, the witness W is then
# removed by exact ID (res_witness_abort; a no-op once res_witness_stop ran).
trap 'RUN_STATUS=$?; bg_stop_all; res_witness_abort || log "witness abort cleanup failed"; it_runner_stop_all; go_tests_finalize || log "go-tests finalization failed"; export_diagnostics || log "diagnostic export failed; retain scanned harness.log"; cleanup' EXIT
setup_path_hygiene
registry_row work_root "$WORK"
registry_row label "g001.run=$RUN_ID"
# tee ignores TERM, INT and HUP, so under a group TERM the EXIT-trap cleanup can still log and run (G003 F26).
exec > >(trap '' TERM INT HUP; exec tee -a "$HARNESS_LOG") 2>&1

log "run $RUN_ID; base $(git -C "$REPO" rev-parse --short "$BASE"); work $WORK; suites: ${SUITES[*]}; areas: $AREAS"

# --- preflight and shared facts ------------------------------------------------------------
compose_version="$(docker compose version --short 2>/dev/null || true)"
major="${compose_version%%.*}"; rest="${compose_version#*.}"; minor="${rest%%.*}"
if [[ -z "$compose_version" || "$major" -lt 2 || ( "$major" -eq 2 && "$minor" -lt 24 ) ]]; then
  record PREFLIGHT FAIL "Docker Compose >= 2.24 required (found '${compose_version:-none}')"; exit 1
fi
record PREFLIGHT INFO "docker $(docker version --format '{{.Server.Version}}'), compose $compose_version"
# G051 build ledger (Plan r3 §3.8): once, before any build; each suite's builds carry LEDGER_SUITE.
LEDGER_SUITE=preflight
ledger_init
# G051 L4 (Plan r3 §3.4): an affected run (feature-web without root) starts a cache-only prebuild of the web-browser
# image, supervised in the background, with exactly the Dockerfile, target, context and arguments of suite_feature_web's
# authoritative build (FEATURE_WEB_BUILD plus web/; no tag); the consumer waits for it to end, whatever its status,
# before building, and every prebuild is waited for before the final checks. A prebuild never records PASS.
# G051 C4: the web-check build (unit, unit-web) also waits for it first: BuildKit merges identical in-flight
# vertices of concurrent builds (web-browser's fixture stage and web.Dockerfile's deps and source stages), and a
# merged vertex carries the other build's stage label and index, which the ledger cannot attribute.
FEATURE_WEB_BUILD=(--label "g001.run=$RUN_ID" -f "$REPO/infra/docker/web-browser.Dockerfile" --target browser)
PREBUILD_WEB=0
if [[ -n "${SEEN_SUITE[feature-web]:-}" && -z "${SEEN_SUITE[root]:-}" ]]; then
  LEDGER_SUITE=feature-web
  bg_start prebuild-web-browser ledger_prebuild prebuild-web-browser -- "${FEATURE_WEB_BUILD[@]}" "$REPO/web"
  PREBUILD_WEB=1
  LEDGER_SUITE=preflight
fi
detect_probe
record PROBE INFO "host probe mode: $PROBE_MODE"

# The SaaS Foundation generator, its manifest and the render-customization register are retired and not part of this
# distribution; the committed repository is the maintained template. The configured identity values come from the
# committed .env.example (B-LITERALS), not from a renderer manifest.
ROOT_PROJECT="$(sed -n 's/^COMPOSE_PROJECT_NAME=//p' "$REPO/.env.example")"
mapfile -t DEFAULT_PORTS < <(for k in WEB_PORT CORE_PORT INTELLIGENCE_PORT POSTGRES_PORT REDIS_PORT MINIO_API_PORT MINIO_CONSOLE_PORT; do sed -n "s/^$k=//p" "$REPO/.env.example"; done)

# Host-port sets for disposable projects (web core intelligence postgres redis minio-api minio-console).
PORTS_MIG=(3400 8480 8400 5832 6779 9400 9401)
PORTS_OPS=(3200 8280 8200 5632 6579 9200 9201)
PORTS_REN=(3300 8380 8300 5732 6679 9300 9301)
PORTS_OFF=(3500 8580 - 5932 6879 9500 9501)

# The reviewed core/cmd/server/main.go: G003's three hunks (Plan §7.1), G004's identity wiring (configuration, the
# WorkOS adapter when configured, the service and its five routes) and G014's MCP wiring (configuration; the Connect
# validator, the handler and its three internal setup routes only when configured) and G016's hostname wiring (the
# domain configuration, the resolver, the DNS and TLS adapters, the relay route and the domain routes registered by
# core/internal/domains/api), and its import set (the rendered set minus fmt, plus the identity packages,
# core/internal/mcp, core/internal/domains, core/internal/domains/api and strings), pinned for B-MCP-SOURCE. G017 adds
# the branding routes registered by core/internal/branding/api (its own access facade over the owned object store built
# from the loaded settings) and that import; G018 the onboarding routes registered by core/internal/onboarding/api (its
# own access facade) and that import; G019 the membership routes registered by core/internal/membership/api (its own
# access facade) and that import; G058 (core slice 1a) the enrollment admission composed by core/internal/enrollment/api
# (its own access facade, bound into identity before serving) and that import, and (HTTP slice) its routes registered
# by the same handler with the shared resolver; G062 (Plan r5 §2, §4.2) the root origin wired into identity and the
# platform root signup routes registered by core/internal/signup/api over the core/internal/signup service, and those
# two imports; G063 (Plan r1 §2) one more MCP setup mount, the rename route; G020 (Plan r2 §2) the label routes
# registered by core/internal/labels/api (its own access facade) and that import; the variable keeps its G016 name.
# G071 (Plan r4 S3, S15): the default-tenant selection and root presentation wired before the resolver (no new import).
MAIN_GO_G016_SHA=d4d3ddd575abcbff638a91e9ab839923fc1da6a5527183e4f15e4f1b591a7914
MAIN_GO_IMPORTS="context core/internal/branding/api core/internal/config core/internal/domains core/internal/domains/api core/internal/enrollment/api core/internal/health core/internal/identity core/internal/identity/workos core/internal/labels/api core/internal/mcp core/internal/membership/api core/internal/observability core/internal/onboarding/api core/internal/signup core/internal/signup/api errors flag github.com/aws/aws-sdk-go-v2/aws github.com/aws/aws-sdk-go-v2/credentials github.com/aws/aws-sdk-go-v2/service/s3 github.com/jackc/pgx/v5/pgxpool log/slog net net/http os os/signal strconv strings syscall time"
# The six core auth routes (G004 Plan r5 §7.2; G016 Plan r4 §1.7: the relay), mounted by main.go through the identity
# route constants.
CORE_AUTH_ROUTES="GET /internal/v1/auth/relay|GET /internal/v1/auth/session|POST /internal/v1/auth/callback|POST /internal/v1/auth/login|POST /internal/v1/auth/logout|POST /internal/v1/auth/provider-logout/return"
# The four internal MCP setup routes (G014 Plan r8 §3.2; G063 Plan r1 §4.3: rename), mounted by main.go through the mcp
# route constants; C-sorted.
CORE_MCP_ROUTES="GET /internal/v1/mcp/connections|POST /internal/v1/mcp/connections|POST /internal/v1/mcp/connections/rename|POST /internal/v1/mcp/connections/revoke"
# The eight G018 onboarding method/path pairs on six paths (Plan r5 §5.3), mounted by main.go through the onboarding
# handler's Register; pinned for B-MCP-SOURCE.
CORE_ONBOARDING_ROUTES="GET /internal/v1/business-profiles|GET /internal/v1/business-profiles/{id}|GET /internal/v1/groups|GET /internal/v1/groups/{id}|POST /internal/v1/groups|POST /internal/v1/groups/slug-suggestion|POST /internal/v1/groups/{id}/business-profile|PUT /internal/v1/business-profiles/{id}"
# The sixteen G019 membership method/path pairs on fifteen paths (Plan r1 §5.3, r2 addendum §7) and the two G060
# Group-authority pairs on /internal/v1/organization/group-authority (Plan r0 §4.2): eighteen pairs on sixteen paths,
# mounted by main.go through the membership handler's Register; C-sorted; pinned for B-MCP-SOURCE.
CORE_MEMBERSHIP_ROUTES="GET /internal/v1/account/retirement-eligibility|GET /internal/v1/invitations|GET /internal/v1/invitations/candidates|GET /internal/v1/invitations/mine|GET /internal/v1/invitations/{id}|GET /internal/v1/members|GET /internal/v1/members/self|GET /internal/v1/organization/group-authority|POST /internal/v1/invitations|POST /internal/v1/invitations/{id}/accept|POST /internal/v1/invitations/{id}/revoke|POST /internal/v1/members/leave|POST /internal/v1/members/{user_id}/demote|POST /internal/v1/members/{user_id}/promote|POST /internal/v1/members/{user_id}/remove|POST /internal/v1/members/{user_id}/role|POST /internal/v1/organization/group-authority|POST /internal/v1/ownership/transfer"
# The G058 enrollment method/path pairs (Plan r3 §5.2 = r0 §5.2's enumerated table plus the revoke route: the public
# projection, the applicant's four own-credential routes and the manager's policy, list, decision and revocation; ten
# pairs on nine paths — the Plan's "9 pairs / 8 paths" count omits the anonymous public route it enumerates), mounted by
# main.go through the enrollment handler's Register; C-sorted; pinned for B-MCP-SOURCE.
# The seven G062 platform root signup method/path pairs (Plan r5 §4.2: status, slug, organizations, challenge, prove,
# complete, abandon), mounted by main.go through the signup handler's Register; C-sorted; pinned for B-MCP-SOURCE.
CORE_SIGNUP_ROUTES="GET /internal/v1/signup/status|POST /internal/v1/signup/abandon|POST /internal/v1/signup/challenge|POST /internal/v1/signup/complete|POST /internal/v1/signup/organizations|POST /internal/v1/signup/prove|POST /internal/v1/signup/slug"
CORE_ENROLLMENT_ROUTES="GET /internal/v1/enrollment/policy|GET /internal/v1/enrollment/public|GET /internal/v1/enrollment/requests|GET /internal/v1/enrollment/status|POST /internal/v1/enrollment/abandon|POST /internal/v1/enrollment/admit|POST /internal/v1/enrollment/policy|POST /internal/v1/enrollment/request|POST /internal/v1/enrollment/requests/{id}/decide|POST /internal/v1/enrollment/requests/{id}/revoke"
# The seven G020 label method/path pairs on six paths (Plan r2 §4.2: the definitions list, create, update and delete,
# the Group page's assignment list, assign and unassign), mounted by main.go through the labels handler's Register;
# C-sorted; pinned for B-MCP-SOURCE.
CORE_LABELS_ROUTES="GET /internal/v1/groups/{id}/labels|GET /internal/v1/labels|POST /internal/v1/groups/{id}/labels/assign|POST /internal/v1/groups/{id}/labels/unassign|POST /internal/v1/labels|POST /internal/v1/labels/{id}/delete|POST /internal/v1/labels/{id}/update"
# The exact SDK package closure the server links (go-sdk v1.8.0 and what it requires; G014 S0): no other schema
# validator, transport or SDK package.
MCP_SDK_CLOSURE="github.com/google/jsonschema-go/jsonschema github.com/modelcontextprotocol/go-sdk/auth github.com/modelcontextprotocol/go-sdk/internal/authutil github.com/modelcontextprotocol/go-sdk/internal/json github.com/modelcontextprotocol/go-sdk/internal/jsonrpc2 github.com/modelcontextprotocol/go-sdk/internal/mcpgodebug github.com/modelcontextprotocol/go-sdk/internal/util github.com/modelcontextprotocol/go-sdk/internal/xcontext github.com/modelcontextprotocol/go-sdk/jsonrpc github.com/modelcontextprotocol/go-sdk/mcp github.com/modelcontextprotocol/go-sdk/oauthex github.com/segmentio/asm/ascii github.com/segmentio/asm/base64 github.com/segmentio/asm/cpu github.com/segmentio/asm/cpu/arm github.com/segmentio/asm/cpu/arm64 github.com/segmentio/asm/cpu/cpuid github.com/segmentio/asm/cpu/x86 github.com/segmentio/asm/internal/unsafebytes github.com/segmentio/asm/keyset github.com/segmentio/encoding/ascii github.com/segmentio/encoding/iso8601 github.com/segmentio/encoding/json github.com/yosida95/uritemplate/v3 golang.org/x/oauth2 golang.org/x/oauth2/internal"

# The intelligence-disabled variant (off suite; derived from the maintained repository, since the generator that once
# rendered it is retired): a copy of the repository without these intelligence-only roots, with the reviewed opt-out
# edits of infra/validation/intelligence-off.tsv applied (off_derive).
OFF_EDITS="$REPO/infra/validation/intelligence-off.tsv"
OFF_REMOVED=(intelligence contracts/intelligence contracts/conformance/intelligence infra/docker/intelligence.Dockerfile
  infra/docker/intelligence.Dockerfile.dockerignore)

# make_env_file DEST PROJECT: .env-format file from .env.example with unique secrets.
make_env_file() {
  local line
  while IFS= read -r line || [[ -n "$line" ]]; do
    if [[ "$line" =~ ^([A-Z0-9_]+)=$ && "${BASH_REMATCH[1]}" != OPENROUTER_API_KEY ]]; then
      local v; v="$(random_hex 24)"; printf '%s\n' "$v" >>"$SECRETS_FILE"
      printf '%s=%s\n' "${BASH_REMATCH[1]}" "$v"
    elif [[ "$line" =~ ^COMPOSE_PROJECT_NAME= ]]; then
      printf 'COMPOSE_PROJECT_NAME=%s\n' "$2"
    else
      printf '%s\n' "$line"
    fi
  done <"$REPO/.env.example" >"$1"
  chmod 600 "$1"
}

save_logs() { compose "$1" logs --no-color >"$WORK/compose-logs-$2.txt" 2>&1 || true; }

# ===================================================================================
# G051 (Plan r3 §3.1): the static checks regrouped by input boundary, each function and assertion unchanged: lint (the
# infrastructure checks), lint-base (repository-wide), lint-web (the web tree) and lint-core (the core module).
# G064 C1 P1: the lint containers go through docker_run_owned (formerly `bash -c "docker run ..."` strings).
lint_shellcheck() {
  local ci=(infra/ci/validate.sh infra/ci/lib.sh infra/ci/junit.sh infra/ci/validate-controls.sh infra/ci/cleanup-owned.sh infra/ci/runner-preflight.sh
    infra/ci/select-profile.sh infra/ci/hosted-context.sh infra/ci/pipeline-union.sh infra/ci/release-verify.sh infra/ci/staging-action.sh)
  # G072: the staging CI adapter, lease/metadata library, publication filter, federation bootstrap and their fixtures.
  docker_run_owned shellcheck --rm --label g001.run=$RUN_ID -v "$REPO:/repo:ro" -w /repo "$SHELLCHECK_IMAGE" -S warning -x -P SCRIPTDIR \
    infra/local/configure-env.sh infra/validation/run.sh infra/validation/lib.sh infra/security/secret-scan.sh infra/security/dependency-scan.sh \
    infra/recovery/recovery.sh infra/recovery/lib.sh \
    infra/staging/staging.sh infra/staging/lib.sh infra/staging/check.sh infra/staging/remote/deploy.sh infra/staging/remote/lib.sh \
    infra/staging/remote/reset.sh infra/staging/remote/start.sh infra/staging/remote/status.sh infra/staging/remote/stop.sh \
    infra/staging/ci.sh infra/staging/ci-lib.sh infra/staging/meta-lib.sh infra/staging/ci-publish.sh \
    infra/staging/tests/ci-adapter.sh infra/staging/tests/env-lease.sh infra/staging/tests/ci-federation.sh \
    infra/staging/tests/ci-launcher.sh infra/staging/tests/ci-build-boundary.sh \
    infra/terraform/bootstrap/bootstrap.sh infra/terraform/bootstrap/ci-federation.sh infra/terraform/modules/gcp-staging-vm/startup.sh \
    infra/validation/run-approved.sh infra/validation/run-approved-controls.sh "${ci[@]}" &&
    docker_run_owned shellcheck --rm --label g001.run=$RUN_ID -v "$REPO:/repo:ro" -w /repo "$SHELLCHECK_IMAGE" -S info --include=SC2251 -x -P SCRIPTDIR "${ci[@]}"
}
lint_hadolint() {
  local f
  for f in migrate web-browser validation-tools ci-context edge staging-ci; do
    docker_run_owned hadolint --rm --label g001.run=$RUN_ID -i "$HADOLINT_IMAGE" hadolint --failure-threshold warning - <"$REPO/infra/docker/$f.Dockerfile" || return 1
  done
}
suite_lint() {
  log "== lint"
  check B-SHELL "shellcheck (severity warning) on the project-owned scripts, including infra/ci (with select-profile.sh, the release-verify.sh helper, the POSIX hosted-context.sh adapter and the POSIX staging-action.sh launcher) and the staging tooling (infra/staging, its remote scripts, the G072 CI adapter, meta-lib.sh, ci-publish.sh and their fixtures, the Terraform bootstrap and CI federation scripts and VM startup scripts); and shellcheck --include=SC2251 on infra/ci (a !-negated statement is ignored by set -e)" \
    lint_shellcheck
  check B-HADOLINT "hadolint (failure threshold warning, as for the archetype's Dockerfiles) on infra/docker/migrate.Dockerfile, infra/docker/web-browser.Dockerfile, infra/docker/validation-tools.Dockerfile, infra/docker/ci-context.Dockerfile, infra/docker/edge.Dockerfile and infra/docker/staging-ci.Dockerfile" \
    lint_hadolint
  static_compose
  static_env_example
  static_deps
  static_migration_set
  static_dockerignore
  static_ci_config
  static_secret_source
}
suite_lint_base() {
  log "== lint-base"
  static_literals
  static_naming
  static_identity_guards
  static_depscan
}
suite_lint_web() {
  log "== lint-web"
  static_web_boundary
  static_web_tokens
  static_web_components
  static_web_design
  static_web_asset_controls
}
suite_lint_core() {
  log "== lint-core"
  static_mcp_source
}

# The static environment file and project of static_compose (lint) and static_mcp_source (lint-core), created once by
# whichever of the two suites runs first (G051: either may run alone).
static_env() {
  [[ -f "$WORK/static.env" ]] && return 0
  make_env_file "$WORK/static.env" "g001-static-$RUN_ID"
  register_project "g001-static-$RUN_ID"
}

# The three control matrices (full profile only; G003 CI revision Plan r2 §2.1).
suite_controls() {
  log "== controls"
  static_deps_controls
  check B-WRAPPER-CONTROLS "infra/validation/run-approved-controls.sh: W1-W9 against a stub harness in a scratch tree (digest, entry, count and path refusals never run the harness; harness failure survives a clean postcheck; drift fails a passing harness)" \
    bash -c "bash '$REPO/infra/validation/run-approved-controls.sh' >'$WORK/wrapper-controls.log' 2>&1 || { tail -30 '$WORK/wrapper-controls.log'; exit 1; }"
}

suite_ci_controls() {
  log "== ci-controls"
  static_ci_controls
}

# B-UNIT-CHECKS (fast and integration; G003 CI revision Plan r2 §2.1): the README platform:check block from
# the actual repository without a stack, with exactly one --env-file .env per line replaced by a generated
# environment file of a run-labelled project. B-CHECKS-ROOT keeps the same commands in full.
suite_unit() {
  log "== unit (stack-free; README platform:check)"
  local env="$WORK/unit.env" project="g001-unit-$RUN_ID" cmds line n=0 bad=0 rc=0 s
  make_env_file "$env" "$project"
  register_project "$project"
  cmds="$(docs_block "$REPO/README.md" platform:check)"
  if [[ -z "$cmds" ]]; then record B-UNIT-CHECKS FAIL "README platform:check block missing"; return; fi
  : >"$WORK/unit.cmds"
  while IFS= read -r line; do
    [[ -n "$line" ]] || continue
    n=$((n + 1))
    if [[ "$(grep -o -- '--env-file \.env' <<<"$line" | wc -l)" != 1 ]]; then bad=1; break; fi
    printf '%s\n' "${line/--env-file .env/--env-file $env}" >>"$WORK/unit.cmds"
  done <<<"$cmds"
  if (( bad || n == 0 )); then record B-UNIT-CHECKS FAIL "README platform:check: every line must carry exactly one --env-file .env"; return; fi
  # G051 build ledger: the block's Compose builds are accounted from its output, before the environment file goes.
  # G051 C4: web-check is built here; a running web-browser prebuild ends first (see PREBUILD_WEB).
  if (( PREBUILD_WEB )); then bg_wait_optional prebuild-web-browser; fi
  ledger_watch_begin "$REPO" "$project" unit
  # G064 D2: the documented block is latch-checked before (refused: 125, nothing run) and recounted after.
  if res_gate "documented block unit"; then
    (cd "$REPO" && bash -euo pipefail "$WORK/unit.cmds") >"$WORK/unit-check.log" 2>&1 || rc=$?
    res_recount "documented block unit" || true
    res_project_networks "$project" remove # G064: the block's `compose run` networks (captured by ID, verified removed)
  else rc=125; printf 'G064 documented block refused (ran nothing): %s\n' "$RES_WHY" >"$WORK/unit-check.log"; fi
  ledger_watch_end "$REPO" "$project" unit "$WORK/unit-check.log" "$(cat -- "$WORK/unit.cmds")"
  rm -f -- "$env" # it holds this run's generated secrets (still listed for B-SECRET-CANARY)
  for s in core-check web-check intelligence-check; do
    if docker image inspect "$project-$s:latest" >/dev/null 2>&1; then record_built_image "$project-$s:latest"; fi
  done
  if (( rc == 0 )); then
    record B-UNIT-CHECKS PASS "README check block without a stack ($n commands, generated environment): core-check (gofmt, vet, unit tests), web-check, intelligence-check, secret-scan"
  else tail -40 "$WORK/unit-check.log"; record B-UNIT-CHECKS FAIL "README check block (exit $rc)"; fi
}

# G051 B-UNIT-WEB, B-UNIT-CORE, B-UNIT-INTEL (unit-web/core/intel; Plan r3 §3.4): exactly that service's line of the
# README platform:check block and its secret-scan line, an exact subset of B-UNIT-CHECKS, run as suite_unit runs them
# (one --env-file .env per line replaced by a generated environment file of a run-labelled project; no stack).
suite_unit_web() { log "== unit-web (stack-free; README platform:check web-check, secret-scan)"; unit_service B-UNIT-WEB web; }
suite_unit_core() { log "== unit-core (stack-free; README platform:check core-check, secret-scan)"; unit_service B-UNIT-CORE core; }
suite_unit_intel() { log "== unit-intel (stack-free; README platform:check intelligence-check, secret-scan)"; unit_service B-UNIT-INTEL intelligence; }
unit_service() { # unit_service ID SERVICE
  local id="$1" svc="$2" env="$WORK/unit-$2.env" project="g001-unit-$2-$RUN_ID" cmds line n=0 nsvc=0 nscan=0 bad=0 rc=0
  cmds="$(docs_block "$REPO/README.md" platform:check)"
  if [[ -z "$cmds" ]]; then record "$id" FAIL "README platform:check block missing"; return; fi
  : >"$WORK/unit-$svc.cmds"
  while IFS= read -r line; do
    [[ -n "$line" ]] || continue
    if [[ "$line" == *" $svc-check" ]]; then nsvc=$((nsvc + 1))
    elif [[ "$line" == *" secret-scan" ]]; then nscan=$((nscan + 1))
    else continue; fi
    n=$((n + 1))
    if [[ "$(grep -o -- '--env-file \.env' <<<"$line" | wc -l)" != 1 ]]; then bad=1; break; fi
    printf '%s\n' "${line/--env-file .env/--env-file $env}" >>"$WORK/unit-$svc.cmds"
  done <<<"$cmds"
  if (( bad || nsvc != 1 || nscan != 1 )); then
    record "$id" FAIL "README platform:check: exactly one $svc-check line and one secret-scan line, each with exactly one --env-file .env"; return
  fi
  make_env_file "$env" "$project"
  register_project "$project"
  # G051 C4: web-check is built here; a running web-browser prebuild ends first (see PREBUILD_WEB).
  if [[ "$svc" == web ]] && (( PREBUILD_WEB )); then bg_wait_optional prebuild-web-browser; fi
  ledger_watch_begin "$REPO" "$project" "unit-$svc"
  if res_gate "documented block unit-$svc"; then
    (cd "$REPO" && bash -euo pipefail "$WORK/unit-$svc.cmds") >"$WORK/unit-$svc-check.log" 2>&1 || rc=$?
    res_recount "documented block unit-$svc" || true
    res_project_networks "$project" remove # G064: the block's `compose run` networks (captured by ID, verified removed)
  else rc=125; printf 'G064 documented block refused (ran nothing): %s\n' "$RES_WHY" >"$WORK/unit-$svc-check.log"; fi
  ledger_watch_end "$REPO" "$project" "unit-$svc" "$WORK/unit-$svc-check.log" "$(cat -- "$WORK/unit-$svc.cmds")"
  rm -f -- "$env" # it holds this run's generated secrets (still listed for B-SECRET-CANARY)
  if docker image inspect "$project-$svc-check:latest" >/dev/null 2>&1; then record_built_image "$project-$svc-check:latest"; fi
  if (( rc == 0 )); then
    record "$id" PASS "README check block subset without a stack ($n commands, generated environment): $svc-check and secret-scan"
  else tail -40 "$WORK/unit-$svc-check.log"; record "$id" FAIL "README check block subset $svc-check, secret-scan (exit $rc)"; fi
}

# G012 B-SECRET-SOURCE (Plan r1 §4.3): the submitted-source scan (infra/security/secret-scan.sh and the overlay
# infra/compose/secret-scan.yml) on disposable Git repositories under $WORK, with the real pinned gitleaks image and no
# network for the scanner. Canaries are generated credential-shaped assignments, registered in SECRETS_FILE and never
# printed (gitleaks --redact). S0 overlay configuration; S1 clean tree with a Vestive-shaped directory link and an
# outside-target link (never followed); S2/S3 tracked and untracked canaries; S4 force-added exempt paths; S5 an ignored
# local .env; S6 an untracked nested repository (skipped) and a gitlink (refused before any directory skip); S7
# incomplete streams (altered regular/link counts) and an invalid configuration; S8 the exact public-digest line
# exceptions of gitleaks.toml (MAINT secret-scan ADR-007, and the G051 O1 credential.go line) stay exact.
ss_canary() { # ss_canary FILE: append one generated credential-shaped assignment to FILE
  # The value always matches the pinned default rule github-pat (ghp_ and 36 of [0-9a-zA-Z], entropy 3): the body is
  # nine groups of three letters and one digit, all 36 characters distinct (Shannon entropy log2 36, about 5.17), so no
  # global allowlist regex or stopword of the pinned configuration (a five-letter run such as false, the alphabet, a
  # trailing null, an anchored placeholder) can apply. Chance picks only which characters and their order (unique per run),
  # never whether the scanner detects it (a G019 S3 miss on a hosted pipeline).
  local r c=ghp_ i j t l=({A..Z} {a..z}) n=({0..9})
  r="$(random_hex 120)"
  [[ ${#r} == 240 ]] || { log "ss_canary: random input unavailable"; return 1; }
  for ((i = 51; i > 0; i--)); do j=$((0x${r:4*(51-i):4} % (i + 1))); t=${l[i]}; l[i]=${l[j]}; l[j]=$t; done
  for ((i = 9; i > 0; i--)); do j=$((0x${r:4*(60-i):4} % (i + 1))); t=${n[i]}; n[i]=${n[j]}; n[j]=$t; done
  for ((i = 0; i < 9; i++)); do c+="${l[3*i]}${l[3*i+1]}${l[3*i+2]}${n[i]}"; done
  printf '%s\n' "$c" >>"$SECRETS_FILE"
  printf 'deploy_api_%s = "%s"\n' key "$c" >>"$1"
}
ss_git() { git -C "$1" -c user.name=g012-fixture -c user.email=fixture@invalid -c commit.gpgsign=false "${@:2}"; }
ss_repo() { # ss_repo DIR: a clean repository-like fixture with the real scanner configuration and a .gitignore
  mkdir -p "$1/infra/security" "$1/docs" "$1/web/.claude/skills/x" "$1/.agents/skills"
  git -c init.defaultBranch=main init -q "$1"
  cp "$REPO/infra/security/gitleaks.toml" "$1/infra/security/gitleaks.toml"
  printf '.env\n.env.local\n' >"$1/.gitignore"
  printf 'Fixture readme for the submitted-source scan.\n' >"$1/README.md"
  printf 'Plain guide text.\n' >"$1/docs/guide.md"
  printf 'Skill text.\n' >"$1/web/.claude/skills/x/SKILL.md"
  ln -s ../../web/.claude/skills/x "$1/.agents/skills/x"
  ss_git "$1" add -A && ss_git "$1" commit -q -m fixture
}
ss_scan() { # ss_scan ROOT LOG -> exit code of the scan
  local rc=0
  bash "$REPO/infra/security/secret-scan.sh" --env-file "$WORK/static.env" --root "$1" secret-scan >"$2" 2>&1 || rc=$?
  res_project_networks "g001-static-$RUN_ID" # G064: its `compose run` network, captured by ID
  return "$rc"
}
static_secret_source() {
  local d="$WORK/secret-source" ok=1 rc f cfg="$WORK/secret-scan-config.json" img base fx w
  static_env
  rm -rf "$d"; mkdir -p "$d"
  # S0: the overlay resets the bind mount and keeps the pinned image, no network, no ports, no environment.
  if (cd "$REPO" && docker compose --env-file "$WORK/static.env" -f infra/compose/local.yml -f infra/compose/secret-scan.yml --profile checks config --format json) >"$cfg" 2>"$d/s0.err"; then
    img="$(cd "$REPO" && docker compose --env-file "$WORK/static.env" -f infra/compose/local.yml --profile checks config --format json | yq_stdin -p json -o tsv '.services["secret-scan"].image')"
    [[ "$(yq_stdin -p json -o tsv '.services["secret-scan"] | [(.volumes // [] | length), .network_mode, (.ports // [] | length), (.environment // {} | length), .image] | join(" ")' <"$cfg")" == "0 none 0 0 $img" && -n "$img" ]] || { ok=0; log "S0: overlay secret-scan must have no volume, network none, no port, no environment and local.yml's image"; }
  else ok=0; log "S0: local.yml + secret-scan.yml config invalid"; fi
  rm -f "$cfg"
  # S1: clean tree with two tracked links; the outside target holds a canary and must never be read.
  fx="$d/s1"; ss_repo "$fx"
  mkdir -p "$d/s1-outside"; ss_canary "$d/s1-outside/held.txt"
  ln -s ../s1-outside/held.txt "$fx/outside-link"; ss_git "$fx" add outside-link && ss_git "$fx" commit -q -m link
  rc=0; ss_scan "$fx" "$d/s1.log" || rc=$?
  [[ $rc == 0 ]] && grep -q 'submitted set of .*: 5 regular files, 2 links, 0 skipped' "$d/s1.log" && grep -q 'accounting ok: 5 regular files, 2 links' "$d/s1.log" ||
    { ok=0; log "S1: clean tree with two links: exit $rc, want 0 with 5 regular files and 2 links"; }
  # S2: canary in a tracked file.
  fx="$d/s2"; ss_repo "$fx"; ss_canary "$fx/docs/deploy.md"; ss_git "$fx" add docs/deploy.md && ss_git "$fx" commit -q -m canary
  rc=0; ss_scan "$fx" "$d/s2.log" || rc=$?
  [[ $rc == 1 ]] && grep -q 'docs/deploy.md' "$d/s2.log" || { ok=0; log "S2: tracked canary: exit $rc, want 1 naming docs/deploy.md"; }
  # S3: canary in an untracked, non-ignored file.
  fx="$d/s3"; ss_repo "$fx"; ss_canary "$fx/docs/draft.md"
  rc=0; ss_scan "$fx" "$d/s3.log" || rc=$?
  [[ $rc == 1 ]] && grep -q 'docs/draft.md' "$d/s3.log" || { ok=0; log "S3: untracked canary: exit $rc, want 1"; }
  # S4: a force-added .env, and a force-added sub/.env.local, each refused before scanning.
  for f in .env sub/.env.local; do
    fx="$d/s4-${f//[\/.]/_}"; ss_repo "$fx"; mkdir -p "$fx/$(dirname "$f")"; ss_canary "$fx/$f"
    ss_git "$fx" add -f "$f" && ss_git "$fx" commit -q -m exempt
    rc=0; ss_scan "$fx" "$d/s4.log" || rc=$?
    [[ $rc == 3 ]] && grep -qF "matches a local-only exemption: $f" "$d/s4.log" && ! grep -q 'accounting ok' "$d/s4.log" ||
      { ok=0; log "S4: submitted $f: exit $rc, want 3 before scanning"; }
  done
  # S5: an ignored, untracked .env holding a canary is not submitted.
  fx="$d/s5"; ss_repo "$fx"; ss_canary "$fx/.env"
  rc=0; ss_scan "$fx" "$d/s5.log" || rc=$?
  [[ $rc == 0 ]] && grep -q 'accounting ok: 5 regular files, 1 links' "$d/s5.log" || { ok=0; log "S5: ignored .env: exit $rc, want 0 with the 5 tracked regular files only"; }
  # S6: an untracked nested repository (skipped, never descended), then a gitlink entry (refused first).
  fx="$d/s6"; ss_repo "$fx"; ss_repo "$fx/nested"; ss_canary "$fx/nested/docs/inner.md"
  rc=0; ss_scan "$fx" "$d/s6.log" || rc=$?
  [[ $rc == 0 ]] && grep -q 'skipped untracked nested repository: nested' "$d/s6.log" && grep -q 'accounting ok: 5 regular files, 1 links' "$d/s6.log" ||
    { ok=0; log "S6: nested repository: exit $rc, want 0 with nested skipped"; }
  ss_git "$fx" update-index --add --cacheinfo "160000,$(ss_git "$fx/nested" rev-parse HEAD),nested"
  rc=0; ss_scan "$fx" "$d/s6g.log" || rc=$?
  [[ $rc == 2 ]] && grep -q 'gitlink (submodule) entry is not supported: nested' "$d/s6g.log" || { ok=0; log "S6: gitlink: exit $rc, want 2"; }
  # S7: incomplete streams straight into the overlay (altered declared counts), and an invalid configuration.
  w="$d/s7"; mkdir -p "$w/scan/src/docs" "$w/meta"; printf 'Plain text.\n' >"$w/scan/src/docs/a.md"; ln -s docs/a.md "$w/scan/src/l"
  : >"$w/scan/link-targets.txt"; cp "$REPO/infra/security/gitleaks.toml" "$w/meta/gitleaks.toml"
  for f in "2 1" "1 0" "1 1"; do
    printf '%s\n' "${f% *}" >"$w/meta/count-regular"; printf '%s\n' "${f#* }" >"$w/meta/count-links"
    rc=0; res_compose_begin --profile checks run --rm -T secret-scan 2>"$d/s7.log" || rc=125
    if (( rc == 0 )); then tar -C "$w" -cf - scan meta | (cd "$REPO" && docker compose --env-file "$WORK/static.env" -f infra/compose/local.yml -f infra/compose/secret-scan.yml --profile checks run --rm -T secret-scan) >"$d/s7.log" 2>&1 || rc=$?; res_compose_end; res_project_networks "g001-static-$RUN_ID"; fi
    case "$f" in
      "1 1") [[ $rc == 0 ]] || { ok=0; log "S7: the correctly declared stream must pass (exit $rc)"; } ;;
      *) [[ $rc == 2 ]] && grep -q 'accounting mismatch' "$d/s7.log" || { ok=0; log "S7: declared counts $f: exit $rc, want 2"; } ;;
    esac
  done
  fx="$d/s7c"; ss_repo "$fx"; printf '\n[[rules]]\nid = "g012-broken"\nregex = """(unclosed"""\n' >>"$fx/infra/security/gitleaks.toml"
  ss_git "$fx" commit -q -am broken
  rc=0; ss_scan "$fx" "$d/s7c.log" || rc=$?
  [[ $rc == 2 ]] || { ok=0; log "S7: invalid configuration: exit $rc, want 2"; }
  # S7: a local-only exemption pattern in RE2-only syntax (the guard evaluates ERE) is refused, never skipped.
  fx="$d/s7r"; ss_repo "$fx"; ss_canary "$fx/.env.production"; ss_git "$fx" add .env.production
  sed -i "/^paths = \[/a\\  '''(?i)\\\\.env\\\\.production\$'''," "$fx/infra/security/gitleaks.toml"
  ss_git "$fx" commit -q -am re2
  rc=0; ss_scan "$fx" "$d/s7r.log" || rc=$?
  [[ $rc == 2 ]] && grep -q 'RE2-only syntax' "$d/s7r.log" && grep -qF "'''(?i)" "$fx/infra/security/gitleaks.toml" ||
    { ok=0; log "S7: an RE2-only exemption pattern: exit $rc, want 2"; }
  # S8 (MAINT secret-scan ADR-007): the two exact public-digest line exceptions stay exact. Each fixture is a tracked
  # docs/inputs.json; a finding is identified by its redacted fingerprint (file:rule:line), never by a value. Keys and
  # digests are joined only at run time, so this source holds no scannable key/digest pair.
  local k7 h7 ko ho l7 g i x rc2 s8c=() s8l=()
  k7=docs/adr/ADR-007-verified-domain-routing-and-authentication.md
  h7=1e9513ab070aefc4ec7afac8ca44d8be184a0e7831f3837f85bce53ee90e495b
  ko=core/internal/identity/credential.go
  ho=12dd3d07db3e8b9a787822725b66fbe6f073a342cdb8af8bb896cb81c675e638
  printf -v l7 '"%s": "%s"' "$k7" "$h7"
  # s8_found LOG: the redacted finding fingerprints of a scan log, space separated; empty, never an abort under
  # set -euo pipefail, when the log holds none.
  s8_found() { if grep -q '^Fingerprint:' "$1"; then grep '^Fingerprint:' "$1" | sed 's|.*/scan/src/||' | tr '\n' ' '; fi; }
  # S8a: the exact ADR-007 line bare (with and without the trailing comma and indentation) and as one inline-code citation.
  fx="$d/s8a"; ss_repo "$fx"
  printf '{\n    %s,\n    %s\n%s,\n%s\n`%s`\n}\n' "$l7" "$l7" "$l7" "$l7" "$l7" >"$fx/docs/inputs.json"
  ss_git "$fx" add docs/inputs.json && ss_git "$fx" commit -q -m s8a
  rc=0; ss_scan "$fx" "$d/s8a.log" || rc=$?
  [[ $rc == 0 ]] && ! grep -q '^Fingerprint:' "$d/s8a.log" ||
    { ok=0; log "S8a: the exact ADR-007 line (bare, with and without comma and indentation; one backtick pair): exit $rc, want 0"; }
  # S8 no-finding path: s8_found returns success with an empty list for a scan log without fingerprints.
  rc2=0; x="$(s8_found "$d/s8a.log")" || rc2=$?
  [[ $rc2 == 0 && -z "$x" ]] || { ok=0; log "S8: a log without findings: s8_found exit $rc2, findings [$x], want 0 and none"; }
  # S8e generic-api-key-only canary: an ss_canary body without its ghp_ prefix (github-pat cannot match), registered in
  # SECRETS_FILE; a body holding a pinned generic-api-key stopword that fits its shape (md5, sha1, sha2, vpn) is drawn again.
  g=""
  for i in 1 2 3 4 5 6 7 8; do
    : >"$d/s8-gen"; ss_canary "$d/s8-gen" || break
    g="$(sed -n 's/.*"ghp_\([0-9A-Za-z]*\)"$/\1/p' "$d/s8-gen")"
    if grep -qiE 'md5|sha[12]|vpn' <<<"$g"; then g=""; else break; fi
  done
  rm -f -- "$d/s8-gen"
  if [[ ${#g} == 36 ]]; then printf '%s\n' "$g" >>"$SECRETS_FILE"; else ok=0; log "S8e: generic-api-key canary unavailable"; fi
  # S8b-S8e: each fixture holds the exempt line (line 2) and one variant (line 3); only line 3 may be reported.
  s8c+=(before);  printf -v x -- '- %s,' "$l7"; s8l+=("$x")
  s8c+=(after);   printf -v x '    %s, "x": 1' "$l7"; s8l+=("$x")
  s8c+=(open);    printf -v x '`%s' "$l7"; s8l+=("$x")
  s8c+=(close);   printf -v x '%s`' "$l7"; s8l+=("$x")
  s8c+=(double);  printf -v x '``%s``' "$l7"; s8l+=("$x")
  s8c+=(key);     printf -v x '    "%s": "%s",' "${k7%.md}.txt" "$h7"; s8l+=("$x")
  s8c+=(digest);  printf -v x '    "%s": "%s",' "$k7" "${h7%?}c"; s8l+=("$x")
  s8c+=(other);   printf -v x '    "%s": "%s",' docs/security/token-rotation.md "$h7"; s8l+=("$x")
  s8c+=(generic); printf -v x '    "deploy_api_%s": "%s"' key "$g"; s8l+=("$x")
  for i in "${!s8c[@]}"; do
    fx="$d/s8-${s8c[i]}"; ss_repo "$fx"
    printf '{\n    %s,\n%s\n}\n' "$l7" "${s8l[i]}" >"$fx/docs/inputs.json"
    ss_git "$fx" add docs/inputs.json && ss_git "$fx" commit -q -m "s8 ${s8c[i]}"
    rc=0; ss_scan "$fx" "$d/s8-${s8c[i]}.log" || rc=$?
    x="$(s8_found "$d/s8-${s8c[i]}.log")"
    [[ $rc == 1 && "$x" == "docs/inputs.json:generic-api-key:3 " ]] ||
      { ok=0; log "S8 ${s8c[i]}: exit $rc, findings [$x], want 1 with only docs/inputs.json:generic-api-key:3"; }
  done
  # S8e: a github-pat canary on the line after the exempt line.
  fx="$d/s8-pat"; ss_repo "$fx"; printf '{\n    %s,\n' "$l7" >"$fx/docs/inputs.json"; ss_canary "$fx/docs/inputs.json"
  ss_git "$fx" add docs/inputs.json && ss_git "$fx" commit -q -m "s8 pat"
  rc=0; ss_scan "$fx" "$d/s8-pat.log" || rc=$?
  x="$(s8_found "$d/s8-pat.log")"
  [[ $rc == 1 && "$x" == "docs/inputs.json:github-pat:3 " ]] || { ok=0; log "S8e pat: exit $rc, findings [$x], want 1 with only docs/inputs.json:github-pat:3"; }
  # S8f: the G051 O1 credential.go line is still exempt; an altered digest is reported.
  fx="$d/s8f"; ss_repo "$fx"; printf '{\n    "%s": "%s",\n}\n' "$ko" "$ho" >"$fx/docs/inputs.json"
  ss_git "$fx" add docs/inputs.json && ss_git "$fx" commit -q -m s8f
  rc=0; ss_scan "$fx" "$d/s8f.log" || rc=$?
  [[ $rc == 0 ]] && ! grep -q '^Fingerprint:' "$d/s8f.log" || { ok=0; log "S8f: the O1 credential.go line: exit $rc, want 0"; }
  fx="$d/s8f-digest"; ss_repo "$fx"; printf '{\n    "%s": "%s",\n}\n' "$ko" "${ho%?}f" >"$fx/docs/inputs.json"
  ss_git "$fx" add docs/inputs.json && ss_git "$fx" commit -q -m s8f-digest
  rc=0; ss_scan "$fx" "$d/s8f-digest.log" || rc=$?
  x="$(s8_found "$d/s8f-digest.log")"
  [[ $rc == 1 && "$x" == "docs/inputs.json:generic-api-key:2 " ]] || { ok=0; log "S8f: an altered credential.go digest: exit $rc, findings [$x], want 1"; }
  if grep -rqF -f "$SECRETS_FILE" "$d"/*.log; then ok=0; log "a canary value reached the scan output"; fi
  # The fixture trees hold the canaries by design: keep only the logs, remove the trees (B-SECRET-CANARY scans $WORK).
  mkdir -p "$WORK/secret-source-logs" && mv -- "$d"/*.log "$WORK/secret-source-logs/" && rm -rf -- "$d" || { ok=0; log "fixture cleanup failed"; }
  [[ $ok == 1 ]] && record B-SECRET-SOURCE PASS "submitted-source scan on disposable repositories (real gitleaks, network none): S0 overlay (no volume, network none, no port, no environment, local.yml's image); S1 clean with a directory link and an outside-target link (exit 0, 5 regular files, 2 links, outside canary never read); S2 tracked and S3 untracked canaries exit 1; S4 force-added .env and sub/.env.local exit 3 before scanning; S5 ignored .env not submitted (exit 0); S6 nested repository skipped, gitlink refused (exit 2); S7 altered regular and link counts, an invalid configuration and an RE2-only exemption pattern exit 2; S8 the exact ADR-007 public-digest line (bare with and without comma and indentation, one backtick pair) exempt (exit 0), and next to it extra text before or after, an unpaired opening or closing backtick, a double backtick pair, a changed key, a changed digest, the digest under another key, a github-pat canary and a generic-api-key-only canary each exit 1 with only that line reported; the G051 O1 credential.go line exempt and its altered digest exit 1; no canary in any output" ||
    record B-SECRET-SOURCE FAIL "submitted-source scan"
}

# G012 B-DEPSCAN (Plan r1 §5.6; D0 dispositions D1/D3): infra/security/dependency-scan.sh in one check with both
# parts required: the fixed offline fixtures (govulncheck and osv-scanner semantics, vulnpolicy over their real
# outputs with the none/valid/expired exception files) and the current-tree scan of core/go.mod, web/package-lock.json
# and intelligence/uv.lock (identity reconciliation, blocking policy, exact exceptions). It downloads only the four
# fixed public advisory objects (plus Go modules for govulncheck, as every core build does); every scan runs with
# network none. Both volumes are registered before creation and removed by the script.
static_depscan() {
  local d="$WORK/depscan" adv="g012-advisory-$RUN_ID" goc="g012-gocache-$RUN_ID" rc=0 cid=() pins steps
  registry_row volume "$adv"
  registry_row volume "$goc"
  if [[ -n "$OWNED_REG" ]]; then cid=(--cid-dir "$(dirname "$OWNED_REG")/cid"); fi
  pins="$(sed -n -E 's/^(OSV_IMAGE|GO_IMAGE|GOVULNCHECK_VERSION|GO_STDLIB_VERSION)="(.*)"$/\1=\2/p' "$REPO/infra/security/dependency-scan.sh" | tr '\n' ' ')"
  record DEPSCAN-PINS INFO "$pins"
  bash "$REPO/infra/security/dependency-scan.sh" --out "$d" --advisory-volume "$adv" --gocache-volume "$goc" \
    --label "g001.run=$RUN_ID" "${cid[@]}" scan >"$WORK/depscan.log" 2>&1 || rc=$?
  if [[ -n "${REPORT_DIR:-}" ]]; then
    cp "$WORK/depscan.log" "$REPORT_DIR/depscan.log" 2>/dev/null || true
    cp "$d/results.tsv" "$REPORT_DIR/depscan-results.tsv" 2>/dev/null || true
    cp "$d/current-report.json" "$REPORT_DIR/depscan-current-report.json" 2>/dev/null || true
  fi
  steps="$(awk -F'\t' '{printf "%s %s; ", $1, $2}' "$d/results.tsv" 2>/dev/null)"
  if (( rc == 0 )) && [[ "$(awk -F'\t' '$2=="PASS"' "$d/results.tsv" | wc -l)" == 6 ]]; then
    record B-DEPSCAN PASS "offline fixtures and current tree (${steps% }): $(awk -F'\t' '$1=="current"{print $3}' "$d/results.tsv"); $(grep -E '^  GRAPH ' "$WORK/depscan.log" | sed -E 's/^  GRAPH ([A-Za-z]+) identities=([0-9]+) locations=([0-9]+).*/\1 \2 identities (\3 locations)/' | tr '\n' ',' | sed 's/,$//'); $(grep -cE '^  FINDING excepted ' "$WORK/depscan.log") excepted finding(s) kept visible"
  else
    grep -E 'dependency-scan: (PASS|FAIL)|^  (FINDING|PROBLEM|INCOMPLETE|RESULT|CHECK)' "$WORK/depscan.log" | tail -40
    record B-DEPSCAN FAIL "dependency scan exit $rc: ${steps:-no step result}$(depscan_fail_detail "$d/results.tsv")"
  fi
}
# depscan_fail_detail RESULTS: the failing steps' own assertion text (exit codes and expectation names only; never a
# value), control characters removed and capped at 800 characters, so a hosted failure is diagnosable from the
# published summary. Empty when no step failed or the file is missing.
depscan_fail_detail() {
  local t
  t="$(awk -F'\t' '$2 != "PASS" {printf "%s: %s | ", $1, $3}' "$1" 2>/dev/null | tr -d '\000-\037' | cut -c1-800)"
  [[ -z "$t" ]] || printf ' -- failing steps: %s' "${t% | }"
}

static_compose() {
  local env="$WORK/static.env" ok=1 cfg="$WORK/compose-config.json" keys line svc key value f
  static_env
  (cd "$REPO" && docker compose --env-file "$env" -f infra/compose/local.yml config -q) || { ok=0; log "local.yml config invalid"; }
  (cd "$REPO" && docker compose --env-file "$env" -f infra/compose/local.yml -f infra/compose/migrate.yml config -q) || { ok=0; log "local.yml + migrate.yml config invalid"; }
  (cd "$REPO" && docker compose --env-file "$env" -f infra/compose/local.yml -f infra/compose/migrate.yml --profile checks --profile migrate config --format json) >"$cfg" || ok=0
  keys="$(yq_stdin -p json -o tsv '.services.migrate.environment | keys | sort | join(" ")' <"$cfg")"
  [[ "$keys" == "APP_ENV DATABASE_NAME DB_HOST DB_MIGRATOR_PASSWORD DB_PORT" ]] || { ok=0; log "migrate environment keys: $keys"; }
  [[ "$(yq_stdin -p json -o tsv '.services.migrate.networks | keys | join(" ")' <"$cfg")" == data ]] || { ok=0; log "migrate must join only the data network"; }
  [[ "$(yq_stdin -p json -o tsv '.services.migrate.ports // [] | length' <"$cfg")" == 0 ]] || { ok=0; log "migrate must publish no port"; }
  # Each secret value may appear only in the services that need it.
  yq_stdin -p json -o tsv '.services | to_entries | .[] | .key as $s | (.value.environment // {}) | to_entries | .[] | [$s, .key, .value]' <"$cfg" >"$WORK/service-env.tsv"
  local -A allowed=(
    [POSTGRES_SUPERUSER_PASSWORD]="postgres" [DB_MIGRATOR_PASSWORD]="postgres migrate"
    [DB_RUNTIME_PASSWORD]="postgres core" [DB_RETENTION_PASSWORD]="postgres"
    [REDIS_PASSWORD]="redis core" [MINIO_ROOT_PASSWORD]="minio minio-init"
    [OBJECT_STORE_SECRET_KEY]="minio-init core" [INTELLIGENCE_SERVICE_TOKEN]="core intelligence"
    # G070 (Plan §3.1): set only in the staging block's derived environment; static.env leaves them unset (skipped).
    [CF_API_TOKEN]="edge" [IDENTITY_WORKOS_API_KEY]="core")
  local secret_key secret_value
  for secret_key in "${!allowed[@]}"; do
    secret_value="$(sed -n "s/^$secret_key=//p" "$env")"
    [[ -n "$secret_value" ]] || continue
    while IFS=$'\t' read -r svc key value; do
      if [[ "$value" == *"$secret_value"* && " ${allowed[$secret_key]} " != *" $svc "* ]]; then ok=0; log "$secret_key reaches service $svc ($key)"; fi
    done <"$WORK/service-env.tsv"
  done
  if awk -F'\t' '$1=="intelligence"{print $2}' "$WORK/service-env.tsv" | grep -qE '^(DB_|REDIS_|OBJECT_STORE_|MINIO_|POSTGRES_)'; then ok=0; log "intelligence holds data-store settings"; fi
  # The browser overlay: a test-only runner in web's network namespace, without ports or secrets.
  local bcfg="$WORK/compose-browser.json"
  (cd "$REPO" && docker compose --env-file "$env" -f infra/compose/local.yml -f infra/compose/browser.yml --profile browser config -q) || { ok=0; log "local.yml + browser.yml config invalid"; }
  (cd "$REPO" && docker compose --env-file "$env" -f infra/compose/local.yml -f infra/compose/browser.yml --profile browser config --format json) >"$bcfg" || ok=0
  keys="$(yq_stdin -p json -o tsv '.services["web-browser"].environment | keys | sort | join(" ")' <"$bcfg")"
  [[ "$keys" == "EXPECTED_PLATFORM_BASE_DOMAIN EXPECTED_PRODUCT_NAME WEB_BASE_URL" ]] || { ok=0; log "web-browser environment keys: $keys"; }
  [[ "$(yq_stdin -p json -o tsv '.services["web-browser"].network_mode' <"$bcfg")" == "service:web" ]] || { ok=0; log "web-browser must share web's network namespace"; }
  [[ "$(yq_stdin -p json -o tsv '.services["web-browser"].ports // [] | length' <"$bcfg")" == 0 ]] || { ok=0; log "web-browser must publish no port"; }
  for f in "$cfg" "$bcfg"; do
    [[ "$(yq_stdin -p json -o tsv '.services.web.networks | keys | join(" ")' <"$f")" == edge ]] || { ok=0; log "web networks must remain exactly edge ($(basename "$f"))"; }
  done
  yq_stdin -p json -o tsv '.services["web-browser"].environment | to_entries | .[] | .value' <"$bcfg" >"$WORK/browser-env.txt"
  for secret_key in "${!allowed[@]}"; do
    secret_value="$(sed -n "s/^$secret_key=//p" "$env")"
    [[ -n "$secret_value" ]] && grep -qF -- "$secret_value" "$WORK/browser-env.txt" && { ok=0; log "$secret_key reaches web-browser"; }
  done
  # G070 (Plan §3.1): the public staging overlay over local.yml and migrate.yml, from a derived environment with
  # synthetic digest-pinned images, the edge settings and generated CF_API_TOKEN/IDENTITY_WORKOS_API_KEY values.
  local senv="$WORK/static-staging.env" scfg="$WORK/compose-staging.json" domain image v nets pinned=0
  domain="$(sed -n 's/^PLATFORM_BASE_DOMAIN=//p' "$env")"
  (umask 077; : >"$senv")
  {
    cat "$env"
    for svc in minio minio-init migrate core web intelligence edge; do
      v="${svc^^}"; printf 'IMAGE_%s=registry.example.invalid/%s@sha256:%s\n' "${v//-/_}" "$svc" "$(random_hex 32)"
    done
    printf 'PLATFORM_AUTH_HOST=auth.%s\nEDGE_BASE_DOMAIN=%s\n' "$domain" "$domain"
    printf 'ACME_EMAIL=ops@example.net\nEDGE_ACME_CA=https://acme-v02.api.letsencrypt.org/directory\n'
    for key in CF_API_TOKEN IDENTITY_WORKOS_API_KEY; do
      v="$(random_hex 24)"; printf '%s\n' "$v" >>"$SECRETS_FILE"
      printf '%s=%s\n' "$key" "$v"
    done
  } >>"$senv"
  (cd "$REPO" && docker compose --env-file "$senv" -f infra/compose/local.yml -f infra/compose/migrate.yml -f infra/compose/staging.yml --profile migrate config --format json) >"$scfg" || { ok=0; log "local.yml + migrate.yml + staging.yml config invalid"; }
  v="$(yq_stdin -p json -o tsv '.services | to_entries | .[] | select(.key != "edge" and ((.value.ports // []) | length) > 0) | .key' <"$scfg" | paste -sd' ')"
  [[ -z "$v" ]] || { ok=0; log "staging: only edge may publish ports (also: $v)"; }
  v="$(yq_stdin -p json -o tsv '.services.edge.ports // [] | .[] | [.published, .protocol]' <"$scfg" | tr '\t' '/' | LC_ALL=C sort | paste -sd' ')"
  [[ "$v" == "443/tcp 80/tcp" ]] || { ok=0; log "staging: edge must publish exactly 80/tcp and 443/tcp (got: $v)"; }
  while IFS=$'\t' read -r svc key image; do
    pinned=$((pinned + 1))
    [[ "$key" == false ]] || { ok=0; log "staging: $svc must not build"; }
    [[ "$image" =~ @sha256:[0-9a-f]{64}$ ]] || { ok=0; log "staging: $svc image is not digest-pinned ($image)"; }
  done < <(yq_stdin -p json -o tsv '.services | to_entries | .[] | select(.key | test("^(core|web|intelligence|minio|minio-init|migrate|edge)$")) | [.key, (.value | has("build")), (.value.image // "-")]' <"$scfg")
  [[ $pinned == 7 ]] || { ok=0; log "staging: $pinned of the 7 pinned services present"; }
  local -A staging_nets=([edge]="edge" [web]="edge" [core]="data edge inference" [intelligence]="inference" [migrate]="data")
  for svc in "${!staging_nets[@]}"; do
    nets="$(yq_stdin -p json -o tsv ".services[\"$svc\"].networks // {} | keys | sort | join(\" \")" <"$scfg")"
    [[ "$nets" == "${staging_nets[$svc]}" ]] || { ok=0; log "staging: $svc networks must be exactly ${staging_nets[$svc]} (got: $nets)"; }
  done
  v="$(yq_stdin -p json -o tsv '.services | keys | .[]' <"$scfg" | { grep -xE 'emulator|proxy|identity-setup|core-replica|web-browser|secret-scan' || true; } | paste -sd' ')"
  [[ -z "$v" ]] || { ok=0; log "staging: test-only services present: $v"; }
  yq_stdin -p json -o tsv '.services | to_entries | .[] | .key as $s | (.value.environment // {}) | to_entries | .[] | [$s, .key, .value]' <"$scfg" >"$WORK/staging-env.tsv"
  v="$(cut -f2 "$WORK/staging-env.tsv" | { grep -E '^(IDENTITY_TEST_|CUSTOM_DOMAIN_DNS_FIXTURE_FILE|DOMAIN_TLS_FIXTURE_FILE|MCP_CONNECT_JWKS_FILE|PLATFORM_HOST_PORT|IDENTITY_WORKOS_API_BASE_URL)' || true; } | LC_ALL=C sort -u | paste -sd' ')"
  [[ -z "$v" ]] || { ok=0; log "staging: test or fixture environment keys present: $v"; }
  v="$(awk -F'\t' '$2=="APP_ENV" && $3=="test"{print $1}' "$WORK/staging-env.tsv" | paste -sd' ')"
  [[ -z "$v" ]] || { ok=0; log "staging: APP_ENV=test in $v"; }
  for secret_key in "${!allowed[@]}"; do
    secret_value="$(sed -n "s/^$secret_key=//p" "$senv")"
    [[ -n "$secret_value" ]] || continue
    while IFS=$'\t' read -r svc key value; do
      if [[ "$value" == *"$secret_value"* && " ${allowed[$secret_key]} " != *" $svc "* ]]; then ok=0; log "staging: $secret_key reaches service $svc ($key)"; fi
    done <"$WORK/staging-env.tsv"
  done
  rm -f "$senv" "$scfg" "$WORK/staging-env.tsv" # they contain the derived environment's generated secrets
  rm -f "$cfg" "$bcfg" "$WORK/service-env.tsv" "$WORK/browser-env.txt" # they contain the static environment's generated secrets
  if [[ $ok == 1 ]]; then record B-COMPOSE PASS "local.yml, local.yml+migrate.yml and local.yml+browser.yml valid; migrate: data network only, no ports, env exactly APP_ENV DATABASE_NAME DB_HOST DB_MIGRATOR_PASSWORD DB_PORT; every secret only in its owning services; intelligence holds no data-store settings; web-browser: network_mode service:web, no ports, env exactly EXPECTED_PLATFORM_BASE_DOMAIN EXPECTED_PRODUCT_NAME WEB_BASE_URL, no secret value; web networks exactly edge; local.yml+migrate.yml+staging.yml valid with a derived static-staging.env: only edge publishes ports (exactly 80/tcp and 443/tcp), minio minio-init migrate core web intelligence edge never build and run sha256-digest images, networks edge=edge web=edge core=data+edge+inference intelligence=inference migrate=data, no test-only service, no test or fixture environment key, APP_ENV never test, CF_API_TOKEN only in edge and IDENTITY_WORKOS_API_KEY only in core"
  else record B-COMPOSE FAIL "compose configuration or secret scoping"; fi
}

# The rendered identity block of .env.example and the reviewed G004 block that replaces it at the end of the file
# (register: .env.example; G004 Plan r5 §4.6, §8.5). B-ENV-EXAMPLE accepts exactly the render with that substitution.
ENV_IDENTITY_MARKER='# --- Identity provider: WorkOS AuthKit (docs/foundation/IDENTITY.md) --------'
ENV_IDENTITY_BLOCK_SHA=3d55f30dedcae4654c9c098557acdfc5f575e8e22249c0f96e1c51793bf300c2
# G016 (Plan r4 §5.4): the reviewed domain block follows the identity block at the end of the file; the identity block
# runs from its marker to the domain marker, and the domain block from its marker to the end.
ENV_DOMAIN_MARKER='# --- Hostnames and custom domains (docs/foundation/DOMAINS.md) -------------'
# G062 (Plan r5 §2, §4.1, D-6): MARKETING_SITE_URL is removed, so the domain block drops its marketing sentence and
# commented variable (16 lines); both pins below were recomputed from the updated file with the same awk ranges.
# G071 (Plan r4 S11 (i), S13, D-10 (i)): the domain block gains the DEFAULT_TENANT_SLUG and PLATFORM_ROOT_PAGE comment
# lines (27 lines); both pins were recomputed again over the integrated bytes with the same awk ranges.
ENV_DOMAIN_BLOCK_SHA=bddf593873c4758c2ba7e5332684a15d08829cf3a78c02f3cec062ce8c170148
# The reviewed head of this template's .env.example (every line before the identity marker, with the configured
# identity values), pinned in place of the retired renderer manifest; the head plus the generator's former rendered
# identity block equals the .env.example hash that manifest recorded, so the guarantee is unchanged. A reviewed edit of
# the head (for example, new identity values) updates this pin.
ENV_HEAD_SHA=deeeed01af67b9ffce3424007fe567dcedbbb741567f795e737c9501d6ca1a42
# The reviewed intelligence dependency inputs, pinned in place of the retired manifest (the same hashes the manifest
# recorded).
INTEL_DEP_PINS=("intelligence/uv.lock 499b5cce611bdc667a2634f5d3aa599ed1a400003aee8f5ab43f93c9fe259401"
  "intelligence/pyproject.toml 70b329706c196b433550c9bf8aba8e1f3c30ad148d0ff60f08e5abb70f4c5ed9")

# The web routes (page.* and route.* under web/app; G051 O3): the routes outside app/(features)/ are exactly the
# reviewed inventory below. A route under app/(features)/<n>/ is allowed only when web area <n> is registered in
# infra/ci/profile-paths.tsv with exactly the root web/app/(features)/<n>/. The (features) namespace alone allows
# nothing: an unregistered area, a malformed name or path, or a missing or extra other route fails. A registration
# classifies validation only; it is no product authority. G062 (Plan r5 §4.4): +8, the root signup pages /signup and
# /signup/details and the route handlers /signup/{start,create,abandon,prove} (root) and /signup/{continue,complete}
# (tenant); /signup/abandon also answers GET (same file). G063 (Plan r1 §2): +1, the MCP setup form route
# /connections/mcp/rename. G020 (Plan r2 §2): +4, the Organization labels page /organization/labels and its form route
# /organization/labels/action, and the Group labels page /groups/[id]/labels and its form route /groups/[id]/labels/action.
WEB_ROUTES="account/page.tsx api/account/route.ts api/health/route.ts api/ready/route.ts auth/callback/route.ts auth/login/route.ts auth/logout-all/route.ts auth/logout-return/route.ts auth/logout/route.ts auth/provider-logout/route.ts branding/logo/route.ts business-profiles/[id]/page.tsx business-profiles/[id]/update/route.ts business-profiles/page.tsx connections/mcp/create/route.ts connections/mcp/page.tsx connections/mcp/rename/route.ts connections/mcp/revoke/route.ts groups/[id]/labels/action/route.ts groups/[id]/labels/page.tsx groups/[id]/members/action/route.ts groups/[id]/members/page.tsx groups/[id]/onboarding/complete/route.ts groups/[id]/onboarding/page.tsx groups/[id]/page.tsx groups/address-check/route.ts groups/create/route.ts groups/new/page.tsx groups/page.tsx invitations/[id]/accept/route.ts invitations/[id]/page.tsx invitations/page.tsx join/abandon/route.ts join/admit/route.ts join/invitations/[id]/page.tsx join/page.tsx join/request/route.ts join/start/route.ts join/status/page.tsx organization/branding/logo/remove/route.ts organization/branding/logo/route.ts organization/branding/objects/delete/route.ts organization/branding/page.tsx organization/branding/reset/route.ts organization/branding/update/route.ts organization/enrollment/action/route.ts organization/enrollment/page.tsx organization/group-authority/action/route.ts organization/group-authority/page.tsx organization/labels/action/route.ts organization/labels/page.tsx organization/members/action/route.ts organization/members/page.tsx page.tsx sign-in/page.tsx signup/abandon/route.ts signup/complete/route.ts signup/continue/route.ts signup/create/route.ts signup/details/page.tsx signup/page.tsx signup/prove/route.ts signup/start/route.ts "
web_routes_ok() { # WEB_DIR MAP
  local w="$1" map="$2" r n c k m rc=0 other=""
  local re='^\(features\)/([a-z][a-z0-9-]{0,31})/([]A-Za-z0-9_.@()[-]+/)*(page|route)\.[a-z]+$'
  local -A reg=()
  [[ -d "$w/app" && ! -L "$w/app" ]] || { log "no web/app directory"; return 1; }
  if [[ -f "$map" && ! -L "$map" ]]; then
    while IFS=$'\t' read -r c k m; do
      if [[ "$c" == area-web && "$k" == dir && "$m" =~ ^web/app/\(features\)/([a-z][a-z0-9-]{0,31})/$ ]]; then reg[${BASH_REMATCH[1]}]=1; fi
    done <"$map"
  fi
  while IFS= read -r -d '' r; do
    if [[ "$r" != "(features)/"* ]]; then other+="$r"$'\n'; continue; fi
    if [[ ! "$r" =~ $re || "/$r/" == *"/../"* || "/$r/" == *"/./"* ]]; then
      log "malformed feature route app/$r"; rc=1; continue
    fi
    n="${BASH_REMATCH[1]}"
    if [[ -z "${reg[$n]+x}" ]]; then
      log "feature route app/$r: web area $n is not registered with root web/app/(features)/$n/ in infra/ci/profile-paths.tsv"; rc=1
    fi
  done < <(cd "$w/app" && find . -type f \( -name 'page.*' -o -name 'route.*' \) -printf '%P\0')
  if [[ "$(printf '%s' "$other" | LC_ALL=C sort | tr '\n' ' ')" != "$WEB_ROUTES" ]]; then
    log "routes outside app/(features)/ differ from /, /api/health, /api/ready, the G004 identity pages and routes, the G014 MCP setup page and routes (G063: rename) the G017 logo route and branding management page and routes and the G018 onboarding pages and routes (with the M1 advisory /groups/address-check) and the G062 signup pages and routes and the G020 label pages and routes"; rc=1
  fi
  return "$rc"
}

static_gitignore() {
  local f
  [[ -f "$REPO/.gitignore" && ! -L "$REPO/.gitignore" ]] || { log ".gitignore missing or symlinked"; return 1; }
  for f in .env .env.local; do
    git --no-optional-locks -c core.fsmonitor=false -C "$REPO" check-ignore --no-index -q -- "$f" || {
      log ".gitignore does not effectively exclude $f"; return 1;
    }
  done
}

static_env_example() {
  local ok=1 f="$REPO/.env.example" n
  static_gitignore || ok=0
  n="$(grep -cxF -- "$ENV_IDENTITY_MARKER" "$f" || true)"
  if [[ "$n" != 1 ]]; then ok=0; log ".env.example must hold the reviewed identity block exactly once (found $n)"
  else
    [[ "$(awk -v m="$ENV_IDENTITY_MARKER" -v d="$ENV_DOMAIN_MARKER" '$0 == d {exit} $0 == m {f = 1} f' "$f" | sha256sum | cut -d' ' -f1)" == "$ENV_IDENTITY_BLOCK_SHA" ]] || { ok=0; log ".env.example identity block differs from the reviewed block"; }
    [[ "$(grep -cxF -- "$ENV_DOMAIN_MARKER" "$f")" == 1 && "$(awk -v m="$ENV_IDENTITY_MARKER" -v d="$ENV_DOMAIN_MARKER" '$0 == m {i = NR} $0 == d {n = NR} END {print (i > 0 && n > i)}' "$f")" == 1 ]] || { ok=0; log ".env.example must hold the reviewed domain block exactly once, after the identity block"; }
    [[ "$(awk -v d="$ENV_DOMAIN_MARKER" '$0 == d {f = 1} f' "$f" | sha256sum | cut -d' ' -f1)" == "$ENV_DOMAIN_BLOCK_SHA" ]] || { ok=0; log ".env.example domain block differs from the reviewed block"; }
    [[ "$(awk -v m="$ENV_IDENTITY_MARKER" '$0 == m {exit} {print}' "$f" | sha256sum | cut -d' ' -f1)" == "$ENV_HEAD_SHA" ]] || { ok=0; log ".env.example differs from the reviewed head before the identity block"; }
  fi
  if grep -E '^[A-Z0-9_]*(PASSWORD|SECRET|TOKEN|API_KEY)[A-Z0-9_]*=.+' "$f"; then ok=0; log "a secret has a value in .env.example"; fi
  if grep -rn 'AI_GATEWAY_DEFAULT' "$f" "$REPO/infra/compose" "$REPO/core" "$REPO/intelligence/src" >/dev/null 2>&1; then ok=0; log "AI_GATEWAY_DEFAULT present"; fi
  [[ $ok == 1 ]] && record B-ENV-EXAMPLE PASS ".env.example is exactly the reviewed head (pinned), then the reviewed G004 identity block (IDENTITY_PROVIDER=workos; every identity value commented), followed by exactly the reviewed G016 domain block (every value commented, none a secret); every secret empty; effective .env/.env.local Git exclusions; no AI_GATEWAY_DEFAULT" || record B-ENV-EXAMPLE FAIL ".env.example or Git secret exclusions"
}

# LITERAL_PATH_EXEMPT: the retired generator Skill's directory path, which the retirement check (SKILL_RETIRED) names
# and which equals the neutral project slug plus its parent and slashes. Only that exact, case-sensitive path component
# is not an identity value; any other occurrence of the literal in the same file, a near-miss path or a re-cased path
# is still a hit.
LITERAL_PATH_EXEMPT=skills/saas-foundation/
literal_outside_exempt() { # LITERAL FILE: 0 when FILE holds LITERAL (case-insensitive) outside every exact exempt path
  local line
  while IFS= read -r line; do
    line="${line//"$LITERAL_PATH_EXEMPT"/}"
    [[ "${line,,}" == *"${1,,}"* ]] && return 0
  done < <(grep -iF -- "$1" "$2" 2>/dev/null || true)
  return 1
}
# product_doc_path PATH: the project's own documentation and operations tooling (the root README.md, docs/foundation/
# and infra/staging/), where the configured product display name may appear. Application roots (web, core,
# intelligence, contracts) and the rest of infra must still take it from configuration.
product_doc_path() { case "$1" in README.md | docs/foundation/* | infra/staging/*) return 0 ;; esac; return 1; }
# literal_hits ROOT LITERAL PRODUCT_ONLY: the scanned files under ROOT holding LITERAL (case-insensitive) outside every
# exact LITERAL_PATH_EXEMPT occurrence, one per line; with PRODUCT_ONLY 1 (LITERAL is the product name and no other
# configured identity value) product_doc_path files are not hits.
literal_hits() {
  local f
  (cd "$1" && grep -rIliF -- "$2" web core intelligence contracts infra docs/foundation README.md .gitignore 2>/dev/null || true) |
    while IFS= read -r f; do
      [[ -n "$f" ]] || continue
      # The former skill-path exclusion is a no-op here (vestive_skill_path): the brand agent skill is not shipped.
      vestive_skill_path "$f" && continue
      [[ $3 == 1 ]] && product_doc_path "$f" && continue
      if literal_outside_exempt "$2" "$1/$f"; then printf '%s\n' "$f"; fi
    done
}

static_literals() {
  local ok=1 lit hits k v c="$WORK/literals-control" want got pname="" ponly=1
  local -a idv=()
  # The five configured identity values of the committed .env.example (formerly the manifest inputs), the product
  # name first.
  : >"$WORK/literals-raw.txt"
  for k in PRODUCT_NAME COMPOSE_PROJECT_NAME PLATFORM_BASE_DOMAIN CUSTOM_DOMAIN_CNAME_TARGET DATABASE_NAME; do
    v="$(sed -n "s/^$k=//p" "$REPO/.env.example" | tail -1 | sed 's/^"\(.*\)"$/\1/')"
    [[ -n "$v" ]] || { ok=0; log ".env.example has no value for $k"; }
    idv+=("$v")
    printf '%s\n' "$v" >>"$WORK/literals-raw.txt"
  done
  grep -v '^$' "$WORK/literals-raw.txt" | LC_ALL=C sort -u >"$WORK/literals.txt" || true
  # The product-name scope applies only to a readable name that equals no other configured value (fails closed).
  pname="${idv[0]}"
  [[ -n "$pname" ]] || ponly=0
  for v in "${idv[@]:1}"; do [[ "${v,,}" == "${pname,,}" ]] && ponly=0; done
  while IFS= read -r lit; do
    [[ -n "$lit" ]] || continue
    if [[ "$lit" == "$pname" ]]; then hits="$(literal_hits "$REPO" "$lit" "$ponly")"; else hits="$(literal_hits "$REPO" "$lit" 0)"; fi
    [[ -z "$hits" ]] || { ok=0; log "identity literal found in: ${hits//$'\n'/ }"; }
  done <"$WORK/literals.txt"
  rm -rf "$c"; mkdir -p "$c/x/infra" "$c/p/web/app" "$c/p/core" "$c/p/infra/ci" "$c/p/infra/staging" "$c/p/docs/foundation"
  # Negative control of the path exemption: a stray literal beside the exact path, a near-miss path and a re-cased
  # path are hits; a file holding the literal only inside the exact path is not.
  lit="${LITERAL_PATH_EXEMPT#skills/}"; lit="${lit%/}"
  printf 'a=%sx\n' "$LITERAL_PATH_EXEMPT" >"$c/x/infra/exempt.sh"
  printf 'a=%sx b=%s\n' "$LITERAL_PATH_EXEMPT" "$lit" >"$c/x/infra/stray.sh"
  printf 'a=skills/%s-x/\n' "$lit" >"$c/x/infra/near.sh"
  printf 'a=SKILLS/%s/\n' "$lit" >"$c/x/infra/upper.sh"
  want=$'infra/near.sh\ninfra/stray.sh\ninfra/upper.sh'
  got="$(literal_hits "$c/x" "$lit" 0 | LC_ALL=C sort)"
  [[ "$got" == "$want" ]] || { ok=0; log "B-LITERALS path-exemption control: expected {${want//$'\n'/ }}, got {${got//$'\n'/ }}"; }
  # Negative control of the product-name scope, with a synthetic name: in web and core (re-cased) and in infra outside
  # infra/staging (a near-miss superstring) it is a hit; in README.md, docs/foundation and infra/staging it is not; and
  # when the name is also another configured value (PRODUCT_ONLY 0) every scanned root is a hit again.
  v=ExportCtlName
  printf 'export const n = "%s";\n' "$v" >"$c/p/web/app/page.tsx"
  printf '// %s\n' "${v,,}" >"$c/p/core/name.go"
  printf '# %ss\n' "$v" >"$c/p/infra/ci/near.sh"
  printf '# %s\n' "$v" >"$c/p/README.md"
  printf '%s\n' "$v" >"$c/p/docs/foundation/x.md"
  printf '# %s\n' "$v" >"$c/p/infra/staging/x.sh"
  want=$'core/name.go\ninfra/ci/near.sh\nweb/app/page.tsx'
  got="$(literal_hits "$c/p" "$v" 1 | LC_ALL=C sort)"
  [[ "$got" == "$want" ]] || { ok=0; log "B-LITERALS product-name control: expected {${want//$'\n'/ }}, got {${got//$'\n'/ }}"; }
  want=$'README.md\ncore/name.go\ndocs/foundation/x.md\ninfra/ci/near.sh\ninfra/staging/x.sh\nweb/app/page.tsx'
  got="$(literal_hits "$c/p" "$v" 0 | LC_ALL=C sort)"
  [[ "$got" == "$want" ]] || { ok=0; log "B-LITERALS product-name control (shared value): expected {${want//$'\n'/ }}, got {${got//$'\n'/ }}"; }
  rm -rf "$c"
  [[ $ok == 1 ]] && record B-LITERALS PASS "$(wc -l <"$WORK/literals.txt") identity values (the configured values of .env.example) absent from platform roots, README, docs/foundation, .gitignore and this harness, except inside the exact retired generator Skill path $LITERAL_PATH_EXEMPT; the product display name alone may also appear in the project's documentation and operations tooling (README.md, docs/foundation, infra/staging) (controls: a stray, near-miss or re-cased occurrence is a hit; the product name stays a hit in web, core and the rest of infra)" || record B-LITERALS FAIL "identity literal outside .env.example"
}

static_naming() {
  local ok=1 d
  for d in frontend backend api app services; do [[ -e "$REPO/$d" ]] && { ok=0; log "generic root $d"; }; done
  # web/public is allowed since G048; its contents are bounded by B-WEB-DESIGN's exact inventory.
  # G069 (Plan r1 S1; Spec §2 "prefer web/src"): web/src is the delimited root of the unaltered kit source, not a
  # redundant tier name; B-WEB-COMPONENTS bounds it to exactly the 69 manifest files.
  for d in core/internal/core core/cmd/core core/cmd/api web/web web/components/domain intelligence/intelligence intelligence/src/intelligence intelligence/app infra/infra contracts/contracts; do
    [[ -e "$REPO/$d" ]] && { ok=0; log "redundant path $d"; }
  done
  [[ $ok == 1 ]] && record B-NAMING PASS "responsibility-named roots; no generic tier roots or redundant paths; no web/components/domain (web/public bounded by B-WEB-DESIGN; the G069 kit source root web/src bounded by B-WEB-COMPONENTS)" || record B-NAMING FAIL "naming"
}

static_deps() {
  local ok=1 f bad
  grep -qE $'^\tgithub.com/pressly/goose/v3 v3\\.28\\.0$' "$REPO/core/go.mod" || { ok=0; log "goose v3.28.0 not a direct requirement"; }
  grep -qE $'^\tgithub.com/modelcontextprotocol/go-sdk v1\\.8\\.0$' "$REPO/core/go.mod" || { ok=0; log "MCP Go SDK v1.8.0 not a direct requirement"; }
  # G004 (Plan r5 §3.7): the WorkOS SDK and the token/key libraries, exact versions, as direct requirements.
  local m
  for m in 'github.com/workos/workos-go/v10 v10.5.0' 'github.com/golang-jwt/jwt/v5 v5.3.1' 'github.com/MicahParks/keyfunc/v3 v3.8.2' 'github.com/MicahParks/jwkset v0.11.3' 'golang.org/x/time v0.15.0' 'golang.org/x/text v0.43.0'; do
    grep -qxF $'\t'"$m" "$REPO/core/go.mod" || { ok=0; log "core/go.mod: $m is not a direct requirement"; }
  done
  # Every FROM that does not name an earlier stage must carry an explicit tag.
  bad="$(awk 'FNR==1{delete st} toupper($1)=="FROM"{ if (!($2 in st) && $2 !~ /:[A-Za-z0-9._-]+$/) print FILENAME ": " $2; if (toupper($3)=="AS") st[$4]=1 }' "$REPO"/infra/docker/*.Dockerfile)"
  bad+="$(grep -hE '^FROM [^ ]+:latest|image: [^ ]+:latest|image: [^ :]+$' "$REPO"/infra/docker/*.Dockerfile "$REPO"/infra/compose/*.yml || true)"
  [[ -z "$bad" ]] || { ok=0; log "unpinned image: $bad"; }
  for f in "${INTEL_DEP_PINS[@]}"; do
    [[ "$(sha_of "$REPO/${f%% *}")" == "${f##* }" ]] || { ok=0; log "${f%% *} differs from its reviewed pin"; }
  done
  # The validation-tools module (test tooling, G003 §13): exactly one direct requirement, go.sum present.
  f="$REPO/infra/validation/tools/go.mod"
  [[ -s "$REPO/infra/validation/tools/go.sum" ]] || { ok=0; log "infra/validation/tools/go.sum missing"; }
  [[ "$(sed -n '/^require (/,/^)/p; /^require [^(]/p' "$f" 2>/dev/null | grep -vE '^require \($|^\)$' | grep -v '// indirect' | sed -E 's/^require //; s/^[[:space:]]+//' | LC_ALL=C sort | tr '\n' ' ')" == "github.com/santhosh-tekuri/jsonschema/v6 v6.0.3 " ]] || { ok=0; log "validation tools: direct requirements are not exactly santhosh-tekuri/jsonschema/v6 v6.0.3"; }
  grep -qx 'module validationtools' "$f" 2>/dev/null || { ok=0; log "validation tools module name"; }
  # web: the reviewed dependency policy (independent, reason-coded rules; G002 plan §15).
  deps_repo_check "$REPO" >"$WORK/deps-codes.txt" || { ok=0; log "web dependency policy: $(tr '\n' ';' <"$WORK/deps-codes.txt")"; }
  [[ $ok == 1 ]] && record B-DEPS PASS "goose v3.28.0 and MCP Go SDK v1.8.0 pinned in core/go.mod, with the G004 WorkOS SDK v10.5.0, golang-jwt v5.3.1, keyfunc v3.8.2, jwkset v0.11.3, x/time v0.15.0 and x/text v0.43.0 (G019, unicode/norm) as direct requirements; every image tag explicit (no latest); intelligence uv.lock and pyproject.toml at their reviewed pins; validation tools: exactly santhosh-tekuri/jsonschema/v6 v6.0.3 with go.sum; web: the 12 existing and ${#DEPS_ADDITIONS[@]} reviewed direct dependencies exactly (Base UI @base-ui/react 1.9.0, G067; the G069 kit runtime packages and shadcn as a dev-only package, no cn), exact pins, lockfile root, versions and dev flags consistent, no radix-ui, no direct @radix-ui package, every @radix-ui lockfile entry at any depth one of the ${#DEPS_RADIX_VIA_CMDK[@]} reviewed cmdk-closure name@version entries with cmdk direct, no next/font, Playwright image tag equals @playwright/test with a full digest" || record B-DEPS FAIL "dependency pins"
}

# B-DEPS-CONTROLS: disposable mutations of copies of the real pair. Each control must fail with
# exactly its expected code set (G002 plan §15); no rule is relaxed to make a control pass.
static_deps_controls() {
  local C="$WORK/deps-controls" ok=1 pkg="$REPO/web/package.json" lock="$REPO/web/package-lock.json" rc got
  local reg='https://registry.npmjs.org'
  mkdir -p "$C"; : >"$WORK/deps-controls.tsv"
  # deps_expect ID WANT GOT RC: WANT and GOT are newline-separated code sets.
  deps_expect() {
    local id="$1" want="$2" got="$3" rc="$4" verdict=pass
    if [[ "$got" != "$want" ]] || { [[ -n "$want" ]] && [[ "$rc" == 0 ]]; } || { [[ -z "$want" ]] && [[ "$rc" != 0 ]]; }; then verdict=FAIL; ok=0; fi
    printf '%s\t%s\t%s\t%s\n' "$id" "${want//$'\n'/;}" "${got//$'\n'/;}" "$verdict" >>"$WORK/deps-controls.tsv"
    [[ $verdict == pass ]] || log "deps control $id: expected {${want//$'\n'/; }} (failing), got {${got//$'\n'/; }} exit $rc"
  }
  # deps_case ID WANT PKG_EXPR LOCK_EXPR: yq mutations of the real pair, checked by deps_pair_check.
  deps_case() {
    local d="$C/$1"; mkdir -p "$d"
    yq_stdin -p json -o json "$3" <"$pkg" >"$d/package.json"
    yq_stdin -p json -o json "$4" <"$lock" >"$d/package-lock.json"
    rc=0; got="$(deps_pair_check "$d/package.json" "$d/package-lock.json")" || rc=$?
    deps_expect "$1" "$2" "$got" "$rc"
  }
  deps_case C1 "DEPS_EXISTING_CHANGED next" '.dependencies.next = "16.3.9"' \
    ".packages[\"\"].dependencies.next = \"16.3.9\" | .packages[\"node_modules/next\"].version = \"16.3.9\" | .packages[\"node_modules/next\"].resolved = \"$reg/next/-/next-16.3.9.tgz\""
  deps_case C2 "DEPS_EXTRA_DIRECT left-pad" '.dependencies["left-pad"] = "1.3.0"' \
    ".packages[\"\"].dependencies[\"left-pad\"] = \"1.3.0\" | .packages[\"node_modules/left-pad\"] = {\"version\": \"1.3.0\", \"resolved\": \"$reg/left-pad/-/left-pad-1.3.0.tgz\"}"
  deps_case C3 "DEPS_ADDITION_WRONG @axe-core/playwright" \
    '.dependencies["@axe-core/playwright"] = .devDependencies["@axe-core/playwright"] | del(.devDependencies["@axe-core/playwright"])' \
    '.packages[""].dependencies["@axe-core/playwright"] = .packages[""].devDependencies["@axe-core/playwright"] | del(.packages[""].devDependencies["@axe-core/playwright"]) | del(.packages["node_modules/@axe-core/playwright"].dev) | del(.packages["node_modules/axe-core"].dev)'
  deps_case C4 "DEPS_EXISTING_MISSING server-only" 'del(.dependencies["server-only"])' \
    'del(.packages[""].dependencies["server-only"]) | del(.packages["node_modules/server-only"])'
  deps_case C5 $'DEPS_ADDITION_WRONG @base-ui/react\nDEPS_NOT_EXACT @base-ui/react' '.dependencies["@base-ui/react"] = "^1.9.0"' \
    '.packages[""].dependencies["@base-ui/react"] = "^1.9.0"'
  deps_case C6 "DEPS_LOCK_MISMATCH next" '.' \
    ".packages[\"node_modules/next\"].version = \"16.3.9\" | .packages[\"node_modules/next\"].resolved = \"$reg/next/-/next-16.3.9.tgz\""
  deps_case C7 "DEPS_LOCK_MISMATCH next" '.' '.packages[""].dependencies.next = "16.3.9"'
  # npm marks @playwright/test devOptional (it is also an optional peer of next): its dev flag is that marking.
  deps_case C8 "DEPS_LOCK_MISMATCH @playwright/test" '.' 'del(.packages["node_modules/@playwright/test"].dev) | del(.packages["node_modules/@playwright/test"].devOptional)'
  deps_case C8b "DEPS_LOCK_MISMATCH typescript" '.' 'del(.packages["node_modules/typescript"].dev)'
  # devOptional is dev-only only when a non-dev package declares the name as an optional peer.
  deps_case C8c "DEPS_LOCK_MISMATCH typescript" '.' 'del(.packages["node_modules/typescript"].dev) | .packages["node_modules/typescript"].devOptional = true'
  deps_case C9 "DEPS_EXISTING_RECATEGORIZED typescript" \
    '.dependencies.typescript = .devDependencies.typescript | del(.devDependencies.typescript)' \
    '.packages[""].dependencies.typescript = .packages[""].devDependencies.typescript | del(.packages[""].devDependencies.typescript) | del(.packages["node_modules/typescript"].dev)'
  # G069 (Plan r1 §3.1): shadcn is a reviewed dev-only addition; as a runtime dependency it is the wrong category.
  deps_case C10 "DEPS_ADDITION_WRONG shadcn" \
    '.dependencies.shadcn = .devDependencies.shadcn | del(.devDependencies.shadcn)' \
    '.packages[""].dependencies.shadcn = .packages[""].devDependencies.shadcn | del(.packages[""].devDependencies.shadcn) | del(.packages["node_modules/shadcn"].dev)'
  # G069 (Plan r1 §3.1, S3), retargeting G067's removed-Radix cases: a cmdk-closure package at its reviewed version as a
  # direct dependency, and an unlisted @radix-ui package nested in the lockfile; then the Radix umbrella and wrong
  # Base UI pins.
  local rdv="" ent
  for ent in "${DEPS_RADIX_VIA_CMDK[@]}"; do [[ "$ent" == "@radix-ui/react-dialog@"* ]] && rdv="${ent##*@}"; done
  deps_case C13 $'DEPS_EXTRA_DIRECT @radix-ui/react-dialog\nDEPS_RADIX_DIRECT @radix-ui/react-dialog' ".dependencies[\"@radix-ui/react-dialog\"] = \"$rdv\"" \
    ".packages[\"\"].dependencies[\"@radix-ui/react-dialog\"] = \"$rdv\" | .packages[\"node_modules/@radix-ui/react-dialog\"] = {\"version\": \"$rdv\", \"resolved\": \"$reg/@radix-ui/react-dialog/-/react-dialog-$rdv.tgz\"}"
  deps_case C14 "DEPS_RADIX_UNEXPECTED @radix-ui/react-popover" '.' \
    ".packages[\"node_modules/cmdk/node_modules/@radix-ui/react-popover\"] = {\"version\": \"1.1.15\", \"resolved\": \"$reg/@radix-ui/react-popover/-/react-popover-1.1.15.tgz\"}"
  deps_case C15 $'DEPS_EXTRA_DIRECT radix-ui\nDEPS_FORBIDDEN radix-ui' '.dependencies["radix-ui"] = "1.6.7"' \
    ".packages[\"\"].dependencies[\"radix-ui\"] = \"1.6.7\" | .packages[\"node_modules/radix-ui\"] = {\"version\": \"1.6.7\", \"resolved\": \"$reg/radix-ui/-/radix-ui-1.6.7.tgz\"}"
  deps_case C16 "DEPS_ADDITION_WRONG @base-ui/react" '.dependencies["@base-ui/react"] = "1.9.1"' \
    ".packages[\"\"].dependencies[\"@base-ui/react\"] = \"1.9.1\" | .packages[\"node_modules/@base-ui/react\"].version = \"1.9.1\" | .packages[\"node_modules/@base-ui/react\"].resolved = \"$reg/@base-ui/react/-/react-1.9.1.tgz\""
  deps_case C17 "DEPS_ADDITION_WRONG @base-ui/react" \
    '.devDependencies["@base-ui/react"] = .dependencies["@base-ui/react"] | del(.dependencies["@base-ui/react"])' \
    '.packages[""].devDependencies["@base-ui/react"] = .packages[""].dependencies["@base-ui/react"] | del(.packages[""].dependencies["@base-ui/react"]) | .packages["node_modules/@base-ui/react"].dev = true'
  deps_case C18 "DEPS_LOCK_MISMATCH @base-ui/react" '.' \
    ".packages[\"node_modules/@base-ui/react\"].version = \"1.8.0\" | .packages[\"node_modules/@base-ui/react\"].resolved = \"$reg/@base-ui/react/-/react-1.8.0.tgz\""
  # C19 (r2): an exact, consistent extra direct dependency isolates the extra-direct rule. G069: date-fns, a transitive-only
  # package of react-day-picker (tailwind-merge is now a reviewed addition); transitive-only packages never become direct.
  deps_case C19 "DEPS_EXTRA_DIRECT date-fns" '.dependencies["date-fns"] = "4.1.0"' \
    ".packages[\"\"].dependencies[\"date-fns\"] = \"4.1.0\" | .packages[\"node_modules/date-fns\"] = {\"version\": \"4.1.0\", \"resolved\": \"$reg/date-fns/-/date-fns-4.1.0.tgz\"}"
  # C11: the repository-level rules on a scratch copy of their inputs, with the image tag changed.
  local d="$C/C11"
  mkdir -p "$d"
  (cd "$REPO" && git ls-files -z --cached --others --exclude-standard -- web infra/docker/web-browser.Dockerfile |
    while IFS= read -r -d '' f; do [[ -e "$f" ]] && { vestive_skill_path "$f" || printf '%s\0' "$f"; }; done | tar --null -cf - -T -) | tar -xf - -C "$d"
  sed -i 's#^FROM mcr\.microsoft\.com/playwright:v1\.63\.0-#FROM mcr.microsoft.com/playwright:v1.62.0-#' "$d/infra/docker/web-browser.Dockerfile"
  rc=0; got="$(deps_repo_check "$d")" || rc=$?
  deps_expect C11 DEPS_IMAGE_MISMATCH "$got" "$rc"
  # C12: the real pair and the real repository-level inputs pass.
  rc=0; got="$(deps_repo_check "$REPO")" || rc=$?
  deps_expect C12 "" "$got" "$rc"
  [[ $ok == 1 ]] && record B-DEPS-CONTROLS PASS "C1-C11, C13-C19, C8b and C8c each fail with exactly their expected reason codes (no missing, no unexpected); C12 (the real inputs) passes; see deps-controls.tsv" || record B-DEPS-CONTROLS FAIL "a dependency control did not produce exactly its expected code set"
}

# code_lines ERE FILE...: matching lines, except lines that are entirely comment: a // line, a
# whole /* ... */ or {/* ... */} line, an opening /* line that does not close, a * continuation
# or a closing */ line. A line with code after a comment is still matched.
code_lines() {
  local re="$1"; shift
  [[ $# -gt 0 ]] || return 0
  grep -nHE -- "$re" "$@" 2>/dev/null |
    grep -vE '^[^:]+:[0-9]+:[[:space:]]*(//.*|/\*.*\*/|\{/\*.*\*/\}|/\*([^*]|\*[^/])*|\*([[:space:]].*)?|\*/)[[:space:]]*$' || true
}
# web_sources: web source files (git view; no node_modules, build output or lockfile), NUL-separated. The
# static assets under web/public (font binaries, licences, the drawing; G048) are not source: B-WEB-DESIGN
# bounds them by an exact pinned inventory.
web_sources() {
  (cd "$REPO" && git ls-files -z --cached --others --exclude-standard -- web | while IFS= read -r -d '' f; do
    [[ -f "$f" && "$f" != web/package-lock.json && "$f" != web/public/* ]] && { vestive_skill_path "$f" || printf '%s\0' "$REPO/$f"; }; done)
}

# The brand agent skill is not shipped in this template, so the former skill-path exclusion is a no-op: B-LITERALS and
# the web-application scans cover all of web/. (VESTIVE_SKILL stays defined for the inert relocation tables below.)
# The kit reference data replaces the skill's manifest and brand images: REF_KIT_MANIFEST holds exactly the 69 kit/src
# rows ("<sha256>  kit/src/<path>") and REF_BRAND_HASHES the content SHA-256 of each omitted brand image (one per line,
# no paths, no images). REF_KIT_MANIFEST is exempt from DESIGN_LITERAL only while its content SHA-256 equals
# REF_KIT_MANIFEST_SHA; any other file under web/tests/reference, and any mismatch, is scanned normally.
# ASSET_CONTROL_FIXTURE is the neutral test-only image that control A2 uses in place of a brand image.
VESTIVE_SKILL=.agents/skills/vestive-brand
vestive_skill_path() { return 1; }
REF_KIT_MANIFEST=web/tests/reference/kit-source.sha256 REF_BRAND_HASHES=web/tests/reference/brand-asset-hashes.txt
REF_KIT_MANIFEST_SHA=572b2b0a1ce4d1f0ea102051fe97232c0446a6c905fa2908814ca6adfb6bc293
ASSET_CONTROL_FIXTURE=web/tests/fixtures/asset-control/neutral.png
# G049 (cycle 2), retargeted by G067: DESIGN_LITERAL scans web/eslint.config.mjs without exactly its two reviewed G067
# lines, which name the excluded skill's path. Each must occur exactly once; otherwise the whole file is scanned and the
# status is 1.
G067_ESLINT_LINES=('  // G067: the adopted brand agent skill (design input, never web source) and its Claude discovery alias.'
  '  { ignores: [".claude/skills/vestive-brand/**", ".agents/skills/vestive-brand/**"] },')
eslint_literal_text() { # FILE
  local l st=0
  for l in "${G067_ESLINT_LINES[@]}"; do [[ "$(grep -cxF -- "$l" "$1")" == 1 ]] || st=1; done
  if (( st == 0 )); then grep -vxF -e "${G067_ESLINT_LINES[0]}" -e "${G067_ESLINT_LINES[1]}" -- "$1" || true; else cat -- "$1"; fi
  return "$st"
}

static_web_boundary() {
  local ok=1 W="$REPO/web" f hits n
  local -a client comps src
  mapfile -t client < <(grep -rlE '^"use client"' "$W/app" "$W/components" "$W/lib" 2>/dev/null)
  mapfile -t comps < <(find "$W/components" -type f)
  mapfile -d '' -t src < <(web_sources)
  hits="$(code_lines '@/lib/config|@/lib/core-client|server-only|process\.env' "${client[@]}" "${comps[@]}")"
  [[ -z "$hits" ]] || { ok=0; log "client or component code reaches server-only configuration: $hits"; }
  for f in lib/config.ts lib/core-client.ts lib/session.ts lib/identity-client.ts lib/mcp-client.ts; do grep -qxF 'import "server-only";' "$W/$f" || { ok=0; log "$f does not import server-only"; }; done
  # G069 (Plan r1 S19): the unaltered kit sidebar (web/src/components/ui/sidebar.tsx) writes its open/closed UI preference
  # to document.cookie (never a credential). Exactly that file is allowed, and no app-owned file (outside web/src) imports
  # it, so that code is unreachable.
  local -a nokit appsrc
  for f in "${src[@]}"; do
    [[ "$f" == "$W/src/components/ui/sidebar.tsx" ]] || nokit+=("$f")
    [[ "$f" == "$W/src/"* ]] || appsrc+=("$f")
  done
  hits="$(code_lines 'NEXT_PUBLIC_' "${src[@]}")$(code_lines 'document\.cookie' "${nokit[@]}")"
  [[ -z "$hits" ]] || { ok=0; log "browser-exposed configuration or cookies: $hits"; }
  hits="$(code_lines "[\"'\`](@/components/ui/sidebar|([.]{1,2}/)+src/components/ui/sidebar)[\"'\`]" "${appsrc[@]}")"
  [[ -z "$hits" ]] || { ok=0; log "app-owned file imports the kit sidebar: $hits"; }
  # G004: lib/session.ts is the only place that may handle cookies through next/headers.
  hits="$(for f in "${src[@]}"; do [[ "$f" == "$W/lib/session.ts" ]] || printf '%s\0' "$f"; done | xargs -0 -r grep -nHE 'cookies\(\)' 2>/dev/null | grep -vE '^[^:]+:[0-9]+:[[:space:]]*//' || true)"
  [[ -z "$hits" ]] || { ok=0; log "cookies() outside lib/session.ts: $hits"; }
  local pages; mapfile -t pages < <(find "$W/app" -name '*.tsx' -type f)
  hits="$(code_lines 'fetch\([[:space:]]*["'\''`]https?:' "${comps[@]}" "${pages[@]}")"
  [[ -z "$hits" ]] || { ok=0; log "absolute fetch in components or pages: $hits"; }
  web_routes_ok "$W" "$REPO/infra/ci/profile-paths.tsv" || ok=0
  # G069 (Plan r1 S5, §3.4): next-themes is the one appearance manager. No app code touches browser storage (the provider
  # keeps the preference under storageKey); the only exception is the S19 kit sidebar file above.
  local -a prodsrc
  for f in "${nokit[@]}"; do [[ "$f" == "$W/app/"* || "$f" == "$W/components/"* || "$f" == "$W/lib/"* || "$f" == "$W/src/"* ]] && prodsrc+=("$f"); done
  hits="$(code_lines 'localStorage|sessionStorage|document\.cookie|indexedDB' "${prodsrc[@]}")"
  [[ -z "$hits" ]] || { ok=0; log "app code references browser storage: $hits"; }
  # Scope (Plan r1 §3.4 "the only file naming next-themes outside web/src"): app-owned production code (web/app,
  # components, lib). Tests are oracles, not a second manager: S-HEAD-ORDER renders next-themes' own script with the
  # S5 props to compare the served one (Plan r1 §6.2, boundary.spec.ts).
  local tp="$W/components/layout/theme-provider.tsx"
  local -a appprod
  for f in "${appsrc[@]}"; do [[ "$f" == "$W/app/"* || "$f" == "$W/components/"* || "$f" == "$W/lib/"* ]] && appprod+=("$f"); done
  hits="$(code_lines "(from|import|require)[[:space:]]*\(?[[:space:]]*[\"'\`]next-themes[/\"'\`]" "${appprod[@]}" | cut -d: -f1 | LC_ALL=C sort -u)"
  [[ "$hits" == "$tp" ]] || { ok=0; log "next-themes must be named in app code (app, components, lib) only by components/layout/theme-provider.tsx (found: ${hits//$'\n'/ })"; }
  for f in 'storageKey="appearance-mode"' 'attribute="class"' 'defaultTheme="system"'; do
    grep -qF -- "$f" "$tp" 2>/dev/null || { ok=0; log "components/layout/theme-provider.tsx does not declare $f"; }
  done
  n="$(cat "${src[@]}" | grep -o 'suppressHydrationWarning' | wc -l)"
  [[ "$n" == 1 ]] && grep -qE '^[[:space:]]*<html lang="en" suppressHydrationWarning>$' "$W/app/layout.tsx" || { ok=0; log "suppressHydrationWarning must occur exactly once, on <html> in app/layout.tsx (found $n)"; }
  ! grep -qE 'next/script|next/font' "$W/app/layout.tsx" || { ok=0; log "layout.tsx imports next/script or next/font"; }
  # G003: request context is server-only and used only by the health routes and the core client,
  # which forwards exactly outboundHeaders(ctx); the error-probe fixtures stay test-only.
  grep -qxF 'import "server-only";' "$W/lib/request-context.ts" || { ok=0; log "lib/request-context.ts does not import server-only"; }
  # G004: also the identity routes and pages and the identity client; G014: the MCP setup page, its two routes and the
  # MCP client; G016: the server-only host helper (lib/host.ts: the tenant-host gate and callback relay, exactly
  # outboundHeaders and no credential), used by the proxy gate and the auth routes. G017: the server-only branding
  # client (lib/branding.ts), the tenant-frame loader (components/branding/tenant-frame.tsx), the logo proxy route and
  # the branding management form routes. G018: the server-only onboarding client (lib/onboarding-client.ts) and the
  # three onboarding form routes and the M1 advisory GET /groups/address-check (exact files; the onboarding pages and
  # components do not import it). G019: the server-only membership client (lib/membership-client.ts) and the three
  # membership form routes (Organization and Group member actions, invitation acceptance; the pages and components do not).
  # G062 (Plan r5 §4.4): the server-only signup client (lib/signup-client.ts), the six signup route handlers
  # (/signup/{start,create,abandon,prove} on the root, /signup/{continue,complete} on a tenant) and the root signup frame
  # components/signup/landing.tsx (its rootPage request facts; exact files; the signup pages and the other signup
  # components do not import it). G063 (Plan r1 §2): the MCP rename form route and the shared account components
  # components/account/connection-manager.tsx and account-sections.tsx (the RequestContext type of the manager's one list
  # call; exact files). G020 (Plan r2 §2): the server-only labels client lib/labels-client.ts (exact file; the label pages,
  # form routes and components do not import it).
  hits="$(grep -rlE 'request-context' "$W/app" "$W/components" "$W/lib" 2>/dev/null | sed "s|^$W/||" | grep -vxE 'app/api/(health|ready|account)/route\.ts|app/auth/(login|callback|logout|logout-all|logout-return|provider-logout)/route\.ts|app/(account|sign-in|connections/mcp)/page\.tsx|app/connections/mcp/(create|revoke|rename)/route\.ts|app/branding/logo/route\.ts|app/organization/branding/(update|reset|logo|logo/remove|objects/delete)/route\.ts|app/groups/(create|address-check)/route\.ts|app/groups/\[id\]/onboarding/complete/route\.ts|app/business-profiles/\[id\]/update/route\.ts|app/organization/members/action/route\.ts|app/groups/\[id\]/members/action/route\.ts|app/invitations/\[id\]/accept/route\.ts|app/signup/(start|create|abandon|continue|prove|complete)/route\.ts|components/branding/tenant-frame\.tsx|components/signup/landing\.tsx|components/account/(connection-manager|account-sections)\.tsx|lib/(core-client|identity-client|mcp-client|request-context|host|branding|onboarding-client|membership-client|enrollment-client|signup-client|labels-client)\.ts' || true)"
  [[ -z "$hits" ]] || { ok=0; log "request-context imported outside the health, identity and MCP setup routes and pages, the core, identity and MCP clients, the G016 host helper and the G017 branding client, tenant frame and routes and the G018 onboarding client, form routes and advisory address-check route and the G062 signup client, route handlers and root frame and the G020 labels client: $hits"; }
  grep -qE 'headers:[[:space:]]*outboundHeaders\(ctx\)' "$W/lib/core-client.ts" || { ok=0; log "core-client.ts must send exactly outboundHeaders(ctx)"; }
  # G004: the identity client sends exactly outboundHeaders(ctx) plus Authorization, X-Login-Binding and Content-Type,
  # built in one place and passed to its one fetch.
  local ic="$W/lib/identity-client.ts"
  [[ "$(grep -cF 'const headers: Record<string, string> = { ...outboundHeaders(ctx) };' "$ic")" == 1 ]] || { ok=0; log "identity-client.ts must start its headers from exactly outboundHeaders(ctx)"; }
  [[ "$(grep -oE 'headers(\.[A-Za-z]+|\["[^"]+"\]) =' "$ic" | sed -E 's/ =$//; s/^headers\.?//; s/^\["//; s/"\]$//' | LC_ALL=C sort | tr '\n' ' ')" == "Authorization Content-Type X-Login-Binding " ]] || { ok=0; log "identity-client.ts may add only Authorization, X-Login-Binding and Content-Type"; }
  [[ "$(grep -c 'fetch(' "$ic")" == 1 ]] && grep -qE '^[[:space:]]+headers,$' "$ic" || { ok=0; log "identity-client.ts must make one fetch with exactly those headers"; }
  [[ "$(cd "$W" && find . -path ./node_modules -prune -o -path ./.next -prune -o -path '*error-probe*' -type f -printf '%P\n' | LC_ALL=C sort | tr '\n' ' ')" == "tests/fixtures/error-probe/control-route.ts tests/fixtures/error-probe/wrapped-route.ts " ]] || { ok=0; log "error-probe fixtures must exist only under web/tests/fixtures/error-probe"; }
  hits="$(grep -rlE 'error-probe' "$W/app" "$W/components" "$W/lib" 2>/dev/null || true)"
  [[ -z "$hits" ]] || { ok=0; log "error-probe fixtures imported by production code: $hits"; }
  # G069 (Plan r1 S5): layout.tsx renders no <script> itself; exactly one <ThemeProvider> element, inside <body>, renders
  # next-themes' classic inline theme script.
  [[ "$(grep -cE '<script\b' "$W/app/layout.tsx")" == 0 ]] || { ok=0; log "layout.tsx must render no <script> element (the ThemeProvider renders the theme script)"; }
  [[ "$(grep -oE '<ThemeProvider([[:space:]>/]|$)' "$W/app/layout.tsx" | wc -l)" == 1 && "$(awk '/<body([[:space:]>]|$)/ {f = 1} f && /<ThemeProvider([[:space:]>\/]|$)/ {n++} /<\/body>/ {f = 0} END {print n + 0}' "$W/app/layout.tsx")" == 1 ]] || { ok=0; log "layout.tsx must render exactly one <ThemeProvider> element, inside <body>"; }
  [[ $ok == 1 ]] && record B-WEB-BOUNDARY PASS "client and component code imports no server configuration, server-only or process.env; lib/config.ts, lib/core-client.ts, lib/session.ts, lib/identity-client.ts and lib/mcp-client.ts stay server-only; no NEXT_PUBLIC_; no document.cookie except the unreachable G069 kit sidebar file (no app-owned file imports @/components/ui/sidebar); cookies() only in lib/session.ts; no absolute fetch in components or pages; routes exactly /, /api/health, /api/ready, the G004 /sign-in, /account, /api/account and six /auth routes (with the C7 /auth/logout-return) and the G014 /connections/mcp page with its create and revoke routes and the G063 rename route, the G017 /branding/logo route and the /organization/branding management page with its five form routes, and the G018 /groups, /groups/new, /groups/[id], /groups/[id]/onboarding and /business-profiles, /business-profiles/[id] pages with the /groups/create, /groups/[id]/onboarding/complete and /business-profiles/[id]/update form routes and the M1 advisory GET /groups/address-check, and the G019 /organization/members, /groups/[id]/members, /invitations and /invitations/[id] pages with their three form routes, and the G060 /organization/group-authority page with its form route, and the G062 /signup and /signup/details pages with the /signup/start, /signup/create, /signup/abandon, /signup/prove, /signup/continue and /signup/complete routes; no localStorage, sessionStorage, document.cookie or indexedDB in app code (web/app, components, lib, src) except that kit file; next-themes imported in app code (outside web/src and tests) only by components/layout/theme-provider.tsx, which declares storageKey=\"appearance-mode\", attribute=\"class\" and defaultTheme=\"system\"; suppressHydrationWarning once, on <html>; layout.tsx: no next/script or next/font, no <script> element, exactly one <ThemeProvider> inside <body> (G069); request-context server-only, used only by the health, identity and MCP setup routes and pages, the core, identity and MCP clients and the G016 server-only host helper lib/host.ts, and the G017 branding client lib/branding.ts, tenant frame, logo route and management form routes, and the G018 onboarding client lib/onboarding-client.ts and its three form routes /groups/create, /groups/[id]/onboarding/complete and /business-profiles/[id]/update and the M1 advisory route /groups/address-check, and the G019 membership client lib/membership-client.ts and its three form routes, and the G062 signup client lib/signup-client.ts, its six route handlers and the root frame components/signup/landing.tsx, and the G063 MCP rename route and the shared account components connection-manager.tsx and account-sections.tsx (core client exactly outboundHeaders; identity client exactly outboundHeaders plus Authorization, X-Login-Binding and Content-Type); error-probe fixtures test-only" || record B-WEB-BOUNDARY FAIL "web boundary"
}

# G069 (Plan r1 §3.3): the kit's own web/src/styles/theme.css is the palette source. web/app/globals.css imports the
# kit chain unchanged as its first statement and declares, as its only colour values, the app-owned accessibility
# variable corrections (S10, S22/D-A2) in one top-level ":root,.light" rule (the kit's own Light selector pair, so an
# isolated Light preview inside a Dark page keeps them) and one top-level .dark rule. Because :root also matches
# <html class="dark">, every name the Light rule corrects is declared again under .dark (Plan r1 §2.4 cascade as
# implemented at P0; Codex P0 gate review), so both lists hold the same names.
# lib/tokens.ts: exactly the 15 saved-branding colour names (core/internal/branding), sorted; no overlay, no roles.
TOKENS_SEMANTIC="accent background border destructive foreground input muted muted-foreground popover popover-foreground primary primary-foreground ring secondary secondary-foreground "

# css_uncomment FILE: the stylesheet with every /* ... */ comment removed (line structure kept).
css_uncomment() {
  awk '{
    line = $0; out = ""
    while (line != "") {
      if (inc) { i = index(line, "*/"); if (!i) { line = ""; break }; line = substr(line, i + 2); inc = 0 }
      else { i = index(line, "/*"); if (!i) { out = out line; line = "" } else { out = out substr(line, 1, i - 1); line = substr(line, i + 2); inc = 1 } }
    }
    print out
  }' "$1"
}
# css_rule_props SELECTOR: on stdin (uncommented CSS), the property names declared directly in each top-level rule whose
# selector is exactly SELECTOR, one per line, then "#RULES <n>" for the number of such rules.
css_rule_props() {
  awk -v sel="$1" '
    { l = $0; gsub(/^[ \t]+|[ \t]+$/, "", l); s = l; sub(/[ \t]*\{$/, "{", s) }
    depth == 0 && s == sel "{" { inb = 1; n++; depth = 1; next }
    inb && depth == 1 && l ~ /^-?-?[a-z][a-z0-9-]*[ \t]*:/ { p = l; sub(/[ \t]*:.*/, "", p); print p }
    { o = gsub(/\{/, "{", l); c = gsub(/\}/, "}", l); depth += o - c; if (depth <= 0) { depth = 0; inb = 0 } }
    END { print "#RULES " n + 0 }'
}

static_web_tokens() {
  local ok=1 W="$REPO/web" css="$REPO/web/app/globals.css" T="$WORK/tokens" hits want f
  mkdir -p "$T"
  # Palette source: the kit's theme.css, byte-equal to its manifest row (B-WEB-COMPONENTS proves the whole tree).
  want="$(sed -nE 's#^([0-9a-f]{64})  kit/src/styles/theme\.css$#\1#p' "$REPO/$REF_KIT_MANIFEST")"
  [[ -n "$want" && -f "$W/src/styles/theme.css" && "$(sha_of "$W/src/styles/theme.css")" == "$want" ]] || { ok=0; log "web/src/styles/theme.css is not the kit's palette source"; }
  css_uncomment "$css" >"$T/app.css"
  [[ "$(grep -m1 -vE '^[[:space:]]*$' "$T/app.css" | sed -E 's/^[[:space:]]+|[[:space:]]+$//g')" == '@import "../src/globals.css";' ]] || { ok=0; log 'web/app/globals.css: the first statement is not @import "../src/globals.css";'; }
  # No palette correction rules: the owner's exact-kit decision (G069-exact-kit-owner-decision-20261026.md) keeps the
  # kit's theme.css values as shipped, so app.css declares no top-level ":root,.light" or .dark rule (the selector is
  # matched exactly, without spaces, as css_rule_props compares it).
  css_rule_props ':root,.light' <"$T/app.css" >"$T/light"
  css_rule_props '.dark' <"$T/app.css" >"$T/dark"
  [[ "$(cat "$T/light")" == '#RULES 0' ]] || { ok=0; log "web/app/globals.css: a :root,.light correction rule remains (got $(tr '\n' ' ' <"$T/light"))"; }
  [[ "$(cat "$T/dark")" == '#RULES 0' ]] || { ok=0; log "web/app/globals.css: a .dark correction rule remains (got $(tr '\n' ' ' <"$T/dark"))"; }
  # lib/tokens.ts: exactly the 15 saved-branding names, TokenOverrides keyed by them; no roles or non-colour tokens.
  sed -n '/^export const SEMANTIC_COLOR_TOKENS = \[$/,/^\] as const;$/p' "$W/lib/tokens.ts" | sed -nE 's/^  "([a-z-]+)",$/\1/p' | LC_ALL=C sort >"$T/list"
  [[ "$(tr '\n' ' ' <"$T/list")" == "$TOKENS_SEMANTIC" ]] || { ok=0; log "SEMANTIC_COLOR_TOKENS is not exactly the 15 saved-branding names: $(tr '\n' ' ' <"$T/list")"; }
  ! grep -qE '^export const (INTERNAL_ROLE_TOKENS|NON_COLOR_TOKENS)\b' "$W/lib/tokens.ts" || { ok=0; log "lib/tokens.ts still exports INTERNAL_ROLE_TOKENS or NON_COLOR_TOKENS"; }
  grep -qxF 'export type TokenValues = Partial<Record<SemanticColorToken, string>>;' "$W/lib/tokens.ts" && grep -qxF 'export type TokenOverrides = { light?: TokenValues; dark?: TokenValues };' "$W/lib/tokens.ts" || { ok=0; log "TokenOverrides is no longer keyed by SemanticColorToken"; }
  # Colour literals: none in app.css or app-owned components, app or lib code (the kit's own files under web/src and
  # tests/fixtures are outside this scan).
  local -a prod
  mapfile -t prod < <(find "$W/components" "$W/lib" "$W/app" -type f \( -name '*.ts' -o -name '*.tsx' \))
  local lit='#[0-9a-fA-F]{3,8}\b|\b(rgba?|hsla?|oklch|oklab)\('
  hits="$(grep -nHE -- "$lit" "${prod[@]}" || true)"
  hits+="$(grep -nE -- "$lit" "$T/app.css" || true)"
  [[ -z "$hits" ]] || { ok=0; log "colour literal in app-owned code: $hits"; }
  # Palette utilities in app-owned code only (kit files are exempt). One recorded exception (S15): the literal
  # focus-visible:outline-white on the landing CTA, the kit CtaButton's own fixed label colour.
  local pal='(slate|gray|zinc|neutral|stone|red|orange|amber|yellow|lime|green|emerald|teal|cyan|sky|blue|indigo|violet|purple|fuchsia|pink|rose)-[0-9]{2,3}|(black|white)\b'
  local pre="\b(bg|text|border|ring|outline|fill|stroke|from|via|to|decoration|shadow|caret|divide|placeholder|accent)-($pal)"
  hits="$(for f in "${prod[@]}"; do
    if [[ "$f" == "$W/components/signup/landing.tsx" ]]; then sed 's/focus-visible:outline-white//g' "$f" | grep -nE -- "$pre" | sed "s|^|$f:|"
    else grep -nHE -- "$pre" "$f"; fi
  done || true)"
  [[ -z "$hits" ]] || { ok=0; log "palette utility: $hits"; }
  # The kit's accent is a neutral fill; app code does not paint it as a surface or pair it with accent-foreground
  # (utility names only: the tenant emitter's --accent-foreground custom property is not a utility).
  hits="$(grep -nHE -- '(^|[^-[:alnum:]_])((bg|from|via|to|fill)-accent\b|[a-z][a-z-]*-accent-foreground\b)' "${prod[@]}" || true)"
  [[ -z "$hits" ]] || { ok=0; log "accent used as a surface or with accent-foreground: $hits"; }
  [[ $ok == 1 ]] && record B-WEB-TOKENS PASS "palette source = the kit's web/src/styles/theme.css (its manifest SHA-256); web/app/globals.css first statement @import \"../src/globals.css\"; no top-level :root,.light or .dark correction rule (G069 exact-kit owner decision); SEMANTIC_COLOR_TOKENS exactly the 15 saved-branding names, TokenOverrides keyed by them, no role or non-colour token exports; no colour literal in app.css or app-owned code; no palette utility in app-owned components, app or lib (the one S15 landing focus-visible:outline-white); no accent surface or accent-foreground utility" || record B-WEB-TOKENS FAIL "token contract"
}

# web_files ROOT: web files (NUL-separated, absolute; no node_modules, build output or lockfile). The real
# repository through git (tracked and untracked, not ignored); a scratch copy built from that view by find.
web_files() {
  local root="$1"
  if [[ -e "$root/.git" ]]; then
    (cd "$root" && git ls-files -z --cached --others --exclude-standard -- web | while IFS= read -r -d '' f; do
      [[ -f "$f" && "$f" != web/package-lock.json ]] && { vestive_skill_path "$f" || printf '%s\0' "$root/$f"; }; done)
  else
    find "$root/web" \( -name node_modules -o -name .next -o -path "$root/$VESTIVE_SKILL" \) -prune -o -type f \
      ! -path "$root/web/package-lock.json" -print0
  fi
}

# ui_kit_manifest ROOT: "<sha256> <path>" for each kit/src/ row of ROOT's kit reference data (REF_KIT_MANIFEST), the
# path relative to kit/src (= relative to web/src). The checked tree's own reference: its row set is pinned by
# UI_KIT_FILES and UI_KIT_DIGEST (B-WEB-COMPONENTS) and its bytes by REF_KIT_MANIFEST_SHA (the DESIGN_LITERAL exemption).
ui_kit_manifest() { sed -nE 's#^([0-9a-f]{64})  kit/src/(.+)$#\1 \2#p' "$1/$REF_KIT_MANIFEST"; }

# web_components_check ROOT (G048 Plan §11.5; rewritten by G069 Plan r1 §3.2): prints the sorted reason codes of the
# adopted kit source under web/src and of its app-owned consumers under ROOT, and fails when there are any. Every rule
# is evaluated independently.
web_components_check() {
  local root="$1" W="$1/web" f rel slot p sha
  local -a src
  local -A allowed=() kitsha=()
  local ucodes=()
  mapfile -d '' -t src < <(web_files "$root")
  # (f) components.json as G067 left it.
  [[ "$(yq_stdin -p json -o tsv '[.style, .rsc, .tsx, .tailwind.css, .tailwind.cssVariables, .tailwind.baseColor, .iconLibrary, .aliases.components, .aliases.ui, .aliases.utils, .aliases.lib, .aliases.hooks] | join(" ")' <"$W/components.json")" == "base-rhea true true app/globals.css true taupe lucide @/components @/components/ui @/lib/utils @/lib @/hooks" ]] || ucodes+=(UI_COMPONENTS_JSON)
  # (a) Fidelity (AC1): the manifest's kit/src rows are the pinned 69-file set (digest UI_KIT_DIGEST); web/src holds
  # exactly those paths, each byte-equal to its row, and the Spec digest recomputed over web/src equals the pin. The
  # retired web/components/ui modules do not exist.
  while read -r sha p; do [[ -n "$p" ]] && kitsha["$p"]="$sha"; done < <(ui_kit_manifest "$root")
  [[ ${#kitsha[@]} == "$UI_KIT_FILES" && "$(printf '%s\n' "${!kitsha[@]}" | LC_ALL=C sort | while IFS= read -r p; do printf '%s  %s\n' "${kitsha[$p]}" "$p"; done | sha256sum | cut -d' ' -f1)" == "$UI_KIT_DIGEST" ]] || ucodes+=(UI_SET)
  [[ "$( { [[ -d "$W/src" ]] && cd "$W/src" && find . ! -type d -printf '%P\n'; } | LC_ALL=C sort)" == "$(printf '%s\n' "${!kitsha[@]}" | LC_ALL=C sort)" ]] || ucodes+=(UI_SET)
  [[ ! -e "$W/components/ui" ]] || ucodes+=(UI_SET)
  for p in "${!kitsha[@]}"; do
    [[ -f "$W/src/$p" && ! -L "$W/src/$p" && "$(sha_of "$W/src/$p")" == "${kitsha[$p]}" ]] || ucodes+=("UI_FIDELITY web/src/$p")
  done
  [[ -d "$W/src" && "$(cd "$W/src" && find . -type f -printf '%P\0' | LC_ALL=C sort -z | xargs -0 -r sha256sum | sha256sum | cut -d' ' -f1)" == "$UI_KIT_DIGEST" ]] || ucodes+=("UI_FIDELITY web/src")
  # (b) No competing implementation (S12, S21): no transitional legacy-ui directory, and none of the retired runtime's
  # markers (portal-container hook, the cn package, the data-appearance attribute, token scopes, the --light-*/--dark-*
  # sources, the old init script) in app code or tests. The appearance control's own data-appearance-* hooks are not
  # the retired attribute.
  [[ ! -e "$W/components/legacy-ui" ]] || ucodes+=("UI_LEGACY web/components/legacy-ui")
  for f in "${src[@]}"; do
    rel="${f#"$root"/}"
    case "$rel" in web/app/* | web/components/* | web/lib/* | web/tests/*) ;; *) continue ;; esac
    grep -qE -- "usePortalContainer|from [\"']cn[\"']|data-appearance([^-[:alnum:]_]|\$)|data-token-scope|--light-|--dark-|appearanceInitScript" "$f" && ucodes+=("UI_LEGACY $rel")
  done
  # Imports, tests included: (c) no Radix anywhere (cmdk reaches its Dialog only inside its own module graph);
  # (d) Base UI only inside the kit source; app code composes the kit modules.
  for f in "${src[@]}"; do
    [[ "$f" =~ \.(ts|tsx|js|jsx|mjs|cjs|mts|cts)$ ]] || continue
    rel="${f#"$root"/}"
    [[ -z "$(code_lines "[\"'\`]@radix-ui/" "$f")" ]] || ucodes+=("UI_IMPORT_RADIX $rel")
    [[ "$rel" == web/src/* || -z "$(code_lines "[\"'\`]@base-ui/" "$f")" ]] || ucodes+=("UI_IMPORT_BASEUI_APP $rel")
  done
  # (e) data-slot literals in app-owned code: the kit's own set (web/src) plus UI_APP_SLOTS.
  for slot in "${UI_APP_SLOTS[@]}"; do allowed[$slot]=1; done
  while IFS= read -r slot; do allowed[$slot]=1; done < <(grep -rhoE 'data-slot="[^"]+"' "$W/src" 2>/dev/null | sed -E 's/^data-slot="(.*)"$/\1/' | LC_ALL=C sort -u)
  while IFS= read -r slot; do
    [[ -n "${allowed[$slot]:-}" ]] || ucodes+=("UI_SLOT $slot")
  done < <(grep -rhoE 'data-slot="[^"]+"' "$W/app" "$W/components" "$W/lib" 2>/dev/null | sed -E 's/^data-slot="(.*)"$/\1/' | LC_ALL=C sort -u)
  # (h) S24: the fixed-white kit CtaButton renders only under the platform palette: CtaButton/ctaButtonVariants are
  # named only by components/controls/button-link.tsx, and variant="cta" is used only by the root landing and by the
  # test-only token fixture page, which renders its one CTA link only for ?tenant=platform (no tenant values; the
  # Plan r1 §4 P0.2 "outside the tenant style's effect" probe of the CTA's rest/hover/active label and inset outline).
  # (i) S22: no placeholder text in app-owned pages and components (so T12 needs no placeholder pair).
  for f in "${src[@]}"; do
    rel="${f#"$root"/}"
    case "$rel" in web/app/* | web/components/* | web/lib/* | web/tests/fixtures/*) ;; *) continue ;; esac
    [[ "$rel" == web/components/controls/button-link.tsx || -z "$(code_lines '\b(CtaButton|ctaButtonVariants)\b' "$f")" ]] || ucodes+=("UI_CTA_TENANT $rel")
    [[ "$rel" == web/components/signup/landing.tsx || "$rel" == web/tests/fixtures/token-scopes/page.tsx || -z "$(code_lines "variant=\{?[\"'\`]cta[\"'\`]" "$f")" ]] || ucodes+=("UI_CTA_TENANT $rel")
    case "$rel" in web/app/* | web/components/*) [[ -z "$(code_lines '\bplaceholder=' "$f")" ]] || ucodes+=("UI_PLACEHOLDER $rel") ;; esac
  done
  [[ ${#ucodes[@]} -eq 0 ]] && return 0
  printf '%s\n' "${ucodes[@]}" | LC_ALL=C sort -u
  return 1
}

static_web_components() {
  local codes
  if codes="$(web_components_check "$REPO")"; then
    record B-WEB-COMPONENTS PASS "components.json as plan §4.2 with style base-rhea and baseColor taupe (G067); G069 fidelity: web/src exactly the $UI_KIT_FILES kit/src files of the kit reference data ($REF_KIT_MANIFEST), each byte-equal to its row, Spec source-manifest digest $UI_KIT_DIGEST over the manifest rows and over web/src; no web/components/ui or web/components/legacy-ui; no retired-runtime marker (usePortalContainer, cn, data-appearance, data-token-scope, --light-/--dark-, appearanceInitScript) in app code or tests; no @radix-ui import anywhere in web (tests included); @base-ui imports only inside web/src; every app data-slot literal a kit slot or button; CtaButton/ctaButtonVariants only in components/controls/button-link.tsx and variant=\"cta\" only on the root landing and the platform-palette token fixture (S24); no placeholder attribute in app pages and components (S22)"
  else
    log "web components: ${codes//$'\n'/; }"; record B-WEB-COMPONENTS FAIL "adopted components"
  fi
}

# web_design_check ROOT (G048 Plan §8.2, §11.5): prints the sorted reason codes of web's design inputs under
# ROOT and fails when there are any. The kit reference data (REF_KIT_MANIFEST, REF_BRAND_HASHES) is ROOT's own.
web_design_check() {
  local root="$1" f rel ent path svg el attr id tok label n urls text sha p
  local lpat='vestive|vitruvian|interlock|restruct|strategic advisor|principal engineer|associate architect|book a free assessment|\$2,000|\$5,000|\$8,000'
  local -a src
  local -A asset=() pins=() seen=() count=() kit=()
  local ucodes=()
  mapfile -d '' -t src < <(web_files "$root")
  while IFS= read -r sha; do [[ "$sha" =~ ^[0-9a-f]{64}$ ]] && asset["$sha"]=1; done <"$root/$REF_BRAND_HASHES"
  for f in "${src[@]}"; do
    rel="${f#"$root"/}"
    # Remote fonts and assets, and @font-face outside globals.css, in code and stylesheets (public assets are
    # bounded by the exact inventory and the SVG rules below).
    if [[ "$rel" != web/public/* ]]; then
      grep -qiE -- 'fonts\.(googleapis|gstatic)|@import[[:space:]]+url\(|url\([[:space:]]*["'\'']?https?://|(src|href)=["'\''{`]+https?://' "$f" && ucodes+=("DESIGN_REMOTE $rel")
      [[ "$rel" != web/app/globals.css ]] && grep -qF '@font-face' "$f" && ucodes+=("FONT_FACE $rel")
    fi
  done
  # globals.css: exactly the five bundled faces (G067), and web CSS url() targets exactly the fonts and the drawing.
  [[ "$(grep -c '@font-face' "$root/web/app/globals.css")" == 5 ]] || ucodes+=("FONT_FACE web/app/globals.css")
  urls="$(for f in "${src[@]}"; do [[ "$f" == *.css ]] && grep -oE 'url\([[:space:]]*["'\'']?[^"'\'')[:space:]]+' "$f"; done | sed -E 's/^url\([[:space:]]*["'\'']?//' | LC_ALL=C sort -u | tr '\n' ' ')"
  [[ "$urls" == "$WEB_CSS_URLS" ]] || ucodes+=("FONT_FACE web/app/globals.css")
  # No web file byte-identical to a design-kit asset; no kit identity or marketing literal. G069 (Plan r1 S17): exempt
  # are exactly the web/src files that are byte-equal to their kit/src manifest row (by manifest, not by glob; an altered
  # or unlisted file is scanned), and in other files only the import lines of the kit's components/vestive modules.
  # The kit reference data itself is exempt only at its exact path and only while its bytes equal REF_KIT_MANIFEST_SHA.
  while read -r sha p; do [[ -n "$p" ]] && kit["web/src/$p"]="$sha"; done < <(ui_kit_manifest "$root")
  for f in "${src[@]}" "$root/web/package-lock.json"; do
    [[ -f "$f" ]] || continue
    rel="${f#"$root"/}"
    [[ -z "${asset[$(sha_of "$f")]:-}" ]] || ucodes+=("DESIGN_KIT_COPY $rel")
    if [[ -n "${kit[$rel]:-}" && "$(sha_of "$f")" == "${kit[$rel]}" ]]; then
      continue
    elif [[ "$rel" == "$REF_KIT_MANIFEST" && "$(sha_of "$f")" == "$REF_KIT_MANIFEST_SHA" ]]; then
      continue
    elif [[ "$rel" == web/eslint.config.mjs ]]; then
      text="$(eslint_literal_text "$f")" || ucodes+=("DESIGN_LITERAL_EXEMPT web/eslint.config.mjs")
      grep -qiE -- "$lpat" <<<"$text" && ucodes+=("DESIGN_LITERAL $rel")
    elif [[ "$rel" != web/src/* && "$rel" =~ \.(ts|tsx|js|jsx|mjs|cjs|mts|cts)$ ]]; then
      text="$(grep -vE '^import .* from "@/components/vestive(/[a-z-]+)?";?$' -- "$f" || true)"
      grep -qiE -- "$lpat" <<<"$text" && ucodes+=("DESIGN_LITERAL $rel")
    else
      grep -qiE -- "$lpat" "$f" && ucodes+=("DESIGN_LITERAL $rel")
    fi
  done
  # web/public: exactly the pinned inventory.
  for ent in "${WEB_PUBLIC_PINS[@]}"; do pins["${ent%% *}"]="${ent##* }"; done
  for f in "${src[@]}"; do
    rel="${f#"$root"/}"
    [[ "$rel" == web/public/* ]] || continue
    seen[$rel]=1
    if [[ -z "${pins[$rel]:-}" ]]; then ucodes+=("ASSET_EXTRA $rel")
    elif [[ "$(sha_of "$f")" != "${pins[$rel]}" ]]; then ucodes+=("ASSET_HASH $rel"); fi
  done
  for path in "${!pins[@]}"; do [[ -n "${seen[$path]:-}" ]] || ucodes+=("ASSET_MISSING $path"); done
  # The blueprint drawing (§8.1-8.2): parse, root, raw-byte denial, element and attribute allow-lists, size, motifs.
  svg="$root/web/public/blueprint/overview.svg"
  if [[ -f "$svg" ]]; then
    [[ "$(stat -c %s "$svg")" -le 40960 ]] || ucodes+=(SVG_SIZE)
    [[ "$(yq_stdin -p xml --xml-attribute-prefix '+@' -o tsv '[(keys | join(",")), .svg."+@viewBox", .svg."+@xmlns"] | join("|")' <"$svg" 2>/dev/null)" == "svg|0 0 960 640|http://www.w3.org/2000/svg" ]] || ucodes+=(SVG_PARSE)
    for tok in '<!' '<?' href 'url(' script style foreignObject '<image' '<use' '<animate' '<set' '@import' 'data:'; do
      label="${tok#<}"
      grep -qF -- "$tok" "$svg" && ucodes+=("SVG_UNSAFE $label")
    done
    while IFS= read -r attr; do ucodes+=("SVG_UNSAFE $attr"); done < <(grep -oE '(^|[[:space:]])on[a-zA-Z]+[[:space:]]*=' "$svg" | sed -E 's/^[[:space:]]*//; s/[[:space:]]*=$//' | LC_ALL=C sort -u)
    while IFS= read -r el; do
      [[ " $SVG_ELEMENTS " == *" $el "* ]] || ucodes+=("SVG_UNSAFE $el")
    done < <(grep -oE '<[A-Za-z][A-Za-z0-9:_.-]*' "$svg" | cut -c2- | LC_ALL=C sort -u)
    while IFS= read -r attr; do
      [[ " $SVG_ATTRIBUTES " == *" $attr "* ]] || ucodes+=("SVG_UNSAFE $attr")
    done < <(grep -oE '[[:space:]][A-Za-z_:][A-Za-z0-9_:.-]*[[:space:]]*=' "$svg" | sed -E 's/^[[:space:]]+//; s/[[:space:]]*=$//' | LC_ALL=C sort -u)
    while IFS= read -r id; do
      [[ " construction dimensions structure " == *" $id "* ]] || ucodes+=("SVG_MOTIF $id")
    done < <(grep -oE '[[:space:]]id="[^"]*"' "$svg" | sed -E 's/^[[:space:]]id="(.*)"$/\1/' | LC_ALL=C sort -u)
    while IFS=$'\t' read -r id n; do count[$id]="$n"; done < <(awk '
      { line = $0
        while (match(line, /<g[[:space:]][^>]*id="[a-z]+"|<\/g>|<(path|line|polyline|polygon|rect|circle)[[:space:]\/>]/)) {
          tok = substr(line, RSTART, RLENGTH); line = substr(line, RSTART + RLENGTH)
          if (tok ~ /^<g/) { sub(/.*id="/, "", tok); sub(/".*/, "", tok); cur = tok; c[cur] += 0 }
          else if (tok == "</g>") cur = ""
          else if (cur != "") c[cur]++
        } }
      END { for (k in c) printf "%s\t%d\n", k, c[k] }' "$svg")
    for ent in construction:10 dimensions:8 structure:10; do
      id="${ent%%:*}"; n="${ent##*:}"
      [[ -n "${count[$id]:-}" && "${count[$id]}" -ge "$n" ]] || ucodes+=("SVG_MOTIF $id")
    done
  fi
  [[ ${#ucodes[@]} -eq 0 ]] && return 0
  printf '%s\n' "${ucodes[@]}" | LC_ALL=C sort -u
  return 1
}

static_web_design() {
  local codes
  if codes="$(web_design_check "$REPO")"; then
    record B-WEB-DESIGN PASS "no remote font or asset URL in web code or styles; @font-face only in globals.css, exactly 5, and web CSS url() targets exactly the 5 bundled fonts and /blueprint/overview.svg; no web file byte-identical to a design-kit asset; no kit identity or marketing literal (G069 S17: except in the web/src files byte-equal to their kit manifest row and in import lines of @/components/vestive modules; and in $REF_KIT_MANIFEST only while its SHA-256 is REF_KIT_MANIFEST_SHA); web/public exactly the ${#WEB_PUBLIC_PINS[@]} pinned files; overview.svg parses with the SVG root and viewBox, has no unsafe byte pattern, only allow-listed elements and attributes, at most 40960 bytes, and the construction, dimensions and structure motifs"
  else
    log "web design: ${codes//$'\n'/; }"; record B-WEB-DESIGN FAIL "design inputs"
  fi
}

# B-WEB-ASSET-CONTROLS (G048 Plan §11.5): each case mutates a scratch copy of web and must produce exactly its
# expected code set from web_components_check plus web_design_check; A13 (the real inputs) produces none.
static_web_asset_controls() {
  local C="$WORK/asset-controls" ok=1 d svg long
  mkdir -p "$C/base"; : >"$WORK/asset-controls.tsv"
  (cd "$REPO" && git ls-files -z --cached --others --exclude-standard -- web |
    while IFS= read -r -d '' f; do [[ -e "$f" ]] && { vestive_skill_path "$f" || printf '%s\0' "$f"; }; done | tar --null -cf - -T -) | tar -xf - -C "$C/base"
  # asset_case ID WANT ROOT: the union of both checks' codes must equal WANT exactly.
  asset_case() {
    local id="$1" want="$2" root="$3" got verdict=pass
    got="$( { web_components_check "$root" || true; web_design_check "$root" || true; } | LC_ALL=C sort -u)"
    [[ "$got" == "$want" ]] || { verdict=FAIL; ok=0; log "asset control $id: expected {${want//$'\n'/; }}, got {${got//$'\n'/; }}"; }
    printf '%s\t%s\t%s\t%s\n' "$id" "${want//$'\n'/;}" "${got//$'\n'/;}" "$verdict" >>"$WORK/asset-controls.tsv"
  }
  fresh() { rm -rf "${C:?}/$1"; cp -a "$C/base" "$C/$1"; printf '%s' "$C/$1"; }
  # G069 (Plan r1 §3.2): a byte-altered kit file at its real web/src path (per-file and aggregate fidelity).
  d="$(fresh A1)"; printf '\n' >>"$d/web/src/components/vestive/ribbon.tsx"
  asset_case A1 $'UI_FIDELITY web/src\nUI_FIDELITY web/src/components/vestive/ribbon.tsx' "$d"
  # A2: the neutral test-only image stands in for a brand image: its hash joins this case's run-local copy of the brand
  # hash list (the shipped list holds only the omitted brand images' hashes), and the image moves to a public path.
  d="$(fresh A2)"; mkdir -p "$d/web/public/brand"; sha_of "$d/$ASSET_CONTROL_FIXTURE" >>"$d/$REF_BRAND_HASHES"
  mv "$d/$ASSET_CONTROL_FIXTURE" "$d/web/public/brand/logo.png"
  asset_case A2 $'ASSET_EXTRA web/public/brand/logo.png\nDESIGN_KIT_COPY web/public/brand/logo.png' "$d"
  d="$(fresh A3a)"; mkdir -p "$d/web/public/fonts/extra"; printf 'x' >"$d/web/public/fonts/extra/x.ttf"
  asset_case A3a "ASSET_EXTRA web/public/fonts/extra/x.ttf" "$d"
  d="$(fresh A3b)"; printf 'X' | dd of="$d/web/public/fonts/oswald/oswald-wght.ttf" bs=1 seek=0 count=1 conv=notrunc 2>/dev/null
  asset_case A3b "ASSET_HASH web/public/fonts/oswald/oswald-wght.ttf" "$d"
  d="$(fresh A4)"; printf '%s\n' '@font-face { font-family: "Remote"; src: url("https://fonts.gstatic.com/x.woff2"); }' >>"$d/web/app/globals.css"
  asset_case A4 $'DESIGN_REMOTE web/app/globals.css\nFONT_FACE web/app/globals.css' "$d"
  # G069: A5 and A12b append to an app-owned module (web/components/ui is retired).
  d="$(fresh A5)"; printf '%s\n' 'export const g048Control = "@font-face { font-family: x }";' >>"$d/web/components/layout/brand-mark.tsx"
  asset_case A5 "FONT_FACE web/components/layout/brand-mark.tsx" "$d"
  svg=web/public/blueprint/overview.svg
  d="$(fresh A6)"; sed -i 's#</svg>#<script/></svg>#' "$d/$svg"
  asset_case A6 $'ASSET_HASH web/public/blueprint/overview.svg\nSVG_UNSAFE script' "$d"
  d="$(fresh A7)"; sed -i '1s#<svg #<svg onload="void 0" #' "$d/$svg"
  asset_case A7 $'ASSET_HASH web/public/blueprint/overview.svg\nSVG_UNSAFE onload' "$d"
  d="$(fresh A8)"; sed -i 's#</svg>#<image href="https://x/y.png"/></svg>#' "$d/$svg"
  asset_case A8 $'ASSET_HASH web/public/blueprint/overview.svg\nSVG_UNSAFE href\nSVG_UNSAFE image' "$d"
  d="$(fresh A9)"; sed -i 's#</svg>#<foreignObject/></svg>#' "$d/$svg"
  asset_case A9 $'ASSET_HASH web/public/blueprint/overview.svg\nSVG_UNSAFE foreignObject' "$d"
  d="$(fresh A10)"; printf -v long '%*s' 11000 ''; long="${long// /L1 1}"; sed -i "s#</svg>#<path d=\"M0 0${long}\"/></svg>#" "$d/$svg"
  asset_case A10 $'ASSET_HASH web/public/blueprint/overview.svg\nSVG_SIZE' "$d"
  d="$(fresh A11)"; sed -i '/<g id="dimensions"/,/<\/g>/d' "$d/$svg"
  asset_case A11 $'ASSET_HASH web/public/blueprint/overview.svg\nSVG_MOTIF dimensions' "$d"
  d="$(fresh A12)"; printf '%s\n' 'import * as Dialog from "@radix-ui/react-dialog";' >>"$d/web/tests/browser/overlay.spec.ts"
  asset_case A12 "UI_IMPORT_RADIX web/tests/browser/overlay.spec.ts" "$d"
  d="$(fresh A12b)"; printf '%s\n' 'import { Menu } from "@base-ui/react/menu";' >>"$d/web/components/layout/brand-mark.tsx"
  asset_case A12b "UI_IMPORT_BASEUI_APP web/components/layout/brand-mark.tsx" "$d"
  # G049 (cycle 2), retargeted by G067: the exact two-line DESIGN_LITERAL exemption in web/eslint.config.mjs, and its limits.
  local el=web/eslint.config.mjs
  d="$(fresh A14)"; printf '%s\n' '// restruct' >>"$d/$el"
  asset_case A14 "DESIGN_LITERAL $el" "$d"
  d="$(fresh A15)"; printf '%s\n' "${G067_ESLINT_LINES[0]}" >>"$d/$el"
  asset_case A15 $'DESIGN_LITERAL web/eslint.config.mjs\nDESIGN_LITERAL_EXEMPT web/eslint.config.mjs' "$d"
  d="$(fresh A16)"; sed -i 's#"\.claude/skills/vestive-brand/\*\*"#".claude/skills/vestive-brand/*"#' "$d/$el"
  asset_case A16 $'DESIGN_LITERAL web/eslint.config.mjs\nDESIGN_LITERAL_EXEMPT web/eslint.config.mjs' "$d"
  d="$(fresh A17)"; grep -vxF -- "${G067_ESLINT_LINES[0]}" "$d/$el" >"$d/$el.tmp" && mv "$d/$el.tmp" "$d/$el"
  asset_case A17 $'DESIGN_LITERAL web/eslint.config.mjs\nDESIGN_LITERAL_EXEMPT web/eslint.config.mjs' "$d"
  d="$(fresh A18)"; mkdir -p "$d/web/.claude/notes"; printf 'vestive notes\n' >"$d/web/.claude/notes/x.md"
  asset_case A18 "DESIGN_LITERAL web/.claude/notes/x.md" "$d"
  # A19, retargeted: another file beside the kit reference data is scanned normally (no directory exemption).
  d="$(fresh A19)"; mkdir -p "$d/web/tests/reference"; printf 'vestive reference\n' >"$d/web/tests/reference/x.md"
  asset_case A19 "DESIGN_LITERAL web/tests/reference/x.md" "$d"
  # G069 (Plan r1 S17): the literal exemption is by manifest: an altered kit file loses it (A20); in app files only the
  # import lines of @/components/vestive modules are exempt, not another line naming it (A21).
  d="$(fresh A20)"; printf '\n' >>"$d/web/src/components/vestive/cta-button.tsx"
  asset_case A20 $'DESIGN_LITERAL web/src/components/vestive/cta-button.tsx\nUI_FIDELITY web/src\nUI_FIDELITY web/src/components/vestive/cta-button.tsx' "$d"
  d="$(fresh A21)"; printf '%s\n' 'export { Kicker } from "@/components/vestive/typography";' >>"$d/web/components/layout/brand-mark.tsx"
  asset_case A21 "DESIGN_LITERAL web/components/layout/brand-mark.tsx" "$d"
  # G069 (Plan r1 S24, S22): a CTA variant on the tenant-served sign-in page (A22), and placeholder text (A23).
  d="$(fresh A22)"; printf '%s\n' 'export const g069Control = <ButtonLink href="/" variant="cta">x</ButtonLink>;' >>"$d/web/app/sign-in/page.tsx"
  asset_case A22 "UI_CTA_TENANT web/app/sign-in/page.tsx" "$d"
  d="$(fresh A23)"; printf '%s\n' 'export const g069Control = <input name="x" placeholder="x" />;' >>"$d/web/app/sign-in/page.tsx"
  asset_case A23 "UI_PLACEHOLDER web/app/sign-in/page.tsx" "$d"
  # The kit reference data: tampered bytes lose the literal exemption while its 69 rows still parse (A24); an altered
  # row breaks the pinned row set and that file's fidelity, and also loses the exemption (A25).
  d="$(fresh A24)"; printf '# tampered\n' >>"$d/$REF_KIT_MANIFEST"
  asset_case A24 "DESIGN_LITERAL $REF_KIT_MANIFEST" "$d"
  d="$(fresh A25)"; sed -i -E 's#^[0-9a-f]{64}(  kit/src/components/ui/accordion\.tsx)$#0000000000000000000000000000000000000000000000000000000000000000\1#' "$d/$REF_KIT_MANIFEST"
  asset_case A25 $'DESIGN_LITERAL web/tests/reference/kit-source.sha256\nUI_FIDELITY web/src/components/ui/accordion.tsx\nUI_SET' "$d"
  asset_case A13 "" "$REPO"
  [[ $ok == 1 ]] && record B-WEB-ASSET-CONTROLS PASS "A1-A12 (G069: a byte-altered kit file at its web/src path; asset copies, extra or altered public files, a remote and a misplaced @font-face, unsafe, oversized and motif-less SVGs, a Radix import in a test; A12b a Base UI import in an app-owned module) and A14-A19 (G049, retargeted by G067: another literal in the ESLint configuration; a duplicated, changed or missing reviewed G067 line; a literal in unrelated hidden web code; a literal in another file beside the kit reference data) and A20-A23 (G069: an altered kit file loses the literal exemption; a non-import line naming the kit modules; a CTA variant on the tenant-served sign-in page; placeholder text) and A24-A25 (the kit reference data: tampered bytes lose the literal exemption; an altered row fails the row set and that file's fidelity) each produce exactly their expected codes; A13 (the real inputs) produces none; see asset-controls.tsv" || record B-WEB-ASSET-CONTROLS FAIL "an asset control did not produce exactly its expected code set"
}

# B-MCP-SOURCE (G014 successor; Plan r8 §7, r1 §7.1): the server links the MCP SDK (exactly its pinned closure),
# core/internal/mcp, the access facade and the tenancy model; one streamable HTTP handler call site, in core/internal/mcp;
# only core/internal/mcp, (G016) core/internal/domains/api and (G017) core/internal/branding/api import the access
# package, only the access package imports tenancy, and only the access package and (G017) core/internal/branding/api
# import objectstore; main.go equals its reviewed digest (six identity routes incl. the relay; G017 branding). Retained
# negatives: no migration code in the server, no fixture, isolation or identity test package in any binary, the migrate
# binary links neither tenancy nor MCP, no other transport and no WithIsolateDDL in core source, no other schema validator.
static_mcp_source() {
  local ok=1 img="g001-static-$RUN_ID-core-check" deps want imp m="$REPO/core/cmd/server/main.go"
  static_env
  ledger_watch_begin "$REPO" "g001-static-$RUN_ID" static-core-check
  (cd "$REPO" && docker compose --env-file "$WORK/static.env" -f infra/compose/local.yml --profile checks build core-check) >"$WORK/static-core-check-build.log" 2>&1 || { ok=0; tail -20 "$WORK/static-core-check-build.log"; }
  ledger_watch_end "$REPO" "g001-static-$RUN_ID" static-core-check "$WORK/static-core-check-build.log" \
    "docker compose --env-file $WORK/static.env -f infra/compose/local.yml --profile checks build core-check"
  deps="$(docker_run_owned go-list --rm --label g001.run=$RUN_ID --network none "$img" go list -deps ./cmd/server 2>&1)" || { ok=0; log "go list ./cmd/server: $deps"; }
  printf '%s\n' "$deps" >"$WORK/deps-server.txt"
  for want in core/internal/mcp core/internal/tenancy/access core/internal/tenancy github.com/modelcontextprotocol/go-sdk/mcp; do
    grep -qxF "$want" "$WORK/deps-server.txt" || { ok=0; log "the server binary does not link $want"; }
  done
  [[ "$(grep -E '^(github\.com/(modelcontextprotocol|google/jsonschema-go|yosida95|segmentio)/|golang\.org/x/oauth2)' "$WORK/deps-server.txt" | LC_ALL=C sort | paste -sd' ')" == "$MCP_SDK_CLOSURE" ]] || { ok=0; log "the server's SDK package closure differs from the pinned set"; }
  if grep -E 'pressly/goose|^core/internal/migrate$|^core/migrations$' "$WORK/deps-server.txt"; then ok=0; log "the server binary links migration code"; fi
  docker_run_owned go-list --rm --label g001.run=$RUN_ID --network none "$img" go list -deps ./cmd/migrate >"$WORK/deps-migrate.txt" 2>&1 || ok=0
  if cat "$WORK/deps-server.txt" "$WORK/deps-migrate.txt" | grep -E 'example\.invalid|isolation|fixture|identitytest'; then ok=0; log "a binary depends on fixture, isolation or identity test code"; fi
  if grep -E '^core/internal/tenancy(/.*)?$|^core/internal/mcp$|modelcontextprotocol' "$WORK/deps-migrate.txt"; then ok=0; log "the migrate binary links tenancy or MCP code"; fi
  grep -qx 'core/internal/migrate' "$WORK/deps-migrate.txt" || { ok=0; log "go list -deps ./cmd/migrate did not list core/internal/migrate"; }
  # Code lines only: test files and // comments are ignored.
  if grep -rnE 'NewSSEHandler|StreamableServerTransport|SSEServerTransport|StdioTransport|CommandTransport|IOTransport|InMemoryTransport|WithIsolateDDL' --include='*.go' "$REPO/core" | grep -v '_test\.go:' | grep -vE '^[^:]+:[0-9]+:[[:space:]]*//'; then ok=0; log "another transport or isolated-DDL symbol in core source"; fi
  [[ "$(grep -rnE 'NewStreamableHTTPHandler\(' --include='*.go' "$REPO/core" | grep -v '_test\.go:' | grep -vE '^[^:]+:[0-9]+:[[:space:]]*//' | cut -d: -f1 | sed "s#^$REPO/##" | paste -sd' ')" == "core/internal/mcp/handler.go" ]] || { ok=0; log "not exactly one streamable HTTP handler call site, in core/internal/mcp"; }
  # Non-test importers of a package, as directories relative to the repository.
  importers() { grep -rlE "^[[:space:]]*(import )?([A-Za-z_]+ )?\"$1\"" --include='*.go' "$REPO/core" | grep -v '_test\.go$' | xargs -r -n1 dirname | sed "s#^$REPO/##" | LC_ALL=C sort -u | paste -sd' '; }
  imp="$(importers core/internal/tenancy/access)"; [[ "$imp" == "core/internal/branding/api core/internal/domains/api core/internal/enrollment/api core/internal/labels/api core/internal/mcp core/internal/membership/api core/internal/onboarding/api" ]] || { ok=0; log "access package importers (want core/internal/branding/api core/internal/domains/api core/internal/enrollment/api core/internal/labels/api core/internal/mcp core/internal/membership/api core/internal/onboarding/api): $imp"; }
  imp="$(importers core/internal/onboarding/api)"; [[ "$imp" == "core/cmd/server" ]] || { ok=0; log "core/internal/onboarding/api importers: $imp"; }
  imp="$(importers core/internal/membership/api)"; [[ "$imp" == "core/cmd/server" ]] || { ok=0; log "core/internal/membership/api importers: $imp"; }
  imp="$(importers core/internal/enrollment/api)"; [[ "$imp" == "core/cmd/server" ]] || { ok=0; log "core/internal/enrollment/api importers: $imp"; }
  imp="$(importers core/internal/signup/api)"; [[ "$imp" == "core/cmd/server" ]] || { ok=0; log "core/internal/signup/api importers: $imp"; }
  imp="$(importers core/internal/labels/api)"; [[ "$imp" == "core/cmd/server" ]] || { ok=0; log "core/internal/labels/api importers: $imp"; }
  imp="$(importers core/internal/tenancy)"; [[ "$imp" == "core/internal/tenancy/access" ]] || { ok=0; log "tenancy importers: $imp"; }
  imp="$(importers core/internal/objectstore)"; [[ "$imp" == "core/internal/branding/api core/internal/tenancy/access" ]] || { ok=0; log "objectstore importers: $imp"; }
  imp="$(importers core/internal/mcp)"; [[ "$imp" == "core/cmd/server" ]] || { ok=0; log "core/internal/mcp importers: $imp"; }
  [[ "$(sha_of "$m")" == "$MAIN_GO_G016_SHA" ]] || { ok=0; log "core/cmd/server/main.go differs from its reviewed G016 digest"; }
  [[ "$(docker_run_owned go-list --rm --label g001.run=$RUN_ID --network none "$img" go list -f '{{join .Imports " "}}' ./cmd/server 2>&1)" == "$MAIN_GO_IMPORTS" ]] || { ok=0; log "core/cmd/server imports differ from the reviewed G016 set (the rendered set minus fmt plus the identity, MCP and domain packages)"; }
  if grep -iE 'jsonschema' "$WORK/deps-server.txt" | grep -vxF 'github.com/google/jsonschema-go/jsonschema'; then ok=0; log "the server binary links another jsonschema package"; fi
  # Routes: the two literal health routes plus exactly the six identity and three MCP route constants, whose patterns
  # are pinned; the MCP handler wraps the mux only inside the configured branch (the digest pins main.go).
  [[ "$(grep -oE 'mux\.Handle(Func)?\("[^"]*"' "$m" | LC_ALL=C sort | tr '\n' ' ')" == 'mux.HandleFunc("GET /healthz" mux.HandleFunc("GET /readyz" ' ]] || { ok=0; log "unexpected literal core routes"; }
  [[ "$(grep -oE 'mux\.Handle(Func)?\([^",)]+' "$m" | LC_ALL=C sort | tr '\n' ' ')" == 'mux.HandleFunc(identity.RouteCallback mux.HandleFunc(identity.RouteLogin mux.HandleFunc(identity.RouteLogout mux.HandleFunc(identity.RouteProviderReturn mux.HandleFunc(identity.RouteRelay mux.HandleFunc(identity.RouteSession mux.HandleFunc(mcp.RouteConnectionsCreate mux.HandleFunc(mcp.RouteConnectionsList mux.HandleFunc(mcp.RouteConnectionsRename mux.HandleFunc(mcp.RouteConnectionsRevoke ' ]] || { ok=0; log "core routes other than the six identity and four MCP setup routes"; }
  [[ "$(sed -nE 's/^[[:space:]]+Route[A-Za-z]+[[:space:]]+= "([^"]+)"$/\1/p' "$REPO/core/internal/identity/http.go" | LC_ALL=C sort | paste -sd'|')" == "$CORE_AUTH_ROUTES" ]] || { ok=0; log "identity route patterns differ from the six reviewed /internal/v1/auth routes"; }
  [[ "$(sed -nE 's/^[[:space:]]+Route[A-Za-z]+[[:space:]]+= "([^"]+)"$/\1/p' "$REPO/core/internal/mcp/routes.go" | LC_ALL=C sort | paste -sd'|')" == "$CORE_MCP_ROUTES" ]] || { ok=0; log "MCP route patterns differ from the four reviewed /internal/v1/mcp routes"; }
  [[ "$(sed -nE 's/^[[:space:]]+Route[A-Za-z]+[[:space:]]+= "([^"]+)"$/\1/p' "$REPO/core/internal/onboarding/api/http.go" | LC_ALL=C sort | paste -sd'|')" == "$CORE_ONBOARDING_ROUTES" ]] || { ok=0; log "onboarding route patterns differ from the eight reviewed G018 pairs"; }
  [[ "$(sed -nE 's/^[[:space:]]+Route[A-Za-z]+[[:space:]]+= "([^"]+)"$/\1/p' "$REPO/core/internal/membership/api/http.go" | LC_ALL=C sort | paste -sd'|')" == "$CORE_MEMBERSHIP_ROUTES" ]] || { ok=0; log "membership route patterns differ from the eighteen reviewed G019/G060 pairs"; }
  [[ "$(sed -nE 's/^[[:space:]]+Route[A-Za-z]+[[:space:]]+= "([^"]+)"$/\1/p' "$REPO/core/internal/enrollment/api/http.go" | LC_ALL=C sort | paste -sd'|')" == "$CORE_ENROLLMENT_ROUTES" ]] || { ok=0; log "enrollment route patterns differ from the ten reviewed G058 pairs"; }
  [[ "$(sed -nE 's/^[[:space:]]+Route[A-Za-z]+[[:space:]]+= "([^"]+)"$/\1/p' "$REPO/core/internal/signup/api/http.go" | LC_ALL=C sort | paste -sd'|')" == "$CORE_SIGNUP_ROUTES" ]] || { ok=0; log "signup route patterns differ from the seven reviewed G062 pairs"; }
  [[ "$(sed -nE 's/^[[:space:]]+Route[A-Za-z]+[[:space:]]+= "([^"]+)"$/\1/p' "$REPO/core/internal/labels/api/http.go" | LC_ALL=C sort | paste -sd'|')" == "$CORE_LABELS_ROUTES" ]] || { ok=0; log "label route patterns differ from the seven reviewed G020 pairs"; }
  [[ $ok == 1 ]] && record B-MCP-SOURCE PASS "server binary links core/internal/mcp, the access facade, the tenancy model and exactly the pinned go-sdk v1.8.0 closure (google/jsonschema-go the only schema package), and no migration code; no binary depends on fixture, isolation or identity test packages; the migrate binary links neither tenancy nor MCP; exactly one streamable HTTP handler call site (core/internal/mcp/handler.go) and no other transport or WithIsolateDDL in core source; only core/internal/mcp, core/internal/domains/api, core/internal/branding/api, core/internal/onboarding/api, core/internal/membership/api, core/internal/enrollment/api and core/internal/labels/api import the access package (only the server imports the onboarding, membership, enrollment, signup and labels adapters), the enrollment adapter declares exactly its ten reviewed method/path pairs, the G062 signup adapter its seven and the G020 labels adapter its seven, only the access package imports tenancy, only the access package and core/internal/branding/api import objectstore, only the server imports core/internal/mcp; server main.go equals its reviewed G020 digest, imports the rendered set minus fmt plus the identity, MCP, domain, branding, onboarding, membership, enrollment, signup and labels packages, the onboarding adapter declares exactly its eight reviewed method/path pairs and the membership adapter its eighteen (G019 sixteen plus the two G060 Group-authority pairs), routes GET /healthz, GET /readyz, exactly the six /internal/v1/auth routes (including the callback relay) and the four /internal/v1/mcp setup routes (G063: rename)" || record B-MCP-SOURCE FAIL "MCP or runtime import boundary"
}

# The accepted migrations stay byte-identical (G005 Plan r4 §6.2; G006 Plan r2 RC-1; G007 Plan r1 RC-1; G008 Plan r3 RC-1):
# 00001 (G001), 00002 (G004), 00003 (G005), 00004 (G006) and 00005 (G007).
MIGRATION_00001_SHA=db18518a954290c71b607eb9413bd2d66b9d4bf78645ad3fb2d1057f2883c60f
MIGRATION_00002_SHA=8d44b30073292f6b0949195b84a852bca40fcb058e21abe65c9c1a25b0fb6e14
MIGRATION_00003_SHA=23c45ee7ad45cd0779220c9cf24af60ef7db263b166e88bd8eb9e4cde8f872e3
MIGRATION_00004_SHA=f73bc9df96a177c1717354afda50e8ee3918ea8c11a2fd0a5233bd08a4b2a677
MIGRATION_00005_SHA=743f161d15221ffe400c341627bdd4d8e737b3a8523e2864f87abc525469cfdd
# G009 (Plan r1 §5.6): the accepted G008 00006_row_level_security.sql is pinned too. G011 (Plan r2 §10.2): the accepted
# G009 00007_stored_objects.sql is pinned; 00008 is the G011 addition. G013 (Plan r3 §11.3): the accepted G011
# 00008_audit_events.sql is pinned; 00009 is the G013 addition. G014 (Plan r8 §11.3): the accepted G013
# 00009_retention_maintenance.sql is pinned; 00010 is the G014 addition. G016: the accepted G014
# 00010_mcp_connections.sql is pinned; 00011 is the G016 addition and 00012 its security amendment (A2). G017: 00013 is the branding addition. G018: 00014 is the Group onboarding addition. G019: 00015 is the membership invitation addition. G058: 00016 is the enrollment addition. G060: 00017 is the Group authority addition. G062: 00018 is the platform signup addition.
MIGRATION_00010_SHA=3ac29b6bcf2078158cf7fbaac13b244ce2610450d2e6f0b53fba734b14a73e19
MIGRATION_00006_SHA=fae1813b6b747a4d385eb16a4b900c92290b3eb82331ca85bb6d41a44c7384ed
MIGRATION_00007_SHA=28f0a073631a0b10ca17602e2b5897a6c02a17e305ca795c959a89d403100d12
MIGRATION_00008_SHA=b1dedf1f56f28c97a9f8c59796ff1b615c1ed79e6b8ab8f2cfb602098c597233
MIGRATION_00009_SHA=7fbfe7f9671d702e9ea1b4ed89f5e7c925e2409e2e3dd2dfc60e68022ade9b2b
static_migration_set() {
  local ok=1 set
  set="$(cd "$REPO/core/migrations" && ls ./*.sql | tr '\n' ' ')"
  [[ "$set" == "./00001_baseline.sql ./00002_identity.sql ./00003_tenancy.sql ./00004_business_profiles.sql ./00005_organization_policy.sql ./00006_row_level_security.sql ./00007_stored_objects.sql ./00008_audit_events.sql ./00009_retention_maintenance.sql ./00010_mcp_connections.sql ./00011_organization_domains.sql ./00012_session_binding_and_claim_exclusivity.sql ./00013_organization_branding.sql ./00014_group_onboarding.sql ./00015_membership_invitations.sql ./00016_organization_enrollment.sql ./00017_group_authority.sql ./00018_platform_signup.sql ./00019_mcp_connection_names.sql ./00020_labels.sql ./00021_platform_labels.sql ./00022_default_tenant.sql " ]] || { ok=0; log "normal migration set: $set"; }
  [[ "$(sha_of "$REPO/core/migrations/00001_baseline.sql")" == "$MIGRATION_00001_SHA" ]] || { ok=0; log "00001_baseline.sql differs from its accepted digest"; }
  [[ "$(sha_of "$REPO/core/migrations/00002_identity.sql")" == "$MIGRATION_00002_SHA" ]] || { ok=0; log "00002_identity.sql differs from its accepted digest"; }
  [[ "$(sha_of "$REPO/core/migrations/00003_tenancy.sql")" == "$MIGRATION_00003_SHA" ]] || { ok=0; log "00003_tenancy.sql differs from its accepted digest"; }
  [[ "$(sha_of "$REPO/core/migrations/00004_business_profiles.sql")" == "$MIGRATION_00004_SHA" ]] || { ok=0; log "00004_business_profiles.sql differs from its accepted digest"; }
  [[ "$(sha_of "$REPO/core/migrations/00005_organization_policy.sql")" == "$MIGRATION_00005_SHA" ]] || { ok=0; log "00005_organization_policy.sql differs from its accepted digest"; }
  [[ "$(sha_of "$REPO/core/migrations/00006_row_level_security.sql")" == "$MIGRATION_00006_SHA" ]] || { ok=0; log "00006_row_level_security.sql differs from its accepted digest"; }
  [[ "$(sha_of "$REPO/core/migrations/00007_stored_objects.sql")" == "$MIGRATION_00007_SHA" ]] || { ok=0; log "00007_stored_objects.sql differs from its accepted digest"; }
  [[ "$(sha_of "$REPO/core/migrations/00008_audit_events.sql")" == "$MIGRATION_00008_SHA" ]] || { ok=0; log "00008_audit_events.sql differs from its accepted digest"; }
  [[ "$(sha_of "$REPO/core/migrations/00009_retention_maintenance.sql")" == "$MIGRATION_00009_SHA" ]] || { ok=0; log "00009_retention_maintenance.sql differs from its accepted digest"; }
  [[ "$(sha_of "$REPO/core/migrations/00010_mcp_connections.sql")" == "$MIGRATION_00010_SHA" ]] || { ok=0; log "00010_mcp_connections.sql differs from its accepted digest"; }
  if find "$REPO/core" -name '9*.sql' | grep -v '/internal/migrate/testdata/synthetic/'; then ok=0; log "synthetic migration outside testdata"; fi
  [[ $ok == 1 ]] && record B-MIGRATE-SET PASS "normal set = {00001_baseline.sql, 00002_identity.sql, 00003_tenancy.sql, 00004_business_profiles.sql, 00005_organization_policy.sql, 00006_row_level_security.sql, 00007_stored_objects.sql, 00008_audit_events.sql, 00009_retention_maintenance.sql, 00010_mcp_connections.sql, 00011_organization_domains.sql, 00012_session_binding_and_claim_exclusivity.sql, 00013_organization_branding.sql, 00014_group_onboarding.sql, 00015_membership_invitations.sql, 00016_organization_enrollment.sql, 00017_group_authority.sql, 00018_platform_signup.sql, 00019_mcp_connection_names.sql, 00020_labels.sql, 00021_platform_labels.sql, 00022_default_tenant.sql}; 00001, 00002, 00003, 00004, 00005, 00006, 00007, 00008, 00009 and 00010 equal their accepted digests; the synthetic probe exists only under core/internal/migrate/testdata" || record B-MIGRATE-SET FAIL "migration set"
}

static_dockerignore() {
  local ok=1 s f
  for s in core web intelligence migrate validation-tools; do
    f="$REPO/infra/docker/$s.Dockerfile.dockerignore"
    [[ -f "$f" ]] && grep -qx '.env' "$f" && grep -qx '.env.\*' "$f" || { ok=0; log "$s.Dockerfile.dockerignore must exclude .env and .env.*"; }
  done
  if find "$REPO/infra/docker" -name '.env*' | grep -q .; then ok=0; log "environment file inside the MinIO build context"; fi
  # Application build contexts exclude operator-private files and the development submodule.
  f="$REPO/.dockerignore"
  [[ -f "$f" && ! -L "$f" ]] && grep -qx '.local' "$f" && grep -qx '.agents/skills/restruct-sdlc' "$f" ||
    { ok=0; log ".dockerignore must exclude .local and the Restruct development submodule"; }
  [[ $ok == 1 ]] && record B-DOCKERIGNORE PASS "service build contexts exclude environment files; the root context excludes .local and the Restruct development submodule" || record B-DOCKERIGNORE FAIL "build-context exclusions"
}

# B-CI-CONFIG, G072 part (Plan r1 §10.2, §10.8, §10.10, §10.13 AC1). ci_g072_check FILE REGISTRY QUIET CLI: FILE (the
# committed .gitlab-ci.yml or a mutation of it) has two documents; no STAGING_* variable and no interpolation other than
# the four plain inputs, candidate and duration only inside double quotes; the spec:inputs header exactly (environment
# options = none + REGISTRY, no images input); the three deployment-mode workflow refusals right after the four push/tag
# never rules; the guard first in validate:checks, validate:integration and identity-staging, none of which holds
# id_tokens (nor, for the validation jobs, an environment); the five staging jobs as exact JSON (credentials, environment
# actions, resource groups, rules, needs, timeouts, scripts, artifacts); and the launcher's context pins. Two yq reads.
# 0 when everything holds; QUIET=1 (the mutation controls) logs nothing and stops at the first failure.
ci_g072_check() {
  local f="$1" reg="$2" quiet="$3" cli="$4" bad=0 opts="" line got want pre ev ia ie ic id ctx sd st svc jq72 df i n
  local -a items=() jw=() names=(prepare check build deploy stop) tmo=()
  g72_fail() { bad=1; [[ $quiet == 1 ]] || log "G072 CI: $*"; }
  g72_stop() { [[ $bad == 1 && $quiet == 1 ]]; }
  [[ "$(grep -c '^---' "$f")" == 1 && "$(grep -m1 -vE '^(#|$)' "$f")" == "spec:" ]] || g72_fail "not exactly two documents with the spec header first"
  if grep -qE '\$\{?STAGING_' "$f"; then g72_fail "a STAGING_* variable is referenced (selection is by typed inputs only)"; fi
  [[ "$(grep -oE '\$\[\[[^]]*\]\]' "$f" | LC_ALL=C sort -u | tr '\n' ' ')" == '$[[ inputs.action ]] $[[ inputs.candidate ]] $[[ inputs.duration ]] $[[ inputs.environment ]] ' ]] ||
    g72_fail "an interpolation other than the four plain inputs"
  if grep -oE '.\$\[\[ inputs\.(candidate|duration) \]\].' "$f" | grep -vqE '^"\$\[\[ inputs\.(candidate|duration) \]\]"$'; then
    g72_fail "candidate or duration interpolated outside double quotes"
  fi
  if [[ -f "$reg" && -s "$reg" ]] && ! grep -qvxE '[a-z][a-z0-9-]{0,30}' "$reg" && ! grep -qxE 'none|validate' "$reg" && [[ -z "$(sort "$reg" | uniq -d)" ]]; then
    opts='["none"'; while IFS= read -r line; do opts+=",\"$line\""; done <"$reg"; opts+=']'
  else g72_fail "infra/staging/ci-environments.txt is missing or malformed"; fi
  g72_stop && return 1
  want='{"spec":{"inputs":{"action":{"type":"string","description":"Validate (the ordinary pipeline), Deploy environment or Stop environment (G072; protected staging web/API pipelines only)","options":["Validate","Deploy environment","Stop environment"],"default":"Validate"},"environment":{"type":"string","description":"none, or the allowlisted environment (infra/staging/ci-environments.txt) for Deploy environment and Stop environment","options":'"$opts"',"default":"none"},"candidate":{"type":"string","description":"Deploy environment only: empty or staging (the commit of this pipeline), a full SHA, or for preview-1 a same-repository branch","default":"","regex":"^([0-9a-f]{40}|[A-Za-z0-9][A-Za-z0-9._/-]{0,199})?$"},"duration":{"type":"string","description":"Deploy environment only: empty (2 h for a stopped VM; a running VM keeps its session) or 600-28800 s as N, Nm or Nh","default":"","regex":"^([1-9][0-9]{0,5}[mh]?)?$"}}}}'
  got="$(yq_stdin -p yaml -o json -I0 'select(document_index == 0)' <"$f" 2>/dev/null)"
  [[ "$got" == "$want" ]] || g72_fail "the spec:inputs header differs (exactly action, environment = none + the registry, candidate and duration; no images input): $got"
  g72_stop && return 1
  mapfile -t items < <(yq_stdin -p yaml -o json -I0 'select(document_index == 1) | [(keys | sort | join(" ")), .workflow.rules, ."validate:checks".rules[0], ."validate:integration".rules[0], ."identity-staging".rules[0], (."validate:checks" | keys | sort | join(" ")), (."validate:integration" | keys | sort | join(" ")), (."identity-staging" | keys | sort | join(" ")), ."validate:checks".services, ."staging:prepare", ."staging:check", ."staging:build", ."staging:deploy", ."staging:stop"] | .[]' <"$f" 2>/dev/null)
  [[ ${#items[@]} == 14 ]] || { g72_fail "document 1 could not be read (${#items[@]} of 14 items)"; return 1; }
  [[ "${items[0]}" == '"identity-staging stages staging:build staging:check staging:deploy staging:prepare staging:stop validate:checks validate:integration variables workflow"' ]] ||
    g72_fail "top-level keys differ (only the validation/staging jobs and fixed checkout policy): ${items[0]}"
  [[ "$(yq_stdin -p yaml -o json -I0 'select(document_index == 1) | .variables' <"$f" 2>/dev/null)" == '{"GIT_SUBMODULE_STRATEGY":"none"}' ]] || g72_fail "global variables must only disable automatic submodule checkout"
  ia='\"$[[ inputs.action ]]\"' ie='\"$[[ inputs.environment ]]\"' ic='\"$[[ inputs.candidate ]]\"' id='\"$[[ inputs.duration ]]\"'
  want='{"if":"$CI_COMMIT_TAG","when":"never"},{"if":"'"$ia"' != \"Validate\" && $CI_PIPELINE_SOURCE != \"web\" && $CI_PIPELINE_SOURCE != \"api\"","when":"never"},{"if":"'"$ia"' != \"Validate\" && $CI_COMMIT_BRANCH != \"staging\"","when":"never"},{"if":"'"$ia"' != \"Validate\" && $CI_COMMIT_REF_PROTECTED != \"true\"","when":"never"},'
  pre="${items[1]%%"$want"*}"
  if [[ "${items[1]}" != *"$want"* || "$(grep -o '{"if":' <<<"$pre" | grep -c .)" != "$(grep -o '"when":"never"}' <<<"$pre" | grep -c .)" ]]; then
    g72_fail "the workflow does not refuse deployment mode outside web/API pipelines on the protected staging branch before admitting anything"
  fi
  for i in 2 3 4; do
    [[ "${items[i]}" == '{"if":"'"$ia"' != \"Validate\"","when":"never"}' ]] || g72_fail "a validation job (rules item $i) does not first refuse a non-default action"
  done
  for i in 5 6 7; do [[ " ${items[i]//\"/} " != *" id_tokens "* ]] || g72_fail "a validation job holds id_tokens"; done
  for i in 5 6; do [[ " ${items[i]//\"/} " != *" environment "* ]] || g72_fail "a validation job holds an environment"; done
  g72_stop && return 1
  svc="${items[8]}" ev='$[[ inputs.environment ]]'
  ctx='$CI_COMMIT_BRANCH == \"staging\" && $CI_COMMIT_REF_PROTECTED == \"true\" && ($CI_PIPELINE_SOURCE == \"web\" || $CI_PIPELINE_SOURCE == \"api\")'
  sd='[{"if":"'"$ia"' == \"Deploy environment\" && '"$ie"' != \"none\" && '"$ctx"'"}]'
  st='[{"if":"'"$ia"' == \"Stop environment\" && '"$ie"' != \"none\" && '"$ic"' == \"\" && '"$id"' == \"\" && '"$ctx"'"}]'
  g72_job() { # ACTION NEEDS RULES EXTRA TIMEOUT SCRIPT_TIMEOUT AFTER_SCRIPT_TIMEOUT ARTIFACTS
    printf '%s' '{"stage":"staging","needs":'"$2"',"rules":'"$3$4"',"tags":["saas-linux-medium-amd64"],"image":"'"$cli"'","services":'"$svc"',"timeout":"'"$5"'","interruptible":false,"variables":{"DOCKER_HOST":"tcp://docker:2376","DOCKER_TLS_CERTDIR":"/certs","DOCKER_TLS_VERIFY":"1","DOCKER_CERT_PATH":"/certs/client","GIT_STRATEGY":"clone","GIT_DEPTH":"0","GIT_CLEAN_FLAGS":"-ffdx","FF_GIT_URLS_WITHOUT_TOKENS":"true","RUNNER_SCRIPT_TIMEOUT":"'"$6"'","RUNNER_AFTER_SCRIPT_TIMEOUT":"'"$7"'"},"script":["sh infra/ci/staging-action.sh '"$1"' --environment '"$ie"' --candidate '"$ic"' --duration '"$id"'"],"after_script":["sh infra/ci/staging-action.sh --after-script '"$1"' --environment '"$ie"'"],"artifacts":'"$8"'}'
  }
  # prepare and build hold no ID token or environment; check reads (read audience, environment action prepare, no
  # resource group); deploy and stop mutate under staging-env-<env>; stop only with an empty candidate and duration.
  jw[0]="$(g72_job prepare '[]' "$sd" '' 10m 7m 2m '{"expire_in":"1 day","access":"none","paths":["staging-prepare/"]}')"
  jw[1]="$(g72_job check '["staging:prepare"]' "$sd" ',"environment":{"name":"'"$ev"'","action":"prepare"},"id_tokens":{"STAGING_CI_ID_TOKEN":{"aud":"staging-ci:'"$ev"':read"}}' 10m 7m 2m '{"expire_in":"1 day","access":"developer","reports":{"dotenv":"staging-check/reuse.env"}}')"
  jw[2]="$(g72_job build '["staging:prepare","staging:check"]' "$sd" ',"resource_group":"staging-image-build"' 45m 40m 4m '{"expire_in":"1 day","access":"none","paths":["staging-build/"]}')"
  jw[3]="$(g72_job deploy '["staging:prepare","staging:check","staging:build"]' "$sd" ',"environment":{"name":"'"$ev"'","action":"start"},"resource_group":"staging-env-'"$ev"'","id_tokens":{"STAGING_CI_ID_TOKEN":{"aud":"staging-ci:'"$ev"':start"}}' 60m 55m 4m '{"when":"always","expire_in":"90 days","access":"developer","paths":["staging-summary/"]}')"
  jw[4]="$(g72_job stop '[]' "$st" ',"environment":{"name":"'"$ev"'","action":"stop"},"resource_group":"staging-env-'"$ev"'","id_tokens":{"STAGING_CI_ID_TOKEN":{"aud":"staging-ci:'"$ev"':stop"}}' 30m 25m 4m '{"when":"always","expire_in":"90 days","access":"developer","paths":["staging-summary/"]}')"
  tmo=("10m 7m 2m" "10m 7m 2m" "45m 40m 4m" "60m 55m 4m" "30m 25m 4m")
  for i in 0 1 2 3 4; do
    got="${items[i + 9]}" n="${names[i]}"
    if [[ "$got" != "${jw[i]}" ]]; then
      if [[ ( $n == prepare || $n == build ) && ( "$got" == *'"id_tokens"'* || "$got" == *'"environment"'* ) ]]; then g72_fail "staging:$n holds an ID token or environment (credential-free job)"; fi
      g72_fail "staging:$n differs from its exact definition: $got"
    fi
    read -r line want pre <<<"${tmo[i]}"
    (( $(dur "$want") + $(dur "$pre") <= $(dur "$line") )) || g72_fail "staging:$n script plus after-script timeouts exceed the job timeout"
  done
  g72_stop && return 1
  if [[ $quiet != 1 ]]; then
    # The launcher's context image: the ci-context CLI stage, check.sh's jq pin and a digest-pinned google-cloud-cli.
    df="$REPO/infra/docker/staging-ci.Dockerfile" jq72="$(sed -n 's/^JQ_IMAGE="\(.*\)"$/\1/p' "$REPO/infra/staging/check.sh")"
    if [[ -n "$cli" && -n "$jq72" ]] && grep -qxF "FROM $cli AS dockercli" "$df" && grep -qxF "FROM $jq72 AS jq" "$df" &&
      grep -qxE 'FROM gcr\.io/google\.com/cloudsdktool/google-cloud-cli:[0-9][0-9.]*-slim@sha256:[0-9a-f]{64}' "$df" &&
      [[ "$(grep -c '^FROM ' "$df")" == 3 ]] && ! grep -q 'TODO-DIGEST' "$df"; then :
    else g72_fail "infra/docker/staging-ci.Dockerfile does not pin the ci-context CLI stage, check.sh's jq and a digest-pinned google-cloud-cli"; fi
  fi
  [[ $bad == 0 ]]
}

# ci_g072_mutations FILE REGISTRY CLI: each control removes one protected behaviour from a copy of the committed file (or
# registry) under $WORK; ci_g072_check must fail on it. A control that does not apply, or passes, fails B-CI-CONFIG.
# Sets G72_MUTS to the count and names.
G72_MUTS=""
ci_g072_mutations() {
  local f="$1" reg="$2" cli="$3" d="$WORK/b-ci-config-mutations" src i n=0 ok=1 sdr hdr unp
  local -a mutname=() tgt=() old=() new=()
  g72_mut() { mutname+=("$1"); tgt+=("$2"); old+=("$3"); new+=("$4"); }
  hdr=$'"staging:prepare":\n  stage: staging\n  needs: []\n'
  sdr=$'  rules:\n    - if: \'"$[[ inputs.action ]]" == "Deploy environment" && "$[[ inputs.environment ]]" != "none" && $CI_COMMIT_BRANCH == "staging" && $CI_COMMIT_REF_PROTECTED == "true" && ($CI_PIPELINE_SOURCE == "web" || $CI_PIPELINE_SOURCE == "api")\'\n'
  g72_mut prepare-id-token yml "$hdr" "$hdr"$'  id_tokens:\n    STAGING_CI_ID_TOKEN:\n      aud: "staging-ci:$[[ inputs.environment ]]:start"\n'
  g72_mut prepare-environment yml "$hdr" "$hdr"$'  environment:\n    name: "$[[ inputs.environment ]]"\n'
  g72_mut build-id-token yml $'  resource_group: staging-image-build\n' $'  resource_group: staging-image-build\n  id_tokens:\n    STAGING_CI_ID_TOKEN:\n      aud: "staging-ci:$[[ inputs.environment ]]:start"\n'
  g72_mut build-environment yml $'  resource_group: staging-image-build\n' $'  resource_group: staging-image-build\n  environment:\n    name: "$[[ inputs.environment ]]"\n'
  g72_mut build-needs-check yml 'needs: ["staging:prepare", "staging:check"]' 'needs: ["staging:prepare"]'
  g72_mut check-audience yml 'aud: "staging-ci:$[[ inputs.environment ]]:read"' 'aud: "staging-ci:$[[ inputs.environment ]]:start"'
  g72_mut check-resource-group yml $'    action: prepare\n' $'    action: prepare\n  resource_group: "staging-env-$[[ inputs.environment ]]"\n'
  g72_mut deploy-audience yml 'aud: "staging-ci:$[[ inputs.environment ]]:start"' 'aud: "staging-ci:$[[ inputs.environment ]]:stop"'
  g72_mut deploy-environment-action yml $'    action: start\n' $'    action: prepare\n'
  g72_mut deploy-no-resource-group yml $'    action: start\n  resource_group: "staging-env-$[[ inputs.environment ]]"\n' $'    action: start\n'
  unp='$CI_COMMIT_REF_PROTECTED == "true" && '
  g72_mut deploy-unprotected yml $'"staging:deploy":\n  stage: staging\n  needs: ["staging:prepare", "staging:check", "staging:build"]\n'"$sdr" \
    $'"staging:deploy":\n  stage: staging\n  needs: ["staging:prepare", "staging:check", "staging:build"]\n'"${sdr/"$unp"/}"
  g72_mut deploy-allow-failure yml $'"staging:deploy":\n  stage: staging\n' $'"staging:deploy":\n  stage: staging\n  allow_failure: true\n'
  g72_mut stop-audience yml 'aud: "staging-ci:$[[ inputs.environment ]]:stop"' 'aud: "staging-ci:$[[ inputs.environment ]]:start"'
  g72_mut stop-resource-group yml $'    action: stop\n  resource_group: "staging-env-$[[ inputs.environment ]]"\n' $'    action: stop\n  resource_group: "staging-stop-$[[ inputs.environment ]]"\n'
  g72_mut stop-with-candidate yml ' && "$[[ inputs.candidate ]]" == ""' ''
  g72_mut stop-with-duration yml ' && "$[[ inputs.duration ]]" == ""' ''
  g72_mut guard-checks yml $'"validate:checks":\n  stage: checks\n  rules:\n    - if: \'"$[[ inputs.action ]]" != "Validate"\'\n      when: never\n' $'"validate:checks":\n  stage: checks\n  rules:\n'
  g72_mut guard-integration yml $'    # runs no validation job.\n    - if: \'"$[[ inputs.action ]]" != "Validate"\'\n      when: never\n' $'    # runs no validation job.\n'
  g72_mut guard-identity yml $'  rules:\n    - if: \'"$[[ inputs.action ]]" != "Validate"\'\n      when: never\n    - if: $CI_COMMIT_BRANCH == "identity-qualification"' $'  rules:\n    - if: $CI_COMMIT_BRANCH == "identity-qualification"'
  g72_mut validation-id-token yml $'"validate:checks":\n  stage: checks\n' $'"validate:checks":\n  stage: checks\n  id_tokens:\n    VALIDATE_ID_TOKEN:\n      aud: validate\n'
  g72_mut workflow-source yml $'    - if: \'"$[[ inputs.action ]]" != "Validate" && $CI_PIPELINE_SOURCE != "web" && $CI_PIPELINE_SOURCE != "api"\'\n      when: never\n' ''
  g72_mut workflow-branch yml $'    - if: \'"$[[ inputs.action ]]" != "Validate" && $CI_COMMIT_BRANCH != "staging"\'\n      when: never\n' ''
  g72_mut workflow-protected yml $'    - if: \'"$[[ inputs.action ]]" != "Validate" && $CI_COMMIT_REF_PROTECTED != "true"\'\n      when: never\n' ''
  g72_mut images-input yml $'    duration:\n' $'    images:\n      type: string\n      options: [build, published]\n      default: build\n    duration:\n'
  g72_mut candidate-regex yml "'^([0-9a-f]{40}|[A-Za-z0-9][A-Za-z0-9._/-]{0,199})?\$'" "'^.*\$'"
  g72_mut candidate-unquoted yml '--candidate "$[[ inputs.candidate ]]"' '--candidate $[[ inputs.candidate ]]'
  g72_mut interpolation-function yml $'name: "$[[ inputs.environment ]]"\n    action: start' $'name: "$[[ inputs.environment | expand_vars ]]"\n    action: start'
  g72_mut variable-selection yml "deploy --environment \"\$[[ inputs.environment ]]\"" 'deploy --environment "$STAGING_ENV"'
  g72_mut variable-rule yml $'  rules:\n    - if: \'"$[[ inputs.action ]]" == "Stop environment"' $'  rules:\n    - if: $DEPLOY_ACTION == "stop"\n    - if: \'"$[[ inputs.action ]]" == "Stop environment"'
  g72_mut top-level-variables yml $'stages: [checks, integration, staging]\n' $'stages: [checks, integration, staging]\nvariables:\n  DEPLOY_TARGET: staging\n'
  g72_mut registry-extra reg "" $'staging\npreview-1\npreview-2\n'
  g72_mut registry-differs reg "" $'staging\npreview-10\n'
  rm -rf "$d"; mkdir -p "$d"
  src="$(<"$f")"
  for i in "${!mutname[@]}"; do
    if [[ ${tgt[i]} == yml ]]; then
      if [[ "$src" != *"${old[i]}"* ]]; then ok=0; log "B-CI-CONFIG mutation control ${mutname[i]} does not apply"; continue; fi
      printf '%s\n' "${src/"${old[i]}"/"${new[i]}"}" >"$d/${mutname[i]}.yml"; cp "$reg" "$d/${mutname[i]}.reg"
    else
      cp "$f" "$d/${mutname[i]}.yml"; printf '%s' "${new[i]}" >"$d/${mutname[i]}.reg"
    fi
    if ci_g072_check "$d/${mutname[i]}.yml" "$d/${mutname[i]}.reg" 1 "$cli"; then ok=0; log "B-CI-CONFIG mutation control ${mutname[i]} passed: the assertion does not detect it"
    else n=$((n + 1)); fi
  done
  rm -rf "$d"
  G72_MUTS="$n of ${#mutname[@]} (${mutname[*]})"
  [[ $ok == 1 && $n == "${#mutname[@]}" ]]
}

# B-CI-CONFIG (G003 Plan §11.1): the single gating job, its rules, artifacts and limits, and the
# README platform:ci command. Parsed with the pinned yq image; nothing is executed.
static_ci_config() {
  local ok=1 f="$REPO/.gitlab-ci.yml" k job stage cli dind want tag s a sum jt prof id suite
  local -A known=() cnt=() ps=()
  [[ -f "$f" ]] || { record B-CI-CONFIG FAIL ".gitlab-ci.yml missing"; return; }
  # G072: two YAML documents, the spec:inputs header (document 0) and the pipeline configuration (document 1).
  q() { yq_stdin -p yaml -o tsv "select(document_index == 1) | $1" <"$f" 2>/dev/null; }
  qj() { yq_stdin -p yaml -o json -I0 "select(document_index == 1) | $1" <"$f" 2>/dev/null; }
  dur() { # Go duration (h, m, s parts) in seconds
    local d="$1" h=0 m=0 s2=0
    [[ "$d" =~ ^(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?$ && -n "$d" ]] || { echo -1; return; }
    [[ -n "${BASH_REMATCH[2]}" ]] && h="${BASH_REMATCH[2]}"; [[ -n "${BASH_REMATCH[4]}" ]] && m="${BASH_REMATCH[4]}"
    [[ -n "${BASH_REMATCH[6]}" ]] && s2="${BASH_REMATCH[6]}"
    echo $((10#$h * 3600 + 10#$m * 60 + 10#$s2))
  }
  [[ "$(q 'keys | sort | join(" ")')" == "identity-staging stages staging:build staging:check staging:deploy staging:prepare staging:stop validate:checks validate:integration variables workflow" ]] || { ok=0; log "top-level validation job keys differ"; }
  [[ "$(qj '.variables')" == '{"GIT_SUBMODULE_STRATEGY":"none"}' ]] || { ok=0; log "CI must not initialize development submodules automatically"; }
  # Owner one-off: automatic documentation Markdown bypass; G049: exactly three handoff files also bypass, nested
  # Claude skills are always admitted. Explicit full contexts remain.
  # ONEOFF RC qualification (owner 2026-10-19: no automatic post-merge cloud checks): no push starts a pipeline
  # (feature, staging and default-branch pushes and every tag); the RC-promotion merge request
  # (release/<canonical RC> to the default branch) is admitted without a changes filter, ahead of the generic MR rule,
  # and both jobs' first rule for it adds the staging and tag refspecs to the runner's credentialed fetch.
  local rcif rcvars guard refuse72
  # G072: every validation job first refuses deployment mode; the workflow admits deployment mode only for web/API
  # pipelines on the protected staging branch (three never rules right after the tag rule).
  guard='{"if":"\"$[[ inputs.action ]]\" != \"Validate\"","when":"never"}'
  refuse72='{"if":"\"$[[ inputs.action ]]\" != \"Validate\" && $CI_PIPELINE_SOURCE != \"web\" && $CI_PIPELINE_SOURCE != \"api\"","when":"never"},{"if":"\"$[[ inputs.action ]]\" != \"Validate\" && $CI_COMMIT_BRANCH != \"staging\"","when":"never"},{"if":"\"$[[ inputs.action ]]\" != \"Validate\" && $CI_COMMIT_REF_PROTECTED != \"true\"","when":"never"}'
  rcif='$CI_PIPELINE_SOURCE == \"merge_request_event\" && $CI_MERGE_REQUEST_TARGET_BRANCH_NAME == $CI_DEFAULT_BRANCH && $CI_MERGE_REQUEST_SOURCE_BRANCH_NAME =~ /^release\\/(0|[1-9][0-9]*)\\.(0|[1-9][0-9]*)\\.(0|[1-9][0-9]*)-rc\\.([1-9][0-9]*)$/'
  rcvars='{"GIT_FETCH_EXTRA_FLAGS":"--prune --quiet +refs/heads/staging:refs/remotes/origin/staging +refs/tags/*:refs/tags/*"}'
  want='[{"if":"$CI_PIPELINE_SOURCE == \"push\" && $CI_COMMIT_BRANCH && $CI_COMMIT_BRANCH != \"staging\" && $CI_COMMIT_BRANCH != $CI_DEFAULT_BRANCH","when":"never"},{"if":"$CI_PIPELINE_SOURCE == \"push\" && $CI_COMMIT_BRANCH == \"staging\"","when":"never"},{"if":"$CI_PIPELINE_SOURCE == \"push\" && $CI_COMMIT_BRANCH == $CI_DEFAULT_BRANCH","when":"never"},{"if":"$CI_COMMIT_TAG","when":"never"},'"$refuse72"',{"if":"'"$rcif"'"},{"if":"$CI_PIPELINE_SOURCE == \"merge_request_event\"","auto_cancel":{"on_new_commit":"interruptible"},"changes":["{*[^d],*[^m]d,*[^.]md,d,md}","{[^.dm]*,.[^a]*,.a[^g]*,.ag[^e]*,.age[^n]*,.agen[^t]*,.agent[^s]*,.agents?*,.,.a,.ag,.age,.agen,.agent,d[^o]*,do[^c]*,doc[^s]*,docs?*,d,do,doc,m[^a]*,ma[^n]*,man[^a]*,mana[^g]*,manag[^e]*,manage[^m]*,managem[^e]*,manageme[^n]*,managemen[^t]*,management?*,m,ma,man,mana,manag,manage,managem,manageme,managemen}/**/*",".agents/*",".agents/{[^s]*,s[^k]*,sk[^i]*,ski[^l]*,skil[^l]*,skill[^s]*,skills?*,s,sk,ski,skil,skill}/**/*",".agents/skills/*",".agents/skills/{[^c]*,c[^l]*,cl[^a]*,cla[^u]*,clau[^d]*,claud[^e]*,claude[^-]*,claude-[^s]*,claude-s[^e]*,claude-se[^s]*,claude-ses[^s]*,claude-sess[^i]*,claude-sessi[^o]*,claude-sessio[^n]*,claude-session[^-]*,claude-session-[^h]*,claude-session-h[^a]*,claude-session-ha[^n]*,claude-session-han[^d]*,claude-session-hand[^o]*,claude-session-hando[^f]*,claude-session-handof[^f]*,claude-session-handoff?*,c,cl,cla,clau,claud,claude,claude-,claude-s,claude-se,claude-ses,claude-sess,claude-sessi,claude-sessio,claude-session,claude-session-,claude-session-h,claude-session-ha,claude-session-han,claude-session-hand,claude-session-hando,claude-session-handof}/**/*",".agents/skills/claude-session-handoff/{[^S]*,S[^K]*,SK[^I]*,SKI[^L]*,SKIL[^L]*,SKILL[^.]*,SKILL.[^m]*,SKILL.m[^d]*,SKILL.md?*,S,SK,SKI,SKIL,SKILL,SKILL.,SKILL.m}",".agents/skills/claude-session-handoff/{[^ar]*,a[^g]*,ag[^e]*,age[^n]*,agen[^t]*,agent[^s]*,agents?*,a,ag,age,agen,agent,r[^e]*,re[^f]*,ref[^e]*,refe[^r]*,refer[^e]*,refere[^n]*,referen[^c]*,referenc[^e]*,reference[^s]*,references?*,r,re,ref,refe,refer,refere,referen,referenc,reference}/**/*",".agents/skills/claude-session-handoff/agents/{[^o]*,o[^p]*,op[^e]*,ope[^n]*,open[^a]*,opena[^i]*,openai[^.]*,openai.[^y]*,openai.y[^a]*,openai.ya[^m]*,openai.yam[^l]*,openai.yaml?*,o,op,ope,open,opena,openai,openai.,openai.y,openai.ya,openai.yam}",".agents/skills/claude-session-handoff/agents/*/**/*",".agents/skills/claude-session-handoff/references/{[^l]*,l[^o]*,lo[^c]*,loc[^a]*,loca[^l]*,local[^-]*,local-[^s]*,local-s[^e]*,local-se[^s]*,local-ses[^s]*,local-sess[^i]*,local-sessi[^o]*,local-sessio[^n]*,local-session[^-]*,local-session-[^h]*,local-session-h[^a]*,local-session-ha[^n]*,local-session-han[^d]*,local-session-hand[^o]*,local-session-hando[^f]*,local-session-handof[^f]*,local-session-handoff[^.]*,local-session-handoff.[^m]*,local-session-handoff.m[^d]*,local-session-handoff.md?*,l,lo,loc,loca,local,local-,local-s,local-se,local-ses,local-sess,local-sessi,local-sessio,local-session,local-session-,local-session-h,local-session-ha,local-session-han,local-session-hand,local-session-hando,local-session-handof,local-session-handoff,local-session-handoff.,local-session-handoff.m}",".agents/skills/claude-session-handoff/references/*/**/*","**/.claude/skills/**/*","docs/**/{*[^d],*[^m]d,*[^.]md,d,md}","management/{*[^d],*[^m]d,*[^.]md,d,md}","management/{[^v]*,v[^0]*,v,v0?*}/**/{*[^d],*[^m]d,*[^.]md,d,md}","management/v0/{*[^dl],*[^m]d,*[^.]md,*[^n]l,*[^o]nl,*[^s]onl,*[^j]sonl,*[^.]jsonl,d,md,l,nl,onl,sonl,jsonl,.jsonl}","management/v0/{[^S]*,S[^E]*,SE[^S]*,SES[^S]*,SESS[^I]*,SESSI[^O]*,SESSIO[^N]*,SESSION[^-]*,SESSION-[^H]*,SESSION-H[^A]*,SESSION-HA[^N]*,SESSION-HAN[^D]*,SESSION-HAND[^O]*,SESSION-HANDO[^F]*,SESSION-HANDOF[^F]*,SESSION-HANDOFF[^S]*,SESSION-HANDOFFS?*,S,SE,SES,SESS,SESSI,SESSIO,SESSION,SESSION-,SESSION-H,SESSION-HA,SESSION-HAN,SESSION-HAND,SESSION-HANDO,SESSION-HANDOF,SESSION-HANDOFF}.jsonl","management/v0/{[^pr]*,p[^l]*,pl[^a]*,pla[^n]*,plan[^s]*,plans?*,r[^e]*,re[^s]*,res[^u]*,resu[^l]*,resul[^t]*,result[^s]*,results?*,p,pl,pla,plan,r,re,res,resu,resul,result}/**/{*[^d],*[^m]d,*[^.]md,d,md}","management/v0/plans/*/**/{*[^d],*[^m]d,*[^.]md,d,md}","management/v0/plans/{*[^dn],*[^o]n,*[^s]on,*[^j]son,*[^.]json,n,on,son,json,*[^m]d,*[^.]md,d,md}","management/v0/plans/{[^G]*,G[^0-9]*,G[0-9][^0-9]*,G[0-9][0-9][^0-9]*,G[0-9][0-9][0-9][^-]*,G[0-9][0-9][0-9]-[^p]*,G[0-9][0-9][0-9]-p[^l]*,G[0-9][0-9][0-9]-pl[^a]*,G[0-9][0-9][0-9]-pla[^n]*,G[0-9][0-9][0-9]-plan[^n]*,G[0-9][0-9][0-9]-plann[^i]*,G[0-9][0-9][0-9]-planni[^n]*,G[0-9][0-9][0-9]-plannin[^g]*,G[0-9][0-9][0-9]-planning[^-]*,G[0-9][0-9][0-9]-planning-[^i]*,G[0-9][0-9][0-9]-planning-i[^n]*,G[0-9][0-9][0-9]-planning-in[^p]*,G[0-9][0-9][0-9]-planning-inp[^u]*,G[0-9][0-9][0-9]-planning-inpu[^t]*,G[0-9][0-9][0-9]-planning-input[^s]*,G,G[0-9],G[0-9][0-9],G[0-9][0-9][0-9],G[0-9][0-9][0-9]-,G[0-9][0-9][0-9]-p,G[0-9][0-9][0-9]-pl,G[0-9][0-9][0-9]-pla,G[0-9][0-9][0-9]-plan,G[0-9][0-9][0-9]-plann,G[0-9][0-9][0-9]-planni,G[0-9][0-9][0-9]-plannin,G[0-9][0-9][0-9]-planning,G[0-9][0-9][0-9]-planning-,G[0-9][0-9][0-9]-planning-i,G[0-9][0-9][0-9]-planning-in,G[0-9][0-9][0-9]-planning-inp,G[0-9][0-9][0-9]-planning-inpu,G[0-9][0-9][0-9]-planning-input}.json","management/v0/results/*/**/{*[^d],*[^m]d,*[^.]md,d,md}","management/v0/results/{*[^dt],*[^m]d,*[^.]md,d,md,*[^x]t,*[^t]xt,*[^.]txt,t,xt,txt,.txt}","management/v0/results/{[^G]*,G[^0-9]*,G[0-9][^0-9]*,G[0-9][0-9][^0-9]*,G[0-9][0-9][0-9][^-]*,G[0-9][0-9][0-9]-[^s]*,G[0-9][0-9][0-9]-s[^c]*,G[0-9][0-9][0-9]-sc[^o]*,G[0-9][0-9][0-9]-sco[^p]*,G[0-9][0-9][0-9]-scop[^e]*,G[0-9][0-9][0-9]-scope[^-]*,G[0-9][0-9][0-9]-scope-[^a]*,G[0-9][0-9][0-9]-scope-a[^l]*,G[0-9][0-9][0-9]-scope-al[^l]*,G[0-9][0-9][0-9]-scope-all[^o]*,G[0-9][0-9][0-9]-scope-allo[^w]*,G[0-9][0-9][0-9]-scope-allow[^l]*,G[0-9][0-9][0-9]-scope-allowl[^i]*,G[0-9][0-9][0-9]-scope-allowli[^s]*,G[0-9][0-9][0-9]-scope-allowlis[^t]*,G[0-9][0-9][0-9]-scope-allowlist?*,G,G[0-9],G[0-9][0-9],G[0-9][0-9][0-9],G[0-9][0-9][0-9]-,G[0-9][0-9][0-9]-s,G[0-9][0-9][0-9]-sc,G[0-9][0-9][0-9]-sco,G[0-9][0-9][0-9]-scop,G[0-9][0-9][0-9]-scope,G[0-9][0-9][0-9]-scope-,G[0-9][0-9][0-9]-scope-a,G[0-9][0-9][0-9]-scope-al,G[0-9][0-9][0-9]-scope-all,G[0-9][0-9][0-9]-scope-allo,G[0-9][0-9][0-9]-scope-allow,G[0-9][0-9][0-9]-scope-allowl,G[0-9][0-9][0-9]-scope-allowli,G[0-9][0-9][0-9]-scope-allowlis}.txt"]},{"if":"$CI_PIPELINE_SOURCE == \"merge_request_event\"","when":"never"},{"if":"$CI_PIPELINE_SOURCE == \"schedule\" && $CI_COMMIT_BRANCH == $CI_DEFAULT_BRANCH"},{"if":"$CI_PIPELINE_SOURCE == \"schedule\" && $CI_COMMIT_BRANCH == \"identity-qualification\" && $CI_COMMIT_REF_PROTECTED == \"true\""},{"if":"$CI_PIPELINE_SOURCE == \"web\" && $CI_COMMIT_BRANCH"},{"if":"$CI_PIPELINE_SOURCE == \"api\" && $CI_COMMIT_BRANCH"},{"when":"never"}]'
  [[ "$(qj '.workflow.rules')" == "$want" && "$(q '.workflow | keys | sort | join(" ")')" == "auto_cancel rules" && "$(qj '.workflow.auto_cancel')" == '{"on_new_commit":"none"}' ]] || { ok=0; log "workflow rules or MR-only auto-cancel policy differ"; }
  # GitLab rejects a rules:changes list with more than 50 entries (a pipeline failed with zero jobs).
  local nchg
  while IFS= read -r nchg; do
    [[ "$nchg" =~ ^[0-9]+$ && "$nchg" -le 50 ]] || { ok=0; log "a workflow rules:changes list has ${nchg:-?} entries (GitLab maximum 50)"; }
  done < <(q 'explode(.) | .workflow.rules[] | select(has("changes")) | .changes | length')
  [[ "$(q '.stages | join(" ")')" == "checks integration staging" ]] || { ok=0; log "stages are not exactly checks, integration, staging"; }
  cli="$(sed -n 's/^FROM \(docker:[^ ]*-cli@sha256:[0-9a-f]\{64\}\) AS dockercli$/\1/p' "$REPO/infra/docker/ci-context.Dockerfile")"
  dind="$(sed -n 's/^FROM docker:\([^@]*\)-cli@.*/\1/p' "$REPO/infra/docker/ci-context.Dockerfile")"
  [[ -n "$cli" && -n "$dind" ]] || { ok=0; log "ci-context.Dockerfile has no pinned docker CLI stage"; }
  for prof in fast gate; do
    if [[ "$prof" == fast ]]; then job=validate:checks; stage=checks; else job=validate:integration; stage=integration; fi
    if [[ "$prof" == fast ]]; then want="after_script artifacts image interruptible rules script services stage tags timeout variables"
    else want="after_script artifacts image interruptible needs rules script services stage tags timeout variables"; fi
    [[ "$(q ".\"$job\" | keys | sort | join(\" \")")" == "$want" ]] || { ok=0; log "$job keys: $(q ".\"$job\" | keys | sort | join(\" \")")"; }
    for k in allow_failure when only except retry parallel trigger environment release pages cache extends include dependencies resource_group coverage hooks secrets id_tokens inherit manual_confirmation; do
      [[ "$(q ".\"$job\" | has(\"$k\")")" == false ]] || { ok=0; log "$job has $k"; }
    done
    [[ "$(q ".\"$job\".stage")" == "$stage" && "$(q ".\"$job\".interruptible")" == false ]] || { ok=0; log "$job stage or interruptible"; }
    [[ "$(q ".\"$job\".image")" == "$cli" ]] || { ok=0; log "$job image is not the pinned CLI equal to the ci-context CLI stage"; }
    a="$(qj ".\"$job\".services")"
    [[ "$a" =~ ^\[\{\"name\":\"docker:${dind//./\\.}-dind@sha256:[0-9a-f]{64}\",\"alias\":\"docker\"\}\]$ ]] || { ok=0; log "$job services are not exactly the pinned TLS DinD daemon alias docker"; }
    tag="$(q ".\"$job\".tags | join(\",\")")"
    [[ "$tag" == saas-linux-medium-amd64 ]] || { ok=0; log "$job tags are not exactly the hosted medium runner tag"; }
    [[ "$(q ".\"$job\".script | length")" == 1 && "$(q ".\"$job\".script[0]")" == "sh infra/ci/hosted-context.sh -- --profile $prof --publish-dir ci-evidence" ]] || { ok=0; log "$job script differs"; }
    [[ "$(q ".\"$job\".after_script | length")" == 1 && "$(q ".\"$job\".after_script[0]")" == "sh infra/ci/hosted-context.sh --after-script" ]] || { ok=0; log "$job after_script differs"; }
    [[ "$(qj ".\"$job\".artifacts")" == '{"when":"always","expire_in":"90 days","access":"developer","paths":["ci-evidence/"],"reports":{"junit":"ci-evidence/junit.xml"}}' ]] || { ok=0; log "$job artifacts differ"; }
    if [[ "$prof" == fast ]]; then s="35m" a="4m" jt="40m"; else s="2h45m" a="10m" jt="3h"; fi
    want="{\"DOCKER_HOST\":\"tcp://docker:2376\",\"DOCKER_TLS_CERTDIR\":\"/certs\",\"DOCKER_TLS_VERIFY\":\"1\",\"DOCKER_CERT_PATH\":\"/certs/client\",\"GIT_STRATEGY\":\"clone\",\"GIT_DEPTH\":\"0\",\"GIT_CLEAN_FLAGS\":\"-ffdx\",\"FF_GIT_URLS_WITHOUT_TOKENS\":\"true\",\"RUNNER_SCRIPT_TIMEOUT\":\"$s\",\"RUNNER_AFTER_SCRIPT_TIMEOUT\":\"$a\"}"
    [[ "$(qj ".\"$job\".variables")" == "$want" ]] || { ok=0; log "$job variables differ"; }
    [[ "$(q ".\"$job\".timeout")" == "$jt" ]] || { ok=0; log "$job timeout is $(q ".\"$job\".timeout")"; }
    # The bound is computed from the definition's own values (script + after_script <= job timeout).
    s="$(q ".\"$job\".variables.RUNNER_SCRIPT_TIMEOUT")" a="$(q ".\"$job\".variables.RUNNER_AFTER_SCRIPT_TIMEOUT")" jt="$(q ".\"$job\".timeout")"
    sum=$(( $(dur "$s") + $(dur "$a") ))
    (( $(dur "$jt") > 0 && $(dur "$s") > 0 && $(dur "$a") > 0 && sum <= $(dur "$jt") )) || { ok=0; log "$job script plus after-script timeouts exceed the job timeout"; }
    if grep -qiE '(password|secret|token|api_?key)' <<<"$(q ".\"$job\".variables | keys | join(\" \")" | sed 's/FF_GIT_URLS_WITHOUT_TOKENS//')"; then ok=0; log "$job has a secret-like variable name"; fi
  done
  want='['"$guard"',{"if":"'"$rcif"'","variables":'"$rcvars"'},{"if":"$CI_PIPELINE_SOURCE == \"merge_request_event\"","interruptible":true},{"when":"on_success"}]'
  [[ "$(qj '."validate:checks".rules')" == "$want" ]] || { ok=0; log "checks rules differ (RC promotion with the fetch refspecs, not interruptible; only other MR checks interruptible; mandatory on-success fallback)"; }
  [[ "$(qj '."validate:integration".needs')" == '[{"job":"validate:checks","artifacts":false}]' ]] || { ok=0; log "gate needs differ (fast, without its artifacts)"; }
  want='['"$guard"',{"if":"'"$rcif"'","variables":'"$rcvars"'},{"if":"$CI_PIPELINE_SOURCE == \"merge_request_event\"","when":"manual","allow_failure":true},{"if":"$CI_PIPELINE_SOURCE == \"schedule\""},{"if":"$CI_PIPELINE_SOURCE == \"web\""},{"if":"$CI_PIPELINE_SOURCE == \"api\""}]'
  [[ "$(qj '."validate:integration".rules')" == "$want" ]] || { ok=0; log "gate rules differ (ordered: automatic required RC-promotion gate with the fetch refspecs; optional manual MR gate; schedule/web/api gate; no staging or default-branch push rule)"; }
  # G004 CI-1: the protected identity-staging job (stage integration, no needs, environment identity-staging on every tier).
  job=identity-staging
  [[ "$(q ".\"$job\" | keys | sort | join(\" \")")" == "after_script artifacts environment image interruptible needs rules script services stage tags timeout variables" ]] || { ok=0; log "$job keys differ"; }
  [[ "$(q ".\"$job\".stage")" == integration && "$(q ".\"$job\".interruptible")" == false && "$(qj ".\"$job\".needs")" == '[]' ]] || { ok=0; log "$job stage, interruptible or needs"; }
  [[ "$(qj ".\"$job\".rules")" == '['"$guard"',{"if":"$CI_COMMIT_BRANCH == \"identity-qualification\" && $CI_COMMIT_REF_PROTECTED == \"true\" && ($CI_PIPELINE_SOURCE == \"web\" || $CI_PIPELINE_SOURCE == \"schedule\")"}]' ]] || { ok=0; log "$job rules differ"; }
  [[ "$(qj ".\"$job\".environment")" == '{"name":"identity-staging"}' ]] || { ok=0; log "$job environment is not exactly identity-staging"; }
  [[ "$(q ".\"$job\".image")" == "$cli" && "$(qj ".\"$job\".services")" == "$(qj '."validate:checks".services')" && "$(q ".\"$job\".tags | join(\",\")")" == saas-linux-medium-amd64 ]] || { ok=0; log "$job image, services or tags differ"; }
  [[ "$(q ".\"$job\".script | length")" == 1 && "$(q ".\"$job\".script[0]")" == "sh infra/ci/hosted-context.sh -- --profile staging --publish-dir ci-evidence" ]] || { ok=0; log "$job script differs"; }
  [[ "$(q ".\"$job\".after_script | length")" == 1 && "$(q ".\"$job\".after_script[0]")" == "sh infra/ci/hosted-context.sh --after-script" ]] || { ok=0; log "$job after_script differs"; }
  [[ "$(qj ".\"$job\".artifacts")" == '{"when":"always","expire_in":"90 days","access":"developer","paths":["ci-evidence/"],"reports":{"junit":"ci-evidence/junit.xml"}}' ]] || { ok=0; log "$job artifacts differ"; }
  want='{"DOCKER_HOST":"tcp://docker:2376","DOCKER_TLS_CERTDIR":"/certs","DOCKER_TLS_VERIFY":"1","DOCKER_CERT_PATH":"/certs/client","GIT_STRATEGY":"clone","GIT_DEPTH":"0","GIT_CLEAN_FLAGS":"-ffdx","FF_GIT_URLS_WITHOUT_TOKENS":"true","RUNNER_SCRIPT_TIMEOUT":"40m","RUNNER_AFTER_SCRIPT_TIMEOUT":"4m"}'
  [[ "$(qj ".\"$job\".variables")" == "$want" && "$(q ".\"$job\".timeout")" == 45m ]] || { ok=0; log "$job variables or timeout differ"; }
  for k in allow_failure when only except retry parallel trigger release pages cache extends include dependencies resource_group coverage hooks secrets id_tokens inherit manual_confirmation; do
    [[ "$(q ".\"$job\" | has(\"$k\")")" == false ]] || { ok=0; log "$job has $k"; }
  done
  # G072 (Plan r1 §10.13 AC1): inputs, deployment-mode refusals, guards, the five staging jobs and the launcher's
  # context pins, on the committed file; then every mutation control (one protected behaviour removed) must fail.
  local g72reg="$REPO/infra/staging/ci-environments.txt"
  ci_g072_check "$f" "$g72reg" 0 "$cli" || ok=0
  ci_g072_mutations "$f" "$g72reg" "$cli" || ok=0
  # README blocks equal the documented local commands.
  [[ "$(docs_block "$REPO/README.md" platform:ci)" == "bash infra/ci/validate.sh --profile full" ]] || { ok=0; log "README platform:ci differs"; }
  [[ "$(docs_block "$REPO/README.md" platform:ci-fast)" == "bash infra/ci/validate.sh --profile fast" ]] || { ok=0; log "README platform:ci-fast differs"; }
  [[ "$(docs_block "$REPO/README.md" platform:ci-integration)" == "bash infra/ci/validate.sh --profile integration" ]] || { ok=0; log "README platform:ci-integration differs"; }
  # G051: the local affected profile against the merge base with the remote integration branch.
  [[ "$(docs_block "$REPO/README.md" platform:ci-affected)" == 'bash infra/ci/validate.sh --profile affected --base "$(git merge-base HEAD origin/staging)"' ]] || { ok=0; log "README platform:ci-affected differs"; }
  # Profile and suite consistency: all = the full profile; every TSV suite label is a run.sh suite; fast is a
  # subset of integration; the required IDs per profile are reported. Suite names may be hyphenated (G051).
  for prof in fast integration full staging; do ps[$prof]="$(sed -n "s/^CI_SUITES_$prof=\"\([a-z -]*\)\"\$/\1/p" "$REPO/infra/ci/lib.sh")"; done
  [[ "${ps[staging]}" == staging ]] || { ok=0; log "the staging profile is not exactly the staging suite"; }
  [[ "${ps[full]}" == "${SUITES_ALL[*]}" ]] || { ok=0; log "the full profile is not run.sh all"; }
  [[ -n "${ps[fast]}" && -n "${ps[integration]}" ]] || { ok=0; log "infra/ci/lib.sh profile suites missing"; }
  for s in ${ps[fast]}; do [[ " ${ps[integration]} " == *" $s "* ]] || { ok=0; log "fast suite $s is not in integration"; }; done
  for s in "${SUITES_KNOWN[@]}" final; do known[$s]=1; done
  while IFS=$'\t' read -r id suite; do
    [[ -z "$id" || "$id" == '#'* ]] && continue
    [[ -n "${known[$suite]+x}" ]] || { ok=0; log "required check $id has unknown suite $suite"; }
    for prof in fast integration full staging; do
      [[ " ${ps[$prof]} final " == *" $suite "* ]] && cnt[$prof]=$(( ${cnt[$prof]:-0} + 1 ))
    done
  done <"$REPO/infra/validation/required-checks.tsv"
  [[ $ok == 1 ]] && record B-CI-CONFIG PASS ".gitlab-ci.yml: two documents, the spec:inputs header with exactly action (Validate|Deploy environment|Stop environment, default Validate), environment (none + infra/staging/ci-environments.txt, default none), candidate and duration (empty defaults, Plan §10.2 regexes) and no images input; exactly workflow/stages (checks, integration, staging)/validate:checks/validate:integration/identity-staging/staging:prepare/staging:check/staging:build/staging:deploy/staging:stop; G072: deployment mode refused by the workflow outside web/API pipelines on the protected staging branch, every validation job and identity-staging first refuses a non-default action, the five staging jobs exact (prepare and build without id_tokens/environment/resource group except build's staging-image-build; check with the read audience, environment action prepare, no resource group, dotenv reuse.env; deploy start and stop stop audiences with environment actions start/stop and the resource group staging-env-<env>; stop only with empty candidate and duration; needs prepare -> check -> build -> deploy; timeouts 10m/10m/45m/60m/30m within script+after-script; the pinned CLI/DinD/tag; scripts through infra/ci/staging-action.sh), no STAGING_* variable and no input outside the four, candidate and duration interpolated only inside double quotes; staging-ci.Dockerfile pins the ci-context CLI stage, check.sh's jq and a digest-pinned google-cloud-cli; mutation controls each refused: $G72_MUTS; the fifteen workflow rules (no push pipeline: feature, staging and default-branch pushes never; every tag (RC or final) never; the three G072 deployment-mode refusals; the RC-promotion merge request release/<canonical RC> to the default branch admitted without a changes filter; MR changed-path admission with the documentation Markdown, planning-input JSON, exact Goal scope-allowlist TXT, exact session-tracking JSONL and exact three handoff-file omission and nested Claude-skill admission, each changes list at most 50 entries; default-branch schedule; the protected identity-qualification schedule; web; api branch pipeline; never); identity-staging: stage integration, no needs, rules protected identity-qualification web or schedule only, environment identity-staging, the pinned CLI/DinD and tag, timeout 45m (40m+4m), script --profile staging through infra/ci/hosted-context.sh; jobs validate:checks and validate:integration on saas-linux-medium-amd64 with the pinned CLI image (= ci-context CLI stage) and pinned TLS DinD service, exact variables, scripts through infra/ci/hosted-context.sh, after_script backstop, artifacts when:always/developer/90 days with the supplemental reports:junit ci-evidence/junit.xml (G050; also on identity-staging), timeouts 40m (35m+4m) and 3h (2h45m+10m), interruptible false by default with exact MR-only checks override and workflow auto-cancel policy, no job-level allow_failure/when/retry/parallel/trigger/environment/release/pages/cache/include/resource_group overrides; gate needs fast without its artifacts (ci-evidence/ never enters the gate's checkout) and has ordered rules: automatic required RC-promotion gate, optional manual MR, schedule, web and api (no staging or default-branch push rule: staging qualification is the coordinator's local --profile full on the exact merged staging SHA); both jobs' RC-promotion rule sets only GIT_FETCH_EXTRA_FLAGS (--prune --quiet plus the staging and tag refspecs); README ci, ci-fast, ci-integration and ci-affected blocks; all = full profile; every required-check suite (including the G051 lint-*, unit-* and feature-* suites) a run.sh suite; fast within integration; required IDs fast ${cnt[fast]:-0}, integration ${cnt[integration]:-0}, full ${cnt[full]:-0}, staging ${cnt[staging]:-0}" || record B-CI-CONFIG FAIL "CI definition"
}

# B-CI-CONTROLS (G003 Plan §11.3): infra/ci/validate-controls.sh groups A-E against stub harnesses in
# scratch trees (group E Docker-real with busybox and control labels; V4 runs the real contracts suite).
static_ci_controls() {
  local rc=0 n want got regenv=()
  # G064 Plan A corr1 C4: this run's registry reaches the controls, which mirror their inventories into it (and so do not
  # self-register); without a registry the controls run self-registered.
  if [[ -n "$OWNED_REG" ]]; then regenv=(CI_CONTROLS_OWNED_REG="$OWNED_REG"); fi
  # G051: --log puts the runner's per-control timings (ci-controls-timings.tsv, beside the log) into $WORK.
  env "${regenv[@]}" bash "$REPO/infra/ci/validate-controls.sh" --log "$WORK/ci-controls-summary.log" >"$WORK/ci-controls.log" 2>&1 || rc=$?
  n="$(grep -cE '^PASS ' "$WORK/ci-controls.log" || true)"
  want="$(bash "$REPO/infra/ci/validate-controls.sh" --list 2>/dev/null | awk '{print $2}' | LC_ALL=C sort)"
  got="$(awk '$1 == "PASS" {print $2}' "$WORK/ci-controls.log" | LC_ALL=C sort)"
  if [[ $rc == 0 ]] && ! grep -qE '^FAIL ' "$WORK/ci-controls.log" && [[ "$n" -ge 80 && -n "$want" && "$got" == "$want" ]]; then
    record B-CI-CONTROLS PASS "infra/ci/validate-controls.sh: $n controls pass (groups A-H: gate semantics and profile completeness incl. the real contracts run, attribution, evidence scanning, finalization/supervision/location/budgets/capacity, positive-ownership cleanup, profile selection, hosted adapter, release helper); see ci-controls.log"
  else tail -30 "$WORK/ci-controls.log"; record B-CI-CONTROLS FAIL "CI controls (exit $rc, $n passing)"; fi
}

# ===================================================================================
suite_provenance() {
  log "== provenance"
  # The renderer identity, repeat-plan and repeat-render checks are retired with the SaaS Foundation generator, which is
  # not part of this distribution. The protected files and the skill roots stay guarded against the base commit.
  skill_relocation
  provenance_protected
  # Every skill root: tracked content equals the base, except the paths the reviewed relocations, the retired root and
  # the verified aliases account for; each untracked path (enumerated NUL-separated, since names may contain spaces) is
  # named exactly by an --expect-hash.
  local skill_ok=1 n_untracked=0 e p moved="" retired=""
  local -A bound=()
  for e in "${EXPECT[@]}"; do bound["${e%%=*}"]="${e#*=}"; done
  if [[ -s "$WORK/skill-relocation.txt" ]]; then
    skill_ok=0
    while IFS= read -r e; do log "relocation: $e"; done <"$WORK/skill-relocation.txt"
  fi
  # The listings go through files so a failing git command fails this check instead of yielding an empty list.
  git -C "$REPO" diff -z --name-only --no-renames --ignore-submodules=none "$BASE" -- "${SKILL_ROOTS[@]}" "${SKILL_EXCLUSIONS[@]}" >"$WORK/skill-diff.z" || { skill_ok=0; log "git diff of the skill roots failed"; }
  while IFS= read -r -d '' p; do
    [[ -n "${RELOC_SRC[$p]+x}" || -n "${RELOC_DST[$p]+x}" || -n "${RETIRED_SRC[$p]+x}" ]] && continue
    skill_ok=0; log "skill path differs from the base commit: $p"
  done <"$WORK/skill-diff.z"
  git -C "$REPO" ls-files -z -o --exclude-standard -- "${SKILL_ROOTS[@]}" "${SKILL_EXCLUSIONS[@]}" >"$WORK/skill-untracked.z" || { skill_ok=0; log "git ls-files of the skill roots failed"; }
  while IFS= read -r -d '' p; do
    n_untracked=$((n_untracked + 1))
    [[ -n "${bound[$p]:-}" && "$(sha_of "$REPO/$p")" == "${bound[$p]}" ]] || { skill_ok=0; log "untracked Skill path not bound by a matching --expect-hash: $p"; }
  done <"$WORK/skill-untracked.z"
  (( RELOC_ACTIVE == 0 )) || moved="; relocation verified: ${#RELOC_SRC[@]} base paths at their destinations (same content and mode), no source left, no unbound file"
  (( ${#RETIRED_SRC[@]} == 0 )) || retired="; retired ${SKILL_RETIRED[*]}: ${#RETIRED_SRC[@]} base paths absent, nothing left there"
  [[ $skill_ok == 1 ]] && record B-SKILL-UNCHANGED PASS "skill roots (${SKILL_ROOTS[*]}) identical to $(git -C "$REPO" rev-parse --short "$BASE")$moved$retired; independent Restruct tooling and its alias are excluded; $n_untracked untracked Skill path(s), each bound by a matching --expect-hash" || record B-SKILL-UNCHANGED FAIL "a Skill differs from the base commit or its approved hashes"
}

# Reviewed skill relocations (G049, owner 2026-10-18; the bundled Restruct SDLC skill: skills/restruct-sdlc/ ->
# distributions/skills/restruct-sdlc/). A move applies only while BASE still holds its source tree, so it is inert once
# the relocation is part of the base. Every BASE file under a source must be at its destination with the same mode and
# content; the source must be gone; no unbound file may appear under a destination. A retired root (the SaaS Foundation
# generator, not part of this distribution) applies the same way: every BASE path under it must be absent and nothing
# may remain there. The discovery aliases (the two Restruct SDLC aliases; this template ships no brand skill alias) are
# Restruct and its discovery alias are outside application validation. Other legacy skill moves remain guarded; problems go to
# $WORK/skill-relocation.txt; RELOC_SRC/RELOC_DST/RETIRED_SRC name the verified paths for B-SKILL-UNCHANGED and
# B-PROTECTED, which accepts a removed base path only through this verification.
SKILL_ROOTS=(skills/ .agents/skills/ .claude/skills/ web/.claude/skills/ web/.agents/skills/)
# Independent development tooling is outside application/CI validation, initialized or not.
SKILL_EXCLUSIONS=(":(top,exclude).agents/skills/restruct-sdlc" ":(top,exclude).claude/skills/restruct-sdlc")
SKILL_MOVES=("skills/claude-session-handoff/ .agents/skills/claude-session-handoff/"
  "skills/vestive-design-skill/ .agents/skills/vestive-design/"
  "web/.claude/skills/vestive-design/ .agents/skills/vestive-design/"
  ".claude/skills/vestive-design/ .agents/skills/vestive-design/"
)
SKILL_RETIRED=(skills/saas-foundation/ distributions/)
SKILL_LINK=.claude/skills/vestive-brand SKILL_LINK_TARGET=../../.agents/skills/vestive-brand
RELOC_ACTIVE=0
declare -A RELOC_SRC=() RELOC_DST=() RETIRED_SRC=()
skill_relocation() {
  local pair src dst rec meta path mode obj d want m e names lt wt alias_source r out="$WORK/skill-relocation.txt"
  local -A expected=() bound=()
  for e in "${EXPECT[@]}"; do bound["${e%%=*}"]="${e#*=}"; done
  : >"$out"
  # The retired roots.
  for r in "${SKILL_RETIRED[@]}"; do
    git -C "$REPO" ls-tree -r -z --name-only "$BASE" -- "$r" >"$WORK/retired-base.z" || { echo "git ls-tree of $r failed" >>"$out"; continue; }
    while IFS= read -r -d '' path; do
      if [[ -e "$REPO/$path" || -L "$REPO/$path" ]]; then echo "retired path remains: $path" >>"$out"; else RETIRED_SRC["$path"]=1; fi
    done <"$WORK/retired-base.z"
    git -C "$REPO" ls-files -z --cached --others --exclude-standard -- "$r" >"$WORK/retired-now.z" || { echo "git ls-files of $r failed" >>"$out"; continue; }
    [[ ! -s "$WORK/retired-now.z" ]] || echo "files remain under the retired $r" >>"$out"
    [[ ! -e "$REPO/${r%/}" && ! -L "$REPO/${r%/}" ]] || echo "the retired root $r remains" >>"$out"
  done
  for pair in "${SKILL_MOVES[@]}"; do
    src="${pair%% *}" dst="${pair##* }"
    names="$(git -C "$REPO" ls-tree -r --name-only "$BASE" -- "$src")" || { echo "git ls-tree of $src failed" >>"$out"; continue; }
    [[ -n "$names" ]] || continue
    RELOC_ACTIVE=1
    # Owner-approved reversal: only the exact former canonical root may remain as its new discovery alias.
    alias_source=0
    if [[ "${src%/}" == "$SKILL_LINK" && -L "$REPO/$SKILL_LINK" && "$(readlink "$REPO/$SKILL_LINK")" == "$SKILL_LINK_TARGET" ]]; then alias_source=1; fi
    [[ $alias_source == 1 || ( ! -e "$REPO/${src%/}" && ! -L "$REPO/${src%/}" ) ]] || echo "source remains: $src" >>"$out"
    while IFS= read -r -d '' rec; do
      meta="${rec%%$'\t'*}" path="${rec#*$'\t'}"
      mode="${meta%% *}" obj="${meta##* }" d="$dst${path#"$src"}"
      expected["$d"]=1
      if [[ -L "$REPO/$d" || ! -f "$REPO/$d" ]]; then echo "missing destination: $d" >>"$out"; continue; fi
      # A reviewed content change of a moved file is accepted only at its exact --expect-hash.
      want="${bound[$d]:-$(git -C "$REPO" cat-file blob "$obj" | sha256sum | cut -d' ' -f1)}"
      if [[ -x "$REPO/$d" ]]; then m=100755; else m=100644; fi
      if [[ "$(sha_of "$REPO/$d")" != "$want" || "$m" != "$mode" ]]; then echo "destination differs: $d" >>"$out"; continue; fi
      if [[ $alias_source == 1 || ( ! -e "$REPO/$path" && ! -L "$REPO/$path" ) ]]; then RELOC_SRC["$path"]=1; fi
      RELOC_DST["$d"]=1
    done < <(git -C "$REPO" ls-tree -r -z "$BASE" -- "$src" || echo "git ls-tree of $src failed" >>"$out")
    git -C "$REPO" ls-files -z --cached --others --exclude-standard -- "$dst" >"$WORK/reloc-dst.z" || echo "git ls-files of $dst failed" >>"$out"
    while IFS= read -r -d '' e; do
      [[ -n "${expected[$e]+x}" ]] || echo "unbound file under $dst: $e" >>"$out"
    done <"$WORK/reloc-dst.z"
  done
  (( RELOC_ACTIVE )) || return 0
  lt="$(realpath -e "$REPO/$SKILL_LINK" 2>/dev/null)" || lt=""
  wt="$(realpath -e "$REPO/$VESTIVE_SKILL" 2>/dev/null)" || wt=""
  if [[ -L "$REPO/$SKILL_LINK" && "$(readlink "$REPO/$SKILL_LINK")" == "$SKILL_LINK_TARGET" && -n "$lt" && "$lt" == "$wt" ]]; then
    RELOC_DST["$SKILL_LINK"]=1
    # The previous Codex alias becomes the verified regular canonical directory.
    if [[ "$(git -C "$REPO" ls-tree "$BASE" -- "$VESTIVE_SKILL" | cut -d' ' -f1)" == 120000 && "$(git -C "$REPO" show "$BASE:$VESTIVE_SKILL")" == ../../.claude/skills/vestive-design ]]; then
      RELOC_SRC["$VESTIVE_SKILL"]=1
    fi
  else
    echo "the alias $SKILL_LINK is not exactly the relative link to $SKILL_LINK_TARGET" >>"$out"
  fi
  git -C "$REPO" ls-files -z --cached --others --exclude-standard -- web/.agents/skills/ >"$WORK/reloc-alias.z" || echo "git ls-files of web/.agents/skills/ failed" >>"$out"
  while IFS= read -r -d '' e; do
    [[ "$e" == "$SKILL_LINK" ]] || echo "unbound path under web/.agents/skills/: $e" >>"$out"
  done <"$WORK/reloc-alias.z"
}

provenance_protected() {
  local ok=1 p want n=0 e path sha nreloc=0 nret=0
  local -A expected=()
  # (paths are compared exactly; ls-tree -z keeps names with spaces intact)
  for e in "${EXPECT[@]}"; do expected["${e%%=*}"]="${e#*=}"; done
  while IFS= read -r p; do
    n=$((n + 1))
    if [[ -n "${expected[$p]:-}" ]]; then continue; fi
    # G049: a base Skill path is accepted as removed only through the verified reviewed relocation or retirement.
    if [[ -n "${RELOC_SRC[$p]+x}" ]]; then nreloc=$((nreloc + 1)); continue; fi
    if [[ -n "${RETIRED_SRC[$p]+x}" ]]; then nret=$((nret + 1)); continue; fi
    [[ -f "$REPO/$p" ]] || { ok=0; log "protected file removed: $p"; continue; }
    want="$(git -C "$REPO" rev-parse "$BASE:$p")"
    [[ "$(git -C "$REPO" hash-object "$p")" == "$want" ]] || { ok=0; log "protected file changed: $p"; }
  done < <(git -C "$REPO" ls-tree -r -z --name-only "$BASE" | tr '\0' '\n' | grep -E '^(AGENTS\.md|CLAUDE\.md|docs/|management/|skills/|distributions/skills/)' | grep -vE '^docs/foundation/')
  for path in "${!expected[@]}"; do
    sha="${expected[$path]}"
    [[ -f "$REPO/$path" && "$(sha_of "$REPO/$path")" == "$sha" ]] || { ok=0; log "$path does not match its expected hash"; }
  done
  [[ $ok == 1 ]] && record B-PROTECTED PASS "$n protected files (AGENTS.md, CLAUDE.md, docs/ except docs/foundation, management/, skills/, distributions/skills/) equal the base commit, $nreloc of them relocated through the verified mapping and $nret removed through the verified retirement; ${#expected[@]} expected-hash file(s) match" || record B-PROTECTED FAIL "protected content changed"
}

# ===================================================================================
# G003 test tooling images (built once per run, run-labelled, removed at cleanup).
VT_IMG="" VT_OK=1 ICHECK_IMG=""
ensure_tools() {
  [[ -n "$VT_IMG" ]] && return "$((1 - VT_OK))"
  VT_IMG="g003-vtools-$RUN_ID"; ICHECK_IMG="g003-intelligence-check-$RUN_ID"
  # G051: explicit builds go through the build ledger (ledger_build: plain progress, an iidfile and one row each).
  if ledger_build "$LEDGER_SUITE" authoritative -- --label g001.run=$RUN_ID -f "$REPO/infra/docker/validation-tools.Dockerfile" --target runtime -t "$VT_IMG" "$REPO/infra/validation/tools" >"$WORK/vtools-build.log" 2>&1; then
    record_built_image "$VT_IMG"
  else VT_OK=0; tail -20 "$WORK/vtools-build.log"; fi
  if ledger_build "$LEDGER_SUITE" authoritative -- --label g001.run=$RUN_ID -f "$REPO/infra/docker/intelligence.Dockerfile" --target check --build-context contracts="$REPO/contracts" -t "$ICHECK_IMG" "$REPO/intelligence" >"$WORK/icheck-build.log" 2>&1; then
    record_built_image "$ICHECK_IMG"
  else VT_OK=0; tail -20 "$WORK/icheck-build.log"; fi
  return "$((1 - VT_OK))"
}
# vtools ARGS...: the validation tools image with no network and the repository's contracts read-only.
vtools() { docker_run_owned vtools --rm --label g001.run=$RUN_ID --network none -v "$REPO/contracts:/contracts:ro" "$@"; }
# pyconf ARGS...: the Python conformance exchange script inside the intelligence check image.
pyconf() { docker_run_owned pyconf --rm --label g001.run=$RUN_ID --network none "$@"; }

# corpus_digest DIR: the shared corpus_sha256 definition (contracts/conformance/README.md).
corpus_digest() { (cd "$1" && find . -type f -printf '%P\n' | LC_ALL=C sort | xargs -d '\n' sha256sum | sha256sum | cut -d' ' -f1); }

suite_contracts() {
  log "== contracts (root only; no stack)"
  local C="$WORK/contracts" X="$WORK/contracts-exchange" ok rc f id corpus="$REPO/contracts/conformance/intelligence" vt_check="g003-vtools-check-$RUN_ID"
  mkdir -p "$C" "$X"; chmod 777 "$C" "$X"

  # B-CONTRACT-TOOLS
  ok=1
  if ledger_build "$LEDGER_SUITE" authoritative -- --label g001.run=$RUN_ID -f "$REPO/infra/docker/validation-tools.Dockerfile" --target check -t "$vt_check" "$REPO/infra/validation/tools" >"$WORK/vtools-check-build.log" 2>&1; then
    record_built_image "$vt_check"
    docker_run_owned vtools-check --rm --label g001.run=$RUN_ID --network none "$vt_check" >"$WORK/vtools-check.log" 2>&1 || { ok=0; tail -30 "$WORK/vtools-check.log"; }
  else ok=0; tail -20 "$WORK/vtools-check-build.log"; fi
  ensure_tools || ok=0
  local irt="g003-intelligence-rt-$RUN_ID"
  if ledger_build "$LEDGER_SUITE" authoritative -- --label g001.run=$RUN_ID -f "$REPO/infra/docker/intelligence.Dockerfile" --target runtime --build-context contracts="$REPO/contracts" -t "$irt" "$REPO/intelligence" >"$WORK/irt-build.log" 2>&1; then
    record_built_image "$irt"
    docker_run_owned irt-check --rm --label g001.run=$RUN_ID --network none --entrypoint sh "$irt" -c 'test ! -e /srv/conformance && test -f /srv/contracts/intelligence/execution-envelope.schema.json' || { ok=0; log "the runtime intelligence image contains /srv/conformance or lacks the schemas"; }
  else ok=0; tail -20 "$WORK/irt-build.log"; fi
  [[ $ok == 1 ]] && record B-CONTRACT-TOOLS PASS "validation-tools check stage (gofmt, vet, tests) passes; validation-tools runtime and intelligence check images build; the runtime intelligence image has no /srv/conformance" || { record B-CONTRACT-TOOLS FAIL "contract tooling"; }
  if [[ $VT_OK != 1 ]]; then
    for id in B-CONTRACT-GO B-CONTRACT-PY B-CONTRACT-AGREE B-CONTRACT-PEER B-CONTRACT-NEGATIVE; do record "$id" FAIL "contract tooling unavailable"; done
    return
  fi

  # B-CONTRACT-GO and B-CONTRACT-PY (normal mode; digests first; every expectation matched; nothing skipped)
  rc=0; vtools -v "$C:/out" "$VT_IMG" contractcheck corpus --schemas /contracts/intelligence --corpus /contracts/conformance/intelligence --out /out/results-go.json >"$WORK/contracts-go.log" 2>&1 || rc=$?
  contract_result_ok "$C/results-go.json" && [[ $rc == 0 ]] && record B-CONTRACT-GO PASS "contractcheck (santhosh-tekuri/jsonschema v6.0.3, Draft 2020-12, format assertion, no remote loader): $(yq_stdin -p json -o tsv '.counts.total' <"$C/results-go.json") cases match their declared expectations; schema digests pinned; 0 skipped" || { tail -20 "$WORK/contracts-go.log"; record B-CONTRACT-GO FAIL "Go contract conformance (exit $rc)"; }
  rc=0; pyconf -v "$C:/out" "$ICHECK_IMG" python tests/conformance_exchange.py corpus --schemas /srv/contracts/intelligence --corpus /srv/conformance/intelligence --out /out/results-python.json >"$WORK/contracts-py.log" 2>&1 || rc=$?
  contract_result_ok "$C/results-python.json" && [[ $rc == 0 ]] && record B-CONTRACT-PY PASS "streaming.schemas validator (ECMA-262 pattern anchoring, RFC 3339 date-time checker, strict JSON): $(yq_stdin -p json -o tsv '.counts.total' <"$C/results-python.json") cases match their declared expectations; schema digests pinned; 0 skipped" || { tail -20 "$WORK/contracts-py.log"; record B-CONTRACT-PY FAIL "Python contract conformance (exit $rc)"; }

  # B-CONTRACT-AGREE: identical verdicts, equal to the declarations; all eight event types; the file set
  # equals the declared entries; the corpus digest agrees across both validators and this harness.
  ok=1
  local exp="$WORK/contracts-expected.tsv" got_go="$WORK/contracts-go.tsv" got_py="$WORK/contracts-py.tsv" digest
  yq_stdin -p json -o tsv '.cases[] | [.id, .valid, .class]' <"$corpus/expectations.json" | LC_ALL=C sort >"$exp" || ok=0
  yq_stdin -p json -o tsv '.results[] | [.id, .valid, .class]' <"$C/results-go.json" 2>/dev/null | LC_ALL=C sort >"$got_go" || ok=0
  yq_stdin -p json -o tsv '.results[] | [.id, .valid, .class]' <"$C/results-python.json" 2>/dev/null | LC_ALL=C sort >"$got_py" || ok=0
  [[ -s "$exp" ]] && cmp -s "$exp" "$got_go" && cmp -s "$exp" "$got_py" || { ok=0; log "verdict sets differ: $(diff "$got_go" "$got_py" | head -5 | tr '\n' ' ')"; }
  [[ "$(wc -l <"$exp")" == 120 ]] || { ok=0; log "expected 120 declared cases, found $(wc -l <"$exp")"; }
  for f in "$C/results-go.json" "$C/results-python.json"; do
    [[ "$(yq_stdin -p json -o tsv '.counts.event_types | sort | join(" ")' <"$f")" == "citation done error structured_output text_delta tool_call tool_result usage" ]] || { ok=0; log "$(basename "$f"): not all eight event types among valid cases"; }
  done
  { yq_stdin -p json -o tsv '.cases[].file' <"$corpus/expectations.json"
    yq_stdin -p json -o tsv '.schema_negative[] | select(.mutation.op != "absent") | .file' <"$corpus/expectations.json"; echo expectations.json; } | LC_ALL=C sort >"$WORK/contracts-declared-files.txt"
  (cd "$corpus" && find . -type f -printf '%P\n' | LC_ALL=C sort) >"$WORK/contracts-actual-files.txt"
  cmp -s "$WORK/contracts-declared-files.txt" "$WORK/contracts-actual-files.txt" || { ok=0; log "corpus files differ from the declared entries: $(diff "$WORK/contracts-declared-files.txt" "$WORK/contracts-actual-files.txt" | head -5 | tr '\n' ' ')"; }
  digest="$(corpus_digest "$corpus")"
  [[ "$(yq_stdin -p json -o tsv '.corpus_sha256' <"$C/results-go.json")" == "$digest" && "$(yq_stdin -p json -o tsv '.corpus_sha256' <"$C/results-python.json")" == "$digest" ]] || { ok=0; log "corpus digests differ (harness $digest)"; }
  cp "$exp" "$WORK/contracts-agree.tsv"
  [[ $ok == 1 ]] && record B-CONTRACT-AGREE PASS "Go and Python verdicts identical and equal to the 120 declared expectations; all eight event types among valid cases; the corpus file set equals the declared entries; corpus digest $(cut -c1-16 <<<"$digest") agrees across both validators and the harness" || record B-CONTRACT-AGREE FAIL "cross-language agreement"

  contract_peer "$X"
  contract_negative "$C"
  cp "$C"/results-go.json "$WORK/contracts-results-go.json" 2>/dev/null || true
  cp "$C"/results-python.json "$WORK/contracts-results-python.json" 2>/dev/null || true
  cp "$C"/schema-negative-go.json "$C"/schema-negative-python.json "$WORK/" 2>/dev/null || true
}

# contract_result_ok FILE: a shared-format result file reporting 0 skipped and pinned schema digests.
contract_result_ok() {
  [[ -s "$1" ]] || return 1
  [[ "$(yq_stdin -p json -o tsv '.skipped' <"$1")" == 0 ]] || return 1
  [[ "$(yq_stdin -p json -o tsv '.schema_sha256["execution-envelope"]' <"$1")" == bf3cec365e291da2cf9b201b9d0194632818721671a5844589954b978accba4f ]] || return 1
  [[ "$(yq_stdin -p json -o tsv '.schema_sha256["stream-event"]' <"$1")" == e1025031134253ac8d19a514f6d7a4d9563281b5006ebe9be40a84e4d63fd102 ]] || return 1
  [[ "$(yq_stdin -p json -o tsv '.counts.total' <"$1")" == 120 ]]
}

# B-CONTRACT-PEER (§8.5): Go envelopes accepted and executed by Python; real Python NDJSON accepted by Go;
# four corruptions of each real file rejected by the peer checker.
contract_peer() {
  local X="$1" ok=1 rc kind src
  vtools -v "$X:/x" "$VT_IMG" contractcheck produce-envelopes --out /x/go-envelopes.ndjson >"$WORK/peer-produce.log" 2>&1 || { ok=0; tail -10 "$WORK/peer-produce.log"; }
  pyconf -v "$X:/x" "$ICHECK_IMG" python tests/conformance_exchange.py peer --schemas /srv/contracts/intelligence --corpus /srv/conformance/intelligence --envelopes /x/go-envelopes.ndjson --out /x/python-events.ndjson >"$WORK/peer-python.log" 2>&1 || { ok=0; tail -10 "$WORK/peer-python.log"; }
  vtools -v "$X:/x:ro" "$VT_IMG" contractcheck ndjson --schemas /contracts/intelligence --corpus /contracts/conformance/intelligence --file /x/python-events.ndjson --as event --require-types text_delta,usage,error,done >"$WORK/peer-go-ndjson.log" 2>&1 || { ok=0; log "Go rejected the real Python NDJSON"; tail -10 "$WORK/peer-go-ndjson.log"; }
  pyconf -v "$X:/x:ro" "$ICHECK_IMG" python tests/conformance_exchange.py ndjson --schemas /srv/contracts/intelligence --corpus /srv/conformance/intelligence --file /x/go-envelopes.ndjson --as envelope >"$WORK/peer-py-ndjson.log" 2>&1 || { ok=0; log "Python rejected the real Go envelopes"; tail -10 "$WORK/peer-py-ndjson.log"; }
  # Corrupt copies: the last line truncated, a garbage line inserted, an invalid record appended, a blank line.
  for src in python-events go-envelopes; do
    [[ -s "$X/$src.ndjson" ]] || { ok=0; continue; }
    head -c -2 "$X/$src.ndjson" >"$X/$src-truncated.ndjson"
    { head -n 1 "$X/$src.ndjson"; echo 'not json'; tail -n +2 "$X/$src.ndjson"; } >"$X/$src-garbage.ndjson"
    { cat "$X/$src.ndjson"; echo '{"type":"unknown","request_id":"req-corrupt-0001"}'; } >"$X/$src-invalid.ndjson"
    { head -n 1 "$X/$src.ndjson"; echo; tail -n +2 "$X/$src.ndjson"; } >"$X/$src-blank.ndjson"
    for kind in truncated garbage invalid blank; do
      rc=0
      if [[ "$src" == python-events ]]; then
        vtools -v "$X:/x:ro" "$VT_IMG" contractcheck ndjson --schemas /contracts/intelligence --corpus /contracts/conformance/intelligence --file "/x/$src-$kind.ndjson" --as event >>"$WORK/peer-corrupt.log" 2>&1 || rc=$?
      else
        pyconf -v "$X:/x:ro" "$ICHECK_IMG" python tests/conformance_exchange.py ndjson --schemas /srv/contracts/intelligence --corpus /srv/conformance/intelligence --file "/x/$src-$kind.ndjson" --as envelope >>"$WORK/peer-corrupt.log" 2>&1 || rc=$?
      fi
      [[ $rc != 0 ]] || { ok=0; log "the peer checker accepted $src-$kind"; }
    done
  done
  [[ $ok == 1 ]] && record B-CONTRACT-PEER PASS "Go-produced envelopes accepted and executed by Python (fake gateway); the real Python NDJSON (success, gateway error, capability not configured) accepted by Go with text_delta, usage, error and done; truncated, garbage, invalid and blank-line corruptions of each file rejected by the peer checker" || record B-CONTRACT-PEER FAIL "peer exchange"
}

# B-CONTRACT-NEGATIVE (§8.2a): audit classes SN01-SN07, positive audit controls, sn-derive and the SD01/SD02
# digest controls in both languages; then EI39 flipped in a scratch corpus fails both validators, naming it.
contract_negative() {
  local C="$1" ok=1 rc f S="$WORK/contracts-ei39"
  vtools -v "$C:/out" "$VT_IMG" contractcheck schema-negative --schemas /contracts/intelligence --corpus /contracts/conformance/intelligence --out /out/schema-negative-go.json >"$WORK/sn-go.log" 2>&1 || { ok=0; tail -10 "$WORK/sn-go.log"; }
  pyconf -v "$C:/out" "$ICHECK_IMG" python tests/conformance_exchange.py schema-negative --schemas /srv/contracts/intelligence --corpus /srv/conformance/intelligence --out /out/schema-negative-python.json >"$WORK/sn-py.log" 2>&1 || { ok=0; tail -10 "$WORK/sn-py.log"; }
  for f in "$C/schema-negative-go.json" "$C/schema-negative-python.json"; do
    [[ -s "$f" ]] || { ok=0; continue; }
    [[ "$(yq_stdin -p json -o tsv '[.results[] | select(.match != true)] | length' <"$f")" == 0 ]] || { ok=0; log "$(basename "$f"): a class differs from its declaration"; }
    [[ "$(yq_stdin -p json -o tsv '[.results[].id] | sort | join(" ")' <"$f")" == "SN01 SN02 SN03 SN04 SN05 SN06 SN07" ]] || { ok=0; log "$(basename "$f"): SN01-SN07 not all classified"; }
    [[ "$(yq_stdin -p json -o tsv '[.digest_controls[] | select(.class == "schema-digest-mismatch" and .evaluated == 0)] | length' <"$f")" == 2 ]] || { ok=0; log "$(basename "$f"): SD01/SD02 not digest mismatches with 0 cases"; }
    [[ "$(yq_stdin -p json -o tsv '[.audit_positive[] | select(.class == "ok")] | length' <"$f")" == 2 ]] || { ok=0; log "$(basename "$f"): positive audit control"; }
  done
  [[ "$(yq_stdin -p json -o tsv '.derive' <"$C/schema-negative-go.json" 2>/dev/null)" == pass ]] || { ok=0; log "sn-derive did not pass"; }
  # EI39 flipped to valid in a scratch corpus copy: both validators must exit nonzero and name EI39.
  mkdir -p "$S"; cp -a "$REPO/contracts/conformance/intelligence/." "$S/"
  yq_stdin -p json -o json '(.cases[] | select(.id == "EI39") | .valid) = true | (.cases[] | select(.id == "EI39") | .class) = "valid"' <"$REPO/contracts/conformance/intelligence/expectations.json" >"$S/expectations.json"
  rc=0; docker_run_owned contractcheck --rm --label g001.run=$RUN_ID --network none -v "$REPO/contracts:/contracts:ro" -v "$S:/scratch:ro" "$VT_IMG" contractcheck corpus --schemas /contracts/intelligence --corpus /scratch --out /dev/null >"$WORK/ei39-go.log" 2>&1 || rc=$?
  [[ $rc != 0 ]] && grep -q 'EI39' "$WORK/ei39-go.log" || { ok=0; log "Go did not fail on EI39 (exit $rc)"; }
  rc=0; pyconf -v "$S:/scratch:ro" "$ICHECK_IMG" python tests/conformance_exchange.py corpus --schemas /srv/contracts/intelligence --corpus /scratch --out /dev/null >"$WORK/ei39-py.log" 2>&1 || rc=$?
  [[ $rc != 0 ]] && grep -q 'EI39' "$WORK/ei39-py.log" || { ok=0; log "Python did not fail on EI39 (exit $rc)"; }
  rm -rf "$S"
  [[ $ok == 1 ]] && record B-CONTRACT-NEGATIVE PASS "both languages, no network: SN01-SN07 audit classes as declared, both canonical schemas audit ok, sn-derive passes (single declared mutation each), SD01/SD02 rejected as schema-digest-mismatch with 0 cases evaluated; EI39 flipped to valid makes both validators exit nonzero naming EI39" || record B-CONTRACT-NEGATIVE FAIL "schema-negative or gate control"
}

# ===================================================================================
suite_root() {
  log "== root (actual repository, project $ROOT_PROJECT)"
  local busy=""
  [[ -e "$REPO/.env" ]] && busy="an existing .env"
  if docker ps -a -q --filter "label=com.docker.compose.project=$ROOT_PROJECT" | grep -q . ||
     docker volume ls -q --filter "label=com.docker.compose.project=$ROOT_PROJECT" | grep -q .; then busy+="${busy:+, }existing Compose resources of project $ROOT_PROJECT"; fi
  ports_busy "${DEFAULT_PORTS[@]}" && busy+="${busy:+, }busy default host ports"
  if [[ -n "$busy" ]]; then record B-ROOT UNMET "refusing to touch $busy; nothing was changed"; return; fi
  register_project "$ROOT_PROJECT"

  local ok_docs=1
  if run_docs_block "$REPO" README.md platform:configure "$WORK/root-configure.log"; then
    ROOT_ENV_SHA="$(sha_of "$REPO/.env")"; remember_secrets "$REPO"; registry_row rootenv "$ROOT_ENV_SHA"
  else ok_docs=0; cat "$WORK/root-configure.log"; record B-DOCS-EXEC-ROOT FAIL "configure block"; return; fi
  local ok=1
  [[ "$(stat -c %a "$REPO/.env")" == 600 ]] || { ok=0; log ".env mode $(stat -c %a "$REPO/.env")"; }
  if (cd "$REPO" && infra/local/configure-env.sh) >"$WORK/root-configure-again.log" 2>&1; then ok=0; log "a second configure run succeeded"; fi
  [[ "$(sha_of "$REPO/.env")" == "$ROOT_ENV_SHA" ]] || { ok=0; log "a second configure run changed .env"; }
  if grep -E '^[A-Z0-9_]*(PASSWORD|SECRET|TOKEN)[A-Z0-9_]*=$' "$REPO/.env"; then ok=0; log "configure left a required secret empty"; fi
  [[ "$(sed -n 's/^OPENROUTER_API_KEY=//p' "$REPO/.env")" == "" ]] || { ok=0; log "optional gateway key was filled"; }
  if grep -F -f "$SECRETS_FILE" "$WORK/root-configure.log" >/dev/null; then ok=0; log "configure printed a secret"; fi
  [[ $ok == 1 ]] && record B-CONFIGURE PASS "configure block creates .env (mode 600) with every required secret generated and the optional gateway key empty; prints no secret; a second run refuses and leaves .env unchanged" || record B-CONFIGURE FAIL "configure helper"

  if ! run_docs_block "$REPO" README.md platform:start "$WORK/root-start.log"; then
    tail -40 "$WORK/root-start.log"; record B-DOCS-EXEC-ROOT FAIL "start block"; return
  fi
  if run_docs_block "$REPO" README.md platform:migrate "$WORK/root-migrate.log" &&
     grep -qP '^migration\t00001_baseline\.sql\tpending' "$WORK/root-migrate.log" &&
     grep -qP '^migration\t00002_identity\.sql\tpending' "$WORK/root-migrate.log" &&
     grep -qP '^migration\t00003_tenancy\.sql\tpending' "$WORK/root-migrate.log" &&
     grep -qP '^migration\t00004_business_profiles\.sql\tpending' "$WORK/root-migrate.log" &&
     grep -qP '^migration\t00005_organization_policy\.sql\tpending' "$WORK/root-migrate.log" &&
     grep -qP '^migration\t00006_row_level_security\.sql\tpending' "$WORK/root-migrate.log" &&
     grep -qP '^migration\t00007_stored_objects\.sql\tpending' "$WORK/root-migrate.log" &&
     grep -qP '^migration\t00008_audit_events\.sql\tpending' "$WORK/root-migrate.log" &&
     grep -qP '^migration\t00009_retention_maintenance\.sql\tpending' "$WORK/root-migrate.log" &&
     grep -qP '^migration\t00010_mcp_connections\.sql\tpending' "$WORK/root-migrate.log" &&
     grep -qP '^migration\t00011_organization_domains\.sql\tpending' "$WORK/root-migrate.log" &&
     grep -qP '^migration\t00012_session_binding_and_claim_exclusivity\.sql\tpending' "$WORK/root-migrate.log" &&
     grep -qP '^migration\t00013_organization_branding\.sql\tpending' "$WORK/root-migrate.log" &&
     grep -qP '^migration\t00014_group_onboarding\.sql\tpending' "$WORK/root-migrate.log" &&
     grep -qP '^migration\t00015_membership_invitations\.sql\tpending' "$WORK/root-migrate.log" &&
     grep -qP '^migration\t00016_organization_enrollment\.sql\tpending' "$WORK/root-migrate.log" &&
     grep -qP '^migration\t00017_group_authority\.sql\tpending' "$WORK/root-migrate.log" &&
     grep -qP '^migration\t00018_platform_signup\.sql\tpending' "$WORK/root-migrate.log" &&
     grep -qP '^migration\t00019_mcp_connection_names\.sql\tpending' "$WORK/root-migrate.log" &&
     grep -qP '^migration\t00020_labels\.sql\tpending' "$WORK/root-migrate.log" &&
     grep -qP '^migration\t00021_platform_labels\.sql\tpending' "$WORK/root-migrate.log" &&
     grep -qP '^migration\t00022_default_tenant\.sql\tpending' "$WORK/root-migrate.log" &&
     grep -qP '^applied\t00001_baseline\.sql$' "$WORK/root-migrate.log" &&
     grep -qP '^applied\t00002_identity\.sql$' "$WORK/root-migrate.log" &&
     grep -qP '^applied\t00003_tenancy\.sql$' "$WORK/root-migrate.log" &&
     grep -qP '^applied\t00004_business_profiles\.sql$' "$WORK/root-migrate.log" &&
     grep -qP '^applied\t00005_organization_policy\.sql$' "$WORK/root-migrate.log" &&
     grep -qP '^applied\t00006_row_level_security\.sql$' "$WORK/root-migrate.log" &&
     grep -qP '^applied\t00007_stored_objects\.sql$' "$WORK/root-migrate.log" &&
     grep -qP '^applied\t00008_audit_events\.sql$' "$WORK/root-migrate.log" &&
     grep -qP '^applied\t00009_retention_maintenance\.sql$' "$WORK/root-migrate.log" &&
     grep -qP '^applied\t00010_mcp_connections\.sql$' "$WORK/root-migrate.log" &&
     grep -qP '^applied\t00011_organization_domains\.sql$' "$WORK/root-migrate.log" &&
     grep -qP '^applied\t00012_session_binding_and_claim_exclusivity\.sql$' "$WORK/root-migrate.log" &&
     grep -qP '^applied\t00013_organization_branding\.sql$' "$WORK/root-migrate.log" &&
     grep -qP '^applied\t00014_group_onboarding\.sql$' "$WORK/root-migrate.log" &&
     grep -qP '^applied\t00015_membership_invitations\.sql$' "$WORK/root-migrate.log" &&
     grep -qP '^applied\t00016_organization_enrollment\.sql$' "$WORK/root-migrate.log" &&
     grep -qP '^applied\t00017_group_authority\.sql$' "$WORK/root-migrate.log" &&
     grep -qP '^applied\t00018_platform_signup\.sql$' "$WORK/root-migrate.log" &&
     grep -qP '^applied\t00019_mcp_connection_names\.sql$' "$WORK/root-migrate.log" &&
     grep -qP '^applied\t00020_labels\.sql$' "$WORK/root-migrate.log" &&
     grep -qP '^applied\t00021_platform_labels\.sql$' "$WORK/root-migrate.log" &&
     grep -qP '^applied\t00022_default_tenant\.sql$' "$WORK/root-migrate.log" &&
     grep -qP '^migration\t00001_baseline\.sql\tapplied\t' "$WORK/root-migrate.log" &&
     grep -qP '^migration\t00002_identity\.sql\tapplied\t' "$WORK/root-migrate.log" &&
     grep -qP '^migration\t00003_tenancy\.sql\tapplied\t' "$WORK/root-migrate.log" &&
     grep -qP '^migration\t00004_business_profiles\.sql\tapplied\t' "$WORK/root-migrate.log" &&
     grep -qP '^migration\t00005_organization_policy\.sql\tapplied\t' "$WORK/root-migrate.log" &&
     grep -qP '^migration\t00006_row_level_security\.sql\tapplied\t' "$WORK/root-migrate.log" &&
     grep -qP '^migration\t00007_stored_objects\.sql\tapplied\t' "$WORK/root-migrate.log" &&
     grep -qP '^migration\t00008_audit_events\.sql\tapplied\t' "$WORK/root-migrate.log" &&
     grep -qP '^migration\t00009_retention_maintenance\.sql\tapplied\t' "$WORK/root-migrate.log" &&
     grep -qP '^migration\t00010_mcp_connections\.sql\tapplied\t' "$WORK/root-migrate.log" &&
     grep -qP '^migration\t00011_organization_domains\.sql\tapplied\t' "$WORK/root-migrate.log" &&
     grep -qP '^migration\t00012_session_binding_and_claim_exclusivity\.sql\tapplied\t' "$WORK/root-migrate.log" &&
     grep -qP '^migration\t00013_organization_branding\.sql\tapplied\t' "$WORK/root-migrate.log" &&
     grep -qP '^migration\t00014_group_onboarding\.sql\tapplied\t' "$WORK/root-migrate.log" &&
     grep -qP '^migration\t00015_membership_invitations\.sql\tapplied\t' "$WORK/root-migrate.log" &&
     grep -qP '^migration\t00016_organization_enrollment\.sql\tapplied\t' "$WORK/root-migrate.log" &&
     grep -qP '^migration\t00017_group_authority\.sql\tapplied\t' "$WORK/root-migrate.log" &&
     grep -qP '^migration\t00018_platform_signup\.sql\tapplied\t' "$WORK/root-migrate.log" &&
     grep -qP '^migration\t00019_mcp_connection_names\.sql\tapplied\t' "$WORK/root-migrate.log" &&
     grep -qP '^migration\t00020_labels\.sql\tapplied\t' "$WORK/root-migrate.log" &&
     grep -qP '^migration\t00021_platform_labels\.sql\tapplied\t' "$WORK/root-migrate.log" &&
     grep -qP '^migration\t00022_default_tenant\.sql\tapplied\t' "$WORK/root-migrate.log" &&
     [[ "$(grep -P '^current_version\t' "$WORK/root-migrate.log" | tr '\n' ' ')" == $'current_version\t0 current_version\t22 current_version\t22 ' ]]; then :
  else ok_docs=0; tail -30 "$WORK/root-migrate.log"; log "migrate block"; fi

  local core_port web_port
  core_port="$(env_get "$REPO" CORE_PORT)"; web_port="$(env_get "$REPO" WEB_PORT)"
  root_defaults
  probe_ports ROOT "$REPO" "${DEFAULT_PORTS[@]}"
  boundaries ROOT "$REPO" yes
  boundary_controls "$REPO"
  mcp_http "$REPO" "$core_port" "$web_port"
  tenancy_http "$REPO" "$core_port" "$web_port"
  web_http "$web_port"
  health_root "$web_port" "$core_port" "$(env_get "$REPO" INTELLIGENCE_PORT)"
  browser_root
  if run_docs_block "$REPO" README.md platform:check "$WORK/root-check.log"; then
    record B-CHECKS-ROOT PASS "README check block: core-check (gofmt, vet, unit tests incl. MCP and migration units), web-check, intelligence-check, secret-scan"
  else tail -40 "$WORK/root-check.log"; record B-CHECKS-ROOT FAIL "README check block"; ok_docs=0; fi
  # Error output of the real artifacts (G003 §10), while the root images exist.
  core_error_output
  web_error_output
  intel_error_output
  save_logs "$REPO" root
  run_docs_block "$REPO" README.md platform:stop "$WORK/root-stop.log" || { ok_docs=0; log "stop block"; }

  local gs=docs/foundation/GETTING-STARTED.md ok_gs=1
  run_docs_block "$REPO" "$gs" archetype:start "$WORK/root-gs-start.log" || { ok_gs=0; tail -20 "$WORK/root-gs-start.log"; }
  run_docs_block "$REPO" "$gs" archetype:check "$WORK/root-gs-check.log" || { ok_gs=0; tail -20 "$WORK/root-gs-check.log"; }
  save_logs "$REPO" root-gs
  run_docs_block "$REPO" "$gs" archetype:stop "$WORK/root-gs-stop.log" || ok_gs=0
  [[ $ok_gs == 1 ]] && record B-DOCS-EXEC-GS PASS "GETTING-STARTED start, check and stop blocks executed verbatim in the actual repository" || record B-DOCS-EXEC-GS FAIL "GETTING-STARTED blocks"

  build_context
  run_docs_block "$REPO" README.md platform:reset "$WORK/root-reset.log" || { ok_docs=0; log "reset block"; }
  if docker volume ls -q --filter "label=com.docker.compose.project=$ROOT_PROJECT" | grep -q .; then ok_docs=0; log "reset left volumes"; fi
  [[ $ok_docs == 1 ]] && record B-DOCS-EXEC-ROOT PASS "README configure, start, migrate (status, up, status), check, stop and reset blocks executed verbatim in the actual repository" || record B-DOCS-EXEC-ROOT FAIL "README blocks"
}

# B-HEALTH-ROOT: the six health routes on the actual root, through the host ports: status, exact body,
# Cache-Control no-store and a valid X-Request-Id on each.
health_root() { # health_root WEB_PORT CORE_PORT INTELLIGENCE_PORT
  local ok=1 spec port path status body h b
  for spec in "$2|/healthz|200|^\{\"status\":\"alive\"\}$" \
              "$2|/readyz|200|^\{\"status\":\"ready\",\"product_name\":\"[^\"]*\",\"checks\":\{\"object_store\":\"ok\",\"postgres\":\"ok\",\"redis\":\"ok\"\}(,\"optional\":\{\"intelligence\":\"ok\"\})?\}$" \
              "$1|/api/health|200|^\{\"status\":\"alive\"\}$" \
              "$1|/api/ready|200|^\{\"status\":\"ready\",\"checks\":\{\"core\":\"ok\"\}\}$" \
              "$3|/healthz|200|^\{\"status\":\"ok\"\}$" \
              "$3|/readyz|200|^\{\"status\":\"ready\"\}$"; do
    IFS='|' read -r port path status b <<<"$spec"
    h="$(http_get_headers 127.0.0.1 "$port" "$path")"; body="$(http_get 127.0.0.1 "$port" "$path" | cut -f2-)"
    [[ "$(head -1 <<<"$h")" == "$status" ]] || { ok=0; log "$port$path status $(head -1 <<<"$h")"; }
    grep -qE '^cache-control:.*\bno-store\b' <<<"$h" || { ok=0; log "$port$path lacks Cache-Control: no-store"; }
    grep -qE '^x-request-id:[[:space:]]*[A-Za-z0-9._-]{8,64}[[:space:]]*$' <<<"$h" || { ok=0; log "$port$path lacks a valid X-Request-Id"; }
    [[ "$body" =~ $b ]] || { ok=0; log "$port$path body: $body"; }
  done
  [[ $ok == 1 ]] && record B-HEALTH-ROOT PASS "core /healthz and /readyz, web /api/health and /api/ready, intelligence /healthz and /readyz on the actual root (host ports $2, $1, $3): 200, exact bodies, Cache-Control no-store and a valid X-Request-Id on each" || record B-HEALTH-ROOT FAIL "health routes on the root stack"
}

# --- G003 error output of the real artifacts (Plan §10; CE1-CE3, WE0-WE6, IE1-IE2) -------------------
# Inputs live in $WORK/obs-inputs (mode 700, excluded from the final scan like secrets*) and are deleted
# after use; canaries are added to the secret list. Outputs keep statuses and categories only.
obs_inputs() { mkdir -p "$WORK/obs-inputs"; chmod 700 "$WORK/obs-inputs"; printf '%s' "$WORK/obs-inputs"; }
new_canary() { local c; c="Cx$(random_hex 12)"; printf '%s\n' "$c" >>"$SECRETS_FILE"; printf '%s' "$c"; }

# core_env FILE LISTEN PRODUCT: a complete, synthetic core configuration (never the root .env).
core_env() {
  (umask 077; cat >"$1" <<ENV
PRODUCT_NAME=$3
PLATFORM_BASE_DOMAIN=probe.example.net
CUSTOM_DOMAIN_CNAME_TARGET=custom.probe.example.net
PLATFORM_AUTH_HOST=app.probe.example.net
DATABASE_NAME=probe_db
APP_ENV=test
CORE_LISTEN_ADDR=$2
DB_HOST=postgres.invalid
DB_PORT=5432
DB_USER=probe_db_runtime
DB_PASSWORD=$(random_hex 16)
REDIS_HOST=redis.invalid
REDIS_PORT=6379
REDIS_PASSWORD=$(random_hex 16)
OBJECT_STORE_ENDPOINT=http://minio.invalid:9000
OBJECT_STORE_REGION=us-east-1
OBJECT_STORE_BUCKET=probe-bucket
OBJECT_STORE_FORCE_PATH_STYLE=true
OBJECT_STORE_ACCESS_KEY=probe-access
OBJECT_STORE_SECRET_KEY=$(random_hex 16)
ENV
  )
}

# B-CORE-ERROR-OUTPUT: CE1 invalid configuration, CE2 listen failure, CE3 a valid bind then TERM.
core_error_output() {
  local ok=1 in img="${ROOT_PROJECT}-core" D="$WORK/error-core" c1 c2 c3 cid rc i
  in="$(obs_inputs)"; mkdir -p "$D"; chmod 777 "$D"
  ensure_tools || { record B-CORE-ERROR-OUTPUT FAIL "validation tools unavailable"; return; }
  c1="$(new_canary)"; c2="$(new_canary)"; c3="$(new_canary)"
  (umask 077; printf 'APP_ENV=test\nDB_PASSWORD=%s\nREDIS_PASSWORD=%s\n' "$c1" "$c1" >"$in/ce1.env")
  core_env "$in/ce2.env" "bad-$c2.invalid:8180" "Probe Product"
  core_env "$in/ce3.env" ":8180" "$c3"
  printf '%s\n%s\n%s\n' "$c1" "$c2" "$c3" >"$in/core-canaries.txt"
  rc=0; docker_run_owned core-ce --rm --label g001.run=$RUN_ID --network none --env-file "$in/ce1.env" "$img" >"$D/ce1.out" 2>&1 || rc=$?; echo "$rc" >"$D/ce1.out.rc"
  rc=0; docker_run_owned core-ce --rm --label g001.run=$RUN_ID --network none --env-file "$in/ce2.env" "$img" >"$D/ce2.out" 2>&1 || rc=$?; echo "$rc" >"$D/ce2.out.rc"
  cid="$(docker_run_owned core-ce3 -d --label g001.run=$RUN_ID --network none --env-file "$in/ce3.env" "$img")" || ok=0
  for ((i = 0; i < 60; i++)); do docker logs "$cid" 2>&1 | grep -q '"event":"core listening"' && break; sleep 1; done
  docker stop "$cid" >/dev/null 2>&1 || true
  rc="$(docker wait "$cid" 2>/dev/null || echo 255)"; docker logs "$cid" >"$D/ce3.out" 2>&1 || true; echo "$rc" >"$D/ce3.out.rc"
  docker rm -f "$cid" >/dev/null 2>&1 || true
  cp "$in/core-canaries.txt" "$D/canaries.txt"
  vtools -v "$D:/in:ro" "$VT_IMG" logcheck startup --service core \
    --case /in/ce1.out=1:core.exit:config_invalid \
    --case /in/ce2.out=1:core.exit:listen_failed --exactly-one-line /in/ce2.out \
    --case "/in/ce3.out=0:core listening:port=8180+app_env=test" --exactly-one-line /in/ce3.out \
    --canaries /in/canaries.txt >"$WORK/error-output-core.log" 2>&1 || { ok=0; tail -20 "$WORK/error-output-core.log"; }
  printf 'case\texit\n' >"$WORK/error-output-core.tsv"
  for c in ce1 ce2 ce3; do printf '%s\t%s\n' "$c" "$(cat "$D/$c.out.rc")" >>"$WORK/error-output-core.tsv"; done
  rm -rf "$D" "$in"/ce*.env "$in/core-canaries.txt"
  [[ $ok == 1 ]] && record B-CORE-ERROR-OUTPUT PASS "real core image, no network: CE1 invalid configuration exit 1 with one core.exit/config_invalid (allow-listed config_keys); CE2 listen failure exit 1 with exactly one line core.exit/listen_failed and no core listening; CE3 one core listening with exactly port=8180 and app_env=test, TERM exit 0 and no further line; no canary; every line catalogued JSON" || record B-CORE-ERROR-OUTPUT FAIL "core error output"
}

# B-WEB-ERROR-OUTPUT: a disposable copy with the two test-only probe routes, built as the production
# runtime image and run without network; httpprobe joins its namespace (WE0-WE6, WE3a, WE3b).
web_error_output() {
  local ok=1 in copy="$WORK/weprobe" img="g003-web-probe-$RUN_ID" D="$WORK/error-web" cid ccid can r i
  in="$(obs_inputs)"; mkdir -p "$D"; chmod 777 "$D"
  ensure_tools || { record B-WEB-ERROR-OUTPUT FAIL "validation tools unavailable"; return; }
  copy_repo "$copy"
  # Next.js treats folders starting with "_" as private; the URL-encoded folder name %5F%5Fprobe is served
  # at /api/__probe (G003 deviation recorded in the Result).
  for r in wrapped control; do mkdir -p "$copy/web/app/api/%5F%5Fprobe/$r"; cp "$copy/web/tests/fixtures/error-probe/$r-route.ts" "$copy/web/app/api/%5F%5Fprobe/$r/route.ts"; done
  if ledger_build "$LEDGER_SUITE" named:B-WEB-ERROR-OUTPUT -- --label g001.run=$RUN_ID -f "$copy/infra/docker/web.Dockerfile" --target runtime -t "$img" "$copy/web" >"$WORK/web-probe-build.log" 2>&1; then record_built_image "$img"
  else tail -20 "$WORK/web-probe-build.log"; rm -rf "$copy"; record B-WEB-ERROR-OUTPUT FAIL "probe image build"; return; fi
  rm -rf "$copy"
  can="$(new_canary)"; printf '%s\n' "$can" >"$D/canaries.txt"
  (umask 077; printf 'PRODUCT_NAME=Probe Product\nPLATFORM_BASE_DOMAIN=probe.example.net\nCUSTOM_DOMAIN_CNAME_TARGET=custom.probe.example.net\nAPP_ENV=test\nCORE_INTERNAL_URL=http://127.0.0.1:9\nNEXT_TELEMETRY_DISABLED=1\nPROBE_CANARY=%s\n' "$can" >"$in/we.env")
  # WE5's URL carries the canary, so it travels in a file; the directory is readable by the nonroot tools image.
  mkdir -p "$in/we5"; chmod 755 "$in/we5"
  (umask 022; printf 'http://127.0.0.1:3000/api/__probe/wrapped?mode=ok&q=%%E0%%A4%%A-%s&r=%%ZZ%s&s=%%-%s\n' "$can" "$can" "$can" >"$in/we5/we5-url.txt")
  cid="$(docker_run_owned web-probe -d --label g001.run=$RUN_ID --network none --env-file "$in/we.env" "$img")" || ok=0
  for ((i = 0; i < 60; i++)); do
    docker_run_owned web-httpprobe --rm --label g001.run=$RUN_ID --network "container:$cid" "$VT_IMG" httpprobe --url http://127.0.0.1:3000/api/health --timeout 2s >/dev/null 2>&1 && break; sleep 1
  done
  we_probe() { # we_probe NAME ARGS...: one httpprobe request in the probe's namespace; JSON to $D/NAME.json
    local name="$1"; shift
    docker_run_owned web-httpprobe --rm --label g001.run=$RUN_ID --network "container:$cid" -v "$in/we5:/in:ro" "$VT_IMG" httpprobe "$@" >"$D/$name.json" 2>"$D/$name.err" || true
  }
  we_probe we1 --url "http://127.0.0.1:3000/api/__probe/wrapped?mode=throw"
  we_probe we2 --url "http://127.0.0.1:3000/api/__probe/wrapped?mode=nonerror"
  we_probe we3a --url "http://127.0.0.1:3000/api/__probe/wrapped?mode=badname-valid"
  we_probe we3b --url "http://127.0.0.1:3000/api/__probe/wrapped?mode=badname"
  we_probe we5 --url-file /in/we5-url.txt
  we_probe we6 --url "http://127.0.0.1:3000/api/__probe/wrapped?mode=slow" --abort-after 500ms
  sleep 4
  docker logs "$cid" >"$D/web-probe.log" 2>&1 || ok=0
  docker rm -f "$cid" >/dev/null 2>&1 || true
  # WE0: every line is either the exact startup banner or a catalogued JSON event.
  vtools -v "$D:/in:ro" "$VT_IMG" logcheck validate --service web --file /in/web-probe.log >"$WORK/error-output-web.log" 2>&1 || { ok=0; log "WE0: unclassified web output"; tail -10 "$WORK/error-output-web.log"; }
  # WE1-WE3b, WE5, WE6: per-request response and exactly one record with the expected outcome/category.
  local spec name want_status want_outcome want_type st xr recs
  for spec in "we1|500|failed|Error" "we2|500|failed|NonError" "we3a|500|failed|Error" "we3b|500|failed|Error" "we5|any|completed|-" "we6|0|cancelled|-"; do
    IFS='|' read -r name want_status want_outcome want_type <<<"$spec"
    st="$(yq_stdin -p json -o tsv '.status' <"$D/$name.json" 2>/dev/null)"; xr="$(yq_stdin -p json -o tsv '.x_request_id // ""' <"$D/$name.json" 2>/dev/null)"
    [[ "$want_status" == any || "$st" == "$want_status" ]] || { ok=0; log "$name status $st, want $want_status"; }
    if [[ "$want_status" == 500 ]]; then
      [[ "$(yq_stdin -p json -o tsv '.body_sha256 // ""' <"$D/$name.json")" == e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 ]] || { ok=0; log "$name body is not empty"; }
      [[ "$(yq_stdin -p json -o tsv '.cache_control // ""' <"$D/$name.json")" == *no-store* ]] || { ok=0; log "$name lacks no-store"; }
      [[ "$xr" =~ ^[A-Za-z0-9._-]{8,64}$ ]] || { ok=0; log "$name lacks X-Request-Id"; }
    fi
    if [[ -n "$xr" ]]; then
      recs="$(vtools -v "$D:/in:ro" "$VT_IMG" logcheck records --service web --file /in/web-probe.log --request-id "$xr" 2>/dev/null || true)"
    else # an aborted request has no response header: select by route and outcome
      recs="$(grep -F '"outcome":"cancelled"' "$D/web-probe.log" || true)"
    fi
    [[ "$(grep -c . <<<"$recs")" == 1 ]] || { ok=0; log "$name: $(grep -c . <<<"$recs") records"; }
    grep -qF "\"outcome\":\"$want_outcome\"" <<<"$recs" || { ok=0; log "$name outcome"; }
    [[ "$want_type" == - ]] || grep -qF "\"error_type\":\"$want_type\"" <<<"$recs" || { ok=0; log "$name error_type"; }
  done
  # No canary anywhere in the retained probe output (the WE4 control below is separate and deleted).
  vtools -v "$D:/in:ro" "$VT_IMG" logcheck scan --canaries /in/canaries.txt /in/web-probe.log /in/we1.json /in/we2.json /in/we3a.json /in/we3b.json /in/we5.json /in/we6.json >>"$WORK/error-output-web.log" 2>&1 || { ok=0; log "a canary reached the web output"; }
  # WE4 positive control: the unwrapped route's error reaches Next's real output with the canary.
  ccid="$(docker_run_owned web-probe-control -d --label g001.run=$RUN_ID --network none --env-file "$in/we.env" "$img")" || ok=0
  for ((i = 0; i < 60; i++)); do
    docker_run_owned web-httpprobe --rm --label g001.run=$RUN_ID --network "container:$ccid" "$VT_IMG" httpprobe --url http://127.0.0.1:3000/api/health --timeout 2s >/dev/null 2>&1 && break; sleep 1
  done
  docker_run_owned web-httpprobe --rm --label g001.run=$RUN_ID --network "container:$ccid" "$VT_IMG" httpprobe --url http://127.0.0.1:3000/api/__probe/control >"$D/we4.json" 2>/dev/null || true
  # G009 correction pass 2: read the whole log into a private file before searching it. A `docker logs | grep -q`
  # pipeline under pipefail fails when grep exits on the match while docker logs still writes (SIGPIPE, exit 141),
  # which reported a present canary as vacuous. A failed log read fails closed; the read is retried for a bounded 10 s
  # until the canary appears, and an absent canary still fails. The canary-bearing copy is deleted at once.
  local we4="vacuous"
  for ((i = 0; i < 20; i++)); do
    (umask 077; docker logs "$ccid" >"$D/we4-control.log" 2>&1) || { we4="log-read-failed"; break; }
    if grep -qF "$can" "$D/we4-control.log"; then we4="canary-present"; break; fi
    sleep 0.5
  done
  rm -f "$D/we4-control.log"
  if [[ $we4 == canary-present ]]; then printf 'we4\tcanary-present\n' >"$D/we4.status"; else ok=0; log "WE4 control: the canary did not reach Next's output ($we4)"; printf 'we4\t%s\n' "$we4" >"$D/we4.status"; fi
  docker rm -f "$ccid" >/dev/null 2>&1 || true
  [[ "$(yq_stdin -p json -o tsv '.status' <"$D/we4.json" 2>/dev/null)" == 500 ]] || { ok=0; log "WE4 control status"; }
  { printf 'case\tresult\n'; for r in we1 we2 we3a we3b we5 we6; do printf '%s\t%s\n' "$r" "$(yq_stdin -p json -o tsv '.status' <"$D/$r.json" 2>/dev/null || echo error)"; done; cat "$D/we4.status"; } >"$WORK/error-output-web.tsv"
  rm -rf "$D" "$in/we.env" "$in/we5"
  [[ $ok == 1 ]] && record B-WEB-ERROR-OUTPUT PASS "production web build with test-only probe routes, no network: WE0 only the startup banner and catalogued JSON; WE1 thrown error -> 500, empty body, no-store, X-Request-Id, one failed/Error record; WE2 NonError; WE3a regex-valid canary name -> Error; WE3b -> Error; WE5 malformed query -> no extra line; WE6 client abort -> one cancelled record; canary absent everywhere except the WE4 control, whose unwrapped error reaches Next's output (deleted after the assertion)" || record B-WEB-ERROR-OUTPUT FAIL "web error output"
}

# B-INTEL-ERROR-OUTPUT: IE1 startup failure exits 3 with one bounded record; IE2 a valid start and stop
# with every line a catalogued JSON event.
intel_error_output() {
  local ok=1 in img="${ROOT_PROJECT}-intelligence" D="$WORK/error-intel" can cid rc i
  in="$(obs_inputs)"; mkdir -p "$D"; chmod 777 "$D"
  ensure_tools || { record B-INTEL-ERROR-OUTPUT FAIL "validation tools unavailable"; return; }
  can="$(random_hex 8 | tr 'a-f' 'A-F')"; printf 'CANARY%s\n' "$can" >>"$SECRETS_FILE"; printf 'CANARY%s\n' "$can" >"$D/canaries.txt"
  (umask 077; printf 'INTELLIGENCE_SERVICE_TOKEN=%s\nAI_MODEL_CANARY%s=openai:x\n' "$(random_hex 16)" "$can" >"$in/ie1.env")
  (umask 077; printf 'APP_ENV=test\nINTELLIGENCE_SERVICE_TOKEN=%s\n' "$(random_hex 16)" >"$in/ie2.env")
  rc=0; docker_run_owned intel-ie --rm --label g001.run=$RUN_ID --network none --env-file "$in/ie1.env" "$img" >"$D/ie1.out" 2>&1 || rc=$?; echo "$rc" >"$D/ie1.out.rc"
  cid="$(docker_run_owned intelligence-ie2 -d --label g001.run=$RUN_ID --network none --env-file "$in/ie2.env" "$img")" || ok=0
  for ((i = 0; i < 60; i++)); do docker logs "$cid" 2>&1 | grep -q '"event":"uvicorn.startup_complete"' && break; sleep 1; done
  docker stop "$cid" >/dev/null 2>&1 || true
  rc="$(docker wait "$cid" 2>/dev/null || echo 255)"; docker logs "$cid" >"$D/ie2.out" 2>&1 || true; echo "$rc" >"$D/ie2.out.rc"
  docker rm -f "$cid" >/dev/null 2>&1 || true
  vtools -v "$D:/in:ro" "$VT_IMG" logcheck startup --service intelligence \
    --case /in/ie1.out=3:intelligence.startup.failed:ConfigError \
    --case /in/ie2.out=0:uvicorn.startup_complete \
    --canaries /in/canaries.txt >"$WORK/error-output-intelligence.log" 2>&1 || { ok=0; tail -20 "$WORK/error-output-intelligence.log"; }
  grep -q 'Traceback' "$D/ie1.out" && { ok=0; log "IE1 printed a traceback"; }
  printf 'case\texit\nie1\t%s\nie2\t%s\n' "$(cat "$D/ie1.out.rc")" "$(cat "$D/ie2.out.rc")" >"$WORK/error-output-intelligence.tsv"
  rm -rf "$D" "$in"/ie*.env
  [[ $ok == 1 ]] && record B-INTEL-ERROR-OUTPUT PASS "real intelligence image, no network: IE1 invalid configuration (a canary in a variable name) exits 3 with exactly one intelligence.startup.failed/ConfigError, no traceback, no canary; IE2 start, uvicorn.startup_complete, TERM exit 0 (uvicorn is PID 1, so its re-raised TERM is ignored): every line a catalogued JSON event, uvicorn templates mapped, no text line" || record B-INTEL-ERROR-OUTPUT FAIL "intelligence error output"
}

web_http() { # web_http WEB_PORT
  local ok=1
  no_store 127.0.0.1 "$1" /api/health 200 || { ok=0; log "web /api/health: $(http_get_headers 127.0.0.1 "$1" /api/health | grep -E '^[0-9]|^cache-control' | tr '\n' ' ')"; }
  no_store 127.0.0.1 "$1" /api/ready 200 || { ok=0; log "web /api/ready: $(http_get_headers 127.0.0.1 "$1" /api/ready | grep -E '^[0-9]|^cache-control' | tr '\n' ' ')"; }
  [[ $ok == 1 ]] && record B-WEB-HTTP PASS "host probes: web /api/health 200 and /api/ready 200, both Cache-Control no-store" || record B-WEB-HTTP FAIL "web health headers"
}

# B-BROWSER-ROOT: README platform:browser executed once with an execution-owned container name instead of --rm,
# internal report paths and PLAYWRIGHT_WORKERS. Copy reports out before removing the container (the existing feature
# stack pattern): docker cp assigns host ownership without changing pwuser or adding a report bind mount.
browser_root() {
  local ok=1 rep="$WORK/browser-root" ver p line runc
  runc="g001-browser-$RUN_ID-$(random_hex 4)"
  mkdir -p "$rep"
  line="$(docs_block "$REPO/README.md" platform:browser)"
  if [[ "$(grep -c . <<<"$line")" != 1 || "$line" != *" run --rm "* ]]; then
    ok=0; log "README platform:browser must be one docker compose run --rm line"
  else
    line="${line/ run --rm / run --name $runc --label g001.run=$RUN_ID -e PLAYWRIGHT_JSON_OUTPUT_FILE=/tmp/report/results.json -e PLAYWRIGHT_OUTPUT_DIR=/tmp/report/out -e PLAYWRIGHT_WORKERS=$PW_WORKERS }"
    printf '%s\n' "$line" >"$WORK/root-browser.log.cmds"
    ledger_watch_begin "$REPO" "$ROOT_PROJECT" browser-root
    (cd "$REPO" && bash -euo pipefail -c "$line") >"$WORK/root-browser.log" 2>&1 || { ok=0; tail -30 "$WORK/root-browser.log"; log "README platform:browser (with the report injection) failed"; }
    ledger_watch_end "$REPO" "$ROOT_PROJECT" browser-root "$WORK/root-browser.log" "$line"
    capture_after "$REPO" run --build
    browser_collect_report "$runc" "$rep" "$WORK/root-browser.log" || ok=0
  fi
  browser_matrix "$rep/results.json" "$WORK/browser-root-tests.tsv" "${BROWSER_MATRIX_FULL[@]}" || ok=0
  init_hash_documented "$rep/results.json" "$REPO/docs/foundation/WEB-SHELL.md" || ok=0
  # S-BOUNDARY saves the delivered bodies as plain files, so the final secret scan reads them.
  for p in shell-light shell-dark; do
    find "$rep/out" -path "*-$p/responses.txt" -size +0 | grep -q . || { ok=0; log "no saved response bodies for $p"; }
  done
  ver="$(chromium_version "$REPO")"
  [[ -n "$ver" ]] || { ok=0; log "Chromium version not determined"; }
  [[ $ok == 1 ]] && record B-BROWSER-ROOT PASS "README platform:browser executed once (exit 0) with the reviewed report collection (named container, internal JSON/output paths copied to host ownership, $PW_WORKERS Playwright workers): $(grep -c . "$WORK/browser-root-tests.tsv") tests, every required scenario ID passes in each required project (shell-light/dark, tokens-light/dark, appearance, dev-hydration), S-ERRORS on every test, 0 skipped/flaky/unexpected; served appearance-init hash equals WEB-SHELL.md; saved response bodies for shell-light/dark; Chromium $ver" || record B-BROWSER-ROOT FAIL "browser suite on the root stack"
}

root_defaults() {
  local ok=1 res
  [[ "$(env_get "$REPO" BIND_ADDRESS)" == 127.0.0.1 ]] || { ok=0; log "BIND_ADDRESS is not loopback"; }
  res="$(http_get 127.0.0.1 "$(env_get "$REPO" CORE_PORT)" /readyz)"
  [[ "$(cut -f1 <<<"$res")" == 200 ]] || { ok=0; log "core /readyz: $res"; }
  grep -q '"status":"ready"' <<<"$res" || grep -q '"checks":{' <<<"$res" || { ok=0; log "readiness body: $res"; }
  if grep -qE '"(postgres|redis|object_store)":"[^o]' <<<"$res"; then ok=0; log "a dependency is not ok: $res"; fi
  grep -qF "\"product_name\":\"$(env_get "$REPO" PRODUCT_NAME)\"" <<<"$res" || { ok=0; log "core does not report the configured product name"; }
  [[ $ok == 1 ]] && record B-DEFAULTS PASS "actual repository started with the assigned defaults (ports ${DEFAULT_PORTS[*]} on 127.0.0.1); core ready with every dependency ok and the configured product name" || record B-DEFAULTS FAIL "defaults"
}

# probe_ports LABEL DEST WEB CORE INTEL|- PG REDIS MINIO_API MINIO_CONSOLE
probe_ports() {
  local label="$1" dest="$2" web="$3" core="$4" intel="$5" pg="$6" redis="$7" mapi="$8" mcon="$9"
  local ok=1 svc port pub p
  local pairs=("web:3000:$web" "core:8180:$core" "postgres:5432:$pg" "redis:6379:$redis" "minio:9000:$mapi" "minio:9001:$mcon")
  [[ "$intel" != "-" ]] && pairs+=("intelligence:8000:$intel")
  for p in "${pairs[@]}"; do
    IFS=: read -r svc port pub <<<"$p"
    [[ "$(compose "$dest" port "$svc" "$port" 2>/dev/null)" == "127.0.0.1:$pub" ]] || { ok=0; log "$svc:$port published as '$(compose "$dest" port "$svc" "$port" 2>/dev/null)'"; }
  done
  if [[ "$PROBE_MODE" == none ]]; then record "B-PORTS-$label" UNMET "no host probe available"; return; fi
  [[ "$(http_get 127.0.0.1 "$web" /api/health | cut -f1)" == 200 ]] || { ok=0; log "web /api/health"; }
  [[ "$(http_get 127.0.0.1 "$core" /healthz | cut -f1)" == 200 ]] || { ok=0; log "core /healthz"; }
  [[ "$intel" == "-" || "$(http_get 127.0.0.1 "$intel" /healthz | cut -f1)" == 200 ]] || { ok=0; log "intelligence /healthz"; }
  [[ "$(http_get 127.0.0.1 "$mapi" /minio/health/live | cut -f1)" == 200 ]] || { ok=0; log "minio api"; }
  [[ "$(http_get 127.0.0.1 "$mcon" / | cut -f1)" =~ ^[1-5][0-9][0-9]$ ]] || { ok=0; log "minio console"; }
  [[ "$(tcp_redis_ping 127.0.0.1 "$redis")" == "-NOAUTH" ]] || { ok=0; log "redis did not answer -NOAUTH"; }
  [[ "$(tcp_pg_sslrequest 127.0.0.1 "$pg")" =~ ^[NS]$ ]] || { ok=0; log "postgres SSLRequest"; }
  local ip; ip="$(host_nonloopback_ip)"
  if [[ -n "$ip" ]]; then
    for port in "$web" "$core" "$pg" "$redis" "$mapi" "$mcon" "$intel"; do
      [[ "$port" == "-" ]] && continue
      tcp_open "$ip" "$port" && { ok=0; log "port $port reachable on non-loopback $ip"; }
    done
  else ok=0; log "no non-loopback address to prove loopback-only binding"; fi
  local core_cid; core_cid="$(cid_of "$dest" core)"
  for p in postgres:5432 redis:6379 minio:9000; do ns_tcp "$core_cid" "${p%%:*}" "${p##*:}" || { ok=0; log "core cannot reach $p internally"; }; done
  [[ $ok == 1 ]] && record "B-PORTS-$label" PASS "published on 127.0.0.1 at $web/$core/$intel/$pg/$redis/$mapi/$mcon; host probes ($PROBE_MODE): HTTP health, Redis -NOAUTH, PostgreSQL SSLRequest; refused on non-loopback $ip; internal DNS on 5432/6379/9000 from core" || record "B-PORTS-$label" FAIL "port publication or reachability"
}

datastore_port() { case "$1" in postgres) echo 5432 ;; redis) echo 6379 ;; minio) echo 9000 ;; *) return 1 ;; esac; }

# forbidden_datastores_reachable PROBE_CID DEST [SVC...]: probes every datastore
# independently, by name and by each container IP, on its own port. Prints one
# line per reachable endpoint; returns 0 if ANY is reachable.
forbidden_datastores_reachable() {
  local probe="$1" dest="$2"; shift 2
  local services=("$@") svc port ip found=1
  [[ ${#services[@]} -gt 0 ]] || services=(postgres redis minio)
  for svc in "${services[@]}"; do
    port="$(datastore_port "$svc")"
    if ns_tcp "$probe" "$svc" "$port"; then printf '%s by name %s:%s\n' "$svc" "$svc" "$port"; found=0; fi
    for ip in $(container_ips "$(cid_of "$dest" "$svc")"); do
      if ns_tcp "$probe" "$ip" "$port"; then printf '%s by IP %s:%s\n' "$svc" "$ip" "$port"; found=0; fi
    done
  done
  return "$found"
}

boundaries() { # boundaries LABEL DEST INTELLIGENCE(yes|no)
  local label="$1" dest="$2" intel="$3" ok=1 web core ic ip svc hits mig
  web="$(cid_of "$dest" web)"; core="$(cid_of "$dest" core)"
  # Negative probes from a missing container would pass vacuously.
  [[ -n "$web" && -n "$core" ]] || { ok=0; log "web or core container not found"; }
  for svc in postgres redis minio; do
    [[ -n "$(container_ips "$(cid_of "$dest" "$svc")" 2>/dev/null | tr -d ' ')" ]] || { ok=0; log "no container IP for $svc"; }
  done
  if hits="$(forbidden_datastores_reachable "$web" "$dest")"; then ok=0; log "web reaches datastores: ${hits//$'\n'/; }"; fi
  for svc in postgres redis minio; do ns_resolves "$web" "$svc" && { ok=0; log "web resolves $svc"; }; done
  if [[ "$intel" == yes ]]; then
    ic="$(cid_of "$dest" intelligence)"
    [[ -n "$ic" ]] || { ok=0; log "intelligence container not found"; }
    ns_tcp "$web" intelligence 8000 && { ok=0; log "web reaches intelligence"; }
    for ip in $(container_ips "$ic"); do ns_tcp "$web" "$ip" 8000 && { ok=0; log "web reaches intelligence by IP"; }; done
    if hits="$(forbidden_datastores_reachable "$ic" "$dest")"; then ok=0; log "intelligence reaches datastores: ${hits//$'\n'/; }"; fi
    docker inspect -f '{{range .Config.Env}}{{println .}}{{end}}' "$ic" | cut -d= -f1 | grep -qE '^(DB_|REDIS_|OBJECT_STORE_|MINIO_|POSTGRES_)' && { ok=0; log "intelligence env holds data-store settings"; }
    ns_tcp "$core" intelligence 8000 || { ok=0; log "core cannot reach intelligence"; }
    execute_probes "$dest" "$core" || ok=0
  fi
  for svc in postgres redis minio; do
    forbidden_datastores_reachable "$core" "$dest" "$svc" >/dev/null || { ok=0; log "control: core cannot reach $svc (probe ineffective)"; }
  done
  # Runtime containers never hold the migrator credential.
  mig="$(env_get "$dest" DB_MIGRATOR_PASSWORD)"
  for svc in core web intelligence; do
    [[ "$svc" == intelligence && "$intel" != yes ]] && continue
    if docker inspect -f '{{range .Config.Env}}{{println .}}{{end}}' "$(cid_of "$dest" "$svc")" | grep -qF -e "DB_MIGRATOR_PASSWORD" -e "$mig"; then ok=0; log "$svc holds the migrator credential"; fi
  done
  if [[ $ok == 1 ]]; then
    record "B-BOUNDARIES-$label" PASS "web isolated from every datastore$([[ $intel == yes ]] && echo ' and intelligence; intelligence isolated from every datastore and holds no data settings') (each by name and every IP on its own port); core reaches each datastore (positive control)$([[ $intel == yes ]] && echo ' and intelligence; execute endpoint 401/422/typed stream without a gateway key'); no runtime container holds the migrator credential"
  else record "B-BOUNDARIES-$label" FAIL "boundary violation"; fi
}

# Negative controls: the boundary check must flag any single reachable datastore.
boundary_controls() { # boundary_controls DEST
  local dest="$1" ok=1 net probe svc hits isolated
  net="$(docker inspect -f '{{range $k, $v := .NetworkSettings.Networks}}{{$k}} {{end}}' "$(cid_of "$dest" postgres)" | tr ' ' '\n' | grep -E '_data$' | head -1)"
  probe="$(docker_run_owned boundary-probe -d --rm --label g001.run=$RUN_ID --network "$net" --name "g001-probe-$RUN_ID" "$BUSYBOX_IMAGE" sleep 300)"
  for svc in postgres redis minio; do
    if hits="$(forbidden_datastores_reachable "$probe" "$dest" "$svc")"; then log "control: $svc reachable from the data network is flagged (${hits//$'\n'/; })"
    else ok=0; log "control FAILED: $svc reachable but not flagged"; fi
  done
  docker rm -f "$probe" >/dev/null 2>&1 || true
  isolated="$(cid_of "$dest" web)"
  for svc in postgres redis minio; do
    if hits="$(
      STUB_SVC_IPS=" $(container_ips "$(cid_of "$dest" "$svc")")" STUB_PORT="$(datastore_port "$svc")"
      ns_tcp() { [[ "$3" == "$STUB_PORT" && "$STUB_SVC_IPS " == *" $2 "* ]]; }
      forbidden_datastores_reachable "$isolated" "$dest"
    )"; then
      grep -q "^$svc by IP" <<<"$hits" || { ok=0; log "control FAILED: stubbed $svc-only reachability attributed wrongly: $hits"; }
    else ok=0; log "control FAILED: stubbed $svc-only reachability not flagged"; fi
  done
  [[ $ok == 1 ]] && record B-BOUNDARY-CONTROLS PASS "negative controls: postgres-only, redis-only and minio-only reachability each fail the boundary check (real data-network probe and stubbed single-endpoint probe)" || record B-BOUNDARY-CONTROLS FAIL "a negative control was not detected"
}

execute_probes() { # From core's network namespace, using busybox wget.
  local dest="$1" core="$2" out body url="http://intelligence:8000/internal/v1/execute"
  out="$(docker_run_owned wget-probe --rm --label g001.run=$RUN_ID --network "container:$core" "$BUSYBOX_IMAGE" wget -S -q -O- --header 'Content-Type: application/json' --post-data '{}' "$url" 2>&1 || true)"
  grep -q ' 401 ' <<<"$out" || { log "execute without token: $(head -1 <<<"$out")"; return 1; }
  # The token reaches the probe container through a file, never the command line.
  local tok="$WORK/exec-token"; (umask 077; printf 'T=%s\n' "$(env_get "$dest" INTELLIGENCE_SERVICE_TOKEN)" >"$tok")
  # The request is built by the shell's printf builtin and sent with nc, so the
  # token never appears in any process's arguments. stdin stays open briefly:
  # nc half-closes on stdin EOF, which uvicorn treats as a lost client.
  local req='{ printf "POST /internal/v1/execute HTTP/1.0\r\nHost: intelligence\r\nAuthorization: Bearer %s\r\nContent-Type: application/json\r\nContent-Length: %d\r\n\r\n%s" "$T" "${#B}" "$B"; sleep 10; } | nc -w 15 intelligence 8000'
  out="$(docker_run_owned wget-probe --rm --label g001.run=$RUN_ID --network "container:$core" --env-file "$tok" -e B='{"bad":true}' "$BUSYBOX_IMAGE" sh -c "$req" 2>&1 || true)"
  grep -q '^HTTP/1\.[01] 422 ' <<<"$out" || { log "execute with bad envelope: $(head -1 <<<"$out")"; rm -f "$tok"; return 1; }
  body='{"request_id":"req-probe-0001","trace_id":"4bf92f3577b34da6a3ce929d0e0e4736","capability":"chat","authorization_ref":"authz-probe","organization_id":"a0000000-0000-4000-8000-000000000001","user_id":"c0000000-0000-4000-8000-000000000001","allowed_tools":[],"deadline":"2999-01-01T00:00:00Z","context_refs":[],"input":{"messages":[{"role":"user","content":"probe"}]}}'
  out="$(docker_run_owned wget-probe --rm --label g001.run=$RUN_ID --network "container:$core" --env-file "$tok" -e B="$body" "$BUSYBOX_IMAGE" sh -c "$req" 2>&1 || true)"
  grep -q '^HTTP/1\.[01] 200 ' <<<"$out" || { log "execute with valid envelope: $(head -1 <<<"$out")"; rm -f "$tok"; return 1; }
  rm -f "$tok"
  grep -q '"type": *"error"' <<<"$out" && grep -q 'capability_not_configured' <<<"$out" && grep -q '"type": *"done"' <<<"$out" || { log "execute with valid envelope returned: $out"; return 1; }
}

# B-MCP-HTTP: MCP is unconfigured in the default stack (no MCP_* variable; G014), so no MCP route exists there:
# JSON-RPC initialize and SSE-style GETs, including a per-connection path and its protected-resource metadata, answer
# 404 on core (inside its own network namespace and through the host port) and web. The configured behaviour (tenant
# hosts only) is B-MCP-E2E.
MCP_HTTP_CONNECTION=/mcp/c/7f000000-0000-4000-8000-000000000001
mcp_http() { # mcp_http DEST CORE_PORT WEB_PORT
  local dest="$1" cport="$2" wport="$3" ok=1 core path out res
  core="$(cid_of "$dest" core)"
  local init='{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"probe","version":"0"}}}'
  for path in /mcp /mcp/ /sse /messages /message /rpc "$MCP_HTTP_CONNECTION" "/.well-known/oauth-protected-resource$MCP_HTTP_CONNECTION" /internal/v1/mcp/connections; do
    out="$(docker_run_owned wget-probe --rm --label g001.run=$RUN_ID --network "container:$core" -e B="$init" "$BUSYBOX_IMAGE" sh -c 'wget -S -q -O- --header "Content-Type: application/json" --header "Accept: application/json, text/event-stream" --post-data "$B" http://127.0.0.1:8180'"$path"' 2>&1' || true)"
    grep -q ' 404 ' <<<"$out" || { ok=0; log "core POST $path: $(head -2 <<<"$out" | tr '\n' ' ')"; }
    grep -qiE '"jsonrpc"|protocolVersion|serverInfo|event-stream' <<<"$out" && { ok=0; log "core POST $path answered like MCP"; }
    res="$(http_get 127.0.0.1 "$cport" "$path")"
    [[ "$(cut -f1 <<<"$res")" == 404 ]] || { ok=0; log "core GET $path via host: $res"; }
  done
  for path in /mcp /sse; do
    [[ "$(http_get 127.0.0.1 "$wport" "$path" | cut -f1)" == 404 ]] || { ok=0; log "web GET $path is not 404"; }
  done
  [[ $ok == 1 ]] && record B-MCP-HTTP PASS "MCP unconfigured in the default stack: anonymous JSON-RPC initialize (POST) and GET on /mcp, /mcp/, /sse, /messages, /message, /rpc, a per-connection /mcp/c/{uuid} path, its protected-resource metadata and /internal/v1/mcp/connections return 404 from core (own namespace and host port) with no MCP payload; web /mcp and /sse 404 (configured tenant-host behaviour: B-MCP-E2E)" || record B-MCP-HTTP FAIL "an MCP-like endpoint answered on the unconfigured default stack"
}

# G005 B-TENANCY-HTTP (Plan r4 §5.2 H-1/H-2): the tenancy model has no HTTP, BFF or provisioning surface. Candidate
# tenancy routes return 404 with no model payload from core (own namespace and host port) and from web.
TENANCY_PROBE_CORE=(/organizations /groups /memberships /role-bindings /tenancy /internal/v1/organizations
  /internal/v1/memberships /internal/v1/role-bindings /internal/v1/tenancy
  /business-profiles /group-business-profiles /business-profile-access
  /internal/v1/group-business-profiles /internal/v1/business-profile-access
  /internal/v1/policy /internal/v1/context /internal/v1/access /policy)
TENANCY_PROBE_WEB=(/organizations /memberships /api/organizations /api/groups /api/memberships /api/role-bindings /api/tenancy
  /onboarding /api/business-profiles /api/group-business-profiles
  /policy /api/policy /api/context)
tenancy_http() { # tenancy_http DEST CORE_PORT WEB_PORT
  local dest="$1" cport="$2" wport="$3" ok=1 core web path out res n=0 model='"(organization_id|group_id|subject_user_id|scope_type|slug|role|memberships?)"'
  core="$(cid_of "$dest" core)"; web="$(cid_of "$dest" web)"
  for path in "${TENANCY_PROBE_CORE[@]}"; do
    out="$(docker_run_owned wget-probe --rm --label g001.run=$RUN_ID --network "container:$core" "$BUSYBOX_IMAGE" sh -c 'wget -S -q -O- --header "Content-Type: application/json" --post-data "{}" http://127.0.0.1:8180'"$path"' 2>&1' || true)"
    grep -q ' 404 ' <<<"$out" || { ok=0; log "core POST $path: $(head -2 <<<"$out" | tr '\n' ' ')"; }
    grep -qiE "$model" <<<"$out" && { ok=0; log "core POST $path answered with a model payload"; }
    res="$(http_get 127.0.0.1 "$cport" "$path")"
    [[ "$(cut -f1 <<<"$res")" == 404 ]] || { ok=0; log "core GET $path via host: $(cut -f1 <<<"$res")"; }
    grep -qiE "$model" <<<"$res" && { ok=0; log "core GET $path answered with a model payload"; }
    n=$((n + 2))
  done
  # G018 (Plan r5 §5.3, D-3): the two served core paths are no longer absent; without a session and origin they answer
  # their closed JSON refusal (GET: ambiguous_parameter; POST {}: origin_rejected on /groups, method_not_allowed on
  # /business-profiles), never 404 and never a model payload. The 401 with a resolvable origin is pinned by
  # TestOnboardingHTTP.
  for path in /internal/v1/groups /internal/v1/business-profiles; do
    res="$(http_get 127.0.0.1 "$cport" "$path")"
    [[ "$(cut -f1 <<<"$res")" == 400 ]] && grep -q '"error":"ambiguous_parameter"' <<<"$res" || { ok=0; log "core GET $path: $(cut -f1 <<<"$res")"; }
    out="$(docker_run_owned wget-probe --rm --label g001.run=$RUN_ID --network "container:$core" "$BUSYBOX_IMAGE" sh -c 'wget -S -q -O- --header "Content-Type: application/json" --post-data "{}" http://127.0.0.1:8180'"$path"' 2>&1' || true)"
    case "$path" in
      /internal/v1/groups) grep -q ' 403 ' <<<"$out" || { ok=0; log "core POST $path: $(head -2 <<<"$out" | tr '\n' ' ')"; } ;;
      *) grep -q ' 405 ' <<<"$out" || { ok=0; log "core POST $path: $(head -2 <<<"$out" | tr '\n' ' ')"; } ;;
    esac
    grep -qiE "$model" <<<"$out$res" && { ok=0; log "core $path answered with a model payload"; }
    n=$((n + 2))
  done
  # G019 (Plan r1 §5.3 B-TENANCY-HTTP positive pins): the four served membership roots answer their closed JSON refusal
  # without a session and origin (GET: ambiguous_parameter, or method_not_allowed on the POST-only transfer; POST {}:
  # origin_rejected on the two mutation roots, method_not_allowed on the GET-only ones), never 404 and never a model
  # payload. The 401 with a resolvable origin, every unlisted method (405) and sub-paths (404) are pinned by
  # TestMembershipHTTP. G060 (Plan r0 §4.2): the Group-authority path (GET without an origin: ambiguous_parameter; POST {}:
  # origin_rejected) joins them.
  local probe getc postc
  for probe in "/internal/v1/members 400 405" "/internal/v1/invitations 400 403" "/internal/v1/ownership/transfer 405 403" \
    "/internal/v1/account/retirement-eligibility 400 405" "/internal/v1/organization/group-authority 400 403"; do
    read -r path getc postc <<<"$probe"
    res="$(http_get 127.0.0.1 "$cport" "$path")"
    [[ "$(cut -f1 <<<"$res")" == "$getc" ]] && grep -qE '"error":"(ambiguous_parameter|method_not_allowed)"' <<<"$res" || { ok=0; log "core GET $path: $(cut -f1 <<<"$res")"; }
    out="$(docker_run_owned wget-probe --rm --label g001.run=$RUN_ID --network "container:$core" "$BUSYBOX_IMAGE" sh -c 'wget -S -q -O- --header "Content-Type: application/json" --post-data "{}" http://127.0.0.1:8180'"$path"' 2>&1' || true)"
    grep -q " $postc " <<<"$out" || { ok=0; log "core POST $path: $(head -2 <<<"$out" | tr '\n' ' ')"; }
    grep -qiE "$model" <<<"$out$res" && { ok=0; log "core $path answered with a model payload"; }
    n=$((n + 2))
  done
  # G058 (Plan r3 §5.2; R13 closed route contract): the enrollment root and any unserved sub-path answer closed JSON 404
  # not_found on every method; a defined path on the wrong method answers 405 method_not_allowed; the three served GET
  # routes reached without an origin answer 400 ambiguous_parameter from their own handler, and the served mutations
  # (manager /policy, applicant /request) answer POST {} 403 origin_rejected: the body decodes, its empty origin is
  # refused before any credential is read (the handler order decode -> origin -> credential of enrollment/api; never
  # 404, never a model payload).
  # Credential, origin and CSRF refusals per route are pinned by TestEnrollmentHTTP.
  for probe in "/internal/v1/enrollment 404 404 not_found" "/internal/v1/enrollment/unknown 404 404 not_found" \
    "/internal/v1/enrollment/public 400 405 ambiguous_parameter" "/internal/v1/enrollment/status 400 405 ambiguous_parameter" \
    "/internal/v1/enrollment/policy 400 403 ambiguous_parameter" "/internal/v1/enrollment/request 405 403 method_not_allowed"; do
    read -r path getc postc code <<<"$probe"
    res="$(http_get 127.0.0.1 "$cport" "$path")"
    [[ "$(cut -f1 <<<"$res")" == "$getc" ]] && grep -q "\"error\":\"$code\"" <<<"$res" || { ok=0; log "core GET $path: $(cut -f1 <<<"$res")"; }
    out="$(docker_run_owned wget-probe --rm --label g001.run=$RUN_ID --network "container:$core" "$BUSYBOX_IMAGE" sh -c 'wget -S -q -O- --header "Content-Type: application/json" --post-data "{}" http://127.0.0.1:8180'"$path"' 2>&1' || true)"
    grep -q " $postc " <<<"$out" || { ok=0; log "core POST $path: $(head -2 <<<"$out" | tr '\n' ' ')"; }
    grep -qiE "$model" <<<"$out$res" && { ok=0; log "core $path answered with a model payload"; }
    n=$((n + 2))
  done
  # G020 (Plan r2 §4.2; the labels/api closed route contract): the served label roots answer their closed JSON refusal
  # without a session and origin, never a model payload: /labels GET 400 ambiguous_parameter (no origin) and POST {} 403
  # origin_rejected; a Group's /labels GET 400 ambiguous_parameter and POST 405 (GET-only); a definition's /delete GET 405
  # method_not_allowed and POST {} 403; an unserved path below /internal/v1/labels/ 404 not_found on both methods.
  # Credential, origin, CSRF and body refusals per route are pinned by TestLabelsHTTP and TestLabelsRefusalMapping.
  for probe in "/internal/v1/labels 400 403 ambiguous_parameter" "/internal/v1/labels/unknown 404 404 not_found" \
    "/internal/v1/groups/00000000-0000-4000-8000-000000000000/labels 400 405 ambiguous_parameter" \
    "/internal/v1/labels/00000000-0000-4000-8000-000000000000/delete 405 403 method_not_allowed"; do
    read -r path getc postc code <<<"$probe"
    res="$(http_get 127.0.0.1 "$cport" "$path")"
    [[ "$(cut -f1 <<<"$res")" == "$getc" ]] && grep -q "\"error\":\"$code\"" <<<"$res" || { ok=0; log "core GET $path: $(cut -f1 <<<"$res")"; }
    out="$(docker_run_owned wget-probe --rm --label g001.run=$RUN_ID --network "container:$core" "$BUSYBOX_IMAGE" sh -c 'wget -S -q -O- --header "Content-Type: application/json" --post-data "{}" http://127.0.0.1:8180'"$path"' 2>&1' || true)"
    grep -q " $postc " <<<"$out" || { ok=0; log "core POST $path: $(head -2 <<<"$out" | tr '\n' ' ')"; }
    grep -qiE "$model" <<<"$out$res" && { ok=0; log "core $path answered with a model payload"; }
    n=$((n + 2))
  done
  for path in "${TENANCY_PROBE_WEB[@]}"; do
    out="$(docker_run_owned wget-probe --rm --label g001.run=$RUN_ID --network "container:$web" "$BUSYBOX_IMAGE" sh -c 'wget -S -q -O- --header "Content-Type: application/json" --post-data "{}" http://127.0.0.1:3000'"$path"' 2>&1' || true)"
    grep -q ' 404 ' <<<"$out" || { ok=0; log "web POST $path: $(head -2 <<<"$out" | tr '\n' ' ')"; }
    res="$(http_get 127.0.0.1 "$wport" "$path")"
    [[ "$(cut -f1 <<<"$res")" == 404 ]] || { ok=0; log "web GET $path via host: $(cut -f1 <<<"$res")"; }
    n=$((n + 2))
  done
  [[ $ok == 1 ]] && record B-TENANCY-HTTP PASS "$n probes: GET and POST {} on ${#TENANCY_PROBE_CORE[@]} candidate core tenancy routes (own namespace and host port; no model payload) and ${#TENANCY_PROBE_WEB[@]} web routes (own namespace and host port) return 404; the two served G018 core paths (/internal/v1/groups, /internal/v1/business-profiles) and the four served G019 roots (/internal/v1/members, /internal/v1/invitations, /internal/v1/ownership/transfer, /internal/v1/account/retirement-eligibility) and the G060 /internal/v1/organization/group-authority answer their closed refusals without a session (no model payload); the G058 enrollment root and an unserved sub-path answer closed JSON 404 not_found on GET and POST, /public and /status 400 ambiguous_parameter without an origin and 405 on POST, /policy 400 ambiguous_parameter and POST {} 403, /request GET 405 and POST {} 403; the G020 /internal/v1/labels 400 ambiguous_parameter without an origin and POST {} 403, a Group's /labels 400 and POST 405, a definition's /delete GET 405 and POST {} 403, an unserved label path 404 not_found on GET and POST" || record B-TENANCY-HTTP FAIL "a tenancy route answered"
}

# An environment file in a service build context never reaches the image.
build_context() {
  local copy="$WORK/ctx" marker ok=1 tag img
  copy_repo "$copy"; marker="ctx-canary-$(random_hex 12)"; printf '%s\n' "$marker" >>"$SECRETS_FILE"
  printf 'CANARY=%s\n' "$marker" >"$copy/core/.env"; printf 'CANARY=%s\n' "$marker" >"$copy/core/.env.local"
  for img in core migrate; do
    tag="g001-ctx-$img-$RUN_ID"
    if ledger_build "$LEDGER_SUITE" named:B-BUILD-CONTEXT -- --label g001.run=$RUN_ID -f "$copy/infra/docker/$img.Dockerfile" --target build -t "$tag" "$copy/core" >"$WORK/ctx-$img.log" 2>&1; then
      docker_run_owned ctx-leak --rm --label g001.run=$RUN_ID --entrypoint sh "$tag" -c "test ! -e /src/.env && test ! -e /src/.env.local && ! grep -rqF '$marker' /src" || { ok=0; log "$img build context leaked the environment file"; }
      docker image rm -f "$tag" >/dev/null 2>&1 || true
    else ok=0; tail -20 "$WORK/ctx-$img.log"; fi
  done
  # The browser image's fixture stage (context web/) builds the same source with a test route.
  printf 'CANARY=%s\n' "$marker" >"$copy/web/.env"; printf 'CANARY=%s\n' "$marker" >"$copy/web/.env.local"
  tag="g001-ctx-web-browser-$RUN_ID"
  if ledger_build "$LEDGER_SUITE" named:B-BUILD-CONTEXT -- --label g001.run=$RUN_ID -f "$copy/infra/docker/web-browser.Dockerfile" --target fixture -t "$tag" "$copy/web" >"$WORK/ctx-web-browser.log" 2>&1; then
    docker_run_owned ctx-leak --rm --label g001.run=$RUN_ID --entrypoint sh "$tag" -c "test ! -e /app/.env && test ! -e /app/.env.local && test ! -e /app/.next/standalone/.env && test ! -e /app/.next/standalone/.env.local && ! grep -rqF '$marker' --exclude-dir=node_modules /app && ! grep -rqF '$marker' /app/.next/standalone" || { ok=0; log "web-browser fixture stage leaked the environment file"; }
    docker image rm -f "$tag" >/dev/null 2>&1 || true
  else ok=0; tail -20 "$WORK/ctx-web-browser.log"; fi
  # The planted canary files are the test input, not a leak: remove the copy so
  # the final canary scan covers only logs, reports and other evidence.
  rm -rf "$copy"
  [[ $ok == 1 ]] && record B-BUILD-CONTEXT PASS "canary .env and .env.local placed in core/ and web/ of a disposable copy are absent from the core and migrate build stages and from the web-browser fixture stage (/app and its standalone output)" || record B-BUILD-CONTEXT FAIL "build context"
}

# ===================================================================================
# Assertion P (plan §8.5) and catalog facts, read as superuser over the
# PostgreSQL container's local socket (validation only).
catalog_p() { # catalog_p DEST -> "objects=2 owners=<db>_owner extra_acl=0 table_privs=0 seq_privs=0"
  local db; db="$(env_get "$1" DATABASE_NAME)"
  psql_super "$1" "
WITH t AS (SELECT 'app.goose_db_version'::regclass AS tbl, pg_get_serial_sequence('app.goose_db_version', 'id')::regclass AS seq),
objs AS (SELECT c.oid, c.relkind, c.relowner, c.relacl FROM pg_class c, t WHERE c.oid IN (t.tbl, t.seq)),
roles AS (SELECT unnest(ARRAY['${db}_runtime', '${db}_retention', '${db}_migrator', 'public']) AS r)
SELECT format('objects=%s owners=%s extra_acl=%s table_privs=%s seq_privs=%s',
  (SELECT count(*) FROM objs),
  (SELECT string_agg(DISTINCT relowner::regrole::text, ',') FROM objs),
  (SELECT count(*) FROM objs o, aclexplode(coalesce(o.relacl, acldefault(CASE WHEN o.relkind = 'S' THEN 's'::\"char\" ELSE 'r'::\"char\" END, o.relowner))) a WHERE a.grantee <> o.relowner),
  (SELECT count(*) FROM t, roles, unnest(ARRAY['SELECT','INSERT','UPDATE','DELETE','TRUNCATE','REFERENCES','TRIGGER']) p WHERE has_table_privilege(roles.r, t.tbl, p)),
  (SELECT count(*) FROM t, roles, unnest(ARRAY['USAGE','SELECT','UPDATE']) p WHERE has_sequence_privilege(roles.r, t.seq, p)))"
}
catalog_roles() { # catalog_roles DEST
  local db; db="$(env_get "$1" DATABASE_NAME)"
  psql_super "$1" "
SELECT format('app_owner=%s runtime_super=%s runtime_bypassrls=%s runtime_member_owner=%s runtime_member_migrator=%s runtime_owned=%s migrator_grant=%s app_objects_beyond_metadata=%s',
  (SELECT nspowner::regrole FROM pg_namespace WHERE nspname = 'app'),
  r.rolsuper, r.rolbypassrls, pg_has_role(r.oid, '${db}_owner', 'MEMBER'), pg_has_role(r.oid, '${db}_migrator', 'MEMBER'),
  (SELECT count(*) FROM pg_class WHERE relowner = r.oid),
  (SELECT format('inherit=%s,set=%s', m.inherit_option, m.set_option) FROM pg_auth_members m WHERE m.roleid = '${db}_owner'::regrole AND m.member = '${db}_migrator'::regrole),
  (SELECT count(*) FROM pg_class c WHERE c.relnamespace = 'app'::regnamespace
     AND c.oid NOT IN ('app.goose_db_version'::regclass, pg_get_serial_sequence('app.goose_db_version', 'id')::regclass)
     AND NOT EXISTS (SELECT 1 FROM pg_index i WHERE i.indexrelid = c.oid AND i.indrelid = 'app.goose_db_version'::regclass)))
FROM pg_roles r WHERE r.rolname = '${db}_runtime'"
}

# G064 (Plan r2 §3.1): the existing app relation pin at the top migration version (catalog_roles'
# app_objects_beyond_metadata), hoisted from the mig_step literals so mig_step, B-MIGRATE-CATALOG and
# db_pristine_verify assert one value; the history of the number is the comment in suite_migrations. G020: 88 at the
# top version 20 (the version-19 rollback and inspect steps keep their literal 79). G057: 89 at the top version 21 (the
# version-20 rollback and inspect steps keep their literal 88). G071: still 89 at the top version 22 (00022 adds one
# function, no relation), so the version-21 rollback and inspect steps also assert $MIG_RELATIONS.
MIG_RELATIONS=89

# G064 B-IT-ISOLATION (Plan r2 §6): T-1, T-1a, T-1b, T-1c, T-2, T-3, T-4, T-5a-c, T-6 (reporter fail-closed) and the T-7
# PG-010 line right after db_template_create on the same DEST and TREE; T-8 (report presence) when the row is recorded. Every
# control is reversible: <db> ends recreated from the template and the shared catalog equal to its snapshot.
# Per-control lines: $WORK/it-isolation.log. it_isolation = it_isolation_controls + it_isolation_record (the S1 private
# run); the migrations suite runs the controls after the template, T-2 again after B-SIGNUP-CATALOG (Plan r2 §3.3, §4
# #50: it_isolation_t2_repeat) and records the row once after that.
ITI_FAILS=() ITI_SUMMARY=() ITI_UNMET=() ITI_TPL=""
it_isolation() { # it_isolation DEST PROJECT TREE STATUS_LOG
  it_isolation_controls "$@"
  it_isolation_record
}
it_isolation_controls() { # it_isolation_controls DEST PROJECT TREE STATUS_LOG
  local dest="$1" project="$2" tree="$3" slog="$4" db tpl rec rt itlog="$WORK/it-isolation.log"
  db="$(env_get "$dest" DATABASE_NAME)"; tpl="${db}_tpl"; rec="$WORK/db-lifecycle/$tpl.rec"; rt="${db}_runtime"
  ITI_FAILS=() ITI_SUMMARY=() ITI_UNMET=() ITI_TPL="$tpl"
  : >"$itlog"
  _iti_t1; _iti_t1a; _iti_t1b; _iti_t1c; _iti_t2; _iti_t3; _iti_t4; _iti_t5; _iti_t6; _iti_t7; _iti_t9; _iti_t10
}
# G064 T-9/T-10 (addendum A §A3; correction 1 §C3/§C4) run in throwaway batches (t9, t10) of the B-MIG runner shape, so
# the B-MIG batch's fault count is untouched. _iti_runner CONTROL BATCH: start the batch runner; on refusal the control is
# recorded here (return 1): the §C4 precondition refusal outside CI is UNMET "not exercised" (no runner can exist, every
# eligible invocation takes the per-invocation container path; unexecuted coverage is never PASS, so B-IT-ISOLATION is
# then UNMET), any other refusal is FAIL.
ITI_RWANT=""
_iti_runner() {
  ITI_RWANT="$(_db_rec "$rec" core_check_image 2>/dev/null | head -1)"
  it_runner_start "$2" "$project-core-check" "${project}_data" "$ITI_RWANT" && return 0
  if [[ "$ITR_WHY" == no-registry || "$ITR_WHY" == no-durable-recovery ]] && [[ "${GITLAB_CI:-}" != true ]]; then
    _iti "$1" UNMET "not exercised: runner unavailable ($ITR_WHY); eligible invocations take the per-invocation container path"
  else _iti "$1" FAIL "runner start refused: $ITR_WHY"; fi
  it_runner_stop "$2" || true
  return 1
}
# T-9 runner residue: exec A leaves a /tmp marker, a $HOME marker and a background sleep; the post-check reports the
# lingering process and the docker diff path (before any reset), the runner is replaced after verified teardown, and
# exec B on the next generation sees neither marker.
_iti_t9() {
  local b=t9 id1 id2 fa ra why="" o="$WORK/it-isolation-t9.out"
  local -a GO_IT_XARGS=()
  _iti_runner T-9 "$b" || return 0
  id1="$(cat "$WORK/it-runner/$b.id")"
  # shellcheck disable=SC2016 # expanded by the runner's sh
  it_runner_exec "$b" "$o" sh -c ': G064-T9-A; echo a >/tmp/g064-t9; echo a >"$HOME/g064-t9"; sleep 300 </dev/null >/dev/null 2>&1 & exit 0' && why+="residue not detected; "
  fa="$ITR_FINDING" ra="$ITR_RC"
  [[ "$fa" == "runner-residue: process"* && "$fa" == *"runner-residue: file "*/g064-t9* ]] || why+="finding '$fa'; "
  [[ "$ra" != 0 ]] || why+="clean rc after residue; "
  [[ -e "$WORK/it-runner/$b.gen1.gone" && ! -e "$WORK/it-runner/$b.held" ]] || why+="generation 1 teardown not verified; "
  if it_runner_ready "$b" "$project-core-check" "${project}_data" "$ITI_RWANT"; then
    id2="$ITR_ID"
    [[ "$id2" != "$id1" && "$(cat "$WORK/it-runner/$b.gen")" == 2 ]] || why+="runner not replaced; "
    # shellcheck disable=SC2016 # expanded by the runner's sh
    it_runner_exec "$b" "$o" sh -c ': G064-T9-B; test ! -e /tmp/g064-t9 && test ! -e "$HOME/g064-t9"' || why+="exec B post-check: $ITR_FINDING; "
    [[ "$ITR_RC" == 0 ]] || why+="exec B saw a marker (rc=$ITR_RC); "
  else why+="replacement refused: $ITR_WHY; "; fi
  it_runner_stop "$b" || why+="final teardown unverified; "
  rm -f -- "$o"
  if [[ -z "$why" ]]; then _iti T-9 PASS "runner residue: lingering process and \$HOME marker reported before any reset ($fa), runner ${id1:0:12} removed and verified, replacement ${id2:0:12} saw neither marker"
  else _iti T-9 FAIL "${why%; }"; fi
}
# T-10 runner fault: the runner is removed (owned, exact ID) while an exec runs `sleep 30`; the exec path returns
# non-zero with runner-fault and the faulted generation's teardown is verified (no hold).
_iti_t10() {
  local b=t10 id ep i res="$WORK/it-isolation-t10.res" o="$WORK/it-isolation-t10.out" why="" rc f seen=0 top out
  local -a GO_IT_XARGS=()
  _iti_runner T-10 "$b" || return 0
  id="$(cat "$WORK/it-runner/$b.id")"; rm -f -- "$res"
  ( it_runner_exec "$b" "$o" sh -c ': G064-T10; sleep 30' || true; printf '%s\t%s\n' "$ITR_RC" "$ITR_FINDING" >"$res" ) &
  ep=$!
  for ((i = 0; i < 100; i++)); do
    if top="$(docker_top_args "$id")" && grep -qx 'sleep 30' <<<"$top"; then seen=1; break; fi
    sleep 0.2
  done
  ((seen)) || why+="exec never observed; "
  # The exec path's fault teardown may already be removing it ("removal ... already in progress"): not a failure here;
  # the teardown's verified-gone proof decides below.
  out="$(docker rm -f "$id" 2>&1)" || docker_rm_pending "$out" || why+="runner removal failed: $(docker_msg_line "$out"); "
  wait "$ep" 2>/dev/null || true
  IFS=$'\t' read -r rc f <"$res" 2>/dev/null || why+="exec path left no status; "
  [[ -n "${rc:-}" && "$rc" != 0 && "${f:-}" == runner-fault:* ]] || why+="exec path rc=${rc:-?} finding '${f:-}'; "
  [[ -e "$WORK/it-runner/$b.gen1.gone" && ! -e "$WORK/it-runner/$b.held" ]] || why+="faulted generation teardown not verified; "
  it_runner_stop "$b" || why+="final teardown unverified; "
  rm -f -- "$res" "$o"
  if [[ -z "$why" ]]; then _iti T-10 PASS "runner fault: runner ${id:0:12} removed during an exec; exec path rc=$rc with ${f%%:*} (${f#runner-fault: }), teardown verified"
  else _iti T-10 FAIL "${why%; }"; fi
}
# T-6 reporter fail-closed (Plan r2 §6, S3): every converter fixture (the passing multi-package stream with exact text and
# per-package rows; one stream per fail-closed class) through `gotestreport selftest` in the validation tools image, and
# the command's own stdin/tar/exit contract on one passing and one failing (omitted expected test) fixture.
_iti_t6() {
  local out="" n bad t6="$WORK/it-isolation-t6" st=0 pst=0 ost=0 prows
  if ! ensure_tools; then _iti T-6 FAIL "validation tools unavailable"; return 0; fi
  out="$(vtools "$VT_IMG" gotestreport selftest 2>&1)" || st=$?
  n="$(grep -c '^PASS ' <<<"$out" || true)"; bad="$(grep '^FAIL ' <<<"$out" | head -3 | tr '\n' ' ' || true)"
  rm -rf -- "$t6"; mkdir -p "$t6"
  vtools "$VT_IMG" gotestreport fixture pass-multi-package >"$t6/pass.events" 2>/dev/null || true
  vtools "$VT_IMG" gotestreport fixture -text >"$t6/pass.want" 2>/dev/null || true
  vtools "$VT_IMG" gotestreport fixture omitted-expected >"$t6/omit.events" 2>/dev/null || true
  vtools -i "$VT_IMG" gotestreport --name t6-pass --rc 0 <"$t6/pass.events" >"$t6/pass.tar" 2>/dev/null || pst=$?
  vtools -i "$VT_IMG" gotestreport --name t6-omit --rc 0 <"$t6/omit.events" >"$t6/omit.tar" 2>/dev/null || ost=$?
  tar -x -C "$t6" -f "$t6/pass.tar" 2>/dev/null || true
  tar -x -C "$t6" -f "$t6/omit.tar" 2>/dev/null || true
  prows="$(awk -F'\t' '$1 == "package" { printf "%s=%s ", $3, $5 }' "$t6/t6-pass.tsv" 2>/dev/null)"
  if [[ $st == 0 && "$n" -ge 20 && -z "$bad" && $pst == 0 && $ost == 1 ]] && cmp -s "$t6/t6-pass.log" "$t6/pass.want" &&
     [[ "$prows" == *"identity/identitytest=skip-no-test-files"* && "$(grep -c '=pass ' <<<"$prows")" == 1 ]] &&
     grep -q '^fail findings=.*omitted:core/internal/identity/workos.TestW2' "$t6/t6-omit.status"; then
    _iti T-6 PASS "gotestreport: $n fixtures as expected (every fail-closed class non-zero; the multi-package stream with its no-test-files package accepted, per-package rows, text byte-equal); CLI exit 0 pass / 1 omitted"
  else
    _iti T-6 FAIL "selftest rc=$st fixtures=$n ${bad:+failed: $bad}; cli pass=$pst omit=$ost rows: ${prows:-none}"
  fi
  rm -rf -- "$t6"
}
# T-8 report presence and completeness so far (Plan r2 §6): the four report files exist (REPORT_DIR, else the work
# copies) and every invocation converted so far has its invocation row and its go-tests/NAME.log.
_iti_t8() {
  local dir="${REPORT_DIR:-$WORK}" f miss="" n=0 name logs="${REPORT_DIR:+$REPORT_DIR/go-tests}"
  for f in go-tests.tsv go-tests-expected.tsv go-tests-junit.xml db-lifecycle.tsv; do [[ -s "$dir/$f" ]] || miss+="$f "; done
  [[ -n "$logs" ]] || logs="$WORK/go-tests"
  if [[ -f "$WORK/go-tests/started.tsv" ]]; then
    while IFS=$'\t' read -r name _; do
      [[ -f "$WORK/go-tests/$name.done" ]] || continue
      n=$((n + 1))
      grep -q $'^invocation\t'"$name"$'\t' "$dir/go-tests.tsv" 2>/dev/null || miss+="row:$name "
      [[ -f "$logs/$name.log" ]] || miss+="log:$name "
    done <"$WORK/go-tests/started.tsv"
  fi
  if [[ -z "$miss" && $n -gt 0 ]]; then _iti T-8 PASS "report files present in ${REPORT_DIR:+REPORT_DIR}${REPORT_DIR:-the work directory (no REPORT_DIR)}; $n invocations so far each with its row and go-tests log"
  else _iti T-8 FAIL "missing: ${miss:-no converted invocation}"; fi
}
# it_isolation_t2_repeat DEST PROJECT AFTER LOG: T-2 again on the fresh lifetime after the role-mutating invocation AFTER.
# Its db_fresh is the next invocation's reset point: a failure writes the reason to LOG and returns 1 (that invocation
# must not run), as it_fresh does.
it_isolation_t2_repeat() {
  local dest="$1" project="$2" after="$3" logf="$4" db tpl rec rt itlog="$WORK/it-isolation.log" nf="${#ITI_FAILS[@]}"
  db="$(env_get "$dest" DATABASE_NAME)"; tpl="${db}_tpl"; rec="$WORK/db-lifecycle/$tpl.rec"; rt="${db}_runtime"
  it_drain "$dest" "$project" "t2r-after-$after"
  _iti_t2 T-2r "t2r-after-$after"
  ((${#ITI_FAILS[@]} == nf)) && return 0
  printf 'G064 T-2 repeat (db_fresh) after %s failed; the invocation did not run (it-isolation.log)\n' "$after" >"$logf"
  return 1
}
# it_isolation_record [REASON]: the one B-IT-ISOLATION row; REASON names a control that could not run (FAIL).
it_isolation_record() {
  [[ -z "${1:-}" ]] || ITI_FAILS+=("$1")
  local itlog="$WORK/it-isolation.log"
  _iti_t8
  if ((${#ITI_FAILS[@]} == 0 && ${#ITI_UNMET[@]} == 0 && ${#ITI_SUMMARY[@]} > 0)); then
    record B-IT-ISOLATION PASS "run-private template $ITI_TPL and canonical-name recreation: $(IFS=';'; printf '%s' "${ITI_SUMMARY[*]}")"
  elif ((${#ITI_FAILS[@]} == 0 && ${#ITI_UNMET[@]} > 0 && ${#ITI_SUMMARY[@]} > 0)); then
    record B-IT-ISOLATION UNMET "required controls not exercised: ${ITI_UNMET[*]} (it-isolation.log)"
  else
    record B-IT-ISOLATION FAIL "controls failed: ${ITI_FAILS[*]:-no control ran}${ITI_UNMET[*]:+; not exercised: ${ITI_UNMET[*]}} (it-isolation.log)"
  fi
}
_iti() { # CONTROL STATUS DETAIL (it_isolation_controls' dynamic scope)
  printf '%s\t%s\t%s\n' "$1" "$2" "$3" >>"$itlog"
  case "$2" in PASS) ITI_SUMMARY+=("$1 ${3%%;*}") ;; UNMET) ITI_UNMET+=("$1") ;; *) ITI_FAILS+=("$1") ;; esac
  log "B-IT-ISOLATION $1 $2 - $3"
}

# G064 (Plan r2 §3.1, §3.6): it_drain DEST PROJECT PURPOSE gives the client sessions of a just-finished invocation a
# bounded time to end (polling only: nothing is cancelled or terminated); a non-zero wait is a PG-010 observation row.
# db_fresh still refuses if any session remains.
IT_DRAIN_POLLS=50 # x 0.2 s
it_drain() {
  local dest="$1" project="$2" purpose="$3" db s="" i t0
  db="$(env_get "$dest" DATABASE_NAME)"; t0="$(_db_ms)"
  for ((i = 0; i < IT_DRAIN_POLLS; i++)); do
    s="$(_db_sessions "$dest" "$db" 2>/dev/null)" || s="?"
    [[ "${s%%$'\t'*}" == 0 ]] && break
    sleep 0.2
  done
  ((i == 0)) || _db_ledger wait "$db" sessions "caller=$project purpose=$purpose polls=$i wait_ms=$(($(_db_ms) - t0)) last=$(_db_oneline "${s#*$'\t'}")" || true
}
# G064 (Plan r2 §4): it_fresh DEST PROJECT PURPOSE LOG: db_fresh immediately before one independent Go invocation (or
# the seed that starts an intentional group). A refusal writes its reason to LOG, so the check's PASS parse fails
# closed with the reason in its tail, and returns 1: the invocation must not run on a database that is not fresh.
# Status lines of the it_* helpers go to stderr (same tee; see go_it_convert) so a captured caller never reads them.
it_fresh() {
  local dest="$1" project="$2" purpose="$3" logf="$4"
  it_drain "$dest" "$project" "$purpose"
  DB_LC_CALLER="$project" db_fresh "$dest" "$project" "$purpose" && return 0
  printf 'G064 db_fresh refused before %s; the invocation did not run: %s\n' "$purpose" "$DB_LC_ERR" >"$logf"
  log "db_fresh before $purpose refused: $DB_LC_ERR" >&2
  return 1
}
_iti_n() { psql_maint "$dest" "SELECT count(*) FROM pg_database WHERE datname = :'n';" n="$1" 2>/dev/null || echo "?"; }
# _iti_login ROLE_SUFFIX DBNAME [SQL]: psql as <db>_<suffix> over TCP inside the container, with the credential the
# postgres container already holds (printenv there; never on a host command line).
_iti_login() {
  printf '%s\n' "${3:-SELECT 1;}" | compose "$dest" exec -T postgres sh -c 'PGPASSWORD="$(printenv "$1")" exec psql -X -q -tA -v ON_ERROR_STOP=1 -h 127.0.0.1 -U "$2" -d "$3" -v c="$4" -v t="$5" -f -' \
    sh "DB_${1^^}_PASSWORD" "${db}_$1" "$2" "${db}_ctlclone" "$tpl"
}

# T-1 template record: source binding, the goose-history predicate, metadata/role facts, closed template, baseline.
_iti_t1() {
  local nok=() f v nb top_rows
  [[ -s "$rec" ]] || { _iti T-1 FAIL "no template record $rec"; return 0; }
  for f in name mig_top bound mig_digest roles_sh_digest provision_sql_digest tree_hash head tree_state core_check_image migrate_image \
    bound_digest history_digest fingerprint_catalog fingerprint_roles baseline_digest size_bytes duration_ms; do
    [[ -n "$(_db_rec "$rec" "$f")" ]] || nok+=("record field $f empty")
  done
  if _db_bound "$tree" && _db_source "$tree"; then
    [[ "$(_db_rec "$rec" mig_top)" == "$DB_MIG_TOP" && "$(_db_rec "$rec" bound)" == "$DB_BOUND" ]] || nok+=("record migration set differs from the tree")
    [[ "$(_db_rec "$rec" mig_digest)" == "$DB_SRC_MIG" ]] || nok+=("mig_digest differs")
    [[ "$(_db_rec "$rec" roles_sh_digest) $(_db_rec "$rec" provision_sql_digest)" == "$DB_SRC_ROLES $DB_SRC_PROV" ]] || nok+=("provisioning digests differ")
    [[ "$(_db_rec "$rec" tree_hash)" == "$DB_SRC_TREE" ]] || nok+=("tested tree hash differs")
  else nok+=("source binding: $DB_LC_ERR"); fi
  nb="$(wc -w <<<"$DB_BOUND")"
  v="$(_db_rec "$rec" status | awk -F'\t' '$1 == "migration" && $3 == "applied" && $4 != "" { n++ } $1 == "current_version" { cv = $2 } END { print n + 0 " " cv }')"
  [[ "$v" == "$nb $DB_MIG_TOP" ]] || nok+=("recorded status listing: $v applied/current (want $nb $DB_MIG_TOP)")
  _db_rec "$rec" history | awk -F'\t' '$2 == "0" { f = 1 } END { exit !f }' || nok+=("recorded history has no version 0 row")
  top_rows="$(_db_rec "$rec" history | awk -F'\t' -v t="$DB_MIG_TOP" '$2 == t { n++ } END { print n + 0 }')"
  _db_goose_check "$dest" "$db" app goose_db_version || nok+=("$DB_LC_ERR")
  v="$(catalog_p "$dest" 2>&1)" || true
  [[ "$v" == "objects=2 owners=${db}_owner extra_acl=0 table_privs=0 seq_privs=0" ]] || nok+=("catalog_p: $v")
  v="$(catalog_roles "$dest" 2>&1)" || true
  [[ "$v" == *" app_objects_beyond_metadata=$MIG_RELATIONS" ]] || nok+=("catalog_roles: $v")
  v="$(psql_maint "$dest" "SELECT format('%s,%s', datistemplate, datallowconn) FROM pg_database WHERE datname = :'n';" n="$tpl" 2>&1)" || true
  [[ "$v" == "t,f" ]] || nok+=("template flags: $v")
  v="$(_db_rec "$rec" shared | awk -v p="db:$tpl:" 'index($0, p) == 1')"
  [[ "$v" =~ ^db:${tpl}:postgres:f,t:(.*)$ ]] && ! tr ',' '\n' <<<"${BASH_REMATCH[1]}" | grep -qv '^postgres=' || nok+=("template ACL/flags in the snapshot: $v")
  v="$(_db_rec "$rec" baseline | awk -F'\t' '$2 != "0" { nz++ } $1 == "app.goose_db_version" { g = 1 } $1 == "app.identity_provider_binding" { p = 1 } { n++ }
    END { print n + 0 " " nz + 0 " " g + 0 " " p + 0 }')"
  [[ "$v" =~ ^[1-9][0-9]*\ 0\ 0\ 1$ ]] || nok+=("baseline tables/non-empty/goose/provider-binding: $v")
  if ((${#nok[@]})); then _iti T-1 FAIL "$(IFS='|'; printf '%s' "${nok[*]}")"; return 0; fi
  _iti T-1 PASS "record bound to tree $(_db_rec "$rec" tree_hash | cut -c1-12) head $(_db_rec "$rec" head | cut -c1-12) ($(_db_rec "$rec" tree_state)); MIG_TOP=$DB_MIG_TOP from $nb bound files, $nb applied status lines, applied set {0}+bound (latest row per version), raw history $(_db_rec "$rec" history | grep -c .) rows recorded (top-version rows $top_rows, not counted); catalog_p/catalog_roles with $MIG_RELATIONS relations; $(_db_rec "$rec" baseline | grep -c .) baseline tables all empty; template closed"
}

# T-1a goose-history and table-set negatives on scratch copies inside <db> (never app/audit), then dropped.
_iti_t1a() {
  local nok=() s=g064_ctl_hist ts=g064_ctl_ts low="${DB_BOUND%% *}" v
  if ! v="$(_psql_on "$dest" "$db" 'DROP SCHEMA IF EXISTS :"s" CASCADE; CREATE SCHEMA :"s";
CREATE TABLE :"s".g0 AS SELECT * FROM app.goose_db_version;
CREATE TABLE :"s".g1 AS SELECT * FROM app.goose_db_version WHERE version_id <> 0;
CREATE TABLE :"s".g2 AS SELECT * FROM app.goose_db_version;
INSERT INTO :"s".g2 (id, version_id, is_applied, tstamp) SELECT max(id) + 1, :top, false, now() FROM :"s".g2;
CREATE TABLE :"s".g3 AS SELECT * FROM app.goose_db_version WHERE version_id <> :low;' s="$s" top="$DB_MIG_TOP" low="$low" 2>&1)"; then
    _iti T-1a FAIL "scratch history copies: $(_db_oneline "$v")"; return 0
  fi
  _db_goose_check "$dest" "$db" "$s" g0 || nok+=("unmodified copy refused: $DB_LC_ERR")
  if _db_goose_check "$dest" "$db" "$s" g1; then nok+=("copy without the version 0 row accepted"); else [[ "$DB_LC_ERR" == *"version 0 row missing"* ]] || nok+=("g1: $DB_LC_ERR"); fi
  if _db_goose_check "$dest" "$db" "$s" g2; then nok+=("superseded is_applied=false top row accepted"); else [[ "$DB_LC_ERR" == *"bound version $DB_MIG_TOP not applied"* ]] || nok+=("g2: $DB_LC_ERR"); fi
  if _db_goose_check "$dest" "$db" "$s" g3; then nok+=("copy missing bound version $low accepted"); else [[ "$DB_LC_ERR" == *"bound version $low not applied"* ]] || nok+=("g3: $DB_LC_ERR"); fi
  if v="$(_psql_on "$dest" "$db" 'DROP SCHEMA IF EXISTS :"t" CASCADE; CREATE SCHEMA :"t";
CREATE TABLE :"t".tb (id int); INSERT INTO :"t".tb VALUES (1);
CREATE SEQUENCE :"t".sq; SELECT nextval(:'"'"'q'"'"');
CREATE VIEW :"t".vw AS SELECT id FROM :"t".tb; CREATE INDEX tb_id ON :"t".tb (id);' t="$ts" q="$ts.sq" 2>&1)"; then
    v="$(_db_baseline "$dest" "$db" "$ts" - 2>&1)" || true
    [[ "$v" == "$ts.tb"$'\t'1 ]] || nok+=("table-set enumeration on table+sequence+view+index: $(_db_oneline "$v") (want only $ts.tb with 1 row)")
  else nok+=("scratch table-set schema: $(_db_oneline "$v")"); fi
  v="$(_psql_on "$dest" "$db" 'DROP SCHEMA IF EXISTS :"s" CASCADE; DROP SCHEMA IF EXISTS :"t" CASCADE;' s="$s" t="$ts" 2>&1)" || nok+=("scratch schemas not dropped: $(_db_oneline "$v")")
  if ((${#nok[@]})); then _iti T-1a FAIL "$(IFS='|'; printf '%s' "${nok[*]}")"; return 0; fi
  _iti T-1a PASS "history copy without version 0, with a superseded is_applied=false top row and without bound version $low each refused by name, unmodified copy accepted; enumeration of table+sequence+view+index lists only the table (1 row)"
}

# T-1b stale/source-mismatch refusal: db_fresh with a record copy whose mig_digest is altered drops nothing.
_iti_t1b() {
  local copy="$WORK/db-lifecycle/t1b-altered.rec" o1 o2 rc=0 why
  o1="$(_db_oid "$dest" "$db" 2>/dev/null)" || o1=""
  awk -F'\t' 'BEGIN { OFS = "\t" } $1 == "mig_digest" { $2 = "0000000000000000000000000000000000000000000000000000000000000000" } { print }' "$rec" >"$copy"
  DB_LC_CALLER=T-1b db_fresh "$dest" "$project" t1b-stale-record "$copy" && rc=1
  why="$DB_LC_ERR"; rm -f -- "$copy"
  o2="$(_db_oid "$dest" "$db" 2>/dev/null)" || o2=""
  if [[ $rc == 0 && "$why" == *"stale or source-mismatched"*mig_digest* && -n "$o1" && "$o1" == "$o2" ]]; then
    _iti T-1b PASS "altered mig_digest refused before any drop ($db oid $o1 unchanged)"
  else _iti T-1b FAIL "rc=$rc oid $o1->$o2: $why"; fi
}

# T-1c dirty-baseline refusal: a provider-binding row makes db_template_create refuse (no control template), and the
# following db_fresh removes it.
_iti_t1c() {
  local rc=0 why n cnt v
  v="$(_psql_on "$dest" "$db" "INSERT INTO app.identity_provider_binding (provider, issuer, client_id) VALUES ('workos', 'https://g064-ctl.invalid', 'g064-ctl');" 2>&1)" ||
    { _iti T-1c FAIL "planting the provider binding failed: $(_db_oneline "$v")"; return 0; }
  DB_LC_CALLER=T-1c db_template_create "$dest" "$project" "$tree" "$slog" "${db}_ctl" && rc=1
  why="$DB_LC_ERR"; n="$(_iti_n "${db}_ctl")"
  DB_LC_CALLER=T-1c db_fresh "$dest" "$project" t1c-cleanup || rc=2
  cnt="$(_psql_on "$dest" "$db" 'SELECT count(*) FROM app.identity_provider_binding;' 2>&1)" || cnt="?"
  if [[ $rc == 0 && "$why" == *"baseline rows not empty"*"app.identity_provider_binding=1"* && "$n" == 0 && "$cnt" == 0 ]]; then
    _iti T-1c PASS "planted provider binding refused (no ${db}_ctl created), removed by the next db_fresh"
  else _iti T-1c FAIL "rc=$rc ${db}_ctl=$n bindings after db_fresh=$cnt: $why${DB_LC_ERR:+ / $DB_LC_ERR}"; fi
}

# T-2 marker isolation: a marker Organization in <db> is absent after db_fresh; the copy's fingerprint equals the record.
_iti_t2() { # [ID PURPOSE] (default T-2 t2-marker; T-2r is the repeat after a role-mutating invocation)
  local tid="${1:-T-2}" purpose="${2:-t2-marker}" id n1 n2 rc=0 why=""
  id="$(_psql_on "$dest" "$db" "INSERT INTO app.organizations (id, name, slug) VALUES (gen_random_uuid(), 'G064 marker', :'slug') RETURNING id;" slug="g064-marker-$(random_hex 4)" 2>&1)" ||
    { _iti "$tid" FAIL "marker insert failed: $(_db_oneline "$id")"; return 0; }
  n1="$(_psql_on "$dest" "$db" "SELECT count(*) FROM app.organizations WHERE id = :'id';" id="$id" 2>&1)" || n1="?"
  DB_LC_CALLER="$tid" db_fresh "$dest" "$project" "$purpose" || { rc=1; why="$DB_LC_ERR"; }
  n2="$(_psql_on "$dest" "$db" "SELECT count(*) FROM app.organizations WHERE id = :'id';" id="$id" 2>&1)" || n2="?"
  if [[ $rc == 0 && "$n1" == 1 && "$n2" == 0 ]]; then
    _iti "$tid" PASS "marker Organization present before, absent after db_fresh; copy fingerprint, history and shared catalog equal the template record"
  else _iti "$tid" FAIL "marker before=$n1 after=$n2 rc=$rc $why"; fi
}

# T-3 active-session refusal: a held session on <db> makes db_fresh refuse without dropping; the session survives and
# is then released by cancelling its own query (owned by this control; nothing is terminated).
_iti_t3() {
  local app hold i n=0 o1 o2 rc=0 why alive left="?"
  app="g064-t3-$(random_hex 4)"
  compose "$dest" exec -T -e PGAPPNAME="$app" postgres psql -X -q -tA -U postgres -d "$db" -c 'SELECT pg_sleep(120)' >/dev/null 2>&1 &
  hold=$!
  for ((i = 0; i < 150; i++)); do
    n="$(psql_maint "$dest" "SELECT count(*) FROM pg_stat_activity WHERE datname = :'d' AND application_name = :'a';" d="$db" a="$app" 2>/dev/null)" || n=0
    [[ "$n" == 1 ]] && break
    sleep 0.2
  done
  o1="$(_db_oid "$dest" "$db" 2>/dev/null)" || o1=""
  DB_LC_CALLER=T-3 db_fresh "$dest" "$project" t3-held-session && rc=1
  why="$DB_LC_ERR"
  o2="$(_db_oid "$dest" "$db" 2>/dev/null)" || o2=""
  alive="$(psql_maint "$dest" "SELECT count(*) FROM pg_stat_activity WHERE datname = :'d' AND application_name = :'a';" d="$db" a="$app" 2>/dev/null)" || alive="?"
  psql_maint "$dest" "SELECT count(pg_cancel_backend(pid)) FROM pg_stat_activity WHERE datname = :'d' AND application_name = :'a';" d="$db" a="$app" >/dev/null 2>&1 || true
  wait "$hold" 2>/dev/null || true
  for ((i = 0; i < 50; i++)); do
    left="$(_db_sessions "$dest" "$db" 2>/dev/null)" || left="?"
    [[ "${left%%$'\t'*}" == 0 ]] && break
    sleep 0.2
  done
  if [[ "$n" == 1 && $rc == 0 && "$why" == *"client session(s) on $db"* && -n "$o1" && "$o1" == "$o2" && "$alive" == 1 && "${left%%$'\t'*}" == 0 ]]; then
    _iti T-3 PASS "held session refused (no drop, no termination; $db oid $o1 unchanged); session survived, then released by cancelling its own query"
  else _iti T-3 FAIL "held=$n rc=$rc oid $o1->$o2 alive=$alive after-release=$(_db_oneline "$left"): $why"; fi
}

# T-4 template closed: the three login roles reach <db> (positive control) but not the template; the migrator cannot
# clone it.
_iti_t4() {
  local r out nok=() t4_seen=()
  for r in migrator runtime retention; do
    out="$(_iti_login "$r" "$db" 2>&1)" || true
    [[ "$out" == 1 ]] || nok+=("${db}_$r cannot reach $db (positive control): $(_db_oneline "$out")")
    if out="$(_iti_login "$r" "$tpl" 2>&1)"; then nok+=("${db}_$r logged in to $tpl"); continue; fi
    case "$out" in
      *"is not currently accepting connections"* | *"permission denied for database"*) t4_seen+=("${db}_$r: $(grep -o 'FATAL: .*' <<<"$out" | head -1)") ;;
      *) nok+=("${db}_$r refused for another reason: $(_db_oneline "$out")") ;;
    esac
  done
  if out="$(_iti_login migrator "$db" 'CREATE DATABASE :"c" TEMPLATE :"t";' 2>&1)"; then nok+=("${db}_migrator created a database from $tpl")
  elif [[ "$out" != *"permission denied to create database"* ]]; then nok+=("clone refused for another reason: $(_db_oneline "$out")"); fi
  [[ "$(_iti_n "${db}_ctlclone")" == 0 ]] || nok+=("${db}_ctlclone exists")
  if ((${#nok[@]})); then _iti T-4 FAIL "$(IFS='|'; printf '%s' "${nok[*]}")"; return 0; fi
  _iti T-4 PASS "logins to $tpl refused for ${db}_migrator, ${db}_runtime, ${db}_retention (each reaches $db); clone by ${db}_migrator refused (permission denied to create database); $(IFS=';'; printf '%s' "${t4_seen[*]}")"
}

# T-5a/b/c shared-catalog detector: each reversible superuser delta is reported exactly, makes db_fresh refuse
# (nothing dropped) and is reverted to the snapshot.
_iti_t5_case() { # ID APPLY-SQL REVERT-SQL EXPECTED-DELTA-PREFIX...
  local id="$1" apply="$2" revert="$3" s0 s1 s2 d e o1 o2 nok=(); shift 3
  s0="$(db_shared_catalog "$dest" "$tpl" 2>&1)" || nok+=("snapshot unreadable")
  psql_maint "$dest" "$apply" r="$rt" d="$db" >/dev/null 2>&1 || nok+=("apply failed")
  s1="$(db_shared_catalog "$dest" "$tpl" 2>&1)" || true
  d="$(_db_delta "$s0" "$s1")"
  [[ "$(grep -c . <<<"$d" || true)" == "$#" ]] || nok+=("delta has $(grep -c . <<<"$d" || true) lines, want $#")
  for e in "$@"; do [[ "$(awk -v p="$e" 'index($0, p) == 1 { n++ } END { print n + 0 }' <<<"$d")" == 1 ]] || nok+=("no single delta line $e"); done
  o1="$(_db_oid "$dest" "$db" 2>/dev/null)" || o1=""
  if DB_LC_CALLER="$id" db_fresh "$dest" "$project" "${id,,}-delta"; then nok+=("db_fresh accepted the delta")
  elif [[ "$DB_LC_ERR" != *"shared catalog delta"* ]]; then nok+=("db_fresh refused for another reason: $DB_LC_ERR"); fi
  o2="$(_db_oid "$dest" "$db" 2>/dev/null)" || o2=""
  [[ -n "$o1" && "$o1" == "$o2" ]] || nok+=("$db oid $o1->$o2")
  psql_maint "$dest" "$revert" r="$rt" d="$db" >/dev/null 2>&1 || nok+=("revert failed")
  s2="$(db_shared_catalog "$dest" "$tpl" 2>&1)" || true
  [[ "$s2" == "$s0" ]] || nok+=("not restored: $(_db_oneline "$(_db_delta "$s0" "$s2")")")
  if ((${#nok[@]})); then _iti "$id" FAIL "$(IFS='|'; printf '%s' "${nok[*]}") delta: $(_db_oneline "$d")"; return 0; fi
  _iti "$id" PASS "delta reported exactly ($(_db_oneline "$d")), db_fresh refused without dropping, reverted to the snapshot"
}
_iti_t5() {
  local inh a1 a2
  inh="$(psql_maint "$dest" "SELECT rolinherit FROM pg_roles WHERE rolname = :'r';" r="$rt" 2>/dev/null)" || inh="?"
  if [[ "$inh" == t ]]; then a1=NOINHERIT a2=INHERIT; else a1=INHERIT a2=NOINHERIT; fi
  _iti_t5_case T-5a "ALTER ROLE :\"r\" $a1;" "ALTER ROLE :\"r\" $a2;" "-role:$rt:" "+role:$rt:"
  _iti_t5_case T-5b 'GRANT pg_read_all_data TO :"r";' 'REVOKE pg_read_all_data FROM :"r";' "+member:$rt>pg_read_all_data:"
  _iti_t5_case T-5c "ALTER ROLE :\"r\" IN DATABASE :\"d\" SET track_functions = 'pl';" 'ALTER ROLE :"r" IN DATABASE :"d" RESET track_functions;' "+setting:$rt:$db:track_functions=pl"
}

# T-7 PG-010 observations so far (template time and size, db_fresh durations, refusals); informational, PASS when read.
_iti_t7() {
  local f="$WORK/db-lifecycle.tsv" v
  [[ -s "$f" ]] || { _iti T-7 FAIL "no lifecycle ledger"; return 0; }
  v="$(awk -F'\t' -v t="$tpl" '
    $2 == "template" && $3 == t && $4 == "create" { tpl = $5 }
    $2 == "fresh" && $4 == "PASS" { n++; if (match($5, /total_ms=[0-9]+/)) ms = ms (ms == "" ? "" : ",") substr($5, RSTART + 9, RLENGTH - 9) }
    $2 == "refuse" { r++ }
    END { sub(/.*size_bytes=/, "size_bytes=", tpl); printf "PG-010 template %s; db_fresh %d (total_ms %s); refusals %d", tpl, n, (ms == "" ? "-" : ms), r }' "$f")"
  _iti T-7 PASS "$v"
}

suite_migrations() {
  log "== migrations (disposable project)"
  local M="$WORK/mig" project="g001-mig-$RUN_ID" db owner ok=1 p_expect r_expect line out step
  if ports_busy "${PORTS_MIG[@]}"; then record B-MIGRATE-CLI FAIL "ports ${PORTS_MIG[*]} busy; refusing to re-map silently"; return; fi
  copy_repo "$M"
  configure "$M" "$project" "${PORTS_MIG[@]}" || { record B-MIGRATE-CLI FAIL "configure"; return; }
  # G064 (Plan r2 §5.1, §5.3): the copied tree's provenance and the declared invocation inventory, before any invocation.
  go_it_source "$M"; go_it_declare migrations
  db="$(env_get "$M" DATABASE_NAME)"; owner="${db}_owner"
  p_expect="objects=2 owners=$owner extra_acl=0 table_privs=0 seq_privs=0"
  # app_objects_beyond_metadata: 0 below version 2; at version 2 the 14 reviewed G004 identity objects (5 tables, 9
  # indexes), which B-IDENTITY-CATALOG names exactly; at version 3 those plus the 13 G005 tenancy relations (3 tables,
  # 2 views, 8 indexes), which B-TENANCY-CATALOG names exactly: 27; at version 4 those plus the 8 G006 BusinessProfile
  # relations (2 tables, 1 view, 5 indexes), which B-BUSINESS-PROFILE-CATALOG names exactly: 35; at version 5 still 35
  # (the G007 policy flag is a column of app.organizations, which B-POLICY-CATALOG names exactly); at version 6 still 35
  # (G008 adds policies, functions and grants only, which B-RLS-CATALOG names exactly); at version 7 38 (the G009
  # stored-object table and its two indexes, which B-OBJ-CATALOG names exactly); at version 8 still 38 (the G011 audit
  # relation, index and functions live in schema audit, which B-AUDIT-CATALOG names exactly); at version 9 still 38 (G013 adds
  # one function, policies and grants, which B-RETENTION-CATALOG names exactly); at version 10 42 (the G014 connection
  # table and its three unique indexes, which B-MCP-CATALOG names exactly); at version 11 48 (the G016 domain table, its
  # four indexes and the Organization slug index, which B-DOMAINS-CATALOG names exactly); at version 12 still 48 (the
  # G016 security amendment adds two sessions columns and constraints and swaps one domain index for another); at
  # version 13 50 (the G017 branding table and its primary key, which B-BRANDING-CATALOG names exactly); at version 14
  # still 50 (G018 adds functions, policies, grants and two business_profiles columns, which B-ONBOARDING-CATALOG names);
  # at version 15 56 (the G019 invitation table, its two keys and three indexes, which B-MEMBERSHIP-CATALOG names); at
  # version 16 71 (the G058 enrollment credential, request and restriction tables, their keys, the six indexes and the
  # ledger's identity sequence, which B-ENROLLMENT-CATALOG names); at version 17 still 71 (G060 adds one organizations
  # column, a guard trigger, replaced policies and helper bodies and grants only; Plan r0 §3, r1 C1); at version 18 79 (the
  # G062 signup table, its primary key, five digest keys and one index, which B-SIGNUP-CATALOG names); at version 19 still
  # 79 (G063 adds one mcp_connections column, its CHECK, two column grants and a trigger, and replaces two audit bodies,
  # which B-MCP-CATALOG names); at version 20 88 (the G020 label tables, their primary keys, the definitions' scoped-name
  # and owner keys, the assignments' unique key and two indexes, which B-LABELS-CATALOG names); at version 21 89 (the
  # G057 Organization platform-label unique index; the label column, its two CHECKs, two column grants, the replaced
  # reader and provisioning bodies, the new slug reader and the widened writer policy add no relation, which
  # B-PLATFORM-ADDRESS-CATALOG names); at version 22 still 89 (G071 adds the auth reader's app.organization_id_by_slug
  # function and its grants only, which TestDefaultTenantCatalog names).
  # The seven steps run only at versions 0 (fresh version and status), 22 (apply, reapply, re-apply) and 21 (rollback,
  # inspect), so B-MIGRATE-CATALOG expects exactly the unique rows P, roles 0 and roles 89 (G071: the apply, reapply and
  # re-apply steps now run at version 22 (89), the rollback and inspect steps at version 21 (also 89), retiring the 88
  # row; G057 retired the 79 row, G063 the 71 row, G060 the 56 row, G018 staging repair R1 the 48).
  r_expect="app_owner=$owner runtime_super=f runtime_bypassrls=f runtime_member_owner=f runtime_member_migrator=f runtime_owned=0 migrator_grant=inherit=f,set=t app_objects_beyond_metadata="
  local doc="$M/docs/foundation/MIGRATIONS.md" P_LOG="$WORK/catalog-p.log"
  : >"$P_LOG"
  local status_line version_line
  status_line="$(docs_block "$doc" migrations:inspect | sed -n 1p)"; version_line="$(docs_block "$doc" migrations:inspect | sed -n 2p)"
  # mig_step NAME OBJECTS CMDS EXPECT_REGEX...: run verbatim lines, assert output lines, then P and the catalog with
  # OBJECTS identity objects.
  mig_step() {
    local name="$1" objects="$2" cmds="$3"; shift 3
    local lf="$WORK/mig-$name.log" re got
    (cd "$M" && bash -euo pipefail -c "$cmds") >"$lf" 2>&1 || { ok=0; log "migration step $name failed: $(tail -3 "$lf" | tr '\n' ' ')"; return; }
    for re in "$@"; do grep -qP "$re" "$lf" || { ok=0; log "step $name: no line matching $re"; }; done
    got="$(catalog_p "$M" 2>&1)"; printf '%s\t%s\n' "$name" "$got" >>"$P_LOG"
    [[ "$got" == "$p_expect" ]] || { ok=0; log "P after $name: $got"; }
    got="$(catalog_roles "$M" 2>&1)"; printf '%s\troles\t%s\n' "$name" "$got" >>"$P_LOG"
    [[ "$got" == "$r_expect$objects" ]] || { ok=0; log "catalog after $name: $got"; }
  }
  mig_step 1-fresh-version 0 "$version_line" '^current_version\t0$'
  compose "$M" down -v >/dev/null 2>&1 || { ok=0; log "down -v before the fresh-status step failed"; }
  docker volume ls -q --filter "label=com.docker.compose.project=$project" | grep -q . && { ok=0; log "volumes survived down -v before the fresh-status step"; }
  mig_step 2-fresh-status 0 "$status_line" '^migration\t00001_baseline\.sql\tpending' '^migration\t00002_identity\.sql\tpending' '^migration\t00003_tenancy\.sql\tpending' '^migration\t00004_business_profiles\.sql\tpending' '^migration\t00005_organization_policy\.sql\tpending' '^migration\t00006_row_level_security\.sql\tpending' '^migration\t00007_stored_objects\.sql\tpending' '^migration\t00008_audit_events\.sql\tpending' '^migration\t00009_retention_maintenance\.sql\tpending' '^migration\t00010_mcp_connections\.sql\tpending' '^migration\t00011_organization_domains\.sql\tpending' '^migration\t00012_session_binding_and_claim_exclusivity\.sql\tpending' '^migration\t00013_organization_branding\.sql\tpending' '^migration\t00014_group_onboarding\.sql\tpending' '^migration\t00015_membership_invitations\.sql\tpending' '^migration\t00016_organization_enrollment\.sql\tpending' '^migration\t00017_group_authority\.sql\tpending' '^migration\t00018_platform_signup\.sql\tpending' '^migration\t00019_mcp_connection_names\.sql\tpending' '^migration\t00020_labels\.sql\tpending' '^migration\t00021_platform_labels\.sql\tpending' '^migration\t00022_default_tenant\.sql\tpending' '^current_version\t0$'
  mig_step 3-apply "$MIG_RELATIONS" "$(docs_block "$doc" migrations:apply)" '^applied\t00001_baseline\.sql$' '^applied\t00002_identity\.sql$' '^applied\t00003_tenancy\.sql$' '^applied\t00004_business_profiles\.sql$' '^applied\t00005_organization_policy\.sql$' '^applied\t00006_row_level_security\.sql$' '^applied\t00007_stored_objects\.sql$' '^applied\t00008_audit_events\.sql$' '^applied\t00009_retention_maintenance\.sql$' '^applied\t00010_mcp_connections\.sql$' '^applied\t00011_organization_domains\.sql$' '^applied\t00012_session_binding_and_claim_exclusivity\.sql$' '^applied\t00013_organization_branding\.sql$' '^applied\t00014_group_onboarding\.sql$' '^applied\t00015_membership_invitations\.sql$' '^applied\t00016_organization_enrollment\.sql$' '^applied\t00017_group_authority\.sql$' '^applied\t00018_platform_signup\.sql$' '^applied\t00019_mcp_connection_names\.sql$' '^applied\t00020_labels\.sql$' '^applied\t00021_platform_labels\.sql$' '^applied\t00022_default_tenant\.sql$' '^current_version\t22$'
  mig_step 4-reapply "$MIG_RELATIONS" "$(docs_block "$doc" migrations:reapply)" '^rolled_back\t00022_default_tenant\.sql$' '^applied\t00022_default_tenant\.sql$' '^current_version\t22$'
  mig_step 5-rollback "$MIG_RELATIONS" "$(docs_block "$doc" migrations:rollback)" '^rolled_back\t00022_default_tenant\.sql$' '^current_version\t21$'
  mig_step 6-inspect "$MIG_RELATIONS" "$(docs_block "$doc" migrations:inspect)" '^migration\t00001_baseline\.sql\tapplied\t' '^migration\t00002_identity\.sql\tapplied\t' '^migration\t00003_tenancy\.sql\tapplied\t' '^migration\t00004_business_profiles\.sql\tapplied\t' '^migration\t00005_organization_policy\.sql\tapplied\t' '^migration\t00006_row_level_security\.sql\tapplied\t' '^migration\t00007_stored_objects\.sql\tapplied\t' '^migration\t00008_audit_events\.sql\tapplied\t' '^migration\t00009_retention_maintenance\.sql\tapplied\t' '^migration\t00010_mcp_connections\.sql\tapplied\t' '^migration\t00011_organization_domains\.sql\tapplied\t' '^migration\t00012_session_binding_and_claim_exclusivity\.sql\tapplied\t' '^migration\t00013_organization_branding\.sql\tapplied\t' '^migration\t00014_group_onboarding\.sql\tapplied\t' '^migration\t00015_membership_invitations\.sql\tapplied\t' '^migration\t00016_organization_enrollment\.sql\tapplied\t' '^migration\t00017_group_authority\.sql\tapplied\t' '^migration\t00018_platform_signup\.sql\tapplied\t' '^migration\t00019_mcp_connection_names\.sql\tapplied\t' '^migration\t00020_labels\.sql\tapplied\t' '^migration\t00021_platform_labels\.sql\tapplied\t' '^migration\t00022_default_tenant\.sql\tpending' '^current_version\t21$'
  mig_step 7-apply "$MIG_RELATIONS" "$(docs_block "$doc" migrations:apply)" '^applied\t00022_default_tenant\.sql$' '^current_version\t22$'
  # Production refuses destructive commands before connecting.
  for step in down reapply; do
    local rc=0; out="$(compose_migrate "$M" run --rm -e APP_ENV=production migrate "$step" 2>&1)" || rc=$?
    [[ $rc == 2 ]] && grep -q "refusing $step with APP_ENV=production" <<<"$out" || { ok=0; log "production $step: exit $rc"; }
  done
  out="$(compose_migrate "$M" run --rm migrate version 2>&1)"; grep -qP '^current_version\t22$' <<<"$out" || { ok=0; log "production refusals changed the version"; }
  [[ $ok == 1 ]] && record B-MIGRATE-CLI PASS "MIGRATIONS.md commands verbatim on a fresh volume with 00001_baseline, 00002_identity, 00003_tenancy, 00004_business_profiles, 00005_organization_policy, 00006_row_level_security, 00007_stored_objects, 00008_audit_events, 00009_retention_maintenance, 00010_mcp_connections, 00011_organization_domains, 00012_session_binding_and_claim_exclusivity, 00013_organization_branding, 00014_group_onboarding, 00015_membership_invitations, 00016_organization_enrollment, 00017_group_authority, 00018_platform_signup, 00019_mcp_connection_names, 00020_labels, 00021_platform_labels and 00022_default_tenant: version, status (fresh, all twenty-two pending), up (to 22), reapply (00022), down (to 21), status+version (00022 pending), up (to 22) with asserted output; down/reapply refused with APP_ENV=production (exit 2, version unchanged)" || record B-MIGRATE-CLI FAIL "migration commands"
  if [[ "$(wc -l <"$P_LOG")" == 14 && "$(cut -f2- "$P_LOG" | LC_ALL=C sort -u | tr '\n' '|')" == "$p_expect|roles"$'\t'"${r_expect}0|roles"$'\t'"${r_expect}${MIG_RELATIONS}|" ]]; then
    record B-MIGRATE-CATALOG PASS "14 catalog lines after each of the 7 steps (incl. fresh version and fresh status before the baseline): $p_expect at every step; ${r_expect}0 at version 0 (fresh version, fresh status) and ${r_expect}${MIG_RELATIONS} at version 22 (apply, reapply, re-apply) and at version 21 (rollback, inspect; the G071 default-tenant reader is a function, not a relation): the 35 G004-G006 relations, the 3 G009 stored-object relations, the 4 G014 connection relations, the 6 G016 domain relations, the 2 G017 branding relations, the 6 G019 invitation relations, the 15 G058 enrollment relations, the 8 G062 signup relations, the 9 G020 label relations and the 1 G057 platform-label index (00014, 00017, 00019 and 00022 add no relation; no step runs at another version)"
  else record B-MIGRATE-CATALOG FAIL "catalog assertions (see $P_LOG)"; fi

  # G064 (Plan r2 §2.1, §3.1, §6): at the bound top version with the verified catalog, one read-only `migrate status` of
  # the bound migrate image feeds the source-bound pristine check, and <db> is copied once into the run-private,
  # write-closed template (B-MIGRATE-CLI and B-MIGRATE-CATALOG must be PASS in this lifecycle); B-IT-ISOLATION's controls
  # follow. The core check image is built first so the template record names it. Every independent Go invocation below
  # starts on a fresh copy of the template (it_fresh); without a template each of them refuses and FAILs.
  local mstatus="$WORK/mig-template-status.log" iti_pre=""
  compose "$M" --profile checks build core-check >"$WORK/mig-check-build-template.log" 2>&1 || log "core check image build before the template failed"
  compose_migrate "$M" run --rm migrate status >"$mstatus" 2>&1 || log "migrate status for the template failed (see $mstatus)"
  if DB_TPL_REQUIRE_ROWS="B-MIGRATE-CLI B-MIGRATE-CATALOG" DB_LC_CALLER="$project" db_template_create "$M" "$project" "$M" "$mstatus"; then
    # G064 (addendum A §A2 B-MIG): the batch opens here; its runner starts with the first eligible invocation (go_it).
    GO_IT_RUNNER_BATCH=mig GO_IT_RUNNER_NET="${project}_data"
    GO_IT_RUNNER_WANT="$(_db_rec "$WORK/db-lifecycle/${db}_tpl.rec" core_check_image 2>/dev/null | head -1)"
    it_isolation_controls "$M" "$project" "$M" "$mstatus"
  else
    iti_pre="template not created: $DB_LC_ERR"; log "G064 $iti_pre"
  fi

  # The migrate container receives only its own credential.
  ok=1
  local cname="g001-migenv-$RUN_ID" envs
  compose_migrate "$M" run --name "$cname" migrate version >/dev/null 2>&1 || ok=0
  envs="$(docker inspect -f '{{range .Config.Env}}{{println .}}{{end}}' "$cname" 2>/dev/null)"; docker rm -f "$cname" >/dev/null 2>&1 || true
  [[ "$(cut -d= -f1 <<<"$envs" | grep -vxE 'PATH|SSL_CERT_FILE|HOME' | LC_ALL=C sort | tr '\n' ' ')" == "APP_ENV DATABASE_NAME DB_HOST DB_MIGRATOR_PASSWORD DB_PORT " ]] || { ok=0; log "migrate env: $(cut -d= -f1 <<<"$envs" | tr '\n' ' ')"; }
  for line in DB_RUNTIME_PASSWORD POSTGRES_SUPERUSER_PASSWORD DB_RETENTION_PASSWORD REDIS_PASSWORD MINIO_ROOT_PASSWORD OBJECT_STORE_SECRET_KEY INTELLIGENCE_SERVICE_TOKEN; do
    grep -qF -- "$(env_get "$M" "$line")" <<<"$envs" && { ok=0; log "migrate env holds the value of $line"; }
  done
  [[ $ok == 1 ]] && record B-MIGRATE-ENV PASS "running migrate container env: exactly APP_ENV DATABASE_NAME DB_HOST DB_MIGRATOR_PASSWORD DB_PORT (plus image defaults); no runtime, superuser, retention, Redis, MinIO, object-store or intelligence secret" || record B-MIGRATE-ENV FAIL "migrate environment"

  # G004: the identity catalog (at version 10 since G014), on its own fresh copy (G064: no longer ordered before the
  # lifecycle test).
  identity_catalog "$M" "$project"

  # Integration tests (real PostgreSQL) in the core check image.
  local img="$project-core-check" itenv="$WORK/it.env" itlog="$WORK/migrate-it.log"
  # G013: the retention value joins the IT env file for TestRetention (mode 600, removed after use).
  (umask 077; printf 'DATABASE_NAME=%s\nDB_HOST=postgres\nDB_PORT=5432\nDB_MIGRATOR_PASSWORD=%s\nDB_RUNTIME_PASSWORD=%s\nDB_RETENTION_PASSWORD=%s\nDB_SUPERUSER=postgres\nDB_SUPERUSER_PASSWORD=%s\nAPP_ENV=test\n' \
    "$db" "$(env_get "$M" DB_MIGRATOR_PASSWORD)" "$(env_get "$M" DB_RUNTIME_PASSWORD)" "$(env_get "$M" DB_RETENTION_PASSWORD)" "$(env_get "$M" POSTGRES_SUPERUSER_PASSWORD)" >"$itenv")
  local migit=1
  if compose "$M" --profile checks build core-check >"$WORK/mig-check-build.log" 2>&1 &&
     it_fresh "$M" "$project" migrate-it "$itlog" &&
     go_it migrate-it "$project" "$img" "$itenv" "$itlog" ./internal/migrate/ '^TestMigrationLifecycle$'; then migit=0; fi
  local passed; passed="$(grep -c -- '--- PASS: TestMigrationLifecycle/' "$itlog" || true)"
  if [[ $migit == 0 ]] && grep -q -- '^--- PASS: TestMigrationLifecycle ' "$itlog" && [[ "$passed" == 10 ]] && ! grep -qE -- '--- (SKIP|FAIL)' "$itlog"; then
    record B-MIGRATE-IT PASS "integration tests against real PostgreSQL: 10/10 subtests pass, none skipped (fresh version/status P before the baseline, failed protection never committed, concurrent first use, up, synthetic DDL as owner with runtime/migrator denials, synthetic down, reapply, down/up, preflight refusal)"
  else tail -40 "$itlog"; record B-MIGRATE-IT FAIL "integration tests ($passed/10 subtests passed)"; fi

  # G005: the tenancy catalog, migrations and persistence (Plan r4 §6.2), then the model (B-TENANCY-IT).
  tenancy_migrations "$M" "$project" "$img" "$itenv"
  tenancy_model "$M" "$project" "$img" "$itenv"
  # G006: the BusinessProfile catalog, migrations and persistence (Plan r2 §7.2, §7.3).
  business_profile_migrations "$M" "$project" "$img" "$itenv"
  business_profile_model "$M" "$project" "$img" "$itenv"
  # G007: the Organization policy catalog, migrations and persistence (Plan r1 §6.2), then the trusted context and guarded
  # facade with the PG-001 footprint (B-POLICY-IT, §6.3, §6.7) and the authorization-to-write ordering (§6.4).
  policy_migrations "$M" "$project" "$img" "$itenv"
  policy_model "$M" "$project" "$img" "$itenv"
  policy_ordering "$M" "$project" "$img" "$itenv"
  # G008: row-level security catalog, migrations and provisioning, isolation and pool, and ordering (Plan r3 §8.3).
  RLS_CANARIES="$(new_canary) $(new_canary) $(new_canary)"
  rls_catalog "$M" "$project" "$img" "$itenv"
  rls_migration "$M" "$project" "$img" "$itenv"
  rls_it "$M" "$project" "$img" "$itenv"
  rls_ordering "$M" "$project" "$img" "$itenv"
  # G009: the private MinIO, then the stored-object catalog, migrations, isolation, ordering and faults (Plan r1 §5.5).
  if obj_minio "$M" "$project" && obj_env "$itenv" "$M"; then
    obj_catalog "$M" "$project" "$img"
    obj_migration "$M" "$project" "$img"
    obj_it "$M" "$project" "$img"
    obj_ordering "$M" "$project" "$img"
    obj_faults "$M" "$project" "$img"
    # G011: the audit catalog, migrations and persistence, then every audit producer (Plan r2 §9, §10).
    audit_catalog "$M" "$project" "$img"
    audit_migration "$M" "$project" "$img"
    audit_it "$M" "$project" "$img"
    # G013: the retention catalog and boundary, then the paired recovery and maintenance drills (Plan r3 §10, §11.4).
    retention_catalog "$M" "$project" "$img"
    retention_it "$M" "$project" "$img"
    # G014: the MCP catalog and integration (TestMCPRecovery expires every session on its own fresh copy; G064 retired
    # its position before the drill).
    mcp_catalog "$M" "$project" "$img"
    mcp_it "$M" "$project" "$img"
    # G016: the domain catalog and the domain lifecycle, resolver, race and footprint tests (Plan r4 §6.3, §11.2).
    domains_checks "$M" "$project" "$img"
    # G017: the branding catalog and the branding integration tests with the object store (Plan r2 §6.3).
    branding_checks "$M" "$project" "$img"
    # G018: the onboarding catalog and the guarded onboarding integration tests (Plan r5 §6.3 b2).
    onboarding_checks "$M" "$project" "$img"
    # G019: the membership catalog and the guarded membership, invitation and retirement integration tests (Plan r1 §6.3 b2).
    membership_checks "$M" "$project" "$img"
    # G058: the enrollment catalog with the replay cases and the guarded enrollment integration tests (Plan r3 §6.3).
    enrollment_checks "$M" "$project" "$img"
    # G062: the signup catalog and the guarded signup provisioning and continuation tests (Plan r5 §2, §5.3 b1/b2).
    signup_checks "$M" "$project" "$img"
    # G064: the one B-IT-ISOLATION row, after its T-2 repeat (signup_checks, after B-SIGNUP-CATALOG).
    it_isolation_record "$iti_pre"
    # G020: the label catalog and migration, then the guarded label and label-ordering integration tests (Plan r2 §2, §5.3 b2).
    labels_checks "$M" "$project" "$img"
    # G057: the opaque platform-address catalog and migration (00021; Plan r5 §3.5, §5.3).
    platform_address_checks "$M" "$project" "$img"
    # G064 (§C3): the B-MIG batch ends before the drill, whose `compose down -v` must not meet an attached runner; an
    # unverified runner teardown holds S-REC (same project and network).
    if go_it_runner_end; then
      res_scope_begin S-REC; recovery_drill "$M" "$project" "$img"; res_scope_end S-REC
    else
      record B-RECOVERY-DRILL FAIL "batch held: runner teardown unverified (diagnostics/it-runner-mig.log); the drill did not run"
      record B-MAINTENANCE-DRILL UNMET "the recovery drill did not run"
    fi
  else
    go_it_runner_end || log "G064 runner mig: teardown unverified at the batch end"
    local objid
    for objid in B-OBJ-CATALOG B-OBJ-MIGRATION B-OBJ-IT B-OBJ-ORDERING B-OBJ-FAULTS B-AUDIT-CATALOG B-AUDIT-MIGRATION B-AUDIT-IT \
      B-RETENTION-CATALOG B-RETENTION-IT B-MCP-CATALOG B-MCP-IT B-DOMAINS-CATALOG B-DOMAINS-IT B-BRANDING-CATALOG B-BRANDING-IT \
      B-ONBOARDING-CATALOG B-ONBOARDING-IT B-MEMBERSHIP-CATALOG B-MEMBERSHIP-IT B-ENROLLMENT-CATALOG B-ENROLLMENT-IT B-SIGNUP-CATALOG B-SIGNUP-IT \
      B-LABELS-CATALOG B-LABELS-IT B-LABELS-ORDERING B-PLATFORM-ADDRESS-CATALOG B-RECOVERY-DRILL; do
      record "$objid" FAIL "the private MinIO or the scoped credential could not be prepared"
    done
    record B-MAINTENANCE-DRILL UNMET "the recovery drill did not run"
    it_isolation_record "${iti_pre:-T-2r (after B-SIGNUP-CATALOG) did not run}"
  fi
  rm -f "$OBJ_ENV"
  rm -f "$itenv"
  save_logs "$M" mig
  go_it_runner_end || log "G064 runner mig: teardown unverified at the batch end"
  compose "$M" down -v >/dev/null 2>&1 || true
}

# it_count LOG TEST: the number of passed subtests (at any depth) of TEST, or FAIL when TEST did not pass, a subtest
# failed or anything was skipped.
it_count() {
  local log="$1" test="$2"
  if grep -q -- "^--- PASS: $test " "$log" && ! grep -qE -- '--- (SKIP|FAIL)' "$log"; then grep -c -- "--- PASS: $test/" "$log" || true; else echo FAIL; fi
}

# G005 B-TENANCY-CATALOG and B-TENANCY-MIGRATION (Plan r4 §5.1, §6.2): TestTenancyCatalog (CT-1..CT-7, the CT-6
# single-authority detector and its eleven rolled-back negative controls, runtime DDL/role/metadata denials) and
# TestTenancyMigration (MG-1..MG-5 and the ESC-1 identity-fixture TRUNCATE ... CASCADE case at version 3) against
# the disposable PostgreSQL as its actual runtime role; then MG-6: a seed phase, a stop/start of that PostgreSQL with its
# volume kept, and a verify phase.
tenancy_migrations() { # tenancy_migrations DEST PROJECT IMAGE ITENV
  local M="$1" project="$2" img="$3" itenv="$4" ok=1 log="$WORK/tenancy-catalog-it.log" mlog="$WORK/tenancy-migration-it.log" n out
  local tout="$WORK/tenancy-persist"
  mkdir -p "$tout"; chmod 777 "$tout"
  it_fresh "$M" "$project" tenancy-catalog "$log" && go_it tenancy-catalog "$project" "$img" "$itenv" "$log" ./internal/migrate/ '^TestTenancyCatalog$' || ok=0
  n="$(it_count "$log" TestTenancyCatalog)"
  if [[ $ok == 1 && "$n" == 21 ]]; then
    record B-TENANCY-CATALOG PASS "TestTenancyCatalog against real PostgreSQL as the runtime role: 21/21 subtests pass, none skipped (CT-1 exact object set and columns, CT-2 owner and trigger functions, CT-3 exact table/column ACLs and privilege matrix, CT-4 constraint definitions, CT-5 indexes, security_invoker views and triggers, CT-6 one durable authority with exactly two read-only projections, CT-6-NEG N1-N9 (11 rolled-back controls each yield exactly their codes; catalog unchanged after), CT-7 no RLS yet, runtime DDL/role/metadata denials)"
  else tail -40 "$log"; record B-TENANCY-CATALOG FAIL "tenancy catalog ($n/21 subtests)"; fi

  ok=1
  it_fresh "$M" "$project" tenancy-migration "$mlog" && go_it tenancy-migration "$project" "$img" "$itenv" "$mlog" ./internal/migrate/ '^TestTenancyMigration$' || ok=0
  n="$(it_count "$mlog" TestTenancyMigration)"
  [[ "$n" == 6 ]] || { ok=0; tail -40 "$mlog"; }
  # MG-6: seed (on its own fresh copy; the seed resets and migrates to version 3 itself, S6 correction 3), restart
  # PostgreSQL with its volume kept, verify.
  it_fresh "$M" "$project" tenancy-persist-seed "$WORK/tenancy-persist-seed.log" && go_it tenancy-persist-seed "$project" "$img" "$itenv" "$WORK/tenancy-persist-seed.log" ./internal/migrate/ '^TestTenancyPersistSeed$' \
    -e TENANCY_PERSIST_FILE=/out/persist.ids -v "$tout:/out" || { ok=0; tail -20 "$WORK/tenancy-persist-seed.log"; }
  compose "$M" stop postgres >"$WORK/tenancy-persist-restart.log" 2>&1 || ok=0
  compose "$M" up -d --wait postgres >>"$WORK/tenancy-persist-restart.log" 2>&1 || { ok=0; tail -20 "$WORK/tenancy-persist-restart.log"; }
  go_it tenancy-persist-verify "$project" "$img" "$itenv" "$WORK/tenancy-persist-verify.log" ./internal/migrate/ '^TestTenancyPersistVerify$' \
    -e TENANCY_PERSIST_FILE=/out/persist.ids -v "$tout:/out:ro" || { ok=0; tail -20 "$WORK/tenancy-persist-verify.log"; }
  for out in "$WORK/tenancy-persist-seed.log" "$WORK/tenancy-persist-verify.log"; do
    grep -qE -- '^--- PASS: TestTenancyPersist(Seed|Verify) ' "$out" && ! grep -qE -- '--- (SKIP|FAIL)' "$out" || ok=0
  done
  [[ $ok == 1 ]] && record B-TENANCY-MIGRATION PASS "TestTenancyMigration: 6/6 subtests pass, none skipped (MG-1 empty to version 3, MG-2 repeated up/status, MG-4 reapply keeps the catalog incl. functions, triggers, constraints and column ACLs, MG-3 populated post-G004 upgrade with identity rows unchanged and runtime CRUD incl. CreateGroup, MG-5 down to 2 restores the version-2 catalog with identity rows unchanged, ESC-1 fixture TRUNCATE ... CASCADE empties bindings and keeps Organizations/Groups); MG-6: rows and version 3 survive a PostgreSQL stop/start with its volume kept" || record B-TENANCY-MIGRATION FAIL "tenancy migrations ($n/6 subtests; persistence)"
}

# G005 B-TENANCY-IT (Plan r4 §5.1-§5.3): TestTenancyModel against real PostgreSQL as the runtime role, with this run's
# canaries as names (a file under the scan-excluded inputs directory, mounted read-only and removed afterwards); then
# logcheck validate and scan of the captured operation log and a planted-canary positive control.
tenancy_model() { # tenancy_model DEST PROJECT IMAGE ITENV
  local M="$1" project="$2" img="$3" itenv="$4" ok=1 log="$WORK/tenancy-model-it.log" n tin tout="$WORK/tenancy-out"
  # The tools image runs as nonroot; $WORK itself stays 0700 (as obs_open).
  tin="$(obs_inputs)/tenancy"; mkdir -p "$tin" "$tout"; chmod 755 "$tin"; chmod 777 "$tout"
  (umask 077; for _ in 1 2 3; do printf '%s\n' "$(new_canary)"; done >"$tin/canaries.txt")
  chmod 644 "$tin/canaries.txt"
  it_fresh "$M" "$project" tenancy-model "$log" && go_it tenancy-model "$project" "$img" "$itenv" "$log" ./internal/tenancy/ '^TestTenancyModel$' \
    -e TENANCY_CANARY_FILE=/in/canaries.txt -e TENANCY_LOG_FILE=/out/tenancy-core.log -v "$tin:/in:ro" -v "$tout:/out" || ok=0
  n="$(it_count "$log" TestTenancyModel)"
  [[ "$n" == 37 ]] || { ok=0; tail -60 "$log"; }
  local lc=""
  if ensure_tools && [[ -s "$tout/tenancy-core.log" ]]; then
    vtools -v "$tout:/o:ro" "$VT_IMG" logcheck validate --service core --file /o/tenancy-core.log >"$WORK/tenancy-logcheck.json" 2>&1 || { ok=0; lc+="validate "; }
    vtools -v "$tout:/o:ro" -v "$tin:/i:ro" "$VT_IMG" logcheck scan --canaries /i/canaries.txt /o/tenancy-core.log >"$WORK/tenancy-logscan.log" 2>&1 || { ok=0; lc+="scan "; }
    # Positive control: a canary planted in a scratch copy must be detected.
    mkdir -p "$tin/control"; cp "$tout/tenancy-core.log" "$tin/control/core.log"; head -1 "$tin/canaries.txt" >>"$tin/control/core.log"
    if vtools -v "$tin:/i:ro" "$VT_IMG" logcheck scan --canaries /i/canaries.txt /i/control/core.log >/dev/null 2>&1; then ok=0; lc+="control "; fi
  else ok=0; lc="no operation log or tools"; fi
  rm -rf "$tin"
  [[ $ok == 1 ]] && record B-TENANCY-IT PASS "TestTenancyModel against real PostgreSQL as the runtime role: 37/37 subtests pass, none skipped (IT-ID-1..5, IT-REF-1..4, IT-VOC-1, IT-UNIQ-1/2, IT-PROJ-1..3, IT-MULTI-1/2, IT-OWN-1..6, IT-SLUG-1..6, IT-SCOPE-1, IT-ATOM-1/2, IT-CONC-1, IT-LOG-1, G019 MEM-1 atomic bootstrap with an active owner and the inactive-owner/post-disable-wait refusals); the $(grep -c . "$tout/tenancy-core.log") operation records are fully catalogued (logcheck validate), carry none of the run's canaries (logcheck scan), and a planted canary is detected" || record B-TENANCY-IT FAIL "tenancy model ($n/37 subtests; logcheck: ${lc:-ok})"
}

# G006 B-BUSINESS-PROFILE-CATALOG and B-BUSINESS-PROFILE-MIGRATION (Plan r2 §5.1, §7.2): TestBusinessProfileCatalog
# (CT-BP-1..9: the additive version-4 catalog with every predecessor row unchanged, ownership and trigger functions, exact
# table/column ACLs and privilege matrix, constraint definitions, indexes, the security_invoker access-path view and
# triggers, the CT-6 detector with the version-4 inventory and its 18 rolled-back negative controls N1-N16, no RLS yet, no
# premature surface, runtime denials) and TestBusinessProfileMigration (MG-BP-1..5 and MG-BP-7) against the disposable
# PostgreSQL as its actual runtime role; then MG-BP-6: a seed phase, a stop/start of that PostgreSQL with its volume kept,
# and a verify phase.
business_profile_migrations() { # business_profile_migrations DEST PROJECT IMAGE ITENV
  local M="$1" project="$2" img="$3" itenv="$4" ok=1 log="$WORK/business-profile-catalog-it.log" mlog="$WORK/business-profile-migration-it.log" n out
  local bout="$WORK/business-profile-persist"
  mkdir -p "$bout"; chmod 777 "$bout"
  it_fresh "$M" "$project" business-profile-catalog "$log" && go_it business-profile-catalog "$project" "$img" "$itenv" "$log" ./internal/migrate/ '^TestBusinessProfileCatalog$' || ok=0
  n="$(it_count "$log" TestBusinessProfileCatalog)"
  if [[ $ok == 1 && "$n" == 29 ]]; then
    record B-BUSINESS-PROFILE-CATALOG PASS "TestBusinessProfileCatalog against real PostgreSQL as the runtime role: 29/29 subtests pass, none skipped (CT-BP-1 additive version-4 catalog: every predecessor row unchanged and exactly 8 relations, 9 column ACLs, 15 constraints, 5 indexes, 1 view, 2 functions and 2 triggers added; CT-BP-2 owner and trigger functions; CT-BP-3 exact table/column ACLs, no PUBLIC entry and the privilege matrix; CT-BP-4 constraint definitions, RESTRICT foreign keys, Group-unique association and nothing unique on a profile or Organization alone; CT-BP-5 indexes, the security_invoker access-path view and triggers; CT-BP-6 one durable authority with the version-4 projection inventory; CT-BP-6-NEG N1-N16 (18 rolled-back controls each yield exactly their codes; catalog unchanged after); CT-BP-7 no RLS yet; CT-BP-8 no premature surface; CT-BP-9 runtime DDL, write, role and metadata denials)"
  else tail -40 "$log"; record B-BUSINESS-PROFILE-CATALOG FAIL "business profile catalog ($n/29 subtests)"; fi

  ok=1
  it_fresh "$M" "$project" business-profile-migration "$mlog" && go_it business-profile-migration "$project" "$img" "$itenv" "$mlog" ./internal/migrate/ '^TestBusinessProfileMigration$' || ok=0
  n="$(it_count "$mlog" TestBusinessProfileMigration)"
  [[ "$n" == 6 ]] || { ok=0; tail -40 "$mlog"; }
  # MG-BP-6: seed (on its own fresh copy; the seed resets and migrates to version 4 itself, S6 correction 3), restart
  # PostgreSQL with its volume kept, verify.
  it_fresh "$M" "$project" business-profile-persist-seed "$WORK/business-profile-persist-seed.log" && go_it business-profile-persist-seed "$project" "$img" "$itenv" "$WORK/business-profile-persist-seed.log" ./internal/migrate/ '^TestBusinessProfilePersistSeed$' \
    -e BUSINESS_PROFILE_PERSIST_FILE=/out/persist.ids -v "$bout:/out" || { ok=0; tail -20 "$WORK/business-profile-persist-seed.log"; }
  compose "$M" stop postgres >"$WORK/business-profile-persist-restart.log" 2>&1 || ok=0
  compose "$M" up -d --wait postgres >>"$WORK/business-profile-persist-restart.log" 2>&1 || { ok=0; tail -20 "$WORK/business-profile-persist-restart.log"; }
  go_it business-profile-persist-verify "$project" "$img" "$itenv" "$WORK/business-profile-persist-verify.log" ./internal/migrate/ '^TestBusinessProfilePersistVerify$' \
    -e BUSINESS_PROFILE_PERSIST_FILE=/out/persist.ids -v "$bout:/out:ro" || { ok=0; tail -20 "$WORK/business-profile-persist-verify.log"; }
  for out in "$WORK/business-profile-persist-seed.log" "$WORK/business-profile-persist-verify.log"; do
    grep -qE -- '^--- PASS: TestBusinessProfilePersist(Seed|Verify) ' "$out" && ! grep -qE -- '--- (SKIP|FAIL)' "$out" || ok=0
  done
  rm -f "$bout/persist.ids"
  [[ $ok == 1 ]] && record B-BUSINESS-PROFILE-MIGRATION PASS "TestBusinessProfileMigration: 6/6 subtests pass, none skipped (MG-BP-1 empty to version 4, MG-BP-2 repeated up/status, MG-BP-4 reapply keeps the catalog, MG-BP-3 populated version-3 upgrade applies only 00004 with identity and every Organization, Group and binding unchanged plus a runtime-SQL profile, link and access-path smoke, MG-BP-5 down to 3 with G006 rows present restores the version-3 catalog with identity and tenancy rows unchanged, MG-BP-7 identity fixture TRUNCATE ... CASCADE empties bindings and keeps Organizations, Groups, profiles and links); MG-BP-6: the profile, its link, the access path and version 4 survive a PostgreSQL stop/start with its volume kept" || record B-BUSINESS-PROFILE-MIGRATION FAIL "business profile migrations ($n/6 subtests; persistence)"
}

# G006 B-BUSINESS-PROFILE-IT (Plan r2 §5.1-§5.3): TestBusinessProfileModel against real PostgreSQL as the runtime role,
# with this run's canaries in every context field (a file under the scan-excluded inputs directory, mounted read-only
# and removed afterwards); then logcheck validate and scan of the captured operation log and a planted-canary control.
business_profile_model() { # business_profile_model DEST PROJECT IMAGE ITENV
  local M="$1" project="$2" img="$3" itenv="$4" ok=1 log="$WORK/business-profile-model-it.log" n tin tout="$WORK/business-profile-out"
  tin="$(obs_inputs)/business-profile"; mkdir -p "$tin" "$tout"; chmod 755 "$tin"; chmod 777 "$tout"
  (umask 077; for _ in 1 2 3; do printf '%s\n' "$(new_canary)"; done >"$tin/canaries.txt")
  chmod 644 "$tin/canaries.txt"
  it_fresh "$M" "$project" business-profile-model "$log" && go_it business-profile-model "$project" "$img" "$itenv" "$log" ./internal/tenancy/ '^TestBusinessProfileModel$' \
    -e TENANCY_CANARY_FILE=/in/canaries.txt -e TENANCY_LOG_FILE=/out/business-profile-core.log -v "$tin:/in:ro" -v "$tout:/out" || ok=0
  n="$(it_count "$log" TestBusinessProfileModel)"
  [[ "$n" == 35 ]] || { ok=0; tail -60 "$log"; }
  local lc=""
  if ensure_tools && [[ -s "$tout/business-profile-core.log" ]]; then
    vtools -v "$tout:/o:ro" "$VT_IMG" logcheck validate --service core --file /o/business-profile-core.log >"$WORK/business-profile-logcheck.json" 2>&1 || { ok=0; lc+="validate "; }
    vtools -v "$tout:/o:ro" -v "$tin:/i:ro" "$VT_IMG" logcheck scan --canaries /i/canaries.txt /o/business-profile-core.log >"$WORK/business-profile-logscan.log" 2>&1 || { ok=0; lc+="scan "; }
    # Positive control: a canary planted in a scratch copy must be detected.
    mkdir -p "$tin/control"; cp "$tout/business-profile-core.log" "$tin/control/core.log"; head -1 "$tin/canaries.txt" >>"$tin/control/core.log"
    if vtools -v "$tin:/i:ro" "$VT_IMG" logcheck scan --canaries /i/canaries.txt /i/control/core.log >/dev/null 2>&1; then ok=0; lc+="control "; fi
  else ok=0; lc="no operation log or tools"; fi
  rm -rf "$tin"
  [[ $ok == 1 ]] && record B-BUSINESS-PROFILE-IT PASS "TestBusinessProfileModel against real PostgreSQL as the runtime role: 35/35 subtests pass, none skipped (IT-BP-ACC-1..7, IT-BP-ID-1..5, IT-BP-CTX-1..4, IT-BP-LINK-1..5, IT-BP-XORG-1..4, IT-BP-UNLINK-1/2, IT-BP-ATOM-1/2, IT-BP-REF-1/2, IT-BP-INDEP-1/2, IT-BP-LOG-1, runtime identity); the $(grep -c . "$tout/business-profile-core.log") operation records are fully catalogued (logcheck validate), carry none of the run's canaries (logcheck scan), and a planted canary is detected" || record B-BUSINESS-PROFILE-IT FAIL "business profile model ($n/35 subtests; logcheck: ${lc:-ok})"
}

# G007 B-POLICY-CATALOG and B-POLICY-MIGRATION (Plan r1 §6.2): TestPolicyCatalog (CT-POL-1 the version-5 delta is
# exactly the policy column and its runtime column ACL, with no table rewrite; CT-POL-2 the exact runtime privilege
# matrix on organizations, no INSERT on the flag and no PUBLIC entry; CT-POL-3 rolled-back malformed assignments, the
# runtime row-lock privileges the ordering contract relies on (direct scoped role_bindings locks; 42501 through the
# SELECT-only projections, Codex review 432d5f17 §1) and unchanged projection write/ACL restrictions; CT-POL-4 no RLS, new object or authority; CT-POL-5
# runtime DDL, role and metadata denials) and TestPolicyMigration (MG-POL-1..5) against the disposable PostgreSQL as its
# actual runtime role; then MG-POL-6: a seed phase, a stop/start of that PostgreSQL with its volume kept, and a verify
# phase.
policy_migrations() { # policy_migrations DEST PROJECT IMAGE ITENV
  local M="$1" project="$2" img="$3" itenv="$4" ok=1 log="$WORK/policy-catalog-it.log" mlog="$WORK/policy-migration-it.log" n out
  local pout="$WORK/policy-persist"
  mkdir -p "$pout"; chmod 777 "$pout"
  it_fresh "$M" "$project" policy-catalog "$log" && go_it policy-catalog "$project" "$img" "$itenv" "$log" ./internal/migrate/ '^TestPolicyCatalog$' || ok=0
  n="$(it_count "$log" TestPolicyCatalog)"
  if [[ $ok == 1 && "$n" == 6 ]]; then
    record B-POLICY-CATALOG PASS "TestPolicyCatalog against real PostgreSQL as the runtime role: 6/6 subtests pass, none skipped (CT-POL-1 the version-5 delta is exactly the group_admins_edit_business_profiles column (boolean NOT NULL DEFAULT true, a fast default without a table rewrite) and its runtime column ACL, every other catalog row unchanged; CT-POL-2 runtime holds exactly SELECT on organizations plus INSERT/UPDATE on name and UPDATE on the flag, no INSERT on the flag, no PUBLIC entry, retention and migrator nothing, and a new Organization reads true; CT-POL-3 runtime and owner NULL/text assignments give 23502/22P02 with the row unchanged, FOR SHARE on group_business_profiles and through both membership projections gives 42501, the direct scoped role_bindings locks and the other row locks of the ordering contract succeed, and runtime writes through the projections are refused with the projection and role_bindings ACLs unchanged; CT-POL-4 no RLS, policy or SECURITY DEFINER function and the version-4 authority inventory; CT-POL-5 runtime DDL, role and metadata denials)"
  else tail -40 "$log"; record B-POLICY-CATALOG FAIL "policy catalog ($n/6 subtests)"; fi

  ok=1
  it_fresh "$M" "$project" policy-migration "$mlog" && go_it policy-migration "$project" "$img" "$itenv" "$mlog" ./internal/migrate/ '^TestPolicyMigration$' || ok=0
  n="$(it_count "$mlog" TestPolicyMigration)"
  [[ "$n" == 5 ]] || { ok=0; tail -40 "$mlog"; }
  # MG-POL-6: seed (on its own fresh copy; the seed resets and migrates to version 5 itself, S6 correction 3), restart
  # PostgreSQL with its volume kept, verify.
  it_fresh "$M" "$project" policy-persist-seed "$WORK/policy-persist-seed.log" && go_it policy-persist-seed "$project" "$img" "$itenv" "$WORK/policy-persist-seed.log" ./internal/migrate/ '^TestPolicyPersistSeed$' \
    -e POLICY_PERSIST_FILE=/out/persist.ids -v "$pout:/out" || { ok=0; tail -20 "$WORK/policy-persist-seed.log"; }
  compose "$M" stop postgres >"$WORK/policy-persist-restart.log" 2>&1 || ok=0
  compose "$M" up -d --wait postgres >>"$WORK/policy-persist-restart.log" 2>&1 || { ok=0; tail -20 "$WORK/policy-persist-restart.log"; }
  go_it policy-persist-verify "$project" "$img" "$itenv" "$WORK/policy-persist-verify.log" ./internal/migrate/ '^TestPolicyPersistVerify$' \
    -e POLICY_PERSIST_FILE=/out/persist.ids -v "$pout:/out:ro" || { ok=0; tail -20 "$WORK/policy-persist-verify.log"; }
  for out in "$WORK/policy-persist-seed.log" "$WORK/policy-persist-verify.log"; do
    grep -qE -- '^--- PASS: TestPolicyPersist(Seed|Verify) ' "$out" && ! grep -qE -- '--- (SKIP|FAIL)' "$out" || ok=0
  done
  rm -f "$pout/persist.ids"
  [[ $ok == 1 ]] && record B-POLICY-MIGRATION PASS "TestPolicyMigration: 5/5 subtests pass, none skipped (MG-POL-1 empty to version 5, MG-POL-2 repeated up/status, MG-POL-3 populated version-4 upgrade applies only 00005: existing Organizations read true, identity and every Organization, Group, binding, profile and link row unchanged, the catalog is version 4 plus the column, and a new Organization reads true; MG-POL-4 reapply keeps the catalog; MG-POL-5 down to 4 with a false flag restores the version-4 catalog with every predecessor row unchanged, then up reads true); MG-POL-6: a false and a true flag and version 5 survive a PostgreSQL stop/start with its volume kept" || record B-POLICY-MIGRATION FAIL "policy migrations ($n/5 subtests; persistence)"
}

# policy_logcheck NAME OUTDIR INDIR LOG: logcheck validate and scan of a captured access.decision log, and the
# planted-canary positive control; prints the failed parts (nothing when all hold).
policy_logcheck() {
  local name="$1" tout="$2" tin="$3" logf="$4" lc=""
  if ensure_tools && [[ -s "$tout/$logf" ]]; then
    vtools -v "$tout:/o:ro" "$VT_IMG" logcheck validate --service core --file "/o/$logf" >"$WORK/$name-logcheck.json" 2>&1 || lc+="validate "
    vtools -v "$tout:/o:ro" -v "$tin:/i:ro" "$VT_IMG" logcheck scan --canaries /i/canaries.txt "/o/$logf" >"$WORK/$name-logscan.log" 2>&1 || lc+="scan "
    mkdir -p "$tin/control"; cp "$tout/$logf" "$tin/control/core.log"; head -1 "$tin/canaries.txt" >>"$tin/control/core.log"
    if vtools -v "$tin:/i:ro" "$VT_IMG" logcheck scan --canaries /i/canaries.txt /i/control/core.log >/dev/null 2>&1; then lc+="control "; fi
  else lc="no access log or tools"; fi
  printf '%s' "$lc"
}

# G007 B-POLICY-IT (Plan r1 §6.3, §6.7): TestAccessPolicy and TestAccessFootprint against real PostgreSQL as the runtime
# role, with this run's canaries in names, display names and profile context (a file under the scan-excluded inputs
# directory, mounted read-only and removed afterwards); then logcheck validate and scan of the access.decision log and a
# planted-canary positive control. The E-PG-1/E-PG-2 summary lines are kept as policy-epg.jsonl in the report.
policy_model() { # policy_model DEST PROJECT IMAGE ITENV
  local M="$1" project="$2" img="$3" itenv="$4" ok=1 log="$WORK/policy-model-it.log" n f tin tout="$WORK/policy-out" lc triggers newtriggers
  tin="$(obs_inputs)/policy"; mkdir -p "$tin" "$tout"; chmod 755 "$tin"; chmod 777 "$tout"
  (umask 077; for _ in 1 2 3; do printf '%s\n' "$(new_canary)"; done >"$tin/canaries.txt")
  chmod 644 "$tin/canaries.txt"
  it_fresh "$M" "$project" policy-model "$log" && GO_IT_TIMEOUT=30m go_it policy-model "$project" "$img" "$itenv" "$log" ./internal/tenancy/access/ '^(TestAccessPolicy|TestAccessFootprint)$' \
    -e TENANCY_CANARY_FILE=/in/canaries.txt -e TENANCY_LOG_FILE=/out/policy-core.log -v "$tin:/in:ro" -v "$tout:/out" || ok=0
  n="$(it_count "$log" TestAccessPolicy)"; f="$(it_count "$log" TestAccessFootprint)"
  [[ "$n" == 39 && "$f" == 4 ]] || { ok=0; tail -80 "$log"; }
  grep -E '^E-PG-(1|1-TRIGGER|2|MEM) ' "$log" >"$WORK/policy-epg.jsonl" || true
  if [[ -n "$REPORT_DIR" ]]; then cp "$WORK/policy-epg.jsonl" "$REPORT_DIR/" 2>/dev/null || true; fi
  # Codex review 432d5f17 §2: the small-fixture sequential scans of statements 6b/7 are an accepted PG-001 observation
  # ("known"); any other trigger is new and returns to Codex.
  triggers="$(grep -c '^E-PG-1-TRIGGER ' "$log" || true)"
  newtriggers="$(grep '^E-PG-1-TRIGGER ' "$log" | grep -c '"known":false' || true)"
  [[ "$newtriggers" == 0 ]] || log "E-PG-1: $newtriggers new escalation trigger line(s): return to Codex (policy-epg.jsonl)"
  lc="$(policy_logcheck policy "$tout" "$tin" policy-core.log)"
  [[ -z "$lc" ]] || ok=0
  rm -rf "$tin"
  [[ $ok == 1 ]] && record B-POLICY-IT PASS "TestAccessPolicy against real PostgreSQL as the runtime role, sessions from the real G004 login: 39/39 subtests pass, none skipped (IT-CTX-1..11 incl. cross-request, cross-subject, same-request and ended-scope binding with zero BeginTx and a browser scope resolving only its session's Organization (G016 A2), IT-MATRIX-1 (the oracle's 418 cells on DB-loaded facts), IT-PROFILE-1..5, IT-FLAG-1..5, IT-FACADE-1..6 incl. definitive aborts and the lost COMMIT reply, IT-REVOKE-1..8, IT-MFA-1, IT-LOG-1); TestAccessFootprint 4/4 (E-PG-1 plan summaries recorded, $triggers escalation trigger lines of which $newtriggers new; E-PG-2 locked-row inventory and G018 onboarding writes asserted; G019 E-PG-MEM membership, ownership and invitation operations, incl. role change, demotion, leave, revoke, retirement eligibility and Group removal on distinct prepared subjects, recorded with their claim-call and read-lock assertions); the $(grep -c . "$tout/policy-core.log") access.decision records are fully catalogued (logcheck validate), carry none of the run's canaries (logcheck scan), and a planted canary is detected" || record B-POLICY-IT FAIL "policy model ($n/39 policy, $f/4 footprint subtests; logcheck: ${lc:-ok})"
}

# G007 B-POLICY-ORDERING (Plan r1 §6.4, §6.7): TestAccessOrdering against real PostgreSQL as the runtime role
# (deterministic hook, tracer and lock-wait barriers; no sleeps), then logcheck as B-POLICY-IT. The ORDERING lines
# (interleaving and outcome of every case, the O-15 victim and the E-PG-3 characterization) are kept as
# policy-ordering.jsonl in the report.
policy_ordering() { # policy_ordering DEST PROJECT IMAGE ITENV
  local M="$1" project="$2" img="$3" itenv="$4" ok=1 log="$WORK/policy-ordering-it.log" n tin tout="$WORK/policy-ordering-out" lc
  tin="$(obs_inputs)/policy-ordering"; mkdir -p "$tin" "$tout"; chmod 755 "$tin"; chmod 777 "$tout"
  (umask 077; for _ in 1 2 3; do printf '%s\n' "$(new_canary)"; done >"$tin/canaries.txt")
  chmod 644 "$tin/canaries.txt"
  it_fresh "$M" "$project" policy-ordering "$log" && GO_IT_TIMEOUT=30m go_it policy-ordering "$project" "$img" "$itenv" "$log" ./internal/tenancy/access/ '^TestAccessOrdering$' \
    -e TENANCY_CANARY_FILE=/in/canaries.txt -e TENANCY_LOG_FILE=/out/policy-ordering-core.log -v "$tin:/in:ro" -v "$tout:/out" || ok=0
  n="$(it_count "$log" TestAccessOrdering)"
  [[ "$n" == 40 ]] || { ok=0; tail -80 "$log"; }
  grep -E '^ORDERING ' "$log" >"$WORK/policy-ordering.jsonl" || true
  if [[ -n "$REPORT_DIR" ]]; then cp "$WORK/policy-ordering.jsonl" "$REPORT_DIR/" 2>/dev/null || true; fi
  lc="$(policy_logcheck policy-ordering "$tout" "$tin" policy-ordering-core.log)"
  [[ -z "$lc" ]] || ok=0
  rm -rf "$tin"
  [[ $ok == 1 ]] && record B-POLICY-ORDERING PASS "TestAccessOrdering against real PostgreSQL as the runtime role: 40/40 subtests pass, none skipped (G060 O-MEM-4a Organization removal of a Group admin versus its promotion to Group owner, O-MEM-4b versus its own Group claim, O-MEM-4c a raw Organization departure versus a Group claim with one bounded class-40 abort and nothing partial; G019 O-MEM-1 concurrent last-owner departures, O-MEM-2a..c the transfer against the recipient's disable and removal and the actor's logout-all, O-MEM-3a..d shared order, facade-versus-direct and mutual demotion with class-40 victims answered unavailable and zero partial writes, promotion versus DisableUser, O-MEM-DUP duplicate claims, O-MEM-RM a Group acceptance versus the recipient's Organization removal in both orders without a dormant Group binding; G018 O-ONB-1..3, O-ONB-3s, O-ONB-3p, O-ONB-D, O-ONB-E and O-ONB-T; O-1..O-14: every conflicting accepted writer in orders A guard-first, B writer-first and C writer-in-flight, each committing the guarded write before the change or denying with the pinned code and an unchanged target; O-15: exactly one 40P01 victim with the safe outcome; E-PG-3a/b lock-queue characterization with safety asserted); the $(grep -c . "$tout/policy-ordering-core.log") access.decision records are fully catalogued, carry none of the run's canaries, and a planted canary is detected" || record B-POLICY-ORDERING FAIL "policy ordering ($n/40 subtests; logcheck: ${lc:-ok})"
}

# ===================================================================================
# G008 B-RLS-CATALOG, B-RLS-MIGRATION, B-RLS-IT and B-RLS-ORDERING (Plan r3 §8.3, §9). Test containers receive this run's
# canaries in an environment variable and return their access.decision records on stdout (prefix RLS-ACCESSLOG); the
# log checks read one fixed host directory, emptied before and after use, so no new per-run host directory is
# bind-mounted (AGENTS.md; docs/foundation/OPERATIONS.md, "Docker Desktop bind-mount leak").
RLS_FIXED="${TMPDIR:-/tmp}/platform-validation-g008-rls"
RLS_CANARIES=""
rls_fixed_reset() { rm -rf -- "$RLS_FIXED"; mkdir -p "$RLS_FIXED"; chmod 755 "$RLS_FIXED"; }

# rls_go NAME PROJECT IMAGE ITENV LOG PACKAGE RUN [docker args...]: one go test invocation in the core check image on the
# data network. G064 (D-6): a compatibility alias of go_it with this run's canaries and the RLS access log (30m).
rls_go() {
  local name="$1" project="$2" img="$3" itenv="$4" log="$5" pkg="$6" run="$7"; shift 7
  GO_IT_TIMEOUT="${GO_IT_TIMEOUT:-30m}" go_it "$name" "$project" "$img" "$itenv" "$log" "$pkg" "$run" \
    -e RLS_CANARIES="$RLS_CANARIES" -e TENANCY_CANARY_FILE=/tmp/rls-canaries.txt -e TENANCY_LOG_FILE=/tmp/rls-access.log "$@"
}

# G064 (Plan r2 §5.1): go_it NAME PROJECT IMAGE ITENV LOG PACKAGE RUN [docker args...]: one Go test invocation in IMAGE on
# PROJECT's data network (GO_IT_NETWORK overrides it; ITENV may be empty; GO_IT_TAG, default integration; GO_IT_TIMEOUT,
# default 10m = go test's own default). In one container: the expected top-level tests of the bound source (`go test
# -json -list RUN PACKAGE`, keyed by each event's Package), then the run (`go test -json -run RUN PACKAGE`), its exit
# status captured before any output handling, the RLS access log (as before), the expectation block, and that status as
# the container's. The stream is saved as LOG.events; gotestreport (validation tools image, no network, stdin/stdout:
# no bind mount) writes LOG = the readable -v text the existing parsers read, the rows and the JUnit fragment, which are
# exported at once (go_tests_export). Returns the run's non-zero status unchanged, 1 when the stream fails a
# completeness rule (missing terminal, omitted/undeclared/duplicate/contradictory/unfinished/skipped test, missing
# expectation...) or cannot be converted, and 0 only for a complete passing invocation. GO_IT_NO_EXPORT=1 defers the
# export to the caller (go_it_publish; mcp_e2e first moves its observed bearer values into the secret registry).
GO_IT_TREE="" GO_IT_HEAD="" GO_IT_TREEFLAG=""
declare -A GO_IT_CHECK=()
# go_it_source DIR: the provenance of the copied tree the suite's images are built from (once per suite).
go_it_source() {
  GO_IT_TREE="$(impl_tree_hash "$1" 2>/dev/null)" || true
  [[ "$GO_IT_TREE" =~ ^[0-9a-f]{64}$ ]] || GO_IT_TREE="unavailable"
  GO_IT_HEAD="$(git -C "$REPO" rev-parse HEAD 2>/dev/null || echo unavailable)"
  if [[ -z "$(git -C "$REPO" status --porcelain -- . ":(top,exclude).agents/skills/restruct-sdlc" ":(top,exclude).claude/skills/restruct-sdlc" 2>/dev/null)" ]]; then GO_IT_TREEFLAG="head"; else GO_IT_TREEFLAG="dirty"; fi
}
# Declared invocation inventories (Plan r2 §4, §5.3): NAME|CHECK. obj-sign-host runs only with a host probe.
# shellcheck disable=SC2034 # read through go_it_declare's nameref
GO_IT_migrations=(migrate-it\|B-MIGRATE-IT identity-catalog\|B-IDENTITY-CATALOG tenancy-catalog\|B-TENANCY-CATALOG
  tenancy-migration\|B-TENANCY-MIGRATION tenancy-persist-seed\|B-TENANCY-MIGRATION tenancy-persist-verify\|B-TENANCY-MIGRATION
  tenancy-model\|B-TENANCY-IT business-profile-catalog\|B-BUSINESS-PROFILE-CATALOG business-profile-migration\|B-BUSINESS-PROFILE-MIGRATION
  business-profile-persist-seed\|B-BUSINESS-PROFILE-MIGRATION business-profile-persist-verify\|B-BUSINESS-PROFILE-MIGRATION
  business-profile-model\|B-BUSINESS-PROFILE-IT policy-catalog\|B-POLICY-CATALOG policy-migration\|B-POLICY-MIGRATION
  policy-persist-seed\|B-POLICY-MIGRATION policy-persist-verify\|B-POLICY-MIGRATION policy-model\|B-POLICY-IT
  policy-ordering\|B-POLICY-ORDERING rls-catalog\|B-RLS-CATALOG rls-migration\|B-RLS-MIGRATION rls-persist-seed\|B-RLS-MIGRATION
  rls-persist-verify\|B-RLS-MIGRATION rls-provisioned-su\|B-RLS-MIGRATION rls-provisioned-adm\|B-RLS-MIGRATION
  rls-provisioned-v5\|B-RLS-MIGRATION rls-it\|B-RLS-IT rls-identity-it\|B-RLS-IT rls-ordering\|B-RLS-ORDERING
  obj-catalog\|B-OBJ-CATALOG obj-migration\|B-OBJ-MIGRATION obj-persist-seed\|B-OBJ-MIGRATION obj-persist-verify\|B-OBJ-MIGRATION
  obj-it\|B-OBJ-IT obj-it-attack\|B-OBJ-IT obj-sign-host\|B-OBJ-IT obj-ordering\|B-OBJ-ORDERING obj-faults\|B-OBJ-FAULTS
  audit-catalog\|B-AUDIT-CATALOG audit-migration\|B-AUDIT-MIGRATION audit-persist-seed\|B-AUDIT-MIGRATION
  audit-persist-verify\|B-AUDIT-MIGRATION audit-it\|B-AUDIT-IT retention-catalog\|B-RETENTION-CATALOG retention-it\|B-RETENTION-IT
  mcp-catalog\|B-MCP-CATALOG mcp-it\|B-MCP-IT mcp-facade-it\|B-MCP-IT domains-catalog\|B-DOMAINS-CATALOG domains-it\|B-DOMAINS-IT
  branding-catalog\|B-BRANDING-CATALOG branding-it\|B-BRANDING-IT onboarding-catalog\|B-ONBOARDING-CATALOG onboarding-it\|B-ONBOARDING-IT
  membership-catalog\|B-MEMBERSHIP-CATALOG membership-it\|B-MEMBERSHIP-IT enrollment-catalog\|B-ENROLLMENT-CATALOG
  enrollment-it\|B-ENROLLMENT-IT signup-catalog\|B-SIGNUP-CATALOG signup-it\|B-SIGNUP-IT labels-catalog\|B-LABELS-CATALOG
  labels-it\|B-LABELS-IT labels-ordering\|B-LABELS-ORDERING platform-address-catalog\|B-PLATFORM-ADDRESS-CATALOG
  rd-seed\|B-RECOVERY-DRILL
  rd-post\|B-RECOVERY-DRILL rd-verify\|B-RECOVERY-DRILL rd-repair1\|B-MAINTENANCE-DRILL rd-repair2\|B-MAINTENANCE-DRILL)
# shellcheck disable=SC2034 # read through go_it_declare's nameref
GO_IT_ops=(identity-it\|B-IDENTITY-IT mcp-e2e\|B-MCP-E2E)
# go_it_declare SUITE: write the suite's declared inventory (before its first invocation).
go_it_declare() {
  local -n inv="GO_IT_$1"
  local e names=()
  for e in "${inv[@]}"; do
    GO_IT_CHECK[${e%%|*}]="${e#*|}"
    [[ "${e%%|*}" == obj-sign-host && "$PROBE_MODE" == none ]] && continue
    names+=("${e%%|*}")
  done
  go_tests_declare "$1" "${names[@]}"
}
# The in-container script: sh -c "$GO_IT_SH" sh RUN PACKAGE TAG TIMEOUT.
# shellcheck disable=SC2016 # expanded by the container's sh
GO_IT_SH='printf "%s\n" $RLS_CANARIES >/tmp/rls-canaries.txt
      go test -tags "$3" -json -list "$1" "$2" >/tmp/go-it-list.json; lrc=$?
      go test -tags "$3" -count=1 -timeout "$4" -json -run "$1" "$2"; rc=$?
      if [ -f /tmp/rls-access.log ]; then sed "s/^/RLS-ACCESSLOG /" /tmp/rls-access.log; fi
      sed "s/^/GO-IT-LIST /" /tmp/go-it-list.json; echo "GO-IT-LIST-EXIT $lrc"; exit $rc'
# G064 batch runner (Plan addendum A §A2/§A3; correction 1 §C3/§C4). suite_migrations opens the B-MIG batch after
# db_template_create (GO_IT_RUNNER_BATCH=mig, the runner network and the template record's core_check_image) and closes
# it at every named batch end (go_it_runner_end). GO_IT_RUNNER=off forces the per-invocation container path.
GO_IT_RUNNER_BATCH="" GO_IT_RUNNER_NET="" GO_IT_RUNNER_WANT=""
# go_it_route IMAGE_ID NET [docker args...]: 0 when the invocation may run in the batch runner: a batch is open, the
# network is the runner's, the image ID is the runner's (template record) and every docker argument is -e K=V or
# --env-file F (a mount, --tmpfs or any other option keeps the container path). Else GO_IT_ROUTE_WHY names the rule.
go_it_route() {
  local iid="$1" net="$2"; shift 2
  GO_IT_ROUTE_WHY=""
  if [[ "${GO_IT_RUNNER:-}" == off ]]; then GO_IT_ROUTE_WHY=off
  elif [[ -z "$GO_IT_RUNNER_BATCH" ]]; then GO_IT_ROUTE_WHY=no-batch
  elif [[ "$net" != "$GO_IT_RUNNER_NET" ]]; then GO_IT_ROUTE_WHY=network
  elif [[ ! "$GO_IT_RUNNER_WANT" =~ ^sha256:[0-9a-f]+$ || "$iid" != "$GO_IT_RUNNER_WANT" ]]; then GO_IT_ROUTE_WHY=image
  fi
  while [[ -z "$GO_IT_ROUTE_WHY" ]] && (($#)); do
    if [[ $# -ge 2 && ( ( "$1" == -e && "$2" =~ ^[A-Za-z_][A-Za-z0-9_]*= ) || ( "$1" == --env-file && -n "$2" ) ) ]]; then shift 2
    else GO_IT_ROUTE_WHY=docker-args; fi
  done
  [[ -z "$GO_IT_ROUTE_WHY" ]]
}
go_it_flat() { printf '%s' "$1" | tr '\t\r\n' '   ' | head -c 300; } # one provenance value: no tab or newline
# go_it_runner_end: a named B-MIG batch end (§C3): the runner is stopped (verified teardown) and the batch closed, so no
# later invocation (the drill's rd_go included) reaches it. Returns 1 when the runner's teardown is unverified: S-REC is
# then held, because it resets the same project and network.
go_it_runner_end() {
  local b="$GO_IT_RUNNER_BATCH" rc=0
  GO_IT_RUNNER_BATCH="" GO_IT_RUNNER_NET="" GO_IT_RUNNER_WANT=""
  [[ -n "$b" ]] || return 0
  it_runner_stop "$b" || rc=1
  if grep -qs 'teardown unverified' "$WORK/it-runner/$b.held"; then rc=1; fi
  return "$rc"
}
go_it() {
  local name="$1" project="$2" img="$3" itenv="$4" log="$5" pkg="$6" run="$7"; shift 7
  local tag="${GO_IT_TAG:-integration}" timeout="${GO_IT_TIMEOUT:-10m}" net="${GO_IT_NETWORK:-${project}_data}" rc=0
  local -a go_envf=()
  [[ -z "$itenv" ]] || go_envf=(--env-file "$itenv")
  go_it_begin "$name" "$img" "$log" "$pkg" "$run" "$tag" "$timeout" || return 1
  # G064 (Plan addendum A §A2/§A3; correction 1 §C3/§C4): transport routing. Only the transport changes; the reporter
  # path (go_it_begin -> GO_IT_SH -> go_it_convert, rows/log per invocation) is the same on both paths.
  local pf="$WORK/go-tests/$name.prov" route=container why=""
  local -a GO_IT_XARGS=("${go_envf[@]}" "$@")
  if go_it_route "$(sed -n 's/^image_id=//p' "$pf")" "$net" "${GO_IT_XARGS[@]}"; then
    if it_runner_ready "$GO_IT_RUNNER_BATCH" "$img" "$net" "$GO_IT_RUNNER_WANT"; then route='exec'
    elif ((ITR_FALLBACK)); then why="$ITR_WHY"
    else route=refused why="$ITR_WHY"; fi
  else why="$GO_IT_ROUTE_WHY"; fi
  case "$route" in
    exec)
      printf 'runner=%s\n' "${ITR_ID:0:12}" >>"$pf"
      it_runner_exec "$GO_IT_RUNNER_BATCH" "$log.events" sh -c "$GO_IT_SH" sh "$run" "$pkg" "$tag" "$timeout" || true
      rc="$ITR_RC"
      if [[ -n "$ITR_FINDING" ]]; then
        printf 'runner_finding=%s\n' "$(go_it_flat "$ITR_FINDING")" >>"$pf"
        log "G064 go_it $name: runner ${ITR_ID:0:12} $ITR_FINDING (invocation fails, rc=$rc)" >&2
      fi ;;
    refused)
      printf 'runner=none\nrunner_finding=%s\n' "$(go_it_flat "$why")" >>"$pf"
      printf 'G064 go_it: %s; %s did not run (no per-test evidence)\n' "$why" "$name" >"$log.events"
      rc=125 ;;
    *)
      printf 'runner=container\nrunner_reason=%s\n' "$(go_it_flat "$why")" >>"$pf"
      docker_run_owned "$name" --rm --label g001.run=$RUN_ID --network "$net" "${go_envf[@]}" "$@" "$img" \
        sh -c "$GO_IT_SH" sh "$run" "$pkg" "$tag" "$timeout" >"$log.events" 2>&1 || rc=$? ;;
  esac
  # G064 C2: each go_it boundary of an open batch is a B-MIG enforcement point (a crossing latches: later invocations FAIL
  # "batch held"; this invocation's status is unchanged).
  if [[ -n "$GO_IT_RUNNER_BATCH" ]] && ! res_recount "go_it $name"; then printf 'res_finding=%s\n' "$(go_it_flat "$RES_WHY")" >>"$pf"; fi
  local conv=0
  go_it_convert "$name" "$log" "$rc" || conv=1
  ((rc == 0)) || return "$rc"
  return "$conv"
}
# go_it_begin NAME IMAGE LOG PACKAGE RUN TAG TIMEOUT: register the invocation (started.tsv, provenance) before it runs;
# returns 1 (LOG names the cause, and a FAIL row is exported) when the name is invalid or repeated or the validation
# tools image is unavailable: the invocation must not run without its per-test evidence.
go_it_begin() {
  local name="$1" img="$2" log="$3" pkg="$4" run="$5" tag="$6" timeout="$7" d="$WORK/go-tests"
  mkdir -p "$d"
  if [[ ! "$name" =~ ^[A-Za-z0-9_.-]{1,120}$ || -e "$d/$name.prov" ]]; then
    printf 'G064 go_it: invalid or repeated invocation name %s\n' "$name" >"$log"; return 1
  fi
  printf '%s\t%s\n' "$name" "$log" >>"$d/started.tsv"
  { printf 'run_id=%s\ncheck=%s\nsuite=%s\ntested_tree=%s\nhead_sha=%s\ntree=%s\nbase=%s\n' "$RUN_ID" "${GO_IT_CHECK[$name]:-unlisted}" \
      "${LEDGER_SUITE:-}" "${GO_IT_TREE:-unavailable}" "${GO_IT_HEAD:-unavailable}" "${GO_IT_TREEFLAG:-unknown}" "${BASE:-}"
    printf 'image_id=%s\npackage=%s\ntag=%s\nselector=%s\ntimeout=%s\n' "$(docker image inspect -f '{{.Id}}' "$img" 2>/dev/null || echo unavailable)" \
      "$pkg" "$tag" "$run" "$timeout"
    printf 'command=go test -tags %s -json -list %s %s; go test -tags %s -count=1 -timeout %s -json -run %s %s\n' "$tag" "$run" "$pkg" "$tag" "$timeout" "$run" "$pkg"
    printf 'lifecycle_row=%s\n' "$(if [[ -s "$WORK/db-lifecycle.tsv" ]]; then printf '%s:' "$(wc -l <"$WORK/db-lifecycle.tsv")"; tail -1 "$WORK/db-lifecycle.tsv" | cut -f2-4 | tr '\t' ':'; else echo none; fi)"
  } >"$d/$name.prov"
  : >"$log.events"
  if ! ensure_tools; then
    printf 'G064 go_it: validation tools unavailable; %s did not run (no per-test evidence)\n' "$name" >"$log.events"
    go_it_convert "$name" "$log" 2 || true
    return 1
  fi
}
# go_it_convert NAME LOG RC [interrupted]: gotestreport on LOG.events; writes LOG and $WORK/go-tests/NAME.{tsv,junit.xml,
# status} and exports (unless GO_IT_NO_EXPORT=1). Without a converter the stream fails closed: LOG keeps the raw stream
# (no parser accepts it) and a FAIL invocation row names the cause. Returns 0 only for a passing invocation.
go_it_convert() {
  local name="$1" log="$2" rc="$3" mode="${4:-}" d="$WORK/go-tests" st=2 k v
  local -a args=(--name "$name" --rc "$rc")
  [[ -z "$mode" ]] || args+=(--interrupted)
  while IFS='=' read -r k v; do [[ -n "$k" ]] && args+=(--prov "$k=$v"); done <"$d/$name.prov"
  rm -f -- "$d/$name.log" "$d/$name.tsv" "$d/$name.junit.xml" "$d/$name.status"
  if [[ -n "$VT_IMG" && "$VT_OK" == 1 ]]; then
    vtools -i "$VT_IMG" gotestreport "${args[@]}" <"$log.events" 2>"$d/$name.stderr" | tar -x -C "$d" -f - 2>>"$d/$name.stderr"
    st="${PIPESTATUS[0]}"
  fi
  if [[ "$st" != 0 && "$st" != 1 ]] || [[ ! -s "$d/$name.tsv" || ! -f "$d/$name.log" ]]; then
    cp -- "$log.events" "$d/$name.log"
    printf 'invocation\t%s\t-\t-\t%s\t%s\t0\t0\t0\t0\t-\tgo-tests/%s.log\t1\t%s\tfindings=converter-unavailable(%s)\n' "$name" \
      "$([[ -n "$mode" ]] && echo interrupted || echo fail)" "$rc" "$name" "$(wc -l <"$d/$name.log")" "$st" >"$d/$name.tsv"
    printf '<testsuite name="%s" tests="1" failures="0" errors="1" skipped="0">\n<testcase classname="%s" name="(invocation)"><error message="converter unavailable"/></testcase>\n</testsuite>\n' "$name" "$name" >"$d/$name.junit.xml"
    st=1
  fi
  cp -- "$d/$name.log" "$log"
  # S6 correction 3: the status line goes to stderr (the run's stderr is the same tee as stdout, so it still reaches the
  # console and the harness log) so that a caller captured by command substitution (rls_provision → rls_provisioned_test →
  # go_it) receives only its own printed verdict, never this line.
  log "G064 go_it $name: rc=$rc $(head -1 "$d/$name.status" 2>/dev/null || echo 'fail findings=converter-unavailable')" >&2
  if [[ "${GO_IT_NO_EXPORT:-0}" != 1 ]]; then go_it_publish "$name" || log "G064 go_it $name: export failed" >&2; fi
  return "$st"
}
# go_it_publish NAME: export one converted invocation once (go_tests_export) and mark it done.
go_it_publish() {
  [[ -f "$WORK/go-tests/$1.done" ]] && return 0
  go_tests_export "$1" || return 1
  : >"$WORK/go-tests/$1.done"
}

# rls_logcheck NAME LOG...: logcheck validate and canary scan of the access.decision records carried in LOG, a canary
# scan of the migrations stack's PostgreSQL log, and the planted-canary positive control; prints the failed parts.
rls_logcheck() {
  local name="$1" M="$2"; shift 2
  local lc=""
  ensure_tools || { printf 'no tools'; return; }
  rls_fixed_reset
  (umask 022; printf '%s\n' $RLS_CANARIES >"$RLS_FIXED/canaries.txt"; sed -n 's/^RLS-ACCESSLOG //p' "$@" >"$RLS_FIXED/access.log")
  compose "$M" logs --no-color postgres >"$RLS_FIXED/postgres.log" 2>&1 || lc+="pglog "
  chmod 644 "$RLS_FIXED"/*
  cp "$RLS_FIXED/access.log" "$WORK/$name-access.log"
  if [[ -s "$RLS_FIXED/access.log" ]]; then
    vtools -v "$RLS_FIXED:/r:ro" "$VT_IMG" logcheck validate --service core --file /r/access.log >"$WORK/$name-logcheck.json" 2>&1 || lc+="validate "
  fi
  vtools -v "$RLS_FIXED:/r:ro" "$VT_IMG" logcheck scan --canaries /r/canaries.txt /r/access.log /r/postgres.log >"$WORK/$name-logscan.log" 2>&1 || lc+="scan "
  { cat "$RLS_FIXED/postgres.log"; head -1 "$RLS_FIXED/canaries.txt"; } >"$RLS_FIXED/control.log"
  if vtools -v "$RLS_FIXED:/r:ro" "$VT_IMG" logcheck scan --canaries /r/canaries.txt /r/control.log >/dev/null 2>&1; then lc+="control "; fi
  rm -rf -- "$RLS_FIXED"
  printf '%s' "$lc"
}

# B-RLS-CATALOG: TestRLSCatalog (CT-RLS-1..6 and the 13 rolled-back negative controls N1-N13).
rls_catalog() { # rls_catalog DEST PROJECT IMAGE ITENV
  local M="$1" project="$2" img="$3" itenv="$4" ok=1 log="$WORK/rls-catalog-it.log" n
  it_fresh "$M" "$project" rls-catalog "$log" && rls_go rls-catalog "$project" "$img" "$itenv" "$log" ./internal/migrate/ '^TestRLSCatalog$' || ok=0
  n="$(it_count "$log" TestRLSCatalog)"
  [[ "$n" == 7 ]] || { ok=0; tail -40 "$log"; }
  [[ $ok == 1 ]] && record B-RLS-CATALOG PASS "TestRLSCatalog against real PostgreSQL at version 6 as the attested runtime login: 7/7 subtests pass, none skipped (CT-RLS-1 the 35 relations unchanged, exactly nine forced row-level security tables and security_invoker views; CT-RLS-2 the 26-policy inventory, every pinned policy expression and the acyclic policy graph; CT-RLS-3 every table, column, function, schema and sequence ACL of every role and PUBLIC, incl. the exact runtime and helper column sets; CT-RLS-4 role attributes and memberships; CT-RLS-5 runtime DDL, policy, SET ROLE, TRUNCATE, metadata, row_security = off and every denied write; CT-RLS-6 owner FORCE; CT-RLS-NEG N1-N13 each yield exactly their detector code and the catalog is unchanged after)" || record B-RLS-CATALOG FAIL "row-level security catalog ($n/7 subtests)"
}

# The version-5 infra/postgres/init/10-roles.sh, byte-identical to its accepted blob (PROV-6 existing-volume upgrade).
RLS_ROLES_V5_SHA=a4617f869b4926e5aab1cb756a1c6b1f56ac258591b8de2ed68da9444ff23360
rls_roles_v5() {
  cat <<'RLS_V5_EOF'
#!/bin/sh
# Runs once, on first initialization of the data volume, as the image's init
# superuser. Creates separated roles; role names derive from DATABASE_NAME.
#   <db>_owner     NOLOGIN; owns schema "app" and every object in it
#   <db>_migrator  LOGIN; may SET ROLE <db>_owner (no inheritance) to migrate
#   <db>_runtime   LOGIN; application role: no superuser, no BYPASSRLS,
#                  no owner membership, so row-level security always applies
#   <db>_retention LOGIN; reserved for the audit-retention path
# New tables are unreachable to runtime until explicitly granted.
set -eu
db="${DATABASE_NAME:?DATABASE_NAME is required}"
: "${DB_MIGRATOR_PASSWORD:?}" "${DB_RUNTIME_PASSWORD:?}" "${DB_RETENTION_PASSWORD:?}"

psql -v ON_ERROR_STOP=1 --username "${POSTGRES_USER}" --dbname "${db}" \
  -v db="${db}" \
  -v owner_role="${db}_owner" -v migrator_role="${db}_migrator" \
  -v runtime_role="${db}_runtime" -v retention_role="${db}_retention" \
  <<'SQL'
-- Passwords are read from the environment (\getenv), never passed on the
-- psql command line.
\getenv migrator_pw DB_MIGRATOR_PASSWORD
\getenv runtime_pw DB_RUNTIME_PASSWORD
\getenv retention_pw DB_RETENTION_PASSWORD
CREATE ROLE :"owner_role" NOLOGIN NOSUPERUSER NOBYPASSRLS NOCREATEDB NOCREATEROLE NOINHERIT;
CREATE ROLE :"migrator_role" LOGIN PASSWORD :'migrator_pw' NOSUPERUSER NOBYPASSRLS NOCREATEDB NOCREATEROLE;
GRANT :"owner_role" TO :"migrator_role" WITH INHERIT FALSE, SET TRUE;
CREATE ROLE :"runtime_role" LOGIN PASSWORD :'runtime_pw' NOSUPERUSER NOBYPASSRLS NOCREATEDB NOCREATEROLE NOINHERIT;
CREATE ROLE :"retention_role" LOGIN PASSWORD :'retention_pw' NOSUPERUSER NOBYPASSRLS NOCREATEDB NOCREATEROLE NOINHERIT;

-- Database: no PUBLIC access (CONNECT, TEMPORARY, CREATE); explicit CONNECT only.
REVOKE ALL ON DATABASE :"db" FROM PUBLIC;
GRANT CONNECT ON DATABASE :"db" TO :"migrator_role", :"runtime_role", :"retention_role";

-- Schemas: nothing may be created in public; application objects live in app.
REVOKE CREATE ON SCHEMA public FROM PUBLIC;
CREATE SCHEMA app AUTHORIZATION :"owner_role";
REVOKE ALL ON SCHEMA app FROM PUBLIC;
GRANT USAGE ON SCHEMA app TO :"runtime_role", :"retention_role";

-- Defaults for objects the owner creates: functions are not executable by
-- PUBLIC; tables receive no grants until a migration grants them explicitly.
ALTER DEFAULT PRIVILEGES FOR ROLE :"owner_role" REVOKE EXECUTE ON FUNCTIONS FROM PUBLIC;
ALTER DEFAULT PRIVILEGES FOR ROLE :"owner_role" IN SCHEMA app GRANT USAGE, SELECT ON SEQUENCES TO :"runtime_role";

ALTER ROLE :"runtime_role" SET search_path = app;
ALTER ROLE :"retention_role" SET search_path = app;
ALTER ROLE :"migrator_role" SET search_path = app;
SQL
RLS_V5_EOF
}

# rls_pg NAME DB PASSWORD: a standalone PostgreSQL of the compose image, without init scripts, on the data network;
# owned (docker_run_owned); echoes its container name once it accepts TCP connections.
rls_pg() {
  local name="g008-prov-$1-$RUN_ID" db="$2" pw="$3" i env="$WORK/rls-pg-$1.env" rc=0
  # The superuser password goes through a mode-600 env file (never argv), deleted once the container exists.
  (umask 077; printf 'POSTGRES_PASSWORD=%s\nPOSTGRES_DB=%s\n' "$pw" "$db" >"$env")
  docker_run_owned "prov-$1" -d --name "$name" --label g001.run=$RUN_ID --network "${RLS_PROJECT}_data" \
    --env-file "$env" "$RLS_PG_IMAGE" >/dev/null || rc=$?
  rm -f "$env"
  [[ $rc == 0 ]] || return 1
  # Bounded cleanup (owner rule, 2026-10-14): the image's anonymous data volume is recorded in the run's registry at
  # creation, before rls_provision removes the container with it.
  local v
  for v in $(docker inspect -f '{{range .Mounts}}{{if eq .Type "volume"}}{{.Name}} {{end}}{{end}}' "$name" 2>/dev/null); do
    id_at_start volume "$v" || registry_row volume "$v"
  done
  for i in $(seq 1 120); do
    if docker exec "$name" psql -X -tA -h 127.0.0.1 -U postgres -d "$db" -c 'SELECT 1' >/dev/null 2>&1; then echo "$name"; return 0; fi
    sleep 1
  done
  return 1
}

# rls_prov CONTAINER DB USER ENVFILE: stream infra/postgres/provision-roles.sql over stdin as USER; the passwords come
# from ENVFILE (docker exec --env-file), never from argv.
rls_prov() {
  docker exec -i --env-file "$4" "$1" psql -X -q -v ON_ERROR_STOP=1 -h 127.0.0.1 -U "$3" -d "$2" -v db="$2" -f - <"$RLS_REPO/infra/postgres/provision-roles.sql"
}

# rls_prov_catalog CONTAINER DB: the role, membership, database, schema, default-privilege and role-setting catalog with
# the database name normalized to DB_ and the database owner's own ACL entries omitted (the PROV-1 equivalence basis).
rls_prov_catalog() {
  docker exec "$1" psql -X -tA -h 127.0.0.1 -U postgres -d "$2" -c "
SELECT coalesce(string_agg(x, '|' ORDER BY x COLLATE \"C\"), '') FROM (
  SELECT 'role:' || replace(rolname, current_database() || '_', 'DB_') || ':' || rolcanlogin::text || ',' || rolinherit::text || ',' ||
         rolsuper::text || ',' || rolbypassrls::text || ',' || rolcreaterole::text || ',' || rolcreatedb::text || ',' || rolreplication::text AS x
  FROM pg_roles WHERE rolname LIKE current_database() || '\_%'
  UNION ALL
  SELECT 'member:' || replace(m.member::regrole::text, current_database() || '_', 'DB_') || '>' ||
         replace(m.roleid::regrole::text, current_database() || '_', 'DB_') || ':' || m.inherit_option::text || ',' || m.set_option::text || ',' || m.admin_option::text
  FROM pg_auth_members m WHERE m.roleid::regrole::text LIKE current_database() || '\_%' AND m.member::regrole::text LIKE current_database() || '\_%'
  UNION ALL
  SELECT 'db:' || CASE a.grantee WHEN 0 THEN 'PUBLIC' ELSE replace(a.grantee::regrole::text, current_database() || '_', 'DB_') END || '=' || a.privilege_type
  FROM pg_database d CROSS JOIN LATERAL aclexplode(coalesce(d.datacl, acldefault('d', d.datdba))) a WHERE d.datname = current_database() AND a.grantee <> d.datdba
  UNION ALL
  SELECT 'schema:' || n.nspname || ':' || replace(n.nspowner::regrole::text, current_database() || '_', 'DB_') || ':' ||
         coalesce((SELECT string_agg(v, ',' ORDER BY v COLLATE \"C\") FROM (
                     SELECT CASE a.grantee WHEN 0 THEN 'PUBLIC' ELSE replace(a.grantee::regrole::text, current_database() || '_', 'DB_') END || '=' || a.privilege_type AS v
                     FROM aclexplode(n.nspacl) a WHERE a.grantee <> n.nspowner) q), '-')
  FROM pg_namespace n WHERE n.nspname IN ('app', 'audit', 'public')
  UNION ALL
  SELECT 'default:' || replace(d.defaclrole::regrole::text, current_database() || '_', 'DB_') || ':' || coalesce(d.defaclnamespace::regnamespace::text, '-') || ':' ||
         d.defaclobjtype::text || ':' || replace(d.defaclacl::text, current_database() || '_', 'DB_')
  FROM pg_default_acl d
  UNION ALL
  SELECT 'setting:' || replace(r.rolname, current_database() || '_', 'DB_') || ':' || array_to_string(s.setconfig, ',')
  FROM pg_db_role_setting s JOIN pg_roles r ON r.oid = s.setrole WHERE r.rolname LIKE current_database() || '\_%'
) t"
}

# rls_provisioned_test CONTAINER DB SUPERPASSWORD LOG: TestRLSProvisioned (PROV-2/3) against that instance.
rls_provisioned_test() {
  local env="$WORK/rls-prov-it.env"
  (umask 077; printf 'DATABASE_NAME=%s\nDB_HOST=%s\nDB_PORT=5432\nDB_MIGRATOR_PASSWORD=%s\nDB_RUNTIME_PASSWORD=%s\nDB_SUPERUSER=postgres\nDB_SUPERUSER_PASSWORD=%s\nAPP_ENV=test\n' \
    "$2" "$1" "$RLS_PW_MIG" "$RLS_PW_RT" "$3" >"$env")
  local rc=0
  go_it "$(basename -- "$4" .log)" "$RLS_PROJECT" "$RLS_IMG" "$env" "$4" ./internal/migrate/ '^TestRLSProvisioned$' || rc=$?
  rm -f "$env"
  [[ $rc == 0 ]] && grep -q -- '^--- PASS: TestRLSProvisioned ' "$4" && ! grep -qE -- '--- (SKIP|FAIL)' "$4"
}

# rls_provision DEST PROJECT IMAGE: PROV-1..6 (Plan r3 §8.3) on owned standalone instances without init scripts. Prints
# the failed parts (nothing when all hold). The detached instances are removed here, as every other detached
# docker_run_owned container is (the run's cleanup tears down Compose projects only).
rls_provision() {
  rls_provision_parts "$@"
  local n
  # -v: the PostgreSQL image's anonymous data volume goes with its container.
  for n in ref su adm fail v5; do docker rm -f -v "g008-prov-$n-$RUN_ID" >/dev/null 2>&1 || true; done
}

rls_provision_parts() {
  local M="$1" bad="" db=prov_db envf="$WORK/rls-prov.env" envmiss="$WORK/rls-prov-missing.env" ref su adm fail v5 a b c rc
  RLS_PROJECT="$2" RLS_IMG="$3" RLS_REPO="$M"
  RLS_PG_IMAGE="$(compose "$M" config --format json 2>/dev/null | yq_stdin -p json -o tsv '.services.postgres.image')"
  [[ -n "$RLS_PG_IMAGE" ]] || { printf 'image '; return; }
  RLS_PW_MIG="$(random_hex 16)" RLS_PW_RT="$(random_hex 16)"
  local pw_ret pw_ref pw_su pw_adm pw_fail pw_v5
  pw_ret="$(random_hex 16)" pw_ref="$(random_hex 16)" pw_su="$(random_hex 16)" pw_adm="$(random_hex 16)" pw_fail="$(random_hex 16)" pw_v5="$(random_hex 16)"
  printf '%s\n' "$RLS_PW_MIG" "$RLS_PW_RT" "$pw_ret" "$pw_ref" "$pw_su" "$pw_adm" "$pw_fail" "$pw_v5" >>"$SECRETS_FILE"
  (umask 077
   printf 'DB_MIGRATOR_PASSWORD=%s\nDB_RUNTIME_PASSWORD=%s\nDB_RETENTION_PASSWORD=%s\nDATABASE_NAME=%s\nPOSTGRES_USER=postgres\n' "$RLS_PW_MIG" "$RLS_PW_RT" "$pw_ret" "$db" >"$envf"
   printf 'DB_MIGRATOR_PASSWORD=%s\nDB_RETENTION_PASSWORD=%s\n' "$RLS_PW_MIG" "$pw_ret" >"$envmiss")
  # PROV-1 reference: a no-init instance initialized by this tree's 10-roles.sh, streamed over stdin.
  ref="$(rls_pg ref "$db" "$pw_ref")" || { printf 'ref '; return; }
  docker exec -i --env-file "$envf" "$ref" bash -s <"$M/infra/postgres/init/10-roles.sh" >"$WORK/rls-prov-ref.log" 2>&1 || bad+="ref-init "
  a="$(rls_prov_catalog "$ref" "$db")"
  # PROV-1/4 superuser: provision, compare, repeat (no change).
  su="$(rls_pg su "$db" "$pw_su")" || { printf 'su '; return; }
  rls_prov "$su" "$db" postgres "$envf" >"$WORK/rls-prov-su.log" 2>&1 || bad+="su-provision "
  b="$(rls_prov_catalog "$su" "$db")"
  [[ -n "$a" && "$a" == "$b" ]] || { bad+="PROV-1-superuser "; printf '%s\n---\n%s\n' "$a" "$b" >"$WORK/rls-prov-1-su.diff"; }
  rls_prov "$su" "$db" postgres "$envf" >>"$WORK/rls-prov-su.log" 2>&1 || bad+="su-repeat "
  [[ "$(rls_prov_catalog "$su" "$db")" == "$b" ]] || bad+="PROV-4-repeat "
  rls_provisioned_test "$su" "$db" "$pw_su" "$WORK/rls-provisioned-su.log" || { bad+="PROV-2/3-superuser "; tail -30 "$WORK/rls-provisioned-su.log"; }
  # PROV-1 non-superuser administrator: a CREATEROLE role that owns the database.
  adm="$(rls_pg adm postgres "$pw_adm")" || { printf 'adm '; return; }
  docker exec "$adm" psql -X -q -v ON_ERROR_STOP=1 -h 127.0.0.1 -U postgres -d postgres -c "CREATE ROLE prov_admin LOGIN CREATEROLE" \
    -c "CREATE DATABASE $db OWNER prov_admin" >"$WORK/rls-prov-adm.log" 2>&1 || bad+="adm-setup "
  rls_prov "$adm" "$db" prov_admin "$envf" >>"$WORK/rls-prov-adm.log" 2>&1 || bad+="adm-provision "
  c="$(rls_prov_catalog "$adm" "$db")"
  [[ -n "$a" && "$a" == "$c" ]] || { bad+="PROV-1-admin "; printf '%s\n---\n%s\n' "$a" "$c" >"$WORK/rls-prov-1-adm.diff"; }
  rc="$(docker exec "$adm" psql -X -tA -h 127.0.0.1 -U postgres -d "$db" -c "SELECT coalesce(string_agg(m.roleid::regrole::text || ':' || m.inherit_option::text || ',' || m.set_option::text || ',' || m.admin_option::text, ' ' ORDER BY m.roleid::regrole::text COLLATE \"C\"), '')
FROM pg_auth_members m WHERE m.member = 'prov_admin'::regrole")"
  [[ "$rc" == "${db}_auth_reader:false,false,true ${db}_claim_writer:false,false,true ${db}_migrator:false,false,true ${db}_owner:false,false,true ${db}_policy_reader:false,false,true ${db}_provision_writer:false,false,true ${db}_retention:false,false,true ${db}_runtime:false,false,true" ]] || { bad+="PROV-1-admin-residual "; log "admin residual: $rc" >&2; }
  rls_provisioned_test "$adm" "$db" "$pw_adm" "$WORK/rls-provisioned-adm.log" || { bad+="PROV-2/3-admin "; tail -30 "$WORK/rls-provisioned-adm.log"; }
  # PROV-5 failure controls: each exits non-zero and changes nothing.
  fail="$(rls_pg fail "$db" "$pw_fail")" || { printf 'fail '; return; }
  a="$(rls_prov_catalog "$fail" "$db")"
  if rls_prov "$fail" "$db" postgres "$envmiss" >"$WORK/rls-prov-fail.log" 2>&1; then bad+="PROV-5-missing-password "; fi
  [[ "$(rls_prov_catalog "$fail" "$db")" == "$a" ]] || bad+="PROV-5-missing-password-changed "
  docker exec "$fail" psql -X -q -h 127.0.0.1 -U postgres -d "$db" -c "CREATE ROLE ${db}_runtime LOGIN BYPASSRLS" >>"$WORK/rls-prov-fail.log" 2>&1
  a="$(rls_prov_catalog "$fail" "$db")"
  if rls_prov "$fail" "$db" postgres "$envf" >>"$WORK/rls-prov-fail.log" 2>&1; then bad+="PROV-5-unsafe-runtime "; fi
  [[ "$(rls_prov_catalog "$fail" "$db")" == "$a" ]] || bad+="PROV-5-unsafe-runtime-changed "
  docker exec "$fail" psql -X -q -h 127.0.0.1 -U postgres -d "$db" -c "DROP ROLE ${db}_runtime" -c "CREATE ROLE ${db}_policy_reader LOGIN" >>"$WORK/rls-prov-fail.log" 2>&1
  a="$(rls_prov_catalog "$fail" "$db")"
  if rls_prov "$fail" "$db" postgres "$envf" >>"$WORK/rls-prov-fail.log" 2>&1; then bad+="PROV-5-helper-login "; fi
  [[ "$(rls_prov_catalog "$fail" "$db")" == "$a" ]] || bad+="PROV-5-helper-login-changed "
  # PROV-6 existing-volume upgrade: an instance initialized by the version-5 10-roles.sh gets only the helpers.
  [[ "$(rls_roles_v5 | sha256sum | cut -d' ' -f1)" == "$RLS_ROLES_V5_SHA" ]] || bad+="PROV-6-v5-script "
  v5="$(rls_pg v5 "$db" "$pw_v5")" || { printf 'v5 '; return; }
  rls_roles_v5 | docker exec -i --env-file "$envf" "$v5" bash -s >"$WORK/rls-prov-v5.log" 2>&1 || bad+="v5-init "
  a="$(rls_prov_catalog "$v5" "$db")"
  rls_prov "$v5" "$db" postgres "$envf" >>"$WORK/rls-prov-v5.log" 2>&1 || bad+="v5-provision "
  b="$(rls_prov_catalog "$v5" "$db")"
  rc="$(LC_ALL=C comm -3 <(tr '|' '\n' <<<"$a" | LC_ALL=C sort) <(tr '|' '\n' <<<"$b" | LC_ALL=C sort) | tr -d '\t' | tr '\n' ' ')"
  [[ "$rc" == "member:DB_owner>DB_auth_reader:false,true,false member:DB_owner>DB_claim_writer:false,true,false member:DB_owner>DB_policy_reader:false,true,false member:DB_owner>DB_provision_writer:false,true,false role:DB_auth_reader:false,false,false,false,false,false,false role:DB_claim_writer:false,false,false,false,false,false,false role:DB_policy_reader:false,false,false,false,false,false,false role:DB_provision_writer:false,false,false,false,false,false,false schema:audit:DB_owner:- " ]] || { bad+="PROV-6-delta "; log "PROV-6 delta: $rc" >&2; }
  rls_provisioned_test "$v5" "$db" "$pw_v5" "$WORK/rls-provisioned-v5.log" || { bad+="PROV-6-migrate "; tail -30 "$WORK/rls-provisioned-v5.log"; }
  rm -f "$envf" "$envmiss"
  printf '%s' "$bad"
}

# B-RLS-MIGRATION: TestRLSMigration (MG-RLS-1..5b and the migration role-switch evidence), MG-RLS-6 around a PostgreSQL
# restart (the seed's IDs pass from its stdout to the verify step's environment) and PROV-1..6.
rls_migration() { # rls_migration DEST PROJECT IMAGE ITENV
  local M="$1" project="$2" img="$3" itenv="$4" ok=1 log="$WORK/rls-migration-it.log" n ids prov
  it_fresh "$M" "$project" rls-migration "$log" && rls_go rls-migration "$project" "$img" "$itenv" "$log" ./internal/migrate/ '^TestRLSMigration$' || ok=0
  n="$(it_count "$log" TestRLSMigration)"
  [[ "$n" == 7 ]] || { ok=0; tail -40 "$log"; }
  it_fresh "$M" "$project" rls-persist-seed "$WORK/rls-persist-seed.log" && rls_go rls-persist-seed "$project" "$img" "$itenv" "$WORK/rls-persist-seed.log" ./internal/migrate/ '^TestRLSPersistSeed$' || { ok=0; tail -20 "$WORK/rls-persist-seed.log"; }
  ids="$(sed -n 's/^RLS-PERSIST //p' "$WORK/rls-persist-seed.log" | head -1)"
  compose "$M" stop postgres >"$WORK/rls-persist-restart.log" 2>&1 || ok=0
  compose "$M" up -d --wait postgres >>"$WORK/rls-persist-restart.log" 2>&1 || { ok=0; tail -20 "$WORK/rls-persist-restart.log"; }
  rls_go rls-persist-verify "$project" "$img" "$itenv" "$WORK/rls-persist-verify.log" ./internal/migrate/ '^TestRLSPersistVerify$' -e RLS_PERSIST_IDS="$ids" || { ok=0; tail -20 "$WORK/rls-persist-verify.log"; }
  for out in "$WORK/rls-persist-seed.log" "$WORK/rls-persist-verify.log"; do
    grep -qE -- '^--- PASS: TestRLSPersist(Seed|Verify) ' "$out" && ! grep -qE -- '--- (SKIP|FAIL)' "$out" || ok=0
  done
  # The capture must hold only rls_provision_parts' printed failed parts: every status line of the helpers it runs
  # (go_it, it_*, its own log calls) goes to stderr (S6 correction 3).
  res_scope_begin PROV; prov="$(rls_provision "$M" "$project" "$img")"; res_scope_end PROV
  [[ -z "$prov" ]] || { ok=0; log "provisioning: $prov"; }
  [[ $ok == 1 ]] && record B-RLS-MIGRATION PASS "TestRLSMigration: 7/7 subtests pass, none skipped (MG-RLS-1 empty to version 6, MG-RLS-2 repeated up/status, MG-RLS-3 populated version 5 to 6 with every row of all ten tables identical, MG-RLS-4 reapply keeps the catalog, MG-RLS-5 down to 5 restores the exact version-5 catalog with rows unchanged and no helper grant, function or CREATE left, then up; MG-RLS-5b G008 code at version 5 fails closed; MG-RLS-C4 a rolled-back up restores the owner role and the version-5 catalog); MG-RLS-6 version 6 and rows survive a PostgreSQL stop/start; PROV-1 a no-init instance provisioned by infra/postgres/provision-roles.sql as the superuser and as a CREATEROLE database owner equals the 10-roles.sh catalog, with the administrator's residual ADMIN-only; PROV-2/3 migrate and runtime/owner-FORCE checks on both; PROV-4 a repeat run changes nothing; PROV-5 a missing password, an unsafe runtime role and a helper with LOGIN each exit non-zero with nothing changed; PROV-6 a version-5 10-roles.sh instance gains exactly the four helpers (auth reader, policy reader; G019 claim writer; G062 provision writer) and their memberships, then migrates" || record B-RLS-MIGRATION FAIL "row-level security migrations ($n/7 subtests; persistence; provisioning: ${prov:-ok})"
}

# B-RLS-IT: TestRLSIsolation, TestRLSPool and TestRLSFootprint (core/internal/tenancy/access) and TestRLSIdentityPool and
# TestRLSFirstSignInRace (core/internal/identity), then the log checks. E-RLS lines are kept as rls-erls.jsonl.
rls_it() { # rls_it DEST PROJECT IMAGE ITENV
  local M="$1" project="$2" img="$3" itenv="$4" ok=1 log="$WORK/rls-it.log" ilog="$WORK/rls-identity-it.log" n p f i r lc triggers scope
  it_fresh "$M" "$project" rls-it "$log" && rls_go rls-it "$project" "$img" "$itenv" "$log" ./internal/tenancy/access/ '^(TestRLSIsolation|TestRLSPool|TestRLSFootprint)$' || ok=0
  n="$(it_count "$log" TestRLSIsolation)"; p="$(it_count "$log" TestRLSPool)"
  grep -q -- '^--- PASS: TestRLSFootprint ' "$log" && f=pass || f=FAIL
  scope="$(sed -n 's/^RLS-SCOPE //p' "$log" | head -1)"
  [[ "$n" == 21 && "$p" == 7 &&"$f" == pass && "$scope" == '{"cases":40,"disagreements":0}' ]] || { ok=0; tail -80 "$log"; }
  it_fresh "$M" "$project" rls-identity-it "$ilog" && rls_go rls-identity-it "$project" "$img" "$itenv" "$ilog" ./internal/identity/ '^(TestRLSIdentityPool|TestRLSFirstSignInRace)$' || ok=0
  i="$(it_count "$ilog" TestRLSIdentityPool)"
  grep -q -- '^--- PASS: TestRLSFirstSignInRace ' "$ilog" && r=pass || r=FAIL
  [[ "$i" == 6 && "$r" == pass ]] || { ok=0; tail -60 "$ilog"; }
  grep -hE '^(E-RLS-(1|2|3|SUBPLAN|TRIGGER)|RLS-POOL|RLS-SCOPE) ' "$log" "$ilog" >"$WORK/rls-erls.jsonl" || true
  if [[ -n "$REPORT_DIR" ]]; then cp "$WORK/rls-erls.jsonl" "$REPORT_DIR/" 2>/dev/null || true; fi
  triggers="$(grep -c '^E-RLS-TRIGGER ' "$WORK/rls-erls.jsonl" || true)"
  [[ "$triggers" == 0 ]] || log "E-RLS: $triggers escalation trigger line(s): return to Codex (rls-erls.jsonl)"
  lc="$(rls_logcheck rls-it "$M" "$log")"
  [[ -z "$lc" ]] || ok=0
  [[ $ok == 1 ]] && record B-RLS-IT PASS "TestRLSIsolation against real PostgreSQL at version 22 on the attested runtime login: 21/21 subtest results pass (G020 RLS-LABEL with its reads, writes and isolated foreign-key parts, RLS-LABEL-FIXTURE with its T5 cases (a), (b) and (d)-(g), the two label relations compared in every RLS-SCOPE case and RLS-ORG's fixed-actor label counts; G018 RLS-SCOPE-INSERT; G019 RLS-SCOPE-MUTATE; RLS-ATTEST incl. the privileged-assertion negative, RLS-SCOPE: the independent oracle's 40 cases (G060: Organization staff with the Group-management override off and on) agree exactly, visible sets, ok/no_update and the Plan-owned mechanism of every refused update; last-path removal and restoration; RLS-CTX every field varied independently; RLS-ORG; RLS-AUTH fixed-actor privacy, E1-E3 and the E5 controls (a)-(e); RLS-IMMUT lock positives and mutation negatives; RLS-PROJ views equal base, read-only, own-fact readers); TestRLSPool 7/7 (P1-P10 on one backend: no carried authority, refused poisoned initialization with discard, clean reuse or discard after cancellation, rollback and errors); TestRLSFootprint (E-RLS-1..3 recorded, $triggers trigger lines); TestRLSIdentityPool 6/6 (I1-I5, E5 (f)) and TestRLSFirstSignInRace (G016 D-2(a): 32 concurrent first sign-ins of one unmapped principal on a routing tenant host each refused organization_not_member by the final issuance guard, no user, principal or session created); access.decision records catalogued, no canary in them or the PostgreSQL log, planted canary detected" || record B-RLS-IT FAIL "row-level security isolation ($n/21 isolation, $p/7 pool, footprint $f, $i/6 identity pool, race $r; scope ${scope:-none}; logcheck: ${lc:-ok})"
}

# B-RLS-ORDERING: TestRLSOrdering (OR-1..5; G060 OR-6 and OR-7; G020 OR-8) with G007's deterministic barriers, then the log checks.
rls_ordering() { # rls_ordering DEST PROJECT IMAGE ITENV
  local M="$1" project="$2" img="$3" itenv="$4" ok=1 log="$WORK/rls-ordering-it.log" n lc
  it_fresh "$M" "$project" rls-ordering "$log" && rls_go rls-ordering "$project" "$img" "$itenv" "$log" ./internal/tenancy/access/ '^TestRLSOrdering$' || ok=0
  n="$(it_count "$log" TestRLSOrdering)"
  [[ "$n" == 8 ]] || { ok=0; tail -60 "$log"; }
  grep -E '^ORDERING ' "$log" >"$WORK/rls-ordering.jsonl" || true
  if [[ -n "$REPORT_DIR" ]]; then cp "$WORK/rls-ordering.jsonl" "$REPORT_DIR/" 2>/dev/null || true; fi
  lc="$(rls_logcheck rls-ordering "$M" "$log")"
  [[ -z "$lc" ]] || ok=0
  [[ $ok == 1 ]] && record B-RLS-ORDERING PASS "TestRLSOrdering against real PostgreSQL: 8/8 subtests pass, none skipped (G020: OR-8 a raw definition delete versus an in-flight assignment of that definition; G060: OR-6 a raw owner override UPDATE queued behind an in-flight DisableUser of that owner; OR-7 a raw Organization removal versus a raw Group-admin-to-co-owner promotion of the same subject in three interleavings; OR-1 last-path removal, OR-2 unlink and OR-3 the flag turned off versus a profile write in orders A, B and C: serialized before the change or denied with the target unchanged; OR-4 an old context, connection and prepared statement after downgrade, removal, unlink, flag, disable, session revoke and expiry, and manager demotion shrinking raw membership visibility at once; OR-5 a real slug change orders a concurrent foreign-key insert behind the upgraded Group lock); access.decision records catalogued, no canary in them or the PostgreSQL log, planted canary detected" || record B-RLS-ORDERING FAIL "row-level security ordering ($n/8 subtests; logcheck: ${lc:-ok})"
}

# G009 B-OBJ-CATALOG, B-OBJ-MIGRATION, B-OBJ-IT, B-OBJ-ORDERING and B-OBJ-FAULTS (Plan r1 §5.2, §5.3, §5.5). The
# migrations project's private MinIO is started from local.yml; the bucket and the scoped service credential are created
# by the built init image with the init script on stdin and an env-file under $WORK (umask 077, removed after use), so no
# per-run host directory is bind-mounted. Test containers receive the scoped OBJECT_STORE_* settings only, never the
# MinIO root credentials, and return their records the RLS way (RLS-ACCESSLOG lines on stdout).
OBJ_ENV=""
obj_minio() { # obj_minio DEST PROJECT -> 0 when MinIO and the scoped credential are ready
  local M="$1" project="$2" envf="$WORK/obj-minio-init.env" rc=0
  compose "$M" build minio minio-init >"$WORK/obj-minio-build.log" 2>&1 || { tail -20 "$WORK/obj-minio-build.log"; return 1; }
  compose "$M" up -d --wait minio >"$WORK/obj-minio-up.log" 2>&1 || { tail -20 "$WORK/obj-minio-up.log"; return 1; }
  (umask 077; printf 'MINIO_ROOT_USER=%s\nMINIO_ROOT_PASSWORD=%s\nOBJECT_STORE_BUCKET=%s\nOBJECT_STORE_ACCESS_KEY=%s\nOBJECT_STORE_SECRET_KEY=%s\n' \
    "$(env_get "$M" MINIO_ROOT_USER)" "$(env_get "$M" MINIO_ROOT_PASSWORD)" "$(env_get "$M" OBJECT_STORE_BUCKET)" \
    "$(env_get "$M" OBJECT_STORE_ACCESS_KEY)" "$(env_get "$M" OBJECT_STORE_SECRET_KEY)" >"$envf")
  docker_run_owned obj-minio-init --rm -i --label g001.run=$RUN_ID --network "${project}_data" --env-file "$envf" "$project-minio-init" \
    sh -s <"$M/infra/minio/init.sh" >"$WORK/obj-minio-init.log" 2>&1 || rc=$?
  rm -f "$envf"
  [[ $rc == 0 ]] || { tail -20 "$WORK/obj-minio-init.log"; return 1; }
  return 0
}

# obj_env ITENV DEST: the object test env-file (ITENV plus the scoped OBJECT_STORE_* settings of DEST).
obj_env() {
  local itenv="$1" M="$2" k
  OBJ_ENV="$WORK/obj-it.env"
  # The redirect sits inside the subshell, after umask 077, so the file is created owner-only.
  (umask 077
   { cat "$itenv"
     for k in OBJECT_STORE_ENDPOINT OBJECT_STORE_REGION OBJECT_STORE_BUCKET OBJECT_STORE_FORCE_PATH_STYLE OBJECT_STORE_ACCESS_KEY OBJECT_STORE_SECRET_KEY; do
       printf '%s=%s\n' "$k" "$(env_get "$M" "$k")"
     done; } >"$OBJ_ENV")
  [[ "$(stat -c %a "$OBJ_ENV")" == 600 ]] || { log "obj-it.env mode $(stat -c %a "$OBJ_ENV")"; return 1; }
  ! grep -q '^MINIO_ROOT' "$OBJ_ENV"
}

# obj_fetch PORT: GET of the signed URL held in OBJ_URL (never an argument, log or file) through the host port, with
# Host 127.0.0.1:PORT exactly as a browser sends it. Prints "STATUS SHA256 DISPOSITION|TYPE|CACHE". OBJ_PATH_ONLY=1
# drops the query (the unsigned control).
obj_fetch() {
  local port="$1" target fd line status="" disp="" type="" cache="" sum hdr
  target="${OBJ_URL#http://127.0.0.1:"$port"}"
  [[ "${OBJ_PATH_ONLY:-0}" == 1 ]] && target="${target%%\?*}"
  if [[ "$PROBE_MODE" == devtcp ]]; then
    exec {fd}<>"/dev/tcp/127.0.0.1/$port" || { printf '000 - -\n'; return; }
    printf 'GET %s HTTP/1.0\r\nHost: 127.0.0.1:%s\r\nConnection: close\r\n\r\n' "$target" "$port" >&"$fd"
    IFS= read -r -t 15 -u "$fd" line || line=""
    line="${line%$'\r'}"; status="${line#* }"; status="${status%% *}"
    while IFS= read -r -t 15 -u "$fd" line; do
      line="${line%$'\r'}"
      [[ -z "$line" ]] && break
      case "${line,,}" in
        content-disposition:*) disp="${line#*: }" ;;
        content-type:*) type="${line#*: }" ;;
        cache-control:*) cache="${line#*: }" ;;
      esac
    done
    sum="$(sha256sum <&"$fd" | cut -d' ' -f1)"
    exec {fd}>&-
  else
    hdr="$WORK/obj-fetch.hdr"
    sum="$(printf 'url = "http://127.0.0.1:%s%s"\nheader = "Host: 127.0.0.1:%s"\n' "$port" "$target" "$port" |
      curl -s -m 15 -K - -D "$hdr" -o - | sha256sum | cut -d' ' -f1)"
    status="$(head -1 "$hdr" 2>/dev/null | awk '{print $2}')"
    disp="$(sed -n 's/^[Cc]ontent-[Dd]isposition: //p' "$hdr" | tr -d '\r')"
    type="$(sed -n 's/^[Cc]ontent-[Tt]ype: //p' "$hdr" | tr -d '\r')"
    cache="$(sed -n 's/^[Cc]ache-[Cc]ontrol: //p' "$hdr" | tr -d '\r')"
    rm -f "$hdr"
  fi
  printf '%s %s %s|%s|%s\n' "${status:-000}" "$sum" "$disp" "$type" "$cache"
}

# obj_sign_host PROJECT IMAGE PORT: OBJ-SIGN-HOST (correction 4). A detached check container issues a URL signed for
# http://127.0.0.1:PORT (the project's published MinIO API port) into a tmpfs; the harness reads it with docker exec
# into a variable, fetches the unchanged URL through the host port, compares status, body digest and overrides, checks
# that the bare path is refused, then releases the container. Prints the failed parts (UNMET without a host probe).
obj_sign_host() {
  local project="$1" img="$2" port="$3" name="g009-signhost-$RUN_ID" bad="" i got want OBJ_URL rc
  if [[ "$PROBE_MODE" == none ]]; then printf 'UNMET'; return; fi
  local olog="$WORK/obj-sign-host-it.log"
  # G064 (Plan r2 §5.1): the detached invocation runs go_it's script; its stream is read from the container log.
  go_it_begin obj-sign-host "$img" "$olog" ./internal/tenancy/access/ '^TestObjectsSignHost$' integration 15m >&2 || { printf 'start '; return; }
  docker_run_owned obj-sign-host -d --name "$name" --label g001.run=$RUN_ID --network "${project}_data" --env-file "$OBJ_ENV" \
    -e RLS_CANARIES="$RLS_CANARIES" -e TENANCY_CANARY_FILE=/tmp/rls-canaries.txt -e TENANCY_LOG_FILE=/tmp/rls-access.log \
    -e OBJ_SIGN_HOST_ENDPOINT="http://127.0.0.1:$port" -e OBJ_SIGN_HOST_DIR=/run/objsign --tmpfs /run/objsign:mode=0700 "$img" \
    sh -c "$GO_IT_SH" sh '^TestObjectsSignHost$' ./internal/tenancy/access/ integration 15m \
    >/dev/null || { go_it_convert obj-sign-host "$olog" 125 >&2 || true; printf 'start '; return; }
  for ((i = 0; i < 900; i++)); do
    # The readiness line is an output event of the -json stream (the plain-text form is still accepted).
    docker logs "$name" 2>&1 | grep -qE '^OBJ-SIGN-HOST-READY$|"Output":"OBJ-SIGN-HOST-READY\\n"' && break
    [[ "$(docker inspect -f '{{.State.Running}}' "$name" 2>/dev/null)" == true ]] || break
    sleep 1
  done
  OBJ_URL="$(docker exec "$name" cat /run/objsign/url 2>/dev/null)"
  want="$(docker exec "$name" cat /run/objsign/sha256 2>/dev/null)"
  if [[ "$OBJ_URL" == "http://127.0.0.1:$port/"* && "$want" =~ ^[0-9a-f]{64}$ ]]; then
    got="$(OBJ_URL="$OBJ_URL" obj_fetch "$port")"
    [[ "$got" == "200 $want attachment|text/plain|no-store" ]] || bad+="fetch(${got%% *}) "
    got="$(OBJ_URL="$OBJ_URL" OBJ_PATH_ONLY=1 obj_fetch "$port")"
    [[ "${got%% *}" == 403 ]] || bad+="unsigned(${got%% *}) "
  else
    bad+="issue "
  fi
  unset OBJ_URL
  docker exec "$name" touch /run/objsign/done >/dev/null 2>&1 || true
  rc="$(timeout 120 docker wait "$name" 2>/dev/null || echo 255)"
  docker logs "$name" >"$olog.events" 2>&1 || true
  docker rm -f -v "$name" >/dev/null 2>&1 || true
  [[ "$rc" =~ ^[0-9]+$ ]] || rc=255
  go_it_convert obj-sign-host "$olog" "$rc" >&2 || bad+="report "
  [[ "$rc" == 0 ]] && grep -q -- '^--- PASS: TestObjectsSignHost ' "$olog" || bad+="test "
  printf '%s' "$bad"
}

# B-OBJ-CATALOG: TestObjectsCatalog (CT-OBJ-1, CT-OBJ-2 and the rolled-back detectors N-OBJ-1..8).
obj_catalog() { # obj_catalog DEST PROJECT IMAGE
  local M="$1" project="$2" img="$3" ok=1 log="$WORK/obj-catalog-it.log" n
  it_fresh "$M" "$project" obj-catalog "$log" && rls_go obj-catalog "$project" "$img" "$OBJ_ENV" "$log" ./internal/migrate/ '^TestObjectsCatalog$' || ok=0
  n="$(it_count "$log" TestObjectsCatalog)"
  [[ "$n" == 3 ]] || { ok=0; tail -40 "$log"; }
  [[ $ok == 1 ]] && record B-OBJ-CATALOG PASS "TestObjectsCatalog against real PostgreSQL at version 7: 3/3 subtests pass, none skipped (CT-OBJ-1 the exact version-7 delta on the unchanged version-6 catalog: stored_objects forced, the 29-policy inventory, the runtime SELECT, six INSERT columns and UPDATE(state, sha256), the invoker guard with no grant and its exact source digest, the columns with the generated key, the constraints, the trigger and the two indexes, 38 relations; CT-OBJ-2 runtime DELETE, TRUNCATE, key/creator/scope writes, state and created_at inserts, DDL and policy changes denied; CT-OBJ-NEG N-OBJ-1..8 (NO FORCE, an extra DELETE grant, SELECT USING (true), the trigger disabled, a SECURITY DEFINER guard, a PUBLIC grant, the key not generated, a replaced guard body) each make the detector fail, and the catalog is unchanged after)" || record B-OBJ-CATALOG FAIL "stored-object catalog ($n/3 subtests)"
}

# B-OBJ-MIGRATION: TestObjectsMigration (MG-OBJ-1..5), then persistence across a PostgreSQL and MinIO restart (the seed's
# identifiers pass from its stdout to the verify step's environment; the guarded Get must return identical bytes).
obj_migration() { # obj_migration DEST PROJECT IMAGE
  local M="$1" project="$2" img="$3" ok=1 log="$WORK/obj-migration-it.log" n ids out
  it_fresh "$M" "$project" obj-migration "$log" && rls_go obj-migration "$project" "$img" "$OBJ_ENV" "$log" ./internal/migrate/ '^TestObjectsMigration$' || ok=0
  n="$(it_count "$log" TestObjectsMigration)"
  [[ "$n" == 5 ]] || { ok=0; tail -40 "$log"; }
  it_fresh "$M" "$project" obj-persist-seed "$WORK/obj-persist-seed.log" && rls_go obj-persist-seed "$project" "$img" "$OBJ_ENV" "$WORK/obj-persist-seed.log" ./internal/tenancy/access/ '^TestObjectsPersistSeed$' || { ok=0; tail -20 "$WORK/obj-persist-seed.log"; }
  ids="$(sed -n 's/^OBJ-PERSIST //p' "$WORK/obj-persist-seed.log" | head -1)"
  compose "$M" stop postgres minio >"$WORK/obj-persist-restart.log" 2>&1 || ok=0
  compose "$M" up -d --wait postgres minio >>"$WORK/obj-persist-restart.log" 2>&1 || { ok=0; tail -20 "$WORK/obj-persist-restart.log"; }
  rls_go obj-persist-verify "$project" "$img" "$OBJ_ENV" "$WORK/obj-persist-verify.log" ./internal/tenancy/access/ '^TestObjectsPersistVerify$' -e OBJ_PERSIST_IDS="$ids" || { ok=0; tail -20 "$WORK/obj-persist-verify.log"; }
  for out in "$WORK/obj-persist-seed.log" "$WORK/obj-persist-verify.log"; do
    grep -qE -- '^--- PASS: TestObjectsPersist(Seed|Verify) ' "$out" && ! grep -qE -- '--- (SKIP|FAIL)' "$out" || ok=0
  done
  [[ $ok == 1 ]] && record B-OBJ-MIGRATION PASS "TestObjectsMigration: 5/5 subtests pass, none skipped (MG-OBJ-1 empty to version 7, MG-OBJ-2 a populated G008 world from 6 to 7 with every row of all ten tables identical and a repeat up, MG-OBJ-3 down refused while a row exists with the version, rows and catalog unchanged, MG-OBJ-4 down and up on the empty table restore the exact version-6 and version-7 catalogs (and reapply), MG-OBJ-5 a rolled-back up restores the owner and the version-6 catalog); a guarded object written before a PostgreSQL and MinIO stop/start is returned with identical bytes by a guarded Get after it; 00001-00006 byte-identical (B-MIGRATE-SET)" || record B-OBJ-MIGRATION FAIL "stored-object migrations ($n/5 subtests; persistence)"
}

# B-OBJ-IT: TestObjects (the oracle's 382 cases in the service and RLS matrices, round trips, scope discovery, schema,
# deletion, signing, administration denials and log checks) and TestObjectsFootprint (E-OBJ-1..4), OBJ-SIGN-HOST, and
# the record checks. E-OBJ lines are kept as obj-eobj.jsonl.
obj_it() { # obj_it DEST PROJECT IMAGE
  local M="$1" project="$2" img="$3" ok=1 log="$WORK/obj-it.log" alog="$WORK/obj-it-attack.log" n f svc rls an iso lc host triggers
  it_fresh "$M" "$project" obj-it "$log" && rls_go obj-it "$project" "$img" "$OBJ_ENV" "$log" ./internal/tenancy/access/ '^(TestObjects|TestObjectsFootprint)$' || ok=0
  # G010: the isolation attack suite, after TestObjects (its own world and log); pinned subtests and oracle line.
  it_fresh "$M" "$project" obj-it-attack "$alog" && rls_go obj-it-attack "$project" "$img" "$OBJ_ENV" "$alog" ./internal/tenancy/access/ '^TestIsolationAttacks$' || ok=0
  an="$(it_count "$alog" TestIsolationAttacks)"; iso="$(sed -n 's/^ISO-ORACLE //p' "$alog" | head -1)"
  [[ "$an" == 24 && "$iso" == '{"cases":163,"disagreements":0}' ]] || { ok=0; tail -60 "$alog"; }
  n="$(it_count "$log" TestObjects)"
  grep -q -- '^--- PASS: TestObjectsFootprint ' "$log" && f=pass || f=FAIL
  svc="$(sed -n 's/^OBJ-ORACLE-SERVICE //p' "$log" | head -1)"; rls="$(sed -n 's/^OBJ-ORACLE-RLS //p' "$log" | head -1)"
  [[ "$n" == 10 && "$f" == pass && "$svc" == '{"cases":297,"disagreements":0}' && "$rls" == '{"cases":85,"disagreements":0}' ]] || { ok=0; tail -80 "$log"; }
  grep -hE '^(E-OBJ-(1|2|3|4|TRIGGER)|OBJ-SCOPE|OBJ-GUARD-TRANSITIONS|OBJ-ORACLE-(SERVICE|RLS)) ' "$log" >"$WORK/obj-eobj.jsonl" || true
  if [[ -n "$REPORT_DIR" ]]; then cp "$WORK/obj-eobj.jsonl" "$REPORT_DIR/" 2>/dev/null || true; fi
  triggers="$(grep -c '^E-OBJ-TRIGGER ' "$WORK/obj-eobj.jsonl" || true)"
  [[ "$triggers" == 0 ]] || log "E-OBJ: $triggers escalation trigger line(s): return to Codex (obj-eobj.jsonl)"
  # G064: TestObjectsSignHost is an independent invocation on its own fresh copy (without a host probe nothing runs: UNMET).
  if [[ "$PROBE_MODE" != none ]] && ! it_fresh "$M" "$project" obj-sign-host "$WORK/obj-sign-host-it.log"; then host="fresh "
  else host="$(obj_sign_host "$project" "$img" "$(env_get "$M" MINIO_API_PORT)")"; fi
  [[ -z "$host" ]] || { ok=0; log "OBJ-SIGN-HOST: $host"; }
  lc="$(rls_logcheck obj-it "$M" "$log" "$alog")"
  [[ -z "$lc" ]] || ok=0
  if [[ "$host" == UNMET ]]; then record B-OBJ-IT UNMET "OBJ-SIGN-HOST needs a host probe (PROBE_MODE=none)"; return; fi
  [[ $ok == 1 ]] && record B-OBJ-IT PASS "TestObjects against real PostgreSQL at version 10 (the attested runtime login) and the private MinIO (the scoped credential, no root credential in the test environment): 10/10 subtests pass (OBJ-ATTEST incl. the privileged-assertion negative; OBJ-READ/OBJ-WRITE: the independent oracle's 297 service cases agree exactly, every denial with zero provider calls, unchanged database and provider state and no disclosure; OBJ-RLS: its 85 runtime and FORCE-owner cases agree exactly (RU05 per the 428C9 ruling), the guard refuses 7 admitted illegal transitions (non-creator finalize, ready-to-ready, any update of a deleted row) with stored_objects_transition and the row unchanged beside 2 positive controls, plus a poisoned pooled connection refused and discarded; OBJ-RT round trips incl. zero bytes and exactly the maximum; OBJ-SCOPE identical statements and fact rows for 1 and 50 Group bindings, only the owning Group read; OBJ-SCHEMA; OBJ-DEL real removal, tombstone and idempotent repeat; OBJ-SIGN real fetches, a real 2 s expiry, alteration and anonymous 403s and the loopback Host binding; OBJ-ADMIN; OBJ-LOG); TestObjectsFootprint (E-OBJ-1..4 recorded, $triggers trigger lines); TestIsolationAttacks (G010 attack chains A-E across context, facade, runtime SQL and provider; G018 B2 and C3; G019 chain F; G020 step G1 cross-tenant and cross-Group label reads, counts, joins and refused writes): 24/24 subtests pass, the independent attack oracle's 163 expectations agree exactly, every denial with its paired control, zero provider calls and unchanged state, the staged E1/E3 effects, objects:/rls: citations resolve; OBJ-SIGN-HOST: the unchanged URL signed for 127.0.0.1:$(env_get "$M" MINIO_API_PORT) fetched through the published port ($PROBE_MODE) returns 200 with the expected digest and the attachment, content-type and no-store overrides, and the bare path 403; storage.operation records catalogued, no canary in them or the PostgreSQL log, planted canary detected" || record B-OBJ-IT FAIL "stored-object isolation ($n/10 subtests, footprint $f; service ${svc:-none}; rls ${rls:-none}; attack $an/24, oracle ${iso:-none}; sign-host ${host:-ok}; logcheck: ${lc:-ok}; failed-subtests: $(obj_fail_codes "$log" "$n"))"
}

# B-OBJ-ORDERING: TestObjectsOrdering (OOR-1..8), then the log checks.
obj_ordering() { # obj_ordering DEST PROJECT IMAGE
  local M="$1" project="$2" img="$3" ok=1 log="$WORK/obj-ordering-it.log" n lc
  it_fresh "$M" "$project" obj-ordering "$log" && rls_go obj-ordering "$project" "$img" "$OBJ_ENV" "$log" ./internal/tenancy/access/ '^TestObjectsOrdering$' || ok=0
  n="$(it_count "$log" TestObjectsOrdering)"
  [[ "$n" == 8 ]] || { ok=0; tail -60 "$log"; }
  grep -E '^ORDERING ' "$log" >"$WORK/obj-ordering.jsonl" || true
  if [[ -n "$REPORT_DIR" ]]; then cp "$WORK/obj-ordering.jsonl" "$REPORT_DIR/" 2>/dev/null || true; fi
  lc="$(rls_logcheck obj-ordering "$M" "$log")"
  [[ -z "$lc" ]] || ok=0
  [[ $ok == 1 ]] && record B-OBJ-ORDERING PASS "TestObjectsOrdering against real PostgreSQL and MinIO: 8/8 subtests pass, none skipped (OOR-1 a Group removal before the publish: denied, compensated, never readable; OOR-2 a removal racing the publish waits for its facts locks, then reads are denied; OOR-3 Delete versus the publish in both orders: a published object is then deleted, a logically deleted pending object is never published and is compensated; OOR-4 the admin-to-member downgrade; OOR-5 a session revoked between insert and publish; OOR-6 a disablement during an in-flight Get; OOR-7 an issued URL until expiry while new issuance is denied; OOR-8 a delete during an open Get stream: no new read, the bounded residual recorded); storage.operation records catalogued, no canary in them or the PostgreSQL log, planted canary detected" || record B-OBJ-ORDERING FAIL "stored-object ordering ($n/8 subtests; logcheck: ${lc:-ok})"
}

# B-OBJ-FAULTS: TestObjectsFaults (every Put and Delete transition under injected and real faults), then the log checks.
obj_faults() { # obj_faults DEST PROJECT IMAGE
  local M="$1" project="$2" img="$3" ok=1 log="$WORK/obj-faults-it.log" n lc
  it_fresh "$M" "$project" obj-faults "$log" && rls_go obj-faults "$project" "$img" "$OBJ_ENV" "$log" ./internal/tenancy/access/ '^TestObjectsFaults$' || ok=0
  n="$(it_count "$log" TestObjectsFaults)"
  [[ "$n" == 19 ]] || { ok=0; tail -60 "$log"; }
  grep -E '^ORDERING ' "$log" >"$WORK/obj-faults.jsonl" || true
  if [[ -n "$REPORT_DIR" ]]; then cp "$WORK/obj-faults.jsonl" "$REPORT_DIR/" 2>/dev/null || true; fi
  lc="$(rls_logcheck obj-faults "$M" "$log")"
  [[ -z "$lc" ]] || ok=0
  [[ $ok == 1 ]] && record B-OBJ-FAULTS PASS "TestObjectsFaults against real PostgreSQL and MinIO: 19/19 subtests pass, none skipped (T1 statement error, acknowledged-then-unknown COMMIT and a real FATAL before COMMIT; P2 provider error, a lost reply after the bytes landed, short and long sources incl. a provider that completes anyway, a blocked first Read and a blocked lookahead within the transfer deadline, zero length; V binding mismatch and unavailability; T2 definitive abort, unknown-but-committed (reconciled ready, bytes kept), unknown with reconcile unavailable (bytes kept, no compensation), unknown-not-committed (compensated); cancellation at each stage; a simulated interruption after each stage repaired by Delete; a compensation failure leaving inaccessible bytes (X3); D1 unknown, D2, D3 and D4 failures as deletion_pending completed by a retry; Get unavailable, integrity mismatch at EOF and missing bytes; pending and deleting objects never served; FX-STALL a frozen database path ends within the decision and cleanup budgets with the connection destroyed); storage.operation records catalogued, no canary in them or the PostgreSQL log, planted canary detected" || record B-OBJ-FAULTS FAIL "stored-object faults ($n/19 subtests; logcheck: ${lc:-ok})"
}

# G011 B-AUDIT-CATALOG, B-AUDIT-MIGRATION and B-AUDIT-IT (Plan r2 §9, §10). Test containers use the stored-object env
# (the scoped MinIO credential for TestAudit's storage producers) and return their records the RLS way.

# B-AUDIT-CATALOG: TestAuditCatalog (CT-AUD-1..6 and the rolled-back detectors N1-N8).
audit_catalog() { # audit_catalog DEST PROJECT IMAGE
  local M="$1" project="$2" img="$3" ok=1 log="$WORK/audit-catalog-it.log" n doc mig trig
  it_fresh "$M" "$project" audit-catalog "$log" && rls_go audit-catalog "$project" "$img" "$OBJ_ENV" "$log" ./internal/migrate/ '^TestAuditCatalog$' || ok=0
  n="$(it_count "$log" TestAuditCatalog)"
  [[ "$n" == 7 ]] || { ok=0; tail -40 "$log"; }
  # U-AUDIT's documentation half (the core check image holds no docs): AUDIT.md's catalog rows name exactly the 52
  # actions of the current action constraint (00020's up section, G020, which adds the five trigger-owned label.*
  # actions; 00019, G063, added the trigger-owned 47th, mcp.connection_renamed; 00018, G062, added no action; 00017,
  # G060, added the 46th), T-rows exactly its 45 trigger-owned actions.
  doc="$(sed -nE 's/^\| [TG][0-9]+ \| `([a-z_.]+)` \|.*/\1/p' "$M/docs/foundation/AUDIT.md" | LC_ALL=C sort | paste -sd' ')"
  trig="$(sed -nE 's/^\| T[0-9]+ \| `([a-z_.]+)` \|.*/\1/p' "$M/docs/foundation/AUDIT.md" | LC_ALL=C sort | paste -sd' ')"
  mig="$(sed -n '/ADD CONSTRAINT events_action_valid/,/revocation\.unaudited/{p;/revocation\.unaudited/q}' "$M/core/migrations/00020_labels.sql" | grep -oE "'[a-z_]+\.[a-z_]+'" | tr -d "'" | LC_ALL=C sort | paste -sd' ')"
  [[ -n "$mig" && "$doc" == "$mig" && "$(wc -w <<<"$doc")" == 52 && "$(wc -w <<<"$trig")" == 45 ]] || { ok=0; log "AUDIT.md catalog differs from the migration: doc [$doc] trigger [$trig] migration [$mig]"; }
  [[ $ok == 1 ]] && record B-AUDIT-CATALOG PASS "AUDIT.md's catalog table names exactly the current migration's 52 actions (45 trigger-owned; G014 adds two, G016 four, G017 one, G019 three trigger-owned invitation.*, G058 six trigger-owned enrollment.policy_changed, enrollment.readiness_changed, enrollment.restriction_backfilled, enrollment.requested, enrollment.decided and enrollment.revoked, G060 the core action access.override_used, G063 the trigger-owned mcp.connection_renamed, G020 the five trigger-owned label.created, label.updated, label.deleted, label.assigned and label.unassigned); TestAuditCatalog against real PostgreSQL at version 8: 7/7 subtests pass, none skipped (CT-AUD-1 audit.events: 19 columns, 17 constraints, the primary key and the partial tenant-time index, no foreign key, schema app unchanged with 38 relations; CT-AUD-2 ENABLE+FORCE and the two pinned policy expressions; CT-AUD-3 every table, column, function and schema ACL: runtime USAGE, SELECT and 12 INSERT columns only, retention nothing, the version-7 app ACLs unchanged; CT-AUD-4 the 18 capture and 3 audit.events trigger definitions, the 11 invoker functions with their pinned bodies, no row-dump construct, exception blocks only in the two revocation-priority branches; CT-AUD-5 runtime and retention UPDATE, DELETE, TRUNCATE, DDL, trigger disable, row_security off, actor/id inserts and GRANT refused with rows and catalog unchanged; CT-AUD-6 owner FORCE: context rows only, UPDATE/DELETE affect 0 rows, TRUNCATE refused; CT-AUD-NEG N1-N8 (NO FORCE, an UPDATE grant, an actor-column INSERT grant, a dropped refusal trigger, a disabled capture trigger, a permissive probe policy, PUBLIC schema USAGE, a SECURITY DEFINER capture function) each yield exactly their detector code, and the catalog is unchanged after)" || record B-AUDIT-CATALOG FAIL "audit catalog ($n/7 subtests)"
}

# B-AUDIT-MIGRATION: TestAuditMigration (MG-AUD-1..6), then persistence across a PostgreSQL restart (the seed's
# identifiers pass from its stdout to the verify step's environment).
audit_migration() { # audit_migration DEST PROJECT IMAGE
  local M="$1" project="$2" img="$3" ok=1 log="$WORK/audit-migration-it.log" n ids out
  it_fresh "$M" "$project" audit-migration "$log" && rls_go audit-migration "$project" "$img" "$OBJ_ENV" "$log" ./internal/migrate/ '^TestAuditMigration$' || ok=0
  n="$(it_count "$log" TestAuditMigration)"
  [[ "$n" == 6 ]] || { ok=0; tail -40 "$log"; }
  it_fresh "$M" "$project" audit-persist-seed "$WORK/audit-persist-seed.log" && rls_go audit-persist-seed "$project" "$img" "$OBJ_ENV" "$WORK/audit-persist-seed.log" ./internal/migrate/ '^TestAuditPersistSeed$' || { ok=0; tail -20 "$WORK/audit-persist-seed.log"; }
  ids="$(sed -n 's/^AUDIT-PERSIST //p' "$WORK/audit-persist-seed.log" | head -1)"
  compose "$M" stop postgres >"$WORK/audit-persist-restart.log" 2>&1 || ok=0
  compose "$M" up -d --wait postgres >>"$WORK/audit-persist-restart.log" 2>&1 || { ok=0; tail -20 "$WORK/audit-persist-restart.log"; }
  rls_go audit-persist-verify "$project" "$img" "$OBJ_ENV" "$WORK/audit-persist-verify.log" ./internal/migrate/ '^TestAuditPersistVerify$' -e AUDIT_PERSIST_IDS="$ids" || { ok=0; tail -20 "$WORK/audit-persist-verify.log"; }
  for out in "$WORK/audit-persist-seed.log" "$WORK/audit-persist-verify.log"; do
    grep -qE -- '^--- PASS: TestAuditPersist(Seed|Verify) ' "$out" && ! grep -qE -- '--- (SKIP|FAIL)' "$out" || ok=0
  done
  [[ $ok == 1 ]] && record B-AUDIT-MIGRATION PASS "TestAuditMigration: 6/6 subtests pass, none skipped (MG-AUD-1 empty to version 8; MG-AUD-2 a populated version-7 world to 8 with every app row and ACL identical, no backfilled history, and a repeat up; MG-AUD-3 a stray object, a PUBLIC grant, a missing schema and a wrong owner of schema audit each refused with the version, rows and catalog unchanged; MG-AUD-4 down refused while history exists, nothing changed; MG-AUD-5 down and up on an empty history restore the exact version-7 and version-8 catalogs (and reapply); MG-AUD-6 a rolled-back up restores the owner and the version-7 catalog); history written before a PostgreSQL stop/start is unchanged after it and readable by the Organization owner; 00001-00007 byte-identical (B-MIGRATE-SET)" || record B-AUDIT-MIGRATION FAIL "audit migrations ($n/6 subtests; persistence)"
}

# B-AUDIT-IT: TestAudit (every producer of the 52-action catalog for real (G020: the five label.* actions from real label
# operations, AT-LAB; G063: mcp.connection_renamed from a real connection rename; G060: access.override_used; G014: 29 plus the two MCP actions; G016: four more; G017: organization.branding_changed; G019: invitation.issued, invitation.revoked, invitation.consumed; G058: enrollment.policy_changed through the owner's guarded policy write, enrollment.readiness_changed and enrollment.restriction_backfilled through the superuser operator procedures; G058 request slice: enrollment.requested by the applicant's own request, enrollment.decided and enrollment.revoked by the owner's approval and revocation), the AUDIT-INV pin, E-AUDIT-1/2 kept as
# audit-eaudit.jsonl) and the record checks.
audit_it() { # audit_it DEST PROJECT IMAGE
  local M="$1" project="$2" img="$3" ok=1 log="$WORK/audit-it.log" n inv lc triggers
  it_fresh "$M" "$project" audit-it "$log" && rls_go audit-it "$project" "$img" "$OBJ_ENV" "$log" ./internal/tenancy/access/ '^TestAudit$' || ok=0
  n="$(it_count "$log" TestAudit)"
  inv="$(sed -n 's/^AUDIT-INV //p' "$log" | head -1)"
  [[ "$n" == 10 && "$inv" == '{"actions":52,"observed":52}' ]] || { ok=0; tail -60 "$log"; }
  grep -E '^E-AUDIT-(1|2|TRIGGER) ' "$log" >"$WORK/audit-eaudit.jsonl" || true
  if [[ -n "$REPORT_DIR" ]]; then cp "$WORK/audit-eaudit.jsonl" "$REPORT_DIR/" 2>/dev/null || true; fi
  triggers="$(grep -c '^E-AUDIT-TRIGGER ' "$log" || true)"
  [[ "$triggers" == 0 ]] || log "E-AUDIT: $triggers escalation trigger line(s): return to Codex (audit-eaudit.jsonl)"
  lc="$(rls_logcheck audit-it "$M" "$log")"
  [[ -z "$lc" ]] || ok=0
  [[ $ok == 1 ]] && record B-AUDIT-IT PASS "TestAudit against real PostgreSQL at version 22 (the attested runtime login; the canonical ACL comparison of amendment 01), the fake WorkOS through the real G004 services and the private MinIO (the scoped credential): 10/10 subtests pass (AT-ATTEST; AT-INV all 52 catalog actions observed from real calls (G020: label.created, label.updated, label.deleted, label.assigned and label.unassigned from real label operations; G063: mcp.connection_renamed from a real RenameMCPConnection; G060: access.override_used from an override-based Group administration write; G058: the owner's enrollment policy write and the operator readiness close, backfill and attestation, and the applicant's request with the owner's approval and revocation, each row's actor, Organization, subject, resource and reason asserted), AUDIT-INV $inv; AT-FMT exact fields for user, system and unknown events, forged actor, foreign or unestablished Organization, foreign and nonexistent Group, trigger-owned and malformed appends refused with their detector codes, malformed correlation stored as NULL, the Group admin's signed URL of a Group object carries that Group (G060: an Organization admin with no binding in the Group is refused not_found and no url row is written); AT-ATOM same-transaction rename and sign-in (both branches), a planted fault fails rename, sign-in, rotation, enable and grant with no effect (G063: a connection rename under a planted fault leaves the name, the row and every security fact unchanged with no event, and after the fault its event shares the row's transaction), revocation priority for logout current/all/provider and disable under a revocation fault (revocation.unaudited recorded) and under every-insert and column-privilege faults (audit.gap failed, grants restored and re-verified), a stalled recorder rollback bounded with its connection destroyed, a recording fault leaves the denial unchanged plus one audit.record_failed; AT-ID sign-in, rotation, logout fan-out over expired-but-unrevoked sessions, provider logout confirmed and failed, system disable/enable, attributed and unattributed post-consume failures, none pre-consume, global rows never tenant-visible, a frozen-database revocation never success, a lost COMMIT reply at most one event; AT-OBJ put, ready, signed URL with TTL, delete and tombstone, no URL, signature or key in any row, an issuance fault returns no URL, provider faults record upload_failed and deletion_pending, T2 lost reply reconciled; AT-READ owner-only reads, every other role forbidden with zero SQL rows, multi-tenant and selected-Group cases, cross-tenant cursor refused, pooled reuse, downgrade and revocation; AT-LIM page and window bounds, cursor binding, strictly decreasing continuation under concurrent appends, E-AUDIT-1/2 ($triggers trigger lines); AT-LAB label producers: attribution, reasons, no-op silence, one-transaction deletion and audit faults; AT-PRIV no canary in any row or read, a planted canary detected, no deletion path); access.decision, storage.operation, audit.record_failed and audit.gap records catalogued, no canary in them or the PostgreSQL log, planted canary detected" || record B-AUDIT-IT FAIL "audit producers ($n/10 subtests; inventory ${inv:-none}; logcheck: ${lc:-ok})"
}

# ===================================================================================
# G013 B-RETENTION-CATALOG, B-RETENTION-IT, B-RECOVERY-DRILL and B-MAINTENANCE-DRILL (Plan r3 §6, §10, §11.4-§11.6).

# B-RETENTION-CATALOG: TestRetentionCatalog (CT-RET-1, CT-RET-2, CT-RET-NEG) and TestRetentionMigration (MG-RET-1..5).
retention_catalog() { # retention_catalog DEST PROJECT IMAGE
  local M="$1" project="$2" img="$3" ok=1 log="$WORK/retention-catalog-it.log" n m
  it_fresh "$M" "$project" retention-catalog "$log" && rls_go retention-catalog "$project" "$img" "$OBJ_ENV" "$log" ./internal/migrate/ '^(TestRetentionCatalog|TestRetentionMigration)$' || ok=0
  n="$(it_count "$log" TestRetentionCatalog)"; m="$(it_count "$log" TestRetentionMigration)"
  [[ "$n" == 3 && "$m" == 5 ]] || { ok=0; tail -40 "$log"; }
  [[ $ok == 1 ]] && record B-RETENTION-CATALOG PASS "TestRetentionCatalog against real PostgreSQL at version 9: 3/3 subtests (CT-RET-1 the exact version-9 delta on the unchanged version-8 catalog: four retention-only policies with pinned expressions, the minimal retention column SELECT and DELETE grants, the owner-owned invoker app.retention_eligible with its pinned body and EXECUTE for retention only, the role-bound audit.events_refuse body digest, schema audit rn+rt USAGE, and the version-8 detector reporting exactly the delta parts; CT-RET-2 retention attributes and memberships, no INSERT/UPDATE/TRUNCATE/REFERENCES/TRIGGER anywhere, exactly seven readable columns, runtime without EXECUTE or DELETE; CT-RET-NEG eight rolled-back detectors each yield exactly their code); TestRetentionMigration: 5/5 (MG-RET-1 empty to 9; MG-RET-2 a populated version-8 world to 9 with every row identical and a repeat up; MG-RET-3 down to 8 with rows present restores the exact version-8 catalog; MG-RET-4 up and reapply; MG-RET-5 a stray grant fails the up, which rolls back to version 8 with the owner restored)" || record B-RETENTION-CATALOG FAIL "retention catalog ($n/3, migration $m/5 subtests)"
}

# B-RETENTION-IT: TestRetention (RT-1..9) with the actual retention login; E-RET lines kept as retention-eret.jsonl.
retention_it() { # retention_it DEST PROJECT IMAGE
  local M="$1" project="$2" img="$3" ok=1 log="$WORK/retention-it.log" n triggers
  it_fresh "$M" "$project" retention-it "$log" && rls_go retention-it "$project" "$img" "$OBJ_ENV" "$log" ./internal/tenancy/access/ '^TestRetention$' || ok=0
  n="$(it_count "$log" TestRetention)"
  [[ "$n" == 9 ]] || { ok=0; tail -60 "$log"; }
  grep -E '^E-RET-(1|1-COUNT|2|2-COUNT|3|3-CANDIDATES|TRIGGER) ' "$log" >"$WORK/retention-eret.jsonl" || true
  [[ "$(grep -cE '^E-RET-(1|1-COUNT|2|2-COUNT|3|3-CANDIDATES) ' "$WORK/retention-eret.jsonl")" == 6 ]] || ok=0
  if [[ -n "$REPORT_DIR" ]]; then cp "$WORK/retention-eret.jsonl" "$REPORT_DIR/" 2>/dev/null || true; fi
  triggers="$(grep -c '^E-RET-TRIGGER ' "$WORK/retention-eret.jsonl" || true)"
  [[ "$triggers" == 0 ]] || log "E-RET: $triggers escalation trigger line(s): return to Codex (retention-eret.jsonl; PG-005 stop rule)"
  [[ $ok == 1 ]] && record B-RETENTION-IT PASS "TestRetention against real PostgreSQL at version 10 with the actual retention login (attested: not superuser, no BYPASSRLS, a member of nothing): 9/9 subtests (RT-1 absent settings raise 22023 on SELECT and DELETE, also after a reset; RT-2 malformed scope, cutoff and holds raise, 256 holds accepted; RT-3 Organization scope deletes exactly the eligible set, held, recent, foreign and global rows remain; RT-4 global scope deletes only eligible global rows in bounded batches; RT-5 UPDATE, TRUNCATE, INSERT, ALTER, CREATE, SET ROLE, trigger disable, row_security off and other tables refused, GRANT grants nothing, catalog and rows unchanged; RT-6 runtime refused, owner affects 0 rows, owner TRUNCATE and a superuser DELETE with an eligible declaration raise append-only; RT-7 only the eligible deleted tombstones are visible and deletable; RT-8 a hold added after the count preserves the row and the tombstone, deleted_by unreadable; RT-9 E-RET-1..3 on the pinned distribution, $triggers trigger lines)" || record B-RETENTION-IT FAIL "retention boundary ($n/9 subtests)"
}

# ===================================================================================
# G014 B-MCP-CATALOG and B-MCP-IT (Plan r8 §10.3, §10.4, §11.6), in the migrations suite after the retention checks and
# before the recovery drill (which resets the database): TestMCPRecovery applies the actual restore invalidation, which
# expires every session, so it runs last. The core records of the MCP tests (MCP_LOG_FILE) return with the access
# records the RLS way and pass the same catalog and canary checks.

# B-MCP-CATALOG: TestMCPCatalog (CT-MCP-1, CT-MCP-NEG) and TestMCPMigration (MG-MCP-1..5); G063 (Plan r1 §6) adds
# TestMCPNamesCatalog (CT-MCPN-1, -CHECK, -DML, -NEG) and TestMCPNamesMigration (MG-MCPN-1..4).
mcp_catalog() { # mcp_catalog DEST PROJECT IMAGE
  local M="$1" project="$2" img="$3" ok=1 log="$WORK/mcp-catalog-it.log" n m nn nm
  it_fresh "$M" "$project" mcp-catalog "$log" && rls_go mcp-catalog "$project" "$img" "$OBJ_ENV" "$log" ./internal/migrate/ '^(TestMCPCatalog|TestMCPMigration|TestMCPNamesCatalog|TestMCPNamesMigration)$' || ok=0
  n="$(it_count "$log" TestMCPCatalog)"; m="$(it_count "$log" TestMCPMigration)"
  nn="$(it_count "$log" TestMCPNamesCatalog)"; nm="$(it_count "$log" TestMCPNamesMigration)"
  [[ "$n" == 3 && "$m" == 5 && "$nn" == 4 && "$nm" == 4 ]] || { ok=0; tail -40 "$log"; }
  [[ $ok == 1 ]] && record B-MCP-CATALOG PASS "TestMCPCatalog against real PostgreSQL at version 10: 3/3 subtests (CT-MCP-1 the exact version-10 delta on the unchanged version-9 catalog: sessions.kind with its CHECK, app.mcp_connections with its constraints, guard trigger and audit capture triggers, the widened audit CHECKs and attribution body, E1 replaced as the auth reader with the browser kind filter, the minimal runtime grants, the three organization-bound policies and the altered sessions_insert, and the version-9 detector reporting exactly the delta parts; CT-MCP-NEG ten rolled-back detectors (runtime DELETE or immutable-column UPDATE, policies without a non-null rls_org_role() or without the organization predicate, sessions_insert admitting browser rows in a tenant shape, E1 without the kind filter, a missing guard trigger, a second activation, a definer reader on the table) each yield exactly their code; CT-MCP-GUARD the selection guard as the runtime role: a foreign or invisible target, a previous selection no longer visible and a selected-Group-shaped write refused, a clear after the loss allowed, and a guard without the previous-selection clause admits that switch (negative control)); TestMCPMigration: 5/5 (MG-MCP-1 empty to 10; MG-MCP-2 a populated version-9 world to 10 with every row identical; MG-MCP-3 down refused with MCP state, allowed when empty, restoring the exact version-9 catalog; MG-MCP-4 up again and reapply; MG-MCP-5 a rolled-back up leaves version 9); G063 TestMCPNamesCatalog at version 19: 4/4 (CT-MCPN-1 the exact version-19 delta over the version-18 catalog: the nullable display_name column with its CHECK, the two runtime column grants, the name capture trigger, the two replaced audit bodies and the widened action CHECK, policies, guard and FORCE unchanged; CT-MCPN-CHECK the stored-shape CHECK and the strict name capture; CT-MCPN-DML the runtime name path writes the name only, on its own row, with a current role; CT-MCPN-NEG rolled-back missing-protection detectors); TestMCPNamesMigration: 4/4 (MG-MCPN-1 a populated version-18 world to 19, every row unnamed and otherwise identical; MG-MCPN-2 down refused with a named row or a rename event; MG-MCPN-3 down and up on an unnamed world restore the exact 18 and 19 catalogs; MG-MCPN-4 a rolled-back up leaves version 18 and the owner)" || record B-MCP-CATALOG FAIL "MCP catalog ($n/3, migration $m/5, names $nn/4, names migration $nm/4 subtests)"
}

# G016 B-DOMAINS-CATALOG and B-DOMAINS-IT (Plan r4 §6.2, §6.3, §11.2), in the migrations suite after the MCP checks and
# before the recovery drill. B-DOMAINS-CATALOG: TestDomainsCatalog (CT-DOM-1 the exact version-11 delta, CT-DOM-NEG,
# CT-DOM-UPDOWN, CT-DOM-12 legacy sessions across the version-12 transition). B-DOMAINS-IT: TestDomains (DOM-1..12 on the attested runtime login; DOM-10 prints the footprint
# observations E-DOM-F1..F4, kept as domains-edom.jsonl; an E-DOM-TRIGGER line returns to Codex; G057 adds DOM-11 and
# DOM-12, the platform-label cases, Plan r5 §5.2 T5), then the record checks.
domains_checks() { # domains_checks DEST PROJECT IMAGE
  local M="$1" project="$2" img="$3" ok=1 log="$WORK/domains-catalog-it.log" dlog="$WORK/domains-it.log" n f lc triggers
  it_fresh "$M" "$project" domains-catalog "$log" && rls_go domains-catalog "$project" "$img" "$OBJ_ENV" "$log" ./internal/migrate/ '^TestDomainsCatalog$' || ok=0
  n="$(it_count "$log" TestDomainsCatalog)"
  [[ "$n" == 4 ]] || { ok=0; tail -40 "$log"; }
  [[ $ok == 1 ]] && record B-DOMAINS-CATALOG PASS "TestDomainsCatalog against real PostgreSQL at versions 11 and 12: 4/4 subtests (CT-DOM-1 the exact version-11 delta on the unchanged version-10 catalog: app.organizations.slug with its CHECKs and partial unique index, app.organization_domains with its constraints, live-hostname and Organization indexes, guard, slug guard and audit capture triggers, the two auth-reader definer functions, policies and exact grants; CT-DOM-NEG rolled-back missing-protection negatives; CT-DOM-UPDOWN down refused with domain state, the exact version-10 catalog when empty, up again; CT-DOM-12 legacy sessions across the version-12 transition)" || record B-DOMAINS-CATALOG FAIL "domain catalog ($n/4 subtests)"
  ok=1
  it_fresh "$M" "$project" domains-it "$dlog" && rls_go domains-it "$project" "$img" "$OBJ_ENV" "$dlog" ./internal/tenancy/access/ '^TestDomains$' || ok=0
  n="$(it_count "$dlog" TestDomains)"
  [[ "$n" == 12 ]] || { ok=0; tail -60 "$dlog"; }
  grep -E '^E-DOM-(F[1-4]|TRIGGER) ' "$dlog" >"$WORK/domains-edom.jsonl" || true
  f="$(grep -cE '^E-DOM-F[1-4] ' "$WORK/domains-edom.jsonl" || true)"
  [[ "$f" == 4 ]] || ok=0
  if [[ -n "$REPORT_DIR" ]]; then cp "$WORK/domains-edom.jsonl" "$REPORT_DIR/" 2>/dev/null || true; fi
  triggers="$(grep -c '^E-DOM-TRIGGER ' "$WORK/domains-edom.jsonl" || true)"
  [[ "$triggers" == 0 ]] || log "E-DOM: $triggers escalation trigger line(s): return to Codex (domains-edom.jsonl)"
  lc="$(rls_logcheck domains-it "$M" "$dlog")"
  [[ -z "$lc" ]] || ok=0
  [[ $ok == 1 ]] && record B-DOMAINS-IT PASS "TestDomains against real PostgreSQL at version 11 on the attested runtime login: 12/12 subtests (DOM-1 operation x role matrix with zero effects after the RLS attestation; DOM-2 concurrent duplicate claims and the claim fence; DOM-3 concurrent slug updates, UUIDs stable; DOM-4 races, late and stale adapter results, the replacement claim and the withdrawal retry; DOM-5 reassignment needs fresh proof; DOM-6 freshness, stored TLS expiry and confirmed proof loss fail closed; DOM-7 the resolver per state through the auth reader; DOM-8 exact audit rows per operation; DOM-9 restore and configuration mismatch; DOM-10 footprint, $f E-DOM-F lines recorded, $triggers trigger lines, no performance guarantee; G057 DOM-11 a slug change keeps the platform label, its routing and the bound sessions, no role can write the label; DOM-12 the platform label routes to the exact UUID, the slug, unknown, reserved and suffix-confusion hosts route nowhere); access records catalogued, no canary in them or the PostgreSQL log, planted canary detected" || record B-DOMAINS-IT FAIL "domain lifecycle ($n/12 subtests, footprint $f; logcheck: ${lc:-ok})"
}

# G017 B-BRANDING-CATALOG and B-BRANDING-IT (Plan r2 §6.2, §6.3), in the migrations suite after the domain checks and
# before the recovery drill. B-BRANDING-CATALOG: TestBrandingCatalog (CT-BRD-1 the exact version-13 delta, CT-BRD-NEG,
# CT-BRD-UPDOWN). B-BRANDING-IT: TestBranding (BRD-1..14 with 9 nested cases, the logo cases on the object
# store) on the attested runtime login, then the record checks.
branding_checks() { # branding_checks DEST PROJECT IMAGE
  local M="$1" project="$2" img="$3" ok=1 log="$WORK/branding-catalog-it.log" blog="$WORK/branding-it.log" n lc
  it_fresh "$M" "$project" branding-catalog "$log" && rls_go branding-catalog "$project" "$img" "$OBJ_ENV" "$log" ./internal/migrate/ '^TestBrandingCatalog$' || ok=0
  n="$(it_count "$log" TestBrandingCatalog)"
  [[ "$n" == 3 ]] || { ok=0; tail -40 "$log"; }
  [[ $ok == 1 ]] && record B-BRANDING-CATALOG PASS "TestBrandingCatalog against real PostgreSQL at version 13: 3/3 subtests (CT-BRD-1 the exact version-13 delta on the unchanged version-12 catalog: app.organization_branding with its constraints, validator, guard and audit capture triggers, runtime and auth-reader policies, the definer projection function and exact grants; CT-BRD-NEG rolled-back missing-protection negatives; CT-BRD-UPDOWN the exact version-12 catalog when empty, up again, down refused with branding state or history)" || record B-BRANDING-CATALOG FAIL "branding catalog ($n/3 subtests)"
  ok=1
  it_fresh "$M" "$project" branding-it "$blog" && rls_go branding-it "$project" "$img" "$OBJ_ENV" "$blog" ./internal/tenancy/access/ '^TestBranding$' || ok=0
  n="$(it_count "$blog" TestBranding)"
  [[ "$n" == 23 ]] || { ok=0; tail -60 "$blog"; }
  lc="$(rls_logcheck branding-it "$M" "$blog")"
  [[ -z "$lc" ]] || ok=0
  [[ $ok == 1 ]] && record B-BRANDING-IT PASS "TestBranding against real PostgreSQL at version 13 with the object store on the attested runtime login: 23/23 subtest results, the 14 BRD cases and their 9 nested cases (BRD-1..4 roles and invalid input with zero effects; BRD-5..7 stale versions, concurrency, reset and audit; BRD-8..9 logo normalization and rejections; BRD-10 upload-association uncertainty and recovery; BRD-11 public projection and logo visibility; BRD-12 statement counts, no performance guarantee; BRD-13 session membership; BRD-14 current-logo delete and purge); access records catalogued, no canary in them or the PostgreSQL log, planted canary detected" || record B-BRANDING-IT FAIL "branding ($n/23 subtest results; logcheck: ${lc:-ok})"
}

# G018 B-ONBOARDING-CATALOG and B-ONBOARDING-IT (Plan r5 §6.2, §6.3 b2), in the migrations suite after the branding
# checks and before the recovery drill. B-ONBOARDING-CATALOG: TestOnboardingCatalog (CT-ONB-1 the exact version-14
# delta with the R-ONB-1 restore-boundary cases, CT-ONB-AUD, CT-ONB-UPDOWN). B-ONBOARDING-IT: TestOnboarding (ONB-1..13
# and ONB-2d/9e) on the attested runtime login, then the record checks. The it_count pins are the real subtest PASS
# lines (5 catalog results incl. the two nested R-ONB-1 cases; 14 TestOnboarding subtests).
onboarding_checks() { # onboarding_checks DEST PROJECT IMAGE
  local M="$1" project="$2" img="$3" ok=1 log="$WORK/onboarding-catalog-it.log" olog="$WORK/onboarding-it.log" n lc
  it_fresh "$M" "$project" onboarding-catalog "$log" && rls_go onboarding-catalog "$project" "$img" "$OBJ_ENV" "$log" ./internal/migrate/ '^TestOnboardingCatalog$' || ok=0
  n="$(it_count "$log" TestOnboardingCatalog)"
  [[ "$n" == 5 ]] || { ok=0; tail -40 "$log"; }
  [[ $ok == 1 ]] && record B-ONBOARDING-CATALOG PASS "TestOnboardingCatalog against real PostgreSQL at version 14: 5/5 subtest results (CT-ONB-1 the exact version-14 delta: the column-limited runtime INSERT grants, the four INSERT policies, the Group bootstrap-owner trigger, the one SECURITY DEFINER eligibility reader and the ACL postcondition, with R-ONB-1 the restore-boundary negative and positive; CT-ONB-AUD the widened access.denied operation constraint; CT-ONB-UPDOWN clean down/up and the history-preserving down refusal)" || record B-ONBOARDING-CATALOG FAIL "TestOnboardingCatalog: ${n} subtest results"
  ok=1
  it_fresh "$M" "$project" onboarding-it "$olog" && rls_go onboarding-it "$project" "$img" "$OBJ_ENV" "$olog" ./internal/tenancy/access/ '^TestOnboarding$' || ok=0
  n="$(it_count "$olog" TestOnboarding)"
  [[ "$n" == 14 ]] || { ok=0; tail -60 "$olog"; }
  lc="$(rls_logcheck onboarding-it "$M" "$olog")"
  [[ -z "$lc" ]] || ok=0
  [[ $ok == 1 ]] && record B-ONBOARDING-IT PASS "TestOnboarding against real PostgreSQL at version 22 on the attested runtime login: 14/14 subtests (ONB-1 permitted creators with the trigger-written owner binding; ONB-2/2d creation denials and fixed-actor runtime SQL with zero effects; ONB-3 atomicity and the bootstrap race; ONB-4 slug conflict and cross-Organization independence; ONB-5/6 new-profile completion and its faults without an orphan profile; ONB-8 a complete Group; ONB-9/9e existing-profile selection and the eligibility reader; ONB-10 complete paged traversal with per-row decisions; ONB-11 retry as an authorized conflict; ONB-12 the display-only can_update hint; ONB-13 member self-service); record check passed" || record B-ONBOARDING-IT FAIL "TestOnboarding: ${n} subtests or a record check failed"
}

# G019 B-MEMBERSHIP-CATALOG and B-MEMBERSHIP-IT (Plan r1 §6.2 CT-MEM-1, §6.3 b2), in the migrations suite after the
# onboarding checks and before the recovery drill, on the onboarding_checks template. B-MEMBERSHIP-CATALOG:
# TestMembershipCatalog (CT-MEM-1 the legacy owner preflight for an ownerless Organization and an ownerless Group with the
# DV-1 rollback postconditions, the exact version-15 delta, least privilege under actual grants; CT-MEM-AUD; CT-MEM-UPDOWN).
# B-MEMBERSHIP-IT: TestMembership on the attested runtime login, then the record checks. The it_count pins are the real
# subtest PASS lines: 6 catalog results (the two preflight cases are slash-named top-level subtests); 15 TestMembership
# subtests plus the 8 nested MEM-11 negatives, the 3 nested MEM-9 parts and the 5 nested MEM-6 denial rows (31); G060
# adds MEM-13b (32) and runs TestAuthority in the same container (the owner-only override write AUT-1..6, override
# administration MEM-OVR-1..3/5..8 and the stale-issuer claims O-OVR-2/4: 15 subtests; Plan r0 §2 authority_integration_test.go).
membership_checks() { # membership_checks DEST PROJECT IMAGE
  local M="$1" project="$2" img="$3" ok=1 log="$WORK/membership-catalog-it.log" mlog="$WORK/membership-it.log" n a lc
  it_fresh "$M" "$project" membership-catalog "$log" && rls_go membership-catalog "$project" "$img" "$OBJ_ENV" "$log" ./internal/migrate/ '^TestMembershipCatalog$' || ok=0
  n="$(it_count "$log" TestMembershipCatalog)"
  [[ "$n" == 6 ]] || { ok=0; tail -40 "$log"; }
  [[ $ok == 1 ]] && record B-MEMBERSHIP-CATALOG PASS "TestMembershipCatalog against real PostgreSQL at version 15: 6/6 subtest results (CT-MEM-1 the legacy owner preflight refusing an ownerless Organization and an ownerless Group with counts only, RLS and FORCE kept on the three lifted tables, version 14 and no partial object, policy, grant or body after each rollback, then passing once resolved; the exact version-15 delta: invitation table, CHECKs, indexes and triggers, the eight SECURITY DEFINER helpers (five policy-reader readers, three claim-writer bodies) with owner, volatility, search_path, ACL and reviewed bodies, the claim writer's whole authority, the policy strings, events_attribute = the 00013 body + the invitation shape; least privilege under actual grants; CT-MEM-AUD the widened audit constraints; CT-MEM-UPDOWN clean down/up and the history-preserving down refusal)" || record B-MEMBERSHIP-CATALOG FAIL "TestMembershipCatalog: ${n} subtest results"
  ok=1
  it_fresh "$M" "$project" membership-it "$mlog" && rls_go membership-it "$project" "$img" "$OBJ_ENV" "$mlog" ./internal/tenancy/access/ '^(TestMembership|TestAuthority)$' || ok=0
  n="$(it_count "$mlog" TestMembership)"; a="$(it_count "$mlog" TestAuthority)"
  [[ "$n" == 32 && "$a" == 15 ]] || { ok=0; tail -60 "$mlog"; }
  lc="$(rls_logcheck membership-it "$M" "$mlog")"
  [[ -z "$lc" ]] || ok=0
  [[ $ok == 1 ]] && record B-MEMBERSHIP-IT PASS "TestMembership and TestAuthority against real PostgreSQL at version 22 on the attested runtime login: 32/32 and 15/15 subtest results (G060 TestAuthority: AUT-1 owner enable/disable with the organization.policy_changed reasons, AUT-2 admin, Group owner and member refused, AUT-3 direct runtime SQL by an admin meets the owner-only guard, AUT-4 the generic policy patch never sets the override, AUT-5 fresh Organizations default to false, AUT-6 a disabled owner's raw UPDATE meets the enabled-owner guard; MEM-OVR-1..3 and 5..8 bounded override administration, its exclusions, revocation, the never-consumed self-targeted intent and the override issuer's invitations; O-OVR-2/4 a stale override or disabled issuer refuses the claim; G060 MEM-13b departure audit-fault rollback; MEM-10 Organization-invitation consumption; MEM-11 its 8 negatives: expired, revoked, consumed by another, wrong and foreign token, token-kind crossover, address case, disabled recipient, issuer demoted; MEM-14 the own-Organizations reader, every-page retirement, stop and resume, and the identity finalization refusing a new membership, a rejoin and a late sole ownership and waiting in both lock orders; MEM-8 the 99→100 pending cap race, the refused 101st and 101-row paging; MEM-9 visibility, in-place upgrade with consumed_binding_id and the refused consumed revoke; MEM-2 both owners administer with hints; MEM-6 five denial attributions; MEM-7 transfer audit-fault rollback with its control; MEM-13 dependent departure after succession and readmission); access records catalogued, no canary in them or the PostgreSQL log, planted canary detected" || record B-MEMBERSHIP-IT FAIL "TestMembership: ${n}, TestAuthority: ${a} subtest results (logcheck: ${lc:-ok})"
}

# G058 B-ENROLLMENT-CATALOG and B-ENROLLMENT-IT (Plan r3 §6.2/§6.3, §11.2; addendum §A.6; correction 01 §C.6), in the
# migrations suite after the membership checks and before the recovery drill, each invocation on its own fresh copy (G064).
# B-ENROLLMENT-CATALOG: TestEnrollmentCatalog with the bound infra/recovery/sql/replay-restrict.sql (ENROLLMENT_REPLAY_SQL_B64,
# the MCP_INVALIDATE_SQL_B64 precedent; the test pins its SHA-256). B-ENROLLMENT-IT: TestEnrollment, TestEnrollmentAdmission
# and TestEnrollmentRequests on the attested runtime login, then the record checks. The it_count pins are the real recursive
# subtest PASS lines: 19 catalog results (12 top-level CT-ENR-1 subtests incl. replay_pre16 and replay, plus the 7 nested
# ENR-16 replay rows); 2 TestEnrollment; 14 TestEnrollmentAdmission (11 rows plus the 3 nested O-ENR-8 interleavings); 15
# TestEnrollmentRequests (11 rows plus the 2 nested O-ENR-RV and the 2 nested O-ENR-6 orders).
enrollment_checks() { # enrollment_checks DEST PROJECT IMAGE
  local M="$1" project="$2" img="$3" ok=1 log="$WORK/enrollment-catalog-it.log" elog="$WORK/enrollment-it.log" n a q lc
  it_fresh "$M" "$project" enrollment-catalog "$log" && rls_go enrollment-catalog "$project" "$img" "$OBJ_ENV" "$log" ./internal/migrate/ '^TestEnrollmentCatalog$' \
    -e ENROLLMENT_REPLAY_SQL_B64="$(base64 -w0 "$REPO/infra/recovery/sql/replay-restrict.sql")" || ok=0
  n="$(it_count "$log" TestEnrollmentCatalog)"
  [[ "$n" == 19 ]] || { ok=0; tail -40 "$log"; }
  [[ $ok == 1 ]] && record B-ENROLLMENT-CATALOG PASS "TestEnrollmentCatalog against real PostgreSQL at versions 15 and 16: 19/19 subtest results (CT-ENR-1 the legacy postcondition closing a version-15 Organization; the exact version-16 delta: the three tables, the policies, least privilege under actual grants, the eight SECURITY DEFINER bodies and the departure and decision triggers, the widened audit constraints and attribution; rolled-back negatives with the trigger-bound ledger atomicity; the operator readiness setter and backfill with the live-side close and the NULL/grammar refusals; the request state machine through the real runtime and claim-writer paths; ENR-16 the bound replay-restrict.sql at version 15 (ledger absent) and on seven emulated restored states: trigger row present, restored row present, materialized newer than a restored approval with admission refused and zero writes, restored backfill present, restore before the backfill materialized and still closed, the ready-checkpoint/later-backfill gap materialized with open admission refused and a control admitted, an unmaterializable entry closing readiness idempotently; CT-ENR-1_updown clean down/up and the history-preserving down refusal)" || record B-ENROLLMENT-CATALOG FAIL "TestEnrollmentCatalog: ${n} subtest results"
  ok=1
  it_fresh "$M" "$project" enrollment-it "$elog" && rls_go enrollment-it "$project" "$img" "$OBJ_ENV" "$elog" ./internal/tenancy/access/ '^(TestEnrollment|TestEnrollmentAdmission|TestEnrollmentRequests)$' || ok=0
  n="$(it_count "$elog" TestEnrollment)"; a="$(it_count "$elog" TestEnrollmentAdmission)"; q="$(it_count "$elog" TestEnrollmentRequests)"
  [[ "$n" == 2 && "$a" == 14 && "$q" == 15 ]] || { ok=0; tail -60 "$elog"; }
  lc="$(rls_logcheck enrollment-it "$M" "$elog")"
  [[ -z "$lc" ]] || ok=0
  [[ $ok == 1 ]] && record B-ENROLLMENT-IT PASS "TestEnrollment, TestEnrollmentAdmission and TestEnrollmentRequests against real PostgreSQL at version 22 on the attested runtime login: 2/2, 14/14 and 15/15 subtest results (ENR-1 closed default and ready by construction; ENR-3 policy authority and readiness; ENR-7 first-time invitation claim and lift; ENR-23 stale invitation rolled back; ENR-8 replay and wrong token; open restricted and closed; ENR-11 claim fault rollback; ENR-HOST, ENR-AUD, ENR-LOST, ENR-REMOVED, ENR-INTENT; O-ENR-8 three claim/departure interleavings; ENR-9/9D/9R approve, deny and revoke; ENR-RQA manager authority; ENR-RQE expiry; ENR-RQB atomic append; ENR-RQL class-40 versus unknown COMMIT; O-ENR-RQ one_pending, O-ENR-7 two consumers, O-ENR-RV and O-ENR-6 both orders); access records catalogued, no canary in them or the PostgreSQL log, planted canary detected" || record B-ENROLLMENT-IT FAIL "enrollment: TestEnrollment ${n}, TestEnrollmentAdmission ${a}, TestEnrollmentRequests ${q} subtest results (logcheck: ${lc:-ok})"
}

# G062 B-SIGNUP-CATALOG and B-SIGNUP-IT (Plan r5 §2 harness row, §5.2 SUP-4..8/12, §5.3 b1/b2), in the migrations suite after
# the enrollment checks and before the recovery drill, each invocation on its own fresh copy (G064). B-SIGNUP-CATALOG:
# TestSignupCatalog (resets the catalog and rebuilds 17 -> 18; the db_fresh after it is T-2r, which checks its role residue). B-SIGNUP-IT: TestSignupProvisioning, TestSignupSlugRace and
# TestSignupContinuation on the runtime pool with the in-process fake provider, then the record checks. The pins are the real
# PASS lines: 9 CT-SUP-1 subtests; 7 provisioning subtests (incl. the generic-plan regression); TestSignupSlugRace has no
# subtests (its top-level line is counted, the TestMCPFootprint precedent); 5 continuation subtests. G057 (Plan r5 §5.2 T4):
# provisioning gains the configured-reserved-label case (8) and TestSignupSlugRace its first subtest, the forced label
# collision (1, counted beside the top-level line).
signup_checks() { # signup_checks DEST PROJECT IMAGE
  local M="$1" project="$2" img="$3" ok=1 log="$WORK/signup-catalog-it.log" slog="$WORK/signup-it.log" n p r s c lc
  it_fresh "$M" "$project" signup-catalog "$log" && rls_go signup-catalog "$project" "$img" "$OBJ_ENV" "$log" ./internal/migrate/ '^TestSignupCatalog$' || ok=0
  n="$(it_count "$log" TestSignupCatalog)"
  [[ "$n" == 9 ]] || { ok=0; tail -40 "$log"; }
  [[ $ok == 1 ]] && record B-SIGNUP-CATALOG PASS "TestSignupCatalog against real PostgreSQL at versions 17 and 18: 9/9 subtests (CT-SUP-1 precondition: 00018 refused without the provision writer or with LOGIN/INHERIT, nothing applied; tables; policies; privileges; the four DEFINER bodies; negatives incl. TENANT/LOGIN/empty shapes, foreign-Organization completion and owner/runtime INSERT refusals; the provision writer's exact grants and non-grants; return-path CHECKs; updown: down refused with signup rows, exact version-17 catalog after down and version-18 after up)" || record B-SIGNUP-CATALOG FAIL "TestSignupCatalog: ${n} subtest results"
  ok=1
  it_isolation_t2_repeat "$M" "$project" signup-catalog "$slog" && rls_go signup-it "$project" "$img" "$OBJ_ENV" "$slog" ./internal/signup/ '^(TestSignupProvisioning|TestSignupSlugRace|TestSignupContinuation)$' || ok=0
  p="$(it_count "$slog" TestSignupProvisioning)"; c="$(it_count "$slog" TestSignupContinuation)"
  r="$(grep -c -- '^--- PASS: TestSignupSlugRace ' "$slog" || true)"; s="$(it_count "$slog" TestSignupSlugRace)"
  [[ "$p" == 8 && "$r" == 1 && "$s" == 1 && "$c" == 5 ]] || { ok=0; tail -60 "$slog"; }
  lc="$(rls_logcheck signup-it "$M" "$slog")"
  [[ -z "$lc" ]] || ok=0
  [[ $ok == 1 ]] && record B-SIGNUP-IT PASS "TestSignupProvisioning, TestSignupSlugRace and TestSignupContinuation against real PostgreSQL at version 22 on the runtime pool with the in-process fake provider: 8/8, 1 with 1/1 and 5/5 results (created with audit attribution and the generated platform label; a configured reserved 26-character label never issued; injected faults retain nothing; SUP-7 lost response, acknowledged retry and re-mint, expired credential, concurrent identical retries, disabled creator and the disable race; the generic-plan backend; 5x two-user slug race with one winner and zero loser writes; a forced platform-label collision retried with a fresh label; continuation happy path to tenant_created, captured R/N/P, wrong root context, wrong tenant host and expiry, completion at another host, account substitution, removed owner and a moved host); no canary in the PostgreSQL log, planted canary detected" || record B-SIGNUP-IT FAIL "signup: TestSignupProvisioning ${p}, TestSignupSlugRace ${r} (subtests ${s}), TestSignupContinuation ${c} results (logcheck: ${lc:-ok})"
}

# G020 B-LABELS-CATALOG, B-LABELS-IT and B-LABELS-ORDERING (Plan r2 §2 harness row, §5.3 b2), in the migrations suite
# after the signup checks and the B-IT-ISOLATION record and before the recovery drill, each invocation on its own fresh copy
# (G064). B-LABELS-CATALOG: TestLabelsCatalog (resets the catalog and rebuilds 19 -> 20, so it runs first in its
# invocation) and TestLabelsMigration. B-LABELS-IT: TestLabels on the
# attested runtime login, then the record checks. B-LABELS-ORDERING: TestLabelsOrdering (rls_ordering_integration_test.go)
# with deterministic barriers, then the record checks. The pins are the real subtest PASS lines: 4 CT-LAB and 4 MG-LAB
# subtests; 15 LAB subtests covering LAB-1..17 (LAB-6..8 are one subtest); 4 O-LAB subtests.
labels_checks() { # labels_checks DEST PROJECT IMAGE
  local M="$1" project="$2" img="$3" ok=1 log="$WORK/labels-catalog-it.log" llog="$WORK/labels-it.log" olog="$WORK/labels-ordering-it.log" n m lc
  it_fresh "$M" "$project" labels-catalog "$log" && rls_go labels-catalog "$project" "$img" "$OBJ_ENV" "$log" ./internal/migrate/ '^(TestLabelsCatalog|TestLabelsMigration)$' || ok=0
  n="$(it_count "$log" TestLabelsCatalog)"; m="$(it_count "$log" TestLabelsMigration)"
  [[ "$n" == 4 && "$m" == 4 ]] || { ok=0; tail -40 "$log"; }
  [[ $ok == 1 ]] && record B-LABELS-CATALOG PASS "TestLabelsCatalog against real PostgreSQL at versions 19 and 20: 4/4 subtests (CT-LAB-1 the exact version-20 delta: the two label tables, their columns, ACLs and FORCE row-level security; CT-LAB-2 production assignment storage closed to group targets; CT-LAB-3 policies, triggers, functions and capture shapes with the widened audit constraints; CT-LAB-4 helper inventories unchanged and the assignments EXISTS one index probe); TestLabelsMigration 4/4 (MG-LAB-1 a populated version-19 world to 20; MG-LAB-2 down refused with a label row or label history, catalog unchanged; MG-LAB-3 down and up on a label-free world restore the exact version-19 and version-20 catalogs; MG-LAB-4 a rolled-back up leaves version 19 and the owner)" || record B-LABELS-CATALOG FAIL "TestLabelsCatalog: ${n}, TestLabelsMigration: ${m} subtest results"
  ok=1
  it_fresh "$M" "$project" labels-it "$llog" && rls_go labels-it "$project" "$img" "$OBJ_ENV" "$llog" ./internal/tenancy/access/ '^TestLabels$' || ok=0
  n="$(it_count "$llog" TestLabels)"
  [[ "$n" == 15 ]] || { ok=0; tail -60 "$llog"; }
  lc="$(rls_logcheck labels-it "$M" "$llog")"
  [[ -z "$lc" ]] || ok=0
  [[ $ok == 1 ]] && record B-LABELS-IT PASS "TestLabels against real PostgreSQL at version 22 on the attested runtime login: 15/15 subtests covering LAB-1..17, LAB-6..8 being one subtest (LAB-1/2 the Organization- and Group-scope lifecycles; LAB-3 name collisions; LAB-4 the 200-definition bound; LAB-5 input refusals before any write; LAB-6..8 assign and unassign on the real Group target; LAB-9 delete removes the definition's assignments in one transaction; LAB-10 foreign, unknown, cross-Group, unsupported, ambiguous and malformed targets; LAB-11 Organization-owned labels for Organization owners and admins only, Group roles and the override confer nothing; LAB-12 Group-owned labels for the exact owning Group's roles only, members read target_members, staff and the override nothing; LAB-13..16 the test-only fixture target; LAB-17 the deletion context); access records catalogued, no canary in them or the PostgreSQL log, planted canary detected" || record B-LABELS-IT FAIL "TestLabels: ${n} subtest results (logcheck: ${lc:-ok})"
  ok=1
  it_fresh "$M" "$project" labels-ordering "$olog" && rls_go labels-ordering "$project" "$img" "$OBJ_ENV" "$olog" ./internal/tenancy/access/ '^TestLabelsOrdering$' || ok=0
  n="$(it_count "$olog" TestLabelsOrdering)"
  [[ "$n" == 4 ]] || { ok=0; tail -60 "$olog"; }
  lc="$(rls_logcheck labels-ordering "$M" "$olog")"
  [[ -z "$lc" ]] || ok=0
  [[ $ok == 1 ]] && record B-LABELS-ORDERING PASS "TestLabelsOrdering against real PostgreSQL at version 22: 4/4 subtests pass, none skipped (O-LAB-1 assignment versus definition deletion in both orders, no orphan; O-LAB-2 narrowed visibility, a removed Group role and a selected-Group switch govern the next read; O-LAB-3 a duplicate assignment race: one already_assigned, one row, one audit row; O-LAB-4 creations racing the 200 bound serialize on the scope row, a mixed pair commits both); access records catalogued, no canary in them or the PostgreSQL log, planted canary detected" || record B-LABELS-ORDERING FAIL "TestLabelsOrdering: ${n} subtest results (logcheck: ${lc:-ok})"
}

# G057 B-PLATFORM-ADDRESS-CATALOG (Plan r5 §3.1, §3.5, §5.3; D-15), in the migrations suite after the label checks and
# before the recovery drill, on its own fresh copy (G064). TestPlatformLabelsCatalog (resets the catalog and rebuilds
# 20 -> 21, so it runs first in its invocation) and TestPlatformLabelsMigration (the D-4 down state matrix). The pins are
# the real subtest PASS lines: 6 CT-PLB and 4 MG-PLB subtests.
platform_address_checks() { # platform_address_checks DEST PROJECT IMAGE
  local M="$1" project="$2" img="$3" ok=1 log="$WORK/platform-address-catalog-it.log" n m
  it_fresh "$M" "$project" platform-address-catalog "$log" && rls_go platform-address-catalog "$project" "$img" "$OBJ_ENV" "$log" ./internal/migrate/ '^(TestPlatformLabelsCatalog|TestPlatformLabelsMigration)$' || ok=0
  n="$(it_count "$log" TestPlatformLabelsCatalog)"; m="$(it_count "$log" TestPlatformLabelsMigration)"
  [[ "$n" == 6 && "$m" == 4 ]] || { ok=0; tail -40 "$log"; }
  [[ $ok == 1 ]] && record B-PLATFORM-ADDRESS-CATALOG PASS "TestPlatformLabelsCatalog against real PostgreSQL at versions 20 and 21: 6/6 subtests (CT-PLB-1 the exact version-21 delta: app.organizations.platform_label with its grammar and reserved-floor CHECKs and partial unique index, 89 relations; CT-PLB-2 the column-ACL delta is exactly auth-reader SELECT and provision-writer INSERT on platform_label, no UPDATE for any role; CT-PLB-3 the writer-only organizations_provision_insert also requires platform_label, every other policy unchanged; CT-PLB-4 the replaced hostname_state and 8-argument signup_provision_organization and the new signup_slug_taken: owners, search_path, no PUBLIC EXECUTE, EXECUTE holders, the 7-argument body absent and the helper inventories; CT-PLB-5 the bodies under their definer roles, rolled back: a label host resolves to its UUID and the slug host to no row, created/already_created with the stored origin, the NULL-label refusal, label_taken, and the 42501 negatives; CT-PLB-6 down and up on a signup-free world restore the exact version-20 and version-21 catalogs); TestPlatformLabelsMigration 4/4 (MG-PLB-1 a populated version-20 world to 21, backfilled under the FORCE bracket with no audit row; MG-PLB-2 down refused for each live, re-mintable or ambiguous label state, catalog unchanged; MG-PLB-3 down passes each dead-history state, restoring version 20 exactly, and up restores version 21; MG-PLB-4 a rolled-back up leaves version 20 and the owner)" || record B-PLATFORM-ADDRESS-CATALOG FAIL "TestPlatformLabelsCatalog: ${n}, TestPlatformLabelsMigration: ${m} subtest results"
}

# B-MCP-IT: the MCP transport, lifecycle, audit, G015 discovery and recovery tests (core/internal/mcp, real HTTP over
# the production constructor, the official SDK client, the real facade and runtime role, the identitytest Connect
# issuer as the test signer; the actual infra/recovery/sql/invalidate.sql), then the facade ordering and footprint
# (E-MCP lines kept as mcp-emcp.jsonl) and the record checks.
mcp_it() { # mcp_it DEST PROJECT IMAGE
  local M="$1" project="$2" img="$3" ok=1 log="$WORK/mcp-it.log" flog="$WORK/mcp-facade-it.log" n l o f a r d lc triggers
  it_fresh "$M" "$project" mcp-it "$log" && rls_go mcp-it "$project" "$img" "$OBJ_ENV" "$log" ./internal/mcp/ '^(TestMCPTransport|TestMCPLifecycle|TestMCPAudit|TestMCPRecovery|TestMCPDiscovery)$' \
    -e MCP_LOG_FILE=/tmp/rls-access.log -e MCP_INVALIDATE_SQL_B64="$(base64 -w0 "$REPO/infra/recovery/sql/invalidate.sql")" || ok=0
  n="$(it_count "$log" TestMCPTransport)"; l="$(it_count "$log" TestMCPLifecycle)"
  a="$(grep -c -- '^--- PASS: TestMCPAudit ' "$log" || true)"; r="$(grep -c -- '^--- PASS: TestMCPRecovery ' "$log" || true)"
  d="$(it_count "$log" TestMCPDiscovery)"
  [[ "$n" == 7 && "$l" == 7 && "$a" == 1 && "$r" == 1 && "$d" == 6 ]] || { ok=0; tail -60 "$log"; }
  it_fresh "$M" "$project" mcp-facade-it "$flog" && rls_go mcp-facade-it "$project" "$img" "$OBJ_ENV" "$flog" ./internal/tenancy/access/ '^(TestMCPOrdering|TestMCPFootprint)$' || ok=0
  o="$(it_count "$flog" TestMCPOrdering)"; f="$(grep -c -- '^--- PASS: TestMCPFootprint ' "$flog" || true)"
  [[ "$o" == 11 && "$f" == 1 ]] || { ok=0; tail -60 "$flog"; }
  grep -E '^E-MCP(-3|-4|-TRIGGER)? ' "$flog" >"$WORK/mcp-emcp.jsonl" || true
  [[ "$(grep -cE '^E-MCP(-3|-4)? ' "$WORK/mcp-emcp.jsonl")" == 5 ]] || ok=0
  if [[ -n "$REPORT_DIR" ]]; then cp "$WORK/mcp-emcp.jsonl" "$REPORT_DIR/" 2>/dev/null || true; fi
  triggers="$(grep -c '^E-MCP-TRIGGER ' "$WORK/mcp-emcp.jsonl" || true)"
  [[ "$triggers" == 0 ]] || log "E-MCP: $triggers escalation trigger line(s): return to Codex (mcp-emcp.jsonl; PG-006 stop rule)"
  lc="$(rls_logcheck mcp-it "$M" "$log" "$flog")"
  [[ -z "$lc" ]] || ok=0
  [[ $ok == 1 ]] && record B-MCP-IT PASS "MCP integration against real PostgreSQL at version 11 (the attested runtime login), real HTTP over the production constructor, the official SDK client and the identitytest Connect test signer: TestMCPTransport 7/7 (IT-MCP-1 discovery, identical stateless metadata and transport bounds; IT-MCP-2 host, route and resolution reassignment: slug and verified-alias hosts route, a removed or reassigned host reaches no connection; IT-MCP-3 setup and activation (G063: with and without a name); IT-MCP-4 endpoint, credential and connection binding; IT-MCP-5 identity and context output; IT-MCP-6 selectable Groups; IT-MCP-7 selection); TestMCPLifecycle 7/7 (IT-MCP-8 same URL shares, separate URLs isolate stored state; IT-MCP-9 renewal and terminal states; IT-MCP-11 membership, selected-Group and key-set loss; IT-MCP-12 live access changes; IT-MCP-13 Organization-bound reads; IT-MCP-14 RV-1 local containment; G063 IT-MCP-15 rename: set, clear, the padded non-BMP vector, same-value without an event, foreign-user, foreign-tenant and terminal targets not_found, wire negatives with no change, every connection and session fact unchanged); TestMCPAudit (exact events, strict activation and selection faults, no token or provider value in the records); TestMCPDiscovery 6/6 (G015 CAT-1 the catalog follows the matrix per role, Organization-only or selected context, link and policy, with test-only capability tools and the production catalog unchanged; CAT-2 selection and external changes govern the next list without a selection-revision change, definitive refusals change nothing, terminal states refuse; CAT-3 hidden and unknown names answer byte-identically in valid and refused contexts with truthful identical records, stale calls fail closed, exact-target refusal after listing; CAT-4 catalogs are connection-local across reconnect and renewal; CAT-5 no list-changed notification, the selection revision is not a catalog signal; G020 CAT-6 label capabilities: Organization staff never reach the Group label tools, hidden equals unknown); TestMCPRecovery (the actual G013 invalidation: pending, active and absent connections refuse, a fresh setup works; G063: a rename with the invalidated browser credential is unauthenticated, then not_found after re-sign-in); TestMCPOrdering 11/11 (one activation winner, O-MCP-1..6 and O-MCP-2b: a selection committed after the decision snapshot makes the switch write nothing; a removal from the selected Group committed after the decision refuses a switch away while a clear succeeds; such a removal attempted during a held switch waits for it; G015 O-MCP-7 an eligibility snapshot during a held switch, or spanning its commit, is the old pair; O-MCP-8 a change committed after the catalog is computed governs the next request and the tools' own checks; G063 O-MCP-9 rename against revocation in both orders and with a held revocation, no deadlock, the lock order user, browser session, target session, facts, row); TestMCPFootprint (constant statement counts over 30 extra Groups; E-MCP-1..3 recorded; E-MCP-4 eligibility statements per catalog shape, production equal to the context check, capability within four more, constant over 30 extra Groups and profiles; $triggers trigger lines); core and access records catalogued, no canary in them or the PostgreSQL log, planted canary detected" || record B-MCP-IT FAIL "MCP integration (transport $n/7, lifecycle $l/7, audit $a, recovery $r, discovery $d/6, ordering $o/11, footprint $f; logcheck: ${lc:-ok})"
}

PORTS_REC=(3700 8780 8700 6132 7079 9700 9701)
# The drill's helpers (rd_*): the source is the migrations project, the target a second project on the same copy with
# its own env file (.env.rec, so no new bind-mount path); the tool is the copy's infra/recovery/recovery.sh.
RD_RES="" RD_M="" RD_P="" RD_L="" RD_SP="" RD_TP="" RD_DB="" RD_IMG="" RD_PGIMG="" RD_OLDV="" RD_SRCENV="" RD_TGTENV=""
RD_ROOT="" RD_RCPT="" RD_LEDGER="" RD_BPOL="" RD_SINK="" RD_CID="" IDS="" POST="" B1="" B2="" B3="" REPAIR="" VERIFIED=fail
rd_rec() { printf '%s\t%s\t%s\n' "$1" "$2" "$3" >>"$RD_RES"; [[ "$2" == FAIL ]] && log "drill $1: $3"; return 0; }
rd_ok() { if [[ "$2" == 1 ]]; then rd_rec "$1" PASS "$3"; else rd_rec "$1" FAIL "$3"; fi; }
rd_hex() { random_hex 24; }
rd_save() { printf -v "$1" '%s' "$2"; }
rd_envf() { if [[ "$1" == tgt ]]; then printf '%s' "$RD_TGTENV"; else printf '%s' "$RD_SRCENV"; fi; }
rd_ev() { sed -n "s/^$2=//p" "$(rd_envf "$1")" | tail -1 | sed 's/^"\(.*\)"$/\1/'; }
rd_pgc() { docker ps -q --filter "label=com.docker.compose.project=$1" --filter label=com.docker.compose.service=postgres | head -1; }
rd_mic() { docker ps -aq --filter "label=com.docker.compose.project=$1" --filter label=com.docker.compose.service=minio | head -1; }
rd_sql() { local p="$RD_SP"; [[ "$1" == tgt ]] && p="$RD_TP"; docker exec -i "$(rd_pgc "$p")" psql -X -q -tA -v ON_ERROR_STOP=1 -U postgres -d "$RD_DB" -c "$2"; }
# G064 D2: product-code creators are latch-checked before (refused: 125) and recounted after (a crossing latches).
rd_tool() {
  res_gate "rd_tool $1" || { printf 'G064 rd_tool %s refused (ran nothing): %s\n' "$1" "$RES_WHY" >&2; return 125; }
  RECOVERY_LABEL="g001.run=$RUN_ID" RECOVERY_CID_DIR="$RD_CID" RECOVERY_STEP_TIMEOUT=60 bash "$RD_M/infra/recovery/recovery.sh" "$@"; local rc=$?; rd_cids
  res_recount "rd_tool $1" || true
  return $rc
}
# rd_tool_bg: only ever as "rd_tool_bg ... &": the background subshell becomes the tool process (exec), so a signal
# sent to $! (the TERM and SIGKILL cases) reaches the tool itself, as an operator's would.
rd_tool_bg() { res_gate "rd_tool_bg $1" || { printf 'G064 rd_tool_bg %s refused (ran nothing): %s\n' "$1" "$RES_WHY" >&2; exit 125; }; RECOVERY_LABEL="g001.run=$RUN_ID" RECOVERY_CID_DIR="$RD_CID" RECOVERY_STEP_TIMEOUT=60 exec bash "$RD_M/infra/recovery/recovery.sh" "$@"; }
rd_dc() { # rd_dc src|tgt ARGS...: Compose for the source (the harness's compose) or the target (its own env file)
  local side="$1"; shift
  if [[ "$side" == src ]]; then compose "$RD_M" "$@"; return; fi
  res_compose_begin "$@" || return 125
  local rc=0; (cd "$RD_M" && docker compose --env-file .env.rec -f infra/compose/local.yml "$@") || rc=$?
  res_compose_end
  rd_capture "$RD_TP"; return $rc
}
rd_capture() { # rd_capture PROJECT: ownership rows for the target project's containers (owned_capture reads only .env)
  [[ -n "$OWNED_REG" ]] || return 0
  local cid x
  for cid in $(docker ps -aq --no-trunc --filter "label=com.docker.compose.project=$1" 2>/dev/null); do
    id_at_start container "$cid" && continue
    grep -qxF "container	$cid" "$OWNED_REG" 2>/dev/null || registry_row container "$cid"
    for x in $(docker inspect -f '{{range $k, $v := .NetworkSettings.Networks}}{{$v.NetworkID}} {{end}}' "$cid" 2>/dev/null); do id_at_start network "$x" || registry_row network "$x"; done
    for x in $(docker inspect -f '{{range .Mounts}}{{if eq .Type "volume"}}{{.Name}} {{end}}{{end}}' "$cid" 2>/dev/null); do id_at_start volume "$x" || registry_row volume "$x"; done
  done
}
rd_cids() { # rd_cids: ownership rows for the cidfiles the tool wrote for its detached holders (distinct, initially absent files)
  [[ -n "$OWNED_REG" && -d "$RD_CID" ]] || return 0
  local f
  for f in "$RD_CID"/*-holder-*.cid; do
    [[ -s "$f" && ! -e "$f.done" ]] || continue
    registry_row standalone "$(cat "$f")" recovery-holder; : >"$f.done"
  done
}
# rd_client_rm NAME...: remove the drill's detached standalone clients (g013-rd-$RUN_ID-NAME) together with the
# anonymous volumes they mount. The PostgreSQL image declares /var/lib/postgresql/data, so every detached psql client
# (adminw, rtw, lock) is created with one; removing the container alone left three volumes behind (MR !29 staging
# gate, exit 14). While the container still exists its anonymous mounts (64-hex names absent from the start
# inventory) are read and registered as owned volume rows; the container is removed with them (docker rm -v) and each
# captured volume is verified gone or removed explicitly by name. Named volumes and anything present at the start are
# never touched; an absent client is not an error.
rd_client_rm() {
  local n cid v vols
  for n in "$@"; do
    cid="$(docker inspect -f '{{.Id}}' "g013-rd-$RUN_ID-$n" 2>/dev/null)" || continue
    [[ -n "$cid" ]] || continue
    vols=""
    for v in $(docker inspect -f '{{range .Mounts}}{{if eq .Type "volume"}}{{.Name}} {{end}}{{end}}' "$cid" 2>/dev/null); do
      [[ "$v" =~ ^[0-9a-f]{64}$ ]] || continue
      id_at_start volume "$v" && continue
      registry_row volume "$v"; vols+="$v "
    done
    docker rm -f -v "$cid" >/dev/null 2>&1 || true
    for v in $vols; do
      docker volume inspect "$v" >/dev/null 2>&1 || continue
      docker volume rm "$v" >/dev/null 2>&1 || log "drill: anonymous volume $v of client $n could not be removed"
    done
  done
  return 0
}
rd_mc() { # rd_mc SIDE admin|svc CMD
  local side="$1" who="$2" proj="$RD_SP" u p; shift 2
  [[ "$side" == tgt ]] && proj="$RD_TP"
  if [[ "$who" == admin ]]; then u="$(rd_ev "$side" MINIO_ROOT_USER)"; p="$(rd_ev "$side" MINIO_ROOT_PASSWORD)"; else u="$(rd_ev "$side" OBJECT_STORE_ACCESS_KEY)"; p="$(rd_ev "$side" OBJECT_STORE_SECRET_KEY)"; fi
  DRO_TIMEOUT=60 docker_run_owned minio-mc --rm -i --label g001.run=$RUN_ID --network "${proj}_data" -e MC_CONFIG_DIR=/tmp/mc --env-file <(printf 'MC_HOST_x=http://%s:%s@minio:9000\n' "$u" "$p") "$proj-minio-init" sh -c "$*"
}
rd_itenv() { # rd_itenv SIDE: the Go phase environment of SIDE (database and scoped object values), mode 600
  local f="$RD_P/it-$1.env" k
  (umask 077
   { printf 'DATABASE_NAME=%s\nDB_HOST=postgres\nDB_PORT=5432\nDB_MIGRATOR_PASSWORD=%s\nDB_RUNTIME_PASSWORD=%s\nDB_SUPERUSER=postgres\nDB_SUPERUSER_PASSWORD=%s\nAPP_ENV=test\n' \
       "$RD_DB" "$(rd_ev "$1" DB_MIGRATOR_PASSWORD)" "$(rd_ev "$1" DB_RUNTIME_PASSWORD)" "$(rd_ev "$1" POSTGRES_SUPERUSER_PASSWORD)"
     for k in OBJECT_STORE_ENDPOINT OBJECT_STORE_REGION OBJECT_STORE_BUCKET OBJECT_STORE_FORCE_PATH_STYLE OBJECT_STORE_ACCESS_KEY OBJECT_STORE_SECRET_KEY; do
       printf '%s=%s\n' "$k" "$(rd_ev "$1" "$k")"; done; } >"$f")
  printf '%s' "$f"
}
rd_go() { # rd_go SIDE LOG RUN [docker args...]: one drill Go phase on SIDE's data network
  local side="$1" log="$2" run="$3" net="${RD_SP}_data" envf; shift 3
  [[ "$side" == tgt ]] && net="${RD_TP}_data"
  envf="$(rd_itenv "$side")"
  local rc=0
  GO_IT_NETWORK="$net" GO_IT_TIMEOUT=20m go_it "rd-${log%.log}" "" "$RD_IMG" "$envf" "$RD_L/$log" ./internal/tenancy/access/ "$run" \
    -v "$RD_OLDV:/recovery-old" -e RECOVERY_OLD_DIR=/recovery-old \
    -e RLS_CANARIES="$RLS_CANARIES" -e TENANCY_CANARY_FILE=/tmp/rls-canaries.txt -e TENANCY_LOG_FILE=/tmp/rls-access.log "$@" || rc=$?
  rm -f "$envf"; return $rc
}
rd_gopass() { grep -q -- "^--- PASS: $1 " "$RD_L/$2" && ! grep -qE -- '--- (FAIL|SKIP)' "$RD_L/$2"; }
rd_policy() { local f="$1"; shift; (umask 077; printf '%s\n' "$@" >"$f"); }
rd_fp() { { docker exec -i "$(rd_pgc "$1")" psql -X -q -tA -U postgres -d "$RD_DB" <"$RD_M/infra/recovery/sql/fingerprint-data.sql"
            docker exec -i "$(rd_pgc "$1")" psql -X -q -tA -U postgres -d "$RD_DB" -v db="$RD_DB" <"$RD_M/infra/recovery/sql/fingerprint-catalog.sql"; } | sha256sum | cut -c1-16; }
rd_flip() { local sz o b; sz="$(stat -c %s "$1")"; o=$((sz / 2)); b="$(dd if="$1" bs=1 skip="$o" count=1 status=none | od -An -tu1 | tr -d ' ')"; printf "$(printf '\\%03o' $(( b ^ 255 )))" | dd of="$1" bs=1 seek="$o" conv=notrunc status=none; }
rd_j() { if [[ $# == 2 ]]; then sed -n "s/.*\"$2\":\"\([^\"]*\)\".*/\1/p" <<<"$1" | head -1; else grep -o "\"$2\":{[^}]*}" <<<"$1" | sed -n "s/.*\"$3\":\"\([^\"]*\)\".*/\1/p" | head -1; fi; }
rd_evcount() { rd_sql tgt "SELECT count(*) FROM audit.events WHERE $1"; }

# The drill stages (Plan r3 §10.2 RD-* and MD-*), in order.
rd_stage_rds() {
  local f rc all=1
  for f in missing-holds unknown-key duplicate-key both-cutoffs future-cutoff oversized-holds; do
    rc=0; rd_tool retention audit --policy "$RD_M/infra/recovery/testdata/policy/$f.policy" --env-file "$RD_SRCENV" --preview "$RD_P/x" >/dev/null 2>&1 || rc=$?
    [[ $rc == 2 ]] || { all=0; rd_rec RD-S FAIL "$f exit $rc"; }
  done
  local lg="$RD_P/ledger/fixture.tsv"; rm -f "$lg"
  local u1 u2 u3 neg=0 k r
  u1="$(cat /proc/sys/kernel/random/uuid)"; u2="$(cat /proc/sys/kernel/random/uuid)"; u3="$(cat /proc/sys/kernel/random/uuid)"
  rd_tool ledger init --ledger "$lg" --env-file "$RD_SRCENV" >/dev/null && rd_tool ledger append --ledger "$lg" --kind profile.unlink --refs "organization=$u1,group=$u2,profile=$u3" --by fixture >/dev/null &&
    rd_tool ledger append --ledger "$lg" --kind membership.remove --refs "organization=$u1,subject=$u2,scope_type=organization,scope_id=$u1,role=org_member" --by fixture >/dev/null || all=0
  # Invalid references refuse at append (exit 2): upper-case UUID, missing key, unknown key, role outside the closed set,
  # role of the other scope, an Organization binding whose scope is not its Organization, and a "downgrade" upwards.
  for k in "profile.unlink|organization=AAAAAAAA-BBBB-4CCC-8DDD-EEEEEEEEEEEE,group=$u2,profile=$u3" "membership.remove|organization=$u1" \
           "profile.unlink|organization=$u1,group=$u2,profile=$u3,extra=$u1" \
           "membership.remove|organization=$u1,subject=$u2,scope_type=group,scope_id=$u3,role=group_superuser" \
           "membership.remove|organization=$u1,subject=$u2,scope_type=group,scope_id=$u3,role=org_member" \
           "membership.remove|organization=$u1,subject=$u2,scope_type=organization,scope_id=$u3,role=org_member" \
           "membership.downgrade|organization=$u1,subject=$u2,scope_type=group,scope_id=$u3,from_role=group_member,to_role=group_admin"; do
    r=0; rd_tool ledger append --ledger "$lg" --kind "${k%%|*}" --refs "${k#*|}" --by fixture >/dev/null 2>&1 || r=$?
    [[ $r == 2 ]] && neg=$((neg + 1)) || { all=0; rd_rec RD-S FAIL "invalid refs accepted (${k%%|*} exit $r)"; }
  done
  rd_tool ledger verify --ledger "$lg" >/dev/null 2>&1 || { all=0; rd_rec RD-S FAIL "valid chain refused"; }
  cp "$lg" "$lg.t"; sed -i '2s/profile.unlink/policy.disable/' "$lg.t"; rd_tool ledger verify --ledger "$lg.t" >/dev/null 2>&1 && { all=0; rd_rec RD-S FAIL "tampered accepted"; }
  cp "$lg" "$lg.f"; sed -i "1s/\t[0-9a-f-]\{36\}\t/\t$(cat /proc/sys/kernel/random/uuid)\t/" "$lg.f"; rd_tool ledger verify --ledger "$lg.f" >/dev/null 2>&1 && { all=0; rd_rec RD-S FAIL "forked header accepted"; }
  rm -f "$lg" "$lg.t" "$lg.f"
  rd_ok RD-S $all "6 policy refusals exit 2; $neg invalid-reference appends refused (exit 2); chain verify passes, tampered entry and forked header fail (stale: RD-L)"
}

rd_stage_rdb() {
  local out rc a b n
  rd_go src seed.log '^TestRecoverySeed$' || true
  rd_gopass TestRecoverySeed seed.log && IDS="$(sed -n 's/^RECOVERY-SEED //p' "$RD_L/seed.log" | head -1)" || { rd_rec RD-B FAIL "seed"; tail -30 "$RD_L/seed.log"; return 1; }
  rd_save IDS "$IDS"
  rd_rec RD-B-SEED PASS "TestRecoverySeed"
  rd_tool ledger init --ledger "$RD_LEDGER" --env-file "$RD_SRCENV" >/dev/null || { rd_rec RD-B FAIL ledger-init; return 1; }
  local prior_acl; prior_acl="$(rd_sql src "SELECT string_agg(x::text, ',' ORDER BY x::text) FROM pg_database d, unnest(d.datacl) x WHERE d.datname = '$RD_DB'")"
  # (1) administrator write in progress -> admin_write_in_progress, exact release.
  local orgB; orgB="$(rd_j "$IDS" OrgB)"
  # The writer starts after X0 (the drill seam pauses after X2), so the X3 lock wait meets it.
  RECOVERY_DRILL_PAUSE_AFTER_X2=15 rd_tool_bg fence --env-file "$RD_SRCENV" --backup-root "$RD_ROOT" --admin-window CHG-RD-1 >"$RD_L/rdb1.out" 2>&1 & local fpid=$!
  local fr=""; for n in $(seq 120); do fr="$(grep -l $'^db_revoked\t' "$RD_ROOT"/fences/*.tsv 2>/dev/null | head -1)"; [[ -n "$fr" ]] && break; sleep 0.5; done
  docker_run_owned rd-stage -d --label g001.run=$RUN_ID --name g013-rd-$RUN_ID-adminw --network "${RD_SP}_data" --env-file <(printf 'PGPASSWORD=%s\nPGAPPNAME=%s\n' "$(rd_ev src POSTGRES_SUPERUSER_PASSWORD)" "g013-rd-$RUN_ID-adminw") \
    $RD_PGIMG psql -h postgres -U postgres -d $RD_DB -c "BEGIN; UPDATE app.organizations SET name = name WHERE id = '$orgB'; SELECT pg_sleep(120); ROLLBACK;" >/dev/null
  rc=0; wait "$fpid" || rc=$?; out="$(cat "$RD_L/rdb1.out")"
  rd_sql src "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE application_name = 'g013-rd-$RUN_ID-adminw'" >/dev/null; rd_client_rm adminw
  a="$(rd_sql src "SELECT string_agg(x::text, ',' ORDER BY x::text) FROM pg_database d, unnest(d.datacl) x WHERE d.datname = '$RD_DB'")"
  rd_ok RD-B-1 "$([[ $rc == 4 && "$out" == *admin_write_in_progress* && "$a" == "$prior_acl" ]] && rd_mc src admin "mc admin user info --json x core-service" | grep -q '"userStatus":"enabled"' && echo 1 || echo 0)" "exit $rc; acl restored $([[ "$a" == "$prior_acl" ]] && echo yes || echo no)"
  # (2) in-flight writers: a runtime transaction and a held-open PUT; then fence.
  local bh; bh="$(rd_hex | head -c 64)"; bh="${bh}$(rd_hex | head -c 16)"; bh="${bh:0:64}"
  docker_run_owned rd-stage -d --label g001.run=$RUN_ID --name g013-rd-$RUN_ID-rtw --network "${RD_SP}_data" --env-file <(printf 'PGPASSWORD=%s\n' "$(rd_ev src DB_RUNTIME_PASSWORD)") $RD_PGIMG \
    psql -h postgres -U ${RD_DB}_runtime -d $RD_DB -c "BEGIN; SELECT set_config('app.login_binding', '$bh', true); INSERT INTO app.login_transactions (state_hash, binding_hash, pkce_verifier, client_id, origin, return_path, expires_at) VALUES (decode('$bh', 'hex'), decode('$bh', 'hex'), repeat('a', 43), 'client_inflight', 'https://inflight.example', '/', now() + interval '5 minutes'); SELECT pg_sleep(120); COMMIT;" >/dev/null
  local orgA inflight; orgA="$(rd_j "$IDS" OrgA)"; inflight="objects/$orgA/$(cat /proc/sys/kernel/random/uuid)"
  local u p; u="$(rd_ev src OBJECT_STORE_ACCESS_KEY)"; p="$(rd_ev src OBJECT_STORE_SECRET_KEY)"
  docker_run_owned rd-stage -d --label g001.run=$RUN_ID --name g013-rd-$RUN_ID-put --network "${RD_SP}_data" -e MC_CONFIG_DIR=/tmp/mc --env-file <(printf 'MC_HOST_x=http://%s:%s@minio:9000\n' "$u" "$p") "$RD_SP-minio-init" \
    sh -c "(printf 'inflight-bytes'; sleep 150) | mc pipe 'x/$(rd_ev src OBJECT_STORE_BUCKET)/$inflight'" >/dev/null
  sleep 3
  local fp0; fp0="$(rd_fp $RD_SP)"
  rc=0; out="$(rd_tool fence --env-file "$RD_SRCENV" --backup-root "$RD_ROOT" --admin-window CHG-RD-2 2>&1)" || rc=$?
  [[ $rc == 0 ]] || { rd_rec RD-B-2 FAIL "fence exit $rc: $(tail -2 <<<"$out")"; return 1; }
  local lt; lt="$(rd_sql src "SELECT count(*) FROM app.login_transactions WHERE client_id = 'client_inflight'")"
  local rtexit; rtexit="$(docker inspect -f '{{.State.Status}}' g013-rd-$RUN_ID-rtw)"
  rd_ok RD-B-2 "$([[ "$lt" == 0 && "$rtexit" == exited ]] && echo 1 || echo 0)" "runtime writer terminated (container $rtexit), its row absent ($lt)"
  # (3) while fenced: refusals and the SHARE-lock safeguard; status held.
  out="$(docker_run_owned rd-stage --rm --label g001.run=$RUN_ID --network "${RD_SP}_data" --env-file <(printf 'PGPASSWORD=%s\n' "$(rd_ev src DB_RUNTIME_PASSWORD)") $RD_PGIMG psql -h postgres -U ${RD_DB}_runtime -d $RD_DB -c 'SELECT 1' 2>&1)"; a=0; [[ "$out" == *"permission denied for database"* ]] && a=1
  rc=0; printf 'x' | rd_mc src svc "cat >/tmp/f && mc cp -q /tmp/f 'x/$(rd_ev src OBJECT_STORE_BUCKET)/objects/$orgA/late'" >/dev/null 2>&1 || rc=$?
  b=0; [[ $rc != 0 ]] && b=1
  out="$(rd_sql src "SET statement_timeout = '3s'; UPDATE app.organizations SET name = name WHERE id = '$orgB'" 2>&1)"; n=0; [[ "$out" == *"statement timeout"* ]] && n=1
  rc=0; rd_tool status --backup-root "$RD_ROOT" --backup-policy "$RD_BPOL" --receipts "$RD_RCPT" >"$RD_L/status-held.jsonl" 2>&1 || rc=$?
  rd_ok RD-B-3 "$([[ $a == 1 && $b == 1 && $n == 1 && $rc == 1 ]] && grep -q '"check":"backup.fence","status":"held"' "$RD_L/status-held.jsonl" && echo 1 || echo 0)" "runtime login refused $a, late PUT refused $b, superuser UPDATE blocked $n, status exit $rc held"
  # (4) capture with the established fence, then unfence.
  rc=0; out="$(rd_tool backup --env-file "$RD_SRCENV" --backup-root "$RD_ROOT" --ledger "$RD_LEDGER" --use-fence 2>&1)" || rc=$?
  local b1; b1="$(sed -n 's/^backup\t\([0-9a-f-]*\)\tcomplete.*/\1/p' <<<"$out")"
  rd_save B1 "$b1"
  rd_policy "$RD_BPOL" policy_id=fixture-backup backup_max_age_s=3600 backup_retain_for_s=1 "backup_holds=$b1" pending_review_after_s=1
  rc2=0; rd_tool status --backup-root "$RD_ROOT" --backup-policy "$RD_BPOL" --receipts "$RD_RCPT" >"$RD_L/status-held2.jsonl" 2>&1 || rc2=$?
  rd_tool unfence --env-file "$RD_SRCENV" --backup-root "$RD_ROOT" >"$RD_L/unfence.log" 2>&1; local urc=$?
  local fp1; fp1="$(rd_fp $RD_SP)"
  rc3=0; rd_tool status --backup-root "$RD_ROOT" --backup-policy "$RD_BPOL" --receipts "$RD_RCPT" >"$RD_L/status-free.jsonl" 2>&1 || rc3=$?
  out2="$(docker_run_owned rd-stage --rm --label g001.run=$RUN_ID --network "${RD_SP}_data" --env-file <(printf 'PGPASSWORD=%s\n' "$(rd_ev src DB_RUNTIME_PASSWORD)") $RD_PGIMG psql -h postgres -U ${RD_DB}_runtime -d $RD_DB -tAc 'SELECT 1' 2>&1)"
  local putok=0; printf 'x' | rd_mc src svc "cat >/tmp/f && mc cp -q /tmp/f 'x/$(rd_ev src OBJECT_STORE_BUCKET)/objects/$orgA/after' && mc rm -q 'x/$(rd_ev src OBJECT_STORE_BUCKET)/objects/$orgA/after'" >/dev/null 2>&1 && putok=1
  local inside; inside="$(grep -c "$inflight" "$RD_ROOT/$b1/manifest.tsv" || true)"
  local present=0; rd_mc src admin "mc stat 'x/$(rd_ev src OBJECT_STORE_BUCKET)/$inflight'" >/dev/null 2>&1 && present=1
  rd_client_rm put rtw
  rd_ok RD-B-4 "$([[ $rc == 0 && -n "$b1" && -f "$RD_ROOT/$b1/COMPLETE" && "$(sed -n 's/^release\t//p' "$RD_ROOT/$b1/backup-receipt.tsv")" == deferred && $rc2 == 1 && $urc == 0 && $rc3 == 0 && "$out2" == 1 && $putok == 1 && "$fp0" == "$fp1" && $inside == 0 && $present == 0 ]] && echo 1 || echo 0)" \
    "B1 $b1 complete, release deferred, held until unfence ($rc2), unfence $urc, status ok ($rc3), runtime login $out2, PUT ok $putok, source fingerprint unchanged $([[ "$fp0" == "$fp1" ]] && echo yes || echo no), in-flight PUT absent (manifest $inside, source $present)"
  # (5) the trusted administrator window is a prerequisite.
  local nf0 ns0; nf0="$(ls "$RD_ROOT/fences" | wc -l)"; ns0="$(ls -d "$RD_ROOT"/*/ | wc -l)"
  rc=0; out="$(rd_tool fence --env-file "$RD_SRCENV" --backup-root "$RD_ROOT" 2>&1)" || rc=$?
  rc2=0; out2="$(rd_tool backup --env-file "$RD_SRCENV" --backup-root "$RD_ROOT" --ledger "$RD_LEDGER" 2>&1)" || rc2=$?
  rd_ok RD-B-5 "$([[ $rc == 3 && $rc2 == 3 && "$out" == *admin_window_unavailable* && "$out2" == *admin_window_unavailable* && "$(ls "$RD_ROOT/fences" | wc -l)" == "$nf0" && "$(ls -d "$RD_ROOT"/*/ | wc -l)" == "$ns0" ]] && echo 1 || echo 0)" "fence and backup without --admin-window refuse before any change"
  # Negatives.
  local neg=1 why=""
  rc=0; out="$(rd_tool backup --env-file "$RD_SRCENV" --backup-root "$RD_ROOT" --ledger "$RD_LEDGER" --use-fence 2>&1)" || rc=$?; [[ $rc != 0 && "$out" == *fence_not_established* ]] || { neg=0; why+="use-fence "; }
  docker stop "$(rd_pgc $RD_SP)" >/dev/null; rc=0; rd_tool backup --env-file "$RD_SRCENV" --backup-root "$RD_ROOT" --ledger "$RD_LEDGER" --admin-window CHG-RD-3 >/dev/null 2>&1 || rc=$?
  rd_dc src up -d --wait postgres >/dev/null 2>&1; [[ $rc != 0 ]] || { neg=0; why+="stopped-pg "; }
  rc=0; rd_tool backup --env-file "$RD_SRCENV" --backup-root "$RD_ROOT" --ledger "$RD_LEDGER" --admin-window CHG-RD-4 --backup-id "$b1" >/dev/null 2>&1 || rc=$?; [[ $rc == 3 ]] || { neg=0; why+="existing-set "; }
  : >"$RD_P/not-a-dir"; rc=0; rd_tool backup --env-file "$RD_SRCENV" --backup-root "$RD_P/not-a-dir" --ledger "$RD_LEDGER" --admin-window CHG-RD-5 >/dev/null 2>&1 || rc=$?; rm -f "$RD_P/not-a-dir"; [[ $rc != 0 ]] || { neg=0; why+="unwritable "; }
  docker pause "$(rd_mic $RD_SP)" >/dev/null; rc=0; rd_tool backup --env-file "$RD_SRCENV" --backup-root "$RD_ROOT" --ledger "$RD_LEDGER" --admin-window CHG-RD-6 >/dev/null 2>&1 || rc=$?; docker unpause "$(rd_mic $RD_SP)" >/dev/null; [[ $rc != 0 ]] || { neg=0; why+="paused-minio "; }
  rd_tool_bg backup --env-file "$RD_SRCENV" --backup-root "$RD_ROOT" --ledger "$RD_LEDGER" --admin-window CHG-RD-7 >"$RD_L/term.log" 2>&1 & local bp=$!
  for n in $(seq 120); do ls "$RD_ROOT"/*/database.dump >/dev/null 2>&1 && [[ -n "$(find "$RD_ROOT" -name database.dump -newer "$RD_ROOT/$b1/COMPLETE" 2>/dev/null)" ]] && break; sleep 0.5; done
  kill -TERM "$bp" 2>/dev/null; wait "$bp"; rc=$?; [[ $rc != 0 ]] || { neg=0; why+="term "; }
  sed 's/^OBJECT_STORE_BUCKET=.*/OBJECT_STORE_BUCKET=no-such-bucket/' "$RD_SRCENV" >"$RD_P/bad-bucket.env"; chmod 600 "$RD_P/bad-bucket.env"
  rc=0; rd_tool backup --env-file "$RD_P/bad-bucket.env" --backup-root "$RD_ROOT" --ledger "$RD_LEDGER" --admin-window CHG-RD-8 >/dev/null 2>&1 || rc=$?; [[ $rc != 0 ]] || { neg=0; why+="pipeline "; }
  # C2: an unexpected capture failure whose INCOMPLETE marker cannot be written still releases the fence (verified state),
  # and the command stays non-zero.
  local mk mko mkrc=0 mkf
  mk="00000000-0000-4000-8000-0000000000c2"
  mko="$(RECOVERY_DRILL_FAIL_CAPTURE=marker rd_tool backup --env-file "$RD_SRCENV" --backup-root "$RD_ROOT" --ledger "$RD_LEDGER" --admin-window CHG-RD-C2 --backup-id "$mk" 2>&1)" || mkrc=$?
  printf '%s\n' "$mko" >"$RD_L/rdb-marker.log"
  mkf="$(grep -l $'^admin_window\tCHG-RD-C2$' "$RD_ROOT"/fences/*.tsv 2>/dev/null | head -1)"
  [[ $mkrc != 0 && -n "$mkf" && -d "$RD_ROOT/$mk/INCOMPLETE" && ! -e "$RD_ROOT/$mk/COMPLETE" && "$mko" == *"INCOMPLETE marker could not be written"* && "$mko" == *"released (verified)"* ]] &&
    grep -q $'^released\t' "$mkf" || { neg=0; why+="marker-failure($mkrc) "; }
  rm -rf -- "${RD_ROOT:?}/$mk" || why+="marker-cleanup "
  local completes unreleased; completes="$(ls "$RD_ROOT"/*/COMPLETE | wc -l)"; unreleased=0
  for f in "$RD_ROOT"/fences/*.tsv; do grep -q $'^released\t' "$f" || unreleased=$((unreleased + 1)); done
  a="$(rd_sql src "SELECT string_agg(x::text, ',' ORDER BY x::text) FROM pg_database d, unnest(d.datacl) x WHERE d.datname = '$RD_DB'")"
  rd_ok RD-B-NEG "$([[ $neg == 1 && $completes == 1 && $unreleased == 0 && "$a" == "$prior_acl" ]] && echo 1 || echo 0)" "negatives exit non-zero incl. a capture failure whose INCOMPLETE marker cannot be written (fence released, verified) ${why:+(failed: $why)}; COMPLETE sets $completes; unreleased fences $unreleased; ACL restored"
  # (6) MR !29 D1: a TERM during a capture under an operator-established fence (--use-fence) never releases that fence;
  # the negatives' CHG-RD-7 case above is the integrated counterpart (released, counted in "unreleased fences 0").
  # Held means all of: the record has no released line, the runtime login is still refused CONNECT (the ACL differs
  # from the prior one), the service user is still disabled and refused by the provider, status reports held; the
  # command exits non-zero with INCOMPLETE (interrupted) and no COMPLETE. An explicit unfence then restores exactly the
  # prior state (ACL equal, runtime login accepted, service user enabled, record released).
  local iset irc=0 iff ia ib=0 ic=0 ir=0 ie=0 urc2 ia2 ib2=0 ic2=0 held=0 restored=0
  rd_tool fence --env-file "$RD_SRCENV" --backup-root "$RD_ROOT" --admin-window CHG-RD-I1 >"$RD_L/int-fence.log" 2>&1 || { rd_rec RD-B-INT FAIL "fence"; return 1; }
  iff="$(grep -l $'^admin_window\tCHG-RD-I1$' "$RD_ROOT"/fences/*.tsv 2>/dev/null | head -1)"
  iset="00000000-0000-4000-8000-0000000000d1"
  rd_tool_bg backup --env-file "$RD_SRCENV" --backup-root "$RD_ROOT" --ledger "$RD_LEDGER" --use-fence --backup-id "$iset" >"$RD_L/int-term.log" 2>&1 & local ip=$!
  for n in $(seq 120); do [[ -e "$RD_ROOT/$iset/database.dump" ]] && break; sleep 0.5; done
  kill -TERM "$ip" 2>/dev/null; wait "$ip"; irc=$?
  ia="$(rd_sql src "SELECT string_agg(x::text, ',' ORDER BY x::text) FROM pg_database d, unnest(d.datacl) x WHERE d.datname = '$RD_DB'")"
  out="$(docker_run_owned rd-stage --rm --label g001.run=$RUN_ID --network "${RD_SP}_data" --env-file <(printf 'PGPASSWORD=%s\n' "$(rd_ev src DB_RUNTIME_PASSWORD)") $RD_PGIMG psql -h postgres -U ${RD_DB}_runtime -d $RD_DB -c 'SELECT 1' 2>&1)"; [[ "$out" == *"permission denied for database"* ]] && ib=1
  rd_mc src admin "mc admin user info --json x core-service" | grep -q '"userStatus":"disabled"' && ic=1
  rd_mc src svc "mc ls 'x/$(rd_ev src OBJECT_STORE_BUCKET)/'" >/dev/null 2>&1 || ir=1
  rc2=0; rd_tool status --backup-root "$RD_ROOT" --backup-policy "$RD_BPOL" --receipts "$RD_RCPT" >"$RD_L/int-status.jsonl" 2>&1 || rc2=$?
  [[ $rc2 == 1 ]] && grep -q '"check":"backup.fence","status":"held"' "$RD_L/int-status.jsonl" && ie=1
  [[ $irc != 0 && -n "$iff" && ! -e "$RD_ROOT/$iset/COMPLETE" && "$(sed -n 's/^reason\t//p' "$RD_ROOT/$iset/INCOMPLETE" 2>/dev/null)" == interrupted && "$ia" != "$prior_acl" && $ib == 1 && $ic == 1 && $ir == 1 && $ie == 1 ]] &&
    ! grep -q $'^released\t' "$iff" && grep -q 'stays held' "$RD_L/int-term.log" && held=1
  rd_tool unfence --env-file "$RD_SRCENV" --backup-root "$RD_ROOT" >"$RD_L/int-unfence.log" 2>&1; urc2=$?
  ia2="$(rd_sql src "SELECT string_agg(x::text, ',' ORDER BY x::text) FROM pg_database d, unnest(d.datacl) x WHERE d.datname = '$RD_DB'")"
  out="$(docker_run_owned rd-stage --rm --label g001.run=$RUN_ID --network "${RD_SP}_data" --env-file <(printf 'PGPASSWORD=%s\n' "$(rd_ev src DB_RUNTIME_PASSWORD)") $RD_PGIMG psql -h postgres -U ${RD_DB}_runtime -d $RD_DB -tAc 'SELECT 1' 2>&1)"; [[ "$out" == 1 ]] && ib2=1
  rd_mc src admin "mc admin user info --json x core-service" | grep -q '"userStatus":"enabled"' && ic2=1
  [[ $urc2 == 0 && "$ia2" == "$prior_acl" && $ib2 == 1 && $ic2 == 1 ]] && grep -q $'^released\t' "$iff" && restored=1
  rd_ok RD-B-INT "$([[ $held == 1 && $restored == 1 ]] && echo 1 || echo 0)" "external --use-fence capture interrupted by TERM: exit $irc, no COMPLETE, INCOMPLETE $(sed -n 's/^reason\t//p' "$RD_ROOT/$iset/INCOMPLETE" 2>/dev/null); fence held (record unreleased, runtime CONNECT refused $ib, service user disabled $ic and refused $ir, status held $ie); unfence $urc2 restores the prior state (ACL equal $([[ "$ia2" == "$prior_acl" ]] && echo yes || echo no), runtime login $ib2, user enabled $ic2)"
}

rd_stage_rdk() {
  local f rc st
  # (1) SIGKILL during fence blocked at X4 against a paused MinIO.
  rd_tool_bg fence --env-file "$RD_SRCENV" --backup-root "$RD_ROOT" --admin-window CHG-RD-K1 >"$RD_L/k1.log" 2>&1 & local fp=$!
  for i in $(seq 120); do f="$(grep -l $'^locks_held\t' "$RD_ROOT"/fences/*.tsv 2>/dev/null | xargs -r grep -L $'^released\t' | head -1)"; [[ -n "$f" ]] && break; sleep 0.5; done
  docker pause "$(rd_mic $RD_SP)" >/dev/null; sleep 2
  kill -KILL "$fp" 2>/dev/null; wait "$fp" 2>/dev/null
  local holder; holder="$(docker ps -q --filter label=g001.run=$RUN_ID --filter label=g013.role=holder | wc -l)"
  docker unpause "$(rd_mic $RD_SP)" >/dev/null
  rd_tool unfence --env-file "$RD_SRCENV" --backup-root "$RD_ROOT" >"$RD_L/k1-unfence.log" 2>&1; rc=$?
  rd_ok RD-K-1 "$([[ -n "$f" && $holder == 1 && $rc == 0 && "$(docker ps -q --filter label=g001.run=$RUN_ID --filter label=g013.role=holder | wc -l)" == 0 ]] && grep -q $'^released\t' "$f" && rd_mc src admin "mc admin user info --json x core-service" | grep -q '"userStatus":"enabled"' && echo 1 || echo 0)" "holder alive after SIGKILL ($holder); unfence $rc restores the prior state"
  # (2) SIGKILL during unfence blocked at the service user against a paused MinIO; a second unfence completes.
  rd_tool fence --env-file "$RD_SRCENV" --backup-root "$RD_ROOT" --admin-window CHG-RD-K2 >"$RD_L/k2.log" 2>&1 || { rd_rec RD-K-2 FAIL fence; return 1; }
  f="$(ls -t "$RD_ROOT"/fences/*.tsv | head -1)"
  docker pause "$(rd_mic $RD_SP)" >/dev/null
  rd_tool_bg unfence --env-file "$RD_SRCENV" --backup-root "$RD_ROOT" >"$RD_L/k2-unfence.log" 2>&1 & local up=$!
  for i in $(seq 120); do grep -q $'^holder_released\t' "$f" && break; sleep 0.5; done
  sleep 2; kill -KILL "$up" 2>/dev/null; wait "$up" 2>/dev/null
  rc=0; rd_tool status --backup-root "$RD_ROOT" --backup-policy "$RD_BPOL" --receipts "$RD_RCPT" >"$RD_L/k2-status.jsonl" 2>&1 || rc=$?
  docker unpause "$(rd_mic $RD_SP)" >/dev/null
  rd_tool unfence --env-file "$RD_SRCENV" --backup-root "$RD_ROOT" >"$RD_L/k2-unfence2.log" 2>&1; local rc2=$?
  rd_ok RD-K-2 "$([[ $rc == 1 && $rc2 == 0 ]] && grep -q '"backup.fence","status":"held"' "$RD_L/k2-status.jsonl" && grep -q $'^released\t' "$f" && echo 1 || echo 0)" "status held after the killed unfence; second unfence $rc2"
  # (3) prior state preserved: user already disabled, retention CONNECT already revoked.
  rd_mc src admin "mc admin user disable x core-service" >/dev/null 2>&1
  rd_sql src "REVOKE CONNECT ON DATABASE $RD_DB FROM ${RD_DB}_retention" >/dev/null
  rd_tool fence --env-file "$RD_SRCENV" --backup-root "$RD_ROOT" --admin-window CHG-RD-K3 >"$RD_L/k3.log" 2>&1 && rd_tool unfence --env-file "$RD_SRCENV" --backup-root "$RD_ROOT" >>"$RD_L/k3.log" 2>&1; rc=$?
  local dis rn; dis="$(rd_mc src admin "mc admin user info --json x core-service" | sed -n 's/.*"userStatus":"\([a-z]*\)".*/\1/p')"
  rn="$(rd_sql src "SELECT has_database_privilege('${RD_DB}_retention', '$RD_DB', 'CONNECT')")"
  rd_mc src admin "mc admin user enable x core-service" >/dev/null 2>&1; rd_sql src "GRANT CONNECT ON DATABASE $RD_DB TO ${RD_DB}_retention" >/dev/null
  rd_ok RD-K-3 "$([[ $rc == 0 && "$dis" == disabled && "$rn" == f ]] && echo 1 || echo 0)" "user stayed $dis, retention CONNECT $rn after fence+unfence; pre-case state restored"
}

rd_stage_rdp() {
  rd_go src post.log '^TestRecoveryPost$' -e RECOVERY_IDS="$IDS" || true
  rd_gopass TestRecoveryPost post.log && POST="$(sed -n 's/^RECOVERY-POST //p' "$RD_L/post.log" | head -1)" || { rd_rec RD-P FAIL post; tail -30 "$RD_L/post.log"; return 1; }
  rd_save POST "$POST"
  local refs
  refs="organization=$(rd_j "$POST" RemovedBinding organization),subject=$(rd_j "$POST" RemovedBinding subject),scope_type=group,scope_id=$(rd_j "$POST" RemovedBinding scope_id),role=$(rd_j "$POST" RemovedBinding role)
organization=$(rd_j "$POST" DowngradedBinding organization),subject=$(rd_j "$POST" DowngradedBinding subject),scope_type=group,scope_id=$(rd_j "$POST" DowngradedBinding scope_id),from_role=$(rd_j "$POST" DowngradedBinding from_role),to_role=$(rd_j "$POST" DowngradedBinding to_role)
organization=$(rd_j "$POST" OrgA),group=$(rd_j "$POST" A2),profile=$(rd_j "$POST" P)"
  rd_tool ledger append --ledger "$RD_LEDGER" --kind membership.remove --refs "$(sed -n 1p <<<"$refs")" --by INC-RD >/dev/null &&
  rd_tool ledger append --ledger "$RD_LEDGER" --kind membership.downgrade --refs "$(sed -n 2p <<<"$refs")" --by INC-RD >/dev/null &&
  rd_tool ledger append --ledger "$RD_LEDGER" --kind profile.unlink --refs "$(sed -n 3p <<<"$refs")" --by INC-RD >/dev/null || { rd_rec RD-P FAIL append; return 1; }
  local s; s="$(rd_tool ledger sync --ledger "$RD_LEDGER" --env-file "$RD_SRCENV" --backup-root "$RD_ROOT" 2>&1)" || { rd_rec RD-P FAIL "sync $s"; return 1; }
  local orgA objA; orgA="$(rd_j "$POST" OrgA)"
  sleep 2
  rd_policy "$RD_P/purge-src.policy" policy_id=fixture-purge-src policy_label=fixture "scope=organization:$orgA" object_retain_for_s=1 holds= batch_size=500 max_batches=10 statement_timeout_ms=5000 lock_timeout_ms=1000 run_timeout_s=120
  local d a; d="$(rd_tool retention objects --policy "$RD_P/purge-src.policy" --env-file "$RD_SRCENV" --preview "$RD_P/purge-src.preview" --ledger "$RD_LEDGER" 2>&1)"
  a="$(rd_tool retention objects --policy "$RD_P/purge-src.policy" --env-file "$RD_SRCENV" --preview "$RD_P/purge-src.preview" --ledger "$RD_LEDGER" --apply --receipts "$RD_P/src-receipts" 2>&1)"; local arc=$?
  objA="$(rd_j "$POST" ObjA)"
  local gone; gone="$(rd_sql src "SELECT count(*) FROM app.stored_objects WHERE id = '$objA'")"
  local rc=0 out; out="$(rd_tool backup --env-file "$RD_SRCENV" --backup-root "$RD_ROOT" --ledger "$RD_LEDGER" --admin-window CHG-RD-B2 2>&1)" || rc=$?
  local b2; b2="$(sed -n 's/^backup\t\([0-9a-f-]*\)\tcomplete.*/\1/p' <<<"$out")"; rd_save B2 "$b2"
  rd_ok RD-P "$([[ "$gone" == 0 && $arc == 0 && $rc == 0 && -n "$b2" ]] && echo 1 || echo 0)" "six kinds applied (post test), 3 operator appends, sync [$s], source purge [$d | $a] removed the deleted A object ($gone left), B2 $b2"
}

rd_stage_rdi() {
  local s="$RD_ROOT/$B2" rc
  rd_flip "$s/database.dump" mid
  rc=0; rd_tool status --backup-root "$RD_ROOT" --backup-policy "$RD_BPOL" --receipts "$RD_RCPT" --sink "$RD_SINK" >"$RD_L/status-incident.jsonl" 2>&1 || rc=$?
  rd_tool quarantine --backup "$s" --reason storage_corruption >/dev/null
  local r2=0; rd_tool restore --backup "$s" --target-env-file "$RD_TGTENV" --receipts "$RD_RCPT" --backup-policy "$RD_BPOL" >"$RD_L/restore-b2.log" 2>&1 || r2=$?
  local v; v="$(rd_tool verify --backup "$s" 2>&1)"
  rd_ok RD-I "$([[ $rc == 1 && $r2 == 3 ]] && grep -q '"backup.integrity","status":"failed"' "$RD_L/status-incident.jsonl" && grep -q '"backup.integrity","status":"failed"' "$RD_SINK" && [[ "$v" == *database.dump* ]] && echo 1 || echo 0)" "status exit $rc with integrity failed (sink line), quarantine, restore refused ($r2), verify names database.dump"
}

rd_stage_rdr() {
  rd_dc tgt up -d --wait postgres minio redis >"$RD_L/up-tgt.log" 2>&1 || { rd_rec RD-R FAIL up-tgt; return 1; }
  docker_run_owned rd-stage --rm -i --label g001.run=$RUN_ID --network "${RD_TP}_data" --env-file <(grep -E '^(MINIO_ROOT_(USER|PASSWORD)|OBJECT_STORE_(BUCKET|ACCESS_KEY|SECRET_KEY))=' "$RD_TGTENV") "$RD_TP-minio-init" sh -s <"$RD_M/infra/minio/init.sh" >"$RD_L/init-tgt.log" 2>&1 || { rd_rec RD-R FAIL init-tgt; return 1; }
  local s1="$RD_ROOT/$B1" all=1 why="" rc v
  rr() { rc=0; rd_tool restore --backup "$1" --target-env-file "${2:-$RD_M/.env.rec}" --receipts "$RD_RCPT" --backup-policy "$RD_BPOL" >"$RD_L/rr-$3.log" 2>&1 || rc=$?; [[ $rc == 3 ]] && grep -q "reason=$4" "$RD_L/rr-$3.log" || { all=0; why+="$3($rc) "; }; }
  rr "$s1" "$RD_SRCENV" alias target_alias
  rd_sql tgt "CREATE TABLE app.not_fresh (id int)" >/dev/null; rr "$s1" "" nonempty target_not_fresh; rd_sql tgt "DROP TABLE app.not_fresh" >/dev/null
  mkset() { rm -rf "${RD_ROOT:?}/${1:?}"; cp -a "$s1" "$RD_ROOT/$1"; }
  mkset 00000000-0000-4000-8000-00000000000a; rm "$RD_ROOT/00000000-0000-4000-8000-00000000000a/database.dump"; rr "$RD_ROOT/00000000-0000-4000-8000-00000000000a" "" missing set_invalid
  mkset 00000000-0000-4000-8000-00000000000b; rd_flip "$RD_ROOT/00000000-0000-4000-8000-00000000000b/objects.tar" mid; rr "$RD_ROOT/00000000-0000-4000-8000-00000000000b" "" corrupt set_invalid
  mkset 00000000-0000-4000-8000-00000000000c; cp "$RD_ROOT/$B2/objects.tar" "$RD_ROOT/00000000-0000-4000-8000-00000000000c/objects.tar"; rr "$RD_ROOT/00000000-0000-4000-8000-00000000000c" "" mixed set_invalid
  mkset 00000000-0000-4000-8000-00000000000d; sed -i '1s/g013-backup\/1/g013-backup\/9/' "$RD_ROOT/00000000-0000-4000-8000-00000000000d/manifest.tsv"; sha256sum "$RD_ROOT/00000000-0000-4000-8000-00000000000d/manifest.tsv" | cut -d' ' -f1 >"$RD_ROOT/00000000-0000-4000-8000-00000000000d/COMPLETE"; rr "$RD_ROOT/00000000-0000-4000-8000-00000000000d" "" format set_invalid
  mkset 00000000-0000-4000-8000-00000000000e; rm "$RD_ROOT/00000000-0000-4000-8000-00000000000e/COMPLETE"; rr "$RD_ROOT/00000000-0000-4000-8000-00000000000e" "" incomplete set_invalid
  rr "$RD_ROOT/$B2" "" quarantined set_quarantined
  mkdir -p "$RD_P/reused"; chmod 700 "$RD_P/reused"; printf 'format\tg013-invalidate/1\nstatus\tok\n' >"$RD_P/reused/invalidate.tsv"
  rc=0; rd_tool restore --backup "$s1" --target-env-file "$RD_TGTENV" --receipts "$RD_P/reused" --backup-policy "$RD_BPOL" >"$RD_L/rr-reused.log" 2>&1 || rc=$?
  [[ $rc == 3 ]] && grep -q reason=receipts_reused "$RD_L/rr-reused.log" || { all=0; why+="reused($rc) "; }
  rm -rf "$RD_P/reused"
  for x in a b c d e; do rm -rf "$RD_ROOT/00000000-0000-4000-8000-00000000000$x"; done
  local nrel nobj; nrel="$(rd_sql tgt "SELECT count(*) FROM pg_class WHERE relnamespace IN ('app'::regnamespace, 'audit'::regnamespace)")"
  nobj="$(rd_mc tgt admin "mc ls --recursive 'x/$(rd_ev tgt OBJECT_STORE_BUCKET)/'" | grep -c . || true)"
  rd_ok RD-R "$([[ $all == 1 && "$nrel" == 0 && "$nobj" == 0 ]] && echo 1 || echo 0)" "9 refusals before mutation (incl. a reused receipts directory) ${why:+(failed: $why)}; target relations $nrel, objects $nobj"
}

rd_stage_rdf() {
  local s1="$RD_ROOT/$B1" rc rc2 rc3 rc4 rc5 r6
  rc=0; rd_tool restore --backup "$s1" --target-env-file "$RD_TGTENV" --receipts "$RD_RCPT" --backup-policy "$RD_BPOL" --steps database >"$RD_L/rdf-db.log" 2>&1 || rc=$?
  docker stop "$(rd_mic $RD_TP)" >/dev/null
  rc2=0; rd_tool restore --backup "$s1" --target-env-file "$RD_TGTENV" --receipts "$RD_RCPT" --backup-policy "$RD_BPOL" --steps objects >"$RD_L/rdf-obj1.log" 2>&1 || rc2=$?
  rc3=0; rd_tool reopen --backup "$s1" --ledger "$RD_LEDGER" --receipts "$RD_RCPT" --sign-off X --runtime-verification pass >"$RD_L/rdf-gate.log" 2>&1 || rc3=$?
  rd_dc tgt up -d --wait minio >/dev/null 2>&1
  local sv="$RD_ROOT/00000000-0000-4000-8000-0000000000f0"; rm -rf "$sv"; cp -a "$s1" "$sv"
  sed -i "s/^backup_id\t.*/backup_id\t00000000-0000-4000-8000-0000000000f0/" "$sv/manifest.tsv"; sha256sum "$sv/manifest.tsv" | cut -d' ' -f1 >"$sv/COMPLETE"
  rd_policy "$RD_BPOL" policy_id=fixture-backup backup_max_age_s=3600 backup_retain_for_s=1 "backup_holds=$B1,00000000-0000-4000-8000-0000000000f0" pending_review_after_s=1
  rc4=0; rd_tool restore --backup "$sv" --target-env-file "$RD_TGTENV" --receipts "$RD_RCPT" --backup-policy "$RD_BPOL" --steps objects >"$RD_L/rdf-other.log" 2>&1 || rc4=$?
  rd_policy "$RD_BPOL" policy_id=fixture-backup backup_max_age_s=3600 backup_retain_for_s=1 "backup_holds=$B1" pending_review_after_s=1
  rm -rf "$sv"
  rc5=0; rd_tool restore --backup "$s1" --target-env-file "$RD_SRCENV" --receipts "$RD_RCPT" --backup-policy "$RD_BPOL" --steps objects >"$RD_L/rdf-src.log" 2>&1 || rc5=$?
  local fname fsize r7=0; fname="$(awk -F'\t' '$1 == "object" { print $2; exit }' "$s1/manifest.tsv")"; fsize="$(awk -F'\t' '$1 == "object" { print $3; exit }' "$s1/manifest.tsv")"
  head -c "$fsize" /dev/zero | rd_mc tgt admin "cat >/tmp/f && mc cp -q /tmp/f 'x/$(rd_ev tgt OBJECT_STORE_BUCKET)/$fname'" >/dev/null 2>&1
  rd_tool restore --backup "$s1" --target-env-file "$RD_TGTENV" --receipts "$RD_RCPT" --backup-policy "$RD_BPOL" --steps objects >"$RD_L/rdf-foreign.log" 2>&1 || r7=$?
  rd_mc tgt admin "mc rm -q 'x/$(rd_ev tgt OBJECT_STORE_BUCKET)/$fname'" >/dev/null 2>&1
  r6=0; rd_tool restore --backup "$s1" --target-env-file "$RD_TGTENV" --receipts "$RD_RCPT" --backup-policy "$RD_BPOL" --steps objects >"$RD_L/rdf-obj2.log" 2>&1 || r6=$?
  rd_ok RD-F "$([[ $rc == 0 && $rc2 != 0 && $rc3 == 3 && $rc4 == 3 && $rc5 == 3 && $r7 == 3 && $r6 == 0 ]] && grep -q reason=continuation_mismatch "$RD_L/rdf-other.log" && grep -q reason=continuation_mismatch "$RD_L/rdf-foreign.log" && grep -q restored-closed "$RD_L/rdf-obj2.log" && echo 1 || echo 0)" \
    "database step $rc; objects with MinIO stopped $rc2 (receipt $(sed -n 's/^status\t//p' "$RD_RCPT/restore.tsv" | tail -1)); gate $rc3; other checkpoint $rc4; source as target $rc5; foreign object on the target $r7; continuation $r6"
}

rd_stage_rdv() {
  local m="$RD_ROOT/$B1/manifest.tsv" out ok1=1
  out="$(docker_run_owned rd-stage --rm --label g001.run=$RUN_ID --network "${RD_TP}_data" --env-file <(printf 'PGPASSWORD=%s\n' "$(rd_ev src DB_RUNTIME_PASSWORD)") $RD_PGIMG psql -h postgres -U ${RD_DB}_runtime -d $RD_DB -c 'SELECT 1' 2>&1)"
  [[ "$out" == *"password authentication failed"* ]] || ok1=0
  rd_ok RD-V "$([[ $ok1 == 1 && "$(sed -n 's/^redis\t//p' "$RD_RCPT/restore.tsv")" == empty && "$(sed -n 's/^database\t//p' "$RD_RCPT/restore.tsv")" == "done" ]] && echo 1 || echo 0)" "fingerprints and catalog equal (restore verification), source runtime value refused on the target, Redis $(sed -n 's/^redis\t//p' "$RD_RCPT/restore.tsv")"
}

rd_stage_rda() {
  local out rc=0; out="$(rd_tool invalidate-auth --target-env-file "$RD_TGTENV" --receipts "$RD_RCPT" 2>&1)" || rc=$?
  # G062 SUP-11: the receipt carries the signup expiry count, and the seeded pre-checkpoint signup row was expired.
  rd_ok RD-A "$([[ $rc == 0 && "$out" == *"live_sessions=0"* && "$out" =~ signup_sessions_expired=[1-9] ]] &&
    grep -qP '^signup_sessions_expired\t[1-9][0-9]*$' "$RD_RCPT/invalidate.tsv" && echo 1 || echo 0)" "$out"
}

rd_stage_rdl() {
  local s1="$RD_ROOT/$B1" all=1 why="" rc
  gate() { rc=0; rd_tool reopen --backup "$s1" --ledger "$1" --receipts "$RD_RCPT" --sign-off X --runtime-verification pass >"$RD_L/gate-$2.log" 2>&1 || rc=$?; [[ $rc == 3 ]] || { all=0; why+="$2($rc) "; }; }
  gate "$RD_LEDGER" no-attestation
  gate "$RD_P/ledger/missing.tsv" missing-ledger
  cp "$RD_LEDGER" "$RD_P/ledger/tampered.tsv"; sed -i '3s/\tsync\t/\toperator\t/' "$RD_P/ledger/tampered.tsv"; gate "$RD_P/ledger/tampered.tsv" tampered
  local mark; mark="$(sed -n 's/^ledger\t[^\t]*\t\([0-9]*\)\t.*/\1/p' "$RD_ROOT/$B1/manifest.tsv")"
  head -n "$mark" "$RD_LEDGER" >"$RD_P/ledger/stale.tsv"; gate "$RD_P/ledger/stale.tsv" stale
  local head; head="$(($(wc -l <"$RD_LEDGER") - 1))"
  rd_tool ledger attest --ledger "$RD_LEDGER" --through-seq "$head" --basis last-sync --by INC-RD >/dev/null; gate "$RD_LEDGER" last-sync-no-owner
  rd_tool ledger attest --ledger "$RD_LEDGER" --through-seq "$head" --basis final-sync --by INC-RD >/dev/null; gate "$RD_LEDGER" before-replay
  rm -f "$RD_P/ledger/"{tampered,stale}.tsv
  cp "$RD_RCPT/invalidate.tsv" "$RD_P/inv.bak"; sed -i 's/^restore_binding\t.*/restore_binding\t0000/' "$RD_RCPT/invalidate.tsv"
  gate "$RD_LEDGER" stale-invalidation; grep -q reason=receipt_unbound "$RD_L/gate-stale-invalidation.log" || { all=0; why+="stale-invalidation-reason "; }
  mv "$RD_P/inv.bak" "$RD_RCPT/invalidate.tsv"
  rd_ok RD-L "$all" "7 gate refusals (incl. an invalidation receipt of another restore) ${why:+(failed: $why)}"
}

rd_stage_rdy() {
  rd_sql tgt "CREATE FUNCTION public.g013_fault() RETURNS trigger LANGUAGE plpgsql AS \$\$ BEGIN RAISE EXCEPTION 'injected audit fault'; END \$\$; CREATE TRIGGER g013_fault BEFORE INSERT ON audit.events FOR EACH ROW WHEN (NEW.action = 'user.disabled') EXECUTE FUNCTION public.g013_fault()" >/dev/null
  docker stop "$(rd_mic $RD_TP)" >/dev/null
  local rc=0 out; out="$(rd_tool replay --backup "$RD_ROOT/$B1" --ledger "$RD_LEDGER" --target-env-file "$RD_TGTENV" --receipts "$RD_RCPT" 2>&1)" || rc=$?
  local st=0; rd_tool status --backup-root "$RD_ROOT" --backup-policy "$RD_BPOL" --receipts "$RD_RCPT" >"$RD_L/rdy-status.jsonl" 2>&1 || st=$?
  local g=0; rd_tool reopen --backup "$RD_ROOT/$B1" --ledger "$RD_LEDGER" --receipts "$RD_RCPT" --sign-off X --runtime-verification pass >/dev/null 2>&1 || g=$?
  local d1; d1="$(rd_j "$POST" Users disable1)"
  local dis; dis="$(rd_sql tgt "SELECT status FROM app.users WHERE id = '$d1'")"
  rd_sql tgt "DROP TRIGGER g013_fault ON audit.events; DROP FUNCTION public.g013_fault()" >/dev/null
  local cat; cat="$(docker exec -i "$(rd_pgc $RD_TP)" psql -X -q -tA -U postgres -d $RD_DB -v db=$RD_DB <"$RD_M/infra/recovery/sql/fingerprint-catalog.sql" | sed -n 's/^catalog\t//p')"
  rd_dc tgt up -d --wait minio >/dev/null 2>&1
  local rc2=0 out2; out2="$(rd_tool replay --backup "$RD_ROOT/$B1" --ledger "$RD_LEDGER" --target-env-file "$RD_TGTENV" --receipts "$RD_RCPT" 2>&1)" || rc2=$?
  rprobe() { docker exec -i "$(rd_pgc $RD_TP)" psql -X -q -tA -v ON_ERROR_STOP=1 -U postgres -d $RD_DB -v kind="$1" -v org="$2" -v grp="$3" -v profile="$4" \
    -v subject="$5" -v scope_type="$6" -v scope_id="$7" -v role="$8" -v from_role=- -v to_role=- <"$RD_M/infra/recovery/sql/replay-restrict.sql" | sed -n 's/^state\t//p'; }
  local zz pr1 pr2 pr3; zz="$(cat /proc/sys/kernel/random/uuid)"
  pr1="$(rprobe profile.unlink "$(rd_j "$POST" OrgA)" "$zz" "$(rd_j "$POST" P)" - - - -)"
  pr2="$(rprobe membership.remove "$(rd_j "$POST" RemovedBinding organization)" - - "$zz" group "$(rd_j "$POST" RemovedBinding scope_id)" "$(rd_j "$POST" RemovedBinding role)")"
  pr3="$(rprobe profile.unlink "$(rd_j "$POST" OrgA)" "$(rd_j "$POST" A2)" "$(rd_j "$POST" P)" - - - -)"
  rd_ok RD-Y "$([[ $rc == 4 && $st == 1 && $g == 3 && "$dis" == disabled && "$cat" == "$(sed -n 's/^catalog\t//p' "$RD_ROOT/$B1/manifest.tsv")" && $rc2 == 0 && "$pr1" == absent && "$pr2" == absent && "$pr3" == unlinked ]] && grep -q '"replay.last","status":"gap"' "$RD_L/rdy-status.jsonl" && grep -q $'^gap\t.*independent=recorded' "$RD_RCPT/replay.tsv" && echo 1 || echo 0)" \
    "first [$out] status $st gap, gate $g, disable1 $dis, fault trigger dropped and catalog equal; second [$out2]; wrong group $pr1, wrong subject $pr2, applied unlink $pr3"
}

rd_stage_rdg() {
  local end; end="$(sed -n 's/^window_db\t[^\t]*\t//p' "$RD_ROOT/$B1/manifest.tsv")"
  rd_go tgt verify.log '^TestRecoveryVerify$' -e RECOVERY_IDS="$POST" -e RECOVERY_CHECKPOINT_END="$end" || true
  rd_gopass TestRecoveryVerify verify.log && VERIFIED=pass || VERIFIED=fail
  rd_save VERIFIED "$VERIFIED"
  rd_ok RD-G "$([[ $VERIFIED == pass ]] && echo 1 || echo 0)" "TestRecoveryVerify $(grep -c -- '--- PASS: TestRecoveryVerify/' "$RD_L/verify.log") subtests passed"
  [[ $VERIFIED == pass ]] || grep -E -- '--- FAIL|_test.go:[0-9]+' "$RD_L/verify.log" | head -20
}

rd_stage_rdo() {
  # C1: a restriction appended and attested after the replay blocks reopen until it is replayed; a ledger change during a
  # replay makes that replay incomplete; replaying the current head then passes and applies the restriction.
  local mrefs h rcs=0 rcd=0 rcn=0 rcr=0 bnd c1=1 c1why=""
  mrefs="organization=$(rd_j "$POST" OrgA),subject=$(rd_j "$POST" Users memberA2),scope_type=group,scope_id=$(rd_j "$POST" A2),role=group_member"
  rd_tool ledger append --ledger "$RD_LEDGER" --kind membership.remove --refs "$mrefs" --by INC-RD >/dev/null || { rd_rec RD-O FAIL "append $mrefs"; return 1; }
  h="$(($(wc -l <"$RD_LEDGER") - 1))"; rd_tool ledger attest --ledger "$RD_LEDGER" --through-seq "$h" --basis final-sync --by INC-RD >/dev/null
  rd_tool reopen --backup "$RD_ROOT/$B1" --ledger "$RD_LEDGER" --receipts "$RD_RCPT" --sign-off SIGN-RD --runtime-verification "$VERIFIED" >"$RD_L/rdo-stale.log" 2>&1 || rcs=$?
  [[ $rcs == 3 ]] && grep -q reason=replay_stale "$RD_L/rdo-stale.log" || { c1=0; c1why+="stale-gate($rcs) "; }
  RECOVERY_DRILL_PAUSE_IN_REPLAY=8 rd_tool_bg replay --backup "$RD_ROOT/$B1" --ledger "$RD_LEDGER" --target-env-file "$RD_TGTENV" --receipts "$RD_RCPT" >"$RD_L/rdo-drift.log" 2>&1 & local rp=$!
  sleep 3; rd_tool ledger append --ledger "$RD_LEDGER" --kind membership.remove --refs "$mrefs" --by INC-RD >/dev/null || c1why+="drift-append "
  wait "$rp"; rcd=$?
  [[ $rcd == 4 ]] && grep -q reason=ledger_drift "$RD_L/rdo-drift.log" && grep -q $'^status\tincomplete' "$RD_RCPT/replay.tsv" || { c1=0; c1why+="drift($rcd) "; }
  h="$(($(wc -l <"$RD_LEDGER") - 1))"; rd_tool ledger attest --ledger "$RD_LEDGER" --through-seq "$h" --basis final-sync --by INC-RD >/dev/null
  rd_tool reopen --backup "$RD_ROOT/$B1" --ledger "$RD_LEDGER" --receipts "$RD_RCPT" --sign-off SIGN-RD --runtime-verification "$VERIFIED" >"$RD_L/rdo-drift-gate.log" 2>&1 || rcn=$?
  [[ $rcn == 3 ]] || { c1=0; c1why+="incomplete-gate($rcn) "; }
  rd_tool replay --backup "$RD_ROOT/$B1" --ledger "$RD_LEDGER" --target-env-file "$RD_TGTENV" --receipts "$RD_RCPT" >"$RD_L/rdo-replay2.log" 2>&1 || rcr=$?
  bnd="$(rd_sql tgt "SELECT count(*) FROM app.role_bindings WHERE subject_user_id = '$(rd_j "$POST" Users memberA2)' AND scope_type = 'group' AND scope_id = '$(rd_j "$POST" A2)'")"
  [[ $rcr == 0 && "$bnd" == 0 && "$(sed -n 's/^ledger_through_seq\t//p' "$RD_RCPT/replay.tsv")" == "$h" ]] || { c1=0; c1why+="replay-head($rcr,$bnd) "; }
  local rc=0 out; out="$(rd_tool reopen --backup "$RD_ROOT/$B1" --ledger "$RD_LEDGER" --receipts "$RD_RCPT" --sign-off SIGN-RD --runtime-verification "$VERIFIED" 2>&1)" || rc=$?
  local rc2=0 o2; o2="$(rd_tool backup --env-file "$RD_TGTENV" --backup-root "$RD_ROOT" --ledger "$RD_LEDGER" --admin-window CHG-RD-B3 2>&1)" || rc2=$?
  local b3; b3="$(sed -n 's/^backup\t\([0-9a-f-]*\)\tcomplete.*/\1/p' <<<"$o2")"; rd_save B3 "$b3"
  local st=0; rd_tool status --backup-root "$RD_ROOT" --backup-policy "$RD_BPOL" --receipts "$RD_RCPT" --sink "$RD_SINK" >"$RD_L/rdo-status.jsonl" 2>&1 || st=$?
  rd_ok RD-O "$([[ $c1 == 1 && $rc == 0 && "$out" == *"gaps="* && $rc2 == 0 && -n "$b3" && $st == 0 ]] && grep -q '"check":"summary","status":"ok"' "$RD_SINK" && echo 1 || echo 0)" \
    "appended+attested after replay: reopen $rcs (replay_stale); drift during replay $rcd (ledger_drift), gate $rcn; replay of the head $rcr, binding left $bnd ${c1why:+(failed: $c1why)}; [$out]; B3 [$o2]; status exit $st"
}

rd_stage_mda() {
  local orgA cross tmid all=1 why="" d a rc
  orgA="$(rd_j "$POST" OrgA)"; cross="$(rd_j "$POST" Users cross)"; tmid="$(rd_j "$POST" TMid)"
  local held recent bcount fp0 fp1
  held="$(rd_evcount "organization_id = '$orgA' AND occurred_at < '$tmid' AND ARRAY[organization_id, group_id, actor_user_id, subject_user_id, resource_id] && ARRAY['$cross'::uuid]")"
  recent="$(rd_evcount "organization_id = '$orgA' AND occurred_at >= '$tmid'")"
  bcount="$(rd_evcount "organization_id = '$(rd_j "$POST" OrgB)'")"
  pol() { rd_policy "$RD_P/$1" policy_id="$1" policy_label=fixture "scope=organization:$orgA" "audit_cutoff=$tmid" "holds=$cross" batch_size=2 "max_batches=$2" statement_timeout_ms=5000 lock_timeout_ms=1000 run_timeout_s=120; }
  pol mda.policy 2
  fp0="$(rd_fp $RD_TP)"
  d="$(rd_tool retention audit --policy "$RD_P/mda.policy" --env-file "$RD_TGTENV" --preview "$RD_P/mda.preview" 2>&1)" || { all=0; why+="dry-run "; }
  fp1="$(rd_fp $RD_TP)"; [[ "$fp0" == "$fp1" ]] || { all=0; why+="dry-run-changed "; }
  rc=0; a="$(rd_tool retention audit --policy "$RD_P/mda.policy" --env-file "$RD_TGTENV" --preview "$RD_P/mda.preview" --apply --receipts "$RD_RCPT" 2>&1)" || rc=$?
  [[ $rc == 1 && "$a" == *"partial"*"deleted=4"* ]] || { all=0; why+="bounded($rc) "; }
  printf 'batch_size=3\n' >>"$RD_P/mda.policy"; rc=0; rd_tool retention audit --policy "$RD_P/mda.policy" --env-file "$RD_TGTENV" --preview "$RD_P/mda.preview" --apply --receipts "$RD_RCPT" >/dev/null 2>&1 || rc=$?
  [[ $rc != 0 ]] || { all=0; why+="changed-policy "; }
  pol mda2.policy 1000
  rd_tool retention audit --policy "$RD_P/mda2.policy" --env-file "$RD_TGTENV" --preview "$RD_P/mda2.preview" >/dev/null 2>&1
  # A held row lock (superuser) on an eligible row: a batch fails on lock_timeout, truthfully; the retry succeeds.
  local lockid; lockid="$(rd_sql tgt "SELECT id FROM audit.events WHERE organization_id = '$orgA' AND occurred_at < '$tmid' AND NOT (ARRAY[organization_id, group_id, actor_user_id, subject_user_id, resource_id] && ARRAY['$cross'::uuid]) ORDER BY occurred_at, id LIMIT 1")"
  docker_run_owned rd-stage -d --label g001.run=$RUN_ID --name g013-rd-$RUN_ID-lock --network "${RD_TP}_data" --env-file <(printf 'PGPASSWORD=%s\n' "$(rd_ev tgt POSTGRES_SUPERUSER_PASSWORD)") $RD_PGIMG \
    psql -h postgres -U postgres -d $RD_DB -c "BEGIN; SELECT id FROM audit.events WHERE id = '$lockid' FOR UPDATE; SELECT pg_sleep(60); ROLLBACK;" >/dev/null
  sleep 2; rc=0; a="$(rd_tool retention audit --policy "$RD_P/mda2.policy" --env-file "$RD_TGTENV" --preview "$RD_P/mda2.preview" --apply --receipts "$RD_RCPT" 2>&1)" || rc=$?
  [[ $rc != 0 && "$a" != *complete* ]] || { all=0; why+="lock-timeout($rc) "; }
  rd_sql tgt "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE query LIKE '%FOR UPDATE; SELECT pg_sleep(60)%' AND pid <> pg_backend_pid()" >/dev/null; rd_client_rm lock
  rd_tool retention audit --policy "$RD_P/mda2.policy" --env-file "$RD_TGTENV" --preview "$RD_P/mda2.preview" >/dev/null 2>&1
  rc=0; a="$(rd_tool retention audit --policy "$RD_P/mda2.policy" --env-file "$RD_TGTENV" --preview "$RD_P/mda2.preview" --apply --receipts "$RD_RCPT" 2>&1)" || rc=$?
  [[ $rc == 0 && "$a" == *complete* ]] || { all=0; why+="complete($rc) "; }
  local left; left="$(rd_evcount "organization_id = '$orgA' AND occurred_at < '$tmid'")"
  [[ "$left" == "$held" && "$(rd_evcount "organization_id = '$orgA' AND occurred_at >= '$tmid'")" -ge "$recent" && "$(rd_evcount "organization_id = '$(rd_j "$POST" OrgB)'")" -ge "$bcount" ]] || { all=0; why+="preserved "; }
  # Global scope: without authorization refused; with it, only eligible global rows.
  rd_policy "$RD_P/mdg.policy" policy_id=mdg policy_label=fixture scope=global "audit_cutoff=$tmid" holds= batch_size=500 max_batches=100 statement_timeout_ms=5000 lock_timeout_ms=1000 run_timeout_s=120
  rc=0; rd_tool retention audit --policy "$RD_P/mdg.policy" --env-file "$RD_TGTENV" --preview "$RD_P/mdg.preview" >/dev/null 2>&1 || rc=$?; [[ $rc == 2 ]] || { all=0; why+="global-unauth($rc) "; }
  printf 'global_authorization=OWNER-RD\n' >>"$RD_P/mdg.policy"
  local gold gnew; gold="$(rd_evcount "organization_id IS NULL AND occurred_at < '$tmid'")"; gnew="$(rd_evcount "organization_id IS NULL AND occurred_at >= '$tmid'")"
  rd_tool retention audit --policy "$RD_P/mdg.policy" --env-file "$RD_TGTENV" --preview "$RD_P/mdg.preview" >/dev/null 2>&1 &&
    rd_tool retention audit --policy "$RD_P/mdg.policy" --env-file "$RD_TGTENV" --preview "$RD_P/mdg.preview" --apply --receipts "$RD_RCPT" >/dev/null 2>&1 || { all=0; why+="global "; }
  [[ "$(rd_evcount "organization_id IS NULL AND occurred_at < '$tmid'")" == 0 && "$(rd_evcount "organization_id IS NULL AND occurred_at >= '$tmid'")" -ge "$gnew" && "$(rd_evcount "organization_id IS NOT NULL")" -gt 0 ]] || { all=0; why+="global-scope "; }
  rd_ok MD-A "$all" "dry-run unchanged, bounded partial, changed policy refused, lock timeout truthful then retry complete, held $held kept, recent/B/global preserved, global scope only with authorization (deleted $gold) ${why:+(failed: $why)}"
}

rd_stage_mds0() { # status --env-file before repair: pending_review pending
  local rc=0; rd_tool status --backup-root "$RD_ROOT" --backup-policy "$RD_BPOL" --receipts "$RD_RCPT" --env-file "$RD_TGTENV" --ledger "$RD_LEDGER" >"$RD_L/mds0.jsonl" 2>&1 || rc=$?
  rd_ok MD-S-0 "$([[ $rc == 1 ]] && grep -q '"maintenance.pending_review","status":"pending"' "$RD_L/mds0.jsonl" && echo 1 || echo 0)" "pending_review pending before repair (exit $rc)"
}

rd_stage_mdr() {
  rd_go tgt repair1.log '^TestRecoveryRepair$' -e RECOVERY_IDS="$POST" -e RECOVERY_REPAIR_STEP=1 || true
  if rd_gopass TestRecoveryRepair repair1.log; then
    REPAIR="$(sed -n 's/^RECOVERY-REPAIR //p' "$RD_L/repair1.log" | head -1)"; rd_save REPAIR "$REPAIR"; rd_rec MD-R PASS "deleting row: deletion_pending under a provider fault, deleted on retry; inert pending row deleted; ready object deleted without a ledger entry (T)"
  else rd_rec MD-R FAIL "repair step 1"; grep -E -- '--- FAIL|_test.go:[0-9]+' "$RD_L/repair1.log" | head; fi
}

# mdo_fail_codes: the closed MD-O subcondition codes (G016 amendment 10), for the existing B-MAINTENANCE-DRILL FAIL row
# only. It reads nothing but RD_MDO_CODES, which rd_stage_mdo resets and fills with fixed literals at its own assertion
# branches (never the raw `why` text or tool output), and prints exactly one token list: the known codes in canonical
# order (then one `unknown` for any token outside the closed set), or `unknown` when none was recorded. Oversize input
# or any character outside the codes' own is `unknown`. It always exits 0 and never echoes input.
mdo_fail_codes() {
  local in="${RD_MDO_CODES:-}" tok c out="" unknown=0 IFS=' '
  local -a order=(plant first T-gone-early sync unavailable second repair2 left)
  local -A known=([plant]=1 [first]=1 [T-gone-early]=1 [sync]=1 [unavailable]=1 [second]=1 [repair2]=1 [left]=1) seen=()
  if (( ${#in} > 128 )) || [[ ! "$in" =~ ^[A-Za-z0-9\ -]*$ ]]; then printf 'unknown\n'; return 0; fi
  for tok in $in; do if [[ -n "${known[$tok]+x}" ]]; then seen[$tok]=1; else unknown=1; fi; done
  for c in "${order[@]}"; do if [[ -n "${seen[$c]+x}" ]]; then out+="${out:+,}$c"; fi; done
  if (( unknown )) || [[ -z "$out" ]]; then out+="${out:+,}unknown"; fi
  printf '%s\n' "$out"
  return 0
}

# off_fail_codes: the closed B-TOGGLE-OFF subcondition codes (G016 attempt 8), for the existing B-TOGGLE-OFF FAIL row
# only. It reads nothing but OFF_CODES, which variants_off resets and fills with fixed literals at its own assertion
# branches (never log text or tool output), and prints exactly one token list: the known codes in canonical order (then
# one `unknown` for any token outside the closed set), or `unknown` when none was recorded. Oversize input or any
# character outside the codes' own is `unknown`. It always exits 0 and never echoes input.
off_fail_codes() {
  local in="${OFF_CODES:-}" tok c out="" unknown=0 IFS=' '
  local -a order=(start migrate-apply migrate-inspect service image control config files pycontainer core-settings ready check stop)
  local -A known=([start]=1 [migrate-apply]=1 [migrate-inspect]=1 [service]=1 [image]=1 [control]=1 [config]=1 [files]=1
    [pycontainer]=1 [core-settings]=1 [ready]=1 [check]=1 [stop]=1) seen=()
  if (( ${#in} > 128 )) || [[ ! "$in" =~ ^[A-Za-z0-9\ -]*$ ]]; then printf 'unknown\n'; return 0; fi
  for tok in $in; do if [[ -n "${known[$tok]+x}" ]]; then seen[$tok]=1; else unknown=1; fi; done
  for c in "${order[@]}"; do if [[ -n "${seen[$c]+x}" ]]; then out+="${out:+,}$c"; fi; done
  if (( unknown )) || [[ -z "$out" ]]; then out+="${out:+,}unknown"; fi
  printf '%s\n' "$out"
  return 0
}

rd_stage_mdo() {
  local orgA tomb pend t all=1 why="" a b
  RD_MDO_CODES=""
  orgA="$(rd_j "$POST" OrgA)"; tomb="$(rd_j "$POST" TombA1)"; pend="$(rd_j "$POST" Pending)"; t="$(rd_j "$REPAIR" tombstone)"
  local cut; cut="$(rd_j "$REPAIR" cut)"
  rd_policy "$RD_P/mdo.policy" policy_id=mdo policy_label=fixture "scope=organization:$orgA" "object_cutoff=$cut" "holds=$pend" batch_size=500 max_batches=10 statement_timeout_ms=5000 lock_timeout_ms=1000 run_timeout_s=120
  purge() { rd_tool retention objects --policy "$RD_P/mdo.policy" --env-file "$RD_TGTENV" --preview "$RD_P/mdo.preview" --ledger "$RD_LEDGER" >/dev/null 2>&1; rd_tool retention objects --policy "$RD_P/mdo.policy" --env-file "$RD_TGTENV" --preview "$RD_P/mdo.preview" --ledger "$RD_LEDGER" --apply --receipts "$RD_RCPT" 2>&1; }
  printf 'planted' | rd_mc tgt svc "cat >/tmp/f && mc cp -q /tmp/f 'x/$(rd_ev tgt OBJECT_STORE_BUCKET)/objects/$orgA/$tomb'" >/dev/null 2>&1 || { all=0; why+="plant "; RD_MDO_CODES+="plant "; }
  a="$(purge)"; [[ "$a" == *"not_preserved=1"* && "$a" == *"bytes_present=1"* ]] || { all=0; why+="first[$a] "; RD_MDO_CODES+="first "; }
  [[ "$(rd_sql tgt "SELECT count(*) FROM app.stored_objects WHERE id = '$t'")" == 1 ]] || { all=0; why+="T-gone-early "; RD_MDO_CODES+="T-gone-early "; }
  rd_tool ledger sync --ledger "$RD_LEDGER" --env-file "$RD_TGTENV" --backup-root "$RD_ROOT" >/dev/null 2>&1 || { all=0; why+="sync "; RD_MDO_CODES+="sync "; }
  docker stop "$(rd_mic $RD_TP)" >/dev/null
  local n0; n0="$(rd_sql tgt "SELECT count(*) FROM app.stored_objects")"
  a="$(purge)"; [[ "$a" == *"provider_unavailable="[1-9]* && "$(rd_sql tgt "SELECT count(*) FROM app.stored_objects")" == "$n0" ]] || { all=0; why+="unavailable[$a] "; RD_MDO_CODES+="unavailable "; }
  rd_dc tgt up -d --wait minio >/dev/null 2>&1
  a="$(purge)"; [[ "$(rd_sql tgt "SELECT count(*) FROM app.stored_objects WHERE id = '$t'")" == 0 && "$a" == *"bytes_present=1"* ]] || { all=0; why+="second[$a] "; RD_MDO_CODES+="second "; }
  rd_go tgt repair2.log '^TestRecoveryRepair$' -e RECOVERY_IDS="$POST" -e RECOVERY_REPAIR_STEP=2 || true
  rd_gopass TestRecoveryRepair repair2.log || { all=0; why+="repair2 "; RD_MDO_CODES+="repair2 "; }
  a="$(purge)"
  local left; left="$(rd_sql tgt "SELECT string_agg(v, ',' ORDER BY v COLLATE \"C\") FROM (SELECT CASE id WHEN '$pend' THEN 'held' WHEN '$(rd_j "$POST" ObjA1)' THEN 'recent' WHEN '$(rd_j "$POST" TombB1)' THEN 'scopeB' WHEN '$(rd_j "$POST" ObjB1)' THEN 'readyB1' ELSE id::text END || ':' || state AS v FROM app.stored_objects) q")"
  [[ "$left" == "held:deleted,readyB1:ready,recent:deleted,scopeB:deleted" ]] || { all=0; why+="left[$left] "; RD_MDO_CODES+="left "; }
  rd_ok MD-O "$all" "not_preserved before sync, provider_unavailable with MinIO stopped (nothing removed), bytes_present until the authorized repeat delete, then purged; held, recent, out-of-scope and ready rows remain ${why:+(failed: $why)}"
}

rd_stage_mde() {
  local all=1 why="" s1="$RD_ROOT/$B1" o rc
  rd_policy "$RD_P/exp.policy" policy_id=exp backup_max_age_s=3600 backup_retain_for_s=1 "backup_holds=$B1,$B3" pending_review_after_s=1
  sleep 2; o="$(rd_tool expire --backup-root "$RD_ROOT" --backup-policy "$RD_P/exp.policy" --apply 2>&1)"
  [[ -f "$s1/database.dump" && -f "$RD_ROOT/$B2/manifest.tsv" && ! -e "$RD_ROOT/$B2/EXPIRED" ]] || { all=0; why+="kept "; }
  rd_policy "$RD_P/exp.policy" policy_id=exp backup_max_age_s=3600 backup_retain_for_s=1 "backup_holds=$B3" pending_review_after_s=1
  rm -f "$s1/objects.tar"; mkdir "$s1/objects.tar"; touch "$s1/objects.tar/obstruction"
  rc=0; o="$(rd_tool expire --backup-root "$RD_ROOT" --backup-policy "$RD_P/exp.policy" --apply 2>&1)" || rc=$?
  [[ -e "$s1/EXPIRED-PARTIAL" && "$o" == *expired-partial* ]] || { all=0; why+="partial[$o] "; }
  rc=0; rd_tool restore --backup "$s1" --target-env-file "$RD_TGTENV" --receipts "$RD_P/r2" --backup-policy "$RD_P/exp.policy" >/dev/null 2>&1 || rc=$?; [[ $rc == 3 ]] || { all=0; why+="restore-partial($rc) "; }
  rm -rf "$s1/objects.tar"
  rc=0; o="$(rd_tool expire --backup-root "$RD_ROOT" --backup-policy "$RD_P/exp.policy" --apply 2>&1)" || rc=$?
  [[ $rc == 0 && ! -e "$s1/EXPIRED-PARTIAL" && -e "$s1/EXPIRED" && ! -e "$s1/manifest.tsv" && ! -e "$s1/COMPLETE" && "$o" == *"replicated or immutable copies are unknown"* ]] || { all=0; why+="rerun[$o] "; }
  rc=0; rd_tool restore --backup "$s1" --target-env-file "$RD_TGTENV" --receipts "$RD_P/r2" --backup-policy "$RD_P/exp.policy" >/dev/null 2>&1 || rc=$?; [[ $rc == 3 ]] || { all=0; why+="restore-expired($rc) "; }
  rd_ok MD-E "$all" "held B1 and quarantined B2 kept; B1 expired-partial with an obstruction, restore refused, rerun completes (manifest and COMPLETE removed) with the other-copies statement; expired restore refused ${why:+(failed: $why)}"
}

rd_stage_mds() {
  local all=1 why="" rc
  rc=0; rd_tool status --backup-root "$RD_ROOT" --backup-policy "$RD_BPOL" --receipts "$RD_RCPT" --env-file "$RD_TGTENV" --ledger "$RD_LEDGER" >"$RD_L/mds1.jsonl" 2>&1 || rc=$?
  grep -q '"maintenance.pending_review","status":"ok"' "$RD_L/mds1.jsonl" || { all=0; why+="pending-ok "; }
  rc=0; rd_tool status --backup-root "$RD_ROOT" --backup-policy "$RD_P/missing.policy" --receipts "$RD_RCPT" >"$RD_L/mds2.jsonl" 2>&1 || rc=$?; [[ $rc == 1 ]] && grep -q '"policy","status":"unknown"' "$RD_L/mds2.jsonl" || { all=0; why+="unknown "; }
  rd_policy "$RD_P/stale.policy" policy_id=stale backup_max_age_s=1 backup_retain_for_s=3600 backup_holds= pending_review_after_s=1
  sleep 2; rc=0; rd_tool status --backup-root "$RD_ROOT" --backup-policy "$RD_P/stale.policy" --receipts "$RD_RCPT" >"$RD_L/mds3.jsonl" 2>&1 || rc=$?; [[ $rc == 1 ]] && grep -q '"backup.latest","status":"stale"' "$RD_L/mds3.jsonl" || { all=0; why+="stale "; }
  mkdir -p "$RD_P/empty-root"; rc=0; rd_tool status --backup-root "$RD_P/empty-root" --backup-policy "$RD_BPOL" --receipts "$RD_RCPT" >"$RD_L/mds4.jsonl" 2>&1 || rc=$?; [[ $rc == 1 ]] && grep -q '"backup.latest","status":"missing"' "$RD_L/mds4.jsonl" || { all=0; why+="missing "; }
  grep -q '"maintenance.last","status":"ok"' "$RD_L/mds1.jsonl" || grep -q '"maintenance.last","status":"partial"' "$RD_L/mds1.jsonl" || { all=0; why+="maint-last "; }
  rd_ok MD-S "$all" "pending_review ok after repair and purge; missing policy unknown, stale, missing, receipts reported ${why:+(failed: $why)}"
}

# rd_measure: the sanitized RD-M measurements (Plan r3 §7.4): sizes, window, durations and the recovery point; never a
# path, value or row content.
rd_measure() {
  local m="$RD_P/b1-manifest.tsv" out="$WORK/recovery-drill.tsv" k
  [[ -f "$m" ]] || cp "$RD_ROOT/$B1/manifest.tsv" "$m" 2>/dev/null || return 0
  {
    printf 'measure\tvalue\n'
    printf 'environment\t%s\n' "$([[ -n "${GITLAB_CI:-}" ]] && echo hosted-dind || echo local) docker $(docker version --format '{{.Server.Version}}' 2>/dev/null)"
    printf 'pg_image\t%s\n' "$(sed -n 's/^pg_image\t//p' "$m")"
    printf 'mc_image\t%s\n' "$(sed -n 's/^mc_image\t//p' "$m")"
    printf 'core_check_image\t%s\n' "$(docker image inspect -f '{{.Id}}' "$RD_IMG" 2>/dev/null)"
    printf 'pg_dump\t%s\n' "$(sed -n 's/^pg_dump_version\t//p' "$m")"
    awk -F'\t' '$1 == "table" { printf "rows %s\t%s\n", $2, $3 }' "$m"
    printf 'objects\t%s\n' "$(grep -c $'^object\t' "$m" || true)"
    printf 'database_dump_bytes\t%s\n' "$(awk -F'\t' '$1 == "component" && $2 == "database.dump" { print $3 }' "$m")"
    printf 'objects_tar_bytes\t%s\n' "$(awk -F'\t' '$1 == "component" && $2 == "objects.tar" { print $3 }' "$m")"
    printf 'window_db\t%s\n' "$(sed -n 's/^window_db\t//p' "$m" | tr '\t' ' ')"
    printf 'window_host\t%s\n' "$(sed -n 's/^window_host\t//p' "$m" | tr '\t' ' ')"
    printf 'recovery_point\t%s (the B1 window end, database clock: changes after it are lost except replayed restrictions)\n' "$(sed -n 's/^window_db\t[^\t]*\t//p' "$m")"
    printf 'restore_database_seconds\t%s\n' "$(sed -n 's/^database_seconds\t//p' "$RD_RCPT/restore.tsv")"
    printf 'restore_objects_seconds\t%s\n' "$(sed -n 's/^objects_seconds\t//p' "$RD_RCPT/restore.tsv")"
    for k in rdb rdr rdf rda rdy rdg rdo mda mdo; do printf 'stage_%s_seconds\t%s\n' "$k" "${RD_T[$k]:-}"; done
  } >"$out"
  if [[ -n "$REPORT_DIR" ]]; then cp "$out" "$REPORT_DIR/" 2>/dev/null || true; cp "$RD_RES" "$REPORT_DIR/recovery-steps.tsv" 2>/dev/null || true; fi
}

# rd_purge: remove the drill's detached clients with their anonymous volumes, its private material and its named
# volume, and stop the target (normal end and any failure).
rd_purge() {
  rd_client_rm adminw rtw put lock
  [[ -n "$RD_P" ]] && rm -rf -- "$RD_P"
  [[ -n "$RD_OLDV" ]] && docker volume rm -f "$RD_OLDV" >/dev/null 2>&1
  [[ -n "$RD_TP" && -f "$RD_TGTENV" ]] && (cd "$RD_M" && docker compose --env-file .env.rec -f infra/compose/local.yml down -v >/dev/null 2>&1)
  rm -f -- "$RD_TGTENV"
  return 0
}

declare -A RD_T=()
RD_MDO_CODES=""
recovery_drill() { # recovery_drill DEST PROJECT IMAGE: B-RECOVERY-DRILL, then B-MAINTENANCE-DRILL on the reopened target
  local M="$1" project="$2" img="$3" s t0 keys i
  RD_MDO_CODES=""
  RD_M="$M" RD_SP="$project" RD_TP="g013-rec-$RUN_ID" RD_DB="$(env_get "$M" DATABASE_NAME)" RD_IMG="$img"
  RD_PGIMG="pgvector/pgvector:0.8.7-pg17" RD_SRCENV="$M/.env" RD_TGTENV="$M/.env.rec" RD_OLDV="g013-old-$RUN_ID"
  mkdir -p "$WORK/recovery"; chmod 700 "$WORK/recovery"
  RD_P="$WORK/recovery/private" RD_L="$WORK/recovery/logs" RD_RES="$WORK/recovery-steps.tsv"
  mkdir -p "$RD_P" "$RD_L"; chmod 700 "$RD_P"; : >"$RD_RES"
  RD_ROOT="$RD_P/backups" RD_RCPT="$RD_P/receipts" RD_LEDGER="$RD_P/ledger/ledger.tsv" RD_BPOL="$RD_P/backup.policy" RD_SINK="$RD_P/sink.jsonl"
  RD_CID="$RD_P/cid"; [[ -n "$OWNED_REG" ]] && RD_CID="$(dirname "$OWNED_REG")/cid"
  if ports_busy "${PORTS_REC[@]}"; then
    record B-RECOVERY-DRILL FAIL "ports ${PORTS_REC[*]} busy; refusing to re-map silently"
    record B-MAINTENANCE-DRILL UNMET "the recovery drill did not reach its reopen step"; rd_purge; return
  fi
  # Source: the migrations project, reset to fresh volumes and migrated to 12; target: its own env file on this copy.
  local setup=1
  compose "$M" down -v >"$RD_L/reset.log" 2>&1 || setup=0
  obj_minio "$M" "$project" || setup=0
  compose_migrate "$M" run --rm migrate up >"$RD_L/migrate.log" 2>&1 && grep -qP '^current_version\t22$' "$RD_L/migrate.log" || setup=0
  make_env_file "$RD_TGTENV" "$RD_TP"
  keys=(WEB_PORT CORE_PORT INTELLIGENCE_PORT POSTGRES_PORT REDIS_PORT MINIO_API_PORT MINIO_CONSOLE_PORT)
  for i in "${!keys[@]}"; do sed -i "s/^${keys[$i]}=.*/${keys[$i]}=${PORTS_REC[$i]}/" "$RD_TGTENV"; done
  register_project "$RD_TP"
  docker volume create --label "g001.run=$RUN_ID" "$RD_OLDV" >/dev/null && registry_row volume "$RD_OLDV" || setup=0
  rd_dc tgt build minio minio-init >"$RD_L/build-tgt.log" 2>&1 || setup=0
  record_built_image "$RD_TP-minio:latest"; record_built_image "$RD_TP-minio-init:latest"
  mkdir -p "$RD_ROOT" "$RD_RCPT" "$RD_P/ledger"; chmod 700 "$RD_ROOT" "$RD_RCPT" "$RD_P/ledger"
  [[ -n "$OWNED_REG" ]] || { mkdir -p "$RD_CID"; chmod 700 "$RD_CID"; }
  rd_policy "$RD_BPOL" policy_id=fixture-backup backup_max_age_s=3600 backup_retain_for_s=1 backup_holds= pending_review_after_s=1
  if [[ $setup == 0 ]]; then
    tail -20 "$RD_L/migrate.log"; record B-RECOVERY-DRILL FAIL "drill setup (source reset, MinIO, migration, target build)"
    record B-MAINTENANCE-DRILL UNMET "the recovery drill did not reach its reopen step"; rd_purge; return
  fi
  # Overall bounds (Plan r3 §11.4): 30 minutes for the recovery drill, 15 for the maintenance drill; a stage that would
  # start past the bound is not started (each tool step has its own timeout).
  local b0=$SECONDS
  for s in rds rdb rdk rdp rdi rdr rdf rdv rda rdl rdy rdg rdo; do
    (( SECONDS - b0 < 1800 )) || { rd_rec STAGE FAIL "recovery drill past its 30-minute bound before $s"; break; }
    t0=$SECONDS
    "rd_stage_$s" || { rd_rec STAGE FAIL "$s aborted"; break; }
    RD_T[$s]=$((SECONDS - t0))
  done
  local fails steps; fails="$(awk -F'\t' '$2 == "FAIL" { printf "%s ", $1 }' "$RD_RES")"; steps="$(awk -F'\t' '$2 == "PASS" { printf "%s ", $1 }' "$RD_RES")"
  [[ -f "$RD_ROOT/$B1/manifest.tsv" ]] && rd_measure
  if [[ -z "$fails" ]] && grep -q $'^RD-O\tPASS' "$RD_RES"; then
    record B-RECOVERY-DRILL PASS "source = the migrations project at version 22, target = $RD_TP on the same copy (own env file, PORTS_REC); steps: $steps(RD-S tool refusals incl. invalid ledger references; RD-B seed, administrator write refused by the X3 lock safeguard with an exact release, in-flight runtime and provider writes drained, late writers refused while fenced, B1 under the established fence with release deferred then unfence, the recorded administrator window required, failing backups never COMPLETE (incl. a failure whose INCOMPLETE marker cannot be written: fence still released, verified), a TERM during a capture under an operator-established fence keeps that fence held (record, database ACL, provider, status) until unfence restores the prior state while the integrated-capture TERM releases; RD-K fence/unfence SIGKILL recovery and preserved prior state; RD-P six restriction kinds, ledger appends and sync, a source purge, B2; RD-I corrupted B2 detected by status and the sink, quarantined, restore refused; RD-R nine refusals before mutation incl. a reused receipts directory; RD-F staged restore with a receipt-bound continuation that refuses a foreign object; RD-V; RD-A invalidation; RD-L seven gate refusals incl. an invalidation receipt of another restore; RD-Y replay with a stopped provider and an injected audit fault, gap receipted, wrong references never already applied; RD-G TestRecoveryVerify V1-V9; RD-O a restriction appended and attested after replay refused until replayed, ledger drift during replay refused, then reopen sign-off, B3 and status ok); measurements in recovery-drill.tsv"
  else
    record B-RECOVERY-DRILL FAIL "failed steps: ${fails:-none} (reached: $steps)"
    record B-MAINTENANCE-DRILL UNMET "the recovery drill did not reach its reopen step"
    [[ -n "$REPORT_DIR" ]] && cp "$RD_RES" "$REPORT_DIR/recovery-steps.tsv" 2>/dev/null
    rd_purge; [[ -e "$RD_P" ]] && log "drill private material could not be removed"
    return
  fi
  : >"$RD_RES.md"; local mres="$RD_RES"; RD_RES="$RD_RES.md"
  b0=$SECONDS
  for s in mda mds0 mdr mdo mde mds; do
    (( SECONDS - b0 < 900 )) || { rd_rec STAGE FAIL "maintenance drill past its 15-minute bound before $s"; break; }
    t0=$SECONDS
    "rd_stage_$s" || { rd_rec STAGE FAIL "$s aborted"; break; }
    RD_T[$s]=$((SECONDS - t0))
  done
  rd_measure
  # MD-C: the private directory and the named volume are removed; nothing private remains under the work directory.
  rd_purge
  local left=0; [[ -e "$RD_P" ]] && left=1; docker volume inspect "$RD_OLDV" >/dev/null 2>&1 && left=1
  find "$WORK" \( -name 'ledger.tsv*' -o -name '*.dump' -o -name 'objects.tar' -o -name 'manifest.tsv' -o -name 'restore.tsv' -o -name '.env.rec' \) 2>/dev/null | grep -q . && left=1
  rd_ok MD-C "$([[ $left == 0 ]] && echo 1 || echo 0)" "private directory, named volume and target removed; no backup, ledger, receipt or env file left"
  cat "$RD_RES" >>"$mres"; [[ -n "$REPORT_DIR" ]] && cp "$mres" "$REPORT_DIR/recovery-steps.tsv" 2>/dev/null
  fails="$(awk -F'\t' '$2 == "FAIL" { printf "%s ", $1 }' "$RD_RES")"; steps="$(awk -F'\t' '$2 == "PASS" { printf "%s ", $1 }' "$RD_RES")"
  RD_RES="$mres"
  if [[ -z "$fails" ]] && grep -q $'^MD-C\tPASS' "$RD_RES"; then
    record B-MAINTENANCE-DRILL PASS "on the reopened target: $steps(MD-A dry-run unchanged, bounded partial, changed policy refused, lock timeout truthful then complete, holds/recent/B/global preserved, global scope only with authorization; MD-S pending before repair; MD-R repair through the guarded delete; MD-O not_preserved, provider_unavailable, bytes_present until the authorized repeat delete, then purge with held, recent, out-of-scope and ready rows kept; MD-E expiry with holds, quarantine kept, EXPIRED-PARTIAL then completion leaving only the EXPIRED record, expired restore refused; MD-S status ok/unknown/stale/missing; MD-C private material removed)"
  else
    local mdo=""; if [[ " $fails" == *" MD-O "* ]]; then mdo="$(mdo_fail_codes)"; fi
    record B-MAINTENANCE-DRILL FAIL "failed steps: ${fails:-none} (reached: $steps)${mdo:+; MD-O subconditions: $mdo}"
  fi
}
# --- end of the G013 drill block ---

# ===================================================================================
# G003 observability on running disposable copies (Plan §4, §5, §6.3, §16 AC1-AC4).
# Every probe runs with httpprobe inside the target service's network namespace; canaries and
# secret-valued identifiers travel only in files under $OBS_IN ($WORK/obs-inputs, excluded from
# B-SECRET-CANARY like secrets*, deleted after use); logs are read with docker logs.
OBS="" OBS_IN="" OBS_OUT="" OBS_SVCS=() OBS_CASES=()
obs_hp() { # obs_hp CID NAME HTTPPROBE-ARGS...: one probe -> $OBS/NAME.json
  local cid="$1" name="$2"; shift 2
  docker_run_owned obs-httpprobe --rm --label g001.run=$RUN_ID --network "container:$cid" -v "$OBS_IN:/in:ro" "$VT_IMG" httpprobe "$@" >"$OBS/$name.json" 2>>"$OBS/httpprobe.err" || true
}
obs_get() { yq_stdin -p json -o tsv "$2" <"$OBS/$1.json" 2>/dev/null || true; } # obs_get NAME EXPR
obs_file() { (umask 022; printf '%s\n' "$2" >"$OBS_IN/$1"); } # obs_file NAME CONTENT (never in argv of a probe)
obs_rid() { printf 'obs-%s-%s' "$1" "$(random_hex 6)"; }
obs_tp() { printf '00-%s-%s-%s' "$(random_hex 16)" "$(random_hex 8)" "${1:-01}"; }
# obs_case ID REQUEST_ID COUNTS [trace_id parent flags]: one chain case (COUNTS: web:core[:intelligence], 0|1|0-1).
obs_case() {
  local id="$1" rid="$2" counts="$3" tid="${4:-}" parent="${5:-}" flags="${6:-}" svc i=0 c j="" cnt
  IFS=: read -r -a cnt <<<"$counts"
  for svc in "${OBS_SVCS[@]}"; do c="${cnt[$i]}"; [[ "$c" == 0-1 ]] && c='"0-1"'; j+="${j:+,}\"$svc\":$c"; i=$((i + 1)); done
  j="{\"id\":\"$id\",\"request_id\":\"$rid\",\"expect\":{$j}"
  [[ -n "$tid" ]] && j+=",\"trace_id\":\"$tid\""
  [[ -n "$parent" ]] && j+=",\"parent_span_id\":\"$parent\""
  [[ -n "$flags" ]] && j+=",\"trace_flags\":\"$flags\""
  OBS_CASES+=("$j}")
}
obs_logs() { # obs_logs DEST: raw logs of the scoped services
  local svc c
  for svc in "${OBS_SVCS[@]}"; do c="$(cid_of "$1" "$svc")"; docker logs "$c" >"$OBS/$svc.log" 2>&1 || true; done
}
obs_record() { # obs_record SVC REQUEST_ID -> the completion records (JSON lines)
  obs_open; vtools -v "$OBS:/o:ro" -v "$OBS_IN:/i:ro" "$VT_IMG" logcheck records --service "$1" --file "/o/$1.log" --request-id "$2" 2>/dev/null || true
}
obs_ms() { echo $(( $(date +%s%N) / 1000000 )); }
obs_open() { chmod -R a+rX "$OBS" "$OBS_IN"; } # the tools image runs as nonroot; $WORK itself stays 0700
obs_chain_file() { printf '{"version":1,"cases":[%s]}\n' "$(IFS=,; echo "${OBS_CASES[*]}")" >"$OBS/chain.json"; }
# obs_recover CORE_CID: core readiness back to 200 within 90 s; prints the recovery time in ms or "timeout".
obs_recover() {
  local t0 i; t0="$(obs_ms)"
  for ((i = 0; i < 90; i++)); do
    obs_hp "$1" recover --url http://127.0.0.1:8180/readyz --timeout 5s
    [[ "$(obs_get recover .status)" == 200 ]] && { echo $(( $(obs_ms) - t0 )); return 0; }
    sleep 1
  done
  echo timeout; return 1
}

# obs_outage DEST LABEL ACTION SVC ...: one row of the §6.3 matrix; writes timings and chain cases.
# ACTION: stop | pause. Checks the statuses for the injected outage; every request <= 5000 ms.
obs_outage() {
  local dest="$1" label="$2" action="$3" svc="$4" intel="$5" ok=1 web core ic rid tp tid par st dur rcv body want_core want_web
  web="$(cid_of "$dest" web)"; core="$(cid_of "$dest" core)"; ic=""; [[ "$intel" == yes ]] && ic="$(cid_of "$dest" intelligence)"
  local target; target="$(cid_of "$dest" "$svc")"
  if [[ "$action" == stop ]]; then compose "$dest" stop "$svc" >/dev/null 2>&1; else docker pause "$target" >/dev/null 2>&1; fi
  sleep 1
  rid="$(obs_rid "$label")"; tp="$(obs_tp 01)"; tid="$(cut -d- -f2 <<<"$tp")"; par="$(cut -d- -f3 <<<"$tp")"
  obs_hp "$web" "$label-web" --url http://127.0.0.1:3000/api/ready --header "X-Request-Id: $rid" --header "traceparent: $tp" --timeout 10s
  st="$(obs_get "$label-web" .status)"; dur="$(obs_get "$label-web" .duration_ms)"
  printf '%s\tweb /api/ready\t%s\t%s\n' "$label" "$st" "$dur" >>"$WORK/obs-timings.tsv"
  # O7 and O8: 200 or 503 here; obs_web_probe then requires O7's pair to be 200/ok or 503/timeout (F11).
  case "$label" in *O7|*O8) want_web="200|503" ;; *) want_web=503 ;; esac
  [[ "$st" =~ ^($want_web)$ ]] || { ok=0; log "$label web /api/ready status $st (want $want_web)"; }
  [[ "$dur" =~ ^[0-9]+$ ]] && (( dur <= 5000 )) || { ok=0; log "$label web /api/ready took ${dur}ms"; }
  [[ "$(obs_get "$label-web" .cache_control)" == *no-store* ]] || { ok=0; log "$label web /api/ready lacks no-store"; }
  if [[ "$svc" != core ]]; then
    obs_hp "$core" "$label-core" --url http://127.0.0.1:8180/readyz --header "X-Request-Id: $rid-c" --timeout 10s
    st="$(obs_get "$label-core" .status)"; dur="$(obs_get "$label-core" .duration_ms)"; body="$(obs_get "$label-core" .body)"
    printf '%s\tcore /readyz\t%s\t%s\n' "$label" "$st" "$dur" >>"$WORK/obs-timings.tsv"
    [[ "$dur" =~ ^[0-9]+$ ]] && (( dur <= 5000 )) || { ok=0; log "$label core /readyz took ${dur}ms"; }
    if [[ "$svc" == intelligence ]]; then want_core=200; else want_core=503; fi
    [[ "$st" == "$want_core" ]] || { ok=0; log "$label core /readyz status $st (want $want_core)"; }
    case "$svc" in
      postgres|redis) grep -q "\"$svc\":\"unavailable\"" <<<"$body" || { ok=0; log "$label core body does not name $svc"; } ;;
      minio) grep -q '"object_store":"unavailable"' <<<"$body" || { ok=0; log "$label core body does not name object_store"; } ;;
      intelligence) grep -q '"optional":{"intelligence":"unavailable"}' <<<"$body" || { ok=0; log "$label optional intelligence not unavailable"; } ;;
    esac
    obs_hp "$core" "$label-core-live" --url http://127.0.0.1:8180/healthz --timeout 10s
    [[ "$(obs_get "$label-core-live" .status)" == 200 ]] || { ok=0; log "$label core liveness"; }
  fi
  obs_hp "$web" "$label-web-live" --url http://127.0.0.1:3000/api/health --timeout 10s
  [[ "$(obs_get "$label-web-live" .status)" == 200 && "$(obs_get "$label-web-live" .cache_control)" == *no-store* ]] || { ok=0; log "$label web liveness"; }
  printf '%s\t%s\t%s\n' "$label" "$rid" "$(obs_get "$label-web" .status)" >>"$OBS/outage-requests.tsv"
  # Recovery (separate from single-request timing; finite deadline 90 s).
  if [[ "$action" == stop ]]; then compose "$dest" start "$svc" >/dev/null 2>&1; else docker unpause "$target" >/dev/null 2>&1; fi
  if [[ "$svc" == core ]]; then
    rcv=timeout; local t0 i; t0="$(obs_ms)"
    for ((i = 0; i < 90; i++)); do obs_hp "$web" recover-web --url http://127.0.0.1:3000/api/ready --timeout 5s; [[ "$(obs_get recover-web .status)" == 200 ]] && { rcv=$(( $(obs_ms) - t0 )); break; }; sleep 1; done
  else rcv="$(obs_recover "$core")" || true; fi
  [[ "$svc" == intelligence ]] && compose "$dest" up -d --wait intelligence >/dev/null 2>&1
  printf '%s\trecovery\t-\t%s\n' "$label" "$rcv" >>"$WORK/obs-timings.tsv"
  [[ "$rcv" =~ ^[0-9]+$ ]] || { ok=0; log "$label did not recover within 90 s"; }
  # Expected chain for the web request (and the direct core request when core was up).
  case "$svc" in
    postgres|redis|minio)
      if [[ "$action" == stop ]]; then obs_case "$label" "$rid" "1:1${ic:+:1}" "$tid" "$par" 01
      else obs_case "$label" "$rid" "1:1${ic:+:0-1}" "$tid" "$par" 01; fi ;;
    intelligence)
      if [[ "$action" == stop ]]; then obs_case "$label" "$rid" "1:1:0" "$tid" "$par" 01
      else obs_case "$label" "$rid" "1:1:0-1" "$tid" "$par" 01; fi ;;
    core)
      if [[ "$action" == stop ]]; then obs_case "$label" "$rid" "1:0${ic:+:0}" "$tid" "$par" 01
      else obs_case "$label" "$rid" "1:0-1${ic:+:0-1}" "$tid" "$par" 01; fi ;;
  esac
  return $((1 - ok))
}

obs_field() { sed -n "s/.*\"$1\":\"\([a-z_]*\)\".*/\1/p" | head -1; } # obs_field KEY <RECORD: a string field
# obs_web_probe LABEL WANT... (after obs_logs): the web record's core_probe for the outage request is one
# of WANT; a WANT of the form STATUS/PROBE must match the web status and core_probe together. The case's
# statuses and categories (never content) go to obs-outages.tsv, which records the observed outcome.
obs_web_probe() {
  local label="$1" rid st wrec crec p wo cn co cs w; shift
  IFS=$'\t' read -r _ rid st < <(awk -F'\t' -v l="$label" '$1==l' "$OBS/outage-requests.tsv")
  wrec="$(obs_record web "$rid")"; crec="$(obs_record core "$rid")"
  p="$(obs_field core_probe <<<"$wrec")"; wo="$(obs_field outcome <<<"$wrec")"
  cn="$(grep -c . <<<"$crec" || true)"; co="$(obs_field outcome <<<"$crec")"; cs="$(sed -n 's/.*"status":\([0-9]*\).*/\1/p' <<<"$crec" | head -1)"
  printf '%s\t%s\t%s\t%s\t%s\t%s\t%s\n' "$label" "$st" "${p:-none}" "${wo:-none}" "$cn" "${co:--}" "${cs:--}" >>"$OBS_OUT"
  for w in "$@"; do
    if [[ "$w" == */* ]]; then [[ "$st/$p" == "$w" ]] && return 0; else [[ "$p" == "$w" ]] && return 0; fi
  done
  log "$label web status $st core_probe '${p:-none}' (want $*)"; return 1
}
# obs_observed LABEL...: the observed "LABEL=web status/core_probe" of each case, from obs-outages.tsv.
obs_observed() {
  local l r=""
  for l in "$@"; do r+="$(awk -F'\t' -v l="$l" '$1 == l {printf "%s=%s/%s ", $1, $2, $3}' "$OBS_OUT")"; done
  printf '%s' "${r% }"
}
# obs_o10 LABEL (after obs_logs): after unpause, core logs at most one record for the request ID, cancelled.
obs_o10() {
  local rid rec; rid="$(awk -F'\t' -v l="$1" '$1==l{print $2}' "$OBS/outage-requests.tsv")"
  rec="$(obs_record core "$rid")"
  [[ -z "$rec" ]] && return 0
  [[ "$(grep -c . <<<"$rec")" == 1 ]] && grep -q '"outcome":"cancelled"' <<<"$rec" && return 0
  log "$1: core record for the paused request is not a single cancelled record"; return 1
}

obs_ops() { # obs_ops DEST: B-OBS-CONTEXT, B-OBS-CHAIN, B-OBS-TIMEOUTS, B-OBS-LOGS, B-OBS-PRIVACY (enabled copy)
  local dest="$1" web core ic ok rid tp tid par i v name st xr rec c
  OBS="$WORK/obs-ops"; OBS_IN="$(obs_inputs)/ops"; OBS_SVCS=(web core intelligence); OBS_CASES=()
  mkdir -p "$OBS" "$OBS_IN"; chmod 755 "$OBS" "$OBS_IN"; : >"$OBS/outage-requests.tsv"
  ensure_tools || { for v in B-OBS-CONTEXT B-OBS-CHAIN B-OBS-TIMEOUTS B-OBS-LOGS B-OBS-PRIVACY; do record "$v" FAIL "validation tools unavailable"; done; return; }
  web="$(cid_of "$dest" web)"; core="$(cid_of "$dest" core)"; ic="$(cid_of "$dest" intelligence)"
  printf 'case\ttarget\tstatus\tduration_ms\n' >"$WORK/obs-timings.tsv"
  OBS_OUT="$WORK/obs-outages.tsv"; printf 'case\tweb_status\tcore_probe\tweb_outcome\tcore_records\tcore_outcome\tcore_status\n' >"$OBS_OUT"

  # --- B-OBS-CONTEXT: invalid request IDs and trace contexts through web /api/ready -------------------
  ok=1; : >"$OBS_IN/invalid-literals.txt"; printf 'row\tstatus\tresponse_id\n' >"$WORK/obs-context.tsv"
  obs_ctx_row() { # obs_ctx_row NAME LITERAL HTTPPROBE-HEADER-ARGS...
    local nm="$1" lit="$2"; shift 2
    [[ -n "$lit" ]] && printf '%s\n' "$lit" >>"$OBS_IN/invalid-literals.txt"
    obs_hp "$web" "ctx-$nm" --url http://127.0.0.1:3000/api/ready "$@"
    st="$(obs_get "ctx-$nm" .status)"; xr="$(obs_get "ctx-$nm" '.x_request_id // ""')"
    printf '%s\t%s\t%s\n' "$nm" "$st" "${xr:0:8}…" >>"$WORK/obs-context.tsv"
    [[ "$st" == 200 ]] || { ok=0; log "context row $nm: status $st"; }
  }
  obs_file rid-big "$(random_hex 4096)"
  obs_file rid-513 "$(random_hex 257 | cut -c1-513)"
  obs_file rid-65 "$(random_hex 33 | cut -c1-65)"
  for name in absent malformed oversized-65 oversized-513 oversized-8k duplicated comma-joined; do
    case "$name" in
      absent) obs_ctx_row absent "" ;;
      malformed) v="bad id with spaces $(random_hex 4)"; obs_file rid-malformed "$v"; obs_ctx_row malformed "$v" --header-file "X-Request-Id=/in/rid-malformed" ;;
      oversized-65) obs_ctx_row oversized-65 "$(cat "$OBS_IN/rid-65")" --header-file "X-Request-Id=/in/rid-65" ;;
      oversized-513) obs_ctx_row oversized-513 "$(cat "$OBS_IN/rid-513")" --header-file "X-Request-Id=/in/rid-513" ;;
      oversized-8k) obs_ctx_row oversized-8k "$(cat "$OBS_IN/rid-big")" --header-file "X-Request-Id=/in/rid-big" ;;
      duplicated) v="dup-$(random_hex 8)"; c="dup-$(random_hex 8)"; obs_file rid-d1 "$v"; obs_file rid-d2 "$c"
        printf '%s\n%s\n' "$v" "$c" >>"$OBS_IN/invalid-literals.txt"; obs_ctx_row duplicated "" --header-file "X-Request-Id=/in/rid-d1" --header-file "X-Request-Id=/in/rid-d2" ;;
      comma-joined) v="cj-$(random_hex 8), cj-$(random_hex 8)"; obs_file rid-cj "$v"; obs_ctx_row comma-joined "$v" --header-file "X-Request-Id=/in/rid-cj" ;;
    esac
    xr="$(obs_get "ctx-$name" '.x_request_id // ""')"
    [[ "$xr" =~ ^[0-9a-f]{32}$ ]] || { ok=0; log "context row $name: response ID is not a generated 32-hex value"; continue; }
    obs_case "ctx-$name" "$xr" 1:1:1 "" "" 00
  done
  # Invalid traceparent rows, each with a valid supplied request ID: a new local trace, no parent.
  for name in zero-trace zero-parent uppercase version-01 version-ff extra-field wrong-length duplicated comma-joined; do
    tid="$(random_hex 16)"; par="$(random_hex 8)"
    case "$name" in
      zero-trace) v="00-00000000000000000000000000000000-$par-01" ;;
      zero-parent) v="00-$tid-0000000000000000-01" ;;
      uppercase) v="00-$(tr 'a-f' 'A-F' <<<"$tid")-$par-01" ;;
      version-01) v="01-$tid-$par-01" ;;
      version-ff) v="ff-$tid-$par-01" ;;
      extra-field) v="00-$tid-$par-01-$(random_hex 4)" ;;
      wrong-length) v="00-${tid:1}-$par-01" ;;
      duplicated) v="00-$tid-$par-01" ;;
      comma-joined) v="00-$tid-$par-01,$(obs_tp 01)" ;;
    esac
    # Neither the raw value nor its trace or parent part may be adopted or logged.
    obs_file "tp-$name" "$v"; printf '%s\n%s\n%s\n' "$v" "$tid" "$par" >>"$OBS_IN/invalid-literals.txt"
    rid="$(obs_rid "tp-$name")"
    if [[ "$name" == duplicated ]]; then obs_file tp-dup2 "$(obs_tp 01)"; cat "$OBS_IN/tp-dup2" >>"$OBS_IN/invalid-literals.txt"
      obs_ctx_row "tp-$name" "" --header "X-Request-Id: $rid" --header-file "traceparent=/in/tp-$name" --header-file "traceparent=/in/tp-dup2"
    else obs_ctx_row "tp-$name" "" --header "X-Request-Id: $rid" --header-file "traceparent=/in/tp-$name"; fi
    [[ "$(obs_get "ctx-tp-$name" '.x_request_id // ""')" == "$rid" ]] || { ok=0; log "context row tp-$name: the valid supplied request ID was not kept"; }
    obs_case "tp-$name" "$rid" 1:1:1 "" "" 00
  done

  # --- B-OBS-CHAIN (i)-(iv) on the healthy stack ----------------------------------------------------
  rid="$(obs_rid chain-i)"; tp="$(obs_tp 01)"; tid="$(cut -d- -f2 <<<"$tp")"; par="$(cut -d- -f3 <<<"$tp")"
  obs_hp "$web" chain-i --url http://127.0.0.1:3000/api/ready --header "X-Request-Id: $rid" --header "traceparent: $tp"
  obs_case chain-i "$rid" 1:1:1 "$tid" "$par" 01
  obs_hp "$web" chain-ii --url http://127.0.0.1:3000/api/ready
  xr="$(obs_get chain-ii '.x_request_id // ""')"; obs_case chain-ii "${xr:-missing0}" 1:1:1 "" "" 00
  rid="$(obs_rid chain-iii)"; obs_file tp-chain-iii "ff-$(random_hex 16)-$(random_hex 8)-01"
  obs_hp "$web" chain-iii --url http://127.0.0.1:3000/api/ready --header "X-Request-Id: $rid" --header-file "traceparent=/in/tp-chain-iii"
  obs_case chain-iii "$rid" 1:1:1 "" "" 00
  obs_hp "$web" chain-iv --url http://127.0.0.1:3000/api/ready --parallel 32
  local iv_ok=1 iv_n=0
  for ((i = 0; i < 32; i++)); do
    rid="$(obs_get chain-iv ".[$i].sent.x_request_id // \"\"")"; tp="$(obs_get chain-iv ".[$i].sent.traceparent // \"\"")"
    [[ -n "$rid" && -n "$tp" ]] || { iv_ok=0; log "chain-iv request $i: no generated context reported"; continue; }
    obs_case "chain-iv-$i" "$rid" 1:1:1 "$(cut -d- -f2 <<<"$tp")" "$(cut -d- -f3 <<<"$tp")" 01
    iv_n=$((iv_n + 1))
  done
  [[ "$(obs_get chain-iv '[.[] | select(.status == 200)] | length')" == 32 ]] || { iv_ok=0; log "chain-iv: not all 32 concurrent requests answered 200"; }

  # --- B-OBS-PRIVACY inputs (healthy stack) ---------------------------------------------------------
  local pc=() k
  for k in auth cookie apikey baggage tracestate xcanary ua referer query body rid tp exec-prompt exec-ref exec-authz; do pc+=("Cx$(random_hex 12)"); done
  printf '%s\n' "${pc[@]}" >"$OBS_IN/privacy-canaries.txt"; printf '%s\n' "${pc[@]}" >>"$SECRETS_FILE"
  obs_file h-auth "Bearer ${pc[0]}"; obs_file h-cookie "session=${pc[1]}"; obs_file h-apikey "${pc[2]}"; obs_file h-baggage "k=${pc[3]}"
  obs_file h-tracestate "v=${pc[4]}"; obs_file h-xcanary "${pc[5]}"; obs_file h-ua "probe/${pc[6]}"; obs_file h-referer "http://referer.invalid/${pc[7]}"
  obs_file b-body "{\"q\":\"${pc[9]}\"}"; obs_file h-rid "bad ${pc[10]}"; obs_file h-tp "00-${pc[11]}-x-01"
  local priv=(--header-file "Authorization=/in/h-auth" --header-file "Cookie=/in/h-cookie" --header-file "X-Api-Key=/in/h-apikey"
    --header-file "baggage=/in/h-baggage" --header-file "tracestate=/in/h-tracestate" --header-file "X-Canary=/in/h-xcanary"
    --header-file "User-Agent=/in/h-ua" --header-file "Referer=/in/h-referer" --header-file "X-Request-Id=/in/h-rid"
    --header-file "traceparent=/in/h-tp" --body-file /in/b-body)
  obs_file u-web-health "http://127.0.0.1:3000/api/health?token=${pc[8]}&q=${pc[8]}"; obs_file u-web-ready "http://127.0.0.1:3000/api/ready?token=${pc[8]}&q=${pc[8]}"
  obs_file u-core-health "http://127.0.0.1:8180/healthz?token=${pc[8]}&q=${pc[8]}"; obs_file u-core-ready "http://127.0.0.1:8180/readyz?token=${pc[8]}&q=${pc[8]}"
  obs_file u-intel-health "http://127.0.0.1:8000/healthz?token=${pc[8]}&q=${pc[8]}"; obs_file u-intel-ready "http://127.0.0.1:8000/readyz?token=${pc[8]}&q=${pc[8]}"
  obs_hp "$web" priv-web-health --url-file /in/u-web-health "${priv[@]}"; obs_hp "$web" priv-web-ready --url-file /in/u-web-ready "${priv[@]}"
  obs_hp "$core" priv-core-health --url-file /in/u-core-health "${priv[@]}"; obs_hp "$core" priv-core-ready --url-file /in/u-core-ready "${priv[@]}"
  obs_hp "$ic" priv-intel-health --url-file /in/u-intel-health "${priv[@]}"; obs_hp "$ic" priv-intel-ready --url-file /in/u-intel-ready "${priv[@]}"
  # Secret-valued request IDs sent directly to core and to intelligence: replaced, never logged.
  obs_file s-core "$(env_get "$dest" REDIS_PASSWORD)"; obs_file s-intel "$(env_get "$dest" INTELLIGENCE_SERVICE_TOKEN)"
  cat "$OBS_IN/s-core" "$OBS_IN/s-intel" >"$OBS_IN/secret-ids.txt"
  obs_hp "$core" priv-core-secret --url http://127.0.0.1:8180/healthz --header-file "X-Request-Id=/in/s-core"
  obs_hp "$ic" priv-intel-secret --url http://127.0.0.1:8000/healthz --header-file "X-Request-Id=/in/s-intel"
  # Prompt, context-ref and authorization-ref canaries through the execute endpoint, from core's namespace.
  obs_file h-exec-auth "Bearer $(env_get "$dest" INTELLIGENCE_SERVICE_TOKEN)"
  obs_file b-exec "{\"request_id\":\"$(obs_rid exec)\",\"trace_id\":\"$(random_hex 16)\",\"capability\":\"chat\",\"authorization_ref\":\"authz-${pc[14]}\",\"organization_id\":\"a0000000-0000-4000-8000-000000000001\",\"user_id\":\"c0000000-0000-4000-8000-000000000001\",\"allowed_tools\":[],\"deadline\":\"2999-01-01T00:00:00Z\",\"context_refs\":[{\"kind\":\"doc\",\"ref\":\"ref-${pc[13]}\"}],\"input\":{\"messages\":[{\"role\":\"user\",\"content\":\"prompt ${pc[12]}\"}]}}"
  obs_hp "$core" priv-exec --url http://intelligence:8000/internal/v1/execute --method POST --header "Content-Type: application/json" --header-file "Authorization=/in/h-exec-auth" --body-file /in/b-exec
  rm -f "$OBS_IN/h-exec-auth"

  # --- §6.3 outage matrix: O1-O10 (B-OBS-TIMEOUTS; chain cases (v)-(vii)) -----------------------------
  local tok_ok=1
  obs_outage "$dest" O1 stop postgres yes || tok_ok=0
  obs_outage "$dest" O2 stop redis yes || tok_ok=0
  obs_outage "$dest" O3 stop minio yes || tok_ok=0
  obs_outage "$dest" O4 pause postgres yes || tok_ok=0
  obs_outage "$dest" O5 pause redis yes || tok_ok=0
  obs_outage "$dest" O6 pause minio yes || tok_ok=0
  obs_outage "$dest" O7 stop intelligence yes || tok_ok=0
  obs_outage "$dest" O8 pause intelligence yes || tok_ok=0
  obs_outage "$dest" O9 stop core yes || tok_ok=0
  obs_outage "$dest" O10 pause core yes || tok_ok=0
  # (vi) after O7's recovery the three-hop chain is restored.
  rid="$(obs_rid chain-vi)"; tp="$(obs_tp 01)"
  obs_hp "$web" chain-vi --url http://127.0.0.1:3000/api/ready --header "X-Request-Id: $rid" --header "traceparent: $tp"
  obs_case chain-vi "$rid" 1:1:1 "$(cut -d- -f2 <<<"$tp")" "$(cut -d- -f3 <<<"$tp")" 01
  sleep 3
  obs_logs "$dest"
  # §6.3 web core_probe per case, then O10's late core record. The F11 nested-deadline race applies to
  # paused (O4-O6, O8) and, by the 2026-10-06 G003 correction-1 direction, to stopped services too: a
  # lookup of a stopped container's name can take the whole 2 s deadline, so web may time out first.
  # O1-O3 not_ready or timeout; O7 200/ok or 503/timeout; O9 unreachable or timeout (observed values
  # are recorded in obs-outages.tsv and the result line).
  obs_web_probe O1 not_ready timeout || tok_ok=0; obs_web_probe O2 not_ready timeout || tok_ok=0; obs_web_probe O3 not_ready timeout || tok_ok=0
  obs_web_probe O4 not_ready timeout || tok_ok=0; obs_web_probe O5 not_ready timeout || tok_ok=0; obs_web_probe O6 not_ready timeout || tok_ok=0
  obs_web_probe O7 200/ok 503/timeout || tok_ok=0; obs_web_probe O8 ok timeout || tok_ok=0
  obs_web_probe O9 unreachable timeout || tok_ok=0; obs_web_probe O10 timeout || tok_ok=0
  obs_o10 O10 || tok_ok=0

  # Context: every invalid literal absent from every raw log; the records exist (chain below).
  local ok_ctx=$ok; obs_open
  vtools -v "$OBS:/o:ro" -v "$OBS_IN:/i:ro" "$VT_IMG" logcheck scan --canaries /i/invalid-literals.txt /o/web.log /o/core.log /o/intelligence.log >"$OBS/ctx-scan.log" 2>&1 || { ok_ctx=0; log "an invalid diagnostic value was reflected in a log: $(head -3 "$OBS/ctx-scan.log" | tr '\n' ' ')"; }
  # Chain: one logcheck chain run over all cases.
  obs_chain_file; obs_open
  vtools -v "$OBS:/o:ro" -v "$OBS_IN:/i:ro" "$VT_IMG" logcheck chain --cases /o/chain.json --log web=/o/web.log --log core=/o/core.log --log intelligence=/o/intelligence.log >"$OBS/chain-result.json" 2>"$OBS/chain.err" && ok=1 || ok=0
  yq_stdin -p json -o tsv '.cases[] | [.id, .pass, ((.failures // []) | join(","))]' <"$OBS/chain-result.json" >"$WORK/obs-chain.tsv" 2>/dev/null || ok=0
  [[ $iv_ok == 1 && $iv_n == 32 && "$(awk -F'\t' '$1 ~ /^chain-iv-/ && $2 == "true"' "$WORK/obs-chain.tsv" | wc -l)" == 32 ]] || { ok=0; log "chain-iv: $iv_n of 32 concurrent cases evaluated"; }
  local ctx_fail; ctx_fail="$(awk -F'\t' '$1 ~ /^(ctx|tp)-/ && $2 != "true"' "$WORK/obs-chain.tsv")"
  [[ -z "$ctx_fail" ]] || { ok_ctx=0; log "context rows without matching records: ${ctx_fail//$'\n'/; }"; }
  # Every context row (7 request-ID rows and 9 traceparent rows) must be present and pass.
  [[ "$(awk -F'\t' '$1 ~ /^(ctx|tp)-/ && $2 == "true"' "$WORK/obs-chain.tsv" | wc -l)" == 16 ]] || { ok_ctx=0; log "context rows: not all 16 evaluated with matching records"; }
  [[ $ok_ctx == 1 ]] && record B-OBS-CONTEXT PASS "enabled copy, through web /api/ready: absent, malformed, oversized (65, 513, 8 KiB), duplicated and comma-joined request IDs each get a generated 32-hex ID with records in web, core and intelligence; zero/uppercase/version 01/ff/extra-field/wrong-length/duplicated/comma-joined traceparents start a new trace without a parent; no invalid literal in any raw log; no 500" || record B-OBS-CONTEXT FAIL "diagnostic context validation"
  [[ $ok == 1 ]] && record B-OBS-CHAIN PASS "logcheck chain over web, core and intelligence logs: $(grep -c . "$WORK/obs-chain.tsv") cases (supplied context with flags 01, no context, invalid replaced, 32 concurrent contexts, O1-O10 outages and O7 recovery): expected record counts, equal request and trace IDs, web parent = supplied or absent, core parent = web span, intelligence parent = core span, distinct spans" || { tail -5 "$OBS/chain.err"; record B-OBS-CHAIN FAIL "cross-service continuity ($(awk -F'\t' '$2!="true"{print $1}' "$WORK/obs-chain.tsv" | tr '\n' ' '))"; }
  [[ $tok_ok == 1 ]] && record B-OBS-TIMEOUTS PASS "O1-O10 on the enabled copy (stop/pause postgres, redis, minio, intelligence, core): statuses as specified (core 503 naming the dependency, O7/O8 core 200 with optional intelligence unavailable, liveness 200), web core_probe within the F11 nested-deadline set for stopped and paused services (observed: $(obs_observed O1 O2 O3 O7 O9)), every single request <= 5000 ms (httpprobe monotonic clock), recovery within 90 s each; see obs-timings.tsv and obs-outages.tsv" || record B-OBS-TIMEOUTS FAIL "outage and timeout matrix"

  # --- B-OBS-LOGS: every line of every service classified; one record per request ---------------------
  ok=1
  local vrc; printf 'service\tresult\n' >"$WORK/obs-logs.tsv"
  for c in web core intelligence; do
    vrc=0; vtools -v "$OBS:/o:ro" -v "$OBS_IN:/i:ro" "$VT_IMG" logcheck validate --service "$c" --file "/o/$c.log" >"$OBS/validate-$c.json" 2>"$OBS/validate-$c.err" || vrc=$?
    # logcheck prints its summary (classes and line numbers only, never content) on stdout.
    (( vrc == 0 )) || { ok=0; log "$c log: logcheck exit $vrc: $(yq_stdin -p json -o tsv '[(.failures // [])[0:5][] | .class + "@" + (.line | to_string)] | join(" ")' <"$OBS/validate-$c.json" 2>/dev/null)"; }
    printf '%s\t%s\n' "$c" "$( (( vrc == 0 )) && echo pass || echo "fail($vrc)")" >>"$WORK/obs-logs.tsv"
  done
  rec="$(obs_record core "$(obs_get chain-i '.x_request_id // ""')")"
  grep -q '"status":200' <<<"$rec" || { ok=0; log "no 200 completion record for chain-i in core"; }
  rec="$(obs_record core "$(awk -F'\t' '$1=="O1"{print $2}' "$OBS/outage-requests.tsv")-c")"
  grep -q '"status":503' <<<"$rec" || { ok=0; log "no handled 503 completion record for O1 in core"; }
  [[ $ok == 1 ]] && record B-OBS-LOGS PASS "enabled copy: every line of the web, core and intelligence raw logs is a catalogued JSON event (or the exact web startup banner); completion records valid and exactly one per request ID; a 200 and a handled 503 recorded on the stack (every running health handler sets its status explicitly; the implicit-200 accounting is proven by core's TestMiddlewareStatusAndResponseBehaviour against a real server)" || record B-OBS-LOGS FAIL "structured logs"

  # --- B-OBS-PRIVACY: no canary in any raw log or retained artifact; positive controls --------------
  ok=1
  vtools -v "$OBS:/o:ro" -v "$OBS_IN:/i:ro" "$VT_IMG" logcheck scan --canaries /i/privacy-canaries.txt /o/web.log /o/core.log /o/intelligence.log >"$OBS/privacy-scan.log" 2>&1 || { ok=0; log "privacy canary in a raw log: $(head -3 "$OBS/privacy-scan.log" | tr '\n' ' ')"; }
  vtools -v "$OBS:/o:ro" -v "$OBS_IN:/i:ro" "$VT_IMG" logcheck scan --canaries /i/secret-ids.txt /o/web.log /o/core.log /o/intelligence.log >"$OBS/secret-scan.log" 2>&1 || { ok=0; log "a secret-valued request ID reached a raw log"; }
  for name in priv-web-health priv-web-ready priv-core-health priv-core-ready priv-intel-health priv-intel-ready priv-core-secret priv-intel-secret; do
    xr="$(obs_get "$name" '.x_request_id // ""')"
    [[ "$xr" =~ ^[0-9a-f]{32}$ ]] && ! grep -qxF -- "$xr" "$OBS_IN/secret-ids.txt" || { ok=0; log "$name: the response ID was not replaced by a generated one"; continue; }
    case "$name" in priv-web-*) c=web ;; priv-core-*) c=core ;; *) c=intelligence ;; esac
    rec="$(obs_record "$c" "$xr")"; [[ "$(grep -c . <<<"$rec")" == 1 ]] && grep -q '"route":"' <<<"$rec" && grep -q '"status":' <<<"$rec" || { ok=0; log "$name: no completion record (diagnostics suppressed?)"; }
  done
  [[ "$(obs_get priv-exec .status)" == 200 ]] || { ok=0; log "execute request with canaries: status $(obs_get priv-exec .status)"; }
  # Positive control: the same scan detects a canary planted in a scratch copy of a log.
  mkdir -p "$OBS/control"; cp "$OBS/core.log" "$OBS/control/core.log"; printf 'planted %s\n' "${pc[5]}" >>"$OBS/control/core.log"; obs_open
  if vtools -v "$OBS:/o:ro" -v "$OBS_IN:/i:ro" "$VT_IMG" logcheck scan --canaries /i/privacy-canaries.txt /o/control/core.log >/dev/null 2>&1; then ok=0; log "privacy control: a planted canary was not detected"; fi
  rm -rf "$OBS/control"
  printf 'check\tresult\nscan\t%s\n' "$([[ $ok == 1 ]] && echo clean || echo finding)" >"$WORK/obs-privacy.tsv"
  [[ $ok == 1 ]] && record B-OBS-PRIVACY PASS "per-run canaries in Authorization, Cookie, X-Api-Key, baggage, tracestate, X-Canary, User-Agent and Referer headers, token= and q= queries, a GET body, invalid IDs, prompt/context-ref/authorization-ref through /internal/v1/execute, and secret-valued request IDs sent to core and intelligence: absent from every raw log; their completion records exist with route and status; a planted canary is detected (positive control)" || record B-OBS-PRIVACY FAIL "privacy"
  rm -rf "$OBS_IN"
}

obs_off() { # obs_off DEST: B-OBS-OFF on the intelligence-disabled copy (web -> core only)
  local dest="$1" ok=1 web core rid tp c body xr k pc=() priv=()
  OBS="$WORK/obs-off"; OBS_IN="$(obs_inputs)/off"; OBS_SVCS=(web core); OBS_CASES=()
  mkdir -p "$OBS" "$OBS_IN"; chmod 755 "$OBS" "$OBS_IN"; : >"$OBS/outage-requests.tsv"
  ensure_tools || { record B-OBS-OFF FAIL "validation tools unavailable"; return; }
  web="$(cid_of "$dest" web)"; core="$(cid_of "$dest" core)"
  rid="$(obs_rid off-chain)"; tp="$(obs_tp 01)"
  obs_hp "$web" off-chain --url http://127.0.0.1:3000/api/ready --header "X-Request-Id: $rid" --header "traceparent: $tp"
  obs_case off-chain "$rid" 1:1 "$(cut -d- -f2 <<<"$tp")" "$(cut -d- -f3 <<<"$tp")" 01
  obs_hp "$core" off-ready --url http://127.0.0.1:8180/readyz
  # Privacy on the disabled variant: canary headers, a canary query and a canary-bearing invalid ID.
  for k in auth cookie xcanary query rid; do pc+=("Cx$(random_hex 12)"); done
  printf '%s\n' "${pc[@]}" >"$OBS_IN/privacy-canaries.txt"; printf '%s\n' "${pc[@]}" >>"$SECRETS_FILE"
  obs_file h-auth "Bearer ${pc[0]}"; obs_file h-cookie "session=${pc[1]}"; obs_file h-xcanary "${pc[2]}"; obs_file h-rid "bad ${pc[4]}"
  obs_file u-web "http://127.0.0.1:3000/api/ready?token=${pc[3]}&q=${pc[3]}"; obs_file u-core "http://127.0.0.1:8180/readyz?token=${pc[3]}&q=${pc[3]}"
  priv=(--header-file "Authorization=/in/h-auth" --header-file "Cookie=/in/h-cookie" --header-file "X-Canary=/in/h-xcanary" --header-file "X-Request-Id=/in/h-rid")
  obs_hp "$web" off-priv-web --url-file /in/u-web "${priv[@]}"; obs_hp "$core" off-priv-core --url-file /in/u-core "${priv[@]}"
  body="$(obs_get off-ready .body)"
  [[ "$(obs_get off-ready .status)" == 200 ]] && ! grep -q optional <<<"$body" || { ok=0; log "disabled core readiness: $body"; }
  local l
  local l lab act svc
  for l in "O1 stop postgres" "O2 stop redis" "O3 stop minio" "O4 pause postgres" "O5 pause redis" "O6 pause minio" "O9 stop core" "O10 pause core"; do
    read -r lab act svc <<<"$l"
    obs_outage "$dest" "off-$lab" "$act" "$svc" no || ok=0
  done
  sleep 3
  obs_logs "$dest"
  OBS_OUT="$WORK/obs-off-outages.tsv"; printf 'case\tweb_status\tcore_probe\tweb_outcome\tcore_records\tcore_outcome\tcore_status\n' >"$OBS_OUT"
  local o expect
  # The F11 nested-deadline interpretation (see obs_ops) for stopped and paused services alike.
  for o in "O1 not_ready timeout" "O2 not_ready timeout" "O3 not_ready timeout" "O4 not_ready timeout" "O5 not_ready timeout" "O6 not_ready timeout" "O9 unreachable timeout" "O10 timeout"; do
    read -r -a expect <<<"$o"
    obs_web_probe "off-${expect[0]}" "${expect[@]:1}" || ok=0
  done
  obs_o10 off-O10 || ok=0
  obs_chain_file; obs_open
  vtools -v "$OBS:/o:ro" -v "$OBS_IN:/i:ro" "$VT_IMG" logcheck chain --cases /o/chain.json --log web=/o/web.log --log core=/o/core.log >"$OBS/chain-result.json" 2>"$OBS/chain.err" || { ok=0; log "disabled-copy chain: $(head -3 "$OBS/chain.err" | tr '\n' ' ')"; }
  for c in web core; do vtools -v "$OBS:/o:ro" -v "$OBS_IN:/i:ro" "$VT_IMG" logcheck validate --service "$c" --file "/o/$c.log" >/dev/null 2>"$OBS/validate-$c.err" || { ok=0; log "disabled $c log unclassified"; }; done
  vtools -v "$OBS:/o:ro" -v "$OBS_IN:/i:ro" "$VT_IMG" logcheck scan --canaries /i/privacy-canaries.txt /o/web.log /o/core.log >"$OBS/privacy-scan.log" 2>&1 || { ok=0; log "disabled variant: a privacy canary reached a raw log"; }
  for c in web core; do
    xr="$(obs_get "off-priv-$c" '.x_request_id // ""')"
    [[ "$xr" =~ ^[0-9a-f]{32}$ ]] && [[ "$(obs_record "$c" "$xr" | grep -c .)" == 1 ]] || { ok=0; log "disabled variant: no generated ID or completion record for the $c canary request"; }
  done
  mkdir -p "$OBS/control"; cp "$OBS/core.log" "$OBS/control/core.log"; printf 'planted %s\n' "${pc[2]}" >>"$OBS/control/core.log"; obs_open
  if vtools -v "$OBS:/o:ro" -v "$OBS_IN:/i:ro" "$VT_IMG" logcheck scan --canaries /i/privacy-canaries.txt /o/control/core.log >/dev/null 2>&1; then ok=0; log "disabled variant privacy control: a planted canary was not detected"; fi
  rm -rf "$OBS/control"
  yq_stdin -p json -o tsv '.cases[] | [.id, .pass]' <"$OBS/chain-result.json" >"$WORK/obs-off.tsv" 2>/dev/null || true
  rm -rf "$OBS_IN"
  [[ $ok == 1 ]] && record B-OBS-OFF PASS "intelligence-disabled copy: core readiness without an optional key; web -> core continuity (one record each, parent = web span); O1-O6, O9, O10 statuses and bounds (each request <= 5000 ms, recovery <= 90 s), web core_probe within the F11 nested-deadline set (observed: $(obs_observed off-O1 off-O2 off-O3 off-O9)); every log line classified; canary headers, query and invalid ID absent from both raw logs while their completion records exist (a planted canary is detected)" || record B-OBS-OFF FAIL "observability on the disabled variant"
}

suite_ops() { log "== ops (disposable copies: observability, outages, persistence; G004 identity)"; variants_ops; res_scope_begin S-IDN; variants_identity; res_scope_end S-IDN; }
suite_rename() { log "== rename (disposable copy: ports, identity, database rename)"; variants_rename; }
suite_off() { log "== off (disposable copy: intelligence disabled)"; variants_off; }

variants_ops() {
  local D="$WORK/ops" project="g001-ops-$RUN_ID"
  if ports_busy "${PORTS_OPS[@]}"; then record B-OUTAGE FAIL "ports ${PORTS_OPS[*]} busy; refusing to re-map silently"; return; fi
  copy_repo "$D"; configure "$D" "$project" "${PORTS_OPS[@]}" || { record B-OUTAGE FAIL "configure"; return; }
  run_docs_block "$D" README.md platform:start "$WORK/ops-start.log" || { tail -30 "$WORK/ops-start.log"; record B-OUTAGE FAIL "start"; return; }
  run_docs_block "$D" README.md platform:migrate "$WORK/ops-migrate.log" || { record B-PERSIST FAIL "migrate"; }
  obs_ops "$D"
  outage "$D" "${PORTS_OPS[1]}" "${PORTS_OPS[0]}"
  # G065 K3: after a failed restart (B-PERSIST FAIL) the copy is down; the checks below need it, so their rows stay
  # missing (B-REQUIRED-CHECKS at final) and the run continues to its joins and results.
  persist "$D" || { save_logs "$D" ops; compose "$D" down -v >/dev/null 2>&1 || true; return; }
  # G004: after G003's log capture and the persistence restart, on the same unconfigured stack.
  identity_unconfigured_probes "$D"
  # G012: last on the disposable copy (it changes the copy's runtime role password and restarts its core).
  credential_rotation "$D" "${PORTS_OPS[1]}"
  save_logs "$D" ops
  compose "$D" down -v >/dev/null 2>&1 || true
}

outage() { # outage DEST CORE_PORT WEB_PORT
  local dest="$1" core="$2" web="$3" ok=1 svc seen i res
  for svc in redis postgres minio; do
    compose "$dest" stop "$svc" >/dev/null 2>&1
    seen=0
    for ((i = 0; i < 30; i++)); do
      res="$(http_get 127.0.0.1 "$core" /readyz)"
      if [[ "$(cut -f1 <<<"$res")" == 503 ]] && grep -q "\"$([[ $svc == minio ]] && echo object_store || echo "$svc")\":\"unavailable\"" <<<"$res"; then seen=1; break; fi
      sleep 1
    done
    [[ $seen == 1 ]] || { ok=0; log "core /readyz did not report the $svc outage"; }
    [[ "$(http_get 127.0.0.1 "$core" /healthz | cut -f1)" == 200 ]] || { ok=0; log "liveness failed during the $svc outage"; }
    compose "$dest" start "$svc" >/dev/null 2>&1
    wait_http 127.0.0.1 "$core" /readyz 200 90 || { ok=0; log "core did not recover after $svc"; }
  done
  compose "$dest" stop intelligence >/dev/null 2>&1; sleep 2
  res="$(http_get 127.0.0.1 "$core" /readyz)"
  [[ "$(cut -f1 <<<"$res")" == 200 ]] && grep -q '"intelligence":"unavailable"' <<<"$res" || { ok=0; log "intelligence outage: $res"; }
  compose "$dest" start intelligence >/dev/null 2>&1
  compose "$dest" stop core >/dev/null 2>&1
  wait_http 127.0.0.1 "$web" /api/ready 503 30 || { ok=0; log "web /api/ready did not report the core outage"; }
  no_store 127.0.0.1 "$web" /api/ready 503 || { ok=0; log "web /api/ready 503 is not no-store"; }
  no_store 127.0.0.1 "$web" /api/health 200 || { ok=0; log "web /api/health during the core outage is not 200 no-store"; }
  compose "$dest" start core >/dev/null 2>&1
  compose "$dest" up -d --wait >/dev/null 2>&1 || { ok=0; log "stack did not recover"; }
  [[ $ok == 1 ]] && record B-OUTAGE PASS "redis/postgres/minio outages -> core 503 naming the dependency with liveness 200, then recovery to 200; intelligence outage reported but not gating; core outage -> web /api/ready 503 (no-store) with /api/health 200 (no-store)" || record B-OUTAGE FAIL "outage behaviour"
}

# G012 B-SECRET-ROTATION (Plan r1 §3.3): one exercised rotation and withdrawal of the core runtime PostgreSQL credential
# on the disposable ops copy, with synthetic values only. Administrative SQL goes to the postgres container's local
# socket as the disposable superuser on stdin (the new value through psql \getenv, never in argv); password logins use
# TCP to the service name (the image's password rule, not the loopback trust lines) with PGPASSWORD passed by name
# through compose exec -e. PostgreSQL does not re-authenticate open connections, so a 200 after the rotation is served
# by a pre-rotation connection; terminating those backends is part of the procedure (docs/SECURITY.md).
rot_sql() { compose "$1" exec -T postgres psql -X -q -At -v ON_ERROR_STOP=1 -U postgres -d "$(env_get "$1" DATABASE_NAME)"; } # SQL on stdin
rot_login() { # rot_login DEST ROLE VALUE: one password login over TCP; prints the result row or psql's error
  PGPASSWORD="$3" compose "$1" exec -T -e PGPASSWORD postgres psql -X -At -h postgres -U "$2" -d "$(env_get "$1" DATABASE_NAME)" -c 'SELECT 1' 2>&1
}
rot_backends() { # rot_backends DEST ROLE CMP T: count of the role's client backends started CMP ('<' or '>') T
  printf "SELECT count(*) FROM pg_stat_activity WHERE usename = '%s' AND backend_type = 'client backend' AND backend_start %s '%s'::timestamptz;\n" "$2" "$3" "$4" | rot_sql "$1"
}
credential_rotation() { # credential_rotation DEST CORE_PORT
  local dest="$1" core="$2" ok=1 role old new t n pre out s r4 r5 r6
  role="$(env_get "$dest" DATABASE_NAME)_runtime"; old="$(env_get "$dest" DB_RUNTIME_PASSWORD)"
  new="Rt$(random_hex 24)"; printf '%s\n' "$new" >>"$SECRETS_FILE"
  # R0: ready, with at least one runtime backend.
  wait_http 127.0.0.1 "$core" /readyz 200 60 || { ok=0; log "R0: core is not ready"; }
  t="$(printf 'SELECT clock_timestamp();\n' | rot_sql "$dest")"
  pre="$(rot_backends "$dest" "$role" '<' "$t")"
  [[ "$pre" =~ ^[0-9]+$ ]] && (( pre >= 1 )) || { ok=0; log "R0: no pre-rotation $role backend ($pre)"; }
  # R1: rotate (the value reaches psql through its environment and \getenv, never argv or a log).
  t="$(printf '%s\n' '\getenv newpw ROT_NEW' "ALTER ROLE \"$role\" PASSWORD :'newpw';" 'SELECT clock_timestamp();' |
    ROT_NEW="$new" compose "$dest" exec -T -e ROT_NEW postgres psql -X -q -At -v ON_ERROR_STOP=1 -U postgres -d "$(env_get "$dest" DATABASE_NAME)" 2>/dev/null)" ||
    { ok=0; log "R1: ALTER ROLE failed"; }
  # R2: the old connection keeps serving.
  [[ "$(http_get 127.0.0.1 "$core" /readyz | cut -f1)" == 200 ]] || { ok=0; log "R2: /readyz is not 200 right after the rotation"; }
  n="$(rot_backends "$dest" "$role" '<' "$t")"; [[ "$n" =~ ^[0-9]+$ ]] && (( n >= 1 )) || { ok=0; log "R2: no pre-rotation backend remains ($n)"; }
  # R3: new logins: the old value fails, the new one works.
  out="$(rot_login "$dest" "$role" "$old")" && { ok=0; log "R3: the old value still logs in"; }
  grep -q 'password authentication failed' <<<"$out" || { ok=0; log "R3: the old value did not fail password authentication"; }
  [[ "$(rot_login "$dest" "$role" "$new" | tr -d '\r\n')" == 1 ]] || { ok=0; log "R3: the new value does not log in"; }
  # R4: terminate the pre-rotation backends; core, still holding the old value, becomes not ready.
  s=$(date +%s)
  n="$(printf "SELECT count(pg_terminate_backend(pid)) FROM pg_stat_activity WHERE usename = '%s' AND backend_type = 'client backend' AND backend_start < '%s'::timestamptz;\n" "$role" "$t" | rot_sql "$dest")"
  [[ "$n" =~ ^[0-9]+$ ]] && (( n >= 1 )) || { ok=0; log "R4: terminated $n backends"; }
  [[ "$(rot_backends "$dest" "$role" '<' "$t")" == 0 ]] || { ok=0; log "R4: pre-rotation backends remain"; }
  wait_http 127.0.0.1 "$core" /readyz 503 30 || { ok=0; log "R4: core did not report 503 within 30 s"; }
  r4="terminated $n backends, 503 after $(( $(date +%s) - s )) s"
  # R5: update the copy's source and restart core; it recovers on fresh connections.
  s=$(date +%s)
  env_set "$dest" DB_RUNTIME_PASSWORD "$new"
  compose "$dest" up -d --no-build --no-deps --force-recreate core >"$WORK/rotation-restart.log" 2>&1 || { ok=0; log "R5: core restart failed"; }
  wait_http 127.0.0.1 "$core" /readyz 200 90 || { ok=0; log "R5: core did not recover within 90 s"; }
  n="$(rot_backends "$dest" "$role" '>' "$t")"; [[ "$n" =~ ^[0-9]+$ ]] && (( n >= 1 )) || { ok=0; log "R5: no post-rotation backend ($n)"; }
  [[ "$(rot_backends "$dest" "$role" '<' "$t")" == 0 ]] || { ok=0; log "R5: a pre-rotation backend reappeared"; }
  r5="200 after $(( $(date +%s) - s )) s on $n post-rotation backends"
  # R6: withdraw without replacement; the consumer fails closed.
  s=$(date +%s)
  printf 'ALTER ROLE "%s" PASSWORD NULL;\n' "$role" | rot_sql "$dest" >/dev/null 2>&1 || { ok=0; log "R6: withdrawal failed"; }
  out="$(rot_login "$dest" "$role" "$new")" && { ok=0; log "R6: the withdrawn value still logs in"; }
  grep -q 'password authentication failed' <<<"$out" || { ok=0; log "R6: the withdrawn value did not fail password authentication"; }
  n="$(printf "SELECT count(pg_terminate_backend(pid)) FROM pg_stat_activity WHERE usename = '%s' AND backend_type = 'client backend';\n" "$role" | rot_sql "$dest")"
  wait_http 127.0.0.1 "$core" /readyz 503 30 || { ok=0; log "R6: core did not report 503 within 30 s after the withdrawal"; }
  r6="terminated ${n:-?} backends, 503 after $(( $(date +%s) - s )) s"
  if grep -rqF -f "$SECRETS_FILE" "$WORK/rotation-restart.log" 2>/dev/null; then ok=0; log "a credential value reached the restart log"; fi
  [[ $ok == 1 ]] && record B-SECRET-ROTATION PASS "disposable ops copy, synthetic values: R0 ready with $pre runtime backends; R1 ALTER ROLE through \\getenv; R2 the pre-rotation connection still serves /readyz 200 (no re-authentication); R3 old value fails password authentication, new value logs in; R4 $r4; R5 source updated and core recreated: $r5; R6 PASSWORD NULL: login fails, $r6" ||
    record B-SECRET-ROTATION FAIL "credential rotation and withdrawal rehearsal"
}

persist() { # persist DEST
  local dest="$1" ok=1 out mc
  mc='export MC_CONFIG_DIR=/tmp/mc MC_HOST_obj="http://${OBJECT_STORE_ACCESS_KEY}:${OBJECT_STORE_SECRET_KEY}@minio:9000"'
  compose "$dest" run --rm --no-deps --entrypoint /bin/sh minio-init -c "$mc; printf kept | mc pipe \"obj/\${OBJECT_STORE_BUCKET}/g001-marker.txt\" >/dev/null" >"$WORK/persist-put.log" 2>&1 || { ok=0; log "minio put (scoped credential) failed"; }
  out="$(compose "$dest" exec -T redis sh -c 'REDISCLI_AUTH="$REDIS_PASSWORD" redis-cli SET g001-marker kept' | tr -d '\r\n')" && [[ "$out" == OK ]] || { ok=0; log "redis SET: $out"; }
  out="$(compose "$dest" exec -T redis sh -c 'REDISCLI_AUTH="$REDIS_PASSWORD" redis-cli GET g001-marker' | tr -d '\r\n')" && [[ "$out" == kept ]] || { ok=0; log "redis GET before restart: $out"; }
  run_docs_block "$dest" README.md platform:stop "$WORK/persist-stop.log" || ok=0
  # G065 K3: a failed restart is this check's failure, returned to the caller; the later reads would fail on a stopped
  # stack and end the harness (set -e) before the background joins and the results.
  run_docs_block "$dest" README.md platform:start "$WORK/persist-start.log" || { ok=0; record B-PERSIST FAIL "restart: platform:start failed (see persist-start.log)"; return 1; }
  out="$(compose_migrate "$dest" run --rm migrate status 2>&1)" || ok=0
  grep -qP '^migration\t00001_baseline\.sql\tapplied\t' <<<"$out" && grep -qP '^migration\t00002_identity\.sql\tapplied\t' <<<"$out" && grep -qP '^migration\t00003_tenancy\.sql\tapplied\t' <<<"$out" && grep -qP '^migration\t00004_business_profiles\.sql\tapplied\t' <<<"$out" && grep -qP '^migration\t00005_organization_policy\.sql\tapplied\t' <<<"$out" && grep -qP '^migration\t00006_row_level_security\.sql\tapplied\t' <<<"$out" && grep -qP '^migration\t00007_stored_objects\.sql\tapplied\t' <<<"$out" && grep -qP '^migration\t00008_audit_events\.sql\tapplied\t' <<<"$out" && grep -qP '^migration\t00009_retention_maintenance\.sql\tapplied\t' <<<"$out" && grep -qP '^migration\t00010_mcp_connections\.sql\tapplied\t' <<<"$out" && grep -qP '^migration\t00011_organization_domains\.sql\tapplied\t' <<<"$out" && grep -qP '^migration\t00012_session_binding_and_claim_exclusivity\.sql\tapplied\t' <<<"$out" && grep -qP '^migration\t00013_organization_branding\.sql\tapplied\t' <<<"$out" && grep -qP '^migration\t00014_group_onboarding\.sql\tapplied\t' <<<"$out" && grep -qP '^migration\t00015_membership_invitations\.sql\tapplied\t' <<<"$out" && grep -qP '^migration\t00016_organization_enrollment\.sql\tapplied\t' <<<"$out" && grep -qP '^migration\t00017_group_authority\.sql\tapplied\t' <<<"$out" && grep -qP '^migration\t00018_platform_signup\.sql\tapplied\t' <<<"$out" && grep -qP '^migration\t00019_mcp_connection_names\.sql\tapplied\t' <<<"$out" && grep -qP '^migration\t00020_labels\.sql\tapplied\t' <<<"$out" && grep -qP '^migration\t00021_platform_labels\.sql\tapplied\t' <<<"$out" && grep -qP '^migration\t00022_default_tenant\.sql\tapplied\t' <<<"$out" && grep -qP '^current_version\t22$' <<<"$out" || { ok=0; log "migration history lost: $out"; }
  out="$(compose "$dest" run --rm --no-deps --entrypoint /bin/sh minio-init -c "$mc; mc cat \"obj/\${OBJECT_STORE_BUCKET}/g001-marker.txt\"" 2>/dev/null)" || ok=0
  [[ "$out" == kept ]] || { ok=0; log "minio object lost"; }
  if out="$(compose "$dest" exec -T redis sh -c 'REDISCLI_AUTH="$REDIS_PASSWORD" redis-cli GET g001-marker')"; then
    [[ -z "$(tr -d '\r\n' <<<"$out")" ]] || { ok=0; log "redis kept data across restart"; }
  else ok=0; log "redis GET after restart failed"; fi
  [[ $ok == 1 ]] && record B-PERSIST PASS "after README stop/start: migration history (version 22: 00001_baseline, 00002_identity, 00003_tenancy, 00004_business_profiles, 00005_organization_policy, 00006_row_level_security, 00007_stored_objects, 00008_audit_events, 00009_retention_maintenance, 00010_mcp_connections, 00011_organization_domains, 00012_session_binding_and_claim_exclusivity, 00013_organization_branding, 00014_group_onboarding, 00015_membership_invitations, 00016_organization_enrollment, 00017_group_authority, 00018_platform_signup, 00019_mcp_connection_names, 00020_labels, 00021_platform_labels and 00022_default_tenant) and a MinIO object (written with the scoped service credential) survive; a Redis key does not (ephemeral)" || record B-PERSIST FAIL "persistence behaviour"
}

variants_rename() {
  local D="$WORK/ren" project="g001-ren-$RUN_ID" ok=1 before after res
  if ports_busy "${PORTS_REN[@]}"; then record B-RENAME FAIL "ports ${PORTS_REN[*]} busy; refusing to re-map silently"; return; fi
  copy_repo "$D"
  [[ "$(impl_tree_hash "$D")" == "$(impl_tree_hash "$REPO")" ]] && record B-SOURCE-HASH PASS "disposable copy is byte-identical to the implementation ($(impl_tree_hash "$REPO" | cut -c1-16))" || record B-SOURCE-HASH FAIL "copy differs from the implementation"
  configure "$D" "$project" "${PORTS_REN[@]}" || { record B-RENAME FAIL "configure"; return; }
  run_docs_block "$D" README.md platform:start "$WORK/ren-start.log" || { tail -30 "$WORK/ren-start.log"; record B-RENAME FAIL "start"; return; }
  probe_ports OVERRIDE "$D" "${PORTS_REN[@]}"
  # The browser image is built before the rename, so the rename itself builds nothing.
  compose_browser "$D" --profile browser build web-browser >"$WORK/ren-browser-build.log" 2>&1 || { ok=0; tail -20 "$WORK/ren-browser-build.log"; }

  before="$(compose "$D" images -q | LC_ALL=C sort | tr '\n' ' ')"
  env_set "$D" PRODUCT_NAME '"Renamed Example"'
  env_set "$D" PLATFORM_BASE_DOMAIN renamed.example.net
  env_set "$D" CUSTOM_DOMAIN_CNAME_TARGET custom.renamed.example.net
  compose "$D" up -d --no-build --wait >"$WORK/ren-up.log" 2>&1 || { ok=0; tail -20 "$WORK/ren-up.log"; }
  after="$(compose "$D" images -q | LC_ALL=C sort | tr '\n' ' ')"
  [[ "$before" == "$after" ]] || { ok=0; log "images changed during rename"; }
  res="$(http_get 127.0.0.1 "${PORTS_REN[0]}" /)"
  grep -q '<title>Renamed Example</title>' <<<"$res" && grep -q 'renamed.example.net' <<<"$res" || { ok=0; log "web did not show the renamed identity"; }
  grep -q '"product_name":"Renamed Example"' <<<"$(http_get 127.0.0.1 "${PORTS_REN[1]}" /readyz)" || { ok=0; log "core did not show the renamed identity"; }
  [[ $ok == 1 ]] && record B-RENAME PASS "product name, base domain and CNAME target changed in .env only; up --no-build; image IDs unchanged ($(wc -w <<<"$before") images); web title/domain and core readiness show the new identity" || record B-RENAME FAIL "rename"

  ok=1
  local brep="$WORK/browser-rename" bimg runc
  runc="g001-browser-$RUN_ID-$(random_hex 4)"
  bimg="$(docker image inspect -f '{{.Id}}' "$project-web-browser" 2>/dev/null)"
  mkdir -p "$brep"
  compose_browser "$D" --profile browser run --name "$runc" --label "g001.run=$RUN_ID" --no-deps -e PLAYWRIGHT_JSON_OUTPUT_FILE=/tmp/report/results.json -e PLAYWRIGHT_OUTPUT_DIR=/tmp/report/out \
    -e PLAYWRIGHT_WORKERS="$PW_WORKERS" web-browser npx playwright test --grep @identity --project=shell-light --project=shell-dark >"$WORK/ren-browser.log" 2>&1 || { ok=0; tail -30 "$WORK/ren-browser.log"; }
  browser_collect_report "$runc" "$brep" "$WORK/ren-browser.log" || ok=0
  browser_matrix "$brep/results.json" "$WORK/browser-rename-tests.tsv" "${BROWSER_MATRIX_IDENTITY[@]}" || ok=0
  [[ -n "$bimg" && "$(docker image inspect -f '{{.Id}}' "$project-web-browser" 2>/dev/null)" == "$bimg" ]] || { ok=0; log "the browser image changed during the rename run"; }
  [[ "$after" == "$(compose "$D" images -q | LC_ALL=C sort | tr '\n' ' ')" ]] || { ok=0; log "service images changed during the rename browser run"; }
  [[ $ok == 1 ]] && record B-BROWSER-RENAME PASS "@identity scenarios pass in shell-light and shell-dark with the renamed product name and domain after the .env-only rename and up --no-build; browser and service image IDs unchanged" || record B-BROWSER-RENAME FAIL "browser identity after rename"

  ok=1
  env_set "$D" DATABASE_NAME renamed_db
  compose "$D" down -v >/dev/null 2>&1
  compose "$D" up -d --no-build --wait >"$WORK/ren-db.log" 2>&1 || { ok=0; tail -20 "$WORK/ren-db.log"; }
  res="$(psql_super "$D" "SELECT string_agg(rolname, ',' ORDER BY rolname) FROM pg_roles WHERE rolname LIKE 'renamed\_db\_%'")"
  [[ "$res" == "renamed_db_auth_reader,renamed_db_claim_writer,renamed_db_migrator,renamed_db_owner,renamed_db_policy_reader,renamed_db_provision_writer,renamed_db_retention,renamed_db_runtime" ]] || { ok=0; log "roles after the database rename: $res"; }
  wait_http 127.0.0.1 "${PORTS_REN[1]}" /readyz 200 60 || { ok=0; log "core not ready after the database rename"; }
  res="$(compose_migrate "$D" run --rm --build migrate up 2>&1)"
  grep -qP '^current_version\t22$' <<<"$res" || { ok=0; log "migrate up after rename: $res"; }
  [[ "$(catalog_p "$D")" == "objects=2 owners=renamed_db_owner extra_acl=0 table_privs=0 seq_privs=0" ]] || { ok=0; log "metadata after rename: $(catalog_p "$D")"; }
  [[ $ok == 1 ]] && record B-DBRENAME PASS "DATABASE_NAME change on a fresh volume: renamed_db with renamed_db_{migrator,owner,retention,runtime} and the G008 helpers renamed_db_{policy_reader,auth_reader} and the G019 helper renamed_db_claim_writer and the G062 helper renamed_db_provision_writer; core ready; migrate up (to version 22) as renamed_db_migrator -> metadata owned by renamed_db_owner and owner-only" || record B-DBRENAME FAIL "database rename"
  save_logs "$D" ren
  compose "$D" down -v >/dev/null 2>&1 || true
}

# control_image IMAGE: a pinned public image that a detector control needs, in the image store. A fresh daemon's
# BuildKit builds keep base images in the build cache only, so the control obtains its image itself. Prints present,
# pulled or unavailable; returns nonzero when unavailable (the caller fails closed).
control_image() {
  if docker image inspect "$1" >/dev/null 2>&1; then echo present; return 0; fi
  if docker pull -q "$1" >/dev/null 2>&1 && docker image inspect "$1" >/dev/null 2>&1; then echo pulled; return 0; fi
  echo unavailable; return 1
}

# The intelligence-disabled variant is derived from the maintained repository instead of a generator render (the
# retired generator's toggle-render composition and its manifest comparisons). off_edit FILE FIRST COUNT
# REPLACEMENT: FILE with the COUNT-line run starting at the only line equal to FIRST replaced by the line REPLACEMENT,
# or removed when REPLACEMENT is "-"; fails unless FIRST occurs exactly once and the run is complete. The values reach
# awk through its environment (never -v), so backslashes stay literal.
off_edit() {
  OFF_FIRST="$2" OFF_REPL="$4" awk -v count="$3" '
    BEGIN { first = ENVIRON["OFF_FIRST"]; repl = ENVIRON["OFF_REPL"] }
    $0 == first { seen++; if (seen == 1) { if (repl != "-") print repl; skip = count } }
    skip > 0 { skip--; next }
    { print }
    END { exit !(seen == 1 && skip == 0) }' "$1"
}
# off_rows: the rows of infra/validation/intelligence-off.tsv (path, count, first line, replacement; tab-separated).
off_rows() { grep -v -e '^#' -e '^$' "$OFF_EDITS"; }
# off_derive DEST: DEST becomes a copy of the repository without the intelligence-only roots, with each row applied in
# order. Returns 1 (reasons logged) when a root is missing from the repository, a row is malformed or a run does not
# occur exactly once.
off_derive() {
  local O="$1" ok=1 r path count first repl n=0
  rm -rf -- "$O"; copy_repo "$O"
  for r in "${OFF_REMOVED[@]}"; do
    [[ -e "$REPO/$r" ]] || { ok=0; log "intelligence-only root $r is missing from the repository"; }
    rm -rf -- "${O:?}/$r"
  done
  [[ -f "$OFF_EDITS" && ! -L "$OFF_EDITS" ]] || { log "infra/validation/intelligence-off.tsv is missing"; return 1; }
  while IFS=$'\t' read -r path count first repl; do
    n=$((n + 1))
    if [[ ! "$count" =~ ^[1-9][0-9]{0,2}$ || -z "$first" || -z "$repl" || "$path" == /* || "/$path/" == */../* ||
      -L "$O/$path" || ! -f "$O/$path" ]]; then ok=0; log "intelligence-off row $n ($path) is malformed or names no regular file"; continue; fi
    if off_edit "$O/$path" "$first" "$count" "$repl" >"$WORK/off-edit.tmp"; then cat "$WORK/off-edit.tmp" >"$O/$path"
    else ok=0; log "intelligence-off row $n: the $count-line run of $path is not present exactly once"; fi
  done < <(off_rows)
  rm -f -- "$WORK/off-edit.tmp"
  (( n > 0 )) || { ok=0; log "infra/validation/intelligence-off.tsv has no rows"; }
  return $((1 - ok))
}

OFF_CODES=""
variants_off() {
  local O="$WORK/off" project="g001-off-$RUN_ID" ok=1 out f r path count first repl nsame=0 nedit=0 nfiles
  OFF_CODES=""
  if ports_busy "${PORTS_OFF[@]}"; then record B-TOGGLE-OFF FAIL "ports busy; refusing to re-map silently"; return; fi
  # B-TOGGLE-OFF-ADDITIONS: the variant is exactly the repository minus the intelligence-only roots,
  # with exactly the reviewed opt-out edits; every other implementation file is byte-identical (links: same target).
  local derived=1
  off_derive "$O" || { ok=0; derived=0; }
  local -A edited=()
  while IFS=$'\t' read -r path count first repl; do edited["$path"]=1; done < <(off_rows)
  for r in "${OFF_REMOVED[@]}"; do [[ ! -e "$O/$r" && ! -L "$O/$r" ]] || { ok=0; log "variant still holds $r"; }; done
  while IFS= read -r -d '' f; do
    [[ -e "$REPO/$f" ]] || continue
    for r in "${OFF_REMOVED[@]}"; do [[ "$f" == "$r" || "$f" == "$r"/* ]] && continue 2; done
    if [[ -n "${edited[$f]+x}" ]]; then
      nedit=$((nedit + 1))
      cmp -s "$REPO/$f" "$O/$f" && { ok=0; log "variant $f is unchanged by its intelligence-off edits"; }
    elif [[ -L "$REPO/$f" ]]; then
      [[ -L "$O/$f" && "$(readlink "$O/$f")" == "$(readlink "$REPO/$f")" ]] || { ok=0; log "variant link $f differs from the repository"; }
      nsame=$((nsame + 1))
    else
      cmp -s "$REPO/$f" "$O/$f" || { ok=0; log "variant $f differs from the repository"; }
      nsame=$((nsame + 1))
    fi
  done < <(impl_files)
  nfiles="$(find "$O" \( -type f -o -type l \) | wc -l)"
  [[ "$nfiles" == $((nsame + nedit)) && "$nedit" == "${#edited[@]}" ]] || { ok=0; log "variant holds $nfiles files and links, want $((nsame + nedit)) ($nedit of ${#edited[@]} edited paths)"; }
  # Detector controls on the first row: applying it again to the variant (the run is gone) and to a copy of the
  # repository file with its first line repeated must both be rejected.
  IFS=$'\t' read -r path count first repl < <(off_rows) || true
  if [[ -n "${path:-}" && -f "$O/$path" ]]; then
    off_edit "$O/$path" "$first" "$count" "$repl" >/dev/null && { ok=0; log "control: a repeated intelligence-off edit of $path was accepted"; }
    { cat "$REPO/$path"; printf '%s\n' "$first"; } >"$WORK/off-control.tmp"
    off_edit "$WORK/off-control.tmp" "$first" "$count" "$repl" >/dev/null && { ok=0; log "control: a repeated first line of $path was accepted"; }
    rm -f -- "$WORK/off-control.tmp"
  else ok=0; log "control: no first intelligence-off row to exercise"; fi
  [[ $ok == 1 ]] && record B-TOGGLE-OFF-ADDITIONS PASS "variant = the repository without the ${#OFF_REMOVED[@]} intelligence-only roots (${OFF_REMOVED[*]}), with the $(off_rows | grep -c .) reviewed opt-out edits of infra/validation/intelligence-off.tsv applied exactly once to $nedit files (each changed); the other $nsame implementation files and links byte-identical to the repository, nothing else present (controls: a repeated edit and a repeated first line are rejected)" || record B-TOGGLE-OFF-ADDITIONS FAIL "variant derivation"
  (( derived )) || { record B-TOGGLE-OFF FAIL "the intelligence-disabled variant could not be derived (see B-TOGGLE-OFF-ADDITIONS)"; return; }

  ok=1
  configure "$O" "$project" "${PORTS_OFF[@]}" || { record B-TOGGLE-OFF FAIL "configure"; return; }
  local gs=docs/foundation/GETTING-STARTED.md
  # G051 build ledger: the variant's Compose and documented-block builds carry the reviewed B-TOGGLE-OFF exception
  # (explicit ledger_build calls keep their own kinds); restored at the end of this function.
  LEDGER_KIND=named:B-TOGGLE-OFF
  run_docs_block "$O" "$gs" archetype:start "$WORK/off-start.log" || { ok=0; OFF_CODES+="start "; tail -30 "$WORK/off-start.log"; }
  run_docs_block "$O" docs/foundation/MIGRATIONS.md migrations:apply "$WORK/off-migrate.log" && grep -qP '^current_version\t22$' "$WORK/off-migrate.log" || { ok=0; OFF_CODES+="migrate-apply "; log "migrate up in the disabled variant"; }
  run_docs_block "$O" docs/foundation/MIGRATIONS.md migrations:inspect "$WORK/off-inspect.log" &&
    grep -qP '^migration\t00001_baseline\.sql\tapplied\t' "$WORK/off-inspect.log" && grep -qP '^migration\t00002_identity\.sql\tapplied\t' "$WORK/off-inspect.log" &&
    grep -qP '^migration\t00003_tenancy\.sql\tapplied\t' "$WORK/off-inspect.log" &&
    grep -qP '^migration\t00004_business_profiles\.sql\tapplied\t' "$WORK/off-inspect.log" &&
    grep -qP '^migration\t00005_organization_policy\.sql\tapplied\t' "$WORK/off-inspect.log" &&
    grep -qP '^migration\t00006_row_level_security\.sql\tapplied\t' "$WORK/off-inspect.log" &&
    grep -qP '^migration\t00007_stored_objects\.sql\tapplied\t' "$WORK/off-inspect.log" &&
    grep -qP '^migration\t00008_audit_events\.sql\tapplied\t' "$WORK/off-inspect.log" &&
    grep -qP '^migration\t00009_retention_maintenance\.sql\tapplied\t' "$WORK/off-inspect.log" &&
    grep -qP '^migration\t00010_mcp_connections\.sql\tapplied\t' "$WORK/off-inspect.log" &&
    grep -qP '^migration\t00011_organization_domains\.sql\tapplied\t' "$WORK/off-inspect.log" &&
    grep -qP '^migration\t00012_session_binding_and_claim_exclusivity\.sql\tapplied\t' "$WORK/off-inspect.log" &&
    grep -qP '^migration\t00013_organization_branding\.sql\tapplied\t' "$WORK/off-inspect.log" &&
    grep -qP '^migration\t00014_group_onboarding\.sql\tapplied\t' "$WORK/off-inspect.log" &&
    grep -qP '^migration\t00015_membership_invitations\.sql\tapplied\t' "$WORK/off-inspect.log" &&
    grep -qP '^migration\t00016_organization_enrollment\.sql\tapplied\t' "$WORK/off-inspect.log" &&
    grep -qP '^migration\t00017_group_authority\.sql\tapplied\t' "$WORK/off-inspect.log" &&
    grep -qP '^migration\t00018_platform_signup\.sql\tapplied\t' "$WORK/off-inspect.log" &&
    grep -qP '^migration\t00019_mcp_connection_names\.sql\tapplied\t' "$WORK/off-inspect.log" &&
    grep -qP '^migration\t00020_labels\.sql\tapplied\t' "$WORK/off-inspect.log" &&
    grep -qP '^migration\t00021_platform_labels\.sql\tapplied\t' "$WORK/off-inspect.log" &&
    grep -qP '^migration\t00022_default_tenant\.sql\tapplied\t' "$WORK/off-inspect.log" &&
    [[ "$(grep -cP '^current_version\t22$' "$WORK/off-inspect.log")" == 2 ]] || { ok=0; OFF_CODES+="migrate-inspect "; log "migrate status/version in the disabled variant"; }
  compose "$O" config --services | grep -qx intelligence && { ok=0; OFF_CODES+="service "; log "intelligence service present"; }
  # Built images are named <project>-<service>, so inspect their sources and contents.
  compose "$O" config --format json 2>/dev/null | yq_stdin -p json -o tsv '.services[] | ((.build.dockerfile // "") + " " + (.image // ""))' | grep -qiE 'intelligence|python' && { ok=0; OFF_CODES+="image "; log "python or intelligence image configured"; }
  # Detector control: python:3.13.16-slim (the intelligence base image) must be detectable by PYTHON_VERSION, so the
  # running-container check below is not vacuous. Its image is obtained if absent; an unavailable image fails closed.
  local pyctl
  pyctl="$(control_image python:3.13.16-slim)" || true
  docker image inspect -f '{{range .Config.Env}}{{println .}}{{end}}' python:3.13.16-slim 2>/dev/null | grep -q '^PYTHON_VERSION=' || { ok=0; OFF_CODES+="control "; log "control: python base image not detectable by PYTHON_VERSION ($pyctl)"; }
  grep -qE 'AI_MODEL_|OPENROUTER_API_KEY|INTELLIGENCE_' "$O/.env.example" "$O/infra/compose/local.yml" && { ok=0; OFF_CODES+="config "; log "AI or intelligence configuration present"; }
  [[ -e "$O/intelligence" || -e "$O/contracts/intelligence" || -e "$O/infra/docker/intelligence.Dockerfile" ]] && { ok=0; OFF_CODES+="files "; log "intelligence files present"; }
  local cid
  for cid in $(docker ps -q --filter "label=com.docker.compose.project=$project"); do
    docker image inspect -f '{{range .Config.Env}}{{println .}}{{end}}' "$(docker inspect -f '{{.Image}}' "$cid")" | grep -q '^PYTHON_VERSION=' && { ok=0; OFF_CODES+="pycontainer "; log "python-based container running"; }
  done
  docker inspect -f '{{range .Config.Env}}{{println .}}{{end}}' "$(cid_of "$O" core)" | grep -qE '^(INTELLIGENCE_|AI_MODEL_|OPENROUTER)' && { ok=0; OFF_CODES+="core-settings "; log "core holds intelligence settings"; }
  out="$(http_get 127.0.0.1 "${PORTS_OFF[1]}" /readyz)"
  [[ "$(cut -f1 <<<"$out")" == 200 ]] && ! grep -q intelligence <<<"$out" || { ok=0; OFF_CODES+="ready "; log "core readiness in the disabled variant: $out"; }
  run_docs_block "$O" "$gs" archetype:check "$WORK/off-check.log" || { ok=0; OFF_CODES+="check "; tail -30 "$WORK/off-check.log"; }
  obs_off "$O"
  # B-BROWSER-OFF: the full browser suite against the disabled variant.
  local bok=1
  browser_run "$O" "$WORK/browser-off" "$WORK/off-browser.log" || { bok=0; tail -30 "$WORK/off-browser.log"; }
  browser_matrix "$WORK/browser-off/results.json" "$WORK/browser-off-tests.tsv" "${BROWSER_MATRIX_FULL[@]}" || bok=0
  init_hash_documented "$WORK/browser-off/results.json" "$O/docs/foundation/WEB-SHELL.md" || bok=0
  [[ $bok == 1 ]] && record B-BROWSER-OFF PASS "full browser suite on the intelligence-disabled variant: $(grep -c . "$WORK/browser-off-tests.tsv") tests, every required scenario ID passes in each required project, 0 skipped/flaky/unexpected" || record B-BROWSER-OFF FAIL "browser suite on the disabled variant"
  save_logs "$O" off
  run_docs_block "$O" "$gs" archetype:stop "$WORK/off-stop.log" || { ok=0; OFF_CODES+="stop "; log "disabled variant archetype:stop failed; see diagnostics/off-stop.log"; }
  [[ $ok == 1 ]] && record B-TOGGLE-OFF PASS "disabled variant: GETTING-STARTED start/check/stop and MIGRATIONS apply verbatim; migrate up to version 22 (00001_baseline, 00002_identity, 00003_tenancy, 00004_business_profiles, 00005_organization_policy, 00006_row_level_security, 00007_stored_objects, 00008_audit_events, 00009_retention_maintenance, 00010_mcp_connections, 00011_organization_domains, 00012_session_binding_and_claim_exclusivity, 00013_organization_branding, 00014_group_onboarding, 00015_membership_invitations, 00016_organization_enrollment, 00017_group_authority, 00018_platform_signup, 00019_mcp_connection_names, 00020_labels, 00021_platform_labels and 00022_default_tenant); no intelligence service, Python image or container, AI/intelligence configuration or files; core ready without an intelligence entry; the PYTHON_VERSION detector control holds (python:3.13.16-slim $pyctl)" || record B-TOGGLE-OFF FAIL "intelligence-disabled variant (detector control image $pyctl); subconditions: $(off_fail_codes)"
  compose "$O" down -v >/dev/null 2>&1 || true
  LEDGER_KIND=authoritative
}

# ===================================================================================
# G004 identity (Plan r5 §8.2–8.5). B-IDENTITY-GUARDS (lint) and B-IDENTITY-CATALOG (migrations); B-IDENTITY-IT, -E2E,
# -PROBES, -PRIVACY and -CONFIG on a separate disposable identity project (local.yml + infra/compose/identity.yml,
# IDENTITY.md's identity:* blocks) after variants_ops. Secrets reach containers only through files; the run's generated
# identity secrets, the browser tests' observed values and the probe handoff join the run's secret registry.
PORTS_IDN=(3600 8680 8600 6032 6979 9600 9601)
IDN_UNCONF="" IDN_P="" IDN_IN=""
# The 13 reviewed G005 tenancy relations of 00003_tenancy.sql (relkind:name), all owned by the owner role.
TENANCY_OBJECTS="i:groups_organization_id_key i:groups_organization_slug_key i:groups_pkey i:organizations_pkey i:role_bindings_group_idx i:role_bindings_pkey i:role_bindings_scope_idx i:role_bindings_subject_scope_key r:groups r:organizations r:role_bindings v:group_memberships v:organization_memberships"
# The 8 reviewed G006 BusinessProfile relations of 00004_business_profiles.sql (relkind:name), all owned by the owner role.
BUSINESS_PROFILE_OBJECTS="i:business_profiles_organization_id_key i:business_profiles_pkey i:group_business_profiles_business_profile_idx i:group_business_profiles_group_key i:group_business_profiles_pkey r:business_profiles r:group_business_profiles v:business_profile_access_paths"
# The 3 reviewed G009 stored-object relations of 00007_stored_objects.sql (relkind:name), all owned by the owner role.
STORED_OBJECT_OBJECTS="i:stored_objects_organization_group_idx i:stored_objects_pkey r:stored_objects"
# The 4 reviewed G014 MCP connection relations of 00010_mcp_connections.sql (relkind:name), all owned by the owner role.
MCP_OBJECTS="i:mcp_connections_pkey i:mcp_connections_resource_key i:mcp_connections_session_key r:mcp_connections"
# The 6 reviewed G016 relations of 00011_organization_domains.sql (relkind:name), all owned by the owner role, as at
# version 12: 00012 replaces the live-hostname index by organization_domains_hostname_owned (proof-time exclusivity).
DOMAIN_OBJECTS="i:organization_domains_hostname_generation_key i:organization_domains_hostname_owned i:organization_domains_organization i:organization_domains_pkey i:organizations_slug_key r:organization_domains"
# The 2 reviewed G017 branding relations of 00013_organization_branding.sql (relkind:name), owned by the owner role.
BRANDING_OBJECTS="i:organization_branding_pkey r:organization_branding"
# The 6 reviewed G019 invitation relations of 00015_membership_invitations.sql (relkind:name), owned by the owner role.
MEMBERSHIP_OBJECTS="i:invitations_pending_idx i:invitations_pkey i:invitations_recipient_idx i:invitations_scope_idx i:invitations_token_digest_key r:invitations"
# The 15 reviewed G058 enrollment relations of 00016_organization_enrollment.sql (relkind:name; enrollmentRelations of
# core/internal/migrate/integration_test.go), owned by the owner role: three tables, their primary keys, the
# credential-digest and seq unique keys, the three 1a partial indexes, the request table's one_pending, scope and user
# indexes and the ledger's identity sequence.
ENROLLMENT_OBJECTS="S:enrollment_restrictions_seq_seq i:enrollment_requests_one_pending i:enrollment_requests_pkey i:enrollment_requests_scope_idx i:enrollment_requests_user_idx r:enrollment_requests i:enrollment_restrictions_open_idx i:enrollment_restrictions_pkey i:enrollment_restrictions_seq_key i:enrollment_sessions_credential_hash_key i:enrollment_sessions_pkey i:enrollment_sessions_provider_idx i:enrollment_sessions_user_idx r:enrollment_restrictions r:enrollment_sessions"
SIGNUP_OBJECTS="i:signup_sessions_challenge_hash_key i:signup_sessions_credential_hash_key i:signup_sessions_nonce_hash_key i:signup_sessions_pkey i:signup_sessions_proof_hash_key i:signup_sessions_ref_hash_key i:signup_sessions_user_idx r:signup_sessions"
# The 9 reviewed G020 label relations of 00020_labels.sql (relkind:name; labelRelations of
# core/internal/migrate/integration_test.go), owned by the owner role: the two tables, their primary keys, the
# definitions' scoped-name and owner keys, the assignments' unique key and two indexes.
LABEL_OBJECTS="i:label_assignments_group_idx i:label_assignments_pkey i:label_assignments_target_idx i:label_assignments_unique i:label_definitions_owner_key i:label_definitions_pkey i:label_definitions_scope_name_key r:label_assignments r:label_definitions"
# The 1 reviewed G057 relation of 00021_platform_labels.sql (relkind:name; platformLabelRelations of
# core/internal/migrate/integration_test.go), owned by the owner role: the partial unique platform-label index on
# app.organizations.
PLATFORM_LABEL_OBJECTS="i:organizations_platform_label_key"
# The 14 reviewed identity objects of 00002_identity.sql (relkind:name), all owned by the owner role.
IDENTITY_OBJECTS="i:identity_principals_pkey i:identity_principals_user_id_key i:identity_provider_binding_pkey i:login_transactions_pkey i:sessions_credential_hash_key i:sessions_pkey i:sessions_provider_logout_ref_hash_key i:sessions_user_live_idx i:users_pkey r:identity_principals r:identity_provider_binding r:login_transactions r:sessions r:users"
# The emulator's public client and issuer and its API base (identity.yml); never part of a delivered web body.
IDN_CLIENT=client_01IDENTITYTESTEMULATOR
IDN_ISSUER=https://workos-emulator.localhost/user_management/client_01IDENTITYTESTEMULATOR
IDN_BASE=http://emulator:4100

compose_identity() { # compose_identity DEST args...  (local.yml plus the identity overlay, as IDENTITY.md)
  local dest="$1" rc=0; shift
  res_compose_begin "$@" || return 125
  (cd "$dest" && docker compose --env-file .env -f infra/compose/local.yml -f infra/compose/identity.yml "$@") || rc=$?
  res_compose_end
  capture_after "$dest" "$@"
  return "$rc"
}

# identity_guards ROOT: one code per violated G004 guard rule; nothing when clean (B-IDENTITY-GUARDS).
identity_guards() {
  local root="$1" hits f
  hits="$(grep -rlE '"github.com/(workos/workos-go|golang-jwt/jwt|MicahParks/(keyfunc|jwkset))' --include='*.go' "$root/core" 2>/dev/null | grep -v "^$root/core/internal/identity/workos/" || true)"
  [[ -z "$hits" ]] || echo PROVIDER_IMPORT
  hits="$(grep -rl '"core/internal/identity/identitytest"' --include='*.go' "$root/core" 2>/dev/null | grep -v '_test\.go$' || true)"
  [[ -z "$hits" ]] || echo TEST_IMPORT
  hits="$( { grep -rlE 'IDENTITY_DEV|BypassLogin|bypass_login|_emulate|workos-emulator|sk_test_' --include='*.go' "$root/core" | grep -v '_test\.go$' | grep -v '/identitytest/'
    grep -rlE 'IDENTITY_DEV|[Bb]ypass|_emulate|workos-emulator|sk_test_' "$root/web/app" "$root/web/lib" "$root/web/components"; } 2>/dev/null || true)"
  [[ -z "$hits" ]] || echo TEST_SYMBOLS
  hits="$(grep -rlE 'cookies\(\)' "$root/web/app" "$root/web/lib" "$root/web/components" 2>/dev/null | grep -vxF "$root/web/lib/session.ts" || true)"
  [[ -z "$hits" ]] || echo COOKIES
  # G069 (Plan r1 §3.4): the appearance preference is next-themes' own (storageKey); no app file is allowed browser storage.
  hits="$(grep -rlE 'localStorage|sessionStorage|document\.cookie|indexedDB' "$root/web/app" "$root/web/lib" "$root/web/components" 2>/dev/null || true)"
  [[ -z "$hits" ]] || echo BROWSER_STORAGE
  ! grep -qE '"(next-auth|@auth/[^"]+|@workos-inc/[^"]+|@workos/[^"]+)"[[:space:]]*:' "$root/web/package.json" || echo AUTH_SDK
  ! grep -rqE 'NEXT_PUBLIC_IDENTITY' "$root/web/app" "$root/web/lib" "$root/web/components" "$root/web/next.config.ts" 2>/dev/null || echo PUBLIC_IDENTITY
  f="$root/web/lib/identity-client.ts"
  { grep -qxF 'import "server-only";' "$f" &&
    [[ "$(grep -cF 'const headers: Record<string, string> = { ...outboundHeaders(ctx) };' "$f")" == 1 ]] &&
    [[ "$(grep -oE 'headers(\.[A-Za-z]+|\["[^"]+"\]) =' "$f" | sed -E 's/ =$//; s/^headers\.?//; s/^\["//; s/"\]$//' | LC_ALL=C sort | tr '\n' ' ')" == "Authorization Content-Type X-Login-Binding " ]]; } || echo CLIENT_HEADERS
  { grep -qiE '^[[:space:]]+(org_id|organization_id|organization|roles?|permissions|entitlements|feature_flags|tenant_id|tenant)[[:space:]]' "$root/core/migrations/00002_identity.sql" ||
    grep -rlE 'json:"(org_id|organization_id|role|roles|permissions|entitlements|feature_flags)"' --include='*.go' "$root/core/internal/identity" 2>/dev/null | grep -qv '_test\.go$'; } && echo VENDOR_CLAIMS
  ! grep -qE 'IDENTITY_[A-Z_]*MFA' "$root/core/internal/config/config.go" "$root/.env.example" "$root"/infra/compose/*.yml 2>/dev/null || echo MFA_CONFIG
  # G014 C10: the MCP test trust anchor (MCP_CONNECT_JWKS_FILE) is refused outside APP_ENV=test by the config loader and
  # is read nowhere else in production code; only the identity test overlay sets it.
  { grep -qE '^[[:space:]]+if !test \{$' "$root/core/internal/config/mcp.go" &&
    grep -qxF '			add("MCP_CONNECT_JWKS_FILE is accepted only with APP_ENV=test")' "$root/core/internal/config/mcp.go" &&
    [[ -z "$( { grep -rnF 'MCP_CONNECT_JWKS_FILE' --include='*.go' "$root/core" | grep -vE '^[^:]+_test\.go:|^[^:]+:[0-9]+:[[:space:]]*//' | cut -d: -f1 | grep -vxF "$root/core/internal/config/mcp.go"
      grep -lF 'MCP_CONNECT_JWKS_FILE' "$root/.env.example" "$root"/infra/compose/*.yml | grep -vxF "$root/infra/compose/identity.yml"; } 2>/dev/null || true)" ]]; } || echo MCP_TEST_ANCHOR
  return 0
}

static_identity_guards() {
  local ok=1 codes base="$WORK/identity-guards" d got n=0
  codes="$(identity_guards "$REPO" | LC_ALL=C sort | tr '\n' ' ')"
  [[ -z "$codes" ]] || { ok=0; log "identity guards on the implementation: $codes"; }
  idfresh() { d="$base/$1"; mkdir -p "$d"; (cd "$REPO" && tar -cf - core/cmd core/internal core/migrations web/app web/lib web/components web/package.json web/next.config.ts .env.example infra/compose) | tar -xf - -C "$d"; }
  idcase() { got="$(identity_guards "$d" | LC_ALL=C sort | tr '\n' ' ')"; [[ "$got" == "$2 " ]] || { ok=0; log "identity guard control $1: got '$got', want '$2'"; }; n=$((n + 1)); rm -rf "$d"; }
  idfresh N1; printf '\nvar _ = "github.com/golang-jwt/jwt/v5"\n' >>"$d/core/internal/identity/service.go"; idcase N1 PROVIDER_IMPORT
  idfresh N2; printf 'package identity\n\nimport _ "core/internal/identity/identitytest"\n' >"$d/core/internal/identity/leak.go"; idcase N2 TEST_IMPORT
  idfresh N3; printf '\nconst devBypass = "IDENTITY_DEV_BYPASS"\n' >>"$d/core/internal/identity/service.go"; idcase N3 TEST_SYMBOLS
  idfresh N4; printf 'export const jar = cookies();\n' >"$d/web/lib/leak.ts"; idcase N4 COOKIES
  idfresh N5; printf 'export const keep = (v: string) => localStorage.setItem("sid", v);\n' >"$d/web/components/account/leak.tsx"; idcase N5 BROWSER_STORAGE
  idfresh N6; sed -i 's/"dependencies": {/"dependencies": {\n    "next-auth": "5.0.0",/' "$d/web/package.json"; idcase N6 AUTH_SDK
  idfresh N7; printf 'export const id = process.env.NEXT_PUBLIC_IDENTITY_CLIENT;\n' >"$d/web/lib/leak.ts"; idcase N7 PUBLIC_IDENTITY
  idfresh N8; sed -i 's|^  if (c.body !== undefined) headers\["Content-Type"\] = "application/json";$|&\n  headers.Cookie = "x";|' "$d/web/lib/identity-client.ts"; idcase N8 CLIENT_HEADERS
  idfresh N9; sed -i 's|^  display_name text NULL|  org_id       text NULL,\n&|' "$d/core/migrations/00002_identity.sql"; idcase N9 VENDOR_CLAIMS
  idfresh N10; printf 'IDENTITY_MFA_REQUIRED=true\n' >>"$d/.env.example"; idcase N10 MFA_CONFIG
  idfresh N11; printf '      MCP_CONNECT_JWKS_FILE: /keys/jwks.json\n' >>"$d/infra/compose/local.yml"; idcase N11 MCP_TEST_ANCHOR
  rm -rf "$base"
  [[ $ok == 1 && $n == 11 ]] && record B-IDENTITY-GUARDS PASS "the implementation violates no G004 guard (WorkOS SDK and token libraries only in core/internal/identity/workos; identitytest only from _test.go; no bypass, emulator or test-key symbol in production code; cookies() only in web/lib/session.ts; no credential in browser storage; no Next auth or WorkOS browser SDK; no NEXT_PUBLIC_IDENTITY; the identity client sends exactly outboundHeaders plus Authorization, X-Login-Binding and Content-Type; no organization, role or permission data persisted or decoded; no MFA configuration variable; the G014 MCP test trust anchor refused outside APP_ENV=test by the config loader and set only by the identity test overlay); each of the 11 seeded violations is reported with exactly its code" || record B-IDENTITY-GUARDS FAIL "identity guards"
}

# B-IDENTITY-CATALOG (migrations suite, at version 3 after the MIGRATIONS.md steps): exactly the 14 reviewed identity
# objects plus the 13 G005 tenancy relations, owned by the owner role, the baseline migration unchanged, and
# TestIdentityCatalog against real PostgreSQL (objects, owner, no RLS or tenant columns, exact runtime grants, permitted
# CRUD, denied DDL/DELETE/role use, CHECKs).
identity_catalog() { # identity_catalog DEST PROJECT
  local M="$1" project="$2" itenv="$WORK/identity-catalog.env" ok=1 got db log="$WORK/identity-catalog-it.log" passed failed
  db="$(env_get "$M" DATABASE_NAME)"
  (umask 077; printf 'DATABASE_NAME=%s\nDB_HOST=postgres\nDB_PORT=5432\nDB_MIGRATOR_PASSWORD=%s\nDB_RUNTIME_PASSWORD=%s\nDB_SUPERUSER=postgres\nDB_SUPERUSER_PASSWORD=%s\nAPP_ENV=test\n' \
    "$db" "$(env_get "$M" DB_MIGRATOR_PASSWORD)" "$(env_get "$M" DB_RUNTIME_PASSWORD)" "$(env_get "$M" POSTGRES_SUPERUSER_PASSWORD)" >"$itenv")
  compose "$M" --profile checks build core-check >"$WORK/identity-catalog-build.log" 2>&1 || ok=0
  got="$(psql_super "$M" "SELECT string_agg(c.relkind::text || ':' || c.relname, ' ' ORDER BY c.relkind::text COLLATE \"C\", c.relname COLLATE \"C\") || ' owners=' || string_agg(DISTINCT c.relowner::regrole::text, ',')
FROM pg_class c WHERE c.relnamespace = 'app'::regnamespace
  AND c.oid NOT IN ('app.goose_db_version'::regclass, pg_get_serial_sequence('app.goose_db_version', 'id')::regclass)
  AND NOT EXISTS (SELECT 1 FROM pg_index i WHERE i.indexrelid = c.oid AND i.indrelid = 'app.goose_db_version'::regclass)" 2>&1)" || ok=0
  local want; want="$(tr ' ' '\n' <<<"$IDENTITY_OBJECTS $TENANCY_OBJECTS $BUSINESS_PROFILE_OBJECTS $STORED_OBJECT_OBJECTS $MCP_OBJECTS $DOMAIN_OBJECTS $BRANDING_OBJECTS $MEMBERSHIP_OBJECTS $ENROLLMENT_OBJECTS $SIGNUP_OBJECTS $LABEL_OBJECTS $PLATFORM_LABEL_OBJECTS" | LC_ALL=C sort | paste -sd' ')"
  [[ "$got" == "$want owners=${db}_owner" ]] || { ok=0; log "identity objects: $got"; }
  [[ "$(git -C "$REPO" hash-object core/migrations/00001_baseline.sql)" == "$(git -C "$REPO" rev-parse "$BASE:core/migrations/00001_baseline.sql")" ]] || { ok=0; log "the baseline migration changed"; }
  it_fresh "$M" "$project" identity-catalog "$log" && go_it identity-catalog "$project" "$project-core-check" "$itenv" "$log" ./internal/identity/ '^TestIdentityCatalog$' || ok=0
  rm -f "$itenv"
  passed="$(grep -c -- '--- PASS: TestIdentityCatalog/' "$log" || true)"; failed="$(grep -cE -- '--- (FAIL|SKIP)' "$log" || true)"
  grep -q -- '^--- PASS: TestIdentityCatalog ' "$log" && [[ "$passed" -ge 1 && "$failed" == 0 ]] || { ok=0; tail -30 "$log"; }
  [[ $ok == 1 ]] && record B-IDENTITY-CATALOG PASS "at version 22 (row-level security on users, sessions, identity_principals and login_transactions; runtime CRUD only in its context shape; constraint controls through the attested setup connection): exactly the 14 reviewed identity objects (5 tables, 9 indexes) plus the 13 G005 tenancy relations, the 8 G006 BusinessProfile relations, the 3 G009 stored-object relations, the 4 G014 MCP connection relations, the 6 G016 domain relations, the 2 G017 branding relations, the 6 G019 invitation relations, the 15 G058 enrollment relations, the 8 G062 signup relations, the 9 G020 label relations and the 1 G057 platform-label index (89 in all), owned by ${db}_owner; 00001_baseline.sql byte-identical to the base; TestIdentityCatalog against real PostgreSQL: $passed/$passed subtests pass, none skipped" || record B-IDENTITY-CATALOG FAIL "identity catalog"
}

# --- identity project probes ------------------------------------------------------------------------------------------
idn_hp() { # idn_hp CID NAME HTTPPROBE-ARGS...: one probe in CID's network namespace -> $IDN_P/NAME.json
  local cid="$1" name="$2"; shift 2
  docker_run_owned idn-httpprobe --rm --label g001.run=$RUN_ID --network "container:$cid" -v "$IDN_IN:/in:ro" "$VT_IMG" httpprobe "$@" >"$IDN_P/$name.json" 2>>"$IDN_P/httpprobe.err" || true
}
idn_get() { yq_stdin -p json -o tsv "$2" <"$IDN_P/$1.json" 2>/dev/null || true; }
idn_file() { (umask 022; printf '%s' "$2" >"$IDN_IN/$1"); } # never in argv of a probe
idn_open() { chmod -R a+rX "$IDN_P" "$IDN_IN" 2>/dev/null || true; } # every probe input is the harness's own (handoffs are copied)
# idn_web_post NAME HEADERFILES... : a form POST to web as the browser would send it through the proxy.
pct_encode() { # pct_encode VALUE: VALUE percent-encoded once (every byte outside A-Z a-z 0-9 - . _ ~)
  local LC_ALL=C s="$1" i c out=""
  for ((i = 0; i < ${#s}; i++)); do
    c="${s:i:1}"
    if [[ "$c" =~ [A-Za-z0-9._~-] ]]; then out+="$c"; else printf -v c '%%%02X' "'$c"; out+="$c"; fi
  done
  printf '%s' "$out"
}
idn_session() { # idn_session CID NAME AUTHFILE ORIGIN -> status of core's GET /session?origin=ORIGIN with that credential
  # G016 A2: core answers a session only for the tenant origin it is bound to, so ORIGIN is mandatory (the caller takes
  # it from the probe handoff and validates it); it is URL-encoded exactly once here. Without it no probe is sent.
  [[ -n "${4:-}" ]] || { echo no-origin; return; }
  idn_hp "$1" "$2" --url "http://127.0.0.1:8180/internal/v1/auth/session?origin=$(pct_encode "$4")" --header-file "Authorization=/in/$3" --timeout 15s
  idn_get "$2" .status
}
# A web request through Node in web's own container, reading the redirect Location (never a secret in argv).
idn_web_location() { # idn_web_location WEB_CID METHOD PATH ORIGIN -> "STATUS LOCATION"
  docker exec "$1" node -e '
const [method, path, origin] = process.argv.slice(1);
const init = { method, redirect: "manual", headers: { origin } };
if (method === "POST") { init.headers["content-type"] = "application/x-www-form-urlencoded"; init.body = "return_to=%2Faccount"; }
fetch("http://127.0.0.1:3000" + path, init).then((r) => console.log(r.status, r.headers.get("location") || "-", r.headers.get("cache-control") || "-")).catch(() => console.log("0 - -"));' "$2" "$3" "$4" 2>/dev/null || echo "0 - -"
}

# On an unconfigured stack (variants_ops; local.yml only, like the root, rename and off variants): the account API is
# 401 and a sign-in start reports that sign-in is unavailable (folded into B-IDENTITY-PROBES).
identity_unconfigured_probes() { # identity_unconfigured_probes DEST
  local dest="$1" web out ok=1
  web="$(cid_of "$dest" web || true)"
  out="$(docker exec "$web" node -e 'fetch("http://127.0.0.1:3000/api/account").then(async (r) => console.log(r.status, r.headers.get("cache-control"), await r.text())).catch(() => console.log("0"))' 2>/dev/null || true)"
  [[ "$out" == '401 no-store {"error":"unauthenticated"}' ]] || { ok=0; log "unconfigured /api/account: $out"; }
  out="$(idn_web_location "$web" POST /auth/login http://127.0.0.1:3000)"
  [[ "$out" == "303 /sign-in?status=unavailable no-store" ]] || { ok=0; log "unconfigured sign-in start: $out"; }
  [[ $ok == 1 ]] && IDN_UNCONF=ok || IDN_UNCONF="unconfigured stack probes failed"
}

variants_identity() {
  local D="$WORK/idn" project="g001-idn-$RUN_ID" id
  if ports_busy "${PORTS_IDN[@]}"; then for id in IT E2E PROBES PRIVACY CONFIG; do record "B-IDENTITY-$id" FAIL "ports ${PORTS_IDN[*]} busy; refusing to re-map silently"; done; record B-MCP-E2E FAIL "ports ${PORTS_IDN[*]} busy"; return; fi
  ensure_tools || { for id in IT E2E PROBES PRIVACY CONFIG; do record "B-IDENTITY-$id" FAIL "validation tools unavailable"; done; record B-MCP-E2E FAIL "validation tools unavailable"; return; }
  copy_repo "$D"
  go_it_source "$D"; go_it_declare ops # G064 (Plan r2 §5.3)
  if ! configure "$D" "$project" "${PORTS_IDN[@]}" ||
     ! run_docs_block "$D" docs/foundation/IDENTITY.md identity:start "$WORK/idn-start.log" ||
     ! run_docs_block "$D" docs/foundation/IDENTITY.md identity:migrate "$WORK/idn-migrate.log"; then
    tail -40 "$WORK/idn-start.log" "$WORK/idn-migrate.log" 2>/dev/null
    for id in IT E2E PROBES PRIVACY CONFIG; do record "B-IDENTITY-$id" FAIL "identity environment did not start (IDENTITY.md identity:start/migrate)"; done
    record B-MCP-E2E FAIL "identity environment did not start (IDENTITY.md identity:start/migrate)"
    compose_identity "$D" down -v >/dev/null 2>&1 || true
    return
  fi
  capture_after "$D" up --build
  IDN_P="$WORK/idn-probes"; IDN_IN="$(obs_inputs)/identity"; mkdir -p "$IDN_P" "$IDN_IN"; chmod 700 "$IDN_IN"
  # The run's generated identity secrets (API key, passwords, canaries) join the registry; the file is deleted.
  compose_identity "$D" cp identity-setup:/identity/secrets.txt "$WORK/secrets-identity-setup" >/dev/null 2>&1 &&
    cat "$WORK/secrets-identity-setup" >>"$SECRETS_FILE" || log "identity setup secrets unavailable"
  rm -f "$WORK/secrets-identity-setup"
  if ! identity_it "$D" "$project"; then
    # G064 (Plan r2 §3.5, D-2 without fallback): a refusal in the isolation window is a reported blocker that stops the
    # overlay; nothing runs on a database whose isolation was not established.
    for id in E2E PROBES PRIVACY CONFIG; do record "B-IDENTITY-$id" FAIL "identity overlay stopped: the B-IDENTITY-IT isolation window refused"; done
    record B-MCP-E2E FAIL "identity overlay stopped: the B-IDENTITY-IT isolation window refused"
    compose_identity "$D" down -v >/dev/null 2>&1 || true
    return
  fi
  # G014: the MCP seed (the first writer on the fresh copy the G064 window leaves; no longer bound to the identity tests'
  # resets), the mcp-* browser projects inside B-IDENTITY-E2E, and B-MCP-E2E after the probes (its sign-out everywhere is
  # last for dave), before the privacy scans.
  mcp_seed "$D"
  identity_e2e "$D"
  identity_probes "$D"
  mcp_e2e "$D" "$project"
  identity_privacy "$D" "$project"
  identity_config "$D" "$project"
  # Logs of this project's own services only: the emulator prints its per-run API key, so its log is never saved.
  compose_identity "$D" logs --no-color core core-replica web proxy postgres redis minio intelligence >"$WORK/compose-logs-idn.txt" 2>&1 || true
  run_docs_block "$D" docs/foundation/IDENTITY.md identity:stop "$WORK/idn-stop.log" || log "identity:stop failed"
  rm -rf "$IDN_IN"
}

# ===================================================================================
# G014 identity-project MCP steps (Plan r8 §10.5): the MCP seed after B-IDENTITY-IT, the mcp-* browser projects inside
# B-IDENTITY-E2E, and B-MCP-E2E after B-IDENTITY-PROBES. dave's account, the two registry Organizations and their Groups
# come from infra/identity/mcp-seed.sql; the registry is infra/identity/mcp-hosts.tsv; the Connect test signer is the
# run's identitysetup key (C10, APP_ENV=test only).
MCP_SEED="" MCP_SUB=""
MCP_E2E_ISSUER=https://connect.identity.test
MCP_E2E_USER=0f14d000-0000-4000-8000-00000000000d
mcp_org() { awk -F'\t' -v h="$2" '$1 == h {print $2}' "$1/infra/identity/mcp-hosts.tsv"; } # mcp_org DEST HOST

# mcp_seed DEST: the seeded users' emulator IDs through the emulator's API (the run's API key only in a 0600 input), then
# the seed as the PostgreSQL superuser with the configured issuer, those IDs and the map's Organization UUIDs. G016
# D-2(a): every browser sign-in user (alice, bob, carol, dave, frank and the hostile user) gets an account and a
# membership; erin stays without one.
mcp_seed() {
  local dest="$1" core key orga orgb orgc db users name id
  local -A sub=()
  MCP_SEED="" MCP_SUB=""
  core="$(cid_of "$dest" core || true)"; key="$(grep -m1 '^sk_test_' "$SECRETS_FILE" || true)"
  [[ -n "$key" ]] || { MCP_SEED="the emulator API key is not in the registry"; return; }
  idn_file emulator-auth "Bearer $key"; idn_open
  idn_hp "$core" mcp-users --url "$IDN_BASE/user_management/users?limit=100" --header-file "Authorization=/in/emulator-auth" --timeout 15s
  rm -f "$IDN_IN/emulator-auth"
  users="$(idn_get mcp-users .body | yq_stdin -p json -o tsv '.data[] | [.email, .id]' 2>/dev/null || true)"
  rm -f "$IDN_P/mcp-users.json"
  while IFS=$'\t' read -r name id; do
    [[ "$id" =~ ^user_[A-Za-z0-9]{1,64}$ ]] || continue
    case "$name" in
      alice@identity.test | bob@identity.test | carol@identity.test | dave@identity.test | frank@identity.test) sub[${name%@*}]="$id" ;;
      canary-*@identity.test) sub[hostile]="$id" ;;
    esac
  done <<<"$users"
  for name in alice bob carol dave frank hostile; do
    [[ -n "${sub[$name]:-}" ]] || { MCP_SEED="the emulator user ID of $name is unavailable"; return; }
  done
  MCP_SUB="${sub[dave]}"
  orga="$(mcp_org "$dest" org-a.localhost:543)"; orgb="$(mcp_org "$dest" org-b.localhost:543)"; orgc="$(mcp_org "$dest" org-c.localhost:543)"
  db="$(env_get "$dest" DATABASE_NAME)"
  compose "$dest" exec -T postgres psql -X -q -v ON_ERROR_STOP=1 -U postgres -d "$db" -v issuer="$IDN_ISSUER" -v dave="$MCP_SUB" \
    -v alice="${sub[alice]}" -v bob="${sub[bob]}" -v carol="${sub[carol]}" -v frank="${sub[frank]}" -v hostile="${sub[hostile]}" \
    -v org_a="$orga" -v org_b="$orgb" -v org_c="$orgc" -f - <"$dest/infra/identity/mcp-seed.sql" >"$WORK/mcp-seed.log" 2>&1 || { MCP_SEED="the MCP seed failed"; tail -10 "$WORK/mcp-seed.log"; }
}

# B-MCP-E2E: the Go driver (tag mcpe2e) in the core check image in web's network namespace with the mcp-dark handoff
# (the later of the two browser projects), then the real core image's MCP configuration refusals (E2E-MCP-6). Minted
# tokens return on stdout (MCP-OBSERVED lines), join the registry and leave the log at once.
mcp_e2e() { # mcp_e2e DEST PROJECT
  local dest="$1" project="$2" ok=1 h="$IDN_IN/handoff-mcp-dark.json" web log="$WORK/mcp-e2e.log" n tokens refusals
  if [[ -n "$MCP_SEED" ]]; then record B-MCP-E2E FAIL "MCP seed: $MCP_SEED"; return; fi
  if [[ ! -f "$h" ]]; then record B-MCP-E2E FAIL "no handoff from the mcp-dark browser project"; return; fi
  web="$(cid_of "$dest" web || true)"; idn_open
  GO_IT_NO_EXPORT=1 GO_IT_TAG=mcpe2e GO_IT_NETWORK="container:$web" go_it mcp-e2e "$project" "$project-core-check" "" "$log" ./internal/mcp/ '^TestMCPE2E$' \
    -v "${project}_identity-run:/identity:ro" -v "$IDN_IN:/in:ro" \
    -e MCP_E2E_HANDOFF=/in/handoff-mcp-dark.json -e MCP_E2E_CA=/identity/ca.pem -e MCP_E2E_SIGNER=/identity/mcp/connect-signing.pem \
    -e MCP_E2E_ISSUER="$MCP_E2E_ISSUER" -e MCP_E2E_SUBJECT="$MCP_SUB" -e MCP_E2E_USER="$MCP_E2E_USER" \
    -e MCP_E2E_ORG_A="$(mcp_org "$dest" org-a.localhost:543)" -e MCP_E2E_ORG_B="$(mcp_org "$dest" org-b.localhost:543)" \
    -e MCP_E2E_EDGE=127.0.0.1:543 -e MCP_E2E_CORE=http://core:8180 || ok=0
  # When the converter is unavailable go_it_convert copies the raw go test -json stream into the converted logs (its
  # fail-closed diagnostic) and the plain-line handling below does not parse that form: register the minted values from
  # the raw stream first, then replace the converted logs and the converter's stderr by one bounded sanitized record,
  # so no raw JSON is printed (the tail below), retained or exported. The invocation stays failed.
  if grep -qE '^\{"' "$log" 2>/dev/null || grep -q 'findings=converter-unavailable' "$WORK/go-tests/mcp-e2e.tsv" 2>/dev/null; then
    cat -- "$log" "$log.events" 2>/dev/null | grep -ao 'MCP-OBSERVED [^" \\]*' | sed 's/^MCP-OBSERVED //' | sort -u >>"$SECRETS_FILE" || true
    printf 'mcp-e2e: the per-test converter was unavailable; the raw test stream is withheld (it carries the minted bearer values, now registered); no per-test evidence for this invocation\n' | tee -- "$log" >"$WORK/go-tests/mcp-e2e.log"
    [[ ! -e "$WORK/go-tests/mcp-e2e.stderr" ]] || printf 'withheld (converter diagnostics may echo the raw stream)\n' >"$WORK/go-tests/mcp-e2e.stderr"
    ok=0
  fi
  sed -n 's/^MCP-OBSERVED //p' "$log" >>"$SECRETS_FILE"
  tokens="$(grep -c '^MCP-OBSERVED ' "$log" || true)"
  sed -i '/^MCP-OBSERVED /d' "$log"
  sed -i 's/^MCP-OBSERVED .*/MCP-OBSERVED [withheld]/' "$WORK/go-tests/mcp-e2e.log" 2>/dev/null || true
  # The raw go test -json stream (LOG.events) still carries the minted values that are now registered; it is only the
  # converter's input and B-SECRET-CANARY scans $WORK, so it goes before anything is retained (the converted log stays).
  rm -f -- "$log.events"
  go_it_publish mcp-e2e || log "G064 go_it mcp-e2e: export failed"
  n="$(it_count "$log" TestMCPE2E)"
  [[ "$n" == 6 && "$tokens" -ge 12 ]] || { ok=0; tail -60 "$log"; }
  refusals="$(mcp_config_refusals "$dest" "$project")"
  [[ -z "$refusals" ]] || { ok=0; log "E2E-MCP-6: $refusals"; }
  [[ $ok == 1 ]] && record B-MCP-E2E PASS "through the TLS edge of the disposable identity environment (real core, replica, web and proxy; the run CA; the per-run Connect test signer): TestMCPE2E 6/6 (E2E-MCP-1 discovery header and metadata, activation by the official SDK client, exactly the five tools, all five called, logout terminal on a connection created through web's setup route; E2E-MCP-2 a second family on one URL shares the selection, a second connection on A and one on B keep isolated stored state; E2E-MCP-3 cross-host, cross-connection, other-client, other-issuer, expired and tampered tokens 401; E2E-MCP-4 the edge serves only the MCP paths on tenant hosts (internal, health and other paths 404 from web; none on the attack origin); E2E-MCP-5 web revoke and sign-out everywhere terminal on both hosts, signing out one browser session leaves connections; E2E-MCP-7 $tokens minted tokens registered for the scans); E2E-MCP-6 real core image, no network: the test trust anchor with APP_ENV=production, an http issuer, the retired host registry (MCP_TENANT_HOSTS_FILE) and the fixture TLS adapter with APP_ENV=production each exit 1 with exactly one core.exit config_invalid line, a valid production MCP configuration (no registry) starts" || record B-MCP-E2E FAIL "MCP end to end ($n/6 subtests, $tokens tokens; refusals: ${refusals:-ok})"
}

# mcp_config_refusals DEST PROJECT: E2E-MCP-6 on the real core image with no network. The retired registry and key-set
# files are copied into each created container (no host directory is mounted). Prints the failed parts. G016: the host
# registry (MCP_TENANT_HOSTS_FILE) is retired and refused; tenant hosts resolve through slugs and verified domains.
mcp_config_refusals() {
  local dest="$1" project="$2" in="$IDN_IN/mcp-config" D="$WORK/mcp-config" c cid rc i out="" cases=""
  local img="$project-core"
  mkdir -p "$in/files" "$D"; chmod 777 "$D"
  core_env "$in/base.env" ":8180" "Probe Product"
  sed -i -e 's/^APP_ENV=test$/APP_ENV=production/' -e 's#^OBJECT_STORE_ENDPOINT=http://#OBJECT_STORE_ENDPOINT=https://#' "$in/base.env"
  local full canary
  canary="$(new_canary)"
  full=$'IDENTITY_PROVIDER=workos\nIDENTITY_WORKOS_CLIENT_ID=client_01PRODUCTIONPROBE\nIDENTITY_WORKOS_ISSUER=https://api.workos.com/user_management/client_01PRODUCTIONPROBE\nIDENTITY_WORKOS_API_KEY=sk_live_'"$canary"$'\nIDENTITY_PUBLIC_ORIGINS=https://app.probe.example.net\nIDENTITY_COOKIE_PROFILE=https'
  local mcp=$'MCP_CONNECT_ISSUER=https://connect.probe.example.net'
  printf 'app.probe.example.net\t0f14e000-0000-4000-8000-000000000001\n' >"$in/files/hosts.tsv"
  printf '{"keys":[]}\n' >"$in/files/jwks.json"
  chmod 644 "$in/files"/*
  # name|extra environment lines: each must exit 1 with one core.exit config_invalid record.
  local -a refuse=(
    "m1-anchor-production|$mcp"$'\nMCP_CONNECT_JWKS_FILE=/mcp/jwks.json'
    "m2-http-issuer|${mcp/https:\/\/connect/http://connect}"
    "m3-retired-registry|$mcp"$'\nMCP_TENANT_HOSTS_FILE=/mcp/hosts.tsv'
    "m4-production-fixture|$mcp"$'\nDOMAIN_TLS_ADAPTER=fixture')
  mcp_run() { # mcp_run NAME ENVFILE DETACH: a created container with the files copied in
    cid="$(docker_create_owned cc-env --label g001.run=$RUN_ID --network none --env-file "$2" "$img")" || return 1
    docker cp "$in/files/." "$cid:/mcp/" >/dev/null 2>&1 || { docker rm -f "$cid" >/dev/null 2>&1; return 1; }
  }
  for c in "${refuse[@]}"; do
    { cat "$in/base.env"; printf '%s\n%s\n' "$full" "${c#*|}"; } >"$in/${c%%|*}.env"
    rc=255
    if mcp_run "${c%%|*}" "$in/${c%%|*}.env"; then
      rc=0; timeout 60 docker start -a "$cid" >"$D/${c%%|*}.out" 2>&1 || rc=$?
      docker rm -f "$cid" >/dev/null 2>&1 || true
    fi
    echo "$rc" >"$D/${c%%|*}.out.rc"
    cases+=" --case /d/${c%%|*}.out=1:core.exit:config_invalid --exactly-one-line /d/${c%%|*}.out"
  done
  # A valid production MCP configuration starts and stops cleanly (the key set is fetched only on first use).
  { cat "$in/base.env"; printf '%s\n%s\n' "$full" "$mcp"; } >"$in/m5-valid.env"
  if mcp_run m5 "$in/m5-valid.env"; then
    docker start "$cid" >/dev/null 2>&1 || out+="m5-start "
    for ((i = 0; i < 60; i++)); do docker logs "$cid" 2>&1 | grep -q '"event":"core listening"' && break; sleep 1; done
    docker stop "$cid" >/dev/null 2>&1 || true
    rc="$(docker wait "$cid" 2>/dev/null || echo 255)"; docker logs "$cid" >"$D/m5-valid.out" 2>&1 || true; echo "$rc" >"$D/m5-valid.out.rc"
    docker rm -f "$cid" >/dev/null 2>&1 || true
  else out+="m5-create "; fi
  sed -n 's/^IDENTITY_WORKOS_API_KEY=sk_live_//p' "$in/m1-anchor-production.env" >"$D/canaries.txt"
  # shellcheck disable=SC2086 # $cases is a list of separate logcheck arguments
  vtools -v "$D:/d:ro" "$VT_IMG" logcheck startup --service core $cases --case "/d/m5-valid.out=0:core listening:port=8180+app_env=production" --canaries /d/canaries.txt >"$WORK/mcp-config-core.log" 2>&1 || { out+="startup "; tail -20 "$WORK/mcp-config-core.log" >&2; }
  rm -rf "$D" "$in"
  printf '%s' "$out"
}

# B-IDENTITY-IT: every identity integration test (L1 against real PostgreSQL) in the core check image on the identity
# project's data network, before the browser runs (G064 retired the former dependency on the tests' identity resets).
# G064 (Plan r2 §3.5; D-2 approved without fallback): the identity tests run inside an isolation window. (1) core and
# core-replica, the only database clients, are stopped and zero client sessions on <db> are asserted; (2) the overlay's
# own source-bound pristine check (its own read-only `migrate status` of the bound migrate image -> db_pristine_verify,
# run by db_template_create, which then copies <db> into the run-private template; no migrations-suite row is required
# or consulted); (3) the identity tests on the canonical <db>; (4) db_fresh recreates <db> from the pristine template;
# (5) core and core-replica start again (no dependency is touched). mcp_seed is then the first writer. A refusal in the
# window records B-IDENTITY-IT FAIL with the cause and returns 1 (the overlay stops); it never falls back to running the
# tests on the shared database. A failing test with the window intact is an ordinary B-IDENTITY-IT FAIL (status 0).
identity_it() { # identity_it DEST PROJECT -> 1 when the isolation window refused
  local dest="$1" project="$2" env="$IDN_IN/it.env" log="$WORK/identity-it.log" passed db s="" i built=1 win="" stopped=0 rec facts
  local irc=0 pkgok=1 pkgs=""
  local slog="$WORK/idn-template-status.log" wlog="$WORK/idn-window.log"
  db="$(env_get "$dest" DATABASE_NAME)"
  printf 'DATABASE_NAME=%s\nDB_HOST=postgres\nDB_PORT=5432\nDB_MIGRATOR_PASSWORD=%s\nDB_RUNTIME_PASSWORD=%s\nDB_SUPERUSER=postgres\nDB_SUPERUSER_PASSWORD=%s\nAPP_ENV=test\n' \
    "$db" "$(env_get "$dest" DB_MIGRATOR_PASSWORD)" "$(env_get "$dest" DB_RUNTIME_PASSWORD)" "$(env_get "$dest" POSTGRES_SUPERUSER_PASSWORD)" >"$env"
  chmod 600 "$env"
  : >"$log"
  compose "$dest" --profile checks build core-check >"$WORK/idn-check-build.log" 2>&1 || built=0
  # (1) Stop the only database clients; bounded drain (polling only), then zero client sessions or a refusal.
  if compose_identity "$dest" stop core core-replica >"$wlog" 2>&1; then
    stopped=1
    for ((i = 0; i < IT_DRAIN_POLLS; i++)); do
      s="$(_db_sessions "$dest" "$db" 2>&1)" || s="?"$'\t'"$(_db_oneline "$s")"
      [[ "${s%%$'\t'*}" == 0 ]] && break
      sleep 0.2
    done
    [[ "${s%%$'\t'*}" == 0 ]] || win="client session(s) on $db with core and core-replica stopped: $(_db_oneline "$s")"
  else win="compose stop core core-replica failed: $(tail -3 "$wlog" | tr '\n' ' ')"; fi
  # (2) The overlay's own pristine check and template.
  if [[ -z "$win" ]]; then
    compose_migrate "$dest" run --rm migrate status >"$slog" 2>&1 || log "identity overlay: migrate status for the template failed (see $slog)"
    DB_LC_CALLER=identity-overlay db_template_create "$dest" "$project" "$dest" "$slog" || win="$DB_LC_ERR"
  fi
  # (3) The identity tests on the canonical <db>, (4) its recreation from the pristine template.
  if [[ -z "$win" ]]; then
    if ((built)); then
      # G064 (Plan r2 §5.2): every package the pattern selects is expected per (package, test) by `-list .` (the
      # same selection as no -run) and must reach its terminal event.
      go_it identity-it "$project" "$project-core-check" "$env" "$log" ./internal/identity/... . || irc=$?
    else irc=1; printf 'core check image build failed (idn-check-build.log); the identity tests did not run\n' >"$log"; fi
    it_drain "$dest" identity-overlay identity-it-after
    DB_LC_CALLER=identity-overlay db_fresh "$dest" "$project" identity-it-after || win="$DB_LC_ERR"
  fi
  rm -f "$env"
  # (5) core and core-replica start again on the fresh copy.
  if [[ -z "$win" ]] && ((stopped)); then
    compose_identity "$dest" up -d --wait --no-deps --no-build core core-replica >>"$wlog" 2>&1 ||
      win="core and core-replica did not start again: $(tail -3 "$wlog" | tr '\n' ' ')"
  fi
  if [[ -n "$win" ]]; then
    log "identity isolation window refused: $win"
    record B-IDENTITY-IT FAIL "isolation window refused (G064 D-2, no fallback; db-lifecycle.tsv, idn-window.log): $win"
    return 1
  fi
  rec="$WORK/db-lifecycle/${db}_tpl.rec"; facts="MIG_TOP=$(_db_rec "$rec" mig_top), $(_db_rec "$rec" baseline | grep -c .) baseline tables empty, catalog fingerprint $(_db_rec "$rec" fingerprint_catalog | cut -c1-12)"
  passed="$(grep -cE -- '^--- PASS: ' "$log" || true)"
  pkgs="$(identity_it_packages "$WORK/go-tests/identity-it.tsv")" || pkgok=0
  if [[ $irc == 0 && $pkgok == 1 ]] && grep -qE '^ok[[:space:]]+core/internal/identity[[:space:]]' "$log" && ! grep -qE -- '--- (FAIL|SKIP)|^FAIL' "$log" && [[ "$passed" -ge 10 ]]; then
    record B-IDENTITY-IT PASS "identity integration tests against real PostgreSQL (tag integration; core check image on the identity project's data network): $passed top-level tests pass, none skipped (catalog, mapping concurrency, atomic consumption, lifetimes and renewal, replicas, outages, disable and logout-all races, namespace binding, continuity, sign-out, log privacy); G064 isolation window: core and core-replica stopped with 0 client sessions, the overlay's own migrate status and pristine check ($facts; pristine row in db-lifecycle.tsv) and template ${db}_tpl, then $db recreated from it and core restarted before mcp_seed; G064 per-package expectations (go test -list, passed/expected): $pkgs"
  else tail -40 "$log"; record B-IDENTITY-IT FAIL "identity integration tests (go test status $irc; per package: ${pkgs:-none})"; fi
}

# G064 (Plan r2 §5.2): identity_it_packages ROWS: every selected package of B-IDENTITY-IT terminated and passed every
# expected test (derived by -list, not pinned): a test package needs terminal pass and passed = expected >= 1, a package
# without test files its skip and 0 expected; prints "package passed/expected" per package; status 1 otherwise.
identity_it_packages() {
  [[ -s "$1" ]] || { printf 'no per-test rows'; return 1; }
  awk -F'\t' '$1 == "package" { n++; ok = ($5 == "pass" && $7 >= 1 && $8 == $7 && $9 == 0 && $10 == 0) || ($5 == "skip-no-test-files" && $7 == 0)
      if (!ok) bad = 1; out = out (out == "" ? "" : ", ") $3 " " ($5 == "skip-no-test-files" ? "no test files" : $8 "/" $7 (ok ? "" : " " $5)) }
    END { printf "%s", out; exit (n == 0 || bad) }' "$1"
}

# G016 (Plan r4 §6.2): E-DOM-1..5 in the two domain projects of identity.spec.ts (the third platform host org-c and the
# custom host portal.acme.test), required beside the inherited matrix. G057 (Plan r5 §5.2 T8): E-DOM-6, the friendly
# slug's host routes nowhere (the fixed 404, no Location), in domain-platform.
BROWSER_MATRIX_DOMAINS=("domain-platform|E-DOM-1 E-DOM-3 E-DOM-4 E-DOM-6 S-ERRORS" "domain-custom|E-DOM-2 E-DOM-5 S-ERRORS")
# B-IDENTITY-E2E: IDENTITY.md identity:browser with an execution-owned one-off and internal reports copied out. The observed bearer values
# join the registry and leave the report; the probe handoff moves into the 0700 probe inputs.
identity_e2e() { # identity_e2e DEST
  local dest="$1" rep="$WORK/browser-identity" line ok=1 f runc
  runc="g001-browser-$RUN_ID-$(random_hex 4)"
  mkdir -p "$rep"
  line="$(docs_block "$dest/docs/foundation/IDENTITY.md" identity:browser)"
  if [[ "$(grep -c . <<<"$line")" != 1 || "$line" != *" run --rm "* ]]; then record B-IDENTITY-E2E FAIL "IDENTITY.md identity:browser must be one docker compose run --rm line"; return; fi
  # G051 L6 is not applied here (cycle 2): identity-light and identity-dark share the seeded users and sessions, so at
  # 2 workers their session-rotation, cross-tab and sign-out-everywhere scenarios interfere. The run keeps the
  # configuration default of 1 worker.
  line="${line/ run --rm / run --name $runc --label g001.run=$RUN_ID -e PLAYWRIGHT_JSON_OUTPUT_FILE=/tmp/report/results.json -e PLAYWRIGHT_OUTPUT_DIR=/tmp/report/out }"
  local idn_project; idn_project="$(env_get "$dest" COMPOSE_PROJECT_NAME)" || idn_project=""
  ledger_watch_begin "$dest" "$idn_project" identity-e2e
  (cd "$dest" && bash -euo pipefail -c "$line") >"$WORK/idn-browser.log" 2>&1 || { ok=0; tail -40 "$WORK/idn-browser.log"; }
  ledger_watch_end "$dest" "$idn_project" identity-e2e "$WORK/idn-browser.log" "$line"
  capture_after "$dest" run --build
  browser_collect_report "$runc" "$rep" "$WORK/idn-browser.log" || ok=0
  if [[ -s "$rep/secrets-observed.txt" ]]; then cat "$rep/secrets-observed.txt" >>"$SECRETS_FILE"; else ok=0; log "the identity browser run observed no bearer values"; fi
  rm -f "$rep/secrets-observed.txt"
  for f in "$rep"/handoff-identity-*.json; do
    [[ -f "$f" ]] || continue
    yq_stdin -p json -o tsv '[.live, .csrf, .revoked] | .[]' <"$f" >>"$SECRETS_FILE"
    # Keep the probe inputs private even though copied reports now have host ownership.
    (umask 077; cat "$f" >"$IDN_IN/$(basename "$f")") && rm -f "$f"
  done
  browser_matrix "$rep/results.json" "$WORK/browser-identity-tests.tsv" "${BROWSER_MATRIX_IDENTITY_E2E[@]}" "${BROWSER_MATRIX_DOMAINS[@]}" || ok=0
  # G014: the mcp-* projects' handoffs (dave's sessions, CSRF tokens and created URLs) for B-MCP-E2E, the same way.
  for f in "$rep"/handoff-mcp-*.json; do
    [[ -f "$f" ]] || continue
    yq_stdin -p json -o tsv '[.live_a, .csrf_a, .live_b, .csrf_b, .spare_a, .spare_csrf_a] | .[]' <"$f" >>"$SECRETS_FILE"
    (umask 077; cat "$f" >"$IDN_IN/$(basename "$f")") && rm -f "$f"
  done
  [[ -f "$IDN_IN/handoff-identity-light.json" ]] || { ok=0; log "no probe handoff from identity-light"; }
  [[ -f "$IDN_IN/handoff-mcp-dark.json" ]] || { ok=0; log "no MCP handoff from mcp-dark"; }
  [[ $ok == 1 ]] && record B-IDENTITY-E2E PASS "IDENTITY.md identity:browser on the disposable HTTPS environment (emulator, TLS proxy, core and replica): $(grep -c . "$WORK/browser-identity-tests.tsv") tests in identity-light, identity-dark, domain-platform, domain-custom, mcp-light, mcp-dark, onboarding-light, onboarding-dark, membership-light, membership-dark, enrollment-light, enrollment-dark, signup-light, signup-dark, labels-light and labels-dark, every required L2 scenario passes in both identity projects, every G016 hostname scenario (E-DOM-1 third platform host, E-DOM-2 custom host through the shared relay, E-DOM-3 unknown hosts, E-DOM-4 public entry serving the G062 landing, E-DOM-5 removed host, G057 E-DOM-6 the friendly slug's host routes nowhere) in its domain project and every G014 MCP setup scenario (M-NOAUTH, M-NONMEMBER, M-SETUP, M-LIST with the seeded old-host row, M-A11Y, M-HANDOFF; G063 M-RENAME and M-ACCOUNT) in both mcp projects (mcp-light and mcp-dark) and every G018 onboarding scenario (E-ONB-1..7 with the member-page S-LANDMARKS and S-NAV) in both onboarding projects and every G019 membership scenario (E-MEM-1..7 with the members-page S-LANDMARKS and S-NAV) in both membership projects and every G058 enrollment scenario (E-ENR-1..5: open signup, Organization invitation entry, denial and closed-after-start, request approve/revoke/admit with the wrong-authority refusals, keyboard and axe) in both enrollment projects and every G062 root-host signup scenario (E-SUP-1..6: root and www landing, fresh and existing identity into a new tenant, field errors, the automatic continuation with its captured-secret, recovery and no-re-mint-on-GET cases, first use without onboarding) in both signup projects, every G020 label scenario (E-LAB-1..6: Organization label lifecycle by keyboard, Organization labels on a Group page, a Group's own labels and their visibility, narrowed visibility and a removed role on the next load, forged writes refused, axe at 1280 and 320) with S-LANDMARKS and S-NAV in both labels projects, 0 skipped/flaky/unexpected; every page request on the app origins and the emulator (E-NET-1)" || record B-IDENTITY-E2E FAIL "identity browser scenarios"
}

# B-IDENTITY-PROBES (with E-LO-5 and E-CSRF's core-restart case): driver probes with the browser run's handoff.
identity_probes() { # identity_probes DEST
  local dest="$1" ok=1 h="$IDN_IN/handoff-identity-light.json" core web replica pg st i out body origin
  if [[ ! -f "$h" ]]; then record B-IDENTITY-PROBES FAIL "no handoff from the identity browser run"; return; fi
  # G016 A2: the handoff's sessions are bound to the tenant origin they signed in on. The session probes send exactly
  # that origin, taken from the handoff; it must be the identity fixture's Organization A origin (identity.spec.ts
  # ORIGIN_A, the Host/Origin of the web probes below). Missing or anything else fails the row; there is no fallback.
  origin="$(yq_stdin -p json -o tsv '.origin' <"$h" 2>/dev/null || true)"
  if [[ "$origin" != "https://org-a.localhost:543" ]]; then record B-IDENTITY-PROBES FAIL "the handoff's session origin is missing or not the fixture's Organization A origin"; return; fi
  idn_file auth-live "Session $(yq_stdin -p json -o tsv '.live' <"$h")"
  idn_file auth-revoked "Session $(yq_stdin -p json -o tsv '.revoked' <"$h")"
  idn_file cookie-live "__Host-sid=$(yq_stdin -p json -o tsv '.live' <"$h")"
  idn_file body-csrf "csrf_token=$(yq_stdin -p json -o tsv '.csrf' <"$h")"
  idn_file cookie-logout "__Host-logout=$(head -c 32 /dev/urandom | base64 | tr '+/' '-_' | tr -d '=\n')"
  idn_open
  core="$(cid_of "$dest" core || true)"; web="$(cid_of "$dest" web || true)"; replica="$(compose_identity "$dest" ps -q core-replica | head -1 || true)"; pg="$(cid_of "$dest" postgres || true)"
  local -a webpost=(--method POST --header "Host: org-a.localhost:543" --header "Origin: https://org-a.localhost:543" --header "Content-Type: application/x-www-form-urlencoded" --header-file "Cookie=/in/cookie-live" --body-file /in/body-csrf --timeout 20s)
  # Live and revoked sessions on core and the replica.
  [[ "$(idn_session "$core" p1-live auth-live "$origin")" == 200 && "$(idn_session "$core" p1-revoked auth-revoked "$origin")" == 401 && "$(idn_session "$replica" p1-replica auth-live "$origin")" == 200 ]] || { ok=0; log "live/revoked before restart"; }
  # Restart: a valid session survives; a revoked one stays denied.
  compose_identity "$dest" restart core >/dev/null 2>&1
  for ((i = 0; i < 60; i++)); do idn_hp "$core" ready --url http://127.0.0.1:8180/readyz --timeout 3s; [[ "$(idn_get ready .status)" == 200 ]] && break; sleep 1; done
  [[ "$(idn_session "$core" p2-live auth-live "$origin")" == 200 && "$(idn_session "$core" p2-revoked auth-revoked "$origin")" == 401 ]] || { ok=0; log "sessions across a core restart"; }
  # Redis stopped: authentication unaffected.
  compose_identity "$dest" stop redis >/dev/null 2>&1
  [[ "$(idn_session "$core" p3-redis auth-live "$origin")" == 200 ]] || { ok=0; log "session with Redis stopped"; }
  compose_identity "$dest" start redis >/dev/null 2>&1
  # E-LO-5: sign-out while PostgreSQL is paused fails visibly (503, never signed_out) and changes nothing.
  docker pause "$pg" >/dev/null 2>&1
  idn_hp "$web" p4-logout --url http://127.0.0.1:3000/auth/logout "${webpost[@]}"
  docker unpause "$pg" >/dev/null 2>&1
  st="$(idn_get p4-logout .status)"; body="$(idn_get p4-logout .body)"
  [[ "$st" == 503 && "$body" == *"Sign-out failed"* && "$(idn_get p4-logout .cache_control)" == no-store ]] || { ok=0; log "sign-out with PostgreSQL paused: $st"; }
  # C7: the provider sign-out's return route with the store unavailable answers 503 and claims no outcome.
  docker pause "$pg" >/dev/null 2>&1
  idn_hp "$web" p4-return --url http://127.0.0.1:3000/auth/logout-return --header "Host: org-a.localhost:543" --header-file "Cookie=/in/cookie-logout" --timeout 20s
  docker unpause "$pg" >/dev/null 2>&1
  st="$(idn_get p4-return .status)"; body="$(idn_get p4-return .body)"
  [[ "$st" == 503 && "$body" == *"Sign-out status unavailable"* && "$body" != *"WorkOS session"* && "$(idn_get p4-return .cache_control)" == no-store ]] || { ok=0; log "logout return with PostgreSQL paused: $st"; }
  for ((i = 0; i < 60; i++)); do idn_hp "$core" ready --url http://127.0.0.1:8180/readyz --timeout 3s; [[ "$(idn_get ready .status)" == 200 ]] && break; sleep 1; done
  [[ "$(idn_session "$core" p4-after auth-live "$origin")" == 200 ]] || { ok=0; log "the session after a failed sign-out"; }
  # PostgreSQL stopped: a bounded 503 and no success, then recovery.
  compose_identity "$dest" stop postgres >/dev/null 2>&1
  st="$(idn_session "$core" p5-down auth-live "$origin")"
  [[ "$st" == 503 && "$(idn_get p5-down .duration_ms)" -le 10000 && "$(idn_get p5-down .body)" == '{"error":"identity_unavailable"}' ]] || { ok=0; log "session with PostgreSQL stopped: $st"; }
  compose_identity "$dest" start postgres >/dev/null 2>&1
  for ((i = 0; i < 90; i++)); do idn_hp "$core" ready --url http://127.0.0.1:8180/readyz --timeout 3s; [[ "$(idn_get ready .status)" == 200 ]] && break; sleep 1; done
  [[ "$(idn_session "$core" p5-up auth-live "$origin")" == 200 ]] || { ok=0; log "session after PostgreSQL recovered"; }
  # Forged identity headers on direct core and web requests are ignored.
  idn_hp "$core" p6-core --url http://127.0.0.1:8180/internal/v1/auth/session --header "X-User-Id: 00000000-0000-0000-0000-000000000000" --header "X-MFA-Verified: true" --header "X-Forwarded-User: admin" --timeout 10s
  idn_hp "$web" p6-web --url http://127.0.0.1:3000/api/account --header "Host: org-a.localhost:543" --header "X-User-Id: 00000000-0000-0000-0000-000000000000" --timeout 10s
  [[ "$(idn_get p6-core .status)" == 401 && "$(idn_get p6-web .status)" == 401 ]] || { ok=0; log "forged identity headers"; }
  # E-CSRF (core restart): the token rendered before the restart still signs out everywhere.
  idn_hp "$web" p7-csrf --url http://127.0.0.1:3000/auth/logout-all "${webpost[@]}"
  [[ "$(idn_get p7-csrf .status)" == 303 && "$(idn_get p7-csrf .cache_control)" == no-store && "$(idn_session "$core" p7-after auth-live "$origin")" == 401 ]] || { ok=0; log "CSRF token across a core restart"; }
  [[ "$IDN_UNCONF" == ok ]] || { ok=0; log "unconfigured stack: ${IDN_UNCONF:-not run (the ops variant did not run)}"; }
  rm -f "$IDN_IN"/auth-* "$IDN_IN"/cookie-* "$IDN_IN"/body-*
  # A live session's answer carries its CSRF token (a registered value): the probe outputs leave the work tree once
  # evaluated, so B-SECRET-CANARY keeps its meaning for everything else.
  rm -f "$IDN_P"/p[0-9]*.json
  [[ $ok == 1 ]] && record B-IDENTITY-PROBES PASS "driver probes with the browser run's handoff (credentials only in 0600 files; every session probe sends the handoff's validated tenant origin, G016 A2): live session on core and the replica, revoked denied; both kept across a core restart; Redis stopped: no effect; PostgreSQL paused: sign-out 503 'Sign-out failed' (no-store, never signed_out) and the session unchanged after recovery (E-LO-5), and the logout return route 503 'Sign-out status unavailable' with no provider claim (C7); PostgreSQL stopped: bounded 503 identity_unavailable, then recovery; forged X-User-Id/MFA headers ignored by core and web; the CSRF token rendered before the restart signs out everywhere (303, no-store); unconfigured stack: /api/account 401 and sign-in start 303 /sign-in?status=unavailable" || record B-IDENTITY-PROBES FAIL "identity probes"
}

# B-IDENTITY-PRIVACY (Plan r5 §8.3): catalogued logs; no registered value in core, web or proxy logs or the identity
# tables; email never stored; P-NEG-1..3 detected and P-POS-1/2 clean, by the existing scanners only.
identity_privacy() { # identity_privacy DEST PROJECT
  local dest="$1" project="$2" ok=1 svc c pv="$WORK/idn-privacy" reg dump planted out
  mkdir -p "$pv"; chmod 777 "$pv"
  for svc in core core-replica web proxy; do c="$(compose_identity "$dest" ps -q "$svc" | head -1 || true)"; docker logs "$c" >"$pv/$svc.log" 2>&1 || true; done
  vtools -v "$pv:/o:ro" "$VT_IMG" logcheck validate --service core --file /o/core.log --file /o/core-replica.log >"$WORK/idn-logcheck-core.log" 2>&1 || { ok=0; tail -10 "$WORK/idn-logcheck-core.log"; }
  vtools -v "$pv:/o:ro" "$VT_IMG" logcheck validate --service web --file /o/web.log >"$WORK/idn-logcheck-web.log" 2>&1 || { ok=0; tail -10 "$WORK/idn-logcheck-web.log"; }
  # The registry so far (this run's secrets, the identity secrets and every observed bearer value) as the canary list.
  reg="$IDN_IN/canaries.txt"; sort -u "$SECRETS_FILE" | awk 'length($0) >= 8' >"$reg"; chmod 644 "$reg"
  dump="$pv/identity-tables.txt"
  psql_super "$dest" "SELECT row_to_json(t)::text FROM (SELECT * FROM app.users) t UNION ALL SELECT row_to_json(t)::text FROM (SELECT issuer, subject, user_id FROM app.identity_principals) t UNION ALL SELECT row_to_json(t)::text FROM (SELECT * FROM app.identity_provider_binding) t" >"$dump" 2>&1 || ok=0
  grep -q '@identity\.test' "$dump" && { ok=0; log "an email address reached the identity tables"; }
  # The dump holds the hostile user's registered display canary as that user's display name (permitted): it is not
  # evidence and leaves the work tree at once, so B-SECRET-CANARY keeps its meaning for everything else.
  rm -f "$dump"
  if vtools -v "$pv:/o:ro" -v "$IDN_IN:/i:ro" "$VT_IMG" logcheck scan --canaries /i/canaries.txt /o/core.log /o/core-replica.log /o/web.log /o/proxy.log >"$WORK/idn-scan.log" 2>&1; then :; else ok=0; log "a registered value appears in a service log: $(head -5 "$WORK/idn-scan.log" | tr '\n' ' ')"; fi
  # The display canary is stored as the owner's display name (permitted) but nowhere else.
  # P-POS-1: the sanitized authorize-request evidence names the client and emulator host and passes the same scan.
  yq_stdin -p json -o tsv '[.. | select(tag == "!!map" and has("name") and .name == "authorize-request.txt") | .body] | .[0]' <"$WORK/browser-identity/results.json" 2>/dev/null | base64 -d >"$pv/authorize-request.txt" 2>/dev/null || true
  grep -qF "client_id=$IDN_CLIENT" "$pv/authorize-request.txt" && grep -qF "$IDN_BASE/user_management/authorize?" "$pv/authorize-request.txt" || { ok=0; log "P-POS-1: no sanitized authorize evidence naming the client and the provider host"; }
  chmod a+r "$pv/authorize-request.txt" 2>/dev/null || true
  vtools -v "$pv:/o:ro" -v "$IDN_IN:/i:ro" "$VT_IMG" logcheck scan --canaries /i/canaries.txt /o/authorize-request.txt >/dev/null 2>&1 || { ok=0; log "P-POS-1: a registered value in the authorize evidence"; }
  vtools -v "$WORK/browser-identity:/b:ro" -v "$IDN_IN:/i:ro" "$VT_IMG" logcheck scan --canaries /i/canaries.txt /b >"$WORK/idn-scan-report.log" 2>&1 || { ok=0; log "a registered value in the browser evidence: $(head -3 "$WORK/idn-scan-report.log" | tr '\n' ' ')"; }
  # P-NEG-1 and P-NEG-3: the emulator API key planted into would-be evidence and into a core log line is detected.
  planted="$pv/planted"; mkdir -p "$planted"
  local key; key="$(grep -m1 '^sk_test_' "$SECRETS_FILE" || true)"
  [[ -n "$key" ]] || { ok=0; log "P-NEG: the emulator API key is not in the registry"; }
  printf 'evidence line %s\n' "$key" >"$planted/evidence.txt"
  printf '{"time":"2026-10-07T00:00:00.000Z","level":"info","service":"core","event":"core listening","port":8180,"app_env":"%s"}\n' "$key" >"$planted/core.log"
  chmod -R a+rX "$planted"
  if vtools -v "$planted:/p:ro" -v "$IDN_IN:/i:ro" "$VT_IMG" logcheck scan --canaries /i/canaries.txt /p/evidence.txt >/dev/null 2>&1; then ok=0; log "P-NEG-1: the planted API key in evidence was not detected"; fi
  if vtools -v "$planted:/p:ro" -v "$IDN_IN:/i:ro" "$VT_IMG" logcheck scan --canaries /i/canaries.txt /p/core.log >/dev/null 2>&1; then ok=0; log "P-NEG-3: the planted API key in a core log line was not detected"; fi
  # P-POS-2 and P-NEG-2: delivered web bodies (pages and every static chunk they load) hold no identity metadata or key;
  # the same check on a copy with the client ID planted fails.
  identity_bundle_clean "$dest" "$project" "$pv" || ok=0
  rm -rf "$planted" "$reg"
  [[ $ok == 1 ]] && record B-IDENTITY-PRIVACY PASS "core, replica and web logs fully catalogued (logcheck validate); no registered value (API key, passwords, display/email/script canaries, observed session, binding, code, state, CSRF and logout values, and the G062 signup reference, nonce, proof, signup credential, challenge and form-record values and organization-name canaries) in core, replica, web or proxy logs; no email in the identity tables; P-POS-1 sanitized authorize evidence names the client and passes the scan; P-NEG-1 and P-NEG-3 planted API key detected in evidence and in a core log line; P-POS-2 delivered identity pages and static chunks carry no client ID, issuer, API base or key, and P-NEG-2 the planted client ID is detected" || record B-IDENTITY-PRIVACY FAIL "identity privacy"
}

# identity_bundle_clean DEST PROJECT DIR: the delivered bodies of the identity pages and of every static chunk, from the
# running web image; then the negative control on a planted copy. G016: pages are requested with the routing platform
# host org-a.localhost:543 as Host (the tenant-host gate answers any other host with the fixed 404 page, which loads no
# chunk); one same-host redirect is followed (/account without a session goes to /sign-in).
identity_bundle_clean() {
  local dest="$1" project="$2" dir="$3" web ok=1 bodies="$3/bodies" cid
  web="$(cid_of "$dest" web || true)"; mkdir -p "$bodies"
  docker exec "$web" node -e '
const http = require("node:http");
const pages = ["/sign-in", "/sign-in?status=unavailable", "/account", "/api/account", "/"];
const get = (p) => new Promise((resolve, reject) => {
  http.get({ host: "127.0.0.1", port: 3000, path: p, headers: { host: "org-a.localhost:543" } }, (r) => {
    let b = ""; r.setEncoding("utf8"); r.on("data", (c) => { b += c; });
    r.on("end", () => resolve({ status: r.statusCode, location: r.headers.location, body: b }));
  }).on("error", reject);
});
(async () => {
  const seen = new Set(); const out = [];
  for (const p of pages) {
    let r = await get(p);
    if (r.status >= 300 && r.status < 400 && typeof r.location === "string" && r.location.startsWith("/")) r = await get(r.location);
    const t = r.body; out.push(t);
    for (const m of t.matchAll(/\/_next\/static\/[^"'"'"' )]+/g)) seen.add(m[0]);
  }
  for (const s of seen) out.push((await get(s)).body);
  process.stdout.write(JSON.stringify({ pages: pages.length, chunks: seen.size }) + "\n" + out.join("\n"));
})().catch(() => process.exit(1));' >"$bodies/delivered.txt" 2>/dev/null || { ok=0; log "P-POS-2: delivered bodies unavailable"; }
  [[ "$(head -1 "$bodies/delivered.txt")" =~ \"chunks\":[1-9] ]] || { ok=0; log "P-POS-2: no static chunk was collected"; }
  if grep -qF -e "$IDN_CLIENT" -e "$IDN_ISSUER" -e "$IDN_BASE" -e 'sk_test_' "$bodies/delivered.txt"; then ok=0; log "P-POS-2: identity metadata or a key in a delivered web body"; fi
  cid="$(docker_create_owned web-next-copy --label g001.run=$RUN_ID "$project-web")" && docker cp "$cid:/app/.next" "$bodies/next" >/dev/null 2>&1; docker rm -f "$cid" >/dev/null 2>&1 || true
  if grep -rqF -e "$IDN_CLIENT" -e "$IDN_ISSUER" -e "$IDN_BASE" -e 'sk_test_' "$bodies/next" 2>/dev/null; then ok=0; log "P-POS-2: identity metadata or a key in the web build output"; fi
  printf '%s\n' "$IDN_CLIENT" >>"$bodies/next/static/planted.js"
  grep -rqF -e "$IDN_CLIENT" "$bodies/next" || { ok=0; log "P-NEG-2: the planted client ID was not detected"; }
  rm -rf "$bodies"
  return $((1 - ok))
}

# B-IDENTITY-CONFIG: the real images refuse unsafe identity configuration in production and carry no test artefact.
identity_config() { # identity_config DEST PROJECT
  local dest="$1" project="$2" ok=1 in="$IDN_IN/config" D="$WORK/identity-config" c key cid rc i canary cases=""
  local img="$project-core"
  mkdir -p "$in" "$D"; chmod 777 "$D"
  canary="$(new_canary)"
  core_env "$in/base.env" ":8180" "Probe Product"
  sed -i 's/^APP_ENV=test$/APP_ENV=production/' "$in/base.env"
  # G009 (Plan r1 §4.7, §5.6): production refuses an http object-store data endpoint, so the production base uses https
  # (never connected: --network none); cc1-cc7 keep their single identity refusal and cc8 still starts.
  sed -i 's#^OBJECT_STORE_ENDPOINT=http://#OBJECT_STORE_ENDPOINT=https://#' "$in/base.env"
  grep -qx 'OBJECT_STORE_ENDPOINT=https://minio.invalid:9000' "$in/base.env" || { ok=0; log "the production base has no https object-store endpoint"; }
  key="IDENTITY_WORKOS_API_KEY=sk_live_$canary"
  local full=$'IDENTITY_PROVIDER=workos\nIDENTITY_WORKOS_CLIENT_ID=client_01PRODUCTIONPROBE\nIDENTITY_WORKOS_ISSUER=https://api.workos.com/user_management/client_01PRODUCTIONPROBE\n'"$key"$'\nIDENTITY_PUBLIC_ORIGINS=https://app.probe.example.net\nIDENTITY_COOKIE_PROFILE=https'
  # name|extra environment lines: each must exit 1 with one core.exit config_invalid record.
  local -a refuse=(
    "cc1-unconfigured|"
    "cc2-partial|IDENTITY_WORKOS_CLIENT_ID=client_01PRODUCTIONPROBE"
    "cc3-emulator-base|$full"$'\n'"IDENTITY_WORKOS_API_BASE_URL=http://emulator:4100"
    "cc4-loopback-http|${full%IDENTITY_COOKIE_PROFILE=https}IDENTITY_COOKIE_PROFILE=loopback-http"
    "cc5-http-issuer|${full/https:\/\/api.workos.com/http://api.workos.com}"
    "cc6-test-lifetime|$full"$'\n'"IDENTITY_SESSION_IDLE=1m"
    "cc7-provider-oidc|${full/IDENTITY_PROVIDER=workos/IDENTITY_PROVIDER=oidc}")
  for c in "${refuse[@]}"; do
    { cat "$in/base.env"; [[ -n "${c#*|}" ]] && printf '%s\n' "${c#*|}"; } >"$in/${c%%|*}.env"
    rc=0; docker_run_owned identity-cc --rm --label g001.run=$RUN_ID --network none --env-file "$in/${c%%|*}.env" "$img" >"$D/${c%%|*}.out" 2>&1 || rc=$?; echo "$rc" >"$D/${c%%|*}.out.rc"
    cases+=" --case /d/${c%%|*}.out=1:core.exit:config_invalid --exactly-one-line /d/${c%%|*}.out"
  done
  # An MFA-labelled variable on a valid production configuration is ignored: core starts as usual and stops cleanly.
  { cat "$in/base.env"; printf '%s\nIDENTITY_MFA_REQUIRED=true\nIDENTITY_MFA_VERIFIED=true\n' "$full"; } >"$in/cc8-mfa.env"
  cid="$(docker_run_owned identity-cc8 -d --label g001.run=$RUN_ID --network none --env-file "$in/cc8-mfa.env" "$img")" || ok=0
  for ((i = 0; i < 60; i++)); do docker logs "$cid" 2>&1 | grep -q '"event":"core listening"' && break; sleep 1; done
  docker stop "$cid" >/dev/null 2>&1 || true
  rc="$(docker wait "$cid" 2>/dev/null || echo 255)"; docker logs "$cid" >"$D/cc8-mfa.out" 2>&1 || true; echo "$rc" >"$D/cc8-mfa.out.rc"; docker rm -f "$cid" >/dev/null 2>&1 || true
  grep -qi 'mfa' "$D/cc8-mfa.out" && { ok=0; log "core mentions MFA configuration"; }
  printf '%s\n' "$canary" >"$D/canaries.txt"
  # shellcheck disable=SC2086 # $cases is a list of separate logcheck arguments
  vtools -v "$D:/d:ro" "$VT_IMG" logcheck startup --service core $cases --case "/d/cc8-mfa.out=0:core listening:port=8180+app_env=production" --canaries /d/canaries.txt >"$WORK/identity-config-core.log" 2>&1 || { ok=0; tail -20 "$WORK/identity-config-core.log"; }
  # Web: the loopback-http profile is refused in production (503 for both layers); unset means https there.
  local w env="$in/web.env"
  printf 'PRODUCT_NAME=Probe Product\nPLATFORM_BASE_DOMAIN=probe.example.net\nCUSTOM_DOMAIN_CNAME_TARGET=custom.probe.example.net\nCORE_INTERNAL_URL=http://127.0.0.1:9\nAPP_ENV=production\nIDENTITY_COOKIE_PROFILE=loopback-http\n' >"$env"
  w="$(docker_run_owned identity-web-config -d --label g001.run=$RUN_ID --network none --env-file "$env" "$project-web")" || ok=0
  for ((i = 0; i < 60; i++)); do idn_hp "$w" wc-health --url http://127.0.0.1:3000/api/health --timeout 3s; [[ "$(idn_get wc-health .status)" == 200 ]] && break; sleep 1; done
  idn_hp "$w" wc-account --url http://127.0.0.1:3000/api/account --header "Host: 127.0.0.1:3000" --timeout 10s
  idn_hp "$w" wc-login --url http://127.0.0.1:3000/auth/login --method POST --header "Origin: http://127.0.0.1:3000" --header "Content-Type: application/x-www-form-urlencoded" --timeout 10s
  [[ "$(idn_get wc-account .status)" == 503 && "$(idn_get wc-account .body)" == '{"error":"unavailable"}' && "$(idn_get wc-login .status)" == 503 ]] || { ok=0; log "web with loopback-http in production: $(idn_get wc-account .status) $(idn_get wc-login .status)"; }
  docker rm -f "$w" >/dev/null 2>&1 || true
  # No identity test package, emulator reference or test key in the production core binary or the web build.
  cid="$(docker_create_owned server-copy --label g001.run=$RUN_ID "$img")" && docker cp "$cid:/usr/local/bin/server" "$D/server" >/dev/null 2>&1; docker rm -f "$cid" >/dev/null 2>&1 || true
  [[ -s "$D/server" ]] || { ok=0; log "core binary unavailable"; }
  grep -qaE 'identitytest|_emulate|workos-emulator|sk_test_' "$D/server" && { ok=0; log "test artefact in the core binary"; }
  rm -rf "$D" "$in"
  [[ $ok == 1 ]] && record B-IDENTITY-CONFIG PASS "real core image, APP_ENV=production, no network: unconfigured, partial, emulator API base, loopback-http, http issuer, a test lifetime and IDENTITY_PROVIDER=oidc each exit 1 with exactly one core.exit config_invalid line; MFA-labelled variables on a valid configuration are ignored (core listening, clean stop, no MFA mention); no canary in any output; real web image: loopback-http in production answers 503 on /api/account and /auth/login; no identitytest, emulator reference or test key in the core binary" || record B-IDENTITY-CONFIG FAIL "identity configuration refusal"
}

# G004 B-IDENTITY-STAGING (staging suite; Plan r5 §8.2 L3, §8.5 CI-1): TestStagingContract (tag workos_staging) against
# the real WorkOS staging API with the credential file that infra/ci/hosted-context.sh forwards in the approved
# identity-staging context (IDENTITY_STAGING_ENV_FILE), on a disposable PostgreSQL. Missing inputs are UNMET, never a
# skip. Only the API key is registered as a secret; only the sanitized result is kept.
PORTS_STG=(3650 8730 8650 6082 7029 9650 9651)
suite_staging() {
  log "== staging (real WorkOS staging API; identity-staging job)"
  local S="$WORK/stg" project="g001-stg-$RUN_ID" f="${IDENTITY_STAGING_ENV_FILE:-}" k n=0 env="$WORK/secrets-staging-db.env" out="$WORK/staging-out" tlog="$WORK/staging-test.log" rc=0 ok=1
  if [[ -z "$f" || -L "$f" || ! -f "$f" ]]; then record B-IDENTITY-STAGING UNMET "no identity staging inputs (IDENTITY_STAGING_ENV_FILE)"; return; fi
  for k in IDENTITY_STAGING_WORKOS_API_KEY IDENTITY_STAGING_WORKOS_CLIENT_ID IDENTITY_STAGING_WORKOS_ISSUER; do
    if grep -qE "^$k=.+" "$f"; then n=$((n + 1)); fi
  done
  if [[ "$n" != 3 ]]; then record B-IDENTITY-STAGING UNMET "identity staging inputs: $n of 3 present"; return; fi
  sed -n 's/^IDENTITY_STAGING_WORKOS_API_KEY=//p' "$f" | awk 'length($0) >= 8' >>"$SECRETS_FILE"
  if ports_busy "${PORTS_STG[@]}"; then record B-IDENTITY-STAGING FAIL "ports ${PORTS_STG[*]} busy; refusing to re-map silently"; return; fi
  copy_repo "$S"
  if ! configure "$S" "$project" "${PORTS_STG[@]}" ||
     ! compose "$S" up -d --wait postgres >"$WORK/staging-up.log" 2>&1 ||
     ! compose_migrate "$S" run --rm --build migrate up >"$WORK/staging-migrate.log" 2>&1 ||
     ! compose "$S" --profile checks build core-check >"$WORK/staging-build.log" 2>&1; then
    tail -20 "$WORK/staging-up.log" "$WORK/staging-migrate.log" "$WORK/staging-build.log" 2>/dev/null
    record B-IDENTITY-STAGING FAIL "staging database or check image"; compose "$S" down -v >/dev/null 2>&1 || true; return
  fi
  (umask 077; printf 'DATABASE_NAME=%s\nDB_HOST=postgres\nDB_PORT=5432\nDB_MIGRATOR_PASSWORD=%s\nDB_RUNTIME_PASSWORD=%s\nAPP_ENV=test\nIDENTITY_STAGING_RESULT_FILE=/out/result.json\n' \
    "$(env_get "$S" DATABASE_NAME)" "$(env_get "$S" DB_MIGRATOR_PASSWORD)" "$(env_get "$S" DB_RUNTIME_PASSWORD)" >"$env")
  mkdir -p "$out"; chmod 777 "$out"
  docker_run_owned staging-test --rm --label g001.run=$RUN_ID --network "${project}_data" --env-file "$env" --env-file "$f" -v "$out:/out" \
    "$project-core-check" go test -tags workos_staging -run '^TestStagingContract$' -count=1 -v -timeout 600s ./internal/identity/workos/ >"$tlog" 2>&1 || rc=$?
  rm -f "$env"
  [[ $rc == 0 ]] && grep -q -- '^--- PASS: TestStagingContract ' "$tlog" && ! grep -qE -- '--- (SKIP|FAIL)' "$tlog" || { ok=0; tail -30 "$tlog"; }
  [[ -s "$out/result.json" ]] && grep -q '"result": "pass"' "$out/result.json" || { ok=0; log "no passing sanitized staging result"; }
  if [[ -n "$REPORT_DIR" && -s "$out/result.json" ]]; then cp "$out/result.json" "$REPORT_DIR/staging-result.json"; fi
  save_logs "$S" stg
  compose "$S" down -v >/dev/null 2>&1 || true
  [[ $ok == 1 ]] && record B-IDENTITY-STAGING PASS "TestStagingContract against the real WorkOS staging API (forwarded credential file; disposable PostgreSQL): exact issuer, client, subject, sid, auth_time and JWKS; S-MFA password alone refused, TOTP accepted, assurance unknown; S-ATTR, S-TAMP, S-CODE, S-REV; S-CLEAN per-run users deleted; sanitized result kept" || record B-IDENTITY-STAGING FAIL "identity staging contract (exit $rc)"
}

# ===================================================================================
# G051 feature areas (Plan r3 §3.2, §3.4). AREAS (--areas) is none or a sorted list of KIND:NAME. An area's roots are
# its registered dir rows in infra/ci/profile-paths.tsv (class area-web, area-core or area-intel); its name is the last
# path segment. A selected area without exactly its allowed roots is UNMET, never skipped.
area_names() { # area_names KIND -> the selected areas of KIND, one per line
  local a list=()
  [[ "$AREAS" == none ]] && return 0
  IFS=, read -r -a list <<<"$AREAS"
  for a in "${list[@]}"; do
    if [[ "${a%%:*}" == "$1" ]]; then printf '%s\n' "${a#*:}"; fi
  done
}
area_roots() { # area_roots KIND NAME -> the area's registered root directories, repository-relative
  awk -F'\t' -v c="area-$1" -v n="$2" '$1 == c && $2 == "dir" { p = $3; sub(/\/$/, "", p); k = split(p, s, "/"); if (s[k] == n) print $3 }' \
    "$REPO/infra/ci/profile-paths.tsv"
}
count_files() { # count_files DIR NAME-PATTERN -> the number of regular files below DIR named like the pattern (0 without DIR)
  if [[ ! -d "$1" ]]; then echo 0; return 0; fi
  { find "$1" -type f -name "$2" || true; } | wc -l
}

# G051 B-FEATURE-WEB (feature-web; Plan r3 §3.4): one Playwright execution with FEATURE_E2E=1 over the selected web
# areas' specs (tests/browser/features/<n>/): the web-browser image (--target browser, which carries the fixture stage's
# production build) in a run-labelled container without network; or, when a selected area has a *.stack.spec.ts, the
# browser service on a disposable full stack (PORTS_FEAT) with FEATURE_STACK=1 as well. A selected area with page routes
# and no spec is UNMET. Every test passes and carries S-ERRORS (expected > 0; 0 skipped, flaky or unexpected), only the
# feature projects run, and every area with specs has tests in the report.
suite_feature_web() {
  log "== feature-web (areas: $AREAS)"
  local rep="$WORK/browser-feature" rows="$WORK/browser-feature-tests.tsv" ok=1 unmet="" a r p rc=0 n pages specs stackspecs
  local mode=fixture fixture=0 img="g001-feature-web-$RUN_ID" D="$WORK/featw" project="g001-featw-$RUN_ID" where
  local tarf="$WORK/browser-feature.tar" runc="g001-featw-$RUN_ID-browser" xfer=0 xerr=""
  local -a areas=() roots=() filters=() projects=() pargs=() spec_areas=() other_areas=() penv=()
  mapfile -t areas < <(area_names web)
  if (( ${#areas[@]} == 0 )); then record B-FEATURE-WEB UNMET "feature-web selected without a selected web area (--areas $AREAS)"; return; fi
  for a in "${areas[@]}"; do
    mapfile -t roots < <(area_roots web "$a")
    if (( ${#roots[@]} == 0 )); then unmet+="web:$a has no registered root; "; continue; fi
    pages=0 specs=0 stackspecs=0
    for r in "${roots[@]}"; do
      case "$r" in
        "web/app/(features)/$a/") pages=$((pages + $(count_files "$REPO/$r" 'page.*'))) ;;
        "web/tests/browser/features/$a/")
          specs="$(count_files "$REPO/$r" '*.spec.ts')"
          stackspecs="$(count_files "$REPO/$r" '*.stack.spec.ts')" ;;
        "web/tests/fixtures/features/$a/") ;;
        *) unmet+="web:$a has the unrecognized root $r; " ;;
      esac
    done
    if (( pages > 0 && specs == 0 )); then unmet+="web:$a has $pages page route(s) and no spec; "; fi
    if (( specs > 0 )); then
      spec_areas+=("$a"); filters+=("tests/browser/features/$a/")
      if (( specs > stackspecs )); then fixture=1; fi
      if (( stackspecs > 0 )); then mode=stack; fi
    else other_areas+=("$a"); fi
  done
  if [[ -n "$unmet" ]]; then record B-FEATURE-WEB UNMET "${unmet%; }"; return; fi
  if (( ${#spec_areas[@]} == 0 )); then record B-FEATURE-WEB UNMET "no feature spec in the selected web areas (${areas[*]})"; return; fi
  if (( fixture )); then projects+=(feature-light feature-dark); fi
  if [[ "$mode" == stack ]]; then projects+=(feature-stack-light feature-stack-dark); fi
  for p in "${projects[@]}"; do pargs+=("--project=$p"); done
  pargs+=("${filters[@]}")
  mkdir -p "$rep"
  # No host report bind mount (G051 O4): Playwright writes its report inside the container. In the fixture branch the
  # container's shell then streams that directory to stdout as a tar and exits with Playwright's own status (its output
  # goes to stderr, the log); the container stays --rm and run-labelled. In the stack branch the Compose path does not
  # give a clean stdout, so the one-off container runs under a unique name and the report is copied out of it
  # (docker cp) before that container is removed; an interrupted run's container is removed with its project.
  # shellcheck disable=SC2016  # expanded by the container's shell
  local wrap='npx playwright test "$@" >&2; rc=$?; tar -C /tmp/report -cf - . || echo "no report in /tmp/report" >&2; exit "$rc"'
  penv=(-e PLAYWRIGHT_JSON_OUTPUT_FILE=/tmp/report/results.json -e PLAYWRIGHT_OUTPUT_DIR=/tmp/report/out
    -e PLAYWRIGHT_WORKERS="$PW_WORKERS" -e FEATURE_E2E=1)
  if [[ "$mode" == fixture ]]; then
    where="the web-browser image's fixture build in a run-labelled container without network"
    if (( PREBUILD_WEB )); then bg_wait_optional prebuild-web-browser; fi
    if ledger_build "$LEDGER_SUITE" authoritative -- "${FEATURE_WEB_BUILD[@]}" -t "$img" "$REPO/web" >"$WORK/feature-web-build.log" 2>&1; then
      record_built_image "$img"
    else tail -20 "$WORK/feature-web-build.log"; record B-FEATURE-WEB FAIL "web-browser image build (--target browser)"; return; fi
    docker_run_owned feature-web --rm --label g001.run=$RUN_ID --network none --init --shm-size 1g "${penv[@]}" \
      -e WEB_BASE_URL=http://127.0.0.1:3001 "$img" sh -c "$wrap" feature-web "${pargs[@]}" >"$tarf" 2>"$WORK/feature-web.log" || rc=$?
    if [[ -s "$tarf" ]] && tar -xf "$tarf" -C "$rep" --no-same-owner --no-same-permissions 2>>"$WORK/feature-web.log"; then xfer=1; fi
    rm -f -- "$tarf"
  else
    where="a disposable full stack (README start, migrate up; ports ${PORTS_FEAT[*]}) with FEATURE_STACK=1"
    if ports_busy "${PORTS_FEAT[@]}"; then record B-FEATURE-WEB FAIL "ports ${PORTS_FEAT[*]} busy; refusing to re-map silently"; return; fi
    copy_repo "$D"
    if ! configure "$D" "$project" "${PORTS_FEAT[@]}" ||
       ! run_docs_block "$D" README.md platform:start "$WORK/featw-start.log" ||
       ! compose_migrate "$D" run --rm --build migrate up >"$WORK/featw-migrate.log" 2>&1; then
      tail -30 "$WORK/featw-start.log" "$WORK/featw-migrate.log" 2>/dev/null || true
      record B-FEATURE-WEB FAIL "the disposable stack for the stack specs did not start (README platform:start, migrate up)"
      save_logs "$D" featw; compose "$D" down -v >/dev/null 2>&1 || true; return
    fi
    if (( PREBUILD_WEB )); then bg_wait_optional prebuild-web-browser; fi
    compose_browser "$D" --profile browser run --build --no-deps --name "$runc" "${penv[@]}" -e FEATURE_STACK=1 \
      web-browser npx playwright test "${pargs[@]}" >"$WORK/feature-web.log" 2>&1 || rc=$?
    if docker cp "$runc:/tmp/report/." "$rep" >"$WORK/feature-web-cp.err" 2>&1; then xfer=1; else
      # The sanitized first line of the copy error (a Docker daemon message naming the container and path) is kept in
      # the check detail; the work files go with the work root.
      xerr="$(head -n 1 "$WORK/feature-web-cp.err" | tr -cd '[:alnum:] _./:()-' | cut -c1-200)"
    fi
    cat "$WORK/feature-web-cp.err" >>"$WORK/feature-web.log" 2>/dev/null || true
    docker rm -f "$runc" >/dev/null 2>&1 || true
    save_logs "$D" featw
    compose "$D" down -v >/dev/null 2>&1 || true
  fi
  (( xfer )) || { ok=0; log "the feature browser report could not be transferred from the container${xerr:+ ($xerr)}"; }
  [[ $rc == 0 ]] || { ok=0; tail -30 "$WORK/feature-web.log"; log "the feature browser run exited $rc"; }
  browser_matrix "$rep/results.json" "$rows" "${projects[*]}|S-ERRORS" || ok=0
  if [[ -s "$rows" ]]; then
    n="$(awk -F'\t' -v p=" ${projects[*]} " '!index(p, " " $1 " ")' "$rows" | wc -l)"
    [[ "$n" == 0 ]] || { ok=0; log "$n tests outside the feature projects ${projects[*]}"; }
  fi
  : >"$WORK/browser-feature-files.txt"
  if [[ -s "$rep/results.json" ]]; then
    yq_stdin -p json -o tsv '.. | select(tag == "!!map" and has("specs")) | .specs[] | .file' <"$rep/results.json" >"$WORK/browser-feature-files.txt" 2>/dev/null || true
  fi
  for a in "${spec_areas[@]}"; do
    grep -qF "features/$a/" "$WORK/browser-feature-files.txt" || { ok=0; log "no test of web:$a in the report"; }
  done
  [[ $ok == 1 ]] && record B-FEATURE-WEB PASS "web areas ${spec_areas[*]} (FEATURE_E2E=1; $PW_WORKERS Playwright workers; $where): $(grep -c . "$rows") tests in ${projects[*]}, each passing with S-ERRORS, 0 skipped/flaky/unexpected; every selected area with page routes has a spec and every area with specs has tests in the report${other_areas[*]:+; selected areas without page routes or specs: ${other_areas[*]}}" || record B-FEATURE-WEB FAIL "feature browser specs${xerr:+; report transfer: $xerr}"
}

# G051 B-FEATURE-CORE (feature-core; Plan r3 §3.4): for each selected core area with a //go:build integration file, its
# packages' tests with -tags integration in the core check image on a disposable project's data network (PORTS_FEAT):
# PostgreSQL migrated up, plus Redis or MinIO only when those packages (with their test dependencies) link their clients;
# the 0600 environment file of B-IDENTITY-IT. Per area: go test exits 0 with an ok package line, at least one top-level
# test passes, and no test fails or is skipped.
suite_feature_core() {
  log "== feature-core (areas: $AREAS)"
  local S="$WORK/featc" project="g001-featc-$RUN_ID" in env a ok=1 ready=1 unmet="" deps rc passed tlog detail="" redis=0 minio=0 img
  local -a areas=() roots=() it=() noit=() pkgs=() svcs=(postgres)
  mapfile -t areas < <(area_names core)
  if (( ${#areas[@]} == 0 )); then record B-FEATURE-CORE UNMET "feature-core selected without a selected core area (--areas $AREAS)"; return; fi
  for a in "${areas[@]}"; do
    mapfile -t roots < <(area_roots core "$a")
    if [[ ${#roots[@]} != 1 || "${roots[0]}" != "core/internal/$a/" || ! -d "$REPO/core/internal/$a" ]]; then
      unmet+="core:$a is not exactly one registered, present core/internal/$a/ root; "; continue
    fi
    if grep -rqE --include='*.go' '^//go:build .*\bintegration\b' "$REPO/core/internal/$a"; then it+=("$a"); pkgs+=("./internal/$a/..."); else noit+=("$a"); fi
  done
  if [[ -n "$unmet" ]]; then record B-FEATURE-CORE UNMET "${unmet%; }"; return; fi
  if (( ${#it[@]} == 0 )); then record B-FEATURE-CORE UNMET "no selected core area (${areas[*]}) has a //go:build integration file"; return; fi
  if ports_busy "${PORTS_FEAT[@]}"; then record B-FEATURE-CORE FAIL "ports ${PORTS_FEAT[*]} busy; refusing to re-map silently"; return; fi
  copy_repo "$S"
  if ! configure "$S" "$project" "${PORTS_FEAT[@]}" || ! compose "$S" --profile checks build core-check >"$WORK/featc-build.log" 2>&1; then
    tail -20 "$WORK/featc-build.log" 2>/dev/null || true
    record B-FEATURE-CORE FAIL "configure or the core check image"; compose "$S" down -v >/dev/null 2>&1 || true; return
  fi
  img="$project-core-check"
  # The stores: PostgreSQL always; Redis or MinIO when the selected packages, with their test dependencies, link the client.
  if ! deps="$(docker_run_owned featc-deps --rm --label g001.run=$RUN_ID --network none "$img" go list -deps -test -tags integration "${pkgs[@]}" 2>"$WORK/featc-deps.err")"; then
    tail -20 "$WORK/featc-deps.err"; record B-FEATURE-CORE FAIL "go list of the selected packages (${pkgs[*]})"; compose "$S" down -v >/dev/null 2>&1 || true; return
  fi
  if grep -qE '^github\.com/(redis/go-redis|gomodule/redigo)(/|$)' <<<"$deps"; then redis=1; svcs+=(redis); fi
  if grep -qE '^github\.com/(aws/aws-sdk-go-v2/service/s3|minio/minio-go)(/|$)' <<<"$deps"; then minio=1; svcs+=(minio); fi
  compose "$S" up -d --wait "${svcs[@]}" >"$WORK/featc-up.log" 2>&1 || ready=0
  if (( ready && minio )); then compose "$S" run --rm minio-init >>"$WORK/featc-up.log" 2>&1 || ready=0; fi
  if (( ready )); then compose_migrate "$S" run --rm --build migrate up >"$WORK/featc-migrate.log" 2>&1 || ready=0; fi
  if (( ! ready )); then
    tail -30 "$WORK/featc-up.log" "$WORK/featc-migrate.log" 2>/dev/null || true
    record B-FEATURE-CORE FAIL "disposable ${svcs[*]} or migrate up"; save_logs "$S" featc; compose "$S" down -v >/dev/null 2>&1 || true; return
  fi
  in="$(obs_inputs)"; env="$in/feature-core-it.env"
  (umask 077; {
    printf 'DATABASE_NAME=%s\nDB_HOST=postgres\nDB_PORT=5432\nDB_MIGRATOR_PASSWORD=%s\nDB_RUNTIME_PASSWORD=%s\nAPP_ENV=test\n' \
      "$(env_get "$S" DATABASE_NAME)" "$(env_get "$S" DB_MIGRATOR_PASSWORD)" "$(env_get "$S" DB_RUNTIME_PASSWORD)"
    if (( redis )); then
      printf 'REDIS_HOST=%s\nREDIS_PORT=%s\nREDIS_PASSWORD=%s\n' "$(env_get "$S" REDIS_HOST)" "$(env_get "$S" REDIS_PORT_INTERNAL)" "$(env_get "$S" REDIS_PASSWORD)"
    fi
    if (( minio )); then
      printf 'OBJECT_STORE_ENDPOINT=%s\nOBJECT_STORE_REGION=%s\nOBJECT_STORE_BUCKET=%s\nOBJECT_STORE_FORCE_PATH_STYLE=%s\nOBJECT_STORE_ACCESS_KEY=%s\nOBJECT_STORE_SECRET_KEY=%s\n' \
        "$(env_get "$S" OBJECT_STORE_ENDPOINT)" "$(env_get "$S" OBJECT_STORE_REGION)" "$(env_get "$S" OBJECT_STORE_BUCKET)" \
        "$(env_get "$S" OBJECT_STORE_FORCE_PATH_STYLE)" "$(env_get "$S" OBJECT_STORE_ACCESS_KEY)" "$(env_get "$S" OBJECT_STORE_SECRET_KEY)"
    fi
  } >"$env")
  for a in "${it[@]}"; do
    tlog="$WORK/feature-core-$a.log"; rc=0
    docker_run_owned feature-core-it --rm --label g001.run=$RUN_ID --network "${project}_data" --env-file "$env" "$img" \
      go test -tags integration -count=1 -v "./internal/$a/..." >"$tlog" 2>&1 || rc=$?
    passed="$(grep -cE -- '^--- PASS: ' "$tlog" || true)"
    if [[ $rc == 0 ]] && grep -qE "^ok[[:space:]]+core/internal/$a(/[^[:space:]]+)?[[:space:]]" "$tlog" && ! grep -qE -- '--- (FAIL|SKIP)|^FAIL' "$tlog" && [[ "$passed" -ge 1 ]]; then
      detail+="$a: $passed top-level tests; "
    else ok=0; tail -40 "$tlog"; log "core:$a integration tests: exit $rc, $passed top-level tests passed"; fi
  done
  rm -f -- "$env"
  save_logs "$S" featc
  compose "$S" down -v >/dev/null 2>&1 || true
  [[ $ok == 1 ]] && record B-FEATURE-CORE PASS "core areas ${it[*]}: go test -tags integration -count=1 in the core check image against a disposable ${svcs[*]} (migrated up): ${detail%; }; every package ok, nothing failed or skipped${noit[*]:+; selected areas without integration tests (unit-core covers them): ${noit[*]}}" || record B-FEATURE-CORE FAIL "feature integration tests"
}

# ===================================================================================
final_checks() {
  local abs leaks
  abs="$(grep -nE '(^|[^A-Za-z0-9_./-])/usr/(local/)?bin/(go|python3?|node|npm|npx|psql|redis-cli)\b' "$REPO"/infra/validation/*.sh "$REPO"/infra/local/*.sh || true)"
  if [[ -s "$SHADOW_LOG" ]]; then record B-DOCKER-ONLY FAIL "host runtime invoked: $(head -3 "$SHADOW_LOG" | tr '\n' ';')"
  elif [[ -n "$abs" ]]; then record B-DOCKER-ONLY FAIL "absolute host runtime path in scripts: $abs"
  else record B-DOCKER-ONLY PASS "no host Go/Python/Node/psql/redis-cli/mc/minio/aws invocation (shadowed; /mnt/* stripped from PATH); no absolute runtime paths"; fi

  # Every generated secret must be absent from evidence and repository files.
  # The de-duplicated list stays under $WORK (removed with it), never beside an external --secrets-file.
  local su="$WORK/secrets.u"
  sort -u "$SECRETS_FILE" | grep -v '^$' >"$su" || true
  if [[ -s "$su" ]]; then
    # Positive control: a generated secret planted in an evidence file must be found by this scan.
    local planted="$WORK/canary-control/planted.txt" found
    mkdir -p "$WORK/canary-control"; head -1 "$su" >"$planted"
    found="$(grep -rlF -f "$su" "$WORK/canary-control" 2>/dev/null || true)"
    rm -rf "$WORK/canary-control"
    [[ "$found" == "$planted" ]] || record B-SECRET-CANARY FAIL "control: a planted secret was not detected by the scan"
    leaks="$(grep -rlF -f "$su" --exclude=.env --exclude='secrets*' --exclude=it.env --exclude=static.env --exclude='exec-token' --exclude-dir=obs-inputs "$WORK" 2>/dev/null | grep -v '/shadow-bin/' || true)"
    leaks+="$(cd "$REPO" && git ls-files -z --cached --others --exclude-standard | xargs -0 grep -lF -f "$su" 2>/dev/null || true)"
    [[ -z "$leaks" ]] && record B-SECRET-CANARY PASS "$(wc -l <"$su") generated secrets absent from harness output, compose and migrate logs, browser reports and saved response bodies, manifests and every non-ignored repository file (a planted secret is detected)" || record B-SECRET-CANARY FAIL "secret found in: ${leaks//$'\n'/ }"
  else record B-SECRET-CANARY UNMET "no secrets were generated"; fi

  # Change scope: every changed path matches the allow-list or is a named exception.
  # Paths are enumerated NUL-separated and exceptions compared exactly (names may contain spaces).
  local allow="$WORK/allow.re" bad="" p e n=0
  { grep -vE '^\s*(#|$)' "$ALLOW_FILE" || true; } >"$allow"
  while IFS= read -r -d '' p; do
    [[ -n "$p" ]] || continue
    n=$((n + 1))
    for e in "${EXCEPTIONS[@]}"; do [[ "$p" == "$e" ]] && { record B-SCOPE-EXCEPTION INFO "$p: $(git -C "$REPO" diff --stat "$BASE" -- "$p" | tail -1)"; continue 2; }; done
    grep -qE -f "$allow" <<<"$p" || bad+="$p "
  done < <( { git -C "$REPO" diff -z --name-only "$BASE" -- . ":(top,exclude).agents/skills/restruct-sdlc" ":(top,exclude).claude/skills/restruct-sdlc"; git -C "$REPO" ls-files -z -o --exclude-standard -- . ":(top,exclude).agents/skills/restruct-sdlc" ":(top,exclude).claude/skills/restruct-sdlc"; } | LC_ALL=C sort -z -u)
  [[ -z "$bad" ]] && record B-SCOPE PASS "$n changed paths since $(git -C "$REPO" rev-parse --short "$BASE"), all within the allow-list" || record B-SCOPE FAIL "outside the allow-list: $bad"

  { grep -h '^FROM ' "$REPO"/infra/docker/*.Dockerfile | awk '{print $2}' | grep -E '/|:'
    grep -h 'image: ' "$REPO"/infra/compose/*.yml | awk '{print $2}'
    printf '%s\n' "$YQ_IMAGE" "$BUSYBOX_IMAGE" "$SHELLCHECK_IMAGE" "$HADOLINT_IMAGE" golang:1.27.3-alpine; } | LC_ALL=C sort -u >"$WORK/images.txt"
  record IMAGES INFO "$(tr '\n' ' ' <"$WORK/images.txt")"
  required_checks
}

# B-REQUIRED-CHECKS: every ID listed for an executed suite (and the final checks) occurs exactly once as
# PASS; a missing, duplicated, FAIL or UNMET required ID fails.
required_checks() {
  local f="$REPO/infra/validation/required-checks.tsv" id suite n st bad="" want=0 s
  local -A run=([final]=1)
  for s in "${SUITES[@]}"; do run[$s]=1; done
  [[ -f "$f" ]] || { record B-REQUIRED-CHECKS FAIL "required-checks.tsv missing"; return; }
  while IFS=$'\t' read -r id suite; do
    [[ -z "$id" || "$id" == \#* || "$id" == B-REQUIRED-CHECKS ]] && continue
    [[ -n "${run[$suite]:-}" ]] || continue
    want=$((want + 1))
    n="$(awk -F'\t' -v i="$id" '$1==i' "$RESULTS" | wc -l)"
    st="$(awk -F'\t' -v i="$id" '$1==i{print $2}' "$RESULTS" | head -1)"
    [[ "$n" == 1 && "$st" == PASS ]] || bad+="$id($n,${st:-absent}) "
  done <"$f"
  [[ -z "$bad" ]] && record B-REQUIRED-CHECKS PASS "$want required IDs of the executed suites (${SUITES[*]}, final) each recorded exactly once as PASS (infra/validation/required-checks.tsv)" || record B-REQUIRED-CHECKS FAIL "required checks missing, duplicated or not PASS: $bad"
}

# G051 L5 (Plan r3 §3.6): with a later suite selected, controls runs as a supervised background child (private results,
# timings and log under $WORK/bg/controls/, merged by bg_join) and writes its own suite timing line; otherwise, and for
# every other suite, in the foreground. Each suite's builds are attributed through LEDGER_SUITE.
bg_controls() {
  local t0; t0="$(date +%s)"
  suite_controls
  printf 'suite\tcontrols\t%s\t%s\n' "$t0" "$(date +%s)" >>"${TIMINGS:-$WORK/timings.tsv}"
}
bg_ci_controls() {
  local t0; t0="$(date +%s)"
  suite_ci_controls
  printf 'suite\tci-controls\t%s\t%s\n' "$t0" "$(date +%s)" >>"${TIMINGS:-$WORK/timings.tsv}"
}
# G064 C2 scopes of the foreground suites; a suite without its own C2 row is its own scope with the run-wide ceiling.
res_suite_scope() { case "$1" in migrations) echo B-MIG ;; root) echo S-ROOT ;; ops) echo S-OPS ;; rename) echo S-REN ;; off) echo S-OFF ;; *) echo "suite-$1" ;; esac; }
CI_CONTROLS_BG=0
CONTROLS_BG=0
# G064 D1: the run-scoped creation witness (with a registry), before any suite allocates; stopped after the joins below.
res_witness_start
for i in "${!SUITES[@]}"; do
  s="${SUITES[$i]}"
  LEDGER_SUITE="$s"
  if [[ "$s" == controls && $((i + 1)) -lt ${#SUITES[@]} ]]; then
    log "== controls (supervised background child; joined before the final checks)"
    bg_start controls bg_controls; CONTROLS_BG=1
    continue
  fi
  if [[ "$s" == ci-controls && $((i + 1)) -lt ${#SUITES[@]} ]]; then
    log "== ci-controls (supervised background child; joined before final checks)"
    bg_start ci-controls bg_ci_controls; CI_CONTROLS_BG=1
    continue
  fi
  res_scope_begin "$(res_suite_scope "$s")"
  t0="$(date +%s)"; "suite_${s//-/_}"
  res_scope_end "$(res_suite_scope "$s")"
  printf 'suite\t%s\t%s\t%s\n' "$s" "$t0" "$(date +%s)" >>"$WORK/timings.tsv"
done
if (( CONTROLS_BG )); then bg_join controls controls; fi
if (( CI_CONTROLS_BG )); then bg_join ci-controls ci-controls; fi
# A prebuild not yet waited for (its consumer returned early) ends here, before the ledger and the final checks.
if (( PREBUILD_WEB )); then bg_wait_optional prebuild-web-browser; fi
# G064 C1/D1: the witness stops (final marker, parser joined, W verified gone), then resources.tsv and its RESOURCES rows.
res_witness_stop
res_report
LEDGER_SUITE=final
ledger_check
t0="$(date +%s)"; final_checks
printf 'suite\tfinal\t%s\t%s\n' "$t0" "$(date +%s)" >>"$WORK/timings.tsv"

log "== summary"
LC_ALL=C sort -t$'\t' -k1,1 "$RESULTS" | awk -F'\t' '{printf "%-24s %-6s %s\n", $1, $2, $3}'
if [[ -n "$REPORT_DIR" ]]; then
  cp "$RESULTS" "$WORK/images.txt" "$WORK/timings.tsv" "$REPORT_DIR/" 2>/dev/null || true
  cp "$WORK"/resources.tsv "$REPORT_DIR/" 2>/dev/null || true
  cp "$WORK"/catalog-p.log "$WORK"/migrate-it.log "$WORK"/deps-server.txt "$REPORT_DIR/" 2>/dev/null || true
  cp "$WORK"/deps-controls.tsv "$WORK"/deps-codes.txt "$WORK"/wrapper-controls.log "$WORK"/browser-*-tests.tsv "$REPORT_DIR/" 2>/dev/null || true
  # G003 evidence (statuses, categories and validator results only; never captured service output).
  cp "$WORK"/ci-controls.log "$WORK"/contracts-results-go.json "$WORK"/contracts-results-python.json "$WORK"/contracts-agree.tsv \
    "$WORK"/schema-negative-go.json "$WORK"/schema-negative-python.json "$WORK"/error-output-core.tsv "$WORK"/error-output-web.tsv \
    "$WORK"/error-output-intelligence.tsv "$WORK"/obs-*.tsv "$REPORT_DIR/" 2>/dev/null || true
  cp "$REPO/infra/validation/required-checks.tsv" "$REPORT_DIR/" 2>/dev/null || true
  # G051: the build ledger, the controls runner's per-control timings (browser-feature-tests.tsv is copied above).
  cp "$WORK"/builds.tsv "$WORK"/ledger-summary.tsv "$WORK"/ci-controls-timings.tsv "$REPORT_DIR/" 2>/dev/null || true
  for d in root rename off feature; do
    if [[ -f "$WORK/browser-$d/results.json" ]]; then cp "$WORK/browser-$d/results.json" "$REPORT_DIR/browser-$d-results.json"; fi
  done
fi
if grep -qE $'\t(FAIL|UNMET)\t' "$RESULTS"; then log "RESULT: FAILED"; exit 1; fi
log "RESULT: PASSED"
