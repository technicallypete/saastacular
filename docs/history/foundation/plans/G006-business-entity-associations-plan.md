> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G006 — BusinessProfile and Group associations: implementation plan (r2, reconciled to merged G005)

- **Plan revision:** r2 DRAFT, 2026-10-09 UTC. Supersedes Plan r1 (not included)
  (archived byte-for-byte, SHA-256 `04a1a4d4ba805bf55eaa452b57c6d1d383574411d806b8c015cb67e6dfbf2416`) and carries
  forward its Codex review (not included) (provisional design ACCEPT, 2026-10-07) where
  still applicable (§0.2).
  - Author: Claude Code (planning-only child, no subagents). Reviewer: Codex. Major decisions: product owner.
- **Status: PLAN DRAFT submitted for Codex review. Not implementation-ready until the gates in §2.3 pass.**
  - G005 is **merged** into staging (MR !10, `570b40c`) but **not formally Codex-ACCEPTed**. That ACCEPT is an
    execution prerequisite (E1), not a planning blocker.
  - Nothing was implemented, built, tested, migrated, run in Docker, pushed or merged. No provider, GitLab write, CI,
    pipeline or settings action. **Zero of five B5 cycles consumed.**
- **Governing inputs:** [Goal r3](../goals/G006-business-entity-associations.md) and
  [Spec r3](../specs/G006-business-entity-associations-spec.md) (history: `-r1`, `-r2`, `-spec-r1`, `-spec-r2`);
  [cardinality addendum](../specs/G006-business-profile-cardinality-addendum.md);
  [ADR-004](../../../adr/ADR-004-group-business-profile-cardinality.md); ARCHITECTURE §53.4–53.7
  (Canonical Multi-Tenant Semantics override).
- **Naming rule location.** No separate naming-addendum file exists at `570b40c`. The rule (canonical `BusinessProfile`,
  UI **Business Profile**, `GroupBusinessProfile`, `business_profile_id`, `business_profile.*`; no `GroupProfile`;
  Groups and Business Profiles are sibling management areas) is in:
  - the "Owner naming clarification — 2026-10-08" sections of the G006 Goal and Spec;
  - the "Naming handoff" and "UX handoff clarification" sections of the cardinality addendum;
  - ADR-004's "Owner-approved naming clarification" and "sibling management UX" sections;
  - ARCHITECTURE §53.4.

  Stable file names (`G006-business-entity-associations-*`) are unchanged.
- **Planning provenance:** isolated worktree `.claude/worktrees/agent-<id>`, detached at
  `origin/staging` `570b40caf47ff9257c5b5aa0588f60fecd58d0a6` (tree `a09895950c8077f7f50b116d05edeb76c81c657e`, equal to
  the MR !10 source `f862f18`'s tree), clean. Every input's SHA-256 at that commit, plus the external final G005 Result
  blob, is in G006-planning-inputs.json (not included). This is planning provenance, not an execution
  binding.

---

## 0. Revision record

### 0.1 What changed from r1, and why

| # | Area | r1 | r2 | Why |
|---|---|---|---|---|
| 1 | Naming | `BusinessEntity`, `GroupBusinessEntity`, `business_entity_id`, `be_*` | `BusinessProfile`, `GroupBusinessProfile`, `business_profile_id`; tables `app.business_profiles`, `app.group_business_profiles`; migration `00004_business_profiles.sql`; IDs `B-BUSINESS-PROFILE-*` | Owner naming clarification (Goal/Spec/addendum/ADR-004, 2026-10-08) |
| 2 | Identity and ownership | Organization-owned UUID entity | Unchanged in substance: canonical immutable UUID, `organization_id` NOT NULL, no Group parent | Spec R1; ADR-004 |
| 3 | Cardinality | Many-to-many; pair uniqueness `(organization_id, group_id, business_entity_id)` | **At most one profile per Group:** `UNIQUE (organization_id, group_id)`, plus a service precheck. One profile may be shared by many Groups of the same Organization; no uniqueness on `business_profile_id` alone | ADR-004; addendum; Spec R3 |
| 4 | Authority | `business_entity` scope added to `role_bindings` (new column, FK, three-branch CHECK), `be_owner`/`be_admin` direct grants, `business_entity_memberships` projection | **None of these.** `role_bindings` is not altered. No BusinessProfile scope, role, source, owner column, grant method or writable mirror | Spec r3 R4; ADR-004 additional decisions |
| 5 | Access facts | Direct-grant projection | **Derived, read-only** `security_invoker` view `app.business_profile_access_paths(user_id, organization_id, business_profile_id, group_id, group_role)`, joining current Group membership bindings to the Group's single attachment. Organization authority stays in the existing `organization_memberships` projection | Spec R4/AC6/AC7; ARCHITECTURE §53.7 |
| 6 | Service surface | 13 new methods (r1 review clarification 1), 33 total | **9 new methods, 29 total** (§4.2). No grant/change/remove methods | Grants removed |
| 7 | Predecessor changes | Altered `role_bindings` (DS-BE-5); Down deleted direct grants | **Purely additive.** No predecessor object changes. Down drops only G006 objects and data; every G005 row, binding and slug survives | Simpler model |
| 8 | Ownership boundaries | G018 noted | Explicit: G018 owns exactly-one-at-completion, onboarding state and its UI. G007 owns capability policy and management classification. G008 owns RLS. Out of G006 (§4.7) | Addendum; ADR-004 |
| 9 | Predecessor facts | Provisional `[G005-PROV]` register P-G005-1..16 | Observed facts of merged G005 at `570b40c` (§2.1). r1's "TestTenancyModel needs no change" was wrong: its `setup()` pins `current_version 3`. "UpTo(3)" does not exist; the existing `subsetFS`/`runnerFor` helpers pin versions | Actual merged code |
| 10 | Migration number | `00004` [provisional] | `00004_business_profiles.sql`, confirmed: the set at `570b40c` is `00001`–`00003` and nothing newer exists | Observed |
| 11 | Authority detector | Inventory adds BusinessEntityMembership (r1 review clarification 2) | The inventory adds the access-path view, with an **explicit expected-source set per projection** (§5.2 CT-BP-6) | The view joins two relations |
| 12 | Import deny-list | Forbid `net` (r1 §4.1) | `net` stays allowed: G005's `errors.go` imports it to classify errors. Forbid `crypto/tls`, `os/exec`, `html/template`, `text/template`, plus the existing `net/http*` | Actual G005 code |
| 13 | Website CHECK | `'^https?://[!-[]-~]+$'` | `'^https?://[!-[]-~]+$'` (printable ASCII without `\`), in G005's `\u`-escape style | r1's bracket expression would not parse as intended |
| 14 | Persist phase | Shares G005's PostgreSQL restart | G006 runs its own seed, stop/start and verify after the G005 steps. G005's harness function is unchanged | Less coupling; one extra restart |
| 15 | Validation and hosted order | X1 evidence-branch push and X2 manual full `web` pipeline; local `validate.sh --profile full` parity | Current owner policy (§7.5): rebase onto the remote target **before** final local validation. Required local full coverage of the final reconciled code comes **before** the MR. The MR `fast` job runs automatically; the MR `gate` is manual, optional and not started. The automatic staging gate follows an owner-authorized conditional integration and is judged by the owner status rule | AGENTS.md 2026-10-09 policies; `070882c`, `b803db2`, `8f5f453` |
| 16 | CI facts | Provisional | G050 JUnit report derives from the registry, so it needs no G006 change. `070882c` made the MR gate manual/optional. G006 paths select `full` through existing rows. **No CI mapping change** | Observed |
| 17 | Skill paths | `skills/project-archetype/…` | `skills/saas-foundation/references/tenancy-authorization.md` (G049 relocation). Its many-to-many and `be_*` examples are historical; ADR-004 supersedes them. No Skill edit | Observed |

**Unchanged:** budget **B5, checkpoint after cycle 3**; Spec AC1–AC12; the r1 field contract (§4.3) and website
contract (§4.4); full-replacement last-commit-wins updates; the no-fetch guarantee; the Unicode/whitespace pins;
package placement in `core/internal/tenancy`; one log event; no public surface; scope; and the G021 barrier.

### 0.2 Disposition of the r1 review

| r1 item | r2 disposition |
|---|---|
| Q1 tenant-anchored BusinessEntityMembership | **Superseded** by the derived access-path view (row 5). The `organization_id` tenant anchor is kept |
| Q2 conflict, not idempotent, on a duplicate link | **Kept and re-pinned.** `association_exists` now means "this Group already has a profile association", whether the profile is the same or different (§4.5) |
| Q3–Q6 | **Kept** (full-replacement updates; Unicode pins; website contract; one package) |
| Q7 typed `role_bindings` extension | **Withdrawn**: no predecessor change |
| Q8 Down deletes direct grants | **Withdrawn**: Down removes G006 objects and data only |
| Q9 predecessor tests pinned at version 3 | **Kept**, with the actual mechanism (§7.2) |
| Q10 three IDs, existing event and probe list | **Kept** (renamed) |
| Q11 forced failure | **Kept**: local F-1 plus inherited gate controls. The hosted part moves to the staging gate (§7.5) |
| Q12 no automatic grants, no membership prerequisite | **Strengthened**: there are no profile grants at all |
| Clarification 1: exact Service surface | Re-pinned to **29** names (§4.2), with the same evidence: the AST check accepts the set and rejects an extra exported method |
| Clarification 2: version-aware detector | Kept. Version 3 uses the two G005 projections. Version 4 uses those two plus `business_profile_access_paths`, with per-projection expected sources. The original meanings are unchanged |
| Clarification 3: website/field validation evidence | Kept unchanged: SQL CHECKs are defense in depth, not the parser contract |

## 1. Approach summary

1. **One forward, purely additive migration**, `core/migrations/00004_business_profiles.sql` (§3), adds:
   - `app.business_profiles`: Organization-owned, UUID-identified, five bounded context fields, no Group parent;
   - `app.group_business_profiles`: UUID-identified links with composite same-Organization foreign keys to both
     endpoints and `UNIQUE (organization_id, group_id)`;
   - the read-only `security_invoker` view `app.business_profile_access_paths`;
   - two key-immutability triggers and exact column-level runtime grants, with G005's preflight and ACL postcondition
     pattern.

   It alters no G001/G004/G005 object.
2. **Access is derived.** `role_bindings` stays the only role store. A profile's access facts are the current Group
   membership bindings of the Groups attached to it. Linking, unlinking or editing a profile writes no binding, and a
   Group role change writes no profile row. G007 classifies roles into capabilities; G008 enforces RLS.
3. **Internal primitives only.** `core/internal/tenancy` gains 9 Service methods in three new files. It stays linked
   into neither binary. No HTTP, BFF, MCP, CLI, environment variable or bootstrap.
4. **Evidence.** Unit tests run in `core-check`. Real-PostgreSQL catalog, migration, model, concurrency and fault tests
   run as the actual runtime role in the existing `migrations` project. Also: the extended route-absence probe,
   `logcheck`, the F-1 forced failure, final local full coverage of the rebased candidate, the automatic MR `fast`
   job and the automatic staging gate after conditional integration.
5. **Bounded harness reuse.**
   - Three new required IDs in `migrations`.
   - Mechanical version-4 reconciliation of the version sites and predecessor tests.
   - Enum additions to one existing log event; an extended probe list.
   - No new suite, profile, job, port set, wrapper, framework or CI mapping change.

## 2. Baseline, observed predecessor facts and gates

### 2.1 Observed facts at `570b40c` (merged G005; read-only)

| ID | Fact | G006 use |
|---|---|---|
| F-1 | `app.groups` has `groups_organization_id_key UNIQUE (organization_id, id)`. `app.organizations(id)` is its PK. Both FKs are `ON DELETE RESTRICT` | Composite FK target. **No parent change needed** (Spec R7) |
| F-2 | `app.role_bindings` holds the Group scope as `scope_type='group'`, `group_id = scope_id`, roles `group_owner\|group_admin\|group_member`, source `membership`. Partial index `role_bindings_group_idx (organization_id, group_id)`. Closed two-branch CHECK `role_bindings_scope_consistent`. `CreateGroup` atomically writes the creator's `group_owner` binding | The view joins on `(organization_id, group_id)`. **Unchanged** |
| F-3 | Projections `organization_memberships` and `group_memberships` are `security_invoker`, SELECT-only. The runtime grants come from a `DO` block deriving `current_database()||'_runtime'`, followed by an exact table/column ACL postcondition and a "no other grantee" check. A preflight asserts the schema owner and that the runtime role exists | Same pattern |
| F-4 | Trigger functions use `SET search_path = pg_catalog`, are not SECURITY DEFINER and raise 23000 with a named constraint. Runtime has no EXECUTE (CT-2: owner only) | Same pattern |
| F-5 | Package `core/internal/tenancy` has 7 production files and 20 pinned Service methods (`boundary_test.go`). Its 20 SQL constants all live in `store.go` (8 writes), and no SQL is allowed elsewhere. Imports allow-list: pgx, pgconn, pgtype, `core/internal/observability`; `net/http` is forbidden | Extended (§4, RC-11) |
| F-6 | Helpers `inTx`/`mutate`/`query`, `exists`, `one`, `collect` and `uuidArg`. Each operation is one READ COMMITTED transaction, with success only after commit, a deferred rollback and no retries. `classify`/`pgCode` map constraint names to codes; anything else maps to `internal` | Reused unchanged |
| F-7 | Log event `tenancy.operation`. Its `events.json` rule **requires** `scope_type` for `membership.*` and **forbids** it for other operations | Profile operations log without `scope_type`. No new scope value |
| F-8 | Tests (tenancy package): `integration_test.go` has `setup()` (asserts `current_version\t3`), the fixtures `e.user`/`e.org`/`e.group`/`e.rows`, the `faultDB` fault wrapper (it records no statements), `waitBlocked` and `h` | `setup()` gets a mechanical pin change (RC-6); the helpers are reused |
| F-9 | Tests (migrate package): `tenancy_integration_test.go` has `catalogConn`, `runtimeConn`, `runtimePool`, `subsetFS`, `runnerFor`, `tenancyCatalog`, `authorityViolations` with `pinnedProjections`, `seedIdentity` and `identitySnapshot`. `TestTenancyCatalog` (21 subtests) and `TestTenancyMigration` (6 subtests) run at "latest" through `newRunner`. The persist seed/verify phases assert version 3. `integration_test.go` has `reset()` with `tenancyDrops`, `TestMigrationLifecycle` (10 subtests, version 3 hard-coded) and `metadataIdentityAndTenancyOnly`. `runner_test.go` pins the normal set | Reconciled (§7.2) |
| F-10 | `core/internal/identity/store_test.go` `TestIdentityCatalog` pins the exact `app` table list. The identity fixtures' `TRUNCATE … CASCADE` reaches only `role_bindings`, because it references `users` | Table list +2 (RC-7). `CASCADE` is unaffected by G006 (MG-BP-7) |
| F-11 | `run.sh` version-3 sites: B-MIGRATE-SET (+ 00001/00002 digests), B-DOCS-EXEC-ROOT migrate block, `suite_migrations` `mig_step`s (27/14 objects), B-MIGRATE-CATALOG, `identity_catalog` (`TENANCY_OBJECTS`), B-PERSIST, B-DBRENAME, the off variant (B-TOGGLE-OFF), `TENANCY_PROBE_CORE/WEB` (36 probes). `APPLICABLE` already lists `core/migrations`, `core/internal/tenancy` and `docs/foundation/TENANCY.md` | §7.2 |
| F-12 | Registry: 83 rows; fast 33, integration 71, full 81, staging 5. `infra/ci/lib.sh`: full = lint controls provenance contracts root migrations ops rename off; `unit` is in fast and integration | +3 `migrations` rows → 86; full 84 |
| F-13 | CI: row `foundation dir core/internal/tenancy/` and control `F-tenancy` exist. G050's `junit.sh` builds reports from `required-checks.tsv` (ID regex `^[A-Z][A-Z0-9-]*$`). `.gitlab-ci.yml` (`070882c`) makes the MR `gate` `when: manual, allow_failure: true`; staging and default-branch pushes run `gate` automatically, selected by paths | No mapping change (§7.4) |
| F-14 | Final G005 Result (local commit `0e32d93`, unpushed): MR `fast` pipeline 2931660933 passed 33/33; MR gate not started; staging gate **pending**. B7 used 1 cycle. Residuals: ESC-1 option B (no post-G005 live staging run); the PostgreSQL DETAIL lines carry UUIDs/slugs; whether `infra/ci/junit.sh` joins the binding is open | E1/E4 inputs (§2.3) |
| F-15 | `docs/foundation/OPERATIONS.md`'s profile table still shows integration 70 and full 77, against the registry's 71 and 81 | Observation only. Not G006 scope (Q-r2-11) |

**Overlap with G051 (Plan r1 at `19dc50e7`, under CORRECTION LOOP).** G051 proposes relabelling and splitting
`run.sh` suites, adding 5 registry rows, and changing `profile-paths.tsv` and `validate-controls.sh`. This is not a
blanket dependency. Whichever Goal integrates second rebases its `run.sh`/`required-checks.tsv` hunks mechanically at
E2/S6. If a G051 change alters `migrations` suite semantics, rather than labels, counts or locations, G006 returns to
Codex.

### 2.2 Reconciliation rule (E2)

This r2 is pre-reconciled at `570b40c`. Before execution, compare every bound input in `G006-planning-inputs.json`
with the execution baseline:
- **Equal:** nothing to do.
- **Mechanically different** (line numbers, counts, a target's unrelated hunks, G051 relabels): re-derive and record
  in the Result.
- **Semantically different** (a G005 correction that changes the tenancy schema, Service, test helpers or harness
  wiring, or any `role_bindings`/projection change): return to Codex with a reissued Plan or a reconciliation record.

### 2.3 Gates

| Gate | Requirement | Established by | State |
|---|---|---|---|
| E1 | **G005 implementation ACCEPT** (Goal dependency). It needs a qualifying post-merge staging disposition under the owner status rule and Codex's review of the final Result | Codex | **Pending.** Merged at `570b40c`; staging gate not observed by this planning child |
| E2 | Reconciliation against the actually integrated G005 model and checks at the execution baseline (§2.2), including any G005 correction-loop changes and any integrated G051 changes | Claude drafts; Codex confirms | Pre-reconciled at `570b40c` |
| E3 | Codex ACCEPT of this Plan r2, including the §13 decisions and B5 | Codex | Pending |
| E4 | Fresh committed baseline; supervisor approved-input snapshot and digest (`run-approved.sh` inputs, including the decision on `infra/ci/junit.sh`); exact scope allow-list; a fresh isolated worktree | Codex | Pending |
| E5 | Named external actions authorized (§12) | Owner | Pending |
| S0 | Start gate (Claude; read-only; no cycle). <ul><li>E1–E5 recorded;</li><li>wrapper `--check-only` passes;</li><li>HEAD equals the bound base; status clean; root `.env` absent;</li><li>`PORTS_MIG/OPS/REN/OFF/IDN` free;</li><li>no conflicting Docker resources, and the shared `local-harness.lock` is free;</li><li>GitLab compute-minute usage checked read-only before X1.</li></ul> | Claude | Pending |

## 3. Schema: `core/migrations/00004_business_profiles.sql` (low–medium risk, additive)

### 3.1 Decisions [Plan pins]

- **DS-BP-1 Keys.** Keys are `gen_random_uuid()` defaults, with nonzero CHECKs. Runtime cannot supply `id` or
  `created_at` (column-level INSERT). Names and websites are never keys. Equal values are allowed within and across
  Organizations, and nothing is merged by name, email, website or provider claim.
- **DS-BP-2 Ownership.** `business_profiles.organization_id` is NOT NULL, references `organizations` with RESTRICT, and
  is immutable. There is no `group_id`, Group-owner column or per-Group copy. A business present in two
  Organizations is two rows.
- **DS-BP-3 Composite alignment.**
  - `business_profiles_organization_id_key UNIQUE (organization_id, id)` is the composite FK target. It is also the
    target for future BillingProfile children.
  - The link carries `organization_id`, `group_id` and `business_profile_id`, all NOT NULL.
  - Its FKs are `(organization_id, group_id) → groups (organization_id, id)` and
    `(organization_id, business_profile_id) → business_profiles (organization_id, id)`, plus the
    `organization_id → organizations` FK (G005 `role_bindings` parity), all RESTRICT.
- **DS-BP-4 Cardinality.** `group_business_profiles_group_key UNIQUE (organization_id, group_id)`: at most one link per
  Group. It also serves as the Group-FK index and the singleton lookup. `business_profile_id` alone is not unique.
  The `group_business_profiles_business_profile_idx` index serves profile-side FK checks and lists.
- **DS-BP-5 Immutability.**
  - First line: column grants. Runtime may UPDATE only the five context columns, and has no UPDATE on links.
  - Second line: BEFORE UPDATE triggers raising 23000 with a named constraint, on `business_profiles (id,
    organization_id)` and on `group_business_profiles (id, organization_id, group_id, business_profile_id)`.
- **DS-BP-6 Deletion.** All FKs are RESTRICT. Runtime has no DELETE on `business_profiles`. Unlink is a runtime `DELETE`
  of one link row, which nothing references. There is no cascade, transfer, merge, reparent or replace operation.
- **DS-BP-7 Derived access view.** `app.business_profile_access_paths` (definition in §3.2). It is `security_invoker`,
  so G008 RLS on both base tables applies to readers, and it is SELECT-only.
  - **Group paths only**, with the raw `group_role` value. It has no capability, "manages" or management column (G007).
  - Organization authority is read from the existing `organization_memberships`, not copied here.
  - A profile without attachments has no rows.
- **DS-BP-8 Context columns** (r1 DS-BE-9, unchanged): `display_name` NOT NULL; `legal_name`, `description`,
  `industry` and `website` nullable (NULL means absent; empty strings are never stored). Only `created_at` metadata.
  No `updated_at`, slug, hostname, owner/creator, billing, payment, tax, onboarding/completion or jsonb column.
- **DS-BP-9 No RLS in G006** (G005 parity). TENANCY.md records the two tables and the view as G008 obligations.

### 3.2 Proposed DDL (exact intent; whitespace and comments may differ)

```sql
-- G006: Organization-owned BusinessProfile context, one-profile-per-Group same-Organization links and read-only access
-- facts derived from the single RoleBinding store. No role, scope or grant is added; role_bindings is not altered.
-- RLS/policy are G007/G008; onboarding completion is G018. Down drops only these objects and their data.

-- +goose Up
-- +goose StatementBegin
DO $$ BEGIN  -- preflight, identical in intent to 00003
  IF current_user::text IS DISTINCT FROM (SELECT nspowner::regrole::text FROM pg_namespace WHERE nspname = 'app') THEN
    RAISE EXCEPTION 'migrations must run as the owner of schema app, not %', current_user; END IF;
  IF to_regrole(current_database() || '_runtime') IS NULL THEN
    RAISE EXCEPTION 'runtime role for database % is missing', current_database(); END IF;
END $$;
-- +goose StatementEnd

CREATE TABLE app.business_profiles (
  id uuid NOT NULL DEFAULT gen_random_uuid(), organization_id uuid NOT NULL,
  display_name text NOT NULL, legal_name text, description text, industry text, website text,
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT business_profiles_pkey PRIMARY KEY (id),
  CONSTRAINT business_profiles_organization_fkey FOREIGN KEY (organization_id)
    REFERENCES app.organizations (id) ON DELETE RESTRICT,
  CONSTRAINT business_profiles_organization_id_key UNIQUE (organization_id, id),
  CONSTRAINT business_profiles_id_nonzero CHECK (id <> '00000000-0000-0000-0000-000000000000'),
  CONSTRAINT business_profiles_display_name_valid CHECK (char_length(display_name) BETWEEN 1 AND 255
    AND display_name = btrim(display_name) AND (display_name COLLATE "C") !~ '[\u0001-\u001f\u007f-\u009f]'),
  CONSTRAINT business_profiles_legal_name_valid CHECK (legal_name IS NULL OR (char_length(legal_name) BETWEEN 1 AND 255
    AND legal_name = btrim(legal_name) AND (legal_name COLLATE "C") !~ '[\u0001-\u001f\u007f-\u009f]')),
  CONSTRAINT business_profiles_description_valid CHECK (description IS NULL OR (char_length(description) BETWEEN 1 AND 4000
    AND description = btrim(description) AND (description COLLATE "C") !~ '[\u0001-\u0008\u000b-\u001f\u007f-\u009f]')),
  CONSTRAINT business_profiles_industry_valid CHECK (industry IS NULL OR (char_length(industry) BETWEEN 1 AND 255
    AND industry = btrim(industry) AND (industry COLLATE "C") !~ '[\u0001-\u001f\u007f-\u009f]')),
  CONSTRAINT business_profiles_website_valid CHECK (website IS NULL OR (char_length(website) BETWEEN 1 AND 2048
    AND (website COLLATE "C") ~ '^https?://[!-[]-~]+$'))
);

CREATE TABLE app.group_business_profiles (
  id uuid NOT NULL DEFAULT gen_random_uuid(), organization_id uuid NOT NULL, group_id uuid NOT NULL,
  business_profile_id uuid NOT NULL, created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT group_business_profiles_pkey PRIMARY KEY (id),
  CONSTRAINT group_business_profiles_organization_fkey FOREIGN KEY (organization_id)
    REFERENCES app.organizations (id) ON DELETE RESTRICT,
  CONSTRAINT group_business_profiles_group_fkey FOREIGN KEY (organization_id, group_id)
    REFERENCES app.groups (organization_id, id) ON DELETE RESTRICT,
  CONSTRAINT group_business_profiles_business_profile_fkey FOREIGN KEY (organization_id, business_profile_id)
    REFERENCES app.business_profiles (organization_id, id) ON DELETE RESTRICT,
  CONSTRAINT group_business_profiles_group_key UNIQUE (organization_id, group_id),
  CONSTRAINT group_business_profiles_id_nonzero CHECK (id <> '00000000-0000-0000-0000-000000000000')
);
CREATE INDEX group_business_profiles_business_profile_idx
  ON app.group_business_profiles (organization_id, business_profile_id);

-- business_profiles_keys_immutable(): (id, organization_id); group_business_profiles_keys_immutable():
-- (id, organization_id, group_id, business_profile_id). Each: LANGUAGE plpgsql SET search_path = pg_catalog,
-- RAISE ... USING ERRCODE = 'integrity_constraint_violation', CONSTRAINT = '<function name>'; BEFORE UPDATE FOR EACH ROW.

CREATE VIEW app.business_profile_access_paths WITH (security_invoker = true) AS
  SELECT rb.subject_user_id AS user_id, gbp.organization_id, gbp.business_profile_id, gbp.group_id,
         rb.role AS group_role
  FROM app.group_business_profiles gbp
  JOIN app.role_bindings rb ON rb.organization_id = gbp.organization_id AND rb.group_id = gbp.group_id
  WHERE rb.scope_type = 'group' AND rb.source = 'membership';

-- +goose StatementBegin
DO $$  -- grants plus the exact-ACL postcondition, as in 00003
DECLARE rt text := current_database() || '_runtime';
  rels text[] := ARRAY['business_profiles', 'group_business_profiles', 'business_profile_access_paths'];
  expected_tables text := 'business_profile_access_paths:SELECT;business_profiles:SELECT;group_business_profiles:DELETE,SELECT';
  expected_columns text := 'business_profiles.description:INSERT,UPDATE;business_profiles.display_name:INSERT,UPDATE;'
    || 'business_profiles.industry:INSERT,UPDATE;business_profiles.legal_name:INSERT,UPDATE;'
    || 'business_profiles.organization_id:INSERT;business_profiles.website:INSERT,UPDATE;'
    || 'group_business_profiles.business_profile_id:INSERT;group_business_profiles.group_id:INSERT;'
    || 'group_business_profiles.organization_id:INSERT';
BEGIN
  REVOKE ALL ON app.business_profiles, app.group_business_profiles, app.business_profile_access_paths FROM PUBLIC;
  EXECUTE format('GRANT SELECT ON app.business_profiles, app.group_business_profiles, app.business_profile_access_paths TO %I', rt);
  EXECUTE format('GRANT INSERT (organization_id, display_name, legal_name, description, industry, website), '
    || 'UPDATE (display_name, legal_name, description, industry, website) ON app.business_profiles TO %I', rt);
  EXECUTE format('GRANT INSERT (organization_id, group_id, business_profile_id), DELETE ON app.group_business_profiles TO %I', rt);
  -- then the 00003 postcondition queries over rels: exact runtime table/column ACLs; no grantee besides owner and runtime
END $$;
-- +goose StatementEnd

-- +goose Down
-- Destroys all BusinessProfile and link data (development/test only; refused with APP_ENV=production by the runner).
DROP VIEW app.business_profile_access_paths;
DROP TABLE app.group_business_profiles;
DROP TABLE app.business_profiles;
DROP FUNCTION app.group_business_profiles_keys_immutable();
DROP FUNCTION app.business_profiles_keys_immutable();
```

**Notes:**
- Every CHECK is NULL-safe. The ranges use `COLLATE "C"` and `\u` escapes, as in 00003.
- The database rule is never stricter than the service rule: `btrim` trims ASCII spaces only, and the website pattern
  is a subset check of the §4.4 parser contract, not a URL parser.
- No `CASCADE` anywhere, so an unexpected dependent fails loudly.
- FK checks run as the table owner, so runtime needs no REFERENCES privilege.

### 3.3 Privilege matrix (asserted by the migration postcondition and CT-BP-3)

| Object | `<db>_runtime` | retention, migrator (without SET ROLE), PUBLIC |
|---|---|---|
| `business_profiles` | SELECT; INSERT (organization_id, the five context columns); UPDATE (the five context columns) | nothing |
| `group_business_profiles` | SELECT; INSERT (organization_id, group_id, business_profile_id); DELETE | nothing |
| `business_profile_access_paths` | SELECT | nothing |
| the two trigger functions | no EXECUTE (owner only, as CT-2 of G005) | no EXECUTE |
| every G001/G004/G005 object | **unchanged** | unchanged |

Never granted: UPDATE on any link column; DELETE, TRUNCATE, REFERENCES or TRIGGER on `business_profiles`; INSERT of
`id`/`created_at`; UPDATE of `id`/`organization_id`; any write on the view; ownership, DDL or role membership.

### 3.4 Indexes

New relations: 2 tables + 1 view + 5 indexes = **8**. The indexes are the two primary keys,
`business_profiles_organization_id_key`, `group_business_profiles_group_key` and
`group_business_profiles_business_profile_idx`. Object counts beyond metadata go from 27 at version 3 to **35 at
version 4**. The view's join uses the existing `role_bindings_group_idx`. The per-user read uses
`role_bindings_subject_scope_key`.

### 3.5 Risk, rollback and irreversibility

- **Risk: low–medium.** The migration is purely additive: no predecessor table is rewritten, scanned for a new CHECK
  or re-granted. Adding FKs that reference `groups` and `organizations` takes brief SHARE ROW EXCLUSIVE locks on those
  parents. There is no production data.
- **Irreversible model decisions for review:**
  - table, column and view names;
  - UUID v4 keys and the composite-FK layout;
  - the Group-unique link and its conflict code;
  - DELETE-based unlink naming the profile;
  - the field bounds and the website contract;
  - the view's columns, including `group_role` and Group-paths-only.
- **Down** removes the G006 objects and **all profile/link data**. Identity/session rows, Organizations, Groups,
  slugs and every binding keep their UUIDs. Down is refused under `APP_ENV=production` and runs only on owned
  disposable databases; it is documented in MIGRATIONS.md and TENANCY.md. No developer or production
  rollback/reset is authorized.

## 4. Internal contracts (extension of `core/internal/tenancy`)

### 4.1 Placement and boundary

- **New production files:**
  - `business_profile.go`: types, `Parse*` functions, Create/Get/Update;
  - `business_context.go`: field and website validation, pure functions;
  - `association.go`: link/get/list/unlink and the access-path reads.
- **Hunks in G005 files:**
  - `store.go`: the 11 new SQL constants and scanners. All SQL stays in `store.go`.
  - `errors.go`: 8 codes and the mapping rows.
  - `log.go`: 4 operation constants.

  `tenancy.go`, `service.go`, `names.go` and `slug.go` are unchanged.
- **Imports:** the existing allow-list plus the standard library `net/url` and `net/netip` (parsing only), `unicode`,
  `unicode/utf8` and `strings`. U-BOUNDARY additionally forbids `crypto/tls`, `os/exec`, `html/template` and
  `text/template`. `net/http*` is already forbidden. `net` remains allowed for G005's error classification.
- **Linked into neither binary** (the existing B-MCP-SOURCE assertion). `main.go` is unchanged.

### 4.2 API additions (U-BOUNDARY pins exactly 20 + 9 = 29 Service methods)

```go
type BusinessProfileID, AssociationID [16]byte                  // distinct types; G005 parse/format rules
func ParseBusinessProfileID(s string) (BusinessProfileID, error)
func ParseAssociationID(s string) (AssociationID, error)
type BusinessContext struct{ DisplayName string; LegalName, Description, Industry, Website *string } // nil = absent
type BusinessProfile struct{ ID BusinessProfileID; OrganizationID OrganizationID; Context BusinessContext; CreatedAt time.Time }
type GroupBusinessProfile struct{ ID AssociationID; OrganizationID OrganizationID; GroupID GroupID
	BusinessProfileID BusinessProfileID; CreatedAt time.Time }
type BusinessProfileAccessPath struct{ UserID UserID; OrganizationID OrganizationID
	BusinessProfileID BusinessProfileID; GroupID GroupID; GroupRole GroupRole } // a fact, not a capability

func (s *Service) CreateBusinessProfile(ctx, org OrganizationID, c BusinessContext) (BusinessProfile, error)
func (s *Service) GetBusinessProfile(ctx, org OrganizationID, p BusinessProfileID) (BusinessProfile, error)
func (s *Service) UpdateBusinessProfile(ctx, org OrganizationID, p BusinessProfileID, c BusinessContext) (BusinessProfile, error) // full replacement
func (s *Service) LinkGroupBusinessProfile(ctx, org OrganizationID, g GroupID, p BusinessProfileID) (GroupBusinessProfile, error)
func (s *Service) GetGroupBusinessProfile(ctx, org OrganizationID, g GroupID) (GroupBusinessProfile, BusinessProfile, error) // singleton lookup
func (s *Service) ListGroupsForBusinessProfile(ctx, org OrganizationID, p BusinessProfileID) ([]GroupBusinessProfile, error) // ORDER BY group_id
func (s *Service) UnlinkGroupBusinessProfile(ctx, org OrganizationID, g GroupID, p BusinessProfileID) error
func (s *Service) ListBusinessProfileAccessPaths(ctx, org OrganizationID, p BusinessProfileID) ([]BusinessProfileAccessPath, error) // ORDER BY user_id, group_id
func (s *Service) ListUserBusinessProfileAccessPaths(ctx, org OrganizationID, u UserID) ([]BusinessProfileAccessPath, error) // ORDER BY business_profile_id, group_id
```

**Not provided:**
- profile delete, move, merge, transfer, cross-Organization list or "list all profiles";
- combined create-and-link (G018 owns the atomic onboarding composite) and replace/reassign;
- any grant, owner or creator write, or "manages" classification;
- deduplication of a profile reachable through several Groups (a caller concern for G007/G018);
- any Organization-membership prerequisite.

A shared profile appears once per Group path.

### 4.3 Business context contract (r1 §4.3, unchanged)

Fields are validated before any SQL, in table order. The first failing field is reported and nothing is written.
1. Valid UTF-8.
2. `strings.TrimSpace` on both ends. No NFC/NFKC normalization, case folding or internal collapsing.
3. `DisplayName` is required. For optional fields, `nil`, `""` or whitespace-only means absent and is stored as NULL.
4. No C0, DEL or C1 code points, except TAB and LF inside `Description`. CR is rejected everywhere.
5. Length in code points: display 1–255, legal 1–255, description 1–4,000, industry 1–255, website 1–2,048. The codes
   are `invalid_display_name`, `invalid_legal_name`, `invalid_description`, `invalid_industry` and `invalid_website`.

Updates replace all five fields in one `UPDATE`; concurrent updates resolve last-committed-wins. Values are bound
parameters only, never rendered, interpolated or logged. The full rule text is r1 §4.3 (archived).

### 4.4 Website contract (r1 §4.4, unchanged)

- Printable ASCII without space or `\`, 1–2,048 bytes.
- Exact lowercase `http://` or `https://` prefix.
- `url.Parse` succeeds with `Opaque == ""`, `User == nil` (rejects any userinfo) and a non-empty `Host`.
- The host contains no `%`. It is an LDH name (which also admits dotted IPv4) or a bracketed IPv6 literal accepted by
  `netip`. A port, when present, is 1–65535.
- The value is stored verbatim. It is **never fetched, resolved, rendered or treated as domain ownership.**

The U-URL valid and invalid lists are carried from r1 §5.2.

### 4.5 Flows, transactions, concurrency and errors

- **Transactions** (G005 parity): one READ COMMITTED transaction per operation, read-write for mutations and
  read-only for reads. Success only after `Commit`. One mutation statement. No retries; a transient failure returns
  `unavailable`. A lost commit reply returns `unavailable` for work that committed, and a retry then reports
  `association_exists`, never a duplicate (documented).
- **Prechecks** run in fixed order, each a scoped indexed `SELECT`:
  1. the Organization exists (mutations and list reads);
  2. the Group exists **in that Organization**;
  3. the BusinessProfile exists **in that Organization**;
  4. for Link only, **the Group has no association** (Spec R3 "service validation").

  Unknown and foreign selectors yield identical codes and identical `Error()` bytes. Single-object reads
  (`GetBusinessProfile`, `GetGroupBusinessProfile`) return the object's not-found code without an Organization
  precheck, as in G005 `GetGroup`.

| Operation | Statement | 0-row / constraint outcome |
|---|---|---|
| CreateBusinessProfile | `INSERT … RETURNING` | 23503 org FK → `organization_not_found` |
| UpdateBusinessProfile | `UPDATE … SET <5 fields> WHERE organization_id=$1 AND id=$2 RETURNING` | 0 rows → `business_profile_not_found` |
| LinkGroupBusinessProfile | `INSERT (organization_id, group_id, business_profile_id) RETURNING` | precheck or 23505 `group_business_profiles_group_key` → **`association_exists`** (same or different profile); 23503 → the endpoint's `*_not_found` |
| GetGroupBusinessProfile | link ⋈ profile `WHERE gbp.organization_id=$1 AND gbp.group_id=$2` | Group precheck → `group_not_found`; no row → `association_not_found` |
| UnlinkGroupBusinessProfile | `DELETE … WHERE organization_id=$1 AND group_id=$2 AND business_profile_id=$3` | 0 rows → `association_not_found` (also when the Group is linked to a different profile; that link is unchanged) |
| access-path reads | `SELECT … FROM app.business_profile_access_paths WHERE organization_id=$1 AND …` | Organization/profile prechecks; an empty list is valid |

- **Concurrency semantics (pinned; tested in §5):**
  - **Distinct-profile or duplicate links on one Group:** at most one row. Losers get `association_exists` from the
    precheck (committed winner) or from 23505 (racing winner). The persisted link is never replaced.
  - **Unlink versus unlink:** one success; the other gets `association_not_found`.
  - **Unlink versus relink:** the end state is either "relinked to the new profile" or "unlinked, and the link reported
    `association_exists`". There are never two rows.
  - **Link versus an owner-level parent delete:** RESTRICT plus FK key-share locks mean exactly one side fails with
    23503.
  - **Context updates:** last committed wins. `FOR NO KEY UPDATE` does not block concurrent link FK checks.
  - **Access paths:** reflect committed bindings and links on the next statement snapshot. Nothing is cached.
- **New codes:** `invalid_display_name`, `invalid_legal_name`, `invalid_description`, `invalid_industry`,
  `invalid_website`, `business_profile_not_found`, `association_not_found`, `association_exists`. G005 codes are reused.
- **Mapping rows added:**
  - 23505 `group_business_profiles_group_key` → `association_exists`;
  - 23503 `business_profiles_organization_fkey` and `group_business_profiles_organization_fkey` →
    `organization_not_found`;
  - 23503 `group_business_profiles_group_fkey` → `group_not_found`;
  - 23503 `group_business_profiles_business_profile_fkey` → `business_profile_not_found`;
  - 23514 `business_profiles_<field>_valid` → `invalid_<field>`;
  - 23514 `*_id_nonzero` → `invalid_id`;
  - anything else, including 42501 and trigger 23000 → `internal`.

### 4.6 Logging (existing event)

`tenancy.operation` adds the operations `business_profile.create`, `business_profile.update`, `association.link` and
`association.unlink`. These carry **no `scope_type`**, as the existing rule requires. `error_code` gains the eight new
codes. Reads are not logged. There are no IDs, names, context values or roles in log records.

### 4.7 Explicitly not added (ownership boundaries)

- **G018:** exactly-one-profile-at-completion, onboarding state or flag, the create-or-select flow, atomic
  create-and-attach, "My Business Profiles", and the Groups/Business Profiles management UI and routes.
- **G007:** TenantContext, capability evaluation and inheritance, the management/read classification of `group_role`,
  and the Organization-owner-only sensitive operations.
- **G008:** RLS on the two tables and the view, and on `role_bindings`.
- **G019:** ownership transfer, last-owner rule, and profile reassignment/removal lifecycle for completed Groups.
  `Unlink` is a foundation primitive. It must not be exposed to silently leave a completed Group without its profile.
- **Not anywhere in G006:** BillingProfile, payment, tax, provider, audit or product objects; HTTP/BFF/MCP/CLI,
  environment, Compose, Dockerfile, `.gitlab-ci.yml` or `infra/ci/**` changes; Skill edits.

## 5. Tests and evidence (designed now; implemented and run only after S0)

### 5.1 Placement and required IDs

- **Unit** (`core-check`; B-UNIT-CHECKS / B-CHECKS-ROOT; the MR `fast` job):
  - new `business_context_test.go`: U-CTX, U-URL, and pre-SQL rejections of the 9 methods;
  - extended `errors_test.go` (U-ERR), `parity_test.go` (U-PARITY) and `boundary_test.go` (U-BOUNDARY).
- **Integration** (tag `integration`; real PostgreSQL in the `migrations` project; runtime-role pool; owner cases via
  migrator `SET ROLE <db>_owner`; failing, never skipping, when settings are absent):
  - **B-BUSINESS-PROFILE-CATALOG** = `TestBusinessProfileCatalog`;
  - **B-BUSINESS-PROFILE-MIGRATION** = `TestBusinessProfileMigration` + `TestBusinessProfilePersistSeed`/`Verify`.
    Both are in the new `core/internal/migrate/business_profile_integration_test.go`.
  - **B-BUSINESS-PROFILE-IT** = `TestBusinessProfileModel` in the new
    `core/internal/tenancy/business_profile_integration_test.go`. It reuses `setup()`, the fixtures and `faultDB`, and
    adds a test-only statement-recording `recordDB` wrapper in its own file.
- **Root:** B-TENANCY-HTTP probe lists extended (H-BP-1/2); no new root ID.
- Each ID passes only with its exact pinned subtest count and no SKIP or FAIL. The counts are fixed when the tests
  are written and recorded in `run.sh` and the Result.

### 5.2 Scenario catalogue

Each negative asserts its SQLSTATE and constraint, or its code, plus before/after row-set hashes over explicit
columns. Multi-violation SQL cases assert set membership of the reported constraint.

**Fixtures** (Spec R8):
- Organizations A and B.
- Groups:
  - A/g1, created by U (so U is `group_owner`);
  - A/g2, created by X;
  - A/g3, created by Y;
  - A/g4, an unlinked foundation Group;
  - B/h1, created by Z, with the same slug as g1.
- Profiles A/p1 and A/p2 with the same display name, and B/q1 with that name too.
- At least six seeded G004 users: U, V, W, X, Y and Z.
- Extra bindings:
  - U: `group_admin`@g3, `org_member`@A and `group_member`@h1 in B;
  - V: `group_member`@g2;
  - W: `org_owner`@A only.
- Links: g1→p1, g2→p1, g3→p2, h1→q1.

| ID | Scenario | Expected |
|---|---|---|
| CT-BP-1 | `reset`; v3 subset `Up`; snapshot S3 (`tenancyCatalog`); full `Up`; snapshot S4 | S4 minus the G006 object lines equals S3: **no predecessor line changes**, including `role_bindings_scope_consistent`. The G006 lines are exactly: 2 tables, 1 view, 5 indexes, the §3.2 constraints, 2 functions, 2 triggers and the column ACLs |
| CT-BP-2 | Ownership and function attributes | All objects are owned by `<db>_owner`. Functions: `prosecdef=false`, `proconfig={search_path=pg_catalog}`, owner-only EXECUTE |
| CT-BP-3 | ACLs | `relacl`/`attacl` equal §3.3 exactly; PUBLIC entries are inspected directly; `has_*_privilege` is false for every unlisted role/privilege |
| CT-BP-4 | Constraints | Names and normalized definitions equal §3.2. Composite FKs are `(organization_id, …)` → `(organization_id, id)` with `confdeltype='r'`. The link's unique indexes are exactly `{pkey(id), group_key(organization_id, group_id)}`. Nothing is unique on `business_profile_id` alone or on `business_profiles(organization_id)` alone |
| CT-BP-5 | Indexes, view, triggers | Exact index definitions. The view has `{security_invoker=true}`, columns `(user_id, organization_id, business_profile_id, group_id, group_role)` and an exact normalized definition (pinned when written). Two enabled BEFORE UPDATE ROW triggers |
| CT-BP-6 | Single authority at v4: `authorityViolations` with the **v4 inventory** | Empty. Role-bearing projections are exactly `organization_memberships.role`, `group_memberships.role` and `business_profile_access_paths.group_role`. Expected sources: `{role_bindings}` for the two G005 views, `{role_bindings, group_business_profiles}` for the new view. The role column is sourced from `role_bindings.role`. No rules other than `_RETURN`, no triggers, no write path |
| CT-BP-6-NEG | G005 N1–N9b rerun at v4 (same expected codes), plus G006 controls. Each runs in its own rolled-back owner transaction | <ul><li>N10 `ALTER TABLE app.business_profiles ADD COLUMN role text` → {`stored_role_outside_authority`};</li><li>N11 the same on `group_business_profiles` (a per-Group profile role) → {`stored_role_outside_authority`};</li><li>N12 `business_profiles ADD COLUMN owner_user_id uuid` → {`stored_role_outside_authority`};</li><li>N13 `CREATE TABLE app.business_profile_memberships (user_id uuid, business_profile_id uuid, role text)` (writable mirror) → {`stored_role_outside_authority`};</li><li>N14 `GRANT INSERT ON app.business_profile_access_paths TO <runtime>` → {`writable_projection`};</li><li>N15 `CREATE OR REPLACE VIEW` of the access view with `'group_owner'::text AS group_role` → {`projection_definition`, `projection_source`};</li><li>N16 `CREATE MATERIALIZED VIEW app.business_profile_access_cache AS SELECT * FROM app.business_profile_access_paths` (cached authority) → {`stored_role_outside_authority`}.</li></ul> Afterwards the detector is empty and the catalog unchanged |
| CT-BP-7 | RLS inventory | RLS off on both tables; no `pg_policy` rows (a G008 obligation, not isolation) |
| CT-BP-8 | No premature surface | Exact columns of both tables: no `group_id` on profiles; no role, owner, creator, permission or capability column. No `app` relation or column matching `billing\|payment\|tax\|invoice\|onboard\|complet\|polic\|capabilit\|permission\|tenant_context`. `role_bindings_scope_consistent` still has exactly the two G005 branches (no BusinessProfile scope). No `business_profile_id` on any predecessor table |
| CT-BP-9 | Runtime negatives | CREATE/ALTER/DROP/TRUNCATE on the new relations; DELETE on `business_profiles`; UPDATE on links; `SET ROLE` owner/migrator; goose metadata read: all 42501 |
| MG-BP-1 | Empty to full | Applies 00001–00004 (exact lines); version 4; v4 object inventory exact |
| MG-BP-2 | Repeated `Up` and `Status` | `no pending migrations`; four applied |
| MG-BP-3 | Populated upgrade: v3 subset `Up`; seed identity and G005 Organizations/Groups/slugs/bindings (existing G005 service); snapshot; full `Up` | `applied 00004` only. Identity and G005 snapshots are unchanged. Runtime smoke via runtime-role SQL (the S3 service does not exist yet at S2): insert a profile, link it, read an access path |
| MG-BP-4 | Reapply | `rolled_back 00004`/`applied 00004`; catalog equal |
| MG-BP-5 | `Down` to 3 with G006 rows present | G006 objects absent; catalog equals S3; identity and every G005 row/binding/slug unchanged; reapply succeeds |
| MG-BP-6 | Persist: seed G006 rows at v4 (runtime-role SQL; parents via the G005 service), PostgreSQL stop/start, verify | Version 4; profile, link and access path present with the same UUIDs |
| MG-BP-7 | Identity fixture `TRUNCATE … CASCADE` at v4 (the G005 ESC-1 analogue) | Identity rows and bindings are emptied; Organizations, Groups, profiles and links are unchanged. The fixture effect is documented |
| IT-BP-ID-1..5 | r1 IT-BE-ID-1..5, renamed | Distinct nonzero v4 UUIDs; equal names give separate rows; updates keep UUID/org/`created_at`/links/bindings; runtime 42501 on key changes and explicit `id`/`created_at`; owner 23000; unknown Organization and zero IDs rejected |
| IT-BP-CTX-1..4 | r1 IT-BE-CTX-1..4 | Field matrix; runtime CHECK negatives with non-canary values; **g1 and g2 read p1 through `GetGroupBusinessProfile`; after one update both show the same UUID and the new context**, while p2, q1, links and bindings are unchanged; inert values; an `httptest` server receives zero requests |
| IT-BP-LINK-1 | Link g1→p1, g2→p1, g3→p2 | Three rows with distinct UUIDs; one p1 row; `ListGroupsForBusinessProfile(p1)={g1,g2}`; `GetGroupBusinessProfile(g3)` returns p2 |
| IT-BP-LINK-2 | Second profile (g1→p2) and duplicate (g1→p1) | Service: `association_exists`, and **zero INSERT reaches the database** (`recordDB`). Runtime SQL `(A,g1,p2)` → 23505 `group_business_profiles_group_key`. The original link's UUID and profile are unchanged |
| IT-BP-LINK-3 | 32 concurrent links of one unlinked Group across p1/p2 (distinct and same), 20 rounds | Exactly 1 success, 31 `association_exists` and 1 row per round, holding the winner's profile |
| IT-BP-LINK-4 | Deterministic interleaving (lock wait detected via `pg_locks` ⋈ `pg_stat_activity`, G005 deviation 5). Tx A inserts `(g,p1)`; the service links `(g,p2)` and blocks. A commits / A rolls back | `association_exists` (23505 path) / success |
| IT-BP-LINK-5 | Unlink(g,p1) versus Link(g,p2): both orders, plus 50 randomized rounds | The end state is in the §4.5 allowed set; never two rows; never p1 after a successful unlink |
| IT-BP-XORG-1..4 | r1 IT-BE-XORG-1..4, renamed (service prevalidation with zero INSERTs; 27-combination runtime INSERT matrix; retarget/reparent; scoped reads) | Only (A,A,A) and (B,B,B) insert; foreign errors equal unknown ones byte-for-byte; rows unchanged |
| IT-BP-UNLINK-1/2 | Unlink g1–p1. Then `Unlink(g2,p2)` while g2→p1. Then 16 concurrent unlinks | g2–p1, p1, Groups and all bindings remain. `association_not_found`, with g2→p1 unchanged. Exactly 1 unlink success |
| IT-BP-ATOM-1/2 | Rejected inputs, and `faultDB` (commit failure, statement failure, cancellation, closed pool), on create/update/link/unlink **and on G005 Group role grant/change/remove at v4** | Rejection codes or `unavailable`; all row-set hashes unchanged |
| IT-BP-REF-1/2 | Owner deletes of a referenced Group, profile or Organization; delete-versus-link races in both orders plus 50 randomized rounds | 23503 RESTRICT; exactly one side fails; no dangling row |
| IT-BP-ACC-1 | Exact access-path rows of the fixture | e.g. `(U,p1,g1,group_owner)`, `(U,p2,g3,group_admin)`, `(V,p1,g2,group_member)`, `(X,p1,g2,group_owner)`, `(Y,p2,g3,group_owner)`, `(Z,q1,h1,group_owner)`, `(U,q1,h1,group_member)`. W has none. `ListUserBusinessProfileAccessPaths(A,U)` lists p1 and p2 (**one global user managing two profiles**). `(B,U)` lists only q1 |
| IT-BP-ACC-2 | Through the G005 service: grant U `group_admin`@g2 → remove U@g1 → change U@g2 to `group_member` → remove U@g2 | After each commit the next read shows: two U paths to p1 → **one path preserved via g2** → **member-only path** (not a management role) → none. Profile and link hashes are unchanged throughout; no G006 statement is issued |
| IT-BP-ACC-3 | Unlink g3–p2, then relink | U's and Y's p2 paths disappear, then reappear. p2's row and the g3 bindings are unchanged; no role write |
| IT-BP-ACC-4 | Organization authority and the unattached profile | W (`org_owner`) has no path rows; organization/group membership counts and hashes are unchanged by every profile and link operation; a new unattached profile has zero paths |
| IT-BP-ACC-5 | Cross-tenant and sibling | `ListBusinessProfileAccessPaths(A,q1)` gives `business_profile_not_found`, byte-identical to unknown. `ListUserBusinessProfileAccessPaths(A,Z)` is empty. V gains no g1 binding or path |
| IT-BP-ACC-6 | Runtime INSERT/UPDATE/DELETE on the view | 42501; rows unchanged |
| IT-BP-ACC-7 | An uncommitted removal of U's g1 binding in an open transaction, with a concurrent reader | The old path is visible until commit and absent after it, on a re-read through the same pool |
| IT-BP-INDEP-1 | Hash snapshots around every G006 mutation and every G005 role mutation | G006 operations never change `role_bindings`; G005 operations never change profiles or links. `recordDB` shows no cross-writes |
| IT-BP-INDEP-2 | `CreateGroup` (creator gets `group_owner`), then link | The creator's path appears from the existing binding. The binding count is unchanged by the link |
| IT-BP-LOG-1 | Success, rejected and failed paths with canary-bearing context | Every `tenancy.operation` record is exact (operation, outcome, level, code, no `scope_type` for profile operations, correlation IDs) |
| H-BP-1 | Core: GET and `POST {}` on `/business-profiles`, `/group-business-profiles`, `/business-profile-access` and `/internal/v1/{business-profiles,group-business-profiles,business-profile-access}` | 404, no model payload |
| H-BP-2 | Web: GET and POST on `/business-profiles`, `/onboarding`, `/api/business-profiles` and `/api/group-business-profiles` | 404. B-TENANCY-HTTP grows from 36 to **56 probes** |
| U-CTX, U-URL | r1 lists (§4.3/§4.4) | Exact codes; trimmed outputs; the URL lists accepted unchanged or rejected with `invalid_website` |
| U-ERR | New mapping rows; bounded `Error()` | Exact codes |
| U-PARITY | Go bounds equal the `00004` CHECK literals; the view text in `00004` equals the pinned text; the 00003 parity is unchanged | Equal |
| U-BOUNDARY | 10 production files; exactly 29 methods; 31 statements (12 writes); write targets are `organizations`, `groups`, `role_bindings`, `business_profiles` and `group_business_profiles`; no projection or view writes; Organization predicate on every scoped read/change; no SQL outside `store.go`; import deny-list | Pass. A dev-only scratch control (an extra exported method) must fail; it is recorded in the Result |

### 5.3 Privacy canaries (r1 §5.3, unchanged mechanism)

- Per-run canaries are passed via `TENANCY_CANARY_FILE` and embedded in all five context fields.
- Canary-bearing rejections are always rejected by the service before any SQL. Database-reached failures in canary
  flows are only unique/FK violations, whose DETAIL carries UUIDs only.
- `logcheck validate`/`scan` run on `/out/business-profile-core.log`, plus a planted-canary control that must be
  detected. B-SECRET-CANARY scans the Compose logs. Test output prints codes and hash prefixes only.

### 5.4 Forced-failure control (AC11)

**F-1 (local, required):** invert one IT-BP-LINK-3 assertion in a scratch clone of the final commit and run
`bash infra/validation/run.sh … migrations` there. The expected result is B-BUSINESS-PROFILE-IT FAIL,
B-REQUIRED-CHECKS FAIL and harness exit 1. B-CI-CONTROLS proves that any FAIL fails `validate.sh` and the gating job.
No hosted failure run (r1 Q11).

### 5.5 Evidence

Evidence goes to `<home>/.local/state/platform-evidence/g006-execution-<base7>/` (never `/tmp`):
- wrapper pre/post outputs and reports;
- catalog dumps, migration output and test logs with subtest counts;
- concurrency outcome counts and `logcheck` outputs;
- NUL-safe status;
- start/end owned-resource inventories;
- rebase/reconciliation records;
- MR and staging pipeline/job/commit/tree identifiers.

No secret, canary or context value is retained.

## 6. Concurrency, transaction and privilege guarantees (summary)

| Guarantee | Mechanism | Evidence |
|---|---|---|
| At most one profile per Group under concurrency | `group_business_profiles_group_key` + precheck | IT-BP-LINK-2..5 |
| Shared context, one row | Full-replacement single-row `UPDATE` | IT-BP-CTX-3 (+ concurrent variant: 16 writers, reader loop through g1/g2) |
| Same-Organization endpoints, before writes and in SQL | Ordered prechecks; composite FKs | IT-BP-XORG-1..4 |
| No dangling rows; no partial operation | RESTRICT FKs; one transaction per operation | IT-BP-REF-1/2, IT-BP-ATOM-1/2, MG-BP-6 |
| Derived access, no second authority | Read-only view over `role_bindings`; detector | CT-BP-6/-NEG, IT-BP-ACC-*, IT-BP-INDEP-* |
| Least privilege; runtime is the real actor | Column grants; preamble identity assertions | CT-BP-3/9, IT-BP-ID-3, IT-BP-ACC-6 |

## 7. Harness, CI and reconciliation (minimal; no new framework)

### 7.1 Required IDs and counts

| ID | Suite | Profile | Added at |
|---|---|---|---|
| B-BUSINESS-PROFILE-CATALOG | migrations | full | S2 |
| B-BUSINESS-PROFILE-MIGRATION | migrations | full | S2 |
| B-BUSINESS-PROFILE-IT | migrations | full | S3 |

| Registry state | Rows | fast | integration | full | staging |
|---|---|---|---|---|---|
| `570b40c` (fact) | 83 | 33 | 71 | 81 | 5 |
| After S2 | 85 | 33 | 71 | 83 | 5 |
| **After S3 = final (forecast)** | **86** | **33** | **71** | **84** | **5** |

The fast and integration counts are unchanged; VP1 and `validate-controls.sh` are untouched. A row and its
recording code land in the same step.

### 7.2 Reconciliation sites (mechanical; same step as the change that needs them)

| Site | Change | Step |
|---|---|---|
| RC-1 `run.sh` B-MIGRATE-SET | Set adds `00004`; add `MIGRATION_00003_SHA=23c45ee7ad45cd0779220c9cf24af60ef7db263b166e88bd8eb9e4cde8f872e3` | S2 |
| RC-2 `run.sh` B-DOCS-EXEC-ROOT, `mig_step`s, B-MIGRATE-CLI/CATALOG | `00004` pending/applied; the version sequence `0 4 4`; apply/reapply at v4 count 35, rollback/inspect at v3 count 27; the catalog set `{0,27,35}`; texts | S2 |
| RC-3 `run.sh` `identity_catalog`, B-PERSIST, B-DBRENAME, the off variant | Add `BUSINESS_PROFILE_OBJECTS` (8 relations) "at version 4"; version 4 and `00004 applied` | S2 |
| RC-4 migrate `integration_test.go` | `reset()` drops the G006 objects first (`businessProfileDrops`). `TestMigrationLifecycle` gets v4 expectations: 02 status, 05 up plus a v4 inventory helper, 07, 08 reapply `00004`, and 09 with one extra down step (4→3, asserting `metadataIdentityAndTenancyOnly`) and up to 4. **Subtest count stays 10** | S2 |
| RC-5 migrate `tenancy_integration_test.go` | `TestTenancyCatalog`, `TestTenancyMigration` and the persist seed/verify use `runnerFor(t, cfg, subsetFS(t, 3))` instead of `newRunner`. **Every G005 assertion is unchanged.** `authorityViolations` takes an explicit projection inventory with expected source sets; the G005 call sites pass the unchanged v3 inventory (r1 review clarification 2) | S2 |
| RC-6 tenancy `integration_test.go` `setup()` | `current_version\t3` → `\t4`. `TestTenancyModel` is otherwise unchanged and runs at v4 (the G005 model contract at v4) | S3 |
| RC-7 identity `store_test.go` | Exact table list adds `business_profiles` and `group_business_profiles` | S2 |
| RC-8 migrate `runner_test.go` | Normal set adds `00004_business_profiles.sql` | S2 |
| RC-9 `run.sh` `TENANCY_PROBE_CORE/WEB` | H-BP-1/2 paths; the B-TENANCY-HTTP ID is unchanged | S3 |
| RC-10 `events.json`, `logrecord_test.go` | §4.6 enums. Accepted records: profile operations without `scope_type`. Rejected records: a profile operation carrying `scope_type`, an unknown operation or code | S3 |
| RC-11 U-BOUNDARY, U-PARITY, U-ERR | §5.2 pins (7→10 files, 20→29 methods, 20/8→31/12 statements, write targets, deny-list) | S3 |
| RC-12 `required-checks.tsv` | 3 rows; header counts | S2, S3 |

Any change to a predecessor test's **meaning**, beyond its version pin, its exact tolerated object set or its
inventory parameter, means **stop and escalate**.

**New `suite_migrations` wiring**, after `tenancy_model`:
- `business_profile_migrations` runs the catalog test, the migration test, then the persist seed → `compose stop`/
  `up -d --wait postgres` → verify. The IDs pass through `BUSINESS_PROFILE_PERSIST_FILE` (0600; mounted read-only for
  verify), as with G005's `TENANCY_PERSIST_FILE`. It records B-BUSINESS-PROFILE-CATALOG and -MIGRATION.
- `business_profile_model` runs `TestBusinessProfileModel` with `TENANCY_CANARY_FILE`,
  `TENANCY_LOG_FILE=/out/business-profile-core.log` and `logcheck` (validate, scan, planted control). It records
  B-BUSINESS-PROFILE-IT.

No new port set, project, image, wrapper or suite.

### 7.3 Order of the `migrations` suite (state between tests)

1. MIGRATIONS.md steps and B-IDENTITY-CATALOG (v4).
2. `TestMigrationLifecycle` (v4).
3. G005 catalog/migration and persist (pinned to v3).
4. `TestTenancyModel` (`setup` applies `00004` → v4).
5. G006 catalog (resets) and migration.
6. G006 persist and restart.
7. `TestBusinessProfileModel` (v4).

### 7.4 CI selection (no mapping change)

- **Path classes.** G006 paths are `full` through the existing rows `core/migrations/`, `core/internal/migrate/`,
  `core/internal/tenancy/`, `infra/` and the non-Markdown `management/`. The identity test is `runtime`, and the
  union is `full`.
- **Pipelines.**
  - The MR pipeline runs `fast` automatically (lint, provenance, contracts, unit: U-CTX/U-URL/U-ERR/U-PARITY/
    U-BOUNDARY). Its `gate` is manual and optional and **is not started**.
  - The staging push after integration runs the automatic `gate`, which the selector resolves to `full` from the
    changed paths. That run includes all three new IDs and every inherited ID.
- **Unchanged:** `.gitlab-ci.yml`, `infra/ci/**` (including `profile-paths.tsv`, `validate-controls.sh`, `lib.sh`,
  `select-profile.sh`, `junit.sh` and `validate.sh`), `APPLICABLE`, `infra/validation/{lib.sh,run-approved*.sh,
  render-customizations.tsv}`.

### 7.5 Validation order and hosted evidence (owner policy, 2026-10-09)

1. **Development:** implement step by step (§10). Focused checks during development are allowed and are not final
   evidence.
2. **No manual-testing or operator-feedback stage.** G006 has no UI or operator qualification (AGENTS 2026-10-08
   rule), so final qualification has no feedback gate.
3. **Rebase before the final local validation:**
   - `git fetch origin`, then rebase the source branch onto `origin/staging` (never a stale local staging);
   - resolve conflicts preserving both changes;
   - verify the target is an ancestor of the source;
   - record both SHAs;
   - reconcile the approved-input snapshot and bindings, with old/new input hashes, as G005 did.
4. **Required local full coverage of the final reconciled code before the MR.** The path matrix requires `full`, so
   run `W(all)` (the full profile, which has no `unit` suite) and `W(unit)` on the rebased candidate. One run listing
   every full-profile suite plus `unit` is equivalent. Unchanged evidence may be reused only with explicit
   byte-identity content/input provenance (G005 r5/L1 precedent). Do not repeat an unchanged identical pass.
5. **Re-check the remote target immediately before opening the MR.** If it advanced: rebase, reconcile and rerun only
   the affected checks.
6. **MR:** push and `glab mr create` to `staging` with source-branch deletion (X1). `fast` runs automatically. The
   `gate` is manual/optional and **not started**; a requested manual gate is extra evidence only. A green MR does not
   claim full hosted coverage.
7. **Codex pre-integration review** of the actual diff and the local/MR evidence. Then an **owner-authorized
   conditional integration** (X2): conditional, not ACCEPT.
8. **Staging gate.** The automatic staging `gate` (full) after integration is the required hosted evidence, bound to
   the integrated content and reconciled with the tested source.
   - **Owner status rule:** `success` or `canceled` satisfies the CI status gate unless any job failed.
   - `failed`, or any failed job or known failing required check, fails it. Pending, running or blocked does not
     qualify.
   - **A canceled run is recorded as canceled and is not passing test evidence** (see Q-r2-10 for AC11).
   - On failure: correct or revert before promotion; dependent G007 execution stays blocked.
9. **Final ACCEPT/completion** by Codex only after a qualifying staging disposition and the review of all criteria.

## 8. File inventory and scope allow-list

**Create (9):**
- `core/migrations/00004_business_profiles.sql`;
- `core/internal/tenancy/{business_profile,business_context,association}.go`;
- `core/internal/tenancy/business_context_test.go`;
- `core/internal/tenancy/business_profile_integration_test.go`;
- `core/internal/migrate/business_profile_integration_test.go`;
- `management/v0/results/G006-scope-allowlist.txt` (first execution write);
- `management/v0/results/G006-business-entity-associations-result.md`.

**Modify (18):**
- `core/internal/tenancy/{errors,store,log}.go`;
- `core/internal/tenancy/{errors,parity,boundary,integration}_test.go`;
- `core/internal/migrate/{integration,tenancy_integration,runner}_test.go`;
- `core/internal/identity/store_test.go`;
- `infra/validation/{run.sh,required-checks.tsv}`;
- `infra/validation/tools/internal/logrecord/{events.json,logrecord_test.go}`;
- `docs/foundation/TENANCY.md`: the BusinessProfile, link and access-path model; the constraint/grant/index
  inventory; the §4.3/§4.4 contracts; codes; concurrency; down effects; the G007/G008/G018/G019 handoffs (§4.7); the
  later-stage table rows for the two tables and the view; future 1:N BillingProfile children keyed by
  `(organization_id, business_profile_id)` through `business_profiles_organization_id_key` with no singleton; the
  updated "Groups may temporarily have no Business Profile" and "access is derived" lines; the identity-fixture
  effect at v4.
- `docs/foundation/MIGRATIONS.md`: current-set row 4, down implications, the validation sentence.
- `docs/foundation/FOLLOW-UP.md` (registered customization; no register change): the STAGE-TENANCY BusinessEntity/
  GroupBusinessEntity item, the BillingProfile-preservation item and the cross-Organization "Verify" line, marked done
  or partly done (G006). The rendered item text is unchanged, with the ADR-004 naming noted in the annotation, in
  G005's wording convention. STAGE-POLICY items stay G007's.

**Untouched:**
- `core/cmd/**`; `core/internal/{config,health,mcp,observability,identity (except store_test.go)}`;
- `core/internal/tenancy/{tenancy,service,names,slug}.go` and their unit tests;
- migrations 00001–00003;
- web, intelligence, contracts;
- Compose, Dockerfiles, `.gitlab-ci.yml`, `infra/ci/**`;
- `infra/validation/{lib.sh,run-approved*.sh,render-customizations.tsv}`;
- `.env.example`, `.archetype`, `skills/**`;
- IDENTITY.md, OPERATIONS.md, VENDORS.md;
- governance, ADRs, Goal/Spec, ROADMAP/PROGRESS, SECURITY.md.

**Scope allow-list** (exact; Codex binds it at E4):

```text
^core/migrations/00004_business_profiles\.sql$
^core/internal/tenancy/(business_profile|business_context|association|errors|store|log)\.go$
^core/internal/tenancy/(business_context|business_profile_integration|errors|parity|boundary|integration)_test\.go$
^core/internal/migrate/(integration|tenancy_integration|runner|business_profile_integration)_test\.go$
^core/internal/identity/store_test\.go$
^infra/validation/(run\.sh|required-checks\.tsv)$
^infra/validation/tools/internal/logrecord/(events\.json|logrecord_test\.go)$
^docs/foundation/(TENANCY|MIGRATIONS|FOLLOW-UP)\.md$
^management/v0/results/G006-(scope-allowlist\.txt|business-entity-associations-result\.md)$
```

## 9. Acceptance-criteria mapping (Spec AC1–AC12)

| AC | Tests and evidence |
|---|---|
| AC1 | CT-BP-4/8 (no `group_id`, exact columns, composite target); IT-BP-ID-1..5; U-BOUNDARY (no move/reparent/delete method) |
| AC2 | U-CTX, U-URL, U-PARITY; IT-BP-CTX-1/2 (service and runtime-SQL negatives, unchanged rows); IT-BP-CTX-3 (g1/g2 see one committed update by the same UUID; unrelated profiles/links/bindings unchanged); IT-BP-CTX-4 (no fetch; inert SQL/HTML/template values) |
| AC3 | IT-BP-LINK-1 (two Groups share p1, a third uses p2; exact UUIDs and counts); IT-BP-LINK-2 (second profile rejected by the service with zero INSERTs, and by runtime SQL 23505; the original link unchanged); IT-BP-LINK-3/4/5 (duplicate, retry, concurrent distinct-profile and interleaved attempts: at most one link, intended conflict); CT-BP-4 (Group-unique; profile side not unique) |
| AC4 | IT-BP-XORG-1 (prevalidation, zero INSERTs); IT-BP-XORG-2 (27 runtime-SQL combinations); IT-BP-XORG-3 (retarget/reparent); IT-BP-XORG-4; CT-BP-4 composite-FK catalog; before/after rows |
| AC5 | IT-BP-UNLINK-1/2 (other link, context, Groups and memberships retained); IT-BP-ATOM-1/2 (profile/link and Group-role operations leave no partial rows); IT-BP-REF-1/2; CT-BP-4 `confdeltype='r'` |
| AC6 | CT-BP-5 (exact view and SELECT-only grant); CT-BP-6/-NEG N10–N16; CT-BP-8 (no BusinessProfile scope, owner column or role row); IT-BP-ACC-1 (one global user owns/administers Groups for p1 and p2, plus a membership in B); IT-BP-ACC-2/3 (grant/change/remove/downgrade/unlink change only the applicable facts immediately; profile rows unchanged; the g2 path preserves access); IT-BP-ACC-6 (projection writes fail) |
| AC7 | IT-BP-INDEP-1/2 and IT-BP-ACC-2..5: profile and unrelated binding snapshots around each mutation; the last qualifying path's removal removes its facts; one of two paths survives; member-only differs from owner/admin; no copied store (CT-BP-6 A1, N11/N13/N16), no automatic Organization role and no sibling-Group authority |
| AC8 | B-TENANCY-HTTP (56 probes, H-BP-1/2); B-MCP-SOURCE; B-WEB-BOUNDARY and contracts unchanged; B-ENV-EXAMPLE; CT-BP-7/8 (no policy, RLS, onboarding, completion or billing object; no singleton on profiles); TENANCY.md future 1:N and G007/G008/G018/G019 owners; B-SCOPE |
| AC9 | MG-BP-1..7 (B-BUSINESS-PROFILE-MIGRATION); CT-BP-1..5 and CT-BP-9 (B-BUSINESS-PROFILE-CATALOG); reconciled B-MIGRATE-*, B-IDENTITY-CATALOG, B-PERSIST, B-DBRENAME, B-TOGGLE-OFF; B-MIGRATE-SET pins the 00001–00003 digests (accepted migrations byte-identical); identity/G005 preservation in MG-BP-3/5 |
| AC10 | Final local full coverage (§7.5 step 4) with every inherited ID: B-TENANCY-* (incl. the G005 model at v4), B-IDENTITY-*, browser, first-paint, tokens, boundary, variant, dependency and provenance. IT-BP-LOG-1 with `logcheck` and the planted control; B-SECRET-CANARY; B-CONTRACT-TOOLS; controls rerun in the same cycle as shared-harness edits; no relaxed reason codes or exclusions |
| AC11 | Wrapper runs on the actual root and hash-matched copies; the final local `W(all)` and `W(unit)` on the rebased candidate (off variant included, no required skips); F-1 plus B-CI-CONTROLS; automatic MR `fast`; the automatic staging full gate after conditional integration, binding the integrated commit/tree/job IDs and the three G006 IDs' results under the owner status rule |
| AC12 | E4 binding; wrapper pre/post; NUL-safe status; allow-list; start/end inventories; rebase reconciliation records; failure history; B5 accounting with the cycle-3 checkpoint; Result; Codex ACCEPT. No inheritance, RLS, onboarding, billing or product claim |

## 10. Sequence, delegation and commands

**Command form** (the accepted approved-input wrapper, unchanged; run under
`flock <home>/.local/state/platform-evidence/local-harness.lock`):

```sh
bash infra/validation/run-approved.sh --snapshot <S> --snapshot-sha256 <D> --count <N> --path-re '<RE>' \
  --plan management/v0/plans/G006-business-entity-associations-plan.md -- \
  --base <BASE> --allow-file management/v0/results/G006-scope-allowlist.txt --report-dir <EVID>/<step> <suites>
```

| Step | Work (tests, wiring and registry rows land together) | Gate `W(…)` |
|---|---|---|
| S0 | Start gate (§2.3) | Pass |
| S1 | Scope allow-list copy | `W(lint)` |
| S2 | Migration 00004; `TestBusinessProfileCatalog` (CT-BP-*, N10–N16) and `TestBusinessProfileMigration` + persist; RC-1..5, RC-7, RC-8, RC-12 (2 rows) | `W(unit lint controls migrations)`, then `W(root ops rename off)` |
| S3 | Package extension and unit tests; `TestBusinessProfileModel`; RC-6, RC-9..12 (1 row); then the read-only security/SQL review | `W(unit)` inner loop, then `W(lint controls contracts unit root migrations ops off)` |
| — | **Cycle-3 checkpoint** (expected around S3–S5) | §11 |
| S4 | Docs: TENANCY.md, MIGRATIONS.md, FOLLOW-UP.md | `W(lint provenance migrations off)` |
| S5 | F-1 in a scratch clone | Expected FAIL recorded |
| S6 | Fetch, then rebase onto `origin/staging`; resolve and reconcile the bindings (§7.5 step 3) | Recorded SHAs and reconciled snapshot |
| S7 | **Final local full coverage of the rebased candidate**: `W(all)` and `W(unit)` (or a byte-identity reuse mapping) | All 84 full IDs plus B-UNIT-CHECKS PASS |
| S8 | Target re-check; X1 push and MR; automatic `fast`; Result commit | `fast` PASS; gate not started |
| S9 | Codex review → X2 owner-authorized conditional integration → automatic staging gate → Result update | Owner status rule; Codex decides ACCEPT |

**Run rules:**
- `controls` always runs with a secret-generating suite (G005 r3 lesson).
- Every run includes the `final` checks.
- The repository is never edited during a run (B-RENDER-PLAN).
- A shared-file correction reruns its covering suites before its step closes.
- No step depends on a later step, and no ID is disabled.

**Delegation:**
- Claude main owns every file and the integration. The Goal's optional database/test delegate is **declined**: the
  migration, catalog/model tests and shared wiring change together.
- One **read-only** security/SQL reviewer after S3, at zero cycles unless its findings require corrections, which
  then count against B5.
- Aggregate cycles are recorded per agent.

## 11. Budget, forecast, checkpoint and stops

**B5: at most five shared meaningful cycles; checkpoint after cycle 3.** Zero consumed by planning. The forecast is
**3–4 of 5. Risk: low–moderate**, which is lower than r1 because no predecessor object changes.

| Cycle | Expected material issue |
|---|---|
| C1 | Migration and catalog: exact view/constraint deparse pins, the detector inventory refactor, the v3 pinning of the G005 tests, `reset` order, lifecycle step 09 |
| C2 | Model tests: interleavings, `recordDB`, website/Unicode edges, access-path assertions |
| C3 | Harness: version-4 sites across root/ops/rename/off, the extra persist restart, `logcheck`, security-review corrections |
| C4 | Rebase reconciliation (G005 corrections or G051 integration) or a staging-gate correction |
| C5 | Reserve |

**Checkpoint after C3** (BUDGET_POLICY §7): ACs done and remaining, repeated failures, any non-mechanical
reconciliation, convergence and boundedness. If the work is not converging, stop and return a bounded proposal.

**Stop and escalate** (in addition to the Goal's list):
- any E-gate missing at S0;
- any G005 schema, Service or helper difference beyond §2.2's mechanical class;
- a needed change to the meaning of a G004/G005 test, or any edit to an accepted migration;
- a need for a BusinessProfile role, scope, owner column, grant, writable/materialized mirror, `role_bindings`
  change, onboarding/completion state, replace workflow, public surface, policy/RLS, billing schema or new dependency;
- CI changes beyond §7.4;
- a staging-gate failure after conditional integration (report it; correct or revert only as authorized);
- thrashing or B5 exhaustion.

## 12. Named external actions (none performed; each needs separate authorization)

- **X1:** push the final rebased branch (proposed name `g006-business-profile-associations`) and run
  `glab mr create --source-branch <it> --target-branch staging`, keeping source-branch deletion. MR `fast` runs
  automatically; the manual MR `gate` is not started unless the owner requests it.
- **X2:** owner-authorized conditional integration via glab after Codex pre-integration review. The automatic staging
  `gate` follows, and its source-branch deletion is verified.

No WorkOS, provider, `identity-qualification`, protected-ref, schedule, settings or runner action. G006 changes no
identity behavior, so no B-IDENTITY-STAGING qualification is required.

## 13. Decisions and questions for Codex review

| ID | Question | Proposal |
|---|---|---|
| Q-r2-1 | Derived-fact representation | `security_invoker` view `business_profile_access_paths` (§3.2), Group paths only, raw `group_role`. Organization authority stays in `organization_memberships`. No management/capability column (G007) |
| Q-r2-2 | Duplicate versus different-profile link | One conflict code, `association_exists`, from the precheck or 23505. No idempotent return (r1 Q2) |
| Q-r2-3 | Unlink shape | `(org, group, profile)` predicate; no replace operation; documented as a foundation primitive that G018/G019 must guard |
| Q-r2-4 | Detector adaptation | Explicit per-projection inventory with expected source sets. The G005 call sites pass the unchanged v3 inventory (r1 clarification 2, re-targeted) |
| Q-r2-5 | Pinning the predecessor tests | G005 catalog, migration and persist tests use `runnerFor(subsetFS(3))`; `TestTenancyModel` `setup()` pins 4; `TestMigrationLifecycle` moves to v4 with 10 subtests |
| Q-r2-6 | Persist phase | A separate G006 stop/start (one extra PostgreSQL restart) instead of editing G005's `tenancy_migrations` |
| Q-r2-7 | Import deny-list | Keep `net`; forbid `crypto/tls`, `os/exec`, `html/template` and `text/template` |
| Q-r2-8 | IDs and names | `B-BUSINESS-PROFILE-{CATALOG,MIGRATION,IT}`; 29 methods (§4.2) |
| Q-r2-9 | Spec r3 residual wording | Risk 4 still says "Direct grants and links are facts only"; AC2/AC5 mention "grants"; the Spec links `skills/project-archetype/…`. Interpretation: "grants" means the existing Organization/Group RoleBindings; no BusinessProfile grant exists. The Spec is Codex-owned and is not edited here |
| Q-r2-10 | A canceled staging gate versus AC11 | Under the owner status rule, `canceled` with no failed job satisfies the CI status gate. AC11 asks the "final hosted job" to bind the model/migration results, and a canceled run produces none. Proposal: the Result records it as canceled, never as a pass, and Codex decides whether AC11's hosted part is met or needs an owner waiver |
| Q-r2-11 | OPERATIONS.md stale profile counts (F-15) | Out of G006 scope; leave it to G051 or a Codex documentation fix |
| Q-r2-12 | `infra/ci/junit.sh` in the approved inputs (G005 residual) | Codex decides at E4. G006 does not edit it |

**Ambiguities recorded, not resolved:**
- The archetype reference keeps many-to-many and `be_*` examples. ADR-004 supersedes them; no Skill edit.
- `docs/SECURITY.md` is empty (SHA-256 `e3b0c442…`). The named architecture, ADR and Spec invariants apply.
- ARCHITECTURE §53.4 lists addresses/contact metadata; Spec R2 makes them non-mandatory, so none are added.

**Context and scope:** one bounded Goal within B5. **No split recommended.**

## 14. Research record (read-only, 2026-10-09)

- **Local sources:** every input in `G006-planning-inputs.json` at `570b40c`, and the external final G005 Result blob
  (`0e32d93`). Also read without binding: the G051 Plan r1 at `19dc50e7` (overlap only) and the root checkout's local
  staging `2e4d4d5`, whose changes from `origin/staging` are PROGRESS and the G051 review only. No Docker, build, test
  or network action besides `git fetch origin`.
- **PostgreSQL 17 and Go standard-library behaviours** are as cited in r1 §15. They are proven by the S2/S3 tests,
  not assumed. A new point: `security_invoker` views apply the base tables' privileges and RLS to the invoker, which
  is why the view needs no grant beyond SELECT and inherits G008 policies.
