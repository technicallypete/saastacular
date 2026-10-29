> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G019 RLS oracle r4: source-derived expectations

2026-10-18 UTC. Supervisor clarification of the already accepted G019 Spec R1–R3
and Plan §4.4–4.5/§6.2. An independent read-only reviewer derived this matrix
from Goal/Spec, AUTHORIZATION and accepted Plan/addenda only; it read no
implementation, executable tests, Results or runtime output. This defines
semantic expectations, not observed database behavior or execution permission.
Preserve historical oracle revisions and cite this source for r4.

## Authority

All actors require enabled status and valid current Organization context.
Ordinary targets have admin/member roles. Each positive remains in the selected
Organization and permitted Group, with immutable keys/scope/source unchanged.

| Actor | Ordinary Organization binding | Ordinary Group binding | Organization owner transition | Group owner transition |
| --- | --- | --- | --- | --- |
| Organization owner | Manage | Manage within permitted Group | Same-scope owner rules | Requires target Group owner role |
| Organization admin | Manage | Manage within permitted Group | Deny | Requires target Group owner role |
| Organization member | No administration of others | Requires target Group owner/admin role | Deny | Requires target Group owner role |
| Target Group owner, Organization member | No administration of others | Manage target Group | Deny | Same-scope owner rules |
| Target Group admin, Organization member | No administration of others | Manage target Group | Deny | Deny |
| Target Group member, Organization member | No administration of others | No administration of others | Deny | Deny |
| Foreign/nonmember/invalid context | Deny | Deny | Deny | Deny |

Manage here permits ordinary admin/member role changes and removal; it does not
permit touching an old/new owner, changing immutable identity/scope/source,
invalid roles, foreign tenants or bypassing mandatory audit.

## Self operations and owner invariants

Facade ChangeMemberRole/RemoveMember reject self_target. Dedicated leave handles
departure. Raw database rules are separate: Plan §4.4 own-row USING preserves
locking, while managerial WITH CHECK controls UPDATE. Raw Organization admin
to member and authorized Group admin to member self-demotions are permitted.
Member-to-admin self-promotion remains denied, as does admin-to-owner without
current same-scope owner authority. A cross-Organization member's active-context
Group authority confers no mutation in their other Organization.

Owner transitions require the exact scope's current owner; Organization-wide
ordinary management never overrides Group-owner authority. Promotion requires
enabled existing same-scope admin and, for Group ownership, current Organization
membership. Demotion/deletion of the last assigned owner is refused; multiple
owners are valid. Atomic transfer promotes recipient before removing/demoting
former owner. Disabled last owners remain assigned; emergency disable is immediate.
Organization departure requires Organization-only context and dependent Group
cleanup, preserving every Group's owner and preventing dormant grants.

## Selected-context interpretation

Spec R1 is authoritative: a selected Group cannot target a sibling Group,
including raw self DELETE of that sibling binding. Base Plan §4.4's abbreviated
own-row predicate must not be used to authorize that exception. The reviewer
flagged this shorthand; this clarification resolves it to the existing Spec,
not a new scope exception. Test the composed read/policy/guard boundary.
An ordinary Organization-role update is not a sibling-Group operation; distinguish
it from Organization departure, which expressly requires Organization-only context.

## Required proof and oracle representation

Name the actual target role and expected unchanged/changed state. Add explicit
positive role demotions with readback, not a blanket permission for role changes.
Preserve rollback/state checks. Pin direct runtime negative pairs:
member-to-admin self-promotion; admin-to-owner without owner authority;
Organization manager touching Group ownership without that Group owner role;
foreign tenant and selected sibling mutation; disabled/non-admin/nonmember
promotion recipients; sole-owner demotion/removal; immutable keys/source and
dependent-binding Organization departure. Cover UPDATE and DELETE as planned.

For guard refusals, the mechanism driver must check the named owner_authority,
recipient_ineligible or last_owner constraint as applicable, rather than accepting
any23000 or any42501. Policy-hidden zero rows are distinct from CHECK/guard errors,
but all forbidden paths preserve state and required audit semantics. Exact SQLSTATE
mechanics belong in the driver, separately from source-derived semantic literals.
Rows becoming admitted writes move to explicit positive cases; do not remove
coverage from the old immutable/denial checks without its replacement.

Competing owner changes retain at least one owner. Accepted r2 class40 outcomes
remain unavailable/aborted, whole-transaction rollback, no automatic retry.
Existing attributable tests may supply negative/concurrency coverage; do not
duplicate entire suites when a mapped existing case proves the same boundary.
