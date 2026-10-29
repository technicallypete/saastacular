> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G065 — Goal-neutral CI-control labels — Result

Status: **FINAL ACCEPT — Codex, 2026-10-23** ([final review](G065-final-acceptance-codex-review-20261023.md)). The rename (R1 + R8) and the RC8 harness
repair (K1, K2, K3 and VB7, B3 attempt 2) have Codex SOURCE ACCEPT (review (not included)) and passed the
owner-selected hosted full profile on `db56209e` (`validate:checks` and the manually started `validate:integration`, both
SUCCESS); MR !87 merged as `4afca837` with the merged `infra` and `docs/foundation` trees equal to the tested trees. The dated
receipt is [Hosted qualification, merge and cleanup](#hosted-qualification-merge-and-cleanup-2026-10-23). The sections between
this paragraph and that receipt are the historical record as written at each handback: wording such as "pending", "draft" or
"not qualified" in them described the state at that time and is superseded by the receipt, with the failed and intermediate
history preserved. The local rename qualification applies to the original rename tree `f77e7823` only, and the local 8-selector run
on the repair was diagnosis only. B3: attempt 1 spent (rename and mapping correction), the repair was attempt 2 (see
[Budget](#budget)); no attempt 3.
The `ai.beginify.*` targets recorded in the sections below belong to the failed intermediate candidate `566fb7cd`; the
current mapping is `sdlc.ci.*` — see [Mapping correction](#mapping-correction-spec-r4-sdlcci).
Goal: [G065](../goals/G065-neutral-ci-control-labels.md). Spec: [r6](../specs/G065-neutral-ci-control-labels-spec.md) (current;
sha256 `2b84be8c7934a140fcaf9940a05893db4b8faa27d401a7087dcb95c177104c1d`; the r3, r4 and r5 texts are retained as snapshots or
history). Plan: [r3](../plans/G065-neutral-ci-control-labels-plan.md) (the r2 body plus the accepted §11 RC8 repair addendum,
Codex ACCEPT of `2af29d26`; sha256 `e2c2a6595d9e5f78b23af8e21aace16df12174fac80658124b216c0aa55762c4`).

## Source

- Base: published staging `675c07fca1aaf19bf39c8bf81efe6f251a88bd8c`. The three source files are byte-identical to the
  Plan base `bb2b4e37` (`git diff --stat bb2b4e37 675c07fc -- <3 files>` empty), so every §2 line number holds.
- Branch `g065-impl`; rename commit SHA: see the branch head / handback (this file is in the same commit).
- Writer: isolated stack-free implementation worker (Opus 5.5 high); fake `docker` first in PATH and
  `DOCKER_HOST=tcp://127.0.0.1:1` throughout. No Docker, build, pull or daemon contact.

## Mapping verification (R1, AC1 source side)

Mapping applied exactly: `g003.controls.run` → `ai.beginify.run`, `g003.control` → `ai.beginify.ci.control`,
`g003.outer` → `ai.beginify.ci.outer`, `g003.controls.owner` → `ai.beginify.ci.owner`. Applied by a boundary-aware
substitution (longer `controls.run`/`controls.owner` matched before `control`; `control` not followed by `s`/word char;
the dot spelling preserved per occurrence: `.`, `\.`, `\\.`), then re-verified line by line.

| File | Changed lines | Plan §2 inventory | Lines |
|---|---|---|---|
| `infra/ci/validate-controls.sh` | 47 | 42 exact-key + 5 escaped forms = 47 | 35 230 237 261 336 371 372 514 694 704 705 710 766 886 887 894 897 1382 2484 2490 2496 2499 2501 2696 2708 2862 2869 2917 2932 2933 2938 2972 3012 3028 3039 3046 3047 3102 3103 3626 3875 3879 4267 4269 4319 7228 7231 |
| `infra/validation/lib.sh` | 7 | 5 exact-key + 2 prose = 7 | 517 522 1422 1431 1512 1519 1550 |
| `docs/foundation/OPERATIONS.md` | 4 | 4 | 722 723 757 759 |

Totals: 58 removed / 58 added lines; no other line changed; file modes unchanged.

Escaped/glob forms (Plan §6 risk 4), old → new:
- 897 `^g003\.control=` → `^ai\.beginify\.ci\.control=`
- 2972 `--label g003\.outer=` → `--label ai\.beginify\.ci\.outer=`
- 3012, 3028 `g003\\.control=` (double-quoted ERE) → `ai\\.beginify\\.ci\\.control=`
- 3102 `--label g003\.controls\.owner=` → `--label ai\.beginify\.ci\.owner=` (3103 likewise: `ai\.beginify\.ci\.owner=[^ ]*` and `-xF "ai.beginify.ci.owner=$5"`)
- 2869 glob `*g003.controls.owner*` → `*ai.beginify.ci.owner*`; its `sed 's/^g003\.controls\.owner=//p'` → `sed 's/^ai\.beginify\.ci\.owner=//p'`
- 2917 `FAKE_PS_FAIL=g003.outer=` → `FAKE_PS_FAIL=ai.beginify.ci.outer=` (trailing-`=` prefix shape kept; consumer at 2847 is a prefix match)

Prose rewording beyond the key token (the four listed prose forms only): lib 517 "a g003.controls.run label" → "an
ai.beginify.run label" (article); lib 1512 "g003 control key" → "CI-control key"; lib 1550 "its `g003.*` labels" → "its
`ai.beginify.*` labels"; vc 3039 "g003 control labels" → "CI-control labels" (line in the §2.1 prose list).

Token check (stack-free script over `git diff -U0`): every removed line, with the four keys mapped and the four prose
rewordings above applied, equals its added line byte for byte — `58 line pairs; key-only: 54; with listed prose
rewording: 4; mismatches: 0`. Hence all unrelated `g003` strings (helpers `g003_*`, `G003_*` names incl. the variable name
`G003_KEYS`, state paths `g003-pending`/`g003-controls-*`/`g003-local`/`g003-validate-*`, `g003ctl-*`, fixture keys
`g003.v27`/`g003.vc9`) and the `g001.*`/`g064.bg` namespaces are byte-identical to the base.

Grep gate (empty output, exit 1, both forms):
```
$ grep -nE 'g003(\.|\\+\.)(controls(\.|\\+\.)run|control|outer|controls(\.|\\+\.)owner)' infra/ci/validate-controls.sh infra/validation/lib.sh docs/foundation/OPERATIONS.md
(no output; exit 1)
$ grep -nE 'g003\\\\\.|g003\\\.' <same three files>
(no output; exit 1)
```

Not changed, deliberately (outside the §2 inventory; prose only, no key token): vc 243 "a surviving g003 object" (a
listed out-of-scope line, `g003_*` helper context) and vc 3040 "no g003 label at all" (seed comment). Both are
reportable for a later wording pass; changing them would exceed the inventory.

Stack-free checks: `bash -n infra/ci/validate-controls.sh` rc 0; `bash -n infra/validation/lib.sh` rc 0;
`bash infra/ci/validate-controls.sh --list` → 528 entries (no Docker on that path: it exits at line 189 before any
daemon use); `source infra/validation/lib.sh` under the fake docker → rc 0, `G003_KEYS=ai.beginify.ci.owner
ai.beginify.run ai.beginify.ci.control ai.beginify.ci.outer`, `_res_live`/`g003_owner_guard` defined.

## Bounded `_res_live` proof script (Plan §4; prepared, NOT run against a daemon)

`<home>/.local/state/platform-evidence/g065-a1/proof/f0-resliveproof.sh` (sha256
`6e7cb84bbabd3d4ebdaf4101b99afc691db4f0c4d7d993b92c841dc84673626a`), outside the repository. Usage:
`f0-resliveproof.sh REPO_ROOT [LOG_FILE]`; sources the repo's unchanged `infra/validation/lib.sh` and calls only the
existing `_res_live container`; steps 1–5 as in the Plan (U unmirrored, M mirrored with a `standalone <id> ci-g065proof`
B row; PASS iff `RES_ALLN − n0 == 2` and neither in `RES_IDS`; then M's row removed → M in `RES_IDS`, `RES_ALLN`
unchanged). Exact IDs logged right after each create; EXIT trap removes only those two IDs (`docker rm -v`), verifies
them gone and `docker ps -a --filter label=ai.beginify.run=$RUN_ID` empty, removes WORK. Exit 0 PASS, 1 FAIL, 2 setup,
3 INCONCLUSIVE (daemon container count moved by other than the two creates), 4 cleanup unverified. Uses `--pull never`
(busybox:1.37.0 must already be present; nothing is pulled). Run it only in the owner's test slot with no controls launch
or other harness run in progress.

Logic check only (fake `docker` implementing `create`/`ps`/`rm`/`inspect`; no daemon; artifacts in
`.../g065-a1/proof/logic-check/`): candidate lib.sh → `RESULT: PASS (rc=0)`, delta 2, M foreground only after row
removal, cleanup verified; base lib.sh (old key) → `RESULT: FAIL (rc=1)`, `RES_ALLN delta 1 != 2 … old key still read?`,
cleanup verified. This is not the §4 proof.

## Cutover receipt (R8, AC2)

Recorded by the coordinator, 2026-10-23, on the shared local daemon. Evidence: `platform-evidence/g065-a1/` (`s0/s0.txt`,
`qual-566fb7cd-ai-beginify-SUPERSEDED/`, `qual2/`).

- **S0 snapshot (15:12:52Z, staging `675c07fc`).** Inventory 15 containers, 12 volumes, 6 networks, 52 images. No
  `validate-controls.sh`, `run.sh` or `validate.sh` process was live; the state root held only `g003-local` with no
  pending marker. **Stop list: empty** (no old-label creator existed, so nothing was stopped or reaped).
- **Absence verification.** The sixteen old-key listings (`g003.controls.run`, `g003.control`, `g003.outer`,
  `g003.controls.owner` x container/network/volume/image) all succeeded and were empty at S0, and again immediately before each
  of the two qualification launches. **No survivor, so no ownership proof was needed and no object was removed** (no
  `dead:*` classification, no exact-ID or `cleanup-owned.sh` action).
- **Intermediate mapping.** The first candidate `566fb7cd` (keys `ai.beginify.*`) ran its selectors on the real daemon
  (V18l/m/n 3 of 3, V27/V27b/VC1/VB1 4 of 4, the bounded `_res_live` invocation PASS) after its fast profile failed
  `B-LITERALS`. It created labelled objects under the intermediate keys; the post-run listings for the old keys and for the
  `ai.beginify.*` keys were all empty and the inventory was 15/12/6/52 again, so **no intermediate-key resource remained**
  and no reconciliation removal was required. Its evidence is kept as `qual-566fb7cd-ai-beginify-SUPERSEDED`.
- **Cutover launch (corrected candidate `f77e7823`).** The first launch on the `sdlc.ci.*` keys started clean (startup guard
  clear, no old-label resource, no live old-source harness run on the daemon).
- **Post-run.** The sixteen `sdlc.ci.*` listings, the sixteen `ai.beginify.*` listings and the sixteen old-key listings were
  all empty; inventory 15/12/6/52 equal to the pre-step counts.
- **Cutover SHA.** Candidate `f77e78231f92677fac30bec051ff8fd292e56d62` (tree `2e15d5cdacff97eabc773d2b3e206c3df251f0c7`,
  parent staging `dce85b89`); the merged SHA of the later MR !87 is `4afca837` (historical wording "recorded after the merge"; see the receipt below).
- **Retained.** No G065 resource. The held `g064-*`/`worktree-agent-*` refs, failed-run evidence and the superseded
  `qual-566fb7cd-ai-beginify-SUPERSEDED` evidence are retained as already recorded (Codex retention disposition; no deletion
  was authorized).

## Validation (AC8)

Scope of this section: the **original rename candidate only** (tested tree `2e15d5cd`, SHA `f77e7823`). It is not
qualification of the K1–K3/VB7 repair, which changed the same three files after this run; that qualification is the hosted
full profile on the final SHA (see "RC8 harness repair K1–K3 + VB7" below), and the repair's local diagnosis is diagnosis only.

Tested once on the corrected candidate before merge (owner timing exception), frozen checkout `g065-run2`, harness lock held,
2026-10-23: SHA `f77e78231f92677fac30bec051ff8fd292e56d62`, tree `2e15d5cdacff97eabc773d2b3e206c3df251f0c7`, staging
parent `dce85b89` (source files unchanged since `675c07fc`).

| Check | Outcome |
| --- | --- |
| Fast profile (`validate.sh --profile fast`: lint, lint-base, lint-web, lint-core, provenance, contracts, unit; includes B-SHELL and B-LITERALS) | PASS, exit 0, final status 0, 15:37:42Z to 15:46:52Z |
| `validate-controls.sh --only V18l,V18m,V18n` | 3 passed, 0 failed (15:46:52Z to 15:50:46Z) |
| `validate-controls.sh --only V27,V27b,VC1,VB1` | 4 passed, 0 failed (15:50:46Z to 15:51:43Z) |
| Bounded `_res_live` invocation (existing function; run id `9de351c27257`) | PASS: baseline `RES_ALLN` 0; U `fe224a5edff6...3481` (label `sdlc.ci.run`, no `g001.run`, no registry row) counted, delta 2; M `60bac657dfdc...b3e5` (`g001.run` plus `sdlc.ci.run`, registry row `ci-g065proof`) excluded from the foreground set (`RES_FGN` 0); after M's row was removed M appeared in the foreground set (`RES_FGN` 1) with `RES_ALLN` unchanged. Both objects removed by exact ID and verified gone; run-label listing empty. |
| Inventory | 15 containers, 12 volumes, 6 networks, 52 images before and after every step; the 48 old, intermediate and new key listings empty after the run |

The candidate that preceded it, `566fb7cd` (`ai.beginify.*`), is recorded as a failed fast profile (`B-LITERALS`, and
`B-REQUIRED-CHECKS` as a consequence) and is not counted toward this qualification. Hosted fast on the MR source SHA, the
Duo disposition, Codex source review and the merged-tree verification were recorded after they occurred (see the receipt below). `B-CI-CONTROLS`
(the 528-control suite) and the full profile were not executed for this Goal and are not reported as passing.

## Deviations

- Target keys changed from the Plan r2 mapping to the Spec r4 `sdlc.ci.*` mapping after `B-LITERALS` rejected the product
  literal (see "Mapping correction"); the rest of the Plan is unchanged. Two prose-only `g003` mentions outside the inventory left unchanged (vc 243, 3040; see above).
- Proof script adds `--pull never` and an INCONCLUSIVE guard on concurrent container-count drift; both only tighten the
  Plan §4 procedure.

## Budget

Current binding: B3 is three shared attempts. Attempt 1 is spent (the rename and its mapping correction); the RC8 harness
repair K1–K3 + VB7 is attempt 2, with its checkpoint at the repair handback; no reset or extension. The ledger remains the
supervisor's (Goal and Spec r6).

History (preserved): B3 had 0 attempts spent at release; the first implementation round was attempt 1 of 3. The mapping
correction was applied by the same owner within that attempt (no restart, no budget reset).

## Self-review (writer, stack-free)

Line-by-line diff reviewed against Plan §2: 47/7/4 changed lines equal the inventory; every change is the exact-key token
(or one of the four listed prose rewordings); escaped forms map dot-for-dot; the `FAKE_PS_FAIL` prefix shape and the
owner-key value format are unchanged; cross-file couplings §2.5 (RUN_LABEL ↔ `_res_live`; OWNER_LABEL ↔ guard
`inspect`/`G003_KEYS` ↔ fake `inspect`/`seed`; nested `--extra-label` ↔ readers ↔ `case_labels` ↔ assertions/`OL`) all
use the same new keys. No runtime validation was executed by the writer; that is the coordinator's S3.

## Mapping correction (Spec r4, sdlc.ci.*)

- **Failed intermediate candidate** `566fb7cd1dd13bd425d1fa5454e914a861d07ef1` (mapping `ai.beginify.*`, preserved on
  `g065-impl`, not force-pushed over): the **local** fast profile (`validate.sh --profile fast`, run by the coordinator under the harness lock; not a hosted pipeline) failed `B-LITERALS` ("identity literal found in
  infra/validation/lib.sh infra/ci/validate-controls.sh docs/foundation/OPERATIONS.md": the target namespace contained
  the manifest identity literal `beginify`) and `B-REQUIRED-CHECKS` as a consequence. Executed-and-failed evidence.
- **Correction input:** Codex handoff f985ba65; Spec r4 on staging `f3181d04` (spec file sha256
  `c8a838a033b2195f0e8f13fa2c4e9c469a541d4d5ea6212ab25631487d12c5c3`), exact mapping:
  `g003.controls.run` → `sdlc.ci.run`, `g003.control` → `sdlc.ci.control`, `g003.outer` → `sdlc.ci.outer`,
  `g003.controls.owner` → `sdlc.ci.owner`. Plan r2 is otherwise unchanged (addendum in the Plan file).
- **Correction commit:** the commit carrying this record, on top of `566fb7cd` (branch head of `g065-impl`). Staging moved
  `675c07fc` → `f3181d04` with one management-only commit (Goal/Spec/review/PROGRESS/handoff; none of the three source
  files, this Result or the Plan); `g065-impl` was **not** rebased, because a rebase would rewrite `566fb7cd` and need a
  force push, which the correction instruction forbids. The branch-workflow reconciliation onto `f3181d04` stays with the
  coordinator before the validation batch (exact-delta: management-only, conflict-free expected).
- **Mechanical change:** full-key patterns only, dot spelling preserved per occurrence (`.`, `\.`, `\\.`):
  `ai.beginify.ci.{owner,control,outer}` → `sdlc.ci.{owner,control,outer}`, `ai.beginify.run` → `sdlc.ci.run`, lib 1550
  prose `ai.beginify.*` → `sdlc.ci.*`. Escaped forms now: 897 `^sdlc\.ci\.control=`; 2869 `*sdlc.ci.owner*` and
  `s/^sdlc\.ci\.owner=//p`; 2917 `FAKE_PS_FAIL=sdlc.ci.outer=`; 2972 `sdlc\.ci\.outer=`; 3012/3028
  `sdlc\\.ci\\.control=`; 3102/3103 `sdlc\.ci\.owner=`. `G003_KEYS="sdlc.ci.owner sdlc.ci.run sdlc.ci.control
  sdlc.ci.outer"`.
- **Counts vs the original base `675c07fc`** (`git diff -U0 675c07fc HEAD -- <3 files>`): validate-controls.sh 47, lib.sh 7,
  OPERATIONS.md 4 = 58 lines, the same line numbers as §Mapping verification. Token check with the sdlc.ci mapping:
  `58 line pairs; key-only: 54; with listed prose rewording: 4; mismatches: 0`.
- **Gates** (three files): old keys, all spellings → no output (exit 1); `ai(\\*\.)beginify` → no output (exit 1).
  B-LITERALS invariant, `grep -ciE beginify` before (base `675c07fc`) / after: lib.sh 0/0, validate-controls.sh 0/0,
  OPERATIONS.md 0/0 — no new identity literal.
- **Stack-free checks:** `bash -n` rc 0 on both scripts; `--list` = 528; `source lib.sh` under the fake docker rc 0 with
  `G003_KEYS=sdlc.ci.owner sdlc.ci.run sdlc.ci.control sdlc.ci.outer`.
- **Proof script** retargeted to `sdlc.ci.run`/`sdlc.ci.control` (labels, listing, text):
  `.../g065-a1/proof/f0-resliveproof.sh` sha256 `6a4f98597056f75d2fe4a1d26edb8aab7ae6a76b8592516f903ac98f611c4026`; the
  previous copy is kept as `f0-resliveproof-ai-beginify.sh` (`6e7cb84b…`). Logic check only (fake docker, no daemon):
  candidate lib.sh → PASS (delta 2, M foreground only after row removal, cleanup verified); base lib.sh → FAIL, delta 1
  ("lib.sh does not read sdlc.ci.run (old key still read?)"), cleanup verified. Not the §4 proof; not run on a daemon.
- **Budget:** this correction is within implementation attempt 1 of 3 (a mapping correction from a Spec revision, not a new
  attempt by itself; supervisor reconciles accounting).

## RC8 harness repair K1–K3 + VB7 — implementation round (attempt 2), 2026-10-23

Status of this round at its handback (historical; superseded by the hosted receipt below, which passed): **implemented, stack-free checks only; not qualified.** Qualification is hosted-only: automatic
`validate:checks` plus manually triggered `validate:integration` in the same MR pipeline on the final SHA (Spec r6, Plan §11.3).

- **Inputs:** Plan §11 (Codex ACCEPT of `2af29d26`, Plan sha256 `e2c2a6595d9e5f78b23af8e21aace16df12174fac80658124b216c0aa55762c4`;
  binding conditions in the review record (not included), "Corrected Plan r3 addendum"); Spec r6
  (sha256 `2b84be8c7934a140fcaf9940a05893db4b8faa27d401a7087dcb95c177104c1d`), K1–K3 and AC9.
- **Source:** branch `g065-repair` = `origin/g065-rename` `db968e2b` rebased onto staging `373aa661` (management-only delta,
  no conflict; `git diff --stat db968e2b <rebased head c6ecc902> -- infra docs/foundation` empty). Commits: `aea36f9e`
  (K1+K2), `eb9c0304` (K3+VB7), then the commit carrying this record (docs count + Result). Writer: isolated stack-free
  worker (Opus 5.5 high); fake `docker` first in PATH (exits 97 with a refusal message) and `DOCKER_HOST=tcp://127.0.0.1:1`.

### What changed (final line numbers on `g065-repair`)

| Item | Change | Where |
|---|---|---|
| K1 | all-image listing (`image ls -aq`) in `gate_count`, `sweep` removal + verification, `lcount`, `vc_clean`, VC4 count; `lcount` splits listing from count, prints `-1` on any failed listing and returns 0; every caller tests `< 0` first and records "the Docker listing of <label> failed; the count is unverified" (VC4: "the image listing failed; the consumer count is unverified"); `vc_clean` makes a failed listing ("cleanup unverified") and each failed `image rm -f` (exact ID + first error line) a `prob`; `sweep` sets `left=1` with `SWEEP_LEFT` text for a failed image listing/removal | `validate-controls.sh` 400, 552-554, 2687-2697, 2763, 2789-2792, 3640, 4355-4363, 4423-4425 |
| K2 | `vc_record_tag TAG WHAT`: `image inspect` failure, `gate_record` failure and a row missing afterwards each a `prob` naming the tag; called after VC8 D and D2 (`$proj-svc:latest`) and after VC9 A's `vc_run` for `one two three four` | `validate-controls.sh` 4364-4373, 4603, 4608, 4817 |
| K3 (a) | failed `platform:start` → `ok=0`, `record B-PERSIST FAIL "restart: platform:start failed (see persist-start.log)"`, `return 1`; `|| ok=0` on the migrate status read and the MinIO read | `run.sh` 5460-5463, 5465 |
| K3 (b) | `variants_ops`: `persist "$D" || { save_logs "$D" ops; compose "$D" down -v …|| true; return; }` (approved early return) | `run.sh` 5353-5355 |
| K3 (c) | `res_witness_abort` (W by exact ID via `_dro_remove`, mark `stopped`; no-op without W or once stopped; 1 when unverified); EXIT trap: `bg_stop_all; res_witness_abort || log …;` before `it_runner_stop_all … cleanup` | `lib.sh` 1043-1055; `run.sh` 110-112 |
| VB7 | table `"D VB7"` (line 110, after VB6); driver mode `witness-abort` (`OWNED_REG=$VB_REG`, trap `bg_stop_all; res_witness_abort || log; drv_cleanup`, `res_witness_start`, `exit 2`); `vb7_persist`, `vb7_rows`, `ctl_VB7` after `ctl_VB5`, before the `vb_copy` comment | `validate-controls.sh` 110, 3508-3510, 3613-3618, 3779-3854 |
| docs | control count 528 → 529 (the only pin; repo grep `-w 528` outside management finds no other) | `OPERATIONS.md` 753 |

VB7 case a/a2 call `if persist` (the production caller's error-handling context): they prove the explicit return and the
ledger row, not an errexit guarantee. Case b requires the real witness: a missing "witness … live" log line, no `w.id`,
or any `hold.d` latch is a `prob`; there is no fallback path. Static assertion: exactly one `trap 'RUN_STATUS=` line in
run.sh, naming `res_witness_abort` after `bg_stop_all` and before `cleanup' EXIT`.

Interpretations (no scope change): VC9 "after 4703" is applied after the A `vc_run` (the 4701-4703 steps execute inside
it; the helper is control-side). VC4's listing also uses `-aq` (Plan §11.2 K1 risk note measures VC4 against `-a`).

### Commands run (stack-free) and outcomes

- `bash -n` on `validate-controls.sh`, `lib.sh`, `run.sh`: rc 0 each. `shellcheck`: not installed on the host; not run.
- `lcount` extracted verbatim, `D` stubbed, under `set -euo pipefail`: `zero=0 some=5 imgfail=-1 psfail=-1`; real path
  through the fake docker `-1`; the shell did not abort.
- VB7a, `vb7_persist`/`vb7_rows` extracted verbatim with the ctl_VB7 a/a2 assertions: candidate run.sh → a rc 1, row
  `B-PERSIST FAIL restart: platform:start failed (see persist-start.log)`; a2 rc 0, row `FAIL persistence behaviour`;
  0 problems. Negative control on the pre-repair run.sh (`origin/g065-rename`): 2 problems (a returned 0; row
  `persistence behaviour`, not the restart row). No call reached the fake docker.
- Diagnostic (not VB7): `persist` called directly, errexit live, start passing, migrate and Compose failing: candidate
  shell rc 0, reached its end, one FAIL row; pre-repair shell rc 1, no row (the unguarded 5458 read aborted it).
- `res_witness_abort` branches (real lib.sh, stubbed `docker`, `RES_RM_S=0`): no W → 0; already stopped → 0; removed
  ("No such container") → 0, `stopped`; daemon refusing → 1, `stopped`, `hold.d/run` latched, `RESOURCES FAIL run:
  teardown unverified` row.
- Static trap assertion (verbatim from ctl_VB7): candidate PASS (1 trap line); pre-repair FAIL.
- Static gates on the three harness files: the four old keys `g003.(control|controls.run|controls.owner|outer)` 0/0/0;
  `ai.beginify` 0/0/0; case-insensitive product literal 0/0/0. `g003.` (any) 5/0/0, unchanged from `db968e2b` (the
  out-of-scope fixture keys `g003.v27`/`g003.vc9`, Plan §2.1). Control table entries 529 (awk over `CONTROLS=(…)`).
  `git diff --stat origin/g065-rename -- infra docs`: the four approved paths only; `lib.sh` hunks only at 1043
  (`_res_live` 517-560 untouched), `run.sh` hunks only at 110, 5352, 5457-5460.

### Not executed (not reported as passing)

Hosted `validate:checks`; hosted `validate:integration` (B-CI-CONTROLS with VB7 among 529, VC1/VC4/VC8/VC9, B-PERSIST,
RESOURCES/BACKGROUND rows, `ci-controls-timings.tsv`, `inventory-leftover`); any local real-Docker check (VB7b's real
witness, the `--only VC1,VC4,VC8,VC9,VB2,VB5,V18n` diagnosis, the K2 inventory rows); the local fast profile; any local
full suite; `validate-controls.sh --list`; `shellcheck` (B-SHELL runs it hosted). VB7b's witness path is unexercised
until the hosted run: whether `res_witness_start` reaches "live" inside the driver is unknown (Plan §11.2).

### Budget (B3)

Attempt 1 spent (rename + mapping correction). This repair is **attempt 2**; its checkpoint is due at this handback. A
hosted failure on the repaired SHA is executed-and-failed evidence inside attempt 2; the one bounded correction it may
force is attempt 3, the last. No reset or extension.

Qualification of this repair is hosted-only, on the final MR SHA: `validate:checks` and manually triggered `validate:integration` in the same pipeline.

### Local diagnosis by the coordinator (2026-10-23; diagnosis only, not qualification)

Run on the writer's head `915aaf5e` (tree `c47baf2a`) in a frozen disposable worktree under the harness lock, concurrently with the Codex source review. It supersedes the "Not executed" list above only for the items named here; the hosted run remains the sole qualification gate (Spec r6).

- **Command:** `validate-controls.sh --only VB7,VC1,VC4,VC8,VC9,VB2,VB5,V18n --keep-scratch`, 16:54:21–16:58:36Z, exit 0, `ci-controls: 8 passed, 0 failed (of 8 selected; 529 defined)`; `--list` printed 529 controls. Evidence: `~/.local/state/platform-evidence/g065-a2/diag/` (`run.txt`, `ctl-a.log`, `ctl-a.stdout.txt`, `ci-controls-timings.tsv`).
- **VB7:** a PASS (strict-mode `persist` after a failed `platform:start` returned 1 with one `B-PERSIST FAIL` restart row; after a failed migrate read returned 0 with one FAIL row). b PASS: the real witness `99b81b45b4a5` reached "live", then exit 2; BACKGROUND FAIL (stopped before join), the witness removed by exact ID and marked stopped, no latch, no labelled leftover; the EXIT trap orders `bg_stop_all`, `res_witness_abort`, `cleanup`. This answers the open question above for the local daemon only (Docker 29.8.2, Compose 5.5.1, buildx 0.37.2).
- **VC1, VC4, VC8, VC9, VB2, VB5, V18n:** PASS, with `image ls -aq` in the cleanup and counting helpers. The K2 inventory rows are not retained after a clean pass (the control inventory is removed), so K2 recording is not independently evidenced locally beyond VC8/VC9 passing; the hosted run supplies that evidence.
- **Resources:** before and after 15 containers / 12 volumes / 6 networks / 52 images; the 16 `sdlc.ci.*` listings (run, control, outer, owner × container, network, volume, image) empty and not failed before and after. The kept scratch was removed and its worktree reclaimed.
- **Not run locally:** the fast profile, any full suite, `B-CI-CONTROLS`; hosted `validate:checks` and `validate:integration` remained pending on the final SHA at that time (both passed; see the receipt below).

## Hosted qualification, merge and cleanup (2026-10-23)

Recorded by the coordinator after the owner-selected hosted full profile. Source ACCEPT, hosted-gate ACCEPT and the prose-delta ACCEPT are Codex's
(review (not included)); final Goal ACCEPT and the B3 ledger remain Codex's.

### Hosted gate (the qualification)

- MR !87, source `g065-rename`, executable head `db56209e` (tree `63420a25`), pipeline 2982223504 (`merge_request_event`).
- `validate:checks` job 17423736174: SUCCESS, 640 s.
- `validate:integration` job 17423736175 (played manually in the same pipeline on the same SHA at 17:10:50Z): SUCCESS, 4514 s
  (75 min). Its log ends `final status 0; published=full`, every wrapper phase 0. Profile `pipeline=full executed=full`, reason `foundation`.
- Codex independently inspected both jobs' artifacts: required-set union 123/123, controls 529/529, VB7b used the real witness `151c09bd27e8`
  (removed and stopped, expected negative-scenario outcome), no FAIL or UNMET rows. The only leftover in each hosted artifact is the existing
  shared `project-archetype-render-cache` volume (not a G065 object; it is not claimed as zero objects).
- First failure policy was not exercised: no hosted failure, no retry, no B3 attempt 3.

### Prose delta and review

- `4ddadaa9` changes only this Result's prose (opening status, Spec r6 and Plan r3 §11 links, budget; the Validation section limited to the
  original rename tree) in answer to Duo's thread on `db56209e`. `infra` `3a0cf7a3…` and `docs/foundation` `29abc394…` equal the
  qualified head's, and the diff against `db56209e` is the one Result file.
- Its pipeline 2982503255: `validate:checks` job 17425756202 SUCCESS (755 s). `validate:integration` (17425756203) was **not** played: the
  delta is documentation only and Codex accepted it under the Spec r6 documentation/management exception (no rerun).
- Duo approved at 18:29:20Z with nothing further to comment on; open discussions: 0 (the legacy-key thread and the opening-status thread
  were both resolved).

### Merge

- MR !87 merged non-squash with source removal as `4afca837` (parents `c5e010e8` staging and `4ddadaa9`), tree `cf51a790`. Auto-merge used
  the standing rule, source-pinned to `4ddadaa9`.
- Merged executable trees: `infra` `3a0cf7a3…` and `docs/foundation` `29abc394…`, equal to the hosted-tested trees. The merge differs from the
  source head only by the management-only staging commits already on the target (G066 definition, RC schedule, review records). No
  executable path changed after `db56209e`, so no post-merge local rerun was performed (tree match).
- `g065-rename` is deleted on the remote.

### Cleanup receipt

- Worktrees reclaimed: the writer worktree `agent-<id>` (clean, `g065-repair` equal to its remote) and `g065-run2`
  (frozen rename candidate `f77e7823`; it was reachable from no branch, so it is preserved first as
  `platform-evidence/g065-a1/g065-run2-f77e7823.bundle`, sha256 `7a93c3b5c3ca827174508fa38710a209593da4546f61176f1cf11b5bea6d5a1f`). The
  diagnosis, planner, push and prose worktrees were removed earlier.
- Retained, awaiting an owner deletion decision (no unique history is deleted): remote and local `g065-impl` (the failed intermediate
  candidate history, `566fb7cd` and `459139de`) and `g065-repair` (`915aaf5e`, the pre-rebase repair commits); remote and local
  `g065-plan` and `g065-plan-a1`, which are ancestors of staging. The held `g064-*` and `worktree-agent-*` refs are unchanged
  (`platform-evidence/g064-a1/HELD-REFS-20261023.md`).
- Docker, observed after the hosted run and the worktree reclaim: 15 containers, 11 volumes, 6 networks, 52 `image ls -q` lines (Codex counts
  50 unique image IDs). The one volume fewer than the G064 baseline is the owner's removal of the unused `project-archetype-render-cache` on
  the local daemon (receipt on staging `67499eda`). G065 created no persistent local objects; hosted runs used their own daemon.
- Local scratch under `platform-evidence/g065-a1/` and `g065-a2/` (diagnosis, qualification and proof records, the bundle above) is retained
  as evidence until final ACCEPT.

### Budget (B3)

Attempt 1 spent (rename and mapping correction); the repair K1–K3 + VB7 was attempt 2 and passed the hosted gate on its first run. No
attempt 3 was used. The ledger is the supervisor's.

### Cleanup receipt, batch 2 (2026-10-23, after Codex's cleanup disposition)

- Verified compact bundles of the unique branch history, all `git bundle verify` OK, under `platform-evidence/g065-a1/`:
  `g065-impl-459139de.bundle` (sha256 `bc2d9ff963be1522a1ef91d4a15cc9c6fdb8a30c23dbb12db3ca3bf602c64152`, 10202 bytes),
  `g065-repair-915aaf5e.bundle` (sha256 `c6aa35f5bddf66ac22899847de9fa2eafe81e4dab0c2831cc1b6b78b85a6bc04`, 24780 bytes) and the earlier
  `g065-run2-f77e7823.bundle`.
- Ancestry verified against staging: `g065-plan` `4b28d20e` and `g065-plan-a1` `2af29d26` are ancestors of staging.
- Ref removal (remote and local `g065-impl`, `g065-repair`, `g065-plan`, `g065-plan-a1`, then `g065-result` after this MR merges) is performed
  at final Goal ACCEPT under the mandatory acceptance cleanup, not before it; none of these refs was deleted at the time of this record. The
  G064 held refs are unchanged.
- Retained evidence, owner Claude coordinator: `platform-evidence/g065-a1/` (about 840 KB) and `g065-a2/` (about 44 KB), holding the
  qualification, diagnosis and proof records and the bundles. Reason: required failure and qualification proof and unique history.
  Release condition: the final-closeout preservation or compaction assessment.
- This MR (documentation only) creates no pipeline: GitLab started no `merge_request_event` pipeline for it, recorded here as the actual
  exemption; no runtime validation was run to populate one.


### Supervisor final acceptance and cleanup — 2026-10-23

Codex merged Result MR !88 at `efe6a9e1`, verified qualified executable-tree
equality and records final Goal ACCEPT in the linked review. The earlier
batch-2 ref-retention statement is superseded: verified impl/repair/run2 bundles
preserve unique history; remote/local `g065-impl`, `g065-repair`, `g065-plan`,
`g065-plan-a1` and `g065-result` are removed, as are stale G065 tracking refs.
Fresh remote/local ref queries confirm zero G065 refs. Finished G065 worktrees
are absent; live coordinator/successor work and G064 holds are preserved.

Local `sdlc.ci.owner` queries confirm zero containers/images/volumes/networks.
The small required failure/qualification proof and verified bundles remain
accepted-Goal evidence under Codex ownership, pending verified durable archive/
recovery replacement; assessment found them already compact (under 1 MiB total).
No disposable worktree or bulk temporary data is retained. No new tests or
permission setting were introduced to close this documentation-only record.
