> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G069 Result — unaltered brand kit adoption (ACCEPTED)

Status: Codex final ACCEPT on 2026-10-27. MR !108 merged to staging as `094403058c3efa395fd5d8e8cd961005fb2ca0ea`; the owner waived post-merge local qualification. RC16 and RC17 did not qualify and are not reported as passes. See the final acceptance and cleanup record below.

## Candidate

- Branch `g069` (MR !108 to `staging`), rebased onto staging `7fe7ba62703db25fdf8f4abd0c2c0c513696f3be` (verified ancestor). Final candidate SHA recorded at push (below).
- MR !108 remote head before this round: `f258532e` (rebased equivalent `dc8b1f2c`). Review delta for this round: `dc8b1f2c..<candidate>` (product, Core, harness and test-oracle commits listed under "This round").
- All commits authored and committed by <owner> <<owner-email>>.

## Owner decisions applied (this round)

1. Exact kit, overrides included (`G069-exact-kit-owner-decision-20261026.md`, `G069-owner-kit-fidelity-delta-20261026.md`): the app's colour correction rules, global reduced-motion rule, forced-colours outline rule, pointer-cursor rule and app colour classes on kit components are removed; `[data-blueprint]` and the `dark` custom variant are kept. Kit contrast shortfalls are OWNER-WAIVED.
2. "UI is perfect now. Ship it." (relayed by Codex): owner M1 manual check complete. M1 walkthrough source: web image built from g069 `10ebc01b` (pre-rebase SHA); Core and core-replica images rebuilt afterwards from `349ef6de` (pre-rebase). Image IDs in `~/.local/state/platform-evidence/m1-preview-20261024/g069-serve/{web,core}-update.log`. The code delta from `349ef6de` to the rebased equivalent is empty for web/core/infra/.agents; this round adds only test, harness, documentation and dead-module deletions on top of that product code.
3. Kit-baseline floor in Core `CheckContrast` (Codex-approved bounded design): for each pair entry and appearance, required = min(existing threshold, the same entry's ratio on `Effective(a, nil)`); same evaluator; no epsilon or persisted-data change. Core platform defaults equal the kit hex values.
4. Validation amendment (owner accepted Codex's recommendation, 2026-10-26): local qualification is limited to the focused Core branding/validator unit check plus the static rows already running; duplicate local browser and database batches are skipped. G069 is qualified through hosted fast plus a full hosted `validate:integration` on the final combined candidate (pre-merge, replacing the full post-merge run when the merged executable tree equals the tested tree). Focused post-merge checks and one `off` run per the Plan still apply. Local tests are used only to diagnose actual failures.
5. Tenant invalid-field waiver (owner, in-session, 2026-10-26): the kit's invalid-field input text (text-destructive on bg-input/50) on a tenant palette is OWNER-WAIVED from 3:1 up to 4.5:1 for axe color-contrast nodes on `input`/`textarea` elements with `aria-invalid="true"`. Anything lower, or any other element, stays a finding. Core holds no such pair. Seeded org A measures about 3.15:1 Light and 3.37:1 Dark.

## OWNER-WAIVED contrast (kit values as shipped)

White on #0075ff 4.21 (rest/focus), 80 % hover 3.18; N3 field boundary 1.26 Light / 1.48 Dark (1.32 hover); Light muted-foreground on muted band 4.40; Light destructive tint 3.99 (hover 3.32); Dark destructive 30 % tint 4.34 / 3.73; CTA S24 4.2097 rest / 3.7713 ×1.10 / 4.0165 ×0.90; destructive focus border (destructive/40) about 2.0–2.4; kit field fill with no boundary 1.12 Light / 1.16 Dark; Dark half-alpha invalid border/halo/fill (≈2.8/1.4/2.4); Light invalid-field value text 4.27. Core floor pairs: Light T4-on-muted 4.40, T8 4.21 / 3.18, T10 tints 3.99 / 3.32, N3 1.26; Dark T8 4.21, T10 30 % 4.34 / 3.73, N3 1.48 / 1.32. The tests assert each waived pair at the kit's own ratio (tolerance 0.03–0.05) and below its threshold, so a regression or a stale waiver fails. Each test also pins its list of waived rows.

Previously refused saves now accepted by the floor (documented in `TestSavedFixtureCompatibility`): seed pre-S23 Light/Dark, #696969 with and without white foreground, BRD-1 #1d4ed8 Light, #2f6fb8.

## This round (since `dc8b1f2c`)

- Product (owner-approved UI): kit Field/FieldSet/Alert/Card/Item/Badge consumers; app colour classes and remaining overrides dropped; kit Dialog close in the help dialog; the globals.css corrections, reduced-motion, forced-colours and pointer rules removed; the unused `field-classes.ts`, `button-classes.ts` and `dialog-close-button.tsx` deleted.
- Core: platform defaults equal the kit hex values; kit-baseline floor (`core/internal/branding/colors.go`); branding tests aligned (`branding_test.go`, `branding_integration_test.go`).
- Harness: B-WEB-TOKENS now requires no `:root,.light` or `.dark` correction rule and no colour literal in app.css (`infra/validation/run.sh`); the globals.css customization row (`render-customizations.tsv`) updated.
- Test oracles: shared browser support (`modes`, `form-controls`, `source-reference`, `spacing`, `contrast`), the token fixture, and `tokens`, `design`, `overlay`, `accessibility`, `account`, `enrollment`, `mcp`, `onboarding`, `signup`, `membership`, `labels` and `identity` specs, retyped to the kit as shipped, with OWNER-WAIVED helpers. No assertion intent dropped. Removed checks measured app rules that no longer exist (reduced-motion ≤0.01 s, solid input boundary, disabled-radio dimming, pointer cursor); each was re-pointed to the kit's actual behaviour.
- Delegation: four isolated Opus 5.5 xhigh test-correction writers (Fable weekly limit hit, HTTP 429; standing fallback Opus 5.5 xhigh for test corrections, actual setting xhigh). Two earlier Fable dispatches died at start without work. The coordinator integrated all work by cherry-pick after base, authorship and scope checks, and removed each worktree.

## Validation (actual outcomes)

- Static batch 10 (head `fcbcf7b0`, rows lint lint-base lint-web lint-core unit-web unit-core): all PASS except B-WEB-TOKENS FAIL (stale correction-rule pin) and B-SCOPE FAIL (a comment-only change to `audit_integration_test.go` outside the allow-list). B-UNIT-CORE PASS, including `core/internal/branding` and `core/internal/tenancy/access`: this is the focused Core branding/validator check. Evidence: `.local/evidence/g069-a1/static-10/`.
- Fixes: `8126bf3e` (scope revert), `1f95c20d` (B-WEB-TOKENS). Static 11 (`1f95c20d`, lint-web): B-WEB-TOKENS PASS (B-REQUIRED-CHECKS UNMET only for B-SECRET-CANARY, which needs the `lint` row).
- Static 12 (`1849cddf`, rows lint lint-base lint-web unit-web): FAIL B-WEB-DESIGN and B-WEB-ASSET-CONTROLS (DESIGN_LITERAL: a comment naming the kit stylesheet's file in `support/form-controls.ts` and `tokens.spec.ts`); FAIL B-UNIT-WEB (tsc: the two parallel writers' `waived` fields on `PairResult` and `KitPairResult` had conflicting types). Fixed in `f7fe9f8a` (modes.ts label renamed `kitWaiver`; comments reworded). Evidence `.local/evidence/g069-a1/static-12/`.
- Static 13 (`169b7619`, rows lint lint-base lint-web unit-web): run.sh rc=0, all rows PASS. Evidence `.local/evidence/g069-a1/static-13/`.
- Codex source review of `dc8b1f2c..169b7619`: REVISE (handoff abe8d840). (1) P2: the icon-only Appearance combobox exposed its value only through `data-appearance-value`. Fixed: a visually hidden label (Light, Dark or System) inside the trigger (`web/components/layout/appearance-control.tsx`), with regression assertions `appearance.spec.ts` `choose()` (toHaveText of the chosen label, label not aria-hidden) and `accessibility.spec.ts` (closed combobox value "System"). (2) P2: `appearance.spec.ts` and `membership.spec.ts` measured the 4px popup inset during the kit's 100 ms zoom-in-95 open motion. Fixed: both wait for the popup's animations to finish before measuring (the `design.spec.ts` pattern). Commit `747b2ef2`. Selectors: browser `appearance.spec.ts`, `accessibility.spec.ts`, `membership.spec.ts` (first executed by the hosted `validate:integration`); static rows below.
- Hosted: superseded by the correction round below (the full hosted run on `0a2a8a12` FAILED; see below).
- Browser oracles were desk-written without a local browser run, under amendment 4. The hosted `validate:integration` was their first execution and it failed on oracle defects; see the correction round.

## Source review

- Codex source review, handoff 2b430ccc: ACCEPT of `169b7619..043e5b82` at `043e5b82` for pushing/updating MR !108 only. Both P2 findings corrected; no remaining material issue; `git diff --check` clean; Codex ran no application tests. This is not source/final acceptance: hosted fast, Duo disposition, the full hosted pre-merge `validate:integration`, merge, focused post-merge checks, cleanup and final Goal ACCEPT remain.
- Previous MR !108 pipeline before this push: 2991500484 on `f258532e`, status success (validate:checks 17482658449 success; validate:integration 17482658450 manual, never played; superseded by this push).

## Duo disposition (pre-full-run)

Duo reviewed `8b11179d` within minutes of the push and left three recommendations, all stale documentation or comment text, no functional defect: (a) `docs/foundation/WEB-SHELL.md` destructive row still named `#bb0009`/`#ff9c9e`; (b) the same file called Core's alignment "a pending delta"; (c) a `branding-form.tsx` comment claimed `PendingSubmit` adds a destructive focus border. All three fixed in `8f5707e2` (comments and documentation only) before the full hosted run, so the tested tree is the merged tree. The earlier thread on draft markers in `core/internal/branding/colors.go` was confirmed fixed by Duo. Threads are left for Duo/reviewer resolution; no review state is altered by the coordinator.

## Pending gates (as of the pre-merge Result; superseded by the closing section)

Combined delta review (Codex) → push with lease → hosted fast + Duo disposition → full hosted `validate:integration` on the final SHA (no pushes during the run; coverage verified against shell/appearance/motion/token/portal/identity/branding) → staging merge under standing authority → focused post-merge checks including one `off` run → closeout receipt.

## Correction round, merge and RC16 (2026-10-26 to 2026-10-27 UTC)

### Hosted full run on `0a2a8a12`: FAILED (preserved)

MR !108 pipeline 2991916404 on `0a2a8a12e3cac0a3be73974a68548563cbd20787`: `validate:checks` job 17486065458 success (773 s). `validate:integration` job 17485979274 (played manually) ended **failed** after about 5942 s; the pipeline shows success only because the job is `allow_failure`. Failed rows: B-BROWSER-ROOT, B-IDENTITY-E2E, B-IDENTITY-PRIVACY, B-OBS-OFF, B-BROWSER-OFF, B-TOGGLE-OFF, B-SECRET-CANARY and the derived B-REQUIRED-CHECKS. The artifact is unpacked in the coordinator's private job directory and the run is not repeated. Evidence: `.local/evidence/g069-a1/` (`root-1`..`root-4`, `ops-1`..`ops-3`, `off-1`, `off-2`, `static-15`).

| Row | Cause (verified from the logs; every cause was a desk-written oracle or harness-list defect, no product defect) | Fix |
|---|---|---|
| B-BROWSER-ROOT | 16 root tests: (a) Base UI `Select.Icon` leaves a never-painted "▼" text node inside the aria-hidden chevron, so the combobox `textContent` was "System▼" while its rendered text is "System"; (b) the spacing check read textContent and expected ""; (c) S-FONTS measured the select item, not its ItemText child; (d) the dialog field autofocused, so its border was the focus ring; (e) WAIVED_TOLERANCE 0.05 was tighter than the 0.0715 observed; (f) a leftover pointer hover audited the hover fill (3.16) instead of the waived rest fill (4.2) | `2fb64d58` (was `c07f9038`) |
| B-IDENTITY-E2E | 18 failures (membership E-MEM-2,3,4,5,6,9,10 and S-SELECT-KIT; labels E-LAB-4): `chooseOption` locator `[data-slot=select-content]` matched two elements (a just-closed popup still in the DOM with `data-closed`, plus the newly opened one) | `71b78039` (was `d28d59d5`): `SELECT_POPUP` names the open popup; `chooseOption` waits for it to leave and for the close motion |
| B-IDENTITY-PRIVACY, B-SECRET-CANARY | derived: a registered value inside the failed tests' `error-context.md` | none needed (passed on rerun) |
| B-BROWSER-OFF, B-OBS-OFF, B-TOGGLE-OFF | the intelligence-disabled variant's web build failed `Can't resolve '@/lib/tenant-css'`: the new `web/lib/tenant-css.ts` was missing from run.sh `APPLICABLE` | `71b78039`: added to `APPLICABLE` |
| (found on rerun) | tenant-fixture S-FIRSTPAINT cases skipped in the project of the other appearance, violating the zero-skip browser rule (4 tests per project without a passing S-ERRORS) | `41ca881d` (was `4dabdbd3`): each case runs in both projects (it emulates its own OS appearance) |
| (found on rerun) | E-MEM-6 and S-SELECT-KIT: three trigger assertions read textContent ("Member▼") | `b4adfa65` (was `a1a01143`): `toHaveText(..., { useInnerText: true })` |

First-failure handling: root and identity reproduced locally on private stacks (`root`/`ops` suites, no repeated pushes); the second layer (membership oracle) only appeared once the first no longer masked it, which is why `ops-2` still failed 4 tests before `ops-3` passed. Parallel fix agents were not needed: each fix was a few lines.

### Focused local reruns on the corrected commits (executed, all PASS)

- `root-4` on `4dabdbd3`: RESULT PASSED, 20 required IDs each PASS once, including B-BROWSER-ROOT (no skips) and B-SECRET-CANARY, B-SCOPE.
- `off-2` on `a1a01143`: RESULT PASSED, 7 required IDs, including B-BROWSER-OFF (172 tests), B-OBS-OFF, B-TOGGLE-OFF, B-TOGGLE-OFF-ADDITIONS.
- `ops-3` on `a1a01143`: RESULT PASSED, 17 required IDs, including B-IDENTITY-E2E (174 tests), B-IDENTITY-PRIVACY, B-SECRET-CANARY.
- The runs on the pre-rebase commits cover the final tree: the pushed head `b4adfa65` differs from `a1a01143` only by staging's docs-only commit `9cbfb633` (README link line, OPERATIONS.md, new TEST-INDEX.md), verified by `git diff --stat` before the push.

### Qualification decision (owner, relayed by Codex: handoffs 95550e05 then 7f970617)

The failed full hosted run stays recorded and is not repeated on the feature MR. The earlier choice (full hosted pre-merge replacing the post-merge full run) was superseded: first by a full local post-merge qualification (95550e05), then by no local post-merge qualification, full or affected (7f970617). So no post-merge local run was performed for !108 and none is claimed. RC16 receives the full hosted RC qualification, which is the next broad gate. Local focused reruns above, hosted fast and Duo are the pre-merge proof; unexecuted checks are not reported as passing.

### Push, hosted fast, Duo, merge

- Rebased `g069` onto staging `9cbfb633` (80 commits; clean), pushed with lease from `0a2a8a12` to `b4adfa65a4c5d580b8cdf61657870861a75e9ed3` under the owner's in-session "You can push on your own".
- Hosted fast: pipeline 2992339779 on `b4adfa65`: `validate:checks` 17488809957 success (771 s); `validate:integration` 17488809958 left manual and unplayed by the owner/Codex decision above.
- Duo reviewed the new head: "nothing to comment", approved. Open discussion threads: 0 of 20 (checked from the discussions list, not the blocking flag).
- Merge: `glab mr merge 108 --sha b4adfa65… --remove-source-branch`; merge commit `094403058c3efa395fd5d8e8cd961005fb2ca0ea` (parents `ca04bb38` staging, then `b4adfa65`). Staging had advanced to `ca04bb38` (management/docs only: G071 and OS001 documents), so the merged tree is the tested tree plus that management-only delta. Source branch `g069` deleted by the merge.

### RC16 (owner: cut RC16 only after !108 merges; full hosted qualification not waived)

Annotated tag `0.2.0-rc.16`, object `bddef120ab08d7cbfee2adf2ed94319602bb93fb`, peeled `094403058c3efa395fd5d8e8cd961005fb2ca0ea`, tree `6fa024ecdb7fb6842150614b6bee7da64151813f`; release helper `name` and `candidate`: PASS. Promotion MR !109 (`release/0.2.0-rc.16` to `main`, squash off, source deletion on), pipeline 2992370523, jobs `validate:checks` 17489025566 and `validate:integration` 17489025567 (automatic, required). The MR was opened with auto-merge off; GitLab then recorded "enabled an automatic merge when all merge checks for 094403058c3efa395fd5d8e8cd961005fb2ca0ea pass" at 2026-10-27T01:30:42Z (`merge_when_pipeline_succeeds=true`), before the closing commit of this Result. Auto-merge is therefore armed, with the required full hosted qualification still pending: no PASS, no main promotion and no stable tag is claimed. Receipt inputs: `.local/evidence/rc020rc16/`.

### RC16 state after the Result write (observed 2026-10-27)

The owner closed promotion MR !109 (`closed_by` `<owner-account>`) at 2026-10-27T01:36:28.448Z (`merged_at` null, `merge_when_pipeline_succeeds` false); the armed auto-merge had been canceled at 01:35:45.262Z. GitLab then ended pipeline 2992370523 as **canceled**: `validate:checks` 17489025566 canceled after 465.105 s (started 01:30:07Z, finished 01:37:52Z), `validate:integration` 17489025567 skipped (never started). The cancellation followed the owner closing the MR; it is neither a pass nor a test failure. No required job completed, so RC16 has **no qualification**, and nothing was promoted to main. The MR was not reopened, auto-merge not re-armed and nothing re-run. Tag `0.2.0-rc.16` (`bddef120…`, peeled `094403058c3efa395fd5d8e8cd961005fb2ca0ea`) remains unchanged and immutable. At the owner's direction the remote branch `release/0.2.0-rc.16` was deleted; `git ls-remote` shows it absent, while the tag and commit `09440305` remain.

### Replacement RC17 (state at the time of this amendment, 2026-10-27)

At the owner's direction a replacement candidate captures the G069 Result: annotated tag `0.2.0-rc.17` (object `8277e64c45b41012feef7eae1f8f33c29fc63ca3`, peeled staging `42aa398055c5049d196c26e60dfa55dcc0e3b85d`, tree `ba3227c8d38d8734dfa1308295e7ff0a273c6d31`), branch `release/0.2.0-rc.17` at the same commit. The executable content equals RC16's; only management records differ. Promotion MR !110 targets `main` (squash off, auto-merge off, manual promotion only). Pipeline 2992392293 was running when observed: `validate:checks` 17489187669 running, required `validate:integration` 17489187670 created and not started. No pass and no promotion is claimed; the outcome is recorded only after the pipeline reaches a terminal state and is verified. Staging commits after the cut (management-only) do not restart the frozen candidate or its pipeline. Receipt inputs: `.local/evidence/rc020rc17/`.

### Closeout state

Not yet cleaned (pending Codex final ACCEPT, which removes them): preview worktree `g069-preview-20261026`, lane branches `g069-c-w4`, `g069-d-f1`, `g069-d-f2`, `g069-d-f3`, local `g069` and `g069-closeout`, my task-owned `g069*` Docker images and the `g069-a1-nm-r1` volume, and the evidence directories above. Retained by name: M1 preview worktree and volumes (G057 frozen inputs). Final cleanup receipt follows at ACCEPT.

### Closeout reconciliation (2026-10-27, append-only; supersedes the "running" and "pending" wording above)

- **RC17 final state.** Pipeline 2992392293 of promotion MR !110 finished with `validate:checks` (17489187669) **success**
  and `validate:integration` (17489187670) **failed** after 4078 s (script_failure). Per the owner's instruction Codex
  closed !110. RC17 (`0.2.0-rc.17`, peeled `42aa3980…`) is **closed and unqualified**; it is not reported as a pass and
  is not reopened in this task. The tag object stays as published. Hosted artifacts and the failed-run analysis are
  preserved in `.local/evidence/rc020rc17/hosted-fail-2992392293/`.
- **Cause of the failure (not G069 content).** G064 resource accounting latched `B-MIG` (`ceiling network 2/1`, compose
  up) when the background ci-controls child's network overlapped B-MIG's stack, and the latch refused every later
  allocation, which failed 31 checks. A local run of the `migrations` suite on the same SHA produced 52 PASS and no
  latch (`.local/evidence/rc020rc17/local-migrations-1/`; its only FAIL is the aggregate `B-REQUIRED-CHECKS` for the
  unselected `B-IT-ISOLATION`). The repair is staging MR !111 (merge `acc3f7d4`, G064 corr8, control VB8). RC18 follows
  G071.
- **Cleanup state verified (read-only, 2026-10-27).** Still owned by this task and not yet removed, to be removed at
  the final ACCEPT under the mandatory cleanup policy (nothing here holds unique work; the merged source is
  `094403058c3e…`): preview worktree `g069-preview-20261026` (detached `349ef6de`); 21 local `g069*` branches (lane
  branches `g069-c-*`, `g069-d-f1/f2/f3`, `g069-w*`, plans, `g069-closeout`, `g069-closeout-r2`, `g069`, diagnostic
  branches) and the 6 remote `g069-closeout`, `g069-closeout-r2`, `g069-plan`, `g069-plan-r1`, `g069-w5c`,
  `g069-w5c2`; 80 task-owned `g069*` Docker images; volume `g069-a1-nm-r1`; evidence directory `.local/evidence/g069-a1/`
  (about 100 MB). Zero `g069*` containers and networks. Retained by name: the M1 preview worktree and volumes (G057
  frozen inputs), the RC16/RC17 evidence directories, and `g069-closeout-r3` until Codex imports this record. A
  cleanup receipt with verified absence follows at the final ACCEPT.

#### Correction to the closeout reconciliation above (append-only, Codex review of `992f876f`)

- **Local `migrations` run.** `B-IT-ISOLATION` was selected and required in that run, not unselected. It is UNMET
  because T-9 (residue) and T-10 (runner removed during an exec) were not exercised: the runner was unavailable
  (`no-registry`; eligible invocations took the per-invocation container path), while T-1…T-2 and the other cases passed.
  The only FAIL row of that run is the aggregate `B-REQUIRED-CHECKS` naming `B-IT-ISOLATION(1,UNMET)`; it is not a product
  failure and the run still shows no resource latch (52 PASS).
- **Hosted RC17 failure count.** The hosted evidence has 37 FAIL rows: 32 nonaggregate product-check failures, plus
  `B-IT-ISOLATION`, the aggregate `B-REQUIRED-CHECKS` and three `RESOURCES` failure rows (`B-MIG: ceiling network 2/1`,
  `B-MIG container: latched`, `B-MIG network: observed_max>1`); `B-MAINTENANCE-DRILL` was also UNMET. "31 checks" above is
  superseded by these counts. The RC17 disposition is unchanged: failed, closed, unqualified.
- **Cleanup.** The final ACCEPT stays blocked until the resources listed above are actually removed and a verified
  receipt is recorded; this record only inventories them.

#### G069 cleanup receipt (2026-10-27, append-only; verified after removal)

Executed by the coordinator for the resources it created; unique unmerged history was preserved first.
- **Preserved.** One verified git bundle of the 20 unmerged local `g069*` lane/plan/closeout/diagnostic branches plus the
  detached preview head `349ef6de` (21 refs; 26 MB; `git bundle verify` OK; sha256
  `2d34803e343d4c8784da5996a630b356cb006265426144015951e9454084b766`) in the private
  `.local/evidence/g069-archive/g069-local-branches-20261027.bundle`.
- **Removed and verified absent.** Preview worktree `g069-preview-20261026`; the local `g069` branch (merged) and the 21
  bundled refs (local `g069*` branches left: only `g069-closeout-r3`, pending Codex import); 80 `g069*` Docker images;
  volume `g069-a1-nm-r1`. Afterwards `g069*` images 0, volumes 0, containers 0, networks 0.
- **Not removed (named retention, release condition in brackets).** The six remote branches `g069-closeout`,
  `g069-closeout-r2`, `g069-plan`, `g069-plan-r1`, `g069-w5c`, `g069-w5c2` [until Codex confirms their records are
  imported, then deleted on request]; `g069-closeout-r3` local and remote [until imported]; evidence directory
  `.local/evidence/g069-a1/` (about 100 MB, contains the hosted-failure, loop and p0x run proofs) and the bundle
  [until Codex's final acceptance disposition]; the M1 preview worktree and volumes (G057 frozen inputs); RC16/RC17
  evidence directories.

#### Codex closeout review and cleanup status (2026-10-27)

- Read-only Result review of `2bd81135` found no material findings. The verified
  bundle (`.local/evidence/g069-archive/g069-local-branches-20261027.bundle`,
  SHA-256 `2d34803e343d4c8784da5996a630b356cb006265426144015951e9454084b766`)
  contains all 21 recorded historical refs and verifies successfully. The six
  retained remote tips matched bundle heads; all six and `g069-closeout-r3` were
  deleted after the receipt was imported to staging. No local G069 branches remain.
- Verified absent: G069 worktree, containers, networks, volumes and tagged
  images. `.local/tmp/g069-a1` and the task-scoped browser credentials under
  `.local/secrets/g069-a1` were removed. RC16/RC17 release evidence is retained
  separately.
- Raw local evidence cleanup is incomplete. An unprivileged removal cleared
  part of `.local/evidence/g069-a1`, leaving about 89 MB; 341 directories remain
  owned by UID 1001 and are not writable by the current user. `sudo -n` returned
  “a password is required,” so cleanup stopped without bypassing that boundary.
  GitLab still holds the failed G069 hosted artifact for job `17485979274`
  (1,128,854 bytes; expires 2027-01-25), preserving the required failure proof.
  The 26 MB verified bundle remains as the named archive of unique branch
  history; it is not a live worktree or an untracked branch.
- **Final Goal ACCEPT is withheld** until `.local/evidence/g069-a1` is removed
  through an authorized filesystem operation and absence is verified. G071's
  integration/final-acceptance dependency remains in effect; its current hosted
  qualification may continue without merging. This interim hold is superseded
  by the final acceptance record below.

#### Codex final Goal ACCEPT and verified cleanup (2026-10-27)

Codex ACCEPTs G069. The owner-directed qualification disposition is followed:
MR !108's hosted fast checks passed, Duo approved with no open discussions, the
focused corrected local browser/ops/off reruns passed, and the selected focused
post-merge row passed. The owner waived a full post-merge local run and directed
the RC path. RC17's full hosted run remains failed and unqualified because of
the separately repaired G064 B-MIG resource-accounting latch; it is not treated
as G069 qualification or as a pass. The earlier failed G069 hosted run and its
correction history remain recorded above.

After the coordinator's owner-authorized cleanup, Codex independently verified
`.local/evidence/g069-a1`, `.local/tmp/g069-a1` and `.local/secrets/g069-a1` are
absent; no local or remote G069 refs, G069 worktree, containers, networks,
volumes or images remain. The verified 21-ref Git bundle remains at
`.local/evidence/g069-archive/g069-local-branches-20261027.bundle` (SHA-256
`2d34803e343d4c8784da5996a630b356cb006265426144015951e9454084b766`) to preserve
unique historical branch commits. Owner: Codex; reason: preserve that deleted
history; release condition: delete once the unique history is imported into
durable repository records or the owner directs disposal. RC16/RC17 release
evidence remains separately retained.

Final ACCEPT releases G071's predecessor integration hold. G071's own hosted
qualification and review gates remain in force. The current Progress index
retains the RC16/RC17 timeline because it records the owner-directed RC18
qualification waiver; no dated history was removed in this closeout.
