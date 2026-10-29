> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# Group ownership, Business Profile access and onboarding addendum

Date: 2026-10-08. Authority: accepted [ADR-004](../../../adr/ADR-004-group-business-profile-cardinality.md) and ARCHITECTURE §§53.2–53.7. Applies to G005–G008 and future G018/G019 planning; no implementation authorization.

## Canonical outcome

Organization → Group remains the tenancy hierarchy. Every completed Group has exactly one attached Business Profile; multiple same-Organization Groups can share it. A global user can manage multiple profiles, but the UI shows only authorized profiles in the active tenant. BusinessProfile is profile/context, not an extra tenancy level or a required identifier on every Group resource.

Organization and Group RoleBindings are the only stored role authority. Group roles are owner/admin/member. The creator starts as Group owner; BusinessProfile has no independent owner/admin grant, writable membership record or authority-bearing owner column. Profile access derives from current roles and attachments. Removing the last qualifying Group owner/admin path removes management access without another write; a second qualifying Group or Organization path remains independently usable. Member-only reading does not qualify for a management list or existing-profile selection.

## Bounded provisional Plan reconciliation

G005 Spec r2 requires the Group owner vocabulary and atomic Group-plus-creator-owner creation. Preserve its original Spec and design-approved Plan r2/review, then amend only affected creation inputs/transactions, role constraints, projections and fixtures. Unknown creator or a failure between writes leaves neither Group nor owner row; no implicit Organization membership.

G006 Goal/Spec r3 replace independent profile grants and many-to-many cardinality. Preserve Plan r1/review, then amend the Group-unique constraint, singleton lookup/conflict contract, read-only current role/attachment joins and affected tests. No BusinessProfile role scope/source, independent grant methods or additional writable projection. Verify multiple Groups share a profile, another Group has another profile, and service/runtime SQL/concurrent distinct-profile attempts cannot attach a second profile to one Group. Role removal/downgrade and unlink change derived facts immediately without changing the profile or unrelated bindings.

Historical archetype fixtures are superseded where they permit multi-profile Groups or independent BusinessProfile roles. No Skill/template edits or new control framework are authorized by this amendment. Preserve all unrelated accepted constraints and actual predecessor reconciliation.

## G007/G008 inherited verification

The revised G007 matrix has eleven capabilities over six roles. Group owners/admins have the reviewed ordinary Group paths, can select their attached profile, and can edit ordinary profile metadata when the Organization flag permits. Group members can read; privileged profile settings remain Organization-owner authority. Group admin membership administration cannot confer owner-binding/transfer authority. Future detailed ownership lifecycle remains G019.

Use a global user owning/administering Groups attached to two profiles in Organization A and another in B. In active A, management eligibility deduplicates those authorized A profiles and excludes B and member-only profiles. Switching tenants re-evaluates facts; selected-Group context limits inheritance to that Group. Profile access alone reveals no sibling Groups/links/members. No public management-list endpoint is introduced in G007/G008.

After committed role removal/downgrade or unlink, reuse the same login/context/connection: the last lost path cannot retain profile edits/selection/list eligibility. An alternate qualifying path can still allow in the correct context. Ordinary member read can survive a management downgrade. RLS covers the real role/attachment joins or read-only projection, not a cached/materialized profile-membership authority.

## G018 onboarding requirements and verification handoff

These are required inputs to its future full Work Spec, not authorization to implement G018 now:

1. **Creator ownership.** Resolve the creator from trusted identity and active-Organization eligibility; create Group and its owner binding atomically. Forged creator, unauthorized tenant and injected failure leave no partial Group/ownership. Do not infer authority from provider claims.
2. **Incomplete until attached.** The Group is unfinished until exactly one profile attachment is durably committed. Refresh/retry cannot present an unattached Group as complete. Multiple-profile completion is rejected by persistence and the completion boundary.
3. **Create new profile.** Create profile and attach it in one transaction. Inject failure after profile insertion: no orphan profile from that attempt, no attachment or completed state; the existing Group stays unfinished. Success uses one UUID/profile/link, with no independent profile-role grant. Retry must not duplicate the completed result.
4. **Select known profile.** List/select only active-Organization profiles manageable through a pre-existing attached Group owner/admin path or qualifying Organization authority. Test member-only, foreign profile, forged IDs and an unrelated profile selected solely using the new Group's ownership: all deny without changing attachment/completion. Recheck authorization at the committing transaction, including a concurrent loss of the qualifying role/link.
5. **My Business Profiles.** Return distinct manageable profiles in the active Organization, with a shared profile listed once. Remove the user's last management path and verify it disappears from that list; another qualifying attached Group preserves it. Keep read-only member access separate and never reveal sibling-Group details or foreign profiles. No second ownership update is performed.
6. **Sibling management and shared editing.** Provide Groups and Business Profiles as sibling management areas in the active Organization. From a Group, Edit Business Profile opens the same profile page and canonical UUID as the Business Profiles list entry. Verify an authorized metadata edit is visible through both attached Groups without creating another profile or changing attachments. On page load and write, recheck current tenant and capability policy: foreign IDs, revoked authority and disallowed edits deny; profile access reveals no sibling-Group details. Navigation does not expand selected-Group authority. Exact routes/context transitions belong in the future reviewed Plan.

Public Group ownership transfer, last-owner protection and profile reassignment/removal for completed Groups require their separately reviewed lifecycle contract. Do not use an internal unlink to silently leave a completed Group without its required profile.

G005/G006 retain twelve criteria each and B7/B5 respectively; G007/G008 retain fourteen criteria and B7. Existing checkpoints/dependencies, G004 C7 bindings and G021 are unchanged. No application, public UX/API, vendor, hosted action or implementation is authorized here.

## Naming handoff — 2026-10-08

Canonical code/domain name is `BusinessProfile`; UI label remains **Business Profile** under **Group settings → Business Profile**. `GroupBusinessProfile`/`business_profile_id` replace historical association/identifier spellings in future reviewed implementation. Do not introduce a separate `GroupProfile` or rename management artifact paths/Goal IDs. Provisional G005/G006 Plans require bounded reconciliation; archived inputs and active G004 execution stay unchanged.

## UX handoff clarification — 2026-10-08

The owner corrects the preceding naming handoff's placement wording: Business Profiles is a sibling management area to Groups, and Group settings supplies a shortcut to the shared profile page. Requirement 6 above carries this into future G018 scope/verification. There is one reusable profile record/page, with current tenant-scoped authorization, not a separately owned Group profile. No G005–G008 UI scope, application work or active G004 input change is authorized.
