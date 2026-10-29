> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G001 — Repository bootstrap result

Goal: [G001-repository-bootstrap.md](../goals/G001-repository-bootstrap.md) · Spec: [G001-repository-bootstrap-spec.md](../specs/G001-repository-bootstrap-spec.md) · Plan: [G001-repository-bootstrap-plan.md](../plans/G001-repository-bootstrap-plan.md) (r2, approved)
Owner: Claude Code · Reviewer: Codex · Budget class: **B5**

## 1. Status

**Submitted for Codex acceptance review. Not accepted.** The final acceptance run (§6.1) exits 0 with 0 FAIL and 0 UNMET. A pre-submission check found that the Spec requires the actual repository to be runnable, and that this was already exercised. A supplementary actual-root evidence pass then made the remaining runtime evidence explicit: 34 PASS, 0 FAIL, 0 UNMET (§13). Only Codex can accept G001 or update PROGRESS.md. Nothing is committed: every G001 path is untracked on `staging` at the starting commit.

## 2. Execution record

### 2.1 S0 starting-state gate

| Item | Value |
|---|---|
| Starting commit (`--base`) | `9b4fba4d0991da1eba73766485a0f95b415af52a` on `staging` (“G001: repository bootstrap implementation plan (r2, approved)”) |
| Spec's recorded baseline | `bf0f20b`. The intervening commits add only the G001 Goal, Spec and Plan (Plan §2.5 item 1). |
| Execution ID | `ec0eadb2` (2026-10-04T21:09:59Z) |
| Working tree at S0 | `git status --porcelain=v1 --untracked-files=all --ignored` was empty; 180 tracked files; index digest `b90f739a…0c1bb` |
| Approved Plan hash | `ff8be307c544df3b471e3f50a7afc819eaead1c49d8794236a26c83fe888e048`, identical to `HEAD`. This is r2 as approved, and is verified by `--expect-hash` on every run. |
| Pre-existing local state | No `.env`; no containers or volumes for the root project; every default and disposable host port free; the renderer's `project-archetype-render-cache` and 50 unrelated volumes present and preserved |

**Evidence loss.** A WSL restart during execution wiped `/tmp`, which held the first session's scratchpad. The lost items are the S0 inventory files (full container, network and image lists), the S1 render report JSON, and the logs from the earlier runs.

The values in this section, the S1 action counts and the earlier check results were recovered verbatim from the session transcript, including the full S0 volume list.

This does not affect the cleanup guarantees:
- every resource G001 created carries `g001`, the exec ID or a run ID in its name (or the root project's Compose label);
- every harness run takes its own start inventory and diff (§7).

Evidence from the final run is in `/tmp/claude-1000/…/cac8663c…/scratchpad/g001-exec/` (outside the repository).

### 2.2 S1 render

```sh
skills/project-archetype/generator/render.sh validate --input <scratch>/render-input.json   # exit 0
skills/project-archetype/generator/render.sh plan     --input <scratch>/render-input.json --dest "$PWD"   # exit 0: 69 create, 0 conflict
skills/project-archetype/generator/render.sh render   --input <scratch>/render-input.json --dest "$PWD" --report <scratch>/s1-render-report.json   # exit 0
```

- **Inputs:** the Spec R2 values, resolved as `{"product_name":"Beginify","project_slug":"beginify","platform_base_domain":"foundation.test","custom_domain_cname_target":"custom.foundation.test","database_name":"beginify","intelligence_enabled":true,"identity_provider":"oidc","production_object_store":"r2","ports":{"web":3100,"core":8180,"intelligence":8100,"postgres":5532,"redis":6479,"minio_api":9100,"minio_console":9101}}`. `ai_models` was omitted. The input file was kept outside the repository.
- **Output ID:** `55bb52881a31a589`; Skill version `0.1.0`.
- **Actions:** render actions equalled plan actions: 69 `create` (Plan Appendix A), 0 `conflict`, 0 `unchanged`. All 180 pre-existing tracked files were unchanged afterwards. Provenance is `.archetype/outputs/55bb52881a31a589/manifest.json`, which has never been edited.
- **Reconciliation:** none was needed.

## 3. Implementation summary and files

### 3.1 Inherited, added and deferred

| | Paths |
|---|---|
| **Rendered, byte-identical (65 files + manifest)** | Plan Appendix A, except the three below. Verified by B-RENDER-PLAN and B-RENDER-REPEAT. |
| **Rendered, deliberately customized (3)** | `core/go.mod` and `core/go.sum` (goose and MCP SDK dependencies); `docs/foundation/FOLLOW-UP.md` (STAGE-BOOTSTRAP labels only: layout and migration tool checked, MCP marked “partly done: dormant construction only”, nothing beyond G001 claimed). Listed in `infra/validation/render-customizations.tsv`. |
| **Added by G001 (23 paths)** | Listed below |
| **Deferred** | As Plan §3.1: web shell/shadcn (G002); CI, metrics and contract validation (G003); identity, sessions and the MCP transport mount with authentication (G004); tenancy schema (G005–G006); policy and RLS (G007–G008); storage authorization (G009); audit (G011); MCP context and dynamic tools (G014–G015); domains (G016); hot reload (unassigned); MinIO replacement (product-owner ADR) |

The G001 additions are:
- **Migrations:** `core/cmd/migrate/main.go`; `core/internal/migrate/{runner.go, metadata.go, runner_test.go, integration_test.go, testdata/synthetic/90000001_synthetic_ddl_probe.sql}`; `core/migrations/{embed.go, 00001_baseline.sql, migrations_test.go}`.
- **MCP:** `core/internal/mcp/{server.go, server_test.go}`.
- **Infrastructure:** `infra/compose/migrate.yml`; `infra/docker/migrate.Dockerfile` and its `.dockerignore`; `infra/local/configure-env.sh`.
- **Validation:** `infra/validation/{run.sh, lib.sh, render-customizations.tsv}`.
- **Docs:** `README.md`; `docs/foundation/{BOOTSTRAP.md, MIGRATIONS.md}`.
- **Management:** `management/v0/results/{G001-scope-allowlist.txt, G001-repository-bootstrap-result.md}`.

`core/cmd/server/main.go` is byte-identical to the render. No file under `AGENTS.md`, `CLAUDE.md`, `docs/` (outside `docs/foundation/`), `management/` (outside the two Result-directory files above) or `skills/` was changed.

### 3.2 Dependencies (pinned)

- **Direct:** `github.com/pressly/goose/v3 v3.28.0`; `github.com/modelcontextprotocol/go-sdk v1.8.0`.
- **Indirect, settled by `go mod tidy` in `golang:1.27.1-alpine`:**
  - new: `google/jsonschema-go v0.4.3`, `mfridman/interpolate v0.0.2`, `segmentio/asm v1.2.1`, `segmentio/encoding v0.5.4`, `sethvargo/go-retry v0.4.0`, `yosida95/uritemplate/v3 v3.0.2`, `go.uber.org/multierr v1.11.0`, `golang.org/x/oauth2 v0.35.0`, `golang.org/x/sys v0.47.0`, `golang.org/x/time v0.15.0`;
  - version-selection upgrades: `golang.org/x/sync v0.17.0 → v0.22.0`, `golang.org/x/text v0.29.0 → v0.41.0`.

  This matches Plan §10.
- **No new images or vendors.** The `migrate` image uses core's pinned bases. Web and intelligence lockfiles are unchanged (B-DEPS).

## 4. Migrations and schema

- **Tool:** goose v3.28.0, in-process, over `pgx/v5/stdlib`. It runs as `<db>_migrator` with `SET ROLE "<db>_owner"` on every pooled connection, and connects only to `DB_HOST:DB_PORT` (pgx fallbacks are cleared).
- **Preflight:** session user, current role, `app` owner, and no superuser or BYPASSRLS. It runs before any goose call.
- **Metadata:** `app.goose_db_version` and its identity sequence are classified as system infrastructure, not tenant domain tables (MIGRATIONS.md). The protected store revokes on exactly these two objects inside goose's creation transaction, then checks the postcondition. Nothing unprotected is ever committed.
- **Baseline:** `00001_baseline.sql` only asserts (owner, and owner-only metadata ACLs). It creates no objects, its Down is a no-op, and migration SQL issues no grants.
- **No Identity, Organization, Group, BusinessEntity, RoleBinding, audit or product objects.**
- **Risk handling:** destructive lifecycle tests ran only in disposable projects `g001-mig-*` and `g001-ren-*`.

## 5. Contracts, API and MCP

- **Local operator contracts only:** the `migrate` CLI (`status|up|down|reapply|version`, exit codes 0/1/2) and the Compose `migrate` service. The service has profile `migrate`, the `data` network only, no published port, and environment `APP_ENV`, `DATABASE_NAME`, `DB_HOST`, `DB_MIGRATOR_PASSWORD` and `DB_PORT`.
- **No public API changes.** Core routes are still only `GET /healthz` and `GET /readyz`.
- **No MCP contract.** `core/internal/mcp.NewServer` is a dormant, tested factory with an empty catalog, and the core server never calls it. The server binary does not link the SDK, and no transport, authenticator, development identity or fixture is present.

## 6. Validation

### 6.1 Final acceptance run

```sh
TMPDIR=<scratchpad> infra/validation/run.sh --base 9b4fba4 \
  --allow-file management/v0/results/G001-scope-allowlist.txt \
  --expect-hash management/v0/plans/G001-repository-bootstrap-plan.md=ff8be307c544df3b471e3f50a7afc819eaead1c49d8794236a26c83fe888e048 \
  --report-dir <scratchpad>/g001-exec/final-all all
```

**Final run `21348d` (2026-10-04, 18:46:52–18:55:21 local): exit 0, `RESULT: PASSED`; 40 PASS, 0 FAIL, 0 UNMET, 3 INFO.** The implementation tree's hash was identical before and after the run, and the run left no resources (empty `inventory-leftover.txt`).

| Suite | Checks (all PASS) |
|---|---|
| static | B-SHELL, B-HADOLINT, B-COMPOSE, B-ENV-EXAMPLE, B-LITERALS, B-NAMING, B-DEPS, B-MCP-SOURCE, B-MIGRATE-SET, B-DOCKERIGNORE |
| provenance | B-RENDER-ID, B-RENDER-PLAN, B-RENDER-REPEAT, B-PROTECTED (180 files plus the Plan hash), B-SKILL-UNCHANGED |
| root (actual repository) | B-CONFIGURE, B-DEFAULTS, B-PORTS-ROOT, B-BOUNDARIES-ROOT, B-BOUNDARY-CONTROLS, B-MCP-HTTP, B-CHECKS-ROOT, B-DOCS-EXEC-GS, B-BUILD-CONTEXT, B-DOCS-EXEC-ROOT |
| migrations | B-MIGRATE-CLI, B-MIGRATE-CATALOG, B-MIGRATE-ENV, B-MIGRATE-IT (10/10, none skipped) |
| variants | B-OUTAGE, B-PERSIST, B-SOURCE-HASH, B-PORTS-OVERRIDE, B-RENAME, B-DBRENAME, B-TOGGLE-OFF-ADDITIONS (variant `850b7f2aa96f0986`), B-TOGGLE-OFF |
| final | B-DOCKER-ONLY, B-SECRET-CANARY (48 generated secrets absent), B-SCOPE (92 paths, exact allow-list) |
| INFO | PREFLIGHT (Docker 27.5.1, Compose 2.32.4), PROBE (`devtcp`), IMAGES |

### 6.1.1 Run history (every run, including failures)

| Run | Suites | Outcome | Notes |
|---|---|---|---|
| v1 (session 1) | static | FAIL: B-HADOLINT, B-MCP-SOURCE | Cycle 2 |
| v2 (session 1) | static, provenance | PASS | Cycle 2 retest |
| v3 (session 1) | migrations | PASS | 10/10 integration; catalog after 7 steps |
| `0b3c0b` (session 1) | root | FAIL: B-SECRET-CANARY only | Cycle 3 |
| v5 | root, variants | PASS (32 secrets absent) | Cycle 3 retest; first variants run |
| `4c9478` | all | 39 PASS, 1 FAIL (B-BOUNDARIES-ROOT) | Cycle 4 retest; regression found, giving cycle 5 |
| `d63754` | all | 39 PASS, 1 FAIL (B-RENDER-PLAN: “plan wrote to the repository”) | **Invalidated by an operator error, not an implementation fault.** I edited this Result file during the run. Its mtime, 18:38:03.25, falls inside the plan snapshot window (18:38:01–18:38:03). The plan output itself matched the register: 3 conflicts, 66 unchanged plus the manifest, and 3 proposal creates. The check correctly detected the write. No code changed, and the run was repeated with no repository edits (not a cycle). |
| `21348d` | all | **40 PASS, 0 FAIL, 0 UNMET** | Final acceptance run |

### 6.2 Tests added

- **`core/internal/mcp`:** `TestInitializeAndEmptyCatalog` checks the initialize handshake over SDK in-memory transports: server name `core`, the configured title, the tools capability, 0 tools, ping, and a clean close. `TestServersAreIndependent` checks that there is no global state.
- **`core/internal/migrate` (unit, no database):**
  - `TestLoadConfigValid` and `TestLoadConfigReportsEveryProblemWithoutValues`: configuration errors are named and secrets are never echoed.
  - `TestRunRefusesBeforeConnecting`: usage errors, and the production refusal of `down`/`reapply`, both exit 2.
  - `TestNormalSetIsBaselineOnly`.
  - `TestProtectionTargetsExactObjects`: no name or prefix matching, and the four quoted roles.
  - `TestCreateVersionTableProtectsOnTheSameTransaction`: revokes and the postcondition run on goose's handle, and a failure is propagated (recording fake).
  - `TestStoreDelegatesEverythingElse`.
  - `TestProviderAcceptsProtectedStore`: `WithIsolateDDL` is not set.
- **`core/migrations`:** `TestEveryMigrationIsVersionedAndReversible` and `TestBaselineCreatesNothing`.
- **`core/internal/migrate` integration (`-tags integration`, real PostgreSQL, 10 subtests):**
  1. fresh `version` creates protected metadata;
  2. fresh `status` before the baseline;
  3. an unprotected creation is never committed;
  4. concurrent first use;
  5. `up` applies the baseline only;
  6. synthetic DDL runs as owner, and runtime has no authority (`42501` for CREATE in `app`/`public`, SELECT on probe and metadata, and SET ROLE to owner, migrator and `postgres`; no superuser, BYPASSRLS, membership or owned relations; the migrator without SET ROLE is denied);
  7. synthetic down restores the baseline state;
  8. reapply;
  9. down to zero keeps only metadata, and up restores it;
  10. preflight refuses without the owner role.

  Assertion P (Plan §8.5) is checked after every step.
- **Project harness:** `infra/validation/run.sh` covers 40 PASS/FAIL check IDs across 5 suites plus final checks, listed in §6.1.

### 6.3 Real systems versus test doubles

- **Real:** PostgreSQL with the rendered role init, Redis, source-built MinIO, Compose networks and published ports, the `migrate` image and goose, the core, web and intelligence images, and host and in-namespace TCP/HTTP probes.
- **Doubles:**
  - an in-memory MCP client (test only);
  - the synthetic DDL migration (test only, never embedded);
  - a stubbed `ns_tcp` for the single-endpoint boundary controls;
  - a recording `DBTxConn` fake for protected-store unit tests.
- **No live AI provider:** intelligence answers `capability_not_configured`.

## 7. Cleanup and Docker inventory

- **Per run:** each harness run takes a start inventory and removes only the Compose projects it registered (`down -v`, plus images built under those project names and absent at start). It also removes the root `.env` it generated, and only when the file's hash matches.
- **Labels:** probe and build containers carry `g001.run=<run-id>`, and Compose resources carry their unique project label.
- **Results:** every harness run's `inventory-leftover.txt` was empty. At S8, the execution-owned `g001-gomod-ec0eadb2` volume was removed after its `g001.exec=ec0eadb2` label was verified. The disposable diagnosis project `g001-diag-ec0eadb2` and its copy were removed in cycle 5. The final volume list is identical to the S0 list. No container, volume, network or image whose name contains `g001`, the root project or `renamed` remains, and there is no root `.env`.
- **Kept and listed, not removed:**
  - base images pulled for builds: `golang:1.27.1-alpine`, `alpine:3.22`, `gcr.io/distroless/static-debian12:nonroot`, `node:24-slim`, `python:3.13.16-slim`, `ghcr.io/astral-sh/uv:0.12.23`, `pgvector/pgvector:0.8.7-pg17`, `redis:8.10.2-alpine`, `busybox:1.37.0`, `mikefarah/yq:4.48.1`, `koalaman/shellcheck:v0.11.0`, `hadolint/hadolint:v2.15.1`, `zricethezav/gitleaks:v8.30.1`;
  - the shared BuildKit cache.
- **Never touched:** pre-existing resources, including `project-archetype-render-cache` and the other projects' containers and volumes.

## 8. Acceptance criteria

| AC | Result | Evidence |
|---|---|---|
| AC1 bootstrapped from the Skill, content preserved, additions attributable | Met | §2: S0 inventory, inputs, 69 create equal to plan, manifest. B-RENDER-ID, B-RENDER-PLAN (zero writes; conflicts exactly the 3 registered), B-RENDER-REPEAT (exit 3; only the 3 proposals, each equal to the manifest), B-PROTECTED (180 files plus the Plan hash), B-SKILL-UNCHANGED. Customization register and BOOTSTRAP.md attribute every addition. |
| AC2 container-only configure, start, check, stop and migrate | Met | B-CONFIGURE, B-DOCS-EXEC-ROOT (README configure, start, migrate, check, stop and reset blocks verbatim), B-DOCS-EXEC-GS, B-MIGRATE-CLI (MIGRATIONS.md blocks verbatim), B-CHECKS-ROOT (core gofmt, vet and unit tests; web lint, types and tests; intelligence ruff and pytest; gitleaks), B-SHELL, B-HADOLINT, B-DOCKER-ONLY (host Go, Python, Node, psql, redis-cli, mc, minio and aws shadowed with `/mnt/*` stripped; 0 invocations) |
| AC3 identity, domain, database and ports configurable | Met | B-DEFAULTS (the actual root at 3100/8180/8100/5532/6479/9100/9101). On a disposable copy of the implementation (B-SOURCE-HASH equals the root): B-RENAME (`.env` only, `up --no-build`, 7 image IDs unchanged, web and core show the new identity), B-DBRENAME (`renamed_db_{owner,migrator,runtime,retention}`, metadata owned by `renamed_db_owner`), B-PORTS-OVERRIDE |
| AC4 health, isolation, least privilege, persistence | Met | B-OUTAGE, B-PERSIST, B-PORTS-ROOT and B-PORTS-OVERRIDE (published as `127.0.0.1:<port>`; host protocol probes; refused on non-loopback; see residual R-3), B-BOUNDARIES-ROOT (each datastore by name and every IP on its own port, from web and intelligence; core positive controls), B-BOUNDARY-CONTROLS (postgres-only, redis-only and minio-only detected, real and stubbed), B-COMPOSE (secret scoping) |
| AC5 migrations under intended roles, no domain schema, no runtime authority | Met | B-MIGRATE-IT (10/10, none skipped), B-MIGRATE-CLI (including production refusal, exit 2), B-MIGRATE-CATALOG (P plus role facts after each of 7 steps, starting before the baseline), B-MIGRATE-ENV, B-MIGRATE-SET, unit tests |
| AC6 MCP functional and dormant | Met | `internal/mcp` tests in core-check; B-MCP-SOURCE (`go list -deps ./cmd/server` excludes the SDK, `internal/mcp` and migration code; no fixture or isolation dependency; no transport symbols; routes only `/healthz` and `/readyz`); B-MCP-HTTP (anonymous JSON-RPC `initialize` and GETs on `/mcp`, `/mcp/`, `/sse`, `/messages`, `/message` and `/rpc` return 404 from core's namespace and the host port; web returns 404) |
| AC7 optional intelligence and secret boundaries | Met | Enabled root: intelligence-check, and execute probes (401, 422, then a typed stream with `capability_not_configured`) with no gateway key. Disabled variant: B-TOGGLE-OFF-ADDITIONS, B-TOGGLE-OFF (no intelligence service, Python-based container or image, AI settings or files; migrate up, status and version; core ready without intelligence). B-ENV-EXAMPLE, B-COMPOSE, B-BUILD-CONTEXT, B-DOCKERIGNORE, B-SECRET-CANARY |
| AC8 bounded diff, docs, Result, budget | Submitted | B-SCOPE: 92 exact-path allow-list. Full `git status` inspected (§9.1). This Result maps AC1–AC8, separates inherited, new and deferred work, and reports B5 usage with the cycle-3 checkpoint (§10). Codex ACCEPT is pending. |

## 9. Deviations, residuals and review findings

### 9.1 Diff inspection

`git status --porcelain --untracked-files=all` lists 92 untracked paths and no modified, deleted or staged tracked paths. All 92 match the exact allow-list. `git diff 9b4fba4 --stat` is empty for tracked files. No Codex-owned file changed during execution, so no `--exception` was used.

### 9.2 Deviations from the Plan (none architectural)

1. **B-MIGRATE-CLI step order.** The order is version, status, up, reapply, down, status+version, up. The Plan lists version, status, up, down, status, reapply, up. The same commands and assertions are covered, and P is checked after every step.
2. **Harness additions within Plan scope.**
   - `--exception PATH`: Plan §11.3 and R12 require exact-path exceptions to be reported. To pass B-PROTECTED, an exception also needs `--expect-hash`.
   - B-CONFIGURE and B-DOCS-EXEC-GS split out evidence the Plan assigned to B-DOCS-EXEC-ROOT.
3. **Compose-managed containers** carry their Compose project label, not `g001.run` (Plan §11.3: “where Docker supports labels”).
4. **The cycle-4 corrections (§10)** tightened several assertions and replaced the prefix allow-list with an exact one. These strengthen validation, and no check was removed or weakened.
5. **S0 evidence files partly lost** (§2.1).

### 9.3 Read-only review (SA-R)

The reviewer read the full untracked diff and the goose and pgx module source. It reported no high-severity findings, used 0 cycles and modified no files.

- **Fixed in cycle 4:**
  1. (medium) prefix allow-list → exact paths;
  2. B-PERSIST Redis half could pass vacuously;
  3. B-TOGGLE-OFF Python detection could not fail for built images. It now inspects build sources and container image environments, with a positive control;
  4. empty container IDs in negative probes;
  5. fresh-status step had an unchecked `down -v`;
  6. B-MIGRATE-CATALOG step count;
  7. service token visible in probe argv. It is now built by the shell's `printf` builtin and sent with `nc`;
  8. pgx fallbacks and TLS derived by `ParseConfig("")` are now cleared explicitly;
  9. MIGRATIONS.md exit-code wording;
  10. MIGRATIONS.md validation command;
  11. run-ID labels;
  12. disabled variant also runs `migrations:inspect`;
  13. configure helper now also removes its temp file on INT, TERM and HUP;
  14. B-DEPS now flags untagged non-stage `FROM`;
  15. lib.sh host-requirement comment;
  16. README database-rename wording.
- **Accepted residuals (not fixed):**
  - **R-1:** a SIGKILL during `configure-env.sh` can leave `.env.tmp.*`. It is mode 600 and not ignored, so B-SCOPE would flag it. Adding an ignore line would customize the rendered `.gitignore`.
  - **R-2:** the migration preflight checks the session user's attributes, not the owner role's. The template creates the owner NOSUPERUSER NOBYPASSRLS. This is relevant to production provisioning (STAGE-RLS).
  - **R-3:** under WSL, refusal on the non-loopback address is weaker evidence than `compose port == 127.0.0.1:<port>`. Both are asserted.
  - **R-4:** the rendered `data` network is not `internal`, so `migrate` and integration-test containers have egress. This is inherited and unchanged.
- **Inherited from G000:**
  - tag-only pins for `alpine:3.22` and distroless;
  - archived upstream MinIO (product-owner ADR pending);
  - hadolint DL3066 at info level on the rendered `core.Dockerfile` (below the archetype's warning threshold).

## 10. Execution budget

**Budget class:** B5 · **Maximum cycles:** 5 · **Cycles consumed:** 5, the full budget (main agent 5; SA-R 0; no implementation subagents)

### 10.1 Cycle log

1. **Unit-test guards too broad.** The first core unit-test run failed `TestProtectionTargetsExactObjects`: its forbidden pattern `_seq` matched the catalog function `pg_get_serial_sequence`. The baseline content test also matched GRANT/REVOKE inside SQL comments. The guards were narrowed (`_id_seq`; comments and literals stripped before matching), with no weakening, and all core packages passed on retest.
2. **Static harness false failures.** B-HADOLINT ran hadolint without the archetype's `--failure-threshold warning`, so the info-level DL3066 failed it; the rendered `core.Dockerfile` has the same note. B-MCP-SOURCE matched `WithIsolateDDL` in Go comments. Both checks were corrected: the archetype threshold, and code lines only (with a negative control). Static and provenance then passed.
3. **B-SECRET-CANARY flagged its own planted canary.** The root suite passed every other check. B-SECRET-CANARY found the deliberately planted build-context canary (`$WORK/ctx/core/.env.local`). The disposable copy is now removed after B-BUILD-CONTEXT, so the scan covers every log, report and the repository. Retest: root and variants all PASS (32 secrets absent).
4. **SA-R findings.** The corrective batch in §9.3 was applied as one implementation attempt. Its retest, the first final `all` run (run `4c9478`), passed 39 checks and failed 1: B-BOUNDARIES-ROOT, with an empty response to the bad-envelope execute probe. The cycle-4 rewrite of the token-bearing probes (finding 7, `printf | nc`) had broken a previously passing check.
5. **nc half-close regression.** Diagnosed on a disposable intelligence-only project (`g001-diag-ec0eadb2`, removed afterwards): busybox `nc` half-closes the socket on stdin EOF, and uvicorn drops the response for a client that has gone away. The earlier busybox-httpd trial did not show this. With stdin held open briefly (`{ printf …; sleep 10; } | nc`), the server returned `422` and the full typed `200` stream (`error capability_not_configured`, then `done`). That probe change was the fix, and it was retested by the final `all` run (§6.1). The token still never appears in argv.

**Not counted:** the execution-owned Go module cache volume was created root-owned, and `go mod tidy` could not write to it. This was an environment setup step before any validation of the implementation (BUDGET_POLICY §4). The volume was handed to the build user and the step rerun, with no implementation change. If Codex counts it conservatively, the total would be 6, one over B5. I flag this for Codex's budget review; I have not extended the budget.

**Thrashing assessment (BUDGET_POLICY §8).** Cycle 4 introduced one regression of a previously satisfied check. It had a single, reproduced root cause, and it was fixed without reverting the security improvement it came from. No failure recurred across cycles, and no fix was reverted.

**Interruptions (no cycles):** an API connection loss and a WSL restart ended the first session after cycle 3's failing run. Work resumed from the transcript with no state change (§2.1).

### 10.2 Cycle-3 checkpoint (BUDGET_POLICY §7)

Recorded after cycle 3, before cycle 4:

- **Criteria completed:**
  - AC1 (static and provenance green);
  - AC2, AC4 and AC6 on the actual root (root suite all PASS);
  - AC5 (migrations suite green: 10/10 integration, catalog after 7 steps);
  - AC3 and AC7: variants all PASS on the retest (rename, database rename, port override, outage, persistence, disabled variant).
- **Criteria remaining:** AC8, which needs the SA-R review triaged, a final `all` run from clean state, and this Result.
- **Suite status:** static, provenance, migrations, root and variants are all PASS.
- **Repeated failures:** none. Each failure was distinct, and all three were in test or harness code, not the product implementation.
- **Uncertainty:** low. The only remaining risk was review findings.
- **Convergence:** yes. No fix reverted or re-broke an earlier check, and scope is unchanged.
- **Forecast:** 4 of 5 (one corrective batch for review findings), with 1 in reserve. No budget change requested. *Outcome:* the reserve cycle was consumed by the cycle-4 regression (cycle 5).

## 11. Unresolved issues and blockers

None blocking. The residuals R-1 to R-4 and the inherited items in §9.3 are recorded for Codex.

## 12. Suggested follow-up (not authorized scope)

- **G003 (CI):** wire `infra/validation/run.sh … static provenance migrations` into CI.
- **G004:** mount MCP only with authentication and session-scoped servers. Revisit R-2 when production database provisioning is planned.
- **Product owner:** the MinIO replacement ADR remains open (G000 residual).

## 13. Actual-repository runtime determination (2026-10-04, pre-submission check)

**Question.** Was the actual repository exercised as a real Docker Compose project? Or only the generator and disposable copies of it?

### 13.1 Determination

**The approved artifacts require it, and it is not deferred.**

| Source | Requirement |
|---|---|
| Spec R6 | “Commands must work as written from the actual repository.” |
| Spec AC2 | “Execute documented commands verbatim against the root implementation using disposable local credentials.” |
| Spec AC3 | “Actual root starts at assigned defaults.” |
| Spec AC7 | “Enabled root passes existing intelligence checks.” |
| Plan §11.3, `root` suite | Executes the README and GETTING-STARTED blocks in the actual root (project from `.env.example`, default ports) |

The Spec assigns the destructive tests to disposable projects: rename, rollback, outage and persistence (Spec §Required validation, “Use uniquely named disposable Compose projects/volumes…”; AC5 “Disposable status/up/down/reapply tests”). They therefore stay on copies proven byte-identical to the implementation (B-SOURCE-HASH). No future goal owns actual-root runnability; G003 (CI) may later automate these checks.

**Already demonstrated before this check.** The final run `21348d` ran the actual repository root as Compose project `beginify` on the default ports. It covered B-CONFIGURE, B-DOCS-EXEC-ROOT (README configure, start `up -d --build --wait`, migrate status/up/status on the real database, check, stop and reset), B-DEFAULTS, B-PORTS-ROOT, B-BOUNDARIES-ROOT, B-BOUNDARY-CONTROLS, B-MCP-HTTP, B-CHECKS-ROOT, and B-DOCS-EXEC-GS (GETTING-STARTED start, check and stop).

**Evidence gaps on the actual root** (shown only indirectly, or only on disposable copies):
- a recorded `docker compose ps` health state;
- web `/api/ready` 200;
- authenticated PostgreSQL, Redis and MinIO data-plane round-trips (previously only core readiness and unauthenticated protocol probes);
- metadata catalog assertion P on the real database;
- the resource state after `stop` (containers and networks gone, durable volumes kept);
- migration and object persistence across a warm restart, and Redis ephemerality;
- a cold re-bootstrap from empty volumes.

These are closed by the pass below. No implementation, Goal, Spec or Plan change was needed.

### 13.2 Supplementary actual-root evidence pass

- **Tool:** a scratchpad script, `<scratchpad>/root-evidence.sh`, outside the repository and not a deliverable. It sources the project harness helpers in `infra/validation/lib.sh` for probes, shadowing and inventory.
- **Commands:** every stack command is a README or MIGRATIONS.md block, run verbatim from the repository root with `docker compose --env-file .env -f infra/compose/local.yml …`. The ports are the architecture's standard+100 defaults.
- **Run:** `08b057`, 2026-10-04 19:37:10–19:39:09 local, at commit `9b4fba4` with the G001 working tree. Result: **exit 0; 34 PASS, 0 FAIL, 0 UNMET.**
- **Isolation:** the implementation tree's hash was identical before and after. Afterwards, `git status --ignored` showed nothing but the 92 untracked G001 paths.

| Phase | Checks (all PASS) | Evidence |
|---|---|---|
| 0 Preconditions | PRECONDITIONS | No root `.env`; no `beginify` containers, networks or volumes; ports 3100, 8180, 8100, 5532, 6479, 9100 and 9101 free. Nothing is touched otherwise. |
| 1 Configure | CONFIGURE | README configure block: `.env` mode 600, project `beginify`, `BIND_ADDRESS=127.0.0.1`, default ports unchanged |
| 2 Resolve configuration | CONFIG | `config -q` valid. Services `core intelligence minio minio-init postgres redis web`. Published `127.0.0.1:3100->3000`, `:8180->8180`, `:8100->8000`, `:5532->5432`, `:6479->6379`, `:9100->9000`, `:9101->9001`. Networks `beginify_{data,edge,inference}`; volumes `beginify_{postgres,minio}-data`. Read with `yq`, without printing environment values. |
| 3 Build and start | START-1 | README start block (`up -d --build --wait`) in 20 s |
| 4 Health and connectivity | PS-1, HEALTH-1, PORTS, POSTGRES, REDIS, MINIO, INTELLIGENCE | See the `compose ps`, `/readyz` and datastore details below |
| 5 Migration workflow (real database) | MIGRATE-1, MIGRATE-INSPECT, CATALOG-1 | README migrate block: `migration 00001_baseline.sql pending` / `current_version 0` → `applied 00001_baseline.sql` / `current_version 1` → `migration 00001_baseline.sql applied 2026-10-04T23:37:43Z` / `current_version 1`. The MIGRATIONS.md inspect block agrees. Catalog: `objects=2 owners=beginify_owner extra_acl=0 table_privs=0 seq_privs=0 beyond_metadata=0 runtime_owned=0`. |
| 6 Stop | STOP-1 | README stop block: no `beginify` containers or networks remain. Only `beginify_minio-data` and `beginify_postgres-data` remain; Redis has no volume. |
| 7 Warm restart | START-2, PS-2, HEALTH-2, PERSIST-MIGRATION, CATALOG-2, PERSIST-MINIO, EPHEMERAL-REDIS | Start again in 18 s, all healthy. 00001 still applied at version 1, P holds, the MinIO object survived, and the Redis key is gone. |
| 8 Stop and reset | STOP-2, RESET-1 | No containers or networks after stop. After the README reset block (`down -v`), no `beginify` containers, networks or volumes remain. |
| 9 Cold re-bootstrap | START-3, PS-3, HEALTH-3, MIGRATE-3, CATALOG-3, FRESH-MINIO | Start from empty volumes in 18 s, all healthy. Migration again pending → applied → version 1, and P holds. The bucket was recreated empty. |
| 10 Final stop, reset, cleanup | RESET-2, CLEANUP, SECRET-CANARY, DOCKER-ONLY | No project resources remain. The images built by the pass were removed, as was the generated `.env` (hash-verified). The end inventory equals the start inventory. 8 generated secrets are absent from all evidence and repository files. No host runtime was invoked (shadowed). |

**Phase 4 details:**
- **`docker compose ps`** (identical in all three starts):

  | Service | State | Health | Exit |
  |---|---|---|---|
  | core, intelligence, minio, postgres, redis, web | running | healthy | 0 |
  | minio-init | exited | — | 0 |

- **Core `/readyz` (127.0.0.1:8180):** `200 {"status":"ready","product_name":"<configured>","checks":{"object_store":"ok","postgres":"ok","redis":"ok"},"optional":{"intelligence":"ok"}}`. Core `/healthz` returns 200.
- **Web (127.0.0.1:3100):** `/api/health` returns 200, and `/api/ready` returns `200 {"status":"ready","checks":{"core":"ok"}}`.
- **Intelligence (127.0.0.1:8100):** `/healthz` returns 200, and core reaches `intelligence:8000`.
- **PostgreSQL:** the host SSLRequest on 127.0.0.1:5532 answers `N`, and `pg_isready` reports “accepting connections”. The database exists with the roles `<db>_{migrator,owner,retention,runtime}`. The runtime role logs in with its password over TCP (`<db>_runtime 5432`).
- **Redis:** an unauthenticated host `PING` on 127.0.0.1:6479 gets `-NOAUTH`. An authenticated `PING` returns `PONG`, and a SET/GET round-trip succeeds.
- **MinIO:** live 200 on 127.0.0.1:9100, and the console answers on 9101. An object was written, listed and read back in the application bucket with the scoped service credential, through the `minio-init` image.
- **Core reachability:** core reaches `postgres:5432`, `redis:6379` and `minio:9000` from its own namespace.

### 13.3 Limitations and accounting

- **Build cache.** The pass's `--build` steps were served from the shared BuildKit cache (31 cached steps), so the starts took 18–20 s. Uncached first builds of the same Dockerfiles ran earlier, in session 1's root run (`0b3c0b`), with MinIO compiled from source. That run's logs were lost with `/tmp` (§2.1).
- **Not committed.** The evidence script is not part of the deliverable. The committed, repeatable actual-root coverage remains the harness `root` suite (§6.1).
- **Credentials.** A superuser query over the PostgreSQL container's local socket was used for validation only, as in the harness. Credentials reached containers only through `.env` and container environments, never through command-line values.
- **Budget.** This is validation only: no implementation or test file changed. It consumes **no** B5 cycle, and the cycle total stays 5 (§10).
- **Inventory.** The pass created and removed only the root `.env` and `beginify` Compose resources, which were absent beforehand. The Docker inventory after the pass equals the inventory before it.

Status is unchanged: **submitted for Codex review. Not accepted.**

