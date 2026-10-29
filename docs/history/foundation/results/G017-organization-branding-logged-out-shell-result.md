> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G017 Result — Organization branding and logged-out shell

## Status

**Initial Result (R0): current-target reconciliation done (nothing material); execution not started; not accepted.**
2026-10-17 UTC.

- This is the execution-readiness record for the accepted
  [Plan r2](../plans/G017-organization-branding-logged-out-shell-plan.md) (SHA-256
  `8e8e848fd1994823534d9ec100c1109cd578b1e253d14cf9bb147ac76859a763`, preserved byte-exact) and its
  R0 addendum (not included), ahead of Codex's R0 review,
  the Git-first binding confirmation and the recorded execution release. It is not implementation and not acceptance.
- Nothing has run: no build, test, harness, Docker, provider call, migration, probe or settings change. No delegate
  has been dispatched and no MR exists. Implementation attempt 1 of B5 (checkpoint after attempt 3) has not begun;
  **0 of 5 attempts used**.
- Design authority: Plan r2 ACCEPT (provisional technical design) in the
  r2 review (not included) with the owner's D-1 decision ("017 owners and admins": owner+admin);
  G016's [final ACCEPT](G016-implementation-codex-review.md#final-implementation-acceptance--2026-10-17-utc) closes
  the dependency. P-1/P-2 belong to G057, not G017 ([split](G016-G057-qualification-split.md)).
- Implementation starts only after (1) Codex's review of the R0 addendum and this Result, (2) the binding confirmation
  below and (3) the recorded supervisor execution release.

## R0 binding (Git-first; prepared for Codex to confirm)

| Label | Bound bytes |
|---|---|
| Target: `origin/staging` at the R0 fetch, 2026-10-17 (G016 final ACCEPT; `e7a1ff36` had advanced by one management-only commit, no source difference) | `3f48ffa413c31b9a5cb1fd9641070e72ecf240d4` (tree `fea6dca25576d6f5655a4e74d9837fbb1b74ddc6`) |
| Plan r2 (P2) / r1 archive (P1) / r0 archive (P0) | `8e8e848fd1994823534d9ec100c1109cd578b1e253d14cf9bb147ac76859a763` / `5b72dbf2c561aadb0c40b6dc2d09b96f62bcef88e817eaf62374453d77204204` / `5d9a80e69c342075270c809954d4bd7f241c084cd17b747366bffd14aa644121` |
| R0 addendum (PA), this branch's commit | recorded at hand-back (see the commit on `g017-r0`) |
| Scope current (SC), `results/G017-scope-allowlist.txt`, 42 executable patterns (41 accepted + 1 forced, addendum §E) | recorded at hand-back; the r2 proposal `plans/G017-scope-allowlist.txt` stays `dc6373012c18f0b6003c615b79ef095a27a86a10dee3f829545678523b4ad8a1` |
| Goal (G) / Spec (S) with the owner D-1 section | `9993d11bfa1d0ece3327ea7d1e0254b37503448a49b6658ebcbff1bbb4c7a257` / `007e60fcc743ec648049495ae7f2bd8573bc433e7fe9b0afd8024abcd9ed39bd` |
| r2 review with the D-1 addendum (R2) / r1 review (R1) / r0 review (R0) | `a51a54997f8637158a685b7ae43f8c1272c61921d9c01552f4cb73a3690b23e0` / `bde78eb4cad312ffd7a65dd1d2568aa4f3051b2d7fda6918a947ba15d683982a` / `c88b8822f5f0f0fdbfa42ebccb4be40b144970a8fd06f54d8ca193598459f7f8` |
| G016 final review (I16) / G016 Result (X16) / G016–G057 split (Q57) | `c9e83621f4087ce4ed30651a47d15970c6f0e336242e93a727a703e1ed5d141f` / `185dfefd6a4b17001cae3478deeac36349f3c25abb5a680fa700d84eaa3ae8b2` / `8feeea4e12a63d1514441952bd47fe7c048f0ae2eedf7aa03e82b201fc234ab1` |
| Architecture (AR) / Security (SE) / Performance incl. PG-008 (PF) | `13ff1ae524a4d5332ff1246f202104e5f536867c28c84a55c32ba88c046753f1` / `75ed4515a43c2ef39d0f549090912d129bc7c546c060171ff61ba770a785992b` / `561b8ea17091cc2cbfebe679d4356ecf46e4d5b6d4ef7a3108a9903e1591e6ef` |
| Foundation notes WEB-SHELL / AUTHORIZATION / OBJECT-STORAGE / IDENTITY / DOMAINS / AUDIT / DATABASE-ISOLATION / RECOVERY | `c65f0a4b…97c96` / `329eb11e…7f350` / `0668545a…fe184` / `919be363…89dfd` / `c052f45f…06f3e` / `e6c26ff8…3ac55` / `1813f138…d19ca` / `76d4ad7a…1739b` (full values in addendum §D) |
| ADR-007 (A7) / Vestive skill (VS) | `4dbec3fd44ba5cd9bf7f539b343d1abb18f27bb1ae6415fc943e5dd7299128c4` / `7dfac5faecadb0b297a1be5641b1b86140dd28e6b4d32c78b140bc1cb65c670e` |
| Harness `run.sh` (RS) / `required-checks.tsv` (RC) / `lib.sh` (LB) / `validate-controls.sh` (VC) | `76efc47ef700ad7242ff867167d27e302b8f954ed7c9450fcc47f157a0c95d2e` / `9e6b1c7a03bc66e090a8f1465345b2f06c183927f1f646aa67e3998d0317a074` / `748b25aebed4f8ecd01d0d45c38c65e810490c047b2051e2f40b7935966e6b58` / `1c3148900b83c80ed201fdd285d016225fbc87d71638d31f041a45c07d510546` |
| Log/route enums (CS/EV), seeds (SD), overlay (IY), Playwright (PW), contrast sources (BT/DG/GC), last migration (M12) | as addendum §D |

**Bound paths.** P2/P1/P0 `management/v0/plans/G017-organization-branding-logged-out-shell-plan{,-r1,-r0}.md`; PA
`management/v0/plans/G017-organization-branding-logged-out-shell-plan-r0-addendum.md`; SC
`management/v0/results/G017-scope-allowlist.txt`; G `management/v0/goals/G017-organization-branding-logged-out-shell.md`;
S `management/v0/specs/G017-organization-branding-logged-out-shell-spec.md`; R2/R1/R0
`management/v0/results/G017-plan-r{2,1,0}-codex-review.md`; I16/X16/Q57
`management/v0/results/G016-{implementation-codex-review,hostname-resolution-domain-lifecycle-result,G057-qualification-split}.md`;
AR/SE/PF `docs/{ARCHITECTURE,SECURITY,PERFORMANCE}.md`; foundation notes `docs/foundation/<NAME>.md`; A7
`docs/adr/ADR-007-verified-domain-routing-and-authentication.md`; VS `web/.claude/skills/vestive-design/SKILL.md`;
RS/RC/LB `infra/validation/{run.sh,required-checks.tsv,lib.sh}`; VC `infra/ci/validate-controls.sh`; CS/EV
`infra/validation/tools/internal/logrecord/{completion.schema.json,events.json}`; SD `infra/identity/mcp-seed.sql`; IY
`infra/compose/identity.yml`; PW `web/playwright.config.ts`; BT/DG/GC `web/components/ui/{button,dialog}.tsx`,
`web/app/globals.css`; M12 `core/migrations/00012_session_binding_and_claim_exclusivity.sql`.

## R0 reconciliation summary (addendum §A–§G)

| Item | Outcome |
|---|---|
| Owner D-1 | **owner+admin**: `organization.manage_branding` = `org_owner` A, `org_admin` A, every other role D; same shape as `organization.manage_domains`; the cleanup route reuses the existing guarded `DeleteObject` policy. Affected assertions enumerated (addendum §A) |
| P-G016-1..12 against the merged code at `3f48ffa4` | all **confirmed** with current file:line; immaterial line shifts only (`run.sh` `domains_checks` `:3173` → `:3186`, `lib.sh` `WEB_PUBLIC_PINS` → `:573`, guard declared in `facade.go:18`). The merged session binding (`requestScope.org`, `BeginRequest(ctx, credential, origin)`, `Resolve` confinement at `context.go:152`) makes core refuse any Organization other than the session's binding, which the branding routes rely on in addition to their own `info.OrganizationID == org` check |
| Migration number | `00012` is the last migration; **`00013` is free** (no file, no reference) |
| Count pins | audit 35 (29 + 6) → 36; capabilities 12 → 13; matrix cells 264 → 286; `required-checks.tsv` 119 → 121; route enums 10 occurrences (+8) and `organization.*` events +1; Playwright projects unchanged |
| Changed bound inputs since r2's base | Goal/Spec (D-1 section), the r2 review addendum, the G016 closeout inputs; **none material** |
| Scope | 41 accepted patterns + `^core/internal/tenancy/access/integration_test[.]go$` (the 264-cell matrix pin) = **42**; the `plans/` r2 proposal bytes are untouched |
| G057 baseline | P-1/P-2 are not G017 requirements; the manual preview is the emulator-backed overlay |
| Executable selectors | r2 §6.3 unchanged; b2 pin files named in addendum §C |

**Budget and effort.** B5, checkpoint after attempt 3 (Goal); planning r0–r2 and this R0 spent **no attempt (0/5
used)**. Implementation effort: Opus at high (default); the owner may choose xhigh for a specific phase (candidate:
slice 2 image normalization / upload R-C seam). The actual runtime setting is reported as unknown unless verifiable.

**Worker record.** R0: task G017, role plan/R0; native isolated worktree
`<source-checkout>` (generated `agent-<id>`
name, the accepted AGENTS.md fallback; task/role/path recorded here), branch `g017-r0` from `3f48ffa4` (the earlier
local `g017-plan-r2` at `18873f74` holds the r2 planning commit, already imported on staging); model Fable 5.1,
assigned effort high, runtime effort unknown; session `<agent-session-url>`.

**Single-writer plan and resources.** As addendum §F: one implementer on `g017-organization-branding-logged-out-shell`
in a native worktree (`g017-agent-<id>` when the tool allows); `g017-core-check:b<N>` / `g017-web-check:b<N>` images,
`g017-focus` database project, the existing emulator-backed identity overlay for b4 and the manual preview; nothing
allocated yet; no new port, vendor, tunnel, deployment or live provider.

## Execution record

Execution release: R0 and execution release review (not included) (SHA-256
`20c1a43c1d746e084cb80c525e3fd76f8f37a2b5e58a47fdc0b0f3b06a5456c7`), recorded at staging `990953fb`.

### Attempt 1 of B5 (checkpoint after attempt 3) — started 2026-10-17 UTC

**Worker record (before the first code write).** Task G017, role **implement**, sole writer. Model Opus 5.5; effort
"high (requested)" — the runtime setting is not verifiable from inside the worker, actual value **unknown**. Native
isolated worktree, generated name `agent-<id>` (not renamed; the accepted AGENTS.md fallback), path
`<source-checkout>`. Feature branch
`g017-organization-branding-logged-out-shell`, created at `origin/staging`
`990953fb54afffced76711f059c9fb18469debf6` (HEAD at creation = that SHA). Session
`<agent-session-url>`.

**Bindings verified at creation (SHA-256, all match the release review):** release review `20c1a43c…56c7`; Plan r2
`8e8e848f…a763`; R0 addendum `c53ad65b…b034`; initial Result `2a92c5fd…6735`; scope `results/G017-scope-allowlist.txt`
`f51f2d1d…8639` (42 executable patterns). The delta `3f48ffa4..990953fb` is management-only (`e7a1ff36` archive/
receipt commit and `990953fb` release import; 8 files under `management/`), so no executable input changed since R0.

**Scope gap found before the first source write (hand-back (c), 2026-10-17).** A read-only sweep for version-12 pins
(every file naming `current_version\t12` or `00012_session_binding…`) found one file outside the 42-pattern scope:
`core/internal/mcp/env_integration_test.go:253` asserts `strings.HasSuffix(out, "current_version\t12\n")` after a full
`migrate up` (added by G016 A2 commit `83d115f3`, the same commit that pinned the other version-12 files). With
`00013_organization_branding.sql` present that assertion fails, so the MCP environment integration tests (hosted full
gate) would fail on any G017 candidate. **Smallest disposition requested:** add
`^core/internal/mcp/env_integration_test[.]go$` to the execution scope, limited to that one version pin (12 → 13) and
its comment; no other change in that file. Every other version-12 pin is already in scope (`migrate/{runner,integration,
domains_integration}_test.go`, `migrations_test.go`, `tenancy/integration_test.go`, `access/{integration,boundary,
recovery_integration}_test.go`, `run.sh`). No source has been written; in-scope slice-1 work proceeds on resume.

**Scope amendment 01 accepted** (release review §"Scope amendment 01", imported at staging `ce79ed9a`): current scope
`results/G017-scope-allowlist.txt` = **43 executable patterns**, SHA-256
`5997cf03e2fee84bb597371e58d80c1eb25daa8917fa4046a720d7ca941c652b` (the 42-pattern `f51f2d1d…8639` stays in history).
`core/internal/mcp/env_integration_test.go` is limited to the terminal migrate-up literal `current_version\t12\n` →
`\t13\n` and its directly corresponding comment. Release review now `aba24ba3f79b57e7a9cf10f4e8ff3be2ef5e1767552d573456a870c76776c48f`.

**Worker record — fresh successor #2 (before any code write).** Task G017, role **implement**, sole writer,
continuing **attempt 1 of B5** (checkpoint after 3; no new attempt, no reset). Model Opus 5.5; effort "high
(requested)" — runtime value **unknown**. Native isolated worktree, generated name `agent-<id>` (not
renamed; accepted fallback), path
`<source-checkout>`. Local branch
`g017-impl` tracking `origin/g017-organization-branding-logged-out-shell`, created at the checkpoint
`0c73e656bbbfd13657dadc0a094e13ff8b981b31`, then merged `origin/staging` `ce79ed9a18118a611daf4a566477575a9197f9fc`
(management-only: SUPERVISOR-HANDOFF, PROGRESS, release review, allow-list). Session
`<agent-session-url>`. Bindings verified after the merge: release review
`aba24ba3…c48f`, scope `5997cf03…652b` (43), Plan r2 `8e8e848f…a763`, R0 addendum `c53ad65b…b034`, Result at the
checkpoint `bc0d79b4…037c`.

**In-design choices carried from the predecessor (followed as recorded):**
- Membership fact (D-6): a new `identity/service.go` method reusing the existing `sqlIsOrganizationMember` from
  `identity/store.go` (no `store.go` change); only the session route pays the extra statement.
- Upload pre-check: `SetBrandingLogo` performs a guarded `branding.set_logo` read (authorization plus
  `expected_version`) **before** the upload; a denied or stale request uploads nothing, and its `access.denied` is
  attributed to the branding operation.
- Audit: `organization.branding_changed` uses resource type `organization` (resource_id = the Organization); the
  trigger derives the reason (`updated`, `reset`, `logo_set`, `logo_removed`) from the old and new rows.
- Expected in-scope edits: `infra/validation/run.sh` (WEB_ROUTES, B-MCP-SOURCE server imports and the `main.go`
  digest, since `cmd/server` imports `objectstore` and `branding/api`), `access/boundary_test.go`
  (`TestProductionImporters`) and `identity/store_test.go` (the forced-RLS table list).

**Slice 1 progress at the second gap (successor #2, uncompiled WIP checkpoint; no build/test/Docker run yet).** Written
in scope: the thirteenth capability (`access/policy.go`, `policy_test.go`, `matrix_test.go` oracle rows/pins 12→13 and
264→286, `access/integration_test.go:1469` 264→286); `core/internal/branding/{branding,colors,branding_test}.go` (label,
grammar, effective palette, T1–T10/N2–N4 oracle with whole-button ×1.12/×0.90, U-BRD-1..5); migration
`00013_organization_branding.sql` (table, validator, guard, capture, version-13 `events_attribute` and constraints
generated from the reviewed 00011 bodies, runtime policies, both auth-reader policies, `app.public_branding`,
postconditions, refusing down); audit catalog 35→36 (`audit.go`, `audit_test.go` reading 00013; reasons 54→58,
operations 24→29). Design note: the label's NFC rule is enforced by the database check `IS NFC NORMALIZED` (the Go
standard library has no normalizer; no third-party package added) — a non-NFC label is refused as `invalid_input`.

**Scope gap found (hand-back (c), 2026-10-17).** `core/internal/tenancy/access/audit_integration_test.go` (`TestAudit`,
subtest "AT-INV", lines 1283–1292) requires **every** action of `audit.TriggerActions`+`CoreActions` to be observed
(`actions never observed`), and `run.sh` B-AUDIT-IT pins `AUDIT-INV {"actions":35,"observed":35}` (`run.sh:3131`, in
scope → 36/36). G016 added its producers to that test (lines 1257–1271: slug change, domain claim/check). With the
thirty-sixth action `organization.branding_changed`, TestAudit (B-AUDIT-IT, hosted full gate) fails unless that test
produces one branding write. The file is not in the 43-pattern list. **Smallest disposition requested:** add
`^core/internal/tenancy/access/audit_integration_test[.]go$`, limited to one guarded branding write (an
`UpdateBranding` by the Organization owner beside the G016 producers at :1257–1271) and its directly corresponding
comment; no other change in that file. Sweep for further forced paths (files naming `organization_domains`,
`manage_domains`, the auth-reader functions, `AUDIT-INV`/count pins, recovery inventories): none other outside scope
(`events.json:243,325` capability enum and `run.sh:3124–3138` are in scope).

**Scope amendment 02 accepted** (release review §"Scope amendment 02", imported at staging `e27c3d39`): current scope
**44 executable patterns**, SHA-256 `30d15b87e8600e7884bab89390de7e57ab3a704bfabde8b60f09a4a4e1de4b1e` (the 43-pattern
`5997cf03…652b` stays in history); `access/audit_integration_test.go` limited to one guarded owner `UpdateBranding`
producer, its ordinary error check and its comment. Release review now
`4cbae5148afea95edd080985b14f3316af9d7409de26135b3aabe15bd05f67cf`. Codex guidance: the label contract keeps NFC
normalization plus trim; a DB-only `IS NFC NORMALIZED` rejection is not that behaviour.

**Worker record — fresh successor #3 (before any code write).** Task G017, role **implement**, sole writer, continuing
**attempt 1 of B5** (checkpoint after 3; no new attempt, no reset). Model Opus 5.5; effort "high (requested)" — runtime
value **unknown**. Native isolated worktree, generated name `agent-<id>` (not renamed; accepted fallback),
path `<source-checkout>`. Local branch
`g017-impl3` tracking `origin/g017-organization-branding-logged-out-shell`, created at the WIP checkpoint
`66594af6c15f5e68d659a26efcbd3ac62b098cec`, then merged `origin/staging` `e27c3d395bd80ef6b077fba462a66a54bb9326f4`
(management-only: release review, allow-list, handoff/progress) as `f786078bfd9dc6d1d299d5be626efda32d27d46a`. Session
`<agent-session-url>`. Bindings verified after the merge: release review
`4cbae514…67cf`, scope `30d15b87…1e4b1e` (44), Plan r2 `8e8e848f…a763`, R0 addendum `c53ad65b…b034`, Result at the
checkpoint `deb9bc08…5004`.

**Slice 1 complete (successor #3; compiled, focused b1/b2 PASS; still attempt 1 of B5).** Source checkpoint
`057b33e6` plus the fixes in the checkpoint commit that carries this record; `origin/staging` `b0873d12` merged
(management/skills only, no bound executable input). Written in scope (44-pattern list, 0 outside):
- `tenancy/branding.go` (five statements: plain read, its `FOR NO KEY UPDATE` lock, version-1 insert, version-guarded
  update, relation-free `normalize($1::text, NFC)`; the live logo state reuses objects.go's `sqlSelectObject`, so no new
  `stored_objects` statement yet) and its pins (`TestBrandingSQLBoundary`, `TestSQLBoundary`/`TestObjectSQLBoundary`
  exemptions, production-file count 14, five exported primitives pinned).
- `access/branding.go`: `GetBranding`, `UpdateBranding`, `ResetBranding`, `RemoveBrandingLogo` through `guard` with
  `organization.manage_branding` (lock order user → session → Organization FOR SHARE → branding row FOR NO KEY UPDATE);
  version 0 = no row; stale version → `conflict`; reset/removal never create a row or delete an object, and an unchanged
  row is not rewritten (no version advance, no audit row). New code `contrast_failed` (field `<pair>:<appearance>`);
  `access/boundary_test.go` pins (files, imports `core/internal/branding` + `encoding/hex`, exported surface, the 00013
  definer exception).
- **NFC (Codex guidance):** the accepted trim + NFC contract is implemented with an existing approved facility —
  PostgreSQL's built-in `normalize(text, NFC)` (UTF-8 database) inside the guarded transaction: `branding.PrepareLabel`
  (trim, UTF-8, no control characters) → `tenancy.NormalizeLabelTx` → `branding.ValidateLabel` (1–80 code points on the
  normalized label). Decomposed input is normalized and stored composed (BRD-3: 160 decomposed code points → 80
  composed, accepted); the `IS NFC NORMALIZED` check stays as defence in depth. No dependency, no contract change, no
  deviation needed. Cost: one extra statement on a labelled update only (PG-008 note for BRD-12).
- Migration 00013 fix found by BRD-1: the CHECKs call `app.branding_overrides_valid` as the writing role, so the
  runtime gets `EXECUTE` on it (the other three functions stay owner-only/auth-reader as reviewed).
- Tests: `migrate/branding_integration_test.go` `TestBrandingCatalog` (CT-BRD-1, CT-BRD-NEG N1–N8, CT-BRD-UPDOWN);
  `access/branding_integration_test.go` `TestBranding` BRD-1..7; amendment 02 producer in `audit_integration_test.go`;
  `branding_test.go` PrepareLabel cases. **BRD-14 moves to slice 2** (it needs a set logo, i.e. `SetBrandingLogo`).
- Version-13 pins: `migrate/{integration,runner}_test.go` (`brandingDrops`, rlsReset drops `public_branding`,
  `metadataThroughVersion13Only`, lifecycle outputs), `migrations_test.go`, `tenancy/integration_test.go` (13 → 12
  step-down), `access/{integration,recovery_integration}_test.go`, `identity/store_test.go` (14 tables, 13 RLS),
  `mcp/env_integration_test.go:253` literal only (amendment 01).
- Harness/enums: `events.json` (access.decision `branding.*` operations, `organization.manage_branding`, op→capability
  rule; `branding.request`/`branding.public` come with slice 2's routes); `run.sh` AUDIT-INV 36/36, B-MIGRATE-SET,
  root migrate block, B-MIGRATE-CLI/CATALOG (50 relations at 13, steps at 0/12/13), B-AUDIT-CATALOG (reads 00013, 36/30),
  B-RECOVERY-DRILL/B-PERSIST/B-DBRENAME/B-TOGGLE-OFF/B-IDENTITY-CATALOG version 13, `APPLICABLE` gains
  `core/internal/branding`; `docs/foundation/AUDIT.md` T30 row added now because B-AUDIT-CATALOG couples it to 00013.

| Batch (focused only) | Outcome |
|---|---|
| b1 core check image (`g017-core-check:b1`, `--network none`): gofmt; vet (+`-tags integration`); `go test ./internal/branding/... ./internal/audit/ ./migrations/`; `TestAccessPolicy|TestToolEligibility|TestProductionImporters|TestSQLBoundary|TestObjectSQLBoundary|TestBrandingSQLBoundary|TestPackageBoundary|TestMigrationParity|TestMembershipLockException`; `TestNormalSetIsTheReviewedSet` | r1–r3 found gofmt alignment, production-file count and exported-function pins (fixed); **r4 PASS** (all) |
| b1 validation-tools check stage (logrecord pin test with the new events.json) | PASS |
| b2 focus DB `g017-focus` (postgres via `compose run --use-aliases`, no host port; MinIO likewise for TestAudit) | r1: `TestBranding` 6 FAIL (runtime lacked EXECUTE on the validator → `internal`), CT-BRD-1 policy split on multi-line expression; `TestMigrationLifecycle` 11/11, `TestIdentityCatalog` 5/5, `TestTenancyModel` 36/36 PASS. **r2 (fixed migration + test): `TestBrandingCatalog` 4/4, `TestBranding` 8/8 (BRD-1..7), `TestMigrationLifecycle` 11/11, `TestIdentityCatalog` 5/5, `TestTenancyModel` 36/36 (13 → 5 step-down), `TestAudit` 10/10 with `AUDIT-INV {"actions":36,"observed":36}` PASS** |
| not run locally (by design) | MCP env pin (amendment 01: hosted); `TestRecoverySeed` passed its version-13 setup and then needs the drill's identity namespace env (`identity_namespace_mismatch`, harness prerequisite, not G017) — B-RECOVERY-DRILL hosted |

Cleanup: `g017-focus` down -v (0 containers/volumes/networks, generated 0600 env files removed); owned images
`g017-core-check:{wip,b1}`, `g017-vtools-check:b1`, `g017-focus-minio{,-init}` removed; evidence in
`platform-evidence/g017/`. No host port, vendor, tunnel or provider action.

**Remaining (slice 2 onward):** image validation (`branding/image.go`), `SetBrandingLogo` with R/C, the
unreferenced-upload listing statement, `PublicBranding`/`PublicBrandingLogo`, `branding/api` + server wiring
(`main.go` digest/imports, WEB_ROUTES, `branding.request`/`branding.public` events, completion enums), BRD-8..12 and
BRD-14, `http_test`; then slices 3–5 per Plan §8. Draft MR after slice 2 compiles and its focused checks pass.

**Checkpoint (hand-back (c), 2026-10-17).** Successor #3 stops at the slice-1 boundary because its context (conservative
estimate about 380k tokens) is past the 300k checkpoint; slice 2 is not started. A fresh successor continues attempt 1
from this pushed checkpoint (rebuild `g017-core-check:b1`, recreate `g017-focus` with the recipe above).

**Worker record — fresh successor #4 (before any code write).** Task G017, role **implement**, sole writer, continuing
**attempt 1 of B5** (checkpoint after 3; no new attempt, no reset). Model Opus 5.5; effort "high (requested)" — runtime
value **unknown**. Native isolated worktree, generated name `agent-<id>` (not renamed; accepted fallback),
path `<source-checkout>`. Local branch
`g017-impl4` tracking `origin/g017-organization-branding-logged-out-shell`, created at the slice-1 checkpoint
`10aa1bcbf72418625b98ee69cb02ce9355cfa93e` (HEAD at creation = that SHA); `origin/staging` `b0873d12` is already an
ancestor (no merge needed). Session `<agent-session-url>`. Bindings verified:
Result at the checkpoint `c82000a1…f8d9`, scope `30d15b87…1e4b1e` (44), release review `4cbae514…67cf`, Plan r2
`8e8e848f…a763`, R0 addendum `c53ad65b…b034`. The b1/focus recipes are preserved durably in
`platform-evidence/g017-scripts/{b1.sh,focus.sh,scope.py}`.

**Slice 2 progress (successor #4; WIP checkpoint, compiles; still attempt 1 of B5).** Written in scope (44-pattern list):
- `branding/image.go`: `NormalizeLogo` per r2 §1.4.1 — read bound 512 KiB, `http.DetectContentType` ∈ {png, jpeg}, `image.DecodeConfig`
  same format and 16..2048 px, buffering-independent PNG chunk walk (signature, length, CRC, allow-list, IHDR first/once,
  IEND ends the buffer) and JPEG marker walk (length-skipped segments, entropy data per scan with FF00/RSTn/fill, EOI ends
  the buffer), full `image.Decode`, draw onto an 8-bit NRGBA canvas, re-encode (PNG default compression; JPEG q90) through a
  writer that refuses beyond 512 KiB. Closed `invalid_input` fields: `logo_too_large`, `logo_type`, `logo_dimensions`,
  `logo_malformed`, `normalized_too_large`. `image_test.go`: 10 positives (tEXt PNG, 16-bit, palette+alpha, hand-built
  Adam7 interlaced PNG, APP1-thumbnail JPEG, the standard library's progressive/CMYK/restart-marker JPEGs, 16 and 2048 px)
  and 21 refusals (header-only/truncated PNG and JPEG, truncated progressive, bytes after IEND/EOI incl. after an APP1
  thumbnail, acTL, private and unknown chunks, bad CRC, SVG, HTML, GIF, WebP, 15/2049 px, 524 289 bytes, normalized too
  large). Note: the standard library's `video-001.progressive.truncated.jpeg` is structurally complete (it ends with EOI
  and decodes), so it is accepted; the truncated negative is a cut of the progressive file.
- `tenancy/branding.go`: `sqlShareLogoObject` (FOR SHARE), `sqlListLogoUploads` (Organization objects of a logo type,
  ready/deleting, not the current reference, `count(*) OVER ()`, LIMIT 100), `sqlPublicBranding` (the definer projection);
  `ShareLogoObjectTx`, `ListLogoUploadsTx`, `ReadPublicBranding` (single statement on a `Querier`). Pins:
  `TestBrandingSQLBoundary` (exact texts of the three), `TestObjectSQLBoundary` exempts branding.go, exported functions.
- `access/branding.go`: `SetBrandingLogo` (V in the guarded pre-check with authorization and `expected_version`; P = the
  unchanged G009 `PutObject` of the normalized bytes; A = guarded version-bound write locking the object FOR SHARE and
  re-checking ready/Organization object/type/size/SHA-256; definitive abort → C = guarded `DeleteObject`, reported
  `compensated`/`compensation_pending`, or `compensation_failed` with the state read live; `commit_outcome_unknown` → R =
  one guarded snapshot read: referenced → the view, else `unavailable` + `publication_outcome_unknown` + object + live
  state, object retained). Seam stages `P`/`A`/`R`/`C` on the existing `objectHooks` only. `GetBranding` lists the
  unreferenced uploads (+total). `DeleteBrandingUpload` (D-3 enforced in the facade: listed upload only, view at the
  caller's version). `PublicBranding`/`PublicBrandingLogo` (organization from the trusted resolver; `v` = sha12; bytes
  verified against key binding, length and SHA-256).
- `access/errors.go`: `Error`/`failure` gain `Object`/`ObjectState`, set only for the four logo outcome fields.
- `branding/api/http.go` (new): routes `GET /internal/v1/branding/public`, `GET …/branding/logo` (24 h
  `public, max-age=86400, immutable`, ETag, nosniff, inline; everything else no-store), `GET/POST /internal/v1/branding`,
  `POST …/reset`, `…/logo` (raw body; `?origin=`, `X-CSRF-Token`, `X-Expected-Version`; Content-Length > 512 KiB → 413,
  unknown length → 400), `…/logo/remove`, `…/objects/delete`; `begin`/`scope`/`originHost` copied byte-for-byte from
  domains/api; events `branding.request` / `branding.public`. `cmd/server/main.go` registers it.
- `access/boundary_test.go`: imports (`bytes`), exported decls/methods, importers (access may be imported by
  `internal/branding/api`; `branding/api` only by `cmd/server`; `internal/branding` only by access; objectstore by access
  and `internal/branding/api`).

**Design notes for Codex review (no contract change claimed; flagged for the S-review):**
1. Store construction lives in `branding/api.New` from the loaded `config.ObjectStore` (so `cmd/server` imports only
   `branding/api`, not `objectstore`); the objectstore importer boundary becomes {access, branding/api}.
2. The anonymous public read cannot use the facade's context-initialized transaction (no subject; `TestRLSBoundary`
   forbids a bare `inTx`): it is one autocommit statement on the pool (via a `tenancy.Querier` assertion on the
   Service's DB) calling only the auth reader's `app.public_branding`, like G016's resolver.
3. Pre-existing slice-1 defect found: `pinnedExportedMethods` was unsorted, so `TestAccessBoundary` failed; slice 1's b1
   selector did not include `TestAccessBoundary`. Fixed (sorted) and the selector now includes `TestAccessBoundary` and
   `TestRLSBoundary`.

| WIP batch (focused, `g017-core-check:wip`, `--network none`) | Outcome |
|---|---|
| build, `gofmt -d` (branding, tenancy, cmd/server), `go vet` (+`-tags integration` access/migrate) | PASS (empty) |
| `go test ./internal/branding/...` | r1 GOROOT lookup (trimpath) / r2 4 expectation fixes / **r3 PASS** |
| `TestAccessPolicy|TestToolEligibility|TestAccessBoundary|TestProductionImporters|TestRLSBoundary|TestSQLBoundary|TestObjectSQLBoundary|TestBrandingSQLBoundary|TestPackageBoundary|TestMigrationParity|TestMembershipLockException` | r1 method pin order / **r2 PASS** |

**Remaining for slice 2 (precise):** (1) `branding/api/http_test.go`: refusal-table equivalence for the eleven shared
codes and byte-equivalence of `begin`/`originHost` against domains/api source (go/parser + go/printer), public
projection exact field set, logo headers/`v` mismatch/foreign-host 404 no-store, 413/unknown length, over a fake
querier (QueryRow routing `hostname_state` vs `public_branding`) and a fake `objectstore.Store`; (2) integration
BRD-8..12 and BRD-14 in `access/branding_integration_test.go` on `setupObjects` (MinIO; `focus.sh minio` + `it … obj`):
BRD-10 via `faultService` hooks (stage P crash; P-stage concurrent update → definitive conflict → compensated;
`commitLostReply` → view; `commitLostNoCommit` → unknown/ready then guarded delete; R-stage concurrent update →
advanced version unknown; R-stage `setFail("C")` → `unread`; C-stage crash → `compensation_failed` with live state;
cancel at P), the in-flight case with a test-local DB wrapper whose Commit returns `io.ErrUnexpectedEOF` and commits
after release (Rollback a no-op while pending); BRD-11 with a service on `e.pool` (hookDB has no QueryRow); BRD-12
counts via `e.hook.counts()`; (3) `events.json` `branding.request` (get/update/reset/set_logo/remove_logo/delete_object)
and `branding.public` (projection/logo) with the closed reasons above; `completion.schema.json` core route enums
(+6 paths ×2 occurrences; the addendum's "+4 routes/+8" predates the final route set) and `logrecord_test.go`;
(4) `run.sh` B-MCP-SOURCE: new `main.go` digest, `MAIN_GO_IMPORTS` + `core/internal/branding/api`, access importers
+ `core/internal/branding/api`, objectstore importers + `core/internal/branding/api`; (5) b1 (incl. validation-tools
check) and b2 (`TestBranding` incl. new cases, `TestBrandingCatalog`); then push and the draft MR.

**Checkpoint (hand-back (c), 2026-10-17).** Successor #4 stops here at a compiled, unit-passing WIP boundary because its
context counter passed the ~250k hand-back threshold. Cleanup: owned image `g017-core-check:wip` removed; no container,
volume, network or port was created; evidence `platform-evidence/g017/{wip1,units}/`, scripts
`platform-evidence/g017-scripts/`.

**Worker record — fresh successor #5 (before any code write).** Task G017, role **implement**, sole writer, continuing
**attempt 1 of B5** (checkpoint after 3; no new attempt, no reset). Model Opus 5.5; effort "high (requested)" — runtime
value **unknown**. Native isolated worktree, generated name `agent-<id>` (not renamed; accepted fallback),
path `<source-checkout>`. Local branch
`g017-impl5` tracking `origin/g017-organization-branding-logged-out-shell`, created at the slice-2 WIP checkpoint
`09166baa0a3c1856be33d81ba67f3049a13cde9b` (HEAD at creation = that SHA); `origin/staging` `b0873d12` is already an
ancestor (no merge needed). Session `<agent-session-url>`. Bindings verified:
Result at the checkpoint `6aacd811…5802`, scope `30d15b87…1e4b1e` (44). Durable scripts repointed to this worktree.

**Slice 2 complete (successor #5; compiled, focused b1/b2 PASS; still attempt 1 of B5).** Written in scope on top of
`09166baa` (44-pattern list):
- `branding/api/http_test.go` (new; fake querier routing the resolver's `hostname_state` and `app.public_branding` by
  statement text, fake `objectstore.Store`): `TestPublicRoute` (exact public field set `dark,display_label,light,logo,
  revision` and `logo` = `content_type,v`; no row → `dark,light,revision` at 0; unknown/inexact/http origin → 404
  `unknown_host`; missing/repeated origin → 400; only the two reviewed statements; store failure → 503, never a clean
  unknown host; no-store/nosniff/JSON and no CORS header on every answer); `TestLogoRoute` (current `v` → bytes with
  exactly the six headers incl. `public, max-age=86400, immutable`, ETag, inline, nosniff; stale/short/uppercase `v`,
  a foreign host (routes to an Organization without that logo) → no-store 404 `not_found` with no store read, a host
  routing nowhere → 404 `unknown_host`; tampered bytes or a foreign binding → 503; missing bytes → 404); 
  `TestSetLogoPreReadBounds` (524 289 and 1 GiB → 413, unknown length → 400, bad/negative version, missing CSRF/version
  headers; at 524 288 the request proceeds to the session scope; zero body bytes read in every case);
  `TestRefusalMapping` (facade codes and logo fields → closed refusals; upload-outcome detail only with `Object`; 16
  codes); `TestDomainsCopyIsExact` (go/parser over both sources: `begin`, `scope`, `originHost` byte-identical; the ten
  shared codes identical in value and status; `errUnavailable` 503 in both under branding's own value);
  `TestNewRefusesIncompleteConfiguration`. The `http.go` comment "first eleven" was corrected to "ten plus
  errUnavailable" (comment only).
- `access/branding_integration_test.go`: `TestBranding` now runs on `setupObjects` (real MinIO; **the hosted
  `branding_checks()` must give `'^TestBranding$'` the object-store env**, as the objects suites have) and gains
  BRD-8 (PNG/JPEG normalized: same dimensions/family, server MIME, bytes differ from the input, `v` over the stored bytes,
  binding Organization-only, replaced logo listed as a ready upload, public projection), BRD-9 (SVG/HTML/GIF/truncated/
  appended/15 px/oversize/declared-size refusals: no object, no row, no event; the full positive/negative matrix stays in
  `image_test.go`), BRD-10 (nine cases: crash after P; definitive conflict at A → `compensated`/`deleted`;
  `commitLostReply` → view; `commitLostNoCommit` → `publication_outcome_unknown`/`ready` then the guarded
  `DeleteBrandingUpload`; the in-flight case through a test-local held-COMMIT wrapper (`brdHeld`: lost reply returned,
  transaction kept open, Rollback a no-op, committed on release → published afterwards); advanced version at R; R's read
  failing → `unread`; C crash → `compensation_failed`/`ready`; cancellation between P and A — observed outcome
  `compensation_failed`/`unread`, object retained ready and listed), BRD-11 (current logo only; replaced, foreign-v,
  malformed and empty `v` → not_found; B's logo object under A's context denied; as `<db>_auth_reader` (superuser
  fixture with `SET LOCAL ROLE`, rolled back): current logos visible, replaced logo and a non-logo object invisible,
  `public_branding(B)` readable), BRD-12 (counts below), BRD-14 (guarded `DeleteObject` of the current logo → view
  `deleted`, projection without logo, bytes not_found; simulated purge → `absent`; `RemoveBrandingLogo` clears with one
  `logo_removed` event). Deviation (wording only): the purge is simulated with the existing `unseed` fixture (attested
  superuser row delete plus key delete), not an owner-role connection.
- BRD-12 / PG-008 counts (hookDB/pool, this run): update 1 transaction / 8 statements; set_logo 4 / 28 (pre-check,
  G009 PutObject's own transactions, association); get 1 / 8; public projection 0 / 1; public logo 0 / 1 plus one object
  read.
- Logrecord: `events.json` `branding.request` (operations get/update/reset/set_logo/remove_logo/delete_object; the 16
  closed reasons; failed ⇔ `branding_unavailable`/`publication_outcome_unknown`) and `branding.public` (projection/logo;
  `ambiguous_parameter`, `unknown_host`, `not_found`, `branding_unavailable`; `host_tag` only); `completion.schema.json`
  both core route enums gain the **6 registered paths (×2 = 12 entries)** — `/internal/v1/branding/public`, `…/logo`,
  `/internal/v1/branding`, `…/reset`, `…/logo/remove`, `…/objects/delete` (8 routes; GET/POST share two paths). **This
  differs from the addendum's prediction of +4 routes / +8 entries**, which predates the final route set (r2 §5.3
  added the public projection, upload deletion and logo removal routes). `logrecord_test.go`: catalogue pin, 8 completion
  accepts, 6 event accepts, 11 rejections.
- `run.sh` B-MCP-SOURCE: `MAIN_GO_G016_SHA` (name kept) = the G017 `main.go` digest
  `d0d0942b33969807fd5f040fd3a566fe637b8517793cf8e28d74d31ff8ea4b36`; `MAIN_GO_IMPORTS` + `core/internal/branding/api`;
  access importers + `core/internal/branding/api`; objectstore importers + `core/internal/branding/api`; comment and
  record text updated to match. `bash -n run.sh` OK; the check itself runs in the hosted static gate.
- `branding/image_test.go`: gofmt of this image's toolchain splits two one-line helpers (formatting only).

| Batch (focused only; staging `b0873d12` ancestor of HEAD before each; mounts check 13 %) | Outcome |
|---|---|
| WIP unit `./internal/branding/api/` | r1 fixture `MaxTTL` above the 300 s bound (objects unconfigured → 503); **r2 PASS** |
| compile (`wip5`): build, gofmt, vet (+integration) | gofmt flagged two files (fixed); vet PASS |
| **b1 `b1-s2r1`** (`g017-core-check:b1`, `--network none`): gofmt (0 files), vet (+`-tags integration`), `./internal/branding/...` (incl. api), audit, migrations, `TestAccessPolicy|TestToolEligibility|TestAccessBoundary|TestRLSBoundary|TestProductionImporters|TestSQLBoundary|TestObjectSQLBoundary|TestBrandingSQLBoundary|TestPackageBoundary|TestMigrationParity|TestMembershipLockException`, `TestNormalSetIsTheReviewedSet`; validation-tools check stage (gofmt, vet, `go test ./...` incl. logrecord) | **PASS** (all) |
| **b2 `s2r1`** focus `g017-focus` (postgres + MinIO via `compose run --use-aliases`, no host port) | **`TestBranding` 14/14 top-level (BRD-1..12, 14; BRD-10's nine cases PASS), `TestBrandingCatalog` 4/4 PASS** |

Cleanup: `g017-focus` down -v (0 containers/volumes/networks, generated env files removed); owned images
`g017-core-check:{b1,wip}`, `g017-vtools-check:b1`, `g017-focus-minio{,-init}` removed. Evidence
`platform-evidence/g017/{wip5,b1-s2r1,s2r1-*.log,units}`. No host port, vendor, tunnel or provider action. No MR opened
(hand-back (b) for the draft MR decision).

**Remaining after slice 2:** BRD-13 (session `membership`), web/BFF and logged-out shell slices 3–5 per Plan §8,
`run.sh` `branding_checks()` (B-BRANDING-CATALOG/B-BRANDING-IT with the object-store env, it_count pins) and
`WEB_ROUTES`/required-checks rows, PG-008 source estimate in `docs/PERFORMANCE.md`, then final qualification.

**Worker record — fresh successor #6 (before any code write).** Task G017, role **implement**, sole writer, continuing
**attempt 1 of B5** (checkpoint after 3; no new attempt, no reset). Model Opus 5.5; effort "high (requested)" — runtime
value **unknown**. Native isolated worktree, generated name `agent-<id>` (not renamed; accepted fallback),
path `<source-checkout>`. Local branch
`g017-impl6` tracking `origin/g017-organization-branding-logged-out-shell`, created at the slice-2 checkpoint
`862db18d1ecedf7a4b8351e5f17033d616fd0f9b` (HEAD at creation = that SHA); `origin/staging` `b0873d12` is already an
ancestor (no merge needed). Session `<agent-session-url>`. Bindings verified:
Result at the checkpoint `98ece795…07a6`, scope `30d15b87…1e4b1e` (44), Plan r2 `8e8e848f…a763`, R0 addendum
`c53ad65b…b034`. Durable scripts repointed to this worktree. Draft MR !40 open for the branch.

**BRD-13 and slice 3 (successor #6; WIP checkpoint, focused checks PASS; still attempt 1 of B5).** Written in scope
(44-pattern list):
- **BRD-13 (D-6):** `identity/service.go` `Membership(ctx, user, org)` — one statement, the existing
  `sqlIsOrganizationMember` (`app.is_organization_member`, auth-reader definer) on the store's pool, no lock/context/
  renewal; `identity/http.go` session route asks it only **after** the origin/Organization binding check and answers
  `membership: "member" | "none"` (a failed read is `identity_unavailable`, never a guessed state); `authService` gains
  `Membership`. `http_test.go` `TestSessionRoute`: 8 fields, member/none/outage, not asked for a session bound to another
  Organization. `access/branding_integration_test.go` BRD-13 through the real handler (`identity.NewHandler(e.id)`):
  member → binding row removed in-test → `none` on the next request; the same credential at another Organization's
  origin stays 401 `unauthenticated` with no membership field.
- **Web:** `lib/identity-client.ts` `SessionView.membership` (closed; anything else malformed) + `identity-client.test.mjs`.
  `lib/branding.ts` (server-only): `readPublicBranding` (exact field set, the 15 names, `#rrggbb` lowercase, label
  bounds/no controls, logo `v` sha12 + png/jpeg; malformed = no branding), `resolveBranding` (diagnostic headers only, no
  credential), `fetchLogo` (bounded 512 KiB, type/length/ETag checked), `indicatorState` (anonymous / member / none /
  unavailable — no identity on failure), and the management client (`getBranding`, `updateBranding`, `resetBranding`,
  `removeBrandingLogo`, `setBrandingLogo` raw body with `X-CSRF-Token`/`X-Expected-Version`, `deleteBrandingUpload`;
  closed refusal codes of `branding/api` plus the closed `field`). `components/branding/tenant-frame.tsx`
  (`tenantRequest`, `viewerSession`, `shellBrand`): loopback profile or `unknown_host` → generic frame unchanged; a
  branding outage/malformed answer → tenant frame with platform defaults (no tenant data). `components/layout/
  brand-mark.tsx` (logo `<img src="/branding/logo?v=…" alt=label|product>` in a `h-10 w-auto` box, else the label text),
  `context-indicator.tsx`; `AppShell.brand` (root `TokenScope` label `tenant` with the overrides server-rendered, brand
  mark, indicator, footer attribution `PRODUCT_NAME · PLATFORM_BASE_DOMAIN`; without `brand` the markup is unchanged).
  `app/page.tsx` tenant entry (brand mark, workspace name, one neutral Sign in action — the header omits its own on the
  entry and sign-in pages — appearance control, keyboard help, attribution; signed in → Account link);
  `sign-in`, `account` (reuses its own session answer), `connections/mcp` pass the frame. `app/branding/logo/route.ts`
  (GET only; one 12-hex `v`; this request's own origin; the six reviewed headers; refusals no-store 404/503).
  Units: `branding-client.test.mjs`, `context-indicator.test.mjs` (the indicator is tested through `indicatorState`, the
  closed state the component renders verbatim — node's test runner cannot import `.tsx`; wording deviation only).
- `run.sh`: `WEB_ROUTES` + `branding/logo/route.ts`; B-WEB-BOUNDARY request-context allow-list + `lib/branding.ts`,
  `components/branding/tenant-frame.tsx`, `app/branding/logo/route.ts` and the five management form routes (comment
  updated; the PASS record text still lists the pre-G017 route set — update with slice 4).

**Design notes for Codex (flagged, no contract change claimed):**
4. Request IDs: a tenant page now makes the public branding read before its own core call; it carries a lookup request
   ID (`lookupContext`, as the G016 gate's lookup), and the MCP page's indicator session read likewise (its own call is
   the list). `docs/foundation/OPERATIONS.md` describes `lookupContext` as a narrow G016 exception and is **not in the
   44-pattern scope**; smallest disposition: one sentence there (scope addition) or Codex direction to record it in
   `WEB-SHELL.md` instead (in scope).
5. The tenant entry's single Sign in action: the header indicator omits its anonymous Sign in on the entry and sign-in
   pages (`ShellBrand.entry`), so each page has exactly one.

| Batch (focused; staging `b0873d12` ancestor of HEAD; mounts 13 %) | Outcome |
|---|---|
| compile `wip6` (`g017-core-check:wip`): build, `gofmt -d` (branding, tenancy, identity, cmd/server), vet (+integration) | PASS (empty) |
| unit `TestSessionRoute|TestResponsesAreNoStoreWithoutCORS|TestMethodsAreExact`, then all of `./internal/identity/` | PASS / PASS |
| web check `web1`–`web4` (`g017-web-check:wip`, `--network none`): lint, typecheck, `npm run test:unit` | web1 resolve-hook path / web2 field grammar (pair ids are uppercase `T7`) / **web3, web4 PASS (158/158)** |
| b2 `s6-brd13` focus `g017-focus` (postgres + MinIO, no host port) `^TestBranding$` | **PASS 15/15 top-level (BRD-1..14)** |

Cleanup: `g017-focus` down -v (0 containers/volumes/networks/secrets). Evidence `platform-evidence/g017/{wip6,web1..web4,
s6-brd13.log,units}`.

**Remaining (precise, for the successor):**
1. Slice 3 rest: `web/tests/fixtures/token-scopes/{fixtures.ts,page.tsx}` `branded` scope (a fixed passing palette) for
   S-TOKENS/S-PORTAL/S-STATES/S-CONTRAST/S-TEXT-SPACING; `web/tests/browser/support/modes.ts` if the branded scope needs a
   mode entry; b1 harness `run.sh lint lint-base lint-web lint-core` (`--base`, `--allow-file`) and b3
   (`tokens-light`, `tokens-dark`, `appearance`).
2. Slice 4: `web/app/organization/branding/page.tsx` (session + `getBranding` with a lookup ID; 403/`forbidden` →
   404 for non-managers; forms following `connections/mcp`: label + 15×2 colour inputs, reset, logo upload (multipart
   form → raw bytes ≤ 512 KiB), logo remove, unreferenced-upload delete; closed status map incl. `contrast_failed` with
   the pair id, `version_conflict`, `publication_outcome_unknown`), the five form routes, `components/branding/*.tsx`
   form parts, the `/account` link (rendered only when `getBranding` succeeds), `WEB_ROUTES` + the six paths,
   `infra/identity/mcp-seed.sql` branded org-a row (and org-c unbranded), E-BRD-1..6 in `identity.spec.ts`, b4.
3. Slice 5: docs (`WEB-SHELL`, `OBJECT-STORAGE`, `AUTHORIZATION`, `DATABASE-ISOLATION`, `MIGRATIONS`, `DOMAINS`,
   `RECOVERY`, `SECURITY`, `PERFORMANCE` PG-008 source estimate from r2 §5.6 + the BRD-12 counts above);
   `run.sh` `branding_checks()` (B-BRANDING-CATALOG `^TestBrandingCatalog$` ./internal/migrate/, B-BRANDING-IT
   `^TestBranding$` ./internal/tenancy/access/ **with the object-store env**, it_count pins 4 and 15),
   `required-checks.tsv` 119 → 121 and `validate-controls.sh` pins; the B-WEB-BOUNDARY record text; M1 preview per r2 §6.4
   (identity overlay L2 emulator, `IDENTITY_PROXY_PUBLISH=127.0.0.1:543`, seeded users) → hand-back (b).

**Worker record — fresh successor #7 (before any code write).** Task G017, role **implement**, sole writer, continuing
**attempt 1 of B5** (checkpoint after 3; no new attempt, no reset). Model Opus 5.5; effort "high (requested)" — runtime
value **unknown**. Native isolated worktree, generated name `agent-<id>` (not renamed; accepted fallback),
path `<source-checkout>`. Local branch
`g017-impl7` tracking `origin/g017-organization-branding-logged-out-shell`, created at the successor-#6 checkpoint
`e9d35f3fa9b73d7278979d30caa012734390017b` (HEAD at creation = that SHA); `origin/staging` `b0873d12` is already an
ancestor (no merge needed). Session `<agent-session-url>`. Bindings verified:
Result at the checkpoint `1fed4a41…1afb`, scope `30d15b87…1e4b1e` (44), Plan r2 `8e8e848f…a763`, R0 addendum
`c53ad65b…b034`. Durable scripts repointed to this worktree. Draft MR !40 open for the branch. Design note 4
(`OPERATIONS.md`, out of scope) is left for Codex's decision; not edited.

**Slices 3 (rest), 4 and 5 (successor #7; focused checks PASS; still attempt 1 of B5).** Written in scope (44-pattern
list; 79 changed paths since `origin/staging`, 0 outside):
- **Slice 3 rest:** `token-scopes/fixtures.ts` `branded` scope — all 15 overridable names in both appearances (the seed
  palette below; a scratch replica of `colors.go` pairs gives min ratio/threshold 1.20 Light, 1.31 Dark); `page.tsx`
  unchanged (it iterates `FIXTURE_SCOPES`); `tokens.spec.ts` scope count 6→7, branded ≠ default primary, S-RUNTIME
  sibling list, and branded joins the S-PORTAL-INIT, S-A11Y, S-SCREENSHOTS and S-TEXT-SPACING dialog lists. `modes.ts`
  not needed (no mode entry).
- **Slice 4:** `lib/branding.ts` form helpers (`readBrandingInput` — blank inherits, `#RRGGBB` lowercased, label trimmed
  and NFC, closed fields `display_label|light|dark`; `readMutationFields`; `isMultipartForm`; `readBoundedBytes`;
  `parseLogoUpload` — exactly `csrf_token`, `expected_version`, one non-empty `logo` ≤ 512 KiB; `brandingAnswer` — the
  closed PRG status map incl. `contrast_failed` + pair id (`PAIR_FIELD`), `version_conflict`,
  `publication_outcome_unknown`, `upload_not_found`; session/CSRF/origin refusals as the form layer's fixed pages;
  forbidden/not_available/unknown_host → fixed 404). `app/organization/branding/page.tsx` (tenant frame lookup read →
  own session read → `getBranding` with a lookup ID; non-manager and unrouted host → `notFound()`; anonymous → sign-in
  first), the five routes `update|reset|logo|logo/remove|objects/delete`, `components/branding/{branding-form,logo-forms}.tsx`,
  the `/account` link only when the management read succeeds. `run.sh` `WEB_ROUTES` + the six paths, B-WEB-BOUNDARY
  record and log text. `infra/identity/mcp-seed.sql`: carol is org-a's `org_owner` (her B/C memberships stay plain);
  org-a branded row (label `Acme <Works> & "Sons"`, the fixture palette, no logo, `updated_by` carol); B and C unbranded.
  E-BRD-1..6 in `identity.spec.ts`; test images `web/tests/fixtures/branding/{logo-teal.png,logo-amber.png,logo-vector.svg}`
  (generated, solid colour). Unit `branding-client.test.mjs` + 6 form cases.
- **Slice 5:** docs `WEB-SHELL` (new G017 section; Status), `OBJECT-STORAGE`, `AUTHORIZATION` (implementation status),
  `DATABASE-ISOLATION` (version 13), `MIGRATIONS` (row 13), `DOMAINS`, `RECOVERY`, `SECURITY`, `PERFORMANCE` (PG-008
  implementation source estimate with the BRD-12 counts). `run.sh` `branding_checks()` (B-BRANDING-CATALOG
  `^TestBrandingCatalog$` ./internal/migrate/, B-BRANDING-IT `^TestBranding$` ./internal/tenancy/access/, both with
  `$OBJ_ENV`), wired after `domains_checks` and into the skip list; `required-checks.tsv` 119 → 121;
  `validate-controls.sh` remainder pins 82 → 84 and full union 113 → 115 (the G016 pattern; integration unchanged).
- **Deviation (pins):** the handoff named it_count pins "4 and 15"; `it_count` counts every `--- PASS: <Test>/` line, and
  the attributable logs give **3** for TestBrandingCatalog (CT-BRD-1/NEG/UPDOWN; `b2r2-brd-catalog.log`) and **23** for
  TestBranding (14 top-level BRD cases + 9 nested; `s6-brd13.log`), so the pins are 3 and 23 (record text says so).
- **MR !40 checks failure (pipeline 2958539961, job 17255629751, head `0594ec3c`; Codex record
  `G017-mr40-checks-failure-20261017.md`):** fixed — `migrate/runner_test.go` `TestProviderAcceptsProtectedStore` 12 →
  13 sources (+ `srcs[12].Version == 13`, comment); `access/boundary_test.go` `TestRLSBoundary` export list + exactly
  `Service.Membership` (D-6 comment; assertion and session-origin negatives unchanged).

Design notes for Codex (flagged, no contract change claimed):
6. WEB-SHELL now states factually that tenant pages' branding reads carry lookup request IDs; `OPERATIONS.md` (design
   note 4) is untouched, pending Codex's decision.
7. E-BRD-1 checks label escaping, palette and attribution only; the logo bytes/headers checks (plan text under E-BRD-1)
   live in E-BRD-5, which uploads first (the seed has no logo) and restores no-logo/no-upload state in `finally`.
8. The management form outcomes are PRG redirects (303 to `?status=`), so a refused mutation is not a 4xx at the form
   route; session/CSRF/origin refusals keep the fixed-page status codes.

| Batch (focused; staging ancestor of HEAD at each; mounts 13 %) | Outcome |
|---|---|
| web `web5` (fixture) / `web6`, `web7` (slice 4) / `web8` (fixes) | PASS 158/158 / typecheck SEED name clash → PASS 164/164 / PASS 164/164 |
| b1 harness `run.sh lint lint-base lint-web lint-core --base <staging> --allow-file` `b1h-1` @`503ca939` | FAIL only B-SECRET-SOURCE S3 (harness self-test "untracked canary exit 0, want 1"; no G017 input; scanner files unchanged) |
| b1 harness `b1h-2` @`0594ec3c` (incl. slice 4) | **PASSED** (B-SECRET-SOURCE incl. S3, B-WEB-BOUNDARY with the six routes, B-WEB-TOKENS/COMPONENTS/DESIGN, B-SECRET-CANARY, B-SCOPE 68/68) |
| b3 tokens-light/dark `b3-2` (runner on the identity overlay's web namespace) | **PASS 28/28**; appearance 31 invalid there (web at 127.0.0.1 is the tenant-gated HTTPS profile) → rerun `b3a` on the plain stack |
| b4 overlay g017-b4 (identity:start/migrate, harness mcp_seed reproduced, identity:browser narrowed) `b4-1` | E-BRD-1/2/3 + E-DOM PASS; E-BRD-4/6 S-ERRORS (expected 404 on the test page), E-BRD-5 duplicate `id="logo"` (real defect) → fixed (`logo-file`, 404 probes via `apiPage`/`statusAt`) |
| `b4-2` | invalid: web recreated without the proxy (proxy shares web's namespace) → proxy recreated; script fixed |
| `b4-3` @`b173bb12` | **PASS 11/11** (E-DOM-1..5, E-BRD-1..6) |
| pins `pins-1` (`TestProviderAcceptsProtectedStore`, `TestRLSBoundary`, gofmt) | **PASS** |
| b3 appearance `b3a-1` @`b173bb12` (plain local stack g017-b3, loopback ports of the downed overlay) | **PASS 31/31** |

Cleanup: `g017-b4` down -v (0 containers/volumes); `g017-b3` down -v after each run (0 left); `g017-core-check:pins` removed.
Evidence `platform-evidence/g017/{web5..web8,b1h-1,b1h-2,b3-1,b3-2,b3a-1,b4,b4-1..b4-3,pins-1,scratch}`.

**M1 preview ready (Plan r2 §6.4; deliberately kept running for the operator).** Compose project `g017-b4` (identity
overlay, L2 emulator) built and started at candidate **`b1042ef121769f99cdf70c879ca6eb70f639915a`** with
`IDENTITY_PROXY_PUBLISH=127.0.0.1:543` (loopback only; the overlay's other services publish the configured
standard+100 loopback ports), `identity:migrate` (version 13), the seed (`mcp-seed.sql` SHA-256
`9e9988a95257d5675045e6fbd871754c342d0dcd0eeb67f3763692ce7bfbdc35`). Run CA `platform-evidence/g017/m1/ca.pem` (trust
only in a dedicated browser profile); seeded users `platform-evidence/g017/m1/users.json` (0600; values never in
chat/logs): carol = org-a owner (management), alice = member. Smoke: `https://org-a.localhost:543/` 200 with the escaped
label; `https://org-c.localhost:543/` 200. Feedback items pending (Light/Dark/System, 320/1280, dialog in the tenant
scope, sign-in as member, logo upload and SVG rejection). Cleanup after M1: `b4.sh down` (`down -v`) and the receipt.

**Remaining (precise):** operator M1 feedback bound to `b1042ef1` → resolve findings, rerun affected selectors; Codex
decisions on design notes 4 and 6–8; S-review; MR !40 checks on the pushed head; final reconciliation and Result; MR to
ready after Codex pre-integration acceptance.

### Successor #8 — owner M1 appearance dropdown (amendment 03), attempt 1 of B5 continued — 2026-10-17 UTC

Identity: Claude Code, model Opus 5.5, effort high (requested; actual unverified), sole writer. Native worktree
`.claude/worktrees/agent-<id>` (generated name), local branch `g017-impl8` tracking
`origin/g017-organization-branding-logged-out-shell` at `bb3c3e88`, merged `origin/staging` **`091be3d6`** (management
only; includes the named `f43abd56` plus G018 review records) → HEAD `1b7f6fcc`; staging ancestor of HEAD verified.

Plan/Result binding delta (no replan): release review `G017-r0-execution-release-codex-review.md` SHA-256
`455fad05…c49` §"Scope amendment 03 — owner M1 appearance dropdown"; allow-list `G017-scope-allowlist.txt` **48
patterns** `826e3368…de4`; prior Result `8035845c…5ad` (at `bb3c3e88`). Owner M1 feedback: "Looks good, just replace
the Light/Dark/System segmented switch with a small dropdown select; the trigger shows the selected symbol." Change:
shared `appearance-control.tsx` → native `<select>` + existing lucide icons (Sun/Moon/Monitor; Monitor whenever the
stored preference is System), accessible name "Appearance" + selected value, tokens only; storage, live OS change,
pre-paint and hydration unchanged; radio-based browser pins adapted to the select. Focused validation only (web
lint/typecheck/unit; b3 appearance + tokens-light/dark; affected a11y/design/shell specs); M1 preview refresh + screenshots.
M1 stays open pending owner feedback on the dropdown.

**Implementation (candidate code commit `3051f7ed`; merged staging `0568d610`, management only, → candidate HEAD
`86ed498b`).** `web/components/layout/appearance-control.tsx`: a `<label data-appearance-control>` trigger (ruled
`border-input`, `min-h-9`, `hover:bg-muted`, `ring` outline on `:focus-visible`) holding an `aria-hidden` symbol span
(lucide `Sun`/`Moon`/`Monitor` + `ChevronDown`; `lucide-react` 1.52.0 already pinned, no new package or menu
framework) and a transparent native `<select aria-label="Appearance">` covering the whole trigger (options Light, Dark,
System; token `bg-background`/`text-foreground` for the native list). `setAppearance`/`useAppearance`, storage, the init
script and its hash are unchanged. `globals.css` §D now shows exactly one `[data-appearance-icon]` keyed on
`html[data-appearance]` (pre-paint; System → Monitor whatever the OS resolves); the radio selected-fill/bar rules are
removed. `docs/foundation/WEB-SHELL.md` Control paragraph updated.
Pins adapted (not deleted): `appearance.spec.ts` — state reads the select value plus the one visible symbol, every
`expectResolved` asserts the symbol equals the stored preference (incl. System under OS light/dark), the held-JS
first-paint test asserts the symbol before hydration, keyboard choice = focused select + ArrowDown/ArrowUp, pointer
choice = `selectOption` by label, new `[S-APPEARANCE-CONTROL]` (name "Appearance", value, labelled options, no radios,
Tab order, keyboard, 320px touch tap on the trigger reaches the select, covered target ≥24px); `shell.spec.ts` — one
combobox "Appearance", no radio, tab-order name via the select; `accessibility.spec.ts` — symbol non-text contrast vs
header and trigger ground at rest/hover, options, keyboard focus ring on the trigger, spacing item = the trigger
(`kind: icon`); `design.spec.ts` — trigger ground, input frame, symbol in `foreground`, two visible SVGs, font node =
the select (Barlow Condensed 600); `support/spacing.ts` — `select` counts as a usable control, reflow uses the trigger.
Design note 9 (flagged): the Vestive skill prefers MDI filled icons; amendment 03's "existing lucide icons" prevails.
Option padding was tried and dropped (ignored by Chromium's native picker; never pushed).

**Scope gap — needs a Codex amendment (not changed):** `web/tests/browser/account.spec.ts` (shell-light/dark,
S-ACCOUNT-LINK and S-ACCOUNT-A11Y) pins the old control — `c.tagName === "FIELDSET"` and a focused
`input[type=radio]` → "appearance" — and matches none of the 48 patterns, so it will fail on the next MR checks until
amended. The exact two-line adaptation (same assertions, select/trigger instead of radio/fieldset) is prepared outside
the repo: `platform-evidence/g017/impl8/account-spec-amendment.patch` (SHA-256 prefix `dac0a9b48430`). Proposed
pattern: `^web/tests/browser/account[.]spec[.]ts$`.

| Batch (focused; staging `0568d610` ancestor of HEAD `86ed498b`; mounts 13 %; owned ports verified free) | Outcome |
|---|---|
| web `i8-web1` (build, lint, typecheck, unit) | **PASS** lint 0, typecheck 0, unit 164/164 |
| browser `i8-b3s1` (plain stack `g017-b3`; `g017-b4` paused with `stop`, volumes kept): appearance, dev-hydration, tokens-light/dark, shell-light/dark narrowed to accessibility/design/shell specs | **PASS 122/122** (appearance 32, dev-hydration 6, tokens 28, a11y 8, design 24, shell 24) |

Not run by instruction: full browser/ops/control campaigns, `account.spec.ts` (out of scope, above). Scope: 86 paths
changed against staging, all inside the 48 patterns.

**M1 preview refreshed:** `g017-b4` restarted (`start`; same CA `m1/ca.pem`, users `m1/users.json`, database and seed
`9e9988a9…`), then `IDENTITY_PROXY_PUBLISH=127.0.0.1:543 b4.sh web` at `86ed498b` (web + proxy recreated). Smoke
`https://org-a.localhost:543/` 200 and `https://org-c.localhost:543/` 200, both serving the select. Screenshots
(headed Chromium on Xvfb framebuffers, so the native list is captured; content area only) in
`platform-evidence/g017/m1/screenshots/`: `desktop-1280-{light,dark,system-os-dark}-{closed,open}{,-detail}.png`,
`mobile-360-system-os-{light,dark}-{closed,open}.png`. Scripts repointed to this worktree; new `b3s.sh`, `shots.sh`,
`shoot.cjs`, `b4.sh stop|start`. Cleanup: `g017-b3` down -v (0 left); `g017-b4` kept running for M1.

**Remaining:** owner M1 feedback on the dropdown (M1 open); Codex amendment for `account.spec.ts`; MR !40 checks on
the pushed head; then the prior remaining items (design notes 4, 6–9, S-review, final reconciliation).

**Amendment 04 binding (recorded before applying).** Merged staging `4c20a2d5` (management only) → `c3db2762`. Release
review SHA-256 `4654b2f4…a5f9` §"Scope amendment 04 — account appearance selectors"; allow-list **49 patterns**
`66328e75…c2d1` (adds exactly `web/tests/browser/account.spec.ts`); patch
`platform-evidence/g017/impl8/account-spec-amendment.patch` `dac0a9b4…d86d` (two selector substitutions, assertions
preserved), applied verbatim and nothing else.

**Amendment 04 applied (commit `f27b4f02`).** `account.spec.ts`: exactly the two selector substitutions of the
patch. Focused run `i8-acct1` @`f27b4f02` (preflight: staging `4c20a2d5` ancestor, mounts 13 %; `g017-b4` paused with
`stop`, plain stack `g017-b3`, `--grep "S-ACCOUNT-LINK|S-ACCOUNT-A11Y"`): **PASS 4/4** (S-ACCOUNT-LINK and
S-ACCOUNT-A11Y in shell-light and shell-dark). Cleanup: `g017-b3` down -v (0 left); `g017-b4` restarted (`start`,
same containers/volumes, web at `86ed498b` = the same web source), smoke org-a/org-c 200; preview kept running. Scope:
87 changed paths against staging, all inside the 49 patterns. Merged staging `3137ca09` (PROGRESS only; SHA-256
`1cf0db58…29a9` records the owner feedback).

**M1 coverage (actual).** The owner reviewed the rest of the G017 UI positively at the earlier preview (feedback
"Looks good", requesting only the dropdown). From the refreshed screenshots the owner accepted the updated
appearance dropdown ("Looks great"). The owner's only question — whether the footer's "Beginify" is hardcoded — was
answered by Codex: it is the configured `PRODUCT_NAME`/`PLATFORM_BASE_DOMAIN` platform attribution (not the tenant
label); no change requested or made. The owner did **not** manually execute the logo-upload, SVG-rejection,
member sign-in or other interactive cases; those remain covered only by the automated E-BRD/E-DOM and shell
evidence recorded above.

**Remaining for final qualification:** S-review (security/self-review of the full G017 diff); Duo review; MR !40
checks on the pushed head; Codex pre-integration acceptance; then the integration gate (authorized conditional merge
and the hosted staging qualification). Codex decisions on design notes 4 and 6–9 remain open. After M1 closes,
`g017-b4` is taken down (`b4.sh down`) with a receipt.

MR !40 checks on `3d715bc2`: pipeline 2958686504 **success** (and 2958677737 on `d951ca53` success).

### Correction round (Codex implementation review CORRECTION LOOP + S-review D-1) — attempt 1 of B5, no reset

Bindings: `G017-implementation-codex-review.md` `366186b4…ec0` (two web corrections), `46d3bc3e…62c2` (§"Security
self-review D-1"), `44af7185…483` (§"Remaining self-review notes disposed"); S-review
`platform-evidence/g017/s-review-3d715bc2.md`. Frozen base `3d715bc2`; staging merged at `82519f23` and `bab675f4`
(management only).

| Fix | Commit | Diff (file:line) |
|---|---|---|
| 1 entry signed-in state | `3686ca02` | `web/app/page.tsx:40-41` `signedIn` = `brand !== undefined && (kind === "member" \|\| kind === "none")` (was `kind !== "anonymous"`, which counted `unavailable`) |
| 2 wide logo preview | `3686ca02` | `web/components/branding/logo-forms.tsx:37,39` box `max-w-full min-w-0`; img `max-w-full min-w-0 object-contain` (alt unchanged); `web/tests/browser/identity.spec.ts` E-BRD-5 +20 lines (2048×16 replacement checked at 320px); `web/tests/fixtures/branding/logo-amber.png` regenerated 64×64 → 2048×16 (same colour; legal 16..2048) |
| D-1 decode after authorization | `ab4c20da` | `core/internal/tenancy/access/branding.go` `SetBrandingLogo`: `pre` keeps only the bounded read and size checks (stores `raw`); `NormalizeLogo` moved after the successful guarded pre-check (authorization, liveness, expected version), outside any transaction, before P; refusal = same closed `brandingValidation` mapping, logged via `logGuarded`; doc comment updated. Later rechecks, audit and unknown-commit handling unchanged. `branding_integration_test.go` BRD-9: a plain member's truncated and trailing-bytes PNGs → `forbidden`, no field, objects unchanged |

Focused outcomes (preflight each: staging ancestor, mounts 13 %, owned resources checked, under the lock):
- `i8-web2` @`3686ca02`: lint 0, typecheck 0, unit **164/164** (incl. the context-indicator file: anonymous,
  member, none, unavailable states). Truth table of the exact `page.tsx` predicate (`tt.mjs` in the check image):
  anonymous false, member true, none true, unavailable false, no brand false — **all as required**.
- Fix 2: E-BRD-5 on the preview overlay `i8-ebrd5` could not run — it stopped at its first step (carol's
  `account-branding-link` absent) because the preview emulator regenerated its user IDs when `g017-b4` was paused and
  restarted for the earlier batches (seeded subjects no longer match; verified read-only with `b4.sh users`). A
  preview-DB repair was refused by the permission system and was not attempted otherwise. Instead, the allowed
  equivalent focused actual render `i8-wide1` @`3686ca02`: the real compiled CSS of the preview web at 320px with the
  exact `Current` markup/classes read from `logo-forms.tsx` and the 2048×16 fixture: natural 2048×16, object-fit
  contain, box 0–320 (inside the viewport), image 10–310 inside its box (300×44, aspect kept by containment), alt
  kept, scrollWidth 320 = innerWidth — **PASS** (`platform-evidence/g017/i8-wide1/{metrics.txt,wide-logo-320.png}`).
  The amended E-BRD-5 runs in the MR/identity checks.
- D-1 `i8-d1` @`e29b7f64`: core check image build 0, gofmt 0 files, `go vet -tags integration` 0;
  `TestBranding/(BRD-2|BRD-8|BRD-9)` on disposable `g017-focus` Postgres+MinIO: **PASS 3/3** (+ parent). Cleanup
  0 containers/volumes/networks/secrets; image removed.

Design notes resolved by Codex: 4 and 6 (WEB-SHELL docs), 7 (E-BRD-5 evidence placement), 8 (PRG), 9 (lucide)
ACCEPTED; the 3/23 count pins VERIFIED. S-review: D-1 fixed here; H-1 deferred to a bounded G021/pre-deployment
assessment (no format or scan change); H-2 and H-4 accepted; **H-3 deferred per Codex** (no route, hint, grant or link
change; GetBranding strict denial audit kept, link still hidden from non-managers).

Preview note: `g017-b4` is still running (web at `3686ca02`), but its seeded users no longer sign in to their
memberships after the emulator restart; anonymous pages work. M1 is already closed from the screenshots. Restoring
it needs either an approved preview-DB re-subject or `b4.sh down` + `up`/`seed`/`ca` (new CA and users file).

Remaining for final qualification: Codex delta review of this round; MR !40 checks on the pushed head; Duo
disposition; Codex pre-integration acceptance; the integration gate (authorized conditional merge, hosted staging).

### Staging repair — correction attempt 2 of B5 (checkpoint 3) — 2026-10-17 UTC

Trigger: staging gate pipeline `2958742766` (job `17256744765`) failed on merge `3cc18228` (MR !40). Diagnosis bundle
`platform-evidence/g017/staging-gate-2958742766/repair-bundle.md` `209d8971…7127` (read-only, no attempt). Release:
staging repair review (not included) `54176506…1f11` at `origin/staging` `b8b6c380` (attempt
2 of B5, checkpoint 3; amendment 05). Rule: autonomous mechanical corrections (not included)
`c1dbd4ce…049ab`. Attempt accounting is Codex's; this round records itself as attempt 2.

- Owner/worker: Claude Code, model **Opus 5.5**, effort **high (requested)**; actual/inherited effort not verifiable from
  inside the session (unknown). Generated worktree `.claude/worktrees/agent-<id>` (tool-generated, not
  task-prefixed; recorded per CLAUDE.md). Branch `g017-staging-repair` (safekeeping push only; no MR).
- Inputs: started at `origin/staging` `c5d1b177`, rebased onto `b8b6c380` (release, 50-pattern allow-list `3307c8ef…c8f5`)
  before the first batch, then onto `61cdd431` and `8c6fdb53` (management-only advances: owner CNAME inclusion, DNS
  proposal reviews, amendment 06 allow-list 56 patterns `10dea627…2ee6` for the separate DNS writer; none of the three
  changed source files) before the push. All four changed paths match the 56-pattern allow-list. The tested source blobs equal the pushed ones (verified
  `git diff --quiet` of the tested commit against the head for every non-management path), so no check was repeated.
- Scope: three changed paths, all in the 50-pattern allow-list — `infra/validation/run.sh`, `web/tests/browser/boundary.spec.ts`
  (amendment 05) and `web/components/branding/branding-form.tsx`. `core/internal/domains/dns.go` untouched (separate DNS
  owner).

| Failed row | Correction |
|---|---|
| B-BROWSER-ROOT / -OFF (S-BOUNDARY) | `boundary.spec.ts:30–32`: the two `getByText("Dark"/"System").click()` interactions become the named `getByRole("combobox", { name: "Appearance", exact: true })` with `selectOption({ label: "Dark" })` / `({ label: "System" })` (the `appearance.spec.ts` pattern). No assertion changed. |
| B-IDENTITY-CATALOG | `run.sh`: `BRANDING_OBJECTS="i:organization_branding_pkey r:organization_branding"` next to `DOMAIN_OBJECTS`, appended to `identity_catalog()`'s want-list; PASS text "… and the 2 G017 branding relations (50 in all)". The SQL, owner check and baseline/TestIdentityCatalog conditions are unchanged. |
| B-TOGGLE-OFF start (and cascades ready, B-OBS-OFF, B-BROWSER-OFF build) | `run.sh` APPLICABLE G017 line: `web/lib/branding.ts web/app/branding web/app/organization/branding` (not the broader directory), comment updated. |
| B-IDENTITY-E2E (E-BRD-5 scrollWidth 324) | Diagnosed on the real page (below): the Reset submit. `branding-form.tsx` ResetForm: its `PendingSubmit` gets `className="max-w-full whitespace-normal text-center leading-[1.2]"` (the `block` variant's wrapping classes without `w-full`) and a one-line comment. Label, variant, behavior and the 320px assertion unchanged. |

E-BRD-5 diagnosis (domain-platform, the unchanged test plus one `console.log` inserted only in the disposable copy just
before the 320px assertion, listing elements outside the viewport or overflowing their box): reproduced
`scrollWidth 324` locally. The only element whose right edge passes `innerWidth` is the Reset submit
"Reset to platform defaults" (`button`, base `whitespace-nowrap … tracking-[1.5px]`, destructive, size default): left 20,
right **324**, width **304** in the 280px content column of `section#reset`; its ancestors (`form`, `div.mt-12`, the
section container, `main`, `body`, `html`) show `scrollWidth 304/324` from it. The bundle's 299px model was 5px short.
After the fix: `scrollWidth 320 = innerWidth`, nothing outside the viewport. Two non-page overflows remain inside the
logo section's 208px column (56–264), both ending inside the viewport and not affecting the assertion: the logo upload
form content (`scrollWidth 238`, the file input) and the "Delete this upload" outline submit (`233`). Not changed here
(not the identified cause; reported for Codex).

Focused outcomes (each batch: fetch/ancestry, `docker-desktop-mounts.sh` 13 %, under `flock local-harness.lock`,
g017-repair-* resources only; `g017-b4` and G018 untouched; evidence `platform-evidence/g017/repair/`):

| Check | Source | Outcome |
|---|---|---|
| `bash -n infra/validation/run.sh` | head | 0 |
| S-BOUNDARY, `-g S-BOUNDARY --project=shell-light --project=shell-dark` on disposable root stack `g017-repair-root` (README start/browser) | `0b2d7662` (same run.sh/boundary blobs) | **2 passed** (`p1/browser.log` `d4d26151…`) |
| Identity catalog: run.sh's exact `pg_class` query on disposable migrated `g017-repair-idn` (IDENTITY.md start+migrate, version 13) vs want-list from the repair run.sh variables | `0b2d7662` and `0178f22a` | **MATCH**, 50 = 50, owners `beginify_owner`; control: the pre-repair want-list (48) is rejected (`p2*/catalog.txt` `3dcee2bd…`). TestIdentityCatalog source unchanged (staging 4/4 retained). |
| Disabled variant: render with `intelligence_enabled=false`, `cp -a` of the 87 APPLICABLE items, the reviewed identity swaps of `.env.example`/VENDORS.md (as `variants_off`) | `0178f22a` | 358 copied files byte-identical to the root, 0 differ; the 8 new files present, `web/app/organization` holds only `branding`; services `core minio minio-init postgres redis web` |
| Disabled variant start/readiness: GETTING-STARTED `archetype:start` (builds the web image), MIGRATIONS apply, `/readyz`, `/api/health` | `0178f22a` | start **0** (no module-not-found; `p3/off-start.log` `a08d442b…`), migrate **0** `current_version 13`, core `/readyz` **200** with no intelligence entry, web `/api/health` **200** |
| Static import closure of the variant's web sources (`repair/variant-imports.sh`) | staging vs head run.sh | control **6 unresolved** `@/lib/branding` (the CI failure); head **0 unresolved** |
| E-BRD-5 diagnostic on the unfixed page | `0b2d7662` | **FAIL 324** reproduced; element named above (`p2-run1/diag1/browser.log` `bf6ee7b9…`) |
| E-BRD-5 only, domain-platform, fixed build (fresh stack): with the diagnostic log, then the unmodified test | `0178f22a` | **1 passed** each (`p2/fixed-diag/browser.log` `07699434…`, `p2/fixed-clean/browser.log` `41621d81…`) |

Attempted approaches: a first fix run rebuilt web with `up --build web` (without `--no-deps`) on the running identity
stack, which recreated `identity-setup`/emulator (new user IDs) and left the in-namespace proxy stale; those two runs
(`fix1` connection refused, `fix2` carol's link absent) are environment results, not test outcomes. The stack was torn
down and the fixed source run on a fresh stack. Cleanup: 0 g017-repair containers, volumes, networks or images; copies
removed. Not run (per release): root/migrations/off/ops umbrella suites, web lint (the two web builds compiled and
type-checked; MR fast lint covers it).

Plan wording reconciliation: Plan r2's harness rows ("two rows and count pins only") now include the identity inventory
set and total and the three web APPLICABLE paths; both are the mechanical parity items named in the release.

Remaining: Codex source/security review of this round; the DNS correction (separate owner); MR and the required hosted
qualification of the combined repaired source.

### Staging repair attempt 2: DNS hop-bound correction and combined branch — 2026-10-17 UTC

Same correction attempt 2 of B5, checkpoint 3 (no reset). Integration worker: Claude Code, model **Opus 5.5**, effort
**high (requested)**, actual/inherited effort unknown; generated worktree `.claude/worktrees/agent-<id>`
(tool-generated, not task-prefixed). DNS writer: Opus 5.5, worktree `agent-<id>`, branch `g017-dns-hopbound`.

Bindings (verified by `git show origin/staging:…` / `sha256sum` at `origin/staging` `7ac4484c`):
DNS RDLENGTH review (not included) `ad135390…5811`;
DNS proposal review r1 (not included) `6ed811f3…00e1`;
staging repair review (not included) (mechanical source ACCEPT at `0a0fc999`);
56-pattern allow-list `G017-scope-allowlist.txt` `10dea627…2ee6`; this Result as imported `534bf285…0799`.
Sources: mechanical repair `origin/g017-staging-repair` `0a0fc999` (ACCEPTED); DNS `origin/g017-dns-hopbound`
`0099f8e6` (source ACCEPTED). Writer evidence `platform-evidence/g017/dns-writer/EVIDENCE.md` `073a8b2d…7929`.

Authority and native refusal history: the owner included the G016 CNAME repair in this effort and natively approved
the DNS work on 2026-10-17 in the coordinator session. An auto-mode classifier denial was resolved by the owner's
allow rule (exact `flock …/local-harness.lock docker run …` form). Claude Code's worktree-isolation guard then held the
focused `go test` command; it was not reworded or routed around. The owner ran the tests natively instead.

Module change (`module-diff-after-tidy.patch` `ba506597…1ab6`, accepted by Codex r1 and the checksum relay):
`core/go.mod` + direct `golang.org/x/net v0.61.0`; indirect `x/sync` 0.22.0→**0.24.0**, `x/sys` 0.47.0→**0.49.0**,
`x/text` 0.41.0→**0.43.0**. `core/go.sum`: x/net v0.61.0 added; x/sync/x/sys/x/text lines replaced; `x/tools` v0.48.0 sums
replaced by v0.51.0 (x/text 0.43.0's upstream graph). No x/crypto/x/term lines.

DNS commits (rebased onto staging `c6ef2411`): `3fa25060` typed immediate-CNAME-edge lookup restores the 3-hop bound
(`dnswire.go` new, `dns.go` SystemChecker, `dns_test.go` constructor/comment only, tests, module, docs);
`4b6ebec3` rejects records whose RDLENGTH runs past the message (parser skip pass over all sections);
`0099f8e6` requires a CNAME name to occupy exactly its declared RDLENGTH (`AnswerHeader` + `UnknownResource` +
`nameExtent`; compression pointer = 2 bytes; decoding stays with dnsmessage) with five malformed-table negatives and a
compressed positive on UDP and TCP. File SHA-256 at `0099f8e6` (= combined branch): `dns.go` `c55c5a1e…2d74`,
`dns_test.go` `7c0ff14b…b563`, `dnswire.go` `c1e406ce…00cb`, `dnswire_test.go` `dd5d0af4…efd`, `go.mod` `8bb77b7c…ede7`,
`go.sum` `805f4e15…f61`.
Checks (writer, under the lock): `gofmt -l ./internal/domains` clean (after two alignment fixes); `go vet` 0;
`go vet -tags integration` 0; core `check` image build with `go mod verify` → "all modules verified", `go build` 0.

Focused tests (owner-run, identical script `run-owner-test.sh` `926ae26c…3a1d`:
`go test -count=1 -v -run '^TestDNS' ./internal/domains/`, `--network none`, `CGO_ENABLED=0`):

| Run | Source | Log (SHA-256) | Outcome |
|---|---|---|---|
| 1 | `f54f7461` | `logs/owner-test-run1-f54f7461.log` `6012812e…c7bcc4` | **13 PASS / 1 FAIL of 14**; FAIL `TestDNSWireMalformed` "RDLENGTH beyond the message" (UDP and TCP). Corrects the earlier "15/16" relay figure. |
| 2 | `0099f8e6` (all four DNS files match by SHA-256) | `logs/owner-test-20261017T215603Z.log` `2021992d…d733da` | **14/14 PASS**, `ok core/internal/domains 3.043s` |

Codex accepted the run-2 focused DNS evidence (review section "Focused DNS source/evidence accepted", same review file,
now `e986fbe9…73f9` at `origin/staging` `1c5f04b4`, a management-only advance merged into this branch; allow-list and
imported Result unchanged).

Not run: `-race` (no cgo toolchain in the image); B-DEPSCAN (hosted, in the MR/staging qualification).

Documentation clarification (Codex-required at the next doc checkpoint; no code or test change):
`docs/foundation/DOMAINS.md` item 2 and the `docs/SECURITY.md` core→DNS row now state that every declared record is
framed and parsed, exact content-length validation applies only to answer CNAME evidence, other RR types keep the
upstream parser's semantic-length limitations and cannot supply the owner-CNAME edge, and bytes after the counted
sections stay unobservable. No general RFC validator is claimed.

Combined branch `g017-staging-repair` (plain merges, no rebase, both repairs keep their commits): `0a0fc999` +
merge `origin/staging` `7ac4484c` (`2b8f6981`, no conflicts) + merge `origin/g017-dns-hopbound` `0099f8e6`
(`f725056e`, no conflicts) + the doc clarification and this Result section + merge `origin/staging` `1c5f04b4`
(management-only). Core files are byte-identical to
`0099f8e6` (`git diff 0099f8e6 HEAD -- core/` empty). Diff against `origin/staging`: `core/go.mod`, `core/go.sum`,
`core/internal/domains/{dns,dns_test,dnswire,dnswire_test}.go`, `docs/SECURITY.md`, `docs/foundation/DOMAINS.md`,
`infra/validation/run.sh`, `web/components/branding/branding-form.tsx`, `web/tests/browser/boundary.spec.ts` and this
Result; all match the 56-pattern allow-list. The integration worker ran no tests. Cleanup: DNS writer volume `g017dns-go`
and image `g017dns-core-check:local` are removed after the push (outcome in the integration handback).

Remaining: Codex review of the combined branch; the coordinator-opened correction MR (Duo, B-DEPSCAN) and hosted staging
qualification of the combined source; G017 final ACCEPT.

## Integration, qualification and closeout (2026-10-17 UTC; closeout worker, after the passing staging gate)

**Status: ready for Codex final ACCEPT. Not self-accepted.** Every section above is preserved byte-exact as history;
this section records the integration, the hosted qualification, the budget and the mandatory cleanup.

- **Worker:** Claude Code, model Opus 5.5 (`claude-opus-5-5`), effort high (requested; actual/inherited unknown).
  Generated worktree `.claude/worktrees/agent-<id>` (tool-generated, not task-prefixed); branch
  `g017-closeout` from `origin/staging` `fad992f07df8ec0f80323850cac0b455ad215e3a`. Sole writer. No tests, no
  pipelines, no implementation change.
- **Bindings verified before writing** (`sha256sum` at `origin/staging` `fad992f0`):
  `PROCESS-mandatory-task-cleanup-20261017.md` `13af3f377f62807bcf2de5f9e851624143c647a47020cf09c1b5468d42e73eaf`;
  `G017-staging-repair-codex-review.md` `e3c317813abde32994e4cff1c6f3efde71ff70ceca09e927b010c24d4bf4645a`;
  `PROCESS-repository-local-storage-20261017.md` `ff3cf18c8040d174100e2e483e71017902b461f6a8724b70f40756c5a9a29ca1`
  (G017 keeps its home-based evidence paths; nothing migrated); this Result before the closeout
  `0171cc386dc2f5961698a791041527695be9b7a844903d6019adb9c8b42a6536` (= the Codex-imported bytes).

### Integration (verified with `glab api` and Git)

- **MR !42** ("G017: staging-gate repair and G016 DNS hop-bound fix", `g017-staging-repair` → `staging`): **merged**
  at `2026-10-17T22:15:44Z` as merge commit `a33029dd3f50a51ff8158a17a93d35d8d9d8fbc0` (parents `7170d02a`, `173d0511`);
  source-branch removal was set and the remote branch is gone.
- **Candidate** `173d0511bc40ae7e27d7ac2ba8d1a1d6a3dceeb9` = the Codex-accepted `0ead89db8e7d81f153c77d9751f2110a8b69c6a7`
  plus a merge of `origin/staging` `22047800`. The diff `0ead89db..173d0511` touches only
  `management/SUPERVISOR-HANDOFF.md`, `management/v0/PROGRESS.md` and `G017-staging-repair-codex-review.md`
  (management-only reconciliation).

### Hosted qualification

- MR pipeline `2959174383` on `0ead89db`: **success**. `validate:checks` job `17259302915` (≈523 s): `B-DEPSCAN`,
  `B-UNIT-CHECKS` and `B-REQUIRED-CHECKS` PASS. `validate:integration` manual (optional; not run).
- MR pipeline `2959183996` on `173d0511`: **success** (`validate:checks` job `17259353163`, ≈534 s; integration manual).
- **GitLab Duo** on MR !42 (note `4036140321`): 0 findings. One advisory: tight timing windows in `dnswire_test.go`
  (for example 400–900 ms and ≥550 ms) could flake on a loaded runner; not a correctness issue. **Coordinator
  disposition: advisory, no change;** any future hosted flake there is a mechanical test-timing correction.
- **Post-merge staging gate: pipeline `2959201147` on `a33029dd`: SUCCESS.** `validate:checks` job `17259440630`
  518 s; `validate:integration` job `17259440631` 3245 s. `results.tsv`: **84 PASS / 0 FAIL** (plus 6 INFO), saved at
  `<home>/.local/state/platform-evidence/g017/staging-gate-2959201147/results.tsv`, SHA-256
  `422ce901af11d1236ebc65cc37352a336a4470debe52617c2f0bcb5d8f3430c0` (copy in `g017-closeout/proof/`, hash-identical).
- **Failure history retained:** the original staging gate `2958742766` FAILED; its diagnosis bundle
  `staging-gate-2958742766/repair-bundle.md` SHA-256 `209d8971c642afcd8047920c8eb71acec199364dc0126191d3f90e9f7e411127`.
  The repair and DNS correction are recorded in the attempt-2 sections above; this pass does not rewrite them.

### Budget

- **Correction attempt 2 of B5, checkpoint 3.** No reset; this closeout spends no attempt. Codex owns the accounting.

### Mandatory cleanup (PROCESS rule)

- **Receipt:** `<home>/.local/state/platform-evidence/g017-closeout/RECEIPT.md`, SHA-256
  `046788eb221b391d8ee7540c5bba1a4e0418d2f488441aab321d87552ce78e52`.
- **Removed:**
  - Docker: the stopped preview `g017-b4` via `flock …/local-harness.lock docker compose -p g017-b4 down --volumes`
    (11 containers, 3 volumes, 3 networks), then its 9 `g017-b4-*:latest` images by exact name with non-force
    `docker image rm` (≈6.4 GB). Afterwards no g017-named container, volume, network or image remains.
  - Remote branch `g017-dns-hopbound` `0099f8e6` (ancestor of `origin/staging`).
  - 11 local branches with `branch -d`: `g017-dns-hopbound`, `g017-impl`, `g017-impl3`…`g017-impl8`,
    `g017-organization-branding-logged-out-shell`, `g017-staging-repair`, `g017-staging-repair-combined`.
  - Disposable evidence: 297 files under `platform-evidence/g017/` plus the five provisional Plan/allow-list files and
    `g017-scripts/` (≈3.4 MB); manifest with sizes in the receipt.
- **Bundled, then deleted locally with `-D`** (unique commits; bundles verified, in `g017-closeout/`): `g017-r0`
  `ace1b302`, `g017-plan` `61d6ff66`, `g017-plan-r2` `18873f74`.
- **Kept:**
  - Remote `g017-r0` `ace1b302`: not contained in `origin/staging`; kept and bundled.
  - The 14 required proof files, copied to `g017-closeout/proof/` with matching SHA-256 (originals still in `g017/`).
  - `/tmp/*g017*` scratch (outside the assigned scope; listed in the receipt for disposal).
  - This worker's worktree until its final removal step.
  - All G018/G019/non-project resources, the coordinator checkout and `worktree-cleanup-20261017/`.

**Hand-back:** returned to Codex for exactly one outcome (ACCEPT, CORRECTION LOOP or ARCHITECTURE DECISION REQUIRED).
G018 integration still awaits final G017 ACCEPT and its own gates; G019 execution stays held; stop before G020 planning.
