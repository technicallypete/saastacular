#!/usr/bin/env bash
# CI controls (docs/foundation/OPERATIONS.md; B-CI-CONTROLS) for infra/ci/validate.sh, infra/ci/lib.sh,
# infra/ci/select-profile.sh, infra/ci/hosted-context.sh, infra/ci/cleanup-owned.sh and this runner's own
# cancellation: groups A-H.
#   A gate semantics and profile completeness (V1-V4, V7-V13f, VP1-VP9), affected and partition completeness
#     (VA1-VA11; G051) and the pipeline union check (VU1-VU9: infra/ci/pipeline-union.sh on published fixtures;
#     VU10-VU12: the RC-promotion pair, mixed tested content, release identity, independent staging observations)
#   B checkout attribution (P1-P10, P3F, N1-N23; ONEOFF RC: the detached promotion merge request P11, P11F, N40-N49
#     and the runner fetch shape RR1, RR2 against a fixture origin with tags and promotion branches)
#   C evidence scanning (V5a-V6d, V5d: the build-ledger diagnostic lines)
#   D finalization, supervision, location, cancellation, budgets and capacity (V14-V18n, V36, V37), the
#     supplemental JUnit report (J1-J8; G050, G051), the per-test Go evidence of the migrations and ops suites (GT1, GT2;
#     G064: a missing report, a failing plus an interrupted invocation) and the harness background supervision
#     (VB1-VB5; G051: drivers that source the real infra/validation/lib.sh; VB1 and VB3b run the real run.sh in a scratch copy
#     with a stub validate-controls.sh; G064 VB8: a background control's networks and volumes are not foreground, the real
#     _res_live through a fake docker)
#   E owned-resource cleanup, Docker-real (V19-V35), the build ledger (VC1-VC8; G051: busybox builds with
#     control labels through the real ledger_build/ledger_check, parser cases through a docker shim), and the
#     G051 cleanup amendment: the storage report (VS1-VS5) and local interrupted-run recovery (VR1-VR8; G064 VR9: the
#     batch runner and a self-registered controls run recovered through the marker/recovery entry path)
#   F profile selection (F-*: infra/ci/select-profile.sh on fixture repositories, no Docker; G051: fixture maps
#     with area rows, the dependency closure, suites, areas and consumers)
#   G hosted adapter (G1-G7i: hosted-context.sh in the pinned docker CLI image, a docker shim, a tmpfs build
#     directory; the synthetic job credentials travel only in a mode-600 environment file), and the G004
#     identity-staging forwarding gate (C-SEC-1..10: forwarding only in the approved context; refusals,
#     the scope-misconfiguration detector, missing inputs and publication of the staging key)
#   H the release helper (ONEOFF RC: infra/ci/release-verify.sh name R1, candidate R2a-R2i, qualified R3a-R3q,
#     promotable R4a-R4c, recheck R5a-R5f, promoted R6a-R6e, final R7a-R7f, merged evidence R8a-R8m) on the RC-promotion
#     fixture pipeline, its repository and bare remote, with the GitLab API JSON as data files
# Every control runs in a disposable scratch tree outside the checkout (git init trees holding copies of
# validate.sh, lib.sh, junit.sh, cleanup-owned.sh, required-checks.tsv, unchanged.allow and gitleaks.toml, a
# declared synthetic contract corpus, a stub infra/validation/run.sh shaped like the real harness and a
# stub runner-preflight.sh). Faults are injected only through the stubs, a docker shim on PATH, file
# modes, GitLab-defined variables and signals sent from outside; validate.sh has no fault switch. V4 runs
# the real infra/validation/run.sh contracts suite on a scratch copy of this checkout with EI39 set to
# valid. Group E uses busybox:1.37.0; every control resource carries sdlc.ci.control=<token> and is removed
# and verified by its control. The real checkout is never written. As root (the hosted context), validate.sh
# launches drop CAP_DAC_OVERRIDE and CAP_DAC_READ_SEARCH (setpriv), so the file-mode faults (V6d, V17, V17b)
# behave as for the unprivileged local user.
#
# Nested launches (validate.sh, the real contracts run, a nested run of this script) start through
# launch(): each its own process group, recorded in the scratch registry before it runs, waited for with a
# bound. On TERM, INT or HUP, at a launch's bound, or when a launch outlives its control, the runner stops
# exactly its own tree (the running control, the registered groups, their descendants and the groups
# those lead): TERM once, a 30 s grace, KILL, reaped, and only then are resources swept and scratch
# removed. The exit status then is 128+n for signal n (a FAIL line names the cancelled control); a bound
# reached fails that control. Nothing is matched by name or pattern. The fixture nested-hold (only by
# --only; never in a group or the --list) is the nested active validator of V18l-V18n.
#
#   bash infra/ci/validate-controls.sh [--group A,B,...] [--only ID,ID,...] [--extra-label KEY=VALUE]...
#                                      [--log FILE] [--list] [--keep-scratch] [--launch-deadline SECONDS]
#
# --keep-scratch keeps the scratch root (cases and per-control logs) for diagnosis; Docker resources are
# always removed. --launch-deadline caps every nested launch's bound (used by V18n). A cleanup that is not verified
# keeps the scratch root anyway (its inventory is the recovery record; see owned_ids) and starts no later control.
# Output: one "PASS <id> <detail>" or "FAIL <id> <detail>" line per selected control (after a held cleanup, one
# "FAIL not-started" line naming the controls not started instead of theirs), then a summary. With
# --log, ci-controls-timings.tsv (ID<TAB>seconds, one line per finished control; G051) is written beside it.
# Before anything is allocated, the startup guard (G064 correction 11 item 4; see "startup guard" below) holds the run
# when a control resource survives whose owner has ended or cannot be attributed.
# A top-level run (no CI_CONTROLS_OWNED_REG from run.sh, not nested, not the CI_CONTROLS_HOLD_SIM simulation) registers
# itself for durable recovery first, as a local validate.sh run does (G064 C4; see "self-registration" below).
# Exit: 0 when every selected control passed; 1 when any failed; 2 on a usage error; 16 when the startup guard
# holds or the self-registration fails (no control ran, nothing was allocated, removed or stopped); 128+n when cancelled
# by signal n.
set -euo pipefail
shopt -s inherit_errexit

SELF_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)"
REPO="$(dirname -- "$(dirname -- "$SELF_DIR")")"
REPO_R="$(realpath -e -- "$REPO")"
# shellcheck source=lib.sh
source "$SELF_DIR/lib.sh"
# The controls' own git operations ignore the caller's git configuration (no signing, hooks or redirects).
export GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1
unset GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE

usage() { echo "usage: validate-controls.sh [--group A,B,C,D,E,F,G,H] [--only ID,...] [--extra-label KEY=VALUE]... [--log FILE] [--list] [--keep-scratch] [--launch-deadline SECONDS]" >&2; exit 2; }
SEL_GROUPS="" SEL_ONLY="" LOG="" LIST=0 KEEP=0 EXTRA_LABELS=() LAUNCH_CAP="" ORIG_ARGS=("$@")
while (( $# > 0 )); do
  case "$1" in
    --group) [[ $# -ge 2 ]] || usage; SEL_GROUPS="$2"; shift 2 ;;
    --only) [[ $# -ge 2 ]] || usage; SEL_ONLY="$2"; shift 2 ;;
    --extra-label)
      [[ $# -ge 2 && "$2" =~ ^[A-Za-z0-9._-]{1,64}=[A-Za-z0-9._-]{1,64}$ ]] || usage
      EXTRA_LABELS+=("$2"); shift 2 ;;
    --log) [[ $# -ge 2 ]] || usage; LOG="$2"; shift 2 ;;
    --list) LIST=1; shift ;;
    --keep-scratch) KEEP=1; shift ;;
    --launch-deadline) [[ $# -ge 2 && "$2" =~ ^[1-9][0-9]{0,4}$ ]] || usage; LAUNCH_CAP="$2"; shift 2 ;;
    *) usage ;;
  esac
done

CONTROLS=(
  "A V1" "A V2" "A V3" "A V4" "A V7" "A V8" "A V9" "A V10" "A V11" "A V12"
  "A V13a" "A V13b" "A V13c" "A V13d" "A V13e" "A V13f"
  "A VP1" "A VP2" "A VP3" "A VP4" "A VP5" "A VP6" "A VP7" "A VP8" "A VP9"
  "A VA1" "A VA2" "A VA3" "A VA4" "A VA4b" "A VA5" "A VA6" "A VA7" "A VA8" "A VA9" "A VA10" "A VA11"
  "A VU1" "A VU2" "A VU3" "A VU4" "A VU5" "A VU6" "A VU7" "A VU8" "A VU9" "A VU10" "A VU11" "A VU12"
  "B P1" "B P2" "B P3" "B P3F" "B P4" "B P5" "B P6" "B P7" "B P8" "B P9" "B P10" "B P11" "B P11F"
  "B N40" "B N41" "B N42" "B N43" "B N44" "B N45" "B N46" "B N47" "B N48" "B N49" "B RR1" "B RR2"
  "B N1" "B N2" "B N3" "B N4" "B N5" "B N6" "B N7" "B N8"
  "B N9" "B N10" "B N11" "B N12" "B N13" "B N14" "B N15" "B N16" "B N17" "B N18"
  "B N19" "B N20" "B N21" "B N22" "B N23" "B N12q"
  "C V5a" "C V5b-planted" "C V5b-clean" "C V5c" "C V5d" "C V6a" "C V6b" "C V6c" "C V6d"
  "D V14" "D V15" "D V16" "D V16b" "D V17" "D V17b" "D V18" "D V18a" "D V18b" "D V18c" "D V18d" "D V18e"
  "D V18f" "D V18g" "D V18h" "D V18i" "D V18j" "D V18k" "D V18l" "D V18m" "D V18n" "D V36" "D V37"
  "D J1" "D J2" "D J3" "D J4a" "D J4b" "D J4c" "D J5" "D J6" "D J7" "D J8" "D GT1" "D GT2"
  "D LR1" "D LR2" "D LR3"
  "D VB1" "D VB2" "D VB3" "D VB3b" "D VB4" "D VB5" "D VB6" "D VB7" "D VB8"
  "E V19" "E V20c" "E V20n" "E V20v" "E V20i" "E V21" "E V22" "E V23" "E V24" "E V25" "E V26" "E V27"
  "E V27b" "E V28" "E V29" "E V30" "E V31" "E V31b" "E V32" "E V33" "E V34" "E V35"
  "E VC1" "E VC2" "E VC3" "E VC4" "E VC5" "E VC6" "E VC7" "E VC8" "E VC9"
  "E VS1" "E VS2" "E VS3" "E VS4" "E VS5" "E VR1" "E VR2" "E VR3" "E VR4" "E VR5" "E VR6" "E VR7" "E VR8"
  "E VR9"
  "F F-web" "F F-core" "F F-intelligence" "F F-contracts" "F F-compose"
  "F F-lock-gomod" "F F-lock-gosum" "F F-lock-pkg" "F F-lock-pkglock" "F F-lock-pyproject" "F F-lock-uvlock"
  "F F-cfg-next" "F F-cfg-dockerfile" "F F-ci-yml" "F F-ci-map" "F F-ci-controls"
  "F F-oracle-browser" "F F-oracle-fixtures" "F F-oracle-conformance" "F F-doc-readme" "F F-doc-foundation"
  "F F-unknown" "F F-delete" "F F-rename" "F F-mixed" "F F-unsafe" "F F-nochange"
  "F F-gitignore-guards" "F F-route-guards" "F F-gitignore" "F F-gitignore-runtime" "F F-gitignore-foundation" "F F-gitignore-delete"
  "F F-session-tracking" "F F-session-tracking-mixed" "F F-session-tracking-near" "F F-session-tracking-moved"
  "F F-status-progress" "F F-status-backlog" "F F-status-history" "F F-mix-spec" "F F-mix-backlog-web"
  "F F-gov-architecture" "F F-gov-security" "F F-gov-decisions" "F F-gov-adr" "F F-gov-handoff" "F F-gov-agents"
  "F F-gov-claude" "F F-gov-goal" "F F-gov-spec" "F F-gov-plan" "F F-gov-allowlist" "F F-gov-roadmap"
  "F F-gov-budget" "F F-gov-newdoc" "F F-gov-newmgmt"
  "F F-review-doc"
  "F F-review-web"
  "F F-review-lock"
  "F F-management-script"
  "F F-management-unknown"
  "F F-management-delete"
  "F F-management-rename"
  "F F-management-outside"
  "F F-doc-script"
  "F F-doc-data"
  "F F-doc-root-control"
  "F F-doc-skill"
  "F F-doc-mixed-runtime"
  "F F-doc-mixed-lock"
  "F F-doc-rename"
  "F F-doc-outside"
  "F F-doc-approval"
  "F F-base-missing" "F F-base-zero" "F F-base-absent" "F F-base-nonancestor" "F F-base-degenerate"
  "F F-feature-new" "F F-push-main" "F F-push-staging" "F F-fast-job" "F F-schedule" "F F-web-intent" "F F-api-intent" "F F-local"
  "F F-rc-promotion" "F F-area-rc"
  "F F-schedule-qualification" "F F-schedule-other"
  "F F-merged-docs" "F F-merged-target" "F F-merged-both-docs" "F F-map-dup" "F F-map-class" "F F-map-abs"
  "F F-diff-fail" "F F-migrate-dep" "F F-config-dep" "F F-tenancy"
  "F F-hand-skill" "F F-hand-meta" "F F-hand-runbook" "F F-hand-docs" "F F-hand-del-skill" "F F-hand-del-meta"
  "F F-hand-del-runbook" "F F-hand-swap" "F F-hand-to-doc" "F F-hand-to-agent" "F F-hand-to-skills" "F F-hand-helper"
  "F F-agent-settings" "F F-agent-claude-skill" "F F-agent-agents-skill" "F F-agent-unknown" "F F-agent-delete"
  "F F-agent-rename" "F F-agent-to-skills" "F F-nested-web" "F F-nested-core" "F F-nested-docs" "F F-vestive-css"
  "F F-vestive-kit" "F F-vestive-md" "F F-vestive-link" "F F-web-agent" "F F-dist-restruct"
  "F F-agent-mixed-runtime" "F F-agent-mixed-lock" "F F-hand-mixed-runtime" "F F-hand-mixed-nested"
  "F F-agent-push-staging" "F F-agent-push-main" "F F-hand-push-staging" "F F-nested-push-main"
  "F F-agent-mixed-push-staging" "F F-hand-web" "F F-hand-api" "F F-hand-schedule"
  "F F-widen-identity" "F F-widen-server" "F F-widen-auth" "F F-widen-sign-in" "F F-widen-account" "F F-widen-account-api"
  "F F-widen-account-ui" "F F-widen-security" "F F-widen-globals" "F F-widen-layout" "F F-widen-public" "F F-widen-layout-ui"
  "F F-widen-ui" "F F-widen-lib" "F F-widen-unit-tests" "F F-widen-mcp" "F F-widen-streaming"
  "F F-keep-health" "F F-keep-api-health" "F F-keep-api-ready" "F F-keep-models" "F F-keep-providers"
  "F F-area-page-spec" "F F-area-page-spec-gate" "F F-area-test-only" "F F-area-fixture-only" "F F-area-core-it"
  "F F-area-core-unit" "F F-area-intel" "F F-area-cross-stack" "F F-area-two-web"
  "F F-mix-area-docs" "F F-mix-area-agent" "F F-mix-area-runtime" "F F-mix-area-lock"
  "F F-unreg-core-pkg" "F F-unreg-feature-test"
  "F F-map-area-root" "F F-map-area-name" "F F-map-area-overlap" "F F-map-area-kind" "F F-map-area-dup" "F F-map-in-repo"
  "F F-area-push-staging" "F F-area-web" "F F-area-api" "F F-area-schedule" "F F-area-merged" "F F-area-feature-push"
  "F F-area-base-missing" "F F-area-local-base" "F F-area-local-no-base" "F F-area-local-base-head" "F F-area-local-compare"
  "F F-area-local-compare-same" "F F-area-local-compare-unknown" "F F-area-ci-compare" "F F-area-ci-base"
  "F F-area-local-refusals" "F F-gate-full-remainder" "F F-gate-integration-remainder" "F F-digest"
  "F F-dep-go-subpkg" "F F-dep-go-subpkg-area" "F F-dep-go-test-only" "F F-dep-go-testdata-path" "F F-dep-web-alias"
  "F F-dep-web-relative" "F F-dep-web-relative-area" "F F-dep-web-css" "F F-dep-web-dynamic" "F F-dep-web-template"
  "F F-dep-web-escape" "F F-dep-py-submodule" "F F-dep-py-relative" "F F-dep-py-relative-unresolved" "F F-dep-py-dynamic"
  "F F-dep-py-dynamic-webarea" "F F-dep-removed-edge" "F F-dep-removed-foundation-consumer" "F F-dep-renamed-area"
  "F F-dep-transitive" "F F-dep-docs-consumer" "F F-dep-harness-route" "F F-fixture-other-area" "F F-fixture-inherited" "F F-feature-test-only"
  "F F-area-unclassified-script" "F F-area-unclassified-gomod" "F F-area-unclassified-config"
  "F F-area-unclassified-pyproject" "F F-area-symlink" "F F-area-route-handler" "F F-closure-base-unreadable"
  "F F-registered-absent" "F F-area-removed-clean" "F F-area-removed-consumer"
  "G G1" "G G2" "G G3" "G G4" "G G5" "G G6a" "G G6b" "G G6c" "G G6d" "G G6e" "G G6f" "G G6g" "G G6h"
  "G G7a" "G G7b" "G G7c" "G G7d" "G G7e" "G G7f" "G G7g" "G G7h" "G G7i"
  "G C-SEC-1" "G C-SEC-2" "G C-SEC-3" "G C-SEC-4" "G C-SEC-5" "G C-SEC-6" "G C-SEC-7" "G C-SEC-8" "G C-SEC-9" "G C-SEC-10"
  "H R1" "H R2a" "H R2b" "H R2c" "H R2d" "H R2e" "H R2f" "H R2g" "H R2h" "H R2i"
  "H R3a" "H R3b" "H R3c" "H R3d" "H R3e" "H R3f" "H R3g" "H R3h" "H R3i" "H R3j" "H R3k" "H R3l" "H R3m" "H R3n" "H R3o" "H R3p" "H R3q"
  "H R4a" "H R4b" "H R4c" "H R5a" "H R5b" "H R5c" "H R5d" "H R5e" "H R5f" "H R6a" "H R6b" "H R6c" "H R6d" "H R6e"
  "H R7a" "H R7b" "H R7c" "H R7d" "H R7e" "H R7f"
  "H R8a" "H R8b" "H R8c" "H R8d" "H R8e" "H R8f" "H R8g" "H R8h" "H R8i" "H R8j" "H R8k" "H R8l" "H R8m"
)
FIXTURES=("Z nested-hold" "Z obj-codes" "Z reg-hold")
if (( LIST )); then for c in "${CONTROLS[@]}"; do printf '%s\n' "$c"; done; exit 0; fi
# G064 Plan A corr1 C4: a top-level run (no run registry passed, not a nested run, not the CI_CONTROLS_HOLD_SIM simulation)
# registers itself for durable recovery (see "self-registration" below). Its recorded process group must be exactly its
# own tree, so a run that is not its own process group leader re-executes once under setsid (same PID: setsid does not
# fork a non-leader), as validate.sh launches its harness.
SELF_REG=0
if [[ -z "${CI_CONTROLS_OWNED_REG:-}" && -z "${CI_CONTROLS_OUTER_OWNED:-}" && -z "${CI_CONTROLS_HOLD_SIM:-}" ]]; then SELF_REG=1; fi
if (( SELF_REG )); then
  sr_pg="$(s="$(cat "/proc/$$/stat" 2>/dev/null)" && read -r -a a <<<"${s##*) }" && printf '%s' "${a[2]}")" || sr_pg=""
  if [[ "$sr_pg" != "$$" ]]; then
    if [[ -n "${CI_CONTROLS_SETSID:-}" ]]; then echo "validate-controls: not a process group leader after setsid (group [$sr_pg]); refused" >&2; exit 2; fi
    exec env CI_CONTROLS_SETSID=1 setsid bash "$SELF_DIR/validate-controls.sh" "${ORIG_ARGS[@]}"
  fi
fi
unset CI_CONTROLS_SETSID

selected() { # GROUP ID
  local g
  if [[ -n "$SEL_ONLY" && ",$SEL_ONLY," != *",$2,"* ]]; then return 1; fi
  if [[ -n "$SEL_GROUPS" ]]; then
    for g in ${SEL_GROUPS//,/ }; do if [[ "$g" == "$1" ]]; then return 0; fi; done
    return 1
  fi
  return 0
}
SELECTED=()
for c in "${CONTROLS[@]}"; do if selected "${c%% *}" "${c#* }"; then SELECTED+=("$c"); fi; done
# A fixture runs only when --only names it.
for c in "${FIXTURES[@]}"; do if [[ ",$SEL_ONLY," == *",${c#* },"* ]]; then SELECTED+=("$c"); fi; done
if (( ${#SELECTED[@]} == 0 )); then echo "validate-controls: the selection matches no control" >&2; exit 2; fi
if [[ -n "$LOG" ]]; then : >"$LOG" || usage; fi
# G051: per-control timing (ID<TAB>seconds) beside the log; the PASS/FAIL lines are unchanged.
TIMES=""
if [[ -n "$LOG" ]]; then TIMES="$(dirname -- "$LOG")/ci-controls-timings.tsv"; : >"$TIMES" || usage; fi
out() { printf '%s\n' "$*"; if [[ -n "$LOG" ]]; then printf '%s\n' "$*" >>"$LOG"; fi; }

REAL_DOCKER="$(command -v docker)" || { out "FAIL setup docker CLI not found"; exit 1; }
# Docker Desktop on WSL: refuse to start when leaked bind mounts have nearly filled its mount table.
bash "$REPO_R/infra/ci/docker-desktop-mounts.sh" >&2 || { out "FAIL setup Docker Desktop mount table near its limit"; exit 1; }
RT="$(ci_random_hex 6)"
# --- startup guard (G064 correction 11 item 4: containment only, not startup recovery) ---------------------------
# OWNER is the sdlc.ci.owner key (infra/validation/lib.sh g003_owner_key: boot id, PID namespace, PID, start time,
# checkout id, run id) that every resource of this run carries from its creation (ctl_labels, case_labels, mk_image,
# the nested-hold simulation). A top-level runner is its own controlling owner. A nested run of this script (V18l-V18n:
# nested_cancel passes CI_CONTROLS_OWNER) carries its CONTROLLING runner's key, never its own: the controlling runner
# stops the nested run and removes whatever it left (owned_cleanup), so the nested resources stay attributed to a live
# owner exactly as long as the process responsible for removing them lives, and a nested runner cancelled on purpose
# never makes them look abandoned to a concurrent start. The nested run's own identity stays in its own
# sdlc.ci.run and sdlc.ci.control labels and in sdlc.ci.outer; the parent's key is only reused, never replaced. The
# nested run uses that key only after proving it names a live ancestor of the same boot, PID namespace and checkout;
# otherwise it holds. g003_owner_guard then runs here, before the scratch root, the template and every control
# (that is, before this runner's first Docker or scratch allocation) for every caller of this script: run.sh's
# ci-controls suite (B-CI-CONTROLS), direct runs and its own nested runs. It does not protect any earlier stage of
# infra/ci/validate.sh or run.sh, which never invoke it. Objects of this controlling run and of other live owners are
# ignored and untouched; a surviving g003 object whose owner has ended or cannot be attributed, or a listing Docker
# refuses, holds the whole run (exit 16). It removes, stops and signals nothing.
G3LIB="$REPO_R/infra/validation/lib.sh"
# shellcheck disable=SC2016  # the helper shell's own script
g3() { bash -c 'set -euo pipefail; source "$1"; shift; "$@"' g3 "$G3LIB" "$@"; } # a lib.sh helper, in its own shell
guard_hold() { # LINES: the first as the FAIL line, the rest as given, the summary; exit 16 (nothing allocated yet)
  local l
  { IFS= read -r l; out "FAIL setup $l"; while IFS= read -r l; do out "$l"; done; } <<<"$1"
  out "ci-controls: startup hold (exit 16); no control ran (of ${#SELECTED[@]} selected; ${#CONTROLS[@]} defined)"
  exit 16
}
if [[ -n "${CI_CONTROLS_OWNER:-}" ]]; then
  if g3 g003_owner_ancestor "$CI_CONTROLS_OWNER" "$$" "$REPO_R"; then OWNER="$CI_CONTROLS_OWNER"; else
    guard_hold "startup guard: hold: CI_CONTROLS_OWNER does not name a live controlling ancestor of this run (same boot, PID namespace and checkout); nothing was allocated"
  fi
elif OWNER="$(g3 g003_owner_key "$$" "$RT" "$REPO_R")"; then :; else
  guard_hold "startup guard: hold: this run's owner identity (boot id, PID namespace, /proc/$$/stat start time, checkout) is unavailable, so its resources could not be attributed; nothing was allocated"
fi
OWNER_LABEL="sdlc.ci.owner=$OWNER"
GUARD_RC=0
GUARD_OUT="$(g3 g003_owner_guard "$REAL_DOCKER" "$OWNER")" || GUARD_RC=$?
if (( GUARD_RC != 0 )); then
  if [[ "$GUARD_OUT" != "startup guard: hold: "* ]]; then GUARD_OUT="startup guard: hold: the guard ended with exit $GUARD_RC before it verified every owner; nothing was allocated, removed or stopped"$'\n'"$GUARD_OUT"; fi
  guard_hold "$GUARD_OUT"
fi
out "$GUARD_OUT"
# --- self-registration (G064 Plan A corr1 C4, corr2 D3) ------------------------------------------------------------
# A top-level run (SELF_REG, above) gets the durable recovery registration of a local validate.sh run before it allocates
# anything: a run registry (validate.sh make_registry: header, job local-<hex>, token, repo) in a private directory under
# the persistent state root (validate.sh state_root_prepare: G003_STATE_ROOT, else the XDG state home), its start inventory
# (cleanup-owned.sh refuses a registry with resource rows but none), and the pending marker STATE_ROOT/g003-pending/<job>.tsv
# (validate.sh marker_write/marker_stage: boot_id, repo; stage launched with this run's process group and itself as the
# recorded harness group and watchdog). The functions are validate.sh's own text, extracted, so there is no second writer.
# Every inventory row is mirrored into this registry (REG_MIRROR, below). On exit the registry is finalized through
# cleanup-owned.sh; only a verified finalize removes it and the marker (marker_finish). Otherwise both stay, and the next
# local validate.sh launch recovers the run (do_recovery: dead creator group, cleanup-owned.sh) or blocks with 16.
# A failure here is exit 16 before anything is allocated.
SR_ON=0 SR_REG="" SR_EV=""
if (( SELF_REG )); then
  sr_fail() { out "FAIL setup self-registration: $1; no control ran, nothing was allocated"; exit 16; }
  # shellcheck disable=SC2034  # read by the extracted validate.sh functions
  CI_MODE=0 REG_CREATED=0 REG_PROVEN=0 REG_JOB="" REG="" EV="" PEND_DIR="" PEND_FILE="" MARKER_BASE="" STATE_ROOT="" REGSTATE=none
  # shellcheck disable=SC2034  # STARTED_UTC: read by the extracted marker_write (the marker's `started` field)
  STARTED_UTC="$(ci_utc)"
  declare -A ST=()
  say() { out "validate-controls: self-registration: $*"; }
  inside_repo() { ci_path_within "$1" "$REPO_R"; }
  sr_defs="$(awk -v want=" make_registry volatile_fs state_root_prepare proc_start_time priv_dir_ok pend_dir_ok boot_id_now marker_write marker_stage marker_finish " '
    !on && match($0, /^[a-z_]+\(\) \{/) { n = substr($0, 1, index($0, "(") - 1); if (index(want, " " n " ")) { on = 1; got = got " " n } }
    on { print } on && $0 == "}" { on = 0 }
    END { print "SR_GOT=\"" got "\"" }' "$REPO_R/infra/ci/validate.sh")" || sr_fail "validate.sh could not be read"
  eval "$sr_defs"
  for f in make_registry volatile_fs state_root_prepare proc_start_time priv_dir_ok pend_dir_ok boot_id_now marker_write marker_stage marker_finish; do
    if [[ " $SR_GOT " == *" $f "* ]] && declare -F "$f" >/dev/null; then :; else sr_fail "validate.sh defines no $f"; fi
  done
  sr_prev="${G003_STATE_ROOT-}" sr_had="${G003_STATE_ROOT+x}"
  state_root_prepare || sr_fail "no private persistent state root"
  if [[ -n "$sr_had" ]]; then G003_STATE_ROOT="$sr_prev"; else unset G003_STATE_ROOT; fi   # never leaks into launches
  EV="$(mktemp -d "$STATE_ROOT/g003-controls-XXXXXX" 2>/dev/null)" || sr_fail "cannot create a registry directory under $STATE_ROOT"
  chmod 700 -- "$EV" || sr_fail "cannot make $EV private"
  REG="$EV/owned" PEND_DIR="$STATE_ROOT/g003-pending"
  SR_EV="$EV" SR_REG="$REG" SR_ON=1
  sr_finalize() { # 0: the registry was finalized verified clean and removed with the marker; 1: both kept (recovery record)
    local rc=0 inv="$SR_REG/cleanup-inventory.tsv"
    if (( SR_ON == 0 )); then return 0; fi
    SR_ON=0
    if [[ -f "$SR_REG/owned.tsv" ]]; then
      bash "$REPO_R/infra/ci/cleanup-owned.sh" --registry "$SR_REG" --mode finalize --out "$inv" >"$SR_REG/cleanup.log" 2>&1 || rc=$?
      if (( rc == 0 )) && [[ -f "$inv" && ! -L "$inv" ]] && ! awk -F'\t' 'NF != 4 || $3 != "removed" { f = 1 } END { exit f ? 0 : 1 }' "$inv"; then
        REGSTATE=removed
      else REGSTATE=kept; fi
    else REGSTATE=removed; fi   # make_registry wrote nothing: nothing can be owned
    # shellcheck disable=SC2034  # ST[cleanup]: read by the extracted marker_finish (the marker's `finalized cleanup=` row)
    ST[cleanup]="$rc"
    if [[ "$REGSTATE" == removed ]]; then rm -rf -- "$SR_EV" 2>/dev/null || true; if [[ -e "$SR_EV" ]]; then REGSTATE=kept; fi; fi
    marker_finish
    [[ "$REGSTATE" == removed ]]
  }
  sr_early_exit() { local rc=$?; sr_finalize || out "FAIL cleanup self-registration: run registry $SR_REG kept with its marker for recovery"; exit "$rc"; }
  trap sr_early_exit EXIT
  make_registry || sr_fail "the run registry could not be created"
  pend_dir_ok || sr_fail "the marker directory $PEND_DIR is not a private directory of this user on persistent storage"
  marker_write || sr_fail "the pending marker could not be written"
  sr_st="$(proc_start_time "$$")" || sr_fail "this run's start time is unreadable"
  marker_stage launched harness_pgid "$$" harness_start "$sr_st" watchdog_pid "$$" watchdog_start "$sr_st" || sr_fail "the marker could not record this run"
  ci_docker_start_inventory "$SR_REG/start-inventory.ids" || sr_fail "the Docker start inventory failed"
  out "validate-controls: self-registered as local run $REG_JOB (registry $SR_REG; marker $PEND_FILE)"
fi
ROOT="$(mktemp -d "${TMPDIR:-/tmp}/ci-controls-XXXXXX")"
ROOT="$(realpath -e -- "$ROOT")"
if ci_path_within "$ROOT" "$REPO_R"; then echo "validate-controls: the scratch root must lie outside the repository" >&2; exit 2; fi
chmod 700 "$ROOT"
mkdir -p "$ROOT/shim" "$ROOT/cases" "$ROOT/logs" "$ROOT/problems" "$ROOT/details" "$ROOT/owned"
RUN_LABEL="sdlc.ci.run=$RT"
# G064 corr2 §D3: a case-tree validate.sh never sees the operator's state root. BASE_ENV names a root that cannot be
# created (/nonexistent/g003-controls: a local launch without its own root fails location 16), and every validate.sh
# launch of a case gets G003_STATE_ROOT=<case state> (case_state): $C/state, mode 700, on the scratch root. validate.sh
# refuses a root on tmpfs/ramfs, so when the scratch root lies there the case roots go under a run-scoped directory on
# persistent storage instead (CI_CONTROLS_STATE_BASE, else $XDG_STATE_HOME or ~/.local/state; never the operator's
# platform-evidence/g003-local root), removed with the scratch root.
STATE_BASE=""
vc_volatile() { local t; t="$(stat -f -c %T -- "$1" 2>/dev/null)" || return 0; [[ "$t" == tmpfs || "$t" == ramfs ]]; }
if vc_volatile "$ROOT"; then
  sb="${CI_CONTROLS_STATE_BASE:-${XDG_STATE_HOME:-${HOME:-/nonexistent}/.local/state}}"
  if (umask 077 && mkdir -p -- "$sb") 2>/dev/null && STATE_BASE="$(mktemp -d "$sb/ci-controls-state-XXXXXX" 2>/dev/null)" &&
    STATE_BASE="$(realpath -e -- "$STATE_BASE")" && chmod 700 "$STATE_BASE" && ! vc_volatile "$STATE_BASE" &&
    ! ci_path_within "$STATE_BASE" "$REPO_R"; then
    echo "validate-controls: the scratch root $ROOT lies on tmpfs/ramfs; case state roots go under $STATE_BASE" >&2
  else
    rm -rf -- "${STATE_BASE:-/nonexistent/none}" 2>/dev/null || true
    echo "validate-controls: the scratch root lies on tmpfs/ramfs and no persistent case state directory could be made under $sb" >&2
    rm -rf -- "$ROOT"; exit 2
  fi
fi
CSTATE=""
case_state() { # -> CSTATE: this case's private state root (created mode 700 when absent; a fresh case gets a fresh one)
  if [[ -z "$STATE_BASE" ]]; then CSTATE="$C/state"; if [[ -d "$CSTATE" ]]; then :; else mkdir -m 700 -- "$CSTATE"; fi; return 0; fi
  if [[ -f "$C/.state-path" ]]; then CSTATE="$(cat "$C/.state-path")"; else
    CSTATE="$(mktemp -d "$STATE_BASE/$CTL_ID-XXXXXX")"; chmod 700 "$CSTATE"; printf '%s' "$CSTATE" >"$C/.state-path"
  fi
}

# --- enforced resource ceilings and complete per-control accounting (G064 Plan A corr1 C1 P6, C2, C4; corr2 D2) ------
# Every control keeps $ROOT/owned/<control>.tsv (the runner starts it before the control). The two allocation paths, D()
# at control level and the docker shim for every case launch, run gate_reserve before each creation: under flock on
# <inventory>.lock it reads the control latch (<inventory>.hold) and the run latch (hold.d/run beside the run registry),
# counts the live objects of that class carrying the scope label plus the live reservations without a created object,
# and refuses (latching, appending a `ceiling` row) when one more would cross the ceiling; a query failure refuses and
# latches too. Ceilings: an ordinary control (label sdlc.ci.control=<run>-<control>) 8 containers, 2 networks, 2 volumes,
# 2 images; the V18l-n nested holder (label sdlc.ci.outer=<run>-<control>, inventory CI_CONTROLS_OUTER_OWNED) 39/0/0/0.
# CI_CONTROLS_CEILING_LOWER=<control>:<n> lowers every ceiling of that control's scopes to at most n (proofs only); a
# value above the smallest container ceiling (8) would raise one and is refused. A container creation always carries a
# cidfile (the caller's, else one under <inventory>.cid/), recorded as a `cidfile` row before the call, so an attached
# or interrupted creation is still attributed; created IDs (network/volume create output, image IDs after build/commit)
# are recorded once each and mirrored into the run registry (CI_CONTROLS_OWNED_REG, else this run's self-registration)
# as existing row kinds: standalone <id> ci-<control>, network <id>, volume <name>, image <tag> <id>. Removal is never
# gated. This file is sourced here and by the shim.
cat >"$ROOT/shim/gate.sh" <<'GATE'
# gate.sh (infra/ci/validate-controls.sh): ceiling check-and-reserve, records and registry mirrors. Sourced; no output
# on success except what the caller prints. GATE_RSV: the reservation; GATE_MSG: why a creation was refused.
GATE_RSV="" GATE_MSG=""
gate_ceil() { # SCOPE(control|outer) CTL KIND -> the ceiling
  local m
  case "$1:$3" in control:container) m=8 ;; control:network|control:volume|control:image) m=2 ;; outer:container) m=39 ;; *) m=0 ;; esac
  # The build-ledger controls (VC1-VC9) build up to ~16 labelled images in one control (a rebuilt tag leaves its old ID dangling
  # with the label, and every live ID counts): a fixed, conservative cap of 24 for that finite fixture family only.
  if [[ "$1:$3" == control:image && "$2" =~ ^VC[1-9]$ ]]; then m=24; fi
  if [[ "${CI_CONTROLS_CEILING_LOWER:-}" =~ ^([A-Za-z0-9._-]+):([0-9]+)$ && "${BASH_REMATCH[1]}" == "$2" ]] && (( BASH_REMATCH[2] < m )); then m="${BASH_REMATCH[2]}"; fi
  printf '%s' "$m"
}
gate_alive() { local s a; read -r s 2>/dev/null <"/proc/$1/stat" || return 1; read -r -a a <<<"${s##*) }"; [[ "${a[0]}" != Z && "${a[19]}" == "$2" ]]; }
gate_count() { # DOCKER KIND LABEL -> the number of live objects of KIND carrying LABEL; 1 when Docker refuses the listing
  local o
  case "$2" in
    container) o="$("$1" ps -aq --no-trunc --filter "label=$3")" || return 1 ;;
    network) o="$("$1" network ls -q --no-trunc --filter "label=$3")" || return 1 ;;
    volume) o="$("$1" volume ls -q --filter "label=$3")" || return 1 ;;
    image) o="$("$1" image ls -aq --no-trunc --filter "label=$3")" || return 1; o="$(LC_ALL=C sort -u <<<"$o")" ;;
    *) return 1 ;;
  esac
  awk 'NF { n++ } END { print n + 0 }' <<<"$o"
}
gate_latch() { # BASE REASON: the control latch (O_EXCL create; a later reason appends one line)
  (umask 077; set -C; printf 'scope\t%s\nreason\t%s\n' "${1##*/}" "$2" >"$1.hold") 2>/dev/null || printf 'reason\t%s\n' "$2" >>"$1.hold"
}
gate_reserve() { # INV DOCKER KIND LABEL MAX OWNER_PID HOLD_RUN [CIDFILE]: 0 with GATE_RSV, or 1 with GATE_MSG (nothing created)
  local inv="$1" d="$2" k="$3" l="$4" max="$5" own="$6" hr="$7" cf="${8:-}" base n f p s c os
  base="${inv%.tsv}" GATE_RSV="" GATE_MSG=""
  if [[ ! -f "$inv" ]]; then GATE_MSG="refused: the inventory $inv of this scope does not exist"; return 1; fi
  mkdir -p -m 700 -- "$base.rsv" 2>/dev/null
  {
    if flock -w 30 9; then :; else gate_latch "$base" "lock timeout"; GATE_MSG="held: resource ceiling lock timeout for ${base##*/}"; return 1; fi
    if [[ -e "$base.hold" ]]; then GATE_MSG="held: resource ceiling latch $base.hold ($(awk -F'\t' '$1 == "reason" {printf "%s%s", s, $2; s = "; "}' "$base.hold" 2>/dev/null))"; return 1; fi
    if [[ -n "$hr" && -e "$hr" ]]; then GATE_MSG="held: resource ceiling run latch $hr"; return 1; fi
    if n="$(gate_count "$d" "$k" "$l")"; then :; else gate_latch "$base" "$k query failed"; GATE_MSG="held: resource ceiling: the $k listing failed for ${base##*/}"; return 1; fi
    for f in "$base.rsv"/*; do
      if [[ -f "$f" ]]; then :; else continue; fi
      read -r p s c <"$f" 2>/dev/null || { rm -f -- "$f"; continue; }
      if gate_alive "$p" "$s"; then :; else rm -f -- "$f"; continue; fi
      if [[ -n "$c" && -s "$c" ]]; then continue; fi    # created: already counted as a live object
      n=$((n + 1))
    done
    if (( n + 1 > max )); then
      gate_latch "$base" "$k ceiling $n/$max"
      printf 'ceiling\t%s\t%s/%s\n' "$k" "$n" "$max" >>"$inv"
      GATE_MSG="resource ceiling $n/$max for ${base##*/} ($k; label $l)"; return 1
    fi
    os=""; if read -r os 2>/dev/null <"/proc/$own/stat"; then read -r -a s <<<"${os##*) }"; os="${s[19]:-}"; fi
    f="$base.rsv/$own-$RANDOM$RANDOM"
    (umask 077; set -C; printf '%s %s %s\n' "$own" "$os" "$cf" >"$f") || { GATE_MSG="refused: the reservation could not be written"; return 1; }
    if [[ -n "$cf" ]]; then printf 'cidfile\t%s\t%s\n' "$cf" "$k" >>"$inv"; fi
    GATE_RSV="$f"
  } 9>>"$base.lock"
  return 0
}
gate_release() { if [[ -n "$GATE_RSV" ]]; then rm -f -- "$GATE_RSV"; fi; GATE_RSV=""; }
gate_record() { # INV KIND ID WHAT REG [TAG]: one inventory row per ID; the registry mirror (validated row formats only)
  local inv="$1" k="$2" id="$3" what="$4" reg="$5" tag="${6:-}" c row=""
  c="${inv##*/}"; c="${c%.tsv}"
  {
    flock -w 30 9 || true
    if grep -qF -- "$(printf '%s\t%s\t' "$k" "$id")" "$inv" 2>/dev/null; then :; else printf '%s\t%s\t%s\n' "$k" "$id" "$what" >>"$inv"; fi
    case "$k" in
      container) if [[ "$id" =~ ^[0-9a-f]{64}$ && "ci-$c" =~ ^[A-Za-z0-9._-]{1,64}$ ]]; then row="standalone	$id	ci-$c"; fi ;;
      network) if [[ "$id" =~ ^[0-9a-f]{64}$ ]]; then row="network	$id"; fi ;;
      volume) if [[ "$id" =~ ^[A-Za-z0-9][A-Za-z0-9_.-]{0,254}$ ]]; then row="volume	$id"; fi ;;
      image) if [[ "$id" =~ ^sha256:[0-9a-f]{64}$ && "$tag" =~ ^[A-Za-z0-9][A-Za-z0-9._/:@+-]{0,254}$ ]]; then row="image	$tag	$id"; fi ;;
    esac
    if [[ -n "$row" && -n "$reg" && -f "$reg" && ! -L "$reg" ]] && ! grep -qxF -- "$row" "$reg"; then printf '%s\n' "$row" >>"$reg"; fi
  } 9>>"${inv%.tsv}.lock"
}
gate_cidarg() { # ARGS...: the caller's cidfile (--cidfile X or --cidfile=X), if any
  while (( $# > 0 )); do case "$1" in --cidfile) printf '%s' "${2:-}"; return 0 ;; --cidfile=*) printf '%s' "${1#--cidfile=}"; return 0 ;; esac; shift; done
  return 1
}
# G064 corr3 (Codex P1): the cidfile the Docker CLI writes at creation lies in the run registry's cid/ directory, the
# durable binding cleanup-owned.sh already resolves (cid/*.cid -> standalone, removed by exact ID). A container created
# by a caller that is then SIGKILLed during an attached call therefore stays recoverable. A caller's own --cidfile is
# taken off the command line and receives the ID once the CLI returns (refused when it already exists, as the CLI would).
GATE_ARGS=() GATE_CALLER_CID=""
gate_cidfile() { # REG CTL OWNER_PID -> the durable cidfile path; 1 when REG is set but cannot take one (refuse); 2 no registry
  local d
  if [[ -z "$1" ]]; then return 2; fi
  d="$(dirname -- "$1")/cid"
  if [[ -f "$1" && ! -L "$1" && -w "$1" && -d "$d" && ! -L "$d" && -w "$d" ]]; then :; else return 1; fi
  printf '%s/ci-%s-%s-%s%s.cid' "$d" "$2" "$3" "$RANDOM" "$RANDOM"
}
gate_cidargs() { # ARGS... -> GATE_ARGS (ARGS without the caller's --cidfile option) and GATE_CALLER_CID (its value)
  GATE_ARGS=() GATE_CALLER_CID=""
  while (( $# > 0 )); do
    case "$1" in
      --cidfile) GATE_CALLER_CID="${2:-}"; if (( $# > 1 )); then shift 2; else shift; fi ;;
      --cidfile=*) GATE_CALLER_CID="${1#--cidfile=}"; shift ;;
      *) GATE_ARGS+=("$1"); shift ;;
    esac
  done
}
GATE
# shellcheck source=/dev/null
source "$ROOT/shim/gate.sh"
if [[ -n "${CI_CONTROLS_CEILING_LOWER:-}" ]]; then
  if [[ ! "$CI_CONTROLS_CEILING_LOWER" =~ ^[A-Za-z0-9._-]+:([0-9]+)$ ]] || (( BASH_REMATCH[1] > 8 )); then
    out "FAIL setup CI_CONTROLS_CEILING_LOWER=$CI_CONTROLS_CEILING_LOWER is malformed or would raise a ceiling (at most 8)"; rm -rf -- "$ROOT"; exit 2
  fi
fi
REG_MIRROR="${CI_CONTROLS_OWNED_REG:-${SR_REG:+$SR_REG/owned.tsv}}" HOLD_RUN=""
if [[ -n "$REG_MIRROR" ]]; then HOLD_RUN="$(dirname -- "$REG_MIRROR")/hold.d/run"; fi
# D: the real CLI; a persisting creation of a control (create, run -d, network/volume create, build, commit) is gated
# and recorded as above. Everything else, removal included, passes through.
D() {
  local k="" a cf="" ccf="" id="" out rc=0 tag="" inv cdir
  case "${1:-}" in
    create) k=container ;;
    run) for a in "${@:2}"; do if [[ "$a" == -d || "$a" == --detach ]]; then k=container; break; fi; done ;;
    network|volume) if [[ "${2:-}" == create ]]; then k="$1"; fi ;;
    build|commit) k=image ;;
    image) if [[ "${2:-}" == build ]]; then k=image; fi ;;
  esac
  # An explicit status: a bare return in a function called from the EXIT trap (on_exit) would return the pre-trap status.
  if [[ -z "$k" || -z "${CTL_ID:-}" ]]; then "$REAL_DOCKER" "$@"; return $?; fi
  inv="$ROOT/owned/$CTL_ID.tsv"
  if [[ "$k" == container ]]; then
    # G064 corr3 (Codex P1): the CLI's cidfile lies in the run registry (gate_cidfile); without a registry one level
    # under $ROOT/owned as before (-p for a concurrent creator, then the private mode). A caller's cidfile gets the ID.
    gate_cidargs "$@"; ccf="$GATE_CALLER_CID"
    if cf="$(gate_cidfile "$REG_MIRROR" "$CTL_ID" "$BASHPID")"; then :; elif [[ -z "$REG_MIRROR" ]]; then
      cdir="${inv%.tsv}.cid"; mkdir -p -- "$cdir" && chmod 700 -- "$cdir"
      cf="$cdir/d-$BASHPID-$RANDOM$RANDOM.cid"
    else echo "validate-controls: refused: the run registry $REG_MIRROR cannot take a cidfile; nothing created" >&2; return 1; fi
    if [[ -n "$ccf" && -e "$ccf" ]]; then echo "validate-controls: refused: the cidfile $ccf exists; nothing created" >&2; return 1; fi
  fi
  if gate_reserve "$inv" "$REAL_DOCKER" "$k" "sdlc.ci.control=$RT-$CTL_ID" "$(gate_ceil control "$CTL_ID" "$k")" "$BASHPID" "$HOLD_RUN" "$cf"; then :; else
    echo "validate-controls: $GATE_MSG" >&2; return 1
  fi
  if [[ "$k" == container ]]; then
    "$REAL_DOCKER" "${GATE_ARGS[0]}" --cidfile "$cf" "${GATE_ARGS[@]:1}" || rc=$?
    id="$(tr -d ' \n' <"$cf" 2>/dev/null)" || id=""
    if [[ -n "$id" ]]; then
      gate_record "$inv" container "$id" "D $1" "$REG_MIRROR"
      if [[ -n "$ccf" ]]; then printf '%s' "$id" >"$ccf" || echo "validate-controls: the cidfile $ccf could not be written (container $id is recorded)" >&2; fi
    fi
  else
    out="$("$REAL_DOCKER" "$@")" || rc=$?
    if [[ -n "$out" ]]; then printf '%s\n' "$out"; fi
    id="$(awk 'NF {l = $NF} END {print l}' <<<"$out")"
    if [[ "$k" == image ]]; then
      if [[ "$1" == commit ]]; then tag="${!#}"; else tag="$(a=0; for x in "$@"; do if (( a )); then printf '%s' "$x"; break; fi; if [[ "$x" == -t || "$x" == --tag ]]; then a=1; fi; done)"; fi
      if [[ ! "$id" =~ ^sha256:[0-9a-f]{64}$ && -n "$tag" ]]; then id="$("$REAL_DOCKER" image inspect -f '{{.Id}}' "$tag" 2>/dev/null)" || id=""; fi
    fi
    if (( rc == 0 )) && [[ -n "$id" ]]; then gate_record "$inv" "$k" "$id" "D $1" "$REG_MIRROR" "$tag"; fi
  fi
  gate_release
  return "$rc"
}

# sweep LABEL: remove every container, network, volume and image carrying LABEL (control resources only;
# the label carries this run's random token), then verify that none remains. A listing Docker refuses during the
# verification is unverified, never absent. SWEEP_LEFT: what remains (exact IDs) or could not be listed.
# G065 K1: images are listed with -a (untagged ones too); a failed image listing or removal is unverified, never success.
SWEEP_LEFT=""
sweep() {
  local l="$1" ids x left=0
  SWEEP_LEFT=""
  ids="$(D ps -aq --no-trunc --filter "label=$l")" || ids=""
  for x in $ids; do D rm -f -v "$x" >/dev/null 2>&1 || true; done
  ids="$(D network ls -q --no-trunc --filter "label=$l")" || ids=""
  for x in $ids; do D network rm "$x" >/dev/null 2>&1 || true; done
  ids="$(D volume ls -q --filter "label=$l")" || ids=""
  for x in $ids; do D volume rm "$x" >/dev/null 2>&1 || true; done
  ids="$(D image ls -aq --no-trunc --filter "label=$l" | LC_ALL=C sort -u)" || { ids=""; left=1; SWEEP_LEFT+="(the image listing failed) "; }
  for x in $ids; do D image rm -f "$x" >/dev/null 2>&1 || { left=1; SWEEP_LEFT+="(image $x could not be removed) "; }; done
  for x in "container ps -aq --no-trunc" "network network ls -q --no-trunc" "volume volume ls -q" "image image ls -aq --no-trunc"; do
    # shellcheck disable=SC2086  # x: the kind, then a fixed subcommand list
    if ids="$(D ${x#* } --filter "label=$l" 2>/dev/null)"; then
      if [[ -n "$ids" ]]; then left=1; SWEEP_LEFT+="${x%% *} $(tr '\n' ' ' <<<"$ids")"; fi
    else left=1; SWEEP_LEFT+="(the ${x%% *} listing failed) "; fi
  done
  return "$left"
}
# --- nested launches: positive process ownership, forwarded termination, bounded grace, reaping -----------
LAUNCHES="$ROOT/launches"   # PGID<TAB>start time<TAB>control, one line per nested launch, written by the launch
: >"$LAUNCHES"
STOP_GRACE_S=30 LPID="" STOP_LEFT="" OWN_SCAN=""
proc_stat() { # PID -> "STATE START" (start time in clock ticks, as /proc/PID/stat field 22); fails when gone
  local s f
  s="$(cat "/proc/$1/stat" 2>/dev/null)" || return 1
  s="${s##*) }"
  read -r -a f <<<"$s"
  if [[ -z "${f[19]:-}" ]]; then return 1; fi
  printf '%s %s' "${f[0]}" "${f[19]}"
}
proc_start() { local s; s="$(proc_stat "$1")" || return 1; printf '%s' "${s#* }"; }
proc_is() { local s; s="$(proc_stat "$1")" || return 1; [[ "${s%% *}" != Z && "${s#* }" == "$2" ]]; } # PID START: same live process
# own_scan ROOTS GROUPS -> OWN_SCAN: "p PID START PGID" for every live owned process and "g PGID" for every
# owned group with a live member. Owned: ROOTS ("PID:START", only while that process has that start time)
# and their descendants; GROUPS given as "PGID:START" (a launch: only while its leader is alive with that
# start time, so a reused group number is never owned) or as a bare PGID (a group already proven owned by
# the same bounded stop); every member of an owned group; every group whose leader is owned; to a
# fixpoint. This shell and the runner's main shell are never owned. Fails when the process table cannot be
# read.
OWN_AWK='
function start(p,   f, l, a) { f = "/proc/" p "/stat"; l = ""
  if ((getline l < f) <= 0) { close(f); return "" }
  close(f); sub(/^.*\) /, "", l); split(l, a, " "); return a[20] }
{ n++; pid[n] = $1; pp[n] = $2; pg[n] = $3; z[n] = ($4 ~ /^Z/) }
END {
  k = split(roots, a, " ")
  for (i = 1; i <= k; i++) { if (split(a[i], b, ":") == 2 && b[2] != "") want[b[1]] = b[2] }
  for (i = 1; i <= n; i++) if ((pid[i] in want) && start(pid[i]) == want[pid[i]]) own[pid[i]] = 1
  k = split(groups, a, " ")
  for (i = 1; i <= k; i++) {
    if (a[i] ~ /^[0-9]+$/ && a[i] + 0 > 1) grp[a[i]] = 1
    else if (split(a[i], b, ":") == 2 && b[1] ~ /^[0-9]+$/ && b[1] + 0 > 1 && b[2] != "") lead[b[1]] = b[2]
  }
  for (i = 1; i <= n; i++) if ((pid[i] in lead) && pg[i] == pid[i] && start(pid[i]) == lead[pid[i]]) grp[pid[i]] = 1
  do {
    ch = 0
    for (i = 1; i <= n; i++) {
      p = pid[i]
      if (p == self || p == me) continue
      if (!(p in own) && ((pp[i] in own) || (pg[i] in grp))) { own[p] = 1; ch = 1 }
      if ((p in own) && pg[i] == p && !(p in grp)) { grp[p] = 1; ch = 1 }
    }
  } while (ch)
  for (i = 1; i <= n; i++) {
    p = pid[i]
    if (!(p in own) || z[i] || p == self || p == me) continue
    s = start(p)
    if (s == "") continue
    live[pg[i]] = 1
    print "p", p, s, pg[i]
  }
  for (g in grp) if (g in live) print "g", g
}'
own_scan() {
  local me="$BASHPID" t
  t="$(ps -eo pid=,ppid=,pgid=,stat=)" || return 1
  OWN_SCAN="$(awk -v roots="$1" -v groups="$2" -v self="$$" -v me="$me" "$OWN_AWK" <<<"$t")" || return 1
}
# stop_tree ROOTS GROUPS: stop exactly the owned tree (own_scan). TERM once to every owned group and to every owned
# process outside those groups (newcomers when they appear), up to STOP_GRACE_S of grace, then KILL each
# round until nothing owned is alive (10 s bound). A process that loses its parent stays owned by PID and
# start time. Returns 0 when nothing owned is left; otherwise 1 with the survivors in STOP_LEFT.
stop_tree() {
  local roots="$1" groups="$2" sig=TERM end k t a b c
  local -A seen=() gseen=() sent=() rg=()
  STOP_LEFT=""
  end=$((SECONDS + STOP_GRACE_S))
  while :; do
    k="$roots"
    for a in "${!seen[@]}"; do k+=" $a:${seen[$a]}"; done
    if own_scan "$k" "$groups ${!gseen[*]}"; then :; else STOP_LEFT="(the process table could not be read)"; return 1; fi
    if [[ -z "$OWN_SCAN" ]]; then return 0; fi
    if (( SECONDS >= end )); then
      if [[ "$sig" == KILL ]]; then
        while read -r t a b c; do if [[ "$t" == p ]]; then STOP_LEFT+="$a "; fi; done <<<"$OWN_SCAN"
        return 1
      fi
      sig=KILL; end=$((SECONDS + 10))
    fi
    rg=()
    while read -r t a b c; do
      if [[ "$t" == g ]]; then
        rg[$a]=1; gseen[$a]=1
        if [[ "$sig" == KILL || -z "${sent[g$a]+x}" ]]; then kill -"$sig" -- "-$a" 2>/dev/null || true; sent[g$a]=1; fi
      fi
    done <<<"$OWN_SCAN"
    while read -r t a b c; do
      if [[ "$t" == p ]]; then
        seen[$a]="$b"
        if [[ -n "${rg[$c]+x}" ]]; then continue; fi
        if [[ "$sig" == KILL || -z "${sent[p$a:$b]+x}" ]]; then kill -"$sig" "$a" 2>/dev/null || true; sent[p$a:$b]=1; fi
      fi
    done <<<"$OWN_SCAN"
    sleep 0.25
  done
}
# launch OUT CMD...: CMD in the background as its own session and process group (setsid keeps the PID, so
# the group ID is $LPID); the launch records itself in $LAUNCHES before CMD runs.
launch() {
  local o="$1"
  shift
  ( p="$BASHPID"; printf '%s\t%s\t%s\n' "$p" "$(proc_start "$p")" "${CTL_ID:-}" >>"$LAUNCHES"; exec setsid "$@" ) >"$o" 2>&1 &
  LPID=$!
}
# await PID SECONDS: wait for a launch, bounded (--launch-deadline caps the bound). At the bound the launch's
# owned tree is stopped and reaped, a problem is recorded and RC is 124; otherwise RC is its exit status.
await() {
  local pid="$1" lim="$2" end
  if [[ -n "$LAUNCH_CAP" ]] && (( LAUNCH_CAP < lim )); then lim="$LAUNCH_CAP"; fi
  end=$((SECONDS + lim))
  while kill -0 "$pid" 2>/dev/null; do
    if (( SECONDS >= end )); then
      prob "deadline: a nested launch was still running after ${lim} s; its process tree was stopped and reaped"
      if stop_tree "" "$pid:$(proc_start "$pid")"; then wait "$pid" 2>/dev/null || true; else prob "deadline: owned processes survived KILL: $STOP_LEFT"; fi
      RC=124
      return 0
    fi
    sleep 0.2
  done
  RC=0
  wait "$pid" || RC=$?
}
launch_groups() { awk -F'\t' -v c="${1:-}" '(c == "" || $3 == c) && $2 != "" {printf "%s:%s ", $1, $2}' "$LAUNCHES" 2>/dev/null || true; }

# --- per-control resource inventory (G064 correction 11; continuous resource accountability) ---------------
# A control whose resources can outlive it (V18l-V18n: a nested run whose holder keeps creating containers) keeps
# $ROOT/owned/<control>.tsv, started before anything it lists exists. Each creator appends a row as soon as it has
# the identity: "process<TAB>PID<TAB>START<TAB>what" for itself, "container|network|volume|image<TAB>ID<TAB>what"
# per object. owned_cleanup runs after the control on every exit path (return, failure, abort) and on cancellation
# (on_exit): it stops the recorded creators and the control's registered launches first (stop_tree: TERM, the
# bounded grace, KILL, reaped), then removes every recorded object and every object still carrying
# sdlc.ci.outer=<run>-<control>, and verifies at once and again after OWNED_SETTLE_S (what appears only then is late
# recreation, removed in turn). Failures (OWNED_MSG, exact IDs): a survivor, a late or unrecorded labelled object, a
# creator surviving KILL, a listing Docker refused. The control's own assertions ran before, inside the control.
# An inventory is deleted only after a verified cleanup. When the cleanup, the control's sweep or the stop of an
# outliving launch is unverified, the control is held (HELD): its inventory stays as the recovery record, with its
# problems file ($ROOT/problems/<control>) as the receipt and the scratch root kept; the FAIL line names both and every
# recorded ID (owned_ids); no later control is started (each is reported as not started, never as passed). on_exit
# retries the cleanup of every remaining inventory and deletes one only when that retry is verified.
OWNED_SETTLE_S=3 OWNED_MSG=() OWNED_N=0 HELD=""
# G064 r13 (real-Docker V18l-n): in a nested run (CI_CONTROLS_OUTER_OWNED: the outer control's inventory; --extra-label
# sdlc.ci.outer=<outer run>-<control>), every shim creation of a case launch is gated and recorded in the OUTER scope
# (SHIM_GATE_OUTER), not in this run's control inventory, yet it also carries this run's own sdlc.ci.control label (the
# case labels). Such an object is accounted by the outer scope: owned_cleanup still removes it here, but it is
# `unrecorded` only when the outer inventory has no row for it either. A cidfile row there whose ID the CLI never
# wrote (a creation KILLed between the daemon's create and the write) counts as accounting for it.
OUTER_INV="${CI_CONTROLS_OUTER_OWNED:-}" OUTER_LABS=""
for ol_e in "${EXTRA_LABELS[@]}"; do case "$ol_e" in sdlc.ci.outer=*) OUTER_LABS+="$ol_e " ;; esac; done
outer_accounts() { # KIND ID: 0 when the outer inventory records ID (or holds an unresolved cidfile row, KIND container)
  local k a b
  if [[ -z "$OUTER_INV" || ! -f "$OUTER_INV" || -L "$OUTER_INV" ]]; then return 1; fi
  while IFS=$'\t' read -r k a b; do
    if [[ "$k" == cidfile && "$1" == container ]]; then
      if [[ "$a" == /* && -f "$a" && ! -L "$a" ]]; then a="$(tr -d ' \n' <"$a" 2>/dev/null)" || a=""; else a=""; fi
      if [[ -z "$a" || "$a" == "$2" ]]; then return 0; fi
    elif [[ "$k" == "$1" && "$a" == "$2" ]]; then return 0; fi
  done <"$OUTER_INV"
  return 1
}
owned_ids() { # CTL: every ID its inventory records ("process PID:START", "KIND ID"), comma-separated; nonsecret
  owned_rows "$1" | awk -F'\t' '$2 ~ /^[A-Za-z0-9:._-]+$/ && $1 == "process" && !u[$1 FS $2 FS $3]++ { printf "%sprocess %s:%s", s, $2, $3; s = ", " }
    $2 ~ /^[A-Za-z0-9:._-]+$/ && $1 ~ /^(container|network|volume|image)$/ && !u[$1 FS $2]++ { printf "%s%s %s", s, $1, $2; s = ", " }' || true
}
owned_file() { : >>"$ROOT/owned/$CTL_ID.tsv"; printf '%s' "$ROOT/owned/$CTL_ID.tsv"; } # the control's inventory (the runner starts it)
owned_rows() { # CTL: its inventory rows, each `cidfile` row resolved to "container<TAB>ID" once the CLI wrote the ID
  local k a b
  while IFS=$'\t' read -r k a b; do
    if [[ "$k" == cidfile ]]; then
      if [[ "$a" == /* && -f "$a" && ! -L "$a" ]]; then a="$(tr -d ' \n' <"$a" 2>/dev/null)" || a=""; else a=""; fi
      if [[ -n "$a" ]]; then printf 'container\t%s\tcidfile\n' "$a"; fi
    else printf '%s\t%s\t%s\n' "$k" "$a" "$b"; fi
  done <"$ROOT/owned/$1.tsv" 2>/dev/null
}
owned_add() { printf '%s\t%s\t%s\n' "$1" "$2" "$3" >>"$ROOT/owned/$CTL_ID.tsv"; }    # KIND ID WHAT
owned_proc() { local s; s="$(proc_start "$1")" || return 0; printf 'process\t%s\t%s\t%s\n' "$1" "$s" "$2" >>"$ROOT/owned/$CTL_ID.tsv"; } # PID WHAT
owned_ls() { # KIND [LABEL]: the full IDs (volumes: names) of every object of KIND, or of those carrying LABEL
  local f=()
  if [[ -n "${2:-}" ]]; then f=(--filter "label=$2"); fi
  case "$1" in
    container) D ps -aq --no-trunc "${f[@]}" ;;
    network) D network ls -q --no-trunc "${f[@]}" ;;
    volume) D volume ls -q "${f[@]}" ;;
    image) D image ls -aq --no-trunc "${f[@]}" ;;
  esac
}
owned_present() { # KIND "LABEL..." [ID...]: the given IDs that still exist and every object carrying a LABEL, once each
  local k="$1" l="$2" all lab="" x y
  shift 2
  all="$(owned_ls "$k")" || return 1
  for y in $l; do x="$(owned_ls "$k" "$y")" || return 1; lab+="$x"$'\n'; done
  { for x in "$@"; do if grep -qxF -- "$x" <<<"$all"; then printf '%s\n' "$x"; fi; done; printf '%s\n' "$lab"; } | awk 'NF && !s[$0]++'
}
owned_rm() { # KIND ID...
  local k="$1"
  shift
  case "$k" in
    container) D rm -f -v "$@" ;;
    network) D network rm "$@" ;;
    volume) D volume rm -f "$@" ;;
    image) D image rm -f "$@" ;;
  esac >/dev/null 2>&1 || true
}
owned_cleanup() { # CTL: 0 when clean; otherwise 1 with OWNED_MSG. OWNED_N: the objects it removed.
  local ctl="$1" inv="$ROOT/owned/$1.tsv" ol="sdlc.ci.outer=$RT-$1" cl="sdlc.ci.control=$RT-$1" k a b roots="" ids x late recs=() y lx ox
  local -A rec=() bad=() s1=()
  OWNED_MSG=() OWNED_N=0
  if [[ ! -f "$inv" ]]; then return 0; fi
  # G064 corr1 P6: every control keeps an inventory; objects carrying its own control label are checked like the
  # outer-labelled ones (an object without a row is `unrecorded`); cidfile rows count as their container.
  while IFS=$'\t' read -r k a b _; do
    case "$k" in
      process) if [[ "$a" =~ ^[1-9][0-9]*$ && "$b" =~ ^[0-9]+$ ]]; then roots+=" $a:$b"; fi ;;
      container|network|volume|image) if [[ "$a" =~ ^[A-Za-z0-9:._-]+$ && " ${rec[$k]:-}" != *" $a "* ]]; then rec[$k]+="$a "; fi ;;
    esac
  done < <(owned_rows "$ctl")
  # Creators first: nothing that could create another object is alive when the removal starts.
  if stop_tree "${roots# }" "$(launch_groups "$ctl")"; then :; else OWNED_MSG+=("creator processes survived KILL: $STOP_LEFT"); fi
  for k in container network volume image; do
    read -r -a recs <<<"${rec[$k]:-}"
    if ids="$(owned_present "$k" "$ol $cl" "${recs[@]}")"; then :; else bad[$k]=1; OWNED_MSG+=("Docker could not list the ${k}s; their removal is unverified"); continue; fi
    ox=""
    if [[ -n "$OUTER_LABS" && -n "$ids" ]]; then for y in $OUTER_LABS; do ox+="$(owned_ls "$k" "$y" 2>/dev/null)"$'\n' || true; done; fi
    for x in $ids; do
      if [[ " ${rec[$k]:-}" == *" $x "* ]]; then continue; fi
      # an object of the outer scope (see OUTER_INV) recorded there: removed below, not unrecorded
      if [[ -n "$ox" ]] && grep -qxF -- "$x" <<<"$ox" && outer_accounts "$k" "$x"; then continue; fi
      y="$ol"; if lx="$(owned_ls "$k" "$cl" 2>/dev/null)" && grep -qxF -- "$x" <<<"$lx"; then y="$cl"; fi
      if [[ -n "$ox" ]] && grep -qxF -- "$x" <<<"$ox"; then y+=" (outer scope: no row in $OUTER_INV either)"; fi
      OWNED_MSG+=("unrecorded $k $x carries $y")
    done
    # shellcheck disable=SC2086  # ids: one Docker ID per word
    if [[ -n "$ids" ]]; then OWNED_N=$((OWNED_N + $(grep -c . <<<"$ids"))); owned_rm "$k" $ids; fi
  done
  for k in container network volume image; do
    if [[ -n "${bad[$k]:-}" ]]; then continue; fi
    read -r -a recs <<<"${rec[$k]:-}"
    if s1[$k]="$(owned_present "$k" "$ol $cl" "${recs[@]}")"; then :; else bad[$k]=1; OWNED_MSG+=("Docker could not verify the ${k}s"); fi
  done
  # Late recreation needs a creator: the settle wait applies when the inventory recorded a creator or objects were removed.
  if [[ -n "$roots" ]] || (( OWNED_N > 0 )); then sleep "$OWNED_SETTLE_S"; fi
  for k in container network volume image; do
    if [[ -n "${bad[$k]:-}" ]]; then continue; fi
    read -r -a recs <<<"${rec[$k]:-}"
    if ids="$(owned_present "$k" "$ol $cl" "${recs[@]}")"; then :; else OWNED_MSG+=("Docker could not verify the ${k}s"); continue; fi
    late=""
    for x in $ids; do if grep -qxF -- "$x" <<<"${s1[$k]}"; then :; else late+="$x "; fi; done
    if [[ -n "$late" ]]; then
      OWNED_MSG+=("late recreation after the sweep: $k ${late% }")
      # shellcheck disable=SC2086  # late: one Docker ID per word
      owned_rm "$k" $late
      if ids="$(owned_present "$k" "$ol $cl" "${recs[@]}")"; then :; else OWNED_MSG+=("Docker could not verify the ${k}s"); continue; fi
    fi
    if [[ -n "$ids" ]]; then OWNED_MSG+=("owned ${k}s survived the cleanup: $(tr '\n' ' ' <<<"$ids")"); fi
  done
  (( ${#OWNED_MSG[@]} == 0 ))
}

# The running control, its process ID and start time; the cancelling signal; the counts so far.
CUR_CTL="" CPID="" CSTART="" CANCEL_SIG="" PASSN=0 FAILN=0
on_cancel() { trap '' TERM INT HUP; CANCEL_SIG="$1"; exit $((128 + $2)); }
on_exit() {
  local rc=$? keep="$KEEP" left=0 cp="$CPID" cs="$CSTART" f c
  set +e
  trap '' TERM INT HUP
  # This run's own nested work stops first (the running control with its descendants, and every registered
  # launch): TERM, bounded grace, KILL, reaped. Only then are resources swept and the scratch removed.
  if [[ -n "$CUR_CTL" && -z "$cp" ]]; then cp="${!:-}"; cs="$(proc_start "$cp")" || cs=""; fi
  if stop_tree "${cp:+$cp:$cs}" "$(launch_groups)"; then
    if [[ -n "$cp" ]]; then wait "$cp" 2>/dev/null; fi
  else
    left=1
  fi
  if [[ -n "$CANCEL_SIG" && -n "$CUR_CTL" ]]; then
    out "FAIL $CUR_CTL cancelled by SIG$CANCEL_SIG; its control and nested launches were stopped and reaped before the sweep"
  fi
  if (( left )); then out "FAIL cancel owned processes survived KILL: $STOP_LEFT; scratch kept at $ROOT"; keep=1; fi
  # The interrupted control's inventory and every held one (G064 correction 11), cleaned as after a control; a failure is
  # terminal. An inventory is deleted only after a verified cleanup; otherwise it stays, with the scratch root, as the
  # recovery record, and the FAIL line names it and its exact IDs.
  for f in "$ROOT/owned/"*.tsv; do
    if [[ -f "$f" ]]; then :; else continue; fi
    c="${f##*/}"; c="${c%.tsv}"
    if owned_cleanup "$c"; then
      rm -rf -- "$f" "${f%.tsv}.hold" "${f%.tsv}.lock" "${f%.tsv}.rsv" "${f%.tsv}.cid"
      if [[ " $HELD" == *" $c "* ]]; then out "cleanup $c: the retry at exit verified the cleanup; its inventory was deleted (receipt kept: $ROOT/problems/$c)"; fi
    else
      out "FAIL cleanup $c: $(printf '%s; ' "${OWNED_MSG[@]}")inventory kept as the recovery record: $f (recorded IDs: $(owned_ids "$c"))"; left=1
    fi
  done
  if [[ -n "$CANCEL_SIG" ]]; then sleep 1; fi
  if sweep "$RUN_LABEL" >/dev/null 2>&1; then :; else out "FAIL cleanup control resources of run $RT remain: ${SWEEP_LEFT% }"; left=1; fi
  # Self-registration (C4): the run registry is finalized last; unverified keeps it and the marker for recovery.
  if (( SR_ON )); then
    if sr_finalize; then :; else
      out "FAIL cleanup self-registration: run $REG_JOB's registry $SR_REG did not finalize verified (see its cleanup-inventory.tsv); it and the marker $PEND_FILE are kept: the next local validate.sh launch recovers it"; left=1
    fi
  fi
  if [[ -n "$CANCEL_SIG" ]]; then
    out "ci-controls: cancelled by SIG$CANCEL_SIG; ${PASSN} passed, ${FAILN} failed before the cancellation (of ${#SELECTED[@]} selected; ${#CONTROLS[@]} defined)"
  fi
  # Unverified cleanup (here or after a control) keeps the scratch root: it holds the recovery record and the receipts.
  if (( left )) || [[ -n "$HELD" ]]; then keep=1; fi
  if (( left && rc == 0 )); then rc=1; fi
  chmod -R u+rwx "$ROOT" 2>/dev/null || true
  if (( keep )); then echo "validate-controls: scratch kept at $ROOT${STATE_BASE:+ (case state roots: $STATE_BASE)}" >&2; else
    rm -rf -- "$ROOT" 2>/dev/null || true
    if [[ -n "$STATE_BASE" ]]; then chmod -R u+rwx "$STATE_BASE" 2>/dev/null || true; rm -rf -- "$STATE_BASE" 2>/dev/null || true; fi
  fi
  exit "$rc"
}
trap on_exit EXIT
trap 'on_cancel TERM 15' TERM
trap 'on_cancel INT 2' INT
trap 'on_cancel HUP 1' HUP

# --- docker shim: pass-through except configured faults; labels every container it creates -----------
cat >"$ROOT/shim/docker" <<'SHIM'
#!/usr/bin/env bash
# docker shim for infra/ci/validate-controls.sh: everything passes through to the real CLI except the
# configured fault patterns. Every container created through it carries the control labels.
real="${SHIM_REAL_DOCKER:?}"
if [[ -n "${SHIM_LOG:-}" ]]; then printf '%s\n' "$*" >>"$SHIM_LOG"; fi
# G064 Plan A corr1 C2 / corr2 D2: every run|create and network|volume create of a case launch is gated (gate.sh) in
# its scope: the V18l-n nested holder (SHIM_GATE_OUTER, its sdlc.ci.outer label; 39 containers) or the control
# (SHIM_GATE_ROOT/owned/<control>.tsv, its sdlc.ci.control label; 8/2/2). A scope whose inventory is gone refuses.
G_INV="" G_LAB="" G_SCOPE="" G_CTL=""
shim_gate() { # KIND -> G_INV/G_LAB/G_SCOPE/G_CTL (empty: no gated scope); 1 when the scope's inventory is gone
  local l o="" c=""
  if [[ -z "${SHIM_GATE_ROOT:-}" || ! -f "${SHIM_GATE_LIB:-}" ]]; then return 0; fi
  # shellcheck source=/dev/null
  source "$SHIM_GATE_LIB"
  for l in ${SHIM_LABELS:-}; do case "$l" in sdlc.ci.outer=*) o="$l" ;; sdlc.ci.control=*) c="$l" ;; esac; done
  if [[ -n "${SHIM_GATE_OUTER:-}" && -n "$o" ]]; then
    G_INV="$SHIM_GATE_OUTER" G_LAB="$o" G_SCOPE=outer
  elif [[ "$c" =~ ^sdlc\.ci\.control=[0-9a-f]+-([A-Za-z0-9._-]+)$ ]]; then
    G_INV="$SHIM_GATE_ROOT/owned/${BASH_REMATCH[1]}.tsv" G_LAB="$c" G_SCOPE=control
  else return 0; fi
  G_CTL="${G_INV##*/}"; G_CTL="${G_CTL%.tsv}"
  if [[ -f "$G_INV" ]]; then return 0; fi
  echo "shim: refused: the inventory $G_INV of this scope does not exist" >&2; return 1
}
shim_reserve() { # KIND CIDFILE
  if gate_reserve "$G_INV" "$real" "$1" "$G_LAB" "$(gate_ceil "$G_SCOPE" "$G_CTL" "$1")" "$$" "${SHIM_GATE_HOLDRUN:-}" "$2"; then return 0; fi
  echo "shim: $GATE_MSG" >&2; return 1
}
# G051 VR4: the daemon is unreachable for every call.
if [[ "${SHIM_DOCKER_DOWN:-0}" == 1 ]]; then echo "shim: cannot connect to the Docker daemon" >&2; exit 1; fi
# G064 corr3 (Codex P2): the image-producing build paths of a case (build, buildx build, image build; the ledger's
# builds) are gated like D()'s build in their scope: run latch, control latch, the image ceiling and a reservation;
# the image ID (the caller's --iidfile, else one injected; else the tag resolved) is recorded and mirrored for recovery.
bsub=() brest=()
case "${1:-}" in
  build) bsub=(build); brest=("${@:2}") ;;
  buildx|image) if [[ "${2:-}" == build ]]; then bsub=("$1" build); brest=("${@:3}"); fi ;;
esac
if (( ${#bsub[@]} > 0 )); then
  if shim_gate image; then :; else exit 1; fi
  if [[ -n "$G_INV" ]]; then
    tag="" iid="" a=0 bextra=()
    for x in "${brest[@]}"; do
      if (( a == 1 )); then tag="$x"; elif (( a == 2 )); then iid="$x"; fi
      a=0
      case "$x" in -t|--tag) a=1 ;; --tag=*) tag="${x#--tag=}" ;; --iidfile) a=2 ;; --iidfile=*) iid="${x#--iidfile=}" ;; esac
    done
    if [[ -z "$iid" ]]; then mkdir -p -m 700 -- "${G_INV%.tsv}.cid"; iid="${G_INV%.tsv}.cid/b-$$-$RANDOM$RANDOM.iid"; bextra=(--iidfile "$iid"); fi
    shim_reserve image "" || exit 1
    rc=0; "$real" "${bsub[@]}" "${bextra[@]}" "${brest[@]}" || rc=$?
    id="$(tr -d ' \n' <"$iid" 2>/dev/null)" || id=""
    if [[ ! "$id" =~ ^sha256:[0-9a-f]{64}$ && -n "$tag" ]]; then id="$("$real" image inspect -f '{{.Id}}' "$tag" 2>/dev/null)" || id=""; fi
    if (( rc == 0 )) && [[ -n "$id" ]]; then gate_record "$G_INV" image "$id" "shim ${bsub[*]}" "${SHIM_GATE_REG:-}" "$tag"; fi
    gate_release; exit "$rc"
  fi
fi
case "${1:-}" in
  pull) if [[ "${SHIM_PULL_NOOP:-0}" == 1 ]]; then exit 0; fi ;;
  system)
    if [[ "${2:-}" == df ]]; then
      case "${SHIM_FAIL_DF:-0}" in
        1) echo "shim: docker system df failed" >&2; exit 1 ;;
        # G063 repair (VS5): an injected synthetic transient failure; every odd-numbered query fails, the next one
        # passes through. No real daemon mutation is involved.
        transient) echo x >>"${SHIM_DF_COUNT:?}"; if (( $(wc -l <"$SHIM_DF_COUNT") % 2 )); then echo "shim: docker system df failed (transient)" >&2; exit 1; fi ;;
      esac
    fi ;;
  rm) if [[ "${SHIM_FAIL_RM:-0}" == 1 ]]; then echo "shim: docker rm refused" >&2; exit 1; fi ;;
  ps) if [[ "${SHIM_FAIL_PS_FILTER:-0}" == 1 && " $* " == *" --filter "* ]]; then echo "shim: docker ps failed" >&2; exit 1; fi ;;
  image)
    # G051 C1 (VB6): a failed image removal (refuse, or gone: after the image really went) and a failed presence query.
    if [[ "${2:-}" == rm && -n "${SHIM_FAIL_IMAGE_RM:-}" ]]; then
      if [[ "$SHIM_FAIL_IMAGE_RM" == gone ]]; then "$real" image rm "${@:3}" >/dev/null 2>&1 || true; fi
      echo "shim: docker image rm failed" >&2; exit 1
    fi
    if [[ "${SHIM_FAIL_IMAGE_REF:-0}" == 1 && "${2:-}" == ls && " $* " == *" reference="* ]]; then
      echo "shim: cannot connect to the Docker daemon" >&2; exit 1
    fi ;;
  run|create)
    if [[ "${SHIM_GITLEAKS_FAIL:-0}" == 1 && " $* " == *" zricethezav/gitleaks:"* ]]; then echo "shim: gitleaks fatal error" >&2; exit 1; fi
    if [[ "${SHIM_YQ_XML_FAIL:-0}" == 1 && " $* " == *" -p xml "* ]]; then echo "shim: yq XML parse failed" >&2; exit 1; fi
    sub="$1"; shift
    extra=()
    for l in ${SHIM_LABELS:-}; do extra+=(--label "$l"); done
    if shim_gate container; then :; else exit 1; fi
    if [[ -z "$G_INV" ]]; then exec "$real" "$sub" "${extra[@]}" "$@"; fi
    # G064 corr3 (Codex P1): the CLI's cidfile lies in the run registry (gate_cidfile); without a registry (simulation
    # modes) under <inventory>.cid/ as before. A caller's cidfile gets the ID after the call.
    gate_cidargs "$@"; ccf="$GATE_CALLER_CID"
    if cf="$(gate_cidfile "${SHIM_GATE_REG:-}" "$G_CTL" "$$")"; then :; elif [[ -z "${SHIM_GATE_REG:-}" ]]; then
      mkdir -p -m 700 -- "${G_INV%.tsv}.cid"; cf="${G_INV%.tsv}.cid/s-$$-$RANDOM$RANDOM.cid"
    else echo "shim: refused: the run registry ${SHIM_GATE_REG} cannot take a cidfile; nothing created" >&2; exit 1; fi
    if [[ -n "$ccf" && -e "$ccf" ]]; then echo "shim: refused: the cidfile $ccf exists; nothing created" >&2; exit 1; fi
    shim_reserve container "$cf" || exit 1
    rc=0; "$real" "$sub" --cidfile "$cf" "${extra[@]}" "${GATE_ARGS[@]}" || rc=$?
    id="$(tr -d ' \n' <"$cf" 2>/dev/null)" || id=""
    if [[ -n "$id" ]]; then
      gate_record "$G_INV" container "$id" "shim $sub" "${SHIM_GATE_REG:-}"
      if [[ -n "$ccf" ]]; then printf '%s' "$id" >"$ccf" || echo "shim: the cidfile $ccf could not be written (container $id is recorded)" >&2; fi
    fi
    gate_release; exit "$rc" ;;
  network|volume)
    if [[ "${2:-}" == create ]]; then
      if shim_gate "$1"; then :; else exit 1; fi
      if [[ -n "$G_INV" ]]; then
        shim_reserve "$1" "" || exit 1
        rc=0; o="$("$real" "$@")" || rc=$?
        if [[ -n "$o" ]]; then printf '%s\n' "$o"; fi
        id="$(awk 'NF {l = $NF} END {print l}' <<<"$o")"
        if (( rc == 0 )) && [[ -n "$id" ]]; then gate_record "$G_INV" "$1" "$id" "shim $1 create" "${SHIM_GATE_REG:-}"; fi
        gate_release; exit "$rc"
      fi
    fi ;;
esac
exec "$real" "$@"
SHIM
chmod 755 "$ROOT/shim/docker"

BASE_ENV=(PATH="$ROOT/shim:$PATH" HOME="${HOME:-/nonexistent}" LC_ALL=C GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1
  SHIM_REAL_DOCKER="$REAL_DOCKER" G003_STATE_ROOT=/nonexistent/g003-controls
  SHIM_GATE_ROOT="$ROOT" SHIM_GATE_LIB="$ROOT/shim/gate.sh" SHIM_GATE_REG="$REG_MIRROR" SHIM_GATE_HOLDRUN="$HOLD_RUN"
  SHIM_GATE_OUTER="${CI_CONTROLS_OUTER_OWNED:-}" CI_CONTROLS_CEILING_LOWER="${CI_CONTROLS_CEILING_LOWER:-}")
DOCKER_ENV=()
for v in DOCKER_HOST DOCKER_CONTEXT DOCKER_CONFIG DOCKER_CERT_PATH DOCKER_TLS_VERIFY; do
  if [[ -n "${!v:-}" ]]; then BASE_ENV+=("$v=${!v}"); DOCKER_ENV+=("$v=${!v}"); fi
done

# --- the scratch template tree ------------------------------------------------------------------------
TEMPLATE="$ROOT/template"
mk_template() {
  local t="$TEMPLATE"
  mkdir -p "$t/infra/ci" "$t/infra/validation" "$t/infra/security"
  cp -- "$REPO_R/infra/ci/validate.sh" "$REPO_R/infra/ci/lib.sh" "$REPO_R/infra/ci/junit.sh" "$REPO_R/infra/ci/cleanup-owned.sh" "$REPO_R/infra/ci/unchanged.allow" "$t/infra/ci/"
  cp -- "$REPO_R/infra/ci/select-profile.sh" "$REPO_R/infra/ci/profile-paths.tsv" "$t/infra/ci/"
  cp -- "$REPO_R/infra/security/gitleaks.toml" "$t/infra/security/"
  cp -- "$REPO_R/infra/validation/required-checks.tsv" "$t/infra/validation/"
  if [[ -f "$REPO_R/infra/validation/lib.sh" ]]; then cp -- "$REPO_R/infra/validation/lib.sh" "$t/infra/validation/lib.sh"; fi
  printf 'Scratch tree of infra/ci/validate-controls.sh (one tracked input).\n' >"$t/README"
  # The declared synthetic corpus of the scratch trees: validate.sh compares the normal contract reports
  # with it, and the stub harness derives its reports from it (one case per line).
  mkdir -p "$t/contracts/conformance/intelligence"
  {
    printf '{\n  "version": 1,\n  "cases": [\n'
    printf '    {"id": "SY01", "file": "cases/sy01.json", "schema": "execution-envelope", "kind": "document", "valid": true, "class": "valid"},\n'
    printf '    {"id": "SY02", "file": "cases/sy02.json", "schema": "execution-envelope", "kind": "document", "valid": false, "class": "schema-invalid"},\n'
    printf '    {"id": "SY03", "file": "cases/sy03.json", "schema": "stream-event", "kind": "document", "valid": false, "class": "invalid-json"}\n'
    printf '  ]\n}\n'
  } >"$t/contracts/conformance/intelligence/expectations.json"
  cat >"$t/infra/ci/runner-preflight.sh" <<'STUB'
#!/usr/bin/env bash
# Stub runner preflight for infra/ci/validate-controls.sh. STUB_PREFLIGHT (comma list): pass | fail |
# block-pgid (occupy state/harness.pgid) | block-p1 (pre-plant state/p1.env as a symlink) | rp6-hold
# (bind a detached httpd container through a cidfile, as RP6 does, and keep it).
set -euo pipefail
REG="" OUT=""
while (( $# > 0 )); do case "$1" in --registry) REG="$2"; shift 2 ;; --out) OUT="$2"; shift 2 ;; *) shift ;; esac; done
EV="$(dirname "$(dirname "$OUT")")"
mode=",${STUB_PREFLIGHT:-pass},"
printf 'invoked\n' >>"$STUB_DIR/preflight-invocations"
st=PASS
if [[ "$mode" == *,fail,* ]]; then st=FAIL; fi
for n in 1 2 3 4 5; do printf 'RP%s\tPASS\tstub\n' "$n" >>"$OUT"; done
printf 'RP6\t%s\tstub\n' "$st" >>"$OUT"
for n in 7 8 9 10 11 12; do printf 'RP%s\tPASS\tstub\n' "$n" >>"$OUT"; done
printf 'RP13\tINFO\tstub\n' >>"$OUT"
if [[ "$mode" == *,block-pgid,* ]]; then mkdir -p "$EV/state/harness.pgid"; fi
if [[ "$mode" == *,block-p1,* ]]; then : >"$STUB_DIR/p1-target"; ln -s "$STUB_DIR/p1-target" "$EV/state/p1.env"; fi
if [[ "$mode" == *,rp6-hold,* ]]; then
  token="$(tr -d ' \n' <"$REG/token")"
  printf 'label\tplatform.validation.preflight=%s\n' "$token" >>"$REG/owned.tsv"
  cf="$REG/cid/preflight-stub-1.cid"
  docker run -d --cidfile "$cf" --label "platform.validation.preflight=$token" --network none busybox:1.37.0 httpd -f -p 80 >/dev/null
  id="$(cat "$cf")"
  printf 'standalone\t%s\tpreflight-httpd\n' "$id" >>"$REG/owned.tsv"
  printf '%s\n' "$id" >"$STUB_DIR/held-id"
  : >"$STUB_DIR/preflight-created"
  sleep 3
fi
if [[ "$st" == FAIL ]]; then exit 1; fi
exit 0
STUB
  cat >"$t/infra/validation/run.sh" <<'STUB'
#!/usr/bin/env bash
# Stub harness for infra/ci/validate-controls.sh, shaped like infra/validation/run.sh: strict mode, an
# EXIT-trap cleanup, and stdout/stderr through a process-substituted tee that ignores TERM, INT and HUP.
# STUB_HARNESS (comma list) selects the behaviour; see the case blocks below. Rows and reports exist only for
# the suites named in the arguments ("all" = the full profile), as in run.sh; STUB_ROWS_SUITES replaces them,
# STUB_DROP_ID omits a row, STUB_FAIL_ID makes one FAIL, STUB_ADD_ROW adds a PASS row and STUB_RM_REPORT
# removes one report. G051: the split suites and --areas are accepted; the suites and areas given are recorded
# (harness-suites, harness-areas); feature-web writes browser-feature-*; every run writes builds.tsv and
# ledger-summary.tsv, and controls ci-controls-timings.tsv, as run.sh does.
set -euo pipefail
REPORT="" OWNED="" SECRETS="" SUITES=() AREAS=none
while (( $# > 0 )); do
  case "$1" in
    --report-dir) REPORT="$2"; shift 2 ;;
    --owned-registry) OWNED="$2"; shift 2 ;;
    --secrets-file) SECRETS="$2"; shift 2 ;;
    --areas) AREAS="$2"; shift 2 ;;
    --base|--allow-file) shift 2 ;;
    --*) shift ;;
    *) SUITES+=("$1"); shift ;;
  esac
done
declare -A ON=([final]=1)
for s in ${STUB_ROWS_SUITES:-${SUITES[*]}}; do
  if [[ "$s" == all ]]; then for x in lint lint-base lint-web lint-core controls ci-controls provenance contracts root migrations ops rename off; do ON[$x]=1; done; else ON[$s]=1; fi
done
on() { [[ -n "${ON[$1]+x}" ]]; }
S="$STUB_DIR"
printf '%s\n' "${SUITES[*]}" >"$S/harness-suites"
printf '%s\n' "$AREAS" >"$S/harness-areas"
SELF_REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)"
MODE=",${STUB_HARNESS:-pass},"
has() { [[ "$MODE" == *",$1,"* ]]; }
printf '%s %s\n' "$$" "$(ps -o pgid= -p $$ | tr -d ' ')" >"$S/harness.pids"
printf 'invoked\n' >>"$S/harness-invocations"
cleanup() { echo "stub harness cleanup: removing its owned resources"; echo ran >"$S/harness-exit-trap"; }
trap cleanup EXIT
if has ignore-term; then trap '' TERM; fi
exec > >(trap '' TERM INT HUP; exec tee -a "$S/stub-tee.log") 2>&1
RUN_ID="$(head -c 3 /dev/urandom | od -An -tx1 | tr -d ' \n')"
WORK="$S/work-$RUN_ID"
mkdir -p "$WORK" "$REPORT"
printf 'work_root\t%s\nlabel\tg001.run=%s\n' "$WORK" "$RUN_ID" >>"$OWNED"
canary="SECRET-$(head -c 12 /dev/urandom | od -An -tx1 | tr -d ' \n')"
if has no-secrets; then :; else printf '%s\n' "$canary" >>"$SECRETS"; fi
printf '%s\n' "$canary" >"$S/canary"
echo "harness console: run $RUN_ID starting"
sleep 300 &
g1=$!
( sleep 300 ) &
g2=$!
printf '%s %s %s %s\n' "$$" "$(ps -o pgid= -p $$ | tr -d ' ')" "$g1" "$g2" >"$S/harness.pids"
if has standalone-kill; then
  # The real harness helper: cidfile in the registry's cid/ directory, then a standalone row.
  # shellcheck source=/dev/null
  source "$SELF_REPO/infra/validation/lib.sh"
  OWNED_REG="$OWNED"
  docker_inventory_ids >"$WORK/inventory-start.ids"
  # Two calls through $(...), as the harness makes them: each needs its own cidfile.
  id1="$(docker_run_owned controls-probe -d --label "g001.run=$RUN_ID" --network none busybox:1.37.0 sleep 300)" || id1=""
  id2="$(docker_run_owned controls-probe-2 -d --label "g001.run=$RUN_ID" --network none busybox:1.37.0 sleep 300)" || id2=""
  printf '%s\n' "$id1" >"$S/standalone-id"; printf '%s\n' "$id2" >"$S/standalone-id2"
  find "$(dirname "$OWNED")/cid" -name '*.cid' -type f | wc -l >"$S/cid-count"
  if [[ -f "$S/docs/README.md" ]]; then
    # The real documented-block helper over the control's fixture: a `run --rm --build` block leaves no container.
    register_project "$(sed -n 's/^COMPOSE_PROJECT_NAME=//p' "$S/docs/.env")"
    rc=0; run_docs_block "$S/docs" README.md stub:check "$S/docs-check.log" || rc=$?
    printf '%s\n' "$rc" >"$S/docs-rc"
    docker image inspect -f '{{.Id}}' "$(cat "$S/docs/fresh-tag")" >"$S/docs-fresh-id" 2>/dev/null || true
    { grep '^image' "$OWNED" || true; } >"$S/docs-rows"
  fi
  kill -KILL $$
fi
if has cidfile-kill; then
  docker run -d --cidfile "$(dirname "$OWNED")/cid/$RUN_ID-1.cid" --label "g001.run=$RUN_ID" --network none busybox:1.37.0 sleep 300 >/dev/null
  cat "$(dirname "$OWNED")/cid/$RUN_ID-1.cid" >"$S/standalone-id"
  kill -KILL $$
fi
if has hold; then
  # Ignores TERM, INT and HUP, and so does its holder descendant, which (re)creates the stub directory and
  # a control container every 2 s for as long as it lives. With STUB_OWNED_LOG (the outer control's inventory;
  # V18l-V18n) the holder records itself (PID, start time) and each container it creates there at once.
  trap '' TERM INT HUP
  ( own="${STUB_OWNED_LOG:-}"
    if [[ -n "$own" ]]; then
      { read -r st <"/proc/$BASHPID/stat" && st="${st##*) }" && read -r -a sf <<<"$st" &&
        printf 'process\t%s\t%s\tnested holder\n' "$BASHPID" "${sf[19]}" >>"$own"; } 2>/dev/null || true
    fi
    while :; do
      { mkdir -p "$S" && date +%s >>"$S/hold-beats"; } || true
      id="$(docker create --network none busybox:1.37.0 true 2>/dev/null)" || id=""
      if [[ -n "$id" ]]; then
        { printf '%s\n' "$id" >>"$S/hold-ids"; } 2>/dev/null || true
        if [[ -n "$own" ]]; then { printf 'container\t%s\tnested holder\n' "$id" >>"$own"; } 2>/dev/null || true; fi
      fi
      sleep 2
    done ) &
  printf '%s\n' "$!" >"$S/holder.pid"
  for ((i = 0; i < 300; i++)); do if [[ -s "$S/hold-ids" ]]; then break; fi; sleep 0.1; done
  : >"$S/holding"
  sleep 600
  exit 0
fi
if has own-hold; then
  # G051 VR1/VR2: one owned standalone container through the real harness helper, then hold (the control then
  # KILLs validate.sh and this group: no finalization).
  # shellcheck source=/dev/null
  source "$SELF_REPO/infra/validation/lib.sh"
  OWNED_REG="$OWNED"
  id1="$(docker_run_owned controls-hold -d --label "g001.run=$RUN_ID" --network none busybox:1.37.0 sleep 300)" || id1=""
  printf '%s\n' "$id1" >"$S/standalone-id"
  : >"$S/harness-running"; sleep 600; exit 0
fi
if has runner-hold; then
  # G064 VR9: the batch runner through the real infra/validation/lib.sh: it_runner_start passes its precondition through
  # this run's registry and pending marker, creates the runner with docker_run_owned (cidfile, standalone row) and checks
  # its baseline; one process is then exec'd into the runner, and the stub holds (the control KILLs validate.sh and this
  # group). busybox has no go, so only `exec <runner> go env GOCACHE GOTELEMETRYDIR` is answered here; every other call
  # reaches the docker on PATH. The runner ID goes to standalone-id (vr_crash), its batch label value to runner-batch.
  # shellcheck source=/dev/null
  source "$SELF_REPO/infra/validation/lib.sh"
  OWNED_REG="$OWNED" REPO="$SELF_REPO" RESULTS="$WORK/runner-results.tsv"
  docker() {
    if [[ "${1:-}" == exec && "${*:3}" == "go env GOCACHE GOTELEMETRYDIR" ]]; then printf '/root/.cache/go-build\n/root/.config/go/telemetry\n'; return 0; fi
    command docker "$@"
  }
  want="$(docker image inspect -f '{{.Id}}' busybox:1.37.0 2>/dev/null)" || want=""
  printf 'vr9-%s\n' "$RUN_ID" >"$S/runner-batch"
  if it_runner_start vr9 busybox:1.37.0 none "$want"; then
    id1="$(cat "$WORK/it-runner/vr9.id")"
    printf '%s\n' "$id1" >"$S/standalone-id"
    docker exec -d "$id1" sleep 300 || echo "exec failed" >"$S/runner-why"
  else
    printf '%s\n' "$ITR_WHY" >"$S/runner-why"
  fi
  : >"$S/harness-running"; sleep 600; exit 0
fi
if has sleep || has ignore-term; then : >"$S/harness-running"; sleep 600; exit 0; fi

# G064 (Plan r2 §5.3): the per-test Go evidence of the executed migrations and ops suites, shaped like run.sh's go_it
# exports (go-tests-expected.tsv, go-tests.tsv with the 15-column header, go-tests-junit.xml, db-lifecycle.tsv and one
# go-tests/<invocation>.log each): one passing invocation per suite. go-tests-fail-interrupted (GT2) declares two
# migrations invocations instead: stub-mig-fail failed (rc 1, a failing test, its log holding the planted generated
# secret) and stub-mig-int interrupted (rc 255, one unfinished test), each with counts consistent with its rows.
go_tests_reports() {
  local s n
  mkdir -p "$REPORT/go-tests"
  printf 'suite\tinvocation\n' >"$REPORT/go-tests-expected.tsv"
  printf 'kind\tinvocation\tpackage\ttest\tstatus\trc\texpected\tpassed\tfailed\tunfinished\telapsed_s\tlog_file\tline_from\tline_to\tdetail\n' >"$REPORT/go-tests.tsv"
  printf '<?xml version="1.0" encoding="UTF-8"?>\n<testsuites>\n' >"$REPORT/go-tests-junit.xml"
  printf 'ts_utc\tkind\tsubject\tfield\tvalue\n%s\ttemplate\tstub_tpl\tcreated\tstub\n' "$(date -u +%Y-%m-%dT%H:%M:%S.000Z)" >"$REPORT/db-lifecycle.tsv"
  for s in migrations ops; do
    if on "$s"; then :; else continue; fi
    if [[ "$s" == migrations ]] && has go-tests-fail-interrupted; then
      printf 'migrations\tstub-mig-fail\nmigrations\tstub-mig-int\n' >>"$REPORT/go-tests-expected.tsv"
      printf '%s\n' \
        $'invocation\tstub-mig-fail\tstub/pkg\t-\tfail\t1\t1\t0\t1\t0\t-\tgo-tests/stub-mig-fail.log\t1\t3\tfindings=package-failed:stub/pkg' \
        $'package\tstub-mig-fail\tstub/pkg\t-\tfail\t-\t1\t0\t1\t0\t0.010\tgo-tests/stub-mig-fail.log\t1\t3\tfindings=package-failed:stub/pkg' \
        $'test\tstub-mig-fail\tstub/pkg\tTestStubFail\tfail\t-\t-\t-\t-\t-\t0.010\tgo-tests/stub-mig-fail.log\t1\t2\tparent=' \
        $'invocation\tstub-mig-int\tstub/pkg\t-\tinterrupted\t255\t1\t0\t0\t1\t-\tgo-tests/stub-mig-int.log\t1\t1\tfindings=interrupted' \
        $'package\tstub-mig-int\tstub/pkg\t-\tmissing\t-\t1\t0\t0\t1\t-\tgo-tests/stub-mig-int.log\t1\t1\tfindings=' \
        $'test\tstub-mig-int\tstub/pkg\tTestStubLong\tunfinished\t-\t-\t-\t-\t-\t-\tgo-tests/stub-mig-int.log\t1\t1\tparent=' \
        >>"$REPORT/go-tests.tsv"
      printf '%s\n' '<testsuite name="stub-mig-fail" tests="1" failures="1" errors="0" skipped="0">' \
        '<testcase classname="stub/pkg" name="TestStubFail" time="0.010"><failure message="fail"/></testcase>' '</testsuite>' \
        '<testsuite name="stub-mig-int" tests="1" failures="0" errors="1" skipped="0">' \
        '<testcase classname="stub/pkg" name="TestStubLong"><error message="unfinished"/></testcase>' '</testsuite>' >>"$REPORT/go-tests-junit.xml"
      printf '=== RUN   TestStubFail\n    stub_test.go:1: leak %s\n--- FAIL: TestStubFail (0.01s)\n' "$canary" >"$REPORT/go-tests/stub-mig-fail.log"
      printf '=== RUN   TestStubLong\n' >"$REPORT/go-tests/stub-mig-int.log"
      continue
    fi
    n="stub-$s-it"
    printf '%s\t%s\n' "$s" "$n" >>"$REPORT/go-tests-expected.tsv"
    printf 'invocation\t%s\tstub/pkg\t-\tpass\t0\t1\t1\t0\t0\t-\tgo-tests/%s.log\t1\t3\tfindings=\n' "$n" "$n" >>"$REPORT/go-tests.tsv"
    printf 'package\t%s\tstub/pkg\t-\tpass\t-\t1\t1\t0\t0\t0.010\tgo-tests/%s.log\t1\t3\tfindings=\n' "$n" "$n" >>"$REPORT/go-tests.tsv"
    printf 'test\t%s\tstub/pkg\tTestStub\tpass\t-\t-\t-\t-\t-\t0.010\tgo-tests/%s.log\t1\t2\tparent=\n' "$n" "$n" >>"$REPORT/go-tests.tsv"
    printf '<testsuite name="%s" tests="1" failures="0" errors="0" skipped="0">\n<testcase classname="stub/pkg" name="TestStub" time="0.010"/>\n</testsuite>\n' "$n" >>"$REPORT/go-tests-junit.xml"
    printf '=== RUN   TestStub\n--- PASS: TestStub (0.01s)\nok  \tstub/pkg\t0.010s\n' >"$REPORT/go-tests/$n.log"
  done
  printf '</testsuites>\n' >>"$REPORT/go-tests-junit.xml"
}

write_reports() {
  local id suite d
  {
    printf 'PREFLIGHT\tINFO\tstub harness\n'
    while IFS=$'\t' read -r id suite; do
      if [[ -z "$id" || "$id" == '#'* ]]; then continue; fi
      if on "$suite" && [[ "$id" != "${STUB_DROP_ID:-}" ]]; then :; else continue; fi
      if [[ "$id" == "${STUB_FAIL_ID:-}" ]]; then printf '%s\tFAIL\tstub failure; failed-subtests: OBJ-SIGN\n' "$id"; continue; fi
      printf '%s\tPASS\tstub %s\n' "$id" "$suite"
    done <"$SELF_REPO/infra/validation/required-checks.tsv"
    if [[ -n "${STUB_ADD_ROW:-}" ]]; then printf '%s\tPASS\tstub extra row\n' "$STUB_ADD_ROW"; fi
    printf 'IMAGES\tINFO\tbusybox:1.37.0\n'
  } >"$REPORT/results.tsv"
  printf 'busybox:1.37.0\n' >"$REPORT/images.txt"
  printf 'suite\tstub\t1\t2\n' >"$REPORT/timings.tsv"
  # J1/J5 (2026-10-19 timing repair): executed fast suites with overlapping timings, 60 + 80 + 80 = 220 s over a
  # 100 s union.
  if [[ -n "${STUB_SUITE_TIMES:-}" ]]; then
    printf 'suite\tlint\t100\t160\nsuite\tcontracts\t110\t190\nsuite\tunit\t120\t200\n' >>"$REPORT/timings.tsv"
  fi
  printf 'seq\tjob\tsuite\tkind\tdockerfile\tdockerfile_sha256\ttarget\tfrontend\tbases\tplatform\tignore_file\tignore_sha256\tcontext_inventory\tnamed_contexts\targs_digest\tdaemon\tprogress\tsteps_executed\tsteps_cached\tresult\timage_id\n' >"$REPORT/builds.tsv"
  printf 'expected\t0\nrows\t0\nunknown\t0\nduplicates\t0\nnear_duplicates\t0\nexceptions\t0\nbootstrap\tnot-applicable\nac5_status\tMET\n' >"$REPORT/ledger-summary.tsv"
  if on ci-controls; then printf 'V1\t1\n' >"$REPORT/ci-controls-timings.tsv"; fi
  for d in root off rename feature; do
    if [[ "$d" == feature ]]; then if on feature-web; then :; else continue; fi; elif on "$d"; then :; else continue; fi
    printf '{"config":{"version":"stub"},"suites":[{"title":"s","specs":[{"title":"shell [S-STUB]","tests":[{"projectName":"p","status":"expected","annotations":[{"type":"S-ERRORS"}],"results":[{"status":"passed","attachments":[{"name":"screenshot","contentType":"image/png","body":"%s"}]}]}]}]}],"stats":{"expected":2,"skipped":0,"flaky":0,"unexpected":0}}\n' \
      "$(printf 'attachment-%s' "$d" | base64)" >"$REPORT/browser-$d-results.json"
    printf 'p\texpected\tshell [S-STUB]\tS-ERRORS\n' >"$REPORT/browser-$d-tests.tsv"
  done
  # Normal contract reports from the declared synthetic corpus ("id valid class" per case); the
  # contracts-* modes alter one report each, keeping its counts consistent unless the mode is about them.
  local cases q t sk mm
  for d in go python; do
    if on contracts; then :; else break; fi
    cases="$(sed -n 's/^ *{"id": "\([A-Za-z0-9]*\)", .*"valid": \(true\|false\), "class": "\([a-z-]*\)"}.*$/\1 \2 \3/p' \
      "$SELF_REPO/contracts/conformance/intelligence/expectations.json")"
    q=0 t=0 sk=0 mm=0
    case "$d:$MODE" in
      go:*,contracts-empty,*) cases="" ;;
      python:*,contracts-truncated,*) cases="$(head -n -1 <<<"$cases")" ;;
      go:*,contracts-duplicate,*) cases="$(awk 'NR == 2 {print} {print}' <<<"$cases")" ;;
      python:*,contracts-verdict,*) cases="$(sed 's/^SY02 false [a-z-]*$/SY02 true valid/' <<<"$cases")" ;;
      go:*,contracts-string,*) q=1 ;;
      python:*,contracts-count,*) t=1 ;;
      go:*,contracts-skipped,*) sk=1 ;;
      python:*,contracts-mismatch,*) mm=1 ;;
    esac
    if [[ -n "$cases" ]]; then printf '%s\n' "$cases"; fi | awk -v v="$d" -v q="$q" -v t="$t" -v sk="$sk" -v mm="$mm" '
      { n++; id[n] = $1; val[n] = $2; cls[n] = $3; if ($2 == "true") nv++; else ni++; bc[$3]++ }
      END {
        printf "{\"validator\":\"stub-%s 1\",\"results\":[", v
        for (i = 1; i <= n; i++) printf "%s{\"id\":\"%s\",\"valid\":%s,\"class\":\"%s\"}", (i > 1 ? "," : ""), id[i], (q ? "\"" val[i] "\"" : val[i]), cls[i]
        printf "],\"counts\":{\"total\":%d,\"valid\":%d,\"invalid\":%d,\"by_class\":{", n + t, nv, ni
        s = ""; for (c in bc) { printf "%s\"%s\":%d", s, c, bc[c]; s = "," }
        printf "}},\"skipped\":%d", sk
        if (v == "python") printf ",\"mismatches\":[%s]", (mm ? "{\"id\":\"SY02\"}" : "")
        printf "}\n"
      }' >"$REPORT/contracts-results-$d.json"
    printf '{"validator":"stub-%s 1","results":[{"id":"SN01","class":"schema-missing","expected_class":"schema-missing","match":true}],"derive":"pass"}\n' "$d" >"$REPORT/schema-negative-$d.json"
  done
  if on contracts; then printf 'SY01\tvalid\tvalid\n' >"$REPORT/contracts-agree.tsv"; fi
  if on root; then for d in core web intelligence; do printf 'X1\tPASS\tcategory\n' >"$REPORT/error-output-$d.tsv"; done; fi
  if on ops; then printf 'O1\tPASS\n' >"$REPORT/obs-context.tsv"; printf 'O1\t120\n' >"$REPORT/obs-timings.tsv"; fi
  if on controls; then
    printf 'W1 PASS\n' >"$REPORT/wrapper-controls.log"
    printf 'C1\tPASS\n' >"$REPORT/deps-controls.tsv"
  fi
  if on ci-controls; then
    if [[ "${STUB_FAIL_ID:-}" == B-CI-CONTROLS ]]; then
      printf 'FAIL LR-PROBE retained failure detail\n' >"$REPORT/ci-controls.log"
    else printf 'PASS V1 stub\n' >"$REPORT/ci-controls.log"; fi
  fi
  printf 'DEPS_STUB x\n' >"$REPORT/deps-codes.txt"
  for d in plan.out repeat-render.out catalog-p.log deps-server.txt; do printf 'stub %s\n' "$d" >"$REPORT/$d"; done
  if on migrations; then printf 'stub migrate-it.log\n' >"$REPORT/migrate-it.log"; fi
  if on migrations || on ops; then go_tests_reports; fi
  : >"$REPORT/inventory-leftover.txt"
  # G064 corr1 C1: run.sh's resources.tsv (validate.sh requires it whenever the harness ran with a registry): no finding.
  printf '# stub resources.tsv\nrow\tscope\tclass\tcreations\tobserved_max\tceiling\tfinal_leftover\tunaccounted\tstatus\ntotal\trun\tcontainer\t0\t0\t63\t0\t0\tPASS\n' \
    >"$REPORT/resources.tsv"
  if [[ -n "${STUB_RM_REPORT:-}" ]]; then rm -f -- "$REPORT/$STUB_RM_REPORT"; fi
}
if has stray; then write_reports; : >"$S/harness-running"; exit 0; fi
kill "$g1" "$g2" 2>/dev/null || true
wait "$g1" "$g2" 2>/dev/null || true
write_reports
second="$(awk -F'\t' '$2 == "PASS" {n++; if (n == 2) {print $1; exit}}' "$REPORT/results.tsv")"
last="$(awk -F'\t' '$2 == "PASS" {l = $1} END {print l}' "$REPORT/results.tsv")"
rc=0
if has fail || has fail-row; then sed -i "s/^$second\tPASS\t.*/$second\tFAIL\tstub failure; failed-subtests: OBJ-SIGN/" "$REPORT/results.tsv"; fi
if has fail; then rc=1; fi
if has missing-id; then sed -i "/^$last\t/d" "$REPORT/results.tsv"; fi
if has unmet; then sed -i "s/^$last\tPASS\t.*/$last\tUNMET\tstub prerequisite/" "$REPORT/results.tsv"; fi
if has malformed; then printf 'X-BAD\tPASS\ttoo\tmany\n' >>"$REPORT/results.tsv"; fi
# G050 J4a/J4c/J5: a row with an unknown raw status; a PASS row of a non-required ID twice (completeness allows it).
if has unknown-status; then printf 'B-J4A-ODD\tSKIPPED\tstub\n' >>"$REPORT/results.tsv"; fi
if has dup-extra; then printf 'B-J4C-EXTRA\tPASS\tstub\nB-J4C-EXTRA\tPASS\tstub\n' >>"$REPORT/results.tsv"; fi
if has browser-skipped; then sed -i 's/"skipped":0/"skipped":1/' "$REPORT/browser-root-results.json"; fi
if has browser-malformed; then printf '{"stats": {"expected": 2,' >"$REPORT/browser-root-results.json"; fi
if has partial; then rm -f "$REPORT/deps-server.txt"; mkdir "$REPORT/deps-server.txt"; printf 'x\n' >"$REPORT/deps-server.txt/inner"; fi
if has locked; then mkdir -p "$REPORT/locked"; printf 'x\n' >"$REPORT/locked/f"; chmod 555 "$REPORT/locked"; fi
if has unreadable; then chmod 000 "$REPORT/migrate-it.log"; fi
if has plant-retained; then printf 'leak %s\n' "$canary" >>"$REPORT/plan.out"; fi
# G051 C6: the canary inside a retained build-ledger diagnostic line (an unknown vertex name keeps Dockerfile text).
if has plant-ledger-context; then printf 'unknown_context\tcall 1 session 1: #24 [web deps 1/4] RUN echo %s\n' "$canary" >>"$REPORT/ledger-summary.tsv"; fi
if has plant-console; then echo "console leak $canary"; fi
if has plant-staging-key; then
  key="sk_live_$(head -c 16 /dev/urandom | od -An -tx1 | tr -d ' \n')"
  printf '%s\n' "$key" >>"$SECRETS"; printf '%s\n' "$key" >"$S/staging-key"
  printf 'provenance key %s\n' "$key" >>"$REPORT/plan.out"
fi
if has staging-metadata; then
  printf 'provenance client_id=client_01CONTROLSFAKE issuer=https://api.workos.com/user_management/client_01CONTROLSENVFAKE\n' >>"$REPORT/plan.out"
fi
if has plant-gitleaks; then
  printf 'token ghp_%s\n' "$(head -c 400 /dev/urandom | tr -dc 'A-Za-z0-9' | head -c 36)" >>"$REPORT/catalog-p.log"
fi
if has drift-tracked; then printf 'drift\n' >>"$SELF_REPO/README"; fi
if has drift-untracked; then printf 'drift\n' >"$SELF_REPO/untracked-drift.txt"; fi
if has block-result; then mkdir -p "$(dirname "$(dirname "$REPORT")")/state/harness.result"; fi
if has replay; then cp "$STUB_REPLAY" "$REPORT/results.tsv"; rc=1; fi
# G050 J4a: results.tsv missing at the end of an otherwise successful run (exit 0).
if has rm-results; then rm -f "$REPORT/results.tsv"; fi
echo "harness console: run $RUN_ID done ($rc)"
exit "$rc"
STUB
  chmod 755 "$t/infra/ci/"*.sh "$t/infra/validation/run.sh"
  git -C "$t" init -q -b main
  git -C "$t" add -A
  GIT_AUTHOR_DATE=2026-01-01T00:00:00Z GIT_COMMITTER_DATE=2026-01-01T00:00:00Z \
    git -C "$t" -c user.name=ci-controls -c user.email=controls@example.invalid commit -q -m "scratch tree"
}

# --- per-control helpers --------------------------------------------------------------------------------
C="" PROB="" DET="" CTL_ID="" LB=() RC=0 VPID=""
prob() { printf '%s\n' "$*" >>"$PROB"; }
detail() { printf '%s' "$*" >"$DET"; }
# Every control resource carries the control and run labels and the owner key (G064 correction 11 item 4; startup guard).
case_labels() { local l="sdlc.ci.control=$RT-$CTL_ID $RUN_LABEL $OWNER_LABEL" e; for e in "${EXTRA_LABELS[@]}"; do l+=" $e"; done; printf '%s' "$l"; }

new_case() { # a fresh scratch git tree in $C/repo
  C="$ROOT/cases/$CTL_ID"
  rm -rf -- "$C"
  mkdir -p "$C/tmp" "$C/stub"
  cp -a -- "$TEMPLATE" "$C/repo"
  : >"$C/stub/harness-invocations"
  : >"$C/stub/preflight-invocations"
}
head_sha() { git -C "$C/repo" rev-parse HEAD; }
ci_env() { # JOB_ID -> CIENV (a branch-push pipeline on the default branch)
  CIENV=(GITLAB_CI=true CI_PROJECT_DIR="$C/repo" CI_JOB_ID="$1" CI_PIPELINE_ID=1001 CI_PIPELINE_IID=7
    CI_PIPELINE_SOURCE=push CI_COMMIT_SHA="$(head_sha)" CI_COMMIT_BRANCH=main CI_COMMIT_REF_NAME=main
    CI_DEFAULT_BRANCH=main CI_JOB_TIMEOUT=10800 RUNNER_SCRIPT_TIMEOUT=2h45m CI_SERVER_VERSION=18.4.0
    CI_JOB_STARTED_AT="$(ci_utc)")
}
# As root, every validate.sh launch runs without CAP_DAC_OVERRIDE and CAP_DAC_READ_SEARCH (an unprivileged
# user's file-mode semantics); an unprivileged runner needs nothing.
NOCAP=()
if [[ "$(id -u)" == 0 ]]; then NOCAP=(setpriv "--bounding-set=-dac_override,-dac_read_search" --); fi
vcmd() { # ENV... -- ARGS... -> VCMD (validate.sh in the case tree, clean environment)
  local envs=()
  while (( $# > 0 )) && [[ "$1" != -- ]]; do envs+=("$1"); shift; done
  if (( $# > 0 )); then shift; fi
  case_state
  VCMD=(env -i -C "$C/repo" "${BASE_ENV[@]}" G003_STATE_ROOT="$CSTATE" TMPDIR="$C/tmp" STUB_DIR="$C/stub" SHIM_LOG="$C/shim.log" SHIM_PULL_NOOP=1
    SHIM_LABELS="$(case_labels)" "${envs[@]}" "${NOCAP[@]}" bash "$C/repo/infra/ci/validate.sh" "$@")
}
run_v() { # ENV... -- ARGS...: run validate.sh (a tracked launch, bound 900 s) to its end; RC and the final line
  vcmd "$@"
  launch "$C/out" "${VCMD[@]}"
  await "$LPID" 900
  parse_final
}
start_v() { # ENV... -- ARGS...: start validate.sh as a tracked launch, its own process group (VPID)
  vcmd "$@"
  launch "$C/out" "${VCMD[@]}"
  VPID="$LPID"
}
wait_file() { # FILE SECONDS
  local i
  for ((i = 0; i < $2 * 10; i++)); do if [[ -e "$1" ]]; then return 0; fi; sleep 0.1; done
  return 1
}
wait_v() { # wait for the started validate.sh (bound 400 s); RC and the parsed final line
  await "$VPID" 400
  parse_final
}
declare -A PH=()
FINAL_LINE="" F_PUB="" F_REG="" F_STOP=""
parse_final() {
  local kv
  FINAL_LINE="$(grep '^validate: final status ' "$C/out" | tail -1)" || FINAL_LINE=""
  F_PUB="$(sed -n 's/.*; published=\([a-z]*\);.*/\1/p' <<<"$FINAL_LINE")"
  F_REG="$(sed -n 's/.*; registry=\([^;]*\);.*/\1/p' <<<"$FINAL_LINE")"
  F_STOP="$(sed -n 's/.*; harness_stop=\([a-z]*\);.*/\1/p' <<<"$FINAL_LINE")"
  PH=()
  for kv in ${FINAL_LINE#*phases:}; do PH[${kv%%=*}]="${kv#*=}"; done
}
PUBDIR() { printf '%s' "$C/repo/ci-evidence"; }

expect_rc() { if [[ "$RC" != "$1" ]]; then prob "exit $RC, want $1"; fi; if [[ -z "$FINAL_LINE" ]]; then prob "no final line"; fi; }
expect_phase() { if [[ "${PH[$1]-<absent>}" != "$2" ]]; then prob "phase $1=${PH[$1]-<absent>}, want $2"; fi; }
expect_no_phase() { if [[ -n "${PH[$1]+x}" ]]; then prob "phase $1=${PH[$1]} recorded, want none"; fi; }
expect_f1f2() { if [[ -z "${PH[postcheck]+x}" || -z "${PH[cleanup]+x}" ]]; then prob "postcheck or cleanup did not run"; fi; }
expect_runs() { local n; n="$(grep -c . "$C/stub/harness-invocations" || true)"; if [[ "$n" != "$1" ]]; then prob "harness invocations $n, want $1"; fi; }
expect_stop() { if [[ "$F_STOP" != "$1" ]]; then prob "harness_stop=$F_STOP, want $1"; fi; }
published_list() { if [[ -d "$(PUBDIR)" ]]; then (cd "$(PUBDIR)" && find . -mindepth 1 -printf '%P\n' | LC_ALL=C sort | tr '\n' ' '); fi; }
expect_pub() { # none | full | minimal:MARKER
  local d; d="$(PUBDIR)"
  case "$1" in
    none)
      if [[ "$F_PUB" != none ]]; then prob "published=$F_PUB, want none"; fi
      if [[ -e "$d" || -e "$d.tmp" ]]; then prob "a publication directory exists"; fi ;;
    full)
      if [[ "$F_PUB" != full ]]; then prob "published=$F_PUB, want full"; fi
      if [[ ! -f "$d/summary.json" || ! -f "$d/phase-status.tsv" ]]; then prob "full set lacks summary.json or phase-status.tsv"; fi
      if [[ -e "$d/WITHHELD" || -e "$d/STAGING-FAILED" ]]; then prob "full set carries a marker"; fi
      if [[ -e "$d/harness-console.log" || -e "$d/deps-codes.txt" || -e "$d/cleanup.log" ]]; then prob "an excluded file was published"; fi
      if grep -q '"status_kind": "pre-publication"' "$d/summary.json" 2>/dev/null; then :; else prob "summary.json is not labelled pre-publication"; fi ;;
    minimal:*)
      if [[ "$F_PUB" != minimal ]]; then prob "published=$F_PUB, want minimal"; fi
      if [[ "$(published_list)" != "${1#minimal:} junit.xml phase-status.tsv summary.json " ]]; then prob "minimal set is [$(published_list)], want ${1#minimal:}, junit.xml, phase-status.tsv, summary.json"; fi ;;
  esac
}
expect_no_canary() {
  if [[ -s "$C/stub/canary" && -d "$(PUBDIR)" ]] && grep -rqF -f "$C/stub/canary" "$(PUBDIR)"; then prob "the canary was published"; fi
}
# The published results.tsv holds exactly one FAIL row and it carries the stub's closed failed-subtest code verbatim
# (G014 object diagnostic publication: the code travels in the existing row through scan, stage and publish).
expect_fail_row_code() {
  local f; f="$(PUBDIR)/results.tsv"
  if [[ ! -f "$f" ]]; then prob "no published results.tsv"; return; fi
  if [[ "$(grep -c $'\tFAIL\t' "$f")" != 1 ]]; then prob "published results.tsv does not hold exactly one FAIL row"; fi
  if grep -qE $'^[A-Z][A-Z0-9-]*\tFAIL\tstub failure; failed-subtests: OBJ-SIGN$' "$f"; then :; else prob "the published FAIL row lacks failed-subtests: OBJ-SIGN"; fi
}
expect_group_gone() {
  local hp="" pg="" g1="" g2="" p
  if [[ -f "$C/stub/harness.pids" ]]; then read -r hp pg g1 g2 <"$C/stub/harness.pids" || true; fi
  for p in $hp $g1 $g2; do if kill -0 "$p" 2>/dev/null; then prob "harness process $p is alive"; fi; done
  if [[ -n "$pg" ]] && kill -0 -- "-$pg" 2>/dev/null; then prob "harness process group $pg is alive"; fi
  if ps -eo args= | grep -F -- "$C/" | grep -vq -- 'grep -F'; then prob "a process of this case is still running"; fi
}
expect_trap() { if [[ "$1" == yes && ! -e "$C/stub/harness-exit-trap" ]]; then prob "the harness EXIT trap did not run"; fi
  if [[ "$1" == no && -e "$C/stub/harness-exit-trap" ]]; then prob "the harness EXIT trap ran"; fi; }
expect_line() { # has|not TEXT
  if [[ "$1" == has && " ${FINAL_LINE#*phases:} " != *" $2 "* ]]; then prob "final line lacks $2"; fi
  if [[ "$1" == not && " ${FINAL_LINE#*phases:} " == *" $2"* ]]; then prob "final line has $2"; fi
}
# G064 D3: a local run's default evidence directory lies under the case state root, a CI run's under TMPDIR.
no_validate_left() { case_state; if compgen -G "$C/tmp/g003-validate-*" >/dev/null || compgen -G "$CSTATE/g003-validate-*" >/dev/null; then prob "an evidence directory was left in TMPDIR or the state root"; fi; }

# --- group A: gate semantics ---------------------------------------------------------------------------
# G051: builds.tsv, ledger-summary.tsv (every stub run) and ci-controls-timings.tsv (controls) from the harness;
# ledger-bootstrap.tsv from validate.sh whenever the harness ran.
FULL_SET="$(printf '%s\n' attribution.json browser-off-results.json browser-off-tests.tsv browser-rename-results.json \
  browser-rename-tests.tsv browser-root-results.json browser-root-tests.tsv builds.tsv catalog-p.log ci-controls.log \
  ci-controls-timings.tsv cleanup-inventory.tsv storage.tsv \
  contracts-agree.tsv contracts-results-go.json contracts-results-python.json deps-controls.tsv deps-server.txt \
  error-output-core.tsv error-output-intelligence.tsv error-output-web.tsv harness.log images.txt inventory-leftover.txt junit.xml \
  resources.tsv db-lifecycle.tsv go-tests go-tests/stub-migrations-it.log go-tests/stub-ops-it.log go-tests-expected.tsv go-tests-junit.xml go-tests.tsv \
  ledger-bootstrap.tsv ledger-summary.tsv migrate-it.log \
  obs-context.tsv obs-timings.tsv phase-status.tsv plan.out repeat-render.out required-checks.tsv results.tsv \
  runner-preflight.tsv schema-negative-go.json schema-negative-python.json selection.tsv summary.json timings.tsv \
  wrapper-controls.log | LC_ALL=C sort | tr '\n' ' ')"
# bootstrap_local: the published ledger-bootstrap.tsv outside the hosted context is the header and one row of none
# with the verdict not-applicable (G051 §3.8.3).
bootstrap_local() {
  local f; f="$(PUBDIR)/ledger-bootstrap.tsv"
  if [[ "$(cat "$f" 2>/dev/null)" != $'ref\tpre_id\tpost_id\tdockerfile_sha256\tbases\tplatform\tverdict\nnone\tnone\tnone\tnone\tnone\tnone\tnot-applicable' ]]; then
    prob "ledger-bootstrap.tsv is not the local not-applicable record: [$(tr '\t\n' ' ;' <"$f" 2>/dev/null)]"
  fi
}
std_ci() { new_case; ci_env "${1:-4242}"; }

ctl_V1() {
  std_ci; run_v "${CIENV[@]}" -- --publish-dir ci-evidence
  expect_rc 0; expect_pub full; expect_f1f2; expect_runs 1; expect_no_canary; expect_line not internal=
  if [[ "$(published_list)" != "$FULL_SET" ]]; then prob "published set differs: [$(published_list)]"; fi
  if grep -q attachments "$(PUBDIR)"/browser-*-results.json 2>/dev/null; then prob "browser JSON still has attachments"; fi
  if grep -q '"candidate": "full"' "$(PUBDIR)/summary.json"; then :; else prob "summary.json candidate is not full"; fi
  if grep -q '^validate: check B-SHELL PASS$' "$C/out"; then :; else prob "the job log lacks the check columns"; fi
  if [[ "$F_REG" != removed ]]; then prob "registry=$F_REG, want removed"; fi
  bootstrap_local
  detail "exit $RC; published full ($(published_list | wc -w) files, no attachments, no console)"
}
ctl_V2() {
  std_ci; run_v "${CIENV[@]}" STUB_HARNESS=fail -- --publish-dir ci-evidence
  expect_rc 1; expect_pub full; expect_f1f2; expect_line has harness=1
  if grep -qP '\tFAIL\t' "$(PUBDIR)/results.tsv" 2>/dev/null; then :; else prob "the published results lack the FAIL row"; fi
  detail "exit $RC; published full with the FAIL row"
}
ctl_V3() { std_ci; run_v "${CIENV[@]}" STUB_HARNESS=fail-row -- --publish-dir ci-evidence; expect_rc 14; expect_phase complete 14; expect_pub full; expect_fail_row_code; detail "exit $RC (FAIL row under harness exit 0; failed-subtests code published)"; }

ctl_V4() {
  local x imp ok=1
  new_case; ctl_labels; imp="$C/impl"
  if [[ ! -f "$REPO_R/contracts/conformance/intelligence/expectations.json" ]]; then prob "prerequisite missing: contracts/conformance/intelligence/expectations.json"; ok=0; fi
  if [[ ! -f "$REPO_R/infra/validation/run.sh" ]] || ! grep -q 'suite_contracts' "$REPO_R/infra/validation/run.sh"; then prob "prerequisite missing: the run.sh contracts suite"; ok=0; fi
  if (( ok == 0 )); then return 0; fi
  if grep -q '"EI39"' "$REPO_R/contracts/conformance/intelligence/expectations.json"; then :; else prob "prerequisite missing: case EI39"; return 0; fi
  mkdir -p "$imp"
  (cd "$REPO_R" && git ls-files -z --cached --others --exclude-standard | while IFS= read -r -d '' x; do if [[ -f "$x" || -L "$x" ]]; then printf '%s\0' "$x"; fi; done | tar --null --no-recursion -cf - -T -) | tar -xf - -C "$imp"
  D run --rm --pull never --network none "${LB[@]}" -i "$CI_YQ_IMAGE" -p json -o json \
    '(.cases[] | select(.id == "EI39") | .valid) = true' <"$REPO_R/contracts/conformance/intelligence/expectations.json" \
    >"$imp/contracts/conformance/intelligence/expectations.json"
  x="$(D run --rm --pull never --network none "${LB[@]}" -i "$CI_YQ_IMAGE" -p json '.cases[] | select(.id == "EI39") | .valid' \
    <"$imp/contracts/conformance/intelligence/expectations.json")"
  if [[ "$x" != true ]]; then prob "EI39 could not be set to valid in the scratch copy"; return 0; fi
  git -C "$imp" init -q -b main
  git -C "$imp" add -A
  git -C "$imp" -c user.name=ci-controls -c user.email=controls@example.invalid commit -q -m "V4 scratch copy (EI39 valid)"
  launch "$C/v4.log" env -i -C "$imp" "${BASE_ENV[@]}" TMPDIR="$C/tmp" SHIM_LABELS="$(case_labels)" \
    bash "$imp/infra/validation/run.sh" --base HEAD --allow-file "$imp/infra/ci/unchanged.allow" \
    --report-dir "$C/v4-report" contracts
  await "$LPID" 3600
  if [[ "$RC" != 1 ]]; then prob "nested run.sh contracts exited $RC, want 1"; fi
  for x in B-CONTRACT-GO B-CONTRACT-PY B-CONTRACT-AGREE; do
    if awk -F'\t' -v i="$x" '$1 == i && $2 == "FAIL" {f = 1} END {exit !f}' "$C/v4-report/results.tsv" 2>/dev/null; then :; else prob "$x is not FAIL in the nested run"; fi
  done
  if grep -q 'EI39' "$C/v4.log" "$C/v4-report/results.tsv" 2>/dev/null; then :; else prob "the nested run does not name EI39"; fi
  cp -- "$C/v4-report/results.tsv" "$C/replay.tsv" 2>/dev/null || { prob "no nested results.tsv to replay"; return 0; }
  ci_env 4243
  run_v "${CIENV[@]}" STUB_HARNESS=replay STUB_REPLAY="$C/replay.tsv" -- --publish-dir ci-evidence
  expect_rc 1; expect_line has harness=1; expect_pub full
  for x in B-CONTRACT-GO B-CONTRACT-PY B-CONTRACT-AGREE; do
    if awk -F'\t' -v i="$x" '$1 == i && $2 == "FAIL" {f = 1} END {exit !f}' "$(PUBDIR)/results.tsv" 2>/dev/null; then :; else prob "replayed publication lacks $x FAIL"; fi
  done
  detail "nested contracts exit 1 (GO, PY, AGREE FAIL, EI39); replay exit $RC, published full"
}
ctl_V7() { std_ci; run_v "${CIENV[@]}" STUB_HARNESS=fail -- --publish-dir ci-evidence; expect_rc 1; expect_phase postcheck 0; expect_pub full; expect_fail_row_code; detail "exit $RC; postcheck passed; failed-subtests code published"; }
ctl_V8() { std_ci; run_v "${CIENV[@]}" STUB_HARNESS=drift-tracked -- --publish-dir ci-evidence; expect_rc 13; expect_phase postcheck 13; expect_pub full; detail "exit $RC (tracked file modified)"; }
ctl_V9() { std_ci; run_v "${CIENV[@]}" STUB_HARNESS=drift-untracked -- --publish-dir ci-evidence; expect_rc 13; expect_phase postcheck 13; detail "exit $RC (untracked file created)"; }
ctl_V10() { std_ci; run_v "${CIENV[@]}" STUB_HARNESS=missing-id -- --publish-dir ci-evidence; expect_rc 14; expect_phase complete 14; detail "exit $RC (required ID missing)"; }
ctl_V11() {
  local r1
  std_ci; run_v "${CIENV[@]}" STUB_HARNESS=unmet -- --publish-dir ci-evidence; expect_rc 14; r1="$RC"
  std_ci 4244; run_v "${CIENV[@]}" STUB_HARNESS=browser-skipped -- --publish-dir ci-evidence; expect_rc 14
  detail "UNMET row exit $r1; browser skipped=1 exit $RC"
}
ctl_V12() { std_ci; run_v "${CIENV[@]}" STUB_HARNESS=fail,drift-untracked -- --publish-dir ci-evidence; expect_rc 1; expect_phase postcheck 13; detail "exit $RC; drift reported (postcheck=13)"; }
# Contract-report completeness against the scratch tree's declared synthetic corpus (SY01 valid, SY02
# schema-invalid, SY03 invalid-json). Every fault keeps every harness row PASS and harness exit 0.
# complete_findings: the job log's completeness findings; do_complete's Q6 diagnostic lines (G064 corr1 C6,
# "complete: diagnostic: ...", printed on every run and never a status) are not findings.
complete_findings() { grep '^validate: complete: ' "$C/out" | grep -v '^validate: complete: diagnostic:' || true; }
c_report() { # MODE MESSAGE: one altered normal contract report -> 14 for completeness, named in the job log
  std_ci; run_v "${CIENV[@]}" STUB_HARNESS="$1" -- --publish-dir ci-evidence
  expect_rc 14; expect_phase complete 14; expect_line has harness=0; expect_pub full
  if grep -qxF "validate: complete: $2" "$C/out"; then :; else prob "$1: the job log lacks [complete: $2]"; fi
  if [[ "$(complete_findings | grep -c .)" != 1 ]]; then prob "$1: want exactly one completeness finding"; fi
  if [[ -f "$(PUBDIR)/results.tsv" ]] && awk -F'\t' '$2 != "PASS" && $2 != "INFO" {f = 1} END {exit !f}' "$(PUBDIR)/results.tsv"; then
    prob "$1: a harness row is not PASS"
  fi
}
ctl_V13a() { c_report contracts-empty "contracts-results-go.json is not the declared case set (a case is missing, extra or has another verdict)"
  detail "exit $RC; an empty Go case set (skipped 0, counts 0) is not the declared corpus; harness rows PASS, harness=0"; }
ctl_V13b() { c_report contracts-truncated "contracts-results-python.json is not the declared case set (a case is missing, extra or has another verdict)"
  detail "exit $RC; a Python case set missing SY03 (counts consistent) is incomplete"; }
ctl_V13c() { c_report contracts-duplicate "contracts-results-go.json repeats a case ID"
  detail "exit $RC; a Go case set with every case plus SY02 twice (counts consistent) is rejected"; }
ctl_V13d() { c_report contracts-verdict "contracts-results-python.json is not the declared case set (a case is missing, extra or has another verdict)"
  detail "exit $RC; a Python report with SY02 valid instead of schema-invalid (counts consistent) is rejected"; }
ctl_V13e() {
  local d
  std_ci; run_v "${CIENV[@]}" -- --publish-dir ci-evidence
  expect_rc 0; expect_phase complete 0; expect_pub full
  if [[ -n "$(complete_findings)" ]]; then prob "a completeness finding was reported"; fi
  for d in go python; do
    if [[ "$(grep -o '"id":"SY0[1-3]"' "$(PUBDIR)/contracts-results-$d.json" 2>/dev/null | LC_ALL=C sort -u | tr '\n' ' ')" != '"id":"SY01" "id":"SY02" "id":"SY03" ' ]]; then
      prob "the published $d report is not the full declared case set"
    fi
  done
  detail "exit $RC; complete=0: both full reports of the declared corpus (3 unique typed cases, consistent counts, nothing skipped) pass"
}
ctl_V13f() {
  local m r=""
  for m in "contracts-string:contracts-results-go.json has a mistyped or malformed case" \
    "contracts-count:contracts-results-python.json counts are inconsistent with its cases" \
    "contracts-skipped:contracts-results-go.json skipped a case or reports a mismatch" \
    "contracts-mismatch:contracts-results-python.json skipped a case or reports a mismatch"; do
    c_report "${m%%:*}" "${m#*:}"; r+=" ${m%%:*}=$RC"
  done
  detail "each explicit check decides alone (exit 14):$r (string verdicts, counts.total off by one, skipped 1, a mismatch entry)"
}

# --- group A: profile completeness (validate.sh --profile) ------------------------------------------------------
FAST_SET="$(printf '%s\n' attribution.json builds.tsv catalog-p.log cleanup-inventory.tsv contracts-agree.tsv contracts-results-go.json \
  contracts-results-python.json deps-server.txt harness.log images.txt inventory-leftover.txt junit.xml ledger-bootstrap.tsv ledger-summary.tsv \
  phase-status.tsv plan.out repeat-render.out resources.tsv \
  required-checks.tsv results.tsv runner-preflight.tsv schema-negative-go.json schema-negative-python.json selection.tsv \
  storage.tsv summary.json timings.tsv | LC_ALL=C sort | tr '\n' ' ')"
FAST_SUITES="lint lint-base lint-web lint-core provenance contracts unit"
stub_args() { # SUITES AREAS: the stub harness was given exactly these suites and --areas
  if [[ "$(cat "$C/stub/harness-suites" 2>/dev/null)" != "$1" ]]; then prob "harness suites [$(cat "$C/stub/harness-suites" 2>/dev/null)], want [$1]"; fi
  if [[ "$(cat "$C/stub/harness-areas" 2>/dev/null)" != "$2" ]]; then prob "harness --areas [$(cat "$C/stub/harness-areas" 2>/dev/null)], want [$2]"; fi
}
sum_has() { # KEY VALUE: the published summary.json carries "KEY": VALUE (VALUE as written: "fast", 31, ...)
  if grep -qF "  \"$1\": $2," "$(PUBDIR)/summary.json" 2>/dev/null; then :; else prob "summary.json $1 is not $2"; fi
}
log_has() { if grep -qxF "validate: $1" "$C/out"; then :; else prob "the job log lacks [$1]"; fi; }
ctl_VP1() {
  std_ci; run_v "${CIENV[@]}" -- --profile fast --publish-dir ci-evidence
  expect_rc 0; expect_pub full; expect_runs 1; expect_line has harness=0; expect_phase complete 0
  if [[ "$(published_list)" != "$FAST_SET" ]]; then prob "published set differs from the fast set: [$(published_list)]"; fi
  sum_has profile '"fast"'; sum_has executed_profile '"fast"'; sum_has required_count 32
  sum_has executed_suites "\"$FAST_SUITES\""; sum_has coverage '"partial"'; sum_has partition '"fast"'
  if grep -q "^validate: harness: starting infra/validation/run.sh $FAST_SUITES (profile fast;" "$C/out"; then :; else prob "the harness did not run exactly the fast suites"; fi
  stub_args "$FAST_SUITES" none; bootstrap_local
  detail "exit $RC; fast job of a full pipeline: the 7 fast suites ($FAST_SUITES, --areas none), 32 required IDs, executed_profile fast, coverage partial, partition fast; published the fast set ($(published_list | wc -w) files)"
}
ctl_VP2() { std_ci; run_v "${CIENV[@]}" STUB_DROP_ID=B-UNIT-CHECKS -- --profile fast --publish-dir ci-evidence
  expect_rc 14; expect_phase complete 14; log_has "complete: required check B-UNIT-CHECKS occurs 0 times"; detail "exit $RC; B-UNIT-CHECKS missing from a fast result"; }
ctl_VP3() { std_ci; run_v "${CIENV[@]}" STUB_FAIL_ID=B-UNIT-CHECKS -- --profile fast --publish-dir ci-evidence
  expect_rc 14; expect_line has harness=0; expect_phase complete 14; log_has "complete: required check B-UNIT-CHECKS is not PASS"
  detail "exit $RC; B-UNIT-CHECKS FAIL under harness exit 0"; }
ctl_VP4() { std_ci; run_v "${CIENV[@]}" STUB_RM_REPORT=contracts-results-go.json -- --profile fast --publish-dir ci-evidence
  expect_rc 14; expect_phase complete 14; log_has "complete: contracts-results-go.json is missing"; detail "exit $RC; the Go contract report of a fast result is missing"; }
ctl_VP5() { std_ci; run_v "${CIENV[@]}" STUB_ADD_ROW=B-BROWSER-ROOT -- --profile fast --publish-dir ci-evidence
  expect_rc 14; expect_phase complete 14; log_has "complete: results.tsv has a row for B-BROWSER-ROOT, outside the fast profile"
  detail "exit $RC; a manufactured B-BROWSER-ROOT PASS row in a fast result"; }
ctl_VP6() { new_case; run_v STUB_RM_REPORT=browser-off-results.json -- --profile integration
  expect_rc 14; expect_phase complete 14; log_has "complete: browser-off-results.json is missing"
  if grep -q '^validate: profile pipeline=integration executed=integration reason=local-explicit base=none$' "$C/out"; then :; else prob "not a local integration run"; fi
  detail "exit $RC; local integration without browser-off-results.json"; }
ctl_VP7() { std_ci; run_v "${CIENV[@]}" "STUB_ROWS_SUITES=lint provenance contracts unit" -- --publish-dir ci-evidence
  expect_rc 14; expect_phase complete 14
  if grep -qE '^validate: complete: required check B-[A-Z0-9-]+ occurs 0 times$' "$C/out"; then :; else prob "no missing required check reported"; fi
  detail "exit $RC; a full run given only the fast rows"; }
# G051: local affected takes --base or --compare (accepted); CI never takes integration, affected, --base or --compare.
vp8_refused() { # MESSAGE: refused at location before any harness run
  expect_rc 16; expect_phase location 16; expect_runs 0
  if [[ -n "$1" ]]; then log_has "location: $1"; fi
}
ctl_VP8() {
  local r="" BASE=""
  std_ci; run_v "${CIENV[@]}" -- --profile integration --publish-dir ci-evidence; vp8_refused "--profile in CI accepts fast, gate or staging"; r+=" ci-integration=$RC"
  std_ci 4252; run_v "${CIENV[@]}" -- --profile affected --publish-dir ci-evidence; vp8_refused "--profile in CI accepts fast, gate or staging"; r+=" ci-affected=$RC"
  std_ci 4253; run_v "${CIENV[@]}" -- --profile fast --base "$(head_sha)" --publish-dir ci-evidence; vp8_refused "--base and --compare are local options"; r+=" ci-base=$RC"
  std_ci 4254; run_v "${CIENV[@]}" -- --profile fast --compare "$(head_sha)" --publish-dir ci-evidence; vp8_refused "--base and --compare are local options"; r+=" ci-compare=$RC"
  new_case; run_v -- --profile gate; vp8_refused "--profile locally accepts fast, affected, integration or full"; r+=" local-gate=$RC"
  new_case; run_v -- --profile everything; vp8_refused ""; r+=" local-everything=$RC"
  new_case; run_v -- --profile full --base "$(head_sha)"; vp8_refused "--base and --compare need --profile affected"; r+=" local-full-base=$RC"
  new_case; run_v -- --profile affected --base "$(head_sha)" --compare "$(head_sha)"; vp8_refused "--base and --compare are mutually exclusive"; r+=" local-both=$RC"
  # Accepted: local affected with --base (an ancestor) and with --compare, each over a docs-only change.
  new_case; vp9_docs_commit; run_v -- --profile affected --base "$BASE"
  expect_rc 0; expect_runs 1; stub_args "$FAST_SUITES" none
  log_has "profile pipeline=fast executed=fast reason=docs-only base=$BASE"; r+=" local-affected-base=$RC"
  new_case; vp9_docs_commit; run_v -- --profile affected --compare "$BASE"
  expect_rc 0; expect_runs 1; stub_args "$FAST_SUITES" none; r+=" local-affected-compare=$RC"
  detail "refused at location with no harness, accepted otherwise:$r"
}
vp9_docs_commit() { # a docs-only commit (management/v0/PROGRESS.md) on the case tree; BASE is the commit before it
  BASE="$(head_sha)"
  mkdir -p "$C/repo/management/v0"
  printf 'status\n' >"$C/repo/management/v0/PROGRESS.md"
  git -C "$C/repo" add management/v0/PROGRESS.md
  GIT_AUTHOR_DATE=2026-01-02T00:00:00Z GIT_COMMITTER_DATE=2026-01-02T00:00:00Z \
    git -C "$C/repo" -c user.name=ci-controls -c user.email=controls@example.invalid commit -q -m "docs-only change"
}
ctl_VP9() {
  local r1 BASE=""
  new_case; vp9_docs_commit
  run_v GITLAB_CI=true CI_PROJECT_DIR="$C/repo" CI_JOB_ID=4250 CI_PIPELINE_ID=1001 CI_DEFAULT_BRANCH=main CI_JOB_TIMEOUT=10800 \
    RUNNER_SCRIPT_TIMEOUT=2h45m CI_JOB_STARTED_AT="$(ci_utc)" CI_PIPELINE_SOURCE=merge_request_event CI_MERGE_REQUEST_IID=7 \
    CI_MERGE_REQUEST_EVENT_TYPE=detached CI_MERGE_REQUEST_SOURCE_BRANCH_NAME=feature CI_MERGE_REQUEST_TARGET_BRANCH_NAME=staging \
    CI_COMMIT_SHA="$(head_sha)" CI_COMMIT_REF_NAME=feature CI_MERGE_REQUEST_DIFF_BASE_SHA="$BASE" -- --profile gate --publish-dir ci-evidence
  r1="$RC"
  expect_rc 0; expect_runs 0; expect_line has harness=not-required; expect_phase complete n/a; expect_pub full
  sum_has pipeline_profile '"fast"'; sum_has executed_profile '"none"'; sum_has selection_reason '"docs-only"'; sum_has comparison_base "\"$BASE\""
  if [[ "$(cat "$(PUBDIR)/selection.tsv" 2>/dev/null)" != $'A\tdocs\tmanagement/v0/PROGRESS.md\t-\t-' ]]; then prob "selection.tsv is not the one docs row"; fi
  std_ci 4251; CIENV=("${CIENV[@]/#CI_COMMIT_BRANCH=*/CI_COMMIT_BRANCH=feature}"); CIENV=("${CIENV[@]/#CI_COMMIT_REF_NAME=*/CI_COMMIT_REF_NAME=feature}")
  run_v "${CIENV[@]}" -- --profile gate --publish-dir ci-evidence
  expect_rc 16; expect_runs 0; expect_phase preflight 16; log_has "preflight: profile selection refused (see selection.log)"
  detail "docs-only MR gate exit $r1 (harness=not-required, pipeline fast, executed none, base recorded); gate on a feature push exit $RC"
}

# --- group A: affected and partition completeness (G051 VA1-VA11; stub harness, real validate.sh) ------------------
# A merge request pipeline whose only change is a page and a spec of the web area "demo" (registered in the case
# tree's committed map, the comparison base) is affected: its fast job executes the whole pipeline, its gate none.
# A push to the default branch without a before-SHA is a full pipeline: the fast job runs the fast partition and the
# gate the remainder. Selection faults come from a committed case-tree selector that runs the real one and then
# rewrites one line.
VA_WEB="lint-base lint-web provenance unit-web feature-web"
VA_REM="controls ci-controls root migrations ops rename off"
mr_env() { # JOB_ID BASE: CIENV for a detached merge request pipeline of HEAD against BASE
  CIENV=(GITLAB_CI=true CI_PROJECT_DIR="$C/repo" CI_JOB_ID="$1" CI_PIPELINE_ID=1001 CI_DEFAULT_BRANCH=main CI_JOB_TIMEOUT=10800
    RUNNER_SCRIPT_TIMEOUT=2h45m CI_JOB_STARTED_AT="$(ci_utc)" CI_PIPELINE_SOURCE=merge_request_event CI_MERGE_REQUEST_IID=7
    CI_MERGE_REQUEST_EVENT_TYPE=detached CI_MERGE_REQUEST_SOURCE_BRANCH_NAME=feature CI_MERGE_REQUEST_TARGET_BRANCH_NAME=staging
    CI_COMMIT_SHA="$(head_sha)" CI_COMMIT_REF_NAME=feature CI_MERGE_REQUEST_DIFF_BASE_SHA="$2")
}
case_commit() { # MESSAGE: commit the whole case tree with the fixed identity
  git -C "$C/repo" add -A
  GIT_AUTHOR_DATE=2026-01-02T00:00:00Z GIT_COMMITTER_DATE=2026-01-02T00:00:00Z \
    git -C "$C/repo" -c user.name=ci-controls -c user.email=controls@example.invalid commit -q -m "$1"
}
va_area() { # JOB_ID: the demo-area merge request case; CIENV its pipeline
  local base
  new_case
  printf '%s\n' $'area-web\tdir\tweb/app/(features)/demo/' $'area-web\tdir\tweb/tests/browser/features/demo/' >>"$C/repo/infra/ci/profile-paths.tsv"
  case_commit "register the web area demo"
  base="$(head_sha)"
  mkdir -p "$C/repo/web/app/(features)/demo" "$C/repo/web/tests/browser/features/demo"
  printf 'export default function Demo() { return null; }\n' >"$C/repo/web/app/(features)/demo/page.tsx"
  printf "test('demo', () => {});\n" >"$C/repo/web/tests/browser/features/demo/demo.spec.ts"
  case_commit "the demo page and its spec"
  mr_env "$1" "$base"
}
va_selector() { # KEY VALUE: the case tree's selector (committed) runs the real one, then rewrites the KEY line to KEY=VALUE
  local d="$C/repo/infra/ci"
  mv -- "$d/select-profile.sh" "$d/select-profile.real.sh"
  {
    printf '#!/usr/bin/env bash\n# VA fault stub of validate-controls.sh: the real selector, then the %s line rewritten.\nset -euo pipefail\n' "$1"
    printf '%s\n' 'out="" a=("$@")' 'while (( $# > 0 )); do if [[ "$1" == --out ]]; then out="$2"; shift 2; else shift; fi; done' \
      'bash "$(dirname "${BASH_SOURCE[0]}")/select-profile.real.sh" "${a[@]}"'
    printf 'sed -i "s|^%s=.*|%s=%s|" "$out"\n' "$1" "$1" "$2"
  } >"$d/select-profile.sh"
  case_commit "VA selection fault: $1"
}
log_prefix() { if grep -q "^validate: $1" "$C/out"; then :; else prob "the job log lacks [$1...]"; fi; }
ctl_VA1() {
  local r1
  va_area 4400; run_v "${CIENV[@]}" -- --profile fast --publish-dir ci-evidence
  expect_rc 0; expect_pub full; expect_runs 1; expect_line has harness=0; expect_phase complete 0
  stub_args "$VA_WEB" web:demo
  sum_has pipeline_profile '"affected"'; sum_has executed_profile '"affected"'; sum_has selection_reason '"area"'
  sum_has pipeline_suites "\"$VA_WEB\""; sum_has executed_suites "\"$VA_WEB\""; sum_has coverage '"complete"'
  sum_has partition '"complete"'; sum_has areas '"web:demo"'; sum_has required_count 17
  if [[ "$(jsuite "$(JX)")" != "fast:affected "* ]]; then prob "junit suite [$(jsuite "$(JX)")], want fast:affected"; fi
  for r1 in browser-feature-results.json browser-feature-tests.tsv; do if [[ -f "$(PUBDIR)/$r1" ]]; then :; else prob "$r1 was not published"; fi; done
  if [[ -e "$(PUBDIR)/contracts-results-go.json" ]]; then prob "a contract report was published without the contracts suite"; fi
  r1="$RC"
  va_area 4401; run_v "${CIENV[@]}" -- --profile gate --publish-dir ci-evidence
  expect_rc 0; expect_runs 0; expect_line has harness=not-required; expect_phase complete n/a
  sum_has pipeline_profile '"affected"'; sum_has executed_profile '"none"'; sum_has executed_suites '"none"'; sum_has coverage '"none"'
  sum_has partition '"none"'
  detail "affected web area: fast job exit $r1 (fast:affected, suites $VA_WEB, --areas web:demo, 17 required IDs, coverage complete, browser-feature reports published, no contract report); gate exit $RC (executed none)"
}
ctl_VA2() { va_area 4402; run_v "${CIENV[@]}" STUB_DROP_ID=B-FEATURE-WEB -- --profile fast --publish-dir ci-evidence
  expect_rc 14; expect_phase complete 14; expect_line has harness=0; log_has "complete: required check B-FEATURE-WEB occurs 0 times"
  detail "exit $RC; an affected result without B-FEATURE-WEB"; }
ctl_VA3() { va_area 4403; run_v "${CIENV[@]}" STUB_ADD_ROW=B-UNIT-WEB -- --profile fast --publish-dir ci-evidence
  expect_rc 14; expect_phase complete 14; log_has "complete: required check B-UNIT-WEB occurs 2 times"
  detail "exit $RC; B-UNIT-WEB recorded twice"; }
ctl_VA4() { va_area 4404; run_v "${CIENV[@]}" STUB_FAIL_ID=B-FEATURE-WEB -- --profile fast --publish-dir ci-evidence
  expect_rc 14; expect_line has harness=0; expect_phase complete 14; log_has "complete: required check B-FEATURE-WEB is not PASS"
  detail "exit $RC; B-FEATURE-WEB FAIL under harness exit 0"; }
ctl_VA4b() { va_area 4405; run_v "${CIENV[@]}" STUB_HARNESS=unmet -- --profile fast --publish-dir ci-evidence
  expect_rc 14; expect_line has harness=0; expect_phase complete 14; log_prefix "complete: required check B-[A-Z0-9-]* is not PASS"
  if grep -qP '\tUNMET\t' "$(PUBDIR)/results.tsv" 2>/dev/null; then :; else prob "the published results lack the UNMET row"; fi
  detail "exit $RC; an UNMET row under harness exit 0"; }
ctl_VA5() { va_area 4406; run_v "${CIENV[@]}" STUB_ADD_ROW=B-BROWSER-ROOT -- --profile fast --publish-dir ci-evidence
  expect_rc 14; expect_phase complete 14; log_prefix "complete: results.tsv has a row for B-BROWSER-ROOT, outside the "
  detail "exit $RC; a B-BROWSER-ROOT row outside the executed suites"; }
ctl_VA6() {
  local kv r=""
  for kv in "AREAS=web:Bad!" "EXECUTED_SUITES=Lint" "COVERAGE=total"; do
    new_case; va_selector "${kv%%=*}" "${kv#*=}"; ci_env 4407
    run_v "${CIENV[@]}" -- --profile fast --publish-dir ci-evidence
    expect_rc 16; expect_runs 0; expect_phase preflight 16; log_has "preflight: the profile selection is malformed"
    r+=" ${kv%%=*}=$RC"
  done
  detail "malformed selection state, refused before the harness:$r"
}
ctl_VA7() {
  std_ci 4408; run_v "${CIENV[@]}" -- --profile gate --publish-dir ci-evidence
  expect_rc 0; expect_pub full; expect_runs 1; expect_phase complete 0
  stub_args "$VA_REM" none
  sum_has pipeline_profile '"full"'; sum_has executed_profile '"full"'; sum_has executed_suites "\"$VA_REM\""
  sum_has coverage '"partial"'; sum_has partition '"remainder"'; sum_has required_count 97
  if [[ "$(jsuite "$(JX)")" != "gate:remainder "* ]]; then prob "junit suite [$(jsuite "$(JX)")], want gate:remainder"; fi
  log_has "suites pipeline=[lint lint-base lint-web lint-core controls ci-controls provenance contracts root migrations ops rename off] executed=[$VA_REM] coverage=partial partition=remainder areas=none"
  detail "exit $RC; the gate of a full pipeline runs the remainder ($VA_REM; 97 required IDs), coverage partial, gate:remainder"
}
ctl_VA8() { std_ci 4409; run_v "${CIENV[@]}" STUB_ADD_ROW=B-UNIT-CHECKS -- --profile gate --publish-dir ci-evidence
  expect_rc 14; expect_phase complete 14; log_prefix "complete: results.tsv has a row for B-UNIT-CHECKS, outside the "
  detail "exit $RC; a fast-suite row (B-UNIT-CHECKS) in the gate's remainder"; }
ctl_VA9() {
  local r1 f
  std_ci 4410; run_v "${CIENV[@]}" -- --profile gate --publish-dir ci-evidence
  expect_rc 0; expect_phase complete 0; r1="$RC"
  if [[ -n "$(complete_findings)" ]]; then prob "a completeness finding was reported for the remainder"; fi
  for f in contracts-results-go.json contracts-results-python.json schema-negative-go.json contracts-agree.tsv; do
    if [[ -e "$(PUBDIR)/$f" ]]; then prob "$f was published without the contracts suite"; fi
  done
  std_ci 4411; run_v "${CIENV[@]}" STUB_RM_REPORT=contracts-results-python.json -- --profile fast --publish-dir ci-evidence
  expect_rc 14; expect_phase complete 14; log_has "complete: contracts-results-python.json is missing"
  detail "contract reports are required only when contracts executed: remainder without them exit $r1; fast partition missing one exit $RC"
}
ctl_VA10() { va_area 4412; run_v "${CIENV[@]}" STUB_RM_REPORT=browser-feature-results.json -- --profile fast --publish-dir ci-evidence
  expect_rc 14; expect_phase complete 14; log_has "complete: browser-feature-results.json is missing"
  detail "exit $RC; an executed feature-web without browser-feature-results.json"; }
ctl_VA11() {
  new_case; va_selector EXECUTED_SUITES "controls root migrations ops rename"; ci_env 4413
  run_v "${CIENV[@]}" -- --profile gate --publish-dir ci-evidence
  expect_rc 16; expect_runs 0; expect_phase preflight 16; log_has "preflight: the selection is not one consistent partition"
  detail "exit $RC; a gate remainder without off (remainder != pipeline - fast) is refused before the harness"
}

# --- group A: the pipeline union check (G051 VU1-VU9; infra/ci/pipeline-union.sh of this checkout) -------------------
# The fixtures are real publications: the fast and the gate job of one pipeline (validate.sh with the stub harness on
# one case-tree commit), made once per run under $ROOT/vu/KIND and copied into each control's case. A fault edits
# one copied file (summary.json through the pinned yq image, the TSV files with sed); VU1 shows that a yq
# re-serialization alone still passes.
VU_FINAL="B-DOCKER-ONLY B-REQUIRED-CHECKS B-SCOPE B-SECRET-CANARY"
vu_make() { # KIND (full|integration|rc): the published fast and gate evidence of one pipeline, in $ROOT/vu/KIND
  local d="$ROOT/vu/$1" base="" j r1
  local -A jid=([fast]=4500 [gate]=4501)
  if [[ -e "$d/ok" ]]; then return 0; fi
  rm -rf -- "$d"; mkdir -p "$d"
  new_case
  if [[ "$1" == integration ]]; then
    base="$(head_sha)" jid=([fast]=4502 [gate]=4503)
    mkdir -p "$C/repo/web/app"; printf 'export default function Page() { return null; }\n' >"$C/repo/web/app/page.tsx"
    case_commit "a runtime change"
  fi
  if [[ "$1" == rc ]]; then
    # ONEOFF RC qualification: main stays at the template commit; staging gains a documentation-only commit R1, the
    # annotated RC tag 0.1.0-rc.1 and the runner-prepared origin/staging point at it, and both jobs run the detached
    # promotion merge request release/0.1.0-rc.1 -> main (IID 7, pipeline 1001, project 999999999).
    jid=([fast]=4504 [gate]=4505)
    git -C "$C/repo" checkout -q -b staging
    mkdir -p "$C/repo/docs"; printf 'release note\n' >"$C/repo/docs/release-note.md"
    case_commit "a documentation-only staging change"
    r1="$(head_sha)"
    git -C "$C/repo" checkout -q --detach "$r1"
    GIT_COMMITTER_DATE=2026-01-02T00:00:00Z git -C "$C/repo" -c user.name=ci-controls -c user.email=controls@example.invalid \
      tag -a -m "RC 0.1.0-rc.1" 0.1.0-rc.1 "$r1"
    git -C "$C/repo" update-ref refs/remotes/origin/staging "$r1"
  fi
  for j in fast gate; do
    if [[ "$1" == rc ]]; then rc_vu_env "${jid[$j]}"
    elif [[ -n "$base" ]]; then mr_env "${jid[$j]}" "$base"; else ci_env "${jid[$j]}"; fi
    run_v "${CIENV[@]}" -- --profile "$j" --publish-dir ci-evidence
    if [[ "$RC" != 0 || ! -f "$(PUBDIR)/summary.json" ]]; then prob "the $1 $j job fixture failed (exit $RC)"; return 1; fi
    mv -T -- "$(PUBDIR)" "$d/$j"
  done
  if [[ "$1" == rc ]]; then
    # The release helper's fixtures (group H): this repository and its bare remote (main, staging, the tag and the
    # promotion branch release/0.1.0-rc.1 at the RC commit).
    cp -a -- "$C/repo" "$d/repo"
    git clone -q --bare -- "$C/repo" "$d/remote.git"
    git -C "$d/remote.git" branch release/0.1.0-rc.1 "$r1"
  fi
  : >"$d/ok"
}
rc_vu_env() { # JOB_ID: CIENV for the detached RC-promotion merge request pipeline of HEAD
  CIENV=(GITLAB_CI=true CI_PROJECT_DIR="$C/repo" CI_JOB_ID="$1" CI_PIPELINE_ID=1001 CI_DEFAULT_BRANCH=main CI_JOB_TIMEOUT=10800
    RUNNER_SCRIPT_TIMEOUT=2h45m CI_JOB_STARTED_AT="$(ci_utc)" CI_PIPELINE_SOURCE=merge_request_event CI_MERGE_REQUEST_IID=7
    CI_MERGE_REQUEST_EVENT_TYPE=detached CI_MERGE_REQUEST_SOURCE_BRANCH_NAME=release/0.1.0-rc.1 CI_MERGE_REQUEST_TARGET_BRANCH_NAME=main
    CI_COMMIT_SHA="$(head_sha)" CI_COMMIT_REF_NAME=release/0.1.0-rc.1 CI_PROJECT_ID=999999999 CI_SERVER_VERSION=18.4.0)
}
vu_case() { # KIND: a fresh case with this control's copies of the KIND pair in $C/fast and $C/gate
  new_case
  vu_make "$1" || return 1
  rm -rf -- "$C/fast" "$C/gate"
  cp -a -- "$ROOT/vu/$1/fast" "$C/fast"; cp -a -- "$ROOT/vu/$1/gate" "$C/gate"
}
vu_run() { # ARGS...: pipeline-union.sh of this checkout; RC and VU_LINE (its one output line)
  RC=0
  env -i -C "$C" "${BASE_ENV[@]}" TMPDIR="$C/tmp" SHIM_LABELS="$(case_labels)" bash "$REPO_R/infra/ci/pipeline-union.sh" "$@" >"$C/union.out" 2>&1 || RC=$?
  VU_LINE="$(cat "$C/union.out" 2>/dev/null)"
  if [[ "$(grep -c . "$C/union.out" 2>/dev/null)" != 1 || "$VU_LINE" != "pipeline-union: "* ]]; then prob "the union check did not print exactly one result line: [$(tr '\n' ';' <"$C/union.out" 2>/dev/null)]"; fi
}
vu_fail() { # NAME: the union check of $C/fast and $C/gate fails (exit 1, a FAIL line)
  vu_run "$C/fast" "$C/gate"
  if [[ "$RC" != 1 || "$VU_LINE" != "pipeline-union: FAIL "* ]]; then prob "$1: union exit $RC [$VU_LINE], want 1 and FAIL"; fi
}
vu_json() { # FILE EXPRESSION: rewrite a copied summary.json through the pinned yq image
  local f="$1"
  D run --rm --pull never --network none -i "${LB[@]}" "$CI_YQ_IMAGE" -p json -o json "$2" <"$f" >"$f.new" && mv -f -- "$f.new" "$f"
}
vu_ids() { # DIR: the registry IDs with a PASS row in DIR/results.tsv, sorted
  awk -F'\t' 'NR == FNR {if ($1 != "" && $1 !~ /^#/) r[$1] = 1; next} ($1 in r) && $2 == "PASS" {print $1}' \
    "$1/required-checks.tsv" "$1/results.tsv" | LC_ALL=C sort -u
}
vu_split() { # KIND FAST GATE UNION: the published pair partitions E_pipeline as stated; the check passes, also re-serialized
  local f g u o
  vu_case "$1" || return 0
  vu_run "$C/fast" "$C/gate"
  if [[ "$RC" != 0 || "$VU_LINE" != "pipeline-union: PASS "* ]]; then prob "union exit $RC [$VU_LINE], want 0 and PASS"; fi
  f="$(vu_ids "$C/fast")"; g="$(vu_ids "$C/gate")"
  u="$(printf '%s\n%s\n' "$f" "$g" | LC_ALL=C sort -u | grep -c .)" || u=0
  o="$(LC_ALL=C comm -12 <(printf '%s\n' "$f") <(printf '%s\n' "$g") | tr '\n' ' ')"
  if [[ "$(grep -c . <<<"$f")/$(grep -c . <<<"$g")/$u" != "$2/$3/$4" ]]; then prob "rows fast/gate/union $(grep -c . <<<"$f")/$(grep -c . <<<"$g")/$u, want $2/$3/$4"; fi
  if [[ "$o" != "$VU_FINAL " ]]; then prob "overlap [$o], want exactly the final IDs"; fi
  ctl_labels
  vu_json "$C/fast/summary.json" '.'; vu_json "$C/gate/summary.json" '.'
  vu_run "$C/fast" "$C/gate"
  if [[ "$RC" != 0 ]]; then prob "the re-serialized pair fails (exit $RC [$VU_LINE]); the edit-based faults would be inconclusive"; fi
}
ctl_VU1() { vu_split full 32 97 125; detail "full pipeline: fast 32 + gate 97 -> 125 distinct, overlap exactly the 4 final IDs; pipeline-union exit 0 (also after a yq re-serialization)"; }
ctl_VU9() { vu_split integration 32 44 72; detail "integration pipeline: fast 32 + gate 44 -> 72 distinct, overlap exactly the 4 final IDs; pipeline-union exit 0"; }
ctl_VU2() {
  local r1
  vu_case full || return 0
  rm -rf -- "$C/fast"; mkdir -p "$C/fast"; vu_fail "an empty fast partition"; r1="$RC"
  vu_run "$C/gate"
  if [[ "$RC" != 2 || "$VU_LINE" != "pipeline-union: usage"* ]]; then prob "the gate alone: exit $RC [$VU_LINE], want 2 (usage)"; fi
  detail "an empty fast partition exit $r1 (FAIL); the gate alone exit $RC (usage)"
}
ctl_VU3() {
  local r1
  new_case; vu_make integration || return 0
  vu_case full || return 0; ctl_labels
  vu_json "$C/fast/summary.json" '.executed_suites = "lint lint-base lint-web lint-core provenance contracts"'
  vu_fail "fast executed suites without unit"; r1="$RC"
  rm -rf -- "$C/fast"; cp -a -- "$ROOT/vu/integration/fast" "$C/fast"
  vu_fail "the fast partition of another pipeline"
  detail "a fast partition that is not the fast suites exit $r1; the fast job of another (integration) pipeline with a full gate exit $RC"
}
ctl_VU4() {
  vu_case full || return 0
  sed -i '/^B-BROWSER-ROOT\t/d' "$C/gate/results.tsv"; vu_fail "B-BROWSER-ROOT absent from the remainder"
  detail "exit $RC; a union gap (B-BROWSER-ROOT absent from the gate)"
}
ctl_VU5() {
  vu_case full || return 0
  printf 'B-UNIT-CHECKS\tPASS\tstub overlap\n' >>"$C/gate/results.tsv"; vu_fail "B-UNIT-CHECKS in both partitions"
  detail "exit $RC; an overlap other than the final IDs (B-UNIT-CHECKS in the gate too)"
}
ctl_VU6() {
  vu_case full || return 0; ctl_labels
  vu_json "$C/gate/summary.json" ".tree_sha = \"$(printf 'a%.0s' {1..40})\""; vu_fail "another candidate tree"
  detail "exit $RC; the gate's candidate tree differs"
}
ctl_VU7() {
  local r1
  vu_case full || return 0; ctl_labels
  vu_json "$C/gate/summary.json" ".selection_digest = \"$(printf 'b%.0s' {1..64})\""; vu_fail "another selection digest"; r1="$RC"
  vu_case full || return 0
  vu_json "$C/gate/summary.json" ".pipeline_set_sha256 = \"$(printf 'c%.0s' {1..64})\""; vu_fail "another pipeline set"
  detail "selection_digest mismatch exit $r1; pipeline_set_sha256 mismatch exit $RC"
}
ctl_VU8() {
  vu_case full || return 0
  sed -i 's/^postcheck\t0$/postcheck\t13/' "$C/gate/phase-status.tsv"
  if grep -qx $'postcheck\t13' "$C/gate/phase-status.tsv"; then :; else prob "the fixture has no postcheck phase to alter"; fi
  vu_fail "a gate phase 13"
  detail "exit $RC; a recorded gate phase is not 0 (postcheck 13)"
}
# ONEOFF RC qualification: the published pair of a detached RC-promotion merge request pipeline (vu_make rc).
ctl_VU10() {
  vu_split rc 32 97 125
  if grep -q '"tested_content": "mr-rc-promotion"' "$C/fast/summary.json" "$C/gate/summary.json" 2>/dev/null; then :; else prob "the pair is not mr-rc-promotion"; fi
  if grep -q '"pipeline_suites": "[a-z -]* ci-controls ' "$C/gate/summary.json"; then :; else prob "the RC pipeline suites lack ci-controls"; fi
  # (vu_split re-serialized the copies through yq, so the object spans lines.)
  if grep -q '"tag": "0.1.0-rc.1"' "$C/gate/summary.json" && grep -q '"source_branch": "release/0.1.0-rc.1"' "$C/fast/summary.json"; then :
  else prob "the summaries lack the release identity"; fi
  detail "RC promotion pipeline (docs-only candidate): fast 32 + gate 97 -> 125 distinct incl. ci-controls, release identity in both summaries; pipeline-union exit 0 (also re-serialized)"
}
ctl_VU11() {
  local r1
  vu_case rc || return 0; ctl_labels
  vu_json "$C/gate/summary.json" '.tested_content = "mr-source" | .release = null'; vu_fail "an mr-source gate with an RC fast job"; r1="$RC"
  if [[ "$VU_LINE" != *"tested_content differs"* ]]; then prob "mixed tested content: [$VU_LINE], want tested_content differs"; fi
  vu_case rc || return 0
  vu_json "$C/fast/summary.json" '.release = null'; vu_fail "an RC fast job without its release identity"
  detail "mixed tested content (RC fast, mr-source gate) exit $r1; an RC partition without its release identity exit $RC"
}
ctl_VU12() {
  local r1 r2 x40
  x40="$(printf 'a%.0s' {1..40})"
  vu_case rc || return 0; ctl_labels
  vu_json "$C/gate/summary.json" ".release.staging_sha = \"$x40\""; vu_run "$C/fast" "$C/gate"; r1="$RC"
  if [[ "$RC" != 0 ]]; then prob "independent staging observations must not be compared: [$VU_LINE]"; fi
  vu_case rc || return 0
  vu_json "$C/gate/summary.json" ".release.tag_object = \"$x40\""; vu_fail "another tag object"; r2="$RC"
  vu_case rc || return 0
  vu_json "$C/gate/summary.json" '.ci.project_id = 1'; vu_fail "another project"
  detail "a different observed staging tip per job exit $r1 (independent); another tag object exit $r2; another project exit $RC"
}

# --- group B: attribution ----------------------------------------------------------------------------
ORIGIN="$ROOT/origin" T_SHA="" S_SHA="" M_SHA=""
mk_origin() {
  cp -a -- "$TEMPLATE" "$ORIGIN"
  T_SHA="$(git -C "$ORIGIN" rev-parse HEAD)"
  git -C "$ORIGIN" branch staging
  git -C "$ORIGIN" checkout -q -b feature
  printf 'feature\n' >"$ORIGIN/feature.txt"
  git -C "$ORIGIN" add feature.txt
  git -C "$ORIGIN" -c user.name=ci-controls -c user.email=controls@example.invalid commit -q -m "source change"
  S_SHA="$(git -C "$ORIGIN" rev-parse HEAD)"
  git -C "$ORIGIN" checkout -q --detach "$T_SHA"
  git -C "$ORIGIN" -c user.name=ci-controls -c user.email=controls@example.invalid merge -q --no-ff -m "merged result" "$S_SHA"
  M_SHA="$(git -C "$ORIGIN" rev-parse HEAD)"
  git -C "$ORIGIN" update-ref refs/merge-requests/1/merge "$M_SHA"
  git -C "$ORIGIN" checkout -q main
}
b_checkout() { # REF: a depth-1 checkout of REF from the fixture origin into $C/repo
  C="$ROOT/cases/$CTL_ID"
  rm -rf -- "$C"
  mkdir -p "$C/tmp" "$C/stub" "$C/repo"
  : >"$C/stub/harness-invocations"
  git -C "$C/repo" init -q
  git -C "$C/repo" fetch -q --depth 1 "file://$ORIGIN" "$1"
  git -C "$C/repo" checkout -q --detach FETCH_HEAD
}
GL=() MR=()
b_env() { GL=(GITLAB_CI=true CI_DEFAULT_BRANCH=main CI_PROJECT_DIR="$C/repo" CI_JOB_ID=5100 CI_PIPELINE_ID=2001 CI_JOB_TIMEOUT=10800 RUNNER_SCRIPT_TIMEOUT=2h45m
    CI_JOB_STARTED_AT="$(ci_utc)")
  MR=(CI_PIPELINE_SOURCE=merge_request_event CI_MERGE_REQUEST_IID=7 CI_MERGE_REQUEST_SOURCE_BRANCH_NAME=feature CI_MERGE_REQUEST_TARGET_BRANCH_NAME=staging); }
b_positive() { # TESTED_CONTENT [CI]: exit as RC already checked by the caller
  if [[ "${2:-}" == local ]]; then
    if grep -q "^validate: tested_content=$1 " "$C/out"; then :; else prob "job log lacks tested_content=$1"; fi
    return 0
  fi
  if grep -q "\"tested_content\": \"$1\"" "$(PUBDIR)/summary.json" 2>/dev/null; then :; else prob "summary.json tested_content is not $1"; fi
  if grep -q "\"tested_content\": \"$1\"" "$(PUBDIR)/attribution.json" 2>/dev/null; then :; else prob "attribution.json tested_content is not $1"; fi
}
b_negative() { # REASON [local]: an attribution refusal for exactly REASON (validate.sh's fixed text, never a
  # variable's value): 16, no harness, F1 and F2, only preflight evidence and the summary published
  local f n reason="$1"
  expect_rc 16; expect_runs 0; expect_f1f2; expect_phase preflight 16; expect_line not internal=
  n="$(grep -c '^validate: attribution rejected: ' "$C/out" || true)"
  if [[ "$n" != 1 ]]; then prob "the job log has $n attribution rejections, want 1"
  elif grep -qxF "validate: attribution rejected: $reason" "$C/out"; then :
  else prob "rejection reason [$(sed -n 's/^validate: attribution rejected: //p' "$C/out")], want [$reason]"; fi
  if [[ "${2:-}" == local ]]; then expect_pub none; return 0; fi
  expect_pub full
  for f in $(published_list); do
    case "$f" in runner-preflight.tsv|attribution.json|cleanup-inventory.tsv|storage.tsv|summary.json|phase-status.tsv|junit.xml) ;; *) prob "unexpected published file $f" ;; esac
  done
  if grep -q '"result": "rejected"' "$(PUBDIR)/attribution.json" 2>/dev/null; then :; else prob "attribution.json does not record the rejection"; fi
  if grep -qxF "  \"reason\": \"$reason\"," "$(PUBDIR)/attribution.json" 2>/dev/null; then :; else prob "attribution.json does not record the reason [$reason]"; fi
}
ctl_P1() { b_checkout refs/heads/staging; b_env; run_v "${GL[@]}" CI_PIPELINE_SOURCE=push CI_COMMIT_SHA="$T_SHA" CI_COMMIT_BRANCH=staging CI_COMMIT_REF_NAME=staging -- --publish-dir ci-evidence
  expect_rc 0; expect_pub full; b_positive branch-push; detail "exit $RC; branch-push (staging)"; }
ctl_P2() { b_checkout refs/heads/main; b_env; run_v "${GL[@]}" CI_PIPELINE_SOURCE=push CI_COMMIT_SHA="$T_SHA" CI_COMMIT_BRANCH=main CI_COMMIT_REF_NAME=main -- --publish-dir ci-evidence
  expect_rc 0; expect_pub full; b_positive branch-push; detail "exit $RC; branch-push (main)"; }
ctl_P3() { b_checkout refs/heads/feature; b_env; run_v "${GL[@]}" "${MR[@]}" CI_MERGE_REQUEST_EVENT_TYPE=detached CI_COMMIT_SHA="$S_SHA" CI_COMMIT_REF_NAME=feature -- --publish-dir ci-evidence
  expect_rc 0; expect_pub full; b_positive mr-source
  if grep -q '"CI_MERGE_REQUEST_IID": "7"' "$(PUBDIR)/attribution.json" 2>/dev/null; then :; else prob "IID not recorded"; fi
  detail "exit $RC; mr-source, IID recorded"; }
ctl_P3F() { b_checkout refs/heads/feature; b_env; run_v "${GL[@]}" "${MR[@]}" CI_MERGE_REQUEST_EVENT_TYPE=detached CI_COMMIT_SHA="$S_SHA" CI_COMMIT_REF_NAME=feature STUB_HARNESS=fail -- --publish-dir ci-evidence
  expect_rc 1; expect_pub full; b_positive mr-source; detail "exit $RC; still mr-source"; }
ctl_P4() { b_checkout refs/merge-requests/1/merge; b_env; run_v "${GL[@]}" "${MR[@]}" CI_MERGE_REQUEST_EVENT_TYPE=merged_result CI_COMMIT_SHA="$M_SHA" CI_MERGE_REQUEST_SOURCE_BRANCH_SHA="$S_SHA" CI_MERGE_REQUEST_TARGET_BRANCH_SHA="$T_SHA" -- --publish-dir ci-evidence
  expect_rc 0; expect_pub full; b_positive mr-merged-result
  if grep -q "\"CI_MERGE_REQUEST_SOURCE_BRANCH_SHA\": \"$S_SHA\"" "$(PUBDIR)/attribution.json" 2>/dev/null; then :; else prob "source SHA not recorded"; fi
  detail "exit $RC; mr-merged-result with both SHAs (depth-1 merge commit)"; }
ctl_P5() { b_checkout refs/heads/feature; run_v -- ; expect_rc 0; expect_pub none; b_positive local local
  if grep -q '^validate: profile pipeline=full executed=full reason=local-explicit base=none$' "$C/out"; then :; else prob "the local default is not the full profile"; fi
  detail "exit $RC; local (default profile full)"; }
ctl_P6() { b_checkout refs/heads/feature; b_env; run_v "${GL[@]}" CI_PIPELINE_SOURCE=push CI_COMMIT_SHA="$S_SHA" CI_COMMIT_BRANCH=feature CI_COMMIT_REF_NAME=feature -- --profile fast --publish-dir ci-evidence
  expect_rc 0; expect_pub full; b_positive branch-push
  sum_has pipeline_profile '"fast"'; sum_has executed_profile '"fast"'; sum_has selection_reason '"feature-push"'
  detail "exit $RC; branch-push on a feature branch: fast (feature-push)"; }
ctl_P7() { b_checkout refs/heads/main; b_env; run_v "${GL[@]}" CI_PIPELINE_SOURCE=schedule CI_COMMIT_SHA="$T_SHA" CI_COMMIT_BRANCH=main CI_COMMIT_REF_NAME=main -- --profile gate --publish-dir ci-evidence
  expect_rc 0; expect_pub full; b_positive branch-schedule
  sum_has pipeline_profile '"full"'; sum_has executed_profile '"full"'; sum_has selection_reason '"schedule"'
  detail "exit $RC; schedule on main: branch-schedule, gate full (schedule)"; }
ctl_P8() { b_checkout refs/heads/feature; b_env; run_v "${GL[@]}" CI_PIPELINE_SOURCE=web CI_COMMIT_SHA="$S_SHA" CI_COMMIT_BRANCH=feature CI_COMMIT_REF_NAME=feature -- --profile gate --publish-dir ci-evidence
  expect_rc 0; expect_pub full; b_positive branch-web
  sum_has pipeline_profile '"full"'; sum_has executed_profile '"full"'; sum_has selection_reason '"web-explicit"'
  detail "exit $RC; web pipeline: branch-web, gate full (web-explicit)"; }
# G048 supplement: branch pipelines started through the pipelines API run fast plus the full gate.
ctl_P9() { b_checkout refs/heads/feature; b_env; run_v "${GL[@]}" CI_PIPELINE_SOURCE=api CI_COMMIT_SHA="$S_SHA" CI_COMMIT_BRANCH=feature CI_COMMIT_REF_NAME=feature -- --profile gate --publish-dir ci-evidence
  expect_rc 0; expect_pub full; b_positive branch-api
  sum_has pipeline_profile '"full"'; sum_has executed_profile '"full"'; sum_has selection_reason '"api-explicit"'
  if grep -qF '"CI_PIPELINE_SOURCE": "api"' "$(PUBDIR)/attribution.json" 2>/dev/null; then :; else prob "attribution.json does not record the api source"; fi
  detail "exit $RC; api pipeline: branch-api, gate full (api-explicit), api recorded"; }
ctl_P10() { b_checkout refs/heads/feature; b_env; run_v "${GL[@]}" CI_PIPELINE_SOURCE=api CI_COMMIT_SHA="$S_SHA" CI_COMMIT_BRANCH=feature CI_COMMIT_REF_NAME=feature -- --profile fast --publish-dir ci-evidence
  expect_rc 0; expect_pub full; b_positive branch-api
  sum_has pipeline_profile '"full"'; sum_has executed_profile '"fast"'; sum_has selection_reason '"api-explicit"'
  detail "exit $RC; api pipeline fast job: branch-api, pipeline full, executed fast (api-explicit)"; }
ctl_N1() { b_checkout refs/heads/feature; b_env; run_v "${GL[@]}" CI_COMMIT_SHA="$S_SHA" -- --publish-dir ci-evidence; b_negative "unsupported pipeline source"
  if grep -q '"CI_PIPELINE_SOURCE"' "$(PUBDIR)/attribution.json" 2>/dev/null; then prob "attribution.json records a pipeline source"; fi
  detail "exit $RC: no pipeline source (unsupported pipeline source; none recorded)"; }
ctl_N2() { b_checkout refs/heads/feature; b_env; run_v "${GL[@]}" "${MR[@]}" CI_MERGE_REQUEST_EVENT_TYPE=detached CI_COMMIT_SHA="$T_SHA" -- --publish-dir ci-evidence; b_negative "checked-out HEAD differs from CI_COMMIT_SHA"; detail "exit $RC: CI_COMMIT_SHA differs from HEAD (checked-out HEAD differs from CI_COMMIT_SHA)"; }
ctl_N3() { b_checkout refs/heads/feature; b_env; run_v "${GL[@]}" "${MR[@]}" CI_MERGE_REQUEST_EVENT_TYPE=detached -- --publish-dir ci-evidence; b_negative "CI_COMMIT_SHA missing or malformed"; detail "exit $RC: CI_COMMIT_SHA missing (CI_COMMIT_SHA missing or malformed)"; }
ctl_N4() { b_checkout refs/heads/feature; b_env; run_v "${GL[@]}" "${MR[@]}" CI_MERGE_REQUEST_EVENT_TYPE=detached CI_COMMIT_SHA="$S_SHA" CI_MERGE_REQUEST_SOURCE_BRANCH_SHA="$S_SHA" -- --publish-dir ci-evidence; b_negative "detached pipeline carries merged-result SHAs"; detail "exit $RC: detached with a source SHA (detached pipeline carries merged-result SHAs)"; }
ctl_N5() { b_checkout refs/merge-requests/1/merge; b_env; run_v "${GL[@]}" "${MR[@]}" CI_MERGE_REQUEST_EVENT_TYPE=merged_result CI_COMMIT_SHA="$M_SHA" CI_MERGE_REQUEST_TARGET_BRANCH_SHA="$T_SHA" -- --publish-dir ci-evidence; b_negative "merged-result pipeline without source/target SHAs"; detail "exit $RC: merged result without the source SHA (merged-result pipeline without source/target SHAs)"; }
ctl_N6() { b_checkout refs/heads/feature; b_env; run_v "${GL[@]}" "${MR[@]}" CI_MERGE_REQUEST_EVENT_TYPE=merged_result CI_COMMIT_SHA="$S_SHA" CI_MERGE_REQUEST_SOURCE_BRANCH_SHA="$S_SHA" CI_MERGE_REQUEST_TARGET_BRANCH_SHA="$T_SHA" -- --publish-dir ci-evidence; b_negative "merged-result HEAD equals a branch head"; detail "exit $RC: merged result with HEAD = source (merged-result HEAD equals a branch head)"; }
ctl_N7() { b_checkout refs/merge-requests/1/merge; b_env; run_v "${GL[@]}" "${MR[@]}" CI_MERGE_REQUEST_EVENT_TYPE=merged_result CI_COMMIT_SHA="$M_SHA" CI_MERGE_REQUEST_SOURCE_BRANCH_SHA="$S_SHA" CI_MERGE_REQUEST_TARGET_BRANCH_SHA="$(printf 'a%.0s' {1..40})" -- --publish-dir ci-evidence; b_negative "merged-result parents are not exactly the source and target SHAs"; detail "exit $RC: merged-result parents differ (merged-result parents are not exactly the source and target SHAs)"; }
ctl_N8() { b_checkout refs/heads/staging; b_env; run_v "${GL[@]}" CI_PIPELINE_SOURCE=push CI_COMMIT_SHA="$T_SHA" CI_COMMIT_BRANCH=staging CI_COMMIT_REF_NAME=staging CI_MERGE_REQUEST_IID=7 -- --publish-dir ci-evidence; b_negative "push pipeline carries CI_MERGE_REQUEST_IID"; detail "exit $RC: push with an MR IID (push pipeline carries CI_MERGE_REQUEST_IID)"; }
ctl_N9() { b_checkout refs/heads/feature; b_env; run_v "${GL[@]}" CI_PIPELINE_SOURCE=push CI_COMMIT_SHA="$S_SHA" CI_COMMIT_BRANCH=feature CI_COMMIT_REF_NAME=other -- --publish-dir ci-evidence; b_negative "CI_COMMIT_REF_NAME differs from CI_COMMIT_BRANCH"; detail "exit $RC: push whose ref name differs from its branch (CI_COMMIT_REF_NAME differs from CI_COMMIT_BRANCH)"; }
ctl_N10() { b_checkout refs/heads/feature; b_env; run_v "${GL[@]}" "${MR[@]}" CI_MERGE_REQUEST_EVENT_TYPE=detached CI_COMMIT_SHA="$S_SHA" CI_COMMIT_BRANCH=feature -- --publish-dir ci-evidence; b_negative "CI_COMMIT_BRANCH set in a merge request pipeline"; detail "exit $RC: MR pipeline with CI_COMMIT_BRANCH (CI_COMMIT_BRANCH set in a merge request pipeline)"; }
ctl_N11() { b_checkout refs/merge-requests/1/merge; b_env; run_v "${GL[@]}" "${MR[@]}" CI_MERGE_REQUEST_EVENT_TYPE=merge_train CI_COMMIT_SHA="$M_SHA" CI_MERGE_REQUEST_SOURCE_BRANCH_SHA="$S_SHA" CI_MERGE_REQUEST_TARGET_BRANCH_SHA="$T_SHA" -- --publish-dir ci-evidence; b_negative "merge_train pipelines are not a supported profile"; detail "exit $RC: merge_train (merge_train pipelines are not a supported profile)"; }
ctl_N12() { b_checkout refs/heads/feature; b_env; run_v "${GL[@]}" CI_PIPELINE_SOURCE=schedule CI_COMMIT_SHA="$S_SHA" CI_COMMIT_BRANCH=feature CI_COMMIT_REF_NAME=feature -- --publish-dir ci-evidence; b_negative "schedule outside the default branch"
  if grep -qF '"CI_PIPELINE_SOURCE": "schedule"' "$(PUBDIR)/attribution.json" 2>/dev/null; then :; else prob "attribution.json does not record the schedule source"; fi
  detail "exit $RC: schedule on a feature branch (schedule outside the default branch; schedule recorded)"; }
# G004 E-1: the identity-qualification schedule is accepted only when the ref is protected.
ctl_N12q() { b_checkout refs/heads/feature; b_env; run_v "${GL[@]}" CI_PIPELINE_SOURCE=schedule CI_COMMIT_SHA="$S_SHA" CI_COMMIT_BRANCH=identity-qualification CI_COMMIT_REF_NAME=identity-qualification CI_COMMIT_REF_PROTECTED=false -- --publish-dir ci-evidence
  b_negative "schedule outside the default branch"
  detail "exit $RC: an unprotected identity-qualification schedule (schedule outside the default branch)"; }
ctl_N13() { b_checkout refs/heads/feature; run_v CI_COMMIT_SHA="$S_SHA" -- ; b_negative "local run carries CI_COMMIT_SHA" local; detail "exit $RC: local run carrying CI_COMMIT_SHA (local run carries CI_COMMIT_SHA)"; }
ctl_N14() { b_checkout refs/heads/main; b_env; run_v "${GL[@]}" CI_PIPELINE_SOURCE=push CI_COMMIT_SHA="$T_SHA" CI_COMMIT_TAG=v1 -- --publish-dir ci-evidence; b_negative "tag pipelines are not a supported profile"; detail "exit $RC: tag pipeline (tag pipelines are not a supported profile)"; }
ctl_N15() { b_checkout refs/heads/feature; b_env; run_v "${GL[@]}" "${MR[@]}" CI_COMMIT_SHA="$S_SHA" -- --publish-dir ci-evidence; b_negative "merge request event type missing or unknown"; detail "exit $RC: MR without an event type (merge request event type missing or unknown)"; }
ctl_N16() { b_checkout refs/heads/feature; b_env; run_v "${GL[@]}" CI_PIPELINE_SOURCE=merge_request_event CI_MERGE_REQUEST_EVENT_TYPE=detached CI_COMMIT_SHA="$S_SHA" CI_MERGE_REQUEST_SOURCE_BRANCH_NAME=feature CI_MERGE_REQUEST_TARGET_BRANCH_NAME=staging -- --publish-dir ci-evidence; b_negative "merge request IID missing"; detail "exit $RC: MR without an IID (merge request IID missing)"; }
# A genuinely unsupported source (G048 supplement: api is now supported, so trigger takes its place).
ctl_N17() { b_checkout refs/heads/feature; b_env; run_v "${GL[@]}" CI_PIPELINE_SOURCE=trigger CI_COMMIT_SHA="$S_SHA" CI_COMMIT_BRANCH=feature CI_COMMIT_REF_NAME=feature -- --publish-dir ci-evidence; b_negative "unsupported pipeline source"
  if grep -qF '"CI_PIPELINE_SOURCE": "trigger"' "$(PUBDIR)/attribution.json" 2>/dev/null; then :; else prob "attribution.json does not record the trigger source"; fi
  detail "exit $RC: trigger pipeline (unsupported pipeline source; trigger recorded)"; }
ctl_N18() { b_checkout refs/heads/feature; b_env; run_v "${GL[@]}" CI_PIPELINE_SOURCE=web CI_COMMIT_SHA="$S_SHA" CI_COMMIT_BRANCH=feature CI_COMMIT_REF_NAME=feature CI_MERGE_REQUEST_IID=7 -- --publish-dir ci-evidence; b_negative "web pipeline carries CI_MERGE_REQUEST_IID"; detail "exit $RC: web pipeline with an MR IID (web pipeline carries CI_MERGE_REQUEST_IID)"; }
# API branch pipelines keep every attribution refusal (G048 supplement).
ctl_N19() { b_checkout refs/heads/feature; b_env; run_v "${GL[@]}" CI_PIPELINE_SOURCE=api CI_COMMIT_SHA="$T_SHA" CI_COMMIT_BRANCH=feature CI_COMMIT_REF_NAME=feature -- --publish-dir ci-evidence; b_negative "checked-out HEAD differs from CI_COMMIT_SHA"; detail "exit $RC: api pipeline whose commit differs from HEAD (checked-out HEAD differs from CI_COMMIT_SHA)"; }
ctl_N20() { b_checkout refs/heads/feature; b_env; run_v "${GL[@]}" CI_PIPELINE_SOURCE=api CI_COMMIT_SHA="$S_SHA" CI_COMMIT_REF_NAME=feature -- --publish-dir ci-evidence; b_negative "api pipeline without CI_COMMIT_BRANCH"; detail "exit $RC: api pipeline without a branch (api pipeline without CI_COMMIT_BRANCH)"; }
ctl_N21() { b_checkout refs/heads/feature; b_env; run_v "${GL[@]}" CI_PIPELINE_SOURCE=api CI_COMMIT_SHA="$S_SHA" CI_COMMIT_BRANCH=feature CI_COMMIT_REF_NAME=other -- --publish-dir ci-evidence; b_negative "CI_COMMIT_REF_NAME differs from CI_COMMIT_BRANCH"; detail "exit $RC: api pipeline whose ref name differs from its branch (CI_COMMIT_REF_NAME differs from CI_COMMIT_BRANCH)"; }
ctl_N22() { b_checkout refs/heads/feature; b_env; run_v "${GL[@]}" CI_PIPELINE_SOURCE=api CI_COMMIT_SHA="$S_SHA" CI_COMMIT_TAG=v1 -- --publish-dir ci-evidence; b_negative "tag pipelines are not a supported profile"; detail "exit $RC: api tag pipeline (tag pipelines are not a supported profile)"; }
ctl_N23() { b_checkout refs/heads/feature; b_env; run_v "${GL[@]}" CI_PIPELINE_SOURCE=api CI_COMMIT_SHA="$S_SHA" CI_COMMIT_BRANCH=feature CI_COMMIT_REF_NAME=feature CI_MERGE_REQUEST_IID=7 -- --publish-dir ci-evidence; b_negative "api pipeline carries CI_MERGE_REQUEST_IID"; detail "exit $RC: api pipeline with an MR IID (api pipeline carries CI_MERGE_REQUEST_IID)"; }

# --- group B: RC-promotion attribution (ONEOFF RC qualification; P11, P11F, N40-N49, RR1, RR2) ----------------
# A copy of the fixture origin with the promotion refs: annotated 0.1.0-rc.1 at T (main = staging), lightweight
# 0.1.0-rc.2 at T, annotated 0.1.0-rc.3 at T with release/0.1.0-rc.3 one commit ahead (X), annotated 0.1.0-rc.4 at the
# feature commit S (not in staging), release/0.1.0-rc.5 without a tag, and GitLab's refs/merge-requests/7/head at T.
# rc_checkout reproduces the runner's fetch shape: init, an origin remote, the merge request head refspec plus the
# GIT_FETCH_EXTRA_FLAGS string word-split as the runner does (default: the value in .gitlab-ci.yml).
RC_ORIGIN="$ROOT/origin-rc" X_SHA="" RC_FLAGS="" PR=()
mk_rc_origin() {
  local o="$RC_ORIGIN" b
  local -a id=(-c user.name=ci-controls -c user.email=controls@example.invalid)
  cp -a -- "$ORIGIN" "$o"
  git -C "$o" "${id[@]}" tag -a -m "RC 0.1.0-rc.1" 0.1.0-rc.1 "$T_SHA"
  git -C "$o" tag 0.1.0-rc.2 "$T_SHA"
  git -C "$o" "${id[@]}" tag -a -m "RC 0.1.0-rc.3" 0.1.0-rc.3 "$T_SHA"
  git -C "$o" "${id[@]}" tag -a -m "RC 0.1.0-rc.4" 0.1.0-rc.4 "$S_SHA"
  git -C "$o" checkout -q --detach "$T_SHA"
  printf 'pushed after the tag\n' >"$o/extra.txt"
  git -C "$o" add extra.txt
  git -C "$o" "${id[@]}" commit -q -m "a commit pushed to the promotion branch"
  X_SHA="$(git -C "$o" rev-parse HEAD)"
  git -C "$o" checkout -q main
  for b in 0.1.0-rc.1 0.1.0-rc.2 0.1.0-rc.5; do git -C "$o" branch "release/$b" "$T_SHA"; done
  git -C "$o" branch release/0.1.0-rc.3 "$X_SHA"
  git -C "$o" branch release/0.1.0-rc.4 "$S_SHA"
  git -C "$o" update-ref refs/merge-requests/7/head "$T_SHA"
  RC_FLAGS="$(sed -n 's/^ *GIT_FETCH_EXTRA_FLAGS: "\(.*\)"$/\1/p' "$REPO_R/.gitlab-ci.yml" | LC_ALL=C sort -u)"
}
rc_checkout() { # REF [FLAGS]: the runner's fetch of REF from the RC origin into $C/repo, detached at it
  local flags="${2-$RC_FLAGS}"
  local -a ex=()
  C="$ROOT/cases/$CTL_ID"
  rm -rf -- "$C"
  mkdir -p "$C/tmp" "$C/stub" "$C/repo"
  : >"$C/stub/harness-invocations"
  git -C "$C/repo" init -q
  git -C "$C/repo" remote add origin "file://$RC_ORIGIN"
  read -r -a ex <<<"$flags"
  git -C "$C/repo" fetch -q origin "+$1:refs/remotes/origin/mr-head" "${ex[@]}"
  git -C "$C/repo" checkout -q --detach refs/remotes/origin/mr-head
}
rc_env() { # RC: GL plus PR, the detached promotion merge request release/RC -> main (IID 7, project 999999999)
  b_env
  PR=(CI_PIPELINE_SOURCE=merge_request_event CI_MERGE_REQUEST_IID=7 CI_MERGE_REQUEST_EVENT_TYPE=detached
    CI_MERGE_REQUEST_SOURCE_BRANCH_NAME="release/$1" CI_MERGE_REQUEST_TARGET_BRANCH_NAME=main CI_COMMIT_REF_NAME="release/$1"
    CI_PROJECT_ID=999999999)
}
rc_positive() { # PROFILE EXECUTED: an accepted RC promotion of T with its release identity in the evidence
  local obj x
  obj="$(git -C "$RC_ORIGIN" rev-parse refs/tags/0.1.0-rc.1)"
  expect_rc 0; expect_pub full; b_positive mr-rc-promotion
  sum_has pipeline_profile '"full"'; sum_has executed_profile "\"$2\""; sum_has selection_reason '"rc-promotion"'
  sum_has release "{\"tag\": \"0.1.0-rc.1\", \"tag_object\": \"$obj\", \"peeled\": \"$T_SHA\", \"staging_sha\": \"$T_SHA\", \"merge_request_iid\": 7, \"source_branch\": \"release/0.1.0-rc.1\", \"target_branch\": \"main\"}"
  if grep -q '"pipeline_suites": "[a-z -]* ci-controls ' "$(PUBDIR)/summary.json" 2>/dev/null; then :; else prob "the RC pipeline suites lack ci-controls"; fi
  if grep -qF '"project_id": 999999999' "$(PUBDIR)/summary.json" 2>/dev/null; then :; else prob "summary.json lacks ci.project_id"; fi
  if grep -qF "\"release\": {\"tag\": \"0.1.0-rc.1\", \"tag_object\": \"$obj\"" "$(PUBDIR)/attribution.json" 2>/dev/null; then :; else prob "attribution.json lacks the release identity"; fi
  for x in CI_COMMIT_BRANCH CI_COMMIT_TAG CI_MERGE_REQUEST_SOURCE_BRANCH_SHA CI_MERGE_REQUEST_TARGET_BRANCH_SHA; do
    if grep -qF "\"$x\"" "$(PUBDIR)/attribution.json" 2>/dev/null; then prob "attribution.json records $x (detached pipelines leave it empty)"; fi
  done
}
ctl_P11() { rc_checkout refs/merge-requests/7/head; rc_env 0.1.0-rc.1; run_v "${GL[@]}" "${PR[@]}" CI_COMMIT_SHA="$T_SHA" -- --profile gate --publish-dir ci-evidence
  rc_positive gate full; detail "exit $RC; detached RC promotion: mr-rc-promotion, gate full remainder incl. ci-controls (rc-promotion), release identity and project recorded, branch/tag/merged-result variables empty"; }
ctl_P11F() { rc_checkout refs/merge-requests/7/head; rc_env 0.1.0-rc.1; run_v "${GL[@]}" "${PR[@]}" CI_COMMIT_SHA="$T_SHA" -- --profile fast --publish-dir ci-evidence
  rc_positive fast fast; detail "exit $RC; detached RC promotion fast job: mr-rc-promotion, pipeline full, executed fast, release identity recorded"; }
ctl_N40() { rc_checkout refs/merge-requests/7/head; rc_env 0.1.0-rc.1; run_v "${GL[@]}" "${PR[@]}" CI_COMMIT_SHA="$T_SHA" CI_MERGE_REQUEST_TARGET_BRANCH_NAME=staging -- --publish-dir ci-evidence
  b_negative "RC promotion target is not the default branch"; detail "exit $RC: release/0.1.0-rc.1 -> staging (RC promotion target is not the default branch)"; }
ctl_N41() {
  local n out=""
  for n in 01.0.0-rc.1 0.1.0-rc.0 v0.1.0-rc.1 0.1.0-RC.1 0.1.0-rc.1+b1; do
    rc_checkout refs/merge-requests/7/head; rc_env "$n"; run_v "${GL[@]}" "${PR[@]}" CI_COMMIT_SHA="$T_SHA" -- --publish-dir ci-evidence
    b_negative "RC promotion source branch is not release/ and a canonical RC name"; out+=" release/$n:$RC"
  done
  detail "malformed RC names in the source branch refused:$out (RC promotion source branch is not release/ and a canonical RC name)"
}
ctl_N42() { rc_checkout refs/heads/release/0.1.0-rc.5; rc_env 0.1.0-rc.5; run_v "${GL[@]}" "${PR[@]}" CI_COMMIT_SHA="$T_SHA" -- --publish-dir ci-evidence
  b_negative "RC refs not prepared: refs/tags of the RC"; detail "exit $RC: release/0.1.0-rc.5 has no tag (RC refs not prepared: refs/tags of the RC)"; }
ctl_N43() { rc_checkout refs/heads/release/0.1.0-rc.2; rc_env 0.1.0-rc.2; run_v "${GL[@]}" "${PR[@]}" CI_COMMIT_SHA="$T_SHA" -- --publish-dir ci-evidence
  b_negative "RC tag is not annotated"; detail "exit $RC: lightweight 0.1.0-rc.2 (RC tag is not annotated)"; }
ctl_N44() { rc_checkout refs/heads/release/0.1.0-rc.3; rc_env 0.1.0-rc.3; run_v "${GL[@]}" "${PR[@]}" CI_COMMIT_SHA="$X_SHA" -- --publish-dir ci-evidence
  b_negative "RC tag peel differs from the merge request head"; detail "exit $RC: a commit pushed to release/0.1.0-rc.3 after the tag (RC tag peel differs from the merge request head)"; }
# The r2 delta's "SOURCE_BRANCH_SHA differs" control is replaced (review condition 1): detached pipelines leave the
# merged-result SHAs, CI_COMMIT_BRANCH and CI_COMMIT_TAG empty, so any value there is contamination and refused.
ctl_N45() {
  local kv out=""
  for kv in CI_MERGE_REQUEST_SOURCE_BRANCH_SHA="$T_SHA" CI_MERGE_REQUEST_TARGET_BRANCH_SHA="$T_SHA" CI_COMMIT_BRANCH=release/0.1.0-rc.1 CI_COMMIT_TAG=0.1.0-rc.1; do
    rc_checkout refs/merge-requests/7/head; rc_env 0.1.0-rc.1; run_v "${GL[@]}" "${PR[@]}" CI_COMMIT_SHA="$T_SHA" "$kv" -- --publish-dir ci-evidence
    case "${kv%%=*}" in
      CI_COMMIT_BRANCH) b_negative "CI_COMMIT_BRANCH set in a merge request pipeline" ;;
      CI_COMMIT_TAG) b_negative "tag pipelines are not a supported profile" ;;
      *) b_negative "RC promotion pipeline carries ${kv%%=*}" ;;
    esac
    out+=" ${kv%%=*}:$RC"
  done
  detail "contamination of the detached promotion refused (variables GitLab leaves empty there):$out"
}
ctl_N46() { rc_checkout refs/heads/release/0.1.0-rc.4; rc_env 0.1.0-rc.4; run_v "${GL[@]}" "${PR[@]}" CI_COMMIT_SHA="$S_SHA" -- --publish-dir ci-evidence
  b_negative "RC commit is not in origin/staging"; detail "exit $RC: 0.1.0-rc.4 on a commit outside staging (RC commit is not in origin/staging)"; }
ctl_N47() { rc_checkout refs/merge-requests/7/head "--prune --quiet +refs/tags/*:refs/tags/*"; rc_env 0.1.0-rc.1
  run_v "${GL[@]}" "${PR[@]}" CI_COMMIT_SHA="$T_SHA" -- --publish-dir ci-evidence
  b_negative "RC refs not prepared: refs/remotes/origin/staging"; detail "exit $RC: fetch without the staging refspec (RC refs not prepared: refs/remotes/origin/staging; no fallback fetch)"; }
ctl_N48() { rc_checkout refs/merge-requests/7/head; rc_env 0.1.0-rc.1
  run_v "${GL[@]}" "${PR[@]}" CI_COMMIT_SHA="$T_SHA" CI_MERGE_REQUEST_EVENT_TYPE=merged_result CI_MERGE_REQUEST_SOURCE_BRANCH_SHA="$T_SHA" CI_MERGE_REQUEST_TARGET_BRANCH_SHA="$S_SHA" -- --publish-dir ci-evidence
  b_negative "RC promotion outside a detached merge request pipeline"; detail "exit $RC: merged-result promotion (RC promotion outside a detached merge request pipeline)"; }
ctl_N49() { rc_checkout refs/merge-requests/7/head; rc_env x; run_v "${GL[@]}" "${PR[@]}" CI_COMMIT_SHA="$T_SHA" -- --publish-dir ci-evidence
  b_negative "RC promotion source branch is not release/ and a canonical RC name"
  detail "exit $RC: an ordinary merge request from release/x to main is refused, never mr-source (a release/ prefix alone never qualifies)"; }
ctl_RR1() {
  local n
  n="$(grep -c '^ *GIT_FETCH_EXTRA_FLAGS: ' "$REPO_R/.gitlab-ci.yml" || true)"
  if [[ "$n" != 2 || "$RC_FLAGS" != "--prune --quiet +refs/heads/staging:refs/remotes/origin/staging +refs/tags/*:refs/tags/*" ]]; then
    prob "the promotion rules do not both set exactly the reviewed GIT_FETCH_EXTRA_FLAGS ($n, [$RC_FLAGS])"
  fi
  rc_checkout refs/merge-requests/7/head
  if [[ "$(git -C "$C/repo" cat-file -t refs/tags/0.1.0-rc.1 2>/dev/null)" != tag ]]; then prob "the annotated tag object was not fetched"; fi
  if [[ "$(git -C "$C/repo" rev-parse refs/remotes/origin/staging 2>/dev/null)" != "$T_SHA" ]]; then prob "refs/remotes/origin/staging was not prepared"; fi
  rc_env 0.1.0-rc.1; run_v "${GL[@]}" "${PR[@]}" CI_COMMIT_SHA="$T_SHA" -- --profile fast --publish-dir ci-evidence
  expect_rc 0; b_positive mr-rc-promotion
  detail "exit $RC; the merge request head fetch plus the .gitlab-ci.yml flags (both rules, word-split) prepares the annotated tag and origin/staging; accepted"
}
ctl_RR2() { rc_checkout refs/merge-requests/7/head "--prune --quiet"; rc_env 0.1.0-rc.1
  run_v "${GL[@]}" "${PR[@]}" CI_COMMIT_SHA="$T_SHA" -- --publish-dir ci-evidence
  b_negative "RC refs not prepared: refs/remotes/origin/staging"
  detail "exit $RC: the default flags only (no extra refspecs) leave origin/staging unprepared; refused without a fallback fetch"; }

# --- group C: evidence scanning ------------------------------------------------------------------------
ctl_V5a() { std_ci; run_v "${CIENV[@]}" STUB_HARNESS=plant-retained -- --publish-dir ci-evidence; expect_rc 15; expect_pub minimal:WITHHELD; expect_no_canary; expect_f1f2; detail "exit $RC; minimal WITHHELD; canary absent"; }
ctl_V5d() { std_ci; run_v "${CIENV[@]}" STUB_HARNESS=plant-ledger-context -- --publish-dir ci-evidence; expect_rc 15; expect_pub minimal:WITHHELD; expect_no_canary; expect_f1f2; detail "exit $RC; canary in a ledger-summary.tsv unknown_context line: minimal WITHHELD, canary absent"; }
ctl_V5b_planted() { std_ci; run_v "${CIENV[@]}" STUB_HARNESS=plant-console -- --publish-dir ci-evidence; expect_rc 15; expect_pub minimal:WITHHELD; expect_no_canary; detail "exit $RC; console-only canary withheld"; }
ctl_V5b_clean() {
  std_ci; run_v "${CIENV[@]}" -- --publish-dir ci-evidence; expect_rc 0; expect_pub full
  if [[ -e "$(PUBDIR)/harness-console.log" ]]; then prob "harness-console.log was published"; fi
  if grep -q 'harness console' "$C/stub/stub-tee.log" 2>/dev/null; then :; else prob "the stub wrote no console output"; fi
  if [[ ! -f "$(PUBDIR)/harness.log" ]]; then prob "scanned harness.log missing"; fi
  detail "exit $RC; full; raw console excluded by name, scanned harness.log published"
}
ctl_V5c() { std_ci; run_v "${CIENV[@]}" STUB_HARNESS=plant-gitleaks -- --publish-dir ci-evidence; expect_rc 15; expect_pub minimal:WITHHELD
  if grep -q 'gitleaks reported a finding' "$C/out"; then :; else prob "the finding did not come from gitleaks"; fi; detail "exit $RC; gitleaks finding withheld"; }
ctl_V6a() { std_ci; run_v "${CIENV[@]}" SHIM_GITLEAKS_FAIL=1 -- --publish-dir ci-evidence; expect_rc 17; expect_pub none; expect_f1f2; detail "exit $RC; nothing published"; }
ctl_V6b() { std_ci; run_v "${CIENV[@]}" STUB_HARNESS=no-secrets -- --publish-dir ci-evidence; expect_rc 17; expect_pub none; detail "exit $RC; secrets file missing"; }
ctl_V6c() { std_ci; run_v "${CIENV[@]}" -- --publish-dir ci-evidence; expect_rc 0; expect_pub full; expect_phase raw_scan 0; expect_phase candidate_scan 0; detail "exit $RC; both scans clean"; }
ctl_V6d() { std_ci; run_v "${CIENV[@]}" STUB_HARNESS=unreadable -- --publish-dir ci-evidence; expect_rc 17; expect_pub none; detail "exit $RC; unreadable raw file"; }

# --- group D: finalization, supervision and location ------------------------------------------------------
ctl_V14() { std_ci; run_v "${CIENV[@]}" STUB_HARNESS=malformed -- --publish-dir ci-evidence; expect_rc 14; expect_pub full; expect_f1f2; detail "exit $RC; full"; }
ctl_V15() { std_ci; run_v "${CIENV[@]}" STUB_HARNESS=browser-malformed -- --publish-dir ci-evidence; expect_rc 14; expect_pub full; expect_f1f2
  if [[ -e "$(PUBDIR)/browser-root-results.json" ]]; then prob "unsanitizable browser JSON was published"; fi; detail "exit $RC; full without the malformed JSON"; }
ctl_V16() { std_ci; run_v "${CIENV[@]}" STUB_HARNESS=partial -- --publish-dir ci-evidence; expect_rc 19; expect_phase stage 19; expect_pub minimal:STAGING-FAILED; detail "exit $RC; minimal STAGING-FAILED, nothing partial"; }
ctl_V16b() { std_ci; run_v "${CIENV[@]}" CI_PIPELINE_ID=not-a-number -- --publish-dir ci-evidence; expect_rc 19; expect_phase summary 19; expect_pub none; detail "exit $RC; none"; }
ctl_V17() {
  std_ci; chmod 555 "$C/repo"
  run_v "${CIENV[@]}" -- --publish-dir ci-evidence
  chmod 755 "$C/repo"
  expect_rc 19; expect_phase publish 19; expect_pub none; detail "exit $RC; nothing published, no .tmp"
}
ctl_V17b() { std_ci; run_v "${CIENV[@]}" STUB_HARNESS=locked -- --publish-dir ci-evidence; expect_rc 19; expect_phase scrub 19; expect_pub full; detail "exit $RC; full published; scrub=19"; }
ctl_V18() {
  std_ci; start_v "${CIENV[@]}" STUB_HARNESS=sleep -- --publish-dir ci-evidence
  if wait_file "$C/stub/harness-running" 60; then kill -TERM "$VPID"; else prob "the stub harness never started"; fi
  wait_v
  expect_rc 143; expect_group_gone; expect_trap yes; expect_f1f2; expect_pub full
  expect_line has signal=143; expect_line has harness=143; expect_stop signal; expect_line not internal=
  detail "exit $RC; group gone; harness cleanup ran; signal=143 harness=143"
}
ctl_V18a() {
  std_ci; start_v "${CIENV[@]}" STUB_HARNESS=ignore-term -- --publish-dir ci-evidence
  if wait_file "$C/stub/harness-running" 60; then kill -TERM "$VPID"; else prob "the stub harness never started"; fi
  wait_v
  expect_rc 143; expect_group_gone; expect_trap no; expect_f1f2; expect_line has signal=143
  if grep -q 'sending KILL' "$C/out"; then :; else prob "no KILL after the grace period"; fi
  detail "exit $RC; KILL after the 150 s grace; group gone"
}
ctl_V18b() { std_ci; run_v "${CIENV[@]}" STUB_HARNESS=fail SHIM_FAIL_PS_FILTER=1 -- --publish-dir ci-evidence; expect_rc 1; expect_phase cleanup 18; detail "exit $RC; cleanup=18 reported"; }
ctl_V18c() { std_ci; run_v "${CIENV[@]}" STUB_PREFLIGHT=fail -- --publish-dir ci-evidence; expect_rc 16; expect_runs 0; expect_f1f2; expect_pub full
  if [[ -f "$(PUBDIR)/runner-preflight.tsv" ]]; then :; else prob "preflight evidence not published"; fi; detail "exit $RC; no harness; preflight evidence published"; }
ctl_V18d() {
  local w s i
  std_ci; start_v "${CIENV[@]}" STUB_HARNESS=sleep -- --publish-dir ci-evidence
  if wait_file "$C/stub/harness-running" 60; then
    # validate.sh starts the watchdog only after the harness group is ready, so the stub's marker can come first.
    for ((i = 0; i < 100; i++)); do
      w="$(ps -eo pid=,ppid=,args= | awk -v p="$VPID" '$2 == p && / watchdog / {print $1}')" || w=""
      s="$(ps -eo pid=,ppid=,args= | awk -v p="${w:-none}" '$2 == p && $3 == "sleep" {print $1}')" || s=""
      if [[ -n "$s" ]]; then break; fi
      sleep 0.1
    done
    if [[ -n "$s" ]]; then kill -TERM "$s"; else prob "watchdog not found"; kill -TERM "$VPID"; fi
  else prob "the stub harness never started"; fi
  wait_v
  expect_rc 124; expect_group_gone; expect_trap yes; expect_stop deadline; expect_line has harness=124; expect_line not internal=; expect_f1f2
  detail "exit $RC; watchdog deadline (its sleep ended from outside); group gone"
}
ctl_V18e() {
  local r d f got want unit=0 H S8
  # (1) unit cases of the real state_import (lib.sh of the scratch tree)
  new_case
  d="$C/import"; mkdir -p "$d"
  H="$(printf 'a%.0s' {1..40})"
  S8='HEAD_SHA=^[0-9a-f]{40}$'
  for r in "valid:accept:HEAD_SHA=$H\nBUDGET_S=8700\n" "unknown-key:reject:HEAD_SHA=$H\nBUDGET_S=8700\nEVIL=1\n" \
    "duplicate:reject:HEAD_SHA=$H\nHEAD_SHA=$H\nBUDGET_S=8700\n" "bad-value:reject:HEAD_SHA=$H\nBUDGET_S=87000000\n" \
    "missing-key:reject:HEAD_SHA=$H\n" "injection:reject:HEAD_SHA=$H\nBUDGET_S=1;touch $d/pwn\n" \
    "lowercase:reject:head_sha=$H\nBUDGET_S=5\n" "symlink:reject:HEAD_SHA=$H\nBUDGET_S=5\n"; do
    f="$d/${r%%:*}.env"; want="${r#*:}"; want="${want%%:*}"
    printf '%b' "${r#*:*:}" >"$d/real-${r%%:*}"
    if [[ "${r%%:*}" == symlink ]]; then ln -s "$d/real-${r%%:*}" "$f"; else mv "$d/real-${r%%:*}" "$f"; fi
    got="$(bash -c 'source "$1"; unset HEAD_SHA BUDGET_S; if state_import "$2" "$3" "$4"; then echo accept; else echo reject; fi' \
      import "$C/repo/infra/ci/lib.sh" "$f" "$S8" 'BUDGET_S=^[1-9][0-9]{0,5}$')"
    if [[ "$got" == "$want" ]]; then unit=$((unit + 1)); else prob "state_import ${r%%:*}: $got, want $want"; fi
  done
  if [[ -e "$d/pwn" ]]; then prob "state_import injection had a side effect"; fi
  # (2) a pre-planted state file (symlink): the preflight write refuses it
  std_ci 4245; run_v "${CIENV[@]}" STUB_PREFLIGHT=block-p1 -- --publish-dir ci-evidence
  expect_rc 16; expect_runs 0
  # (3) an out-of-range value reaches the importer (CI_JOB_TIMEOUT is GitLab-defined)
  std_ci 4246; CIENV=("${CIENV[@]/#CI_JOB_TIMEOUT=*/CI_JOB_TIMEOUT=99999999}")
  CIENV=("${CIENV[@]/#RUNNER_SCRIPT_TIMEOUT=*/RUNNER_SCRIPT_TIMEOUT=}")
  run_v "${CIENV[@]}" -- --publish-dir ci-evidence
  expect_rc 16; expect_runs 0; expect_phase state 16
  detail "state_import unit cases $unit/8; planted p1.env -> 16; out-of-range BUDGET_S -> import 16; harness 0"
}
ctl_V18f() {
  std_ci; run_v "${CIENV[@]}" STUB_PREFLIGHT=block-pgid STUB_HARNESS=sleep -- --publish-dir ci-evidence
  expect_rc 143; expect_group_gone; expect_f1f2; expect_line has harness=143; expect_line has internal=19
  expect_stop finalizer; expect_line not signal=
  detail "exit $RC; harness=143 internal=19 harness_stop=finalizer; no signal"
}
pre_loc() { # ARGMODE: the pre-location cases
  local inv1 inv2 tmpdir="$C/tmp" args=()
  case "$1" in
    tmp-inside) tmpdir="$C/repo/tmp-inside" ;;
    ev-inside) args=(--evidence-dir "$C/repo/ev") ;;
    ev-symlink) ln -s "$C/repo" "$C/link"; args=(--evidence-dir "$C/link/ev") ;;
  esac
  inv1="$(find "$C/repo" -printf '%p %y %s %T@\n' | LC_ALL=C sort | sha256sum)"
  case_state
  launch "$C/out" env -i -C "$C/repo" "${BASE_ENV[@]}" G003_STATE_ROOT="$CSTATE" TMPDIR="$tmpdir" STUB_DIR="$C/stub" SHIM_PULL_NOOP=1 \
    bash "$C/repo/infra/ci/validate.sh" "${args[@]}"
  await "$LPID" 300
  parse_final
  inv2="$(find "$C/repo" -printf '%p %y %s %T@\n' | LC_ALL=C sort | sha256sum)"
  expect_rc 16; expect_phase location 16; expect_runs 0; expect_line not internal=
  if [[ -z "${PH[postcheck]+x}" ]]; then prob "postcheck did not run"; fi
  expect_phase cleanup n/a; expect_no_phase preflight
  if [[ "$F_PUB" != none || "$F_REG" != none ]]; then prob "published=$F_PUB registry=$F_REG, want none and none"; fi
  if [[ "$inv1" != "$inv2" ]]; then prob "the repository tree changed"; fi
  no_validate_left
}
ctl_V18g() { new_case; pre_loc tmp-inside; detail "exit $RC; TMPDIR inside the repository; tree unchanged"; }
ctl_V18h() { new_case; pre_loc ev-inside; detail "exit $RC; --evidence-dir inside; tree unchanged"; }
ctl_V18i() { new_case; pre_loc ev-symlink; detail "exit $RC; --evidence-dir through a symlink into the repository; tree unchanged"; }
ctl_V18j() { std_ci; run_v "${CIENV[@]}" STUB_HARNESS=stray -- --publish-dir ci-evidence; expect_rc 0; expect_group_gone; expect_pub full
  if grep -q 'stragglers=1' "$C/out"; then :; else prob "stragglers not reported"; fi; detail "exit $RC; stragglers stopped; group gone"; }
ctl_V18k() { std_ci; run_v "${CIENV[@]}" STUB_HARNESS=block-result -- --publish-dir ci-evidence; expect_rc 19; expect_line has harness=0; expect_line has internal=19
  expect_line not signal=; expect_f1f2; expect_pub full; expect_phase report 0; expect_phase scrub 0; detail "exit $RC; harness=0 internal=19; F1-F11 ran"; }

# --- group D: budgets and capacity (infra/ci/lib.sh of the case tree, clean environments) -------------------------------
lib_out() { # ENV... -- EXPR: EXPR evaluated by bash with the case tree's infra/ci/lib.sh sourced; prints its output
  local envs=()
  while (( $# > 0 )) && [[ "$1" != -- ]]; do envs+=("$1"); shift; done
  shift
  # shellcheck disable=SC2016  # the inner shell's own positional parameters
  env -i PATH="$PATH" HOME="${HOME:-/nonexistent}" LC_ALL=C "${envs[@]}" bash -c 'source "$1"; eval "$2"' lib "$C/repo/infra/ci/lib.sh" "$1" 2>/dev/null || true
}
ago() { date -u -d "@$(( $(date -u +%s) - $1 ))" +%Y-%m-%dT%H:%M:%SZ; }
ctl_V36() {
  local r args want got n=0 hosted_want=fail
  local G4=4194304 G7=7340032 G8=8388608 G15=15728640 G36=37748736 G39=40894464 G40=41943040
  new_case
  for r in "fast 2 $G4 $G15 same:ok" "fast 1 $G4 $G15 same:fail" "fast 2 $((G4 - 1)) $G15 same:fail" \
    "fast 2 $G4 $((G15 - 1)) same:fail" "fast 2 $G4 $G15 $((G15 - 1)):fail" \
    "integration 4 $G8 $G40 $G40:ok" "integration 4 $G8 $G39 $G40:fail" "integration 4 $G8 $G40 $G39:fail" \
    "integration 3 $G8 $G40 same:fail" "integration 4 $G7 $G40 same:fail" \
    "full 4 $G8 $G40 same:ok" "full 4 $G8 $G39 same:fail" "full 3 $G8 $G40 same:fail" "full 4 $G7 $G40 same:fail" \
    "full x $G8 $G40 same:fail" "full 4 $G8 $G40 none:fail" "gate 4 $G8 $G40 same:fail"; do
    args="${r%%:*}" want="${r##*:}"
    got="$(lib_out -- "if ci_capacity_ok $args; then echo ok; else echo fail; fi")"
    if [[ "$got" == "$want" ]]; then n=$((n + 1)); else prob "ci_capacity_ok $args: $got, want $want"; fi
  done
  # The hosted 35 GiB floor: never from a caller flag; only in the qualified adapter context (variables plus the
  # client certificate and /.dockerenv of a container).
  got="$(lib_out PLATFORM_CI_CONTEXT=hosted-context/1 -- "if ci_capacity_ok full 4 $G8 $G36 same; then echo ok; else echo fail; fi")"
  if [[ "$got" == fail ]]; then n=$((n + 1)); else prob "a caller flag activated the hosted floor"; fi
  if [[ -f /certs/client/cert.pem && -r /certs/client/cert.pem && -f /.dockerenv ]]; then hosted_want=ok; fi
  got="$(lib_out GITLAB_CI=true CI_JOB_ID=1 PLATFORM_CI_CONTEXT=hosted-context/1 DOCKER_HOST=tcp://localhost:2376 DOCKER_TLS_VERIFY=1 \
    DOCKER_CERT_PATH=/certs/client -- "if ci_capacity_ok full 4 $G8 $G36 same; then echo ok; else echo fail; fi")"
  if [[ "$got" == "$hosted_want" ]]; then n=$((n + 1)); else prob "36 GiB with the full adapter variables: $got, want $hosted_want"; fi
  got="$(lib_out GITLAB_CI=true CI_JOB_ID=1 PLATFORM_CI_CONTEXT=hosted-context/1 DOCKER_HOST=tcp://localhost:2376 DOCKER_TLS_VERIFY=1 \
    DOCKER_CERT_PATH=/certs/client -- "if ci_capacity_ok fast 2 $G4 $((G15 - 1)) same; then echo ok; else echo fail; fi")"
  if [[ "$got" == fail ]]; then n=$((n + 1)); else prob "the hosted context lowered the fast floor"; fi
  detail "ci_capacity_ok $n/20 (fast 2 CPU/4 GiB/15 GiB and just below; integration/full 4/8/40 and 39 GiB, 3 CPU, 7 GiB; malformed; a caller flag never selects the hosted 35 GiB floor; adapter variables here: 36 GiB -> $hosted_want)"
}
ctl_V37() {
  local r want got n=0 b
  new_case
  for r in "fast 1800:ok" "fast 900:fail" "gate 10800:ok" "gate 3600:fail" "fast 1h:fail" "full 10800:fail" \
    "staging 2700:ok" "staging 2699:fail" "staging 45m:fail"; do
    want="${r##*:}"; got="$(lib_out -- "if ci_time_ok ${r%%:*}; then echo ok; else echo fail; fi")"
    if [[ "$got" == "$want" ]]; then n=$((n + 1)); else prob "ci_time_ok ${r%%:*}: $got, want $want"; fi
  done
  for r in "CI_JOB_TIMEOUT=3h RUNNER_SCRIPT_TIMEOUT=2h45m CI_JOB_STARTED_AT=$(ci_utc)" \
    "CI_JOB_TIMEOUT=10800 RUNNER_SCRIPT_TIMEOUT=2hours CI_JOB_STARTED_AT=$(ci_utc)" \
    "CI_JOB_TIMEOUT=10800 RUNNER_SCRIPT_TIMEOUT=2h45m CI_JOB_STARTED_AT=2026-10-06T20:00:00" \
    "CI_JOB_TIMEOUT=10800 RUNNER_SCRIPT_TIMEOUT=2h45m" "RUNNER_SCRIPT_TIMEOUT=2h45m CI_JOB_STARTED_AT=$(ci_utc)"; do
    # shellcheck disable=SC2086  # r is a fixed list of NAME=VALUE words
    got="$(lib_out GITLAB_CI=true $r -- 'ci_profile_budget_seconds full || echo refused')"
    if [[ "$got" == refused ]]; then n=$((n + 1)); else prob "budget with [$r]: $got, want refused"; fi
  done
  got="$(lib_out GITLAB_CI=true CI_JOB_TIMEOUT=10800 RUNNER_SCRIPT_TIMEOUT=2h45m CI_JOB_STARTED_AT="$(ci_utc)" -- 'ci_profile_budget_seconds fast; ci_profile_budget_seconds integration; ci_profile_budget_seconds full' | tr '\n' ' ')"
  read -r -a b <<<"$got"
  if [[ "${b[0]:-}" == 1500 && "${b[1]:-}" == 5400 && "${b[2]:-0}" -ge 8690 && "${b[2]:-0}" -le 8700 ]]; then n=$((n + 1)); else prob "profile caps: [$got], want 1500 5400 8690-8700"; fi
  got="$(lib_out GITLAB_CI=true CI_JOB_TIMEOUT=2400 RUNNER_SCRIPT_TIMEOUT=35m CI_JOB_STARTED_AT="$(ago 120)" -- 'ci_profile_budget_seconds fast; ci_profile_budget_seconds affected' | tr '\n' ' ')"
  if [[ "$got" == "1500 1500 " ]]; then n=$((n + 1)); else prob "configured fast/affected after 120 s startup: [$got], want 1500 1500"; fi
  got="$(lib_out GITLAB_CI=true CI_JOB_TIMEOUT=10800 RUNNER_SCRIPT_TIMEOUT=2h45m CI_JOB_STARTED_AT="$(ago 9000)" -- 'b="$(ci_profile_budget_seconds full)"; if (( b < CI_MINBUDGET_full )); then echo below; else echo "$b"; fi')"
  if [[ "$got" == below ]]; then n=$((n + 1)); else prob "9000 s elapsed: budget $got, want below the full minimum"; fi
  got="$(lib_out -- 'ci_profile_budget_seconds integration')"
  if [[ "$got" == 14400 ]]; then n=$((n + 1)); else prob "local budget $got, want 14400"; fi
  # End to end: a malformed start time and an exhausted budget each refuse before the harness.
  std_ci 4260; CIENV=("${CIENV[@]/#CI_JOB_STARTED_AT=*/CI_JOB_STARTED_AT=yesterday}")
  run_v "${CIENV[@]}" -- --profile fast --publish-dir ci-evidence
  expect_rc 16; expect_runs 0; expect_phase preflight 16; log_has "preflight: CI_JOB_TIMEOUT, RUNNER_SCRIPT_TIMEOUT or CI_JOB_STARTED_AT is missing or malformed"
  std_ci 4261; CIENV=("${CIENV[@]/#CI_JOB_STARTED_AT=*/CI_JOB_STARTED_AT=$(ago 9000)}")
  run_v "${CIENV[@]}" -- --publish-dir ci-evidence
  expect_rc 16; expect_runs 0; expect_phase preflight 16
  if grep -qE '^validate: preflight: harness budget -?[0-9]+s is below 3600s \(full\)$' "$C/out"; then :; else prob "the exhausted budget is not reported"; fi
  detail "time and budget unit cases $n/18 (job floors fast 1800/gate 10800/staging 2700; malformed timeouts and start time refused; caps fast 1500, integration 5400, full uncapped; 9000 s elapsed below the full minimum; local 14400); end to end: malformed start and exhausted budget exit 16 before the harness"
}

# --- group D: cancellation of this runner (V18l-V18n) -----------------------------------------------------
# Fixture nested-hold (only by --only): the nested active validator of V18l-V18n. Its validate.sh runs a
# stub harness that ignores TERM, INT and HUP, with a descendant that does too and creates a control
# container every 2 s.
ctl_nested_hold() {
  if [[ -n "${CI_CONTROLS_HOLD_SIM:-}" ]]; then hold_sim "$CI_CONTROLS_HOLD_SIM"; return 0; fi
  std_ci; start_v "${CIENV[@]}" STUB_HARNESS=hold STUB_OWNED_LOG="${CI_CONTROLS_OUTER_OWNED:-}" -- --publish-dir ci-evidence
  printf '%s\n' "$VPID" >"$C/stub/validate.pid"
  if wait_file "$C/stub/holding" 60; then :; else prob "the holding harness never started"; fi
  wait_v
  prob "the hold fixture ended (exit $RC)"
}
# CI_CONTROLS_HOLD_SIM (G064 correction 11; the cleanup driver of V18n, no Docker daemon; nested_cancel starts its
# nested runner with env -i, so never there): the fixture is instead a creator standing in for a nested run's holder,
# a registered launch that records itself and every container it creates (one per 0.5 s; a foreign run's labels and
# this control's sdlc.ci.outer label, as the leaked V18m containers carried) in this control's inventory. MODE: startup
# (the creator ignores TERM; the fixture fails before holding and returns while it runs), hold (holding, then a
# bounded wait: the deadline under --launch-deadline, or the runner is cancelled), late (the creator stopped; an
# unrecorded creator outside the control adds one labelled container 1 s after the runner's removal began), left (the
# creator stopped; its containers left to the runner), done (the fixture also removes them itself).
# G064 corr1 C2 / corr2 D2/D4: with CI_CONTROLS_HOLD_SIM_GATE=1 the creator goes through this run's docker shim in the
# nested-holder scope (SHIM_GATE_OUTER = this inventory, its sdlc.ci.outer label), as the real nested holder does; then
# also race (ceiling - 1 sequential creations, then two concurrent shim creators: exactly one creates) and latch (a
# nested case shell sets the latch; the parent's shim creation is refused, nothing created).
hold_sim() {
  local mode="$1" f hf hp st first ign="" i lab=() gd v ok=0
  new_case; ctl_labels; f="$(owned_file)"; hf="$f"
  lab=(--label "sdlc.ci.control=n$RT-nested-hold" --label "sdlc.ci.run=n$RT" --label "sdlc.ci.outer=$RT-$CTL_ID" --label "$OWNER_LABEL")
  gd="$REAL_DOCKER"
  if [[ "${CI_CONTROLS_HOLD_SIM_GATE:-}" == nested ]]; then
    # G064 r13: the real nested-run shape (this runner is nested: CI_CONTROLS_OUTER_OWNED, --extra-label sdlc.ci.outer=...):
    # the creator stands in for validate.sh's stub holder: `docker create` through the shim with this case's labels (this
    # run's own sdlc.ci.control label and the outer label), gated in the outer scope (BASE_ENV SHIM_GATE_OUTER); it records
    # itself and each container in the OUTER inventory (as STUB_OWNED_LOG), never in this control's.
    if [[ ! -f "$OUTER_INV" ]]; then prob "CI_CONTROLS_HOLD_SIM_GATE=nested needs CI_CONTROLS_OUTER_OWNED"; return 0; fi
    hf="$OUTER_INV" lab=() gd="$C/gdocker"
    { printf '#!/usr/bin/env bash\nexec env'
      for i in "${BASE_ENV[@]}" SHIM_LABELS="$(case_labels)"; do printf ' %q' "$i"; done
      printf ' %q "$@"\n' "$ROOT/shim/docker"; } >"$gd"
    chmod 700 "$gd"
  elif [[ "${CI_CONTROLS_HOLD_SIM_GATE:-}" == 1 || "$mode" == race || "$mode" == latch ]]; then
    gd="$C/gdocker" v="${lab[*]}"; v="${v//--label /}"
    { printf '#!/usr/bin/env bash\nexec env'
      for i in "${BASE_ENV[@]}" SHIM_GATE_OUTER="$f" SHIM_LABELS="$v"; do printf ' %q' "$i"; done
      printf ' %q "$@"\n' "$ROOT/shim/docker"; } >"$gd"
    chmod 700 "$gd"
  fi
  case "$mode" in
    race)
      v="$(gate_ceil outer "$CTL_ID" container)"
      for ((i = 1; i < v; i++)); do "$gd" create "${lab[@]}" --network none "$CI_BUSYBOX_IMAGE" true >/dev/null 2>"$C/race-$i.err" || prob "race: sequential creation $i of $((v - 1)) refused: $(cat "$C/race-$i.err")"; done
      "$gd" create "${lab[@]}" --network none "$CI_BUSYBOX_IMAGE" true >"$C/race-a.out" 2>"$C/race-a.err" & i=$!
      "$gd" create "${lab[@]}" --network none "$CI_BUSYBOX_IMAGE" true >"$C/race-b.out" 2>"$C/race-b.err" || :
      wait "$i" || :
      ok="$(cat "$C/race-a.out" "$C/race-b.out" | grep -c .)" || ok=0
      if [[ "$ok" != 1 ]]; then prob "race: $ok of 2 concurrent creators at ceiling - 1 created (want exactly 1)"; fi
      detail "simulated race: ceiling $v; $ok of 2 concurrent creators created; $(grep -c '^container' "$f") recorded containers"
      return 0 ;;
    latch)
      # shellcheck disable=SC2016  # the nested case shell's own script
      bash -c 'source "$1"; gate_latch "$2" "nested case shell breach"' latch "$ROOT/shim/gate.sh" "${f%.tsv}"
      if "$gd" create "${lab[@]}" --network none "$CI_BUSYBOX_IMAGE" true >"$C/latch.out" 2>"$C/latch.err"; then prob "latch: the parent shim created despite the latch set in a nested case shell"; fi
      detail "simulated latch: parent shim refusal [$(head -1 "$C/latch.err")]; $(grep -c '^container' "$f") recorded containers"
      return 0 ;;
    unrec)
      # an object carrying this control's own label created outside both allocation paths: no inventory row
      "$REAL_DOCKER" create "${LB[@]}" --network none "$CI_BUSYBOX_IMAGE" true >"$C/unrec.id"
      detail "simulated unrec: $(cut -c1-12 "$C/unrec.id")"
      return 0 ;;
  esac
  if [[ "$mode" == startup ]]; then ign=ignore; fi
  # shellcheck disable=SC2016  # the creator's own script
  launch "$C/holder.out" bash -c 'if [[ "$1" == ignore ]]; then trap "" TERM INT HUP; fi
    f="$2" stop="$3" d="$4" img="$5"; shift 5
    read -r s <"/proc/$BASHPID/stat"; s="${s##*) }"; read -r -a a <<<"$s"
    printf "process\t%s\t%s\tsimulated holder\n" "$BASHPID" "${a[19]}" >>"$f"
    while [[ ! -e "$stop" ]]; do
      if id="$("$d" create "$@" --network none "$img" true 2>/dev/null)"; then printf "container\t%s\tsimulated holder\n" "$id" >>"$f"; fi
      sleep 0.5
    done' holder "$ign" "$hf" "$C/holder.stop" "$gd" "$CI_BUSYBOX_IMAGE" "${lab[@]}"
  hp="$LPID"
  # (unique IDs: a gated creation is recorded by the shim as well as by the creator)
  for ((i = 0; i < 100; i++)); do if (( $(awk -F'\t' '$1 == "container" && !u[$2]++' "$hf" | grep -c .) >= 2 )); then break; fi; sleep 0.1; done
  first="$(awk -F'\t' '$1 == "container" {print $2; exit}' "$hf")"
  case "$mode" in
    startup) prob "the holding harness never started (simulated: its TERM-ignoring creator still runs)" ;;
    hold) : >"$C/holding"; await "$hp" 400; prob "the hold fixture ended (exit $RC)" ;;
    late|left|done)
      # The creator ends between two creations (its stop file), so each container it made is recorded.
      st="$(proc_start "$hp")" || st=""
      : >"$C/holder.stop"
      for ((i = 0; i < 50; i++)); do if [[ -n "$st" ]] && proc_is "$hp" "$st"; then sleep 0.1; else break; fi; done
      if stop_tree "" "$hp:$st"; then wait "$hp" 2>/dev/null || true; else prob "the creator survived KILL: $STOP_LEFT"; fi
      if [[ "$mode" == late ]]; then
        # shellcheck disable=SC2016  # the late creator's own script
        setsid bash -c 'd="$1" first="$2" img="$3"; shift 3
          for ((i = 0; i < 600; i++)); do if "$d" inspect --type container "$first" >/dev/null 2>&1; then sleep 0.1; else break; fi; done
          sleep 1; "$d" create "$@" --network none "$img" true' late "$REAL_DOCKER" "$first" "$CI_BUSYBOX_IMAGE" "${lab[@]}" </dev/null >/dev/null 2>&1 &
      elif [[ "$mode" == 'done' ]]; then
        while IFS=$'\t' read -r i first _; do if [[ "$i" == container ]]; then D rm -f "$first" >/dev/null 2>&1 || true; fi; done <"$hf"
      fi ;;
    *) prob "unknown simulation $mode" ;;
  esac
  detail "simulated $mode: $(grep -c '^container' "$hf") recorded containers"
}
# Fixture reg-hold (only by --only; G064 C4, VR9 (b)): in a self-registered top-level run, one container created through D
# (recorded in the inventory and mirrored into the run registry), then a hold of CI_CONTROLS_REG_HOLD_S seconds (default
# 600; VR9 KILLs the run meanwhile). CI_CONTROLS_REG_HOLD_OUT names a file that receives "<container ID> <run registry>".
ctl_reg_hold() {
  local id o="${CI_CONTROLS_REG_HOLD_OUT:-}" s="${CI_CONTROLS_REG_HOLD_S:-600}"
  ctl_labels
  id="$(D create "${LB[@]}" --network none "$CI_BUSYBOX_IMAGE" true)" || { prob "the fixture container was not created"; return 0; }
  if [[ -n "$o" ]]; then printf '%s %s\n' "$id" "${SR_REG:-none}" >"$o.tmp" && mv -f -- "$o.tmp" "$o"; fi
  if [[ "$s" =~ ^[0-9]{1,4}$ ]]; then sleep "$s"; fi
  detail "created ${id:0:12} through D and held ${s} s"
}
# Fixture obj-codes (only by --only): the producer/extractor of the B-OBJ-IT failed-subtest codes
# (infra/validation/lib.sh obj_fail_codes, G014 object diagnostic publication). Crafted go test -v logs in the case's
# scratch tree, the tracked lib.sh of the template sourced in a strict-mode subshell; every output must be the closed
# grammar, the expected token list, and free of every planted private or spoofed string; helper errors leave the
# record expression a FAIL row with `failed-subtests: unknown`. CI_CONTROLS_HOLD_SIM_NEXT (the V18n cleanup driver only,
# like CI_CONTROLS_HOLD_SIM): obj-codes is instead a second simulated allocating control in that mode, selected after
# nested-hold, so that the driver can show that a held cleanup starts no later control (no creation by it).
ctl_obj_codes() {
  if [[ -n "${CI_CONTROLS_HOLD_SIM_NEXT:-}" ]]; then hold_sim "$CI_CONTROLS_HOLD_SIM_NEXT"; return 0; fi
  new_case
  local lib="$C/repo/infra/validation/lib.sh" d="$C/stub/obj-codes" names=() nm got cx url asrt pub=0 fails=0
  local gram='^((OBJ-[A-Z-]+|PARENT|FOOTPRINT)(,(OBJ-[A-Z-]+|PARENT|FOOTPRINT))*(,unknown)?|unknown|none)$'
  mkdir -p "$d"
  if [[ ! -f "$lib" ]]; then prob "the template lacks infra/validation/lib.sh"; return; fi
  # The ten exact Go-printed subtest names of TestObjects (core/internal/tenancy/access/objects_integration_test.go),
  # kept here as the fixture's own expectation.
  names=('OBJ-ATTEST_runtime_login_and_scoped_credential' 'OBJ-READ_and_OBJ-WRITE:_the_oracle_service_matrix'
    'OBJ-RLS:_the_oracle_RLS_matrix,_guard_transitions,_FORCE_owner,_poisoned_context' 'OBJ-RT_round_trips'
    'OBJ-SCOPE_scope_discovery_is_constant_in_the_binding_count' 'OBJ-SCHEMA_constraints,_generated_key_and_forged_binding'
    'OBJ-DEL_physical_deletion,_tombstone_and_idempotent_repeat' 'OBJ-SIGN_real_fetch,_expiry,_alteration_and_loopback_binding'
    'OBJ-ADMIN_the_scoped_credential_administers_nothing;_anonymous_access_is_denied' 'OBJ-LOG_records_catalogued,_no_canary_or_secret')
  sub() { printf '    --- %s: TestObjects/%s (0.10s)\n' "$1" "$2"; }            # STATUS NAME
  top() { printf -- '--- %s: TestObjects (11.00s)\n' "$1"; }                     # STATUS
  foot() { printf -- '--- %s: TestObjectsFootprint (1.00s)\n' "$1"; }           # STATUS
  subs() { local n; for n in "${names[@]}"; do sub PASS "$n"; done; }           # the ten, all PASS
  # codes LOG N -> stdout of obj_fail_codes under set -euo pipefail in a fresh shell; the fixture fails on a non-zero exit.
  codes() {
    local rc=0
    got="$(bash -c 'set -euo pipefail; source "$1"; obj_fail_codes "$2" "$3"' _ "$lib" "$1" "$2" 2>>"$C/obj-codes.err")" || rc=$?
    if (( rc != 0 )); then prob "$(basename -- "$1"): helper exit $rc"; fails=$((fails + 1)); fi
    if [[ ! "$got" =~ $gram ]]; then prob "$(basename -- "$1"): output [$got] is outside the closed grammar"; fails=$((fails + 1)); fi
  }
  want() { if [[ "$got" != "$3" ]]; then prob "$1: got [$got], want [$3]"; fails=$((fails + 1)); fi; }   # CASE LOG WANT
  # 1 single known failure (assertion lines precede the header, as Go prints them).
  { subs | sed "s|PASS: TestObjects/OBJ-SIGN_|FAIL: TestObjects/OBJ-SIGN_|"; printf '        objects_integration_test.go:1389: before expiry: 403\n'; top FAIL; foot PASS; } >"$d/single.log"
  codes "$d/single.log" FAIL; want single "$d/single.log" OBJ-SIGN
  # 2 multiple, listed in reverse order in the log -> canonical order.
  { sub FAIL "${names[7]}"; sub FAIL "${names[3]}"; top FAIL; foot PASS; } >"$d/multiple.log"
  codes "$d/multiple.log" FAIL; want multiple "$d/multiple.log" OBJ-RT,OBJ-SIGN
  # 3 parent-only failure (setup failed before any subtest).
  { top FAIL; foot PASS; } >"$d/parent.log"
  codes "$d/parent.log" FAIL; want parent "$d/parent.log" PARENT
  # 4 footprint failure with every TestObjects subtest passing.
  { subs; top PASS; foot FAIL; } >"$d/footprint.log"
  codes "$d/footprint.log" FAIL; want footprint "$d/footprint.log" FOOTPRINT
  # 5 unknown subtest name; a known one plus an unknown one.
  { sub FAIL 'OBJ-EVIL_x'; top FAIL; foot PASS; } >"$d/unknown.log"
  codes "$d/unknown.log" FAIL; want unknown "$d/unknown.log" unknown
  { sub FAIL "${names[7]}"; sub FAIL 'OBJ-EVIL_x'; top FAIL; foot PASS; } >"$d/mixed.log"
  codes "$d/mixed.log" FAIL; want mixed "$d/mixed.log" OBJ-SIGN,unknown
  # 6a private and spoofed text around one real failure: a canary, a signed URL, an assertion line, the header text
  #    embedded in an RLS-ACCESSLOG line, a header without the subtest indent, a 1 MiB line.
  cx="SECRET-$(head -c 12 /dev/urandom | od -An -tx1 | tr -d ' \n')"
  url='http://minio:9000/bucket/objects/a/b?X-Amz-Signature=0123456789abcdef&X-Amz-Credential=core-service'
  asrt='        objects_integration_test.go:1389: before expiry: 403'
  { printf 'leak %s\n' "$cx"; printf 'url %s\n' "$url"; printf '%s\n' "$asrt"
    printf 'RLS-ACCESSLOG     --- FAIL: TestObjects/%s (0.10s)\n' "${names[3]}"
    printf -- '--- FAIL: TestObjects/%s (0.04s)\n' "${names[6]}"
    head -c 1048576 /dev/zero | tr '\0' 'x'; printf '\n'
    sub FAIL "${names[7]}"; top FAIL; foot PASS; } >"$d/spoof.log"
  codes "$d/spoof.log" FAIL; want spoof "$d/spoof.log" OBJ-SIGN
  for nm in "$cx" 'X-Amz' 'before expiry' 'OBJ-RT' 'OBJ-DEL' 'xxxx'; do
    if grep -qF -- "$nm" <<<"$got"; then prob "spoof: the output carries [$nm]"; fails=$((fails + 1)); fi
  done
  # 6b a prefix spoof of a code is not a code.
  { sub FAIL 'OBJ-SIGNATURE_leak'; top FAIL; foot PASS; } >"$d/prefix.log"
  codes "$d/prefix.log" FAIL; want prefix "$d/prefix.log" unknown
  # 7 malformed and oversize input: NUL bytes, CRLF, a bare 1 MiB line, an empty file, a missing path, over the cap,
  #   and (as an unprivileged user) an unreadable file.
  { sub FAIL "${names[7]}"; printf 'bin \000 ary\n'; top FAIL; foot PASS; } >"$d/nul.log"
  codes "$d/nul.log" FAIL; want nul "$d/nul.log" unknown
  sed 's/$/\r/' "$d/single.log" >"$d/crlf.log"
  codes "$d/crlf.log" FAIL; want crlf "$d/crlf.log" unknown
  { head -c 1048576 /dev/zero | tr '\0' 'x'; printf '\n'; } >"$d/longline.log"
  codes "$d/longline.log" FAIL; want longline "$d/longline.log" unknown
  : >"$d/empty.log"
  codes "$d/empty.log" FAIL; want empty "$d/empty.log" unknown
  codes "$d/missing.log" FAIL; want missing "$d/missing.log" unknown
  truncate -s 67108865 "$d/big.log"
  codes "$d/big.log" FAIL; want big "$d/big.log" unknown
  rm -f -- "$d/big.log"
  if [[ "$(id -u)" != 0 ]]; then
    cp -- "$d/single.log" "$d/unreadable.log"; chmod 000 "$d/unreadable.log"
    codes "$d/unreadable.log" FAIL; want unreadable "$d/unreadable.log" unknown
    chmod 600 "$d/unreadable.log"
  fi
  # 8 none needs N == 10 and both top-level PASS lines; a truncated or skipped result is unknown.
  { subs; top PASS; foot PASS; } >"$d/none.log"
  codes "$d/none.log" 10; want none "$d/none.log" none
  { subs; top PASS; } >"$d/nofoot.log"
  codes "$d/nofoot.log" 10; want nofoot "$d/nofoot.log" unknown
  { subs; top PASS; foot PASS; printf -- '--- SKIP: TestObjectsOrdering (0.00s)\n'; } >"$d/skip.log"
  codes "$d/skip.log" FAIL; want skip "$d/skip.log" unknown
  # 9 the FAIL record survives a helper error unchanged apart from `failed-subtests: unknown` (strict mode, as run.sh).
  if bash -c 'set -euo pipefail; source "$1"; RESULTS="$2/results.tsv"; WORK="$2"; n=FAIL; log="$2/missing.log"
      record B-OBJ-IT FAIL "stored-object isolation ($n/10 subtests, footprint pass; service none; rls none; attack 17/17, oracle none; sign-host ok; logcheck: ok; failed-subtests: $(obj_fail_codes "$log" "$n"))"' \
      _ "$lib" "$d" >"$d/record.out" 2>&1; then :; else prob "record expression exit under strict mode"; fails=$((fails + 1)); fi
  if grep -qxF $'B-OBJ-IT\tFAIL\tstored-object isolation (FAIL/10 subtests, footprint pass; service none; rls none; attack 17/17, oracle none; sign-host ok; logcheck: ok; failed-subtests: unknown)' "$d/results.tsv" 2>/dev/null; then pub=1; else
    prob "the FAIL row after a helper error is not the expected row"; fails=$((fails + 1)); fi
  detail "obj_fail_codes: single, multiple (canonical order), parent, footprint, unknown and mixed, spoofed/private text excluded, prefix spoof, NUL/CRLF/long-line/empty/missing/oversize$([[ "$(id -u)" != 0 ]] && printf '/unreadable') unknown, none and its truncated/skipped negatives, FAIL row unchanged on helper error (row check $pub); $fails problems"
}
# G065 K1: the listing is taken apart from the count (grep -c's exit 1 on zero lines no longer masks a failed listing),
# untagged images included. Any failed listing prints -1 (unverified) and still returns 0: a nonzero return inside $(...)
# would end the control under its set -e. Every caller tests < 0 before its own comparison.
lcount() { # LABEL: containers, volumes, networks and images carrying LABEL; -1 when a listing failed
  local x n=0 o bad=0
  for x in "ps -aq" "volume ls -q" "network ls -q" "image ls -aq"; do
    # shellcheck disable=SC2086  # x is a fixed subcommand list
    if o="$(D $x --filter "label=$1")"; then n=$((n + $(awk 'NF { k++ } END { print k + 0 }' <<<"$o"))); else bad=1; fi
  done
  if (( bad )); then printf '%s' -1; else printf '%s' "$n"; fi
}
# nested_cancel MODE: a nested run of this script (--only nested-hold; its own scratch root under this
# case's TMPDIR; everything it creates also carries sdlc.ci.outer=<run>-<control>) is cancelled while its
# validate.sh holds TERM-resistant processes that keep creating containers. MODE: group-term (TERM to the
# nested runner's process group, as validate.sh's supervisor sends it), pid-hup (HUP to its PID only) or
# deadline (--launch-deadline 8). Decoys that must survive: a process whose command line names the nested
# harness path, a "sleep 300" like the stub's, and a container of this control.
# G064 correction 11: the nested runner starts with HUP and TERM at their default disposition (env
# --default-signal, GNU coreutils 8.31+: the CI context image and the local hosts). A caller that ignores HUP
# (nohup) would otherwise pass SIG_IGN down, bash cannot trap a signal ignored on entry, and the HUP of pid-hup
# would be a no-op until the 240 s deadline. Without that env option the launch proceeds only when this runner
# ignores neither signal; otherwise, and when the nested runner does not catch the signal its mode sends, the
# control fails at once with the diagnosis (no signal, no 240 s wait; the runner's cleanup stops the tree).
nested_cancel() {
  local mode="$1" ol="sdlc.ci.outer=$RT-$CTL_ID" e i args=() npid nroot="" snap p s left="" n0 n1 n2 d1 d2 s1 s2 dctr want b
  local sigenv=() ign own
  new_case; ctl_labels
  if env --default-signal=HUP,TERM true 2>/dev/null; then sigenv=("--default-signal=HUP,TERM"); else
    ign="$(awk '/^SigIgn:/ {print $2}' "/proc/$BASHPID/status" 2>/dev/null)" || ign=""
    if [[ ! "$ign" =~ ^[0-9a-f]{16}$ ]] || (( (16#$ign & 16#4001) != 0 )); then
      prob "env has no --default-signal (GNU coreutils 8.31+) and this runner has SIGHUP or SIGTERM ignored (SigIgn ${ign:-unreadable}; e.g. started under nohup): a nested runner would inherit it and could not be cancelled; launch the harness without nohup"
      return 0
    fi
  fi
  args=(--only nested-hold --log "$C/nested.log" --extra-label "$ol")
  for e in "${EXTRA_LABELS[@]}"; do args+=(--extra-label "$e"); done
  if [[ "$mode" == deadline ]]; then args+=(--launch-deadline 8); fi
  # The inventory exists before the nested run: its holder records itself and each container there at once.
  # CI_CONTROLS_OWNER (G064 correction 11 item 4): the nested run's resources carry this runner's owner key, the
  # controlling owner that removes them (see "startup guard"); the nested run checks that it names its live ancestor.
  own="$(owned_file)"
  launch "$C/nested.out" env -i "${sigenv[@]}" -C "$C" PATH="$PATH" HOME="${HOME:-/nonexistent}" LC_ALL=C GIT_CONFIG_GLOBAL=/dev/null \
    GIT_CONFIG_NOSYSTEM=1 "${DOCKER_ENV[@]}" TMPDIR="$C/tmp" CI_CONTROLS_OUTER_OWNED="$own" CI_CONTROLS_OWNER="$OWNER" CI_CONTROLS_OWNED_REG="$REG_MIRROR" \
    bash "$REPO_R/infra/ci/validate-controls.sh" "${args[@]}"
  npid="$LPID"
  owned_proc "$npid" "nested runner"
  for ((i = 0; i < 1200; i++)); do
    if [[ -z "$nroot" ]]; then nroot="$(compgen -G "$C/tmp/ci-controls-*" | head -1)" || nroot=""; fi
    if [[ -n "$nroot" && -e "$nroot/cases/nested-hold/stub/holding" ]]; then break; fi
    if kill -0 "$npid" 2>/dev/null; then sleep 0.1; else break; fi
  done
  if [[ -z "$nroot" || ! -e "$nroot/cases/nested-hold/stub/holding" ]]; then
    prob "the nested fixture never held (nested.out: $(tail -3 "$C/nested.out" | tr '\n' ' '))"; return 0
  fi
  # The nested tree as it runs (PID and start time), and proof that it holds what the control is about.
  e="$npid:$(proc_start "$npid")"
  own_scan "$e" "$e"
  snap="$(awk '$1 == "p" {print $2, $3}' <<<"$OWN_SCAN")"
  for e in validate.pid holder.pid; do
    p="$(cat "$nroot/cases/nested-hold/stub/$e" 2>/dev/null)" || p=""
    if [[ -n "$p" ]] && grep -q "^$p " <<<"$snap"; then :; else prob "the nested tree lacks the process of $e"; fi
  done
  read -r p _ <"$nroot/cases/nested-hold/stub/harness.pids" || p=""
  if [[ -n "$p" ]] && grep -q "^$p " <<<"$snap"; then :; else prob "the nested tree lacks the stub harness"; fi
  # The signal this mode sends reaches the nested runner's trap (its handler is installed: SigCgt bit set);
  # otherwise the control fails now instead of signalling a runner that cannot act on it.
  case "$mode" in group-term) b=14 ;; pid-hup) b=0 ;; *) b="" ;; esac
  if [[ -n "$b" ]]; then
    s="$(awk '/^SigCgt:/ {print $2}' "/proc/$npid/status" 2>/dev/null)" || s=""
    if [[ "$s" =~ ^[0-9a-f]{16}$ ]] && (( (16#$s >> b) & 1 )); then :; else
      ign="$(awk '/^SigIgn:/ {print $2}' "/proc/$npid/status" 2>/dev/null)" || ign=""
      prob "the nested runner does not catch SIG$([[ "$mode" == group-term ]] && echo TERM || echo HUP) (SigCgt ${s:-unreadable}, SigIgn ${ign:-unreadable}): started with it ignored? launch the harness without nohup; not signalled, no 240 s wait"
      return 0
    fi
  fi
  n0="$(lcount "$ol")"
  if (( n0 < 0 )); then prob "the Docker listing of $ol failed; the count is unverified"
  elif (( n0 < 1 )); then prob "the nested holder created no container"; fi
  # Unrelated decoys of this control.
  launch "$C/decoy1.out" bash -c 'exec -a "$1" sleep 600' decoy "bash $nroot/cases/nested-hold/repo/infra/validation/run.sh --report-dir"
  d1="$LPID"
  launch "$C/decoy2.out" sleep 300
  d2="$LPID"
  dctr="$(D create "${LB[@]}" --network none "$CI_BUSYBOX_IMAGE" true)"
  owned_add container "$dctr" "decoy container"
  sleep 0.5
  s1="$(proc_start "$d1")" || s1=""; s2="$(proc_start "$d2")" || s2=""
  case "$mode" in
    group-term) kill -TERM -- "-$npid"; want=143 ;;
    pid-hup) kill -HUP "$npid"; want=129 ;;
    deadline) want=1 ;;
  esac
  await "$npid" 240
  if [[ "$RC" != "$want" ]]; then prob "nested exit $RC, want $want"; fi
  # Every process of the nested tree is gone (same PID and start time).
  while read -r p s; do if [[ -n "$p" ]] && proc_is "$p" "$s"; then left+=" $p:$s"; fi; done <<<"$snap"
  if [[ -n "$left" ]]; then prob "nested processes survived:$left"; stop_tree "$left" "" || true; fi
  # Resources removed, the scratch removed; nothing reappears afterwards.
  n1="$(lcount "$ol")"
  if [[ -e "$nroot" ]]; then prob "the nested scratch root remains"; fi
  sleep 6
  n2="$(lcount "$ol")"
  if (( n1 < 0 )); then prob "the Docker listing of $ol failed; the count is unverified"
  elif (( n1 != 0 )); then prob "$n1 nested control resources remained after the nested run"; fi
  if (( n2 < 0 )); then prob "the Docker listing of $ol failed; the count is unverified"
  elif (( n2 != 0 )); then prob "$n2 nested control resources exist 6 s later (later creation)"; fi
  if [[ -e "$nroot" ]]; then prob "the nested scratch root reappeared"; fi
  # No unsafe publication: no PASS line, the cancelled control named, no publication directory.
  if grep -q '^PASS ' "$C/nested.log" "$C/nested.out" 2>/dev/null; then prob "the nested run published a PASS line"; fi
  case "$mode" in
    group-term|pid-hup)
      e="SIG$([[ "$mode" == group-term ]] && echo TERM || echo HUP)"
      if grep -q "^FAIL nested-hold cancelled by $e; " "$C/nested.log"; then :; else prob "the nested log does not name the cancelled control"; fi
      if grep -q "^ci-controls: cancelled by $e; 0 passed, 0 failed before the cancellation" "$C/nested.log"; then :; else prob "the nested summary does not record the cancellation"; fi ;;
    deadline)
      if grep -q '^FAIL nested-hold deadline: a nested launch was still running after 8 s' "$C/nested.log"; then :; else prob "the nested log does not record the deadline"; fi
      if grep -q '^ci-controls: 0 passed, 1 failed ' "$C/nested.log"; then :; else prob "the nested summary does not record the failure"; fi ;;
  esac
  if [[ -n "$(find "$C" -name 'ci-evidence*' -print -quit)" ]]; then prob "a publication directory exists"; fi
  # The unrelated decoys survived; this control then stops its own decoys.
  if [[ -n "$s1" ]] && proc_is "$d1" "$s1"; then :; else prob "the pattern-named decoy process was stopped"; fi
  if [[ -n "$s2" ]] && proc_is "$d2" "$s2"; then :; else prob "the sleep 300 decoy process was stopped"; fi
  if exists container "$dctr"; then :; else prob "the decoy container was removed"; fi
  if stop_tree "" "$d1:$s1 $d2:$s2"; then :; else prob "the decoys survived KILL: $STOP_LEFT"; fi
  # Whatever the nested run left (counted and reported above), on this or any earlier return, is removed by the
  # runner after the control (owned_cleanup: the recorded IDs and label $ol), after every assertion here.
  detail "nested exit $RC; $(grep -c . <<<"$snap") processes of the nested tree (validate.sh, the TERM-ignoring harness and its container-creating holder) stopped and reaped; $n0 nested containers before, $n1 after, $n2 6 s later; nested scratch removed; no PASS line or publication; decoys survived"
}
# cleanup_driver (G064 correction 11): V18n also proves the runner's per-control cleanup on every exit path with no
# Docker daemon. Nested runs of this script select the fixture nested-hold in a CI_CONTROLS_HOLD_SIM mode against a
# pure-bash docker stand-in (containers as label files; every call logged): startup failure, deadline, cancellation,
# late recreation, a refused removal (non-zero, its ID named) and the normal path; then held cleanup (correction 11
# review): a refused removal and an unreadable presence query each keep the inventory and the scratch root, name the
# inventory and the exact ID and start no later allocating control (zero creations by it), and a refused removal on
# the EXIT path of a cancelled run keeps them too, with exit 143 unchanged. G064 correction 11 item 4: then the
# startup guard through the same real start, each runner against a seeded stand-in daemon (WORK/guard.txt). Stack-free:
# extract this heredoc and run it as `bash cleanup-drv.sh REPO WORK` with REPO a checkout (WORK/cleanup.txt and
# WORK/guard.txt: one ok|bad verdict per case).
cleanup_driver() {
  cat >"$C/cleanup-drv.sh" <<'CLDRV'
#!/usr/bin/env bash
# Cleanup driver of infra/ci/validate-controls.sh (V18n): cleanup-drv.sh REPO WORK. Needs no Docker daemon.
set -uo pipefail
REPO="$1" WORK="$2"
mkdir -p "$WORK/bin" "$WORK/state"
: >"$WORK/cleanup.txt"
cat >"$WORK/bin/docker" <<'FAKE'
#!/usr/bin/env bash
# docker stand-in: a container is the file $FAKE_STATE/<id> holding its labels; volumes, networks and images do not exist.
s="${FAKE_STATE:?}"
mkdir -p "$s"; printf '%s\n' "$*" >>"$s.log"
sub="${1:-}"; if (( $# )); then shift; fi
case "$sub" in
  create|run)
    # FAKE_CREATE_DELAY: seconds before the object exists (a slow daemon; the V18n race case); --cidfile gets the ID.
    if [[ -n "${FAKE_CREATE_DELAY:-}" ]]; then sleep "$FAKE_CREATE_DELAY"; fi
    id="$(head -c 32 /dev/urandom | od -An -tx1 | tr -d ' \n')"; : >"$s/$id"
    while (( $# > 0 )); do
      if [[ "$1" == --label ]]; then printf '%s\n' "$2" >>"$s/$id"; shift; elif [[ "$1" == --cidfile ]]; then printf '%s' "$id" >"$2"; shift; fi
      shift
    done
    if [[ "${FAKE_REFUSE:-}" == first && ! -e "$s.refuse" ]]; then printf '%s\n' "$id" >"$s.refuse"; fi
    printf 'created %s\n' "$id" >>"$s.log"; printf '%s\n' "$id" ;;
  ps)
    lab=""; while (( $# > 0 )); do if [[ "$1" == --filter ]]; then lab="${2#label=}"; shift; fi; shift; done
    # FAKE_PS_FAIL=<prefix>: a listing whose label filter starts with it fails (an unreadable presence query).
    if [[ -n "${FAKE_PS_FAIL:-}" && "$lab" == "$FAKE_PS_FAIL"* ]]; then echo "ps refused: $lab" >&2; exit 1; fi
    for f in "$s"/*; do
      if [[ ! -f "$f" ]]; then continue; fi
      if [[ -z "$lab" ]] || grep -qxF -- "$lab" "$f" || { [[ "$lab" != *=* ]] && grep -q "^$lab=" "$f"; }; then printf '%s\n' "${f##*/}"; fi
    done ;;
  rm)
    rc=0
    for a in "$@"; do
      if [[ "$a" == -* ]]; then continue; fi
      if [[ -e "$s.refuse" ]] && grep -qxF -- "$a" "$s.refuse"; then echo "refused: $a" >&2; rc=1
      elif [[ -f "$s/$a" ]]; then rm -f -- "$s/$a"; printf 'removed %s\n' "$a" >>"$s.log"
      else echo "No such container: $a" >&2; rc=1; fi
    done
    exit "$rc" ;;
  inspect)
    # --format naming sdlc.ci.owner prints that label's value ("" when absent), as Docker's index does.
    # FAKE_INSPECT_FAIL=<ID>: that object stays listed but cannot be inspected (the guard's unreadable case).
    fmt="" ids=()
    while (( $# > 0 )); do case "$1" in --type) shift ;; --format|-f) fmt="${2:-}"; shift ;; -*) ;; *) ids+=("$1") ;; esac; shift; done
    for a in "${ids[@]}"; do
      [[ -f "$s/$a" ]] || exit 1
      if [[ -n "${FAKE_INSPECT_FAIL:-}" && "$a" == "$FAKE_INSPECT_FAIL" ]]; then echo "inspect refused: $a" >&2; exit 1; fi
      if [[ "$fmt" == *sdlc.ci.owner* ]]; then printf '%s\n' "$(sed -n 's/^sdlc\.ci\.owner=//p' "$s/$a" | head -1)"; fi
    done ;;
esac
exit 0
FAKE
chmod 755 "$WORK/bin/docker"
declare -A P=() RC=()
start() { # NAME MODE [VAR=VALUE...] [-- ARGS...]: a runner of the fixture in simulation MODE, its own session
  # START_NOHUP=1: launched as an outer harness under nohup launches it (SIGHUP ignored) through nested_cancel's
  # `env --default-signal=HUP,TERM`, which restores the default disposition the runner's HUP trap needs.
  local n="$1" m="$2" envs=() args=() pre=() eo=()
  shift 2
  while (( $# > 0 )) && [[ "$1" != -- ]]; do envs+=("$1"); shift; done
  if (( $# > 0 )); then shift; args=("$@"); fi
  if [[ -n "${START_NOHUP:-}" ]]; then pre=(nohup); eo=(--default-signal=HUP,TERM); fi
  mkdir -p "$WORK/$n/tmp"
  "${pre[@]}" env -i "${eo[@]}" PATH="$WORK/bin:/usr/local/bin:/usr/bin:/bin" HOME="$WORK/$n" LC_ALL=C TMPDIR="$WORK/$n/tmp" FAKE_STATE="$WORK/state/$n" \
    DOCKER_HOST=unix:///nonexistent/docker.sock \
    CI_CONTROLS_HOLD_SIM="$m" "${envs[@]}" setsid bash "$REPO/infra/ci/validate-controls.sh" --only nested-hold --log "$WORK/$n/log" "${args[@]}" \
    </dev/null >"$WORK/$n/out" 2>&1 &
  P[$n]=$!
}
finish() { # NAME SECONDS: its exit status in RC[NAME] (999: still running at the bound, then KILLed)
  local i
  for ((i = 0; i < $2 * 10; i++)); do if kill -0 "${P[$1]}" 2>/dev/null; then sleep 0.1; else break; fi; done
  if kill -0 "${P[$1]}" 2>/dev/null; then kill -KILL -- "-${P[$1]}" 2>/dev/null; wait "${P[$1]}" 2>/dev/null; RC[$1]=999
  else wait "${P[$1]}"; RC[$1]=$?; fi
}
verdict() { # NAME WANT_RC PATTERN LEFT: exit, the runner's log matches PATTERN, LEFT containers remain (1 s later too)
  local n="$1" why="" c r l
  if [[ "${RC[$n]}" != "$2" ]]; then why+="exit ${RC[$n]}, want $2; "; fi
  if grep -qE -- "$3" "$WORK/$n/log" 2>/dev/null; then :; else why+="the log lacks [$3]; "; fi
  sleep 1
  l="$(find "$WORK/state/$n" -maxdepth 1 -type f 2>/dev/null | grep -c .)"
  if [[ "$l" != "$4" ]]; then why+="$l containers remain, want $4; "; fi
  c="$(grep -c '^created ' "$WORK/state/$n.log" 2>/dev/null)"; r="$(grep -c '^removed ' "$WORK/state/$n.log" 2>/dev/null)"
  if [[ -z "$why" ]]; then printf 'ok %s: exit %s; %s created, %s removed, %s left\n' "$n" "${RC[$n]}" "$c" "$r" "$l"
  else printf 'bad %s: %s(%s created, %s removed; log: %s)\n' "$n" "$why" "$c" "$r" "$(grep -E '^(FAIL|PASS)' "$WORK/$n/log" 2>/dev/null | tr '\n' ' ' | cut -c1-600)"; fi >>"$WORK/cleanup.txt"
}
start startup startup
start deadline hold -- --launch-deadline 2
start interrupt hold
start late late
start refuse left FAKE_REFUSE=first
start done done
# Held cleanup: a refused removal and an unreadable presence query, each followed by a second allocating control
# (obj-codes as CI_CONTROLS_HOLD_SIM_NEXT), and a refused removal on the EXIT path of a cancelled run.
start held-rm left FAKE_REFUSE=first CI_CONTROLS_HOLD_SIM_NEXT=done -- --only nested-hold,obj-codes
start held-query left FAKE_PS_FAIL=sdlc.ci.outer= CI_CONTROLS_HOLD_SIM_NEXT=done -- --only nested-hold,obj-codes
start held-exit hold FAKE_REFUSE=first
# G064 corr1 C5 / corr2 D4: enforced ceilings through the shim (CI_CONTROLS_HOLD_SIM_GATE=1: the creator is gated as the
# real nested holder). cap: the lowered ceiling 2 refuses the third creation (control FAIL naming it), nothing more is
# created, 0 left. held: the same with a refused removal: held, the later control not started. race: two concurrent
# shim creators at ceiling - 1 (slow daemon): exactly one creates. nested-hold: a latch set in a nested case shell; the
# parent shim refuses and creates nothing.
start cap hold CI_CONTROLS_HOLD_SIM_GATE=1 CI_CONTROLS_CEILING_LOWER=nested-hold:2 -- --launch-deadline 4
start held hold CI_CONTROLS_HOLD_SIM_GATE=1 CI_CONTROLS_CEILING_LOWER=nested-hold:2 FAKE_REFUSE=first CI_CONTROLS_HOLD_SIM_NEXT=done -- --launch-deadline 4 --only nested-hold,obj-codes
mkdir -p -m 700 "$WORK/race-reg" "$WORK/race-reg/cid"; printf 'header\njob\tlocal-0123456789ab\n' >"$WORK/race-reg/owned.tsv"
start race race CI_CONTROLS_CEILING_LOWER=nested-hold:3 FAKE_CREATE_DELAY=0.4 CI_CONTROLS_OWNED_REG="$WORK/race-reg/owned.tsv"
start nested-hold latch
# unrec: an object with the control's own label and no inventory row is an `unrecorded` FAIL naming its ID; removed.
start unrec unrec
# G064 r13 (real-Docker V18l-n): the runner under test is itself nested, as nested_cancel launches it (an outer inventory
# CI_CONTROLS_OUTER_OWNED, --extra-label sdlc.ci.outer=<outer run>-<control>), and its holder is shaped like validate.sh's stub
# holder (CI_CONTROLS_HOLD_SIM_GATE=nested): shim creations carrying this run's own sdlc.ci.control label and the outer label,
# recorded in the OUTER inventory only. outer-scope: cancelled (TERM), a clean cancelled run: no unrecorded FAIL, scratch
# removed, nothing left. outer-unrec: an object with both labels and no row in either inventory is still unrecorded.
# outer-hup (V18m under nohup): the outer-scope shape launched under nohup through env --default-signal=HUP,TERM (as
# nested_cancel launches it) and cancelled by SIGHUP: exit 129, the same clean cancelled run.
OL="sdlc.ci.outer=0123456789ab-V18m"
for n in outer-scope outer-unrec outer-hup; do mkdir -p "$WORK/$n"; : >"$WORK/$n/outer.tsv"; done
start outer-scope hold CI_CONTROLS_HOLD_SIM_GATE=nested CI_CONTROLS_OUTER_OWNED="$WORK/outer-scope/outer.tsv" -- --extra-label "$OL"
start outer-unrec unrec CI_CONTROLS_HOLD_SIM_GATE=nested CI_CONTROLS_OUTER_OWNED="$WORK/outer-unrec/outer.tsv" -- --extra-label "$OL"
START_NOHUP=1 start outer-hup hold CI_CONTROLS_HOLD_SIM_GATE=nested CI_CONTROLS_OUTER_OWNED="$WORK/outer-hup/outer.tsv" -- --extra-label "$OL"
for n in interrupt held-exit outer-scope outer-hup; do
  for ((i = 0; i < 600; i++)); do if compgen -G "$WORK/$n/tmp/ci-controls-*/cases/nested-hold/holding" >/dev/null; then break; fi; sleep 0.1; done
  if [[ "$n" == outer-hup ]]; then kill -HUP "${P[$n]}" 2>/dev/null; else kill -TERM "${P[$n]}" 2>/dev/null; fi
done
for n in startup deadline interrupt late refuse done held-rm held-query held-exit cap held race nested-hold unrec outer-scope outer-unrec outer-hup; do finish "$n" 150; done
rid="$(cat "$WORK/state/refuse.refuse" 2>/dev/null)" || rid=""
verdict startup 1 '^FAIL nested-hold the holding harness never started .*;cleanup after the control removed [0-9]+ ' 0
verdict deadline 1 '^FAIL nested-hold deadline: a nested launch was still running after 2 s.*;cleanup after the control removed [0-9]+ ' 0
verdict interrupt 143 '^ci-controls: cancelled by SIGTERM; 0 passed, 0 failed' 0
verdict late 1 '^FAIL nested-hold cleanup: late recreation after the sweep: container [0-9a-f]{64};' 0
verdict refuse 1 "^FAIL nested-hold cleanup: owned containers survived the cleanup: ${rid:-unknown} ;" 1
verdict done 0 '^PASS nested-hold simulated done' 0
hverdict() { # NAME WANT_RC PATTERN ID LEFT [next]: verdict's checks (exit, PATTERN, LEFT containers remain; rec: as many as
  # the inventory records) plus the held record: the scratch root and its inventory kept, the inventory recording
  # container ID, a log line naming that inventory and ID; no create call made by the later control obj-codes. next:
  # its not-started line, no PASS or FAIL line of it and the held control's receipt (problems file) kept.
  local n="$1" why="" root inv l c r o w="$5"
  if [[ "${RC[$n]}" != "$2" ]]; then why+="exit ${RC[$n]}, want $2; "; fi
  if grep -qE -- "$3" "$WORK/$n/log" 2>/dev/null; then :; else why+="the log lacks [$3]; "; fi
  sleep 1
  root="$(compgen -G "$WORK/$n/tmp/ci-controls-*" | head -1)" || root=""
  inv="$root/owned/nested-hold.tsv"
  if [[ -z "$root" || ! -d "$root" ]]; then why+="the scratch root was removed; "
  elif [[ ! -f "$inv" ]]; then why+="the inventory was deleted; "
  elif [[ -z "$4" ]] || ! grep -qF -- "$(printf 'container\t%s\t' "$4")" "$inv"; then why+="the inventory lacks container ${4:-(none)}; "; fi
  if [[ -n "$root" && -n "$4" ]] && grep -F -- "$inv" "$WORK/$n/log" 2>/dev/null | grep -qF -- "container $4"; then :; else why+="no log line names the kept inventory and container ${4:-(none)}; "; fi
  l="$(find "$WORK/state/$n" -maxdepth 1 -type f 2>/dev/null | grep -c .)"
  if [[ "$w" == rec ]]; then w="$(grep -c '^container' "$inv" 2>/dev/null)" w="${w:-0}"; fi
  if [[ "$l" != "$w" || "$l" == 0 ]]; then why+="$l containers remain, want $w; "; fi
  o="$(grep -cE '^create .*--label sdlc\.ci\.outer=[0-9a-f]+-obj-codes( |$)' "$WORK/state/$n.log" 2>/dev/null)"
  if [[ "${o:-0}" != 0 ]]; then why+="${o} create calls by the later control obj-codes; "; fi
  if [[ "${6:-}" == next ]]; then
    if grep -qx 'FAIL not-started 1 later control(s) were not started after the unverified cleanup of nested-hold (not passed): obj-codes' "$WORK/$n/log"; then :
    else why+="no not-started line for obj-codes; "; fi
    if grep -qE '^(PASS|FAIL) obj-codes ' "$WORK/$n/log"; then why+="obj-codes ran; "; fi
    if [[ -s "$root/problems/nested-hold" ]]; then :; else why+="the receipt problems/nested-hold was not kept; "; fi
  fi
  c="$(grep -c '^created ' "$WORK/state/$n.log" 2>/dev/null)"; r="$(grep -c '^removed ' "$WORK/state/$n.log" 2>/dev/null)"
  if [[ -z "$why" ]]; then printf 'ok %s: exit %s; %s created, %s removed, %s left; inventory and scratch kept naming %s; %s later creations\n' "$n" "${RC[$n]}" "$c" "$r" "$l" "${4:0:12}" "${o:-0}"
  else printf 'bad %s: %s(%s created, %s removed; log: %s)\n' "$n" "$why" "$c" "$r" "$(grep -E '^(FAIL|PASS)' "$WORK/$n/log" 2>/dev/null | tr '\n' ' ' | cut -c1-600)"; fi >>"$WORK/cleanup.txt"
}
hid="$(cat "$WORK/state/held-rm.refuse" 2>/dev/null)" || hid=""
hverdict held-rm 1 "^FAIL nested-hold cleanup: owned containers survived the cleanup: ${hid:-unknown} ;.*;cleanup unverified, held: inventory /.*/owned/nested-hold\\.tsv \\(recorded IDs: [^)]*container ${hid:-unknown}" "$hid" 1 next
qid="$(awk '$1 == "created" {print $2; exit}' "$WORK/state/held-query.log" 2>/dev/null)" || qid=""
hverdict held-query 1 "^FAIL nested-hold cleanup: Docker could not list the containers; their removal is unverified;cleanup unverified, held: inventory /.*/owned/nested-hold\\.tsv \\(recorded IDs: [^)]*container ${qid:-unknown}" "$qid" rec next
xid="$(cat "$WORK/state/held-exit.refuse" 2>/dev/null)" || xid=""
hverdict held-exit 143 "^FAIL cleanup nested-hold: owned containers survived the cleanup: ${xid:-unknown} ; inventory kept as the recovery record: /.*/owned/nested-hold\\.tsv \\(recorded IDs: [^)]*container ${xid:-unknown}" "$xid" 1
cverdict() { # NAME WANT_CREATED: exactly that many create calls reached the stand-in daemon (a refusal creates nothing)
  local c
  c="$(grep -c '^created ' "$WORK/state/$1.log" 2>/dev/null)" c="${c:-0}"
  if [[ "$c" != "$2" ]]; then printf 'bad %s-creations: %s objects created, want exactly %s\n' "$1" "$c" "$2" >>"$WORK/cleanup.txt"; fi
}
verdict cap 1 '^FAIL nested-hold .*resource ceiling 2/2 for nested-hold \(container\): creation refused;.*held: resource ceiling latch of nested-hold set \(container ceiling 2/2\).*cleanup after the control removed 2 ' 0
cverdict cap 2
cid="$(cat "$WORK/state/held.refuse" 2>/dev/null)" || cid=""
hverdict held 1 "^FAIL nested-hold .*resource ceiling 2/2 for nested-hold \\(container\\): creation refused;.*cleanup: owned containers survived the cleanup: ${cid:-unknown} ;.*cleanup unverified, held: inventory /.*/owned/nested-hold\\.tsv \\(recorded IDs: [^)]*container ${cid:-unknown}" "$cid" 1 next
cverdict held 2
verdict race 1 '^FAIL nested-hold resource ceiling 3/3 for nested-hold \(container\): creation refused' 0
if grep -q '^FAIL nested-hold.*race: ' "$WORK/race/log" 2>/dev/null; then printf 'bad race-outcome: %s\n' "$(grep '^FAIL nested-hold' "$WORK/race/log" | cut -c1-300)" >>"$WORK/cleanup.txt"; fi
cverdict race 3
verdict nested-hold 1 '^FAIL nested-hold held: resource ceiling latch of nested-hold set \(nested case shell breach\)' 0
cverdict nested-hold 0
# the race run's registry mirror: one standalone row per created container (ci-<control>), none twice, none other
rc_ids="$(awk '$1 == "created" {print $2}' "$WORK/state/race.log" | LC_ALL=C sort)"
rg_ids="$(awk -F'\t' '$1 == "standalone" && $3 == "ci-nested-hold" {print $2}' "$WORK/race-reg/owned.tsv" | LC_ALL=C sort)"
if [[ -z "$rc_ids" || "$rc_ids" != "$rg_ids" || "$(grep -c '^standalone' "$WORK/race-reg/owned.tsv")" != "$(grep -c . <<<"$rc_ids")" ]]; then
  printf 'bad race-mirror: registry rows [%s] for created [%s]\n' "$(tr '\n' ' ' <<<"$rg_ids")" "$(tr '\n' ' ' <<<"$rc_ids")" >>"$WORK/cleanup.txt"
fi
uid="$(awk '$1 == "created" {print $2; exit}' "$WORK/state/unrec.log" 2>/dev/null)" || uid=""
verdict unrec 1 "^FAIL nested-hold cleanup: unrecorded container ${uid:-unknown} carries sdlc\\.ci\\.control=[0-9a-f]{12}-nested-hold;" 0
overdict() { # NAME: (bad lines only) the nested-shape run left no FAIL cleanup line and no scratch root, and every container
  # it created (at least 2) is recorded in its outer inventory
  local n="$1" why="" c x
  if grep -q '^FAIL cleanup' "$WORK/$n/log" 2>/dev/null; then why+="[$(grep '^FAIL cleanup' "$WORK/$n/log" | cut -c1-200)]; "; fi
  if compgen -G "$WORK/$n/tmp/ci-controls-*" >/dev/null; then why+="the scratch root remains; "; fi
  c="$(awk '$1 == "created" {print $2}' "$WORK/state/$n.log" 2>/dev/null)"
  if (( $(grep -c . <<<"$c") < 2 )); then why+="fewer than 2 creations; "; fi
  for x in $c; do grep -qF -- "$(printf 'container\t%s\t' "$x")" "$WORK/$n/outer.tsv" || why+="container ${x:0:12} has no outer row; "; done
  if [[ -n "$why" ]]; then printf 'bad %s-outer: %s\n' "$n" "$why" >>"$WORK/cleanup.txt"; fi
}
verdict outer-scope 143 '^ci-controls: cancelled by SIGTERM; 0 passed, 0 failed' 0
overdict outer-scope
verdict outer-hup 129 '^ci-controls: cancelled by SIGHUP; 0 passed, 0 failed' 0
overdict outer-hup
ouid="$(awk '$1 == "created" {print $2; exit}' "$WORK/state/outer-unrec.log" 2>/dev/null)" || ouid=""
verdict outer-unrec 1 "^FAIL nested-hold cleanup: unrecorded container ${ouid:-unknown} carries sdlc\\.ci\\.control=[0-9a-f]{12}-nested-hold \\(outer scope: no row in /.*/outer\\.tsv either\\);" 0
# G064 correction 11 item 4: the startup guard through the runner's real start (WORK/guard.txt: one verdict per case).
# Each guard runner has its own stand-in daemon, seeded before it starts; besides its subject every case holds two
# decoys, a container of a live other run of this checkout and an unlabelled one. Owner keys come from the real lib.sh
# (g003_owner_key) for live helper processes of this driver, which is also the runners' parent. A refusal: exit 16
# naming the reason, no allocation (no create or run call, no scratch root) and no removal call. A pass: exit 0, PASS
# nested-hold. Either way every seeded container is still there and no removal names it.
: >"$WORK/guard.txt"
g3() { bash -c 'set -euo pipefail; source "$1"; shift; "$@"' g3 "$REPO/infra/validation/lib.sh" "$@"; }
hex() { head -c "$1" /dev/urandom | od -An -tx1 | tr -d ' \n'; }
declare -A SEEDED=()
seed() { # NAME OWNER: a container with CI-control labels and sdlc.ci.owner=OWNER in NAME's stand-in daemon
  # (OWNER "-": no owner label; "unlabelled": no g003 label at all); its ID is added to SEEDED[NAME] and kept in LAST
  local id r
  id="$(hex 32)" r="$(hex 6)"
  mkdir -p "$WORK/state/$1"
  case "$2" in
    unlabelled) printf 'com.example.unrelated=%s\n' "$r" ;;
    -) printf 'sdlc.ci.control=%s-seeded\nsdlc.ci.run=%s\n' "$r" "$r" ;;
    *) printf 'sdlc.ci.control=%s-seeded\nsdlc.ci.run=%s\nsdlc.ci.owner=%s\n' "$r" "$r" "$2" ;;
  esac >"$WORK/state/$1/$id"
  SEEDED[$1]+="$id " LAST="$id"
}
decoys() { seed "$1" "$KLIVE"; seed "$1" unlabelled; }
sleep 900 & LIVEP=$!
sleep 900 & GONEP=$!
KLIVE="$(g3 g003_owner_key "$LIVEP" "$(hex 6)" "$REPO")" || KLIVE=""         # a live other run of this checkout
KFOREIGN="$(g3 g003_owner_key "$LIVEP" "$(hex 6)" /)" || KFOREIGN=""         # a live owner of another checkout
KGONE="$(g3 g003_owner_key "$GONEP" "$(hex 6)" "$REPO")" || KGONE=""
kill "$GONEP" 2>/dev/null; wait "$GONEP" 2>/dev/null
KDRV="$(g3 g003_owner_key "$$" "$(hex 6)" "$REPO")" || KDRV=""               # this driver: the runners' live parent
IFS=. read -r _ kb kns kp kst kco krun <<<"$KLIVE"
KBOOT="v1.$(tr '0-9a-f' '1-9a-f0' <<<"$kb").$kns.$kp.$kst.$kco.$krun"         # another boot, the live PID and start
KREUSE="v1.$kb.$kns.$kp.$((kst + 1)).$kco.$krun"                             # the live PID, another start time
if [[ -z "$KLIVE" || -z "$KFOREIGN" || -z "$KGONE" || -z "$KDRV" || -e "/proc/$GONEP" ]]; then
  printf 'bad guard-setup: an owner key could not be made or the ended process %s is still present\n' "$GONEP" >>"$WORK/guard.txt"
fi
seed g-live "$KLIVE"; seed g-live "$KFOREIGN"; seed g-live unlabelled
seed g-dead "$KGONE"; decoys g-dead
seed g-boot "$KBOOT"; decoys g-boot
seed g-reuse "$KREUSE"; decoys g-reuse
seed g-malformed "v1.$kb.not-an-owner-key"; decoys g-malformed
seed g-missing -; decoys g-missing
seed g-unreadable "$KLIVE"; unread="$LAST"; decoys g-unreadable
seed g-nested "$KDRV"; decoys g-nested
decoys g-nested-bad
GUARD_CASES="g-live g-dead g-boot g-reuse g-malformed g-missing g-unreadable g-nested g-nested-bad"
for n in g-live g-dead g-boot g-reuse g-malformed g-missing; do start "$n" done; done
start g-unreadable done FAKE_INSPECT_FAIL="$unread"
start g-nested done CI_CONTROLS_OWNER="$KDRV"
start g-nested-bad done CI_CONTROLS_OWNER="$KLIVE"
for n in $GUARD_CASES; do finish "$n" 150; done
gverdict() { # NAME WANT_RC WANT_HELD PATTERN [OWNER]: exit, held-object lines, the log matches PATTERN, the seeded
  # containers kept; 16: the hold and its summary, no allocation, removal or scratch root; 0: PASS nested-hold and,
  # with OWNER, every create call carrying exactly that owner key
  local n="$1" lg="$WORK/state/$1.log" why="" id a r h k=0 c o
  if [[ "${RC[$n]}" != "$2" ]]; then why+="exit ${RC[$n]}, want $2; "; fi
  if grep -qE -- "$4" "$WORK/$n/log" 2>/dev/null; then :; else why+="the log lacks [$4]; "; fi
  h="$(grep -cE '^startup guard: (container|network|volume|image) ' "$WORK/$n/log" 2>/dev/null)"
  if [[ "${h:-0}" != "$3" ]]; then why+="${h:-0} held objects named, want $3; "; fi
  for id in ${SEEDED[$n]}; do
    if [[ -f "$WORK/state/$n/$id" ]]; then k=$((k + 1)); else why+="seeded ${id:0:12} was removed; "; fi
    if grep -qE "^(rm|removed) .*$id" "$lg" 2>/dev/null; then why+="a removal named seeded ${id:0:12}; "; fi
  done
  a="$(grep -cE '^(create|run|commit|build|buildx|compose|network create|volume create|image build) ' "$lg" 2>/dev/null)" a="${a:-0}"
  r="$(grep -cE '^(rm|network rm|volume rm|image rm|rmi|kill|stop|prune|system prune) ' "$lg" 2>/dev/null)" r="${r:-0}"
  if [[ "$2" == 16 ]]; then
    if grep -qE '^FAIL setup startup guard: hold: ' "$WORK/$n/log" && grep -qE '^ci-controls: startup hold \(exit 16\); no control ran' "$WORK/$n/log"; then :
    else why+="no hold line and summary; "; fi
    if [[ "$a" != 0 || "$r" != 0 ]]; then why+="$a allocation and $r removal calls; "; fi
    if compgen -G "$WORK/$n/tmp/ci-controls-*" >/dev/null; then why+="a scratch root was created; "; fi
  else
    if grep -q '^PASS nested-hold simulated done' "$WORK/$n/log" 2>/dev/null; then :; else why+="no PASS line; "; fi
    if [[ -n "${5:-}" ]]; then
      c="$(grep -c '^create ' "$lg" 2>/dev/null)" o="$(grep -c "^create .*--label sdlc\.ci\.owner=$5 " "$lg" 2>/dev/null)"
      if [[ "${c:-0}" == 0 || "$c" != "$o" ]] || grep '^create ' "$lg" | grep -o 'sdlc\.ci\.owner=[^ ]*' | grep -qvxF "sdlc.ci.owner=$5"; then
        why+="${o:-0} of ${c:-0} create calls carry the controlling owner key alone; "
      fi
    fi
  fi
  if [[ -z "$why" ]]; then printf 'ok %s: exit %s; %s held named; %s allocation and %s removal calls; %s seeded kept\n' "$n" "${RC[$n]}" "${h:-0}" "$a" "$r" "$k"
  else printf 'bad %s: %s(log: %s)\n' "$n" "$why" "$(grep -E '^(FAIL|PASS|startup guard)' "$WORK/$n/log" 2>/dev/null | tr '\n' ' ' | cut -c1-600)"; fi >>"$WORK/guard.txt"
}
gverdict g-live 0 0 '^startup guard: clear: .*[(]0 of the controlling run and 2 of other live owners ignored'
gverdict g-dead 16 1 '^startup guard: container [0-9a-f]{12}: owner gone [(]pid-absent; pid [0-9]+, run [0-9a-f]{12}, this checkout[)]$'
gverdict g-boot 16 1 '^startup guard: container [0-9a-f]{12}: owner gone [(]boot; '
gverdict g-reuse 16 1 '^startup guard: container [0-9a-f]{12}: owner gone [(]pid-reuse; '
gverdict g-malformed 16 1 '^startup guard: container [0-9a-f]{12}: owner cannot be attributed [(]malformed[)]$'
gverdict g-missing 16 1 '^startup guard: container [0-9a-f]{12}: owner cannot be attributed [(]missing[)]$'
gverdict g-unreadable 16 1 "^startup guard: container ${unread:0:12}: owner cannot be attributed [(]unreadable[)]\$"
gverdict g-nested 0 0 '^startup guard: clear: .*[(]1 of the controlling run and 1 of other live owners ignored' "$KDRV"
gverdict g-nested-bad 16 0 '^FAIL setup startup guard: hold: CI_CONTROLS_OWNER does not name a live controlling ancestor'
kill "$LIVEP" 2>/dev/null; wait "$LIVEP" 2>/dev/null
if grep -q '^bad ' "$WORK/cleanup.txt" "$WORK/guard.txt"; then exit 1; fi
exit 0
CLDRV
}
ctl_V18l() { nested_cancel group-term; }
ctl_V18m() { nested_cancel pid-hup; }
ctl_V18n() {
  local l n g
  nested_cancel deadline
  cleanup_driver; mkdir -p "$C/cleanup"
  launch "$C/cleanup.out" env -i -C "$C" PATH="$PATH" HOME="${HOME:-/nonexistent}" LC_ALL=C TMPDIR="$C/tmp" \
    bash "$C/cleanup-drv.sh" "$REPO_R" "$C/cleanup"
  await "$LPID" 300
  if [[ "$RC" != 0 ]]; then prob "cleanup driver exit $RC"; fi
  while IFS= read -r l; do prob "cleanup driver: ${l#bad }"; done < <(grep -h '^bad ' "$C/cleanup/cleanup.txt" "$C/cleanup/guard.txt" 2>/dev/null)
  n="$(grep -c '^ok ' "$C/cleanup/cleanup.txt" 2>/dev/null)" || n=0
  if [[ "$n" != 17 ]]; then prob "cleanup driver: $n of 17 cases ok"; fi
  # G064 correction 11 item 4: the startup guard cases of the same driver.
  g="$(grep -c '^ok ' "$C/cleanup/guard.txt" 2>/dev/null)" || g=0
  if [[ "$g" != 9 ]]; then prob "cleanup driver: $g of 9 startup guard cases ok"; fi
  detail "$(cat "$DET"); runner cleanup without a daemon ($n/17): startup failure (TERM-ignoring creator KILLed first), deadline and SIGTERM 143 removed every recorded container, late recreation and a refused removal fail naming the ID, normal path passes, a refused removal and an unreadable query keep the inventory and scratch and start no later control, a refused removal at exit keeps them with exit 143, the lowered ceiling refuses through the shim (control FAIL, no later creation, 0 left; with a refused removal held), two concurrent creators at ceiling - 1 create exactly one, a latch set in a nested case shell refuses the parent shim, created containers are mirrored once each into the run registry, an object with the control label and no row is an unrecorded FAIL naming its ID, a nested runner removes its outer-scope holder objects (its own control label, recorded only in the outer inventory) as a clean cancelled run, also when launched under nohup and cancelled by SIGHUP (129), and a doubly-labelled object with no row in either inventory is still unrecorded; startup guard ($g/9): live owners of this and another checkout and the controlling parent pass untouched, a dead, other-boot, reused-PID, malformed, missing or unreadable owner and a non-ancestor controlling key hold (exit 16) before any allocation, decoys kept"
}

# --- group D: the supplemental JUnit report (G050; J1-J6) --------------------------------------------------------
# validate.sh's junit and junit_withheld phases (infra/ci/junit.sh): check cases come only from the sanitized
# status columns, phase cases only from failed numeric phases; a converter error is 19 with no junit.xml; the
# job status keeps its precedence. Expected cases are derived from the published results.tsv or the case tree's
# required-checks.tsv, never from junit.xml itself.
JX() { printf '%s' "$(PUBDIR)/junit.xml"; }
jcases() { # FILE -> "CLASS NAME TYPE" per testcase, in document order (TYPE pass when it has no failure child)
  sed -E 's/ time="[0-9]{1,5}"//g; s/<testcase /\n<testcase /g' "$1" | sed -nE \
    -e 's/^<testcase classname="([^"]*)" name="([^"]*)"><failure type="([^"]*)".*/\1 \2 \3/p' \
    -e 's/^<testcase classname="([^"]*)" name="([^"]*)"\/>.*/\1 \2 pass/p'
}
jsuite() { # FILE -> "NAME TESTS FAILURES" of the one testsuite element
  sed -nE 's/.*<testsuite name="([^"]*)" tests="([0-9]+)" failures="([0-9]+)" errors="0" skipped="0">.*/\1 \2 \3/p' "$1"
}
jexpect() { # JOB PHASE...: the well-formed non-INFO rows of the published results.tsv, then a failure per PHASE
  local j="$1" p
  shift
  if [[ -f "$(PUBDIR)/results.tsv" ]]; then
    awk -F'\t' -v j="$j" 'NF == 3 && $1 ~ /^[A-Z][A-Z0-9-]*$/ && $2 ~ /^(PASS|FAIL|UNMET)$/ {
      print "validate." j ".check " $1 " " ($2 == "PASS" ? "pass" : $2) }' "$(PUBDIR)/results.tsv"
  fi
  for p in "$@"; do printf 'validate.%s.phase phase:%s phase-error\n' "$j" "$p"; done
}
jreq() { # JOB PROFILE: a passing case per required ID of PROFILE, in required-checks.tsv order (the stub's row order)
  awk -F'\t' -v j="$1" -v s=" $(ci_profile_suites "$2") final " '$1 != "" && $1 !~ /^#/ && index(s, " " $2 " ") {
    print "validate." j ".check " $1 " pass" }' "$C/repo/infra/validation/required-checks.tsv"
}
expect_jx() { # SUITE EXPECTED: the published junit.xml has exactly the EXPECTED cases ("CLASS NAME TYPE" lines)
  local f t=0 fl=0
  f="$(JX)"
  if [[ -n "$2" ]]; then t="$(grep -c . <<<"$2")" || t=0; fl="$(grep -vc ' pass$' <<<"$2")" || fl=0; fi
  if [[ ! -f "$f" ]]; then prob "junit.xml was not published"; return 0; fi
  if [[ "$(jsuite "$f")" != "$1 $t $fl" ]]; then prob "junit suite [$(jsuite "$f")], want [$1 $t $fl]"; fi
  if [[ "$(jcases "$f")" != "$2" ]]; then prob "junit cases [$(jcases "$f" | tr '\n' ';')], want [$(tr '\n' ';' <<<"$2")]"; fi
}
expect_jmsg() { # TEXT: junit.xml carries the fixed failure message TEXT
  if grep -qF "message=\"$1\"" "$(JX)" 2>/dev/null; then :; else prob "junit.xml lacks the message [$1]"; fi
}
expect_no_jx() { if [[ -e "$(JX)" ]]; then prob "junit.xml was published"; fi; }
# J1 (AC1): a passing fast run: an independent yq parse (no network, no external resolution) agrees with the
# expected cases and counts; the suite is JOB:PROFILE; a second identical run gives identical bytes.
ctl_J1() {
  local exp n got h1
  std_ci 4300; run_v "${CIENV[@]}" -- --profile fast --publish-dir ci-evidence
  expect_rc 0; expect_pub full; expect_phase junit 0; expect_runs 1
  exp="$(jexpect fast)"; n="$(grep -c . <<<"$exp")" || n=0
  if (( n < 1 )); then prob "no expected check case"; fi
  expect_jx fast:fast "$exp"
  if grep -q ' validate[.]fast[.]phase ' <<<" $(jcases "$(JX)" | tr '\n' ' ')"; then prob "a passing run has a phase case"; fi
  ctl_labels
  got="$(D run --rm --pull never --network none "${LB[@]}" -i "$CI_YQ_IMAGE" -p xml -o tsv '[.testsuites.testsuite."+@name",
    .testsuites."+@tests", .testsuites.testsuite."+@failures", ([.testsuites.testsuite.testcase] | flatten(1) | map(select(tag == "!!map")) | length),
    ([.. | select(tag == "!!map") | keys | .[] | select(. == "+content" or . == "+directive")] | length)]' <"$(JX)" 2>&1)" || got="yq failed: $got"
  if [[ "$got" != "fast:fast"$'\t'"$n"$'\t0\t'"$n"$'\t0' ]]; then prob "independent yq parse [$got], want [fast:fast $n 0 $n 0]"; fi
  if grep -q '<!' "$(JX)"; then prob "junit.xml has a DOCTYPE or another declaration"; fi
  log_has "junit: junit.xml with $n cases, 0 failing (supplementary)"
  h1="$(sha256sum <"$(JX)")" || h1=""
  std_ci 4301; run_v "${CIENV[@]}" -- --profile fast --publish-dir ci-evidence
  expect_rc 0
  if [[ -z "$h1" || "$(sha256sum <"$(JX)" 2>/dev/null)" != "$h1" ]]; then prob "a second identical run gave different junit.xml bytes"; fi
  # 2026-10-19 timing repair: overlapping executed-suite timings give one timing case (self-checked), no suite case.
  std_ci 4302; run_v "${CIENV[@]}" STUB_SUITE_TIMES=1 -- --profile fast --publish-dir ci-evidence
  expect_rc 0; expect_phase junit 0
  expect_jx fast:fast "$exp"$'\n'"validate.fast.timing timing:validation-elapsed pass"
  jtiming
  detail "exit 0; fast:fast, $n passing check cases in results.tsv order, no phase case; yq parse agrees; byte-identical on a second run; overlapping suite timings (220 s summed): one timing:validation-elapsed case, the only time attribute [$JT] s, no suite case"
}
# jtiming: the published junit.xml has exactly one time attribute, on timing:validation-elapsed, and no suite
# case (sets JT). Shape only: the measured value depends on runner speed; J4b proves non-additivity with a fixed
# elapsed row.
jtiming() {
  local f
  f="$(JX)"; JT="$(grep -oE ' time="[0-9]+"' "$f" 2>/dev/null | tr -dc '0-9\n' | paste -sd' ' || true)"
  if [[ ! "$JT" =~ ^[0-9]+$ ]]; then prob "case times [$JT], want exactly one value"; fi
  if ! grep -qE 'classname="validate[.][a-z]+[.]timing" name="timing:validation-elapsed" time="[0-9]+"/>' "$f" 2>/dev/null; then prob "no timing:validation-elapsed case with the time"; fi
  if grep -q 'name="suite:' "$f" 2>/dev/null; then prob "a suite case is reported"; fi
}
# J2 (AC1, AC2): mixed PASS, FAIL, UNMET and INFO rows (B-SCOPE-EXCEPTION twice) whose detail column carries XML-
# sensitive text and a secret-like token: exact cases and fixed messages, INFO absent, no detail in the report or
# the job log.
ctl_J2() {
  local tok exp
  std_ci 4302; tok="J2DETAIL$(ci_random_hex 6)"
  {
    printf 'PREFLIGHT\tINFO\t<script>&"%s\x27\n' "$tok"
    printf 'B-SHELL\tPASS\t<b a="1">&amp; %s\n' "$tok"
    printf 'B-SCOPE-EXCEPTION\tINFO\tfirst %s\n' "$tok"
    printf 'B-SCOPE-EXCEPTION\tINFO\tsecond %s\n' "$tok"
    printf 'B-HADOLINT\tFAIL\tSECRETLIKE-%s <&>\n' "$tok"
    printf 'B-COMPOSE\tUNMET\t"quoted" \x27apos\x27 %s\n' "$tok"
    printf 'IMAGES\tINFO\t%s\n' "$tok"
  } >"$C/replay.tsv"
  run_v "${CIENV[@]}" STUB_HARNESS=replay STUB_REPLAY="$C/replay.tsv" -- --profile fast --publish-dir ci-evidence
  expect_rc 1; expect_line has harness=1; expect_phase complete 14; expect_phase junit 0; expect_pub full
  exp="validate.fast.check B-SHELL pass
validate.fast.check B-HADOLINT FAIL
validate.fast.check B-COMPOSE UNMET
validate.fast.phase phase:harness phase-error
validate.fast.phase phase:complete phase-error"
  expect_jx fast:fast "$exp"
  expect_jmsg "check failed"; expect_jmsg "required prerequisite unmet"; expect_jmsg "phase harness exited 1"; expect_jmsg "phase complete exited 14"
  if grep -qE 'PREFLIGHT|IMAGES|B-SCOPE-EXCEPTION' "$(JX)" 2>/dev/null; then prob "an INFO row reached junit.xml"; fi
  if grep -qF "$tok" "$(JX)" 2>/dev/null; then prob "detail text reached junit.xml"; fi
  if grep -qF "$tok" "$C/out"; then prob "detail text reached the job log"; fi
  detail "exit $RC; 3 check cases (PASS, FAIL check failed, UNMET required prerequisite unmet) plus phase:harness and phase:complete; 4 INFO rows omitted; no detail canary in junit.xml or the job log"
}
# J3 (AC2): an executed profile none (a docs-only merge-request gate) gives a valid empty suite; required checks
# outside the executed profile are never invented; a row the harness does report outside the profile is shown as
# reported, beside the phase:complete failure that completeness records for it.
ctl_J3() {
  local BASE="" ex n=0 id r1 r2
  new_case; vp9_docs_commit
  run_v GITLAB_CI=true CI_PROJECT_DIR="$C/repo" CI_JOB_ID=4303 CI_PIPELINE_ID=1001 CI_DEFAULT_BRANCH=main CI_JOB_TIMEOUT=10800 \
    RUNNER_SCRIPT_TIMEOUT=2h45m CI_JOB_STARTED_AT="$(ci_utc)" CI_PIPELINE_SOURCE=merge_request_event CI_MERGE_REQUEST_IID=7 \
    CI_MERGE_REQUEST_EVENT_TYPE=detached CI_MERGE_REQUEST_SOURCE_BRANCH_NAME=feature CI_MERGE_REQUEST_TARGET_BRANCH_NAME=staging \
    CI_COMMIT_SHA="$(head_sha)" CI_COMMIT_REF_NAME=feature CI_MERGE_REQUEST_DIFF_BASE_SHA="$BASE" -- --profile gate --publish-dir ci-evidence
  r1="$RC"; expect_rc 0; expect_line has harness=not-required; expect_phase complete n/a; expect_phase junit 0
  expect_jx gate:none ""
  std_ci 4304; run_v "${CIENV[@]}" -- --profile fast --publish-dir ci-evidence
  expect_rc 0; expect_phase junit 0
  ex="$(awk -F'\t' -v s=" $(ci_profile_suites fast) final " '$1 != "" && $1 !~ /^#/ && !index(s, " " $2 " ") {print $1}' "$C/repo/infra/validation/required-checks.tsv")"
  for id in $ex; do
    n=$((n + 1))
    if grep -qF "name=\"$id\"" "$(JX)" 2>/dev/null; then prob "$id, outside the fast profile, is in junit.xml"; fi
  done
  if (( n == 0 )); then prob "no required check lies outside the fast profile"; fi
  r2="$RC"
  std_ci 4309; run_v "${CIENV[@]}" STUB_ADD_ROW=B-BROWSER-ROOT -- --profile fast --publish-dir ci-evidence
  expect_rc 14; expect_phase complete 14; expect_phase junit 0
  expect_jx fast:fast "$(jexpect fast complete)"
  if grep -qF 'name="B-BROWSER-ROOT"/>' "$(JX)" 2>/dev/null; then :; else prob "the reported out-of-profile row is not shown as reported"; fi
  detail "docs-only MR gate exit $r1: gate:none with 0 cases; fast run exit $r2: none of the $n required IDs outside the fast profile is invented; a reported B-BROWSER-ROOT row in a fast run: exit $RC, shown as reported with phase:complete"
}
# J4a (AC2): raw-input defects keep the existing completeness precedence (complete=14, overall 14, junit=0); the
# report shows phase:complete and only the well-formed rows; a harness failure keeps its own status.
ctl_J4a() {
  local r=""
  std_ci 4305; run_v "${CIENV[@]}" STUB_HARNESS=malformed -- --profile fast --publish-dir ci-evidence
  expect_rc 14; expect_line has harness=0; expect_phase complete 14; expect_phase junit 0
  expect_jx fast:fast "$(jexpect fast complete)"
  if grep -q 'X-BAD' "$(JX)" 2>/dev/null; then prob "the malformed row reached junit.xml"; fi
  r+=" malformed=$RC"
  std_ci 4306; run_v "${CIENV[@]}" STUB_HARNESS=unknown-status -- --profile fast --publish-dir ci-evidence
  expect_rc 14; expect_line has harness=0; expect_phase complete 14; expect_phase junit 0
  expect_jx fast:fast "$(jexpect fast complete)"
  if grep -q 'B-J4A-ODD' "$(JX)" 2>/dev/null; then prob "the unknown-status row reached junit.xml"; fi
  r+=" unknown-status=$RC"
  std_ci 4307; run_v "${CIENV[@]}" STUB_HARNESS=rm-results -- --profile fast --publish-dir ci-evidence
  expect_rc 14; expect_line has harness=0; expect_phase complete 14; expect_phase junit 0
  expect_jx fast:fast "validate.fast.phase phase:complete phase-error"
  r+=" missing=$RC"
  std_ci 4308; run_v "${CIENV[@]}" STUB_HARNESS=fail,malformed -- --profile fast --publish-dir ci-evidence
  expect_rc 1; expect_line has harness=1; expect_phase complete 14; expect_phase junit 0
  expect_jx fast:fast "$(jexpect fast harness complete)"
  r+=" malformed-with-harness-1=$RC"
  detail "exit:$r; each publishes phase:complete with only the well-formed rows (none when results.tsv is missing); junit=0"
}
# J4b (AC2): the converter alone, on sanitized inputs: each defect returns 19 and writes nothing; the intentional
# phase-only cases (executed profile none, harness not-run) succeed with no check case.
ctl_J4b() {
  local d names k n=0 rc out
  new_case; d="$C/j4b"; mkdir -p "$d"
  names="$(sed -n '/^PHASES=(/,/)/p' "$C/repo/infra/ci/validate.sh" | tr '\n' ' ' | sed -E 's/^PHASES=\(//; s/\).*$//; s/ +/ /g')"
  printf 'B-A\tPASS\textra\n' >"$d/bad-line.tsv"
  printf 'B-A\tSKIP\n' >"$d/unknown.tsv"
  printf 'B-A\tPASS\nB-A\tFAIL\n' >"$d/dup.tsv"
  printf 'B-A\tPASS\n' >"$d/ok.tsv"
  # NAME JOB PROFILE COLUMNS PHASE=STATUS...: each must return 19 with no OUT and no OUT.tmp
  for k in "bad-line fast fast $d/bad-line.tsv complete=0" "unknown fast fast $d/unknown.tsv complete=0" \
    "duplicate fast fast $d/dup.tsv complete=0" "missing fast fast $d/absent.tsv complete=0" \
    "job fast:x fast $d/ok.tsv complete=0" "profile fast everything $d/ok.tsv complete=0" \
    "phase-name fast fast $d/ok.tsv bogus=1" "phase-status fast fast $d/ok.tsv harness=weird"; do
    set -- $k
    rc=0; out="$(bash -c 'source "$1"; shift; junit_report "$@"' j4b "$C/repo/infra/ci/junit.sh" "$d/$1.xml" "$2" "$3" "$4" "$names" "${@:5}" 2>&1)" || rc=$?
    if [[ "$rc" != 19 || -n "$out" ]]; then prob "$1: exit $rc output [$out], want 19 and none"; fi
    if [[ -e "$d/$1.xml" || -e "$d/$1.xml.tmp" ]]; then prob "$1: a report was written"; fi
    n=$((n + 1))
  done
  rc=0; out="$(bash -c 'source "$1"; shift; junit_report "$@"' j4b "$C/repo/infra/ci/junit.sh" "$d/none.xml" gate none "" "$names" \
    location=0 preflight=0 state=0 harness=not-required postcheck=0 cleanup=0 raw_scan=0 complete=n/a stage=0 summary=0)" || rc=$?
  if [[ "$rc" != 0 || "$out" != "0 0" ]]; then prob "executed profile none: exit $rc [$out], want 0 [0 0]"; fi
  if [[ "$(jsuite "$d/none.xml" 2>/dev/null)" != "gate:none 0 0" || -n "$(jcases "$d/none.xml" 2>/dev/null)" ]]; then prob "executed profile none is not an empty gate:none suite"; fi
  rc=0; out="$(bash -c 'source "$1"; shift; junit_report "$@"' j4b "$C/repo/infra/ci/junit.sh" "$d/not-run.xml" fast fast "" "$names" \
    location=0 preflight=16 harness=not-run postcheck=0 cleanup=0 raw_scan=0 complete=n/a stage=0 summary=0)" || rc=$?
  if [[ "$rc" != 0 || "$out" != "1 1" ]]; then prob "harness not-run: exit $rc [$out], want 0 [1 1]"; fi
  if [[ "$(jcases "$d/not-run.xml" 2>/dev/null)" != "validate.fast.phase phase:preflight phase-error" ]]; then prob "harness not-run: cases [$(jcases "$d/not-run.xml" 2>/dev/null)], want phase:preflight only"; fi
  # Non-additive timing (2026-10-19): overlapping suites (12 + 9 + 30 = 51 s) and a failed harness phase (40 s)
  # inside a measured validation elapsed of 41 s give exactly one time attribute, 41, on the one
  # timing:validation-elapsed case, for a passing and a failing input alike; suite rows are never cases; check
  # and phase failures and their count are unchanged; without an elapsed row there is no timing case.
  printf 'B-A\tFAIL\n' >"$d/fail.tsv"
  printf 'suite\troot\t12\nsuite\tmigrations\t9\nsuite\tci-controls\t30\nphase\tharness\t40\nelapsed\tvalidation\t41\n' >"$d/times.tsv"
  printf 'suite\troot\t12\nsuite\tmigrations\t9\nphase\tharness\t40\n' >"$d/times-noelapsed.tsv"
  local tc="validate.gate.timing timing:validation-elapsed pass" hc="validate.gate.phase phase:harness phase-error" want
  # NAME TIMES COLUMNS WANT-COUNTS WANT-TIMES PHASE=STATUS...
  for k in "pass times $d/ok.tsv 2_0 41 harness=0" "fail times $d/fail.tsv 3_2 41 harness=1" \
    "phase-only times - 2_1 41 harness=1" "no-elapsed times-noelapsed $d/ok.tsv 1_0 - harness=0"; do
    set -- $k
    [[ "$3" == - ]] && set -- "$1" "$2" "" "${@:4}"
    [[ "$5" == - ]] && set -- "${@:1:4}" "" "${@:6}"
    case "$1" in
      pass) want="validate.gate.check B-A pass"$'\n'"$tc" ;;
      fail) want="validate.gate.check B-A FAIL"$'\n'"$tc"$'\n'"$hc" ;;
      phase-only) want="$tc"$'\n'"$hc" ;;
      no-elapsed) want="validate.gate.check B-A pass" ;;
    esac
    rc=0; out="$(JUNIT_DURATIONS_FILE="$d/$2.tsv" bash -c 'source "$1"; shift; junit_report "$@"' jt "$C/repo/infra/ci/junit.sh" "$d/t-$1.xml" gate full "$3" "$names" "${@:6}" 2>&1)" || rc=$?
    if [[ $rc != 0 || "$out" != "${4/_/ }" ]]; then prob "timing $1: exit $rc [$out], want 0 [${4/_/ }]"; continue; fi
    if [[ "$(jcases "$d/t-$1.xml")" != "$want" ]]; then prob "timing $1: cases [$(jcases "$d/t-$1.xml" | tr '\n' ';')], want [$(tr '\n' ';' <<<"$want")]"; fi
    if [[ "$(grep -oE ' time="[0-9]+"' "$d/t-$1.xml" | tr -dc '0-9\n' | paste -sd' ' || true)" != "$5" ]]; then
      prob "timing $1: case times [$(grep -oE ' time="[0-9]+"' "$d/t-$1.xml" | tr -dc '0-9\n' | paste -sd' ' || true)], want exactly [$5] (no suite/phase addition)"; fi
    if [[ -n "$5" ]] && ! grep -qF 'name="timing:validation-elapsed" time="41"/>' "$d/t-$1.xml"; then prob "timing $1: the elapsed time is not on the timing case"; fi
  done
  for k in "suite\troot\tinvalid" "elapsed\tvalidation\t1\nelapsed\tvalidation\t2" "elapsed\tsuites\t3"; do
    rm -f -- "$d/bad-time.xml"; printf "$k\n" >"$d/bad-time.tsv"
    rc=0; JUNIT_DURATIONS_FILE="$d/bad-time.tsv" bash -c 'source "$1"; shift; junit_report "$@"' jt "$C/repo/infra/ci/junit.sh" "$d/bad-time.xml" gate full "$d/ok.tsv" "$names" >/dev/null 2>&1 || rc=$?
    if [[ $rc != 19 || -e "$d/bad-time.xml" ]]; then prob "invalid duration row accepted [$k]"; fi
  done
  detail "$n converter defects (bad line, unknown status, duplicate, missing columns, job, profile, phase name, phase status) each 19 with nothing written; executed profile none: empty gate:none suite; harness not-run: phase:preflight only; timing: overlapping suites 51 s + harness 40 s under a 41 s elapsed give one time attribute, 41, on timing:validation-elapsed, no suite case (pass 2/0, fail 3/2, phase-only 2/1; no elapsed row: no timing case); invalid, repeated or misnamed duration rows 19"
}
# J4c (AC3): on an otherwise successful job, a duplicate non-required PASS ID (which completeness allows) and a
# failed parse self-check (the shim fails the yq XML parse) each give junit=19, overall 19, no junit.xml; the rest
# of the candidate is published unchanged.
ctl_J4c() {
  local want r1
  want="${FAST_SET/junit.xml /}"
  std_ci 4310; run_v "${CIENV[@]}" STUB_HARNESS=dup-extra -- --profile fast --publish-dir ci-evidence
  expect_rc 19; expect_line has harness=0; expect_phase complete 0; expect_phase junit 19; expect_pub full; expect_runs 1
  expect_no_jx
  if [[ "$(published_list)" != "$want" ]]; then prob "duplicate: published set [$(published_list)], want the fast set without junit.xml"; fi
  log_has "junit: the report inputs are invalid; no junit.xml"
  r1="$RC"
  std_ci 4318; run_v "${CIENV[@]}" SHIM_YQ_XML_FAIL=1 -- --profile fast --publish-dir ci-evidence
  expect_rc 19; expect_line has harness=0; expect_phase complete 0; expect_phase junit 19; expect_pub full; expect_runs 1
  expect_no_jx
  if [[ "$(published_list)" != "$want" ]]; then prob "self-check: published set [$(published_list)], want the fast set without junit.xml"; fi
  log_has "junit: the written report failed its self-check; no junit.xml"
  detail "duplicate non-required ID: exit $r1; failed parse self-check: exit $RC; each complete=0, junit=19, the fast set published without junit.xml"
}
# J5 (AC3): a failing harness keeps its status with a valid report; with a report defect as well it is still 1;
# postcheck, cleanup and scrub run and the harness ran once.
ctl_J5() {
  local r1 r2
  std_ci 4311; run_v "${CIENV[@]}" STUB_HARNESS=fail -- --profile fast --publish-dir ci-evidence
  expect_rc 1; expect_line has harness=1; expect_phase junit 0; expect_f1f2; expect_phase scrub 0; expect_runs 1; expect_pub full
  expect_jx fast:fast "$(jexpect fast harness complete)"
  expect_jmsg "check failed"; expect_jmsg "phase harness exited 1"
  r1="$RC"
  std_ci 4312; run_v "${CIENV[@]}" STUB_HARNESS=fail,dup-extra -- --profile fast --publish-dir ci-evidence
  expect_rc 1; expect_line has harness=1; expect_phase junit 19; expect_f1f2; expect_phase scrub 0; expect_runs 1; expect_pub full
  expect_no_jx
  r2="$RC"
  # 2026-10-19 timing repair: the failing harness with overlapping suite timings keeps the same failure cases and
  # count; the one timing case is added between the checks and the phases.
  std_ci 4313; run_v "${CIENV[@]}" STUB_HARNESS=fail STUB_SUITE_TIMES=1 -- --profile fast --publish-dir ci-evidence
  expect_rc 1; expect_line has harness=1; expect_phase junit 0; expect_runs 1; expect_pub full
  expect_jx fast:fast "$(jexpect fast)"$'\n'"validate.fast.timing timing:validation-elapsed pass"$'\n'"validate.fast.phase phase:harness phase-error"$'\n'"validate.fast.phase phase:complete phase-error"
  jtiming
  detail "harness 1 with a valid report: exit $r1, the FAIL row and phase:harness reported; harness 1 with a report defect: exit $r2, junit=19, no junit.xml; postcheck, cleanup and scrub ran; one harness run each; harness 1 with overlapping suite timings: exit $RC, the same failure cases plus one timing case, the only time attribute [$JT] s"
}
# J6 (AC2, AC4): a complete-phase failure without a failing check; a STAGING-FAILED candidate (checks plus
# phase:stage); WITHHELD candidates from the raw scan (the V5a and V5c plants) and from the candidate scan (a
# gitleaks token in a tracked comment of the case tree, outside the raw evidence) carry phase cases only.
ctl_J6() {
  local r=""
  std_ci 4313; run_v "${CIENV[@]}" STUB_DROP_ID=B-UNIT-CHECKS -- --profile fast --publish-dir ci-evidence
  expect_rc 14; expect_phase complete 14; expect_phase junit 0
  expect_jx fast:fast "$(jexpect fast complete)"; expect_jmsg "phase complete exited 14"
  if grep -q 'type="FAIL"\|type="UNMET"' "$(JX)" 2>/dev/null; then prob "a check case fails"; fi
  r+=" complete-only=$RC"
  std_ci 4314; run_v "${CIENV[@]}" STUB_HARNESS=partial -- --profile fast --publish-dir ci-evidence
  expect_rc 19; expect_phase stage 19; expect_phase junit 0; expect_pub minimal:STAGING-FAILED
  expect_jx fast:fast "$(jreq fast fast; printf 'validate.fast.phase phase:stage phase-error')"
  r+=" staging-failed=$RC"
  std_ci 4315; run_v "${CIENV[@]}" STUB_HARNESS=plant-retained -- --profile fast --publish-dir ci-evidence
  expect_rc 15; expect_phase raw_scan 15; expect_phase junit 0; expect_pub minimal:WITHHELD; expect_no_canary
  expect_jx fast:fast "validate.fast.phase phase:raw_scan phase-error"; expect_jmsg "phase raw_scan exited 15"
  if [[ -s "$C/stub/canary" ]] && grep -qF -f "$C/stub/canary" "$C/out"; then prob "the planted canary reached the job log"; fi
  r+=" raw-withheld=$RC"
  std_ci 4316; run_v "${CIENV[@]}" STUB_HARNESS=plant-gitleaks -- --profile fast --publish-dir ci-evidence
  expect_rc 15; expect_phase raw_scan 15; expect_phase junit 0; expect_pub minimal:WITHHELD
  expect_jx fast:fast "validate.fast.phase phase:raw_scan phase-error"
  r+=" gitleaks-withheld=$RC"
  new_case
  printf '# J6 plant: token ghp_%s\n' "$(head -c 400 /dev/urandom | tr -dc 'A-Za-z0-9' | head -c 36)" >>"$C/repo/infra/validation/required-checks.tsv"
  git -C "$C/repo" add infra/validation/required-checks.tsv
  GIT_AUTHOR_DATE=2026-01-02T00:00:00Z GIT_COMMITTER_DATE=2026-01-02T00:00:00Z \
    git -C "$C/repo" -c user.name=ci-controls -c user.email=controls@example.invalid commit -q -m "J6 candidate-scan plant"
  ci_env 4317; run_v "${CIENV[@]}" -- --profile fast --publish-dir ci-evidence
  expect_rc 15; expect_phase raw_scan 0; expect_phase junit 0; expect_phase candidate_scan 15; expect_phase summary_withheld 0
  expect_phase junit_withheld 0; expect_phase rescan 0; expect_pub minimal:WITHHELD
  expect_jx fast:fast "validate.fast.phase phase:candidate_scan phase-error"; expect_jmsg "phase candidate_scan exited 15"
  if grep -q 'ghp_' "$(JX)" 2>/dev/null; then prob "the planted token reached junit.xml"; fi
  r+=" candidate-withheld=$RC"
  detail "exit:$r; phase:complete without a failing check; STAGING-FAILED: checks plus phase:stage; WITHHELD (raw scan, gitleaks, candidate scan with junit_withheld): phase cases only"
}
# J7 (G051): the fast job of an affected pipeline reports fast:affected with exactly the published TSV's cases.
ctl_J7() {
  local exp n
  va_area 4320; run_v "${CIENV[@]}" -- --profile fast --publish-dir ci-evidence
  expect_rc 0; expect_phase junit 0; expect_pub full
  exp="$(jexpect fast)"; n="$(grep -c . <<<"$exp")" || n=0
  if (( n != 20 )); then prob "the affected result has $n check rows, want 20"; fi
  expect_jx fast:affected "$exp"
  log_has "junit: junit.xml with $n cases, 0 failing (supplementary)"
  detail "exit $RC; fast:affected with the $n check cases of results.tsv, in its order"
}
# J8 (G051): the gate's partial coverage reports gate:remainder with exactly the TSV's cases; the suite never says full.
ctl_J8() {
  local exp n
  std_ci 4321; run_v "${CIENV[@]}" -- --profile gate --publish-dir ci-evidence
  expect_rc 0; expect_phase junit 0; expect_pub full
  exp="$(jexpect gate)"; n="$(grep -c . <<<"$exp")" || n=0
  if (( n != 97 )); then prob "the remainder result has $n check rows, want 97"; fi
  expect_jx gate:remainder "$exp"
  if grep -q 'name="gate:full"' "$(JX)" 2>/dev/null; then prob "the gate's suite states full"; fi
  detail "exit $RC; gate:remainder with the $n check cases of results.tsv (coverage partial), never gate:full"
}

# --- group D: per-test Go evidence (G064 Plan r2 §5.4; stub harness, real validate.sh adapter and scanners) ---------
# GT1: the full run executes migrations and ops and writes every PASS row but no go-tests.tsv: complete 14 names the
# missing report; the scanned evidence set is published as on any completeness failure (V3), without a go-tests.tsv.
# That partial publication is a recorded, accepted deviation from G064 Plan r2 §5.4's "nothing published" wording
# (Codex source review, G064 Result "S4 correction 2"): the existing V3 publication path is kept, go-tests.tsv is
# omitted and the missing report is never synthesized or treated as complete.
ctl_GT1() {
  std_ci 4330; run_v "${CIENV[@]}" STUB_RM_REPORT=go-tests.tsv -- --publish-dir ci-evidence
  expect_rc 14; expect_line has harness=0; expect_phase complete 14; expect_pub full
  log_has "complete: go-tests: go-tests.tsv is missing"
  if [[ -e "$(PUBDIR)/go-tests.tsv" ]]; then prob "a go-tests.tsv was published"; fi
  if [[ "$(published_list)" != "${FULL_SET/go-tests.tsv /}" ]]; then prob "published set differs from the full set without go-tests.tsv: [$(published_list)]"; fi
  detail "exit $RC; every row PASS, harness=0, go-tests.tsv absent: complete=14 [go-tests: go-tests.tsv is missing]; nothing stands in for the missing report; the other scanned evidence is published without it (recorded deviation from Plan r2 §5.4 'nothing published')"
}
# GT2: migrations declares two invocations, one failed (a failing test row, a FAIL row in results.tsv and a
# go-tests/<n>.log holding the planted generated secret) and one interrupted: complete 14 names the interrupted
# invocation (not the failed one, which is complete) and still reports the FAIL row; the real raw scan finds the secret
# under go-tests/ and withholds the evidence (exit 15 ranks above 14), so the secret is never published.
ctl_GT2() {
  std_ci 4331; run_v "${CIENV[@]}" STUB_HARNESS=go-tests-fail-interrupted STUB_FAIL_ID=B-MIGRATE-IT -- --publish-dir ci-evidence
  expect_rc 15; expect_line has harness=0; expect_phase raw_scan 15; expect_phase complete 14
  log_has "complete: go-tests: invocation stub-mig-int is interrupted (rc 255)"
  log_has "complete: required check B-MIGRATE-IT is not PASS"
  log_has "scan: raw: a generated secret was found"
  if grep -q '^validate: complete: go-tests: .*stub-mig-fail' "$C/out"; then prob "the failed (complete) invocation was reported as incomplete"; fi
  expect_pub minimal:WITHHELD; expect_no_canary
  detail "exit $RC; complete=14 [invocation stub-mig-int is interrupted (rc 255)] and [required check B-MIGRATE-IT is not PASS]; raw_scan=15 on the secret in go-tests/stub-mig-fail.log: WITHHELD, the secret unpublished"
}

# --- group D: background supervision (G051 VB1-VB5; Plan §3.6) ----------------------------------------------------
# VB2-VB5: a harness-shaped driver sources the real infra/validation/lib.sh (bg_start, bg_join, bg_wait_optional,
# bg_stop_all) with run.sh's EXIT order (bg_stop_all, then cleanup); its cleanup records whether any child was still
# alive and removes the run-labelled container by its cidfile. The driver is a tracked launch (its own process
# group), so a group TERM reaches the driver and its children as validate.sh's would reach run.sh. VB5 also runs the
# driver's join-lost cases (G064 correction 11: a child status bash no longer holds, a reused PID), which need no Docker.
vb_driver() { # write $C/driver.sh
  cat >"$C/driver.sh" <<'VBDRV'
#!/usr/bin/env bash
# VB driver of infra/ci/validate-controls.sh: the real background helpers in a run.sh-shaped shell.
#   driver.sh REPO WORK MODE LABELS   MODE: exit2 | term | prebuild | join | join-foreign | join-lost | witness-abort
# join-lost needs no Docker: extract this heredoc and run it with REPO = a checkout and LABELS = "" (lost.txt: verdicts).
# witness-abort (VB7): VB_REG names the case's registry.
set -euo pipefail
REPO="$1" WORK="$2" MODE="$3" LABELS="$4"
# shellcheck source=/dev/null
source "$REPO/infra/validation/lib.sh"
RUN_ID="vb$$" RESULTS="$WORK/results.tsv" HARNESS_LOG="$WORK/harness.log"
: >"$RESULTS"; : >"$HARNESS_LOG"; : >"$WORK/child.pids"
lbl=()
for l in $LABELS; do lbl+=(--label "$l"); done
drv_cleanup() {
  local p alive=""
  while read -r p; do if [[ -n "$p" ]] && kill -0 "$p" 2>/dev/null; then alive+=" $p"; fi; done <"$WORK/child.pids"
  printf 'alive=%s\n' "${alive:-none}" >"$WORK/cleanup-saw"
  if [[ -s "$WORK/ctr.cid" ]]; then docker rm -f "$(cat "$WORK/ctr.cid")" >/dev/null 2>&1 || true; fi
}
trap 'bg_stop_all; drv_cleanup' EXIT
hold() { # a child holding a run-labelled container; ignore-term: TERM is ignored, the child ends after 5 s
  printf '%s\n' "$BASHPID" >>"$WORK/child.pids"
  if [[ "${1:-}" == ignore-term ]]; then trap '' TERM; fi
  docker run -d --cidfile "$WORK/ctr.cid" "${lbl[@]}" --network none busybox:1.37.0 sleep 300 >/dev/null
  : >"$WORK/holding"
  if [[ "${1:-}" == ignore-term ]]; then sleep 5 & else sleep 300 & fi
  printf '%s\n' "$!" >>"$WORK/child.pids"
  wait "$!"
}
holding() { local i; for ((i = 0; i < 600; i++)); do if [[ -e "$WORK/holding" ]]; then return 0; fi; sleep 0.1; done; return 1; }
fake_prebuild() { record B-VB-OWNER PASS "a prebuild row (never merged)"; sleep 1; exit 3; }
ctl_rows() {
  record B-DEPS-CONTROLS PASS vb; record B-WRAPPER-CONTROLS PASS vb
  if [[ "${1:-}" == foreign ]]; then record B-SHELL PASS "a foreign row"; fi
  return 0
}
ctl_ci_rows() { record B-CI-CONTROLS PASS vb; }
# join-lost (G064 correction 11): bash no longer attributes the child's status. lost_join joins from a subshell of this
# shell, where the ended child is not a child: its wait returns 127 ("not a child of this shell"), as after a later
# child of the harness reused the PID (G064 full run 2). reuse*: the PID names another live child of this shell.
# reuse-diff: the child recorded exit 3, and its PID now names an ended, reaped child of this shell whose saved status
# (5) wait returns. Each case has its own RESULTS; lost_expect appends one verdict (ok|bad CASE: ...) to lost.txt.
ctl_exit3() { ctl_rows; exit 3; }
ctl_hold() { ctl_rows; : >"$WORK/$BG_CHILD.ready"; while [[ ! -e "$WORK/$BG_CHILD.release" ]]; do sleep 0.1; done; }
lost_gone() { local i; for ((i = 0; i < 600; i++)); do if [[ ! -e "/proc/${BG_PID[$1]}" ]]; then return 0; fi; sleep 0.05; done; return 1; }
lost_ready() { local i; for ((i = 0; i < 600; i++)); do if [[ -e "$WORK/$1.ready" ]]; then return 0; fi; sleep 0.05; done; return 1; }
lost_join() { RESULTS="$WORK/case-$1.tsv"; ( bg_join "$1" controls ); BG_DONE[$1]=1; }
lost_expect() { # CASE STATUS PATTERN: the controls rows once each and one BACKGROUND row with STATUS matching PATTERN
  local f="$WORK/case-$1.tsv" why="" row
  row="$(awk -F'\t' '$1 == "BACKGROUND" {printf "%s%s %s", s, $2, $3; s = "; "}' "$f")"
  if [[ "$(grep -c . "$f")" != 3 || "$(grep -c $'^B-DEPS-CONTROLS\tPASS\t' "$f")" != 1 || "$(grep -c $'^B-WRAPPER-CONTROLS\tPASS\t' "$f")" != 1 ]]; then
    why+="the controls rows were not merged exactly once; "
  fi
  if awk -F'\t' -v s="$2" -v t="$3" '$1 == "BACKGROUND" {n++; if ($2 == s && $3 ~ t) m++} END {exit !(n == 1 && m == 1)}' "$f"; then :
  else why+="want one BACKGROUND $2 matching [$3]; "; fi
  if [[ -z "$why" ]]; then printf 'ok %s: %s\n' "$1" "$row"; else printf 'bad %s: %s(BACKGROUND rows: %s)\n' "$1" "$why" "${row:-none}"; fi >>"$WORK/lost.txt"
}
case "$MODE" in
  exit2) bg_start holder hold; holding; exit 2 ;;
  term) bg_start holder hold ignore-term; holding; : >"$WORK/ready"; sleep 600 & wait "$!" ;;
  prebuild)
    bg_start prebuild-vb fake_prebuild
    bg_wait_optional prebuild-vb
    record B-VB-OWNER FAIL "the authoritative (fake) build failed" ;;
  join) bg_start controls ctl_rows; bg_start ci-controls ctl_ci_rows; bg_join controls controls; bg_join ci-controls ci-controls ;;
  join-foreign) bg_start controls ctl_rows foreign; bg_start ci-controls ctl_ci_rows; bg_join controls controls; bg_join ci-controls ci-controls ;;
  join-lost)
    : >"$WORK/lost.txt"
    for c in lost-ok lost-norc lost-badrc lost-bigrc reuse reuse-norc; do bg_start "$c" ctl_rows; done
    bg_start lost-exit3 ctl_exit3; bg_start reuse-diff ctl_exit3
    for c in term lost-term live-127; do bg_start "$c" ctl_hold; done
    for c in term lost-term live-127; do lost_ready "$c"; done
    kill -TERM "${BG_PID[term]}" "${BG_PID[lost-term]}"
    for c in lost-ok lost-norc lost-badrc lost-bigrc lost-exit3 reuse reuse-norc reuse-diff term lost-term; do lost_gone "$c"; done
    rm -f -- "$WORK/bg/lost-norc/rc" "$WORK/bg/reuse-norc/rc"
    printf 'x0\n' >"$WORK/bg/lost-badrc/rc"; printf '256\n' >"$WORK/bg/lost-bigrc/rc"
    for c in lost-ok lost-norc lost-badrc lost-bigrc lost-exit3 lost-term live-127; do lost_join "$c"; done
    : >"$WORK/live-127.release"; wait "${BG_PID[live-127]}" || true
    RESULTS="$WORK/case-term.tsv"; bg_join term controls
    # The ended child's PID now names another live child of this shell: never waited for (it would block on or
    # attribute that child); the status comes from the child's own record.
    sleep 10 & decoy=$!
    BG_PID[reuse]="$decoy" BG_PID[reuse-norc]="$decoy"
    for c in reuse reuse-norc; do
      RESULTS="$WORK/case-$c.tsv"; bg_join "$c" controls
      if kill -0 "$decoy" 2>/dev/null; then :; else printf 'bad %s: bg_join waited for the unrelated child %s\n' "$c" "$decoy" >>"$WORK/lost.txt"; fi
    done
    kill "$decoy" 2>/dev/null || true; wait "$decoy" 2>/dev/null || true
    # The ended child's PID now names an ended child of this shell that bash reaped and whose status (5) it saved.
    ( exit 5 ) & decoy=$!
    BG_PID[reuse-diff]="$decoy"
    for ((i = 0; i < 600; i++)); do if [[ ! -e "/proc/$decoy" ]]; then break; fi; sleep 0.05; done
    RESULTS="$WORK/case-reuse-diff.tsv"; bg_join reuse-diff controls
    RESULTS="$WORK/results.tsv"
    lost_expect lost-ok INFO '^lost-ok joined [(]exit 0, [0-9]+ s[)]$'
    lost_expect lost-exit3 FAIL '^lost-exit3 exited 3$'
    lost_expect lost-norc FAIL '^lost-norc: exit status unknown [(]pid [0-9]+ is no longer a child of this shell[)] and none recorded$'
    lost_expect lost-badrc FAIL '^lost-badrc: invalid recorded exit status$'
    lost_expect lost-bigrc FAIL '^lost-bigrc: invalid recorded exit status$'
    lost_expect lost-term FAIL '^lost-term: exit status unknown [(]pid [0-9]+ is no longer a child of this shell[)] and none recorded$'
    if [[ -e "$WORK/bg/lost-term/rc" ]]; then printf 'bad lost-term: the TERMed child left an exit record\n' >>"$WORK/lost.txt"; fi
    lost_expect live-127 FAIL '^live-127: exit status unknown [(]wait for its running process [(]pid [0-9]+[)] returned 127[)] and none recorded$'
    lost_expect term FAIL '^term exited 143$'
    lost_expect reuse INFO '^reuse joined [(]exit 0, [0-9]+ s[)]$'
    lost_expect reuse-norc FAIL '^reuse-norc: exit status unknown [(]pid [0-9]+ now names another process[)] and none recorded$'
    lost_expect reuse-diff FAIL '^reuse-diff: wait reported exit 5 but the child recorded exit 3$'
    if grep -q '^bad ' "$WORK/lost.txt"; then exit 1; fi ;;
  witness-abort)
    # G065 K3 (VB7b): the real witness on the case's registry, then an abort before res_witness_stop, in run.sh's EXIT order.
    OWNED_REG="${VB_REG:?}"
    trap 'bg_stop_all; res_witness_abort || log "witness abort cleanup failed"; drv_cleanup' EXIT
    res_witness_start
    exit 2 ;;
esac
if grep -qE $'\t(FAIL|UNMET)\t' "$RESULTS"; then exit 1; fi
exit 0
VBDRV
}
vb_launch() { # MODE: the driver as a tracked launch (VPID); its WORK is $C/work
  new_case; ctl_labels; vb_driver; mkdir -p "$C/work"
  launch "$C/drv.out" env -i -C "$C" "${BASE_ENV[@]}" TMPDIR="$C/tmp" SHIM_LABELS="$(case_labels)" \
    bash "$C/driver.sh" "$C/repo" "$C/work" "$1" "$(case_labels)"
  VPID="$LPID"
}
vb_row() { # STATUS TEXT-PATTERN: RESULTS has a BACKGROUND (or PREBUILD) row with STATUS whose text matches
  if awk -F'\t' -v s="$1" -v t="$2" '($1 == "BACKGROUND" || $1 == "PREBUILD") && $2 == s && $3 ~ t {f = 1} END {exit !f}' "$C/work/results.tsv" 2>/dev/null; then :
  else prob "no $1 row matching [$2] (rows: $(awk -F'\t' '$1 == "BACKGROUND" || $1 == "PREBUILD"' "$C/work/results.tsv" 2>/dev/null | tr '\t\n' ' ;'))"; fi
}
vb_gone() { # every recorded child and every process of this case is gone; the container and labelled resources too
  local p n
  while read -r p; do if [[ -n "$p" ]] && kill -0 "$p" 2>/dev/null; then prob "child process $p is alive"; fi; done <"$C/work/child.pids"
  if ps -eo args= | grep -F -- "$C/" | grep -vq -- 'grep -F'; then prob "a process of this case is still running"; fi
  if [[ -s "$C/work/ctr.cid" ]] && exists container "$(cat "$C/work/ctr.cid")"; then prob "the held container still exists"; fi
  n="$(lcount "sdlc.ci.control=$RT-$CTL_ID")"
  if (( n < 0 )); then prob "the Docker listing of sdlc.ci.control=$RT-$CTL_ID failed; the count is unverified"
  elif [[ "$n" != 0 ]]; then prob "$n labelled resources remain before the sweep (inventory leftover)"; fi
}
ctl_VB2() {
  vb_launch exit2; await "$VPID" 180
  if [[ "$RC" != 2 ]]; then prob "driver exit $RC, want 2"; fi
  if [[ "$(cat "$C/work/cleanup-saw" 2>/dev/null)" != alive=none ]]; then prob "cleanup ran with a live child: $(cat "$C/work/cleanup-saw" 2>/dev/null)"; fi
  vb_row FAIL '^holder stopped before join'; vb_gone
  detail "foreground exit 2 with a child holding a container: exit $RC; the child was reaped before cleanup (BACKGROUND FAIL stopped before join); container removed; no labelled leftover"
}
ctl_VB3() {
  vb_launch term
  if wait_file "$C/work/ready" 60; then kill -TERM -- "-$VPID"; else prob "the driver never held"; fi
  await "$VPID" 180
  if [[ "$RC" != 143 ]]; then prob "driver exit $RC, want 143"; fi
  if [[ "$(cat "$C/work/cleanup-saw" 2>/dev/null)" != alive=none ]]; then prob "cleanup ran with a live child: $(cat "$C/work/cleanup-saw" 2>/dev/null)"; fi
  vb_row FAIL '^holder stopped before join'; vb_gone
  detail "group TERM while the child ignores TERM for 5 s: exit $RC; reaped before cleanup; container removed"
}
ctl_VB4() {
  vb_launch prebuild; await "$VPID" 180
  if [[ "$RC" != 1 ]]; then prob "driver exit $RC, want 1"; fi
  vb_row INFO '^prebuild-vb exit 3; authoritative build follows$'
  if [[ "$(awk -F'\t' '$1 == "B-VB-OWNER" {print $2}' "$C/work/results.tsv" 2>/dev/null | tr '\n' ' ')" != "FAIL " ]]; then prob "the owning check is not exactly one FAIL row (a prebuild row was merged?)"; fi
  if grep -qP '^PREBUILD\tPASS' "$C/work/results.tsv" 2>/dev/null; then prob "a prebuild recorded PASS"; fi
  vb_gone
  detail "a failing prebuild then a failing authoritative build: exit $RC; PREBUILD INFO plus the owning FAIL only; no live child"
}
# VB6 (G051 attempt 5 (a)): the real lib.sh cleanup() in run.sh's EXIT order. A driver takes the start inventory,
# builds a control-labelled image it records as built, starts a container from it and exits. cleanup() waits (about
# 15 s) until no container uses that image before removing it without -f. wait: a short-lived --rm container and a
# registered project whose name prefixes the image (listed twice, as a unit suite's); the image is removed once and
# the run keeps exit 0. timeout: a long-lived container, the image is kept, the reason logged
# and an otherwise passing run exits 1. earlier-failure: the same, but the run's earlier exit 3 is kept. G051 C1: no
# container, and the shim fails the removal. gone: the image really went, the presence query shows it absent, exit 0.
# kept: the image is still present, logged kept, exit 1. unknown: the presence query fails too, logged unverified,
# exit 1; unknown-earlier: the same after an earlier exit 3, which is kept. The image is recorded with its tag, as
# run.sh records it, and the shim log shows it removed once. lib.sh's log() writes to stdout (run.sh tees it into
# harness.log); here that is the driver's launch output, drv6.out.
vb6_case() { # MODE (wait|timeout|earlier-failure|gone|kept|unknown|unknown-earlier) [SHIM_VAR=VALUE...] -> RC, VB6_SECS
  local img="g001-vb6-$RT-$1" t0
  new_case; ctl_labels
  cat >"$C/driver6.sh" <<'VB6DRV'
#!/usr/bin/env bash
set -euo pipefail
REPO="$1" WORK="$2" MODE="$3" IMG="$4" LABELS="$5"
# shellcheck source=/dev/null
source "$REPO/infra/validation/lib.sh"
RUN_ID="vb6$$" HARNESS_LOG="$WORK/harness.log" KEEP=1
lbl=(); for l in $LABELS; do lbl+=(--label "$l"); done
docker_inventory >"$WORK/inventory-start.txt"
trap 'RUN_STATUS=$?; bg_stop_all; cleanup' EXIT
printf 'FROM busybox:1.37.0\nLABEL g001.vb6=%s\n' "$RUN_ID" | docker build -q "${lbl[@]}" -t "$IMG" - >/dev/null
printf '%s:latest\n' "$IMG" >"$WORK/images-built"
case "$MODE" in
  # wait also registers a project whose name prefixes the image, so cleanup sees the image twice (a unit suite's case).
  wait) printf '%s\n' "${IMG%-wait}" >"$WORK/projects"; docker run -d --rm "${lbl[@]}" --network none "$IMG" sleep 3 >"$WORK/ctr.id" ;;
  gone|kept|unknown|unknown-earlier) ;;
  *) docker run -d "${lbl[@]}" --network none "$IMG" sleep 120 >"$WORK/ctr.id" ;;
esac
if [[ "$MODE" == earlier-failure || "$MODE" == unknown-earlier ]]; then exit 3; fi
exit 0
VB6DRV
  mkdir -p "$C/work"
  t0="$SECONDS"
  launch "$C/drv6.out" env -i -C "$C" "${BASE_ENV[@]}" TMPDIR="$C/tmp" SHIM_LOG="$C/shim.log" "${@:2}" \
    bash "$C/driver6.sh" "$C/repo" "$C/work" "$1" "$img" "$(case_labels)"
  await "$LPID" 120
  VB6_SECS=$((SECONDS - t0)) VB6_IMG="$img"
  VB6_CTR="$(cat "$C/work/ctr.id" 2>/dev/null)" || VB6_CTR=""
}
vb6_tidy() { # the control's own container and image, by exact ID
  if [[ -n "$VB6_CTR" ]]; then D rm -f "$VB6_CTR" >/dev/null 2>&1 || true; fi
  D image rm "$VB6_IMG" >/dev/null 2>&1 || true
}
vb6_calls() { grep -cxF "$1" "$C/shim.log" 2>/dev/null || true; } # LINE -> how often the shim saw exactly that call
vb6_failed_rm() { # MODE RC_WANT PRESENT(yes|no) LOG_TEXT (empty: no cleanup record) -> a problem per mismatch
  local i="$VB6_IMG:latest"
  if [[ "$RC" != "$2" ]]; then prob "$1: exit $RC, want $2"; fi
  if exists image "$VB6_IMG"; then [[ "$3" == yes ]] || prob "$1: the image is still present"
  else [[ "$3" == no ]] || prob "$1: the image is gone although its removal was refused"; fi
  if [[ -z "$4" ]]; then
    if grep -q 'cleanup: ' "$C/drv6.out" 2>/dev/null; then prob "$1: cleanup reported [$(grep 'cleanup: ' "$C/drv6.out" | head -1)]"; fi
  elif grep -qF "] cleanup: image $i $4" "$C/drv6.out" 2>/dev/null; then :; else prob "$1: no [cleanup: image $i $4] record"; fi
  local r q
  r="$(vb6_calls "image rm $i")" q="$(vb6_calls "image ls -q --filter reference=$i")"
  if [[ "$r" != 1 || "$q" != 1 ]]; then prob "$1: the shim saw $r removal(s) and $q presence query(ies), want 1 and 1"; fi
}
ctl_VB6() {
  local d1 d2 d3 n
  vb6_case wait
  if [[ "$RC" != 0 ]]; then prob "wait: exit $RC, want 0"; fi
  if exists image "$VB6_IMG"; then prob "wait: the built image survived its container's teardown"; fi
  if grep -q 'cleanup: ' "$C/drv6.out" 2>/dev/null; then prob "wait: cleanup reported [$(grep 'cleanup: ' "$C/drv6.out" | head -1)]"; fi
  n="$(vb6_calls "image rm $VB6_IMG:latest")"
  if [[ "$n" != 1 ]]; then prob "wait: the image listed by both records was removed $n time(s), want once"; fi
  d1="exit $RC, image removed once after the --rm container ended (${VB6_SECS} s)"; vb6_tidy
  vb6_case timeout
  if [[ "$RC" != 1 ]]; then prob "timeout: exit $RC, want 1"; fi
  if exists image "$VB6_IMG"; then :; else prob "timeout: the image was removed although a container still used it"; fi
  if grep -q 'cleanup: containers still use images this run built after 15 s; the images are kept' "$C/drv6.out" 2>/dev/null; then :; else prob "timeout: no kept-image record in the driver output"; fi
  if (( VB6_SECS < 15 || VB6_SECS > 60 )); then prob "timeout: took ${VB6_SECS} s, want about 15"; fi
  d2="exit $RC after ${VB6_SECS} s, image kept and logged"; vb6_tidy
  vb6_case earlier-failure
  if [[ "$RC" != 3 ]]; then prob "earlier-failure: exit $RC, want 3 (the earlier status kept)"; fi
  if exists image "$VB6_IMG"; then :; else prob "earlier-failure: the image was removed although a container still used it"; fi
  d3="exit $RC kept, image kept"; vb6_tidy
  vb6_case gone SHIM_FAIL_IMAGE_RM=gone; vb6_failed_rm gone 0 no ""; vb6_tidy
  vb6_case kept SHIM_FAIL_IMAGE_RM=refuse; vb6_failed_rm kept 1 yes "could not be removed (kept)"; vb6_tidy
  vb6_case unknown SHIM_FAIL_IMAGE_RM=refuse SHIM_FAIL_IMAGE_REF=1
  vb6_failed_rm unknown 1 yes "could not be removed and Docker could not confirm it is gone (unverified)"; vb6_tidy
  vb6_case unknown-earlier SHIM_FAIL_IMAGE_RM=refuse SHIM_FAIL_IMAGE_REF=1
  vb6_failed_rm unknown-earlier 3 yes "could not be removed and Docker could not confirm it is gone (unverified)"; vb6_tidy
  detail "real cleanup(): wait: $d1; timeout: $d2; earlier failure: $d3; failed removal (shim): image gone exit 0," \
    "still present exit 1 kept, presence query failing exit 1 unverified, the same after exit 3 keeps 3"
}
ctl_VB5() {
  local id r1 r2 n l
  vb_launch join; await "$VPID" 180; r1="$RC"
  if [[ "$RC" != 0 ]]; then prob "driver exit $RC, want 0"; fi
  for id in B-DEPS-CONTROLS B-WRAPPER-CONTROLS B-CI-CONTROLS; do
    if [[ "$(awk -F'\t' -v i="$id" '$1 == i && $2 == "PASS"' "$C/work/results.tsv" 2>/dev/null | grep -c .)" != 1 ]]; then prob "$id was not merged exactly once"; fi
  done
  vb_row INFO '^controls joined [(]exit 0, [0-9]+ s[)]$'; vb_gone
  vb_launch join-foreign; await "$VPID" 180; r2="$RC"
  if [[ "$RC" != 1 ]]; then prob "foreign variant: driver exit $RC, want 1"; fi
  vb_row FAIL 'foreign or malformed rows'; vb_gone
  # G064 correction 11: the child's status no longer held by bash (wait 127) or its PID naming another live child.
  vb_launch join-lost; await "$VPID" 180
  if [[ "$RC" != 0 ]]; then prob "lost-status variant: driver exit $RC, want 0"; fi
  while IFS= read -r l; do prob "lost-status variant: ${l#bad }"; done < <(grep '^bad ' "$C/work/lost.txt" 2>/dev/null)
  n="$(grep -c '^ok ' "$C/work/lost.txt" 2>/dev/null)" || n=0
  if [[ "$n" != 11 ]]; then prob "lost-status variant: $n of 11 cases ok"; fi
  vb_gone
  detail "normal join: exit $r1, the 3 controls rows merged once each, BACKGROUND INFO joined; a foreign B-SHELL row: exit $r2, BACKGROUND FAIL;" \
    "status lost or PID reused (exit $RC, $n/11 cases): exit 0 joined from the child's record, exit 3 named, no or invalid record," \
    "TERMed or running child FAIL (never 127 or success), TERM before join exited 143, a reused PID never waited for," \
    "a reused PID's saved status differing from the record FAIL naming both"
}
# VB7 (G065 K3; AC9's forced persistence-start failure, inside the hosted profile). a: persist() of this checkout's run.sh,
# extracted verbatim, in a strict-mode shell sourcing the template's lib.sh; Compose, the migrate overlay and the README
# blocks are stubbed to fail (platform:start passes in a2). It is called as `if persist`, the error-handling context of its
# caller (variants_ops: `persist || ...`): this proves the explicit failure return and the ledger row, not an errexit
# guarantee. b: the driver's witness-abort mode: the real res_witness_start on a registry of this case, then exit 2 in
# run.sh's EXIT order; W removed by exact ID and marked stopped, no latch, no leftover; and run.sh's EXIT trap names
# res_witness_abort after bg_stop_all and before cleanup. The real witness is required: a witness that does not start is
# a problem of this control, never replaced by another proof.
vb7_persist() { # CASE START_RC -> $C/vb7-CASE/{persist.sh,results.tsv,persist.rc,out}; VB7_RC: the strict-mode shell's exit
  local d="$C/vb7-$1"
  mkdir -p "$d"
  awk '/^persist\(\) \{/ {f = 1} f {print} f && /^}/ {exit}' "$REPO_R/infra/validation/run.sh" >"$d/persist.sh"
  VB7_RC=0
  bash -c 'set -euo pipefail; source "$1"; source "$2"; WORK="$3" RESULTS="$3/results.tsv" START_RC="$4"; : >"$RESULTS"
    compose() { return 1; }
    compose_migrate() { echo "vb7 stub: migrate status failed"; return 1; }
    run_docs_block() { if [[ "$3" == platform:start ]]; then return "$START_RC"; fi; return 1; }
    if persist "$3/dest"; then echo 0 >"$3/persist.rc"; else echo "$?" >"$3/persist.rc"; fi' \
    _ "$C/repo/infra/validation/lib.sh" "$d/persist.sh" "$d" "$2" >"$d/out" 2>&1 || VB7_RC=$?
}
vb7_rows() { # CASE DIR TEXT: DIR/results.tsv holds exactly one B-PERSIST row, FAIL, its text matching TEXT
  local rows
  rows="$(awk -F'\t' '$1 == "B-PERSIST" {printf "%s%s %s", s, $2, $3; s = "; "}' "$2/results.tsv" 2>/dev/null)"
  if awk -F'\t' -v t="$3" '$1 == "B-PERSIST" {n++; if ($2 == "FAIL" && $3 ~ t) m++} END {exit !(n == 1 && m == 1)}' "$2/results.tsv" 2>/dev/null; then :
  else prob "$1: B-PERSIST rows [${rows:-none}], want exactly one FAIL matching [$3]"; fi
}
ctl_VB7() {
  local d r1 r2 w o p s n t
  new_case; ctl_labels
  # a: platform:start fails: persist returns 1 with the restart's FAIL row; the shell is not ended by errexit.
  vb7_persist start-fails 1; d="$C/vb7-start-fails"
  if [[ ! -s "$d/persist.sh" ]]; then prob "a: no persist() found in infra/validation/run.sh"; fi
  if [[ "$VB7_RC" != 0 ]]; then prob "a: the strict-mode shell ended with $VB7_RC: $(tail -2 "$d/out" 2>/dev/null | tr '\n' ' ')"; fi
  r1="$(cat "$d/persist.rc" 2>/dev/null)" || r1=none
  if [[ "$r1" != 1 ]]; then prob "a: persist returned $r1 after a failed platform:start, want 1"; fi
  vb7_rows a "$d" '^restart: platform:start failed'
  # a2: platform:start passes and the migrate status fails: one FAIL row, the history named lost; persist returns 0.
  vb7_persist migrate-fails 0; d="$C/vb7-migrate-fails"
  if [[ "$VB7_RC" != 0 ]]; then prob "a2: the strict-mode shell ended with $VB7_RC: $(tail -2 "$d/out" 2>/dev/null | tr '\n' ' ')"; fi
  r2="$(cat "$d/persist.rc" 2>/dev/null)" || r2=none
  if [[ "$r2" != 0 ]]; then prob "a2: persist returned $r2 after a failed migrate status, want 0 (its FAIL row recorded)"; fi
  vb7_rows a2 "$d" '^persistence behaviour$'
  if grep -q 'migration history lost' "$d/out" 2>/dev/null; then :; else prob "a2: no 'migration history lost' log line"; fi
  # b: the real witness on this case's registry, then the abort (exit 2) before res_witness_stop.
  vb_driver; mkdir -p "$C/work"
  WORK="$(realpath -e "$C/work")"; mk_reg "$C/reg" "vb7-$RT"
  launch "$C/drv.out" env -i -C "$C" "${BASE_ENV[@]}" TMPDIR="$C/tmp" SHIM_LABELS="$(case_labels)" VB_REG="$C/reg/owned.tsv" \
    bash "$C/driver.sh" "$C/repo" "$C/work" witness-abort "$(case_labels)"
  VPID="$LPID"; await "$VPID" 180
  if [[ "$RC" != 2 ]]; then prob "b: driver exit $RC, want 2"; fi
  if grep -q 'G064-RES witness [0-9a-f]* live' "$C/drv.out" 2>/dev/null; then :; else
    prob "b: the real witness did not start: $(grep -m2 -e 'G064-RES' -e 'RESOURCES' "$C/drv.out" 2>/dev/null | tr '\n' ' ')"; fi
  w="$(cat "$C/work/res-witness/w.id" 2>/dev/null)" || w=""
  if [[ ! "$w" =~ ^[0-9a-f]{64}$ ]]; then prob "b: no witness ID recorded (res-witness/w.id)"
  elif o="$(D inspect --type container -f '{{.Id}}' "$w" 2>&1)"; then prob "b: the witness ${w:0:12} still exists after the abort"
  elif grep -qi 'no such' <<<"$o"; then :; else prob "b: the witness ${w:0:12} could not be verified gone ($(head -1 <<<"$o"))"; fi
  if [[ -e "$C/work/res-witness/stopped" ]]; then :; else prob "b: the witness was not marked stopped"; fi
  if compgen -G "$C/reg/hold.d/*" >/dev/null; then
    prob "b: a resource latch was set ($(cat "$C/reg/hold.d/"* 2>/dev/null | awk -F'\t' '$1 == "reason" {printf "%s%s", s, $2; s = "; "}'))"; fi
  { read -r p s <"$C/work/res-witness/events.pid"; } 2>/dev/null || p=""
  if [[ -n "$p" ]] && proc_is "$p" "$s"; then prob "b: the witness event stream (pid $p) survived the abort"; fi
  vb_row FAIL '^res-witness stopped before join'
  if awk -F'\t' '$1 == "BACKGROUND" && $3 ~ /^res-witness survived KILL/ {f = 1} END {exit !f}' "$C/work/results.tsv" 2>/dev/null; then prob "b: the witness parser survived KILL"; fi
  vb_gone
  # The production EXIT trap (static): res_witness_abort after bg_stop_all, before cleanup.
  n="$(grep -c "^trap 'RUN_STATUS=" "$REPO_R/infra/validation/run.sh")" || n=0
  t="$(grep -m1 "^trap 'RUN_STATUS=" "$REPO_R/infra/validation/run.sh")" || t=""
  case "$n:$t" in
    1:*"bg_stop_all; res_witness_abort "*"; cleanup' EXIT") ;;
    *) prob "run.sh's EXIT trap does not name res_witness_abort after bg_stop_all and before cleanup ($n trap lines)" ;;
  esac
  detail "a: persist (run.sh, strict mode, stubbed stack) after a failed platform:start returned $r1 with one B-PERSIST FAIL (restart) row;" \
    "after a failed migrate status returned $r2 with one FAIL row (migration history lost); b: the real witness ${w:0:12}," \
    "then exit $RC: BACKGROUND FAIL res-witness stopped before join, W removed by exact ID and marked stopped, no latch," \
    "no labelled leftover; run.sh's EXIT trap orders bg_stop_all, res_witness_abort, cleanup"
}
# VB1 and VB3b: the real run.sh of a scratch copy of this checkout (V4's copy), its validate-controls.sh a stub that
# prints its PASS line and then kills (VB_STUB=kill) the background controls child (the topmost ancestor below
# run.sh's group leader) or holds there (hold); run-approved-controls.sh exits 0. The later suite is staging, which
# without IDENTITY_STAGING_ENV_FILE records UNMET at once.
vb_copy() { # [fast-profile]: $C/impl, committed; fast-profile also stubs runner-preflight.sh and sets CI_SUITES_fast
  local x imp="$C/impl"
  mkdir -p "$imp" "$C/vbs"
  (cd "$REPO_R" && git ls-files -z --cached --others --exclude-standard | while IFS= read -r -d '' x; do if [[ -f "$x" || -L "$x" ]]; then printf '%s\0' "$x"; fi; done | tar --null --no-recursion -cf - -T -) | tar -xf - -C "$imp"
  cat >"$imp/infra/ci/validate-controls.sh" <<'VBSTUB'
#!/usr/bin/env bash
# VB stub of infra/ci/validate-controls.sh (validate-controls.sh VB1/VB3b).
set -euo pipefail
if [[ "${1:-}" == --list ]]; then printf 'A V1\n'; exit 0; fi
printf 'PASS V1 stub\n'
m="${VB_STUB_DIR:?}" pg="$(ps -o pgid= -p "$$" | tr -d ' ')" p="$PPID" pp=""
while :; do
  pp="$(ps -o ppid= -p "$p" | tr -d ' ')" || pp=""
  if [[ -z "$pp" || "$pp" == "$pg" || "$pp" -le 1 ]]; then break; fi
  p="$pp"
done
printf '%s %s %s\n' "$p" "$pp" "$pg" >"$m/stub-chain"
if [[ "$pp" != "$pg" || "$p" == "$pg" ]]; then : >"$m/stub-not-background"; exit 1; fi
printf '%s\n' "$p" >"$m/child.pid"
if [[ "${VB_STUB:-hold}" == kill ]]; then kill -KILL "$p"; exit 0; fi
: >"$m/holding"
sleep 600
VBSTUB
  printf '#!/usr/bin/env bash\n# VB stub: the wrapper controls are not the subject of VB1/VB3b.\nexit 0\n' >"$imp/infra/validation/run-approved-controls.sh"
  if [[ "${1:-}" == fast-profile ]]; then
    cp -- "$TEMPLATE/infra/ci/runner-preflight.sh" "$imp/infra/ci/runner-preflight.sh"
    sed -i 's/^CI_SUITES_fast=.*/CI_SUITES_fast="ci-controls staging"/' "$imp/infra/ci/lib.sh"
    if grep -qx 'CI_SUITES_fast="ci-controls staging"' "$imp/infra/ci/lib.sh"; then :; else prob "the scratch fast profile could not be set"; return 1; fi
  fi
  git -C "$imp" init -q -b main
  git -C "$imp" add -A
  git -C "$imp" -c user.name=ci-controls -c user.email=controls@example.invalid commit -q -m "VB scratch copy"
}
vb_no_survivor() {
  local p
  p="$(cat "$C/vbs/child.pid" 2>/dev/null)" || p=""
  if [[ -z "$p" ]]; then prob "the stub found no background controls child ($(cat "$C/vbs/stub-chain" 2>/dev/null))"; elif kill -0 "$p" 2>/dev/null; then prob "the controls child $p is alive"; fi
  if [[ -e "$C/vbs/stub-not-background" ]]; then prob "the controls suite did not run as a background child of run.sh"; fi
  if ps -eo args= | grep -F -- "$C/" | grep -vq -- 'grep -F'; then prob "a process of this case is still running"; fi
}
ctl_VB1() {
  local r1 rep
  new_case; ctl_labels; rep="$C/vb1-report"
  vb_copy || return 0
  launch "$C/vb1.log" env -i -C "$C/impl" "${BASE_ENV[@]}" TMPDIR="$C/tmp" SHIM_LABELS="$(case_labels)" VB_STUB=kill VB_STUB_DIR="$C/vbs" \
    bash "$C/impl/infra/validation/run.sh" --base HEAD --allow-file "$C/impl/infra/ci/unchanged.allow" --report-dir "$rep" ci-controls staging
  await "$LPID" 900; r1="$RC"
  if [[ "$RC" != 1 ]]; then prob "run.sh exit $RC, want 1"; fi
  if awk -F'\t' '$1 == "BACKGROUND" && $2 == "FAIL" && $3 ~ /^ci-controls exited/ {f = 1} END {exit !f}' "$rep/results.tsv" 2>/dev/null; then :; else prob "results.tsv lacks BACKGROUND FAIL ci-controls exited"; fi
  if awk -F'\t' '$1 == "B-CI-CONTROLS" {f = 1} END {exit !f}' "$rep/results.tsv" 2>/dev/null; then prob "a B-CI-CONTROLS row exists although the child was killed first"; fi
  vb_no_survivor
  # The killed child's result through validate.sh: harness 1, the BACKGROUND FAIL row published, not successful.
  cp -- "$rep/results.tsv" "$C/replay.tsv" 2>/dev/null || { prob "no results.tsv to replay"; return 0; }
  ci_env 4330; run_v "${CIENV[@]}" STUB_HARNESS=replay STUB_REPLAY="$C/replay.tsv" -- --publish-dir ci-evidence
  expect_rc 1; expect_line has harness=1; expect_pub full
  if grep -q '"harness": 1' "$(PUBDIR)/summary.json" 2>/dev/null; then :; else prob "summary.json does not record harness 1"; fi
  if grep -qP '^BACKGROUND\tFAIL\t' "$(PUBDIR)/results.tsv" 2>/dev/null; then :; else prob "the published results lack the BACKGROUND FAIL row"; fi
  detail "real run.sh ci-controls staging, the CI-controls child KILLed after its PASS line: exit $r1, BACKGROUND FAIL, no B-CI-CONTROLS row, no live child; through validate.sh: exit $RC, summary harness 1"
}
ctl_VB3b() {
  local i w s
  new_case; ctl_labels
  vb_copy fast-profile || return 0
  case_state
  launch "$C/out" env -i -C "$C/impl" "${BASE_ENV[@]}" G003_STATE_ROOT="$CSTATE" TMPDIR="$C/tmp" STUB_DIR="$C/stub" SHIM_LOG="$C/shim.log" SHIM_PULL_NOOP=1 \
    SHIM_LABELS="$(case_labels)" VB_STUB=hold VB_STUB_DIR="$C/vbs" "${NOCAP[@]}" bash "$C/impl/infra/ci/validate.sh" --profile fast
  VPID="$LPID"
  if wait_file "$C/vbs/holding" 300; then
    s=""
    for ((i = 0; i < 100; i++)); do
      w="$(ps -eo pid=,ppid=,args= | awk -v p="$VPID" '$2 == p && / watchdog / {print $1}')" || w=""
      s="$(ps -eo pid=,ppid=,args= | awk -v p="${w:-none}" '$2 == p && $3 == "sleep" {print $1}')" || s=""
      if [[ -n "$s" ]]; then break; fi
      sleep 0.1
    done
    if [[ -n "$s" ]]; then kill -TERM "$s"; else prob "watchdog not found"; kill -TERM "$VPID"; fi
  else prob "the stub never held (run.sh did not reach the background controls)"; kill -TERM "$VPID"; fi
  await "$VPID" 400; parse_final
  expect_rc 124; expect_stop deadline; expect_line has harness=124
  if [[ "$F_PUB" != none ]]; then prob "published=$F_PUB, want none"; fi
  vb_no_survivor
  detail "validate.sh deadline while the background controls child holds: exit $RC (harness 124, harness_stop deadline), nothing published, no survivor"
}
# VB8 (G064 corr8; RC17 `RESOURCES FAIL B-MIG: ceiling network 2/1 (compose up)`): the real _res_live of the template's
# lib.sh on fixture objects, through a fake docker on PATH that renders the caller's --format (no daemon, nothing created).
# fix: the run's foreground network, volume and container (a registered project; the container also g001.run) beside a
# background control's network, volume and container (sdlc.ci.run; mirrored as gate_record does: bare `network|volume
# <id>` rows, `standalone <id> ci-V19`) and an unlabelled default network: one foreground ID per class, a run-wide sample
# of 2. nolabel: the same rows without the sdlc.ci.run label: the bare row alone keeps a network or volume foreground (2),
# the count the pre-fix rule gave the labelled control objects (B-MIG's network ceiling is 1).
vb8_driver() { # write $C/driver8.sh and the fake docker $C/work/fake/docker
  mkdir -p "$C/work/fake"
  cat >"$C/work/fake/docker" <<'VB8FAKE'
#!/usr/bin/env bash
# VB8 fake docker of infra/ci/validate-controls.sh: `network ls`, `volume ls` and `ps -a` render the caller's --format for
# the objects of $VB8_FIX/<networks|volumes|containers> (ID|labels); any other call is refused. Each call goes to $VB8_LOG.
set -euo pipefail
printf '%s\n' "$*" >>"$VB8_LOG"
fmt="" a=0 pi='{{.ID}}' pn='{{.Name}}' re='[{][{][.]Label "([^"]*)"[}][}]'
for x in "$@"; do if (( a )); then fmt="$x"; a=0; elif [[ "$x" == --format ]]; then a=1; fi; done
case "${1:-} ${2:-}" in
  "network ls") f="$VB8_FIX/networks" ;;
  "volume ls") f="$VB8_FIX/volumes" ;;
  "ps -a") f="$VB8_FIX/containers" ;;
  *) echo "vb8 fake docker: refused: $*" >&2; exit 1 ;;
esac
while IFS='|' read -r id labels; do
  l="${fmt//"$pi"/$id}"; l="${l//"$pn"/$id}"
  while [[ "$l" =~ $re ]]; do
    t="${BASH_REMATCH[0]}" k="${BASH_REMATCH[1]}" v=""
    for kv in $labels; do if [[ "${kv%%=*}" == "$k" ]]; then v="${kv#*=}"; fi; done
    l="${l//"$t"/$v}"
  done
  printf '%s\n' "${l//\\t/$'\t'}"
done <"$f"
VB8FAKE
  chmod 755 "$C/work/fake/docker"
  cat >"$C/driver8.sh" <<'VB8DRV'
#!/usr/bin/env bash
# VB8 driver of infra/ci/validate-controls.sh: the real _res_live per fixture set and class, the fake docker first on PATH.
#   driver8.sh REPO WORK   -> WORK/live.tsv: set, class, RES_FGN, RES_ALLN, RES_IDS (or the query failure)
set -euo pipefail
REPO="$1" WORK="$2"
# shellcheck source=/dev/null
source "$REPO/infra/validation/lib.sh"
PATH="$WORK/fake:$PATH" RUN_ID=vb8run OWNED_REG="$WORK/reg/owned.tsv"
export VB8_LOG="$WORK/fake.log"
_res_d
: >"$WORK/live.tsv"
for set in fix nolabel; do
  export VB8_FIX="$WORK/$set"
  for cls in network volume container; do
    if _res_live "$cls"; then printf '%s\t%s\t%s\t%s\t%s\n' "$set" "$cls" "$RES_FGN" "$RES_ALLN" "$(paste -sd ' ' <<<"$RES_IDS")" >>"$WORK/live.tsv"
    else printf '%s\t%s\tfailed\t-\t%s\n' "$set" "$cls" "$RES_WHY" >>"$WORK/live.tsv"; fi
  done
done
VB8DRV
}
vb8_hex() { printf '%064d' 0 | tr 0 "$1"; } # DIGIT -> a 64-character fixture ID
ctl_VB8() {
  local fn cn dn fc cc fv=g001-vb8_pgdata cv=g003ctl-vb8-v19-vol ctl f want got s c n q
  new_case; vb8_driver; mkdir -p "$C/work/reg" "$C/work/fix" "$C/work/nolabel"
  fn="$(vb8_hex 1)" cn="$(vb8_hex 2)" dn="$(vb8_hex 3)" fc="$(vb8_hex 4)" cc="$(vb8_hex 5)"
  ctl="com.docker.compose.project=g003ctl-vb8-v19 sdlc.ci.control=vb8ctl-V19 sdlc.ci.run=vb8ctl sdlc.ci.owner=vb8owner"
  printf 'g001-vb8\n' >"$C/work/projects"
  printf 'network\t%s\nnetwork\t%s\nvolume\t%s\nstandalone\t%s\tci-V19\n' "$fn" "$cn" "$cv" "$cc" >"$C/work/reg/owned.tsv"
  printf '%s|com.docker.compose.project=g001-vb8\n%s|%s\n%s|\n' "$fn" "$cn" "$ctl" "$dn" >"$C/work/fix/networks"
  printf '%s|com.docker.compose.project=g001-vb8\n%s|%s\n' "$fv" "$cv" "$ctl" >"$C/work/fix/volumes"
  printf '%s|com.docker.compose.project=g001-vb8 g001.run=vb8run\n%s|%s\n' "$fc" "$cc" "$ctl" >"$C/work/fix/containers"
  for f in networks volumes containers; do sed 's/ sdlc\.ci\.run=vb8ctl//' "$C/work/fix/$f" >"$C/work/nolabel/$f"; done
  launch "$C/drv8.out" env -i -C "$C" "${BASE_ENV[@]}" TMPDIR="$C/tmp" bash "$C/driver8.sh" "$C/repo" "$C/work"
  await "$LPID" 60
  if [[ "$RC" != 0 ]]; then prob "driver exit $RC: $(tail -2 "$C/drv8.out" 2>/dev/null | tr '\n' ' ')"; fi
  for want in "fix network 1 2 $fn" "fix volume 1 2 $fv" "fix container 1 2 $fc" \
    "nolabel network 2 2 $fn $cn" "nolabel volume 2 2 $fv $cv" "nolabel container 1 2 $fc"; do
    read -r s c _ <<<"$want"
    got="$(awk -F'\t' -v s="$s" -v c="$c" '$1 == s && $2 == c {print $1, $2, $3, $4, $5}' "$C/work/live.tsv" 2>/dev/null)" || got=""
    if [[ "$got" != "$want" ]]; then prob "$s $c: [${got:-no row}], want [$want] (set class RES_FGN RES_ALLN RES_IDS)"; fi
  done
  n="$(grep -c . "$C/work/fake.log" 2>/dev/null)" || n=0
  q="$(grep -E '^(network|volume) ls ' "$C/work/fake.log" 2>/dev/null | grep -cF '{{.Label "sdlc.ci.run"}}')" || q=0
  if [[ "$n" != 6 || "$q" != 4 ]]; then prob "the fake docker saw $n calls ($q network/volume queries asking sdlc.ci.run), want 6 (4)"; fi
  detail "real _res_live, fake docker ($n queries; every network/volume query asks sdlc.ci.run): a background control's network" \
    "and volume (bare registry rows, sdlc.ci.run) and container (ci- row) are run-wide only: foreground/run-wide network 1/2," \
    "volume 1/2, container 1/2 (unchanged); without the label the bare rows keep the network and volume foreground (2/2):" \
    "the pre-fix rule counted the labelled control objects so, 2 against B-MIG's network ceiling 1 (RC17)"
}

# --- group E: owned-resource cleanup (Docker-real) ---------------------------------------------------------
REG="" WORK="" PROJ="" JOB=""
e_case() { # a scratch tree with cleanup-owned.sh and a local registry skeleton; start inventory taken now
  C="$ROOT/cases/$CTL_ID"
  rm -rf -- "$C"
  mkdir -p "$C/repo/infra/ci" "$C/tmp" "$C/work"
  cp -- "$TEMPLATE/infra/ci/cleanup-owned.sh" "$TEMPLATE/infra/ci/lib.sh" "$C/repo/infra/ci/"
  WORK="$(realpath -e "$C/work")"
  REG="$C/reg"
  JOB="local-$(ci_random_hex 6)"
  PROJ="g003ctl-$RT-$(tr 'A-Z' 'a-z' <<<"$CTL_ID")"
  mk_reg "$REG" "$JOB"
}
mk_reg() { # DIR JOB: registry skeleton plus a fresh start inventory
  local tok
  mkdir -m 700 "$1" "$1/cid"
  tok="$(ci_random_hex 16)"
  printf '%s\n' "$tok" >"$1/token"
  printf '%s\njob\t%s\ntoken\t%s\nrepo\t%s\nwork_root\t%s\n' "$CI_REGISTRY_HEADER" "$2" "$tok" "$(realpath -e "$C/repo")" "$WORK" >"$1/owned.tsv"
  ci_docker_start_inventory "$1/start-inventory.ids"
}
row() { local IFS=$'\t'; printf '%s\n' "$*" >>"$REG/owned.tsv"; }
ctl_labels() { LB=(--label "sdlc.ci.control=$RT-$CTL_ID" --label "$RUN_LABEL" --label "$OWNER_LABEL"); local e; for e in "${EXTRA_LABELS[@]}"; do LB+=(--label "$e"); done; }
mk_image() { # TAG -> a new image ID made from busybox (no build, no network)
  local cid id
  cid="$(D create "${LB[@]}" --network none "$CI_BUSYBOX_IMAGE" true)"
  id="$(D commit --change "LABEL sdlc.ci.control=$RT-$CTL_ID" --change "LABEL $RUN_LABEL" --change "LABEL $OWNER_LABEL" "$cid" "$1")"
  D rm -f "$cid" >/dev/null
  printf '%s' "$id"
}
compose_ctr() { # PROJECT WORKING_DIR [docker run args...] -> a detached Compose-labelled container
  local p="$1" wd="$2"
  shift 2
  D run -d "${LB[@]}" --label "com.docker.compose.project=$p" --label "com.docker.compose.project.working_dir=$wd" "$@"
}
run_cleanup() { # ENV...: finalize mode
  RC=0
  env -i -C "$C/repo" "${BASE_ENV[@]}" TMPDIR="$C/tmp" SHIM_LABELS="$(case_labels)" "$@" \
    bash "$C/repo/infra/ci/cleanup-owned.sh" --registry "${CLEAN_REG:-$REG}" --mode finalize --out "$C/inv.tsv" >"$C/cleanup.out" 2>&1 || RC=$?
}
inv_has() { awk -F'\t' -v k="$1" -v i="$2" -v a="$3" '$1 == k && index($2, i) && $3 == a {f = 1} END {exit !f}' "$C/inv.tsv"; }
inv_count() { awk -F'\t' -v a="$1" '$3 == a {n++} END {print n + 0}' "$C/inv.tsv"; }
exists() { D inspect --type "$1" "$2" >/dev/null 2>&1; }
want_inv() { if inv_has "$@"; then :; else prob "inventory lacks $1 ${2:0:12} $3"; fi; }
want_gone() { if exists "$1" "$2"; then prob "$1 ${2:0:12} still exists"; fi; }
want_kept() { if exists "$1" "$2"; then :; else prob "$1 ${2:0:12} was removed"; fi; }
want_rc() { if [[ "$RC" != "$1" ]]; then prob "cleanup exit $RC, want $1: $(tail -1 "$C/cleanup.out" 2>/dev/null)"; fi; }

ctl_V19() {
  local img net vol ctr
  e_case; ctl_labels
  img="$(mk_image "$PROJ-core:latest")"
  net="$(D network create --internal "${LB[@]}" --label "com.docker.compose.project=$PROJ" "$PROJ-net")"
  vol="$(D volume create "${LB[@]}" --label "com.docker.compose.project=$PROJ" "$PROJ-vol")"
  ctr="$(compose_ctr "$PROJ" "$WORK/copy/infra/compose" --network "$net" -v "$vol:/data" "$PROJ-core:latest" sleep 300)"
  row project "$PROJ"; row image "$PROJ-core:latest" "$img"; row container "$ctr"; row network "$net"; row volume "$vol"
  run_cleanup
  want_rc 0
  want_inv container "$ctr" removed; want_inv network "$net" removed; want_inv volume "$vol" removed; want_inv image "$PROJ-core:latest" removed
  if [[ "$(inv_count removed)" != 4 || "$(grep -c . "$C/inv.tsv")" != 4 ]]; then prob "inventory is not exactly 4 removed rows"; fi
  want_gone container "$ctr"; want_gone network "$net"; want_gone volume "$vol"; want_gone image "$img"
  detail "exit $RC; $(tail -1 "$C/cleanup.out")"
}
ctl_V20c() {
  local ctr
  e_case; ctl_labels
  ctr="$(compose_ctr "$PROJ" /nonexistent/other-checkout/infra/compose --network none "$CI_BUSYBOX_IMAGE" sleep 300)"
  row project "$PROJ"
  run_cleanup; want_rc 18; want_inv container "$ctr" refused; want_kept container "$ctr"
  detail "exit $RC; foreign working_dir refused and kept"
}
ctl_V20n() {
  local net
  e_case; ctl_labels
  net="$(D network create --internal "${LB[@]}" --label "com.docker.compose.project=$PROJ" "$PROJ-net")"
  row project "$PROJ"
  run_cleanup; want_rc 18; want_inv network "$net" unproven; want_kept network "$net"
  detail "exit $RC; unattached network unproven and kept"
}
ctl_V20v() {
  local vol
  e_case; ctl_labels
  vol="$(D volume create "${LB[@]}" --label "com.docker.compose.project=$PROJ" "$PROJ-vol")"
  row project "$PROJ"
  run_cleanup; want_rc 18; want_inv volume "$vol" unproven; want_kept volume "$vol"
  detail "exit $RC; unmounted volume unproven and kept"
}
ctl_V20i() {
  local img
  e_case; ctl_labels
  img="$(mk_image "$PROJ-intruder:latest")"
  row project "$PROJ"
  run_cleanup; want_rc 18; want_inv image "$PROJ-intruder:latest" unproven; want_kept image "$img"
  detail "exit $RC; registered-project image name without a row unproven and kept"
}
ctl_V21() {
  local q ctr net vol
  e_case; ctl_labels
  q="g003ctl-$RT-other"
  net="$(D network create --internal "${LB[@]}" --label "com.docker.compose.project=$q" "$q-net")"
  vol="$(D volume create "${LB[@]}" --label "com.docker.compose.project=$q" "$q-vol")"
  ctr="$(compose_ctr "$q" "$WORK/copy/infra/compose" --network "$net" -v "$vol:/data" "$CI_BUSYBOX_IMAGE" sleep 300)"
  row project "$PROJ"
  run_cleanup; want_rc 0
  if grep -qF "${ctr:0:12}" "$C/inv.tsv" || grep -qF "$q" "$C/inv.tsv"; then prob "an unregistered project appears in the inventory"; fi
  want_kept container "$ctr"; want_kept network "$net"; want_kept volume "$vol"
  detail "exit $RC; unregistered project untouched and absent from the inventory"
}
ctl_V22() {
  local ctr
  e_case; ctl_labels
  ctr="$(compose_ctr "$PROJ" "$WORK/copy/infra/compose" --network none "$CI_BUSYBOX_IMAGE" sleep 300)"
  row project "$PROJ"
  run_cleanup SHIM_FAIL_RM=1; want_rc 18; want_inv container "$ctr" remaining
  if [[ ! -s "$C/inv.tsv" ]]; then prob "the inventory is empty"; fi
  detail "exit $RC; docker rm failure reported as remaining"
}
ctl_V23() {
  local ctr out reg2
  e_case; ctl_labels
  ctr="$(D run -d "${LB[@]}" --network none "$CI_BUSYBOX_IMAGE" sleep 300)"
  row standalone "$ctr" controls-probe
  # (a) the registry is a symlink
  ln -s "$REG" "$C/reglink"
  CLEAN_REG="$C/reglink" run_cleanup; want_rc 18; want_inv registry "$C/reglink" refused; want_kept container "$ctr"
  # (b) token row and token file disagree
  cp -a "$REG" "$C/reg-b"; ci_random_hex 16 >"$C/reg-b/token"
  CLEAN_REG="$C/reg-b" run_cleanup; want_rc 18; want_inv registry "$C/reg-b" refused; want_kept container "$ctr"
  # (c) the job row does not match CI_JOB_ID
  CLEAN_REG="$REG" run_cleanup GITLAB_CI=true CI_JOB_ID=999; want_rc 18; want_inv registry "$REG" refused; want_kept container "$ctr"
  # (d) after_script: a symlinked registry and a job mismatch are refused, exit 0, nothing removed
  reg2="$C/.platform-validation-owned-4711"
  ln -s "$REG" "$reg2"
  RC=0; out="$(env -i PATH="$PATH" "${DOCKER_ENV[@]}" CI_PROJECT_DIR="$C/repo" CI_JOB_ID=4711 bash "$C/repo/infra/ci/cleanup-owned.sh" --after-script 2>&1)" || RC=$?
  if [[ "$RC" != 0 || "$out" != *refused* ]]; then prob "after_script symlink: exit $RC, output lacks the refusal"; fi
  rm -f "$reg2"; cp -a "$REG" "$reg2"
  RC=0; out="$(env -i PATH="$PATH" "${DOCKER_ENV[@]}" CI_PROJECT_DIR="$C/repo" CI_JOB_ID=4711 bash "$C/repo/infra/ci/cleanup-owned.sh" --after-script 2>&1)" || RC=$?
  if [[ "$RC" != 0 || "$out" != *"job mismatch"* ]]; then prob "after_script job mismatch: exit $RC, output lacks the refusal"; fi
  want_kept container "$ctr"
  detail "symlink, token and job mismatches refused (18); after_script refusals exit 0; nothing removed"
}
ctl_V24() {
  local ctr reg out
  e_case; ctl_labels
  reg="$C/.platform-validation-owned-4712"
  mk_reg "$reg" 4712
  ctr="$(D run -d --cidfile "$reg/cid/run-1.cid" "${LB[@]}" --network none "$CI_BUSYBOX_IMAGE" sleep 300)"
  printf 'SECRET-%s\n' "$(ci_random_hex 8)" >"$reg/secrets"
  RC=0; out="$(env -i PATH="$PATH" "${DOCKER_ENV[@]}" CI_PROJECT_DIR="$C/repo" CI_JOB_ID=4712 bash "$C/repo/infra/ci/cleanup-owned.sh" --after-script 2>&1)" || RC=$?
  printf '%s\n' "$out" >"$C/after.out"
  if [[ "$RC" != 0 ]]; then prob "after_script exit $RC"; fi
  if [[ "$out" != *removed=1*remaining=0* ]]; then prob "after_script did not report the removal"; fi
  want_gone container "$ctr"
  if [[ -e "$reg" ]]; then prob "the registry was not deleted after a verified cleanup"; fi
  RC=0; out="$(env -i PATH="$PATH" "${DOCKER_ENV[@]}" CI_PROJECT_DIR="$C/repo" CI_JOB_ID=4799 bash "$C/repo/infra/ci/cleanup-owned.sh" --after-script 2>&1)" || RC=$?
  if [[ "$RC" != 0 || "$out" != "cleanup: no registry (finalization completed)" ]]; then prob "no-registry case: exit $RC, output [$out]"; fi
  detail "env -i hand-off removed the bound container and the registry; no registry -> message, exit 0"
}
ctl_V25() {
  e_case; ctl_labels
  printf 'X=1\n' >"$C/repo/.env"
  row rootenv "$(sha256sum <"$C/repo/.env" | cut -c1-64)"
  run_cleanup; want_rc 0; want_inv env "$C/repo/.env" removed
  if [[ -e "$C/repo/.env" ]]; then prob "the matching .env survived"; fi
  printf 'X=2\n' >"$C/repo/.env"
  run_cleanup; want_rc 18; want_inv env "$C/repo/.env" refused
  if [[ ! -e "$C/repo/.env" ]]; then prob "the mismatched .env was removed"; fi
  detail "matching hash removed; mismatched hash refused and kept"
}
ctl_V26() {
  local ctr
  C="$ROOT/cases/$CTL_ID"; rm -rf -- "$C"; mkdir -p "$C/repo/infra/ci" "$C/tmp" "$C/work"; WORK="$(realpath -e "$C/work")"; ctl_labels
  cp -- "$TEMPLATE/infra/ci/cleanup-owned.sh" "$TEMPLATE/infra/ci/lib.sh" "$C/repo/infra/ci/"
  PROJ="g003ctl-$RT-v26"
  ctr="$(compose_ctr "$PROJ" "$WORK/copy/infra/compose" --network none "$CI_BUSYBOX_IMAGE" sleep 300)"
  REG="$C/reg"; mk_reg "$REG" "local-$(ci_random_hex 6)"
  row project "$PROJ"
  run_cleanup; want_rc 0; want_kept container "$ctr"
  if grep -qF "${ctr:0:12}" "$C/inv.tsv"; then prob "a pre-existing container appears in the inventory"; fi
  detail "exit $RC; the pre-existing registered-project container is untouched"
}
v_standalone() { # MODE [PREP]: a killed stub harness left a detached standalone container (PREP runs after case setup)
  local id
  std_ci; if [[ -n "${2:-}" ]]; then "$2"; fi
  run_v "${CIENV[@]}" STUB_HARNESS="$1" -- --publish-dir ci-evidence
  id="$(cat "$C/stub/standalone-id" 2>/dev/null)" || id=""
  expect_rc 137; expect_phase cleanup 0; expect_line has harness=137
  if [[ ! "$id" =~ ^[0-9a-f]{64}$ ]]; then prob "the stub recorded no container ID"; return 0; fi
  want_gone container "$id"
  if awk -F'\t' -v i="$id" '$1 == "container" && $2 == i && $3 == "removed" {f = 1} END {exit !f}' "$(PUBDIR)/cleanup-inventory.tsv" 2>/dev/null; then :; else prob "the published inventory lacks the removal"; fi
}
ctl_V27() {
  if [[ -f "$TEMPLATE/infra/validation/lib.sh" ]] && grep -q 'docker_run_owned()' "$TEMPLATE/infra/validation/lib.sh"; then :; else prob "prerequisite missing: docker_run_owned in infra/validation/lib.sh"; return 0; fi
  local id2 proj fresh old kept inv
  v_standalone standalone-kill v27_docs_prep
  id2="$(cat "$C/stub/standalone-id2" 2>/dev/null)" || id2=""
  if [[ "$id2" =~ ^[0-9a-f]{64}$ ]]; then want_gone container "$id2"; else prob "the second docker_run_owned call started no container"; fi
  if [[ "$(cat "$C/stub/cid-count" 2>/dev/null)" != 2 ]]; then prob "want two cidfiles for two calls"; fi
  # Documented block `run --rm --build` (2026-10-19 cleanup repair): the new image has a build row and is removed;
  # the pre-existing IDs (the tag's earlier image and an unbuilt service's image) are never recorded or removed.
  proj="$(vc_tag docs)" inv="$(PUBDIR)/cleanup-inventory.tsv"
  fresh="$(cat "$C/stub/docs-fresh-id" 2>/dev/null)" || fresh=""
  old="$(cat "$C/v27-old-id")" kept="$(cat "$C/v27-kept-id")"
  if [[ "$(cat "$C/stub/docs-rc" 2>/dev/null)" != 0 ]]; then prob "the documented block failed: $(tail -3 "$C/stub/docs-check.log" 2>/dev/null | tr '\n' ' ')"; fi
  if [[ ! "$fresh" =~ ^sha256:[0-9a-f]{64}$ || "$fresh" == "$old" ]]; then prob "the block built no new image (fresh [$fresh])"; fresh="sha256:none"; fi
  if [[ "$(cat "$C/stub/docs-rows" 2>/dev/null)" != "image"$'\t'"$proj-fresh:latest"$'\t'"$fresh" ]]; then prob "image rows [$(tr '\t\n' ' ;' <"$C/stub/docs-rows" 2>/dev/null)], want exactly the fresh build row"; fi
  # The tag's earlier image may go with the retag itself (containerd image store), so only the kept one must remain.
  want_gone image "$fresh"; want_kept image "$kept"
  if awk -F'\t' -v i="$proj-fresh:latest|$fresh" '$1 == "image" && $2 == i && $3 == "removed" {f = 1} END {exit !f}' "$inv" 2>/dev/null; then :; else prob "the published inventory lacks the fresh image removal"; fi
  if grep -qF -e "${old#sha256:}" -e "${kept#sha256:}" "$inv" 2>/dev/null; then prob "a pre-existing image ID appears in the published inventory"; fi
  detail "exit $RC (harness KILLed); two docker_run_owned containers (each via \$(...)) removed through their cidfile rows; a run --rm --build documented block's new image recorded by build and removed, the pre-existing image IDs never recorded or removed by cleanup (the unbuilt service image kept); cleanup=0"
}
v27_docs_prep() { # the documented-block fixture of V27; the tag's earlier image and the kept image pre-exist the run
  local d="$C/stub/docs" proj i n s
  ctl_labels; proj="$(vc_tag docs)" n="$(ci_random_hex 8)"
  mkdir -p "$d/infra/compose" "$d/fresh" "$d/kept"
  printf 'COMPOSE_PROJECT_NAME=%s\n' "$proj" >"$d/.env"
  printf '%s-fresh:latest\n' "$proj" >"$d/fresh-tag"
  printf 'FROM busybox:1.37.0\nLABEL g003.v27=new-%s\n' "$n" >"$d/fresh/Dockerfile"
  printf 'FROM busybox:1.37.0\nLABEL g003.v27=kept-%s\n' "$n" >"$d/kept/Dockerfile"
  {
    printf 'services:\n'
    for s in fresh kept; do
      printf '  %s:\n    build:\n      context: ../../%s\n      labels:\n' "$s" "$s"
      for i in $(case_labels); do printf '        %s: "%s"\n' "${i%%=*}" "${i#*=}"; done
      printf '    network_mode: none\n    command: ["true"]\n'
    done
  } >"$d/infra/compose/local.yml"
  printf '<!-- stub:check -->\n```sh\ndocker compose --env-file .env -f infra/compose/local.yml run --rm --build fresh\n```\n' >"$d/README.md"
  D build -q "${LB[@]}" -t "$proj-kept:latest" "$d/kept" >"$C/v27-kept-id"
  printf 'FROM busybox:1.37.0\nLABEL g003.v27=old-%s\n' "$n" | D build -q "${LB[@]}" -t "$proj-fresh:latest" - >"$C/v27-old-id"
}
ctl_V27b() { v_standalone cidfile-kill; detail "exit $RC; container bound only by its cidfile removed; cleanup=0"; }
ctl_V28() {
  local id out
  std_ci; start_v "${CIENV[@]}" STUB_PREFLIGHT=rp6-hold -- --publish-dir ci-evidence
  if wait_file "$C/stub/preflight-created" 60; then kill -TERM "$VPID"; else prob "the stub preflight never created its container"; fi
  wait_v
  id="$(cat "$C/stub/held-id" 2>/dev/null)" || id=""
  expect_rc 143; expect_line has signal=143; expect_phase cleanup 0; expect_runs 0
  if [[ "$id" =~ ^[0-9a-f]{64}$ ]]; then want_gone container "$id"; else prob "no preflight container ID"; fi
  # SIGKILL of the whole job: only after_script can clean up.
  CTL_ID="$CTL_ID-kill"; std_ci 4801; CTL_ID="${CTL_ID%-kill}"
  start_v "${CIENV[@]}" STUB_PREFLIGHT=rp6-hold -- --publish-dir ci-evidence
  if wait_file "$C/stub/preflight-created" 60; then kill -KILL -- "-$VPID"; else prob "the stub preflight never created its container (kill case)"; fi
  wait "$VPID" 2>/dev/null || true
  id="$(cat "$C/stub/held-id" 2>/dev/null)" || id=""
  RC=0; out="$(env -i PATH="$PATH" "${DOCKER_ENV[@]}" CI_PROJECT_DIR="$C/repo" CI_JOB_ID=4801 bash "$C/repo/infra/ci/cleanup-owned.sh" --after-script 2>&1)" || RC=$?
  printf '%s\n' "$out" >"$C/after.out"
  if [[ "$RC" != 0 ]]; then prob "after_script exit $RC"; fi
  if [[ "$id" =~ ^[0-9a-f]{64}$ ]]; then want_gone container "$id"; else prob "no preflight container ID (kill case)"; fi
  if [[ -e "$C/.platform-validation-owned-4801" ]]; then prob "after_script kept a verified registry"; fi
  detail "TERM: finalize removed the RP6-style container (exit 143); KILL: after_script removed it through the cidfile row"
}
label_only() { # LABEL-ROW: a stopped container carrying a registered label, with no binding
  local ctr
  e_case; ctl_labels
  row label "$1"
  ctr="$(D create "${LB[@]}" --label "$1" --network none "$CI_BUSYBOX_IMAGE" true)"
  run_cleanup; want_rc 18; want_inv container "$ctr" unproven; want_kept container "$ctr"
}
ctl_V29() { label_only "g001.run=$RT"; detail "exit $RC; run label alone is not authority (unproven, kept)"; }
ctl_V30() { label_only "platform.validation.preflight=$(ci_random_hex 16)"; detail "exit $RC; preflight label alone is not authority (unproven, kept)"; }
ctl_V31() {
  local ctr
  e_case; ctl_labels
  row label "g001.run=$RT"
  ctr="$(D run -d --rm "${LB[@]}" --label "g001.run=$RT" --network none "$CI_BUSYBOX_IMAGE" sleep 6)"
  run_cleanup; want_rc 0
  if grep -qF "$ctr" "$C/inv.tsv"; then prob "the ephemeral container was reported"; fi
  want_gone container "$ctr"
  detail "exit $RC; ephemeral --rm container exited within the wait; not reported"
}
ctl_V31b() {
  local ctr
  e_case; ctl_labels
  row label "g001.run=$RT"
  ctr="$(D run -d "${LB[@]}" --label "g001.run=$RT" --network none "$CI_BUSYBOX_IMAGE" sleep 300)"
  run_cleanup; want_rc 18; want_inv container "$ctr" unproven; want_kept container "$ctr"
  detail "exit $RC; still running after 60 s: unproven and kept"
}
reg_kept() { # LAYOUT-PATH: the registry evidence survives (without secrets), mode 700, non-removed row
  local r="$1" f
  if [[ "$F_REG" != "kept:$r" ]]; then prob "registry=$F_REG, want kept:$r"; fi
  for f in owned.tsv start-inventory.ids token cid cleanup-inventory.tsv; do if [[ ! -e "$r/$f" ]]; then prob "the kept registry lacks $f"; fi; done
  if [[ -e "$r/secrets" ]]; then prob "the generated-secret list survived"; fi
  if awk -F'\t' '$3 != "removed" {f = 1} END {exit !f}' "$r/cleanup-inventory.tsv" 2>/dev/null; then :; else prob "the kept inventory has no non-removed row"; fi
  if [[ "$(stat -c %a "$r" 2>/dev/null)" != 700 ]]; then prob "registry mode is not 700"; fi
}
local_ev() { local e; case_state; for e in "$CSTATE"/g003-validate-*; do if [[ -d "$e" ]]; then printf '%s' "$e"; return 0; fi; done; return 0; }
ctl_V32() {
  local ev
  new_case; run_v SHIM_FAIL_PS_FILTER=1 -- ; ev="$(local_ev)"
  expect_rc 18; expect_phase cleanup 18
  if [[ -z "$ev" ]]; then prob "the local evidence directory is gone"; else
    reg_kept "$ev/owned"
    if [[ "$(find "$ev" -mindepth 1 -maxdepth 1 -printf '%f ')" != "owned " ]]; then prob "\$EV is not reduced to owned/"; fi
  fi
  detail "exit $RC; registry kept at \$EV/owned without secrets; \$EV reduced to owned/"
}
ctl_V33() { std_ci 4802; run_v "${CIENV[@]}" SHIM_FAIL_PS_FILTER=1 -- --publish-dir ci-evidence
  expect_rc 18; reg_kept "$C/.platform-validation-owned-4802"; no_validate_left; detail "exit $RC; sibling registry kept; \$EV deleted"; }
ctl_V34() {
  local ev r1
  new_case; run_v STUB_HARNESS=fail SHIM_FAIL_PS_FILTER=1 -- ; ev="$(local_ev)"; r1="$RC"
  expect_rc 1; expect_phase cleanup 18
  if [[ -n "$ev" ]]; then reg_kept "$ev/owned"; else prob "local registry not kept"; fi
  std_ci 4803; run_v "${CIENV[@]}" STUB_HARNESS=fail SHIM_FAIL_PS_FILTER=1 -- --publish-dir ci-evidence
  expect_rc 1; expect_phase cleanup 18; reg_kept "$C/.platform-validation-owned-4803"
  detail "local exit $r1, CI exit $RC (harness precedence); cleanup=18; registry kept in both layouts"
}
ctl_V35() {
  local r1
  new_case; run_v -- ; r1="$RC"
  expect_rc 0
  if [[ "$F_REG" != removed ]]; then prob "local registry=$F_REG"; fi
  no_validate_left
  std_ci 4804; run_v "${CIENV[@]}" -- --publish-dir ci-evidence
  expect_rc 0
  if [[ "$F_REG" != removed ]]; then prob "CI registry=$F_REG"; fi
  if [[ -e "$C/.platform-validation-owned-4804" ]]; then prob "the CI registry remains"; fi
  no_validate_left
  detail "local exit $r1, CI exit $RC; registry=removed; nothing left"
}

# --- group E: the build ledger (G051 VC1-VC8; Plan §3.8) -------------------------------------------------------------
# A driver sources the real infra/validation/lib.sh and runs ledger_init, control-authored steps (ledger_build calls of
# busybox builds that carry this control's labels, and file or tag changes between them) and ledger_check, each run
# with its own WORK (one ledger). Each control asserts the exact rows and keys, expected == rows, the ac5_status and
# the removal of its labelled images. The parser cases (VC8) replay progress through a docker shim first on PATH;
# the progress they derive from is a real build's. Builds in these scratch trees are the controls' own scenarios
# (Plan §3.8.3), never the harness's ledger.
VC_K="dockerfile dockerfile_sha256 target frontend bases platform ignore_file ignore_sha256 context_inventory named_contexts args_digest"
vc_case() { # a fresh case, its labels and the driver $C/vcdrv.sh
  new_case; ctl_labels
  cat >"$C/vcdrv.sh" <<'VCDRV'
#!/usr/bin/env bash
# VC driver of infra/ci/validate-controls.sh: the real build ledger over control-authored steps.
#   vcdrv.sh REPO WORK STEPS
set -euo pipefail
REPO="$1" WORK="$2"
# shellcheck source=/dev/null
source "$REPO/infra/validation/lib.sh"
RUN_ID="vc$$" RESULTS="$WORK/results.tsv" LEDGER_JOB=vc LEDGER_SUITE=vc LEDGER_IMAGES_FILE="$WORK/images.txt"
: >"$RESULTS"; : >"$WORK/rcs"
vc_b() { local k="$1" rc=0; shift; ledger_build vc "$k" -- "$@" || rc=$?; printf '%s\n' "$rc" >>"$WORK/rcs"; }
ledger_init
# shellcheck source=/dev/null
source "$3"
ledger_check
VCDRV
}
vc_tag() { printf 'g003ctl-%s-%s-%s' "$RT" "$(tr 'A-Z' 'a-z' <<<"$CTL_ID")" "$1"; }
vc_ctx() { # DIR [copy-only]: a build context with a per-run nonce; FROM, COPY and (unless copy-only) RUN
  mkdir -p "$1"
  ci_random_hex 16 >"$1/nonce"
  chmod 644 "$1/nonce"
  if [[ "${2:-}" == copy-only ]]; then printf 'FROM busybox:1.37.0\nCOPY nonce /nonce\n' >"$1/Dockerfile"
  else printf 'FROM busybox:1.37.0\nCOPY nonce /nonce\nRUN cat /nonce >/dev/null\n' >"$1/Dockerfile"; fi
}
vc_b() { # STEPS KIND DOCKER-BUILD-ARGS...: append one ledger_build step
  local s="$1"; shift
  printf 'vc_b' >>"$s"; printf ' %q' "$@" >>"$s"; printf '\n' >>"$s"
}
vc_sh() { printf '%s\n' "$2" >>"$1"; } # STEPS LINE: append a control-authored shell line
vc_build() { # STEPS CTX TAG [KIND [EXTRA...]]: one labelled build of CTX
  local s="$1" ctx="$2" tag="$3" kind="${4:-authoritative}"
  shift 4 2>/dev/null || shift $#
  vc_b "$s" "$kind" "$@" -t "$tag" "${LB[@]}" -f "$ctx/Dockerfile" "$ctx"
}
vc_run() { # WORK STEPS [IMAGE...]: the driver over STEPS with its own ledger in WORK (images.txt: busybox and IMAGE...)
  local w="$1" s="$2"; shift 2
  mkdir -p "$w"; printf '%s\n' busybox:1.37.0 "$@" >"$w/images.txt"
  launch "$w.out" env -i -C "$C" "${BASE_ENV[@]}" TMPDIR="$C/tmp" SHIM_LABELS="$(case_labels)" "${VC_ENV[@]}" \
    bash "$C/vcdrv.sh" "$C/repo" "$w" "$s"
  await "$LPID" 900
  VW="$w"
  if [[ "$RC" != 0 ]]; then prob "${w##*/}: ledger driver exit $RC ($(tail -2 "$w.out" 2>/dev/null | tr '\n' ' '))"; fi
}
VC_ENV=() VW=""
vc_f() { # ROW FIELD: FIELD of data row ROW (1-based) of $VW/builds.tsv, by its header
  awk -F'\t' -v r="$1" -v f="$2" 'NR == 1 {for (i = 1; i <= NF; i++) c[$i] = i; next} NR - 1 == r {if (f in c) print $(c[f]); exit}' "$VW/builds.tsv" 2>/dev/null
}
vc_key() { local f k=""; for f in $VC_K; do k+="$(vc_f "$1" "$f")|"; done; printf '%s' "$k"; }
vc_s() { awk -F'\t' -v k="$1" '$1 == k {print $2; exit}' "$VW/ledger-summary.tsv" 2>/dev/null; }
vc_rows() { awk 'NR > 1' "$VW/builds.tsv" 2>/dev/null | grep -c . || true; }
vc_rcs() { tr '\n' ' ' <"$VW/rcs" 2>/dev/null; }
vc_want() { # ROWS AC5: rows == expected == ROWS, the ac5_status AC5, a BUILD-LEDGER INFO row
  local n w="${VW##*/}"
  n="$(vc_rows)"
  if [[ "$n" != "$1" || "$(vc_s rows)" != "$1" || "$(vc_s expected)" != "$1" ]]; then prob "$w: rows $n, summary rows $(vc_s rows), expected $(vc_s expected); want $1 each"; fi
  if [[ "$(vc_s ac5_status)" != "$2" ]]; then prob "$w: ac5_status $(vc_s ac5_status), want $2 (unknown $(vc_s unknown), duplicates $(vc_s duplicates), near_duplicates $(vc_s near_duplicates))"; fi
  if awk -F'\t' '$1 == "BUILD-LEDGER" && $2 == "INFO" {f = 1} END {exit !f}' "$VW/results.tsv" 2>/dev/null; then :; else prob "$w: no BUILD-LEDGER INFO row"; fi
}
vc_unaccounted() { # one expected invocation, kept in the denominator as an unknown row or a missing-row placeholder; UNMET
  local w="${VW##*/}" p
  p="$(vc_f 1 progress)"
  if [[ "$(vc_rows)" != 1 || "$(vc_s expected)" != 1 ]]; then prob "$w: builds.tsv rows $(vc_rows), expected $(vc_s expected); want 1 each"; fi
  if [[ -z "$p" || "$p" == parsed ]]; then prob "$w: progress [$p], want unknown or missing"; fi
  if [[ "$p" == missing && "$(vc_s rows)" != 0 ]] || [[ "$p" != missing && "$(vc_s rows)" != 1 ]]; then prob "$w: summary rows $(vc_s rows) for progress $p"; fi
  if [[ "$(vc_s unknown)/$(vc_s ac5_status)" != 1/UNMET ]]; then prob "$w: unknown/ac5_status $(vc_s unknown)/$(vc_s ac5_status), want 1/UNMET"; fi
  if awk -F'\t' '$1 == "BUILD-LEDGER" && $2 == "INFO" {f = 1} END {exit !f}' "$VW/results.tsv" 2>/dev/null; then :; else prob "$w: no BUILD-LEDGER INFO row"; fi
}
vc_parsed() { # ROW EXEC: progress parsed with steps_executed 0, or + (at least one)
  local p e
  p="$(vc_f "$1" progress)" e="$(vc_f "$1" steps_executed)"
  if [[ "$p" != parsed ]]; then prob "${VW##*/} row $1: progress $p ($(vc_f "$1" image_id)), want parsed"; return 0; fi
  if [[ "$2" == 0 && "$e" != 0 ]] || [[ "$2" == + && ! "$e" =~ ^[1-9][0-9]*$ ]]; then prob "${VW##*/} row $1: steps_executed $e, want $2"; fi
}
vc_known() { if [[ "$(vc_key "$1")" == *unknown* ]]; then prob "${VW##*/} row $1: a key field is unknown [$(vc_key "$1")]"; fi; }
vc_same() { if [[ "$(vc_key "$1")" != "$(vc_key "$2")" ]]; then prob "${VW##*/}: rows $1 and $2 have different keys"; fi; }
vc_differ() { if [[ "$(vc_key "$1")" == "$(vc_key "$2")" ]]; then prob "${VW##*/}: rows $1 and $2 have the same key"; fi; }
vc_clean() { # this control's labelled images (untagged too) removed, and verified; G065 K1: a failed listing or removal is a problem
  local x ids n o
  if ids="$(D image ls -aq --no-trunc --filter "label=sdlc.ci.control=$RT-$CTL_ID" | LC_ALL=C sort -u)"; then :; else
    prob "the image listing failed; cleanup unverified"; ids=""; fi
  for x in $ids; do o="$(D image rm -f "$x" 2>&1 >/dev/null)" || prob "the image $x could not be removed ($(head -1 <<<"$o"))"; done
  n="$(lcount "sdlc.ci.control=$RT-$CTL_ID")"
  if (( n < 0 )); then prob "the Docker listing of sdlc.ci.control=$RT-$CTL_ID failed; the count is unverified"
  elif [[ "$n" != 0 ]]; then prob "labelled resources remain after the control's own cleanup"; fi
}
# vc_record_tag TAG WHAT (G065 K2): a Compose build of the ledger driver bypasses D and the shim's build gate; its image is
# recorded by exact ID in this control's inventory (and the registry mirror) before cleanup. A failed inspection, a failed
# row write or a row missing afterwards is a problem naming TAG, never skipped (owned_cleanup then holds what is unrecorded).
vc_record_tag() {
  local id inv="$ROOT/owned/$CTL_ID.tsv"
  if id="$(D image inspect -f '{{.Id}}' "$1" 2>&1)"; then :; else
    prob "K2: the Compose-built image $1 could not be inspected after its build ($(head -1 <<<"$id"))"; return 0; fi
  gate_record "$inv" image "$id" "$2" "$REG_MIRROR" "$1" || prob "K2: the inventory row for $1 ($id) could not be written"
  grep -qF $'image\t'"$id"$'\t' "$inv" 2>/dev/null || prob "K2: $1 ($id) is not in the inventory after recording"
}
ctl_VC1() {
  local s
  vc_case; s="$C/s1"; vc_ctx "$C/b1"; : >"$s"
  vc_build "$s" "$C/b1" "$(vc_tag a)"; vc_build "$s" "$C/b1" "$(vc_tag a)"
  vc_run "$C/w1" "$s"
  vc_want 2 MET; vc_parsed 1 +; vc_parsed 2 0; vc_known 1; vc_same 1 2
  if [[ "$(vc_s duplicates)/$(vc_s unknown)" != 0/0 ]]; then prob "duplicates/unknown $(vc_s duplicates)/$(vc_s unknown), want 0/0"; fi
  vc_clean
  detail "identical input twice: one key; first parsed and executing, second parsed with 0 executed (a confirmed hit); rows = expected = 2; MET; images removed"
}
ctl_VC2() {
  local s
  vc_case; s="$C/s2"; vc_ctx "$C/b2"; : >"$s"
  vc_build "$s" "$C/b2" "$(vc_tag a)"
  vc_sh "$s" "printf 'changed\n' >>$(printf '%q' "$C/b2/nonce")"
  vc_build "$s" "$C/b2" "$(vc_tag a)"
  vc_run "$C/w2" "$s"
  # Plan r3 §3.8.3 (normative): executing rows sharing every K field except context_inventory are a near-duplicate,
  # listed with their differing paths, and need a named exception; so this pair is accounted and UNMET (the VC2 row's
  # "MET" contradicts §3.8.3; the G051 Result records the reading).
  vc_want 2 UNMET; vc_parsed 1 +; vc_parsed 2 +; vc_known 2; vc_differ 1 2
  if [[ "$(vc_f 1 context_inventory)" == "$(vc_f 2 context_inventory)" ]]; then prob "the context inventory did not change"; fi
  if [[ "$(vc_s duplicates)/$(vc_s near_duplicates)/$(vc_s unknown)" != 0/1/0 ]]; then prob "duplicates/near_duplicates/unknown $(vc_s duplicates)/$(vc_s near_duplicates)/$(vc_s unknown), want 0/1/0"; fi
  if [[ "$(vc_s near_duplicate)" != *" unexcepted paths=nonce" ]]; then prob "near_duplicate [$(vc_s near_duplicate)], want the pair unexcepted with paths=nonce"; fi
  vc_clean
  detail "a changed context file: a new key (context inventory), the COPY executed again; accounted as an unexcepted near-duplicate naming nonce (§3.8.3), so UNMET"
}
ctl_VC3() {
  local s t
  vc_case; s="$C/s3a" t="$C/s3b"; vc_ctx "$C/b3a"; vc_ctx "$C/b3b"; : >"$s"; : >"$t"
  vc_build "$s" "$C/b3a" "$(vc_tag a)"; vc_build "$s" "$C/b3a" "$(vc_tag a)" named:prebuild-fallback --no-cache
  vc_build "$t" "$C/b3b" "$(vc_tag b)"; vc_build "$t" "$C/b3b" "$(vc_tag b)" authoritative --no-cache
  vc_run "$C/w3a" "$s"; vc_want 2 MET; vc_parsed 1 +; vc_parsed 2 +; vc_same 1 2
  if [[ "$(vc_f 2 kind)" != named:prebuild-fallback ]]; then prob "w3a row 2 kind $(vc_f 2 kind)"; fi
  vc_run "$C/w3b" "$t"; vc_want 2 UNMET; vc_parsed 1 +; vc_parsed 2 +; vc_same 1 2
  if [[ "$(vc_s duplicates)" != 1 ]]; then prob "w3b duplicates $(vc_s duplicates), want 1"; fi
  vc_clean
  detail "two executing invocations of one key: with named:prebuild-fallback MET; both authoritative: duplicates 1, UNMET"
}
ctl_VC4() {
  local s n o
  vc_case; s="$C/s4"; vc_ctx "$C/b4"; : >"$s"
  vc_b "$s" prebuild -f "$C/b4/Dockerfile" "$C/b4"
  vc_build "$s" "$C/b4" "$(vc_tag consumer)"
  vc_run "$C/w4" "$s"
  if [[ "$(vc_rcs)" != "0 0 " ]]; then prob "build statuses [$(vc_rcs)], want 0 0"; fi
  vc_want 2 MET; vc_parsed 1 +; vc_parsed 2 0; vc_same 1 2
  if [[ "$(vc_f 1 kind)/$(vc_f 1 image_id)" != prebuild/cacheonly ]]; then prob "the prebuild row is $(vc_f 1 kind)/$(vc_f 1 image_id), want prebuild/cacheonly"; fi
  # G065 K1: the listing (untagged images too), then the count; a failed listing is unverified, never zero.
  if o="$(D image ls -aq --no-trunc --filter "label=sdlc.ci.control=$RT-$CTL_ID" | LC_ALL=C sort -u)"; then n="$(awk 'NF { k++ } END { print k + 0 }' <<<"$o")"; else n=-1; fi
  if (( n < 0 )); then prob "the image listing failed; the consumer count is unverified"
  elif [[ "$n" != 1 ]]; then prob "$n labelled images, want only the consumer's"; fi
  vc_clean
  detail "a cache-only prebuild then the consumer: the consumer executed 0 steps; the prebuild left no image or tag (cacheonly); one executing row; MET"
}
ctl_VC5() {
  local s ia ib base b1 b2
  vc_case; s="$C/s5"; mkdir -p "$C/b5"; ci_random_hex 16 >"$C/b5/nonce"; : >"$s"
  base="$(vc_tag base):latest"
  ia="$(mk_image "$(vc_tag base-a):latest")"; ib="$(mk_image "$(vc_tag base-b):latest")"
  printf 'FROM %s\nCOPY nonce /nonce\n' "$base" >"$C/b5/Dockerfile"
  vc_sh "$s" "docker tag $ia $base"; vc_build "$s" "$C/b5" "$(vc_tag a)"
  vc_sh "$s" "docker tag $ib $base"; vc_build "$s" "$C/b5" "$(vc_tag b)"
  vc_run "$C/w5" "$s" "$base"
  b1="$(vc_f 1 bases)" b2="$(vc_f 2 bases)"
  if [[ "$b1" != unknown && "$b2" != unknown && "$(vc_f 1 progress)" == parsed && "$(vc_f 2 progress)" == parsed ]]; then
    vc_want 2 MET; vc_differ 1 2
    if [[ "$b1" == "$b2" ]]; then prob "the re-pointed base did not change bases"; fi
    detail "the local base tag re-pointed between builds: bases differ, keys differ; both accounted; MET"
  else
    vc_want 2 UNMET
    detail "the local base identity was not resolved (bases [$b1] [$b2]): the rows are unknown, kept, and AC5 is UNMET (the refusal path)"
  fi
  vc_clean
}
ctl_VC6() {
  local s def other p1 p2
  vc_case; s="$C/s6"; vc_ctx "$C/b6" copy-only; : >"$s"
  def="$(D version --format '{{.Server.Os}}/{{.Server.Arch}}' 2>/dev/null)" || def=""
  if [[ "$def" == linux/arm64 ]]; then other=linux/amd64; else other=linux/arm64; fi
  vc_build "$s" "$C/b6" "$(vc_tag a)"; vc_build "$s" "$C/b6" "$(vc_tag b)" authoritative --platform "$other"
  vc_run "$C/w6" "$s"
  p1="$(vc_f 1 platform)" p2="$(vc_f 2 platform)"
  if [[ "$(vc_rcs)" == "0 0 " && "$p1" != unknown && "$p2" != unknown && "$(vc_f 2 progress)" == parsed ]]; then
    vc_want 2 MET; vc_differ 1 2
    if [[ "${p1%%;*}" == "${p2%%;*}" || "${p2%%;*}" != "$other" ]]; then prob "platforms [$p1] [$p2], want the default and $other"; fi
    detail "a COPY-only build for the default platform and for $other: platforms differ, keys differ; both accounted; MET"
  else
    vc_want 2 UNMET
    detail "the $other base was unavailable or unresolved (statuses [$(vc_rcs)], platforms [$p1] [$p2]): unknown, kept, UNMET"
  fi
  vc_clean
}
ctl_VC7() {
  local x d sha r=""
  vc_case
  # (a) a Dockerfile-specific ignore file added between two builds
  d="$C/b7a"; vc_ctx "$d"; x="$C/s7a"; : >"$x"
  vc_build "$x" "$d" "$(vc_tag a)"; vc_sh "$x" "printf 'unused-*\n' >$(printf '%q' "$d/Dockerfile.dockerignore")"; vc_build "$x" "$d" "$(vc_tag a)"
  vc_run "$C/w7a" "$x"; vc_want 2 MET; vc_differ 1 2
  if [[ "$(vc_f 1 ignore_file)/$(vc_f 2 ignore_file)" != none/Dockerfile.dockerignore ]]; then prob "(a) ignore files [$(vc_f 1 ignore_file)] [$(vc_f 2 ignore_file)]"; fi
  r+=" a"
  # (b) both ignore files: the Dockerfile-specific one is chosen, both are recorded
  d="$C/b7b"; vc_ctx "$d"; printf 'unused-b\n' >"$d/.dockerignore"; printf 'unused-a\n' >"$d/Dockerfile.dockerignore"; x="$C/s7b"; : >"$x"
  vc_build "$x" "$d" "$(vc_tag b)"
  vc_run "$C/w7b" "$x"; vc_want 1 MET
  sha="$(sha256sum <"$d/Dockerfile.dockerignore" | cut -c1-64)"
  if [[ "$(vc_f 1 ignore_file)" != Dockerfile.dockerignore*.dockerignore* || "$(vc_f 1 ignore_sha256)" != "$sha"*"$(sha256sum <"$d/.dockerignore" | cut -c1-64)"* ]]; then
    prob "(b) ignore [$(vc_f 1 ignore_file)] [$(vc_f 1 ignore_sha256)], want the Dockerfile-specific file first and both recorded"
  fi
  r+=" b"
  # (c) only a file mode changed
  d="$C/b7c"; vc_ctx "$d"; x="$C/s7c"; : >"$x"
  vc_build "$x" "$d" "$(vc_tag c)"; vc_sh "$x" "chmod 600 $(printf '%q' "$d/nonce")"; vc_build "$x" "$d" "$(vc_tag c)"
  vc_run "$C/w7c" "$x"; vc_want 2 UNMET; vc_differ 1 2; vc_parsed 2 +
  if [[ "$(vc_s near_duplicate)" != *" unexcepted paths=nonce" ]]; then prob "(c) near_duplicate [$(vc_s near_duplicate)], want the mode-changed nonce"; fi
  r+=" c"
  # (d) a symlink retargeted
  d="$C/b7d"; vc_ctx "$d"; printf 'one\n' >"$d/t1"; printf 'two\n' >"$d/t2"; ln -s t1 "$d/link"; x="$C/s7d"; : >"$x"
  vc_build "$x" "$d" "$(vc_tag d)"; vc_sh "$x" "ln -sfn t2 $(printf '%q' "$d/link")"; vc_build "$x" "$d" "$(vc_tag d)"
  vc_run "$C/w7d" "$x"; vc_want 2 MET; vc_differ 1 2; r+=" d"
  # (e) a regular file replaced by a symlink to a file with the same content
  d="$C/b7e"; vc_ctx "$d"; printf 'same\n' >"$d/f"; printf 'same\n' >"$d/g"; x="$C/s7e"; : >"$x"
  vc_build "$x" "$d" "$(vc_tag e)"; vc_sh "$x" "rm -f $(printf '%q' "$d/f") && ln -s g $(printf '%q' "$d/f")"; vc_build "$x" "$d" "$(vc_tag e)"
  vc_run "$C/w7e" "$x"; vc_want 2 MET; vc_differ 1 2; r+=" e"
  # (f) only an ignored file differs: distinct keys, the second a confirmed hit, no duplicate
  d="$C/b7f"; vc_ctx "$d"; printf 'ignored.txt\n' >"$d/.dockerignore"; printf 'v1\n' >"$d/ignored.txt"; x="$C/s7f"; : >"$x"
  vc_build "$x" "$d" "$(vc_tag f)"; vc_sh "$x" "printf 'v2\n' >$(printf '%q' "$d/ignored.txt")"; vc_build "$x" "$d" "$(vc_tag f)"
  vc_run "$C/w7f" "$x"; vc_want 2 MET; vc_differ 1 2; vc_parsed 2 0
  if [[ "$(vc_s duplicates)" != 0 ]]; then prob "(f) duplicates $(vc_s duplicates), want 0"; fi
  r+=" f"
  # (g) a near-duplicate executing pair (only an ignored file differs, the second forced): without, then with a named exception
  d="$C/b7g"; vc_ctx "$d"; printf 'ignored.txt\n' >"$d/.dockerignore"; printf 'v1\n' >"$d/ignored.txt"; x="$C/s7g"; : >"$x"
  vc_build "$x" "$d" "$(vc_tag g)"; vc_sh "$x" "printf 'v2\n' >$(printf '%q' "$d/ignored.txt")"; vc_build "$x" "$d" "$(vc_tag g)" authoritative --no-cache
  vc_run "$C/w7g" "$x"; vc_want 2 UNMET; vc_parsed 1 +; vc_parsed 2 +
  if [[ ! "$(vc_s near_duplicates)" =~ ^[1-9][0-9]*$ ]]; then prob "(g) near_duplicates $(vc_s near_duplicates), want at least 1"; fi
  d="$C/b7h"; vc_ctx "$d"; printf 'ignored.txt\n' >"$d/.dockerignore"; printf 'v1\n' >"$d/ignored.txt"; x="$C/s7h"; : >"$x"
  vc_build "$x" "$d" "$(vc_tag h)"; vc_sh "$x" "printf 'v2\n' >$(printf '%q' "$d/ignored.txt")"; vc_build "$x" "$d" "$(vc_tag h)" named:B-BUILD-CONTEXT --no-cache
  vc_run "$C/w7h" "$x"; vc_want 2 MET; vc_parsed 1 +; vc_parsed 2 +; r+=" g"
  vc_clean
  detail "context and ignore configuration ($r): a Dockerfile-specific ignore file, a mode change, a retargeted symlink and a file replaced by a same-content symlink each give a new key (MET; the mode change re-executes the COPY and is an unexcepted near-duplicate naming nonce, UNMET, per §3.8.3); with both ignore files the Dockerfile-specific one is chosen and both recorded; an ignored-only change is a new key whose build is a confirmed hit; a forced near-duplicate is UNMET without and MET with named:B-BUILD-CONTEXT"
}
# VC8: progress and accounting that cannot be confirmed stay in the denominator and make AC5 UNMET.
VC8_CASES="empty tty no-terminal odd-token dup-terminal no-index exit-no-error no-base"
vc8_fixture() { # NAME SOURCE OUT: one derived parser fixture
  local n
  n="$(awk '/^#[0-9]+ \[[0-9]+\/[0-9]+\] COPY / {sub(/^#/, "", $1); print $1; exit}' "$2")"
  case "$1" in
    empty) : >"$3" ;;
    tty) sed -E 's/^#[0-9]+ /=> /' "$2" >"$3" ;;
    no-terminal) sed "/^#$n DONE/d" "$2" >"$3" ;;
    odd-token) sed "s/^#$n DONE/#$n WEIRD/" "$2" >"$3" ;;
    dup-terminal) sed "/^#$n DONE/p" "$2" >"$3" ;;
    no-index) sed "/^#$n /d" "$2" >"$3" ;;
    exit-no-error) cp -- "$2" "$3" ;;
    no-base) sed -E 's/busybox:1\.37\.0(@sha256:[0-9a-f]{64})?/busybox:0.0.0-g051-absent/g' "$2" >"$3" ;;
  esac
  if [[ -z "$n" ]]; then prob "the real progress has no COPY vertex to derive $1 from"; fi
}
ctl_VC8() {
  local s d i n p iid proj e r=""
  vc_case; d="$C/b8"; vc_ctx "$d"
  # The no-base case builds a context whose base is absent from the local store (its fixture names the same base).
  mkdir -p "$C/b8n"; cp -- "$d/nonce" "$C/b8n/nonce"; sed 's/^FROM busybox:1\.37\.0$/FROM busybox:0.0.0-g051-absent/' "$d/Dockerfile" >"$C/b8n/Dockerfile"
  if grep -qx 'FROM busybox:0.0.0-g051-absent' "$C/b8n/Dockerfile"; then :; else prob "the absent-base context could not be written"; fi
  # A: a real build; its progress and image ID become the parser fixtures' source.
  s="$C/s8a"; : >"$s"; vc_build "$s" "$d" "$(vc_tag a)"
  vc_sh "$s" 'cp -- "$LEDGER_DIR/progress/$LG_CALL.log" "$WORK/fx.log"; cp -- "$LEDGER_DIR/progress/$LG_CALL.iid" "$WORK/fx.iid"'
  vc_run "$C/w8a" "$s"; vc_want 1 MET; vc_parsed 1 +
  iid="$(cat "$C/w8a/fx.iid" 2>/dev/null)" || iid=""
  if [[ ! -s "$C/w8a/fx.log" || ! "$iid" =~ ^sha256:[0-9a-f]{64}$ ]]; then prob "no real progress or image ID to derive the parser fixtures from"; vc_clean; return 0; fi
  # The docker shim: a build replays VC_FIXTURE and exits VC_RC (VC_IID into --iidfile on 0); all else passes through.
  mkdir -p "$C/vcshim" "$C/fx"
  cat >"$C/vcshim/docker" <<'VCSHIM'
#!/usr/bin/env bash
# VC8 docker shim of infra/ci/validate-controls.sh. VC_WITHHOLD=1: a Compose call that can build runs for real and
# its output is withheld (progress genuinely absent from the capture), whatever the Compose version prints.
if [[ "${VC_WITHHOLD:-0}" == 1 && "${1:-}" == compose ]]; then
  b=0; for a in "$@"; do case "$a" in config|ps|down|images|ls|version) b=0; break ;; build|run|up|create) b=1 ;; esac; done
  if (( b )); then "${VC_NEXT:?}" "$@" >/dev/null 2>&1; exit $?; fi
fi
if [[ -n "${VC_FIXTURE:-}" ]] && [[ "${1:-}" == build || ( "${1:-}" == buildx && "${2:-}" == build ) ]]; then
  iid="" prev=""
  for a in "$@"; do if [[ "$prev" == --iidfile ]]; then iid="$a"; fi; prev="$a"; done
  cat -- "$VC_FIXTURE" >&2
  if [[ "${VC_RC:-0}" == 0 && -n "$iid" ]]; then printf '%s' "$VC_IID" >"$iid"; fi
  exit "${VC_RC:-0}"
fi
exec "${VC_NEXT:?}" "$@"
VCSHIM
  chmod 755 "$C/vcshim/docker"
  VC_ENV=(PATH="$C/vcshim:$ROOT/shim:$PATH" VC_NEXT="$ROOT/shim/docker")
  # B: the well-formed real progress replayed: parsed and executing (the replay channel is sound).
  s="$C/s8b"; : >"$s"
  vc_sh "$s" "export VC_FIXTURE=$(printf '%q' "$C/w8a/fx.log") VC_RC=0 VC_IID=$iid"; vc_build "$s" "$d" "$(vc_tag a)"
  vc_run "$C/w8b" "$s"; vc_want 1 MET; vc_parsed 1 +
  # C: the eight defects, each one row of one ledger.
  s="$C/s8c"; : >"$s"
  for e in $VC8_CASES; do
    vc8_fixture "$e" "$C/w8a/fx.log" "$C/fx/$e.log"
    vc_sh "$s" "export VC_FIXTURE=$(printf '%q' "$C/fx/$e.log") VC_RC=$([[ "$e" == exit-no-error ]] && echo 1 || echo 0) VC_IID=$iid"
    if [[ "$e" == no-base ]]; then vc_build "$s" "$C/b8n" "$(vc_tag a)"; else vc_build "$s" "$d" "$(vc_tag a)"; fi
  done
  vc_run "$C/w8c" "$s"
  n=0
  for e in $VC8_CASES; do
    n=$((n + 1)); p="$(vc_f "$n" progress)"
    if [[ "$p" != parsed || "$(vc_f "$n" bases)" == unknown ]]; then :; else prob "$e: progress parsed with known bases"; fi
    if [[ "$p" == parsed && "$(vc_f "$n" steps_executed)" == 0 ]]; then prob "$e: counted as a confirmed cache hit"; fi
    r+=" $e=$p"
  done
  vc_want 8 UNMET
  if [[ "$(vc_s unknown)" != 8 ]]; then prob "unknown $(vc_s unknown), want 8 (every defect kept and counted)"; fi
  VC_ENV=()
  # D: a real Compose build whose output the shim withholds (G051 C5: BUILDKIT_PROGRESS=quiet is not absent progress
  # under the pinned Compose v5); D2: the same build observed (a positive Compose row); E: a new <project>-<service>
  # image with no build progress.
  proj="$(vc_tag p)"
  mkdir -p "$C/cdest/infra/compose"
  cp -a -- "$d" "$C/cdest/ctx"; ci_random_hex 16 >"$C/cdest/ctx/nonce"
  printf 'COMPOSE_PROJECT_NAME=%s\n' "$proj" >"$C/cdest/.env"
  {
    printf 'services:\n  svc:\n    build:\n      context: ../../ctx\n      labels:\n'
    for i in $(case_labels); do printf '        %s: "%s"\n' "${i%%=*}" "${i#*=}"; done
  } >"$C/cdest/infra/compose/local.yml"
  VC_ENV=(PATH="$C/vcshim:$ROOT/shim:$PATH" VC_NEXT="$ROOT/shim/docker")
  s="$C/s8d"; : >"$s"
  vc_sh "$s" "export VC_WITHHOLD=1; compose $(printf '%q' "$C/cdest") build svc || printf 'compose\n' >>\"\$WORK/rcs\"; unset VC_WITHHOLD"
  vc_run "$C/w8d" "$s"; vc_unaccounted
  vc_record_tag "$proj-svc:latest" "VC8 compose build (withheld)"
  r+=" compose-withheld=$(vc_f 1 progress)"
  s="$C/s8d2"; : >"$s"
  vc_sh "$s" "compose $(printf '%q' "$C/cdest") build svc || printf 'compose\n' >>\"\$WORK/rcs\""
  vc_run "$C/w8d2" "$s"; vc_want 1 MET
  vc_record_tag "$proj-svc:latest" "VC8 compose build (observed)"
  if [[ "$(vc_f 1 progress)" != parsed ]]; then prob "w8d2: the observed Compose build is $(vc_f 1 progress), want parsed"; fi
  r+=" compose-observed=$(vc_f 1 progress)"
  VC_ENV=()
  e="$(mk_image "$(vc_tag e-src):latest")"
  s="$C/s8e"; : >"$s"
  vc_sh "$s" "ledger_watch_begin $(printf '%q' "$C/cdest") $proj vc8; docker tag $e $proj-svc:latest; : >\"\$WORK/w.log\"; ledger_watch_end $(printf '%q' "$C/cdest") $proj vc8 \"\$WORK/w.log\" ''"
  vc_run "$C/w8e" "$s"; vc_unaccounted
  r+=" image-without-build=$(vc_f 1 progress)"
  D image rm "$proj-svc:latest" >/dev/null 2>&1 || true
  vc_clean
  detail "a real build parsed (MET) and its replay parsed (MET); defects kept as unknown rows, UNMET, none a cache hit:$r"
}

# VC9 (G051 C3): Compose Bake progress of the pinned hosted toolchain. The CLI and plugins of the docker CLI image
# that infra/docker/ci-context.Dockerfile pins (Compose v5.5.1, buildx v0.37.2: asserted, the captured formats' toolchain)
# are copied out and serve, behind the controls' shim, as the driver's docker against the local daemon (its builder
# is only read, never created or selected). A: real builds through the real ledger: one service (an unprefixed Bake
# session), two services sharing one Dockerfile (prefixed, one shared build definition), a command block of two
# sequential invocations, and a direct build; every row parsed with the verified driver, MET; the Compose streams
# kept by the shim (tee) as replay fixtures. B: those streams replayed through the real parser and evaluator: the
# same keys, MET. C: derived defects, each its own ledger, UNMET with its rows kept: an export naming two services
# of the call, an export naming no service built here, one session naming a service while another goes unbuilt,
# builder evidence unavailable, a builder that changes during the call, a header contradicting the evidence, a step
# without terminal line, no export, an unresolved base, truncated progress, no Bake definitions vertex, mixed prefixes.
# G051 C4: D: the captured single-service and shared-definition streams with their build definition vertex re-printed
# in a second print block (BuildKit re-prints a vertex name per block): each vertex counts once, MET; C adds the
# same re-print under a different name (a conflict), UNMET. G051 C5: E: the captured two-target stream with a shared
# step re-printed under the other service's prefix (a valid alias), MET; C adds the re-print with another operation,
# stage or index, or without the service prefix (mixed), each UNMET (mixed keeps the row of its stage name read as a
# service). G051 C6: E2: the two frontend vertices with a lone service label (the d7 matched-daemon form) injected
# into the captured two-target stream and re-printed under the other bound target, MET with no context; C adds the
# re-print with the other operation, under a project service this call does not build, without a label and under
# the reserved auth token, each UNMET with its rows kept and the named vertex's reduced context in ledger-summary.tsv
# (names verbatim, a planted log line only as "<log>").
VC9_COMPOSE=v5.5.1 VC9_BUILDX=v0.37.2
vc9_pin() { # the pinned CLI and plugins into $C/pin; VC9_VER; returns 1 when unavailable
  local img cid ok=1
  img="$(awk '/^FROM docker:[^ ]*-cli@sha256:[0-9a-f]+ / {print $2; exit}' "$REPO_R/infra/docker/ci-context.Dockerfile")"
  mkdir -p "$C/pin/bin" "$C/pin/cfg/cli-plugins"; printf '{}\n' >"$C/pin/cfg/config.json"
  cid="$(D create "${LB[@]}" "$img" 2>/dev/null)" || { prob "the pinned CLI image [$img] is unavailable"; return 1; }
  D cp "$cid:/usr/local/bin/docker" "$C/pin/bin/docker" >/dev/null 2>&1 || ok=0
  D cp "$cid:/usr/local/libexec/docker/cli-plugins/." "$C/pin/cfg/cli-plugins/" >/dev/null 2>&1 || ok=0
  D rm "$cid" >/dev/null 2>&1 || true
  if (( ! ok )); then prob "the pinned CLI and plugins could not be copied out of $img"; return 1; fi
  VC9_VER="$(DOCKER_CONFIG="$C/pin/cfg" "$C/pin/bin/docker" --version 2>&1 | head -1); compose $(DOCKER_CONFIG="$C/pin/cfg" "$C/pin/bin/docker" compose version --short 2>&1 | head -1); buildx $(DOCKER_CONFIG="$C/pin/cfg" "$C/pin/bin/docker" buildx version 2>&1 | awk '{print $2; exit}')"
  if [[ "$VC9_VER" != *"compose $VC9_COMPOSE"* && "$VC9_VER" != *"compose ${VC9_COMPOSE#v}"* ]] || [[ "$VC9_VER" != *"buildx $VC9_BUILDX"* ]]; then
    prob "the pinned toolchain [$VC9_VER] is not the captured one (Compose $VC9_COMPOSE, buildx $VC9_BUILDX)"; return 1
  fi
}
vc9_defs() { # STREAM: its distinct build definition lines. BuildKit re-prints an unfinished vertex's name when it
  # resumes that vertex after another one's output, so an identical repeated line is another observation of the same
  # vertex; another vertex number, or the same number under another name, is a distinct line.
  { grep -E '^#[0-9]+ (\[[^]]*\] )?load build definition from ' "$1" || true; } | LC_ALL=C sort -u
}
vc9_one_def() { # STREAM: exactly one distinct build definition line, service-prefixed (two targets, one shared definition)
  [[ "$(vc9_defs "$1")" =~ ^#[0-9]+\ \[(two|three)\ internal\]\ load\ build\ definition\ from\ Dockerfile$ ]]
}
vc9_fx() { # NAME SRC OUT: one derived fixture of a captured Compose stream
  local n r
  r="$(awk '/^#[0-9]+ .*RUN cat \/nonce/ {sub(/^#/, "", $1); print $1; exit}' "$2")"
  n="$(awk '/^#[0-9]+ (\[[^]]*\] )?exporting to image$/ {sub(/^#/, "", $1); print $1; exit}' "$2")"
  case "$1" in
    no-terminal) sed "/^#$r DONE/d" "$2" >"$3" ;;
    no-export) sed "/^#$n /d" "$2" >"$3" ;;
    truncated) sed "/^#$r DONE/,\$d" "$2" >"$3" ;;
    no-bake-vertex) sed '/^#1 /d' "$2" >"$3" ;;
    unbuilt-name) sed -E 's#naming to [^ ]+#naming to docker.io/library/g003ctl-vc9-not-built:latest#' "$2" >"$3" ;;
    header) { printf '#0 building with "vc9" instance using docker-container driver\n\n'; cat -- "$2"; } >"$3" ;;
    mixed) { cat -- "$2"; printf '\n#99 [internal] load build definition from Dockerfile\n#99 DONE 0.0s\n'; } >"$3" ;;
    reprint|conflict) # the build definition vertex re-printed after the next vertex's name (conflict: under another name)
      awk -v mode="$1" '
        NR == FNR { if (d == "" && $0 ~ /^#[0-9]+ (\[[^]]*\] )?load build definition from /) { d = $1; h = $0 }
                    else if (d != "" && nx == "" && $1 != d && $0 ~ /^#[0-9]+ \[/) nx = $0
                    next }
        { print }
        $1 == d && /transferring dockerfile/ && !x { x = 1; print nx; r = h; if (mode == "conflict") sub(/\[internal\]/, "[vc9 internal]", r); print r }
      ' "$2" "$2" >"$3"
      if ! grep -q 'transferring dockerfile' "$2"; then prob "the captured stream $2 has no build definition transfer to derive $1 from"; fi ;;
    absent) sed -E -e 's#docker\.io/library/busybox:1\.37\.0(@sha256:[0-9a-f]{64})?#docker.io/library/busybox:0.0.0-g051-absent#g' \
              -e "s#naming to [^ ]+#naming to docker.io/library/$VC9_ABSENT:latest#" "$2" >"$3" ;;
  esac
  if [[ -z "$r" || -z "$n" ]]; then prob "the captured stream $2 has no RUN or export vertex to derive $1 from"; fi
}
vc9_alias() { # VARIANT(alias|op|stage|index|mixed) SRC OUT: the first service-prefixed stage step of a two-target
  # stream re-printed, before its terminal line and after the next vertex name, under the other service (alias) or
  # with another operation, stage or index, or with no service prefix (mixed)
  awk -v var="$1" '
    NR == FNR {
      if (n == "" && $0 ~ /^#[0-9]+ \[(two|three) [a-z0-9-]+ [0-9]+\/[0-9]+\] /) { n = $1; h = $0; next }
      if (n != "" && !t && $1 == n && ($2 == "DONE" || $2 == "CACHED")) { t = 1; next }
      if (t && m == "" && $1 != n && $0 ~ /^#[0-9]+ \[/) m = $0
      next
    }
    n != "" && !x && $1 == n && ($2 == "DONE" || $2 == "CACHED") {
      x = 1; i1 = index(h, "["); i2 = index(h, "] "); split(substr(h, i1 + 1, i2 - i1 - 1), tk, " "); text = substr(h, i2 + 2)
      other = (tk[1] == "two") ? "three" : "two"; stg = tk[2]; idx = tk[3]
      if (var == "op") { w = text; sub(/ .*$/, "", w); text = ((w == "COPY") ? "ADD" : "COPY") substr(text, length(w) + 1) }
      if (var == "stage") stg = stg "x"
      if (var == "index") { split(idx, q, "/"); idx = ((q[1] == q[2]) ? q[1] - 1 : q[1] + 1) "/" q[2] }
      lab = (var == "mixed") ? stg " " idx : other " " stg " " idx
      print m; print n " [" lab "] " text
    }
    { print }
  ' "$2" "$2" >"$3"
  if ! grep -qE '^#[0-9]+ \[(two|three) [a-z0-9-]+ [0-9]+/[0-9]+\] ' "$2"; then prob "the captured stream $2 has no service-prefixed stage step to derive $1 from"; fi
}
vc9_lone() { # VARIANT(alias|text|target|mixed|reserved) SRC OUT PLANT: after the Bake definitions vertex, the lone-label
  # frontend vertices #90 (resolve image config, with the log line PLANT) and #91 (docker-image) printed under two and
  # re-printed in a later block under three (alias); #90's re-print instead takes the other operation (text), the
  # unbuilt project service four (target), no label (mixed) or the reserved [auth] token (reserved)
  awk -v var="$1" -v plant="$4" '
    { print }
    /^#1 DONE / && !x {
      x = 1; rs = "resolve image config for docker-image://docker.io/docker/dockerfile:1"
      di = "docker-image://docker.io/docker/dockerfile:1@sha256:4edf897a3ffa55b89f906fc8cc78afdb3f1834cc9c7083565e611a8a7d5fe99e"
      r90 = "[three] " rs
      if (var == "text") r90 = "[three] " di; else if (var == "target") r90 = "[four] " rs
      else if (var == "mixed") r90 = rs; else if (var == "reserved") r90 = "[auth] " rs
      printf "\n#90 [two] %s\n#90 0.100 %s\n#91 [two] %s\n#90 %s\n#90 DONE 0.5s\n\n#91 [three] %s\n#91 DONE 0.8s\n", rs, plant, di, r90, di
    }
  ' "$2" >"$3"
  if ! grep -q '^#1 DONE ' "$2"; then prob "the captured stream $2 has no Bake definitions terminal to inject the frontend vertices after"; fi
}
vc9_context() { # NAME PLANT REPRINT: ledger-summary.tsv names #90 with its first name and REPRINT verbatim, its planted
  # log line only as "<log>", and at most 121 context lines
  local f="$VW/ledger-summary.tsv"
  if ! grep -qF '#90 [two] resolve image config for docker-image://docker.io/docker/dockerfile:1' "$f" || ! grep -qF -- "#90 $3" "$f"; then
    prob "$1: unknown_context lacks the #90 names [$3]"; fi
  if ! grep -qF '#90 <log>' "$f"; then prob "$1: unknown_context lacks the reduced log line of #90"; fi
  if grep -qF -- "$2" "$f"; then prob "$1: the planted log line reached ledger-summary.tsv"; fi
  if [[ "$(grep -c '^unknown_context' "$f")" -gt 121 ]]; then prob "$1: more than 121 unknown_context lines"; fi
}
vc9_unmet() { # NAME EXPECTED: rows kept (expected), at least one unknown row, UNMET
  if [[ "$(vc_s expected)" != "$2" ]]; then prob "$1: expected $(vc_s expected), want $2 (every predicted or printed build kept)"; fi
  if [[ ! "$(vc_s unknown)" =~ ^[1-9][0-9]*$ ]]; then prob "$1: unknown $(vc_s unknown), want at least 1"; fi
  if [[ "$(vc_s ac5_status)" != UNMET ]]; then prob "$1: ac5_status $(vc_s ac5_status), want UNMET"; fi
}
vc9_dest() { # DEST [dup]: a Compose project: one, two and three (one context and Dockerfile), four, absent [and dup: one's image]
  local x i
  mkdir -p "$1/infra/compose"; printf 'COMPOSE_PROJECT_NAME=%s\n' "$(vc_tag p)" >"$1/.env"
  {
    printf 'services:\n'
    for x in one:one two:shared three:shared four:four absent:absent ${2:+dup:one}; do
      printf '  %s:\n    image: %s\n    build:\n      context: %s\n' "${x%%:*}" \
        "$([[ "${x%%:*}" == dup ]] && vc_tag one || vc_tag "${x%%:*}")" "$C/c9/${x#*:}"
      case "${x%%:*}" in two|three) printf '      target: %s\n' "${x%%:*}" ;; esac
      printf '      labels:\n'
      for i in $(case_labels); do printf '        %s: "%s"\n' "${i%%=*}" "${i#*=}"; done
    done
  } >"$1/infra/compose/local.yml"
}
ctl_VC9() {
  local s e d f w b blk r="" n k
  vc_case; mkdir -p "$C/c9"
  vc9_pin || { vc_clean; return 0; }
  # Each context's Dockerfile is its own (a LABEL): builds differing only in context content would be near-duplicates.
  for d in one shared four; do vc_ctx "$C/c9/$d"; printf 'LABEL g003.vc9=%s\n' "$d" >>"$C/c9/$d/Dockerfile"; done
  # two and three: one context and Dockerfile, two targets over a shared stage (as minio and minio-init).
  printf 'FROM busybox:1.37.0 AS base\nCOPY nonce /nonce\nRUN cat /nonce >/dev/null\nFROM base AS two\nRUN echo two >/two\nFROM base AS three\nRUN echo three >/three\nLABEL g003.vc9=shared\n' >"$C/c9/shared/Dockerfile"
  mkdir -p "$C/c9/absent"; cp -- "$C/c9/one/nonce" "$C/c9/absent/nonce"
  sed 's/^FROM busybox:1\.37\.0$/FROM busybox:0.0.0-g051-absent/' "$C/c9/one/Dockerfile" >"$C/c9/absent/Dockerfile"
  VC9_ABSENT="$(vc_tag absent)"
  vc9_dest "$C/cd9"; vc9_dest "$C/cd9dup" dup
  blk="docker compose --env-file .env -f infra/compose/local.yml build four"$'\n'"docker compose --env-file .env -f infra/compose/local.yml build one"
  mkdir -p "$C/vc9shim"
  cat >"$C/vc9shim/docker" <<'VC9SHIM'
#!/usr/bin/env bash
# VC9 docker shim of infra/ci/validate-controls.sh: a Compose call that can build is teed to $VC_TEE.<n>.{out,err}
# (VC_TEE set) or replayed from VC_FX_OUT/VC_FX_ERR with @N@ = the call's number and exit VC_RC (VC_FX_ERR set);
# `buildx inspect` answers from VC_BLD (fail, or a file per call, @N@ = its number, the last repeated); all else passes.
cnt() { local n; n="$(cat -- "$1" 2>/dev/null)" || n=0; n=$((n + 1)); printf '%s\n' "$n" >"$1"; printf '%s' "$n"; }
if [[ "${1:-}" == compose ]]; then
  b=0; for a in "$@"; do case "$a" in config|ps|down|images|ls|version) b=0; break ;; build|run|up|create) b=1 ;; esac; done
  if (( b )) && [[ -n "${VC_FX_ERR:-}" ]]; then
    n="$(cnt "$VC_FX_N")"; o="${VC_FX_OUT//@N@/$n}"; e="${VC_FX_ERR//@N@/$n}"
    if [[ -f "$o" ]]; then cat -- "$o"; fi
    if [[ -f "$e" ]]; then cat -- "$e" >&2; fi
    exit "${VC_RC:-0}"
  fi
  if (( b )) && [[ -n "${VC_TEE:-}" ]]; then
    n="$(cnt "$VC_TEE.n")"; rc=0
    "${VC_NEXT:?}" "$@" >"$VC_TEE.$n.out" 2>"$VC_TEE.$n.err" || rc=$?
    cat -- "$VC_TEE.$n.out"; cat -- "$VC_TEE.$n.err" >&2
    exit "$rc"
  fi
fi
if [[ "${1:-}" == buildx && "${2:-}" == inspect && -n "${VC_BLD:-}" ]]; then
  if [[ "$VC_BLD" == fail ]]; then echo "vc9 shim: builder unavailable" >&2; exit 1; fi
  n="$(cnt "$VC_BLD_N")"; f="${VC_BLD//@N@/$n}"
  while [[ ! -f "$f" && "$n" -gt 1 ]]; do n=$((n - 1)); f="${VC_BLD//@N@/$n}"; done
  cat -- "$f"; exit 0
fi
exec "${VC_NEXT:?}" "$@"
VC9SHIM
  chmod 755 "$C/vc9shim/docker"
  VC_ENV=(PATH="$C/vc9shim:$ROOT/shim:$PATH" VC_NEXT="$ROOT/shim/docker" SHIM_REAL_DOCKER="$C/pin/bin/docker" DOCKER_CONFIG="$C/pin/cfg")
  # A: real builds with the pinned toolchain; streams teed.
  s="$C/s9a"; : >"$s"
  vc_sh "$s" "export VC_TEE=\"\$WORK/tee\""
  vc_sh "$s" "compose $(printf '%q' "$C/cd9") build one || printf 'compose-one\n' >>\"\$WORK/rcs\""
  vc_sh "$s" "compose $(printf '%q' "$C/cd9") build two three || printf 'compose-two-three\n' >>\"\$WORK/rcs\""
  vc_sh "$s" "ledger_run_block $(printf '%q' "$C/cd9") $(printf '%q' "$blk") \"\$WORK/block.log\" || printf 'block\n' >>\"\$WORK/rcs\""
  vc_sh "$s" "unset VC_TEE"
  vc_build "$s" "$C/c9/one" "$(vc_tag direct)"
  vc_sh "$s" 'cp -- "$LEDGER_DIR/progress/$LG_CALL.log" "$WORK/direct.log"; cp -- "$LEDGER_DIR/progress/$LG_CALL.iid" "$WORK/direct.iid"'
  vc_sh "$s" 'docker buildx inspect >"$WORK/builder.txt" 2>&1 || true'
  vc_run "$C/w9a" "$s"; w="$C/w9a"
  # G065 K2: the Compose-built images of A (one, two, three; the block's four), recorded by exact ID.
  for n in one two three four; do vc_record_tag "$(vc_tag "$n"):latest" "VC9 A compose build ($n)"; done
  if [[ -n "$(vc_rcs)" && "$(vc_rcs)" != "0 " ]]; then prob "A: statuses [$(vc_rcs)], want only the direct build's 0"; fi
  vc_want 6 MET
  # Rows: one; three and two (sorted; each runs its own target's step, Bake runs the shared stage once); the
  # block's four and one; the direct build. Executing: one, three, two, four; the repeated one and direct: hits.
  for n in 1 2 3 4; do vc_parsed "$n" +; done
  for n in 5 6; do vc_parsed "$n" 0; done
  for n in 1 2 3 4 5 6; do vc_known "$n"
    if [[ "$(vc_f "$n" daemon)" != */docker ]]; then prob "A row $n: daemon/driver $(vc_f "$n" daemon), want the verified docker driver"; fi
  done
  vc_same 1 5; vc_same 1 6; vc_differ 2 3
  # The pinned Compose writes BuildKit progress to stdout and its Image status lines to stderr.
  for f in 1 2 3 4; do if [[ ! -s "$w/tee.$f.out" ]]; then prob "A: no captured progress for Compose call $f"; fi; done
  if grep -q '^#0 building with' "$w"/tee.*.out "$w"/tee.*.err 2>/dev/null; then prob "A: the pinned Compose printed a #0 header (not the captured headerless format)"; fi
  if grep -q '^#1 \[internal\] load local bake definitions$' "$w/tee.1.out" && grep -q '^#[0-9]* \[[0-9]/[0-9]\] RUN cat /nonce' "$w/tee.1.out"; then :; else prob "A: call 1 is not an unprefixed Bake stream"; fi
  if vc9_one_def "$w/tee.2.out"; then :; else prob "A: call 2 is not a prefixed Bake stream with one shared build definition [$(vc9_defs "$w/tee.2.out" | tr '\n' '|')]"; fi
  if grep -q '^#0 building with "[^"]*" instance using docker driver$' "$w/direct.log"; then :; else prob "A: the direct build has no #0 header"; fi
  r+=" A: real [$VC9_VER; builder $(awk '/^Name:/ {print $2; exit}' "$w/builder.txt" 2>/dev/null)/$(awk '/^Driver:/ {print $2; exit}' "$w/builder.txt" 2>/dev/null)] 6 rows parsed, MET;"
  # B: the captured streams replayed (one, two three, the block's two calls, the direct build).
  for f in 1 2 3 4; do cp -- "$w/tee.$f.out" "$C/fx9-$f.out" 2>/dev/null || : >"$C/fx9-$f.out"; cp -- "$w/tee.$f.err" "$C/fx9-$f.err" 2>/dev/null || : >"$C/fx9-$f.err"; done
  cat >"$C/vc8shim" <<'VC8SHIM'
#!/usr/bin/env bash
# VC9 direct-build replay: VC_DFX on stderr and VC_DIID into --iidfile; all else passes through.
if [[ -n "${VC_DFX:-}" ]] && [[ "${1:-}" == build || ( "${1:-}" == buildx && "${2:-}" == build ) ]]; then
  iid="" prev=""; for a in "$@"; do if [[ "$prev" == --iidfile ]]; then iid="$a"; fi; prev="$a"; done
  cat -- "$VC_DFX" >&2; if [[ -n "$iid" ]]; then printf '%s' "$VC_DIID" >"$iid"; fi; exit 0
fi
exec "${VC_NEXT2:?}" "$@"
VC8SHIM
  mkdir -p "$C/vc9d"; mv -- "$C/vc8shim" "$C/vc9d/docker"; chmod 755 "$C/vc9d/docker"
  VC_ENV=(PATH="$C/vc9d:$C/vc9shim:$ROOT/shim:$PATH" VC_NEXT2="$C/vc9shim/docker" VC_NEXT="$ROOT/shim/docker" SHIM_REAL_DOCKER="$C/pin/bin/docker" DOCKER_CONFIG="$C/pin/cfg")
  s="$C/s9b"; : >"$s"
  vc_sh "$s" "export VC_FX_N=\"\$WORK/fx.n\" VC_FX_OUT=$(printf '%q' "$C/fx9-@N@.out") VC_FX_ERR=$(printf '%q' "$C/fx9-@N@.err") VC_RC=0"
  vc_sh "$s" "compose $(printf '%q' "$C/cd9") build one"; vc_sh "$s" "compose $(printf '%q' "$C/cd9") build two three"
  vc_sh "$s" "ledger_run_block $(printf '%q' "$C/cd9") $(printf '%q' "$blk") \"\$WORK/block.log\""
  vc_sh "$s" "unset VC_FX_ERR VC_FX_OUT; export VC_DFX=$(printf '%q' "$w/direct.log") VC_DIID=$(cat "$w/direct.iid" 2>/dev/null)"
  vc_build "$s" "$C/c9/one" "$(vc_tag direct)"
  vc_run "$C/w9b" "$s"
  vc_want 6 MET
  for n in 1 2 3 4 5 6; do
    if [[ "$(awk -F'\t' -v r="$n" 'NR - 1 == r' "$C/w9a/builds.tsv" | cut -f4-19)" != "$(awk -F'\t' -v r="$n" 'NR - 1 == r' "$C/w9b/builds.tsv" | cut -f4-19)" ]]; then
      prob "B row $n differs from the real row (kind, key, frontend, bases, platform, driver, progress, counts)"
    fi
  done
  r+=" B: replay identical, MET;"
  # C: derived defects (one stream: call 1's; block: calls 3 and 4).
  VC_ENV=(PATH="$C/vc9shim:$ROOT/shim:$PATH" VC_NEXT="$ROOT/shim/docker" SHIM_REAL_DOCKER="$C/pin/bin/docker" DOCKER_CONFIG="$C/pin/cfg")
  printf 'Name:   vc9-a\nDriver: docker\n\nNodes:\nName: vc9-a0\n' >"$C/bld-1.txt"; printf 'Name:   vc9-b\nDriver: docker-container\n\nNodes:\nName: vc9-b0\n' >"$C/bld-2.txt"
  printf 'Name:   default\nDriver: docker\n\nNodes:\nName: default\n' >"$C/bld-d.txt"
  for e in two-names unbuilt-name same-twice builder-unavailable builder-changes header no-terminal no-export absent truncated no-bake-vertex mixed conflict alias-op alias-stage alias-index alias-mixed \
    lone-text lone-target lone-mixed lone-reserved; do
    s="$C/s9c-$e"; : >"$s"; d="$C/cd9"; k="build one"; b=""
    cp -- "$C/fx9-1.err" "$C/fxc-$e-1.err"
    case "$e" in
      two-names) d="$C/cd9dup"; k="build one dup"; cp -- "$C/fx9-1.out" "$C/fxc-$e-1.out" ;;
      unbuilt-name|no-terminal|no-export|truncated|no-bake-vertex|header|conflict) vc9_fx "$e" "$C/fx9-1.out" "$C/fxc-$e-1.out" ;;
      same-twice) for n in 1 2; do cp -- "$C/fx9-4.out" "$C/fxc-$e-$n.out"; cp -- "$C/fx9-4.err" "$C/fxc-$e-$n.err"; done ;;
      builder-unavailable) cp -- "$C/fx9-1.out" "$C/fxc-$e-1.out"; b=fail ;;
      builder-changes) cp -- "$C/fx9-1.out" "$C/fxc-$e-1.out"; b="$C/bld-@N@.txt" ;;
      absent) k="build absent"; vc9_fx absent "$C/fx9-1.out" "$C/fxc-$e-1.out" ;;
      mixed) k="build two three"; vc9_fx mixed "$C/fx9-2.out" "$C/fxc-$e-1.out"; cp -- "$C/fx9-2.err" "$C/fxc-$e-1.err" ;;
      alias-*) k="build two three"; vc9_alias "${e#alias-}" "$C/fx9-2.out" "$C/fxc-$e-1.out"; cp -- "$C/fx9-2.err" "$C/fxc-$e-1.err" ;;
      lone-*) k="build two three"; vc9_lone "${e#lone-}" "$C/fx9-2.out" "$C/fxc-$e-1.out" "vc9-planted-$(vc_tag "$e")"; cp -- "$C/fx9-2.err" "$C/fxc-$e-1.err" ;;
    esac
    vc_sh "$s" "export VC_FX_N=\"\$WORK/fx.n\" VC_BLD_N=\"\$WORK/bld.n\" VC_FX_OUT=$(printf '%q' "$C/fxc-$e-@N@.out") VC_FX_ERR=$(printf '%q' "$C/fxc-$e-@N@.err") VC_RC=0 VC_BLD=$(printf '%q' "${b:-$C/bld-d.txt}")"
    if [[ "$e" == same-twice ]]; then vc_sh "$s" "ledger_run_block $(printf '%q' "$d") $(printf '%q' "$blk") \"\$WORK/block.log\" || true"
    else vc_sh "$s" "compose $(printf '%q' "$d") $k || true"; fi
    vc_run "$C/w9c-$e" "$s"
    case "$e" in
      two-names) vc9_unmet "$e" 3 ;;       # the unattributable session's unknown row, one and dup missing
      unbuilt-name|no-export|truncated) vc9_unmet "$e" 2 ;;  # the unattributable session's unknown row, one missing
      no-bake-vertex) vc9_unmet "$e" 1 ;;  # no session at all: one missing
      same-twice) vc9_unmet "$e" 3 ;;      # both sessions are one's (two rows); four missing
      alias-mixed|lone-target) vc9_unmet "$e" 3 ;;  # two, three and the stage name or unbuilt service read as a service
      mixed|alias-*|lone-*) vc9_unmet "$e" 2 ;;
      *) vc9_unmet "$e" 1 ;;
    esac
    case "$e" in
      lone-text) vc9_context "$e" "vc9-planted-$(vc_tag "$e")" "[three] docker-image://docker.io/docker/dockerfile:1@sha256:" ;;
      lone-target) vc9_context "$e" "vc9-planted-$(vc_tag "$e")" "[four] resolve image config" ;;
      lone-mixed) vc9_context "$e" "vc9-planted-$(vc_tag "$e")" "resolve image config for docker-image://docker.io/docker/dockerfile:1" ;;
      lone-reserved) vc9_context "$e" "vc9-planted-$(vc_tag "$e")" "[auth] resolve image config" ;;
    esac
    r+=" $e=$(vc_s expected)/$(vc_s unknown)/$(vc_s ac5_status)"
  done
  # D: re-printed build definitions (one service; two targets sharing one definition): each vertex counts once, MET.
  vc9_fx reprint "$C/fx9-1.out" "$C/fxd-1.out"; vc9_fx reprint "$C/fx9-2.out" "$C/fxd-2.out"
  cp -- "$C/fx9-1.err" "$C/fxd-1.err"; cp -- "$C/fx9-2.err" "$C/fxd-2.err"
  s="$C/s9d"; : >"$s"
  vc_sh "$s" "export VC_FX_N=\"\$WORK/fx.n\" VC_BLD_N=\"\$WORK/bld.n\" VC_FX_OUT=$(printf '%q' "$C/fxd-@N@.out") VC_FX_ERR=$(printf '%q' "$C/fxd-@N@.err") VC_RC=0 VC_BLD=$(printf '%q' "$C/bld-d.txt")"
  vc_sh "$s" "compose $(printf '%q' "$C/cd9") build one"; vc_sh "$s" "compose $(printf '%q' "$C/cd9") build two three"
  vc_run "$C/w9d" "$s"
  vc_want 3 MET
  for n in 1 2 3; do vc_parsed "$n" "$(vc_f "$n" steps_executed | grep -qx '0' && echo 0 || echo +)"; done
  if [[ "$(grep -c 'load build definition from' "$C/fxd-1.out")" != 2 ]]; then prob "D: the derived stream does not re-print the build definition"; fi
  r+=" D: re-printed definitions $(vc_s expected)/$(vc_s unknown)/$(vc_s ac5_status);"
  # A's call-2 format check on derived streams: the re-printed shared definition is one definition; the same vertex
  # re-printed under the other service, an added unprefixed definition or a second prefixed definition vertex is not.
  if [[ "$(grep -c 'load build definition from' "$C/fxd-2.out")" != 2 ]]; then prob "D: the derived call-2 stream does not re-print the build definition"; fi
  if ! vc9_one_def "$C/fxd-2.out"; then prob "D: the re-printed shared definition is not one definition [$(vc9_defs "$C/fxd-2.out" | tr '\n' '|')]"; fi
  awk '/^#[0-9]+ \[(two|three) internal\] load build definition from / && ++k == 2 { if (!sub(/\[two internal\]/, "[three internal]")) sub(/\[three internal\]/, "[two internal]") } { print }' \
    "$C/fxd-2.out" >"$C/fxa-conflict.out"
  vc9_fx mixed "$C/fx9-2.out" "$C/fxa-mixed.out"
  { cat -- "$C/fx9-2.out"; printf '\n#98 [two internal] load build definition from Dockerfile\n#98 DONE 0.0s\n'; } >"$C/fxa-second.out"
  for f in conflict mixed second; do if vc9_one_def "$C/fxa-$f.out"; then prob "D: A's call-2 check accepts the $f build definition"; fi; done
  r+=" A call-2 check: re-print one definition, conflict/mixed/second rejected;"
  # E: a shared step re-printed under the other service (a valid alias): both rows parsed, MET; the aliased block keeps
  # its own service row, the step counted once (executed by the first, cached for the other).
  vc9_alias alias "$C/fx9-2.out" "$C/fxe-1.out"; cp -- "$C/fx9-2.err" "$C/fxe-1.err"
  s="$C/s9e"; : >"$s"
  vc_sh "$s" "export VC_FX_N=\"\$WORK/fx.n\" VC_BLD_N=\"\$WORK/bld.n\" VC_FX_OUT=$(printf '%q' "$C/fxe-@N@.out") VC_FX_ERR=$(printf '%q' "$C/fxe-@N@.err") VC_RC=0 VC_BLD=$(printf '%q' "$C/bld-d.txt")"
  vc_sh "$s" "compose $(printf '%q' "$C/cd9") build two three"
  vc_run "$C/w9e" "$s"
  vc_want 2 MET
  if [[ "$(grep -cE '^#[0-9]+ \[(two|three) [a-z0-9-]+ [0-9]+/[0-9]+\] ' "$C/fxe-1.out")" -le "$(grep -cE '^#[0-9]+ \[(two|three) [a-z0-9-]+ [0-9]+/[0-9]+\] ' "$C/fx9-2.out")" ]]; then prob "E: the derived stream does not re-print a shared step"; fi
  r+=" E: service-prefix alias $(vc_s expected)/$(vc_s unknown)/$(vc_s ac5_status);"
  # E2 (G051 C6): the lone-label frontend vertices re-printed under the other bound target: both rows parsed, MET, each
  # block keeping its own service row; no unknown_context.
  vc9_lone alias "$C/fx9-2.out" "$C/fxe2-1.out" "vc9-planted-$(vc_tag e2)"; cp -- "$C/fx9-2.err" "$C/fxe2-1.err"
  s="$C/s9e2"; : >"$s"
  vc_sh "$s" "export VC_FX_N=\"\$WORK/fx.n\" VC_BLD_N=\"\$WORK/bld.n\" VC_FX_OUT=$(printf '%q' "$C/fxe2-@N@.out") VC_FX_ERR=$(printf '%q' "$C/fxe2-@N@.err") VC_RC=0 VC_BLD=$(printf '%q' "$C/bld-d.txt")"
  vc_sh "$s" "compose $(printf '%q' "$C/cd9") build two three"
  vc_run "$C/w9e2" "$s"
  vc_want 2 MET
  if [[ "$(grep -c '^#9[01] \[three\] ' "$C/fxe2-1.out")" != 2 || "$(grep -c '^#9[01] \[two\] ' "$C/fxe2-1.out")" != 2 ]]; then prob "E2: the derived stream does not print both lone-label vertices under two and three"; fi
  if grep -q '^unknown_context' "$VW/ledger-summary.tsv"; then prob "E2: a parsed call left unknown_context lines"; fi
  r+=" E2: lone-label frontend alias $(vc_s expected)/$(vc_s unknown)/$(vc_s ac5_status);"
  VC_ENV=()
  vc_clean
  detail "pinned hosted Compose Bake progress through the real ledger:$r"
}

# --- group E: storage report and interrupted-run recovery (G051 cleanup amendment; VS1-VS5, VR1-VR9) -------------
# validate.sh observes Docker storage read-only before the harness and after cleanup (raw/storage.tsv: scope, time,
# job, candidate, unit and status on every row; "unavailable", never zero, when an observation fails) and never lets
# the report change a run's status. A local run marks itself in STATE_ROOT/g003-pending (G064 D3; the case state root,
# case_state); the next local run recovers an
# unfinalized run through cleanup-owned.sh on its registry, or blocks (16) and keeps the marker.
ST_HDR=$'observed_utc\tphase\tjob\tcandidate\tscope\tmetric\tvalue\tunit\tstatus'
expect_storage() { # FILE ok|unavailable
  local f="$1" ph n bad
  if [[ -L "$f" || ! -f "$f" ]]; then prob "no storage report $f"; return 0; fi
  if [[ "$(head -1 "$f")" != "$ST_HDR" ]]; then prob "storage.tsv header [$(head -1 "$f" | tr '\t' ' ')]"; return 0; fi
  for ph in pre post; do
    n="$(awk -F'\t' -v p="$ph" 'NR > 1 && $2 == p && $5 ~ /^daemon/' "$f" | grep -c .)" || n=0
    if [[ "$n" != 6 ]]; then prob "storage $ph: $n daemon and builder rows, want 6"; fi
  done
  if awk -F'\t' 'NR > 1 && (NF != 9 || $1 !~ /^[0-9-]+T[0-9:]+Z$/ || $3 !~ /^(local-[0-9a-f]+|[0-9]+)$/ || $4 !~ /^[0-9a-f]+$/ || length($4) != 40 || $5 == "" || $8 == "" || $9 == "") {f = 1} END {exit !f}' "$f"; then
    prob "a storage row lacks its time, job, candidate, scope, unit or status"
  fi
  case "$2" in
    ok) # G063 repair: the offending rows (phase:metric=value[status]) travel in the problem, so a failure is attributable.
        bad="$(awk -F'\t' 'NR > 1 && $5 ~ /^daemon/ && ($9 != "ok" || $7 == "unavailable") { print $2 ":" $6 "=" $7 "[" $9 "]" }' "$f" | LC_ALL=C sort -u | tr '\n' ' ')"
        if [[ -n "$bad" ]]; then prob "an available observation is not ok: ${bad% }"; fi
        bad="$(awk -F'\t' 'NR > 1 && $5 ~ /^daemon/ && $8 != "count" && $7 !~ /^[0-9.]+[kMGTP]?B$/ { print $2 ":" $6 "=" $7 "[" $9 "]" }' "$f" | LC_ALL=C sort -u | tr '\n' ' ')"
        if [[ -n "$bad" ]]; then prob "a size is not a Docker CLI size: ${bad% }"; fi ;;
    unavailable) if awk -F'\t' 'NR > 1 && $5 ~ /^daemon/ && ($9 !~ /^unavailable:/ || $7 != "unavailable")' "$f" | grep -q .; then prob "a failed observation is not explicitly unavailable"; fi ;;
  esac
  if [[ "$(awk -F'\t' 'NR > 1 && $2 == "post" && ($6 == "cleanup_status" || $6 == "owned_remaining")' "$f" | grep -c .)" != 2 ]]; then
    prob "post lacks cleanup_status and owned_remaining"
  fi
}
st_val() { awk -F'\t' -v p="$2" -v m="$3" 'NR > 1 && $2 == p && $6 == m { print $7; exit }' "$1" 2>/dev/null; }
pend_marker() { case_state; compgen -G "$CSTATE/g003-pending/local-*.tsv" | head -1 || true; } # G064 D3: the state root
# Local report retention must survive the real wrapper's final scrub, including on failure.
ctl_LR1() {
  local f
  new_case; run_v STUB_HARNESS=fail STUB_FAIL_ID=B-CI-CONTROLS -- --evidence-dir "$C/ev"
  expect_rc 1; expect_phase raw_scan 0; expect_phase candidate_scan 0; expect_phase scrub 0
  expect_phase cleanup 0; expect_pub none; expect_runs 1
  for f in harness.log results.tsv ci-controls.log junit.xml summary.json phase-status.tsv; do
    if [[ ! -s "$C/ev/candidate/$f" ]]; then prob "missing retained failure artifact $f"; fi
  done
  if ! grep -q '^FAIL LR-PROBE retained failure detail$' "$C/ev/candidate/ci-controls.log"; then prob "the underlying failing control was lost"; fi
  if ! grep -q $'^B-CI-CONTROLS\tFAIL\t' "$C/ev/candidate/results.tsv"; then prob "the failing check ledger was lost"; fi
  if ! grep -q 'name="B-CI-CONTROLS"' "$C/ev/candidate/junit.xml"; then prob "the failed check is absent from JUnit"; fi
  for f in raw staged staged.tmp final state scan owned; do
    if [[ -e "$C/ev/$f" ]]; then prob "temporary or secret-bearing directory $f survived"; fi
  done
  log_has "evidence: scanned local report retained at $C/ev/candidate (logs, check ledger and supplemental JUnit when available)"
  detail "failed local wrapper exit 1 preserved; scanned harness log, underlying control detail, ledger and JUnit survive; raw evidence, state, secrets and owned resources scrubbed"
}
ctl_LR2() {
  new_case; run_v STUB_HARNESS=plant-retained -- --evidence-dir "$C/ev"
  expect_rc 15; expect_phase raw_scan 15; expect_phase scrub 0
  if [[ ! -s "$C/ev/candidate/summary.json" ]]; then prob "the safe withheld summary was not retained"; fi
  if [[ -e "$C/ev/candidate/harness.log" || -e "$C/ev/candidate/plan.out" || -e "$C/ev/raw" ]]; then prob "unsafe/raw evidence survived withholding"; fi
  if grep -rFq -- "$(cat "$C/stub/canary")" "$C/ev"; then prob "a canary survived in retained local evidence"; fi
  detail "unsafe local evidence exit 15 preserved; only scanned minimal candidate retained, no canary or raw diagnostic logs"
}
ctl_LR3() {
  new_case; run_v SHIM_GITLEAKS_FAIL=1 -- --evidence-dir "$C/ev"
  expect_rc 17; expect_phase raw_scan 17; expect_phase scrub 0
  if [[ -e "$C/ev/candidate" || -e "$C/ev/raw" ]]; then prob "unscanned evidence survived a scanner failure"; fi
  detail "scanner failure exit 17 preserved; no unscanned candidate or raw evidence retained"
}
ctl_VS1() {
  new_case; run_v -- --evidence-dir "$C/ev"
  expect_rc 0; expect_phase recovery 0; expect_phase cleanup 0
  expect_storage "$C/ev/storage.tsv" ok
  if [[ "$(st_val "$C/ev/storage.tsv" post owned_remaining)" != 0 || "$(st_val "$C/ev/storage.tsv" post cleanup_status)" != 0 ]]; then
    prob "post: cleanup_status or owned_remaining is not 0 after a verified cleanup"
  fi
  if grep -q '^validate: storage pre: images_count=' "$C/out" && grep -q '^validate: storage post: images_count=' "$C/out"; then :; else prob "no storage summary lines"; fi
  if grep -q '^validate: storage: daemon-wide and builder-wide observations, not project usage; report .*/ev/storage.tsv$' "$C/out"; then :; else prob "the summary does not name the kept report"; fi
  if [[ -n "$(pend_marker)" ]]; then prob "a run marker was left after a verified cleanup"; fi
  if [[ ! -s "$C/ev/candidate/junit.xml" || ! -s "$C/ev/candidate/results.tsv" ]]; then prob "successful local run lost its scanned report"; fi
  detail "exit $RC; local run: pre and post observations (daemon images, builder cache, cleanup, owned left 0) kept in --evidence-dir; summary names it; no marker left"
}
ctl_VS2() {
  std_ci; run_v "${CIENV[@]}" STUB_HARNESS=fail -- --publish-dir ci-evidence
  expect_rc 1; expect_line has harness=1; expect_pub full; expect_line not recovery=
  expect_storage "$(PUBDIR)/storage.tsv" ok
  if [[ "$(st_val "$(PUBDIR)/storage.tsv" post cleanup_status)" != 0 ]]; then prob "post cleanup_status is not 0"; fi
  log_has "storage: daemon-wide and builder-wide observations, not project usage; report ci-evidence/storage.tsv when published"
  detail "exit $RC (harness failure kept); storage observed before and after cleanup and published; CI has no recovery phase"
}
ctl_VS3() {
  local m
  new_case; start_v STUB_HARNESS=sleep -- --evidence-dir "$C/ev"
  if wait_file "$C/stub/harness-running" 60; then kill -TERM "$VPID"; else prob "the stub harness never started"; fi
  wait_v
  expect_rc 143; expect_line has signal=143; expect_stop signal; expect_phase cleanup 0; expect_group_gone
  expect_storage "$C/ev/storage.tsv" ok
  m="$(pend_marker)"; if [[ -n "$m" ]]; then prob "a run marker was left after an interrupted run cleaned up"; fi
  detail "exit $RC (TERM kept); harness group stopped and reaped before cleanup; cleanup verified; storage reported after cleanup"
}
ctl_VS4() {
  local r1 n
  std_ci; run_v "${CIENV[@]}" SHIM_FAIL_DF=1 -- --publish-dir ci-evidence
  expect_rc 0; expect_pub full; expect_storage "$(PUBDIR)/storage.tsv" unavailable; r1="$RC"
  if grep -q '^validate: storage pre: images_count=unavailable ' "$C/out"; then :; else prob "the summary does not show the unavailable observation"; fi
  # G063 repair: the retry is bounded; a persistent failure is queried exactly three times per phase (pre and post).
  n="$(grep -c '^system df ' "$C/shim.log" 2>/dev/null)" || n=0
  if [[ "$n" != 6 ]]; then prob "a persistently failing docker system df was queried $n times, want 6 (3 per phase)"; fi
  std_ci 4243; run_v "${CIENV[@]}" SHIM_FAIL_DF=1 STUB_HARNESS=fail -- --publish-dir ci-evidence
  expect_rc 1; expect_storage "$(PUBDIR)/storage.tsv" unavailable
  detail "docker system df failing: every observation unavailable (no zero), status unavailable:docker-system-df-exit-1 after exactly 3 bounded attempts per phase; pass exit $r1, failing harness exit $RC (not masked)"
}
# G063 repair (a post-merge full qualification run, VS3 read unavailable). Unproven hypothesis for that failure: a
# read-only docker system df on the shared daemon failed once while the foreground harness removed an image. VS5 does
# not reproduce any daemon race; the shim injects a synthetic transient docker system df failure
# (SHIM_FAIL_DF=transient: odd-numbered queries fail) and proves that the bounded retry (validate.sh storage_observe)
# records the next, passing observation instead of unavailable. A persistent failure stays unavailable (VS4).
ctl_VS5() {
  local n
  std_ci; run_v "${CIENV[@]}" SHIM_FAIL_DF=transient SHIM_DF_COUNT="$C/df-count" -- --publish-dir ci-evidence
  expect_rc 0; expect_pub full; expect_storage "$(PUBDIR)/storage.tsv" ok
  if grep -q '^validate: storage pre: images_count=unavailable \|^validate: storage post: images_count=unavailable ' "$C/out"; then prob "a transient failure was reported unavailable"; fi
  n="$(grep -c '^system df ' "$C/shim.log" 2>/dev/null)" || n=0
  if [[ "$n" != 4 ]]; then prob "docker system df was queried $n times, want 4 (one failed and one passing per phase)"; fi
  detail "exit $RC; an injected transient docker system df failure once per phase is retried once and both observations read ok with real Docker CLI figures; 4 queries (2 per phase), no unavailable row"
}
# G051 cycle 4: process survivors and descriptors, not only Docker resources. A crashed run's launch gets a
# case lock on descriptor 7, inherited by every process of that run (the watchdog closes it); fd_holders lists who
# still holds it. vr_ids reads the run's recorded harness group and watchdog from its marker.
fd_holders() { # FILE: the PIDs holding FILE open (from /proc), this shell excluded
  # find exits non-zero for other users' unreadable descriptor directories and vanished processes; that is not a failure.
  { find /proc/[0-9]*/fd -maxdepth 1 -lname "$1" 2>/dev/null || true; } | awk -F/ -v me="$$" '$3 != me { print $3 }' | sort -un | tr '\n' ' '
}
pg_members() { ps -eo pgid=,stat= | awk -v g="$1" '$1 == g && $2 !~ /^Z/ { f = 1 } END { exit f ? 0 : 1 }'; } # PGID: a live member
vr_ids() { # MARKER -> VR_STAGE VR_HG VR_HS VR_WP VR_WS
  VR_STAGE="$(awk -F'\t' '$1 == "stage" { print $2 }' "$1" 2>/dev/null)" || VR_STAGE=""
  VR_HG="$(awk -F'\t' '$1 == "harness_pgid" { print $2 }' "$1" 2>/dev/null)" || VR_HG=""
  VR_HS="$(awk -F'\t' '$1 == "harness_start" { print $2 }' "$1" 2>/dev/null)" || VR_HS=""
  VR_WP="$(awk -F'\t' '$1 == "watchdog_pid" { print $2 }' "$1" 2>/dev/null)" || VR_WP=""
  VR_WS="$(awk -F'\t' '$1 == "watchdog_start" { print $2 }' "$1" 2>/dev/null)" || VR_WS=""
}
vr_left() { # what of the crashed run is still alive or holds the case lock (empty: nothing)
  local s="" h
  if [[ "$VR_WP" =~ ^[0-9]+$ ]] && proc_is "$VR_WP" "$VR_WS"; then s+="watchdog $VR_WP; "; fi
  if [[ "$VR_WP" =~ ^[0-9]+$ ]] && pg_members "$VR_WP"; then s+="watchdog group $VR_WP; "; fi
  if [[ "$VR_HG" =~ ^[0-9]+$ ]] && pg_members "$VR_HG"; then s+="harness group $VR_HG; "; fi
  h="$(fd_holders "$C/case.lock")"; if [[ -n "$h" ]]; then s+="case-lock holders $h; "; fi
  printf '%s' "$s"
}
vr_settled() { # SECONDS: nothing of the crashed run survives (the watchdog ends by itself); survivors are stopped
  local i left=""
  for ((i = 0; i < $1 * 10; i++)); do left="$(vr_left)"; if [[ -z "$left" ]]; then return 0; fi; sleep 0.1; done
  prob "the KILLed run left processes or descriptors after $1 s: $left"
  stop_tree "" "${VR_WP:+$VR_WP:$VR_WS} ${VR_HG:+$VR_HG:$VR_HS}" || prob "survivors outlived KILL: $STOP_LEFT"
  return 1
}
vr_start() { # [MODE]: a local own-hold (or MODE) run whose processes inherit the case lock; waits until it holds
  : >"$C/case.lock"
  vcmd STUB_HARNESS="${1:-own-hold}" --
  launch "$C/out" "${VCMD[@]}" 7>"$C/case.lock"
  VPID="$LPID"
  if wait_file "$C/stub/harness-running" 60; then :; else prob "the stub harness never held"; fi
  VR_ID="$(cat "$C/stub/standalone-id" 2>/dev/null)" || VR_ID=""
  local i
  for ((i = 0; i < 100; i++)); do VR_M="$(pend_marker)"; vr_ids "$VR_M"; if [[ "$VR_STAGE" == launched ]]; then break; fi; sleep 0.1; done
  VR_REG="$(awk -F'\t' '$1 == "registry" { print $2 }' "$VR_M" 2>/dev/null)" || VR_REG=""
  VR_JOB="${VR_M##*/}"; VR_JOB="${VR_JOB%.tsv}"
  if [[ "$VR_STAGE" != launched || ! "$VR_HG" =~ ^[0-9]+$ || ! "$VR_WP" =~ ^[0-9]+$ ]]; then
    prob "the marker does not record the launched harness group and watchdog (stage [$VR_STAGE] harness [$VR_HG] watchdog [$VR_WP])"
  fi
}
vr_crash() { # [MODE]: a local run with one owned container, then validate.sh and its harness group KILLed (no finalization)
  local hg _r
  new_case; ctl_labels
  vr_start "${1:-own-hold}"
  read -r _r hg _r <"$C/stub/harness.pids" || hg=""
  if [[ "$hg" != "$VR_HG" ]]; then prob "the marker's harness group [$VR_HG] is not the stub's [$hg]"; fi
  kill -KILL "$VPID" 2>/dev/null || true
  if [[ "$hg" =~ ^[0-9]+$ ]]; then kill -KILL -- "-$hg" 2>/dev/null || true; fi
  await "$VPID" 30
  if [[ ! "$VR_ID" =~ ^[0-9a-f]{64}$ ]]; then prob "the stub recorded no owned container"; fi
  if [[ -z "$VR_M" || ! -d "$VR_REG" ]]; then prob "no unfinalized marker and registry after the crash"; fi
  if exists container "$VR_ID"; then :; else prob "the fixture's owned container did not outlive the crash"; fi
  if grep -q '^validate: final status ' "$C/out"; then prob "the KILLed run printed a final line"; fi
  vr_settled 5
}
ctl_VR1() {
  local row
  vr_crash
  run_v -- --evidence-dir "$C/ev2"
  expect_rc 0; expect_phase recovery 0; expect_runs 2
  want_gone container "$VR_ID"
  if [[ -e "$VR_M" ]]; then prob "the marker remains after a verified recovery"; fi
  row="$(tail -n 1 "$VR_REG/recovery.tsv" 2>/dev/null)"
  if [[ "$row" != "$VR_JOB"$'\t'"$VR_REG"$'\t'*$'\tdelayed-recovery-verified' ]]; then prob "the registry lacks the delayed-recovery record [$row]"; fi
  if [[ "$(tail -n 1 "$C/ev2/recovery.tsv" 2>/dev/null)" != "$row" ]]; then prob "the recovering run did not keep the recovery report"; fi
  if [[ -e "$VR_REG/secrets" ]]; then prob "the recovered run's secret list survived"; fi
  if grep -q "^validate: recovery: run $VR_JOB did not finalize; delayed recovery verified " "$C/out"; then :; else prob "the job log does not report the delayed recovery"; fi
  expect_storage "$C/ev2/storage.tsv" ok
  if [[ -n "$(vr_left)" ]]; then prob "after recovery: $(vr_left)"; fi
  detail "validate.sh and its harness KILLed (no final line, marker unfinalized); its watchdog ended by itself and no process or case-lock holder of the run remained; the next run recovered the owned container through cleanup-owned.sh, recorded the delayed recovery for $VR_JOB, then ran (exit $RC)"
}
ctl_VR2() {
  local dctr
  vr_crash
  printf '%032d\n' 0 >"$VR_REG/token"
  dctr="$(D create "${LB[@]}" --network none "$CI_BUSYBOX_IMAGE" true)"
  run_v --
  expect_rc 16; expect_phase recovery 16; expect_runs 1; expect_line not preflight=
  want_kept container "$VR_ID"; want_kept container "$dctr"
  if [[ ! -f "$VR_M" ]]; then prob "the marker was removed although recovery was refused"; fi
  if grep -q "^validate: recovery: run $VR_JOB left owned resources that cleanup did not verify removed (exit 18; " "$C/out"; then :; else prob "the job log does not name the refused recovery"; fi
  detail "crashed run left no process or case-lock holder; tampered registry: cleanup-owned.sh refused (exit 18); launch blocked (exit $RC), harness not started; the owned and an unrelated container kept; marker kept for the operator"
}
# G064 D3: hand-made markers go to the case state root (STATE_ROOT/g003-pending) with this boot's boot_id (32 hex, as
# validate.sh boot_id_now) and this case tree as repo; VR_MDIR=legacy plants a pre-change marker (no boot_id, no repo) in
# $TMPDIR/g003-pending instead, which validate.sh reads under the same-boot rules and never writes.
VR_MDIR="" VR_MF=""
vr_boot() { local b; b="$(cat /proc/sys/kernel/random/boot_id 2>/dev/null)" || b=""; printf '%s' "${b//-/}"; }
vr_marker() { # REGISTRY PID PID_START [ROWS]: a hand-made marker of an earlier run (ROWS: its stage and identities)
  local d prov=""
  case_state
  if [[ "$VR_MDIR" == legacy ]]; then d="$C/tmp/g003-pending"; else
    d="$CSTATE/g003-pending"; prov="$(printf 'boot_id\t%s\nrepo\t%s\n' "${VR_BOOT-$(vr_boot)}" "${VR_REPO-$(realpath -e "$C/repo")}")"$'\n'
  fi
  if [[ -d "$d" ]]; then :; else mkdir -m 700 "$d"; fi
  VR_MF="$d/local-0123456789ab.tsv"
  printf 'registry\t%s\nevidence\t%s\npid\t%s\npid_start\t%s\nstarted\t2026-01-01T00:00:00Z\n%s%b' "$1" "$C/gone" "$2" "$3" \
    "$prov" "${4:-stage\tprelaunch\n}" >"$VR_MF"
}
vr_blocked() { # REASON: launch blocked by the marker, which is kept
  expect_rc 16; expect_phase recovery 16; expect_runs 0
  if [[ ! -f "$VR_MF" ]]; then prob "the marker was removed"; fi
  if grep -qF "validate: recovery: $1" "$C/out"; then :; else prob "the job log lacks [recovery: $1]"; fi
}
ctl_VR3() {
  local p
  new_case; p="$(bash -c 'echo $$')"
  vr_marker "$C/no-registry" "$p" 1
  run_v --
  vr_blocked "the ownership registry of run local-0123456789ab ($C/no-registry) is missing"
  detail "exit $RC; a marker whose registry is missing blocks the launch; nothing guessed"
}
ctl_VR4() {
  local p
  new_case; p="$(bash -c 'echo $$')"
  vr_marker "$C/no-registry" "$p" 1
  run_v SHIM_DOCKER_DOWN=1 --
  vr_blocked "Docker is unavailable; the unfinalized run local-0123456789ab cannot be recovered"
  detail "exit $RC; Docker unavailable with an unfinalized run: launch blocked, marker kept"
}
ctl_VR5() {
  local d s
  new_case
  launch "$C/live.out" sleep 120; d="$LPID"; sleep 0.3
  s="$(proc_start "$d")" || s=""
  vr_marker "$C/no-registry" "$d" "$s"
  run_v --
  vr_blocked "local validation run local-0123456789ab (pid $d) is still running"
  if stop_tree "" "$d:$s"; then :; else prob "the live decoy survived KILL"; fi
  detail "exit $RC; a live run's marker (same PID and start time) blocks a second local run"
}
# VR6: only validate.sh is KILLed; its harness keeps running. The watchdog ends by itself, the next run is blocked
# (the harness still runs), nothing of it is signalled; after the operator stops that exact group, recovery proceeds.
ctl_VR6() {
  local hs i
  new_case; ctl_labels
  vr_start
  hs="$(proc_start "$VR_HG")" || hs=""
  if [[ -z "$hs" || "$hs" != "$VR_HS" ]]; then prob "the marker's harness start [$VR_HS] is not the live harness's [$hs]"; fi
  kill -KILL "$VPID" 2>/dev/null || true
  await "$VPID" 30
  for ((i = 0; i < 50; i++)); do if proc_is "$VR_WP" "$VR_WS" || pg_members "$VR_WP"; then sleep 0.1; else break; fi; done
  if proc_is "$VR_WP" "$VR_WS" || pg_members "$VR_WP"; then prob "the watchdog outlived its KILLed parent by 5 s"; fi
  if proc_is "$VR_HG" "$hs"; then :; else prob "the fixture harness did not outlive its parent"; fi
  if [[ -n "$(fd_holders "$C/case.lock")" ]]; then :; else prob "the surviving harness holds no case lock (fixture)"; fi
  run_v -- --evidence-dir "$C/ev2"
  expect_rc 16; expect_phase recovery 16; expect_runs 1; expect_line not preflight=
  if grep -qF "validate: recovery: the harness of run $VR_JOB (process group $VR_HG) is still running; " "$C/out"; then :; else prob "the job log does not name the running harness"; fi
  if proc_is "$VR_HG" "$hs"; then :; else prob "the blocked recovery stopped or signalled the surviving harness"; fi
  if [[ ! -f "$VR_M" ]]; then prob "the marker was removed although recovery was blocked"; fi
  want_kept container "$VR_ID"
  # Operator resolution: stop exactly that harness group; the next run then recovers the run's owned container.
  if stop_tree "" "$VR_HG:$hs"; then :; else prob "the fixture harness survived KILL: $STOP_LEFT"; fi
  if [[ -n "$(vr_left)" ]]; then prob "after the operator stop: $(vr_left)"; fi
  run_v -- --evidence-dir "$C/ev3"
  expect_rc 0; expect_phase recovery 0; expect_runs 2
  want_gone container "$VR_ID"
  if [[ -e "$VR_M" ]]; then prob "the marker remains after a verified recovery"; fi
  detail "only validate.sh KILLed: its watchdog ended by itself, the harness kept running and holding the case lock; the next run was blocked (16) without signalling it; after the operator stopped that group, recovery removed the owned container (exit $RC)"
}
# VR7: the watchdog of the case tree's validate.sh, run directly against disposable decoy parents.
vr7_decoy() { # NAME: a parent that records every USR1 and never ends by itself -> D7 D7S
  launch "$C/$1.out" bash -c 'trap "echo usr1 >>\"\$1\"" USR1; while :; do sleep 0.2; done' decoy "$C/$1.usr1"
  D7="$LPID"; sleep 0.3; D7S="$(proc_start "$D7")" || D7S=""
}
ctl_VR7() {
  local WATCHDOG_SH="" w h k da dsa db dsb
  new_case
  eval "$(sed -n "/^WATCHDOG_SH='\$/,/^'\$/p" "$C/repo/infra/ci/validate.sh")"
  if [[ -z "$WATCHDOG_SH" ]]; then prob "no WATCHDOG_SH in validate.sh"; return 0; fi
  # (a) Deadline: USR1 to the verified parent; no inherited descriptor kept; its only child is the timer.
  vr7_decoy a; da="$D7"; dsa="$D7S"
  : >"$C/wd.lock"
  launch "$C/wd-a.out" bash -c "$WATCHDOG_SH" watchdog 3 "$da" "$dsa" 7>"$C/wd.lock"
  w="$LPID"; sleep 0.8
  h="$(fd_holders "$C/wd.lock")"; if [[ -n "$h" ]]; then prob "(a) the watchdog kept an inherited descriptor (holders $h)"; fi
  k="$(ps -eo ppid=,args= | awk -v p="$w" '$1 == p { $1 = ""; print }' | tr '\n' ';')"
  if [[ "$k" != " sleep 3;" ]]; then prob "(a) the watchdog's children are [$k], want only its timer"; fi
  await "$w" 10
  if [[ "$RC" != 0 ]]; then prob "(a) watchdog exit $RC"; fi
  if wait_file "$C/a.usr1" 3 && [[ "$(grep -c usr1 "$C/a.usr1")" == 1 ]]; then :; else prob "(a) the verified parent did not get exactly one USR1"; fi
  if pg_members "$w"; then prob "(a) the watchdog's group outlived it"; fi
  # (b) Identity mismatch (a reused PID): the watchdog ends at once and signals nothing.
  vr7_decoy b; db="$D7"; dsb="$D7S"
  launch "$C/wd-b.out" bash -c "$WATCHDOG_SH" watchdog 2 "$db" "$((dsb + 1))"
  w="$LPID"; await "$w" 5
  if [[ "$RC" != 0 ]]; then prob "(b) watchdog exit $RC (did not end on the mismatch)"; fi
  sleep 3
  if [[ -e "$C/b.usr1" ]]; then prob "(b) a process with another start time was signalled"; fi
  if proc_is "$db" "$dsb"; then :; else prob "(b) the unrelated process did not survive"; fi
  if pg_members "$w"; then prob "(b) the watchdog's timer outlived it"; fi
  # (c) The parent is KILLed: the watchdog and its timer end within about a second.
  vr7_decoy c
  launch "$C/wd-c.out" bash -c "$WATCHDOG_SH" watchdog 600 "$D7" "$D7S"
  w="$LPID"; sleep 0.5; kill -KILL "$D7"
  await "$w" 5
  if [[ "$RC" != 0 ]]; then prob "(c) the watchdog outlived its KILLed parent (exit $RC)"; fi
  if pg_members "$w"; then prob "(c) the watchdog's timer outlived it"; fi
  if stop_tree "" "$da:$dsa $db:$dsb"; then :; else prob "the decoys survived KILL: $STOP_LEFT"; fi
  detail "watchdog: deadline USR1 only to the verified parent, no inherited descriptor, its only child the timer; a mismatched start time ends it without a signal; a KILLed parent ends it and its timer"
}
# VR8: the recorded harness/watchdog identities of a hand-made marker. Unrecorded, malformed, live or unattributable
# processes block (16, marker kept, nothing signalled); a reused PID outside the recorded group is ignored.
vr8_rows() { printf 'stage\tlaunched\nharness_pgid\t%s\nharness_start\t%s\nwatchdog_pid\t%s\nwatchdog_start\t%s\n' "$@"; }
ctl_VR8() {
  local p g gs l ls wd wds nd nds
  new_case; p="$(bash -c 'echo $$')"
  vr_marker "$C/no-registry" "$p" 1 'stage\tlaunching\n'
  run_v --
  vr_blocked "run local-0123456789ab ended while starting its harness, so its harness and watchdog were not recorded"
  vr_marker "$C/no-registry" "$p" 1 "$(vr8_rows x 1 2 3)"
  run_v --
  vr_blocked "the harness and watchdog identities in the marker of run local-0123456789ab are malformed"
  # A live harness leader with the recorded start time.
  launch "$C/hg.out" sleep 120; g="$LPID"; sleep 0.3; gs="$(proc_start "$g")" || gs=""
  vr_marker "$C/no-registry" "$p" 1 "$(vr8_rows "$g" "$gs" "$p" none)"
  run_v --
  vr_blocked "the harness of run local-0123456789ab (process group $g) is still running"
  if proc_is "$g" "$gs"; then :; else prob "the live harness decoy was signalled"; fi
  # A group whose leader has ended while a member lives: not attributable.
  launch "$C/orph.out" bash -c 'sleep 120 & sleep 0.3'; l="$LPID"; ls="$(proc_start "$l")" || ls=""; sleep 1
  vr_marker "$C/no-registry" "$p" 1 "$(vr8_rows "$l" "$ls" "$p" none)"
  run_v --
  vr_blocked "process group $l, run local-0123456789ab's harness group, still has members but its leader has ended"
  if pg_members "$l"; then :; else prob "the orphaned group member was signalled"; fi
  # A live watchdog with the recorded start time.
  launch "$C/wd.out" sleep 120; wd="$LPID"; sleep 0.3; wds="$(proc_start "$wd")" || wds=""
  vr_marker "$C/no-registry" "$p" 1 "$(vr8_rows "$p" none "$wd" "$wds")"
  run_v --
  vr_blocked "the deadline watchdog of run local-0123456789ab (pid $wd) is still running"
  if proc_is "$wd" "$wds"; then :; else prob "the live watchdog decoy was signalled"; fi
  # A reused PID: a live process outside the recorded group, with another start time, is not the run's.
  sleep 120 & nd=$!; disown "$nd"; sleep 0.3; nds="$(proc_start "$nd")" || nds=""
  vr_marker "$C/no-registry" "$p" 1 "$(vr8_rows "$nd" "$((nds + 1))" "$nd" "$((nds + 1))")"
  run_v --
  vr_blocked "the ownership registry of run local-0123456789ab ($C/no-registry) is missing"
  if proc_is "$nd" "$nds"; then :; else prob "the unrelated process with a reused PID was signalled"; fi
  if stop_tree "$nd:$nds" "$g:$gs $l $wd:$wds"; then :; else prob "the decoys survived KILL: $STOP_LEFT"; fi
  detail "launching stage, malformed identities, a live harness group, a leaderless group with a live member and a live watchdog each block (16, marker kept, nothing signalled); a reused PID outside the group is ignored (the run then stops at its missing registry)"
}
# VR9 (G064 Plan A corr1 C5, corr2 D4): the marker/recovery entry path of the batch runner and of a self-registered
# controls run, through the real case-tree validate.sh (case state root, case_state). Every recovering launch has a
# refusing stub preflight (STUB_PREFLIGHT=fail): it ends after the recovery phase (preflight=16), so neither a harness
# nor a new resource batch follows. Same-boot SIGKILL only; the boot_id rewrite and the planted markers prove the boot
# rule of do_recovery, never a reboot.
VR_BATCH="" VR9_N=0 VR9_T=""
vr9_sub() { VR9_T="$1"; VR9_N="$(grep -c . "$PROB" 2>/dev/null)" || VR9_N=0; } # TAG: the problems after this get TAG
vr9_tag() { # prefix this sub-case's problems with its tag
  local t="$PROB.vr9"
  if [[ -s "$PROB" ]]; then awk -v n="$VR9_N" -v t="$VR9_T" 'NR > n { $0 = t " " $0 } { print }' "$PROB" >"$t" && cat "$t" >"$PROB"; rm -f -- "$t"; fi
}
vr9_recover() { run_v STUB_PREFLIGHT=fail -- --evidence-dir "$1"; } # EVDIR: a recovering launch
# The g001.run containers of this control only (MAINT-rc14-vr9): every container a case launch creates goes through the
# shim with case_labels, so it carries sdlc.ci.control=$RT-$CTL_ID, this runner's and this control's exact identity. The
# daemon-wide g001.run population is not this control's: in the full profile ci-controls is a background child of run.sh
# while the other suites create and remove their own g001.run containers concurrently (RC14 hosted: 2 before, 3 after).
# Both label filters are ANDed by the CLI; a failed listing stays `unavailable` (a failure of (c)).
vr9_g001() { local q; q="$(D ps -aq --no-trunc --filter label=g001.run --filter "label=sdlc.ci.control=$RT-$CTL_ID")" || { printf 'unavailable'; return 0; }; grep -c . <<<"$q" || true; }
vr9_crash() { # (a)'s crash: the runner-hold stub's runner and its exec'd process outlive validate.sh and its harness group
  local t
  vr_crash runner-hold
  VR_BATCH="$(cat "$C/stub/runner-batch" 2>/dev/null)" || VR_BATCH=""
  if [[ -s "$C/stub/runner-why" ]]; then prob "the runner-hold stub: $(head -1 "$C/stub/runner-why")"; fi
  if awk -F'\t' -v i="$VR_ID" '$1 == "standalone" && $2 == i { f = 1 } END { exit !f }' "$VR_REG/owned.tsv" 2>/dev/null; then :; else
    prob "the run registry has no standalone row for the runner (docker_run_owned)"
  fi
  # The daemon refuses a ps format without a PID column (`-eo args`); as lib.sh docker_top_args: rows with a numeric
  # first field of `-eo pid,args`, PID stripped.
  t="$(D top "$VR_ID" -eo pid,args 2>/dev/null)" || t=""
  t="$(awk '$1 ~ /^[0-9]+$/ { sub(/^[ \t]*[0-9]+[ \t]*/, ""); sub(/[ \t]+$/, ""); print }' <<<"$t")"
  if grep -qx 'sleep 300' <<<"$t"; then :; else prob "the process exec'd into the runner did not outlive the crash"; fi
}
vr9_recovered() { # the crashed run's marker removed after a verified delayed recovery recorded in its registry
  local row
  if [[ -e "$VR_M" ]]; then prob "the marker remains after the recovery"; fi
  row="$(tail -n 1 "$VR_REG/recovery.tsv" 2>/dev/null)" || row=""
  if [[ "$row" != "$VR_JOB"$'\t'"$VR_REG"$'\t'*$'\tdelayed-recovery-verified' ]]; then prob "the registry lacks the delayed-recovery record [$row]"; fi
  if grep -q "^validate: recovery: run $VR_JOB did not finalize; delayed recovery verified " "$C/out"; then :; else prob "the job log does not report the delayed recovery"; fi
  if [[ -n "$(vr_left)" ]]; then prob "after the recovery: $(vr_left)"; fi
}
vr9_blocked_kept() { # REASON ID: a crashed run's recovery refused (16, no harness); its marker and runner kept, then removed
  expect_rc 16; expect_phase recovery 16; expect_runs 1; expect_line not preflight=
  if grep -qF "validate: recovery: $1" "$C/out"; then :; else prob "the job log lacks [recovery: $1]"; fi
  if [[ -e "$VR_M" ]]; then :; else prob "the marker was removed although recovery was refused"; fi
  want_kept container "$2"
  D rm -f "$2" >/dev/null 2>&1 || true
  want_gone container "$2"
}
ctl_VR9() {
  local rid q n0 n1 p cid reg g gs ctr ctr2 nb=0123456789abcdef0123456789abcdef
  # (a) The batch runner (it_runner_start through the run's marker and registry) and a process exec'd into it outlive
  # the KILL of validate.sh and its harness group; the next launch verifies the dead group and watchdog
  # (recovery_survivors), removes the runner through cleanup-owned.sh and records the delayed recovery.
  vr9_sub "(a)"; vr9_crash; rid="$VR_ID"
  vr9_recover "$C/ev2"
  expect_rc 16; expect_phase recovery 0; expect_phase preflight 16; expect_runs 1
  want_gone container "$rid"
  if [[ "$VR_BATCH" =~ ^vr9-[0-9a-f]{6}$ ]]; then
    sleep 6
    if q="$(D ps -aq --no-trunc --filter "label=g064.batch=$VR_BATCH")"; then
      if [[ -n "$q" ]]; then prob "the g064.batch=$VR_BATCH query is not empty 6 s after the recovery: $(tr '\n' ' ' <<<"$q")"; fi
    else prob "the g064.batch=$VR_BATCH query failed"; fi
  else prob "no runner batch recorded [$VR_BATCH]"; fi
  vr9_recovered; vr9_tag
  # (b) A self-registered top-level validate-controls.sh (its own setsid group; registry and marker under the case state
  # root) holding one container created through D, KILLed with its group: recovered by the same do_recovery.
  vr9_sub "(b)"; new_case; case_state
  cp -- "$REPO_R/infra/ci/validate-controls.sh" "$REPO_R/infra/ci/docker-desktop-mounts.sh" "$C/repo/infra/ci/" || prob "the controls runner was not copied"
  if git -C "$C/repo" add -A && git -C "$C/repo" -c user.name=ci-controls -c user.email=controls@example.invalid commit -q -m "controls runner"; then :; else
    prob "the case tree commit failed"
  fi
  launch "$C/rh.out" env -i -C "$C/repo" "${BASE_ENV[@]}" G003_STATE_ROOT="$CSTATE" CI_CONTROLS_STATE_BASE="$CSTATE" TMPDIR="$C/tmp" \
    SHIM_LABELS="$(case_labels)" CI_CONTROLS_REG_HOLD_OUT="$C/rh" bash "$C/repo/infra/ci/validate-controls.sh" --only reg-hold
  p="$LPID"
  if wait_file "$C/rh" 120; then :; else prob "the self-registered run never held: $(tail -n 3 "$C/rh.out" | tr '\n' ' ')"; fi
  read -r cid reg <"$C/rh" 2>/dev/null || { cid=""; reg=""; }
  VR_M="$(pend_marker)"; vr_ids "$VR_M"
  VR_JOB="${VR_M##*/}"; VR_JOB="${VR_JOB%.tsv}"; VR_REG="$reg"
  if [[ "$VR_STAGE" != launched || "$VR_HG" != "$p" || "$VR_WP" != "$p" ]]; then
    prob "the marker does not record the run as its own launched group (stage [$VR_STAGE] harness [$VR_HG] watchdog [$VR_WP], run $p)"
  fi
  if [[ "$(awk -F'\t' '$1 == "registry" { print $2 }' "$VR_M" 2>/dev/null)" != "$reg" || ! -d "$reg" ]]; then prob "the marker does not name the run registry [$reg]"; fi
  if awk -F'\t' -v i="$cid" '$1 == "standalone" && $2 == i { f = 1 } END { exit !f }' "$reg/owned.tsv" 2>/dev/null; then :; else
    prob "the run registry has no standalone row for the D container [${cid:0:12}] (registry mirror)"
  fi
  kill -KILL -- "-$p" 2>/dev/null || true
  await "$p" 30
  vr_settled 5
  if [[ "$cid" =~ ^[0-9a-f]{64}$ ]] && exists container "$cid"; then :; else prob "the D container [${cid:0:12}] did not outlive the KILL"; fi
  if [[ -f "$VR_M" && -d "$reg" ]]; then :; else prob "no marker and registry after the KILL"; fi
  vr9_recover "$C/ev2"
  expect_rc 16; expect_phase recovery 0; expect_phase preflight 16; expect_runs 0
  if [[ -n "$cid" ]]; then want_gone container "$cid"; fi
  vr9_recovered; vr9_tag
  # (c) Crash, then a tampered registry token: cleanup-owned.sh refuses, the launch blocks (16) with no harness, so no new
  # resource batch: the runner is kept and this control's g001.run container count (vr9_g001) is unchanged.
  vr9_sub "(c)"; vr9_crash; rid="$VR_ID"
  printf '%032d\n' 0 >"$VR_REG/token" || prob "the registry token could not be tampered"
  n0="$(vr9_g001)"
  vr9_recover "$C/ev2"
  n1="$(vr9_g001)"
  if [[ "$n0" == unavailable || "$n0" != "$n1" ]]; then prob "g001.run containers of this control $n0 before and $n1 after the refused recovery"; fi
  vr9_blocked_kept "run $VR_JOB left owned resources that cleanup did not verify removed (exit 18; " "$rid"; vr9_tag
  # (d) Crash, then the marker's boot_id rewritten: the owner is dead by definition, the process checks are skipped and
  # the runner is recovered.
  vr9_sub "(d)"; vr9_crash; rid="$VR_ID"
  sed -i "s/^boot_id\t.*/boot_id\t$nb/" "$VR_M" || prob "the marker's boot_id could not be rewritten"
  vr9_recover "$C/ev2"
  expect_rc 16; expect_phase recovery 0; expect_phase preflight 16; expect_runs 1
  if grep -qF "validate: recovery: run $VR_JOB was started in an earlier boot ($nb); its processes ended with that boot, so the process checks are skipped" "$C/out"; then :; else
    prob "the job log does not report the earlier boot"
  fi
  want_gone container "$rid"
  vr9_recovered; vr9_tag
  # (e) Crash, then the registry deleted: 16, no harness, the surviving labelled runner untouched.
  vr9_sub "(e)"; vr9_crash; rid="$VR_ID"
  rm -rf -- "$VR_REG"
  vr9_recover "$C/ev2"
  vr9_blocked_kept "the ownership registry of run $VR_JOB ($VR_REG) is missing; operator resolution required (marker kept)" "$rid"; vr9_tag
  # (f) Crash, then the marker mode 000: 16, no harness, the runner untouched; readable again, it is recovered.
  vr9_sub "(f)"; vr9_crash; rid="$VR_ID"
  chmod 000 "$VR_M" || prob "the marker mode could not be changed"
  vr9_recover "$C/ev2"
  expect_rc 16; expect_phase recovery 16; expect_runs 1; expect_line not preflight=
  if grep -qF "validate: recovery: the marker of run $VR_JOB ($VR_M) is unreadable; operator resolution required (marker kept)" "$C/out"; then :; else
    prob "the job log does not report the unreadable marker"
  fi
  if [[ -e "$VR_M" ]]; then :; else prob "the unreadable marker was removed"; fi
  want_kept container "$rid"
  chmod 600 "$VR_M" || prob "the marker mode could not be restored"
  vr9_recover "$C/ev3"
  expect_rc 16; expect_phase recovery 0; expect_runs 1
  want_gone container "$rid"
  vr9_recovered; vr9_tag
  # (g) Planted markers over a registry with one bound container: a malformed boot_id or another checkout blocks (16,
  # marker kept); another boot_id whose harness group is a live process of this boot skips the process checks and is
  # recovered, the live process untouched; a legacy-directory marker is processed.
  vr9_sub "(g)"; new_case; ctl_labels
  WORK="$C/work"; mkdir -p "$WORK"
  mk_reg "$C/reg" local-0123456789ab
  ctr="$(D create "${LB[@]}" --network none "$CI_BUSYBOX_IMAGE" true)" || ctr=""
  printf 'standalone\t%s\tvr9-g\n' "$ctr" >>"$C/reg/owned.tsv"
  p="$(bash -c 'echo $$')"
  launch "$C/hg.out" sleep 120; g="$LPID"; sleep 0.3; gs="$(proc_start "$g")" || gs=""
  VR_BOOT=bad vr_marker "$C/reg" "$p" 1 "$(vr8_rows "$g" "$gs" "$g" "$gs")"
  vr9_recover "$C/ev2"
  vr_blocked "the marker of run local-0123456789ab has a missing or malformed boot_id; operator resolution required (marker kept)"
  VR_REPO=/elsewhere/repo vr_marker "$C/reg" "$p" 1 "$(vr8_rows "$g" "$gs" "$g" "$gs")"
  vr9_recover "$C/ev3"
  vr_blocked "run local-0123456789ab belongs to checkout /elsewhere/repo; recover it (or wait for it) from that checkout (marker kept)"
  want_kept container "$ctr"
  VR_BOOT="$nb" vr_marker "$C/reg" "$p" 1 "$(vr8_rows "$g" "$gs" "$g" "$gs")"
  vr9_recover "$C/ev4"
  expect_rc 16; expect_phase recovery 0; expect_runs 0
  if grep -qF "validate: recovery: run local-0123456789ab was started in an earlier boot ($nb); its processes ended with that boot, so the process checks are skipped" "$C/out"; then :; else
    prob "the job log does not report the earlier boot of the planted marker"
  fi
  want_gone container "$ctr"
  if [[ -e "$VR_MF" ]]; then prob "the other-boot marker remains after the recovery"; fi
  if proc_is "$g" "$gs"; then :; else prob "the live process named as the other boot's harness group was signalled"; fi
  mk_reg "$C/reg2" local-0123456789ab
  ctr2="$(D create "${LB[@]}" --network none "$CI_BUSYBOX_IMAGE" true)" || ctr2=""
  printf 'standalone\t%s\tvr9-g-legacy\n' "$ctr2" >>"$C/reg2/owned.tsv"
  VR_MDIR=legacy vr_marker "$C/reg2" "$p" 1
  vr9_recover "$C/ev5"
  expect_rc 16; expect_phase recovery 0; expect_runs 0
  if grep -q "^validate: recovery: run local-0123456789ab did not finalize; delayed recovery verified " "$C/out"; then :; else prob "the legacy marker was not processed"; fi
  want_gone container "$ctr2"
  if [[ -e "$VR_MF" ]]; then prob "the legacy marker remains after the recovery"; fi
  if stop_tree "" "$g:$gs"; then :; else prob "the live decoy survived KILL: $STOP_LEFT"; fi
  vr9_tag
  detail "(a) the batch runner and its exec'd process outlived the KILL of validate.sh and its harness group; the next launch verified the dead group and watchdog, recovered the runner (g064.batch query empty after 6 s), recorded the delayed recovery and removed the marker; (b) a self-registered validate-controls.sh holding a D container, KILLed with its group, recovered the same way; (c) tampered token: 16, no harness, runner kept, this control's g001.run count unchanged; (d) rewritten boot_id: process checks skipped, recovered; (e) registry deleted and (f) marker mode 000: 16, no harness, runner untouched; (g) planted markers: malformed boot_id and another checkout 16 (kept), another boot with a live harness group recovered without signalling it, a legacy-directory marker processed (same boot only; no reboot claimed)"
}

# --- group F: profile selection (infra/ci/select-profile.sh on fixture repositories; no Docker) -------------------
# The fixture base holds the selector and two tracked files; each case is the base plus one commit. Unless a
# control says otherwise, the case is a detached merge request against the base and the selector runs as the
# gate job with the reviewed map of this checkout.
F_BASE="$ROOT/fbase" F_BASE_SHA="" F_AMAP="$ROOT/famap.tsv"
f_commit_in() { # DIR MESSAGE: one commit of the staged tree (empty allowed), fixed identity and dates
  GIT_AUTHOR_DATE=2026-01-03T00:00:00Z GIT_COMMITTER_DATE=2026-01-03T00:00:00Z \
    git -C "$1" -c user.name=ci-controls -c user.email=controls@example.invalid commit -q --allow-empty -m "$2"
}
mk_fbase() {
  mkdir -p "$F_BASE/infra/ci" "$F_BASE/web/app" "$F_BASE/management/v0"
  # G051: the selector sources infra/ci/lib.sh (suite lists and order) from its own directory.
  cp -- "$REPO_R/infra/ci/select-profile.sh" "$REPO_R/infra/ci/lib.sh" "$F_BASE/infra/ci/"
  printf 'export default function Page() { return null; }\n' >"$F_BASE/web/app/page.tsx"
  printf 'progress\n' >"$F_BASE/management/v0/PROGRESS.md"
  printf '.env\n.env.local\n' >"$F_BASE/.gitignore"
  # Owner correction: seed the exact delivery log so its rename control exercises an actual base-file rename.
  printf 'tracking\n' >"$F_BASE/management/v0/SESSION-HANDOFFS.jsonl"
  # G049: seeded agent paths, so deletion and rename controls have a base (the handoff files, a root agent skill and
  # settings, the Vestive skill and its relative Claude discovery link).
  mkdir -p "$F_BASE/.claude" "$F_BASE/.agents/skills/example" "$F_BASE/.agents/skills/claude-session-handoff/agents" \
    "$F_BASE/.agents/skills/claude-session-handoff/references" "$F_BASE/.agents/skills/vestive-brand/kit/src/styles" "$F_BASE/web/.agents/skills"
  printf 'settings\n' >"$F_BASE/.claude/settings.json"
  printf 'example skill\n' >"$F_BASE/.agents/skills/example/SKILL.md"
  printf 'handoff skill\n' >"$F_BASE/.agents/skills/claude-session-handoff/SKILL.md"
  printf 'handoff metadata\n' >"$F_BASE/.agents/skills/claude-session-handoff/agents/openai.yaml"
  printf 'handoff runbook\n' >"$F_BASE/.agents/skills/claude-session-handoff/references/local-session-handoff.md"
  printf 'vestive styles\n' >"$F_BASE/.agents/skills/vestive-brand/kit/src/styles/theme.css"
  mkdir -p "$F_BASE/.claude/skills"
  ln -s ../../.agents/skills/vestive-brand "$F_BASE/.claude/skills/vestive-brand"
  git -C "$F_BASE" init -q -b main
  git -C "$F_BASE" add -A
  f_commit_in "$F_BASE" "fixture base"
  F_BASE_SHA="$(git -C "$F_BASE" rev-parse HEAD)"
  # G051: the fixture map of the area controls, outside every case repository: the reviewed map plus area rows.
  cp -- "$REPO_R/infra/ci/profile-paths.tsv" "$F_AMAP"
  printf '%s\n' $'area-web\tdir\tweb/app/(features)/a/' $'area-web\tdir\tweb/tests/browser/features/a/' \
    $'area-web\tdir\tweb/tests/fixtures/features/a/' $'area-web\tdir\tweb/app/(features)/b/' \
    $'area-web\tdir\tweb/tests/browser/features/b/' $'area-web\tdir\tweb/tests/browser/features/w/' \
    $'area-web\tdir\tweb/app/(features)/c/' $'area-web\tdir\tweb/app/(features)/z/' \
    $'area-core\tdir\tcore/internal/a/' $'area-core\tdir\tcore/internal/b/' $'area-core\tdir\tcore/internal/c/' \
    $'area-intel\tdir\tintelligence/src/a/' $'area-intel\tdir\tintelligence/tests/a/' >>"$F_AMAP"
}
f_apply() { # DIR OPS...: add:PATH (append a line; a new or a modified file), put:PATH=TEXT (exactly TEXT, printf %b),
  # del:PATH, mv:OLD:NEW, ln:PATH:TARGET
  local d="$1" op p
  shift
  for op in "$@"; do
    case "$op" in
      add:*) p="${op#add:}"; mkdir -p "$(dirname -- "$d/$p")"; printf 'change %s\n' "$CTL_ID" >>"$d/$p"; git -C "$d" add -- "$p" ;;
      put:*) p="${op#put:}"; mkdir -p "$(dirname -- "$d/${p%%=*}")"; printf '%b\n' "${p#*=}" >"$d/${p%%=*}"; git -C "$d" add -- "${p%%=*}" ;;
      del:*) git -C "$d" rm -q -r -- "${op#del:}" ;;
      mv:*) p="${op#mv:}"; mkdir -p "$(dirname -- "$d/${p#*:}")"; git -C "$d" mv -- "${p%%:*}" "${p#*:}" ;;
      ln:*) p="${op#ln:}"; mkdir -p "$(dirname -- "$d/${p%%:*}")"; ln -sfn -- "${p#*:}" "$d/${p%%:*}"; git -C "$d" add -- "${p%%:*}" ;;
    esac
  done
}
f_repo() { # OPS...: the case repository, the fixture base plus one commit of OPS (none: an empty commit)
  C="$ROOT/cases/$CTL_ID"
  rm -rf -- "$C"
  mkdir -p "$C"
  cp -a -- "$F_BASE" "$C/repo"
  f_apply "$C/repo" "$@"
  f_commit_in "$C/repo" "case $CTL_ID"
}
FRC=0 FS_P="" FS_E="" FS_R="" FS_B="" FS_A="" FS_PS="" FS_ES="" FS_C="" FS_D=""
f_select() { # JOB TESTED_CONTENT [--requested P] [--base R] [--compare R] [--map FILE] [NAME=VALUE]...: the selector in
  # the case repository (its own TMPDIR under the case)
  local job="$1" tc="$2" map="$REPO_R/infra/ci/profile-paths.tsv" extra=() envs=()
  shift 2
  while (( $# > 0 )); do
    case "$1" in
      --requested|--base|--compare) extra+=("$1" "$2"); shift 2 ;;
      --map) map="$2"; shift 2 ;;
      *) envs+=("$1"); shift ;;
    esac
  done
  rm -f -- "$C/sel.env" "$C/sel.tsv"
  mkdir -p "$C/tmp"
  FRC=0
  env -i PATH="$PATH" HOME="${HOME:-/nonexistent}" LC_ALL=C GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 TMPDIR="$C/tmp" CI_CONTROLS_FORCE=1 "${envs[@]}" \
    bash "$C/repo/infra/ci/select-profile.sh" --job "$job" "${extra[@]}" --tested-content "$tc" --map "$map" \
    --out "$C/sel.env" --tsv "$C/sel.tsv" >"$C/sel.log" 2>&1 || FRC=$?
  FS_P="$(sed -n 's/^PIPELINE_PROFILE=//p' "$C/sel.env" 2>/dev/null)" || FS_P=""
  FS_E="$(sed -n 's/^EXECUTED_PROFILE=//p' "$C/sel.env" 2>/dev/null)" || FS_E=""
  FS_R="$(sed -n 's/^SELECTION_REASON=//p' "$C/sel.env" 2>/dev/null)" || FS_R=""
  FS_B="$(sed -n 's/^COMPARISON_BASE=//p' "$C/sel.env" 2>/dev/null)" || FS_B=""
  # G051: the areas, suites, coverage and digest lines.
  FS_A="$(sed -n 's/^AREAS=//p' "$C/sel.env" 2>/dev/null)" || FS_A=""
  FS_PS="$(sed -n 's/^PIPELINE_SUITES=//p' "$C/sel.env" 2>/dev/null)" || FS_PS=""
  FS_ES="$(sed -n 's/^EXECUTED_SUITES=//p' "$C/sel.env" 2>/dev/null)" || FS_ES=""
  FS_C="$(sed -n 's/^COVERAGE=//p' "$C/sel.env" 2>/dev/null)" || FS_C=""
  FS_D="$(sed -n 's/^SELECTION_DIGEST=//p' "$C/sel.env" 2>/dev/null)" || FS_D=""
}
f_want() { # PIPELINE EXECUTED REASON BASE: the exact selection
  if [[ "$FRC" != 0 ]]; then prob "selector exit $FRC: $(tail -1 "$C/sel.log" 2>/dev/null)"; return 0; fi
  if [[ "$FS_P/$FS_E/$FS_R" != "$1/$2/$3" ]]; then prob "selection $FS_P/$FS_E/$FS_R, want $1/$2/$3"; fi
  if [[ "$FS_B" != "$4" ]]; then prob "comparison base $FS_B, want $4"; fi
}
f_refused() { if [[ "$FRC" != 16 ]]; then prob "selector exit $FRC, want 16"; fi; if [[ -e "$C/sel.env" ]]; then prob "a selection was written"; fi; }
f_no_table() { if [[ -s "$C/sel.tsv" ]]; then prob "a diff was consulted ($(grep -c . "$C/sel.tsv") rows)"; fi; }
f_text() { local s="$*"; printf '%s' "${s//$'\t'/\\t}"; }
# Table-driven controls: "PIPELINE EXECUTED REASON|OP|OP..." for a detached MR against the base, gate job.
declare -A F_ROWS=(
  [F-gitignore]="fast none docs-only|add:.gitignore"
  [F-gitignore-runtime]="integration integration runtime|add:.gitignore|add:web/app/page.tsx"
  [F-gitignore-foundation]="full full foundation|add:.gitignore|add:infra/compose/local.yml"
  [F-gitignore-delete]="fast none docs-only|del:.gitignore"
  [F-web]="integration integration runtime|add:web/app/page.tsx"
  [F-core]="integration integration runtime|add:core/internal/health/checks.go"
  [F-intelligence]="integration integration runtime|add:intelligence/src/main.py"
  [F-contracts]="full full foundation|add:contracts/intelligence/execution-envelope.schema.json"
  [F-compose]="full full foundation|add:infra/compose/local.yml"
  [F-lock-gomod]="full full foundation|add:core/go.mod"
  [F-lock-gosum]="full full foundation|add:core/go.sum"
  [F-lock-pkg]="full full foundation|add:web/package.json"
  [F-lock-pkglock]="full full foundation|add:web/package-lock.json"
  [F-lock-pyproject]="full full foundation|add:intelligence/pyproject.toml"
  [F-lock-uvlock]="full full foundation|add:intelligence/uv.lock"
  [F-cfg-next]="full full foundation|add:web/next.config.ts"
  [F-cfg-dockerfile]="full full foundation|add:infra/docker/web.Dockerfile"
  [F-ci-yml]="full full foundation|add:.gitlab-ci.yml"
  [F-ci-map]="full full foundation|add:infra/ci/profile-paths.tsv"
  [F-ci-controls]="full full foundation|add:infra/ci/validate-controls.sh"
  [F-oracle-browser]="full full foundation|add:web/tests/browser/shell.spec.ts"
  [F-oracle-fixtures]="full full foundation|add:web/tests/fixtures/error-probe/route.ts"
  [F-oracle-conformance]="full full foundation|add:contracts/conformance/intelligence/expectations.json"
  [F-doc-readme]="fast none docs-only|add:README.md"
  [F-doc-foundation]="fast none docs-only|add:docs/foundation/OPERATIONS.md"
  [F-unknown]="full full unknown-path|add:tools/new.txt"
  [F-delete]="full full deletion-or-rename|del:web/app/page.tsx"
  [F-rename]="full full deletion-or-rename|mv:web/app/page.tsx:web/app/home.tsx"
  [F-mixed]="integration integration runtime|add:management/v0/PROGRESS.md|add:web/app/page.tsx"
  [F-unsafe]="full full unsafe-path|add:web/app/a"$'\t'"b.tsx"
  [F-nochange]="fast none no-change|"
  [F-session-tracking]="fast none docs-only|add:management/v0/SESSION-HANDOFFS.jsonl"
  [F-session-tracking-mixed]="integration integration runtime|add:management/v0/SESSION-HANDOFFS.jsonl|add:web/app/page.tsx"
  [F-session-tracking-near]="full full foundation|add:management/v0/SESSION-HANDOFFS-extra.jsonl"
  [F-session-tracking-moved]="full full deletion-or-rename|mv:management/v0/SESSION-HANDOFFS.jsonl:management/v0/control.jsonl"
  [F-status-progress]="fast none docs-only|add:management/v0/PROGRESS.md"
  [F-status-backlog]="fast none docs-only|add:management/v0/BACKLOG.md"
  [F-status-history]="fast none docs-only|add:docs/history/initial-product-architecture-discussion.md"
  [F-mix-spec]="fast none docs-only|add:management/v0/PROGRESS.md|add:management/v0/specs/G003-ci-runner-feedback-revision-spec.md"
  [F-mix-backlog-web]="integration integration runtime|add:management/v0/BACKLOG.md|add:web/app/page.tsx"
  [F-gov-architecture]="fast none docs-only|add:docs/ARCHITECTURE.md"
  [F-gov-security]="fast none docs-only|add:docs/SECURITY.md"
  [F-gov-decisions]="fast none docs-only|add:docs/DECISIONS.md"
  [F-gov-adr]="fast none docs-only|add:docs/adr/ADR-002-gitlab-hosted-ci-runner.md"
  [F-gov-handoff]="fast none docs-only|add:docs/agents/CLAUDE-SESSION-HANDOFF.md"
  [F-gov-agents]="fast none docs-only|add:AGENTS.md"
  [F-gov-claude]="fast none docs-only|add:CLAUDE.md"
  [F-gov-goal]="fast none docs-only|add:management/v0/goals/G003-foundation-observability-ci.md"
  [F-gov-spec]="fast none docs-only|add:management/v0/specs/G003-ci-runner-feedback-revision-spec.md"
  [F-gov-plan]="fast none docs-only|add:management/v0/plans/G003-ci-runner-feedback-revision-plan.md"
  [F-gov-allowlist]="fast none docs-only|add:management/v0/results/G003-scope-allowlist.txt"
  [F-gov-roadmap]="fast none docs-only|add:management/v0/ROADMAP.md"
  [F-gov-budget]="fast none docs-only|add:management/BUDGET_POLICY.md"
  [F-gov-newdoc]="fast none docs-only|add:docs/x.md"
  [F-gov-newmgmt]="fast none docs-only|add:management/x.md"
  [F-review-doc]="fast none docs-only|add:management/v0/results/G048-implementation-codex-review.md"
  # G051 (Plan §3.2, R3): web/components/ui/ is widened to full.
  [F-review-web]="full full foundation|add:management/v0/results/G048-implementation-codex-review.md|add:web/components/ui/button.tsx"
  [F-review-lock]="full full foundation|add:management/v0/results/G048-implementation-codex-review.md|add:web/package-lock.json"
  [F-management-script]="full full foundation|add:management/helpers/check.sh"
  [F-management-unknown]="full full foundation|add:management/new-control.xyz"
  [F-management-delete]="fast none docs-only|del:management/v0/PROGRESS.md"
  [F-management-rename]="fast none docs-only|mv:management/v0/PROGRESS.md:management/v0/Status.md"
  [F-management-outside]="full full deletion-or-rename|mv:management/v0/PROGRESS.md:infra/progress.md"
  [F-doc-script]="full full foundation|add:docs/agents/helper.sh"
  [F-doc-data]="full full foundation|add:docs/foundation/check.json"
  [F-doc-root-control]="full full unknown-path|add:new-control.txt"
  [F-doc-skill]="full full foundation|add:distributions/skills/example/SKILL.md"
  [F-doc-mixed-runtime]="integration integration runtime|add:AGENTS.md|add:docs/ARCHITECTURE.md|add:web/app/page.tsx"
  [F-doc-mixed-lock]="full full foundation|add:CLAUDE.md|add:docs/agents/GITLAB-BRANCH-WORKFLOW.md|add:web/package-lock.json"
  [F-doc-rename]="fast none docs-only|add:docs/Existing.md|mv:management/v0/PROGRESS.md:docs/Status with spaces.md"
  [F-doc-outside]="full full deletion-or-rename|mv:management/v0/PROGRESS.md:infra/Status.md"
  [F-doc-approval]="fast none docs-only|add:AGENTS.md|add:CLAUDE.md|add:docs/agents/GITLAB-BRANCH-WORKFLOW.md|add:management/v0/plans/G004-global-identity-sessions-plan.md"
  [F-migrate-dep]="full full foundation|add:core/internal/observability/logging.go"
  [F-config-dep]="full full foundation|add:core/internal/config/config.go"
  [F-tenancy]="full full foundation|add:core/internal/tenancy/service.go"
  # G049 (Plan r2 §3.2), detached MR source. The selector result is what an invocation would select; whether an
  # automatic pipeline exists at all is the workflow's admission (the three exact handoff files alone are omitted).
  [F-hand-skill]="fast none docs-only|add:.agents/skills/claude-session-handoff/SKILL.md"
  [F-hand-meta]="fast none docs-only|add:.agents/skills/claude-session-handoff/agents/openai.yaml"
  [F-hand-runbook]="fast none docs-only|add:.agents/skills/claude-session-handoff/references/local-session-handoff.md"
  [F-hand-docs]="fast none docs-only|add:.agents/skills/claude-session-handoff/SKILL.md|add:.agents/skills/claude-session-handoff/agents/openai.yaml|add:.agents/skills/claude-session-handoff/references/local-session-handoff.md|add:README.md|add:docs/foundation/OPERATIONS.md|add:management/v0/PROGRESS.md"
  [F-hand-del-skill]="fast none docs-only|del:.agents/skills/claude-session-handoff/SKILL.md"
  [F-hand-del-meta]="fast none docs-only|del:.agents/skills/claude-session-handoff/agents/openai.yaml"
  [F-hand-del-runbook]="fast none docs-only|del:.agents/skills/claude-session-handoff/references/local-session-handoff.md"
  [F-hand-swap]="fast none docs-only|del:.agents/skills/claude-session-handoff/references/local-session-handoff.md|mv:.agents/skills/claude-session-handoff/SKILL.md:.agents/skills/claude-session-handoff/references/local-session-handoff.md"
  [F-hand-to-doc]="fast none docs-only|mv:.agents/skills/claude-session-handoff/SKILL.md:docs/handoff.md"
  [F-hand-to-agent]="fast none docs-only|mv:.agents/skills/claude-session-handoff/SKILL.md:.agents/skills/claude-session-handoff/references/notes.md"
  [F-hand-to-skills]="full full deletion-or-rename|mv:.agents/skills/claude-session-handoff/SKILL.md:distributions/skills/example/SKILL.md"
  [F-hand-helper]="fast none agent-only|add:.agents/skills/claude-session-handoff/scripts/helper.sh"
  [F-agent-settings]="fast none agent-only|add:.claude/settings.json"
  [F-agent-claude-skill]="fast none agent-only|add:.claude/skills/example/SKILL.md"
  [F-agent-agents-skill]="fast none agent-only|add:.agents/skills/example/SKILL.md"
  [F-agent-unknown]="fast none agent-only|add:.agents/config.toml"
  [F-agent-delete]="fast none agent-only|del:.agents/skills/example/SKILL.md"
  [F-agent-rename]="fast none agent-only|mv:.claude/settings.json:.agents/settings.json"
  [F-agent-to-skills]="full full deletion-or-rename|mv:.agents/skills/example/SKILL.md:distributions/skills/example/SKILL.md"
  [F-nested-web]="full full foundation|add:web/.claude/skills/example/SKILL.md"
  [F-nested-core]="full full foundation|add:core/.claude/skills/example/SKILL.md"
  [F-nested-docs]="full full foundation|add:docs/.claude/skills/example/SKILL.md"
  [F-vestive-css]="full full foundation|add:.agents/skills/vestive-brand/kit/src/styles/theme.css"
  [F-vestive-kit]="full full foundation|add:.agents/skills/vestive-brand/kit/src/components/ui/button.tsx"
  [F-vestive-md]="full full foundation|add:.agents/skills/vestive-brand/SKILL.md"
  [F-vestive-link]="full full foundation|ln:.claude/skills/vestive-brand:../../.agents/skills/other"
  [F-web-agent]="full full foundation|add:web/.claude/settings.json"
  [F-dist-restruct]="full full foundation|add:distributions/skills/restruct-sdlc/claude-session-handoff/scripts/handoff.py"
  [F-agent-mixed-runtime]="integration integration runtime|add:.claude/settings.json|add:web/app/page.tsx"
  [F-agent-mixed-lock]="full full foundation|add:.claude/settings.json|add:web/package-lock.json"
  [F-hand-mixed-runtime]="integration integration runtime|add:.agents/skills/claude-session-handoff/SKILL.md|add:web/app/page.tsx"
  [F-hand-mixed-nested]="full full foundation|add:.agents/skills/claude-session-handoff/SKILL.md|add:docs/.claude/skills/example/SKILL.md"
  # G051 (Plan §3.2, R3): identity, first paint, UI primitives, shared helpers and contracts widened to full; the
  # remaining runtime prefixes unchanged (integration).
  [F-widen-identity]="full full foundation|add:core/internal/identity/x.go"
  [F-widen-server]="full full foundation|add:core/cmd/server/main.go"
  [F-widen-auth]="full full foundation|add:web/app/auth/login/route.ts"
  [F-widen-sign-in]="full full foundation|add:web/app/sign-in/page.tsx"
  [F-widen-account]="full full foundation|add:web/app/account/page.tsx"
  [F-widen-account-api]="full full foundation|add:web/app/api/account/route.ts"
  [F-widen-account-ui]="full full foundation|add:web/components/account/x.tsx"
  [F-widen-security]="full full foundation|add:intelligence/src/security/auth.py"
  [F-widen-globals]="full full foundation|add:web/app/globals.css"
  [F-widen-layout]="full full foundation|add:web/app/layout.tsx"
  [F-widen-public]="full full foundation|add:web/public/x.svg"
  [F-widen-layout-ui]="full full foundation|add:web/components/layout/x.tsx"
  [F-widen-ui]="full full foundation|add:web/components/ui/button.tsx"
  [F-widen-lib]="full full foundation|add:web/lib/x.ts"
  [F-widen-unit-tests]="full full foundation|add:web/tests/unit/x.test.mjs"
  [F-widen-mcp]="full full foundation|add:core/internal/mcp/x.go"
  [F-widen-streaming]="full full foundation|add:intelligence/src/streaming/x.py"
  [F-keep-health]="integration integration runtime|add:core/internal/health/x.go"
  [F-keep-api-health]="integration integration runtime|add:web/app/api/health/route.ts"
  [F-keep-api-ready]="integration integration runtime|add:web/app/api/ready/route.ts"
  [F-keep-models]="integration integration runtime|add:intelligence/src/models/x.py"
  [F-keep-providers]="integration integration runtime|add:intelligence/src/providers/x.py"
)
f_row() { # ID: a table-driven selector control
  local row="${F_ROWS[$1]:-}" p e r ops=()
  if [[ -z "$row" ]]; then prob "control $1 is not defined"; return 0; fi
  read -r p e r <<<"${row%%|*}"
  if [[ -n "${row#*|}" ]]; then IFS='|' read -r -a ops <<<"${row#*|}"; fi
  f_repo "${ops[@]}"
  f_select gate mr-source CI_MERGE_REQUEST_DIFF_BASE_SHA="$F_BASE_SHA"
  f_want "$p" "$e" "$r" "$F_BASE_SHA"
  if [[ "$p" != fast || "$r" != no-change ]] && [[ ! -s "$C/sel.tsv" ]]; then prob "the selection table is empty"; fi
  detail "MR [$(f_text "${ops[*]:-empty commit}")] -> pipeline $FS_P, gate $FS_E ($FS_R)"
}
ctl_F_gov_allowlist() {
  f_row F-gov-allowlist
  # Extend the existing control: exact metadata only; mixed and near-match controls remain covered.
  f_repo add:management/v0/results/G014-scope-allowlist.txt add:web/app/page.tsx
  f_select gate mr-source CI_MERGE_REQUEST_DIFF_BASE_SHA="$F_BASE_SHA"
  f_want integration integration runtime "$F_BASE_SHA"
  f_repo add:management/v0/results/G014-scope-allowlist.txt add:infra/validation/run.sh
  f_select gate mr-source CI_MERGE_REQUEST_DIFF_BASE_SHA="$F_BASE_SHA"
  f_want full full foundation "$F_BASE_SHA"
  local p
  for p in management/v0/results/G014-scope-allowlist-extra.txt management/v0/results/G014-scope-allowlist.txt.sh \
    management/v0/results/nested/G014-scope-allowlist.txt management/v0/results/G14-scope-allowlist.txt \
    management/v1/results/G014-scope-allowlist.txt management/v0/results/.claude/skills/check/SKILL.md; do
    f_repo "add:$p"
    f_select gate mr-source CI_MERGE_REQUEST_DIFF_BASE_SHA="$F_BASE_SHA"
    f_want full full foundation "$F_BASE_SHA"
  done
  f_repo add:management/v0/results/G014-scope-allowlist.txt
  f_select gate branch-push CI_COMMIT_BRANCH=staging CI_DEFAULT_BRANCH=main CI_COMMIT_BEFORE_SHA="$F_BASE_SHA"
  f_want fast none docs-only "$F_BASE_SHA"
  detail "exact Goal scope metadata omitted; mixed runtime/control, malformed/nested/other-release paths and nested Claude skills remain covered; staging scope-only selects no gate"
}
ctl_F_gitignore_guards() {
  f_repo
  cp -- "$C/repo/.gitignore" "$C/base.ignore"
  sed -n '/^static_gitignore() {$/,/^}$/p' "$REPO_R/infra/validation/run.sh" >"$C/ignore-check.sh"
  local mode want rc
  for mode in valid env-negation local-negation broad-negation missing symlink; do
    rm -f -- "$C/repo/.gitignore"
    cp -- "$C/base.ignore" "$C/repo/.gitignore"
    want=1
    case "$mode" in
      valid) printf '\n/management/**/SESSION-HANDOFFS.jsonl\n' >>"$C/repo/.gitignore"; want=0 ;;
      env-negation) printf '\n!.env\n' >>"$C/repo/.gitignore" ;;
      local-negation) printf '\n!.env.local\n' >>"$C/repo/.gitignore" ;;
      broad-negation) printf '\n!*\n' >>"$C/repo/.gitignore" ;;
      missing) rm -- "$C/repo/.gitignore" ;;
      symlink) rm -- "$C/repo/.gitignore"; ln -s ../base.ignore "$C/repo/.gitignore" ;;
    esac
    rc=0
    REPO="$C/repo" bash -c 'log() { :; }; source "$1"; static_gitignore' ignore-check "$C/ignore-check.sh" || rc=$?
    [[ "$rc" == "$want" ]] || prob "$mode returned $rc, expected $want"
  done
  detail "effective Git ignore check: valid local-log addition passes; env/local/broad negations, missing file and symlink fail"
}
# G051 O3: run.sh's route inventory (web_routes_ok) on a scratch web/app tree: the exact non-feature inventory, and a
# feature route only with its exact registered web area root; everything else fails closed.
ctl_F_route_guards() {
  local w map mode want rc r
  new_case
  w="$C/rt/web" map="$C/rt/map.tsv"
  mkdir -p "$C/rt"
  sed -n '/^WEB_ROUTES=/,/^}$/p' "$REPO_R/infra/validation/run.sh" >"$C/route-check.sh"
  if ! grep -q '^web_routes_ok() {' "$C/route-check.sh"; then prob "no web_routes_ok in run.sh"; return 0; fi
  for mode in inventory registered unregistered tests-root-only bad-name bad-segment missing-route extra-route; do
    rm -rf -- "$w"; mkdir -p "$w/app"
    for r in $(bash -c 'source "$1"; printf "%s" "$WEB_ROUTES"' rc "$C/route-check.sh"); do mkdir -p "$w/app/$(dirname "$r")"; : >"$w/app/$r"; done
    printf '%s\n' $'docs\tdir\tdocs/' >"$map"
    want=1
    case "$mode" in
      inventory) want=0 ;;
      registered) mkdir -p "$w/app/(features)/demo/x"; : >"$w/app/(features)/demo/x/page.tsx"
        printf '%s\n' $'area-web\tdir\tweb/app/(features)/demo/' >>"$map"; want=0 ;;
      unregistered) mkdir -p "$w/app/(features)/demo"; : >"$w/app/(features)/demo/page.tsx" ;;
      tests-root-only) mkdir -p "$w/app/(features)/demo"; : >"$w/app/(features)/demo/page.tsx"
        printf '%s\n' $'area-web\tdir\tweb/tests/browser/features/demo/' >>"$map" ;;
      bad-name) mkdir -p "$w/app/(features)/Demo"; : >"$w/app/(features)/Demo/page.tsx"
        printf '%s\n' $'area-web\tdir\tweb/app/(features)/Demo/' >>"$map" ;;
      bad-segment) mkdir -p "$w/app/(features)/demo/a b"; : >"$w/app/(features)/demo/a b/page.tsx"
        printf '%s\n' $'area-web\tdir\tweb/app/(features)/demo/' >>"$map" ;;
      missing-route) rm -- "$w/app/api/health/route.ts" ;;
      extra-route) mkdir -p "$w/app/admin"; : >"$w/app/admin/page.tsx" ;;
    esac
    rc=0
    bash -c 'log() { :; }; source "$1"; web_routes_ok "$2" "$3"' route-check "$C/route-check.sh" "$w" "$map" || rc=$?
    [[ "$rc" == "$want" ]] || prob "$mode returned $rc, expected $want"
  done
  detail "route inventory: the exact non-feature routes pass, a feature route with its exact registered web/app/(features)/<n>/ root passes; an unregistered or tests-root-only area, a malformed name or path segment, and a missing or extra other route fail"
}
ZERO40="$(printf '0%.0s' {1..40})"
ctl_F_base_missing() { f_repo add:web/app/page.tsx; f_select gate mr-source; f_want full full missing-base none; f_no_table
  detail "MR without CI_MERGE_REQUEST_DIFF_BASE_SHA -> full (missing-base), no diff"; }
ctl_F_base_zero() { f_repo add:web/app/page.tsx; f_select gate branch-push CI_COMMIT_BRANCH=main CI_DEFAULT_BRANCH=main CI_COMMIT_BEFORE_SHA="$ZERO40"
  f_want full full missing-base none; f_no_table; detail "protected push with the zero before-SHA -> full (missing-base)"; }
ctl_F_base_absent() { f_repo add:web/app/page.tsx; f_select gate mr-source CI_MERGE_REQUEST_DIFF_BASE_SHA="$(printf 'e%.0s' {1..40})"
  f_want full full missing-base none; f_no_table; detail "a base absent from the clone -> full (missing-base)"; }
ctl_F_base_nonancestor() {
  local x
  f_repo add:web/app/page.tsx
  git -C "$C/repo" checkout -q -b side "$F_BASE_SHA"
  f_apply "$C/repo" add:side.txt; f_commit_in "$C/repo" side
  x="$(git -C "$C/repo" rev-parse HEAD)"
  git -C "$C/repo" checkout -q main
  f_select gate mr-source CI_MERGE_REQUEST_DIFF_BASE_SHA="$x"; f_want full full degenerate-base none; f_no_table
  detail "a base that is not an ancestor of HEAD -> full (degenerate-base)"
}
ctl_F_base_degenerate() { f_repo add:web/app/page.tsx; f_select gate mr-source CI_MERGE_REQUEST_DIFF_BASE_SHA="$(git -C "$C/repo" rev-parse HEAD)"
  f_want full full degenerate-base none; f_no_table; detail "base = HEAD -> full (degenerate-base)"; }
ctl_F_feature_new() {
  f_repo add:web/app/page.tsx
  f_select fast branch-push CI_COMMIT_BRANCH=feature CI_DEFAULT_BRANCH=main CI_COMMIT_BEFORE_SHA="$ZERO40"; f_want fast fast feature-push none; f_no_table
  f_select gate branch-push CI_COMMIT_BRANCH=feature CI_DEFAULT_BRANCH=main CI_COMMIT_BEFORE_SHA="$ZERO40"; f_refused
  detail "a new feature branch (zero before-SHA): the fast job runs fast (feature-push); a gate job is refused (16)"
}
# Owner 2026-10-19 (no automatic post-merge cloud checks): the workflow starts no pipeline for a push to staging,
# the default branch or a feature branch, nor for any tag. The selector keeps its branch-push semantics (an explicit
# local or diagnostic attribution); these controls also assert the workflow suppression in .gitlab-ci.yml.
f_push_never() { # BRANCH-CONDITION: the workflow has exactly "if: $CI_PIPELINE_SOURCE == "push" && CONDITION / when: never"
  local cond="\$CI_PIPELINE_SOURCE == \"push\" && $1" w
  w="$(sed -n '/^workflow:$/,/^stages:/p' "$REPO_R/.gitlab-ci.yml")"
  if grep -qxF -A1 "    - if: $cond" <<<"$w" && [[ "$(grep -xF -A1 "    - if: $cond" <<<"$w" | tail -1)" == "      when: never" ]]; then :
  else prob "the workflow does not suppress [$cond] (when: never)"; fi
  if grep -q '^    - if: \$CI_PIPELINE_SOURCE == "push" && \$CI_COMMIT_BRANCH$' <<<"$w"; then prob "the workflow still admits branch pushes"; fi
  if grep -qxF -A1 '    - if: $CI_COMMIT_TAG' <<<"$w" && [[ "$(grep -xF -A1 '    - if: $CI_COMMIT_TAG' <<<"$w" | tail -1)" == "      when: never" ]]; then :
  else prob "the workflow does not suppress every tag"; fi
}
ctl_F_push_main() { f_repo add:management/v0/PROGRESS.md; f_select gate branch-push CI_COMMIT_BRANCH=main CI_DEFAULT_BRANCH=main CI_COMMIT_BEFORE_SHA="$F_BASE_SHA"
  f_want fast none docs-only "$F_BASE_SHA"; f_push_never '$CI_COMMIT_BRANCH == $CI_DEFAULT_BRANCH'
  detail "a main push starts no pipeline (workflow never; tags never); the selector's branch-push semantics remain (docs-only -> fast, gate none)"; }
ctl_F_push_staging() { f_repo add:web/app/page.tsx; f_select gate branch-push CI_COMMIT_BRANCH=staging CI_DEFAULT_BRANCH=main CI_COMMIT_BEFORE_SHA="$F_BASE_SHA"
  f_want integration integration runtime "$F_BASE_SHA"; f_push_never '$CI_COMMIT_BRANCH == "staging"'
  detail "a staging push starts no pipeline, not even validate:checks (workflow never; staging is qualified locally on the merged SHA); the selector's branch-push semantics remain (runtime -> integration)"; }
# ONEOFF RC qualification: the verified promotion context is full without a diff, the CI meta-tests included even with
# CI_CONTROLS_FORCE=0 and documentation-only candidate content.
ctl_F_rc_promotion() { f_repo add:management/v0/PROGRESS.md; f_select gate mr-rc-promotion CI_CONTROLS_FORCE=0
  f_want full full rc-promotion none; f_no_table
  if [[ " $FS_PS " != *" ci-controls "* ]]; then prob "the RC promotion suites lack ci-controls [$FS_PS]"; fi
  detail "RC promotion, docs-only candidate, CI_CONTROLS_FORCE=0 -> full (rc-promotion) with ci-controls, no diff"; }
ctl_F_fast_job() { f_repo add:web/app/page.tsx; f_select fast mr-source CI_MERGE_REQUEST_DIFF_BASE_SHA="$F_BASE_SHA"
  f_want integration fast runtime "$F_BASE_SHA"; detail "MR runtime change, fast job: pipeline integration, the fast job runs fast"; }
ctl_F_schedule() { f_repo add:management/v0/PROGRESS.md; f_select gate branch-schedule CI_COMMIT_BRANCH=main CI_DEFAULT_BRANCH=main CI_COMMIT_BEFORE_SHA="$F_BASE_SHA"
  f_want full full schedule none; f_no_table; detail "schedule (docs-only change) -> full (schedule), no diff"; }
ctl_F_web_intent() { f_repo add:management/v0/PROGRESS.md; f_select gate branch-web CI_COMMIT_BRANCH=main CI_DEFAULT_BRANCH=main CI_COMMIT_BEFORE_SHA="$F_BASE_SHA"
  f_want full full web-explicit none; f_no_table; detail "web (release intent, docs-only change) -> full (web-explicit), no diff"; }
# G049: push contexts for root-agent, handoff, nested-skill and mixed changes; explicit contexts with handoff-only.
ctl_F_agent_push_staging() { f_repo add:.claude/settings.json; f_select gate branch-push CI_COMMIT_BRANCH=staging CI_DEFAULT_BRANCH=main CI_COMMIT_BEFORE_SHA="$F_BASE_SHA"
  f_want fast none agent-only "$F_BASE_SHA"; f_push_never '$CI_COMMIT_BRANCH == "staging"'
  detail "a staging push starts no pipeline (workflow never); the selector's branch-push semantics remain (root agent change -> fast, gate none)"; }
ctl_F_agent_push_main() { f_repo add:.claude/settings.json; f_select gate branch-push CI_COMMIT_BRANCH=main CI_DEFAULT_BRANCH=main CI_COMMIT_BEFORE_SHA="$F_BASE_SHA"
  f_want fast none agent-only "$F_BASE_SHA"; f_push_never '$CI_COMMIT_BRANCH == $CI_DEFAULT_BRANCH'
  detail "a main push starts no pipeline (workflow never); the selector's branch-push semantics remain (root agent change -> fast, gate none)"; }
ctl_F_hand_push_staging() { f_repo add:.agents/skills/claude-session-handoff/SKILL.md; f_select gate branch-push CI_COMMIT_BRANCH=staging CI_DEFAULT_BRANCH=main CI_COMMIT_BEFORE_SHA="$F_BASE_SHA"
  f_want fast none docs-only "$F_BASE_SHA"; detail "push to staging, handoff file only (if invoked) -> fast, gate none"; }
ctl_F_nested_push_main() { f_repo add:docs/.claude/skills/example/SKILL.md; f_select gate branch-push CI_COMMIT_BRANCH=main CI_DEFAULT_BRANCH=main CI_COMMIT_BEFORE_SHA="$F_BASE_SHA"
  f_want full full foundation "$F_BASE_SHA"; detail "push to main, docs-subtree Claude skill -> full"; }
ctl_F_agent_mixed_push_staging() { f_repo add:.claude/settings.json add:web/app/page.tsx; f_select gate branch-push CI_COMMIT_BRANCH=staging CI_DEFAULT_BRANCH=main CI_COMMIT_BEFORE_SHA="$F_BASE_SHA"
  f_want integration integration runtime "$F_BASE_SHA"; detail "push to staging, root agent plus runtime -> integration"; }
ctl_F_hand_web() { f_repo add:.agents/skills/claude-session-handoff/SKILL.md; f_select gate branch-web CI_COMMIT_BRANCH=main CI_DEFAULT_BRANCH=main CI_COMMIT_BEFORE_SHA="$F_BASE_SHA"
  f_want full full web-explicit none; f_no_table; detail "web, handoff file only -> full (web-explicit), no diff"; }
ctl_F_hand_api() { f_repo add:.agents/skills/claude-session-handoff/SKILL.md; f_select gate branch-api CI_COMMIT_BRANCH=feature CI_DEFAULT_BRANCH=main CI_COMMIT_BEFORE_SHA="$F_BASE_SHA"
  f_want full full api-explicit none; f_no_table; detail "api, handoff file only -> full (api-explicit), no diff"; }
ctl_F_hand_schedule() { f_repo add:.agents/skills/claude-session-handoff/SKILL.md; f_select gate branch-schedule CI_COMMIT_BRANCH=main CI_DEFAULT_BRANCH=main CI_COMMIT_BEFORE_SHA="$F_BASE_SHA"
  f_want full full schedule none; f_no_table; detail "schedule, handoff file only -> full (schedule), no diff"; }
ctl_F_api_intent() { f_repo add:management/v0/PROGRESS.md; f_select gate branch-api CI_COMMIT_BRANCH=feature CI_DEFAULT_BRANCH=main CI_COMMIT_BEFORE_SHA="$F_BASE_SHA"
  f_want full full api-explicit none; f_no_table; detail "api (branch pipeline, docs-only change) -> full (api-explicit), no diff"; }
# G004 E-1 (Plan r5 §8.5): the protected identity-qualification schedule is attributed branch-schedule and selected
# full; a schedule of any other non-default branch is refused at attribution, protected or not.
ctl_F_schedule_qualification() {
  local r1
  f_repo add:management/v0/PROGRESS.md
  f_select gate branch-schedule CI_COMMIT_BRANCH=identity-qualification CI_DEFAULT_BRANCH=main CI_COMMIT_REF_PROTECTED=true CI_COMMIT_BEFORE_SHA="$F_BASE_SHA"
  f_want full full schedule none; f_no_table
  [[ -n "$S_SHA" ]] || mk_origin
  b_checkout refs/heads/feature; b_env
  run_v "${GL[@]}" CI_PIPELINE_SOURCE=schedule CI_COMMIT_SHA="$S_SHA" CI_COMMIT_BRANCH=identity-qualification CI_COMMIT_REF_NAME=identity-qualification CI_COMMIT_REF_PROTECTED=true -- --profile gate --publish-dir ci-evidence
  r1="$RC"; expect_rc 0; b_positive branch-schedule; sum_has executed_profile '"full"'
  detail "selection full (schedule), no diff; validate.sh exit $r1: tested_content branch-schedule, executed full"
}
ctl_F_schedule_other() {
  [[ -n "$S_SHA" ]] || mk_origin
  b_checkout refs/heads/feature; b_env
  run_v "${GL[@]}" CI_PIPELINE_SOURCE=schedule CI_COMMIT_SHA="$S_SHA" CI_COMMIT_BRANCH=staging CI_COMMIT_REF_NAME=staging CI_COMMIT_REF_PROTECTED=true -- --profile gate --publish-dir ci-evidence
  b_negative "schedule outside the default branch"
  detail "exit $RC: a protected schedule of another non-default branch is refused (schedule outside the default branch)"
}
ctl_F_local() {
  f_repo add:management/v0/PROGRESS.md
  f_select local local --requested integration; f_want integration integration local-explicit none; f_no_table
  f_select gate local; f_refused
  detail "local explicit integration -> integration (local-explicit), no diff; a gate job in a local run is refused (16)"
}
f_merged() { # SOURCE_OPS TARGET_OPS ("|"-joined, may be empty): HEAD = the merge of a source and a target commit on the base
  local s tg so=() to=()
  C="$ROOT/cases/$CTL_ID"; rm -rf -- "$C"; mkdir -p "$C"; cp -a -- "$F_BASE" "$C/repo"
  if [[ -n "$1" ]]; then IFS='|' read -r -a so <<<"$1"; fi
  if [[ -n "$2" ]]; then IFS='|' read -r -a to <<<"$2"; fi
  git -C "$C/repo" checkout -q -b source "$F_BASE_SHA"; f_apply "$C/repo" "${so[@]}"; f_commit_in "$C/repo" source
  s="$(git -C "$C/repo" rev-parse HEAD)"
  git -C "$C/repo" checkout -q -b target "$F_BASE_SHA"; f_apply "$C/repo" "${to[@]}"; f_commit_in "$C/repo" target
  tg="$(git -C "$C/repo" rev-parse HEAD)"
  git -C "$C/repo" checkout -q --detach "$tg"
  GIT_AUTHOR_DATE=2026-01-03T00:00:00Z GIT_COMMITTER_DATE=2026-01-03T00:00:00Z \
    git -C "$C/repo" -c user.name=ci-controls -c user.email=controls@example.invalid merge -q --no-ff -m "merged result" "$s"
  f_select gate mr-merged-result CI_MERGE_REQUEST_DIFF_BASE_SHA="$F_BASE_SHA" CI_MERGE_REQUEST_SOURCE_BRANCH_SHA="$s" CI_MERGE_REQUEST_TARGET_BRANCH_SHA="$tg"
  f_want full full merged-result none; f_no_table
}
ctl_F_merged_docs() { f_merged "add:management/v0/PROGRESS.md" ""; detail "merged result, docs-only source change -> full (merged-result), no diff"; }
ctl_F_merged_target() { f_merged "" "add:web/app/page.tsx"; detail "merged result, source unchanged, target-side runtime change -> full (merged-result), no diff"; }
ctl_F_merged_both_docs() { f_merged "add:management/v0/PROGRESS.md" "add:management/v0/BACKLOG.md"; detail "merged result, source and target docs-only -> full (merged-result), no diff"; }
f_map_bad() { # LINE: the reviewed map plus LINE -> refused (16)
  f_repo add:web/app/page.tsx
  cp -- "$REPO_R/infra/ci/profile-paths.tsv" "$C/map.tsv"
  printf '%s\n' "$1" >>"$C/map.tsv"
  f_select gate mr-source --map "$C/map.tsv" CI_MERGE_REQUEST_DIFF_BASE_SHA="$F_BASE_SHA"; f_refused
}
ctl_F_map_dup() { f_map_bad "$(awk -F'\t' '$1 == "runtime" {print; exit}' "$REPO_R/infra/ci/profile-paths.tsv")"; detail "a duplicated map row -> 16"; }
ctl_F_map_class() { f_map_bad $'urgent\tdir\ttools/'; detail "a map row with an unknown class -> 16"; }
ctl_F_map_abs() { f_map_bad $'foundation\tdir\t/etc/'; detail "a map row with an absolute path -> 16"; }
ctl_F_diff_fail() {
  local tree obj
  f_repo add:web/app/page.tsx
  tree="$(git -C "$C/repo" rev-parse HEAD:web/app)"
  obj="$C/repo/.git/objects/${tree:0:2}/${tree:2}"
  if [[ -f "$obj" ]]; then rm -f -- "$obj"; else prob "the fixture tree object is not a loose object"; return 0; fi
  f_select gate mr-source CI_MERGE_REQUEST_DIFF_BASE_SHA="$F_BASE_SHA"; f_want full full diff-failed none; f_no_table
  detail "a diff that cannot be computed (a missing tree object behind a valid base) -> full (diff-failed), no table"
}

# --- group F: areas and the dependency closure (G051 Plan §3.2-§3.3, §6.1) -------------------------------------------
# A case is the fixture base plus the common area content FA_AB (and the case's own base additions) committed as the
# comparison base, then the case's change as HEAD; the selector runs with the fixture map $F_AMAP (outside the case
# repository). Registered: web a (app, browser and fixtures roots), b (app, browser), w (browser), c (app), z (app; never
# present); core a (with an integration test), b, c (no files at the base); intelligence a (src, tests). Each control
# asserts all nine selection lines (the executed partition derived from the job, the pipeline profile and suites), the
# five-column table and the exact closure consumers ("path<via>" of the "+" rows).
FA_AB="put:web/app/(features)/a/page.tsx=export default function A() { return null; }|put:web/app/(features)/a/w.ts=export const w = 1;"
FA_AB+="|put:web/tests/browser/features/a/a.spec.ts=test('a', () => {});|put:web/tests/fixtures/features/a/data.json={}"
FA_AB+="|put:web/app/(features)/b/page.tsx=export default function B() { return null; }|put:web/tests/browser/features/b/b.spec.ts=test('b', () => {});"
FA_AB+="|put:web/app/(features)/c/page.tsx=export default function C() { return null; }"
FA_AB+="|put:core/internal/a/a.go=package a|put:core/internal/a/sub/s.go=package sub|put:core/internal/a/a_it_test.go=//go:build integration\n\npackage a"
FA_AB+="|put:core/internal/b/b.go=package b|put:intelligence/src/a/__init__.py=|put:intelligence/src/a/sub.py=X = 1"
FA_AB+="|put:intelligence/src/models/m.py=from . import (\n    resolver,\n)\nfrom .settings import Settings"
FA_ALL="core:a,core:b,intel:a,web:a,web:b,web:c"
FA_FULL="lint lint-base lint-web lint-core controls ci-controls provenance contracts root migrations ops rename off feature-web feature-core"
FA_INT="lint lint-base lint-web lint-core provenance contracts unit root ops rename off feature-web feature-core"
FA_WEB="lint-base lint-web provenance unit-web feature-web"
FA_FASTWEB="lint lint-base lint-web lint-core provenance contracts unit unit-web feature-web"
FA_CORE="lint-base lint-core provenance unit-core"
FA_CORE_IT="lint-base lint-core provenance unit-core feature-core"
FA_INTEL="lint-base provenance contracts unit-intel"
B_SHA="" FX_E="" FX_ES="" FX_C=""
a_repo() { # BASEOPS CHANGEOPS ("|"-joined, may be empty): FA_AB plus BASEOPS committed as B_SHA, then CHANGEOPS as HEAD
  local bo=() co=()
  C="$ROOT/cases/$CTL_ID"; rm -rf -- "$C"; mkdir -p "$C"; cp -a -- "$F_BASE" "$C/repo"
  IFS='|' read -r -a bo <<<"$FA_AB${1:+|$1}"
  if [[ -n "$2" ]]; then IFS='|' read -r -a co <<<"$2"; fi
  f_apply "$C/repo" "${bo[@]}"; f_commit_in "$C/repo" "area base"
  B_SHA="$(git -C "$C/repo" rev-parse HEAD)"
  f_apply "$C/repo" "${co[@]}"; f_commit_in "$C/repo" "case $CTL_ID"
}
fa_minus() { # LIST MINUS: LIST without MINUS's suites, order kept
  local s out=""
  for s in $1; do if [[ " $2 " != *" $s "* ]]; then out+=" $s"; fi; done
  printf '%s' "${out# }"
}
fa_exec() { # JOB PIPELINE PIPELINE_SUITES: FX_E FX_ES FX_C, the job's partition (Plan §3.3 rule 8, §3.7)
  case "$1:$2" in
    local:*) FX_E="$2" FX_ES="$3" FX_C=complete ;;
    fast:affected) FX_E=affected FX_ES="$3" FX_C=complete ;;
    fast:fast) FX_E=fast FX_ES="$FAST_SUITES" FX_C=complete ;;
    fast:*) FX_E=fast FX_ES="$FAST_SUITES" FX_C=partial ;;
    gate:fast|gate:affected) FX_E=none FX_ES=none FX_C=none ;;
    *) FX_E="$2" FX_ES="$(fa_minus "$3" "$FAST_SUITES")" FX_C=partial ;;
  esac
}
f_cons() { awk -F'\t' '$1 == "+" {print $3 "<" $5 ">"}' "$C/sel.tsv" 2>/dev/null | LC_ALL=C sort | paste -sd, -; }
fa_want() { # JOB PIPELINE/EXECUTED/REASON BASE AREAS PIPELINE_SUITES CONSUMERS ("-" none, "*" not checked)
  local p e r want_c
  IFS=/ read -r p e r <<<"$2"
  if [[ "$FRC" != 0 ]]; then prob "selector exit $FRC: $(tail -1 "$C/sel.log" 2>/dev/null)"; return 0; fi
  fa_exec "$1" "$p" "$5"
  if [[ "$e" != "$FX_E" ]]; then prob "control table: $1 job of a $p pipeline executes $FX_E, not $e"; fi
  if [[ "$FS_P/$FS_E/$FS_R" != "$p/$FX_E/$r" ]]; then prob "selection $FS_P/$FS_E/$FS_R, want $p/$FX_E/$r"; fi
  if [[ "$FS_B" != "$3" ]]; then prob "comparison base $FS_B, want $3"; fi
  if [[ "$FS_A" != "$4" ]]; then prob "AREAS [$FS_A], want [$4]"; fi
  if [[ "$FS_PS" != "$5" ]]; then prob "PIPELINE_SUITES [$FS_PS], want [$5]"; fi
  if [[ "$FS_ES/$FS_C" != "$FX_ES/$FX_C" ]]; then prob "EXECUTED_SUITES/COVERAGE [$FS_ES/$FS_C], want [$FX_ES/$FX_C]"; fi
  if [[ ! "$FS_D" =~ ^[0-9a-f]{64}$ ]]; then prob "SELECTION_DIGEST [$FS_D] is not a SHA-256"; fi
  if [[ "$(grep -c . "$C/sel.env")" != 9 ]]; then prob "the selection has $(grep -c . "$C/sel.env") lines, want 9"; fi
  if awk -F'\t' 'NF != 5 {f = 1} END {exit !f}' "$C/sel.tsv" 2>/dev/null; then prob "a selection table row is not five columns"; fi
  if [[ "$6" != '*' ]]; then
    want_c="$(if [[ "$6" != - ]]; then tr ',' '\n' <<<"$6" | LC_ALL=C sort | paste -sd, -; fi)"
    if [[ "$(f_cons)" != "$want_c" ]]; then prob "consumers [$(f_cons)], want [$want_c]"; fi
  fi
}
fa_ps() { # PIPELINE [SUITES]: the expected pipeline suites (full and integration with every present area's feature suite)
  case "$1" in full) printf '%s' "$FA_FULL" ;; integration) printf '%s' "$FA_INT" ;; fast) printf '%s' "$FAST_SUITES" ;; *) printf '%s' "${2:-}" ;; esac
}
fa_t() { # JOB BASEOPS CHANGEOPS P/E/R|16 [AREAS] [SUITES] [CONSUMERS]: a detached merge request of the case, selected by JOB
  local job="$1" want="$4" p
  a_repo "$2" "$3"
  f_select "$job" mr-source --map "$F_AMAP" CI_MERGE_REQUEST_DIFF_BASE_SHA="$B_SHA"
  if [[ "$want" == 16 ]]; then f_refused; detail "MR [$(f_text "$3")] -> refused (16)"; return 0; fi
  p="${want%%/*}"
  fa_want "$job" "$want" "$B_SHA" "${5:-$FA_ALL}" "$(fa_ps "$p" "${6:-}")" "${7:-*}"
  detail "MR [$(f_text "$3")]${2:+ over [$(f_text "$2")]}, $job job -> $FS_P/$FS_E/$FS_R areas $FS_A suites [$FS_ES]${FS_C:+ ($FS_C)}$( [[ -n "$(f_cons)" ]] && printf '; consumers %s' "$(f_cons)")"
}
# The r1 area set, mixed changes and unregistered paths.
ctl_F_area_page_spec() { fa_t fast "" "add:web/app/(features)/a/page.tsx|add:web/tests/browser/features/a/a.spec.ts" affected/affected/area web:a "$FA_WEB" -; }
ctl_F_area_page_spec_gate() { fa_t gate "" "add:web/app/(features)/a/page.tsx|add:web/tests/browser/features/a/a.spec.ts" affected/none/area web:a "$FA_WEB" -; }
ctl_F_area_test_only() { fa_t fast "" "put:core/internal/b/b_test.go=package b" affected/affected/area core:b "$FA_CORE" -; }
ctl_F_area_fixture_only() { fa_t fast "" "add:web/tests/fixtures/features/a/data.json" affected/affected/area web:a "$FA_WEB" -; }
ctl_F_area_core_it() { fa_t fast "" "add:core/internal/a/a.go" affected/affected/area core:a "$FA_CORE_IT" -; }
ctl_F_area_core_unit() { fa_t fast "" "add:core/internal/b/b.go" affected/affected/area core:b "$FA_CORE" -; }
ctl_F_area_intel() { fa_t fast "" "add:intelligence/src/a/sub.py" affected/affected/area intel:a "$FA_INTEL" -; }
ctl_F_area_cross_stack() { fa_t fast "put:web/tests/browser/features/w/x.stack.spec.ts=stack" "add:core/internal/a/a.go" affected/affected/area core:a,web:w \
  "lint-base lint-web lint-core provenance unit-web unit-core feature-web feature-core" "web/tests/browser/features/w/x.stack.spec.ts<core:a>"; }
ctl_F_area_two_web() { fa_t fast "" "add:web/app/(features)/a/page.tsx|add:web/app/(features)/b/page.tsx" affected/affected/area web:a,web:b "$FA_WEB" -; }
ctl_F_mix_area_docs() { fa_t fast "" "add:web/app/(features)/a/page.tsx|add:management/v0/PROGRESS.md" affected/affected/area web:a "$FA_FASTWEB" -; }
ctl_F_mix_area_agent() { fa_t fast "" "add:web/app/(features)/a/page.tsx|add:.claude/settings.json" affected/affected/area web:a "$FA_FASTWEB" -; }
ctl_F_mix_area_runtime() { fa_t gate "" "add:web/app/(features)/a/page.tsx|add:web/app/page.tsx" integration/integration/runtime; }
ctl_F_mix_area_lock() { fa_t gate "" "add:web/app/(features)/a/page.tsx|add:web/package-lock.json" full/full/foundation; }
ctl_F_unreg_core_pkg() { fa_t gate "" "add:core/internal/newpkg/x.go" integration/integration/runtime; }
ctl_F_unreg_feature_test() { fa_t gate "" "add:web/tests/browser/features/zz/x.spec.ts" full/full/foundation; }
# Map refusals (the fixture map plus one row): bad roots, bad names, file kind, overlap with a reviewed row, a duplicate.
fa_map_bad() { # LINE...: each refused (16)
  local l n=0
  for l in "$@"; do
    a_repo "" "add:web/app/(features)/a/page.tsx"
    cp -- "$F_AMAP" "$C/map.tsv"; printf '%s\n' "$l" >>"$C/map.tsv"
    f_select gate mr-source --map "$C/map.tsv" CI_MERGE_REQUEST_DIFF_BASE_SHA="$B_SHA"; f_refused
    n=$((n + 1))
  done
  detail "$n malformed area rows, each refused (16)"
}
ctl_F_map_area_root() { fa_map_bad $'area-web\tdir\tweb/features/x/' $'area-web\tdir\tweb/lib/x/' $'area-core\tdir\tcore/cmd/x/'; }
ctl_F_map_area_name() { fa_map_bad $'area-core\tdir\tcore/internal/Bad/' $'area-core\tdir\tcore/internal/a-b/' $'area-intel\tdir\tintelligence/src/a/b/'; }
ctl_F_map_area_overlap() { fa_map_bad $'area-core\tdir\tcore/internal/identity/' $'area-core\tdir\tcore/internal/tenancy/' $'foundation\tfile\tweb/app/(features)/a/x.ts'; }
ctl_F_map_area_kind() { fa_map_bad $'area-web\tfile\tweb/app/(features)/q/page.tsx'; }
ctl_F_map_area_dup() { fa_map_bad $'area-web\tdir\tweb/app/(features)/a/'; }
# The committed map names every area root: it is the registration, never a consumer (validate.sh's case trees rely on it).
ctl_F_map_in_repo() {
  local bo=()
  C="$ROOT/cases/$CTL_ID"; rm -rf -- "$C"; mkdir -p "$C"; cp -a -- "$F_BASE" "$C/repo"
  IFS='|' read -r -a bo <<<"$FA_AB"; f_apply "$C/repo" "${bo[@]}"
  cp -- "$F_AMAP" "$C/repo/infra/ci/profile-paths.tsv"; git -C "$C/repo" add infra/ci/profile-paths.tsv
  f_commit_in "$C/repo" "base with the map"; B_SHA="$(git -C "$C/repo" rev-parse HEAD)"
  f_apply "$C/repo" "add:core/internal/a/a.go"; f_commit_in "$C/repo" "case"
  f_select fast mr-source --map "$C/repo/infra/ci/profile-paths.tsv" CI_MERGE_REQUEST_DIFF_BASE_SHA="$B_SHA"
  fa_want fast affected/affected/area "$B_SHA" core:a "$FA_CORE_IT" -
  detail "the committed map with area rows is no consumer: core area a -> affected, core:a, no consumer"
}
# Contexts: a push to staging selects by paths; web, api, schedule and merged results never narrow; a feature push is fast.
ctl_F_area_push_staging() {
  local r1
  a_repo "" "add:web/app/(features)/a/page.tsx"
  f_select gate branch-push --map "$F_AMAP" CI_COMMIT_BRANCH=staging CI_DEFAULT_BRANCH=main CI_COMMIT_BEFORE_SHA="$B_SHA"
  fa_want gate affected/none/area "$B_SHA" web:a "$FA_WEB" -; r1="$FS_E"
  f_select fast branch-push --map "$F_AMAP" CI_COMMIT_BRANCH=staging CI_DEFAULT_BRANCH=main CI_COMMIT_BEFORE_SHA="$B_SHA"
  fa_want fast affected/affected/area "$B_SHA" web:a "$FA_WEB" -
  detail "push to staging, web area a: pipeline affected; gate $r1, fast job [$FS_ES]"
}
fa_context() { # TESTED_CONTENT REASON [fast]: a never-narrowing context over an area change, no diff
  a_repo "" "add:web/app/(features)/a/page.tsx"
  f_select "${3:-gate}" "$1" --map "$F_AMAP" CI_COMMIT_BRANCH=main CI_DEFAULT_BRANCH=main CI_COMMIT_BEFORE_SHA="$B_SHA"
  fa_want "${3:-gate}" "full/$([[ "${3:-gate}" == fast ]] && echo fast || echo full)/$2" none "$FA_ALL" "$FA_FULL" -; f_no_table
  detail "$1 with an area change -> full ($2), every present area ($FS_A), the ${3:-gate} job [$FS_ES], no diff"
}
ctl_F_area_web() {
  # The legacy context assertion retains its complete forced profile. Then verify
  # automatic CI-control inclusion independently with force explicitly disabled.
  fa_context branch-web web-explicit
  f_select gate branch-web --map "$F_AMAP" CI_CONTROLS_FORCE=0
  fa_want gate full/full/web-explicit none "$FA_ALL" "${FA_FULL/ ci-controls/}" -
  f_apply "$C/repo" "put:infra/ci/control-fixture.txt=meta-test dependency"
  f_commit_in "$C/repo" "CI dependency"
  f_select gate branch-web --map "$F_AMAP" CI_CONTROLS_FORCE=0
  fa_want gate full/full/web-explicit none "$FA_ALL" "$FA_FULL" -
  f_apply "$C/repo" "put:unmapped-control-fixture.txt=unknown dependency"
  f_commit_in "$C/repo" "unknown dependency"
  f_select gate branch-web --map "$F_AMAP" CI_CONTROLS_FORCE=0
  fa_want gate full/full/web-explicit none "$FA_ALL" "$FA_FULL" -
  detail "forced full retained; automatic selection omits CI meta-tests for an application-only delta, includes CI dependencies and unknown paths"
}
ctl_F_area_api() { fa_context branch-api api-explicit; }
# ONEOFF RC qualification: the area map never narrows an RC promotion; a docs-only candidate with present registered
# areas selects every present area and the CI meta-tests with CI_CONTROLS_FORCE=0, gate and fast job.
ctl_F_area_rc() {
  local r1
  a_repo "" "add:management/v0/PROGRESS.md"
  f_select gate mr-rc-promotion --map "$F_AMAP" CI_CONTROLS_FORCE=0
  fa_want gate full/full/rc-promotion none "$FA_ALL" "$FA_FULL" -; f_no_table; r1="$FS_A"
  f_select fast mr-rc-promotion --map "$F_AMAP" CI_CONTROLS_FORCE=0
  fa_want fast full/fast/rc-promotion none "$FA_ALL" "$FA_FULL" -; f_no_table
  detail "RC promotion, docs-only candidate, CI_CONTROLS_FORCE=0 -> full (rc-promotion), every present area ($r1), ci-controls included; gate remainder and fast job, no diff"
}
ctl_F_area_schedule() { fa_context branch-schedule schedule fast; }
ctl_F_area_merged() {
  local s tg
  a_repo "" ""
  git -C "$C/repo" checkout -q -b source "$B_SHA"; f_apply "$C/repo" "add:web/app/(features)/a/w.ts"; f_commit_in "$C/repo" source
  s="$(git -C "$C/repo" rev-parse HEAD)"
  git -C "$C/repo" checkout -q -b target "$B_SHA"; f_commit_in "$C/repo" target; tg="$(git -C "$C/repo" rev-parse HEAD)"
  git -C "$C/repo" checkout -q --detach "$tg"
  GIT_AUTHOR_DATE=2026-01-03T00:00:00Z GIT_COMMITTER_DATE=2026-01-03T00:00:00Z \
    git -C "$C/repo" -c user.name=ci-controls -c user.email=controls@example.invalid merge -q --no-ff -m "merged result" "$s"
  f_select gate mr-merged-result --map "$F_AMAP" CI_MERGE_REQUEST_DIFF_BASE_SHA="$B_SHA" CI_MERGE_REQUEST_SOURCE_BRANCH_SHA="$s" CI_MERGE_REQUEST_TARGET_BRANCH_SHA="$tg"
  fa_want gate full/full/merged-result none "$FA_ALL" "$FA_FULL" -; f_no_table
  detail "merged result with an area-only source change -> full (merged-result), every present area, no diff"
}
ctl_F_area_feature_push() {
  a_repo "" "add:web/app/(features)/a/page.tsx"
  f_select fast branch-push --map "$F_AMAP" CI_COMMIT_BRANCH=feature CI_DEFAULT_BRANCH=main CI_COMMIT_BEFORE_SHA="$B_SHA"
  fa_want fast fast/fast/feature-push none none "$FAST_SUITES" -; f_no_table
  detail "a feature-branch push with an area change -> fast (feature-push), no areas, no diff"
}
# Comparison base and local selection (Plan §3.3 rule 9).
ctl_F_area_base_missing() { a_repo "" "add:web/app/(features)/a/page.tsx"; f_select gate mr-source --map "$F_AMAP"
  fa_want gate full/full/missing-base none "$FA_ALL" "$FA_FULL" -; f_no_table; detail "an area change without a comparison base -> full (missing-base)"; }
ctl_F_area_local_base() {
  local r1
  a_repo "" "add:web/app/(features)/a/page.tsx"
  f_select local local --requested affected --base "$B_SHA" --map "$F_AMAP"; fa_want local affected/affected/area "$B_SHA" web:a "$FA_WEB" -; r1="$FS_P"
  f_select local local --requested affected --base HEAD~1 --map "$F_AMAP"; fa_want local affected/affected/area "$B_SHA" web:a "$FA_WEB" -
  detail "local affected with --base (a SHA: $r1; HEAD~1: $FS_P), complete coverage of [$FS_ES]"
}
ctl_F_area_local_no_base() { a_repo "" "add:web/app/(features)/a/page.tsx"; f_select local local --requested affected --map "$F_AMAP"
  fa_want local full/full/missing-base none "$FA_ALL" "$FA_FULL" -; f_no_table
  detail "local affected with neither --base nor --compare -> full (missing-base), no diff"; }
ctl_F_area_local_base_head() { a_repo "" "add:web/app/(features)/a/page.tsx"; f_select local local --requested affected --base HEAD --map "$F_AMAP"
  fa_want local full/full/degenerate-base none "$FA_ALL" "$FA_FULL" -; f_no_table; detail "local affected with --base HEAD -> full (degenerate-base)"; }
ctl_F_area_local_compare() { a_repo "" "add:web/app/(features)/a/page.tsx"; f_select local local --requested affected --compare "$B_SHA" --map "$F_AMAP"
  fa_want local affected/affected/area "$B_SHA" web:a "$FA_WEB" -; detail "local --compare against the previous candidate -> affected, exactly [$FS_ES] to rerun"; }
ctl_F_area_local_compare_same() { a_repo "" "add:web/app/(features)/a/page.tsx"; f_select local local --requested affected --compare HEAD --map "$F_AMAP"
  fa_want local fast/fast/no-change "$(git -C "$C/repo" rev-parse HEAD)" none "$FAST_SUITES" -; detail "local --compare against an identical tree -> fast (no-change)"; }
ctl_F_area_local_compare_unknown() { a_repo "" "add:web/app/(features)/a/page.tsx"
  f_select local local --requested affected --compare "$(printf 'e%.0s' {1..40})" --map "$F_AMAP"
  fa_want local full/full/missing-base none "$FA_ALL" "$FA_FULL" -; f_no_table; detail "local --compare with an unknown revision -> full (missing-base)"; }
ctl_F_area_ci_compare() { a_repo "" "add:web/app/(features)/a/page.tsx"
  f_select gate mr-source --compare "$B_SHA" --map "$F_AMAP" CI_MERGE_REQUEST_DIFF_BASE_SHA="$B_SHA"; f_refused
  detail "--compare in a CI job -> refused (16)"; }
ctl_F_area_ci_base() { a_repo "" "add:web/app/(features)/a/page.tsx"
  f_select fast mr-source --base "$B_SHA" --map "$F_AMAP" CI_MERGE_REQUEST_DIFF_BASE_SHA="$B_SHA"; f_refused
  detail "--base in a CI job -> refused (16)"; }
ctl_F_area_local_refusals() {
  a_repo "" "add:web/app/(features)/a/page.tsx"
  f_select local local --requested full --base "$B_SHA" --map "$F_AMAP"; f_refused
  f_select local local --requested affected --base "$B_SHA" --compare "$B_SHA" --map "$F_AMAP"; f_refused
  f_select local local --requested affected --base --help --map "$F_AMAP"; f_refused
  detail "--base with a non-affected profile, --base with --compare and an option-like revision -> each refused (16)"
}
# The gate's remainder of full and integration pipelines (Plan §3.7), and the selection digest.
ctl_F_gate_full_remainder() { fa_t gate "" "add:infra/compose/local.yml" full/full/foundation; }
ctl_F_gate_integration_remainder() { fa_t gate "" "add:web/app/page.tsx" integration/integration/runtime; }
ctl_F_digest() {
  local d1 d2 d3
  a_repo "" "add:web/app/(features)/a/page.tsx"
  f_select fast mr-source --map "$F_AMAP" CI_MERGE_REQUEST_DIFF_BASE_SHA="$B_SHA"; d1="$FS_D"
  f_select fast mr-source --map "$F_AMAP" CI_MERGE_REQUEST_DIFF_BASE_SHA="$B_SHA"; d2="$FS_D"
  f_select gate mr-source --map "$F_AMAP" CI_MERGE_REQUEST_DIFF_BASE_SHA="$B_SHA"; d3="$FS_D"
  if [[ ! "$d1" =~ ^[0-9a-f]{64}$ || "$d1" != "$d2" || "$d1" != "$d3" ]]; then prob "digests [$d1] [$d2] [$d3], want one SHA-256 for the same selection in both jobs"; fi
  a_repo "" "add:web/app/(features)/b/page.tsx"
  f_select fast mr-source --map "$F_AMAP" CI_MERGE_REQUEST_DIFF_BASE_SHA="$B_SHA"
  if [[ "$FS_D" == "$d1" ]]; then prob "another area selection has the same digest"; fi
  detail "one selection digest for the same candidate and selection (twice, fast and gate job); another selection differs"
}
# C1 additions (Plan §6.1): supported forms, old and new edges, ambiguity, allow-list, absent and removed areas.
ctl_F_dep_go_subpkg() { fa_t gate "put:core/internal/tenancy/t.go=package tenancy\nimport _ \"core/internal/a/sub\"" "add:core/internal/a/sub/s.go" full/full/area-imported \
  "" "" "core/internal/tenancy/t.go<core:a>"; }
ctl_F_dep_go_subpkg_area() { fa_t fast "put:core/internal/b/b.go=package b\nimport \"core/internal/a/sub\"" "add:core/internal/a/sub/s.go" affected/affected/area core:a,core:b \
  "$FA_CORE_IT" "core/internal/b/b.go<core:a>"; }
ctl_F_dep_go_test_only() { fa_t gate "put:core/internal/tenancy/t_test.go=package tenancy\nimport \"core/internal/a\"" "add:core/internal/a/a.go" full/full/area-imported \
  "" "" "core/internal/tenancy/t_test.go<core:a>"; }
ctl_F_dep_go_testdata_path() { fa_t gate "put:core/internal/migrate/x.go=package migrate\nvar p = \"core/internal/a/testdata/x\"" "add:core/internal/a/a.go" full/full/area-imported \
  "" "" "core/internal/migrate/x.go<core:a>"; }
ctl_F_dep_web_alias() { fa_t gate "put:web/lib/x.ts=import w from \"@/app/(features)/a/x\";" "add:web/app/(features)/a/page.tsx" full/full/area-imported "" "" "web/lib/x.ts<web:a>"; }
ctl_F_dep_web_relative() { fa_t gate "put:web/app/page.tsx=import { w } from \"./(features)/a/w\";" "add:web/app/(features)/a/w.ts" integration/integration/area-imported \
  "" "" "web/app/page.tsx<web:a>"; }
ctl_F_dep_web_relative_area() { fa_t fast "put:web/app/(features)/b/page.tsx=import { w } from \"../a/w\";" "add:web/app/(features)/a/w.ts" affected/affected/area web:a,web:b \
  "$FA_WEB" "web/app/%28features%29/b/page.tsx<web:a>"; }
ctl_F_dep_web_css() { fa_t gate "put:web/app/styles/feature.css=.x { background: url(../(features)/a/x.svg); }" "add:web/app/(features)/a/page.tsx" \
  integration/integration/area-imported "" "" "web/app/styles/feature.css<web:a>"; }
ctl_F_dep_web_dynamic() { fa_t gate "put:web/app/dyn.ts=const m = await import(name);" "add:web/app/(features)/a/page.tsx" full/full/area-ambiguous "" "" "web/app/dyn.ts<ambiguous>"; }
ctl_F_dep_web_template() { fa_t gate "put:web/app/tpl.ts=const m = await import(\`./(features)/\${n}/w\`);" "add:web/app/(features)/a/page.tsx" full/full/area-ambiguous \
  "" "" "web/app/tpl.ts<ambiguous>"; }
ctl_F_dep_web_escape() { fa_t gate "put:web/app/esc.ts=import x from \"../../core/x\";" "add:web/app/(features)/a/page.tsx" full/full/area-ambiguous "" "" "web/app/esc.ts<ambiguous>"; }
ctl_F_dep_py_submodule() { fa_t gate "put:intelligence/src/main.py=from a.sub import x" "add:intelligence/src/a/sub.py" integration/integration/area-imported \
  "" "" "intelligence/src/main.py<intel:a>"; }
ctl_F_dep_py_relative() { fa_t gate "put:intelligence/src/models/r.py=from ..a import sub" "add:intelligence/src/a/sub.py" integration/integration/area-imported \
  "" "" "intelligence/src/models/r.py<intel:a>"; }
ctl_F_dep_py_relative_unresolved() { fa_t gate "put:intelligence/src/models/u.py=from ... import x" "add:intelligence/src/a/sub.py" full/full/area-ambiguous \
  "" "" "intelligence/src/models/u.py<ambiguous>"; }
ctl_F_dep_py_dynamic() { fa_t gate "put:intelligence/src/models/d.py=import importlib\nm = importlib.import_module(name)" "add:intelligence/src/a/sub.py" full/full/area-ambiguous \
  "" "" "intelligence/src/models/d.py<ambiguous>"; }
ctl_F_dep_py_dynamic_webarea() { fa_t gate "put:intelligence/src/models/d.py=m = __import__(name)" "add:web/app/(features)/a/page.tsx" full/full/area-ambiguous \
  "" "" "intelligence/src/models/d.py<ambiguous>"; }
ctl_F_dep_removed_edge() { fa_t fast "put:core/internal/b/b.go=package b\nimport \"core/internal/a/sub\"" "del:core/internal/a/sub/s.go" affected/affected/area core:a,core:b \
  "$FA_CORE_IT" "core/internal/b/b.go<core:a>"; }
ctl_F_dep_removed_foundation_consumer() { fa_t gate "put:core/internal/tenancy/t.go=package tenancy\nimport _ \"core/internal/a/sub\"" "del:core/internal/a/sub/s.go" \
  full/full/area-imported "" "" "core/internal/tenancy/t.go<core:a>"; }
ctl_F_dep_renamed_area() { fa_t fast "put:core/internal/b/b.go=package b\nimport \"core/internal/a\"" \
  "mv:core/internal/a/a.go:core/internal/c/a.go|mv:core/internal/a/a_it_test.go:core/internal/c/a_it_test.go|mv:core/internal/a/sub/s.go:core/internal/c/sub/s.go|put:core/internal/b/b.go=package b\nimport \"core/internal/c\"" \
  affected/affected/area-removed core:a,core:b,core:c "$FA_CORE_IT" "core/internal/b/b.go<core:a>,core/internal/b/b.go<core:c>"; }
ctl_F_dep_transitive() { fa_t fast "put:core/internal/b/b.go=package b\nimport \"core/internal/a\"|put:web/app/(features)/c/page.tsx=// uses core/internal/b" \
  "add:core/internal/a/a.go" affected/affected/area core:a,core:b,web:c "lint-base lint-web lint-core provenance unit-web unit-core feature-web feature-core" \
  "core/internal/b/b.go<core:a>,web/app/%28features%29/c/page.tsx<core:b>"; }
# G051 O3: a harness file naming a feature route is a genuine foundation consumer (the former probe route pin): full.
ctl_F_dep_harness_route() { fa_t gate "put:infra/validation/run.sh=routes=\"(features)/a/page.tsx account/page.tsx\"" \
  "add:web/app/(features)/a/page.tsx|add:web/tests/browser/features/a/a.spec.ts" full/full/area-imported "" "" "infra/validation/run.sh<web:a>"; }
ctl_F_dep_docs_consumer() { fa_t fast "put:docs/notes.md=see core/internal/a/x" "add:core/internal/a/a.go" affected/affected/area core:a "$FA_CORE_IT" "docs/notes.md<core:a>"; }
ctl_F_fixture_other_area() { fa_t fast "put:web/tests/browser/features/b/b.spec.ts=import d from '../../fixtures/features/a/data.json';" "add:web/tests/fixtures/features/a/data.json" \
  affected/affected/area web:a,web:b "$FA_WEB" "web/tests/browser/features/b/b.spec.ts<web:a>"; }
ctl_F_fixture_inherited() { fa_t gate "put:web/tests/browser/shell.spec.ts=import d from '../fixtures/features/a/data.json';" "add:web/tests/fixtures/features/a/data.json" \
  full/full/area-imported "" "" "web/tests/browser/shell.spec.ts<web:a>"; }
ctl_F_feature_test_only() { fa_t fast "" "add:web/tests/browser/features/a/a.spec.ts" affected/affected/area web:a "$FA_WEB" -; }
ctl_F_area_unclassified_script() { fa_t gate "" "add:core/internal/a/run.sh" full/full/area-unclassified "" "" -; }
ctl_F_area_unclassified_gomod() { fa_t gate "" "add:core/internal/a/go.mod" full/full/area-unclassified "" "" -; }
ctl_F_area_unclassified_config() { fa_t gate "" "add:web/app/(features)/a/next.config.ts" full/full/area-unclassified "" "" -; }
ctl_F_area_unclassified_pyproject() { fa_t gate "" "add:intelligence/src/a/pyproject.toml" full/full/area-unclassified "" "" -; }
ctl_F_area_symlink() { fa_t gate "" "ln:web/app/(features)/a/link.ts:page.tsx" full/full/area-unclassified "" "" -; }
ctl_F_area_route_handler() { fa_t gate "" "add:web/app/(features)/a/api/route.ts" integration/integration/area-route-handler "" "" -; }
ctl_F_closure_base_unreadable() {
  local bb
  a_repo "" "add:web/app/(features)/a/w.ts"
  bb="$(git -C "$C/repo" rev-parse "$B_SHA:web/app/(features)/a/w.ts")"
  if [[ -f "$C/repo/.git/objects/${bb:0:2}/${bb:2}" ]]; then rm -f -- "$C/repo/.git/objects/${bb:0:2}/${bb:2}"; else prob "the base blob is not a loose object"; return 0; fi
  f_select gate mr-source --map "$F_AMAP" CI_MERGE_REQUEST_DIFF_BASE_SHA="$B_SHA"
  fa_want gate full/full/closure-failed "$B_SHA" "$FA_ALL" "$FA_FULL" '*'
  detail "the area file's comparison-base blob unreadable (the tree diff still works) -> full (closure-failed)"
}
ctl_F_registered_absent() {
  local r1
  fa_t fast "" "add:web/app/(features)/a/page.tsx" affected/affected/area web:a "$FA_WEB" -
  if grep -q '^select-profile: INFO registered-absent ' "$C/sel.log"; then :; else prob "no registered-absent INFO for the never-present areas"; fi
  r1="$FS_A"
  fa_t gate "" "add:web/app/page.tsx" integration/integration/runtime
  detail "registered areas with no files at the base or HEAD (web z, web w, core c) add nothing: an area change -> $r1; a runtime change -> areas $FS_A, suites [$FS_PS]"
}
ctl_F_area_removed_clean() {
  local r1
  fa_t fast "" "del:web/app/(features)/c/page.tsx" affected/affected/area-removed web:c "lint-base lint-web provenance unit-web" -; r1="$FS_ES"
  fa_t fast "" "del:core/internal/b/b.go" affected/affected/area-removed core:b "$FA_CORE" -
  detail "every file of an area deleted, no consumer: web c -> [$r1]; core b -> [$FS_ES] (lint and unit of its kind, no feature suite)"
}
ctl_F_area_removed_consumer() { fa_t gate "put:web/lib/use-c.ts=import C from \"@/app/(features)/c/page\";" "del:web/app/(features)/c/page.tsx" full/full/area-removed \
  core:a,core:b,intel:a,web:a,web:b "" "web/lib/use-c.ts<web:c>"; }

# --- group G: the hosted adapter (infra/ci/hosted-context.sh) ------------------------------------------------------
# Each case runs the adapter as the job script does, in the pinned docker CLI image of the jobs (busybox ash): as
# this user without CAP_DAC_OVERRIDE/CAP_DAC_READ_SEARCH, without a network, with a tmpfs as the build directory's
# filesystem (the disk guard reads real df figures) and a docker shim first on PATH that records every call's
# arguments (never the environment) and plays the job daemon. The synthetic job credentials reach the container
# only through a mode-600 environment file, never an argument.
CLI_IMAGE="$(sed -n 's/^FROM \(docker:[^ ]*-cli@sha256:[0-9a-f]\{64\}\) AS dockercli$/\1/p' "$REPO_R/infra/docker/ci-context.Dockerfile" 2>/dev/null)" || CLI_IMAGE=""
G_TOKEN="$(ci_random_hex 16)" G_REGPW="$(ci_random_hex 16)" G_NS=/builds/grp G_TMPFS=4128m
# The reviewed environment allow-list of the context (hosted-context.sh run_context).
declare -A G_ALLOW=()
for v in GITLAB_CI CI_PIPELINE_SOURCE CI_COMMIT_SHA CI_COMMIT_BRANCH CI_COMMIT_REF_NAME CI_COMMIT_TAG CI_DEFAULT_BRANCH \
  CI_MERGE_REQUEST_IID CI_MERGE_REQUEST_EVENT_TYPE CI_MERGE_REQUEST_SOURCE_BRANCH_NAME CI_MERGE_REQUEST_TARGET_BRANCH_NAME \
  CI_MERGE_REQUEST_SOURCE_BRANCH_SHA CI_MERGE_REQUEST_TARGET_BRANCH_SHA CI_MERGE_REQUEST_DIFF_BASE_SHA CI_MERGE_REQUEST_APPROVED \
  CI_PIPELINE_ID CI_PIPELINE_IID CI_PIPELINE_URL CI_JOB_ID CI_JOB_URL GITLAB_USER_LOGIN CI_RUNNER_ID CI_RUNNER_DESCRIPTION \
  CI_RUNNER_TAGS CI_SERVER_VERSION CI_JOB_TIMEOUT CI_PROJECT_DIR CI_JOB_STARTED_AT CI_COMMIT_BEFORE_SHA RUNNER_SCRIPT_TIMEOUT \
  CI_COMMIT_REF_PROTECTED CI_PROJECT_ID; do
  G_ALLOW[$v]=1
done
g_context() { # BEHAVIOUR: what the context does once the adapter starts it (the shim execs this script)
  {
    printf '#!/bin/sh\n# group G context: %s\n' "$1"
    cat <<'GCTX'
NS=/builds/grp
wait_stop() { i=0; while [ ! -e /gs/stopped ] && [ "$i" -lt 600 ]; do sleep 0.2; i=$((i + 1)); done; exit 143; }
fill() { dd if=/dev/zero of="$NS/fill" bs=1048576 count=64 2>/dev/null; }
GCTX
    case "$1" in
      "exit "*) printf 'sleep 2\nexit %s\n' "${1#exit }" ;;
      healthy) printf 'sleep 4\nexit 0\n' ;;
      staging-inspect) printf 'f="$(ls /tmp/hosted-context-staging.* 2>/dev/null | head -1)"\nif [ -n "$f" ]; then stat -c %%a "$f" >/gs/stg-mode; cut -d= -f1 "$f" | sort | tr "\\n" " " >/gs/stg-keys; fi\nexit 0\n' ;;
      config-copy) printf 'cp "$CI_PROJECT_DIR/.git/config" /gs/config-after\nsleep 1\nexit 0\n' ;;
      fill) printf 'sleep 1\nfill\nwait_stop\n' ;;
      vanish) printf 'sleep 1\nmv "$NS" "$NS.gone"\nwait_stop\n' ;;
      race) printf 'sleep 1\nfill\ni=0\nwhile ! ls /tmp/hosted-context-guard.*/fail >/dev/null 2>&1 && [ "$i" -lt 300 ]; do sleep 0.2; i=$((i + 1)); done\nexit 0\n' ;;
      hang-sample) printf 'sleep 2\n: >/gs/exec-hang\nwait_stop\n' ;;
      empty-beat) printf 'i=0\nwhile [ "$i" -lt 60 ]; do\n  for f in /tmp/hosted-context-guard.*/beat; do : >"$f"; done\n  sleep 0.1; i=$((i + 1))\ndone\nexit 0\n' ;;
      daemon-drop) printf 'sleep 2\nprintf "Filesystem 1024-blocks Used Available Capacity Mounted on\\noverlay 104857600 89128960 15728640 85%%%% /\\n" >/gs/daemon-df\nsleep 3\nexit 0\n' ;;
      kill-guard|stop-guard)
        if [[ "$1" == kill-guard ]]; then printf 'SIG=KILL\n'; else printf 'SIG=STOP\n'; fi
        cat <<'GKILL'
# The guard: a child of the adapter (PID 1) with the adapter's own command line that is not an ancestor of this
# context; it is killed outright (an unexpected death) or stopped (alive but no longer sampling).
sleep 2
anc=" $$ " p=$$
while [ "$p" -gt 1 ]; do p="$(awk '{print $4}' "/proc/$p/stat")"; anc="$anc$p "; done
want="$(tr '\0' ' ' </proc/1/cmdline)"
for d in /proc/[0-9]*; do
  q="${d#/proc/}"
  case "$anc" in *" $q "*) continue ;; esac
  [ "$(awk '{print $4}' "$d/stat" 2>/dev/null)" = 1 ] || continue
  [ "$(tr '\0' ' ' <"$d/cmdline" 2>/dev/null)" = "$want" ] || continue
  kill -"$SIG" "$q" 2>/dev/null && echo "$q" >>/gs/killed
done
wait_stop
GKILL
      ;;
    esac
  } >"$C/fx/context.sh"
}
g_config() { # KIND: the job's .git/config (@TOKEN@ is replaced by the job token inside the container)
  local u=$'[core]\n\trepositoryformatversion = 0\n[remote "origin"]\n\turl = https://gitlab-ci-token@gitlab.example.invalid/group/project.git\n\tfetch = +refs/heads/*:refs/remotes/origin/*\n'
  case "$1" in
    username) printf '%s' "$u" ;;
    password) printf '%s' "${u/gitlab-ci-token@/gitlab-ci-token:@TOKEN@@}" ;;
    include-pw) printf '%s[include]\n\tpath = inc.conf\n' "$u" ;;
    include-nested) printf '%s[include]\n\tpath = a.conf\n' "$u" ;;
    include-ok) printf '%s[include]\n\tpath = clean.conf\n[includeIf "gitdir:/nonexistent/"]\n\tpath = /nonexistent/outside.conf\n' "$u" ;;
    header-cred) printf '%s[url "https://u:p@example.invalid/"]\n\tinsteadOf = https://example.invalid/\n' "$u" ;;
    quoted) printf '[core]\n\trepositoryformatversion = 0\n[remote "origin"]\n\turl = "https://gitlab-ci-token@gitlab.example.invalid/group/project.git"\n' ;;
  esac >"$C/fx/git-config"
}
g_extra() { printf 'NS=/builds/grp P=/builds/grp/proj\n%s\n' "$1" >"$C/fx/extra.sh"; } # SH: case setup in the container
g_env() { # [-NAME | NAME=VALUE]...: the job environment file (mode 600), without NAME or with NAME=VALUE
  local line k o skip=" "
  local -A setv=()
  for o in "$@"; do if [[ "$o" == -* ]]; then skip+="${o#-} "; else setv[${o%%=*}]="${o#*=}"; fi; done
  : >"$C/env"
  chmod 600 "$C/env"
  for line in GITLAB_CI=true CI_JOB_ID=777 "CI_PROJECT_DIR=$G_NS/proj" DOCKER_HOST=tcp://docker:2376 DOCKER_TLS_VERIFY=1 \
    DOCKER_CERT_PATH=/certs/client DOCKER_TLS_CERTDIR=/certs GSTATE=/gs PLATFORM_CI_GUARD_INTERVAL=1 CI_PIPELINE_SOURCE=push \
    CI_COMMIT_BRANCH=feature CI_COMMIT_REF_NAME=feature "CI_JOB_TOKEN=$G_TOKEN" "CI_REGISTRY_PASSWORD=$G_REGPW" \
    PATH=/fx/shim:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin; do
    k="${line%%=*}"
    if [[ "$skip" == *" $k "* ]]; then continue; fi
    if [[ -n "${setv[$k]+x}" ]]; then line="$k=${setv[$k]}"; unset "setv[$k]"; fi
    printf '%s\n' "$line" >>"$C/env"
  done
  for k in "${!setv[@]}"; do printf '%s=%s\n' "$k" "${setv[$k]}" >>"$C/env"; done
}
g_case() { # [CONTEXT_BEHAVIOUR]: a fresh case with the default Git configuration (username-only) and environment
  C="$ROOT/cases/$CTL_ID"
  rm -rf -- "$C"
  mkdir -p "$C/fx/shim" "$C/gs"
  ctl_labels
  if [[ -z "$CLI_IMAGE" ]]; then prob "infra/docker/ci-context.Dockerfile has no pinned docker CLI stage"; fi
  cp -- "$REPO_R/infra/ci/hosted-context.sh" "$C/fx/hosted-context.sh"
  cp -- "$REPO_R/infra/docker/ci-context.Dockerfile" "$C/fx/ci-context.Dockerfile"
  cat >"$C/fx/shim/docker" <<'GSHIM'
#!/bin/sh
# docker shim of validate-controls.sh group G: records each call's arguments (never the environment) and plays the
# job daemon from the case state directory.
GS="${GSTATE:?}"
printf '%s\n' "$*" >>"$GS/docker.log"
case "${1:-}" in
  version) if [ -e "$GS/daemon-down" ]; then exit 1; fi; exit 0 ;;
  image) exit 1 ;;
  build) cat >/dev/null; exit 0 ;;
  container) case " $* " in *" platform-ci-777 "*) if [ -e "$GS/ctx-running" ]; then exit 0; fi ;; esac; exit 1 ;;
  stop) : >"$GS/stopped"; exit 0 ;;
  exec) if [ -e "$GS/exec-hang" ]; then exec sleep 30; fi; cat "$GS/daemon-df"; exit 0 ;;
  volume) if [ "${2:-}" = create ] && [ -e "$GS/stg-vol-fail" ]; then exit 1; fi; exit 0 ;;
  run)
    case " $* " in
      *" --name platform-ci-777-stg "*)
        if [ -e "$GS/stg-fail" ]; then exit 1; fi
        cut -d= -f1 | sort | tr '\n' ' ' >"$GS/stg-delivered-keys"; exit 0 ;;
      *" --name platform-ci-777-after "*) : >"$GS/after-ran"; exit 0 ;;
      *" --name platform-ci-777 "*) : >"$GS/context-ran"; exec sh /fx/context.sh ;;
      *" -d "*" --name platform-ci-guard-777 "*) : >"$GS/sampler-ran"; echo 0123456789ab; exit 0 ;;
      *" --label platform.ci.guard=777 "*) cat "$GS/daemon-df"; exit 0 ;;
    esac ;;
esac
exit 0
GSHIM
  cat >"$C/fx/setup.sh" <<'GSETUP'
#!/bin/sh
# group G case setup: the build directory on the tmpfs, then the adapter as the job script (it becomes PID 1).
set -eu
P="$CI_PROJECT_DIR" NS="$(dirname "$CI_PROJECT_DIR")"
mkdir -p "$P/.git" "$P/infra/ci" "$P/infra/docker"
cp /fx/hosted-context.sh "$P/infra/ci/hosted-context.sh"
cp /fx/ci-context.Dockerfile "$P/infra/docker/ci-context.Dockerfile"
printf 'readme\n' >"$P/README.md"
mkdir -p "$P.tmp"
printf 'a file variable of the runner\n' >"$P.tmp/FILE_VAR"
while IFS= read -r line || [ -n "$line" ]; do
  case "$line" in
    *@TOKEN@*) printf '%s%s%s\n' "${line%%@TOKEN@*}" "${CI_JOB_TOKEN:-}" "${line#*@TOKEN@}" ;;
    *) printf '%s\n' "$line" ;;
  esac
done </fx/git-config >"$P/.git/config"
if [ -f /fx/extra.sh ]; then . /fx/extra.sh; fi
# Daemon storage reads as the build directory's filesystem unless the case wrote its own figures.
[ -f /gs/daemon-df ] || df -Pk "$NS" >/gs/daemon-df
exec sh "$P/infra/ci/hosted-context.sh" "$@"
GSETUP
  chmod 755 "$C/fx/shim/docker"
  g_context "${1:-healthy}"
  g_config username
  g_env
}
g_run() { # ADAPTER_ARGS...: the case setup, then the adapter, in the pinned CLI image; RC and $C/out
  launch "$C/out" "$REAL_DOCKER" run --rm --network none --pull never "${LB[@]}" --user "$(id -u):$(id -g)" \
    --cap-drop DAC_OVERRIDE --cap-drop DAC_READ_SEARCH --mount "type=tmpfs,destination=/builds,tmpfs-size=$G_TMPFS,tmpfs-mode=1777" \
    -v "$C/fx:/fx:ro" -v "$C/gs:/gs" --env-file "$C/env" --entrypoint sh "$CLI_IMAGE" /fx/setup.sh "$@"
  await "$LPID" 180
}
g_rc() { if [[ "$RC" != "$1" ]]; then prob "adapter exit $RC, want $1 [$(tail -1 "$C/out" 2>/dev/null | cut -c1-160)]"; fi; }
g_has() { if grep -qF -- "$1" "$C/out" 2>/dev/null; then :; else prob "adapter output lacks [$1]"; fi; }
g_no_docker() { if [[ -s "$C/gs/docker.log" ]]; then prob "a docker call was made: [$(head -1 "$C/gs/docker.log" | cut -c1-80)]"; fi; }
g_no_context() { if [[ -e "$C/gs/context-ran" ]]; then prob "the context was started"; fi; }
g_context_ran() { if [[ -e "$C/gs/context-ran" ]]; then :; else prob "the context was not started"; fi; }
g_call() { if grep -qxF -- "$1" "$C/gs/docker.log" 2>/dev/null; then :; else prob "no daemon call [$1]"; fi; }
g_secret_free() { # neither synthetic credential in the adapter output, the call log or the state directory
  local f
  for f in "$C/out" "$C/gs"/*; do
    if [[ -f "$f" ]] && grep -qF -e "$G_TOKEN" -e "$G_REGPW" -- "$f"; then prob "a synthetic credential reached ${f##*/}"; fi
  done
}
g_calls_owned() { # every daemon call is one of the adapter's own: no prune, kill, or stop/removal of anything else
  local line
  while IFS= read -r line; do
    case "$line" in
      version|"image inspect platform-ci-context:777"|"build -q -t platform-ci-context:777 -") ;;
      "container inspect platform-ci-777"|"container inspect platform-ci-guard-777"|"stop -t 150 platform-ci-777") ;;
      "rm -f platform-ci-guard-777"|"exec platform-ci-guard-777 df -Pk /") ;;
      "run --rm --network none --label platform.ci.guard=777 busybox:1.37.0 df -Pk /") ;;
      "run -d --rm --name platform-ci-guard-777 --network none --label platform.ci.guard=777 busybox:1.37.0 sleep 86400") ;;
      "run --rm --init --name platform-ci-777 "*|"run --rm --init --name platform-ci-777-after "*) ;;
      "volume create --label platform.ci.staging=777 platform-ci-stg-777"|"volume rm -f platform-ci-stg-777") ;;
      "run --rm -i --network none --name platform-ci-777-stg --label platform.ci.staging=777 -v platform-ci-stg-777:/d platform-ci-context:777 sh -c umask 077 && cat >/d/env && test -s /d/env") ;;
      *) prob "unexpected daemon call [$line]" ;;
    esac
  done <"$C/gs/docker.log"
}
g_cap() { # FS: the adapter's capacity line for FS
  grep "^hosted-context: capacity: fs=$1 " "$C/out" 2>/dev/null | tail -1 || true
}
ctl_G1() {
  local r1 r2
  g_case; g_env -GITLAB_CI; g_run -- --profile fast; r1="$RC"; g_rc 16; g_no_docker
  g_has "refused: not a GitLab CI job with the job-private TLS daemon (no Docker call made)"
  g_case; g_env -DOCKER_TLS_VERIFY; g_run -- --profile fast; r2="$RC"; g_rc 16; g_no_docker
  g_case; g_env DOCKER_HOST=tcp://localhost:2375; g_run -- --profile fast; g_rc 16; g_no_docker
  detail "outside CI exit $r1; without DOCKER_TLS_VERIFY exit $r2; another daemon address exit $RC; no docker call in any"
}
ctl_G2() {
  local r1
  g_case "exit 0"; g_run -- --profile fast; r1="$RC"; g_rc 0; g_context_ran; g_has "final: context=0 guard=ok"
  g_case "exit 14"; g_run -- --profile fast; g_rc 14; g_has "final: context=14 guard=ok"
  detail "context exit 0 -> adapter $r1; context exit 14 -> adapter $RC"
}
ctl_G3() {
  local line tok=() i e n=0 vols=() mnts=() bad=""
  g_case; g_run -- --profile gate --publish-dir ci-evidence; g_rc 0; g_context_ran
  line="$(grep -E -- '^run .* --name platform-ci-777 ' "$C/gs/docker.log" 2>/dev/null | head -1)" || line=""
  if [[ -z "$line" ]]; then prob "no context run recorded"; return 0; fi
  read -r -a tok <<<"$line"
  for ((i = 0; i < ${#tok[@]}; i++)); do
    case "${tok[i]}" in
      --mount) mnts+=("${tok[i + 1]:-}"); i=$((i + 1)) ;;
      --privileged*|--pid*|--cap-add*|--device*|--userns*|--ipc*|--uts*|--volumes-from*|--mount=*|--group-add*) bad+=" ${tok[i]}" ;;
      *docker.sock*) bad+=" a docker socket" ;;
      -v) vols+=("${tok[i + 1]:-}"); i=$((i + 1)) ;;
      -e)
        e="${tok[i + 1]:-}"; i=$((i + 1))
        if [[ "$e" == *=* ]]; then
          case "$e" in
            "TMPDIR=$G_NS/.platform-tmp-777"|DOCKER_HOST=tcp://localhost:2376|DOCKER_TLS_VERIFY=1|DOCKER_CERT_PATH=/certs/client|PLATFORM_CI_CONTEXT=hosted-context/1) ;;
            *) bad+=" -e ${e%%=*}=<value>" ;;
          esac
        elif [[ -n "${G_ALLOW[$e]+x}" ]]; then n=$((n + 1))
        else bad+=" -e $e"; fi ;;
    esac
  done
  if [[ -n "$bad" ]]; then prob "forbidden context arguments:$bad"; fi
  if [[ "${vols[*]}" != "$G_NS:$G_NS /certs/client:/certs/client:ro" ]]; then prob "context mounts [${vols[*]}], want exactly the build directory parent and the client certificates (read-only)"; fi
  if [[ "${mnts[*]}" != "type=tmpfs,destination=$G_NS/proj.tmp,tmpfs-size=1m" ]]; then prob "context --mount [${mnts[*]}], want exactly the empty mask over the runner's proj.tmp"; fi
  for e in "--init" "--network host" "--security-opt no-new-privileges" "-w $G_NS/proj"; do
    if [[ " $line " != *" $e "* ]]; then prob "the context lacks [$e]"; fi
  done
  if [[ "$(grep -o -- '--security-opt' <<<"$line" | wc -l)" != 1 ]]; then prob "more than one --security-opt"; fi
  if [[ "$line" != *" platform-ci-context:777 bash infra/ci/validate.sh --profile gate --publish-dir ci-evidence" ]]; then prob "the context does not run the entry point with the job's arguments"; fi
  if (( n != ${#G_ALLOW[@]} )); then prob "allow-listed names passed $n, want ${#G_ALLOW[@]}"; fi
  g_secret_free; g_calls_owned
  detail "exit $RC; context: --init, --network host, no-new-privileges, mounts exactly $G_NS and /certs/client:ro with the runner's proj.tmp masked by an empty tmpfs, $n allow-listed -e NAME plus 5 fixed values, no privilege/PID/device/socket; canary credentials absent from argv, output and state"
}
ctl_G4() {
  local r1 r2 r3 line
  g_case; : >"$C/gs/ctx-running"; g_run --after-script; r1="$RC"; g_rc 0
  g_call "stop -t 150 platform-ci-777"; g_has "after-script: stopping the remaining context"; g_has "scan=clean"
  line="$(grep -E -- '^run .* --name platform-ci-777-after ' "$C/gs/docker.log" 2>/dev/null | head -1)" || line=""
  if [[ "$line" != *" platform-ci-context:777 bash infra/ci/cleanup-owned.sh --after-script" ]]; then prob "no backstop cleanup context"; fi
  g_calls_owned; g_secret_free
  g_case; g_extra 'mkdir -p "$NS/cache"; printf "t=%s\n" "$CI_JOB_TOKEN" >"$NS/cache/creds"'; g_run --after-script; r2="$RC"; g_rc 0
  if [[ -e "$C/gs/after-ran" ]]; then prob "a backstop context started without proven credentials"; fi
  g_has "after-script: credentials not proven absent; no context started"; g_secret_free
  g_case; g_extra 'mkdir -p "$NS/.platform-tmp-777/w"; mkfifo "$NS/.platform-tmp-777/w/fifo"; printf "x\n" >"$NS/.platform-tmp-777/w/left"; chmod 000 "$NS/.platform-tmp-777/w/left"; mkfifo "$NS/elsewhere.fifo"'
  g_run --after-script; r3="$RC"; g_rc 0; g_has "scan=clean"
  if [[ -e "$C/gs/after-ran" ]]; then :; else prob "the backstop did not run past a FIFO and leftovers of the context temporary directory"; fi
  g_case; g_env -GITLAB_CI; g_run --after-script; g_rc 0; g_no_docker
  detail "running context stopped then backstop cleanup context, exit $r1; residual credential: no context, exit $r2; FIFOs and an unreadable leftover in the context temporary directory: backstop ran, exit $r3; outside CI: no docker call, exit $RC"
}
ctl_G5() {
  local r1
  g_case; g_extra 'mkdir "$NS/.platform-tmp-777"'; g_run -- --profile fast; r1="$RC"; g_rc 16; g_no_context
  g_has "refused: unsafe or pre-existing temporary directory"
  g_case; g_extra 'ln -s /tmp "$NS/.platform-tmp-777"'; g_run -- --profile fast; g_rc 16; g_no_context
  g_has "refused: unsafe or pre-existing temporary directory"
  detail "a pre-existing temporary directory exit $r1; a symlink in its place exit $RC; no context in either"
}
ctl_G6a() {
  g_case config-copy; g_run -- --profile fast; g_rc 0; g_context_ran
  g_has "credentials: remotes=1 sanitized=0 includes=0 include_inside=0 include_outside=0 include_absent=0 password_urls=0 token_files=0 scan=clean"
  if cmp -s "$C/fx/git-config" "$C/gs/config-after"; then :; else prob "the username-only configuration changed"; fi
  g_secret_free
  detail "exit $RC; a username-only remote is accepted unchanged; the context started"
}
ctl_G6b() {
  g_case config-copy; g_config password; g_run -- --profile fast; g_rc 0; g_context_ran
  g_has "credentials: remotes=1 sanitized=1 includes=0 include_inside=0 include_outside=0 include_absent=0 password_urls=0 token_files=0 scan=clean"
  if grep -qxF $'\turl = https://gitlab-ci-token@gitlab.example.invalid/group/project.git' "$C/gs/config-after" 2>/dev/null; then :
  else prob "the remote was not reduced to the username form"; fi
  g_secret_free; g_calls_owned
  detail "exit $RC; the job token is removed from the remote URL before the context mounts the build directory (scheme, user, host and path kept); the token is in no argument, log or state file"
}
ctl_G6c() {
  g_case; g_extra 'mkdir -p "$NS/cache"; printf "t=%s\n" "$CI_JOB_TOKEN" >"$NS/cache/creds"'; g_run -- --profile fast; g_rc 16; g_no_context
  g_has "credentials: a job credential is stored in 1 file(s) under the build directory (fail closed)"
  g_has "refused: credentials not proven absent; no context started"; g_secret_free
  detail "exit $RC; a residual job token in another file under the build directory is refused; no context"
}
ctl_G6d() {
  local r1
  g_case; g_extra 'printf "x\n" >"$NS/locked.txt"; chmod 000 "$NS/locked.txt"'; g_run -- --profile fast; r1="$RC"; g_rc 16; g_no_context
  g_has "credentials: the build-directory scan failed or met unreadable entries (fail closed)"
  g_case; g_extra 'mkdir "$NS/lockdir"; printf "x\n" >"$NS/lockdir/f"; chmod 000 "$NS/lockdir"'; g_run -- --profile fast; g_rc 16; g_no_context
  g_has "credentials: the build-directory scan failed or met unreadable entries (fail closed)"
  detail "an unreadable file exit $r1; an unreadable directory exit $RC (busybox grep -r alone would report it as no match); the scan is inconclusive and refused; no context"
}
ctl_G6e() {
  local r1
  g_case; g_config include-pw; g_extra 'printf "[remote \"mirror\"]\n\turl = https://mirror-user:not-the-job-token@mirror.example.invalid/x.git\n" >"$P/.git/inc.conf"'
  g_run -- --profile fast; r1="$RC"; g_rc 16; g_no_context; g_has "credentials: an included configuration holds a password-bearing URL"
  g_case; g_config include-nested; g_extra 'printf "[include]\n\tpath = b.conf\n" >"$P/.git/a.conf"; printf "[core]\n" >"$P/.git/b.conf"'
  g_run -- --profile fast; g_rc 16; g_no_context; g_has "credentials: nested include (fail closed)"
  detail "an included configuration with a password-bearing URL exit $r1; a nested include exit $RC; no context"
}
ctl_G6f() {
  g_case config-copy; g_config include-ok; g_extra 'printf "[core]\n\tbare = false\n" >"$P/.git/clean.conf"'
  g_run -- --profile fast; g_rc 0; g_context_ran
  g_has "credentials: remotes=1 sanitized=0 includes=2 include_inside=1 include_outside=0 include_absent=1 password_urls=0 token_files=0 scan=clean"
  detail "exit $RC; a clean include inside the build directory is examined, an absent outside include is counted; the context started"
}
ctl_G6g() {
  local r1
  g_case; g_config header-cred; g_run -- --profile fast; r1="$RC"; g_rc 16; g_no_context
  g_has "credentials: unparseable or unsafe Git configuration (fail closed)"
  g_case; g_config quoted; g_run -- --profile fast; g_rc 16; g_no_context
  g_has "credentials: unparseable or unsafe Git configuration (fail closed)"
  detail "a credential URL in a section name exit $r1; a quoted remote URL exit $RC; no context"
}
ctl_G6h() {
  g_case; g_env -CI_JOB_TOKEN; g_run -- --profile fast; g_rc 16; g_no_context
  g_has "credentials: no job credential to check against"
  detail "exit $RC; without a job token nothing can be proven absent; no context"
}
ctl_G7a() {
  local b s
  g_case healthy; g_run -- --profile fast; g_rc 0; g_context_ran
  g_has "shared=yes"; g_has "capacity: fs=daemon-storage shared_with=build-directory (same filesystem; not counted twice)"
  b="$(g_cap build-directory)"; s="$(sed -n 's/.* samples=\([0-9]*\) .*/\1/p' <<<"$b")"
  if [[ "$b" != *" durable_floor_gib=35 baseline_meets_durable_floor=no" ]]; then prob "build-directory capacity [$b]"; fi
  if [[ -z "$s" ]] || (( s < 3 )); then prob "the guard took ${s:-no} samples, want at least 3"; fi
  g_has "final: context=0 guard=ok"; g_calls_owned
  detail "exit $RC; shared filesystem confirmed; $s samples; durable floor max(35, ceil(1.5 x peak + 4)) = 35 GiB reported"
}
ctl_G7b() {
  g_case fill; g_run -- --profile fast; g_rc 143
  g_has "disk guard: FAIL (available space below 4 GiB); stopping the context"
  g_has "final: context=143 guard=FAIL (available space below 4 GiB)"; g_call "stop -t 150 platform-ci-777"; g_calls_owned
  detail "exit $RC; below 4 GiB during the workload: the owned context was stopped, the job fails; no other daemon call"
}
ctl_G7c() {
  g_case vanish; g_run -- --profile fast; g_rc 143
  g_has "final: context=143 guard=FAIL (measurement failed)"; g_call "stop -t 150 platform-ci-777"; g_calls_owned
  detail "exit $RC; a failed measurement fails the guard and stops the owned context"
}
ctl_G7d() {
  g_case kill-guard; g_run -- --profile fast; g_rc 143
  if [[ "$(grep -c . "$C/gs/killed" 2>/dev/null || true)" != 1 ]]; then prob "the guard was not identified exactly once"; fi
  g_has "final: context=143 guard=FAIL (guard died)"; g_call "stop -t 150 platform-ci-777"; g_calls_owned
  detail "exit $RC; the guard killed outright: its death is detected, the owned context stopped, the job fails"
}
ctl_G7e() {
  g_case race; g_run -- --profile fast; g_rc 20
  g_has "final: context=0 guard=FAIL (available space below 4 GiB)"
  detail "exit $RC; the context succeeded while the guard failed: the failure is kept (20)"
}
ctl_G7f() {
  g_case; G_TMPFS=4000m g_run -- --profile fast; g_rc 20; g_no_context
  g_has "disk guard: FAIL (available space below 4 GiB before the workload); no context started"
  detail "exit $RC; below 4 GiB at the baseline: no context"
}
ctl_G7g() {
  local d
  g_case daemon-drop
  g_extra 'printf "Filesystem 1024-blocks Used Available Capacity Mounted on\noverlay 104857600 57671680 47185920 55%% /\n" >/gs/daemon-df'
  g_run -- --profile fast; g_rc 0; g_has "shared=no"
  g_call "run -d --rm --name platform-ci-guard-777 --network none --label platform.ci.guard=777 busybox:1.37.0 sleep 86400"
  g_call "exec platform-ci-guard-777 df -Pk /"; g_call "rm -f platform-ci-guard-777"
  d="$(g_cap daemon-storage)"
  if [[ "$d" != "hosted-context: capacity: fs=daemon-storage baseline_avail_bytes=48318382080 min_avail_bytes=16106127360 incremental_peak_bytes=32212254720 samples="*" durable_floor_gib=49 baseline_meets_durable_floor=no" ]]; then
    prob "daemon-storage capacity [$d]"
  fi
  if [[ "$(g_cap build-directory)" != *" durable_floor_gib=35 baseline_meets_durable_floor=no" ]]; then prob "build-directory capacity [$(g_cap build-directory)]"; fi
  g_calls_owned
  detail "exit $RC; distinct filesystems both monitored through one sampler container (started, sampled, removed); 45 -> 15 GiB: incremental peak 30 GiB, durable floor ceil(1.5 x 30 + 4) = 49 GiB, baseline 45 GiB does not meet it"
}
ctl_G7h() {
  g_case hang-sample
  g_extra 'printf "Filesystem 1024-blocks Used Available Capacity Mounted on\noverlay 104857600 57671680 47185920 55%% /\n" >/gs/daemon-df'
  g_run -- --profile fast; g_rc 143; g_has "shared=no"
  g_has "final: context=143 guard=FAIL (measurement failed)"; g_call "stop -t 150 platform-ci-777"; g_calls_owned
  detail "exit $RC; a daemon-storage sample that hangs is cut off by its time limit: measurement failed, the owned context stopped"
}
ctl_G7i() {
  g_case stop-guard; g_run -- --profile fast; g_rc 143
  if [[ "$(grep -c . "$C/gs/killed" 2>/dev/null || true)" != 1 ]]; then prob "the guard was not identified exactly once"; fi
  g_has "final: context=143 guard=FAIL (guard stalled)"; g_call "stop -t 150 platform-ci-777"; g_calls_owned
  g_has "reserve 4 GiB; interval 1s; stall 43s"
  # The production window: the default 30 s interval gives a 420 s stall window (it was 180 s), with the floor unchanged.
  g_case healthy; g_env -PLATFORM_CI_GUARD_INTERVAL; g_run -- --profile fast; g_rc 0
  g_has "reserve 4 GiB; interval 30s; stall 420s"; g_has "final: context=0 guard=ok"
  # A heartbeat read that finds the file empty (mid-rewrite) is not a stall: the context truncates it for 6 s.
  g_case empty-beat; g_run -- --profile fast; g_rc 0
  g_has "final: context=0 guard=ok"
  detail "exit $RC; a guard that stays alive but stops sampling (SIGSTOP) is detected as stalled (43 s window at a 1 s interval); the default window is 420 s; an emptied heartbeat file is not a stall; the owned context stopped"
}

# --- group H: the release helper (ONEOFF RC qualification; infra/ci/release-verify.sh of this checkout) ------------
# Fixtures: the RC promotion pipeline of vu_make rc (its repository: main R0 = the template commit, staging R1 a
# documentation-only commit, the annotated tag 0.1.0-rc.1 at R1; its bare remote with release/0.1.0-rc.1 at R1; the
# published fast and gate partitions of pipeline 1001, jobs 4504/4505, MR !7, project 999999999), with this checkout's
# release-verify.sh and pipeline-union.sh beside it, the GitLab API JSON as data files, and a receipt. Faults edit one
# input (yq through the pinned image, sed, or refs of the bare remote); the helper must print exactly one FAIL line.
RV_PEEL="" RV_OBJ="" RV_TREE="" RV_MAIN="" RV_LINE=""
RV_QREC=(PROJECT_ID=999999999 MR_IID=7 PIPELINE_ID=1001 FAST_JOB_ID=4504 GATE_JOB_ID=4505)
RV_ID=(-c user.name=ci-controls -c user.email=controls@example.invalid)
rv_case() { # a fresh case with the RC fixtures in $C (repo, remote.git, fast, gate) and the API JSON
  new_case
  vu_make rc || return 1
  rm -rf -- "$C/repo" "$C/remote.git" "$C/fast" "$C/gate"
  cp -a -- "$ROOT/vu/rc/repo" "$C/repo"; cp -a -- "$ROOT/vu/rc/remote.git" "$C/remote.git"
  cp -a -- "$ROOT/vu/rc/fast" "$C/fast"; cp -a -- "$ROOT/vu/rc/gate" "$C/gate"
  cp -- "$REPO_R/infra/ci/release-verify.sh" "$REPO_R/infra/ci/pipeline-union.sh" "$REPO_R/infra/ci/lib.sh" "$C/repo/infra/ci/"
  RV_PEEL="$(git -C "$C/repo" rev-parse 'refs/tags/0.1.0-rc.1^{commit}')"
  RV_OBJ="$(git -C "$C/repo" rev-parse refs/tags/0.1.0-rc.1)"
  RV_TREE="$(git -C "$C/repo" rev-parse "$RV_PEEL^{tree}")"
  RV_MAIN="$(git -C "$C/repo" rev-parse refs/heads/main)"
  ctl_labels
  printf '{"iid": 7, "project_id": 999999999, "source_project_id": 999999999, "target_project_id": 999999999, "source_branch": "release/0.1.0-rc.1", "target_branch": "main", "sha": "%s", "state": "opened", "squash": false, "merge_when_pipeline_succeeds": false, "head_pipeline": {"id": 1001, "sha": "%s", "status": "success"}}\n' \
    "$RV_PEEL" "$RV_PEEL" >"$C/mr.json"
  printf '{"id": 1001, "iid": 7, "project_id": 999999999, "sha": "%s", "ref": "refs/merge-requests/7/head", "status": "success", "source": "merge_request_event", "tag": false}\n' \
    "$RV_PEEL" >"$C/pipeline.json"
  printf '[{"id": 4505, "name": "validate:integration", "status": "success", "allow_failure": false, "pipeline": {"id": 1001, "project_id": 999999999}, "commit": {"id": "%s"}}, {"id": 4504, "name": "validate:checks", "status": "success", "allow_failure": false, "pipeline": {"id": 1001, "project_id": 999999999}, "commit": {"id": "%s"}}]\n' \
    "$RV_PEEL" "$RV_PEEL" >"$C/jobs.json"
  git ls-remote "$C/remote.git" >"$C/ls.txt"
}
rv_receipt() { # [KEY=VALUE]...: $C/receipt, the cut and candidate record of the fixture RC plus KEY=VALUE lines
  local kv
  {
    printf 'FORMAT=release-receipt/1\nTAG=0.1.0-rc.1\nCUT_STAGING=%s\nCUT_LSREMOTE_SHA256=%s\n' "$RV_PEEL" "$(printf 'c%.0s' {1..64})"
    printf 'TAG_OBJECT=%s\nPEELED=%s\nTREE=%s\n' "$RV_OBJ" "$RV_PEEL" "$RV_TREE"
    for kv in "$@"; do printf '%s\n' "$kv"; done
  } >"$C/receipt"
}
rv_cut_receipt() { printf 'FORMAT=release-receipt/1\nTAG=0.1.0-rc.1\nCUT_STAGING=%s\nCUT_LSREMOTE_SHA256=%s\n' "$RV_PEEL" "$(printf 'c%.0s' {1..64})" >"$C/receipt"; }
rv() { # ARGS...: release-verify.sh of the case repository; RC and RV_LINE (its one output line)
  RC=0
  env -i -C "$C" "${BASE_ENV[@]}" TMPDIR="$C/tmp" SHIM_LABELS="$(case_labels)" bash "$C/repo/infra/ci/release-verify.sh" "$@" >"$C/rv.out" 2>&1 || RC=$?
  RV_LINE="$(cat "$C/rv.out" 2>/dev/null)"
  if [[ "$(grep -c . "$C/rv.out" 2>/dev/null)" != 1 || "$RV_LINE" != "release-verify: "* ]]; then prob "the helper did not print exactly one result line: [$(tr '\n' ';' <"$C/rv.out" 2>/dev/null)]"; fi
}
rv_pass() { rv "$@"; if [[ "$RC" != 0 || "$RV_LINE" != "release-verify: PASS "* ]]; then prob "$1: exit $RC [$RV_LINE], want 0 and PASS"; fi; }
rv_fail() { # REASON ARGS...: exit 1 and a FAIL line containing REASON
  local why="$1"; shift
  rv "$@"
  if [[ "$RC" != 1 || "$RV_LINE" != "release-verify: FAIL "*"$why"* ]]; then prob "$1: exit $RC [$RV_LINE], want 1 and FAIL [..$why..]"; fi
}
rv_q() { rv_qx "$C/receipt" 999999999 "$@"; }
rv_qx() { # RECEIPT PROJECT [fail REASON]: qualified over the case JSON and partitions
  local rf="$1" pid="$2"
  shift 2
  if [[ "${1:-}" == fail ]]; then rv_fail "$2" qualified 0.1.0-rc.1 "$rf" "$pid" "$C/mr.json" "$C/pipeline.json" "$C/jobs.json" "$C/fast" "$C/gate"
  else rv_pass qualified 0.1.0-rc.1 "$rf" "$pid" "$C/mr.json" "$C/pipeline.json" "$C/jobs.json" "$C/fast" "$C/gate"; fi
}
rv_commit() { # PARENT TREE_FROM [PATH=CONTENT]...: a new commit (no work tree change) with PARENT's or TREE_FROM's tree plus the files
  local parent="$1" from="$2" kv blob idx="$C/tmp/idx.$RANDOM"
  shift 2
  GIT_INDEX_FILE="$idx" git -C "$C/repo" read-tree "$from"
  for kv in "$@"; do
    blob="$(printf '%s\n' "${kv#*=}" | git -C "$C/repo" hash-object -w --stdin)"
    GIT_INDEX_FILE="$idx" git -C "$C/repo" update-index --add --cacheinfo "100644,$blob,${kv%%=*}"
  done
  GIT_AUTHOR_DATE=2026-01-03T00:00:00Z GIT_COMMITTER_DATE=2026-01-03T00:00:00Z \
    git -C "$C/repo" "${RV_ID[@]}" commit-tree "$(GIT_INDEX_FILE="$idx" git -C "$C/repo" write-tree)" -p "$parent" -m "control commit"
  rm -f -- "$idx"
}
rv_merge() { # TREE P1 P2: a merge commit with exactly these parents
  GIT_AUTHOR_DATE=2026-01-04T00:00:00Z GIT_COMMITTER_DATE=2026-01-04T00:00:00Z \
    git -C "$C/repo" "${RV_ID[@]}" commit-tree "$1" -p "$2" -p "$3" -m "merge"
}
rv_retag() { # NAME TARGET [lightweight]: replace the tag NAME locally and in the bare remote
  git -C "$C/repo" tag -d "$1" >/dev/null 2>&1 || true
  if [[ "${3:-}" == lightweight ]]; then git -C "$C/repo" tag "$1" "$2"
  else GIT_COMMITTER_DATE=2026-01-05T00:00:00Z git -C "$C/repo" "${RV_ID[@]}" tag -a -m "replaced $1" "$1" "$2"; fi
  git -C "$C/repo" push -q -f "$C/remote.git" "refs/tags/$1:refs/tags/$1"
  git ls-remote "$C/remote.git" >"$C/ls.txt"
}

ctl_R1() {
  local n bad=""
  rv_case || return 0
  for n in 0.1.0-rc.1 10.20.30-rc.12 0.0.0-rc.1; do rv_pass name "$n"; done
  for n in 01.0.0-rc.1 0.1.0-rc.0 v0.1.0-rc.1 0.1.0-RC.1 0.1.0-rc.1+b1 1.0.0 0.1.0-rc.01 0.1-rc.1 0.1.0-rc; do
    rv_fail "not a canonical RC name" name "$n"; bad+=" $n"
  done
  rv_pass name 0.2.0-rc.1 "$C/ls.txt" "$C/receipt-0.2"
  if grep -qx "CUT_STAGING=$(git -C "$C/remote.git" rev-parse refs/heads/staging)" "$C/receipt-0.2" 2>/dev/null; then :; else prob "the cut receipt does not record the listed staging commit"; fi
  rv_fail "receipt already exists" name 0.2.0-rc.1 "$C/ls.txt" "$C/receipt-0.2"
  rv_fail "must be absent" name 0.1.0-rc.1 "$C/ls.txt" "$C/receipt-0.1"
  git -C "$C/remote.git" branch release/0.3.0-rc.1 "$RV_PEEL"; git ls-remote "$C/remote.git" >"$C/ls.txt"
  rv_fail "must be absent" name 0.3.0-rc.1 "$C/ls.txt" "$C/receipt-0.3"
  rv_fail "not a regular file" candidate 0.4.0-rc.1 "$C/ls.txt" "$C/receipt-none"
  detail "canonical names accepted; refused:$bad; with the pre-cut listing: absence recorded in a new receipt; reused receipt, existing tag and existing release branch refused; candidate before a cut receipt refused"
}
ctl_R2a() { rv_case || return 0; rv_cut_receipt
  rv_pass candidate 0.1.0-rc.1 "$C/ls.txt" "$C/receipt"
  if grep -qx "TAG_OBJECT=$RV_OBJ" "$C/receipt" && grep -qx "PEELED=$RV_PEEL" "$C/receipt" && grep -qx "TREE=$RV_TREE" "$C/receipt"; then :; else prob "the receipt does not record the tag object, peel and tree"; fi
  rv_pass candidate 0.1.0-rc.1 "$C/ls.txt" "$C/receipt"
  detail "annotated tag object and peel equal the remote lines, peel in staging: recorded, then re-verified unchanged"; }
ctl_R2b() { rv_case || return 0; rv_cut_receipt
  git -C "$C/remote.git" tag -d 0.1.0-rc.1 >/dev/null
  GIT_COMMITTER_DATE=2026-01-05T00:00:00Z git -C "$C/remote.git" "${RV_ID[@]}" tag -a -m moved 0.1.0-rc.1 "$RV_MAIN"
  git ls-remote "$C/remote.git" >"$C/ls.txt"
  rv_fail "differs from the remote one" candidate 0.1.0-rc.1 "$C/ls.txt" "$C/receipt"
  detail "exit $RC: the remote tag moved to another commit (local and remote differ)"; }
ctl_R2c() { rv_case || return 0; rv_cut_receipt
  rv_pass candidate 0.1.0-rc.1 "$C/ls.txt" "$C/receipt"
  rv_retag 0.1.0-rc.1 "$RV_PEEL"
  rv_fail "substituted or recreated tag" candidate 0.1.0-rc.1 "$C/ls.txt" "$C/receipt"
  detail "exit $RC: same peel, another annotated tag object than the cut receipt's"; }
ctl_R2d() { rv_case || return 0; rv_cut_receipt
  grep -F "$(printf '\t')refs/tags/0.1.0-rc.1" "$C/ls.txt" | grep -v '\^{}$' >"$C/dup.txt"; cat "$C/dup.txt" >>"$C/ls.txt"
  rv_fail "want exactly 1" candidate 0.1.0-rc.1 "$C/ls.txt" "$C/receipt"
  detail "exit $RC: a duplicate remote tag line"; }
ctl_R2e() { rv_case || return 0; rv_cut_receipt
  sed -i '/\trefs\/tags\/0\.1\.0-rc\.1/d' "$C/ls.txt"
  rv_fail "want exactly 1" candidate 0.1.0-rc.1 "$C/ls.txt" "$C/receipt"
  detail "exit $RC: the tag is missing from the remote listing"; }
ctl_R2f() { rv_case || return 0; rv_cut_receipt
  rv_retag 0.1.0-rc.1 "$RV_PEEL" lightweight
  rv_fail "refs/tags/0.1.0-rc.1^{}" candidate 0.1.0-rc.1 "$C/ls.txt" "$C/receipt"
  detail "exit $RC: a lightweight tag (no peel line, not annotated)"; }
ctl_R2g() { rv_case || return 0; rv_cut_receipt
  git -C "$C/remote.git" update-ref refs/heads/staging "$RV_MAIN"; git ls-remote "$C/remote.git" >"$C/ls.txt"
  rv_fail "not in the listed staging" candidate 0.1.0-rc.1 "$C/ls.txt" "$C/receipt"
  detail "exit $RC: the RC commit is not in the listed staging"; }
# Codex preliminary review finding 1: the first binding requires peel = the staging commit recorded at cut.
ctl_R2h() { local r1 r2; rv_case || return 0; rv_cut_receipt
  rv_retag 0.1.0-rc.1 "$RV_MAIN"
  rv_fail "not the staging commit recorded at cut" candidate 0.1.0-rc.1 "$C/ls.txt" "$C/receipt"; r1="$RC"
  r2="$(rv_commit "$RV_PEEL" "$RV_PEEL" "docs/later-note.md=a later staging change")"
  git -C "$C/repo" push -q "$C/remote.git" "$r2:refs/heads/staging"
  rv_retag 0.1.0-rc.1 "$r2"
  rv_fail "not the staging commit recorded at cut" candidate 0.1.0-rc.1 "$C/ls.txt" "$C/receipt"
  if grep -q '^TAG_OBJECT=' "$C/receipt"; then prob "a wrong cut was recorded"; fi
  detail "first binding refused for a tag on an older staging ancestor (exit $r1) and on a newer commit than the recorded cut (exit $RC); nothing recorded"; }
ctl_R2i() { local r2; rv_case || return 0; rv_cut_receipt
  rv_pass candidate 0.1.0-rc.1 "$C/ls.txt" "$C/receipt"
  r2="$(rv_commit "$RV_PEEL" "$RV_PEEL" "docs/later-note.md=a later staging change")"
  git -C "$C/repo" push -q "$C/remote.git" "$r2:refs/heads/staging"; git ls-remote "$C/remote.git" >"$C/ls.txt"
  rv_pass candidate 0.1.0-rc.1 "$C/ls.txt" "$C/receipt"
  detail "the unchanged correct cut is still accepted after staging advanced past it (recorded identity, staging ancestry)"; }

ctl_R3a() { rv_case || return 0; rv_receipt; rv_q
  if grep -qx 'PIPELINE_ID=1001' "$C/receipt" && grep -qx 'MR_IID=7' "$C/receipt" && grep -qx 'GATE_JOB_ID=4505' "$C/receipt"; then :; else prob "the receipt does not record the qualified pipeline"; fi
  if [[ "$RV_LINE" != *"eligible (owner promotion instruction still required)"*"fast 32 + remainder 97 -> 125 required IDs incl. ci-controls"* ]]; then prob "the PASS line lacks the eligibility and the recomputed coverage: [$RV_LINE]"; fi
  detail "MR !7 head = RC peel, latest pipeline 1001 successful with both jobs required and successful, artifacts attributed to project/pipeline/job/MR/RC/tree, union PASS, coverage recomputed from the candidate (32 + 97 -> 125 incl. ci-controls): eligible only"; }
ctl_R3b() { rv_case || return 0; rv_receipt; vu_json "$C/pipeline.json" '.status = "failed"'; rv_q fail "did not succeed"; detail "exit $RC: a failed pipeline"; }
ctl_R3c() { rv_case || return 0; rv_receipt; vu_json "$C/jobs.json" '(.[] | select(.name == "validate:integration") | .status) = "manual"'; rv_q fail "validate:integration is manual"; detail "exit $RC: a manual (unplayed) gate"; }
ctl_R3d() { local r1; rv_case || return 0; rv_receipt
  vu_json "$C/jobs.json" '(.[] | select(.name == "validate:integration") | .status) = "skipped"'; rv_q fail "validate:integration is skipped"; r1="$RC"
  rv_case || return 0; rv_receipt; vu_json "$C/jobs.json" 'map(select(.name != "validate:integration"))'; rv_q fail "the pipeline has 1 jobs"
  detail "skipped gate exit $r1; missing gate exit $RC"; }
ctl_R3e() { rv_case || return 0; rv_receipt; vu_json "$C/jobs.json" '(.[] | select(.name == "validate:integration") | .allow_failure) = true'; rv_q fail "is optional (allow_failure)"; detail "exit $RC: an optional gate"; }
ctl_R3f() { rv_case || return 0; rv_receipt; rv_qx "$C/receipt" 999999998 fail "another project"; detail "exit $RC: proof of another project"; }
ctl_R3g() { rv_case || return 0; rv_receipt; vu_json "$C/gate/summary.json" '.ci.pipeline_id = 1002'; rv_q fail "gate: evidence of another pipeline"; detail "exit $RC: a gate artifact of another pipeline"; }
ctl_R3h() { rv_case || return 0; rv_receipt; vu_json "$C/jobs.json" '(.[] | select(.name == "validate:checks") | .id) = 4599'; rv_q fail "fast: evidence of another job"; detail "exit $RC: the fast artifact is not the API job's"; }
ctl_R3i() { local r1; rv_case || return 0; rv_receipt
  vu_json "$C/gate/summary.json" ".tree_sha = \"$(printf 'a%.0s' {1..40})\""; rv_q fail "gate: the candidate tree is not the RC tree"; r1="$RC"
  rv_case || return 0; rv_receipt; vu_json "$C/pipeline.json" ".sha = \"$RV_MAIN\""; rv_q fail "did not test the RC peeled commit"
  detail "another tree exit $r1; a pipeline of another commit exit $RC"; }
# A reduced pair that is self-consistent (pipeline-union passes) without ci-controls is refused by the recomputation.
ctl_R3j() {
  local reg ps fast rem rs ps_ids rcount
  rv_case || return 0; rv_receipt
  reg="$C/gate/required-checks.tsv"
  ps="$(sed -n 's/^  "pipeline_suites": "\([a-z -]*\)",$/\1/p' "$C/gate/summary.json")"
  fast="$(sed -n 's/^CI_SUITES_fast="\([a-z -]*\)"$/\1/p' "$REPO_R/infra/ci/lib.sh")"
  ps="${ps/ ci-controls/}"
  rem="$(for s in $ps; do if [[ " $fast " != *" $s "* ]]; then printf '%s ' "$s"; fi; done)"
  rs="$(awk -F'\t' -v s=" ${rem}final " '$1 != "" && $1 !~ /^#/ && index(s, " " $2 " ") {print $1}' "$reg" | LC_ALL=C sort)"
  ps_ids="$(awk -F'\t' -v s=" $ps $fast final " '$1 != "" && $1 !~ /^#/ && index(s, " " $2 " ") {print $1}' "$reg" | LC_ALL=C sort)"
  rcount="$(grep -c . <<<"$rs")"
  awk -F'\t' '$2 == "ci-controls" {print $1}' "$reg" | while IFS= read -r id; do sed -i "/^$id\t/d" "$C/gate/results.tsv"; done
  vu_json "$C/gate/summary.json" ".pipeline_suites = \"$ps\" | .executed_suites = \"${rem% }\" | .required_count = $rcount | .required_set_sha256 = \"$(printf '%s\n' "$rs" | sha256sum | cut -c1-64)\" | .pipeline_set_sha256 = \"$(printf '%s\n' "$ps_ids" | sha256sum | cut -c1-64)\""
  vu_json "$C/fast/summary.json" ".pipeline_suites = \"$ps\" | .pipeline_set_sha256 = \"$(printf '%s\n' "$ps_ids" | sha256sum | cut -c1-64)\""
  vu_run "$C/fast" "$C/gate"
  if [[ "$RC" != 0 ]]; then prob "the reduced pair is not self-consistent (union exit $RC [$VU_LINE]); the control would be inconclusive"; fi
  rv_q fail "differ from the candidate's recomputation"
  detail "a self-consistent reduced pair without ci-controls (union exit 0) is refused against the candidate's own recomputation (exit $RC)"
}
ctl_R3k() { rv_case || return 0; rv_receipt; vu_json "$C/fast/summary.json" '.tested_content = "mr-source" | .release = null'; rv_q fail "fast: tested_content is not mr-rc-promotion"; detail "exit $RC: wrong tested content"; }
ctl_R3l() { rv_case || return 0; rv_receipt; vu_json "$C/mr.json" '.source_project_id = 99'; rv_q fail "another project or a fork"; detail "exit $RC: a fork merge request (source project differs)"; }
ctl_R3m() { rv_case || return 0; rv_receipt; vu_json "$C/mr.json" ".sha = \"$RV_MAIN\""; rv_q fail "merge request head is not the RC peeled commit"; detail "exit $RC: the merge request head moved (a commit pushed to release/0.1.0-rc.1)"; }
ctl_R3n() { rv_case || return 0; rv_receipt; vu_json "$C/mr.json" '.head_pipeline.id = 1002'; rv_q fail "not the merge request's latest pipeline"; detail "exit $RC: a newer pipeline is the merge request's latest"; }
ctl_R3o() { local r1; rv_case || return 0; rv_receipt
  vu_json "$C/mr.json" '.merge_when_pipeline_succeeds = true'; rv_q fail "queued auto-merge"; r1="$RC"
  rv_case || return 0; rv_receipt; vu_json "$C/mr.json" '.squash = true'; rv_q fail "squashes"
  detail "queued auto-merge exit $r1; squash exit $RC"; }
ctl_R3p() { rv_case || return 0; rv_receipt
  vu_json "$C/pipeline.json" '.ref = "0.1.0-rc.1" | .tag = true | .source = "push"'; rv_q fail "a tag pipeline is not a promotion qualification"
  detail "exit $RC: tag-shaped proof (ref = tag, tag = true, source push) is refused"; }
# Codex preliminary review finding 2: the latest pipeline must be freshly successful (same ID/SHA can be re-running).
ctl_R3q() { local r1; rv_case || return 0; rv_receipt
  vu_json "$C/mr.json" '.head_pipeline.status = "running"'; rv_q fail "latest pipeline is running, not success"; r1="$RC"
  rv_case || return 0; rv_receipt; vu_json "$C/mr.json" '.head_pipeline.status = "failed"'; rv_q fail "latest pipeline is failed, not success"
  detail "same pipeline ID and SHA but running (a retry) exit $r1, failed exit $RC"; }

ctl_R4a() { rv_case || return 0; rv_receipt "${RV_QREC[@]}"
  rv_pass promotable 0.1.0-rc.1 "$C/receipt" "$RV_MAIN"
  if grep -qx "EXPECTED_MAIN=$RV_MAIN" "$C/receipt"; then :; else prob "the expected main is not recorded"; fi
  detail "main R0 behind the RC: conflict-free, prospective tree = RC tree; expected main recorded"; }
ctl_R4b() { local m r1; rv_case || return 0; rv_receipt
  rv_fail "lacks PIPELINE_ID" promotable 0.1.0-rc.1 "$C/receipt" "$RV_MAIN"; r1="$RC"
  rv_receipt "${RV_QREC[@]}"; m="$(rv_commit "$RV_MAIN" "$RV_MAIN" "main-only.txt=main change")"
  rv_fail "differs from the RC tree" promotable 0.1.0-rc.1 "$C/receipt" "$m"
  detail "before qualification exit $r1; main changed a file the RC lacks exit $RC"; }
ctl_R4c() { local m; rv_case || return 0; rv_receipt "${RV_QREC[@]}"
  m="$(rv_commit "$RV_MAIN" "$RV_MAIN" "docs/release-note.md=conflicting note")"
  rv_fail "conflicts" promotable 0.1.0-rc.1 "$C/receipt" "$m"
  detail "exit $RC: main conflicts with the RC"; }

ctl_R5a() { rv_case || return 0; rv_receipt "${RV_QREC[@]}" "EXPECTED_MAIN=$RV_MAIN"
  rv_pass recheck 0.1.0-rc.1 "$C/receipt" "$C/ls.txt" "$C/mr.json" "$C/jobs.json"
  if [[ "$RV_LINE" != *"not compared atomically"* ]]; then prob "the PASS line does not state the non-atomic target limitation"; fi
  detail "fresh listing: tag object, peel, release/0.1.0-rc.1 = peel, main = expected; MR head and latest pipeline unchanged; non-atomic target stated"; }
ctl_R5b() { local p0; rv_case || return 0; rv_receipt "${RV_QREC[@]}" "EXPECTED_MAIN=$RV_MAIN"
  p0="$(rv_merge "$RV_TREE" "$RV_MAIN" "$RV_PEEL")"
  git -C "$C/repo" push -q "$C/remote.git" "$p0:refs/heads/main"; git ls-remote "$C/remote.git" >"$C/ls.txt"
  rv_fail "main moved" recheck 0.1.0-rc.1 "$C/receipt" "$C/ls.txt" "$C/mr.json" "$C/jobs.json"
  rv_pass promotable 0.1.0-rc.1 "$C/receipt" "$p0"
  rv_pass recheck 0.1.0-rc.1 "$C/receipt" "$C/ls.txt" "$C/mr.json" "$C/jobs.json"
  detail "main moved (an earlier promotion merge with an equal tree): recheck refused, promotable on the new main passed, recheck passed"; }
ctl_R5c() { local r1; rv_case || return 0; rv_receipt "${RV_QREC[@]}" "EXPECTED_MAIN=$RV_MAIN"
  git -C "$C/remote.git" update-ref refs/heads/release/0.1.0-rc.1 "$RV_MAIN"; git ls-remote "$C/remote.git" >"$C/ls.txt"
  rv_fail "is not the RC peeled commit" recheck 0.1.0-rc.1 "$C/receipt" "$C/ls.txt" "$C/mr.json" "$C/jobs.json"; r1="$RC"
  git -C "$C/remote.git" update-ref -d refs/heads/release/0.1.0-rc.1; git ls-remote "$C/remote.git" >"$C/ls.txt"
  rv_fail "refs/heads/release/0.1.0-rc.1, want exactly 1" recheck 0.1.0-rc.1 "$C/receipt" "$C/ls.txt" "$C/mr.json" "$C/jobs.json"
  detail "release branch moved exit $r1; release branch missing from the fresh listing exit $RC"; }
ctl_R5d() { rv_case || return 0; rv_receipt "${RV_QREC[@]}" "EXPECTED_MAIN=$RV_MAIN"
  git -C "$C/remote.git" tag -d 0.1.0-rc.1 >/dev/null
  GIT_COMMITTER_DATE=2026-01-05T00:00:00Z git -C "$C/remote.git" "${RV_ID[@]}" tag -a -m recreated 0.1.0-rc.1 "$RV_PEEL"
  git ls-remote "$C/remote.git" >"$C/ls.txt"
  rv_fail "remote tag object changed" recheck 0.1.0-rc.1 "$C/receipt" "$C/ls.txt" "$C/mr.json" "$C/jobs.json"
  detail "exit $RC: the remote tag was recreated (same peel, another object)"; }
# Codex preliminary review finding 2 at the pre-merge recheck: fresh success and the qualified job identities.
ctl_R5e() { local r1; rv_case || return 0; rv_receipt "${RV_QREC[@]}" "EXPECTED_MAIN=$RV_MAIN"
  vu_json "$C/mr.json" '.head_pipeline.status = "running"'
  rv_fail "latest pipeline is running, not success" recheck 0.1.0-rc.1 "$C/receipt" "$C/ls.txt" "$C/mr.json" "$C/jobs.json"; r1="$RC"
  rv_case || return 0; rv_receipt "${RV_QREC[@]}" "EXPECTED_MAIN=$RV_MAIN"
  vu_json "$C/jobs.json" '(.[] | select(.name == "validate:integration") | .id) = 4599'
  rv_fail "re-run qualified" recheck 0.1.0-rc.1 "$C/receipt" "$C/ls.txt" "$C/mr.json" "$C/jobs.json"
  detail "same pipeline re-running at the pre-merge recheck exit $r1; a retried gate job (new identity) without a refreshed qualification exit $RC"; }
ctl_R5f() { rv_case || return 0; rv_receipt
  rv_q
  vu_json "$C/jobs.json" '(.[] | select(.name == "validate:integration") | .id) = 4599'
  vu_json "$C/gate/summary.json" '.ci.job_id = 4599'
  rv_q
  if grep -qx 'GATE_JOB_ID=4599' "$C/receipt"; then :; else prob "the refreshed qualification did not record the retried job"; fi
  rv_pass promotable 0.1.0-rc.1 "$C/receipt" "$RV_MAIN"
  rv_pass recheck 0.1.0-rc.1 "$C/receipt" "$C/ls.txt" "$C/mr.json" "$C/jobs.json"
  detail "a successful retried gate job with its refreshed artifact: qualified re-run records it, then promotable and recheck accept"; }

rv_publish() { # COMMIT: the bare remote's main becomes COMMIT (a fast-forward), and the listing is refreshed
  git -C "$C/repo" push -q "$C/remote.git" "$1:refs/heads/main"
  git ls-remote "$C/remote.git" >"$C/ls.txt"
}
ctl_R6a() { local mg; rv_case || return 0; rv_receipt "${RV_QREC[@]}" "EXPECTED_MAIN=$RV_MAIN"
  mg="$(rv_merge "$RV_TREE" "$RV_MAIN" "$RV_PEEL")"; rv_publish "$mg"
  rv_pass promoted 0.1.0-rc.1 "$C/receipt" "$mg" "$C/ls.txt"
  if grep -qx "MAIN_MERGE_SHA=$mg" "$C/receipt"; then :; else prob "the main merge is not recorded"; fi
  detail "published merge: parents (expected main, RC), tree = RC tree, on the listed remote main; recorded"; }
# Codex preliminary review finding 3: correct local parents/tree do not prove publication.
ctl_R6d() { local mg; rv_case || return 0; rv_receipt "${RV_QREC[@]}" "EXPECTED_MAIN=$RV_MAIN"
  mg="$(rv_merge "$RV_TREE" "$RV_MAIN" "$RV_PEEL")"
  rv_fail "is not on the remote main" promoted 0.1.0-rc.1 "$C/receipt" "$mg" "$C/ls.txt"
  if grep -q '^MAIN_MERGE_SHA=' "$C/receipt"; then prob "an unpublished merge was recorded"; fi
  detail "exit $RC: a correctly shaped merge that exists only locally (remote main unchanged) is refused and not recorded"; }
ctl_R6e() { local mg nx; rv_case || return 0; rv_receipt "${RV_QREC[@]}" "EXPECTED_MAIN=$RV_MAIN"
  mg="$(rv_merge "$RV_TREE" "$RV_MAIN" "$RV_PEEL")"; nx="$(rv_commit "$mg" "$mg" "later.txt=a later main change")"; rv_publish "$nx"
  rv_pass promoted 0.1.0-rc.1 "$C/receipt" "$mg" "$C/ls.txt"
  detail "a genuinely promoted merge after main advanced past it (remote main contains it): accepted"; }
ctl_R6b() { local mg r1; rv_case || return 0; rv_receipt "${RV_QREC[@]}" "EXPECTED_MAIN=$RV_MAIN"
  mg="$(rv_merge "$RV_TREE" "$RV_PEEL" "$RV_MAIN")"
  rv_fail "parents are not exactly" promoted 0.1.0-rc.1 "$C/receipt" "$mg" "$C/ls.txt"; r1="$RC"
  rv_fail "parents are not exactly" promoted 0.1.0-rc.1 "$C/receipt" "$RV_PEEL" "$C/ls.txt"
  detail "reversed parents exit $r1; a fast-forward to the RC (no merge commit) exit $RC"; }
ctl_R6c() { local mg; rv_case || return 0; rv_receipt "${RV_QREC[@]}" "EXPECTED_MAIN=$RV_MAIN"
  mg="$(rv_merge "$(git -C "$C/repo" rev-parse "$RV_MAIN^{tree}")" "$RV_MAIN" "$RV_PEEL")"
  rv_fail "tree differs from the RC tree" promoted 0.1.0-rc.1 "$C/receipt" "$mg" "$C/ls.txt"
  detail "exit $RC: the main merge does not carry the RC tree"; }

rv_final_case() { # MG: a promoted receipt whose main merge is a correct merge commit (printed)
  local mg
  rv_case || return 1
  mg="$(rv_merge "$RV_TREE" "$RV_MAIN" "$RV_PEEL")"
  rv_receipt "${RV_QREC[@]}" "EXPECTED_MAIN=$RV_MAIN" "MAIN_MERGE_SHA=$mg"
  RV_MG="$mg"
  if [[ "${1:-}" != unpublished ]]; then rv_publish "$mg"; fi
}
RV_MG=""
ctl_R7a() { rv_final_case || return 0
  rv_pass final pre 0.1.0-rc.1 0.1.0 "$C/receipt" "$C/ls.txt"
  GIT_COMMITTER_DATE=2026-01-06T00:00:00Z git -C "$C/repo" "${RV_ID[@]}" tag -a -m "release 0.1.0" 0.1.0 "$RV_MG"
  git -C "$C/repo" push -q "$C/remote.git" refs/tags/0.1.0; git ls-remote "$C/remote.git" >"$C/ls.txt"
  rv_pass final post 0.1.0-rc.1 0.1.0 "$C/receipt" "$C/ls.txt"
  if grep -qx "FINAL_TAG_OBJECT=$(git -C "$C/repo" rev-parse refs/tags/0.1.0)" "$C/receipt"; then :; else prob "the final tag object is not recorded"; fi
  detail "pre-creation absence recorded; after creation one annotated 0.1.0 on the verified main merge with the RC tree; recorded"; }
ctl_R7b() { local v out=""; rv_final_case || return 0
  for v in 0.2.0 v0.1.0 0.1.0-rc.1 0.1; do rv_fail "exactly the RC's MAJOR.MINOR.PATCH" final pre 0.1.0-rc.1 "$v" "$C/receipt" "$C/ls.txt"; out+=" $v:$RC"; done
  detail "wrong final versions refused:$out"; }
ctl_R7c() { local r1 m; rv_final_case || return 0
  rv_pass final pre 0.1.0-rc.1 0.1.0 "$C/receipt" "$C/ls.txt"
  GIT_COMMITTER_DATE=2026-01-06T00:00:00Z git -C "$C/repo" "${RV_ID[@]}" tag -a -m "release 0.1.0" 0.1.0 "$RV_PEEL"
  git -C "$C/repo" push -q "$C/remote.git" refs/tags/0.1.0; git ls-remote "$C/remote.git" >"$C/ls.txt"
  rv_fail "does not point at the verified main merge" final post 0.1.0-rc.1 0.1.0 "$C/receipt" "$C/ls.txt"; r1="$RC"
  rv_case || return 0
  m="$(rv_merge "$(git -C "$C/repo" rev-parse "$RV_MAIN^{tree}")" "$RV_MAIN" "$RV_PEEL")"
  rv_receipt "${RV_QREC[@]}" "EXPECTED_MAIN=$RV_MAIN" "MAIN_MERGE_SHA=$m"
  rv_fail "does not carry the RC tree" final pre 0.1.0-rc.1 0.1.0 "$C/receipt" "$C/ls.txt"
  detail "final tag on the RC commit instead of the main merge exit $r1; a recorded main merge without the RC tree exit $RC"; }
ctl_R7d() { local r1; rv_final_case || return 0
  rv_fail "lacks FINAL_TAG" final post 0.1.0-rc.1 0.1.0 "$C/receipt" "$C/ls.txt"; r1="$RC"
  GIT_COMMITTER_DATE=2026-01-06T00:00:00Z git -C "$C/repo" "${RV_ID[@]}" tag -a -m "release 0.1.0" 0.1.0 "$RV_MG"
  git -C "$C/repo" push -q "$C/remote.git" refs/tags/0.1.0; git ls-remote "$C/remote.git" >"$C/ls.txt"
  rv_fail "must be absent" final pre 0.1.0-rc.1 0.1.0 "$C/receipt" "$C/ls.txt"
  detail "post-creation verification before the absence proof exit $r1; absence proof with the final tag already published exit $RC"; }
# Codex preliminary review finding 3 (final): the verified merge must be on the freshly listed remote main, pre and post.
ctl_R7e() { local r1; rv_final_case unpublished || return 0
  rv_fail "is not on the remote main" final pre 0.1.0-rc.1 0.1.0 "$C/receipt" "$C/ls.txt"; r1="$RC"
  rv_publish "$RV_MG"
  rv_pass final pre 0.1.0-rc.1 0.1.0 "$C/receipt" "$C/ls.txt"
  GIT_COMMITTER_DATE=2026-01-06T00:00:00Z git -C "$C/repo" "${RV_ID[@]}" tag -a -m "release 0.1.0" 0.1.0 "$RV_MG"
  git -C "$C/repo" push -q "$C/remote.git" refs/tags/0.1.0
  git -C "$C/remote.git" update-ref refs/heads/main "$RV_MAIN"; git ls-remote "$C/remote.git" >"$C/ls.txt"
  rv_fail "is not on the remote main" final post 0.1.0-rc.1 0.1.0 "$C/receipt" "$C/ls.txt"
  detail "an unpublished merge: final pre exit $r1; remote main replaced after the absence proof: final post exit $RC"; }
ctl_R7f() { local nx; rv_final_case || return 0
  nx="$(rv_commit "$RV_MG" "$RV_MG" "later.txt=a later main change")"; rv_publish "$nx"
  rv_pass final pre 0.1.0-rc.1 0.1.0 "$C/receipt" "$C/ls.txt"
  GIT_COMMITTER_DATE=2026-01-06T00:00:00Z git -C "$C/repo" "${RV_ID[@]}" tag -a -m "release 0.1.0" 0.1.0 "$RV_MG"
  git -C "$C/repo" push -q "$C/remote.git" refs/tags/0.1.0; git ls-remote "$C/remote.git" >"$C/ls.txt"
  rv_pass final post 0.1.0-rc.1 0.1.0 "$C/receipt" "$C/ls.txt"
  detail "main advanced past the published merge: final pre and post accept, the tag on the verified merge (not the newer main tip)"; }

# merged (evidence only): an already merged promotion MR verified on a candidate receipt. The fixture: main merge
# RV_MG published on the bare remote's main, release/0.1.0-rc.1 deleted, mr.json merged as RV_MG (queue flag true).
# Every refusal must leave the receipt byte for byte unchanged (rv_m checks it).
rv_merged_setup() { # MG [unpublished]
  RV_MG="$1"
  rv_receipt
  if [[ "${2:-}" != unpublished ]]; then git -C "$C/repo" push -q "$C/remote.git" "$1:refs/heads/main"; fi
  git -C "$C/remote.git" update-ref -d refs/heads/release/0.1.0-rc.1
  git ls-remote "$C/remote.git" >"$C/ls.txt"
  vu_json "$C/mr.json" ".state = \"merged\" | .merge_when_pipeline_succeeds = true | .merge_commit_sha = \"$1\""
}
rv_m() { # [fail REASON]: merged over the case evidence with RV_MG
  local before
  before="$(sha256sum <"$C/receipt")"
  if [[ "${1:-}" == fail ]]; then
    rv_fail "$2" merged 0.1.0-rc.1 "$C/receipt" 999999999 "$C/mr.json" "$C/pipeline.json" "$C/jobs.json" "$C/fast" "$C/gate" "$RV_MG" "$C/ls.txt"
    if [[ "$(sha256sum <"$C/receipt")" != "$before" ]]; then prob "the refused merged evidence changed the receipt"; fi
  else rv_pass merged 0.1.0-rc.1 "$C/receipt" 999999999 "$C/mr.json" "$C/pipeline.json" "$C/jobs.json" "$C/fast" "$C/gate" "$RV_MG" "$C/ls.txt"; fi
}
rv_mcase() { rv_case || return 1; rv_merged_setup "$(rv_merge "$RV_TREE" "$RV_MAIN" "$RV_PEEL")"; } # the genuine merged fixture
ctl_R8a() { local want; rv_mcase || return 0
  rv_m
  want="$(printf 'FORMAT=release-receipt/1\nTAG=0.1.0-rc.1\nCUT_STAGING=%s\nCUT_LSREMOTE_SHA256=%s\nTAG_OBJECT=%s\nPEELED=%s\nTREE=%s\n' \
    "$RV_PEEL" "$(printf 'c%.0s' {1..64})" "$RV_OBJ" "$RV_PEEL" "$RV_TREE"
    printf 'PROJECT_ID=999999999\nMR_IID=7\nPIPELINE_ID=1001\nFAST_JOB_ID=4504\nGATE_JOB_ID=4505\nEXPECTED_MAIN=%s\nMAIN_MERGE_SHA=%s\nVERIFICATION_MODE=merged-evidence\n' "$RV_MAIN" "$RV_MG")"
  if [[ "$(cat "$C/receipt")" != "$want" ]]; then prob "the merged receipt is not exactly the candidate record plus the verified proof"; fi
  if [[ "$RV_LINE" != *"no merge mode, actor or authority inferred"*"fast 32 + remainder 97 -> 125 required IDs incl. ci-controls"* ]]; then prob "the PASS line lacks the evidence-only statement or the recomputed coverage: [$RV_LINE]"; fi
  rv_pass final pre 0.1.0-rc.1 0.1.0 "$C/receipt" "$C/ls.txt"
  if git -C "$C/repo" rev-parse -q --verify refs/tags/0.1.0 >/dev/null || [[ -n "$(git -C "$C/remote.git" tag -l 0.1.0)" ]]; then prob "a final tag exists"; fi
  detail "merged MR (queue flag true) with the full qualification, parents (main, RC), RC tree recreated from P1, on remote main, tag unchanged, release branch deleted: recorded once with VERIFICATION_MODE=merged-evidence (no actor or mode); final pre then passes (no tag created)"; }
ctl_R8b() { local nx; rv_mcase || return 0
  vu_json "$C/mr.json" '.merge_when_pipeline_succeeds = false'
  nx="$(rv_commit "$RV_MG" "$RV_MG" "later.txt=a later main change")"; rv_publish "$nx"
  rv_m
  detail "queue flag false and main advanced past the merge: accepted (the flag is not evidence either way)"; }
ctl_R8c() { local r1 r2; rv_mcase || return 0
  vu_json "$C/mr.json" ".sha = \"$RV_MAIN\""; rv_m fail "merge request head is not the RC peeled commit"; r1="$RC"
  rv_mcase || return 0; rv_retag 0.1.0-rc.1 "$RV_PEEL"; rv_m fail "the local tag differs from the receipt"; r2="$RC"
  rv_mcase || return 0
  rv_fail "belongs to another tag" merged 0.1.0-rc.2 "$C/receipt" 999999999 "$C/mr.json" "$C/pipeline.json" "$C/jobs.json" "$C/fast" "$C/gate" "$RV_MG" "$C/ls.txt"
  detail "another candidate (MR head) exit $r1; a recreated tag object exit $r2; another tag's receipt exit $RC"; }
ctl_R8d() { local r1 r2; rv_mcase || return 0
  vu_json "$C/pipeline.json" '.status = "failed"'; rv_m fail "did not succeed"; r1="$RC"
  rv_mcase || return 0; vu_json "$C/mr.json" '.head_pipeline.id = 1002'; rv_m fail "not the merge request's latest pipeline"; r2="$RC"
  rv_mcase || return 0; vu_json "$C/jobs.json" '(.[] | select(.name == "validate:integration") | .id) = 4599'; rv_m fail "gate: evidence of another job"
  detail "failed pipeline exit $r1; not the latest pipeline exit $r2; a retried gate job without its artifact exit $RC"; }
ctl_R8e() { local r1; rv_mcase || return 0
  vu_json "$C/gate/summary.json" '.ci.pipeline_id = 1002'; rv_m fail "gate: evidence of another pipeline"; r1="$RC"
  rv_mcase || return 0; printf '# not the candidate registry\n' >>"$C/fast/required-checks.tsv"; printf '# not the candidate registry\n' >>"$C/gate/required-checks.tsv"
  rv_m fail "a published registry is not the candidate's registry"
  detail "a gate artifact of another pipeline exit $r1; both published registries differ from the candidate's exit $RC"; }
ctl_R8f() { local x r1; rv_case || return 0
  rv_merged_setup "$(rv_merge "$RV_TREE" "$RV_PEEL" "$RV_MAIN")"; rv_m fail "parents are not exactly"; r1="$RC"
  rv_case || return 0; x="$(rv_commit "$RV_MAIN" "$RV_MAIN" "extra.txt=a third parent")"
  rv_merged_setup "$(GIT_AUTHOR_DATE=2026-01-04T00:00:00Z GIT_COMMITTER_DATE=2026-01-04T00:00:00Z \
    git -C "$C/repo" "${RV_ID[@]}" commit-tree "$RV_TREE" -p "$RV_MAIN" -p "$RV_PEEL" -p "$x" -m "octopus")"
  rv_m fail "parents are not exactly"
  detail "swapped parents exit $r1; three parents exit $RC"; }
ctl_R8g() { rv_case || return 0
  rv_merged_setup "$(rv_merge "$RV_TREE" "$RV_MAIN" "$RV_PEEL")" unpublished; rv_m fail "is not on the remote main"
  detail "exit $RC: after the qualification passed, a merge that is not on the listed remote main is refused; receipt unchanged"; }
ctl_R8h() { rv_case || return 0
  rv_merged_setup "$(rv_merge "$(git -C "$C/repo" rev-parse "$RV_MAIN^{tree}")" "$RV_MAIN" "$RV_PEEL")"; rv_m fail "tree differs from the RC tree"
  detail "exit $RC: the merge does not carry the RC tree"; }
ctl_R8i() { local p r1; rv_case || return 0
  p="$(rv_commit "$RV_MAIN" "$RV_MAIN" "docs/release-note.md=conflicting note")"
  rv_merged_setup "$(rv_merge "$RV_TREE" "$p" "$RV_PEEL")"; rv_m fail "conflicts"; r1="$RC"
  rv_case || return 0; p="$(rv_commit "$RV_MAIN" "$RV_MAIN" "main-only.txt=main change")"
  rv_merged_setup "$(rv_merge "$RV_TREE" "$p" "$RV_PEEL")"; rv_m fail "recreates tree"
  detail "P1 conflicting with the RC exit $r1; P1 whose merge with the RC is not the RC tree (main content dropped) exit $RC"; }
ctl_R8j() { local r1; rv_mcase || return 0
  git -C "$C/remote.git" tag -d 0.1.0-rc.1 >/dev/null
  GIT_COMMITTER_DATE=2026-01-05T00:00:00Z git -C "$C/remote.git" "${RV_ID[@]}" tag -a -m recreated 0.1.0-rc.1 "$RV_PEEL"
  git ls-remote "$C/remote.git" >"$C/ls.txt"
  rv_m fail "remote tag object changed"; r1="$RC"
  rv_mcase || return 0; git -C "$C/remote.git" branch release/0.1.0-rc.1 "$RV_PEEL"; git ls-remote "$C/remote.git" >"$C/ls.txt"
  rv_m fail "is still on the remote"
  detail "remote tag recreated after the merge exit $r1; release branch still live exit $RC (both after the qualification passed; receipt unchanged)"; }
ctl_R8k() { local r1 r2; rv_mcase || return 0
  rv_cut_receipt; rv_m fail "lacks TAG_OBJECT"; r1="$RC"
  rv_receipt PIPELINE_ID=1001; rv_m fail "already holds PIPELINE_ID"; r2="$RC"
  rv_receipt "MAIN_MERGE_SHA=$RV_MG"; rv_m fail "already holds MAIN_MERGE_SHA"
  detail "partial (cut only) receipt exit $r1; pre-qualified receipt with PIPELINE_ID exit $r2, with MAIN_MERGE_SHA exit $RC"; }
ctl_R8l() { local r1 r2; rv_mcase || return 0
  vu_json "$C/mr.json" '.state = "opened"'; rv_m fail "the merge request is not merged"; r1="$RC"
  rv_mcase || return 0; vu_json "$C/mr.json" ".merge_commit_sha = \"$RV_MAIN\""; rv_m fail "merge commit is not MERGE_SHA"; r2="$RC"
  rv_mcase || return 0; rv_q fail "the merge request is not open"
  detail "an open MR exit $r1; another merge commit exit $r2; manual qualified still refuses the merged MR exit $RC"; }
# MR58 Duo finding: after merged, no manual step may extend (overwrite EXPECTED_MAIN or MAIN_MERGE_SHA in) the receipt.
ctl_R8m() { local before p0 mg2 out=""; rv_mcase || return 0
  rv_m
  before="$(sha256sum <"$C/receipt")"
  # Arguments the manual steps would otherwise accept: the published merge as the "new main" (conflict-free, RC tree)
  # and a correctly shaped (expected main, RC) merge; without the guard promotable would overwrite EXPECTED_MAIN.
  p0="$RV_MG"
  mg2="$(rv_merge "$RV_TREE" "$RV_MAIN" "$RV_PEEL")"
  rv_fail "manual and merged evidence never mix" promotable 0.1.0-rc.1 "$C/receipt" "$p0"; out+=" promotable:$RC"
  rv_fail "manual and merged evidence never mix" recheck 0.1.0-rc.1 "$C/receipt" "$C/ls.txt" "$C/mr.json" "$C/jobs.json"; out+=" recheck:$RC"
  rv_fail "manual and merged evidence never mix" promoted 0.1.0-rc.1 "$C/receipt" "$mg2" "$C/ls.txt"; out+=" promoted:$RC"
  rv_q fail "manual and merged evidence never mix"; out+=" qualified:$RC"
  if [[ "$(sha256sum <"$C/receipt")" != "$before" ]]; then prob "a manual step changed the merged-evidence receipt"; fi
  rv_pass final pre 0.1.0-rc.1 0.1.0 "$C/receipt" "$C/ls.txt"
  GIT_COMMITTER_DATE=2026-01-06T00:00:00Z git -C "$C/repo" "${RV_ID[@]}" tag -a -m "release 0.1.0" 0.1.0 "$RV_MG"
  git -C "$C/repo" push -q "$C/remote.git" refs/tags/0.1.0; git ls-remote "$C/remote.git" >"$C/ls.txt"
  rv_pass final post 0.1.0-rc.1 0.1.0 "$C/receipt" "$C/ls.txt"
  detail "after a successful merged, each manual step (otherwise well-formed arguments) is refused with the receipt unchanged:$out; final pre and post (fixture tag in the scratch remote) still pass"; }

# --- runner -------------------------------------------------------------------------------------------
mk_template
need_b=0 need_f=0 need_g=0
for c in "${SELECTED[@]}"; do case "${c%% *}" in B) need_b=1 ;; F) need_f=1 ;; G) need_g=1 ;; esac; done
# G051: the two F schedule controls check out the fixture origin too; prepare it once in this shell so a focused
# --group F run does not build it twice in their subshells.
for c in "${SELECTED[@]}"; do case "${c#* }" in F-schedule-qualification|F-schedule-other) need_b=1 ;; esac; done
if (( need_b )); then mk_origin; mk_rc_origin; fi
if (( need_f )); then mk_fbase; fi
if (( need_g )) && [[ -n "$CLI_IMAGE" ]]; then
  if D image inspect "$CLI_IMAGE" >/dev/null 2>&1; then :; else D pull "$CLI_IMAGE" >/dev/null 2>&1 || true; fi
fi
# --- G004 CI-1: the identity-staging forwarding gate (Plan r5 §8.5, C-SEC-1..10) --------------------------------------
# The approved project path is the template's synthetic example project (re-pin it with infra/ci/hosted-context.sh), so no
# identity literal is written here, and the path's SHA-256 must equal the adapter's pin.
G_STG_KEY="sk_live_$(ci_random_hex 16)"
g_stg_path() {
  printf 'example-group/example-project/platform'
}
g_stg_env() { # [NAME=VALUE | -NAME]...: the approved identity-staging job environment, with overrides and removals
  local o args=() drops=()
  local -A v=([CI_PROJECT_ID]=999999999 [CI_PROJECT_PATH]="$(g_stg_path)" [CI_COMMIT_BRANCH]=identity-qualification
    [CI_COMMIT_REF_NAME]=identity-qualification [CI_COMMIT_REF_PROTECTED]=true [CI_PIPELINE_SOURCE]=web [CI_JOB_NAME]=identity-staging
    [CI_ENVIRONMENT_NAME]=identity-staging [IDENTITY_STAGING_WORKOS_API_KEY]="$G_STG_KEY"
    [IDENTITY_STAGING_WORKOS_CLIENT_ID]=client_01CONTROLSFAKE
    [IDENTITY_STAGING_WORKOS_ISSUER]=https://api.workos.com/user_management/client_01CONTROLSENVFAKE)
  for o in "$@"; do
    if [[ "$o" == -* ]]; then unset "v[${o#-}]"; drops+=("$o"); else v[${o%%=*}]="${o#*=}"; fi
  done
  for o in "${!v[@]}"; do args+=("$o=${v[$o]}"); done
  g_env "${drops[@]}" "${args[@]}"
}
g_stg_none() { # the adapter refused before any forwarding: exit 16, no context, no staging mount, the key nowhere
  g_rc 16; g_no_context
  if grep -q 'platform-identity-staging.env' "$C/gs/docker.log" 2>/dev/null; then prob "a staging mount was passed"; fi
  if grep -qF 'identity staging: approved context' "$C/out" 2>/dev/null; then prob "the adapter reported forwarding"; fi
  g_stg_secret_free
}
g_stg_secret_free() {
  local f
  for f in "$C/out" "$C/gs"/*; do
    if [[ -f "$f" ]] && grep -qF -e "$G_STG_KEY" -- "$f"; then prob "the staging key reached ${f##*/}"; fi
  done
  g_secret_free
}
g_stg_refused() { # MESSAGE ENV...: one refused case
  local msg="$1"; shift
  g_case; g_stg_env "$@"; g_run -- --profile staging --publish-dir ci-evidence; g_stg_none; g_has "$msg"
}
STG_OUTSIDE="refused: the staging profile outside the approved identity-staging context (nothing forwarded)"
ctl_C_SEC_1() {
  g_stg_refused "$STG_OUTSIDE" CI_PIPELINE_SOURCE=merge_request_event CI_MERGE_REQUEST_IID=12 -CI_COMMIT_BRANCH
  local r1="$RC"
  g_stg_refused "$STG_OUTSIDE" CI_MERGE_REQUEST_IID=12
  detail "a merge request pipeline exit $r1; a web pipeline carrying a merge request IID exit $RC; nothing forwarded"
}
ctl_C_SEC_2() {
  g_stg_refused "$STG_OUTSIDE" CI_COMMIT_REF_PROTECTED=false; local r1="$RC"
  g_stg_refused "$STG_OUTSIDE" -CI_COMMIT_REF_PROTECTED; local r2="$RC"
  g_stg_refused "$STG_OUTSIDE" CI_COMMIT_BRANCH=main CI_COMMIT_REF_NAME=main
  detail "unprotected exit $r1; protection unknown exit $r2; another protected branch exit $RC; nothing forwarded"
}
ctl_C_SEC_3() {
  g_stg_refused "$STG_OUTSIDE" CI_PIPELINE_SOURCE=push; local r1="$RC"
  g_stg_refused "$STG_OUTSIDE" CI_PIPELINE_SOURCE=api
  detail "push exit $r1; api exit $RC; nothing forwarded"
}
ctl_C_SEC_4() {
  g_stg_refused "$STG_OUTSIDE" CI_JOB_NAME=gate; local r1="$RC"
  g_stg_refused "$STG_OUTSIDE" -CI_JOB_NAME
  detail "job gate exit $r1; no job name exit $RC; nothing forwarded"
}
ctl_C_SEC_5() {
  g_stg_refused "$STG_OUTSIDE" -CI_ENVIRONMENT_NAME; local r1="$RC"
  g_stg_refused "$STG_OUTSIDE" CI_ENVIRONMENT_NAME=identity-staging-x; local r2="$RC"
  g_stg_refused "$STG_OUTSIDE" CI_ENVIRONMENT_NAME=production
  detail "no environment exit $r1; identity-staging-x exit $r2; production exit $RC; nothing forwarded"
}
ctl_C_SEC_6() {
  g_stg_refused "$STG_OUTSIDE" CI_JOB_NAME=fast -CI_ENVIRONMENT_NAME -IDENTITY_STAGING_WORKOS_API_KEY -IDENTITY_STAGING_WORKOS_CLIENT_ID -IDENTITY_STAGING_WORKOS_ISSUER
  local r1="$RC"
  g_stg_refused "$STG_OUTSIDE" CI_JOB_NAME=gate -CI_ENVIRONMENT_NAME
  detail "--profile staging from the fast job context exit $r1; from the gate job context (variables visible) exit $RC; refused"
}
ctl_C_SEC_7() {
  local p r=""
  for p in fast gate; do
    g_case; g_env "IDENTITY_STAGING_WORKOS_API_KEY=$G_STG_KEY" CI_JOB_NAME="$p"; g_run -- --profile "$p" --publish-dir ci-evidence
    g_stg_none; g_has "refused: an IDENTITY_STAGING_WORKOS_* variable is visible to the $p profile (scope misconfiguration; nothing forwarded)"
    r+=" $p=$RC"
  done
  g_case; g_env IDENTITY_STAGING_WORKOS_CLIENT_ID=client_01CONTROLSFAKE; g_run -- --profile fast --publish-dir ci-evidence; g_stg_none
  detail "a staging variable visible to the job:$r; metadata only, fast=$RC; fail closed, no context"
}
ctl_C_SEC_8() {
  local line want
  want="$(sed -n 's/^STG_PROJECT_PATH_SHA=\([0-9a-f]\{64\}\)$/\1/p' "$REPO_R/infra/ci/hosted-context.sh")"
  if [[ "$(printf '%s' "$(g_stg_path)" | sha256sum | cut -c1-64)" != "$want" ]]; then prob "the derived project path does not match the adapter's pin"; fi
  g_case staging-inspect; g_stg_env; g_run -- --profile staging --publish-dir ci-evidence; g_rc 0; g_context_ran
  g_has "identity staging: approved context; 3 of 3 inputs forwarded as one read-only file (values not shown)"
  line="$(grep -E -- '^run .* --name platform-ci-777 ' "$C/gs/docker.log" 2>/dev/null | head -1)" || line=""
  if [[ ! "$line" =~ \ --mount\ type=volume,src=platform-ci-stg-777,dst=/run/platform-identity-staging\.env,volume-subpath=env,readonly\ -e\ IDENTITY_STAGING_ENV_FILE=/run/platform-identity-staging\.env\  ]]; then prob "the context lacks the read-only staging volume file and its path"; fi
  if [[ "$line" == *" -v /tmp/"* ]]; then prob "the context bind-mounts a job-side path (it would name the daemon's filesystem)"; fi
  if [[ "$line" == *"IDENTITY_STAGING_WORKOS"* ]]; then prob "a staging variable name reached the context arguments"; fi
  if [[ "$(cat "$C/gs/stg-mode" 2>/dev/null)" != 600 ]]; then prob "the staging file mode is $(cat "$C/gs/stg-mode" 2>/dev/null), want 600"; fi
  if [[ "$(cat "$C/gs/stg-keys" 2>/dev/null)" != "IDENTITY_STAGING_WORKOS_API_KEY IDENTITY_STAGING_WORKOS_CLIENT_ID IDENTITY_STAGING_WORKOS_ISSUER " ]]; then prob "the staging file holds [$(cat "$C/gs/stg-keys" 2>/dev/null)]"; fi
  if [[ "$(cat "$C/gs/stg-delivered-keys" 2>/dev/null)" != "IDENTITY_STAGING_WORKOS_API_KEY IDENTITY_STAGING_WORKOS_CLIENT_ID IDENTITY_STAGING_WORKOS_ISSUER " ]]; then prob "the delivery helper received [$(cat "$C/gs/stg-delivered-keys" 2>/dev/null)] on stdin"; fi
  local vc vd cx vr
  vc="$(grep -nxF 'volume create --label platform.ci.staging=777 platform-ci-stg-777' "$C/gs/docker.log" | head -1 | cut -d: -f1)"
  vd="$(grep -n -- '--name platform-ci-777-stg ' "$C/gs/docker.log" | head -1 | cut -d: -f1)"
  cx="$(grep -nE -- '^run .* --name platform-ci-777 ' "$C/gs/docker.log" | head -1 | cut -d: -f1)"
  vr="$(grep -nxF 'volume rm -f platform-ci-stg-777' "$C/gs/docker.log" | tail -1 | cut -d: -f1)"
  if [[ -n "$vc" && -n "$vd" && -n "$cx" && -n "$vr" ]] && (( vc < vd && vd < cx && cx < vr )); then :; else prob "the volume is not created, filled, mounted and removed in that order [$vc $vd $cx $vr]"; fi
  g_stg_secret_free; g_calls_owned
  # Delivery to the job daemon fails closed: no context, and the volume is still removed.
  local f r6=""
  for f in stg-vol-fail stg-fail; do
    g_case; g_stg_env; g_extra ": >/gs/$f"
    g_run -- --profile staging --publish-dir ci-evidence; g_rc 16; g_no_context
    g_has "refused: cannot deliver the identity staging file to the job daemon; no context started"
    g_call "volume rm -f platform-ci-stg-777"; g_stg_secret_free; g_calls_owned
    r6+=" $f=$RC"
  done
  g_stg_real
  # The API key joined the credential patterns: a copy below the build directory is refused.
  g_case; g_stg_env; g_extra 'mkdir -p "$NS/cache"; printf "k=%s\n" "$IDENTITY_STAGING_WORKOS_API_KEY" >"$NS/cache/k"'
  g_run -- --profile staging --publish-dir ci-evidence; g_rc 16; g_no_context; g_has "refused: credentials not proven absent; no context started"
  local r3="$RC"
  # The project ID and the path are each required (checkpoint C4 item 10): a wrong ID with the approved path, the
  # approved ID with a wrong or missing path; the pin is the adapter's literal, never a hash of the incoming path.
  g_stg_refused "$STG_OUTSIDE" CI_PROJECT_ID=999999998; local r4="$RC"
  g_stg_refused "$STG_OUTSIDE" CI_PROJECT_PATH="$(g_stg_path)-fork"; local r5="$RC"
  g_stg_refused "$STG_OUTSIDE" -CI_PROJECT_PATH
  detail "approved project, protected identity-qualification, web, job and environment identity-staging: exit 0, one mode-600 file of exactly the three names delivered over a helper's stdin into a job-scoped volume, mounted read-only (volume-subpath) and removed after, no job-side bind mount, no value in argv, output or state; delivery faults fail closed:$r6; real daemon: a regular mode-600 read-only file of exactly the three names; the key below the build directory refused ($r3); wrong project ID exit $r4, wrong path exit $r5, no path exit $RC, nothing forwarded"
}
# g_stg_real: the delivery mechanism on the real daemon, with the adapter's exact helper command and mount form. No
# daemon-host path is involved, so the result holds for the hosted job-private daemon reached over TCP.
g_stg_real() {
  local vol="g003ctl-stg-$RT" src="$C/stg-real.env" got
  grep -qF -- "sh -c 'umask 077 && cat >/d/env && test -s /d/env'" "$REPO_R/infra/ci/hosted-context.sh" &&
    grep -qF -- 'stg="--mount type=volume,src=$STG_VOL,dst=$STG_MOUNT,volume-subpath=env,readonly -e IDENTITY_STAGING_ENV_FILE=$STG_MOUNT"' "$REPO_R/infra/ci/hosted-context.sh" ||
    prob "the adapter's delivery helper or mount form differs from the one proven here"
  printf 'IDENTITY_STAGING_WORKOS_API_KEY=%s\nIDENTITY_STAGING_WORKOS_CLIENT_ID=client_01CONTROLSFAKE\nIDENTITY_STAGING_WORKOS_ISSUER=https://api.workos.com/user_management/client_01CONTROLSENVFAKE\n' "$G_STG_KEY" >"$src"; chmod 600 "$src"
  D volume rm -f "$vol" >/dev/null 2>&1 || true
  if D volume create "${LB[@]}" "$vol" >/dev/null 2>&1 &&
     D run --rm -i --network none --pull never "${LB[@]}" -v "$vol:/d" "$CLI_IMAGE" sh -c 'umask 077 && cat >/d/env && test -s /d/env' <"$src" >/dev/null 2>&1; then
    got="$(D run --rm --network none --pull never "${LB[@]}" --mount "type=volume,src=$vol,dst=/run/platform-identity-staging.env,volume-subpath=env,readonly" \
      "$CLI_IMAGE" sh -c 'f=/run/platform-identity-staging.env; if [ -f "$f" ] && [ ! -L "$f" ]; then printf "file "; fi
        printf "%s " "$(stat -c %a "$f")"; cut -d= -f1 "$f" | tr "\n" " "; if (echo x >>"$f") 2>/dev/null; then echo writable; else echo read-only; fi' 2>&1)" || got="run failed"
    [[ "$got" == "file 600 IDENTITY_STAGING_WORKOS_API_KEY IDENTITY_STAGING_WORKOS_CLIENT_ID IDENTITY_STAGING_WORKOS_ISSUER read-only" ]] || prob "real daemon: the mounted staging file is [$got]"
    if [[ "$got" == *"$G_STG_KEY"* ]]; then prob "real daemon: the key reached the probe output"; fi
  else prob "real daemon: the staging volume could not be created or filled"; fi
  D volume rm -f "$vol" >/dev/null 2>&1 || true
  if D volume inspect "$vol" >/dev/null 2>&1; then prob "real daemon: the staging volume was not removed"; fi
  rm -f "$src"
}
ctl_C_SEC_9() {
  local r1
  g_case staging-inspect; g_stg_env -IDENTITY_STAGING_WORKOS_ISSUER; g_run -- --profile staging --publish-dir ci-evidence; r1="$RC"; g_rc 0
  g_has "identity staging: approved context; 2 of 3 inputs forwarded as one read-only file (values not shown)"
  # The staging suite records a missing input as UNMET, never a skip (the real harness, staging suite only).
  local w="$C/stgrun" f="$C/stgrun-in.env" res
  mkdir -p "$w"; printf 'IDENTITY_STAGING_WORKOS_API_KEY=%s\nIDENTITY_STAGING_WORKOS_CLIENT_ID=client_01CONTROLSFAKE\n' "$G_STG_KEY" >"$f"; chmod 600 "$f"
  launch "$C/stgrun.out" env IDENTITY_STAGING_ENV_FILE="$f" TMPDIR="$w" bash "$REPO_R/infra/validation/run.sh" --base HEAD \
    --allow-file "$REPO_R/infra/ci/unchanged.allow" --report-dir "$w/report" staging
  await "$LPID" 600
  res="$(awk -F'\t' '$1 == "B-IDENTITY-STAGING" {print $2}' "$w/report/results.tsv" 2>/dev/null)"
  if [[ "$RC" == 0 || "$res" != UNMET ]]; then prob "the staging suite with two of three inputs: exit $RC, B-IDENTITY-STAGING ${res:-absent}, want nonzero and UNMET"; fi
  if grep -rqF -- "$G_STG_KEY" "$w/report" "$C/stgrun.out" 2>/dev/null; then prob "the staging key reached the harness output"; fi
  local r2="$RC"
  # validate.sh --profile staging (checkpoint C4 item 9): the 2700 s job floor holds first, then the hosted context
  # and the forwarded file; each refusal is at location, before preflight, the registry and the harness.
  local floor="location: the identity-staging job timeout is missing or below its 2700s floor" t r3=""
  for t in 2699 45m ""; do
    std_ci 4280; CIENV=("${CIENV[@]/#CI_JOB_TIMEOUT=*/CI_JOB_TIMEOUT=$t}")
    run_v "${CIENV[@]}" -- --profile staging --publish-dir ci-evidence
    expect_rc 16; expect_phase location 16; expect_runs 0; log_has "$floor"
    r3+=" [${t:-unset}]=$RC"
  done
  std_ci 4281; CIENV=("${CIENV[@]/#CI_JOB_TIMEOUT=*/CI_JOB_TIMEOUT=2700}")
  run_v "${CIENV[@]}" PLATFORM_CI_CONTEXT=hosted-context/1 DOCKER_HOST=tcp://localhost:2376 DOCKER_TLS_VERIFY=1 DOCKER_CERT_PATH=/certs/client \
    -- --profile staging --publish-dir ci-evidence
  expect_rc 16; expect_phase location 16; expect_runs 0
  if grep -qxF "validate: $floor" "$C/out"; then prob "the floor refused exactly 2700 s"; fi
  if grep -qxE 'validate: location: the staging profile (runs only in the hosted adapter context|has no forwarded identity staging file)' "$C/out"; then :; else prob "a staging run without the hosted context or the forwarded file was not refused at location"; fi
  detail "approved context with the issuer missing: adapter exit $r1 (2 of 3 forwarded); harness staging suite exit $r2 with B-IDENTITY-STAGING UNMET; validate.sh staging below the 2700 s floor:$r3 at location, no harness; at 2700 s without the forwarded file exit $RC at location"
}
ctl_C_SEC_10() {
  local r1
  # P-NEG-1 at CI level: the registered staging key in a would-be-published report is withheld.
  std_ci; run_v "${CIENV[@]}" STUB_HARNESS=plant-staging-key -- --publish-dir ci-evidence; r1="$RC"; expect_rc 15; expect_pub minimal:WITHHELD
  if [[ -s "$C/stub/staging-key" ]] && grep -rqF -f "$C/stub/staging-key" "$(PUBDIR)" 2>/dev/null; then prob "the planted staging key was published"; fi
  # P-POS-1 at CI level: the public client ID and issuer in sanitized provenance are published in full.
  std_ci; run_v "${CIENV[@]}" STUB_HARNESS=staging-metadata -- --publish-dir ci-evidence; expect_rc 0; expect_pub full
  if grep -qF 'client_01CONTROLSFAKE' "$(PUBDIR)/plan.out" 2>/dev/null; then :; else prob "the public client ID was not published"; fi
  detail "planted registered staging key: exit $r1, withheld; public client ID and issuer only: exit $RC, published in full"
}

# A control is its function, or a row of the selector table.
ctl_run() { if declare -F "ctl_${1//-/_}" >/dev/null; then "ctl_${1//-/_}"; else f_row "$1"; fi; }

NOTRUN=()
for c in "${SELECTED[@]}"; do
  CTL_ID="${c#* }"
  # G064 correction 11: after an unverified cleanup (HELD) nothing more is allocated: no later control starts.
  if [[ -n "$HELD" ]]; then NOTRUN+=("$CTL_ID"); continue; fi
  held=0
  PROB="$ROOT/problems/$CTL_ID"; DET="$ROOT/details/$CTL_ID"
  : >"$PROB"; : >"$DET"
  # G064 corr1 P6: every control keeps an inventory, started before it can allocate anything.
  : >"$ROOT/owned/$CTL_ID.tsv"
  # The control runs in the background and is waited for, so a signal reaches this shell's trap at once.
  CUR_CTL="$CTL_ID"
  t0="$SECONDS"
  ( set -e; ctl_run "$CTL_ID" ) >"$ROOT/logs/$CTL_ID.log" 2>&1 &
  CPID=$!
  CSTART="$(proc_start "$CPID")" || CSTART=""
  set +e
  wait "$CPID"
  rc=$?
  set -e
  CPID="" CSTART=""
  # A registered launch of this control that is still running is stopped and reaped before the control's
  # resources are swept and its scratch tree removed.
  g="$(launch_groups "$CTL_ID")"
  if [[ -n "$g" ]] && own_scan "" "$g" && [[ -n "$OWN_SCAN" ]]; then
    prob "a nested launch outlived the control; its process tree was stopped"
    if stop_tree "" "$g"; then :; else prob "owned processes survived KILL: $STOP_LEFT"; held=1; fi
  fi
  # G064 correction 11: the control's inventory, on whatever path the control ended (see owned_cleanup). It is deleted
  # only when the cleanup, the sweep and any stop above are verified; otherwise the control is held (see owned_ids).
  # G064 corr1 C2 / corr2 D2: a refused creation (ceiling row) or a set latch fails the control; cleanup still runs.
  while IFS=$'\t' read -r m g _; do prob "resource ceiling $g for $CTL_ID ($m): creation refused"; done < <(awk -F'\t' '$1 == "ceiling" {print $2 "\t" $3}' "$ROOT/owned/$CTL_ID.tsv" 2>/dev/null)
  if [[ -e "$ROOT/owned/$CTL_ID.hold" ]]; then prob "held: resource ceiling latch of $CTL_ID set ($(awk -F'\t' '$1 == "reason" {printf "%s%s", s, $2; s = "; "}' "$ROOT/owned/$CTL_ID.hold" 2>/dev/null))"; fi
  if owned_cleanup "$CTL_ID"; then :; else for m in "${OWNED_MSG[@]}"; do prob "cleanup: $m"; done; held=1; fi
  if [[ -s "$PROB" ]] && (( OWNED_N > 0 )); then prob "cleanup after the control removed $OWNED_N of its recorded or outer-labelled objects"; fi
  if sweep "sdlc.ci.control=$RT-$CTL_ID"; then :; else prob "control resources could not be removed: ${SWEEP_LEFT% }"; held=1; fi
  if (( held )); then
    HELD+="$CTL_ID "
    if [[ -f "$ROOT/owned/$CTL_ID.tsv" ]]; then m="inventory $ROOT/owned/$CTL_ID.tsv (recorded IDs: $(owned_ids "$CTL_ID"))"; else m="no inventory; label sdlc.ci.control=$RT-$CTL_ID"; fi
    prob "cleanup unverified, held: $m and receipt $ROOT/problems/$CTL_ID kept with the scratch root; no later control is started"
  else
    rm -rf -- "$ROOT/owned/$CTL_ID.tsv" "$ROOT/owned/$CTL_ID.hold" "$ROOT/owned/$CTL_ID.lock" "$ROOT/owned/$CTL_ID.rsv" "$ROOT/owned/$CTL_ID.cid"
  fi
  chmod -R u+rwx "$ROOT/cases" 2>/dev/null || true
  if (( rc != 0 )); then prob "control aborted (exit $rc): $(tail -3 "$ROOT/logs/$CTL_ID.log" | tr '\n' ' ')"; fi
  CUR_CTL=""
  if [[ -n "$TIMES" ]]; then printf '%s\t%s\n' "$CTL_ID" "$((SECONDS - t0))" >>"$TIMES"; fi
  if [[ -s "$PROB" ]]; then
    FAILN=$((FAILN + 1)); out "FAIL $CTL_ID $(tr '\n' ';' <"$PROB")"
  else
    PASSN=$((PASSN + 1)); out "PASS $CTL_ID $(cat "$DET")"
  fi
  if (( KEEP == 0 && held == 0 )); then rm -rf -- "$ROOT/cases/$CTL_ID" "$ROOT/cases/$CTL_ID-kill" 2>/dev/null || true; fi
done
if (( ${#NOTRUN[@]} > 0 )); then
  out "FAIL not-started ${#NOTRUN[@]} later control(s) were not started after the unverified cleanup of ${HELD% } (not passed): ${NOTRUN[*]}"
fi
if sweep "$RUN_LABEL"; then :; else FAILN=$((FAILN + 1)); out "FAIL cleanup control resources of run $RT remain: ${SWEEP_LEFT% }"; fi
# The run is complete: a signal from here on no longer cancels it.
trap '' TERM INT HUP
out "ci-controls: ${PASSN} passed, ${FAILN} failed (of ${#SELECTED[@]} selected; ${#CONTROLS[@]} defined)${NOTRUN[*]:+; ${#NOTRUN[@]} not started (cleanup held)}"
if (( FAILN > 0 )); then exit 1; fi
exit 0
