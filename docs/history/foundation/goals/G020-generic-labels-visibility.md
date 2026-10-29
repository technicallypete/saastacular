> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G020 — Generic labels and visibility

> Current phase: FINAL ACCEPT. Implementation merged at `504fe8fc`; Result closeout MR !98 merged at `1f593023`. See the [final Codex acceptance](../results/G020-final-acceptance-codex-review-20261024.md).

Epic: EPIC-008 — Labels and administrative organization

Status: **FINAL ACCEPT** — implementation merged at `504fe8fc`; closeout Result MR !98 merged at `1f593023` on 2026-10-24. Earlier holds below remain as history and are superseded by the dated amendments and final acceptance.

## Objective

Provide reusable administrative labels with independent owner-scope and target
checks, default privacy and current-authority enforcement. Use G020 as the first
normal task exercising the accepted G055 consolidated Restruct workflow.

## In Scope

- Organization- and Group-owned LabelDefinition/LabelAssignment primitives,
  tenant-bound storage, scoped lifecycle and bounded query contracts.
- Organization-owned labels on existing Groups; private to Organization admins/owners;
  Group owner/admin/member roles do not reveal them, with no member-visible opt-in.
- Group-owned definitions and a closed target-validation boundary for future
  Group-owned objects; test-only target fixtures prove assignment semantics.
- Organization owner/admin manages labels on Groups; exact Group owner/admin
  manages labels on its Group objects. Organization authority alone grants no
  Group-owned object-label management path. Visibility-aware reads, RLS, atomic
  mandatory audit, thin HTTP/MCP surfaces and minimal existing-shell UI.
- Focused adversarial, migration, cross-scope, revocation and failure evidence;
  actual CI selector/pin reconciliation and ordinary cleanup.
- A brief Result observation of normal G055 discovery/roles/handoffs/cleanup;
  reuse its accepted qualification, without another workflow test framework.

## Out of Scope

Product objects (Ideas, Projects, Workstreams and Roadmaps), labeling global users,
Business Profiles or storage objects, public tags, role grants through labels,
bulk import, hierarchy, automation, search engine, colors/taxonomy designer,
label-based access control, new vendor or live deployment. No fake target type
registered in production, G054 revival or G021 barrier waiver.

## Dependencies

- Final ACCEPT/integration of G019, G058 and G055; accepted G018 creation,
  current tenancy/authorization/RLS, audit, identity and visual foundations.
- [Spec](../specs/G020-generic-labels-visibility-spec.md), Architecture §53.8,
  SECURITY, AUTHORIZATION and applicable accepted ADRs.
- Reconcile the actual accepted predecessors before Plan review/execution;
  no migration number, route/catalog or runtime binding is reserved now.
- G057 live qualification does not block G020; it still gates G021.

## Completion Definition

Spec AC1–AC8 have reviewed source and attributable evidence, required current-head
CI/Duo/staging disposition, scoped resource cleanup and Codex final ACCEPT.
Default-private labels never leak through member reads, listings or aggregates.
Unsupported production target types remain closed, with the extension limit clear.

## Execution Policy

Architecture/security/data tier: new scoped data, RLS, audit and public contracts
require a concise Claude Plan and explicit review/release. Owner's 2026-10-18
“Do that” authorizes this Goal/Spec drafting only. G055 final ACCEPT is required
before Claude planning; the existing stop before G020 planning is not revoked.
Any later planning/execution requires recorded owner phase authority. No active
G019/G058/G055 implementation or immutable Plan is rewritten by this definition.

## Iteration Budget

Provisional B5, checkpoint after attempt3 under BUDGET_POLICY. Reassess at Plan
review against actual schema/adapter/visibility work; definition spends no attempt.
No budget reset or permission to consume all attempts.

## Context / Scope Budget

Two Organizations, two Groups per tenant, existing owner/admin/member roles,
one real target type and isolated synthetic Group-target fixtures. One scoped
implementation writer by default; 300k context checkpoint/350k safe transfer.
Change-impact selection connects changed policy/migration/audit/catalog consumers
to focused checks; no full local umbrella suite by default.

## Suggested Subagents / Delegation

Claude owns one Plan and implementation branch. Optional bounded read-only
security/data review for owner/target visibility; no duplicate implementation or
new coordinator. Codex owns final Plan/source/evidence disposition.

## Stop / Escalation Conditions

Hold before planning until G055 ACCEPT and recorded phase release. Stop affected
work for authority/visibility ambiguity, architecture conflict, unapproved target,
unsafe migration/retention, new dependencies, contract/scope expansion, native
refusal, budget exhaustion or failing isolation/audit guarantees. Do not weaken
oracles, hide private labels, route around native refusals or add a product model
to fill the missing target gap. Resolve major product decisions with the owner.

## G060 dependency reconciliation — 2026-10-19

[G060](G060-explicit-group-authority.md) is a required accepted predecessor before G055
completion and this task's planning. Reconcile its metadata/content separation:
Organization staff may manage Organization-private Group labels under existing
Organization authority; Group-owned object labels still require actual Group
roles and cannot inherit the ordinary Group-administration override. No G020
planning/execution release; existing eight criteria and privacy scope remain.


## Owner authorizes provisional planning now — 2026-10-20

Owner agreed “Ya i agree” to starting G020's provisional Plan now rather than waiting for G055 qualification. This supersedes this document's earlier G055/G060 acceptance-based **planning** hold and specification-only phase stop. Codex releases the verified existing coordinator to assign one planner (reuse any existing G020 planner; do not duplicate it). Read-only repository inspection and Plan/input/history artifacts only: no implementation, migration reservation, installs, builds, DB/browser probes or shared test resources.

Plan against current integrated G060 source and current approved G055 design, marking pending final qualification/discovery/cutover as provisional. Reconcile actual accepted G060/G055 and any later relevant deltas before implementation release; do not infer acceptance from their merges or from this planning permission. Preserve explicit Group authority, private Organization labels and actual Group-role requirements for Group object labels; G060's ordinary management override does not grant object-label access. Use current canonical local-after-merge qualification and permission/resource rules rather than older hosted-staging wording in historical AC8 notes. Keep G055 as the first normal workflow exercise when qualified, without requalifying it here.

G019/G058 are closed. G060/G055 final acceptance, reviewed Plan, exact input/scope/budget/resource readiness and supervisor execution release remain implementation gates. Provisional B5/checkpoint3 unchanged; planning consumes no implementation attempt. No new roadmap range, G061 dispatch, product feature or live deployment authority. Earlier holds remain as history, superseded only for this planning phase.


Current owner context amendment,2026-10-20:150k checkpoint/200k safe handoff supersedes earlier300k/350k task defaults at the next safe boundary. Existing planner/phase/budget/bindings and no-runtime planning restriction remain; no restart or additional review round solely for policy carry-forward. Coordinator `/compact` stays owner-only.


Owner superseding amendment,2026-10-20: Claude/Codex self-compaction at safe points is now authorized through supported native controls; the earlier owner-only rule above is historical. Keep150k/200k assessment, save state, preserve active batches/children/budgets and verify actual completion. If the live interface has no agent-accessible control, report the limitation once; no synthetic input or replacement-session workaround. G055 carries the policy into consolidated guidance/templates/export; no task restart or added validation campaign.


## Owner scheduling amendment — G062 before G020 execution — 2026-10-20

Owner clarified the delivery sequence as G060 → G055 → G062 → G020 → G057 →
G021 → OS001 → OS002. G020's already accepted Plan and technical scope remain
unchanged. G062 final ACCEPT is now a sequence gate before G020 implementation
release; this is owner scheduling direction, not a technical dependency
assessment. Reconcile the accepted G062 source and current G055 integration with
the Plan/bindings before execution release. No Plan re-review or replan is
required solely for this sequence update.

## Owner scheduling amendment — G063 next before G020 — 2026-10-21

Owner now prioritizes [G063](G063-account-mcp-connections.md), the shared
My Account MCP manager and friendly names, next after G062. Current order is
**G062 → G063 → G020 → G057 → G021 → OS001 → OS002**. G063 final ACCEPT
gates G020 integration/final acceptance; isolated overlap needs normal supervisor
release.
This supersedes earlier immediate-G020 sequencing, adds a scheduling gate only,
and changes no G020 requirements or accepted Plan bytes. Reconcile inputs and
readiness before the existing supervisor execution release; retain prior history.


## Owner sequencing amendment — RC7 before G020 — 2026-10-21

The owner supersedes the prior immediate G020 start order and the implementation release's initial dispatch timing with: **G063 final ACCEPT → RC7 full candidate qualification → G020 → combined staging preview → RC8 full candidate qualification → G057 → G021**. The previously recorded G020 implementation release, accepted Plan, scope, budget and S0 requirements remain valid, but implementation dispatch/resumption is queued until RC7 qualification completes. If work has already started, preserve it and pause at a safe boundary; do not discard a checkpoint or spend/reset an attempt. The combined preview follows G020 and covers the then-current staging state, including G062 signup and G063 MCP connections. RC8 is cut from the latest staging commit after that preview. RC7/RC8 promotion to `main` remains subject to the existing owner-controlled promotion authority. G063 final ACCEPT remains required before G020 integration/final acceptance.


## Owner sequencing supersession — G020 implementation before RC7 qualification — 2026-10-21

The owner supersedes the RC7-before-G020-implementation hold: the existing G020 implementation writer is released to resume now, before RC7 full candidate qualification completes. Preserve its S0/checkpoint and B5 budget; do not restart or reset attempts. RC7 qualification remains a separate track and must complete before G020 integrates to staging. G063 final ACCEPT remains required before G020 integration/final acceptance. The staging preview remains after G020 and before RC8, and is not a G063 acceptance gate. The accepted Plan and technical scope remain unchanged; apply the separately approved narrow account-sections allow-list amendment through the existing scope record.


## Owner priority amendment — G064 before G020 resumes — 2026-10-21

The owner now pauses G020 at its safe checkpoint and prioritizes G064 test-isolation Goal/Spec, Claude Plan and reviewed implementation. This supersedes the previous immediate-resumption instruction. Preserve S0, current S1 source/checkpoints (last inspected `origin/g020-impl` `df2c46f4`), safe active tests, worktrees and the existing B5 ledger (one attempt reported); do not start S2 or a replacement writer while held. G020 resumes after G064 final ACCEPT, with supervisor reconciliation of changed harness/Plan inputs, remaining budget and resource readiness. Its accepted label design is unchanged; no automatic replan or attempt reset. G063 final ACCEPT and RC7 qualification still gate G020 staging integration/final acceptance. The combined staging preview remains after G020 and before RC8. G063 frozen qualification is independently owned and is not stopped by this instruction.


## Owner-confirmed delivery order — 2026-10-21

The latest owner order is **G063 → G064 → RC7 → G020 → combined staging preview → RC8 (if needed) → G057 → G021 → RC9**. This supersedes earlier sequencing only. G064 planning and approved isolated implementation may overlap G063's frozen qualification; G063 final ACCEPT precedes G064 staging integration/final acceptance. Both must be accepted before RC7 is cut and fully qualified. G020 remains held through G064 final ACCEPT and RC7 qualification, then resumes its preserved checkpoint/budget after supervisor input/readiness reconciliation. Preview follows G020 acceptance; RC8 is conditional, and every RC actually cut requires its normal complete qualification. G057 and G021 follow before RC9. Existing main-promotion decisions, technical scope and budgets remain unchanged.


## Owner resumes isolated implementation alongside G064 — 2026-10-22

The owner explicitly requested Claude resume G064 and G020. Resume the existing
G020 checkpoint after bounded reconciliation of current inputs, accepted Plan,
scope, budget and resource readiness; isolated implementation may overlap G064
qualification. This supersedes the prior G064/RC implementation-only hold, not
integration gates: G064 final ACCEPT and RC8 qualification still precede G020
integration/final acceptance. Preserve task owners, source/checkpoints and B5
attempt history; report material Plan conflicts rather than silently redesigning.
The current delivery order is G063 accepted → G064 → RC8 → G020 → staging preview
→ RC9 if needed → G057 → G021 → RC10. Running frozen qualification and shared
resource limits remain binding; no later task implementation is released.
