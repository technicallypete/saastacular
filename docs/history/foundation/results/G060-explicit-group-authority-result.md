> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G060 — Explicit Group authority — Implementation Result

Goal: [G060](../goals/G060-explicit-group-authority.md) · Spec: [G060 Spec](../specs/G060-explicit-group-authority-spec.md) ·
Plan: [r0](../plans/G060-explicit-group-authority-plan.md) + r1 (not included) +
correction02 (not included) · Review: G060-plan-codex-review.md (not included).

## Status

**Current (2026-10-19, attempt 5 of 5, final): attempt-5 corrections, Option B and b3/b4 browser evidence
review-ready for Codex pre-integration review, not accepted — see
[§ Attempt 5](#attempt-5-b5-attempt-5-final--2026-10-19).** Earlier status lines are preserved as history.

(Attempt 2) **Corrected S2 checkpoint review-ready, not accepted — see
[§ S2 correction attempt 2](#s2-correction-attempt-2-b5-attempt-2-checkpoint-3--2026-10-19).** The attempt-1 status
below is preserved as history.

(Attempt 1) **S2 checkpoint submitted for Codex early-risk review (review-ready). Not accepted.** S1–S2 core/SQL code, migration
00017 and unit/static evidence are complete; **b2 (database) evidence is NOT RUN** (native refusal, below). Two scope
escalations (§Escalations) block a fully green b1. S3 (UI) not started. Manual review: OWNER-WAIVED, not performed.

## Worker record

- Task G060, role S1–S2 implementation writer (single writer), model Opus 5.5, intended effort high; runtime effort not
  verifiable from inside the agent (unknown).
- Worktree `<source-checkout>` (native
  `agent-<id>`); session start branch `worktree-agent-<id>` at `c91c945b`; implementation branch `g060-impl`
  created at `origin/staging` `0c10fb6d681d9ba9df6ded53a1a350afb783325a` (branch/HEAD verified before any edit; no 00017
  existed). Rebased onto `origin/staging` `72912efadff60d4f4cafbdc15ff3748242c171b1` (management-only: the review record's
  dispatch-verified section and PROGRESS.md) before the first batch.
- Service skill: `skills/saas-foundation/SKILL.md` (database-rls / tenancy-authorization invariants); core has no
  service-local skill file.

## Bound inputs (release `0c10fb6d`; current base `72912efa`, management-only delta)

| Input | SHA-256 |
|---|---|
| Plan r0 | `71d405f41f450bc198c5303338f3cac2481c9234341d3a522b634a8b1eaf3936` |
| r1 addendum | `b3858eccd0524da0f3c69ed6539f3218a70ccbc8289caf1b06ee5f83a1813794` |
| correction02 addendum | `4c44c1e9559f40bc021299f05c822dd9d5a919f54c64f6bb74f65c652e116998` |
| planning inputs r0 / r1 / correction02 | `bcb40f10…17f3` / `b76544fb…1c16` / `669edbe6…f477` |
| scope allow-list (28 patterns) | `e261f70bba43e5810ad76a0f748525c3a0f91ea2f3582687354bc250217065b3` |
| Codex review record at release (latest section governs) | `b5839957a3250f137c4a4e1bdbd8bbd14d46a6cf2307122389c449fc17cccf00` |

Codex mechanical correction applied: the raw own-row self-role UPDATE expects `Guard=self_escalation` / 42501 with the
override OFF as well as ON (own-row USING admits the row; the BEFORE UPDATE guard fires before WITH CHECK); "no role
change" and the independent-issuer upgrade positives (MEM-9 "upgraded", MEM-OVR-8) are kept.

## Budget and attempts

B5, checkpoint after attempt 3. **Attempt 1** started 2026-10-19 (S1); still attempt 1 at this checkpoint. No correction
round yet.

## Implementation summary (S1–S2)

- **Migration `core/migrations/00017_group_authority.sql`** (SHA-256 `f96bfb09…1afe`, 1694 lines, assembled from the
  reviewed pieces with the version-15/16 bodies extracted byte-exact for the down section): precondition (exact v16
  shapes, claim-writer inventory 7, FORCE); `group_management_override boolean NOT NULL DEFAULT false` (no backfill);
  column grants runtime `UPDATE`, claim writer `SELECT`, policy reader `SELECT` (C02-1); owner-owned SECURITY INVOKER
  `app.organizations_group_override_guard` (BEFORE UPDATE OF the column; `group_override_owner`, 42501, unless the
  context role is `org_owner` or a superuser session); inline `OVR` (no new reader); re-created `role_bindings_update`
  (WITH CHECK only: `GMGR OR (MGR AND OVR AND subject <> ctx_user)`), `groups_update`, `group_business_profiles_insert`,
  `business_profiles_update/insert`, `stored_objects_select/insert/update` (Group objects: explicit Group role only),
  `invitations_select/insert/update`; replaced `rls_owner_count`, `rls_group_invitation_candidates` (policy reader),
  `rls_user_enabled_lock`, `invitation_claim_group` (claim writer: issuer = recipient → `not_found`; Organization row
  FOR SHARE, issuer+recipient users FOR SHARE ascending, bindings FOR SHARE; issuer authority = Group owner/admin of the
  scope or Organization owner/admin while OVR is true, and issuer `active`); `role_bindings_authority_guard` step (3b)
  (`current_user IN (rt, ow)`, own Group binding, not Group owner → `self_escalation`); audit actions 45→46
  (`access.override_used`, core), reasons 68→70, operations 64→65, `events_attribute` shape for the new action and the
  two `organization.policy_changed` reasons, `capture_organizations` + `audit_capture_update` watching the column;
  postconditions (column/ACL strings, guard, triggers, inventories, override-reading policies, Group-object policies,
  body fragments, nothing enabled). Down: refuses while any Organization is enabled or any G060 audit row exists, then
  restores v16 (capture, attribution, constraints, policies, bodies, guard) and drops the column last.
- **Go**: `OrganizationPolicy.GroupManagementOverride` read with the Organization fact row (the same FOR SHARE /
  FOR NO KEY UPDATE lock); `decide`: metadata reads via `orgAdmin`, ordinary administration via the explicit Group path
  first, then `orgMgr = orgAdmin && override` (`allowed_override`); `eligible` mirrors it; `decideObject`: Group objects
  need an explicit Group role (read/put/delete) under either setting; new `authority.go` (`GetGroupAuthority`,
  owner-only `SetGroupManagementOverride` under `organization.manage_policy` + `policyWriteQuery`, operation
  `organization.set_group_override`; `overrideUsed` appends `access.override_used` atomically on the write's transaction
  for rename/slug, completions (Group and selected-profile targets), profile update, member role change/removal,
  invitation issue/revoke); Group-scope self-invitation refused before SQL (`self_target`, also in tenancy validation);
  `SetOrganizationPolicy` never sets the override; refusal mapping `self_escalation`/`group_override_owner` → forbidden,
  `self_target` → invalid_input; decision log reason `allowed_override` for override-path allows.
- **HTTP (internal BFF only)**: `GET|POST /internal/v1/organization/group-authority` (`{override_enabled, can_change}`;
  body exactly `origin, csrf_token, enabled`), membership routes 16→18 pairs on 16 paths; Group rows gain `self_role`,
  `can_manage`, `management_source`; Business Profile list rows gain `can_select`, `can_update` and list by
  `profile.read` (C2). No MCP wire change (D-6 retained).

### Deviation D-13 (needs Codex disposition)

`role_bindings_select` keeps its 00006 predicate and `role_bindings_delete`/`role_bindings_update` keep their 00015
USING. r1 C3 already keeps update/delete USING; narrowing `role_bindings_select` as r0 A11 proposed would break the
accepted G019 Organization departure (MEM-13; `SubjectBindingsLock` must read and lock the departing subject's Group
bindings under the actor's RLS, otherwise the Organization DELETE meets `dependent_bindings`). Residual: with the
override off, Organization staff can still read Group rosters and delete ordinary (never owner) Group bindings through
raw runtime SQL; Go denies both (`group.manage_members`). Options: (a) accept; (b) a later reviewed DEFINER departure
path; (c) accept the G019 regression. Consequence noted with C02-2: step (3b) also refuses a runtime Group admin's raw
self-demotion (former r4 "ok" rows for adminA2 and crossUser now expect `self_escalation`); Leave is unchanged.

## Escalations (scope)

1. **`core/internal/tenancy/access/boundary_test.go` (not in the allow-list).** Mechanical pins required by the
   approved interfaces: `authority.go` in `accessFiles`; `type GroupAuthorityView`, `Service.GetGroupAuthority`,
   `Service.SetGroupManagementOverride` in the exported pins; a G060 clause in the SECURITY DEFINER allowance (00017
   re-creates four existing DEFINER bodies, 8 occurrences, all `CREATE OR REPLACE`). Proposed diff (29 lines, not
   applied): `platform-evidence/g060-s12/proposed-access-boundary_test.diff` (SHA-256 `07a071aa…59f0`). Applied only to a
   scratch copy: `TestAccessBoundary` PASS, `TestRLSBoundary` PASS, `TestAuthority*` PASS.
2. **`infra/validation/tools/internal/logrecord/events.json` (not in the allow-list).** The decision log adds operation
   `organization.set_group_override` and reason `allowed_override`; `membership.request` adds `group_authority` and
   `set_group_authority`. Needed before any harness logcheck (B-POLICY-IT/B-TENANCY-HTTP); not run in S2.

## Validation (exact commands and outcomes)

Pre-batch for every batch: `git fetch origin staging`, base `72912efa`, HEAD descendant (`git merge-base
--is-ancestor` rc 0); `bash infra/ci/docker-desktop-mounts.sh` → 15584–15600 of 100000 mounts (15%), rc 0.

**b1** (`platform-evidence/g060-s12/b1.sh`; `docker run --rm --network none`, source `:ro`, module cache volume
`g058-a1-gomod` `:ro`, `GOFLAGS=-mod=readonly GOPROXY=off`, disposable in-container `GOCACHE`, `golang:1.27.2-alpine`;
no lock, no DB, no ports). Selectors exactly Plan §5.3 b1.

| Run | Selector group | Outcome |
|---|---|---|
| 1 | access (17 names) | FAIL: oracle cells (expected before oracle rewrite), `TestAccessBoundary`, `TestRLSBoundary` |
| 1 | tenancy / migrations / migrate / membership api | FAIL (pins: self-target spec fixtures, set 16→17, routes) |
| 1 | audit / onboarding api / mcp | PASS |
| 2 | access | **FAIL only `TestAccessBoundary`, `TestRLSBoundary`** (escalation 1); all oracle/matrix/eligibility/object/authority tests PASS |
| 2 | tenancy (5), audit (2), migrations (2), migrate (2), membership api (5), onboarding api (3), mcp (2) | **PASS** |

Also: `go vet` (default and `-tags integration`) over tenancy, audit, membership, onboarding, mcp, migrations,
migrate: clean (compile proof of the integration tests only — not execution).

**b2: NOT RUN.** Owned resources created then removed (network `g060s12-data` internal, `g060s12-pg`
pgvector/pgvector:0.8.7-pg17 on tmpfs, no published port; removal verified by label). The first test command was refused
natively (verbatim):

> This agent is isolated in the worktree <source-checkout>, but this command runs flock with the text test in a plain command, so what it runs cannot be shown not to be git. Refusing to run it — a worktree-isolated agent's git operations must target its own worktree. Run the plain command from <source-checkout>

Refused operation: `flock -n <home>/.local/state/platform-evidence/local-harness.lock docker run --rm … golang:1.27.2-alpine go test -tags integration … ./internal/migrate/`. Not retried or wrapped. Every b2 selector
(migrate catalogs, `TestAccessPolicy`, `TestRLSIsolation`, `TestIsolationAttacks`, `TestMembership`, `TestObjects`,
`TestOnboarding`, `TestAudit`, `TestAuthority`, ordering subsets, `TestMCPDiscovery`/`TestMCPTransport`) is NOT RUN.
`TestObjects`/`TestIsolationAttacks`/`TestAudit` additionally need the harness MinIO build.

Earlier worktree-guard refusals (non-test, recorded for completeness): "…this command is too complex to verify that it
stays inside the worktree. Refusing to run it — … Split it into plain, separate commands…" (twice: a shell extraction
pipeline and a Python heredoc edit) and "…this command runs docker with the text sh in a plain command…" (an image
inspection). The file extraction/edits were then done by plain `bash <file>`/`python3 <file>` scripts under the private
evidence directory (no git); the image inspection was dropped.

## Tests added or changed (S1–S2)

- Unit (PASS in b1): `authority_test.go` (owner-only capability, C7 hint cases, override reason ordering);
  `matrix_test.go` (four default tables rewritten, two override-on tables, +ovr subjects, 627 asserted cells, scope and
  manageable rows, eligibility literals); `policy_test.go` (G060 subtests); `objects_test.go` / `objects_expected_test.go`
  (Group content denied to staff, oracle R20/R21/P24/P25/P34/D25/D26 → not_found); `audit_test.go` (parity at 00017);
  tenancy `invitations_test.go` (self_target), `boundary_test.go` (87/24 statements); `migrations_test.go`,
  `runner_test.go` (17 sources); membership `http_test.go` (+2 pairs, body shapes); onboarding `http_test.go` (Group keys
  and hints).
- Integration (compiled, **NOT RUN**): `authority_integration_test.go` `TestAuthority` (AUT-1..5, MEM-OVR-1,2,3,5,6,7,8,
  O-OVR-2, O-OVR-4); `migrate/integration_test.go` (v17 lifecycle strings, down 17→16, `groupAuthorityAtVersion17`,
  reset drops the guard); `rls_expected_test.go` mutation flips (Organization staff Group role changes denied; own Group
  role changes `self_escalation`) and `rlsGuardState`.
- **Pending first b2 run (not yet reconciled):** `rlsExpectedCases` Updates/Inserts for Organization staff (groups and
  profiles), r4 override-on mutation rows on a new `bAdminAG1` binding (needs a driver override toggle),
  `TestIsolationAttacks`/`TestOnboarding`/`TestMembership` existing staff expectations, O-OVR-1/3/5 barriers, AT-OVR-1..3,
  `TestMCPDiscovery` staff eligibility, C2 list cases.

## Acceptance criteria (S1–S2 view)

AC1–AC7: code and unit evidence in place; runtime SQL/HTTP evidence NOT RUN. AC8: S3. AC9: S4. No AC is claimed PASS.

## Security review

One bounded read-only review subagent (no edits, no runs) over 00017 up/down and the Go authorization slice. **No
blocker.** Findings and dispositions:

| # | Finding | Severity | Disposition |
|---|---|---|---|
| SR-1 | `role_bindings_select`/`_delete` (and `_update` USING lock) keep MGR without the override: raw-SQL roster reads, row locks and deletion of non-owner Group bindings by Organization staff; no content, object or self-escalation path | should-fix (documented D-13) | **Escalated to Codex** with the reviewer's two narrower options: (i) a bounded SECURITY DEFINER departure body (MGR, Organization-only, non-owner subject; small DELETE grant to its definer role) then narrow both policies to `GMGR OR (MGR AND OVR)`; (ii) Organization DELETE first with guard step (3) as a deferred constraint trigger and MGR admitted only for orphaned Group bindings. Both change the G019 departure sequence. |
| SR-2 | `invitations_insert` lacked an `intended_user_id <> issuer_user_id` check (self-invitation unclaimable anyway) | note | **Fixed**: Group branch adds `intended_user_id IS DISTINCT FROM issuer_user_id`. |
| SR-3 | A raw-SQL claim locked the invitation before the Organization: with a revoke holding the Organization FOR SHARE and a queued override change, a 3-way wait is possible (40P01 abort; liveness only) | note | **Fixed**: `invitation_claim_group` locks the context Organization FOR SHARE first (step 0), the facade's order; step (2b) removed; postcondition pins the order. |
| SR-4 | Any current Organization member could append a forged `access.override_used` through raw SQL | note | **Fixed**: the attribution shape now requires `app.rls_org_role()` owner/admin and the override currently enabled. |
| SR-5 | Self-escalation, privileges/ACLs, audit reasons, live revocation, down byte-identity (bodies diffed against 00015/00016/00011) | — | None found. |

After the fixes (00017 `f96bfb09…1afe`): b1 run 3 (same selectors and container shape) — access: FAIL only
`TestAccessBoundary`, `TestRLSBoundary` (escalation 1); tenancy, audit, migrations, migrate, membership api, onboarding
api, mcp: PASS. Scratch copy with the proposed pin diff: `TestAccessBoundary`, `TestRLSBoundary` PASS.

## Commits

- `d099ae87` G060 S1 WIP (rebased onto `72912efa`); `55807c1a` S1–S2 oracles/tests/Result (pushed to `origin/g060-impl`, no MR); security-review fixes in the following commit (pushed).

## S2 correction attempt 2 (B5 attempt 2, checkpoint 3) — 2026-10-19

**Status: corrected S2 checkpoint submitted for Codex review (review-ready). Not accepted.** Amendment01 and the
release's corrections 1–7 are implemented; the focused b1 selectors and the amendment's b2 selectors PASS on
`999922fb`; `TestAccessPolicy` and `TestOnboarding` FAIL on still-unreconciled G060 staff expectations (pending, below);
`TestMCPDiscovery`/`TestMCPTransport` NOT RUN (out-of-scope version pin). S3 (UI) HELD. Manual review OWNER-WAIVED.

### Worker record (attempt 2)

- Task G060, role S2-correction writer (attempt 2, single writer), model Opus 5.5, intended effort high; runtime effort
  not verifiable from inside the agent (unknown).
- Worktree `<source-checkout>` (native
  `agent-<id>`; session branch `worktree-agent-<id>` at `c91c945b`). `git switch g060-impl` verified at
  `4e668521dad368aed9ed9cdd61706b2abc305af7`; rebased onto `origin/staging` `aa796e639b694455196b527bbec799af3b36731d`
  (management/prose-only delta since `72912efa`). The only conflict was this Result (add/add): staging's copy was
  byte-identical to the source's, so staging's version was kept in each replayed commit (both histories preserved).
  Rebased tree = source tree + staging delta exactly (`git diff 4e668521 0751523c` lists only the 15 staging files).
  Published with `--force-with-lease=g060-impl:4e668521…` → `0751523c`. `origin/staging` ancestor of HEAD verified
  before every batch (staging stayed `aa796e63`).
- Inputs verified: review record `7952637673bfe75c…017bb`, scope allow-list (32 patterns) `5b73b14ed9da6b39…36034`,
  pin diff `07a071aa…59f0`, preview checklist `b88f5ddd…f58f6c`. Amendment01 and review sections read in full.
- Commits: `00f2d50d` (00017 cascade/policies/enabled owner, Go, pins, catalog), `39285dfb` (first b2 fixes, oracles,
  new proofs), `999922fb` (ordering/attack/object/audit reconciliation, AUTHORIZATION.md caveat), then this Result.
  Safekeeping pushes after each round; no MR.

### Corrections 1–7 (what changed, where)

1. **Departure cascade (A1.3)** — `core/migrations/00017_group_authority.sql`: `role_bindings_select`/`_delete` Group
   branch `GMGR OR (MGR AND (OVR OR DEP))`, `_update` USING `GMGR OR (MGR AND OVR)` (WITH CHECK unchanged), `DEP` =
   `pg_trigger_depth() >= 1`; guard step (3) deletes the subject's Group bindings in ascending `group_id` after
   `context_required` (another subject's `group_owner` → 42501 `owner_subject`), then the unchanged dependent check.
   Precondition 11→13 shapes; PC3 ten override-reading policies plus an exact inventory of runtime/owner policies naming
   `pg_trigger_depth()` (deviation D-A2-1); PC4 cascade fragments; down restores `role_bindings_select` (00006) and
   `role_bindings_delete` (00015) beside the existing restores. No new function, DEFINER, grant, trigger or role.
   Go: `departGroups` removed (`access/membership.go`; callers `RemoveMember`, `LeaveScope`, `ownership.go`
   `TransferOwnership`, `retirement.go` `retireDepart`); `tenancy.SubjectBindingsLock`/`sqlSubjectGroupBindingsLock`
   removed; `CodeOwnerSubject` + `privilegeCodes`; `membershipFailure` maps it to forbidden/`owner_subject`;
   `RemoveMember` keeps the recordable `ownerSubject` denial. Pins: `tenancy/boundary_test.go` (86 statements),
   `membership_test.go` mapping rows.
2. **Enabled-owner guard (A2)** — `organizations_group_override_guard`: after the owner-role check, `SELECT us.status …
   FROM app.users us WHERE us.id = app.ctx_user() FOR SHARE`, refuse unless `active` (`group_override_owner`); PC2 pins.
   Tests: AUT-3 reused (active owner positive), **AUT-6** new (facade-promoted second owner, DisableUser, raw UPDATE →
   42501 `group_override_owner`, value unchanged), **OR-6** new (raw owner UPDATE waits behind an in-flight DisableUser
   and is refused after the wait), O-3 reused.
3. **Lock-order caveat (A1.3 correction)** — **O-MEM-4c** new in `TestAccessOrdering`: the facade claim paused after its
   Organization `FOR SHARE` (before step OB), a raw self-departure DELETE of the same subject waits on the Organization
   row (G058 departure trigger), release → wait cycle. Observed: raw DELETE `40P01` victim, claim committed; asserted:
   exactly one class-40 victim, no partial Group/Organization binding, no partial audit, exactly one claim
   `access.decision` (no retry), facade step order O before OB. Caveat stated in `docs/foundation/AUTHORIZATION.md`
   (Deadlocks) and here: raw departure and facade order are NOT the same; this is a new wait edge.
4. **Literal oracles (A1.5)** — `rls_expected_test.go`: four Organization-A staff cases (and, same model, ownerB/adminB:
   deviation D-A2-2) drop the Group labels from RoleBindings, empty GroupMemberships and ProfileAccessPaths; staff Group
   and profile updates/inserts become no_update/denied (r0 A12/A13; the previous Result's pending list); independent
   override-ON cases (driver toggle `rlsContext.Override`: the Organization owner sets the override inside the
   rolled-back transaction) restore the pre-G060 literals. Mutations: :1130/:1132 denied (OFF) with ON positives;
   :1163 `last_owner`; :1164 ok with `Also: [bAdminA2G2]`; :1165 `owner_subject` (OFF and ON — see D-A2-3); :1143-1147
   Guard "" (OFF) with ON `owner_authority`; correction02 `bAdminAG1` fixture and its OFF/ON `self_escalation` rows;
   `rlsGuardState` + `owner_subject`; `r4Mechanism` narrowed; RLS-ORG literal 14→9 plus an ON block.
   `attack_expected_test.go` F1-05 → `0`; F1-09 → `sqlstate 23000 last_owner` (deviation D-A2-4). Membership mapping row
   in `core/internal/tenancy/membership_test.go`.
5. **Transfer to none** — MEM-7 extended: in a second Organization the sole owner who solely owns a Group transfers
   with `none` → `last_owner`, everything rolled back (recipient stays admin, one owner, no role_changed/removed row);
   after the Group succession the same transfer succeeds atomically (recipient owner, former owner holds nothing, three
   `membership.removed` rows).
6. **Earlier pending items** — pin diff applied unchanged (`patch` of the normal-format diff; `TestAccessBoundary`,
   `TestRLSBoundary` PASS on the candidate); `events.json` + `organization.set_group_override`, `allowed_override`,
   `group_authority`, `set_group_authority`; raw own-row self-update OFF = `self_escalation` kept.
7. **Optional preview** — checklist `<home>/.local/state/platform-evidence/g058-a1/r16/G060-preview-checklist.md`
   (SHA-256 `b88f5ddd83afa27dd86ae3a5ba81847f9a0196aa27e0cf1b5b9f3239b2f58f6c`) referenced for the optional post-merge
   owner preview; **not performed** (not a FAIL, not a gate).

Additional proofs (A3): MEM-13 extended (three `membership.removed` rows by the acting owner, `role_binding`, Groups
ascending then the Organization); **MEM-13b** new (audit fault on the Organization row's capture after the nested Group
DELETEs → removal fails, three bindings remain, no removed row, one failed record; control succeeds); **O-MEM-4a**
(promotion first → removal `forbidden:owner_subject`, bindings intact; removal first → promotion `not_found`; the test
also accepts `owner_authority` for the EPQ interleaving, documented in the test); **O-MEM-4b** (claim first → both Group
bindings and the Organization binding depart, three rows; removal first → claim `context_denied`, nothing consumed).

### First-run defects found and fixed (b2 had never executed before)

- 00017 postcondition 1 `IF body IS DISTINCT FROM CASE … END THEN` never parsed (PL/pgSQL ends the IF condition at the
  CASE's first THEN: `syntax error at end of input`) → parenthesized. The organizations trigger inventory pin omitted the
  existing `organizations_keys_immutable`/`organizations_slug_guard`; the stored-object postcondition matched across the
  concatenated USING and WITH CHECK (false positive) → each expression checked separately.
- Version-17 pins: `TestMigrationLifecycle` (`metadataThroughVersion17`, schema-independent trigger definition),
  `setupIT` latest version 16→17, the hook step pattern `O` (+`group_management_override` column; it had stopped
  matching, which broke `failOn "O"`/D2-04).
- Test bugs in the S1 tests: MEM-OVR-6/8 counted fixture grants (now the claim's Group grant / +1 +1 +0 deltas).
- G060 consequences in existing suites (approved model, reconciled with the minimum change and the reason inline):
  E-PG-3b, O-2, O-6, O-ONB-T enable the override for their Organization-path writes; attack E2/E3 Group-object actor
  aOwner → g1's admin gAdmin (g1's owner left orgA at C3-06); object RLS rows RS01/RS02/RS03/RS16/RI09/RT20/RF01/RF02
  and the two Organization-admin guard transitions (now refused by RLS first: `hidden 2`, refused 7→5); AT-FMT positive
  signer → Group admin (admin sign `not_found`); AT-INV enables the override for the profile update (observes
  `access.override_used`).

### Validation (exact; private per-run DB, no lock)

Pre-batch every time: `git fetch origin staging` (`aa796e63`), ancestor check rc 0, `bash infra/ci/docker-desktop-mounts.sh`
→ 15596–15608 of 100000 (15%). Scripts: `platform-evidence/g060-s12/a2/b1.sh` (attempt-1 shape: `--network none`,
source `:ro`, `g058-a1-gomod:ro`, `-mod=readonly GOPROXY=off`, disposable GOCACHE) and `a2/b2.sh` (per run: internal
network `g060a2-<run>-net`, `pgvector/pgvector:0.8.7-pg17` on tmpfs with `infra/postgres/init` `:ro`, optional MinIO,
generated credentials mode 600 under `a2/secrets` deleted at exit, no published port, labels `g060-a2=<run>`, cleanup
by label verified and printed; slot check: at most two running `pgvector/pgvector:0.8.7-pg17` containers — other
repositories' dev databases on `pg16` images are not this repository's runs). Logs `a2/b1-run*.log`, `a2/b2-<run>.log`,
`a2/b2-<run>.pg.log`.

| Batch | Selector | Source | Outcome |
|---|---|---|---|
| b1 run1 | Plan §5.3 b1 (8 commands) | `00f2d50d` | **PASS** all (incl. `TestAccessBoundary`, `TestRLSBoundary`, `TestMembershipRefusalCodes`, `TestMembershipFailure`) |
| b1 run2 | same + `TestMembershipRejectsBeforeSQL` | `999922fb` | **PASS** |
| m1–m5 | `./internal/migrate '^(TestMigrationLifecycle\|TestMembershipCatalog)$'` | `00f2d50d`+fixes | FAIL (the 00017 defects above), fixed in turn |
| m6 | same | working tree → `39285dfb` | **PASS** |
| m7 | `+TestEnrollmentCatalog\|TestPolicyCatalog\|TestObjectsCatalog\|TestOnboardingCatalog\|TestAuditCatalog` | `999922fb` | 6 PASS; `TestEnrollmentCatalog` FAIL: my runner lacked the harness's `ENROLLMENT_REPLAY_SQL_B64` |
| m8 | `'^TestEnrollmentCatalog$'` with the bound `infra/recovery/sql/replay-restrict.sql` | `999922fb` | **PASS** |
| r1–r3, t1, t2 | `TestRLSIsolation` / `+TestMembership\|TestAuthority` | iterations | FAILs as reconciled above |
| o1, o2, a1, t3 | `TestAccessOrdering` subsets, then whole (justified: G060 changes the Organization-path authority its O-2/O-6/O-ONB-T cases use) | iterations | t3 hung after O-2/O-6/O-ONB-T failures leaked a paused writer: test container removed (rc 137), cleanup verified; a1 whole **PASS 40/40** |
| **f1** | `'^(TestRLSIsolation\|TestMembership\|TestAuthority\|TestAccessOrdering)$'` | `999922fb` | **PASS** (97 subtests; RLS-SCOPE 40/0 disagreements; RLS-SCOPE-MUTATE 78 cases; O-MEM-4c raw `40P01` victim) |
| q1, q2 | `'^TestRLSOrdering$/^(OR-3\|OR-6)'` | `39285dfb`+test edits / `999922fb` | q1 **PASS**; q2 **PASS** (OR-3, OR-6) |
| k1–k6 | MinIO suites | iterations | k1 MinIO tmpfs not writable by `nonroot` (fixed `mode=1777`); k2/k3/k5 FAILs reconciled |
| **k7** | `'^(TestIsolationAttacks\|TestObjects\|TestAudit)$'` + private MinIO | `999922fb` | **PASS** (44 subtests; OBJ oracle service 297/0, RLS 85/0; AUDIT-INV 46/46) |
| p1 | `'^(TestAccessPolicy\|TestOnboarding)$'` | `39285dfb`+edits | **FAIL** — pending (below) |

MinIO provenance (each MinIO run): `platform-minio:server-bc184510` = `sha256:c1194d7e…6f4e20b` (label minio
RELEASE.2025-10-15T17-29-55Z; mc RELEASE.2025-08-13T08-35-41Z) and `platform-minio:init-bc184510` =
`sha256:46b1f9f1…5368e7` (mc RELEASE.2025-08-13T08-35-41Z), inspected before use; no rebuild, no substitute; server
`server /data --address :9000 --console-address :9001`, init `infra/minio/init.sh` via `sh -s`.

**Pending / FAIL (not reconciled in this checkpoint):** `TestAccessPolicy` IT-CTX-9, IT-PROFILE-3, IT-FLAG-4,
IT-FACADE-1/2/3/4/6, IT-REVOKE-1/8, IT-LOG-1 and `TestOnboarding` (Organization admin selects an unlinked profile;
C2 profile listing for members/owners) fail on Organization-owner/admin Group and profile writes in the itWorld
fixture (the approved r0 A2/A4 model; same class as the reconciled cases above). **NOT RUN:** `TestMCPDiscovery`,
`TestMCPTransport` — `core/internal/mcp/env_integration_test.go:253` pins `current_version 16` and is outside the
32-pattern scope (escalation E-A2-1); logcheck validate/scan of the access logs (needs the validation tools image, a
shared build); harness IDs (B-*) not run. Also outside scope with a version-16 pin (not in my selectors):
`core/internal/tenancy/access/recovery_integration_test.go:92`, `core/internal/tenancy/integration_test.go:144`.

### Deviations and escalations (attempt 2)

- **D-A2-1** PC3's trigger-depth pin lists the exact runtime/owner policies naming `pg_trigger_depth()`:
  `role_bindings_select`, `role_bindings_delete` plus the accepted 00014 `role_bindings_insert_bootstrap` (`> 0`) and
  00016 `enrollment_restrictions_owner_insert` (owner-only); amendment01 said "in no other rt/ow policy", which the
  existing objects contradict.
- **D-A2-2** ownerB/adminB visible sets changed exactly like the four named Organization-A staff cases (same model; the
  amendment enumerated only Organization A).
- **D-A2-3** :1165 with the override ON is `owner_subject`, not "today's" `dependent_bindings`: the cascade is
  override-independent, so the superseded literal is unreachable; not copied.
- **D-A2-4** F1-09: rb:aMember:gF is gF's sole owner binding (F1-01), so the cascade raises `last_owner` (A1.4), not the
  amendment's "one affected row".
- **E-A2-1 (scope)** `core/internal/mcp/env_integration_test.go` (one-line version pin 16→17) to run the planned
  `TestMCPDiscovery`/`TestMCPTransport` selectors.
- Budget: B5 attempt 2 of 5 (checkpoint 3), not reset. Native refusals (worktree guard, quoted verbatim, not retried;
  the work continued with plain separate commands and script files): "This agent is isolated in the worktree
  <source-checkout>, but this command is too
  complex to verify that it stays inside the worktree. Refusing to run it — a worktree-isolated agent's git operations
  must target its own worktree. Split it into plain, separate commands and run them from
  <source-checkout>" (also its variants for
  `git … ; git …`, a `b2.sh` call with a shell variable, and a `sh -c` with chained `go` commands). No test was marked
  NOT RUN because of a refusal.

### Security review (attempt 2)

One bounded read-only review subagent (no edits, runs or tests) over the 00017 cascade, policies, enabled-owner guard,
down restoration, the Go departure callers and the lock/40P01 work at `999922fb`. **No blocker, no privilege widening.**

| # | Finding | Severity | Disposition |
|---|---|---|---|
| SR2-1 | The `DEP` branch is not reachable from raw runtime SQL: runtime has no TEMP/CREATE/TRIGGER; no trigger body builds caller-supplied SQL; the two invoker bodies touching role_bindings run fixed SQL | note | accepted |
| SR2-2 | The cascade refuses Group-owner subjects (`owner_subject`), keeps owner authority/last owner for self-departure in the nested guard, `context_required` precedes it, any raise rolls back the statement | note | accepted |
| SR2-3 | The loop's `r.role` read is unlocked; a promotion committing before the nested DELETE's row lock is re-checked by the nested guard (owner authority/last owner); worst case is an existing co-owner, already permitted | note | accepted; no change (an optional `FOR NO KEY UPDATE` on the loop is left for Codex) |
| SR2-4 | The 40P01 caveat was true but too narrow (other facade operations locking the subject's binding after the Organization share; owner raw departure locks the Organization in step (4); the G058 edge pre-existed for subjects without Group bindings) | should-fix (doc) | **fixed**: AUTHORIZATION.md paragraph widened (doc only, no code change) |
| SR2-5 | Enabled-owner check: no lock cycle with DisableUser/FinalizeRetirement (users then sessions, never organizations); fails closed | note | accepted |
| SR2-6 | Down restores the version-16 select/delete policies and guard byte-identically | note | accepted |
| SR2-7 | Go sound; the `fl == nil` arm in `RemoveMember` is dead (harmless) | note | accepted, unchanged |
| SR2-8 | No grants, DEFINER or functions added; PC2/PC3/PC4 cover the new pieces | note | accepted |
| SR2-9 | Pre-existing (G058): two concurrent facade departures in one Organization can deadlock (Organization share then the G058 NO KEY UPDATE) | note | out of scope; reported for Codex (verify it is recorded under G058) |

## S2 correction attempt 3 (B5 attempt 3, checkpoint 3) — 2026-10-19

**Status: corrected S2 checkpoint submitted for Codex review (review-ready). Not accepted.** SR2-3 fixed in 00017 with a
deterministic raw interleaving proof that fails on the pre-fix migration; `TestAccessPolicy` (39/39) and `TestOnboarding`
(14/14) PASS after reconciliation; `TestMCPDiscovery` (5/5), `TestMCPTransport` (7/7) and logcheck validate/scan now
RUN and PASS; focused regression PASS on the final source. S3 (UI) HELD. Manual review OWNER-WAIVED.

### Worker record (attempt 3)

- Task G060, role S2-correction writer (attempt 3, single writer), model Opus 5.5, intended effort high; actual runtime
  effort not verifiable from inside the agent (unknown).
- Worktree `…/.claude/worktrees/agent-<id>` (native `agent-<id>`, session branch at `c91c945b`). Local
  `g060-impl` checked out, branch/HEAD asserted `a3fba87dd5fea54b4f63544e2e92442c37ec3e5f` before the rebase; rebased onto
  `origin/staging` `4face233` (management-only: SUPERVISOR-HANDOFF, PROGRESS, this Result, review record, scope list) with
  no conflict (staging's Result already carried attempt 2) → `64fa8131`, published with
  `--force-with-lease=g060-impl:a3fba87d…`. Before every batch: `git fetch origin staging` (stayed `4face233`), ancestor
  check rc 0, `bash infra/ci/docker-desktop-mounts.sh` (15608–15664 of 100000, 15%).
- Inputs verified: review record SHA-256 `2195be52…41e9`, scope allow-list (35 patterns) `5620573c…b7a7`, prior Result
  `463dad3b…df86d`. Commit `a75d294a` (code/tests/docs/catalog), then this Result; safekeeping pushes; no MR.

### Corrections

1. **SR2-3 owner-subject race** (`core/migrations/00017_group_authority.sql`). The guard step (3) loop query is now
   `… ORDER BY rb.group_id FOR NO KEY UPDATE LOOP`: each Group binding is locked, ascending, before `r.role` is read. A
   concurrent promotion either committed first (locked read/EPQ returns `group_owner` → 42501 `owner_subject`) or waits
   for the departure and then updates no row. Under RLS a locking read also needs UPDATE USING, so the
   `role_bindings_update` USING Group branch became `GMGR OR (MGR AND (OVR OR DEP))`; WITH CHECK is unchanged (no DEP;
   `GMGR OR (MGR AND OVR AND subject <> actor)`), so the widening is lock-only (runtime's only UPDATE grant is
   `UPDATE (role)`, and no trigger body updates role_bindings — security review). PC3: the trigger-depth inventory is now
   `enrollment_restrictions_owner_insert,role_bindings_delete,role_bindings_insert_bootstrap,role_bindings_select,role_bindings_update`
   plus a check that `role_bindings_update` names `pg_trigger_depth()` in USING only and keeps the override in both USING
   and WITH CHECK. PC4: the lock fragment must exist and precede the role check. Down unchanged (restores the 00015
   policy, byte-identical to version 16). No new function, DEFINER, grant, role or trigger. AUTHORIZATION.md (raw
   departure lock paragraph) records the locked read, OR-7 and the lock-timing note.
   - **OR-7** new in `TestRLSOrdering` (raw SQL, two runtime connections, lock-wait observation): remover = Organization
     admin and actual owner of g; target = Organization member, g admin and member of g2 (owned by another; override off,
     so g2 is reached only through the DEP branch, including its lock). Promotion committed first, and promotion in flight
     first → removal 42501 `owner_subject`; bindings `g2:group_member,g:group_owner,organization:org_member`; audit =
     baseline + exactly one `membership.role_changed`, no removed row. Removal in flight first → the promotion waits,
     updates 0 rows, its commit writes nothing; no bindings remain; audit = baseline + three `membership.removed`, no role
     change.
   - **Reproducer control r0:** the same OR-7 on a disposable copy of the tree with the committed (pre-fix) 00017
     (SHA-256 `bc448713…b4bb`): **FAIL** — "promotion in flight first: removal <nil> (SQLSTATE "" constraint ""), want
     42501 owner_subject" (the race is real and OR-7 detects it). Copy deleted afterwards.
   - **O-MEM-4a narrowed:** promotion first now accepts only `owner_subject` (the unlocked read's `owner_authority`
     alternative is no longer accepted); f2 observed `refused forbidden:owner_subject`.
2. **p1 reconciliation** (no production change). Actors moved to real Group roles where the test is about the Group path
   (IT-FACADE-1 ChangeGroupSlug → g1 owner, UpdateBusinessProfile → g1 admin; IT-FACADE-6 slug conflict/no-op → g1
   owner); explicit owner override setup (new helper `withOverride`: facade `SetGroupManagementOverride`, turned off at
   subtest end) where the subject is the Organization path (IT-CTX-9, IT-FACADE-2/3/4, IT-REVOKE-1, ONB-9 admin selection
   of unlinked pU). OFF negatives added with no-write checks: IT-PROFILE-3 owner selectable set empty off / p1,p2,p3 on;
   IT-FLAG-4 owner/admin refused off (rows unchanged), allowed on with the flag off; IT-REVOKE-8 admin refused off,
   allowed on; ONB-9 off `forbidden` with nothing attached; ONB-12 owner/admin `can_update` false off, true on, Group owner
   with the flag off stays false with the override on. ONB-10 follows r1 C2 (read listing, independent hints): the owner
   lists all 7 (3 pages) with select/update true only for its two Group-linked profiles off and all true on; gown lists
   ps0+ps1 in 2 pages (hints false/false and true/true); the member lists ps0 read-only; selected gs[0] lists ps0
   read-only. C2 paging kept.
   - **IT-LOG-1 diagnosis:** the two p1 `audit.record_failed` records came from IT-FACADE-2 "server-rejected COMMIT"
     (`commitDefinitive`) and IT-FACADE-3 "lost reply" (`commitLostReply`): with the stale owner actor the guarded call
     was a marked denial (`role_not_permitted`, as the p1 lines show) and its `access.denied` record ran through the same
     hooked DB while the COMMIT fault was still armed, so the record failed (exactly those two; the W-statement and
     after-W faults never reach the record's INSERT/COMMIT). With the override set before arming, the fault hits the
     guarded write as intended. p2: 842 records, all `access.decision`, no `audit.record_failed`; IT-LOG-1 unchanged.
3. **Pins:** `mcp/env_integration_test.go:253` and `access/recovery_integration_test.go:92` → `current_version 17`;
   `tenancy/integration_test.go` cleanup gains a version-17 branch (override reset + audit purge through the attested
   superuser, down to 16) before the unchanged conditional version-16 branch.
4. **MCP CAT-1** (`mcp_integration_test.go`, found by c1): the Organization admin without Group bindings lists no
   Group/profile tool with the override off (Organization-only and A1 selected); with the override on (fixture setup on
   the setup connection, the setFlag pattern, reset at CAT-1 end) the pre-G060 expectations hold, including after unlink
   and with the flag off.
5. **Log catalog** (`events.json`, found by logcheck-1: 23 `event-keys-invalid`): allow records may carry
   `allowed_override` (new rule: only with outcome allow and capability group.update, group.manage_members,
   group.complete_onboarding, business_profile.profile.update or business_profile.select_for_onboarding); the
   `business_profile.list` rule admits `business_profile.profile.read` (r1 C2). See D-A3-1.

### Validation (exact; private per-run DB, no lock)

Scripts under `platform-evidence/g060-s12/a3/`: `b1.sh`, `b1tools.sh`, `b2.sh` (internal network `g060a3-<run>-net`,
pgvector 0.8.7-pg17 on tmpfs, no published port (inventory `ports=[]`), read-only source and `g058-a1-gomod`,
`-mod=readonly`, `GOPROXY=off`, label `g060-a3=<run>`, credentials mode 600 in `a3/secrets`, operation logs and canaries in
`a3/logs/<run>/`), `b2pre.sh` (b2 plus one prior selector on the same DB), `logcheck.sh`, `mkold.sh`, `scope.sh`. At most
two concurrent stacks.

| Run | Selector | Source | Outcome |
|---|---|---|---|
| b1-c1, b1-c2 | integration-tag compile of tenancy/access/mcp/migrate | working tree | PASS |
| r1 | `^TestRLSOrdering$/^OR-7` | working tree (fix) | **PASS** (3 variants) |
| r0 | same, pre-fix 00017 copy | control | **FAIL as expected** (race reproduced) |
| p2 | `^(TestAccessPolicy\|TestOnboarding)$` | working tree | **PASS** 39/39 + 14/14, 0 skipped, 842 log records |
| c1 | `^(TestMCPDiscovery\|TestMCPTransport)$` | working tree | Transport **PASS 7/7**; Discovery FAIL CAT-1 (fixed, item 4) |
| c2 | `^TestMCPDiscovery$` | working tree | **PASS 5/5** |
| t1 | pre `TestRLSOrdering/OR-5` (DB to 17), then `^TestTenancyModel$` ./internal/tenancy | working tree | **PASS** 37/37 (17→16→…→5 step-down exercised) |
| b1-tools 1–3 | tools `./internal/logrecord ./cmd/logcheck` | events.json iterations | 1–2 FAIL (pinned vector, D-A3-1); 3 **PASS** |
| m1 | `^(TestMigrationLifecycle\|TestMembershipCatalog\|TestEnrollmentCatalog\|TestPolicyCatalog\|TestObjectsCatalog\|TestOnboardingCatalog\|TestAuditCatalog)$` (bound replay SQL) | final | **PASS** 7 tests, 49 subtests |
| f2 | `^(TestMembership\|TestRLSOrdering\|TestRLSIsolation\|TestAuthority\|TestAccessOrdering)$/…` | final | **PASS** 80 subtests = attempt-2 f1's 79 + OR-7 (D-A3-2) |
| k1 | `^(TestIsolationAttacks\|TestObjects\|TestAudit)$` + private MinIO | final | **PASS** 27 subtests; OBJ oracle service 297/0, RLS 85/0; AUDIT-INV 46/46 |
| v1 | `^TestRecoverySeed$` + MinIO | final | setup with the version-17 pin passed; then `RECOVERY_OLD_DIR is required` (drill-only env; the drill NOT RUN) |
| b1-run1 | Plan §5.3 b1 (8 commands + compile) | final | **PASS** all (incl. TestAccessBoundary, TestRLSBoundary) |
| logcheck-1 | validate/scan p2 c1 c2 t1 | — | p2 validate FAIL (catalog, item 5); scans 0 hits; controls detected |
| logcheck-3 | validate/scan p2 c1 c2 t1 f2 k1 (+ each PostgreSQL log) | final catalog | **PASS**: validate rc 0 ×6, 0 canary hits, planted canary detected ×6 |

Logcheck build: lock-free `golang:1.27.2-alpine --network none`, read-only `infra/validation/tools` and `g058-a1-gomod`
(which already held jsonschema v6.0.3 and x/text v0.14.0), `-mod=readonly GOPROXY=off`, disposable GOCACHE and binary
inside the container; no download, no shared image, no lock. MinIO provenance inspected before k1/v1: server
`sha256:c1194d7e…6f4e20b`, init `sha256:46b1f9f1…5368e7` (labels as attempt 2); no rebuild or substitute.

**Evidence attribution.** "final" = the `a75d294a` tree. p2/c1/c2/t1/r1 ran on the working tree before the last edits: one
PC3 postcondition line (a migration-time assertion, exercised by every later migration including m1/f2/k1) and
events.json/docs (not core inputs); c2 ran after the CAT-1 edit. They are retained with that delta. From 999922fb:
f1/k7/m6/m8 are superseded by f2/k1/m1 (re-run because 00017 policies changed); q2 (OR-3/OR-6) and the attempt-2 b1 runs
are superseded by f2/b1-run1. Nothing else from 999922fb is relied on.

> **Attribution correction (attempt 4, per the Codex checkpoint-3 review).** The attempt-3 run headers record `64fa8131`
> with a **dirty working tree** (for example `b2-r1.log`: `dirty=8`), not a committed `a75d294a` checkout. The statement
> "final = the `a75d294a` tree" above is the worker's **reported** reconciliation of the tested working-tree files to the
> later commit `a75d294a` (with the declared late postcondition/catalog/prose deltas); it is not an independently captured
> per-run tree proof. Also, f2 selected all 40 `TestAccessOrdering` cases but only OR-7 of `TestRLSOrdering`, so it did
> **not** supersede attempt-2 q2 (OR-3/OR-6). Both points are superseded by the attempt-4 committed-source runs `or1` and
> `e1` below (clean checkouts; commit, tree and core-tree hashes captured in each run header).

**NOT RUN:** harness IDs (B-*) as such; B-RECOVERY-DRILL (TestRecoverySeed needs the drill's directories; only its setup
pin was exercised); TestMCPLifecycle/Audit/Recovery and TestMCPOrdering/Footprint (outside this correction's selectors);
S3/S4.

### Deviations for disposition (attempt 3)

- **D-A3-1** The `business_profile.list` rule admits both `business_profile.profile.read` (what core emits since r1 C2)
  and the old `select_for_onboarding`, because the pinned accept vector at
  `infra/validation/tools/internal/logrecord/logrecord_test.go:322` is outside the 35 patterns. Narrowing to read-only
  needs that exact path approved.
- **D-A3-2** f2 selected whole TestMembership/TestRLSIsolation/TestAuthority/TestAccessOrdering (my subtest regex covered
  every subtest; the same composition as attempt-2 f1) rather than departure/promotion subsets. Technical reason: the
  role_bindings_update USING change reaches every raw binding lock/UPDATE path those suites exercise. Cost ~30 s.
- **D-A3-3** IT/ONB/CAT reconciliation uses explicit override setup for Organization-path subjects (facade in access, the
  setup connection in MCP, the existing setFlag pattern). No oracle widened, no production change.

### Security review (attempt 3)

One bounded read-only reviewer on the 00017 delta: **no blocker, no should-fix**. Notes: the locked read closes both
orders (the EPQ re-check rebuilds `r.role`; RLS quals do not depend on role; `group_id` is immutable); removal first gives
0 rows; DEP in UPDATE USING is lock-only (no trigger body updates role_bindings; runtime grant `UPDATE (role)`; WITH CHECK
strict); no FK references role_bindings, so the NO KEY UPDATE→DELETE upgrade meets no KEY SHARE holder; no new cycle
class beyond the documented Organization-row caveat; PC3/PC4 parse and fail closed; down byte-identical. Optional items
applied: PC3 also asserts USING still names the override; AUTHORIZATION.md qualifies "no further wait edge" (locks are
taken earlier, and an `owner_subject` refusal may wait).

### Native refusals (quoted once, not retried; work continued with Edit and plain commands)

"This agent is isolated in the worktree <source-checkout>,
but this command is too complex to verify that it stays inside the worktree. Refusing to run it — a worktree-isolated
agent's git operations must target its own worktree. Split it into plain, separate commands and run them from
<source-checkout>" (a python heredoc edit and a
heredoc append to this Result), and its variant "… but this command runs sed with a program computed at runtime inside a
construct too complex to verify, so what it runs cannot be shown not to be git. …" (a shell loop). No check was skipped
because of a refusal.

### Cleanup

Every run printed `CLEANUP label=g060-a3=<run> containers=0 networks=0 secrets=0`. Final check: no `g060-a3` container or
network, `a3/secrets` empty, no pgvector 0.8.7-pg17 container running. The pre-fix copy `a3/oldcore` was deleted. Logs and
canary files stay under `a3/logs/` as evidence (test canaries, not secrets).

### Checkpoint-3 assessment (mandatory)

- **Convergence:** converging. All three consolidated items of the attempt-2 review are addressed with runtime proof, and
  every selected suite passes on the final source. The only red runs were first-contact findings fixed within scope (the
  CAT-1 stale positive and two catalog rules) and the intended pre-fix control.
- **Remaining criteria and uncertainty:** harness-level B-* qualification (including B-RECOVERY-DRILL, the rest of B-MCP-IT
  and the ordering suites not run here), S3 web/UI (HELD), S4, and final local and hosted qualification. D-A3-1..3 need
  disposition. The pre-existing G058 two-departure deadlock note (SR2-9) remains open for Codex.
- **Budget:** B5 attempt 3 of 5 consumed (not reset). No attempt-4 release or budget change follows from this checkpoint.

## S3–S4 attempt 4 (B5 attempt 4) — 2026-10-19

**Status: S3 (web) and S4 (harness/docs) source complete; every lock-free check run and PASS; the web unit/lint/typecheck
build and the browser checks are NOT RUN (pending: a native refusal stopped the private build, see below). Submitted for
Codex review (review-ready), not accepted.** Manual review OWNER-WAIVED, not performed. No MR.

### Worker record (attempt 4)

- Task G060, role S3–S4 writer (attempt 4, single writer), model Opus 5.5, intended effort high; actual runtime effort
  not verifiable from inside the agent (unknown).
- Worktree `…/.claude/worktrees/agent-<id>` (native `agent-<id>`). `g060-impl` switched in, branch/HEAD
  asserted `a410de78c6075f91297f5c7995dd5c41ea4f83cb`, rebased onto `origin/staging` `0215b237` (management-only; the
  Result-only checkpoint commit dropped as already upstream) → `22d9bec2`, pushed with
  `--force-with-lease=g060-impl:a410de78…`. Staging then advanced to `26a0a720` (governance/management-only: the owner's
  private-browser/sequential-build scheduling decision); rebased again (no conflict) → catalog `4aac5443`, S3 `12aca202`,
  pushed with `--force-with-lease=g060-impl:909f0ad1…`; S4 `761c990c`; then this Result. Before every batch: fetch,
  ancestor check rc 0, `bash infra/ci/docker-desktop-mounts.sh` (15688–15712 of 100000, 15%).
- Inputs verified: review record SHA-256 `8fd400b2…dedda` at `0215b237` (the `26a0a720` addition, the owner scheduling
  decision, read in full), scope allow-list 36 patterns `bb8a106a…d39e`, prior Result `8584aeb4…b2eb`.

### D-A3-1 catalog (done)

`infra/validation/tools/internal/logrecord/events.json`: the `business_profile.list` rule is `const
business_profile.profile.read` (the legacy `select_for_onboarding` alternative dropped; core emits only `profile.read`
for the listing, `access/onboarding.go` `listed(…, capProfileRead, …)`). `logrecord_test.go`: the positive vector now
uses `profile.read`, plus one `allowed_override` positive (`membership.change_role`/`group.manage_members`) and four
negatives (the obsolete capability on the listing; `allowed_override` on the metadata listing, on a denial, on
`group.manage_owners`). `allowed_override` stays limited to allow decisions of the five management capabilities.

| Run | Command | Outcome |
|---|---|---|
| b1-tools1 | `a4/b1tools.sh`: tools module `go test -count=1 ./...` (`golang:1.27.2-alpine --network none`, read-only tools source and `g058-a1-gomod`, `-mod=readonly GOPROXY=off`, disposable GOCACHE) | **PASS** all 7 packages (events.json `4050e9f0…2bfb57`) |
| logcheck-a3logs | `a4/logcheck.sh` validate + scan of the retained attempt-3 logs p2 c1 c2 t1 f2 k1 (+ each PostgreSQL log) with the corrected catalog | **PASS**: validate rc 0 ×6 (p2 holds 14 `business_profile.list` records, all accepted), 0 canary hits, planted canary detected ×6 |
| logcheck-a4 | same on or1, e1 | **PASS** ×2 |

### D-A3-2 committed-source ordering proof (done)

Private per-run DB (`a4/b2.sh`, the attempt-3 runner with label `g060-a4=<run>`, network `g060a4-<run>-net` internal,
`pgvector/pgvector:0.8.7-pg17` on tmpfs, inventory `ports=[]`, read-only source and module cache, credentials mode 600
under `a4/secrets` deleted at exit; the new `clean` mode refuses to start unless `git status --porcelain` is empty and
the header records commit, tree and core-tree; one stack at a time).

| Run | Selector (`./internal/tenancy/access/`, `-tags integration -count=1 -v`) | Source | Outcome |
|---|---|---|---|
| **or1** | `-run '^TestRLSOrdering$/^(OR-3\|OR-6\|OR-7)'` | commit `909f0ad1` (pre-second-rebase; tree `5f9a64b7…`), core tree `3caebabf…` = `a75d294a:core` = current `HEAD:core`, dirty 0 before and after | **PASS** OR-3, OR-6, OR-7 |
| **e1** | `-run '^(TestRLSOrdering\|TestEnrollment\|TestEnrollmentAdmission\|TestEnrollmentRequests)$'` (whole) | commit `761c990c`, tree `0597fdcb…`, core tree `3caebabf…`, dirty 0 | **PASS** TestRLSOrdering 7/7 (OR-1..OR-7), TestEnrollment 2/2, TestEnrollmentAdmission 14/14, TestEnrollmentRequests 15/15; 0 FAIL/SKIP; 345 access records |

Both supersede the attempt-2 q2 OR-3/OR-6 record and the attempt-3 working-tree attribution for these cases. e1 also
fixes two S4 pins from execution (B-RLS-ORDERING 7; B-ENROLLMENT-IT 2/14/15 at version 17). Every run printed
`CLEANUP label=g060-a4=<run> containers=0 networks=0 secrets=0`. Core source is unchanged in attempt 4 (core tree
`3caebabf` throughout), so the attempt-3 core evidence keeps its stated (reported) attribution; no broad repeat.

### S3 web (Plan r0 §1.6, r1 C2/C6/C7, correction02 C02-3/C02-4/§P)

| AC | Change (files) |
|---|---|
| AC3 owner-only setting | `app/organization/group-authority/page.tsx`: core read + own Organization role; owner → `GroupAuthorityForm` (native radio fieldset Off/On, scope text by `aria-describedby`, `PendingSubmit`), admin → `group-authority-readonly` effective state, anyone else → fixed 404. `…/action/route.ts` → `groupAuthorityActionRoute` (form layer, closed `enabled=on\|off`, session, then core; 303 `?status=enabled\|disabled` from core's answer; non-owner "Not permitted"; possibly-committed → unknown page). `lib/membership-client.ts` `groupAuthority`/`setGroupAuthority` (strict `{override_enabled, can_change}`, the answer must report the requested value). Links from `/organization/members` (`staffLinks`, owner/admin only) and `/account` (same audience as the branding link). |
| AC1/AC2/AC8 truthful Group UI | `lib/onboarding-client.ts`: Group DTO exactly 8 keys with `selfRole`/`canManage`/`managementSource` and their consistency (source ⇔ can_manage; `group_role` ⇔ own owner/admin; override never with an owner/admin role), row budget 1,500 → 1,600 (arithmetic in the comment); profile summary 5 keys with `canSelect`/`canUpdate`; `selectableProfiles` (filter, cursor untouched); `ONBOARDING_NOT_ADMINISTERED`. `components/membership/members-view.tsx`: no own-membership read without `self_role`; nonmember + override → `override-admin-notice`, admin lists, no self card/Leave/owner controls (owner candidates only for an actual owner); nonmember without management → `members-not-administered` (no table/form), `left` status kept; member + override → self card + `override-admin-member-notice`. `groups/[id]/members/page.tsx` passes the hints. `groups/[id]/page.tsx`, `groups/page.tsx`: Resume only when `can_manage`, otherwise the notice (and Open); Members link only for an actual role or a manager. `groups/[id]/onboarding/page.tsx`: metadata + notice and no forms without `can_manage`; picker = `selectableProfiles`, "No Business Profile you can attach". `business-profiles/page.tsx`: "Business Profiles" — every readable row, `profile-read-only` marker. Refusal sentences: `self_target` reworded for invitations, `self_escalation` added. No new component, dependency, asset import or branding change; semantic tokens and existing `Text`/`TextLink`/`Button`/`PendingSubmit` only. |
| AC4/AC6 (UI) | E-MEM-9 and E-ONB-8 script enable → administer → disable → next request denied, genuine memberships kept. |

Tests (extend/new; each file in scope): unit `onboarding-client.test.mjs` (G1/G2 hints; the four C7 authority cases plus
the group_admin case; eight inconsistent-hint negatives and a missing key; summary hints and negatives; `selectableProfiles`
mixed/all-false/empty and through the client; create answer hints; byte arithmetic +90/+38), `membership-client.test.mjs`
(Group-authority read/write requests, mismatch and forbidden, malformed answers, size bound), `onboarding-forms.test.mjs`
(fixture hints; forged completion by a non-manager → 403 "Not permitted", no redirect), **new**
`group-authority-forms.test.mjs` (closed choice incl. repeated field; statuses and sentences; 303 on/off with the exact
core body; non-owner 403; unknown outcomes ×3 incl. a value mismatch; form/origin/session refusals before core; 405s).
Browser `membership.spec.ts`: E-MEM-3 L279-287 rewritten (frank, now Group Member with the override off: self card +
Leave, no table/form, the "Only the administrators" sentence); E-MEM-5 (carol and frank on Design → 200
`members-not-administered`, no table/form/self card/Leave/owner controls; alice/absent/Organization B Operations → 404
for carol and frank); **E-MEM-8** (keyboard enable by carol, scope text and `aria-describedby`, admin read-only, forged
admin POST 403 with nothing changed, alice no links and 404, axe both themes; `finally` resets off); **E-MEM-9** (on a
per-run Group created by bob — see deviation D-A4-2 — override off metadata only; on: Members link, `override-admin-notice`,
table and invitation form, no self card/Leave/owner controls, axe; frank invites alice who joins; bob invites frank:
self card Member + Leave + `override-admin-member-notice`; off: frank keeps self card only, Design shows
`members-not-administered`, alice keeps her membership; `finally` off). `onboarding.spec.ts`: E-ONB-3 updated for r1 C2
(heading "Business Profiles"; dave lists Research Labs marked Read-only); **E-ONB-8** (alice's truly incomplete Group;
carol and frank off: badge + notice, no Resume/forms/Members, list Open only, forged completion 403, Design profile
read-only from Group page and index; on: Resume, both forms, picker offers Design Studio and Research Labs, not
Operations Partner; off: notice again, alice's owner path unchanged; axe; `finally` off). `infra/validation/lib.sh`
matrix rows: onboarding + E-ONB-8, membership + E-MEM-8 E-MEM-9.

Offline syntax only (`node:24-slim --network none`, read-only web source): `node --check` of the four changed/new unit
test files **PASS**. Not a substitute for b3 (TypeScript and TSX are not checked by `--check`).

### S4 harness and docs

`infra/validation/run.sh` (exact consumers): `CORE_MEMBERSHIP_ROUTES` 16 → 18 pairs (C-sorted; equals the 18 `Route*`
constants of `membership/api/http.go`) and the B-MCP-SOURCE strings; `WEB_ROUTES` + `organization/group-authority/
action/route.ts`, `organization/group-authority/page.tsx`; APPLICABLE copy list + `web/app/organization/group-authority`;
B-WEB-BOUNDARY record; B-MIGRATE-SET + `00017_group_authority.sql`; the README migrate block (pending/applied/status for
00017, `current_version 0 17 17`); B-MIGRATE-CLI/B-MIGRATE-CATALOG steps (fresh status + 00017 pending; apply/reapply
00017 at 17; rollback/inspect at 16 with 71 relations; the unique roles rows become {0, 71} — the 56 row came from the
version-15 rollback and can no longer occur); B-RECOVERY-DRILL, B-PERSIST, B-DBRENAME and B-TOGGLE-OFF version 16 → 17;
B-AUDIT-CATALOG reads 00017's `events_action_valid` (46 actions, 39 trigger-owned; verified locally: the AUDIT.md
catalog and the 00017 constraint are identical, 46 words); B-AUDIT-IT inventory 46/46 (attempt-3 k1 observed 46/46);
B-RLS-IT RLS-SCOPE 36 → 40 cases (attempt-3 f2); B-RLS-ORDERING 5 → 7 (e1); B-POLICY-ORDERING 37 → 40 (attempt-3 f2:
O-MEM-4a/4b/4c); B-MEMBERSHIP-IT 31 → 32 (MEM-13b) and **TestAuthority 15** run in the same invocation (D-A4-1);
B-TENANCY-HTTP + the group-authority probe (GET 400 `ambiguous_parameter`, POST {} 403 like `/invitations`);
B-ENROLLMENT-IT and B-IDENTITY-CATALOG record text at version 17. Unchanged pins confirmed from attempt-3/4 counts:
TestAccessPolicy 39, TestOnboarding 14, TestMCPDiscovery 5, TestMCPTransport 7, TestTenancyModel 37, TestMigrationLifecycle
10, catalogs (Membership 6, Enrollment 19, Policy 6, Objects 3, Onboarding 5, Audit 7), TestObjects 10 (297/85),
TestIsolationAttacks 22 (142), TestAudit 9, TestRLSIsolation 10. Static: `bash -n` of `run.sh` and of `lib.sh`
(`node:24-slim --network none`, read-only) **PASS** (rc 0 each).

Docs: AUTHORIZATION.md (G060 subsection superseding the `all` management cells: matrix, override, self-grant refusals,
liveness, objects), DATABASE-ISOLATION.md (00017 policies/helpers/guard), AUDIT.md (T3 reasons, G7
`access.override_used`, 65 operations, 00017 paragraph), TENANCY.md (G060 section), OBJECT-STORAGE.md (Group object
table), MCP.md (selection vs management), WEB-SHELL.md (G060 pages/hints/notices, project rows E-ONB-1..8, E-MEM-1..9),
MIGRATIONS.md (row 17, "from version 17", down 17), ARCHITECTURE.md and SECURITY.md (implementation-status notes; target
text preserved). DECISIONS.md (Codex-owned) untouched.

Scope: `a4/scope.sh` — 36 patterns, SHA-256 `bb8a106a…d39e`; 85 changed paths vs `origin/staging` `26a0a720` (committed
and untracked), **all in scope**.

### Pending web/browser selectors (NOT RUN; need image builds and private stacks)

| ID | Selector | Image/target and stack |
|---|---|---|
| b3 / B-UNIT-WEB (B-UNIT-CHECKS web-check) | `npm run check` = `eslint --max-warnings=0 .`, `tsc --noEmit`, `node --conditions=react-server --test "tests/unit/*.test.mjs"` (incl. new `group-authority-forms`) | `infra/docker/web.Dockerfile` target `check`, context `web/`, run `--network none` (the `npm ci` deps layer may need the npm registry if not cached) |
| b4 membership | `IDENTITY.md` `identity:browser` line, `IDENTITY_E2E=1`, `--project=membership-light --project=membership-dark --grep 'E-MEM\|S-LANDMARKS\|S-NAV'` (E-MEM-1..9) | `infra/docker/web-browser.Dockerfile` (Playwright 1.63.0) through `infra/compose/browser.yml` on the private identity environment (`infra/compose/identity.yml`: core, web runtime, PostgreSQL, emulator, MinIO) |
| b4 onboarding | same line, `--project=onboarding-light --project=onboarding-dark --grep 'E-ONB-3\|E-ONB-4\|E-ONB-8'` (E-ONB-3 changed for r1 C2) | same |
| account page regression | `identity.spec.ts` E-BRD-* (the branding link now sits in a flex row with the Group-administration link) on `domain-platform`/`domain-custom`; `shell-*` S-ACCOUNT-* | same stack (domain projects) / fixture shell |
| final affected union before the MR (Plan §5.3) | B-UNIT-WEB, B-IDENTITY-E2E (membership + onboarding projects), B-WEB-BOUNDARY, B-TENANCY-HTTP, B-MCP-SOURCE, B-MIGRATE-SET/CLI/CATALOG/IT, B-AUDIT-CATALOG/IT, B-RLS-IT/ORDERING, B-POLICY-ORDERING, B-MEMBERSHIP-IT (+TestAuthority), B-ENROLLMENT-IT, B-RECOVERY-DRILL, B-SCOPE, B-REQUIRED-CHECKS | harness stacks |

### Deviations for disposition (attempt 4)

- **D-A4-1** B-MEMBERSHIP-IT now runs `'^(TestMembership|TestAuthority)$'` and pins TestAuthority 15. Reason: the Plan
  placed AUT-*/MEM-OVR-*/O-OVR-* in `authority_integration_test.go` (implementer's choice recorded in S1), but no harness ID
  ran it, so the AC3/AC4/AC6 integration proofs would be absent from the final local qualification. Minimum mapping
  change inside an approved path; no new ID or CI trigger.
- **D-A4-2** E-MEM-9 administers a per-run Group created by bob (a plain Organization member) instead of seeded Design /
  Research (Plan T11/C7 text). Reason: inviting alice into Design and frank into Research would persist across the two
  sequential membership projects (the second project could not invite them again, and E-MEM-5's alice-on-Design 404
  would fail). The seeded Design is still used for the read-only checks.
- **D-A4-3** The `members-not-administered`/`override-*` C7 "unit four cases" are covered at the DTO level in
  `onboarding-client.test.mjs` (the unit suite imports `lib/*` and route modules only and cannot render a server
  component); the rendered proof is E-MEM-3/5/9.
- **D-A4-4** Pins taken from attempt-3 executed counts rather than an attempt-4 rerun (core unchanged): RLS-SCOPE 40,
  TestAccessOrdering 40, TestMembership 32, TestAuthority 15, AUDIT-INV 46. Their final confirmation is the harness run.
- Observation (not changed, outside this task): the B-POLICY-IT record says "the oracle's 242 cells" while
  `integration_test.go` asserts 418 (pre-existing stale prose).

### Native refusals (quoted once, not retried, not routed around)

After the coordinator relayed the owner's private-build approval, writing the b3 build script
(`a4/b3.sh`: one private `docker build --target check` with a task-owned tag, then the image's check `--network none`)
was refused: "Permission for this action was denied by the Claude Code auto mode classifier. Reason: [Auto-Mode
Bypass]." A following read-only `sed` of `run.sh` got the same refusal; the work continued with Read/Edit and plain
commands. No image was built, no Compose or browser stack started, and the exclusive build slot was never used (never
held by an active build). Remedy for the coordinator/owner: the approved private build/browser operations need a native
permission the session will accept (for example an allow rule for the task's build and browser commands), or the
coordinator runs b3/b4 itself. Also self-reported: one read-only `git -C <worktree> status --short` slipped through
despite the dispatch's no-`git -C` rule; nothing else used it.

### Cleanup

`or1` and `e1` each printed `CLEANUP … containers=0 networks=0 secrets=0`; `a4/secrets` is empty; no `g060-a4`
container, network or volume remains; no image was built or pulled (no task-owned image exists). Evidence and scripts stay
under `platform-evidence/g060-s12/a4/` (logs, canaries — test canaries, not secrets). Two transient `/tmp/g060a4-*.txt`
comparison files were written by the AUDIT.md parity check.

### Budget

B5 attempt 4 of 5 consumed (not reset). Remaining: b3/b4 execution and any first-contact fixes, Codex pre-integration
review, MR fast/Duo, merged-SHA local full and cleanup. No attempt 5 is implied.

## Attempt 5 (B5 attempt 5, final) — 2026-10-19

**Status: the bounded attempt-5 work (read-only onboarding navigation, reporting fixes, Option B invitation message)
is committed; b3 and the focused b4 browser selectors PASS on frozen committed source after one first-contact test
correction; submitted for Codex pre-integration review (review-ready), not accepted.** No MR. Manual review not a gate
(owner-waived as before). Lock-requiring harness checks are listed for the merged-SHA local full.

### Worker record (attempt 5)

- Task G060, role attempt-5 writer (single writer), model Opus 5.5, intended effort high; actual runtime effort not
  verifiable from inside the agent (**unknown**). Effective permission mode: **unknown** (dispatched as the owner's
  default mode; the session's own context text also referred to auto mode; no prompt and no refusal was observed for any
  operation of this attempt).
- Native worktree `…/.claude/worktrees/agent-<id>` (`agent-<id>`). `g060-impl` switched in; branch and HEAD
  asserted `815e4d13e86eadcfe9f9ff60a15c02cf903f593b` before any rebase. Rebased onto `origin/staging` `7eaa4bbe`
  (management-only; the attempt-4 Result commit `815e4d13` dropped as already upstream, Result byte-identical to
  staging's imported copy `f542631a…`) → `5f2c8272`, pushed with `--force-with-lease=g060-impl:815e4d13…`. Staging then
  advanced to `b539e615` (governance/permissions/management only: the attempt-5 permission installation records and
  the reusable project test permissions); the attempt-5 source commit was rebased onto it (no conflict) → `643bb15f`,
  pushed with `--force-with-lease=g060-impl:5f2c8272` (deviation D-A5-3). Before every batch: fetch, ancestor check rc 0;
  `bash infra/ci/docker-desktop-mounts.sh` before the first build (15724 of 100000, 15%).
- Inputs verified: review record SHA-256 `4f5181b6…981c` at `7eaa4bbe` (both binding sections read in full; the later
  `b539e615` additions are the permission installation/adoption records only), scope allow-list 39 patterns
  `919ebace…cb17`, Goal/Spec Option B append at `c4def98c`. Plan/Goal/Spec/review files not edited.

### Changes (what, where)

| Item | Files | Change |
|---|---|---|
| Read-only onboarding navigation | `web/app/groups/[id]/onboarding/page.tsx` | The read-only view (`!canManage`) renders only `#new-profile`; its Primary navigation now offers only that section, labelled "Group onboarding" (no promise of a form). The manager view keeps "Create new"/"Select existing". No new navigation framework: the page passes its own `ShellNavItem[]` to the shared frame and narrows it before `r.frame` (the frame renders the array at call time); `components/onboarding/onboarding-frame.tsx` is out of scope and unchanged. Observation (not changed, G018 behaviour outside the finding): the completed-Group view also renders only `#new-profile` under the full navigation. |
| B-POLICY-IT record | `infra/validation/run.sh` | "the oracle's 242 cells" → 418 (`integration_test.go` asserts 418). |
| Latest-state IT prose | `infra/validation/run.sh` | B-RLS-IT 6 → 17, B-AUDIT-IT 10 → 17, B-ONBOARDING-IT 14 → 17 (the shared access-package environment asserts `current_version 17`). Historical catalog/migration records keep their versions. Observation for review (not changed, outside the named three): B-OBJ-IT and B-RETENTION-IT say 10, B-DOMAINS-IT/B-MCP-IT 11, B-BRANDING-IT 13. |
| B-AUDIT-IT signer | `infra/validation/run.sh` | "an admin's URL of an unbound Group's object carries the Group" → the Group admin's signed URL carries that Group; an Organization admin with no binding is refused not_found and no url row is written (exactly `audit_integration_test.go` AT-FMT). |
| DEP exceptions | `docs/foundation/DATABASE-ISOLATION.md` | G060 summary names `DEP` (`pg_trigger_depth() >= 1`, the guard's departure cascade) and the reviewed exceptions: `role_bindings_select`/`_delete` Group branches `GMGR OR (MGR AND (OVR OR DEP))`, `role_bindings_update` USING the same for the cascade's `FOR NO KEY UPDATE`; WITH CHECK has no DEP; depth-zero staff roster/lock/UPDATE/DELETE without the override still see no row. Matches 00017 L191–223. No migration or core change. |
| Option B (Organization) | `web/lib/enrollment-client.ts` | New exported `INVALID_INVITATION` = "Invalid invitation link. Switch to a different account, or ask for a new invitation link."; `JOIN_NOTICES.invitation_mismatch` uses it. Callback redirect (`/join?status=invitation_mismatch`) unchanged. |
| Option B (Group) | `web/app/invitations/[id]/page.tsx` | A malformed id (no core call) and core's closed `not_found` render the same sentence in the page frame instead of `notFound()`. Unchanged: the shared frame (non-tenant host 404, sign-in first, unavailable session), other refusals via `r.refused` (incl. `origin_rejected` → 404, forbidden/not_available, infrastructure → unavailable), Organization-scope/closed/pending bodies, the accept route, core. The malformed-id check moved after the frame request, so an anonymous visitor now sees "Sign in first" for every id alike. |
| Tests | `web/tests/unit/enrollment-client.test.mjs` (extend), `web/tests/browser/membership.spec.ts` (new **E-MEM-10**; E-MEM-3 correction), `web/tests/browser/onboarding.spec.ts` (E-ONB-8 extend), `infra/validation/lib.sh` (matrix row + E-MEM-10) | Unit: exact sentence via `INVALID_INVITATION`, `joinNotice` and `JOIN_NOTICES`; no account/email/existence wording; `not_found`/`unavailable` notices unchanged. E-MEM-10: frank invites alice to his per-run Group; dave (plain Organization A member, not recipient, not administrator) opens the real link, a random nonexistent UUID and `/invitations/not-an-invitation`: each 200, heading "Invitation", `status-message` = the sentence, no code field or Accept button, no alice name/email in `main`, the three `main` texts identical; axe at 1280/320; alice then accepts the same link (`joined`), frank's table shows her as Member. E-ONB-8: the read-only view's Primary navigation is exactly ["Group onboarding" → `#new-profile`], the target exists and keyboard Enter reaches it; the override manager view keeps exactly both links, each reaching its section. |
| Docs | `docs/foundation/WEB-SHELL.md` | membership row E-MEM-1..10; G060 bullet: the read-only onboarding navigation and the Option B presentation. |

### Validation (frozen committed source; exact)

Source: **`643bb15f`** (tree `2b0e641e…`, web tree `fd9ddd40…`, core tree `3caebabf…` = accepted `a75d294a:core`, clean
checkout) for b3 and b4 run 1; **`04d897f8`** (tree `5d26cd4a…`, web tree `b9ec3bd8…`; differs from `643bb15f` only in
`web/tests/browser/membership.spec.ts`) for the E-MEM-3 rerun and b3b. `origin/staging` `b539e615` an ancestor before
each batch.

Build order (one exclusive build slot, strictly sequential, task-owned tags, bound context of the clean checkout, no
shared cache prepared, no base pull needed — every base resolved from the local store/BuildKit cache):

| # | Tag | Dockerfile / target / context | Image ID |
|---|---|---|---|
| 1 | `g060-a5-web-check:643bb15f` | `web.Dockerfile` check, `web/` | `ae0c0e507187` |
| 2 | `g060-a5-core:643bb15f` | `core.Dockerfile` runtime, `core/` | `da3cfcd32f13` |
| 3 | `g060-a5-migrate:643bb15f` | `migrate.Dockerfile` runtime, `core/` | `c924344aebda` |
| 4 | `g060-a5-web:643bb15f` | `web.Dockerfile` runtime, `web/` | `34a83cdd3da1` |
| 5 | `g060-a5-web-browser:643bb15f` | `web-browser.Dockerfile` browser, `web/` | `5ac5e1f1c9ce` |
| 6 | `g060-a5-validation-tools:643bb15f` | `validation-tools.Dockerfile` runtime, `infra/validation/tools/` | `d2de145cac3c` |
| 7 | `g060-a5-web-browser:04d897f8` | as 5 at `04d897f8` | `9bfecb2a88ee` |
| 8 | `g060-a5-web-check:04d897f8` | as 1 at `04d897f8` | `441ce9f08c63` |

MinIO reused read-only, not rebuilt: `platform-minio:server-bc184510` `sha256:c1194d7e…` and `:init-bc184510`
`sha256:46b1f9f1…` (the provenance IDs of the attempt-3/4 runs; `minio.Dockerfile` is outside G060). The build slot was
held only during builds 1–8 and **released after build 8**; no build ran concurrently with another.

| Run | Command (exact shape) | Source | Outcome |
|---|---|---|---|
| **b3** | `docker run --rm --network none --label g060-a5=b3 g060-a5-web-check:643bb15f` (`npm run check`) | `643bb15f` | **PASS**: eslint `--max-warnings=0` clean, `tsc --noEmit` clean, unit **267/267** (incl. the new Option B case), 0 fail/skip (`a5/b3-check.log`). A first attempt with an added `--read-only` root failed only at `tsc` writing `tsconfig.tsbuildinfo` (EROFS; lint had passed) — my own extra flag, dropped. |
| **b4i run 1** | private identity stack `docker compose -p g060-a5-b4i … up -d --wait --no-build --pull never` (local + identity + migrate + browser + `a5/b4i/override.yml`), `migrate up` (→ 17), `a5/seed.sh` (mcp-seed as the harness's `mcp_seed`), then `… --profile browser run --rm --no-deps -e IDENTITY_E2E=1 -v identity-run:/identity:ro … web-browser npx playwright test --project=domain-platform --project=domain-custom --project=onboarding-light --project=onboarding-dark --project=membership-light --project=membership-dark --grep "E-MEM\|E-ONB-3\|E-ONB-4\|E-ONB-8\|E-BRD\|S-LANDMARKS\|S-NAV"` | `643bb15f` | **38/40 pass, 2 fail** (E-MEM-3 light/dark). Passed: E-BRD-1..6; E-ONB-3, E-ONB-4, **E-ONB-8** (with the new navigation checks), onboarding S-LANDMARKS/S-NAV, both themes; E-MEM-1, 2, 4..9, **E-MEM-10**, membership S-LANDMARKS/S-NAV, both themes. Every test carries the S-ERRORS annotation (80 = 2 × 40). |
| **b4i rerun** | same stack, browser image `04d897f8`, `--project=membership-light --project=membership-dark --grep "E-MEM-3\]"` | `04d897f8` | **PASS 2/2** |
| **b4r** | private root stack `docker compose -p g060-a5-b4r …` (local + migrate + browser + override; no identity provider), `migrate up` (→ 17), `… run … web-browser npx playwright test --project=shell-light --project=shell-dark --grep "S-ACCOUNT"` | `04d897f8` browser; app images `643bb15f` (identical app source) | **PASS 8/8** (S-ACCOUNT-ANON/UNAVAILABLE/LINK/A11Y × 2) |
| **b3b** | as b3 with `g060-a5-web-check:04d897f8` (`--label g060-a5=b3b`) | `04d897f8` | **PASS**: lint, tsc, unit 267/267 (`a5/b3b-check.log`) |
| s1 | `docker run --rm --network none --label g060-a5=s1 -v …/infra/validation:/v:ro node:24-slim bash -n /v/run.sh` and `… /v/lib.sh` | `04d897f8` | **PASS** rc 0 each |
| scope | `a5/scope.sh`: 39 patterns `919ebace…`; 88 changed paths vs `origin/staging` `b539e615` (committed + untracked) at `643bb15f`; 89 at the Result commit (+ this Result) | `643bb15f`, Result commit | **all in scope** |

First-contact failure and fix (recorded, not hidden): **E-MEM-3** (attempt-4 rewrite) still asserted alice's Owner
row on frank's members page right after the transfer, while the same test (and AC2) asserts that frank, now a plain
Member with the override off, sees no member table — the row could never exist. Fix `04d897f8`: the new owner's own
self card ("Your role in this Group: Owner") proves the transfer; every assertion on frank's view is unchanged. Rerun of
the failed case in both themes PASS (above). No other failure.

Private-run resources (owner direction 2026-10-19): each stack had its own project `g060-a5-<run>`, label
`g060-a5=<run>` on every service, network and volume, internal project networks, generated credentials (`a5/secrets/<run>.env`,
0600, deleted after the run), no published port on any service (override resets every `ports:`; the browser runner
shares web's namespace), read-only source mounts only, prebuilt task images with `pull_policy: never`. Test-pool usage:
**one run at a time** (b4i, then b4r; b3/b3b/s1 are network-less single containers without a database); never more than
one of the two reserved slots in use.

### Option B privacy cases (E-MEM-10 + unit)

| Case | Path | Presentation |
|---|---|---|
| Wrong authenticated account (dave on alice's link) | core `GET /invitations/{id}` → closed not_found | the fixed sentence, 200, no form |
| Nonexistent id (random UUID) | core not_found | identical |
| Malformed id (`not-an-invitation`) | no core call | identical |
| Organization callback mismatch | `/join?status=invitation_mismatch` (unchanged redirect) | the same fixed sentence (unit) |
| Intended recipient (alice) | unchanged accept form and POST | `joined`; member row Member |

No case names the intended account/email or distinguishes existence; `main` text is byte-identical across the three
browser cases.

### Pending lock-requiring harness checks (for the coordinator's merged-SHA local full)

Not run here (they run inside `infra/validation/run.sh` with its shared lock, fixed `PORTS_IDN`/harness ports or shared
stacks): B-IDENTITY-E2E (full identity:browser incl. the new matrix entry E-MEM-10 through `browser_matrix`),
B-BROWSER-ROOT (shell projects incl. S-ACCOUNT through the README command), B-WEB-BOUNDARY (static; the new import of
the server-only `lib/enrollment-client.ts` by the invitation page was inspected: the page is a server page, no
`request-context`, no client/component import), B-UNIT-CHECKS web-check, B-TENANCY-HTTP, B-MCP-SOURCE,
B-MIGRATE-SET/CLI/CATALOG/IT, B-AUDIT-CATALOG/IT, B-RLS-IT/ORDERING, B-POLICY-IT/ORDERING, B-ONBOARDING-IT,
B-MEMBERSHIP-IT (+TestAuthority), B-ENROLLMENT-IT, B-RECOVERY-DRILL, B-SCOPE, B-REQUIRED-CHECKS. The four record-text
edits (B-POLICY-IT, B-RLS-IT, B-AUDIT-IT, B-ONBOARDING-IT) change prose only, not conditions.

Reused with attribution (core tree `3caebabf` unchanged throughout attempt 5, so no core suite was repeated): the
attempt-3 core/catalog evidence as stated in its sections, the attempt-4 committed OR-3/6/7 and enrollment proofs (or1
`909f0ad1`, e1 `761c990c`), the attempt-4 tools/logcheck runs (events.json unchanged).

### Deviations for disposition (attempt 5)

- **D-A5-1** E-MEM-3 assertion correction (test-only, in-scope path; above). Intent preserved; it now verifies the
  transfer from the new owner's view.
- **D-A5-2** S-ACCOUNT needs the ordinary stack without an identity provider, so it ran on a second private root stack
  (b4r) after the identity stack was removed (sequential, one pool slot).
- **D-A5-3** Staging advanced (`b539e615`, governance/management only) after the first rebase; the second rebase needed
  one more lease push (`--force-with-lease=g060-impl:5f2c8272`); later pushes plain.
- **D-A5-4** MinIO images reused by verified ID instead of rebuilt (as attempts 3–4); all other stack images are
  task-built from the committed checkout.
- **D-A5-5** E-MEM-10 added to `BROWSER_MATRIX_IDENTITY_E2E` in `infra/validation/lib.sh` so the final qualification
  requires it (approved path; minimum mapping change, no new ID family or CI trigger).

### Native refusals and prompts

None. The first Write under `a5/` (`a5/scope.sh`), the first `docker build` (`docker build --label g060-a5=b3 -t
g060-a5-web-check:643bb15f -f …/infra/docker/web.Dockerfile --target check …/web`) and the first `docker compose`
(`docker compose -p g060-a5-b4i --env-file …/a5/secrets/b4i.env -f …/local.yml -f …/identity.yml -f …/migrate.yml -f
…/browser.yml -f …/a5/b4i/override.yml up -d --wait --no-build --pull never …`) each ran **without a prompt**; the
build outcome was reported to the coordinator at once.

### Cleanup

`a5/verify-clean.sh`: b4i and b4r each `containers=0/0 networks=0/0 volumes=0/0 secrets=0` (label/name), after `down
-v --remove-orphans`; b3/b3b/s1 containers were `--rm`. All eight task images removed (`docker image rm g060-a5-…`);
**no task image kept**. `a5/secrets` empty. The browser reporters' observed-token lists were checked against the reports
and failure artifacts (0 matches) and deleted. Kept evidence under `platform-evidence/g060-s12/a5/`: scripts, b3/b3b
logs, both b4i reports (run 1 incl. E-MEM-3 failure screenshots, rerun), b4r report, overrides, `cleanup.log`. Two
`/tmp` files from attempt 4 are not this attempt's.

### Budget

B5 attempt 5 of 5 consumed (final; no attempt 6, no reset). Remaining after review: Codex pre-integration review, MR
fast/Duo, authorized merge, the merged-SHA local full (incl. the pending list above) and cleanup.

## Merged-SHA local full qualification — FAILED (terminal, 2026-10-20)

Recorded by the coordinator at Codex's request before any correction dispatch. The run is preserved as **failed**. No
rerun, code or test change, check relaxation or attempt-6 dispatch has happened. B5 attempts used: 5 (budget cap 7 per
the recorded extension); 2 remain, not reset.

### Run identity

- Source: staging merge `2d55e37ff33fc75f5ed8096293a1d0411ef0576d` (MR !59; parents `8146abb8` + `92f7369c`).
  Coordinator checkout `.claude/worktrees/0-agent-5c6f436`, detached and clean at launch;
  `bash infra/ci/docker-desktop-mounts.sh` ran first.
- Command (background, shared harness lock):
  `flock -n <home>/.local/state/platform-evidence/local-harness.lock bash infra/ci/validate.sh --profile full --evidence-dir <home>/.local/state/platform-evidence/g060-s12/postmerge-full-2d55e37f/evidence > …/postmerge-full-2d55e37f/run.log 2>&1`
- Selection: profile full, executed full, reason local-explicit. Suites: lint, lint-base, lint-web, lint-core,
  controls, ci-controls, provenance, contracts, root, migrations, ops, rename, off. Coverage and partition complete;
  harness deadline 14400 s.
- Timing: started 2026-10-20T00:47:22Z; run.log closed 02:23:44Z (about 96 min). The usual local full takes about
  70 min; most of the overrun is the B-OBJ-ORDERING 1800 s timeout.
- Outcome: `validate: final status 1`. Phases: harness=1, complete=14 ("required check B-OBJ-ORDERING is not PASS"),
  every other phase 0; published=none, registry removed. Cleanup verified (0 resources removed; owned_remaining=0).
  Lock released at exit.
- Check tally in run.log: 114 PASS, 6 FAIL, 7 INFO. Supplementary JUnit: 123 cases, 8 failing (the JUnit file is not
  retained, so the 8 case names are **unavailable**).

### Evidence retained

- `…/g060-s12/postmerge-full-2d55e37f/run.log`, SHA-256 `bcc2ec07bfddb6ab0f5f511066af9b5c6d055c3d8273f241b2485a70be9fee0f`.
- `…/postmerge-full-2d55e37f/evidence/storage.tsv`.
- Nothing else. A local (non-CI) run never publishes, and its scrub phase removed `evidence/raw` (harness-console.log,
  report/*.jsonl), the JUnit file and the candidate.
- The details below come from coordinator reads of `evidence/raw/harness-console.log` and `evidence/raw/report/*.jsonl`
  taken while the run was live (about 02:08–02:15Z), and from the coordinator checkpoint
  (`platform-evidence/coordinator-checkpoint-20261018.md`). They are transcriptions, not retained files.

### Failure inventory

| Row | Harness summary line (verbatim) | Retained detail | Unavailable |
|---|---|---|---|
| B-POLICY-IT | `[21:10:56 local] FAIL B-POLICY-IT - policy model (FAIL/39 policy, FAIL/4 footprint subtests; logcheck: ok)` | Package result `FAIL core/internal/tenancy/access 6.520s`. `TestAccessPolicy` printed `--- PASS (3.91s)` with every listed IT-CTX/MATRIX/PROFILE/FLAG/FACADE/REVOKE/MFA/LOG subtest PASS. The harness then printed `E-PG-1: 8 new escalation trigger line(s): return to Codex (policy-epg.jsonl)`. policy-epg.jsonl had 113 records; the 8 `E-PG-1-TRIGGER` records were 4 × `business_profile.update` and 4 × `group.rename`, all reasons `["a sequential scan in statements 4-7"]`, step `OB`, mode `org-only`. The environment: cardinality N_g 51, b_org 4000, groups_x 200, m_org 400, n_g 1; 16 CPUs; PostgreSQL 17.11 in the harness check container. | Which Go test or subtest made the package FAIL: no `--- FAIL` line appeared in the console text I read. The 4 failing footprint subtest names. The full plans. |
| B-RLS-IT | `[21:12:34] FAIL B-RLS-IT - row-level security isolation (FAIL/10 isolation, FAIL/7 pool, footprint pass, 6/6 identity pool, race pass; scope {"cases":40,"disagreements":0}; logcheck: ok)` | Package result `FAIL core/internal/tenancy/access 4.704s`; `TestRLSFootprint --- PASS (2.27s)`. The harness printed `E-RLS: 40 escalation trigger line(s): return to Codex (rls-erls.jsonl)`. rls-erls.jsonl had 314 records; the 40 `E-RLS-TRIGGER` records were 9 × business_profile.update, 9 × business_profile.get, 8 × group.rename, 8 × group.change_slug, 4 × group.get and 2 × organization.get. All carried the reason "reader calls exceed twice the rows all of the statement's scans examined (descriptive heuristic, not the Plan's candidate-row term; unresolved observation)"; the examples had scanned_rows 1 and reader_calls 3. Steps were G and Gt; modes org-only, plus selected for organization.get and group.get. An example `E-RLS-3` (group.get, selected, s50, n_g 50) had statements 7 and function_calls ctx_group 15, ctx_org 26, ctx_user 10, rls_group_role 1, rls_org_role 5. An `E-RLS-SUBPLAN` record showed organization.get step U with SubPlan 17 over organizations and role_bindings, seq_scan true, loops 0. | The failing isolation (10) and pool (7) subtest names and assertions. |
| B-OBJ-ORDERING | `[21:44:18] FAIL B-OBJ-ORDERING - stored-object ordering (FAIL/8 subtests; logcheck: ok)` | Package result `FAIL core/internal/tenancy/access 1800.008s`: the Go test hit its 30-minute timeout. The trace tail showed a goroutine `core/internal/tenancy/access.start.func1()` at `ordering_integration_test.go:86`, created by `access.start` at `ordering_integration_test.go:84`. The RLS-ACCESSLOG lines around the end (01:44:27Z) show repeated `object.delete` deny with reason `target_not_found` (stage t1) interleaved with `object.put` allow. Before that, the access log had nothing between 01:14:17Z and 01:44:27Z, consistent with a 30-minute stall. This is the required row named by phase complete=14. | The `panic: test timed out` header, the running-test list and the full goroutine dump (truncated in the console). The names of the 8 subtests. |
| B-OBJ-FAULTS | `[21:44:35] FAIL B-OBJ-FAULTS - stored-object faults (FAIL/19 subtests; logcheck: ok)` | It ran right after the ordering timeout. E-OBJ, printed earlier at 21:14:02, said `E-OBJ: 2 escalation trigger line(s): return to Codex (obj-eobj.jsonl)`: 2 × `E-OBJ-TRIGGER` on `object.put T1`, reason "statement above its hypothesized reader bound". obj-eobj.jsonl had 76 records. B-OBJ-IT itself passed (10/10). | The names and assertions of the 19 subtests, and whether they are a downstream effect of the ordering stall or independent. |
| B-IDENTITY-PRIVACY | run.log: `validate: check B-IDENTITY-PRIVACY FAIL` | Reported after B-IDENTITY-E2E (PASS at 22:10:10 local) and B-IDENTITY-PROBES (PASS); B-IDENTITY-IT, B-IDENTITY-GUARDS, B-IDENTITY-CATALOG and B-IDENTITY-CONFIG passed. Its console line was written after my last live read (about 02:10:52Z), so I never saw it. | **Everything else**: the summary line, the failing privacy condition, the matched token class and the source log. The scrub removed it. |
| B-REQUIRED-CHECKS | run.log: `validate: check B-REQUIRED-CHECKS FAIL` | Downstream aggregate of the rows above. | — |

Times in `[..]` are host local (UTC−4), as printed by the harness.

### Checkout-integrity incident during the run (disclosed)

- **What happened.** At 02:01:16Z a G020 planner, which I had dispatched without worktree isolation (its EnterWorktree
  call was natively refused), ran `checkout 40792883` in this frozen checkout. It committed `c8ab9e4a` at 02:09:55Z on
  local branch `g020-plan-r1`.
- **Delta.** The only difference from `2d55e37f` was under `management/` (G020 Plan files plus management commits up to
  `c1e724c6`). There was no core, web, infra, contracts or docs change.
- **Restore.** I restored the detached `2d55e37f` (clean) at about 02:10:45Z.
- **Rows that finished inside the window (all PASS):** B-OBS-CONTEXT, B-OBS-CHAIN, B-OBS-TIMEOUTS, B-OBS-LOGS,
  B-OBS-PRIVACY, B-OUTAGE, B-PERSIST, B-SECRET-ROTATION, B-IDENTITY-IT and B-IDENTITY-E2E.
- **Timing of the failed rows.** The five failed rows all completed before the window, except B-IDENTITY-PRIVACY, which
  reported after the restore. Its relation to the window is **unknown**.
- **Pending.** Codex's attribution disposition. The root cause was a coordinator dispatch error; the corrective rule is
  that writing agents always get worktree isolation.

### State and next step

G060 is merged and **not qualified**. Its acceptance stays conditional. Waiting on Codex's bounded disposition of the
E-PG-1, E-RLS and E-OBJ escalation triggers, the ordering timeout, the 19 fault subtests, the identity-privacy gap and
the window attribution. Attempt 6 waits on that disposition. Repairs remain with the G060 owner, through a fresh worker.

## Attempt 6 (B5 attempt 6 of cap 7) — focused diagnosis and bounded corrections, 2026-10-20

Released by Codex's disposition (review record `G060-plan-codex-review.md` SHA-256 `89e7264f…a886` at `origin/staging`
`adf45a13`): reproduce every failed row focused and preserve its output before editing; corrections only inside the
39-pattern scope allow-list (`919ebace…cb17`); no threshold, oracle or check change; scope changes return to Codex before
editing; no unchanged full-suite rerun.

### Worker record (attempt 6)

- Isolated worktree `.claude/worktrees/agent-<id>`, branch `g060-attempt6` from `origin/staging`
  `adf45a13` (asserted after the switch), rebased once onto `f4e12c15` (management/`.codex/config.toml`/AGENTS.md only;
  no core, web, infra or contracts delta) before the correction batch; the fetched target was an ancestor of HEAD before
  every batch. Staging's `core/`, `web/`, `contracts/` trees equal the merged `2d55e37f`; `infra/` and `docs/` differ
  only by the owner-approved Vestive canonical reversal `d0b06691` (harness path constants, eslint comment), which no
  failed row exercises.
- Evidence (durable, outside the checkout): `platform-evidence/g060-s12/a5/a6/` — `b6.sh` (runner, adapted from
  attempt 3's `b2.sh`; SHA-256 `027aab6d…`), `b6-<run>.log` (full `go test -v` output), `b6-<run>.pg.log` (PostgreSQL
  log), `b6-o.pgstat.log` (20 s `pg_stat_activity` samples), `logs/<run>/` (access log + the run's canaries),
  `scratch-core/` (a copy of `core/` carrying the proposed patches, see D-A6-2), `commit1.txt`/`commit2.txt`.
- Private per-run stacks (owner direction 2026-10-19): per run one internal network, PostgreSQL `pgvector/pgvector:0.8.7-pg17`
  on tmpfs initialised by `infra/postgres/init`, the pinned MinIO images by verified ID for the object rows, generated
  credentials (0600, deleted at exit), no published port, read-only source and module cache (`g058-a1-gomod`,
  `-mod=readonly`, `GOPROXY=off`), disposable `GOCACHE`, toolchain `golang:1.27.2-alpine` (= the core check image's
  base). Names `g060-a5-a6-<run>-{net,pg,minio,go}`, label `g060-a5=a6-<run>`. Never more than two stacks live; no
  image built; no harness lock taken. `bash infra/ci/docker-desktop-mounts.sh` first (16% of the mount table).

### Reproduction (source `adf45a13`, core tree = merged `2d55e37f`; before any edit)

| Run | Selector (exact harness selector of the row) | Outcome | Evidence (SHA-256) |
|---|---|---|---|
| p | `-timeout 15m -v -run '^(TestAccessPolicy\|TestAccessFootprint)$' ./internal/tenancy/access/` | **FAIL** 5.2 s: `TestAccessPolicy` PASS (39/39); `TestAccessFootprint` FAIL — `E-PG-2 onboarding writes (G018)`: `footprint_integration_test.go:386: x_admin group.onboarding.complete_existing: access: forbidden`; `E-PG-MEM membership operations (G019)`: `footprint_integration_test.go:486: owner invitation.issue_group: access: forbidden`. E-PG-1 and E-PG-2 locked-row inventory PASS. 8 E-PG-1 trigger lines (the known sequential scans, statements 4–7). | `b6-p.log` `a4fcccec…` |
| r | `-timeout 15m -v -run '^(TestRLSIsolation\|TestRLSPool\|TestRLSFootprint)$' ./internal/tenancy/access/` | **FAIL** 3.9 s: `TestRLSIsolation` PASS 10/10 (scope 40 cases, 0 disagreements); `TestRLSFootprint` PASS; `TestRLSPool` FAIL — `P3/P4/P5`: `rls_integration_test.go:1541: commit: access: forbidden`; `P8`: `rls_integration_test.go:1633: lost reply: access: forbidden, want "unavailable"`. 40 E-RLS-TRIGGER lines (the descriptive reader-call heuristic). | `b6-r.log` `eababe0a…` |
| o | `-timeout 8m -v -run '^TestObjectsOrdering$' ./internal/tenancy/access/` (MinIO; `PGSTAT=1`) | **FAIL** 480 s = the Go timeout: `OOR-3 Delete D1 versus Put T2, both orders`: `objects_ordering_integration_test.go:245: fewer than 1 lock waits observed`; `OOR-8`: `objects_ordering_integration_test.go:438: delete: {Outcome:} access: not_found`; OOR-1, 2, 4, 5, 6, 7 PASS. Then `panic: test timed out after 8m0s` with the dump: goroutine 218 (`start.func1` ← `objEnv.put` ← `PutObject` ← `storageTx` ← `hookTx.Commit` ← `pause.hook`, `ordering_integration_test.go:60`, "chan receive, 7 minutes") and goroutine 20 (`t.Cleanup` → `pgxpool.Pool.Close` → `puddle.Pool.Close` → `WaitGroup.Wait`, 7 minutes). `pg_stat_activity`: one `g060it_runtime` backend `idle in transaction` on T2's `UPDATE app.stored_objects SET state = 'ready' …` from the first sample to the end. | `b6-o.log` `d9e0d621…`, `b6-o.pgstat.log` `4fbe76e4…` |
| f | `-timeout 15m -v -run '^TestObjectsFaults$' ./internal/tenancy/access/` (MinIO) | **FAIL** 8.0 s: 12/19 PASS, **7 FAIL** — FT-UNKNOWN, FT2, FX-CANCEL, FX-CRASH, FC, FD, `pending and deleting objects never serve new reads or signs` — every failure line is `repair delete: {Outcome:} access: not_found` (`objects_faults_integration_test.go:180/364/389/412/438/571`) or, in FD, `FD-D1-UNKNOWN/D2/D3/D4: access: not_found, row ready` and `… retry: {Outcome:} access: not_found` (`:481/:483/:489`). FX-STALL PASS. | `b6-f.log` `9b2ad6df…` |

B-IDENTITY-PRIVACY was **not reproduced**: `identity_privacy` runs only inside the harness's `ops` suite
(`variants_identity`: the identity Compose variant on the fixed `PORTS_IDN`, the run's secrets registry, the
browser-identity results and `vtools`), which needs `local-harness.lock` and coordinator awareness. Per the dispatch
rule it is requested through this report rather than waited on; no cause is inferred (D-A6-3).

### Root cause (one cause, four rows)

Every failing assertion is a `forbidden`/closed `not_found` refusal of an Organization owner or admin acting on a Group
without a Group role, with the Group-management override off — exactly the G060 contract (Spec R2 table: ordinary Group
management by Organization staff needs the override; AC5: Group object read/upload/delete/sign needs an explicit Group
role under either setting). The four tests still carry their pre-G060 fixtures, which relied on the removed implicit
Organization-staff shortcut:

- **B-POLICY-IT** — `TestAccessFootprint`: the X admin (`OrgAdmin`, `GroupMember` of 50 Groups) attaches its existing
  profile through the Organization path (`CompleteGroupOnboardingWithProfile`); the owner `mo` (no Group role) issues a
  Group invitation. Both are Organization-path Group management, refused without the override. `observe`/`run` use
  `t.Fatalf`, so later operations of those subtests are unobserved until these are repaired.
- **B-RLS-IT** — `TestRLSPool`: the Organization owner (no Group role) renames `g3` in P3 (commit), P7 and P8; the
  rename is refused before any transaction, so P3 fails outright and P8 sees `forbidden` instead of the lost-reply
  `unavailable`. P7's cancellation was unaffected only because the refusal arrives before its hook.
- **B-OBJ-ORDERING** — `OOR-3` deletes as `w.admin` (Organization admin, no `g1` role): the delete is refused at
  Resolve, so no row-lock wait is ever observed (`:245`, fatal); the fatal leaves `pPut.release` unclosed, the paused
  putter keeps its pooled connection in T2 (the `idle in transaction` backend), and the package-level
  `t.Cleanup(pool.Close)` waits for it forever — that is the 1800 s harness timeout, not a database lock. `OOR-8`
  deletes as `w.admin` likewise (`:438`).
- **B-OBJ-FAULTS** — `repairDelete` (`:105`) and the FD cases (`:477`, `:486`) delete as `w.admin`; every "repair" is
  refused `not_found`. The 7 failures are independent of the ordering stall (same cause, not downstream).

The E-PG-1 (8), E-RLS (40) and E-OBJ trigger lines are `t.Logf`/stdout observations of unchanged instrumentation and
were not involved in any failure; no threshold or oracle was touched.

### Correction 1 (in scope, committed): `core/internal/tenancy/access/rls_integration_test.go`

`TestRLSPool` turns the override on through the existing fixture helper `e.withOverride(t, w.owner, w.A)` right after
`w := e.world(t)` (before the pool's first acquisition; the helper uses the fixture service, not the MaxConns=1 pool
under test, and turns the override off at the test's end). No assertion changed: P1/P2 same-backend checks, P4's member
denial, P5/P6/P7/P9/P10 and P8's `unavailable` + unchanged `g3` name are as before. Commit `72b91502` (rebased from
`fa120488`), 3 lines.

Focused reruns on the frozen committed `72b91502` (clean checkout, `origin/staging` `f4e12c15` an ancestor):

| Run | Selector | Outcome | Evidence |
|---|---|---|---|
| r2 | the B-RLS-IT access selector (above) | **PASS** 3.7 s: `TestRLSIsolation` 10/10, `TestRLSPool` 7/7, `TestRLSFootprint` PASS; `RLS-SCOPE {"cases":40,"disagreements":0}`; 40 E-RLS-TRIGGER lines (unchanged observation) | `b6-r2.log` `429b5287…` |
| r2i | `-run '^(TestRLSIdentityPool\|TestRLSFirstSignInRace)$' ./internal/identity/` (the row's second package, control) | **PASS**: 6/6, race PASS | `b6-r2i.log` `86d14bdf…` |

Not run here: the row's `logcheck validate/scan` of the access records and the planted-canary control (harness-internal
`vtools`); the access log and canaries of r2 are kept under `a6/logs/r2/` for the coordinator's merged-SHA full.

### Deviations for disposition (attempt 6)

- **D-A6-1 Scope question (blocking the other three rows).** `footprint_integration_test.go`,
  `objects_ordering_integration_test.go` and `objects_faults_integration_test.go` are **not** in the 39-pattern
  allow-list (checked against `a5/allow.re`: only `rls_integration_test.go` and `integration_test.go` of the five
  files match), so per disposition §4 they were not edited in the checkout. The proposed, same-class fixture
  corrections (no assertion, threshold or oracle change) are staged in `a6/scratch-core/internal/tenancy/access/` for
  Codex to diff against `core/`:
  - `footprint_integration_test.go`: `e.withOverride(t, fw.owner, fw.X)` at the start of
    `E-PG-2 onboarding writes (G018)` (s1/s50 are members, unaffected) and `e.withOverride(t, mo, fw.X)` after
    `e.bind(t, acc, fw.X)` in `E-PG-MEM membership operations (G019)`; 2 inserted statements + comments.
  - `objects_ordering_integration_test.go`: OOR-3 and OOR-8 delete as `w.g1admin` (the fixture's `GroupAdmin` of `g1`)
    instead of `w.admin`; 2 lines.
  - `objects_faults_integration_test.go`: `repairDelete` and the two FD deletes use `w.g1admin`; 3 lines.
  Alternative for the footprint rows: Group-role grants instead of the override (changes the dataset's binding counts).
  Requested: extend the allow-list by exactly these three paths (or direct a different repair), then a bounded attempt
  (6 continued or 7) applies, commits and reruns the three selectors. Until then B-POLICY-IT, B-OBJ-ORDERING and
  B-OBJ-FAULTS remain failing on the merged SHA. The observation that `OOR-1`'s admin read (`:139`) is now vacuous
  under G060 (not_found either way) is noted, not changed.
- **D-A6-2 Proposed patches unverified.** Running the patched scratch tree (`CORE_DIR=…/scratch-core`, run `sp`) was
  refused by the native auto-mode classifier ("Security Test Removal"); so was writing `diff -u` files of the scratch
  tree. Neither was retried or routed around; the scratch tree itself is the proposal. The footprint subtests may hit
  further pre-G060 expectations after these first fatals (each `t.Fatalf` hides later operations).
- **D-A6-3 B-IDENTITY-PRIVACY** needs the lock-requiring identity variant (above). Static inspection found no
  uncatalogued event in the G060 delta (`access.override_used` is an audit action, not a log event; `allowed_override`
  is catalogued in `events.json`); this is context only, not a cause.
- **D-A6-4** The imported Result version (`49a2c35b` on `origin/g060-result-fullrun`, the failure-inventory section)
  is carried on this branch unchanged, since `origin/staging` lacks it.

### Native refusals and prompts (quoted once, not retried)

Two classifier refusals, both "Security Test Removal": (1) `bash …/a6/b6.sh sp …` with `CORE_DIR` pointing at the
patched scratch tree; (2) the `diff -u` loop writing `a6/proposed-*.diff`. Every other command (private stacks,
reruns on the checkout, commits, rebase) ran without a prompt.

### Cleanup receipt (attempt 6)

Per run `CLEANUP label=g060-a5=a6-<run> containers=0 networks=0 volumes=0 secrets=0` for p, r, o, f, r2, r2i
(`b6-*.log`); verified after the last run: 0 containers/networks/volumes matching `g060-a5-a6`, 0 `g060-a5*` images
(none built), `a6/secrets/` empty. Removed by name: `g060-a5-a6-{p,r,o,f,r2,r2i}-{go,pg}`, `g060-a5-a6-{o,f}-minio`,
networks `g060-a5-a6-{p,r,o,f,r2,r2i}-net`. Kept: the evidence files above and `a6/scratch-core/` (the proposal; delete
at disposition). The original `postmerge-full-2d55e37f/run.log` is untouched.

### Budget

B5 attempt 6 of 7 consumed; attempt 7 reserved. G060 stays unqualified until the exact merged SHA passes the
coordinator's local full; this attempt changes no acceptance gate.

## Attempt 6 correction round (same B5 attempt 6 of cap 7) — 2026-10-20

Status: **submitted for Codex review** (not accepted). Released by Codex's binding approval of D-A6-1: **scope
amendment 01** at `origin/staging` `7d2d6221` adds exactly `footprint_integration_test.go`,
`objects_ordering_integration_test.go` and `objects_faults_integration_test.go` (fixture-only). Canonical allow-list
`management/v0/results/G060-scope-allowlist.txt`: **42 patterns**, SHA-256
`5e5c11720a1fa1a61d6a674bfd0721549b430ef6cd508f62d1f04e45e9fd2b56` (identical to `origin/staging`). No production
code, assertion, expected outcome, threshold, oracle or unrelated fixture changed. This round is a continuation of
attempt 6, not attempt 7.

### Worker record (correction round)

- Single write owner, model Opus 5.5 (effort inherited; exact runtime value unknown). Isolated worktree
  `.claude/worktrees/agent-<id>`, local branch `g060-a6-correction` created from `origin/g060-attempt6`
  `d27c77dc` (asserted). The local `g060-attempt6` branch is checked out in the previous writer's locked worktree
  `agent-<id>` (left untouched), so this branch publishes with `HEAD:g060-attempt6`.
- Rebased onto `7d2d6221`, then, after a pre-batch fetch showed staging at `7afee895` (management, AGENTS.md and
  CLAUDE.md Codex model-default prose only), onto `7afee895`. Pushed with
  `--force-with-lease=g060-attempt6:d27c77dc…` → `4a939840`. `origin/staging` `7afee895` was an ancestor of HEAD and
  the checkout was clean before the test batch.
- Bindings versus the merged `2d55e37f`: `core/` differs only by the four fixture files of attempt 6 (this round's
  three plus `rls_integration_test.go`); `contracts/` identical; `web/` differs only by `web/eslint.config.mjs` and
  `infra/` and `docs/foundation`/`docs/ARCHITECTURE.md` only by the Vestive canonical reversal `d0b06691`. Precision
  added to the attempt-6 record: `docs/agents/CLAUDE-SESSION-HANDOFF.md` also differs (agent-handoff prose of
  `30a6d497`, `d1fe56f0`, `86df5a69`). None of these is exercised by the rows below.
- Evidence: `platform-evidence/g060-s12/a5/a6/` — runner `b6c.sh` (SHA-256 `e13c1578…a2f1`; `b6.sh` adapted to this
  checkout, no `CORE_DIR` override, refuses to overwrite a run log, outer `timeout` 900 s around Go's own `-timeout`),
  `b6-c{p,f,o}.log`, `b6-c{p,f,o}.pg.log`, `b6-co.pgstat.log`, `logs/c{p,f,o}/` (access log + canaries),
  `commit3.txt`, `c-scope-files.txt`. All earlier attempt-6 files, including the failed-run logs and
  `scratch-core/`, are preserved unchanged.

### Corrections (commit `4a939840`, 3 files, +22/−10)

| File | Change | Cause (attempt-6 diagnosis) |
|---|---|---|
| `footprint_integration_test.go` | `e.withOverride(t, fw.owner, fw.X)` at the start of `E-PG-2 onboarding writes (G018)` and after the credential binds of `E-PG-MEM membership operations (G019)` (the existing helper; turned off when each subtest ends) | The X admin's existing-profile completion and the owner's Group invitation are Organization-path Group management |
| `objects_ordering_integration_test.go` | OOR-3's deleter and OOR-8's delete run as `w.g1admin` (the world's `GroupAdmin` of `g1`) | Group object deletion needs an explicit Group role (AC5) |
| `objects_faults_integration_test.go` | `repair := &objOrderWorld{admin: w.g1admin, org: w.org}` for the six `repairDelete` calls; the FD delete and retry run as `w.g1admin` | Same |

Two deliberate differences from the staged, unverified `scratch-core/` draft (re-derived from the real source):

- **E-PG-MEM override set by `fw.owner`, not `mo`.** `mo` ends the subtest with `TransferOwnership(…, adm2, OrgAdmin)`
  and is then an Organization admin, so the helper's owner-only turn-off at subtest end would be refused (a `t.Fatalf`
  in cleanup). `fw.owner` stays an owner of X; the setting and its effect are the same.
- **`repairDelete` itself unchanged.** `audit_integration_test.go:988` (B-AUDIT-IT) calls it with
  `&objOrderWorld{admin: w.gadmin, org: w.orgA}`, which has no `g1admin`; changing the helper would have broken that
  neighbour. The faults test uses the audit test's existing form at its own call sites, so no shared helper changed.

Static pre-check (disposable `--network none` containers, read-only source, `golang:1.27.2-alpine`, same core content
as the commit): `gofmt -l internal/tenancy/access` printed nothing; `go vet -tags integration ./internal/tenancy/access/`
exit 0.

### Validation (frozen committed `4a939840`, clean checkout; private per-run stacks, no lock)

Each run: `HEAD_SHA=4a939840 bash …/a6/b6c.sh <run> [minio] -- go test args`, executing
`go test -tags integration -count=1 <args>` in `golang:1.27.2-alpine` against its own PostgreSQL
`pgvector/pgvector:0.8.7-pg17` (tmpfs, internal network, no published port) and, for the object rows, the pinned MinIO
images verified by ID. At most two stacks live (cp and cf ran together, then co); no image built.

| Row | Run | Exact args | Exit | Outcome | Evidence (SHA-256) |
|---|---|---|---|---|---|
| B-POLICY-IT | cp | `-timeout 10m -v -run '^(TestAccessPolicy\|TestAccessFootprint)$' ./internal/tenancy/access/` | 0 | **PASS** 8.8 s: `TestAccessPolicy` 39/39, `TestAccessFootprint` **4/4** (E-PG-1 plan summaries, E-PG-2 locked-row inventory, E-PG-2 onboarding writes, E-PG-MEM); x_admin `complete_existing`, owner `invitation.issue_group`, `list.candidates`, `membership.remove_group` and the final `ownership.transfer` all observed; 0 FAIL, 0 SKIP | `b6-cp.log` `ce0ba9a1…6ec7` |
| B-OBJ-FAULTS | cf | `-timeout 10m -v -run '^TestObjectsFaults$' ./internal/tenancy/access/` (MinIO) | 0 | **PASS** 8.1 s: **19/19** incl. FT-UNKNOWN, FT2, FX-CANCEL, FX-CRASH, FC, FD and the pending/deleting case (the 7 attempt-6 failures) | `b6-cf.log` `e2e7b709…4a8f` |
| B-OBJ-ORDERING | co | `-timeout 8m -v -run '^TestObjectsOrdering$' ./internal/tenancy/access/` (MinIO, `PGSTAT=1`) | 0 | **PASS** 4.9 s: **8/8** (OOR-1..8); OOR-3 both orders observed its lock wait; no stall, no timeout | `b6-co.log` `e9ff6cbc…8b5a`, `b6-co.pgstat.log` `d8063eb0…d095` |

PostgreSQL logs: `b6-cp.pg.log` `777ac7ae…de96`, `b6-cf.pg.log` `899cb26f…110b`, `b6-co.pg.log` `0c1906cf…10a5`.

Controls and reuse:

- **B-RLS-IT**: not rerun; the attempt-6 PASS is reused (`b6-r2.log` `429b5287…` on `72b91502`: `TestRLSIsolation`
  10/10, `TestRLSPool` 7/7, `TestRLSFootprint` PASS; `b6-r2i.log` identity pool 6/6 + race). Attribution: `core/`
  between `72b91502` and `4a939840` differs only by this round's three files, which define no helper the RLS tests
  use (`withOverride`, `repairDelete`, `deleteAs`, `orderWorld` unchanged); the package compiles (vet and the three
  runs).
- **B-OBJ-IT / B-AUDIT-IT**: not rerun, because no shared helper was edited (the edits are confined to the three
  tests' bodies; see the `repairDelete` note).
- Not run here (harness-internal `vtools`): each row's `logcheck validate/scan` of the access records and the
  planted-canary control. The access logs and canaries are kept under `a6/logs/c{p,f,o}/`.
- **B-IDENTITY-PRIVACY: UNVERIFIED.** It needs the coordinator's lock-requiring identity variant. No pass is inferred.

Trigger observations (instrumentation and thresholds untouched): E-PG-1 printed 8 `E-PG-1-TRIGGER` lines, the same
class as before (subject s50; `business_profile.update` and `group.rename`; steps OB and GB; org-only and selected;
"a sequential scan in statements 4-7"; `"known":false`). The harness logs these for Codex and does not fail the row on
them. E-RLS (40 lines, r2) is unchanged. E-OBJ's 2 lines come from B-OBJ-IT's `TestObjectsFootprint`, which was not
rerun. All three remain open for Codex's disposition.

Residual observation, not changed (outside the approved correction): three Organization-admin **reads** still pass but
now prove less. Under G060 their asserted `not_found` comes from the missing Group role rather than from object state:
OOR-1 `objects_ordering_integration_test.go:139`, FC `objects_faults_integration_test.go:435` and the pending/deleting
check `:559`. A possible follow-up, for Codex to decide: run them as `w.g1admin` so they again prove the lifecycle
guard.

### Failure inventory status (merged-SHA rows)

| Row | Status after this round |
|---|---|
| B-POLICY-IT | Focused **PASS** (cp). Harness logcheck/canary pending the coordinator's full run |
| B-RLS-IT | Focused **PASS** (attempt-6 r2, reused with attribution above) |
| B-OBJ-ORDERING | Focused **PASS** 8/8 (co); the 1800 s stall cause (OOR-3 fatal leaving T2 paused) is removed |
| B-OBJ-FAULTS | Focused **PASS** 19/19 (cf) |
| B-IDENTITY-PRIVACY | **UNVERIFIED**: needs the lock-requiring identity variant (coordinator) |
| B-REQUIRED-CHECKS | Aggregate; pending the coordinator's local full on the merged SHA |

These are focused results on branch source `4a939840`, not qualification. G060 remains merged-and-unqualified until
this correction is integrated and the coordinator's exact merged-SHA local full passes.

### Native refusals and prompts

None. Every edit used the Edit tool. The three runs, the static checks, the commits, the rebases and the lease push
each ran without a refusal.

### Cleanup receipt (correction round)

Per run `CLEANUP label=g060-a5=a6-<run> containers=0 networks=0 volumes=0 secrets=0` for cp, cf and co. Verified
after the last run: 0 containers labelled `g060-a5`, 0 networks named `g060-a5-a6*`, `a6/secrets/` empty, and no
image built. Removed by label/name: `g060-a5-a6-{cp,cf,co}-{pg,go}`, `g060-a5-a6-{cf,co}-minio` and networks
`g060-a5-a6-{cp,cf,co}-net`. The `--rm` static-check containers `g060-a5-a6-cfmt-go` and `g060-a5-a6-cvet-go` exited
and were removed. Kept: all evidence files listed above, every earlier attempt-6 log (including the failed runs) and
`scratch-core/` (for deletion at Codex disposition).

### Budget

Still B5 attempt 6 of 7 (this correction round continues it under Codex's binding approval). Attempt 7 is reserved.

## Residual finding (Codex review of f361842e) — same B5 attempt 6 of cap 7, 2026-10-20

Status: **submitted for Codex review** (not accepted). Codex's binding finding on `f361842e`: three read-denial probes
still ran as `w.admin`, an Organization admin with no Group role. Their `not_found` could therefore come from the
missing Group authority instead of the object lifecycle state. Directed fix: run them as the existing `w.g1admin`
fixture. Assertion text and expected outcomes stay unchanged. No new paths, production code, threshold, oracle or
unrelated fixture change. This continues attempt 6 and does not start attempt 7.

### Worker record (residual finding)

- Single write owner, model Opus 5.5 (effort inherited; exact runtime value unknown). Isolated worktree
  `.claude/worktrees/agent-<id>`, local branch `g060-a6-residual` created from `origin/g060-attempt6`
  `f361842e` (branch and HEAD asserted after the switch). The local `g060-attempt6` branch is still checked out in
  `agent-<id>` (left untouched), so this branch publishes with `HEAD:g060-attempt6`.
- No rebase. The pre-batch fetch showed `origin/staging` still at `7afee895`, an ancestor of HEAD
  (`git merge-base --is-ancestor` exit 0). The checkout was clean before the batch. There was no target delta to
  reconcile. The pre-push fetch then showed staging at `bd0b43ae` (`management: sequence RC3 after G060 acceptance`,
  `management/v0/PROGRESS.md` and `ROADMAP.md` only). That is not a code delta and touches no branch path, so the branch
  was not rebased and was pushed as a fast-forward of `f361842e`.
- Inputs: this Result before the section, SHA-256 `4776549baaecfe13d9d93aaf34724be45d89a3aad44176f6292f6701aed43570`;
  allow-list `G060-scope-allowlist.txt`, 42 patterns, SHA-256 `5e5c11720a1fa1a61d6a674bfd0721549b430ef6cd508f62d1f04e45e9fd2b56`
  (unchanged). Branch delta versus `origin/staging` (`a6/r-scope-files.txt`, `f52f0357…6282`): 5 paths, 0 outside the
  allow-list (`grep -Evc -f`).
- `bash infra/ci/docker-desktop-mounts.sh` ran first: 16% of the mount table in use.

### Correction (2 files, 3 lines)

| File | Probe | Change |
|---|---|---|
| `objects_ordering_integration_test.go` | OOR-1 compensated-object metadata read (`:139`) | `beginAs(…, w.admin, …)` → `w.g1admin` |
| `objects_faults_integration_test.go` | FC leftover-bytes metadata read (`:435`) | same |
| `objects_faults_integration_test.go` | pending/deleting `check` metadata, get and sign probes (`:559`) | same |

Only the actor and the line comment (now citing G060 AC5) changed. `w.g1admin` is the `GroupAdmin` of `g1` in each
world: OOR-1's own `o1` world removes only `m1`, and the faults world never changes `g1admin`'s role. No other
`w.admin` read remains in these two files. The remaining uses are the world setup and the shared `repairDelete` helper
body, which stays unchanged as recorded in the correction round.

Static pre-check (disposable `--network none` containers, read-only source, `golang:1.27.2-alpine`, names
`g060-a5-a6-r{fmt,vet}-go`): `gofmt -l internal/tenancy/access` printed nothing; `go vet -tags integration
./internal/tenancy/access/` exit 0.

### Validation (frozen committed source, clean checkout; private per-run stacks, no lock)

The tested commit is `19cbe674` (the fixture change alone, parent `f361842e`). The pushed commit amends it with this
Result section only, so its `core/` tree `09927363d19af0cada67d24a06e10c1b3e1f4e4b` (access package tree
`eac95ab73ae7a60da525545979696a2aa8d2c9c0`) is the tested tree. Runner `a6/b6r.sh` (SHA-256 `ebcb5d19…65a6`) is
`b6c.sh` with only the bound worktree and the outer wall-clock default (300 s) changed. Go's `-timeout 3m` panics with
the full goroutine dump on a hang, inside an outer `timeout -k 30 300`. The two stacks ran concurrently within the pool
of two (no other repository stack was live). MinIO images were verified by ID; every image already existed by exact
tag; nothing was built.

- `HEAD_SHA=19cbe674… PGSTAT=1 bash …/a6/b6r.sh ro minio -- -timeout 3m -v -run '^TestObjectsOrdering$' ./internal/tenancy/access/`
- `HEAD_SHA=19cbe674… bash …/a6/b6r.sh rf minio -- -timeout 3m -v -run '^TestObjectsFaults$' ./internal/tenancy/access/`

| Row | Run | Exit | Outcome | Evidence (SHA-256) |
|---|---|---|---|---|
| B-OBJ-ORDERING | ro | 0 | **PASS** 5.1 s: **8/8** (OOR-1..8), 0 FAIL, 0 SKIP; no stall (the `pg_stat_activity` sampler saw only itself) | `b6-ro.log` `81b89eb9…721e`, `b6-ro.pg.log` `023315af…ebe8`, `b6-ro.pgstat.log` `559abcdc…b1f0` |
| B-OBJ-FAULTS | rf | 0 | **PASS** 8.1 s: **19/19** incl. FC and the pending/deleting case, 0 FAIL, 0 SKIP | `b6-rf.log` `d231b0d8…5fa1`, `b6-rf.pg.log` `4ecce4a8…669b` |

The access logs show that the probes now hit the lifecycle guard. These are the `storage.operation` records:

- **Before** (correction round on `4a939840`, `logs/co`, `logs/cf`): OOR-1's `object.metadata`, FC's `object.metadata`
  and the pending check's `object.metadata`/`object.get`/`object.sign` were denied with `reason: target_not_found`.
  That is the authority refusal.
- **After** (`logs/ro` `0c7ecf41…726a`, `logs/rf` `62376811…cc14`): the same probes are denied with
  `reason: object_not_ready`. The asserted `not_found` now comes from the object state.

Controls and reuse: B-POLICY-IT and B-RLS-IT were not rerun, because neither their files nor any shared helper changed.
The correction-round focused passes still apply. B-OBJ-IT and B-AUDIT-IT were not rerun, because the edits stay inside
the two tests' bodies. Not run here (harness-internal `vtools`): `logcheck validate/scan` and the planted-canary
control. The access logs and canaries are kept under `a6/logs/{ro,rf}/`.

### Finding disposition

**Resolved in source, pending Codex review.** All three named probes run as `w.g1admin`. Assertions and expected
outcomes are unchanged. Both selectors pass, and the denials are attributed to `object_not_ready`. This closes the
correction round's "Residual observation".

### Failure inventory status (merged-SHA rows, after this commit)

| Row | Status |
|---|---|
| B-POLICY-IT | Focused **PASS** (correction round cp; unchanged) |
| B-RLS-IT | Focused **PASS** (attempt-6 r2; unchanged) |
| B-OBJ-ORDERING | Focused **PASS** 8/8 (ro, tested tree `09927363…`) |
| B-OBJ-FAULTS | Focused **PASS** 19/19 (rf, tested tree `09927363…`) |
| B-IDENTITY-PRIVACY | **UNVERIFIED**: pending the coordinator's lock-requiring harness run (identity variant); no pass is inferred |
| B-REQUIRED-CHECKS | Aggregate; pending the coordinator's local full on the merged SHA |

These focused results are not qualification. G060 stays merged and unqualified until this correction is integrated
and the coordinator's exact merged-SHA local full passes.

### Native refusals and prompts

None. The source edits used the Edit tool. The static checks, both runs, the commit and the push ran without a
refusal.

### Cleanup receipt (residual finding)

Per run: `CLEANUP label=g060-a5=a6-{ro,rf} containers=0 networks=0 volumes=0 secrets=0`. Verified afterwards: 0
containers and 0 volumes labelled `g060-a5`, 0 networks named `g060-a5-a6*`, and `a6/secrets/` empty. The `--rm`
static-check containers exited and were removed. Nothing was built and no image was removed. Kept: the new evidence
(`b6r.sh`, `b6-r{o,f}.*`, `logs/r{o,f}/`, `r-scope-files.txt`, `commit5.txt`) and every earlier file in `a6/`.

### Budget

Still B5 attempt 6 of 7. Attempt 7 is reserved.

## Merged-SHA local full qualification — second run on `f24f0eb9` — ONE ROW FAILED (2026-10-20)

Recorded by the coordinator. Not a qualification: **B-IDENTITY-PRIVACY FAIL**, so B-REQUIRED-CHECKS FAIL. G060 stays unqualified
and not accepted. No code, test, check or threshold change has been made since MR !60; no rerun has been made; budget unchanged
(B5 attempt 6 of cap 7 consumed; attempt 7 reserved).

### Run identity

- Source: staging merge `f24f0eb9` (MR !60 = `g060-attempt6` `66b7969c`; staging parent `fb6632f9`; core tree `09927363…` as validated).
  Coordinator checkout `.claude/worktrees/0-agent-5c6f436`, detached and clean; `bash infra/ci/docker-desktop-mounts.sh` first (16%).
- Command (background, shared harness lock):
  `flock -n <home>/.local/state/platform-evidence/local-harness.lock bash infra/ci/validate.sh --profile full --evidence-dir <home>/.local/state/platform-evidence/g060-s12/postmerge-full-f24f0eb9/evidence > …/run.log 2>&1`
- A first launch exited at once with `--evidence-dir already exists` (the coordinator had pre-created the directory); no check ran
  (`final status 16`, phase `location=16`). The empty directory was removed and the run relaunched on the same clean checkout.
- Timing: about 66 minutes (shorter than the first run: no ordering timeout). Outcome `validate: final status 1`; `published=none`,
  `registry=removed`, cleanup verified (`cleanup_status=0 owned_remaining=0`), lock released at exit. Phases: harness=1, complete=14.
- Tally: 118 rows PASS, 2 FAIL (B-IDENTITY-PRIVACY and the aggregate B-REQUIRED-CHECKS), INFO rows BACKGROUND/BUILD-LEDGER/IMAGES.

### Rows that failed in the first full run and now pass

B-POLICY-IT, B-RLS-IT, B-OBJ-ORDERING, B-OBJ-FAULTS, B-OBJ-IT. Also PASS: B-IDENTITY-IT (88 tests), B-IDENTITY-E2E (133 browser
tests), B-IDENTITY-PROBES, B-MCP-E2E, B-IDENTITY-CONFIG, B-SECRET-CANARY, B-SCOPE and every other row.

### Retained evidence (snapshotted while the run was live; a local run scrubs `evidence/raw`)

- `…/postmerge-full-f24f0eb9/run.log` SHA-256 `71a44b667e294f99cc0fe28c91fb89bd81a4fd071e3efeccfe125312d5c6f984`.
- `…/postmerge-full-f24f0eb9/live-copy/` (private snapshot of `evidence/raw` log/jsonl/tsv files taken every 90 s):
  `harness-console.log` SHA-256 `82de9a53f1bef7fa84bf27087949520f25c69eeb68c6af815c2742e6aa0ef47c`;
  `report/diagnostics/idn-logcheck-core.log` SHA-256 `79cdc19acc4888e60bed6e238caa1f7580d8e55e6b109a68984d38b91c328001`;
  `idn-logcheck-web.log`, `idn-scan.log`, `idn-scan-report.log`, all other `report/` files.
- **Not retained:** the raw service logs the check scanned (`core.log`, `core-replica.log`, `web.log`, `proxy.log` in the identity run's
  volume). They are not in the snapshot, so the 56 offending lines themselves are unavailable.

### B-IDENTITY-PRIVACY (the only failing row) — what is known

- The check is `infra/validation/run.sh` L5344–5361: `logcheck validate --service core --file core.log --file core-replica.log`, then
  `logcheck scan --canaries` over the service logs and over the browser evidence.
- Canary scans **pass**: `scan: 275 canaries, 0 hit lines` on both the service logs and the browser evidence. **No secret or canary
  value is exposed.**
- The failing step is `logcheck validate` on core: `lines 16568, json_records 16512, unclassified 56, failures_total 56`, every
  failure class `completion-invalid` (in `infra/validation/tools/internal/logrecord/logrecord.go`: an `http.request.completed` record
  that fails `completion.schema.json`). The web log passed (257/257, 0 failures).
- Every route in the report's route/status table is inside the schema's route enum (checked by a coordinator script); the failing
  line numbers fall in repeating triples (gaps 9, 11, then a longer gap), so the 56 completion records appear to share a property
  (request-id or trace form, level-versus-status rule, parent_span_id or similar) — a guess, not a finding. **The failing field is not determinable
  from the retained evidence.**
- The same row also failed in the first full run (`2d55e37f`, run.log `bcc2ec07…`, cause then unavailable). The failure therefore
  survives the attempt-6 fixture corrections and appears unrelated to them. Whether it predates G060 or comes from G060 code (for
  example the Option B / membership-invitation paths or the new logged requests) is **unknown**.
- Codex's earlier disposition holds: do not infer the cause without the actual output.

### State and next step

G060 is merged (MR !59 and the correction MR !60) and **not qualified**. The next step needs Codex's disposition: a bounded attempt-7
dispatch to reproduce `logcheck validate` on the identity stack with the raw `core.log` retained, identify the 56 records and the
violated schema rule, and correct in scope (or return to Codex if the fix is outside the 42-pattern allow-list or changes the
completion schema/oracle). Then a focused identity-privacy rerun and, if the source changes, a new merged-SHA full.

## Attempt 7 (B5 attempt 7 of cap 7) — B-IDENTITY-PRIVACY reproduced and isolated; the fix is outside scope — RETURNED TO CODEX (2026-10-20)

Writer: the attempt-7 Claude worker (Opus 5.5), the only write owner, in the isolated worktree
`.claude/worktrees/agent-<id>`. Branch `g060-attempt7` was created from `origin/g060-result-fullrun2` `59b08dae`
(= staging `f24f0eb9` plus the second-run section), and branch/HEAD were checked. **No source, test, schema, oracle, catalogue or
threshold change was made.** This is the final attempt under the cap; no further attempt is available without Codex/owner
reconciliation.

### Reproduction (focused, before any edit)

- Pre-batch: `origin/staging` = `f24f0eb9` (unchanged and an ancestor of HEAD). `bash infra/ci/docker-desktop-mounts.sh` →
  16484 of 100000 (16%). No platform containers were running and the coordinator was running nothing.
- Selector: `ops` is the smallest harness selector that contains `identity_privacy`. `suite_ops` = `variants_ops` + `variants_identity`,
  and the identity variant needs the ops variant's unconfigured-stack result. Command (background; shared lock taken; no native refusal):
  `flock -n <home>/.local/state/platform-evidence/local-harness.lock bash infra/validation/run.sh --base HEAD --allow-file infra/ci/unchanged.allow --keep --report-dir …/g060-s12/a5/a7/report ops`.
  `--keep` retains only the work tree; the harness still removed every resource it created.
- Run `8ebbdb` on source `59b08dae` (core/web/infra trees = `f24f0eb9`), 00:42:28–00:57:47 local, exit 1. **16 PASS, 2 FAIL,
  4 INFO.** The FAILs are B-IDENTITY-PRIVACY and the aggregate B-REQUIRED-CHECKS. PASS rows include B-IDENTITY-IT (88), B-IDENTITY-E2E
  (133 browser tests), B-IDENTITY-PROBES, B-MCP-E2E, B-IDENTITY-CONFIG, B-OBS-LOGS/PRIVACY, B-SECRET-CANARY (207) and B-SCOPE (0 changed
  paths).
- **The failure is identical to both merged-SHA fulls.** Core `logcheck validate`: lines 16559, json_records 16503, unclassified 56,
  failures_total 56. Every failure is `completion-invalid` in `/o/core.log` (none in the replica), at the same triple pattern (6277,
  6286, 6297, … 13576, 13585, 13596). The web log passes (258 lines, 0 failures). Both canary scans report `206 canaries, 0 hit
  lines`, so **no secret is exposed**. The harness logged no other privacy sub-condition.
- Raw logs were copied out before cleanup. The work tree, which held the run's secrets, was then removed.

### Offending records and violated rule

- **All 56 records have the same shape**: `http.request.completed` from core with
  `route: "/internal/v1/organization/group-authority"`. The split is GET 200 ×38, POST 200 ×16 and POST 403 ×2. Each record has
  `level: info`, `outcome: completed`, `trace_source: inbound` with `parent_span_id`, `request_id_source: inbound` and no
  `error_type`. All 56 have the same 16 keys: the 15 required ones plus `parent_span_id`.
- The raw logs hold exactly 56 records with that route (core 56, replica 0). So **every** Group-authority completion fails, and
  **every** failure is a Group-authority completion. The clusters (9, 9, 13+6, 13+6) are GET page → POST action → GET redirect
  sequences. They fit the `/organization/group-authority` form steps that E-ONB-8 and E-MEM-8 drive in their light and dark
  projects.
- **Rule isolated.** This used a task-built `validation-tools` runtime image with the same Dockerfile and source, network none:
  1. The 56 original records give exit 1 and 56/56 `completion-invalid`.
  2. The same 56 records, changing only the route value to the listed `/internal/v1/organization/slug`, give exit 0: 56
     completions, 0 failures, `pass: true`.
  3. The whole run's `core.log` + `core-replica.log`, with only those 56 route values changed, give exit 0: 16559/16559
     classified, 0 failures.

  No other schema rule is violated anywhere in the run: level/status, outcome/error_type, trace parent and the ID patterns all hold.
- **Root cause.** G060 registered `GET` and `POST /internal/v1/organization/group-authority` in core
  (`core/internal/membership/api/http.go`). `observability.logCompletion` labels each completion with the matched ServeMux path
  (`routeOf(r.Pattern)`). `infra/validation/tools/internal/logrecord/completion.schema.json` (SHA-256 `0db40ac6…2bfb57`) does not
  list that path in either route enum: the top-level `properties.route.enum` and the `service: core` branch of `allOf`. Earlier
  Goals that added core routes extended this schema's route labels in the same change; the file history shows G004, G014, G016,
  G017, G018, G019 and G058. G060's Plan and scope amendments added `events.json` and `logrecord_test.go` but never
  `completion.schema.json`.
- This also explains why the row failed in both merged-SHA fulls, with or without the attempt-6 fixture changes. Only the identity
  browser run sends traffic to these two routes, and no other harness log check sees them. **This is not a security or privacy
  exposure.** It is a catalogue gap in the completion-record oracle.

### Why no correction was made (scope stop)

The only correction consistent with the contract is an oracle change. It adds `"/internal/v1/organization/group-authority"`
to both route enum lists in `infra/validation/tools/internal/logrecord/completion.schema.json`, plus positive vectors in
`logrecord_test.go` (GET 200 and POST 403), as earlier Goals did. `completion.schema.json` matches **none** of the 42 allow-list
patterns (scope file SHA-256 `5e5c1172…9e2b56`). It is also the completion schema/oracle, which the dispatch reserves to Codex.
Renaming the core route or changing how it is labelled would change an approved interface and the Plan, so neither is in scope.
Per the dispatch rule, attempt 7 stopped before editing.

**Question for Codex:** may G060 add the exact path `^infra/validation/tools/internal/logrecord/completion[.]schema[.]json$`
(scope amendment 02)? It would admit the one route label `/internal/v1/organization/group-authority` in the two existing enum lists.
No other key, rule, pattern or threshold would change. The `logrecord_test.go` vectors are already in scope. If approved, the
proposed validation is:

- the `validation-tools` check target (`gofmt`/`vet`/`go test ./...`, which covers `logrecord_test.go`);
- `logcheck validate --service core` on this attempt's retained raw logs (expected 16559/16559, 0 failures);
- a focused `ops` harness rerun on the frozen committed SHA (B-IDENTITY-PRIVACY);
- after integration, a new merged-SHA local full.

The budget needs Codex/owner reconciliation, because attempt 7 was the last under the cap.

### Evidence (`<home>/.local/state/platform-evidence/g060-s12/a5/a7/`, SHA-256)

| File | SHA-256 |
|---|---|
| `harness-console.log` (run `8ebbdb`) | `4ad4499daf5defc90124dde9460d3e218a8808caa883ee8ba6e1593bbc0afd15` |
| `report/results.tsv` | `5c93f22ec884b976b8e2d3bd5c6a0d5d77d13b7bebeae15e99899360d6702d26` |
| `report/diagnostics/harness.log` | `37faea4b70b8bfe93fa072100347e9f7c0afc3a68e23ad1a728b927e2a4c7ccf` |
| `raw/core.log` (raw, retained) | `ecedb108b6a117ad74f31a81ec507726a4e4ff5e5df39e6227dc36c05a6002cc` |
| `raw/core-replica.log` | `f32468bd04145d018ce5955ab7b11408e512f196795bd50dd5de5a708902b73a` |
| `raw/web.log` / `raw/proxy.log` | `d10d4635…b5aa2e` / `eefc51b0…9982f` |
| `raw/idn-logcheck-core.log` | `3bc430ca521439e77ee0fa169057fe9e3b301f3518680cf55f655c671415637a` |
| `raw/idn-logcheck-web.log` | `621f0f69d1f9a748c1d9cbd4281c10ccf4c30353b2dbb924d493b3ee7fa294bf` |
| `raw/idn-scan.log` = `raw/idn-scan-report.log` | `4b3e5c0401debf256efde03920ac05908da8b62b8aebe1eef88b0845402aeb5c` |
| `offending-56.jsonl` (the 56 records, verbatim) | `8b7129c9862b2a278569ce558fb778bd30314b5e8a5ecae59791f33153971b12` |
| `iso-1-original-56.json` (exit 1) | `7023a13dbe5a5a80b87d3e3d684bde5d468d6cdf7839fa1c55bd48c391bfcf06` |
| `iso-2-relabelled-56.json` (exit 0) | `bf93a50bdeb4b4d547b9e3747cc8c74671f5fd9b9c7b67868409b16850331bdd` |
| `iso-3-relabelled-full.json` (exit 0) | `bb480d2a92113f5e437b50d6ab2cf27d8ec8d073d873daf4d296fc81fc87e575` |

`isolate/` holds the isolation inputs: route-relabelled synthetic copies, which are diagnostic only and are not evidence of
service behaviour.

### Cleanup receipt

- Harness run `8ebbdb` removed every resource it created: the leftover inventory is empty, and no container, image, volume or
  network carries `g001.run=8ebbdb`.
- The lock was released when the run exited.
- The retained work tree `/tmp/project-validation-7fE6Mh`, which held the run's secrets, was removed and its absence verified.
- The task image `g060-a5-a7-vtools:local` was removed. No `g060-a5=a7` container or image remains.
- The raw evidence above is kept. The monitor was stopped.

### State

G060 is merged and **not qualified**, and B-IDENTITY-PRIVACY still fails on staging `f24f0eb9`. Attempt 7 delivered diagnosis only.
It awaits Codex's scope/oracle disposition; there is no self-acceptance.

## Attempt 7 fix (continuation, scope amendment 02) — 2026-10-20

Status: **submitted for Codex review** (not accepted). This continues B5 attempt 7 of cap 7 under Codex's bounded scope
amendment. It is not attempt 8, and the budget is unchanged.

### Scope amendment 02 (recorded before any schema edit)

- **Codex disposition (binding).** The B-IDENTITY-PRIVACY failure is a validator allow-list omission. G060 added the core
  routes `GET` and `POST /internal/v1/organization/group-authority` (`core/internal/membership/api/http.go`), but
  `infra/validation/tools/internal/logrecord/completion.schema.json` lists neither in its two route enums (top-level and
  the core `allOf` branch). So all 56 core completion records on that route fail: GET 200 ×38, POST 200 ×16 and POST 403
  ×2. The canary scans are clean.
- **Approved only:**
  - (a) Append exactly `^infra/validation/tools/internal/logrecord/completion[.]schema[.]json$` to the allow-list and
    keep every prior line.
  - (b) Add exactly one enum value, `/internal/v1/organization/group-authority`, to each of the two route lists. GET and
    POST share the matched path; the method is a separate field. `additionalProperties` and every other route, method
    or record constraint stay unchanged. No canary or validation weakening. `events.json` stays untouched unless a
    focused check proves group-authority events missing; in that case, stop and ask.
  - (c) Focused `logrecord_test.go` vectors: GET 200, POST 200 and POST 403 accepted for core, and the route rejected for
    web and intelligence. They must be red on the pre-fix schema and green after the fix. `logrecord_test.go` was
    already in scope. No other new path is authorized.
- **Review history.** Attempt 7 (diagnosis only, section above) stopped before editing and asked Codex about scope
  amendment 02. Codex approved the bounded amendment above. This continuation writer applies it.
- **Allow-list** `management/v0/results/G060-scope-allowlist.txt`: before, 42 patterns, SHA-256
  `5e5c11720a1fa1a61d6a674bfd0721549b430ef6cd508f62d1f04e45e9fd2b56`; after, **43 patterns**, SHA-256
  `f5c8802f40a993dcdddc9cc083df64fbe10a47f80fcaa32bd2334663afe2d5f0`. The diff is the one appended line (+1/−0). All
  prior lines are byte-identical, including the header comment, which still reads "42 ERE patterns" because prior lines
  are preserved.

### Worker record (fix continuation)

- Single write owner: the attempt-7 fix writer, model Opus 5.5 (effort inherited; the exact runtime value is unknown).
  It works in the isolated worktree `.claude/worktrees/agent-<id>` on local branch `g060-attempt7-fix`,
  created from `origin/g060-attempt7` `44d9a083`. Branch and HEAD were asserted after the switch. The attempt-7
  worktree `agent-<id>` (locked) was left untouched.
- Inputs: this Result before the section, SHA-256 `77d8bd31f46dccea1acd9c91f592c93cd1166e2c3fbb406b6a9530ff0d585492`
  (matches the dispatch). Pre-fix `completion.schema.json`, SHA-256
  `0db40ac6af6d29bd03abf978adaaf3cac6c7c744d0c4de7f7ab45682b9f37c80`. A precision note on the attempt-7 section:
  its "`0db40ac6…2bfb57`" elision has the wrong suffix. `…2bfb57` is the suffix of `events.json`
  (`4050e9f0…2bfb57`); the earlier section is not rewritten.
- Pre-batch: `git fetch`. `origin/staging` was `f24f0eb9`, an ancestor of HEAD (`git merge-base --is-ancestor` exit 0).
  There was no target delta, so no rebase. The checkout was clean before each run.
  `bash infra/ci/docker-desktop-mounts.sh` showed 16560 of 100000 mounts (16%). No platform container or image build was
  active before the build.
- Evidence: `<home>/.local/state/platform-evidence/g060-s12/a5/a7f/`.

### Commits (red → green)

| Commit | Parent | Content |
|---|---|---|
| `1b33abfc` | `44d9a083` | Scope amendment 02: allow-list +1 line and the amendment note above |
| `02502dab` | `1b33abfc` | **Red.** `logrecord_test.go` +8 lines. `TestClassifyAccepts` gets three core completion vectors on `/internal/v1/organization/group-authority`: GET 200, POST 200 and POST 403. The `completion()` defaults apply: level info, outcome completed, inbound trace with `parent_span_id`, no `error_type`. `TestClassifyRejects` gets the route on web and on intelligence (`completion-invalid`), in the existing style for core-only routes |
| `66034158` | `02502dab` | **Green.** `completion.schema.json` +2/−2: the one label is appended to the top-level `properties.route.enum` and to the `service: core` `allOf` enum. Post-fix SHA-256 `fe2cf6ea4859c573168a7921867094db12a231ad8f8d7404544cd17d47e5cd82` |

The commit that records this section is management-only and does not change any tested tree. `logrecord_test.go` after
the change has SHA-256 `456464429f2a5d7d759c65a3eb023e3dc81e553abd320a8b1b7f4c3dec7dfa2b`.

### Validation (unit/static; isolated, no lock)

Method: one image build in the sequential build slot (no other build active), with the harness's B-CONTRACT-TOOLS check
stage and a context of `02502dab`. Build log: `build-check.log` `dec4b6a8…b480`.

```
docker build --label g060-a5=a7f -f infra/docker/validation-tools.Dockerfile --target check -t g060-a5-a7f-vtools-check:local infra/validation/tools
```

Each run:

```
docker run --rm --label g060-a5=a7f --network none -v <worktree>/infra/validation/tools:/src:ro -e GOFLAGS="-trimpath -mod=readonly" -e GOPROXY=off -e GOCACHE=/tmp/gocache g060-a5-a7f-vtools-check:local <cmd>
```

The checkout was clean at the named commit for each run. The read-only mount replaces the image's baked source, so each
run tests exactly that committed tree. The module cache is the image's own, from the Dockerfile's
`go mod download && go mod verify`. GOCACHE is disposable, there is no published port and no shared cache was written.
The logcheck runs also mount attempt 7's retained raw logs read-only at `/o`: `core.log` `ecedb108…02cc` and
`core-replica.log` `f32468bd…735a`, the same files as in attempt 7's table.

| Run | Source | `<cmd>` | Exit | Outcome | Evidence (SHA-256) |
|---|---|---|---|---|---|
| red-focused | `02502dab` (pre-fix schema) | `go test -count=1 -v -run '^(TestClassifyAccepts\|TestClassifyRejects)$' ./internal/logrecord/` | 1 | **FAIL as intended.** `TestClassifyAccepts` FAIL with exactly 3 errors: the three new group-authority vectors, each `class "completion-invalid"`. `TestClassifyRejects` PASS, because the new web/intelligence rejections also hold before the fix | `red-focused.log` `e7f103c22353023cc1d393c9ae090d289ffa1b6ca434f05375848b1fc7fa2146` |
| red-logcheck | `02502dab` | `go run ./cmd/logcheck validate --service core --file /o/core.log --file /o/core-replica.log` | 1 | Lines 16559, json_records 16503, unclassified 56, failures_total 56, `pass:false` (identical to attempt 7) | `red-logcheck-core.json` `b6303cb3e8c3092f7e59437cff962f9900931e71cba820b2e78afdf4e1c6f7c7` |
| green-focused | `66034158` | Same as red-focused | 0 | **PASS 2/2** | `green-focused.log` `d07bcb11dd47251dca3900483afdcc609f2b3d80e4a86c741e425ffbea43d2df` |
| green-pkg | `66034158` | `go test -count=1 -v ./internal/logrecord/` | 0 | **PASS 10/10** tests, 0 FAIL, 0 SKIP | `green-logrecord-pkg.log` `ddfb53669807435e24ad85578b7107ff26d749aec68a18e0e2edfb93a826b57c` |
| green-check | `66034158` | The image's default check CMD (`gofmt -l .`, `go vet ./...`, `go test ./...`) | 0 | `gofmt` printed nothing, vet clean, **7/7 packages ok** | `green-check.log` `364a4298707c14569ad9781724cb031d6e6925345afe56ec82b349f953b2d407` |
| green-logcheck core | `66034158` | Same as red-logcheck | 0 | **Lines 16559, json_records 16559, unclassified 0, failures_total 0, `pass:true`.** Completions 7903 (= 7847 + 56), including GET `…/group-authority` 200 ×38, POST 200 ×16 and POST 403 ×2 | `green-logcheck-core.json` `070cd271d0aa8473874604350ef29330b464568f48295e5c9e8cb1eff22c53fc` |
| green-logcheck web | `66034158` | `go run ./cmd/logcheck validate --service web --file /o/web.log` | 0 | 258 lines (253 JSON, 5 allowed non-JSON), 0 failures, `pass:true`. Byte-identical to the harness's `idn-logcheck-web.log` | `green-logcheck-web.json` `621f0f69d1f9a748c1d9cbd4281c10ccf4c30353b2dbb924d493b3ee7fa294bf` |

- **`events.json` was not changed** (SHA-256 `4050e9f0276741060523343584655fbbbd0ec083e12e98af00b3c7322c2bfb57`).
  The green core logcheck classifies every line of the run, so no group-authority event is missing. No stop was needed.
- **Not run, by dispatch:** the `ops` harness and the full suite. The coordinator owns the focused B-IDENTITY-PRIVACY
  confirmation and the new merged-SHA local full.
- **Canary scans not rerun:** `logcheck scan` does not read the completion schema, so attempt 7's scans (206 canaries, 0
  hit lines) stand.

### Scope check

The branch delta versus `origin/staging` `f24f0eb9` is 4 paths (`a7f/scope-files.txt` `2f5a0d6f…c407`):
`completion.schema.json`, `logrecord_test.go`, this Result and the allow-list. Paths outside the 43-pattern allow-list:
**0** (`grep -Evc -f`). The schema diff adds only the one label to each list; no other key, rule, enum, method,
`additionalProperties` or `$defs` changed.

### Security review

- The core route label set grows by exactly the one path core serves. It is `RouteGetGroupAuthority` /
  `RouteSetGroupAuthority` in `core/internal/membership/api/http.go`, and `run.sh` `CORE_MEMBERSHIP_ROUTES` already lists
  both pairs.
- The enum is exact-match, so no prefix, wildcard, query or trailing slash is admitted.
- The web and intelligence branches are unchanged, and their rejection of this route is now pinned by tests.
- The closed key set (`additionalProperties: false`) is unchanged, as are the level/status, outcome/`error_type`,
  trace-parent and ID rules. Canary and secret scanning are untouched.
- No secret was read or printed.

### Native refusals and prompts

None. The source edits used the Edit tool. The build, the runs, the commits and the image removal ran without a
refusal.

### Cleanup receipt (fix continuation)

- All seven test containers ran with `--rm`; 0 containers labelled `g060-a5=a7f` remain.
- The task image `g060-a5-a7f-vtools-check:local` (`fd422f6a…`) was removed, and 0 images labelled `g060-a5=a7f` remain.
- No volume or network was created (both verified at 0). The BuildKit cache was not pruned, since a blanket prune is not
  allowed.
- The build slot was released when the build completed. No lock was taken.
- Kept: the `a7f/` evidence (commit messages, build log, run logs and `scope-files.txt`) and all attempt-7 evidence in
  `a7/`, including the raw logs and the failed-run outputs, unchanged.

### State

G060 is merged and **not qualified**. The source fix is on `g060-attempt7-fix`. The next steps are the coordinator's MR,
Codex review and the new merged-SHA local full (B-IDENTITY-PRIVACY). Budget: B5 attempt 7 of cap 7, of which this is the
continuation. There is no attempt 8 and no self-acceptance.

## Merged-SHA local full qualification — third run on `e0a52c6a` (corrected source) — 2026-10-20

Recorded by the coordinator. Result: **B-IDENTITY-PRIVACY now PASSES**; the run's only failing row is **B-CI-CONTROLS (2 of 521
controls: VS3 and G3)**, which did not reproduce on a targeted rerun of exactly those two controls on the same source. This is
reported as run, not as a clean full PASS: the whole-run verdict stays `final status 1`. G060 is unqualified until Codex decides
whether the targeted rerun suffices or directs a further full run. No code, test, check or threshold change since MR !61.
Budget: B5 attempt 7 consumed (the schema fix continued attempt 7, no source was edited before the scope amendment); attempt 8
and any extension amount are **not** started or assumed — the owner's "another budget" amount is pending Codex's record.

### Run identity

- Source: staging merge `e0a52c6a50c42f9e006280d81747d0eb4bc4280c` (MR !61 = `g060-attempt7-fix` `96b2f65b`; first parent staging
  `f24f0eb9`). Coordinator checkout `.claude/worktrees/0-agent-5c6f436`, detached and clean at launch and after the run (`git
  status --porcelain` empty); `bash infra/ci/docker-desktop-mounts.sh` first (16%). The `--evidence-dir` was NOT pre-created.
- Command (background, shared harness lock):
  `flock -n <home>/.local/state/platform-evidence/local-harness.lock bash infra/ci/validate.sh --profile full --evidence-dir <home>/.local/state/platform-evidence/g060-s12/postmerge-full-e0a52c6a/evidence > …/run.log 2>&1`
- Timing: started about 05:33Z (harness pre-observation 05:32:52Z); finished about 06:41Z (about 68 minutes). `validate: final
  status 1; published=none; registry=removed`; phases harness=1, complete=14; cleanup verified (`cleanup_status=0 owned_remaining=0`).
- Tally: 119 rows PASS, 2 FAIL (B-CI-CONTROLS and the aggregate B-REQUIRED-CHECKS), INFO BACKGROUND/BUILD-LEDGER/IMAGES.
- Previously failing rows now PASS: B-POLICY-IT, B-RLS-IT, B-OBJ-ORDERING, B-OBJ-FAULTS, B-OBJ-IT and **B-IDENTITY-PRIVACY**;
  also B-IDENTITY-IT, B-IDENTITY-E2E, B-IDENTITY-PROBES, B-MCP-E2E, B-IDENTITY-CONFIG, B-SECRET-CANARY (296 generated secrets
  absent) and B-SCOPE (0 changed paths since `e0a52c6a`, all within the allow-list).

### Failure history preserved

| Run | Source | Result | run.log SHA-256 |
|---|---|---|---|
| first | `2d55e37f` | 114 PASS, 6 FAIL: POLICY-IT, RLS-IT, OBJ-ORDERING (1800 s timeout), OBJ-FAULTS, IDENTITY-PRIVACY, aggregate | `bcc2ec07bfddb6ab0f5f511066af9b5c6d055c3d8273f241b2485a70be9fee0f` |
| second | `f24f0eb9` | 118 PASS, 2 FAIL: IDENTITY-PRIVACY (56 completion-invalid on the new group-authority route), aggregate | `71a44b667e294f99cc0fe28c91fb89bd81a4fd071e3efeccfe125312d5c6f984` |
| third | `e0a52c6a` | 119 PASS, 2 FAIL: B-CI-CONTROLS (VS3, G3), aggregate; IDENTITY-PRIVACY PASS | `f0e2e64749abfe55646404bf1a0c66a70f8aa3fba9d3fa6ae643e07e6e122222` |

### B-CI-CONTROLS (the only non-aggregate failure) — what is known

- Harness line (verbatim): `FAIL B-CI-CONTROLS - CI controls (exit 1, 519 passing)`; `ci-controls: 519 passed, 2 failed (of 521
  selected; 521 defined)`; the suite ran concurrently with the rest of the harness (BACKGROUND, joined after 3703 s).
- The two failing controls (from the retained `ci-controls.log`):
  - `FAIL VS3 an available observation is not ok;a size is not a Docker CLI size;` — the control interrupts a stub harness with
    TERM and checks the storage observation file of that scratch run. The real run's own `evidence/storage.tsv` is valid
    (every pre and post row `ok`, sizes in Docker CLI form such as `48.88GB`), so the real storage observation is not the defect.
  - `FAIL G3 adapter exit 20, want 0 [hosted-context: final: context=0 guard=FAIL (guard stalled)]` — the hosted-context
    adapter's disk-guard sampler reported a stall.
- **Neither control touches any path G060 changed** (the two controls exercise `infra/ci` storage/guard behaviour; the corrected
  source changes only the log-record schema, its test, the allow-list and the Result). Both passed in the first and second full
  runs on the same machine. A timing/load sensitivity under the concurrent run is plausible, but the cause is **not established**.
- **Targeted rerun on the same exact source `e0a52c6a`** (clean checkout; under the shared lock; mounts 17%):
  `flock -n …/local-harness.lock bash infra/ci/validate-controls.sh --only VS3,G3 --log …/postmerge-full-e0a52c6a/ci-controls-rerun-vs3-g3.log`
  → `PASS VS3`, `PASS G3`, `ci-controls: 2 passed, 0 failed (of 2 selected; 521 defined)`. Rerun log SHA-256
  `c5e786dd…7630d`.

### Retained evidence

- `…/postmerge-full-e0a52c6a/run.log` SHA-256 `f0e2e64749abfe55646404bf1a0c66a70f8aa3fba9d3fa6ae643e07e6e122222`;
  `…/ci-controls-rerun-vs3-g3.log` SHA-256 `c5e786dddc1a12878e06ad78f69d66a5f7718a4c4d84b466e10bb2d1d327630d`;
  `…/live-copy/report/ci-controls.log` SHA-256 `5a8768074c9abd0dfa5b01932de1e7eff4916288aebee5f91f835b2a5faef81f`;
  `…/live-copy/harness-console.log` SHA-256 `4526fcc6e78c7804aad2723ab7224e62807c37aa1c84da9898d2ec00c17f8586`;
  `evidence/storage.tsv` (retained by the local run); the private `live-copy/` snapshot of `evidence/raw` files.
- Not retained: the ci-controls scratch directories of the two failed controls (the case workspaces are removed by the harness),
  so the exact stall detail of G3 and the failing observation of VS3 are **unavailable**.

### State and next step

G060 is merged and **not yet qualified**: the required exact merged-SHA full run did not end `final status 0`. Needs Codex's
disposition: accept the targeted VS3/G3 rerun on the same source as covering the two unreproduced controls (with a durable
statement that B-CI-CONTROLS passed 519 of 521 in the full run and 2 of 2 on rerun), or direct another full run. If a further
implementation fix were required it would need the owner's budget amount; none is indicated by this evidence.

## Qualification disposition and closeout receipt (2026-10-20)

### Codex qualification disposition (relayed to the coordinator; recorded here, not self-issued)

Codex: no fourth full-suite run is required for the two residual controls. Under the targeted-rerun / attributable-coverage policy,
the third full run on exact merged SHA `e0a52c6a` plus the focused VS3 and G3 reruns on that same SHA provide complete attributable
coverage. Codex independently inspected `run.log`, `attribution.json`, `ci-controls-rerun-vs3-g3.log` and `cleanup.log`.

**The third full run stays recorded as FAILED** (`validate: final status 1`; the aggregate B-REQUIRED-CHECKS FAIL is derived from
B-CI-CONTROLS). It is not relabelled and no full-pass report is synthesized.

| Item | Detail |
|---|---|
| Failed rows in the run | B-CI-CONTROLS (controls **VS3** and **G3**; 519 of 521 passed) and the derived aggregate B-REQUIRED-CHECKS |
| VS3 original outcome | `FAIL VS3 an available observation is not ok;a size is not a Docker CLI size;` |
| G3 original outcome | `FAIL G3 adapter exit 20, want 0 [hosted-context: final: context=0 guard=FAIL (guard stalled)];` |
| Rerun source | exact `e0a52c6a50c42f9e006280d81747d0eb4bc4280c`, clean detached checkout, shared lock held |
| Rerun command | `flock -n …/local-harness.lock bash infra/ci/validate-controls.sh --only VS3,G3 --log …/postmerge-full-e0a52c6a/ci-controls-rerun-vs3-g3.log` |
| Rerun outcome | exit 0; `PASS VS3`, `PASS G3`, `ci-controls: 2 passed, 0 failed (of 2 selected; 521 defined)` |
| Cleanup | full run `cleanup_status=0 owned_remaining=0`; rerun: harness cleanup verified |
| Cause | **unknown** (the failed controls' scratch case directories are removed by the harness); no control, threshold or check was changed |

Complete attributable coverage: 119 rows PASS on the full run (B-IDENTITY-PRIVACY among them) plus 2 of 2 on the same-SHA targeted
rerun of exactly the failed controls. B-IDENTITY-PRIVACY passes on this SHA.

### Attempt and budget accounting (as recorded in this Result)

- B5 attempts 1–5 as recorded in their sections; attempt 6 (diagnosis, correction round, residual finding) and attempt 7 (diagnosis
  with scope stop; fix continued inside attempt 7 after scope amendment 02, no source edited before the amendment). Cap 7 per the
  recorded extension; **no attempt 8 was started**; the owner's later "another budget" amount is unrecorded and was not used.
- Scope: 42-pattern allow-list after amendment 01, **43** after amendment 02 (SHA-256 `f5c8802f…d5f0`); every changed path was inside.

### Closeout receipt (mandatory cleanup; verified by the coordinator)

- Docker: no `g060-a5*` image, container, network or volume remains (checked by label and tag); every run reported its own cleanup.
- Worktrees removed (clean, HEAD verified, tool locks released): the attempt-6, correction, residual, attempt-7 diagnosis and attempt-7
  fix writers, and the G058 `group-admin-404` worktree. Local branches deleted (content superseded by staging or patch-equivalent):
  `g060-impl`, `g060-plan`, `g060-plan-r2`, `g060-plan-amend01`, `g060-result-fullrun`, `g060-result-fullrun2`, `g060-attempt6`,
  `g060-attempt6-mr`, `g060-attempt7`, `g060-attempt7-fix`, `g060-a6-correction`, `g060-a6-residual`, and eight `worktree-agent-*`
  placeholders of G060 writers. Remote branches deleted: `g060-attempt7`, `g060-plan`, `g060-plan-r2`, `g060-plan-amend01`,
  `g060-result-fullrun`, `g060-result-fullrun2` (`g060-attempt6` and `g060-attempt7-fix` were removed by their merges).
- Evidence: scratch draft patches, empty secrets directories removed; **retained** under `platform-evidence/g060-s12/` (about 70 MB:
  the three full-run logs and snapshots, per-attempt evidence and the receipts cited above by hash) pending Codex's import of this
  Result.

### Pending closeout items (not done by the coordinator)

1. The six G060 bridge permission rules (`g060-a5` Write/Edit evidence path, docker build/run/compose/image-rm) remain in the main
   checkout's ignored `.claude/settings.local.json`. The coordinator does not edit its own permission settings; request: Codex or
   the owner removes exactly those six rules after final ACCEPT (the permanent generic test rules stay).
2. Remote and local `g060-result-fullrun3` (carries this Result update): delete once Codex has imported the Result.
3. `g058-group-admin-404` (local and remote, `5f94a705`): not patch-equivalent in staging; kept pending Codex's disposition of that
   commit (worktree already removed).
4. `platform-evidence/g060-s12/` purge after the Result is imported; `g058-retained` and the r16 toolkit per reuse (Codex).
5. Older `worktree-agent-*` placeholder branches of earlier attempts (identifiers not recorded) were not enumerated or removed.

### Status

G060 final handback submitted to Codex for ACCEPT. The coordinator has not accepted anything. RC3 and the G055 resume wait for
Codex's recorded final ACCEPT.
