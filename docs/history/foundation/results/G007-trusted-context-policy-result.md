> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G007 — Trusted context and shared policy evaluation: Result (in progress)

- **Status: IN PROGRESS — not submitted.** *(Superseded at S9: see "S9: submission status and AC1–AC14 evidence"
  at the end; this header is kept as history.)* This Result is opened early to hold the outstanding-findings checklist
  (AGENTS.md "Validation batches: finish, fix, rerun, then advance", staging `9bee708`). It is completed at S9; only
  Codex accepts G007.
- **Plan:** [Plan r1](../plans/G007-trusted-context-policy-plan.md) (`b6bf98f4…0e60`), unchanged; release
  G007-execution-binding-codex-confirmation.md (not included) (`916295f2…`).
- **Execution amendment authority:** G007-execution-blockers-codex-review.md (not included)
  (SHA-256 `432d5f171e417f334676e140858a1dbbb371ea909d411b973cea77a2649bf9a3`, ACCEPT of bounded execution amendments, not
  implementation acceptance), with the amended Goal (`664ec86b9b35a47f7839d7866938746623f651745d7080ff108d316081710aae`) and
  Spec (`d6918879dab5a39c6f6b35c673b3e2ffba77c404441add55f07ba7d5674fe181`), all committed on staging at `d98964e`.
- **Checkout:** `.claude/worktrees/g007-agent-5c6f436`, branch `g007-trusted-context-policy`. Effort: xhigh.
- **Accounting (Codex review §4; owner goal-level attempt policy, BUDGET_POLICY §3, staging `06b7776`):** zero historical
  counted cycles; **attempt 1 active, 0 completed; B7 = seven shared attempts, checkpoint after attempt 4.** Scoped
  debugging, test reruns and approved subagents share attempt 1; a failed-attempt retry or a supervisor correction round
  uses the next attempt. No reset of prior counts. (This replaces the earlier B7-cycles line.) **Attempt 2 of B7**
  is authorized for C1 and C2 together (Codex review `1bed0ddb…`, handoff `d4a53b90`); attempt 1's history is
  preserved, and the checkpoint stays after attempt 4. See "Attempt 2" below.

## Outstanding-findings checklist

| # | Source | Finding | State |
|---|---|---|---|
| F1 | Batch `s23-unit-1` (unit; HEAD `fed3e8e`, base `6cd33d6`, 2026-10-10T19:34:37Z, exit 1) | B-UNIT-CHECKS FAIL: the README check block's secret-scan (gitleaks, rule generic-api-key) flags `management/v0/plans/G007-planning-inputs.json:89` and `G007-planning-inputs-r0.json:50` (the `credential.go` path plus its SHA-256, a false positive). core-check (gofmt, vet, every unit test incl. the oracle matrix), web-check and intelligence-check completed | **Closed** by `s23-unit-2` (unit, HEAD `8e3bdef`, base `d8d2c62`, 22:24:28–22:25:25Z, exit 0, leftover 0). Fix: per Codex review §3, G007 carries the exact reviewed G051 O1 configuration at the one added scope path `infra/security/gitleaks.toml` (SHA-256 `df25add66bcf6819e0e94066d318f8692adcdbc692989b26087ceb4b483ece45`, from G051 candidate `49220be`, shared provenance; staging still `4f26187c…`). It permits only the exact `credential.go` path-plus-digest line for generic-api-key. No other G051 content. A development scan of all tracked files reported no leaks before the rerun |
| F2 | Batch `s23-unit-1` | B-REQUIRED-CHECKS FAIL, solely because of F1 | **Closed** with F1 (`s23-unit-2`) |
| F3 | Plan r1 §11 stop (scratch probe, PostgreSQL 17.11) | The runtime role cannot row-lock the SELECT-only membership projections (42501 "permission denied for view"); §4.1 steps 4–5 and CT-POL-3 as written could not run | Resolved by Codex review §1 (option A): see the execution amendment; bound validation (CT-POL-3, U-BOUNDARY, tenancy/policy/migration regressions, O-1..O-15, E-PG-2) pending |
| F4 | Scratch dev run 2 (E-PG-1, not evidence) | The literal §6.7 trigger fires at s50 for the Organization-only profile write: statements 6b (groups) and 7 (associations) use sequential scans on small fixture tables (260 and 200 rows). No Organization-wide membership path (statement 4 examined 2 entries; 5a/5b 68–121): H-1 resolves to the subject path | Disposed by Codex review §2: accepted as a recorded PG-001 observation under the owner's unknown-workload/no-latency-target disposition; trigger and plan-summary reporting kept; the same scan at 6b/7 is recorded (`"known": true`) without re-escalation; anything new returns to Codex. No index, forced plan, SQL reshaping, M1/M2 or benchmark |
| F5 | Batch `s1-lcm-1` (lint controls migrations; HEAD `8e3bdef`, base `d8d2c62`, 22:33:23–22:57:36Z, exit 1, leftover 0) | B-CI-CONTROLS FAIL (290 of 291 controls pass): V17b (`STUB_HARNESS=locked`) exit 17, want 19; published none, want full. Cause: staging commit `bed2622`'s streamed evidence scan in `infra/ci/validate.sh` (`scan_set`: `tar -xf - -C /tmp/w \|\| exit 122` inside the gitleaks container, run as the non-root job user) cannot unpack the V17b fixture's mode-555 directory (`validate-controls.sh:593`), so gitleaks returns 122 and the raw scan 17 before the expected scrub failure (19). Inherited: the same signature failed G051 `d1-affected` on staging content; G051's unintegrated cycle-3 `validate.sh` fixes it. `infra/ci/**` is outside the G007 scope (a bound input). B-REQUIRED-CHECKS FAIL follows. Every G007 database check passed: B-MIGRATE-*, B-IDENTITY-CATALOG, B-TENANCY-*, B-BUSINESS-PROFILE-*, B-POLICY-CATALOG/MIGRATION/IT/ORDERING | **Closed** by `s1-lcm-2` (lint controls migrations, HEAD `f549283`, base `38ad78c`, 04:17:40–04:44:57Z, exit 0, leftover 0): B-CI-CONTROLS 291/291 incl. V17b "exit 19; full published; scrub=19". Fix: owner-approved minimum correction (Codex handoff `126605f5`), carried as a narrowly recorded scope amendment (execution amendment 2) |
| F6 | G051 O6 (predicted, not observed by G007) | The carried scanner configuration (`df25add6…`) is an unregistered render customization: `infra/security/gitleaks.toml` is archetype-rendered (manifest `4f26187c…`) and `infra/validation/render-customizations.tsv` registers 24 rows, which `run.sh` pins, so B-RENDER-PLAN and B-RENDER-REPEAT (suite `provenance`) would fail. The likely fix (a register row, out of scope, plus `run.sh`'s pinned count 24 to 25, in scope) is not made | Open — escalated (Codex); no batch including `provenance` until decided → **Disposed** by the shared register correction (follow-up review `92b226dc…`; execution amendment 3); register and count fix applied, the provenance rerun is pending → **Closed** by `s6-prov-1` (lint provenance; run `9385bd`, HEAD `ac7fe93`, binding `g007-binding-reconciled-1a278d1-aaaafe0ea`, exit 0, leftover 0): B-RENDER-PLAN and B-RENDER-REPEAT pass at 25 registered customizations |
| F7 | Batch `s1-lcm-2`, B-POLICY-IT E-PG-1 (`policy-epg.jsonl`) | Two new trigger lines (`"known": false`): at s50, statement 6a (the target Group point read of `group.rename`, `organization_id = $1 AND id = $2 FOR NO KEY UPDATE`) used a sequential scan of `groups` in both context modes (4 pages, 4 shared hits, 1 row, about 0.03 ms). The accepted 6b/7 scans recur (marked known), in this run also in selected-Group mode (one candidate). No Organization-wide membership path (statement 4 on the subject index). Not covered by the §2 disposition (6b/7 only), so it returns to Codex. No index, forced plan or SQL change | Open — escalated (Codex) → **Disposed** (follow-up review `92b226dc…`): an exact small-fixture PG-001 observation (execution amendment 3) |
| F8 | S5 read-only review (finding 1; report copy SHA-256 `d08d2342a026e6164516dfa807354e7afe049cf3a0070854b7a5b5e40ecb92d8`) | Plan-statement inaccuracy, not a code defect. `slug` belongs to `groups_organization_slug_key UNIQUE (organization_id, slug)` (00003:51), so PostgreSQL treats it as a key column: a `ChangeGroupSlug` that really changes the slug takes `FOR UPDATE` on the Group row at step 9, above step 6's `FOR NO KEY UPDATE`. That conflicts with the foreign-key `FOR KEY SHARE` locks of `LinkGroupBusinessProfile` and `GrantGroupRole` inserts on that Group (the guard waits for one in flight; later ones wait until it ends). A profile update also takes an exclusive row lock on P at step 9. Behaviour equals the accepted raw `ChangeGroupSlug`; the reviewer traced those writers (their `FOR KEY SHARE` is their last contended lock and they never wait for a guard lock) and found no new deadlock. Inaccurate statements: Plan r1 §4.4 ("at most two exclusive locks", the implied no-upgrade), §4.7.1 (G(t) `FOR NO KEY UPDATE` as the only Group mode) and §4.7.4 ("never delayed" for grants and `Link`). E-PG-2 uses the unchanged slug (`footprint_integration_test.go:269`), so the upgrade is not exercised | **Corrected in the guide** (AUTHORIZATION.md "Accepted writers" and "Locks", supervisor direction after S5); Plan bytes unchanged; the S9 AC8 statement will state the upgrade and the P lock. No lock-mode change (taking `FOR UPDATE` up front would be a §11 stop). Offered to Codex, not added: an optional E-PG-2 variant with a changed slug. Sent to Codex by the supervisor → **ACCEPTED** as the recorded Plan-statement deviation (Codex review `1bed0ddb…` §F8). The unchanged-slug footprint did not measure the slug-change upgrade; AUTHORIZATION.md now says so. No lock-order or lock-mode change, no footprint variant or benchmark |
| F9 | S5 read-only review (finding 2), batch `s1-lcm-2` E-PG-1 (`runs/s1-lcm-2/report/policy-epg.jsonl`, SHA-256 `97b545e8d90111d992bfb497f78cf4d0b0b59a3b0282efb56c017ac8c874e0ef`) | Sequential scans outside the §6.7 trigger set (statements 4–7), so no trigger line, not recorded until now. Each returned 1 row with shared-buffer hits only (no reads), in both context modes and for both subjects (s1, s50): statement 2 (S, `sessions` `FOR SHARE`): 7 removed by filter, 1 hit, about 0.016–0.017 ms, 8 plans; statement 3 (O, `organizations` `FOR SHARE`): 5 removed, 1 hit, about 0.016–0.018 ms, 8 plans; statement 8 (P, `business_profiles` read, `business_profile.update`): 103 removed, 2 hits, about 0.015–0.019 ms, 4 plans; statement 9 (W, the UPDATEs): `business_profiles` 103 removed, 2 hits, 4 plans, and `groups` (`group.rename`) 263 removed, 4 hits, 4 plans, statement times about 0.08–0.10 ms including the write. They contradict §4.7.2's primary-key-probe hypotheses for rows 2, 3, 8 and 9 (tables of 8, 6, 104 and 264 rows) | **Pending Codex disposition** (sent by the supervisor) → **ACCEPTED** as additional PG-001 observations under the owner's unknown-workload disposition (Codex review `1bed0ddb…` §F9): they invalidate the Plan's universal point-probe hypothesis but establish no production regression or scalability guarantee. Recorded as small-fixture observations only; the code's `known` flag and the raw trigger lines are unchanged; no query, index or lock change. AUTHORIZATION.md states the observed facts without calling them accepted, and (advisory 3) that only the locking forms of statements 4–5 are measured, the projection snapshot forms inferred |
| F10 | Batch `s46-clpmo-1` (contracts lint provenance migrations off; run `ab2634`, HEAD `70e5f5e`, base `1a278d1`, exit 1, leftover 0) | B-TOGGLE-OFF-ADDITIONS FAIL ("variant composition"; log: `customized infra/security/gitleaks.toml: rendered base or customization differs`), and B-REQUIRED-CHECKS FAIL solely because of it. Cause: the shared register correction (F6, amendment 3) registered `infra/security/gitleaks.toml` as a customization, and `variants_off` requires every registered customization in the disabled variant's manifest to equal the root's customized file. Every other such path reaches the variant through `run.sh`'s APPLICABLE list (or is toggle-rendered, or not rendered without intelligence); `infra/security/gitleaks.toml` is not in APPLICABLE, so the variant keeps its rendered base (`4f26187c…`, toggle-independent, equal to the manifest) while the root carries `df25add6…`. Not G007 code: `s1-rorf-1` passed before the register row existed, when the path was compared only with the variant's own manifest. The O6 analysis did not cover this consumer of the register. This is **G051's O7, shared with G007**: G051 hit the identical failure in `f1-final-qual` | **Open — pending Codex approval; not applied** (O6 limited `run.sh` to the count line). Proposed fix: append two lines to `run.sh`'s APPLICABLE after G007's last entry `docs/foundation/AUTHORIZATION.md`, byte-identical to G051's reference hunk (`o7-proposal/run.sh.applicable.diff`, SHA-256 `b3d3109d…5531`) except for the preceding line (see "s46-clpmo-1 and O7" below). Rerun plan after approval: `off` plus the required-check evaluation; every other suite reused with provenance → **Accepted by Codex** (review section "G051 O7 / G007 shared toggle-off correction — 2026-10-11", SHA-256 `5cd3e383…ef1f`; handoff `b85a6792`); the one APPLICABLE entry is applied (`run.sh` SHA-256 `9efc3541…7d68`); the `lint off` rerun is pending. `s46-clpmo-1` stays recorded as failed → **Closed** by `s7-lo-1` (lint off; run `1d21ae`, HEAD `a315a59`, base `0706f9f`, exit 0, leftover 0): B-TOGGLE-OFF-ADDITIONS and B-REQUIRED-CHECKS pass |
| F11 | Batch `s8-final-3` (all unit on the merged harness; run `e1b472`, HEAD `e3d01d1`, base `7dc3621`, exit 1, leftover 0) | B-CI-CONTROLS FAIL ("CI controls (exit 1, 436 passing)"); B-REQUIRED-CHECKS FAIL solely with it. Three G051 controls in `infra/ci/validate-controls.sh` pin the full-pipeline split from the pre-G007 register: VA7 (`summary.json required_count is not 56`; line 1110, detail line 1113), VU1 (`rows fast/gate/union 33/60/89, want 33/56/85`; line 1204) and J8 (`the remainder result has 60 check rows, want 56`; line 1945). The controls compute from the real `required-checks.tsv`, so G007's 4 B-POLICY-* rows (suite `migrations`, full only) move the gate remainder from 56 to 60 and the union from 85 to 89; the fast partition stays 33 (fast suites in full: 17 lint + 5 provenance + 6 contracts = 28; 88 − 28 = 60; 33 + 60 − 4 final = 89). This is the mechanical controls reconciliation Plan r1 §7 anticipates for whichever Goal integrates second, but `infra/ci/**` is outside G007's 13-expression scope (an immutable input) | **Open — reported to the supervisor; not fixed** (batch rule; scope). Proposed minimal change in `infra/ci/validate-controls.sh`, 4 lines: VA7 `sum_has required_count 56`→`60` and its detail "56 required IDs"→"60"; VU1 `vu_split full 33 56 85`→`33 60 89` and its detail "gate 56 -> 85"→"gate 60 -> 89"; J8 `(( n != 56 ))`→`60` and "want 56"→"want 60". Rerun plan after approval: `controls` plus the required-check evaluation; the other 87 passing checks reused from `s8-final-3` → **Approved** as C1 (Codex review `1bed0ddb…` §C1, attempt 2): the four control sites and the OPERATIONS.md coverage row updated; the `controls` rerun is held → **Closed** by `a2-c1c2-1` (run `1eb376`, HEAD `12d2d8e`, base `da1ed80`, exit 0, leftover 0): B-CI-CONTROLS 439/439 incl. VA7, VU1 (33/60/89), VU9 (33/42/71) and J8; B-REQUIRED-CHECKS pass |
| F12 | Batch `s9-ops-1` (lint-base; run `35cfc5`, HEAD `b5a5f40`, base `a324b70`, exit 1, leftover 0) | B-SECRET-CANARY UNMET ("no secrets were generated"; `run.sh:3525`) and B-REQUIRED-CHECKS FAIL solely with it. `lint-base` generates no secret, so the canary check has nothing to scan; every run needs a secret-generating suite (Plan r1 §7: "`controls` always runs with a secret-generating suite"). The affected content check passed: B-LITERALS (the edited OPERATIONS.md), with B-NAMING, B-IDENTITY-GUARDS, B-DOCKER-ONLY and B-SCOPE. Batch selection, not a content failure | **Open — reported to the supervisor; not rerun** (batch rule). Proposed: one rerun `lint-base unit` (unit generates its environment secrets, so B-SECRET-CANARY has a real scan; about 1 minute). No content change → The supervisor recorded the composition as its own specification error (it specified `lint-base` alone) and approved the rerun. **Closed** by `s9-ops-2` (lint-base unit; run `12725f`, HEAD `f323b30`, base `a324b70`, exit 0, leftover 0): B-SECRET-CANARY and B-REQUIRED-CHECKS pass. `s9-ops-1` stays recorded as failed |


Closed findings are kept below the table with their resolution and rerun evidence.

## Batches

| Step | Suites | HEAD / base | Outcome |
|---|---|---|---|
| `s23-unit-1` | unit | `fed3e8e` / `6cd33d6` | FAIL (F1, F2); B-DOCKER-ONLY, B-SECRET-CANARY, B-SCOPE PASS; leftover 0 |
| `s23-unit-2` | unit (rerun of the failed batch) | `8e3bdef` / `d8d2c62` | PASS; leftover 0 |
| `s1-lcm-1` | lint controls migrations | `8e3bdef` / `d8d2c62` | FAIL (F5: B-CI-CONTROLS V17b, B-REQUIRED-CHECKS); 37 PASS incl. every database check; leftover 0 |
| `s1-lcm-2` | lint controls migrations (rerun of the failed batch) | `f549283` / `38ad78c` | PASS: 39 PASS (B-CI-CONTROLS 291/291; B-POLICY-CATALOG 6/6, -MIGRATION 5/5, -IT 38/38 + footprint 2/2, -ORDERING 17/17); E-PG-1: 6 trigger lines, 2 new (F7); leftover 0 |
| `s1-rorf-1` | root ops rename off | `7c2ea2a` / `38ad78c` | PASS: 04:50:28–05:28:54Z, exit 0, 42 PASS, `RESULT: PASSED`, `run-approved: harness exit 0; postcheck passed`; leftover 0. It ran on through the worker's session-limit interruption (below) |
| `s6-prov-1` | lint provenance | `ac7fe93` / `1a278d1` | PASS: run `9385bd`, binding `g007-binding-reconciled-1a278d1-aaaafe0ea`, launcher 09:14:37Z (queued on the shared lock), end 09:47:59Z, exit 0; 26/26 checks PASS incl. B-RENDER-PLAN/REPEAT (25 customizations), B-SCOPE (37 paths), B-REQUIRED-CHECKS; leftover 0 |
| `s46-clpmo-1` | contracts lint provenance migrations off (S4 + S6, combined by supervisor direction) | `70e5f5e` / `1a278d1` | FAIL (F10: B-TOGGLE-OFF-ADDITIONS, B-REQUIRED-CHECKS); run `ab2634`, binding `g007-binding-reconciled-1a278d1-aaaafe0ea`, launcher 10:19:54Z (queued on the shared lock), harness 10:31:42–10:47:17Z, exit 1, `run-approved: harness exit 1; postcheck passed`; 49 PASS: contracts 6/6, lint 17/17, provenance 5/5 (B-RENDER-PLAN/REPEAT at 25), migrations 15/15 (B-POLICY-CATALOG 6/6, -MIGRATION 5/5, -IT 38/38 + footprint 2/2, -ORDERING 17/17), B-TOGGLE-OFF, B-OBS-OFF, B-BROWSER-OFF (141 tests), B-DOCKER-ONLY, B-SECRET-CANARY, B-SCOPE (42 paths); E-PG-1: the same 6 trigger lines (2 `"known": false` = the disposed F7 6a lines) and the identical sequential-scan set of F9 (`policy-epg.jsonl` SHA-256 `23541e82…24c3`); leftover 0 |
| `s7-lo-1` | lint off (the F10/O7 rerun after execution amendment 4; the name is not S7) | `a315a59` / `0706f9f` | PASS: run `1d21ae`, binding `g007-binding-reconciled-0706f9f-aaaafe0ea` (snapshot `ac1d9fdd…10fb`, 93 inputs), launcher 12:14:27Z, end 12:26:53Z, exit 0, `RESULT: PASSED`, `run-approved: harness exit 0; postcheck passed`; 25/25 PASS: lint 17/17, B-TOGGLE-OFF-ADDITIONS (all 54 applicable additions byte-identical to the root), B-TOGGLE-OFF, B-OBS-OFF, B-BROWSER-OFF (141 tests), B-DOCKER-ONLY, B-SECRET-CANARY (20 secrets), B-SCOPE (42 paths since `0706f9f`), B-REQUIRED-CHECKS (24 IDs of the executed suites); leftover 0 |
| `s7-f1` | S7 F-1 forced failure: W(unit), then W(migrations), in a scratch clone with two flips | `72c6b9c` + flips / `2f6b409` | **Expected FAIL, as required.** W(unit): run `504a97`, exit 1, B-UNIT-CHECKS FAIL on exactly TestOracleMatrix's four `group.update/group_member` subtests, B-REQUIRED-CHECKS FAIL on exactly B-UNIT-CHECKS, 3 other checks PASS. W(migrations): run `c865d2`, exit 1, B-POLICY-IT FAIL on exactly IT-MATRIX-1 (5 cell errors), B-REQUIRED-CHECKS FAIL on exactly B-POLICY-IT, 17 other checks PASS. Both wrapper postchecks passed; 12:34:56–12:37:53Z; leftover 0; clone removed; checkout unchanged |
| `s8-final` | all unit (S8 final) | `9044c87` / `beca305` | **DEFERRED, not a failure: no test ran.** The pre-batch gates passed at 12:41:06Z and the launcher queued on the shared lock behind G051's `k1-c5-o7`. The owner reprioritized G051 (Codex handoff `18b87bd0`, PROGRESS `6a335428`), so at 12:47:36Z the supervisor sent SIGTERM to that childless flock waiter only. `run-approved.sh` and `run.sh` never started and `harness.out` is empty. `runs.tsv`'s `143 … leftover=?` is the killed waiter, not a harness result. Zero `g001.run` resources. Record: `runs/s8-final/DEFERRED.md` (SHA-256 `1ac26408…7667`) |
| `s8-final-2` | all unit (S8 final, the Plan's required final coverage) | `f56b1c1` / `beca305` | **PASS:** run `c85fa5`, binding `g007-binding-reconciled-beca305-aaaafe0ea` (snapshot `e4383b79…`, 93 inputs), launcher 13:22:10Z (lock acquired at once), harness 13:22:11–14:23:20Z, end 14:24:06Z, exit 0, `RESULT: PASSED`, `run-approved: harness exit 0; postcheck passed`; **89/89 PASS** (all 88 full-profile IDs + B-UNIT-CHECKS), no FAIL or SKIP; B-REQUIRED-CHECKS "88 required IDs of the executed suites … each recorded exactly once as PASS"; leftover 0 |
| `s8-final-3` | all unit on the merged harness (staging `7dc3621`, after the G051/G052 integration) | `e3d01d1` / `7dc3621` | FAIL (F11: B-CI-CONTROLS, B-REQUIRED-CHECKS); run `e1b472`, binding `g007-binding-reconciled-7dc3621-aaaafe0ea` (snapshot `360555f5…1b40`, 93 inputs), launcher 14:32:39Z, harness 14:32:40–15:16:02Z, exit 1, `run-approved: harness exit 1; postcheck passed`; **87 PASS, 2 FAIL**: lint 8, lint-base 3, lint-web 5, lint-core 1, provenance 5, contracts 6, root 17, migrations 15 (B-POLICY-CATALOG/MIGRATION/IT/ORDERING pass), ops 12, rename 5, off 4, unit 1, controls 2 of 3 (B-DEPS-CONTROLS, B-WRAPPER-CONTROLS pass), final 3 of 4; E-PG-1 trigger and scan sets and E-PG-2 identical to `s46-clpmo-1`; leftover 0 |
| `a2-c1c2-1` | lint lint-base controls unit migrations (attempt 2: C1 + C2; one G051-granted lock window) | `12d2d8e` / `da1ed80` | **PASS:** run `1eb376`, binding `g007-binding-reconciled-da1ed80-ad4a53b90` (`ff894a8c…`, snapshot `d43b96cf…`, 92 inputs, scope `15b0fe19…`, 15 expressions), launcher 16:03:58Z, harness 16:03:59–16:41:18Z, end 16:41:24Z, exit 0, `RESULT: PASSED`, `run-approved: harness exit 0; postcheck passed`; **34/34 PASS**: lint 8, lint-base 3, controls 3 (B-CI-CONTROLS 439 controls), unit 1, migrations 15 (B-POLICY-IT 38/38 + footprint 2/2, B-POLICY-ORDERING 17/17), final 4 (B-SCOPE 42 paths since `da1ed80`; B-REQUIRED-CHECKS "33 required IDs of the executed suites … each recorded exactly once as PASS"); leftover 0 |
| `s9-ops-1` | lint-base (S9 affected checks for the OPERATIONS.md cell) | `b5a5f40` / `a324b70` | **FAIL, batch composition** (F12): run `35cfc5`, binding `g007-binding-reconciled-a324b70-ad4a53b90` (`9353b7a4…`, 92 inputs), 17:02:54–17:02:58Z, exit 1, postcheck passed; B-LITERALS, B-NAMING, B-IDENTITY-GUARDS, B-DOCKER-ONLY and B-SCOPE (42 paths since `a324b70`) PASS; **B-SECRET-CANARY UNMET** ("no secrets were generated": `lint-base` alone generates no secret) and B-REQUIRED-CHECKS FAIL solely with it; leftover 0 |
| `s9-ops-2` | lint-base unit (the F12 rerun; supervisor-approved) | `f323b30` / `a324b70` | **PASS:** run `12725f`, binding `g007-binding-reconciled-a324b70-ad4a53b90`, 17:04:00–17:04:48Z, exit 0, `RESULT: PASSED`, `run-approved: harness exit 0; postcheck passed`; 8/8 PASS: B-LITERALS (the edited OPERATIONS.md), B-NAMING, B-IDENTITY-GUARDS, B-UNIT-CHECKS, B-DOCKER-ONLY, B-SECRET-CANARY (8 generated secrets), B-SCOPE (42 paths since `a324b70`), B-REQUIRED-CHECKS (7 required IDs); `results.tsv` SHA-256 `f1046fb0…b93a`; leftover 0. `off` is reused (it copies OPERATIONS.md byte-for-byte; the change is a table cell with no executed block). Staging then moved to `65954bc` (records only) after the launch; the supervisor reconciles before the MR |

Development-only (not evidence, no cycle): two supervisor-approved scratch PostgreSQL runs with option (A) applied in
the streamed copy only (`g007-scratch-20261010T195859Z` at `b3e1368`; `g007-scratch-20261010T200905Z` at `0c41f4d`). The
second ran migrate (90), tenancy (72), TestAccessPolicy 38/38, TestAccessFootprint 2/2 and TestAccessOrdering 17/17,
all passing; zero leftovers. The first found four test-side bugs, fixed in `0c41f4d`.

## Deviations from Plan r1 (accepted by the supervisor; recorded here)

- RC-6 lands in S1 with migration 00005, not S2 (Plan §6.6 "same step as the change").
- RC-9's `APPLICABLE` entry for `docs/foundation/AUTHORIZATION.md` lands in S6 with the document (`variants_off` copies
  every APPLICABLE path under `set -e`).

## Execution amendment (appended; Plan r1 bytes unchanged)

Authority: Codex review `432d5f17…` (§1–§4).

1. **Membership-fact locks (§1, option A).** Only `sqlFactOrganizationBindingLock`, `sqlFactGroupBindingsLock` and
   `sqlFactAllGroupBindingsLock` change: they read `app.role_bindings` directly with `organization_id = $1`,
   `subject_user_id = $2`, the exact `scope_type`, `source = 'membership'`, the candidate filter and `ORDER BY group_id`
   where applicable, and end in `FOR SHARE`, returning the same columns and rows as the projection queries through the
   existing UPDATE(role) grant. Snapshot reads stay on the projections; transaction, lock order and modes, the
   selected-Group restriction and deny behaviour are unchanged. No grant, schema, index or lock-mode change.
   - U-BOUNDARY: a name-and-exact-text exception for these three only (`membershipLocks`); statement inventory unchanged
     at 47 statements (13 writes) and 8 pinned locking statements; `TestMembershipLockException` proves that another name,
     a missing tenant/subject/scope/source/candidate filter or order, another scope, other columns and a changed or
     missing lock mode all fail it.
   - CT-POL-3: runtime 42501 through both projections; the three direct scoped locks succeed; runtime writes through
     the projections are refused (42501); the projection and `role_bindings` table and column ACLs are exactly unchanged.
2. **E-PG-1 (§2).** Trigger lines carry `"known"` (true only for the accepted 6b/7 small-table scans); E-PG-1 summaries
   carry the dataset cardinalities; `policy_model` reports total and new trigger counts. E-PG-1..3 are recorded with
   environment, cardinality, rows, buffers and timing, including the compatible-reader bypass of a queued writer
   (E-PG-3a/b) and no fairness or maximum-wait promise. E-PG-2 and every ordering assertion remain mandatory.
3. **Scanner (§3).** `infra/security/gitleaks.toml` carries exactly the G051 O1 bytes (`df25add6…ce45`). The scope
   allow-list is amended to the approved 10 expressions plus `^infra/security/gitleaks\.toml$` (11; SHA-256
   `ab51581bf285be5f5da1cfac0a76ab4673e7e56b912a13d4a0f3faf892d8786d`; previous `be184423…26b6`). If the upstream fix
   integrates first, rebased staging is used instead of a duplicate patch.
4. **Accounting (§4).** Goal-level attempts adopted as recorded above.

## Execution amendment 2 — F5 scanner unpack (appended)

Authority: owner approval via Codex handoff `126605f5`; supervisor-verified minimum correction; still attempt 1.

- **File:** `infra/ci/validate.sh` only. Before: staging bytes SHA-256 `ee57450f96bc5ac18dba339e6cc70b3ea99a45f7a17ce3542d311a325aedd2e9`.
  After: SHA-256 `6ff35c3a1d7a1dc4bd0ed714b73832eee34e37f1a02c0e01c3c1c5986ce3c9c9` = staging plus only G051's V17b hunk: a
  three-line comment and `--mode=u+rwX` on the packing `tar` of the streamed evidence scan (`scan_set`). Diffed against
  the previous file: exactly those two hunks. `validate-controls.sh` and V17b's expectations are unchanged: the scan now
  unpacks a read-only evidence directory and the scrub still fails with 19.
- **Finding:** F5 (`s1-lcm-1`, B-CI-CONTROLS V17b exit 17, want 19), the `bed2622` regression.
- **Reproduction:** the supervisor's `bin/repro-v17b.sh` (pinned scanner image, non-root user, no mounts, `--rm`, 0
  leftovers): without the flag busybox tar reports "Permission denied" for the mode-555 directory's file (unpack rc 1);
  with `--mode=u+rwX` the unpack rc is 0.
- **Portability:** the packing tar is GNU tar in both execution contexts (the hosted context image is
  `buildpack-deps:trixie-scm`; local runs use the host's GNU tar 1.35).
- **Dev check:** shellcheck (pinned image, stdin, severity warning and SC2251): no new finding; the one SC2154 warning
  (`CI_JOBFLOOR_staging`, defined in the sourced `lib.sh`) is identical on the previous bytes.
- **Scope:** the allow-list becomes 12 expressions, adding `^infra/ci/validate\.sh$` (SHA-256
  `67171bac98c429d22b1c3b065c6dc799201f13f705c7ba6a710513c4f80a5476`; previous `ab51581b…786d`, approved `be184423…26b6`).
- **Next:** rerun `lint controls migrations` (re-qualifies the CI-control and lint consumers of the changed file), then
  `root ops rename off`.

The remaining Result sections (implementation summary, files, schema, contracts, tests, commands and outcomes, the
AC1–AC14 checklist, the G008 handoff, PG-001 evidence, cleanup and attempt history) are completed at S9.

## Execution amendment 3 — F6 render register, F7 disposition, interruption (appended)

Authority: G007-G051-G052-followup-codex-review.md (not included) (SHA-256
`92b226dc1a34c5b209de8a69386703e3806b9510beff83f739129870317c95fd`, committed on staging at `3769186`): ACCEPT of the
dispositions, not implementation acceptance. Amendment bytes: `amendment-aaafe0ea/` in the G007 evidence directory.

- **F5:** closed by the review, on the `s1-lcm-2` evidence above.
- **F6 (G051 O6):** disposed by the shared register correction, approved in both G007 and G051.
  - `infra/validation/render-customizations.tsv` is now byte-equal to the authorized register (SHA-256
    `90a258399171f8ef7b05969894c83edb80a090c8beeaead3814fd864ab88d119`; previous `35a52fa5…`): staging's 24 rows plus
    exactly one row, `infra/security/gitleaks.toml` (row SHA-256
    `10ce151639275ca1dc5f638482e5d15d66d1d3ec0c6eb8827706de6e251b8519`), byte-identical to G051's proposal.
  - `infra/validation/run.sh` gains only the count hunk (reference `run.sh.count.diff`, SHA-256
    `6634761362c518cfd41e199dae7e118a4e16d13e2700a64930d37d4a605f8a7d`, applied to G007's own file at line 983): the
    expected register count is 25 instead of 24 and the message adds `+ 1 G051`. No other `run.sh` change.
  - The register moves from the immutable inputs to the mutable scope: 93 immutable inputs; the scope allow-list is 13
    expressions, adding `^infra/validation/render-customizations\.tsv$` (SHA-256
    `34e402e321b58824de494ea0db26432c052f89a8596adb45fc3ef053f8f26ad9`; previous `67171bac…`).
  - Next: the first provenance batch covers the failed-by-prediction B-RENDER-PLAN/B-RENDER-REPEAT and the register's
    consumers, with the covering lint and boundary checks.
- **F7:** statement 6a (Gt, the tenant-and-UUID point lookup and `FOR NO KEY UPDATE` of `group.rename`) is recorded as
  an exact small-fixture PG-001 observation under the owner's unknown-workload disposition. Evidence:
  `runs/s1-lcm-2/report/policy-epg.jsonl` (SHA-256 `97b545e8d90111d992bfb497f78cf4d0b0b59a3b0282efb56c017ac8c874e0ef`).
  Raw trigger lines are preserved, and the reporting code is unchanged (6a still reports `"known": false`). This is not
  production performance acceptance. No query, index or lock change.
- **Interruption:** the worker session reached its usage limit at about 05:25Z while waiting on `s1-rorf-1`, and the
  coordinator session was respawned under the same UUID (owner resume handoff `3535405f`). The batch completed on its own
  under the launcher with its evidence intact (supervisor-verified: no harness processes and zero `g001.run` containers
  left). Nothing was rerun, and no attempt increment was spent: still attempt 1 of B7, checkpoint after attempt 4.
- **Effort:** the owner policy at staging `1a278d1` now defaults implementation to high. The supervisor reports this
  worker's runtime metadata as inherited xhigh; per owner direction it was not restarted solely for effort.

## S5 review and S6 documentation (appended)

- **S5 review (Plan §10):** read-only general-purpose subagent `a666d35a0c268ab51` (effort not set: inherited/default,
  actual unknown; 305,919 tokens, 56 tool uses, 892 s), on HEAD `625a845` and the S6 draft. No writes, builds, docker or
  state-changing git. Verdict: no blocking findings; 2 should-fix (F8, F9 above), 3 advisory. Report copy:
  `s5-review/a666d35a0c268ab51-report.txt` in the G007 evidence directory (SHA-256
  `d08d2342a026e6164516dfa807354e7afe049cf3a0070854b7a5b5e40ecb92d8`); transcript `agent-<id>.jsonl` in the
  session's subagent directory. Checked without issue: lock order and modes, predicates, privileges, B1–B6, error and log
  non-disclosure, the bypass boundary, commit ambiguity, option A, E-PG-2/3 and O-1..O-15.
  - **Advisory 3** (E-PG-1 plans only the locking forms of statements 4–5): stated in AUTHORIZATION.md; no new test.
  - **Advisory 4, open question for Codex (Plan §4.5/Q10):** `commitFailure` (`access/errors.go:82`) treats every
    `*pgconn.PgError` from COMMIT as definitive (`aborted`), including connection-ending FATAL/PANIC errors such as 57P01
    or 57P02, which pgconn also returns as a `PgError`; in a narrow window after a durable commit, only the logged internal
    reason would be wrong. Externally every case is `unavailable`, so no caller can conclude a rollback. AUTHORIZATION.md
    now tells adapters to treat any `unavailable` from a write as possibly durable. Remapping FATAL/PANIC to
    `commit_outcome_unknown` is Codex's call; no code change.
  - **Advisory 5, PG-001/Q14 note:** `SetOrganizationPolicy` locks the Organization row `FOR NO KEY UPDATE` at step 3 and
    Group writes lock the target Group at step 6, both before `decide` (Plan-pinned, "taken up front"). Any subject with
    a valid context in the Organization (for example an `org_member`) can briefly serialize that Organization's guarded
    writes, or a Group's unlink and writes, and observe a weak timing signal of a same-Organization Group's existence.
    Other Organizations' rows are never locked. No change; relevant to future adapter rate limits and deadlines.
- **S6 documentation** (commit `cc9d065`, then the S5 corrections in the following commit): new
  `docs/foundation/AUTHORIZATION.md` (Plan §8 contents, including the G008 handoff); TENANCY.md, MIGRATIONS.md (version
  5), IDENTITY.md (`LockLiveSession`/`CheckLiveSession`) and FOLLOW-UP.md (STAGE-POLICY annotations); `run.sh`
  APPLICABLE gains `docs/foundation/AUTHORIZATION.md` (the recorded RC-9 deviation). The six IDENTITY.md docs blocks that
  harness suites execute are byte-identical before and after (`docs_block` extraction hashes). Development checks (not
  evidence): every changed path inside the 13-expression scope; B-LITERALS terms absent; `bash -n` and shellcheck
  (findings identical to HEAD); gitleaks over the tracked tree: no leaks.
- **Next batch (supervisor direction):** S4's not-yet-run `contracts` combined with S6's `lint provenance migrations off`
  in one batch; Plan §10 lists them separately but does not forbid combining, and no code changes between S4 and S6.
- **Evidence reuse (supervisor direction; not rerun solely for a new SHA).** Since the evidence commits, the only changes
  outside `management/` are documentation (staging's AGENTS.md, CLAUDE.md, PERFORMANCE.md and handoff guide; the G007
  foundation docs), one `render-customizations.tsv` row and two `run.sh` hunks (the provenance count at line 985 and
  APPLICABLE). Consumers of those inputs: APPLICABLE and the register feed `variants_off` (suite `off`, rerun), the
  provenance count and the register feed `suite_provenance` (rerun), the register also feeds B-DEPS through
  `deps_repo_check` (suite `lint`, rerun), and MIGRATIONS.md feeds `migrations` and `off` (rerun).
  - `controls` from `s1-lcm-2` (HEAD `f549283`, base `38ad78c`): `validate-controls.sh` runs the real `run.sh` contracts
    suite on a scratch copy and extracts `static_gitignore`; neither consumes APPLICABLE, the provenance count or the
    register. B-DEPS-CONTROLS reads the register only through `deps_repo_check`, which tests that the
    `web/package.json` and `web/package-lock.json` rows are present (control C11 copies it for that); both rows are
    unchanged, and the one added row (`infra/security/gitleaks.toml`) is outside that test. No `infra/ci/**` input
    changed.
  - `root` and `rename` from `s1-rorf-1` (HEAD `7c2ea2a`, base `38ad78c`): they execute README, GETTING-STARTED and
    WEB-SHELL blocks and build `core`/`web`/`intelligence`; none of those inputs changed.
  - `ops` from `s1-rorf-1`: its only changed input file is IDENTITY.md, read solely through the `identity:start`,
    `identity:migrate`, `identity:browser` and `identity:stop` blocks, which are byte-identical.
  - S8's final `all` + `unit` reruns every suite on the final tree.

## s46-clpmo-1 and O7 (appended)

- **Batch `s46-clpmo-1`** (supervisor-verified independently: `runs.tsv` 10:19:54Z–10:47:29Z, exit 1, HEAD `70e5f5e`,
  postcheck ok, leftover 0; no harness process and zero `g001.run` containers afterwards). Run `ab2634` on base
  `1a278d1`, 93 inputs, snapshot `6994e3c7…63d1`. Per-check outcomes, 49 PASS and 2 FAIL:
  - `contracts` (first G007 run): B-CONTRACT-TOOLS, -GO (120 cases), -PY (120), -AGREE (corpus `a01f1a85…`), -PEER and
    -NEGATIVE PASS.
  - `lint` 17/17 PASS, including B-SHELL, B-LITERALS (4 identity values), B-DEPS, B-MCP-SOURCE, B-MIGRATE-SET and
    B-IDENTITY-GUARDS.
  - `provenance` 5/5 PASS. The F6-related render checks hold: B-RENDER-PLAN (conflict exactly for the 25 registered
    customizations, zero writes) and B-RENDER-REPEAT (only the 25 proposals written); also B-RENDER-ID, B-PROTECTED and
    B-SKILL-UNCHANGED.
  - `migrations` 15/15 PASS: B-MIGRATE-CLI (00001–00005 verbatim), -CATALOG, -ENV, -IT, B-IDENTITY-CATALOG, B-TENANCY-*,
    B-BUSINESS-PROFILE-*, B-POLICY-CATALOG 6/6, B-POLICY-MIGRATION 5/5, B-POLICY-IT 38/38 plus TestAccessFootprint 2/2
    (590 `access.decision` records catalogued, canaries absent, planted canary detected) and B-POLICY-ORDERING 17/17.
    E-PG-1 is unchanged from `s1-lcm-2`: the same 6 trigger lines (the 2 `"known": false` lines are the disposed F7 6a
    scans) and the identical F9 sequential-scan set (`policy-epg.jsonl` SHA-256
    `23541e829056cc85d6596916ff09e71d72ad48f77f8030710b70cc0172c724c3`).
  - `off`: B-TOGGLE-OFF, B-OBS-OFF and B-BROWSER-OFF (141 tests) PASS; **B-TOGGLE-OFF-ADDITIONS FAIL** (F10, harness.out
    line 57).
  - B-DOCKER-ONLY, B-SECRET-CANARY and B-SCOPE (42 changed paths since `1a278d1`, all in the allow-list) PASS;
    **B-REQUIRED-CHECKS FAIL** solely because of B-TOGGLE-OFF-ADDITIONS.
  - `results.tsv` SHA-256 `de019a7e1eb8a4da1f54d11ff73533bd2d84225802ef0fb47728a793217ccc74`.
- **Evidence reuse** is unchanged from "S5 review and S6 documentation" above: `controls` from `s1-lcm-2`; `root`,
  `rename` and `ops` from `s1-rorf-1`; with the provenance recorded there.
- **F10 / O7 proposed lines** (not applied; Codex approval pending). G007's APPLICABLE would end:

  ```text
    docs/foundation/AUTHORIZATION.md
    # G051 O1/O6: the customized secret-scan configuration (the exact credential.go provenance-digest exception).
    infra/security/gitleaks.toml)
  ```

  This replaces the current closing `docs/foundation/AUTHORIZATION.md)`. The two appended lines are byte-identical to
  G051's reference hunk (`b3d3109d…5531`); only the preceding line differs.
  - **Rerun plan after approval and the supervisor's rebase/rebind:** `off`, plus the required-check evaluation.
    `contracts`, `lint`, `provenance` and `migrations` are reused from `s46-clpmo-1`, and `controls`, `root`, `rename`
    and `ops` as recorded, each with exact provenance. The APPLICABLE change is consumed only by `variants_off`.
  - S7 (F-1) and S8 (final `all` + `unit`) wait under the batch rule. F8 and F9 remain pending Codex.
- **Stop and resume.** After `s46-clpmo-1` completed, the worker reached a weekly usage limit while recording the
  outcome. The uncommitted Result bookkeeping and all batch evidence were intact; nothing ran, and nothing was lost. The
  owner lifted the stop with a usage upgrade (Codex handoff `43d2d6eb`), and work resumed under the same coordinator UUID.
  Still attempt 1 of B7, no increment; checkpoint after attempt 4.
- **Staging** is now `0706f9f` (management-only: PROGRESS and the G052 review). The supervisor rebases and rebinds at the
  next pre-batch boundary, after Codex rules on O7. No `run.sh` edit, rebase or launch has been made.

## Execution amendment 4 — O7/F10 APPLICABLE entry (appended)

Authority: the section "G051 O7 / G007 shared toggle-off correction — 2026-10-11" appended to
G007-G051-G052-followup-codex-review.md (not included) (SHA-256
`5cd3e38392e477f51d59d209cfb426506f1963db27b2e3ad51ec35cb1a94ef1f`; not yet on staging, which carries `92b226dc…`).
Supervisor handoff `b85a6792`, with the exact bytes preserved at `amendment-b85a6792/` in the G007 evidence directory.
This is an ACCEPT of the bounded correction, not final Goal acceptance. It supplements O6's count-only `run.sh`
authorization, and every other scope restriction remains.

- **Change:** `infra/validation/run.sh` APPLICABLE gains exactly one `infra/security/gitleaks.toml` entry, with the G051
  O1/O6 comment, after G007's last entry `docs/foundation/AUTHORIZATION.md`. The closing parenthesis moves to the new
  last line, and every other entry is unchanged. The two added lines are byte-identical to the reference patch's added
  lines (`run.sh.applicable.diff`, SHA-256 `b3d3109d90c7e11a9d3ce62ed8330050690986f5ad8bb2be4f91e0a879295531`). They were
  applied to G007's own `run.sh`; no file was copied between checkouts. The entry occurs once.
  - `run.sh` SHA-256: `9efc354147c6214f7bfa5efcdd4557e7be4969bcf04709008f813a3e68f27d68` (previous `35886a9a…`).
  - The `git diff` of `run.sh` shows only this change.
  - Development checks: `bash -n` OK; shellcheck findings identical to HEAD; scope check clean.
- **No other change:** scanner rules, comparisons, the register and the count are unchanged.
- **Next batch (after the supervisor's rebase onto `0706f9f` and rebind recording this amendment):** `lint off`. That is
  the failed `off` suite, its required-check evaluation, and the covering shell and static checks for the changed list.
  - The report will verify that the variant's scanner configuration equals the approved root bytes `df25add6…`, and
    keep the manifest and customization comparisons and the cleanup evidence.
  - Reused with explicit provenance: `contracts`, `provenance` and `migrations` from `s46-clpmo-1`; `controls` from
    `s1-lcm-2`; `root`, `rename` and `ops` from `s1-rorf-1`.
  - S7 (F-1) and S8 (final `all` + `unit`) follow.
- **Accounting:** still attempt 1 of B7, checkpoint after attempt 4. This correction consumes no new attempt.

## s7-lo-1: the F10/O7 rerun (appended)

- **Batch** `s7-lo-1` (suites `lint off`; the step name is a label, not Plan step S7). Rebased HEAD `a315a59` (tree
  `c5e7d080`), base `0706f9f`. Binding `g007-binding-reconciled-0706f9f-aaaafe0ea` (`binding-reconciled.json`
  `b6b580f9…ce65b`, which records amendment b85a6792), mount safeguard 2%. Run `1d21ae`, 12:14:27Z–12:26:53Z, exit 0;
  `results.tsv` SHA-256 `f89366c83e5816cd38fd39752ab824d841cd06e07baf1abf20ef8520bed610ea`. All 25 checks PASS.
  - Lint, 17/17: B-SHELL, B-HADOLINT, B-COMPOSE, B-ENV-EXAMPLE, B-LITERALS, B-NAMING, B-DEPS, B-WEB-BOUNDARY,
    B-IDENTITY-GUARDS, B-WEB-TOKENS, B-WEB-COMPONENTS, B-WEB-DESIGN, B-WEB-ASSET-CONTROLS, B-MCP-SOURCE, B-MIGRATE-SET,
    B-DOCKERIGNORE and B-CI-CONFIG.
  - Off: B-TOGGLE-OFF-ADDITIONS, B-TOGGLE-OFF, B-OBS-OFF and B-BROWSER-OFF (141 tests).
  - Final: B-DOCKER-ONLY, B-SECRET-CANARY, B-SCOPE, and B-REQUIRED-CHECKS ("24 required IDs of the executed suites (lint
    off, final) each recorded exactly once as PASS").
- **B-TOGGLE-OFF-ADDITIONS:** "variant `850b7f2aa96f0986` keeps its own manifest (no root manifest); its rendered files
  match it except the registered customizations, whose rendered base hashes equal the root manifest's; the
  toggle-rendered .env.example and VENDORS.md are the variant's own render … (controls: the uncustomized render and an
  altered customization are rejected); all 54 applicable additions byte-identical to the root". 54 is the APPLICABLE
  count with the new entry (53 before).
- **The variant's scanner configuration equals the approved root bytes.** The harness does not print the variant file's
  hash, and the disposable work tree is removed. Equality follows from the check semantics:
  - `variants_off` copies each APPLICABLE item from the repository (`run.sh:2654`).
  - For a registered customization, it requires the variant manifest's rendered hash to equal the root manifest's, and
    the variant file to equal the repository file (`run.sh:2689`).
  - Its additions loop requires every file of every APPLICABLE item to be byte-identical to the repository's
    (`run.sh:2697`).

  The repository's `infra/security/gitleaks.toml` at `a315a59` hashes to
  `df25add66bcf6819e0e94066d318f8692adcdbc692989b26087ceb4b483ece45`. With the check passing, and no "customized … differs"
  or "addition … differs" line in `harness.out` (`s46-clpmo-1` logged the former at line 57), the variant's bytes are
  `df25add6…`.
- **Cleanup:** `inventory-leftover.txt` is empty; zero `g001.run` containers remain; `runs.tsv` records `leftover=0`; and
  the wrapper postcheck passed.
- **Reuse with explicit provenance (supervisor direction and Codex review `5cd3e383`):**
  - `contracts`, `provenance` and `migrations` from `s46-clpmo-1` (HEAD `70e5f5e`, base `1a278d1`). Since then, the only
    changes are the APPLICABLE entry, which only `variants_off` consumes; Result text; and staging's management-only
    advance (PROGRESS.md, the G052 review). None is an input of those suites, and B-SCOPE reran on the new base.
  - `controls` from `s1-lcm-2`, and `root`, `rename` and `ops` from `s1-rorf-1`, with the provenance recorded in "S5
    review and S6 documentation". The APPLICABLE entry is not consumed by them.
- **Next:** S7 F-1, on the supervisor's go. Attempt 1 of B7.

## S7: F-1 forced failure (appended)

Plan r1 §6.5: "in a scratch clone, flip one U-POL-MATRIX cell and one IT-MATRIX expectation. `W(unit)` and
`W(migrations)` must FAIL with exactly those IDs, combined with the inherited B-CI-CONTROLS gate controls." Supervisor
go after `s7-lo-1`; attempt 1 of B7.

- **Gates and clone.**
  - `g007-vw.sh check` passed at 12:33:33Z: binding `g007-binding-reconciled-2f6b409-aaaafe0ea`, target `2f6b409` equal
    to the binding base and an ancestor of HEAD `72c6b9c`, clean checkout, mounts 2%, wrapper precheck 93 inputs.
  - The scratch clone was created by a plain `git clone --no-hardlinks --branch g007-trusted-context-policy` of this
    checkout, at `runs/s7-f1/clone` in the G007 evidence directory.
  - The runner (`scratch/f1-run.sh`, with `f1-flips.py`, `f1-tree.py` and `f1-w.sh`) repeated the launcher's gates and
    ran git only in the checkout. It proved the clone equal to HEAD (882 tracked files identical, by content or symlink
    target, and no extra file), applied the flips, confirmed both flipped files are gofmt-clean, and checked the scope
    copy.
  - The first runner start, at 12:34:23Z, stopped at its own identity check before any flip or harness run: it hashed a
    tracked directory symlink, `.agents/skills/vestive-design`. After the fix (symlink targets compared), the run
    started at 12:34:56Z. Both starts appear in `prebatch.tsv`.
- **Flips** (`runs/s7-f1/flips.diff`, SHA-256 `418cc4d560aa9d3c18b127bc3799a23785bb4a1e0709c71e96e975ae4128b7f6`):
  1. **One U-POL-MATRIX cell.** The Spec R3 cell `group.update × group_member` was flipped from deny to allow. It is
     restated in all four primary oracle tables (`matrix_test.go`: Organization-only and selected-g1, flag on and off),
     so TestOracleTables' shape and difference checks still hold and only the cell disagrees with `decide`.
  2. **One IT-MATRIX expectation.** IT-MATRIX-1 (`integration_test.go`) inverts its expectation for
     `organization.manage_policy × org_owner`, in Organization-only context with the flag on.

  IT-MATRIX-1 applies the oracle's table (Plan §6.3, "the same independent 11×6 table"). Flip 1 therefore also flips
  IT-MATRIX-1's expectation for that cell, so the migrations failure shows both flips.
- **W(unit)** (wrapper from the clone, under the shared lock; run `504a97`; exit 1; `run-approved: harness exit 1;
  postcheck passed`; `results.tsv` SHA-256 `55cdd1a4…3295`):
  - B-UNIT-CHECKS FAIL. core-check's tests fail only in `core/internal/tenancy/access`, and only in
    `TestOracleMatrix/{org-only,selected-g1}/{flag-on,flag-off}/group.update/group_member`. Each subtest reports
    `decide("group.update") … allow = false, want true` for the complete and the loaded facts. Every other core package
    passes, and TestOracleTables passes.
  - B-REQUIRED-CHECKS FAIL: "B-UNIT-CHECKS(1,FAIL)" only.
  - B-DOCKER-ONLY, B-SECRET-CANARY and B-SCOPE (42 paths) PASS.
- **W(migrations)** (run `c865d2`; exit 1; postcheck passed; `results.tsv` SHA-256 `3c0fb728…4b81`):
  - B-POLICY-IT FAIL. In TestAccessPolicy, only `IT-MATRIX-1` fails, with exactly five errors:
    - `organization.manage_policy / org_owner (selected=false, flag=true): allow=true, want false` (flip 2);
    - `group.update / group_member: allow=false, want true` in all four context-and-flag modes (flip 1).

    The other 37 subtests PASS. The harness labels both counts "FAIL/38 policy, FAIL/2 footprint". That is its shared-log
    count (`it_count`, `run.sh:1887`: any `--- FAIL` in the one log marks both), not a footprint failure.
    TestAccessFootprint ran in full: 64 E-PG-1 lines, the 6 trigger lines, and 24 E-PG-2 lines byte-identical to
    `s46-clpmo-1`'s (`policy-epg.jsonl` SHA-256 `8bfa3644…8eb4`).
  - B-REQUIRED-CHECKS FAIL: "B-POLICY-IT(1,FAIL)" only.
  - The other 17 checks PASS:
    - B-MIGRATE-CLI, -CATALOG, -ENV, -IT and B-IDENTITY-CATALOG;
    - B-TENANCY-* and B-BUSINESS-PROFILE-*;
    - B-POLICY-CATALOG, -MIGRATION and -ORDERING (17/17);
    - B-DOCKER-ONLY, B-SECRET-CANARY and B-SCOPE.
- **Gate controls:** the inherited B-CI-CONTROLS (291/291 in `s1-lcm-2`) prove that any failing required check fails
  `infra/ci/validate.sh` and the gating job.
- **Cleanup:** both `inventory-leftover.txt` files are empty. Zero `g001.run` containers, volumes and networks and zero
  `g007.devcheck` containers remain. The clone was removed at 12:37:53Z. The checkout is unchanged (HEAD `72c6b9c`,
  clean). `runs.tsv` records `s7-f1 … unit_rc=1 migrations_rc=1 … checkout-unchanged leftover=unit:0 migrations:0`.
- **Outcome:** F-1 meets Plan §6.5. Each wrapper run fails with exactly the flipped IDs and nothing else. Next: S8 final
  (`all` + `unit`) on the supervisor's go.

## S8 deferred (appended)

- `s8-final` (`all unit`) on HEAD `9044c87`, binding `g007-binding-reconciled-beca305-aaaafe0ea`, passed its pre-batch
  gates at 12:41:06Z: target `beca305`, mounts 2%. It queued on the shared lock behind G051's `k1-c5-o7`.
- The owner then prioritized finishing G051 first (Codex handoff `18b87bd0`, PROGRESS `6a335428`). At 12:47:36Z, after
  confirming the waiter had no child, the supervisor sent SIGTERM to the `s8-final` flock waiter (PID 94906) only.
  - `run-approved.sh` and `run.sh` never started; `harness.out` is empty.
  - The launcher's exit 143 and its "leftover=?" line come from the missing report directory of a run that never
    began. No owned resource existed: zero `g001.run` containers, volumes and networks. The only harness process
    afterwards was G051's.
  - Staging was unchanged at `beca305` (`git ls-remote`, 12:48:20Z).
  - Supervisor record: `runs/s8-final/DEFERRED.md` in the G007 evidence directory (SHA-256
    `1ac26408cf1905c863d23ac276253941c99da30603c4f427ecb9b373dd897667`).
- This is a scheduling deferral, not a test failure; no attempt is spent (attempt 1 of B7). S7 F-1 and all earlier
  evidence stay valid.
- On resume, after G051's local batches, the supervisor fetches, rebases and rebinds onto the then-current staging and
  assigns a new S8 step name (for example `s8-final-2`). S8 then runs `all unit` as the Plan's required final coverage.

## S8: final local qualification (appended)

- **Batch `s8-final-2`** (`all unit`; Plan §10 S8: "all 88 full IDs + B-UNIT-CHECKS PASS").
  - HEAD `f56b1c1`, base `beca305`. Staging was unchanged since the last rebind, so no rebase or rebind was needed.
  - Pre-batch gates passed at 13:22:10Z with mounts at 3%. Run `c85fa5` executed the suites lint, controls, provenance,
    contracts, root, migrations, ops, rename, off and unit, plus final.
  - Exit 0, wrapper postcheck passed. `results.tsv` SHA-256
    `adcff1d80f8bc5496448acd12351da8e002f9b4b5da6506804ed1cf6cc0c7ad9`; `required-checks.tsv` SHA-256
    `7abe6d5b10f69adeaa14c4a912ea3f2481c88ed6822c394ba42acbbcc3df224d`.
- **Per check, 89/89 PASS:**
  - lint 17: B-SHELL, B-HADOLINT, B-COMPOSE, B-ENV-EXAMPLE, B-LITERALS, B-NAMING, B-DEPS, B-WEB-BOUNDARY,
    B-IDENTITY-GUARDS, B-WEB-TOKENS, B-WEB-COMPONENTS, B-WEB-DESIGN, B-WEB-ASSET-CONTROLS, B-MCP-SOURCE, B-MIGRATE-SET,
    B-DOCKERIGNORE, B-CI-CONFIG.
  - controls 3: B-DEPS-CONTROLS, B-WRAPPER-CONTROLS, B-CI-CONTROLS (291 controls).
  - provenance 5: B-RENDER-ID, B-RENDER-PLAN (25 registered customizations), B-RENDER-REPEAT, B-PROTECTED,
    B-SKILL-UNCHANGED.
  - contracts 6: B-CONTRACT-TOOLS, -GO, -PY, -AGREE, -PEER, -NEGATIVE.
  - root 17: B-CONFIGURE, B-DEFAULTS, B-PORTS-ROOT, B-BOUNDARIES-ROOT, B-BOUNDARY-CONTROLS, B-MCP-HTTP, B-TENANCY-HTTP
    (70 route-absence probes), B-WEB-HTTP, B-HEALTH-ROOT, B-BROWSER-ROOT, B-CHECKS-ROOT, B-CORE-ERROR-OUTPUT,
    B-WEB-ERROR-OUTPUT, B-INTEL-ERROR-OUTPUT, B-DOCS-EXEC-GS, B-BUILD-CONTEXT, B-DOCS-EXEC-ROOT.
  - migrations 15: B-MIGRATE-CLI, -CATALOG, -ENV, B-IDENTITY-CATALOG, B-MIGRATE-IT, B-TENANCY-CATALOG, -MIGRATION, -IT,
    B-BUSINESS-PROFILE-CATALOG, -MIGRATION, -IT, B-POLICY-CATALOG (6/6), B-POLICY-MIGRATION (5/5), B-POLICY-IT (38/38,
    plus TestAccessFootprint 2/2), B-POLICY-ORDERING (17/17).
  - ops 12: B-OBS-CONTEXT, -CHAIN, -TIMEOUTS, -LOGS, -PRIVACY, B-OUTAGE, B-PERSIST, B-IDENTITY-IT, -E2E, -PROBES,
    -PRIVACY, -CONFIG.
  - rename 5: B-SOURCE-HASH, B-PORTS-OVERRIDE, B-RENAME, B-BROWSER-RENAME, B-DBRENAME.
  - off 4: B-TOGGLE-OFF-ADDITIONS (all 54 applicable additions identical), B-OBS-OFF, B-BROWSER-OFF, B-TOGGLE-OFF.
  - unit 1: B-UNIT-CHECKS (core-check with gofmt, vet and unit tests including U-POL-MATRIX, U-CTX and
    U-ACCESS-BOUNDARY; web-check; intelligence-check; secret-scan).
  - final 4: B-DOCKER-ONLY, B-SECRET-CANARY, B-SCOPE (42 changed paths since `beca305`, all in the allow-list),
    B-REQUIRED-CHECKS.
- **PG-001 evidence is unchanged.** E-PG-1 has the same 6 trigger lines (the 2 lines with `"known": false` are the
  disposed F7 6a scans), and the sequential-scan set is identical to F9's. E-PG-2 lines are byte-identical to
  `s46-clpmo-1`'s. `policy-epg.jsonl` SHA-256: `430c810c1e9b8c493199f505118b43cf11dd1dc2a2f78159d6a19c5e03e05766`.
  F8 and F9 remain pending Codex.
- **Cleanup:** `inventory-leftover.txt` is empty, and the launcher's label check at 14:24:06Z found zero `g001.run`,
  `g003.controls.run` and `g003.outer` containers.
  - A `g003.controls.run` container seen afterwards was created at 14:24:36Z. It belongs to G051's next batch
    (`c1-a6-cleanup-u`, which took the lock after this run released it), not to G007.
  - Zero `g001.run` volumes and networks.
- `s8-final` stays recorded as DEFERRED. Next: Codex's actual-diff review. No push or MR without authorization.

## Merge reconciliation onto staging 7dc3621 and s8-final-3 (appended)

- **Rebase** (supervisor direction): `75398bb` onto `7dc3621` (G051 harness merge `c84a633`, G052 `8ae2b6a`, records)
  gave HEAD `e3d01d1` (tree `07b7e687`), 24 commits replayed. Conflicts occurred only in the two expected files:
  - `infra/validation/required-checks.tsv` header, at S1 (2 G007 rows) and S4 (4 rows): G051's header text (lint
    relabel, unit/feature IDs) combined with G007's rows note: "+ 3 G006 + 5 G051 + 4 G007 = 95 rows". The profile
    counts were derived with `scratch/profile-counts.py`: rows per `CI_SUITES_<profile>` of `infra/ci/lib.sh` plus the 4
    final rows. That script reproduces staging's 91 rows and fast 33, integration 71, full 84, staging 5. The merged file
    has 95 unique rows: all 91 staging rows once plus the 4 B-POLICY-* rows (suite `migrations`, which only full lists),
    so fast 33, integration 71, full 88, staging 5. SHA-256 `e2a64f2f…3303`.
  - `infra/validation/run.sh` APPLICABLE tail, at S6: G007's authorization-guide entry, then the single shared G051
    O1/O6 `infra/security/gitleaks.toml)` entry (once). Amendment 4's identical change then merged cleanly. SHA-256
    `c7214016…7ec8`. Against staging, `run.sh` changes exactly G007's original hunks minus the three parts staging
    already carries: the closing-parenthesis change, the scanner entry, and the `== 25` count line (unchanged). That is
    172 of the original 178 changed lines; no G051 content is reverted. `bash -n` passes; shellcheck findings are
    identical to staging's (3).
  - `infra/ci/validate.sh`, `infra/security/gitleaks.toml` and `infra/validation/render-customizations.tsv` now equal
    staging. G051 carries the same `--mode=u+rwX` fix, so the F5 hunk drops out of G007's diff.
  - 39 of the 42 G007 files are byte-identical to `75398bb`: every implementation, test, migration, doc and Result
    file. 39 paths differ from staging, all inside the 13-expression scope.
- **Rebind** (supervisor): binding `g007-binding-reconciled-7dc3621-aaaafe0ea` (`dce948bb…`), snapshot `360555f5…`,
  93 entries.
- **Batch `s8-final-3`** (`all unit` on the merged harness): the target changed the harness every suite runs through
  (`run.sh`, `validation/lib.sh`, `validate.sh`, the CI controls, the lint split, image reuse and cleanup, the browser
  configuration). The affected set is therefore every suite; this is not a restart for a new SHA.
  - 87 checks pass. B-CI-CONTROLS fails on VA7, VU1 and J8 (F11), and B-REQUIRED-CHECKS fails with it.
    `results.tsv` SHA-256 `f32919a4871973f380970a30ac984c8a6f75017f1264847aa6aa57ccf0d78029`; `ci-controls.log`
    SHA-256 `9571cb6a28acd74f83020e5f60181101acae976028e3b85bce1105b17558d515`.
  - Cleanup: `inventory-leftover.txt` is empty, and zero `g001.run`, `g003.controls.run` and `g003.outer` resources
    remain.
- `s8-final-2` stays recorded as pre-merge-harness evidence at `beca305`. S7 F-1, the S5 review and earlier evidence
  keep their recorded provenance. F8, F9 and advisory 4 remain pending Codex. Attempt 1 of B7.

## Attempt 2 — C1 count reconciliation and C2 COMMIT classifier (appended)

Authority: G007-integration-findings-codex-review.md (not included), SHA-256
`1bed0ddbd2bdec61d42bc0fc40a445928645dcaae0095c0baf718bad429f6706`. This is a CORRECTION LOOP, not implementation
acceptance. Supervisor handoff `d4a53b90`; bytes preserved at `amendment-d4a53b90/` in the G007 evidence directory.
Attempt 2 of B7 covers C1 and C2 together. `s8-final-3` stays recorded as failed (87/89).

- **Scope and input amendment (C1).**
  - `infra/ci/validate-controls.sh` moves from the immutable inputs to the scope: 92 immutable paths
    (`immutable-paths.txt`, SHA-256 `661f084a180b88d70afd949dd6bc69b35f7b6e2188ff69ea796c9a766f8b1074`).
  - The scope gains that exact path and `docs/foundation/OPERATIONS.md`, 15 expressions. The in-repository
    `management/v0/results/G007-scope-allowlist.txt` is byte-identical to `amendment-d4a53b90/scope-allowlist.txt`
    (SHA-256 `15b0fe19b278d31d916a561001d4dd2575b93e34360434b61abe230ff4993c1b`; previous `34e402e3…`).
  - The supervisor binds it with `g007-reconcile5.sh` and `g007-bind6.py`.
- **C1, exact counts (F11).** Numbers and detail text only, at the four authorized control sites and one doc row:
  - `infra/ci/validate-controls.sh`:
    - VA7: `required_count 56` → `60`, detail "56 required IDs" → "60 required IDs".
    - VU1: `vu_split full 33 56 85` → `33 60 89`, detail "fast 33 + gate 56 -> 85" → "fast 33 + gate 60 -> 89".
    - J8: `(( n != 56 ))` → `60`, "want 56" → "want 60".
  - `docs/foundation/OPERATIONS.md` full coverage row: 33/56/85 → 33/60/89, and "the 84 full IDs" → "the 88 full IDs".

  The fast and integration numbers and logic are unchanged. No exact cardinality was replaced by a computed value, and
  no control was removed. OPERATIONS.md's profile table above (line 93, the full profile's "Required IDs" cell, still
  84) is outside the authorized sites and is left for the supervisor's decision.
- **C2, COMMIT classifier (advisory 4).** `core/internal/tenancy/access/errors.go` `commitFailure` claims a definitive
  abort only when one is established:

  | COMMIT failure | Internal reason |
  |---|---|
  | Proven unsent (`pgconn.SafeToRetry`) | `failureReason`: `canceled` if the context is done, otherwise `store_unavailable` |
  | `pgx.ErrTxCommitRollback` (the server answered ROLLBACK), wrapped or not | `aborted` |
  | `*pgconn.PgError` with severity ERROR, wrapped or not: `SeverityUnlocalized`, or the localized `Severity` only when no nonlocalized severity was sent; exact match | `aborted` |
  | `*pgconn.PgError` with FATAL (for example 57P01) or PANIC (for example 57P02), including a nonlocalized FATAL with a localized ERROR | `commit_outcome_unknown` |
  | `*pgconn.PgError` with a missing, translated-only, lowercase or other severity | `commit_outcome_unknown` |
  | Any other error: EOF, unexpected EOF, connection loss, cancellation or deadline during COMMIT | `commit_outcome_unknown` |

  - The public code stays `unavailable` with no field and no driver text. There is still no automatic retry. The
    statement-level class-40 mapping (`failureReason`) and the reason set are unchanged.
  - `tx.go`'s comment no longer says a failed COMMIT rolls back.
  - No new dependency, vendor, restart test or transaction redesign.
- **Tests (C2):**
  - New unit file `core/internal/tenancy/access/commit_test.go` (U-COMMIT) covers 21 cases: ordinary ERROR 40001,
    23505, wrapped 40P01, translated and localized-only ERROR, ROLLBACK reply (plain and wrapped), FATAL 57P01 (plain,
    wrapped, nonlocalized FATAL with localized ERROR), PANIC 57P02 (plain, wrapped), no, unrecognized, translated-only and
    lowercase severity, EOF, unexpected EOF, cancellation during COMMIT, unsent, and unsent after cancellation.
    - `TestCommitClassification` asserts code `unavailable`, no field, the exact reason, the error text `access:
      unavailable`, and no driver text.
    - `TestCommitFailureTransaction` runs each case through `inTx` with a stub database. It asserts one BeginTx, one
      Commit (no retry) and the deferred rollback. It also asserts that the `access.decision` record holds only time,
      level, msg, operation, capability, outcome and reason, with outcome `failed`, the exact reason and no driver text.
  - `integration_test.go`: the hook's injected server-rejected COMMIT (IT-FACADE-2) now carries severity ERROR
    (nonlocalized and localized), as a real ordinary command error does. The real definitive-abort (IT-FACADE-2) and
    durable lost-reply (IT-FACADE-3) integration tests are otherwise unchanged.
- **Docs.** AUTHORIZATION.md:
  - the caller table separates "never sent", "ROLLBACK reply or ordinary ERROR" and "FATAL/PANIC, unknown severity or
    lost reply", and states that a server error alone is never a confirmed rollback;
  - the deadlock paragraph names ordinary (severity ERROR) 40P01/40001;
  - the F8 limit is stated: the unchanged-slug footprint did not measure the upgrade.
- **F8, F9:** ACCEPTED as dispositions (§F8, §F9). No footprint variant, benchmark, or lock, query or index change.
- **Development checks** (no Docker, harness or `go test` under the test hold): `bash -n` on `validate-controls.sh`;
  scope check against the 15 expressions; text checks. gofmt-sensitive constructs (aligned fields, doc comments) were
  checked by hand; gofmt and the tests run with the released batch.
- **Proposed minimum batch once the hold lifts:** `g007-vw.sh <step> lint lint-base controls unit migrations`, plus
  final.
  - `lint` re-shellchecks the edited `infra/ci/validate-controls.sh` (B-SHELL covers `infra/ci`).
  - `lint-base` re-scans the edited `docs/foundation` files (B-LITERALS).
  - `controls` covers C1: VA7, VU1 and J8 with the inherited union and completeness negative controls.
  - `unit` covers C2's U-COMMIT, gofmt, vet and the core unit tests.
  - `migrations` covers B-POLICY-IT, with IT-FACADE-2/3 on the changed classifier.
  - Not proposed: outage, rename or browser suites. `off` reads `docs/foundation/OPERATIONS.md` only as an APPLICABLE
    copy (`cp -a`, then a byte comparison with the root), so that edit cannot change its outcome. Reusing `off` from
    `s8-final-3` is the supervisor's decision.
  - The other passing `s8-final-3` checks are reused with exact provenance.

## a2-c1c2-1: attempt-2 validation (appended)

- **Batch `a2-c1c2-1`**, run in the single lock window that G051 granted. HEAD `12d2d8e`: the supervisor rebased
  `6983fe8` onto staging `da1ed80`, whose delta is records only; tree `74f4c538`. Binding
  `g007-binding-reconciled-da1ed80-ad4a53b90`, recording amendment `d4a53b90`. Mounts 4%. Exit 0, wrapper postcheck
  passed. `results.tsv` SHA-256 `90e4036f28fe0ad44a3d1c750d8e12f9194ce5fa124cf81d64d07b8ded96df77`; `ci-controls.log`
  SHA-256 `ac97d60a64d9d4ed5c5cf4e0cd413201b6c5813d1632f45feaaee11aa159899c`.
- **Per check, 34/34 PASS:**
  - lint (8): B-SHELL re-shellchecked the edited `infra/ci/validate-controls.sh`; also B-HADOLINT, B-COMPOSE,
    B-ENV-EXAMPLE, B-DEPS, B-MIGRATE-SET, B-DOCKERIGNORE, B-CI-CONFIG.
  - lint-base (3): B-LITERALS over the edited `docs/foundation` files; B-NAMING, B-IDENTITY-GUARDS.
  - controls (3): B-DEPS-CONTROLS, B-WRAPPER-CONTROLS, and B-CI-CONTROLS at 439/439, zero FAIL. VA7 reports "60
    required IDs", VU1 "fast 33 + gate 60 -> 89", VU9 is unchanged at "fast 33 + gate 42 -> 71", and J8 reports "60
    check cases" (C1).
  - unit (1): B-UNIT-CHECKS. Core-check's gofmt, vet and unit tests include U-COMMIT's `TestCommitClassification` and
    `TestCommitFailureTransaction`; web-check, intelligence-check and secret-scan also pass (C2).
  - migrations (15): B-MIGRATE-*, B-IDENTITY-CATALOG, B-TENANCY-*, B-BUSINESS-PROFILE-*, B-POLICY-CATALOG,
    B-POLICY-MIGRATION, B-POLICY-IT (38/38, including IT-FACADE-2 aborted and IT-FACADE-3 commit_outcome_unknown on the
    new classifier; TestAccessFootprint 2/2) and B-POLICY-ORDERING (17/17) (C2).
  - final (4): B-DOCKER-ONLY, B-SECRET-CANARY, B-SCOPE, B-REQUIRED-CHECKS.
- **PG-001:** E-PG-1 trigger and scan sets and the E-PG-2 lines are identical to `s46-clpmo-1`'s. `policy-epg.jsonl`
  SHA-256 `e6fc938d74ba31c1bd086b0224324adf49ef6f76bfb4f9e6b44a288565a17c67`.
- **Cleanup:** `inventory-leftover.txt` is empty, and zero `g001.run`, `g003.controls.run` and `g003.outer` containers,
  volumes and networks remain. The lock was released at the end, and G007 is back on hold until G051 releases priority.
- **Evidence reused from `s8-final-3`** (HEAD `e3d01d1`, base `7dc3621`), by supervisor direction. Since then the
  changed inputs are: `infra/ci/validate-controls.sh` (C1); `docs/foundation/OPERATIONS.md` and AUTHORIZATION.md;
  `core/internal/tenancy/access/{errors,tx}.go`, the new `commit_test.go` and `integration_test.go` (C2); the scope
  allow-list and Result; and staging's record-only `da1ed80`.
  - `lint-web`: no web input changed.
  - `lint-core` (B-MCP-SOURCE): no production import changed. `errors.go` and `tx.go` keep their import sets,
    `commit_test.go` is a test file, and the access package is still linked into no binary. U-ACCESS-BOUNDARY
    re-verified the imports in `unit`.
  - `provenance`: neither edited file is archetype-rendered (absent from the render manifest), so the render plan,
    repeat render, protected and skill checks have unchanged inputs.
  - `contracts`: no contract or validation-tool input changed.
  - `root`, `ops`, `rename`: the built binaries are unchanged, because the access package is linked into no binary.
    B-CHECKS-ROOT runs the same README check commands that B-UNIT-CHECKS reran on this HEAD, and their README and
    GETTING-STARTED blocks are unchanged.
  - `off`: it reads `docs/foundation/OPERATIONS.md` and `core/internal/tenancy` only as APPLICABLE copies (`cp -a`,
    then a byte comparison with the root), so these edits cannot change its outcome; its binaries are unchanged as
    above.
- **C1 and C2 are complete** under attempt 2 of B7. F11 is closed, and F8 and F9 are accepted. Advisory 4 is resolved
  by C2.
  - OPERATIONS.md line 93 (the full profile's "Required IDs" cell, 84) remains pending Codex and is unchanged.
  - Next: Codex's actual-diff and readiness review. No push or MR without authorization.

## S9: submission status and AC1–AC14 evidence (appended)

**Current submission status.** Codex accepted actual-code and local-evidence readiness at candidate `c151407`, with the
tested implementation at `12d2d8e`. The authority is
[G007-implementation-codex-review.md](G007-implementation-codex-review.md), SHA-256
`f62c1927e33fa0171e8197f02a578d668958293f2117f15338ec7b70933afa94` (supervisor handoff `8552d9cd`). That is not final
Goal acceptance or merge authorization.

S9 is in progress under attempt 2. Its items:
- the OPERATIONS.md full-profile cell (below);
- this checklist;
- reconciliation with the actual staging target;
- the affected-check batch;
- push and the non-draft MR to `staging` with source deletion;
- the MR `fast` evidence.

Still pending: owner-authorized integration, the qualifying post-integration staging disposition, and final Codex
Goal acceptance. G008 is not released and G021 stays closed.

**Attempt accounting.**
- Attempt 1 (initial execution) ended with the failed merged-harness `s8-final-3` (F11).
- **Attempt 2 of B7** covers C1 and C2 (Codex review `1bed0ddb…`) and these S9 handback items (Codex review
  `f62c1927…`, "without another implementation correction cycle").
- The checkpoint stays after attempt 4, and the history is preserved.

**OPERATIONS.md amendment (authorized by `f62c1927…` §S9 item 1).** The full profile's "Required IDs" cell (line 93)
changes from 84 to 88; nothing else changes.
- It now agrees with the full coverage row (line 120: 33/60/89, "the 88 full IDs plus B-UNIT-CHECKS").
- `scratch/profile-counts.py` over the merged `required-checks.tsv` and `infra/ci/lib.sh` derives full 88.
- No control logic or scope change; `docs/foundation/OPERATIONS.md` is already in the 15-expression scope.
- Affected checks: B-LITERALS (`lint-base`) re-scans the doc; B-SCOPE and B-REQUIRED-CHECKS follow. `off` copies
  OPERATIONS.md byte-for-byte as an APPLICABLE item, and the change is a table cell with no executed block, so `off`
  is reused with that provenance.

**Composite final local evidence (accepted by `f62c1927…`).** `a2-c1c2-1` (34/34 PASS: lint, lint-base, controls,
unit, migrations and final, on the tested implementation `12d2d8e`) replaces the failed obligations. The unaffected
passing `s8-final-3` checks (lint-web, lint-core, provenance, contracts, root, ops, rename, off) are reused with
content and input provenance. Together they give **89 distinct PASS obligations** (88 full IDs plus B-UNIT-CHECKS) and
no failing obligation. The historical `s8-final-3` run remains recorded as **failed** (87/89); this is composite
evidence, not a claim that the full run passed.

| AC | Status | Evidence |
|---|---|---|
| AC1 trusted, bounded context | Met | B-POLICY-IT IT-CTX-1..10 (cross-request with equal or absent IDs, cross-subject, same-request, outside or after the request, nested; zero BeginTx on binding rejections), U-CTX-1..9 and U-ACCESS-BOUNDARY (B-UNIT-CHECKS), IT-REVOKE in an unchanged scope (`a2-c1c2-1`) |
| AC2 membership required, no minting | Met | IT-CTX-3..6 with row-set hashes; U-POL-GATE |
| AC3 closed matrix | Met | U-POL-MATRIX (the independent oracle literal) and IT-MATRIX-1 (242 cells on database-loaded facts); U-POL-GATE/SCOPE; store-error and cancellation denials. F-1 (`s7-f1`) shows both gates fail on a flipped cell or expectation |
| AC4 profile authority derived | Met | IT-PROFILE-1..5; U-POL-SCOPE; no profile-role input type |
| AC5 policy flag | Met | IT-FLAG-1..5; CT-POL-2/3 (B-POLICY-CATALOG); MG-POL-3/6 (B-POLICY-MIGRATION) |
| AC6 enforced operations, honest failures | Met | IT-FACADE-1..6 (definitive abort versus lost reply); the operation-capability mapping test; U-ACCESS-BOUNDARY and B-MCP-SOURCE (no bypass wiring); C2: U-COMMIT, 21 cases (ERROR or ROLLBACK aborted; FATAL 57P01, PANIC 57P02, unknown severity or lost reply commit_outcome_unknown; public unavailable, one attempt, no driver text) |
| AC7 revocation | Met | IT-REVOKE-1..8 |
| AC8 ordering | Met, with recorded limits | B-POLICY-ORDERING O-1..O-15 (17/17) and the E-PG-2 locked-row inventory. **Limits:** (a) a slug-changing UPDATE upgrades the Group row to `FOR UPDATE` (a unique-key column) and the profile UPDATE takes its own exclusive row lock (F8, accepted). (b) The measured footprint changes a slug to its unchanged value, so it did not measure that upgrade; the statement rests on PostgreSQL's documented semantics. (c) O-15's lifecycle deadlock aborts one side safely (`unavailable`, no unauthorized write). (d) E-PG-3: compatible lockers may pass a queued writer, with no fairness or maximum-wait guarantee |
| AC9 optional MFA | Met | IT-MFA-1; U-POL-MFA; inherited B-IDENTITY-* verified-MFA and cancellation controls (`s8-final-3` ops, reused) |
| AC10 persistence and privileges | Met | CT-POL-1..5, MG-POL-1..6, CT-POL-4 (no copied grants, no RLS or new object); B-MIGRATE-*, B-PERSIST, B-DBRENAME, B-TOGGLE-OFF; 00001–00004 digests (B-MIGRATE-SET) |
| AC11 later isolation | Met | B-TENANCY-HTTP, B-MCP-SOURCE, U-ACCESS-BOUNDARY, CT-POL-4; the AUTHORIZATION.md G008 handoff inventory; RLS isolation not claimed |
| AC12 privacy and inherited validation | Met | IT-LOG-1 with logcheck validate, scan and planted canary; B-SECRET-CANARY; inherited suites in the composite evidence |
| AC13 local and hosted gates | **Local part met; hosted PENDING** | The wrapper on the real root with pre- and postchecks; the composite final local evidence above; F-1 (`s7-f1`); B-CI-CONTROLS 439/439 (`a2-c1c2-1`). **Pending:** MR `fast`, then the post-integration staging disposition under the owner status rule |
| AC14 binding, records, acceptance | **Records met; final acceptance PENDING** | Bound inputs and amendments (cb784148, 126605f5, aaafe0ea, b85a6792, d4a53b90), wrapper pre/post, NUL-safe status and allow-list checks, owned cleanup (every run leftover 0), failed, interrupted and deferred runs recorded, rebase and reconcile records, attempt accounting, Codex readiness ACCEPT `f62c1927…`. **Pending:** MR evidence, owner-authorized integration, staging disposition, final Codex Goal ACCEPT |
| PG-001 (not an AC) | Recorded | E-PG-1..3 on every policy run. F4 (6b/7), F7 (6a) and F9 (statements 2/3/8/9) are accepted small-fixture observations under the owner's unknown-workload disposition; there is no production performance claim. Revisit before public adapters |

## S9: push, MR !18 and the MR fast evidence (appended)

- **Push** (supervisor direction under Codex readiness ACCEPT `f62c1927…`). Commit `3b45e8754fab022f6f303911c3ef7469806fa70c`,
  on staging `65954bc`, was pushed to `refs/heads/g007-trusted-context-policy` with a lease against `c151407`.
  Pre-push inspection against `origin/staging`: 42 files, all in the 15-expression scope; no secrets or machine-local
  paths in the added lines.
- **MR !18:** https://gitlab.example/<namespace>/<project>/-/merge_requests/18 (IID 18), titled "G007: trusted
  context and organization policy". It is open and non-draft, targets `staging`, and deletes the source branch on
  merge (`force_remove_source_branch: true`). No earlier MR existed for the branch. The factual description (scope,
  reviews `f62c1927…` and `1bed0ddb…`, local evidence, pending gates, and the merge-order note about G051's !17) is
  saved at `scratch/mr-description.md` in the G007 evidence directory.
- **MR pipeline 2936197429** (merge-request event, `refs/merge-requests/18/head` at `3b45e87`): success,
  17:07:06–17:13:40Z.
  - **fast job 17095287168: SUCCESS**, 17:07:06.8–17:13:38.3Z (about 391 s). Profile: pipeline full, executed fast
    (reason foundation, base `65954bc`; lint, lint-base, lint-web, lint-core, provenance, contracts, unit). Harness exit
    0; **33/33 checks PASS** including B-REQUIRED-CHECKS; JUnit 33 cases, 0 failing. Every phase reported status 0:
    location, preflight, state, harness, postcheck, cleanup, raw_scan, complete, stage, publish, summary. Artifacts:
    `artifacts.zip`, `metadata.gz`, `junit.xml.gz`. The log is kept read-only at `scratch/fast-17095287168.log`.
  - The gate job 17095287169 is manual and was not run (no retry, skip, cancel or manual gate).
- **Refresh.** Staging moved past the MR source with a docs-only commit, `7c0eb7a`
  (`docs/agents/GITLAB-BRANCH-WORKFLOW.md`). The supervisor rebased `3b45e87` onto it, with G007 content identical,
  and rebound to `g007-binding-reconciled-7c0eb7a-ad4a53b90` (`13d28530…`, 92 inputs, scope `15b0fe19…`). This
  commit is pushed with that rebase. Docs and records only, so no local batch is needed.
- **Pending:** the refreshed MR fast run, owner-authorized merge, a qualifying post-integration staging disposition,
  and final Codex Goal acceptance. G008 is not released and G021 stays closed. Attempt 2 of B7.

## S9: refreshed MR fast run and the refresh onto 1300044 (appended)

- **MR pipeline 2936210385** (merge-request event at `36e66d9`, after the refresh onto `7c0eb7a`): success,
  17:16:10–17:22:56Z. The fast job 17095364272 was a **SUCCESS**, 17:16:10.6–17:22:54.4Z (about 404 s).
  - Profile: pipeline full, executed fast (reason foundation, base `7c0eb7a`).
  - Harness exit 0; **33/33 checks PASS**, including B-REQUIRED-CHECKS; JUnit 33 cases, 0 failing.
  - Every phase reported status 0: location, preflight, state, harness, postcheck, cleanup, raw_scan, complete, stage,
    summary, junit, candidate_scan, publish.
  - Artifacts: `artifacts.zip`, `metadata.gz`, `junit.xml.gz`; the log is kept read-only at
    `scratch/fast-17095364272.log`. The gate job 17095364273 is manual and was not run.
- **Refresh onto staging `1300044`** (owner-requested; Codex handoff `c8ab8a90`; mechanical only). The supervisor
  rebased `36e66d9` onto `1300044`, giving `bb30e0f` (tree `79ac52b2`).
  - Staging's advance is ADR-005, tenant-bound MCP connections: `contracts/mcp/connection-context.md`,
    `docs/ARCHITECTURE.md`, `docs/DECISIONS.md`, `docs/adr/ADR-005-tenant-bound-mcp-connections.md`,
    `docs/foundation/FOLLOW-UP.md`, ROADMAP and PROGRESS. It is documentation for future G014–G016 direction, with no
    G007 implementation impact.
  - `docs/foundation/FOLLOW-UP.md` (a G007 scope file) merged cleanly. Staging's change is in STAGE-MCP. G007's 14
    change lines (STAGE-TENANCY and STAGE-POLICY) are byte-identical before and after the refresh (`-U0` diffs against
    each base compared).
  - The same 42 paths differ from staging. Between `36e66d9` and the refresh, no other G007 file changed.
  - Rebound to `g007-binding-reconciled-1300044-ad4a53b90` (`6c389a04…`, 92 inputs, 18 changed cumulatively, scope
    `15b0fe19…`).
- **No G007 implementation byte changed,** so the local evidence (the composite 89 PASS obligations, `a2-c1c2-1`,
  `s9-ops-2`, `s7-f1`) and the MR fast evidence above are retained with this provenance. No local batch and no full
  qualification were run.
- **Pending:** the refreshed MR fast run, owner-authorized merge, a qualifying post-integration staging disposition,
  and final Codex Goal acceptance. G008 is not released and G021 stays closed. Attempt 2 of B7.

## S9: failed MR fast run on 1300044, staging fix and refresh onto c0852dc (appended)

- **MR pipeline 2936222549** (merge-request event at `72bd325`, on staging `1300044`): **FAILED**, 17:27:10–17:33:48Z.
  - The fast job 17095436332 **FAILED**, 17:27:10.2–17:33:46.8Z (about 397 s). Profile: pipeline full, executed fast
    (reason foundation, base `1300044`). Harness exit 1. **31 checks PASS, 2 FAIL:**
    - **B-LITERALS** FAIL;
    - **B-REQUIRED-CHECKS** FAIL ("required check B-REQUIRED-CHECKS is not PASS").
  - JUnit: 35 cases, 4 failing (the two checks and the harness and complete phase cases). Phases `harness` 1 and
    `complete` 14; every other phase 0.
  - Artifacts: `artifacts.zip`, `metadata.gz`, `junit.xml.gz`; the log is kept read-only at
    `scratch/fast-17095436332.log`. The pipeline skipped the gate job 17095436333 after the fast failure. Nothing was
    retried, skipped, canceled or run manually.
  - **Cause, solely staging `1300044`:** ADR-005 added `contracts/mcp/connection-context.md`, whose line 7 used an
    example host containing the manifest's identity literal. B-LITERALS forbids that literal under `contracts/`. No G007
    file contains it. The previous MR run on `7c0eb7a`, before ADR-005, passed 33/33. This run stays recorded as
    **failed**.
- **Staging fix `c0852dc`** (Codex handoff `dab93549`: a supervisor documentation defect, no new implementation attempt
  and no local suite). The contract example now reads `https://acme.platform.example/mcp`; B-LITERALS and the CI guards
  are unchanged. A scan of the B-LITERALS roots in this checkout finds no identity literal.
- **Refresh:** the supervisor rebased `72bd325` onto `c0852dc`, giving `84f729b` (tree `cd97d339`); the delta is only
  that contract file. Rebound to `g007-binding-reconciled-c0852dc-ad4a53b90` (`f4ef1d11…`, 92 inputs, scope
  `15b0fe19…`). No G007 byte changed, so local and MR evidence are retained with that provenance.
- **Pending:** the refreshed MR fast run, owner-authorized merge, a qualifying post-integration staging disposition,
  and final Codex Goal acceptance. G008 is not released and G021 stays closed. Attempt 2 of B7, no increment.

## S9: MR fast run on c0852dc and the refresh onto the G051 merge 20b2841 (appended)

- **MR pipeline 2936272784** (merge-request event at `4a56422`, on staging `c0852dc`): success, 17:53:44–18:00:14Z.
  - The fast job 17095721527 was a **SUCCESS**, 17:53:44.5–18:00:12.6Z (about 388 s). Profile: pipeline full,
    executed fast (reason foundation, base `c0852dc`).
  - Harness exit 0; **33/33 checks PASS**, including B-LITERALS and B-REQUIRED-CHECKS; JUnit 33 cases, 0 failing.
    Every phase reported status 0.
  - Artifacts: `artifacts.zip`, `metadata.gz`, `job.log`, `junit.xml.gz`; the log is kept read-only at
    `scratch/fast-17095721527.log`. The gate job 17095721528 is manual and was not run.
- **Refresh onto staging `20b2841`, the merge of G051's MR !17** (Codex handoff `71c7ee67`: Codex holds merge
  ownership and merges !18 under owner authority; this worker does not merge).
  - The supervisor rebased `4a56422` onto `20b2841`, giving `2173b64` (tree `e3483899`), with no conflict (a clean
    trial merge tree).
  - Merge reconciliation, verified here:
    - `infra/validation/lib.sh` equals staging (G051's C3 parser).
    - `infra/ci/validate-controls.sh` carries G051's VC9 together with G007's four count reconciliations. G007's
      diff against staging there is exactly 8 changed lines: VA7 `required_count 60` and its detail, VU1
      `33 60 89` and its detail, and J8 `60` with its message.
    - `docs/foundation/OPERATIONS.md` carries G051's build-ledger sub-list and G007's rows 93 (full, 88 IDs) and 120
      (33/60/89).
  - The merged `required-checks.tsv` still has 95 unique rows (fast 33, integration 71, full 88, staging 5;
    `profile-counts.py`), so the expected 60/89 split stays exact.
  - Rebound to `g007-binding-reconciled-20b2841-ad4a53b90` (`993dc8cd…`, 92 inputs, scope `15b0fe19…`).
- **Evidence:** there was no textual conflict, so Codex requires no local covering run. The combined controls (VC9
  with G007's counts) are exercised by the post-merge staging gate. No G007 implementation byte changed.
- **Pending:** the refreshed MR fast run, the Codex merge under owner authority, a qualifying post-integration staging
  disposition, and final Codex Goal acceptance. G008 is not released and G021 stays closed. Attempt 2 of B7.

## Integration and staging pipeline 2936313877 (appended)

- MR !18 (source `b126252`) was merged by Codex under owner authority as staging `e5b1823` (Codex handoff `3eec9608`);
  the remote source branch was deleted.
- **Staging pipeline 2936313877** (push, `e5b1823`), **canceled**, 18:25:08–18:33:07Z.
  - fast job 17095957029: **success**, 18:25:08.8–18:31:48.3Z (about 400 s). Profile: pipeline full, executed fast
    (reason foundation, base `20b2841`). Harness exit 0; 33/33 checks PASS, including B-REQUIRED-CHECKS; JUnit 33
    cases, 0 failing.
  - gate job 17095957030: **canceled**, 18:31:49.8–18:33:05.6Z (about 76 s). It started the full remainder (controls,
    root, migrations, ops, rename, off; deadline 8666 s) at 18:32:23Z, received TERM at 18:32:55Z ("script canceled
    externally (UI, API)"), and its process group was gone at 18:32:56Z. The canceling actor is not established. No
    check result exists for the remainder, so B-CI-CONTROLS (G051's VC9 with G007's VA7/VU1/J8 counts) was **not
    exercised** in this pipeline.
  - Logs are kept read-only at `scratch/staging-fast-17095957029.log` and `scratch/staging-gate-17095957030.log` in the
    G007 evidence directory. This worker did nothing to the pipeline (read-only polling, stopped on supervisor
    direction).
  - Under the owner status rule, canceled is recorded as canceled, never as a test pass. Its disposition is Codex's.

## Attempt 3 — post-integration Duo disposition (appended)

Authority: [G007-implementation-codex-review.md](G007-implementation-codex-review.md), section "Post-integration Duo
disposition — 2026-10-11 — CORRECTION LOOP" (file SHA-256
`9536ede8effea1a7d6f9ff6c3e5b3be4b7c5df9e39a049bb45ff04d54eaad38a`; Codex handoff `325f8431`).
**Attempt 3 of B7**: a bounded documentation correction; the checkpoint stays after attempt 4.

- **Finding:** Duo !18 note 3993788301 (repeated in 3993992512 and 3994034607). Codex found the availability concern
  valid. Policy writes request the Organization's `FOR NO KEY UPDATE` lock before the capability decision, and
  selected or Organization-context Group writes can lock the target Group before denial. So a live subject lacking the
  capability can contend with authorized writes. No unauthorized mutation is possible, and ordinary guarded reads are
  lock-free and not blocked. Duo's earlier truncated no-finding reviews are retained as such, not as comprehensive
  proof.
- **Disposition:** documentation, as Duo offered, under PG-001's existing owner risk disposition and with no public
  tenant adapter. The reviewed locked-facts decision, mutation and lock order are retained. There is no change to
  authorization, SQL, schema, indexes or lock order, no speculative pre-decision, and no numeric guarantee.
- **Change, append only, on a new branch `g007-a3-duo-docs` from staging `c845bdb`:**
  - `docs/foundation/AUTHORIZATION.md` gains "Availability: denied exclusive-write requests". It names the affected
    operations (`SetOrganizationPolicy` at step 3; `RenameGroup` and `ChangeGroupSlug` at step 6 in Organization-only
    or selected-Group context) and what they delay. It states that reads are not blocked and nothing unauthorized can
    change. It requires future adapters to provide per-subject admission or rate limits and bounded request and
    transaction lifetimes before exposing these operations, and states that neither exists yet.
  - `docs/PERFORMANCE.md` gains "PG-001 post-integration Duo disposition — 2026-10-11" with the same content.
- **Owner exception:** direct staging publication with **no local test runs**, recorded by Codex. Local runs are
  **WAIVED** by owner direction: no Docker, build, lint, controls, harness or `g007-vw.sh`. They are not PASS. The
  supervisor is the single publisher to staging; this worker only commits. Hosted evidence will be reviewed later.
- **Status:** Codex final Goal ACCEPT is still pending the applicable staging disposition and review of the completed
  handback. G008 is not released and G021 stays closed.

## Final status — ACCEPTED, completed and integrated (appended)

- **Status: ACCEPTED by Codex. G007 is completed and integrated on staging.** Authority:
  [G007-implementation-codex-review.md](G007-implementation-codex-review.md), section "Final implementation acceptance
  — 2026-10-11 — ACCEPT", on staging `12c6ca6`. This acceptance closes the earlier conditional integration and the
  AC13/AC14 hosted and final gates; the "IN PROGRESS" header and the S9 status above are superseded and kept as
  history. G007 closes at **attempt 3 of B7**.
- **Hosted reference:** staging pipeline **2936644423 SUCCESS** on `4bab0e796ad6a0bb0a599431cdd35d70e83bc165`.
  - fast job 17097945983 and gate job 17097945984 both succeeded on that commit.
  - All 93 harness obligations PASS, with no FAIL, UNMET or duplicate outcome IDs; 441/441 controls PASS.
  - B-POLICY-CATALOG, B-POLICY-MIGRATION, B-POLICY-IT and B-POLICY-ORDERING all PASS.
  - Both build ledgers MET, 9/9 and 69/69, with 0 unknown and no unexcepted duplicates.
  - Final job traces show every wrapper phase 0, full publication, the registry removed, `owned_remaining` 0 and
    scrub 0.
  - No core or contracts change occurred between the G007 merge and this qualification.
- **Integration chain:**
  - MR !18 was merged under owner authority as staging `e5b1823`. Its MR fast evidence is pipeline 2936306602,
    33/33.
  - The attempt-3 documentation correction was published at `47c503a`. AUTHORIZATION.md is
    `0363d9d177e06395e3f1db528081791cbeeb6a928d2b48c27eb5627bda884f22` and PERFORMANCE.md is
    `e5798a83da933b280debaffcf880c70343d650aca0018ecb900f5eadaf756db3`, the exact accepted attempt-3 hashes.
- **History is kept as recorded.** These stay recorded as they happened and are not rewritten:
  - staging pipeline 2936313877, canceled (gate canceled externally);
  - the failed merged-harness `s8-final-3`;
  - the failed `s1-lcm-1`, `s23-unit-1`, `s46-clpmo-1` and `s9-ops-1` runs, and MR pipeline 2936222549;
  - the deferred `s8-final`;
  - every correction run.

  The attempt-3 local runs were waived by owner direction and are not PASS.
- **Dependencies:** G008's dependency on G007 is satisfied. That is not authorization for G008 implementation. G021
  stays closed.
- **Limits carried forward:**
  - PG-001's recorded lock and fairness limitations stand, as do the slug-key lock upgrade, the small-fixture scan
    observations and denied exclusive-write contention.
  - Future public adapters must provide per-subject admission or rate limits and bounded request and transaction
    lifetimes (deadlines) before exposing the guarded write operations.
  - None of those mitigations exists yet, and no numeric performance guarantee is made.
