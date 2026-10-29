> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G016 — Hostname resolution and verified domain lifecycle

Epic: EPIC-006 — Organization domains and branding

Status: Definition ready for provisional planning at G015 implementation MR; implementation gated

Owner: Codex Goal/Spec/review; Claude Code Plan/implementation/validation

Date: 2026-10-15 UTC

## Objective

Resolve registered platform and verified custom hostnames to canonical
Organization UUIDs, with bounded ownership verification and TLS lifecycle,
without turning routing aliases into identity or weakening browser/MCP authority.

## In Scope

- Trusted hostname normalization, explicit proxy/edge trust, registered platform
  subdomains, reserved labels and unique Organization-domain associations.
- Authorized custom-domain claim, ownership challenge, CNAME/configuration check,
  activation, recheck, suspension/removal and a replaceable TLS lifecycle adapter.
- Domain-aware browser entry/authentication and sign-out under the selected,
  reviewed WorkOS callback mechanism; one WorkOS web application per environment.
- Existing MCP endpoint/credential/session agreement under domain lifecycle;
  exact stored host/resource audiences survive no implicit alias rewrite.
- Optional Terraform for shared deployment resources; one provider API adapter
  for tenant hostname lifecycle, with exclusive per-resource management ownership.
- The reviewed generated-baseline configuration transition, domain runbook and
  focused security/lifecycle tests using existing fixtures and CI entry points.

## Out of Scope

Production DNS/ingress/cloud deployment or provider purchase/provisioning without
separate owner authority; a DNS hosting product, generic scheduler or certificate
platform; marketing-site implementation, agency signup, global workspace picker;
G017 branding/tenant landing content; onboarding, product schemas/features; new
identity vendor, token issuer or shared-cookie model; public domain-management
MCP tools or broad administrative UI unless separately added to reviewed scope;
unrelated identity/MCP refactors, test-framework or runner redesign.

## Dependencies

- G002 and G015 finally ACCEPTED/integrated before implementation. Reconcile the
  actual accepted G014/G015 authentication, exact audience/host, connection-state,
  discovery and revocation contracts rather than treating their Plans as delivery.
- Accepted identity, tenancy, policy/RLS, audit, security and recovery foundation;
  Architecture §§3, 50–52 and 60; ADR-003/004/005/006; MCP connection-context
  contract and [G004 redirect-scaling addendum](../specs/G004-tenant-redirect-scaling-addendum.md).
- [Work Spec](../specs/G016-hostname-resolution-domain-lifecycle-spec.md) and its
  unresolved DNS/TLS, browser cross-origin and management-authority decisions.

## Completion Definition

Spec AC1–AC12 have reviewed source/diff and attributable verification evidence.
Required provider/HTTPS/manual evidence, configuration transition, integration
disposition and scoped resource cleanup are complete. Codex records final ACCEPT;
merging or a simulated adapter alone is not production DNS/TLS qualification.
The G021 product-feature barrier remains closed.

## Execution Policy

**Architecture/security/data tier.** Owner authorizes this G016 draft only after
the bounded browser-runner correction; this does not release G016 planning or
implementation or authorize G017+ specification. Earlier G016 stop is lifted
only for this definition. Root supervisor finalizes the draft/status separately.

Before planning release, resolve the listed scope/owner questions or identify
which bounded alternatives the Plan may compare. Before Plan ACCEPT, record the
selected DNS/TLS adapter, browser callback/handoff security contract, management
authority and deployment qualification boundary; obtain required owner/ADR
decisions for new infrastructure/vendors or changed security contracts. No
settings/provider probes or live DNS changes are authorized by this draft.

Implementation requires Plan ACCEPT, all dependency ACCEPTs, current Git-first
binding, decisions/budget/resource readiness and a recorded supervisor release.
Reuse accepted controls; do not create another orchestration framework. Applicable
automatic lifecycle rules remain subordinate to this explicit draft-only hold.

## Iteration Budget

**Provisional B7: seven shared Goal-level execution attempts; checkpoint after
attempt 4**, under BUDGET_POLICY (not included). Domain lifecycle,
authentication/transport trust and migration coupling justify this upper bound;
Claude may recommend a smaller class or bounded split before Plan approval.
Definition/planning uses no implementation attempt. Inner loops/delegates share
the parent attempt; no reset for rebase, interruption or new provider evidence.

Apply the current model-aware effort policy at later dispatch; record actual
configuration or unknown. This draft does not change an active worker's settings.

## Context / Scope Budget

One existing core/web authority path, two platform Organizations plus a third
growth case, one custom-domain lifecycle and six grouped security scenarios.
The Plan limits adapter/provider qualification to concrete unresolved guarantees;
reuse unchanged identity/MCP proofs. Do not duplicate a full stack or full suite
per hostname/case or begin a DNS/TLS/performance benchmark campaign.

## Suggested Subagents / Delegation

One implementer owns overlapping domain/authentication integration. Up to two
bounded delegates may review hostname/DNS/TLS safety and browser/MCP security.
Reviewers are read-only by default; Claude integrates work and retains one B7
ledger. Use the existing coordinator and supported native isolated worktree
mapping; naming limitations alone do not block work.

## Stop / Escalation Conditions

Any owner stop or unreleased phase; unresolved production adapter/custom-domain
authentication decision; unavailable prerequisite evidence; architecture/ADR
conflict; broad-cookie or bearer-in-redirect proposal; unknown/unverified host
fallback; issuer/subject/client namespace change; MCP audience/resource retarget
or local binding weakening; domain collision/reassignment race; unsafe outbound
DNS/TLS work; major provider/infrastructure need; destructive migration or new
capability/public surface outside reviewed scope; unbounded host lookup/recheck
cost, scans or contention; native permission refusal; context overrun, repeated
material failure or exhausted budget. Escalate, preserving work and evidence.

## Authority amendment — EPIC-006 continuation (2026-10-15)

Owner instruction “Spec 17 and continue auto for epic 5 and 6” supersedes the
initial definition-only/preplanning hold above. G016 is included in the exact
G014–G017 delivery set recorded in PROGRESS (not included).
Authorize provisional planning at the initial G015 implementation MR under normal
lifecycle rules; the Plan may compare bounded DNS/TLS/edge, browser-authentication
and management-surface choices. The owner still selects major infrastructure/
security decisions before Plan ACCEPT. No live probe, provider settings change or
production deployment is granted by this amendment. Implementation requires final
dependency ACCEPTs, reviewed Plan/current binding, decisions/budget/resource
readiness and explicit recorded supervisor release. B7/checkpoint4 unchanged.
The mandate stops after G017 closeout, before G018 work.

## Owner amendment — sqlc assessment in G016 planning (2026-10-16)

Owner direction: “So try to introduce in 016 planning?” Include a bounded sqlc
trial assessment in the G016 Plan for one suitable new domain-query package.
The Plan may propose adoption within this Goal, with exact files/tooling/schema
inputs and focused checks; Codex Plan review decides that bounded scope before
implementation. If compatibility or added complexity makes it unsuitable, record
a concrete deferral and continue G016 with pgx. No separate Goal or unconditional
adoption prerequisite is imposed. Existing B7, dependency/owner/release gates and
architecture/security invariants remain; no active G014/G015 tooling change.


## Owner amendment — live qualification split to G057 (2026-10-16)

The owner approved the [G016 → G057 scope/dependency amendment](../results/G016-G057-qualification-split.md).
It supersedes earlier mandatory-in-G016 real-provider/manual qualification and
no-deferral wording only for P-1/P-2: actual WorkOS callback/sign-out scaling and
live custom-host HTTPS/browser proof now belong to
G057 (not included), required before
G021. Mark these guarantees unqualified/transferred, never passed by fixtures.
G016 still requires reviewed implementation, the already-planned provider-test
extension, local security/HTTPS/emulator evidence, applicable MR/Duo and required
post-integration CI, Result and cleanup before final ACCEPT. Unresolved code or
security defects are not deferred by this amendment. P-3 remains optional; P-4
keeps its existing explicit production-adapter hold. No immutable Plan/history
is rewritten, no budget is reset and no broad rerun follows from management edits.
G017–G020 need accepted implementation contracts, not G057 live qualification;
the G021 foundation barrier additionally requires final G057 ACCEPT.
No owner DNS/VM/WorkOS/browser action or G057 dispatch is required/authorized now.


## Owner budget override — optional eighth attempt, 2026-10-17 UTC

Owner instruction: **“Approve an 8 attempt of needed.”** Interpreted as one
additional G016 execution attempt if attempt7 requires a further bounded
correction. The maximum is now **eight total attempts (B7 + one owner override)**,
not eight additional attempts. Attempts1–7 and checkpoint4 history are preserved;
no reset, automatic consumption or new scope is implied.

Attempt7 continues under its existing release/binding. If needed, Codex may review
and release a bounded attempt8 without another owner budget/proceed prompt. Its
specific correction, scope/input binding and minimum relevant checks still need
supervisor disposition before execution. Attempt9 or a wider scope needs new
owner authority. This supersedes earlier statements that another attempt after7
lacks budget, without altering those historical approvals or failed evidence.

No interruption/restart/replan solely for this budget update. Reconcile at the
next safe handoff/batch boundary; preserve any safely running immutable checks.
Required source review, MR/Duo/staging disposition, cleanup/final ACCEPT,
architecture/security/native-permission stops and G017/G057/G018 boundaries remain.


## Owner budget override — optional ninth attempt, 2026-10-17 UTC

Owner: **“Give claude an extra attempt since we should have logs this time.”**
The G016 maximum is now **nine total execution attempts**, one additional slot
beyond the prior cap8. Attempts1–8 and checkpoint4 history remain; no reset.
This supersedes prior no-attempt9 statements for budget only. No attempt10.

If the combined hosted run exposes a failure, Claude may inspect its actual
published diagnostics and propose the smallest evidence-based correction. Codex
may review and release bounded attempt9 within existing G016 scope without another
owner budget prompt. The correction and minimum affected validation still need
supervisor disposition; this budget amendment alone consumes no attempt and
releases no speculative patch or blind full retry. Continue the reviewed MR !38
and current safe checks unchanged; reconcile this management-only amendment at
the next safe boundary without restarting validation.

Retain source/MR/Duo/required staging evidence, cleanup and final ACCEPT,
architecture/security/scope/native stops, and G017/G057/G018/G019 owner boundaries.
