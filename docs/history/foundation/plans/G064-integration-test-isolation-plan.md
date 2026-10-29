> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G064 — Integration-test isolation and per-test evidence — Implementation Plan (r2)

Goal: [G064](../goals/G064-integration-test-isolation.md) (sha256 `438df4fc6cf6de5b34a2c7dc6293b4964aaa8afb4dd531e410fc574d7a0ad0c0`; r0 bound `4260218d…`) ·
Spec: [G064 Work Spec](../specs/G064-integration-test-isolation-spec.md) (sha256 `d8e23cb71ba157a8b57b0ae79123537dea9c963f31a52606ca9c8720d1a3ffb3`; r0 bound `e669c8b6…`; includes the
[bounded lifecycle clarification](../specs/G064-integration-test-isolation-spec.md#bounded-lifecycle-clarification--plan-r0-review-2026-10-21)) ·
r0 Codex review (not included) (sha256 `bf70ed69b2f3b34547fd789337183a781d0c2e25ddbb59c5ffeb78d3d5d5d7f1`, CORRECTION LOOP, D-1..D-6) ·
r1 Codex review (not included) (sha256 `f98fad76d97258a2a82a81b3da13a52992fbbaab16f82b18491bb5f51fe7d683`, CORRECTION LOOP, three bounded corrections + one mechanical §8 fix) ·
[PG-010](../PERFORMANCE.md#pg-010--g064-test-database-cloning-cost-and-connection-contention) ·
[DATABASE-ISOLATION.md](../../../foundation/DATABASE-ISOLATION.md) · [MIGRATIONS.md](../../../foundation/MIGRATIONS.md) ·
[OPERATIONS.md](../../../foundation/OPERATIONS.md).
Status: **r2 submitted for Codex Plan re-review (one bounded revision answering the r1 CORRECTION LOOP's three
corrections and its mechanical §8 fix; the five r1 design corrections and D-1..D-6 are settled and unchanged). PLANNING
ONLY: no build, test, install, Docker, Compose, database or runtime command was run for r0, r1 or r2; implementation is
NOT started and NOT released. Zero of B5 execution attempts spent (checkpoint after attempt 3).**
Base: `g064-plan` r1 (`bf70e268b4a8f6a6950993a7d4247f7b2920450d`; r0 `e3117f2d9ec0e4454a9bd30b72c3bedf3e8d6627`,
branched from `origin/staging` `26240112acba2094009668f543a78a17098b87fe`, merged with `origin/staging`
`5a52fe8586c7e81fb23433904c3512c34f8e6b7f` at r1) merged with `origin/staging`
`79e457ddabbfb0379e99ee7122344126966b32be` (merge commit; r0/r1 history preserved). The staging deltas
`26240112..5a52fe85` and `5a52fe85..79e457dd` touch `management/` only (verified each time: `git diff --stat ...
':(exclude)management'` is empty), so **every code file:line reference below is still at `26240112` and every
non-management blob in §1.3 is unchanged**. G020's paused source `origin/g020-impl`
`df2c46f486480ad13a61ef4c4b8f5716836e74ae` was read read-only (`git show`) for §12 only (r1).
Planner: Fable 5.1 (requested high; the runtime effort is not observable from inside the agent), isolated worktree
`.claude/worktrees/agent-<id>` (r1: `agent-<id>`; r0: `agent-<id>`), branch
`g064-plan`. Read-only inspection only (git/grep/sed/cat).

## 0. Order, gates and tier

- **Order (owner-confirmed, 2026-10-21; Goal/ROADMAP/PROGRESS):** **G063 → G064 → RC7 → G020 → combined staging preview
  → RC8 (if needed) → G057 → G021 → RC9.** G064 planning and approved isolated implementation may overlap G063's frozen
  qualification (its checkout/source is never touched or switched, §3.6, §9); **G063 final ACCEPT precedes G064
  staging integration/final acceptance**; both ACCEPTs precede the RC7 cut and its full qualification; **G020 stays
  paused at S1** (`origin/g020-impl` `df2c46f4`, migration `00020` unmerged, B5 round 1 / attempt 1 preserved) **until
  G064 final ACCEPT and RC7 qualification**, then resumes after supervisor reconciliation (§12) without a budget reset.
  No G020 artifact, branch or ledger is edited by G064.
- **Tier:** architecture/security/data (privileged test credentials, database lifecycles, security evidence). Explicit
  Codex Plan ACCEPT, exact input/scope binding and a recorded supervisor execution release are required before any
  implementation, build or runtime probe.
- **Gates before implementation:** Codex Plan ACCEPT of **r2** (the D-1..D-6 rulings of the r0 review are recorded in
  §10 as approved with their evidence corrections; the r1 review's three corrections are resolved per the r2 mapping); imported scope allow-list (§9); bound inputs reconciled at the exact
  staging SHA of release (§1.3; exact-delta rule for irrelevant staging changes); isolated task branch `g064-impl` from
  the release SHA; one build slot and the shared two-run pool; recorded execution release. Plan existence, merge or CI
  alone is not release.
- **Gates after implementation:** Codex source ACCEPT → MR to `staging` (auto-merge after gates) → **G063 final ACCEPT
  before the G064 merge/final acceptance** → coordinator-owned local full qualification on the exact merged SHA →
  cleanup receipt → Codex final ACCEPT → RC7 → G020 reconciliation (§12).
- **Budget:** B5, checkpoint after attempt 3, **zero spent** (r0, r1 and r2 are planning only). Forecast: 2 attempts
  (lifecycle slice; reporter slice).
- **Context policy:** 150k checkpoint / 200k safe handoff for the writer.

## 1. Current state (what the Plan is grounded in)

### 1.1 One database, one role family, sequential invocations

- The `migrations` suite is one disposable Compose project `g001-mig-$RUN_ID` (`infra/validation/run.sh:2471-2477`) with
  one PostgreSQL (`pgvector/pgvector:0.8.7-pg17`, `infra/compose/local.yml:9-29`; init script
  `infra/postgres/init/10-roles.sh` mounted at volume creation, line 21) and **one database `$DATABASE_NAME`**
  (`beginify` in `.env.example:13`). Every Go integration invocation runs in the `$project-core-check` image (source is
  `COPY`ed into the image at build, `infra/docker/core.Dockerfile:5-11`, target `check`) on the project's `data` network
  with the `it.env` file (`DATABASE_NAME, DB_HOST=postgres, DB_PORT=5432`, migrator/runtime/retention/superuser
  passwords, `APP_ENV=test`; `run.sh:2557-2560`), through `docker_run_owned` (`lib.sh:320-331`) or `rls_go`
  (`run.sh:2874-2882`: `sh -c 'go test -tags integration -count=1 -timeout 30m -v -run "$1" "$2"; rc=$?; sed ...
  /tmp/rls-access.log; exit $rc'`).
- The image's source is the harness copy `$WORK/mig` made by `copy_repo` (`lib.sh:162-165`: `git ls-files --cached
  --others --exclude-standard` of `$REPO`, i.e. the **working tree**, not necessarily `HEAD`); `impl_tree_hash DIR`
  (`lib.sh:168-171`) already hashes exactly that path/content set and `B-SOURCE-HASH` (`run.sh:4826`) compares copy and
  checkout. `--base` is only the comparison baseline of `B-SCOPE`/`B-SKILL-UNCHANGED` (`run.sh:81,116,5866`); nothing
  today records the tested tree beside a Go result.
- `psql_super DEST SQL` (`lib.sh:697-699`) connects as `postgres` **to `$DATABASE_NAME` itself**; there is no
  maintenance-database helper yet (needed for `CREATE/DROP DATABASE`, §3.1).
- **Role names derive from the database name in three places:** `migrate.Config.Database` prefixes every role
  (`core/internal/migrate/runner.go:26-43`; the runner connects to `cfg.Database`, `:110`); `10-roles.sh:10-18` and
  `provision-roles.sql:118-125` build `<db>_{owner,migrator,runtime,retention,policy_reader,auth_reader,claim_writer,
  provision_writer}`; and **every migration from 00002 to 00019 derives role names from `current_database()` while it
  runs** (DO blocks such as `00006_row_level_security.sql:35,343,607`, `00014_group_onboarding.sql:33-35,212-214`,
  `00015_membership_invitations.sql:422-427,1010-1013`), baking role names as literals into generated bodies
  (`00009_retention_maintenance.sql:306-310`: `audit.events_refuse` is created with `format(... current_user = %L ...)`).
  Consequence: **any database on which migration SQL executes must carry the canonical name `<db>`**; a clone of a
  migrated database keeps ownership/ACLs by role name and its generated function bodies keep the baked `<db>_*` literals.
- `CREATE DATABASE ... TEMPLATE` copies the per-database catalogs (relations, functions, policies, grants, default
  privileges, extensions) but **not** the shared catalogs: database-level ACL (`datacl`), database owner, `ALTER DATABASE
  ... SET` rows (`pg_db_role_setting`, keyed by database OID), roles, memberships and role settings. The provisioning
  script establishes `REVOKE ALL ON DATABASE ... FROM PUBLIC; GRANT CONNECT ... TO <db>_migrator, <db>_runtime,
  <db>_retention` (`10-roles.sh:1044-1045`) and role-level `search_path` settings (`:1061-1063`, cluster-global).
- The existing shared-catalog query `rls_prov_catalog` (`run.sh:2998-3027`) lists only roles `LIKE '<db>\_%'`
  (`:3003`), **memberships whose member and role are both in the family** (`:3008`), the database ACL of
  `current_database()` only (`:3010`) and role settings **without the `setdatabase` column** (`:3023-3024`): a grant
  such as `GRANT pg_read_all_data TO <db>_runtime` or a `<db>`-qualified role setting is invisible to it. It is reused
  as the PROV-1 basis, not as the G064 detector (§3.3).
- No migration seeds rows at migration time: every `INSERT INTO app.*` in `core/migrations/*.sql` is inside a function
  body (`00014:128`, `00015:807,891`, `00016:797,967,1019,1049,1304`, `00017:507,1711`, `00018:642,648`; the `audit.*`
  inserts are trigger bodies). The migrated baseline is therefore **goose's own history satisfied for the bound
  migration set and every other row-bearing `app`/`audit` relation empty** (verified again at S0; §3.1 derives it, no
  pin). `app.goose_db_version` is **not** "one row per migration": goose v3.28.0 (`core/go.mod:14`) inserts a
  **version 0 row** when it creates the table (`core/internal/migrate/metadata.go:17-25`; `protectedStore` only adds
  owner-only protection and delegates every metadata operation to goose's PostgreSQL store, `:26-47`), `Down`/`Reapply`
  (`runner.go:245-280`) change the history of the top version, and the verbatim CLI steps (`run.sh:2524-2528`: apply,
  reapply, rollback, apply) leave `00019`'s history with several rows. The applied set is what goose reports
  (`Status`, `runner.go:204-216`: `migration\t<file>\t<state>\t<applied-at>` per known file, then
  `current_version\t<n>` from `GetDBVersion`); the SQL counterpart is the latest row per `version_id` (§3.1).
  Row-bearing relations are `pg_class.relkind IN ('r','p','m')` (ordinary tables, partitioned parents, materialized
  views); the two schemas also hold sequences (`relkind 'S'`: goose's identity sequence, the enrollment ledger
  sequence of `00016`), indexes (`'i'`) and views (`'v'`), which are not emptiness evidence (`core/migrations` has no
  `MATERIALIZED VIEW`/`PARTITION BY` today; re-grepped at S0 and recorded).

### 1.2 Lifecycle coupling between invocations (the order dependence to remove)

| Coupling | Where |
|---|---|
| `identity_catalog` must run "before the lifecycle test resets the database" | `run.sh:2546-2547`; `reset(t, cfg)` drops every object down to `goose_db_version` (`migrate/integration_test.go:254-300`) |
| `TestMCPRecovery` "expires every session, so before the drill's reset" | `run.sh:2597, 3552-3556`; it applies the real `invalidate.sql` (`MCP_INVALIDATE_SQL_B64`, `:3589-3591`) |
| `recovery_drill` does `compose down -v` and migrates afresh | `run.sh:4263-4265` (last consumer of the project's database) |
| `TestTenancyModel` and `TestBusinessProfileModel` share `setup(t)` whose version ladder is conditional on the version the shared database is at (19 → purges + 14 downs → version 5; a database already at 5 skips everything) and **leaves the database at version 5** | `tenancy/integration_test.go:94-315`, `business_profile_integration_test.go:211-212` — the known fresh-DB run-order artifact |
| `setupIT`/`setupMCP` migrate `Up` to 19 and `DELETE FROM app.identity_provider_binding` to evict the previous test's provider namespace; all other rows stay | `access/integration_test.go:228-240`, `mcp/env_integration_test.go:248-258` |
| identity/signup `TestMain` migrate `Up`, keep one package pool (40/20 conns) and release fixture slugs at exit for the later `mcp-seed.sql` consumer | `identity/integration_test.go:161-200`, `signup_integration_test.go:52-91` |
| Identity overlay: core and core-replica hold pools on `<db>` while `identity_it` runs every identity test on that same database (`./internal/identity/...`, `run.sh:5353-5364`), PASS needs `^ok` **of the `identity` package only** (`:5363`: `^ok[[:space:]]+core/internal/identity[[:space:]]`), no FAIL/SKIP and **at least ten** `--- PASS:` lines although the pattern selects three packages: `identity` (35 `integration`-tagged top-level tests **plus 39 untagged ones** in `allowlist/assurance/attest/credential/enroll/http/identity/limiter_copy/mcp_name/mq_eval/retire_test.go`), `identity/workos` (21 untagged tests in `adapter/connect/jwks/token_test.go`; `staging_test.go` is `workos_staging`-tagged and excluded) and `identity/identitytest` (no test files: `?  ... [no test files]`); a `workos` failure is caught only by the `^FAIL` grep, and a `workos` package that never terminates is not caught at all; then `mcp_seed` applies `infra/identity/mcp-seed.sql` (Organizations `0f14a000…000a/b/c` with slugs `org-a/b/c`, `:43`) and the browser projects consume it | `run.sh:5216-5220`, `web/playwright.config.ts:18-21`, specs `identity/membership/enrollment/mcp/onboarding/signup.spec.ts` |
| Persist pairs: seed → `compose stop postgres` / `up -d --wait postgres` → verify (IDs through env/files) | tenancy `run.sh:2657-2665`, business profile `:2720-2729`, policy `:2786-2794`, RLS `:3116-3123` (`RLS_PERSIST_IDS`), objects `:3288-3292` (`OBJ_PERSIST_IDS`), audit `:3378-3382` (`AUDIT_PERSIST_IDS`) |
| Cluster-global mutations by tests: `ALTER ROLE <db>_provision_writer RENAME/INHERIT` (restored in-test, `migrate/signup_integration_test.go:133-141`); N7 `ALTER ROLE ... LOGIN` (rolled back, `migrate/rls_integration_test.go:516`); `ALTER DATABASE <db> SET/RESET track_functions` (`access/objects_footprint_integration_test.go:134-138`, `audit_integration_test.go:1736-1739`, `rls_integration_test.go:1707,1789`); `TestRLSProvisioned`/PROV-1..6 on standalone no-init instances (`run.sh:2969-3060`) | — |
| Evidence today: `go test -v` text per invocation under `$WORK/*.log`; only `migrate-it.log` is in `STAGE_REPORT` (`infra/ci/validate.sh:122-131`); every other log survives only through `export_diagnostics` (`lib.sh:35-57`: on non-zero status or a FAIL/UNMET row, generated-secret redaction, 1 MiB bound) from the EXIT trap (`run.sh:109`: `RUN_STATUS=$?; bg_stop_all; export_diagnostics; cleanup`); `stage_one` silently skips an absent allow-listed name (`validate.sh:1378`) and the `complete` step checks presence only for `results.tsv`, `required-checks.tsv`, browser, contract and leftover files (`:1255-1372`) | — |

### 1.3 Bound inputs (blob SHA-1; code paths at `26240112` = unchanged at `5a52fe85` and at `79e457dd`; the writer re-verifies at S0 and records the release SHA)

```text
35270b470dd65f323175ef16e4b0faafe8e8c278 core/internal/migrate/runner.go
6882ff6edaa393f9c33c51076b5307e0ea72e65a core/internal/migrate/integration_test.go
df83c52977b0c316c18e2f88d530672257e1ea5c core/internal/migrate/tenancy_integration_test.go
a7f84ec7d68b177a62b2a9b4d80c9f219bd7673c core/internal/migrate/rls_integration_test.go
49b83cec41018769d83e9d5ba09fbe371d7924d0 core/internal/migrate/signup_integration_test.go
720a0cba3c6bf501a56926303b50bbd77f9d7026 core/internal/migrate/mcp_integration_test.go
f06d82d8078e184adac9241d0e823194fd6c4bb7 core/internal/identity/integration_test.go
471fcb1342c4f87ed825914e52d0a3a8bd3334bf core/internal/identity/rls_pool_integration_test.go
76e5181170cbb75eafd6b9c5b13a541ddd80e5fc core/internal/identity/store_test.go
f0bcb2eae0ba37cf41a322b7989be56c044bfc36 core/internal/signup/signup_integration_test.go
c9625d534fafa5f7386add8aa0d4d0b718cdd56c core/internal/tenancy/integration_test.go
e31dfc949e21dae6e693b5f2084c1fee58c3b6b0 core/internal/tenancy/business_profile_integration_test.go
ee2c0d28164a3100993b5a4ae1f69e174c3de107 core/internal/tenancy/access/integration_test.go
a344dc55eea0affcfd44291120fc8d7c3e4202f7 core/internal/tenancy/access/rls_integration_test.go
3b8a95492b67e3620da0c74b1d02d598af6229a6 core/internal/tenancy/access/recovery_integration_test.go
6d4c1359e171fdc15e5f8a558887d7e2e622c087 core/internal/tenancy/access/objects_integration_test.go
c6bc8433f50fb9b16092637294786256a38db1fa core/internal/tenancy/access/audit_integration_test.go
3e59c9b14d5b855730d763d88b4837459bca2e30 core/internal/tenancy/access/retention_integration_test.go
fbd3425eed26fe0941c8e70dd2b194a31ce07f57 core/internal/tenancy/access/mcp_integration_test.go
7830631f20365835e58b2be58a471657af9bf4fa core/internal/tenancy/access/objects_footprint_integration_test.go
aedc76b19c6160939bd9f3921e0eb6c955c2d941 core/internal/mcp/env_integration_test.go
9083a7cad375dea5856be87d01532fd6551089b9 core/internal/mcp/mcp_integration_test.go
f7625daacb2ec727d956d192a39480f2ec645640 infra/postgres/init/10-roles.sh
8ac2befb115e5a356defa12c17d7f8dd5e0ef6f6 infra/postgres/provision-roles.sql
4b0207d3c31ffe904f71b62118f16a644676cff0 infra/validation/run.sh
62dc5210de2eef44e65744949be8149198774658 infra/validation/lib.sh
7a78da2ce98dd3dfa1bfd40e0c065a1a040924d1 infra/validation/required-checks.tsv
f5c305cf34dbf11dacdaf5f8586d0ee2da6ff0a7 infra/ci/validate.sh
fcc5178ba6ca92ddd86810ad9673b54ef75552e9 infra/ci/lib.sh
adb4bfa4bc7bf66a6f145d763044ef7bc1b8f74a infra/ci/junit.sh
d6e5fa88493b3f53e93913cc47e8ee0516864d63 infra/compose/local.yml
3d31d158758f414dd6690f9e62d33362f86b67a0 infra/compose/migrate.yml
37573b8daf89a873a17cf4198251737b4bf04942 infra/compose/identity.yml
39c0dac522d7902cf65797962484ab40adaf9a61 infra/compose/browser.yml
753df665a294450322b5a54a52735e99e6ede680 infra/identity/mcp-seed.sql
192af8e955d372aa9e3dedb19e8e077a8a09c97c infra/identity/mcp-hosts.tsv
541e1126afcfd4f5742e86f354cc1bfa38c4a2e8 infra/recovery/recovery.sh
033821e7cad001bfa39088d79cf9a149bdf82a12 infra/recovery/sql/invalidate.sql
f4edfb564bb0d0b79dbe98db6bdd81cd1fa4b4db infra/docker/core.Dockerfile
8dfcdedc61cc7e29550c2d087cda94db86973479 web/playwright.config.ts
a17881c77dd7129c9c4bf92ddd753622609f425e docs/PERFORMANCE.md
6c05966e8c932d01fcb26675a3d92e96287b7fad docs/foundation/MIGRATIONS.md
d77da3242e01c57a8d2745ca933b6cf77887bea7 docs/foundation/DATABASE-ISOLATION.md
17b23b97afc663dc6011ee276d80a71bd1eb42d4 docs/foundation/OPERATIONS.md
d1d459d891486c4f460f278934dea49b6f8fa77a management/v0/goals/G064-integration-test-isolation.md   (r0: 8829716e)
357b8f1b97a1524c8ff4d9d4b4d2b11e1edd6c06 management/v0/specs/G064-integration-test-isolation-spec.md   (r0: b53dc106)
```

r1 adds `infra/ci/validate-controls.sh` (the existing home of validate.sh adapter controls, groups A-H; `:1-48`) to the
bound set; the writer records its blob at S0. The 52 integration-tagged test files (112 top-level `Test*` functions)
and the 19 migrations `00001`..`00019` are bound by the tree at the release SHA; §4 lists every file. r2 adds
`core/internal/migrate/metadata.go` (goose store/version-0 facts) and `core/go.mod` (goose pin) as read-only facts
of §1.1/§3.1, and the untagged `core/internal/identity/*_test.go` and `core/internal/identity/workos/*_test.go` files
as the non-database part of `identity_it`'s per-package expectations (§4 #53, §5.2); the writer records their blobs at
S0. None of these files is edited.

## 2. Approach (one sentence each)

1. **Run-private template, canonical-name recreation (D-1 approved):** after the verbatim CLI steps leave `<db>` at the
   bound top version with the verified catalog (`run.sh:2529-2535`), the harness copies it once into a run-private,
   write-closed template `<db>_tpl` (`IS_TEMPLATE true, ALLOW_CONNECTIONS false`, CONNECT revoked) whose record is
   **bound to the tested source** and whose pristine state is **verified by a source-bound check the caller runs
   itself** (`db_pristine_verify`, §3.1: goose history against the bound migration set, catalog/role/fixture/ACL
   facts; the `migrations` suite additionally reuses its `B-MIGRATE-CLI`/`B-MIGRATE-CATALOG` rows, the standalone
   identity overlay and the private early-risk run do not depend on them), and, **before every independent Go
   invocation**, drops and recreates `<db>`
   from that template (`db_fresh`), restoring the database-level grants that cloning does not copy; the canonical name
   is kept so role derivation, baked role literals and migration SQL stay production-equivalent, and the only database
   a package can reach is its own fresh lifetime (state isolation, not distinct package credentials).
2. **Cluster-global state stays serialized and verified:** roles, memberships and role settings are shared by
   construction inside the run-private instance; every `db_fresh` compares the **complete** shared-catalog view
   (family roles with attributes, every membership crossing either endpoint, database-qualified settings, database
   ACL/owner/flags of `<db>` and `<db>_tpl`; §3.3) with the snapshot taken at template creation and refuses to proceed
   on any residue (fail closed, named delta), so the tests that mutate roles remain serialized proofs whose residue is
   detected, never silently inherited; no private role family, cluster or ACL rewrite.
3. **Lifecycle/recovery exceptions keep their transitions:** migration catalog/migration/lifecycle tests, the persist
   pairs, provisioning, the recovery drill and `TestMCPRecovery` run on a fresh canonical database at the start of their
   group and execute their own down/up/reset/invalidation on it; nothing is pre-migrated past the version a test
   expects, and nothing runs on the template.
4. **Browser seed consumers are decoupled (D-2 approved, no fallback):** the identity overlay's Go tests run inside a
   window where core and core-replica are stopped and no client is attached, on the canonical database, which is then
   dropped and recreated from a pristine copy **verified** (not presumed) before the tests, so `mcp-seed.sql` and every
   browser project consume a database with zero identity-test residue; a refusal anywhere in the window is a reported
   blocker, never a fallback to the shared-state path.
5. **Per-test evidence (D-5 approved with corrections):** every Go invocation moves to `go test -json` with the
   expected top-level tests derived from the bound source in the same container (`go test -json -list`), **keyed by
   `(package, test)` and covering every package the selector matches** (untagged unit tests and packages without test
   files included), and a small converter in the existing validation tools module (`gotestreport`) reconstructs the
   exact `-v` text for the unchanged PASS parsers, writes per-test rows (package, test/subtest, status, elapsed,
   tested source tree, command, selector, image, check) and a supplemental JUnit fragment, and fails closed on
   truncated, skipped, zero-test, missing-expected-test, unfinished, duplicate/contradictory, build-failed or malformed
   streams and on **any selected package without its terminal event**; the `go test` exit status is captured before
   any conversion and propagated unchanged. The database-isolation campaign (reset points, direct selectors) stays
   limited to the database-integration inventory (§4, §4.1); per-package expectations only make every selected test
   accountable.
6. **One new required check `B-IT-ISOLATION`** (suite `migrations`, D-3 approved) carries the template-record/source
   binding, stale/mismatch and dirty-baseline refusals, marker isolation, active-session refusal, template-closed,
   shared-catalog detector (attribute, cross-family membership, database-qualified setting), reporter fail-closed and
   report-presence controls as attributable evidence (T-1..T-8, §6); the two validate.sh adapter controls live in the
   existing `B-CI-CONTROLS` runner (§5.4).
7. **Retained evidence and safe cleanup by construction:** template and copies never leave the run-owned PostgreSQL
   container/volume; each create/drop is recorded in `db-lifecycle.tsv`; every invocation's per-test rows and its
   redacted readable log are exported to `REPORT_DIR` **as soon as the invocation ends**, an interrupted invocation is
   finalized as `interrupted` from the EXIT trap before `export_diagnostics`/`cleanup`, and `validate.sh` refuses a
   candidate whose declared invocation inventory is not complete; `db_fresh` never terminates sessions it does not own
   and refuses while any session is attached; the existing project teardown (`compose down -v`, `lib.sh:2193-2271`)
   removes everything on success, failure and interruption.

## 3. Database lifecycle design (R2, R5)

### 3.1 Helpers (new, `infra/validation/lib.sh`; explicit arguments; superuser `psql` through `compose exec` as `psql_super` does)

| Helper | Behaviour |
|---|---|
| `psql_maint DEST SQL` | `psql_super` with `-d postgres` (the instance's maintenance database): the only connection used for `CREATE DATABASE`, `DROP DATABASE`, `ALTER DATABASE` and the shared-catalog queries. Neither `<db>` nor `<db>_tpl` is ever the lifecycle connection target. |
| `db_pristine_verify DEST TREE STATUS_LOG` | **The source-bound pristine check every template caller runs itself** (r1 review item 2): `DEST` is the caller's configured copy (`psql_super`/`psql_maint` and `compose_migrate` work on it unchanged), `TREE` the copied source tree whose `core/migrations/*.sql` are the bound migration set, `STATUS_LOG` the output of the **caller's own run of the bound migrate image's `status` command** (`compose_migrate DEST run --rm migrate status`: the image embeds the same tree's migrations; `migrations` suite: one read-only extra `status` run after step 7-apply and the production refusals, `run.sh:2528-2534`; identity overlay: on the overlay project after `identity:migrate`, `IDENTITY.md:433-436`, inside the §3.5 window; private early-risk run: the same image on the private network). Checks, in order, any failure → FAIL naming the predicate (nothing is created or dropped): **(a) goose history, no row count:** let `bound` = the sorted `NNNNN` prefixes of `TREE/core/migrations/*.sql` and `MIG_TOP` = its maximum (19 here, derived); `STATUS_LOG` contains exactly one `migration\t<file>\tapplied\t<ts>` line per bound file in `bound` order, no other `migration` line, no `pending`, and `current_version\tMIG_TOP` (`runner.go:204-227`); the SQL counterpart on `app.goose_db_version` — `applied_sql` = `version_id` of the **latest row per `version_id` (highest `id`) whose `is_applied` is true** — equals `{0} ∪ bound` (the version 0 row goose inserts at table creation is required, `metadata.go:17-25`; a superseded down/reapply row, whether goose deletes it or marks it `is_applied=false`, never counts), `max(applied_sql)` = `MIG_TOP`, and the raw history (`id, version_id, is_applied, tstamp` ordered by `id`, including the `00019` reapply/rollback/re-apply rows of the verbatim steps) is **recorded verbatim, never counted or pinned**; **(b) metadata protection and roles (reuse):** `catalog_p DEST` (`run.sh:2444-2456`) = `objects=2 owners=<db>_owner extra_acl=0 table_privs=0 seq_privs=0` and `catalog_roles DEST` (`:2457-2469`) = the `r_expect` role facts (`app_owner=<db>_owner runtime_super=f runtime_bypassrls=f runtime_member_owner=f runtime_member_migrator=f runtime_owned=0 migrator_grant=inherit=f,set=t`, derived from `<db>`) with `app_objects_beyond_metadata=$MIG_RELATIONS` — the **existing** relation pin that `mig_step` carries as the literal `79` (`:2524-2528`, 88 after G020), hoisted into one `run.sh` variable read by both `mig_step` and this check (a mechanical move, not a new pin); **(c) catalog fingerprint recorded:** `fingerprint-catalog.sql` `catalog`/`roles` md5 of `<db>` (reuse `infra/recovery/sql/fingerprint-catalog.sql`; recorded for the template record and later `db_fresh` equality, not compared with a pin); **(d) empty fixture baseline, precise table set:** `baseline_tables` = every `pg_class` relation in schemas `app`/`audit` with `relkind IN ('r','p','m')` (ordinary tables, partitioned parents, materialized views) **except `app.goose_db_version`**; sequences (`'S'`), indexes (`'i'`), views (`'v'`), TOAST and composite relations are excluded (not row-bearing or not queryable); each is counted with `SELECT count(*) FROM <rel>` (a partitioned parent counts its partitions; a materialized view its stored rows) and **every count must be zero** (so `app.identity_provider_binding`, `app.users`, `app.organizations`, `app.sessions`, `audit.events` and any future row-bearing relation are covered; a migration that seeds a reference table is a Plan-amendment stop, §10); the `(relation, count)` table is recorded; **(e) database-level state:** `pg_db_role_setting` has no `<db>`-qualified row; the database ACL of `<db>` equals `10-roles.sh:1044-1045` (`PUBLIC` revoked; CONNECT for `<db>_migrator`, `<db>_runtime`, `<db>_retention`); **(f) sessions:** zero sessions on `<db>` other than the caller (`pg_stat_activity`). Output: a `pristine` row in `db-lifecycle.tsv` (caller, `MIG_TOP`, `bound` digest, history listing digest, fingerprint md5s, baseline table list digest) consumed by `db_template_create`. |
| `db_template_create DEST PROJECT TREE STATUS_LOG [NAME]` | **Source binding (recorded first):** `MIG_TOP` and `bound` from `TREE/core/migrations/*.sql` (19 at this base; **derived, no pin**); `mig_digest` = `sha256sum` over the sorted migration files; digests of `infra/postgres/init/10-roles.sh` and `infra/postgres/provision-roles.sql`; `impl_tree_hash TREE` (`lib.sh:168`) and `git -C "$REPO" rev-parse HEAD` plus a `tree=head|dirty` flag (`git status --porcelain` empty or not); the core-check image ID; `--base` (recorded as *baseline*, never as tested source). **Preconditions (any failure → FAIL, no template):** `db_pristine_verify DEST TREE STATUS_LOG` PASS — the only gate every caller satisfies by itself; **in the `migrations` lifecycle additionally** `B-MIGRATE-CLI` and `B-MIGRATE-CATALOG` recorded PASS for this run (they already ran there; their observed `catalog_roles` lines are reused as corroboration, `run.sh:2457-2470`), **never required of the standalone identity overlay or the private early-risk run**, which do not execute those checks and must not add a migrations/ops/full-suite run to create a template. Then `CREATE DATABASE <db>_tpl TEMPLATE <db>`; `ALTER DATABASE <db>_tpl IS_TEMPLATE true ALLOW_CONNECTIONS false`; `REVOKE ALL ON DATABASE <db>_tpl FROM PUBLIC, <db>_migrator, <db>_runtime, <db>_retention`. **Record** (`$WORK/db-lifecycle.tsv`, one `template` row + one `template-record` block): template name and flags, `MIG_TOP`, `bound`, the `status` listing and the raw goose history (verbatim), `mig_digest`, provisioning digests, tested tree hash/HEAD/flag, image ID, `baseline_tables` list and counts digest, `fingerprint-catalog.sql` `catalog`/`roles` md5 of `<db>`, the §3.3 shared-catalog snapshot, `pg_database_size`, duration. `NAME` (default `<db>_tpl`) lets T-1c create a throwaway refusal target. |
| `db_fresh DEST PROJECT PURPOSE` | **Preconditions (any failure → FAIL naming the cause; nothing dropped):** the template exists with the recorded flags; the template record's `mig_digest`, provisioning digests and tested tree hash **equal the bound checkout now** (a mismatch is a stale/source-mismatch refusal: the template cannot be reused for another source); the §3.3 shared catalog equals the snapshot (delta → FAIL naming roles/memberships/settings; a `<db>`-level database setting/ACL delta is recorded as `residue-removed-by-recreate`, since the drop destroys it, while any `<db>_tpl`-level delta is a refusal); zero sessions on `<db>` (any → FAIL; **no `pg_terminate_backend`**). `DROP DATABASE <db>` (plain, never `FORCE`); `CREATE DATABASE <db> TEMPLATE <db>_tpl`; `REVOKE ALL ON DATABASE <db> FROM PUBLIC; GRANT CONNECT ON DATABASE <db> TO <db>_migrator, <db>_runtime, <db>_retention` (exactly `10-roles.sh:1044-1045`). **Postconditions:** the §3.1 goose-history predicate (a) holds on the copy (`applied_sql` = `{0} ∪ bound`, `max` = `MIG_TOP`; the cloned history listing equals the record's), catalog fingerprint = template record, database ACL/owner = snapshot, every `baseline_tables` count = 0 (the recorded table set, re-enumerated and compared by name). Appends one `fresh` row (purpose, invocation name, timestamps, durations for PG-010). |
| `db_shared_catalog DEST` | The §3.3 shared-catalog query (roles, memberships crossing either endpoint, database-qualified settings, database ACL/owner/flags of `<db>` and `<db>_tpl`), used by `db_template_create`, `db_fresh` and T-5. |
| `MIG_RELATIONS` (variable, `run.sh`) | The existing relation-count pin (`79` today, `88` after G020) hoisted from the seven `mig_step` literals (`run.sh:2524-2528`, and the `${r_expect}79` text of `B-MIGRATE-CATALOG`, `:2536-2537`) into one file-scope variable, so `mig_step` and `db_pristine_verify` assert the same value and G020 updates one literal; the `B-MIGRATE-CATALOG` PASS text is unchanged. |
| `db_lifecycle_export` | Copies `db-lifecycle.tsv` to `REPORT_DIR` (atomic tmp+mv) after every row (not only at suite end) so an interrupted run retains the create/drop ledger. |

Identifiers are quoted with psql `:"var"` quoting and validated against `^[a-z_][a-z0-9_]{0,47}$` (`runner.go:27`);
`<db>_tpl` stays within the 63-byte limit because `_retention` (10) is the longest suffix the regex reserves and `_tpl`
is shorter.

### 3.2 What a recreated database needs (resolves owner item 1)

Copied by cloning: schema `app`/`audit` ownership by `<db>_owner`, every relation/function/policy/trigger/grant/default
privilege, the migration metadata, extensions. Not copied and restored by `db_fresh`: database ACL (CONNECT for the
three login roles, PUBLIC revoked), database owner (creator `postgres`, equal to the image's `POSTGRES_DB` owner), and
`ALTER DATABASE ... SET` rows (none are provisioned; the three test-time `track_functions` settings are reset in-test and
vanish with the drop, recorded by the §3.3 comparison if a test left one). Not copied and not needed: roles/memberships/
role `search_path` settings (cluster-global, already present, verified unchanged). Because the recreated database keeps
the canonical name, `migrate.Config` role derivation, `current_database()` in migration SQL and the baked `events_refuse`
literal all resolve to the existing family — **no change to `runner.go`, `10-roles.sh`, `provision-roles.sql`, any
migration, any policy or any production code.**

### 3.3 Cluster-global isolation and the complete shared-state detector (resolves owner item 2; r0 finding 3)

Roles, memberships, `ALTER ROLE ... SET` and `ALTER DATABASE ... SET` are shared or database-OID-bound and are never
cloned. The instance is run-private (project `g001-mig-$RUN_ID`), so the sharing boundary is the run, not the host.
Mutating tests remain serialized in the `migrations` suite and gain a detector. `db_shared_catalog` (new; the existing
`rls_prov_catalog` is left unchanged for PROV-1) returns one sorted, `C`-collated line set:

- `role:<name>:<rolcanlogin,rolinherit,rolsuper,rolbypassrls,rolcreaterole,rolcreatedb,rolreplication,rolconnlimit,
  rolvaliduntil>` for every role `LIKE '<db>\_%'` **and** for `postgres`;
- `member:<member>><role>:<inherit,set,admin>` for every `pg_auth_members` row **where the member OR the role is in the
  family** (so `GRANT pg_read_all_data TO <db>_runtime`, `GRANT <db>_owner TO postgres` or a membership in a
  predefined role is seen; counterparties outside the family are printed by name);
- `setting:<role|->:<database|->:<setconfig>` for every `pg_db_role_setting` row whose role is in the family **or**
  whose `setdatabase` resolves (by name, through `pg_database`) to `<db>` or `<db>_tpl` (`setdatabase` is an OID, so a
  row whose database no longer exists prints `?` and is a refusal);
- `db:<name>:<owner>:<datallowconn,datistemplate>:<aclexplode of datacl or acldefault>` for `<db>` and `<db>_tpl`.

`db_fresh` refuses on any role/member/setting delta or any `<db>_tpl` db-line delta; a `<db>` db-line delta is recorded
as residue removed by recreation. `B-IT-ISOLATION` proves the detector with three reversible superuser deltas, each made
and reverted inside the control after `db_shared_catalog` has reported exactly that delta: **T-5a attribute** (`ALTER
ROLE <db>_runtime NOINHERIT` → `INHERIT`), **T-5b cross-family membership** (`GRANT pg_read_all_data TO <db>_runtime` →
`REVOKE`; the counterparty is outside the family), **T-5c database-qualified setting** (`ALTER ROLE <db>_runtime IN
DATABASE <db> SET track_functions = 'pl'` → `RESET`). No extension is created by any test (`CREATE EXTENSION`/`ALTER
SYSTEM` absent from `core/internal` tests; the writer re-greps at S0 and records it). `TestRLSProvisioned`/PROV-1..6 keep
their standalone instances unchanged. Across fresh lifetimes the following stay proven by the unchanged tests rerun under
`db_fresh`: active-session refusal (T-3 on every lifetime class), runtime/template denial (T-4 plus RLS-ATTEST/AT-ATTEST/
OBJ-ATTEST negatives) and marker removal (T-2 repeated once after a role-mutating invocation, #50).

### 3.4 Lifecycle and recovery exceptions (resolves owner item 3)

| Group | Treatment |
|---|---|
| `TestMigrationLifecycle` (01→10 ordered subtests; `reset` to zero, up, synthetic DDL, down/up) | `db_fresh` before; it owns its copy; subtest order is its documented proof (selection boundary = the function) |
| Every `Test*Catalog`/`Test*Migration` in `core/internal/migrate` (`reset`/`runnerFor(subsetFS N)`/down/up inside) | `db_fresh` before each invocation; transitions unchanged; the catalog tests that start by `reset` prove the same thing from a fresh copy |
| Persist pairs (tenancy, business profile, policy, RLS, objects, audit): seed → postgres stop/start → verify | `db_fresh` before the **seed** only; the pair is one intentional group (IDs via `*_PERSIST_FILE`/`*_PERSIST_IDS`); verify never recreates |
| `TestRLSProvisioned`, PROV-1..6 | unchanged (own `rls_pg` instances, own passwords, removed by `rls_provision`) |
| `TestMCPRecovery` (applies `invalidate.sql`, expires every session) | runs inside `mcp-it` on a fresh copy; its position in the suite no longer matters; the "before the drill's reset" dependency note is retired |
| `recovery_drill` (RD-S..RD-O, MD-*; `TestRecoverySeed/Post/Verify/Repair` on source and target) | unchanged chain; `compose down -v` at its start removes template and copy (it is the last database consumer of the suite); the drill's own migrate-up produces its fresh source; its `rd_go` phases move to `go_it` for reporting only |
| `TestTenancyModel`, `TestBusinessProfileModel` (version ladder top→5, leaves 5) | `db_fresh` before each; the ladder always starts at the top (deterministic); `setup` gains one `t.Logf("TENANCY-SETUP version=%s")` line of the observed starting version and a fail-closed assertion that it is **the embedded migration set's top version or 5** (top derived from the embedded FS at test time, **not a literal `19`**, so G020's `00020` needs no edit of this line; unchanged behaviour otherwise; no ladder redesign) |
| identity overlay `identity_it` | §3.5 |

### 3.5 Identity overlay and browser seed consumers (resolves owner item 4; D-2 approved without fallback)

Consumers of the overlay database: `identity_it` (Go, all 35 identity tests), `mcp_seed` (`mcp-seed.sql`), the live core
and core-replica, and the browser projects `identity-*`, `domain-*`, `mcp-*`, `onboarding-*`, `membership-*`,
`enrollment-*`, `signup-*` plus `B-IDENTITY-PROBES` and `B-MCP-E2E` (handoffs). New `variants_identity` order after the
documented `identity:up` block has reported healthy: (1) `compose stop core core-replica` (the only database clients;
web, proxy, emulator and `identity-setup` hold no connection — asserted: zero sessions on `<db>`, else FAIL closed);
(2) **the overlay's own source-bound pristine check**: `compose_migrate "$D" run --rm migrate status` (the bound
migrate image already built for `identity:migrate`, `IDENTITY.md:433-436`; read-only) → `db_pristine_verify "$D"
"$D" <status log>` (goose history = `{0} ∪ bound` with `current_version` = `MIG_TOP`, `catalog_p`/`catalog_roles`
with `MIG_RELATIONS`, fingerprint recorded, every row-bearing `app`/`audit` table empty, ACL/settings, zero sessions —
§3.1) → `db_template_create "$D" "$project" "$D" <status log>`; **no `B-MIGRATE-CLI`/`B-MIGRATE-CATALOG` row is
required or consulted here** (those belong to the `migrations` suite and never run in `ops`; the overlay gives the
same guarantees through its own check, without adding a migrations/full-suite run) — core's startup namespace
inspection is read-only, so a healthy start leaves no row; this is checked, not presumed (there is no source basis to
assume the overlay is seeded before `mcp_seed`); (3) `identity_it` on the canonical `<db>` (its `TestMain` `Up` is a
no-op at the top version; per-test `resetIdentity` TRUNCATE and `releaseFixtureSlugs` unchanged; its per-package
expectations and terminal rule are §5.2); (4) `db_fresh` → `<db>` is recreated from the pristine template (identity residue gone
by construction); (5) `compose up -d --wait core core-replica` + `capture_after`; (6) `mcp_seed` is the **first writer**
after the fresh copy, then browser runs, probes, E2E as today. Result: the seed and every browser spec consume a database
whose only content is `mcp-seed.sql`; `identity_it` no longer precedes them by necessity, and the "the tests reset the
identity tables" dependency (`run.sh:5354`) is retired. **A refusal in this window (a lingering session, a failed stop/
restart, a dirty baseline) records `B-IDENTITY-IT` FAIL with the session inventory and stops the overlay; it is a
reported blocker, not authority to run the identity tests on the shared database and claim isolation** (the r0 fallback
sentence is withdrawn under D-2).

### 3.6 Resources, caps and cleanup (R5; resolves owner item 7)

- No new container, network, volume, port or bind mount: template and copies live in the project's PostgreSQL volume;
  the lifecycle log is one file under `$WORK`, exported after every row. The migrations-suite harness run keeps its
  existing lock/ports; the writer's private runs follow §6.
- Pools: one package at a time; the largest package pools are tenancy 48, identity 40, access 16+16, mcp 16+16 (§1.2),
  under PostgreSQL's default `max_connections` — unchanged.
- Cleanup on success/failure/interrupt: unchanged `cleanup()`/`compose down -v`; `db_fresh` leaves a database it could
  not drop in place and records `FAIL` with the session inventory (never terminates unknown sessions); the drill's own
  purge is unchanged. `db-lifecycle.tsv` names every create/drop with purpose and timestamps (ownership register).
- PG-010 observations collected without a benchmark campaign: template creation time, each `db_fresh` duration, number
  of recreations per suite (**50 in `migrations`, 1 in `ops`**, §4), template size (`pg_database_size`), zero-session
  waits, and the direct-selector campaign's per-selector totals (§4.1, once).
- G063's frozen qualification checkout and run are never read, written, locked or displaced by G064 work; the writer's
  private runs take their own pool slots and build slot under the existing inventory.

## 4. Inventory of every invocation and test (R1; resolves owner item 5; r0 finding 4)

Classes: **I** independently runnable by its direct selector from a fresh canonical database (prerequisites built by its
own `TestMain`/`setup`); **G** intentional group (serial, harness-orchestrated); **L** lifecycle/catalog proof that
executes migration SQL (independent, but must run on the canonical name and starts from a fresh copy). Treatment
`fresh ×N` = `db_fresh` immediately before each of the N named invocations; `group` = fresh at group start only;
`own` = own instance. **Reset points are counted per row and total 50 in `migrations` (#1-51) plus 1 in `ops` (#53).**

| # | Check ID | `run.sh` line | Package / selector | Top-level tests | Class | Treatment / reset points / retained dependency |
|---|---|---|---|---|---|---|
| 1 | B-IDENTITY-CATALOG | 5145 | identity `^TestIdentityCatalog$` | TestIdentityCatalog | I | fresh ×1 |
| 2 | B-MIGRATE-IT | 2560 | migrate `^TestMigrationLifecycle$` | TestMigrationLifecycle (10 ordered subtests) | L | fresh ×1 |
| 3 | B-TENANCY-CATALOG | 2645 | migrate `^TestTenancyCatalog$` | TestTenancyCatalog | L | fresh ×1 |
| 4-6 | B-TENANCY-MIGRATION | 2653, 2659, 2664 | migrate `^TestTenancyMigration$`; PersistSeed → restart → PersistVerify | TestTenancyMigration (L); TestTenancyPersistSeed, TestTenancyPersistVerify (G) | L/G | fresh ×2 (before Migration; before Seed = group start); `TENANCY_PERSIST_FILE` |
| 7 | B-TENANCY-IT | 2682 | tenancy `^TestTenancyModel$` | TestTenancyModel | I | fresh ×1; `TENANCY_CANARY_FILE`/`TENANCY_LOG_FILE` |
| 8-10 | B-BUSINESS-PROFILE-CATALOG/MIGRATION | 2709, 2717, 2723, 2728 | migrate Catalog; Migration; PersistSeed/Verify | TestBusinessProfileCatalog, TestBusinessProfileMigration (L); PersistSeed/Verify (G) | L/G | fresh ×3 (Catalog; Migration; Seed); `BUSINESS_PROFILE_PERSIST_FILE` |
| 11 | B-BUSINESS-PROFILE-IT | 2746 | tenancy `^TestBusinessProfileModel$` | TestBusinessProfileModel | I | fresh ×1 (formerly order-coupled with #7) |
| 12-14 | B-POLICY-CATALOG/MIGRATION | 2774, 2782, 2788, 2793 | migrate Catalog; Migration; PersistSeed/Verify | TestPolicyCatalog, TestPolicyMigration (L); PersistSeed/Verify (G) | L/G | fresh ×3; `POLICY_PERSIST_FILE` |
| 15 | B-POLICY-IT | 2825 | access `^(TestAccessPolicy|TestAccessFootprint)$` | TestAccessPolicy, TestAccessFootprint | I | fresh ×1 before the invocation; each test also by direct selector (§4.1); the two share one `TENANCY_LOG_FILE` append (kept) |
| 16 | B-POLICY-ORDERING | 2852 | access `^TestAccessOrdering$` | TestAccessOrdering (O-*, serial barriers) | I | fresh ×1; subtests not promised independent |
| 17 | B-RLS-CATALOG | 2906 | migrate `^TestRLSCatalog$` | TestRLSCatalog (N7 role mutation rolled back) | L | fresh ×1; shared catalog checked by the next `db_fresh` |
| 18-20 | B-RLS-MIGRATION | 3115, 3118, 3122, 3033 | migrate Migration; PersistSeed/Verify; `^TestRLSProvisioned$` ×PROV | TestRLSMigration (L); TestRLSPersistSeed/Verify (G, `RLS_PERSIST_IDS`); TestRLSProvisioned (own) | L/G/own | fresh ×2 (Migration; Seed); PROV-1..6 unchanged |
| 21-22 | B-RLS-IT | 3135, 3140 | access `^(TestRLSIsolation|TestRLSPool|TestRLSFootprint)$`; identity `^(TestRLSIdentityPool|TestRLSFirstSignInRace)$` | TestRLSIsolation, TestRLSPool, TestRLSFootprint; TestRLSIdentityPool, TestRLSFirstSignInRace | I | fresh ×2 (one per invocation); each of the five by direct selector (§4.1); `RLS_CANARIES`, `RLS-SCOPE`/`E-RLS` lines |
| 23 | B-RLS-ORDERING | 3156 | access `^TestRLSOrdering$` | TestRLSOrdering | I | fresh ×1 |
| 24-26 | B-OBJ-CATALOG/MIGRATION | 3275, 3285, 3288, 3292 | migrate Catalog; Migration; access PersistSeed/Verify | TestObjectsCatalog, TestObjectsMigration (L); TestObjectsPersistSeed/Verify (G, `OBJ_PERSIST_IDS`) | L/G | fresh ×3; private MinIO (`OBJ_ENV`) |
| 27-28 | B-OBJ-IT | 3304, 3306, 3246 | access `^(TestObjects|TestObjectsFootprint)$`; `^TestIsolationAttacks$`; `^TestObjectsSignHost$` | TestObjects, TestObjectsFootprint; TestIsolationAttacks; TestObjectsSignHost | I | fresh ×3 (one per invocation); TestObjects and TestObjectsFootprint also by direct selector; `OBJ_SIGN_HOST_*`; object-store bucket cleanup stays in-test (§11) |
| 29 | B-OBJ-ORDERING | 3328 | access `^TestObjectsOrdering$` | TestObjectsOrdering | I | fresh ×1 |
| 30 | B-OBJ-FAULTS | 3341 | access `^TestObjectsFaults$` | TestObjectsFaults | I | fresh ×1 |
| 31-33 | B-AUDIT-CATALOG/MIGRATION | 3357, 3375, 3378, 3382 | migrate Catalog; Migration; PersistSeed/Verify | TestAuditCatalog, TestAuditMigration (L); PersistSeed/Verify (G, `AUDIT_PERSIST_IDS`) | L/G | fresh ×3 |
| 34 | B-AUDIT-IT | 3393 | access `^TestAudit$` | TestAudit (AT-*, `AUDIT-INV`) | I | fresh ×1 |
| 35 | B-RETENTION-CATALOG | 3412 | migrate `^(TestRetentionCatalog|TestRetentionMigration)$` | both | L | fresh ×1; each by direct selector |
| 36 | B-RETENTION-IT | 3421 | access `^TestRetention$` | TestRetention (`DB_RETENTION_PASSWORD`) | I | fresh ×1 |
| 37 | B-MCP-CATALOG | 3442 | migrate `^(TestMCPCatalog|TestMCPMigration|TestMCPNamesCatalog|TestMCPNamesMigration)$` | four | L | fresh ×1; each of the four by direct selector |
| 38-39 | B-MCP-IT | 3589, 3595 | mcp `^(TestMCPTransport|TestMCPLifecycle|TestMCPAudit|TestMCPRecovery|TestMCPDiscovery)$`; access `^(TestMCPOrdering|TestMCPFootprint)$` | five; two | I | fresh ×2 (one per invocation); each of the seven by direct selector; `MCP_LOG_FILE`, `MCP_INVALIDATE_SQL_B64`; TestMCPRecovery no longer order-bound |
| 40-41 | B-DOMAINS-CATALOG/IT | 3455, 3460 | migrate `^TestDomainsCatalog$`; access `^TestDomains$` | TestDomainsCatalog (L); TestDomains (I) | L/I | fresh ×2 |
| 42-43 | B-BRANDING-CATALOG/IT | 3480, 3485 | migrate `^TestBrandingCatalog$`; access `^TestBranding$` | (L); (I) | L/I | fresh ×2 |
| 44-45 | B-ONBOARDING-CATALOG/IT | 3500, 3505 | migrate `^TestOnboardingCatalog$`; access `^TestOnboarding$` | (L); (I) | L/I | fresh ×2 |
| 46-47 | B-MEMBERSHIP-CATALOG/IT | 3524, 3529 | migrate `^TestMembershipCatalog$`; access `^(TestMembership|TestAuthority)$` | (L); TestMembership, TestAuthority (I) | L/I | fresh ×2; TestMembership and TestAuthority also by direct selector |
| 48-49 | B-ENROLLMENT-CATALOG/IT | 3547, 3553 | migrate `^TestEnrollmentCatalog$`; access `^(TestEnrollment|TestEnrollmentAdmission|TestEnrollmentRequests)$` | (L); three (I) | L/I | fresh ×2; each of the three by direct selector |
| 50-51 | B-SIGNUP-CATALOG/IT | 3569, 3574 | migrate `^TestSignupCatalog$` (renames/alters `<db>_provision_writer`, restores); signup `^(TestSignupProvisioning|TestSignupSlugRace|TestSignupContinuation)$` | (L); three (I; `-p 1` within the package: fault triggers on the shared catalog) | L/I | fresh ×2; shared catalog verified by the `db_fresh` after #50 (T-2 repeated there); each of the three signup tests by direct selector |
| 52 | B-RECOVERY-DRILL / B-MAINTENANCE-DRILL | 4240-4310, `rd_go` 3691-3700 | access `TestRecoverySeed`, `TestRecoveryPost`, `TestRecoveryVerify`, `TestRecoveryRepair` on source/target | four | G | unchanged chain (`RECOVERY_*`); no reset point; `go_it` reporting only |
| 53 | B-IDENTITY-IT (ops) | 5353-5364 | `./internal/identity/...` = three packages: `identity`, `identity/workos`, `identity/identitytest` (no test files) | **Database-integration inventory (the 35 `integration`-tagged tests of `identity`, the only ones with reset points and direct selectors):** TestIdentityCatalog, TestRLSIdentityPool, TestRLSFirstSignInRace, TestIssuanceGuardFailsClosed, TestMCPIdentityCreate/Activation/RenewalAndRevocation/LogoutAllInterleaving/Rename, TestCustomDomainRelay, TestLocalLogoutAndCSRF, TestLogoutAll, TestProviderLogout, TestDisableAndEnable, TestLoginStart, TestAssuranceIsAlwaysUnknown, TestCallbackSignup/Pinned/ParametersAndTransactions/Continuity/Refusals, TestNamespaceBinding, TestMappingIsByIssuerAndSubject, TestEnrollmentPendingCredentialRevocation, TestEnrollmentRevocationFaultRollback, TestConcurrentFirstSignIns, TestParallelIdenticalCallbacksConsumeOnce, TestSignInIsAtomic, TestSessionLifetimes, TestRevokeVersusRenewNeverResurrects, TestReplicasAndRestart, TestStoreOutageIsUnavailable, TestDisableVersusSignInRace, TestLogoutAllVersusConcurrentSignIn, TestLogsCarryNoValues. **Also selected and expected per `(package, test)`, not database-integration tests (no reset point, no direct selector):** `identity` 39 untagged tests (`allowlist_test.go` 3, `assurance` 2, `attest` 1, `credential` 4, `enroll_test.go` 6, `http` 11, `identity_test.go` 3, `limiter_copy` 2, `mcp_name` 1, `mq_eval` 3, `retire` 3) and `identity/workos` 21 untagged tests (`adapter` 8, `connect` 4, `jwks` 5, `token` 4); excluded by tag: `manual_qualification_test.go` (`workos_manual`), `workos/staging_test.go` (`workos_staging`). Expected at this base: `identity` 74, `workos` 21, `identitytest` 0 — **derived by `-list`, not pinned** | I (each of the 35; per-test `resetIdentity`); the 60 unit tests are database-free | §3.5 window: pristine check + template create + fresh ×1 after the invocation; package `TestMain` shared pool; **each of the 35 by direct selector (§4.1)**; every selected package must reach its terminal event (§5.2) |
| 54 | B-MCP-E2E (ops) | 5288 | mcp `-tags mcpe2e ^TestMCPE2E$` (live stack, handoff) | TestMCPE2E | G | unchanged; `go_it` reporting only |
| 55 | B-IT-ISOLATION (new) | after `db_template_create` | harness controls (§6) | — | — | — |

Every one of the 112 integration-tagged top-level tests appears above (migrate 36, identity 35, access 31, mcp 5,
signup 3, tenancy 2); `TestStagingContract` (`workos_staging`, hosted identity-staging job, no database) and
`manual_qualification_test.go` (`workos_manual`) are out of scope. **112 is the database-integration inventory**, the
set with reset points and direct selectors; `identity_it` (#53) is the one invocation whose package pattern also
selects untagged unit tests (60 at this base) and a package without test files — those are accounted for by the
`(package, test)` expectations of §5.2 (every selected test expected, every selected package terminal) and are **not**
added to the database-isolation campaign, which has no database behaviour to prove for them. Supported selection boundary: the **top-level test
function** by `-run '^Name$'` on a fresh canonical database with the invocation's environment; nested subtests of
ordering/concurrency/attack/audit/fault tests are documented proofs, not promised independent. Within-package serial
execution stays (`signup`: shared-catalog fault triggers; Go's default `-p 1` per package binary is unchanged; no
`t.Parallel` added or removed).

### 4.1 Direct clean-start selectors for every independently classified test (AC3; r0 finding 4)

**Classification totals:** I = 70 top-level tests (identity 35, tenancy 2, access 25, mcp 5, signup 3); L = 25 (the
`core/internal/migrate` catalog/migration/lifecycle tests); G = 11 (five migrate persist pairs = 10, the objects persist
pair in access = 2, the four recovery tests, `TestMCPE2E`; 17 functions in all); own = 1 (`TestRLSProvisioned`).
Reconciliation: 70 + 25 + 17 = 112.

**Documented direct selector** of an I- or L-class test `Name` in package `P`: `db_fresh` (postconditions recorded) then
`go_it ac3-<pkg>-<Name> ... P '^Name$'` with the invocation's own environment file/variables from §4 (the §4 row's
env, e.g. `RLS_CANARIES`, `OBJ_ENV`, `MCP_INVALIDATE_SQL_B64`, `DB_RETENTION_PASSWORD`). The expected set of that
selector (`go test -json -list '^Name$' P`) must be exactly `{(P, Name)}`.

**Proof campaign (batch 2b, once, run mechanically after `go_it`/`gotestreport` exist — §8 S4 — and recorded per
selector in `go-tests.tsv` and the Result):** every one of the **95**
I- and L-class tests executes its direct selector from a fresh copy — including each constituent of the multi-test
invocations #15, #21, #22, #27, #35, #37, #38, #39, #47, #49, #51 and **each of the 35 identity tests** (the three
identity tests that also appear in #1/#22 run once). Representative shuffles of the existing multi-test invocations are
*not* accepted as proof of their constituents. To bound cost without weakening the proof, the campaign may reuse one
owned, long-lived `core-check` container per package (`docker exec` per selector, so the Go build cache is reused);
the `go test` process exits between selectors, `db_fresh`'s zero-session precondition is asserted before each, and
each selector still produces its own `-list` expectation, events file, rows and redacted log. The campaign is linear
(95 executions), not combinatorial, and **stays the database-integration inventory**: the 60 untagged unit tests that
`identity_it`'s pattern also selects (§4 #53) are not campaign members (the coincidence that `identity_it` expects 95
`(package, test)` pairs at this base is unrelated to the 95 selectors). G-class groups run only by their documented group selector (the persist pair with
its restart; the drill chain; `TestMCPE2E` with its handoff) in batch 3/post-merge. Bounded representative order
evidence is retained in addition (§6 AC3 row): four reversed formerly coupled pairs, one repeat on fresh copies, two
`-shuffle=on` invocations with recorded seeds.

**Fixture rule:** fixtures are changed only if a direct selector exposes accidental coupling in that test's own
prerequisites; any additional test file beyond the one in §7 is an escalation (stop) before editing.

## 5. Per-test reporter, completeness, provenance and retention (R4; resolves owner item 6; r0 findings 1-2)

### 5.1 Invocation (`go_it`) and the expected-test inventory

- **`go_it NAME PROJECT IMAGE ITENV LOG PACKAGE RUN [docker args...]`** (new, `run.sh`) replaces the body of `rls_go`
  (which keeps its name and signature as a compatibility alias, D-6 approved, including G020's later callers) and the
  direct `docker_run_owned ... go test` sites (#1-16, 19, 52-54; `identity_it` and `rd_go` included). Inside the
  container, in this order: `go test -tags <tag> -json -list "$1" "$2" >/tmp/expected.json` (the expected top-level
  `Test*` names **from the bound source in the image, keyed by `(package, test)`**: in `-json` mode each listed name
  is an `output` event whose `Package` field names the package that owns it, and a package without test files yields
  its `[no test files]` line and a package-level `skip`; same container so the build cache is reused), then `go test
  -tags <tag> -count=1 -timeout <t> -json -run "$1" "$2"; rc=$?; <existing RLS-ACCESSLOG sed>; printf
  'GO-IT-EXPECTED %s\n' "$(<one 'package<TAB>test' pair per line, plus 'package<TAB>-' for a selected package with no
  expected test>)"; exit $rc` — the exit status is captured before any output handling and is the container's status,
  exactly as today (`run.sh:2878-2880`). Stdout (events + prefixed log lines + merged stderr) is saved as
  `$LOG.events`; the `-list` failure (build error) makes the invocation FAIL closed. The **selected package set** is
  the set of `Package` values in the `-list` output (every package the pattern matches, with or without tests); if the
  writer finds at S3 that `-list` events do not carry `Package`, the fallback is `go list "$2"` followed by one `-list`
  per package in the same container — recorded in the Result, same expectation format.
- **Provenance per invocation** (harness side, written into the rows): `run_id`, `check_name`, invocation `NAME`,
  `tested_tree` = `impl_tree_hash "$WORK/mig"` (the exact copied tree the image was built from; equals the checkout
  per `B-SOURCE-HASH`), `head_sha` = `git rev-parse HEAD` of `$REPO` with `tree=head|dirty`, `base` (comparison
  baseline only), `image_id` (core-check), `package`, `tag`, `selector`, `timeout`, the full in-container command, and
  the template record ID the fresh copy came from (`db-lifecycle.tsv` row).

### 5.2 Converter (`gotestreport`)

New command in the existing `infra/validation/tools` module (built into the validation-tools image beside
`contractcheck httpprobe logcheck identitysetup`, `infra/docker/validation-tools.Dockerfile:10`; run with `vtools`,
`run.sh:1700`, network none, on the fixed host directory pattern of `RLS_FIXED`, `run.sh:2866-2870`, so no per-run bind
mount is added). It reads `$LOG.events` and the provenance fields and writes:

- (a) `$LOG` = the exact `-v` text (concatenated `Output` fields; non-JSON lines such as `RLS-ACCESSLOG`, `E-*`,
  `GO-IT-EXPECTED`, build errors preserved verbatim in place), so `it_count` (`run.sh:2630-2634`), the 72 `--- PASS`
  greps, `rd_gopass` and the identity `^ok` parser keep working unchanged;
- (b) `$WORK/go-tests/<NAME>.tsv`: one `invocation` row (`status`, `rc`, selected package list, expected count per
  package, observed counts per package, provenance), one `package` row per selected package (`package, terminal
  = pass|fail|skip-no-test-files|missing, expected, passed, failed, unfinished`) and one row per test/subtest
  (`package, test, status, elapsed_s, log_file, line_from, line_to`);
- (c) a per-invocation JUnit fragment (no text nodes, no time on non-measured cases, `junit.sh` discipline).

**Completeness and termination rules, keyed by `(package, test)` (exit non-zero → harness records FAIL; the original
`rc` is reported with it):** **a selected package without exactly one package-level terminal event** — `pass` or
`fail`, or `skip` only for a package whose expectation is `package<TAB>-` and whose output is the `[no test files]`
line — is a truncated/interrupted/unterminated package and fails the invocation (one package's `ok` never completes
another: `identity`'s `ok` line does not terminate `workos`); `build-fail`/`[build failed]`/`[setup failed]` for any selected
package; zero `run` events in a package that expects at least one test; **an expected `(package, test)` (from
`GO-IT-EXPECTED`) with no `run` event in that package (omitted expected test)**; **a started test or subtest with no
terminal event (unfinished)**; **more than one terminal event for one `(package, test)`, or contradictory terminals
(e.g. `pass` and `fail`)**; any `skip` of a `Test*`; a `run` for a `(package, top-level name)` not in the expected
set (undeclared test, including a name expected in another package); a malformed JSON line between the first and the
terminal event; a package `pass` terminal with `rc != 0` (panic/`TestMain` exit); an expected set that is empty for
a package that has test files; a selected package absent from the events entirely. The harness combines `rc == 0 &&
converter ok && existing count checks` (`it_count`, the `10/10`, `-ge 10`, `^ok` rules all stay as the required
subtest controls); PASS text alone never passes. `identity_it` additionally requires, **per selected package**,
expected = observed terminal passes (`identity` 74 = 35 database-integration + 39 untagged, `workos` 21,
`identitytest` 0 with its `[no test files]` skip, at this base — derived, not pinned) and reports the per-package
numbers beside the unchanged `-ge 10`/`^ok identity` rules; **35 remains the database-integration inventory of
`identity`, not the total of enabled tests**, and the per-package rule adds accountability for the unit tests
without adding them to the database-isolation campaign (§4, §4.1).

### 5.3 Retained evidence, interruption and readable output (r0 finding 2)

- **Export as each invocation ends, not at suite end:** after conversion the harness appends the invocation's rows to
  `$WORK/go-tests.tsv`, regenerates `$WORK/go-tests-junit.xml`, and copies both plus `db-lifecycle.tsv` atomically
  (tmp + `mv`) to `REPORT_DIR`; it also writes `REPORT_DIR/go-tests/<NAME>.log` = the reconstructed `$LOG` passed
  through the **same generated-secret redaction and 1 MiB bound as `export_diagnostics`** (factored into
  `redact_log SRC DST`, `lib.sh:49-57`, behaviour unchanged for diagnostics) — so every successful and failing Go
  invocation's readable output survives cleanup, and rows reference `go-tests/<NAME>.log` plus line ranges, never a
  deleted `$WORK` file. Canaries used as fixture names appear exactly where they appear in today's exported `-v` logs
  (`migrate-it.log` is already staged); the final `B-SECRET-CANARY` and the `validate.sh` raw/candidate scans remain the
  fail-closed privacy gates for every file under `REPORT_DIR`, including `go-tests/`.
- **Declared inventory:** at suite start, before any invocation, the harness writes `REPORT_DIR/go-tests-expected.tsv`
  (`suite, invocation_name`) from a static per-suite array in `run.sh` (`GO_IT_migrations`, `GO_IT_ops`; the §4 names);
  the `-list` expectation of each invocation is the per-test counterpart.
- **Interruption:** the EXIT trap (`run.sh:109`) becomes `RUN_STATUS=$?; bg_stop_all; go_tests_finalize;
  export_diagnostics ...; cleanup`. `go_tests_finalize` converts every `$WORK/*.events` that has no `invocation` row
  yet (the container was stopped by `stop_harness_group`/signal) with status `interrupted`, keeping every per-test
  terminal row observed so far and marking started-unfinished tests `unfinished`, exports the rows/JUnit/redacted log
  as above, and records `FAIL`-equivalent status (an interrupted run is never green: `RUN_STATUS` is non-zero and the
  `complete` step refuses). Declared-but-never-started invocations stay absent from `go-tests.tsv` and are reported by
  `validate.sh` as missing. Local `--evidence-dir` retention (`36d1189b`) keeps the scanned `REPORT_DIR` as today.
- **No raw sensitive copy:** `.events` files never leave `$WORK` (they are not `*.log`, so `export_diagnostics` does not
  export them); only the redacted, bounded reconstruction is published.

### 5.4 Adapter (`infra/ci/validate.sh`, `infra/ci/validate-controls.sh`)

Narrow adapter changes (within the already proposed `validate.sh` path, as the review permits):

- `STAGE_REPORT` (`validate.sh:122-131`) gains `go-tests.tsv go-tests-expected.tsv go-tests-junit.xml
  db-lifecycle.tsv`; `do_stage` (`:1395-1410`) stages the `go-tests/` directory exactly like `diagnostics/`
  (`:1407-1410`: name pattern `^[A-Za-z0-9_.-]{1,120}$`, regular files only); both raw and candidate scans gate them.
- `complete` step (`:1250-1372`), **suite-aware:** when the executed set contains `migrations` or `ops`, (i) the four
  files exist, are regular, non-empty and well-formed; (ii) **every invocation declared in `go-tests-expected.tsv` for
  the executed suites has exactly one `invocation` row in `go-tests.tsv` with status `pass` or `fail` and `rc`
  recorded; `interrupted`, `unfinished`, duplicate, missing or undeclared invocations → `complete: ... ` finding, return
  14**; (iii) every invocation row's per-test rows are consistent with its counts and every `log_file` it names exists
  under `go-tests/`; (iv) the authoritative `results.tsv`/required-ID rules are unchanged and still decide PASS; the
  per-test files supplement them (D-5). When neither suite executed, the files are not required.
- `validate-controls.sh` (group D, beside J1-J8): **GT1 missing-report** — the stub harness executes `migrations`,
  writes PASS rows but no `go-tests.tsv` → `complete` 14 with the named finding and nothing published; **GT2
  failing/interrupted invocation** — the stub harness writes `go-tests-expected.tsv` with two invocations, one
  `invocation` row `fail` (with a failing per-test row and a `go-tests/<n>.log` that contains a planted generated secret
  and a FAIL row in `results.tsv`) and one `interrupted` → `complete` 14 naming the interrupted invocation, the FAIL row
  is reported, and the raw scan/redaction path withholds the planted secret (the real scanner, real adapter). The
  `B-CI-CONTROLS` expected-control list and count (`run.sh:1497-1507`) absorb the two new IDs (mechanical).

### 5.5 Unchanged

The authoritative `ci-evidence/junit.xml` (`junit.sh`, G050 fixed shape) is **not** changed (D-5); required-check
completeness stays the ledger. Browser outcomes: unchanged Playwright JSON reporter (`web/playwright.config.ts:172`,
`browser-*-results.json`, `browser_matrix`); no new runner or format.

## 6. Tests and validation (reuse / extend / new; the uncovered behaviour of each new item)

| Selector / control | Kind | Covers |
|---|---|---|
| `B-IT-ISOLATION` (new required ID, suite `migrations`, run right after `db_template_create`) | **new** (no existing control exercises template validity/source binding, the goose-history/baseline predicate, recreation, session refusal, the shared-catalog detector, reporter fail-closed or report presence) | **T-1** template record: the §3.1 goose-history predicate (`status` listing = one `applied` line per bound file, `current_version` = `MIG_TOP`; `applied_sql` = `{0} ∪ bound` by latest row per version; the raw history recorded, containing the version 0 row and the `00019` reapply/rollback/re-apply rows, never counted), `mig_digest`/provisioning digests/tested tree hash/image ID, `catalog_p`/`catalog_roles` with `MIG_RELATIONS`, fingerprint, shared-catalog snapshot, the `baseline_tables` list (every `app`/`audit` `relkind r/p/m` relation except `app.goose_db_version`, by name) with all-zero counts, flags `datistemplate/datallowconn`; **T-1a** history/table-set negatives: a copy of `app.goose_db_version` in the control's scratch schema with its version 0 row removed, with a superseded `is_applied=false` row for `MIG_TOP` as the latest row, and with a bound version missing each make the predicate (evaluated against that copy) report the named failure; the `baseline_tables` enumeration (schema list parameterised, default `app,audit`) run on a scratch schema holding one ordinary table with one row, one advanced sequence, one view and one index lists **only the table with count 1** (sequence, view and index excluded), then the schema is dropped — proving the table-set boundary without touching `app`/`audit`; **T-1b** stale/source-mismatch refusal: `db_fresh` against a copy of the record with one altered digest refuses without dropping; **T-1c** dirty-baseline refusal: a superuser row in `app.identity_provider_binding` (provider binding) on `<db>` makes `db_template_create ... <db>_ctl` refuse (no `<db>_ctl` created), row removed by the following `db_fresh`; **T-2** marker isolation: a marker Organization in `<db>`, `db_fresh`, marker absent, template fingerprint unchanged (repeated once after #50); **T-3** active-session refusal: a held `psql` session on `<db>` makes `db_fresh` FAIL without dropping and the session survives, then released; **T-4** template closed: `<db>_runtime`, `<db>_migrator`, `<db>_retention` logins to `<db>_tpl` refused (SQLSTATE recorded), `CREATE DATABASE ... TEMPLATE <db>_tpl` as `<db>_migrator` refused; **T-5a/b/c** shared-catalog detector: attribute, cross-family membership and database-qualified setting deltas each reported exactly and reverted (§3.3); **T-6** reporter fail-closed on fixture streams (truncated, skip, zero tests, build-fail text, malformed line, pass-with-rc-1, **omitted expected test, unfinished test, duplicate terminal, contradictory terminal, undeclared test, a multi-package stream whose second package has no terminal, an expected test observed in the wrong package**) → non-zero; a passing multi-package fixture (two test packages plus one `[no test files]` package, the §4 #53 shape) → per-package rows, the no-test-files `skip` accepted, and exact text reconstruction equality; **T-7** PG-010 observations line (template time, `db_fresh` durations so far, size, waits); **T-8** report presence/completeness so far: `go-tests.tsv`, `go-tests-expected.tsv`, `go-tests-junit.xml`, `db-lifecycle.tsv` present in `REPORT_DIR`, every invocation so far has its row and its `go-tests/<NAME>.log` |
| `infra/validation/tools/internal/gotest/report_test.go` | **new** (new code) | the converter on table-driven NDJSON fixtures: reconstruction byte-equality with a recorded `-v` log, per-test rows for pass/fail/skip/subtests, elapsed capture, provenance pass-through, every fail-closed class of §5.2 including the r1 fixtures (omitted expected, unfinished started, duplicate terminal, contradictory terminal) and the r2 fixtures (`(package, test)` keys: multi-package stream with one unterminated package, `[no test files]` package skip accepted, same test name expected in one package and observed in another, package absent from events) |
| Identity overlay window and private early-risk run: `db_pristine_verify` (§3.1, §3.5, batch 2a) | **new** evidence of an existing-helper reuse (`catalog_p`/`catalog_roles`/fingerprint) plus the new goose-history/table-set predicate | the standalone callers' own source-bound pristine proof: recorded `pristine` row (history listing, `MIG_TOP`, fingerprint, baseline table list) in `db-lifecycle.tsv` and named in the `B-IDENTITY-IT` PASS text (overlay) and the batch 2a log (private run); no dependency on `B-MIGRATE-CLI`/`B-MIGRATE-CATALOG` rows outside the `migrations` suite |
| `validate-controls.sh` GT1, GT2 (`B-CI-CONTROLS`) | **new** (the adapter's completeness/staging/scan path for the new reports is unexercised) | §5.4: missing report → 14; failing + interrupted invocation → 14, FAIL reported, planted secret withheld |
| `core/internal/tenancy/integration_test.go` `setup` | extend | `TENANCY-SETUP version=` line + fail-closed starting-version assertion against the embedded set's top version or 5 (the former run-order artifact becomes observable; no literal version) |
| Every existing integration selector (#1-54) | reuse, unchanged assertions | run through `go_it`+`db_fresh`; their existing count/skip/`E-*`/logcheck controls plus the new completeness rule are the proof that no assertion moved |
| AC3 direct-selector campaign (§4.1; batch 2b, after S3) | reuse (unchanged tests), new evidence | 95 direct selectors (the database-integration inventory only) from fresh copies, rows + logs per selector through `go_it`/`gotestreport` |
| AC3 order/repeat evidence (recorded commands and seeds in the Result) | reuse | reversed formerly coupled pairs: (a) `TestMigrationLifecycle` then `TestIdentityCatalog`; (b) `TestBusinessProfileModel` then `TestTenancyModel`; (c) `mcp-it` (TestMCPRecovery) then `domains-it`; (d) `signup-catalog` then `rls-catalog` with shared-catalog checks between; repeat: `TestAccessPolicy` twice on fresh copies; shuffle: `-shuffle=on` on #21 and #38, seeds recorded; no combinatorial campaign |
| `B-SHELL`, `B-CI-CONFIG`, `B-CI-CONTROLS`, `B-REQUIRED-CHECKS`, `B-SCOPE`, `B-SECRET-CANARY` | reuse | shellcheck of the harness delta, required-ID/profile counts, validate.sh controls, allow-list, evidence privacy |

**Local batches (writer; fetch/rebase + `bash infra/ci/docker-desktop-mounts.sh` before each; verify what each selector
actually runs before dispatch):**

1. *Stack-free:* `bash infra/validation/run.sh --base HEAD --allow-file <scope> lint lint-base lint-core unit-core
   ci-controls` (B-SHELL on `run.sh`/`lib.sh`, B-CI-CONFIG, B-CI-CONTROLS with GT1/GT2, B-UNIT-CORE for the one-line
   tenancy test change) plus the tools module tests through the validation-tools image build (one build slot) or, if no
   existing selector runs them, a network-free pinned `golang:1.27.3-alpine` container with read-only source,
   `-mod=readonly GOPROXY=off` and a disposable `GOCACHE` (owner rule: no harness lock).
2. *Early-risk private DB run (pool slot 1), in two parts:* **2a (S1)** a private pinned PostgreSQL
   (`pgvector/pgvector:0.8.7-pg17`, `10-roles.sh` mounted read-only, unique name/label/network, no published ports,
   generated passwords in a mode-600 env file; driven as a private `DEST` copy so `psql_super`/`psql_maint`/
   `compose_migrate` work unchanged) migrated to the top version with the bound migrate image (build slot), then **its
   own pristine check**: `migrate status` from that image → `db_pristine_verify` → `db_template_create` (**no
   `B-MIGRATE-CLI`/`B-MIGRATE-CATALOG` row exists in a private run and none is required**), then the T-1..T-5 controls
   driven by a disposable script that sources the new `lib.sh` helpers; **2b (S4, after the reporter exists)** the
   §4.1 direct-selector campaign and the AC3 order/repeat/shuffle selectors through `go_it` and the core-check image on
   that network, plus T-6..T-8; images are built from the bound checkout, never pulled or substituted; every selector's
   provenance rows name that checkout's tree hash.
3. *`migrations` suite once* (`bash infra/validation/run.sh ... migrations`, harness lock, PORTS_MIG). Recorded
   technical reason for this umbrella: every one of its 52 invocations changes lifecycle (`db_fresh`) and reporter
   (`go_it`), so the smallest selector containing the affected IDs is the suite (D-4 approved; no blanket `ops`).
4. *Identity overlay focused (pool slot):* the documented `identity:up` block, the §3.5 window (its own `migrate
   status` + `db_pristine_verify` + `db_template_create`, no `migrations`-suite row consulted), `identity_it` with the
   per-package expectations (`identity`/`workos`/`identitytest`), `db_fresh`, core restart, `mcp_seed`, then
   `identity:browser` restricted to `--project mcp-light --grep "M-LIST"` (the seeded old-host row proves the seed
   consumer) — the `ops` suite itself runs post-merge.
5. *Hosted:* MR fast pipeline automatic; the changed paths map to the foundation/full profile, so the optional MR full
   gate stays manual; the coordinator-owned local full qualification runs on the exact merged staging SHA.

**Manual/operator feedback:** none required by Goal/Spec; writer self-review (§8 S7) and Codex source review are the
human-equivalent reviews.

## 7. Files and modules

| Path | Change |
|---|---|
| `infra/validation/lib.sh` | **new** `psql_maint`, `db_pristine_verify` (goose-history predicate, `catalog_p`/`catalog_roles` reuse, fingerprint, `baseline_tables` enumeration, ACL/settings, sessions), `db_template_create`, `db_fresh`, `db_shared_catalog`, `db_lifecycle_export`, `redact_log` (factored from `export_diagnostics`, behaviour unchanged), `go_tests_finalize` (§3.1, §5.3); `catalog_p`/`catalog_roles` stay in `run.sh` and are called from there (no move) |
| `infra/validation/run.sh` | `go_it` (§5.1, `-json -list` expectations keyed `(package, test)`); `rls_go` → alias; `MIG_RELATIONS` hoisted from the `mig_step` literals (§3.1); one read-only `migrate status` run → `db_pristine_verify` → `db_template_create` after `mig_step 7-apply` and the production-refusal checks (`:2529-2535`); `db_fresh` at the 50 reset points of §4; `GO_IT_migrations`/`GO_IT_ops` declared inventories; `it_isolation` check function → `B-IT-ISOLATION`; `variants_identity` window with its own `migrate status` + `db_pristine_verify` (§3.5); `identity_it` per-package completeness rule (§5.2); `identity_it`/`rd_go`/`obj_sign_host`/`mcp_e2e` through `go_it`; EXIT trap line (`:109`) gains `go_tests_finalize`; retire the three order comments (`:2546, 2597, 3552-3556, 5354`); PASS texts of B-MIGRATE-CATALOG unchanged; `B-CI-CONTROLS` expected list +GT1/GT2 |
| `infra/validation/required-checks.tsv` | +1 row `B-IT-ISOLATION	migrations`; header counts (full 122 → 123 per the header's profile line; fast/integration/staging unchanged; row count 129 → 130) |
| `infra/validation/tools/cmd/gotestreport/`, `infra/validation/tools/internal/gotest/` | **new** converter + tests (§5.2); `infra/docker/validation-tools.Dockerfile:10` adds `./cmd/gotestreport` to the build list |
| `infra/ci/validate.sh` | `STAGE_REPORT` += four names; `go-tests/` directory staging; suite-aware `complete` rules (§5.4) |
| `infra/ci/validate-controls.sh` | GT1, GT2 in group D (§5.4) |
| `core/internal/tenancy/integration_test.go` | `setup`: version line + derived-top assertion (§3.4) |
| `docs/foundation/OPERATIONS.md`, `MIGRATIONS.md` (`:140`), `DATABASE-ISOLATION.md`, `IDENTITY.md` (overlay order), `BOOTSTRAP.md` (if it lists suite evidence), `docs/PERFORMANCE.md` PG-010 (observations, status) | concise runbook/evidence updates (new report names, lifecycle ledger, interruption semantics) |
| `management/v0/results/G064-integration-test-isolation-result.md` | **new** at execution |

Not touched: `core/internal/migrate/runner.go`, every migration, `10-roles.sh`, `provision-roles.sql`, every policy/grant,
`infra/ci/junit.sh`, `.gitlab-ci.yml`, `infra/ci/lib.sh` profiles, Compose files, `web/**` (Playwright reporter reused as
is), `infra/identity/**`, `infra/recovery/**`, every other test file (assertions, counts, `t.Parallel`, skips unchanged),
Codex-owned Goal/Spec/roadmap/progress, G020 and G063 artifacts and branches.

**Database/schema changes:** none (test-only databases inside a disposable instance; no migration, role, grant or
policy change). **Contracts/public API/MCP:** none. **Dependent callers/pins:** `rls_go` callers (all keep working
through the alias); `it_count` and the `--- PASS` greps (read the reconstructed text, unchanged); `B-CI-CONFIG` reads
profile counts from the TSV (`run.sh:1481-1494`; the header prose is updated); `B-CI-CONTROLS` count/list
(`run.sh:1503-1506`); `export_diagnostics` (`$WORK/*.log` — the reconstructed logs keep the `.log` suffix; `.events`
files are excluded by pattern); `stage_one`/`do_stage` name validation; G020's planned `labels_checks` (§12).

## 8. Sequence (one writer; checkpoint slices)

S0 reconcile bound inputs at the release SHA (exact-delta rule), re-verify the grep facts (no test/migration `CREATE
EXTENSION`/`ALTER SYSTEM`; no migration-time `INSERT`; no `MATERIALIZED VIEW`/`PARTITION BY` in `core/migrations`;
the identity package/tag inventory of §4 #53; tools-module test selector; validation-tools image build list; EXIT-trap
execution on `stop_harness_group`), record them →
**S1 early-risk slice (real PostgreSQL lifecycle proof first):** `lib.sh` helpers (`db_pristine_verify` included) +
`MIG_RELATIONS` hoist + `B-IT-ISOLATION` T-1/T-1a/T-1b/T-1c/T-2..T-5 proven on the private instance (batch 2a: the
private run's own `migrate status` → pristine check → template) and PG-010 observations — the highest-risk item
(canonical-name recreation, goose-history/baseline predicate, source binding, grant restoration, session refusal,
complete shared-catalog detector) is proven before any other edit → S2 wire the `migrations`-suite `status` run +
`db_pristine_verify` + `db_template_create`, `db_fresh` into the 50 reset points, the overlay window (§3.5, its own
pristine check) and the tenancy setup line — **no direct-selector campaign yet: it needs `go_it`** → S3 reporter:
converter + tests, `go_it` with `-json -list` `(package, test)` expectations and the per-package terminal rule, all
invocations, `identity_it` per-package rule, per-invocation export, `go_tests_finalize`, `validate.sh` adapter,
GT1/GT2, T-6/T-7/T-8 → **S4 direct-selector campaign (batch 2b), run mechanically now that `go_it`/`gotestreport`
exist:** the 95 §4.1 selectors + AC3 reversed pairs/repeat/shuffle on the private instance → S5 required-checks, docs,
PG-010 → S6 batches 1, 3, 4 (finish, fix, rerun failed + affected, then advance) → S7 self-review (no assertion/count/
skip change outside the one tenancy line; no production path touched; privacy of new evidence incl. `go-tests/`;
cleanup receipt) → S8 fetch/rebase/reconcile, safekeeping push, Result → Codex source review → push/MR to `staging`
(auto-merge after gates; **G063 final ACCEPT precedes this merge**) → coordinator-owned merged-SHA local full
qualification (one run may also satisfy pending Goals on the same candidate) → acceptance cleanup receipt. Checkpoint
after S1 and after S3 (context 150k/200k policy).

## 9. Delegation, ownership, resources, scope

One implementation writer (Opus 5.5 high; isolated worktree `g064-impl`). Optional, at the writer's request: one
read-only Fable 5.1 high review of §3.1/§5 (SQL lifecycle and fail-closed reporter) after S1/S3; no second writer. At
most two concurrent private test runs (the writer's private instance and, when needed, the identity overlay stack), one
image build at a time, no harness lock for qualifying private runs, images built from the bound checkout only. G063's
frozen qualification checkout and G020's `g020-impl` are never touched.

Scope allow-list (ERE; implementation may touch only these paths):

```text
^infra/validation/(run|lib)[.]sh$
^infra/validation/required-checks[.]tsv$
^infra/validation/tools/cmd/gotestreport/.*[.]go$
^infra/validation/tools/internal/gotest/.*[.]go$
^infra/validation/tools/go[.](mod|sum)$
^infra/docker/validation-tools[.]Dockerfile$
^infra/ci/validate[.]sh$
^infra/ci/validate-controls[.]sh$
^core/internal/tenancy/integration_test[.]go$
^docs/foundation/(OPERATIONS|MIGRATIONS|DATABASE-ISOLATION|IDENTITY|BOOTSTRAP)[.]md$
^docs/PERFORMANCE[.]md$
^management/v0/plans/G064-.*$
^management/v0/results/G064-.*$
```

## 10. Codex rulings (r0 review), r1 disposition, risks, escalations

| ID | Ruling (r0 review) | r1 disposition |
|---|---|---|
| **D-1** | **Approved:** serial canonical-name recreation from a run-private, write-closed template; each independent invocation gets a fresh disposable lifetime; state isolation, not distinct package credentials or parallel package databases; no production configuration/role derivation change | §2.1, §3.1-3.3 as approved; the r0 alternative (b) (renamed clones with a `ConnectDatabase` override) is **withdrawn** and not retained as an option |
| **D-2** | **Approved:** stop/copy/test/reset/restart/seed window with verified pristine baseline and no active clients; a refusal is a reported blocker, not fallback authority; startup inspection is read-only; no presumption that the overlay is seeded before `mcp_seed` | §3.5 rewritten: baseline verified by `db_template_create`, refusal = FAIL + stop, **fallback sentence removed** |
| **D-3** | **Approved:** one `B-IT-ISOLATION` required check in `migrations`; parity and fail-closed evidence preserved | §6 (T-1..T-8), §7 TSV +1 |
| **D-4** | **Approved:** one focused pre-merge `migrations` umbrella run; corrected independent-selector evidence required; no blanket `ops`/repeated full suite | §6 batch 3 unchanged; the independent-selector evidence is §4.1 (batch 2) |
| **D-5** | **Approved:** supplemental TSV/JUnit with the authoritative JUnit/check ledger preserved, subject to completeness, provenance and retained-output corrections | §5.1-5.4 (expected-list completeness, tested-tree provenance, per-invocation retained redacted logs, interruption finalization, suite-aware `complete`) |
| **D-6** | **Approved:** `rls_go` kept as a compatibility alias including G020's later callers | §5.1, §12 |

- **Risks:** `DROP DATABASE` while a client lingers (fail closed, recorded; the Go test process has exited and
  `--rm` containers are gone before `db_fresh`; the §4.1 long-lived campaign container holds no connection between
  selectors); core-replica restart timing in the overlay window (`--wait`); `go test -json` interleaving of stderr
  (merged lines are preserved verbatim and never counted as results); JSON `Output` reconstruction must be byte-exact
  for the 72 text greps (T-6 equality control); `go test -list` doubling build time if the cache is not shared (same
  container, measured in PG-010 observations); `B-CI-CONFIG`/`B-CI-CONTROLS` count prose; a `validate.sh` allow-list
  typo would silently drop a new file (**T-8** asserts presence in `REPORT_DIR` during the run and the suite-aware
  `complete` step plus GT1 refuse a candidate without it — r0 wrongly cited T-7, which is the PG-010 observations line);
  EXIT-trap behaviour under `stop_harness_group` (verified at S0; if the trap cannot run on the delivered signal, the
  finalizer is also invoked from `bg_stop_all`'s path and the writer records the actual mechanism).
- **r2 risks:** goose's down semantics on the metadata table (row deleted or marked `is_applied=false`) are not
  inspected on the host (Docker-first, no module cache) — the predicate uses the latest row per version and is correct
  under either, and T-1a proves both negative shapes; `go test -json -list` package attribution is verified at S3 with
  the documented per-package fallback; the `MIG_RELATIONS` hoist must leave the `B-MIGRATE-CATALOG` PASS text
  byte-identical (`B-REQUIRED-CHECKS`/`B-CI-CONFIG` unaffected).
- **Mechanical pins under standing authority:** `required-checks.tsv` header counts, `run.sh` line-pinned comments, the
  `STAGE_REPORT` list, `B-CI-CONTROLS` expected list, PASS-text wording that mentions ordering, the `MIG_RELATIONS`
  hoist of the existing `79` literal.
- **Stops (return to Codex, never autonomous):** any need to touch `runner.go`, a migration, `10-roles.sh`,
  `provision-roles.sql`, a policy/grant, `junit.sh`, `.gitlab-ci.yml`, CI profiles, a test assertion/count/skip, or to
  `pg_terminate_backend` a session the harness does not own; a §3.5 window refusal that cannot be cleared by the
  harness's own owned stop/restart; any package that cannot start from a fresh canonical copy without a schema/role
  change; a shared-catalog delta that is not restored by its test; a direct selector (§4.1) that fails from a fresh copy
  for a reason outside that test's own prerequisites (fixture coupling in another file); a migration that seeds rows at
  migration time (baseline definition change); evidence that a private run needs a third concurrent slot; PG-010
  observations materially above expectation (template or recreation taking more than the invocation it precedes).

## 11. Out of scope (explicit)

G061 parallel full-suite orchestration or any concurrent package execution; Dagger/Bazel/new scheduler/framework;
production migration, role, grant, policy, schema or application changes; changing how roles derive from the database
name; private role families or clusters; object-store (MinIO) fixture isolation beyond the existing in-test cleanup;
`ops`/`rename`/`off` suite stacks (live services on their own disposable copies, not integration packages); browser
test parallelism or reporters; the hosted CI job definitions and the authoritative JUnit report shape; the retention
repair at `36d1189b` (already landed); benchmarking campaigns; weakening or removing any assertion, count, no-retry or
commit-fault semantics; `TestStagingContract`/manual WorkOS qualification; the r0 alternative D-1(b) and the r0 D-2
fallback (withdrawn).

## 12. G020 interaction and migration `00020` reconciliation (no G020 artifact is edited; r0 finding 5)

Read-only facts from `origin/g020-impl` `df2c46f4` (`core/migrations/00020_labels.sql`, 1187 lines, absent from
staging and from every G064 template by construction):

- **Precondition/effects:** `00020` refuses unless `current_version` is exactly 19 (`:79-80`), checks the version-19
  `audit.events_attribute()` body md5 and FORCE flags (`:120-126`), adds `app.label_definitions`/`app.label_assignments`
  (RLS enabled + FORCE, `:133-196`), two audit capture triggers (`:238-280`), per-role grants to **existing**
  `<db>_runtime`/`policy_reader`/… resolved from `current_database()` (`:42,56-61,287,327-331,692-698`), replaces
  `audit.events_attribute()` and its constraints (`:372-374,440`, new md5 at `:806`), and **creates no role, no
  PUBLIC/retention/migrator grant, no `ALTER ROLE`/`ALTER DATABASE`, no SECURITY DEFINER** (`:18`). Relation count at
  the top version becomes 88 (`labelRelations` = 9; G020's `metadataThroughVersion20`). Its `Down` lifts FORCE on the
  label tables and `audit.events` inside its own transaction (`:843-845`).
- **Template/clone consequence:** none for the design. The template is created by the verbatim CLI steps at the bound
  top version; `MIG_TOP`, `mig_digest`, fingerprint, baseline row counts and the shared-catalog snapshot are **derived
  at creation** (§3.1), so with `00020` in the bound set the template is version 20 with 88 relations and the same
  role family **without any G064 pin edit**. The only fixed version/relation pins remain the existing `mig_step`/
  `B-MIGRATE-CATALOG`/`B-MIGRATE-CLI` texts (`run.sh:2525-2535`, `r_expect` 79) and the Go-side `metadataThroughVersion*`
  tables — all already updated on `g020-impl` (its own exact later update: `current_version 20`, `00020` applied/
  rolled back lines, 88). The shared-catalog detector sees no delta from `00020` (no role/membership/setting change).
- **Tenancy starting version:** G020 adds a 20→19 ladder step to `tenancy/integration_test.go` `setup` (+12 lines at
  `:144-155` on `g020-impl`); G064's one-line assertion compares against the embedded set's top version, so both edits
  coexist (textual rebase only, same function; no semantic conflict).
- **G020's `setupIT`/`setupMCP`/`setupRecovery` `current_version 20` expectations** (`g020-impl` deltas to
  `access/integration_test.go:233-236`, `mcp/env_integration_test.go:253`, `recovery_integration_test.go:92`) are
  satisfied by a fresh copy at the bound top version (their `Up` is then a no-op), exactly as the version-19 ones are
  today.
- **What the supervisor/writer reconciles when G020 resumes (after G064 final ACCEPT and RC7 qualification):**
  (1) G020 rebases its `run.sh` line pins (shifts from `go_it`/`db_fresh`/`it_isolation`/inventory arrays) and
  updates `MIG_RELATIONS` once (`79` → `88`) instead of the several `mig_step` literals it edits today; the
  goose-history predicate derives `bound` = `00001..00020` and `MIG_TOP` = 20 without any G064 edit;
  (2) its three `labels_checks` invocations (B-LABELS-CATALOG/IT/ORDERING, through the `rls_go` alias) each get a
  `db_fresh` before them (independent: catalog, IT, ordering) and are appended to `GO_IT_migrations`;
  (3) `required-checks.tsv` header/profile counts absorb G064's +1 before G020's +3 (G020's recorded "127 rows" is
  reconciled against the then-current TSV — 129 today, 130 after G064);
  (4) no G020 test-code change beyond its own rebase: `labels_integration_test.go` starts from a fresh canonical
  database like every other access test;
  (5) `B-CI-CONTROLS`/`B-CI-CONFIG` prose counts;
  (6) the G063/G064 full-qualification evidence is not reinterpreted as G020 coverage; G020 keeps its B5 ledger
  (attempt 1 reported) and S1 checkpoint. The supervisor records these in G020's reconciliation at resumption; this
  section edits nothing of G020.

## 13. Acceptance-criteria mapping

| AC | Evidence (where) |
|---|---|
| AC1 complete inventory | §4 rows #1-55 (every selector, required ID, class, treatment, **reset points = 50 + 1**, retained dependency); §4.1 classification totals reconciled to 112 (the database-integration inventory); #53's three packages with their untagged tests and the no-test-files package expected per `(package, test)` (§5.2) without joining the database campaign; §3.4 exceptions; the Result accounts for each row |
| AC2 correct template, private copies, fail-safe controls, cross-package containment, meaningful role negatives | `B-IT-ISOLATION` T-1/T-1a/T-1b/T-1c/T-2/T-3/T-4/T-5a-c (source-bound template record with the goose-history predicate — version 0 row, applied set = bound files by latest row per version, raw history recorded not counted — and the precise `relkind r/p/m` baseline table set; history/table-set negatives; stale/mismatch and dirty-provider-binding refusals; marker isolation; session refusal; template closed to all three login roles and to cloning by migrator; complete shared-catalog detector incl. cross-family membership); `db_pristine_verify` run by the `migrations` suite, the identity overlay window and the private run themselves (§3.1, §3.5, §6) with `B-MIGRATE-CLI`/`B-MIGRATE-CATALOG` reused only where they run; `db-lifecycle.tsv`; unchanged RLS-ATTEST/AT-ATTEST/OBJ-ATTEST and privileged-assertion negatives rerun under the new lifecycle |
| AC3 independent groups by direct selector; repeat/shuffle/reversed | §4.1 campaign (batch 2b, S4, after `go_it` exists): **every** I/L-class test (95) by its documented direct selector from a fresh copy, rows per selector; §6 AC3 row: four reversed formerly coupled pairs, one repeat, two shuffled invocations with recorded seeds; batch 2b and 3 logs |
| AC4 lifecycle/recovery/ordering/attack/audit/fault proofs intact | #2-6, 8-10, 12-14, 16-20, 23-26, 29-33, 35, 37, 52 pass under `fresh`/`group` with unchanged subtest counts and complete expected lists; `recovery_drill` chain unchanged; `B-RLS-MIGRATION` PROV-1..6 unchanged |
| AC5 per-test outcomes retained; failures cannot become PASS | `go-tests.tsv`/`go-tests-expected.tsv`/`go-tests-junit.xml`/`go-tests/<NAME>.log` in `REPORT_DIR`, exported per invocation; T-6 fail-closed classes (incl. omitted expected, unfinished, duplicate/contradictory, unterminated second package, wrong-package test) + converter unit tests; expected set keyed `(package, test)` from `-json -list` on the bound source with terminal completion required of **every** selected package (`identity_it`: `identity`, `workos`, `identitytest`), the per-package `identity_it` rule beside the unchanged `-ge 10`/`^ok` rules; tested-tree/HEAD/image provenance per row; exit status captured before conversion (`go_it`); `go_tests_finalize` on interruption; suite-aware `complete` + GT1/GT2; `B-REQUIRED-CHECKS`/`B-SECRET-CANARY`/`validate.sh` scans unchanged and fail closed; Playwright JSON unchanged |
| AC6 cleanup proves only owned resources removed; evidence retained; caps respected | `db-lifecycle.tsv` (every create/drop, purpose, timestamps, exported per row); T-3 (no termination, no drop with a live session); interrupted-run evidence retained before `cleanup` (§5.3, GT2); existing cleanup inventory and receipt; run inventory shows one build slot, at most two runs, G063's frozen checkout untouched |
| AC7 focused pre-merge checks, reviews, hosted fast, authorized merge, merged-SHA full qualification, Result, runbook, cleanup | batches 1-4; Codex source/Duo/hosted fast gates; MR/auto-merge after gates with G063 final ACCEPT first; coordinator-owned full run on the merged SHA; Result ledger; docs of §7; cleanup receipt; Codex final ACCEPT separate |

## 14. Readiness gates remaining (restated)

Codex Plan ACCEPT of **r2** → accepted Plan imported to the integration branch → bound inputs reconciled at the release
SHA → supervisor execution release (attempt 1) → implementation (S0-S8) → Codex source ACCEPT → **G063 final ACCEPT** →
MR/auto-merge after gates → merged-SHA local full qualification → cleanup → Codex final ACCEPT → **RC7** cut and full
qualification → G020 reconciliation/resumption (§12) → combined staging preview → RC8 (if needed) → G057 → G021 → RC9.
**Nothing in this document starts, builds, tests or releases anything.**

## Revision history

- r0, 2026-10-21 — initial Plan on `origin/staging` `26240112` (planning worktree `agent-<id>`, branch
  `g064-plan`; Fable 5.1 planner; read-only inspection only; zero of B5 spent).
- r1, 2026-10-21 — one bounded revision answering the r0 Codex review (not included)
  (CORRECTION LOOP; D-1..D-6 approved subject to evidence corrections). Branch `g064-plan` merged with `origin/staging`
  `5a52fe85` (management-only delta; code bindings unchanged); worktree `agent-<id>`; Fable 5.1 planner;
  read-only inspection only (`git show` of `origin/g020-impl` for §12); zero of B5 spent. No architecture escalation,
  no wholesale replan; prior r0 text retained where unchanged. Changes per finding:
  - **Finding 1 — Complete test selection and termination accounting (R4/AC5):** `go_it` derives the expected
    top-level list from the bound source (`go test -list`, same container); the converter rejects omitted expected
    tests, unfinished started tests, duplicate/contradictory terminal events and undeclared tests, with four new
    fixtures; existing subtest/count controls kept; `identity_it` requires expected = observed; every row records the
    tested tree hash (`impl_tree_hash` of the copied tree), HEAD + clean/dirty flag, image ID, command, selector and
    template record, with `--base` recorded only as the comparison baseline (§5.1-5.2, §6, §13 AC5).
  - **Finding 2 — Retained evidence and interruption handling (AC5/AC6):** per-invocation export of rows/JUnit/ledger
    and a redacted bounded readable log to `REPORT_DIR/go-tests/` (no row points at a deleted work file); declared
    invocation inventory; `go_tests_finalize` in the EXIT trap marks interrupted/unfinished evidence before
    `export_diagnostics`/`cleanup`; `validate.sh` suite-aware `complete` rules and `go-tests/` staging; GT1
    missing-report and GT2 failing/interrupted controls in `validate-controls.sh`; T-7 corrected to mean the PG-010
    line and the presence control renumbered T-8 (§5.3-5.4, §6, §7, §10 risks).
  - **Finding 3 — Verified template baseline and complete shared-state detector (R2/AC2):** template record bound to
    the migration set digest, provisioning digests, tested tree hash and image; approved baseline defined (goose rows =
    bound count, every other `app`/`audit` relation empty, no provider binding) and derived, not pinned; `db_fresh`
    refuses stale/source-mismatch (T-1b) and the template refuses a dirty provider binding (T-1c); `psql_maint` on the
    `postgres` maintenance database for every lifecycle operation; `db_shared_catalog` covers role attributes, every
    membership crossing either endpoint, database-qualified settings and db ACL/owner/flags of `<db>` and `<db>_tpl`;
    T-5a attribute, T-5b cross-family membership, T-5c database-qualified setting deltas; active-session refusal,
    runtime/template denial and marker removal preserved across lifetimes (§1.1, §3.1, §3.3, §6).
  - **Finding 4 — Prove the independence claimed (AC1/AC3):** reset points recounted and named per row (**50** in
    `migrations` + 1 in `ops`, not 24); §4.1 classifies all 112 tests (I 70, L 25, G 17), documents the direct selector
    and requires every I/L-class test — each of the 35 identity tests and each constituent of multi-test invocations —
    to execute it from a fresh copy once (95 linear executions, bounded container reuse allowed); shuffles of grouped
    invocations are not accepted as constituent proof; G-class groups keep their group selectors; fixture-change and
    extra-file escalation rule (§3.6, §4, §4.1, §6, §13 AC1/AC3).
  - **Finding 5 — Reconcile sequencing and future baseline changes:** owner order G063 → G064 → RC7 → G020 → preview →
    RC8 (if needed) → G057 → G021 → RC9 applied to §0, §8, §13 AC7 and §14 (G063 ACCEPT before G064 integration; RC7
    before G020 resumes; G020 S1/B5 attempt 1 preserved; G063 frozen run untouched); `00020_labels.sql` read read-only
    and its role/ACL/audit effects reconciled (no role/membership/setting change; version 20 / 88 relations derived by
    the template, no new G064 pin; tenancy assertion uses the embedded top version); the six G020 resumption items
    recorded without editing G020 (§3.4, §12); Goal/Spec sha256 and blob SHAs rebound after the staging merge (§1.3).
  - **D-1..D-6:** recorded as approved with their evidence corrections (§10); the r0 D-1 alternative (b) and the D-2
    fallback are withdrawn (§3.5, §10, §11).
  - **Scope/bindings:** `infra/ci/validate-controls.sh` added to the allow-list and bound set; no other scope change;
    budgets unchanged (zero spent).
- r2, 2026-10-21 — one bounded revision answering the r1 Codex review (not included)
  (CORRECTION LOOP: three bounded corrections and one mechanical §8 fix; the r1 design corrections and D-1..D-6 are
  settled and unchanged). Branch `g064-plan` (r1 `bf70e268`) merged with `origin/staging` `79e457dd`
  (management-only delta: PROGRESS.md and the r1 review record; code bindings unchanged); worktree
  `agent-<id>`; Fable 5.1 planner; read-only inspection only (git/grep/sed/cat; no module cache or
  runtime probe); zero of B5 spent. No architecture escalation, no replan; prior r0/r1 text retained where unchanged.
  Changes per r1 item:
  - **Item 1 — Correct the baseline predicate (§3.1):** the `MIG_TOP`-row-count baseline is withdrawn. The template
    baseline is goose's own history: the bound migrate image's `status` listing (one `applied` line per bound file,
    `current_version` = `MIG_TOP`) and, in SQL, the latest row per `version_id` of `app.goose_db_version` with
    `is_applied` true = `{0} ∪ bound` (the version 0 row goose inserts at table creation, `metadata.go:17-25`, is
    required; superseded down/reapply rows never count; the raw history is recorded, never counted or pinned). The
    empty-fixture set is defined precisely as every `app`/`audit` relation with `relkind IN ('r','p','m')` except
    `app.goose_db_version`; sequences, indexes, views, TOAST and composite relations are excluded. T-1 restated,
    T-1a history/table-set negatives added; provider-binding (T-1c) and other residue refusals kept (§1.1, §3.1,
    §6, §13 AC2).
  - **Item 2 — Make prerequisites usable in each approved lifecycle (§3.1/§3.5/§6):** new `db_pristine_verify DEST
    TREE STATUS_LOG`, the source-bound check every template caller runs itself (goose history against the bound
    migration files, `catalog_p`/`catalog_roles` reuse with the existing relation pin hoisted into `MIG_RELATIONS`,
    recorded catalog fingerprint, empty baseline tables, ACL/settings, sessions). `db_template_create` requires only
    that check; `B-MIGRATE-CLI`/`B-MIGRATE-CATALOG` PASS rows are reused **only in the `migrations` lifecycle** and
    are never required of the standalone identity overlay (§3.5 step 2: its own `migrate status` on the overlay
    project) or the private early-risk run (batch 2a) — no added migrations/ops/full-suite run (§2.1, §3.1, §3.5,
    §6, §7, §12, §13 AC2).
  - **Item 3 — Qualify expected results by package (§5):** `go_it` derives expectations with `go test -json -list`
    keyed by `(package, test)` over every package the selector matches; the converter requires exactly one terminal
    event for **every** selected package (`skip` accepted only for a `[no test files]` package) and keys omitted/
    undeclared/duplicate/contradictory rules by package. `identity_it`'s three packages are inventoried (§4 #53:
    `identity` 35 database-integration + 39 untagged, `workos` 21 untagged, `identitytest` none; derived, not
    pinned); its PASS rule adds per-package expected = observed beside the unchanged `-ge 10`/`^ok` controls. **35
    stays the database-integration inventory**; reset points, the 95-selector campaign and the 112 total are
    unchanged — the unit tests are accountable, not campaign members (§1.2, §1.3, §2.5, §4, §4.1, §5.1, §5.2, §6
    converter fixtures, §13 AC1/AC5).
  - **Mechanical §8 fix:** the direct-selector campaign moves from S2 to a new **S4, after the S3 reporter**, run
    mechanically once `go_it`/`gotestreport` exist; batch 2 is split into 2a (S1 lifecycle controls) and 2b (S4
    campaign + AC3 pairs/shuffle/repeat + T-6..T-8); S1's real-PostgreSQL lifecycle proof stays first; later steps
    renumbered S5-S8 (§4.1, §6, §8, §14).
  - **Scope/bindings:** no allow-list change; `metadata.go`, `core/go.mod` and the untagged identity/workos test
    files recorded as read-only facts (§1.3); budgets unchanged (zero spent).

### Correction-to-section mapping (r2)

| r1 review item | Where resolved in r2 |
|---|---|
| 1. Correct the baseline predicate (§3.1) | §1.1 (goose version 0 row, `Status` output, row-bearing `relkind`), §3.1 (`db_pristine_verify` (a)/(d), `db_template_create` record, `db_fresh` postconditions), §6 (T-1, T-1a), §10 r2 risks, §12 (1), §13 AC2 |
| 2. Make prerequisites usable in each approved lifecycle (§3.1/§3.5/§6) | §2.1, §3.1 (`db_pristine_verify`, `db_template_create` preconditions, `MIG_RELATIONS`), §3.5 step (2), §6 (pristine-check row; batches 2a and 4), §7 (`lib.sh`, `run.sh`), §13 AC2 |
| 3. Qualify expected results by package (§5) | §1.2 (overlay row), §1.3, §2.5, §4 (#53, inventory paragraph), §4.1 (campaign scope), §5.1 (`-json -list`, selected package set), §5.2 (per-package rules, `identity_it` rule, `package` rows), §6 (T-6, converter fixtures), §13 AC1/AC5 |
| Mechanical §8: campaign after its reporter dependencies | §4.1 (batch 2b), §6 (batch 2a/2b, S7 reference), §8 (S2 no campaign; S4 campaign; S5-S8), §13 AC3, §14 (S0-S8) |
| Settled, unchanged | r1 Findings 1-5 dispositions, D-1..D-6 (§10), scope allow-list (§9), B5 zero spent (§0) |

### Correction-to-section mapping (r1)

| r0 review item | Where resolved in r1 |
|---|---|
| 1. Complete test selection and termination accounting (R4/AC5) | §2.5, §5.1, §5.2, §6 (T-6, converter fixtures), §13 AC5 |
| 2. Retained evidence and interruption handling (AC5/AC6) | §1.2 (evidence row), §2.7, §3.1 (`db_lifecycle_export`), §5.3, §5.4, §6 (T-8, GT1/GT2), §7, §10 risks (T-7/T-8), §13 AC5/AC6 |
| 3. Verified template baseline and complete shared-state detector (R2/AC2) | §1.1 (`rls_prov_catalog` gaps, `psql_super`, no seed rows), §2.2, §3.1, §3.3, §6 (T-1, T-1b, T-1c, T-5a-c), §13 AC2 |
| 4. Prove the independence claimed (AC1/AC3) | §3.6 (counts), §4 (reset points per row), §4.1, §6 (campaign row), §10 stops, §13 AC1/AC3 |
| 5. Reconcile sequencing and future baseline changes | §0, §3.4 (derived top version), §8, §12, §13 AC7, §14 |
| D-1 | §2.1, §3.1-3.3, §10, §11 (alternative (b) withdrawn) |
| D-2 | §2.4, §3.5, §10, §11 (fallback withdrawn) |
| D-3 | §2.6, §6, §7, §10 |
| D-4 | §6 batch 3, §10 |
| D-5 | §5.1-5.5, §10 |
| D-6 | §5.1, §12, §10 |

## Plan addendum A (2026-10-22): batch-level Docker reuse

Append-only delta to Plan r2; nothing above is edited. Inputs: owner direction relayed by Codex (handoff `de1a7f0f`);
[G064 Goal](../goals/G064-integration-test-isolation.md) (sha256 `202f1a3054d434565ce69fca19903fc15df45949a739bae6bdb72bfa305557dc`,
"Immediate owner-authorized extension"); [Spec](../specs/G064-integration-test-isolation-spec.md) (sha256
`76a9886d1090903bc4fb402585b877d723938709d7eb5d478cc77faebea831e7`, "Owner extension"); the
[batch-resource reuse Spec addendum](../specs/G064-batch-resource-reuse-spec-addendum-20261022.md) (sha256
`0b862409919998dd115dae7389c200b7019a9afe46b31c023bd36315c40f224e`; requirements **BR1-BR4**, criteria **BR-AC1-BR-AC5**);
continuous cleanup mandate (not included). Base `origin/staging` `9c557e58`; every
non-management path is byte-identical to the second full-run candidate `79085025` (`git diff --stat 79085025 9c557e58 --
infra core web` is empty), so file:line references below hold at both. Planner: Opus 5.5 (effort not observable from
inside the agent), isolated worktree, read-only inspection of source and retained evidence; no Docker, build, test or
runtime command was run. **Planning only; implementation awaits bounded supervisor acceptance and release.** Budget:
inside the current B5 attempt (attempt 1 continues per the Result); no reset, extension or new attempt is forecast.

### A1. Problem and measured baseline

**Sources actually read (retained, read-only):** `platform-evidence/g064-a1/qual-79085025/ci-evidence/candidate/`
(`timings.tsv`, `ci-controls-timings.tsv`, `go-tests.tsv`, `db-lifecycle.tsv`, `builds.tsv`, `inventory-leftover.txt`,
`storage.tsv`), the same files of `qual-6ea4cc31`, the console summaries, the S6 timing record of the S4 exec-mode
campaign (Result "S6 timing disposition") and the owner cleanup inventory
(`.local/evidence/owner-cleanup-20261022/containers-{before,removed}.json`). **Not retained, therefore not used:** the
run's owned registry (`ci-evidence/owned/`; `validate.sh` deletes it after a verified cleanup, console line "cleanup:
verified (0 resources removed)") and any `raw/harness-console.log`. The registry row counts quoted in the dispatch
(containers 427, networks 527, volumes 198, images 452) could not be re-read; they are also **capture events, not
unique resources**: `owned_capture` re-appends every container of a project, with its networks, volumes and image,
after each `up|run|start|create|build|restart` call (`lib.sh:264-277, 330-351`), and `--rm` Compose one-offs that end
before the next capture, `yq_stdin` (`lib.sh:529`) and `vtools`/`pyconf` (`run.sh:1700-1702`) are never recorded.

**Wall clock (run 2, `79085025`, `timings.tsv`; run 1 `6ea4cc31` in brackets):**

| Suite | Minutes | Mode | Docker shape (source) |
|---|---:|---|---|
| lint / lint-base / lint-web / lint-core | 4.2 / 1.2 / 1.4 / 0.2 | fg | one-off tool containers |
| provenance / contracts | 0.4 / 1.7 | fg | one-off `vtools`/`pyconf`, 4 builds |
| root | 13.1 [13.1] | fg | root project full stack (`run.sh:1835`), browser, probes |
| migrations | 27.4 [26.9] | fg | `g001-mig-*`: postgres + minio, 64 Go invocations, 5 PROV instances, recovery target project |
| ops | 21.7 [21.8] | fg | `g001-ops-*` full stack, then `g001-idn-*` identity overlay + browser |
| rename / off | 2.2 / 10.8 [2.1 / 9.9] | fg | one full stack each |
| controls | 2.4 | bg | stack-free |
| **ci-controls** | **103.7 [82.0]** | **bg** | 526 controls run serially (sum 103.6 min); nested `validate.sh`, busybox fixtures |
| total | 111.0 (`PREFLIGHT`..`final`) | | foreground provenance..off = **77.3 [76.9]** min |

**Finding 1 (measured):** `ci-controls` is the critical path in both runs: it starts with `provenance` and ends
26.4 min [5.1 min] after `off`. Reusing PostgreSQL/MinIO/Redis or runners in the foreground suites therefore does
**not** shorten a run in which `ci-controls` is selected (it is selected whenever CI/harness dependencies change,
`run.sh:83-88`, which includes G064's own consolidated qualification). Foreground savings shorten runs without it.

**Migrations Go window (run 2):** `B-IDENTITY-CATALOG`..`B-SIGNUP-IT` = 895 s for 59 invocations (+5 drill invocations
inside `B-RECOVERY-DRILL`/`B-MAINTENANCE-DRILL`, 652 s). Package elapsed (go-tests `package` rows) 143 s; 53 `db_fresh`
rows of 1.9-2.4 s (`db-lifecycle.tsv`). Single-test catalog checks (package elapsed ≤ 1.2 s) take 9-10 s each, i.e.
**≈ 6-7 s of per-invocation fixed cost** beyond `db_fresh` and the test: one `docker run --rm` of the core-check image
per invocation (`run.sh:3273-3285`) compiling the test binary for `-list` and `-run` with a cold build cache, plus one
`gotestreport` container (`go_it_convert`, `run.sh:3313-3321`). The S4 campaign at `17fdad66` already used the
reviewed **exec mode** (one long-lived core-check container per package, `docker exec` per selector, `/tmp` emptied
before each; driver `management/v0/results/G064-s4-private-run-driver.sh:1-40`): 95 invocations, invocation time
357,678 ms against 166,640 ms of test elapsed, **≈ 2.0 s per-invocation overhead** (Result, S6 timing disposition).

**Ops:** `B-IDENTITY-IT` 151 s with 45 s package elapsed in **one** invocation; `B-IDENTITY-E2E` 395 s (browser);
`B-OBS-CONTEXT` 490 s (stack build/start/observability); no repeated Go invocations to batch.

**Container accounting available today (three distinct numbers):**

| Figure | Run 2 | Basis |
|---|---|---|
| Cumulative creations, Go path of `migrations` | 128 (64 runner + 64 converter containers) | measured invocation count (64, `go-tests.tsv`) × 2 containers per `go_it` (source) |
| Cumulative creations, everything else | **unknown** | registry not retained; helpers never recorded (above) |
| Peak live project-owned containers | **≥ ~135-142 (lower bound, derived)** | the 127 V18m holder containers (below) were created at ≈45.5-50.4 min into `ci-controls` (reconstructed by summing `ci-controls-timings.tsv` in order) and survived to the owner cleanup, concurrent with the `ops`/identity stacks (source-derived 8 and ≈15 containers). Not derivable exactly: registry rows carry no timestamps and inventories are taken only at start/end |
| Terminal leftovers | **127 containers, 0 networks, 0 volumes** [126/0/0] | `inventory-leftover.txt`; every one is `busybox:1.37.0` with `g003.control=*-nested-hold`, `g003.outer=*-V18m`, state `created`, matched by name to the owner cleanup inventory and removed there |

**Finding 2 (measured):** the leftovers are the V18m nested holder's containers; they carry no `g001.run` label and no
registry row, so `cleanup-owned.sh` verified "0 removed", and `validate.sh`'s existing leftover rule
(`validate.sh:1495-1498`, `complete` 14) was masked by the earlier `B-CI-CONTROLS` failure in the summary line.
Correction 11 owns that repair (A4).

**Finding 3 (source):** PostgreSQL/MinIO are **already batch-shared** across all six Go integration packages: the
`migrations` suite runs every one of its 59 ordinary invocations against one `g001-mig-*` PostgreSQL and one private
MinIO, with a fresh canonical database per invocation from the write-closed template (Plan r2 §2.1; `run.sh:2870-2877,
2926`). Redis is not started there and no Go integration test reads a Redis variable (grep of `Getenv` in the six
packages' `*_test.go`: only database, persist/canary/log, recovery, object sign-host and identity-issuer variables). The
remaining per-test repetition is the **runner container and its cold compile**, not the data services.

### A2. Reuse boundaries (BR1, BR2)

**Package/batch runner decision.** A single `go test` over several database packages is **not safe**: package
binaries would share one database lifetime with no `db_fresh` between them (and run concurrently under default `-p`),
violating BR2/R2. The safe reuse is **one long-lived runner container per batch with one `docker exec` per
invocation**, `db_fresh` before each, exactly the S4 exec mode promoted into `go_it`. A multi-package single
invocation stays limited to what already exists (`identity_it`'s three packages, whose 60 untagged unit tests are
database-free).

| ID | Batch / scenario | Members (§4 rows) | Class | Shared resources | Fresh state / reset boundary | Reason if scenario-owned |
|---|---|---|---|---|---|---|
| B-MIG | `migrations` ordinary Go invocations | #1-51 except the mount-bearing 10 and obj-sign-host: **48 runner-eligible** invocations incl. 3 `TestRLSProvisioned` (env-file only, same data network) | **BATCH-SHARED** | `g001-mig-*` postgres + minio (existing), **one runner** `g064-itr-<run>` (core-check image of the template record), `<project>_data` network | `db_fresh` per invocation (unchanged); runner `/tmp` wiped and verified empty, process table = PID 1 only and `docker diff` confined to the Go build cache, before and after every exec; exec env per invocation only | — |
| B-MIG-c | mount-bearing invocations: tenancy/business-profile/policy persist seed+verify, tenancy-model, business-profile-model, policy-model, policy-ordering (`run.sh:3003, 3007, 3024, 3063, 3067, 3084, 3124, 3128, 3159, 3185`); obj-sign-host (`--tmpfs`, own namespace, `run.sh:3712`) | 11 | BATCH-SHARED data services, **per-invocation container** (unchanged) | postgres + minio | as today | a runner cannot add per-invocation bind mounts; 8 of the 10 use `:ro` mounts whose EROFS semantics an exec cannot reproduce (A8 Q2) |
| S-PERSIST | the six persist pairs (seed → `compose stop/up postgres` → verify) | inside B-MIG / B-MIG-c | scenario **inside** the batch stack | batch postgres (restarted) | fresh at seed only (Plan §3.4) | restart with volume kept is the assertion; the runner holds no connection between execs and survives the restart |
| S-PROV | PROV-1..6 standalone no-init PostgreSQL ×5 (`run.sh:3432-3451, 3504-3509`) | | **SCENARIO-OWNED** | — | own clusters | role provisioning on a bare cluster is the assertion; roles are cluster-global and all five use the same `prov_db` family |
| S-REC | recovery/maintenance drills: source = the `g001-mig-*` project reset (`compose down -v`, `run.sh:4728`), target `g013-rec-*` (postgres, minio, redis), `rd-stage` writers, 5 `rd_go` invocations on two networks (`run.sh:4163-4170`) | #52 | **SCENARIO-OWNED** | reuses the migrations project as its source (existing) | drill chain | backup/restore/recovery is the assertion; `rd_go` switches networks per phase |
| S-ROOT | root project full stack, browser, probes | root suite | **SCENARIO-OWNED** | — | — | documented root `platform:start`, default ports and network boundaries are the assertion |
| S-OPS | `g001-ops-*` full stack | `variants_ops` | **SCENARIO-OWNED** | — | — | outages, persistence restart and credential rotation mutate the stack (`run.sh:5170-5185`) |
| S-IDN | `g001-idn-*` identity overlay + browser; `identity_it` (1 invocation) and `mcp-e2e` (`container:$web`, `run.sh:5763`) | #53-54 | **SCENARIO-OWNED** stack; Go path unchanged | — | D-2 window (Plan §3.5) | documented `identity:start`/overlay configuration (emulator, replica, proxy) is the assertion; one Go invocation, nothing to batch (A8 Q3) |
| S-REN / S-OFF | renamed / intelligence-disabled full stacks | rename, off | **SCENARIO-OWNED** | — | — | startup configuration is the assertion |
| S-CI | `ci-controls` nested launches and fixtures | 526 controls | **SCENARIO-OWNED** (this slice) | — | per control | each control asserts `validate.sh`'s own inventory/registry/cancellation/cleanup of the Docker resources its launch creates (groups D-E, VB, VR, VS, V18*); a scanner or fixture container shared across launches would appear as a foreign/pre-existing resource to each launch's inventory diff and change the subject under test. Not Postgres/MinIO/Redis. Its 103.7 min critical path is a separate decision (A8 Q1) |
| helpers | `yq_stdin`, `vtools`, `pyconf`, documented `compose run --rm migrate` steps, the `gotestreport` converter | all suites | unchanged, **counted** (A5) | — | — | documented CLI commands are B-MIGRATE-CLI's assertion; the converter's no-network, no-mount sandbox is reviewed (§5.2) and its distroless image cannot host a long-lived process |

**Ports, network, credentials, volumes (B-MIG):** the runner publishes no port, mounts nothing (no bind mount, no
volume; any anonymous volume found on it at creation is a refusal), attaches only to `<project>_data` (the network
today's per-invocation containers use), carries labels `g001.run=$RUN_ID` and `g064.batch=mig-$RUN_ID`, and is created
with **no environment**. Each exec receives exactly the env-file and `-e` arguments today's `docker run` receives
(`docker exec --env-file`, supported by CLI 27.5.1, Result S4 validation table); exec environment is not persisted in
the container configuration (today's `docker run --env-file` values are inspectable while the container exists).
Image: the suite's core-check image; the runner's `.Image` must equal the template record's `core_check_image`.

**Per-test outcomes and direct selectors (BR3):** unchanged reporter path: `go_it_begin` → GO_IT_SH (`-list` then
`-run`, exit status captured first) → `go_it_convert`/`gotestreport` → immediate export. Only the transport changes
(`docker exec` instead of `docker run`). The prov gains `runner=<12-hex id>|container` and, for the container path,
`runner_reason=<rule>` (`gotestreport` accepts arbitrary `--prov k=v`, `main.go:28-35`). A direct selector runs alone
from clean prerequisites through the same `go_it` with `GO_IT_RUNNER=off` (container path) or in a one-invocation
runner (driver), never requiring the batch.

### A3. Design deltas (exact files; no new framework, registry, daemon, scheduler or required-check ID)

| File | Change |
|---|---|
| `infra/validation/lib.sh` | new `it_runner_start BATCH PROJECT IMAGE NET` (`docker_run_owned it-runner -d --pull never --label g001.run=… --label g064.batch=… --network NET --name g064-itr-$RUN_ID-<n> --entrypoint sh IMAGE -c '<PID-1 sleep loop that exits on TERM>'`; records ID and `State.StartedAt` in `$WORK/it-runner/BATCH`; inspects mounts; refuses when `BATCH.stopped` exists), `it_runner_check` (running, same ID and StartedAt; `ps` shows PID 1 only; `/tmp` wiped then verified empty; host-side `docker diff` lists changes only under the Go build cache, `go env GOCACHE` recorded at start, and `/tmp`), `it_runner_stop BATCH` (`docker rm -f <exact ID>`, verify "No such object", 6 s late-recreation query by `g064.batch` label as V18m does, write `BATCH.stopped`), `it_runner_stop_all` (EXIT path), `res_snapshot TAG` (A5) and one `res_start HELPER PURPOSE` line appended directly to `$HARNESS_LOG` from `docker_run_owned`, `yq_stdin` and `compose_call` (`run|up|create`) so captured callers and redirections cannot swallow it |
| `infra/validation/run.sh` | `go_it` routing: exec path when `GO_IT_RUNNER_BATCH` is set, the network equals the runner's and every docker argument is `-e K=V` or `--env-file F`; otherwise the container path with a recorded reason. Exec path: `it_runner_check` → `docker exec --env-file … RUNNER sh -c "$GO_IT_SH" …` → rc captured → `it_runner_check` again → `go_it_convert` with that rc. A failed post-check (runner gone/replaced, lingering process, `/tmp` residue) makes the invocation FAIL with finding `runner-fault`/`runner-residue` whatever the rc, removes the runner and lets the next eligible invocation start a new one; a **second** fault in one batch holds the batch (remaining eligible invocations FAIL "batch held: repeated runner fault"; no silent fallback). `suite_migrations`: `it_runner_start` after `db_template_create` (`run.sh:2873`); `it_runner_stop mig` before `recovery_drill` (`run.sh:2957`, whose `compose down -v` must not meet an attached container), in the MinIO-failure branch (`:2958-2967`) and before the final `compose down -v` (`:2971`). `vtools`/`pyconf` (`:1700-1702`) gain the `res_start` line. EXIT trap (`:109`): `bg_stop_all; it_runner_stop_all; go_tests_finalize …` (rebased onto correction 11's trap). `B-IT-ISOLATION` gains **T-9** (runner residue: exec A writes a `/tmp` marker and a `$HOME` marker and leaves a background `sleep`; the post-check must report the lingering process and the `docker diff` path outside the build cache and replace the runner; exec B on the new runner sees neither marker) and **T-10** (runner fault: a throwaway runner is removed during a `sleep` exec; `go_it`'s exec path must return non-zero with `runner-fault` and verified teardown). Same check ID |
| `docs/foundation/OPERATIONS.md`, `docs/PERFORMANCE.md` | runbook lines for the runner, `G064-RES`/`res_start` accounting and the ceilings; PG-010 observation (no benchmark) |
| `management/v0/results/G064-br-proof-driver.sh` (new, task evidence) | the focused proof driver (A7), reusing the S4 driver's P0-P5 instance phases and exec helpers by sourcing it in `S4_DRIVER_SOURCE_ONLY=1` mode |
| not changed | `cleanup-owned.sh` (an unknown row kind is a refusal, `cleanup-owned.sh:141-159`; the runner is an ordinary `standalone` row through its cidfile), `validate.sh` (unless A8 Q6), `required-checks.tsv`, `gotestreport`, any test file, production code, Compose files |

**Allow-list:** no addition is required; every changed path matches the existing
`G064-integration-test-isolation-allowlist.txt` (lines 1 and 10-13). If implementation finds a needed change to
`infra/ci/cleanup-owned.sh` or a Compose file, stop and propose `^infra/ci/cleanup-owned[.]sh$` (or the exact file)
with its reason before editing.

**Test mapping (reuse / extend / new):**

| Item | Kind | Uncovered behaviour it proves |
|---|---|---|
| 48 B-MIG invocations' existing PASS parsers and §4 checks | reuse | unchanged assertions over the new transport |
| T-2 marker isolation, T-3 active session, T-5 shared catalog, T-6 reporter fail-closed (B-IT-ISOLATION) | reuse | database isolation is transport-independent |
| T-9 runner residue (new subcontrol) | new | files/processes left by one exec cannot reach the next; residue is detected and the runner replaced |
| T-10 runner fault (new subcontrol) | new | a persistent runner cannot hide a failed or killed child; exec-path status is non-zero and attributed |
| GT1/GT2, VB5, V18l-n, VR1-VR8 (`ci-controls`) | reuse | adapter fail-closed, background join, nested cancellation, interrupted-run recovery (correction 11 versions) |
| S4 direct-selector campaign evidence (`17fdad66`) | reuse | 95 direct selectors from clean prerequisites (the exec mode already proven there) |

### A4. Resource accounting and cleanup on all paths (BR4)

| Path | Runner (new) | Existing resources |
|---|---|---|
| allocation | cidfile written by the Docker CLI at creation (`docker_run_owned`, `lib.sh:383-393`) → `standalone` row; ID + StartedAt in `$WORK/it-runner/mig`; mounts inspected | unchanged (`register_project`, `owned_capture`, cidfiles) |
| success | `it_runner_stop` at batch end, absence verified, late-recreation query, `G064-RES` row | `cleanup()` (`lib.sh:2620-2704`) |
| test failure | runner kept (process exited, post-check clean) | unchanged |
| setup failure / suite early return | `it_runner_stop` in every `suite_migrations` exit branch; EXIT trap backstop | unchanged |
| runner fault / residue | removed at once, verified, replacement counted; second fault holds the batch | — |
| deadline (validate.sh 14,400 s) / TERM / INT / HUP | EXIT trap: `bg_stop_all`, then `it_runner_stop_all` **before** `go_tests_finalize` (an exec'd `go test` survives its killed `docker exec` client until the container goes); the interrupted invocation is finalized `interrupted` | correction 11: nested HUP/TERM default disposition, `bg_join` ownership |
| power loss / restart | no trap; the runner is an ordinary registry `standalone` row, so `validate.sh`'s existing recovery phase (`cleanup-owned.sh --mode finalize` on the stale registry, `validate.sh:958-1000`) removes and verifies it before any new run; a direct `run.sh` without a registry leaves a `g001.run`/`g064.batch`-labelled container for the mandatory startup reconcile | unchanged |
| terminal | absent from `inventory-leftover.txt`; any project-owned leftover keeps `validate.sh`'s `complete` 14 | unchanged |

**Dependency on correction 11 (not duplicated):** `bbff838c` (`bg_join` child ownership under PID reuse), `82b3cc59`
(nested runner HUP/TERM default disposition, `SigCgt` precondition, V18m `g003.outer` leftover sweep; WIP on
`g064-corr11b`/`g064-corr11c`) and the remaining cleanup-coverage repair named by the cleanup mandate. This addendum
assumes those land first (or as separately attributable commits ahead of it in the same MR), adds nothing to
`bg_join`, `nested_cancel` or the V18m sweep, and only appends `it_runner_stop_all` to the EXIT order that correction
11 leaves. Terminal-leftover proof for V18m (127 → 0) is correction 11's evidence, cited, not re-implemented.

### A5. Concurrency, capacity and live accounting (BR4, BR-AC5)

**Explicit maximum simultaneous project-owned resources.** Source-derived ceilings from the Compose definitions
(`local.yml:221-228`: networks `edge, data, inference`, volumes `postgres-data, minio-data`; `identity.yml:155-156`:
`identity-run`) and the suites' call sites; each is **confirmed at S-B0** against `docker compose config --services`
of each project and recorded. B-MIG excess fails `B-IT-ISOLATION`; a scenario excess is a Result finding (report-only
in this slice, A8 Q6).

| Batch / scenario | Containers | Networks | Volumes | Composition |
|---|---:|---:|---:|---|
| **B-MIG** outside PROV | **5** | **1** | **2** | postgres, minio, runner, ≤ 1 serial transient (converter, per-invocation container, migrate/yq one-off), obj-sign-host only during its invocation |
| **B-MIG** PROV window | **9** | **1** | **7** | + ≤ 5 standalone PROV PostgreSQL and their anonymous volumes; runner idle |
| S-REC | 10 | 2 | 6 | source postgres + minio; target postgres, minio, redis; ≤ 4 `rd-stage` writers; 1 transient; volumes 2 + 2 named, redis anonymous, `g013-old-*` |
| S-ROOT | 11 | 3 | 3 | 6 services, browser, check/probe one-offs (≤ 2 detached error-output probes) |
| S-OPS | 8 | 3 | 3 | 6 services, 1 transient, 1 probe |
| S-IDN | 15 | 3 | 4 | 9 long-running services, exited `identity-setup`/`minio-init`, browser, 1 Go/transient, ≤ 2 detached config probes; volumes 3 named + redis anonymous |
| S-REN / S-OFF | 8 / 7 | 3 / 2 | 3 / 3 | full stack ± intelligence, browser, transient |
| S-CI (background) | ≤ 5 per ordinary control; V18l-n holder burst counted by its own `n0` | ≤ 2 | ≤ 2 | correction 11 requires 0 after each control's sweep |

**Run-wide:** foreground suites are sequential, so the run-wide maximum is the largest foreground scenario plus the
background control: **≤ 20 containers, ≤ 5 networks, ≤ 9 volumes** outside a V18l-n holder burst (previously ≥ ~135
live after V18m, A1). Two-run private pool: the focused proof (A7 F1) takes one slot; a full or single-suite harness run
holds the shared fixed ports (`run.sh:161-164`) and lock and never overlaps another. One build at a time: F1 builds its
five images sequentially in one build slot; the runner adds no build (it reuses the core-check image already built
for the template).

**Memory (WSL capped near 12 GB with 4 GB swap per the coordinator; not re-measured here):** the runner idles in a
sleep loop; compilation happens in it instead of in a fresh container, so peak memory is expected unchanged
(**estimate**) and its writable layer holds the warm Go build cache until batch end (disk, removed with it).
Measured method: one `docker stats --no-stream` of owned containers plus `free -m` at B-MIG start/end and at the S-IDN
window (one-shot calls, no polling).

**Live accounting method (existing artifacts only):**

- **Cumulative creations:** exact for cidfile-recorded containers (count of `cid/<run>-<purpose>-*` files per batch
  window); the `res_start` lines (A3) for `yq_stdin`/`vtools`/`pyconf`/Compose `run|up|create`; unique Compose
  container IDs from `owned_capture`. Any remaining direct `docker run` site reachable inside a batch window is
  enumerated at S-B0 and either routed through an instrumented helper or listed as uninstrumented ("unknown") in the
  Result.
- **Peak live:** `res_snapshot` writes one `G064-RES` line to `harness.log` (retained, scanned) with owned container
  counts by state (label `g001.run=$RUN_ID` ∪ registered Compose projects) at every Go invocation boundary, and
  networks/volumes at batch/window boundaries (the only points where they change by source). Observed peak = max at
  boundaries, plus at most the one serial transient the source allows between boundaries.
- **Terminal leftovers:** existing `inventory-leftover.txt` and the `cleanup-owned.sh` finalize receipt.
- **Result table** per batch: cumulative creations, peak live, final leftovers, start/stop timestamps, separately for
  containers, networks, volumes and images.

### A6. Acceptance mapping, security and stops

| ID | Evidence in this slice |
|---|---|
| **BR1** / **BR-AC1** | A2 matrix with members and reasons; A3 paths; A5 ceilings; runner reuse shown by one runner ID across the 48 eligible invocations in `go-tests.tsv` prov |
| **BR2** / **BR-AC2** | `db_fresh` per exec (unchanged rows in `db-lifecycle.tsv`); T-9; F1 two-invocation same-runner marker proof in both orders; direct single selectors; S4 evidence reused for the 95 selectors |
| **BR3** / **BR-AC3** | F1 deliberate real failure in a reused runner, individual rows/log, non-zero batch, focused rerun under a distinct name; T-10; setup failure, timeout, interruption, truncated stream (runner killed mid-exec) |
| **BR4** / **BR-AC4** | A4 paths; F1 resource receipts (cidfiles, `G064-RES`, late-recreation query, finalize verified); restart recovery case; correction 11's V18m 0-leftover evidence |
| **BR-AC5** | A5 three-number table from F1 and from the consolidated qualification; no intermediate full run (A7) |
| AC1 | §4 rows unchanged; A2 adds the transport column per row |
| AC2 / AC4 | unchanged assertions; RLS/audit/tenancy, ordering, fault and no-retry semantics untouched (no test file edited) |
| AC3 | direct selectors remain runnable alone (`GO_IT_RUNNER=off` or one-invocation runner) |
| AC5 | exit status and completeness unchanged; runner faults are additional non-green findings |
| AC6 | runner teardown on every path, owned-only removal by exact ID |
| AC7 | one consolidated qualification after combined review |

**Security and tenancy (sharing a runner across packages):** credentials are those every invocation of the suite
already receives (one `it.env`, `OBJ_ENV` for object tests, the PROV env), now delivered per exec and never written in
the runner; database state is recreated per invocation (canonical name, template, shared-catalog residue check
unchanged); runner files and processes are wiped and verified at both edges of every exec (process table, empty `/tmp`, `docker diff` confined to the build cache; T-9); the Go build cache
holds content-addressed compiled packages only and test-result caching stays off (`-count=1`); MinIO sharing is the
accepted pre-existing boundary (in-test cleanup, Plan §11; reset by the drill's `compose down -v`); Redis is not
started in B-MIG and no Go integration test uses it; no new privilege, socket, host mount, port or role. Cross-run and
cross-task reuse are impossible by construction (runner name and labels carry `$RUN_ID`; image by ID).

**Escalation stops:** any production/role/schema change; a required-check ID change; a `cleanup-owned.sh` or Compose
change; converting `:ro` mounts to writable; reuse across runs, tasks or sources; inability to record a runner ID at
creation; correction 11 not landed; a ceiling that cannot be confirmed; native refusal; nonconvergence after two
correction cycles of the same issue.

### A7. Focused validation first, then one consolidated qualification

Each step runs after the standard fetch/rebase/reconcile and `bash infra/ci/docker-desktop-mounts.sh`, and records
exact commands, SHAs, counts and cleanup receipts in the Result.

| Step | Command (shape) | Proves |
|---|---|---|
| F0 stack-free | `bash -n` and the B-SHELL shellcheck command on `run.sh`/`lib.sh`; an out-of-repo logic test with a fake `docker` (as the S4 logic test) driving `go_it`'s routing, exec rc propagation (0, 1, 2, 124, 125-127, 137), post-check faults, second-fault hold and the EXIT order | routing/attribution without Docker |
| F1 private real PG (one pool slot, no published port) | `bash management/v0/results/G064-br-proof-driver.sh <checkout> <private evidence dir>` | (a) one runner, ≥ 6 representative invocations across migrate/access/mcp/signup/identity packages with `db_fresh` each; (b) T-2-style marker A→B and B→A in the same runner; (c) T-9, T-10; (d) deliberate real failure: `TestRLSPersistVerify` with `RLS_PERSIST_IDS` naming absent rows on a fresh copy (env-only, runner-eligible) → its own FAIL rows/log, the next invocation in the same runner PASS and separately attributed, batch exit non-zero; then a **focused rerun** `rerun-rls-persist-verify-1` after a real seed → PASS, both records retained; (e) setup failure (wrong runtime password env), `GO_IT_TIMEOUT=5s` on `TestAccessOrdering`, zero-match selector; (f) TERM and, separately, HUP to the driver during an exec → runner removed, invocation `interrupted`, exit 128+n; (g) runner killed mid-exec → truncated stream incomplete + `runner-fault`; (h) restart recovery: `kill -9` of a second driver instance mid-batch (no traps), then the existing `cleanup-owned.sh --registry <reg> --mode finalize` → runner and stack removed and verified, late-recreation query empty; (i) the A5 three-number table and `docker stats`/`free -m` snapshots |
| F2 controls (only if correction 11 changed them) | `bash infra/ci/validate-controls.sh --only GT1,GT2,VB5,V18l,V18m,V18n --log <dir>/ci-controls.log` | adapter, background join and nested cancellation under the combined source |
| Source/Duo/fast gates | normal MR flow | — |
| **Q-final** | one consolidated merged-SHA qualification after correction 11 + this slice are reviewed and integrated | at-scale counts for BR-AC5; supersedes nothing failed |

**No full or single-suite run between the current repairs and this addition**, and no second broad run after a
failure without diagnosis (Spec addendum sequencing). F1 is the only Docker proof planned before the MR; an extra
`migrations`-suite run is not planned (A8 Q5). Budget: inside the current B5 attempt; anything forecast to exceed it is
reported before it starts.

### A8. Open questions for the supervisor and expected savings

| Q | Decision requested | Default in this Plan |
|---|---|---|
| Q1 | `ci-controls` (82-104 min, critical path in both runs) keeps scenario ownership in this slice; scope any reduction (per-launch fixed costs, slow controls V18m 294 s, V18a 167 s, VB6 150 s, V4 131 s) as a separate bounded item? | yes, separate |
| Q2 | Convert the 10 mount-bearing invocations to the runner via `docker cp` exchange (loses EROFS on 8 `:ro` mounts) | no |
| Q3 | Move `identity_it` into B-MIG (would retire the approved D-2 window and change `B-IDENTITY-IT`'s suite) | no |
| Q4 | Leave helper one-offs (converter 64/run, `yq`, `vtools`, `pyconf`) unbatched, counted only | yes |
| Q5 | Accept F1 plus the consolidated qualification as the at-scale proof instead of a separate `migrations`-suite run | yes |
| Q6 | Scenario ceilings report-only in this slice (B-MIG excess fails `B-IT-ISOLATION`); optionally make `validate.sh` always print the leftover count even when an earlier `complete` rule fails | report-only; optional line |

**Expected savings (all estimates unless marked measured):**

| Item | Before | After | Basis |
|---|---|---|---|
| Go-path container creations in `migrations` | 128 (measured count × source) | ≈ 82 (1-2 runners + 16 per-invocation + 64 converters) | 48 eligible × 1 runner container avoided |
| `migrations` wall clock | 27.4 min (measured) | ≈ 24 min | ≈ 4-5 s × 48 (container-mode ≈ 6-7 s measured on catalog checks vs exec-mode ≈ 2.0 s measured in S4) |
| Full-run wall clock with `ci-controls` selected | 111 min (measured) | **≈ unchanged** | `ci-controls` critical path (Finding 1) |
| Full-run wall clock without `ci-controls` | — | foreground −3 to −4 min | as above |
| Peak live project-owned containers | ≥ ~135 (derived lower bound, dominated by the V18m leak) | ≤ 20 outside holder bursts | correction 11 removes the leak; this slice adds +1 runner to B-MIG |
| Terminal leftovers | 127 / 0 / 0 (measured) | 0 / 0 / 0 required | correction 11 + A4 |
| Peak memory | not measured | expected unchanged | A5 method measures it |

The honest summary: the data services are already shared per batch; this slice removes the repeated runner container
and cold compile (the measurable remaining per-test cost), formalizes the scenario list, budgets and three-number
accounting, and depends on correction 11 for the only large observed resource problem (the V18m leak). The run's
wall-clock limiter is `ci-controls`, which this slice deliberately does not change without a supervisor decision.

## Plan addendum A, correction 1 (2026-10-22): Codex review dispositions

Append-only; the original Plan and addendum A (lines 1-1177) are unchanged. Answers the
Codex review (not included) (sha256
`057633422b83fb313ede1d24fe2917d20f0c9edf3e0e911371258cdbd6047511`: findings 1-2, Q1-Q6 and the "Correction 11 item 4"
ruling). Base `origin/staging` `bddfff78`; infra/core/web are byte-identical to `9c557e58`, so A's file:line citations
hold. Correction-11 source cited as `vc@25e878e9` / `lib@25e878e9` (`origin/g064-corr11c`); unchanged code is cited at
staging. Planner: Opus 5.5, isolated worktree, read-only source and evidence; no Docker, build or test was run. Planning
only; no attempt charged or reset. **Figure classes, used everywhere below and in the Result: M = measured observation
(exact or sampled, stated), D = source-derived bound, U = unknown historical value.** A sampled maximum is a lower bound
of the true peak and is never called a peak; the enforced ceilings (D) are the only bounds.

### C1. Exact successful-allocation accounting (finding 1; BR4, BR-AC1/4/5)

A creation is counted only from an identity Docker returned: a cidfile written by the CLI, the ID printed by `create`,
`run -d`, `network create` or `volume create`, an inspect of an existing object, or a witnessed create event. A logged
launch or a non-zero call without an ID is never a creation. Counts deduplicate by full ID per class. The registry
(`owned.tsv` + `cid/`) remains the only ownership record; accounting files are reports derived from it. No new registry
row kind (`cleanup-owned.sh:159` refuses unknown kinds); `cleanup-owned.sh` stays unchanged.

| # | Allocation path (today) | Exact-ID rule (change) |
|---|---|---|
| P1 | `docker run`/`create` in harness code: 34 raw `docker run --rm --label g001.run` sites in `run.sh`, `lib.sh:520,523,529,2698`, `vtools`/`pyconf` (`run.sh:1700-1702`), `docker create` (`run.sh:5817,6093,6149`). Only the 24 `docker_run_owned` call sites write a cidfile (`lib.sh:383-393`) | every site goes through `docker_run_owned` (plus a `create` mode): cidfile at creation, `standalone` row. Detached: mounts inspected at once, anonymous volumes become `volume` rows. Attached `--rm` keeps `--rm` unless the image declares `Config.Volumes` (one cached `image inspect` per image) or the arguments hold an anonymous mount; then `create` → inspect mounts → `start -a [-i]` → `rm -v` → verified gone, with the container's exit status returned. That makes the transient anonymous volumes of `$RD_PGIMG` psql helpers (`run.sh:4246-4444`) exact |
| P2 | Compose long-running services: `owned_capture` after up/run/start/create/build/restart (`lib.sh:264-277, 326-351`) re-appends rows | unchanged; counted once per ID. Project-labelled networks/volumes of a registered project without a capture row are an accounting finding (the same objects `cleanup-owned.sh` already reports `unproven`, `:339-359`) |
| P3 | Compose one-offs: harness `run --rm` calls (`run.sh:534, 2856, 2859, 2872, 4730, 5279, 5284, 5286, 5337, 5878, 6174, 6281, 6350, 6351`) and documented blocks run verbatim (`lib.sh:400-427`); the container is gone before any capture | call-scoped creation witness in `compose_call`/`ledger_run_block` for calls or blocks with a `run` subcommand: `docker events --filter label=com.docker.compose.project=P --filter type=container --filter event=create` is started as a child, then a fence container (cidfile-recorded, `--network none`, that project label) is created and removed. The witness waits, bounded to 10 s, until the fence ID appears; the call runs; a second fence follows, then a bounded wait for it, then the child is stopped and reaped. IDs between the fences, fences excluded, become `container` capture rows: accounting-only, because an absent ID classifies `absent` in `cleanup-owned.sh:216`. Volume `mount` events whose `container` attribute is a witnessed ID record anonymous volumes. A fence timeout, a lost child or a failed event stream is an accounting failure, never "unknown". Documented commands stay verbatim |
| P4 | Images: build rows (`lib.sh:354-378`, ledger) | unchanged. Image IDs present at the end but not at the start and not built are reported `pulled` and retained (shared-cache policy, cleanup receipt) |
| P5 | Wrapper helpers `validate.sh:1202,1232,1510,1718` (no label, no cidfile); `runner-preflight.sh:67` already writes registry cidfiles | add `--cidfile "$REG/cid/wrapper-<n>.cid"`, verify each container is gone after the call (bounded 5 s). An unverified one sets the cleanup status to 18, which keeps the registry and marker (`validate.sh:1829-1831, 896-906`) |
| P6 | CI controls (`g003.controls.run`/`g003.control`/`g003.outer`): correction 11 keeps `$ROOT/owned/<ctl>.tsv` for V18l-n only (`vc@25e878e9:356-445, 667-690`) | every control keeps the inventory, filled at the two existing allocation paths: `D()` (`vc:213`) at control level and the shim's `run\|create` (`vc@25e878e9:485-525`) for nested runs. A cidfile is added when the arguments have none, or the caller's cidfile is read. Also recorded: printed IDs of network/volume create, image IDs after `build`/`commit`, the holder's self-records (deduplicated). Each row is mirrored into the run registry (C4). At control end, an object labelled `g003.control=$RT-$CTL_ID` or `g003.outer=$RT-$CTL_ID` without a row is an `unrecorded` FAIL; correction 11 checks only the outer label |

**Report and reconciliation.** `run.sh` writes `resources.tsv` to the report directory (added to `validate.sh`'s staged
list beside `inventory-leftover.txt`, `validate.sh:129`). It has one row per scope (B-MIG, PROV window, S-REC, S-ROOT,
S-OPS, S-IDN, S-REN, S-OFF, helpers, wrapper, S-CI by control) and class (container/network/volume/image) with these
fields: `creations` (unique IDs, M exact), `observed_max` (M sampled), `ceiling` (D), `final_leftover` (M),
`unaccounted` (M) and `status`. `created` detail rows give kind, full ID, purpose and scope. Scope boundaries are
registry line offsets in `$WORK/res-scopes.tsv`. Concurrent appends from the background CI child are single short
`printf >>` writes (one `write(2)` each, O_APPEND); F0 checks two concurrent writers. `validate-controls.sh` writes its
per-control rows to a file that `run.sh` passes in (`run.sh:1502`) and merges.

**Observation points (sampled, M).** These replace A3/A5's `res_snapshot` (which filtered only `g001.run` and registered
projects) and `res_start` lines (attempts, not creations). Points: the return of every allocation helper, each Go
invocation boundary and each batch/suite boundary. Containers are counted at every point; networks/volumes at
Compose/batch points and after creations with anonymous volumes. Each point also samples all containers carrying any
`g003.controls.run` label. That gives the run-wide simultaneous sample (a lower bound of the true peak). Docker query
failure at a point is an accounting failure, not zero.

**Accounting failures are non-green and attributable.** `RESOURCES FAIL <scope>: <finding>` is a non-required row, and
any FAIL row fails `do_complete` (`validate.sh:1386, 1413` → 14). Findings: rc 0 without a cidfile ID; an
unrecorded/unaccounted labelled object; a witness failure; a query failure; a ceiling breach (C2); unverified teardown
(C3); `resources.tsv` missing when the harness ran with a registry. In CI, the failing control reports the finding and
B-CI-CONTROLS fails. A direct `run.sh` without a registry makes no accounting claim: it logs
`G064-RES accounting=unavailable (no registry)` and is never cited as BR evidence.

**A figures superseded by this correction** (A itself unedited): A1's registry-row discussion stays as history.
Superseded:

- A3: `res_start`/`res_snapshot` and the runner's "PID-1 sleep loop" (C3).
- A3: replacement without verified teardown (C3).
- A4: the power-loss row's "direct `run.sh` without a registry ... startup reconcile" (C4).
- A5: the S-CI row; S-REC volumes 6 (C2).
- A5: "scenario excess report-only"; the run-wide "≤ 20 / ≤ 5 / ≤ 9 outside a V18l-n holder burst", which left the
  leaking class out.
- A5: the cumulative-creation method based on `res_start`; "observed peak = max at boundaries plus one transient"
  (now `observed_max`, M sampled).
- A7: F1(h) and the F2 list (C5).
- A8: Q1-Q6 defaults (C6); the savings row "≤ 20 outside holder bursts".

Retained, with their class:

- Run 2/1 leftovers 127/126 containers (M).
- 64 Go invocations (M).
- "≥ ~135-142 live" (D, lower bound, not a peak).
- Historical cumulative creations and the run-wide peak (U).

### C2. Finite, enforced ceilings at named boundaries (finding 1)

| Scope | Containers / networks / volumes (D) | Enforced at |
|---|---|---|
| B-MIG outside PROV / PROV window | 5/1/2 and 9/1/7 (A5, PROV anonymous volumes included) | every `go_it` boundary and allocation-helper return |
| S-REC | 10/2/**10**: named 2+2, redis anonymous 1, `g013-old` 1, anonymous volumes of ≤ 3 detached `$RD_PGIMG` writers (`run.sh:4224, 4232, 4552`) plus 1 transient psql. A5's 6 omitted the writers' anonymous volumes (postgres images declare `VOLUME /var/lib/postgresql/data`; the pinned image is confirmed at S-B0) | helper returns, drill phase boundaries |
| S-ROOT, S-OPS, S-IDN, S-REN, S-OFF | A5 rows unchanged: 11/3/3, 8/3/3, 15/3/4, 8/3/3, 7/2/3 | Compose and helper returns |
| S-CI ordinary control (label `g003.control=$RT-$CTL_ID`) | **8/2/2, images 2**. D: fixture objects ≤ 2 containers/1 network/1 volume/1 image (`vc@25e878e9:3104-3163`, `mk_image` `:3079-3085`) plus one nested launch ≤ 3 (stub standalone `vc@25e878e9:663/692-700`, preflight httpd `runner-preflight.sh:126`, one attached helper). Total 5/1/1. A margin of +3/+1/+1 covers controls not enumerated one by one here. S-B0 statically enumerates all 526 control functions and stops with a proposal if any exceeds the derived 5 | the shim and `D()` before each persisting creation (create, `run -d`, network/volume create) |
| V18l-n nested holder (label `g003.outer=$RT-$CTL_ID`) | **40/0/0** (images 0). The holder creates one container per loop and sleeps 2 s, so creations ≤ 1 + ⌊L/2⌋ for holder lifetime L. Passing path (D): deadline mode L ≤ 8 s nested launch deadline (`vc@25e878e9:2152`) + 30 s `STOP_GRACE_S` (`vc:236`), so ≤ 20. Signal modes: L ≤ pre-signal steps + 30 s; the pre-signal steps are local calls that the source does not bound. M: n0 = 1 at the observation point in all 6 passing V18l-n observations (qual-6ea4cc31, qual-79085025, s6-e7c4efc6, corr11c proof logs); control walls V18l 56/48 s and V18n 74/73 s bound L, so ≤ 38. Ceiling 40 = 38 + 2 for the check-then-create race of ≤ 2 concurrent creators (holder plus one preflight/helper transient). The failed run 2 had 127 (M), which matches a 240 s undelivered-HUP wait (D: ≈ 254 s / 2) and would now stop at 40 | the nested shim, before each holder `docker create` (ceiling passed down with `CI_CONTROLS_OUTER_OWNED`) |
| Run-wide | **63/5/12** (D) = the largest foreground scope (S-IDN 15/3; volumes S-REC 10) plus the largest background control: V18l-n 40 outer-labelled + 8 control-labelled. Foreground suites are serial; `controls` and the web prebuild create no containers | sum of the independently enforced ceilings, never a sampled figure |

**Enforcement, never report-only.** `run.sh` checks after allocation. A breach records `RESOURCES FAIL <scope> ceiling`
and sets `RES_HOLD=<scope>`. While it is set, the allocation helpers refuse with "held: resource ceiling" and create
nothing: `docker_run_owned`/create, Compose `up|run|create|start|restart` via `compose_call`, the witness and
`it_runner_start`. B-MIG's remaining invocations FAIL `batch held`. Removal and teardown are never blocked. The
scenario's existing teardown runs at once. The next suite starts only after a verified zero foreground count; an
unverified count sets `RES_HOLD=run`, so every later allocation is refused and all later suites fail.

CI controls refuse before creation, so a ceiling is never crossed. The shim/`D()` counts live labelled objects of that
class. At the ceiling it returns 1 ("resource ceiling N/MAX for <ctl>") and appends `ceiling` to the inventory, and the
control FAILs. A query failure is a refusal. `owned_cleanup` (`vc@25e878e9:400-445`) then runs as today. The runner
loop (`vc@25e878e9:6216-6245`) dispatches the next control only after verified clean. Otherwise it holds: each remaining
selected control is reported FAIL `held: cleanup of <ctl> unverified` and the summary is non-zero. Assertions, decoys,
the 240 s bound and the V18m cancellation, reaping and scratch assertions (`vc@25e878e9:2139-2237`) are unchanged. There
is no holder-burst exception. For the proofs only, `RES_CEILING_LOWER=<scope>:<n>` (and `CI_CONTROLS_CEILING_LOWER`)
may lower a ceiling; an override that would raise one is refused.

### C3. Fail-closed teardown, replacement and the runner baseline (finding 2)

**Runner and idle baseline.** The runner is created by `docker_run_owned it-runner -d --init --pull never --network
<project>_data --label g064.batch=mig-$RUN_ID ... --entrypoint sleep IMAGE 2147483647`. That command has no shell; S4's
`sh -c 'while :; do sleep 1; done'` (`G064-s4-private-run-driver.sh:274`) always has a child and is not used. The exact
accepted process set, read host-side with `docker top <id> -eo pid,ppid,stat,args`, is:

1. PID 1 = `/sbin/docker-init -- sleep 2147483647`.
2. Exactly one child of PID 1 = `sleep 2147483647`, not a zombie, whose host PID is recorded at start and must never
   change.

`--init` reaps orphans, so a finished exec leaves no zombie, while a lingering live process stays visible. At start the
runner also records ID, `StartedAt`, image ID (must equal the template record's `core_check_image`) and
`go env GOCACHE GOTELEMETRYDIR`, then refuses anonymous volumes.

**Per-exec ordering.** Before an exec:

1. Running, with the same ID, `StartedAt` and sleep PID.
2. The process set equals the baseline.
3. `docker diff`: every `A`/`D` path lies under an allowed root (GOCACHE, the telemetry directory, `/tmp`), and `C` is
   allowed only on those roots and their ancestors.
4. `/tmp` is empty (`find /tmp -mindepth 1 -print -quit`).

After the exec, in this order:

1. (a) Runner identity. A change is `runner-fault`.
2. (b) Process set. Checked before any reset, so a lingering writer cannot act after a wipe; an extra process is
   `runner-residue: process`.
3. (c) `docker diff` outside the allowed roots (for example a `$HOME` marker) is `runner-residue: file <path>`.
4. (d) Only then the legitimate `/tmp` output is reset: `GO_IT_SH` has already streamed its `/tmp` files into the
   event log (`run.sh:3268-3272`). The reset is `rm -rf` of `/tmp`'s entries, then a verified-empty check.

Residue is refused, with no reset attempted, for (a)-(c), and for a reset failure or a non-empty `/tmp` after reset. The
invocation FAILs with that finding whatever its rc; its rows and log are kept.

**Generation versus terminal state (`$WORK/it-runner/`).** The files are:

- `mig.gen` = N and `mig.id`: the current generation's identity.
- `mig.gen<N>.gone`: written only after verified teardown.
- `mig.faults`: the fault count.
- `mig.stopped`: terminal. Written only by `it_runner_stop` at the named batch ends: before `recovery_drill`
  (`run.sh:2957`), the MinIO-failure branch (`:2958-2967`), before the final `compose down -v` (`:2971`), and on EXIT.
- `mig.held`: terminal, with its reason.

`it_runner_start` allocates generation N+1 only when there is no `stopped` and no `held`, `gen<N>.gone` exists, faults
are fewer than 2, `RES_HOLD` is unset and the durable registration holds (C4). A second fault writes `held`.

**Verified teardown** is used both for a fault and for a terminal stop:

1. `docker rm -f <id>`.
2. `docker inspect --type container <id>` must fail with "No such object". Any other error, Docker unavailable
   included, is unverified.
3. `docker ps -aq --filter label=g064.batch=mig-$RUN_ID` must succeed and be empty, now and again after 6 s (late
   recreation: any hit is removed, recorded and still holds).

Any unverified step writes `mig.held` and `RESOURCES FAIL`. B-MIG's remaining invocations FAIL `batch held: runner
teardown unverified (<step>: <first Docker error line>)`. S-REC is held too, because it resets the same project and
network (`run.sh:4728`). Unaffected suites continue.

Evidence is preserved: the failed invocation's rows and log, `diagnostics/it-runner-mig.txt` through `export_diagnostics`
(`lib.sh:33-50`), and the registry with its `standalone` row. `validate.sh`'s `do_cleanup` → finalize retries (`:833-846`).
A failure there is cleanup 18, which keeps the registry and marker for the next launch's `do_recovery`.

### C4. Durable recovery: one existing mechanism (finding 2; item-4 ruling)

**Precondition for any reusable or outliving allocation** (the runner; mirrored control inventories) is a valid run
registry (`OWNED_REG`, header/job/token/repo, the same checks as `cleanup-owned.sh:115-190`) plus a durable recovery
registration:

- CI: job registry, fresh daemon and the `after_script` backstop (`validate.sh:854`; `cleanup-owned.sh --after-script`).
- Local: a pending marker `$TMPDIR/g003-pending/<job>.tsv` whose `registry` row names this registry, with stage
  `launching|launched` (`validate.sh:871-894`, wired at `:1984-1990`).

Without it, **`it_runner_start` refuses** and `go_it` takes today's per-invocation container path with
`runner_reason=no-registry|no-durable-recovery`. There is no deferred startup reconcile. Direct selectors and tests
still run.

**Correction 11 converges on the same records.** `run.sh` passes `CI_CONTROLS_OWNED_REG="$OWNED_REG"` to
`validate-controls.sh` (`run.sh:1502`). Nested instances receive it with `CI_CONTROLS_OUTER_OWNED` (`vc@25e878e9:2155`).
Every inventory row (P6) is mirrored at creation as an existing row kind: containers `standalone <id> ci-<ctl>`, networks
and volumes as capture rows, images `image <tag> <id>`. A crash therefore leaves registry-owned objects. The next local
launch's `do_recovery` (`validate.sh:954-1020`) removes them through `cleanup-owned.sh` after it has verified that the
creator group and watchdog are dead (`recovery_survivors`, `:926-952`). Otherwise it blocks with 16, the marker is kept
and the harness is not started.

A top-level direct `validate-controls.sh` (no registry passed; not nested; not simulation mode `CI_CONTROLS_HOLD_SIM`)
self-registers first. It uses validate.sh's own `make_registry`/`marker_write`/`marker_stage`/`marker_finish` text,
`eval`-extracted (precedent: `vc:3967` extracts `WATCHDOG_SH`), so there is no second writer implementation. If it is
not its own group leader it re-execs once under `setsid` (as `validate.sh:752` launches the harness), so the recorded
group is exactly its tree. It finalizes on exit; when finalize is unverified the marker stays.

The F1 driver does the same: a registry as `s4_registry_init` (`G064-s4-private-run-driver.sh:98-129`), the extracted
marker writer, and a refusal to start while any marker is pending ("launch validate.sh to recover").

**Item-4 guard.** The `g003.controls.owner` key binds boot id, PID and start ticks, and canonical repository/run
identity (ruling, review file). It is referenced, not redesigned. It is containment only. A guard hold is a
`RESOURCES FAIL`/non-green outcome and never a clean receipt. With mirrored rows, objects whose owner is dead are
recovered by `do_recovery`. Only leaks that predate this change need the recorded one-time operator procedure
(OPERATIONS.md); there is no standing manual cleanup.

### C5. Focused proofs (new delta only; no full suite, no timing campaign)

| Step | Proves (reuse / extend / new) |
|---|---|
| F0 stack-free, extended (fake `docker` as the S4 logic test) | `docker_run_owned` modes: cidfile, `--rm` split, exit status and stdin propagation. Witness fences, including a timeout → FAIL. Deduplication: the same ID captured twice, and again after a restart, counts 1. Ceiling breach → `RESOURCES FAIL` and later allocations refused. Teardown outcomes: rm error, inspect error, Docker down, late hit → `held`. `stopped`/fault/second-fault semantics. Precondition refusals (no registry; no marker; marker naming another registry) → container path. Concurrent registry appends |
| F1 private real PG (one pool slot), extended | Retains A7 F1 (a)-(g) and (i). Adds the accounting fixtures: (1) `yq_stdin` → exactly one recorded ID, then gone; (2) `$RD_PGIMG` psql `--rm` → container plus anonymous volume recorded and gone; (3) `compose_migrate run --rm migrate status` → witnessed ID; (4) a nested-control allocation through the shim (V18n run inside F2) is mirrored into F1's registry; (5) `RES_CEILING_LOWER=B-MIG:3` → the crossing invocation FAILs, the rest are `batch held`, the runner and stack are torn down and verified, and the next scope's zero check passes; (6) T-9/T-10 against the exact baseline. F1(h) is **replaced** by VR9 |
| F2 controls (direct, self-registered) | `--only GT1,GT2,VB5,V18l,V18m,V18n,VR1,VR2,VR6,VR9`. VR1/VR2/VR6 are rerun because `docker_run_owned` (their stub's helper, `vc:591-600`) changed. V18l-n show real holder counts under the 40 ceiling. In V18n's daemon-free cleanup driver (`vc@25e878e9:2243-2333`), cases `cap` (refusal at the lowered ceiling, control FAIL naming it, 0 left) and `held` (refused removal; the next control is not dispatched and is reported held) are added. S-B0 greps the controls that assert `complete:` messages; any of them is added if Q6 touched it |
| VR9 (new control, real `validate.sh` in a case tree; the marker/recovery entry path) | (a) A `runner-hold` stub sources the real `lib.sh`, passes the precondition through the case's real marker, calls `it_runner_start`, starts `docker exec -d <runner> sleep 300`, then holds. `vr_crash`-style KILL of `validate.sh` and its harness group; the runner and its exec'd process survive. `vr_settled` and the next launch's `recovery_survivors` verify that the creator group and watchdog are dead. The next run recovers: runner gone, the `g064.batch` query empty after 6 s, marker removed, delayed-recovery row. (b) The same with a self-registered top-level `validate-controls.sh` creating one recorded container, then KILLed: recovered by the same `do_recovery`. (c) Crash, then a tampered token: exit 16, `expect_runs 1` (no harness, so no new resource batch), runner kept, and the `g001.run` container count unchanged. VR3-VR5/VR7/VR8 evidence is reused unchanged |

### C6. Rulings carried as decisions; scope; acceptance delta

- **Q1:** CI-controls scenario isolation stays, with no CI redesign, new task or speed commitment. Finite CI limits (C2)
  and complete accounting (P6, C4) are in G064 now.
- **Q2:** no `docker cp` conversion; the read-only-mount per-invocation exception (B-MIG-c) stays.
- **Q3:** the D-2 identity overlay/window and `B-IDENTITY-IT` suite attribution stay.
- **Q4:** isolated helper/converter sandboxes stay; their IDs are accounted (P1, P3).
- **Q5:** focused F0/F1/F2/VR9, then one consolidated final qualification. No extra migrations run or full umbrella.
- **Q6:** report-only maxima are rejected (C2). `do_complete` will print the leftover count (`validate.sh:1495-1498`)
  and the `resources.tsv` findings before its early returns (`:1410-1413`), so cleanup and leftover findings appear
  next to any earlier failure. The status rules are unchanged; this is a bounded diagnostic in the allowed path.

**Touched paths:** `infra/validation/{run,lib}.sh`, `infra/ci/validate.sh`, `infra/ci/validate-controls.sh`,
`docs/foundation/OPERATIONS.md`, `docs/PERFORMANCE.md`, `management/v0/results/G064-br-proof-driver.sh`. All of them are
in the existing allow-list. **No additional allow-list pattern is proposed.** Not changed: `cleanup-owned.sh`,
`runner-preflight.sh`, `infra/ci/lib.sh`, Compose files, `required-checks.tsv` (no new required check ID;
`RESOURCES` is a non-required row like `BACKGROUND`), test files, production code. A6's escalation stops stand. If
implementation needs `infra/ci/lib.sh` (for example a shared marker helper instead of `eval` extraction), the pattern
`^infra/ci/lib[.]sh$` will be proposed with the reason before any edit.

**BR3/BR-AC3 unchanged.** Each test/subtest keeps its rows, timing, selector, sanitized log and JUnit/TSV, with
completeness and child-status propagation. A held, faulted or refused invocation is an additional non-green finding,
never a pass. Direct selectors run alone (container path or a one-invocation runner), and focused reruns stay separately
attributable.

**Acceptance delta (amends A6):**

- **BR-AC1:** C1/C2 tables, exact touched paths, enforced ceilings.
- **BR-AC4:** P1-P6 receipts, C3 verified teardown, VR9 recovery and refusal, zero `unaccounted`, holds on cleanup
  failure.
- **BR-AC5:** `resources.tsv` (M creations and sampled maxima, D ceilings and run-wide bound, U history) from F1/F2 and
  Q-final.
- **AC6:** C3/C4.

### C7. Could not verify (planning, no Docker)

- The Docker CLI still writes `--cidfile` for `--rm` runs.
- `docker events` label/type filtering and fence ordering on CLI/daemon 27.5.1.
- `docker top -eo` on Docker Desktop.
- That `sleep 2147483647` and `/sbin/docker-init` behave as stated in the core-check image.
- That the pinned `$RD_PGIMG` declares `VOLUME`.
- Go telemetry writes under `GOTELEMETRYDIR`.
- The single-`write(2)` append atomicity.
- The per-control maximum over all 526 controls (S-B0 static enumeration).

F0/F1/F2/S-B0 confirm each of these. A refutation stops the affected step for a Plan correction; it is never silently
widened. The added cost of observation points, witnesses, ceiling queries and per-helper cidfiles has not been measured.
F1/F2 will report it.

## Plan addendum A, correction 2 (2026-10-22): C1/C2/C4 guarantee fixes

Append-only; lines 1-1435 are unchanged. Answers "Plan addendum A correction 1 — delta review" in the
Codex review (not included) (sha256
`49682f9fdd25547fb668b6a4e7675e72f89b70c111415d8eb139a26415957277`). Base `origin/staging` `391296d9` (management-only
since `bddfff78`; correction 11 cited as `vc@25e878e9`). Planner: Opus 5.5, isolated worktree, read-only; no Docker,
build or test. C1-C7 direction, runner/fresh-state/reporting design, scope and Q1-Q6 stand; only D1-D4 change them.
Figure classes M/D/U as in C1. Event behaviour is quoted from the
[events reference](https://docs.docker.com/reference/cli/docker/system/events/) (fetched, not run).

### D1. Event coverage and positive attribution (replaces C1 P3's two fences per call)

Documented: container events include `create`, `destroy` and `top`. Volume events are `create`, `destroy`, `mount` and
`unmount`; the mount example carries `read/write`, `container`, `destination`, `driver` and `propagation`. A repeated
filter key is OR, different keys AND, and without `--since` only new events arrive. The page does not say volume events
carry labels, so no daemon-side `label=` filter is used (it could drop every volume event).

**One run-scoped witness**, started with a registry by `res_witness_start` in run.sh's main shell before the suite loop
(run.sh:6460); never inside `$(...)`, where a child holding stdout would block the substitution.
1. A supervised child (`bg_start res-witness`, lib.sh:1172; joined after run.sh:6477; EXIT's `bg_stop_all`, run.sh:109)
   runs `docker events --format <TSV of .Type .Action .Actor.ID and the Actor.Attributes com.docker.compose.project,
   ...project.working_dir, container, destination, image> --filter type=container --filter type=volume --filter
   event=create --filter event=destroy --filter event=mount --filter event=top`, which means (container OR volume) AND
   (create OR destroy OR mount OR top), into one bash parser.
2. W is created through D2's latch and lock: `docker_run_owned witness -d --pull never --network none --label
   g001.run=$RUN_ID busybox:1.37.0 sleep 2147483647` (lib.sh:14), giving a cidfile and a `standalone` row. P1's cached
   `image inspect` must show no `Config.Volumes`, else the start is refused.
3. A bounded wait (C1's 10 s) for W's own `create` event proves the stream is live before any witnessed call.

**Attribution (live, in the parser).**
- **Container C:** its own `create` event carries a project in `$WORK/projects` (lib.sh:308) and a working_dir under
  REPO or WORK (owned_capture's rule, lib.sh:337-338), and C is absent from the start inventory (lib.sh:324). A
  `container C` capture row is appended at once.
- **Anonymous volume V:** its own `mount` event has Actor.ID = V's name and a `container` attribute with a full ID. If
  that ID is an attributed C and V is absent from the start inventory, a `volume V` capture row (accepted by
  cleanup-owned.sh:24) is appended at once, detail `mount:<C>:<destination>`. The binding is the mount relation in V's
  own event; no label of V is taken from C. V counts as anonymous when its own `create` falls in the current window and
  its own labels (`docker volume inspect V`) lack `com.docker.compose.project` (Compose named volumes keep project
  teardown); a `com.docker.volume.anonymous` label is recorded, not relied on.
- **Foreign events** from other runs on the daemon are ignored and never FAIL. An unbound volume is never removed.

**Markers instead of fences.** After every witnessed call or block on C1 P3's list, and at each scope boundary, the
caller holds `flock` on `$WORK/res-witness/marker.lock`, reads the parser's ack, runs `docker top W` (a documented
event on W's exact ID that creates nothing) and waits up to 10 s for a larger ack. The parser writes the ack by atomic
rename and the window's events to `win-<n>.tsv`. With no shell counter, markers work inside `$(...)` (run.sh:2856, 2859,
5284, 5286, 5337) and in nested shells. Per window:
- **Volume removal.** Each V created in the window, bound to an attributed C and not Compose-named is handled once C is
  gone (`destroy`, or `inspect` "No such container"). If V's `destroy` was seen, `docker volume inspect V` must fail
  "No such volume"; otherwise `docker volume rm V`, then the same check (P5's 5 s bound). An unverified removal is an
  accounting failure (D2), and the row stays for finalize.
- **`unbound-volume`** is an accounting failure: an attributed C's image (its `image` attribute, P1's cached inspect)
  declares a `Config.Volumes` destination, no `mount` names it before C's `destroy`, and C cannot be inspected.
- **Stream failure** (marker timeout, parser or stream EOF, W not running) is an accounting failure, never "unknown".
  The parser lives for the whole run, so late events are still recorded.

**Instrumentation counted.** Per run: 1 container (W), 0 networks, 0 volumes (D), and 2 host processes. Per witnessed
call: 0 containers, 1 `docker top`, 1 ack wait, 1 inspect per window volume. It replaces C1 P3's 2 fence creations per
call: 28 at the 14 static sites (D, one execution each when selected; loops and blocks add U more). Cost unmeasured
(U); F1 reports it. resources.tsv gets an `instrumentation` scope; W is excluded by exact ID from the other scopes and
included in run-wide totals.

### D2. Persistent latch and serialized check-and-create (replaces C2's `RES_HOLD` variable)

lib.sh:387 already notes that callers run inside `$(...)`, "where a counter would not persist". The hold is therefore
files under RESDIR = `$(dirname "$OWNED_REG")`, the registry directory (mode 700, validate.sh:176; checked by
run.sh:89-91). A run without a registry uses `$WORK/res` and makes no accounting claim. CI controls use the run's RESDIR
(`CI_CONTROLS_OWNED_REG`, C4) and `$ROOT/owned/` per control (vc@25e878e9).
- **Latch.** `hold.d/<scope>` or `hold.d/run` is created by `(umask 077; set -C; printf 'scope\t…\nreason\t…\n…'
  >FILE)`: noclobber is O_EXCL, the primitive of marker_write/marker_stage (validate.sh:877, 890). A later reason
  appends one line. Existence means held; an unreadable `hold.d` is held. cleanup-owned.sh reads only `owned.tsv`,
  `token`, `cid/` and `start-inventory.ids` (cleanup-owned.sh:115-191), so these files are inert to it; it is unchanged.
- **Lifetime.** Only the main shell moves `hold.d/<scope>` to `held.d/`, at scope end after C2's verified zero
  foreground count; otherwise it writes `hold.d/run`, which is never lifted in the run. It lasts until verified cleanup
  removes the registry (validate.sh:1829-1830), or stays in a kept registry for do_recovery (validate.sh:961-1020).
- **Lock.** `flock`, the existing primitive of `infra/recovery/recovery.sh:77-78` and the harness-lock check
  (`G064-s6-b4-identity-driver.sh:177-179`); the kernel releases it on death. Each allocator runs `( flock -w 30 9 ||
  refuse; … ) 9>>"$RESDIR/res-<scope>.lock"` (30 s = validate.sh:988's Docker query bound) and, under it: (1) checks the
  latch; (2) counts live objects of the scope and class (C1 observation query, W excluded) plus reservation cidfile IDs
  plus reservations without an ID, a query failure refusing and latching; (3) refuses and latches `<scope> ceiling` if
  count + 1 exceeds the ceiling; (4) else writes a `set -C` reservation `rsv.d/<scope>/<BASHPID>-<hex>` (owner PID,
  start ticks, cidfile path). Then it unlocks, creates, appends the registry row and drops the reservation on return. A
  dead owner's reservation is dropped under the lock; its container, if any, is counted live. Compose calls,
  documented blocks and product-code creators are latch-checked before and recounted after under the same lock (fixed
  per-call counts, A5); a breach latches as in C2.
- **Readers** (latch-checked; capacity reserved or recounted): lib.sh:383 `docker_run_owned` and its create mode;
  lib.sh:297 `compose_call` for `up|run|create|start|restart`; lib.sh:415 `ledger_run_block` when a block line
  allocates or is unparseable (as `_lg_block_buildable`, lib.sh:2210); C1 P1's routed raw sites (lib.sh:520, 523, 529;
  run.sh:1700, 1702, 5817, 6093, 6149) and run.sh:534 via compose_call; the `rd_tool`/`rd_tool_bg` call sites
  (run.sh:4094, 4097; recovery.sh unchanged); C3's `it_runner_start` (its "RES_HOLD is unset" becomes this latch); D1's
  witness start; validate-controls.sh `D()` (vc:213) for persisting creations and the shim's `run|create` plus
  `network create|volume create` (vc@25e878e9:485-527), locked on `$ROOT/owned/<ctl>.lock`, reading both the control
  latch and `hold.d/run`.
- **Never gated:** teardown (Compose `down|stop|rm|kill`; Docker `rm`, `volume rm`, `network rm`; lib.sh:2698;
  cleanup-owned.sh; `owned_cleanup`, vc@25e878e9:400-445; witness and runner stops). validate.sh's post-harness
  evidence helpers (validate.sh:1202, 1232, 1510, 1718) log `hold.d/run` and still run: serial, `--network none`,
  verified gone (P5).
- **Outcome preserved.** A refusal is an extra `held` finding. The first FAIL row and the harness status remain
  (validate.sh:1855-1856; `do_complete` 14), and teardown always proceeds.

| C2 row (only these change) | Was | Now (D) | Derivation |
|---|---|---|---|
| V18l-n | 40/0/0 | 39/0/0 | The serialized check removes the race term: 38 holder creations (C2, L ≤ 74 s) + 1 concurrent preflight/helper transient |
| instrumentation | none | 1/0/0 | W (D1) |
| Run-wide | 63/5/12 | 63/5/12 | 15 (S-IDN) + 39 + 8 control-labelled + 1 W; networks 3 + 2; volumes 10 + 2 |

### D3. Recovery state on persistent private storage (C4)

Today both the marker directory `${TMPDIR:-/tmp}/g003-pending` and the default evidence directory, with its registry
`$EV/owned` (`mktemp -d ${TMPDIR:-/tmp}/g003-validate-XXXXXX`), are volatile (validate.sh:197, 249, 256). do_recovery
passes when no marker exists (validate.sh:964-965, 1019), so a wiped TMPDIR silently starts a new batch. Local runs
change as follows; CI writes no marker and is unchanged.
- **STATE_ROOT** = `${G003_STATE_ROOT:-${XDG_STATE_HOME:-$HOME/.local/state}/platform-evidence/g003-local}`, inside the
  private evidence tree already used for task storage (PROCESS-repository-local-storage-20261017.md:30-37). Checks as
  `pend_dir_ok` (validate.sh:863-869): mode 0700, owned, not a symlink, outside the repository (validate.sh:198), and
  `stat -f -c %T` not tmpfs/ramfs; a failure is location 16. validate.sh:256 sets `PEND_DIR="$STATE_ROOT/g003-pending"`
  and validate.sh:249 puts the default evidence directory under STATE_ROOT; never `/tmp`. A local `--evidence-dir` on
  tmpfs/ramfs is refused. Formats, names, marker_stage, marker_finish and do_scrub are unchanged.
- **Provenance.** `MARKER_BASE` (validate.sh:876) gains `boot_id` (from `/proc/sys/kernel/random/boot_id`; unreadable
  means 16) and `repo` (REPO_R) rows, in the same key/value format read by `marker_get`. The registry already binds job,
  token and repo (validate.sh:187).
- **do_recovery, per marker:**
  - Unreadable, malformed or missing `boot_id`: 16, marker kept.
  - `repo` other than REPO_R: 16, "recover from checkout <repo>"; today it ends 16 too, via cleanup-owned.sh:170.
  - `boot_id` other than now: the owner is dead by definition. Skip the PID check (:977), `recovery_survivors` (:983)
    and the `launching` block (:984), then take the unchanged Docker/registry/cleanup-owned path (:987-1017): verified
    recovery writes the delayed-recovery row and removes the marker, anything else is 16. Never "clean".
  - Pre-change markers in `${TMPDIR:-/tmp}/g003-pending` are still read, under today's same-boot rules; never written.
- **Cleanup boundary (unchanged).** The marker goes only through `marker_finish` after registry removal
  (validate.sh:896-900) or verified recovery (:1010); do_scrub deletes a default evidence directory after verified
  cleanup (:1832-1837); a recovered registry stays as evidence until acceptance cleanup or operator removal, never by
  age. Review flag: a kept registry's generated-secret list (mode 0700 directory) now survives reboots until recovery
  removes it (:1009); same lifecycle, not a security-model change.
- **Reuse and controls.** C4's self-registered validate-controls.sh and the F1 driver extract the STATE_ROOT and boot_id
  lines too. BASE_ENV (vc:436) sets `G003_STATE_ROOT=/nonexistent/g003-controls`, so a case-tree validate.sh fails
  location 16 rather than touch the operator's root; the validate.sh launches (vc:806, 1772, 2679) set
  `G003_STATE_ROOT="$C/state"`; `pend_marker` (vc:3715) and the planted-marker helper (vc:3894-3900, now with `boot_id`
  and `repo`) follow.
- **Outside the recovery guarantee.** Survivors with no recovery state at all (a deleted root, pre-change runs, direct
  runs without a registry) remain the containment domain of the item-4 guard (g064-corr11d, not inspected, not
  redesigned) and of C4's one-time OPERATIONS.md procedure.

### D4. Focused proofs (C5 delta)

| Step | Adds |
|---|---|
| F0 (stack-free, fake `docker`) | (a) A lowered ceiling breaches inside `x="$(docker_run_owned …)"` and writes `hold.d/<scope>`. The parent's next docker_run_owned and `compose_call up` are refused with no create in the fake log, `compose down` still runs, the first FAIL row stays and the exit is non-zero. The same holds for an accounting failure inside `bash -c 'source lib.sh; …'` (fake `ps` fails; `hold.d/run`). (b) Overlap: 8 background creators, ceiling 3, fake create delayed 0.2 s, 20 repetitions: exactly 3 live, 5 refused, never 4. A KILLed reserver's reservation is dropped; a lock timeout refuses and latches. (c) The parser on scripted lines: container plus volume via mount; foreign objects ignored; label-like volume attributes never attribute; `unbound-volume`; EOF or marker timeout latches; the ack works from `$(...)` |
| F2 | V18n daemon-free driver cases `race` (two shim creators at ceiling − 1: exactly one creates) and `nested-hold` (a latch set in a nested case shell; the parent shim refuses). V18l-n under 39. VR1-VR8 and VS1-VS5 are rerun for the state-root change; C5's reuse of VR3-VR5/VR7/VR8 is withdrawn |
| VR9, extended (real validate.sh, root `$C/state`) | (d) Crash as in (a), then rewrite the marker's `boot_id`: the next run recovers runner and container, writes the delayed-recovery row and removes the marker. (e) The same with the registry deleted: 16, `expect_runs 1`, the surviving labelled container untouched. (f) Marker mode 000 with a surviving container: 16, no harness, container untouched. (g) Planted markers (helper vc:3894-3900): another `boot_id` with a live PID of this boot as `harness_pgid` → process checks skipped, registry recovered; malformed `boot_id` or repo mismatch → 16, marker kept; a legacy-directory marker is processed. The same-boot SIGKILL proof (a) stays; nothing in (a)-(g) is claimed as reboot proof |
| F1 (7), real, later by the coordinator in F1's pool slot | `bash management/v0/results/G064-br-proof-driver.sh <checkout> <private evidence dir> --only witness-volume`. Setup: registry and marker (C4), witness, fixture project `g064wv-<hex>` at `$WORK/wv` (`.env`; `infra/compose/wv.yml` with service `t`, image `pgvector/pgvector:0.8.7-pg17` = `$RD_PGIMG` at run.sh:4716, entrypoint `["true"]`, no volumes), then `compose_call "$WORK/wv" infra/compose/wv.yml -- run --rm --no-deps t`. Expected events, all before the marker `container top W`: `volume create V`, `container create C project=g064wv-<hex>`, `volume mount V container=C destination=/var/lib/postgresql/data`, `container destroy C`. `owned.tsv` holds `container C` and `volume V` by the time the marker is acknowledged. resources.tsv: `created volume V mount:C:/var/lib/postgresql/data`, `gone volume V call-end removed_by=compose\|harness`, scope volume creations 1, final_leftover 0, unaccounted 0, instrumentation 1 container and 0 left. `docker volume inspect V` then fails "No such volume". A project network is reported, not asserted (U). The driver prints `WV PASS container=<C12> volume=<V12> removed_by=<…>` |

### D5. Allow-list, acceptance delta, could not verify

**Allow-list:** no additional pattern; all edits stay in C6's listed paths. `cleanup-owned.sh`, `infra/ci/lib.sh` and
`required-checks.tsv` are unchanged. No new required check ID, Goal, budget change or intermediate full run.

| Row (amends C6's acceptance delta) | Change |
|---|---|
| C1 P3 / BR-AC4 | D1: positive container and volume attribution, `top` markers, verified volume removal, instrumentation counted |
| C2 / BR-AC1 | D2: file latch, flock check-and-create, the listed readers; V18l-n 39; instrumentation 1/0/0 |
| C4 / AC6 | D3: persistent STATE_ROOT marker and registry, boot/repo provenance, boot-transition recovery, refusals |
| C5 / BR-AC4, BR-AC5 | D4: F0 (a)-(c), F2 additions and reruns, VR9 (d)-(g), F1 (7) |

**Could not verify** (no Docker run; a refutation stops the affected step for a Plan correction): labels in container
event attributes and none in volume events (the page is silent); `.Actor.Attributes`/`index` in `--format` on CLI
27.5.1 and the hosted 29.8.2; daemon-order delivery of a call's create/mount before the later `top` marker; Compose
`run --rm`'s own volume removal; busybox:1.37.0 having no `VOLUME` and accepting `sleep 2147483647`; `flock` in the
hosted context (`buildpack-deps:trixie-scm`, util-linux; S-B0 runs `command -v flock`); `stat -f` types on WSL2 and
`boot_id` changing on `wsl --shutdown`; the item-4 guard's interface on g064-corr11d.
