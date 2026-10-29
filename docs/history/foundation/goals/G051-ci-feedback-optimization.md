> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G051 — CI feedback and build optimization

Epic: EPIC-001 — Repository and local platform follow-up
Status: Defined for Claude planning; implementation not authorized
Implementation owner: Claude Code
Supervisor and acceptance reviewer: Codex
Budget: B7 — maximum seven shared meaningful cycles; original checkpoint after cycle 3, extension checkpoint after cycle 6

## Objective

Make routine development use a bounded standard build and affected tests rather than hour-long foundation qualification. Meet measured cold-run pipeline targets of five minutes for routine UI/static work, fifteen minutes for affected backend integration, and forty-five minutes for comprehensive qualification on the accepted medium hosted runner. Preserve required behavior/security coverage and honest evidence.

## In Scope

- Attribute current CI time from existing evidence and inspect actual accepted code before proposing changes.
- Refine the existing fast/affected/full selection and completeness machinery to select the affected service and its transitive contracts/tests.
- Stop feature-specific browser tests from forcing unrelated foundation qualification; shared/security-sensitive tests retain wider coverage.
- Reduce duplicate builds, startup and repeated identical validation within existing jobs; bounded cache reuse only under the Work Spec's input/trust rules.
- Required selection/completeness/cache-negative checks, hosted performance evidence, operations guidance and a Result.

## Out of Scope

- Application behavior, product features, schema/migrations, identity/authorization redesign or fixing unrelated tests.
- New runner/vendor/subscription, privilege changes, deployment, generalized orchestration, broad harness rewrite or framework replacement.
- Removing tests, weakening assertions, tolerating required failures, converting UNMET to success, or dropping full qualification.
- Rewriting accepted G003/G004/G048/G049 evidence or changing active G005/G050 execution bindings.

## Dependencies

- G049 accepted/integrated; G050 accepted/integrated before execution, preserving its final reporting/publication contract.
- Reconcile any G005 or other predecessor integrated before the actual execution baseline. G005 need not wait for G051 and G051 does not require G005 completion to begin planning.
- Canonical governance, architecture/security/decisions, ADR-002, roadmap/progress, BUDGET_POLICY, actual CI/harness/required-check registry and inherited acceptance evidence.
- A concise Claude Plan ACCEPT, fresh committed input/scope binding and named hosted actions before implementation. Current definition baseline is staging `e59bc40`; this is not an execution snapshot.

## Completion Definition

AC1–AC10 in the [Work Spec](../specs/G051-ci-feedback-optimization-spec.md) pass. Codex accepts the actual implementation and timing/coverage evidence. Routine paths meet the cold-run targets; full qualification retains all required current checks. G021 remains the product-feature barrier and assesses the final integrated CI contract.

## Execution Policy

Architecture/security/data execution tier because this changes merge-gate coverage. Use a short concrete Plan, explicit approval and existing wrappers/resource/privacy controls. Before approval, inspect existing logs/code and define checks; do not build a replacement validation system or run broad qualification probes. During approved execution, use focused controls before the bounded final hosted acceptance runs. Hosted CI supplies final equivalent automated coverage; retain distinct BASE-aware/manual evidence.

The Plan must map existing check IDs and relevant test paths to their new execution obligations. Ordinary implementation choices within that reviewed mapping are autonomous. No blanket permission to reduce coverage or alter accepted ADRs. Do not block independent approved feature work; coordinate shared CI integration separately.

## Iteration Budget

**B5: five shared meaningful cycles maximum; checkpoint after cycle 3.** Forecast three to four with one reserve. Stop on exhaustion or a repeated strategy that cannot meet targets without weakening guarantees. Hosted measurement is bounded by the Spec; repeated full runs need a relevant change/failure, not a desire for additional benchmark samples.

## Context / Scope Budget

Existing GitLab jobs, selector/profile map, required-check/completeness logic and only the harness/build sites responsible for measured duplication. No new generalized scheduler, resource registry or execution framework. If these bounds cannot deliver the targets, propose a concrete split before implementation rather than starting an unbounded CI redesign.

## Suggested Subagents / Delegation

Claude may use one read-only coverage/security reviewer. Additional implementation delegation must have disjoint file ownership and remain within B5. Codex acceptance and integration responsibility are not delegated.

## Stop / Escalation Conditions

- Required predecessor/input drift, active-checkout overlap or unavailable hosted authorization.
- A selected plan cannot account for every existing required check or needs an unreviewed coverage exclusion.
- Runner/privilege/vendor change, new credentials or registry publishing, security/authority change, application refactor or substantial harness redesign.
- Targets appear infeasible within scope: provide measured bottlenecks and the smallest proposed adjustment; do not silently raise targets or drop checks.
- Selection cannot resolve a dependency safely: use full coverage and report the missing mapping; do not infer irrelevance.
- Budget exhaustion, unsafe cache/resource behavior or unavailable required acceptance evidence.

## Revision Record

- 2026-10-09 — Owner requests normal-sized build/feedback for the current project. Bounded specification only; no implementation, test execution, schedule/settings change or active-goal gate amendment.

## Owner execution amendment — Batched failure recovery, 2026-10-10

The global AGENTS.md rule, **Validation batches: finish, fix, rerun, then advance**,
is part of G051's feedback optimization. Let each safe batch finish collecting
failures; fix all outstanding errors across batches, rerun failed batches and
affected checks, then continue remaining stages. Reuse unchanged passing evidence
with exact provenance instead of restarting unrelated batches. Preserve immediate
safety/invalid-prerequisite stops, cleanup, reconciliation, B5 and all required
final local/hosted evidence. This amends active execution sequencing without
rewriting historical Plan/snapshot bytes. No new runner, scheduler, general retry
framework or CI job redesign is authorized. Record adoption in the Result.

## Codex bounded correction amendment — 2026-10-10

The post-cycle-4 review (not included) authorizes the remaining cycle 5 for O1–O5. Added scope is limited to the pinned secret-scan configuration and the responsive browser test observer, alongside existing G051 harness/operations sites. L6 may be requalified only after the observer fix and three consecutive complete two-worker passes; targets/coverage remain unchanged. Per-run feature-report bind mounts are corrected rather than exempted. Preserve historical Plan/input bindings and explicitly reconcile the expanded paths before testing. No budget extension or implementation acceptance is granted.

### 2026-10-10 — Owner-authorized budget extension

Current allowance is **B7: seven total shared meaningful cycles maximum**, superseding earlier B5 limits and cycle-5 exhaustion statements. Four cycles are completed and cycle 5 is active; cycles 6–7 are reserves. Record a convergence checkpoint after cycle 6 before using cycle 7. Preserve prior cycle accounting and stop on exhaustion or repeated unresolved strategy failure. Scope and acceptance are unchanged; no duplicate unchanged full-suite runs are authorized. See the budget extension and bounded recovery release (not included).

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


### 2026-10-11 — Owner starts bounded C6 correction attempt10

Owner directs “Start the fix.” Codex raises maximum9to10total attempts, retaining nine used and B7/override history. Scope: VC9 three-row malformed-prefix expectation; positively target-bound lone-service-label aliases; focused existing controls and minimal sanitized unknown-vertex diagnostics. No speculative terminal relaxation, application change, target relaxation or whole local suite. Focused changed-behavior checks are released; handback checkpoint precedes publication/hosted qualification. See [C6 correction contract](../results/G051-implementation-codex-review.md#owner-starts-bounded-c6-correction--2026-10-11--correction-loop). No automatic extension beyond10.


### 2026-10-11 — Final ACCEPT, completed/integrated

G051 is ACCEPT at staging4bab0e7 after successful2936644423, complete/MET/cleanup evidence and owner's one-off46.59wall/93.05compute timing disposition. Ten used attempts/max10 retained; no further execution authorized for closeout. Historical header/status/budget and original45/90targets are preserved by the dated amendments. [Final review](../results/G051-implementation-codex-review.md#owner-timing-disposition-and-final-implementation--2026-10-11--accept).
