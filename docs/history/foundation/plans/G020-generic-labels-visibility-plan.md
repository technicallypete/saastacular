> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G020 — Generic labels and visibility — Implementation Plan (r2, PROVISIONAL)

Goal: [G020](../goals/G020-generic-labels-visibility.md) (SHA-256 at `origin/staging` `c28671ed`
`f61dbc39f1478db7fa555f565238cefb6067ceec2bef0731a4adf8ab51b297b4`; r0 bound `4a02248d…18b9eb` at `c1e724c6`) ·
Spec: [G020 Spec](../specs/G020-generic-labels-visibility-spec.md) (SHA-256 at `c28671ed`
`798017e9141f0fb064adf0edd2034728b763a9a084cf6bfb8b6528bc4259fbc1`; r0 bound `03da9a57…8df6` at `c1e724c6`) ·
Review: G020-plan-codex-review.md (not included) R0 CORRECTION LOOP (SHA-256
`e0b6ff31b7da105a191a45d54e1bf9c03d09b676e797b0ea21fb295f6bb00309`); R1 bounded correction remains (SHA-256
`de984df8a0e9b2ee54858cfc18dd281f0405868673f98b7611a37d07e3b9b5ed` at `origin/staging` `3b4b1dc2`) ·
Architecture §53.8 · [ADR-008](../../../adr/ADR-008-explicit-group-authority.md) · [ADR-006](../../../adr/ADR-006-audit-schema-boundary.md).
Status: **provisional Plan r2 (one bounded correction pass on the two R1 items: the fixture target contract and the
T3 owner-fact oracle) submitted for Codex review; planning only; zero execution attempts spent; implementation HELD.**
The r0/r1 text below is kept and amended in place where a finding required it (each amendment is marked "r1" or
"r2"); §14 maps every R0 finding/decision (r1 table) and every R1 item (r2 table) to its change. R1 settled findings
1, 2, 3 and 5, decisions D-1 and D-4, and accepted the production target set closed to `group`.
Authority: owner "Ya i agree" recorded by Codex on `origin/staging` `c1e724c6` (Goal/Spec section "Owner authorizes provisional planning now — 2026-10-20");
r1 under the R0 review's "one existing planner may make a bounded r1/correction addendum" and the coordinator's r1 dispatch.
Planner: task G020, role Plan writer. r0: Fable 5.1 high, agent `a99201e21a22d60bf`, worktree
`<source-checkout>` (stopped by the coordinator
after its context passed the 200k handoff threshold). r1: successor planner, Fable 5.1, requested effort high (runtime
effort **unknown** from inside the agent), isolated worktree `…/.claude/worktrees/0-agent-5c6f436` (§12 r1 quotes the
native refusal to enter the r0 worktree), local branch `g020-plan-r1` created at `g020-plan` `40792883`
(= `origin/g020-plan`) and pushed to `origin/g020-plan`. r2: successor planner, Fable 5.1, requested effort high (runtime
effort **unknown** from inside the agent), isolated worktree `…/.claude/worktrees/agent-<id>`, local branch
`g020-plan-r2` created at `origin/g020-plan` `c8ab9e4a` (verified; tree clean) and pushed fast-forward to `origin/g020-plan`.
Base: `g020-plan` from `origin/staging` `c1e724c642bcf419df6c431911df07779dd4ada6`; r1 reconciled against `origin/staging`
`c28671edb254fe9670775e8045a7826b36ef46d1` (three documentation-only commits after `d0b06691`: context limits, self-compaction,
the R0 review; no change under `core/`, `web/`, `infra/`, `contracts/` or `docs/foundation/`); r2 reconciled against
`origin/staging` `3b4b1dc20aa608091b284e818652733c8e74cf38` (two documentation-only commits after `c28671ed`: Codex delegate
defaults, the R1 review; `git diff --stat c28671ed 3b4b1dc2` over `core/`, `web/`, `infra/`, `contracts/`, `docs/foundation/`
and the G020 Goal/Spec is empty, so the r1 bindings stand).
Inputs: G020-planning-inputs.json (not included). File:line references below are at `c1e724c6` unless stated.

## 0. Provisional status, gates and bindings

- **Baseline.** Integrated G060 source on staging (merge `2d55e37f`; G060 final ACCEPT **pending** its exact-SHA local full
  qualification, cap 7 / 5 spent) and the current approved G055 design (Plan r3; G055 final ACCEPT and live cutover
  **pending**). Nothing in this Plan infers acceptance from those merges or from this planning permission.
- **Implementation gates (unchanged):** G060 and G055 final ACCEPT, Codex Plan ACCEPT, imported scope allow-list (§11),
  current input bindings, exact-source reconciliation (§13) and recorded execution release. Architecture/security/data
  tier: every D-item in §10 is flagged for explicit review. No migration number is reserved by this Plan (§3 describes
  the next free slot without claiming it).
- **Budget.** Provisional B5, checkpoint after attempt 3; planning spent none. Forecast 3 attempts (§11).
- **Context policy (r1 reconciliation of the Goal/Spec amendments of 2026-10-20 at `30a6d497`/`d1fe56f0`).** Worker and
  planner checkpoint at **150k** active context and safe handoff at **200k** (supersedes the 300k/350k figures in §6 r0).
  Claude/Codex self-compaction at safe points is authorized through supported native controls: save state, preserve
  active batches/children/budgets, verify actual completion; if the live interface exposes no agent-accessible control,
  report that once. No restart, added review round or validation campaign for the policy carry-forward.
- **Qualification rules.** Current canonical: focused local proof + feature-MR fast + Codex + Duo → reviewed staging
  merge → coordinator-owned exact-merged-SHA **local** full qualification → cleanup → final ACCEPT. Private per-run
  DB/browser stacks (repository pool of two, third only with owner approval), sequential image builds, isolated
  lock-free unit checks. Historical AC8 "hosted staging" wording is superseded by these rules.
- **Later staging delta noted (no rebase needed):** `d0b06691` (2 commits after the base) moves the canonical Vestive
  skill to `.agents/skills/vestive-design/` (`.claude/skills/vestive-design` becomes the alias); `SKILL.md` content is
  byte-identical (`7dfac5fa…c670e`). Only harness/CI path pins and doc references changed (`infra/validation/run.sh`
  +17/−9 around L847–L1545, `infra/ci/profile-paths.tsv`, `infra/ci/validate-controls.sh`, `web/eslint.config.mjs`,
  docs). This Plan cites the canonical path; `run.sh` line references below shift by up to +10 after L847 at `d0b06691`
  and are re-pinned at release.
- **Out of scope** (Goal): product objects, labels on users/profiles/objects, public tags, role grants via labels, bulk
  import, hierarchy, automation, search, colours/taxonomy designer, label-based access control, new vendors, live
  deployment, any fake production target type, G054 revival, G021 waiver, G061, roadmap changes.

## 1. Approach

Two forced tables in `app` (`label_definitions`, `label_assignments`) with one additive migration, three new closed
capabilities in the existing evaluator (`decide`), a closed typed target registry in core with exactly one production
target (`group`, Organization-owned definitions only) and a test-only Group-owned fixture adapter, guarded facade
operations under the existing lock order, capture-trigger audit, one thin internal HTTP adapter for the web BFF, MCP
parity through capability eligibility (no production tool, D-1), and minimal pages in the existing shell. No new role,
store, reader role, DEFINER body, dependency or component library.

Authority model (the finite set; `orgAdmin` = org_owner/org_admin, `gr(g)` = the caller's own role in Group g,
`OVR` = G060 override):

| Capability (kind) | org_owner | org_admin | org_member | group_owner(g) | group_admin(g) | group_member(g) | OVR effect |
|---|---|---|---|---|---|---|---|
| `label.organization.manage` (organization) | Y | Y | — | — | — | — | none |
| `label.group.manage` (group g) | — | — | — | own | own | — | **none** (never consulted) |
| `label.group.read` (group g) | — | — | — | own | own | own | **none** |

- Organization-owned definitions and their assignments on Groups are read and written only through
  `label.organization.manage`; a Group role (owner/admin/member) never reads them (Spec §2, owner clarification
  2026-10-18). `Target members` visibility is refused for this owner scope (CHECK + core).
- Group-owned definitions and assignments are Group **content** (G060 Spec R1 "protected object labels"): an
  explicit role in the exact owning Group is required; `orgAdmin`, `orgMgr`/OVR, Group selection and shared profiles
  confer nothing (Goal: "G060's ordinary management override does NOT grant object-label access"). Managers
  (owner/admin) read and write; members read only definitions whose visibility is `target_members` (and, for
  assignments, only with current read authority on the registered target — proven with the test adapter).
- A subject who independently holds both roles (fixture Grace: A admin + research admin) reads Organization-owned
  labels through the Organization path and research's Group-owned labels through the Group path; the paths never
  combine into a third one.
- Target existence/authority is a second decision in the same transaction: Organization-owned assign/unassign/read
  additionally requires `group.read` on the exact target Group (existence, same Organization, selected-Group context);
  Group-owned assignments go through the closed registry adapter (§4.1) that derives the target's owning Group from
  authoritative storage in the same transaction.

Riskiest slice first (§7 S1): the two tables, their RLS and the evaluator rows are proven by unit oracles, runtime RLS
and attack suites before any facade write, HTTP or web work.

### 1.1 What exists today (audited read-only; file:line at `c1e724c6`)

| # | Surface | Fact | Use in G020 |
|---|---|---|---|
| E1 | `access/policy.go:12-41` 19 capabilities; `capabilityKind` :112-123; `decide` :214-311; `orgAdmin` :252, `orgMgr` :256; `eligible` :374-434 | closed evaluator; count pinned at `matrix_test.go:878` (19), `policy_test.go:27`, `events.json:448-455` | +3 capabilities (22); new rows in the same switch; `orgMgr` not consulted |
| E2 | `access/facade.go:19-85` `guard`/`guardedTx`; `groupQuery` :115 (target Group FOR NO KEY UPDATE on writes); `policyWriteQuery` :140; `denial` :93-102 | one capability per operation; not_found vs forbidden by the same-kind read capability | reused verbatim; label operations are `guard` calls with `groupQuery`/`organizationQuery`; `readCapability` gains the two label kinds |
| E3 | `access/invitations.go:317-367` `invitationAdmin`/`guardEvaluated` | capability known only after the row is read | `label.definition.update/delete` and `label.unassign` by id: read the definition row first, then decide `label.organization.manage` or `label.group.manage` |
| E4 | `access/authority.go:77-97` `overrideUsed` | mandatory same-transaction audit append; failure refuses | pattern only; no override path exists for labels, so no `access.override_used` is ever appended |
| E5 | `access/objects_policy.go:57-101` `decideObject` | Group content needs explicit role; `orgAdmin` only for Organization objects | the `label.group.*` rows copy this rule |
| E6 | `tenancy/facts.go:47-95`, `store.go:84-123`; lock order AUTHORIZATION.md L410-428 | users → session → organization (FOR SHARE) → own bindings → Group rows (target FOR NO KEY UPDATE) → associations → target read → write | label rows are locked **after** step 7 (§1.3); no new Group-level lock |
| E7 | `00017:140-150` macros `U/O/G/R/MGR/GMGR/OVR/DEP`; `stored_objects_*` :263-277 (explicit Group role shape); `invitations_*` :281-294 | policy conventions | label policies use `MGR` for the Organization scope and `GMGR`/`rls_group_role` for the Group scope; never `OVR`, never `DEP` |
| E8 | `00015:134-185` `app.invitations` (composite `(organization_id, group_id)` FK MATCH SIMPLE, scope CHECK, FORCE) | table template | both label tables copy the shape |
| E9 | `00008:256-394` `events_attribute`; `00017:612-667` catalog CHECKs; `:694-703` trigger list; `:724-735` direct-append scope; `:835-836` denied resource list | audit shape rules | +5 trigger actions, +2 reasons, +7 operations, +1 resource type (§3.6) |
| E10 | `audit/audit.go:85-98` 39 trigger / 7 core actions; `:333-350` 11 resource types; `audit_test.go:98,118` counts 46/11/70/65 | parity | re-pinned at the new migration |
| E11 | `membership/api/http.go:37-66` route constants and `Paths`; `:245-306` `begin` (origin, session binding, CSRF); `:310-330` failure mapping; `:724-758` strict `mutation`; `TestMembershipCopyIsExact` :421 | the thin-adapter pattern | new `core/internal/labels/api` copies it byte-for-byte for the shared helpers (pinned by its own CopyIsExact test) |
| E12 | `mcp/catalog.go:76-131` production = 5 control tools, capability-mapped tools only from the test fixture; `server.go:63-95` hidden = unknown | parity path | three fixture tools mapped to the label capabilities in tests only (D-1) |
| E13 | `onboarding/api/http.go:373-386` Group DTO (8 keys); web `GROUP_KEYS` `onboarding-client.ts:198` | DTO pins | unchanged (labels are read by their own route, D-9) |
| E14 | `web/app/groups/[id]/page.tsx` (Members link gated :96), `organization/members/page.tsx` `staffLinks` :36-40, `organization/group-authority/{page.tsx,action/route.ts}`, `components/membership/group-authority-form.tsx`, `components/onboarding/state-badge.tsx` | page/form/badge patterns | §1.5 |
| E15 | `infra/validation/run.sh` route pins :178-196, importer allow-lists :1231-1247, migration set :1271, version/relation pins :2413-2428, `suite_migrations` :2464-2505, B-TENANCY-HTTP probes :2237-2306; `required-checks.tsv` (127 rows + header); `lib.sh:643-648` browser matrix; `infra/ci/profile-paths.tsv:25,30-33` | harness pins | §5.4, §9 |
| E16 | `logrecord/events.json` (`membership.request` :317; operations :429-446; capabilities :448-455; reasons :458-467), `completion.schema.json` core route enums | log-vector pins | new `labels.request` event, operations, capabilities, routes |
| E17 | Fixtures `infra/identity/mcp-seed.sql:43-102`: carol (A owner), frank (A admin), alice (A member), bob (research owner), dave (research member; Group-only in C), Grace (A admin + research admin, no login), Dana (design owner, no login), Organization B/C foreign | two tenants, two Groups per tenant suffice | §5 |
| E18 | Name validation `tenancy/names.go:15` `ValidateName` (1–200 code points, trimmed, no control characters); `MaxNameLength` :9 | reuse | label names reuse the validator with a tighter 80-point bound (D-5) |
| E19 | Identifier collision: branding `NormalizeLabelTx`/`sqlNormalizeLabel` (`tenancy/branding.go:178`, pinned `boundary_test.go:50`) and `display_label`; enrollment web `label` key (`enrollment-client.ts:322`) | naming | Go identifiers use `LabelDefinition*`/`LabelAssignment*`; SQL uses `label_definitions`/`label_assignments`; no bare `Label` type |

### 1.2 Model (Spec §1)

**LabelDefinition** `app.label_definitions`: `id uuid` (core-generated, nonzero), `organization_id`, `owner_scope_type`
(`organization`|`group`), `owner_scope_id`, `group_id` (NULL for the Organization scope; else the owning Group),
`name text`, `name_key text GENERATED ALWAYS AS (lower(name COLLATE "C")) STORED`, `visibility`
(`owner_scope`|`target_members`), `created_at`.

- `label_definitions_scope_consistent`: `(owner_scope_type='organization' AND owner_scope_id=organization_id AND
  group_id IS NULL) OR (owner_scope_type='group' AND group_id IS NOT NULL AND owner_scope_id=group_id)`.
- `label_definitions_visibility_valid`: `visibility='owner_scope' OR owner_scope_type='group'` (Organization-owned
  definitions can never opt into target-member visibility; Spec §2).
- `label_definitions_name_valid`: 1–80 code points after trimming, no C0/C1/DEL control characters, no leading or
  trailing space (the `ValidateName` rule with the tighter bound; D-5).
- Scoped uniqueness `label_definitions_scope_name_key UNIQUE (organization_id, owner_scope_type, owner_scope_id,
  name_key)`: ASCII-case-insensitive, otherwise byte-exact (the slug convention "validated, never normalized";
  TENANCY.md L128). Collision → `conflict` (`Field` `name_conflict`), zero writes.
- Bound `label_definitions_bound` (the `invitations_bound` pattern, 00015): at most **200** definitions per owner scope,
  counted after locking the scope row (Organization or Group) `FOR NO KEY UPDATE`; refusal `conflict` (`scope_full`).
- Keys immutable (`label_definitions_keys_immutable`, BEFORE UPDATE, SQLSTATE 23000): id, organization_id,
  owner_scope_type, owner_scope_id, group_id, created_at. Owner scope is immutable (Spec §1).
- **r1 (finding 1) — owner facts are non-nullable in both scopes.** `organization_id` and `owner_scope_id` are `NOT NULL`
  (`owner_scope_id` = the Organization id or the owning Group id by `label_definitions_scope_consistent`), so
  `label_definitions_organization_fkey (organization_id) → organizations (id)` and
  `label_definitions_group_fkey (organization_id, group_id) → groups (organization_id, id)` (`group_id NOT NULL` whenever
  `owner_scope_type = 'group'` by the CHECK) enforce definition existence, exact Organization and exact owner facts for
  every row; no definition row can skip its owner's referential check.
- `label_definitions_owner_key UNIQUE (organization_id, id, owner_scope_type, owner_scope_id)` is the composite target of
  the assignment foreign key below (the `business_profiles_organization_id_key` trick, TENANCY.md L96). **r1:** the
  key uses the non-nullable `owner_scope_id` instead of the nullable `group_id` (finding 1).

**LabelAssignment** `app.label_assignments`: `id uuid` (core-generated, nonzero), `organization_id NOT NULL`,
`label_id NOT NULL`, `owner_scope_type NOT NULL`, `owner_scope_id NOT NULL` (copied from the definition and enforced
by the composite FK; **r1:** replaces the nullable `owner_group_id` of r0), `target_type text NOT NULL`, `target_id uuid
NOT NULL`, `target_group_id uuid NOT NULL` (the Group that owns the target: the Group itself for `group`, the exact
owning Group for a future Group-owned target), `created_at`.

- **r1 (finding 1)** `label_assignments_label_fkey (organization_id, label_id, owner_scope_type, owner_scope_id) →
  label_definitions (organization_id, id, owner_scope_type, owner_scope_id)` ON DELETE RESTRICT. Every referencing column
  is `NOT NULL`, so under MATCH SIMPLE the check is total: a raw runtime INSERT with an absent, foreign-Organization or
  owner-mismatched (`owner_scope_type`/`owner_scope_id` differing from the definition's) label id is refused with 23503
  regardless of scope — the Organization scope no longer skips the FK. `owner_scope_type`/`owner_scope_id` are copied
  facts, not authority: RLS (§3) still decides who may write. Deleting a definition with assignments is refused by the
  FK (23503) unless the facade's two-statement delete removed them first (§1.3), so hidden orphans are impossible.
- `label_assignments_target_group_fkey (organization_id, target_group_id) → groups (organization_id, id)` ON DELETE
  RESTRICT (a Group with labels cannot be deleted silently; Groups have no runtime DELETE today).
- **r1 (finding 4 / D-2 closed)** `label_assignments_target_shape`: `owner_scope_type = 'organization' AND owner_scope_id =
  organization_id AND target_type = 'group' AND target_group_id = target_id`. Production storage admits exactly the one
  production target kind: an Organization-owned label on a same-Organization Group whose existence is enforced by
  `label_assignments_target_group_fkey`. A Group-owned assignment row is impossible in production storage until a future
  registered type's own migration widens this CHECK under the registration contract (§1.4); the r0 grammar openness
  (`^[a-z][a-z0-9_]{0,31}$`) is withdrawn.
- `label_assignments_unique UNIQUE (label_id, target_type, target_id)` (Spec §1); duplicate assign → `conflict`
  (`already_assigned`), zero writes, no audit row (D-6).
- Keys immutable (every column; an assignment is never updated, only inserted or deleted; no runtime UPDATE grant).
- Indexes: `label_assignments_target_idx (organization_id, target_type, target_id)`, `label_assignments_group_idx
  (organization_id, target_group_id)`.

Relations added: 2 tables + 2 primary keys + 2 unique keys on definitions + 1 unique key on assignments + 2 indexes =
**9** (71 → 80; the exact count is pinned by `TestMigrationLifecycle` at implementation; the Plan does not claim it).
r1 changes constraints and one column, not the relation count. Triggers: **five** (r1 tally, replacing r0's "four") —
`label_definitions_keys_immutable`, `label_assignments_keys_immutable`, `label_definitions_bound`,
`audit.capture_label_definitions`, `audit.capture_label_assignments`.

### 1.3 Decision inputs and lock order

Every label operation is a `guard` call (E2). Facts: Organization row (`FOR SHARE` on writes, **except** definition
creation in the Organization scope, below), own bindings, the candidate Group rows (target or owning Group `FOR NO KEY
UPDATE` on writes through `groupQuery`), then (new, after AUTHORIZATION.md step 7 and before the write) the **label
definition row** `FOR NO KEY UPDATE` for definition update/delete and `FOR SHARE` for assign/unassign. No existing
writer locks label rows and label operations never lock a row an existing writer holds in a different order (Group
rows first, then label rows), so no new wait cycle is added; two assigns of one label to two Groups hold the definition
`FOR SHARE` (compatible). A definition delete takes the definition `FOR NO KEY UPDATE` **without** locking Group rows,
then deletes its assignments and the definition in two statements; an assign that holds the definition `FOR SHARE`
blocks the delete until it commits, after which the delete's DELETE statement (READ COMMITTED, fresh statement
snapshot) removes the new row too — no orphan (O-LAB-1). An assign that starts after the delete's lock waits and then
finds no definition → `not_found`. A visibility narrowing is a definition UPDATE under the same lock; every later read
(REPEATABLE READ snapshot taken after the commit) observes it (O-LAB-2); no session, MCP catalog or web state caches
label visibility (the G060 A21 argument).

**r1 (finding 2) — scope-row lock for definition creation is taken exclusively up front; no upgrade.**
`CreateLabelDefinition` in the Organization scope uses the existing `policyWriteQuery` (`facade.go:139-144`:
`organizationQuery` with `ExclusiveOrganization`, `facts.go:75-76`), so the Organization row is locked `FOR NO KEY
UPDATE` at AUTHORIZATION.md step 3 — the same up-front rule `SetOrganizationPolicy` follows ("so a shared lock is never
upgraded", `facade.go:201-203`, AUTHORIZATION.md L413-414). In the Group scope it uses `groupQuery(g, write)`
(`facade.go:115-125`), which locks the owning Group `FOR NO KEY UPDATE` at step 6 (`TargetGroup`). The
`label_definitions_bound` trigger then re-locks the scope row `FOR NO KEY UPDATE` **in the same transaction at the same
strength** (a re-lock, not a conversion) before counting; for a raw runtime INSERT that took no facts the trigger's lock
is the only one, which is the intended backstop. Consequence for T6 O-LAB-4: two concurrent Organization-scope creates
serialize on the Organization row at step 3 (one waits, then counts 200 and refuses `scope_full`); two Group-scope
creates in one Group serialize on the Group row at step 6; a Group-scope create and an Organization-scope create of the
same Organization do not block each other (the Group create holds the Organization `FOR SHARE`, compatible with nothing
held by the other until its step 3 `FOR NO KEY UPDATE`, which conflicts with `FOR SHARE` — so they serialize without a
cycle: both take the Organization row first, in step order). Only creation takes the exclusive scope lock; update/
delete/assign/unassign keep the r0 strengths (Organization `FOR SHARE`, definition row lock), since the bound trigger
fires on INSERT only. Cost: Organization-scope definition creation serializes with policy writes and enrollment
approvals of the same Organization (all `FOR NO KEY UPDATE` on the Organization row) — acceptable for an administrator
operation bounded at 200 rows per scope.

**r1 (finding 3) — Organization-owned definition deletion requires Organization-wide context.** The assignment DELETE
policy is selected-Group filtered (`G IS NULL OR target_group_id = G`, §3), so a delete issued with a selected Group
could never remove the assignments of other Groups and, with the r1 FK, would be refused by 23503 — which would reveal
through the error whether other Groups carry the label. Therefore `DeleteLabelDefinition` refuses **before any SQL**
when the trusted context has a selected Group (`TenantContext.group` nonzero, `context.go:89`) and the definition
turns out, after the same-facts read, to be Organization-owned: `invalid_input` (`Field` `organization_context_required`),
zero writes, no audit row. The refusal is uniform (it does not depend on whether any assignment exists in any Group)
and non-enumerating; the trusted context is never cleared or widened by the operation, RLS is not bypassed, and the
`assignments_removed` count is complete because it is only ever produced under the Organization-wide context. A
Group-owned definition is deletable under Organization-wide context or with its own Group selected (every assignment of
a Group-owned definition has `target_group_id = owner_scope_id`, so the filtered DELETE is complete); with a different
Group selected the definition itself is invisible (`G IS NULL OR group_id = G`) → `not_found`, the same-facts read rule.
The order inside the operation is: facts → read the definition (owner scope) → context check → definition
`FOR NO KEY UPDATE` → DELETE assignments → DELETE definition. HTTP/web: the BFF forwards the session's context unchanged;
`/organization/labels` renders the closed status "Deleting an Organization label requires the Organization-wide
context" (`?status=organization_context_required`, `data-testid="labels-status"`), never a partial deletion.

Target liveness: Groups have no runtime DELETE, so "racing target deletion" is the FK RESTRICT backstop plus the
`FOR NO KEY UPDATE` on the target Group (an assign waits for a concurrent rename/slug change and re-reads).

### 1.4 Target registry (Spec §1 "closed typed core target-validation contract")

```go
// core/internal/tenancy/access/labels_targets.go (production): the closed set.
type labelTargetKind string                      // "group" only in production
type labelTarget struct{ Kind labelTargetKind; ID tenancy.UUID; Group tenancy.GroupID }
type labelTargetAdapter interface {
    // resolve reads the target inside the operation's transaction from authoritative storage and returns its
    // owning Group (and existence); the caller supplies nothing but the id.
    resolve(ctx, tx *trackedTx, org tenancy.OrganizationID, id tenancy.UUID) (owning tenancy.GroupID, ok bool, err error)
    // readable decides current read authority on the resolved target for the caller (target-member visibility).
    readable(sel contextSelection, f Facts, owning tenancy.GroupID) bool
}
var labelTargets = map[labelTargetKind]labelTargetAdapter{labelTargetGroup: groupTargetAdapter{}}
```

- `groupTargetAdapter`: `resolve` = the Group row already in `Facts` (locked), `readable` = `decide(capGroupRead, …)`.
  It is admitted **only** for Organization-owned definitions (`label_assignments_target_shape` and the core check);
  a Group-owned definition targeting a Group is `invalid_input` (`target_unsupported`).
- Unknown/unregistered kinds refuse before any SQL (`invalid_input`, `Field` `target_unsupported`); ambiguous selectors
  (both `group_id` and another target field) refuse (`invalid_input`, `target_ambiguous`).
- **r1 (finding 4) — what production SQL guarantees versus what the test adapter proves.** Production storage (§1.2,
  §3) guarantees, for every assignment row: the definition exists with exact Organization/owner facts (the total FK);
  the target kind is `group`; the target Group exists in the same Organization (`label_assignments_target_group_fkey`);
  and the row is readable only by a subject who can read the definition, which for an Organization-owned definition
  means `MGR` (org_owner/org_admin) — and `capGroupRead` holds for `orgAdmin` on every Group of the Organization
  (`policy.go:280-282`, "the Organization path holds regardless of the override"). So in production, visible definition
  ⇒ existing, readable target, **by construction of the closed CHECK and FK**, not by an extension guarantee. The
  r0 claim that "an assignment is visible exactly when its definition is" is restated as this closed statement.
  **Residual, explicit:** production SQL carries no per-type target-visibility predicate because no Group-owned target
  type exists; the registration contract below is what a future type must satisfy, and it is proven only with the
  test-only widening — it is not evidence of a live production Group-target guarantee or leak.
- **r1 — registration contract for a future target type `T`** (documented in `labels_targets.go` and TENANCY.md; owned
  and reviewed by the Goal that introduces `T`): (1) a core adapter (`resolve` from authoritative storage in the same
  transaction, `readable` from the current facts); (2) a migration that widens `label_assignments_target_shape` with
  `(owner_scope_type = 'group' AND target_type = 'T' AND target_group_id = owner_scope_id)`; (3) **r2 (R1 item 1)** a
  per-type, NULL-skipping **existence-and-ownership key**: a stored generated column `target_T_id uuid GENERATED ALWAYS AS
  (CASE WHEN target_type = 'T' THEN target_id END) STORED` (the `storage_key` precedent, `00007:133`; PostgreSQL 17 per
  `run.sh:4105`) and the foreign key `(organization_id, target_T_id, target_group_id) → T (organization_id, id, group_id)`
  (`T` carries `UNIQUE (organization_id, id, group_id)`, the `business_profiles_organization_id_key` trick) with `ON UPDATE
  RESTRICT` (a generated referencing column admits only RESTRICT/NO ACTION on update) and `ON DELETE RESTRICT` or `CASCADE`
  (`T`'s decision). Under MATCH SIMPLE the key column is NULL on every row of another type, so the FK constrains **only
  `T` rows** and ordinary production `group` rows coexist untouched; on a `T` row it binds the target's existence in the
  same Organization **and** its actual `group_id` to `target_group_id`, which the widened CHECK equates to
  `owner_scope_id`, i.e. to the definition's owning Group through the composite definitions FK — a subject who manages
  two Groups cannot point a research-owned assignment at a design-owned target. r1's two-column `(organization_id,
  target_id) → T` form is withdrawn: it constrained every row regardless of type and bound no ownership; and (4) a
  disjunct in `label_assignments_select` requiring the target row to be visible to the invoker under `T`'s own RLS and
  to carry the row's ownership facts (**r2:** `EXISTS (SELECT 1 FROM T t WHERE t.organization_id = organization_id AND
  t.id = target_id AND t.group_id = target_group_id)`), so that target-member visibility holds in storage, not only in
  core. Core and SQL must agree (`readable` ⇔ the EXISTS), which the fixture proves for the template.
  `TestLabelTargetsClosed` pins the core map at `{group}`; `TestLabelsCatalog` CT-LAB-2 pins the production CHECK body,
  the SELECT policy body and the production column list of `label_assignments` as `group`-only (no per-type key column
  exists in production storage; the first one arrives with `T`'s own reviewed migration).
- **Test-only fixture adapter and schema widening** (`labels_fixture_test.go`, build tag `integration`): registers kind
  `fixture` through an unexported `labelHooks` pointer that is nil in production (the `objectHooks`/`mcpHooks`
  precedent, `access/access.go:45-46`, `access/mcp.go:33-38`). The superuser `itSetup` of the disposable database (never
  a `core/migrations` file; `migrate/testdata` precedent, MIGRATIONS.md L127) creates `app.label_fixture_targets
  (organization_id, id, group_id, members_may_read boolean)` with FORCE RLS and one SELECT policy (`rls_group_role
  (group_id)` manager, or member when `members_may_read`), then applies the contract's (2)–(4) for `fixture` by
  replacing the CHECK and the SELECT policy and adding the FK; the teardown restores the production bodies, and
  `TestLabelsCatalog` (its own database, migrate package) pins the production bodies independently. The adapter's
  `resolve` reads `label_fixture_targets` in the operation's transaction; `readable` mirrors the fixture policy.
  **r2 (R1 item 1) — the exact test-only storage.** `label_fixture_targets` gets `PRIMARY KEY (organization_id, id)`,
  `UNIQUE (organization_id, id, group_id)` and `FOREIGN KEY (organization_id, group_id) → groups (organization_id, id)`.
  The widening on `label_assignments` is: (2) `label_assignments_target_shape` replaced by the production disjunct `OR
  (owner_scope_type = 'group' AND target_type = 'fixture' AND target_group_id = owner_scope_id)`; (3) `ALTER TABLE
  app.label_assignments ADD COLUMN target_fixture_id uuid GENERATED ALWAYS AS (CASE WHEN target_type = 'fixture' THEN
  target_id END) STORED` and `ADD CONSTRAINT label_assignments_fixture_target_fkey FOREIGN KEY (organization_id,
  target_fixture_id, target_group_id) REFERENCES app.label_fixture_targets (organization_id, id, group_id) ON UPDATE
  RESTRICT ON DELETE RESTRICT`; (4) the SELECT disjunct of §3. Consequences: a production-shape row (`target_type =
  'group'`) has `target_fixture_id IS NULL`, so the fixture FK never evaluates it and the production CHECK, Group FK and
  policies are the only constraints on it — ordinary Group assignments coexist with fixture rows without any fake
  fixture row and without altering the production target constraint; a `fixture` row must name a fixture target that
  exists in the same Organization and is owned by exactly `target_group_id` (= `owner_scope_id` by the CHECK = the
  definition's owning Group by `label_assignments_label_fkey`); moving (`UPDATE … SET group_id`) or deleting an
  assigned fixture target is refused 23503. The generated column is invisible to core: core statements name their
  columns (`TestLabelStatements`/`TestSQLBoundary` pins, no `SELECT *`), the runtime roles' column-limited INSERT list
  excludes it (a generated column cannot be inserted anyway), `label_assignments_keys_immutable` compares the
  production columns and `audit.capture_label_assignments` reads only `label_id`/`organization_id`/`target_group_id`.
  The teardown drops the FK, the column and the fixture table and restores the production CHECK/SELECT bodies; the
  widening exists only in the disposable database of the suites that register the fixture (`TestLabels`,
  `TestRLSIsolation`), never in a migration file. Negatives and the coexistence positive: T5 (d)–(g).
  Production startup registers nothing and the widening exists in no production path (`TestLabelTargetsClosed`;
  `TestPackageBoundary` pins that no non-test file names `fixture`; the fixture SQL lives only in `*_test.go`). No HTTP
  route or MCP tool reaches a Group-owned assignment; the UI says so (§1.5). The fixture is the extension **template**
  proof demanded by Spec §1/AC4, labelled as test-only evidence in the Result, never as production coverage.

### 1.5 Web and BFF (Spec §4; minimal, existing shell; canonical Vestive skill `.agents/skills/vestive-design/SKILL.md`)

- `web/lib/labels-client.ts`: `listLabels(scope, groupId?, after?)`, `createLabel`, `updateLabel`, `deleteLabel`,
  `listGroupLabels(groupId)`, `assignLabel`, `unassignLabel`; the membership-client `call`/`readBounded`/`hasExactly`/
  `failure` pattern (pinned by a CopyIsExact-style unit test), closed DTO key sets (`LABEL_KEYS` =
  `id, scope, group_id, name, visibility`; `GROUP_LABEL_KEYS` = `label_id, name`), byte budgets (row ≤ 400 B, page of
  100 + envelope), closed `?status=` maps, the form routes (`labelActionRoute`, `groupLabelActionRoute`).
- `web/app/organization/labels/page.tsx` + `action/route.ts`: Organization label definitions for owners/admins (list
  with rename and delete, one create form: name only; visibility fixed "Organization administrators" and stated); anyone
  else the fixed 404 (`refused("not_found")`, the group-authority page pattern). Linked from `/organization/members`
  `staffLinks` and `/account` (owner/admin gates already there). `data-testid`: `label-list`, `label-create-form`,
  `label-row`, `label-rename-form`, `label-delete-form`, `labels-status`. r1: the closed `?status=` map gains
  `organization_context_required` ("Deleting an Organization label requires the Organization-wide context"); the page
  never deletes partially (§1.3).
- `web/app/groups/[id]/page.tsx`: a "Labels" section rendered **only** when `listGroupLabels` answers `ok` (core
  decides; Organization owner/admin): chips (StateBadge-style `<span>`, `data-testid="group-label"`), an assign form
  (native `<select>` of unassigned Organization-owned definitions, the existing `components/ui/select`) and per-chip
  remove forms, posting to `app/groups/[id]/labels/action/route.ts`. A `forbidden`/`not_found` answer renders
  **nothing** (no hidden markup, no empty heading; Spec §4 "not serialized into member HTML/JSON and hidden with
  CSS"); `unavailable` renders the existing unavailable notice without the section.
- `web/app/groups/[id]/labels/page.tsx` + `action/route.ts`: Group-owned definitions. Owner/admin of the exact Group
  (core decides `label.group.manage`): list, create (name + visibility radio "Group administrators" / "Group members"),
  rename, change visibility, delete, and the fixed sentence "Group labels cannot be applied to anything yet: no
  Group-owned object type exists in this foundation" (`data-testid="labels-no-targets"`). A Group member: the
  read-only list of `target_members` definitions (`label.group.read`; `data-testid="labels-read-only"`). A readable
  Group without a Group role (Organization staff, override on or off): the notice `labels-not-administered` ("Labels of
  this Group are managed by its owners and admins") and **no** list — never the definitions. Absent/foreign/unreadable
  Group: the fixed 404. Linked from `/groups/[id]` only when `self_role` is set (an actual Group role); never from the
  `can_manage` override hint.
- Forms: native `<fieldset>`/radio, `PendingSubmit`, hidden `csrf_token`, exact-Origin `readForm`, 303 to the closed
  `?status=` map, "Not permitted" on `forbidden`, the honest unknown-state page on `unavailable`, light/dark through
  the semantic tokens, keyboard-only operation; no new dependency, kit module, `UI_FILES`/`UI_SLOTS` or B-DEPS change
  (D-7). Names render as escaped text; no colours.

## 2. Files and modules

**Create**

- `core/migrations/000NN_labels.sql` (§3; NN = the next free number at execution, 18 at `c1e724c6`; not reserved).
- `core/internal/tenancy/labels.go` (model types, `ValidateLabelName`, store statements: definition/assignment reads,
  inserts, update, deletes, locks, keyset pages), `core/internal/tenancy/labels_test.go` (validation, statement pins).
- `core/internal/tenancy/access/labels.go` (guarded operations §4.1), `labels_targets.go` (registry §1.4),
  `labels_test.go` (unit: capability rows, registry closure, denial mapping), `labels_integration_test.go`
  (`TestLabels` LAB-* and `TestLabelsOrdering` O-LAB-*), `labels_fixture_test.go` (integration-tagged fixture adapter).
- `core/internal/labels/api/http.go`, `http_test.go` (§4.2; shared helpers byte-copied from `membership/api` with a
  `TestLabelsCopyIsExact` pin).
- `core/internal/migrate/labels_integration_test.go` (`TestLabelsCatalog` CT-LAB-*).
- `web/lib/labels-client.ts`; `web/components/labels/{label-list,label-form,group-labels-section}.tsx`;
  `web/app/organization/labels/{page.tsx,action/route.ts}`; `web/app/groups/[id]/labels/{page.tsx,action/route.ts}`;
  `web/tests/unit/{labels-client,labels-forms}.test.mjs`; `web/tests/browser/labels.spec.ts` (E-LAB-1..6).
- `management/v0/results/G020-generic-labels-visibility-result.md`; `management/v0/results/G020-scope-allowlist.txt`
  (Codex import of §11).

**Modify — core**

- `access/policy.go` (3 capabilities, `capabilityKind`, `readCapability`, `decide` rows, `eligible`, header comment
  "nineteen" → "twenty-two"), `policy_test.go` (`allCapabilities`), `matrix_test.go` (`oracleCapabilities` 19 → 22,
  tables with the three new rows and the override-on/off subjects, cell count 627 → re-pinned, `TestToolEligibility`),
  `access/log.go` (7 `opLabel*` constants, 45 → 52), `access/audit.go` (`targetScope`: the label kinds → resource
  `label`), `access/errors.go` (fields `name_conflict`, `scope_full`, `already_assigned`, `not_assigned`,
  `target_unsupported`, `target_ambiguous`, `visibility_not_allowed`; 23505/23514 mapping for the two tables),
  `access/boundary_test.go` (`accessFiles` + 3, exported pins, `inContextTx` sites 18 → unchanged: all through `guard`),
  `rls_expected_test.go` (two new tables in `rlsExpectedFixture`/`rlsExpectedCases`/`Inserts`/`Mutations`),
  `attack_expected_test.go` (label rows), `rls_integration_test.go`/`attack_integration_test.go` (fixture rows),
  `mcp_test.go`/`mcp_integration_test.go` (fixture tools for the 3 capabilities; `TestMCPDiscovery` staff/override
  rows), `tenancy/boundary_test.go` (19 → 20 production files, store statements 86/24 → re-pinned, `writeTargets` + 2
  tables, `lockingStatements` + 2), `tenancy/errors.go` (codes), `core/internal/audit/audit.go` + `audit_test.go`
  (actions 46 → 51, reasons 70 → 72, operations 65 → 72, resource types 11 → 12, parity source → the new migration),
  `core/migrations/migrations_test.go` (set + 1), `migrate/runner_test.go` (18 sources), `migrate/integration_test.go`
  (version-NN lifecycle strings, down NN → NN−1 refusal, relation count 71 → 80, `metadataThroughVersion`),
  `migrate/audit_integration_test.go` (catalog constants), `access/integration_test.go` + `recovery_integration_test.go`
  + `tenancy/integration_test.go` + `mcp/env_integration_test.go` (`current_version` 17 → NN), `access/audit_integration_test.go`
  (AT-INV observed set + 5 actions), `core/cmd/server/main.go` (register the labels handler; digest re-pinned).

**Modify — web**: `web/app/groups/[id]/page.tsx` (Labels section; Group labels link), `web/app/organization/members/page.tsx`
and `web/app/account/page.tsx` (link), `web/tests/browser/membership.spec.ts` only if S-NAV/S-LANDMARKS rows name the new
link (expected: no change), `web/playwright.config.ts` (D-10: `labels-light/dark` projects on `labels.spec.ts`).

**Modify — harness/CI/docs**: `infra/validation/run.sh` (new `labels_checks` in `suite_migrations`: B-LABELS-CATALOG,
B-LABELS-IT, B-LABELS-ORDERING; `CORE_LABELS_ROUTES` + B-MCP-SOURCE pins; `MAIN_GO_G016_SHA`/`MAIN_GO_IMPORTS`; importer
allow-list for `core/internal/labels/api`; B-TENANCY-HTTP probes of the served roots; B-MIGRATE-SET/CLI/CATALOG version
and object-count literals 17/71 → NN/80; AUDIT-INV/B-AUDIT-CATALOG 46 → 51 and source file; `WEB_ROUTES`; request-context
importer allow-list for `lib/labels-client.ts`), `infra/validation/lib.sh` (`BROWSER_MATRIX_IDENTITY_E2E` + E-LAB row),
`infra/validation/required-checks.tsv` (+3 rows, header counts), `infra/ci/profile-paths.tsv` (`foundation dir
core/internal/labels/` — the 00009/00011 precedent, so a labels-only change selects full), `docs/foundation/IDENTITY.md`
(`identity:browser` `--project=` list + 2), `infra/validation/tools/internal/logrecord/events.json` + `logrecord_test.go`
(`labels.request` event; 7 operations and 3 capabilities in `access.decision`), `logrecord/completion.schema.json` (the
labels routes), docs §9. `.gitlab-ci.yml`: no change.

## 3. Schema and migration — `core/migrations/000NN_labels.sql`

Slot: `ls core/migrations` ends at `00017_group_authority.sql` at `c1e724c6`; **00018 is free at planning time and is
not reserved** (re-verified at release; G060 attempts 6–7, if any, could add a migration first). `00006`–`00009` are
SHA-pinned (`migrations_test.go:76-85`); 00017 is not edited. Up, in one transaction:

1. Precondition `DO`: owner, the helpers, version NN−1 shapes (no label relation, the 00017 catalog constraint bodies,
   the policy reader/claim writer inventories 9/7, relation count 71).
2. `CREATE TABLE app.label_definitions` and `app.label_assignments` (§1.2), indexes, ENABLE + FORCE RLS; the two
   `*_keys_immutable` triggers (owner-owned SECURITY INVOKER, SQLSTATE 23000, the 00003 pattern) and
   `label_definitions_bound` (owner-owned, scope-row lock then count, limit 200; the `invitations_bound` pattern).
3. Runtime grants (column-limited): `label_definitions` SELECT; INSERT (`id, organization_id, owner_scope_type,
   owner_scope_id, group_id, name, visibility`); UPDATE (`name, visibility`); DELETE. `label_assignments` SELECT; INSERT
   (`id, organization_id, label_id, owner_scope_type, owner_scope_id, target_type, target_id, target_group_id`; r1 column set); DELETE;
   no UPDATE. No helper, PUBLIC, retention or migrator grant; no new SECURITY DEFINER; no new role.
4. Policies (`TO rt, ow`, command-specific; macros as 00017):
   - `label_definitions_select USING (organization_id = O AND R IS NOT NULL AND ((owner_scope_type = 'organization' AND
     MGR) OR (owner_scope_type = 'group' AND (G IS NULL OR group_id = G) AND (GMGR OR (visibility = 'target_members' AND
     app.rls_group_role(group_id) IS NOT NULL)))))`.
   - `label_definitions_insert WITH CHECK (organization_id = O AND R IS NOT NULL AND ((owner_scope_type = 'organization'
     AND MGR) OR (owner_scope_type = 'group' AND (G IS NULL OR group_id = G) AND GMGR)))`;
     `label_definitions_update USING/WITH CHECK` and `label_definitions_delete USING` the same manager predicate.
   - `label_assignments_select USING (organization_id = O AND R IS NOT NULL AND (G IS NULL OR target_group_id = G) AND
     EXISTS (SELECT 1 FROM app.label_definitions d WHERE d.id = label_id AND d.organization_id = organization_id) AND
     target_type = 'group')` — **r1 (finding 4):** the definition policy applies inside the EXISTS under the invoker's
     RLS (graph: assignments → definitions → readers; acyclic, no 42P17) and the explicit `target_type = 'group'`
     disjunct is the production target-visibility predicate (target existence by the Group FK; every reader of an
     Organization-owned definition holds `group.read` on the target, §1.4). A future type adds its own disjunct under
     the registration contract; the test-only widening adds `OR (target_type = 'fixture' AND EXISTS (SELECT 1 FROM
     app.label_fixture_targets t WHERE t.organization_id = organization_id AND t.id = target_id AND t.group_id =
     target_group_id))` in the disposable database only (**r2:** the predicate carries the same Organization and
     owning-Group facts as the fixture FK of §1.4, so a readable fixture target in the wrong Group is not a visible row).
   - `label_assignments_insert WITH CHECK (organization_id = O AND R IS NOT NULL AND (G IS NULL OR target_group_id = G)
     AND ((owner_scope_type = 'organization' AND MGR AND target_type = 'group') OR (owner_scope_type = 'group' AND
     app.rls_group_role(target_group_id) IN ('group_owner','group_admin'))))`; `label_assignments_delete USING` the
     same. (r1: the Group-scope branch is the type-independent owner-authority rule and stays in production; it admits
     no row in production because the CHECK admits no Group-owned assignment. Deleting an Organization-owned
     definition's assignments from a selected-Group context is refused in core before SQL, §1.3.)
   - No policy names `group_management_override` or `pg_trigger_depth()` (the 00017 postcondition lists at
     `:1026-1028` and `:1039-1040` stay exact).
5. Audit: `audit.capture_label_definitions` (INSERT → `label.created`; UPDATE OF name → `label.updated` reason
   `renamed`; UPDATE OF visibility → `label.updated` reason `visibility_changed`; DELETE → `label.deleted`) and
   `audit.capture_label_assignments` (INSERT → `label.assigned`; DELETE → `label.unassigned`), owner-owned SECURITY
   INVOKER, resource `label` / the definition id, Organization, `group_id` = the owning Group (definitions of the Group
   scope) or the `target_group_id` (assignments); never the name, target type or target id of a non-Group target
   (D-4). Catalog CHECKs re-created: actions +5 (46 → 51), reasons +2 (`renamed`, `visibility_changed`; 70 → 72),
   operations +7 (65 → 72), resource types +1 (`label`; 11 → 12); `events_attribute` replaced with the five trigger
   actions in the `trig` list and their shape rule (`success`, no reason except `label.updated`'s two, no role/subject/
   TTL, Organization required, `resource_type = 'label'`), and `label` added to the `access.denied` resource list
   (`:835-836`). One validating scan of `audit.events` per widened constraint.
6. Postcondition `DO`: the two tables' columns/ACLs/FORCE, the 9 relations (80), the 8 policies' commands/roles/bodies
   (deparsed fragments: no `group_management_override`, no `pg_trigger_depth`), the **five** triggers (r1 tally: two
   `*_keys_immutable`, `label_definitions_bound`, two `audit.capture_label_*`), the capture functions' owner/ACL, the
   catalog constraints, the closed `label_assignments_target_shape` body (`target_type = 'group'` only, r1), the four
   foreign keys all validated (definitions → organizations, definitions → groups, assignments → definitions with every
   referencing column NOT NULL, assignments → groups), inventories unchanged (policy reader 9, claim writer 7).

Down (development/test only): refuses while any label row exists or `audit.events` holds a `label.*` action, a label
operation or a label reason (history is never discarded); otherwise drops the policies, triggers, functions and tables,
restores the 00017 constraint bodies and `events_attribute` byte-identically and revokes the grants. Production
rollback is "deploy the previous code with NN applied" (version-17 code never names the tables; the rows are inert).
Locks: ACCESS EXCLUSIVE on the new tables only, SHARE ROW EXCLUSIVE on `organizations`/`groups` for the foreign keys,
one validating scan of `audit.events`: migrate in a quiescent window, then deploy. No data migration, no backfill, no
edit to accepted migrations, no cascade across unrelated entities (definition → own assignments is an explicit
two-statement facade delete, FK RESTRICT as the backstop).

## 4. Contracts and interfaces

### 4.1 Go (`core/internal/tenancy/access`, guarded; capability per operation)

| Method | Operation | Capability (target) | Lock/facts | Write |
|---|---|---|---|---|
| `ListLabelDefinitions(tc, scope, g, after, limit)` | `label.definition.list` | Organization scope: `label.organization.manage` (Organization); Group scope: `label.group.manage` on g, falling back to `label.group.read` for the `target_members` subset (one decision each, same facts; the record carries the capability that allowed) | REPEATABLE READ READ ONLY | — (keyset by id, page 1..100 default 50, cursor bound to the scope) |
| `CreateLabelDefinition(tc, scope, g, name, visibility)` | `label.definition.create` | as above (manage) | **r1:** Organization scope `policyWriteQuery` (Organization row `FOR NO KEY UPDATE` up front, step 3); Group scope `groupQuery(g, write)` (owning Group `FOR NO KEY UPDATE`, step 6); the bound trigger re-locks the same row at the same strength (raw-SQL backstop, no upgrade) | one INSERT (core-generated id) |
| `UpdateLabelDefinition(tc, id, name, visibility)` | `label.definition.update` | read the row first (`guardEvaluated`), then manage of its owner scope | definition `FOR NO KEY UPDATE` after the facts | absolute set of `name` and `visibility`; `visibility_not_allowed` for the Organization scope before SQL; setting the current values is a successful no-op (no audit row: the capture's `WHEN` clauses skip no-ops) |
| `DeleteLabelDefinition(tc, id)` | `label.definition.delete` | as update; **r1:** an Organization-owned definition additionally requires Organization-wide context (`tc.group` zero) → otherwise `invalid_input` `organization_context_required` before any write | definition `FOR NO KEY UPDATE` | `DELETE label_assignments WHERE label_id`, then `DELETE label_definitions` (count observed; complete by construction, §1.3) |
| `ListLabelAssignments(tc, target)` (labels on one target) | `label.assignment.list` | `group` target: `label.organization.manage` **and** `group.read` on the exact Group; fixture target: `label.group.read`/manage on the owning Group plus the adapter's `readable` | READ ONLY | — |
| `AssignLabel(tc, id, target)` | `label.assign` | definition read first; Organization-owned + `group` target: manage **and** `group.read`(g); Group-owned + registered target: `label.group.manage`(owning Group) and the adapter's `resolve` in the same transaction | target Group `FOR NO KEY UPDATE`, definition `FOR SHARE` | one INSERT; 23505 → `conflict` `already_assigned` |
| `UnassignLabel(tc, id, target)` | `label.unassign` | as assign | as assign | one DELETE; zero rows → `conflict` `not_assigned` |

- Failures: `invalid_input` (`invalid_id`, `invalid_name`, `invalid_visibility`, `invalid_scope`, `target_unsupported`,
  `target_ambiguous`, `visibility_not_allowed`, `organization_context_required` (r1), `invalid_cursor`, `invalid_page_size`); `conflict` (`name_conflict`,
  `scope_full`, `already_assigned`, `not_assigned`); `not_found` for an unknown/foreign/outside-context/unreadable
  definition, Group or target (the same-facts read rule); `forbidden` for a readable one whose capability denies;
  `unavailable` for aborts/unknown COMMIT (no retry, no idempotency key; the writes are an INSERT, an absolute set and
  idempotent-by-conflict deletes, so a retry cannot accumulate effects).
- Records: the seven operations join `access.decision` and the `access.denied` operations; no name, target id or
  payload in any record. No `access.override_used` is ever appended for labels (no override path exists).
- `eligible` (MCP hint): the three capabilities follow `decide` (Organization staff eligible for
  `label.organization.manage` only; `label.group.*` only with an explicit role; override irrelevant).

### 4.2 HTTP (`/internal/v1`, web BFF only; `core/internal/labels/api`; no public API)

Seven method/path pairs on six paths, C-sorted, pinned by `TestLabelsRoutesPinned` and `CORE_LABELS_ROUTES`:

| Route | Body / query | Answer |
|---|---|---|
| `GET /internal/v1/labels?scope=organization\|group&group_id=&after=&limit=` | — | `{"labels":[{id, scope, group_id, name, visibility}], "next": cursor\|null}` |
| `POST /internal/v1/labels` | `{origin, csrf_token, scope, group_id?, name, visibility}` (exactly; `visibility` must be `owner_scope` for `organization`) | 200 the row |
| `POST /internal/v1/labels/{id}/update` | `{origin, csrf_token, name, visibility}` | 200 the row |
| `POST /internal/v1/labels/{id}/delete` | `{origin, csrf_token}` | 200 `{"deleted": true, "assignments_removed": n}` |
| `GET /internal/v1/groups/{id}/labels` | — | `{"labels":[{label_id, name}]}` (Organization owner/admin only; others `forbidden`/`not_found` as core) |
| `POST /internal/v1/groups/{id}/labels/assign` | `{origin, csrf_token, label_id}` | 200 `{"assigned": true}` |
| `POST /internal/v1/groups/{id}/labels/unassign` | `{origin, csrf_token, label_id}` | 200 `{"assigned": false}` |

Status mapping, `maxBody` 8 KiB, strict JSON, origin/session-binding/CSRF and the `labels.request` operational event
copy `membership/api` (E11). No route exposes a Group-owned assignment (no production target); the limit is stated in
the adapter comment and WEB-SHELL.md.

### 4.3 MCP

Production catalog unchanged (five control tools; `contracts/mcp/control-plane.md:111` stays true; no
`tools.listChanged`, no new tool, no public contract change). Parity proof: three test-only capability-mapped fixture
tools (`labels.list_organization`, `labels.list_group`, `labels.read_group`) through the existing fixture path
(`catalog_fixture_test.go`), asserting eligibility (`TestToolEligibility`, `TestMCPDiscovery`: frank with the override
on is eligible for the Organization tool and never for the Group tools; bob for the Group tools of research only) and
hidden = unknown (C-1) on direct call. **r1 (D-1 decided; Spec "MCP foundation clarification — provisional Plan review,
2026-10-20"):** the five production control tools are kept; the fixture capability/discovery/hidden-call evidence is
proof of finite label-capability eligibility, current-context refusal and hidden direct-call behaviour only, and the
Result labels it as such — never as a production label operation implemented or tested. A production tenant-data tool
remains a separately reviewed public-contract change; the r0 alternative is withdrawn from this Plan.

### 4.4 Web

`labels-client.ts` (§1.5), pages and components; DTO key pins `LABEL_KEYS`/`GROUP_LABEL_KEYS`; no change to
`GROUP_KEYS`/`MEMBER_KEYS` or any existing DTO (D-9).

## 5. Validation and test mapping

### 5.1 Existing coverage inspected (reuse unless stated)

Unit: `matrix_test.go` (oracle tables, 627 cells, `TestToolEligibility`), `policy_test.go`, `objects_test.go`,
`authority_test.go`, `audit_test.go` (`TestCatalogParity`), `tenancy/boundary_test.go`, `access/boundary_test.go`,
`migrations_test.go`, `migrate/runner_test.go`, `membership/api/http_test.go` (routes/copy/body pins), `mcp_test.go`
(`TestToolCatalog`, `TestCatalogFixtureIsolation`), web unit tests. Integration: `TestAccessPolicy`, `TestRLSIsolation`
(+`rls_expected_test.go`), `TestIsolationAttacks` (+`attack_expected_test.go`), `TestMembership`/`TestAuthority`,
`TestAccessOrdering`/`TestRLSOrdering`, `TestAudit` (AT-INV), `TestMCPDiscovery`, migrate catalogs
(`TestMigrationLifecycle`, `TestMembershipCatalog`, `TestEnrollmentCatalog`, `TestAuditCatalog`). Browser:
`membership.spec.ts` E-MEM-8/9 (settings and override pages), `onboarding.spec.ts` E-ONB-3/4/8, S-LANDMARKS/S-NAV.

### 5.2 Scenario groups → tests (reuse / extend / new; the distinct boundary of each new case)

| Group | Tests | Kind | Distinct boundary |
|---|---|---|---|
| T1 evaluator | `TestOracleTables`/`TestOracleMatrix`/`TestOracleGates` + 3 capability rows (subjects incl. frank with OVR on/off, Grace dual role, dave, foreign); `TestPolicyGates`/`TestPolicyScope`; new `TestLabelPolicy` (readCapability kinds, `not_found` vs `forbidden`) | extend + new | the three label rows did not exist; OVR must be a no-op for `label.group.*` |
| T2 registry/validation | new `TestLabelTargetsClosed` (production map = {group}; unknown/ambiguous/unsupported refuse before SQL; Group-owned + `group` target refused), `tenancy/labels_test.go` (`ValidateLabelName` bounds, trimming, control characters, 80-point limit, key collision cases) | new | the registry and the name grammar are new |
| T3 runtime RLS | `TestRLSIsolation` (`rlsExpectedFixture`/`Cases`/`Inserts`/`Mutations` + the two tables: **r1 (finding 5):** carol, frank and Grace (A admin — her independent Organization authority) read Organization-owned rows; bob, dave and Dana (Group roles only) see none of them; bob and Grace (research admin) read research's Group-owned rows, dave only `target_members`; frank with OVR on and carol with OVR off see no Group-owned row (override-without-Group-role negatives retained unweakened); Grace sees design's Group-owned rows **not at all** (no design role; Organization admin confers nothing); selected-Group shapes; foreign Organization rows invisible; raw INSERT/UPDATE/DELETE by each subject. **r1 (finding 1) / r2 (R1 item 2) raw-SQL substitutions on `label_assignments`.** Refusal order of a raw INSERT in PostgreSQL 17 (`ExecInsert`): RLS `WITH CHECK` (42501) → table CHECK/NOT NULL (23514/23502) → unique index (23505) → referential AFTER triggers (23503); every case below either admits exactly one refusal or names its expected first refusal, runs under its own savepoint, and is followed by a superuser count proving **zero** rows for the attempted key and **zero** `audit.events` rows (`label.assigned` for that definition/Group; the capture trigger is AFTER ROW and the statement rolls back whole). **Isolated FK cases** (valid production shape — `('organization', A)` owner facts, `target_type = 'group'`, `target_group_id = target_id` = research — by carol Organization-wide, so the INSERT policy and the CHECK admit and only `label_assignments_label_fkey` can refuse): (i) `label_id` unknown → 23503; (ii) a foreign Organization's definition id with A's `organization_id` → 23503; (iii) **owner-fact substitution:** bob's research-owned definition id supplied with Organization owner facts → 23503 (the definition exists, but its key is `('group', research)`; the composite key does not match). **Converse direction** (an Organization-owned definition id supplied with `('group', research)` facts): no production shape admits Group owner facts, so the FK cannot be isolated in production storage; the actual first refusal is asserted instead — by carol → 42501 (the Group branch of `label_assignments_insert` needs a research owner/admin role she lacks; RLS precedes the CHECK), by bob (research owner; the policy admits) → 23514 (`label_assignments_target_shape`, closed); r1's claim that this direction proves 23503 is withdrawn, and its isolated FK case lives under the test-only widening, T5 (f). (iv) a Group-owned definition with its exact copied facts and `target_type = 'group'` by bob → 23514 (closed CHECK; the policy admits bob). Paired positive: carol's exact copied owner facts insert on research (one row, one `label.assigned`). If the first b2 run observes a different first-refusal code for a multi-check case, the Result records the observed code and that pin is corrected under standing authority; the isolated cases' 23503 and the zero-row/zero-audit assertions are not negotiable. Deletion cleanup: raw `DELETE label_definitions` while an assignment exists → 23503 (RESTRICT), paired positive after the assignments are deleted), `TestIsolationAttacks` (cross-tenant/cross-Group label reads, counts and joins; A-rows with the other subjects' rows present) | extend | both tables and both scopes are new policy bodies; the total FK is new |
| T4 lifecycle | new `TestLabels` LAB-1..12: create/list/rename/visibility/delete in both scopes; name collision (ASCII case, exact non-ASCII); 200-bound; Organization-owned `target_members` refused; assign/unassign the real Group target; duplicate assign `already_assigned`; unassign twice `not_assigned`; delete removes assignments atomically; foreign/unknown/cross-Group/unsupported/ambiguous targets; member/owner/admin/staff/override positives and denials per §1 table. **r1 (finding 3) LAB-17:** one Organization-owned definition assigned to research and design; carol with research selected → `invalid_input` `organization_context_required`, both assignments and the definition still present, no audit row (negative); carol Organization-wide → `assignments_removed = 2`, definition gone, audit `label.unassigned` × 2 then `label.deleted` (positive); bob (research owner) deleting his Group-owned definition with research selected → complete; with design selected → `not_found` | new (`labels_integration_test.go`) | the operations did not exist; the context rule is new |
| T5 fixture adapter | `TestLabels` LAB-13..16 with the `fixture` kind under the test-only widening (§1.4): Group-owned assignment to a synthetic target owned by research (bob positive; Dana/frank-with-OVR/carol denied; a target owned by design refused `not_found`); `target_members` read by dave requires `readable`; the adapter's facts are read in the same transaction (an **unassigned** target moved by the superuser between setup and the assign is observed: bob's assign of a research definition to it is `not_found`; **r2:** moving an *assigned* target is refused by the fixture FK, (d)). **r1 (finding 4) direct-SQL observations in `TestRLSIsolation` (fixture rows added by `itSetup`):** (a) **unreadable target:** an assignment of a research `target_members` definition to a fixture target with `members_may_read = false` — dave's raw SELECT returns zero rows while bob's returns one; the core `readable` answers agree (parity pin); (b) **nonexistent target:** a raw INSERT with an unknown `target_id` → 23503 (fixture FK), and a superuser DELETE of an assigned fixture target → 23503 (RESTRICT in the template), paired positive with a live target; (c) ~~after the fixture teardown the CHECK/SELECT bodies equal the production bodies~~ (r2: superseded by (g)). **r2 (R1 item 1) wrong-Group negatives, coexistence and the converse FK case** (same refusal order, savepoint and zero-row/zero-audit verification as T3): (d) **assignment to the wrong Group under direct SQL:** a per-run dual-Group subject (Dana, design owner, granted `group_admin` in research through the existing `grantGroup` helper for this case) raw-INSERTs an assignment of bob's research-owned definition with its exact copied facts `('group', research)`, `target_type = 'fixture'`, `target_id` = a **design**-owned fixture target and `target_group_id` = research: the widened CHECK admits (`target_group_id = owner_scope_id`), the INSERT policy admits (she manages research), and only `label_assignments_fixture_target_fkey` refuses → 23503 (no fixture row `(A, that id, research)`); variant with `target_group_id` = design → 23514 first (CHECK: `target_group_id ≠ owner_scope_id`); variant by bob (research only) with `target_group_id` = design → 42501 first (policy: no design role); superuser `UPDATE label_fixture_targets SET group_id = design` on an assigned research target → 23503 (ON UPDATE RESTRICT). The core path for the same placement (`resolve` returns design ⇒ `not_found`, LAB-13) is asserted alongside, so core and SQL refuse the same cross-Group placement; (e) **production rows coexist:** in the widened database carol's ordinary production-shape assignment (Organization-owned definition → `group` research) inserts and reads normally while fixture rows exist (`target_fixture_id IS NULL`, the fixture FK is skipped under MATCH SIMPLE), the production `label_assignments_target_group_fkey` still refuses an unknown Group id (23503) and the production branch of the CHECK still refuses `target_type = 'fixture'` with Organization owner facts (23514); no fixture row exists for any Group; (f) **isolated FK, Organization→Group direction:** bob raw-INSERTs carol's Organization-owned definition id with `('group', research)` facts, `target_type = 'fixture'`, a research fixture target and `target_group_id` = research — the widened CHECK and the policy admit and the fixture FK matches, so only `label_assignments_label_fkey` refuses → 23503; (g) **teardown:** after the fixture teardown `target_fixture_id`, `label_assignments_fixture_target_fkey` and `label_fixture_targets` are gone and the CHECK/SELECT bodies and the `label_assignments` column list equal the production ones (byte comparison against `TestLabelsCatalog`'s pinned fragments) | new | Spec §1/AC4 extension-template proof; no production registration; the storage half of the registration contract is exercised; r2: exact target ownership and coexistence with production Group rows are proven under direct SQL |
| T6 ordering/revocation | `TestLabelsOrdering` O-LAB-1 (assign vs definition delete: serialized, no orphan either order), O-LAB-2 (visibility narrowed → next read empty; role removed → next read `not_found`; selected-Group switch), O-LAB-3 (duplicate assign race: one `already_assigned`, one row, one audit row), O-LAB-4 (**r1 (finding 2):** two Organization-scope creates racing the 200 bound serialize on the up-front `FOR NO KEY UPDATE` Organization lock: exactly one `scope_full`, 200 rows, no deadlock (40P01) observed; the Group-scope pair on one Group likewise on the Group row; a mixed Organization/Group pair commits both), `TestRLSOrdering` OR-8 (raw delete of a definition under an in-flight assign: with the total FK the raw delete is refused 23503 or the assign observes `not_found`, never an orphan) | new subtests in existing suites | label rows are new mutable facts in the linearization argument |
| T7 migration/catalog | `TestMigrationLifecycle` v NN (up/down/up, down refusal with rows and with history), new `TestLabelsCatalog` CT-LAB-1..4 (tables/ACLs/policies/triggers/functions; no `OVR`/`DEP` in label policies; inventories unchanged), `TestAuditCatalog`/`TestCatalogParity` (51/12/72/72), `TestNormalSetIsTheReviewedSet` | extend + new | version-NN delta |
| T8 audit | `TestAudit` AT-LAB (five actions with the right resource/Group, no name; `label.updated` reasons; no-op update records nothing; delete records `label.unassigned` × n then `label.deleted` in one transaction; an injected audit fault refuses the create/assign; unknown-COMMIT classification reused `TestCommitClassification`), AT-INV observed set | extend | new actions |
| T9 HTTP | new `TestLabelsHTTP`/`TestLabelsRoutesPinned`/`TestLabelsBodyShapes`/`TestLabelsCopyIsExact` (strict bodies, CSRF/origin refusals, unknown fields, size, member payload never contains labels, 404 parity for unknown/foreign ids) | new | new adapter |
| T10 MCP parity | `TestToolEligibility` (+3 capabilities), `TestMCPDiscovery` (fixture tools: staff/override/Group-role rows; hidden = unknown), `TestToolCatalog` (production still 5/0) | extend | discovery truthfulness for the new capabilities |
| T11 web unit | new `labels-client.test.mjs` (7 calls, strict keys, budgets, error map), `labels-forms.test.mjs` (closed fields, 303, statuses, "Not permitted", unknown page) | new | new client/forms |
| T12 browser | new `labels.spec.ts`: **E-LAB-1** carol creates/renames/deletes an Organization label on `/organization/labels` (keyboard), frank sees it, alice 404; **E-LAB-2** carol assigns/unassigns it on a Group page; bob (research owner) and dave see no Labels section and no label markup on research; the Group page HTML of bob contains no label name; **E-LAB-3** bob manages Group-owned labels on `/groups/<research>/labels` with the "cannot be applied" sentence; dave sees only `target_members` ones; frank (OVR on) sees `labels-not-administered`; carol (OVR off) the same; alice 404; **E-LAB-4** visibility narrowed by bob → dave's next load empty; carol removes dave's Group role → 404; **E-LAB-5** forged POSTs (alice assigns; bob creates an Organization label; frank creates a research label with OVR on) → "Not permitted"; **E-LAB-6** axe WCAG 2.2 AA at 1280/320 light+dark on the three pages, S-ERRORS; plus S-LANDMARKS/S-NAV on `/organization/labels` | new | UI truthfulness and privacy of the new pages; `lib.sh` matrix row; `--project=labels-*` (D-10) |

### 5.3 Batches (focused executable selectors; working directory `core/`; fetch/rebase/reconcile before each; private
per-run resources, no harness lock for qualifying runs; one build slot)

- **b1 unit (no DB, `docker run --rm --network none`, source `:ro`, `-mod=readonly GOPROXY=off`, disposable GOCACHE):**
  `go test ./internal/tenancy/access -run '^(TestPolicyGates|TestPolicyScope|TestOracleTables|TestOracleMatrix|TestOracleScope|TestOracleManageable|TestOracleGates|TestOracleCrossOrganization|TestToolEligibility|TestLabelPolicy|TestLabelTargetsClosed|TestAccessBoundary|TestRLSBoundary)$'`;
  `go test ./internal/tenancy -run '^(TestSQLBoundary|TestPackageBoundary|TestMigrationParity|TestLabelName.*|TestLabelStatements)$'`;
  `go test ./internal/audit -run '^(TestCatalogParity|TestStatements)$'`;
  `go test ./migrations -run '^(TestMigrationSet|TestEveryMigrationIsVersionedAndReversible)$'`;
  `go test ./internal/migrate -run '^(TestNormalSetIsTheReviewedSet|TestProtectionTargetsExactObjects)$'`;
  `go test ./internal/labels/api -run '^(TestLabelsHTTP|TestLabelsRefusalMapping|TestLabelsBodyShapes|TestLabelsRoutesPinned|TestLabelsCopyIsExact)$'`;
  `go test ./internal/mcp -run '^(TestToolCatalog|TestCatalogFixtureIsolation|TestPackageBoundary)$'`.
- **b2 runtime PostgreSQL (private per-run database, pgvector image as the harness pins, `-tags integration`):**
  `./internal/migrate -run '^(TestMigrationLifecycle|TestLabelsCatalog|TestAuditCatalog|TestMembershipCatalog)$'`;
  `./internal/tenancy/access -run '^(TestLabels|TestAccessPolicy|TestRLSIsolation|TestIsolationAttacks|TestAudit)$'`
  (whole `TestRLSIsolation`/`TestIsolationAttacks` justified: new tables in their fixtures and oracles);
  `./internal/tenancy/access -run '^(TestLabelsOrdering|TestRLSOrdering)$/^(O-LAB|OR-8)'`; `./internal/mcp -run '^(TestMCPDiscovery|TestMCPTransport)$'`.
  `TestAudit`/`TestIsolationAttacks` need the pinned MinIO images (reuse by verified ID; no rebuild).
- **b3 web unit:** `npm run test:unit`; lint/typecheck through the web check image (one build slot).
- **b4 browser (identity mode, private Compose stack):** the `IDENTITY.md` `identity:browser` line with
  `--project=labels-light --project=labels-dark --grep 'E-LAB|S-LANDMARKS|S-NAV'`; E-MEM-8/E-ONB-8 only if a shared
  page changed (expected: the Group page link only → rerun E-ONB-8's Group-page assertions).
- **Required final local coverage before the MR (affected union, changed-input provenance in the Result):** B-UNIT-CORE,
  B-UNIT-WEB, B-MIGRATE-SET/CLI/CATALOG/IT, B-LABELS-CATALOG/IT/ORDERING, B-POLICY-IT, B-RLS-CATALOG/IT/ORDERING,
  B-AUDIT-CATALOG/IT, B-MCP-IT, B-MCP-SOURCE, B-TENANCY-HTTP, B-WEB-BOUNDARY, B-IDENTITY-E2E (labels projects),
  B-SCOPE, B-REQUIRED-CHECKS. No default `full` locally before the MR; the coordinator owns the exact-merged-SHA local
  full after the authorized conditional merge; no automatic staging cloud run.

### 5.4 Change-impact inventory (consumers to reconcile before the first batch)

Migration version/file pins (`run.sh:1271,1283,1767-1803,2385-2428,2424,3038,3253-3254,3283,3290,3395,3419,3443,4119,
4674-4680,4727,4840-4858,5007`; `migrations_test.go:44-56`; `runner_test.go:308-313`; `migrate/integration_test.go:
764-800,941-965,1008-1030`; the four `current_version 17` test pins), capability pins (`matrix_test.go:76,878,1036,1373,
1477`; `policy_test.go:27`; `events.json:448-455`), audit counts (`audit_test.go:98,118`; AUDIT.md L93-96; `run.sh:3254,
3283`), route pins (`run.sh:178-196,1231-1247`; `completion.schema.json`), `WEB_ROUTES` (`run.sh:656`) and the
request-context importer list (`:900`), `required-checks.tsv` header counts, `lib.sh:643-648`, `IDENTITY.md:325`,
`profile-paths.tsv`, `tenancy/boundary_test.go:100,203,436`, `access/boundary_test.go:25,55,78`. Disabled variants: none
(no feature flag; labels are always on). Concurrency: §1.3 and T6.

## 6. Delegation

One implementation writer (Opus 5.5 high; test corrections Fable 5.1 high per CLAUDE.md), single write owner, **150k
checkpoint / 200k safe handoff** with supported native self-compaction at safe points (r1, Goal/Spec amendments of
2026-10-20; r0 said 300k/350k), phase handoff between S2 and S3 if needed. One bounded **read-only** security/data review
subagent after S2 (scope: the two policy sets, the registry, the fixture isolation, the audit shape, the UI's absence
of label markup for non-staff); findings consolidated by the writer. No delegated product/authority decisions, no
overlapping edits; the writer validates and integrates everything. Planning spent no attempt.

## 7. Implementation sequence and slicing

- **S1 (risky slice, schema + policy):** migration (§3), `tenancy/labels.go` statements, capabilities and `decide`
  rows, oracle rewrites (T1), RLS/attack fixtures and oracles (T3), catalog/lifecycle pins (T7), registry (T2);
  b1, then b2 minus ordering/audit. **Early-risk checkpoint** to Codex with the policy bodies and the RLS oracle
  outcome before S2.
- **S2 (facade, audit, HTTP, MCP parity):** guarded operations (§4.1), fixture adapter (T5), audit captures/shapes (T8),
  ordering (T6), `labels/api` (T9), MCP fixture tools (T10), log vectors; b1/b2 reruns of changed inputs; security review
  subagent; checkpoint.
- **S3 (web):** client, pages, components, unit tests, `labels.spec.ts`, matrix/project rows; b3, b4.
- **S4:** docs (§9), harness pins/record strings, Result (incl. the G055 workflow observation, §8 AC8), final affected
  coverage, safekeeping push, MR (fast + Codex + Duo), conditional merge, coordinator-owned local full, cleanup receipt.

No split is required (D-12): one migration, two tables, three capabilities, seven operations, three pages; S2→S3 is the
natural fresh-worker boundary.

## 8. Acceptance-criteria mapping

| AC | Proof |
|---|---|
| AC1 Model/lifecycle | T4 LAB-1..8 (both scopes, uniqueness/bounds, duplicate/removal/deletion semantics), T2 (grammar), T7 (`TestMigrationLifecycle` up/down/up, down refusal with rows and history on real PostgreSQL, `TestLabelsCatalog`), B-MIGRATE-* and B-LABELS-CATALOG |
| AC2 Current authority | T1 rows (owner/admin positives, member denials, dual-role Grace, frank OVR on/off), T4 LAB-9..12 (selected context; direct facade), T9 (direct HTTP), T10 (direct MCP fixture call = unknown); ordinary membership/label never elevates (T3 raw mutations denied) |
| AC3 Privacy | T3 (Group roles see no Organization-owned row; members see only `target_members` Group-owned rows), T4 (`visibility_not_allowed` for the Organization scope; counts/pages/errors non-enumerating), T6 O-LAB-2 (narrowing/revocation immediate), T9 (member payloads), E-LAB-2/3/4 (no label markup for bob/dave/alice; `labels-not-administered` for staff) |
| AC4 Target boundary | T4 (same-tenant Group positive; foreign/unknown/cross-Group/unsupported/ambiguous rejected), T5 (fixture adapter proves owner + target + same-transaction semantics without a production registration; r1: plus the storage half of the registration contract — unreadable and nonexistent fixture targets under direct SQL; r2: the NULL-skipping fixture FK binds exact target ownership — wrong-Group direct-SQL negatives (d), coexistence with production Group rows (e), teardown parity (g)), `TestLabelTargetsClosed`, `TestLabelsCatalog` CT-LAB-2 (production CHECK/SELECT bodies and column list `group`-only, r1/r2), `TestPackageBoundary` |
| AC5 Storage isolation | T3 `TestRLSIsolation`/`TestIsolationAttacks` direct-SQL negatives and paired positives on both tables (no cross-tenant/orphan row; r1: the total assignment→definition FK refuses unknown/foreign/owner-mismatched label ids in both scopes and RESTRICTs definition deletion with assignments; r2: the owner-fact FK is isolated by a valid-shape authorized case (Group-owned id with Organization facts → 23503) and by T5 (f) for the converse, whose production first refusals are pinned as 42501/23514; every refusal verified zero rows and zero audit records), T4 LAB-17 (selected-context deletion refused, Organization-wide deletion complete, r1), T7 grants/helper inventory unchanged (9/7), B-RLS-CATALOG/IT, legacy-path check: version-17 code against version NN (lifecycle) |
| AC6 Atomicity/audit | T8 (fault refuses the mutation; shapes/attribution/no names; no-op silence), T6 O-LAB-1/3/4 and OR-8 (concurrent duplicate/delete/bound), unknown-COMMIT handling (`TestCommitClassification` reused), T7 catalog parity 51/12/72/72, AT-INV and historical consumers (`events.json`, AUDIT.md) updated |
| AC7 Surfaces | E-LAB-1..6 (minimal accessible shell pages, light/dark/320 px, keyboard, axe, forged POST refusals), T9 (parser/bounds/CSRF/context; member payloads), T10 (HTTP/MCP parity through the same capabilities; hidden direct call denied), `TestLabelsCopyIsExact` |
| AC8 Delivery/workflow | current-head feature-MR fast + Codex + Duo disposition, conditional merge, exact-merged-SHA **local** full qualification (every new row selected: B-LABELS-*, E-LAB in the matrix), scoped resource/evidence/branch cleanup receipt, final Codex ACCEPT. The Result records, in one bounded section, how the accepted G055 `restruct-sdlc` entrypoint, discovery, roles, handoffs and cleanup were used (or, if G055 is not yet cut over at execution, that the legacy layout was used and why) — an observation, not a requalification; no G021 claim |

## 9. Adopter docs (supersession notes, not rewrites)

`docs/ARCHITECTURE.md` §53.8 (implementation status paragraph), `docs/SECURITY.md` (a "Labels — G020" note: private
metadata, no access control via labels), `docs/foundation/AUTHORIZATION.md` (capabilities 19 → 22, the §1 table, the
operations table, lock-order step for label rows; r1: the up-front exclusive Organization lock for Organization-scope
definition creation at step 3 and the Organization-wide-context rule for Organization-definition deletion),
`DATABASE-ISOLATION.md` (two tables, eight policies, readers unchanged), `TENANCY.md` (model, constraints, runtime
grants, a "Labels (G020)" section; r1: the total assignment→definition FK, the closed target CHECK and the
registration contract for a future target type), `AUDIT.md` (T40–T44 rows,
count sentence 51/12/72/72, resource `label`), `MIGRATIONS.md` (row NN; D-11 notes the missing row 16),
`MCP.md` (no production tool; eligibility of the three capabilities), `WEB-SHELL.md` (pages, test IDs, browser project
row), `IDENTITY.md` (`identity:browser` project list), `contracts/mcp/control-plane.md` (unchanged unless D-1 chooses
production tools), `docs/DECISIONS.md` (Codex-owned).

## 10. Decisions, risks and escalations (flagged for explicit review)

- **D-1 No production MCP label tool.** Parity is proven through capability eligibility with test-only fixture tools;
  the public MCP contract ("exactly five tools") is unchanged. Alternative: three production capability-mapped tools
  (`labels.list`, `labels.assign`, `labels.unassign`) — a public MCP contract change (AGENTS.md escalation list) and the
  first tenant-data tools, which MCP.md L141 currently excludes. ~~**Question for Codex/owner.**~~ **r1: decided** —
  keep the five production tools; fixture capability/discovery/hidden-call evidence is proof only (Spec clarification
  2026-10-20; review D1). `contracts/mcp/control-plane.md` is untouched; its allow-list pattern stays (28 patterns
  unchanged) but no edit to it is planned.
- **D-2 Group-owned target type in SQL.** The CHECK admits any grammar-valid non-`group` type for Group-owned rows
  (owning Group enforced), while the closed set is core's registry (one production kind); the test adapter needs the
  openness. Residual: a holder of the runtime credential could insert a Group-owned row with an unregistered type inside
  a Group it manages (within the G008 trust limit, DATABASE-ISOLATION.md L199-212). Alternative: `target_type IN
  ('group')` in SQL, which makes every Group-owned assignment impossible until a future migration and reduces AC4's
  extension proof to core-only. ~~**Question for Codex.**~~ **r1 (held pending finding 4; proposed resolution):** the
  closed alternative is adopted — production SQL admits `target_type = 'group'` for Organization-owned rows only
  (§1.2), the grammar openness is withdrawn and the G008 residual disappears. AC4's extension proof is **not** reduced
  to core-only: the test-only widening applied by `itSetup` exercises the full registration contract (CHECK, FK, SELECT
  predicate, core adapter) and its direct-SQL unreadable/nonexistent observations (§1.4, T5). Residual, explicit: the
  contract is a reviewed template; the Goal that introduces a real Group-owned target owns and proves its own migration.
  **r2 (R1 item 1):** R1 accepted the closed production set; the contract's existence FK is corrected to the per-type
  NULL-skipping existence-and-ownership key (§1.4 (3), generated column + three-column FK), which coexists with
  production `group` rows and binds the target's actual owning Group. The production CHECK stays closed to `group`.
  Remains **for Codex disposition** as the finding-4 reconciliation.
- **D-3 Three capabilities** (`label.organization.manage`, `label.group.manage`, `label.group.read`; 19 → 22) rather
  than reusing `group.manage_members`/`group.read` — the existing Group capabilities are held by Organization staff
  (metadata/override) and would leak Group-owned labels. Alternative: two capabilities with Group-member reads under
  `label.group.manage` + a visibility check in `act` (weaker: the record would not distinguish read from manage).
- **D-4 Audit shape:** five trigger-owned actions on one new resource type `label` (definition id as the resource; the
  owning or target Group as `group_id`; the target id of a non-Group target is not recorded — no column fits and the
  fixture is test-only); two reasons on `label.updated`. Alternative: a second resource type `label_assignment` with the
  assignment id (12 → 13) — more faithful for future targets, more pins now. ~~**Question for Codex.**~~ **r1: decided
  (review D4)** — one `label` resource with five actions for the current Group targets: `resource_id` = the definition
  id (exact definition identity), `group_id` = the actual owning Group (Group-scope definitions) or the actual target
  Group (assignments; equals the owning Group for a future Group-owned target), `organization_id` always; no name,
  target type, non-Group target id or other content in any event, and no new resource column. **Future target
  mapping, explicit:** the audit attribution of a non-Group target type (whether a second resource type, a target-id
  column or another shape) is decided by the Goal that registers that type and reviewed there; the test-only fixture
  establishes no product audit coverage for future targets and records only the owning Group. The alternative is
  withdrawn from this Plan.
- **D-5 Bounds:** name 1–80 code points (branding label precedent) under the `ValidateName` grammar; ASCII-case-insensitive
  uniqueness via a generated `lower(name COLLATE "C")` key (deterministic, no Go/PostgreSQL folding mismatch); 200
  definitions per owner scope; pages 1..100 default 50 (MCP/list precedent).
- **D-6 Repeat semantics:** duplicate assign → `conflict` `already_assigned`; unassign of a missing assignment →
  `conflict` `not_assigned`; update to the current values → success with no audit row. Alternative: idempotent 200 for
  both repeats (simpler clients, but hides a lost race from the caller).
- **D-7 UI placement:** `/organization/labels` (new page, owner/admin), a Labels section on `/groups/[id]` only for
  Organization staff (absent markup otherwise), `/groups/[id]/labels` for the Group's own owners/admins/members; chips
  as StateBadge-style spans (no kit `badge.tsx` adoption, so no `UI_FILES`/B-DEPS/lineage work). Alternative: adopt the
  kit badge (`.agents/skills/vestive-design/assets/shadcn/components/ui/badge.tsx`) with the documented translations.
- **D-8 Group-owned definitions are content:** Organization owners/admins (override on or off) never read or list them;
  the Group page shows `labels-not-administered` to them. This is the Goal/ADR-008 reading; confirm, since G060 Spec R1
  names "protected object labels" as content.
- **D-9 Separate adapter and route:** `core/internal/labels/api` with its own six paths instead of widening the Group DTO
  (`GROUP_KEYS`, byte budgets, every Group list consumer) — keeps every existing DTO and pin unchanged; `profile-paths.tsv`
  classifies `core/internal/labels/` as foundation so a labels-only MR selects full.
- **D-10 Browser projects:** new `labels-light/dark` projects on `labels.spec.ts` (two more `--project=` flags in
  IDENTITY.md and the `lib.sh` matrix row). Alternative: fold E-LAB into `membership.spec.ts` (no new project; a larger
  shared spec).
- **D-11 Observations outside scope (no action in G020 unless Codex adds them):** `docs/foundation/MIGRATIONS.md` has no
  row 16; `logrecord/completion.schema.json` core route enums were not seen to include `/internal/v1/organization/group-authority`
  (G060) — verify against the G060 full-run outcome before adding the labels routes to the same enums.
- **D-12 No split; forecast 3 attempts of B5.** Cross-layer but finite; the S1 early-risk checkpoint bounds the schema
  risk before the facade/UI work.
- **Risks:** (r1) oracle growth (627 cells → ~730) must add explicit deny rows for every staff/override cell of
  `label.group.*`; (r2) the assignments SELECT policy's EXISTS over `label_definitions` must stay acyclic and
  constant-cost (one index probe; E-PG-style plan summary in CT-LAB); (r3) version/count pin churn (17 → NN, 71 → 80,
  46 → 51, 19 → 22) is mechanical under standing authority; (r4) the G060 full run or attempts 6–7 may change 00017
  bodies, counts or line numbers — §13; (r5) G055 cutover may move the handoff/receipt tooling mid-task — §13; (r6)
  hosted DinD/base-image gotchas (memory); (r7) if G061 is scheduled before G020 execution, batch scheduling (pool of
  two, sequential builds) is unchanged but the coordinator's slot inventory must include its runs; (r8, r1) the
  test-only fixture widening must be restored exactly on teardown and must never share a database with
  `TestLabelsCatalog` — the byte comparison in T5(c) and the migrate package's own database pin this; (r9, r1) the
  Organization-scope create's exclusive Organization lock serializes with policy/enrollment writes of the same
  Organization — bounded administrator traffic, no retry logic added. Stops: any path by
  which a Group role reveals an Organization-owned label or Organization authority reveals a Group-owned one; any
  production reachability of the fixture adapter; a native refusal; budget exhaustion.

## 11. Budget, pins and scope allow-list

Budget **B5, checkpoint after attempt 3**, zero execution attempts spent; planning consumed none. Pins to re-pin at
release against the integrated base: migrations 17 → NN, `current_version` NN, relations 71 → 80, capabilities 19 → 22,
oracle cells 627 → actual, access operations 45 → 52 (`log.go`), audit actions 46 → 51 (trigger 39 → 44), reasons
70 → 72, operations 65 → 72, resource types 11 → 12, policy reader 9 / claim writer 7 (unchanged), tenancy production
files 19 → 20, core routes +7 pairs / +6 paths, `required-checks.tsv` +3, browser matrix +1 row (E-LAB-1..6),
`--project=` +2.

Proposed exact file allow-list (extended regular expressions; Codex imports as
`management/v0/results/G020-scope-allowlist.txt`):

```
^core/migrations/000[0-9][0-9]_labels[.]sql$
^core/migrations/migrations_test[.]go$
^core/internal/tenancy/(labels|labels_test|errors|boundary_test|integration_test)[.]go$
^core/internal/tenancy/access/labels(_targets|_test|_integration_test|_fixture_test)?[.]go$
^core/internal/tenancy/access/(policy|policy_test|matrix_test|log|audit|errors|boundary_test|rls_expected_test|rls_integration_test|rls_ordering_integration_test|attack_expected_test|attack_integration_test|audit_integration_test|integration_test|recovery_integration_test|mcp_test|mcp_integration_test)[.]go$
^core/internal/audit/(audit|audit_test)[.]go$
^core/internal/migrate/(labels_integration_test|runner_test|integration_test|audit_integration_test)[.]go$
^core/internal/labels/api/http(_test)?[.]go$
^core/internal/mcp/(mcp_test|catalog_fixture_test|mcp_integration_test|env_integration_test)[.]go$
^core/cmd/server/main[.]go$
^web/lib/labels-client[.]ts$
^web/components/labels/(label-list|label-form|group-labels-section)[.]tsx$
^web/app/organization/labels/(page[.]tsx|action/route[.]ts)$
^web/app/groups/\[id\]/labels/(page[.]tsx|action/route[.]ts)$
^web/app/groups/\[id\]/page[.]tsx$
^web/app/organization/members/page[.]tsx$
^web/app/account/page[.]tsx$
^web/playwright[.]config[.]ts$
^web/tests/unit/(labels-client|labels-forms)[.]test[.]mjs$
^web/tests/browser/labels[.]spec[.]ts$
^infra/validation/(run|lib)[.]sh$
^infra/validation/required-checks[.]tsv$
^infra/ci/profile-paths[.]tsv$
^infra/validation/tools/internal/logrecord/(events[.]json|logrecord_test[.]go|completion[.]schema[.]json)$
^docs/(ARCHITECTURE|SECURITY)[.]md$
^docs/foundation/(AUTHORIZATION|DATABASE-ISOLATION|TENANCY|AUDIT|MIGRATIONS|MCP|WEB-SHELL|IDENTITY)[.]md$
^contracts/mcp/control-plane[.]md$
^management/v0/results/G020-(scope-allowlist[.]txt|generic-labels-visibility-result[.]md)$
```

**28 patterns** (every named existing file verified present at `c1e724c6`; `contracts/mcp/control-plane.md` only for D-1's
production-tool alternative, otherwise untouched). Excluded on purpose: `core/migrations/0000[1-9]*`, `00010`–`00017`
(accepted history), `web/package.json` (no dependency), `web/components/ui/**` and `web/components/membership/**` (no
edit), `infra/ci/**` other than `profile-paths.tsv`, `.gitlab-ci.yml`, seeds/fixtures (`infra/identity/**`: tests create
their own rows and restore them), `core/internal/onboarding/**` and `core/internal/membership/**` (DTOs unchanged, D-9),
`.agents/skills/**`.

## 12. Revision history and transfer state

- r0, 2026-10-20 (UTC): initial provisional Plan from two bounded read-only audits (core authority/RLS/audit/HTTP/MCP/tests;
  web/harness/docs/G055) and the governance inputs in `G020-planning-inputs.json`; load-bearing lines spot-checked by the
  planner (`policy.go:84-311`, `facade.go:1-160`, `invitations.go:310-367`, `objects_policy.go:40-130`,
  `authority.go:70-97`, `00017:130-194,260-295,612-667,690-735,828-866,1000-1044`, `00015:130-210`,
  `audit.go:80-98,328-383`, `membership/api/http.go:35-74`, `run.sh:168-196`, `required-checks.tsv:1-20`,
  `profile-paths.tsv:18-47`, `playwright.config.ts:88-120`, `names.go:8-27`). Native guard refusal during planning,
  quoted once and not retried: a monitor loop was refused as "this command evaluates a arithmetically inside a construct
  too complex to verify … Refusing to run it — a worktree-isolated agent's git operations must target its own
  worktree"; plain separate commands were used instead. No installs, builds, tests, Docker, DB/browser probes, settings
  or Goal/Spec/review/roadmap edits.
- Transfer state (for a successor planner, if any): the open items are exactly D-1…D-12; the audits' numbers are in §11;
  the base is `c1e724c6` with `d0b06691` noted (§0).
- r1, 2026-10-20 (UTC): one bounded correction pass on the R0 review (`G020-plan-codex-review.md` at `c28671ed`, SHA-256
  `e0b6ff31…b00309`), by a successor planner after the r0 planner was stopped above the 200k context threshold. Delta
  reads only: the review, the Goal/Spec diff `c1e724c6..c28671ed` (two context-policy amendments and the Spec "MCP
  foundation clarification"; no other G020 Goal/Spec change), `git log d0b06691..origin/staging` (three documentation
  commits; nothing under `core/`, `web/`, `infra/`, `contracts/`, `docs/foundation/`), and the source lines cited in
  §1.3/§1.4 (`facade.go:105-144,196-215`, `facts.go:70-96`, `policy.go:276-290`, `context.go:89`, AUTHORIZATION.md
  L408-430, lock-strength grep over `tenancy`/`access`). No installs, builds, tests, Docker, DB/browser probes, migration
  reservation, settings or Goal/Spec/review/roadmap edits; the r0 inventory is preserved (inputs JSON `inputs_at_c1e724c6`
  untouched; r1 bindings added beside it). Native refusal, quoted once and not routed around: `EnterWorktree` of the r0
  worktree answered "Entering …/agent-<id> would not give this agent write access there, so the switch was
  not made. This session is isolated in the worktree …/0-agent-5c6f436, so its writes are limited to that folder."
  The r1 edits were therefore made in the isolated worktree on local branch `g020-plan-r1` created at `40792883`
  (= `origin/g020-plan`, verified) and pushed to `origin/g020-plan`; the r0 worktree's local `g020-plan` is left at
  `40792883` (behind origin by the r1 commit) for the coordinator to fast-forward or remove at cleanup.
- Transfer state after r1: open for Codex = the finding-4 reconciliation (D-2 closed design, §1.4/§10) and the five
  finding dispositions in §14; D-1/D-4 decided; D-3, D-5…D-10 accepted in principle; D-11 observations; D-12 unchanged.
  Budget B5 / checkpoint 3, zero attempts spent. Next safe boundary: Codex r1 review.
- r2, 2026-10-20 (UTC): one bounded correction pass on the R1 review (`G020-plan-codex-review.md` at `origin/staging`
  `3b4b1dc2`, SHA-256 `de984df8…3b9b5ed`, section "R1 — bounded correction remains"), by a successor planner in the
  isolated worktree `agent-<id>` on local branch `g020-plan-r2` created at `origin/g020-plan` `c8ab9e4a`
  (HEAD and clean tree verified before editing). Delta reads only: the R1 section, this Plan's §1.2/§1.4/§3/T3/T5/§8/
  §10/§12/§14, the inputs JSON, `git log c8ab9e4a..origin/staging` (two documentation commits) and an empty
  `git diff --stat c28671ed 3b4b1dc2` over `core/`, `web/`, `infra/`, `contracts/`, `docs/foundation/` and the G020
  Goal/Spec; source facts: `00007:133` (stored generated column precedent), `00008:606-618`/`00015:208-209` (audit
  captures are AFTER ROW triggers), `run.sh:4105` (`pgvector/pgvector:0.8.7-pg17`), `infra/identity/mcp-seed.sql:14-29,63-64`
  (fixture roster: bob owns research, Dana owns design, Grace A admin + research admin, dave research member; no seeded
  subject manages both research and design, hence the per-run `grantGroup` binding in T5 (d)). Changes: §1.4 contract
  item (3) replaced by the NULL-skipping generated-column key and (4) tightened; §1.4 exact test-only storage paragraph;
  §3 fixture SELECT disjunct tightened; T3 substitution oracle rewritten (isolated 23503 cases, pinned first refusals
  42501/23514, zero-row/zero-audit verification); T5 (d)–(g); AC4/AC5 rows; D-2 r2 note; §13 item 4; §14 r2 table.
  No installs, builds, tests, Docker, DB/browser probes, migration reservation, settings or Goal/Spec/review/roadmap
  edits; r0/r1 text preserved (one r1 sentence struck through, not deleted). Native refusal, quoted once and not routed
  around: the first `git commit` (a `$'…'` trailer argument) was refused — "this command names git in a form too complex
  to verify that it stays inside the worktree. Refusing to run it"; the commit was re-issued as a plain `git commit -F`
  with the message in a file. All other git operations were plain single commands in this worktree.
- Transfer state after r2: open for Codex = R1 items 1 and 2 (§14 r2 table) and q3 (§14). Findings 1, 2, 3, 5 and D-1/
  D-4 settled by R1; closed production target set accepted by R1; D-3, D-5…D-10 accepted in principle; D-11
  observations; D-12 unchanged. Budget B5 / checkpoint 3, zero attempts spent. Next safe boundary: Codex r2 review.

## 13. Reconcile before execution release (provisional dependencies)

1. **G060 final ACCEPT.** Rebind to the accepted source: re-read `00017` (policy bodies, postcondition lists at
   `:1026-1028`/`:1039-1040`, inventories 9/7), `policy.go`/`matrix_test.go` (627 cells), `audit.go`/`audit_test.go`
   (46/70/65/11), `migrate/integration_test.go` (71, version 17), the four `current_version 17` pins, `run.sh` version/
   count literals, `events.json`, `completion.schema.json`; any G060 correction attempt (cap 7, 5 spent) that touches
   these moves the baseline. Verify the next free migration number again.
2. **G055 final ACCEPT and cutover.** If `skills/restruct-sdlc/` is live, the writer and coordinator use its
   entrypoint/discovery/handoff/receipt tooling and the Result records that use (AC8); if not, record the legacy layout
   used. No G055 requalification by G020; a tooling defect is a G055 repair.
3. **Staging deltas after `c1e724c6`**: `d0b06691` (Vestive canonical path, `run.sh` line shifts ≥ L847, `profile-paths.tsv`
   rows 47-48) and anything later; re-pin every `run.sh`/`lib.sh` line in §1.1/§5.4 and the Vestive path; re-read the
   G057 Spec refresh (`089b425d`) only for conflicts (none expected: live qualification is not a G020 dependency).
4. **Open decisions** (r1): D-1 and D-4 decided by the R0 review (no change to §3.6/§9 beyond the r1 notes; allow-list
   unchanged); D-2's closed design awaits the finding-4 disposition (r2: the closed production set was accepted by R1;
   the fixture-contract correction of R1 items 1–2 awaits the r2 review); D-3, D-5…D-10 accepted in principle (actual
   source/policy still reviewed at implementation); D-11 observations verified against the actual G060 outcome.
5. **Resource readiness**: pinned pgvector/MinIO/web-check images available by verified ID; private-run pool slots
   (two) and one build slot inventoried by the coordinator; G061 scheduling (undecided) does not change this Plan.
6. **Input rebinding**: refresh `G020-planning-inputs.json` hashes at the release commit; Goal/Spec bytes expected
   unchanged from the r1 bindings (`f61dbc39…`, `798017e9…` at `c28671ed`).
7. **Context policy** (r1): the writer runs under 150k checkpoint / 200k handoff with supported self-compaction (§0, §6);
   no restart or added review for the carry-forward.

## 14. r1 correction record — finding → change map

| R0 item | Disposition | Change (section) |
|---|---|---|
| 1 Referential integrity | **Fixed.** Nullable `owner_group_id` replaced by `NOT NULL owner_scope_id` on assignments; composite FK `(organization_id, label_id, owner_scope_type, owner_scope_id)` → definitions' `(organization_id, id, owner_scope_type, owner_scope_id)`, every referencing column NOT NULL so MATCH SIMPLE is total in both scopes; definitions gain explicit `organizations`/`groups` FKs with `group_id NOT NULL` under the Group-scope CHECK; RLS retained. T3 raw-SQL substitutions (unknown/foreign/owner-mismatch, deletion RESTRICT) with paired positives, in the existing `TestRLSIsolation` fixtures | §1.2, §3 step 6, T3, AC5 |
| 2 Creation locking | **Fixed.** Organization-scope create uses the existing `policyWriteQuery` (Organization `FOR NO KEY UPDATE` up front, step 3); Group-scope create `groupQuery(g, write)` (Group `FOR NO KEY UPDATE`, step 6); the bound trigger re-locks at the same strength (raw-SQL backstop), no SHARE→NO KEY UPDATE conversion; O-LAB-4 proves count/conflict with no 40P01 | §1.3, §4.1, T6, D-risk r9 |
| 3 Selected-context deletion | **Fixed.** Organization-owned definition deletion requires Organization-wide context (`tc.group` zero): otherwise `invalid_input` `organization_context_required` before any SQL, uniform and non-enumerating; context never cleared, RLS not bypassed; `assignments_removed` complete by construction; Group-owned deletion rule stated; LAB-17 two-Group positive/negative; web status string | §1.3, §1.5, §4.1, T4, AC5 |
| 4 Target-read extension proof | **Reconciled; D-2 proposed closed, for Codex disposition.** Production SQL guarantee stated exactly (total FK + `target_type = 'group'` CHECK + Group FK + readers are `MGR`, who hold `group.read` on every Group, `policy.go:280-282`) versus the test-only adapter + schema widening; registration contract (adapter, CHECK, FK, SELECT predicate) documented; fixture direct-SQL observations for an unreadable and a nonexistent target with core/SQL parity; production bodies pinned `group`-only by CT-LAB-2; residual limitation explicit | §1.2, §1.4, §3 policies, T5, AC4, D-2 |
| 5 Dual-role oracle | **Fixed.** Grace reads Organization-owned rows through her A-admin authority and research's Group-owned rows through her research role, nothing of design; Group-role-only negatives (bob/dave/Dana) and override-without-Group-role negatives (frank OVR on, carol OVR off) retained unweakened | T3 |
| D1 | Decided: five production tools kept; fixture evidence labelled proof-only | §4.3, §10 D-1 |
| D2 | Held → closed production SQL proposed under finding 4 | §1.2, §1.4, §10 D-2 |
| D3, D5–D10 | Accepted in principle; no text change beyond cross-references | — |
| D4 | Decided: one `label` resource, actual Group attribution, explicit future-target mapping deferred to the registering Goal, no new column | §10 D-4 |
| D11 | Observations retained; verify against the actual G060 outcome at release | §13 item 1 |
| D12 | Unchanged: B5 / checkpoint 3, zero attempts, forecast is not permission | §11 |
| Trigger tally | Five (2 immutable + bound + 2 capture), replacing r0's "four" | §1.2, §3 step 6 |
| Context policy | 150k/200k and supported self-compaction reconciled; no restart | §0, §6, §13 item 7 |

No new owner product choice is requested by r1. Questions for Codex: (q1) accept the closed D-2 design as the finding-4
reconciliation, or require the test-only widening to be narrowed to core-only proof; (q2) `organization_context_required`
as `invalid_input` (context shape, decided before SQL) rather than `forbidden` — chosen because the subject's authority is
not in question, only the context shape.

### r2 correction record — R1 item → change map

R1 (at `origin/staging` `3b4b1dc2`) answered q1 (closed production set accepted; fixture widening retained as the
template proof) and q2 (`invalid_input` consistent with the existing contract), and left two items.

| R1 item | Disposition | Change (section) |
|---|---|---|
| 1 Fixture target FK not isolated by type; no owner binding | **Fixed.** The r1 two-column FK `(organization_id, target_id) → label_fixture_targets` is withdrawn. Test-only storage: `label_fixture_targets` with `UNIQUE (organization_id, id, group_id)` and a Group FK; on `label_assignments` a stored generated column `target_fixture_id = CASE WHEN target_type = 'fixture' THEN target_id END` and `label_assignments_fixture_target_fkey (organization_id, target_fixture_id, target_group_id) → label_fixture_targets (organization_id, id, group_id)` ON UPDATE/DELETE RESTRICT. MATCH SIMPLE skips production `group` rows (key column NULL), so they coexist without fake fixture rows or any change to the production CHECK, which stays closed to `group`; on fixture rows the FK binds existence and the target's actual `group_id` to `target_group_id` = `owner_scope_id` = the definition's owning Group. The registration contract (3) and (4) and the fixture SELECT disjunct carry the same binding. Direct-SQL negatives: wrong-Group placement by a dual-Group subject → 23503 isolated, `target_group_id` variants → 23514/42501 first, superuser move of an assigned target → 23503; coexistence positive for production rows; teardown parity incl. the column list | §1.4 (3)–(4) and the r2 storage paragraph, §3 policies, T5 (d)(e)(g), AC4, D-2, §13 item 4 |
| 2 T3 over-pins 23503 for owner-fact substitutions | **Fixed.** Refusal order stated (RLS 42501 → CHECK 23514 → unique → FK 23503). The isolating case is a Group-owned definition id supplied with Organization owner facts in a valid production shape by carol → 23503 (only the composite FK can refuse). The converse (Organization-owned id with Group facts) asserts the actual first refusal — carol 42501, bob 23514 — and its isolated FK case is T5 (f) under the widening (bob, research fixture target → 23503). Every refused statement runs under a savepoint and is followed by a superuser count: zero rows, zero `audit.events` (`label.assigned`). The multi-check pins may be corrected to observed codes under standing authority; the isolated 23503s and the zero-row/zero-audit assertions may not | T3, T5 (f), AC5 |
| Decided by R1 | Findings 1, 2, 3, 5 fixed; D-1 and D-4 decided; production target set closed to `group` accepted; q1/q2 answered; no text change beyond the header/§13 notes | header, §13 item 4 |
| Budget | Unchanged: B5 / checkpoint 3, zero attempts; r2 spends no implementation attempt | §11 |

Question for Codex (q3, test-fixture detail only, no product choice): the dual-Group subject for T5 (d) is a per-run
`grantGroup` binding (Dana + research `group_admin`) rather than a seed change; if Codex prefers a seeded dual-Group
subject, that is a fixture-roster edit outside this Plan's allow-list and would be scheduled separately.
