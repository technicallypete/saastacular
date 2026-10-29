> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G003 — Health, logging, contracts, and baseline CI

Epic: EPIC-001 — Repository and local platform
Status: Authorized under the owner-requested sequential execution addendum; gates apply
Implementation owner: Claude Code
Supervisor and acceptance reviewer: Codex
Budget class: B7 — Complex; maximum seven shared meaningful cycles, checkpoint after cycle 4

## Owner-authorized sequential execution override — 2026-10-05

The product owner requests one uninterrupted execution batch: G047 first, G003 second, with both Results returned together for Codex review. The supervisor execution authorization (not included) governs this bounded override. G047 must meet all seven criteria and complete its final validation before G003 implementation begins; intermediate Codex implementation acceptance is deferred to the joint review. G003 r4 technical implementation is authorized within that contract. Local implementation may proceed while hosted access is investigated; hosted actions require the unchanged runner/security checks. Required hosted evidence is not waived. This dated override supersedes earlier planning-only, intermediate-acceptance and literal baseline-supply wording solely as specified in the authorization. B7, scope, security and all AC1–AC10 remain unchanged.

## Objective

Complete the remaining operational foundation of the accepted repository: demonstrate bounded health/readiness behavior, propagate diagnostic request/trace context across the existing web → core → intelligence health path, produce structured privacy-safe service logs, validate existing cross-service contracts in both languages, and run the containerized checks in GitLab CI with attributable failure and success evidence.

This is a foundation increment. It introduces no product feature, tenant schema, authenticated user surface, public AI execution path, telemetry vendor, or deployment. The product-feature barrier remains closed through G021.

## Context and rationale

On 2026-10-05, Codex inspected commit `7193ed84237be9362b7b596bdd135dd592447b07` (`G002: accessible web shell (accepted)`). Normal and ignored-inclusive working-tree inventories were empty before these artifacts. This commit includes accepted G001/G002 implementation and PROGRESS.md acceptance. Historical submission wording in closed Goal/Spec/Result files does not override that acceptance.

G003 remains the next dependency-ready roadmap Goal. Actual state narrows its work:

- Core already has dependency-aware readiness, dependency-independent liveness, JSON redaction and inbound request/trace IDs. Web already has liveness and a bounded server-only core readiness bridge. Intelligence already has health/readiness and privacy-safe execution export scaffolding.
- Core's optional intelligence probe and web's core readiness fetch do not propagate the diagnostic context. Core currently retains a trace ID, not an inspectable end-to-end parent relationship. Web lacks corresponding request-completion logging; intelligence lacks correlated health-request logging.
- Two canonical intelligence JSON Schemas and Python validation/tests already exist. Core has no independent cross-language validation of their shared corpus. Do not add a production execution client to demonstrate that validation.
- Docker builds/check profiles, the project harness, migration/isolation negative controls, and the G002 browser matrices already work. There is no CI definition. The configured remote is GitLab, not GitHub.

The increment is integration and evidence for these gaps, not a replacement health stack, logging framework, or general CI platform.

## In Scope

- Preserve and verify existing health/readiness paths, status bodies, optional intelligence semantics, deadlines, no-store responses and outage/recovery behavior.
- Diagnostic request/trace context validation and propagation on the existing readiness calls; service request-completion logs and bounded error redaction.
- Shared fixture-based validation of the existing intelligence envelope/event schemas by Python and an independent Go validator, using synthetic data and fake providers only.
- A GitLab CI definition and repository-owned containerized entry point covering builds, static checks, service checks, contracts and integrated regressions, including accepted G002 browser coverage.
- CI failure controls, sanitized change/run artifacts, execution-owned cleanup, and bounded project harness/provenance reconciliation.
- Focused README/foundation documentation of commands, operational conventions, inherited versus new guarantees, and remaining work.

## Out of Scope

- Authentication, sessions, IdentityProvider selection, tenant context, Organization/Group/BusinessEntity records, RLS, storage authority, audit/usage persistence or retention jobs.
- Public MCP mounting, new API routes, public AI streaming, production core → intelligence execution, callbacks/tool execution, product AI workflows or live provider calls.
- UI redesign, token/palette/component changes, marketing work, font assets, G002 accessibility follow-up implementation, and Organization branding storage.
- Metrics endpoints/exporters, collectors, dashboards, alerting, Langfuse/OpenTelemetry backend adoption, external trace export or full distributed tracing certification.
- Deployment, registry publishing, production credentials/provisioning, vendor replacement, runner fleet provisioning, privileged-runner security changes, vulnerability-program expansion or broad dependency upgrades.
- Refactors or cleanup unrelated to an explicit Work Spec requirement; changes to accepted Skills/templates, original renderer manifest, canonical architecture/governance or closed-goal records.

## Dependencies

- **G047 accepted before G003 execution:** the owner-requested [successor Vestive shell refresh](G047-vestive-shell-refresh.md) is inserted ahead of G003. G003 planning corrections may finish, but its final Plan must reconcile affected baseline/provenance, browser expectations and feasibility after G047 acceptance. Obtain a fresh Codex approval and execution binding; no G002/G003 continuity snapshot authorizes work against the refreshed tree. G003's UI implementation exclusion and B7 remain unchanged.
- Roadmap dependency: **G001 accepted**. Its actual Compose stack, roles, migrations, health behavior, dormant MCP boundary and container checks remain required regressions.
- **Accepted G002 is the inspected repository baseline and protected regression input**, not a new roadmap renumbering/dependency decision. Preserve its appearance/first-paint, scoped tokens, browser suites, dependency controls and wrapper semantics.
- Read AGENTS.md, CLAUDE.md, ARCHITECTURE.md, SECURITY.md, DECISIONS.md and accepted ADRs if present, ROADMAP.md, PROGRESS.md, BUDGET_POLICY.md, prior Goal/Spec/Plan/Result and acceptance/deviation records, and the [G003 Work Spec](../specs/G003-foundation-observability-ci-spec.md).
- The product owner confirms an **existing Docker-capable GitLab runner** is to be used. Before Plan approval, identify its execution profile and verify Docker/Compose/browser support, isolation, capacity and hosted-evidence permissions. No runner configuration was inspected during this task. The Plan must resolve those details or return a specific blocker; local CI parity is not a substitute for hosted evidence.

## Completion Definition

1. Work Spec AC1–AC10 pass with no required scenario skipped, unmet or deferred.
2. The Result maps each criterion to exact commands and sanitized artifacts from the current implementation, distinguishes local from hosted results, and records all failures, deviations, cleanup and cycle accounting.
3. Hosted GitLab evidence identifies the tested commit/tree, pipeline, jobs and the required success/failure cases. No production publish/deploy action occurs.
4. Codex reviews implementation, evidence, architecture/security conformance and budget behavior, returns ACCEPT, and updates PROGRESS.md. Plan approval and a green pipeline alone do not complete G003.

## Execution Policy

This Goal and Work Spec authorize read-only inspection and Claude Plan preparation only. Claude must submit `management/v0/plans/G003-foundation-observability-ci-plan.md` for Codex approval before installing dependencies, changing implementation/tests/CI, running initialization or building application images.

After explicit Plan approval, Claude may implement and validate the enumerated changes within this Goal. The Plan must pin a starting commit and reviewed inputs, enumerate exact changed paths/dependencies/observable header changes, resolve the CI execution profile and failure controls, and map every acceptance criterion. Use the existing approved-input wrapper with a fresh G003 binding; do not reuse G002's historical snapshot or approval as authorization for G003.

Preserve existing developer .env, Docker resources and data. Destructive/outage/migration trials run only on disposable owned copies containing the submitted implementation. No commit, push, pipeline trigger, external configuration or deployment is authorized by this planning handoff. Obtain separately requested authorization for hosted evidence actions before taking them.

## Iteration Budget

**B7: maximum seven meaningful implementation/debug cycles shared across all implementation agents; mandatory checkpoint after cycle 4 and before beginning cycle 5.** Follow BUDGET_POLICY.md. No prior Goal's budget carries forward and no automatic correction allowance or extension is granted.

B7 is the smallest appropriate class for the three-runtime correlation/privacy boundary, independent contract semantics, Docker/GitLab integration and full accepted-stack regressions. B3 is too narrow; B5 leaves insufficient room for the first hosted execution and failure-control integration. The Plan must forecast completion within B7 after checking the runner and validators, or propose a bounded split before implementation. The class does not authorize extra scope.

At the checkpoint record per-criterion status, failure convergence, remaining work and forecast. A shared harness/logger/contract change requires its own affected checks as well as relevant runtime regressions; do not repeat G002's omission of static checks after a harness change. Stop at exhaustion or thrashing. Planning, read-only review and successful checks do not independently consume cycles.

## Context / Scope Budget

One bounded operational foundation increment using existing runtime paths and commodity test tooling. No new runtime subsystem or future domain scaffolding. Keep contract conformance code separate from application authority and CI execution separate from production configuration. If a dependable Plan requires service redesign, a runner security-model change or deferred required evidence, stop and ask Codex to revise/split scope.

## Suggested Subagents / Delegation

Prefer one implementation owner. Claude may propose bounded read-only reviewers for correlation/privacy and CI/contracts/provenance. Any implementation delegation requires non-overlapping exact file ownership, integration responsibility and shared-cycle accounting in the reviewed Plan. Delegation cannot decide architecture, scope, runner privileges or acceptance.

## Stop / Escalation Conditions

- Canonical architecture/accepted ADR conflict, scope expansion, security-model change, new major infrastructure/vendor, privileged runner adoption or irreversible action.
- Any proposed product/domain schema, durable entity identity change, authenticated API/MCP contract, public AI execution, live provider dependency or tenant authority in trace headers/fixtures.
- Need to remove/weaken an accepted check, relax correlation/privacy assertions, rewrite a closed record, alter Skill/manifest provenance, or substitute local evidence for required hosted evidence.
- Unavailable runner, insufficient Docker/Compose/port resources, denied hosted-action authorization, incomplete baseline/snapshot binding, or an existing developer-resource conflict. Stop affected work and report the exact unmet prerequisite; do not delete unrelated state or silently use another CI vendor.
- B7 exhaustion/thrashing, or a forecast/context size that cannot satisfy every criterion within the approved budget.

## Revision Record

- 2026-10-05 — Initial G003 definition after committed accepted G002 and repository reassessment. No Plan or implementation created; user requested management/spec artifacts only and no push.
- 2026-10-05 — Product owner requests updated Vestive visuals before G003 execution. Add G047 acceptance and subsequent Plan/baseline reconciliation as prerequisites; no operational scope, acceptance criterion, budget or implementation authorization changed. Original inspected baseline remains historical.


## Proposed bounded CI revision — 2026-10-06

The owner requests that G003 settle routine fast/integration feedback and a lower-cost GitLab hosted runner before acceptance. The [linked revision Work Spec](../specs/G003-ci-runner-feedback-revision-spec.md) preserves the original Goal/Spec/Plan and Result history and defines the exact limited supersessions, AC7a–AC7g, profile triggers, measured feedback targets and hosted evidence. Medium is the proposed default for fast, integration and full jobs; small is an optional later cost optimization. [ADR-002](../../../adr/ADR-002-gitlab-hosted-ci-runner.md) is PROPOSED for the changed DinD security profile, not accepted by this note.

Existing local correction validation may finish under its current binding. Do not inject revised inputs into that active run. Revised implementation requires owner ADR acceptance, a short linked Claude Plan revision, Codex Plan ACCEPT and a fresh exact binding. Proposed separate B3 revision budget: maximum three shared meaningful cycles, checkpoint after cycle 2; activation waits for that review. Original B7 and correction-pass accounting remain historical/separate. Full tests, hosted pass/failure proof, artifact privacy, isolation, G021 and all unchanged acceptance requirements remain mandatory. No application, G048, deployment, paid runner, purchase, schedule creation, push or pipeline action is authorized by this draft.

## Approved CI revision budget extension — 2026-10-06

At the owner's explicit request, Codex extends the active CI revision from **B3 (3 maximum, 3 used) to B5 (5 maximum, 3 used): two additional shared meaningful cycles**. Record the extension separately from the original B7 and correction pass 1; none of those counters resets. Reassess after revision cycle 4 before using cycle 5. This is a budget/scope supplement to the accepted r2/P0 execution contract, not a new Plan or permission to overwrite bound Goal/Spec/Plan/snapshot bytes in the execution checkout.

Reason: hosted full and integration runs completed but each failed B-TOGGLE-OFF; their remaining checks and runner controls passed. The Result §15 identifies an unproven cold-daemon assumption about Python base-image visibility. Remaining work is bounded: establish the failing assertion, make the smallest self-sufficient control correction without weakening intelligence-disabled guarantees, run focused regression validation, then obtain passing hosted full/integration evidence on the exact corrected tree. Preserve failed-run evidence and append diagnosis, correction and shared-cycle accounting to the Result.

Authorized implementation paths for this correction are `infra/validation/run.sh`, the existing relevant control test if required, and `docs/foundation/OPERATIONS.md` only where its description becomes inaccurate; append Result evidence. No application/schema/dependency/framework redesign or unrelated cleanup. Reuse the existing approved wrapper, nine-input binding, isolation/cleanup, exact-content evidence rules and named hosted actions. Do not copy subsequent main-checkout architecture/progress edits into the bound execution checkout. H2 must remain attributable under the approved reuse rules; rerun on the final tree when required, without weakening the intentional-failure proof.

Stop for a different material failure requiring expanded scope, two unsuccessful corrections of the same material issue, architecture/security changes or the five-cycle limit. Capacity/timing dispositions and project settings remain separate acceptance decisions; this extension does not accept G003, release G048, resume OS001 or authorize main/staging pushes, merges, settings, purchases or deployment.

### Owner-approved six-cycle maximum — 2026-10-06

The owner explicitly raises the active CI revision maximum from five to **six shared meaningful cycles** (custom B5 override). Four are used/in progress, leaving two available subject to actual recorded accounting; counters do not reset. This supersedes only the five-cycle stop above. The cycle-4 convergence checkpoint, bounded B-TOGGLE-OFF correction scope, protected inputs, acceptance/validation and all other stop conditions remain. Reason: the cold-daemon cause is now proven, its minimal control correction passed local validation, and final hosted runs have started; reserve one further cycle for a related correction if necessary. No unrelated work or repeated failing strategy is authorized.
