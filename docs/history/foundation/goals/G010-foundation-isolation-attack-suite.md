> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G010 — Foundation isolation attack suite

Epic: EPIC-003 — Authorization and isolation  
Status: r0 + owner amendment — planning released; implementation gated  
Owner: Codex Goal/Spec/review; Claude Code Plan/implementation/validation  
Date: 2026-10-14 UTC

## Objective

Prove that accepted identity, Organization/Group membership, derived
BusinessProfile policy, PostgreSQL RLS and object storage compose under realistic
isolation attacks. Deliver a repeatable permission/attack matrix and integrated
negative tests, with bounded repairs to existing foundation conformance defects.

## In Scope

- Independently specified actors, exact expected data/actions and multi-step attacks
  spanning the existing guarded core facade, runtime PostgreSQL and private MinIO.
- Cross-tenant privilege bleed, sibling/selected-Group substitution, shared-profile
  inheritance, sensitive-capability denial, current authority and connection reuse.
- Storage upload/read/delete/signing attacks, including current member own-upload
  deletion, denial without provider effects and documented bearer-URL residuals.
- Minimum test/CI entry-point wiring, focused foundation documentation and Result.
- Small conformance fixes in existing foundation modules, reviewed in the Plan or
  a bounded Codex correction; no redesign or changed authorization semantics.

## Out of Scope

Product features/UI; new public HTTP/BFF/MCP routes or connection behavior;
new capabilities, schema/migrations, profile grants, vendors or infrastructure;
live WorkOS/R2/cloud/operator qualification; durable audit (G011), retention
(G013), onboarding/admin workflows, identity redesign, a new attack/harness
framework, exhaustive combinatorial testing or benchmarking. G021 stays closed.

## Dependencies

- **Final G009 implementation ACCEPT and committed integration** before execution;
  transitively accepted G004–G008. G009 source reviewed at `5605a6c`, MR !24 at
  `4c7569e` with unchanged application bytes, merged as `e087f2c700df5d7db79b745e5d751143493900da`.
  Its staging disposition/final ACCEPT is pending at drafting, not inferred from merge.
- [Work Spec](../specs/G010-foundation-isolation-attack-suite-spec.md), canonical
  architecture/ADRs/governance, accepted predecessor Results/reviews/deviations,
  and the then-current CI/input/cleanup procedures.
- Conditional Plan preparation may use this reviewed merged source only after
  separate Codex planning release. Reconcile final G009 acceptance and any later
  changes before execution; G009's delivery-loop authority does not release G010.

## Completion Definition

Work Spec AC1–AC10 have attributable evidence, required local and hosted gates
are satisfied, and Codex reviews the actual diff and records final ACCEPT.
Close EPIC-003 only with the recorded progress-compaction disposition and owned
cleanup receipt; hand remaining audit/retention/public-adapter obligations to
their owning Goals. G011 becomes dependency-ready; G021 remains closed.

## Execution Policy

**Architecture/security/data tier.** This draft authorizes Goal/Spec work only.
Claude planning needs its own release; implementation requires final G009 ACCEPT,
explicit G010 Plan ACCEPT and fresh Git-bound inputs/scope. After release,
Claude may implement, diagnose and validate autonomously within approved scope.
Reuse existing fixtures/oracles/entry points and current batch/reconciliation,
safekeeping, MR review/auto-merge and post-integration acceptance rules. Testing
must exercise existing internal operations, without inventing public surfaces.

## Iteration Budget

**B5: maximum five shared Goal-level execution attempts; checkpoint after
attempt 3**, under BUDGET_POLICY (not included). Planning spends none;
delegates and scoped debugging share the allowance. Use B5 because this is a
bounded integration-test increment over accepted controls, not a new security
model. Recommend a split before implementation if that boundary is infeasible.

## Context / Scope Budget

One finite attack matrix over the accepted foundation. Add missing cross-boundary
chains and reuse inherited individual-control tests; do not clone their oracles,
build a second suite framework or repeatedly requalify unchanged controls.
No generalized fuzzing, pen-test platform, benchmark or broad predecessor refactor.

## Suggested Subagents / Delegation

One integration owner; optionally one independent expected-outcome author and
one read-only security reviewer with non-overlapping ownership. Reuse accepted
literal oracles; author only new chain expectations independently of production
policy/SQL/results. Claude owns integration and one shared stack. Planning xhigh,
implementation high; verify actual effort or record unknown. Follow current
task-prefixed worktree/role rules; no inherited G009 naming exception.

## Stop / Escalation Conditions

- Missing dependency/release, invalid binding or unavailable required runtime evidence.
- Ambiguous expected authority, oracle disagreement or architecture/ADR conflict.
- Repair requires changed security/capability semantics, schema, public contracts,
  broader identity/model redesign, new vendor/infrastructure or framework work.
- Need runtime owner/BYPASSRLS/root credentials, stale allow reuse, copied grants,
  cross-tenant provider claims or immediate revocation of issued bearer URLs.
- Credential/private-data leak, unowned destructive action, unsupported cleanup,
  unexplained material query/locking cost, repeated nonconvergence or B5 exhaustion.
  Preserve evidence and escalate; do not weaken assertions or conceal required skips.

## Owner planning/autopilot amendment — 2026-10-14 UTC

Owner directs: "Start planning and go autopilot again" and asks that automation
not inflate validation again. This supersedes the specification-only/planning-
dispatch hold above. Claude may prepare a concise Plan at xhigh using the merged,
source-reviewed G009 state while its staging qualification finishes. Record the
exact baseline and reconcile any subsequent G009 correction/final acceptance.

Codex may autonomously review the Plan, request bounded corrections and, after
final G009 ACCEPT plus explicit G010 Plan ACCEPT/input binding, release Claude
implementation at high. Continue the B5 implementation/review/correction and
authorized MR/integration/closeout loop without another routine owner start
prompt. Architecture, security-model, scope, budget and permission stops remain.
No implementation is authorized merely by this planning release.

Validation is part of the bounded outcome, not a parallel framework project.
The Plan must distinguish new composition guarantees from inherited coverage,
justify every additional validation mechanism and state expected run time/profile.
Use existing entry points and fixtures wherever they can prove the cases ran;
do not add a new orchestration layer, repeated qualification campaign or duplicate
control/oracle suite. Apply the Spec's validation limits below, preserving all ACs.

See planning start authorization (not included)
for the single planner's dispatch/ownership and native-path exception.
