# ADR-004 — Group-owned access to one Business Profile

> **Editorial note (public export).** This historical decision record is retained for its reusable architecture and security reasoning. Links to private planning, review and evidence records that are not part of this repository are marked "(not included)"; references to "Architecture §N" or ARCHITECTURE.md point to the source project's architecture document, which is not included. Goal IDs, dates, approvals and superseded history are kept as written; product-specific passages are redacted where marked "[Editorial redaction: …]". Current behaviour is documented in the foundation guides under [`docs/foundation/`](../foundation/).

Date: 2026-10-08
Status: ACCEPTED — explicit product-owner decision
Decision owner: Product owner
Recorded by: Codex

## Decision

The owner confirms one required Business Profile per completed Group and rejects multiple profiles per Group as unnecessary complexity. Global Beginify/Saastacular users may manage multiple BusinessProfiles, but the UX shows only authorized profiles in their active tenant.

- The tenancy hierarchy remains **Organization → Group**. BusinessProfile is reusable business profile/context, not another tenant/container tier. Group-owned resources do not universally require a BusinessProfile ID or a three-part key.
- Each Group has at most one profile association. A completed/onboarded Group has exactly one. Multiple Groups in the same Organization may reuse one profile and see the same context updates.
- Profiles remain Organization-owned, UUID-identified records. A global user can own/administer Groups attached to different profiles within multiple Organizations. There are no independent BusinessProfile owner/admin grants or writable profile-membership records. This does not make profiles globally shared or visible: reads/lists/selection require active-Organization ownership and current authorization.
- Client-facing onboarding creates a new profile or selects an authorized existing profile. Managing a profile does not grant access to all Groups using it. Existing capability-specific inheritance and direct sensitive authority remain separate.
- A profile can have multiple future BillingProfiles; this does not require multiple Business Profiles on one Group.

## Scope and consequences

G005 supports Group owner/admin/member and atomically creates the creator owner binding with a Group. G006 enforces at-most-one association and shared reuse, exposing read-only derived access facts rather than a new RoleBinding scope, with actual runtime SQL/concurrency tests. G018 enforces exactly one at onboarding completion and owns client-facing creation/selection. Unfinished foundation Groups may have no profile. A future reassignment/removal lifecycle for onboarded Groups needs explicit review; G006 does not add it.

G007/G008 enforce scoped profile access independently of global user identity and preserve sibling-Group isolation. No global profile picker, cross-tenant profile sharing, independently writable BusinessProfile role, public onboarding implementation or third tenancy scope is authorized here. Creator ownership is the explicit Group creation contract, not provider-claim or global authority.

This supersedes the multiple-profiles-per-Group and independent BusinessProfile-grant portions of G006 Spec/Goal r1, provisional Plan r1/review and historical archetype guidance. Preserve those historical bytes; reconcile Claude's Plan before execution. It does not alter G004's bound inputs, current C7 work, predecessor dependencies, budgets or the G021 foundation barrier. No application or Skill implementation is performed by recording this decision.

## Additional owner decisions — 2026-10-08

The owner approves deriving profile management through any attached Group they own/administer and then directs: "The creator should start as owner." These are the revised canonical decisions:

- Store Organization/Group roles once in RoleBinding; Group creator receives `group_owner` atomically. Do not duplicate ownership or update a second BusinessProfile role when Group roles change.
- In the active Organization, Group owner/admin derive management/selection of their attached profile; members derive read only. Organization authority is independently evaluated. Losing the last qualifying path removes management automatically; another qualifying path remains valid.
- Client onboarding creates the Group with its owner, then creates or selects exactly one profile. New-profile creation/attachment is atomic. Existing selection requires a qualifying pre-existing attached-Group or Organization path. No Group role can select an unrelated existing profile merely because it owns the new Group.
- Owner and admin share bounded ordinary Group/profile capabilities. Admin cannot use membership administration to promote itself or modify owner bindings. Public ownership transfer/last-owner/destructive lifecycle remains G019's separately reviewed work; G005 model primitives do not implement that workflow.

G005 provisional Plan r2 and G006 Plan r1 require bounded amendments; previous approval does not authorize superseded contracts. G007/G008 criteria are reconciled to current derived facts. Preserve archived Goal/Spec versions and all Plan/review history. Budgets and G021 remain unchanged; G004 C7 inputs and work are unaffected.

## Owner-approved naming clarification — 2026-10-08

The owner selects canonical `BusinessProfile` in place of `BusinessEntity`, with UI label **Business Profile** and placement **Group settings → Business Profile**. The record is reusable across Groups; `GroupProfile` would wrongly imply exclusive Group ownership. Use `GroupBusinessProfile`/`business_profile_id` and `business_profile.*` for future reviewed domain/interface contracts. Earlier management filenames, archived Plans/Specs/reviews and source Skill names remain historical, with stable Goal IDs and UUIDs. No persisted rename/migration is performed because G006 is not implemented. Cardinality, authority, active-tenant visibility, budgets and G004 bindings are unchanged.

## Owner-approved sibling management UX — 2026-10-08

The owner clarifies that **Groups** and **Business Profiles** are sibling management areas within the active Organization. This supersedes the preceding naming clarification's placement wording as the primary/exclusive management location; **Business Profile** remains the label and `BusinessProfile` the canonical name. A Group's **Edit Business Profile** action opens the same shared profile page as the Business Profiles area, identified by the same canonical profile UUID, rather than a second editor or copied Group-owned profile.

An authorized edit changes the shared record used by its attached Groups. The page and every read/write re-evaluate current active-Organization capabilities, including the ordinary-profile edit policy. The shortcut does not confer authority from navigation, expand selected-Group operations or reveal sibling-Group details. Cardinality, derived roles, G018 ownership of this UX, dependencies, budgets, active G004 bindings and G021 remain unchanged. No UI implementation or exact route contract is authorized by this clarification.

## Later authority refinement — ADR-008, 2026-10-19

[ADR-008](ADR-008-explicit-group-authority.md) supersedes unconditional Organization
staff ordinary-profile/Group management. Metadata read remains; ordinary management
requires a qualifying Group role or the owner-enabled G060 override, subject to
existing edit policy. Neither path implies sibling Group content or owner powers.
Cardinality, shared-record identity and RoleBindings remain unchanged. Implementation
and cross-layer evidence belong to G060, not a historical rewrite of this ADR.
