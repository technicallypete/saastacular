> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G009 — Tenant-aware object storage boundary: Result

## Status

**Integrated; hosted staging qualification PASS (retry); awaiting Codex final ACCEPT.** G009 is not finally accepted.

- G009 merged as MR !24 (`e087f2c`). Its post-merge gate failed on B-WEB-ERROR-OUTPUT WE4.
- The WE4 correction merged as MR !25 (`944ecf3`). Its first gate failed on B-DOCS-EXEC-ROOT (start block; cause
  unknown).
- The Codex-authorized single exact-job retry of that gate passed.
- B5: 3 of 5 attempts used, no extension.
- The hosted record is at the end of this Result. The earlier status and evidence of attempt 1 and correction passes 1
  and 2 follow unchanged.

- **Final local validation (b4): b18 + b19.**
  - **b18** (`run.sh --base f629847 … all unit`, run 481155, candidate `9487675`) ran every suite (lint lint-base lint-web
    lint-core controls ci-controls provenance contracts root migrations ops rename off unit). **95 PASS, 2 FAIL:**
    - B-PROTECTED failed because this Result (in base `f629847` since the prep import) had been updated.
    - B-TOGGLE-OFF failed because the intelligence-disabled variant lacked `core/internal/objectstore`
      (`APPLICABLE` gap), so its core build failed.
    - Everything else passed, including all five B-OBJ-* checks, B-MIGRATE-CATALOG, B-IDENTITY-CONFIG, B-PERSIST,
      B-DBRENAME, B-CI-CONTROLS (background ci-controls exit 0), B-UNIT-CHECKS, B-SECRET-CANARY and B-SCOPE.
    - B-REQUIRED-CHECKS failed only for those two IDs.
  - **Fixes (checkpoint `6466f6f`):**
    - `APPLICABLE` lists the G009 additions (`core/internal/objectstore`, `objects_config_test.go`, `OBJECT-STORAGE.md`).
    - This Result was returned to its base bytes for validation and rewritten afterwards, as G008 did. No
      `--expect-hash` exemption was used.
  - **b19** (`lint lint-base provenance off`, run 371ed5, candidate `6466f6f`): **PASS**. 24 PASS, including:
    - B-PROTECTED: 505 protected files equal base;
    - B-TOGGLE-OFF-ADDITIONS, B-TOGGLE-OFF (migrate to 7 in the variant), B-OBS-OFF, B-BROWSER-OFF (141 tests);
    - B-SKILL-UNCHANGED, B-SECRET-CANARY, B-SCOPE (49 paths) and B-REQUIRED-CHECKS (23/23).
  - **Coverage:** every required ID of the full profile passed in b18 or b19 on the same `f629847` target. The
    `9487675` → `6466f6f` delta is the `APPLICABLE` lines (consumed only by `off`) and the Result bytes (consumed by
    provenance and B-SCOPE). b19 reran both, so b18's other evidence carries over unchanged.
  - INFO: BUILD-LEDGER `expected 76, rows 75, unknown 1 … ac5_status UNMET` is an informational G051 ledger line, not a
    required check. It is recorded as observed.
- This Result is written after b19. Its bytes therefore differ from base (allow-listed, B-SCOPE), and only B-PROTECTED
  would see that difference.

## Execution binding (Git-first)

| Item | Value |
|---|---|
| Base staging | `bd31633e865832b0ec6b780a01794564d15436c4` (origin/staging at fetch, 2026-10-14) |
| Accepted Plan | [G009-tenant-object-storage-plan.md](../plans/G009-tenant-object-storage-plan.md), r1, accepted commit `1c9bae34c468c11376de49ebc3fb44b110eed107`, SHA-256 `bb3496be5b175ef5693d6fd55083202b3cff8a17e59fbf13486ff9b241f04b53` (staging bytes equal the `1c9bae3` blob) |
| r0 archive | G009-tenant-object-storage-plan-r0.md (not included), SHA-256 `e2843c781042a18fb4d8623620d3405e0982ac7020c41f2a9c8ba5dd26afc6f7` |
| Plan r1 ACCEPT | G009-plan-r1-codex-review.md (not included), SHA-256 `c73be7e1dc238b8c52a8dd0c46eb0870e614ac1b66003e0f02df69f585724dc4` |
| B-SCOPE allow-list | G009-scope-allowlist.txt (not included), SHA-256 `e1cb865d3ccb189ae76f7dc2f6b4518425344230bfb8d03b455e3c4e2bda0578`, 56 lines (28 patterns) |

**Approved inputs, bound by commit `bd31633` and path** (no snapshot or per-file inventory; AGENTS.md Git-first
provenance):
- Goal r1 `management/v0/goals/G009-tenant-object-storage.md`; Spec r1 `management/v0/specs/G009-tenant-object-storage-spec.md`.
- `management/v0/results/G009-storage-write-policy-r1.md`; `management/v0/results/G009-planning-start-authorization.md`.
- `management/v0/results/G008-implementation-codex-review.md` (G008 FINAL ACCEPT, staging `80ebe22`).
- `AGENTS.md`, `CLAUDE.md`, `management/BUDGET_POLICY.md`, `docs/ARCHITECTURE.md`, `docs/foundation/*`,
  `skills/saas-foundation/references/{storage-audit,database-rls}.md`.
- Implementation sources and harness: the tree at `bd31633`.

**Allow-list derivation.** Each pattern maps to a Plan §2 create/modify row or a §5.6 RC row, including the §5.6
version-7 reconciliation tests, the approved B-IDENTITY-CONFIG production-probe https data endpoint (in
`infra/validation/run.sh`), the scoped `tx.go` helper change E11 with its `facade`/`context`/`commit_test.go` call
sites, and the conditional E9 `core/go.mod` classification. The §2 category "version-6 pins in G005–G008 Go tests"
is resolved to the files named in §5.6 plus the §5.6 grep inventory (`00006|version 6|current_version|nine|26
policies`) of latest-version test files at `bd31633`. Excluded on purpose: `rls_expected_test.go` (G008 oracle,
unchanged), `core/go.sum`, production identity/migrate code, `manual_qualification_test.go` (its RC-12 guard is
unaffected) and the version-subset migrate tests (`policy_`, `tenancy_`, `business_profile_integration_test.go`).
Any further path needs a Codex-reviewed allow-list amendment. A read-only `grep -E -f` self-check matched every
planned path and none of the excluded paths (not a harness run).

**Harness binding.** Plan §6 uses `infra/validation/run.sh --base <target> --allow-file
management/v0/results/G009-scope-allowlist.txt --report-dir <EVID>/<batch> <suites>` directly. `run.sh` binds by
commit (`--base`) and the committed allow-list path. B-SCOPE strips comment and blank lines and matches every
changed and untracked path. `infra/validation/run-approved.sh` (snapshot, path-re, count) is **not used**, so no
binding directory, snapshot, path-regex or count outside Git is needed. Batch evidence goes under
`<home>/.local/state/platform-evidence/g009-impl-bd31633/<batch>`.


### Binding history (later authority, Git-first)

| Record | Commit / SHA-256 | Effect |
|---|---|---|
| Binding confirmation | staging `c2df228`; `G009-execution-binding-codex-confirmation.md` `c6fe3782…a42961d9` | released implementation; prep commit `e125af1` (allow-list + initial Result) imported to staging byte-identically, then dropped from this branch by the rebase as already upstream (delivery history) |
| Amendment 01 | staging `24c8bda`; `G009-execution-amendment-01.md` `3a692989…65916d2c` | allow-list `e1cb865d…` → `98311013…` (32 patterns): `observability/exit.go` (5 config keys), `logrecord_test.go` (catalogue pin), `validate-controls.sh` (VA7/VU1/J8 counts), `OPERATIONS.md` (counts) |
| Amendment 02 | staging `6b7ede8`; `G009-execution-amendment-02.md` `b76a07f8…6e62a9dfe1` | allow-list → `154c188a…f8dadc98` (33 patterns): `observability/exit_test.go`, one fixture line (production base data endpoint https) |
| Oracle RU05 ruling | staging `2485ede`; `G009-oracle-ru05-codex-ruling.md` `a466be31…6e7502` | 428C9 accepted only for RU05's non-DEFAULT generated storage-key assignment; oracle r2 |
| Query-cost ruling | staging `f629847`; `G009-query-cost-codex-ruling.md` `e1075bb0…09ac26` | the INSERT … RETURNING 4/2 reader observation accepted as characterized; RETURNING and the key comparison before upload kept; PG-003 open |

**Oracle provenance** (delegate-written, imported byte-exact into `core/internal/tenancy/access/objects_expected_test.go`):
r0 `123f3866…ffbaa8` (failed gofmt only, never imported); r1 `d9f1061c…f530a0e4` (mechanical gofmt of r0, no
non-whitespace change); **r2 `e1d26df1…7f9ab96f`** (current: adds `oeRLSImmutableGenerated`, used by RU05 only, per
the ruling). The glue accepts 428C9 only for that outcome and the `object_key` column, with the unchanged-row check.

## Ownership

| Item | Value |
|---|---|
| Task / role | G009 / impl (implementation and integration owner, single stack owner; Plan §7) |
| Agent id | `a37b9d28086da7aaf` |
| Path | `<source-checkout>` (native name under the G009 task-only exception) |
| Repository | `git@gitlab.example:<namespace>/<project>.git` |
| Branch | `g009-tenant-object-storage` from `origin/staging` `bd31633` |
| Worktree lock | `locked claude agent agent-<id> (pid 68375 start 10178190)` |
| Effort | high as instructed; actual runtime setting unknown (not verifiable from inside the worker) |
| Dispatch record | `<home>/.local/state/platform-evidence/g009-impl-bd31633/dispatch-record-impl.md` (outside Git) |
| Coordinator | `0-agent-5c6f436` (dispatch intent in the same directory) |

The oracle delegate was dispatched separately and blind by the coordinator (output folder
`<home>/.local/state/platform-evidence/g009-oracle-c2df228/`; this worker never read its inputs). The read-only
security reviewer is requested at the S-review hand-back.

## G008 exact-delta check (Plan §0, §8 step 1)

- G008 merge `81b186c`, correction `d0e8d40` and FINAL ACCEPT `80ebe22` are all ancestors of `bd31633`.
- `git diff --name-only 81b186c bd31633` changes only `management/**` (13), `skills/**` (4),
  `.agents/skills/repository-supervisor/SKILL.md`, `AGENTS.md`, `docs/PERFORMANCE.md` and
  `docs/agents/GITLAB-BRANCH-WORKFLOW.md`.
- None of the §0 G008 items changed: `core/migrations/00006_row_level_security.sql`, `core/internal/tenancy/access/tx.go`
  (`inContextTx`, discard lifecycle), identity `LockLiveSession`/`CheckLiveSession`, and the `infra/validation/run.sh`
  RLS pins. Nothing under `core/`, `infra/`, `web/`, `intelligence/`, `contracts/`, `docs/foundation/` or
  `.gitlab-ci.yml` changed.
- Since the Plan r1 base `c944ef7`: only management files (PROGRESS, the Goal/Spec status lines, the imported Plan and
  archive, the G008 final review and the r1 review). The Goal/Spec diffs are status text only.
- The `AGENTS.md`/workflow commits since `81b186c` (`74fc694`, `72c4d30`, `c944ef7`) are governance guidance already
  reflected in the accepted Plan's authority (feature-MR auto-merge default, epic-closeout compaction).
- **Result: unchanged.** The integrated G008 state the Plan relies on is byte-identical at `bd31633`; no
  re-reconciliation of any §0 item is needed.

## Attempt ledger (B5)

| Attempt | State |
|---|---|
| Budget | B5: five shared Goal-level execution attempts, checkpoint after attempt 3 (Plan §12, r1 ACCEPT). Planning spent none. |
| Binding prep | not an attempt (management records only) |
| 1 | **complete; Codex review: CORRECTION LOOP (C1)** (started 2026-10-14 08:03Z (preflight PASS: docker-desktop-mounts 7%, harness lock free, staging `c2df228`): implementation, focused batches b01–b12, covering b14 PASS. Two allow-list amendments, one oracle ruling, one query-cost ruling and the S-review corrections were absorbed within attempt 1 (no new attempt); b15–b17 focused reruns PASS; b4 = b18 (2 harness-reconciliation failures) + b19 rerun PASS.) |
| 2 | **correction pass 1, complete; Codex ACCEPT (pre-integration); merged as MR !24 (`e087f2c`); post-merge staging gate FAILED (WE4)** (Codex-authorized as attempt 2 of B5 in `G009-implementation-codex-review.md`, staging `6ed5363`; checkpoint after attempt 3 remains): C1 fix, regressions, race-detector run, covering batches b20–b27 |
| 3 | **post-integration correction pass 2, complete; Codex ACCEPT for integration; merged as MR !25 (`944ecf3`); hosted staging qualification PASS on the authorized retry (job 17173244141)** (Codex-authorized as attempt 3 of B5 in `G009-implementation-codex-review.md`, staging `6bcddc1`): WE4 diagnosis, run.sh WE4-block fix on `g009-we4-correction`, focused cases and the root-suite rerun. **Attempt-3 checkpoint due (Plan §12 / BUDGET_POLICY):** attempts used 3 of 5, 2 remain; the shared budget is not extended. |

## Implementation summary

A route-less, core-internal storage increment on the G007/G008 boundary ([OBJECT-STORAGE.md](../../../foundation/OBJECT-STORAGE.md)):

- **Migration `00007_stored_objects.sql`** (SHA-256 `28f0a073…0d12`): `app.stored_objects` exactly as Plan §3.1 (named
  constraints), the owner-owned SECURITY INVOKER guard trigger, FORCE RLS with the three policies (SCOPE, the upload
  check, W), the runtime grants SELECT / INSERT(6) / UPDATE(state, sha256); exact version-6 preconditions and
  version-7 postconditions (RLS tables, policy inventory, ACLs, function allowlist, schema privileges); down refuses
  with rows, else restores the exact version-6 inventories.
- **`core/internal/objectstore`**: `Store`, validated `Key`, `Binding`, closed errors; the S3 adapter (unsigned
  streaming payload, `WhenRequired` checksums, explicit length, no retries, bounded transport, offline presign with the
  fixed overrides) and `EndpointProblems`/`New` re-enforcing the §4.7 transport rules.
- **`core/internal/tenancy/objects.go`**: `ObjectID`, states, `ValidateContentType`, `ObjectStorageKey`, the six
  statements and transaction-scoped primitives; input codes in `errors.go`.
- **`core/internal/tenancy/access`**: `NewServiceWithObjects`, `ObjectConfig`, the five operations (Plan §4.4 sequences:
  P0/P1/T1/P2/V/T2/R/C and D1–D4), `decideObject` (the closed storage mapping) and the same-facts denial rule, the owned
  upload source, the exact reader with final-byte holdback, the verifying download reader, the redacting `SignedURL`,
  inert stage/crash seams, the `storage.operation` record; `storageTx` with the 5 s decision budget (the budgeted
  context is passed to every statement) and the 5 s cleanup budget through the scoped `tx.go` helper edit (E11:
  `cleanup` parameter; `cleanupUnbounded` keeps every G007/G008 call byte-for-byte in behaviour).
- **`core/internal/config`**: the five optional settings with defaults and bounds, and the endpoint transport rules
  (shape everywhere; https for both endpoints in production; mixed http-data/https-signing refused everywhere).
- **Harness**: version-7 pins (`MIGRATION_00006_SHA`, 00007 in the set, root/migrations/ops/rename/off/identity
  catalog), B-IDENTITY-CONFIG's https production data endpoint, the private MinIO start and stdin init in the
  migrations project, `obj_catalog`/`obj_migration`/`obj_it` (incl. OBJ-SIGN-HOST)/`obj_ordering`/`obj_faults`, five
  required IDs, the `storage.operation` event and five `core.exit` keys, one `profile-paths.tsv` row; amendment-01
  count reconciliations.
- **Docs**: `OBJECT-STORAGE.md` (new); MIGRATIONS, DATABASE-ISOLATION, AUTHORIZATION, FOLLOW-UP, OPERATIONS updated.

## Files changed (50 paths since staging `f629847` with this final Result, 49 at b19; all within the 33-pattern allow-list, B-SCOPE PASS; the S-review corrections touched only paths already in this set)

Created: `core/migrations/00007_stored_objects.sql`; `core/internal/objectstore/{objectstore,s3,objectstore_test}.go`;
`core/internal/tenancy/{objects,objects_test}.go`; `core/internal/tenancy/access/{objects,objects_io,objects_policy}.go`;
`core/internal/tenancy/access/objects_{,expected_,integration_,ordering_integration_,faults_integration_,footprint_integration_}test.go`;
`core/internal/migrate/objects_integration_test.go`; `core/internal/config/objects_config_test.go`;
`docs/foundation/OBJECT-STORAGE.md`.
Modified: `core/internal/config/config.go`; `core/internal/observability/{exit,exit_test}.go`;
`core/internal/tenancy/{errors,boundary_test,integration_test}.go`;
`core/internal/tenancy/access/{access,tx,facade,context,errors,log,boundary_test,commit_test,integration_test}.go`;
`core/internal/identity/store_test.go`; `core/migrations/migrations_test.go`;
`core/internal/migrate/{runner_test,integration_test,rls_integration_test}.go`; `infra/validation/{run.sh,required-checks.tsv}`;
`infra/validation/tools/internal/logrecord/{events.json,logrecord_test.go}`; `infra/ci/{profile-paths.tsv,validate-controls.sh}`;
`docs/foundation/{MIGRATIONS,DATABASE-ISOLATION,AUTHORIZATION,FOLLOW-UP,OPERATIONS}.md`; this Result.
Not modified: `core/go.mod`/`go.sum` (E9 not needed: the SDK middleware is passed as a value, no direct smithy-go
import), `cmd/**`, `health`, identity production code, 00001–00006, compose, `.env.example`, `infra/minio/**`,
`docs/foundation/VENDORS.md` (deviation 1).

## Version-7 reconciliation of accepted checks (Plan §5.6 RC list)

| Location | Change |
|---|---|
| `run.sh` B-MIGRATE-SET | 00007 in the exact set; `MIGRATION_00006_SHA` byte pin |
| `run.sh` root migrate block | 00007 pending/applied; `current_version 0/7/7` |
| `run.sh` B-MIGRATE-CLI/CATALOG | reapply/rollback act on 00007; versions 7/6; app relations 38 at version 7 (35 at 6) |
| `run.sh` B-PERSIST, B-DBRENAME, B-TOGGLE-OFF | version 7 (00007 applied) |
| `run.sh` B-IDENTITY-CATALOG | the 3 stored-object relations join the expected set at version 7 |
| `run.sh` B-IDENTITY-CONFIG | production base `OBJECT_STORE_ENDPOINT=https://minio.invalid:9000` (asserted); cc1–cc8 unchanged |
| `migrations_test.go` | set 00001–00007; 00006 SHA pin; new static `TestStoredObjectsMigrationShape` |
| `runner_test.go` | normal set and provider sources through 7 |
| `migrate/integration_test.go` | lifecycle to 7 (38 relations + guard), reapply 00007, down 7→6→…→0, `reset` drops the object table first |
| `migrate/rls_integration_test.go` | `TestRLSCatalog` migrates to 7 and steps down from 7 on the empty object table; the other G008 migration tests use a version-6 subset runner, keeping their exact version-6 expectations (deviation 3) |
| `tenancy/integration_test.go` | rolls back 7→6→5 |
| `access/integration_test.go` | `setupIT` expects version 7; hookDB gains `commitLostNoCommit` (faults) |
| `identity/store_test.go` | exact app table set + `stored_objects`, 10 forced tables |
| `access/commit_test.go` | `inTx(..., cleanupUnbounded, ...)`; assertions unchanged |
| `access/boundary_test.go` | files/imports/exported surface; objectstore and AWS SDK importer rules; 4 `inContextTx` call sites (+`storageTx`) |
| `tenancy/boundary_test.go` | 12 production files; pinned functions; `TestObjectSQLBoundary`; `constStrings` skips numeric constants |
| `exit_test.go` (amendment 02) | production fixture data endpoint https |
| `validate-controls.sh`, `OPERATIONS.md`, `logrecord_test.go` (amendment 01) | derived counts 69/98, catalogue pin |

`rls_expected_test.go` (G008 oracle) is unchanged.

## Migrations / schema / contracts

- Schema: version 7 (`00007_stored_objects.sql`) as above. 00001–00006 byte-identical (B-MIGRATE-SET).
- Public API/MCP: **none**. No route, handler, contract, BFF or MCP tool; `NewServiceWithObjects` and the five
  methods are an internal Go API pinned by U-ACCESS-BOUNDARY; nothing imports `access` in production.

## Tests added

Unit: U-OBJ-KEY/ADAPTER/PRESIGN/transport (objectstore), U-OBJ-CONFIG (config), tenancy object validators and SQL
boundary, U-OBJ-INPUT, U-OBJ-POLICY (mapping table + the oracle's service cases on synthetic RLS-modelled facts),
U-OBJ-READER (exact reader incl. holdback and Size+1, owned source: blocked first Read, blocked lookahead, size-0 probe,
non-conforming source; 25 MiB streaming allocation bound; verifying reader), U-OBJ-EVENT, U-OBJ-SEAM (budgets with a
fake `pgx.Tx`, incl. a statement inside the transaction), config bounds.
Integration: `TestObjects` (10 subtests incl. the oracle's 297 service and 85 RLS cases), `TestObjectsFootprint`,
`TestObjectsOrdering` (OOR-1..8), `TestObjectsFaults` (19 subtests incl. FX-STALL), `TestObjectsSignHost`,
`TestObjectsPersistSeed/Verify`, `TestObjectsCatalog` (3, incl. the guard source digest and N-OBJ-1..8), `TestObjectsMigration` (5).
S-review additions: the guard transition cases in OBJ-RLS (7 refusals, 2 controls), the `stored_objects_guard`
`md5(prosrc)` pin and N-OBJ-8.

## Commands and validation

Every batch: `git fetch origin` and the exact-delta/ancestor check (`shas` file per batch), then
`bin/batch.sh` (`bash infra/ci/docker-desktop-mounts.sh`, then `flock -n …/local-harness.lock`). Focused batches use
the pinned `golang:1.27.1-alpine` (sha256:8a5910f3…) and `pgvector/pgvector:0.8.7-pg17` (sha256:ac08538c…) and MinIO
built from `infra/docker/minio.Dockerfile` (minio `RELEASE.2025-10-15T17-29-55Z`, mc `RELEASE.2025-08-13T08-35-41Z`);
inner loop, not acceptance evidence. Evidence: `<home>/.local/state/platform-evidence/g009-impl-bd31633/<batch>/`.

| Batch | Candidate / target | Content | Outcome |
|---|---|---|---|
| b01 | f18bfdd+dirty / 24c8bda | focus-go `check` (config, objectstore, tenancy, observability, migrations, migrate) | FAIL: 3 `TestLogExitIdentityConfigKeys` subtests (production http fixture, → amendment 02), my pending version-7 pins and two boundary assertions |
| b02 | 40640bf+dirty / 6b7ede8 | focus-go `check ./...` | PASS |
| b03 | f041517+dirty / 6b7ede8 | focus-go `check ./...` | FAIL: a `TestKeys` case (uppercase of an all-digit UUID), oracle glue mis-routed R74 (empty raw selector) |
| b04 | f041517+dirty / 6b7ede8 | rerun objectstore, tenancy | PASS |
| b05 | 42f9a50 / 6b7ede8 | focus-db migrate: lifecycle, G008 RLS at v6, objects | lifecycle PASS, RLS PASS; objects FAIL (catalog query `"char"` casts; cascade) |
| b06 | 42f9a50+dirty / 6b7ede8 | object catalog/migration rerun | TestObjectsMigration PASS 5/5; TestObjectsCatalog FAIL (N-OBJ-7 detector query) — fixed, covered by b14 |
| b07 | ea9aee0 / 6b7ede8 | TestObjects | FAIL: glue panic (RLS immutable fresh value); service oracle 297/297 |
| b08 | ea9aee0+1 line / 6b7ede8 | TestObjects | FAIL: **RU05 ORACLE-DISAGREEMENT** (428C9, row unchanged; all else PASS) → ruling, oracle r2 |
| b09 | ea9aee0+dirty / 6b7ede8 | TestObjectsOrdering | **interrupted** (rc 143, stopped by this worker): OOR-3 delete-first glue armed the D1 pause on Resolve's COMMIT and hung; OOR-1/2/4–8 printed passing |
| b10 | 4c2127d / 2485ede | TestObjects (r2), ordering, faults | TestObjects PASS (382/382), ordering PASS 8/8; faults: FG level casing (test) and **FX-STALL hung: a real defect** — storage statements used the caller's context, not the decision-budgeted one (fixed: `storageTx` passes its budgeted context to every statement; U-OBJ-SEAM now checks a statement inside the transaction) |
| b11 | 780ca45 / 2485ede | focus-go `check ./...` | PASS |
| b12 | 780ca45 / 2485ede | faults, objects, ordering, footprint | **PASS** all four (FX-STALL unavailable at 5.0 s, connection destroyed); E-OBJ-TRIGGER (INSERT 4/2) → query-cost ruling |
| b13 | b65f665 / f629847 | covering | invocation error (`--owned-registry: not a registry file`); no check ran |
| **b14** | **b65f665 / f629847** | **`run.sh --base f629847 --allow-file …G009-scope-allowlist.txt --report-dir <EVID>/b14-covering/report lint lint-base lint-core unit migrations`** (run d111fb, devtcp probe) | **PASS**: 40 required IDs each PASS once; exit 0; no leftovers |
| b15 | 9487675 / f629847 | focus-go `check ./...` after the S-review corrections | PASS |
| b16 | 9487675 / f629847 | focus-db migrate: lifecycle, G008 RLS at v6, TestObjectsCatalog/Migration | PASS (CT-OBJ-1 with the guard digest; N-OBJ-8 detected) |
| b17 | 9487675 / f629847 | focus-db TestObjects, Ordering, Faults, Footprint | PASS: oracle 297/0 and 85/0, `OBJ-GUARD-TRANSITIONS {"refused":7,"controls":2}` |
| b18 (b4) | 9487675 / f629847 | `run.sh --base f629847 --allow-file … --report-dir <EVID>/b18-b4-all/report all unit` (run 481155) | **FAILED, 95 PASS / 2 FAIL**: B-PROTECTED (this Result updated against base) and B-TOGGLE-OFF (variant missing `core/internal/objectstore`, `APPLICABLE` gap); all B-OBJ-*, B-CI-CONTROLS, B-UNIT-CHECKS and the rest PASS; `OBJ-GUARD-TRANSITIONS {"refused":7,"controls":2}`; 2 E-OBJ-TRIGGER lines |
| **b19** | **6466f6f / f629847** | **`lint lint-base provenance off`** (run 371ed5): the failed suites plus lint for the `run.sh` change | **PASS**: 24 PASS incl. B-PROTECTED (505 files), B-TOGGLE-OFF(-ADDITIONS), B-OBS-OFF, B-BROWSER-OFF, B-SCOPE, B-SECRET-CANARY, B-REQUIRED-CHECKS 23/23 |

**Oracle agreement (b14 B-OBJ-IT):** `OBJ-ORACLE-SERVICE {"cases":297,"disagreements":0}`,
`OBJ-ORACLE-RLS {"cases":85,"disagreements":0}` (RU05 via `oeRLSImmutableGenerated`).

**OBJ-SIGN-HOST (b14):** the unchanged URL signed for `127.0.0.1:9400` fetched through the published port (devtcp,
URL only in a shell variable and the request line written by `printf`) → 200, expected digest, attachment /
`text/plain` / `no-store`; the bare path 403.

**PG-003 (E-OBJ, b12 and b14):** E-OBJ-3 equal for s1 and s50; reader calls per statement: select 2/1, lock 4/3,
update 2/0 (finalize/delete), **insert 4/2 against the ≤2/≤1 hypothesis** (INSERT WITH CHECK + the SELECT policy of
RETURNING). Raw lines kept (`obj-eobj.jsonl`, 2 `E-OBJ-TRIGGER` lines). Disposition: Codex ruling `e1075bb0…`
(characterized observation; hypotheses unchanged; PG-003 open; not a risk acceptance). E-OBJ-2: FOR SHARE on the
Organization, binding, Group binding and Group per write transaction, FOR UPDATE on the one object row only. E-OBJ-4:
0 runtime backends idle in transaction during a blocked transfer.

## Acceptance criteria (final local evidence: b14, b18 and b19)

| AC | Evidence so far |
|---|---|
| AC1 | U-ACCESS-BOUNDARY, U-OBJ-KEY, U-OBJ-INPUT (zero adapter/DB calls), B-MCP-SOURCE PASS, B-SCOPE PASS (no contract/route path) |
| AC2 | OBJ-SCHEMA, oracle RLS forged-creator/key cases, B-OBJ-CATALOG PASS |
| AC3 | oracle service reads 0 disagreements; OBJ-SCOPE 1/50 bindings identical; U-OBJ-POLICY |
| AC4 | oracle put/delete 0 disagreements (admin→member, former-admin Organization object); OBJ-RLS transitions |
| AC5 | oracle RLS incl. FORCE-owner positives/negatives and no-context; the guard transition refusals; N-OBJ-1..8 with the guard digest; B-RLS-* PASS at b14 and b18 |
| AC6 | OBJ-RT on real MinIO with the scoped credential |
| AC7 | OBJ-DEL, FD-* |
| AC8 | OBJ-SIGN (real fetch, 2 s expiry, reuse, alteration, anonymous, loopback Host) and OBJ-SIGN-HOST |
| AC9 | OOR-1..8 (B-OBJ-ORDERING PASS) |
| AC10 | U-OBJ-INPUT, U-OBJ-READER, U-OBJ-SEAM budgets, OBJ-RT boundaries |
| AC11 | B-OBJ-FAULTS 19/19 incl. FP-LONG/BLOCK1/BLOCKN/ZERO and FX-STALL |
| AC12 | TestObjectsMigration, the RC pins, B-RLS-*, B-POLICY-*; root/ops/rename at version 7 (b18), off at version 7 (b19) |
| AC13 | OBJ-LOG, OBJ-ADMIN, U-OBJ-EVENT, U-OBJ-CONFIG, B-SECRET-CANARY PASS; B-IDENTITY-CONFIG (b18); OBJECT-STORAGE.md |
| AC14 | this Result; full local profile PASS across b18 + b19; hosted staging disposition after an authorized conditional merge |

## Security review (Plan §6 S-review)

- Reviewer: native Agent (type Plan, effort high, read-only, no worktree), dispatched by the coordinator
  (`dispatch-record-sreview.md`). Candidate `9f8f0c3` on staging `f629847`. Report:
  `<home>/.local/state/platform-evidence/g009-impl-bd31633/s-review.md`, SHA-256
  `2e03b76f9d1b5afbae0d5f361e740f395f9e2f2b2f7b96e26ec3d607de81d47c`. The coordinator verified that the oracle at the
  candidate is byte-identical to r2.
- **Verdict: FINDINGS, highest Medium (test adequacy).** No exploitable isolation, authorization or secret-leak defect.
- Corrections (Plan §8 step 8, coordinator-relayed, all in paths already in the 33-pattern allow-list; checkpoint
  `9487675`; the oracle file is untouched):

| # | Severity | Finding | Fix |
|---|---|---|---|
| 1 | Medium (test adequacy) | 00007's creator-only finalize (:178) and same-state refusal (:184-186) were untested and the guard body unpinned | `migrate/objects_integration_test.go`: `md5(prosrc)` of `stored_objects_guard` pinned (`af1143d8…e0e4`, the exact 00007 body) in the version-7 detector; rolled-back N-OBJ-8 replaces the guard body and is caught (`obj-guard-src`). `access/objects_integration_test.go` OBJ-RLS: `guardTransitions`, 7 runtime refusals each expecting 23000 `stored_objects_transition` with the row unchanged (g1Admin and aAdmin finalizing g1MemberX's pending row; the creator rewriting sha256 of, and re-readying, a ready row; the g1 admin updating a deleted row to deleted and to deleting; aAdmin rewriting a deleted row's sha256) beside 2 positive controls (the creator's finalize, the deleter's tombstone); printed as `OBJ-GUARD-TRANSITIONS`. No new subtest, so the harness counts are unchanged |
| 2 | Low | `ownedSource.Read` checked `closed` before registering in flight, so a Read could slip past Close and the reap check | `objects_io.go`: `closed` is a plain bool under `mu`; Read checks it and increments `inflight` in one critical section; Close sets it under `mu` before closing Body |
| 3 | Low | `obj_env` opened `obj-it.env` before `umask 077` applied | `run.sh`: the redirect is inside the subshell after `umask 077`, and the file mode is asserted 600 |
| 4 | Low (ordering) | Delete ran RLS discovery before the liveness check | `access/objects.go`: D1 and D4 run `LockLiveSession` first (binding, liveness, RLS discovery, facts, decision per Plan §1 item 3); removed from `decide` |

- Reruns on `9487675` against `f629847`: b15 (focus-go), b16 (migrate focused), b17 (storage tests focused) PASS; b18
  (b4 `all unit`) PASS except two harness-reconciliation failures unrelated to the corrections, fixed in `6466f6f`
  and rerun in b19 (PASS).

## Architecture deviations and decisions within scope

1. **`docs/foundation/VENDORS.md` not modified.** B-TOGGLE-OFF-ADDITIONS requires the root VENDORS.md to equal its
   render with exactly the pinned G004 identity rows, so an R2 note there would fail an accepted check. The R2 notes
   (with exact citations and the unverified items) are in `OBJECT-STORAGE.md`. Narrowing only (no new path).
2. **`MaxObjectBytes` lives in `objects.go`**; `tenancy/boundary_test.go`'s `constStrings` skips numeric constants
   (a first attempt put it in `tenancy.go`, outside the allow-list; reverted before any harness run).
3. **G008 migration tests at version 6.** `TestRLSCatalog` steps down from 7; `TestRLSMigration`, persistence and
   provisioning use a version-6 subset runner, so their exact version-6 guarantees are kept unchanged.
4. **Raw selectors** in the oracle glue are parsed at the API boundary (`tenancy.ParseObjectID`/`ParseGroupID`); the
   all-zero UUID is passed as the zero selector so the service's own refusal runs.
5. **Defect found and fixed in attempt 1 (b10):** statements inside storage transactions ran on the caller's context,
   so the decision budget bound only the context statement and COMMIT; FX-STALL exposed it. Now every statement uses
   the budgeted context.
6. **`APPLICABLE` reconciliation (found by b18).** Plan r1's RC list missed that the intelligence-disabled variant
   copies only the `run.sh` `APPLICABLE` additions. The three G009 additions are now listed, as G008 RC-13 did for its
   paths. This is a `run.sh` edit inside the allow-list: no new path, and no check was weakened.
7. **Result at base bytes during final validation.** B-PROTECTED requires `management/` to equal base. This Result was
   imported to staging at prep, so it was returned to its base bytes for b19 and written after validation, with no
   `--expect-hash` exemption. Codex can rerun B-PROTECTED on the final bytes with `--expect-hash` if it wants them bound.
8. No architecture change, new vendor, role, definer, public surface or schema beyond Plan r1.

## Unresolved issues, blockers and residuals

- Pending: Codex pre-integration review; then, only after Codex acceptance, the MR, and the hosted staging disposition
  after an authorized conditional merge.
- PG-003 stays open (query-cost ruling `e1075bb0…`; not a risk acceptance).
- Residuals X1–X7 as documented in OBJECT-STORAGE.md; R2 not cloud-qualified.

## Suggested follow-up

None beyond the Plan's G010/G011/G013 handoffs.

## Correction pass 1 (attempt 2 of B5)

### Review input

- Codex review **CORRECTION LOOP**: staging `6ed5363f9b9519a6f42357dfec776f418f9e0802`,
  `management/v0/results/G009-implementation-codex-review.md`, SHA-256
  `eb0ef6c2264156faafd1e4fe245f443e4e2cd1b9c226e3decdc6e8667c009e7d`.
- Reviewed candidate `787ca4e`; Result `b0de505b…`. RLS and security were otherwise accepted, and deviations 1–4 and
  6–7 were accepted as bounded reconciliation.
- Correction pass 1 is attempt 2 of B5. Scope and the fourteen ACs are unchanged.
- **Finding C1 (P2):**
  - `upload()` read `xr.sum()` unconditionally after `ownedSource.waitReleased` (objects.go:536).
  - `ownedSource.Read` deregisters the Body.Read before the surrounding `exactReader.Read` performs its `hash.Write`
    (objects_io.go:46, :210, :265).
  - So an asynchronous transport returning on cancellation, while a conforming source releases data, can run
    `hash.Sum` concurrently with `hash.Write`. This violates R7/AC11 cancellation safety.

### Reconciliation

- Staging advanced `f629847` → `6ed5363`. The exact delta is two management paths: `management/v0/PROGRESS.md` and the
  Codex review. Neither is a test, harness or code input.
- I rebased the branch onto `6ed5363` with no conflicts: `787ca4e` → `80a918c`.
- Every batch below records `shas`: target `6ed5363`, which is an ancestor of the candidate.

### Fix (checkpoint `a6fae5e`; `objects_io.go`, `objects.go`)

- `exactReader.hash` is touched only by the transport goroutine.
- `final()` performs the last byte's `hash.Write`, then takes the digest snapshot (`hash.Sum`) on that same goroutine.
  It then publishes the snapshot together with `verdictExact` under `mu`, and only if no other verdict was set.
- `result()` returns the verdict, and the digest only with `verdictExact` (zero otherwise). `upload()` uses `result()`
  and never reads the hash.
- `sum()` is removed.
- No transport join was added. The transfer deadline, reap grace, compensation and residual X6 are unchanged.
- Test seam: `objectHooks.sourceRead`, mirrored as `exactReader.afterRead`. It is inert (nil in production) and runs on
  the transport goroutine after each source Read returns, before digesting. It follows the existing `stage`/`crash`
  seams.

### Regressions (`objects_test.go`, `TestUploadDigestSnapshot`; `TestExactReader` uses `result()`)

- **No premature publication.**
  - While the final byte's lookahead is blocked, the reader reports pending and no digest.
  - After a data-bearing release with EOF, it reports exact with the digest of all Size bytes.
- **Asynchronous transport cancelled while the source is blocked; the source releases data as it is closed** (50
  iterations each):
  - mid-stream data;
  - the final byte followed by closed;
  - the final byte with EOF, where exact is permitted only with the full digest.
  - Every iteration requires: no digest without exact, Body closed exactly once, the return within the reap bound and
    no leaked goroutine.
  - Each iteration waits for the transport goroutine's last write.
- **The held C1 window, deterministic:**
  - The `sourceRead` seam holds the transport goroutine after the source's data-bearing release, and before the
    digest write, until a timer armed before `upload` fires.
  - `upload` must return inside the hold with neither the exact verdict nor a digest. Cases: mid-stream, and final
    byte with EOF.
- **Data-bearing conforming sources** (`iotest.DataErrReader`, `OneByteReader`, `HalfReader`; sizes 1, 2, 10 and
  70000):
  - through the reader and through `upload` with an SDK-like transport;
  - each gives exact with the digest of exactly the Size bytes, Body closed once.

### Race-detector prerequisite

- `go test -race` needs cgo, and the pinned `golang:1.27.1-alpine` has no C toolchain.
- The focused race runs therefore use the Debian variant of the same pinned Go version: `golang:1.27.1`,
  `golang@sha256:162be5298a40ed317005c8339c6de4d10d3eef336d66dc8e9259b03ab9d3a6d2`, `CGO_ENABLED=1`, via
  `bin/race.sh`.
- This is an inner-loop tool, not a harness change.

### Batches (candidate / target)

| Batch | Candidate / target | Content | Outcome |
|---|---|---|---|
| b20 | 80a918c+3 dirty / 6ed5363 | focus-go `check ./...` (first fix, before the seam) | PASS |
| b21 | 80a918c+3 dirty / 6ed5363 | `-race -count=3` reader/upload tests (first fix) | PASS, 0 data races |
| b22 | scratch: pre-fix reader/upload + shim / — | the first regression on the pre-fix code | FAIL on the new digest contract ("verdict 4 published a digest"), but **no race reproduced**: the transport goroutine always won the window. This motivated the deterministic `sourceRead` hold |
| b23 | 80a918c+3 dirty (= `a6fae5e` content) / 6ed5363 | focus-go `check ./...` with the seam and the held cases | PASS |
| b24 | 80a918c+3 dirty (= `a6fae5e` content) / 6ed5363 | `-race -count=3`: TestExactReader, TestOwnedUploadSource, TestEmptyUploadProbe, TestUploadStreamsWithoutBuffering, TestVerifyingReader, TestUploadDigestSnapshot, TestObjectInputs, TestStorageTransactionBudgets | **PASS, 0 data races** (one run hit the snapshot window: 49/50 exact, consistent) |
| b25 | scratch negative control: `a6fae5e` with only the C1 snapshot reverted (`upload` reads `xr.verdict(), xr.sum()`) / — | `-race` TestUploadDigestSnapshot | **FAIL as intended**: 6 `WARNING: DATA RACE` reports, "Write by `(*exactReader).Read`, previous read by `(*exactReader).sum()` from `(*Service).upload()`", plus both held cases "returned … with a digest". The regression detects C1 |
| b26 | a6fae5e / 6ed5363 | focus-db TestObjects, Ordering, Faults, Footprint | **PASS**: oracle 297/0 and 85/0; guard 7/2; OBJ-RT exact round trips (digest-verified reads); Faults 19/19 incl. FP-SHORT/LONG/LONG-SUCCESS/BLOCK1/BLOCKN/ZERO, FX-CANCEL, FX-CRASH, FX-STALL (5000 ms) |
| b27 | a6fae5e / 6ed5363 | the b14 covering profile: `run.sh --base 6ed5363 --allow-file … lint lint-base lint-core unit migrations` (run a66d00) | **PASS**, 40/40 required IDs (details below) |

**b27: PASS** (run a66d00, candidate `a6fae5e`, base `6ed5363`, harness exit 0). 41 PASS. All 40 required IDs of the
executed suites passed exactly once (B-REQUIRED-CHECKS), including:

- B-UNIT-CHECKS;
- B-OBJ-CATALOG 3/3, B-OBJ-MIGRATION 5/5, B-OBJ-IT and B-OBJ-ORDERING 8/8;
- B-OBJ-FAULTS 19/19;
- B-SECRET-CANARY;
- B-SCOPE: 50 changed paths since `6ed5363`, all within the allow-list.

B-OBJ-IT recorded `OBJ-ORACLE-SERVICE {"cases":297,"disagreements":0}`, `OBJ-ORACLE-RLS {"cases":85,"disagreements":0}`
and `OBJ-GUARD-TRANSITIONS {"refused":7,"controls":2}`, plus the 2 ruled E-OBJ-TRIGGER lines. No resources were left
behind.

### Evidence reuse (exact delta)

- `6466f6f` (b19) → this candidate changes only:
  - `objects.go` (`upload` uses `result()`; `sourceRead` wiring);
  - `objects_io.go` (snapshot, `result()`, seam);
  - `objects_test.go`;
  - this Result;
  - the staging delta's two management paths, which are not test inputs.
- b27 reruns every suite whose inputs these changes touch: `unit` (B-UNIT-CHECKS), `migrations` (the B-OBJ-* storage
  tests) and the lint/scope/canary checks.
- Reused unchanged from b18/b19 (inputs unchanged): ci-controls, controls, contracts, root, ops, rename, off,
  lint-web and provenance.
- B-PROTECTED (b19) does not see this Result's report-only bytes, as in attempt 1.
- No CI-controls, browser or full local run was repeated for this correction.

### Status

- C1 is resolved.
- No new path. All changed paths are within the 33-pattern allow-list, and the oracle is unchanged.
- PG-003 stays open. No MR is open. Returned to Codex for re-review.

## Post-integration correction pass 2 (attempt 3 of B5)

### Input

- G009 was merged as MR !24 (`e087f2c`). The post-merge staging pipeline `2946564278` gate job `17169253060` **failed**
  after 2527.7 s. Fast job `17169253059` succeeded.
- Only B-WEB-ERROR-OUTPUT failed, on WE4 `vacuous`, plus the B-REQUIRED-CHECKS aggregate.
  - All five B-OBJ-* checks passed. Cleanup: 4 removed, 0 remaining.
  - `error-output-web.tsv`: we1–we3b 500, we5 200 and we6 0 match the passing runs; only we4 differs.
- Codex authorized this bounded correction (attempt 3 of B5): staging `6bcddc1`,
  `management/v0/results/G009-implementation-codex-review.md`, SHA-256
  `91d2eb04acd10dbe21f200a87681396f212d258d17604bd8ac8d65aff274586c`.
- Scope: only the WE4 control block of `infra/validation/run.sh`, and this Result.
- Hosted evidence: `<home>/.local/state/platform-evidence/g009-impl-bd31633/hosted-staging-gate/` (job log,
  results.tsv, error-output-web.tsv, summary.json).

### Diagnosis

The WE4 check was `sleep 2` followed by `if docker logs "$ccid" 2>&1 | grep -qF "$can"`, under run.sh's
`set -euo pipefail`. The hosted context is the CI context image: Debian trixie with GNU grep and the docker 29.8.2 CLI
(`infra/docker/ci-context.Dockerfile`).

Focused characterization (`bin/we4-probe.sh`, `bin/we4-block.sh`) used the production web runtime image with the probe
routes and the validation tools image, both built from staging `6bcddc1` exactly as `web_error_output` builds them. Log
reads used both the host CLI 27.5.1 and the hosted CLI 29.8.2, extracted from the pinned `docker:29.8.2-cli` image and
run on the host.

| Probe | Observation (CLI 27.5.1 / 29.8.2) |
|---|---|
| T1 timing | the canary is in a full `docker logs` capture **immediately** after the 500 response (t = 0) and unchanged at 0.1, 0.25, 0.5, 1 and 2 s; the log has 7 lines |
| T3 shape | **one log entry follows the canary line** (Next's error output) |
| T2 run.sh's exact pipeline, 60× | 60/60 pass locally (`PIPESTATUS 0/0`) |
| T4 the same 7-line log, reader exits after one byte, 60× | docker logs exit **141 (SIGPIPE)** in 1/60 (27.5.1) and **4/60 (29.8.2)**, otherwise 0. The CLI writes the entries in more than one write, so a reader that leaves early can SIGPIPE it on this very log |
| T6 a 200 000-line log, early reader, 10× | 141 in 10/10 with both CLIs (the CLI dies on a broken pipe) |
| S-old (stub: the canary line, then one more entry 0.3 s later) | the old pipeline reports **vacuous with `PIPESTATUS 141 0`**: grep matched and exited 0, docker logs died with SIGPIPE, and pipefail made the `if` false |

**Conclusion.**

- **Hypothesis (b) is the demonstrated mechanism; the original hosted cause remains unproven** (see the clarification
  below). `grep -q` exits on the canary line while docker logs still has the
  following entry to write. docker logs then dies with SIGPIPE (141), and under pipefail the condition is false, so a
  present canary is reported as `vacuous`.
  - The mechanism is demonstrated on the exact WE4 log with the hosted CLI (T4, 4/60). Whether it triggers depends on
    scheduling, which explains the local passes and the single failure on the loaded hosted runner (the gate runs the
    background controls in parallel).
- **Hypothesis (a) is not supported.** The canary is present at t = 0 in every local read, well inside the existing
  2 s wait.
- The retained hosted artifact cannot prove which branch fired. Its symptom (status 500 checked separately and passed;
  presence false) is exactly what (b) produces.

### Change (branch `g009-we4-correction` from staging `6bcddc1`; `infra/validation/run.sh` WE4 block only)

- The fixed `sleep 2` and the pipeline are replaced. Each attempt reads the whole log into a private file
  (`umask 077`, `$D/we4-control.log`). A failed read sets `log-read-failed` and fails closed. The file is then searched
  with `grep -qF`, so no early consumer is attached to docker logs.
- The read is retried for at most 20 non-following captures with 0.5 s delays, about 10 s of polling (not an
  end-to-end deadline; see the clarification below), until the canary appears. Absence after the
  bound is `vacuous` and fails.
- The canary-bearing copy is deleted immediately; `$D` is still removed at the end as before.
- `we4.status` records `canary-present`, `vacuous` or `log-read-failed`. Only run.sh reads it.
- Unchanged: the WE4 500-status check, the exact canary requirement, the control route and fixtures, every other check,
  production web code and CI profiles.
- Commits:
  - `3c8d2ab` the fix;
  - `8b5108e` quoting of the three status values, because B-SHELL (shellcheck SC2100) misread the hyphenated
    assignments as arithmetic. No behaviour change.
- Diff: `infra/validation/run.sh`, +12/−2 against staging.

### Validation (candidate / target; every batch preceded by fetch, docker-desktop-mounts and the harness flock)

| Batch | Candidate / target | Content | Outcome |
|---|---|---|---|
| c2-b01, c2-b02 | 6bcddc1 / 6bcddc1 | `we4-probe.sh` T1–T4 (host CLI) | as tabled above; no local reproduction of the pipeline failure |
| c2-b03, c2-b04 | 6bcddc1 / 6bcddc1 | `we4-probe.sh` T1–T4, T6 with CLI 27.5.1 and 29.8.2 | T4 141 in 1/60 and 4/60; T6 141 10/10 each |
| c2-b05, c2-b06 | 6bcddc1 + the fix (= `3c8d2ab`) / 6bcddc1 | `we4-block.sh`: the WE4 block extracted verbatim from run.sh and run under `set -euo pipefail`; CLI 27.5.1 and 29.8.2 | **P** present: ok=1 `canary-present` (≈30 ms). **A** absent (a different fresh canary): ok=0 `vacuous` after the bound (≈10.5 s). **F** read failure (removed container): ok=0 `log-read-failed` (≈25 ms). No canary copy left. **S-old** `vacuous (PIPESTATUS 141 0)` vs **S-new** ok=1 `canary-present` |
| c2-b07 | 3c8d2ab / 6bcddc1 | `run.sh --base 6bcddc1 --allow-file … lint lint-base root` | FAILED: B-WEB-ERROR-OUTPUT **PASS** (`we4 canary-present`; we1–we3b 500, we5 200, we6 0) and B-CORE-ERROR-OUTPUT PASS; **B-SHELL FAIL** (SC2100 on the two unquoted hyphenated assignments), so B-REQUIRED-CHECKS failed; fixed in `8b5108e` |
| c2-b08 | 8b5108e / 6bcddc1 | `we4-block.sh` on the final bytes, CLI 29.8.2 | P / A / F / S identical to c2-b06 |
| **c2-b09** | **8b5108e / 6bcddc1** | **`lint lint-base root`** (run 22977b; the failed batch rerun) | **PASS**: 31/31 required IDs; B-WEB-ERROR-OUTPUT with `we4 canary-present`; B-SHELL PASS |

**c2-b09: PASS** (run 22977b, candidate `8b5108e`, base `6bcddc1`, harness exit 0). 32 PASS. B-REQUIRED-CHECKS:
all 31 required IDs of the executed suites passed exactly once. Highlights:

- **B-WEB-ERROR-OUTPUT PASS**; `error-output-web.tsv`: we1–we3b 500, we5 200, we6 0, **we4 `canary-present`**.
- B-CORE-ERROR-OUTPUT PASS.
- B-SHELL PASS.
- B-SECRET-CANARY PASS (22 secrets).
- B-SCOPE PASS: 1 changed path since `6bcddc1`, within the allow-list.

No owned containers remain.

At commit time this Result is the second changed path, which is allow-listed and report-only.

### Evidence reuse

- Every other suite's inputs are unchanged by this run.sh-WE4-only delta. The merged `e087f2c` content and its hosted
  results for all other checks, including the five B-OBJ-* checks, stand.
- No full local, controls or browser run was repeated.

### Status

- WE4 is corrected in the minimum scope.
- The attempt-3 checkpoint is due: B5 has 2 attempts remaining, and the budget is not extended.
- Next: Codex diff review, then the normal MR/Duo route and the required staging disposition. No MR is open.

### Codex review and attempt-3 checkpoint (report-only addendum)

- **Codex ACCEPT for integration.** Candidate `7f8da8a` was reviewed in
  `management/v0/results/G009-implementation-codex-review.md`, section "WE4 correction pass 2 / attempt 3
  pre-integration review — 2026-10-14 UTC", on staging `558e727`; the file's SHA-256 is
  `c15d56b94f54ad8cba5e235bdfcc5891a70c176e25a28041d48851cedd216d07`. Result `83e60440…` was imported exactly with it.
- **Attempt-3 checkpoint disposition**, recorded by Codex in that section's "Attempt-3 checkpoint disposition":
  - The scoped strategy is converging.
  - The only failed staging obligation, the inherited WE4 control, is corrected and locally covered.
  - Remaining work: MR/Duo disposition, qualifying staging evidence, integrated-content reconciliation and bounded
    final closeout.
  - No scope split, budget extension or new harness project is justified.
  - **3 of B5 used, 2 remaining.**
- **Clarification 1, the cause.** The original hosted cause of the failed WE4 run is **unproven**. Local timing
  observations do not rule out every hosted delay, and the exact old `grep -qF` pipeline did not fail in the local
  60-run samples. The correction removes the demonstrated early-consumer SIGPIPE mechanism (T4, T6, S-old/S-new). No
  evidence was recovered from the failed runner.
- **Clarification 2, the bound.** "Bounded 10 s" means at most 20 non-following `docker logs` captures with 0.5 s
  delays, about ten seconds of polling. It is **not** a hard end-to-end deadline if Docker itself is slow; the existing
  outer execution limits still apply.
- The two in-place wording changes above, under the diagnosis's conclusion and the change description, apply these
  clarifications. No harness bytes changed, and no tests ran for this addendum.

## Integration and hosted staging qualification

| MR | Merge commit | Content |
|---|---|---|
| !24 | `e087f2c` | G009 implementation (attempt 1 and correction pass 1) |
| !25 | `944ecf3e4e295c196a406f56b55f35226b511c35` | correction pass 2: the run.sh WE4 block (bytes equal the reviewed `7f8da8a`) |

Hosted staging gate history. The failed runs stay failed; all three outcomes are preserved:

| Pipeline / job | Commit | Outcome |
|---|---|---|
| 2946564278 / gate 17169253060 | `e087f2c` | **FAIL**: B-WEB-ERROR-OUTPUT WE4 `vacuous`, plus the B-REQUIRED-CHECKS aggregate (2527.7 s). All five B-OBJ-* checks passed; cleanup 4 removed, 0 remaining. Led to correction pass 2 (above) |
| 2946943462 / gate 17172092975 (fast 17172092974 succeeded) | `944ecf3` | **FAIL**: B-DOCS-EXEC-ROOT `start block`, plus the missing-check aggregate (2201.8 s). The root suite's later checks, including WE4, were not reached, so this run is not evidence of a WE4 failure. The published build ledger lacks the completed MinIO server steps/export and one minio-init step, with no captured build error. **Cause unknown**: it is consistent with a build interruption, but it is **not a confirmed network or runner issue**, and no deterministic source failure was identified. Cleanup verified (4 resources, 0 owned remaining) |
| 2946943462 / gate **17173244141** (the single Codex-authorized exact-job retry of 17172092975, same pipeline) | `944ecf3` | **PASS**: final status 0, finished 2026-10-14T14:37:17Z after 2680.23 s. 69 PASS, 0 FAIL (6 INFO), including all 68 required IDs (B-REQUIRED-CHECKS) of `controls ci-controls root migrations ops rename off`, B-CI-CONTROLS (441 controls), the five B-OBJ-* checks, B-DOCS-EXEC-ROOT, and B-WEB-ERROR-OUTPUT with `we4 canary-present`. `cleanup_status=0`, `owned_remaining=0`. Verified independently by Codex |

Retry evidence (coordinator copies, `<home>/.local/state/platform-evidence/g009-impl-bd31633/hosted-staging-gate-2-retry/`):

| File | SHA-256 |
|---|---|
| `gate-job-17173244141.log` | `32ee5683bf0e5798f9ce2b00cc9d5fa08bab87e86a9092e20c8f01d64c92b0cf` |
| `results.tsv` | `3400d23a8333a72ebfd31e8e19bf5fde4529b7353a83a4800b4c2c079c93b9e3` |
| `summary.json` | `a53fc595a2afaca857bd2a793e26876f48e88cc1020d738313d29cfdb8b98b28` |
| `error-output-web.tsv` | `fa4ef5dc72ca27baee468425227082ec2286adc23f52f3e3eafad48beb3e934c` |

- The retry was authorized in `G009-implementation-codex-review.md`, section "Second staging gate / single hosted retry
  — 2026-10-14 UTC". It is an unchanged-source qualification retry and consumes no Goal attempt.
- **B5 ledger:** 3 of 5 attempts used (attempt 1; correction pass 1; post-integration correction pass 2). The attempt-3
  checkpoint was reviewed by Codex. No extension.
- PG-003 remains open, per the query-cost ruling.
- Remaining: integrated-content and cleanup confirmation, and Codex final ACCEPT. G009 is **not** claimed accepted.

