> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G014 — Tenant-bound MCP connections and optional Group selection

Epic: EPIC-005 — MCP control plane

Status: Owner-approved definition; provisional planning authorized, implementation gated

Owner: Codex Goal/Spec/review; Claude Code Plan/implementation/validation

Date: 2026-10-15 UTC

## Objective

Expose the existing generic foundation through authenticated external MCP
connections. Each connection belongs to one canonical Organization UUID and
supports current user/context, authorized Group listing, select/clear and
connection termination. Another Organization requires another authorization and
connection. No product functionality is introduced.

## In Scope

- Actual SDK-backed HTTP transport, registered tenant routing and a reviewed
  external authentication/credential/session lifecycle.
- Immutable endpoint/authorization/Organization agreement, fresh core-owned
  request authority and connection-local optional Group selection.
- The five architecture control-plane tools, their bounded response/error
  contracts and current-session/permission enforcement.
- Minimal connection storage/configuration and additive migrations/audit events
  only when justified by the approved Plan, with existing RLS/role/recovery rules.
- Replace obsolete dormant-MCP guards with authenticated-exposure guards and
  reconcile current reusable guidance/fixtures under ADR-005.
- Focused external-transport/security evidence and applicable existing CI.

## Out of Scope

G015's dynamic domain catalogs/notifications and broader stale/concurrency
matrix; G016's hostname/custom-domain/DNS/TLS lifecycle; product tools, business
onboarding/UI, AI execution/RAG, outgoing MCP integration, machine/service-account
access, a new identity vendor or authorization-server platform, new infrastructure,
production provisioning, automatic cross-tenant discovery/switching, mandatory
MFA, whole-repository refactoring or CI redesign.

## Dependencies

- Final accepted/committed G013 before implementation; its recovery/session and
  maintenance contracts are provisional while this draft is written.
- Accepted G004 identity, G005/G006 tenancy, G007 trusted policy, G008 RLS,
  G009 storage, G011 audit and G012 secrets/security; ADR-003/004/005/006.
- [Work Spec](../specs/G014-tenant-bound-mcp-connections-spec.md), Architecture
  §§3/4/21/37/52/53 and the existing MCP connection-context contract.
- Explicit owner review/release of this definition and its open decisions before
  Claude planning. Creating the files does not satisfy that human-input gate.

## Completion Definition

Spec AC1–AC12 have actual-diff and source-bound evidence, reviewed external
transport/authentication and audit contracts, required local/MR/staging
qualification and scoped cleanup. Codex records final ACCEPT and a G015 handoff.
G021 remains closed; no claim of client compatibility beyond tested scope.

## Execution Policy

**Architecture/security/public-contract tier.** Owner direction “Draft it now”
(2026-10-15) authorizes this Goal/Spec before G013 closes. It changes definition
timing only. **Stop for owner input before G014 planning or implementation**, even
if G013 integrates or a ready Spec exists. Standing automatic transitions do not
release this hold. G013 work and its separate decisions continue unchanged.

After owner planning release, Claude inspects actual accepted predecessors and
writes a concise Plan. Explicit Codex Plan ACCEPT, resolved owner decisions,
final G013 ACCEPT and current Git-first input/budget/resource release precede
implementation. A major new authentication vendor/platform requires its own
owner/ADR decision; do not hide it inside transport plumbing.

After execution release, use ordinary autonomous scoped implementation/debugging
and accepted batch/integration procedures. A local emulator or private bearer
probe is not proof of general external-client interoperability.

## Iteration Budget

**Provisional B7: seven shared Goal-level attempts, checkpoint after attempt 4**,
per BUDGET_POLICY (not included). One attempt contains its focused
implement/fix/retest batches; no per-command or per-agent counters. No attempt
spent by definition. Reassess the budget/split if the selected authentication
flow requires a separate platform; do not expand silently.

Planning xhigh; implementation high unless the owner chooses otherwise. Remind
and record actual setting as verified or unknown at dispatch.

## Context / Scope Budget

One external transport/authentication path, five control tools, one registered
shared-origin route scheme and one two-Organization fixture. Reuse real identity,
policy/RLS/audit helpers. No replicated coordinator, generalized token platform,
new oracle/registry, copied predecessor attack matrix or full stack per negative.

## Suggested Subagents / Delegation

One implementer owns overlapping transport/session/policy/audit changes. Up to
two bounded delegates: protocol/authentication analysis and read-only security
review. Shared B7, actual native worktree/role mapping, one integration owner.
Naming limitations do not block authorized work.

## Stop / Escalation Conditions

Unresolved owner client/authentication choice; G013 reconciliation changes;
endpoint/credential/session mismatch without a fail-closed design; cached
request context or authority; cross-connection Group leakage; unreviewed
issuer/audience/refresh/logout change; audit/recovery omission; new privileged
role/vendor/platform; destructive migration; query/index/fan-out/lock concerns
(PG-006); native permission refusal; scope/context overrun or exhausted budget.

No product implementation before final G021. No G015 work is authorized by this
definition. The owner can retain/extend the planning hold independently of G013.

## Owner definition approval and planning release — 2026-10-15

Owner answered “Ol approve” after reviewing the G014 summary and clarifying that
clearing a Group preserves Organization-level permissions. This approves the
Goal/Spec and releases the original human-input hold **for planning only**.
Connection-local logout remains the definition's default; no implicit global,
other-connection, browser or WorkOS logout is authorized.

The existing coordinator may dispatch one xhigh provisional planner while G013
implements. Use accepted G012 plus the current G013 Plan/binding as provisional
recovery inputs; reconcile actual final G013 before Plan/execution release.
Planning may propose the initial clients, issuer/bootstrap and wire contract;
this delegates proposal preparation, not those unresolved security/vendor
choices or implementation. No new vendor/platform/provider settings or account
actions. Flag decisions requiring the owner during Plan review.

This dated release supersedes the earlier G014 pre-planning hold only. G013 work
continues, final G013 ACCEPT remains an implementation dependency, and explicit
Codex Plan ACCEPT/current input/budget/resource release remain mandatory. No
G014 probes/builds/tests or application changes during planning. The earlier
G011–G013 mandate gains this named planning phase only, not G014 execution or
G015+ specification/work. Owner can interrupt or stop either task.

## Owner UX/route amendment — 2026-10-15

Owner directed recording tenant-hosted per-connection URLs, per-user/Organization
connection tracking, Group switching inside one connection and stable URLs during
ordinary reauthentication. The dated [Spec amendment](../specs/G014-tenant-bound-mcp-connections-spec.md#owner-amendment--tenant-hosted-per-connection-ux-2026-10-15)
supersedes shared-origin-only wording and authorizes a bounded Plan revision for
the WorkOS per-connection proposal. Include minimal authenticated setup; agency
staff and client users retain their existing roles with no special MCP grant.
One user/Organization/AI client is the normal single-connection UX, not a limit.
Revocation/expired pending setup requires a new connection; routine renewal does
not. G016 hostname lifecycle is not pulled forward. Provider qualification and
host readiness, ADR/Plan/input/resource acceptance, B7/checkpoint4 and final G013
ACCEPT remain gates. No provider changes/probe or implementation released here;
G013 F4 remains a separate undecided owner risk.

## Owner selection — managed WorkOS grants, 2026-10-15

Owner answered "Ok" to the recommendation to retain WorkOS for browser login
and MCP token issuance (option B), accepting broader same-user/client credential
reach while preserving local tenant/Group authorization. The dated
[Spec amendment](../specs/G014-tenant-bound-mcp-connections-spec.md#owner-amendment--managed-workos-grants-and-compromise-containment-2026-10-15)
supersedes the prior independent-refresh-family and QP-3 acceptance requirements.
Their failures/history are retained; no dedicated MCP issuer is authorized.

Another Organization still requires another separately authorized local
connection. Separate Group/lifecycle state is retained; distinct URLs do not
promise isolated provider credentials. Normal connection logout remains local.
Before implementation, qualify supported provider user/client-wide revocation,
including old refresh values after fresh provider authorization; no future app
bootstrap is needed. Prove local invalidation/replacement containment during
approved implementation before acceptance.
Provider revocation uncertainty is a release blocker, not a reason to claim local
logout contains the credential or silently waive the gate.

The existing planner may reconcile the bounded Plan/assessment with this decision
and identify supported operations and any remaining provider question. Preserve
r4/history and twelve ACs; B7/checkpoint4 remains provisional pending review.
Planning uses the current workload-model policy (Fable/high default; existing
planner may finish its current phase under its recorded actual setting). No
provider probe/settings mutation, new dependency, implementation or G015 work is
released. Final G013 ACCEPT and revised Plan/input/budget/resource release remain.
