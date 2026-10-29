> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G005 — Organization and Group membership model: implementation plan (r4, provisional)

- **Plan revision:** r4 DRAFT, 2026-10-08 (2026-10-09 UTC). A bounded, owner-authorized amendment (via Codex; planning only) that reconciles the unreviewed r3 to the **accepted** G004 implementation and the current CI controls: the items of the r4 change map (§0). No redesign: r3's model, contracts, tests, budget and sequence are reused except where §0 says otherwise.
  - Author: Claude Code (planning-only subagent; no delegation). Reviewer: Codex. Major decisions: product owner.
- **History (preserved, linked):**
  - Plan r1 (not included): SHA-256 `e01cccf155af3cdd4cbbb761fcb486c1f715396ad583a4e2c10f92bcbf8c2244`.
  - r1 Codex review (not included): SHA-256 `d3aedd867751ad5dc1a30b8b74e1cb634e2a3df4e700fbd7fb260bf4f83a7568`. CORRECTION LOOP; Q1–Q10 resolved.
  - Plan r2 (not included): SHA-256 `eb7e2c749f4dec0beff376abd2885158ff91a237ccbbec4db916a509d304cb5d`. It was this path's content at the baseline; its exact bytes were archived with `cp` and hash-verified before r4 was written.
  - r2 Codex review (not included): SHA-256 `b1f7a2c9da121eb0433506f1f4916f48cacc5a73511882126777a839252c4c30`. Provisional design **ACCEPT**, including two CT-6 catalog clarifications (§5.2); execution still gated.
  - Plan r3 (not included): SHA-256 `7607b6f3f58bc91c75b0f4e28c97ac5e6c21cbdcd7cf1a52fc6dd50055545c8d`. Written in the isolated clone at root staging `19aa3c2` and mapped provisionally to G004 candidate `4a46da9`; imported byte-exactly with `cp` and hash-verified. **Never reviewed** (no r3 review exists); r4 supersedes it as the amendment submitted for review.
- **Status: PROVISIONAL PLAN DRAFT — submitted for Codex amendment review; NOT implementation-ready.**
  - **Provisional baseline.** r4 is reconciled at `150544058d152fd72f87955477b84ef374b16d46` (tree `ef05f4110a65a7266ef97633bc669d0ace6c3f59`), the G004 MR !7 reconciliation merge. Parent 1 is the accepted G004 delivery `7dc4c5cc28fb05d79daa52be3903599c09be34b8`; parent 2 is staging `c1412355c841627920acc7912f25c2fdaf6f3b5b`. **MR !7 is open, not merged**, and its MR pipeline was still running at handoff, so `1505440` is not a `staging` commit. The execution binding (E3, E5) must be refreshed against the actual integrated G004 staging commit before execution approval.
  - **G004 is implementation-accepted** (Codex ACCEPT, AC1–AC12) but **not yet integrated**. Observations marked **[G004-ACC]** are read from the accepted code, byte-identical in `1505440` (§2.1). They replace r3's provisional [G004-OBS] observations at `4a46da9`, which differs from `7dc4c5c` only in the G004 Result.
  - Gates E2 (its integration part), E3–E6 and S0 (§2.3) are unsatisfied.
  - Nothing was implemented, installed, built, migrated, run in Docker, committed or pushed. No provider, GitLab, CI or settings action occurred. Only the four planning files listed in §14 were written. No other checkout was read or touched; the G049 Plan drafted in parallel was not read.
  - **Zero of seven B7 cycles consumed.**
- **Planning provenance:**
  - Checkout `<source-checkout>`, detached at `1505440`, clean before writing.
  - r4 inventory `G005-planning-inputs-r4.json` (not included): SHA-256 at `1505440` of the repository inputs read, the r3 blob comparison, and the two external r3 files with absolute paths and hashes.
  - r3 inventory (external): `<home>/.local/state/platform-evidence/g005-planning-19aa3c2/g005-planning-inputs-r3.json`, SHA-256 `1fb6663250928639a4ac91f5ba1a2253de8f789f4537510aad476d8486e41d56`. The r1/r2 inventories stay with the historical `g005-planning-42b21d1` clone.
  - This is planning provenance, not an execution binding.

| Governing input (r4, at `1505440`) | SHA-256 |
|---|---|
| G005 Goal r2 / Spec r2 (unchanged since r3) | `9ed47e7a…c892` / `9be24ef8…ae9f` |
| G005 Goal r1 / Spec r1 archives | `6a5f978b…a411` / `e176531f…428f` |
| ADR-004 / cross-Goal addendum `G006-business-profile-cardinality-addendum.md` | `51a48784…5a68` / `495a1509…0810` |
| `docs/ARCHITECTURE.md` (§53; "Consultant engagement functions — 2026-10-08") | `c0357d1c…1dfe` |
| `docs/DECISIONS.md`; ADR-003 | `4ac654da…53e0`; `537522dc…2a62` |
| `docs/SECURITY.md` | `e3b0c442…b855` — still the empty file (Q9 unchanged) |
| AGENTS.md / CLAUDE.md / BUDGET_POLICY | `b89d61ac…afaa` / `37467aca…3137` / `b183179b…f437` |
| ROADMAP / PROGRESS / BACKLOG | `2c787298…52b0` / `256eba7c…fa66` (changed since r3: G004 acceptance and the G005 delegation record) / `700d6e50…f007` |
| G004 Result / implementation review / integration reconciliation | `45210c3a…a4a9` (equal to the review's bound Result digest) / `7bfabe24…5190` / `65928e48…c69a` |
| G049 Goal / Work Spec (draft r2 plus owner clarifications) | `75166d9b…93e7` / `66a7f4c7…d5ac` |
| Accepted G004 code read ([G004-ACC]) | `00002_identity.sql` `8d44b300…6e14`; `workos/staging_test.go` `3240d0fe…612c`; the rest in the inventory |
| CI controls read | `.gitlab-ci.yml` `72958be8…eeec` and `run.sh` `623a5570…7cd1` (both equal the reconciliation record's resolved digests); `profile-paths.tsv` `b8189994…70b3`; `select-profile.sh` `763e7b43…438d`; `validate-controls.sh` `17f65d23…bf17`; `required-checks.tsv` `79334c76…458c` |

---

## 0. Revision record

| Rev | Date | Outcome |
|---|---|---|
| r1 | 2026-10-07 | Initial provisional Plan, drafted while G004 is unimplemented. Concrete schema, contracts, tests, harness changes, gates and forecast. Every G004-dependent decision is listed for reconciliation (§2.2). **CORRECTION LOOP**: two bounded corrections; model direction accepted; Q1–Q10 resolved. |
| r2 | 2026-10-07 | The two corrections below, and the review's decisions incorporated without redesign. Provisional design **ACCEPT** (r2 review (not included)). |
| **r3** | 2026-10-08 | Owner-authorized bounded amendment under ADR-004 and G005 Goal/Spec r2: the six reconciliation items of the r3 change map below; no redesign. One escalation (ESC-1, §13). |
| **r4** | 2026-10-08 (2026-10-09 UTC) | Owner-authorized bounded reconciliation of the unreviewed r3 to the accepted G004 implementation (`7dc4c5c`, observed in the provisional MR !7 merge `1505440`) and the current CI controls: the r4 change map below; no redesign. ESC-1 kept with re-assessed options; no new escalation; three questions (§13.2). |

**r2 correction map (r1 review).** The model, schema, DDL, contracts, error mapping, budget and dependency register are unchanged.

| Correction | Resolution | Changed sections |
|---|---|---|
| 1 — CT-6 wrongly required `role_bindings.role` to be the only `role` column in `app`, although both projection views expose `role` by design. | CT-6 becomes a relation-kind inventory with three positive assertions: exactly one durable role column; exactly the two required SELECT-only projections, each deriving `role` only from `role_bindings`; no write path to any role-bearing column except `role_bindings.role`. Rolled-back negative controls prove that an extra stored role column, a materialized or writable mirror, a third role-bearing view, a writable or trigger/rule-backed projection, and a projection sourced elsewhere each fail. View write denial (IT-PROJ-2) remains. | §5.2 CT-6 and CT-6-NEG; §8 AC4 |
| 2 — r1's S2/S3 gates needed tenancy IDs whose `run.sh`/registry wiring was scheduled only at S4. Adding `00003` would also have broken the inherited migration-version checks before S4. | Wiring is minimal and step-local. Each step adds a required ID's recording, its registry row and the inherited expectations its files change, then runs the covering suites through the existing wrapper. `TestTenancyCatalog` moves to the `migrate` package, so S2 does not depend on the S3 package. Registry completeness (B-REQUIRED-CHECKS, per executed suite) holds at every step. | §5.1, §6.1, §6.2, §6.4, §7, §10 |

**Review decisions incorporated:**
- Q1 wording: DS-7, §7.
- Q4 re-inventory: P-G004-11.
- Q6: E6, §5.4, §8 AC11 and §12 no longer include X3; one passing hosted full run on the final tree remains mandatory.
- Production versus test-only import scope: §5.1.
- §13 now records all resolutions.

**r3 change map.** The r2 model is otherwise reused unchanged: UUID keys, typed-scope CHECK, composite FK, single RoleBinding store, `security_invoker` projections, CT-6 detector, column grants plus triggers, slug contract, error mapping, B7 forecast and CI placement.

| Item | Resolution | Changed sections |
|---|---|---|
| 1 — Atomic Group creation with the explicit creator's `group_owner` binding | `CreateGroup` takes an explicit existing creator UUID. One transaction inserts the Group and that one `group_owner` membership binding; either failure rolls back both. There is no owner column and no Organization membership is minted. Owner bindings remain ordinary RoleBindings for the internal primitives; transfer and last-owner policy stay with G019. CT-6 now also rejects stored owner/creator columns. | DS-9; §4.2, §4.5, §4.6, §4.7; §5.2 IT-OWN-1..6, CT-6 A1/N9; §8 AC3/AC7 |
| 2 — Organization and Group owner/admin/member role codes | Organization `org_owner`, `org_admin`, `org_member`; Group `group_owner`, `group_admin`, `group_member`; source `membership`. Applied to the CHECK, Go types, projections and affected tests: six valid combinations. | §3.2 DDL; §4.2/§4.3; §5.2 IT-VOC-1, IT-PROJ-1, IT-MULTI-1, CT-4; U-PARITY |
| 3 — Canonical BusinessProfile/GroupBusinessProfile terminology; derived access | BusinessProfile is **not** a RoleBinding scope. G006 derives profile access from current Group/Organization bindings and the single attachment, and adds no scope, source or column to `role_bindings`. r2's G006-extension wording is withdrawn. G005 creates no BusinessProfile schema. The sibling-navigation UX is G018 context, not G005 scope. | DS-2, DS-3; §1; §4.7; §5.2 CT-1, IT-VOC-1; §7 docs; §8 AC8 |
| 4 — Advisor/Principal/Associate are future staffing functions | They are not RoleBinding roles. `advisor`, `principal` and `associate` are explicit negatives in the role matrix. No staffing schema. | §4.3; §5.2 IT-VOC-1 |
| 5 — G004 reconciliation by read-only inspection of candidate `4a46da9` | Every P-G004 item now records its [G004-OBS] observation. Confirmed: `app.users(id uuid)` key, grants, `00002` numbering, runtime-role derivation pattern, routes, counts and events. Concrete mechanical consequences are added to the inventory: migrate `reset`/helpers/`runner_test.go`, identity `TRUNCATE` fixtures, `TestIdentityCatalog` and B-IDENTITY-CATALOG whole-schema sets, B-MIGRATE-CLI object counts. **ESC-1:** G004's credential-bearing staging test also truncates `users` and can only be revalidated by the protected identity-staging job. Final reconciliation against the accepted integrated tree remains gate E3. | §2.1, §2.2, §2.3, §3.2 (G004 grant pattern), §6.2, §7, §10, §12, §13 |
| 6 — Current CI path/profile mapping and accepted controls | Recorded current facts: Markdown under `docs/`, `management/` and root is `docs` (fast) and never promotes a mixed change; `core/` is runtime; migrations are foundation; `branch-web`/`branch-api` select full. The G005 delivery still selects full. The one `core/internal/tenancy/` foundation row plus F-tenancy stays. No harness/control framework, no reduction in required tests; the proportional-testing backlog and G049 stay separate. No manual-test or operator-feedback stage exists. | §2.1, §6.3, §6.4, §12 |

The r2 review's two CT-6 catalog clarifications (A2 source dependency, A3 PUBLIC handling) are now stated in §5.2. Editorial: §10 drops a stray empty table header left in r2.

**r4 change map.** Items 1–4 of the r4 handoff were already in r3. They are re-verified against Goal/Spec r2, ADR-004, the addendum and ARCHITECTURE, whose bytes are unchanged since r3. Everything not listed is reused from r3 unchanged: model, DDL, contracts, error mapping, tests, budget, file set and scope allow-list.

| Item | Resolution | Changed sections |
|---|---|---|
| 1 — Atomic Group plus creator `group_owner` | Unchanged from r3 (DS-9; IT-OWN-1..6; CT-6 A1/N9); matches Spec r2 R5/AC7 and the addendum. | — |
| 2 — Six membership role codes | Unchanged from r3: `org_owner`/`org_admin`/`org_member`, `group_owner`/`group_admin`/`group_member`, source `membership` (Spec r2 R2). | — |
| 3 — BusinessProfile/GroupBusinessProfile; derived authority | Unchanged from r3: G005 adds no profile scope, role, grant, column or schema. G006 derives access read-only from current bindings and the single attachment; sibling management UX is G018's. | — |
| 4 — Advisor/Principal/Associate | Unchanged from r3: future engagement functions (ARCHITECTURE "Consultant engagement functions — 2026-10-08"), rejected as RoleBinding roles (IT-VOC-1). | — |
| 5 — Reconciliation to the accepted G004 code | The observation basis moves from candidate `4a46da9` [G004-OBS] to the accepted delivery `7dc4c5c` [G004-ACC]. Verified that they differ only in the G004 Result, and that 12 of the 15 blobs r3 observed are byte-identical at `1505440` (the other three: the Result, plus `run.sh` and `validate-controls.sh`, changed only by MR !7's CI hunks). Re-checked at `1505440` with line numbers: `00002` (schema, grants, preflight/postcondition), both owner `TRUNCATE` sites, the whole-schema sets and every `run.sh` version/count site. No **Differs** finding; no consequence beyond r3's (a)–(d). E2/E3 re-scoped to the open MR !7. | Header; §2.1–§2.3; §6.1; §6.2 site map; §7; §13; `[G004-OBS]` → `[G004-ACC]` in §1 and §3–§12 |
| 6 — Current CI controls and the expected profile | Recorded at `1505440`: staging's documentation bypass, the `profile-paths.tsv`/`select-profile.sh` fast/affected (integration)/full mapping and G004's protected `identity-staging` admission, as combined by MR !7. G005's changed paths select **full** in every gating context (reason `foundation`); `identity-staging` never runs for them. No CI-control change beyond r3's one map row and one control. | §2.1; §6.3 (expected-profile table); §12 |
| 7 — ESC-1: the staging fixture truncates `app.users` | Kept. Re-assessed against the accepted code, the protected-job admission and G004's consolidated-pipeline precedent. Options: A (recommended: one protected qualification pipeline that also supplies X2's full evidence), A′ (local staging suite with the owner's credential file), B (inspection plus the byte-identical statement validated locally; deferral with recorded residual risk), C (a G004 follow-up). Records the local CAP-3 block's effect on the root database. No hosted or provider action is authorized. | §5.4; §6.4 (feedback note); §8 AC11; §12 (X1, X4, X5); §13.1 |
| 8 — G049 overlaps | Concrete overlaps only: three shared files, two read-only dependencies and documentation paths. No blanket G049 dependency; a sequencing and rebase rule is added. | §2.4 (new); §7; §11 stops; §13.2 |
| Precision (r4) | S7's `W(all)` executes the 81 full-profile IDs, not all 83. `W(unit)` on the final tree adds B-UNIT-CHECKS; B-IDENTITY-STAGING follows ESC-1. X1 pushes with `ci.skip` (G004 precedent), so the evidence push starts no automatic pipeline. | §6.1; §6.4; §10; §12 |

B7 with the checkpoint after cycle 4, the AC mapping and the scope allow-list are unchanged.

## 1. Approach summary

1. **One forward migration.** `core/migrations/00003_tenancy.sql` [G004-ACC numbering; re-confirmed at E3] adds:
   - `app.organizations`, `app.groups` and `app.role_bindings`;
   - two read-only `security_invoker` projection views;
   - three key-immutability triggers;
   - exact column-level runtime grants.

   It adds no RLS, policy, BusinessProfile/GroupBusinessProfile, onboarding, routing/domain, branding, audit, staffing or product object.
2. **Declarative typed-scope integrity.** `role_bindings` keeps the canonical `(subject, scope_type, scope_id, role, source)` columns and `organization_id`. It also has a typed nullable `group_id` with a composite same-Organization foreign key. One closed CHECK ties scope, shape, role and source together, and has no unconstrained branch. Roles are Organization and Group owner/admin/member (r3). The subject is a foreign key to the global user table `app.users(id)` [G004-ACC].
3. **One writable authority.** `role_bindings` is the only table with a role. `organization_memberships` and `group_memberships` are views over it, granted SELECT only. Grant, change and remove are each one write to one row. A role change updates the row in place and keeps its UUID. Removal deletes the row; a regrant creates a new row with a new UUID and the role given at regrant time. BusinessProfile access is derived later (G006/G007) from these bindings; no profile role exists.
4. **Internal core primitives only.** `core/internal/tenancy` provides Organization, Group, slug and membership operations for trusted in-process use and tests. Group creation atomically writes the Group and its explicit creator's `group_owner` binding (r3).
   - It is not linked into the server or migrate binaries.
   - There is no HTTP, BFF, MCP, CLI, environment hook or bootstrap.
   - Every scope-sensitive read and write carries an explicit Organization predicate.
5. **Group slugs.** Stored grammar `[a-z0-9]+(-[a-z0-9]+)*`, 1–63 ASCII bytes, with a pinned 24-entry reserved set (§4.4). Slugs are unique per Organization. Suggestions are deterministic; explicit values must already be canonical. Collisions are decided by the unique constraint.
6. **Evidence.** The evidence layers are:
   - unit tests in the existing `core-check` image;
   - real-PostgreSQL model, catalog, concurrency, fault-injection and migration tests, run as the actual runtime role in the existing disposable `migrations` project;
   - a route-absence probe on the root stack;
   - existing log and privacy tooling;
   - one authorized hosted full pipeline.
7. **Bounded harness/CI reuse.**
   - Four new required IDs.
   - Mechanical reconciliation of existing migration-version and catalog expectations.
   - One profile-path row plus one selector control.
   - One log-catalogue event.
   - No new suite, job, profile, wrapper, framework or disposable project.

## 2. Baseline, provisional dependencies and gates

### 2.1 Observed state (facts, read-only)

**r4 update: provisional baseline `1505440` (facts at that commit; [G004-ACC] for the accepted G004 code).** These supersede the r3 and r1/r2 observations below, which are kept as history.

- **Baseline identity.** `1505440` (tree `ef05f411…`) is the G004 MR !7 reconciliation merge: parent 1 is the accepted delivery `7dc4c5c` (tree `6f24080b…`), parent 2 is staging `c141235`. MR !7 is **open**, so the merge is not on `staging`. Since r3's `19aa3c2`, staging gained only management records (`19aa3c2..c141235`: PROGRESS and the G004 implementation review).
- **G004 is accepted, not integrated** ([implementation review](../results/G004-implementation-codex-review.md): ACCEPT, AC1–AC12; B7 7/7).
  - The hosted-qualified commit `4a46da9` and the delivery `7dc4c5c` differ only in `management/v0/results/G004-global-identity-sessions-result.md` (`git diff --stat 4a46da9 7dc4c5c`).
  - Of the 15 blobs r3 observed at `4a46da9`, 12 are byte-identical at `1505440`: `00002_identity.sql`, `migrations_test.go`, `migrate/{integration,runner}_test.go`, `identity/{integration_test,store_test,identity,http}.go`, `workos/staging_test.go`, `required-checks.tsv`, `events.json` and `infra/ci/lib.sh`.
  - The other three: the G004 Result (now the accepted version, SHA-256 `45210c3a…`), and `run.sh` and `validate-controls.sh`, changed only through the MR !7 merge.
- **What MR !7 changed relative to the accepted delivery** (reconciliation record (not included); `git diff --stat 7dc4c5c 1505440`):
  - No `core/`, `web/`, `intelligence/`, contract, migration, Compose or Docker path.
  - `.gitlab-ci.yml`: staging's documentation-bypass globs combined with G004's `identity-staging` job and protected schedule rule (ten workflow rules).
  - `run.sh`: only `static_ci_config`'s bypass comment, its workflow `want` string and the B-CI-CONFIG PASS text.
  - `validate-controls.sh`: staging's nine `F-doc-*` controls added to G004's version, which already holds the identity admission controls.
  - `profile-paths.tsv` and `select-profile.sh`: staging's versions, byte-identical to r3's `19aa3c2` observation.
  - Otherwise management and governance records, and the `claude-session-handoff` skill from staging.
- **Identity schema** (`00002_identity.sql`, [G004-ACC]):
  - Five tables: `identity_provider_binding`, `users(id uuid PRIMARY KEY DEFAULT gen_random_uuid(), status, display_name, created_at, updated_at, disabled_at)`, `identity_principals`, `sessions` and `login_transactions`. With nine indexes, 14 relations.
  - Up opens with the owner/runtime preflight DO block (`to_regrole(current_database() || '_runtime')`). Grants run in a DO block whose postcondition requires the exact runtime table ACL and no other grantee.
  - Runtime grants: SELECT/INSERT/UPDATE on `users`, `identity_principals` and `sessions`; SELECT/INSERT on the binding; SELECT/INSERT/DELETE on `login_transactions`. No DELETE on users.
  - No identity column matches CT-6's role- or owner-bearing patterns.
- **Owner `TRUNCATE` of `app.users`.** Both sites use the byte-identical statement `TRUNCATE app.login_transactions, app.sessions, app.identity_principals, app.users, app.identity_provider_binding`:
  - `core/internal/identity/integration_test.go:134`, `resetIdentity` (tag `integration`). It is called from `integration_test.go`, `service_test.go`, `logout_test.go` and `store_test.go` (`TestIdentityCatalog`), and exercised by B-IDENTITY-CATALOG (`migrations`) and B-IDENTITY-IT (`ops`), each after `migrate up` to the latest version.
  - `core/internal/identity/workos/staging_test.go:258` (tag `workos_staging`), after `r.Up` on the embedded `migrations.FS`. Only B-IDENTITY-STAGING exercises it (see CI below and ESC-1).
  - No other site truncates, drops or deletes identity tables: `store_test.go:161–167` are runtime-denial negatives, and `manual_qualification_test.go` (`workos_manual`) and `identitysetup` do not truncate.
- **Whole-schema sets that a version-3 schema changes** (exact; P-G004-5/6):
  - `TestIdentityCatalog` "objects and owner" (`store_test.go:48–61`): every `app` table except `goose_db_version` equals the five identity tables.
  - `metadataAndIdentityOnly` (`migrate/integration_test.go:207–240`): any relation that is not a table or an identity index is a stray.
  - `reset` (`migrate/integration_test.go:72–89`): drops the identity tables directly.
  - `run.sh`: `identity_catalog` (2310–2391; `IDENTITY_OBJECTS` = the 14 relations) and the `mig_step` object counts (1612–1648).
  - Unaffected, because they are scoped to the identity tables, `00002` or `core/internal/identity`: `TestIdentityCatalog`'s forbidden-column and grant subtests, `TestIdentityMigrationShape` and B-IDENTITY-GUARDS.
- **Registry** (`required-checks.tsv`, counted with `infra/ci/lib.sh` `CI_SUITES_*` plus `final`): 79 rows; fast 33, integration 70, full 77, staging 5. The register (`render-customizations.tsv`) holds 23 rows and is path-keyed: B-RENDER-PLAN reads only the path column.
- **CI at `1505440`** (item 6; G005's expected profile in §6.3):
  - **Workflow admission.** Automatic push/MR pipelines need a changed path outside root `*.md`, `docs/**/*.md` and `management/**/*.md`. Feature-branch pushes with an open MR are suppressed. Default-branch schedules, the protected `identity-qualification` schedule, and `web` and `api` pipelines are admitted.
  - **Profile selection** (`select-profile.sh` with `profile-paths.tsv`): an exact file row wins, then the Markdown documentation rule (`docs`, fast, never promotes), then the longest directory prefix.
    - `core/`, `web/` and `intelligence/` are runtime (integration).
    - Foundation (full): `core/migrations/`, `core/internal/migrate/`, `core/cmd/migrate/`, `core/internal/{observability,config}/`, `infra/`, `skills/`, `.archetype/`, CI files, locks, and non-Markdown under `docs/` and `management/`.
    - Unknown paths, and deletions or renames outside documentation, select full.
    - `branch-web`, `branch-api`, `branch-schedule` and `mr-merged-result` select full.
  - **G004 `identity-staging` admission.** It runs only on the protected `identity-qualification` branch, for a `web` or `schedule` source, in environment `identity-staging`, with profile `staging` (B-IDENTITY-STAGING plus `final`). The owner starts the `web` pipeline in the GitLab UI; an `api` pipeline is refused by the forwarding gate (G004 Result). Per the G004 Result the branch stays at `4a46da9`.
  - **Local staging paths.** `run.sh … staging` with `IDENTITY_STAGING_ENV_FILE` (the owner's 0600 file outside the repository). IDENTITY.md's `identity:staging-local` (CAP-3) block, which migrates and truncates the **root** Compose database.
- **Routes at `1505440`** (P-G004-11): web `account`, `api`, `auth` and `sign-in`; public `blueprint` and `fonts`; core `/healthz`, `/readyz` and the five `/internal/v1/auth/*` routes. The 24-entry reserved set covers all of them.
- **Governance.** Every r3 governing input is byte-identical except PROGRESS (G004 acceptance; the G005 delegation record). ROADMAP records G049 with "No application changes or implied G005 blocker", and G021 depends on G049.

**r3 observations at root staging `19aa3c2` and G004 candidate `4a46da9` (history; re-verified at `1505440` as stated above):**

- **G048 is accepted and integrated** at `19aa3c2`.
  - Registry: 71 rows (fast 32, integration 64, full 70).
  - VP1 is 32; the render register holds 20 rows.
  - `web/public/fonts/**` and `web/public/blueprint/overview.svg` are present.
- **Current CI map** (`infra/ci/profile-paths.tsv`, `select-profile.sh`, item 6):
  - Markdown under `docs/` and `management/`, and root `*.md`, classify as `docs` (fast). Automatic documentation-only pipelines are omitted, and a docs path never promotes a mixed change.
  - Non-Markdown under `docs/` and `management/` stays foundation.
  - `core/` (and `web/`, `intelligence/`) is runtime → integration.
  - `core/migrations/`, `core/internal/migrate/`, `core/cmd/migrate/`, `core/internal/{observability,config}/`, `infra/`, CI files and locks are foundation → full.
  - `branch-web` and `branch-api` pipelines select full.
  - Recorded limitation (AGENTS 2026-10-08): all `web/tests/browser/` and `web/tests/fixtures/` changes are full. G005 changes no web path.
- **Governance additions** (AGENTS/CLAUDE 2026-10-08):
  - feature integration through a glab MR to `staging` after ACCEPT;
  - proportional UI validation (not applicable: G005 has no UI);
  - manual testing and operator feedback before final CI (not applicable: G005 has **no** manual-test or operator-feedback stage, so the final local `all` and the hosted run remain last).
- **G004 candidate `4a46da9`** [G004-OBS] (base `201f5d1`, an ancestor of `19aa3c2`; locally validated interim, **not accepted or integrated**):
  - **Migration and schema:**
    - `core/migrations/00002_identity.sql` creates `identity_provider_binding`, `users(id uuid PRIMARY KEY DEFAULT gen_random_uuid(), status, display_name, created_at, updated_at, disabled_at)`, `identity_principals`, `sessions` and `login_transactions`; 14 relations in `app`.
    - Its Up starts with an owner/runtime-role preflight DO block, using `to_regrole(current_database() || '_runtime')`. Its grants run in a DO block with an exact-ACL postcondition.
    - Runtime holds SELECT, INSERT and UPDATE on `users`, with no DELETE.
  - **Core:** the server routes `GET /healthz`, `GET /readyz` and five `/internal/v1/auth/*` routes. `MAIN_GO_IMPORTS` adds the identity packages. The identity package exports no user-ID type; `UserID` is a string field.
  - **Web routes:** `account`, `api/account`, `auth/{callback,login,logout,logout-all,logout-return,provider-logout}`, `sign-in`.
  - **Harness:**
    - Registry 79 rows (fast 33, integration 70, full 77, staging 5); VP1 33; register 23.
    - `events.json` adds five identity events with optional `request_id` (`^[A-Za-z0-9._-]{8,64}$`) and `trace_id` (`^[0-9a-f]{32}$`).
    - New port sets `PORTS_IDN` and `PORTS_STG`.
  - **Result:** "interim, not submitted for acceptance". The final local `all` on the C7 candidate passed 77/77 ([G004-OBS]); the hosted run on the final candidate and ACCEPT are pending.

**r1/r2 observations at `42b21d1` (history):**

- **Core:**
  - The server serves only `GET /healthz` and `GET /readyz`. MCP is dormant and unmounted.
  - The migration set is `{00001_baseline.sql}`.
  - Direct dependencies include goose `v3.28.0` and pgx `v5.11.0`; Go is `1.27.1`.
  - G005 needs **no new dependency**.
- **PostgreSQL:** the image is `pgvector/pgvector:0.8.7-pg17`. `gen_random_uuid()` is built in, and `security_invoker` views are available (PostgreSQL 15 and later).
- **Roles** (`infra/postgres/init/10-roles.sh`):
  - `<db>_owner` is NOLOGIN and owns schema `app`.
  - `<db>_migrator` reaches the owner through `SET ROLE` only.
  - `<db>_runtime` is NOINHERIT, with no superuser and no BYPASSRLS.
  - `<db>_retention` is reserved.
  - Default privileges revoke function EXECUTE from PUBLIC and grant sequences to runtime. Tables receive nothing until a migration grants it.
  - Role names derive from `DATABASE_NAME`, which is configurable and exercised by B-DBRENAME.
- **Harness:**
  - The `migrations` suite runs only in the **full** profile.
  - Path classes: `core/` maps to integration; `core/migrations/` and `core/internal/migrate/` map to full.
- **Pinned counts:**
  - VP1 asserts a fast count of 31.
  - The render register holds 19 rows.
  - The registry has 70 rows (fast 31, integration 63, full 69).
- **Hard-coded migration-version expectations** in `infra/validation/run.sh`:
  - B-MIGRATE-SET;
  - the B-DOCS-EXEC-ROOT migrate block;
  - B-MIGRATE-CLI/CATALOG;
  - B-PERSIST, B-DBRENAME;
  - the off-variant inspect step.

  `TestMigrationLifecycle` expects a baseline-only set.
- **Routes and logs:**
  - Web: `/`, `/api/health`, `/api/ready`. Intelligence: `/healthz`, `/readyz`, `/internal/*`.
  - The log catalogue (`events.json`) is closed (`additionalProperties: false`), and `logrecord_test.go` pins the core event-name list.
- **Rendered paths:**
  - Rendered `docs/foundation` paths are FOLLOW-UP.md (registered), GETTING-STARTED.md and VENDORS.md (unregistered; G005 must not touch them).
  - MIGRATIONS.md and the new TENANCY.md are consumer additions.
- **Reference shape:** the archetype fixture `skills/project-archetype/validation/isolation/sql/10_schema.sql` gives a reference shape ("recommended defaults, not decisions").
  - It is reused where compatible.
  - Its `ON DELETE CASCADE` is **not** reused, because the Spec forbids cascade removal.
  - Its single generic immutability function is replaced by per-table functions, so PL/pgSQL never references a field that a table lacks.

### 2.2 Dependency register (accepted G004 at `7dc4c5c`, observed in `1505440`; final re-confirmation at E3)

**Legend:**
- **[G004-ACC]** = read from the accepted G004 code, byte-identical in `1505440`. These are facts of the accepted implementation; E3 re-confirms them at the integrated staging commit.
- **Differs** = an E3 finding (a changed bound path) that returns to Codex.

| ID | Assumption (r2/r3) | Observation at accepted G004 [G004-ACC] | G005 consequence (r4) |
|---|---|---|---|
| P-G004-1 | `app.users(id uuid PK DEFAULT gen_random_uuid())`, owned by `<db>_owner`. | **Matches** (`00002_identity.sql`). | FK `role_bindings_subject_fkey` → `app.users (id)` (§3.2). |
| P-G004-2 | Runtime holds SELECT and INSERT on `users`; no DELETE. | **Matches, plus table-level UPDATE.** | As r3: the user and creator prechecks use SELECT; test-only synthetic users use INSERT; a runtime change of a referenced `users.id` fails on the subject FK (IT-REF-3). G005 adds no grant on G004 objects. |
| P-G004-3 | `00001_baseline.sql`, then `00002_identity.sql`. | **Matches.** SHA-256: `00001` `db18518a…`, `00002` `8d44b300…`. | G005 is `00003_tenancy.sql`. B-MIGRATE-SET pins both predecessor digests, taken from the integrated tree at E3. |
| P-G004-4 | Identity rows in `users`, `identity_principals`, `sessions`, `identity_provider_binding`. | **Matches** (r3's column details are unchanged). Runtime may INSERT into all four. No role- or owner-bearing column. | Synthetic identity fixtures are seeded through runtime SQL; identity snapshots hash those four tables; `login_transactions` is excluded. |
| P-G004-5 | G004 reconciles version and catalog sites to version 2. | **Sites at `1505440`** (line numbers in §6.2): <ul><li>B-MIGRATE-SET;</li><li>the root migrate block (`0 2 2`);</li><li>`mig_step` counts 0/0/14/14/0/0/14 and the production-refusal version;</li><li>B-PERSIST, B-DBRENAME and both off-variant greps (version 2);</li><li>`identity_catalog`;</li><li>`migrate/integration_test.go`: `identityTables`, `reset`, `metadataAndIdentityOnly`, subtests 02/05/07/08/09;</li><li>`migrate/runner_test.go`: `TestNormalSetIsTheReviewedSet` and `TestProviderAcceptsProtectedStore` (2 sources).</li></ul> | Unchanged from r3 (§6.2): <ul><li>version 3;</li><li>`reset` drops the G005 objects and the three functions first;</li><li>counts 0/0/27/27/14/14/27;</li><li>`runner_test.go` in scope.</li></ul> `migrations_test.go` needs no change: its file rules are generic, and `TestIdentityMigrationShape` reads only `00002`. |
| P-G004-6 | G004 tests may assume 00002 is latest, or may TRUNCATE `users`. | **Confirmed** (§2.1): <ul><li>(a) the `resetIdentity` TRUNCATE (`integration_test.go:134`);</li><li>(b) `TestIdentityCatalog`'s exact table list (`store_test.go:48–61`);</li><li>(c) `run.sh` `identity_catalog`'s 14-object set (2311, 2383);</li><li>(d) the `staging_test.go:258` TRUNCATE.</li></ul> PostgreSQL refuses a `TRUNCATE` of an FK-referenced table without `CASCADE` (SQLSTATE `0A000`; §14 R-7), even when the referencing table is empty. | Unchanged from r3: (a)/(d) append `CASCADE`; (b)/(c) expected sets become identity ∪ G005, still exact. (a)–(c) are validated locally (B-IDENTITY-CATALOG in `migrations`; B-IDENTITY-IT in `ops`). (d): **ESC-1**. |
| P-G004-7 | Registry 79; fast 33, integration 70, full 77, staging 5. | **Fact.** | Forecast unchanged: 83; 33 / 71 / 81 / 5. |
| P-G004-8 | `main.go` digest, imports and routes pinned in B-MCP-SOURCE. | **Matches** (`run.sh` 123, 755–771; `main.go` SHA-256 `923b8238…`). | `main.go` stays byte-identical; G005 only adds the "tenancy not linked" assertion. |
| P-G004-9 | `events.json` gains identity events. | **Matches:** five events, with `request_id` `^[A-Za-z0-9._-]{8,64}$` and `trace_id` `^[0-9a-f]{32}$`. | `tenancy.operation` reuses those patterns. |
| P-G004-10 | APPLICABLE and the register gain G004 entries. | **Matches:** register 23; `APPLICABLE` at `run.sh` 132. | G005 adds two APPLICABLE items and no register row: FOLLOW-UP.md is already registered and the register is path-keyed. |
| P-G004-11 | Route inventory. | **Matches** (§2.1). | No change to the 24-entry set (Q4). R3-Q1 (§13.3) stays an optional proposal. |
| P-G004-12 | G004 may export a user-ID type. | **No exported type** (`UserID` is a string field). | `tenancy.UserID` stays local. |
| P-G004-13 | CI-1 (`identity-staging`, `staging` profile) and the identity `ops` project. | **Matches**, now combined with staging's documentation bypass by MR !7 (§2.1, §6.3). `PORTS_IDN` and `PORTS_STG` exist. | G005 changes no admission rule, job, profile, port set or secret boundary. F-tenancy and the map row coexist with the `F-doc-*` and identity controls. |
| P-G004-14 | Runtime-role derivation pattern. | **Matches:** a preflight DO block, and grants in a DO block with an exact-ACL postcondition. | G005 adopts it, extended to column ACLs (§3.2). |
| P-G048-1..3 | VP1 32, register 20, shared harness bytes, `fonts`/`blueprint` segments. | **Resolved facts** (r3), present in `1505440`; after G004, VP1 is 33 and the register 23. | Unchanged. Reserved `fonts`/`blueprint` confirmed. |
| P-G049 | G049 (draft) relocates skills and changes CI classification. | Concrete overlaps in §2.4. | No G049 work in G005; sequencing and rebase rule in §2.4. |

### 2.3 Gates (single source)

| Gate | Requirement | Established by | Depends on | State (r4) |
|---|---|---|---|---|
| E1 | G048 implementation ACCEPT and integration. | Codex / owner | — | **Satisfied** |
| E2 | **G004 implementation ACCEPT and integration** (Goal dependency, unchanged). | Codex / owner | E1 | **ACCEPT satisfied** ([review](../results/G004-implementation-codex-review.md)). **Integration pending:** MR !7 is open; its reconciliation merge `1505440` had its MR pipeline running at handoff. |
| E3 | Re-confirmation against the **integrated** staging commit that contains G004. Every repository path this Plan binds as [G004-ACC] or as a CI fact (the r4 inventory) is byte-identical there, or each difference is re-derived mechanically (line numbers, counts) and recorded. §2.4 applies too if G049 integrated first. Output: a reviewed reconciliation record or a reissued Plan. A **Differs** in any G004 code, migration or test path, or any non-mechanical difference, returns to Codex. | Claude drafts; Codex confirms | E2 | Pending. r4 pre-reconciles at `1505440`; if MR !7 merges with no further change, E3 reduces to a hash comparison. |
| E4 | Codex approval of the r4 amendment and the E3 result, including the §13 decisions, the ESC-1 disposition and the §2.4 sequencing rule. | Codex | E3 | Pending |
| E5 | Fresh committed base, supervisor approved-input snapshot and digest (`infra/validation/run-approved.sh` inputs) and a fresh isolated worktree. | Codex | E4 | Pending |
| E6 | Named external actions (§12) under the ESC-1 disposition: X1 + X2 (options B, C); X1 + X4 (option A); X1 + X2 + X5 (option A′). No hosted failure run is required (Q6). | Owner | E4 | Pending |
| S0 | Start gate (Claude, read-only, consumes no cycle). <ul><li>E1–E6 recorded;</li><li>wrapper `--check-only` passes;</li><li>HEAD equals the bound base;</li><li>status clean;</li><li>root `.env` absent;</li><li>`PORTS_MIG`, `PORTS_OPS`, `PORTS_REN`, `PORTS_OFF` and `PORTS_IDN` free (`PORTS_STG` too under ESC-1 A′);</li><li>no conflicting Docker resources, including a concurrent G049 wrapper run (§2.4 rule 5);</li><li>GitLab compute-minute quota checked before X2 or X4.</li></ul> | Claude | E1–E6 | Pending |

### 2.4 G049 overlaps (concrete; no blanket dependency)

**Basis.** G049 Goal/Spec at `1505440` (draft r2 plus three owner clarifications), and the coordinator's note that the parallel G049 Plan relocates `skills/project-archetype/` to `skills/saas-foundation/`. The G049 Plan itself was not read. ROADMAP and the G049 Goal record no G005 blocker, and AGENTS.md forbids inventing a blanket CI dependency.

| # | Overlap | G049 change (expected) | G005 change | Nature |
|---|---|---|---|---|
| O1 | `infra/validation/run.sh` (both edit) | `RENDER` (line 99, `skills/project-archetype/generator/render.sh`); Vestive asset paths (629, 709); B-SKILL-UNCHANGED (937–942); B-PROTECTED (972–977); B-CI-CONFIG `want` and PASS text (794–880) if the workflow globs change | APPLICABLE (132); B-MCP-SOURCE (755–771); B-MIGRATE-SET (774–780); root migrate block (1148–1155); `suite_migrations` (1612–1675 plus tenancy wiring); B-PERSIST (2110–2116); B-DBRENAME (2161–2163); B-TOGGLE-OFF (2268–2299); `identity_catalog` (2310–2391); `suite_root` (`tenancy_http` after `mcp_http`, 1164) | Disjoint hunks: a textual rebase, then rerun the covering suites |
| O2 | `infra/ci/profile-paths.tsv` (both edit) | `skills/`, root and nested agent rows, the three handoff files, the Vestive link; possibly a new class | One tab-separated row: `foundation`, `dir`, `core/internal/tenancy/` | Disjoint rows. G005's row uses an existing class; any grammar change in `select-profile.sh` must still accept it |
| O3 | `infra/ci/validate-controls.sh` (both edit) | New agent/skill path controls next to `F-doc-skill` (CONTROLS list at 72–120; table at 1748) | `F-tenancy` after `F-config-dep` (list line 115; table line 1755) | Adjacent edits: a likely textual conflict with a mechanical resolution. The B-CI-CONTROLS count is reported from the integrated tree |
| O4 | `infra/ci/select-profile.sh` and `.gitlab-ci.yml` (G049 only) | Classification and workflow-admission rules | None | Read-only dependency. Under G049 R5, G005's paths (`core/**`, `infra/**`, non-Markdown `management/**`) stay full; the §6.3 table is re-verified at E3/E5 |
| O5 | `infra/validation/required-checks.tsv` | None expected: G049 R4 reuses existing guards | Four rows | If G049 adds rows, G005's count forecast shifts by exactly those rows; restated at E3 |
| O6 | Archetype path references | `skills/project-archetype/` → `skills/saas-foundation/`; archived records keep the old naming (G049 R2) | TENANCY.md cites the archetype tenancy contract (`references/tenancy-authorization.md`); r1–r3 history cites the isolation fixture | TENANCY.md uses the path present in the bound base. The Codex-owned Spec link and this Plan's history stay as written |
| O7 | Protected and unchanged-skill guards (Spec R9) | The guard set follows the new layout | G005 touches no skill or `.archetype` path under either layout | Evidence is reported against the guard set at the bound base |
| O8 | `docs/foundation/` and the render register | Live references (`GETTING-STARTED.md`, `BOOTSTRAP.md`), possibly register rows | `TENANCY.md` (new), `MIGRATIONS.md`, `FOLLOW-UP.md`; no register edit | No shared file. The register count (23) is re-read at E3 |

There is no overlap in `core/**`, `core/migrations/**`, the identity tests, `logrecord`, `web/**` or `intelligence/**`.

**Sequencing and rebase rule.**
1. No blanket dependency in either direction. Neither Goal imports the other's unintegrated changes.
2. **G049 integrates before G005's E5:** E3 re-derives O1–O8 at the integrated commit: paths, line numbers, counts, and the archetype path in TENANCY.md. Mechanical differences are recorded; anything else returns to Codex.
3. **G049 integrates after G005's E5:** the G005 execution checkout and binding are not rebased mid-execution. At G005's MR, staging is merged into the G005 branch as in G004's MR !7. Resolve only the O1–O3 textual conflicts, then rerun the covering suites (`lint controls provenance`, plus `migrations` for `run.sh`) and the MR pipeline (full). This is integration work recorded separately, not a G005 cycle, unless a real failure occurs.
4. **G005 integrates first:** G049 rebases by the same rule. G005's four registry rows and F-tenancy must survive that merge.
5. **No concurrent local runs:** both Goals' wrapper runs use the same `PORTS_*` sets and Docker resources, so they must not run at the same time on one workstation. S0 and `ports_busy` refuse rather than re-map.

## 3. Schema: `core/migrations/00003_tenancy.sql` (medium risk)

### 3.1 Decisions [Plan pins]

- **DS-1 Keys.**
  - IDs are `uuid` values generated in the database by `gen_random_uuid()` (v4).
  - Runtime **cannot supply** `id` or `created_at`, because INSERT is granted on columns only.
  - CHECK constraints refuse the all-zero UUID for every id and for the subject.
  - Names, slugs and provider subjects are never keys.
- **DS-2 Typed scope.**
  - The canonical `scope_type` and `scope_id` columns are kept.
  - `group_id` is set only for Group scope, with FK `(organization_id, group_id) → groups(organization_id, id)`. This proves the Group exists **in the same Organization**.
  - An Organization scope requires `scope_id = organization_id`, with an FK to `organizations`.
  - The single CHECK `role_bindings_scope_consistent` enumerates exactly two branches: Organization scope with `org_owner`/`org_admin`/`org_member`, and Group scope with `group_owner`/`group_admin`/`group_member`, both with source `membership` (r3 item 2). Any other scope, shape, role or source fails that one named rule.
  - **r3 item 3 (supersedes r2's G006-extension note):** BusinessProfile is **not** a RoleBinding scope (ADR-004, Spec r2 R2). G006 derives profile access read-only from current Organization/Group bindings and the single GroupBusinessProfile attachment, and adds no scope, source, column or branch here. A future scope or source (for example `invitation`/`system`, owned by later Goals) needs its own reviewed migration of this CHECK; there is no generic branch.
  - **r3 item 4:** Advisor, Principal and Associate are future engagement functions (ARCHITECTURE "Consultant engagement functions"), not RoleBinding roles. The CHECK rejects them like any other unsupported role.
- **DS-3 Uniqueness.**
  - `UNIQUE (subject_user_id, scope_type, scope_id)` across all sources: at most one membership binding per subject and scope. In G005 every binding is `membership`.
  - Leading on `subject_user_id`, it also serves the subject FK.
- **DS-4 Projections.**
  - `security_invoker = true` views, so G008's RLS on `role_bindings` applies to reads through them.
  - Runtime gets SELECT only. The views would otherwise be auto-updatable; INSERT, UPDATE and DELETE privilege on a view is never granted, so writes fail with 42501 (§14 R-3).
- **DS-5 Removal and regrant.**
  - Removal is `DELETE` of the binding row. The projection loses the row at commit.
  - A regrant inserts a new row with a new UUID and its own role, so an obsolete role cannot be resurrected.
  - There is no lifecycle table or soft-revoke column; audit history is G011's.
- **DS-6 Immutability.**
  - **First line:** column-level UPDATE grants. Runtime can update only `organizations.name`, `groups.name`, `groups.slug` and `role_bindings.role`. Any other column fails with 42501 before any trigger runs.
  - **Second line:** per-table `BEFORE UPDATE` triggers raise SQLSTATE `23000` with a named constraint when any key column changes. This also covers owner-level and future-grant mistakes.
  - Foreign keys use `ON DELETE RESTRICT`: no cascade, and no dangling Group or binding when a parent or subject is deleted directly.
- **DS-7 No RLS in G005.**
  - The tables carry `organization_id`, but RLS stays disabled. G008 owns ENABLE/FORCE, policies and context. This is recorded in TENANCY.md as G008's inventory.
  - Until then, no runtime code path links the model (§4.1).
  - Per Q1 (resolved): MIGRATIONS.md and TENANCY.md gain only a scoped foundation-stage explanation and the mandatory G008 inventory. The eventual ENABLE/FORCE requirement is not weakened, and no tenant data is exposed before its gates.
- **DS-8 Minimal metadata.** Only `name` and `created_at`; no `updated_at`, preferences, branding, onboarding flag or hostname slug.
- **DS-9 Creator ownership (r3 item 1; ADR-004, Spec r2 R5/AC7).**
  - Group creation takes an explicit existing global creator UUID from the trusted in-process caller.
  - One transaction inserts the Group row and exactly one RoleBinding (`scope_type='group'`, `scope_id = group_id = new Group`, `role='group_owner'`, `source='membership'`, subject = creator).
  - A failure of either write, including unknown creator, slug conflict, FK race, statement failure, commit failure or cancellation, rolls back both: neither row remains.
  - **No duplicate authority:** there is no owner, creator or `created_by` column on Groups or anywhere else. Ownership exists only as that binding, enforced by CT-6 A1/N9.
  - **No Organization membership is minted** for the creator, and no provider-claim or global authority is consulted.
  - G018 later supplies the creator from verified onboarding context, never from a browser assertion.
  - Afterwards the owner binding is an ordinary RoleBinding: the internal grant/change/remove primitives accept `group_owner` like other roles. Ownership transfer, last-owner protection and "admin cannot modify owner bindings" are actor-based policy owned by G007/G019; G005 primitives are actor-less model operations and claim none of it.
  - The runtime role already holds INSERT on both tables (§3.3); no grant changes.

### 3.2 Proposed DDL (exact intent; whitespace and comments may differ)

```sql
-- G005: Organization/Group tenancy model and the single RoleBinding authority store.
-- Tenant-owned rows carry organization_id. RLS, policy and request context are G007/G008
-- obligations (docs/foundation/TENANCY.md). goose runs this file in one transaction.

-- +goose Up
-- +goose StatementBegin
DO $$                                                       -- the accepted G004 preflight pattern [G004-ACC P-G004-14]
BEGIN
  IF current_user::text IS DISTINCT FROM
     (SELECT nspowner::regrole::text FROM pg_namespace WHERE nspname = 'app') THEN
    RAISE EXCEPTION 'migrations must run as the owner of schema app, not %', current_user;
  END IF;
  IF to_regrole(current_database() || '_runtime') IS NULL THEN
    RAISE EXCEPTION 'runtime role for database % is missing', current_database();
  END IF;
END
$$;
-- +goose StatementEnd

CREATE TABLE app.organizations (
  id         uuid        NOT NULL DEFAULT gen_random_uuid(),
  name       text        NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT organizations_pkey PRIMARY KEY (id),
  CONSTRAINT organizations_id_nonzero CHECK (id <> '00000000-0000-0000-0000-000000000000'),
  CONSTRAINT organizations_name_valid CHECK (char_length(name) BETWEEN 1 AND 200
    AND name = btrim(name) AND (name COLLATE "C") !~ '[\u0001-\u001f\u007f-\u009f]')
);

CREATE TABLE app.groups (
  id              uuid             NOT NULL DEFAULT gen_random_uuid(),
  organization_id uuid             NOT NULL,
  name            text             NOT NULL,
  slug            text COLLATE "C" NOT NULL,
  created_at      timestamptz      NOT NULL DEFAULT now(),
  CONSTRAINT groups_pkey PRIMARY KEY (id),
  CONSTRAINT groups_organization_fkey FOREIGN KEY (organization_id)
    REFERENCES app.organizations (id) ON DELETE RESTRICT,
  CONSTRAINT groups_organization_id_key UNIQUE (organization_id, id),
  CONSTRAINT groups_organization_slug_key UNIQUE (organization_id, slug),
  CONSTRAINT groups_id_nonzero CHECK (id <> '00000000-0000-0000-0000-000000000000'),
  CONSTRAINT groups_name_valid CHECK (char_length(name) BETWEEN 1 AND 200
    AND name = btrim(name) AND (name COLLATE "C") !~ '[\u0001-\u001f\u007f-\u009f]'),
  CONSTRAINT groups_slug_format CHECK (char_length(slug) BETWEEN 1 AND 63
    AND octet_length(slug) = char_length(slug)
    AND slug ~ '^[0-9a-z]+(-[0-9a-z]+)*$'),
  CONSTRAINT groups_slug_reserved CHECK (slug NOT IN (
    'account', 'admin', 'api', 'auth', 'billing', 'blueprint', 'fonts', 'health',
    'healthz', 'internal', 'login', 'logout', 'mcp', 'members', 'message', 'messages',
    'new', 'onboarding', 'readyz', 'rpc', 'settings', 'sign-in', 'sse', 'system'))
);

CREATE TABLE app.role_bindings (
  id              uuid        NOT NULL DEFAULT gen_random_uuid(),
  organization_id uuid        NOT NULL,
  subject_user_id uuid        NOT NULL,
  scope_type      text        NOT NULL,
  scope_id        uuid        NOT NULL,
  group_id        uuid,
  role            text        NOT NULL,
  source          text        NOT NULL,
  created_at      timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT role_bindings_pkey PRIMARY KEY (id),
  CONSTRAINT role_bindings_organization_fkey FOREIGN KEY (organization_id)
    REFERENCES app.organizations (id) ON DELETE RESTRICT,
  CONSTRAINT role_bindings_subject_fkey FOREIGN KEY (subject_user_id)
    REFERENCES app.users (id) ON DELETE RESTRICT,                      -- [G004-ACC] P-G004-1 (re-confirmed at E3)
  CONSTRAINT role_bindings_group_fkey FOREIGN KEY (organization_id, group_id)
    REFERENCES app.groups (organization_id, id) ON DELETE RESTRICT,   -- MATCH SIMPLE: unchecked when group_id IS NULL
  CONSTRAINT role_bindings_subject_scope_key UNIQUE (subject_user_id, scope_type, scope_id),
  CONSTRAINT role_bindings_id_nonzero CHECK (id <> '00000000-0000-0000-0000-000000000000'),
  CONSTRAINT role_bindings_subject_nonzero CHECK (subject_user_id <> '00000000-0000-0000-0000-000000000000'),
  CONSTRAINT role_bindings_scope_consistent CHECK (
       (scope_type = 'organization' AND scope_id = organization_id AND group_id IS NULL
        AND role IN ('org_owner', 'org_admin', 'org_member') AND source = 'membership')
    OR (scope_type = 'group' AND group_id IS NOT NULL AND scope_id = group_id
        AND role IN ('group_owner', 'group_admin', 'group_member') AND source = 'membership'))  -- r3 item 2
);
CREATE INDEX role_bindings_scope_idx ON app.role_bindings (organization_id, scope_type, scope_id);
CREATE INDEX role_bindings_group_idx ON app.role_bindings (organization_id, group_id) WHERE group_id IS NOT NULL;

-- +goose StatementBegin
CREATE FUNCTION app.organizations_keys_immutable() RETURNS trigger
  LANGUAGE plpgsql SET search_path = pg_catalog AS $$
BEGIN
  IF NEW.id IS DISTINCT FROM OLD.id THEN
    RAISE EXCEPTION 'organizations key columns are immutable'
      USING ERRCODE = 'integrity_constraint_violation', CONSTRAINT = 'organizations_keys_immutable';
  END IF;
  RETURN NEW;
END $$;
-- +goose StatementEnd
-- groups_keys_immutable():        (id, organization_id) IS DISTINCT FROM old values -> same RAISE, CONSTRAINT 'groups_keys_immutable'
-- role_bindings_keys_immutable(): (id, organization_id, subject_user_id, scope_type, scope_id, group_id, source)
--                                 -> same RAISE, CONSTRAINT 'role_bindings_keys_immutable'
CREATE TRIGGER organizations_keys_immutable BEFORE UPDATE ON app.organizations
  FOR EACH ROW EXECUTE FUNCTION app.organizations_keys_immutable();
-- (likewise groups_keys_immutable and role_bindings_keys_immutable)

CREATE VIEW app.organization_memberships WITH (security_invoker = true) AS
  SELECT subject_user_id AS user_id, organization_id, role
  FROM app.role_bindings WHERE scope_type = 'organization' AND source = 'membership';
CREATE VIEW app.group_memberships WITH (security_invoker = true) AS
  SELECT subject_user_id AS user_id, organization_id, group_id, role      -- organization_id: see Q2
  FROM app.role_bindings WHERE scope_type = 'group' AND source = 'membership';

-- +goose StatementBegin
DO $$
DECLARE rt text := current_database() || '_runtime';      -- role names derive from DATABASE_NAME (P-G004-14)
BEGIN
  REVOKE ALL ON app.organizations, app.groups, app.role_bindings,
                app.organization_memberships, app.group_memberships FROM PUBLIC;
  EXECUTE format('GRANT SELECT ON app.organizations, app.groups, app.role_bindings, '
              || 'app.organization_memberships, app.group_memberships TO %I', rt);
  EXECUTE format('GRANT INSERT (name), UPDATE (name) ON app.organizations TO %I', rt);
  EXECUTE format('GRANT INSERT (organization_id, name, slug), UPDATE (name, slug) ON app.groups TO %I', rt);
  EXECUTE format('GRANT INSERT (organization_id, subject_user_id, scope_type, scope_id, group_id, role, source), '
              || 'UPDATE (role), DELETE ON app.role_bindings TO %I', rt);
  -- r3 postcondition (G004 pattern, extended to column ACLs): the five G005 relations' table ACLs (relacl) and
  -- column ACLs (attacl) name only the owner and the runtime role, with exactly the §3.3 privileges; any other
  -- grantee or privilege raises, so the migration commits nothing. (Exact expected strings pinned in execution.)
END
$$;
-- +goose StatementEnd

-- +goose Down
DROP VIEW app.group_memberships;
DROP VIEW app.organization_memberships;
DROP TABLE app.role_bindings;
DROP TABLE app.groups;
DROP TABLE app.organizations;
DROP FUNCTION app.role_bindings_keys_immutable();
DROP FUNCTION app.groups_keys_immutable();
DROP FUNCTION app.organizations_keys_immutable();
```

**Notes:**
- CHECK constraints fire in alphabetical name order after NOT NULL, and pass on UNKNOWN (§14 R-2). Every CHECK above is NULL-free by construction: all columns are NOT NULL except `group_id`, which is guarded by `IS [NOT] NULL`.
- Regex ranges follow the collating sequence (§14 R-4), hence `COLLATE "C"` on `slug` and on the name control-character test, plus explicit `[0-9a-z]` and an ASCII-only `octet_length` guard.
- Foreign-key checks run as the table owner. Runtime needs no REFERENCES privilege.
- Down uses no `CASCADE`, so an unexpected dependent fails loudly.

### 3.3 Privilege matrix (asserted exactly by B-TENANCY-CATALOG)

| Object | `<db>_runtime` | retention, migrator (without SET ROLE), PUBLIC |
|---|---|---|
| `organizations` | SELECT; INSERT (name); UPDATE (name) | nothing |
| `groups` | SELECT; INSERT (organization_id, name, slug); UPDATE (name, slug) | nothing |
| `role_bindings` | SELECT; INSERT (organization_id, subject_user_id, scope_type, scope_id, group_id, role, source); UPDATE (role); DELETE | nothing |
| `organization_memberships`, `group_memberships` | SELECT only | nothing |
| three trigger functions | no EXECUTE (default privileges revoke it from PUBLIC). Positive runtime updates prove the triggers still fire (contingency in §11). | no EXECUTE |

- **Never granted to runtime:**
  - TRUNCATE, REFERENCES or TRIGGER;
  - DELETE on Organizations or Groups;
  - INSERT of `id` or `created_at`;
  - UPDATE of any key column;
  - ownership, DDL or role membership.
- **Unchanged:** no sequence is created, and no object of G001 or G004 changes.

### 3.4 Indexes

| Index | Serves |
|---|---|
| `groups_organization_slug_key` | Organization-local slug uniqueness; slug lookups and ordered listing per Organization |
| `groups_organization_id_key` | Composite FK target; scoped id lookups |
| `role_bindings_subject_scope_key` | Membership uniqueness; subject FK; per-user lookups |
| `role_bindings_scope_idx` | Members of a scope under an explicit Organization predicate; Organization FK |
| `role_bindings_group_idx` (partial) | Composite Group FK checks on owner-level parent deletes |

### 3.5 Risk, rollback and irreversibility

- **Risk: medium.** The migration is additive: new objects only, and no change to G001 or G004 objects. The only coupling is the subject FK onto the G004 user table [G004-ACC]. Its test-fixture consequences are in §6.2 and ESC-1.
- **Irreversible model decisions for review:**
  - table and column names;
  - UUID v4 database-generated keys;
  - the typed-scope representation;
  - DELETE-based removal;
  - the reserved set.

  Once production tenancy data exists, changes need a reviewed forward migration.
- **Down:**
  - drops all G005 objects and **destroys all G005 tenancy data**;
  - leaves identity/session tables and rows untouched, proven by MG-3/MG-5;
  - is refused with `APP_ENV=production` by the existing runner;
  - runs only on owned disposable databases;
  - is documented in MIGRATIONS.md and TENANCY.md.

  No developer or production rollback or reset is authorized.

## 4. Internal contracts (`core/internal/tenancy`)

### 4.1 Layout and boundary

- **One package** `core/internal/tenancy` (Q3) with these files:
  - `tenancy.go`: IDs, entities, vocabulary;
  - `names.go`, `slug.go`, `errors.go`;
  - `store.go`: SQL and constraint-to-code mapping;
  - `service.go`: operations and the transaction helper;
  - `log.go`.
- **Imports:**
  - pgx v5 (`pgx`, `pgconn`, `pgtype`), `log/slog`, the standard library and `core/internal/observability` (for request/trace IDs).
  - **Never** `net/http`, the MCP SDK, `core/internal/identity` (or its WorkOS adapter) or `core/internal/config`.
- **Not linked** into `./cmd/server` or `./cmd/migrate`. B-MCP-SOURCE asserts this, and `main.go` stays byte-identical [G004-ACC P-G004-8].
- The service takes a `DB` interface, which `*pgxpool.Pool` satisfies. Callers supply a runtime-role pool. The package reads no environment, credentials or configuration.

### 4.2 API (exported surface pinned by U-BOUNDARY)

```go
type OrganizationID, GroupID, BindingID, UserID [16]byte   // distinct named types; UserID local (P-G004-12 [G004-ACC])
func ParseOrganizationID(s string) (OrganizationID, error) // likewise for each; see §4.3
type OrganizationRole string // "org_owner" | "org_admin" | "org_member"
type GroupRole string        // "group_owner" | "group_admin" | "group_member"   (r3 item 2)
type ScopeType string        // "organization" | "group";  const SourceMembership = "membership"

type Organization struct{ ID OrganizationID; Name string; CreatedAt time.Time }
type Group struct{ ID GroupID; OrganizationID OrganizationID; Name, Slug string; CreatedAt time.Time }
type Binding struct{ ID BindingID; OrganizationID OrganizationID; SubjectUserID UserID; ScopeType ScopeType
	ScopeID [16]byte; GroupID *GroupID; Role, Source string; CreatedAt time.Time }
type OrganizationMembership struct{ UserID UserID; OrganizationID OrganizationID; Role OrganizationRole }
type GroupMembership struct{ UserID UserID; OrganizationID OrganizationID; GroupID GroupID; Role GroupRole }

type DB interface{ BeginTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error) }
func NewService(db DB, logger *slog.Logger) *Service

// Organizations
func (s *Service) CreateOrganization(ctx, name string) (Organization, error)
func (s *Service) GetOrganization(ctx, org OrganizationID) (Organization, error)
func (s *Service) RenameOrganization(ctx, org OrganizationID, name string) (Organization, error)
// Groups (every call names its owning Organization)
func (s *Service) SuggestGroupSlug(ctx, org OrganizationID, name string) (string, error)
func (s *Service) CreateGroup(ctx, org OrganizationID, creator UserID, name, slug string) (Group, Binding, error)
	// r3 item 1: atomically the Group plus the creator's group_owner Binding (DS-9); slug "" => derived base (§4.4)
func (s *Service) GetGroup(ctx, org OrganizationID, g GroupID) (Group, error)
func (s *Service) GetGroupBySlug(ctx, org OrganizationID, slug string) (Group, error)
func (s *Service) ListGroups(ctx, org OrganizationID) ([]Group, error)               // ORDER BY slug
func (s *Service) RenameGroup(ctx, org OrganizationID, g GroupID, name string) (Group, error) // slug unchanged
func (s *Service) ChangeGroupSlug(ctx, org OrganizationID, g GroupID, slug string) (Group, error)
// Memberships: writes go to role_bindings only; reads go through the projections only
func (s *Service) GrantOrganizationRole(ctx, org OrganizationID, u UserID, r OrganizationRole) (Binding, error)
func (s *Service) ChangeOrganizationRole(ctx, org OrganizationID, u UserID, r OrganizationRole) (Binding, error)
func (s *Service) RemoveOrganizationMembership(ctx, org OrganizationID, u UserID) error
func (s *Service) GrantGroupRole(ctx, org OrganizationID, g GroupID, u UserID, r GroupRole) (Binding, error)
func (s *Service) ChangeGroupRole(ctx, org OrganizationID, g GroupID, u UserID, r GroupRole) (Binding, error)
func (s *Service) RemoveGroupMembership(ctx, org OrganizationID, g GroupID, u UserID) error
func (s *Service) GetOrganizationMembership(ctx, org OrganizationID, u UserID) (OrganizationMembership, error)
func (s *Service) ListOrganizationMemberships(ctx, org OrganizationID) ([]OrganizationMembership, error)
func (s *Service) GetGroupMembership(ctx, org OrganizationID, g GroupID, u UserID) (GroupMembership, error)
func (s *Service) ListGroupMemberships(ctx, org OrganizationID, g GroupID) ([]GroupMembership, error)
```

These 20 methods are the whole surface. There is:
- no delete for Organizations or Groups;
- no move or reparent;
- no cross-Organization listing (G014 owns `account.list_organizations`);
- no Organization bootstrap owner grant; Group creation's creator ownership is the only automatic binding (DS-9);
- no last-owner rule or ownership-transfer workflow (G019);
- no actor/caller parameter, "admin" bypass or "admin cannot touch owners" check (actor-based policy is G007/G019's);
- no BusinessProfile/GroupBusinessProfile operation or profile role (G006 derives access; G018 owns onboarding UX).

### 4.3 Validation (before any SQL)

- **IDs:**
  - Accepted: exactly 36-character hyphenated hexadecimal, case-insensitive, normalized to lowercase.
  - Refused: all-zero, braces, `urn:` forms, 32-hex without hyphens, empty.
  - Code: `invalid_id`.
- **Names (Organization and Group):**
  - valid UTF-8;
  - Unicode whitespace trimmed (`strings.TrimSpace`) and stored trimmed;
  - no C0, DEL or C1 characters (U+0000–U+001F, U+007F, U+0080–U+009F);
  - 1–200 code points.
  - Code: `invalid_name`.
  - The database rule (length, ASCII-space trim, C0/DEL/C1) is never stricter than the service rule.
- **Roles:** typed per scope, from a closed set (r3): Organization `org_owner`/`org_admin`/`org_member`; Group `group_owner`/`group_admin`/`group_member`. Unknown values give `invalid_role`, including:
  - BusinessProfile-style roles (`be_owner`, `be_admin`, `business_profile_*`);
  - staffing functions `advisor`, `principal`, `associate`;
  - presentation labels.

  These are internal semantic codes; client-facing labels resolve through configuration in later Goals. Scope and source are never caller-supplied: each method fixes them.
- **Creator (r3):** `CreateGroup`'s `creator` follows the ID rule (`invalid_id` for zero or malformed). Existence is prechecked (`user_not_found`) and FK-enforced. User status is not consulted: disabled-user bindings remain a model fact (Q7), and authentication of the creator is the caller's/G018's.

### 4.4 Slugs

- **Grammar:** 1–63 ASCII bytes matching `[a-z0-9]+(-[a-z0-9]+)*` (`ValidateSlug`).
  - Explicit slugs are **validated, never normalized**. A non-canonical value gives `invalid_slug`; a reserved value gives `slug_reserved`.
  - No percent-decoding, Unicode folding or transliteration is performed anywhere.
- **Reserved set (24; pinned in Go and in `groups_slug_reserved`; parity unit-tested):**

| Tier | Entries |
|---|---|
| Spec minimum | `account`, `api`, `auth`, `health`, `healthz`, `login`, `logout`, `mcp`, `readyz`, `rpc`, `sse`, `system` |
| Current tree | `internal` (intelligence `/internal/*`); `message`, `messages` (MCP endpoint candidates probed absent by B-MCP-HTTP) |
| Predecessor trees | `sign-in` [G004-ACC; present at `1505440`]; `fonts`, `blueprint` (G048, integrated fact) |
| Archetype-recommended Group reservations (Q4) | `admin`, `billing`, `members`, `new`, `onboarding`, `settings` |

  Other top-level segments (`_next`, `favicon.ico`) cannot pass the grammar. E3 re-inventories the integrated tree.
- **Base suggestion** (`suggestBase(name)`, after the name rule passes):
  1. ASCII `A–Z` becomes lowercase.
  2. `a–z` and `0–9` are kept.
  3. Every other rune, including non-ASCII, `_`, `.`, `/`, `%` and whitespace, is a separator.
  4. Runs of separators collapse to one `-`. Leading and trailing `-` are trimmed.
  5. Cut to 63 bytes and trim a trailing `-`.
  6. An empty result means no base.
- **`SuggestGroupSlug`** (advisory, read-only):
  - Requires that the Organization exists (`organization_not_found`).
  - With no base, returns `slug_required` (an explicit slug is needed).
  - Otherwise it tries candidates `base`, then `base-2` through `base-99`. For each suffix the base is cut so that `base + "-" + n` is at most 63 bytes, and a trailing `-` is trimmed. All candidates are checked in one query, `SELECT slug … WHERE organization_id = $1 AND slug = ANY($2)`.
  - Returns the first candidate that is neither reserved nor used, else `slug_unavailable`.
  - A suggestion grants and reserves nothing.
- **`CreateGroup` with `slug == ""`:**
  - uses the **base only**, with no suffix search, so collisions surface deterministically;
  - no base gives `slug_required`; a reserved base gives `slug_reserved`; a taken base gives `slug_conflict`.
- **`RenameGroup`** never changes the slug. **`ChangeGroupSlug`** changes only the slug; setting the same value is a no-op success.

| Name | Base | Notes |
|---|---|---|
| `Acme Revenue` | `acme-revenue` | Spec example |
| `  Acme   Revenue  ` | `acme-revenue` | whitespace runs |
| `ACME_Revenue.2026` | `acme-revenue-2026` | `_` and `.` are separators |
| `-Acme--Revenue-` | `acme-revenue` | collapse and trim |
| `../Acme%2FRevenue` | `acme-2frevenue` | no decoding; traversal characters are separators |
| `Café Ops` | `caf-ops` | non-ASCII is a separator |
| `Ｒｅｖｅｎｕｅ`, `東京チーム`, `🚀` | none | gives `slug_required` |
| `API` | `api` | reserved: Suggest gives `api-2`; Create without a slug gives `slug_reserved` |
| 70 × `a` | 63 × `a` | cut |
| 62 × `a` + ` b` | 62 × `a` | cut at 63, then trailing `-` trimmed |
| `2026`, `a` | `2026`, `a` | digits-only and one character are valid |

### 4.5 Transactions, concurrency, retries and errors

- **Transactions:**
  - Each operation is one transaction at READ COMMITTED: read-write for mutations, read-only for reads.
  - Success is returned **only after `Commit` returns nil**. Any error rolls back.
  - Every mutation is one statement writing one row, preceded only by scoped existence reads. **Exception (r3 item 1, by design): `CreateGroup`** does exactly two writes in one transaction:
    1. Prechecks: the Organization exists; the creator exists.
    2. `INSERT INTO app.groups (organization_id, name, slug) … RETURNING id, …`.
    3. `INSERT INTO app.role_bindings (organization_id, subject_user_id, scope_type, scope_id, group_id, role, source) VALUES ($org, $creator, 'group', $gid, $gid, 'group_owner', 'membership') RETURNING …`.
    4. Commit.

    Any error at steps 1–4 rolls back both rows. The result returns the Group and its owner Binding.
- **Prechecks** (Organization exists; Group exists **in that Organization**; user or creator exists) give deterministic codes. FKs and unique constraints remain the race-proof enforcement.
- **Concurrency semantics (pinned):**
  - **Duplicate grant or slug:** the unique constraint admits exactly one row. Losers get `membership_exists` or `slug_conflict`; nothing is overwritten.
  - **Concurrent role changes on one binding:** row locks serialize them. The **last committed change wins**, and the binding UUID is unchanged. There is **no** optimistic-concurrency or lost-update-protection claim.
  - **Change versus remove:** if the remove commits first, the waiting change re-evaluates its predicate, matches 0 rows and returns `membership_not_found`.
  - **Grant versus owner-level parent or subject delete:** RESTRICT plus FK key-share locks mean either the grant fails (23503) or the delete fails. There is never a dangling row.
  - **Concurrent `CreateGroup` (r3):** slug collisions leave exactly one Group and exactly one owner binding. Losers get `slug_conflict` with no Group and no binding. A creator deleted (owner-level) between precheck and binding insert makes the binding insert fail 23503 (`user_not_found`), and the Group insert rolls back with it.
- **Retries:** none. Every call is a single attempt. Transient failures return `unavailable`, and callers may retry.
- **Error type:**
  - `*tenancy.Error{Code}`. `Error()` is `"tenancy: <code>"`, never IDs, names, slugs or driver text.
  - `Unwrap` exposes only context cancellation or deadline errors. PostgreSQL `Detail` text, which can contain row values, never leaves the package.

| Source | Code |
|---|---|
| Input validation | `invalid_id`, `invalid_name`, `invalid_slug`, `slug_reserved`, `slug_required`, `invalid_role` |
| Precheck miss or 0-row update/delete. Unknown and foreign selectors give the **identical** code. | `organization_not_found`, `group_not_found`, `user_not_found`, `membership_not_found` |
| 23505 `groups_organization_slug_key` / `role_bindings_subject_scope_key` | `slug_conflict` / `membership_exists` |
| 23503 `groups_organization_fkey`, `role_bindings_organization_fkey` / `role_bindings_group_fkey` / `role_bindings_subject_fkey` | `organization_not_found` / `group_not_found` / `user_not_found` |
| 23514 `*_name_valid` / `groups_slug_format` / `groups_slug_reserved` / `*_nonzero` / `role_bindings_scope_consistent` | `invalid_name` / `invalid_slug` / `slug_reserved` / `invalid_id` / `invalid_role` |
| Suggestion exhausted | `slug_unavailable` |
| SQLSTATE class 08, 40, 53 or 57; connection or pool closed; context canceled or deadline exceeded; commit error | `unavailable` |
| Anything else, including 42501 and trigger 23000 | `internal` |

### 4.6 Logging (one record per mutation; G003 logger and catalogue)

- **Event:** `tenancy.operation`.
- **Levels:** `info` for `ok`, `warn` for `rejected` (validation, not-found, conflict), `error` for `failed` (`unavailable` or `internal`).
- **Fields** (closed):
  - `operation` ∈ {`organization.create`, `organization.rename`, `group.create`, `group.rename`, `group.change_slug`, `membership.grant`, `membership.change`, `membership.remove`}. `group.create` is one record for the atomic Group-plus-owner transaction (r3);
  - `scope_type` (membership operations only);
  - `outcome` ∈ {`ok`, `rejected`, `failed`};
  - `error_code`, required exactly when the outcome is not `ok`;
  - `request_id` and `trace_id`, only when `observability.FromContext` provides them. They reuse the G004-observed catalogue patterns `^[A-Za-z0-9._-]{8,64}$` and `^[0-9a-f]{32}$` [G004-ACC P-G004-9].
- **Excluded:** no UUIDs (consistent with G004's restricted-reference policy), names, slugs, roles, driver errors or payloads. Reads are not logged.
- The event is registered in `events.json` with exact enums. This is the only catalogue addition.

### 4.7 Explicitly not added

- HTTP, BFF or MCP routes; tools; CLI; environment variables; startup hooks; test authenticators; seed data.
- Provider-claim or hostname-derived membership; global or implicit roles; synthesized cross-scope memberships.
- Onboarding state, invitations, BusinessProfile/GroupBusinessProfile schema or derived-access joins (G006), policy, RLS, audit ledger, domains or branding.
- Any owner/creator/`created_by` column (DS-9), Organization membership minted by Group creation, ownership transfer or last-owner policy.
- Advisor/Principal/Associate engagement functions, staffing assignments or any staffing schema (future G022/G024/G027/G037).
- UI of any kind, including the Groups/Business Profiles sibling management areas, which are G018 context only.
- Any change to `.env.example`, configuration, `main.go`, Compose, Dockerfiles or `.gitlab-ci.yml`.

## 5. Tests and evidence (designed now; implemented and run only after E4/E5)

### 5.1 Placement

- **Unit tests** (`core-check`; covered by the existing B-UNIT-CHECKS and B-CHECKS-ROOT):
  - **U-SLUG:** the §4.4 table, all 24 reserved entries, and the grammar negatives listed under IT-SLUG-2.
  - **U-NAME:** `""`, whitespace only, 201 code points, NUL, TAB, DEL, U+0085, invalid UTF-8; valid 200 × `é` (400 bytes); trimming.
  - **U-ID**, **U-ROLE**, **U-ERR:** the mapping table is driven from synthetic `*pgconn.PgError` values, and `Error()` never contains an input.
  - **U-PARITY:** the reserved set, roles (the six r3 codes), scope types, source and bounds in Go equal the literals parsed from the embedded `00003_tenancy.sql`.
  - **U-BOUNDARY** (`go/parser`, **production files only**, i.e. non-`_test.go`): imports stay inside the §4.1 allow-list; there is no exported `http.Handler`; the exported `Service` method set equals the pinned 20 names.
  - Test-only files may use `net/http/httptest`, `observability.Middleware` (IT-LOG-1) and direct-SQL fixtures. They are scoped separately: B-MCP-SOURCE proves that neither the tenancy package nor any test helper is linked into a production binary.
- **Integration tests** (build tag `integration`; real PostgreSQL in the existing `migrations` disposable project, run in the run-labelled `core-check` image on `${project}_data`):
  - **B-TENANCY-CATALOG** = `TestTenancyCatalog`, in `core/internal/migrate/tenancy_integration_test.go` (package `migrate`). It needs only the migration and catalog queries, so S2 does not depend on the S3 package.
  - **B-TENANCY-MIGRATION** = `TestTenancyMigration` (same file) plus the persist seed/verify phases across a PostgreSQL restart.
  - **B-TENANCY-IT** = `TestTenancyModel` in `core/internal/tenancy/integration_test.go`, plus the log checks.
  - All model operations use a pool logged in as **`<db>_runtime`**. The tests first assert `current_user = session_user = <db>_runtime` with no superuser, BYPASSRLS or memberships.
  - Owner-level cases use a test-only migrator connection with `SET ROLE <db>_owner`.
  - Credentials arrive through the existing 0600 env-file pattern of B-MIGRATE-IT.
  - Absent settings **fail, never skip**.
- **Root:** **B-TENANCY-HTTP** route-absence probes (§5.2 H).

### 5.2 Scenario catalogue (expected state outcomes; each negative asserts its SQLSTATE and constraint or code, plus before/after row-set hashes)

**Fixtures:**
- **Users:** ≥ 5 synthetic users, seeded by runtime SQL into `users`, `identity_principals` and `sessions` [G004-ACC P-G004-4].
  - A dedicated creator **F** creates the fixture Groups. The r3 creator-owner bindings therefore never collide with the users under test.
  - Membership tests use non-creator users unless they deliberately test the creator.
- **Organizations:** A and B.
- **Groups:** A/g1 and A/g2; B/g3, with g3's slug equal to g1's. Each has exactly one `group_owner` binding for F at creation, and those rows are included in the expected row sets.
- Every assertion is scoped to fixture IDs.
- Row-set hash = `md5(string_agg(t::text, E'\n' ORDER BY id))` over fixture rows.

| ID | Scenario | Expected |
|---|---|---|
| CT-1 | `app` object set | Exactly: metadata ∪ the 14 identity relations [G004-ACC] ∪ the G005 set (3 tables, 2 views, 8 indexes, 3 functions, 3 triggers). No sequence added. No `tenants`, agencies, client organizations, business profiles or GroupBusinessProfile links, billing, policies, labels, domains, onboarding, staffing/engagement or audit objects. The exact column lists of the three G005 tables are pinned; Groups have no owner/creator column. |
| CT-2 | Ownership and attributes | Every G005 object owned by `<db>_owner`. Functions: `prosecdef = false`, `proconfig = {search_path=pg_catalog}`, no EXECUTE for PUBLIC, runtime or retention. |
| CT-3 | ACLs | Table, view and **column** ACLs (`relacl`, `attacl`) equal the §3.3 matrix exactly. `has_*_privilege` is false for every unlisted role and privilege. |
| CT-4 | Constraint catalogue | Names, types and `pg_get_constraintdef` (whitespace-normalized) equal the pinned §3.2 text. FK `confdeltype = 'r'`, MATCH SIMPLE. `role_bindings_scope_consistent` has exactly the two branches, with the six r3 role codes. |
| CT-5 | Indexes, views, triggers | Index definitions exact; views have `reloptions = {security_invoker=true}` and pinned definitions; three BEFORE UPDATE ROW triggers, enabled. |
| CT-6 | Single durable authority versus exactly two read-only projections (relation-kind inventory; definitions follow this table) | All three assertions A1–A3 below hold for the proposed schema, with zero violations. |
| CT-7 | RLS inventory | `relrowsecurity` and `relforcerowsecurity` are false for the three tables, and no `pg_policy` rows exist. Recorded as G008's obligation, not as isolation. |
| IT-ID-1 | Create A and B; create g1–g3 with creator F | Nonzero distinct v4 UUIDs; Groups carry their Organization; same-slug g1/g3 succeed. Each Group has exactly one F `group_owner` binding. |
| IT-ID-2 | Rename the Organization and a Group; change g1's slug | UUIDs, `organization_id`, `created_at` and binding row-set hash unchanged; a rename leaves the slug unchanged. |
| IT-ID-3 | Runtime SQL: UPDATE `groups.organization_id`, `organizations.id`, each `role_bindings` key column; INSERT with an explicit `id` or `created_at` | 42501 every time; rows unchanged. |
| IT-ID-4 | Owner SQL: the same key updates | 23000 with constraint `*_keys_immutable`; rows unchanged. |
| IT-ID-5 | Owner SQL: zero UUID for each id and for the subject | 23514 `*_nonzero`. |
| IT-REF-1 | Runtime grants of each Organization and Group role through the service | One row each, with the exact columns. |
| IT-REF-2 | Service: unknown or foreign Organization, Group or user; zero or malformed IDs | `organization_not_found` / `group_not_found` / `user_not_found` / `invalid_id`; no rows. Foreign and unknown errors are byte-identical. |
| IT-REF-3 | Runtime direct INSERT: unknown subject or Organization; a Group from B with Organization A; Organization scope with `group_id`; Group scope without `group_id`; `scope_id` not equal to `organization_id`/`group_id`; zero subject. **r3:** runtime `UPDATE app.users SET id = …` of a referenced user (G004 grants runtime table-level UPDATE [G004-ACC P-G004-2]) | 23503 on the named FK, or 23514 `role_bindings_scope_consistent` or `role_bindings_subject_nonzero`; the referenced-user key change fails 23503 on `role_bindings_subject_fkey`; row sets unchanged. |
| IT-REF-4 | Deterministic races, as owner: delete a user, Group or Organization versus a runtime grant, in both commit orders (lock wait detected through `pg_stat_activity`) | Exactly one side fails 23503; there is no dangling row. 50 randomized rounds hold the same invariant. |
| IT-VOC-1 | Runtime matrix (r3): <ul><li>scope type {organization, group, business_profile, business_entity, `''`, `Organization`};</li><li>role {the six valid codes, `owner`, `admin`, `member`, `be_owner`, `be_admin`, `business_profile_owner`, `advisor`, `principal`, `associate`, `''`};</li><li>source {membership, direct_grant, invitation, system, `''`}.</li></ul> | Only the **six** valid combinations insert: Organization × `org_owner`/`org_admin`/`org_member`, and Group × `group_owner`/`group_admin`/`group_member`, all with `membership`. Every other combination, including cross-scope (`group_owner` at Organization scope) and BusinessProfile and staffing cases, fails 23514 `role_bindings_scope_consistent`. Service typed roles reject the rest with `invalid_role`. |
| IT-UNIQ-1 | 32 concurrent grants of the same **non-creator** user and scope, for both Organization and Group scope, repeated 20 rounds; plus a grant to the Group's creator F | Exactly 1 success and 31 `membership_exists`; 1 row. Granting F any role on its own Group gives `membership_exists` (the owner binding already occupies that scope). |
| IT-UNIQ-2 | Deterministic interleaving: transaction A inserts, transaction B inserts the same binding and blocks, A commits | B gets 23505 `role_bindings_subject_scope_key`. |
| IT-PROJ-1 | Grant `org_member`, change to `org_admin`, change to `org_owner`, remove, regrant `org_member`; the same sequence for Group scope with `group_member` → `group_admin` → `group_owner` | 1 / 1 / 1 / 0 / 1 rows for the user. The UUID is unchanged across changes; the regrant has a new UUID and the regranted role. Projection rows and shapes, including `organization_id` on the Group view, are exact at each step. The creator's owner row is untouched throughout. |
| IT-PROJ-2 | Runtime INSERT, UPDATE and DELETE on both views | 42501; rows unchanged. |
| IT-PROJ-3 | A reader on a separate connection after each commit; a concurrent reader loop during changes | Sees the committed state immediately; never two rows, and never an uncommitted role. |
| IT-PROJ-4 | Snapshot of unrelated bindings around every change and removal | Hash unchanged. |
| IT-MULTI-1 | One seeded G004 user U [G004-ACC] with: <ul><li>`org_owner`@A, `org_member`@B;</li><li>`group_owner`@g4 (a Group U creates in A);</li><li>`group_admin`@g1, `group_member`@g2, `group_member`@g3.</li></ul> Then change A and remove g1. | The other bindings are unchanged. U's user, principal and session row hashes are unchanged. |
| IT-MULTI-2 | User V with only a Group role in A; user W with only an Organization role in B; **r3:** user C who only *created* a Group in B (no other binding) | V has no Organization projection row; W has no Group projection row; C has exactly the one `group_owner` Group row and **no** Organization row in B. Nothing is synthesized. |
| IT-OWN-1 (r3) | `CreateGroup(A, creator C, name, slug)` | Exactly one new Group row and exactly one binding: subject C, scope Group, `group_owner`, `membership`, `organization_id` A, `group_id = scope_id` = the new Group. Organization projection unchanged for C. Returned Group and Binding equal the committed rows. No owner column exists (CT-1/CT-6). |
| IT-OWN-2 (r3) | Creator unknown, zero or malformed; Organization unknown or foreign | `user_not_found` / `invalid_id` / `organization_not_found`; **no** Group and no binding (row-set hashes). |
| IT-OWN-3 (r3) | Injected failures through the fault `DB` wrapper: (a) failure at the binding insert, after the Group insert; (b) commit failure; (c) context cancellation between the two inserts | `unavailable`; neither the Group nor the binding exists, and no success is returned. |
| IT-OWN-4 (r3) | Deterministic race: owner-level delete of creator C commits between the creator precheck and the binding insert (lock wait detected through `pg_stat_activity`); the reverse order too | Either the delete fails 23503 and the creation succeeds, or the binding insert fails 23503 (`user_not_found`) and the Group rolls back. Never a Group without its owner, and never a dangling binding. |
| IT-OWN-5 (r3) | 16 concurrent `CreateGroup` calls for one Organization and slug, each with a different creator | Exactly one Group and exactly one owner binding (its own creator). 15 `slug_conflict` with no Group or binding. |
| IT-OWN-6 (r3) | Model primitives on an owner binding: `ChangeGroupRole(creator → group_admin)`, then `RemoveGroupMembership(creator)` | Both succeed as ordinary single-row binding writes (same UUID on change; row gone on remove). This documents that last-owner/transfer policy is **not** in G005 (G019), and no other row changes. |
| IT-SLUG-1 | Every reserved entry, by service and by runtime SQL | `slug_reserved` / 23514 `groups_slug_reserved`. |
| IT-SLUG-2 | Grammar negatives by service and runtime SQL: `""`, 64 × `a`, `Acme`, `acme_revenue`, `acme.revenue`, `-acme`, `acme-`, `acme--revenue`, `acme%2frevenue`, `%2e%2e`, `..`, `a/b`, `a b`, `é`, `ａｃｍｅ`, Cyrillic `асme`, ` acme`, `acme\n` | `invalid_slug` / 23514 `groups_slug_format`. |
| IT-SLUG-3 | Same slug in A and B; same-Organization create and edit collisions | Cross-Organization succeeds; within an Organization, `slug_conflict` with rows unchanged. |
| IT-SLUG-4 | 16 concurrent creates of one slug; 16 Groups concurrently changed to one slug; create versus edit; deterministic interleavings | Exactly 1 success each time. The losers' slug, name and UUID are unchanged. |
| IT-SLUG-5 | `SuggestGroupSlug`: free base, taken base, reserved base, unsluggable name; suggestion taken before create | `base`, `base-2`, `api-2`, `slug_required`; the later create gives `slug_conflict`. |
| IT-SLUG-6 | `GetGroupBySlug` with Organization A for a slug that exists only in B | `group_not_found`. |
| IT-SCOPE-1 | With Organization A and B's Group: Get, Rename, ChangeSlug, Grant, Change, Remove, Get/List memberships | `group_not_found` / `membership_not_found`; B rows unchanged; errors identical to the unknown-ID case. |
| IT-ATOM-1 | Each rejected input class | Row-set hashes unchanged; no success returned. |
| IT-ATOM-2 | Fault-injecting test `DB` wrapper around the runtime pool: (a) commit failure (real rollback); (b) failure at the mutation statement; (c) context cancellation mid-transaction; (d) closed pool | `unavailable`; rows unchanged; never a success result. |
| IT-CONC-1 | 32 concurrent role changes on one binding; deterministic A-then-B interleaving; change versus remove both orders | One row with the same UUID and role ∈ the requested set, equal to the last committed value in the deterministic case. After a remove, a change gives `membership_not_found`. The projection equals the row. |
| IT-LOG-1 | Success, rejected and failed paths with canary names (§5.3), run through `observability.Middleware` in-test for correlation | Every `tenancy.operation` record has the expected `operation`, `outcome`, `error_code` and level, and the supplied `request_id`/`trace_id`. |
| MG-1 | Empty to full: reset, then `Up` | Applies 00001–00003 (output lines asserted); version 3. |
| MG-2 | Repeated `Up` and `Status` | `no pending migrations`; all three applied. |
| MG-3 | Populated upgrade: reset, `UpTo(2)`, seed identity rows [G004-ACC], identity snapshot, `UpTo(3)` | Identity snapshot unchanged. Runtime CRUD smoke on the new objects, including `CreateGroup` with a seeded creator: one Group plus its owner binding. |
| MG-4 | Catalog after `Up` equals catalog after reapply | Equal, including `pg_proc`, `pg_trigger`, `pg_constraint` and column ACLs. |
| MG-5 | `Down` to 2 | G005 objects absent; identity snapshot unchanged; catalog equals the pre-00003 snapshot; G005 fixture rows gone (documented). |
| MG-6 | Persist: seed phase writes IDs to `/out` (0600); harness stops and restarts the M project's PostgreSQL with its volume kept; verify phase | Rows and version 3 present after the restart. |
| H-1 | Core (own namespace and host port): GET and POST `{}` on `/organizations`, `/groups`, `/memberships`, `/role-bindings`, `/tenancy`, `/internal/v1/{organizations,groups,memberships,role-bindings,tenancy}` | 404 with no JSON model payload. |
| H-2 | Web host: GET and POST on `/organizations`, `/groups`, `/memberships`, `/api/{organizations,groups,memberships,role-bindings,tenancy}` | 404. |

**CT-6 definition (r2 correction 1).** CT-6 is implemented as one catalog function, `authorityViolations`, in `TestTenancyCatalog`. It returns violation codes, and the empty set means pass.

**Inventory.**
- **Relations covered:** every relation in schemas `app` and `public`, plus any relation owned by one of the four application roles in any other schema. (The owner cannot create schemas, so this is a backstop.)
- **Classification by `pg_class.relkind`:**
  - **storage** = `r` (table), `p` (partitioned table), `m` (materialized view), `f` (foreign table). A materialized view stores rows, so it counts as storage, never as a projection.
  - **projection** = `v` (view).
  - Sequences, indexes and composite types carry no user-visible role column of their own. A composite type used as a column type is caught through that column.
- **Role-bearing column:** a live user column (`attnum > 0`, not dropped) whose name matches `^(role|roles)$` or `_roles?$`.
- **Owner-bearing column (r3 item 1):** a live user column whose name matches `(^|_)(owner|owners|creator)(_|$)` or `^created_by(_|$)`. A1 treats it like a role-bearing column in storage, because Group ownership may exist only as a `group_owner` binding. G004 identity tables have none [G004-ACC]; any match there is reconciled at E3.

| Assertion | Positive requirement (the proposed schema passes) | Violation code if broken |
|---|---|---|
| A1 Durable storage | Across all **storage** relations, the role-bearing columns are exactly `{app.role_bindings.role}`, and there are **no** owner-bearing columns (r3). G004 identity tables have neither [G004-ACC P-G004-4; any match is reconciled at E3]. | `stored_role_outside_authority` |
| A2 Projections | Across all **projection** relations, the role-bearing columns are exactly `{app.organization_memberships.role, app.group_memberships.role}`; no other view is exempt. For each of the two views: <ul><li>`relkind = 'v'`;</li><li>`reloptions = {security_invoker=true}`;</li><li>exact column list: `(user_id, organization_id, role)` and `(user_id, organization_id, group_id, role)`;</li><li>the pinned `pg_get_viewdef` text (whitespace-normalized);</li><li>its `_RETURN` rule's `pg_depend` references exactly one relation, `app.role_bindings`, including the column `role_bindings.role`;</li><li>no rewrite rule other than `_RETURN`;</li><li>no trigger (so no `INSTEAD OF` write path).</li></ul> | `unexpected_role_projection`, `projection_source`, `projection_definition`, `projection_write_path` |
| A3 No other write path | Checked for runtime, retention, migrator and PUBLIC using `has_table_privilege` (INSERT, UPDATE, DELETE, TRUNCATE) and `has_column_privilege` (INSERT, UPDATE) on every role-bearing relation. <ul><li>**Storage:** only `app.role_bindings` may be writable, and only by runtime (INSERT/UPDATE of `role`, DELETE).</li><li>**Projections:** no write privilege for any role; runtime holds SELECT only.</li></ul> | storage: `writable_role_column`; projection: `writable_projection` |

**CT-6-NEG (negative controls; run during approved execution only).**
- Each control applies its defect as the owner (migrator with `SET ROLE <db>_owner`) inside its own transaction on the disposable `migrations` database. It asserts that `authorityViolations` returns **exactly** the listed set, then `ROLLBACK`s. PostgreSQL DDL is transactional.
- After all controls, `authorityViolations` must again return the empty set and the CT-1..CT-5 catalog snapshot must be unchanged.

| Control | Deliberate defect (one transaction) | Expected violation set (exact) |
|---|---|---|
| N1 | `ALTER TABLE app.groups ADD COLUMN role text` | {`stored_role_outside_authority`} |
| N2 | `CREATE TABLE app.membership_mirror (user_id uuid, organization_id uuid, role text)` | {`stored_role_outside_authority`} |
| N3 | `CREATE MATERIALIZED VIEW app.membership_cache AS SELECT subject_user_id, role FROM app.role_bindings` | {`stored_role_outside_authority`} |
| N4 | `CREATE VIEW app.extra_memberships AS SELECT subject_user_id, role FROM app.role_bindings` | {`unexpected_role_projection`} |
| N5 | `CREATE OR REPLACE VIEW app.organization_memberships WITH (security_invoker = true)` with the same columns but `'org_owner'::text AS role`, so the role is not derived from `role_bindings.role` | {`projection_definition`, `projection_source`} |
| N6 | `GRANT INSERT ON app.group_memberships TO <db>_runtime` | {`writable_projection`} |
| N7a | An `INSTEAD OF INSERT` trigger (with its function) on `app.organization_memberships` | {`projection_write_path`} |
| N7b | `CREATE RULE … AS ON INSERT TO app.organization_memberships DO INSTEAD …` | {`projection_write_path`} |
| N8 | N1, then `GRANT UPDATE (role) ON app.groups TO <db>_runtime` | {`stored_role_outside_authority`, `writable_role_column`} |
| N9 (r3) | `ALTER TABLE app.groups ADD COLUMN owner_user_id uuid`, and separately `ADD COLUMN created_by uuid` | {`stored_role_outside_authority`} each |

**r2-review catalog clarifications (part of the r2 ACCEPT; carried into r3 unchanged):**
- **A2 source dependency.** A2's one-relation requirement concerns **source** relations. The `_RETURN` rule's internal dependency on its own owning view is classified separately. A2 asserts exactly the approved source relation (`app.role_bindings`) and source column (`role`), and rejects any other source dependency. It neither counts the owning view as a second authority nor ignores external dependencies.
- **A3 PUBLIC handling.** A3 checks PUBLIC by inspecting PUBLIC ACL entries (table `relacl` and column `attacl`, including owner/default-privilege semantics), not by passing PUBLIC to the role privilege functions. Named application roles are checked separately with the privilege functions, without `SET ROLE` owner. No privilege is waived.
- The detector stays bounded to this schema/privilege contract.

The view write-denial evidence remains separate and unchanged: runtime INSERT, UPDATE and DELETE on both views fail with 42501 (IT-PROJ-2).

### 5.3 Privacy canaries (existing tools only)

- **Canaries:**
  - The harness generates per-run canaries with `new_canary`, registered in the run secrets file and therefore covered by B-SECRET-CANARY.
  - They reach the test through a 0600 file in a read-only mount.
  - The test uses them as Organization and Group names on success, rejected and failed paths.
- **Avoiding DB-log echo of row values.** PostgreSQL CHECK violations put the failing row in `DETAIL` in the server log, and B-SECRET-CANARY scans Compose logs. So:
  - canary-bearing rejections are always rejected by **service validation before any SQL**;
  - database-reached failures in canary flows are only unique or FK violations, whose DETAIL shows key columns only;
  - direct-SQL CHECK negatives use non-canary values;
  - bind parameters are not logged on error (§14 R-6).

  The scan of Compose logs therefore also proves that no canary reached the PostgreSQL log. Changing the server log configuration is out of scope; see the follow-up in Q8.
- **Harness checks on the captured `/out/tenancy-core.log`:**
  - `logcheck validate --service core`: every line is classified;
  - `logcheck scan --canaries`: no canary present;
  - positive control: a canary planted in a scratch copy **must** be detected.
- Test output never formats names. Failure messages print codes and hash prefixes only.

### 5.4 Forced-failure control (AC11)

- **F-1 (local, required):**
  - In a scratch clone of the final commit, invert one assertion in `integration_test.go` (the expectation of IT-UNIQ-1).
  - Run `infra/validation/run.sh … migrations` there.
  - Expected: B-TENANCY-IT FAIL and harness exit 1.
  - The accepted B-CI-CONTROLS gate semantics already prove that any FAIL record fails `validate.sh` and the gating job.
- **Q6 (resolved):** F-1 plus the inherited gate-failure controls compose the required guarantee. No hosted failure run is planned. The one mandatory hosted run is a **passing** full run on the final tree (X2, or X4 under ESC-1 option A).

### 5.5 Evidence (outside the repository; never `/tmp`, which WSL restarts lose)

- **Location:** `<home>/.local/state/platform-evidence/g005-execution-<base7>/`.
- **Contents:**
  - wrapper pre/post outputs and report directories (`results.tsv`, logs);
  - catalog dumps (constraint, index and view definitions; table/column ACLs; role attributes);
  - migration step outputs;
  - test logs with subtest counts and no SKIP;
  - `logcheck` outputs;
  - concurrency outcome counts;
  - NUL-safe tracked, untracked and ignored status;
  - start/end owned-resource inventories;
  - hosted pipeline, job and commit/tree identifiers and published summaries.
- No secret, canary value or display name is retained.

## 6. Harness, CI and inherited-guard reconciliation (minimal; no new framework)

### 6.1 New required IDs and counts

| ID | Suite | Profiles | Recording code and registry row added at (r2 correction 2) |
|---|---|---|---|
| B-TENANCY-CATALOG | migrations | full | S2, with `00003` and `TestTenancyCatalog` |
| B-TENANCY-MIGRATION | migrations | full | S2, with `TestTenancyMigration` and the restart phase |
| B-TENANCY-IT | migrations | full | S3, with `core/internal/tenancy`, `TestTenancyModel` and the tenancy log event |
| B-TENANCY-HTTP | root | integration, full | S4, with the probe |

| Registry state | Rows | fast | integration | full | staging |
|---|---|---|---|---|---|
| Clone base (fact) | 70 | 31 | 63 | 69 | — |
| After G048 (fact at `19aa3c2`) | 71 | 32 | 64 | 70 | — |
| After G048 + G004 (fact at `1505440`) | 79 | 33 | 70 | 77 | 5 |
| After S2 | 81 | 33 | 70 | 79 | 5 |
| After S3 | 82 | 33 | 70 | 80 | 5 |
| **After S4 = final G005 (forecast)** | **83** | **33** | **71** | **81** | **5** |

- The fast count is unchanged, so VP1 is not touched.
- **Counting (r4).** Counts come from `required-checks.tsv` and `infra/ci/lib.sh` `CI_SUITES_*`: each profile's suites plus `final`. `run.sh all` equals the full profile and contains neither `unit` nor `staging`. The 83 rows are therefore the 81 full-profile IDs plus B-UNIT-CHECKS and B-IDENTITY-STAGING (§6.4 S7; ESC-1).
- A registry row and its recording code always land in the same step. A row is **never** added before its code (which would make a missing required ID), and code never records an unregistered ID.
- `B-REQUIRED-CHECKS` keeps its completeness semantics in every focused run: within each executed suite, a missing, duplicated or skipped ID fails. No check is disabled or deferred at any step.

### 6.2 Existing checks reconciled (exact G005 additions only; reason codes and negative controls kept)

Each reconciliation happens in the step that first makes it necessary, never later. Adding `00003` at S2 immediately changes every migration-version expectation, so all of them move to S2.

| Check | G005 change | Step | Covering suites rerun in that step |
|---|---|---|---|
| B-MIGRATE-SET | Set = {00001, 00002, 00003}. Adds: the 00001 and accepted 00002 SHA-256 values equal pinned digests (accepted migrations byte-identical). | S2 | lint |
| B-MIGRATE-CLI / B-MIGRATE-CATALOG | Per-step expectations on G004's observed `mig_step <count>` form [G004-ACC P-G004-5]: <ul><li>fresh status lists 00003 pending;</li><li>apply → 00001–00003 applied, version 3, count **27**;</li><li>reapply → `rolled_back`/`applied 00003`, count 27;</li><li>rollback → `rolled_back 00003`, version 2, count **14**;</li><li>inspect → 00003 pending, version 2, count 14;</li><li>apply → `applied 00003`, version 3, count 27;</li><li>production refusals unchanged.</li></ul> The count is 14 identity + 13 G005 relations; it is recomputed at E3. | S2 | migrations |
| B-MIGRATE-IT (`migrate/integration_test.go`) and `migrate/runner_test.go` | [G004-ACC]: <ul><li>`reset` drops the G005 views, `role_bindings`, `groups`, `organizations` and the three functions **before** G004's identity tables (the subject FK blocks `DROP TABLE app.users` otherwise);</li><li>a `metadataIdentityAndTenancyOnly` helper (identity ∪ the 13 G005 relations) replaces `metadataAndIdentityOnly` at version 3;</li><li>tests 02/05/07/08 add the 00003 lines and version 3;</li><li>test 09 downs three times (3→2→1→0);</li><li>`TestNormalSetIsTheReviewedSet` adds `00003_tenancy.sql`;</li><li>`TestProviderAcceptsProtectedStore` expects 3 sources.</li></ul> The subtest count is kept. | S2 | unit, migrations |
| B-DOCS-EXEC-ROOT migrate block; B-PERSIST; B-DBRENAME; B-TOGGLE-OFF inspect | Version 3 and `00003_tenancy.sql` pending/applied lines (G004's observed form: root `0 3 3`). | S2 | root, ops, rename, off |
| G004 database tests and checks (P-G004-6, [G004-ACC]) | Mechanical only: <ul><li>(a) `identity/integration_test.go` `resetIdentity` TRUNCATE gains `CASCADE`;</li><li>(b) `identity/store_test.go` `TestIdentityCatalog` "objects and owner" expects identity ∪ the three G005 tables (still exact);</li><li>(c) `run.sh` `identity_catalog`: `IDENTITY_OBJECTS` comparison becomes identity ∪ the 13 G005 relations at version 3 (still exact; B-IDENTITY-CATALOG detail text updated);</li><li>(d) `identity/workos/staging_test.go` TRUNCATE gains `CASCADE`: **ESC-1**.</li></ul> No identity semantics, guard, code or route changes. | S2 | migrations (B-IDENTITY-CATALOG), ops (B-IDENTITY-IT); (d) only by X4 or the ESC-1 disposition |
| B-MCP-SOURCE | Adds: `core/internal/tenancy` is absent from `go list -deps` of both `./cmd/server` and `./cmd/migrate`. `main.go` digest, imports and routes keep the accepted G004 pins. | S3 | lint |
| B-TOGGLE-OFF-ADDITIONS | `APPLICABLE` += `core/internal/tenancy` (S3), then `docs/foundation/TENANCY.md` (S5, when the file exists). | S3, S5 | off |
| MIGRATIONS.md and FOLLOW-UP.md text (no docs-block change) | Current-set row, down implications, validation sentence, Q1 note; STAGE-TENANCY partial ticks. | S5 | provenance, migrations, off |
| B-CONTRACT-TOOLS (`logrecord_test.go`) | The core event list gains `tenancy.operation`, with positive and negative cases (missing `error_code` on a non-`ok` outcome, extra key, unknown enum). | S3 | contracts, ops (catalogue consumers) |
| B-CI-CONTROLS | Adds control `F-tenancy` (§6.3), together with the `profile-paths.tsv` row, once the package exists. | S3 | controls, lint |
| B-CI-CONFIG / B-REQUIRED-CHECKS | Registry rows per §6.1; counts reported for each step. | S2, S3, S4 | lint, final |
| B-RENDER-PLAN / B-RENDER-REPEAT | **Unchanged**: FOLLOW-UP.md is already registered, and MIGRATIONS.md and TENANCY.md are not rendered paths. The register is path-keyed, so its count (23) and reason text need no change. | S5 | provenance |
| B-SCOPE | The G005 allow-list (§7); every step's paths are inside it from S1. | all | final |

**r4 site map at `1505440`** (`run.sh` SHA-256 `623a5570…`). The S2 edits are those of the table above, unchanged in content.

| Check | Site(s) | Change at S2 |
|---|---|---|
| B-MIGRATE-SET | `static_migration_set`, 774–780 | Set gains `00003_tenancy.sql`; predecessor digests pinned |
| B-DOCS-EXEC-ROOT migrate block | 1148–1155 | `00003` pending and applied lines; versions `0 3 3` |
| B-MIGRATE-CLI / B-MIGRATE-CATALOG | `mig_step` 1631–1639; production refusal 1645; PASS texts 1646 and 1648; comment 1612 | Counts 0/0/27/27/14/14/27; version 3 or 2 per step; the refusal keeps version 3 |
| B-IDENTITY-CATALOG | Call 1662 (now at version 3); `IDENTITY_OBJECTS` 2311; comparison 2383; PASS text 2390 | Expected set is identity ∪ the 13 G005 relations (`r:`, `v:` and `i:` entries); still exact |
| B-MIGRATE-IT | 1664–1675 (10/10 subtests) | Count unchanged |
| B-PERSIST | 2110, 2116 | `00003` applied; version 3 |
| B-DBRENAME | 2161, 2163 | Version 3 |
| B-TOGGLE-OFF | Apply 2268; inspect 2270–2271; PASS text 2299 | Version 3; `00003` applied; two version-3 lines |
| Go tests | `migrate/integration_test.go` 72–89 (`reset`), 207–240 (`metadataAndIdentityOnly`) and subtests 02/05/07/08/09 (266–410); `runner_test.go` 115–122 and 296; `identity/integration_test.go` 134; `store_test.go` 48–61; `workos/staging_test.go` 258 | As in the table above |

New `run.sh` wiring, each piece in the same step as its test files and registry row:
- **S2, `suite_migrations` after B-MIGRATE-IT:** `go test -tags integration -run '^TestTenancyCatalog$'` and `-run '^TestTenancyMigration$'` in `./internal/migrate/`, followed by the persist seed → stop/start of PostgreSQL → verify phases. Records B-TENANCY-CATALOG and B-TENANCY-MIGRATION.
- **S3, `suite_migrations` after those:** `-run '^TestTenancyModel$' ./internal/tenancy/`, followed by `logcheck validate`/`scan` and the planted control. Records B-TENANCY-IT.
- **S4, `suite_root` after `mcp_http`:** `tenancy_http`. Records B-TENANCY-HTTP.

Each ID is recorded PASS only when its exact pinned subtest count passes with no SKIP or FAIL. No new port set, project, image, wrapper or suite.

### 6.3 CI selection (smallest mapping change)

**Current controls (r4 item 6; facts at `1505440`).** MR !7 combined staging's documentation bypass with G004's identity CI and changed no path class.

G005's new or changed paths classify as follows:
- **Foundation (full):** `core/migrations/00003_tenancy.sql`; `core/internal/migrate/*_test.go`; `infra/validation/**` and `infra/ci/**`; `management/v0/results/G005-scope-allowlist.txt` (non-Markdown under `management/`); `core/internal/tenancy/**` once S3's row exists.
- **Runtime (integration):** `core/internal/identity/{integration,store}_test.go`, `core/internal/identity/workos/staging_test.go`, and `core/internal/tenancy/**` before S3's row exists.
- **Docs (fast; never promotes or demotes a mixed change):** `docs/foundation/{TENANCY,MIGRATIONS,FOLLOW-UP}.md` and the Result `.md`.

**Expected CI profile for the G005 delivery** (as AGENTS.md requires):

| Context | Admission | Selected profile (reason) | Jobs and IDs (forecast) |
|---|---|---|---|
| Local wrapper `W(all)`; `validate.sh --profile full` | Explicit | full (`local-explicit`) | 81 full-profile IDs; `W(unit)` adds B-UNIT-CHECKS |
| Push to a feature branch with an open MR | Suppressed | — | None |
| Push to a feature branch without an MR (X1 uses `ci.skip`) | Admitted | fast (`feature-push`) | `fast` only (33) |
| X2: `web`/`api` pipeline on the evidence branch | Admitted | full (`web-explicit`/`api-explicit`) | `fast` 33 + `gate` full 81 |
| X4 (ESC-1 A): `web` pipeline on protected `identity-qualification` | Admitted | full (`web-explicit`); `staging` for the staging job | `fast` 33 + `gate` 81 + `identity-staging` 5 |
| Later MR to `staging` (integration after ACCEPT; MR source) | Admitted (non-documentation paths) | full (`foundation`: in byte order, the first full-class path is `core/internal/migrate/integration_test.go`); a merged-result pipeline selects full | `fast` + `gate` full; no `identity-staging` |
| Push to `staging` after the merge | Admitted | full (`foundation`) | As above |
| A later Markdown-only edit of `TENANCY.md` or `MIGRATIONS.md` | Automatic pipeline omitted | An explicit `web`/`api` request selects full | — |

- Per the G004 review's integration-profile disposition, the union of the hosted `fast` and `gate` (full) jobs covers every integration ID; there is no separate integration run.
- `identity-staging` never runs in G005's ordinary contexts (ESC-1).
- TENANCY.md contains no harness-executed block. MIGRATIONS.md's executed blocks keep the accepted documentation-bypass behaviour: a Markdown-only change starts no automatic pipeline, and B-MIGRATE-CLI and B-TOGGLE-OFF execute those blocks in full runs.
- No AGENTS-recorded browser-test limitation applies: G005 changes no `web/` path.
- The proportional-testing backlog and G049 are not adopted here; required tests are not reduced.

- **Problem:** B-TENANCY-IT lives in `migrations`, which runs only in full. Today a later change confined to `core/internal/tenancy/` maps to `runtime`, which selects integration and would skip the model tests.
- **Change:**
  - `infra/ci/profile-paths.tsv` gains `foundation	dir	core/internal/tenancy/`, next to the existing migrate-dependency rows.
  - `infra/ci/validate-controls.sh` gains the table-driven control `[F-tenancy]="full full foundation|add:core/internal/tenancy/service.go"` and its `CONTROLS` entry, following the F-migrate-dep precedent. At `1505440` that means CONTROLS line 115 and the table after `[F-config-dep]` at line 1755.
  - A focused `--only F-tenancy` inner-loop run may need the shared B-group fixture (`N12q`) first, as the G004 reconciliation recorded for the schedule controls. The gate is the complete `controls` suite.
- **Unchanged:** `.gitlab-ci.yml`, `infra/ci/lib.sh` profiles and budgets, `select-profile.sh`, `hosted-context.sh` and VP1.
- **Alternative (Q5):** run B-TENANCY-IT in `ops`. Rejected because it needs a second disposable project with migrator credentials and splits the tenancy evidence.

### 6.4 Commands (execution only; container-only; existing entry points)

There is **one command form**: the accepted approved-input wrapper around the existing `run.sh` suites. It is used from S2 onward; only the suite list changes per step.

```sh
# W(<suites>): the approved-input wrapper (inputs from E5); <EVID> per §5.5
bash infra/validation/run-approved.sh --snapshot <S> --snapshot-sha256 <D> --count <N> --path-re '<RE>' \
  --plan management/v0/plans/G005-organization-group-memberships-plan.md -- \
  --base <BASE> --allow-file management/v0/results/G005-scope-allowlist.txt --report-dir <EVID>/<step> <suites>
```

| Step | Files at this step (§10) | Command(s), in order | IDs that must PASS (every executed suite complete) |
|---|---|---|---|
| S2 | `00003`, `migrate` tests (including `runner_test.go`), `run.sh` hunks of §6.2/S2, 2 registry rows, G004 test reconciliation (a)–(d) [G004-ACC] | `W(unit lint migrations)`, then `W(root ops rename off)` | All `unit`, `lint` and `migrations` IDs, including B-UNIT-CHECKS (`runner_test.go`), B-MIGRATE-SET/CLI/CATALOG/ENV/IT, B-IDENTITY-CATALOG, **B-TENANCY-CATALOG** and **B-TENANCY-MIGRATION**; then all `root`, `ops`, `rename` and `off` IDs (version-3 sites; B-IDENTITY-IT/-E2E/-PROBES/-PRIVACY/-CONFIG in the identity `ops` project). B-IDENTITY-STAGING is **not** in any local suite: see ESC-1. |
| S3 | `core/internal/tenancy` and unit tests, `TestTenancyModel`, `events.json`, `logrecord_test.go`, B-MCP-SOURCE hunk, `APPLICABLE` item, profile row, F-tenancy, 1 registry row | `W(unit)` as the fast inner loop (stack-free README `platform:check`: gofmt, vet, U-* tests), then `W(lint controls contracts unit migrations ops off)` | B-UNIT-CHECKS; B-MCP-SOURCE, B-CI-CONFIG; B-CI-CONTROLS (including F-tenancy); B-CONTRACT-TOOLS; **B-TENANCY-IT** plus the S2 IDs; B-OBS-*; B-TOGGLE-OFF-ADDITIONS |
| S4 | `tenancy_http` hunk, 1 registry row | `W(lint root)` | B-SHELL, B-CI-CONFIG; **B-TENANCY-HTTP**, B-MCP-HTTP, B-DOCS-EXEC-ROOT |
| S5 | TENANCY.md, MIGRATIONS.md, FOLLOW-UP.md, `APPLICABLE` item | `W(lint provenance migrations off)` | B-LITERALS (scans `docs/foundation`), B-RENDER-*, B-PROTECTED; B-MIGRATE-CLI (runs MIGRATIONS.md blocks verbatim) and the tenancy IDs; B-TOGGLE-OFF-ADDITIONS and B-TOGGLE-OFF (also runs MIGRATIONS.md blocks) |
| S6 | Scratch clone only (F-1) | `bash infra/validation/run.sh … migrations` in that clone | **Expected FAIL** of B-TENANCY-IT and harness exit 1 |
| S7 | Final tree | `W(all)`, then `W(unit)`; local CI parity `bash infra/ci/validate.sh --profile full` | All 81 full-profile IDs [forecast] plus B-UNIT-CHECKS. B-IDENTITY-STAGING follows ESC-1 (83 registry rows in total; §6.1) |

- **Final gates in every run.** Every `W(…)` run includes the harness `final` checks (B-SCOPE, B-SECRET-CANARY, B-REQUIRED-CHECKS, B-DOCKER-ONLY).
- **No reliance on later steps.** No step depends on a file or wiring from a later step. No registry row, ID or test is disabled to make an intermediate run pass.
- **No edits during runs.** The repository is never edited while a harness run is active (B-RENDER-PLAN tree-state check).
- **Covering suites after shared edits.** A shared-file edit made while correcting inside a step reruns that row's covering suites before the step closes.
- TENANCY.md documents the commands that run the tenancy checks: the `migrations` suite and `platform:ci`.
- **No manual-testing or operator-feedback stage** (AGENTS 2026-10-08 rule): G005 has no UI or operator qualification. Final qualification (S7 `all` and `unit`, then X2 or X4) therefore needs no feedback gate. X4 and X5 (ESC-1) are automated real-provider regression runs, not manual-testing or operator-feedback stages; either one uses the same final candidate, after S7.

## 7. File inventory and scope allow-list

**Create:**
- `core/migrations/00003_tenancy.sql` [G004-ACC number; final at E3]
- `core/internal/tenancy/{tenancy,names,slug,errors,store,service,log}.go`
- `core/internal/tenancy/{tenancy,names,slug,errors,parity,boundary}_test.go`
- `core/internal/tenancy/integration_test.go` (`//go:build integration`): `TestTenancyModel`, including the fault-injecting `DB` wrapper (S3)
- `core/internal/migrate/tenancy_integration_test.go` (`//go:build integration`): `TestTenancyCatalog` with CT-6/CT-6-NEG, `TestTenancyMigration` and the persist phases (S2)
- `docs/foundation/TENANCY.md`. Contents:
  - the model, constraint and grant inventory;
  - contracts, error codes and concurrency semantics;
  - the slug rules and reserved set;
  - down/destructive notes;
  - the G007/G008 handoff: tables, views and queries to secure; a Group without its BusinessProfile attachment is not completed onboarding (G018);
  - the role vocabulary (Organization/Group owner/admin/member) and creator ownership (DS-9). These are semantic codes, not client-facing labels. BusinessProfile access is derived (G006/G007), not a role. Advisor/Principal/Associate are future staffing functions, not roles;
  - commands. No harness-executed docs block (§6.3);
  - the effect of the identity fixtures' `TRUNCATE … CASCADE` (§6.2 (a)/(d)): disposable and test databases, and the root database used by the local CAP-3 block, lose G005 bindings together with the identity rows (ESC-1);
  - the archetype tenancy contract, cited at its path in the bound base (§2.4 O6).
- `management/v0/results/G005-scope-allowlist.txt` (first execution write) and `management/v0/results/G005-organization-group-memberships-result.md`

**Modify:**
- `core/internal/migrate/integration_test.go` (expectations, `reset` order, version-3 helper) and `core/internal/migrate/runner_test.go` (normal set and source count) [G004-ACC]
- G004 test files, mechanical only (§6.2 (a)/(b)/(d)) [G004-ACC]: `core/internal/identity/integration_test.go`, `core/internal/identity/store_test.go`, `core/internal/identity/workos/staging_test.go` (ESC-1)
- `infra/validation/run.sh` (including the `identity_catalog` expected set, (c)), `infra/validation/required-checks.tsv` (`run.sh` is shared with G049: §2.4 O1)
- `infra/validation/tools/internal/logrecord/{events.json,logrecord_test.go}`
- `infra/ci/{profile-paths.tsv,validate-controls.sh}` (shared with G049: §2.4 O2/O3)
- `docs/foundation/MIGRATIONS.md`: current-set row, down implications, validation sentence, and the RLS note per Q1
- `docs/foundation/FOLLOW-UP.md` (registered customization): STAGE-TENANCY Group/membership items marked partly done (G005)
- E3 re-derives these G004 paths from the accepted, integrated tree; any addition returns to Codex.

**Untouched:**
- `core/cmd/**`, `core/internal/{config,health,mcp,observability}`; all **non-test** files of `core/internal/identity`; accepted migrations (`00001`, `00002`)
- web, intelligence, contracts
- Compose files, Dockerfiles, `.gitlab-ci.yml`
- `infra/ci/{lib,validate,select-profile,hosted-context,cleanup-owned,runner-preflight}.sh`
- `infra/validation/{lib.sh,run-approved*.sh,render-customizations.tsv}`
- `.env.example`, `.archetype`, skills
- VENDORS.md, GETTING-STARTED.md
- governance, ADRs, Goal/Spec, ROADMAP/PROGRESS, SECURITY.md

Scope allow-list (exact regular expressions; r3 added the G004 test paths; unchanged in r4; re-derived at E3 and after any G049 integration, §2.4):

```text
^core/migrations/00003_tenancy\.sql$
^core/internal/tenancy/(tenancy|names|slug|errors|store|service|log)\.go$
^core/internal/tenancy/(tenancy|names|slug|errors|parity|boundary|integration)_test\.go$
^core/internal/migrate/(integration|tenancy_integration|runner)_test\.go$
^core/internal/identity/(integration|store)_test\.go$
^core/internal/identity/workos/staging_test\.go$
^infra/validation/(run\.sh|required-checks\.tsv)$
^infra/validation/tools/internal/logrecord/(events\.json|logrecord_test\.go)$
^infra/ci/(profile-paths\.tsv|validate-controls\.sh)$
^docs/foundation/(TENANCY|MIGRATIONS|FOLLOW-UP)\.md$
^management/v0/results/G005-(scope-allowlist\.txt|organization-group-memberships-result\.md)$
```

## 8. Acceptance-criteria mapping (Spec AC1–AC12)

| AC | Tests and evidence |
|---|---|
| AC1 | <ul><li>CT-1/2/6;</li><li>IT-ID-1..5 (UUIDs, parent, no duplicate tenant model, rename/slug preservation, runtime 42501, owner trigger 23000);</li><li>U-BOUNDARY (no move/reparent method).</li></ul> |
| AC2 | IT-REF-1..4 (runtime positives; service and direct-SQL negatives for unknown, foreign, mismatched and zero cases; races), CT-4, post-failure row hashes. |
| AC3 | <ul><li>IT-VOC-1: the full r3 matrix with six valid combinations, including the valid `group_owner` grant/change cases and premature `direct_grant`/`invitation`, BusinessProfile and staffing-function negatives;</li><li>IT-OWN-1: `group_owner` creation;</li><li>IT-PROJ-1: `group_owner` change;</li><li>CT-4: two-branch definition with the six codes;</li><li>IT-UNIQ-1/2.</li></ul> |
| AC4 | <ul><li>IT-PROJ-1..4 (IT-PROJ-2: view write denial);</li><li>CT-3;</li><li>CT-6 A1–A3: one durable role column, exactly the two SELECT-only projections deriving `role` from `role_bindings`, no other write path;</li><li>CT-6-NEG N1–N9: an extra stored role or owner/creator column, a materialized or writable mirror, a third role-bearing view, a re-sourced projection, write grants and trigger/rule write paths each fail;</li><li>U-BOUNDARY: the only SQL writes target `role_bindings`.</li></ul> |
| AC5 | IT-MULTI-1/2 with a seeded G004 user and its principal and session rows [G004-ACC], including a creator-owned Group and a creator with no Organization role; no provider-claim path (U-BOUNDARY: no identity import). |
| AC6 | U-SLUG; IT-SLUG-1..6 (every reserved entry, the grammar matrix, cross/same-Organization, concurrency, suggestion, scoped lookup); IT-ID-2. |
| AC7 | <ul><li>IT-SCOPE-1, IT-ATOM-1/2, IT-CONC-1, IT-UNIQ-2; the documented last-committed-wins semantics (§4.5);</li><li>**creator ownership (r3):** IT-OWN-1 (exactly one Group plus its `group_owner` binding; no Organization role), IT-OWN-2 (unknown/zero creator → neither row), IT-OWN-3 (injected failure between or after the two writes → neither row), IT-OWN-4/5 (races and concurrent creates), IT-OWN-6 (no last-owner policy claimed).</li></ul> |
| AC8 | <ul><li>B-MCP-SOURCE (not linked; `main.go`, imports and routes at the accepted pins);</li><li>B-TENANCY-HTTP H-1/H-2;</li><li>B-WEB-BOUNDARY route set unchanged;</li><li>CT-1/7 (no BusinessProfile/GroupBusinessProfile, policy, RLS, onboarding, routing, staffing or audit objects; no owner column);</li><li>`.env.example` and configuration unchanged (B-ENV-EXAMPLE);</li><li>scope allow-list.</li></ul> |
| AC9 | <ul><li>MG-1..6 (B-TENANCY-MIGRATION);</li><li>CT-1..5 and runtime DDL/role/metadata negatives (B-TENANCY-CATALOG: CREATE/ALTER/DROP/TRUNCATE on G005 tables, `SET ROLE` owner/migrator and goose metadata access all 42501);</li><li>reconciled B-MIGRATE-*, B-PERSIST, B-DBRENAME;</li><li>B-MIGRATE-SET predecessor digests.</li></ul> |
| AC10 | <ul><li>Full wrapper run: all inherited auth/session (B-IDENTITY-*, [G004-ACC]; B-IDENTITY-STAGING per ESC-1), browser, first-paint, token, boundary, variant, dependency and provenance IDs;</li><li>IT-LOG-1 plus `logcheck` validate/scan and the planted control (§5.3);</li><li>B-SECRET-CANARY;</li><li>B-CONTRACT-TOOLS.</li></ul> |
| AC11 | <ul><li>Step-wise and final wrapper runs (§6.4) execute the commands verbatim on the actual root and hash-matched copies (`all`, off variant included).</li><li>F-1 forced failure plus the inherited gate-failure controls (Q6).</li><li>One **passing** hosted full pipeline (X2, or X4 under ESC-1 option A) on the final tree, with commit, tree, job and required-ID results.</li></ul> |
| AC12 | E5 binding, wrapper pre/post, NUL-safe status, allow-list, start/end inventories, failure history, B7 accounting with the cycle-4 checkpoint, Result; Codex ACCEPT. No policy, RLS, onboarding or product claim. |

## 9. Delegation

- **Planning (r1–r4):** none, per instruction.
- **Execution:** Claude main owns **every** file and the integration. The Goal suggests an optional test/migration delegate; I am **declining** it. The store contract, its tests and the shared harness wiring overlap tightly, and a separate owner would add integration risk without parallel benefit at this size.
- **Read-only reviewers** (zero cycles unless their findings trigger corrective work, which then counts against B7):
  1. a security/SQL reviewer after S3, covering DDL, grants, triggers, error mapping, non-disclosure and the privacy design;
  2. a harness/evidence reviewer before S7.

## 10. Sequence (after E1–E6)

Each step is self-contained (r2 correction 2). Its files, its `run.sh`/registry wiring and the inherited expectations it changes land together, and its gate is the §6.4 command for that step.

| Step | Work (tests, wiring and registry together) | Gate (§6.4) |
|---|---|---|
| S0 | Start gate (§2.3) | Pass |
| S1 | Write the scope allow-list | `W(lint)`: B-SCOPE and lint IDs; nothing else changed yet |
| S2 | <ul><li>Migration 00003 (with the G004 preflight/postcondition pattern and the six role codes);</li><li>`migrate` tests: `TestTenancyCatalog` with CT-6/CT-6-NEG including N9, `TestTenancyMigration`, persist phases;</li><li>all migration-version and catalog reconciliations, including `runner_test.go` and the `reset` order;</li><li>G004 test reconciliation (a)–(d) [G004-ACC];</li><li>`suite_migrations` wiring and 2 registry rows.</li></ul> | `W(unit lint migrations)`, then `W(root ops rename off)` |
| S3 | <ul><li>`core/internal/tenancy` and unit tests, including atomic `CreateGroup` with creator ownership (DS-9); `TestTenancyModel` including IT-OWN-1..6;</li><li>`events.json` and `logrecord_test.go`; B-MCP-SOURCE hunk; `APPLICABLE` item;</li><li>profile row and F-tenancy;</li><li>B-TENANCY-IT wiring and 1 registry row.</li></ul> Then the security review. | `W(unit)` inner loop, then `W(lint controls contracts unit migrations ops off)` |
| S4 | `tenancy_http` wiring and 1 registry row | `W(lint root)` |
| S5 | Docs: TENANCY.md, MIGRATIONS.md, FOLLOW-UP.md; `APPLICABLE` item | `W(lint provenance migrations off)` |
| S6 | F-1 forced-failure control in a scratch clone | Expected FAIL recorded |
| — | **Cycle-4 checkpoint** (expected around S4–S7) | Reassessment (§11) |
| S7 | Final wrapper `all` and `unit` on the final tree; local `validate.sh --profile full` parity | All 81 full-profile IDs and B-UNIT-CHECKS PASS |
| S8 | Hosted qualification on the same tree, per the ESC-1 disposition: X1 + X2 (B, C); X1 + X4, the consolidated protected pipeline that replaces X2 (A); or X1 + X2 + X5 (A′). The owner starts any protected `web` pipeline in the GitLab UI. | Recorded |
| S9 | Result; return control to Codex. Integration afterwards follows the GitLab MR-to-`staging` workflow only after ACCEPT and owner authorization; it is not G005 execution. | — |

## 11. Budget, forecast and stops

**B7: at most seven shared meaningful cycles; checkpoint after cycle 4** (unchanged; r3 and r4 planning consumed zero). Forecast: **5 of 7, two in reserve. Risk: moderate.** There is no provider or browser work; the main uncertainty is the G004 harness interplay, now concretely mapped (§6.2). r3's creator-ownership transaction and tests fall inside C2; the G004 test reconciliation falls inside C3. The forecast is unchanged.

| Cycle | Expected material issue |
|---|---|
| C1 | Catalog expectations: column ACL text, regex/collation behavior, trigger firing without EXECUTE, definition normalization. |
| C2 | Model tests: concurrency harness timing, error mapping, fault wrapper. |
| C3 | Harness and G004 reconciliation: version sites, reset/TRUNCATE interplay, restart phase, log catalogue. |
| C4 | First full run: off/rename variants, controls, inherited suites. |
| C5 | Hosted full-run correction (runner, disk, timing; the G003 history shows this risk). |

**Checkpoint after C4:**
- record the ACs satisfied and remaining;
- record the repeated failures;
- state whether G004 reconciliation produced unplanned semantic changes;
- forecast the hosted run.

Do not spend the reserve on harness redesign.

**Pre-declared contingencies** (within scope; recorded in the Result):
1. **Trigger EXECUTE.** If runtime updates fail on trigger EXECUTE, grant EXECUTE on the three SECURITY INVOKER, table-access-free functions to runtime.
2. **Regex escapes.** If the PostgreSQL regex rejects the `\u` escapes, use an equivalent `chr()`-built bracket expression under `COLLATE "C"`.

**Stop and escalate** (in addition to the Goal's list):
- any E-gate missing at S0;
- a non-UUID or unreachable G004 user key;
- a need to change accepted G004 semantics or tests beyond mechanical reconciliation;
- any edit to an accepted migration;
- a need for RLS, policy, onboarding, BusinessProfile/GroupBusinessProfile schema, profile roles, staffing roles, a public surface or a new dependency;
- a second role store, an owner/creator column or a writable projection;
- any **Differs** finding at E3 versus the [G004-ACC] observations, or a G004 test change beyond §6.2 (a)–(d);
- B-IDENTITY-STAGING coverage not resolved per the ESC-1 disposition before final qualification;
- the integrated G004 staging commit differs from `1505440` in a bound G004 code, migration or test path (an E3 **Differs**);
- a G049 change that overlaps G005 beyond §2.4 O1–O8, or a G049 integration that needs a non-mechanical G005 change;
- CI changes beyond §6.3;
- missing hosted access (**UNMET, never skipped**);
- thrashing or B7 exhaustion.

## 12. Named external actions (none performed; each needs separate owner authorization)

- **X1:** push the final tree to a dedicated evidence branch (for example `g005-tenancy-evidence`) with `ci.skip`, as in G004, so the push itself starts no automatic pipeline.
- **X2:** run the existing pipeline on X1, triggered from the web (`branch-web`) or the API (`branch-api`); both select full (fact at `1505440`, §6.3). The `fast` job plus `gate` with the **full** profile; record the pipeline, jobs, commit, tree and published summaries.
  - Check the GitLab compute-minute quota first (Free plan). `ci_quota_exceeded` is a resource stop recorded as **UNMET**, not a CI failure and never a skip.
  - Not run separately under ESC-1 option A: X4 supplies the same full evidence on the same tree.
- r1's optional X3 (a hosted failure run) is **withdrawn** per the Q6 resolution. F-1 plus the inherited gate-failure controls suffice; the mandatory hosted evidence is one passing full run (X2, or X4 under option A).
- **X4 (ESC-1 option A; r3, made concrete in r4 from G004's accepted precedent, pipeline 2928428140):**
  - After S7, an authorized fast-forward push of the final G005 commit to the protected `identity-qualification` branch. It is a descendant of `4a46da9` for any base that contains G004, so no force is needed.
  - The owner starts the `web` pipeline in the GitLab UI; the staging job and the forwarding gate refuse an `api` source.
  - It runs `fast`, `gate` (full) and `identity-staging` on one commit. That revalidates B-IDENTITY-STAGING with the `CASCADE` fixture change (§6.2 (d)) and replaces X2, so no full run is duplicated.
  - Record the pipeline, jobs, commit, tree and published summaries. Check the GitLab compute-minute quota first. Per-run WorkOS staging users are created and deleted by the test.
  - It uses the existing WorkOS staging access and protected-variable boundary; G005 changes no setting, secret, variable or branch protection. If the BACKLOG qualification-simplification follow-up has replaced the branch by then, X4 uses the approved replacement.
- **X5 (ESC-1 option A′; new in r4):** a local `W(staging)` run on the workstation, with the owner's existing 0600 `IDENTITY_STAGING_ENV_FILE` outside the repository (CAP-3 precedent). It is a real-provider action: per-run WorkOS users are created and deleted, only the API key is registered as a secret, and only the sanitized result is kept. No hosted action or protected-branch update is involved. X2 still runs.
- **After ACCEPT (not G005 execution):** integration through `glab mr create` to `staging` with source-branch deletion, per `docs/agents/GITLAB-BRANCH-WORKFLOW.md` and owner authorization. If G049 integrated after G005's E5, the MR first takes the §2.4 rule-3 reconciliation.

No provider account, secret, project-setting, schedule or runner change is needed. X4 and X5 use only the existing protected qualification path or the owner's existing credential file.

## 13. Escalation, r4/r3 decisions, earlier resolutions and review pins

### 13.1 ESC-1 (kept from r3; re-assessed in r4; needs a Codex/owner disposition before E4)

**ESC-1: G005's required subject FK breaks G004's credential-bearing staging fixture, and only a real-provider run can revalidate the fix.**

- **Finding [G004-ACC]** (substance unchanged from r3):
  - `core/internal/identity/workos/staging_test.go:258` (`workos_staging`) migrates the disposable database to the latest embedded version. It then runs the owner `TRUNCATE app.login_transactions, app.sessions, app.identity_principals, app.users, app.identity_provider_binding`. The hosted `suite_staging` runs `migrate up` before the test (`run.sh` 2689–2692).
  - Once `00003` exists, `role_bindings_subject_fkey` references `app.users`, and PostgreSQL refuses that TRUNCATE (SQLSTATE `0A000`; §14 R-7) even with zero bindings. B-IDENTITY-STAGING would fail on its first run against a post-G005 tree.
  - The byte-identical statement in `resetIdentity` (`integration_test.go:134`) gets the same `CASCADE`. B-IDENTITY-CATALOG and B-IDENTITY-IT validate it locally on the same version-3 schema.
  - The G004 implementation review anticipated this: "Its future schema/foreign keys require checking the existing disposable staging test setup; this is a downstream reconciliation obligation, not a G004 defect or authorization for G005 schema work."
- **Exposure (new in r4):**
  - No G005 gating pipeline detects it. `identity-staging` runs only for `web` and `schedule` pipelines on the protected `identity-qualification` branch, and B-IDENTITY-STAGING is in no local G005 suite.
  - Integrating G005 into `staging` breaks no automatic pipeline either. The protected branch stays at `4a46da9` (pre-G005) until it is updated, and a protected schedule keeps testing that commit. The failure would surface at the first identity qualification of a post-G005 tree.
  - The local CAP-3 block (IDENTITY.md `identity:staging-local`) runs against the **root** Compose database. It already truncates every identity table there. With `CASCADE` it also empties `role_bindings`, since bindings cannot outlive their users, but leaves Organizations and Groups. TENANCY.md records this (§7); IDENTITY.md is not edited (R4-Q3).
- **Why it is not avoidable inside G005** (unchanged). Spec R2 requires the subject FK, and any TRUNCATE of a referenced table needs `CASCADE` or co-truncation. Rejected alternatives:
  - dropping or weakening the FK, which violates R2;
  - limiting the staging test to version 2, which changes G004 semantics;
  - naming `app.role_bindings` in the identity statement, which couples an identity fixture to G005's table list and breaks again for any later table that references users.
- **Minimal fix** (unchanged): append `CASCADE` to that one statement (§6.2 (d)). No identity semantics, guard, route or secret changes.
- **Validation options** (none performed or authorized here):
  - **A (recommended): one protected qualification pipeline on the final candidate (X4),** following G004's accepted precedent (pipeline 2928428140).
    - After S7: an authorized fast-forward push to `identity-qualification`, and to the evidence branch with `ci.skip` (X1); the owner then starts the `web` pipeline in the GitLab UI.
    - It runs `fast`, `gate` (full) and `identity-staging` on one commit, so it also supplies X2's mandatory full evidence. X2 is not run separately and no full run is duplicated.
    - Extra cost over X2: the staging job (142 s in G004's run) and per-run WorkOS staging users, which the test deletes.
    - Needs owner authorization for the protected-branch update and the provider run.
  - **A′: the local staging suite (X5).** `W(staging)` with the owner's 0600 `IDENTITY_STAGING_ENV_FILE` on the workstation (CAP-3 precedent). No hosted action or protected-branch update, but it is a real-provider action with owner credentials, and X2 still runs.
  - **B: inspection plus local equivalence, with a deferred real-provider run.**
    - Codex accepts (d) by inspection: the identical statement is validated by (a) on the same schema. B-IDENTITY-STAGING reruns at the next authorized identity qualification, and the Result records the residual risk (the first post-G005 staging run).
    - B is consistent with Spec R9 ("model fixtures require no additional live provider/vendor") and with the BACKLOG direction to reserve real-provider runs for provider-affecting changes. The residual is a test-fixture regression, not a product or security defect.
  - **C: a G004-owned follow-up before G005 integration.** Change the staging fixture so it does not truncate FK-referenced identity tables. That is separate G004 scope, not a G005 change. Not recommended: it delays G005 for no stronger guarantee than A.
- **Recommendation: A, with B as the fallback** if the owner declines the protected-branch update or the provider run for G005. A gives final-tree evidence for every required ID, including B-IDENTITY-STAGING, at almost no cost beyond the mandatory full run. No option changes WorkOS configuration, secrets, settings, variables or branch protection.

### 13.2 r4 decisions and questions for Codex

| ID | Decision or question | Where |
|---|---|---|
| R4-D1 | The observation basis is the accepted G004 code [G004-ACC] (`7dc4c5c`), seen in the provisional `1505440`. E2/E3 are re-scoped to the open MR !7 and the integrated staging commit. | Header; §2.1–§2.3 |
| R4-D2 | CI facts and the expected profile: full in every gating context; `identity-staging` never runs for G005; no CI change beyond r3's one row and one control. | §2.1; §6.3 |
| R4-D3 | G049: concrete overlaps O1–O8 and the five-point sequencing/rebase rule; no blanket dependency. | §2.4 |
| R4-D4 | Precision: S7 runs `W(all)` (81 full-profile IDs) plus `W(unit)`; 83 = 81 + B-UNIT-CHECKS + B-IDENTITY-STAGING. X1 uses `ci.skip`. | §6.1; §6.4; §10; §12 |
| R4-Q1 | The ESC-1 disposition: A (recommended), A′, B or C. | §13.1 |
| R4-Q2 | Under A: confirm that X4 replaces X2 (the consolidated G004-style pipeline), with X1 as the evidence-branch push with `ci.skip`. | §12 |
| R4-Q3 | Confirm that the CAP-3 root-database effect of `CASCADE` is recorded in TENANCY.md only, without adding `docs/foundation/IDENTITY.md` to the allow-list. | §7; §13.1 |
| R3-Q1 | Unchanged: the optional `groups`/`business-profiles` reservation is not applied. | §13.3 |

There is no new escalation beyond ESC-1.

### 13.3 r3 decisions recorded (owner direction applied; carried into r4 unchanged)

| ID | Decision | Where |
|---|---|---|
| R3-D1 | `CreateGroup(org, creator, name, slug)` returns the Group and its `group_owner` Binding atomically. No owner column, no Organization membership, no actor-based checks; owner bindings are ordinary RoleBindings for the internal primitives. | DS-9, §4.2, §4.5, IT-OWN-1..6 |
| R3-D2 | Role vocabulary: `org_owner`/`org_admin`/`org_member`, `group_owner`/`group_admin`/`group_member`, source `membership`; six valid combinations. | §3.2, §4.3, IT-VOC-1 |
| R3-D3 | BusinessProfile terminology; profile access derived in G006/G007; no G005 profile schema, role, scope or method. G006 needs no change to `role_bindings`. | DS-2, §4.7 |
| R3-D4 | Advisor/Principal/Associate are rejected as RoleBinding roles; no staffing schema. | DS-2, IT-VOC-1 |
| R3-D5 | The G004 pattern is adopted: owner/runtime preflight DO block and exact-ACL postcondition, extended to column ACLs. | §3.2 |
| R3-D6 | The CT-6 owner-bearing column rule (A1/N9) protects "no duplicate owner column". | §5.2 |

**R3-Q1 (optional proposal, not applied).** ADR-004's future Groups and Business Profiles sibling management areas may later occupy top-level route segments such as `groups` or `business-profiles`. Reserving them now would avoid a later data migration. Per Q4 any substantive change to the 24-entry set needs Codex confirmation, so r3 leaves the set unchanged unless Codex directs otherwise.

### 13.4 Earlier resolutions (r1 review, unchanged)

All of r1's Q1–Q10 are resolved by the r1 review (not included). r2 applies them without redesign; no open question remains.

| ID | Resolution (r1 review) | Where applied in r2 |
|---|---|---|
| Q1 | Defer RLS to G008 exactly as the Spec requires. Add only a scoped foundation-stage explanation and the mandatory G008 inventory in MIGRATIONS.md and TENANCY.md. Do not weaken the eventual ENABLE/FORCE requirement or expose tenant data before its gates. | DS-7; §7 docs |
| Q2 | Include `organization_id` in GroupMembership's read-only projection as its tenant anchor. No second authority or tenant entity. | §3.2 view; CT-6 A2 column list |
| Q3 | One `core/internal/tenancy` package accepted. | §4.1 |
| Q4 | The 24-entry reserved set is accepted provisionally. Re-inventory the actual accepted routes at E3 and return substantive changes for confirmation. | §4.4; P-G004-11 |
| Q5 | Tenancy checks in the existing `migrations` suite plus the one foundation-path row and control accepted. | §6.3 |
| Q6 | Local F-1 plus the existing gate-failure controls suffice. One passing hosted full run on the final tree remains mandatory. X3 is not required. | §5.4; E6; §12 |
| Q7 | Disabled users may retain model bindings. G004 denies disabled authentication and G007 owns effective policy. | §4 (no lifecycle policy) |
| Q8 | PostgreSQL failing-row DETAIL acknowledged. Service rejection and parameterized queries preserve the canary guarantees; direct constraint cases use non-sensitive fixtures; all captured logs remain scanned. No infrastructure change now. | §5.3 |
| Q9 | SECURITY.md is empty and is not security evidence. Apply the named architecture, accepted-ADR and Spec invariants; no certification and no new security-document workstream. | Header; §2.1 |
| Q10 | Keep the distinct B-TENANCY-HTTP ID. | §6.1 |
| Test scope | Production linkage and import checks are scoped separately from explicitly test-only middleware and SQL fixtures; test helpers never enter production binaries. | §5.1 U-BOUNDARY, B-MCP-SOURCE |

**Accepted design choices (review pins; unchanged from r1/r2 except the r3 items above):**
- DS-1..8, plus DS-9 (r3);
- the §4.2 surface (r3: `CreateGroup` signature and `GroupRole` codes);
- name rules (§4.3);
- slug algorithm, base-only create and the suffix range 2–99 (§4.4);
- the error-code set and mapping;
- last-committed-wins with no retries (§4.5);
- logs without UUIDs (§4.6);
- the 24-entry reserved set;
- `UNIQUE` across all sources;
- column grants plus triggers;
- `security_invoker` views.

**Context and scope:** G005 fits one Goal within B7. **No split is recommended.** The G004 harness interplay is concretely mapped from the accepted G004 code (§6.2); E3 re-confirms it at the integrated staging commit.

## 14. Research record (read-only, 2026-10-07; r3 additions 2026-10-08; r4 additions 2026-10-08/09 UTC)

- **r4 sources (read-only):**
  - the repository inputs bound in `G005-planning-inputs-r4.json` at `1505440`;
  - `git diff --stat` of `4a46da9..7dc4c5c`, `7dc4c5c..1505440`, `c141235..1505440` and `19aa3c2..c141235`; the combined merge diff of `.gitlab-ci.yml`, `run.sh` and `validate-controls.sh`; `git ls-tree` comparisons of r3's 15 observed G004 paths;
  - PostgreSQL TRUNCATE semantics (R-7).

  No other checkout was read, and the G049 Plan was not read. r4 wrote four files: `G005-organization-group-memberships-plan-r2.md` (the exact r2 archive), `G005-organization-group-memberships-plan-r3.md` (the exact r3 import), this Plan and `G005-planning-inputs-r4.json`.
- **r3 sources (read-only):**
  - the clone inputs bound in `../g005-planning-inputs-r3.json`;
  - G004 candidate `4a46da9`, read only through object queries in the G004 checkout (file views, greps, diffs against base `201f5d1`, blob IDs). Files: `core/migrations/00002_identity.sql`, `core/internal/migrate/{integration,runner}_test.go`, `core/migrations/migrations_test.go`, `core/internal/identity/{integration_test,store_test,identity,http}.go`, `core/internal/identity/workos/staging_test.go`, `infra/validation/{run.sh,required-checks.tsv}`, `infra/validation/tools/internal/logrecord/events.json`, `infra/ci/{lib.sh,validate-controls.sh}`, `web/app/**` and `web/public/**` listings, and the interim Result.

  No G004 file, evidence, binding or resource was modified or executed.
- **Local sources (r1/r2):** every governing input in the header table; G004 Plan r1/r2 (§5–§7, §10) for the deferred schema and harness details; the clone's `core/`, `infra/postgres`, `infra/validation`, `infra/ci`, `.gitlab-ci.yml`, `docs/foundation`, the archetype tenancy contract and its isolation fixture; G006 Spec (extension expectations); the G048 Plan (public asset paths).
- **PostgreSQL 17 documentation:**
  - **R-1** [CREATE TRIGGER](https://www.postgresql.org/docs/17/sql-createtrigger.html) and [Privileges](https://www.postgresql.org/docs/17/ddl-priv.html): EXECUTE is required to *create* a trigger. Whether it is checked when a trigger fires is not documented; hence the S2 runtime proof and the §11 contingency. The Privileges page also documents column-level INSERT/UPDATE and that PUBLIC receives no table privileges by default.
  - **R-2** [CREATE TABLE](https://www.postgresql.org/docs/17/sql-createtable.html): CHECK constraints are tested in alphabetical order by name after NOT NULL; TRUE or UNKNOWN passes.
  - **R-3** [CREATE VIEW](https://www.postgresql.org/docs/17/sql-createview.html): `security_invoker` applies the invoker's privileges and RLS policies. Simple views are auto-updatable, and a write needs the INSERT/UPDATE/DELETE privilege on the view itself.
  - **R-4** [Pattern matching](https://www.postgresql.org/docs/17/functions-matching.html): ARE escapes, including `\uwxyz`, are allowed inside bracket expressions. Ranges follow the collating sequence, hence the `COLLATE "C"` pins.
  - **R-5** Image `pgvector/pgvector:0.8.7-pg17` (`infra/compose/local.yml`): PostgreSQL 17, built-in `gen_random_uuid()`, `security_invoker` available.
  - **R-6** [Error reporting and logging](https://www.postgresql.org/docs/17/runtime-config-logging.html): `log_parameter_max_length_on_error` defaults to 0, so bind parameters are not included in error messages.
  - **R-7** [TRUNCATE](https://www.postgresql.org/docs/17/sql-truncate.html): TRUNCATE cannot be used on a table that has foreign-key references from other tables, unless all such tables are also truncated in the same command; `CASCADE` truncates every referencing table. The refusal raises "cannot truncate a table referenced in a foreign key constraint" (SQLSTATE `0A000`) and does not depend on whether the referencing table has rows.
