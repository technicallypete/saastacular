> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G015 — Dynamic capability exposure and stale-context handling

Epic: EPIC-005 — MCP control plane

Status: Definition ready; automatic provisional planning at G014 implementation MR creation; execution gated

Owner: Codex Goal/Spec/review; Claude Code Plan/implementation/validation

Date: 2026-10-15 UTC

## Objective

Make MCP discovery reflect each logical connection's current Organization/Group
authority, while every invocation independently enforces current permissions.
Selection changes, revocation and concurrent requests must never turn a cached
catalog into authority or share selection between independent connections.

## In Scope

- Permission-filtered `tools/list` over G014's five approved control tools and a
  reusable, closed capability-filtering path using existing core policy.
- Freshness after Group select/clear and committed membership, role, association,
  policy and identity/session/lifecycle changes; stale and hidden-call refusal.
- Connection-local catalog/selection ordering, independent logical connections,
  reconnect/resume and supported catalog-change notification/re-list behavior.
- Focused unit, real SDK HTTP and targeted database-ordering evidence for the
  actual discovery/stale/concurrency delta, with existing audit/privacy controls.
- Test-only in-process capability-mapped tools to demonstrate role-dependent
  filtering without adding a production capability or public tool.

## Out of Scope

Product tools or workflows; new capabilities/roles; Organization switching;
G014 authentication/provider qualification, connection setup or Group-control
redesign; G016 hostname/DNS/TLS lifecycle; new identity vendors, token issuers,
services, event buses, schedulers, Redis authority caches or discovery frameworks;
speculative schema/indexes; broad CI/harness redesign; production performance
targets, idle polling or an instantaneous client-UI update guarantee.

## Dependencies

- G014 integrated and finally ACCEPTED before execution, including its accepted
  transport, five controls, local connection state, WorkOS qualification,
  recovery, expiry and control-mutation ordering contracts. G014 Plan r7 design
  and QP-B provider-stage qualification are accepted; its implementation and final
  Goal acceptance remain pending. Proposed interfaces are not shipped behavior.
- Accepted G004–G012 identity, tenancy, policy/RLS, storage, isolation, audit and
  security controls, plus final G013 recovery as inherited through G014.
- [Work Spec](../specs/G015-dynamic-capability-exposure-spec.md), Architecture
  §52 and §53.9, ADR-003/004/005/006, the MCP connection-context contract and
  foundation AUTHORIZATION/DATABASE-ISOLATION documentation.
- Apply the existing automatic lifecycle: when the G014 implementation MR first opens (including draft), dispatch
  G015 planning provisionally without another owner prompt. Reconcile the Plan
  against final G014 content before execution; preserve any explicit owner stop.

## Completion Definition

Spec AC1–AC10 have attributable source/diff and required validation evidence;
notification observation/delivery limits and concurrency ordering are explicit.
Codex reviews the actual implementation, required integration/staging disposition
and scoped cleanup, then records final ACCEPT. G021's foundation barrier remains
closed; this Goal does not authorize G016 or product implementation.

## Execution Policy

**Architecture/security/public-contract tier.** Owner direction “Do it” authorized
this definition; the owner's subsequent clarification “isn't that our default
automated workflow?” confirms the standing lifecycle applies. There is no extra
G015 owner-approval hold. When the G014 implementation MR first opens (including
a draft), automatically dispatch a concise provisional Claude Plan against the
actual MR candidate/source/target baseline. Do not wait for CI/Duo, ready status
or merge to begin planning; planning-only/placeholder/probe MRs do not qualify. Pending
predecessor acceptance remains an execution gate, not a planning blocker.

After final G014 ACCEPT, automatically reassess the Plan against accepted content
and revise/review material changes as required. Explicit Codex Plan ACCEPT,
every dependency ACCEPT, reconciliation
of actual Goal/Spec and predecessor contracts, accepted Git-first bindings,
budget/resource readiness and a recorded execution release precede implementation.
The Plan may choose bounded filtering/invalidation/notification mechanics; it
cannot change the accepted identity or permission model. Reuse existing controls
and required checks rather than constructing another validation framework.

## Iteration Budget

**B7: seven shared Goal-level execution attempts; checkpoint after attempt 4**,
per BUDGET_POLICY (not included). This class covers authorization and
concurrency-sensitive enforcement, despite the small production tool surface.
Definition/planning spends no implementation attempt. Inner fix/retest loops and
delegates share the parent attempt; corrections retain prior counts. Stop on
exhaustion or repeated material failure; do not spend the allowance by default.

Apply the current model-aware policy at dispatch: Fable 5.1 high for planning and
test corrections, selective xhigh for difficult security/concurrency work;
owner-selected Opus 5.5 planning/test corrections xhigh, implementation/routine
execution high. Verify actual settings or report unknown; no budget reset.

## Context / Scope Budget

One accepted G014 HTTP path, five production controls, one reused two-Organization
fixture with two distinct A connections and one B connection, and five bounded
scenario groups. Reuse existing role/policy expectations and deterministic
ordering controls. No general oracle, new harness, exhaustive permutation
campaign, repeated live-provider/browser qualification or copied G007–G014 suite.

## Suggested Subagents / Delegation

One implementer owns overlapping MCP/policy integration. Up to two bounded
delegates may inspect SDK/filtering behavior or review security/concurrency;
reviewers are read-only by default. Claude owns integration and one shared B7
ledger. Use existing verified coordinator/worktree rules when later dispatched.

## Stop / Escalation Conditions

Any explicit owner stop; unreconciled or materially changed G014 interfaces; identity,
audience, user/client/Organization or liveness checks weakened; control recovery,
clear or logout exceptions lost; stale authority used for an effect; shared
selection/catalog authority; new public tool/capability/role, runtime discovery,
schema/index or infrastructure need; architecture/security conflict; unexpected
query fan-out, scans or contention under PERFORMANCE.md; native permission
refusal; context overrun, repeated failure or exhausted budget.

Escalate owner-owned decisions rather than silently choosing. Do not reopen
G014's WorkOS qualification or claim per-connection credential containment.
No product implementation before final G021 ACCEPT.

## Revision — automatic continuation clarification, 2026-10-15

Supersedes the extra pre-planning hold introduced in the initial definition at
`9396bdc`; it was a supervisor interpretation, not a required new owner gate.
Existing automatic planning begins at G014 integration, with execution gated on
final predecessor ACCEPT and the normal review/input/budget/resource release.
No planner is dispatched before that trigger by this amendment. No G016+ automatic
specification or larger roadmap mandate is added; missing Specs and explicit owner
stops still govern. Ten ACs, technical scope and B7/checkpoint4 are unchanged.

## Revision — earlier automatic planning trigger, 2026-10-15

Owner clarified that planning begins when the predecessor implementation MR is
initially made. This supersedes the integration trigger in the preceding revision.
A verified G014 draft implementation MR is sufficient; no CI/Duo/merge wait.
Planning stays provisional, uses the actual candidate and is reconciled against
final accepted content before execution. Missing Specs, explicit owner stops and
all execution gates remain. No early G015 implementation or G016+ auto-spec scope.
