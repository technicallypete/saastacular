> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G060 — Explicit Group authority and owner-controlled administration — Implementation Plan (r0)

Goal: [G060](../goals/G060-explicit-group-authority.md) · Spec: [G060 Spec](../specs/G060-explicit-group-authority-spec.md)
(SHA-256 `21f4cf89fa95c31b39220271790549b4f0e88ec9f98b4903031ec9761f6d805e`) · Decision: [ADR-008](../../../adr/ADR-008-explicit-group-authority.md).
Status: **Plan r0 submitted for Codex review; planning only, zero execution attempts spent.**
Planner: task G060, role Plan writer, Fable 5.1 high (runtime effort not verifiable from inside the agent);
branch `g060-plan` from `origin/staging` `70f61393ce1bb27451c0a500de2df10d15889fe3`; worktree
`<source-checkout>` (native `agent-<id>` name).
Inputs: G060-planning-inputs.json (not included). Line numbers below are at `70f61393`.
Manual review: **OWNER-WAIVED, not performed** (Goal Completion Definition); automated evidence remains mandatory.

## 0. Gates and bindings

- Implementation needs Codex Plan ACCEPT, an imported scope allow-list (§11), current input bindings and recorded
  execution release. Architecture/security/data tier: every D-item in §10 is flagged for explicit review.
- Base at execution: `origin/staging` at release time (rebind; `70f61393` is the planning base). G058 original-scope
  closeout runs separately and does not block G060; G055 completion, G021 and OS001 wait for G060 final ACCEPT.
- Not in scope (Goal Out of Scope): combined Organization+Group admission, new roles/stores, public/open Group joining,
  Group-ownership override, emergency recovery, products, vendors, deployment, G059, G020 planning, RC/main/tag actions.

## 1. Approach

Today one Go local (`orgAdmin`, `core/internal/tenancy/access/policy.go:248`, `:360`, `objects_policy.go:77`) and one SQL
macro (`MGR` = `app.rls_org_role() IN ('org_owner','org_admin')`) grant every Organization owner/admin implicit Group
management and Group content. G060 keeps that path **only for metadata reads**, makes ordinary Group management depend on
an explicit Group owner/admin binding **or** a new default-false Organization boolean that only an Organization owner can
set, and makes Group content depend on an explicit Group role under either setting. The change is a finite predicate
rewrite at three enforcement layers (core evaluator, runtime SQL/RLS/DEFINER helpers, storage mapping) plus one additive
column, one audited owner-only write, one HTTP read/write pair, two Group DTO hints and a truthful UI. No new capability,
role or store; the capability catalog stays at 19.

Riskiest slice first (§8 S1–S2): the predicate rewrite with its database parity and the owner-only setting write are
proven by unit oracles, runtime RLS/attack suites and ordering tests before any web change.

### 1.1 Current authority audit (what changes; file:line at `70f61393`)

| # | Surface | Current shortcut / Group-authority predicate | G060 disposition |
|---|---|---|---|
| A1 | `policy.go:272-273` `group.read` | `orgAdmin \|\| roles[g].Valid()` | keep (metadata read); stop using it for object reads (A9) |
| A2 | `policy.go:274-276` `group.update`, `group.manage_members`, `group.complete_onboarding` | `orgAdmin \|\| gr∈{owner,admin}` | `gr∈{owner,admin} \|\| orgMgr` where `orgMgr = orgAdmin && Policy.GroupManagementOverride` |
| A3 | `policy.go:277-281` `business_profile.profile.read` | `if orgAdmin {allow}` then linked path | keep for read (ordinary profile fields are metadata, Spec R1) |
| A4 | `policy.go:277-281` `profile.update`, `select_for_onboarding` | `if orgAdmin {allow}` (bypasses the `group_admins_edit_business_profiles` flag) | `orgMgr` or linked path; the flag applies to the Group path only (unchanged) — D-5 |
| A5 | `policy.go:262-263` `organization.manage_policy`, `audit.read`, `privileged_settings.manage`, `organization.manage_owners` | `orgRole == OrgOwner` | unchanged; the override write reuses `manage_policy` (already owner-only) through a distinct operation — D-3 |
| A6 | `policy.go:266-271` `group.manage_owners`, `group.leave` | own binding only | unchanged (already G060-shaped) |
| A7 | `policy.go:344-399` `eligible` (MCP discovery hint) | `:369-373` orgAdmin ⇒ every Group/profile capability | metadata capabilities via `orgAdmin`; management via `orgMgr`; privileged unchanged |
| A8 | `policy.go:430-452` `contextCheck` (Group selectability) `:441-443`; `00006:265-274` `rls_org_role()` selected-Group validity; `access/mcp.go:348` `manager` list | Organization owner/admin may select/list any Group | keep (selection is metadata; it confers nothing once A9/A10 hold) — D-6 |
| A9 | `objects_policy.go:55-60` read ⇒ `group.read`; `:77-96` put `orgAdmin \|\| role`, delete `orgAdmin \|\| owner/admin \|\| creator-member` | Organization staff read/upload/delete/sign Group objects | Group object read/put: `roles[g].Valid()`; delete: `gr∈{owner,admin} \|\| (member && creator)`; Organization objects unchanged |
| A10 | `00007:227-243` `stored_objects_select/insert/update` Group branch | `MGR \|\| rls_group_role(...)` | explicit `rls_group_role` only (both settings) |
| A11 | `00006:422-427` `role_bindings_select` Group branch; `00015:924-931` `role_bindings_update/delete` Group branch | `MGR \|\| GMGR` | `GMGR \|\| (MGR AND OVR)` (rosters are administration, Spec R1) |
| A12 | `00006:439-441` `groups_update` WITH CHECK | `MGR \|\| GMGR` | `GMGR \|\| (MGR AND OVR)`; `groups_select` unchanged (metadata) |
| A13 | `00006:445-454` `business_profiles_select/update`; `00014:195-203` `business_profiles_insert`, `group_business_profiles_insert` | `MGR` branches | select unchanged; update/insert/attach: `(MGR AND OVR)` or the existing Group path |
| A14 | `00015:909-918` `invitations_select/insert/update` Group branch | `MGR \|\| GMGR` | `GMGR \|\| (MGR AND OVR)`; insert additionally `intended_user_id IS DISTINCT FROM app.ctx_user()` — D-4 |
| A15 | `00015:548` `rls_owner_count` (group), `:626` `rls_group_invitation_candidates`, `:707-709` `rls_user_enabled_lock` Group admission | `org_owner/org_admin OR …` | `(MGR AND OVR) OR GMGR` (owner count: `OVR` or any own Group role) |
| A16 | `00015:780-787` `invitation_claim_group` stale-issuer check | Organization owner/admin counts as current Group-invitation issuer | `(org owner/admin AND override currently true) OR Group owner/admin of the scope`, read under an Organization row share lock — §1.5 |
| A17 | Owner-only settings: `facade.go:193-210` `SetOrganizationPolicy` (`manage_policy`; `00006:434-435` `organizations_update` CHECK `org_owner`); `00016:1138-1163` enrollment mode (MGR policy + owner guard trigger) | pattern | reused for the new column: owner-only in Go, a guard trigger in SQL (permissive UPDATE policies on `organizations` would otherwise admit an admin) — D-2 |
| A18 | Self-escalation: `IssueInvitation` (`access/invitations.go:201-257`) does not refuse a self-targeted recipient; with A2 an Organization admin can today invite itself into a Group and accept (A16 holds) | the only Organization-role path into explicit Group authority | refused in Go and SQL (A14) — D-4; `ChangeMemberRole`/`RemoveMember` already refuse `self_target`; `group.manage_owners` is own-owner only |
| A19 | HTTP (`membership/api/http.go:37-53`, `onboarding/api/http.go:37-45`, `enrollment/api/http.go:48-57`) | every Group member/invitation route inherits A2 through `membership.go:88-107`; no route exposes `SetOrganizationPolicy` | inherit the new predicates; add the override read/write route pair (§5.3); Group DTO gains hints (§5.2) |
| A20 | Metadata DTOs: Group `id,name,slug,business_profile_id,complete` (`onboarding/api/http.go:370-377`); profile `id,display_name,legal_name,description,industry,website,can_update` (`:496-499`); MCP `group_id,slug,name` | no counts, previews, URLs, rosters or tokens | metadata; unchanged except the two Group hints. Member rows, invitations (and `token`), candidates = administration; `ObjectInfo`/signed URLs = content (not HTTP-exposed) |
| A21 | Stale flags: core computes per request; MCP caches nothing (`mcp/catalog.go:17`, `access/mcp.go:190-220`); web renders per request; `app.mcp_connections.selected_group_id` persists but every call re-runs `contextCheck` | none | prove with AC6/AC8 tests; no cache invalidation work |
| A22 | Audit: `organization.policy_changed` capture (`00008:425-431`, trigger `00011:527-530`) with reasons `group_admin_profile_edit_enabled/disabled`; mandatory atomic append precedent `SignedGetURL` (`objects.go:313-318`) | pattern | reused: two new reasons, one new operation, one new core action — D-8 |

Fixture subjects (`infra/identity/mcp-seed.sql`, `infra/identity/emulator-seed.yaml`): carol = Organization A owner
with no binding in Design/Research; frank = A admin with none; Grace Admin = A admin + Research admin (no login);
bob = Research owner (plain Organization member); Dana = Design owner (no login); dave = Research member (and Group-only
in C); alice = plain A member; Organization B/C Groups are foreign to A. Two tenants and four Groups suffice.

### 1.2 Authority matrix (R2 pin; the finite set)

Caller without a Group binding, in Organization A, `OVR` = `organizations.group_management_override`:

| Capability / action | org_owner, OVR off | org_admin, OVR off | org_owner/admin, OVR on | org_member |
|---|---|---|---|---|
| `group.read`, `business_profile.profile.read`, Group list, Group selection (MCP) | Y | Y | Y | own Groups only |
| `group.update`, `group.complete_onboarding`, `group.manage_members`, `profile.update`, `select_for_onboarding`, invitation issue/list/revoke/candidates | — | — | Y (bounded) | — |
| `group.manage_owners`, owner-role invitations, last-owner bypass, Leave, self card | — | — | — | — |
| Group object read/metadata/sign/put/delete | — | — | — | — |
| Set `group_management_override` | Y | — | owner only | — |
| Organization-scoped capabilities (`organization.*`, `audit.read`, `privileged_settings`, `group.create`, Organization objects) | unchanged | unchanged | unchanged | unchanged |

A caller **with** a Group binding gets exactly that role's existing powers unioned with the rows above. An Organization
admin who is a Group member keeps member-level content; `can_manage` is true only through the override. Turning the
override off removes only the override row. Foreign/selected-other Groups stay `not_found`/`target_outside_context`.

### 1.3 Decision inputs and linearization

`Facts.Policy` gains `GroupManagementOverride`, scanned with the Organization row by the same statement
(`core/internal/tenancy/facts.go:116`, `sqlFactOrganization{,Share,Exclusive}` in `store.go`). A guarded write already
locks that row `FOR SHARE` at step 3 of the documented order (AUTHORIZATION.md L369-384); the policy write takes it
`FOR NO KEY UPDATE` up front. The override is therefore read under the lock that a concurrent disable must wait for, so
every override-based write either commits before the disable or observes it and denies (the existing linearization
argument, unchanged). Guarded reads use one REPEATABLE READ snapshot. No new lock, order change or advisory lock.

### 1.4 Override write (D-2, D-3)

`SetGroupManagementOverride(tc, enabled bool)` in `core/internal/tenancy/access/authority.go`: guard with
`capOrganizationManagePolicy`, `Target{Kind: kindOrganization}`, `policyWriteQuery` (same shape as
`SetOrganizationPolicy`, `facade.go:131-136, 193-210`), absolute assignment, operation `organization.set_group_override`,
returns the view. `GetGroupAuthority(tc)` reads `{override_enabled, can_change}` under `organization.read` (`can_change`
= decide `manage_policy`, a hint). In SQL the runtime gets `UPDATE (group_management_override)`, admitted by the existing
permissive `organizations_*` UPDATE policies, and the owner-owned SECURITY INVOKER guard
`app.organizations_group_override_guard` (BEFORE UPDATE OF the column) raises `group_override_owner` unless
`app.rls_org_role() = 'org_owner'` or the session is a superuser operator — the 00016 enrollment-guard pattern. A direct
runtime statement by an admin therefore meets the same refusal (AC3). Missing/invalid/unavailable policy: the column is
`NOT NULL`, the reader coalesces to `false`, a read error is `unavailable`, never an allow.

### 1.5 Invitations under the override (A14–A16, D-4)

Issue: `IssueInvitation` refuses `recipient == tc.subject` (`invalid_input`, field `self_target`) in every scope; the
`invitations_insert` policy repeats it. Consume: `invitation_claim_group` locks the invitation row (as today), then the
Organization row `FOR SHARE` (new; claim writer gains `SELECT (id, group_management_override)` on `organizations` under a
`organizations_claim_select` policy `USING (id = app.ctx_org())`), then the issuer/recipient bindings `FOR SHARE`; the
issuer check becomes `(organization owner/admin AND override true under that lock) OR Group owner/admin of the scope`.
Order: invitation → organization → bindings; the policy write holds only users/sessions/organizations and never an
invitation row, and guarded writes hold `organizations` `FOR SHARE` (compatible), so no new wait cycle. An invitation
issued solely under the override is `issuer_authority_lost` (conflict) after the override is disabled or the issuer is
demoted (AC6). Owner-role invitations remain impossible (`InvitationRoleValid`, CHECK `00015:163-167`).

### 1.6 Web and BFF

- Group DTO hints drive the UI; `SelfMembership` keeps its contract (`not_found` without a binding; `self:true`
  asserted by `web/lib/membership-client.ts:238-249`) — an override is never a self membership.
- `/groups/[id]/members`: `getGroup` first; `membersView` takes `authority`: `administers = can_manage`; `SelfCard`,
  Leave and last-owner notice only when `self_role` is set (then `selfMembership` is read as today); owner controls only
  from core's per-row hints (already false for non-owners). `management_source === "organization_override"` renders
  `data-testid="override-admin-notice"` ("You administer this Group's members through the Organization setting
  Group administration by Organization staff. You are not a member of this Group.") and no "your role" wording. A
  readable Group without management shows `data-testid="members-not-administered"` (no table, no form) — D-10; absent,
  foreign or unreadable Groups keep the fixed 404.
- `/groups/[id]`: the Members link only when `can_manage || self_role`; no content navigation exists in web.
- Settings: `web/app/organization/group-authority/page.tsx` + `action/route.ts` and
  `web/components/membership/group-authority-form.tsx`, patterned on `web/components/enrollment/policy-form.tsx`
  (native radio fieldset Off/On, `aria-describedby` scope text, `PendingSubmit`, closed `?status=` map); owners get the
  form, admins the read-only effective status (`data-testid="group-authority-readonly"`), others 404. Scope text states:
  applies to all Organization owners and admins; ordinary member/invitation/metadata administration only; no content,
  ownership or self-join; delegated administrators can grant access to other people. Links from `/organization/members`
  (owner/admin) and `/account` (owner/admin). Light/dark, keyboard and S-ERRORS conventions as the membership projects.
- Fold of `5f94a705` (§10 Q-fold): reuse its own-binding/administers split, the stripped owner controls and its negative
  oracles (alice, absent Group, carol/frank on Organization B's Operations → 404); replace its `listMembers` probe trigger
  with the core `authority` fact; add the override notice; keep the override-off negatives; rewrite E-MEM-3 L279-287.
  Re-implemented by hand in S3; no cherry-pick.

## 2. Files and modules

**Create**

- `core/migrations/00017_group_authority.sql` (§3).
- `core/internal/tenancy/access/authority.go`, `authority_test.go`, `authority_integration_test.go` (override
  read/write, AUT-* and O-OVR-* scenarios; the integration file may instead extend `membership_integration_test.go` —
  implementer's choice recorded in the Result).
- `web/app/organization/group-authority/page.tsx`, `web/app/organization/group-authority/action/route.ts`,
  `web/components/membership/group-authority-form.tsx`, `web/tests/unit/group-authority-forms.test.mjs`.
- `management/v0/results/G060-explicit-group-authority-result.md`; `management/v0/results/G060-scope-allowlist.txt`
  (Codex import of §11).

**Modify — core**

- `access/policy.go` (A2, A4, A7, header comment "nineteen" unchanged), `policy_test.go`, `matrix_test.go` (oracle rows
  with orgAdmin cells: S06, S15, S20, S25, S45, L1257-1267, L1304-1311; `TestToolEligibility`),
  `objects_policy.go` (A9), `objects_test.go`, `objects_expected_test.go`, `facts.go`/`tenancy/facts.go`/`tenancy/store.go`
  (policy column in the Organization fact statements and `OrganizationPolicy`), `tenancy/boundary_test.go` (statement
  text pins), `access/facade.go` (no behavior change; `SetOrganizationPolicy` unchanged), `access/invitations.go` (D-4),
  `tenancy/invitations.go` (self-target refusal before SQL), `access/errors.go` + `tenancy/errors.go`
  (`group_override_owner` → forbidden/`owner_authority`; `self_target` for invitations), `access/log.go`
  (`opSetGroupOverride`, decision reason `allowed_override`), `access/audit.go` (mandatory override record, D-8),
  `access/mcp.go` (no change expected; `eligible` consumers), `core/internal/audit/audit.go` + `audit_test.go`
  (reasons 68→70, operations 64→65, core actions 6→7, parity source 00017), `rls_expected_test.go`,
  `attack_expected_test.go`, `membership_integration_test.go`, `invitations_integration_test.go`,
  `ordering_integration_test.go`, `onboarding_integration_test.go`, `integration_test.go` (`TestAccessPolicy`),
  `audit_integration_test.go`, `access/mcp` tests (`TestMCPDiscovery`), `membership_test.go` (`TestMemberHints`,
  `TestMembershipCapabilities`).
- `core/internal/onboarding/api/http.go` (+ tests): Group JSON hints. `core/internal/membership/api/http.go` (+ tests):
  the two `organization/group-authority` routes, `TestMembershipRoutesPinned`. `core/cmd/server/main.go` only if the
  route registration list changes (pinned by B-MCP-SOURCE; expected: membership `Register` covers it, no main.go edit).
- `core/migrations/migrations_test.go` (file list), `core/internal/migrate/runner_test.go` (17 sources),
  `core/internal/migrate/integration_test.go` (version-17 catalog, `current_version 17`),
  `migrate/membership_integration_test.go` (policy inventory, predicate fragments, DEFINER catalog +1 reader),
  `migrate/enrollment_integration_test.go` (organizations policy count, attribute body md5 if replaced),
  `migrate/policy_integration_test.go` (column set), `migrate/objects_integration_test.go` (policy expressions),
  `migrate/onboarding_integration_test.go` (00014 policy predicates if restated), `migrate/audit` catalog pins.
- `core/internal/mcp` integration tests (`TestMCPDiscovery` orgAdmin eligibility, `TestMCPTransport` version pin).

**Modify — web**: `web/lib/onboarding-client.ts` (Group DTO `hasExactly` + byte budget), `web/lib/membership-client.ts`
(authority read/write functions, route helper), `web/components/membership/members-view.tsx`,
`web/app/groups/[id]/page.tsx`, `web/app/groups/[id]/members/page.tsx`, `web/app/organization/members/page.tsx` and
`web/app/account/page.tsx` (link), `web/tests/unit/{onboarding-client,membership-client,membership-forms}.test.mjs`,
`web/tests/browser/membership.spec.ts`.

**Modify — harness/docs**: `infra/validation/run.sh` (route pins L189-195/L653/L897/L910, B-MIGRATE-CATALOG versions
16→17 comments/expectations, record strings), `infra/validation/lib.sh` L644 (E-MEM-8, E-MEM-9), docs in §9.
`infra/validation/required-checks.tsv`, `profile-paths.tsv`, `.gitlab-ci.yml`: no change.

## 3. Schema and migration — `core/migrations/00017_group_authority.sql`

Slot verified from the repository: `ls core/migrations` ends at `00016_organization_enrollment.sql`; `00017` is free.
`00006`–`00009` are SHA-pinned (`migrations_test.go:58-80`), so every policy/function change is a `DROP POLICY`/`CREATE
POLICY` or `CREATE OR REPLACE FUNCTION` in 00017 (the 00015/00016 pattern). Up, in one transaction:

1. Precondition `DO`: exact version-16 shapes (column absent, the 00015/00016 policy names and bodies present, the
   claim writer's function inventory, FORCE on the touched tables).
2. `ALTER TABLE app.organizations ADD COLUMN group_management_override boolean NOT NULL DEFAULT false` — every existing
   and new Organization reads `false`; no backfill, no inference from old behavior. `GRANT UPDATE
   (group_management_override) ON app.organizations TO <db>_runtime`; column grant postcondition restated.
3. Reader `app.rls_group_management_override(p_org uuid) RETURNS boolean` (policy-reader-owned, STABLE, SECURITY
   DEFINER, `search_path = pg_catalog, pg_temp`, `coalesce(false)`; EXECUTE to runtime, owner, claim writer; PUBLIC
   revoked). Macro `OVR` = `app.rls_group_management_override(app.ctx_org())`.
4. Guard `app.organizations_group_override_guard()` (owner-owned SECURITY INVOKER, `BEFORE UPDATE OF
   group_management_override`): raises `group_override_owner` for a session role other than `org_owner`.
5. Policies re-created with the §1.1 predicates: `role_bindings_select`, `role_bindings_update`, `role_bindings_delete`,
   `groups_update`, `group_business_profiles_insert`, `business_profiles_update`, `business_profiles_insert`,
   `stored_objects_select`, `stored_objects_insert`, `stored_objects_update`, `invitations_select`, `invitations_insert`,
   `invitations_update`; new `organizations_claim_select` (claim writer) + `GRANT SELECT (id,
   group_management_override) ON app.organizations TO <db>_claim_writer`.
6. `CREATE OR REPLACE` under `SET LOCAL ROLE`: `rls_owner_count`, `rls_group_invitation_candidates` (policy reader),
   `rls_user_enabled_lock`, `invitation_claim_group` (claim writer; §1.5). `rls_org_role`, `rls_group_role`,
   `groups_select`, `business_profiles_select`, `group_business_profiles_select`, `groups_insert`, bootstrap policies:
   unchanged (D-6).
7. Audit: `audit.capture_organizations` replaced to record `organization.policy_changed` with reason
   `group_management_override_enabled|disabled` when the column changes; `events_reason_valid` +2,
   `events_operation_valid` +1 (`organization.set_group_override`), `events_action_valid` +1 core action
   `access.override_used` with its `events_attribute` shape (actor user, Organization, Group target, operation; no
   reason/token) — D-8. One validating scan of `audit.events` per widened constraint.
8. Postcondition `DO`: the column and grants, the reader/guard ownership/ACL/volatility, the 14 policy bodies (deparsed
   fragments), the claim writer's function inventory (4), the catalog constraints; relation count unchanged (71).

Down (development/test only): refuses while any Organization has the override `true` or `audit.events` holds a G060
reason/operation/action (history is never discarded); otherwise restores the 00015/00016 bodies and policies, drops
the guard, reader, claim policy and column, and revokes the grants. Rollback limit: the column values are destroyed by a
down; production rollback is "deploy the previous code with 00017 applied" (the extra column is inert to version-16
code: it reads neither the column nor the reader; the stricter policies then deny Organization staff Group management,
which is fail-closed, not an escalation). Restore from a pre-00017 backup: migration re-applies, default `false`.
Restore from a post-00017 backup: values retained; `infra/recovery` replay scripts untouched. Locks: brief ACCESS
EXCLUSIVE on `organizations` (constant default, no rewrite) and on the re-created policies' tables; migrate in a quiescent
window, then deploy. No data migration, no destructive step, no edit to accepted migrations.

## 4. Contracts and interfaces

### 4.1 Go (`core/internal/tenancy/access`)

- `OrganizationPolicy{GroupAdminsEditBusinessProfiles, GroupManagementOverride bool}`.
- `decide` (§1.2); `decideObject` (A9); `eligible` (A7); `IssueInvitation` self-target refusal; new
  `SetGroupManagementOverride`, `GetGroupAuthority`, `GroupAuthorityView{OverrideEnabled, CanChange bool}`.
- `GroupOnboarding` (or its HTTP projection) gains `SelfRole tenancy.Role` (zero when none), `CanManage bool`,
  `ManagementSource` (`group_role` | `organization_override` | none), computed from the same facts (two extra `decide`
  calls, no extra statement).
- Audit: operation `organization.set_group_override`; core action `access.override_used` appended on the guarded write's
  transaction whenever the allow path was `orgMgr` (`ChangeMemberRole`, `RemoveMember`, `IssueInvitation`,
  `RevokeInvitation`, `RenameGroup`/`ChangeGroupSlug`, onboarding completion, `UpdateBusinessProfile`, profile select);
  an append failure refuses the write (`unavailable`), as `SignedGetURL`. Reads (`ListMembers`, `ListInvitations`,
  candidates) record the decision reason `allowed_override` in the `access.decision` log record only.
- Refusals: `group_override_owner` → `forbidden` (`owner_authority`); `self_target` → `invalid_input`;
  `issuer_authority_lost` unchanged (`conflict`).

### 4.2 HTTP (`/internal/v1`, web BFF only; no public API or MCP wire change)

- `GET /internal/v1/groups`, `GET /internal/v1/groups/{id}`: rows gain `self_role` (string|null), `can_manage` (bool),
  `management_source` (`"group_role"` | `"organization_override"` | null). Exact key set becomes
  `id,name,slug,business_profile_id,complete,self_role,can_manage,management_source`.
- `GET /internal/v1/organization/group-authority` → `{"override_enabled": bool, "can_change": bool}` (any member;
  `organization.read`). `POST /internal/v1/organization/group-authority` body `{origin, csrf_token, enabled}` →
  200 same view; 403 `Not permitted` for non-owners; CSRF/origin as the enrollment policy route. Registered by the
  membership handler (`TestMembershipRoutesPinned`, B-TENANCY-HTTP probe list, run.sh route pins).
- `GET /internal/v1/members/self`: unchanged contract; `can_manage` now false for an Organization admin who is a
  Group member unless the override is on.
- Everything else: unchanged shapes; refusals shift per §1.2.

### 4.3 MCP

Control tools unchanged; `context.list_groups`/`switch_group` keep listing/selecting for Organization staff (D-6);
`tools/list` eligibility for capability-mapped (test-only) tools follows A7. No `tools.listChanged`, no new tool.

### 4.4 Web

`web/lib/onboarding-client.ts` `GroupView` + three fields; `web/lib/membership-client.ts` `groupAuthority()`,
`setGroupAuthority()`, `groupAuthorityActionRoute`; members view and pages per §1.6.

## 5. Validation and test mapping

### 5.1 Existing coverage inspected (reuse unless stated)

Unit: `matrix_test.go` oracle (S-rows), `policy_test.go`, `objects_test.go`, `membership_test.go`, `audit_test.go`
(`TestCatalogParity`), `tenancy/boundary_test.go`, `migrations_test.go`, `migrate/runner_test.go`, HTTP tests of
`membership/api` and `onboarding/api`, `mcp` unit tests, web unit tests. Integration: `TestAccessPolicy`,
`TestRLSIsolation` (+`rls_expected_test.go`), `TestIsolationAttacks` (+`attack_expected_test.go`), `TestMembership`
(MEM-*), invitations subtests, `TestObjects` (+`objects_expected_test.go`), `TestOnboarding`, `TestAccessOrdering`,
`TestRLSOrdering`, `TestAudit`, `TestMCPDiscovery`, migrate catalogs (`TestMigrationLifecycle`, `TestRLSCatalog`,
`TestPolicyCatalog`, `TestObjectsCatalog`, `TestOnboardingCatalog`, `TestMembershipCatalog`, `TestEnrollmentCatalog`,
`TestAuditCatalog`). Browser: `membership.spec.ts` E-MEM-1..7, `onboarding.spec.ts` E-ONB-3/4, `enrollment.spec.ts`
E-ENR-1/2/5 (settings pattern).

### 5.2 Scenario groups → tests (reuse / extend / new; the distinct boundary of each new case)

| Group | Tests | Kind | Distinct boundary |
|---|---|---|---|
| T1 evaluator matrix | `TestOracleMatrix`/`TestOracleTables`/`TestOracleGates` rows rewritten for §1.2; `TestPolicyGates` | extend | orgAdmin cells → metadata-only; `orgMgr` cells; override-on/off is a new oracle input column |
| T2 object mapping | `TestObjectStoragePolicyMapping`, `TestObjectPolicyOracle`, `objectsExpectedCases` | extend | Group object read/put/delete deny Organization staff under both settings; member/owner positives kept |
| T3 runtime RLS | `TestRLSIsolation` (`rlsExpectedCases`, `Inserts`, `Mutations`), `TestIsolationAttacks` | extend | roster/invitation/object/profile rows for carol/frank with OVR off and on; selected-Group shape; direct UPDATE of the override as admin → guard refusal (AC3 SQL) |
| T4 owner-only write | new `AUT-1..6` (owner enable/disable + audit reasons; admin/Group owner/member refusal in Go; direct runtime SQL refusal; coalesce/unavailable fail-closed; absolute no-op set) | new (`authority_integration_test.go`) | the distinct operation/guard did not exist |
| T5 bounded enabled mode | `TestMembership` MEM-OVR-1..4: frank with OVR on manages Design members/invitations, rename, onboarding; cannot invite self (D-4), promote/demote/transfer, leave, read objects; `SelfMembership` not_found; `can_manage` false for a member-admin with OVR off | extend | override path and its exclusions are new |
| T6 stale issuer / revocation | `TestAccessOrdering` O-OVR-1 (disable racing an override-based removal: serialized, coherent audit), O-OVR-2 (invitation issued under OVR, OVR disabled → claim `issuer_authority_lost`), O-OVR-3 (issuer demoted); `TestRLSOrdering` OR-3 pattern reused | new subtests in existing suites | the override is a new mutable fact in the linearization argument |
| T7 migration/catalog | `TestMigrationLifecycle` v17, `TestMembershipCatalog` (+reader, +policy bodies, claim inventory 4), `TestEnrollmentCatalog` (organizations policies 4→5), `TestPolicyCatalog` (columns), `TestObjectsCatalog` (expressions), `TestAuditCatalog`/`TestCatalogParity` (counts) | extend | version-17 delta and down refusal |
| T8 audit | `TestAudit` AT-OVR: `organization.policy_changed` reasons; `access.override_used` atomic with the mutation; append fault refuses; unknown-commit classification reused | extend | new action/reasons |
| T9 HTTP/BFF | `TestMembershipHTTP`/`TestMembershipRoutesPinned`/`TestMembershipBodyShapes` (+2 routes), `TestOnboardingHTTP` (Group hints), web unit `onboarding-client`, `membership-client`, `group-authority-forms` | extend/new | contract shapes |
| T10 MCP parity | `TestToolEligibility`, `TestMCPDiscovery` (orgAdmin hint under OVR off/on; selection kept) | extend | discovery truthfulness after a policy change |
| T11 browser | E-MEM-3 (L279-287 rewritten), E-MEM-5 (override-off: carol/frank on Design → `members-not-administered`, no table/form/self-card; alice/absent/foreign → 404), **E-MEM-8** new: settings page — carol enables by keyboard, frank read-only, alice 404, forged POST as frank 403, audit-visible outcome; **E-MEM-9** new: with OVR on frank administers Design (table, invitation form, `override-admin-notice`, no self card/Leave/owner controls), invites alice, alice joins; carol disables; frank's next visit is `members-not-administered`; alice's membership survives; axe both themes | extend + new | UI truthfulness for the override state; `lib.sh:644` gains E-MEM-8/9 |

### 5.3 Batches (focused executable selectors; working directory `core/`; fetch/rebase/reconcile before each)

- **b1 unit (no DB, no harness lock, `-mod=readonly GOPROXY=off`, disposable GOCACHE, network-disabled container):**
  `go test ./internal/tenancy/access -run '^(TestPolicyGates|TestPolicyScope|TestOracleTables|TestOracleMatrix|TestOracleScope|TestOracleManageable|TestOracleGates|TestOracleCrossOrganization|TestToolEligibility|TestObjectStoragePolicyMapping|TestObjectPolicyOracle|TestMembershipCapabilities|TestMemberHints|TestMembershipFailure|TestAccessBoundary|TestRLSBoundary|TestAuthority.*)$'`;
  `go test ./internal/tenancy -run '^(TestSQLBoundary|TestMigrationParity|TestMembershipRefusalCodes|TestInvitationRejectsBeforeSQL|TestInvitationSpecValidate)$'`;
  `go test ./internal/audit -run '^(TestCatalogParity|TestStatements)$'`;
  `go test ./migrations -run '^(TestMigrationSet|TestEveryMigrationIsVersionedAndReversible)$'`;
  `go test ./internal/migrate -run '^(TestNormalSetIsTheReviewedSet|TestProtectionTargetsExactObjects)$'`;
  `go test ./internal/membership/api -run '^(TestMembershipHTTP|TestMembershipRefusalMapping|TestMembershipBodyShapes|TestMembershipRoutesPinned|TestMembershipCopyIsExact)$'`;
  `go test ./internal/onboarding/api -run '^(TestOnboardingHTTP|TestOnboardingBodyBudgets|TestOnboardingCopyIsExact)$'`;
  `go test ./internal/mcp -run '^(TestToolCatalog|TestPackageBoundary)$'`.
- **b2 runtime PostgreSQL (harness lock; `rls_go` pattern, `-tags integration`):**
  `./internal/migrate -run '^(TestMigrationLifecycle|TestMembershipCatalog|TestEnrollmentCatalog|TestPolicyCatalog|TestObjectsCatalog|TestOnboardingCatalog|TestAuditCatalog)$'` (first run `-v` lists subcases; thereafter only changed-input subcases by name, recorded in the Result);
  `./internal/tenancy/access -run '^(TestAccessPolicy|TestRLSIsolation|TestIsolationAttacks|TestMembership|TestObjects|TestOnboarding|TestAudit|TestAuthority.*)$'` (whole `TestRLSIsolation`/`TestIsolationAttacks`/`TestObjects` are justified: policy bodies change);
  `./internal/tenancy/access -run '^(TestAccessOrdering|TestRLSOrdering)$/^(O-OVR|O-MEM|OR-3|E-PG-3)'`; `./internal/mcp -run '^(TestMCPDiscovery|TestMCPTransport)$'`.
- **b3 web unit:** `npm run test:unit`.
- **b4 browser (identity mode):** the `IDENTITY.md` `identity:browser` line with `--project=membership-light
  --project=membership-dark --grep 'E-MEM|S-LANDMARKS|S-NAV'`; E-ONB-3/E-ONB-4 only if the profile hint/oracle is touched.
- **Required final local coverage before the MR (affected union, changed-input provenance in the Result):** B-UNIT-CORE,
  B-UNIT-WEB, B-MIGRATE-CATALOG/IT, B-POLICY-IT/ORDERING, B-RLS-CATALOG/IT/ORDERING, B-OBJ-IT/ORDERING,
  B-ONBOARDING-IT, B-MEMBERSHIP-CATALOG/IT, B-ENROLLMENT-CATALOG, B-AUDIT-CATALOG, B-MCP-IT, B-MCP-SOURCE,
  B-TENANCY-HTTP, B-IDENTITY-E2E (membership projects), B-SCOPE, B-REQUIRED-CHECKS. No default `full` locally.
  Hosted: MR fast automatically; after the authorized conditional merge the existing coordinator owns the exact
  merged-SHA local full qualification (AC9); no automatic staging cloud run.

## 6. Delegation

One implementation writer (Opus 5.5 high; test corrections Fable 5.1 high per CLAUDE.md), single write owner, 300k
checkpoint / 350k transfer, with a phase handoff between S2 and S3 if needed. One bounded **read-only** security
review subagent after S2 (scope: the §1.1 rows, the 00017 bodies, the claim function, the self-escalation and stale
paths), findings consolidated by the writer. No delegated product/authority decisions, no overlapping edits; the writer
validates and integrates everything. Planning spent no attempt.

## 7. Implementation sequence

- **S1 (risky slice, core + SQL):** 00017 (§3), `OrganizationPolicy`, `decide`/`decideObject`/`eligible`, fact
  statements and boundary pins, oracle rewrites (T1–T3), migrate catalog pins (T7); b1 then b2 (minus ordering).
- **S2 (write path, invitations, audit, HTTP, MCP):** `SetGroupManagementOverride`/`GetGroupAuthority`, guard refusal
  mapping, D-4, claim function, audit action/reasons/operation, routes and Group hints, T4–T6, T8–T10; b1/b2 reruns of
  changed inputs; security review subagent; checkpoint.
- **S3 (web):** client DTOs, members view, Group page link, settings page, unit tests, E-MEM-3/5/8/9, `lib.sh:644`;
  b3, b4.
- **S4:** docs (§9), harness pins/record strings, Result, final affected coverage, safekeeping push, MR (fast + Codex
  + Duo), conditional merge, coordinator-owned local full, cleanup receipt.

## 8. Acceptance-criteria mapping

| AC | Proof (selector groups from §5) |
|---|---|
| AC1 Metadata visibility | T1 (group.read/profile.read rows), T3 (`groups_select`, `business_profiles_select` unchanged; roster/invitation/object rows denied for carol/frank), T9 Group hints, T11 E-MEM-5/E-MEM-9 (Group page readable, Members link gated), `TestOnboarding` ONB list cases; foreign/selected-other/absent `not_found` retained (E-ONB-4, E-MEM-5, `TestIsolationAttacks`) |
| AC2 Default separation | T7 (`DEFAULT false`, upgraded fixture reads false), T1/T3 OVR-off cells, T5 (Grace Admin / bob keep bounded Group powers; member-admin `can_manage` false), E-MEM-3 rewritten |
| AC3 Owner-only setting | T4 AUT-1..6 (UI/API/forged/generic policy patch — `SetOrganizationPolicy` cannot set it —/direct runtime SQL; invalid/unavailable fail closed), T8 reasons, E-MEM-8 |
| AC4 Bounded enabled mode | T5 MEM-OVR-1..4 (manage metadata/members/invitations; no self binding, self-invite, owner powers, Leave, content), T10, E-MEM-9 |
| AC5 Content isolation | T2 + `TestObjects` (real SQL + facade: read/metadata/sign/put/delete denied for carol/frank under OVR off and on, in Organization-only and selected-Group shapes; bob/dave positives), T3 `stored_objects_*`; shared profile and MCP selection confer none (`TestMCPDiscovery`, `TestIsolationAttacks`) |
| AC6 Live revocation/concurrency | T6 O-OVR-1..3 (serialized disable vs override write; stale invitation consumption; demotion), E-MEM-9 (disable → next request denied; alice's membership survives), A21 (no caches) |
| AC7 Data/audit continuity | T7 (up/down/up, refusal with history, version-17 catalogs, grants/ACLs, relation count 71), T8 (success/failure/unknown-commit via `TestCommitClassification` reuse), existing MEM last-owner/creation/enrollment regressions rerun in b2 |
| AC8 UI and machine surfaces | T11 (settings/metadata/membership, light/dark, keyboard, refusals, axe), T10 (owner/admin/member catalog and execution parity), `TestMembershipHTTP`; manual review OWNER-WAIVED, not performed |
| AC9 Integration/closeout | §5.3 final union, MR fast + Codex + Duo, coordinator-owned merged-SHA local full, failure disposition and cleanup receipt in the Result, docs §9; no duplicate hosted full |

## 9. Adopter docs (supersession notes, not rewrites of accepted history)

`docs/ARCHITECTURE.md` §53.7 (L4517-4521 bullets; L4553-4576 G019 boundary; storage action policy L3172-3184),
`docs/SECURITY.md` L453-463 (target → implemented evidence), `docs/foundation/AUTHORIZATION.md` (L40-52 Resolve note,
matrix L94-125, G019 table L244-263, storage table L356-359), `DATABASE-ISOLATION.md` (policies L129-141, readers
L155+), `TENANCY.md` (L292-343), `AUDIT.md` (catalog counts, 00017 paragraph), `OBJECT-STORAGE.md` (L18-32),
`MCP.md` (selection note), `WEB-SHELL.md` (members page/settings, browser project row), `MIGRATIONS.md` (row 17),
`IDENTITY.md` (fixture narrative L225-237 if wording implies admin Group authority), `docs/DECISIONS.md` (ADR-008 row
status after ACCEPT, Codex-owned).

## 10. Decisions, risks and escalations (flagged for explicit review)

- **D-1 Column, not a policy map.** One boolean on `app.organizations` (00005/00016 precedent). Risk: none beyond pins.
- **D-2 Owner-only enforcement in SQL by guard trigger**, because the permissive `organizations_update_slug` /
  `organizations_enrollment_update` policies (MGR) would admit an admin's UPDATE of a runtime-granted column.
- **D-3 Reuse `organization.manage_policy` (already owner-only) with a distinct operation `organization.set_group_override`**
  rather than a new capability (catalog stays 19) or widening `SetOrganizationPolicy`. Alternative with fewer pins:
  reuse operation `organization.set_policy` and add the field to `SetOrganizationPolicy` — rejected because the Spec asks
  for a distinct checked operation/field; Codex may select it.
- **D-4 Universal self-targeted invitation refusal** (all scopes), Go and SQL. Small behavior change for Group/Organization
  owners (self-invite was a no-op `already_member` path). Closes the only Organization-role route into explicit Group
  authority (A18).
- **D-5 Profile paths:** `profile.update` and `select_for_onboarding` on the Organization path require the override;
  `profile.read` stays metadata; `ListBusinessProfiles` keeps the selectable set, so without the override an
  Organization staff member's "My Business Profiles" page lists only Groups they manage (possibly empty); unlinked
  profiles (fixture "Operations Partner") are reachable only with the override or by a future reviewed path.
  **Escalation question:** confirm this reading of Spec R1/R2 against Architecture §53.7 "Organization owner/admin can
  manage profiles"; alternative: treat unlinked profiles as Organization-scoped (orgAdmin without override).
- **D-6 Group selection/listing by Organization staff is retained** (MCP `context.list_groups`/`switch_group`,
  `rls_org_role` selected-Group validity) because selection is metadata and every content/management predicate now
  evaluates the explicit role. Alternative: restrict selection to bound Groups unless the override is on (more MCP
  churn, no security gain once A9–A16 hold). **Question for Codex.**
- **D-7 RLS predicate set** (13 re-created policies, 4 replaced helpers, 1 claim policy, 1 reader). Risk: pin churn in
  `TestMembershipCatalog`/`TestObjectsCatalog`/`TestEnrollmentCatalog`; mechanical, within standing authority.
- **D-8 Audit of override-based administration:** one new core action `access.override_used` appended atomically on the
  mutating transaction (mandatory-write; failure refuses), plus the `organization.policy_changed` reasons. Alternative:
  decision-log reason only (no catalog growth) — weaker than Spec R4 "audit … privileged override-based
  administration". **Question for Codex.**
- **D-9 Group DTO hints** (`self_role`, `can_manage`, `management_source`) on the existing Group read/list; the web
  `hasExactly` pins and byte budgets change. Alternative: a separate authority read per Group (extra round trip).
- **D-10 Members page for a readable Group without management:** a truthful `members-not-administered` notice instead of
  the accepted E-MEM-5 fixed 404 (the Group itself is readable metadata; 404 would be untruthful). This changes an
  accepted G019 oracle and needs explicit sign-off; alternative: keep the 404.
- **D-11 Settings placement:** a new `/organization/group-authority` page (not the enrollment page, whose subject is
  joining); native radio fieldset, no Switch component or dependency (B-DEPS unchanged). Alternative: a second fieldset
  on `/organization/enrollment`.
- **D-12 No split required.** Cross-layer but finite (one column, one matrix); S2→S3 is the natural fresh-worker
  boundary if context nears 300k. Forecast 2–3 attempts of B5.
- **Risks:** (r1) oracle rewrites must not weaken negatives — every removed orgAdmin allow becomes an explicit deny row;
  (r2) `invitation_claim_group` edit is claim-writer code — the security review subagent covers it; (r3) `TestMigrationLifecycle`
  and B-MIGRATE-CATALOG version strings (16→17) are many small pins; (r4) hosted DinD gotchas per memory; (r5) web route
  pins (`run.sh` L653/L897/L910) for the new page. Stops: any implicit content grant found at runtime, a foreign/sibling
  exposure, inability to enforce the same boundary in SQL and HTTP/MCP, a native refusal, budget exhaustion.
- **Fold disposition of `5f94a705`:** see §1.6; no cherry-pick; no test pass inferred.

## 11. Budget, pins and scope allow-list

Budget **B5, checkpoint after attempt 3**, zero execution attempts spent; planning consumed none. Count pins to re-pin at
release against the integrated base: migrations 16→17, `current_version 17`, relations 71 (unchanged), policies on
`organizations` 4→5, claim-writer functions 3→4, policy-reader readers +1, audit actions 45→46 (core 6→7), reasons
68→70, operations 64→65, resource types 11, capabilities 19, membership routes +2 (16→18 method/path pairs),
`BROWSER_MATRIX_IDENTITY_E2E` membership row +2 IDs.

Proposed exact file allow-list (extended regular expressions; Codex imports as `management/v0/results/G060-scope-allowlist.txt`):

```
^core/migrations/00017_group_authority[.]sql$
^core/migrations/migrations_test[.]go$
^core/internal/tenancy/access/authority(_test|_integration_test)?[.]go$
^core/internal/tenancy/access/(policy|policy_test|matrix_test|objects_policy|objects_test|objects_expected_test|objects_integration_test|facade|invitations|membership|membership_test|membership_integration_test|invitations_integration_test|ordering_integration_test|onboarding|onboarding_test|onboarding_integration_test|integration_test|rls_expected_test|rls_integration_test|rls_ordering_integration_test|attack_expected_test|attack_integration_test|audit|audit_integration_test|errors|commit_test|log|mcp|mcp_test|mcp_integration_test|ownership)[.]go$
^core/internal/tenancy/(facts|store|invitations|invitations_test|errors|boundary_test|membership|mcp)[.]go$
^core/internal/audit/(audit|audit_test)[.]go$
^core/internal/migrate/(runner_test|integration_test|membership_integration_test|enrollment_integration_test|policy_integration_test|objects_integration_test|onboarding_integration_test|audit_integration_test|rls_integration_test)[.]go$
^core/internal/membership/api/http(_test)?[.]go$
^core/internal/onboarding/api/http(_test)?[.]go$
^core/internal/mcp/(mcp_test|catalog_fixture_test|mcp_integration_test)[.]go$
^core/cmd/server/main[.]go$
^web/lib/(onboarding-client|membership-client)[.]ts$
^web/components/membership/(members-view|group-authority-form)[.]tsx$
^web/app/groups/\[id\]/(page[.]tsx|members/page[.]tsx)$
^web/app/organization/group-authority/(page[.]tsx|action/route[.]ts)$
^web/app/organization/members/page[.]tsx$
^web/app/account/page[.]tsx$
^web/tests/unit/(onboarding-client|membership-client|membership-forms|group-authority-forms)[.]test[.]mjs$
^web/tests/browser/membership[.]spec[.]ts$
^infra/validation/(run|lib)[.]sh$
^docs/(ARCHITECTURE|SECURITY)[.]md$
^docs/foundation/(AUTHORIZATION|DATABASE-ISOLATION|TENANCY|AUDIT|OBJECT-STORAGE|MCP|WEB-SHELL|MIGRATIONS|IDENTITY)[.]md$
^management/v0/results/G060-(scope-allowlist[.]txt|explicit-group-authority-result[.]md)$
```

23 patterns (every named existing file verified present at `70f61393`). Excluded on purpose: `core/migrations/0000[1-9]*`, `00010`–`00016` (accepted history), `policy-form.tsx`
and enrollment code (pattern only), `web/playwright.config.ts` (no new project), `infra/ci/**`, `.gitlab-ci.yml`,
`required-checks.tsv`, `profile-paths.tsv`, seeds/fixtures (`infra/identity/**`: tests toggle the override at runtime
and restore it), `web/package.json` (no dependency).

## 12. Revision history

- r0, 2026-10-19: initial Plan from the two read-only audits (core/SQL/storage/HTTP/MCP; web/tests/fixtures/5f94a705)
  and the governance inputs listed in `G060-planning-inputs.json`.
