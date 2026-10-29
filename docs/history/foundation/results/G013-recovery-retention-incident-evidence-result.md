> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G013 — Recovery, retention and incident evidence: Result

## Status

**Integrated; corrected hosted staging qualification PASS; awaiting Codex final ACCEPT. Not accepted.** Codex accepted
attempt 4 (the B7 checkpoint; 4 of 7 attempts used); MR !30 merged at the exact head 8281ab7 into
`29a0da5816ee2f6e5db5650ee55e1e2143873715`, and the corrected staging pipeline 2952919063 succeeded (gate 17216727596:
77 PASS, 0 FAIL, 6 INFO; B-CI-CONTROLS 441/441; the leftover inventory holds only the baseline render cache). The
earlier staging gate failure on MR !29's merge stays recorded as history. See "Integration and hosted staging gate",
"Evidence retention" and "Closeout owed".

The status recorded at the checkpoint hand-back follows unchanged.

**Attempt 4 complete (the B7 CHECKPOINT attempt); returned to Codex for the checkpoint review; not accepted.**
Attempt 3 was accepted and merged to staging as 8b0823a (MR !29), but the staging gate (pipeline 2952592867, job
17214158089) failed at finalization with exit 14: three anonymous volumes of the recovery drill's detached psql
clients were left behind although every check passed. Attempt 4 fixes that client lifecycle in run.sh (`rd_client_rm`)
with a narrow reproduction, a three-client normal/abort lifecycle proof and a covering run whose own leftover inventory
is now empty (see "Attempt 4 correction (staging cleanup)"); branch `g013-drill-volume-cleanup`, no MR yet. Attempt 1
ended in
a Codex CORRECTION LOOP (G013-implementation-codex-review.md: C1 replay not bound to the ledger head, C2 exit-guard
cleanup under errexit, C3 F4 wording); attempt 2 fixed exactly those (see "Attempt 2 correction") and was accepted for
integration (MR !29). The Duo review of MR !29 raised D1 (fence_abort could release an operator-established fence on a
signalled `--use-fence` capture) and D2 (bounded purge head-of-line blocking); attempt 3 fixed D1 with a drill
regression and documented D2 as an accepted limitation (see "Attempt 3 correction"). The F1 identity-line digest
deviation was accepted, and the F4 availability residual was accepted by the owner.

Attempt-3 covering evidence: `run.sh migrations lint lint-base provenance` a3-02 exit 0 (53 PASS, 52 required IDs,
RD-B-INT PASS) on staging ae5aea4; after the G014 docs-only staging advance (fd21bdc, SECURITY.md merged to
66e80909…), `run.sh lint lint-base provenance` a3-03 exit 0 (22 PASS, 21 required IDs) with the code byte-identical.

Attempt-2 covering evidence: the focus drill b3-11 29/29; `run.sh migrations lint lint-base provenance` b8-01 exit 0
(52 required IDs); after the staging policy delta, `run.sh lint lint-base provenance` b8-03 exit 0 (21 required IDs);
b5-02 B-LITERALS PASS after Codex's contract-literal fix. The attempt-1 final broad run b6-02 stays recorded as failed
(104 PASS, 3 FAIL, 7 INFO; finalize exit 18); its other suites and 439 CI controls are reused with provenance, VS2/VS3
passed on a focused rerun, and `beginify-core-check:latest` remains an explicit ownership hold (unresolved provenance).

## Execution binding (Git-first)

Hash lines below deliberately carry neutral labels; the artifact paths are listed under "Bound paths".

| Label | Bound bytes |
|---|---|
| Base staging (origin/staging at fetch, 2026-10-15) | `4e53de6f952c030a591bf7e9b066f33b9d69a9d7` |
| Plan r3 (P3), reviewed source `6492a144282e21a95a69f09ea95d70b8fc721ca3` | `4c936221bf680a1d870c26ecc648fd6413d525ee5a77160effc366549c8184ab` |
| Plan r2 archive (P2) | `d659406ae20723179365cb8cc265dd18f9081b615f19378833c3b9a50eec8c9e` |
| Plan r1 archive (P1) | `7b80791b19c36a52694bf892220451835d4abbf0b8f96fd954b9052d0d8ad006` |
| Plan r0 archive (P0) | `4822a1430de442391588e9ef745c8c24d6ccc4f1fee66af965ca34d7caaaab36` |
| Plan r3 Codex review, design ACCEPT (R3) | `181e7a85d615d7c8b7e7090178f725bb5e9484224b513ae43785435570eb569c` |
| Plan r2 Codex review, historical CORRECTION LOOP (R2) | `a5d23177c99334bea9acb8d2c44f7f1ef97f6a749ca1edb5e80380a6000160da` |
| Plan r1 Codex review, historical CORRECTION LOOP (R1) | `9e37a004f307cee68af46cd7b46988225e9b9900de26d9382c1d95c15aa70d5e` |
| Plan r0 Codex review, historical CORRECTION LOOP (R0) | `fef0ab96194e3bbfe4dd1c89b5d2c32811749c72aeddb3295b0cec79e45ee949` |
| Goal (G) | `62de3c7524def3a7517c46441fbb59b828a5c80f238c58ac85ef49cece97db40` |
| Spec, scanner-safe wording correction (S) | `c43f8c9746d7b0c072e870699e064dad84b3f0c05ec5f20b0a9376aad0657ba8` |
| Definition review (SR) | `99827f7bd0b50386a2090038ab902ca4f9e7dba75cbb3fa43c82d93b171cb0b2` |
| G012 implementation review with final ACCEPT (F1) | `f76a050bd6010121b99c662ef43d29b5f95283e516e71bafa998477a815bdd57` |
| G012 final Result (F2) | `e9224e6197daa4d23ac594caaef6ab029bb1422f38a5e2887738030334b6f161` |
| G012 closeout receipt (F3) | `f9964e9b6a7fcabfb5b29a67e5be4a6e9c5a8013e8493aaeea92aaa02eb64342` |
| G012 D0 disposition (F4) | `c7dde64fe0d6dfabd141574d5283ac40ef44e3de1c4c5dd707135db1e2bc16bc` |
| G012 Plan r1 (F5) | `cd90a79099bf320fadd9aad39ef4501d1971fb339d6053a533efe69f7f70d51a` |
| G012 execution binding confirmation (F6) | `260d370d6fc56969815db73cd74c6475f9a684ccf30225cd81393369229f0026` |
| G012 merge (MR !28) | `5e1e98735f6224888889d516ef2397a0ea8dc89a` |
| SECURITY.md (D1) | `e6bb1a70ec98225cd4146d9a463d2d5f7fe1ec6f6bd54f0c931f33a1b87d555f` |
| G011 implementation review with final ACCEPT (E1) | `481afdb96a5ff91152f238233967b9026650d04311667aa2996cfb6b0bd73219` |
| G011 final Result (E2) | `e62069daa0107d82c08bf0f15f63bba47a5e0d8a39c7f3382689e1e0e3b36572` |
| G011 closeout receipt (E3) | `95e0cd46ef819e2a45c10435f376e837d052e2d828ee2578cc0b40f41940ce02` |
| G011 merge | `1af469634225e6e6527aca485b40aa55953227a8` |
| AUDIT.md (D2) | `33a9bac93b867342e81a949681e742b90406d06cdf5471f9e495e4c889c44760` |
| Migration 00008 (M8) | `b1dedf1f56f28c97a9f8c59796ff1b615c1ed79e6b8ab8f2cfb602098c597233` |
| Migration 00007 (M7) | `28f0a073631a0b10ca17602e2b5897a6c02a17e305ca795c959a89d403100d12` |
| B-SCOPE allow-list (L), 28 lines, 16 patterns | `b618436d47b610b0d1f1da53b98345943f0663618e587e914ae90bbe8f459f8d` |

**Provenance (the G012 binding-confirmation clarification applied).** Every input above except L binds to base
`4e53de6` and its path; each SHA-256 was computed at `4e53de6` and equals the value cited by the Plan header, the r3
review and the dispatch. The new allow-list (L) and this initial Result bind to the prepared candidate commit (the
direct child of `4e53de6` that adds only these two files) and their own hashes, not to the base, where neither exists.
This Result cannot carry its own hash; the hand-back reports it.

**Bound paths** (Git-first provenance, no per-file inventory):
- P3 `management/v0/plans/G013-recovery-retention-incident-evidence-plan.md`; P2, P1, P0 the `-plan-r2.md`,
  `-plan-r1.md` and `-plan-r0.md` archives beside it.
- R3 G013-plan-r3-codex-review.md (not included); R2 G013-plan-r2-codex-review.md (not included);
  R1 G013-plan-r1-codex-review.md (not included); R0 G013-plan-r0-codex-review.md (not included).
- G `management/v0/goals/G013-recovery-retention-incident-evidence.md`;
  S `management/v0/specs/G013-recovery-retention-incident-evidence-spec.md`;
  SR G013-specification-codex-review.md (not included).
- F1 [G012-implementation-codex-review.md](G012-implementation-codex-review.md);
  F2 [G012-secrets-secure-engineering-result.md](G012-secrets-secure-engineering-result.md);
  F3 G012-closeout-receipt.md (not included); F4 G012-D0-codex-disposition.md (not included);
  F5 `management/v0/plans/G012-secrets-secure-engineering-plan.md`;
  F6 G012-execution-binding-codex-confirmation.md (not included); D1 `docs/SECURITY.md`.
- E1 [G011-implementation-codex-review.md](G011-implementation-codex-review.md);
  E2 [G011-durable-audit-events-result.md](G011-durable-audit-events-result.md);
  E3 G011-closeout-receipt.md (not included); D2 `docs/foundation/AUDIT.md`;
  M8 `core/migrations/00008_audit_events.sql`; M7 `core/migrations/00007_stored_objects.sql`.
- L G013-scope-allowlist.txt (not included).
- Governance: `AGENTS.md`, `CLAUDE.md`, `management/BUDGET_POLICY.md`, `docs/ARCHITECTURE.md`, `docs/DECISIONS.md`,
  ADR-006, `docs/PERFORMANCE.md` (PG-005), `docs/foundation/*`. Implementation sources and harness: the tree at
  `4e53de6`.

The staging Plan and its three archives are byte-identical to the reviewed source (`git diff 6492a14 4e53de6` over
the four Plan files is empty). `1af4696`, `5e1e987`, `728165e` and the Plan base `df3d1b0` are ancestors of `4e53de6`.

## Binding checks (Plan §12 step 1; read-only, plain Git and grep)

**Delta since the Plan base `df3d1b0`.** Three commits, management and docs records only:
- `de11391` drafts the G014 Goal, Spec and definition review, with PG-006 in `docs/PERFORMANCE.md`;
- `deb660c` normalizes the G014 draft header formatting;
- `4e53de6` imports Plan r3 with its archives and the r3 review, and records the PG-005 note "r3 design accepted;
  owner answer still pending".

Also touched: ROADMAP (one line), PROGRESS and SUPERVISOR-HANDOFF. No core, infra, test, migration, harness,
contract, SECURITY.md, foundation-doc or G013 Goal/Spec/definition-review byte changed. This is not a G013 design
input change. It is exactly the delta the r3 review asked to reconcile mechanically at binding. PG-006 belongs to
G014 and does not touch G013.

**A-G012-n reconciliation against the accepted final G012.**
- A-G012-1: SECURITY.md at `4e53de6` equals the Plan's hash (D1). Sections §3 inventory, §5 least privilege (the
  `<db>_retention` row still reads "no consumer and no audit privilege until G013"), §6 rotation, §7 at-rest, §8
  (F-7 still open, "note for G013") and §9 handoffs are present. **Holds.**
- A-G012-2: `required-checks.tsv` has 110 non-comment rows; the header totals 110; the suite `migrations` has 27.
  Profiles from OPERATIONS.md: fast 35 (line 90), integration 74 (line 92), full 103 (line 93); the line 125 full
  pipeline reads 35 / 73 / 104. `validate-controls.sh`: line 1112 `required_count 73`, line 1115 "73 required IDs",
  line 1206 `vu_split full 35 73 104`, line 1948 `n != 73`. Line 1207 has the VU9 integration split
  `35 43 74`, which the Plan leaves unchanged. **Holds**, and it matches the §11.5 "Now" column.
- A-G012-3: B-SECRET-ROTATION is the last `ops` row; no migration after 00008 exists. **Holds.**
- A-G012-4: the braces row in `infra/security/dependency-exceptions.tsv` still `expires 2026-10-29`. **Holds.**
  The external timing risk is listed below.
- The G012 final review (F1) and final Result (F2) hashes match. The G012 closeout receipt, D0 disposition, Plan r1
  and execution binding match the Plan's prefixes.

**A-G011-n.** The migration set is 00001–00008 (eight SQL files). M8, M7 and AUDIT.md equal the Plan's hashes; the G011
review, Result and receipt match. **Holds.**

**Migration number.** `core/migrations/` holds `00001`–`00008`, `embed.go` and `migrations_test.go`. No tracked or
untracked path contains `00009`. **00009 is the next free number.**

**Paths.**
- All eight planned create targets are absent: 00009, the three new Go test files, `infra/recovery/`
  (no tracked file; directory absent), `docs/foundation/RECOVERY.md`, and the two binding files before this commit.
- All seventeen modify targets exist: the six version-9 reconciliation test files; `run.sh`, `required-checks.tsv`,
  `render-customizations.tsv` and `validate-controls.sh`; SECURITY.md and the six foundation docs.
- The Git index has one link, `.agents/skills/vestive-design` (mode 120000), and no gitlink (mode 160000).

**Version-8 pin inventory (Plan §11.3 grep, `00008|version 8|current_version.t8|_retention`).**
- Under `infra/validation` and `infra/ci` there are 38 hit lines, all in `infra/validation/run.sh`, which is in the
  allow-list.
- Core test files with hits, in the allow-list: `migrations_test.go` 6, `migrate/integration_test.go` 7,
  `migrate/audit_integration_test.go` 15, `migrate/runner_test.go` 4, `tenancy/integration_test.go` 4,
  `tenancy/access/integration_test.go` 1.
- Core test files with raw hits outside the allow-list (5), classified read-only at binding:
  - `access/objects_expected_test.go` (2): fixture UUIDs whose zero-padded suffix ends in 8, false positives.
  - `audit/audit_test.go` (2): reads the 00008 file and its catalog; 00008 stays byte-identical.
  - `access/audit_integration_test.go` (1): a "version 8" comment.
  - `access/rls_integration_test.go` (1) and `migrate/rls_integration_test.go` (6): retention-role attribute and
    refusal checks, plus a version-7 subset comment.
  - None is planned for edit. Whether any version-9 behavior moves one of them is known only at b2. Any needed edit
    is a conditional amendment for Codex review, never a silent edit (§11.3).
- Production files with hits: `migrate/runner.go`, `tenancy/store.go`, `audit/audit.go`. They are comments and the
  `RetentionRole()` name. Production Go stays unchanged (§8).

## Allow-list derivation

- The 16 EREs are copied exactly from Plan r3 §11.8 (Plan lines 988–1007, inside the `text` fence). The block from
  `# Create` to the end of the file is byte-identical to the Plan's fenced block (`diff` empty).
- A seven-line comment header names the Plan path and SHA-256, in the G012 style.
- The §11.8 conditional entries are not listed: `infra/validation/lib.sh`, reconciliation hits outside the list, and
  `core/internal/tenancy/access/attack_*_test.go`. Each needs a Codex-reviewed amendment.

Read-only self-checks over the comment-stripped patterns:
- every pattern compiles (`python3 -I` `re.compile`; `grep -E -f` exit status not 2);
- `python3 -I` and `grep -E -f` both match all 35 planned paths of Plan §8/§11.8 (example `sql/` and `testdata/` names
  included);
- they match none of 45 representative excluded paths, including:
  - the conditional `infra/validation/lib.sh` and both `attack_*_test.go` files;
  - the five out-of-list pin files above;
  - `00008` and a hypothetical `00010`;
  - `infra/ci/{lib.sh,profile-paths.tsv,select-profile.sh,validate.sh}`, compose files and `minio.Dockerfile`;
  - an off-pattern `infra/recovery` script, an upper-case or nested `sql/` file, and a `.sh` file under `sql/`;
  - `.gitlab-ci.yml`, `.env.example`, ARCHITECTURE, PERFORMANCE, DECISIONS, VENDORS;
  - PROGRESS, ROADMAP, SUPERVISOR-HANDOFF, the G013 Goal/Spec/Plan/review, a G014 review, AGENTS.md and CLAUDE.md.

This is not a scan or harness run.

**Harness binding.** Plan §11.1 standard form: `bash infra/validation/run.sh --base <target> --allow-file
management/v0/results/G013-scope-allowlist.txt --expect-hash docs/SECURITY.md=<sha> --report-dir <EVID>/<batch>
<suites>`. Batch evidence goes under `<home>/.local/state/platform-evidence/g013-impl-4e53de6/<batch>`. During
validation this Result and the allow-list stay at their committed bytes (B-PROTECTED); `docs/SECURITY.md` uses
`--expect-hash`.

## Ownership

| Item | Value |
|---|---|
| Task / role | G013 / impl (the one implementation owner of all files, Plan §12) |
| Agent id | `agent-<id>` |
| Path | `<source-checkout>` (native name under the standing owner fallback, AGENTS.md "Task-prefixed Claude worktrees"; no worktree created or switched, no rename) |
| Repository | `git@gitlab.example:<namespace>/<project>.git` (SSH) |
| Branch | `g013-recovery-retention-incident-evidence`, created from the base staging commit above (upstream to staging unset before the first push) |
| Worktree lock at dispatch | `locked claude agent agent-<id> (pid 68375 start 10178190)` |
| Model and effort | Opus 5.5 (session default); effort high at spawn, actual unknown (not verifiable from inside the worker); the clarified routing policy applies from the next round |
| File ownership | every path in the allow-list; no delegate edits files |
| Dispatch record | `<home>/.local/state/platform-evidence/g013-impl-4e53de6/dispatch-record-impl.md` (outside Git) |
| Coordinator | `0-agent-5c6f436` (dispatch intent in the same directory) |

Delegate 1 (required read-only security review): coordinator-dispatched native Agent, type Plan, effort high, read-only, no worktree,
no file edits (record `dispatch-record-sreview.md`); see "Security review (Delegate 1) and dispositions". The optional
Delegate 2 was not used.

## Execution Budget

Budget class: B7 (Goal; approved in the r3 review)
Maximum goal-level attempts: 7, checkpoint after attempt 4
Attempts used: **4** (checkpoint reached: Codex reviews before any further meaningful attempt)

### Attempt Log

1. Attempt 1 (2026-10-15, b0 after the release through b6-02): implementation, focused batches b0-b5, the Delegate-1
   review and its fix round (b3-09, b4-04, b4-05, b7-01), the provisional broad run b6-01, the F4 owner hold, the
   single final broad run b6-02 and the focused control rerun b7-02, as one attempt (AGENTS.md; routine failures and
   the review's correction loop do not each spend an increment). Ended in the Codex CORRECTION LOOP (C1-C3).
2. Attempt 2 (2026-10-15, after the Codex review on staging ecc04c7): C1-C3 fixes, the focus drill b3-10/b3-11 and
   the covering runs b5-02, b8-01, b8-02 (superseded) and b8-03. Accepted for integration; MR !29 opened.
3. Attempt 3 (2026-10-15, after the Codex CORRECTION LOOP on the MR !29 Duo findings, staging 1740149): the D1 fix
   with the RD-B-INT drill regression, the D2 documentation, the preliminary shellcheck a3-00, the gate-blocked a3-01
   (staging advanced; rebased), the covering run a3-02 and the docs-reconciliation run a3-03, as one attempt (Fable 5.1
   at high; no xhigh escalation). Accepted for integration; merged to staging as 8b0823a; the staging gate failed
   at finalization (leftover inventory).
4. Attempt 4 (2026-10-15, after the Codex CORRECTION LOOP "Failed staging cleanup and attempt-4 release", staging
   dd2f569): the drill client anonymous-volume fix (`rd_client_rm`), the reproduction and lifecycle proof a4-00, the
   covering run a4-01 with the leftover inventory counted as hosted `complete` counts it, as one attempt (Fable 5.1 at
   high; no xhigh escalation). **Checkpoint after attempt 4 reached.**

## Gates (historical; all satisfied at release)

1. PG-005: the owner answered "Yes" (bounded query and lock measurements, no upfront index); not a production SLO or
   measured performance approval.
2. Resource, budget and input release: relayed by the coordinator with the execution release.
3. Codex confirmation of the binding and explicit execution release: G013-execution-binding-codex-confirmation.md.

Every launch since ran the per-batch gate of Plan §11.1: fetch and ancestry, `bash infra/ci/docker-desktop-mounts.sh`,
the coordinator-owned harness flock, port and ownership checks, the README secret-scan line, target and candidate SHAs.

**External timing risk.** The G012 braces exception expires on 2026-10-29 (inclusive; A-G012-4, R-e). After that
date B-DEPSCAN fails every profile until a fix or a renewed owner disposition. If final validation runs after
the expiry, the lapse is reported as an external blocker for Codex/owner disposition; G013 never edits the exception.

**G014:** held for the owner's explicit release; G013 changes nothing about that hold.

## Execution release and amendment 01

- **Release:** Codex "Owner disposition and execution release", G013-execution-binding-codex-confirmation.md on staging
  e03c553 (file digest `91f15a18a4356f59ed2a804bca987e89cc7eaab09f1d5f62744e6bce708910fc`). Owner PG-005 answer "Yes":
  bounded query and lock measurements, no upfront index; not measured or production performance approval.
- **Amendment 01** (same file, staging 3bbadd6, file digest `dc85100f104d6819cfb56a601751b25b560e2b7b064eac4c08b126861633b77f`):
  `core/internal/tenancy/access/audit_integration_test.go` added to the allow-list (17 patterns, list digest
  `6f9030387c47321ba5f7c2df12e3afaa63dc66699a46ca66a38a127433c8386c`), restricted to `auditACL` and its focused control.
  Cause: at version 9 `audit.events.organization_id` carries runtime INSERT and retention SELECT items, and TestAudit's
  privilege fault (REVOKE and re-GRANT of runtime's INSERT) reorders them; the old helper compared raw aclitem text.
  Change: `audCanonACL` sorts complete entries (duplicates kept, NULL `null` versus explicit `{}`, columns keyed by
  attname) and `audCanonACLControl` (inside AT-ATOM) shows reordered equal entries equal and a changed privilege, grant
  option, lost duplicate, NULL or explicit-empty ACL unequal. No other expectation changed.
- The 16-pattern list digest `b618436d…` is historical; validation uses the 17-pattern list from staging.

## Implementation summary, files changed, migrations, contracts, tests

- **Migration** `core/migrations/00009_retention_maintenance.sql` (version 9): exact version-8 preconditions (the 00008
  inventories plus the version-8 refusal-body digest), `app.retention_eligible(organization, at, ids)` (owner-owned,
  SECURITY INVOKER, STABLE, pinned search_path, PUBLIC EXECUTE revoked, EXECUTE to retention only; raises 22023 on an
  absent or malformed `retention.scope` (`organization:<uuid>`|`global`), `retention.cutoff` (strict UTC, before the
  transaction time) or `retention.holds` (`none` or up to 256 UUIDs); no exception block), four retention-only policies
  (SELECT and DELETE on `audit.events`; SELECT and DELETE of `state = 'deleted'` rows on `app.stored_objects`), USAGE on
  `audit`, SELECT on `events(id, occurred_at, organization_id)` and `stored_objects(id, organization_id, state,
  deleted_at)`, DELETE on both, and the replaced `audit.events_refuse` body (DELETE by `<db>_retention` of an eligible
  row only). Exact version-9 postconditions including both body digests. Down restores the exact version-8 catalog with
  rows present. Implementation details recorded against the Plan: holds use the token `none` for an empty set (a reset
  placeholder then still counts as absent); the holds check is `(,uuid)*` plus a 256 cardinality bound (the first
  `{0,255}` regex was catastrophically slow on long values and hung TestRetention; replaced in b2); the relation-count
  postcondition counts 41 (38 plus the migration ledger's three relations).
- **Operator commands** `infra/recovery/recovery.sh`, `infra/recovery/lib.sh` and `infra/recovery/sql/*.sql`
  (fingerprint-data, fingerprint-catalog, fence-state, fence-revoke, fence-restore, invalidate, ledger-sync,
  replay-object-start/finish, replay-user-disable, replay-unaudited, replay-restrict, retention-audit,
  retention-objects, status-counts): fence/unfence, backup (C1-C7, manifest, COMPLETE), verify, restore (preflight
  refusals, staged continuation), invalidate-auth, ledger init/append/sync/verify/attest, replay, reopen, retention
  audit/objects, status (JSON lines and sink), quarantine, expire. `infra/recovery/testdata/policy/*.policy`: six
  synthetic refusal fixtures.
- **Go tests:** `core/internal/migrate/retention_integration_test.go` (TestRetentionCatalog 3 subtests,
  TestRetentionMigration 5), `core/internal/tenancy/access/retention_integration_test.go` (TestRetention RT-1..9 with
  E-RET-1..3), `core/internal/tenancy/access/recovery_integration_test.go` (TestRecoverySeed, TestRecoveryPost,
  TestRecoveryVerify V1-V9, TestRecoveryRepair), `core/migrations/migrations_test.go` (set through 00009, 00008 digest
  pin, TestRetentionMigrationShape), amendment 01 in `audit_integration_test.go`.
- **Version-9 reconciliation (Plan §11.3):** `migrate/integration_test.go` (lifecycle 0→9→0→9, reset drops the
  version-9 function, metadataThroughVersion9Only), `migrate/audit_integration_test.go` (`auditFS` = subset 8),
  `migrate/runner_test.go` (sources 9), `tenancy/integration_test.go` (9→8 step), `access/integration_test.go` (version 9);
  `infra/validation/run.sh` B-MIGRATE-SET (00008 pinned), root migrate block, B-MIGRATE-CLI/CATALOG steps, B-PERSIST,
  B-DBRENAME, B-TOGGLE-OFF, B-OBJ-IT/B-AUDIT-IT/B-IDENTITY-CATALOG texts.
- **Harness:** four required IDs in `migrations` (B-RETENTION-CATALOG, B-RETENTION-IT, B-RECOVERY-DRILL,
  B-MAINTENANCE-DRILL); `PORTS_REC`; B-SHELL list; `APPLICABLE` (+ infra/recovery, RECOVERY.md); the retention value in
  the IT env file; `required-checks.tsv` 114 rows / migrations 31, full 107; `validate-controls.sh` 77 and 35/77/108;
  the FOLLOW-UP customization reason.
- **Docs:** new `docs/foundation/RECOVERY.md`; SECURITY.md §3, §5, §7, §8 (F-7 closed), §9; AUDIT, OBJECT-STORAGE,
  DATABASE-ISOLATION, MIGRATIONS, OPERATIONS, FOLLOW-UP.
- **Contracts:** no public API, MCP, BFF or admin contract; no new role, vendor, image or runtime service; the 29-action
  audit catalog is unchanged. Operator interface formats: g013-backup/1, g013-fence/1, g013-ledger/1, g013-restore/1,
  g013-replay/1, g013-reopen/1, g013-maintenance/1, g013-preview/1, g013-attest/1.

## Security review (Delegate 1) and dispositions

Delegate 1 (Plan r3 §12): a native read-only Agent (type Plan, effort high, no worktree), dispatched by the coordinator
on candidate 459c694 (staging 6aca14e); record `dispatch-record-sreview.md`, report `s-review.md` (file digest
`e41fd1b727c4eb278644435a48416423e9f843d63395e92e3de1fbf43e7f1b77`), both under the batch evidence directory outside
Git. Verdict: BLOCKING-FINDINGS (1 Blocking, 1 High, 4 Medium, 10 Low). It found the migration, fence lifecycle, value
handling, harness block and amendment 01 sound. Fixes stayed inside the 17-pattern allow-list and attempt 1; no
contract or design change beyond the review's minimal fixes.

| Finding | Disposition | Evidence |
|---|---|---|
| F1 Blocking: stale invalidation/replay receipts satisfied the gate for a new target | Fixed. A first restore refuses a receipts directory holding invalidation, replay or reopen receipts (`receipts_reused`). Invalidation and replay receipts record the restore binding and the target identity (project, PostgreSQL container, system identifier); replay and reopen accept only receipts bound to the current restore (`receipt_unbound`); gaps carry forward only from a same-binding replay receipt. The binding is the digest of the restore receipt's identity lines (set, manifest digest, target project, container, volume, system identifier, start time), not of the whole receipt, because a staged object step legitimately adds lines after invalidation. | RD-R 9th refusal (reused receipts directory); RD-L 7th gate (invalidation receipt of another restore, reason `receipt_unbound`) |
| F2 High: a strict replay entry with wrong or missing references reported already_applied | Fixed. `refs_valid` (lib.sh) at append and at replay: exactly the kind's keys, canonical lowercase UUIDs (object.delete included), closed-set scope types and roles of that scope, an Organization binding scoped to its Organization; anything else is refused at append (exit 2) and replays as `conflict`. `replay-restrict.sql` returns `unlinked`/`none` only when every referenced entity exists in the checkpoint and the link or binding is absent, otherwise `absent` (conflict). | RD-S: 7 invalid-reference appends refused; RD-Y probes: wrong Group `absent`, wrong subject `absent`, applied unlink `unlinked` |
| F3: membership.downgrade not checked to be a downgrade | Fixed in `refs_valid` (owner > admin > member, same scope), at append and replay. | RD-S "downgrade upwards" refused |
| F4: retention login can hold any table lock on its two tables (availability) | **Residual accepted by the owner** ("Yes for 013 question", 2026-10-15; section "Owner risk acceptance and final-batch release" of the same Codex record on staging 31444dd, file digest `9c8309f079b512886520011cb506754fc038f9427cd24605c0a6ef546e11c68e`); the approved eligible-row DELETE boundary is retained, no broader grant, no redesign. Codex review (`G013-security-f4-codex-review.md`, first at staging c1e16de): inherent to the approved table-level DELETE grants (PostgreSQL LOCK rule), not a privilege expansion; documented acceptance recommended. 00009 unchanged; no redesign. RECOVERY.md and SECURITY.md §5 state the affected tables, the cross-Organization impact, that timeouts are no guarantee against a compromised retention login and that no separate LOCK grant exists, with blocking-session inspection (`pg_locks`/`pg_stat_activity` query), termination and rotation. No RT case added. | docs |
| F5: set -e exits skipped fence release and INCOMPLETE | Fixed. An EXIT guard (`GUARD` = fence, capture-integrated, capture-external) releases what the command established and writes INCOMPLETE on any exit that did not pass through `fence_abort`/`capture_fail`. | code review; RD-B/RD-K unchanged PASS |
| F6: fence did not check PUBLIC CONNECT | Fixed. X0 refuses when PUBLIC holds CONNECT (NULL ACL read as the type default); after X2 no non-superuser login may still pass `has_database_privilege(..., 'CONNECT')`, or the fence aborts. | RD-B/RD-K PASS with the new checks |
| F7: `--max-duration` unvalidated | Fixed: 1-7200 seconds. | code |
| F8: retention objects reported complete at max_batches | Fixed: eligible rows left at the stop give `partial` for both kinds. | MD-A/MD-O PASS |
| F9: continuation re-uploaded mismatched objects | Fixed: a continuation refuses `continuation_mismatch` when an object already on the target differs in bytes or binding (never overwritten). | RD-F foreign-object case (exit 3) |
| F10: replay receipt lacked the executing operator | Fixed: `executed_by`. | code |
| F11: expire kept manifest and COMPLETE | Fixed: after every component is gone, the manifest, COMPLETE (and INCOMPLETE/QUARANTINED) are removed; only `EXPIRED` remains. | MD-E checks both absent |
| F12: roles fingerprint missed cross-boundary memberships | Fixed: memberships where either side is a project role. | RD-V/restore equality PASS |
| F13: admin_window_end not recorded | Fixed in the fence record at release (the manifest is written before release). | code |
| F14: concurrent fence/backup not serialized | Fixed: `flock` on `<root>/fences/.lock` (`fence_busy`). | code; drill PASS |
| F15: preview/sink not private-checked; unvalidated status basename; eval | Fixed: absolute file paths in private directories; status considers only UUID-named sets; eval replaced by a case. | code |
| F16: no overall drill bound; dead line; unscoped holder filter | Fixed: 30/15-minute bounds checked before each stage; dead line removed; RD-K-1 filters on the run label. | run.sh drill block |

Focused retests after the fixes: b3-09 focus drill 29/29 PASS (with the new negatives); B-SHELL form shellcheck clean;
b4-04 found a measurement defect (fixed); b4-05 `run.sh migrations` exit 0 with both drills PASS and a clean finalize
cleanup; b7-01 VR1/VS1 PASS.

## Attempt 2 correction (Codex CORRECTION LOOP)

Review: `management/v0/results/G013-implementation-codex-review.md` (staging ecc04c7, file digest
`d1ff6fcf59e79302ea91d9ff33c0c9fa6132ef03463edd367c81803053a75f5a`) on source 8607534 and the attempt-1 Result
65914e55… It accepted the F1 identity-line digest deviation and required C1-C3 only. Model Opus 5.5 (session
default); effort high at spawn, actual unknown. The clarified routing policy (staging 9da7580: Fable 5.1 at high for
planning and test corrections) applies from the next round; this round was not switched.

| Correction | Change | Focused evidence |
|---|---|---|
| C1 replay not bound to the ledger head (AC4) | `replay` reads and validates the ledger head once at its start and records `ledger_uuid`, `ledger_through_seq` and `ledger_head` in replay.tsv; it applies only entries up to that sequence; a head that changed before the end makes the run `incomplete` with `ledger_drift` (exit 4). `reopen` requires the replay receipt's ledger identity, sequence and head to equal the current validated and attested head (`replay_stale` otherwise), on top of the backup, restore and target binding. Drill seam `RECOVERY_DRILL_PAUSE_IN_REPLAY` (inert unless set). | RD-O (b3-10, b3-11, b8-01): a membership.remove appended and attested after the replay makes reopen refuse 3 (`replay_stale`); a ledger append during a paused replay gives exit 4 (`ledger_drift`), receipt `incomplete`, reopen 3; replaying the new head exits 0, covers the attested sequence and the binding is gone; the normal reopen then passes |
| C2 exit guard inherited errexit (AC1/AC7) | `guard_exit` turns errexit off, writes the INCOMPLETE marker and log lines without letting their failure skip anything, releases exactly the fence this command established, then reports the fence's actual recorded state (released, verified, or still held: run unfence) and keeps a non-zero exit. `fence_abort` and `capture_fail` guard their marker writes the same way (`capture_fail` now releases first). Drill seam `RECOVERY_DRILL_FAIL_CAPTURE=marker` (inert unless set): a capture failure whose INCOMPLETE path is a directory. | RD-B-NEG (b3-11, b8-01): exit non-zero, "the INCOMPLETE marker could not be written", "fence … released (verified)", the fence record ends `released`, no COMPLETE, the drill removes its own set; b3-10 first failed only because the guard sent its log lines to /dev/null (fixed in 332459c) |
| C3 F4 still described as pending (AC8) | SECURITY.md §5 and RECOVERY.md now cite the owner acceptance (2026-10-15, G013-security-f4-codex-review.md at staging 31444dd); risk, scope and safeguards unchanged. SECURITY.md digest now `cab01f6a0474759cbd3211f3b19fa34c44aa7bfdde0111cee5b88922775ea58e` (was 834a4285…). | B-PROTECTED, B-SCOPE (b8-01, b8-03) |

Files changed in attempt 2: `infra/recovery/recovery.sh`, `infra/validation/run.sh` (drill cases and the PASS text),
`docs/SECURITY.md`, `docs/foundation/RECOVERY.md` (C1/C2 behaviour, the seams, the C3 wording); no migration, grant,
contract or new file.

**Attempt-2 batches** (each after `prebatch.sh`: fetch and ancestry, mounts, ports and ownership, README secret scan;
this Result at its base bytes b9012164… during every harness run):

| Batch | Target / candidate | Outcome |
|---|---|---|
| b5-02 `run.sh lint lint-base` (the dcb68a4 literal fix) | dcb68a4 / 14865f2 | exit 0: B-LITERALS PASS (4 identity values absent), B-SECRET-CANARY, B-SHELL, 16 required IDs; finalize exit 0 |
| b3-10 focus drill | 827070d / 507b33b | 28/29: RD-O (C1) PASS; RD-B-NEG FAIL (the guard's log lines were discarded, see C2) |
| b3-11 focus drill | 4f195c5 / 9caee67 (332459c rebased) | 29/29 PASS |
| b8-01 `run.sh migrations lint lint-base provenance` | 4f195c5 / 5d05d3f | **exit 0**: 53 PASS, 0 FAIL; B-RECOVERY-DRILL, B-MAINTENANCE-DRILL, B-RETENTION-CATALOG, B-RETENTION-IT, B-SHELL, B-LITERALS, B-RENDER-ID/PLAN/REPEAT, B-PROTECTED (575), B-SKILL-UNCHANGED, B-SECRET-CANARY (55), B-SCOPE (46 paths), B-REQUIRED-CHECKS (52 IDs); no E-RET-TRIGGER; finalize exit 0 |
| b8-02 `run.sh provenance` after staging moved to 9800db4 (AGENTS/CLAUDE/skills/templates delta) | 9800db4 / 5a6f995 | exit 1: B-RENDER-*, B-PROTECTED (576), B-SKILL-UNCHANGED and B-SCOPE PASS; B-SECRET-CANARY UNMET only because provenance alone generates no secrets, hence B-REQUIRED-CHECKS FAIL; superseded by b8-03 |
| b8-03 `run.sh lint lint-base provenance` | 9800db4 / 5a6f995 | **exit 0**: B-SHELL, B-LITERALS, B-RENDER-ID/PLAN/REPEAT, B-PROTECTED (576), B-SKILL-UNCHANGED (identical to 9800db4), B-SECRET-CANARY (16), B-SCOPE (46), B-REQUIRED-CHECKS (21 IDs); finalize exit 0 |

**Evidence reuse, with provenance** (no unchanged full suite and no full ci-controls rerun, as the review directed):
- The other suites of the full profile (root, ops, rename, off, controls, contracts, unit, lint-web, lint-core) come from
  b6-02 (candidate 67d99f5 on 31444dd), where they passed. Attempt 2 changed only `infra/recovery/recovery.sh`, the
  run.sh drill block (exercised only by `migrations`) and two documents, none of them inputs to those suites. The later
  staging deltas (dcb68a4 contract literal, G014 records, the AGENTS/CLAUDE/budget/skill/template policy changes) are
  covered by b5-02 and b8-03.
- B-CI-CONTROLS: 439 controls passed in b6-02; VS2 and VS3 failed there and passed in the focused rerun b7-02
  (537d124 on aae1101, exit 0); VR1 failed in b6-01 and passed in b7-01. The controls test `infra/ci` and validate.sh,
  which G013 did not change. This reuse does not show the load-sensitive storage-observation defect is fixed.
- b6-02 stays recorded as a failed run (104 PASS, 3 FAIL, 7 INFO; finalize exit 18), never relabelled green.
- **Image ownership hold:** `beginify-core-check:latest` was
  sha256:a5ab2420c8d92c1c06c66f7a3fe3786a6f38c3cb384a39c3afc417ea3d0e5150 at the start of b6-02 and
  sha256:2db75c3eba6cf56a371ca0b7baca2c82e7d18e8587c47e70205982bf1e4274d5 after it, with b6-02's finalize exit 18
  ("unproven"). The attempt-2 batches did not run the root/unit suites that build it, and their finalize runs exited 0.
  It remains an explicit ownership hold with unresolved provenance: no deletion, no capture row, no prune.

## Attempt 3 correction (MR !29 Duo D1/D2)

Review: `management/v0/results/G013-implementation-codex-review.md` section "MR !29 Duo disposition — 2026-10-15 UTC"
(staging 1740149, file digest `20248f618eb5a889d16fab9a3e9b468385bc37711551458808137bb577be7a4f`) on the MR head
4ea4cfb (fast pipeline 2952382395 passed). It reopened only the two Duo findings D1 and D2 and authorized attempt 3 of
B7. Worker: a new native isolated agent (agent-<id>; dispatch record `dispatch-record-attempt3.md` in the
evidence folder), model **Fable 5.1** (claude-fable-5-1) at effort **high** at spawn per the owner's routing policy for
test corrections (ONEOFF-workload-model-routing.md); the actual CLI setting is unknown from inside the agent. No
escalation to xhigh was needed: the defect reproduced from the source on first reading, the fix stayed bounded and the
regression passed on its first harness run. Branch handling: the MR branch stayed checked out in the idle attempt-2
worktree, so this round used the local branch `g013-attempt3` from 4ea4cfb, rebased onto staging 1740149 (586444f) and,
after the gate of a3-01 found staging advanced again, onto ae5aea4 (399d9d4; both deltas management-only), and pushed to
the MR branch with an exact-tip lease on 4ea4cfb.

| Finding | Change | Focused evidence |
|---|---|---|
| D1 (blocking; AC1/AC7, Plan §3.1/§3.2): `fence_abort` released `FENCE_FILE` unconditionally, and `cmd_backup --use-fence` installs it for TERM/INT/HUP, so a signal during a capture under an operator-established fence reopened that fence | `fence_abort` now reads the ownership state **before** clearing anything (`GUARD`, and `fence_mode`), turns errexit off, and releases only a fence this command established (`fence`, or an integrated capture). Under `--use-fence` (`GUARD=capture-external`, `fence_mode=external`) it never releases: it logs that the external fence stays held (release deferred to `unfence`), writes `INCOMPLETE` with the closed reason, and exits 4 as before. For an owned fence it reports the verified recorded state like `guard_exit` (released (verified), or still held: run unfence). `cmd_backup` now clears the signal trap right after `COMPLETE` in both modes (the external branch previously left it installed after the capture). `unfence`, `capture_fail`, `guard_exit` and the integrated release are unchanged. | New drill step **RD-B-INT** in the existing `rd_stage_rdb` (B-RECOVERY-DRILL), no new seam, suite or framework: `fence --admin-window CHG-RD-I1`, then `backup --use-fence --backup-id …d1` in the background and TERM once `database.dump` exists (the same mechanism as the existing integrated TERM negative). Asserts all of: exit non-zero (4), no `COMPLETE`, `INCOMPLETE` reason `interrupted`, the tool log says the external fence stays held, the fence record has no `released` line, the database ACL differs from the recorded prior ACL and the runtime login is refused `permission denied for database`, the MinIO service user is `disabled` and its `mc ls` refused, `status` exits 1 with `backup.fence: held`; then an explicit `unfence` exits 0 and the ACL equals the prior ACL, the runtime `SELECT 1` succeeds, the user is `enabled` and the record carries `released`. The existing integrated-capture TERM negative (CHG-RD-7 in RD-B-NEG, asserted through "unreleased fences 0") is unchanged and still passes. a3-02: RD-B-INT PASS ("exit 4, no COMPLETE, INCOMPLETE interrupted; fence held (record unreleased, runtime CONNECT refused 1, service user disabled 1 and refused 1, status held 1); unfence 0 restores the prior state (ACL equal yes, runtime login 1, user enabled 1)"), RD-B-NEG PASS unchanged ("unreleased fences 0") |
| D2 (accepted bounded limitation; Spec R5/AC6, Plan §6.5): the oldest fully skipped tombstone batch ends the invocation before younger eligible rows | `docs/foundation/RECOVERY.md` (retention section) documents the bounded head-of-line blocking: candidates are selected oldest first and a skipped row stays at the head of every later batch, so the invocation ends at the first batch that deletes nothing; the receipt is `partial` (exit non-zero) with the skip counts and the remaining count, never a claim that every eligible row was visited; an unchanged retry selects the same head and makes no progress; the operator resolves the actual blockers only through the existing paths (ledger synchronization or the missing `object.delete` entries for `not_preserved`, the authorized repeat delete through the guarded delete for `bytes_present`, provider recovery for `provider_unavailable`), then a new dry run and apply; never delete or alter blockers, erase ledger or replay evidence, or read `partial` as completion; no cursor-based (seek) paging or starvation-free traversal is claimed. No code, policy, migration or contract change. `docs/SECURITY.md` does not describe purge traversal and is unchanged (digest `cab01f6a0474759cbd3211f3b19fa34c44aa7bfdde0111cee5b88922775ea58e`). | B-PROTECTED, B-SCOPE and B-RENDER-* (a3-02); the existing skipped/partial maintenance coverage (MD-O, MD-A in B-MAINTENANCE-DRILL) is reused, as the review allowed |

The RECOVERY.md backup section also states the `--use-fence` rule in operator terms (a signal, timeout, refusal or
unexpected exit during that capture never releases the operator's fence; the record, CONNECT and the service user stay
restricted, `status` reports held, the set is `INCOMPLETE`, only `unfence` restores the prior state).
RECOVERY.md digest after attempt 3: `12670f744937ea98e62d44ca3177f88d22d6d86f3446f3512497e0ba4a7b405e` (was 4be58e1c…).

Files changed in attempt 3: `infra/recovery/recovery.sh` (`fence_abort`, the trap end in `cmd_backup`),
`infra/validation/run.sh` (RD-B-INT and the B-RECOVERY-DRILL PASS text), `docs/foundation/RECOVERY.md`, and this Result
(committed last); all inside the 17-pattern allow-list (digest 6f903038…); no migration, grant, contract, new file,
suite or seam.

**Attempt-3 batches** (each after the a3 pre-batch gate under the harness flock: fetch and ancestry, Docker Desktop
mounts, ports, ownership (the idle attempt-2 worker's `g013.batch=focus` environment listed and left alone), the README
secret scan; this Result at its base bytes 7a1e0d4f… during every run; scripts `a3-*.sh` in the evidence folder):

| Batch | Target / candidate | Outcome |
|---|---|---|
| a3-00 preliminary shellcheck (the B-SHELL flags) on run.sh, lib.sh, recovery.sh, recovery lib.sh | 1740149 / 586444f + the three edits | `-S warning` exit 0, `--include=SC2251` exit 0 |
| a3-01 `run.sh migrations lint lint-base provenance` | ae5aea4 / 586444f | gate FAIL before any run: staging had advanced to ae5aea4 (management-only) and was no longer an ancestor; rebased, nothing ran |
| a3-02 `run.sh migrations lint lint-base provenance` | ae5aea4 / 399d9d4 | **exit 0**: 53 PASS, 0 FAIL, 0 UNMET, 5 INFO; B-RECOVERY-DRILL (29 recovery steps incl. RD-B-INT), B-MAINTENANCE-DRILL (7), B-RETENTION-CATALOG 3/3 + 5/5, B-RETENTION-IT 9/9 (no E-RET-TRIGGER), B-SHELL, B-LITERALS, B-SECRET-CANARY (55), B-SCOPE (46 changed paths, all within the allow-list), B-PROTECTED (577, 1 expect-hash match), B-RENDER-ID/PLAN/REPEAT, B-SKILL-UNCHANGED, B-REQUIRED-CHECKS (52 IDs); finalize exit 0 (removed 0, unproven 0, remaining 0), 0 leftover containers |
| a3-03 `run.sh lint lint-base provenance` (docs reconciliation; `--expect-hash docs/SECURITY.md=66e80909…`) | fd21bdc / 7415180 | **exit 0**: 22 PASS, 0 FAIL, 0 UNMET, 5 INFO; B-SHELL, B-LITERALS, B-SECRET-CANARY (16), B-SCOPE (46 changed paths since fd21bdc, all within the allow-list), B-PROTECTED (579, the expect-hash 66e80909… matches), B-RENDER-ID/PLAN/REPEAT, B-SKILL-UNCHANGED (identical to fd21bdc), B-REQUIRED-CHECKS (21 IDs); finalize exit 0 (removed 0, unproven 0, remaining 0), 0 leftover containers |

**Reconciliation after a3-02 (coordinator notice).** Staging advanced to fd21bdc (the G014 owner decision B: the MCP
connection-context contract, ARCHITECTURE.md, DECISIONS.md, ADR-005, the G014 Goal/Spec/reviews, PROGRESS.md and a new
`docs/SECURITY.md` section 6.2 "Planned G014 MCP grant boundary", +20 lines; no code, migration or harness change). The
branch was rebased from 399d9d4 onto fd21bdc (7415180) without conflict: the SECURITY.md hunks merged, staging's G014
section is kept byte-for-byte and G013's F4/C3 content is unchanged. `infra/`, `core/` and `docs/foundation/` are
byte-identical to 399d9d4, so a3-02's code evidence carries over unchanged; only the docs-merge-affected checks were
rerun in a3-03 (the README secret scan in the gate, B-LITERALS, B-PROTECTED with the new expect-hash,
B-SKILL-UNCHANGED, B-SCOPE, B-RENDER-*). Hashes: staging's SECURITY.md
44549d92a2362d4657b94bb3f97ea7e6f4cc52e2a511a7b568b2ca80bed39ed1; the merged SECURITY.md
`66e80909c686de57700c754443bbf12e287fb6850548e18130e9d9d7f4b04229` (the `--expect-hash` value from now on);
RECOVERY.md 12670f74…; allow-list 6f903038… unchanged.

**Evidence reuse, with provenance** (no unchanged full local suite and no all-controls rerun, as the review directed):
- a3-02 supersedes b8-01/b8-03 for the suites it ran (`migrations`: B-RETENTION-CATALOG, B-RETENTION-IT,
  B-RECOVERY-DRILL, B-MAINTENANCE-DRILL; `lint lint-base`: B-SHELL, B-LITERALS, B-SECRET-CANARY; `provenance`:
  B-SCOPE, B-PROTECTED, B-RENDER-*, B-SKILL-UNCHANGED); a3-03 repeats the lint/lint-base/provenance suites on the
  reconciled tip. The drills were not rerun for the docs merge: their inputs are byte-identical to a3-02's.
- The other suites of the full profile (root, ops, rename, off, controls, contracts, unit, lint-web, lint-core) are
  reused from b6-02 (67d99f5 on 31444dd) as in attempt 2: attempt 3 changed only `recovery.sh`, the run.sh drill block
  (exercised only by `migrations`), RECOVERY.md and this Result, none of them inputs to those suites. The staging deltas
  since 62af67d (1740149: PROGRESS.md, the Codex review, the G014 resource-qualification record; ae5aea4: PROGRESS.md,
  the G014 provider-qualification review) are management-only and are covered by a3-02's lint and provenance suites.
- B-CI-CONTROLS: 439 of 441 passed in b6-02; VS2/VS3 passed in b7-02 and VR1 in b7-01. b6-02 and b8-02 stay recorded
  as failed runs; the load-sensitive storage observation is not claimed fixed.
- **Image ownership hold:** `beginify-core-check:latest` was sha256:2db75c3eba6cf56a371ca0b7baca2c82e7d18e8587c47e70205982bf1e4274d5 before and after a3-02 and a3-03 (unchanged since b6-02); still an explicit hold with unresolved
  provenance: no deletion, no capture row, no prune. The a3-02 finalize exit code is reported as observed.

Budget: attempt 3 of B7 used (routine gate block a3-01 and the rebase do not spend an increment); checkpoint after
attempt 4 not reached. The earlier evidence and dispositions are preserved unchanged; this section is appended only.

## Attempt 4 correction (staging cleanup)

Review: `management/v0/results/G013-implementation-codex-review.md` section "Failed staging cleanup and attempt-4
release — 2026-10-15 UTC" (staging dd2f569, file digest
`52d3ac79d49b3af770fbd7298ed84f95f566717d42fa34070d5255b264bedb50`). G013 merged to staging as 8b0823a (MR !29); the
staging pipeline 2952592867, gate job 17214158089, passed every check (77 required, 441 controls) and failed at
finalization with exit 14, "the harness left 3 resources behind": three anonymous volumes (64-hex names) in
`inventory-leftover.txt`, beside the separately allowed render-cache volume. The original failed gate is not claimed
to have passed. Worker: the existing verified agent (agent-<id>; dispatch record
`dispatch-record-attempt4.md`), model **Fable 5.1** at effort **high** at spawn (actual unknown); no xhigh escalation:
the cause was confirmed from retained local evidence before any change, and the fix and proofs stayed bounded. New
branch `g013-drill-volume-cleanup` from staging dd2f569 (the MR branch is merged and deleted); no MR opened.

**Cause.** The PostgreSQL image `pgvector/pgvector:0.8.7-pg17` declares `VOLUME /var/lib/postgresql/data`, so each of
the recovery drill's detached psql clients (`docker run -d … $RD_PGIMG psql …`: `adminw` in RD-B-1, `rtw` in RD-B-2,
`lock` in MD-A) is created with an anonymous volume. The drill removed those containers with `docker rm -f` (no `-v`)
at run.sh 3199, 3237 and 3526, and `rd_purge` repeated the volume-less removal, so the three volumes outlived the
run. `docker_run_owned` registers only the container ID; by the time `cleanup-owned.sh` runs, the containers are
gone and their mounts can no longer be read, so the volumes were neither registered nor removed (and, absent from the
start inventory, were correctly counted as leftovers). The `put` client's minio-init image declares no volume, which
matches the count of exactly three. Confirmed before any change: the local runs' own `report/inventory-leftover.txt`
already listed exactly three anonymous volumes in b4-05, b8-01 and a3-02, each with the harness line "resources
created and not removed" in `harness.log`.

**Why local missed it.** `run.sh` writes `inventory-leftover.txt` (start inventory vs end inventory, images excluded)
and logs it, but treats it as informational; hosted `validate.sh` `complete` counts that same file and fails with 14.
The local batch scripts (b4/b8/a3) looked only at `cleanup-owned.sh --mode finalize` (registry-based: 0 owned
remaining, exit 0, "0 leftover g001.run containers") and never read the harness's leftover file. The signal was
written on every local run and unread. The attempt-4 batch script reads and counts that file exactly as hosted
`complete` does. This host holds 71 unlabelled anonymous volumes accumulated by those local runs; none was removed
(no sweep; they are not attributable to a registry).

**Fix** (`infra/validation/run.sh`, commit 8668e5b; inside the allow-list): `rd_client_rm NAME...` reads a client's
volume mounts while the container still exists, keeps only anonymous ones (64-hex names) that are absent from the
run's start inventory, registers each as an owned `volume` row, removes the container with `docker rm -f -v`, then
verifies each captured volume is gone or removes it explicitly by name (a failure is logged). Named volumes and
anything listed at the start are never touched; an absent client is not an error. It replaces the three `docker rm
-f` sites (adminw; put and rtw; lock) and the loop in `rd_purge` (normal end and every failure path). No sweep,
prune, allow-listing, inventory or acceptance change, and no fence, data or recovery-semantics change.

**Proofs (focused).**

| Batch | Target / candidate | Outcome |
|---|---|---|
| a3-00-style shellcheck (B-SHELL flags) on the edited run.sh | dd2f569 + the edit | `-S warning` exit 0, `--include=SC2251` exit 0 |
| a4-00 narrow reproduction and three-client lifecycle proof (`scripts/a4-repro.sh`; `rd_client_rm`, `registry_row` and `id_at_start` extracted verbatim from the checkout; same image and `docker run -d` shape as the drill; a labelled control volume created before the start inventory and mounted by the `lock` client as a named volume) | dd2f569 + the edit (first run), 8668e5b (logged re-run, `a4-00/repro.log`) | 10/10 PASS on both runs: R1 a detached client is created with an anonymous volume; R2 the previous `docker rm -f` leaves it behind (the leak reproduced); R3 the captured volume removed by name. Normal path N1 adminw, then put+rtw, then lock through `rd_client_rm`: 0 containers and 0 captured volumes left; N2 the three volumes registered as owned rows before removal; N3 the control volume survives, unregistered. Abort form A1 `rd_client_rm adminw rtw put lock` with all three present (put absent): 0/0 left; A2 a repeat with every client absent returns 0. E1 the host's anonymous-volume set is unchanged (71 = 71); E2 the control volume still present, then removed by the script itself |
| a4-01 `run.sh migrations lint lint-base provenance` with `--owned-registry`, a labelled control volume created before the start inventory, and the harness's `report/inventory-leftover.txt` counted as hosted `complete` counts it | dd2f569 / 8668e5b | **exit 0**: 53 PASS, 5 INFO, 0 FAIL, 0 UNMET; B-RECOVERY-DRILL (29 steps incl. RD-B-1, RD-B-2, RD-B-INT) and B-MAINTENANCE-DRILL (7 steps incl. MD-A, the `lock` client) PASS, B-RETENTION-CATALOG/IT, B-SHELL, B-LITERALS, B-SECRET-CANARY (55), B-SCOPE (1 changed path, in the allow-list), B-PROTECTED (594, expect-hash 66e80909… matches), B-RENDER-*, B-SKILL-UNCHANGED, B-REQUIRED-CHECKS (52 IDs). **`report/inventory-leftover.txt`: 0 lines (0 leftovers counted as hosted `complete` counts them; 3 on every earlier drill run)**, no "resources created and not removed" line; anonymous volumes on the host 71 before and 71 after; the labelled control volume present after the run and absent from the cleanup rows (removed by the batch script itself); registry volume rows 41 (38 in a3-02: the three clients' anonymous volumes registered before removal); no "could not be removed" line; finalize exit 0 (removed 0, unproven 0, remaining 0), 0 leftover containers |

**Evidence reuse, with provenance.** Attempt 4 changed only the drill's client removal in run.sh (exercised by
`migrations` through B-RECOVERY-DRILL and B-MAINTENANCE-DRILL) and this Result. The recovery/retention behaviour, the
tool `recovery.sh`, the migration, RECOVERY.md (12670f74…) and SECURITY.md (66e80909…) are unchanged since a3-02/a3-03
(staging 8b0823a), so their evidence stands; the other full-profile suites and the 441 CI controls are reused from the
hosted gate 17214158089 on 8b0823a (all PASS there; only finalization failed) and from b6-02/b7-01/b7-02 as before.
Earlier failed runs stay recorded as failed; `beginify-core-check:latest` was sha256:2db75c3eba6cf56a371ca0b7baca2c82e7d18e8587c47e70205982bf1e4274d5 before and after a4-01 (the hold is unchanged; nothing deleted).

**Correction of the attempt-3 counts.** a3-02 recorded 53 PASS, 0 FAIL, 0 UNMET and **5 INFO** (not "0 INFO");
a3-03 recorded 22 PASS, 0 FAIL, 0 UNMET and **5 INFO**. The INFO rows (PREFLIGHT, PROBE, DEPSCAN-PINS,
BUILD-LEDGER, IMAGES) are the harness's informational records, not checks, and did not affect the exit codes; the
attempt-3 section's "0 FAIL/UNMET/INFO" wording was wrong about INFO only; the two batch rows there now read
"0 FAIL, 0 UNMET, 5 INFO" (corrected in place in this attempt, as Codex asked), and this note records the change. The same two logs also carry, after "RESULT: PASSED", the line "resources created and not
removed" naming three anonymous volumes: the leak was visible locally and unread (see "Why local missed it").

Budget: **attempt 4 of B7 used — the CHECKPOINT attempt.** Per the Plan and the review, Codex reviews this diff and
Result at the checkpoint before any further meaningful attempt. Attempts 1-4 all spent on 2026-10-15: implementation
and review loop (1), C1-C3 (2), D1/D2 (3), this cleanup fix (4). Remaining: 3, released only after the checkpoint.

## Integration and hosted staging gate

Facts verified by the coordinator from the GitLab API and the job artifacts (2026-10-15).

- **MR !29 (attempt 3, history).** Merged to staging as `8b0823a`. Staging pipeline 2952592867, gate job 17214158089:
  **FAILED** at finalization (exit 14): three anonymous volumes left by the drill's detached PostgreSQL-image clients.
  Every one of the 77 required checks and all 441 CI controls had passed in that job; the failure stands as recorded.
- **MR !30 (attempt 4).** Codex accepted attempt 4 (checkpoint). MR !30 auto-merged at the exact head `8281ab7` into
  merge commit `29a0da5816ee2f6e5db5650ee55e1e2143873715`; `infra/validation/run.sh` equals the reviewed version and the
  other merge differences are only staging's `e71f998` guidance files. The source branch was deleted. MR fast job
  17216568332: success. GitLab Duo found nothing (note 4024464400).
- **Corrected staging pipeline 2952919063: success.**
  - fast job 17216727595: success;
  - gate job 17216727596: success, 22:07:34 to 22:53:29 UTC (2804.5 s); head `29a0da5`, tree `3ba2a789…`, comparison
    base `e71f998`; required count 77, stragglers 0, harness stop none;
  - results: 77 PASS, 0 FAIL, 6 INFO; B-CI-CONTROLS 441/441; B-RETENTION-CATALOG, B-RETENTION-IT, B-RECOVERY-DRILL and
    B-MAINTENANCE-DRILL PASS;
  - `inventory-leftover.txt` lists only `project-archetype-render-cache`, the same baseline as G011/G012: the drill's
    anonymous-volume leak is fixed;
  - the owner status rule applies: success qualifies the CI acceptance gate.
- **Gate artifact digests** (SHA-256, neutral labels):

| Artifact | SHA-256 |
|---|---|
| results.tsv | `e5c184e29f0d6ad96a9f86b315da7a1678c4e31e09f5ef808a4120d5933a6419` |
| summary.json | `185b0c4c28037babd0316443d2e336b49d3270d6c0f551007e8c5c45dc070d76` |
| inventory-leftover.txt | `77377f63e413eaf06f6135d14f827346869f79593bebb3d21d77d42f17fbf7aa` |
| ci-controls.log | `f10bea22cd7196f5df3dfdf80727190ffce35bc548622c7e0721cd43fee91b58` |
| required-checks.tsv | `c7b41dfdd073e323371289faec0398900d7be5a77eff3cda13c2330f0add5c60` |

## Evidence retention

**Necessary proof holds: none.** This Result, the Codex review records and the hosted job artifacts carry the proof.
The local folders are temporary and eligible for purge at closeout after a compact receipt:
- `<home>/.local/state/platform-evidence/g013-impl-4e53de6/` (implementation batches, drill and harness runs,
  the S-review report, the attempt-2..4 evidence and the batch log);
- `<home>/.local/state/platform-evidence/g013-plan-5e1e987/` (planning).

**Explicit holds (untouched; not G013 cleanup):** the 71 pre-existing unlabelled anonymous volumes on this host
(accumulated by earlier local drill runs; no sweep) and `beginify-core-check:latest` at
sha256:2db75c3eba6cf56a371ca0b7baca2c82e7d18e8587c47e70205982bf1e4274d5 (provenance unresolved). G021 stays closed.

**Open external expiry:** the G012 braces exception expires 2026-10-29 (inclusive); after that date B-DEPSCAN fails by
design until a fixed release or a renewed owner disposition.

## Closeout owed

- Codex final ACCEPT, the PROGRESS update and the G013 handoff (Codex).
- The coordinator's bounded cleanup of the resources the G013 implementer created, as listed in its closeout
  hand-back: the focus environment (containers `g013-focus-pg` and `g013-focus-minio`, volumes `g013-focus-pg` and
  `g013-focus-minio`, network `g013-focus-data`, all labelled `g013.batch=focus`), the images built by the focus and b3
  drivers and the first harness drill run (`g013-focus-*`, `g013-b3src-*`, `g013-b3tgt-*`, `g013-rec-029344-*`), the
  worktree `<source-checkout>` with its
  local branches (`g013-closeout`, the former feature branch `g013-recovery-retention-incident-evidence`, the native
  `worktree-agent-<id>`) and the two evidence folders above. The attempt-3/4 branches
  (`g013-attempt3`, `g013-drill-volume-cleanup`) belong to the session that ran those attempts.
- The holds above stay untouched.

## Commands and validation

Binding prep ran read-only Git and grep only. Every execution batch ran under the coordinator-owned harness flock,
after `scripts/prebatch.sh` (fetch and target-ancestor check, `bash infra/ci/docker-desktop-mounts.sh`, port and
ownership checks, the README secret-scan line, target and candidate SHAs). The batch log with script hashes and
resource lists is `<home>/.local/state/platform-evidence/g013-impl-4e53de6/batch-log.md` (outside Git); every
batch's evidence is in its directory there.

| Batch | Target / candidate | Outcome |
|---|---|---|
| b0, b0b tool contract | 1b15499 / 1b15499 | Plan §3.5 contract confirmed; two mechanism details recorded: catalog ACLs compare normalized; holder and in-flight sessions end by `pg_terminate_backend` |
| b1-01..03 static/unit focus | 1b15499 / WIP | provider source count and shape pin; gofmt; then PASS |
| b2-01..07 focused DB tests | 1b15499..ce55227 / WIP | 00009 nested dollar quote and relation count fixed; holds regex replaced; TestRetentionCatalog 3/3, TestRetentionMigration 5/5, TestRetention RT-1..9 with E-RET-1..3, reconciled migrate (13 tests), access (incl. TestAudit after amendment 01), tenancy models, identity: PASS |
| b3-01..08 recovery/maintenance focus drill | to 30b5528 / to 308982c | b3-08: all 29 steps PASS |
| b4-01, b4-02 `run.sh migrations`, owned registry | 30b5528 / 1e33adc, 80ca454 | run.sh exit 1: drill port defects (below) |
| b4-03 `run.sh migrations`, owned registry | 6aca14e / 212b58f | run.sh exit 0; both drills PASS; finalize exit 0 |
| b5-01 `run.sh lint lint-base` | 6aca14e / 459c694 | exit 0 (B-SHELL, B-SCOPE, B-SECRET-CANARY, 16 required IDs) |
| Delegate-1 review | 6aca14e / 459c694 | BLOCKING-FINDINGS; see "Security review (Delegate 1) and dispositions" |
| b3-09 focus drill after the fixes | fb8fd07 / d7d1a41 | all 29 steps PASS with the new negatives |
| b4-04 `run.sh migrations` | fb8fd07 / 015bf83 | exit 2: the post-maintenance measurement read the manifest that F11 now removes; fixed |
| b4-05 `run.sh migrations` | fb8fd07 / ddb89ac | exit 0; B-RECOVERY-DRILL and B-MAINTENANCE-DRILL PASS; 34 required IDs; finalize exit 0 |
| b6-01 broad `all` (provisional) | fb8fd07 / d0e1b1a | Launched before the F4 hold arrived and kept as provisional evidence only: 105 PASS, 2 FAIL (B-CI-CONTROLS VR1, B-REQUIRED-CHECKS); finalize exit 18 (one unproven image) |
| b7-01 focused control rerun | c1e16de / 652157f | `validate-controls.sh --only VR1,VS1`: both PASS |
| **b6-02 final broad `all`** | **31444dd / 67d99f5** | **run.sh exit 1: 104 PASS, 3 FAIL, 7 INFO**; finalize exit 18; details below |
| b7-02 focused control rerun | aae1101 / 537d124 | `validate-controls.sh --only VS2,VS3`: both PASS |

**Final run b6-02** (`bash infra/validation/run.sh --base 31444ddb0b3e0d78e2e8b8a42d43b37f5795c930 --allow-file
management/v0/results/G013-scope-allowlist.txt --expect-hash docs/SECURITY.md=834a4285bb4c7fa8e9bae2e17f04c1a9acf696fd28000107280a59f6d76effd8
--owned-registry <reg>/owned.tsv --secrets-file <reg>/secrets --report-dir EVID/b6-02/report all`; this Result was at
its base bytes, digest b9012164…, during the run; fresh start inventory; 16:38-17:37 UTC). The registry and
finalize were exercised as in `infra/ci/validate.sh`.
- PASS on every G013 check: B-RETENTION-CATALOG, B-RETENTION-IT, B-RECOVERY-DRILL (all RD steps), B-MAINTENANCE-DRILL
  (all MD steps), the version-9 reconciled checks (B-MIGRATE-*, B-AUDIT-IT with amendment 01, B-OBJ-IT, B-IDENTITY-CATALOG,
  B-PERSIST, B-DBRENAME, B-TOGGLE-OFF), B-SHELL, B-SCOPE (46 paths since 31444dd, all within the allow-list),
  B-PROTECTED (570 files equal the base), B-SECRET-CANARY (220 generated secrets absent).
- FAIL B-LITERALS: the identity literal `beginify` in `contracts/mcp/connection-context.md` line 49 (`acme.app.foundation.test`).
  Staging introduced it in 0449101 (G014 planning); it is still there at aae1101. It is not a G013 path and lies outside
  the allow-list, so G013 does not edit it.
- FAIL B-CI-CONTROLS (439 passing): VS2 and VS3, "an available observation is not ok; a size is not a Docker CLI
  size", the same Docker storage-observation class as VR1 in b6-01 under the broad run's load. Focused rerun of only
  those controls: b7-02 PASS (VR1 and VS1 likewise passed in b7-01). G013 did not change `validate.sh` or the storage
  report.
- FAIL B-REQUIRED-CHECKS: only B-LITERALS and B-CI-CONTROLS.
- E-RET-1..3 have the same plans as in b2 (below); no E-RET-TRIGGER. E-RLS (40) and E-OBJ (2) trigger lines are the
  existing G007/G009 observations that G012 also recorded; G013 does not change their tests.
- `cleanup-owned.sh --mode finalize` exited 18 (removed 0, refused 0, unproven 1, remaining 0, failed 0), reported as
  is. The unproven image is `beginify-core-check:latest`. Before the run it was
  sha256:a5ab2420c8d92c1c06c66f7a3fe3786a6f38c3cb384a39c3afc417ea3d0e5150 (in the start inventory), as Codex's ruling
  required us to preserve. After the run the tag resolves to
  sha256:2db75c3eba6cf56a371ca0b7baca2c82e7d18e8587c47e70205982bf1e4274d5, with the same Created time
  (2026-10-15T13:03:14.551880295Z), and the old ID no longer exists. The README platform:check block of the root
  project rebuilds this harness-wide name without a build or capture row, so the accepted procedure records no
  ownership. Its provenance stays unresolved. I deleted nothing and the harness was not changed. 0 g001.run containers
  were left.

Defects found by the drills and fixed (each rerun in full):
- the restore object loop and the replay loop read the manifest and ledger on stdin while `docker run -i` clients in
  the loop body consumed it: both loops now read on fd 3 (the continuation preflight loop too);
- replay, tombstone purge and status refused a stopped MinIO: they accept it and record the provider unavailable;
  restore still refuses before mutation;
- status could exit through `die` in its pending-review probe: the probe runs in a subshell;
- TestRecoveryVerify expectation errors (V1 omitted the creators' group_owner bindings; V4 asked for 500 events, the
  maximum is 100);
- harness port defects: an application name inside single quotes, unregistered target images, backgrounded tool runs
  signalling a function subshell, and (b4-04) the measurement reading a manifest that expiry now removes.

RD-M measurements (b6-02, local Docker 27.5.1; `recovery-drill.tsv`): pg_dump 17.11; B1 holds 80 audit events, 15 role
bindings and 5 objects; database dump 159,634 bytes, objects tar 22,528 bytes; capture window
17:06:44.71Z-17:06:50.27Z (database clock); recovery point 2026-10-15T17:06:50.274401Z (the B1 window end: later changes
are lost except replayed restrictions); restore database step 2 s, object step 21 s; stage seconds rdb 202 (includes
the 15 s drill pause and the drain), rdr 11, rdf 41, rda 1, rdy 12, rdg 6, rdo 26, mda 40, mdo 50. These are
fixture-scale observations on one host, not production recovery objectives.

PG-005 measurements (E-RET, RT-9; b2-06, b4 runs and b6-02 `retention-eret.jsonl`; pinned distribution: 8,757 events,
Organization A 4,001 with 200 held, global 2,002, Organization B 1,051; 253 stored-object fixture rows); b6-02 times:
- E-RET-1 Organization batch (500 rows): Index Scan on `events_tenant_time_idx` plus `events_pkey`, 24.9 ms; count by
  a bitmap scan of `events_tenant_time_idx`, 62.2 ms.
- E-RET-2 global batch: Seq Scan plus Sort, as the Plan expected (the partial index excludes NULL organization),
  29.8 ms; count Seq Scan 18.6 ms.
- E-RET-3 tombstone candidates: bitmap scan of `stored_objects_organization_group_idx` plus Sort, 2.6 ms;
  confirmed-ID DELETE by `stored_objects_pkey`, 1.7 ms.
- Locks during each DELETE (b2-06): AccessShare/RowExclusive on the table and its indexes only; no helper fan-out.
- No `E-RET-TRIGGER`; no index added. Bounded fixture observations, not production performance approval.

## Acceptance criteria

Mapping: Plan §10.3. Status is the implementer's evidence claim; Codex decides acceptance.

| AC | Evidence | Status |
|---|---|---|
| AC1 complete consistent checkpoint | RD-B-1..5, RD-B-NEG incl. the failed-marker case (C2), RD-K-1..3 (b3-11, b8-01; earlier b6-02); F5/C2 exit guard, F6 PUBLIC/connectable-login checks, F14 root lock | met |
| AC2 isolated restore rejecting unsafe or incomplete inputs | RD-R (9 refusals before mutation incl. a reused receipts directory), RD-F (staged, receipt-bound continuation; other checkpoint, source alias and foreign object refused), RD-I | met |
| AC3 restored content and authority match | RD-V, RD-A, RD-G TestRecoveryVerify V1-V9; receipts bound to the current restore (F1) | met |
| AC4 restrictions and deletions not resurrected | RD-P, RD-L (7 gate refusals incl. an invalidation receipt of another restore), RD-Y (stopped provider, injected audit fault, gap receipted; wrong references never already applied), RD-G V8-V9, RD-O replay head binding and ledger drift (C1); strict references (F2/F3) | met |
| AC5 restricted, bounded audit retention | B-RETENTION-CATALOG, B-RETENTION-IT (RT-1..9, E-RET-1..3), MD-A; F4 residual disclosed and accepted by the owner | met (with the accepted F4 residual) |
| AC6 object maintenance and backup expiry | MD-O, MD-R, MD-E (only the EXPIRED record remains) | met |
| AC7 failures reach an operational signal | RD-I, RD-O, MD-S, RD-Y status gap, RD-B-3 fence held; the exit guard reports the actual fence state (C2) | met |
| AC8 runbooks | docs/foundation/RECOVERY.md (incl. the F4 diagnosis and terminate-and-rotate runbook, the owner acceptance wording of C3, the replay head binding); RD-M above | met |
| AC9 validation, provenance and privacy | corrected hosted staging gate 17216727596 on 29a0da5: 77 required PASS, 0 FAIL, B-CI-CONTROLS 441/441, all four G013 IDs PASS, leftover inventory only the baseline render cache (artifact digests under "Integration and hosted staging gate"); local covering runs per attempt; failed history kept (b6-02, staging gate 17214158089); image ownership hold | met (Codex decides) |
| AC10 review, cleanup and closeout | Delegate-1 review and dispositions; Codex reviews through the attempt-4 checkpoint ACCEPT; MR !29 and !30 integration with source branches deleted; corrected hosted gate PASS; the owned-resource list under "Closeout owed" | pending Codex final ACCEPT and the coordinator's bounded cleanup |

## Architecture deviations

None. Implementation details recorded against the Plan (no contract or architecture change):
- **Amendment 01** (Codex accepted): the TestAudit ACL comparison is canonical; the only edit outside the original
  allow-list.
- **F1 binding digest:** the restore binding is the digest of the restore receipt's identity lines (set, manifest
  digest, target project, PostgreSQL container and volume, system identifier, start time), not of the whole receipt,
  because a staged object step legitimately appends lines after invalidation. Invalidation and replay receipts also
  record the target identity.
- Restore applies the TOC (minus the two provisioned schema entries) as a rendered script in one `psql
  --single-transaction` whose first statement withdraws the owner's provisioned default sequence grant to runtime.
- Catalog fingerprints normalize ACLs and constraint/trigger expression text; the roles section includes memberships
  crossing the project boundary (F12).
- Fence X0 refuses a visible other superuser session or PUBLIC CONNECT; the drill uses the test-only pause
  `RECOVERY_DRILL_PAUSE_AFTER_X2`; a backup against a paused MinIO is refused at X0.
- `retention.holds` uses `none` for an empty set; the optional trace was not used.
- Replay, tombstone purge and status accept a stopped MinIO and record it unavailable; restore refuses it.
- **F4 (accepted residual):** the approved table-level DELETE grants let the retention login hold any table lock on
  `audit.events` and `app.stored_objects`, blocking writes across Organizations; owner acceptance "Yes for 013
  question" (2026-10-15). Disclosed in SECURITY.md §5 and RECOVERY.md with the safeguards Codex required.

## Unresolved issues and blockers

- B-LITERALS failed in b6-02 on staging content outside G013; Codex corrected it on staging dcb68a4 and b5-02/b8-01/
  b8-03 record B-LITERALS PASS (attempt 2).
- B-CI-CONTROLS VS2/VS3 (b6-02) and VR1 (b6-01) failed in the broad runs and passed on focused reruns; the
  load-sensitive storage observation is not shown fixed: for Codex's status decision.
- `cleanup-owned.sh` finalize exit 18 in b6-02: `beginify-core-check:latest` was rebuilt during that run to an
  unregistered new ID (sha256:2db75c3e…, still current); explicit ownership hold, provenance unresolved, nothing deleted.
- External timing risk unchanged: the G012 braces exception expires 2026-10-29 (inclusive).
- Staging gate 17214158089 on 8b0823a failed at finalization (exit 14, three anonymous volumes of the drill's detached
  psql clients); that run is not claimed to have passed. Attempt 4 fixed the lifecycle, and the corrected staging gate
  17216727596 on 29a0da5 passed (see "Integration and hosted staging gate"). This host keeps 71 unlabelled anonymous volumes accumulated by earlier local
  drill runs; none was removed (no sweep); they are the operator's to reclaim.
- Accepted limitation (MR !29 D2, documented in RECOVERY.md): tombstone purge can stop at an oldest fully skipped
  batch; the operator resolves the blockers through the existing ledger, deletion and provider paths and retries. No
  paging redesign was requested or made.

## Suggested follow-up

None proposed by G013 (no follow-up features added).
