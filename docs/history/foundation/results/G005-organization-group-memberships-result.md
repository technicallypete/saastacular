> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G005 — Organization and Group membership model: Result

**Status: SUBMITTED FOR CODEX IMPLEMENTATION REVIEW. G005 is not accepted; only Codex may accept it and update
PROGRESS.**

**State (2026-10-09).** Implementation, a read-only security review and its bounded corrections (cycle 1 of B7)
are complete. Required local coverage of the final reconciled code passes:
- r5 on the rebased code;
- r1, r2 and r4 under explicit byte-identity provenance;
- the F-1 forced failure, which fails as required;
- L1 after the target's docs-only advance.

The final candidate is the rebased branch on `origin/staging` `8f5f453`, plus the commit adding this Result.

**Hosted state.** MR !10 (not included) is open from
pushed source `f862f1843fb96636d8801e6f0ae751bfa3a9db5c`.
- **Automatic MR `fast` job: passed 33/33.** It ran in MR pipeline 2931660933, alongside push pipeline 2931660648.
- **MR gate:** manual and not started.
- **Staging gate: pending.** This is the required full hosted evidence, after an owner-authorized conditional
  integration.

Nothing is merged. This hosted record is a local commit, not pushed, per the coordinator's correction.

## Bindings and start gate

| Input | Value |
|---|---|
| Authorization | Execution release (not included) `02efd1e8fe321960c7691a86ced3e37976a0deb4bfcc13205d29a3c79fd1df5a` (read from the root checkout; not copied); checkout relocation confirmation (not included) `e1d0ed9a082456fa8d5cf64010c6bd4894dbfac25017ce5a3806b480f7a08ac8` (read from the root checkout; not copied) |
| Execution checkout | `.claude/worktrees/agent-<id>` (the parent's harness isolation worktree), in place of `.claude/worktrees/g005-exec`: a mechanical location exception confirmed by Codex. Base, tree, snapshot, scope and `path_re` bytes unchanged |
| Branch | `g005-organization-group-memberships`, created at BASE after the relocation proof |
| BASE | `0d524212f185d9157719bc92671c8d098bc249eb`, tree `706ad4de7eebdd67d9bdc8c811ba921cc2370396` |
| Approved bundle | `<home>/.local/state/platform-evidence/g005-execution-binding-proposed/`: `binding-proposal.json` `d1ee5411…975e3`, `approved-inputs.sha256` `fb295f31…6ee74c` (75 inputs), `scope-allowlist.txt` `4796482e…2d2dc` (12 expressions), `path_re` `0d49ad2b…fdf13` (each recomputed and matched) |
| Plan | r4 `6e8e7cd0dad1ac9183f383a8c5ceb0dd48b7dafa34b13ca9d8b15f4b9f648251` with the r4 review (not included) `201cfc8f…539d62` (unchanged; every review disposition kept) |
| Scope copy | `management/v0/results/G005-scope-allowlist.txt`, `cmp`-identical to the bundle file (`4796482e…2d2dc`), first execution write (commit `e79ac7c`; after the rebases `fd308f2`) |
| Budget | B7, checkpoint after cycle 4 |
| Commit history | Tested pre-rebase commits `e79ac7c`, `4aa6db8`, `9bc7efe`, `2324f94` (r1–r3) and `0fc1731` (cycle 1; r4, F-1) on BASE; rebased onto `070882c` as `96a81bc`…`ed3ff2c` (r5) and onto `8f5f453` as `fd308f2`, `9ccc0c1`, `590a974`, `94e2e04`, `8818e86` (L1), plus this Result |

**Relocation proof (before any write; `s0/relocation-proof-before-scope-copy.txt`).**
1. `git fetch origin`, then `git checkout --detach 0d524212…`: HEAD `0d524212f185d9157719bc92671c8d098bc249eb`, tree
   `706ad4de7eebdd67d9bdc8c811ba921cc2370396`.
2. Every one of the 771 BASE tree entries is byte- and mode-identical on disk (`git hash-object --no-filters` against the
   tree blob, executable bit against the mode): 0 mismatches, 0 extra paths, `git status --porcelain --ignored
   --untracked-files=all` empty, no root `.env`.
3. Wrapper precheck with the bundle fields: `run-approved: precheck passed (75 inputs, snapshot
   fb295f311595d474c92c24e3feb6808b5abf015f1dce1cc3b788cda3af6ee74c)`, exit 0.

**S0 (`s0/`).** Docker 27.5.1. No resource of the root project `beginify` (0 containers, 0 volumes); every harness port
set (`PORTS_MIG/OPS/REN/OFF/IDN/STG` and the defaults) free. Unrelated resources seen and left untouched:
`payments-assistant-*` (postgres on 5433), `outreacher-*`/`outreacher-pg` (5432), `needknow-*` (7777/7778), their
volumes and networks, and the `project-archetype-render-cache` volume. GitLab compute-minute usage (read-only GraphQL):
October 2026 = 3478 minutes (purchased minutes are not shown). Every non-check-only wrapper run, and the F-1 run, held
the shared `<home>/.local/state/platform-evidence/local-harness.lock`.

## Implementation summary

One forward migration, one internal package and their tests, plus the bounded harness, CI-map, log-catalogue and
documentation reconciliation of Plan r4. No dependency, route, contract, configuration, Compose, Dockerfile,
`.gitlab-ci.yml` or `main.go` change.

- **`core/migrations/00003_tenancy.sql` (Plan r4 §3, DS-1..DS-9).** `app.organizations`, `app.groups` (editable `slug`
  with `COLLATE "C"`, grammar and 24-entry reserved CHECKs, `UNIQUE (organization_id, slug)`) and `app.role_bindings`
  (the only role store: subject FK to `app.users(id)`, Organization FK, composite same-Organization Group FK, one closed
  two-branch `role_bindings_scope_consistent` CHECK over the six owner/admin/member codes and source `membership`,
  `UNIQUE (subject_user_id, scope_type, scope_id)`), eight indexes, three per-table `*_keys_immutable` BEFORE UPDATE
  triggers (SQLSTATE 23000, `search_path=pg_catalog`, not SECURITY DEFINER), two read-only `security_invoker`
  projections, and column-level runtime grants with the G004 preflight and an exact table-and-column ACL
  postcondition. All foreign keys are `ON DELETE RESTRICT`; no cascade, no RLS (G008), no owner/creator column. Down
  drops exactly these objects.
- **`core/internal/tenancy` (Plan r4 §4).** `tenancy.go` (distinct UUID types, parsing, roles, entities), `names.go`,
  `slug.go` (grammar, reserved set, deterministic base, `base-2`..`base-99` candidates), `errors.go` (bounded codes; the
  constraint and SQLSTATE mapping; `Unwrap` exposes only context errors), `store.go` (all 20 statements, each with the
  explicit Organization predicate), `service.go` (the 20 pinned methods; one READ COMMITTED transaction each, success
  only after commit, deferred rollback; atomic `CreateGroup` = Group plus the explicit creator's `group_owner`
  binding, no Organization membership) and `log.go` (one `tenancy.operation` record per mutation). It imports only
  the standard library, pgx (`pgx`, `pgconn`, `pgtype`) and `core/internal/observability`, and is linked into neither
  binary.
- **Harness and CI (Plan r4 §6).** Version-3 reconciliation of every migration-version and whole-schema site; the new
  B-TENANCY-CATALOG, B-TENANCY-MIGRATION (with a real PostgreSQL stop/start between seed and verify) and B-TENANCY-IT
  (with `logcheck validate`/`scan` and a planted-canary control) in `migrations`, and B-TENANCY-HTTP in `root`;
  B-MCP-SOURCE asserts neither binary links the package; B-MIGRATE-SET pins the accepted 00001/00002 digests;
  `APPLICABLE` gains the package and `TENANCY.md`. Registry 83 rows (fast 33, integration 71, full 81, staging 5).
  `profile-paths.tsv` gains `foundation dir core/internal/tenancy/` and `validate-controls.sh` the `F-tenancy` control.
  `events.json` gains the closed `tenancy.operation` event (outcome/level/code/scope/correlation rules).
- **G004 test reconciliation (P-G004-6 (a)–(d), mechanical).** Both identity fixture `TRUNCATE` literals gain `CASCADE`
  (byte-identical; ESC-1 option B); `TestIdentityCatalog`'s exact table list is identity ∪ the three tenancy tables;
  `identity_catalog`'s exact relation set is identity ∪ the 13 tenancy relations. No identity semantics, guard, route
  or production file changed.
- **Docs.** New `docs/foundation/TENANCY.md`; `MIGRATIONS.md` (current set, down implications, foundation-stage RLS
  note per Q1, validation); `FOLLOW-UP.md` STAGE-TENANCY partial ticks; the approved `IDENTITY.md` warning (R4-Q3)
  immediately before the `identity:staging-local` block, command bytes unchanged.

### Files (33 paths, all inside the 12 scope expressions; B-SCOPE PASS)

Created (19): `core/migrations/00003_tenancy.sql`; `core/internal/tenancy/{tenancy,names,slug,errors,store,service,log}.go`;
`core/internal/tenancy/{tenancy,names,slug,errors,parity,boundary,integration}_test.go`;
`core/internal/migrate/tenancy_integration_test.go`; `docs/foundation/TENANCY.md`;
`management/v0/results/G005-scope-allowlist.txt`; this Result.

Modified (14): `core/internal/migrate/{integration,runner}_test.go`; `core/internal/identity/{integration,store}_test.go`;
`core/internal/identity/workos/staging_test.go`; `infra/validation/{run.sh,required-checks.tsv}`;
`infra/validation/tools/internal/logrecord/{events.json,logrecord_test.go}`; `infra/ci/{profile-paths.tsv,validate-controls.sh}`;
`docs/foundation/{MIGRATIONS,FOLLOW-UP,IDENTITY}.md`.

## Migrations and schema

Version 3, `00003_tenancy.sql` (forward; 00001/00002 byte-identical, pinned by B-MIGRATE-SET). Relations at version 3:
the 14 identity relations plus 13 tenancy relations (3 tables, 2 views, 8 indexes) = 27, plus three trigger functions
and three triggers. Runtime privilege matrix exactly Plan r4 §3.3 (asserted by the migration postcondition and CT-3).
Down to 2 destroys tenancy data only (MG-5: identity rows and the version-2 catalog unchanged); refused with
`APP_ENV=production` by the existing runner. **Risk handling:** additive; the only coupling is the subject FK onto
`app.users`, whose fixture consequence is the ESC-1 `CASCADE` (below).

## Contracts, API and MCP

No public API, BFF, MCP, CLI, environment or route change. B-TENANCY-HTTP proves 36 route-absence probes (GET and
`POST {}` on 10 candidate core routes, own namespace and host port, no model payload; 8 web routes) return 404, and
B-MCP-SOURCE proves neither binary links the package. The only new contract is the internal Go surface (Plan r4 §4.2,
pinned by U-BOUNDARY) and the `tenancy.operation` log event.

## Tests added

| Layer | Tests |
|---|---|
| Unit (`core-check`; B-UNIT-CHECKS, B-CHECKS-ROOT) | U-ID/U-ROLE/service pre-SQL validation (`tenancy_test.go`, 33 rejection cases), U-NAME, U-SLUG (24 reserved, 18 grammar negatives, the §4.4 table, candidates), U-ERR (34 mapping cases, bounded `Error()`, `Unwrap`), U-PARITY (Go vocabulary and bounds equal the embedded SQL), U-BOUNDARY (imports, no `ServeHTTP`, exactly 20 `Service` methods; SQL writes only to the three tables, roles only to `role_bindings`, no projection writes, Organization predicate everywhere, no SQL outside `store.go`); migrate `runner_test.go` reconciled |
| B-TENANCY-CATALOG | `TestTenancyCatalog`, 21 subtests: runtime identity; CT-1..CT-5; CT-6 `authorityViolations`; CT-6-NEG N1, N2, N3, N4, N5, N6, N7a, N7b, N8, N9a, N9b (each in a rolled-back owner transaction yields exactly its codes; catalog unchanged afterwards); CT-7; runtime DDL/role/metadata denials |
| B-TENANCY-MIGRATION | `TestTenancyMigration`, 6 subtests (MG-1, MG-2, MG-4, MG-3, MG-5, ESC-1), plus `TestTenancyPersistSeed`/`TestTenancyPersistVerify` around a PostgreSQL stop/start (MG-6) |
| B-TENANCY-IT | `TestTenancyModel`, 35 subtests (runtime identity; IT-ID-1..5; IT-REF-1..4 incl. both commit orders and 50 randomized rounds; IT-VOC-1 480-combination matrix; IT-UNIQ-1/2; IT-PROJ-1..3; IT-MULTI-1/2; IT-OWN-1..6; IT-SLUG-1..6; IT-SCOPE-1; IT-ATOM-1/2; IT-CONC-1; IT-LOG-1) |
| B-CONTRACT-TOOLS | `logrecord_test.go`: catalogue list plus 7 accepted and 18 rejected `tenancy.operation` records |
| B-CI-CONTROLS | `F-tenancy` (MR adding `core/internal/tenancy/service.go` → pipeline full, gate full, `foundation`) |
| Reconciled | `TestMigrationLifecycle` subtests 02/05/07/08/09 and `reset`; `TestIdentityCatalog` table list; the identity fixtures' `CASCADE` |

## Validation commands and outcomes

All harness runs are the accepted wrapper, launched by `g005-vw.sh` (BASE binding), `g005-vw2.sh` (binding reconciled
at `070882c`) or `g005-vw3.sh` (reconciled at `8f5f453`; check-only in L1). Each launcher re-verifies the bundle digests, then runs, under
`flock <home>/.local/state/platform-evidence/local-harness.lock`:

```sh
bash infra/validation/run-approved.sh --snapshot <S> --snapshot-sha256 <D> --count 75 --path-re <bundle path_re> \
  --plan management/v0/plans/G005-organization-group-memberships-plan.md -- \
  --base <B> --allow-file management/v0/results/G005-scope-allowlist.txt --report-dir <EV>/runs/<step>/report <suites>
```

It records NUL-safe `git status` and the Docker/port inventory before and after each run. Evidence root:
`<home>/.local/state/platform-evidence/g005-exec-0d52421/` (`runs/<step>/{meta.txt,harness.out,report/}`,
`runs.tsv`, `s0/`, `reconciled-070882c/`, `reconciled-8f5f453/`, `rebase-reconciliation.txt`,
`rebase-reconciliation-2.txt`; launcher and helper scripts in `bin/`).

| Run | Suites | Tested commit (tree) | `--base` / snapshot | Outcome |
|---|---|---|---|---|
| r1 | lint unit contracts provenance migrations | `2324f94` (`0066938…`) | `0d52421` / `fb295f31…` | **exit 0**; 41 PASS, 0 FAIL/UNMET; B-REQUIRED-CHECKS 40/40; pre/post passed; B-SCOPE 32 paths; B-PROTECTED 314 files + 75 expected hashes |
| r2 | root ops rename off | `2324f94` | `0d52421` / `fb295f31…` | **exit 0**; 42 PASS; 41/41; pre/post passed; B-TENANCY-HTTP 36 probes; B-BROWSER-ROOT 141 tests; B-IDENTITY-IT 57 tests |
| r3 | controls | `2324f94` | `0d52421` / `fb295f31…` | **exit 1** (recorded failure, not a product defect): B-CI-CONTROLS PASS (274 controls incl. F-tenancy), but a controls-only run generates no secrets, so B-SECRET-CANARY is UNMET and B-REQUIRED-CHECKS fails. Controls are thereafter run with a secret-generating suite |
| (review) | read-only security/SQL review of `2324f94` | — | — | No blocker; two should-fix and several nits → cycle 1 (below) |
| r4 | lint unit controls provenance migrations ops | `0fc1731` (`233881c…`) | `0d52421` / `fb295f31…` | **exit 0**; 50 PASS, 0 FAIL/UNMET; 49/49; pre/post passed; B-CI-CONTROLS 274; B-TENANCY-CATALOG 21/21, B-TENANCY-MIGRATION 6/6 + restart, B-TENANCY-IT 35/35 (1745 records, logcheck clean, planted control detected); B-IDENTITY-IT 57; B-SECRET-CANARY 157 secrets |
| F-1 | `run.sh … migrations` in a scratch clone of `0fc1731` with IT-UNIQ-1 inverted (`ok.Load() != 2 \|\| exists.Load() != 30`) | `0fc1731` + 1-line test edit | `0fc1731` | **harness exit 1 as required**: B-TENANCY-IT FAIL (`--- FAIL: TestTenancyModel/IT-UNIQ-1…`), B-REQUIRED-CHECKS FAIL, the other 10 checks PASS; scratch removed (`runs/f1-forced-failure/`) |
| r5 | lint unit controls provenance root off | `ed3ff2c` (`fea633d9d7a1c36df44e92e10c760cc59f23c2c3`), rebased on `070882c` | `070882c` / reconciled `bee7341d…` | **exit 0**; 51 PASS, 0 FAIL/UNMET; B-REQUIRED-CHECKS 50/50; pre/post passed; B-CI-CONFIG and B-SHELL on the merged `run.sh`/G050 files; B-CI-CONTROLS 282 controls (G050 J1–J6 and G005 F-tenancy together); B-PROTECTED 321 files + 75 expected hashes; B-SCOPE 32 paths since `070882c`; B-TENANCY-HTTP 36 probes; B-CHECKS-ROOT; B-BROWSER-ROOT 141 and B-BROWSER-OFF 141 tests; B-TOGGLE-OFF version 3 |
| L1 | wrapper `--check-only` with the snapshot reconciled at `8f5f453`, plus B-SCOPE, B-PROTECTED and B-LITERALS (for `OPERATIONS.md`) evaluated with the harness's semantics by script, no suite rerun | `8818e86` (`da0e9aa12bfdba12d324d74f010b18ddb25442c3`) | `8f5f453` / reconciled `d5f47345…` | precheck passed (75 inputs), exit 0; scope 32 paths in the allow-list; 321 protected files equal the base; no literal (`reconciled-8f5f453/light-checks.txt`) |

Development-only checks (not harness evidence): the dev helper ran the unit, vet and integration packages against an
owned disposable PostgreSQL (`g005dev-pg` on `g005dev-net`, tmpfs data, no published port), `logcheck validate/scan` on
a dev operation log (1745 records, clean) and a scratch negative control for the new SQL-boundary assertion (fails as
required). Unit totals in `core-check`: every core package passes `gofmt`, `go vet ./...` and `go test ./...`.

### Coverage of the final tree `8818e86`

The final candidate is the rebased branch: the five G005 commits on `origin/staging` `8f5f453` (`fd308f2`, `9ccc0c1`,
`590a974`, `94e2e04`, `8818e86`, tree `da0e9aa…`), plus the commit adding this Result. Its code is byte-identical to
r5's tree `ed3ff2c`: the target's only advance after r5 (`b803db2`, `8f5f453`) is docs-only governance (AGENTS.md,
CLAUDE.md, GITLAB-BRANCH-WORKFLOW.md, OPERATIONS.md, PROGRESS.md); of these only `docs/foundation/OPERATIONS.md` is read
by a check (B-LITERALS scans it, re-checked in L1; B-TOGGLE-OFF-ADDITIONS copies it).
Every required suite covers it either by a run on that tree or by byte identity of everything its checks depend on:

| Suite | Evidence for the final tree |
|---|---|
| lint, unit, controls, provenance, root, off | r5 on `ed3ff2c` (code byte-identical to the final tree; the docs-only advance is covered by L1; the Result commit adds one Markdown file under `management/` that no check executes) |
| migrations, ops | r4 on `0fc1731`. All 32 G005 paths are unchanged in content (30 identical blobs; the two merged files keep identical G005 hunks, `rebase-reconciliation.txt`). The target's changes since BASE are `infra/ci/*` and `.gitlab-ci.yml` (not executed by these suites), the harness `run.sh` hunks confined to `suite_lint`'s B-SHELL list and `static_ci_config` (lint only), `docs/foundation/OPERATIONS.md` (copied into the off variant, not executed) and governance/management Markdown. Every other file these suites execute is byte-identical |
| contracts, rename | r1/r2 on `2324f94`. Cycle 1 changed only tenancy/migrate tests, `service.go` (linked into no binary) and `TENANCY.md`, none of which these suites execute or compare beyond the repository copy that B-SOURCE-HASH checks against itself; the target's changes are as described for migrations and ops |
| F-1 | `0fc1731`; the G005 code is byte-identical in the final tree |

Hosted: see [Hosted MR pipeline](#hosted-mr-pipeline). Under the solo-operator policy the MR runs `fast` automatically;
its `gate` is manual and optional and was **not** started. The required full hosted evidence is the automatic staging
gate after an owner-authorized conditional integration: **pending**.

## Acceptance criteria (Spec AC1–AC12)

Local run identifiers are in [Validation](#validation-commands-and-outcomes); the hosted pipeline is in
[Hosted MR pipeline](#hosted-mr-pipeline).

| AC | Outcome | Evidence |
|---|---|---|
| AC1 canonical Organization/Group identity and ownership | Met | CT-1 (exact object set and columns; no tenant/Agency/BusinessProfile/billing/policy/audit object; no owner column), CT-2, CT-6; IT-ID-1 (distinct nonzero v4 UUIDs, Group parent), IT-ID-2 (renames and slug edit keep UUID, parent, `created_at` and the binding row), IT-ID-3 (runtime 42501 on every key column and on explicit `id`/`created_at`), IT-ID-4 (owner 23000 `*_keys_immutable`), IT-ID-5 (23514 `*_nonzero`); U-BOUNDARY (no move/reparent/delete method) |
| AC2 bindings reference valid users and same-Organization typed scopes | Met | IT-REF-1 (runtime grants of all six roles, exact columns), IT-REF-2 (service: unknown/foreign Organization, Group and user; zero/malformed IDs; byte-identical unknown and foreign errors), IT-REF-3 (runtime SQL: 23503 on each named FK incl. the composite Group FK, 23514 `role_bindings_scope_consistent`/`role_bindings_subject_nonzero`, referenced-user key change fails on `role_bindings_subject_fkey`), IT-REF-4 (deterministic and 50 randomized parent-delete races; no dangling row), CT-4; row-set hashes unchanged after every negative |
| AC3 supported role/scope/source only; one membership per subject and scope | Met | IT-VOC-1 (480 runtime combinations: exactly the six valid ones insert; every other, incl. cross-scope `group_owner` at Organization scope, BusinessProfile scopes, `direct_grant`/`invitation`/`system` sources and `advisor`/`principal`/`associate`, fails 23514 on that one named CHECK; service `invalid_role`), IT-OWN-1 (`group_owner` creation), IT-PROJ-1 (`group_owner` change), CT-4 (two branches, six codes), IT-UNIQ-1 (20 rounds × 2 scopes × 32 concurrent grants: exactly 1 success and 31 `membership_exists`), IT-UNIQ-2 (deterministic 23505 `role_bindings_subject_scope_key`) |
| AC4 read-only projections follow one binding immediately | Met | IT-PROJ-1 (grant → change → change → remove → regrant: 1/1/1/0/1 rows, same UUID across changes, new UUID on regrant, exact projection shapes incl. the Group view's `organization_id`, creator and unrelated bindings unchanged), IT-PROJ-2 (runtime INSERT/UPDATE/DELETE on both views: 42501), IT-PROJ-3 (separate and concurrent readers never see two rows or an uncommitted role), CT-3, CT-6 A1–A3 and CT-6-NEG N1–N9b, U-BOUNDARY |
| AC5 one global user, independent scoped roles | Met | IT-MULTI-1 (one seeded G004 user with `org_owner`@A, `org_member`@B, `group_owner`@own Group, `group_admin`, `group_member` ×2; change and removal leave every other binding and the user, principal and session rows unchanged), IT-MULTI-2 (nothing synthesized across scopes; a creator has no Organization membership); no identity/provider import (U-BOUNDARY) |
| AC6 slug suggestion/edit/uniqueness/reserved rules | Met | U-SLUG; IT-SLUG-1 (all 24 reserved by service create, service edit and runtime SQL), IT-SLUG-2 (18 grammar negatives by service and runtime SQL), IT-SLUG-3, IT-SLUG-4 (16 concurrent creates, 16 concurrent edits, create versus edit, deterministic interleaving: exactly one winner, losers unchanged), IT-SLUG-5 (`Acme Revenue` → `acme-revenue`, `acme-revenue-2`, `api-2`, `slug_required`, suggestion taken before create), IT-SLUG-6; IT-ID-2 |
| AC7 scoped, transactional, fail-closed operations incl. creator ownership | Met | IT-SCOPE-1 (8 operations with a foreign Group: `group_not_found`, identical to unknown; B rows unchanged), IT-ATOM-1 (11 rejection classes, each with its exact code; rows unchanged), IT-ATOM-2 (commit failure, statement failure, closed pool, mid-transaction cancellation: `unavailable`, no result, rows unchanged), IT-CONC-1 (32 concurrent changes keep one row and UUID; last committed wins; change versus remove both orders); IT-OWN-1..6 (Group plus exactly one `group_owner` binding; unknown/zero/malformed creator or unknown Organization leaves neither; injected failure after the Group insert, commit failure and cancellation between the writes leave neither; creator deleted mid-creation rolls the Group back, and the reverse order fails the delete; 16 concurrent creations give one Group with its own creator's binding; owner bindings are ordinary rows, no last-owner policy claimed) |
| AC8 no public surface or future-Goal implementation | Met | B-MCP-SOURCE (neither binary links the package; `main.go`, imports and routes at the G004 pins), B-TENANCY-HTTP (36 probes 404), B-WEB-BOUNDARY, CT-1/CT-7 (no BusinessProfile/GroupBusinessProfile, policy, RLS, onboarding, routing, staffing or audit object), B-ENV-EXAMPLE, B-SCOPE |
| AC9 migrations, persistence and least privilege on real PostgreSQL | Met | B-TENANCY-MIGRATION (MG-1..MG-6), B-TENANCY-CATALOG (CT-1..CT-5; runtime CREATE/ALTER/DROP/TRUNCATE/DELETE, trigger create/disable, `SET ROLE` owner/migrator and metadata read all 42501), reconciled B-MIGRATE-CLI/CATALOG/IT, B-PERSIST, B-DBRENAME, B-MIGRATE-SET (accepted migrations byte-identical) |
| AC10 inherited contracts and redaction | Met locally (every suite covers the final tree per [Validation](#validation-commands-and-outcomes)); hosted full pending the staging gate | B-IDENTITY-* (catalog, IT, E2E, probes, privacy, config), browser, first-paint, token, boundary, variant, dependency and provenance IDs; IT-LOG-1 plus `logcheck validate` (every operation record catalogued), `logcheck scan` (no canary) and the planted control; B-SECRET-CANARY; B-CONTRACT-TOOLS |
| AC11 container and hosted CI gate the submitted model | Local part met; MR `fast` passed (33/33, pipeline 2931660933); hosted full **pending** (staging gate after conditional integration) | Wrapper runs on the actual root and hash-matched copies (r1–r5); F-1 forced failure (harness exit 1, B-TENANCY-IT FAIL); B-CI-CONTROLS (a FAIL record fails the gate); MR `fast` per [Hosted MR pipeline](#hosted-mr-pipeline); the required full hosted job is the automatic staging gate under the solo-operator policy |
| AC12 provenance and bounded Result | Met, pending Codex review | Bindings, relocation proof, wrapper pre/post on every run, NUL-safe status, scope, start/end inventories, failures, B7 accounting below; no policy, RLS, onboarding or product claim |

## Deviations (architecture deviations: none)

1. **Checkout relocation.** Executed in the parent's isolation worktree `.claude/worktrees/agent-<id>`
   instead of `g005-exec`; confirmed by Codex as a mechanical location exception (record above). BASE, Plan, snapshot,
   scope, `path_re`, branch and B7 unchanged; `g005-exec` and every other worktree were never touched.
2. **Validation placement (owner policies, applied in order).** The execution release first replaced Plan S7's
   duplicate final local `all` with the MR pipeline's fast/full evidence. The later solo-operator policy (staging
   `070882c`, relayed by the parent) supersedes that for G005: the required local coverage must cover the final
   reconciled code before the MR; the MR runs `fast` automatically; its `gate` is manual/optional and was not started;
   the required full hosted evidence is the automatic staging gate after an owner-authorized conditional integration.
   Local coverage of the final tree therefore comes from r5 (rerun of the suites whose inputs changed) plus r1/r2/r4
   reused under the byte-identity mapping in [Validation](#validation-commands-and-outcomes). No local
   `validate.sh --profile full` parity run and no repeated unchanged suite.
3. **Integration route.** The accepted MR route (rebase onto `origin/staging`, push, `glab mr create` to `staging` with
   source-branch deletion) replaced the Plan's proposed X1/X2 evidence-branch route. No protected-branch update,
   `identity-qualification` push or provider action (ESC-1 option B: no post-G005 live staging run; residual below).
4. **IDENTITY.md** is in scope with the approved warning (R4-Q3); `ops` (which executes IDENTITY.md's `identity:*`
   blocks verbatim) ran in r4, whose inputs are byte-identical in the final tree.
5. **Lock-wait detection** in the race tests uses `pg_locks` (ungranted) joined to `pg_stat_activity` for backends of the
   test database, rather than `pg_stat_activity` wait events alone: PostgreSQL hides other roles' wait events from a
   non-superuser, and the tests run without a superuser.
6. **Step order.** S2–S5 were implemented before the first wrapper run instead of one wrapper gate per step; the step
   gates' suites were then run as wrapper runs r1–r3 (registry completeness held in every executed suite). The
   security review ran after S5 instead of right after S3.
7. **Scope note.** `TENANCY.md` cites `skills/saas-foundation/references/tenancy-authorization.md` (Codex direction);
   the Codex-owned Spec's older `skills/project-archetype/` link is unchanged.

## Residuals and follow-ups (no new scope)

- **ESC-1 option B residual.** `core/internal/identity/workos/staging_test.go`'s `TRUNCATE … CASCADE` was validated by
  inspection plus the byte-identical `resetIdentity` statement (B-IDENTITY-CATALOG, B-IDENTITY-IT) and by the ESC-1
  subtest at version 3 (plain statement refused 0A000; with `CASCADE` bindings emptied, Organizations/Groups kept, no
  other object cascaded). No post-G005 real-provider staging run was performed; a future authorized qualification
  checks the updated fixture.
- **Security-review nits not applied** (no requirement gap; recorded for Codex): the migration postcondition does not
  separately reject grant options (CT-3's exact `relacl`/`attacl` strings do); trigger-function EXECUTE relies on the
  init script's default privileges (CT-2 asserts owner-only EXECUTE); a lost commit reply returns `unavailable` for work
  that committed (documented in TENANCY.md; a retry then reports a conflict, never a duplicate).
- **PostgreSQL DETAIL lines** on conflicts and foreign-key failures carry UUIDs and slugs (documented). A slug derived
  from a name can therefore reach the server log; the per-run canaries are mixed-case and the scans case-sensitive, so
  the canary scan proves the absence of verbatim names only. Changing the server log configuration is out of scope (Q8).
- Proposed (not authorized): a later Goal could reserve `groups`/`business-profiles` route segments (R3-Q1).

## Rebase onto `origin/staging` and binding reconciliation

### First rebase: `070882c`

Pre-MR rebase (GITLAB-BRANCH-WORKFLOW rule 2) of the tested branch `0fc1731` onto the fetched target
`070882c1e2cfcfbd2c58ade5cc17c5fcfd4fb534` (G050 integration `ff6a97c`/`dae307d` and the solo-operator policy commit).
**No conflict.** Rebased commits: `96a81bc` (scope copy), `2593647` (model and tests), `b9f8adc` (harness and CI map),
`1a57db0` (docs), `ed3ff2c` (cycle 1); `070882c` is an ancestor of the source.

- **G005 content unchanged:** the same 32 G005 paths change against the new base; 30 are byte-identical to the tested
  `0fc1731` blobs, and in the two shared files (`infra/validation/run.sh`, `infra/ci/validate-controls.sh`) the G005
  hunk bodies are identical (hash of the `-U0` `+/-` lines: `2f59009347e4ba7f` and `ba6314349ae6f1bd`, before and
  after). The target's bytes are kept: G050's `static_ci_config` artifact pins, PASS text and B-SHELL lists, the
  policy commit's MR-gate rule pin, and G050's groups A–D, stubs, shims and J1–J6 are untouched by G005's group-F,
  migration, identity-catalog, staging-site and tenancy hunks.
- **Reconciled snapshot** (`reconciled-070882c/approved-inputs.sha256`, SHA-256
  `bee7341d73693e3e3815501eedea0ba24a44940e47ba1877eea7eb8711409302`): the same 75 paths, `path_re` (`0d49ad2b…`),
  count, Plan (`6e8e7cd0…`, unchanged) and scope (`4796482e…`, unchanged); every entry equals the new base's blob. The
  frozen bundle and Plan bytes are not modified. Eight inputs changed, all through integrated target commits:

| Input | Old SHA-256 | New SHA-256 | Why |
|---|---|---|---|
| `.gitlab-ci.yml` | `a04aee56850447c44cb9dd618ca032a8777237d37c2b64205bcb5f5fc27f7a52` | `f95d2f1fcdc957b9e4afe4eea0835df2c0134d9dc98064a47d81ff6ba056dd38` | G050 JUnit `reports` on fast/gate/identity-staging; policy `070882c` MR gate `when: manual`, `allow_failure: true` |
| `AGENTS.md` | `022264431e2f51bd57a515fe3185583562a6883cda1a03a87768d8b44cffb7ee` | `d3dda66f18e9a98293f71f4114983fe515f2bddfcdff0bc387554e3a6becd0a3` | Policy `070882c`: temporary solo-operator validation and integration policy |
| `CLAUDE.md` | `04b7a9b9a5db19da5f3081d7acf3f4fa1137361fa2decc648ecc7e5e3caccea3` | `80ea721f4bfaf315dc408f70761fbf1df237a309535fa70f2aa86bdcfe5a74fc` | Policy `070882c` |
| `docs/agents/GITLAB-BRANCH-WORKFLOW.md` | `3108aeed987afa6a647007d2d7e56fd6874250d768d08a45bde500d0e73605cb` | `418eca25a9b899988e1e6a3dbc9847658d74b74b8f9afe2b8b37edb8ef4b2a4b` | Policy `070882c`: gate placement and conditional integration |
| `infra/ci/validate.sh` | `78a275e5efd2c1214410f82b49d25842ecb6d53f0da97602b59a3a345419d333` | `a3a2b862bd8cf111fc858a3570a742b31f4178e2c4ccba14e63af8af7dc95330` | G050: JUnit conversion and publication |
| `management/v0/BACKLOG.md` | `700d6e50ead49e24a059e93ec440b1b24c0c5a0290618da2172ce25411fdf007` | `c810fca1f83e21c0985d0eb284aa640d51f055f1fa66d43d082d28b4b96280cf` | G051 definition |
| `management/v0/PROGRESS.md` | `c208e1954268aeed3cb54651dc093fab1cda4dc0e993255ff0724c76e2110ba8` | `337d88ce20af30ba3f1e27f66525c36360448570a6594981c3aab01abaea388f` | Progress records (G050/G005 releases, G005 relocation, G050 acceptance and integration, policy) |
| `management/v0/ROADMAP.md` | `33c31f7fc2e96ea49770fcdd4b8f1519e068236fbdeebe048faae0f2a8ca4705` | `3ec51245a5e71421ae25feb7367abe668b91f00181f1900454ec19c08d3a5a48` | G051 roadmap row |

- **Prechecks:** the reconciled wrapper precheck passes (`precheck passed (75 inputs, snapshot bee7341d…)`, exit 0); the
  original snapshot now refuses with exit 11 on exactly those eight inputs (`reconciled-070882c/old-snapshot-check.txt`).
- **Not in the snapshot:** `infra/ci/junit.sh` (new in G050) did not exist at BASE, so the reconciled snapshot keeps the
  approved 75 paths rather than widening `path_re`; G005 does not touch it (B-SCOPE) and r5's lint and controls
  execute it. Codex may bind it explicitly if required.
- r5 ran with `--base 070882c` and the reconciled snapshot, so B-SCOPE, B-PROTECTED and the wrapper pre/post checks are
  BASE-aware against that base. Hosted runs use `--base HEAD` and do not replace them.

### Second rebase: target advanced to `8f5f453` (docs-only)

Before opening the MR the target advanced by `b803db2` and `8f5f453` (governance: rebase before final local
qualification; staging status rule), touching only AGENTS.md, CLAUDE.md, `docs/agents/GITLAB-BRANCH-WORKFLOW.md`,
`docs/foundation/OPERATIONS.md` and PROGRESS.md. The branch was rebased again without conflict onto
`8f5f453b2dae555f82585fe8834db544ec65d584`: `fd308f2` (scope copy), `9ccc0c1` (model and tests), `590a974` (harness and
CI map), `94e2e04` (docs), `8818e86` (cycle 1), tree `da0e9aa12bfdba12d324d74f010b18ddb25442c3`. All 32 G005 paths are
byte-identical to r5's tested `ed3ff2c` (`rebase-reconciliation-2.txt`); `8818e86` differs from `ed3ff2c` only in those
five target files. Per the updated governance, only affected checks are repeated (none of these files is a harness,
CI-config or code input; `OPERATIONS.md` is re-checked for B-LITERALS in L1).

Reconciled snapshot at `8f5f453`: `reconciled-8f5f453/approved-inputs.sha256`, SHA-256
`d5f473456c9d4a139eeca2b3bd049c6d03e637929de49f7dbafca2cf87ddea2c` (same 75 paths, count, `path_re`, Plan, scope).
Against the approved bundle the same eight inputs differ; four of them changed again relative to the `070882c`
reconciliation, all for the reason "docs-only governance update":

| Input | SHA-256 at `070882c` | SHA-256 at `8f5f453` |
|---|---|---|
| `AGENTS.md` | `d3dda66f18e9a98293f71f4114983fe515f2bddfcdff0bc387554e3a6becd0a3` | `00d8a79b2ca3d0c56f47e3fa56c1419e908414b18cde8253d30ed7ebc0b97aab` |
| `CLAUDE.md` | `80ea721f4bfaf315dc408f70761fbf1df237a309535fa70f2aa86bdcfe5a74fc` | `8490b1d35c56f4dec047638bd0e64d71c126e712756c01b9cd460cbccbcce28b` |
| `docs/agents/GITLAB-BRANCH-WORKFLOW.md` | `418eca25a9b899988e1e6a3dbc9847658d74b74b8f9afe2b8b37edb8ef4b2a4b` | `093690f7ed59bb2acbb14a7e06a419900bca21e27130d7f5192d269b0d66727c` |
| `management/v0/PROGRESS.md` | `337d88ce20af30ba3f1e27f66525c36360448570a6594981c3aab01abaea388f` | `d2734e8683895fc3ea48d175660563db96f26c724f9c646db7b1dc089466aa51` |

The `8f5f453` wrapper precheck passes (`precheck passed (75 inputs, snapshot d5f47345…)`, exit 0), and L1 records
B-SCOPE (32 paths), B-PROTECTED (321 files) and B-LITERALS (`OPERATIONS.md`) against `8f5f453`.

## Execution budget

Budget class: B7. Maximum cycles: 7, checkpoint after cycle 4. **Cycles consumed: 1** (main agent 1; security
reviewer 0, read-only).

### Cycle log

1. **Read-only security/SQL review → bounded corrections.** The review of `2324f94` found no blocker; its two
   should-fix items and five nits were corrected in `0fc1731` (IT-ATOM-1 exact codes; lock waits tied to the test
   database's backends; IT-ID-3 explicit `id`/`created_at` on `role_bindings` and `created_at` on `groups`; hash
   prefixes instead of named entities in failure messages; no SQL outside `store.go`, with a scratch negative control;
   deferred rollback in `Service.inTx`; TENANCY.md log-DETAIL and lost-commit notes). Retested by dev runs, then r4.

Not counted: the first dev runs of the freshly written integration tests found two test-code parameter-typing errors
(session seed `$1` used as two types; IT-VOC-1's `$4` unreferenced in Organization-shaped rows), fixed before any
harness validation as part of the initial implementation. r3's exit 1 was a suite-selection error in the validation
procedure (no product change).

**Checkpoint:** not reached (1 of 4). No repeated failure, no G004 semantic change beyond the mechanical
reconciliation, no thrashing. The only failing harness results were r3 (procedure) and F-1 (required).

## Resource cleanup

Owned resources only. The harness removed its own projects, images and volumes per run (start/end inventories in each
`runs/<step>/`). The dev database container `g005dev-pg` and network `g005dev-net` (label `g005.dev=0d52421`) were
removed after the last local run; the dev Go caches live in the evidence directory. The F-1 scratch clone was removed.
Unrelated containers, volumes and networks (`payments-assistant-*`, `outreacher-*`, `needknow-*`,
`project-archetype-render-cache`) were never touched. The end inventory (`s0/resource-inventory-end.txt`) shows the
same unrelated resources as at S0 and no resource labelled for G005 or a harness run.

## Hosted MR pipeline

The candidate is MR !10 (`g005-organization-group-memberships` → `staging`, `force_remove_source_branch: true`, not
draft), source `f862f1843fb96636d8801e6f0ae751bfa3a9db5c`, target `8f5f453b2dae555f82585fe8834db544ec65d584`. The
source was pushed at 2026-10-09T17:58Z. Read-only API data, traces and artifacts are in `hosted/`.

| Pipeline | Source | Selection (trace) | Jobs | Outcome |
|---|---|---|---|---|
| 2931660648 | push (first push of the new branch, before the MR existed) | `profile pipeline=fast executed=fast reason=feature-push base=none` | `fast` 17068826240 | **success**, 375 s. Test report: fast 33/33, 0 failed, 0 skipped. `final status 0; published=full`, cleanup completed |
| 2931660933 | merge_request_event | `profile pipeline=full executed=fast reason=foundation base=8f5f453…` | `fast` 17068829156; `gate` 17068829157 | `fast` **success**, 375 s. Test report: fast 33/33, 0 failed, 0 skipped. `final status 0; published=full`, cleanup completed. `gate` is `manual` with `allow_failure: true`; **not started** |

Under the solo-operator policy (staging `070882c`, refined by `b803db2`/`8f5f453`):
- The MR's comprehensive job is the optional manual gate, which was not run.
- The required full hosted evidence is the automatic staging push gate after an owner-authorized conditional
  integration, judged by the owner status rule: `success`, or `canceled` without a failed job, qualifies.
- That evidence is **pending**. Final ACCEPT and G006 release wait for it.

No pipeline was started manually, retried or canceled. Compute usage before the push was 3771 October minutes,
read-only; purchased minutes are not shown.

## Security and privacy self-review

- No secret, token or WorkOS value was printed, copied or used; the WorkOS staging env file was not read. No live
  WorkOS action, protected-ref update, project/provider setting change, pipeline cancellation or merge.
- The model adds no authority path: no route, no binary linkage, no provider-claim or hostname membership, no global
  role, no BusinessProfile scope, no owner column. Runtime has column-limited CRUD only; keys are immutable by grants and
  triggers; every scope-sensitive statement names its Organization.
- Logging carries bounded codes only; the canary scan and planted control pass; the PostgreSQL DETAIL limitation is
  documented (residuals above).

## Reasoning effort and delegation

One read-only general-purpose reviewer (security/SQL review, no edits, no runs); all files, runs and integration were
done by the main implementer. I cannot independently attest the configured reasoning effort of this session or of the
reviewer; the parent session reports the dispatch configuration.

## Suggested follow-up (not authorized scope)

- Decide whether `infra/ci/junit.sh` should join the G005 binding (see reconciliation).
- After conditional integration, the staging gate's B-TENANCY-* and inherited results bind the integrated content.
- R3-Q1 route reservations and any server-log DETAIL policy belong to later reviewed Goals.
