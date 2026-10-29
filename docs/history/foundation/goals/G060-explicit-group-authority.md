> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G060 — Explicit Group authority and owner-controlled administration

Epic: EPIC-007 — Group and Business Profile onboarding.
Status: Plan ACCEPT; S1–S2 core/SQL implementation released, 2026-10-19.
UI composition follows risky-slice review under the recorded release.
Owner decisions: 2026-10-19, recorded in ADR-008. G059 remains deferred.

## Objective

Separate Organization oversight, ordinary Group administration and Group content
access. Explicit Group membership is the default; an Organization owner may opt
into bounded Organization-wide administration without granting content access.
Deliver this foundation increment before G055 completes and before SaaStacular.

## In Scope

- Current Organization owners/admins can list/read same-Organization Group and
  Business Profile metadata, without implicit access to Group objects/content.
- Default-off, Organization-scoped management override, configurable only by a
  current Organization owner, applying to Organization owners and admins equally.
- Explicit Group roles remain authoritative for content and ownership. Reconcile
  existing core policy, SQL/RLS/helpers, object storage, HTTP/MCP and UI consumers.
- Existing invitation-only Group joining, creator-owner bootstrap, ownership
  continuity, revocation and current-context protections remain intact.
- Minimal settings UI, authorized read-only metadata views, truthful management
  controls and automated regressions. Fold in the paused Group members UI work
  only where compatible with this policy; do not merge the old patch wholesale.
- Forward migration, fail-closed defaults, audit, catalogs/pins and adopter docs.

## Out of Scope

Combined Organization+Group admission for nonmembers, new invitation delivery or
email service, public/open/request-access Group enrollment, new roles or profile
role stores, automatic self-join, Group ownership override, emergency recovery
redesign, product objects, new vendors/infrastructure or deployment. No G059 revival,
G020 planning, public export execution, real RC/main promotion or final tag.

## Dependencies

Accepted G018/G019 and integrated reviewed G058 enrollment/identity foundation;
G058 original-scope final closeout is separate and is not blocked by G060 delivery.
Reconcile accepted G005–G009, G011–G015 and current shell/onboarding contracts.
Read [Spec](../specs/G060-explicit-group-authority-spec.md), ADR-008, Architecture,
SECURITY and current roadmap. G060 does not depend on G055 consolidation.

## Completion Definition

Spec AC1–AC9 satisfied by attributable source/automated evidence, required MR
fast/Codex/Duo disposition, post-merge local full qualification and scoped cleanup,
then Codex final ACCEPT. G055 completion and G021/SaaStacular require G060 ACCEPT.
Owner explicitly waives G060 manual review: record owner-waived/not performed,
never manual PASS. No new owner preview or sign-off gate; automated UI proof remains.

## Execution Policy

Architecture/security/data tier. Owner: “make a spec and a goal and a plan ...
work on this next”, plus explicit SaaStacular prerequisite and manual-review waiver.
This authorizes definition and Claude planning now. Implementation requires Codex
Plan ACCEPT, current inputs/file scope/budget/readiness and recorded release.
Preserve current G055 source/Plan/evidence and earlier Goal history. No automatic
budget reset or implementation authorization from this document alone.

## Iteration Budget

Provisional B5, checkpoint after attempt3 under management/BUDGET_POLICY.md.
Zero implementation attempts spent. Reassess at Plan review for cross-layer
feasibility; propose a split if necessary rather than weakening the boundary.
All delegates and corrections share the Goal ledger; planning spends no attempt.

## Context / Scope Budget

One Organization boolean, existing roles and one finite authority matrix; reuse
existing guards, APIs and fixtures. Two tenants and a few Groups suffice for
isolation; no exhaustive role×surface×mode×theme campaign. One writer by default,
300k context checkpoint/350k safe transfer. Private task files use the current
shared home-local override with strict permissions; no repo .local migration.

## Suggested Subagents / Delegation

Claude owns one concise Plan and implementation branch. Prefer one bounded
read-only cross-layer security review of the risky authorization slice before UI
composition. Codex reviews the Plan/diff and evidence; no new persistent coordinator.

## Stop / Escalation Conditions

Stop affected work for authority ambiguity, missing content-vs-metadata separation,
implicit content/owner grant, foreign/sibling exposure, stale-policy admission,
unreviewed destructive migration/public-contract expansion, audit failure, new
vendor, native refusal, exhausted budget or inability to enforce the same boundary
at runtime SQL and HTTP/MCP. Do not “fix” genuine authorization failures by changing
oracles alone. Preserve G021, G020 phase stop and OS001 pause.


## Owner amendment — invalid-invitation presentation, 2026-10-19

Owner approved Option B after the 1–1.5h combined incremental estimate: fold the
minimal Organization and Group invitation message fix into G060's current UI work.
For the Organization callback's existing invitation_mismatch status, show a fixed
“Invalid invitation link” message, optionally suggesting switching accounts or
requesting a new link. For the Group invitation page, malformed IDs and core's
closed not_found (wrong recipient/nonexistent/foreign) render the same generic
message locally. Never disclose the intended account/email or distinguish whether
the invitation exists. Preserve sign-in/recipient/verified-identity checks and all
successful-invite behavior. Other authentication, unavailable and context errors
keep their existing handling; do not mislabel infrastructure failures as invalid
invitations. Accept POST refusal, callback redirect/routes and core remain unchanged.

Exact added paths: web/lib/enrollment-client.ts,
web/tests/unit/enrollment-client.test.mjs, web/app/invitations/[id]/page.tsx.
web/tests/browser/membership.spec.ts is already approved. Required evidence:
message unit assertion and a focused Group wrong-account browser case, with the
same generic presentation for nonexistent/malformed IDs and existing intended-
recipient success retained. Reuse current browser/build setup and relevant checks;
no duplicate campaign. No new role, endpoint, provider or authorization/error-code
contract. Direct presentation amendment, no separate Plan cycle. Preserve accepted
Plan history; same in-flight attempt4 and budget, no reset. A context handoff alone
spends no new attempt; a later post-submission correction follows normal counting.


## Owner budget extension — 2026-10-20

Owner: “Ectend budget by 2”, referring to active G060 after its final original
attempt and MR59 submission. Add exactly two execution attempts: total cap **7**,
**5 spent / 2 remaining**. Attempts6 and7 are available for necessary bounded
corrections under normal supervisor review/release; no counter reset or automatic
attempt8. Original B5/checkpoint3 and attempt history remain; checkpoint3 was
completed and current source is pre-integration accepted. This explicit cap
override does not require repeating old checkpoints or restarting workers.
Scope39, approved Plan, native permissions, MR/Duo/fast, exact-merged-SHA full local
qualification and final cleanup/ACCEPT remain required. No other Goal budget or
new architecture/product scope is changed.
