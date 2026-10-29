> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G065 — Goal-neutral CI-control labels — Implementation Plan

Revision: r2, 2026-10-23 — the Codex correction loop (not included)
(sha256 `9363a6594b1efeecfe88cc37691316f82a9b09b2f31ce44a197b26459f393415`) applied in place to the
reviewed r1 (`45a91202`, estimate `62cb2c44`), with the correction record in §9. Planning only; nothing was
implemented, built, run or cleaned.
Goal: [G065](../goals/G065-neutral-ci-control-labels.md) (sha256 `2ac2b8ef0e0cff09d61dfb88c9c87ae701ed731fdb3d68d685d6a3bd1545c8ff`).
Spec: [r3](../specs/G065-neutral-ci-control-labels-spec.md) with the appended RC8-independent start
clarification and the pre-merge validation-timing exception (sha256
`6696ac213bc1f32bef8690af7a9c4754afc7fccca127fbab8a3b7d491c82494a`, superseding the r1-bound
`19757259…`; r1/r2 snapshots are historical), both at published staging
`f99099ac22044b1b7ea3fea7687c92aab6764ae0`.
Source base: `bb2b4e37f7c69c273065581fb82122884177424b` (G064 final ACCEPT,
[review](../results/G064-final-acceptance-codex-review-20261023.md)); the three source files are
byte-identical between `bb2b4e37` and `f99099ac` (`git diff --stat` empty), so every line number
below still holds. Budget: proposed B3, zero spent (a planning correction is not an execution attempt).

> **Current budget binding (r3 correction, 2026-10-23; the historical statements above and in §9 are preserved as
> written).** B3 = three shared execution attempts, checkpoint after attempt 2. **Attempt 1 is spent** on the original
> rename candidate (`f77e7823`, Result "Budget"); the owner-directed RC8 repair of §11 is **attempt 2** and does not
> reset the counter. Sources: Goal "Iteration Budget" (lines 88–92 at staging `492e5ca5`: "Attempt 1 is spent on the
> original rename candidate; the owner-directed RC8 repair does not reset the counter") and Spec r6 "Plan, budget and
> execution boundary" (lines 105–106: "the original rename has spent attempt 1, and the amendment does not reset the
> attempt ledger"). Where §9/§8 say "zero spent", read them as the r2 planning-time statement.

## 1. Approach

One mechanical commit renames exactly four Docker label keys everywhere they are
written, read, queried, seeded or documented in the current tree, preserving every
value, the owner-key format (`v1.<boot>.<pidns>.<pid>.<start>.<checkout>.<run>`),
run/control/nesting identity and all cleanup/recovery behaviour:

| Old key | New key |
|---|---|
| `g003.controls.run` | `ai.beginify.run` |
| `g003.control` | `ai.beginify.ci.control` |
| `g003.outer` | `ai.beginify.ci.outer` |
| `g003.controls.owner` | `ai.beginify.ci.owner` |

R8 wraps that commit in one verified maintenance boundary on the shared local
daemon: stop old-label creators, remove their exact owned leftovers with the
accepted G064 mechanism, prove old-key absence once, cut over, prove the new
emitted labels and cleanup, record the cutover SHA. No dual writes, no legacy
readers, no compatibility layer, no new framework, no other resources, no rewrite of
historical evidence. Only the exact keys change: helper names (`g003_*`, `G003_*`),
state paths (`g003-pending`, `g003-controls-*`, `g003-local`, `g003-validate-*`),
image/project/volume names (`g003ctl-*`, `g003-vtools-*`), the distinct fixture keys
`g003.v27`/`g003.vc9` and every management record stay as they are.

## 2. Exact inventory of the four keys (grep over `bb2b4e37`)

Three tracked files carry the keys outside management records. Line numbers are at the base SHA.

### 2.1 `infra/ci/validate-controls.sh` — 42 exact-key lines + 5 escaped-regex forms

| Role | Lines | Notes |
|---|---|---|
| writer | 261 `OWNER_LABEL`, 336 `RUN_LABEL`, 1382 `case_labels` (shim), 3875 `ctl_labels`, 3879 `mk_image` `commit --change LABEL` | the four real creators of labelled objects |
| writer (fixture) | 2496 nested-hold simulation (`control`, `controls.run`, `outer`, owner) | `hold_sim` labels |
| reader | 710 (`case g003.outer=*` on `--extra-label`), 894 (`shim_gate` `case` on `SHIM_LABELS`), **897 escaped regex `^g003\.control=`** | derive outer scope / control id from a label |
| query | 514 `gate_reserve` label arg, 766 `owned_cleanup` `ol`/`cl`, 2708 `nested_cancel` `ol`, 3626 `lcount`, 4267/4319 `image ls --filter label=`, 4269 `lcount`, 7228 `sweep`, 7231 held message | `--filter label=KEY=VALUE` and the sweep |
| fixture (stand-in daemon, V18n driver) | 2869 fake `inspect` `--format` match + `sed`, 2917 `FAKE_PS_FAIL=g003.outer=` (keep the trailing `=` shape), 2938 `OL=`, 3046/3047 `seed` label lines | driver heredocs `FAKE` and `CLDRV` |
| test assertion | 3103 exact; **escaped regex at 2972, 3012, 3028, 3102** | log/verdict patterns |
| comment/prose | 35, 230, 237, 371, 372, 694, 704, 705, 886, 887, 2484, 2490, 2499, 2501, 2696, 2862, 2932, 2933, 3039 | reword with the same commit |

Out of scope in this file (match `g003` only): `g003_*` helper calls (239, 243, 255, 258, 263, 3032, 3054–3058), state paths (273, 301, 303, 338, 342, 999, 1493, 4142, 4843, 4871, 5049–5058), names `g003ctl-*` (3863, 3951, 4026, 4205, 4568, 7130), fixture keys `g003.v27` (4071–4083) and `g003.vc9` (4658–4660).

### 2.2 `infra/validation/lib.sh` — 5 exact-key lines + 2 prose forms

| Role | Lines |
|---|---|
| reader (resource accounting) | 522 `_res_live` `{{.Label "g003.controls.run"}}` — column four of the container listing; at 545 `$4 != ""` admits the object to the run-wide all-objects count `RES_ALLN` only. Foreground membership/exclusion is decided at 543–544 by `g001.run`, registered projects, registry `R` rows and the background set (`B` rows `standalone <id> ci-*`/`bg-*`, or `g064.bg`), never by this key |
| query (startup guard) | 1431 `G003_KEYS` (the four filter keys), 1519 `inspect --format` on `g003.controls.owner` |
| prose | 517, 1422 comments; 1512 comment "a g003 control key"; 1550 operator text "its `g003.*` labels" (guard output; no control asserts on it) |

Out of scope: function/variable names `g003_*`, `G003_OWNER_RE`, `G003_KEYS`'s name, `G003_GUARD_MAX`, `g003_state_root`/`g003-pending` (1141–1161), 3854 image-name comment.

### 2.3 `docs/foundation/OPERATIONS.md` — 4 lines

722–723 (resource accounting: `g003.control=<run>-<control>`, `g003.outer`) and 757, 759 (one-time leak procedure: `--filter label=g003.controls.owner`, "likewise for `g003.outer`, `g003.control`"). Lines 387 and 749 are state paths: unchanged.

### 2.4 Verified absent / historical

- No other tracked file carries the keys: `infra/ci/validate.sh`, `run.sh`, `cleanup-owned.sh`, `gate`/shim libs, `.gitlab-ci.yml`, `infra/ci/profile-paths.tsv`, `.claude/settings.json`, `skills/`, `.agents/`, `docs/agents/` all clean (`validate.sh`/`run.sh` only hold paths and image names).
- **Archetype: not touched.** `.archetype/outputs/2ea491555634dde6/manifest.json` lists none of the three files and no template renders them (`skills/saas-foundation/templates` has no `OPERATIONS`, `validate-controls` or `validation/lib.sh`). `B-RENDER-ID/PLAN/REPEAT` read the manifest/templates only, so the G064-maintenance renderer regeneration is **not** triggered. `OPERATIONS.md` appears in `run.sh` `APPLICABLE` (the `off` variant copies it; content-agnostic).
- Historical, unchanged (15 management files): `goals/G065*`, `specs/G065*-r1/-r2/r3` (Codex-owned mapping tables), `plans/G003-*`, `plans/G064-*`, `results/G004-*`, `G007-*`, `G051-*`, `G062-*`, `G064-batch-*`, `G064-integration-test-isolation-result.md`, `ONEOFF-rc-helper-*` (2), `PROCESS-continuous-cleanup-20261022.md`.
- Machine-local, outside the tree: any `.claude/settings.local.json` or shell history with `label=g003.*` filters is operator convenience, not a repository reader; note in the Result if found, change nothing there under this Goal.

### 2.5 Cross-file couplings the single commit must keep consistent

1. `RUN_LABEL` (vc 336) ↔ `_res_live` (lib 522/545): the controls-run label feeds `RES_ALLN`, the run-wide simultaneous sample (`resources.tsv` `observed_max`, a lower bound), not foreground exclusion. A mismatch undercounts that sample for every control container that has no registry row yet (no latch, wrong accounting evidence). Mirrored controls stay out of foreground sets through their registry `B` rows and `g064.bg` (543–544), which the rename does not touch. (Corrected in r2; r1 described this as the foreground-exclusion mechanism.)
2. `OWNER_LABEL` (vc 261) ↔ guard `inspect --format` (lib 1519) and `G003_KEYS` filters (lib 1431) ↔ the V18n fake daemon `inspect` (vc 2869) and `seed` (3046–3047).
3. `nested_cancel` `--extra-label` (2708) ↔ nested runner readers (710, 894, 897) ↔ `case_labels` (1382) ↔ assertions (2972, 3012, 3028, 3102–3103) and `OL` (2938).

## 3. R8 cutover and cleanup boundary (reuse only; nothing new)

**Who/what must be stopped (shared local daemon only).** Every live creator of the old labels: any `infra/ci/validate-controls.sh` process (direct, through `run.sh ci-controls`/`all`/`static`, or a `validate.sh --profile full` run from any checkout — coordinator qualification runs included) and its nested runners (V18l–V18n) and the V18n cleanup driver. Not stopped: source work, unit/private DB/browser runs and G020's isolated implementation (they cannot create these labels); hosted jobs (own daemon).

**Absence verification (one-off, not a runtime layer).**
1. No live creator: `pgrep -af validate-controls.sh` empty; `<state root>/g003-pending/*.tsv` holds no `launched` marker with a live PID; no `<state root>/g003-controls-*/owned` or `<evidence>/owned` registry with unfinalized rows.
2. Old-key absence query, 4 keys × containers/networks/volumes/images, exactly the OPERATIONS.md "one-time procedure" listing (`docker ps -a --filter label=<key>`, `network ls`, `volume ls`, `image ls -a`): all sixteen listings must succeed and be empty. A failed listing is "unavailability is not absence": hold.
3. Any survivor is first classified by its `g003.controls.owner` value exactly as `g003_owner_class` does. **That result establishes liveness only**: `live` → wait for its owner to finish, then re-query; `unknown:*` → hold. `dead:*` authorizes nothing by itself. Before any fallback removal, ownership of each dead-owner object is proven independently from the existing records (r2, review item 2), and only an object that passes every applicable test is removed by exact ID:
   - **Run and checkout.** The owner key's checkout id must equal `g003_checkout_id` of this checkout's canonical root, and its run id must appear in one existing record of a run of this checkout: the pending marker `<state root>/g003-pending/<job>.tsv` (`repo`, `boot_id`, `pid`/`pid_start`, registry and evidence paths) or the run registry header (`job`, `repo`, `work_root`, `label`, `project` rows) of `<evidence>/owned` / `<state root>/g003-controls-*/owned`.
   - **Exact ID.** The object's full ID appears as a resource row of that registry (`container|network|volume|image`, `standalone <id> ci-<control>`, a `cid/` cidfile) or of the control inventory `owned/<control>.tsv` (its `problems/<control>` file is the held receipt), or in the control's own labels (`control`/`outer` values naming that run and control). An ID present in that run's **start inventory** is never the run's.
   - **Volumes.** A named volume carrying the control label and the registry row; an anonymous volume only through the res-witness journal (`res-witness/attr.tsv`, `bound.tsv`, `gone.tsv`: attributed solely by its witnessed mount to an attributed container) and `docker inspect` showing no other container mounting it. A volume mounted by any live container or carrying `com.docker.compose.project` of a persistent stack holds.
   - **Images.** Only an image carrying the control label whose ID is a registry/inventory row of that dead run, not listed in the registry's `images-shared` note (an ID that pre-existed the build), with `docker ps -a --filter ancestor=<id>` empty and no other repo tag; shared base/tool images hold. Precedent: the G064 corr7 `g003ctl-6f3da70d051a-vc8-p-svc` image (inspected, labelled, no containers, then removed).
   - **Mechanism.** Registry-owned → `infra/ci/cleanup-owned.sh --mode finalize` on that registry (its `refused`/`unproven` rows hold the cutover; nothing is forced). Otherwise the exact-ID commands of OPERATIONS.md lines 755–762 (`docker rm -f -v <id>`, `network rm`, `volume rm`, `image rm`).
   - **Hold.** A foreign checkout id, a shared or in-use object, a project-labelled object, or an ID with no record → **hold cutover, escalate** with the object's bounded description. Never remove by label, name pattern or dead-process evidence alone. No new framework, campaign or sweeper.

**Order.** (1) Record the stop in the coordinator run inventory; no new old-label launch on this daemon from any checkout until step 7. (2) Stop/reap task-owned live creators (TERM, bounded grace, KILL, reaped), finalize their registries. (3) Absence verification above; reconcile pending markers/registries. (4) Commit the rename (one commit: all writers/readers/queries/fixtures/docs of §2). (5) First launch on the new keys: the startup guard (new `G003_KEYS` values) must print `clear`; run the §4 focused selectors; confirm emitted labels by `docker inspect` of one control's live object during a group-E control and by the V18n fake-daemon logs. (6) Post-run: new-key listings (same 16 queries) empty; Docker inventory back to the pre-step counts (G064 final baseline 15 containers / 12 volumes / 6 networks / 52 images, or the actual pre-step count recorded in step 1). (7) Record the receipt and the cutover SHA; resume workloads on post-cutover source only.

**Holds.** Unresolved/unknown/foreign-live survivor; a dead-owner object whose ownership the records above do not prove; failed listing; refused removal; a frozen checkout older than the cutover SHA that still needs a `ci-controls` run on this daemon (see §6 risk 2); a live old-source harness run overlapping the first new-label launch (§9). A hold is reported, not worked around.

**Receipt (in the Result, compact).** Pre-step inventory counts; the stop list (PID, checkout, registry/marker, outcome); every removed object (kind, exact ID, old label values, owner class, mechanism used); the sixteen old-key listing outputs (empty) with timestamp; marker/registry reconciliation; post-cutover sixteen new-key listings after the §4 runs (empty) and final counts; cutover commit SHA; retained items with owner/reason/release condition. Historical labels in old evidence stay as written.

## 4. Validation (existing checks only; focused first)

| Check / selector | Why it is affected | Where |
|---|---|---|
| `B-SHELL` (`lint` suite; run with `lint-base` locally per the canary rule) | shellcheck over the two edited scripts | hosted fast job (automatic) + local, once on the reconciled candidate |
| `bash infra/ci/validate-controls.sh --only V18l,V18m,V18n` | nested runners receive `--extra-label ai.beginify.ci.outer=`; V18m outer-label sweep; V18n cleanup driver (17 cases incl. `held-query` `FAKE_PS_FAIL`, `unrec`, `outer-*`) and the 9 startup-guard cases on the renamed fake-daemon labels — couplings 2 and 3 | local, once on the reconciled candidate, before merge |
| `--only V27,V27b,VC1,VB1` | real-Docker image labels via `mk_image`/`commit` (V27/V27b), the `image ls --filter label=` ledger counts (VC1, 4267/4319), `ctl_labels` + `lcount` in `vb_gone` (VB1, 3626), and every control's `owned_cleanup`/`sweep`/unrecorded check on the new labels | local, once on the reconciled candidate, before merge (the G064 affected set reduced to its label-carrying members; VR1/VR2/VR9 carry no label and are not repeated) |
| Bounded real-Docker invocation of the existing `_res_live container` (below) | coupling 1: the renamed run label must feed `RES_ALLN`; mirrored objects must stay out of the foreground set — neither selector above samples `_res_live` with a control object present | local, once on the reconciled candidate, before merge; estimated, unmeasured |
| Hosted MR fast job (`validate:checks --profile fast`) | automatic on every MR push; `infra/ci` and `infra/validation` are class `foundation` → pipeline profile `full`, so the `validate:integration` remainder on a feature MR stays the optional manual job and is **not** requested | hosted |

**Timing (owner validation-timing exception, 2026-10-23).** The four local rows run **once on the final reconciled candidate before merge** (after the fetch/rebase reconciliation of the branch-workflow rule) and replace the post-merge local run: the Result records the tested SHA and tree, commands, outcomes and cleanup; after merge the coordinator verifies the merged SHA's tree against the tested tree, and an equal tree needs no repeated execution. A later executable delta (candidate or integration) gets only the affected rows rerun; a documentation/management-only delta is reconciled explicitly in the Result, not rerun. Results are never transferred silently to changed executable inputs. Hosted fast, Codex source review, Duo, cleanup and final ACCEPT gates are unchanged.

**Bounded `_res_live` invocation (existing function, no new control or suite).** Run inside the owner's test slot while no controls launch is in progress (its objects carry no owner label and would hold a concurrent startup guard as `unknown:missing`), following the G064 stack-free bounded-reproduction convention (`f0-*` in the G064 Result); a short log under the task evidence root, exact IDs recorded before anything else:
1. Setup: `RUN_ID` 12 hex; `WORK=$(mktemp -d)` with an empty `$WORK/projects`; `RESD=$WORK/res` with an empty `exclude.ids`; `OWNED_REG=$WORK/owned.tsv` (header rows only). Baseline: source `infra/validation/lib.sh`, call `_res_live container`, record `RES_ALLN` as n0 and `RES_IDS`.
2. U (unmirrored): `docker create --cidfile $WORK/u.cid --label ai.beginify.run=$RUN_ID --label ai.beginify.ci.control=$RUN_ID-g065proof busybox:1.37.0 true` — no `g001.run`, no project, no registry row.
3. M (mirrored): `docker create --cidfile $WORK/m.cid --label g001.run=$RUN_ID --label ai.beginify.run=$RUN_ID busybox:1.37.0 true`, then append the registry row `standalone<TAB><M full id><TAB>ci-g065proof`.
4. Call `_res_live container` again. PASS iff `RES_ALLN − n0 == 2` and neither U nor M is in `RES_IDS`. Then remove M's row and call once more: M must now appear in `RES_IDS` (the `B` row, not the key, is what excludes it) while `RES_ALLN` is unchanged.
5. Cleanup in a trap: `docker rm -v` by the two cidfile IDs, `docker ps -a --filter label=ai.beginify.run=$RUN_ID` empty, `$WORK` removed; IDs, outputs and cleanup go into the Result.

Why U depends on the renamed key: U has no `g001.run`, no project and no registry row, so the only clause that admits it to the all-objects set is `$4 != ""` (lib 545), i.e. the `{{.Label "ai.beginify.run"}}` column. If `lib.sh` still read `g003.controls.run`, column four would be empty for U, `RES_ALLN − n0` would be 1, and step 4 fails; no registry row exists for U that could mask that. M's inclusion is key-independent by design (it is the foreground-exclusion observation). This proves the single `_res_live` reader; it is not a `B-CI-CONTROLS` PASS.

Not proposed: the complete 528-control `ci-controls` suite, `run.sh ci-controls migrations`, any local full run, a new control, fixture, compatibility matrix, dual-label test or duplicate campaign; `AC3–AC7` stay retired. The selected controls are reported as exactly those controls; `B-CI-CONTROLS` remains unexecuted for this Goal and is never called PASS. Allow-file for direct `run.sh --base <staging SHA>` runs, if used instead of `validate.sh --base HEAD`: `^infra/ci/validate-controls[.]sh$`, `^infra/validation/lib[.]sh$`, `^docs/foundation/OPERATIONS[.]md$`, `^management/v0/results/G065-.*[.]md$` (materialize only if needed). No edits to the checkout during a running batch.

## 5. Files, sequence, delegation

Changed: `infra/ci/validate-controls.sh`, `infra/validation/lib.sh`, `docs/foundation/OPERATIONS.md`; new `management/v0/results/G065-neutral-ci-control-labels-result.md`. Unchanged: schemas/migrations (none), public API/MCP (none), `.archetype/*`, `skills/*`, `profile-paths.tsv`, `.gitlab-ci.yml`, all historical records.

Sequence: S0 read-only daemon/process inspection and inventory snapshot → S1 R8 steps 1–3 (stop, finalize, ownership-proven absence receipt) → S2 the rename commit (diff reviewed against §2 line by line; `grep -nE 'g003\.(controls\.run|control\b|outer|controls\.owner)'` must return only management records) → S3 fetch/rebase reconciliation, then the focused pre-merge qualification once on that reconciled candidate: `lint lint-base`, `--only V18l,V18m,V18n`, `--only V27,V27b,VC1,VB1`, the bounded `_res_live` invocation → S4 R8 steps 5–7 receipt (tested SHA/tree recorded), Result, safekeeping push → S5 Codex source review → MR to staging (hosted fast, Duo) → merge → S6 merged SHA/tree verified against the tested tree (equal: no post-merge execution; executable delta: affected rows only; management-only delta: reconciled in the Result) → task cleanup receipt → final ACCEPT.

Delegation: one Claude implementation owner (Opus 5.5 high) does S0–S4; no subagent for the rename; Codex's optional read-only review of the §2 mapping/diff. The owner holds the single test slot while the selectors and the bounded invocation run (shared-daemon operation, not a private run).

## 6. Risks, questions, escalation

1. **`_res_live` coupling (lib 522/545).** A rename error there would not latch anything but would undercount the run-wide sample in later full runs. It is covered by the bounded invocation in §4 (key-dependent by construction); no broad run and no supervisor decision remain open (r1's option (a)/(b) is withdrawn).
2. **Pre-cutover checkouts on the shared daemon.** After cutover, any `ci-controls`/full run launched from a checkout older than the cutover SHA (frozen qualification trees, old worktrees) writes old labels that the new guard cannot see. Hold such launches or finish them before step 4; record it in the run inventory. Hosted jobs are unaffected.
3. **Foreign/unknown survivors** hold cutover (Spec R8); no deletion authority from a label, no silent compatibility reader.
4. **Escaped regex and glob forms** (vc 897, 2972, 3012, 3028, 3102; the `*g003.controls.owner*` glob at 2869; `FAKE_PS_FAIL` shape) are easy to miss with a plain replace; the S2 grep gate catches the exact keys, the reviewer checks these five lines explicitly.
5. Scope stop: any further `g003` string, other label namespaces (`g001.*`, `g064.bg`), state-path or name renames, or a label rollout to other resources is outside R1+R8 and is reported, not done.

## 7. Acceptance mapping

| AC | Evidence |
|---|---|
| AC1 | §2 inventory vs. the diff (line-by-line review), the S2 grep gate, emitted-label assertions: V18n guard/cleanup cases, group-E sweeps and ledger counts in the seven §4 selectors, and the bounded `_res_live` invocation for the lib.sh reader |
| AC2 | §3 receipt: stop list, ownership proof per object from the named records, exact-ID removals via the accepted mechanism, sixteen empty old-key listings, marker/registry reconciliation, old-reader accounting (§9), cutover SHA; any hold recorded as a hold |
| AC8 | `B-SHELL`, the seven §4 selectors and the bounded invocation PASS once on the tested candidate SHA/tree before merge, merged tree verified equal (or affected rows rerun on an executable delta), pre/post inventory counts equal, post-run new-key listings empty, hosted fast green on the MR source SHA, task cleanup receipt; historical evidence and unchanged G064 passes untouched; unexecuted coverage (`B-CI-CONTROLS`, full profile) reported as unexecuted |

## Revision history

- r1, 2026-10-23 — initial Plan from the current tree at `bb2b4e37`; no execution (`45a91202`; coordinator estimate §8 added in `62cb2c44`, reviewed hash `ab067b69…`).
- r2, 2026-10-23 — Codex CORRECTION LOOP applied in place (§2.2/§2.5 coupling corrected; §3 ownership proof before deletion; §4/§5/§6/§7/§8 validation reduced to the seven selectors plus the bounded `_res_live` invocation, placed before merge under the owner timing exception; §9 bindings and old-reader accounting). No execution.
- r3 addendum, 2026-10-23 — §11 RC8 harness repair amendment (K1–K3, AC9) under Spec r6 (r5 repair scope, r6 hosted-only qualification) at staging `a63c2ac9`; additive, no execution.
- r3 correction, 2026-10-23 — Codex CORRECTION LOOP (`134464a3…`) applied in §11 at staging `492e5ca5`: K3 guard gap, K2 fail-closed, VB7 hosted regression, `lcount` sentinel handling and `run.sh` inventory, header budget binding (§11.6). No execution.

## 8. Estimate (coordinator addendum, 2026-10-23; provisional, not a commitment)

Basis: recorded per-control durations from run 4 (`qual-pre-9cffb495`, `ci-controls-timings.tsv`), the measured hosted fast pipeline (about 10 minutes on !82/!84) and the measured local fast profile (about 7 minutes). Unmeasured items are stated as unknown.

| Part | Work | Estimate | Basis |
| --- | --- | --- | --- |
| Coding | The mechanical exact-key rename in `validate-controls.sh` (42 lines plus the five escaped-regex/glob forms), `lib.sh` (5 lines plus 2 prose lines) and `OPERATIONS.md` (4 lines); self-review of the diff; one writer, stack-free | 30-60 min | 51 matching lines, no logic change |
| Cutover | Stop old-label creators, sixteen one-off old-key listings, exact-ID removal of dead-owner survivors, one rename commit, first-launch label inspection, new-key listings, baseline inventory, receipt | 15-30 min | depends on whether any old-label resource or live creator exists; unknown until the first listing |
| Checks (once, pre-merge, on the reconciled candidate) | `lint lint-base` (B-SHELL) about 6 min; `validate-controls.sh --only V18l,V18m,V18n` 59+60+123 = 242 s plus scratch/startup; `--only V27,V27b,VC1,VB1` 28+14+10+19 = 71 s plus startup; the bounded `_res_live` invocation (two `docker create`, three function calls, exact-ID removal) | 20-30 min | recorded selector durations; controls-run startup overhead and the bounded invocation (estimated 2-5 min) are **unmeasured** |
| Post-merge | tree verification of the merged SHA against the tested tree; affected-row reruns only on an executable delta | 0-5 min (equal tree) | no post-merge local run under the owner timing exception; a delta rerun is not estimated |
| CI and review waits | hosted fast about 10 min per push; Duo about 5-10 min per round; Codex source review, plan review and the cutover receipt review | 45-120 min | Codex review time is not measured; Duo and hosted times are |

Total: about 2 to 3 hours end to end, of which review waits are the largest and least predictable part; this counts only the checks §4 actually selects. No `ci-controls` suite, `ci-controls migrations` or full-profile run is scheduled or estimated, and none is called PASS (r2 removed r1's optional broad-coverage row, which the review found inconsistent with the estimate). Not estimated: an unexpected live creator, an unproven-ownership or foreign-owner survivor found by the cutover listing (holds the cutover until resolved), or an affected-row rerun after a later executable delta.

## 9. Correction record r2 — bindings and old-reader accounting

**Bindings reconciled (review item 3).** Staging advanced `bb2b4e37` → `f99099ac` with management-only commits (`75196a23` RC8 candidate snapshot, `1dc56f69` RC8-independent release, `d5072d73` Plan review, `f99099ac` owner pre-merge timing exception). `git diff bb2b4e37 f99099ac -- infra/ci/validate-controls.sh infra/validation/lib.sh docs/foundation/OPERATIONS.md` is empty: the §2 inventory, line numbers and couplings are unchanged and no test is repeated for this reconciliation. Current hashes: Spec `6696ac21…` (r3 + RC8-independent clarification + validation-timing exception), Goal `2ac2b8ef…`, Plan review `9363a659…`. Consequences: RC8 qualification is independent and gates neither G065 planning, implementation, integration nor final ACCEPT; the selected local checks run before merge (§4 timing); the "full local post-merge" default of CLAUDE.md is superseded for G065 by the owner exception; B3 stays at zero spent; implementation still requires Plan ACCEPT and current-input/resource reconciliation.

**Old readers accounted at cutover (review item 3; Spec R8 step 1 covers creators).** Consumers of the four old keys that may still be live when the rename commit lands:
1. The startup guard (`g003_owner_guard`, lib 1431/1519) — runs only at a controls launch; stopped with the creators in §3 step 2.
2. `_res_live` (lib 522) — runs in **every** `validate.sh`/`run.sh` harness run of any profile, fast included, from any checkout. A harness run of pre-cutover source that overlaps the first new-label launch (§3 step 5) would omit the new-label control objects from its `RES_ALLN` (accounting only; no latch). Rule: before step 5, verify no live harness run (`pgrep -af 'infra/validation/run.sh|infra/ci/validate.sh'`, pending markers) from pre-cutover source on this daemon; otherwise wait or hold. Not stopped: runs that do not share the daemon (hosted), private DB/browser runs and source work.
3. `owned_cleanup`/`sweep`/`shim_gate`/nested readers (vc 710, 766, 894, 897, 2708, 7228) — live only inside a running controls run; stopped as creators.
4. The operator runbook text (OPERATIONS.md 757–759) and the guard's recovery text (lib 1550) — cut over in the same commit; no live consumer.
5. `cleanup-owned.sh`, the `validate.sh` recovery gate and the registries are exact-ID/registry based and read no label key: unaffected.
A local run of the pinned RC8 source (or any pre-cutover tree) on this daemon after cutover would be both an old reader and an old creator: that is §6 risk 2 and a hold; hosted RC8 qualification uses its own daemon and is unaffected.

**Unchanged by r2.** R1 mapping, the three-file inventory, the §3 stop/absence/order/receipt, the selector set, delegation, risks 2–5 and the AC IDs.

## 10. r2 key-mapping supersession (addendum, 2026-10-23)

Spec r4 (staging `f3181d04`, Codex handoff f985ba65) replaces only the R1 **target keys**, because the r3 targets
contained the manifest identity literal `beginify` and failed `B-LITERALS` on candidate `566fb7cd`. The binding mapping is:
`g003.controls.run` → `sdlc.ci.run`, `g003.control` → `sdlc.ci.control`, `g003.outer` → `sdlc.ci.outer`,
`g003.controls.owner` → `sdlc.ci.owner`. Wherever §1–§9 name an `ai.beginify.*` key (including the §4 bounded
`_res_live` invocation labels), read the corresponding `sdlc.ci.*` key. Everything else in r2 — the three-file inventory and
line numbers, R8 cutover, selectors, timing, delegation, risks, AC mapping and budget — is unchanged and remains ACCEPTED;
this addendum is a mechanical supersession, not a new Plan revision, and needs no new approval round.

## 11. RC8 harness repair amendment (r3 addendum, 2026-10-23)

Additive; §1–§10 stand. Planning only: no Docker, no source edit, no build, no test was run for this section (fake
`docker` on PATH, `DOCKER_HOST=tcp://127.0.0.1:1`). Authority: Spec r6 at staging `a63c2ac9` (file sha256
`2b84be8c7934a140fcaf9940a05893db4b8faa27d401a7087dcb95c177104c1d`): its r5 amendment (snapshot
`G065-neutral-ci-control-labels-spec-r5.md`; r5 file sha256 `b1911cf0…`) adds K1–K3 and AC9 to G065 and **supersedes r4
only for the added repair scope**; its r6 amendment (owner: "waive the local pre merge run, and just trigger
`validate:integration` on CI", relayed via supervisor handoffs 4482c49b/a656b14c) **supersedes r5's local selector
requirements for qualification**: qualification is hosted only (§11.3). R1 + R8 and the accepted rename candidate are
unchanged; B3 stays 3 attempts with attempt 1 spent and no reset. Codex review of this amendment gates the added code.

### 11.1 Bindings and established facts

- Base: staging `492e5ca5b26cef9df86a4cabb25bcfd0d6e04929` (this branch after the r3 correction; `d2bf553a` →
  `a63c2ac9` → `492e5ca5` are management-only: Spec r6 + r5 snapshot, then the Codex review record; no `infra/` change). Implementation branch `origin/g065-rename` `db968e2b` = `67f55f7b` +
  rename `0708492a` + docs `59460d93`/`b271f504`/`db968e2b` (MR !87, draft, per supervisor; its pipeline `2981988185`
  is fast-only and pre-repair — never a qualification input). `git diff --stat d2bf553a origin/g065-rename -- infra/`
  is 54/54 lines in the two renamed files; all three harness files keep their line counts (7255/3924/6689), so the
  proposal's `bb2b4e37` `path:line` references hold on `g065-rename` and every line cited below was re-read there.
- CI path map, verified from the config: `infra/ci/profile-paths.tsv:61` `foundation<TAB>dir<TAB>infra/`;
  `infra/ci/select-profile.sh:55` `class_profile` maps `foundation` → `full` (rank 4, :53-54, doubt resolves upward);
  `.gitlab-ci.yml:136-138` makes `validate:integration` `when: manual`/`allow_failure: true` on an ordinary MR
  (`:122-139`), so the full profile for an `infra/` change is `validate:checks` plus the manually started
  `validate:integration` of the same MR pipeline.
- RC8 evidence: pipeline `2981578878`, job `17419324226` (3,082 s, exit 1), artifact zip sha256
  `a4e0aed5196a644728753ad018b683424073bd012632640e3afb2707eb92558b` (re-opened read-only in a fresh directory), tag
  object `65d3345f` → `bb2b4e37`, tree `f0c920cb`. MR !86 is closed by the owner; tag and failed record unchanged.
  Proposal: `~/.local/state/platform-evidence/rc8-0.2.0-rc.8/repair-proposal.md` sha256
  `751f8964fda834b6a5abc3bea87c7756f2ed645da48fdd1eff7fbec3d919af16`.
- Re-verified: **F4** (`ci-controls-summary.log:183-186`: VC8 FAIL "unrecorded image sha256:26c38bee… carries
  `g003.control=33e3e5c08c23-VC8`; removed 1; held", 181 PASS, 346 not started, exit-time retry verified; the only
  "cancelled by SIG" text is inside V18n's PASS detail at :129, so the control runner ended on its own exit path);
  **F9** (VC 4435 A via the gating shim 918-934; 4496/4500 D/D2 `compose build svc` fall through shim 994 with no
  inventory row; 4505 `mk_image` 3876-3882 recorded by `D` 525-532; 4507 re-tag; 4510 untag; 4511 `vc_clean`); **F10**
  (`vc_clean` 4267 `|| ids=""`, 4268 `|| true`; `lcount` 2690 `|| k=0`; `sweep` 551-552; `owned_cleanup` lists `-aq` at
  745 and classifies `unrecorded` at 785-791; held at 7226-7232/7197). One correction to F10: the gate ceiling count
  (`gate_count` 400) **already fails closed** (`|| return 1` → `gate_reserve` 417 latches "query failed"); its remaining
  K1 defect is only the `-q` listing. Also verified: F3 (`persist-start.log:26-27,43`; the error vertex prints under the
  `intelligence` target, the failing target is `web`), F2 (`harness.log:169-174`), F5 (`cleanup-inventory.tsv:1`), RP3–RP5
  (docker 29.8.2 overlay2, compose 5.5.1, buildx 0.37.2), F7 (`run.sh` 31/111/5450-5466/6654-6663), F8 (`lib.sh` 883).
- Unknown and left unknown: which of m1/m2/m3 produced the hosted VC8 hold; X ≡ `26c38bee` (inference); the `persist`
  inner failure at 5458; the authorization deadline cause. The local daemon version (27.5.1 per the proposal) was not
  re-read here. Hosted VC8 timings, `owned/VC8.tsv`, `problems/VC8` are not in the artifacts.

### 11.2 Design per item (prose; no code here)

**K1 — fail-closed, consistent image listing/removal** (`infra/ci/validate-controls.sh` only).
`vc_clean` 4265-4270: list with `image ls -aq --no-trunc`; a failed listing is a `prob` ("the image listing failed;
cleanup unverified") instead of `ids=""`; each failed `image rm -f` is a `prob` naming the exact ID instead of `|| true`;
the `lcount` re-check stays. `lcount` 2686-2694: `image ls -aq`; **separate the listing from the count** — today
`k="$(D $x … | grep -c .)" || k=0` masks both a failed listing and grep's exit 1 on zero matches (pipefail), which is
exactly why m1 is invisible; capture `o="$(D $x --filter …)"` first (a failure marks the call unverified), then count
`NF` lines of `$o`. On any failed listing `lcount` prints the sentinel **`-1`** and still returns 0 (a nonzero return
inside `$(…)` under the control's `set -e` at 7206 would end the control as an unclassified "control aborted (exit 1)",
7237). Every caller (complete list by grep: 2759, 2781, 2784 in `nested_cancel`; 3626 `vb_gone`; 4269 `vc_clean`; no
other) tests `(( n < 0 ))` **before** its existing comparison and records `prob "the Docker listing of <label> failed;
the count is unverified"`, then skips the existing numeric branch for that value (an `-1` would otherwise read as
"created no container" at 2760 or "-1 resources remained" at 2785-2786/3627/4269). VC4:4319 is kept with the same split
(listing captured, then counted; `-1` → `prob "the image listing failed; the consumer count is unverified"`, 4320
otherwise unchanged). `sweep` 551-552 and the verification list at 553: `-aq`; a failed listing/removal in the removal
pass sets `left=1` with a `SWEEP_LEFT` text (the verification pass at 555-557 already does this). `gate_count` 400: `-aq`
(fail-closed already). Why `-a` creates no new local false failures: `-a` adds only images that `image ls` hides
(intermediate images: a parent that has children); the label filter `sdlc.ci.control=<run>-<control>` still scopes the
result to this control's objects, the VC contexts all `FROM busybox`/commit from busybox (no labelled image is a parent),
and `owned_cleanup` already lists `-aq` with the same label after every control, so anything `-aq` adds is already
removed and reported today. Risks: a hidden labelled image mid-control now counts toward the VC ceiling (24, line 389);
VC4 asserts exactly one labelled image (4320) — the hosted VC1/VC4/VC8/VC9 outcomes measure both; a transient listing
failure now fails/holds a control instead of passing silently (intended; it is what AC9 asks).

**K2 — record Compose-built images, fail closed** (`validate-controls.sh` only). One helper `vc_record_tag TAG WHAT`:
`id="$(D image inspect -f '{{.Id}}' "$TAG" 2>&1)"` — on failure `prob "K2: the Compose-built image $TAG could not be
inspected after its build ($id)"` and return (the image, if it exists, is then found by `owned_cleanup` as `unrecorded`
→ hold, the existing fail-closed path); on success `gate_record "$ROOT/owned/$CTL_ID.tsv" image "$id" "$WHAT"
"$REG_MIRROR" "$TAG" || prob "K2: the inventory row for $TAG ($id) could not be written"` (439-453; same row the gated
builds write, deduplicated by ID, mirrored to the run registry as `image <tag> <id>`, `cleanup-owned.sh` 149), then
**verify** `grep -qF $'image\t'"$id"$'\t' "$ROOT/owned/$CTL_ID.tsv" || prob "K2: $TAG ($id) is not in the inventory
after recording"`. Nothing is skipped silently: every outcome other than "row present" is a `prob` naming the exact
service tag. VC8: call it after D (4497) and after D2 (4501) with `"$proj-svc:latest"` (`proj=$(vc_tag p)` =
`g003ctl-<RT>-vc8-p`; a failed build there already fails the control through `vc_want`/`vc_unaccounted`, the extra
`prob` is honest). VC9 A: after 4703, once per real tag `$(vc_tag one|two|three|four):latest` (`vc9_dest` 4644-4646;
`absent` never builds; B–E replays and `cd9dup` build nothing real). Interplay: before K2, X is removed by `vc_clean` (recorded or not) and
`owned_cleanup` sees nothing; after K2, if `vc_clean` ever misses X again, `owned_cleanup` removes it as a **recorded**
object (OWNED_N=1, no `unrecorded` hold) and 7227 reports it only when the control already has a problem — which K1
guarantees when the miss was a listing/removal failure. The 4507 re-tag makes the recorded tag stale (X becomes
untagged); the registry row's tag is informational, removal is by ID. VC8's `vc_want`/`vc_unaccounted` assert ledger rows,
not image counts, so recording changes no assertion.

**K3 — persist-start failure is a check failure, not a harness abort** (`infra/validation/run.sh` 5352, 5457-5460,
EXIT trap 111; one small helper in `infra/validation/lib.sh` beside `res_witness_stop` 1026). Every command of
`persist` 5450-5466 was checked under `set -euo pipefail`: 5453 (`|| { ok=0 … }`), 5454-5455 (`… && [[ ]] || { }`), 5456-5457
(`|| ok=0`), 5459 (`… || { }`), 5461 (`[[ ]] || { }`), 5462-5464 (`if` condition), 5465 (`&&`/`||` list) are guarded;
**5458** (`out="$(compose_migrate …)"`) and **5460** (`out="$(compose … run … mc cat …)"`) are not — either fails on a
stopped stack and re-triggers errexit. (a) Record-and-return at 5457: when `platform:start` fails, `ok=0`, `record
B-PERSIST FAIL "restart: platform:start failed (see persist-start.log)"`, `return 1`; and guards `|| ok=0` on 5458 and
5460 for the other failure modes (a D2 latch refusal `125` from `compose_call` 303, a `mc cat` error on a live stack),
so the grep/compare lines report them as "migration history lost"/"minio object lost". (b) `variants_ops` 5352:
`persist "$D" || { save_logs "$D" ops; compose "$D" down -v >/dev/null 2>&1 || true; return; }` — on a dead stack the
later `identity_unconfigured_probes` (5837 `out="$(docker exec "$web" …)"`, unguarded) and `credential_rotation` (5445
`n="$(… | rot_sql …)"`, unguarded) would abort the harness the same way; skipping them is the existing early-return
convention of this function (5346-5348), and their required rows (`required-checks.tsv` ops: B-IDENTITY-*, B-MCP-E2E,
B-SECRET-ROTATION) are then reported missing by B-REQUIRED-CHECKS at `final`. **Reviewer decision:** (b) is the
smallest change that actually preserves finalization; the alternative (guarding every later ops command) is out of
K1–K3 scope. Effect: B-PERSIST FAIL, `suite_ops` continues (`variants_identity`), the loop reaches the joins 6658-6659,
`res_witness_stop` 6663, `res_report`, `results.tsv`, `B-REQUIRED-CHECKS`, and the exit-1 contract at 6688. (c) Abort path: add `res_witness_abort` — when `$WORK/res-witness/w.id` exists and `stopped` does not, `_dro_remove`
(801-809: `rm -f -v`, verified gone, latch on failure) the recorded W and mark `stopped`; call it in the trap between
`bg_stop_all` and `cleanup`, guarded `|| log …` (the trap runs under `set -e`, like 111's existing guards). Not
`res_witness_stop`: on the abort path `bg_stop_all` has already stopped and `BG_DONE`-marked the parser, so its marker
wait would time out (10 s latch) and its `bg_join` would record "was not running". Interactions: joins unchanged;
`bg_stop_all` rows (2495) still appear for a real abort; W's registry row still lets `cleanup-owned.sh` remove it in the
hosted wrapper (F5) — K3(c) only closes the gap for `run.sh` runs without a wrapper and makes `inventory-leftover` honest;
no change to `cleanup()` 3839-3918, to `res_report` or to any cleanup semantics.

**K3 regression inside the hosted full profile (AC9 forced-start-failure proof; reuse of group D).** The hosted normal
path (`B-PERSIST PASS`) never executes (a)–(c), so one new control **VB7** in the existing group-D family ("background
supervision", `validate-controls.sh` 3492-3497; harness-shaped driver sourcing the real `lib.sh`) is added: table line
110 (`"D VB1" … "D VB6"` → append `"D VB7"`), function inserted after `ctl_VB5` (ends 3762, before the `vb_copy`
comment at 3763). Two deterministic cases, both stack-free:
- **VB7a (persist guard).** As the strict-mode precedent at 2679-2681: `bash -c 'set -euo pipefail; source lib.sh; …'`
  with `persist()` extracted verbatim from the candidate `run.sh` (awk from `^persist() {` to its `^}`), stub
  `compose`/`compose_migrate`/`run_docs_block` where `platform:start` and every later call fail, `RESULTS` in the case.
  Asserts: the shell exits 0 (no errexit abort), `if persist …` returns 1, `results.tsv` holds exactly one
  `B-PERSIST<TAB>FAIL` row and no PASS row; a second variant with `platform:start` succeeding and `migrate status`
  failing asserts one FAIL row ("migration history lost") — covers both 5457 and 5458/5460 guards. No Docker; < 2 s.
- **VB7b (witness abort).** `vb_driver` gains a MODE `witness-abort`: the driver sets `OWNED_REG` to a `mk_reg`
  registry (3866-3873) made by the control, calls `res_witness_start` (869-893), then `exit 2` with the trap order
  `bg_stop_all; res_witness_abort; drv_cleanup`. Asserts: driver exit 2, `BACKGROUND FAIL res-witness stopped before
  join` row (as VB2's), W (`$WORK/res-witness/w.id`) gone by exact ID, `stopped` present, no `hold.d` latch, no
  labelled leftover (`vb_gone`). Plus a static assertion on the candidate: `run.sh:111`'s trap string names
  `res_witness_abort` after `bg_stop_all` and before `cleanup`. One busybox container + one `docker events` stream;
  estimated 5–15 s (unmeasured; VB2 is 6 s). Unknown until the first run: whether `res_witness_start`'s preconditions
  (`_res_d` registry layout 466-469, `_dro` reservation 644) hold inside the driver without run.sh's scope stack; if
  the witness cannot start there, the control records it as a `prob` and the fallback is `res_witness_abort` against a
  `docker_run_owned` container plus the static trap assertion (reviewer-visible, not silent).
Consequences of 528 → 529 controls: counts are computed (`${#CONTROLS[@]}`, 251/861/7253; `--list` prints the table),
`required-checks.tsv:40` keeps its single `B-CI-CONTROLS ci-controls` row, no catalog/parity file pins 528 (repo grep:
only `docs/foundation/OPERATIONS.md:753` "prints 528 controls" — update to 529 in the same commit; management prose is
historical). VB7 runs in `B-CI-CONTROLS` of `validate:integration` (all groups selected, as RC8's "528 selected"), not in
`validate:checks` (fast has no `ci-controls`). It is a hosted-profile regression, never a local gate.

### 11.3 Qualification (hosted only) and optional local diagnosis

**Gate (Spec r6; supersedes r5's selector sets as gates).** On the exact final MR head of !87 carrying every executable
repair and its docs: (1) `validate:checks` (automatic on push) must pass; (2) then the optional `validate:integration`
job is started manually **in that same MR pipeline**, and both jobs must pass on that same SHA — the full hosted profile
the path map selects for `infra/` (§11.1), and the only qualification of K1–K3 and AC9. `validate:integration` is **not**
run on the current head `db968e2b` (pipeline `2981988185`, fast-only, pre-repair). Both jobs have one named monitoring
owner (the coordinator) recorded with pipeline/job IDs and SHA; the first failure is diagnosed from its artifacts before
any further push. Any executable change after the hosted full run changes its input: the new candidate is requalified
(both jobs on the new SHA) before merge. After merge, the merged tree is compared with the tested tree: equal → no local
post-merge run; a management-only delta is recorded, not rerun. AC9
evidence from the hosted run: `results.tsv` (B-CI-CONTROLS PASS with VC1/VC4/VC8/VC9/VB2/VB5/V18n **and VB7** among the
529, B-PERSIST PASS, RESOURCES/BACKGROUND rows), `ci-controls-timings.tsv` (VB7 row), `inventory-leftover` empty, the
wrapper cleanup inventory; no FAIL/UNMET/held row. VB7 is the forced-start-failure proof of AC9 (K3 (a)–(c)); K1/K2 are
proved by VC1/VC4/VC8/VC9 on the hosted daemon. A hosted run that never reaches the repaired path (e.g. a fast-job lint
failure) is not evidence.

**Optional local diagnosis (no gate; cannot replace the hosted result).** The Spec r5 selector sets remain available to
the coordinator for diagnosis only: `validate-controls.sh --only VC1,VC4,VC8,VC9,VB2,VB5,V18n --keep-scratch` (run 4
timings 10+10+22+90+6+7+123 = 268 s + startup; qual2 overhead ≤ 20 s) — it also yields the exact K2 records in
`owned/VC8.tsv`/`owned/VC9.tsv` and the registry mirror if Codex wants them in the Result; the fast profile (9 min
measured) before pushing, to catch B-SHELL/B-LITERALS locally instead of on the hosted job; a stack-free K3(a) proof
(`persist` 5450-5466 extracted from the candidate `run.sh`, run under `set -euo pipefail` with stubbed `compose`/
`run_docs_block`/`record`/`log` and a failing `compose_migrate`: must return with one `B-PERSIST FAIL` row; seconds) and a
f0-style K3(b) witness proof (`res_witness_start` with a real registry, then `bg_stop_all; res_witness_abort`; W gone by
exact ID; 1–3 min, unmeasured). The proposal's `run.sh … ops` with a forced-failing README block is not planned (edited
non-candidate copy, stack + builds, 10–20 min unmeasured, proves nothing the hosted ops suite does not). The local
29.8.2 dind probe (proposal §3 R1 against `docker:29.8.2-dind@sha256:7dcdfc4a…` with `--storage-driver overlay2`, driven
from `docker:29.8.2-cli@sha256:b1805116…`) is **dropped as a gate**; it stays an optional diagnostic only if clearly cheap
— it needs two image pulls (unmeasured) and a privileged container (owner approval in-session; a relayed approval is not
one), and its only question (is an untagged labelled image hidden from `-q` but listed by `-aq` on 29.8.2/overlay2, m2)
does not change K1's design. Recommendation: do not run it before coding; the hosted `validate:integration` run on the
repaired SHA exercises the real daemon. What local runs cannot show on any path: hosted 29.8.2/overlay2 listing
behaviour, runner load, registry latency, the hosted wall clock.

**Prior G065 evidence after the K1–K3 edits.** Invalidated as qualification: the qual2 fast profile and the two
`--only` runs (all three edited files are their inputs/helpers); they are superseded by the hosted full profile, not
re-run locally. Still valid: the `_res_live` proof (`lib.sh` 517-560 untouched — confirm by diff hunk), the R8 cutover
receipt and absence listings; the rename's Codex source review for the rename hunks. Any optional local run records
cleanup: `--keep-scratch` scratch copied to `~/.local/state/platform-evidence/g065-a2/` then removed, exact IDs of every
proof object recorded before removal, inventory counts (15/12/6/52 baseline) before and after.

### 11.4 Estimate and budget feasibility (B3: 3 attempts, 1 spent, checkpoint after 2)

This repair is **attempt 2**; the checkpoint falls at its end. Attempt 1 stays spent.

| Part | Estimate | Basis |
|---|---|---|
| Coding K1 (≈20 lines: 5 helper sites + 6 caller branches) + K2 (`vc_record_tag` ≈8 lines, 2 VC8 + 1 VC9 call sites) + K3 (5457 record-and-return, 2 guards, 5352 early return, `res_witness_abort` ≈8 lines, 1 trap edit) + VB7 (≈40 lines: driver mode, two cases, OPERATIONS.md count), self-review, optional local fast + VB7 dry run | 2.5–4 h | line counts; VB7 is the largest unmeasured part |
| Hosted pipeline elapsed per push: `validate:checks` ≈ 10 min (measured on !82/!84); `validate:integration` **> 51 min** — RC8's 3,082 s is one data point of a run that failed inside ops with 346 controls not started and rename/off/final unrun, so a green full profile is longer; unmeasured, planning figure 60–90 min | 1.2–1.7 h per full hosted round | RC8 job 17419324226; !82/!84 |
| Review waits: Codex amendment review, source review of the repair hunks, Duo on the MR | 1–3 h | unmeasured |

Total ≈ 5–9 h for one clean round (VB7 adds about an hour of coding and 5–20 s of hosted time), dominated by the
hosted integration job and reviews. Budget rule (AGENTS.md
"Iteration budget"): a hosted job that fails on the repaired SHA is executed-and-failed evidence inside attempt 2, not
a separate increment; the correction round it forces (fix, push, `validate:checks`, re-trigger `validate:integration`)
is **attempt 3** and consumes the last of B3. So the remaining budget contains the repair plus exactly one hosted-found
correction loop (≈ 2–4 h more). It does not contain a second material correction, a mechanism outside K1–K3, or a hosted
failure in unrelated controls that needs its own repair. Stop for an owner decision when: attempt 2's hosted full
profile fails for a reason outside K1–K3; attempt 3 would be consumed by anything but one bounded correction; or the
checkpoint after attempt 2 shows the hosted job cannot be re-triggered on the same SHA (quota, runner) — a resource stop,
not an attempt.

### 11.5 Sequencing, delegation, escalation

The repair commits go on `g065-rename` after the rename commit (same MR !87, still draft until amendment acceptance,
repair, hosted `validate:checks` + manually triggered `validate:integration` on the final SHA, review disposition and
input/tree reconciliation are complete), after fetch/rebase reconciliation onto the current staging (K1–K3 touch only
`infra/`; management commits do not conflict). The rename's source review stays valid for the rename hunks; the repair
hunks get their own Codex source review. Order: P0 Codex review of this amendment → P1 code K1, K2, K3 (one Claude
implementation owner, Opus 5.5 high, stack-free writer in an isolated worktree; no subagent; the coordinator runs any
optional real-Docker diagnosis in its own test slot) → P2 safekeeping push (hosted `validate:checks` runs; monitored) →
P3 Result addendum (AC9 mapping with the VB7 PASS row and its `ci-controls-timings.tsv` entry, the K2 inventory rows if a
diagnostic `--keep-scratch` run was made, exact SHA/tree, optional-diagnosis outcomes if run, attempt ledger) → P4 Codex source
review → P5 manual `validate:integration` on the final SHA's MR pipeline, Duo disposition → merge → merged-tree
verification; cleanup receipt; final ACCEPT. Escalation stops: any change beyond K1–K3 (other listing sites, `cleanup()`,
`owned_cleanup`, `cleanup-owned.sh`, `res_witness_stop` semantics, ceilings), any change to cleanup/hold semantics other
than reporting a failure that was silent, anything touching the `DeadlineExceeded` (no retry, pin, timeout, Dockerfile or
Compose change), a replacement RC tag, any local full/528-control run presented as qualification, and a hosted
`validate:integration` failure outside K1–K3 (report with job ID, SHA and the failing row; no retry of unchanged source).

### 11.6 Correction record r3c (Codex review `134464a3…`, staging `492e5ca5`)

Changed paths of the repair (supersedes §5's "Changed:" list for §11 work): `infra/ci/validate-controls.sh` (K1, K2,
VB7), `infra/validation/lib.sh` (`res_witness_abort`), **`infra/validation/run.sh`** (K3: 5352, 5457-5460, 111),
`docs/foundation/OPERATIONS.md:753` (control count), `management/v0/results/G065-*-result.md`. Path mechanisms: the CI
path map `infra/ci/profile-paths.tsv:61` (`foundation dir infra/`) already covers all three scripts; the harness
change-scope check `B-SCOPE` (`run.sh` 6585-6594) matches every path changed since `--base` against the allow-file —
hosted and `validate.sh` runs use `--base HEAD` with `infra/ci/unchanged.allow` (`validate.sh:799`; 4 lines, committed
candidates change nothing), so no tracked allow-file changes; the §4 allow-file for a direct `run.sh --base <staging
SHA>` diagnosis gains `^infra/validation/run[.]sh$`. Spec r6 names the directories `infra/ci` and `infra/validation`
(line 264); the Goal names no file list. VB7 adds no path outside this inventory.

| Item | Change | Where |
|---|---|---|
| 1 K3 guard gap | all `persist` commands enumerated; record-and-return at 5457, guards at 5458 and 5460, early return in `variants_ops` 5352 (reviewer decision flagged) | §11.2 K3 (a)–(b) |
| 2 K2 fail closed | `vc_record_tag`: inspect failure, `gate_record` failure and a missing row each `prob` with the exact tag | §11.2 K2 |
| 3 AC9 proof in hosted profile | new group-D control VB7 (a: persist guard under strict mode; b: witness abort + static trap assertion), insertion points, cost, 528→529 consequences | §11.2 "K3 regression", §11.3 evidence, §11.4 estimate, §11.5 P3 |
| 4 K1 sentinel + inventory | `-1` sentinel, listing/count split, every caller's `< 0` branch, VC4:4319 kept; `run.sh` added to the inventory and the §4 allow-file; B-SCOPE/path-map citations | §11.2 K1, §11.6 |
| 5 budget header | current-binding note under the header, Goal/Spec citations, history preserved | header (after "Budget: proposed B3…") |
