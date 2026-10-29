> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G051 — CI feedback and build optimization work specification

Goal: [G051-ci-feedback-optimization.md](../goals/G051-ci-feedback-optimization.md)
State: Defined for Claude planning; implementation requires separate Plan ACCEPT

## Rationale, Scope and Prerequisites

The accepted G049 pipeline `2929968148` spent approximately 6.1 minutes in fast and 60.5 minutes in gate. Current integration selects much of the whole-stack suite; feature browser/fixture paths select full regardless of the feature's boundary. Such qualification is useful for foundation changes, but should not be the default cost of ordinary page or service work. These observations are repository evidence, not an industry benchmark.

Optimize the actual post-G050 implementation using existing GitLab fast/gate jobs and Docker-first tools. Preserve accepted G049 selection exceptions, G050 JUnit/TSV/privacy semantics and any integrated G005 model/migration evidence. Inspect real changes before planning; do not assume historical counts such as 33/77 remain current. G005/G050 continue under their current approvals. Only management artifacts are authorized by this definition.

## Requirements and Architectural/Security Invariants

**R1 — Measure before optimizing.** Use existing successful job traces/reports to separate image/tool pulls, builds, container startup/readiness, check execution, controls, scanning and publication/cleanup. Where existing timing is insufficient, add bounded timing during approved execution. Identify concrete repeated work and the baseline suites/builds for each profile. No additional preapproval full run or broad benchmark campaign is required.

**R2 — Explicit coverage obligations.** Define a reviewed mapping from changed production/test/fixture/helper/configuration paths to affected check IDs/suites and transitive service/contract dependencies. Selection must include the changed behavior's tests; adding a test cannot make its production path qualify for less coverage. Account for every current required check: retained in full qualification, executed on a stated routine trigger, or also required whenever its underlying boundary changes. Split an aggregate check only with a documented mapping to its original scenarios and equivalent assertions. No test or scenario removal is authorized. Completeness validates the selected obligation set exactly; an unexplained missing, duplicate, malformed, failing or UNMET required outcome fails. An affected pass must not be labeled full.

**R3 — Required trigger matrix.** The following are the minimum new obligations, using the existing profiles rather than inventing a second gate authority:

| Change/context | Required automatic coverage |
|---|---|
| Existing exempt documentation/planning-input JSON/exact handoff files only | No automatic pipeline; preserve precise accepted exemptions. |
| Other root `.agents/` or `.claude/` changes | Fast as accepted by G049; no automatic full merely by root location. |
| Known ordinary page/copy/layout change within existing UI behavior, plus its feature-specific tests/fixtures | Fast static/build checks and the relevant browser/accessibility/responsive cases. No unrelated datastore, outage or rename campaign. |
| Known ordinary core/intelligence runtime change without a foundation/security boundary change | Fast plus affected service unit/integration checks and its transitive contracts/callers. Start PostgreSQL, Redis or object storage only when selected scenarios use them. |
| A feature-specific test/fixture added, changed, deleted or moved | Its owning feature/service coverage; existing shared/security cases remain mandatory when affected. An unmapped test/fixture selects full. |
| Migrations/migration runner; identity/session/authorization/tenant isolation; shared tokens/themes/fonts/portals/first paint; shared global UI primitives; contracts/dependencies/locks; Compose/build/CI controls; root `skills/**`; subtree `.claude/skills/**`, Vestive link and other existing foundation triggers | Full qualification, preserving the accepted boundary. Refining an existing trigger needs an explicit reviewed coverage mapping; the listed security/data/foundation classes must not silently become ordinary fast changes. |
| Exact release candidate, existing main weekly diagnostics, or explicit full request | Full qualification. Manual/API defaults remain full; narrower overrides require an explicitly reviewed contract and cannot narrow a required full change. |
| Unknown/unmapped path, missing/invalid diff/base, unresolved dependency or classification | Full, fail closed. |
| Mixed changes or cross-boundary rename | Union of obligations; full dominates. Evaluate old and new paths. Deletions cannot hide the former owner's coverage. |

Do not claim all UI is fast or all backend is full. Thin page changes that affect authentication or another security contract inherit that wider contract. Routine core-only work must not build/test intelligence or web unless its actual dependency closure requires them. Preserve supported push/MR/staging/main contexts and duplicate-pipeline suppression. Existing protected identity qualification remains separately controlled; ordinary changes do not invoke live WorkOS.

**R4 — Normal build and test reuse.** Build each required image/input combination at most once per job, unless a scenario explicitly tests rebuild/reset/different variants. Reuse that build for relevant checks. Avoid repeating successful static/build work for identical inputs merely to cross a profile boundary; reuse only attributable same-candidate results, and retain required distinct environment/variant scenarios. Full outage, persistence, empty-volume, rename and isolation cases remain in full qualification. Do not parallelize checks against shared ports/databases/volumes without positive isolation. Stay within ADR-002 medium hosted runner/TLS/privilege boundaries.

Cache use is optional, not a prerequisite for acceptance. If used, exact source/build-context, lockfiles, tool/base-image versions/digests, architecture and configuration inputs define validity. Cache contents must contain no credentials, runtime data or raw evidence. An untrusted branch cannot supply executable/image state to a protected pipeline without independent input/content validation; when trust/validity is absent, rebuild. A cache hit is not evidence that tests passed. No new registry credential, vendor, persistent daemon or publication scheme is implicitly authorized.

**R5 — Measurable feedback targets.** On the existing medium hosted runners, the final candidate must meet these cold-run upper limits:

| Representative change | Required coverage | Pipeline wall time | Sum of billed compute minutes at current factor 2 |
|---|---|---:|---:|
| Ordinary UI/static feature | Fast plus required focused UI checks | 5 minutes | 10 |
| Ordinary backend feature | Fast plus affected integration | 15 minutes | 30 |
| Foundation/full change | Fast plus comprehensive qualification | 45 minutes | 90 |

These are project acceptance targets. Cold means fresh job daemons/workspaces and no project build/dependency cache; normal published base/tool-image pulls are included. Measure earliest required job start through last required job completion, including bootstrap, pulls/builds, tests, scans, artifact/report upload and cleanup. Exclude only GitLab runner queue time; report it separately. Sum all required job durations/multipliers so parallelism cannot hide increased compute consumption. A failing, canceled, timed-out, incomplete or retried run cannot demonstrate the target; retain its history. Do not meet targets by relaxing timeouts, terminating required checks early, hiding setup costs or moving checks into an uncounted pipeline.

Use one successful cold representative run per class, with exact scenario paths/check obligations and candidate/input provenance. This is an acceptance demonstration, not a throughput/SLA percentile claim. Reuse the ordinary final full integrated staging candidate run for full correctness and timing under the temporary solo-operator policy. Plan the minimum additional representative hosted runs (at most two) for routine UI/backend selection; bind any disposable fixture commit and prove qualification-only changes do not become application delivery. No fabricated local runtime estimate substitutes for hosted evidence. Warm-cache evidence is optional and cannot replace cold acceptance. If a target cannot be reached within scope, stop for Codex/owner revision rather than accepting a higher duration silently.

**R6 — Retain execution and evidence guarantees.** Preserve approved-input pre/post checks, original nonzero harness/phase precedence, complete-required validation, ownership-based cleanup, credential protection, artifact scan-before-publication, retention/access and G050 report consistency. Account for each selected/excluded check and why; do not publish skipped work as PASS. Use the same selected obligation set for execution/completeness/TSV/JUnit reporting. Keep exact-candidate evidence and rebase/reconciliation rules; old timings are not acceptance of changed inputs. No application/public API/MCP/schema/vendor/authority changes or G021 product work.

## Expected Artifacts

A concise Claude Plan with bottleneck evidence, concrete changed paths, check/trigger coverage mapping, cache decision, representative scenarios and criterion mapping; minimum existing selector/harness/build/completeness changes; focused controls; operations documentation; Result with actual timings, profile/check selection, provenance, failures/limitations, cleanup and B5 accounting. No new general orchestration framework.

## Acceptance Criteria and Verification

| ID | Required outcome | Required verification/evidence |
|---|---|---|
| AC1 | Approved implementation is based on accepted integrated predecessors and measured bottlenecks. | Pin baseline/tree/inputs and G049/G050 reviews; reconcile intervening G005 changes. Attach existing baseline phase evidence and concrete repeated-build/suite findings; Plan identifies only the affected optimization sites. |
| AC2 | Every existing required full check/scenario is preserved; selected sets have exact completeness. | Compare pre/post registries and scenario mappings. Full run executes all final required obligations. Negative fixtures omit/duplicate/corrupt an obligated outcome and plant FAIL/UNMET; each fails rather than reporting success. |
| AC3 | Routine UI and backend changes select their reviewed affected coverage without unrelated whole-stack work. | Existing selector/control fixtures cover at least page+its browser test, standalone feature-test/fixture changes, core-only, intelligence-only and a cross-service dependency. Inspect actual selected IDs and built/started services; backend dependency changes include transitive consumers. |
| AC4 | Boundary, mixed and invalid-input selection fails closed across supported pipeline contexts. | Controls cover migrations, auth/isolation, global UI/theming, lock/contract/infra, root/subtree skills, root-agent fast, exact documentation exceptions, unknown paths, invalid/missing bases, moves/deletions, shared helpers and mixed changes. Assert both workflow admission and selected obligations; explicit full/release/weekly cases cannot narrow. GitLab configuration lint passes and aliased changes lists remain within platform limits. |
| AC5 | Build/startup reuse reduces redundant work while preserving isolation and cache correctness. | Record required images/input digests/build counts; no repeat identical build per job without a named rebuild scenario. If cache is used, stale source/lock/base-image/config and untrusted-cache negatives rebuild/refuse; cold run passes without project caches. Prove cleanup/ownership and independent variant scenarios remain. |
| AC6 | Ordinary UI/static cold pipeline meets five-minute/ten-compute-minute targets. | One attributable successful hosted representative run; exact changed paths, required checks, complete job timestamps/durations, multiplier and no project-cache state. Include focused browser cases when the representative UI behavior requires them. |
| AC7 | Affected backend cold pipeline meets fifteen-minute/thirty-compute-minute targets. | One attributable successful hosted representative run with real affected integration and required datastore round-trip, not a static-only stand-in; job timings, check completeness and service dependency closure recorded. |
| AC8 | Full final qualification meets forty-five-minute/ninety-compute-minute targets with inherited required scenarios. | Final candidate's applicable hosted full run, all required IDs/scenarios, clean status and timestamps/compute accounting. Full outage/persistence/rename/variant/security checks remain; no distinct necessary scenario is replaced by cached evidence. |
| AC9 | Selection/optimization cannot weaken status, privacy, reports or cleanup. | Focused existing failure-path/privacy/cleanup controls cover changed sites. Successful postchecks cannot mask failing tests; malformed selection/completeness fails. G050 TSV/JUnit/Tests-tab counts/statuses reflect actual executed obligations; no unsafe artifact/cache publication. Preserve protected identity job configuration without requiring an unrelated live provider run. |
| AC10 | Operations and Result describe a reproducible final CI contract and limitations. | Document profile/trigger matrix, precise runtime/queue/cold/warm definitions, full/release/weekly use, cache trust and focused debugging. Record commits/inputs, exact commands/jobs/artifacts, failed/interrupted history, exclusions, resource cleanup and cycles. Codex reviews actual diff before acceptance. |

## Completion Gates, Permitted Deviations and Open Risks

Claude owns implementation choices and concrete file/check mappings. Existing suite organization may change only with preserved scenario coverage and explicit reviewed mapping. No target relaxation, security boundary change, scenario deletion, new runner/cache vendor or application refactor is a permitted deviation.

Before implementation approval, resolve: the actual slow phases from accepted logs; dependency ownership for feature browser tests/shared helpers; which minimum harness split achieves selective execution without duplicating authority; whether same-job reuse alone reaches cold targets; and the exact hosted representative runs/action permissions. If the current harness cannot support these within B5, propose the smallest split with evidence before writing implementation. Cache trust is a gate only if caching is chosen. No purchase, GitLab setting/schedule creation or broad reconnaissance run is needed to draft the Plan.

All ten criteria and required evidence remain in G051. Full qualification remains mandatory for listed boundaries, exact releases and existing weekly diagnostics; no new schedule is created by this Spec. Until G051 is accepted/integrated, current path mappings continue unchanged except for the owner-authorized temporary solo-operator gate placement recorded below. The final integrated CI contract is reviewed at G021; no Beginify product feature is authorized.


## Owner sequencing amendment — 2026-10-09

The temporary solo-operator policy in AGENTS.md supersedes automatic MR full
qualification: local required coverage before MR, automatic MR fast (optional
manual gate), then existing automatic staging gate after conditional integration.
G051 must preserve that placement while providing affected selection; its final
full hosted correctness/timing evidence comes from the integrated staging
candidate. Final ACCEPT and dependent execution await required staging success.
Ten criteria, targets, budgets, cold evidence and all mandatory scenarios remain;
no implementation is authorized by this amendment.

## Owner execution amendment — Batched failure recovery, 2026-10-10

Under AC9/AC10, G051 must apply and document the global validation-batch sequence
in AGENTS.md: finish a safe batch, preserve outcomes and verify cleanup, diagnose
and fix every outstanding failure, rerun failed batches and checks affected by
the fixes, then advance. Do not start later batches with unresolved failures or
repeat unrelated passing batches without changed relevant inputs. Unsafe/invalid
work still stops immediately.

Result evidence names batch selections, failed checks, fixes, rerun outcomes,
remaining stages and exact provenance for reused passes. Maintain an outstanding
findings checklist so errors from earlier batches are not omitted. Reconcile
inputs before reruns under existing rules. All ten ACs, timing targets, B5 and
required final/hosted qualification remain unchanged. Use existing suites and
controls; this is an execution/operations optimization, not authorization for a
new retry engine or changes to hosted job sequencing.

## Codex bounded correction amendment — 2026-10-10

The post-cycle-4 review (not included) authorizes the remaining cycle 5 for O1–O5. Added scope is limited to the pinned secret-scan configuration and the responsive browser test observer, alongside existing G051 harness/operations sites. L6 may be requalified only after the observer fix and three consecutive complete two-worker passes; targets/coverage remain unchanged. Per-run feature-report bind mounts are corrected rather than exempted. Preserve historical Plan/input bindings and explicitly reconcile the expanded paths before testing. No budget extension or implementation acceptance is granted.

### 2026-10-10 — Budget extension only

Owner authorizes **B7 maximum total shared cycles**, with a checkpoint after cycle 6, as recorded in the Goal and Codex review (not included). This supersedes earlier B5/no-extension statements; all ten criteria, timing caps, validation, scope and escalation rules remain. Existing counts are retained; extra cycles do not authorize redundant unchanged qualification or coverage reduction.

### 2026-10-10 — Goal-level attempt accounting

Owner clarifies that a budget increment is a goal-level execution attempt, with autonomous scoped debugging, test-batch recovery and approved subagents inside that attempt. Apply management/BUDGET_POLICY.md prospectively to G051: preserve the four historically consumed cycles; the current cycle 5 becomes active execution attempt 5. Its routine scoped failures/fixes do not independently consume attempts 6 or 7. Those reserves are used only for a new attempt after an unsuccessful goal-level attempt or supervisor-authorized correction round. B7 remains seven total, with a convergence checkpoint after attempt 6. Record boundaries, failures and delegated contributions; do not reset earlier counts. Escalate scope/security changes and a demonstrably failing strategy regardless of remaining allowance. Required acceptance, timing caps, input reconciliation, cleanup and selective validation remain unchanged.

### 2026-10-11 — Bounded post-integration correction override

Codex authorizes attempt 8: retain seven used attempts and B7 history, with one
exceptional additional attempt (maximum eight total) for H2 overlapping-build
attribution and repeated-definition vertex counting only. See the
[post-integration correction review](../results/G051-implementation-codex-review.md#post-integration-h2duo-review--2026-10-11--correction-loop).
Owner directs direct staging publication without local test runs before push;
review hosted evidence later. Required criteria, targets, isolation, fail-closed
ledger and final acceptance are unchanged. No automatic further extension.

### 2026-10-11 — H4 failure: bounded observer/control correction

Codex extends the maximum from eight to nine total goal-level attempts, preserving
eight used and all prior B7/override history. Attempt 9 is limited to VC8's
missing-progress fixture and the C4 shared-vertex service-prefix predicate,
with a code handback checkpoint before replacement hosted qualification. See
[H4 correction review](../results/G051-implementation-codex-review.md#h4-2936339771-review--2026-10-11--correction-loop).
Targets 45/90 and all ten criteria remain; timing miss stays unresolved. No
automatic further extension or immediate duplicate full/local test campaign.


## Owner-approved AC8 closeout deviation — 2026-10-11

Owner accepts the measured timing of successful pipeline2936644423 on4bab0e7:46.59wall-minutes (46.55first-job-to-last-job boundary),93.05compute-minutes, against original45/90targets. Response “I'm ok with 46 minutes” accepts this exact run after those measurements were presented. This one-off exception supersedes the no-target-relaxation clause only for G051AC8 closeout; the original criterion and historical misses remain unchanged, not falsely markedMET. No general limit change, scenario waiver or authorization for later slow runs. See [final ACCEPT](../results/G051-implementation-codex-review.md#owner-timing-disposition-and-final-implementation--2026-10-11--accept).
