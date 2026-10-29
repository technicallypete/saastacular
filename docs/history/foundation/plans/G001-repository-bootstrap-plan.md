> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G001 — Repository bootstrap implementation plan

Goal: [G001-repository-bootstrap.md](../goals/G001-repository-bootstrap.md)
Spec: [G001-repository-bootstrap-spec.md](../specs/G001-repository-bootstrap-spec.md)
Owner: Claude Code · Reviewer: Codex · Budget class: **B5** (5 shared meaningful cycles; checkpoint after cycle 3)

| Revision | Date | Change |
|---|---|---|
| r1 | 2026-10-04 | Initial Plan, written after inspecting the repository. Submitted for Codex plan review. Nothing has been rendered or implemented. |
| r2 | 2026-10-04 | Codex plan review: CORRECTIONS REQUIRED (three bounded areas). (1) Metadata privileges are now protected at creation time, inside goose's own creation transaction, for exact objects (§8.5, §8.4, §11.1–§11.3). (2) The S0 starting-state gate permits and hashes only this Plan file and does not require committing it (§13, §11.3). (3) Caches and other Docker resources are execution-owned and uniquely named; pre-existing ones are preserved (§17, §13). Q1–Q3 are recorded as resolved (§0, §3.2, §9.2). No other scope changed. Still not rendered or implemented. |

## 0. Plan review record

**r1 → Codex: CORRECTIONS REQUIRED (2026-10-04).** Corrections and where r2 addresses them:

| # | Codex correction | r2 change |
|---|---|---|
| 1 | Fresh `status` creates metadata before the baseline revoke runs. Protect at creation time, so metadata and its sequence never acquire runtime privileges, including before the baseline and on fresh `status`/`version`. Target the exact objects, not a name prefix. Add pre-baseline privilege assertions and keep them after rollback and reapply. | §8.5 rewritten: a goose `Store` wrapper revokes non-owner privileges on exactly `app.goose_db_version` and its identity sequence, and verifies the result, inside the same transaction that creates the table. The baseline's prefix-based revoke is removed; the baseline now only asserts (§8.4). Privilege assertion P runs before the baseline and after every lifecycle step (§11.2, B-MIGRATE-CATALOG). |
| 2 | The Plan is untracked. Permit and hash that exact artifact, reject unrelated changes, record the starting commit and working-tree inventory, and do not require committing the Plan. | S0 gate rewritten (§13). Harness `--expect-hash` keeps verifying the Plan hash, whether committed or not (§11.3). |
| 3 | Replace the fixed `g001-gomod-cache` with a uniquely named, execution-owned cache, or preserve any pre-existing cache. Remove only resources created by this execution. | Execution ID, labelled `g001-gomod-<exec-id>` volume, S0 Docker inventory, and inventory-diff cleanup (§13, §17). |

**Resolved questions:**

- **Q1: yes.** A tested factory that core startup does not call satisfies dormant construction. Transport stays absent; future instances are session-scoped (§9.2).
- **Q2: yes.** goose is adopted without an ADR, as the pinned migration dependency authorized by R4 (§8.1).
- **Q3: yes.** The variant uses disabled archetype inputs plus identical applicable G001 additions, keeps its own manifest, and verifies that those additions match (§3.2, B-TOGGLE-OFF).

B5 remains suitable (Codex).

## 1. Summary

G001 renders the accepted `project-archetype` Skill, unchanged, into this repository. A small project-owned layer then adds three things:

- containerized migration tooling (goose v3.28.0) with an empty-schema baseline;
- a dormant MCP server construction boundary (official Go SDK v1.8.0) with in-process tests;
- a root README, a `.env` configure helper, foundation docs, and a project validation harness that tests the actual repository.

The approach keeps the renderer's output close to byte-identical:

- Only **3 rendered files** are deliberately customized: `core/go.mod`, `core/go.sum` and `docs/foundation/FOLLOW-UP.md`. These are listed in a customization register and are the only expected repeat-render conflicts.
- Everything else G001 adds lives in **new files**. The migration service is a Compose *overlay*, so the rendered `local.yml` stays untouched.

No tenant or product schema is created, no MCP transport is mounted, and the Skill is not modified.

## 2. Inspection findings

### 2.1 Starting state (verified 2026-10-04)

- `HEAD` is `3c45545` ("Goal 001"). It adds only the G001 Goal and Spec on top of the Spec's recorded baseline `bf0f20b`.
  - The working tree was clean before this Plan.
  - The actual starting commit used for execution is recorded at render time (R7).
- Root `web/`, `core/`, `intelligence/`, `contracts/`, `infra/`, `README.md`, `.env*` and `.gitignore` do not exist.
- `docs/SECURITY.md` and `docs/DECISIONS.md` are empty. No ADRs exist.
- **Docker:** Engine 27.5.1, Compose 2.32.4. No `beginify` containers, volumes or networks exist. Default host ports 3100, 8180, 8100, 5532, 6479, 9100 and 9101 are free.
- **Secret scan baseline:** gitleaks v8.30.1 with the template's `gitleaks.toml`, run read-only over the whole repository, found no leaks. So the rendered `secret-scan` check has a clean baseline even though it scans `management/`, `docs/` and `skills/`.

### 2.2 What G000 provides (inherited; reused, not rebuilt)

| Area | Provided by the accepted Skill (`skills/project-archetype/`, v0.1.0) |
|---|---|
| Renderer | `generator/render.sh` (containerized Go). Commands validate, plan and render. Content-addressed manifest. Collision policy `create`/`unchanged`/`conflict`/`ensure-lines`. Never writes protected paths. |
| Services | Go core: config, health/readiness, redacted logging, correlation. Next.js web: semantic tokens, server-side core bridge. FastAPI intelligence: envelope auth, ModelGateway/OpenRouter adapter, typed events. Contracts: envelope and stream-event schemas. |
| Infrastructure | `infra/compose/local.yml`: edge, data and inference networks; loopback standard+100 ports; `${VAR:?}` fail-closed environment; per-service secret scoping; check profile; `secret-scan`. Dockerfiles with per-Dockerfile `.dockerignore`. Source-built MinIO plus scoped `minio-init`. `infra/postgres/init/10-roles.sh` (owner NOLOGIN, migrator via `SET ROLE` with `INHERIT FALSE`, runtime, retention, no PUBLIC rights). |
| Docs | `docs/foundation/GETTING-STARTED.md` (marked start, check and stop blocks), `FOLLOW-UP.md` (STAGE-* obligations), `VENDORS.md`. |
| Evidence / limits | G000 Result: template-enforced vs fixture-tested vs prescribed requirements. Isolation fixtures are contract demonstrations only. Residuals: tag-only pins for alpine and distroless; MinIO is an archived upstream (product-owner ADR pending). |

G000 did **not** provide a migration runner, MCP initialization in core, a root README, a configure helper, or validation of a real consumer repository.

### 2.3 Renderer preview (dry run against this repository)

I ran these with the Spec's R2 values. The input file was kept in the session scratchpad, outside the repository.

- `render.sh validate`: exit 0. Resolved configuration:
  `{"product_name":"Beginify","project_slug":"beginify","platform_base_domain":"foundation.test","custom_domain_cname_target":"custom.foundation.test","database_name":"beginify","intelligence_enabled":true,"identity_provider":"oidc","production_object_store":"r2","ports":{"web":3100,"core":8180,"intelligence":8100,"postgres":5532,"redis":6479,"minio_api":9100,"minio_console":9101}}`. `ai_models` is omitted.
- **`output_id` = `55bb52881a31a589`.**
- `render.sh plan --dest <repo>`: exit 0. **69 `create`, 0 `conflict`, 0 `unchanged`**. A sha256 of every file before and after showed the tree unchanged. The full action list is in Appendix A. No protected path (`AGENTS.md`, `CLAUDE.md`, `README.md`, `docs/ARCHITECTURE.md`, `management/`, `skills/`) is touched.

### 2.4 Dependency research (read-only, containerized)

I ran this in `golang:1.27.1-alpine`, in a throwaway module outside the repository:

- Latest versions: `github.com/pressly/goose/v3` **v3.28.0**; `github.com/modelcontextprotocol/go-sdk` **v1.8.0**, the same pin as the G000 MCP fixtures.
- Both resolve together with the template's `pgx/v5 v5.11.0` under Go 1.27.1.
- goose's Provider API offers `Status`, `Up`, `UpByOne`, `Down`, `DownTo`, `GetDBVersion`, a session locker and a custom store/table.
- The SDK provides `NewInMemoryTransports` and `ServerOptions.Capabilities`.

### 2.5 Discrepancies found (none blocking)

1. **Baseline commit.** The Spec names `bf0f20b`, but `HEAD` is `3c45545`, which only adds the G001 Goal and Spec. I record the real starting commit; this is not a conflict.
2. **MCP wording.** The Skill's FOLLOW-UP says "Add the MCP server initialization in `core` … only together with authentication". The Spec (R5) reconciles this explicitly: dormant construction now, and transport mounting with authentication later. I follow the Spec (§9.2, Q1).
3. **Hot reload.** The Skill lists hot reload under STAGE-BOOTSTRAP, but the Goal excludes it. It stays unchecked in FOLLOW-UP and is recorded as deferred; it is not done in G001.

None of these is a Goal or Spec conflict, a completed-work overlap, or a reason to change B5.

## 3. Implementation approach

1. **Render once, unmodified.** Run the accepted renderer (validate, then plan, then render) into the repository root with the R2 inputs. Commit the content-addressed manifest `.archetype/outputs/55bb52881a31a589/manifest.json` as the initial provenance. The renderer's `--report` JSON and logs stay outside the repository as Result evidence.
2. **Add, don't edit.** G001 functionality goes into new files:
   - migration runner and baseline in `core/`;
   - MCP construction in `core/internal/mcp`;
   - Compose overlay and Dockerfile in `infra/`;
   - configure helper, harness and docs.

   Rendered files are customized only where unavoidable:
   - Go module metadata, for the new dependencies;
   - `FOLLOW-UP.md` labels, which R6 requires.
3. **Customization register.** `infra/validation/render-customizations.tsv` lists each deliberately customized rendered path and the reason. The harness asserts that a repeat `plan` reports `conflict` for exactly these paths and `unchanged` for every other rendered path. This explains each conflict and catches drift. The manifest is never edited: provenance describes the template output only.
4. **The overlay keeps the toggle clean.** All G001 additions are independent of the intelligence toggle. The disabled variant can therefore be composed as "disabled render + the same G001 additions" without a destructive toggle on the root.
5. **Project validation lives in the project.** `infra/validation/` tests the actual repository and its disposable copies. Its helpers are adapted from the Skill's harness (§11.3) and attributed. It does not source the Skill's internals, so later Skill revisions cannot silently change this repository's acceptance checks. The Skill's own harness is not run, because the Skill is provably unchanged (§11.3, B-SKILL-UNCHANGED).

### 3.1 What G001 adds, inherits and defers

| | Items |
|---|---|
| **Inherited (rendered, unchanged)** | 65 of the 68 rendered files (plus the manifest) byte-identical: all service skeletons, Compose, Dockerfiles, role init, MinIO init, gitleaks configuration, `.env.example`, `.gitignore`, GETTING-STARTED, VENDORS, contracts, intelligence |
| **Inherited (rendered, customized)** | `core/go.mod`, `core/go.sum` (dependencies only); `docs/foundation/FOLLOW-UP.md` (G001 completion labels only) |
| **Added by G001** | Migration runner, CLI and image; baseline migration; Compose `migrate` overlay; dormant MCP construction and tests; configure helper; root README; `docs/foundation/{BOOTSTRAP,MIGRATIONS}.md`; project validation harness and customization register; G001 Plan, Result and scope allow-list |
| **Deferred (unchanged obligations)** | Each to its roadmap goal: web shell/shadcn (G002); CI, metrics, contract validation (G003); identity and sessions, plus the MCP transport mount with authentication (G004); tenancy schema (G005–G006); policy/RLS application (G007–G008); storage authorization (G009); audit (G011); MCP context switching and dynamic tools (G014–G015); domains (G016); hot reload (not assigned; it stays an unchecked STAGE-BOOTSTRAP item); MinIO replacement (product-owner ADR) |

### 3.2 Intelligence-disabled consumer variant (Q3 resolved)

**Composition.** The variant is built in the harness work directory, outside the repository:

1. Render the root manifest's inputs with `intelligence_enabled: false` into an empty disposable directory. The variant keeps **its own** `.archetype/outputs/<variant-id>/manifest.json`.
2. Copy in the **applicable** G001 additions, byte-identical to the root's.

**Applicable additions:**

- the three customized rendered files: `core/go.mod`, `core/go.sum`, `docs/foundation/FOLLOW-UP.md`;
- `core/cmd/migrate/**`, `core/internal/migrate/**`, `core/internal/mcp/**`, `core/migrations/**`;
- `infra/compose/migrate.yml`, `infra/docker/migrate.Dockerfile`, `infra/docker/migrate.Dockerfile.dockerignore`, `infra/local/configure-env.sh`;
- `docs/foundation/MIGRATIONS.md`.

**Not applicable, with reasons:**

| Addition | Why it is excluded |
|---|---|
| `README.md` | Documents the intelligence-enabled root; its check block names `intelligence-check` |
| `docs/foundation/BOOTSTRAP.md` | Records the root's provenance (`55bb…`) |
| `infra/validation/**` | The harness runs from the root |
| The root manifest | The variant has its own |

Verification: B-TOGGLE-OFF-ADDITIONS and B-TOGGLE-OFF (§11.3). The root's `intelligence/` service is never removed or toggled.

## 4. Files and modules

### 4.1 Rendered (69 actions, Appendix A)

These are written by `render.sh render`: 68 files plus the manifest. 65 files stay byte-identical.

### 4.2 Rendered files deliberately customized (expected repeat-render conflicts)

| Path | Customization |
|---|---|
| `core/go.mod` | Add `github.com/pressly/goose/v3 v3.28.0` and `github.com/modelcontextprotocol/go-sdk v1.8.0`. Indirect requirements are settled by `go mod tidy` in `golang:1.27.1-alpine` (§10). |
| `core/go.sum` | Checksums for the above (`go mod verify` in the image build). |
| `docs/foundation/FOLLOW-UP.md` | STAGE-BOOTSTRAP only. Check "adopt the rendered layout" and "choose and pin a migration tool", each with a pointer to MIGRATIONS/BOOTSTRAP. Annotate the MCP item "dormant construction only (G001); mount with authentication remains open". Leave hot reload unchecked. No other stage is touched. |

### 4.3 New project files

| Path | Purpose |
|---|---|
| `README.md` | Root entry point, identity-neutral. Prerequisites; marked executable blocks `<!-- platform:configure|start|migrate|check|stop|reset -->`; port table; reset implications; links to foundation docs and governance. |
| `core/cmd/migrate/main.go` | Migration CLI (`status`, `up`, `down`, `reapply`, `version`); §9.1. |
| `core/internal/migrate/runner.go` | Configuration from environment; migrator connection with `SET ROLE <db>_owner` on every connection; identity preflight; goose Provider over the embedded set; redacted output. |
| `core/internal/migrate/metadata.go` | Protected metadata store: wraps goose's Postgres store and secures the exact metadata objects inside the creating transaction (§8.5). |
| `core/internal/migrate/runner_test.go` | Unit tests: argument and configuration validation, role-name derivation, identifier quoting, production guard, redaction, embedded set contents. No database needed. |
| `core/internal/migrate/integration_test.go` | `//go:build integration`. Lifecycle and role assertions against a disposable PostgreSQL (§11.2). Never compiled by `core-check`. |
| `core/internal/migrate/testdata/synthetic/90000001_synthetic_ddl_probe.sql` | Test-only reversible DDL probe. Never embedded in the normal set. |
| `core/migrations/embed.go` | `package migrations`; `//go:embed *.sql` is the normal migration set. |
| `core/migrations/00001_baseline.sql` | Empty-schema baseline (§8.4). |
| `core/migrations/migrations_test.go` | Every file has goose Up/Down annotations and versioned naming; the baseline contains no `CREATE TABLE/VIEW/FUNCTION/POLICY`. |
| `core/internal/mcp/server.go` | Dormant MCP construction boundary (§9.2). |
| `core/internal/mcp/server_test.go` | In-process SDK lifecycle tests (§11.1). |
| `infra/compose/migrate.yml` | Overlay adding the one-shot `migrate` service (profile `migrate`, `data` network only, migrator secret only). |
| `infra/docker/migrate.Dockerfile` + `.dockerignore` | Builds `core/cmd/migrate` from the `core/` context onto the same pinned bases as core. The `.dockerignore` excludes `.git`, `.env` and `.env.*`. |
| `infra/local/configure-env.sh` | Creates `.env` from `.env.example` (§6.2). |
| `infra/validation/run.sh`, `infra/validation/lib.sh` | Project validation harness (§11.3). |
| `infra/validation/render-customizations.tsv` | Customization register (§3, item 3). |
| `docs/foundation/BOOTSTRAP.md` | Provenance (output ID, manifest, how to re-plan); inherited, added and deferred table; dormant MCP boundary; validation-harness usage. |
| `docs/foundation/MIGRATIONS.md` | Tool, roles, metadata classification, commands (marked blocks), adding migrations, rollback and reset semantics. |
| `management/v0/plans/G001-repository-bootstrap-plan.md` | This Plan. |
| `management/v0/results/G001-repository-bootstrap-result.md` | Result (after validation). |
| `management/v0/results/G001-scope-allowlist.txt` | Exact path patterns that G001 may change (B-SCOPE input), mirroring this section. |

The following are **never created or modified**:

- `AGENTS.md`, `CLAUDE.md`, `docs/ARCHITECTURE.md`, `docs/SECURITY.md`, `docs/DECISIONS.md`, `docs/history/**`;
- `management/**`, except the three G001 artifacts above;
- `skills/**`;
- any other rendered file.

`core/cmd/server/main.go` stays byte-identical (§9.2).

## 5. Database and schema changes

- **No tenant, product or domain schema.** No Identity, Organization, Group, BusinessEntity, RoleBinding, audit or product objects are created.
- **Migration infrastructure only:**
  - goose's metadata table `app.goose_db_version` (plus its identity sequence), created and owned by `<db>_owner` through the migrator's `SET ROLE`;
  - baseline version 1, which creates no objects (§8.4).
- **Classification.** Migration metadata is **system infrastructure, not a tenant-owned domain table**. It carries no `organization_id`, gets no RLS policy, and no runtime grant. This is recorded in MIGRATIONS.md, so that the later "ENABLE + FORCE on every tenant table" checks (G008) exclude it explicitly rather than by accident.
- **Risk handling.** All destructive lifecycle tests (down, reapply, synthetic DDL) run only against uniquely named disposable Compose projects and volumes. The template role script, the `app` schema and its default privileges are unchanged. The only grant statements G001 issues are the creation-time revokes on the two exact metadata objects, in the transaction that creates them (§8.5). Migration SQL issues no grants.

## 6. Configuration

### 6.1 Inputs and environment

- **Renderer inputs:** exactly the R2 table (§2.3). These literals appear only in `.env.example`, `.archetype/outputs/55bb52881a31a589/manifest.json`, and management evidence. They must not appear in source, README, foundation docs or harness code (enforced by B-LITERALS).
- **No new environment variables.** The `migrate` service reuses `DATABASE_NAME`, `DB_MIGRATOR_PASSWORD` and `APP_ENV`, and sets fixed internal wiring (`DB_HOST=postgres`, `DB_PORT=5432`). So `.env.example` stays byte-identical, with empty secrets.
- **No change to core, web or intelligence environments.** Core keeps only `<db>_runtime` credentials. No `AI_GATEWAY_DEFAULT`, model identity or provider key is added, and `ai_models` stays omitted.

### 6.2 `infra/local/configure-env.sh`

Run from the repository root. It does the following:

- Refuses if `.env` already exists, so a developer's file is never overwritten. It explains how to recreate one.
- Runs with `umask 077`, writing to a temporary file in the repository root and then using `mv`, so `.env` ends with mode 600.
- Copies `.env.example` line by line. It fills every empty `KEY=` with 48 hex characters (`head -c 24 /dev/urandom | od -An -tx1 | tr -d ' \n'`), the format the Skill specifies.
- Leaves explicitly optional keys empty (`OPENROUTER_API_KEY`).
- Prints only the names of generated keys, never their values.

It needs only bash and coreutils, which are already prerequisites. The manual procedure in GETTING-STARTED remains valid.

## 7. Docker and runtime implications

- **Host prerequisites stay:** Git, Docker with Compose 2.24 or later, and bash with coreutils. No host Go, Python, Node, PostgreSQL, Redis or MinIO.
- **New build:** `migrate.Dockerfile`:
  - `golang:1.27.1-alpine` build stage (`CGO_ENABLED=0`, `GOTOOLCHAIN=local`, `-trimpath`, `go mod download && go mod verify`);
  - runtime on `gcr.io/distroless/static-debian12:nonroot` as non-root, with `ENTRYPOINT ["/usr/local/bin/migrate"]`.

  These are the same pins as core; **no new base image**. The migration SQL is embedded in the binary, with no bind mounts.
- **New service (overlay):** `migrate`. It has profile `migrate`, `restart: "no"`, network `data` only, and `depends_on: postgres: service_healthy`. Its environment is `DATABASE_NAME`, `DB_HOST`, `DB_PORT`, `DB_MIGRATOR_PASSWORD` and `APP_ENV`, so it never receives runtime, superuser, Redis, MinIO or AI secrets. It publishes no port.
- **Invocation:** `docker compose --env-file .env -f infra/compose/local.yml -f infra/compose/migrate.yml run --rm --build migrate <command>`.
  - The core runtime image is unchanged and contains no migration code: `migrate` is a separate binary and image.
  - Runtime startup never migrates.
- **Unchanged:** networks, published ports and loopback binding, volumes (PostgreSQL and MinIO durable; Redis ephemeral), health and readiness, check profile, and the source-built MinIO.

## 8. Migration foundation (R4)

### 8.1 Tool and pin

- **Tool:** goose `github.com/pressly/goose/v3 v3.28.0`, used as an in-process Go library through `database/sql` (`github.com/jackc/pgx/v5/stdlib`, already in the template's pgx module).
- **Why goose:**
  - It has native `status`, `up`, one-step `down`, version and re-apply primitives (re-apply = `Down` + `UpByOne`).
  - It supports SQL files with Up/Down sections, embeddable migrations, and a Postgres advisory-lock session locker.
  - It is named in the Skill's follow-up guidance.
- **Not chosen:**
  - golang-migrate has no status listing.
  - Atlas adds a separate engine and CLI image.
  - tern is viable but has no status or re-apply primitives.
- **Not infrastructure:** goose is a library inside our image, not a service or vendor (Q2).

### 8.2 Roles and privilege path

| Step | Mechanism |
|---|---|
| Login | `<db>_migrator` with `DB_MIGRATOR_PASSWORD`, over TCP to `postgres:5432` |
| Escalation | pgx `AfterConnect` runs `SET ROLE "<db>_owner"` on **every** pooled connection. The template grants `<db>_owner TO <db>_migrator WITH INHERIT FALSE, SET TRUE`. |
| Preflight (before any goose call) | `current_user = <db>_owner`, `session_user = <db>_migrator`, owner of schema `app` = `<db>_owner`, and session user is neither superuser nor BYPASSRLS. Otherwise exit 1 and do nothing. |
| Metadata creation | Inside goose's creation transaction, through the protected store. The metadata table and its sequence are committed owner-only, or not at all (§8.5). |
| Locking | goose Postgres session locker (advisory lock; no extra privilege) |
| Runtime | `<db>_runtime` is never used by the runner. Core never receives migrator credentials. |

Names derive from `DATABASE_NAME`, validated with the template's rule `^[a-z_][a-z0-9_]{0,47}$`. Identifiers are quoted with pgx's `Identifier.Sanitize`.

### 8.3 Commands (container only)

| Command | Behaviour | Exit |
|---|---|---|
| `status` | Per-migration applied/pending state and current version. goose creates the metadata table on first use of **any** command, including `status` and `version`. That creation always goes through the protected store (§8.5). | 0 / 1 |
| `up` | Apply all pending migrations in order, each in a transaction | 0 / 1 |
| `down` | Roll back the latest applied migration only. **Refused when `APP_ENV=production`.** | 0 / 1 / 2 |
| `reapply` | Roll back the latest migration and apply it again. **Refused when `APP_ENV=production`.** | 0 / 1 / 2 |
| `version` | Current version only | 0 / 1 |

Exit 2 means usage, configuration or refusal. Output never contains the DSN or password; the logger is built with the secret value for redaction (`core/internal/observability`, reused unchanged).

### 8.4 Baseline (`00001_baseline.sql`)

- **Up:**
  - a `DO` block that raises unless `current_user` owns schema `app`, so a runner that forgot `SET ROLE` fails closed in SQL as well;
  - a read-only `DO` block that raises if `app.goose_db_version` or its identity sequence (`pg_get_serial_sequence`) carries any privilege for a role other than its owner. This is defense in depth: the baseline **changes no privileges**, because protection happens at creation (§8.5);
  - no tables, views, functions, policies or grants.
- **Down:** no-op. The baseline creates nothing, and goose's metadata table is never dropped by migrations.
- **Rollback semantics:**
  - `down` from version 1 leaves version 0, with an empty `app` schema except migration metadata;
  - `up` restores version 1;
  - both are asserted by the integration tests.

### 8.5 Metadata creation-time protection (r2, correction 1)

**Problem.** The template sets `ALTER DEFAULT PRIVILEGES FOR ROLE <db>_owner IN SCHEMA app GRANT USAGE, SELECT ON SEQUENCES TO <db>_runtime`. goose creates its metadata on the first use of *any* command, which can be a fresh `status` or `version` long before the baseline runs. The identity sequence would then pick up the runtime grant when it is created.

**Verified goose v3.28.0 behaviour** (read from the module source in a container):

- Every Provider entry point used here calls `initialize`, which calls `ensureVersionTable`: `Status`, `GetDBVersion` (`version`), `Up`/`UpByOne`, `Down` and `ApplyVersion`.
- When the table is missing, goose runs `Store.CreateVersionTable(ctx, tx)` and `Store.Insert(ctx, tx, {Version: 0})` inside **one transaction**, `tx`, unless the `WithIsolateDDL` option is set. The runner never sets it.
- The Postgres DDL is `id integer PRIMARY KEY GENERATED BY DEFAULT AS IDENTITY, …`, so exactly one implicit identity sequence belongs to the table.

**Mechanism.** `core/internal/migrate/metadata.go` defines `protectedStore`. It embeds goose's `database.NewStore(DialectPostgres, "app.goose_db_version")`, keeping all of its other methods, including `TableExists`. Only `CreateVersionTable(ctx, db)` is overridden. It runs on the **same** `db` handle, which is goose's creation transaction:

1. The embedded `CreateVersionTable`.
2. Resolve the two exact objects: `'app.goose_db_version'::regclass` and `pg_get_serial_sequence('app.goose_db_version', 'id')`. **No name-prefix matching.**
3. `REVOKE ALL ON TABLE app.goose_db_version FROM PUBLIC, <runtime>, <retention>, <migrator>` and `REVOKE ALL ON SEQUENCE <that sequence> FROM PUBLIC, <runtime>, <retention>, <migrator>`. Role identifiers are derived from `DATABASE_NAME` and quoted.
4. Postcondition, in the same transaction:
   - both objects are owned by `<db>_owner`;
   - `aclexplode(coalesce(relacl, acldefault(<kind>, relowner)))` returns no grantee other than the owner.

   Any other result returns an error.

goose commits the creation only if every step succeeds. On any error the transaction rolls back and nothing is committed. So **no committed state, and no other session, can ever observe** the metadata objects with a runtime, retention, migrator or PUBLIC privilege: not on fresh `status`, fresh `version`, `up`, or concurrent first use. goose's retry treats a concurrent creator's duplicate-table failure as retryable, and the loser's transaction rolls back.

**Scope of the change.** The template's default privileges stay intact for future application sequences, which is their purpose. The protected store only affects goose's own metadata objects. No migration SQL issues grants.

**Assertion P** (used in §11.2 and B-MIGRATE-CATALOG):

- `app.goose_db_version` and its identity sequence exist, are owned by `<db>_owner`, and have owner-only ACLs.
- `has_table_privilege` is false for every privilege (SELECT, INSERT, UPDATE, DELETE, TRUNCATE, REFERENCES, TRIGGER).
- `has_sequence_privilege` is false for USAGE, SELECT and UPDATE.

Both are checked for `<db>_runtime`, `<db>_retention`, `<db>_migrator` (as session user, without `SET ROLE`) and PUBLIC. They are corroborated by real runtime attempts (`SELECT`, `nextval`), which fail with `42501`.

### 8.6 Synthetic DDL proof (test only)

`testdata/synthetic/90000001_synthetic_ddl_probe.sql` creates and drops one probe table in `app`.

- It is used only by `integration_test.go`, through an `fstest.MapFS` that combines the embedded set with this file.
- It is excluded from the normal set, which is asserted by a unit test (the embedded set is exactly `{00001_baseline.sql}`) and by B-MIGRATE-SET.

## 9. Contracts and interfaces

### 9.1 Contracts that are added

- **Migration CLI** (§8.3) and the **Compose `migrate` service** (§7) are local operator contracts, not public APIs.
- **No public HTTP or API change, and no MCP contract.** Core routes remain `GET /healthz` and `GET /readyz`. Web, intelligence and the `contracts/` schemas are unchanged.

### 9.2 Dormant MCP boundary (R5)

```go
// Package mcp is core's MCP boundary. G001: dormant construction only — no
// transport, no tools, no authentication. G004 adds authentication; G014/G015
// build one server per authenticated session context and register tools.
package mcp

type Options struct{ Title, Version string } // Title comes from configuration (no literal)

// NewServer returns an SDK server with an empty tool catalog. The tools
// capability is declared (ListChanged), but no tool is registered.
func NewServer(opts Options) *sdkmcp.Server
```

- **Not wired into `cmd/server`.** The core binary neither imports the SDK nor mounts any handler. This is deliberate:
  - The Skill reference requires **one server per authenticated session context**. A process-global server built at startup would be the pattern G014/G015 must avoid, and it would serve no purpose before authentication exists.
  - `main.go` therefore stays byte-identical to the template.
- **Functional:** the in-process SDK lifecycle tests prove it works. **Dormant:** route, dependency and HTTP checks prove it is not exposed. The alternative, constructing an unused instance at startup, was raised as Q1. **Resolved by Codex:** the tested factory, uncalled by core startup, satisfies dormant construction. Transport stays absent, and future instances are session-scoped (G014/G015).
- **Excluded:**
  - `NewStreamableHTTPHandler`, SSE handlers and any network transport;
  - any authenticator, development identity adapter or context tool;
  - any import of `example.invalid/archetype/isolation` or fixture code.

## 10. Dependencies (pinned)

| Dependency | Pin | Scope |
|---|---|---|
| `github.com/pressly/goose/v3` | v3.28.0 | `core/cmd/migrate`, `core/internal/migrate` |
| `github.com/modelcontextprotocol/go-sdk` | v1.8.0 | `core/internal/mcp` (tests only, at G001) |
| Indirect (from resolution research) | `github.com/google/jsonschema-go v0.4.3`, `github.com/mfridman/interpolate v0.0.2`, `github.com/sethvargo/go-retry v0.4.0`, `github.com/segmentio/{asm v1.2.1, encoding v0.5.4}`, `github.com/yosida95/uritemplate/v3 v3.0.2`, `go.uber.org/multierr v1.11.0`, `golang.org/x/{oauth2 v0.35.0, sys v0.47.0, time v0.15.0}` | Settled by `go mod tidy`; the final list goes in the Result |
| Version-selection upgrades | `golang.org/x/sync` v0.17.0 → v0.22.0, `golang.org/x/text` v0.29.0 → v0.41.0 (expected) | Verified by `core-check` and build |

There are no new images or vendors. Web and intelligence lockfiles are unchanged.

## 11. Test and validation strategy

### 11.1 Unit tests (inside `core-check`, `network_mode: none`)

- **`internal/mcp`.** Uses `sdkmcp.NewInMemoryTransports()` with a synthetic in-test client:
  - the initialize handshake succeeds; server info name is `core` and the title equals the supplied option;
  - the tools capability is present; `tools/list` returns 0 tools; ping succeeds;
  - session close is clean;
  - two `NewServer` calls give independent, empty catalogs, so there is no global state.
- **`internal/migrate`.** Covers:
  - command parsing and usage exit 2;
  - missing or invalid configuration names the variable and never echoes the secret;
  - role and database derivation, and identifier quoting;
  - `down`/`reapply` refused for `APP_ENV=production`;
  - the embedded set equals `{00001_baseline.sql}`, and the synthetic file is not embedded;
  - `protectedStore`:
    - targets exactly the metadata table and its serial sequence, with no prefix patterns;
    - revokes from PUBLIC and the three derived, quoted roles;
    - returns the postcondition error when any statement or check fails, which is checked with a recording `DBTxConn` fake;
    - every method except `CreateVersionTable` delegates unchanged;
  - the Provider is constructed with the protected store, and `WithIsolateDDL` is not set.
- **`migrations`.** Annotations, naming and baseline content rules (§4.3).
- **Existing:** template unit tests (config, health, observability; web lint, type checks and tests; intelligence pytest/ruff) run unchanged.

### 11.2 Integration tests (`-tags integration`, real PostgreSQL)

The harness runs these in the `core-check` image, attached to a disposable project's `data` network. Credentials go in through `--env-file`, never on the command line. "Fresh" means the metadata table does not exist. Between fresh-start cases the test drops it through the runner's own owner connection.

P is assertion P from §8.5. Sequence and assertions:

1. **Fresh `version`** (before any baseline): returns 0; the metadata was just created; **P holds**.
2. **Fresh `status`** (metadata dropped again, still before any baseline): version 0, `00001` pending; **P holds**. Real runtime `SELECT` and `nextval` on the metadata objects fail with `42501`.
3. **Creation-failure atomicity:** a test hook makes the protected store's postcondition fail on a fresh database. The hook is an unexported field set only from the package's own tests; there is no runtime flag or environment switch. The command fails, and the metadata table **does not exist** afterwards. The unprotected creation was never committed.
4. **Concurrent first use:** two runners run `status` at the same time on a fresh database. Both succeed, there is exactly one metadata table, and **P holds**.
5. **`up`:** version 1. `app` contains only metadata. Nothing in `public` is owned by any application role. **P holds.** Catalog snapshot S1 is taken, covering relations, owners and ACLs in `app` and `public`, and metadata rows.
6. **Synthetic up** (combined set): the probe table is owned by `<db>_owner`; **P holds**. Runtime negatives, each with SQLSTATE `42501`:
   - `CREATE TABLE app.x` and `CREATE TABLE public.x`;
   - `SELECT` on the probe table and on the metadata table;
   - `SET ROLE` to owner, migrator and `postgres`.

   Also: runtime is not superuser and not BYPASSRLS, is not a member of owner, and owns 0 relations. The migrator *without* `SET ROLE` cannot create in `app` (`42501`), which proves the bounded path.
7. **Synthetic down:** version 1, the probe table is gone, the catalog equals S1, and **P holds**.
8. **`reapply`:** version 1, the catalog equals S1, and **P holds**.
9. **`down`** to version 0: metadata is kept, with no other objects, and **P holds**. **`up`** again gives version 1, equal to S1, and **P holds**.
10. **Preflight negative:** a runner configured so `SET ROLE` is skipped (test hook) refuses before touching goose. No metadata is created.

### 11.3 Project harness: `infra/validation/run.sh`

**Usage:** `infra/validation/run.sh --base <start-commit> --allow-file <file> [--expect-hash PATH=SHA256]… [--keep] [--report-dir DIR] SUITE…`, where SUITE is `static | provenance | root | migrations | variants | all`.

`--expect-hash` names a file that legitimately differed from `--base` at the starting-state gate. For G001 that is only this Plan, with the sha256 recorded at S0. The file must still match that hash, whether or not it has been committed since (§13).

**Adapted from G000** (attributed): PATH shadowing of host runtimes including `/mnt/*`, `make_env`, host probes, busybox namespace probes, the corrected `forbidden_datastores_reachable`, `trials_boundary_controls`, outage, persistence, rename, docs-block execution, and the canary scan.

**Safety:**
- The work directory is outside the repository.
- **Docker inventory.** Each run takes a Docker inventory (containers, volumes, networks, images) at start. Every resource the run creates carries the run ID in its name, and the label `g001.run=<run-id>` where Docker supports labels. Disposable Compose projects are named `g001-<purpose>-<run-id>` and registered for `down -v`.
- **Cleanup.** Removes only registered projects (their containers, networks and volumes, plus images built under those project names), and only resources absent from the start inventory. Pre-existing resources are never removed or pruned (§17).
- The **root** suite first checks that no `.env`, `beginify` containers or volumes exist. If any are found it records UNMET and never deletes or overwrites. At the end it removes only what it created: the `.env` it generated (hash-verified) and its own volumes, via the README reset block.

Check IDs (all PASS/FAIL/UNMET; INFO is evidence only):

| Suite | Checks |
|---|---|
| static | **B-SHELL** (shellcheck on new scripts); **B-HADOLINT** (`migrate.Dockerfile`); **B-COMPOSE** (`config -q` for root and root+overlay; `migrate` env holds exactly the five planned keys; core, web and intelligence carry no `DB_MIGRATOR_PASSWORD` or superuser secret; intelligence has no data settings); **B-ENV-EXAMPLE** (byte-identical to render; all secrets empty; no `AI_GATEWAY_DEFAULT`); **B-LITERALS** (manifest identity values absent from platform roots, README, docs/foundation and the harness, case-insensitive); **B-NAMING** (no `frontend/backend/api/app/services` roots or redundant paths); **B-DEPS** (exact pins as §10; no `latest`); **B-MCP-SOURCE** (`go list -deps ./cmd/server` excludes the SDK and `core/internal/mcp`; neither binary depends on isolation or fixture packages; no transport or handler symbols in non-test core source; core routes exactly `/healthz` and `/readyz`); **B-MIGRATE-SET** (normal set = baseline); **B-DOCKERIGNORE** (each Dockerfile has a `.dockerignore` excluding `.env` and `.env.*`) |
| provenance | **B-RENDER-ID** (inputs reconstructed from the committed manifest re-validate to `55bb52881a31a589`); **B-RENDER-PLAN** (`plan` against the actual root: conflicts are exactly the customization register, all else `unchanged`, zero writes by tree hash); **B-RENDER-REPEAT** (`render` on a disposable copy: exit 3, the same conflicts, writes only under `.archetype/outputs/<id>/proposed/`, manifest unchanged); **B-PROTECTED** (blob hashes of `AGENTS.md`, `CLAUDE.md`, `docs/**` except `docs/foundation/**`, `management/**` except the G001 artifacts, and `skills/**` equal `--base`; each `--expect-hash` file equals its recorded sha256); **B-SKILL-UNCHANGED** (`git diff --quiet <base> -- skills/` and no untracked Skill files) |
| root | **B-DOCS-EXEC-ROOT** (README `configure`, `start`, `migrate` (status, up, status), `check`, `stop` blocks, and GETTING-STARTED `start`, `check`, `stop` blocks, executed verbatim in the actual root); **B-DEFAULTS** (published at 3100, 8180, 8100, 5532, 6479, 9100, 9101 on 127.0.0.1; `/readyz` shows all `ok` and the configured product name); **B-PORTS-ROOT** (host protocol probes; refused on a non-loopback address); **B-BOUNDARIES-ROOT** (web and intelligence cannot reach any datastore, each by name and every IP on its own port; web cannot reach intelligence; DNS; core positive controls; execute endpoint 401/422/typed stream without a gateway key); **B-BOUNDARY-CONTROLS** (postgres-only, redis-only and minio-only negative controls, real and stubbed); **B-MCP-HTTP** (POST JSON-RPC `initialize` and GET `/mcp`, `/mcp/`, `/sse`, `/messages` on core, from its own namespace and the host port, return 404 with no MCP payload; web `/mcp` also 404); **B-CHECKS-ROOT** (`core-check`, `web-check`, `intelligence-check`, `secret-scan`); **B-BUILD-CONTEXT** (an `.env` canary in the `core/` context of a disposable copy is absent from the core and migrate build stages) |
| migrations | Disposable project `g001-mig-*`, PostgreSQL only. **B-MIGRATE-IT** (§11.2, `go test -tags integration -count=1 -json`; any skip is a FAIL); **B-MIGRATE-CLI** (MIGRATIONS.md blocks verbatim, starting from a fresh volume: `version` → `status` → up → down → status → reapply → up, with outputs and exit codes asserted; `down`/`reapply` refused with `APP_ENV=production`, exit 2); **B-MIGRATE-CATALOG** (superuser catalog read, validation only, over the container socket. After the fresh `version`, after the fresh `status` and before the baseline, then after every later CLI step: assertion P (§8.5) on the exact metadata table and sequence; owners of `app`, metadata and sequence; role attributes and memberships; no runtime-owned objects); **B-MIGRATE-ENV** (`migrate` container env: only migrator credential, no runtime or superuser secret) |
| variants | **P1 `g001-ops-*`** (copy of the root implementation, ports 3200/8280/8200/5632/6579/9200/9201): **B-OUTAGE** (Redis, PostgreSQL and MinIO outage gives 503 naming the dependency, liveness 200, then recovery; intelligence outage is non-gating; core outage makes web `/api/ready` 503); **B-PERSIST** (after `migrate up`, then `down`/`up`: migration version and a MinIO object survive, a Redis key does not). **P2 `g001-ren-*`** (ports 3300/8380/8300/5732/6679/9300/9301): **B-PORTS-OVERRIDE** (published at overrides, internal DNS on 5432/6379/9000 works); **B-RENAME** (`PRODUCT_NAME`, base domain and CNAME changed via `.env`, `up --no-build`, image IDs unchanged, web and core show the new identity); **B-DBRENAME** (new `DATABASE_NAME` on a fresh volume gives `<new>_owner/_migrator/_runtime/_retention`, `migrate up` succeeds, and metadata is owned by `<new>_owner`); **B-SOURCE-HASH** (copy tree hash equals root). **P3 `g001-off-*`** (variant per §3.2, ports 3500/8580/5932/6879/9500/9501): **B-TOGGLE-OFF-ADDITIONS** (the variant keeps its own manifest and has no root manifest; its rendered files match that manifest, except the register's customized paths; for each customized path, both manifests record the same intended hash, so the same customization applies to the same rendered base; each applicable addition is byte-identical to the root's); **B-TOGGLE-OFF** (the variant's GETTING-STARTED start and check blocks and the MIGRATIONS.md blocks run verbatim, including `migrate up`, with no intelligence service, Python image, AI or intelligence variables, or `intelligence/` directory; core readiness omits intelligence) |
| final | **B-DOCKER-ONLY** (shadow log empty; no absolute runtime paths in scripts); **B-SECRET-CANARY** (every generated secret, including the root `.env`, absent from harness output, compose logs, migrate output, reports, manifests and all non-ignored repository files); **B-SCOPE** (`git status --porcelain --untracked-files=all` plus `git diff --name-only <base>` match the allow-list exactly; every path outside it is listed as a FAIL; any Codex-owned exception must be named by exact path with its diff inspected); **IMAGES** (INFO: every image and tag used) |

### 11.4 Real versus test doubles

- **Real:**
  - PostgreSQL (with the rendered role init), Redis, source-built MinIO, Compose networks and published ports;
  - the `migrate` binary and image, the goose Provider, the core, web and intelligence images;
  - host and in-namespace TCP and HTTP probes.
- **Doubles:**
  - an in-memory MCP client peer (test only);
  - the synthetic DDL migration (test only);
  - a stubbed `ns_tcp` for the independence control (as in G000);
  - no live AI provider: intelligence answers `capability_not_configured`.

## 12. Acceptance-criteria mapping

| AC | Demonstrated by | Command(s) |
|---|---|---|
| AC1 bootstrap from Skill, content preserved, additions attributable | Render evidence: starting commit, path inventory, resolved inputs, plan and render output, report JSON outside the repository, manifest. B-RENDER-ID, B-RENDER-PLAN, B-RENDER-REPEAT, B-PROTECTED, B-SKILL-UNCHANGED, customization register, BOOTSTRAP.md | `render.sh validate/plan/render` (S1); `run.sh … provenance` |
| AC2 container-only configure, start, check, stop, migrate | B-DOCS-EXEC-ROOT, B-MIGRATE-CLI, B-CHECKS-ROOT (builds, lint, types, unit tests, secret scan), B-SHELL, B-HADOLINT, B-DOCKER-ONLY (zero shadow invocations) | `run.sh … static root migrations` |
| AC3 identity, domain, database and ports configurable | B-DEFAULTS (actual root at assigned defaults); disposable copies of the implementation: B-RENAME (no rebuild, image IDs compared), B-DBRENAME (roles), B-PORTS-OVERRIDE, B-SOURCE-HASH | `run.sh … root variants` |
| AC4 health, isolation, least privilege, persistence | B-OUTAGE, B-PERSIST, B-PORTS-ROOT (loopback and non-loopback refusal), B-BOUNDARIES-ROOT, B-BOUNDARY-CONTROLS (independent per-port probes; negative and positive controls), B-COMPOSE env scoping | `run.sh … root variants static` |
| AC5 migrations under intended roles, no domain schema, no runtime authority | B-MIGRATE-IT, B-MIGRATE-CLI, B-MIGRATE-CATALOG, B-MIGRATE-ENV, B-MIGRATE-SET, unit tests | `run.sh … migrations static` |
| AC6 MCP functional and dormant | `internal/mcp` unit tests (in `core-check`), B-MCP-SOURCE, B-MCP-HTTP | `run.sh … static root` |
| AC7 optional intelligence and secret boundaries | Enabled root: intelligence-check plus execute probes without a key (B-CHECKS-ROOT, B-BOUNDARIES-ROOT). B-TOGGLE-OFF-ADDITIONS and B-TOGGLE-OFF (disabled variant, §3.2). B-ENV-EXAMPLE, B-COMPOSE, B-BUILD-CONTEXT, B-DOCKERIGNORE, B-SECRET-CANARY | `run.sh … all` |
| AC8 bounded diff, docs, Result, budget | B-SCOPE against `G001-scope-allowlist.txt`; full `git status`/diff inspection recorded; Result maps AC1–AC8, separates inherited, new and deferred work, and reports B5 usage and the cycle-3 checkpoint | Final `run.sh … all`; Result |

**Final acceptance run:**

```sh
infra/validation/run.sh --base <start-commit> \
  --allow-file management/v0/results/G001-scope-allowlist.txt \
  --expect-hash management/v0/plans/G001-repository-bootstrap-plan.md=<plan-sha256-from-S0> \
  --report-dir <outside-repo> all
```

It must exit 0 with 0 FAIL and 0 UNMET.

## 13. Implementation sequence

| Step | Work | Gate |
|---|---|---|
| S0 | Starting-state gate (below). Generate the execution ID; record the Docker inventory and the free-port check, outside the repository. | Gate passes, ports free |
| S1 | `validate`, then `plan` (expect 69 `create`), then `render --report <scratch>`. Save outputs. Expect exit 0 and the manifest at `55bb…`. | Any conflict or unexpected path means **stop** and return to Codex |
| S2 | Core: `core/migrations` (embed and baseline), `internal/migrate`, `cmd/migrate`, `internal/mcp`; `go mod tidy` in `golang:1.27.1-alpine`, with the module cache in the execution-owned volume `g001-gomod-<exec-id>` (§17); `gofmt` and `vet` in containers | Builds in a container |
| S3 | Infrastructure: `migrate.Dockerfile` with `.dockerignore`, `compose/migrate.yml`, `local/configure-env.sh` | `config -q` |
| S4 | Docs: README, MIGRATIONS, BOOTSTRAP, FOLLOW-UP labels; customization register; scope allow-list | — |
| S5 | Harness: `infra/validation/{lib.sh,run.sh}` | shellcheck |
| S6 | Validate incrementally: static, then provenance, then migrations, then root, then variants. Fix and retest (cycles counted per §15). | All suites green |
| S7 | Read-only review subagent (§14) on the full diff. Triage and fix findings. | Findings resolved or recorded |
| S8 | Final `run.sh … all` from clean state. Inspect the full `git status`/diff. Execution-owned cleanup (§17), with a post-cleanup Docker inventory diff. Write the Result. Stop for Codex. | Exit 0 |

Nothing is committed unless the user asks. No G002 or G003 work.

### 13.1 S0 starting-state gate (r2, correction 2)

Execution must not implicitly require committing the Plan. The gate:

1. **Record:**
   - the actual starting commit (`git rev-parse HEAD`) and branch;
   - the full working-tree inventory: `git status --porcelain=v1 --untracked-files=all --ignored`, plus `git ls-files -s` digested to a single sha256;
   - the sha256 of every file that differs from `HEAD` (modified, staged, untracked or ignored).

   All of this goes into the execution evidence outside the repository, and is summarized in the Result.
2. **Permit exactly one difference:** `management/v0/plans/G001-repository-bootstrap-plan.md`, whether untracked, modified relative to `HEAD`, or already committed with no difference. Its sha256 is recorded as the **approved Plan hash**, and the Result names the reviewed revision (r2 or later) it corresponds to.
3. **Reject everything else.** Any other modified, staged, deleted, untracked or ignored path means **stop before rendering** and report the inventory. There are no implicit exceptions:
   - A Codex-owned change present at that moment (for example `management/v0/PROGRESS.md`) is accepted only if Codex names that exact path in its approval. It is then hashed, its diff is recorded, and it is passed to the harness as a named exception.
   - Otherwise I stop and ask.
4. **Base for later checks.**
   - `--base` = the recorded starting commit.
   - `--expect-hash` = the approved Plan hash.
   - B-PROTECTED and B-SCOPE therefore pass whether the Plan stays untracked or is committed later unchanged, and fail if it changes during execution. Any change to the Plan during execution would itself need Codex review.

## 14. Delegation

- **During planning: no subagents.** The analysis was sequential and tightly coupled (renderer behaviour, template contracts, dependency resolution, harness reuse). Delegating it would have duplicated reading without improving the result.
- **Implementation: no implementation subagents.** `go.mod`, the runner, the docs and the harness are interdependent, and the harness exercises every piece. Parallel editing would create overlapping ownership and integration risk for little time saved. All files are owned by the main agent.
- **One read-only review subagent (SA-R) after S6.** Its brief:
  - Inputs: the Goal, Spec, this Plan and the full diff (including untracked files).
  - Review: the migration privilege path and SQL; secret handling (environment, command lines, logs, build contexts); MCP dormancy; Docker-only paths; README and doc command accuracy; scope against the allow-list.
  - Output: a findings list with severity and file:line, plus cycles used (expected 0) and blockers.
  - It edits no files.

  I validate, fix and integrate everything. SA-R's opinions do not replace Codex review.

## 15. Execution budget (B5)

- **Ceiling:** 5 meaningful implement → validate → diagnose → fix → retest cycles, shared by all agents. SA-R is read-only, so it uses 0 unless it modifies code, which it will not.
- **Counting** (BUDGET_POLICY §3–§5):
  - One cycle = a material validation failure that is diagnosed, fixed and retested.
  - Related commands against the same implementation attempt are one cycle.
  - Successful runs, formatting, inspection and the initial implementation before its first validation do not count.
  - Doubtful cases are counted.
- **Checkpoint after cycle 3:** a record in the Result of criteria completed and remaining, suite status, repeated failures, uncertainty, scope convergence, and a forecast against 5.
- **Forecast: 2–4 cycles.** The riskiest areas are harness integration (WSL host probes, ported functions in a new layout) and migration details (goose metadata creation timing, the §8.5 sequence grant). Both are mitigated: the G000 trial logic is reused rather than rebuilt, and suites are validated in increments that isolate failures.
- **B5 suitability: appropriate; no change recommended.**
  - The skeleton is already validated.
  - G001's new code is small: about 600 lines of Go including tests, about 900 lines of shell and docs.
  - Integration risk is concentrated in a harness whose logic is proven.
  - If thrashing appears (BUDGET_POLICY §8), or the cycle-3 checkpoint forecasts more than 5, I stop and return to Codex rather than extend.
- **Context/scope:** this fits one bounded implementation cycle. **No split recommended.**

## 16. Risks, interpretations and escalation points

**Questions raised in r1, all resolved by Codex on 2026-10-04** (original wording kept):

- **Q1 — Dormant MCP placement (§9.2).** Is the construction factory with in-process initialization tests, deliberately *not* invoked by `cmd/server` (so no global server), acceptable as the G001 "dormant initialization"? The alternative is to construct one unused instance at startup and keep it unmounted. That would add a process-global object, which G014/G015 must replace, and a `main.go` customization.
  - **Resolved: yes.** Keep transport absent and future instances session-scoped.
- **Q2 — goose without an ADR.** goose is an in-process Go library selected under Spec R4 ("The Plan must select and pin a migration tool"). It is not a service, vendor or infrastructure component, so I treat it as not requiring an ADR. Please confirm.
  - **Resolved: yes.** goose is the pinned migration dependency authorized by R4.
- **Q3 — Disabled-variant composition.** Is "disabled render of the same inputs + the identical G001 additions" acceptable as the "disposable intelligence-disabled consumer variant"? The root is never toggled destructively.
  - **Resolved: yes.** Use disabled archetype inputs plus identical applicable additions, keep the variant's own manifest, and verify that the additions match (§3.2).

No open questions remain in r2.

**Risks and handling:**

| # | Risk | Handling |
|---|---|---|
| R1 | Default sequence grant reaches the metadata sequence (§8.5) | r2: revoke and verify the exact table and sequence inside goose's creation transaction; the baseline only asserts; P is checked before the baseline and after every step |
| R2 | Any goose command creates metadata on a fresh database (`status`, `version`, …) | r2: every creation path goes through the protected store; fresh-`status`, fresh-`version`, failure-atomicity and concurrency tests (§11.2 steps 1–4) |
| R2a | A future goose upgrade changes the creation transaction or the DDL | Pinned at v3.28.0. The unit test plus integration steps 1–4 fail closed if the store is bypassed or creation is no longer transactional. Any upgrade is a reviewed change. |
| R3 | Default ports busy at execution time | Harness refuses (FAIL/UNMET), never silently re-maps; ask the user |
| R4 | Pre-existing developer `.env` or `beginify` resources at execution | Root suite records UNMET and touches nothing; stop and ask |
| R4a | Unrelated working-tree changes at S0 (for example an uncommitted supervisor edit) | The S0 gate stops before rendering unless Codex names the exact path (§13.1) |
| R5 | Unexpected render conflict, or render touching protected content | Stop before reconciling; return to Codex |
| R6 | An archetype defect is discovered | Stop; return to Codex for a separately bounded Skill correction. No Skill edits. |
| R7 | Spec baseline is `bf0f20b`, actual is `3c45545` | Record the actual starting commit; protected-hash base = starting commit |
| R8 | Version-selection upgrades of `x/sync` and `x/text` | Recorded; `core-check` and image builds verify |
| R9 | Inherited tag-only pins (distroless, alpine); archived MinIO | Unchanged residuals from G000; not resolved by G001 (product-owner ADR for MinIO) |
| R10 | `secret-scan` covers the whole repository | Baseline is clean. A future finding means stop, with no allow-list broadening without review. |
| R11 | Long first builds (MinIO from source, Next.js) and WSL probe timing | Build cache; same probe modes as G000 (`devtcp`/`curl`); not a scope issue |
| R12 | Codex edits `PROGRESS.md` during review | B-SCOPE reports it as an exact-path exception with its diff inspected; it is never covered by a broad allow-list |

**Escalation (stop affected work):** as Goal §Stop/Escalation. That covers architecture contradictions, scope expansion, security-model change, risky migrations, new public API or MCP contracts, new major infrastructure or vendors, irreversible data-model decisions, protected-content conflicts, required Skill changes, the inability to validate without weakening checks, budget exhaustion, and context overrun. MCP network exposure and identity bypasses are out of bounds.

## 17. Evidence handling and cleanup

- **Evidence location:** render outputs, `--report` JSON, harness reports, logs and the protected-hash snapshot go to the session scratchpad, outside the repository. They are summarized in the Result with exact commands and outcomes.
- **Secrets:** never printed, written to evidence or committed. B-SECRET-CANARY enforces this. Harness credentials go into containers through `--env-file` or files, not command-line values. The G000 residual `-e PGPASSWORD` pattern is not reused: superuser catalog reads use the container's local socket.
- **Cleanup ownership (r2, correction 3).** Only resources created by this execution are removed:
  - **Execution ID.** S0 generates `<exec-id>` (random hex) and records a Docker inventory: container, volume, network and image IDs and names. All resources this execution creates embed the ID in their name and, where Docker supports it, carry the label `g001.exec=<exec-id>`. Harness runs use their own run IDs (§11.3) under the same rules.
  - **Go module cache.**
    - S2 uses `g001-gomod-<exec-id>`.
    - Before creating it, the execution verifies that the name does not exist (`docker volume inspect` must fail), then creates it with `docker volume create --label g001.exec=<exec-id>`.
    - At S8 it is removed only if its label matches.
    - No fixed cache name is used, and no pre-existing cache volume is reused or removed.
  - **Harness projects.** The harness removes only the Compose projects it registered (`down -v`, plus `--rmi local` for images built under those project names), plus the root `.env` it generated (hash-verified). The commands are recorded.
  - **Root project.** The root project's containers, volumes and built images are removed only if they were absent from the S0 inventory.
  - **Final inventory diff.** S8 takes a final inventory and diffs it against S0. Anything created and not removed is listed in the Result.
  - **Kept and listed, not removed:**
    - base images newly pulled for builds;
    - the BuildKit cache, which is shared and cannot be attributed to one execution.
  - **Pre-existing resources are preserved and never pruned.** This includes the renderer's `project-archetype-render-cache` volume, which existed before G001 and belongs to the renderer.

## 18. Readiness

**r2: ready for Codex re-review.** All three r1 corrections are addressed (§0), and Q1–Q3 are resolved. No Goal/Spec conflict, completed-work overlap or budget mismatch was found. The discrepancies in §2.5 are recorded and are not blocking. B5 is unchanged. Implementation waits for an explicit Codex ACCEPT of this Plan revision.

*(r1 text: "Ready for Codex plan review … Implementation waits for an explicit Codex ACCEPT of this Plan, with answers to Q1–Q3 or an equivalent correction.")*

## Appendix A — Renderer dry-run actions (`plan`, output `55bb52881a31a589`, all `create`)

`.archetype/outputs/55bb52881a31a589/manifest.json`, `.env.example`, `.gitignore`, `contracts/README.md`, `contracts/intelligence/execution-envelope.schema.json`, `contracts/intelligence/stream-event.schema.json`, `core/cmd/server/main.go`, `core/go.mod`, `core/go.sum`, `core/internal/config/config.go`, `core/internal/config/config_test.go`, `core/internal/health/checks.go`, `core/internal/health/health.go`, `core/internal/health/health_test.go`, `core/internal/observability/correlation.go`, `core/internal/observability/logging.go`, `core/internal/observability/observability_test.go`, `docs/foundation/FOLLOW-UP.md`, `docs/foundation/GETTING-STARTED.md`, `docs/foundation/VENDORS.md`, `infra/compose/local.yml`, `infra/docker/core.Dockerfile`, `infra/docker/core.Dockerfile.dockerignore`, `infra/docker/intelligence.Dockerfile`, `infra/docker/intelligence.Dockerfile.dockerignore`, `infra/docker/minio.Dockerfile`, `infra/docker/web.Dockerfile`, `infra/docker/web.Dockerfile.dockerignore`, `infra/minio/init.sh`, `infra/postgres/init/10-roles.sh`, `infra/security/gitleaks.toml`, `intelligence/pyproject.toml`, `intelligence/src/main.py`, `intelligence/src/models/__init__.py`, `intelligence/src/models/resolver.py`, `intelligence/src/models/settings.py`, `intelligence/src/observability/__init__.py`, `intelligence/src/observability/export.py`, `intelligence/src/providers/__init__.py`, `intelligence/src/providers/gateway.py`, `intelligence/src/providers/openrouter.py`, `intelligence/src/security/__init__.py`, `intelligence/src/security/auth.py`, `intelligence/src/streaming/__init__.py`, `intelligence/src/streaming/events.py`, `intelligence/src/streaming/schemas.py`, `intelligence/tests/conftest.py`, `intelligence/tests/test_auth.py`, `intelligence/tests/test_config.py`, `intelligence/tests/test_execute.py`, `intelligence/tests/test_gateway.py`, `intelligence/tests/test_lint.py`, `intelligence/tests/test_observability.py`, `intelligence/tests/test_schema.py`, `intelligence/uv.lock`, `web/app/api/health/route.ts`, `web/app/api/ready/route.ts`, `web/app/globals.css`, `web/app/layout.tsx`, `web/app/page.tsx`, `web/eslint.config.mjs`, `web/lib/config.ts`, `web/lib/core-client.ts`, `web/next-env.d.ts`, `web/next.config.ts`, `web/package-lock.json`, `web/package.json`, `web/postcss.config.mjs`, `web/tsconfig.json`
