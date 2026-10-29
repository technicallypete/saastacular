> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G006 — BusinessProfile associations: Result

**Status: SUBMITTED FOR CODEX IMPLEMENTATION REVIEW. G006 is not accepted; only Codex may accept it and update
PROGRESS.**

**Codex decision items (Plan expectations the implementation could not meet literally; escalated, not
self-accepted):** D1 (CT-BP-6-NEG N11's exact expected set) and D2 (IT-BP-ACC-6's SQLSTATE). Both are test
expectations; neither changes the schema, grants or detector. See [Deviations](#deviations-architecture-deviations-none).

**State (2026-10-09 local / 2026-10-10 UTC).** Implementation, a read-only security/SQL review with its bounded
fixes, and cycle 1 of B5 are complete. On the candidate rebased onto `origin/staging` `d4bb478`:
- the final local `W(all unit)` (`s7-final`) passes: exit 0, 85 PASS, 0 FAIL/UNMET, B-REQUIRED-CHECKS 84/84;
- the F-1 forced failure fails exactly as required.

The candidate is `08d279f` plus the commit adding this Result. Hosted: see [Hosted MR pipeline](#hosted-mr-pipeline).
The MR `gate` is manual and is not started. The required full hosted evidence is the automatic staging gate after an
owner-authorized conditional integration (X2, not authorized here): **pending**. Nothing is merged.

## Bindings and start gate

| Input | Value |
|---|---|
| Authorization | Execution binding confirmation (not included) `3d13a8ca06c4e15612ff72e73d392a4c6b05d869566e8d22cc409afb40b77f9e` (read from the root checkout; not copied). Resume instruction via Codex after the usage-limit stop: same Plan, scope, binding and B5 budget, no cycle reset |
| Execution checkout | `.claude/worktrees/g006-exec`, branch `g006-business-profile-associations` |
| BASE | `8488b6ad01f6d8152a6e40f12399e110c962422d`, tree `f256668f912a801f4d4da2a29548cf467086b3ff` |
| Approved bundle | `<home>/.local/state/platform-evidence/g006-execution-binding-8488b6a/`: `binding.json` `bf314a981cf96d0707eee0808a16bf88afe38d14b6d74ff6d0f12de17b71c02d`, `approved-inputs.sha256` `f3acda69…f15600` (74 inputs), `scope-allowlist.txt` `85497a2d…69eda3` (9 expressions); each recomputed and matched by the launcher before every run |
| Plan | r2 `3bd79dc555833cdc55cb5df4549693e9cbc02c9c627091216e1bc852b35842e6` with the r2 review (not included) `b2b15b72…6321` and its six binding clarifications (applied as recorded below) |
| Scope copy | `management/v0/results/G006-scope-allowlist.txt`, byte-identical to the bundle file (`85497a2d…69eda3`), first execution write (commit `7a5d6ab`) |
| Budget | B5, checkpoint after cycle 3 |
| Commit history | `7a5d6ab` (S1), `5ff922c` (S2 WIP checkpoint at the usage-limit stop), `92b6bda` (S2 harness), `be1fe59` (S3 + cycle 1), `eac2b61` (S4); rebased onto `d4bb478` as `82988f7`, `727aa29`, `43c606d`, `7533cc8`, `08d279f` (tested by F-1 and `s7-final`), plus this Result |

**S0 (`s0/s0.txt`, 2026-10-09T19:26:59Z).** Wrapper precheck passed (74 inputs, snapshot `f3acda69…`); HEAD and tree
equal BASE; `git status --porcelain --ignored` empty; no root `.env`; every harness and default port free; root project
`beginify` had 0 containers and 0 volumes. Unrelated running resources seen and left untouched:
`payments-assistant-*`, `needknow-*`, `outreacher-*`. GitLab compute minutes (read-only): October 2026 = 3939. Every
wrapper run and the F-1 run held the shared `local-harness.lock`; G051's child shared it in turn.

**Binding clarifications applied.**
1. RC-6 (the G005 model's `setup()` version pin, 3 → 4) landed in S2 with the migration; the G005 model scenarios run
   unchanged at version 4 and the G005 catalog/migration/persist tests explicitly qualify version 3.
2. Q-r2-9: no BusinessProfile grant, role scope or owner column exists (CT-BP-6/-NEG, CT-BP-8); the Skill and closed
   artifacts were not edited.
3. Q-r2-10: the owner staging-status rule governs the hosted portion; a cancellation is recorded honestly and is never
   model/migration evidence, which comes from the final local full/unit runs on the reconciled candidate.
4. Q-r2-11: `OPERATIONS.md` profile counts were left alone.
5. Q-r2-12: `infra/ci/junit.sh` is in the 74 approved inputs and unchanged.
6. Proportional execution: focused gates as changes landed, then one final full/unit qualification after the
   remote-target rebase.

**Owner sequencing amendment (via Codex, handoff `f3a65c0f`, received 2026-10-09 ~20:45 local).** Before every new
local test/validation batch (focused, preliminary, cycle, background, builds, lint/type checks, controls, full), fetch
`origin/staging`, verify it is an ancestor of the candidate HEAD, and rebase/reconcile the approved input bindings
before launching tests. It is a sequencing rule, not a scope or budget reset; the approved Plan and snapshot bytes are
unchanged. Batch `c1` had already started and was not canceled; it is historical preliminary evidence, reused only with
exact content/input provenance. Every later batch is preceded by a recorded fetch and ancestry check (below).

**Owner direct instruction (2026-10-10 ~02:40Z, in the parent session).** Continue G006 and G051 only under their
approved Plans, exact bindings, budgets and gates; report completed steps rather than restart them; escalate any
material Plan ambiguity; no new scope or permissions. Applied: nothing was restarted or duplicated, and D1 and D2 are
escalated as Codex decision items.

## Implementation summary

One forward migration, the BusinessProfile extension of the internal `tenancy` package and their tests, plus the bounded
harness, log-catalogue and documentation reconciliation of Plan r2. No dependency, route, contract, configuration,
Compose, Dockerfile, `.gitlab-ci.yml`, CI-map or `main.go` change.

- **`core/migrations/00004_business_profiles.sql` (Plan r2 §3).** `app.business_profiles` (database-generated UUID,
  immutable non-null `organization_id` RESTRICT FK, editable `display_name` and optional `legal_name`, `description`,
  `industry`, `website`, `created_at`; bounds/trim/control-character CHECKs and the website CHECK;
  `UNIQUE (organization_id, id)` as the composite target), `app.group_business_profiles` (composite RESTRICT FKs to
  `groups (organization_id, id)` and `business_profiles (organization_id, id)`, `UNIQUE (organization_id, group_id)` =
  at most one association per Group, nothing unique on the profile side), the read-only `security_invoker` view
  `app.business_profile_access_paths` (current Group membership bindings of associated Groups with their raw Group
  role), two `*_keys_immutable` BEFORE UPDATE trigger functions and triggers, column-level runtime grants, the G004
  preflight and an exact table-and-column ACL postcondition. No BusinessProfile role, owner, grant, mirror, RLS,
  onboarding or billing object. Down drops exactly these objects.
- **`core/internal/tenancy` (Plan r2 §4).** `business_context.go` (the context and website contracts, validated before
  any SQL), `business_profile.go` (Create/Get/Update), `association.go` (Link/GetGroup/ListGroupsFor/Unlink and the two
  access-path reads, with ordered Organization-scoped prechecks), `store.go` (the 11 new statements, each with the
  explicit Organization predicate), `errors.go` (eight new codes and the constraint mapping), `log.go` (four new
  operations). `Service` now has exactly 29 methods. One READ COMMITTED transaction per operation; one
  `tenancy.operation` record per mutation; reads unlogged.
- **Harness (Plan r2 §7).** Version-4 reconciliation of every migration-version and whole-schema site (RC-1..RC-3:
  B-MIGRATE-SET four-file set with the 00003 digest pinned, root migrate block, `mig_step`s, P_LOG sets, identity
  catalog, B-PERSIST, B-DBRENAME, off variant); the new B-BUSINESS-PROFILE-CATALOG, B-BUSINESS-PROFILE-MIGRATION (real
  PostgreSQL stop/start between seed and verify) and B-BUSINESS-PROFILE-IT (with `logcheck validate`/`scan` and a
  planted-canary control) in `migrations`; RC-9 adds 20 route-absence probes to B-TENANCY-HTTP (56 total). Registry 86
  rows (fast 33, integration 71, full 84, staging 5) (RC-12).
- **Predecessor test reconciliation (RC-4..RC-8, mechanical).** `TestMigrationLifecycle` at version 4 with the version-3
  step; the G005 catalog/migration/persist tests explicitly qualify version 3 (`runnerFor(…, subsetFS(t, 3))`) and the
  detector compares exact projection source sets; the G005 model runs at version 4; `TestIdentityCatalog` table list +2;
  the runner's normal-set and provider-source pins are the four-file set.
- **Docs.** `TENANCY.md` (BusinessProfile, association, access paths, constraints, grants, Service surface, contracts,
  errors, log operations, future owners, down effects, validation IDs); `MIGRATIONS.md` (version 4, down, validation);
  `FOLLOW-UP.md` (STAGE-TENANCY BusinessEntity item done under ADR-004's names; BillingProfile item partly done).

### Files (27 paths, all inside the 9 scope expressions; B-SCOPE PASS)

Created (9): `core/migrations/00004_business_profiles.sql`;
`core/internal/tenancy/{business_context,business_profile,association}.go`;
`core/internal/tenancy/{business_context,business_profile_integration}_test.go`;
`core/internal/migrate/business_profile_integration_test.go`; `management/v0/results/G006-scope-allowlist.txt`; this
Result.

Modified (18): `core/internal/tenancy/{errors,store,log}.go`; `core/internal/tenancy/{errors,parity,boundary,integration}_test.go`;
`core/internal/migrate/{integration,tenancy_integration,runner}_test.go`; `core/internal/identity/store_test.go`;
`infra/validation/{run.sh,required-checks.tsv}`; `infra/validation/tools/internal/logrecord/{events.json,logrecord_test.go}`;
`docs/foundation/{TENANCY,MIGRATIONS,FOLLOW-UP}.md`. G006 changes none of the 74 bound inputs (B-PROTECTED, B-SCOPE);
the nine input changes reconciled below come only from target commits.

## Migrations and schema

Version 4, `00004_business_profiles.sql` (forward; 00001–00003 byte-identical, pinned by B-MIGRATE-SET). Purely
additive: no earlier object changes (CT-BP-1 compares the version-3 catalog part by part). Runtime privilege matrix
exactly Plan r2 §3.3 (migration postcondition, CT-BP-3). Down to 3 destroys every profile and association and leaves
identity, sessions, Organizations, Groups, slugs and bindings unchanged (MG-BP-5); refused with `APP_ENV=production` by
the existing runner. **Risk handling:** additive; RESTRICT everywhere, so no G005 delete path can cascade into profiles.

## Contracts, API and MCP

No public API, BFF, MCP, CLI, environment or route change. B-TENANCY-HTTP proves 56 route-absence probes return 404;
B-MCP-SOURCE proves neither binary links the package. The only new contracts are the internal Go surface (pinned by
U-BOUNDARY) and four `tenancy.operation` operation names with eight error codes in the log catalogue.

## Tests added

| Layer | Tests |
|---|---|
| Unit (`core-check`; B-UNIT-CHECKS, B-CHECKS-ROOT) | U-CTX (`TestValidateBusinessContext`: field matrix, trimming, absent optionals, bounds, control characters, format characters as data), U-URL (`TestValidateWebsite`: the valid and invalid vectors of Plan r2 §4.3), pre-SQL rejections (`TestBusinessProfileRejectionsBeforeSQL`), U-PARITY (`TestBusinessProfileMigrationParity`: Go bounds and the pinned view text equal the 00004 literals), U-ERR (new codes and constraint mappings), U-BOUNDARY (29 methods; SQL only in `store.go`; Organization predicate everywhere; writes only to the two G006 tables; no projection write) |
| B-BUSINESS-PROFILE-CATALOG | `TestBusinessProfileCatalog`, 29 subtests: runtime identity; CT-BP-1..CT-BP-5; CT-BP-6 (detector with the version-4 inventory; the version-3 inventory yields exactly `unexpected_role_projection`); CT-BP-6-NEG (G005 N1–N9b rerun at version 4 plus N10–N16, each in a rolled-back owner transaction); CT-BP-7..CT-BP-9 |
| B-BUSINESS-PROFILE-MIGRATION | `TestBusinessProfileMigration`, 6 subtests (MG-BP-1, 2, 4, 3, 5, 7), plus `TestBusinessProfilePersistSeed`/`Verify` around a PostgreSQL stop/start (MG-BP-6) |
| B-BUSINESS-PROFILE-IT | `TestBusinessProfileModel`, 35 subtests (IT-BP-ACC-1..7, IT-BP-ID-1..5, IT-BP-CTX-1..4, IT-BP-LINK-1..5, IT-BP-XORG-1..4, IT-BP-UNLINK-1/2, IT-BP-ATOM-1/2, IT-BP-REF-1/2, IT-BP-INDEP-1/2, IT-BP-LOG-1) as the runtime role, with per-run canaries in every context field |
| B-CONTRACT-TOOLS | `logrecord_test.go`: four accepted and three rejected G006 `tenancy.operation` records |
| Reconciled | `TestMigrationLifecycle` (02/05/07/08/09, `reset`), the G005 catalog/migration/persist tests (version 3), the G005 model (version 4), `TestIdentityCatalog`, the runner pins |

## Validation commands and outcomes

All harness runs are the accepted wrapper, launched by `bin/g006-vw.sh` (BASE binding: `--base 8488b6a`, approved
snapshot) or, after the rebase, `bin/g006-vw2.sh` (`--base d4bb478`, reconciled snapshot, preceded by the recorded
fetch and ancestry check). Each launcher re-verifies the bundle digests, runs
`infra/validation/run-approved.sh … -- --base <B> --allow-file management/v0/results/G006-scope-allowlist.txt
--expect-hash … --report-dir <EVID>/runs/<step>/report <suites>` under the shared lock, and records the post-check.
F-1 runs `run.sh` directly in its scratch clone, under the same lock.
Evidence: `<home>/.local/state/platform-evidence/g006-execution-8488b6a/runs/`.

| Step | Suites | Commit | Outcome |
|---|---|---|---|
| s1-lint | lint | `7a5d6ab` | exit 0; postcheck passed |
| s2-a | unit lint controls migrations | `92b6bda` | exit 1: 32 PASS, 3 INFO, 3 FAIL — B-UNIT-CHECKS (`TestProviderAcceptsProtectedStore` still pinned three migration sources), B-BUSINESS-PROFILE-CATALOG (CT-BP-6-NEG N11, see Deviations) and the resulting B-REQUIRED-CHECKS; postcheck passed. Cycle 1 |
| c1 | unit lint contracts provenance root migrations ops rename off | `eac2b61` (pre-rebase, BASE `8488b6a`) | **exit 0**; 82 PASS, 3 INFO, 0 FAIL/UNMET; B-REQUIRED-CHECKS 81/81; postcheck passed; no leftover resource. B-BUSINESS-PROFILE-CATALOG 29/29, B-BUSINESS-PROFILE-MIGRATION 6/6 + restart, B-BUSINESS-PROFILE-IT 35/35 (1162 records catalogued, no canary, planted control detected); B-TENANCY-IT 35/35 at version 4; B-TENANCY-HTTP 56 probes; B-IDENTITY-IT 57; B-BROWSER-ROOT and B-BROWSER-OFF 141 tests; B-TOGGLE-OFF version 4; B-SCOPE 26 paths; B-SECRET-CANARY 194 secrets. Retest of cycle 1. Started before the owner sequencing amendment: historical preliminary evidence |
| F-1 | `run.sh --base d4bb478 … migrations` in a scratch clone of the rebased `08d279f` with IT-BP-LINK-3 inverted (`ok.Load() == 1 || exists.Load() != 31`), after the pre-batch check | `08d279f` + 1-line test edit | **harness exit 1 as required**: B-BUSINESS-PROFILE-IT FAIL (`--- FAIL: TestBusinessProfileModel/IT-BP-LINK-3…`; logcheck ok), B-REQUIRED-CHECKS FAIL on exactly that ID; the other 13 checks PASS (B-SCOPE 26 paths since `d4bb478`); no leftover resource; scratch clone removed (`runs/s5-f1/`). B-CI-CONTROLS proves any FAIL fails `validate.sh` and the gating job |
| s7-final | all unit (`bin/g006-vw2.sh`, reconciled snapshot `3a73e3b8…`, `--base d4bb478`), after the pre-batch check | `08d279f` (tree `d182afe…`) | **exit 0**; 85 PASS, 3 INFO, 0 FAIL/UNMET; B-REQUIRED-CHECKS 84/84 (lint controls provenance contracts root migrations ops rename off unit, final); wrapper pre/post passed; no leftover resource. B-BUSINESS-PROFILE-CATALOG 29/29, -MIGRATION 6/6 + restart, -IT 35/35 (logcheck, planted control); B-TENANCY-IT 35/35 at version 4; B-TENANCY-HTTP 56 probes; B-CI-CONTROLS 282 (incl. staging's `F-vestive-link`); B-CI-CONFIG; B-PROTECTED 341 files + 75 expected hashes; B-IDENTITY-IT 57; B-BROWSER-ROOT and B-BROWSER-OFF 141 tests; B-TOGGLE-OFF version 4; B-UNIT-CHECKS and B-CHECKS-ROOT; B-SCOPE 26 paths since `d4bb478`; B-SECRET-CANARY 194 secrets |

Development-only checks (not harness evidence): `go vet` (with and without the integration tag), `gofmt` and the core
and `logrecord` unit tests in the pinned `golang:1.27.1-alpine` image, on scratch copies before each apply and on the
checkout before each commit; three disposable scratch PostgreSQL checks of the 00004 DDL (website CHECK, view DML).

### Coverage of the final tree

The final candidate is `08d279f` (the five G006 commits on `d4bb478`) plus the commit adding this Result, one
Markdown file under `management/` that no check executes (B-SCOPE allows it). `s7-final` ran every required full suite
and `unit` on exactly `08d279f` after the rebase, so no evidence is reused for the final tree. `c1` (pre-rebase) is
historical preliminary evidence only; F-1 ran on `08d279f`.

## Acceptance criteria (Spec AC1–AC12)

Local runs are in [Validation](#validation-commands-and-outcomes); hosted state in
[Hosted MR pipeline](#hosted-mr-pipeline).

| AC | Outcome | Evidence |
|---|---|---|
| AC1 Organization-owned profile identity | Met | CT-BP-4/8 (no `group_id`, exact columns, `business_profiles_organization_id_key` composite target); IT-BP-ID-1..5 (database-generated nonzero UUIDs, updates keep identity, runtime cannot supply or change keys, owner key changes hit the triggers, unknown Organization and zero IDs); U-BOUNDARY (no move/reparent/delete/merge method) |
| AC2 context contract | Met | U-CTX, U-URL, U-PARITY; IT-BP-CTX-1 (service field matrix), IT-BP-CTX-2 (runtime CHECK negatives incl. DEL/C1/space/non-ASCII websites, plus Go-valid values inserted by the runtime role without hitting a CHECK; rows unchanged), IT-BP-CTX-3 (g1/g2 read one committed update through the same UUID; unrelated rows unchanged), IT-BP-CTX-4 (inert values; an `httptest` server receives zero requests) |
| AC3 shared profile, at most one per Group | Met | IT-BP-LINK-1 (two Groups share p1, a third uses p2), IT-BP-LINK-2 (second profile refused with zero INSERTs, runtime SQL 23505), IT-BP-LINK-3/4/5 (32 concurrent links × 20 rounds: exactly one; deterministic interleaving; unlink versus relink), CT-BP-4 (Group-unique; profile side not unique); F-1 inverts LINK-3 and fails |
| AC4 same-Organization enforcement | Met | IT-BP-XORG-1 (ordered prechecks, foreign = unknown byte-for-byte, incl. the unlink-foreign-profile pair; zero writes), IT-BP-XORG-2 (27-combination runtime INSERT matrix: only (A,A,A) and (B,B,B)), IT-BP-XORG-3 (no retarget/reparent), IT-BP-XORG-4; CT-BP-4 composite FKs |
| AC5 unlink, atomicity, references | Met | IT-BP-UNLINK-1/2; IT-BP-ATOM-1 (rejections), IT-BP-ATOM-2 (commit failure, statement failure, closed pool and mid-transaction cancellation on create/update/link/unlink and the G005 Group grant/change/remove at version 4: `unavailable`, rows unchanged); IT-BP-REF-1/2 (RESTRICT, races both orders); CT-BP-4 `confdeltype='r'` |
| AC6 derived access facts, no profile authority | Met | CT-BP-5 (exact view, SELECT-only grant); CT-BP-6 and CT-BP-6-NEG N10–N16 (N11 see D1); CT-BP-8 (no profile scope, owner column or role row); IT-BP-ACC-1..3; IT-BP-ACC-6 (projection writes fail; D2) |
| AC7 independence of facts and authority | Met | IT-BP-INDEP-1/2 and IT-BP-ACC-2..5 (row snapshots around each mutation; last path removal; one of two paths survives; member versus owner/admin; no sibling-Group or Organization authority); no copied store (CT-BP-6 A1, N11/N13/N16) |
| AC8 no public surface or future-Goal implementation | Met | B-TENANCY-HTTP (56 probes 404), B-MCP-SOURCE, B-WEB-BOUNDARY, B-ENV-EXAMPLE, CT-BP-7/8 (no policy, RLS, onboarding, completion or billing object; no singleton on profiles), TENANCY.md hand-offs (G007/G008/G018/G019, future BillingProfiles 1:N), B-SCOPE |
| AC9 migration safety and persistence | Met | B-BUSINESS-PROFILE-MIGRATION (MG-BP-1..7 incl. the restart), B-BUSINESS-PROFILE-CATALOG (CT-BP-1..5, CT-BP-9), reconciled B-MIGRATE-*, B-IDENTITY-CATALOG, B-PERSIST, B-DBRENAME, B-TOGGLE-OFF (version 4); B-MIGRATE-SET pins 00001–00003; identity/G005 preservation in MG-BP-3/5 |
| AC10 inherited contracts and redaction | Met locally (`s7-final` on the final tree); hosted full pending the staging gate | Every inherited ID in `s7-final` (B-TENANCY-* incl. the G005 model at version 4, B-IDENTITY-*, browser, first-paint, token, boundary, variant, dependency, provenance); IT-BP-LOG-1 with canaries in all five context fields, `logcheck validate`/`scan` and the planted control; B-SECRET-CANARY; B-CONTRACT-TOOLS; controls rerun on the final tree after the shared-harness edits |
| AC11 container and hosted CI gate the model | Local part met; MR `fast` per [Hosted MR pipeline](#hosted-mr-pipeline); hosted full **pending** (staging gate after conditional integration, owner status rule; a cancellation is recorded as such and is not test evidence) | Wrapper runs on the actual root with pre/post checks; `s7-final` (`all unit`, off variant included, no required skip) on the rebased candidate; F-1 (harness exit 1, B-BUSINESS-PROFILE-IT FAIL); B-CI-CONTROLS |
| AC12 provenance and bounded Result | Met, pending Codex review | E4 binding and launcher digests on every run; wrapper pre/post; NUL-safe status; allow-list; inventories; rebase reconciliation records; failure history; B5 accounting; no inheritance, RLS, onboarding, billing or product claim |

## Deviations (architecture deviations: none)

- **D1 — CT-BP-6-NEG N11 expected set.** Plan r2 §5.1 expects exactly `{stored_role_outside_authority}` for N11
  (`ALTER TABLE app.group_business_profiles ADD COLUMN role text`). The Plan's own grants (§3.3) give the runtime role
  table-level DELETE on associations, so the accepted G005 detector's A3 rule also reports `writable_role_column`,
  exactly as for G005's N8. The expected set is now both codes; the detector is unchanged, and the match is still exact.
  Found by s2-a (cycle 1).
- **D2 — IT-BP-ACC-6 SQLSTATE.** Plan r2 §5.2 expects 42501 for DML through `business_profile_access_paths`. PostgreSQL's
  rewriter rejects writes to a non-updatable join view with 55000 before any privilege check, so IT-BP-ACC-6 asserts
  55000; the privilege side is proven by CT-BP-3 (`has_table_privilege` matrix) and CT-BP-6-NEG N14 (a granted INSERT
  is detected as `writable_projection`).
- **D3 — U-BOUNDARY literal exemption.** The SQL-literal check flags `"business_profile.update"` (the word `update`).
  `logOperationNames` exempts exactly that one log operation name; no other literal is exempt.
- **D4 — missed RC-8 site.** `runner_test.go` `TestProviderAcceptsProtectedStore` also pinned the migration sources; it
  now pins the four-file set (cycle 1).
- **D5 — function EXECUTE evidence.** Plan r2 §3.3 credits the postcondition and CT-BP-3 for owner-only EXECUTE on the
  trigger functions; the actual check is CT-BP-2 (default privileges from `infra/postgres/init/10-roles.sh`), as in G005.
- No service-specific governance or skill exists under `core/`; none was needed.

## Security and privacy self-review

No secret, token or WorkOS value was printed, copied or used; the WorkOS staging env file was not read. No live WorkOS
action, protected-ref update, project/provider/runner/schedule setting change, pipeline start, cancellation or merge.
The model adds no authority path: no route, no binary linkage, no BusinessProfile role, scope, grant or owner column;
runtime has column-limited writes only; keys are immutable by grants and triggers; every statement names its
Organization.

A read-only security/SQL review subagent reviewed the S3 tree (`review/s3-security-review.md`): **no high or medium
findings**. Verified: tenant isolation (every statement Organization-scoped; composite FKs; identical foreign/unknown
errors), exact privileges and ACL postcondition, the `security_invoker` view, parameterized SQL and race outcomes, the
input contracts (Go never looser than needed, SQL never stricter), logging without IDs or values, additive migration.
Its five low findings were fixed before the S3 commit (gofmt; escaped format characters; canaries in all five fields
with unlogged reads and created UUIDs; the unlink foreign/unknown pair and cancellation on all seven mutations; database
positives plus DEL/C1/space/non-ASCII negatives). The reviewed production bytes equal the committed ones except gofmt
whitespace in `business_profile.go`, `errors.go` and `store.go` (whitespace-insensitive equality checked).
For G007/G018: valid websites may contain quotes, angle brackets and backticks and may be `http://`, IP literals or
localhost; context fields accept bidi and format characters as data. Renderers must escape, never fetch, and consider
spoofing in profile pickers.

## Rebase onto `origin/staging` and binding reconciliation

Under the owner sequencing amendment the rebase preceded F-1 and the final runs (Plan r2 listed S5 before S6).
After `c1`, the fetch gave `origin/staging` = `d4bb47882cf47fb48043b041d2fb7834441e9547`. The tested branch `eac2b61`
was rebased onto it **without conflict**: `82988f7` (S1), `727aa29` (S2 WIP), `43c606d` (S2), `7533cc8` (S3 + cycle 1),
`08d279f` (S4), tree `d182afe26fa557065cc9d9dc3de9ad88991c21e0`; `d4bb478` is an ancestor of the source.

- **G006 content unchanged** (`rebase-reconciliation-d4bb478.txt`): the same 26 G006 paths change against the new base;
  25 are byte-identical to the tested `eac2b61` blobs, and in the one shared file (`infra/validation/run.sh`) the G006
  hunk bodies are identical (hash of the `-U0` `+/-` lines `a59027ef83cd9fdb` before and after). The target's own
  `run.sh` change (`VESTIVE_ALIAS=.agents/skills/vestive-design`, `e23799b`) is kept.
- **Root governance, the Vestive discovery link and the production-domain documentation** reach the branch only
  through this rebase onto the target, as integrated target commits; nothing was copied into the checkout by hand.
- **Reconciled snapshot** (`reconciled-d4bb478/approved-inputs.sha256`, SHA-256
  `3a73e3b8ef5820ebfce3c7ac3d7aa2e3fe98b12ac7536ae1db121ff4aa2561f5`): the same 74 paths, count, `path_re`, Plan
  (`3bd79dc5…`, unchanged) and scope (`85497a2d…`, unchanged); every entry equals the new base's blob. The frozen bundle
  and Plan bytes are not modified. Nine inputs changed, all through integrated target commits:

| Input | Old SHA-256 | New SHA-256 | Changed by |
|---|---|---|---|
| `AGENTS.md` | `00d8a79b2ca3d0c56f47e3fa56c1419e908414b18cde8253d30ed7ebc0b97aab` | `2549bad60ab94e7589a08c74b0bad81aa271322d69ddeb34982e4f9c05493fc0` | `d4bb478`, `e23799b`, `f656741` |
| `CLAUDE.md` | `8490b1d35c56f4dec047638bd0e64d71c126e712756c01b9cd460cbccbcce28b` | `cf4de436dfa971ddf0bbb03c3fd010650c3c3e4835b9341c9780167f5ce02f9e` | `d4bb478` |
| `docs/ARCHITECTURE.md` | `62e1c53c85b01e37a8046b4ac20e61799b1d3c8dd1551f164936a03ca18cb024` | `77dfbd768497a769314d74b263caba39b63a27146fbf71d05d9f0f38ab5c57b0` | `e234504` |
| `docs/agents/GITLAB-BRANCH-WORKFLOW.md` | `093690f7ed59bb2acbb14a7e06a419900bca21e27130d7f5192d269b0d66727c` | `bf0732cfd499ea42553f3f4cfdfbadbc71559d801c24592828ddd04f5f626a27` | `d4bb478` |
| `docs/foundation/OPERATIONS.md` | `35bd05614e3b53a05dc5705ec9764eea2aebcc035399c90e0ac01159bfb1cf4f` | `266d9fe35a7c18a8aedd7db8a6904d51ba8012811d006ea67af589bc35d6bae4` | `e23799b` |
| `infra/ci/profile-paths.tsv` | `6c2c2e0a4736536950bf0db509430d1beb9497fa3c9a714e1d3d80a932b10fb3` | `8cc14314bc65153eed11ef2944481e186861a44c299a8509c0c1b0657c5d3920` | `e23799b` |
| `infra/ci/validate-controls.sh` | `0fb81a84eb39ec851d60cae08fef78f61a25046dc030f851e372bcd4c3e6676d` | `b8fad7d1ea6029ae1b471b398e6614d21e2b9f036cce92de749ed1c62d2fdb4d` | `e23799b` |
| `management/v0/PROGRESS.md` | `09638966d41f4f337dd3e0695003e435071f0505fba595812d26de3461733f4d` | `e87a9b077aaf43fc938b374f06b0ddad0bd9477ed33208dc2a198eb75dad3486` | `6ba3720`, `e37f954`, `9db19b5`, `056a2fe` |
| `management/v0/ROADMAP.md` | `3ec51245a5e71421ae25feb7367abe668b91f00181f1900454ec19c08d3a5a48` | `e8633db5f85400c46a8d81589919b1c1ed25379d7f294bdb4e36df6e1133496c` | `e234504` |

- **Prechecks:** the reconciled wrapper precheck passes (`precheck passed (74 inputs, snapshot 3a73e3b8…)`, exit 0);
  the original snapshot now refuses with exit 11 on exactly those nine inputs (`reconciled-d4bb478/old-snapshot-check.txt`).
- Runs after the rebase use `bin/g006-vw2.sh`: `--base d4bb478` and the reconciled snapshot, after a recorded fetch and
  ancestry check (`prebatch.tsv`); the launcher refuses if the target has advanced beyond `d4bb478`.

## Execution budget

B5, checkpoint after cycle 3. **Cycles consumed: 1 of 5** (parent 1; security reviewer 0, read-only). **Checkpoint
(after cycle 3): not reached.** No repeated failure, no thrashing, no semantic change to accepted G004/G005 behaviour
beyond the planned mechanical reconciliation.

Not counted: s2-a's expected failure is the cycle-1 trigger itself; F-1's exit 1 is required; development-only scratch
compiles and scratch PostgreSQL checks before the first apply were part of the initial implementation.

### Cycle log

| Cycle | Trigger | Fix | Retest |
|---|---|---|---|
| 1 | s2-a: B-UNIT-CHECKS (runner source pin), B-BUSINESS-PROFILE-CATALOG (N11) | D4 pin; D1 expected set; together with S3 and S4 | c1 exit 0 (82 PASS) |

## Resource cleanup

s2-a's global inventory diff reported one leftover anonymous volume. Cause: two development scratch scripts
(`bin/g006-scratch-pg.sh`, `bin/g006-scratch-viewdml.sh`) removed their PostgreSQL containers with `docker rm -f`
without `-v`. Three such volumes were removed after positive ownership evidence (creation times matching the scripts;
WAL segments naming `group_business_profiles_group_key`, which exists only in 00004); the scripts now use `docker rm -f
-v` (`cleanup/scratch-volumes-20261009.txt`). No other resource was touched. Every later harness run (`c1`, F-1,
`s7-final`) left no resource (`inventory-leftover.txt` empty); the F-1 scratch clone was removed. Unrelated
`payments-assistant-*`, `needknow-*` and `outreacher-*` resources were never touched.

## Residuals and follow-ups (no new scope)

- G007 classifies the raw Group roles in the access paths into capabilities, with Organization authority.
- G008 adds RLS on both tables (the view inherits it through `security_invoker`).
- G018 owns exactly-one-at-completion and create-or-select onboarding; G019 the reassignment/removal lifecycle. Any
  exposure of `UnlinkGroupBusinessProfile` must be guarded so a completed Group never silently loses its profile.
- `docs/foundation/OPERATIONS.md` profile counts remain stale (Q-r2-11; out of G006 scope).

## Hosted MR pipeline

This Result is committed before the push. The commit adding it is the pushed source of the MR
`g006-business-profile-associations` → `staging` (source-branch deletion; X1). Under the solo-operator policy:
- the MR's automatic `fast` job is the hosted MR evidence;
- its `gate` is manual/optional and is **not** started;
- the required full hosted evidence is the automatic staging gate after an owner-authorized conditional integration
  (X2, not authorized), judged by the owner status rule.

The MR URL, pipeline IDs and the `fast` outcome are recorded in a local follow-up commit that is **not** pushed, so the
MR diff stays this source; the parent's hand-back reports them. No pipeline is started manually, retried or canceled.

## Reasoning effort and delegation

Implementation, integration and all validation by the parent session. One read-only security/SQL review subagent
(general-purpose; no edits). G051 ran in parallel in its own relocated checkout through its existing child, sharing
only the harness lock.

## Suggested follow-up (not authorized scope)

None beyond the residuals above.
