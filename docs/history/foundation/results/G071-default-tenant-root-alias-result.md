> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G071 Result — optional default-tenant root alias (candidate, pre-MR)

Coordinator session 580cdb2f (Sonnet 5.5). Not accepted; Codex owns acceptance. Budget B3, attempt 1.
Plan: `plans/G071-default-tenant-root-alias-plan.md` r4 (+ r4.1 annotations, commit `39456e70`). Base S0 `6d94f5bc8b6566535ff82d2ec852afe884e8770b`.
Validated head `a45277cc7a328ce5755f68ce12491edde58465f0` (18 commits, 42 files). After the RC17 harness repair (MR !111,
merged into staging `acc3f7d4`) the branch was rebased onto that staging tip: code head `0dae5d7d` (the same 18 commits
plus the Plan r4.2/Result docs commit). The rebased tree differs from the validated tree only by the staging delta
(`infra/validation/lib.sh`, `infra/ci/validate-controls.sh`, `docs/foundation/OPERATIONS.md`, `management/v0/PROGRESS.md`,
Spec addendum); no test was launched for the rebase, per the owner's sequence.

## Delivered (by Plan slice)

- Core: migration `00022` (default-tenant reader, version-22 pins, build-tagged integration pins), runtime resolution,
  `root_page` on every `public_entry` hosts/resolve answer, startup records catalogued (`8acb8198`, `6247ae37`,
  `64068616`, `7bb1c0f6`, `d2ed314b`, `c3158720`, `eaf597e6`).
- Web: aliased platform root gate and E-DT-1, `root_page` consumption, exact-root redirect for `signup_redirect`,
  E-DT-2/3/5/6 (`e8877908`, `d6e65aca`, `01ada6aa`, `a1f9693e`, `e9948e70`).
- Configuration and docs: `.env.example` domain block (S11(i)), README S20 section, DOMAINS / MIGRATIONS /
  DATABASE-ISOLATION / IDENTITY docs (`e36f026a`, `25e1c471`).
- AC8 (G057 observation path hardening, deferred Duo finding on !110): `stgPathAllowed` rejects dot segments and
  backslashes (`33425d86`).
- Pins: `run.sh` pins for `main.go`, the domain block and the identity-domain tail recomputed (`f4a947e4`).

## Validation (actual outcomes; every run on a frozen scratch checkout; evidence under `.local/evidence/g071-a1/`)

| Run | Head | Scope | Outcome |
|---|---|---|---|
| unit-a1 | `e8877908` | U1–U4, vet, U8 | PASS |
| r1-1 | pre-fix | R1 | FAIL (missing environment only; kept as failure history) |
| r1-2 | `d2ed314b` | U9, U6, R1 CT-DT-1..3, TestMigrationLifecycle, TestPlatformLabelsCatalog | PASS |
| u10-1 | `33425d86` | TestStagingDomainObservationFixtures, vet, gofmt | PASS |
| b1-1 | `33425d86` | B1 phase 1 (E-DT-1) | PASS; no edits under identity/auth/emulator paths |
| unit-b-1 / unit-b-2 | `25e1c471` / `a45277cc` | U3, U6, vet, gofmt, web check (lint, typecheck, 379+ unit tests incl. new redirect cells) | PASS |
| D1 | `g071` vs S0 | README `platform:*` 12, IDENTITY `identity:*` 6, MIGRATIONS `migrations:*` 4 marker blocks byte-identical; parity listing consistent | PASS |
| r2r3-1 | `f4a947e4` | R2 `TestSignupContinuation`, `TestCustomDomainRelay`; R3 (a) exit 1 `default_tenant_unavailable`, (b)/(c) exit 0 with `invalid_slug` / `invalid_value` records, `logcheck startup`, canaries absent, 0 leftovers | PASS |
| b1-2 | `f4a947e4` | B1 phases 1 2 3 4 5 0 | phases 2, 3, 0 PASS; **phase 1 FAIL** (E-DT-2 S-ERRORS console 404s), **phases 4, 5 FAIL** (E-DT-5/6: HTTP 500, web log `TypeError: Invalid URL … ERR_INVALID_URL input '/signup'`) |
| b1-3 | `663343e4` | phase 1 | FAIL again (www landing misdiagnosis) |
| b1-4 | diagnostic commit (throwaway) | phases 1 4 5 | evidence only: console 404s are the auth host's font files; E-DT-5/6 GET body is the absolute target URL |
| b1-5 | `a45277cc` | phases 1 4 5 | all PASS (E-DT-2, E-DT-5, E-DT-6) |
| l1-1 | `a45277cc` | `lint lint-base lint-core` against S0 with an allow-file from the exact changed paths | PASS (B-ENV-EXAMPLE, B-MIGRATE-SET, B-COMPOSE, B-MCP-SOURCE, B-IDENTITY-GUARDS, B-SECRET-CANARY, B-SHELL, B-LITERALS, B-CI-CONFIG, B-SCOPE 42 paths) |

B1 phases 2, 3 and 0 reuse b1-2 at `f4a947e4` (unaffected files); the later commits changed only `web/proxy.ts`,
`web/lib/host.ts` and its test, `web/tests/browser/default-tenant.spec.ts` and docs. No ci-controls and no full local suite
ran, as the Plan and owner direction require. The full hosted suite is the pre-merge gate of the G071 MR.

## First failure and Plan r4.1 (preserved)

The b1-2 failure is kept. Cause: Next.js 16.4.1's proxy rejects a relative `Location` for the exact-root redirect. Codex
decided (handoff 7986f9b4/61e16ddb; Spec addendum on staging `16fcc483`): an absolute same-origin Location built from the
core-resolved `signupOrigin` (not request-derived), proxy gate kept, 307, no-store, fixed `/signup`, no query or cookie,
exact-root GET/HEAD, alias precedence. Plan S17/U5/AC7/D-9 were annotated `[r4.1 …]` (`39456e70`) and the original
500 evidence retained; only the failed and affected root cases were retested.

## Deviations and corrections to record

1. The coordinator performed W2-C's mechanical `run.sh` pin lane (`f4a947e4`) instead of dispatching W2-C.
2. Two E-DT test assertions were corrected (`a45277cc`, Plan r4.1 A3, mechanical, no product change):
   E-DT-2 reads the auth host's `/sign-in` on its own page (that host serves four font files as the pre-existing fixed
   404, which tripped the shared page's S-ERRORS guard); E-DT-5/6 now accept a GET body that is empty or exactly the
   Location (Next's proxy adapter writes the target URL as the GET body), while HEAD stays empty. Status, Location,
   no-store and no-cookie assertions are unchanged. Codex accepted this (source review at `0b284596`): the pinned Next
   adapter exposes only the fixed Location text as the GET body; the application constructs a bodyless `Response`, and
   HEAD remains empty.
3. A www-side-page variant (`663343e4`) was a misdiagnosis and was replaced by `a45277cc`; both stay in history.
4. Diagnostic commit and throwaway branch are not part of the candidate (deleted at closeout).
5. Known pre-existing item left out of scope: the auth host's sign-in page font 404s.

## Hosted qualification of MR !112 (record as of the head below; the final outcome is appended in the closeout receipt)

- Pipeline 2992651907 on `547cbfd0`: `validate:checks` success; the played `validate:integration` was superseded by a
  push and canceled at about 4 minutes (not a pass).
- Duo review found a real defect: the 21 to 22 pin bump had dropped the space or separator in ten harness PASS
  messages (`version 22on` and similar). Fixed in `3d0e23f0`, then finished in `e5be46a3` (six `onthe` and one
  `target =$RD_TP` left by the first correction), each verified by comparing every `version 21` context of the base
  file with the current one.
- Pipeline 2992703035 on `3d0e23f0`: `validate:checks` and the played `validate:integration` both success; executed
  profile `full` (selection reason `foundation`), 97 PASS, 0 FAIL, ci-controls 530 passed (VB8 included),
  B-MAINTENANCE-DRILL PASS; evidence `.local/evidence/g071-a1/hosted-2992703035/`. The later `e5be46a3` is a text-only
  correction to ten messages, so the full gate is rerun on it (a new pipeline) before any merge.

## Pending

- (Done) Rebase onto staging with the RC17 repair (MR !111) and binding reconciliation; no rerun was needed.
- (Done) Codex source review ACCEPTED for MR creation and hosted qualification at `0b284596` (tree `80de7c04`): no
  blocking source or security findings (`results/G071-source-review-codex-review-20261027.md`). The G069 final ACCEPT
  still gates G071 merge, integration and final acceptance; it does not gate opening the MR or the hosted gate.
- MR to staging: hosted fast, then the manual `validate:integration` job played and its executed profile verified to be
  the full profile and passing (it selects `ci-controls` because the change includes `infra/validation`); repairs go
  directly in this MR. No auto-merge until that gate and the G069 hold are clear.
- Phase E after merge only as the Plan requires (frozen merged SHA, D-10 invocation); then closeout cleanup (lane
  branches, images `g071*`, scratch worktrees) and RC18 under the RC-series authority after G071 merges.

## Closeout receipt (append-only; written after the merge of MR !112)

Coordinator session 580cdb2f (Sonnet 5.5). Not accepted; Codex owns acceptance.

**Final hosted outcome (MR !112, head `326d90d7b5f26031f7dd9276f7911f9784a94592`, tree `6971ce5f`).** Pipeline 2992900450:
`validate:checks` (job 17492538466) success; the played `validate:integration` (job 17492538467) success, ended
07:23:08. Downloaded artifacts (`.local/evidence/g071-a1/hosted-2992900450/`): `executed_profile` full, `selection_reason`
foundation, comparison base `0d86b545`, executed suites `controls ci-controls root migrations ops rename off`,
`required_count` 97, results 97 PASS / 0 FAIL / 7 INFO, ci-controls 530 passed 0 failed (VB8 included),
B-CI-CONTROLS and B-MAINTENANCE-DRILL PASS, no stragglers, no harness stop. The earlier pipelines on `547cbfd0` and
`3d0e23f0` are recorded above; only the last head is the pre-merge gate (the later commits changed message text and
this Result only).

**Review.** Duo's three spacing findings (ten messages, then six `onthe` plus `target =$RD_TP`) were fixed and
resolved; the final review of `326d90d7` found nothing and approved; 0 unresolved discussions. The G069 final ACCEPT
(staging `874d0e8b`) lifted the predecessor hold before the merge.

**Merge.** `glab mr merge 112 --sha 326d90d7…` (merge commit `37bde74a3aca8826fa7b22a1d59b6615ad60773d` on staging,
source branch removed, no squash). Staging was then `874d0e8b` plus the merge; the only delta from the tested head was
management-only documentation.

**Phase E disposition (for Codex).** The Plan §5.4 D-10 post-merge row (`root migrations ops rename off` plus the lint
suites on a frozen merged SHA) was not run locally. The hosted full gate above executed `root migrations ops rename off`
and `ci-controls` on the exact MR head, and the later staging delta is management-only; the owner direction for this
sequence is no further full or post-merge run. Reported as an unexecuted local row, not as local proof; Codex
dispositions whether the hosted evidence satisfies it.

**RC18.** Cut as a tag only from the staging tip at cut time, per the owner-directed RC18 instruction; no promotion
branch, MR or hosted qualification, no main promotion (see the RC18 tag message and the receipt line handed to Codex).

**Cleanup (pending final ACCEPT).** Kept until acceptance: lane branches `g071-w1-p0`, `g071-w2-a`, `g071-w3-p0`,
`g071-w4-a`, `g071-w2-b`, `g071-w1-p1`, `g071-w3-p1`, `g071-w4-b`, `g071-plan-r4`, throwaway `g071-diag-tmp`, this
`g071-closeout`, `g071*` images and the evidence under `.local/evidence/g071-a1/`.

## Phase E disposition and cleanup receipt (append-only; supersedes the "Cleanup (pending final ACCEPT)" paragraph above)

**D-10 (Codex disposition, handoff 155296d8).** The Plan §5.4 D-10 local invocation is **UNEXECUTED** and explicitly
accepted by hosted substitution: pipeline 2992900450 passed `validate:checks` (lint, lint-base, lint-core) and the full
`validate:integration` (root, migrations, ops, rename, off, ci-controls) on MR head `326d90d7` (tree `6971ce5f`); merge
`37bde74a` (tree `9aa879ec`) differs from that head only by three G069 management files. It is not a local pass and no
new run was made.

**Cleanup (coordinator 580cdb2f, after the merge; no test run).**

| Category | Result |
|---|---|
| Containers / networks / volumes | G071-labelled: 0 found at inventory and 0 after; none created outside the private runs, all reaped at each run's end. The 15 remaining containers belong to unrelated projects (payments-assistant, needknow, outreacher), not touched. |
| Images | 0 `g071*` images; 0 dangling. |
| Worktrees | none left (frozen scratch checkouts and lane worktrees were removed earlier); `git worktree list` shows only the main checkout, this coordinator checkout, `g070-integ` and `m1-preview-20261024` (retained). |
| Local branches | deleted 11 after bundling: `g071`, `g071-diag-tmp`, `g071-plan-r4`, `g071-w1-p0`, `g071-w1-p1`, `g071-w2-a`, `g071-w2-b`, `g071-w3-p0`, `g071-w3-p1`, `g071-w4-a`, `g071-w4-b`. The eight lane branches are patch-equivalent to staging (`git cherry`: 0 unique); `g071-diag-tmp` (1 diagnostic commit) and `g071-plan-r4` (2) hold unique history, preserved in the bundle. |
| Bundle | `.local/evidence/g071-archive/g071-local-branches-20261027.bundle`, verified, sha256 `39a223769d4c3ae39ca9394ab7f4e07af195d262a07716fe50845e27efbddd1e`; refs list `refs-at-bundle.txt` (also contains `origin/g071-plan`). |
| Remote refs | `origin/g071` removed by the merge. Still present: `g071-plan` (4 unique plan commits), `g071-plan-r4` (2) and `g071-closeout` (this receipt). Left for Codex to delete after import and verification of their tips against the bundle; not deleted by the coordinator. |
| Evidence | `.local/evidence/g071-a1` reduced from 12 MB to 2.1 MB: kept `inventory.tsv`, `rc18-receipt.env`, all per-run evidence including the preserved failures (r1-1, b1-2, b1-3, b1-4) and, for both hosted full-gate pipelines, `job.json`, `summary.json` and `results.tsv` (the downloaded artifact archives and extractions were removed; the job artifacts remain on GitLab). `.local/evidence/g071-plan` (12 KB) kept. `.local/secrets`, `.local/tmp`: no G071 content. |
| Not touched | G070 (`g070-integ`, `g070-w1/w2/w3/w4/r2`, image `g070a1-edge:p0-r1`), `m1-preview-20261024`, the 118 older `worktree-agent-*` branches, `g064-*` branches and `refs/cleanup/*` (held for owner approval), RC17/RC16 evidence. |
| Unresolved | none for G071 beyond the remote refs named above. |

RC18 remains tag-only and unqualified (tag `0.2.0-rc.18`, object `fd0d0554`, peel `37bde74a`).

## Post-import cleanup and Codex final ACCEPT — 2026-10-27

This append supersedes the earlier closeout table's pending-final-ACCEPT and
remote-ref statements.

**Result import and remote refs.** The closeout receipt at `866e0bc5` is preserved
on staging by publication commit `307e23bd`. After verifying the archival
bundles, the remote refs `g071-plan`, `g071-plan-r4` and `g071-closeout` were
deleted. `origin/g071` had already been removed by MR merge. No remote G071 refs
remain.

**Supplemental preservation.** The final closeout branch tip and the two remote
Plan tips were preserved before ref deletion in
`.local/evidence/g071-archive/g071-closeout-refs-20261027.bundle`, SHA-256
`5804b4cbae3daa5d7d32422489f95d8c0887d0fcc2d9c01acf63dcaa5a17ee64`; `git
bundle verify` passed and listed all three refs. The earlier branch-history
bundle remains at SHA-256
`39a223769d4c3ae39ca9394ab7f4e07af195d262a07716fe50845e27efbddd1e`.

**Remaining named retention.** The clean local `g071-closeout` branch is still
checked out by the live Claude coordinator worktree
`.claude/worktrees/0-agent-5c6f436` (session `580cdb2f`). It is retained as the
active shared coordinator checkout while G070 resumes; release it at the next
safe coordinator checkout handoff. This is active coordination state, not a
dangling task worktree. The separate G070 implementation worktree and the
owner-retained M1 preview worktree were not touched. G071 Docker objects,
`g071*` images, temporary files and secrets are absent. The 2.1 MB G071 evidence
set remains for the final hosted proof and preserved failure history; retain it
with this historical Result.

**Spec acceptance mapping.** The reported results and final hosted gate satisfy
the accepted Plan's mapping:

| Spec criterion | Evidence and disposition |
|---|---|
| AC1 — configuration | U1–U4/U8, R1–R3, B1 phases 0–3; recorded in `unit-a1`, `r1-2`, `r2r3-1`, `b1-5`, and the final hosted `root`/`migrations` suites. |
| AC2 — usable root alias | B1 E-DT-1/E-DT-2 and final hosted `B-BROWSER-ROOT` plus identity checks. |
| AC3 — route boundary | U3/U5, B1 E-DT-2/E-DT-4 and R2; final hosted browser/root suites pass. Signup APIs and permissions were unchanged as recorded in the Plan's source inventory. |
| AC4 — isolation and switching | U6, R2, B1 E-DT-3/E-DT-4; final hosted identity E2E/privacy checks pass. |
| AC5 — least privilege and operation | Codex SQL/source review, R1 CT-DT-1..3, U8/R3 log-canary proof, L1 and generic deployment docs; hosted `migrations` passes. |
| AC5 — README carry-through | D1 block invariance/parity, L1 B-LITERALS/B-CI-CONFIG/B-SCOPE, hosted fast unit check; behavior claims map to the B1/U/R evidence above. |
| AC6 — delivery and cleanup | MR !112 merged after hosted fast/full success and Duo approval; D-10 accepted by equivalent hosted proof; cleanup receipt and ref verification are recorded here. |
| AC7 — multi-tenant root presentation | U1–U3/U5, B1 E-DT-3/E-DT-5/E-DT-6 and the final hosted `root`/browser suites. |
| AC8 — staging observation path | U10 `TestStagingDomainObservationFixtures`, `go vet` and `gofmt` passed in `u10-1`; no production/runtime file was changed. |

**Codex final disposition: ACCEPT.** The required hosted fast/review/merge gates,
focused implementation evidence, accepted D-10 substitution and all-category
cleanup are complete. D-10's local invocation remains **UNEXECUTED**; the hosted
checks and full integration job exercised its named checks on the identical
executable tree, so no redundant local run was needed. RC18 remains tag-only and
unqualified, with no main promotion.
