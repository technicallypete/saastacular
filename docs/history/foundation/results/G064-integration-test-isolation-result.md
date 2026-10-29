> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G064 — Integration-test isolation and per-test evidence — Result

Goal: [G064](../goals/G064-integration-test-isolation.md) · Spec: [G064 Work Spec](../specs/G064-integration-test-isolation-spec.md) ·
Plan: [r2](../plans/G064-integration-test-isolation-plan.md) (sha256 `a524a0737ed151cbec1fbd113626cd478d512c740da7b9594b7212d00a7c4837`) ·
Allow-list: 13 ERE lines (not included) (sha256 `4ee27e22c31dab708c78524d3fba71bb859a0fc7f4b2d17ea299db4ecb999627`) ·
Release: Plan r2 ACCEPT and execution release (not included) (sha256 `36da81ee11479719439a331ccc5e79baac62e397f92154d8bd2b816cabd4acf1`).

## Status

**S2 checkpoint — slices S0-S2 submitted; NOT accepted.** S1's real-PostgreSQL lifecycle proof (batch 2a) PASSED on a
private instance (coordinator evidence `platform-evidence/g064-a1/s1b-f58b33f2/`). S2 (template in the `migrations`
lifecycle, the 50 reset points, the identity isolation window, the tenancy setup line) is written and passes stack-free
checks; its real-PostgreSQL batch is the coordinator's next step (see "S2 record" below). S3 (reporter), S4 (campaign),
S5 (required-checks, docs) are not started. Only Codex accepts.

## Ownership, inputs and resources (S0)

| Item | Value |
|---|---|
| Writer | single implementation writer; Claude Opus 5.5 (`claude-opus-5-5`); effort requested high by dispatch, runtime effort not observable from inside the agent (unknown) |
| Checkout | isolated worktree `.claude/worktrees/agent-<id>`, branch `g064-impl` |
| Base | `origin/staging` `310baf2c9eb22de7bfc31f38f225b48bd4b8e48c`; release commit `c4af9f06` is an ancestor; `c4af9f06..310baf2c` changes only `management/SUPERVISOR-HANDOFF.md` |
| Governing hashes | Plan, allow-list and release record match the values above (re-hashed at HEAD); Goal sha256 `438df4fc…a0ad0c0`, Spec sha256 `d8e23cb7…ffb3`, both equal the Plan's bindings |
| Bound blobs (Plan §1.3) | **46/46 match at HEAD; 0 mismatches** |
| r1/r2 additional bindings (recorded, not edited) | `infra/ci/validate-controls.sh` `154733b5b86a21af91e7afe1e0acb4401d4a39da`; `core/internal/migrate/metadata.go` `02a215e929a86b9f4d793234fa0c8fbca1fe1a0f`; `core/go.mod` `df94f5f489f1cb4b290a2a7e3cfc9a525de43ee2`; `infra/docker/validation-tools.Dockerfile` `b1e82fdda161b1540d36a0ce602e002d300d947a`; `infra/validation/tools/go.mod` `e997b492b647c8428c05657a800f559908d7e90c`; `infra/recovery/sql/fingerprint-catalog.sql` `6eff4d6823fe1c39ebff3622ebd99f2cee963e90`; identity untagged tests: `allowlist_test.go` `fa7be0bb…`, `assurance_test.go` `d734915b…`, `attest_test.go` `0f477a10…`, `credential_test.go` `fa0b1655…`, `enroll_test.go` `889a6662…`, `http_test.go` `06a1cb7b…`, `identity_test.go` `41b80aac…`, `limiter_copy_test.go` `29018857…`, `mcp_name_test.go` `c987a9c6…`, `mq_eval_test.go` `45d9f7a4…`, `retire_test.go` `4239d6d9…`; workos: `adapter_test.go` `75d86fc8…`, `connect_test.go` `f81950ab…`, `jwks_test.go` `2e8571ab…`, `token_test.go` `c63512dd…`, `staging_test.go` (excluded tag) `b7a60a26…` |
| Migration set | 19 files `00001`..`00019`, contiguous; sha256 over `sha256sum *.sql` = `37e09c4cb5f1d40935eee1b835bb1f65c11a7676e616e4c4053a71378777373f` (the `mig_digest` form `_db_source` records) |
| Copied-tree provenance | `copy_repo`/`impl_files`/`impl_tree_hash` (`infra/validation/lib.sh:159-171`) unchanged: working-tree path set of `$REPO`; the template record binds `impl_tree_hash` of the copy, HEAD and a `head`/`dirty` flag, `--base` only as baseline. Observed: `impl_tree_hash` exits non-zero because `sha256sum` skips the three tracked directory symlinks (`.agents/skills/restruct-sdlc`, `.claude/skills/restruct-sdlc`, `.claude/skills/vestive-design`) — existing behaviour shared with `B-SOURCE-HASH` (`run.sh:5053` ignores the status); `_db_source` therefore requires a well-formed digest instead of a zero status |
| Grep facts | `CREATE EXTENSION`/`ALTER SYSTEM` in `core/internal/**/*_test.go`: 0. `MATERIALIZED VIEW`/`PARTITION BY` in `core/migrations`: 0. `INSERT INTO app.|audit.` in migrations: 66 lines, **0 top-level, 0 directly in a DO body**: 50 inside function bodies, 16 inside nested `EXECUTE` strings of DO blocks (dollar-quote depth ≥ 2) — no migration-time seed rows. Integration-tagged files 52, top-level `Test*` 112 (migrate 36, identity 35, access 31, mcp 5, signup 3, tenancy 2). Untagged identity tests 39, workos 21, `identitytest` has no test files — all as §4 #53 |
| Tools module / image | tests run by the validation-tools `check` target (`B-CONTRACT-TOOLS`, `run.sh:1714-1724`); build list at `infra/docker/validation-tools.Dockerfile:10` (contractcheck httpprobe logcheck identitysetup; `vulnpolicy` exists under `cmd/` but is not in the list) — S3 inputs |
| EXIT trap | `run.sh:109` `trap '…; bg_stop_all; export_diagnostics …; cleanup' EXIT`; `stop_harness_group` (`infra/ci/validate.sh:668-688`) sends TERM to the group, then KILL after `STOP_GRACE_DS`; TERM runs the EXIT trap by the existing G003 F26 design (`run.sh:113-114`), KILL does not — S3 input for `go_tests_finalize` |
| Citation drift (non-material) | Plan cites `10-roles.sh:1044-1045` and `:1061-1063`; the bound blob (75 lines) has the database grants at `:55-56` and the role `search_path` settings at `:72-74`, content as the Plan describes. `provision-roles.sql:182-183` repeats the grants. `<db>_runtime` is provisioned `NOINHERIT` (`:38`) |
| Resources used by the writer | none at runtime: no Compose stack, database, browser, image build/pull, network, volume or port. Static checks only (below); one read-only, network-disabled `koalaman/shellcheck:v0.11.0` container per lint run (image already local) |
| Resources for batch 2a (coordinator) | one build slot (migrate image), one slot of the repository-wide two-run pool (DB-only private run), no published port, no pull; G063's frozen qualification checkout/run untouched |

## Implementation summary (S1)

`infra/validation/lib.sh` (new block `:701-1061`, Plan §3.1-§3.3):

- `_psql_on`/`psql_maint` — superuser psql over the container socket, SQL on stdin (CREATE/DROP DATABASE outside a
  transaction block), psql `:"v"`/`:'v'` variables; every lifecycle statement on the maintenance database `postgres`.
- `db_shared_catalog DEST [TPL]` (`:840`) — family roles + `postgres` with attributes, memberships crossing either
  endpoint, family/`<db>`/`<tpl>`/dangling-database settings, owner/flags/ACL of `<db>` and the template; sorted (C).
- `db_pristine_verify DEST TREE STATUS_LOG` (`:877`) — (a) status listing per bound file + `current_version`=MIG_TOP and
  the SQL latest-row applied set `{0} ∪ bound` (`_db_goose_check`, `:791`), raw history recorded; (b) `catalog_p`/
  `catalog_roles` with `MIG_RELATIONS`; (c) fingerprint recorded; (d) every `app`/`audit` `relkind r/p/m` relation
  except `app.goose_db_version` empty (`_db_baseline`, `:811`); (e) owner/ACL/no `<db>` setting; (f) no client session.
- `db_template_create DEST PROJECT TREE STATUS_LOG [NAME]` (`:944`) — source binding first (derived MIG_TOP/bound,
  `mig_digest`, provisioning digests, tree hash, HEAD/flag, image IDs, base as baseline), `DB_TPL_REQUIRE_ROWS` (empty
  until S2 sets `B-MIGRATE-CLI B-MIGRATE-CATALOG` in the migrations lifecycle), its own `db_pristine_verify`, then
  `CREATE … TEMPLATE`, `IS_TEMPLATE true ALLOW_CONNECTIONS false`, CONNECT revoked; record `$WORK/db-lifecycle/NAME.rec`.
- `db_fresh DEST PROJECT PURPOSE [RECORD]` (`:1003`) — preconditions (template closed; record digests/tree = tree now;
  shared catalog = snapshot with `<db>`-level ACL/setting deltas recorded as residue; zero client sessions), plain
  `DROP DATABASE`, `CREATE … TEMPLATE`, exact 10-roles grants, postconditions (goose predicate and history listing,
  fingerprint, shared catalog = snapshot, baseline set by name all zero); one `fresh` row with PG-010 durations.
- `db_lifecycle_export` (`:743`) — atomic copy of `db-lifecycle.tsv` to `REPORT_DIR` after every row.

`infra/validation/run.sh`: `MIG_RELATIONS=79` (`:2474`) hoisted from the five `mig_step` literals and the two
`${r_expect}79` uses of `B-MIGRATE-CATALOG` (`:2751-2763`; the PASS text is byte-identical at 79); `it_isolation` (`:2480`)
with `_iti_t1` T-1, `_iti_t1a` T-1a, `_iti_t1b` T-1b, `_iti_t1c` T-1c, `_iti_t2` T-2, `_iti_t3` T-3, `_iti_t4` T-4,
`_iti_t5` T-5a/b/c, `_iti_t7` T-7, recording `B-IT-ISOLATION`. **Not called by any suite yet** (S2 wires it; the ID
joins `required-checks.tsv` at S5).

`management/v0/results/G064-s1-private-run-driver.sh` — the disposable batch 2a driver (Plan §6 "driven by a
disposable script that sources the new lib.sh helpers"), kept beside this Result because the isolated writer cannot
write the main checkout's `.local/`; it is not harness code and no suite calls it.

Schema/migrations: **none**. Contracts/API/MCP: **none**. Production code, migrations, roles, grants, policies,
Compose files, test files: **untouched**.

## Validation performed (writer, stack-free)

| Command | Outcome |
|---|---|
| `bash -n infra/validation/lib.sh infra/validation/run.sh management/v0/results/G064-s1-private-run-driver.sh` | OK ×3 |
| `docker run --rm --network none -v <worktree>:/repo:ro -w /repo koalaman/shellcheck:v0.11.0 -S warning -x -P SCRIPTDIR infra/validation/run.sh infra/validation/lib.sh` (the B-SHELL warning pass for these two files) | rc 0 (first pass found SC2178/SC2179/SC2128/SC2190/SC2209 from name collisions of new local arrays with existing string/associative variables; renamed `lc_*`/`nok`/`t4_seen`/`iti_*`, quoted `state="head"`; rerun rc 0) |
| shellcheck `-S warning` on the driver | rc 0 (after exporting `DB_LC_CALLER`) |
| Out-of-repo logic test (`/tmp`, not committed): `lib.sh` sourced with psql/compose/docker stubbed by a fake catalog state machine | 23/23: pristine PASS; pending/current_version and dirty-provider-binding negatives named; `<db>_ctl` refused; template created and recorded; `db_fresh` PASS; stale `mig_digest`, held session and cross-family membership refused with **no drop**; `<db>`-level setting accepted as residue and recorded; ledger exported byte-equal to `REPORT_DIR` |
| Driver extraction check | every helper and `it_isolation`/`_iti*` defined after `source lib.sh` + the run.sh block eval; `MIG_RELATIONS=79` |

Not run (by dispatch rule; coordinator-owned): batch 2a real-PostgreSQL proof; `lint`/`ci-controls`/`B-SHELL` through
`run.sh` (no harness run by the writer); every stack-bearing check. Nothing is reported as passing that did not run.

## Coordinator: S1 real-PostgreSQL proof (batch 2a)

From an isolated checkout of `g064-impl` at the pushed S1 head (after `git fetch` + ancestry check and
`bash infra/ci/docker-desktop-mounts.sh`; holds one build slot, one private run slot; no harness lock needed):

```sh
bash management/v0/results/G064-s1-private-run-driver.sh "$PWD" <private evidence dir, e.g. .local/evidence/G064-s1-2a>
```

- Images (local, never pulled): `pgvector/pgvector:0.8.7-pg17`; migrate build bases `golang:1.27.3-alpine`,
  `gcr.io/distroless/static-debian12:nonroot` (the build's `go mod download` needs the module cache/network the
  existing migrate build uses). Ports: **none** (driver override `ports: !reset []`). Project `g064s1-<run id>`.
- Expected: `db_template_create PASS`; `it-isolation.log` with PASS for T-1, T-1a, T-1b, T-1c, T-2, T-3, T-4, T-5a,
  T-5b, T-5c, T-7; `results.tsv` row `B-IT-ISOLATION PASS`; exit 0. Evidence dir: `db-lifecycle.tsv` (pristine,
  template, template-record, refuse rows for T-1b/T-1c/T-3/T-5, fresh rows for T-1c/T-2), redacted logs,
  `<db>_tpl.rec`, `cleanup.log` (left containers/volumes/networks/image all 0).
- Checkpoint observations to report (release record: unexpected goose history/test-list behaviour): the recorded raw
  goose history and its top-version row count (T-1 detail; goose v3's down is expected to delete the version row, so
  one row per version is expected here — unverified until this run); T-4's FATAL texts; T-7 template time/size and
  `db_fresh` durations (PG-010). Any FAIL: preserve the evidence dir and route to the writer.

## Acceptance criteria (S1 contribution only)

| AC | S1 status |
|---|---|
| AC1 inventory | S0 counts re-verified (112 / 52 files / identity 35+39, workos 21, identitytest 0); reset-point wiring S2 |
| AC2 template, copies, fail-safe controls | helpers + T-1..T-5 written; **proof pending batch 2a** |
| AC3-AC5 | not started (S2-S4) |
| AC6 cleanup/ledger | `db-lifecycle.tsv` + per-row export written; driver cleanup receipt; proof pending |
| AC7 | not started |

## Deviations (none architectural)

1. `db_fresh` takes an optional 4th argument RECORD (default the `<db>_tpl` record) so T-1b can pass an altered copy.
2. `db_template_create` runs `db_pristine_verify` itself (the gate cannot be skipped by a caller); the migrations-only
   corroboration is the `DB_TPL_REQUIRE_ROWS` variable (set at S2).
3. Derived names (`<db>_tpl`, `<db>_ctl`) are validated against the 63-byte identifier rule; `<db>` against the
   runner's rule (Plan §3.1 states the 63-byte bound for `<db>_tpl`).
4. T-5a toggles `<db>_runtime`'s current INHERIT attribute and restores it (the role is provisioned NOINHERIT, so the
   Plan's "NOINHERIT → INHERIT" is applied as current → opposite → restore). T-5a/b/c additionally assert that
   `db_fresh` refuses (no drop) while the delta is present; T-4 adds a positive login control per role.
5. Session checks count `backend_type = 'client backend'` only (autovacuum workers are not sessions).
6. The batch 2a driver lives under `management/v0/results/` (allow-listed) instead of `.local/` (writer isolation).

## Budget and history

- B5 ledger: **attempt 1 in progress** (this implementation round; checkpoint after attempt 3). Not reset or extended.
- Correction cycles: 0. Inner loops: shellcheck name-collision fix; logic-test stub-state fix (test-only); tree-hash
  status handling (above).

## Transfer note (checkpoint)

Next writer step after a green batch 2a: S2 — in `suite_migrations` after the production refusals (`run.sh` 7-apply
block) run one read-only `migrate status`, set `DB_TPL_REQUIRE_ROWS="B-MIGRATE-CLI B-MIGRATE-CATALOG"`, call
`db_template_create` and `it_isolation`, then `db_fresh` before the 50 reset points of Plan §4 (T-2 repeat after #50),
the overlay window (§3.5) and the tenancy setup line (§3.4). B-IT-ISOLATION must not be recorded by a suite before its
`required-checks.tsv` row (S5) unless S2 and S5 land together.

## Allow-list check

Script (out of repo): every path in `git diff --name-only 310baf2c` plus untracked files tested with
`grep -qEf management/v0/plans/G064-integration-test-isolation-allowlist.txt` (sha256 `4ee27e22…9627`, 13 lines).
Result before the S1 commit: **changed 4, outside 0** — `infra/validation/lib.sh`, `infra/validation/run.sh`,
`management/v0/results/G064-integration-test-isolation-result.md`, `management/v0/results/G064-s1-private-run-driver.sh`.

## Correction record: S1 batch 2a first attempt (2026-10-21)

- The coordinator's first batch 2a run (evidence `g064-a1/s1-ab4468cd`, driver at `ab4468cd`) failed at the first
  `docker compose run` step: `unknown flag: --no-build`. Docker Compose 2.32.4 `run` has no `--no-build` (it uses the
  existing image unless `--build` is passed). Nothing after that step ran; the EXIT-trap cleanup was verified with 0
  leftovers (containers, volumes, networks, migrate image). The evidence is kept as a failed driver run.
- Fix: driver lines 63 and 65 now use `run --rm --pull never migrate up|status` (the image was just built by the
  preceding build step; `run` neither rebuilds nor pulls). Other driver flags checked against `docker compose
  up|run|down --help` (2.32.4): `up -d --wait --pull never`, `--profile migrate build`, `down -v --remove-orphans` are
  valid. `lib.sh`/`run.sh` contain no `compose ... run --no-build` (the harness's `up --no-build` is valid; unchanged).
- Validation: `bash -n` on the driver. Stack-free only; no Docker run, build or pull.
- Budget: a correction within B5 attempt 1; no new attempt.

## S1 batch 2a outcome (coordinator, 2026-10-21)

PASS on a private instance at `f58b33f2` (evidence `~/.local/state/platform-evidence/g064-a1/s1b-f58b33f2/`:
`results.tsv`, `it-isolation.log`, `status.log`). T-7: template 2114 ms, 9.7 MB; `db_fresh` about 1.6 s each. Goose raw
history 20 rows (version 0 + 19 bound versions), top-version rows 1 (the `00019` down deletes the row; recorded, never
counted). No correction needed.

## S2 record (B5 attempt 1 continues)

Writer: Claude Opus 5.5 (`claude-opus-5-5`), runtime effort not observable (unknown). Checkout: isolated worktree
`.claude/worktrees/agent-<id>`, local branch `g064-impl-s2` from `origin/g064-impl` `f58b33f2` (asserted),
pushed fast-forward to `g064-impl`.

### Implementation (Plan r2 §2.1, §3.4, §3.5, §4, §8 S2)

`infra/validation/run.sh`:

- **Template in the `migrations` lifecycle** (after `B-MIGRATE-CLI`/`B-MIGRATE-CATALOG` are recorded): core-check image
  built first (so the record names its ID), one read-only `compose_migrate run --rm migrate status` →
  `db_template_create` with `DB_TPL_REQUIRE_ROWS="B-MIGRATE-CLI B-MIGRATE-CATALOG"` (it runs `db_pristine_verify`
  itself) → `it_isolation_controls`. No template → every reset point refuses and its check FAILs; B-IT-ISOLATION FAIL.
- **B-IT-ISOLATION split**: `it_isolation` = `it_isolation_controls` + `it_isolation_record` (the S1 driver is
  unchanged in behaviour); the suite records the single row after `signup_checks`, so it includes **T-2r** (T-2 repeated
  on the lifetime after `B-SIGNUP-CATALOG`, Plan §3.3, §4 #50). MinIO-failure branch records FAIL with the reason.
- **`it_drain` / `it_fresh`**: bounded drain of the previous invocation's client sessions (polling only, 50 × 0.2 s;
  a non-zero wait is a `wait` ledger row for PG-010), then `db_fresh`. A refusal writes its reason to the invocation's
  log (so the existing PASS parse fails closed, with the reason in the tail) and the invocation does not run.
- **50 reset points in `migrations`** (Plan §4 #1-#51, by class): **L 21** (migrate-it, the 14 `*-catalog` invocations incl.
  retention and mcp, the 6 `*-migration` invocations — every `core/internal/migrate` invocation except persist pairs);
  **G 6** (group start = persist seed: tenancy, business-profile, policy, rls, obj, audit; verify never recreates);
  **I 23** (identity-catalog, tenancy-model, business-profile-model, policy-model, policy-ordering, rls-it,
  rls-identity-it, rls-ordering, obj-it, obj-it-attack, obj-sign-host, obj-ordering, obj-faults, audit-it,
  retention-it, mcp-it, mcp-facade-it, domains-it, branding-it, onboarding-it, membership-it, enrollment-it, and
  signup-it whose reset is T-2r's `db_fresh`). Recovery drill, PROV-1..6 and `TestMCPE2E` unchanged (no reset point).
- **Identity isolation window (ops, #53; D-2 without fallback)** inside `identity_it`: core-check build; `compose stop
  core core-replica` + drain, zero client sessions asserted; the overlay's own `migrate status` → `db_template_create
  "$D" "$project" "$D" <status>` (its own pristine check; no migrations-suite row required or consulted); the
  unchanged identity `go test`; drain + `db_fresh`; `up -d --wait --no-deps --no-build core core-replica` (no dependency
  such as `identity-setup` is touched). Any refusal → `B-IDENTITY-IT FAIL "isolation window refused …"`, return 1, and
  `variants_identity` records E2E/PROBES/PRIVACY/CONFIG and B-MCP-E2E FAIL ("overlay stopped") and tears the project
  down; a failing test with the window intact stays an ordinary FAIL and the overlay continues. The PASS text appends
  the window facts (MIG_TOP, baseline table count, fingerprint prefix). `mcp_seed` is the first writer afterwards.
- Ordering comments retired (identity-catalog, MCP before the drill, enrollment/signup "template" notes, the identity
  reset dependency).

`core/internal/tenancy/integration_test.go` `setup`: `t.Logf("TENANCY-SETUP version=%s")` of the observed start and a
fail-closed check that it equals the embedded set's top version (derived from `migrations.FS`, no literal) or 5.

Not changed: assertions, counts, PASS parses (other than the two PASS-text additions above), `rls_go` (S3 aliases it),
`required-checks.tsv` (B-IT-ISOLATION is recorded but not yet listed: `B-REQUIRED-CHECKS` and `validate.sh` ignore an
unlisted PASS row), reporter, adapter, docs.

### Validation (writer, stack-free)

| Command | Outcome |
|---|---|
| `bash -n infra/validation/run.sh` | OK |
| `koalaman/shellcheck:v0.11.0 -S warning -x -P SCRIPTDIR run.sh lib.sh` (network none, ro) | rc 0 |
| `golang:1.27.3-alpine gofmt -l -d` on the tenancy test (network none, ro) | clean |
| `go vet -tags integration ./internal/tenancy/` (network none, ro source, ro module cache volume `g058-a1-gomod`, `-mod=readonly GOPROXY=off`, disposable GOCACHE) | rc 0 (compiles; one rename to avoid a `:=` redeclaration found before the run) |
| Out-of-repo stubbed logic test (`/tmp`, not committed): `lib.sh` + the run.sh blocks with psql/Compose/Docker stubbed | **23/23**: `it_fresh` success and refusal (invocation skipped, reason in log); drain wait row / no row / bounded; T-2r pass and failure (next invocation blocked); record PASS/FAIL/no-template; identity window happy path and exact call order (stop → status → template → tests → fresh → up `--no-deps`), refusals on lingering session, stop failure, template, fresh and restart (each rc 1, nothing after it runs), test failure and build failure (rc 0, ordinary FAIL, window completes) |
| Static: 49 `it_fresh` + 1 T-2r = 50 reset points; 0 on any persist-verify line | confirmed by grep |

First-pass fixes (inner loop): test-stub subshell state and an order-parser bug in the logic test (test-only).

### Coordinator: S2 real-PostgreSQL batch

After fetch/ancestry check, `bash infra/ci/docker-desktop-mounts.sh`, from an isolated checkout at the pushed S2 head:

1. **Batch 3 (`migrations` once; harness lock, `PORTS_MIG`; umbrella per D-4):**
   `bash infra/validation/run.sh --base 310baf2c9eb22de7bfc31f38f225b48bd4b8e48c --allow-file management/v0/plans/G064-integration-test-isolation-allowlist.txt migrations`
   (substitute the reconciled staging SHA if staging moved). Images built locally by the suite (core-check, migrate,
   minio-init, validation tools); pinned bases already local; nothing pulled. Expected: every existing migrations ID
   PASS with unchanged counts; `B-IT-ISOLATION PASS` naming T-1..T-5, T-7 and **T-2r**; `db-lifecycle.tsv` with
   pristine/template rows (caller `g001-mig-<run>`, `core_check_image` set) and **50 suite `fresh PASS` rows** (the
   purposes listed above) besides the controls' rows; no suite `refuse` row; `tenancy-model-it.log` and
   `business-profile-model-it.log` contain `TENANCY-SETUP version=19`; `wait` rows (if any) small. Each reset point
   needs a fresh DB by design (the suite creates it).
2. **Batch 4 (identity overlay; Plan §6 — normally S6):** no identity-only selector exists; the window runs inside
   `ops` (`... run.sh --base … --allow-file … ops`, which also runs `variants_ops`; broad — needs the supervisor's
   umbrella disposition) or in a focused private driver still to be written (S6). Expected: `B-IDENTITY-IT PASS` with
   the window facts, a second template block (caller `identity-overlay`), one `fresh PASS identity-it-after`, then
   `B-IDENTITY-E2E` (M-LIST seeded old-host row) PASS.

Any FAIL: preserve the evidence directory and route to the writer.

### Deviations (S2; none architectural)

1. `db_pristine_verify` is invoked through `db_template_create` (S1 deviation 2), not as a separate preceding call, in
   both template callers; the pristine row is recorded once.
2. `it_drain` (bounded, polling-only) precedes every `db_fresh`: client backends of a just-exited `--rm` container may
   linger milliseconds; nothing is terminated and `db_fresh` still refuses a remaining session (T-3 unchanged).
3. The migrations suite builds the core-check image just before the template so the record carries its ID (the later
   builds are cache hits).
4. `B-IT-ISOLATION` is recorded after T-2r (after `signup_checks`) instead of right after the template.
5. Window restart uses `up --no-deps --no-build` so `identity-setup` (a completed one-shot dependency) is never re-run.

### Allow-list (S2)

`/tmp` script: every path in `git diff --name-only <base>` plus untracked, tested with `grep -qEf` against the
allow-list (sha256 `4ee27e22…9627`). Against release base `310baf2c`: **changed 5, outside 0**; against S2 base
`f58b33f2`: **changed 3, outside 0** (`infra/validation/run.sh`, `core/internal/tenancy/integration_test.go`, this Result).

### Budget

B5 attempt 1 continues (no new attempt). Correction cycles 0.

### Transfer note (S2 checkpoint; writer context past 150k, completion-imminent exception recorded)

Next writer: S3 — `go_it` with `-json -list` `(package, test)` expectations, `gotestreport` + tests, `rls_go` →
alias, `identity_it`/`rd_go`/`obj_sign_host`/`mcp_e2e` through `go_it`, per-invocation export, `go_tests_finalize`
in the EXIT trap, T-6/T-8, `validate.sh` adapter + GT1/GT2. `it_fresh` stays in front of each `go_it` call at the 50
points; `identity_it`'s window wraps its `go_it` call. A focused identity-window driver is needed for batch 4 (S6).

## S3 record, part 1 (reporter, `go_it`, completeness rules; B5 attempt 1 continues)

Writer: Claude Opus 5.5 (`claude-opus-5-5`), runtime effort not observable (unknown). Checkout: isolated worktree
`.claude/worktrees/agent-<id>`, local branch `g064-impl-s3` from `origin/g064-impl` `4f9e6a48` (asserted),
pushed fast-forward to `g064-impl`. Staging `42a2bf1e` (management-only) not rebased, per dispatch. **Context
checkpoint reached (150k policy): S3 is split; part 2 (the `validate.sh` adapter and GT1/GT2) is listed under
"Remaining" below and is not started.**

### Implemented (Plan r2 §5.1-§5.3, §6 T-6/T-8, §7)

- **Converter** `infra/validation/tools/internal/gotest/report.go` + `cmd/gotestreport/main.go` (stdlib only; built into
  the validation tools image: `infra/docker/validation-tools.Dockerfile` build list and runtime COPY gain
  `gotestreport`). Input: the invocation stream (`go test -json` events, non-JSON lines verbatim, then the expectation
  block `GO-IT-LIST <event>` per `go test -json -list` event and `GO-IT-LIST-EXIT <n>`). Output (tar on stdout): `NAME.log`,
  `NAME.tsv` (invocation / package / test rows, 15 columns, header `GO_TESTS_HEADER`), `NAME.junit.xml` (no text nodes,
  time only when measured), `NAME.status`. Exit 0 pass, 1 fail/interrupted (archive complete), 2 usage/read error.
  Every §5.2 rule is a named finding: `package-unterminated`, `package-duplicate-terminal`, `package-failed`,
  `package-skipped` (skip accepted only for an expected `[no test files]` package), `package-pass-with-exit`,
  `build-failed`, `zero-tests`, `omitted`, `undeclared-test`, `undeclared-package`, `package-absent`, `unfinished`,
  `duplicate-terminal`, `contradictory`, `skipped` (any depth, as every existing parser), `malformed-json-line`,
  `truncated-final-line`, `expectation-missing|-exit|-empty|-failed|-unterminated|-duplicate`, `exit:<rc>`,
  `interrupted`. Expected set = the `Test*` names of each `-list` package (keyed by the event's `Package`; S3 check of
  the Plan's open question: **`-json -list` events do carry `Package`**, verified offline with go1.27.3, so the
  per-package fallback is not needed).
- **Reconstruction fidelity (finding, resolved inside the converter):** plain concatenation of `Output` is *not* the
  `go test -v` text — in test2json mode a subtest's `--- PASS/FAIL` line is printed unindented at the subtest's end, while
  `-v` buffers it after its parent's line, indented four spaces per level, and prints packages contiguously. Two
  existing parsers depend on the `-v` layout: `identity_it`'s `^--- PASS: ` top-level count (in its PASS text) and
  `obj_fail_codes`'s `^    --- FAIL: TestObjects/…`. The converter therefore re-lays results out as `-v` does (buffer per
  parent, indent, package-contiguous in start order). Offline comparison against real go1.27.3 output of a probe module
  (nested, slash-named and parallel subtests, `fmt.Println` lines, a `[no test files]` package, a failing package with
  a skip): **byte-equal** after normalizing durations, except two documented, parser-irrelevant differences: the `=== NAME`
  attribution lines (consumed by test2json, not re-synthesized) and the go command's final bare `FAIL` line of a failing
  run. No harness parser was edited for layout; the 72 `--- PASS` greps, `it_count`, `rd_gopass`, the identity `^ok`
  rule and `obj_fail_codes` read the same layout as before.
- **`go_it NAME PROJECT IMAGE ITENV LOG PACKAGE RUN [docker args...]`** (`run.sh`): `go_it_begin` registers the
  invocation (`$WORK/go-tests/started.tsv`, provenance file: run_id, check, suite, tested_tree = `impl_tree_hash` of the
  suite's copy, head_sha, tree=head|dirty, base, image_id, package, tag, selector, timeout, full command, last
  `db-lifecycle.tsv` row) and requires the tools image (no per-test evidence → the invocation does not run, FAIL);
  one container runs `GO_IT_SH`: canaries file (as `rls_go`), `go test -tags T -json -list RUN PKG`, `go test -tags T
  -count=1 -timeout X -json -run RUN PKG; rc=$?`, the RLS-ACCESSLOG sed, the expectation block, `exit $rc`. Stream →
  `LOG.events`; `go_it_convert` runs `gotestreport` through `vtools -i` (stdin/stdout, network none: **no bind mount at
  all**, stricter than the Plan's fixed-directory pattern) and writes LOG; returns the container's non-zero status
  unchanged, else 1 on any finding, else 0. Without a converter LOG keeps the raw stream (no parser accepts it) and a FAIL
  row names `converter-unavailable`. `GO_IT_TAG` (default integration), `GO_IT_TIMEOUT` (default 10m = go test's own
  default, so former direct sites keep their timeout), `GO_IT_NETWORK`, `GO_IT_NO_EXPORT`.
- **`rls_go`** is a compatibility alias (D-6): same name/signature, `GO_IT_TIMEOUT` 30m, the three RLS `-e` arguments.
- **All invocations through `go_it`** (Plan §4 #1-16, 19, 52-54): 17 direct `docker_run_owned … go test` sites
  converted mechanically (tenancy 5, business profile 5, policy 6, identity catalog 1; the two 30m sites keep 30m), plus
  `migrate-it` (its go test status now joins the PASS decision, previously ignored), `rls_provisioned_test`
  (names `rls-provisioned-su|adm|v5` from its log), `rd_go` (`rd-seed|post|verify|repair1|repair2`, 20m, own network),
  `obj_sign_host` (detached: `go_it_begin` + `GO_IT_SH` in the detached container + `go_it_convert` of its `docker logs`;
  the readiness grep also accepts the `-json` output-event form of `OBJ-SIGN-HOST-READY`), `identity_it` (`-run .` =
  the default selection; its status now joins the PASS decision, previously `|| true`) and `mcp_e2e`
  (`GO_IT_TAG=mcpe2e`, network `container:$web`, export deferred until its `MCP-OBSERVED` bearer values are in the
  secret registry; the published copy shows `MCP-OBSERVED [withheld]` so line ranges stay exact). Every `rls_go`
  caller is unchanged. Out of scope and unchanged: `B-STAGING-IDENTITY` (`TestStagingContract`) and `B-FEATURE-CORE`.
- **`identity_it` per-package rule:** `identity_it_packages` reads the package rows: every selected package must be
  `pass` with passed = expected ≥ 1 and no failed/unfinished, or `skip-no-test-files` with 0 expected (derived by
  `-list`, not pinned). The PASS text keeps its existing wording and appends `G064 per-package expectations (go test
  -list, passed/expected): core/internal/identity N/N, …/workos N/N, …/identitytest no test files`; the unchanged
  `-ge 10`/`^ok identity`/no FAIL/SKIP rules stay.
- **Export and inventory** (`lib.sh`): `redact_log` factored from `export_diagnostics` (behaviour unchanged);
  `go_tests_export` appends rows to `$WORK/go-tests.tsv`, regenerates `go-tests-junit.xml`, publishes both,
  `db-lifecycle.tsv` and the redacted, 1 MiB-bounded `go-tests/NAME.log` to `REPORT_DIR` atomically (tmp + mv) as each
  invocation ends; `.events` never leave `$WORK` and are not `*.log` (excluded from `export_diagnostics` by pattern).
  `go_tests_declare` + `go_it_declare` write `go-tests-expected.tsv` from the static `GO_IT_migrations` (64 entries;
  `obj-sign-host` omitted when `PROBE_MODE=none`, where B-OBJ-IT is UNMET anyway) and `GO_IT_ops` (2) arrays at suite
  start (`NAME|CHECK`, the check also feeds provenance).
- **Interruption:** EXIT trap = `RUN_STATUS=$?; bg_stop_all; go_tests_finalize …; export_diagnostics …; cleanup`.
  `go_tests_finalize` converts every registered invocation without its `.done` marker with `--interrupted` (rc 255 =
  unknown), keeping observed terminal rows and marking started tests `unfinished`, and exports it.
- **B-IT-ISOLATION T-6** (`_iti_t6`, in the controls after T-5): `gotestreport selftest` in the tools image (every
  fixture: the passing B-IDENTITY-IT-shaped stream with exact text and per-package rows, and one stream per fail-closed
  class), plus the CLI contract on a passing (exit 0, text byte-equal, `identitytest=skip-no-test-files`) and a failing
  (`omitted-expected`, exit 1) fixture. **T-8** (`_iti_t8`, when the row is recorded after T-2r): the four report files
  present (REPORT_DIR, else the work copies) and every converted invocation so far with its row and `go-tests/NAME.log`.

Files: `infra/validation/run.sh`, `infra/validation/lib.sh`, `infra/docker/validation-tools.Dockerfile`, new
`infra/validation/tools/internal/gotest/{report.go,fixtures.go,report_test.go}`, new
`infra/validation/tools/cmd/gotestreport/main.go`, this Result. Not changed: any test file, assertion, count, skip,
production code, `junit.sh` (authoritative JUnit untouched, D-5), `required-checks.tsv`, `validate.sh`,
`validate-controls.sh`, docs.

### Validation (writer, stack-free; exact outcomes)

| Command | Outcome |
|---|---|
| `bash -n infra/validation/run.sh infra/validation/lib.sh` | OK |
| `koalaman/shellcheck:v0.11.0 -S warning -x -P SCRIPTDIR run.sh lib.sh` (network none, ro) | rc 0 (first pass: SC2178/SC2128 from a local `envf` array colliding with existing string variables → renamed `go_envf`; SC2209 quoted; SC2034 on the two nameref-read arrays disabled with reason; rerun rc 0) |
| `golang:1.27.3-alpine`, network none, ro source, `GOFLAGS=-mod=readonly GOPROXY=off GOTOOLCHAIN=local`, disposable GOCACHE: `gofmt -l`, `go vet`, `go test -count=1 ./internal/gotest ./cmd/gotestreport` | gofmt clean; vet rc 0; tests `ok` (TestFixtures: 25 fixtures, TestRowsAndProvenance, TestInterruptedRows) |
| Same container: `go build ./cmd/gotestreport` (static binary for the host checks below, outside the repo) | rc 0 |
| Real-stream comparison (out of repo): offline probe module, go1.27.3 `-json`/`-list`/`-v` captured, converted by the binary | multi-package pass stream: status pass, text **EQUAL** to `-v`; failing package stream: status fail (`package-failed`, `skipped`, `exit:1`), text **EQUAL** (both modulo durations, `=== NAME`, final bare `FAIL`) |
| Stubbed `go_it` dry-run (out of repo, `/tmp`): `lib.sh` + the `run.sh` go_it block sourced; docker/vtools/ensure_tools stubbed; the "container" runs `GO_IT_SH` locally with a fake `go` replaying converter fixtures | **18/18**: all-pass (go_it 0, row pass, LOG = fixture text); `[no test files]` package row `skip-no-test-files`; identity per-package rule (`… 2/2, … 1/1, … no test files`); export to REPORT_DIR with the planted secret redacted; provenance fields; one failing test (status 3 propagated unchanged); package without terminal, duplicate result, omitted expected test, undeclared test (each go_it 1 with rc 0); pass events with exit 1 (1 propagated); `rls_go` alias (30m, RLS env); repeated name refused; interruption finalized (`interrupted` row, `unfinished` test row, exported); 9 invocations exported exactly once; no `.events` published; no converter → FAIL `converter-unavailable` |
| Allow-list (`git diff --name-only 310baf2c` + untracked, `grep -vEf` the 13-line allow-list) | **changed 10, outside 0** |
| U+2028/U+2029 (`grep -c -P '\x{2028}\|\x{2029}'`) on every changed file and this Result | 0 each |

Not run (by dispatch rule; coordinator-owned): any stack, database, harness suite or image build. **The real-PostgreSQL
exercise of the S2 wiring and of the S3 wiring is not yet executed**: per coordinator disposition the first real run of
S2 is the S4 campaign / S6 batch 3 (Plan §8 has no S2 real-PostgreSQL batch; the single D-4 `migrations` umbrella run
stays reserved). Every claim above about `go_it`, the reset points, the identity window, T-6/T-8 and the exports against
real PostgreSQL, the real core-check image and the real tools image build remains "not yet executed".

### Deviations (S3 part 1; none architectural)

1. The expectation block carries the raw `-list` events (`GO-IT-LIST <event>`) and the converter derives the `(package,
   test)` pairs (Go JSON parsing instead of in-container shell); the reconstructed LOG renders them in the Plan's
   `GO-IT-EXPECTED package<TAB>test` / `package<TAB>-` form. Order as the Plan: `-list` before the run (a `-list` run
   executes `TestMain`; identity/signup `TestMain` migrate `Up` (no-op at the top version) and clear fixture slugs that
   do not exist on a fresh copy, so state is unchanged).
2. The converter re-lays results out in `-v` form (finding above); the Plan's "concatenated Output fields" alone would
   have broken two parsers.
3. `gotestreport` reads stdin and writes a tar to stdout (no bind mount) instead of the RLS_FIXED fixed-directory pattern.
4. Converter fixtures live in `fixtures.go` (a non-test file, allow-listed `.go`) so the same table serves the unit test
   and T-6 (`gotestreport selftest`); no testdata files (not allow-listed).
5. `migrate-it` and `identity_it` now require the go test status (the Plan's `rc == 0 && converter ok && counts` rule);
   their PASS descriptions are unchanged except B-IDENTITY-IT's appended per-package numbers (Plan §5.2).
6. `obj_sign_host`'s readiness grep accepts the `-json` event form (mechanical parity: the line is now an output event).
7. `rd-repair1`/`rd-repair2` are attributed to B-MAINTENANCE-DRILL in the inventory's provenance (check names are
   provenance only; the drill's records are unchanged).
8. The S1 batch 2a driver calls `it_isolation`, which now includes T-6 (needs the tools image) and T-8 (needs converted
   invocations); a rerun of that one-off driver would report T-6/T-8 FAIL. It is not rerun; S4's driver will need both.

### Remaining (exact; S3 part 2, not started)

1. `infra/ci/validate.sh`: `STAGE_REPORT` += `go-tests.tsv go-tests-expected.tsv go-tests-junit.xml db-lifecycle.tsv`;
   `do_stage` stages `go-tests/` like `diagnostics/` (name pattern `^[A-Za-z0-9_.-]{1,120}$`, regular files only); the
   suite-aware `complete` rules (i)-(iv) of Plan §5.4 (when `migrations` or `ops` executed: four files present/regular/
   non-empty/well-formed; exactly one `invocation` row with `pass|fail` and an `rc` per declared invocation of the
   executed suites; `interrupted`/`unfinished`/duplicate/missing/undeclared → finding, return 14; per-test rows
   consistent with counts and every `log_file` present under `go-tests/`; `results.tsv`/required-ID rules unchanged).
   Row format: `GO_TESTS_HEADER` (15 columns; invocation `status` col 5, `rc` col 6, `log_file` col 12).
2. `infra/ci/validate-controls.sh` group D: GT1 (missing report → 14, nothing published) and GT2 (one `fail` + one
   `interrupted` invocation, planted generated secret in `go-tests/<n>.log` withheld by the real scanner, FAIL row
   reported, → 14); `run.sh` `B-CI-CONTROLS` expected list/count +2 (mechanical).
3. Then the stack-free checks for part 2 (`bash -n`, shellcheck, a local `validate-controls.sh` GT1/GT2 run if it is
   stack-free) and the S3 checkpoint handback. S4 (95-selector campaign), S5 (required-checks row, docs, PG-010) and S6
   batches stay as Plan §8.

### Budget

B5 attempt 1 continues (no new attempt). Correction cycles 0. Inner loops: shellcheck rename; stub-test harness fixes
(test-only: stub argument shift, an unset `RLS_CANARIES` in the stub, a temp-dir path).

## S3 record, part 2 (`validate.sh` adapter, GT1/GT2; B5 attempt 1 continues)

Writer: Claude Opus 5.5 (`claude-opus-5-5`), runtime effort not observable (unknown). Checkout: isolated worktree
`.claude/worktrees/agent-<id>`, local branch `g064-impl-s3b` from `origin/g064-impl` `592c2e5e` (asserted
before editing), pushed fast-forward to `g064-impl`. Not rebased onto staging (staging's G063 changes to `validate.sh`
and `validate-controls.sh` are reconciled by the coordinator at S8, per dispatch). Stack-free only.

### Implemented (Plan r2 §5.4, §6 GT1/GT2, §7)

- **`infra/ci/validate.sh`:** `STAGE_REPORT` += `go-tests.tsv go-tests-expected.tsv go-tests-junit.xml
  db-lifecycle.tsv`; `do_stage` stages `go-tests/` with the `diagnostics/` loop (same name pattern
  `^[A-Za-z0-9_.-]{1,120}$`, regular non-link files only; one loop over both directories). Both scans already cover
  every file under `raw/` and the staged candidate, so `go-tests/` is gated by the literal-secret and gitleaks scans
  with no scanner change. New `go_tests_complete REP SET` (Bash + one awk program, no Docker): when the executed set
  holds `migrations` or `ops`: **(i)** the four files exist, are regular, non-link, non-empty and well-formed
  (`go-tests.tsv` = exact `GO_TESTS_HEADER` + 15 columns, kind `invocation|package|test`, safe invocation name,
  `log_file` = `go-tests/<safe name>`; `go-tests-expected.tsv` = `suite<TAB>invocation` header + 2 columns, no name
  declared twice; `go-tests-junit.xml` = the closed line grammar the harness writes (declaration, `<testsuites>`, one
  `<testsuite ...>` ... `</testsuite>` block per invocation holding `<testcase .../>` or
  `<testcase ...><failure|error|skipped .../></testcase>` lines, `</testsuites>` last); `db-lifecycle.tsv` = its
  header + at least 5 columns with a UTC millisecond timestamp); **(ii)** every executed suite declared at least one
  invocation; every invocation declared for an executed suite has exactly one `invocation` row with status
  `pass|fail` and a numeric `rc`; `interrupted`/`unfinished`/any other status, duplicate, missing, undeclared,
  other-suite invocations and orphan rows are named findings; **(iii)** each invocation's
  expected/passed/failed/unfinished equal its rows (sum of package `expected`; top-level `parent=` test rows
  `pass`/`fail`; test rows `unfinished` at any depth: the converter's own counting), one JUnit suite per invocation
  row, and every `log_file` a regular non-link file under a non-link `go-tests/`; **(iv)** `do_complete` calls it right
  after the executed set is known, keeps its status and then applies every existing rule unchanged (results.tsv,
  required IDs, browser/contract reports, leftovers still decide and still return their own 14); its two `return 0`
  exits return the kept status. Every finding is said (`complete: go-tests: ...`, at most 20, then a count) and the
  status is 14. When neither suite executed the files are not required. The authoritative `junit.xml` (`junit.sh`) is
  untouched (D-5).
- **`infra/ci/validate-controls.sh`:** the stub harness writes the four reports and `go-tests/stub-<suite>-it.log`
  (one passing invocation per executed `migrations`/`ops` suite; fixture parity, otherwise every full/integration/gate
  control would now fail completeness) and `FULL_SET` (V1's exact published set) gains `db-lifecycle.tsv go-tests
  go-tests/stub-migrations-it.log go-tests/stub-ops-it.log go-tests-expected.tsv go-tests-junit.xml go-tests.tsv`.
  New stub mode `go-tests-fail-interrupted`. **GT1** (`STUB_RM_REPORT=go-tests.tsv`, full CI run, every row PASS,
  harness 0): exit 14, `complete` 14, `complete: go-tests: go-tests.tsv is missing`, published set = `FULL_SET`
  without `go-tests.tsv`. **GT2** (`STUB_HARNESS=go-tests-fail-interrupted STUB_FAIL_ID=B-MIGRATE-IT`): `stub-mig-fail`
  (`fail`, rc 1, a failing test row, its log holding the planted generated secret) + `stub-mig-int` (`interrupted`,
  rc 255, one unfinished test): `complete` 14 naming `invocation stub-mig-int is interrupted (rc 255)` (and not the
  complete failed invocation), `required check B-MIGRATE-IT is not PASS` reported, `raw_scan` 15 (`a generated secret
  was found`), exit 15, minimal `WITHHELD` publication, canary absent. Both in group D after J8; header comment updated.
- **`infra/validation/run.sh`: no edit needed.** `static_ci_controls` derives the `B-CI-CONTROLS` expected list from
  `validate-controls.sh --list` and the count from the PASS lines (`n -ge 80`, `got == want`), so adding `"D GT1"
  "D GT2"` to `CONTROLS` is the whole mechanical change (`--list`: 524 to 526 entries); its PASS description is
  unchanged.

Files: `infra/ci/validate.sh`, `infra/ci/validate-controls.sh`, this Result. Not changed: `run.sh`, `lib.sh`,
`junit.sh`, `required-checks.tsv`, any test, assertion, count, skip or production code, docs.

### Validation (writer, stack-free; exact outcomes)

| Command | Outcome |
|---|---|
| `bash -n infra/ci/validate.sh infra/ci/validate-controls.sh` | OK (both) |
| `koalaman/shellcheck:v0.11.0 -S warning -x -P SCRIPTDIR validate.sh validate-controls.sh` (`--pull never --network none`, ro) | first pass rc 1: SC2178/SC2179/SC2128, a local `suites` string colliding with an existing `suites` array, which also exposed a real slip (the awk `-v suites=` read that name); renamed `gsuites` and fixed; rerun **rc 0** |
| Same image, `-S info --include=SC2251` (the B-SHELL second pass) | rc 0 |
| Fixture test of the adapter (out of repo, `/tmp`): `GO_TESTS_HDR` through `do_complete` and `suite_set` extracted from `validate.sh` and sourced (a global `suites` array planted), synthetic report directories, `say` captured | **63/63** with gawk and **63/63** with mawk (Debian's default awk): (i) valid (2), each file missing/empty/symlink (12), directory, header/column/kind/name/expected-row faults (6), JUnit unclosed/declaration/nested/text/trailing (5) and a failure testcase accepted, db-lifecycle header/timestamp/columns (3); (ii) `fail` accepted, interrupted, unfinished, missing, duplicate, undeclared, other suite, declared twice, no rc, executed suite undeclared, ops-only accepted, orphan rows, more than 20 findings capped; (iii) passed/failed/unfinished/expected mismatches, subtest fail not counted (accepted), log missing/symlink/outside/directory symlink, JUnit suite missing/extra, converter-unavailable row accepted; (iv) neither suite gives 0 without files and with garbage; `do_complete` wiring: all pass 0, FAIL row only 14, go-tests only 14, both findings reported, missing report 14, missing required ID still decides |
| Stub parity (out of repo): the stub's `go_tests_reports` extracted from `validate-controls.sh`, run per profile, checked by the real `go_tests_complete` | **5/5**: full, integration, fast give 0 (fast writes nothing); GT1 shape gives 14 with exactly `go-tests.tsv is missing`; GT2 shape gives 14 with exactly the `stub-mig-int` finding; the full report entries equal the names added to `FULL_SET`; the canary is in `stub-mig-fail.log` |
| `bash infra/ci/validate-controls.sh --list` | 526 entries, `D GT1`, `D GT2` present |
| Allow-list (`git diff --name-only 310baf2c` + untracked, `grep -vEf` the 13-line allow-list) | **changed 12, outside 0** |
| U+2028/U+2029 (`grep -c -P '\x{2028}\|\x{2029}'`) on both changed scripts and this Result | 0 each |

**Not run (Docker-dependent; coordinator-owned):** GT1/GT2 themselves and the regression of the existing controls
whose stub fixture changed. Every validate-controls control runs the real `validate.sh` through a docker shim to the
real daemon (gitleaks and yq images, owned-resource cleanup), so none is stack-free. Commands for the coordinator (no
stack, database or image build; needs the pinned gitleaks and yq images present locally):
`bash infra/ci/validate-controls.sh --only GT1,GT2,V1,V3,V13e,VA9,VP7,J8` (the new controls plus the controls that
consume the changed stub/`FULL_SET`: full publication set, single-finding count, gate remainder, rows-only), then the
whole `B-CI-CONTROLS` selector in S6 batch 1 (`bash infra/validation/run.sh --base HEAD --allow-file <scope> lint
lint-base ci-controls`). **Nothing from S2/S3 has yet run against real PostgreSQL, a real core-check image or a real
tools image build**; that remains S4/S6.

### Deviations (S3 part 2; none architectural)

1. GT1 "nothing published" is read as "no go-tests report published, nothing stands in for it": the adapter publishes
   the scanned evidence set on any completeness failure (existing V3/VP4 semantics) and Plan §5.4 lists no
   publication change, so GT1 asserts `published=full` equal to `FULL_SET` minus `go-tests.tsv`. Withholding a whole
   publication on a completeness failure would be a new adapter rule needing review.
2. GT2's exit is 15, not 14: the planted unredacted secret makes the raw scan withhold the evidence, and the existing
   final-status ranking puts 15 above 14; the `complete` phase records 14 with the named finding. This is the Plan's
   "complete 14 ... and the raw scan ... withholds the planted secret" combination under the unchanged ranking.
3. The stub harness gained the four reports for every executed `migrations`/`ops` run and `FULL_SET` the seven new
   published entries (fixture parity under standing authority: without it every full, integration and gate control
   would fail completeness for missing per-test evidence). No existing assertion was changed.
4. `run.sh` is not edited: the `B-CI-CONTROLS` list and count are derived (above).
5. The adapter requires `db-lifecycle.tsv` whenever `ops` executed (Plan §5.4). The real `ops` suite writes it only
   through the identity window's ledger rows; the coordinator's first `ops`/integration run must confirm it is always
   present (else complete 14, fail-closed).

### Remaining (exact; S4 onward, not started)

S4 the 95 §4.1 direct selectors + AC3 reversed pairs/repeat/shuffle on the private instance (batch 2b; needs the tools
image and core-check builds; T-6/T-8 in the driver); S5 `required-checks.tsv` row `B-IT-ISOLATION migrations` and
header counts, docs, PG-010; S6 batches 1, 3, 4 (including the Docker-dependent GT1/GT2 and existing-control
regression above); S7 self-review; S8 staging reconciliation (including G063's `validate.sh`/`validate-controls.sh`
changes) and handback.

### Budget

B5 attempt 1 continues (no new attempt). Correction cycles 0. Inner loops: shellcheck name collision (and the awk
argument it exposed); one test-script mutation fix (test-only).

## S4 record (direct-selector campaign driver, batch 2b; B5 attempt 1 continues)

Writer: Claude Opus 5.5 (`claude-opus-5-5`), runtime effort not observable from inside the agent (unknown). Checkout:
isolated worktree `.claude/worktrees/agent-<id>`, local branch `g064-impl-s4` from `origin/g064-impl`
`45a263fd` (asserted before editing), pushed fast-forward to `g064-impl`. Not rebased onto staging. Stack-free only: no
Compose stack, database, harness suite, image build or pull was run by the writer.

### Implemented (Plan r2 §4.1, §6 batch 2b and AC3 row, §8 S4)

One disposable task-evidence driver, `management/v0/results/G064-s4-private-run-driver.sh`, in the S1 driver's pattern
(private Compose project, `--pull never`, no published host port, one build at a time, EXIT-trap cleanup with a
leftover receipt, generated-secret redaction, evidence directory output). No harness, test or validate script changed:
the Plan's S4 text needs none. The driver sources `lib.sh` and evaluates, at top level, the `run.sh` blocks it calls
(tools image helpers; `obs_inputs`/`new_canary`; `catalog_p` .. `it_isolation*`, `it_drain`, `it_fresh`; `rls_go`, the
`go_it` block, `go_it_publish`; `obj_minio`, `obj_env`), and stops (exit 2) if any called function or `GO_IT_SH`/
`MIG_RELATIONS` is missing. It calls the real `it_fresh`/`db_fresh`, `db_template_create`, `go_it_begin`, `go_it_convert`,
`go_it`, `rls_go`, the isolation controls and `go_tests_finalize`; it reimplements none of them.

- **P0 (no Docker):** copy of the checkout; the campaign selectors **derived from that copy's source**: every top-level
  `func TestX(t *testing.T)` of a `core/internal` test file whose build line is exactly `//go:build integration`, minus
  the G-class groups (`*PersistSeed`/`*PersistVerify`, `TestRecoverySeed|Post|Verify|Repair`; `TestMCPE2E` is
  `mcpe2e`-tagged) and the own-instance `TestRLSProvisioned`; class L = `core/internal/migrate`, else I. Fail closed
  (exit 2, before any container) unless tagged 112, excluded 17, **95 selectors, I 70 / L 25** and the §4.1 package
  numbers (identity 35, tenancy 2, access 25, mcp 5, signup 3, migrate 25) hold and every invocation name is unique.
  The declared inventory (`go-tests-expected.tsv`, suite `g064-s4`) lists the 95 + 12 invocation names before any runs.
- **P1-P4 (private instance):** PostgreSQL; migrate image build, `migrate up`, the source-bound `migrate status`;
  core-check image build (before the template so its record names it); `db_template_create` (its own
  `db_pristine_verify`; no migrations-suite row required); the validation tools image (`gotestreport`) built under a
  task-owned tag and preset so `ensure_tools` reuses it; `minio` and `minio-init` images built one after the other, then
  the harness's `obj_minio`/`obj_env`; `it.env`, the identity env and `OBJ_ENV` exactly as `run.sh` writes them,
  `RLS_CANARIES` (three registered canaries) and `MCP_INVALIDATE_SQL_B64`. A no-published-port check after postgres,
  after MinIO and after the campaign (any `->` mapping on a project or run-labelled container is FAIL).
- **P5:** `it_isolation_controls` (T-1, T-1a, T-1b, T-1c, T-2..T-5, **T-6**, T-7).
- **P6 campaign (95):** per selector, in package order then name: `it_fresh` (bounded drain, then `db_fresh` with its
  zero-session precondition and postconditions; a refusal means the selector does not run and is FAIL), then one
  invocation `ac3-<pkg>-<Name>` with `-run '^Name$'` and the §4 row's environment (a rule table mirroring the `run.sh`
  sites: 7 migrate `go_it` sites on `it.env` at 10m; RLS sites on `it.env` with `rls_go`'s canaries/access log at 30m;
  identity on the identity env at 10m; the four `TENANCY_*` sites; object/audit/retention/MCP/domain/branding/
  onboarding/membership/enrollment/signup sites on `OBJ_ENV` at 30m; MCP adds `MCP_LOG_FILE`/`MCP_INVALIDATE_SQL_B64`;
  sign-host 15m). §4.1 cost bound: one owned, long-lived core-check container per package (`docker exec` per selector;
  the Go build cache in `/root/.cache` is reused; `/tmp` is emptied before each; the `go test` process exits between
  selectors). The invocation goes through `go_it_begin` + the harness's unchanged `GO_IT_SH` + `go_it_convert` (thin
  wrapper `s4_go`, same return rule as `go_it`). PASS needs `s4_go` 0 **and** the converted rows: one invocation row
  `pass`, rc 0, no failed/unfinished, **expected exactly 1, selected package exactly the test's package, exactly one
  top-level test row `(package, Name)` passed**; tenancy selectors additionally `TENANCY-SETUP version=` the derived top
  version. `S4_MODE=run` uses one `--rm` container per selector instead (same wrapper and rule).
- **P7 AC3 (12 invocations, real `go_it`/`rls_go`, each on a fresh copy):** pairs (a) `TestMigrationLifecycle` then
  `TestIdentityCatalog`, (b) `TestBusinessProfileModel` then `TestTenancyModel` (both must log the top version), (c)
  `mcp-it` (with `TestMCPRecovery`) then `domains-it`, (d) `signup-catalog` then `rls-catalog` with an explicit
  `db_shared_catalog` equality check between (plus `db_fresh`'s own snapshot comparison); repeat: `TestAccessPolicy`
  twice; shuffle: `rls-it` (#21) and `mcp-it` (#38) with `GOFLAGS=-trimpath -shuffle=on`, the seed read from the test
  binary's `-test.shuffle N` line (fail closed if absent).
- **P8:** T-7 again (PG-010 line with every campaign `db_fresh` duration), then `it_isolation_record` (T-8 and the one
  `B-IT-ISOLATION` row).

Exit 0 only when the 95 selectors PASS (I 70, L 25), AC3 12/12, `B-IT-ISOLATION` PASS and no published port; 1 on a
test/control failure (the batch still finishes collecting every failure); 2 on a preflight/infrastructure stop; 4 when
everything passed but cleanup left an owned resource; 130/143 on interruption (invocations finalized `interrupted`).

### Coordinator: the real campaign (run once, private instance)

After fetch/ancestry check and `bash infra/ci/docker-desktop-mounts.sh`, from an isolated checkout at the pushed S4 head,
holding one slot of the two-run pool and the build slot for P2-P4 (five sequential builds):

`bash management/v0/results/G064-s4-private-run-driver.sh <checkout at the S4 head> <absent or empty private evidence dir>`

- Resources: Compose project `g064s4-<UTC stamp>-<pid>` (postgres, minio, their volumes, the `data` network); images
  `g064s4-<run>-{migrate,core-check,minio,minio-init,vtools}`; per-package containers `g064s4-<run>-gocheck-<pkg>`,
  short-lived go/tools containers labelled `g001.run=<run>`; zero published ports; `--pull never` for Compose up/run
  (builds use local base images; Dockerfile `go mod download` steps behave as in every harness build); no
  `local-harness.lock`; `COMPOSE_PARALLEL_LIMIT=1`.
- Expected duration: not measured (first real run); estimated 1-3 hours, dominated by the 95 test executions plus one
  test-binary compile per package (six) and 12 AC3 compiles; `s4-phases.tsv` records each phase.
- Evidence layout (the evidence directory): `s4-summary.tsv` (11 columns: section, name, class, package, test, status,
  rc, fresh, fresh_ms, go_ms, detail; sections `count` derivation/campaign/class-I/class-L/ac3, `campaign` 95 rows,
  `ac3` 12 rows with seeds/setup versions/shared-catalog result, `control` B-IT-ISOLATION and no-published-ports,
  `result`), `s4-phases.tsv`, `s4-selectors.tsv`, `go-tests.tsv`, `go-tests-expected.tsv`, `go-tests-junit.xml`,
  `go-tests/<name>.log` (redacted, one per invocation), `db-lifecycle.tsv`, `results.tsv`, `driver-logs/` (redacted
  driver logs, template record), `cleanup.log` (leftover receipt).
- Any FAIL: keep the evidence directory and route to the writer with `s4-summary.tsv` and the failing
  `go-tests/<name>.log`.

### Validation (writer, stack-free; exact outcomes)

| Command | Outcome |
|---|---|
| `bash -n` on the driver | OK |
| `koalaman/shellcheck:v0.11.0 -S warning` (`--pull never --network none`, ro) on the driver | first pass rc 1: SC2034 on `VT_OK`, `GO_IT_CHECK`, the dynamic-scope `itlog`/`tpl` (all read by the evaluated harness code) → targeted disables with reasons; rerun **rc 0**; `-S info --include=SC2251` rc 0 |
| Same image `-S warning -s bash` on the evaluated `run.sh` blocks extracted as one file (538 lines) | one SC2120 on `it_isolation_record` (an excerpt artefact: its argument-passing caller `suite_migrations` is outside the excerpt; the full `run.sh` passes B-SHELL); nothing else |
| Reference scan: `run.sh` functions referenced by the evaluated blocks but defined neither there nor in `lib.sh` | none (the five hits are comments) |
| Out-of-repo logic test (`/tmp`, not committed): the driver sourced with `S4_DRIVER_SOURCE_ONLY=1`, the instance phase and `it_fresh`/`it_drain`/`db_shared_catalog`/the controls stubbed, a fake `docker` (synthetic `go test -json` streams per selector, the locally built `gotestreport` binary for the tools image, call/order log) | **46/46**: derivation of the real source = 112/17/95, I 70 / L 25, package order, no G/own/mcpe2e/TestMain names; **94 and 96 selectors and a 95 with a wrong I/L split refused** (also at main level: exit 2 before the instance); every selector has an environment rule (it.env 20, identity env 33, OBJ_ENV 42; rls_go, sign-host and MCP shapes); all-pass exit 0 with 95 + 12 PASS rows, 107 invocation rows/logs, declared inventory 107, **`db_fresh` before every selector in selector order**, one exec per selector carrying the scratch prefix and `GO_IT_SH`, one container per package started and removed, AC3 through `go_it` `--rm` containers, B-IT-ISOLATION with post-campaign T-7 and T-8, seeds, setup versions, pair-d equality, provenance `check=`/`suite=`; **redaction** (a canary in test output and a generated secret in a driver log both `[REDACTED]`, values absent from the evidence); cleanup receipt zero leftovers and work dir removed; failing test (row FAIL rc 1, 94 PASS, exit 1); undeclared test (reporter finding, FAIL); `db_fresh` refusal (selector not run, FAIL REFUSED); container status 3 with passing events (propagated); missing seed (AC3 10/12); setup version 5 (four rows FAIL); shared-catalog delta (FAIL); control failure (exit 1); leftover (exit 4); TERM mid-selector (exit 143, invocation `interrupted`, result INTERRUPTED, container removed, receipt); `S4_MODE=run` (107 `--rm` invocations, exit 0); non-empty evidence dir refused (exit 2) |
| `docker exec --help` (CLI 27.5.1) | `--env-file` supported (exec mode relies on it) |
| Allow-list (`git diff --name-only` + untracked, `grep -vEf` the 13-line allow-list) | vs `310baf2c`: **changed 13, outside 0**; vs `45a263fd`: changed 1 (the driver) + this Result, outside 0 |
| U+2028/U+2029 (`grep -c -P '\x{2028}\|\x{2029}'`) on the driver and this Result | 0 each |

**Not yet executed:** nothing from S2, S3 or S4 has run against real PostgreSQL, a real core-check image or a real
validation tools image build. The coordinator's S4 campaign is that first run (the reset points' `db_fresh` path, `go_it`
with `-json -list`, `gotestreport` in its image, the exports, T-6/T-8 on real streams).

### Deviations (S4; none architectural)

1. **TestObjectsSignHost without the host fetch.** Its Go test runs by its direct selector with its environment
   (`OBJ_SIGN_HOST_ENDPOINT` naming the unpublished MinIO API port, `OBJ_SIGN_HOST_DIR=/tmp/objsign` because an exec'd
   container cannot add the harness's tmpfs), and a bounded in-container watcher touches the `done` marker once the URL
   is written. The harness's published-port fetch (200/403, part of B-OBJ-IT, not of the test function) cannot run with
   zero published ports; it stays covered by batch 3 (`migrations` suite with a host probe). Open question below.
2. **`TENANCY_*` files container-local.** The four bind-mount sites (tenancy model, business profile model, access
   policy/footprint, access ordering) get `RLS_CANARIES` + `TENANCY_CANARY_FILE=/tmp/rls-canaries.txt` (written by
   `GO_IT_SH`, as `rls_go` does) and `TENANCY_LOG_FILE=/tmp/s4-core.log` instead of `-v` host mounts (no bind mount in
   the campaign); the suite-level logcheck of those outputs is not part of a selector's proof.
3. **Exec-mode wrapper.** `s4_go` = `go_it_begin` + `docker exec` of a scratch prefix and the unchanged `GO_IT_SH` +
   `go_it_convert` (Plan §4.1 permits the reuse and a thin wrapper over the same reporter path); AC3 uses the real
   `go_it`/`rls_go`.
4. **Shuffle through `GOFLAGS`** (`-trimpath -shuffle=on`, keeping the image's `-trimpath`): `GO_IT_SH` has no flag slot
   and no harness edit is needed; the seed comes from the binary's `-test.shuffle N` line and is required.
5. **Tools image** built directly as `g064s4-<run>-vtools` and preset for `ensure_tools` (which would otherwise also build
   the intelligence check image, unused here); task-owned tag per the private-run rule.
6. T-7 appears twice in the `B-IT-ISOLATION` text (after the controls and after the campaign).
7. Selectors are derived from the bound source with the Plan's counts asserted (including per package) rather than read
   from a Plan file; `go_it`'s own `-list` step then proves each selector's expectation is exactly `{(package, Name)}`.
8. Observation (not changed, S3 code): `_iti_t8`'s PASS text prints the `REPORT_DIR` path after the word `REPORT_DIR`
   (`${REPORT_DIR:-…}` expands to the path when set); cosmetic.

### Open questions

1. Deviation 1: is the sign-host selector's Go-test-only proof acceptable for the campaign, with the published-port
   fetch left to batch 3, or should the coordinator grant one reserved loopback port for MinIO in this run (the
   private-browser rule's pattern) so the driver can also perform the fetch?

### Remaining (exact)

The coordinator's real S4 campaign (above); S5 `required-checks.tsv` row `B-IT-ISOLATION migrations` and header counts,
docs, PG-010 (with the campaign's T-7 numbers); S6 batches 1, 3, 4 (including the Docker-dependent GT1/GT2 and
existing-control regression); S7 self-review; S8 staging reconciliation and handback.

### Budget

B5 attempt 1 continues (no new attempt). Correction cycles 0. Inner loops: `set -e` hazards in command substitutions
found by review before any run (`|| true` added), one wrong expectation in the out-of-repo test (test-only), the
shellcheck SC2034 disables.

### S4 correction 1 (writer, stack-free; first real campaign run on `8d712163`, 2026-10-21)

Evidence (coordinator's run, read-only): `~/.local/state/platform-evidence/g064-a1/s4-8d712163/run/s4-summary.tsv`
(campaign **93/95**, AC3 **11/12**, T-1..T-8 and `B-IT-ISOLATION` PASS, no published port, cleanup 0 leftovers),
`go-tests/ac3-migrate-TestEnrollmentCatalog.log`, `go-tests/ac3-tenancy-TestBusinessProfileModel.log`,
`go-tests/ac3-order-b1-business-profile-model.log`. Two distinct causes; both corrections are inside the allow-list.

1. **`ac3-migrate-TestEnrollmentCatalog` — driver/Plan-row environment gap (proven).** `enrollment_integration_test.go:200`
   and the seven `CT-ENR-1_replay/ENR-16*` subtests fail with `ENROLLMENT_REPLAY_SQL_B64 is required`; the other ten
   subtests pass. `run.sh:4010-4011` (B-ENROLLMENT-CATALOG) passes
   `-e ENROLLMENT_REPLAY_SQL_B64="$(base64 -w0 "$REPO/infra/recovery/sql/replay-restrict.sql")"`; the driver's `s4_env_of`
   `migrate:*` rule and Plan §4 row 48 omitted it. Fix: `s4_infra` computes `S4_REPLAY_B64` from the bound
   `infra/recovery/sql/replay-restrict.sql` beside `S4_INVALIDATE_B64`; a `migrate:TestEnrollmentCatalog` case adds the one
   `-e` to the `rls_go` trio (row text `OBJ_ENV + rls_go + ENROLLMENT_REPLAY_SQL_B64 (30m)`); header comment updated. Plan
   §4 row 48 is not amended (see the coordinator disposition below); this Result is the amendment record for the
   omitted variable. Not a product defect.
2. **`ac3-tenancy-TestBusinessProfileModel` and `ac3-order-b1-business-profile-model` — connection exhaustion at
   `IT-BP-REF-2` (proven from source; the stated hypothesis was half of it).** `connect` registers the close in `t.Cleanup`
   (`core/internal/tenancy/integration_test.go:76`), so (a) the 50 race rounds of `IT-BP-REF-2` each hold an owner
   connection until the subtest ends (`business_profile_integration_test.go:1082`, `:1121`), and (b) every superuser purge
   helper of `setup`'s ladder (`purgeAuditHistory` `:340`, `purgeMCPState` `:360`, `purgeDomainState` `:381`) holds its
   connection until the **top-level** test ends: from the embedded top version 19 the ladder (`:162-279`) makes 19 such
   calls (v19 2, v18 2, v17 2, v16 2, v15 2, v14 1, v13 2, v12 1, v11 2, v10 2, v8 1). The runtime pool (`MaxConns=48`,
   `:312`) holds about 32 idle connections after `IT-BP-LINK-3`'s 32 concurrent linkers (`business_profile_integration_test.go:653`;
   pgxpool keeps idle connections), plus `e.own`/`e.rt`. Budget: 19 + 32 + 2 + 50 = 103 > 97 (PostgreSQL default
   `max_connections=100` minus 3 superuser-reserved; no `max_connections` setting anywhere in `infra/`, `compose.yml`
   or the harness) → `53300` at about round 44, deterministically, as both logs show. Why it passed before: the
   established `run.sh` order runs B-TENANCY-IT first, whose ladder leaves the shared database at 5, so
   B-BUSINESS-PROFILE-IT started at 5 with **zero** purge connections (32 + 2 + 50 = 84 ≤ 97) — exactly the fragility
   recorded in [G062 review note 13](G062-root-host-tenant-signup-result.md) ("each step's superuser purge connection is
   held until test cleanup … exceeding `max_connections=100`"), which the G064 fresh-copy-at-top-version start now
   exercises on every run. `TestTenancyModel` (b2, PASS) has the same 50-round `ownerConn` loop (`integration_test.go:956-959`)
   but its 20×N fan-out (`:1039`) comes after it, so its pool is small there. Fix (smallest, allow-listed, G062's own
   suggestion): each of the three purge helpers now `defer`s `c.Close(ctx)` at the end of the helper (the `t.Cleanup` close
   stays; pgx v5.11.0 `conn.go:302-305` returns nil from `Close` on a closed connection), so the ladder leaves no held connection and the budget returns to the
   established 84. No assertion, race count, pool size, server setting or other test changed.
   `business_profile_integration_test.go` (the per-round pattern) is **not** in the allow-list and is untouched; by the
   accounting above the helper fix alone restores the margin that every prior B-BUSINESS-PROFILE-IT PASS relied on.
   Same per-round pattern elsewhere (listed, not changed): `integration_test.go:918-921` (3 connections), `:956-959`
   (50, `TestTenancyModel`), `:1463`/`:1476` (2); `core/internal/tenancy/access` has the same `t.Cleanup` connect helper
   (`access/integration_test.go:146`) but no per-round connection loop.

Environment cross-check of every campaign selector rule (`s4_env_of`) against the `run.sh` call sites and Plan §4:

| Selectors | `run.sh` site | Harness env | Driver rule | Delta |
|---|---|---|---|---|
| migrate Lifecycle, Tenancy/BusinessProfile/Policy Catalog+Migration | 2901, 2989-3115 `go_it` | it.env | it.env, 10m | none |
| migrate RLSCatalog/RLSMigration; identity RLS*; access RLS* | 3362-3611 `rls_go` | it.env + `RLS_CANARIES`, `TENANCY_CANARY_FILE`, `TENANCY_LOG_FILE` (3205-3206) | same trio, 30m | none |
| migrate Objects/Audit/Retention/MCP*/Domains/Branding/Onboarding/Membership/Signup catalogs; access Objects*, IsolationAttacks, Audit, Retention, Domains, Branding, Onboarding, Membership, Authority, Enrollment*, MCPOrdering/Footprint; signup ×3 | 3736-4058 `rls_go` | OBJ_ENV + trio | same | none (signup-it's `it_isolation_t2_repeat` is a suite control, not env) |
| migrate TestEnrollmentCatalog | 4010-4011 | OBJ_ENV + trio + `ENROLLMENT_REPLAY_SQL_B64` | was OBJ_ENV + trio | **missing → fixed** |
| access TestObjectsSignHost | 3700-3703 | + `OBJ_SIGN_HOST_ENDPOINT`, `OBJ_SIGN_HOST_DIR=/run/objsign` tmpfs | + endpoint, `/tmp/objsign` | recorded deviation 1 (passed) |
| mcp ×5 | 4052-4053 | + `MCP_LOG_FILE`, `MCP_INVALIDATE_SQL_B64` | same (from the bound `invalidate.sql`) | none |
| tenancy ×2; access Policy/Footprint/Ordering | 3021-3022, 3080-3081, 3154-3155, 3180-3181 | `TENANCY_CANARY_FILE`/`TENANCY_LOG_FILE` bind mounts; 30m for the access sites | trio with `/tmp` files; 10m / 30m | recorded deviation 2 (passed) |
| identity ×33 (non-RLS) | 5597-5607, 5839-5864 | identity env (no `DB_RETENTION_PASSWORD`) | identity.env, 10m | none |
| G-class persist/recovery (`*_PERSIST_FILE`, `*_PERSIST_IDS`, `RECOVERY_*`), TestMCPE2E | excluded from the campaign | - | - | n/a |

Plan §4 row 48 (coordinator disposition, 2026-10-22): the writer amended row 48 and the Plan revision history in place;
the coordinator struck that edit in the next commit so the accepted Plan r2 bytes stay frozen (CLAUDE.md: record
amendments instead of altering approved Plan bytes). The omitted variable is recorded here instead: row 48's
`TestEnrollmentCatalog` selector runs with `ENROLLMENT_REPLAY_SQL_B64` (base64 of the bound
`infra/recovery/sql/replay-restrict.sql`), as `run.sh:4010-4011` does. The allow-list line in the check table above
(4 changed) describes the writer's commit; after the strike the correction touches 3 files (driver, tenancy
`integration_test.go`, this Result), 0 outside the allow-list.

Stack-free checks (this correction; no stack, database, build or network):

| Command | Outcome |
|---|---|
| `bash -n` on the driver | OK |
| `koalaman/shellcheck:v0.11.0 -S warning` on the driver (`--pull never --network none`, ro; file-level disables unchanged) | **rc 0** |
| Driver sourced with `S4_DRIVER_SOURCE_ONLY=1` (`/tmp/g064s4-scratch/s4-corr1-envrule-check.sh`, not committed): `s4_env_of` for the enrollment, MCP catalog, RLS catalog and tenancy catalog selectors | enrollment rule = trio + exactly one `ENROLLMENT_REPLAY_SQL_B64` (8 args) whose value decodes to the bound `replay-restrict.sql` (sha256 prefix `8ef8f9d9589508f9` both ways); the other rules unchanged (6/6/0 args) |
| The previous writer's 46-case out-of-repo logic test (`/tmp/g064s4-scratch/test.sh`, fake `docker`, not committed) | first run 15/46: its stubbed `s4_infra` and env-section setup predate `S4_REPLAY_B64` (`line 139: S4_REPLAY_B64: unbound variable` under `set -u`; the real `s4_infra` sets it) → the two scratch stubs mirror the new variable (`S4_REPLAY_B64=c3Q=` / `=R`, test-only); rerun **46/46**; the enrollment row in `env.out` carries the trio + `-e ENROLLMENT_REPLAY_SQL_B64=R`, E1-E8 counts unchanged |
| `golang:1.27.3-alpine` (`--network none`, ro source, ro `g058-a1-gomod` module cache, `GOFLAGS=-mod=readonly GOPROXY=off GOTOOLCHAIN=local`, disposable `GOCACHE`): `gofmt -l ./internal/tenancy/`; `go vet -tags integration ./internal/tenancy/` | gofmt clean; vet **rc 0** |
| `git diff --check` | clean |
| U+2028/U+2029 on the four touched files | 0 each |
| Allow-list: touched `core/internal/tenancy/integration_test.go`, the driver, the Plan, this Result | 4 changed, **0 outside** |

**Rerun pending:** nothing here was run against PostgreSQL. The coordinator reruns the full S4 campaign on the new head
(about 11 minutes); expected delta: `ac3-migrate-TestEnrollmentCatalog` 19/19 subtests, `ac3-tenancy-TestBusinessProfileModel`
and `ac3-order-b1-business-profile-model` 35/35 (`IT-BP-REF-2` included), campaign 95/95, AC3 12/12. No test is claimed
passed by this correction. Budget: B5 attempt 1 continues; correction cycle 1 (S4).

### S4 rerun and S5 record

**S4 rerun (coordinator; recorded by the S5 writer from read-only evidence).** Frozen checkout at `03e50356` (S4
correction 1 plus the Plan-bytes strike), evidence `~/.local/state/platform-evidence/g064-a1/s4b-03e50356/`; started
2026-10-22T00:50:55Z, finished 01:01:50Z, driver exit 0; Docker Desktop mounts 21% before the run; checkout status after
the run empty.

| Item | Outcome |
|---|---|
| Campaign | **95/95** PASS (I 70/70, L 25/25); derivation tagged 112, excluded 17, per package identity 35, mcp 5, migrate 25, signup 3, tenancy 2, tenancy/access 25 |
| AC3 | **12/12**: reversed pairs 4/4 (a lifecycle then identity catalog; b business-profile model then tenancy model, both `setup=19`; c `mcp-it` then `domains-it`; d signup catalog then RLS catalog, shared catalog equal, 19 lines); repeat 2/2 (`TestAccessPolicy`); shuffle 2/2 (`rls-it` seed 1792630887984484990, `mcp-it` seed 1792630898333332727) |
| B-IT-ISOLATION | PASS: T-1, T-1a, T-1b, T-1c, T-2, T-3, T-4, T-5a, T-5b, T-5c, T-6 (24 fixtures), T-7, T-8 |
| Reports | 107 declared invocations, 107 `invocation` rows all `pass`, 107 `go-tests/<name>.log` |
| Lifecycle ledger | 2 pristine (1 PASS, 1 deliberate T-1c FAIL), 1 template, 109 fresh (all PASS, 0 residue), 6 refusals (all deliberate controls), 0 drain waits |
| Ports / cleanup | no published port (checked after the campaign); `cleanup.log`: 0 left containers, run-labelled containers, volumes, networks, images; work directory removed |

sha256: `run/s4-summary.tsv` `48eacc1202e5f67a15307528823d33edbd16c14302bf2d279c20c0845d22bddb`; `driver.console.txt`
`2b813311d0a3a31b2ba2e48f7049d3ad3dd69185ee676032f9d8f787cafbde7d`; `run/cleanup.log`
`412fb95b32cf7cf3c46638a924decf88dcda8b9fff5f5c006259051ad9ebed2b`; `run/results.tsv`
`99d5473058653da44ef2d1a661cdb2ce86771f189f0b4e73e6215cadc8ef1f50`.

The first campaign run `s4-8d712163` stays recorded as **FAILED** (campaign 93/95, AC3 11/12; `run/s4-summary.tsv`
sha256 `d76797a133e3b04935c1b1820a1faf6efd52eae3f62a95d3fe345d0ad3253df9`, evidence unchanged); its two causes and
corrections are "S4 correction 1" above. The rerun shows the predicted delta and nothing else failed.

**S5 (writer, stack-free).** Writer: Claude Opus 5.5 (`claude-opus-5-5`), runtime effort not observable (unknown).
Isolated worktree `.claude/worktrees/agent-<id>`, local branch `g064-s5` from `origin/g064-impl`
`03e50356` (asserted), pushed fast-forward to `g064-impl`; not rebased onto staging (S8).

- `infra/validation/required-checks.tsv`: row `B-IT-ISOLATION<TAB>migrations` after `B-SIGNUP-IT`; header `... + 2 G062
  + 1 G064 = 130 rows`, `G064: B-IT-ISOLATION migrations` in the per-Goal list, profile line `full 122` → `full 123`.
  Before/after, re-derived from the file and `infra/ci/lib.sh`: rows 129 → 130; fast 35 → 35, integration 75 → 75
  (no `migrations`), full 122 → 123, staging 5 → 5. Pins checked: `B-CI-CONFIG` derives the per-profile counts from the
  TSV (`run.sh:1487-1494`) and prints them, no literal; `B-CI-CONTROLS` derives its list/count from
  `validate-controls.sh --list` (S5 adds no control); `validate-controls.sh` writes the stub's PASS rows from the TSV
  (`:647`) and its `FULL_SET` names files, not counts; `validate.sh` reads the TSV for required IDs and the selection
  digest. No other pin exists in an allow-listed file; `run.sh`, `validate.sh` and `validate-controls.sh` are unchanged.
  The ID was already recorded by `suite_migrations` (S2); it is now required.
- `docs/foundation/OPERATIONS.md`: new `### Per-test Go evidence (G064)` (fresh copy per invocation; `go_it` and
  `gotestreport` fail-closed classes; the five retained reports with provenance, redaction and 1 MiB bound; TERM/KILL
  interruption semantics; `validate.sh complete` rules and GT1/GT2; D-4 as approved). Profile table full required IDs
  `112` → `123` (the doc's 112 was already stale against the TSV's 122). The B-MCP-IT note "runs last before the drill
  resets the database" is retired (Plan §3.4).
- `docs/foundation/MIGRATIONS.md` (Validation, `:140`): one paragraph on the template, the pristine predicate,
  `db_fresh` and its refusals, and the lifecycle-test exceptions.
- `docs/foundation/DATABASE-ISOLATION.md`: Evidence row B-IT-ISOLATION and `### Test databases (G064)` (canonical-name
  recreation, restored `10-roles.sh` grants, shared-catalog detector, T-1..T-8).
- `docs/foundation/IDENTITY.md`: the `ops` harness order (isolation window steps 1-5, then `mcp-seed.sql`, browsers,
  probes, B-MCP-E2E; refusal stops the overlay, no fallback) and one clause at the seed paragraph.
- `docs/foundation/BOOTSTRAP.md`: unchanged (it lists suites, not suite evidence files).
- `docs/PERFORMANCE.md` PG-010: status "observed on one private test instance ..., not benchmarked; **open** until the
  G064 `migrations`-suite and identity-window observations (S6) and Codex review"; observations derived with a script
  outside the repository from `run/db-lifecycle.tsv`, `run/s4-summary.tsv`, `run/s4-phases.tsv`: template 9,822,735
  bytes, 1,921 ms; `db_fresh` ×109 min 1,398 / median 1,478 / max 1,628 ms (checks 545, drop 144, create 145,
  postconditions 640 ms median); 6 refusals; no drain wait; campaign 459 s, AC3 118 s; recreations 141 s vs 296 s of
  campaign test time; the `53300` history of the first run as the connection-contention observation.

**Observation for review (PG-010 stop condition, not decided by the writer).** Plan §10 lists "PG-010 observations
materially above expectation (template or recreation taking more than the invocation it precedes)" as a stop. In the
campaign, `db_fresh` alone exceeded the following single-test direct-selector invocation for **5 of 95** selectors (by at
most 73 ms: `migrate-TestPolicyCatalog`, `identity-TestEnrollmentRevocationFaultRollback`, `identity-TestIdentityCatalog`,
`identity-TestIssuanceGuardFailsClosed`, `identity-TestRLSFirstSignInRace`, invocations of 1.39-1.50 s on a warm build
cache), and with the bounded drain poll included (`it_fresh`) for 23 of 95. No AC3 invocation (4.6-13.5 s) was shorter.
The suite's 50 reset points precede multi-test invocations and are first observed in S6 batch 3. Recorded in PG-010 for
Codex disposition.

**Observation (cosmetic, unchanged).** `_iti_t8`'s PASS text prints the `REPORT_DIR` path after the word `REPORT_DIR`
(`${REPORT_DIR:+REPORT_DIR}${REPORT_DIR:-...}` expands to the path when set; seen in the rerun's `results.tsv`), and
`_iti` keeps each control's detail only up to its first `;`, which drops T-3's and T-6's closing parenthesis in the
summary. Not changed: a correct fix is not a single-line edit, and editing `run.sh` would separate the S4 evidence from the
tested `run.sh` blob; no assertion depends on either text.

Stack-free checks (S5; no stack, database, build or network):

| Check | Outcome |
|---|---|
| `bash -n` / `shellcheck` | not applicable: no shell file touched in S5 |
| Required-checks self-consistency (`/tmp/g064-s5/tsvcheck.sh`, out of repo): header row total and per-Goal sum vs rows; per-profile rows (suites from `infra/ci/lib.sh` + `final`) vs the header profile line; duplicate IDs | base `03e50356`: 129/129/129, fast 35, integration 75, full 122, staging 5; S5: 130/130/130, fast 35, integration 75, full 123, staging 5; no duplicate; **OK** both |
| `git diff --check` (vs `03e50356`, before the commit) | clean (rc 0) |
| U+2028/U+2029 (`grep -c -P '\x{2028}\|\x{2029}'`) on the 7 touched files; the S5 diff | 0 in each file (pre-existing 0); 0 in the diff |
| Allow-list (`git diff --name-only 310baf2c` + untracked, `grep -vEf` the 13-line allow-list, sha256 `4ee27e22…9627`, read from `origin/staging`) | **changed 19, outside 0**; vs `03e50356`: 7 files |
| Markdown links added (`/tmp/g064-s5/links.py`, out of repo: target file exists, `#anchor` equals a heading slug) | 6 links, 0 unresolved |

**Remaining for S6 (coordinator):** batch 1 (`lint lint-base lint-core unit-core ci-controls`, including GT1/GT2 and the
regression of the controls on the changed stub, B-CI-CONFIG on the new TSV counts, B-SHELL on the S1-S4 harness delta),
batch 3 (`migrations` suite once: the 50 reset points, `go_it` on every invocation, B-IT-ISOLATION as a required ID,
B-REQUIRED-CHECKS with 130 rows, the deferred sign-host host fetch), batch 4 (identity overlay focused, `db-lifecycle.tsv`
presence in `ops`), then S7 self-review and S8 staging reconciliation. The PG-010 short-selector observation above needs
a Codex disposition. Budget: B5 attempt 1 continues; no new attempt; correction cycles unchanged.

### S4 correction 2 (Codex source review f42317b0)

Writer: Claude Fable 5.1 (`claude-fable-5-1`), runtime effort not observable from inside the agent (unknown). Checkout:
isolated worktree `.claude/worktrees/agent-<id>`, local branch `g064-s4-corr2` from `origin/g064-rb`
`e7c4efc6` (asserted before editing; `g064-rb` is `g064-impl` rebased onto staging `f9048e8a`), pushed as the new branch
`g064-s4-corr2` (no MR, no rebase; `g064-impl` and `g064-rb` untouched). Stack-free only: no Compose stack, database,
harness suite, image build, pull or real Docker mutation by the writer. Input: the Codex review
G064-s3-s4-source-codex-review-20261022.md (not included) (`origin/staging`, commit
`f42317b0`): CORRECTION LOOP, two required safety corrections to the S4 driver before the campaign is rerun, three
bounded deviation rulings. Files touched: the driver, this Result and (text only) `infra/ci/validate-controls.sh`.

**Correction 1 — exact owned resources registered before cleanup** (`management/v0/results/G064-s4-private-run-driver.sh`).
The driver no longer sets `OWNED_REG=""`; it uses the harness's owned-resource registry and the existing capture path:

- `s4_registry_init` (P0, before the first Docker mutation): `$WORK/owned/` in the `validate.sh` registry shape (mode
  700; `owned.tsv` with `CI_REGISTRY_HEADER`, `job local-<12 hex>`, `token`, `repo` = the resolved checkout; `token`
  file; `cid/`; `start-inventory.ids`), then `work_root` and `label g001.run=<run id>` rows (`lib.sh registry_row`), then
  `s4_inventory_start`: the four start listings (`docker ps -aq --no-trunc`, `network ls -q --no-trunc`,
  `volume ls -q`, `image ls -q --no-trunc`) each checked for exit status and written both as `$WORK/inventory-start.ids`
  (read by `lib.sh id_at_start`) and as the registry's `start-inventory.ids`. A failed listing is a stop (exit 2),
  never an empty inventory.
- The driver's `compose`/`compose_migrate` wrappers now call `capture_after` after every call, as `lib.sh`'s own
  wrappers do (`owned_capture`: every container of the project absent at the start with a working directory under
  the checkout or `$WORK`, its network IDs, mounted volume names and image tag+ID), with the `build`/`--build` words
  dropped from the capture arguments so `record_project_images` (which runs the yq tool image) is not used.
- `s4_own_image TAG` after each of the five builds (migrate, core-check, vtools, minio, minio-init; minio/minio-init
  again after `obj_minio`'s cache-hit rebuild): one `docker image inspect -f '{{.Id}}'` of the exact task-owned tag,
  fail-closed (stop 2 if it cannot be inspected), `image TAG ID` registry row; an ID that already existed before the
  run (a fully cached identical build) is noted in `images-shared` and is only untagged at cleanup.
- `s4_capture_check` after `postgres` and after `minio`: every container of the project (fail-closed `docker ps`)
  must have a `container <id>` capture row, otherwise stop 2 naming the container; the registry's row counts are logged.
- Standalone containers were already bound by `docker_run_owned` once a registry exists (cidfile written by the
  Docker CLI at creation plus a `standalone <id> <purpose>` row): the per-package core-check containers, every
  `go_it`/`rls_go` `--rm` container of AC3 and run mode, and `obj_minio`'s minio-init.
- Cleanup (`s4_finish` → `s4_cleanup`) is now one removal pass by the existing
  `infra/ci/cleanup-owned.sh --registry $WORK/owned --mode finalize --out owned-inventory.tsv` (the `validate.sh` F2
  component: removal only of resources it verifies as run-owned from the registry and the start inventory, never by
  label or name alone; a registered label or `<project>-` image name only reports `unproven`), after a last
  `owned_capture` while the resources exist and the untagging of run tags on pre-existing image IDs (only while another
  reference keeps the image). The driver's by-project `compose down -v`, by-name `docker image rm` loop and by-label
  `docker rm -f` loop are gone. Then every registered resource is checked by exact ID/name: the finalizer's inventory
  row where it has one, otherwise a fail-closed presence check (`docker inspect --type ...`: absent only on an answered
  "No such ..."; any other failure is `failed`).
- Receipt `<evidence dir>/cleanup.log`: the finalizer's summary and exit, one row per registered resource
  (`kind<TAB>id<TAB>state<TAB>reason`; state `removed` | `gone` (already absent: `--rm` containers the daemon
  removed, per-package containers the campaign removed) | `untagged` | `remaining` | `refused` | `unproven` |
  `failed` | `unknown`), the finalizer's rows without a registry row (project-labelled candidates, query/daemon/
  registry failures), the counts line `owned resources: removed= gone= untagged= leftover= unknown=` and
  `cleanup verdict: CLEAN | LEFTOVER | UNKNOWN`. `<evidence dir>/owned-inventory.tsv` (the finalizer's inventory) and
  `<evidence dir>/owned-registry.tsv` (the registry rows recorded while the resources existed) are kept.

**Correction 2 — fail closed when Docker cannot be read.**

- `s4_no_ports`: the two `docker ps` listings run separately; each exit status is checked; only stdout is grepped for
  `->` (stderr goes to `docker-errors.log`, exported redacted under `driver-logs/`). A failed listing sets
  `S4_PORTS="UNKNOWN at <when>: ..."` and returns 2: the early sites stop (exit 2), the post-campaign site adds
  `published ports: UNKNOWN ...` to the verdict reasons (exit 1); an UNKNOWN or FAIL is never overwritten by a later
  PASS, and the `no-published-ports` summary row shows `UNKNOWN`.
- Cleanup: no `|| true` or `2>/dev/null | wc -l` on any listing or inspection. Any `failed` row (finalizer query/
  inspect/daemon failure, driver inspect failure, untag failure), a finalizer that did not complete (exit other than
  0/18) or a missing inventory makes the verdict **UNKNOWN**; a `remaining`/`refused`/`unproven` row makes it
  **LEFTOVER**; only all-removed/gone is **CLEAN**. Exit codes (driver header): **4** = everything passed but a verified
  leftover; **5 (new)** = everything passed but cleanup could not be verified (UNKNOWN takes precedence over LEFTOVER);
  an earlier non-zero status is kept and the receipt still states the verdict. The receipt is always written before
  the work directory is removed.

**Codex rulings recorded (bounded deviations; no code change beyond the label):**

- Sign-host selector: accepted **only as the isolated URL-issuance half** of the direct-selector campaign; the
  in-container `done` marker does not prove the host fetched the signed URL. Batch 3's B-OBJ-IT host probe must still
  prove HTTP 200, body/headers and the unsigned-path 403; the S4 selector is not that proof. No extra published port
  is added for the selector.
- Container-local `TENANCY_*` files: acceptable for the direct selectors provided the existing suite-level
  access/log checks still run in batch 3 and their outcomes are recorded separately in this Result.
- GT1 partial publication: the existing V3 behavior is kept (exit 14 on completeness, `go-tests.tsv` not published,
  other scanned evidence published); accepted as a bounded adapter deviation from Plan r2 §5.4's "nothing published"
  wording and labeled as such here and in the GT1 description. The missing report is never synthesized or treated as
  complete. `infra/ci/validate-controls.sh`: the `ctl_GT1` comment gains three lines naming the deviation and its
  acceptance, and the `detail` text appends "; the other scanned evidence is published without it (recorded deviation
  from Plan r2 §5.4 'nothing published')". Text only; no expectation, selector or logic change (`B-CI-CONTROLS`
  derives its list and count from `--list`, not from detail text). The G063-updated `validate*.sh`/objects test are
  otherwise untouched.

**Rerun required.** The S4 campaign must be rerun on the corrected head (this branch's head or its reconciled
successor): the `s4b-03e50356` rerun recorded above stays valid evidence of the selectors on `03e50356` but does not
attribute the corrected driver; the rerun must attribute the corrected source and retain every selector outcome. No
selector or assertion changed: `s4_derive`, `s4_env_of`, `s4_campaign`, `s4_ac3`, `s4_verdict` and the 95-selector
inventory are untouched (the diff is confined to the header, the wrappers, the registry helpers, `s4_no_ports`, the
`s4_infra` ownership hooks, `s4_main`'s registry setup and `s4_finish`/`s4_cleanup`).

Stack-free checks (this correction; no stack, database, build, pull, network or real daemon mutation):

| Command | Outcome |
|---|---|
| `bash -n` on the driver | OK |
| `koalaman/shellcheck:v0.11.0 -S warning` (existing pinned local image; `--pull never --network none`, ro) on the driver | first pass rc 1: SC2034 one unused local (`line`) → removed; rerun **rc 0**; `-S info --include=SC2251` **rc 0** |
| Failure-path harness `/tmp/g064s4-corr2/{fakedocker/docker,test.sh}` (outside the repository, not committed): a `docker` shim with an on-disk fake inventory, injectable failures (`FAKE_FAIL` per listing/inspect/version, `all` = daemon unreachable) and a removal log; the driver sourced with `S4_DRIVER_SOURCE_ONLY=1`; `s4_cleanup` runs the **real** `infra/ci/cleanup-owned.sh --mode finalize` against the shim | **41/41**: (a) `s4_no_ports`: `docker ps` failing → return 2, `UNKNOWN`, never `PASS`; a published port → return 1 `FAIL at postgres: <name> 0.0.0.0:5432->5432/tcp`; no port with a stderr warning containing `->` → `PASS` (stderr not merged); an earlier UNKNOWN/FAIL never overwritten; the failure in `docker-errors.log`. (b) cleanup inventory failing: listings failing → **exit 5**, `cleanup verdict: UNKNOWN`, no `CLEAN`, `query ... failed` rows, finalizer exit 18, receipt/inventory/registry kept; daemon unreachable → exit 5, every registered resource rowed `failed` with its exact ID, `docker daemon failed`, nothing removed; an earlier status 1 kept; finalizer missing → `did not complete (exit 127)`, inventory missing, exit 5; inspect failing → exit 5 with the finalizer's `docker inspect failed` rows. (c) successful empty inventory: every registered resource already absent → **exit 0, CLEAN**, rows `gone` with exact IDs, finalizer exit 0, nothing removed, foreign resources untouched (`removed=0 gone=5 leftover=0 unknown=0`); all present and owned → exit 0 CLEAN, rows `removed` with exact IDs, the removal log exactly the five owned resources. (d) through the driver's real capture path: registry shape (header, job/token/repo/work_root/label/project rows, start inventory holding the foreign IDs, identical copies); `compose up` via the wrapper captured the exact container ID, network ID and volume name and no row for the pre-existing base image; `s4_capture_check` passes and logs counts; `s4_own_image` records tag + exact ID, and a build whose ID pre-existed is noted in `images-shared`; `docker_run_owned` bound a `-d` and a `--rm` container by cidfile + standalone row (two cidfiles); cleanup → exit 0 CLEAN with rows `removed` (container, network, volume, core-check image, gocheck container `bound at creation (s4-gocheck-migrate)`), `gone` for the `--rm` container, `untagged` for the shared-ID tag; removal set exactly the owned set plus the untag, base image and foreign resources kept (`removed=5 gone=1 untagged=1 leftover=0 unknown=0`); a container with the run's project label but a foreign working directory → `refused`, a label-only container → `unproven`, both kept, owned ones still removed, **exit 4 LEFTOVER**; `s4_capture_check` fails closed on a failed listing and on a project container without a capture row (named); `s4_registry_init` fails closed on a failed start listing (no start inventory written); `s4_own_image` fails closed. (e) `s4_finish` end to end (subshells): daemon unreachable → exit 5 with the receipt kept and the work directory removed; clean → exit 0 with `driver-logs/docker-errors.log` exported; refused leftover → exit 4; earlier status 1 kept; no work directory → entry status |
| Previous writer's 46-case logic test `/tmp/g064s4-scratch/test.sh` (unchanged, its own fake docker, stubbed `s4_infra`) | **43/46**; the 3 failures (S1, S10, S11) assert the old receipt wording (`left project containers: 0`, `left run-labelled containers: 1`) and, S11, a by-name `container rm` of the per-package container in the fake's order log (the finalizer now removes it by its cidfile-bound ID, which that fake does not produce). The new receipts: `allpass` CLEAN exit 0; `left` **LEFTOVER exit 4** naming `leftover1` as `unproven` (label only); `die` (TERM) exit 143 with the receipt written. Derivation (112/17/95, I 70 / L 25, 94/96/split refusals), environment rules, campaign/AC3/control/redaction/exit-code scenarios: unchanged, 43 PASS |
| `git diff --check` | clean |
| U+2028/U+2029 (`grep -c -P`) on the driver, `validate-controls.sh`, this Result | 0 each |
| Allow-list (`git diff --name-only e7c4efc6` + untracked) | 3 files (driver, `validate-controls.sh`, this Result), **0 outside** |

Receipt sample from scenario (d) (tabs shown as ` | `):

```
cleanup: removed=5 refused=0 unproven=0 remaining=0 failed=0
finalizer exit: 0
container | fbd56235...70af | removed | registered Compose project, working_dir under work_root
network | efebbe3a...76d13 | removed | capture row or owned attachment
volume | g064s4-d1-58131_pgdata | removed | capture row or owned mount
image | g064s4-d1-58131-core-check:latest|sha256:cee3e026...be4c | removed | build or capture row matching tag and ID
container | 93cd4d0c...d3b0 | removed | bound at creation (s4-gocheck-migrate)
container | 1251dca6...2f69 | gone | absent at the end
owned resources: removed=5 gone=1 untagged=1 leftover=0 unknown=0 (finalizer rows 5, exit 0)
cleanup verdict: CLEAN
```

**Not executed:** nothing here ran against a real daemon or PostgreSQL. The coordinator's rerun is the first real
exercise of the registry/finalizer path inside the driver (`cleanup-owned.sh` itself is the existing `validate.sh` F2
component). Two run-time notes for that rerun: (1) Compose's `working_dir` label must lie under `$WORK` for the
project's containers to be owned (as in `lib.sh owned_capture`; a canonicalized `TMPDIR` would show as `refused` →
exit 4, not a silent removal); (2) a run tag whose image ID pre-existed is untagged only while another reference keeps
the image, otherwise it is reported `remaining` (exit 4) rather than deleted, following the finalizer's never-remove-
start-inventory rule. Budget: B5 attempt 1 continues; correction cycle 2 (S4); no new attempt.

### S6 batch 3 failure and correction 3

Writer: Claude Fable 5.1 (`claude-fable-5-1`), runtime effort not observable from inside the agent (unknown). Checkout:
isolated worktree `.claude/worktrees/agent-<id>`, local branch `g064-s6-corr3` from `origin/g064-s4-corr2`
`ded22f8c` (asserted before editing; `ded22f8c` = `e7c4efc6` plus S4 correction 2), pushed as the new branch
`g064-s6-corr3` (no MR, no rebase, no other branch touched). Stack-free only: no Compose stack, database, harness
suite, image build, pull or image substitution by the writer. Files touched: `infra/validation/run.sh` and this Result
(`infra/validation/lib.sh` read, unchanged; `infra/ci/validate-controls.sh` unchanged: no count or pin depends on the
reset-point count, see below).

**Batch 3 (coordinator; the one `migrations` suite run on `e7c4efc6`), recorded as FAILED and preserved.** Evidence
`~/.local/state/platform-evidence/g064-a1/s6-e7c4efc6/b3/` (`run.console.txt` sha256 `6d3972c5…6598a`, `report/results.tsv`
`1b7584af…203d`): run `3679a2`, base `f9048e8a`, 2026-10-22T01:28:01Z to 01:54:28Z, `run exit 1`; 45 PASS, 4 FAIL plus
B-REQUIRED-CHECKS (`run.console.txt:182`, `:240` `RESULT: FAILED`). Every other check of the suite passed, including
B-IT-ISOLATION (T-1..T-8) and the untouched persist pairs (`rls-persist-seed/verify`, `obj-persist-seed/verify`,
`audit-persist-seed/verify`, `run.console.txt:97-98`, `:114-115`, `:129-130`). Failed commands and logs:

| Class | Check | Invocation (`go_it` row) | Console | Per-test log (`report/go-tests/`) |
|---|---|---|---|---|
| A | B-TENANCY-MIGRATION | `tenancy-persist-seed` rc=1 `package-failed:core/internal/migrate,exit:1`: `tenancy_integration_test.go:915: up: "no pending migrations\ncurrent_version\t19\n"`; `tenancy-persist-verify` rc=1: `tenancy_integration_test.go:947: open /out/persist.ids: no such file or directory` | `:27-44` (FAIL `:44`) | `tenancy-persist-seed.log` `2ced2472…b44c8`, `tenancy-persist-verify.log` `7fc15666…bd024` |
| A | B-BUSINESS-PROFILE-MIGRATION | `business-profile-persist-seed` rc=1: `business_profile_integration_test.go:718: up: "no pending migrations\ncurrent_version\t19\n"`; `business-profile-persist-verify` rc=1: `:747: open /out/persist.ids` | `:49-66` (FAIL `:66`) | `business-profile-persist-seed.log` `4caa86f1…c083d`, `business-profile-persist-verify.log` `67ee2bfc…f01d1` |
| A | B-POLICY-MIGRATION | `policy-persist-seed` rc=1: `policy_integration_test.go:558: up: "no pending migrations\ncurrent_version\t19\n"`; `policy-persist-verify` rc=1: `:579: open /out/persist.ids` | `:71-88` (FAIL `:88`) | `policy-persist-seed.log` `fe2d4237…5367`, `policy-persist-verify.log` `05222b37…8e21` |
| B | B-RLS-MIGRATION | every RLS `go_it` invocation rc=0 pass (`rls-migration`, `rls-persist-seed`, `rls-persist-verify`, `rls-provisioned-su`, `rls-provisioned-adm`, `rls-provisioned-v5`, `:96-101`); the check FAILs with `provisioning: [21:34:49] G064 go_it rls-provisioned-su: rc=0 pass findings=` … (`:99-104`) | `:102-104` | `rls-provisioned-su.log` `ff0c237c…06d1` (PASS) |
| C | (end of run) | `resources created and not removed: container adoring_yalow` | `:242` | `report/inventory-leftover.txt` `902a9f11…5ac4` (one line) |

**Class A — the version-scoped seeds ran on a fresh v19 copy.** `it_fresh` before each `*-persist-seed` (G064 S2, Plan
r2 §4 "`db_fresh` before the **seed** only") gives the seed a fresh copy of the version-19 template, but the three
seeds are historical-subset tests: `TestTenancyPersistSeed` runs `runnerFor(cfg, subsetFS(3)).Up` and requires the
output to end in `current_version\t3`; `TestBusinessProfilePersistSeed` `subsetFS(4)` → `current_version\t4`;
`TestPolicyPersistSeed` `subsetFS(5)` → `current_version\t5`. On a version-19 database `Up` prints `no pending
migrations` / `current_version 19`, the seed fails before writing `/out/persist.ids`, and the verify fails on the missing
file. Before G064 these seeds ran on the database their `Test*Migration` had just left, which is the state they are
written for.

Per pair, read from each Migration test's end state (the test must leave the database at the seed's expected version):

| Pair | Migration end state | Seed expects | Changed |
|---|---|---|---|
| tenancy | `TestTenancyMigration`: MG-5 `down` to 2 then `up` → 3; ESC-1 runs at version 3 and ends with `metadataIdentityAndTenancyOnly(... subsetFS(3))` → **version 3** | `subsetFS(3)` Up suffix `current_version\t3` (satisfied by `no pending migrations\ncurrent_version\t3\n`); verify `version` = 3 | **yes** |
| business profile | `TestBusinessProfileMigration`: MG-BP-5 `down` to 3 then `up` → 4; MG-BP-7 at version 4, ends with `...BusinessProfileOnly(... subsetFS(4))` → **version 4** | `subsetFS(4)` → 4; verify 4 | **yes** |
| policy | `TestPolicyMigration`: MG-POL-5 `down` to 4 then `up` → **version 5** (last subtest) | `subsetFS(5)` → 5; verify 5 | **yes** |
| RLS | `TestRLSPersistSeed` calls `reset(t, cfg)` itself, then `rlsRunner6` Up → 6 (independent of the prior state; passed on the fresh v19 copy at `:97`) | `current_version\t6` after its own reset | no |
| objects | `TestObjectsPersistSeed` (`core/internal/tenancy/access`) uses `setupObjects` on the full schema; no version subset (passed `:113`) | the version-19 schema | no |
| audit | `TestAuditPersistSeed` calls `reset(t, cfg)` then `auditUp` → 8 (passed `:129`) | `current_version\t8` after its own reset | no |
| other suites | no other `*PersistSeed` exists (`grep func Test.*PersistSeed core/`: migrate tenancy/business_profile/policy/rls/audit, access objects); `ops`/`identity` suites declare none | — | — |

**Codex ruling (bounded Result amendment for these failing pairs only, quoted from the dispatch):** "one fresh database
copy before the Migration invocation, then Migration → PersistSeed → `compose stop postgres` / `up -d --wait postgres`
(volume retained) → PersistVerify serially on that same database; NO db_fresh between those steps; keep separate fresh
copies for Catalog and other independent selectors; leave the direct-selector campaign unchanged. Remove the `it_fresh
... *-persist-seed` call only for the pairs that fail (tenancy, business profile, policy); do not touch pairs that
passed (RLS, objects, audit seed/verify) unless you prove they are affected." Plan r2 §4's row "`db_fresh` before the
**seed** only" and its count of **50** reset points are therefore amended by this Result for these three pairs (47
reset points in the `migrations` suite; the three Migration invocations keep theirs). No check, pin, control stub or
required-checks row depends on the reset-point count (`run.sh`, `validate.sh`, `validate-controls.sh` grep: the count
appears only in the frozen Plan and in comments), and the declared invocation inventory `GO_IT_migrations` is unchanged
(the seed and verify invocations still run, with their `go_it` rows).

**New sequence per affected pair** (`tenancy_migrations`, `business_profile_migrations`, `policy_migrations`; the
catalog invocation before each keeps its own fresh copy):

1. `it_fresh … <pair>-migration` (one fresh copy of the template) → `go_it <pair>-migration` (`TestTenancyMigration` 6/6,
   `TestBusinessProfileMigration` 6/6, `TestPolicyMigration` 5/5), leaving version 3 / 4 / 5;
2. if the Migration step passed: `go_it <pair>-persist-seed` on that database (no `db_fresh`), `compose stop postgres`,
   `compose up -d --wait postgres` (volume retained), `go_it <pair>-persist-verify` on the same database (no `db_fresh`);
3. if the Migration step did not pass (refused `db_fresh`, non-zero invocation or subtest count short): the new
   `it_group_skip` writes `G064 group after <pair>-migration: the Migration step did not pass; this invocation did not run`
   to the seed and verify logs and logs the skip (stderr); the seed, restart and verify do not run on a database whose
   state is not the group's precondition, and the check FAILs closed with that reason in its tail (the same shape as an
   `it_fresh` refusal). This gate is the one addition beyond the ruling's literal sequence (a bounded fail-closed guard;
   nothing runs on an unknown state).

The RLS, objects and audit pairs are unchanged (`it_fresh` before their seeds remains; asserted by the scratch test).

**Class B — the `go_it` status line leaked into a command-substitution capture.** `rls_migration` captures
`prov="$(rls_provision …)"` (`run.sh`, now `:3606`) and treats any non-empty value as failed provisioning parts.
`rls_provision_parts` calls `rls_provisioned_test` three times, which (S3) runs `go_it`; `go_it_convert`'s status line
`G064 go_it <name>: rc=… ` is printed by `lib.sh log` on **stdout**, so the capture received three status lines
(`:99-101`, verbatim in the FAIL detail `:102-104`) although every part passed. Fix: in `go_it_convert` the status line
and the `export failed` line go to **stderr** (`>&2`); `run.sh:114` `exec > >(tee -a "$HARNESS_LOG") 2>&1` sends stderr
through the same tee, so both lines still reach the console and the harness/run log (ordering preserved: one pipe). The
reporter's findings semantics are unchanged (`$name.status` is read, not the console). For consistency the other
status lines of the `it_*` family go to stderr too: `it_fresh`'s `db_fresh before … refused` line and the new
`it_group_skip` line; and the two pre-existing `log` calls inside the captured `rls_provision_parts` (`admin residual`,
`PROV-6 delta`; failure paths only) likewise, so the capture can hold only the printed `bad` parts. Not changed: the
`tail -30` of a failed provisioned-test log inside `rls_provision_parts` still lands in the capture by design
(pre-existing; failure path only; it is the diagnostic of an already-failed part).

Sites checked (every `$(` capture or pipe in `run.sh` and `lib.sh` wrapping `go_it`, `rls_go`, `rls_provisioned_test`,
`rls_provision`, `rd_go` or an `it_*`/`_iti*` helper):

| Site | Wraps | Hazard | Disposition |
|---|---|---|---|
| `run.sh` `rls_migration`: `prov="$(rls_provision …)"` | `rls_provision_parts` → `rls_provisioned_test` ×3 → `go_it` → `go_it_convert`; `rls_pg`, `rls_prov`, `rls_prov_catalog`, `compose … config`, `yq_stdin`, `docker exec` | **yes** (the b3 failure) | fixed (stderr); `rls_pg`/`rls_prov`/`rls_prov_catalog`/`docker_run_owned`/`registry_row`/`id_at_start` contain no `log` call |
| `run.sh` ×44 `n="$(it_count LOG TEST)"` (every suite) | `grep` only | none (no helper, no `log`) | unchanged |
| `run.sh` ×13 `lc="$(rls_logcheck …)"`, ×2 `lc="$(policy_logcheck …)"` | `vtools logcheck`, `compose logs`; no `go_it`, no `log` call inside either function | none | unchanged |
| `run.sh` `_iti_t1c`/`_iti_t4`: `"$(_iti_n …)"`, `"$(_iti_login … 2>&1)"` | `psql_maint`/`compose exec`; no `go_it`, no `log` | none | unchanged |
| `run.sh` ×12 `cid=/probe=/out=/w="$(docker_run_owned …)"` (ops, web, identity, drill) | `docker run`, `registry_row`; no `log` | none | unchanged |
| `run.sh` `rd_go` (drill): called directly, never captured; `envf="$(rd_itenv …)"` only | — | none | unchanged |
| pipes: `go_it`/`rls_go`/`it_fresh` appear only in `a && b || c` chains, never on the left of `\|` | — | none | unchanged |
| `lib.sh`: no capture of any `go_it`-family helper; `go_tests_export`, `redact_log`, `docker_run_owned`, `capture_after`, `owned_capture`, `_db_ledger` have no `log` call reachable from `go_it` | — | none | unchanged |

**Class C — `container adoring_yalow`.** The end-of-run message comes from `lib.sh cleanup`
(`comm -13 inventory-start inventory-end`, `docker_inventory` = `docker ps -a --format 'container {{.Names}}'` …): the
inventory is **host-global by name**, so any container created on the host between 01:28:01Z and the end inventory at
01:54:28Z and still present is reported, whoever created it (precedent: `maint-secretscan-r1/run1/console.log`
reported another task's `container platform-test-g062-unit`). The name is daemon-generated (no `--name`). Every
`docker run` in `run.sh`/`lib.sh` is `--rm` (`go_it`/`rls_go`/`rd_go` through `docker_run_owned "$name" --rm …`,
`vtools` incl. `gotestreport`, `yq_stdin`, the busybox fallback, `obj-minio-init`, the drill's one-shot `rd-stage`
runs) or named and removed by its owner (`rls_pg` `g008-prov-*`, `obj-sign-host`, the drill's `g013-rd-*`); the three
`docker create` sites (`:5795`, `:6071`, `:6127`) are in the `ops`/`web`/`identity` suites, not in `migrations`, and are
removed by name; `compose run` containers carry project names. No G064-changed path (`go_it`, `rls_go`,
`rls_provisioned_test`, `vtools`/`gotestreport`, `it_fresh`/`db_fresh`) creates a container without `--rm`, and the
registry (`owned-registry.tsv`: 92 `standalone` rows = the 103 cidfiles of this run's `docker_run_owned` containers,
28 Compose containers) binds every harness container by ID. The daemon's event buffer no longer covers the window
(`docker events --since 2026-10-22T01:27:00Z --until 02:00:00Z --filter container=adoring_yalow`: empty) and the
container is gone (`docker ps -a --filter name=adoring_yalow`: none), so the creator cannot be proven from the host.
The evidence points to a **concurrent foreign container**: the S4 correction 2 writer worked on this host inside the
window (its scratch directory `/tmp/g064s4-corr2` timestamps 01:34-01:41Z, its commit `ded22f8c` at 01:42:20Z) and ran
the pinned `koalaman/shellcheck:v0.11.0` image with `docker run` (the command line recorded in "S4 correction 2" does not mention `--rm`); an
exited unnamed container of that run would appear exactly as `container adoring_yalow` in this diff and would then be
removed by that worker's cleanup. Disposition: **not a G064 harness leak; pre-existing host-global inventory
behaviour; no `run.sh`/`lib.sh` change for (C)**. The rerun should keep other Docker activity off the host for its
window, or the coordinator should expect to attribute a foreign name the same way. (This writer's own shellcheck runs
below used `docker run --rm`.)

**Stack-free checks (no stack, database, build, pull, network or image substitution):**

| Command | Outcome |
|---|---|
| `bash -n infra/validation/run.sh`, `bash -n infra/validation/lib.sh` | OK |
| `docker run --rm --pull never --network none -v "$PWD:/repo:ro" -w /repo koalaman/shellcheck:v0.11.0 -S warning -x -P SCRIPTDIR infra/validation/run.sh infra/validation/lib.sh` (the B-SHELL flags; existing local image `61862eba1fcf`) | **rc 0** (a first pass without `-x -P SCRIPTDIR` showed only the two cross-file SC2034s `KEEP`/`LEDGER_KIND`, both pre-existing and resolved by the harness flags); `-S info --include=SC2251` rc 0 |
| Scratch logic test `/tmp/g064s6-corr3/test.sh` (outside the repository, not committed): the changed functions extracted verbatim from the working tree (`it_count`, `it_group_skip`, `it_fresh`, `go_it`, `go_it_convert`, `go_it_publish`, `rls_provisioned_test`, `tenancy_migrations`, `business_profile_migrations`, `policy_migrations`) and the base `go_it_convert`/`it_fresh` from `ded22f8c`; `docker_run_owned`, `vtools` (a tar of `NAME.tsv/.log/.status`), `go_tests_export`, `compose`, `db_fresh`, `record` stubbed; `lib.sh log` verbatim | **40/40**: (B) `x="$(go_it …)"` is empty with rc 0 and the status line on stderr; the `rls_provision_parts` shape `$(… rls_provisioned_test … \|\| bad+=…; printf '%s' "$bad")` captures exactly `""` when passing and exactly `PROV-2/3-superuser ` when failing (rc 1 status line on stderr); the **base** `go_it_convert` in the same capture yields `…G064 go_it rls-provisioned-su: rc=0 pass findings=…` (the b3 class reproduced), the new one is clean again; the `export failed` line on stderr; `it_fresh` refusal: rc 1, capture empty, reason in LOG and on stderr (the base version polluted the capture); `it_group_skip` writes both logs, stdout empty. (A) for each of the three pairs the exact call sequence is `it_fresh <p>-catalog; go_it <p>-catalog; it_fresh <p>-migration; go_it <p>-migration; go_it <p>-persist-seed; compose stop postgres; compose up -d --wait postgres; go_it <p>-persist-verify` with no `it_fresh` after the Migration and none for the seed, and `B-<PAIR>-MIGRATION PASS` recorded; a Migration one subtest short → no seed/restart/verify, both logs carry the group reason, FAIL recorded; a refused `db_fresh` before the Migration → no Go invocation of the group, FAIL recorded; `it_fresh` before `rls-`, `obj-` and `audit-persist-seed` still present in `run.sh`; the diff against `ded22f8c` under `infra/` is `run.sh` only |
| `git diff --check` | clean |
| U+2028/U+2029 (`grep -c -P '\x{2028}\|\x{2029}'`) on `run.sh`, `lib.sh`, this Result | 0 each |
| Allow-list (`git diff --name-only ded22f8c` + untracked) | 2 files (`infra/validation/run.sh`, this Result), **0 outside** |
| Harness-lib offline controls | not run: `infra/ci/validate-controls.sh` needs Docker for its controls and is outside the stack-free scope of this correction; B-CI-CONTROLS/B-SHELL run in the coordinator's batch 1 |

**Rerun required (coordinator).** Batch 3 must be rerun on this corrected head (or its reconciled successor) for the
failed checks **B-TENANCY-MIGRATION, B-BUSINESS-PROFILE-MIGRATION, B-POLICY-MIGRATION, B-RLS-MIGRATION and their
dependent B-REQUIRED-CHECKS**; the `migrations` suite is one run, so the rerun is the whole suite and must attribute this
source. The first run `s6-e7c4efc6/b3` stays recorded as **FAILED** with its evidence unchanged. Expected delta: the
three seeds now print `no pending migrations` / `current_version 3|4|5` (or apply nothing) and write `persist.ids`,
the verifies read it back after the restart; B-RLS-MIGRATION's `prov` capture is empty with the three provisioned
status lines still on the console. Nothing here ran against a real daemon or PostgreSQL. Budget: B5 attempt 1
continues; correction cycle 3 (S6); no new attempt.

### S6 correction 3, revised (Codex correction of the class A ruling; handoff 6af948ec)

Supersedes the class A part of the subsection above; classes B and C are unchanged. Writer, checkout, branch and
evidence as above. The grouped source correction `fdffa92c` (one commit on `g064-s6-corr3`: Migration → seed →
restart → verify on one database, `it_group_skip`) was **never run** against a stack and is **withdrawn**: Codex
corrected the ruling because carrying the Migration invocation's end state into the seed preserves order coupling.
This revision is the next commit on the same branch; the run.sh grouping is fully reverted (the three pair functions'
code equals `ded22f8c`'s, comments aside; `it_group_skip` removed; `it_fresh` before each seed restored exactly as in
Plan r2 §4, "`db_fresh` before the **seed** only"; the reset-point count stays **50**).

**Corrected Codex ruling (quoted from the dispatch):** "keep ONLY the intentional PersistSeed → `compose stop
postgres` / `up -d --wait postgres` (volume retained) → PersistVerify group. Each persistence group starts from its OWN
fresh copy (keep `it_fresh` before the Seed, exactly as in the Plan), and PersistSeed must establish its expected
historical schema version independently: use the existing `reset(t,cfg)` + bounded `Up` pattern in the migrate-package
integration tests where appropriate … Migration remains a separate fresh invocation; Catalog and all other independent
selectors stay isolated; the direct-selector campaign is unchanged." Allow-list extension (Codex disposition, limited):
"you may edit ONLY the Seed test setup (and the minimal Verify adjustment if the Seed change requires it) in
`core/internal/migrate/tenancy_integration_test.go`, `business_profile_integration_test.go` and
`policy_integration_test.go` — no assertion removed or weakened, no other test body changes".

**Every `*PersistSeed`/`*PersistVerify` pair in `run.sh` (confirmed before editing; `grep 'func Test.*Persist'` over
`core/`: no other pair exists in any package or suite):**

| Pair (`run.sh` flow) | Seed start expects | Verify expects | Fresh v19 clone before the seed |
|---|---|---|---|
| tenancy (`tenancy_migrations`) | `runnerFor(subsetFS(3)).Up` output ending `current_version\t3` — needs a database at version ≤ 3 holding only migrations ≤ 3 | `subsetFS(3)` `Version` = `current_version\t3`; rows by ID | **broke** (`no pending migrations` / 19; b3 `:28`) — fixed by `reset` |
| business profile (`business_profile_migrations`) | `subsetFS(4).Up` ending `current_version\t4` | `subsetFS(4)` `Version` = 4; profile and link rows | **broke** (b3 `:50`) — fixed by `reset` |
| policy (`policy_migrations`) | `subsetFS(5).Up` ending `current_version\t5` | `subsetFS(5)` `Version` = 5; `false,true` flags | **broke** (b3 `:72`) — fixed by `reset` |
| RLS (`rls_migration`) | `reset(t, cfg)` then `rlsRunner6.Up` ending `current_version\t6` (already independent) | `rlsRunner6` `Version` = 6; detector; runtime context A sees only org A | compatible (passed b3 `:97-98`) — unchanged |
| objects (`obj_migration`, `core/internal/tenancy/access`) | `setupObjects`/`setupIT` on the current full schema; a PutObject | guarded Get of the same bytes by a new reader; `unseed` | compatible (passed `:114-115`) — unchanged |
| audit (`audit_migration`) | `reset(t, cfg)` then `auditUp` ending `current_version\t8` (already independent) | `auditFS` `Version` = 8; detector; event row md5; owner reads it through RLS | compatible (passed `:129-130`) — unchanged |

Migration-subset compatibility of the fix: `reset(t, cfg)` (`integration_test.go:254`) is the same helper every
`Test*Migration` MG-1 subtest calls first on its fresh v19 copy in batch 3 (`tenancy-migration`,
`business-profile-migration`, `policy-migration`, `rls-migration`, `obj-migration`, `audit-migration` all passed), and
the same step `TestRLSPersistSeed`/`TestAuditPersistSeed` already take; it drops every object of versions 19…1 in
dependency order and `app.goose_db_version`, so the following bounded `Up` applies exactly migrations 1..N from the
subset filesystem. The seeds' own writes (`seedIdentity`, `CreateOrganization`/`CreateGroup`/`GrantOrganizationRole`,
`insertProfileAndLink`, `seedPolicyFixture`) run at version N as before; the verifies are unchanged and still compare
the runner's `Version` at the subset and the rows by the IDs the seed wrote.

**Exact Go diff (3 files, +3 lines each, nothing removed; `git diff ded22f8c -- core`):** in each `Test*PersistSeed`,
immediately before `r := runnerFor(t, cfg, subsetFS(t, N))`, two comment lines and one statement:

```
+	// G064: the seed runs on its own fresh copy of the current template; it establishes version 3 itself (reset to
+	// empty, then the bounded Up) instead of relying on TestTenancyMigration's end state.
+	reset(t, cfg)
```

(`tenancy_integration_test.go` N = 3, `TestTenancyMigration`; `business_profile_integration_test.go` N = 4,
`TestBusinessProfileMigration`; `policy_integration_test.go` N = 5, `TestPolicyMigration`.) Post-change expected
version per seed: tenancy **3**, business profile **4**, policy **5**; the `Up` now prints `applied 00001…0000N` /
`current_version N` (the unchanged `HasSuffix` assertion accepts it). No assertion removed or weakened; the verifies
are untouched; the `*_PERSIST_FILE` contract, the harness restart and the `go_it` rows are unchanged.

`run.sh` after this revision differs from `ded22f8c` only by class B (stderr status lines) and three MG-6/MG-BP-6/
MG-POL-6 comment lines naming the seeds' self-reset.

**Stack-free checks (this revision; no stack, database, network, pull, build or image substitution):**

| Command | Outcome |
|---|---|
| `bash -n infra/validation/run.sh` | OK |
| shellcheck B-SHELL flags (`-S warning -x -P SCRIPTDIR`, pinned local `koalaman/shellcheck:v0.11.0`, `--rm --pull never --network none`, ro) on `run.sh` + `lib.sh`; `-S info --include=SC2251` | rc 0; rc 0 |
| `docker run --rm --pull never --network none -v core:/src:ro golang:1.27.3-alpine gofmt -l` on the three test files (the project's pinned Go image, `infra/docker/core.Dockerfile`) | no output, rc 0 (formatted) |
| `docker run --rm --pull never --network none -e GOPROXY=off -e GOFLAGS=-mod=readonly -e GOTOOLCHAIN=local -e GOCACHE=/tmp/gocache -v core:/src:ro -v g058-a1-gomod:/go/pkg/mod:ro -w /src golang:1.27.3-alpine go vet -tags integration ./internal/migrate/` (an existing local module-cache volume mounted read-only; disposable in-container GOCACHE; no network) | **rc 0**, no findings (the cache held every module `core` requires; nothing downloaded) |
| Scratch logic test `/tmp/g064s6-corr3/test.sh` (revised; outside the repository, not committed) | **37/37**: class B as before (capture around `go_it`/`rls_provisioned_test`/`it_fresh` empty or exactly the `bad` part, base `go_it_convert` reproduces the leak); (A) each pair's exact sequence is `it_fresh <p>-catalog; go_it <p>-catalog; it_fresh <p>-migration; go_it <p>-migration; it_fresh <p>-persist-seed; go_it <p>-persist-seed; compose stop postgres; compose up -d --wait postgres; go_it <p>-persist-verify` with PASS recorded; each pair function's code equals `ded22f8c`'s with comments stripped; a refused `db_fresh` before the seed keeps the base fail-closed behaviour; `it_fresh` before all six `*-persist-seed` present; in each seed `reset(t, cfg)` immediately precedes `runnerFor(t, cfg, subsetFS(t, N))`; the Go diff is exactly three added `reset(t, cfg)` lines and no removal; the `infra/` diff is `run.sh` only |
| `git diff --check`; U+2028/U+2029 on the 5 touched files | clean; 0 each |
| Allow-list (`git diff --name-only ded22f8c` + untracked) | 5 files: `run.sh`, this Result, and the 3 test files under the Codex extension; **0 outside** the extended allow-list (3 outside the original one, all covered by the extension) |

**Isolated rerun pending (coordinator).** Batch 3 (`migrations` suite) must be rerun on this revised head for
B-TENANCY-MIGRATION, B-BUSINESS-PROFILE-MIGRATION, B-POLICY-MIGRATION, B-RLS-MIGRATION and B-REQUIRED-CHECKS; the
`e7c4efc6` run `s6-e7c4efc6/b3` stays FAILED and unchanged; the grouped commit `fdffa92c` was never run. Expected
delta: the three seeds print `applied 00001…0000N` / `current_version N` after their reset and write `persist.ids`;
the verifies read it back after the restart; B-RLS-MIGRATION's `prov` capture is empty. Budget: B5 attempt 1
continues; correction cycle 3 (S6); no new attempt.

### S6 batch 4 driver

Writer: one bounded writer, Claude Opus 5.5 (`claude-opus-5-5`) under the standing Fable-limit fallback (Codex record
"Confirmed Fable limit", checkpoint 127); effort requested xhigh by the dispatch, runtime effort not observable from
inside the agent (**unknown**). Checkout: isolated worktree `agent-<id>`, branch `g064-s6-b4b` from
`origin/g064-s6-b4` `25ab7a76` (verified; the stopped Fable writers' worktree had no partial edits), child of
`2d25cf7a`. Commits `025aafc2` (driver body), `478a1243` (receipt/scan-status fix found by the scratch test) and this
Result/PG-010 commit. Driver blob `20cc3be59a062213e3d745458efeafa19666d7ac`, mode `100755`, 723 lines; `run.sh` blob
`4fc59cd7` (sha256 `adbcfd30…abb2d`), unchanged since `2d25cf7a`. Plan bytes unchanged. **Not run against a stack.**

**What it runs (Plan r2 §3.5, §6 batch 4).** Entry point `management/v0/results/G064-s6-b4-identity-driver.sh
CHECKOUT EVIDENCE_DIR`, launched by the coordinator as `flock -n
<home>/.local/state/platform-evidence/local-harness.lock env G064_HARNESS_LOCK=held bash <driver> <frozen
checkout> <evidence dir>` (PORTS_IDN are fixed host ports, so the shared harness lock is required; the driver refuses
unless the marker is `held` AND a `flock -n LOCK true` probe fails, i.e. someone actually holds the lock; it never
creates or takes the lock). It runs the documented `identity:start`/`identity:migrate` blocks, the §3.5 isolation window
(stop core/core-replica, zero client sessions, the overlay's own `migrate status` → `db_pristine_verify` →
`db_template_create`), `identity_it` through `go_it` with the per-package expectations (`identity`, `identity/workos`,
`identity/identitytest`), `db_fresh identity-it-after`, the core restart, `mcp_seed` as the first writer, then
`identity:browser` restricted to `--project=mcp-light --grep "M-LIST"`, then `identity:stop` — all on the real `run.sh`
code paths.

**Derived-copy approach.** `run.sh` is not sourceable (top-level parsing, EXIT trap, suite loop, final checks), so the
driver writes a derived copy outside the repository (`<scratch>/run-derived.sh`) that differs from `run.sh` in exactly
three `diff -U3` hunks: H1 line 33 `REPO=` → the frozen checkout; H2 line 61 the accepted-suite case gains
`|identity-focused` (parser only); H3 a 95-line block inserted immediately before the suite loop (`CI_CONTROLS_BG=0`,
line 6445) defining `suite_identity_focused`, `variants_identity_focused` (= `variants_identity` lines 5669-5700 through
`mcp_seed "$D"` verbatim, then `identity_e2e_focused "$D"`, then lines 5706-5710 verbatim; the five replaced calls at
5701-5705 must be exactly `identity_e2e`, `identity_probes`, `mcp_e2e`, `identity_privacy`, `identity_config`) and
`identity_e2e_focused` (identity_e2e's exact one-off substitution line, required to occur exactly once in `run.sh`,
with the trailing project chain replaced by `--project=mcp-light --grep "M-LIST"`; read-only `psql_super` count of the
seeded old-host row `app.mcp_connections.host = 'old.acme.test:543'` must be 1 before the browser run; real
`ledger_watch_*`, `capture_after run --build`, `browser_collect_report`, `browser_matrix "mcp-light|M-LIST S-ERRORS"`;
distinct ID `B-IDENTITY-E2E-FOCUSED`, never `B-IDENTITY-E2E`). Every structural fact is re-checked at run time and fails
closed (exit 2): one anchor line each, the inverse (block dropped, H1/H2 restored) `cmp`-equal to `run.sh`, `bash -n`,
exactly the three new functions, none pre-existing in `run.sh`. The diff, both sha256 values and the derived copy itself
are written to the evidence (`run-derived.diff`, `run-derived.sha256`, `run-derived.sh`; the derived sha256 depends on
the checkout path in H1, so it is recorded per run).

**Deviation from `suite_ops` (Plan D-4, §6 batch 4; recorded, not new).** `ops` itself (variants_ops, the full
`identity:browser` chain, `B-IDENTITY-PROBES`/`-PRIVACY`/`-CONFIG`, `B-MCP-E2E`) runs post-merge in the full
qualification; batch 4 covers the window, `identity_it` and the one browser seed consumer. `go_it_declare ops` (part
of the verbatim prefix) declares `identity-it` and `mcp-e2e`; `mcp-e2e` does not run by design, so this run is not a
`validate.sh` `complete`-step candidate (it is driven by `run.sh` directly).

**Differences from the committed design header (corrected in the header).** (1) Dockerfile `FROM` bases are recorded in
`preflight.tsv` but are not a refusal: `gcr.io/distroless/static-debian12:nonroot` is not in `docker image ls` on this
host (BuildKit's own cache), and BuildKit resolves `FROM` tags itself, so a presence refusal would be false; the run
images (`image:` lines of `local.yml`/`identity.yml`/`browser.yml`, `YQ_IMAGE`, `BUSYBOX_IMAGE`) must be present locally
and nothing is pulled. (2) `identity_e2e_focused` registers observed values and moves handoffs exactly as
`identity_e2e`, but requires neither (M-LIST calls neither `sessionValue` nor `csrfOf` and writes no handoff; requiring
them would fail a correct run). (3) The start inventory uses `docker image ls -aq` as `validate.sh`'s
`ci_docker_start_inventory` (the S4 driver used `-q`; a superset, more conservative). (4) The harness console goes to the
scratch directory (live: `tail -f`) and only its redacted copy enters the evidence. (5) Added: an evidence secret scan
with a planted positive control (a matching file is withheld and the run fails) and a registered `--rm --pull never
--network none` busybox fallback (cidfile in the registry's `cid/`) for container-owned files when removing the kept
`$WORK`. (6) A signal to the driver is forwarded once as TERM to the harness process group (`setsid`), then the driver
waits for the harness's own EXIT-trap cleanup (no KILL escalation).

**Resources.** Compose project `g001-idn-<run id>` on PORTS_IDN `3600 8680 8600 6032 6979 9600 9601` (run.sh's own
`ports_busy` refusal applies) plus the proxy's random loopback `:543` publish; no other published port; images of the
documented blocks and `ensure_tools`; the `identity-run` volume; one build slot (`COMPOSE_PARALLEL_LIMIT=1`); the shared
harness lock for the fixed ports. Registry in the `validate.sh` shape created before `run.sh` starts (`--owned-registry`),
finalized by `infra/ci/cleanup-owned.sh --mode finalize`, then the S4 exact-ID presence check; receipt `cleanup.log`
(verdict CLEAN/LEFTOVER/UNKNOWN; a non-CLEAN registry is retained for recovery). Exit codes: 0 pass; 1 check failure (a
leaked secret too); 2 preflight/infrastructure stop; 4 passed but LEFTOVER; 5 passed but cleanup UNKNOWN; 130/143
interrupted; an earlier non-zero status is kept.

**Extracted outcomes (`s6b4-summary.tsv`; all required for exit 0).** `run.sh` exit 0; `B-IDENTITY-IT` exactly once
PASS with the window facts in its text ("core and core-replica stopped with 0 client sessions", "the overlay's own
migrate status and pristine check", "pristine row in db-lifecycle.tsv", "recreated from it and core restarted before
mcp_seed", the per-package expectations); `db-lifecycle.tsv`: one `pristine` PASS row with `caller=identity-overlay`
(its checks include zero sessions), one `template create status=PASS`, ≥ 1 `template-record`, one `fresh` PASS row
`caller=identity-overlay purpose=identity-it-after`, no `refuse`/FAIL row; `go-tests.tsv`: the `identity-it`
invocation row pass/rc 0/passed = expected, package rows `identity` and `identity/workos` pass with passed = expected
≥ 1 and `identity/identitytest` `skip-no-test-files` (exactly three; Plan §4 #53 base figures 74/21/0 reported, derived
not pinned); `go-tests/identity-it.log`; `go-tests-expected.tsv` ops = `identity-it,mcp-e2e`; the only row of
`browser-identity-focused-tests.tsv` = M-LIST in mcp-light, `expected`; old-host rows before the run = 1 (from the
`B-IDENTITY-E2E-FOCUSED` text); `B-DOCKER-ONLY`, `B-SECRET-CANARY`, `B-SCOPE` (0 changed paths; the allow file holds no
pattern), `B-REQUIRED-CHECKS` PASS and no FAIL/UNMET row; evidence scan clean; cleanup CLEAN. Also recorded: PG-010
figures of this run (template size/duration, the single `ops` `db_fresh`).

**What it cannot prove.** The full `identity:browser` matrix (`B-IDENTITY-E2E`), probes, privacy, config, `B-MCP-E2E`
and `variants_ops` (post-merge `ops`); that dependency fetches during image builds come from cache (BuildKit metadata
resolution and `go mod download`/`npm ci`/`uv sync` behave as in every harness run); a `validate.sh` publication of
this run.

**Stack-free checks (writer; no stack, database, build, pull or harness run; the real harness lock never probed):**

| Check | Outcome |
|---|---|
| `bash -n` on the driver | OK |
| `shellcheck -S warning -x -P SCRIPTDIR` (local pinned `koalaman/shellcheck:v0.11.0`, `--rm --pull never --network none`, read-only mount) | rc 0, no findings (`-S info`: SC2015 ×2 and SC2317 notes only, reviewed; one SC2016 note silenced as literal `run.sh` text) |
| `git diff --check`; U+2028/U+2029 | clean; 0 |
| Scratch logic test `/tmp/g064-s6b4-scratch/test.sh` (outside the repository, not committed; docker and flock PATH shims; a scratch lock file for the real-`flock` cases), clean checkout at `478a1243`, driver blob `20cc3be5` | **68/68 PASS**: usage (0/1 args), non-directory, non-top-level, non-empty evidence and evidence-inside-checkout refusals (exit 2, nothing created); marker absent/wrong/lock free → exit 2; lock check with real `flock` on a scratch file (absent, free, held by another holder accepted, marker unset, symlink); derivation on the real `run.sh`: success, `bash -n`, exactly 3 hunks, only the REPO and suite-case lines removed, `REPO="<checkout>"`, `identity-focused` only in the derived copy, block immediately before `CI_CONTROLS_BG=0`, independent inverse `cmp`-equal, exactly the 3 functions, layout 5668/5700, prefix 5669-5700 and suffix 5706-5710 verbatim, 5701-5705 the five replaced calls, the suite reachable by `suite_${s//-/_}`; negative derivations refused (extra line among the replaced calls, two loop anchors, a changed one-off substitution, a path with a space); `identity_e2e_focused` with stubs (docker = recording shim): the executed line is `identity:browser` with identity_e2e's substitution ending `web-browser npx playwright test --project=mcp-light --grep "M-LIST"`, one `--project=`, no `--rm`, compose files/profile as documented, the read-only old-host query, ledger watch/capture/collection/matrix calls, PASS text with the count; MCP-seed failure and missing old-host row → FAIL with no browser run; registry: unreadable start inventory fails (no empty inventory), validate.sh shape; cleanup with Docker unreachable and a registered container → verdict UNKNOWN, exit 5, the container listed by exact ID (`failed`), earlier status 1 kept; main path: old Compose → exit 2 with the derived copy recorded, no registry → CLEAN (nothing owned), scratch removed, only the lock probe and `compose version` issued; run image absent → exit 2, no pull/run/build; start-inventory listing failure → exit 2, receipt "registered resource rows: 0 (start inventory not recorded …)", only `compose version`/`image inspect`/the failing `ps -aq` issued, `run.sh` never started; extraction on a passing fixture → no failure, on a negative fixture → missing `fresh` row, a refusal, a `workos` failure, an extra browser row, a missing window fact and old-host 0 all reported; evidence scan → PASS with the planted control, a leak → FAIL and the file withheld |
| Same test, dirty checkout (this Result and `docs/PERFORMANCE.md` uncommitted), driver blob `20cc3be5` | **56/56 PASS**: the refusal, lock, derivation, `identity_e2e_focused`, registry/cleanup, extraction and scan cases above plus "dirty checkout → exit 2, no evidence directory created" (the main-path preflight cases need a clean checkout and run in clean mode) |
| Allow-list | 3 files vs `25ab7a76`: the driver, this Result, `docs/PERFORMANCE.md`; 0 outside |

Recorded during the first clean-mode run: 4 expectations of the start-inventory-failure case assumed verdict UNKNOWN;
the finalizer correctly reported CLEAN because the registry held no resource row and no start inventory (nothing could
have been created). The expectations were corrected and the receipt now states the registered row count and whether the
start inventory was recorded (`478a1243`); no driver assertion was weakened.

**Run pending (coordinator).** Batch 4 runs this driver once on the frozen checkout of the reviewed head, holding the
shared harness lock (`flock -n`), one pool slot and the build slot, with an absent or empty evidence directory under
`platform-evidence/g064-a1/`. Its `s6b4-summary.tsv`, `report/`, `cleanup.log` and the PG-010 row are recorded here
afterwards. Budget: B5 attempt 1 continues; no new attempt; correction cycles unchanged.

### S6 timing disposition

**Corrected S4 campaign at `17fdad66`** (coordinator run 2026-10-22 02:21:40Z-02:35:05Z, evidence
`platform-evidence/g064-a1/s6-17fdad66/s4/run/`: `s4-summary.tsv`, `db-lifecycle.tsv`, `go-tests.tsv`, `s4-phases.tsv`,
`cleanup.log`; S4 result PASS: campaign 95/95, I 70/70, L 25/25, AC3 12/12, B-IT-ISOLATION PASS, no published port,
cleanup CLEAN). Recomputed by the writer, read-only, with an inline `awk` over those files (per selector: the
`db-lifecycle.tsv` `fresh` row with `purpose=<invocation>` → raw `db_fresh` `total_ms`; `s4-summary.tsv` `fresh_ms` →
the `it_fresh` wrapper, bounded drain included; `go_ms` → the invocation it precedes):

| Figure | Value |
|---|---|
| Template | `pg_database_size` 9,740,815 bytes; `db_template_create` 2,126 ms |
| `db_fresh`, all 109 recreations | all PASS, no residue; total min/median/max 1,441 / 1,694 / 2,997 ms; medians: checks 599, drop 159, create 166, postconditions 733 ms; refusals 6 (the deliberate controls) |
| Campaign raw `db_fresh` (95) | min/median/max 1,441 / 1,663 / 2,997 ms; sum 167,452 ms against 357,678 ms of invocation time |
| Campaign wrapper (`it_fresh`) | sum 192,296 ms |
| Raw `db_fresh` longer than the invocation it precedes | **12/95**, maximum excess **948 ms** |
| Wrapper longer than the invocation it precedes | **20/95**, maximum excess **1,141 ms** |
| AC3 (12) | all PASS; none shorter than its recreation (raw or wrapper) |
| Phases | campaign 551,336 ms, AC3 146,651 ms, all phases 792,686 ms (wall clock 805 s) |
| Supplementary basis (not Codex's) | against the tests' own reported elapsed time (top-level `test` rows of `go-tests.tsv`, sum 166,640 ms; excludes `go test -list`, the cached build/link and container exec), raw recreation is longer for 76/95, maximum 2,667 ms |

**Verification against Codex's figures:** identical on Codex's basis (raw 1,441/1,663/2,997 ms; 167,452 vs 357,678 ms;
12/95, max 948 ms; wrapper 20/95, max 1,141 ms, total 192,296 ms; AC3 12/12 PASS, none shorter than its recreation;
cleanup CLEAN). No difference; the supplementary test-elapsed basis is added for completeness and shows the per-test
isolation cost is of the same order as a single short test's own run time.

**Codex disposition (2026-10-22, `G064-s3-s4-source-codex-review-20261022.md`, "Timing and continuation
disposition"):** bounded continuation approved with a recorded performance limitation in PG-010; this is measured
isolation cost, not a correctness failure or permission to weaken resets; no optimization, new benchmark campaign or
repeat of passing S4 is required to clear the stop; materially different later observations are escalated through the
risk register. **Superseded:** the S4 rerun/S5 record's per-selector figures from the `03e50356` run ("5 of 95
selectors, by at most 73 ms"; "with the bounded drain poll included, 23 of 95") are superseded by the figures above;
that text is kept unchanged as history. PG-010 (`docs/PERFORMANCE.md`) gains a dated "Corrected S4 campaign at
`17fdad66`" observation; its status stays open, observed, not benchmarked, pending the batch-4 identity-window
observation and Codex review. **`_iti_t8`/summary text:** deferred per Codex; no code edit; the disclosure in the S4
rerun record stands.

### S6 results, S7 self-review and S8 bindings (coordinator, 2026-10-22)

All S6 batches are recorded here with the exact tested source of each. Evidence lives outside the repository under
`~/.local/state/platform-evidence/g064-a1/` (sha256 prefixes below are of the primary evidence file). Original failed
runs stay recorded as failed; nothing below replaces them.

| Batch | Tested source | Result | Evidence (sha256 prefix) |
|---|---|---|---|
| 1a stack-free harness suites (`run.sh lint lint-base lint-core unit-core`) | `e7c4efc6` (first rebase onto staging `f9048e8a`) | PASS, 38 PASS rows, `RESULT: PASSED` | `s6-e7c4efc6/b1/static.console.txt` 005cf2d2 |
| 1b full `validate-controls.sh` (527 controls) | `e7c4efc6` | **FAILED 521/6**: VA7, VU1, VU10, J8, R3a, R8a hard-code the old gate/union counts 92/123, while S5's B-IT-ISOLATION row makes them 93/124 (the S5 handback's "counts derive" held for `run.sh` and `B-CI-CONFIG`, not for `validate-controls.sh`) | `s6-e7c4efc6/b1/controls.console.txt` 5429e251 |
| 1b-rerun focused controls (the six, GT1, GT2, VB1-VB6, V1, V3) after the mechanical pin correction (`validate-controls.sh`, 9 lines; no logic changed) | `2d25cf7a` | PASS 17/17, exit 0 (independently inspected by Codex) | `s6-2d25cf7a/controls/controls.console.txt` f62b3155 |
| 3 `migrations` suite once (D-4), harness lock | `e7c4efc6` | **FAILED**: B-TENANCY-MIGRATION, B-BUSINESS-PROFILE-MIGRATION, B-POLICY-MIGRATION (version-scoped PersistSeed on a fresh version-19 copy: `up: no pending migrations current_version 19`), B-RLS-MIGRATION (all `go_it` invocations rc=0, but `go_it`'s status line leaked into the `$(rls_provision ...)` capture), B-REQUIRED-CHECKS (dependent). The end-of-run `container adoring_yalow` note was a foreign container from a concurrent writer's shellcheck run (no G064 path creates an unnamed container) | `s6-e7c4efc6/b3/run.console.txt` 6d3972c5 |
| 3-rerun `migrations` suite once, harness lock, after correction 3 (seeds `reset(t, cfg)` to their historical version on their own fresh copy; status lines to stderr; Codex disposition 6af948ec) | `17fdad66` | PASS: 50 PASS rows, `RESULT: PASSED`, exit 0, no leftover note | `s6-17fdad66/b3/run.console.txt` c7d9dd95 |
| S4 corrected campaign (driver with exact-ID owned inventory and fail-closed Docker checks; Codex source review f42317b0) | `17fdad66` | PASS 95/95 selectors, AC3 12/12 (4 reversed pairs, repeat 2/2, shuffle 2/2), `B-IT-ISOLATION` PASS, 115 PASS rows, 0 non-PASS; cleanup `CLEAN` (removed=10 gone=21 leftover=0 unknown=0, exact IDs in the receipt); `TestEnrollmentCatalog` runs with `ENROLLMENT_REPLAY_SQL_B64` | `s6-17fdad66/s4/run/s4-summary.tsv` ce36cdd4; `cleanup.log` 2583fc57 |
| 4 identity overlay focused (documented `identity:start`/`identity:migrate`, the §3.5 isolation window, `identity_it`, `db_fresh`, core restart, `mcp_seed`, then `--project=mcp-light --grep M-LIST`), harness lock | `0696791f` (final rebase onto staging `b66930a1`) | PASS in 181 s: `B-IDENTITY-IT` 95/95 top-level tests (identity 74/74, workos 21/21), window facts (zero client sessions, own pristine check, template, recreated copy, core restarted before `mcp_seed`), `B-IDENTITY-E2E-FOCUSED` PASS, `B-SECRET-CANARY`, `B-DOCKER-ONLY`, `B-SCOPE` PASS; cleanup `CLEAN` (leftover=0 unknown=0) | `s6-0696791f/b4/s6b4-summary.tsv` b153d80f; `cleanup.log` f325d3ea |

Binding statement. The Go, harness and test sources of `17fdad66` (S4 campaign and batch 3) are byte-identical to the
final head `0696791f` apart from `infra/ci/validate-controls.sh` (pin parity, covered by the 2d25cf7a controls),
management/docs files, and the batch 4 driver; so those results carry to the final head without repetition (CLAUDE.md
targeted reruns). Batch 1a ran on `e7c4efc6`; `run.sh`/`lib.sh` changed afterward only by the `go_it` status-line
stderr routing and comments, and batch 3, S4 and batch 4 executed that code. Plan §10's timing stop was dispositioned by
Codex (7e61e18e) as bounded continuation with an honest PG-010 limitation (see `S6 timing disposition`).
The merged-SHA local full qualification still runs after merge and is the proof for everything not exercised above
(full browser matrix, `variants_ops`, `B-MCP-E2E`, hosted gates).

Deviations and amendments recorded in this Result (none architectural; Plan r2 bytes unchanged):
- S4 correction 1 and 2, the class-A persist-seed correction 3 (Codex first approved grouping Migration with Seed,
  then corrected it before any run: do not carry Migration state into Seed; the grouped attempt `d287be4c`/`fdffa92c`
  was never run and is superseded by `4e20bffe`/`17fdad66`), the controls pin parity, the batch 4 focused driver in
  place of the `ops` suite (Plan D-4/§6 batch 4), GT1's retained V3 partial publication labelled as a deviation from
  Plan §5.4, and the sign-host selector accepted only as the URL-issuance half (the actual host probe stays in
  `B-OBJ-IT`, which the migrations/ops suites run).
- Allow-list extension (Codex 6af948ec, limited): `core/internal/migrate/{tenancy,business_profile,policy}_integration_test.go`
  seed setup only (one `reset(t, cfg)` plus two comment lines each); `infra/ci/validate-controls.sh` text/pin parity;
  `management/v0/results/G064-s6-b4-identity-driver.sh`.
- Deferred by Codex: the cosmetic `_iti_t8`/summary-text issue (no code edit or rerun). Observation only: in the
  batch 4 summary the window fact row for per-package expectations prints an empty detail; the go-tests rows carry the
  data (identity 74/74, workos 21/21).

#### S7 self-review (coordinator; stack-free reading plus the recorded evidence)
- Diff against staging `b66930a1`: 23 files, +5485/-176, `git diff --check` clean, no U+2028/U+2029. **No production
  path touched** (no `core`, `web`, `intelligence`, `contracts` or Compose file outside tests, harness, docs and
  management files).
- **No assertion, count or skip changed outside approved lines:** the Go test diff is additions only (30 lines: the one
  tenancy `setup` starting-version assertion and its three purge-helper `defer c.Close`, plus three `reset(t, cfg)` seed
  setups); zero removed lines; every other selector runs unchanged through `go_it` + `db_fresh`.
- Required checks: 130 rows, full profile 123 (union 124 with `ci-controls`), fast 35 / integration 75 / staging 5
  unchanged; pins in `validate-controls.sh` corrected and rerun (above).
- Privacy of new evidence: `go-tests/` logs and exports pass the harness's generated-secret redaction; `B-SECRET-CANARY`
  PASS in batch 1a and batch 4; the failed-run diagnostics were exported redacted. Private runs published no host port.
- Cleanup: S4 and batch 4 cleanup receipts `CLEAN` with exact IDs; batch 3 reruns end with no leftover note; no
  G001/G064-labelled container, volume, network or image remains on the host after the batches.
- Open items for Codex source review: the pending S3 deviations (the `-v` re-layout, GT1/GT2 readings), the S4 deviations
  (`TestObjectsSignHost` Go-test-only half, container-local TENANCY files, `docker exec` wrapper, shuffle via `GOFLAGS`,
  tools image built directly, T-7 text duplication), the extra deviations above, and PG-010's open performance limitation.

#### Budget
B5 attempt 1 continues (no new goal-level attempt). Correction passes this slice: S4 corrections 1 and 2, S6 correction 3
(one superseded before execution), one mechanical pin correction. Worker effort: the Fable 5.1 writers' and Opus 5.5
fallback writer's actual runtime effort is unknown (not observable inside the agent); Fable stopped at its quota
(API 429) and the standing Opus 5.5 xhigh fallback was applied per Codex (b66930a1), recorded in the checkpoint.

### S6 source-review correction 4 (Codex, candidate 08b109b8)

Writer: one bounded writer, Claude Opus 5.5 (`claude-opus-5-5`) under the standing Fable-limit fallback; effort
requested xhigh by the dispatch, runtime effort not observable inside the agent (**unknown**). Checkout: isolated
worktree `agent-<id>`, branch `g064-mr2` from `origin/g064-mr` `08b109b8` (verified). Allow-list: the
driver and this Result (append-only). Driver commit `5611aadd`: blob `20cc3be5` → `e2c29f92`, mode `100755`, 723 → 859
lines. `run.sh`, `lib.sh`, tests, docs and the Plan are unchanged.

**Issue (Codex review of candidate `08b109b8`, CORRECTION LOOP).** The driver suppressed `redact_log` errors at lines
489, 496, 499 and 502, and the collection row still said PASS (503). The finalizer then removed `$WORK` (633) and,
after a CLEAN cleanup, the scratch directory (645), so a failed export could discard its only source. The generated-secret
scan also treated a grep error as no hits (605/608). Both break AC5/AC6 (retained, fail-closed failure evidence).

**Fix (driver only; line numbers at `5611aadd`).**

| Site | Before | After |
|---|---|---|
| `s6b4_export` (new, :509) | — | Every redacted copy is verified: the source must be a readable regular file (symbolic links are not followed), `redact_log` must return status 0, and the destination must then be a regular file, non-empty when the source is. The status alone proves nothing, because `redact_log` ends with `rm -f`, so a failed `mv` still returns 0. On failure the partial output (destination and `.redacting`) is removed and the failure is recorded in `S6B4_EXPORT_FAILS` |
| `s6b4_collect` (:531, was 489-503) | `redact_log … \|\| s6b4_log` / `\|\| true`; the row said PASS regardless | Every copy goes through `s6b4_export`. The redaction input must be a readable regular file: lib.sh's awk skips a directory with only a warning and status 0, so nothing is redacted (checked with the host gawk). More than one work directory, or a `driver-logs/` that cannot be created, also counts as a failure. `s6b4_collect_row` (:572) writes FAIL with every failure listed, never PASS |
| `s6b4_main` (:843) | — | Export failures are added to the FAIL reasons: the verdict is FAIL and the exit is 1 |
| `s6b4_rm_work` (:581; the removal at 633) | always removed `$WORK` | Keeps `$WORK` when any export failed (`collect work-removal RETAINED`) |
| `s6b4_cleanup` (:379-383) | the registry copy was unchecked | A failed copy to `owned-registry.tsv` counts as an export failure and gets a receipt line |
| `s6b4_scan` (:679, was 602-608) | `sort \| grep -v \|\| true`; `grep … 2>/dev/null \|\| true` | A `sort` failure is UNKNOWN. Empty lines are dropped with `sed '/^$/d'`, which has no "nothing selected" status. For the control and evidence greps, exit 0 means a hit, 1 no hit, and ≥ 2 an error, reported as UNKNOWN (status 2) and never PASS. A leak found alongside an error stays FAIL with the error noted. A withholding overwrite that fails is reported |
| `s6b4_retain` (new, :721) and `s6b4_finish` (:742, was 645) | the scratch directory was removed after a CLEAN cleanup | When an export failed, or the scan is FAIL or UNKNOWN, the scratch directory is kept (mode 700, contents go-rwx: `$WORK`, the raw console, `secrets`/`secrets.u`, the registry) and never copied raw into the evidence. `cleanup.log` records `retained private source: <path> (directory mode 700 …)`, the reason, each export failure, the retained entries and four `recovery:` lines: redact with lib.sh and the run's secrets file; rescan with grep's exit semantics; remove the directory, including through the busybox fallback; a `/tmp` TMPDIR may not survive a WSL restart. The summary gets a `retention scratch RETAINED` row. If the PASS verdict was already written, a `result S6B4 FAIL "after the verdict: …; the result row above is superseded"` row follows |
| `s6b4_main`, derived-copy diff (:824) | `diff … \|\| true` | Requires diff exit 1 and a non-empty file; anything else stops with exit 2 |

Runtime-resource cleanup (the finalizer with `--mode finalize`, the exact-ID check and the `cleanup.log` receipt) is
unchanged and runs whatever the collection or scan outcome. Files withheld by the scan keep only the marker, as the
retention policy requires: never copy unsafe raw files to get around withholding. The scan still runs after `$WORK` is
removed and after the cleanup, so `cleanup.log` is still scanned and the busybox fallback stays covered by the
finalizer's receipt. As a result, a scan FAIL or UNKNOWN on a run whose exports all verified keeps the redaction inputs
and the registry but not `$WORK`, whose redacted copies are already in the evidence.

**Exit-code precedence (header, `s6b4_finish`).**
1. The status on entry to the EXIT trap is kept whenever it is non-zero; no later step replaces it. That status is
   `s6b4_main`'s verdict (0/1/2), a stop, an errexit, or 130/143 after a signal.
2. Only when the entry status is 0 does a failure found by the finalizer give 1. Such failures are an export failure
   (the registry copy included) or a scan FAIL or UNKNOWN.
3. Otherwise cleanup UNKNOWN gives 5 and LEFTOVER gives 4 (a leftover `$WORK` too); else the exit is 0.

Exits 4 and 5 are qualified passes, so step 2 outranks them, as it already did for the scan.

**Not changed (scope).** These are untouched:
- The derived copy and the three-hunk derivation. The regression comparison below produced an identical
  `run-derived.diff`.
- The `identity_e2e_focused` text. Its `grep -c … || true` (line 222 at `08b109b8`) belongs to the derived copy.
- The remaining suppressed sites, because none of them gates a PASS or fails open:
  - logging (`s6b4_log`, the inventory counts), signal handling and the informational FROM-base inspection;
  - preflight `compose version`, where an empty value is refused;
  - the derivation's `s6b4_line_of` and hunk count, where an empty result is refused;
  - the derivation's "run.sh already defines" check, which reads run.sh only after the same function has read it
    successfully and `cmp`-verified it;
  - extraction reads, where an empty value gives a FAIL row.

**Checks (writer, stack-free; no stack, database, Compose, browser, build, pull, harness run or real harness-lock
probe):**

| Check | Outcome |
|---|---|
| `bash -n` on the driver | OK |
| shellcheck `-S warning -x -P SCRIPTDIR` (local pinned `koalaman/shellcheck:v0.11.0`, `--rm --pull never --network none`, read-only mount) | rc 0, no findings (`-S info`: SC2015 ×2 and SC2317, all pre-existing) |
| `git diff --check 08b109b8 HEAD`; U+2028/U+2029 | clean; 0 |
| Scratch test `/tmp/g064-s6b4-c4/test.sh` (outside the repository, not committed): docker PATH shim, fake `s6b4_run` harness, a scratch lock held by the test shell and probed with the real `flock`; clean checkout `5611aadd`, driver blob `e2c29f92` | **95/95 PASS** (cases below) |

Scratch cases:
- **A (9), refusals with the real driver file.** 0 or 1 arguments, a checkout that is not a directory, a checkout that
  is not a Git top level, a non-empty evidence directory, an evidence directory inside the checkout, and the lock
  marker absent or wrong: each exits 2, creates nothing and makes no docker call.
- **B (5), lock check with the real `flock`.** A lock held by another holder is accepted; a free lock, an absent lock,
  a symlinked lock and an unset marker are refused.
- **C (7), derivation on the real `run.sh`.** It succeeds and passes `bash -n`, with exactly 3 hunks. Only the REPO and
  suite-case lines are removed, `REPO="<checkout>"` is set, and `identity-focused` appears only in the copy. A path
  with a space is refused.
- **D (2), start inventory.** A `docker ps` failure is refused, and no empty inventory is written.
- **E (3), cleanup with Docker unreachable and a registered container.** The verdict is UNKNOWN with exit 5, the
  container is listed by exact ID as `failed`, and an earlier status 1 is kept.
- **F (16), function-level cases.**
  - Export: ok (redacted); unreadable source; symlinked source; `redact_log` failure on an unwritable destination (no
    `.redacting` or destination left behind); status 0 without a destination; empty destination.
  - Scan: no hit (grep exit 1) → PASS. An unreadable evidence file (grep exit 2) → UNKNOWN, status 2. A leak plus an
    error → FAIL with the error noted. The leaked file is withheld (marker only). A leak alone → FAIL with the
    unchanged text. An unreadable secrets list → UNKNOWN. No secret registered, or only empty lines → status 0
    without a PASS.
- **G (53), main path (`s6b4_main` with the fake harness).**
  - Success: exit 0; result, collect and scan PASS; cleanup CLEAN; no RETAINED, FAIL or UNKNOWN row; scratch removed;
    redaction applied; exit row 0.
  - **Regression against the `08b109b8` driver** on the same success fixture: the same exit 0, an identical evidence
    file list (21 files), identical summary section/name/status columns, byte-identical redacted copies, and identical
    `cleanup.log` and collect/scan detail text (paths normalized).
  - Export failures: a read failure (chmod 000 log), a redaction failure (read-only `driver-logs/`), `redact_log`
    status 0 without output, the redaction input as a directory (nothing exported; scan UNKNOWN), and two work
    directories. Each gives exit 1, a collect FAIL naming each failure and a result FAIL. `$WORK` and the scratch
    directory are kept with mode 700 and `secrets` at 600, the retention row and recovery note are in `cleanup.log`,
    no raw secret is in the evidence, and no `.redacting` file is left behind.
  - A scan error after a clean collection gives exit 1 and scan UNKNOWN, and a FAIL row supersedes the PASS verdict.
    The scratch directory is kept with the secrets; `$WORK` is removed.
  - A leak gives exit 1; the file is withheld and the scratch directory kept.
  - Cleanup continues after a collection failure: the registered standalone container is removed by exact ID
    (`rm -f -v <id>`), the receipt is CLEAN, the exit is 1 and the scratch directory is kept.
  - Earlier status kept: run.sh exit 2 plus an export failure gives exit 2 (STOP); signal 143 plus an export failure
    gives exit 143 (INTERRUPTED).
  - Cleanup LEFTOVER alone gives exit 4 with the registry kept and no retention row, the same rows and exit as the
    `08b109b8` driver. LEFTOVER plus a scan error gives exit 1.

The first test run passed 91 of 95. All four failures were test-expectation errors, not driver faults: the
collect-detail comparison did not normalize the random `project-validation-*` suffix, and the "removal refused" fixture
yields LEFTOVER/4, not UNKNOWN/5, for both the new and the `08b109b8` driver. The expectations were corrected with no
driver change, and the rerun passed 95/95.

**Limits.** Only these stack-free scratch checks (shimmed Docker, a fake harness, injected faults) exercised the changed
error paths. The passing real batch-4 run took only the success path. lib.sh's internal truncation write (the
`> 1 MiB` branch) is not separately verified beyond status, presence and non-emptiness.

**Rerun disposition (Codex).** As Codex's candidate-`08b109b8` review states, this driver-only correction requires no
rerun of S4, the migrations suite or batch 4. The candidate's executable source is identical to tested `0696791f`, and
harness, Go and test sources are untouched here. This correction changes only the driver's collection, scan and
cleanup error handling; the regression comparison above shows the success path unchanged. The passing batch-4
evidence at `0696791f` therefore stays valid. Budget: B5 attempt 1 continues; this is the single correction pass Codex
authorized (correction 4); no new attempt.

### S6 source-review correction 5 (Codex delta re-review of 112cddb4)

Writer: the same bounded writer, Claude Opus 5.5 (`claude-opus-5-5`) under the standing Fable-limit fallback; runtime
effort **unknown** (not observable inside the agent). Branch `g064-mr3` from `origin/g064-final` `112cddb4` (verified:
same driver and Result content as `c8ac637f`, rebased onto staging `5e1575a8`). Allow-list: the driver and this Result
(append-only). Driver commit `2e5d6d34`: blob `e2c29f92` → `4948de14`, mode `100755`, +12/−7 lines. lib.sh is unchanged.

**Issue (Codex re-review of `112cddb4`, second pass on the same finding).** `s6b4_export` accepted status 0 with a
non-empty destination. lib.sh `redact_log`'s > 1 MiB truncation branch (`lib.sh:61-63`) can report a write error after
partial output and still return 0 from its final `rm`. A truncated copy would then pass the check, and `$WORK`, the only
complete source, would be deleted. The correction-4 Result had disclosed this limit; AC5/AC6 still require retention.

**Fix (driver only; line numbers at `2e5d6d34`).** `s6b4_export` (:521) now treats any diagnostic text captured from
`redact_log` (its stderr; it writes nothing to stdout) as an export failure, even with status 0 and a non-empty
destination: "redact_log reported an error despite status 0 (a write may have failed; the output is untrusted)". The
check comes before the presence and emptiness checks. The existing correction-4 path then applies: the partial
destination and `.redacting` are removed, the collect row is FAIL, the verdict is FAIL with exit 1, and `$WORK` and the
scratch directory are kept privately with the recovery receipt. The header (:92-96) and the function comment (:506-512)
are updated. Exit precedence and runtime-resource cleanup are unchanged.

Any stderr text from `redact_log`, a warning included, now fails the export closed. It is never treated as a pass.
Read-only proxy on the real batch-4 evidence (`s6-0696791f/b4`): `redact_log` was run on all 23 evidence copies (console,
`results.tsv`, browser JSON, `driver-logs/*`; these are the redacted forms of the exported sources) with a dummy secret.
Result: 0 files with diagnostics, all outputs byte-identical. The new check would therefore not have changed that run's
outcome.

**Checks (writer, stack-free; no stack, database, Compose, browser, build, pull, harness run or real harness-lock
probe):**

| Check | Outcome |
|---|---|
| `bash -n` | OK |
| shellcheck `-S warning -x -P SCRIPTDIR` (local pinned `koalaman/shellcheck:v0.11.0`, `--rm --pull never --network none`) | rc 0, no findings (`-S info`: the same pre-existing SC2015 ×2 and SC2317) |
| `git diff --check 112cddb4 HEAD`; U+2028/U+2029 | clean; 0 |
| Scratch test `/tmp/g064-s6b4-c4/test.sh` (correction-4 test extended; clean checkout `2e5d6d34`, driver blob `4948de14`) | **109/109 PASS**, first run |
| Control: the same test against the correction-4 driver (blob `e2c29f92`) | 100/109: exactly the 9 new partial-write checks fail (that driver exits 0 and deletes the source), so the new case discriminates |

New cases (14). The fault is a `tail` PATH shim, inert unless the case enables it, under which lib.sh's real
`tail -c 524288` reports "No space left on device" and writes nothing. The real `redact_log` therefore leaves the head
plus the marker and returns 0.
- **Function-level:**
  - F8a: a 1.5 MiB source with no fault exports through the real truncation branch (marker present, no failure).
  - F8b: the same source with the fault is an export failure with the new reason.
  - F8c: the partial destination and `.redacting` are removed.
  - F8d (control): the same fault leaves `redact_log` with status 0 and a non-empty output, the case the old check
    accepted.
- **Main path:**
  - G42: the > 1 MiB success case gives exit 0; collect, result and scan PASS; the copy is truncated and redacted; the
    scratch directory is removed.
  - G43: regression against the `08b109b8` driver on that case: same exit, same evidence file list and rows,
    byte-identical `big.log`.
  - G44-G47: the partial write gives exit 1, a collect FAIL naming `big.log` with the new reason, result FAIL and exit
    row 1. No partial copy is left in the evidence. The complete 1.5 MiB source is retained (sha256 matches the
    original), and the export failure is in `cleanup.log`.
  - G48 (4): the scratch directory is kept with mode 700 and `secrets` at 600, the retention row and recovery note are
    present, and no raw secret is in the evidence.

All 95 correction-4 cases pass unchanged. They cover the success case and its regression against the `08b109b8` driver
(identical files, rows, redacted copies and receipt), every fault negative, the scan error versus no hit, the
refusals, the 3-hunk derivation, the fail-closed inventory, cleanup continuing after a collection failure, and the exit
precedence (exit 2 and 143 kept; LEFTOVER plus a scan error gives 1).

**Limits.** Only the stack-free scratch checks exercised the new path, with a shimmed `tail` to produce lib.sh's
partial-write behavior. The passing real batch-4 run took only the success path.

**Rerun disposition.** As Codex's re-review states, this caller guard requires no shared-library change and no
database, browser, S4 or migrations rerun. Harness, Go and test sources are untouched. The batch-4 evidence at
`0696791f` stays valid: the read-only proxy above shows the guard does not change that run's exports. Budget: B5
attempt 1 continues; this is the second and last authorized correction pass on this finding (correction 5); no new
attempt.

### MR !80 Duo finding: S4 driver export fail-closed (correction 6, coordinator, 2026-10-22)

**Finding (GitLab Duo, thread `e4eb4e71`, head `927b999c`).** `G064-s4-private-run-driver.sh` `s4_finish` copied
`results.tsv` with `cp -f … 2>/dev/null`, called `redact_log` without checking its status or stderr, and then ran
`rm -rf -- "$WORK"` unconditionally. A failed export (full disk, failed `mv`, the `redact_log` truncation branch) would
delete the only complete evidence copy and the run could still end PASS with a CLEAN receipt. It is the defect class of
corrections 4 and 5 for the batch-4 driver, which the S4 driver did not carry. Disposition: accepted and fixed (this is a
real defect in an evidence driver of this Goal, AC5/AC6, not a follow-up).

**Change (driver only; harness, Go, test and production sources untouched).**
- New `s4_export SRC DST`, the same contract as the batch-4 `s6b4_export`: readable regular source, `redact_log` status 0
  and no stderr text (a diagnostic with status 0 is a failure), DST a regular file and non-empty when SRC is; a failure
  removes DST and the `.redacting` temporary and records the failure in `S4_EXPORT_FAILS`.
- `s4_finish` exports `results.tsv` and every driver log/record through it (and a failed `mkdir` of `driver-logs` is a
  failure). When any export failed it keeps `$WORK`, writes `export-failures.txt` and a `cleanup.log` line, logs the
  failure and exits 1 (a status that is already non-zero, including cleanup's 4 and 5, is kept). Otherwise unchanged:
  `$WORK` is removed and the existing exit codes apply.

**Check (stack-free).** `bash -n` passes. A scratch test extracts `s4_export` verbatim from the driver and runs it
against `redact_log` stand-ins: good copy (no failure, non-empty output); status 1; status 0 with stderr (the truncation
shape); status 0 with no destination; empty destination for a non-empty source; missing source; symlink source (each one
failure and no partial destination); empty source with empty destination (no failure). 8/8 pass. No S4, database,
browser or migrations rerun: the S4 campaign evidence (`17fdad66`) exercised only the success path, which this change
does not alter. Budget: B5 attempt 1 continues; no new attempt.

### MR !80 delta review: owned-registry export in `s4_cleanup` (correction 7, coordinator, 2026-10-22)

**Finding (Codex delta review of `f2ce4a7a`, staging `1ff263f1`).** The correction-6 helper and retention are sound, but
one export stayed unchecked: `s4_cleanup` ran `cp -f -- "$OWNED_REG" "$EVID/owned-registry.tsv" 2>>"$rcpt"` under
`set +e`, so a failed copy left `S4_EXPORT_FAILS` empty and `s4_finish` could delete `$WORK` (the only copy of the
registry) and exit 0. Fixed (the same guard as the accepted batch-4 cleanup).

**Change (driver only).** The copy is now `if ! cp … 2>>"$rcpt"; then`, which appends `owned registry export: FAILED
(the registry stays in the retained work directory)` to the receipt and `$OWNED_REG -> owned-registry.tsv: cp failed
(cleanup.log)` to `S4_EXPORT_FAILS`. The existing `s4_finish` branch then retains `$WORK`, writes
`export-failures.txt` and exits non-zero. The runtime finalizer and the removal pass run before the copy and are
unaffected. The retention line in `s4_finish` was reworded (no quoted text inside the printf format).

**Check (stack-free).** `bash -n` passes. A scratch FINALIZER-path test extracts `s4_finish`, `s4_export`, `s4_cleanup`,
`s4_present` and `s4_norm_tag` verbatim from the driver, with `docker`, `redact_log` and `owned_capture` stubs, a fake
`cleanup-owned.sh` that records that it ran, and a `cp` stub that fails only for `owned-registry.tsv`: A success
(exit 0, finalizer ran, registry exported, `$WORK` removed, no `export-failures.txt`); B failed registry copy (exit 1,
finalizer still ran, `$WORK` and registry source retained, `export-failures.txt` names the registry, one exact RETAINED
receipt line, no stderr output from the driver, receipt records the export FAILED); C the same failure with an entry
status of 2 (2 kept, `$WORK` retained); D entry status 130 without a failure (130 kept, `$WORK` removed). 16/16 pass.
Written first against the new retention wording, the test failed on a stray shell diagnostic (an apostrophe in that
printf format shifted its arguments); the wording fix above passes it. No S4, database, browser or migrations rerun;
the S4 evidence on `17fdad66` exercised only the success path. Budget: B5 attempt 1 continues; second correction pass
on this driver finding (corrections 6 and 7), no new attempt.

### MR !80 Duo threads on the retention path: superseding summary row (correction 8, coordinator, 2026-10-22)

**Finding (GitLab Duo threads `4b0aa2e6` and `755eda7e` on `f2ce4a7a`/`6e1157a9`).** With an export failure `s4_finish`
kept `$WORK` and exited 1, but `s4_main` had already appended `result S4 PASS` to `s4-summary.tsv`, so a reader of the
summary alone saw a green run; the batch-4 driver (correction 4) appends a superseding FAIL row. The receipt half (the
retention line saying the earlier "removed" line is superseded) was already fixed in correction 7. Disposition: valid for
the summary, accepted and fixed; Codex's source ACCEPT at `6e1157a9` covers the earlier corrections, and this is a
small additional delta over it (one line in the driver).

**Change (driver only).** In the `s4_finish` export-failure branch, after the exit status is fixed, one row is appended
with `s4_row result S4 - - - FAIL "$rc" - - - "after the verdict: N evidence export(s) failed and $WORK is retained
(export-failures.txt); any PASS result row above is superseded and the receipt line saying the work directory was
removed is wrong for this run"` (only when the summary file exists).

**Check (stack-free).** `bash -n` passes. The FINALIZER-path scratch test now seeds `s4-summary.tsv` with a PASS verdict
row and asserts the last row: A success keeps PASS (no extra row); B failed registry copy ends with `result S4 FAIL 1`
naming the supersession; C (entry status 2) ends with `FAIL 2`. 20/20 pass, and the same test run against the
`6e1157a9` driver fails exactly the three summary checks (the summary stayed PASS), proving the new assertions bite. The
earlier `s4_export` test passes unchanged (8/8). No S4, database, browser or migrations rerun. Budget: B5 attempt 1
continues; no new attempt.

### Merged-SHA full qualification failure and the raw MCP stream correction (correction 9, coordinator, 2026-10-22)

**Run record (preserved, not waived, not a pass).** Full local qualification of staging `6ea4cc319c1ebf75a203f64fc6514a7286581b91`
(frozen checkout `.claude/worktrees/g064-full`, clean, `validate.sh --profile full`, run label `qual-6ea4cc31`),
started 2026-10-22T12:30:13Z, wrapper exit **1** at 14:04:23Z (harness 82 min, final status 1, cleanup status 0, owned
resources remaining 0, build ledger 76/76 MET). `results.tsv`: 130 rows, 120 PASS, **3 FAIL**, the rest INFO. Evidence
root `<home>/.local/state/platform-evidence/g064-a1/` (unmodified): `qual-6ea4cc31.console.txt` sha256
`138deb2f…8719`; `qual-6ea4cc31/ci-evidence/candidate/` `results.tsv` `0d75d73c…905b`, `harness.log` `6ce882a0…b8e5`,
`junit.xml` `180ccc8b…c92d`, `go-tests-junit.xml` `e08b80a7…6e892`, `diagnostics/ci-controls-summary.log`
`0a5f7ade…5421`, `ci-controls-timings.tsv` `3c0667ca…8273`. The three failures:

| Check | Row | Disposition |
|---|---|---|
| `B-SECRET-CANARY` | secret found in the harness work directory file `mcp-e2e.log.events` | confirmed G064 defect; corrected below (Codex-authorized, handoff `1e09b3f5`, record `G064-fullrun-failure-disposition-codex-review-20261022.md`) |
| `B-CI-CONTROLS` | `ci-controls: 526 passed, 1 failed (of 527 selected)`; the failed control is **V18m** | separate failure, diagnosed below; the full run's row is unchanged |
| `B-REQUIRED-CHECKS` | derived: `B-CI-CONTROLS(1,FAIL) B-SECRET-CANARY(1,FAIL)` | remains until both are resolved on a qualifying run |

**V18m (separate, not attributed to the MCP defect).** V18m sends SIGHUP to a nested validator and expects exit 129
within a 240 s await. In the full run it took 289 s (nested exit 124, want 129; the nested tree was then stopped by the
control itself), while its siblings V18l and V18n of the same mechanism took 48 s and 73 s; it ran overlapped with the
heavy harness suites. Focused retry, one control on the same frozen checkout (`--only V18m`, started 14:07:02Z, evidence
`qual-6ea4cc31-v18m/` and `qual-6ea4cc31-v18m.console.txt` sha256 `c96b1257…f143`, summary `a96c3176…082a`): **PASS in
52 s**, nested exit 129, 14 processes reaped, 0 resources after, scratch removed. This is consistent with load
sensitivity of the 240 s await but does not prove it, and it is not a pass of the full run; the disposition (and whether
the full profile needs a repeat) is Codex's.

**Defect (B-SECRET-CANARY).** `go_it` saves the raw `go test -json` stream as `LOG.events`. `mcp_e2e` registers the
minted bearer values from the converted log, deletes the `MCP-OBSERVED` lines from the converted text log and replaces
them with `[withheld]` in `go-tests/mcp-e2e.log`, but leaves `LOG.events` (`$WORK/mcp-e2e.log.events`) with the raw
values; `B-SECRET-CANARY` scans the work directory and found them. No other `.events` file holds a registered secret.

**Change (`infra/validation/run.sh`, three lines, one in `mcp_e2e`).** After the registration and both scrubs and before
`go_it_publish`, `rm -f -- "$log.events"` with a two-line comment. The converted per-test log stays and stays redacted.
Nothing else changes: product authentication, test assertions, `lib.sh`, the driver and every other check are untouched.
(Observation, not changed: if the converter image is unavailable `go_it_convert` copies the raw stream into the
converted log; that path fails closed already and is outside this authorization.)

**Check (stack-free, scratch, out of repository: `canary-fix/g064-c9-test.sh` sha256 `7b5713ec…e6da`,
`g064-c9-extract.py` `1036c500…0101`, output `test-output.txt` `dc9c8011…4b2`; run.sh sha256 old `adbcfd30…2d`,
new `b07e0052…c0`; blobs old `4fc59cd7`, new `00ce16ee`).** The test extracts the real `mcp_e2e` and the real
`B-SECRET-CANARY` scan block from each `run.sh`, stubs everything else (no Docker, database or network), plants 12
generated bearer values per scenario and reports counts only, never a value: **14/14 pass.**
Negative control on the old staging `run.sh` (79ef7416): all 12 values registered, the raw `.events` stream retains the
12 planted lines, and the real scan records `B-SECRET-CANARY FAIL` naming `mcp-e2e.log.events` (reproducing the
reported failure). Fixed `run.sh`, passing invocation: `B-MCP-E2E` reaches its PASS row, all 12 values registered, no
`.events` file under the work directory, the converted log and the per-test log hold no planted value (12 withheld
lines), the exported evidence holds none, `B-SECRET-CANARY PASS`, and registration precedes removal inside the
function. Fixed `run.sh`, failing invocation (`go_it` status 1): the raw stream is still removed, all 12 values are
registered, `B-MCP-E2E FAIL` is recorded and `B-SECRET-CANARY PASS`. `bash -n` passes; the pinned shellcheck
(`-S warning -s bash`) reports 2 findings on both the old and the new `run.sh` (none new).

No S4, database, browser, `migrations`, `ops` or full-profile run was repeated for this correction. Budget: B5 attempt 1
continues; no new attempt. This is submitted for Codex source review only; it is not a merge or final acceptance, and
G064's final acceptance stays pending until the required validation disposition for the failed full run is complete.

### Converter-unavailable path of the raw MCP stream (correction 10, Codex source review `e27d84cd`, 2026-10-22)

**Finding (Codex, handoff `3bdfac0c`, record `G064-canary-correction-codex-review-20261022.md`).** Correction 9 was sound
for the normal path, but when the validation-tools image or the conversion is unavailable `go_it_convert` copies the
raw `LOG.events` JSON into the converted logs as its fail-closed diagnostic. `mcp_e2e` registered and removed only plain
lines beginning `MCP-OBSERVED`, so the raw JSON stayed in `$WORK/mcp-e2e.log` and `$WORK/go-tests/mcp-e2e.log` with its
values unregistered, the later `tail -60 "$log"` could print it, and the failure exporter could retain or export it;
the secret canary would not recognize unregistered values. Accepted: this was visible to me in correction 9 as an
observation and should have been fixed there instead of being scoped out.

**Change (`infra/validation/run.sh`, `mcp_e2e`, 10 added lines, before the plain-line handling).** If the converted log
begins any line with `{"` (a raw `go test -json` event) or the invocation's tsv row names `converter-unavailable`
(either alone is enough): register every `MCP-OBSERVED <value>` found in the converted log and in `LOG.events` (the value
ends at a space, a quote or a backslash, the JSON string's own delimiters), replace both converted logs by one bounded
sanitized record that states that the raw stream is withheld and why, replace the converter's stderr file (it may echo a
raw line) by a withheld note, and set the invocation failed. The existing steps then run unchanged: nothing plain is
left to register, the diagnostic `tail` prints only the sanitized record, `LOG.events` is removed, and the export
publishes the sanitized record. The normal path is byte-for-byte as in correction 9; product authentication, test
assertions, `lib.sh` and every other check are untouched.

**Check (stack-free, scratch, out of repository; counts only, never a value).** The regression (`canary-fix/g064-c9-test.sh`
sha256 `dabdd064…65e3`, `g064-c9-extract.py` `1036c500…0101`, output `test-output.txt` `0a136e8b…dad8`; this replaces the
correction 9 script `7b5713ec…e6da`, which is kept in the earlier section's record) extracts the real `mcp_e2e` and the
real `B-SECRET-CANARY` scan block from each run.sh, with 12 generated planted values per scenario, and now runs **28
checks, 28 pass**: the three old-code negative-control checks and the eight normal-path checks of correction 9 (staging
`79ef7416` `run.sh` `adbcfd30…abb2d` leaks and fails the real scan; candidate passes), the failing normal invocation
(three checks), a **negative control on the first fix** (`592ab2bb` `run.sh` `b07e0052…a5c0`): with the converter
unavailable the 12 planted values stay in the work file `mcp-e2e.log` and the failure diagnostic prints 12 of them to
stdout (this reproduces Codex's finding), and for the candidate (`run.sh` sha256 `262819f5…e1d5`, blob `4c827c7b`) with
the converter unavailable (raw copy, tsv marker, stderr echoing a raw line): the invocation stays failed (B-MCP-E2E FAIL,
no PASS row, B-SECRET-CANARY PASS), all 12 values are registered from the raw stream, no `.events` file survives, no
planted value occurs in stdout/stderr, in the remaining work files (converter stderr included), in the exported
diagnostics, or anywhere in the retained scratch outside the registry, and both converted logs hold the sanitized
record; plus the content-based detection alone (raw copy without the tsv marker): 12 registered, no value in
stdout/stderr, work, exports or scratch, no `.events` file. `bash -n` passes; the pinned shellcheck (`-S warning -s bash`)
reports 2 findings on the staging, first-fix and candidate `run.sh` alike (none new).

No S4, database, browser, `migrations`, `ops` or full-profile run was repeated. The failed full run (`6ea4cc31`) stays
failed and unwaived; the V18m focused pass stays separate and does not erase its failed row. Budget: B5 attempt 1
continues; no new attempt. Source review only, not an acceptance or merge.

## One-off exception: one pre-merge full run replaces the post-merge full run (owner direction, 2026-10-23)

**Authorization (recorded before the candidate is frozen).** The owner told the coordinator directly in this session:
"Ok you can do a lre merge full run right before merge" (read as "pre-merge") and, after the coordinator said it would
relay this to Codex, "That's great, tell codex". An earlier owner turn, "The final rerun should be after merge", is
superseded by these. Codex relays 069e707c (conditions) and 0fd94fc0 (the owner's instruction given to Codex directly)
confirm the exception.

**Scope (one-off, G064 MR !82 only).** ONE complete full local run on the final rebased candidate, taken after every other
pre-merge gate has cleared (hosted fast pipeline, fresh Codex pre-integration review, fresh Duo review and disposition,
target/source reconciliation, open threads 0, every hold). If it succeeds it REPLACES the post-merge full run; the
coordinator does not run both. This is not a general policy change and creates no precedent for other MRs.

**Conditions.**
1. The merge uses `--sha` on the frozen candidate and the staging tip stays an ancestor of the candidate. After the merge,
   the merge commit's tree must equal the tested candidate tree exactly, with no intervening source change.
2. If the merged tree differs, or any source changes after the run, STOP: reconcile the changed coverage before any
   qualification claim; the run is then candidate-only and the bound post-merge qualification still runs.
3. A failed run stays recorded as failed (no waiver, no budget reset). All cleanup receipts are preserved.
4. The evidence record (candidate SHA and tree, the complete profile evidence, the final merged SHA and tree, the cleanup
   receipt) is added by a separate docs-only Result MR after the merge, as for G063. Until then this section records the
   authorization only, not a result.

The earlier failed merged-SHA full qualifications (`qual-6ea4cc31` for !80, `qual-79085025` for !81) remain failed.

## Correction record: Codex source review of `cb7e3f3d` (MR !82, 2026-10-23)

**Findings and rulings.** Codex's source review (handoff 9feed637) returned a correction loop with two resource-safety
findings in `infra/ci/validate-controls.sh`; handoff 426a3fca narrowed it (see below); handoff 389d8de7 added one valid
Duo finding (thread a22da37d); handoff 4c92678e ruled on the image ceiling. Hosted fast pipeline 2979909464 was green on
`cb7e3f3d` (including B-SHELL and B-DEPSCAN). Duo's three earlier threads (OPERATIONS.md 528, `f1_child_env` guards,
nested-scope latch lift) were fixed in `26a19a3f`..`48d267f0`, before the rebase, and resolved.

**Evidence attribution (Codex 426a3fca, which withdrew the demand for a new complete F1 run).** The F1 real-Docker driver
PASS on `b54627bb` (30 PASS rows, exit 0, `f1-proof/full-051143`) stays attributable: the later F1 driver delta is two
defensive `${..:-}` defaults, `lib.sh` changes the nested-scope baseline/latch logic, shellcheck fixes and the Next pins,
`run.sh` changes naming/quoting. A second full F1 run on `cb7e3f3d` also PASSED (exit 0, 07:21-07:29Z, `full-072115`; the
inventory was identical before and after: g003=0, containers 15, volumes 12, networks 6, images 57; the extra volume
`project-archetype-render-cache`, created 05:30Z inside the Next maintenance fast-profile window, is unlabelled and
attributed to it only probably, not removed). It predates the corrections below and is not claimed as proof of them. The
earlier FAILED real-Docker runs (r12, F1 runs 1-2) and the failed merged-SHA full qualifications (`qual-6ea4cc31`,
`qual-79085025`) stay recorded as failed.

**Corrections (writer 16, then the coordinator; owner time-boxes 2026-10-23: 20 minutes, extended by 20; the owner's
"blind" wording was corrected to "bound", so each fix keeps one focused proof).**
- `7790739e` P1: a gated container creation (the control's `D()` and the case shim alike) hands the Docker CLI a cidfile
  under the run registry's `cid/` directory, the existing `cleanup-owned.sh` cid provenance resolved by exact ID, so a
  creator killed during an attached call leaves a recoverable object; a caller's `--cidfile` receives the ID after the
  call; a registry that cannot take a cidfile refuses the creation. Basis: Plan addendum A C4 (existing mechanism, cid
  provenance), review finding 2 (no new recovery framework).
- `67995784` P2: the shim's `build`, `buildx build` and `image build` pass `shim_gate`/`shim_reserve` (kind image): run
  hold, control latch, ceiling and reservation; the built ID (the caller's `--iidfile`, else an injected one, else a tag
  inspect) is recorded in the inventory and mirrored to the registry. `commit` through the shim is not reached by any
  control (`D()` already gates it).
- `15ad40aa` Duo `_res_live`: registry rows `standalone <id> ci-<control>` are background objects of the run: not counted
  by any foreground scope, ceiling or zero check, still owned, cleaned, recovered and counted in the run-wide sample.
- `f0c2c264` P2 ceiling (Codex 4c92678e): the per-control image ceiling is a fixed 24 for the build-ledger controls
  VC1-VC9 only, 2 for every other control; container, network, volume and outer limits and the run hold are unchanged;
  `D()` and the shim use the same `gate_ceil`; every live image ID counts, dangling ones included. Reason: with the shim
  gated, the universal 2 would latch the existing VC3 (4 builds), VC5 (2 fixture outputs and 2 builds) and VC7 (15
  builds) scenarios on real Docker. 24 is a conservative operational cap for that finite existing fixture family, not a
  measured peak and not a general increase.

**Proofs (stack-free, guard docker and a dead `DOCKER_HOST`, the real CLI never reached; evidence under
`.../g064-a1/reuse/s11-fix5/` and `s13-vcceil/`).** `f0-p1p2` 10/10 on the corrected tree; on `cb7e3f3d` P1 x3 and P2 x2
FAIL (no durable binding: `cleanup-owned.sh` finalizes with exit 0 and removes nothing while the container exists; builds
pass under the run hold). `f0-reslive` 5/5; on `cb7e3f3d` R1/R1b FAIL (`hold.d/run` on a live background control), the real
foreground leftover cases R3/R4 hold on both. `f0-vcceil` 34/34: VC1-VC9 image 24, every other control 2, other kinds
unchanged, a lowered injected cap crosses and is refused (latch and ceiling row), cap 24 admits, the run hold refuses a VC
image creation, and the created image lands as an inventory row and a registry row; on `15ad40aa` 10 of its checks FAIL.
Suites touching the changed files, run by writer 16 on `15ad40aa`: `f0-dgate` 22/22 (two expectations updated to the
registry-cidfile design), `f0-controls-state` 27/27, `f0-latch` 59/59, `f0-selfreg` 23/23, `cldrv` 17/17 and guard 9/9.
After `f0c2c264` (a three-line `gate_ceil` change) only `f0-dgate` 22/22, `f0-controls-state` 27/27, `f0-p1p2` 10/10,
`f0-reslive` 5/5 and `f0-vcceil` 34/34 were rerun. `bash -n` passes and `--list` prints 528. ShellCheck v0.11.0 (both `lint_shellcheck` invocations, pinned local image) exits 0 with no findings on
`f0c2c264`.

**Not run (owner and Codex bounds) and residual risks.** No mutation campaign, whole-suite list, complete F1 rerun or
real-Docker VC control run on the corrected tree; the one authorized final full run exercises the real VC controls (the
24 cap is untested on a real daemon until then). A caller's `--cidfile` is written after the CLI returns, so a concurrent
reader sees it empty until then (a conservative over-count only). Pre-launch binding for network/volume creators and the
CLI's create-returned-but-cidfile-unwritten window are not closed (no reached path needs the former; the latter is the
existing `cleanup-owned.sh` empty-cidfile rule). Attempt count and budget are unchanged.

### Correction record corr4 — the qual-pre-c070aff8 false foreground latch (writer 18, 2026-10-23)

**Failure.** The authorized pre-merge full local run on `c070aff8` (evidence `.../g064-a1/qual-pre-c070aff8/`) passed
lint, lint-base, lint-web, lint-core, provenance and all six contracts checks (B-CONTRACT-NEGATIVE PASS 04:30:53 local),
then at 04:30:54 `FAIL RESOURCES - run: scope suite-contracts ended with foreground containers 1 (verified zero required):
9dda2516…`. `hold.d/run` refused every later allocation: the root start block and the migrations down/reapply (exit 125),
then `bg_stop_all` recorded res-witness, controls and ci-controls "stopped before join" (04:31:32). root (start block),
migrations, ops, rename, off, the controls suite (B-DEPS-CONTROLS, B-WRAPPER-CONTROLS) and ci-controls (V1-V3 PASS; V4,
V7, V8 `published=none` and V9 cancelled after the latch; 3 passed, 3 failed of 528) are therefore unexecuted or
incomplete on `c070aff8`. The host returned to its 15-container baseline; cleanup removed exactly the witness.

**Root cause (by elimination from the harness code; the registry was removed with the run, so the row of `9dda2516` is
not recoverable).** `_res_live` classifies a container as foreground by the run label `g001.run=$RUN_ID`, a registered
Compose project, or a registry row other than `standalone <id> ci-<control>`. Every container of the contracts suite is
an attached `docker run --rm` through `docker_run_owned`, which the Docker 27 CLI returns from only after the daemon's
removal (`WaitConditionRemoved`), so none can be listed one second after B-CONTRACT-NEGATIVE's last step. The only
concurrent creator sharing the run label and the registry is the `controls` background child (`bg_start controls`
04:29:00, forked from the same run.sh with the same `RUN_ID` and `OWNED_REG`): `static_deps_controls` runs `yq_stdin`
(`docker_run_owned yq --rm --label g001.run=$RUN_ID`) for its deps cases, each registered `standalone <id> yq`, so a `yq`
container alive at the instant of the foreground scope-end query counts as a foreground container of that scope. The
controls suite took 142 s and 174 s in the two preserved earlier full runs (`qual-79085025`, `qual-6ea4cc31`); here the
contracts scope ended 114 s after its start under load 8-9, inside the deps-controls phase. The race is the window of one
short-lived attached container at one query instant, which is why earlier runs passed. The container is this run's own
creation (run label, registry row), not a foreign host container; the auto-named `gracious_elion` the post-run inventory
saw at 04:31:39 and later not at all is consistent with a `docker run --rm` container of the killed background child
auto-removed by the daemon after its client died, but its identity is unknown. ci-controls V4/V7/V8/V9 are consequences:
their case creations are refused under `hold.d/run` (`gate_reserve` reads it), and V9 was cancelled by `bg_stop_all`.

**Fix (`infra/validation/lib.sh`, 8 lines).** `_dro` registers a creation made inside a supervised background child
(`BG_CHILD`, set by `bg_start`) as `standalone <id> bg-<child>-<purpose>`; `_res_live` classifies `bg-*` rows like
`ci-*` rows (background: owned, cleaned, in the run-wide sample) and excludes a background-row container from the
foreground set even when it carries the run label (the row decides). A run-labelled container without a registry row, a
main-shell creation, a registered foreground leftover and an unknown resource all still latch; no ceiling or limit
changed. `cleanup-owned.sh` accepts the purpose (`^[A-Za-z0-9._-]{1,64}$`); `res_report` attributes the row as before.
Residual: a background child's own detached creations no longer count toward its own reservation ceiling through `_res_live`
(the run-wide `observed_max` row in resources.tsv still samples them; the controls child only runs attached `--rm` yq
containers, one at a time), the same accounting the accepted `ci-*` rule already has. *Incomplete (Codex review, corr5
below): the row is appended only after the attached call returns, so the live `yq` interval itself had no row and was
still foreground; the f0-bgchild cases above proved classification after registration, not during the call.*

**Proofs (stack-free: fake docker first in PATH, dead `DOCKER_HOST`, real CLI never reached; evidence under
`.../g064-a1/reuse/s14-fix6/`).** New `f0-bgchild.sh` (standalone proof, not a catalogued control; `--list` stays 528):
on the `c070aff8` lib.sh B1/B1a/B1b/B2/B5 FAIL (`hold.d/run` on a live background-child container at the foreground
scope end; the row is `standalone <id> yq`) and the guard cases B3/B4 PASS; on the corrected tree 7/7. Existing proofs on
the corrected tree: f0-reslive 5/5, f0-p1p2 10/10, f0-vcceil 34/34, f0-dgate 22/22, f0-controls-state 27/27, f0-selfreg
23/23, f0-runsh-reg 5/5, f0-latch 53/53, f0-statebase-exit 2/2; f0-vr9 12/14 with V3/V4 failing identically on the
untouched `c070aff8` infra tree (its two cidfile-path expectations predate the accepted corr3 P1 registry-cidfile
relocation; not a regression of this change). `bash -n` passes on lib.sh and the proof. The project ShellCheck gate and
the real-daemon full qualification of the corrected candidate remain to run.

### Correction record corr5 — the in-flight attached background creation (writer 19, 2026-10-23)

**Codex finding (CORRECTION LOOP on corr4).** At `_dro` an ordinary attached `docker run` blocks, and the
`standalone <id> bg-<child>-<purpose>` row is appended only after it returns (`registry_row` after the cidfile read).
`yq_stdin` is attached `--rm`, and for its known zero-volume image the split condition is false, so throughout the live
`yq` interval the container carried `g001.run=$RUN_ID` and had no row: `_res_live` still admitted it as foreground and the
scope-end race stayed reachable. The f0-bgchild proof of corr4 used `-d`, waited for the call to return and only then
ended the scope: classification after registration, not during the attached call. Full local run 2 on `88c51416` failed
at the same point as run 1 (contracts scope end latched on "foreground containers 1", container `189f7423…`, 05:08:48
local; the coordinator's run record), which confirms corr4 was incomplete.

**Change (`infra/validation/lib.sh`, 15 insertions / 7 deletions; no ceiling, limit, spec or catalogue change).**
Inside a supervised background child `_dro` now also passes a creation-time label `--label g064.bg=$BG_CHILD` to the
`docker run` (and, prepended to `DRO_ARGS`, to the split `docker create`); `_res_live`'s single `docker ps` query reads
that label as a fifth column, and a container carrying it together with this run's `g001.run` label is background (set B:
never foreground, in the run-wide sample) from its first listing, unless a registry row other than `bg-*`/`ci-*` names it
(a dead owner's `reconciled` leftover): then the row decides and it is foreground again. The post-return `bg-*` row is
kept unchanged (ownership for cleanup and recovery; `g001.run` stays on the container so the sweeps are unchanged). The
label was chosen over reading the `-bg-` cidfiles in `$RESD/cid/` because the CLI writes the cidfile after the daemon
has created the container, which leaves a create→cidfile window in which the container is listable but unattributed,
while a label is part of the create request itself and needs no second file read in the accounting query. Guards kept:
a run-labelled container with neither the label nor a row, a main-shell creation (no `BG_CHILD`, so no label) and a
registered foreground leftover still latch; a `g064.bg` label without this run's `g001.run` label counts for nothing.

**Proofs (stack-free: fake docker first in PATH, dead `DOCKER_HOST`, real CLI never reached; evidence under
`.../g064-a1/reuse/s15-fix7/`: `run-proofs.sh`, `run-proofs.out`, `sha256.txt`).** `f0-bgchild.sh` copied from
s14-fix6 and extended; the copied fake docker gained the fifth `ps` column and `FAKE_HOLD_FILE` (an attached `run` writes
`.held` once the container is listed and its cidfile written, then blocks until released). B6 runs the held attached
`run --rm` in a background shell with `BG_CHILD=controls` and `RES_SCOPES=""` (run.sh forks the controls child before its
first `res_scope_begin`), and while it is held the main shell asserts no registry row yet (B6-pre), `_res_live` gives
`RES_FGN=0 RES_ALLN=1` (B6a), `res_scope_end suite-contracts` writes no `hold.d/run` and no "ended with foreground
containers" row (B6), then releases the call and checks rc 0 and the `bg-controls-yq` row (B6b). Guards: B7 the same held
call from the main shell still latches; B8 a run-labelled, `g064.bg`-labelled container whose only row is `reconciled`
still latches; B3/B4 unchanged. B9 shows the split `create` path carries the label (`FAKE_IMG_VOLS=1`). On the `88c51416`
lib.sh (`pre/`, sha256 `0c3e0458…`): B6a, B6 and B9 FAIL, 11 PASS (the gap exactly as found); on the corrected tree
(sha256 `aa3dd045…`): 14/14. Existing proofs on the corrected tree: f0-reslive 5/5, f0-p1p2 10/10, f0-vcceil 34/34,
f0-dgate 22/22, f0-controls-state 27/27, f0-selfreg 23/23, f0-runsh-reg 5/5, f0-latch 53/53 (its static A12 guard pinned
the exact `docker create --cidfile "$cf" "${DRO_ARGS[@]}"` text, which is why the label is prepended to `DRO_ARGS`),
f0-statebase-exit 2/2; f0-vr9 V3/V4 are stale private-proof expectations (corr4 record) and were not rerun.
`bash -n` passes on lib.sh, the proof and the fake; ShellCheck is not installed in this worker (the coordinator runs the
gate); `validate-controls.sh --list` stays 528 (the proof is standalone, not catalogued).

**Status.** The root-cause identity is inferred: neither failing run retained the offending container's registry row
(the registry is removed with the run), and both runs latched at the contracts scope end while the controls child's
deps cases ran; the in-flight attached `yq` container is the only creator that fits, and B6 reproduces that interval on
the corrected and the base lib.sh. Final runtime qualification (the real-daemon full local run on the corrected source)
and the source review remain pending; attempt count and budget are unchanged.

### Correction record corr6 — a re-created Compose volume read as a destroyed anonymous one (writer 19, 2026-10-23)

**Finding.** Full local run 3 on `4f73e412` failed in the `migrations` suite: `hold.d/run` "accounting: anonymous volume
g001-mig-437… of b6d44ef3c0de not verified removed". The volume is the Compose-named `g001-mig-<RUN_ID>_postgres-data`:
the suite's `compose down -v` destroys it and a later `compose up` re-creates it under the same name. Cause, INFERRED by
code reading (the witness work dir was removed with the run; the event sequence is unrecoverable): `_res_witness_parse`
set `vgone[V]` on `volume/destroy` and the later `volume/create` did not clear it, so at the window the obsolete marker
skipped the label inspection (Compose project ⇒ `named`) and `_res_vol_gone` found the re-created volume alive ⇒ latch.
The reproducer confirms this: fed through the real parser, the sequence container create → volume create → mount →
volume destroy → volume create (same name, Compose label, existing at the window) → container destroy → window latches
exactly "anonymous volume g001-mig-437 of … not verified removed" on `4f73e412` and not on the fix.

**Change (`infra/validation/lib.sh`, one line).** `volume/create` now also does `unset "vgone[$id]"`: a witnessed
re-create voids the earlier destroy, so the volume's own labels decide again at the window. No generation tracker,
ceiling, limit, run.sh, migration or application change. One line in OPERATIONS.md (D1 "Volume removal").

**Proof (stack-free: fake docker first in PATH, dead `DOCKER_HOST`, real CLI never reached; evidence under
`.../g064-a1/reuse/s16-fix8/`: `f0-volrecreate.sh`, `run-proof.sh`, `f0-volrecreate-{pre,post}.out`, `run-existing.sh`,
`run-existing.out`, `sha256.txt`).** `f0-volrecreate.sh` drives `_res_witness_parse`/`_res_witness_window` from the
sourced lib.sh with a proof-local fake docker (volume state, labels, inspect errors, `rm` failure knob). V1 the case
above ⇒ `named … compose project g001-mig-4374d1 (project teardown)`, no latch; V1c re-created then destroyed again ⇒
`gone removed_by=compose`; V3c re-created without a Compose label ⇒ harness `volume rm`, `gone`. Guards kept: V2 a
genuine anonymous volume destroyed and gone ⇒ `gone`, no `rm`; V3 one left behind ⇒ harness `volume rm`, verified; V3b
the same with an unverifiable removal still latches "not verified removed"; V4 a non-"no such" inspect error still
latches "not inspectable"; V5 a pre-existing volume stays `named`. On `4f73e412`: V1, V1b, V3c FAIL (7 PASS; V1's
reason is the production text); on the fix 10/10. Existing proofs on the fix: f0-reslive 5/5, f0-p1p2 10/10, f0-vcceil
34/34, f0-dgate 22/22, f0-controls-state 27/27, f0-selfreg 23/23, f0-runsh-reg 5/5, f0-latch 53/53, f0-statebase-exit
2/2, f0-bgchild 14/14; f0-vr9 12/14 (V3/V4 the known stale expectations, identical on base). `bash -n` ok on lib.sh and
run.sh; ShellCheck not installed in this worker (coordinator gate); `validate-controls.sh --list` stays 528. The
real-daemon full local run on the corrected source and the source review remain pending; attempt count unchanged.

### Correction record corr7 — the S-OFF container ceiling (writer 20, 2026-10-23)

**Finding.** Full local run 4 on `9cffb495` failed B-BROWSER-OFF: scope S-OFF latched `ceiling container 8/7 (run yq)`
at 11:27:09Z (resources.tsv: S-OFF container creations 184, observed_max 7, ceiling 7, `FAIL latched`; held.log: seven
`refused run yq` in that second; evidence `.../g064-a1/qual-pre-9cffb495/ci-evidence/candidate/` and the snapshot
`.../g064-a1/run4-off-diag/`). Cause, verified in code: `browser_run` (lib.sh) runs `compose_browser … run --name <runc>
--build --no-deps web-browser`; the one-off is retained until `browser_collect_report` removes it, but `capture_after`
sees `run` + `--build`, so `record_project_images` runs one `yq_stdin` per compose file — a `docker_run_owned`
admission (`res_reserve`) — while the one-off still exists. `_res_live` counts `docker ps -a` (state-agnostic), so the
off stack is 6 containers, not the A5 note's 5: redis, postgres, minio, the exited one-shot minio-init, core, web.
6 + 1 retained one-off + 1 yq = 8 > 7 ⇒ latch; the later yq refusals are consequences. The identity of the 7 counted
containers is INFERRED from code and the registry order (no `docker ps` was captured at 11:27:09Z). S-REN never crossed
because `variants_rename` builds the browser image first (yq with the 7-container stack = 8/8) and runs it without
`--build` (7 + 1 = 8/8, no yq); S-ROOT has the same overlap under 11.

**Change (`infra/validation/lib.sh`, one value).** `[S-OFF]=8/2/3` (was 7/2/3), the derivation in the comment: the
measured/derived S-OFF peak, not a blanket raise. Not changed: every other ceiling, deadline and guard, the
`browser_run`/`--build` structure, the `_res_admit` refusal text (appending counted IDs was skipped: it would alter the
exact strings f0-latch, f0-dgate and this proof assert), run.sh, migrations, application code. OPERATIONS.md's ceiling
row updated; the approved Plan's A5 row (7) stays as the historical input, superseded by this record.

**Proof (stack-free: fake docker first in PATH, dead `DOCKER_HOST`, real CLI never reached; evidence under
`.../g064-a1/reuse/s17-fix9/`: `f0-offceil.sh`, `run-proof.sh`, `f0-offceil-{pre,post}.out`, `run-existing.sh`,
`rerun-latch.sh`, `run-existing.out`, `suites/`, `sha256.txt`).** `f0-offceil.sh` sources lib.sh with a proof-local
fake docker whose `ps -a` lists the 6 off-stack containers of the registered project plus the run-labelled retained
one-off. O1 the capture yq through `yq_stdin` → `docker_run_owned` → `res_reserve`; O2 the production path
`capture_after DEST … run … --build … web-browser` (owned_capture, then one yq per compose file); O3 one more container
⇒ refused `ceiling container 9/8 (run yq)`; O4 the table (only S-OFF changed; `_res_ceiling` 8/2/3); O5 S-REN: 7 stack
+ build yq admitted 8/8, 7 stack + one-off + yq refused 9/8; O6 S-ROOT: 7 + 1 + 1 admitted under 11. On the base
lib.sh (`b7cdcc9c…` = 9cffb495): O1, O1a, O2, O3, O4, O4b FAIL (5 PASS), O1/O2 latching the production text `ceiling
container 8/7 (run yq)`; on the fix 11/11. Existing proofs on the fix: f0-reslive 5/5, f0-p1p2 10/10, f0-vcceil 34/34,
f0-dgate 22/22, f0-controls-state 27/27, f0-selfreg 23/23, f0-runsh-reg 5/5, f0-bgchild 14/14, f0-volrecreate 10/10,
f0-statebase-exit 2/2; f0-vr9 12/14 (V3/V4 the known stale expectations, identical on base); f0-latch 53/53 on an
s17 copy whose only delta is its two hard-coded S-OFF expectations (C6 crossing 9/8 with 9 fake containers, W3 table
8/2/3; the shared `reuse/tools/f0-latch.sh` untouched, 51/53 against the fix). `bash -n` ok on lib.sh and run.sh;
ShellCheck not installed in this worker (the run's lint covers it); `validate-controls.sh --list` stays 528. The
real-daemon full local run on the corrected source and the source review remain pending; attempt count unchanged.

### Correction record corr7-B — `_res_live` read the caller's stdin without a registry (writer 20, 2026-10-23)

**Finding.** ci-controls VC8 (w8d2) and VC9 (w9a/w9d/w9e) failed on `9cffb495` with "compose config unavailable for
build" (every Compose plan row unknown). The coordinator reproduced both standalone on the real daemon and pinned the
failing step by instrumented replay with a logging docker wrapper (`.../g064-a1/run4-vc89/`: `stdout.txt`,
`replay3.out`, `wrap/real.log` lines 2-8: `config stdout=735`, then `run -i yq … stdin=0`; `stage1-4` pass with a file
stdin; `awk-empty.sh`: gawk reads stdin on an empty operand). Cause, verified in code: the VC driver
(`validate-controls.sh` `vcdrv.sh`, `env -i`) sources lib.sh with no registry (`OWNED_REG` empty); `_lg_config` pipes
`docker compose … config` into `yq_stdin` → `docker_run_owned` → `res_reserve` → `_res_admit` → `_res_count` →
`_res_sample` → `_res_live`, whose group command (inside `$(…)`, stdin inherited) ran `awk … "$OWNED_REG"` with an empty
file operand; gawk 5.2.1 skips it and reads stdin, so the config was consumed and the real `docker run -i yq` got 0
bytes. The operand came with the bg/ci rows (corr3/corr4); run.sh always sets `OWNED_REG`, so only registry-less lib.sh
users are hit. Audit of the resource block: every other `"$OWNED_REG"` operand (lines 407, 593, 709, 1066, 1115) sits
behind an `-n`/`-z` guard; the `sed` operands cannot read stdin (sed errors on an empty operand); no other
awk/grep/cat/head/wc operand in `_res_live`, `_res_count`, `_res_sample`, `_res_obs`, the `res_*` helpers, `_dro` or
`registry_row` can be empty.

**Change (`infra/validation/lib.sh`, `_res_live`).** The awk operand is `"${OWNED_REG:-/dev/null}"` and the group
command takes `</dev/null` (nothing in it reads stdin). Not changed: the classification (ci-/bg- rows background, run
label foreground, `reconciled` override), ceilings, latch texts, `_lg_config`, the VC driver, run.sh; OPERATIONS.md
does not describe `_res_live`'s inputs and is untouched.

**Proof (stack-free: guard docker first in PATH, dead `DOCKER_HOST`, real CLI never reached; evidence under
`.../g064-a1/reuse/s18-fix10/`: `f0-stdin.sh`, `run-proof.sh`, `f0-stdin-{pre,post}.out`, `run-existing.sh`,
`run-existing.out`, `suites/`, `sha256.txt`).** The guard docker records the bytes its `run -i` receives. S1
`_res_live container` with a 735-byte pipe on stdin and no registry: bytes left on the pipe 0 on `ecd0cecf`, 735 on the
fix. S2 `payload | yq_stdin` (the `_lg_config` shape): the shim received 0 bytes and yq printed nothing on `ecd0cecf`,
735 bytes and output on the fix (the admission is reserved and released on both). Controls holding on both: S3 with a
registry the pipe passes through and the corr3/4/5 classification is unchanged (ci- row and bg-labelled container
background, run-labelled foreground, a `reconciled` row overrides the bg label); S4 no registry, non-stdin caller:
result unchanged; S5 a failed query still fails `accounting:`. Pre 6/8 (S1, S2 FAIL), post 8/8. Existing proofs on the
fix: f0-reslive 5/5, f0-p1p2 10/10, f0-vcceil 34/34, f0-dgate 22/22, f0-controls-state 27/27, f0-selfreg 23/23,
f0-runsh-reg 5/5, f0-latch 53/53 (s17 copy), f0-bgchild 14/14, f0-volrecreate 10/10, f0-offceil 11/11,
f0-statebase-exit 2/2; f0-vr9 12/14 (known stale V3/V4). `bash -n` ok; ShellCheck not installed here;
`validate-controls.sh --list` stays 528.

**Dependents to rerun on the real daemon (`validate-controls.sh --only …`).** Directly affected: VC8, VC9. The same
registry-less VC driver, never reaching `_res_live` (no Compose plan step): VC1–VC7, a cheap regression set.
Registry-set callers of the edited group command (`docker_run_owned`/`res_*` through the stub harness or the real
run.sh): V27, V27b, VR1, VR2, VR9, VB1, VB3b. VB2–VB6 drivers source lib.sh without a registry but use only the bg_*
helpers and `docker_inventory`, so they do not reach `_res_live`. Confidence is high — the failing step was identified
by instrumented replay on the real daemon and the stack-free proof reproduces it byte for byte — but the fix is proven
only once the real VC8/VC9 rerun passes; the full local run and the source review remain pending; attempt count
unchanged.

## Final qualification, maintenance and merge record (coordinator, 2026-10-23)

Recorded after the merge of MR !82. Failed runs and pipelines stay failed; nothing below relabels one as green. Identities named as "inferred" were reconstructed from code and registry order, not measured, because the original event streams of the failed runs were not retained.

### Local full runs (all FAILED, preserved)

| Run | Source | Evidence | Outcome |
| --- | --- | --- | --- |
| 1 | `c070aff8` | `qual-pre-c070aff8*` | FAILED (581 s) at the contracts-end false foreground latch. |
| 2 | `88c51416` (corr4) | `qual-pre-88c51416*` | FAILED (597 s) at the same latch. |
| 3 | `4f73e412` (corr5) | `qual-pre-4f73e412*` | FAILED (1548 s). The contracts latch did not recur. Passed: lint, lint-base, lint-web, lint-core, provenance, contracts, root. Latch in migrations: `accounting: anonymous volume g001-mig-<id> of <container> not verified removed`. Mechanism, reproduced stack-free (`f0-volrecreate` 7/10 on the base, 10/10 fixed): `compose down -v` destroyed and recreated the Compose-named volume `g001-mig-<RUN_ID>_postgres-data`, and `_res_witness_parse`'s `volume/create` did not clear `vgone[name]`, so `_res_vol_gone` found the recreated volume. The original event stream is unavailable, so the identities are inferred. |
| 4 | `9cffb495` (corr6) | `qual-pre-9cffb495*` | FAILED (6207 s). Passed: lint, lint-base, lint-web, lint-core, provenance, contracts, root (with B-DOCS-EXEC-ROOT), migrations (past the run-3 latch), ops, rename (with B-BROWSER-RENAME), off: B-TOGGLE-OFF-ADDITIONS, B-OBS-OFF, B-TOGGLE-OFF, and the controls child. Failure A: B-BROWSER-OFF, scope S-OFF latched `ceiling container 8/7` (scope latch only, lifted at scope end). Failure B: ci-controls 526 of 528, VC8 and VC9 failed. Each run's cleanup was verified and the Docker inventory returned to baseline. |

Merged-SHA fulls of MR !80 (`6ea4cc31`) and MR !81 (`79085025`) also failed and stay recorded as failed. The owner's conditional waiver relayed by Codex (6c940243: if the same false contracts-end latch recurred after corr5, waive its acceptance impact and stop looping) was **not triggered**: the latch never recurred.

### Corrections after run 1 (each with a stack-free focused proof; the Codex source disposition is stated per correction)

- corr4 `88c51416`: background-child classification in `_res_live`. Codex source review returned a **CORRECTION LOOP**, not a SOURCE ACCEPT: the post-return registry registration missed the live attached-run interval.
- corr5 `4f73e412`: the creation-time label `g064.bg=<child>` in `_dro` closed that gap (`f0-bgchild` 14/14 is the corr5 proof); Codex SOURCE ACCEPT.
- corr6 `9cffb495`: `volume/create` clears the gone mark (`f0-volrecreate`, above).
- corr7-A `ecd0cecf`: S-OFF container ceiling 8 (`f0-offceil` 11/11; a ninth container is still refused). The legitimate S-OFF peak is six stack containers (including the exited one-shot minio-init), the retained browser one-off and the capture yq; this derivation is inferred from code and registry order, not measured.
- corr7-B `74091550`: VC8/VC9. The VC driver sources lib.sh with no registry, so `_res_live`'s `awk … "$OWNED_REG"` ran with an empty operand and gawk read stdin, swallowing the compose config piped into `yq_stdin` (the real `docker run -i yq` received 0 of 735 bytes; instrumented replay `run4-vc89/wrap/real.log`). Fix: `"${OWNED_REG:-/dev/null}"` and `</dev/null` on the group command (`f0-stdin` 6/8 before, 8/8 after). Both VC8 and VC9 also failed standalone in about 100 s with no concurrency, so load was ruled out.

### Consolidated local qualification (Codex ACCEPT 8b18f0bc on `740915507b14`, tree `57d572b9`)

Complete attributable coverage is the run-4 passes on unchanged code (all 123 required check IDs appear once in run 4, 512 unaffected CI controls carried forward), plus on the clean corrected SHA: the 16 dependent controls VC1–VC9, V27, V27b, VR1, VR2, VR9, VB1, VB3b, 16/16 PASS (`g064-a1/corr7-ctl`), and the disabled-variant suite PASS with B-BROWSER-OFF (146 tests, S-OFF container peak 7 against ceiling 8, 191 creations, required-checks PASS; `g064-a1/corr7-off`, run through `run.sh off` with an owned registry in the wrapper's format, not through the `validate.sh` wrapper). The first controls launch was refused by the startup guard because of my own leftover VC8 diagnostic image (`g003ctl-6f3da70d051a-vc8-p-svc`, label `g003.control=6f3da70d051a-VC8`, no containers); it was inspected and removed, then the run started. A first `off` launch without a registry was aborted and kept as `corr7-off-noreg-aborted` because it would not have exercised the S-OFF ceiling. Docker inventory 15 containers, 12 volumes, 6 networks, 57 images before and after each rerun. The combined wording is "full required coverage qualified across the recorded runs"; run 4 itself is not relabelled.

### Hosted pipelines, Duo and the dependency maintenance

- Pipeline `2980126450` on `c070aff8`: success; Duo found nothing.
- Pipeline `2981256868` on `740915507b14`: **FAILED** at `B-DEPSCAN` and `B-REQUIRED-CHECKS` (job `17416903967`): the `braces` 3.0.3 exception row matched no current finding (unused rows fail the policy) and a Medium advisory for `opentelemetry-api` 1.45.0 (GHSA-xcq7-p3pf-p88h) was reported. G064 changes no dependency file. Duo requested one change (thread `7c0a31c9`, a bare `go_it_runner_end; r=$?` under `set -e` in the retained proof driver); valid, fixed as `r=0; go_it_runner_end || r=$?` (syntax check only, authorized by Codex).
- Maintenance MR !84 (`MAINT-otel-1.45.1-braces-exception-result.md`): local fast on `c6dfe281` **FAILED** (`B-DEPS`, `B-RENDER-ID`, `B-RENDER-PLAN`, `B-RENDER-REPEAT`; the archetype manifest named the old template hash), then the regenerated renderer output on `b2a4e2a7` passed locally (exit 0, `B-DEPSCAN` 0 findings) and hosted (`2981364350` success, Duo nothing, Codex pre-integration ACCEPT cb310c84). Merged as `b308f848` (parents `840b4e28`, `b2a4e2a7`; tree `8a20b22d` equals the tested tree); source branch deleted.
- MR !82 rebased onto `b308f848` as `b0302883` (tree `c911ec16`, 51 commits over staging); hosted pipeline `2981420316` success, Duo approved with nothing to comment on, open threads 0. Merged as `b5f47088` (parents `b308f848`, `b0302883`); the merged tree `c911ec16` equals the tested hosted source tree, and the source branch is deleted.

Important qualification of "tested tree": the merged tree is **not** the tree of the local qualification SHA `740915507b14`. It differs by exactly eight files, none of them harness code. The list is the output of `git diff --name-status -M 740915507b14 b0302883` (`b0302883` is the hosted-green, merged source; the merged tree `c911ec16` equals its tree):

1. `.archetype/outputs/2ea491555634dde6/manifest.json`, shown by git as a 99% rename from `.archetype/outputs/adda1ceba47bd16a/manifest.json` (the regenerated renderer output; the output directory holds only this one file, so the diff does include it, as one renamed path; without rename detection the same diff lists nine names)
2. `docs/SECURITY.md`
3. `docs/foundation/BOOTSTRAP.md`
4. `infra/security/dependency-exceptions.tsv`
5. `intelligence/uv.lock`
6. `skills/saas-foundation/templates/intelligence/intelligence/uv.lock`
7. `management/v0/results/MAINT-otel-1.45.1-braces-exception-result.md` (new)
8. `management/v0/results/G064-br-proof-driver.sh` (the one-line `r=0; go_it_runner_end || r=$?` fix)

Files 1 to 7 are the maintenance of MR !84; file 8 is the Duo fix. Codex directed (058f84f4, cb310c84) that the accepted `74091550` local coverage carries forward for the unchanged code, with the maintenance's focused evidence and no further local test for the one-line fix. No post-merge staging run was executed, and none is reported as passing; the one-off pre-merge exception therefore rests on that accepted carry-forward, not on an identical tree.

### Cleanup receipt (2026-10-23)

Removed: worktrees `g064-proof`, `g064-full3` to `g064-full7`, the maintenance worktree and the writer 17 to 20 worktrees (all clean, no unique unpushed commits; the SHAs of the four failed local full runs, `c070aff8`, `88c51416`, `4f73e412` and `9cffb495`, stay reachable from the remote temporary branches `g064-reuse18` to `g064-reuse21`, as do the locally qualified SHA `74091550` and the corr7-A commit `ecd0cecf` (both reachable from `g064-reuse21`); `74091550` is the qualified source, not a failed run); the local maintenance branch; the five `g064f1-*:t` proof images (images 57 to 52, containers 15, volumes 12, networks 6). Kept: failed-run evidence under `platform-evidence/g064-a1/`, the checkpoint file, the remote temporary `g064-*` branches and the local `worktree-agent-*` refs pending an owner decision, and the volume `project-archetype-render-cache`, which is not owned by this task.
