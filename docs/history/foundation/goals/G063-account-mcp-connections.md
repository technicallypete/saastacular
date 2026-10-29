> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G063 — Account MCP connections and friendly names

Epic: EPIC-008 — Reusable foundation completion, before G021.
Status: Plan r1 accepted; implementation released under the owner-independence amendment and Codex release record.
Owner: Codex Goal/Spec/review; Claude Code Plan/implementation/validation.
Spec: [G063 Work Spec](../specs/G063-account-mcp-connections-spec.md).

## Objective

Let a signed-in user manage their own current-workspace MCP connections from
My Account, reusing the existing setup interface, and give connections optional
friendly names so the user can distinguish their AI clients. Preserve the existing
tenant-bound connection, client binding and lifecycle model.

## In Scope

- A shared server-rendered connection-management component embedded in `/account`
  and the existing `/connections/mcp` compatibility page, with create/list/revoke
  behavior preserved and page-specific pagination/status/return handling.
- Optional friendly name on creation and rename/clear for an owned live pending
  or active connection in the current Organization; neutral legacy fallback.
- A narrow additive migration, authoritative core validation and mutation,
  least-privilege SQL/RLS/grant review, thin web forms and bounded contracts.
- Focused automated/browser, isolation, migration and lifecycle evidence, updated
  connection documentation, normal source/MR reviews and post-merge qualification.

## Out of Scope

Provider/OAuth/issuer changes or live provider operations; new MCP tools; different
tenant, Group, client, URL or session semantics; per-Group connection provisioning;
global connection administration or tenant discovery; terminal recovery; fabricated
name backfill; a new frontend framework, component library, audit framework or
infrastructure; clipboard enhancements; product-specific Beginify features.

## Dependencies

Accepted G014/G015/G016 and their transitive identity, tenant authorization,
session, recovery and audit controls. Current source contracts are
[ADR-005](../../../adr/ADR-005-tenant-bound-mcp-connections.md),
[connection context](../../../../contracts/mcp/connection-context.md) and
[control plane](../../../../contracts/mcp/control-plane.md).

The owner's 2026-10-21 instruction authorizes this one reusable foundation task
and requires it before SaaStacular completion. Supervisor scheduling places G063
**after G057 and before G021**, preserving **G062 → G020 → G057 → G063 → G021**;
the exact insertion is a scheduling decision, not an owner-selected technical
dependency. G062's active work does not technically gate G063 planning. G063 final
ACCEPT is a prerequisite for G021 and the accepted SaaStacular/OS001 foundation.
No broader Goal/Spec mandate or change to G062's M1 gate follows.

## Completion Definition

Spec AC1–AC8 have attributable evidence against the actual diff and Result.
Focused pre-merge validation, supervisor/source and applicable Duo review, required
hosted fast checks, authorized staging integration, coordinator-owned full local
qualification on the exact merged staging SHA and scoped cleanup are complete.
Codex records final ACCEPT; merge alone does not complete this Goal. G021 retains
its independent foundation review and product implementation barrier.

## Execution Policy

**Architecture/security/data tier: concise explicit Claude Implementation Plan
and Codex Plan ACCEPT required**, because names add persisted metadata/SQL and a
mutation at the session/tenant authorization boundary. At initial publication,
this Goal/Spec did not itself release implementation. Plan r1 is now accepted;
the specific owner override and supervisor execution release are recorded in
G063 implementation release (not included).
All other current-input/scope, reviewed migration/contract, budget, isolated
ownership, resource-readiness and validation gates remain. Preserve existing
active bindings and owner holds.

After release, Claude may implement, validate and make bounded in-scope corrections
autonomously under normal repository rules. Use current model-aware effort defaults
and fallback policy in AGENTS.md/BUDGET_POLICY; verify actual settings or report
unknown. No extra owner manual-feedback gate is introduced. Major product or
architecture decisions remain owner-owned.

## Iteration Budget

**B3: three shared Goal-level execution attempts; checkpoint after attempt 2**
under BUDGET_POLICY (not included). Zero implementation attempts spent.
Planning/reviews and ordinary inner fix/test loops do not each spend an attempt.
Reassess convergence at the checkpoint; no silent extension, reset or weaker ACs.

## Context / Scope Budget

One shared management surface, one optional metadata field and one rename/clear
operation, using existing routes, controls and test harnesses. **150k active-context
checkpoint / 200k compaction-or-safe-handoff threshold**, subject to smaller native
limits. Transfers preserve task, attempt, evidence and bindings. Recommend a split
before implementation if this cannot remain bounded; no general account redesign.

## Suggested Subagents / Delegation

One Claude implementation owner/writer; optional one bounded read-only security/SQL
or independent source review. Avoid overlapping writers or duplicate coordinators.
All delegated work shares this Goal's attempt allowance and existing inventory.

## Stop / Escalation Conditions

Stop affected work for broadened user/tenant authority or unsafe grants; a rename
that changes binding/lifetime/Group state or revives terminal state; arbitrary
redirects; raw name/credential leakage; architecture/ADR conflict; unreviewed
contract/migration expansion; provider or infrastructure changes; unsafe active-task
bindings; native refusal; unavailable required evidence; nonconverging strategy or
budget exhaustion. Preserve work/evidence and request a bounded disposition.

History: 2026-10-21 — bounded owner-requested foundation Goal/Spec published;
planning authorized, implementation release remains pending.

## Owner priority amendment — 2026-10-21

Owner directs: plan this task next and make this simple addition next on the
roadmap. This supersedes the initial supervisor placement after G057 above.
Current order is **G062 → G063 → G020 → G057 → G021 → OS001 → OS002**.
Prioritize the existing G063 planner now; do not create a duplicate or interrupt
G062 validation. Scope, B3 and Plan/readiness/release requirements are unchanged.
G063 final ACCEPT gates G020 integration/final acceptance; isolated overlap needs
the normal supervisor execution release; G020's accepted technical Plan does not need rewriting for order alone.

## Owner execution-independence amendment — 2026-10-21

The owner clarifies that G063 implementation does not depend on G062's remaining
owner M1 walk-through, cleanup or final Goal ACCEPT. G062's code integration and
full local qualification are complete; G063 may implement in its isolated task
checkout while G062 closeout remains pending. Do not disturb or reuse G062's M1
preview stack. G063 final ACCEPT remains required before G020 integration/final
acceptance and before G021; this amendment changes no G063 technical scope,
budget, validation or other release gates.


## Owner sequencing amendment — RC7, G020 and RC8 — 2026-10-21

The current owner order is **G063 final ACCEPT → RC7 full candidate qualification → G020 → combined staging preview → RC8 full candidate qualification → G057 → G021**. This supersedes earlier sequencing text. G063 final ACCEPT is required before RC7 and before G020 integration/final acceptance. The staging preview follows G020 and can review the current G062 signup, G063 MCP connection and G020 label surfaces together. G063 final ACCEPT may be recorded before the preview; the preview is a separate follow-up and is not an acceptance gate. RC8 is cut from latest staging after the preview; candidate qualification does not itself authorize merging its promotion MR to `main`.
