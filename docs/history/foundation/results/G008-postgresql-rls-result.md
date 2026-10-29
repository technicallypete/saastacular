> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G008 — PostgreSQL RLS and least-privilege roles: Result

## Status

**Submitted for Codex review; not accepted.** Only Codex accepts and updates PROGRESS.
- **Implementation:** Plan r3 (`management/v0/plans/G008-postgresql-rls-plan.md`, 0887a6fb…), within the scope
  allow-list (798fafee…, 15 expressions).
- **Final qualification:**
  - **Local:** b34-final (`all unit`) FAILED with 90 PASS and 3 FAIL. b37 reran the failed and drift-affected suites
    (58 PASS, B-CI-CONTROLS FAIL). b38 reran controls alone (B-CI-CONTROLS FAIL).
  - **Owner authorization:** the owner then authorized the MR-hosted full gate (staging 13503d4) to verify the CI
    optimization together with G008.
  - **First MR !23 gate:** job 17159164435, pipeline 2945136878, on 95113af. **FAILED**, kept as failed history: every
    G008 ID passed, and B-CI-CONTROLS failed on 32 controls of the untested 3a50a7e change (§6.5). It ran before the
    S5 corrections.
  - **Second MR !23 gate: PASSED.** Job 17160250730, pipeline 2945315031, on the final candidate f0ed177 (base
    e80077d). 64 PASS rows, no FAIL or UNMET. B-CI-CONTROLS passed 441/441, verifying Codex's CI change and its
    correction. B-REQUIRED-CHECKS: 63 IDs, each PASS exactly once (§6.5).
  - **Still required:** final acceptance needs Codex review, then a conditional integration and the staging gate under
    the owner status rule.
- **Forced F-1 check:** done on 95113af, before the S5 corrections, which touch only provisioning wording and the harness's rls_pg password handling, not the detectors. The expected FAILs were observed, plus B-RLS-MIGRATION (stricter than the Plan required); see §6.2.
- **S5 read-only security review: NO-BLOCKING-FINDINGS** (two LOW findings, both corrected; §6.6).
- **Final candidate:** code 02f3bc7008370acddd62ca9eaa1801533ba533d9 on staging e80077d2763d245add1ff2674c5eac524d47ddab (the S5 corrections, rebased onto Codex's CI correction). The hosted gate passed on f0ed177 (code plus the first Result commit). This update records that gate and is the only later change (Result only, documentation). The handback states the exact final commit.
- **MR:** draft MR !23 to staging, with source-branch deletion, was opened by the coordinator. Merging needs Codex
  acceptance and separate owner authorization. Staging qualification follows any authorized merge.
- **Budget:** B7, attempt 1 (in progress since GO, never reset). The checkpoint after attempt 4 was not reached.
- **PG-002:** the owner accepted the characterized query-cost risk ("Option 1 is good.",
  G008-performance-owner-disposition.md ec85cdc0…). PG-002 stays open; see §8.

## 1. Worker, binding and authority

- **Worker:**
  - G008 single implementation owner (task G008, role impl), agent `agent-<id>`, effort high;
  - checkout `.claude/worktrees/agent-<id>`, branch `g008-postgresql-rls`;
  - dispatch and S0 records: `<EVID>/dispatch-record-impl.md` and `<EVID>/s0/S0-results.md`;
  - `<EVID>` = `<home>/.local/state/platform-evidence/g008-impl-c20ada0` (private local evidence; not
    published).
- **S0 (2026-10-13):**
  - GO file seen at 21:56:12Z;
  - branch created at BASE c20ada0d2ec912a8380e90efc83c1576139cdf4b;
  - check-only `precheck passed (90 inputs, snapshot aa474dd5…)`;
  - mounts at 5%; ports and lock free; no harness resources;
  - GitLab minutes 6277 used of 10000 + 3973; not a resource stop.
- **Authority chain** (all under `management/v0/results/` unless noted):

| Record | SHA-256 | Effect |
|---|---|---|
| G008-execution-binding-codex-confirmation.md | ad704d83… | authorized binding c20ada0 (90 inputs, snapshot aa474dd5…, receipt d8cf9059…) |
| G008-execution-amendment-01.md | ccd03824… | C1 scope + binding (15 expressions, 88 inputs); C2 oracle delegate; C3 `inTx`/`inContextTx`; C4 migration `SET LOCAL ROLE` |
| G008-oracle-a1-codex-review.md | bba05bb0… | oracle r0 (8523cc47…) superseded; `no_update` semantic outcome |
| G008-oracle-d1-d2-codex-review.md + specs/G008-context-validity-clarification.md | 6cee1389… / 4df154eb… | D-1/D-2 ruling; amendment-02 binding (89 inputs) |
| G008-erls-escalation-03-codex-review.md | 4aeb1aba… | E-RLS measurement correction accepted; 44 alerts kept unresolved |
| G008-pool-discard-deviation-codex-review.md | 1fd1cf0b… | pool-discard deviation ACCEPTED (bounded) |
| G008-performance-owner-disposition.md | ec85cdc0… | owner: continue under the characterized risk |

## 2. Binding history and E3 reconciliations

Every receipt is under `<EVID>/`. Each one re-derives only the changed bound entries, keeps the authorized or amended
snapshot as history, and was followed by a passing check-only. The path-re is amendment-02's (a9e8c49e…) from
amendment-02 on; the scope is 798fafee… from amendment-01 on.

| Step | Base | Snapshot (count) | Changed entries | Note |
|---|---|---|---|---|
| authorized | c20ada0 | aa474dd5… (90) | — | receipt d8cf9059… |
| reconcile-c56f46c | c56f46c | 09e6c539… | management/docs | |
| reconcile-3df5a37 | 3df5a37 | 4f3d1cb2… (90) | ARCHITECTURE (+13, §40, future-G009 input), PROGRESS | |
| amended-01 (`g008-execution-binding-amended-01-f375463`) | f375463 | 5854f1cf… (88) | runner_test and validate-controls leave the snapshot; Goal | path-re e48ad3b3…; verified `amend-01/verify.log` PASS |
| reconcile-91faf2c | 91faf2c | 6253c57c… (88) | management only | |
| reconcile-981c67d | 981c67d | f009d558… (88) | PROGRESS, ROADMAP | b12 process note (§9) |
| amended-02 (`…-amended-02-cebc16a`) | cebc16a | 4b9c4883… (89) | PROGRESS, + clarification | path-re a9e8c49e…; `amend-02/verify.log` PASS |
| reconcile-682d799 | 682d799 | 5d79c65e… (89) | GITLAB-BRANCH-WORKFLOW (6db5615c…), PROGRESS | evidence-closeout policy |
| reconcile-df68eba | df68eba | ff739067… (89) | PROGRESS | |
| reconcile-de56f3f | de56f3f | d396f566… (89) | PROGRESS | |
| reconcile-bd9bb25 | bd9bb25 | 852dac51… (89) | PERFORMANCE (PG-002 corrected observations), PROGRESS | |
| reconcile-19a165e | 19a165e | e7398abe… (89) | PROGRESS | pool-deviation ACCEPT |
| reconcile-f081e48 | f081e48 | 9e6e0ba0… (89) | AGENTS (bounded cleanup), PROGRESS | OPERATIONS also changed on staging (not bound; no overlap with G008's two rows) |
| reconcile-cef0ad6 | cef0ad6 | 5e0d900a… (89) | PERFORMANCE, PROGRESS, owner disposition | first re-derivation omitted the disposition entry; check-only refused it (FAIL 11); nothing ran |
| reconcile-37d9190 | 37d9190 | 380e820f… (89) | AGENTS (Git-first provenance), ARCHITECTURE (§21.1 audit schema, ADR-006), DECISIONS, PROGRESS, ROADMAP | one pass for 44820ff, eea8f90 and 37d9190; ADR-006 states no change to active G008 |
| reconcile-0dc83a3 | 0dc83a3 | fd97610f… (89) | PROGRESS | |
| reconcile-dffe9cd | dffe9cd | 0f6f4d56229423790c32aefd6a2748cb659bffe61c909e6d8a8f40e123e57b79 (89) | AGENTS, CLAUDE, GITLAB-BRANCH-WORKFLOW, infra/ci/lib.sh, select-profile.sh, validate.sh, PROGRESS (469a53f3…) | owner-authorized, untested CI optimization 3a50a7e (ci-controls split and `ci_controls_required` selector) plus 5e9f454/13503d4/dffe9cd; one rebase conflict (§6.3) |
| reconcile-e80077d | e80077d | d50e0f976ce8104681201f58657a7e345990d67efb589792f5f50d3583f1740e (89) | PROGRESS (66061830…) | Codex CI correction e80077d (VB1/VB3b/VB5 supervision, legacy selector fixtures); rebase clean; G008 delta in validate-controls.sh = the four count lines; not charged to G008 |

Pushes went only to `g008-postgresql-rls`, with `--force-with-lease` against the exact previous remote tip after
each rebase: c2fe613 → a53993e → ec22958 → 3741ad2 → a74b3ce → 6e5345c → 6b11649 → 07bfa78 → 7437bfc → fd6979c (unpushed) → 417d2de → ea08c05 → 95113af (MR !23 opened by the coordinator; first hosted gate) → ed7225e (S5 corrections) → the final candidate (this Result on 02f3bc7).
There was no MR, no staging/main push and no pipeline start or cancel.

## 3. Implementation summary and files

The changed paths (43 against staging) are all inside the 15-expression allow-list (B-SCOPE PASS).

- **Migration `core/migrations/00006_row_level_security.sql`** (fae1813b…). One goose transaction as the owner:
  - preconditions: owner, role attributes, memberships, helpers holding nothing, exact version-5 grants;
  - four context parsers `app.ctx_user/org/group/login_binding()`: malformed UUID 22P02; zero UUID, bad binding and
    illegal shape 22023;
  - two policy readers (`rls_org_role()`, `rls_group_role(uuid)`) owned by NOLOGIN NOINHERIT `<db>_policy_reader`;
  - three authentication readers E1 `auth_resolve_session`, E2 `auth_principal_subject` and E3 `auth_logout_ref`,
    owned by `<db>_auth_reader`. Helper-owned grants run through migration-only `SET LOCAL ROLE` (C4), with the owner
    restored explicitly;
  - the narrowed runtime grants; ENABLE + FORCE on nine tables; 26 policies (21 `TO runtime, owner` plus 5 helper
    policies);
  - postconditions: RLS flags, policy inventory, ACL strings, function allowlist and schema ACL;
  - down: drops policies, readers and parsers, restores the exact version-5 grants, and leaves the helpers holding
    nothing.
- **Context statement:**
  - one guarded `set_config` statement for all four fields, byte-identical in `core/internal/tenancy/store.go` and
    `core/internal/identity/store.go`;
  - `tenancy.InitContext`; `CodeContextPresent`;
  - facade: lifecycle-only `inTx` plus mandatory `inContextTx` at every production path (C3);
  - a poisoned connection is closed while held.
- **Identity store (`core/internal/identity/store.go`, `logout.go`):**
  - EMPTY, IDENTITY and LOGIN initializers;
  - E1, E2 and E3 for lookups;
  - F3 login start as one guarded transaction;
  - core-generated v4 user UUIDs;
  - the first-sign-in race path under RLS;
  - a zero UUID in Disable/Enable returns `invalid_request`.
- **Provisioning:**
  - `infra/postgres/init/10-roles.sh` adds the helpers and the owner's SET-only memberships;
  - `infra/postgres/provision-roles.sql` is portable, idempotent and uses `\getenv` passwords. It refuses before any
    change. On the non-superuser admin path it grants itself SET on the owner temporarily (revoked before commit) and
    makes the app-schema grants and owner default privileges under `SET LOCAL ROLE` owner.
- **Harness:**
  - `infra/validation/run.sh`: B-RLS-CATALOG/MIGRATION/IT/ORDERING, PROV-1..6, version-6 texts and the superuser test
    env;
  - provisioning instances are removed explicitly with `-v`, and their anonymous volumes are recorded in the run
    registry at creation (bounded-cleanup rule);
  - the render register is 26;
  - `required-checks.tsv`: 99 rows, full 92; `render-customizations.tsv`: the 10-roles.sh row;
  - amendment-01: `core/internal/migrate/runner_test.go` (six migrations), `infra/ci/validate-controls.sh`
    (VA7 64, VU1 33/64/93, J8 64), `docs/foundation/OPERATIONS.md` (92; 33/64/93).
- **Docs:**
  - new `docs/foundation/DATABASE-ISOLATION.md`;
  - annotations in MIGRATIONS, FOLLOW-UP (STAGE-RLS checked), AUTHORIZATION, TENANCY and IDENTITY.

Changed files at the code candidate 02f3bc7008370acddd62ca9eaa1801533ba533d9, against staging e80077d (SHA-256; this Result is the only later change):

| Path | SHA-256 |
|---|---|
| `core/internal/identity/integration_test.go` | 05fc8691c5e6ca6afa3e8f6be8f40d74cd4c861300f81f34697fa3d9b72422d9 |
| `core/internal/identity/logout.go` | c89fa42645db4c36701d4d3b0476e9a9c7d4568879a914549d883460312eb076 |
| `core/internal/identity/logout_test.go` | 387c59392ca39e4da6200cde8160ec4202e76d2f84a450c8ba713e4730a715b0 |
| `core/internal/identity/manual_qualification_test.go` | 15ac9e3186812ff3c6a3f95e132823da49238ed00229cc104dcb6ab1e484cf4a |
| `core/internal/identity/rls_pool_integration_test.go` | c1e25a8b998765850e50c1c3f6a72dc4c94cc322e0e1298528af2b90171e0e56 |
| `core/internal/identity/service_test.go` | b289d5a8bb447f940bdae1e1f5ab8f78b34477e998bf1fbcb4a54934dfd1a3a8 |
| `core/internal/identity/store.go` | 626b04da563a88c99e4bad9574d671a1f2db2af2ac2ab5bdf3fbd56d83ea9241 |
| `core/internal/identity/store_test.go` | cbfdb52676aad2ad633961ca18f939d61528dd0b42edbaaa8ad8d241f74ddfde |
| `core/internal/migrate/integration_test.go` | 6af4d1c60f72d332f175a9cf8dd028364ada60ebe9467648beda3e4cabddb974 |
| `core/internal/migrate/policy_integration_test.go` | 8a7503a1eea7e256074fbfcbf253d96ddd9e8528b37ae0aa0949653831ffd533 |
| `core/internal/migrate/rls_integration_test.go` | 1eab93ee8610f55a1238812835d57023b014382bda7f648b723a5ecad1449da6 |
| `core/internal/migrate/runner_test.go` | 7ea53e19fc4186ecc1167daed2abe8e0709cf11fb27116e4ce96c35c18fa4a2f |
| `core/internal/tenancy/access/boundary_test.go` | 105d2004b267313c3457bf96ae3a105dfbfc65225c47a988b61f6e4c1c856260 |
| `core/internal/tenancy/access/context.go` | 9280f59d37e515369b30d3a03ea181e6c8e0ac227fbe4a34e94055c51b4611f2 |
| `core/internal/tenancy/access/facade.go` | b33177c06361f3d7a4cb81927e99d31d6fb42365fab46225c544d862510fe0dd |
| `core/internal/tenancy/access/footprint_integration_test.go` | de24ca754e026bcc7032ad51550e12269c090c5b92609f9ae583609a58c8a1d9 |
| `core/internal/tenancy/access/integration_test.go` | 9a04dae878f97168b9681d41544be3065558a288702e022e9370c01d1033c087 |
| `core/internal/tenancy/access/ordering_integration_test.go` | dd308a15fb7507154487bda8b9f98d05a5011418f41131183e11a8c3b7a4a55c |
| `core/internal/tenancy/access/rls_expected_test.go` | 4d76e68a2c0af2ea24040ef3f42dbe4db4dacbfdf47b9326f9832d3ba095998d |
| `core/internal/tenancy/access/rls_integration_test.go` | 7bc0150d23aaddcbd905c63354942149fd2c694e6e9acf169a1d0530883314b4 |
| `core/internal/tenancy/access/rls_ordering_integration_test.go` | 0ac3201c35ca43497be2f1f1f4ccf3e3fd9d8a3493e803d50c07e2dc5e5b8ccc |
| `core/internal/tenancy/access/tx.go` | 15f56c701e5fd42ba9783d9dbb70fddd0d10bae77a988eb6bb4b6ffb3731aef6 |
| `core/internal/tenancy/boundary_test.go` | 6469557201e0a2bf0b34ac25458359958d8aecee02d6d5b554111fe40e3e552a |
| `core/internal/tenancy/errors.go` | 8f4fef0e2e7d6048a894e40e2fc55f8681a4d82db51f3838df6bde967aeebbc5 |
| `core/internal/tenancy/facts.go` | e2b640b6db112817e4f70959b4f4afa609435b51a27a14485dceefbd8418e845 |
| `core/internal/tenancy/integration_test.go` | 3e21a598b49e0824a90fa0ccddb7478cc12032a26c01f6b6f93343101520efda |
| `core/internal/tenancy/store.go` | 00c1e38d692a4cfb671eaefc4507ef446f08881d75d0b0980afc603c67026dfe |
| `core/migrations/00006_row_level_security.sql` | fae1813b6b747a4d385eb16a4b900c92290b3eb82331ca85bb6d41a44c7384ed |
| `core/migrations/migrations_test.go` | 664238392fe039ede369927bf2029d5ed14fdf9a8d1450f121e655d9f632b72f |
| `docs/foundation/AUTHORIZATION.md` | b5aaff584493908b44b1ed41bd3fe73477430e7b8c74753485c5e2ab0877aa72 |
| `docs/foundation/DATABASE-ISOLATION.md` | d2455ed5e99068ab1dcc8517ff215bfe25b8a7157ba346c98b5d7bca82c94062 |
| `docs/foundation/FOLLOW-UP.md` | bc5a07bd6f4349baf976dd7c1b06a99e71cf273da2a7ec43e71ec7a7a3b7ebac |
| `docs/foundation/IDENTITY.md` | 620966750ab1989a7a513a3bb47c0e9a9ec20f4457fb8ec61bc62a75dff6285b |
| `docs/foundation/MIGRATIONS.md` | cfda0dd99eceaf30e50f36c0663b3d159dd73bc79198c3915bf5ef8ffa7c017e |
| `docs/foundation/OPERATIONS.md` | 86067ab848073d3764264c0c7c950502f213e8f68397723d1f12e818b5c00575 |
| `docs/foundation/TENANCY.md` | cde06f213c4812c86e4ddaaaf71605e06497dfd0dcc0441c30ecc2c2d99af44a |
| `infra/ci/validate-controls.sh` | bc6ed0a5ec29e28c36b7f4813441f1316c6822a8eac2aa07887302e49ccb163f |
| `infra/postgres/init/10-roles.sh` | 7124374086eda3c023e0eb6190f1eef668432716ee75f33795224fdf9dc65f94 |
| `infra/postgres/provision-roles.sql` | c95cb68ce45200918e69f0e316bace87b946971aeca28272096f22515a49a08c |
| `infra/validation/render-customizations.tsv` | 491e2647867ac42709b9fc21a0de2079b7387d42084c4915bc9c82059febdc33 |
| `infra/validation/required-checks.tsv` | 7f938d3bd799c04db708dcfc35aab1f4f0d9028ff8780d9aaa555fdbddda7be5 |
| `infra/validation/run.sh` | 6469e3986616257dab2d5abd59d4cc74f76e7451e9a6d0037d36d28d1fc51e85 |
| `management/v0/results/G008-scope-allowlist.txt` | 798fafee7f65d8705875c3491b4644ae858655e9850d8cc6c2cd0a98964260f7 |


## 4. Migrations, contracts and interfaces

- **Schema/migration:** 00006 as above. No relation, column, constraint, index or data change. It takes ACCESS
  EXCLUSIVE on the nine tables for its transaction. Deploy order: provisioning, then migrate plus G008 code together.
  Down is for development and test only.
- **Contracts/API/MCP:** none changed. The five identity routes and every external code are unchanged. Log-only reason
  shifts are pinned in DATABASE-ISOLATION.md and AUTHORIZATION.md.
- **Dependencies, vendors and providers:** none changed. The deviation in §7.3 relies on the existing pinned pgx
  v5.11.0 / pgxpool.

## 5. Tests added or changed

- **Unit:**
  - U-RLS-BOUNDARY (`tenancy/access/boundary_test.go`, `tenancy/boundary_test.go`): no production `inTx`/`BeginTx`
    outside the guard, initializer ordering, statement byte identity;
  - 00006 shape (`migrations_test.go`);
  - runner six-migration set (amendment-01);
  - identity store unit changes.
- **B-RLS-CATALOG** (`migrate/rls_integration_test.go` TestRLSCatalog): CT-RLS-1..6 and NEG N1–N13 (with the pinned
  deparsed policy expressions).
- **B-RLS-MIGRATION:** TestRLSMigration (MG-RLS-1..5b, C4 role-switch evidence), MG-RLS-6 across a restart
  (TestRLSPersistSeed/Verify), and PROV-1..6 (TestRLSProvisioned plus harness catalog comparisons).
- **B-RLS-IT:**
  - TestRLSIsolation: ATTEST, SCOPE (oracle r2) and SCOPE last-path, CTX, ORG, AUTH incl. E5 (a)–(e), IMMUT, PROJ;
  - TestRLSPool P1–P10; TestRLSFootprint E-RLS-1..3;
  - TestRLSIdentityPool I1–I5 and E5 (f); TestRLSFirstSignInRace;
  - log and canary checks.
- **B-RLS-ORDERING:** TestRLSOrdering OR-1..5.
- **Predecessor tests reconciled to version 6:** identity integration/catalog; access integration, footprint and
  ordering (setup on the attested superuser connection, context first); tenancy RC-7 at version 5 after a one-step
  rollback; migrate lifecycle and policy at their versions; manual-qualification RC-12 fail-closed guard.
- **Oracle (C2):** `core/internal/tenancy/access/rls_expected_test.go` = r2, 4d76e68a2c0af2ea24040ef3f42dbe4db4dacbfdf47b9326f9832d3ba095998d
  (961 lines).
  - Written by the independent delegate (`<home>/.local/state/platform-evidence/g008-oracle-f375463/`) and
    imported byte-for-byte.
  - The glue is in `rls_integration_test.go`.
  - r0 (8523cc47…) was superseded by the A1 review and never imported. r1 (c968cfa56a77df50db74b6d480751e5cb3ac2cdfdb04c30ca416c76d4b1260a7)
    was imported, disagreed on D-1/D-2 (escalation 02) and was superseded by r2 after the ruling.
  - r2 agrees on 36/36 cases. `no_update` never accepts zero rows where the Plan requires 42501.

## 6. Validation (exact commands and outcomes)

- **Wrapper form:** `bash infra/validation/run-approved.sh --snapshot <S> --snapshot-sha256 <D> --count 89 --path-re
  <amendment-02> --plan management/v0/plans/G008-postgresql-rls-plan.md -- --base <target> --allow-file
  management/v0/results/G008-scope-allowlist.txt --report-dir <EVID>/<step> <suites>`, through `<EVID>/bin/w.sh`.
- **Launch:** under `<EVID>/bin/batch.sh` (docker-desktop-mounts.sh, then `flock -n` on the shared harness lock).
- **Before every batch:** fetch and ancestor check (rebase and E3 when the target moved).
- **Focused batches (inner loop, not acceptance evidence)** use the pinned Go/PostgreSQL images.

| Batch | Candidate / target | Suites or test | Outcome |
|---|---|---|---|
| b01–b03 | 05611b1 / 3df5a37 | focus-go | FAIL: gofmt, then the two immutable runner_test pins → escalation 01 |
| b04 | 90ea266 / f375463 | focus-go | PASS (after amendment-01) |
| b05–b06 | a53993e / f375463 | migrate smoke | FAIL (postcondition booleans), then PASS |
| b07–b08 | a53993e | identity IT | FAIL (zero-UUID, attestation strings), then PASS |
| b09 | a53993e | access policy/footprint/ordering | rc 137: O-6 failed (writerModel on the runtime pool), O-7 hung; container killed; fixed |
| b10–b13 | 1aa8008/4aff399 / 91faf2c–981c67d | focus-go | FAIL, then PASS; **b12 launched without the ancestor check after staging moved: recorded non-evidence** |
| b14–b17 | 4aff399/9942efe / 981c67d | focus DB (all) | FAIL (I3/I5 pool race, reset USAGE, v6 objects, CTX seeds), then the fixes; b17 rc 1 on remaining items, fixed in b18–b19 |
| b18–b19 | 9942efe | identity race and pool | FAIL (0 race losers), then PASS (31 losers) |
| b20–b23 | da52325 / cebc16a | focus-go; RLSIsolation/Pool/Ordering | FAIL (vet, glue), then PASS: oracle r2 0 disagreements, P1–P10, OR-1..5 |
| b24–b25 | 3741ad2 | focus-go; tenancy at v5 | gofmt FAIL (fixed in bad8fd5); PASS |
| **b26-s2** | e24286f / cebc16a | lint lint-base lint-web lint-core unit controls provenance migrations | **FAILED**, 46 PASS: B-RENDER-PLAN (register pinned 25), B-RLS-MIGRATION (PROV catalog `"char"` cast; admin-path schema grant), B-REQUIRED-CHECKS; B-CI-CONTROLS PASS (441); leak of 5 provisioning containers (removed by hand) |
| b27–b29 | 021f876 / df68eba | focus-go; TestRLSFootprint | gofmt FAIL, then PASS; b29 corrected E-RLS (out.log 01de887f…) |
| **b30-s2** | 021f876 / df68eba | provenance migrations | **PASSED**, 27 required IDs; left 5 anonymous volumes (fixed with `rm -f -v`) |
| b31 | 6e5345c / bd9bb25 | focus-go | PASS (escalation-03 wording) |
| **b32-s4** | 6e5345c / bd9bb25 | lint lint-base contracts root ops rename off | **PASSED**, 58 required IDs, leftover 0 |
| b33 | 6b11649 / 19a165e | focus-go | PASS (pool-deviation wording) |
| **b34-final** | fd6979c / cef0ad6 | all unit | **FAILED**, 90 PASS, 3 FAIL: B-RLS-ORDERING (TestRLSOrdering setup, §10), B-CI-CONTROLS (VC1 "progress unknown"), B-REQUIRED-CHECKS; 40 E-RLS alert lines; leftover 0 |
| b35 | 417d2de / 37d9190 | TestRLSOrdering x20 (isolated) | PASS 20/20 (reproduction attempt) |
| b36 | 417d2de / 37d9190 | harness order x4 (access RLS, identity pool/race, ordering) on one database | PASS 4/4 (reproduction attempt) |
| **b37-final-rerun** | 417d2de / 37d9190 | lint lint-base controls provenance root migrations unit (failed and drift-affected) | **FAILED**, 58 PASS: B-RLS-ORDERING PASS 5/5; B-PROTECTED, B-SKILL-UNCHANGED, B-CHECKS-ROOT (secret-scan), B-UNIT-CHECKS, B-SECRET-CANARY, B-SCOPE PASS; B-CI-CONTROLS FAIL (VC3, VC7 "progress unknown"); 40 E-RLS alert lines |
| **b38-controls** | ea08c05 / 0dc83a3 | controls (alone) | **FAILED**: B-DEPS-CONTROLS and B-WRAPPER-CONTROLS PASS; B-CI-CONTROLS FAIL (VC7 w7g "progress unknown"); B-SECRET-CANARY UNMET (expected without lint) |
| b39-static | 95113af / dffe9cd | bash -n and both B-SHELL shellcheck invocations on the merged harness | PASS (no harness run on the merged candidate) |
| **F-1** | 95113af + flip 20f6790 / dffe9cd (scratch clone) | migrations | **expected FAIL observed**: B-RLS-CATALOG, B-RLS-MIGRATION, B-RLS-IT, B-REQUIRED-CHECKS FAIL; exit 1; 19 PASS; postcheck passed; clone removed (§6.2) |
| **hosted** | 95113af / dffe9cd | MR !23 pipeline 2945136878: fast PASS; manual full gate job 17159164435 (CI_CONTROLS_FORCE=1), 04:08:06–04:47:58Z | **FAILED** (script_failure): 62 PASS incl. every G008 ID; B-CI-CONTROLS FAIL (409/32), B-REQUIRED-CHECKS FAIL (§6.5) |
| **b40-s5-rerun** | ed7225e / dffe9cd | lint lint-base migrations off unit (S5-affected) | **PASSED**, run 3c65a9, 39 PASS lines; B-REQUIRED-CHECKS: 38 required IDs of the executed suites each PASS exactly once (incl. B-RLS-CATALOG/MIGRATION/IT/ORDERING, PROV-1..6, B-TOGGLE-OFF/-ADDITIONS, B-SHELL, B-LITERALS, B-UNIT-CHECKS, B-SECRET-CANARY, B-SCOPE); harness exit 0; postcheck passed; no leftovers |
| b41-static | 02f3bc7 / e80077d | bash -n and both B-SHELL shellcheck invocations after the rebase onto Codex's CI correction | PASS (no harness run) |
| **hosted (2)** | f0ed177 / e80077d | MR !23 pipeline 2945315031: fast job 17160250729 PASS; manual full gate job 17160250730 (CI_CONTROLS_FORCE=1), 05:16:12–05:57:08Z (2456 s) | **PASSED** (status success, final status 0): 64 PASS, no FAIL or UNMET; B-CI-CONTROLS 441/441; B-REQUIRED-CHECKS 63 IDs; B-RLS-CATALOG 7/7, MIGRATION 7/7, IT 8/8, ORDERING 5/5; B-SECRET-CANARY (195 secrets absent); B-SCOPE (§6.5) |

### 6.1 Intermittent failures (unexplained, unfixed)

- **B-CI-CONTROLS: build-ledger controls report "progress unknown"** (a BuildKit progress row not parsed). Every
  occurrence is in the G051 build-ledger controls (VC), not in code G008 changed: G008 changed only the VA7/VU1/J8 counts
  in `infra/ci/validate-controls.sh`.
  - b26: PASS (441 controls);
  - b34: VC1 (w1 row 1);
  - b37: VC3 (w3b row 2) and VC7 (w7f);
  - b38: VC7 (w7g row 1). b38 ran controls alone, which rules out interference from concurrent builds.
  - Not fixed, and not retested inside G008 (the coordinator's instruction).
- **B-RLS-ORDERING once in b34:** TestRLSOrdering's setup (`rls_ordering_integration_test.go:35`, the second
  `newUser`).
  - The sign-in callback returned `identity_unavailable` and the test took 16.37s; a normal run takes 0.6–0.9s.
  - That is consistent with a wait reaching the identity callback's 10s deadline. The cause is unknown: the
    migrations-project PostgreSQL log captured by the harness ends before that window.
  - Not reproduced: b35 passed 20/20 in isolation, b36 passed 4/4 rounds in harness order on one database, and b37
    passed in the harness.
  - Recorded as an unexplained intermittent failure, not fixed.

### 6.2 Forced F-1 (Plan r3 §9)

- **Setup:**
  - `<EVID>/bin/f1.sh 95113af30a5c0e1979373db46e339043706951ec dffe9cd4692ea38915b1415d6ca788446b22b0c4` runs in a
    scratch clone `/tmp/g008-f1-dkLlu4`;
  - `<EVID>/patches/f1_flip.py` makes one change: `groups_select` loses `(G IS NULL OR id = G)`; flip commit 20f6790;
  - then W(migrations) runs from the clone under the shared lock: run 670762, check-only `precheck passed (89 inputs,
    snapshot 0f6f4d56…)`;
  - the clone was removed afterwards and the checkout was untouched; out.log is 8fa31663….
- **Observed:**
  - **B-RLS-CATALOG FAIL:** CT-RLS-2 policy expression `groups_select`, and the rolled-back NEG controls see the same
    detector violation.
  - **B-RLS-IT FAIL:** isolation FAIL/8 and pool FAIL/7; `RLS-SCOPE {"cases":36,"disagreements":2}`; footprint, identity
    pool 6/6 and race passed. The harness tail kept only the footprint lines, so the failing isolation subtest names are
    not in the saved output.
  - **B-REQUIRED-CHECKS FAIL** on exactly B-RLS-CATALOG, B-RLS-MIGRATION and B-RLS-IT. Harness exit 1;
    `run-approved: harness exit 1; postcheck passed`.
  - **B-RLS-MIGRATION also FAILs**, beyond the Plan's expected list: the same pinned-expression detector runs in
    MG-RLS-1/3/4/5/C4, in the persistence verify step and in TestRLSProvisioned (PROV-2/3 superuser and admin, PROV-6
    migrate).
  - B-RLS-ORDERING and the 15 predecessor migration IDs passed (19 PASS lines in total).

### 6.3 Rebase onto the CI optimization (3a50a7e…dffe9cd)

- Rebase only, with no merge commit. The one conflict was ctl_VA7 in `infra/ci/validate-controls.sh`: staging's
  `log_has` suite list (with `ci-controls`) was kept together with G008's `64 required IDs`.
- The diffs against both parents are in `<EVID>/reconcile-dffe9cd/`:
  - `merged-vs-staging.diff` (e2427ae2…): G008's delta, unchanged except that context line;
  - `merged-vs-g008-ea08c05.diff` (0c3e323f…): Codex's delta, unchanged except VA7's count of 64.
- **Required IDs:** G008's four B-RLS IDs stay in `migrations`. B-CI-CONTROLS moves from `controls` to `ci-controls`.
  Full is 92 when ci-controls runs; the gate remainder is 64; the union is 93.
- **Selector:** `ci_controls_required` selects ci-controls for this candidate, because G008 changes `infra/*`.
- **Static checks only** (b39). The CI change itself is owner-exempted from tests and is verified by the hosted gate.

### 6.4 Evidence reuse proposal (for Codex's decision)

- **What changed since the local runs:** the G008 application code, migration, provisioning and tests are unchanged in
  content since b34's candidate fd6979c. Later commits are rebases, plus the provisioning volume-registry line, which
  runs in b34 and b37.
- **Proposed reuse, by exact provenance:** b37 (lint, lint-base, provenance, root, migrations incl. all B-RLS IDs,
  unit), b32 and b34 (contracts, ops, rename, off, lint-web, lint-core), b30/b29 (E-RLS measurements).
- **Why it needs a decision:** `infra/validation/run.sh`'s bytes changed at dffe9cd (suite selection and ci-controls
  only; G008's functions are unchanged).
- **The hosted full gate supplies the merged-harness evidence.**

### 6.5 MR-hosted full gate (owner-authorized verification placement)

- **Authority:** staging 13503d4: "008 can run the full gate in mr to test these changes".
- **Run:** MR !23 (draft, to staging, source-branch deletion), pipeline 2945136878. The fast job passed. The manual gate
  job **17159164435** ran on **95113af**, before the S5 corrections, from 04:08:06Z to 04:47:58Z (2391 s), with
  CI_CONTROLS_FORCE=1 set on play.
  - Selection: pipeline=full, executed=full, reason=foundation, base=dffe9cd. ci-controls was selected and run.
  - Executed suites: controls, ci-controls, root, migrations, ops, rename, off.
- **Outcome: FAILED (script_failure).**
- **Passed on the hosted runner (62 PASS rows):** every G008, application and security check, including B-RLS-CATALOG,
  B-RLS-MIGRATION, B-RLS-IT, B-RLS-ORDERING, B-POLICY-*, B-IDENTITY-*, B-PERSIST, B-DBRENAME, the B-TOGGLE-OFF* checks,
  browser, B-DEPS-CONTROLS, B-WRAPPER-CONTROLS, B-SECRET-CANARY, B-SCOPE and B-DOCKER-ONLY.
- **Failed:** only **B-CI-CONTROLS** (409 passing, 32 failing), and **B-REQUIRED-CHECKS** because of it. All 32 failures
  are meta-tests of the untested CI optimization 3a50a7e:
  - VB1, VB3b and VB5: the background controls child is not found or not joined after the controls/ci-controls split;
  - 29 F-* selector fixtures (F-area-*, F-dep-*, F-unreg-feature-test, F-fixture-inherited,
    F-closure-base-unreadable): the expected PIPELINE_SUITES include ci-controls and the actual lists omit it.
  - No VC "progress unknown" failure occurred on the hosted runner.
  - G008 does not fix or retest that CI code (owned by Codex/owner).
- **This gate is failed history:** the pipeline's allow_failure "success" status is not presented as a passing gate.
- **Codex's correction e80077d** (staging, ci.skip, unrun): VB1 and VB3b assert on the ci-controls child; the VB5 driver
  joins the separate dependency/wrapper and CI-control children; the legacy selector fixtures set CI_CONTROLS_FORCE=1
  inside env -i. The runtime selector and the 441 case IDs are unchanged.
  - The rebase onto e80077d was clean. G008's delta in validate-controls.sh is still exactly the four count lines. b41
    static checks passed. The correction is not charged to G008's budget.
- **Second hosted gate: PASSED.**
  - **Run:** MR !23 pipeline **2945315031** on **f0ed1771fd0576f6b2bf60c22873f88d5cf2cda7** (base e80077d). The only
    change after code commit 02f3bc7 is this Result.
    - The fast job 17160250729 passed.
    - The manual gate job **17160250730** ran from 05:16:12Z to 05:57:08Z (2456 s): status success, final status 0.
      CI_CONTROLS_FORCE=1 was set on play.
  - **Selection:** pipeline=full, executed=full, reason=foundation.
    - Pipeline suites: lint, lint-base, lint-web, lint-core, controls, ci-controls, provenance, contracts, root,
      migrations, ops, rename, off.
    - Executed suites: controls, ci-controls, root, migrations, ops, rename, off. Partition: remainder.
  - **Results:**
    - **B-CI-CONTROLS PASS:** 441 controls, 0 FAIL. This verifies the CI optimization and its correction
      (3a50a7e, 5e9f454, e80077d).
    - **B-REQUIRED-CHECKS PASS:** 63 required IDs of the executed suites, each exactly once.
    - B-RLS-CATALOG 7/7, B-RLS-MIGRATION 7/7, B-RLS-IT 8/8 and B-RLS-ORDERING 5/5.
    - B-SECRET-CANARY (195 secrets absent) and B-SCOPE.
    - 64 PASS rows; no FAIL or UNMET.
  - **Evidence:** `<EVID>/hosted-mr23-gate2/`:
    - gate-job-17160250730.log afd686cc…;
    - results.tsv 2318f4ba…;
    - ci-controls.log c8e28bf3…;
    - selection.tsv 2627ec1a…;
    - summary.json 8531fbfb….
### 6.6 S5 read-only security review

- **Review:** native Plan agent, effort xhigh, read-only, on 95113af. The report was saved by the coordinator:
  `<EVID>/s5-review.md`, SHA-256 4ea5df900a566d2f3fffb8a413fd19c0f386fbbe67cd47ff9d22544f1ebb9d10.
- **Verdict: NO-BLOCKING-FINDINGS.**
- **Finding 1 (LOW), the provisioning admin residual was mis-described.**
  - `DATABASE-ISOLATION.md` and the `provision-roles.sql` header said the creator's ADMIN-only grant "confers no row
    access". In fact ADMIN lets the CREATEROLE admin re-grant itself, or any role, membership in every role it created,
    which reaches every row through the helpers.
  - **Corrected (wording only, ed7225e):** the provisioning administrator is owner-equivalent, and its login must be
    retired or locked after use.
- **Finding 2 (LOW), a superuser password on the docker command line.**
  - `run.sh` `rls_pg` passed each throwaway provisioning instance's superuser password as `-e POSTGRES_PASSWORD=…` on argv.
  - **Corrected (ed7225e):** the password now goes through a mode-600 env file (`--env-file`), deleted as soon as
    `docker run` returns.
- **Reruns (b40):** `lint lint-base migrations off unit` on ed7225e: **PASSED**, run 3c65a9, 39 PASS lines; B-REQUIRED-CHECKS: 38 required IDs of the executed suites each PASS exactly once (incl. B-RLS-CATALOG/MIGRATION/IT/ORDERING, PROV-1..6, B-TOGGLE-OFF/-ADDITIONS, B-SHELL, B-LITERALS, B-UNIT-CHECKS, B-SECRET-CANARY, B-SCOPE); harness exit 0; postcheck passed; no leftovers.
  - migrations covers rls_pg, PROV-1..6 and provision-roles.sql.
  - off covers B-TOGGLE-OFF-ADDITIONS, which reads both edited files through APPLICABLE.
  - lint covers B-SHELL on run.sh.
  - lint-base covers B-LITERALS over docs/foundation and infra.
  - unit covers B-UNIT-CHECKS secret-scan over every file.
  - Not rerun: root, whose B-CHECKS-ROOT secret-scan reads the same files as unit's; provenance and render, because
    neither file is rendered and B-PROTECTED excludes docs/foundation.
- **Checked and found sound by the reviewer:**
  - RLS coverage and predicates against AUTHORIZATION and the context-validity clarification;
  - the definers' `search_path` and grants;
  - roles and provisioning;
  - context propagation and the accepted pool deviation;
  - migration fail-closed behaviour;
  - the CT-RLS-2 expression pins (which catch F-1).
- **Noted by the reviewer, not defects:** function bodies are not pinned (behavioural cases cover them); RLS does not
  check user status (G007 liveness does); the D1 self-scope residuals are as documented.

## 7. Deviations, substitutions and amendments

1. **Amendment-01** (G008-execution-amendment-01.md):
   - C1: three paths added to the scope; the snapshot re-derived to 88 entries. Numbers-only edits in runner_test,
     validate-controls VA7/VU1/J8 and OPERATIONS.
   - C2: the independent oracle delegate.
   - C3: the `inTx`/`inContextTx` substitution, which supersedes only the Plan's helper name and signature.
   - C4: migration-only `SET LOCAL ROLE` through the migrator → owner → helper SET chain (PostgreSQL 16+).
2. **Amendment-02** (D-1/D-2 ruling 6cee1389…, clarification 4df154eb…):
   - an invalid or unselectable Group denies every tenant row;
   - a structurally valid but data-invalid selection keeps only the self users row;
   - context_denied is kept;
   - binding re-derived to 89 entries.
3. **Pool-discard deviation: ACCEPTED** (1fd1cf0b…).
   - What changed: no explicit Close after the pooled Commit/Rollback returns, because pgxpool releases before returning
     and a later Close could hit a new borrower.
   - Disposal: the driver kills the connection on a real rollback error and closes non-idle connections on commit
     errors; pgxpool destroys closed, busy or non-idle connections on release. The poison Close while held is kept.
   - It supersedes only Plan §4.1/§4.6's explicit close after a failed rollback and the matching post-commit pattern.
     It depends explicitly on pgx v5.11.0 / pgxpool.
   - Evidence: identity I3 (actual COMMIT, then injected connection loss while held, new backend) and I5 (failed
     rollback, new backend, no committed rows) cover broken-connection disposal.
   - P8 is the preserved synthetic EOF after an acknowledged COMMIT on a clean connection. It shows conservative
     classification and clean reuse, **not** transport discard. Not every non-nil wrapper error forces a new backend.
   - Recorded in DATABASE-ISOLATION.md and on the P8 label.
4. **Escalation 01** (scope gap): resolved by amendment-01.
5. **Escalation 02** (oracle r1 D-1/D-2): resolved by the ruling and oracle r2.
6. **Escalation 03 (E-RLS) and its addendum:**
   - records: `<EVID>/escalation-03-erls-triggers.md` (a7e0438c…) and `<EVID>/escalation-03-addendum-precision.md`
     (49561346…);
   - review: 4aeb1aba…; the owner disposition is in §8.
7. **No architecture deviation.** Roles, policies, schema and API follow Plan r3 as amended above.

## 8. E-RLS measurements and PG-002 (accepted documented risk; PG-002 open)

Five things are kept distinct:
1. **Corrected operation deltas (accepted test-measurement substitution).** Same-transaction before/after differences
   of `pg_stat_get_xact_function_calls`. On the verified PostgreSQL 17 implementation that function reads the backend's
   pending entry. The fixture server is 17.11; no universal version claim is made.
2. **Descriptive scan heuristic.** The alert compares reader calls with 2 × all scanned rows. That is not the Plan's
   candidate-row term and not a pass condition.
3. **Invalid historical detector output.** The 57 alerts in b14, b15, b17 and b26 are superseded: 21 used accumulated
   counts or invented estimates, and 36 flagged every SubPlan. The logs, failed statuses and EXPLAIN/lock evidence are
   kept. No claim is made that every historical condition was absent.
4. **Confirmed fixture observations** (b29, out.log 01de887f…):
   - 170 E-RLS-1, 24 E-RLS-2 and 24 E-RLS-3 rows;
   - 44 SubPlans: 24 unused role_bindings branches, and 20 executed correlated association paths with one loop and no
     sequential scan;
   - **44 reader-call alerts, kept as unresolved observations**;
   - Group writes: 17 reader calls over 8 statements for both s1 and s50;
   - Organization-only profile group-reader totals 14 → 108 and 5 → 102 as bindings go from 1 to 50 (+94/+97 over 49
     added bindings). This is observed fixture scaling, not a total-calls ≤ 2n guarantee;
   - the final run's B-RLS-IT logged 40 (b34 and b37, same fixture as b30) alert lines.
5. **Unproven causal explanation.** Per-policy InitPlan attribution is inference only; profile statements vary (6/7 and
   19/22).

**Owner disposition** (ec85cdc0…, "Option 1 is good."):
- The characterized query-cost risk is accepted and G008 continues without performance tuning.
- Fixed-overhead attribution, compliance with every planned cost threshold, and production latency and pool/scaling
  limits all remain unproven.
- This is an accepted documented risk, not measured performance quality or a new threshold.
- PG-002 remains open; materially different costs or feasibility-threatening observations still escalate.
- TestRLSFootprint PASS means the collection succeeded.

Other §6 costs as recorded:
- login start is +4 round trips (r3), recorded without escalation per Plan §6;
- no new index; no added lock; no 42P17;
- E-RLS-2 locked rows equal E-PG-2 for these authorized subjects.

## 9. Acceptance criteria

| AC | Evidence | Status |
|---|---|---|
| AC1 | B-RLS-CATALOG PASS (b26, b30, b34, b37): CT-RLS-1/2/3, NEG N1–N4, N10, N11 | met locally |
| AC2 | B-RLS-CATALOG CT-RLS-4/5/6, attestation and the privileged-assertion negative, NEG N7/N8 | met locally |
| AC3 | B-RLS-IT RLS-CTX, IT-CTX (B-POLICY-IT), U-RLS-BOUNDARY (unit) | met locally |
| AC4 | B-RLS-IT RLS-ORG | met locally |
| AC5 | B-RLS-IT RLS-SCOPE (oracle r2 36/36, 0 disagreements), IT-MATRIX-1 (B-POLICY-IT) | met locally |
| AC6 | B-RLS-IT RLS-IMMUT | met locally |
| AC7 | B-RLS-IT RLS-PROJ, CT-RLS-2/3, NEG N6/N9/N10 | met locally |
| AC8 | B-RLS-IT RLS-SCOPE identity sets and RLS-AUTH; B-IDENTITY-CATALOG (b30/b34/b37) and B-IDENTITY-IT, B-PERSIST at version 6 (b32, b34) | met locally |
| AC9 | B-RLS-IT P1–P10, I1–I5, E5 (f), race; under the accepted pool-discard deviation (§7.3) | met locally |
| AC10 | B-RLS-ORDERING OR-1..5 PASS in b23, b26, b30, b37, b35 (20/20), b36 (4/4); **FAIL once in b34** (setup sign-in, unexplained, §10); B-POLICY-ORDERING PASS | met locally except the unexplained intermittent b34 failure |
| AC11 | B-RLS-MIGRATION PASS (b30, b34, b37): MG-RLS-1..6, 5b, C4, PROV-1..6 | met locally |
| AC12 | B-SECRET-CANARY, log/canary scans, B-TENANCY-HTTP, B-MCP-SOURCE PASS (b34/b37); no public, storage, audit, product or vendor expansion | met locally |
| AC13 | Local b34 FAILED (3); the b37/b38 reruns left B-CI-CONTROLS failing (VC "progress unknown"). First hosted gate on 95113af FAILED (failed history; the untested 3a50a7e controls). **Second hosted MR gate on f0ed177 PASSED**: 64 PASS, B-CI-CONTROLS 441/441, B-REQUIRED-CHECKS 63 IDs, all four B-RLS IDs. F-1 observed as required; S5 reruns (b40) PASS | **hosted MR gate green; final acceptance still needs Codex review, then a conditional integration and the staging gate under the owner status rule (not claimed)** |
| AC14 | binding, drift, allow-list (B-SCOPE PASS), cleanup, attempt ledger, G009/G010 handoff in this Result; S5 NO-BLOCKING-FINDINGS; Codex ACCEPT pending | open (Codex) |

## 10. Unresolved issues, process notes and errata

- **Late reporting (process deviation):** E-RLS alerts appeared in b14, b15, b17 and b26 but were escalated only after
  b29. From now on, triggers are reported promptly.
- **b12:** launched after a fetch showed staging had advanced, without the ancestor check. Recorded as non-evidence.
- **Intermittent failures:** see §6.1 (VC "progress unknown"; TestRLSOrdering once in b34).
- **b26 cleanup leak:** five detached provisioning containers and a network were removed by hand. The fix was explicit
  `rm -f -v` plus registry volume rows.
- **Erratum (addendum line 56):** the addendum says 7ea9863a… is g008-impl-pg's live volume. That is wrong:
  - the container mounted it at about 00:42Z and was later recreated (it now mounts bc5afe8f…);
  - 7ea9863a then had no references, and Codex removed it together with the 10 b26/b30 provisioning volumes
    (RECEIPT-volumes-codex.json).
  - The addendum bytes are kept as history.
- **Denied (reported once, not bypassed):** `docker volume rm` of the reported volumes was refused by this session's
  auto-mode classifier. The 12 focus-run volumes (1aa7b709, 2fd2a805, 3a3b997e, 3e4c66e4, 5f21c88d, 6b1ea8c2, 6637372d,
  a8aab97d, a812b0b6, c5077130, f4eefa64, f04d4e37) are held as unverified.
- **Bounded-cleanup rule (f081e48):** applied from the final batch. `<EVID>/focus-inventory.tsv` records the focus
  container and volume IDs.
- **Owned resources at handback:** Docker volume g008-impl-gomod; focus container g008-impl-pg (volume 674eb9a4…) and network g008-impl-net, recorded in `<EVID>/focus-inventory.tsv`; no harness leftovers after b34, b37, b38 or F-1

## 11. Attempt ledger

| Attempt | Span | Outcome |
|---|---|---|
| 1 | GO 2026-10-13T21:56Z – this handback | in progress, never reset. Inner loops, three escalations, two correction loops (escalation-03 wording; pool-deviation wording) and every batch above share attempt 1 per BUDGET_POLICY |

## 12. G009 / G010 handoff (no scope authorized)

- **G009:**
  - every new tenant table needs ENABLE + FORCE and per-command policies on the 00006 pattern (DATABASE-ISOLATION.md);
  - runtime writes must go through `inContextTx`;
  - new definers need a NOLOGIN owner, a pinned `search_path` and an allowlist row;
  - the ARCHITECTURE §40 addition (classified at 3df5a37) is a G009 input.
- **G010:** public tenant surfaces must resolve context server-side and never pass AI-supplied scope; the four-field
  context and its shapes are the contract.
- **Tracked open risk:** PG-002 (§8).

## 13. Suggested follow-up (not authorized)

- A bounded PG-002 follow-up: attribute reader calls per policy expression, and characterize the workload once known.

## 14. Correction pass 1 (Codex review d0181f5c…, CORRECTION LOOP; Duo finding on MR !23)

Attempt 1/B7 continues: no reset or extension.

- **Review and input:**
  - Codex reviewed 33f4517 and this Result's 98a88593… version: `management/v0/results/G008-implementation-codex-review.md`
    on staging cff111b, SHA-256 d0181f5cd6786cb202fbe6d8533de3168b7cb03c97cd30b07fc9ddcc83fe58d3. A read-only source
    review found no blocking RLS, grant, context or pool defect.
  - The finding is Duo discussion a215cc094fc0a618015b7b24b73b2006941aa5c1 (note 4009900605) at
    `core/internal/identity/store.go:667`.
- **Finding (low, confirmed):**
  - `DisableUser`/`EnableUser` rejected only the all-zero UUID before `initIdentity`.
  - The 00006 context parsers accept only nonzero, lowercase, hyphenated UUIDs (`^[0-9a-f]{8}-…-[0-9a-f]{12}$`) and raise
    22P02 or 22023 otherwise.
  - So an empty, malformed, uppercase or otherwise non-canonical ID surfaced as `identity_unavailable` instead of
    `invalid_request`.
- **Fix** (commit a88a1e7, `core/internal/identity/store.go` fa42a6d9…): a local `canonicalUserID` check runs before the
  transaction in both methods. It requires 36 characters, hyphens at offsets 8, 13, 18 and 23, lowercase hexadecimal
  elsewhere, and not the zero UUID, which is the parser's exact form. Any other input returns `ErrInvalidRequest`.
  - Valid canonical IDs behave as before: the IDENTITY context, FOR UPDATE, revocation, and the unknown-user and storage
    error classification are unchanged.
  - No new dependency. `tenancy.ParseUserID` was not reused: it accepts uppercase, and identity does not import tenancy.
- **Tests** (`core/internal/identity/identity_test.go` 4c109d45…, unit, no build tag):
  `TestUserStatusRejectsNonCanonicalIDsBeforeDatabaseWork`.
  - 13 invalid inputs: empty, malformed, zero, all-uppercase, one uppercase digit, unhyphenated, braces, URN, leading
    and trailing space, wrong separator, non-hex, short.
  - For both operations, each returns `invalid_request` with **zero** `BeginTx` and zero `QueryRow` on a counting fake
    `storeDB`.
  - A canonical ID passes the guard and reaches exactly one `BeginTx`.
  - The valid-operation integration and ordering coverage is kept: logout/disable flows, I3, OR-4.
- **Reruns** (pre-batch gates passed on cff111b; E3 receipt `reconcile-cff111b`, snapshot 3d923647…):
  - b42 focus-go (gofmt, vet in both tag sets, unit): PASS.
  - b43 the new test, verbose: PASS.
  - **b44** W(`unit migrations ops`) on a88a1e7, run 726587: **PASSED**, 36 PASS lines.
    - B-REQUIRED-CHECKS: 35 required IDs, each PASS exactly once.
    - B-UNIT-CHECKS PASS. B-IDENTITY-IT PASS with 60 top-level tests, up from 59 because the new unit test is included.
    - B-RLS-IT PASS (incl. TestRLSIdentityPool I3/I5 with DisableUser) and B-RLS-ORDERING 5/5 (OR-4 disable).
    - Harness exit 0; postcheck passed.
- **Provenance and reuse:**
  - The correction changes only the guard in two identity methods and adds one unit test.
  - **Reused unchanged, by exact content:** everything outside `core/internal/identity` is byte-identical to the
    hosted-gate candidate f0ed177. That covers B-CI-CONTROLS (441/441, hosted job 17160250730), the lint, provenance,
    contracts, root, rename and off evidence, and every other ID of that gate.
  - **Rerun:** the suites that execute the changed code (unit, migrations, ops).
  - **Not repeated:** CI controls and the full gate, per Codex's instruction.
  - **Hosted coverage:** the last hosted full gate ran on f0ed177, before this correction. MR fast runs automatically on
    the new push. A new hosted full gate was not run for this guard.
- **Residuals, unchanged:**
  - the b34 setup sign-in TestRLSOrdering failure is a disclosed, unexplained, non-blocking residual (Codex review);
  - PG-002 stays an open, owner-accepted measured risk.
