> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G003 — Health, logging, contracts and baseline CI implementation plan

Goal: [G003-foundation-observability-ci.md](../goals/G003-foundation-observability-ci.md)
Spec: [G003-foundation-observability-ci-spec.md](../specs/G003-foundation-observability-ci-spec.md)
Owner: Claude Code · Reviewer: Codex · Budget class: **B7**: at most 7 shared meaningful cycles; mandatory checkpoint after cycle 4 and before cycle 5. Planning consumed no cycle.

| Revision | Date | Change |
|---|---|---|
| r1 | 2026-10-05 | Initial Plan. Submitted for Codex review. **Not approved; no implementation has begun.** One unmet prerequisite blocks execution (the GitLab runner profile, §12 and Q1); the remaining questions are in §22. — *Appended at r2:* r1 SHA-256 `a43aa523624bab98c38e0fe6c7ff86cffef2cd70d7a3da7b6e806d5c60adbdfc` (1333 lines). Codex review: **CORRECTION LOOP**, six corrections (§0.1). Codex recorded that Q2 and Q9 were also prerequisites, so the "one unmet prerequisite" statement above was incomplete; r2 corrects it. The complete r1 text is preserved verbatim in Appendix R1. |
| r2 | 2026-10-05 | Addresses the six r1 corrections (map in §0.1). **(1)** Runner facts, hosted authorizations and the fresh binding are recorded as unresolved blocking prerequisites (Q1, Q2, Q9), with an exact owner fact list (§12.2), named authorizations (§12.3) and a runner-qualification step R1 that runs the unchanged baseline harness cold on the runner **before S1** (§12.4). The working tree is in an owner-authorized skill consolidation; baseline verification is **PENDING** and the S0 gate is unchanged, with no new exception (§2.1, §15.1a, Q14). **(2)** Checkout attribution uses only documented variables per pipeline type, with 5 positive and 16 negative metadata controls that run before, and independently of, the harness (§11.2, §11.3). **(3)** An explicit phase and finalizer model for `validate.sh`: postcheck, owned cleanup, raw and staged scans and publication run on every exit path; the harness status always survives; scanner and report errors fail closed; a job-unique registry, discoverable by `after_script`, binds owned resources to label-based removal commands (§11.1–§11.4). **(4)** Schema-negative cases run through a separate audit path with declared failure classes and single-mutation derivations; normal runs still reject any unpinned bytes (§8.1a–§8.3). **(5)** Error output is established from the actual frameworks: Next.js 16.3.8 logs a thrown route error with message and stack, so the web wrapper now answers 500 itself; `core/cmd/server/main.go` gets a bounded JSON exit record and a classifying `net/http` error log; intelligence gains bounded startup-failure, uvicorn, library and warning records; three real-artifact checks assert this (§5.3–§5.6, §7, §10). **(6)** An executable step order: the scope allow-list is the first write (S1); CI files, documentation and harness changes are integrated in one step before their first static run (S6); exact intermediate commands are given (§15.3, §15.6). B7 is unchanged. Submitted for Codex review. **Not approved; no implementation has begun. Execution is blocked by Q1, Q2, Q9 and Q14 (§22).** — *Appended at r3:* r2 SHA-256 `6e6356f3bf7e6b0155fe4a32764e1e89d5cf25afa7a96dd0642096fd0220bc82` (3279 lines). Codex review: **CORRECTION LOOP**, five corrections (§0.4). Codex accepted the attribution, schema-negative isolation and step-order work, and accepted the Q15–Q19 directions subject to the corrections. The complete r2 text is preserved in Appendix R2 (body) plus the unchanged Appendix R1 section; the Appendix R2 header gives the hash-verifiable extraction. |
| r3 | 2026-10-05 | Addresses the five r2 corrections (map in §0.4).<br>**(1) State hand-off, signals and location.** Steps that create state (location, harness launch and supervision) run in the parent. Isolated phases return values only through a validated state file. The harness runs as its own `setsid` process group, under a parent-owned watchdog deadline (GNU `timeout` is dropped). Signals stop the whole group with a bounded grace. Containment is checked before any directory is created. A real-signal model passes 24 of 24 cases (§15.5a).<br>**(2) Finalizer consistency.** Publication has exactly three outcomes (full, minimal, none), and staging is all-or-nothing. The artifact summary is labelled pre-publication; the job's exit status and final log line are authoritative. The job-log status columns come from a validated, scanned file that outlives the scrub. `harness-console.log` is now genuinely excluded. The aggregation has a structural no-masking invariant, which fixes the `stage` omission Codex found in r2's model.<br>**(3) Positive ownership.** Networks, volumes and images are removed only with positive ownership: capture rows taken from owned containers, or the run's own build records. Unrelated post-inventory resources of every kind survive (V20c, V20n, V20v, V20i).<br>**(4) Error output.** `main.go` binds before logging and logs only the bound port and `app_env` (exactly three reviewed hunks). CE2 asserts exactly one `core.exit` record; a new CE3 asserts the bounded startup record. Web `error_type` comes from code-controlled `instanceof` categories, with a syntactically valid canary-bearing `name` control (WE3a).<br>**(5) Baseline and prerequisites.** The skill consolidation is now complete but uncommitted; route A is still recommended. Q1 and Q2 remain unresolved and must be supplied before execution approval is sought.<br>B7 unchanged. Submitted for Codex review. **Not approved; no implementation has begun. Execution is blocked by Q1, Q2, Q9 and Q14 (§22).** — *Appended at r4:* r3 SHA-256 `b100cb5266f70a90332afa16e3cffa70391eeed2e43fe170b9326eb7e00f54b2` (5375 lines). Codex review: **CORRECTION LOOP** with three items (§0.6). Two are Plan and evidence corrections; the third restates the unresolved execution prerequisites. The complete r3 text is preserved in Appendix R3 (body) plus the unchanged Appendix R2 and R1 sections. |
| r4 | 2026-10-05 | Addresses the r3 review (map in §0.6).<br>**(1) Parent-error provenance.** `signal`, `internal` and `harness_stop` are recorded in the parent before aggregation. `aggregate` runs in the parent and no longer mutates state, which fixes the lost `internal=19` (r3 F15). V18f now requires `harness=143` **and** `internal=19` in the authoritative final line. The new V18k covers a parent error after a successful harness (exit 19, `harness=0`, `internal=19`). External signals are recorded as `signal=`, never as internal errors.<br>**(2a) Standalone ownership.** Standalone `docker run` containers are bound positively at creation through a `--cidfile` in the run's registry (`docker_run_owned`), separate from the Compose `working_dir` rule. Labels only locate candidates for the report. Added controls: interrupted-standalone positives V27, V27b and V28; label-only negatives V29 and V30; ephemeral-container controls V31 and V31b.<br>**(2b) Registry retention.** F10 no longer deletes a local `$EV` that holds a retained registry. On failed cleanup, the registry, start inventory, cidfiles and cleanup inventory survive in both layouts, while the secret list is scrubbed. Never an empty claim (V32–V35).<br>**(3)** The prerequisites are unchanged and unresolved (§12.6, §22). Not repeated beyond the handback.<br>Model: **29 of 29**, with the strengthened assertions; the r4 driver fails the r3 model on F15 and F15b. 80 V-controls. B7 unchanged. **Not approved; no implementation has begun.** |

## 0. Review record and planning-session disclosures

### 0.1 r1 → Codex: CORRECTION LOOP (2026-10-05)

Codex reviewed r1 (SHA-256 `a43aa523…dbfc`) at HEAD `9461c4bdbd8c1ed3e9a9986c70647cd1add57101`. Codex confirmed that the embedded S0 script extracts byte-exactly to `0409148a…6723` and passes `bash -n`. That was script-text verification, not execution approval. The review is recorded in PROGRESS.md under `### 2026-10-05 — G003 Plan r1 review: CORRECTION LOOP`. **The review text governs.** The left column below abbreviates it.

| # | Codex correction (abbreviated) | Where r2 addresses it |
|---|---|---|
| 1 | **Runner, hosted authorization and current starting state** (R6/R7, AC7/AC10). Q1, Q2 and Q9 remain prerequisites. Record the runner's identity and tag, executor, isolation and permissions, host, path and network profile, capacity and timeouts, and artifact access. Obtain explicit authorization for the named hosted success and failure actions and for evidence retrieval. Check runner compatibility and cold-build feasibility early enough for the B7 forecast, not first at S12. Reassess or isolate the execution baseline (untracked `vestive-shadcn/**`) without deleting that work or adding a broad unrelated-path exception. A fresh supervisor binding follows approval and a resolved baseline. | §2.1 (current state; baseline **PENDING**), §12.2 (owner fact list, all **UNRESOLVED**), §12.3 (named authorizations A-R1 to A-E), §12.4 (R1 runner qualification before S1, with a decision rule), §15.1a (S0 unchanged; no new exception), §15.3 (S0R), §19 (forecast conditional on R1), §22 (Q1, Q2, Q9, Q14) |
| 2 | **GitLab checkout attribution** (R6/R7, AC7). `CI_MERGE_REQUEST_SOURCE_BRANCH_SHA` is empty in detached merge-request pipelines and set only in merged-results pipelines. Use the variables actually present for each pipeline type. Distinguish detached source, merged result and branch push, without inventing a merged-result claim. Add positive controls for each metadata profile and negative controls for missing or contradictory metadata, producing the intended preflight outcome independently of harness failures. | §3 (F13), §11.2 (attribution rules, cited), §11.3 group B (P1–P5, P3F, N1–N16), §15.5 (P-ATTR prototype: 21 of 21 as expected) |
| 3 | **Aggregate finalization, scanning and owned cleanup** (R7, AC7/AC8). Specify an explicit failure and finalization model: postcheck and safe cleanup reporting run on ordinary failure paths; a nonzero harness status survives; scanner and tool errors and missing or malformed reports fail closed. Resolve the raw-versus-staged scan mismatch. The cleanup registry must be job-unique, discoverable by `after_script`, and bound to the configuration needed for removal (`docker compose -p <name> down` has neither the Compose file nor the environment). Refuse unowned resources and test it. Keep cleanup failures observable. `after_script` runs in a separate shell, cannot fail a successful job and does not run on job timeout by default. | §11.1 (job timeouts and `after_script`), §11.2 (phases, finalizer, exit codes, raw and staged scans), §11.4 (registry and `cleanup-owned.sh`), §11.3 groups C–E, §10 (proposed B-ARTIFACT-SCAN withdrawn; registry hooks), §15.5 (P-FIN, P-TERM and P-SCAN probes) |
| 4 | **Schema-negative isolation** (R5/R7, AC5/AC8). Normal runs must reject bytes that differ from the two pinned hashes. SN02–SN07 must reach the intended validator checks through a disposable negative path, so a shared digest failure cannot make them pass vacuously. Each language asserts the intended classification. Keep an ordinary-run digest-mismatch control and the EI39 control. No network access and no production schema mutation. | §8.1a, §8.2a (declared classes and single-mutation derivations), §8.3 (stage order; normal mode versus audit mode), §10 (B-CONTRACT-NEGATIVE) |
| 5 | **Actual error-output conformance** (R4, AC4). The web wrapper rethrew, and r1 assumed Next prints only startup text. Prove the handling with a real-framework error probe that uses message and stack canaries. Core's protected `cmd/server/main.go` prints `run()` errors with `fmt.Fprintln(os.Stderr, err)`; resolve it, and if `main.go` changes, reconcile the register, scope list, disabled variant and equality guard under R8. Error events need bounded categories or types; arbitrary messages under a renamed `event` key are not enough. | §3 (F15–F19), §5.3, §5.4, §5.6, §7.1–§7.3, §10 (B-CORE-, B-WEB- and B-INTEL-ERROR-OUTPUT), §15.5 (P-GO probe and P-NEXT source evidence), §17 (`main.go` reconciliation), Q15, Q16 |
| 6 | **Executable step order** (R6/R8, AC6/AC9/AC10). S6 ran static checks that read CI files created at S7, and the scope allow-list was created only at S8, although earlier wrapper runs needed it. Reorder so every required file exists when it is first checked; name the exact intermediate commands and scope binding; keep the rule that a shared harness change reruns the affected static and control checks in the same cycle. | §15.3 (new sequence, with an r1→r2 step map), §15.6 (exact commands), §14.3 (the allow-list is created at S1) |

**Planning decisions Codex recorded at r1 (not re-opened by r2):**
- **Q3:** the bounded Python format and anchor corrections enforce the unchanged canonical schemas; they authorize no new schema restriction.
- **Q4:** keep the nested 2 s deadlines, prove optional intelligence is non-gating at core, and disclose the web timeout result.
- **Q5:** scoped to the credentials each service actually holds; web gains no credential access.
- **Q6, Q11, Q12:** the fixture location, the separate pinned test-tool module and the disposable local CI-parity checkout are within scope; no new vendor ADR is needed for the validation library.
- **Q7:** developer-only artifacts with a 90-day expiry are an acceptable intended policy, subject to verifying actual instance and project access and limits (§12.2, F-R11).
- **Q8:** failure of the required job is G003's gate; owner-managed merge and protection settings are not changed.
- **Q13:** the single exact ignored client-settings exception is acceptable, without reading, deleting or broadening it.
- **Budget:** B7 retained as the Goal's ceiling, at most seven shared meaningful cycles, checkpoint after cycle 4; zero cycles consumed. The six-cycle forecast stays conditional on runner compatibility, cold-run capacity and hosted authorization.

### 0.2 r2 planning-session disclosures

- **Delegation.** This revision was prepared by a planning subagent at the product owner's direction, in parallel with a separate owner-authorized skill consolidation run by another agent. That delegation authorizes Plan revision only. It is **not** implementation authorization and **not** hosted-action authorization.
- **Repository writes.** The only repository write is this file.
  - It was revised in place in this checkout, not in a worktree.
  - To keep the r1 text byte-exact (Appendix R1), the r2 file was assembled from a hash-verified scratch copy of r1 and new sections under `/tmp/g003-plan-r2/build/`. It was installed onto this path with one scripted copy, after re-checking that the path still held the r1 bytes. This is the scripted in-place edit the owner authorized.
- **Git.** Only read-only commands were used (`status`, `log`, `rev-parse`, `ls-files`, `diff`), always with `GIT_OPTIONAL_LOCKS=0`. No index, ref, stash, commit or push operation was performed.
- **Not done.** No install, initialization, application build, `docker build`, `docker compose up/build/pull`, harness run, GitLab API call, pipeline trigger or external configuration change.
- **What was run** (scratch only, under `/tmp/g003-plan-r2/`; full list in §15.5):
  - a stdlib-only Go probe (`go run`) inside the existing local `golang:1.27.1-alpine` image (`--pull never --network none`);
  - read-only `grep` of that image's Go 1.27.1 `net/http` source;
  - Bash probes of `errexit`, traps, `timeout` and finalization;
  - Git fixture probes of shallow merge-commit parents and the attribution classifier;
  - a gitleaks v8.30.1 exit-code probe on scratch directories (`--pull never --network none`);
  - `docker image ls` and `docker ps -a` inventory listings.

  The probes created only short-lived `--rm` containers, from images already present. They created no volumes, networks or images.
- **Documentation research.** Official GitLab documentation (predefined variables, merge-request and merged-results pipelines, `CI_PIPELINE_SOURCE` values, `after_script`, runner timeouts). Next.js 16.3.8 and uvicorn 0.54.0 source files at their release tags on GitHub.
- **The other agent's skill files were not inspected as inputs.** They were listed by path, and their bytes were hashed only for the B-WEB-DESIGN identity check (§2.1). Their content was not inspected, modified or bound.
- **No credential, token or environment secret was read or printed.**

### 0.3 r1 planning-session disclosures (verbatim from r1 §0, as submitted)

No Codex review has taken place yet.

Disclosures from this planning session:

- **Writes.** The only repository write is this file.
  - Scratch writes were made only outside the repository, under the job's temporary directory: the S0 gate script, its demonstration clones and their evidence (§15.1–§15.2), and a staging copy of this Plan.
  - The background session's write guard required this Plan to be staged outside the repository first. It is placed at its canonical path only as the single repository write.
- **One Git call with optional locks enabled.** The first baseline check of the session ran an ordinary `git status --porcelain=v1 -uall`. It printed nothing, but because optional locks were enabled it may have refreshed `.git/index` stat data. It cannot have changed content. Every later Git inspection used `git --no-optional-locks` with `GIT_OPTIONAL_LOCKS=0`.
- **Tool-local ignored file.** During the session, at 17:22 local time, the Claude Code client created `.claude/settings.local.json`. It is ignored by the user's global excludes (`~/.config/git/ignore`), and was not created, read or modified by me. It is not a repository input. §15.1 treats it as the single exact ignored-path exception.
- **What I did not do.** I did not install anything, initialize any tool, build any application image, commit, push, trigger a pipeline or change any external configuration.
- **What I did run.** Docker was used only for read-only tooling on scratch files and inventories:
  - `mikefarah/yq` reading the committed manifest;
  - `koalaman/shellcheck` on the scratch S0 gate;
  - inventory and port queries.
- **Remote queries (read-only).**
  - `git ls-remote --symref origin` over SSH, with batch mode on.
  - One unauthenticated GitLab API probe. It returned 404, because the project is private.
- **No credential, token or environment secret was read or printed.**

### 0.4 r2 → Codex: CORRECTION LOOP (2026-10-05)

Codex reviewed r2 (SHA-256 `6e6356f3…bc82`) at HEAD `9461c4b…`. Codex independently verified the unchanged S0 script (`0409148a…6723`) and the verbatim r1 appendix (`a43aa523…dbfc`). The review is recorded in PROGRESS.md under `### 2026-10-05 — G003 Plan r2 review: CORRECTION LOOP`. **The review text governs.** The left column below abbreviates it.

**Resolved by Codex (preserved unchanged in r3):**
- detached-MR attribution, the 21-case prototype and the raw parent check;
- schema-negative audit path, declared classes, single-mutation derivation and positive controls, with SD01/SD02 normal-mode digest rejection;
- S1 creating the allow-list, and S6 creating the CI, docs and harness inputs before the static and provenance gates;
- the unchanged S0 rejecting unrelated uncommitted changes;
- hosted evidence remaining inside G003.

**Directions accepted subject to the corrections:** Q15 (own empty-body 500), Q16 (bounded `main.go` customization), the real production-framework error controls, Q17, Q18 and Q19.

| # | Codex correction (abbreviated) | Where r3 addresses it |
|---|---|---|
| 1 | **Phase state, signals and the read-only location gate** (R7, AC7/AC8). Subshell assignments do not propagate, so the paths from P0, the budget and precheck from P1, and the background PID from P2 never reach the parent. Define behaviour when P0 has no valid location. The default `mktemp -d` creates a directory before containment is checked. Provide disposable evidence of the actual state hand-off and background mechanism: externally signal the supervisor, prove the harness child and process group stop, and prove postcheck and cleanup still run with the intended status. | §11.2.1 (parent-held state; isolated phases plus a validated state import; parent-launched `setsid` harness; watchdog deadline; supervisor signal handling; group readiness), §11.2.2 (L0 location in the parent, containment before creation; pre-location mode), §3 (F24–F27), §15.5a (P-STATE, P-SETSID, P-TEE, P-IMPORT, P-MODEL F15–F22), §11.3 group D (V18–V18j) |
| 2 | **Finalizer, publication and control consistency** (R7, AC7/AC8). F8 gated publication on scans alone, so a partial F5 could be published. F9 deleted evidence before F10 read `results.tsv`. F6 wrote summaries before F7–F9 while §11.2.5 claimed they listed every failure. V5b's "excluded" file was on the staging list. The supplied model omitted `stage` from aggregation; Codex reproduced `stage=19`, a published report and **exit 0**. Correct the model, rerun the faults, and add staging-after-partial-copy, summary, scrub and pre-location failures. | §11.2.3 (all-or-nothing staging; three publication outcomes; status columns retained and scanned; scrub last), §11.2.5 (pre-publication summary versus the authoritative final status; structural no-masking invariant), §11.3 groups C and D (V5b clean and planted; V16, V16b, V17, V17b, V18g–V18i), §15.5a (P-MODEL: 24 of 24 cases, including F11 staging after a partial copy, F12 summary, F14 scrub, F20–F22 pre-location) |
| 3 | **Owned cleanup across resource kinds** (R7, AC7/AC8). The network, volume and image rules relied only on a project label or name plus absence from the start inventory, so an unrelated resource created later under the same label could be removed. Specify positive ownership; reconcile the `<project>-<service>` image fallback with the ban on name-pattern-only removal; split V20 into container, network, volume and image cases, each surviving. | §11.4 (positive ownership rules, capture points, image build and capture records, `unproven` status), §10 (`run.sh` capture hooks), §11.3 group E (V19, V20c, V20n, V20v, V20i, V21–V26), §15.5a (P-OWN decision model) |
| 4 | **Remaining error-output paths** (R4, AC4). CE2 cannot pass under the two-hunk rule: `main.go` logs `core listening` with the raw `cfg.ListenAddr`, carrying CE2's canary, before returning the listen failure. Add a bounded startup-field treatment, reconcile the hunk restriction, guard, register and disabled-variant mapping, and assert exactly one exit event while classifying any permitted startup event. Web copies `Error.name`, which is a writable instance property; map to code-controlled categories and add a syntactically valid canary-bearing `name` control. | §7.1 (exact three-hunk diff: bind first, then a bounded `core listening`), §5.6 (startup record fields; web categories), §5.1–§5.2 r3 additions, §7.2 (`instanceof` categories), §10 (CE1–CE3, WE3a and WE3b), §17 (reconciliation), Q16, Q21 |
| 5 | **Execution prerequisites and refreshed baseline** (R6/R7, AC7/AC9/AC10). Q1 facts and Q2 authorizations remain unresolved, and the proposed profiles and R1 supply neither. Obtain them before seeking execution approval. Update the current state: the consolidation has finished and its 63 kit entries verify, but its 2 modified documents and 65 new files remain uncommitted. Route A remains recommended; a fresh binding follows technical approval and baseline resolution. | §2.1 (refreshed state), §12.2 and §12.3 (unchanged gates, restated), §12.6 (order of prerequisites), §15.1a, §22 (Q1, Q2, Q9, Q14), §23 |

**Budget (Codex):** retain B7 as a ceiling, at most seven shared cycles, checkpoint after cycle 4 and before cycle 5; zero cycles used. Feasibility is conditional on runner qualification and the corrected control flow. Return for split or budget review if the corrected forecast or the qualified runner cannot support all acceptance evidence.

### 0.5 r3 planning-session disclosures

- **Authority.** The owner authorized sending Codex's r2 corrections to this planning agent and revising the Plan. That is not implementation, hosted-action or execution authorization.
- **Repository writes.** The only repository write is this file.
  - It was revised in place, without a worktree.
  - To keep r1 and r2 byte-exact, it was assembled under `/tmp/g003-plan-r3/` from a hash-verified copy of r2 (`6e6356f3…`) and new sections, then installed with one scripted copy after re-checking that the path still held the r2 bytes.
- **Git.** Only read-only commands with `GIT_OPTIONAL_LOCKS=0`: `status`, `rev-parse`, `ls-files`.
- **Not done.** No install, initialization, application build, `docker build`, Compose command, harness run, GitLab API call, pipeline trigger or external configuration change. No Docker container, network, volume or image was created in r3. P-OWN is a pure-Bash decision model over simulated inventories.
- **What was run** (scratch only, under `/tmp/g003-plan-r3/`; listed in §15.5a):
  - Bash and process probes: `setsid`, process-group signalling, `timeout`, `wait`, traps, process-substitution `tee`;
  - a control-flow model of `validate.sh` with stub harness, preflight, cleanup and scanner, driven through 24 fault cases with real signals;
  - a state-importer unit test;
  - the ownership decision model;
  - read-only hashing of the skill consolidation (its manifest check and the B-WEB-DESIGN identity check).
- **Leftover processes.**
  - Two driver-side `sleep 60` guard processes were orphaned by the case driver itself (not by the model) and were terminated.
  - During two aborted model runs, stuck stub processes, caused by a stub-only `wait` deadlock and a watchdog readiness race in an early model version, were terminated by PID and process group.
  - The final run left no model process. All were scratch processes under `/tmp/g003-plan-r3/`; no developer process was touched.
- **The other agent's skill files** were hashed, not inspected or modified.
- **No credential, token or environment secret was read or printed.**

### 0.6 r3 → Codex: CORRECTION LOOP (2026-10-05)

Codex reviewed r3 (SHA-256 `b100cb52…54b2`) at HEAD `9461c4b…`. It reproduced S0 (`0409148a…`), r1 (`a43aa523…`), r2 (`6e6356f3…`) and the scratch `main.go` (`987c9335…`). It replayed the r3 model: 24 of 24 under the supplied driver, but its additional assertion found that F15's final line lacked `internal=19` (`/tmp/g003-codex-r3-review-sk21rl2t/`). The review is recorded in PROGRESS.md under `### 2026-10-05 — G003 Plan r3 review: CORRECTION LOOP`. **The review text governs.**

**Resolved by Codex (preserved unchanged in r4):**
- parent-held location and P1 import, and supervised harness state;
- containment before creation;
- external TERM, the watchdog and group termination;
- all-or-nothing staging, and the scanned full/minimal/none outcomes;
- genuinely excluded console scanning, and report-before-scrub;
- the no-false-success staging fix;
- core bind-before-log with CE2 cardinality and CE3;
- code-controlled web categories and the regex-valid name canary;
- Q20 and Q21 directions;
- the earlier attribution, schema-negative, step-order and regression work.

| # | Codex correction (abbreviated) | Where r4 addresses it |
|---|---|---|
| 1 | **Retain parent-error provenance** (R7, AC7/AC8/AC10). `final="$(aggregate)"` loses `ST[internal]=19` in the command-substitution subshell, so F15 exits 143 without the marker V18f requires. Keep the harness status and its precedence while retaining the marker in parent state before the final record. Require both `harness=143` and `internal=19`. Add a parent error after a successful harness, requiring nonzero internal status and retained provenance. Rerun all cases. Do not report an orchestration failure as only an external interruption. | §11.2.1 (Provenance), §11.2.2 (P2's post-harness record), §11.2.3 F11 (final-line fields), §11.2.5 (precedence including `signal`; parent-only aggregation), §11.3 V18, V18d, V18f, **V18k**, V18g–i, §15.5b (P-MODEL r4 29 of 29; F15, F15b; non-vacuity run against r3's model) |
| 2 | **Complete cleanup ownership and registry retention** (R7, AC7/AC8). The table required a Compose `working_dir` even for containers selected by a run label, but the harness and RP6 create standalone `docker run` containers without Compose labels. Specify a positively bound path for them, separate from the Compose rule; add interrupted-standalone positives and unrelated label-only negatives; a label or name alone must not authorize deletion. Reconcile F10: deleting `$EV` destroys a local `$REG=$EV/owned` that should survive failed cleanup. Assert retention of the registry and its start and ownership records for local and CI layouts while the secret list is scrubbed; keep 18 (or an earlier failing harness status); never claim an empty inventory. | §10 (`docker_run_owned`), §11.4 (standalone capture through the cidfile; Compose, standalone and ephemeral rules; registry retention), §11.5 (RP6), §11.2.3 F2 and F10, §11.3 V27–V35, §15.5b (P-OWN r4; model F10, F10c, F10h, F10hc, F01, F01c, F14) |
| 3 | **Execution prerequisites.** Q1, Q2, Q14 and Q9 remain unresolved or pending. No Plan revision is needed only to repeat this. | Unchanged §12.6 and §22, still accurate. Restated in the handback, not expanded here. |

**Recurrence and budget (Codex):** the remaining defects are Plan and model state-recording and resource-class/retention coverage, not an architecture contradiction. Retain the single Goal and B7 (checkpoint after cycle 4; zero cycles used). If these mechanisms do not converge in this bounded resubmission, reassess a split before further planning or implementation.

### 0.7 r4 planning-session disclosures

- **Authority.** The owner directed this r4 Plan revision. It is not implementation, hosted-action or execution authorization.
- **Repository write.** The only repository write is this file. It was assembled under `/tmp/g003-plan-r4/build/` from a hash-verified copy of r3 (`b100cb52…`) with scripted, exactly-once replacements and new sections, then installed with one scripted copy after re-checking that the path still held the r3 bytes. No worktree was used.
- **Git.** Only read-only inspection with `git --no-optional-locks` / `GIT_OPTIONAL_LOCKS=0`.
- **Not done.** No install, build, Compose command, harness run, Docker resource creation, GitLab API call, pipeline trigger or external change. No application, test, CI, harness, Goal, Spec, PROGRESS, architecture or skill file was touched.
- **Scratch work under `/tmp/g003-plan-r4/`:**
  - the r4 model, stubs and driver (29 cases);
  - the non-vacuity control run of the r4 driver against the r3 model;
  - P-OWN r4.
- **A driver defect, disclosed.** The first r4 model run accidentally used the **r3 driver copy**, because my write of the new driver was refused (the file had not been read in this session). That copy still contained r3's broad `find /tmp -maxdepth 1 -name 'g003-validate-*' … -exec rm -rf` clean-up, which Codex's replay had deliberately omitted.
  - It matched **nothing**: zero `/tmp/g003-validate-*` entries existed afterwards, and every model evidence root lives under its case's own `TMPDIR`, so none could have matched.
  - Codex's evidence directory `/tmp/g003-codex-r3-review-sk21rl2t/` was confirmed intact.
  - The r4 driver removes that deletion entirely and confines every case to its own directory.
  - Two orphaned driver-guard `sleep 60` processes from that run were terminated.
- **No credential, token or environment secret was read or printed.**

## 1. Summary (r3; r1 and r2 summaries are in the appendices)

G003 completes the existing operational paths. It does not rebuild them. The points below are current.

**r4 changes (Codex r3 review; detail in §0.6):**
- **Provenance.** `signal`, `internal` and `harness_stop` are recorded in the parent before aggregation, and the final line always carries them. An orchestration error during or after the harness shows `internal=19` beside the harness status (V18f, V18k).
- **Standalone ownership.** Standalone containers are bound by a creation-time cidfile; labels are never deletion authority (V27–V31b).
- **Registry retention.** A failed cleanup keeps the registry and its records in both layouts, while the secret list is always scrubbed (V32–V35).
- **Evidence and counts.** Model: 29 of 29. Controls: 80.

- **Diagnostic context (R2/R3)** is unchanged (§4): request ID and W3C `traceparent` version-00 validation on the six health routes, per-hop spans, and propagation on exactly two existing calls (web `/api/ready` → core `/readyz` → intelligence `/healthz`).
- **Logs and privacy (R4).**
  - Exactly one JSON completion record per scoped request (§5.2).
  - Application `event` values are code-authored constants. Framework, library and error text is never copied into a record.
  - **r3:** every `error_type` comes from a code-controlled category (§5.1–§5.2 r3 additions, §5.6).
  - **Error output (r2 direction; r3 corrections):**
    - **Web:** the wrapper returns Next's own 500 with an empty body (plus `no-store` and `X-Request-Id`) instead of rethrowing. `error_type` is an `instanceof` category, never the writable `Error.name`.
    - **Core:** `main.go` has exactly three reviewed hunks — a bounded `core.exit` record, a classifying `ErrorLog`, and **bind-then-log**, so `core listening` carries only the bound port and `app_env` and is never emitted when binding fails.
    - **Intelligence:** bounded startup-failure, uvicorn, library and warning records.
    - **Real-artifact checks:** CE1–CE3, WE0–WE6 (with WE3a and WE3b), IE1 and IE2. These are mandatory at the first approved runtime gate; source reading is not runtime evidence.
- **Health (R1)** is unchanged: statuses, bodies and 2 s deadlines; additive headers only; each single request ≤ 5 s.
- **Contracts (R5)** are unchanged from r2, as accepted by Codex: an independent Go validator, the corrected Python validator, normal-mode digest enforcement, the audit-mode schema negatives with declared classes and derivations, and SD01/SD02.
- **CI (R6/R7).** One gating job runs `infra/ci/validate.sh`. **r3:**
  - **State and signals.** Location and harness supervision run in the parent shell. Isolated phases hand values back only through a validated state file. The harness runs as its own `setsid` process group under a parent-owned watchdog deadline. On a signal or the deadline, the supervisor terminates the whole group with a bounded grace, then finalizes.
  - **Location.** It is checked for containment **before** any directory is created; without a valid location, the run takes a read-only pre-location path and exits 16.
  - **Publication.** Exactly one outcome — full, minimal or none. Staging is all-or-nothing. Scanner errors publish nothing. The artifact summary is labelled pre-publication; the authoritative status is the job's exit code and its final log line.
  - **No masking.** A structural invariant makes any recorded failure produce a nonzero exit.
  - **Harness change.** `run.sh`'s `tee` ignores TERM, INT and HUP, so the harness's own cleanup still runs under a group TERM (probe P-TEE).
  - **Cleanup.** Positive ownership for every resource kind: containers by working directory and absence at start; networks and volumes only through capture rows from owned containers; images only through the run's build or capture records. Unproven resources are kept and reported (18).
  - **Attribution** is unchanged from r2, as accepted.
- **Runner and hosted evidence.** Q1 (runner facts) and Q2 (named hosted actions) are **unresolved owner prerequisites** and must be supplied before execution approval is sought (§12.6). R1 (runner qualification before S1) and H0–H3 are unchanged and need their authorizations.
- **Baseline.** The owner-authorized skill consolidation **has finished**: its 63 kit entries verify against its manifest. But its 2 modified documents and 65 new files are **uncommitted**, so baseline verification stays **PENDING**. S0 is unchanged; route A (commit, then bind BASE) is recommended (Q14).
- **Step order** is unchanged from r2, as accepted: the allow-list is the first write (S1), and CI, docs and harness are integrated before their first static run (S6). The S3, S4 and S6 gates add CE3, WE3a/WE3b and SC2251 shellcheck.
- **Not changed.** No database, schema, migration, public domain API, MCP or production-route change. No production execution client, telemetry vendor, metrics endpoint, deployment or UI change. `local.yml`, `core/go.mod`, the intelligence lockfile and the web dependencies are unchanged. The G021 barrier stays closed.

## 2. Verified baseline (2026-10-05, read-only)

| Item | Value |
|---|---|
| HEAD | `9461c4bdbd8c1ed3e9a9986c70647cd1add57101` ("Define G003 operational foundation and CI specification"), branch `staging` |
| Working tree | Clean at the start of the session: `git --no-optional-locks status --porcelain=v1 -z --untracked-files=all --ignored` was empty. Since 17:22 local time it shows one ignored, tool-local path, `!! .claude/settings.local.json` (§0), plus this Plan once placed. |
| Goal SHA-256 | `63f2b26676c9f1c385639c2078abc586beb48e36d6722c957aa4b03b70d73bb2` (matches handoff) |
| Spec SHA-256 | `d64a691d49219e6945b3fb115bcea59977a9b491f7e2746702e329d1c1c2e949` (matches handoff) |
| ARCHITECTURE.md SHA-256 | `6a6fd49b854ce90acf1a2ce6bd281a4deee3e68e8f4e021d8d737c3b04d90726` (matches handoff; §60 and §61 read) |
| Remote | `origin` `git@gitlab.example:<namespace>/<project>.git`. `ls-remote --symref`: `HEAD → refs/heads/main` (`99e2c79`, "Initial planning"), `refs/heads/staging` = `9461c4b` (equal to local HEAD). The default branch is `main`. |
| GitLab project | Private: the unauthenticated API returns 404. No `glab`, `gitlab-runner` or API token is available in this session, so runners, settings, pipelines and artifacts cannot be inspected here (§12). |
| Governance | SECURITY.md and DECISIONS.md are empty; there is no accepted ADR. BUDGET_POLICY, ROADMAP and PROGRESS were read. G001/G002 acceptance is recorded in PROGRESS and supersedes historical submission wording. |
| Schema digests (must remain unchanged) | `execution-envelope.schema.json` `bf3cec365e291da2cf9b201b9d0194632818721671a5844589954b978accba4f`; `stream-event.schema.json` `e1025031134253ac8d19a514f6d7a4d9563281b5006ebe9be40a84e4d63fd102` |
| Protected harness files | `run-approved.sh` `620a3e84…c53c695` and `run-approved-controls.sh` `314a16a7…7ba9` (both unchanged by G003); `run.sh` `d5e7e5be…41ac8`, `lib.sh` `9f7a3997…57c3` and `render-customizations.tsv` `9abf9e92…279f` (all three modified, §10) |
| `core/cmd/server/main.go` | `95aa5b23…7951`. It equals the rendered manifest and stays **unchanged**, so B-MCP-SOURCE's byte-equality guard needs no reconciliation. |
| Local Docker | Engine 27.5.1, Compose 2.32.4 (Docker Desktop on WSL2); 16 CPUs, 15 GiB RAM, 920 GB free. No `beginify` project containers, volumes or networks. Default ports 3100, 8180, 8100, 5532, 6479, 9100 and 9101 are free, as are the variant port sets. Unrelated developer resources exist (for example `payments-assistant-postgres-1`, `outreacher-*`, `needknow-drafter-*` and `atlas-*`/`harbor-*` images) and are never touched. |

### 2.1 r3 current working-tree state (baseline verification **PENDING**; r2's observation is preserved in Appendix R2)

Read-only, 2026-10-05 at about 23:56 UTC, with `GIT_OPTIONAL_LOCKS=0`.

| Item | r3 observation |
|---|---|
| HEAD | Unchanged: `9461c4bdbd8c1ed3e9a9986c70647cd1add57101` |
| Goal, Spec, ARCHITECTURE | Unchanged: `63f2b266…3bb2`, `d64a691d…e949`, `6a6fd49b…0726` |
| PROGRESS.md | `1d023c5ebea2160e4cc556ef11715a32482d77c88581884f73067df6432e71a3`. Modified by Codex's r2 review record; supervisor-owned. |
| Plan | r2 `6e6356f3…bc82` before this edit. The r3 hash is reported at handoff, not self-bound. |
| **Skill consolidation** | **Finished.** `sha256sum -c skills/vestive-design-skill/references/shadcn-kit.sha256` (run from the skill root): exit 0, **63 of 63 OK**. This agrees with Codex's review.<br>**Still uncommitted:**<br>• ` M skills/vestive-design-skill/README.md` and ` M skills/vestive-design-skill/SKILL.md` (2 modified documents);<br>• 63 untracked files under `assets/shadcn/` and 2 under `references/` (65 new resources).<br>`vestive-shadcn/**` is gone. |
| Everything else | The only other differing paths are `management/v0/PROGRESS.md` (` M`) and this Plan (`??`). **No production, harness, dependency or application file differs.** Status counts: 3 modified, 66 untracked, 1 ignored. |
| B-WEB-DESIGN interaction | **0** web files are byte-identical to any of the 63 consolidated asset files (re-checked at r3). Rechecked by the static suite at S6 and S10 against the BASE Codex binds. |
| Ignored | Only `!! .claude/settings.local.json` (Q13) |

**Baseline status: PENDING. Execution cannot start.**
- **S0 is unchanged** (§15.1, `0409148a…6723`). It rejects this tree at step 5 (exit 24), because the skill paths are not snapshot inputs; B-SKILL-UNCHANGED and B-PROTECTED would likewise fail against the current BASE.
- **r3 adds no exception.**
- **Route A remains recommended (Q14).** The owner commits the completed consolidation under their own authorization; Codex then binds BASE to that commit, together with the supervisor files, in the fresh binding (Q9). This review commits nothing.

**r1 row superseded (as noted in r2).** r1 §2 says `core/cmd/server/main.go` "stays **unchanged**". It is now a planned three-hunk customization (§7.1, Q16, Q21). Its current SHA-256, `95aa5b23…7951`, is the rendered base that B-RENDER-PLAN and B-TOGGLE-OFF-ADDITIONS bind.

## 3. Findings that shape the plan

| # | Area | Finding (inspected) | Consequence |
|---|---|---|---|
| F1 | Core correlation | `Middleware` accepts a single regex-valid request ID and a trace ID. It keeps no span or parent and logs the raw `path`. `r.Header.Get` reads only the first of duplicated headers. `statusRecorder` records the **last** `WriteHeader`. `ParseTraceparent` accepts any version except `ff`. | Rework inside `core/internal/observability` (§7.1). `main.go` is unchanged, because the middleware is already wired there. |
| F2 | Core → intelligence | `health.Intelligence` builds a fresh request with no diagnostic headers. Its 2 s client timeout equals core's 2 s readiness deadline. | Propagate from the request context (§7.1). No new route. |
| F3 | Web | The handlers take no `Request`. `coreReady()` sends no context and returns only a boolean. There is no completion logging. Next's standalone server prints only startup lines. | Add `web/lib/request-context.ts`; customize the two routes and `core-client.ts` (§7.2). |
| F4 | Intelligence logging | `uvicorn main:app` keeps uvicorn's default text access log, which writes the raw request line with its query string to stderr, and its default text error log, which writes tracebacks. `JsonFormatter` has no key or value redaction. `/healthz` and `/readyz` send no `Cache-Control`. | Pure-ASGI completion middleware scoped to the two health paths. Take over the `uvicorn.access` and `uvicorn.error` loggers inside `configure_logging`, so the Dockerfile `CMD` is unchanged. Add redaction and no-store (§7.3). |
| F5 | Python format assertion | `jsonschema` 4.26 registers the `date-time` checker only when `rfc3339_validator` imports. It is absent from `uv.lock`, so **`format: date-time` is silently not asserted**. `main.py` already compensates at runtime, but a conformance check would treat the keyword as success. | Register a project-owned RFC 3339 `date-time` checker, with no new dependency (§8.4). |
| F6 | Python `pattern` | `jsonschema` evaluates `pattern` with `re.search`. Python's `$` also matches before a final `\n`, so `"req-12345678\n"` passes `^[A-Za-z0-9._-]{8,64}$` in Python. It fails under ECMA-262 (JSON Schema's dialect) and under Go RE2. | An ECMA-anchored `pattern` keyword in the existing validator, with an audit of the supported syntax (§8.4, Q3). |
| F7 | Python JSON parsing | `json.loads` accepts `NaN` and `Infinity`; RFC 8259 and Go reject them. | Conformance parsing is strict. Runtime envelope parsing is unaffected, because an envelope has no free numeric fields (§8.4). |
| F8 | Go validator candidates | `google/jsonschema-go` (already indirect via the MCP SDK) ignores `format` during validation. `santhosh-tekuri/jsonschema/v6` v6.0.3 (2026-06-28, Apache-2.0) supports Draft 2020-12, `AssertFormat()`, `AddResource`/`UseLoader` and meta-validation, and implements `date-time` and `uuid`. Its `go.mod` requires `golang.org/x/text v0.14.0` (`regexp2` is test-only). | Use santhosh v6.0.3 in a separate test-tool module, so core's `go.mod` is untouched (§9). |
| F9 | Harness assumptions | The harness probes published ports on `127.0.0.1`, bind-mounts host paths (`$REPO`, `$WORK`, report directories) and refuses busy fixed port sets. | The runner must share loopback and filesystem with its Docker daemon. Docker-in-Docker and socket-only Docker executors do not meet this (§12). |
| F10 | Provenance | `contracts/README.md` and `local.yml` are rendered from templates (`.tmpl`) that differ in the intelligence-disabled render. Customizing them would break B-TOGGLE-OFF-ADDITIONS' equal-base rule. | Neither is changed. The corpus lives in `contracts/conformance/` and compose overlays are avoided (§14). |
| F11 | Nested deadlines | Web's `AbortSignal.timeout(2000)` starts before core's 2 s readiness deadline, which bounds its intelligence probe. A *hung* (paused) intelligence therefore makes core answer at about 2.0 s, and web usually aborts first, so web reports `core: unavailable`. A *stopped* intelligence fails fast and is unaffected. | Existing behaviour; the Spec preserves the deadlines. Assertions follow §6.3. Codex decision in Q4. |
| F12 | Ignored-file gate | G002's S0 gate rejects every ignored path. The Claude Code client's ignored `.claude/settings.local.json` now exists in the checkout. | One exact, reported exception in the G003 gate. Every other ignored path, `.env` and `.env.local` included, still fails (§15.1, demonstrations T1–T3). |

**r2 findings.** These are added to F1–F12 above. Where they conflict, the r2 finding governs. In particular, F1's "`main.go` is unchanged" is superseded by F16 and F17, and F3's "Next's standalone server prints only startup lines" by F15. Sources are cited, and the probes are in §15.5.

| # | Area | Finding (inspected or probed) | Consequence |
|---|---|---|---|
| F13 | GitLab attribution | GitLab's [predefined-variables reference](https://docs.gitlab.com/ci/variables/predefined_variables/) documents:<br>• `CI_MERGE_REQUEST_SOURCE_BRANCH_SHA` and `CI_MERGE_REQUEST_TARGET_BRANCH_SHA`: "The variable is empty in merge request pipelines. The SHA is present only in merged results pipelines."<br>• `CI_MERGE_REQUEST_EVENT_TYPE` is `detached`, `merged_result` or `merge_train`.<br>• `CI_COMMIT_BRANCH` is "Not available in merge request pipelines or tag pipelines".<br>• `CI_COMMIT_SHA` is "The commit revision the project is built for".<br>Further:<br>• [Merge-request pipelines](https://docs.gitlab.com/ci/pipelines/merge_request_pipelines/) "Run on the contents of the source branch only".<br>• [Merged-results pipelines](https://docs.gitlab.com/ci/pipelines/merged_results_pipelines/) (Premium/Ultimate) test "a temporary merged commit … This commit doesn't exist in either branch", and fall back to a standard merge-request pipeline on conflict.<br>• The documented `CI_PIPELINE_SOURCE` values are on the [job rules](https://docs.gitlab.com/ci/jobs/job_rules/) page. | r1's `mr-source` rule required HEAD to equal a variable that is empty in exactly that pipeline type. r2 attributes from `CI_COMMIT_SHA` plus the event type, and treats the merged-result SHAs as contradictory in a detached pipeline (§11.2). |
| F14 | `after_script` and timeouts | GitLab's [`after_script`](https://docs.gitlab.com/ci/yaml/#after_script) documentation says it:<br>• runs in a separate shell, with the working directory reset and no access to `script` changes;<br>• has its own timeout, 5 minutes by default (`RUNNER_AFTER_SCRIPT_TIMEOUT`);<br>• cannot fail a job whose `script` succeeded;<br>• "For jobs that time out: `after_script` commands do not execute by default", unless `RUNNER_SCRIPT_TIMEOUT` and `RUNNER_AFTER_SCRIPT_TIMEOUT` together fit within the job timeout ([runner configuration](https://docs.gitlab.com/ci/runners/configure_runners/)).<br>`artifacts:when: always` excludes timeouts. | Gating, scanning and cleanup verification all happen inside `validate.sh`, under an internal deadline below the job timeout. `after_script` is a non-gating backstop only (§11.1, §11.4). |
| F15 | Next.js error output | Next.js 16.3.8 source, at tag `v16.3.8`:<br>• `packages/next/src/build/templates/app-route.ts` rethrows a route-handler error. The outer catch calls `routeModule.onRequestError(req, err, …, silenceLog=false, …)` and then sends `new Response(null, { status: 500 })`.<br>• `route-modules/route-module.ts` `onRequestError` calls `routerServerContext.logErrorWithOriginalStack(err, 'app-dir')` or `console.error(err)`. Both print the error's message and stack. | r1's §5.4 assumption ("route handlers print nothing") was wrong for thrown errors. The wrapper now catches, logs the bounded record and returns that 500 with an empty body itself, plus `Cache-Control: no-store` and `X-Request-Id`. A real-framework probe (B-WEB-ERROR-OUTPUT) proves it (§5.3, §7.2, §10). |
| F16 | Core exit path | `main.go` prints `run()` errors with `fmt.Fprintln(os.Stderr, err)`. Probe P-GO shows a listen error carrying the configured host verbatim: `listen tcp: lookup bad-CANARY-….invalid …`. `config.Load` messages name variables only, never values, but they are still non-JSON free text. | A bounded JSON `core.exit` record with an `error_type` category replaces the print (§5.6, §7.1). `main.go` becomes a reviewed customization (Q16). |
| F17 | Core `net/http` error log | Go 1.27.1 `net/http/server.go` (`grep` of the local `golang:1.27.1-alpine` image) logs these through `Server.ErrorLog`, or the standard `log` package when it is nil:<br>• superfluous and hijacked `WriteHeader`/`Write` calls (with the caller's function, file and line);<br>• an invalid `Content-Length` value;<br>• conflicting length headers;<br>• `http: panic serving <addr>: <value>` with a stack;<br>• TLS handshake and `Accept` errors.<br>Probe P-GO reproduced the plain-text lines, including a canary-bearing panic value. Go 1.27.1 does **not** log semicolon queries (no such call site; the probe produced no line). `http.ErrAbortHandler` is silent. | `main.go` installs `ErrorLog: observability.ServerErrorLog(logger)`. A classifying writer turns each line into one bounded `http.server.error` record and never copies the text. The probe's prototype produced only canary-free JSON (§7.1, §15.5). |
| F18 | Intelligence startup failure | `main:app` is built lazily in the module's `__getattr__`. In uvicorn 0.54.0, `Config.load_app` catches only `ImportFromStringError`, so a `ConfigError` raised by `create_app` escapes as an uncaught traceback: non-JSON, and it carries the exception message. uvicorn's `LOGGING_CONFIG` gives the `uvicorn` logger a stderr text handler (`uvicorn.error` propagates to it) and `uvicorn.access` a stdout handler whose line includes the request line. `server.py` logs a bind `OSError` object directly (`logger.error(exc)`). | `__getattr__` catches startup failures, writes one bounded `intelligence.startup.failed` record and exits with `SystemExit(3)`. The takeover now replaces the `uvicorn` logger's handlers, and uvicorn records are mapped by message *template* to fixed events, with arguments dropped (§7.3). |
| F19 | Python fallbacks | Records from a logger that has no handlers reach `logging.lastResort`, which writes the raw message to stderr. `warnings` writes plain text unless captured. | A single JSON handler on the root logger, with third-party records reduced to `library.log`, plus `logging.captureWarnings(True)` (§7.3). The real-image and full-log checks fail on any remaining raw line (R2 in §21). |
| F20 | Schema-negative masking | r1 mounted the SN files over the pinned schema paths. In normal mode the digest check runs first, so every SN case would have failed on the digest alone. | Separate audit mode, declared classes and a derivation check, while normal mode keeps digest enforcement (§8.3). |
| F21 | Working tree | The owner-authorized skill consolidation is in progress (§2.1). | Baseline PENDING. No S0 change (§15.1a, Q14). |
| F22 | gitleaks exit codes | Probe P-SCAN (gitleaks v8.30.1 `dir`): a clean directory exits 0. A planted generated token exits with the configured `--exit-code 23`. A missing configuration or path exits **1**, the same as gitleaks' default leak code. | Scans use `--exit-code 23`. Exit 0 means clean, 23 means a finding, and anything else is a scanner failure that fails closed (§11.2). |
| F23 | r1 count error | r1 §14.3 listed 18 validation-tool files but said "(17)". | Corrected in §14.3 (r2: 20 files). |

**r3 findings.** These are added to F1–F23. Where they conflict, the r3 finding governs. The probes are in §15.5a.

| # | Area | Finding (inspected or probed) | Consequence |
|---|---|---|---|
| F24 | Phase state | Codex showed, and P-STATE reproduces, that a body run as `( set -e; f )` cannot set parent variables: r2's location phase left `EV` and `REG` unset in the parent, and its `mktemp -d` had already created a directory. | State-creating steps run in the parent. Isolated phases return values only through a validated state file (§11.2.1). |
| F25 | Process groups | P-SETSID: in a non-interactive shell, `setsid CMD &` keeps `$!` as the PID and makes it the process-group and session leader (`pid = pgid = sid`), so `kill -- -$!` reaches the whole tree. The model's F15 first run showed the group is created **asynchronously**: a signal sent immediately after launch found no group and missed the tree. | A bounded `group_ready` wait follows each launch, and every stop path uses it (§11.2.1). |
| F26 | Harness cleanup under signals | P-TEE, with a stub shaped like `run.sh` (`set -euo pipefail`, `trap cleanup EXIT`, `exec > >(tee -a LOG) 2>&1`): on a group TERM the harness died of SIGPIPE (141) **before** its cleanup, because `tee` died first. When `tee` ignores TERM, INT and HUP, the cleanup ran every time and the harness exited 143. Under GNU `timeout`, a variant with process substitution also skipped its EXIT trap. | `run.sh` changes one line: its `tee` ignores TERM, INT and HUP. GNU `timeout` is replaced by a supervisor-owned watchdog. The harness's own cleanup is best-effort; validate.sh's positive-ownership cleanup is the guaranteed path (§10, §11.2, §11.4). |
| F27 | `errexit` pitfall | The r3 model's first run passed a failing preflight: `! grep …` as a statement is exempt from `set -e` in Bash. | Every check in `validate.sh`, `lib.sh` and `cleanup-owned.sh` uses `if …; then return N; fi`. B-SHELL and G6a add `shellcheck --include=SC2251` for `infra/ci/*.sh` (§10). |
| F28 | Core startup record | `main.go` starts `ListenAndServe` in a goroutine and immediately logs `core listening` with `slog.String("addr", cfg.ListenAddr)`. A failing address is therefore logged raw before `core.exit` (Codex). | Bind first with `net.Listen`; log `core listening` only after a successful bind, with the numeric bound port and the validated `app_env` (§7.1). |
| F29 | Web error category | `Error.name` is a writable instance property, so a name that matches a syntax regex is still data (Codex). | `instanceof` mapping to a fixed category set (§5.1–§5.2 r3 additions) |
| F30 | Cleanup ownership | Network, volume and image rules relied on label or name plus absence at start (Codex). | Positive ownership through capture and build records (§11.4); P-OWN models the decisions. |
| F31 | Working tree | The consolidation is finished but uncommitted (§2.1). | Baseline PENDING; route A (Q14) |

## 4. Diagnostic context contract (R2)

### 4.1 Request ID

- **Accepted inbound value.** Exactly one `X-Request-Id` field value matching `^[A-Za-z0-9._-]{8,64}$`, evaluated with full-string semantics:
  - Go RE2 `MatchString`;
  - JavaScript without the `m` flag;
  - Python `re.fullmatch`.
- **Rejected inbound value.** A missing value, more than one field line, a comma-joined value (Node joins duplicated fields as `a, b`), an over-long or non-matching value, or a value that contains a credential the receiving service holds (§4.4).
  - A rejected value is **never reflected or logged**.
  - Instead the service generates `request_id = 32 lowercase hex` from 16 random bytes (`crypto/rand`, `node:crypto` `randomBytes`, `secrets.token_hex`), and records `request_id_source` as `generated` rather than `inbound`.
- **Response header.** All six scoped health responses carry the effective `X-Request-Id`. Core already returns it on every response.

### 4.2 `traceparent` profile (narrower than W3C; no full-compliance claim)

Based on the [W3C Trace Context Recommendation](https://www.w3.org/TR/2021/REC-trace-context-1-20211123/).

- **Accepted.** Exactly one field value, at most 512 bytes before parsing, that fully matches `^00-([0-9a-f]{32})-([0-9a-f]{16})-([0-9a-f]{2})$`, where:
  - the trace ID is not all zeros;
  - the parent ID is not all zeros;
  - the length is exactly 55, which follows from the pattern.
- **Flags.** Any two lowercase hex digits are accepted. Only the `sampled` bit (`flags & 0x01`) is interpreted. Unknown bits are ignored and **never propagated**: the outbound flags are `01` when sampled is set, otherwise `00`. A newly generated trace uses `00`. The flag confers no authority, sampling, content capture or export.
- **Versions.**
  - `ff` is invalid.
  - **Every version other than `00` is unsupported in this profile.** The service restarts with a new local trace. This is narrower than W3C's SHOULD-parse rule for higher versions, and is documented as such.
  - Uppercase hex, the wrong length, extra trailing fields on `00`, internal whitespace, a non-hex character, a duplicated field or a comma-joined value are all invalid.
- **On an invalid value.** The service generates a new nonzero 16-byte trace ID. The invalid value is never reflected, logged or forwarded, no HTTP 500 is produced, and `trace_source` is recorded as `generated`.
- **Not supported:** `tracestate` and `baggage`. They are never read, forwarded or logged.

### 4.3 Spans and parent relationship (server-span-as-caller convention)

- **Per-hop span.** Every receiving service creates a fresh nonzero 8-byte `span_id` for its server operation. It regenerates if the value is all zeros, and it never copies the incoming parent ID.
- **Parent.** `parent_span_id` is the inbound `traceparent` parent ID, present only when that `traceparent` is valid.
- **Outbound.** Each outbound call carries `traceparent: 00-<trace_id>-<caller span_id>-<flags>`. The parent ID is "the ID of the current operation", as W3C requires when mutating `traceparent`. No separate client span is recorded.
- **Resulting chain on the full stack:**
  - `web.parent_span_id` = the caller's parent, absent when the context was generated;
  - `core.parent_span_id` = `web.span_id`;
  - `intelligence.parent_span_id` = `core.span_id`.

  All three records share `trace_id` and `request_id`, and their span IDs are distinct.
- **Datastores.** Probes to PostgreSQL, Redis and MinIO carry no tracing protocol.

### 4.4 Credentials supplied as diagnostic values (Q5)

- **Core** treats an inbound request ID that contains any configured secret value (`Config.SecretValues()`, reached through the redacting logger's handler) as invalid, and generates a new one. Core never echoes, logs or forwards that ID. Its logger additionally scrubs secret values from every string (existing behaviour).
- **Intelligence** does the same for `INTELLIGENCE_SERVICE_TOKEN` and `OPENROUTER_API_KEY`, and its formatter scrubs them (§7.3).
- **Web** holds no credentials, so it cannot classify a value as one. It logs only regex-bounded IDs. Core refuses to adopt or forward such a value.

### 4.5 Header allow-list and isolation

- **Outbound headers set by application code:** exactly `X-Request-Id` and `traceparent`.
  - Transport defaults remain: Go `User-Agent` and `Accept-Encoding`, and undici's standard request headers.
  - No inbound header, cookie, `Authorization`, `baggage` or `tracestate` is ever copied.
- **Request-local state:**
  - **Core:** values stored in the per-request `context.Context`.
  - **Web:** an explicit `RequestContext` argument. There is no module-level or `AsyncLocalStorage` state.
  - **Intelligence:** local variables inside the ASGI call. There is no `ContextVar`, because the health handlers make no outbound call.
- Isolation therefore holds by construction. It is still tested under concurrency, failure and cancellation (§16, AC2).

## 5. Log contract (R4)

### 5.1 Envelope for every application-owned JSON line (all three services)

| Key | Type | Rule |
|---|---|---|
| `time` | string | RFC 3339 UTC, millisecond precision, `Z`. Go: slog `ReplaceAttr`; Node: `toISOString()`; Python: `datetime.fromtimestamp(created, UTC)` |
| `level` | string | one of `debug`, `info`, `warn`, `error`. Python `WARNING` → `warn`, `CRITICAL` → `error` |
| `service` | string | `web`, `core` or `intelligence` |
| `event` | string | Core: slog's message key is renamed to `event`. Intelligence: `msg` becomes `event`; the existing `logger` key is kept. |

Startup and error records keep their messages (for example core's `core listening` and intelligence's `configured`) under `event`.

### 5.2 Completion record (`event = "http.request.completed"`)

| Key | Type | Rule |
|---|---|---|
| `method` | string | `GET`, `HEAD`, `POST`, `PUT`, `PATCH`, `DELETE` or `OPTIONS`; anything else becomes `OTHER` |
| `route` | string | Normalized, never the raw path:<br>• core: the matched `ServeMux` pattern path (`/healthz`, `/readyz`), else `unmatched`;<br>• web: the fixed handler route (`/api/health`, `/api/ready`);<br>• intelligence: `/healthz`, `/readyz` |
| `status` | integer | The final HTTP status: the first `WriteHeader`, or an implicit 200 on the first write. `0` only when the connection aborts before headers. |
| `duration_ms` | integer ≥ 0 | Monotonic elapsed time, floored |
| `request_id`, `request_id_source` | string, enum | §4.1 |
| `trace_id`, `trace_source` | 32 hex nonzero, enum | §4.2 |
| `span_id` | 16 hex nonzero | §4.3 |
| `parent_span_id` | 16 hex nonzero | **Present only** when the inbound parent was valid |
| `trace_flags` | `00` or `01` | The outbound flags (§4.2) |
| `outcome` | enum | `completed`, `cancelled` (client gone or context cancelled) or `failed` (panic or exception) |
| `error_type` | string `^[A-Za-z][A-Za-z0-9_.]{0,63}$` | Only when `outcome = failed`. Go: `panic`; Node: the error's `name`; Python: the exception class name. Never message text or a stack. |
| `core_probe` | enum | **Web `/api/ready` only:** `ok`, `not_ready`, `timeout` or `unreachable` |

- **Level.** `failed` → `error`; status 500 or above → `warn`; otherwise `info`.
- **Strictness.** No other key is allowed, apart from intelligence's `logger`.
- **Machine-checked contract.** The record is checked by `infra/validation/tools/internal/logrecord/completion.schema.json`, through the `logcheck` command (§9).

**r3 additions to §5.1–§5.2 (corrections 4 and 5; they supersede r2's additions, preserved in Appendix R2).**
- **Constant events** (unchanged from r2). An application-owned `event` value is a constant authored in repository code: a template, never formatted with runtime values. `logcheck` validates every `event` against the per-service catalogue `infra/validation/tools/internal/logrecord/events.json`. Any uncatalogued value fails.
- **Web `error_type` is a code-controlled category, never instance data.** The wrapper maps a thrown value with `instanceof` checks against built-in constructors, in this fixed order:
  1. a `DOMException` whose `name` is exactly `AbortError` or `TimeoutError` → that name; any other `DOMException` → `DOMException`;
  2. `TypeError`, `RangeError`, `SyntaxError`, `ReferenceError`, `EvalError`, `URIError`, `AggregateError` → the matching constructor's name, **as written in the mapping table, not read from the value**; a subclass reports its built-in ancestor;
  3. any other `Error` → `Error`;
  4. a non-`Error` value → `NonError`.

  The value's `name`, `message`, `cause`, stack and own properties are never read into the record. The category set is closed, and `logcheck` enforces it through `events.json`.
- **Core and intelligence `error_type`.**
  - **Core** uses only the fixed categories in §5.6 (`panic` for handler panics).
  - **Intelligence** uses `type(exc).__name__`. That is taken from the exception's class (code), never from instance attributes, and is validated against `^[A-Za-z][A-Za-z0-9_.]{0,63}$`; `Exception` otherwise.
- **Startup, exit and framework records** use the same §5.1 envelope (§5.6).

### 5.3 Exactly one record (r2; supersedes r1 §5.3 in Appendix R1)

- **Core** is unchanged from r1.
  - The middleware logs once in a `defer` guarded by a flag.
  - A handler panic is logged with `outcome: failed`, `error_type: panic`. The middleware then re-panics with `http.ErrAbortHandler`, which `net/http` neither prints nor passes to `ErrorLog` (F17). The connection aborts as before.
  - A request whose context was cancelled is logged with `outcome: cancelled`.
- **Web** (changed in r2; F15). `withCompletionLog` logs exactly once, in `finally`.
  - **Normal path.** The handler's own response is returned, with `X-Request-Id` added.
  - **On an exception** the wrapper does **not** rethrow. It records `outcome: failed` and the bounded `error_type`, and returns `new Response(null, { status: 500, headers: { "Cache-Control": "no-store", "X-Request-Id": <effective id> } })`.
    - This is the status and empty body that Next 16.3.8 itself sends after logging, so the only observable differences are the two additive headers (Q15).
    - Because no error reaches Next, Next prints nothing.
  - **On abort.** When `request.signal.aborted` is set at completion, the wrapper records `outcome: cancelled`, with the same no-rethrow rule.
- **Intelligence** is unchanged from r1.
  - The ASGI middleware logs once in `finally`.
  - An exception becomes `failed` with its class name, then is re-raised, so uvicorn's "Exception in ASGI application" is emitted as the bounded `uvicorn.asgi_exception` record (§5.6).
  - `CancelledError` becomes `cancelled`, then is re-raised.
- **Scope** is exactly the six health routes, as in r1.
  - Core keeps logging every request it receives, including 404 and 405, with `route: unmatched`.
  - Intelligence's `/internal/*` path keeps its execution export record and no longer has an access line (F4).
- **Duplicate writes.** The recorder forwards every call unchanged, as before, but reports the first status.

### 5.4 Framework output (r2; supersedes r1 §5.4 in Appendix R1)

`logcheck` classifies **every** line of each service's raw output, in both stacks and in the error-output checks. Any line that is not valid JSON with a catalogued event, and does not match the exact allow-list below, fails.

| Service | Actual framework behaviour (source) | G003 handling | Allowed non-JSON lines |
|---|---|---|---|
| core | `net/http` writes `ErrorLog` lines in plain text; panics other than `ErrAbortHandler` print value and stack (F17, probe P-GO). `main()` prints `run()` errors raw (F16). | `ErrorLog` is the classifying adapter (`http.server.error` with a fixed `error_type`). `main()` writes one `core.exit` record (§5.6). Panics in scoped handlers are re-raised as `ErrAbortHandler`. | none |
| web | Next 16.3.8 prints its startup banner. For a thrown route error it calls `console.error(err)` or `logErrorWithOriginalStack`, which print message and stack (F15). | Route handlers never let an error escape the wrapper (§5.3). | Exactly r1's startup patterns: `^ ▲ Next\.js [0-9]+\.[0-9]+\.[0-9]+$`, `^   - Local:        http://[^ ]+$`, `^   - Network:      http://[^ ]+$`, `^ ✓ Starting\.\.\.$`, `^ ✓ Ready in [0-9.]+m?s$`, and the empty line. They are confirmed against the observed output of the real-framework probe (WE0, §10); any deviation is recorded and returned for review, and the list is never widened silently. |
| intelligence | uvicorn 0.54.0 sends text through the `uvicorn` logger's stderr handler and access lines (with the request line) to stdout. `load_app` lets non-import errors escape as tracebacks. Unhandled loggers fall back to `logging.lastResort`; `warnings` writes text (F18, F19). | `configure_logging` takes over the `uvicorn` logger, disables `uvicorn.access`, adds the root JSON handler and captures warnings. `main.__getattr__` turns a startup failure into one bounded record and `SystemExit(3)` (§7.3). | none |

Runtime crashes outside request handling, for example a Go `fatal error` or a Python interpreter abort, are not request or startup paths G003 controls. If one occurs in a tested run, `logcheck` fails it as an unclassified line; it is reported, never suppressed.

### 5.5 Redaction (existing protections preserved; equivalents added)

- **Core (existing, retained):**
  - sensitive-key regex `password|passwd|secret|token|api_?key|authorization|cookie|dsn|credential`, including nested groups;
  - configured-secret-value scrubbing in the message and in string and non-string attributes.
- **Intelligence (new, in `JsonFormatter`):**
  - the same key regex, applied recursively to `fields`;
  - configured-secret-value scrubbing of `event` and of every string;
  - exception type only (existing);
  - string values capped at 512 characters.

  `JsonFormatter()` keeps its no-argument constructor, because `test_observability.py` uses it. Secrets are supplied through `configure_logging(level, secrets=...)`, which `create_app` calls.
- **Web.** Records are built only from the fixed keys in §5.2, with values that are validated IDs, enums or integers. A key-regex guard drops any sensitive-looking key. No URL, header, error message or configuration value is ever logged.
- **Never logged anywhere:**
  - raw headers, cookies or query strings;
  - request or response bodies;
  - prompts, completions, retrieved content or envelopes;
  - authorization references;
  - tenant or user identifiers;
  - URLs, DSNs or exception text.

### 5.6 Startup, exit and framework error records (r3; r2 version in Appendix R2)

All of these use the §5.1 envelope (`time`, `level`, `service`, `event`). They never contain message text, stack frames, argument values, host names, URLs, addresses or configuration values, apart from the single numeric port and validated `app_env` named below.

| Service | `event` | Extra keys (bounded) | Emitted when |
|---|---|---|---|
| core | **`core listening`** (level `info`; the existing constant) | **r3:** exactly `port` (an integer from the **bound** listener's `*net.TCPAddr`; `0` if not TCP) and `app_env` (`development`, `test` or `production`; validated by `config.Load`). The configured listen address and host are **never** logged. | Only **after** `net.Listen` succeeds (§7.1, hunk 3). A failed bind produces no `core listening` record. |
| core | `core.exit` (level `error`) | `error_type`, one of `config_invalid`, `database_pool_invalid`, `listen_failed` (`*net.OpError` with `Op == "listen"`, now returned directly by `net.Listen`), `serve_failed`, `shutdown_timeout`, `unknown`. `config_invalid` adds `config_keys` from the allow-list derived from `config.Load` (§7.1). | `run()` returns an error; exit status 1 |
| core | `http.server.error` (level `warn`) | `error_type` ∈ {`superfluous_write_header`, `hijacked_write_header`, `hijacked_write`, `invalid_content_length`, `conflicting_length_headers`, `handler_panic`, `tls_handshake_error`, `accept_error`, `other`} | `net/http` writes to `Server.ErrorLog` |
| core | `http.request.completed` | §5.2; `error_type: panic` only | §5.3 |
| intelligence | `intelligence.startup.failed` (level `error`) | `error_type`: `type(exc).__name__` (class, regex-validated) | `create_app()` raises while uvicorn builds `main:app`; exit 3 |
| intelligence | `uvicorn.process_started`, `uvicorn.startup_waiting`, `uvicorn.startup_complete`, `uvicorn.listening`, `uvicorn.shutting_down`, `uvicorn.shutdown_waiting`, `uvicorn.shutdown_complete`, `uvicorn.process_finished`, `uvicorn.app_load_failed`, `uvicorn.invalid_request`, `uvicorn.asgi_exception`, `uvicorn.server_error`, `uvicorn.other` | `error_type` (class) when exception information is present | Mapped by **template** (`record.msg`), with arguments dropped (as in r2). `uvicorn.listening` carries **no** host or port: the template's arguments are dropped. |
| intelligence | `library.log`; `python.warning` | `logger` (regex-validated name), or `category` (warning class) | As in r2 |
| web | `http.request.completed` | §5.2; `error_type` from the closed `instanceof` category set (§5.1–§5.2 r3 additions) | §5.3 |

**Catalogued startup events per service** (permitted, classified, never exempted):
- core: `core listening` with exactly `port` and `app_env`;
- intelligence: the `uvicorn.*` lifecycle events and `configured`;
- web: none (Next's startup banner stays the exact non-JSON allow-list in §5.4).

Every other startup line fails `logcheck`. The canary checks (CE1–CE3, IE1, IE2, WE0) scan these records like any other line.

Intelligence's own logger records keep `event = record.msg`, as in r2.

## 6. Health semantics and bounds (R1)

### 6.1 Endpoints (bodies and status unchanged; additive headers only)

| Service | Route | Body (exact bytes) | Status | Headers after G003 |
|---|---|---|---|---|
| core | `GET /healthz` | `{"status":"alive"}\n` | 200 | `Content-Type: application/json`, `Cache-Control: no-store`, `X-Request-Id` (all existing) |
| core | `GET /readyz` | `{"status":…,"product_name":…,"checks":{"object_store":…,"postgres":…,"redis":…}[,"optional":{"intelligence":…}]}\n` | 200 / 503 | as above (existing) |
| web | `GET /api/health` | `{"status":"alive"}` | 200 | `Cache-Control: no-store` (existing), **`X-Request-Id` (new)** |
| web | `GET /api/ready` | `{"status":"ready"\|"unready","checks":{"core":"ok"\|"unavailable"}}` | 200 / 503 | `Cache-Control: no-store` (existing), **`X-Request-Id` (new)** |
| intelligence | `GET /healthz` | `{"status":"ok"}` | 200 | **`Cache-Control: no-store` and `X-Request-Id` (new)** |
| intelligence | `GET /readyz` | `{"status":"ready"}` (also when no model is configured) | 200 | **`Cache-Control: no-store` and `X-Request-Id` (new)** |

- Core's required checks remain `postgres`, `redis` and `object_store`. `intelligence` stays under `optional` and is absent when intelligence is disabled.
- The deadlines remain:
  - core readiness: 2 s;
  - core's intelligence client: 2 s;
  - web's fetch: 2 s.
- No error text, DSN, URL or credential appears in any body.
- Intelligence's body bytes are unchanged: FastAPI's default dict response and `JSONResponse` render identically (`separators=(",", ":")`).

### 6.2 Timing protocol

- **Clock.** Each single request is timed by `httpprobe` (§9) with Go's monotonic clock.
- **Where it runs.** Inside the target service's network namespace (`--network container:<cid>`): web on `127.0.0.1:3000`, core on `127.0.0.1:8180`, intelligence on `127.0.0.1:8000`.
- **Bound.** Every injected timeout or outage case must return in **≤ 5000 ms**, or the criterion fails.
- **Recovery.** Recovery polling is separate. It has a documented finite deadline of **90 s per case** and is never used to time a single request.

### 6.3 Outage and timeout matrix (disposable copies only)

The cases run on the enabled copy (`variants_ops`, ports 3200, 8280, 8200, 5632, 6579, 9200 and 9201) and on the disabled copy (`variants_off`).

| Case | Injection | Assertions (each single request ≤ 5 s) | Enabled | Disabled |
|---|---|---|---|---|
| O1–O3 | `compose stop` postgres / redis / minio | core `/readyz` 503 naming the dependency; web `/api/ready` 503 (`core_probe: not_ready`); both liveness probes 200; continuity records (§16 AC3); recovery to 200 | ✓ | ✓ |
| O4–O6 | `docker pause` postgres / redis / minio | Same statuses; web `core_probe` is `not_ready` or `timeout`, because core answers at about 2 s (F11); continuity in web and core; unpause, then recovery | ✓ | ✓ |
| O7 | stop intelligence | core 200 with `optional.intelligence = unavailable`; web 200; no intelligence record; recovery restores the three-hop chain | ✓ | — |
| O8 | pause intelligence | **Core `/readyz` (direct) 200 with `optional.intelligence = unavailable`.** Web `/api/ready` returns within the bound with either 200 or 503 `core_probe: timeout`. The observed value is recorded (F11, Q4). After unpause, intelligence may log 0 or 1 record for the queued probe; more than one, or a mismatched one, fails. | ✓ | — |
| O9 | stop core | web `/api/ready` 503 (`core_probe: unreachable`), `/api/health` 200, both no-store; only a web record | ✓ | ✓ |
| O10 | pause core | web `/api/ready` 503 (`core_probe: timeout`). After unpause, core logs at most one record for that request ID, with `outcome: cancelled` and matching IDs. | ✓ | ✓ |

The existing B-OUTAGE and B-PERSIST then run unchanged on the same copy, after recovery has been asserted.

## 7. Service changes (r3; r2 §7 is in Appendix R2, r1 §7 in Appendix R1)

### 7.1 Core (`core/internal/observability`, `core/internal/health`, and `core/cmd/server/main.go`, which is customized)

- **`correlation.go`, `logging.go`, `checks.go`** are unchanged from r2 (§7.1 of r2: the context profile, per-hop spans, completion records, the `ReplaceAttr` envelope and outbound propagation).
- **`exit.go` (new)** contains:
  - `LogExit(w io.Writer, err error)`, which writes exactly one `core.exit` record (§5.6) and never `err.Error()`;
  - the `config_keys` allow-list;
  - **`ListenPort(a net.Addr) int`**, which returns `a.(*net.TCPAddr).Port`, or 0.
- **`errorlog.go` (new):** `ServerErrorLog(logger *slog.Logger) *log.Logger`, the classifying writer (§5.6), as in r2.
- **`core/cmd/server/main.go` (customized; Q16, Q21): exactly three unified-diff hunks (3 lines of context), shown below.**
  - Hunk 1 drops the unused `fmt` import.
  - Hunk 2 replaces the raw exit print.
  - Hunk 3 installs `ErrorLog` and **binds before logging**.

  The intended file's SHA-256 is **`987c933560ef77ea9a2daa4a3bec4c9747ea69d578f3c9c9d8e6179d013551ea`**. It was produced from the current file (`95aa5b23…7951`) in a scratch copy and is gofmt-clean (`gofmt` in `golang:1.27.1-alpine`, `--network none`; not compiled, because the module's dependencies are not downloaded in planning). This value is `MAIN_GO_G003_SHA` in B-MCP-SOURCE (§10). Any other byte change is an escalation (§21).

```diff
@@ -8,7 +8,6 @@
 	"context"
 	"errors"
 	"flag"
-	"fmt"
 	"log/slog"
 	"net"
 	"net/http"
@@ -33,7 +32,7 @@
 		os.Exit(healthcheck(os.Args[2:]))
 	}
 	if err := run(); err != nil {
-		fmt.Fprintln(os.Stderr, err)
+		observability.LogExit(os.Stderr, err)
 		os.Exit(1)
 	}
 }
@@ -107,10 +106,15 @@
 		Addr:              cfg.ListenAddr,
 		Handler:           observability.Middleware(logger, mux),
 		ReadHeaderTimeout: 5 * time.Second,
+		ErrorLog:          observability.ServerErrorLog(logger),
+	}
+	ln, err := net.Listen("tcp", cfg.ListenAddr)
+	if err != nil {
+		return err
 	}
 	errCh := make(chan error, 1)
-	go func() { errCh <- srv.ListenAndServe() }()
-	logger.Info("core listening", slog.String("addr", cfg.ListenAddr), slog.String("app_env", cfg.AppEnv))
+	go func() { errCh <- srv.Serve(ln) }()
+	logger.Info("core listening", slog.Int("port", observability.ListenPort(ln.Addr())), slog.String("app_env", cfg.AppEnv))
 
 	select {
 	case <-ctx.Done():
```

  **Behaviour.**
  - A failing bind now returns the `*net.OpError` from `net.Listen` before any `core listening` record exists, so the run emits exactly one `core.exit`/`listen_failed`.
  - A successful bind logs `core listening` with the bound port and `app_env` only.
  - `Serve(ln)` and `Shutdown` keep the existing semantics. `http.ErrServerClosed` is still the normal shutdown result.
  - Routes, the remaining imports, timeouts, the health wiring, `healthcheck` and shutdown are unchanged.

  **Reconciliation** (unchanged in kind from r2; updated in detail):
  - one register row;
  - the allow-list entry;
  - `APPLICABLE` + `core/cmd/server/main.go` (a plain base-template file, so the disabled render has the same rendered base);
  - B-MCP-SOURCE pins `MAIN_GO_G003_SHA` and keeps every security assertion (§10, §17).
- **New tests:**
  - `context_test.go` and `propagation_test.go`: as in r2.
  - **`exit_test.go`:** as in r2, plus `ListenPort` for a `TCPAddr`, a `UnixAddr` (0) and nil (0).
  - **`errorlog_test.go`:** as in r2.

  `observability_test.go` and `health_test.go` are unchanged and must pass.
- **Real-binary checks** (B-CORE-ERROR-OUTPUT, §10):
  - **CE1:** configuration invalid.
  - **CE2:** listen failure. **Exactly one** output line, `core.exit`/`listen_failed`; **no** `core listening` line; no canary.
  - **CE3 (new):** a valid bind. Exactly one `core listening` line, with `port` and `app_env` keys only; then on TERM, a clean exit 0 with no further line; no canary.

### 7.2 Web (routes and BFF; no UI file changes)

- **`web/lib/request-context.ts` (new; `server-only`; no relative imports)** is as in r2: `deriveContext`, `outboundHeaders` and `withCompletionLog`. **r3:**
  - `error_type` comes from the closed `instanceof` category mapping (§5.1–§5.2 r3 additions), implemented as a fixed table of built-in constructors.
  - The thrown value's `name`, `message`, `cause`, stack and own properties are never read into the record.
  - The catch path still returns `new Response(null, { status: 500, headers })` with `Cache-Control: no-store` and `X-Request-Id`, and never rethrows (Q15).
  - Completion semantics are unchanged: one record, `outcome` `failed` or `cancelled`.
- **`core-client.ts` and both routes, `web/package.json` scripts:** as in r2.
- **`web/tests/unit/request-context.test.mjs`:** r2's coverage, plus:
  - an `Error` whose `name` is a **syntactically valid canary-bearing identifier** (for example `Cx<hex>`, matching `^[A-Za-z][A-Za-z0-9_.]{0,63}$`) → `Error`, with the canary absent;
  - a `TypeError` subclass with a custom `name` → `TypeError`;
  - a `DOMException` with an arbitrary name → `DOMException`;
  - an `AbortError` → `AbortError`;
  - the message, `cause`, own-property and non-`Error` canaries, as in r2.
- **`web/tests/fixtures/error-probe/{wrapped-route.ts,control-route.ts}`:** as in r2. **r3:** `wrapped` gains `mode=badname-valid` (a regex-valid canary-bearing `name`, giving WE3a) next to `badname` (an invalid name, giving WE3b).

### 7.3 Intelligence (`src/`, check stage only in the image)

- **`intelligence/src/observability/http.py` (new)** is `CompletionLogMiddleware`, as in r1.
- **`intelligence/src/observability/export.py` (customized).**
  - **`JsonFormatter`:** the §5.1 envelope, `service`, redaction (§5.5) and a 512-character value cap, as in r1. Plus the r2 event rules (§5.6):
    - intelligence loggers use the template and never `getMessage()`;
    - uvicorn records are mapped by template;
    - other libraries become `library.log`;
    - warnings become `python.warning`;
    - exceptions keep their type only.
  - **`configure_logging(level=INFO, secrets=())`** is idempotent, with the secrets updatable as in r1. **r2:**
    - it installs **one** JSON stream handler on the root logger, with `intelligence` at INFO and the root at WARNING;
    - it replaces the `uvicorn` logger's handlers with the same handler, keeping `propagate=False`;
    - it sets `uvicorn.access` to `disabled=True` with no handlers;
    - it calls `logging.captureWarnings(True)`.

    The test asserts there are no duplicate lines. `JsonFormatter()` keeps its no-argument constructor.
  - `ObservabilityHook` is unchanged, so content capture stays constructor-only and off by default.
- **`intelligence/src/main.py` (customized).** As in r1: `create_app` passes the two secrets to `configure_logging` and the middleware, and the health handlers return `JSONResponse(..., headers={"Cache-Control": "no-store"})`. **r2:**
  - `__getattr__` builds `_app` inside `try: … except Exception as exc:`.
  - On failure it calls `configure_logging()` if it has not yet run, writes one `intelligence.startup.failed` record (`error_type` only) and raises `SystemExit(3) from None`.
  - uvicorn's `Config.load_app` does not catch `SystemExit`, so the process exits 3 without a traceback. The real-image check verifies this (B-INTEL-ERROR-OUTPUT, IE1).
- **`intelligence/src/streaming/schemas.py` (customized):** §8.4, unchanged from r1.
- **New tests** (no new packages):
  - `test_health_context.py`, `test_logging_privacy.py`, `test_conformance.py` and `conformance_exchange.py`, as in r1. r2 extends the last two (§8.3).
  - **`tests/test_error_output.py` (new):**
    - the startup-failure path through `main.__getattr__` with an invalid environment that includes a canary-bearing variable *name*: `SystemExit(3)`, one record, no canary;
    - the uvicorn template mapping for each template in §5.6;
    - an unknown template becomes `uvicorn.other`;
    - a third-party logger warning becomes `library.log` with no message;
    - `warnings.warn(<canary>)` becomes `python.warning` with no canary;
    - exactly one handler, with no duplicate lines;
    - the `uvicorn.access` logger is disabled.
- **`infra/docker/intelligence.Dockerfile` (customized)** is as in r1: the `check` stage only copies `conformance/intelligence/`.

## 8. Contract conformance (R5)

### 8.1 Layout (fixtures only; `contracts/intelligence/*.schema.json` unchanged)

```text
contracts/conformance/README.md                         provenance, case IDs, how to run, stream-check rules
contracts/conformance/intelligence/expectations.json    independently declared outcome per case, schema digests, corpus rules
contracts/conformance/intelligence/cases/envelope/*.json
contracts/conformance/intelligence/cases/event/*.json
contracts/conformance/intelligence/cases/ndjson/*.ndjson
contracts/conformance/intelligence/schema-negative/*.json
```

- **Why this location.**
  - It is kept outside `contracts/intelligence/`, so it is not copied into the runtime image's `/srv/contracts/intelligence`.
  - It is absent from the intelligence-disabled render, which B-TOGGLE-OFF already requires.
  - It is fixture-only, consistent with "contract schemas' test/fixture/documentation areas" (Q11).
- **Expectation pinning.** `expectations.json` pins the two schema SHA-256 values from §2. Both validators refuse to run if the bytes they load differ.
- **Fixture data.** All data is synthetic, hand-authored for G003, and uses RFC-style UUIDs. Values never look like real credentials, so gitleaks' generic rules are not triggered; secret-scan is not loosened.

### 8.1a r2 addition: declared schema-negative entries

`expectations.json` gains a `schema_negative` array, declared independently of either validator's output. Each entry gives:
- `id`;
- `base`: `execution-envelope` or `stream-event`;
- `mutation`: one JSON-Patch-style `add` or `replace` at a JSON Pointer, or `truncate-bytes` with a byte length, or `absent`;
- `class`: the expected audit classification.

It also declares the two normal-mode digest controls, SD01 and SD02. They have no fixture file, because they are derived at test time inside a scratch directory (§8.2a).

### 8.2 Case catalogue (file name = lowercase ID plus a slug; the expected result is fixed here)

**Envelope, valid (14):**

| ID | Case |
|---|---|
| EV01 | Minimal: required fields only, empty `allowed_tools` and `context_refs`, one message |
| EV02 | Full: `group_id`, `business_entity_id`, `callback_grant_ref`, all three tool levels, context refs, system/user/assistant messages |
| EV03 | `request_id` of length 8 |
| EV04 | `request_id` of length 64 with `.`, `_` and `-` |
| EV05 | Offset `+02:00` with a 9-digit fraction |
| EV06 | `2028-02-29` (leap year) |
| EV07 | `23:59:60Z` (leap second) |
| EV08 | Uppercase UUIDs |
| EV09 | 256 tools and 256 context refs |
| EV10 | 256 messages |
| EV11 | Every string maximum: content 200000, `authorization_ref` 512, `callback_grant_ref` 512, tool name 128, ref 256, kind 64 |
| EV12 | `trace_id` all zeros. The existing wire behaviour is preserved: the envelope schema has no nonzero rule, and none is imported. |
| EV13 | Nil UUID (shape-valid) |
| EV14 | `capability: embedding` |

**Envelope, invalid (50):**

| ID | Case |
|---|---|
| EI01 | Truncated object |
| EI02 | Trailing comma |
| EI03 | `NaN` literal |
| EI04 | Top-level array |
| EI05–EI14 | Each of the 10 required fields missing (in schema `required` order) |
| EI15 | `allowed_tools` as an object |
| EI16 | `request_id` as a number |
| EI17 | Unknown field |
| EI18 | `api_key` |
| EI19 | `password` |
| EI20 | `access_token` |
| EI21 | `authorization` |
| EI22 | `capability: vision` |
| EI23 | Tool level `admin` |
| EI24 | Extra property on a tool |
| EI25 | Role `tool` |
| EI26 | Empty messages |
| EI27 | 257 messages |
| EI28 | 257 tools |
| EI29 | Content of 200001 characters |
| EI30 | `request_id` of 7 characters |
| EI31 | `request_id` of 65 characters |
| EI32 | `request_id` with a space |
| EI33 | **`request_id` with a trailing `\n`** (F6) |
| EI34 | Uppercase `trace_id` |
| EI35 | `trace_id` of 31 characters |
| EI36 | `organization_id` not a UUID |
| EI37 | `user_id` of 35 characters |
| EI38 | `group_id` containing `g` |
| EI39 | **`2027-02-30T00:00:00Z`** |
| EI40 | **Month 13** |
| EI41 | **Hour 24** |
| EI42 | **Minute 60** |
| EI43 | **Second 61** |
| EI44 | **`12:00:60Z`** |
| EI45 | **`2027-02-29`** |
| EI46 | No offset |
| EI47 | Space separator |
| EI48 | 10-digit fraction |
| EI49 | Lowercase `t`/`z` |
| EI50 | **Offset `+24:00`** |

The bold cases EI39–EI45 and EI50 satisfy the `pattern` and fail **only** on `format: date-time`. They prove the format assertion is effective in both validators.

**Event, valid (15):**

| ID | Case |
|---|---|
| XV01 | `text_delta` with empty text |
| XV02 | `tool_call` with nested arguments |
| XV03 | `tool_result`, `is_error: false`, object content |
| XV04 | `tool_result`, `is_error: true`, null content |
| XV05 | `tool_result`, string content |
| XV06 | `citation` without locator |
| XV07 | `citation`, locator of 256 characters |
| XV08 | `structured_output` without `schema_ref` |
| XV09 | `structured_output` with `schema_ref` |
| XV10 | `usage`, all values |
| XV11 | `usage` with null `input_tokens`, `output_tokens` and `estimated_cost` |
| XV12 | `usage` with `units: 0` and zero tokens |
| XV13 | `usage` with tokens `1.0` (an integer-valued number) |
| XV14 | `error`, `retryable: true`, message of 1024 characters |
| XV15 | `done` |

**Event, invalid (34):**

| ID | Case |
|---|---|
| XI01 | Malformed JSON |
| XI02 | Unknown type |
| XI03 | Missing type |
| XI04 | Extra field |
| XI05 | Numeric text |
| XI06 | `done` without `request_id` |
| XI07 | `request_id` with a space |
| XI08 | `request_id` with a trailing `\n` |
| XI09 | Bad `call_id` |
| XI10 | Array arguments |
| XI11 | `tool_result` without `is_error` |
| XI12 | Empty citation ref |
| XI13 | String data |
| XI14 | Negative input tokens |
| XI15 | Negative output tokens |
| XI16 | Negative latency |
| XI17 | Negative cost |
| XI18 | Negative units |
| XI19 | Units `1.5` |
| XI20 | `usage` without `trace_id` |
| XI21 | Unknown usage capability |
| XI22–XI25 | `usage` with `organization_id`, `user_id`, `group_id`, `business_entity_id` (one case each) |
| XI26–XI30 | `usage` with the content fields `prompt`, `completion`, `messages`, `content`, `input` (one case each) |
| XI31 | Uppercase error code |
| XI32 | Error message of 1025 characters |
| XI33 | Extra field on `done` |
| XI34 | Uppercase `configured_provider` |

**NDJSON stream files (7).** Rules:
- every record is LF-terminated;
- no blank line;
- at least one record;
- every line parses strictly and validates against `stream-event`.

| ID | Case | Expected |
|---|---|---|
| NV01 | `text_delta`, `usage`, `done` | valid |
| NV02 | `error`, `done` | valid |
| NI01 | Final line truncated, no LF | invalid |
| NI02 | Non-JSON line in the middle | invalid |
| NI03 | Schema-invalid event in the middle | invalid |
| NI04 | Blank line | invalid |
| NI05 | Empty file | invalid |

These rules are test-profile rules for checking producer output. They make no wire-contract change.

### 8.2a Schema-negative and digest controls (r2; supersedes r1's "Schema-negative (7)" block in Appendix R1)

**Two paths, never mixed (correction 4):**
- **Normal mode** (`corpus`, `ndjson`, `peer` and the pytest corpus test). The schema digests are checked **before anything else**. Any byte difference from the two pinned SHA-256 values fails with class `schema-digest-mismatch`, with zero cases evaluated. No option or argument disables this.
- **Audit mode** (`schema-audit`). It accepts only files under `contracts/conformance/intelligence/schema-negative/`, or the canonical schema paths for the positive control. It never evaluates cases and never runs the digest step. Its only output is a classification. It exits 0 only for `ok`.

**Audit stage order** (identical in both languages, implemented separately; the first failing stage sets the class):

| Stage | Check | Class on failure |
|---|---|---|
| 1 | The file exists and is a regular file | `schema-missing` |
| 2 | Strict JSON parse (RFC 8259; `NaN` and `Infinity` rejected); the top level is an object | `schema-malformed-json` |
| 3 | `$schema` is exactly `https://json-schema.org/draft/2020-12/schema` | `schema-wrong-draft` |
| 4 | Meta-validation against the embedded Draft 2020-12 meta-schema, with no network | `schema-meta-invalid` |
| 5 | Every keyword is in the audited profile (§8.3) | `schema-unsupported-keyword` |
| 6 | Every `format` is `date-time` or `uuid` | `schema-unsupported-format` |
| 7 | Every `$ref` starts with `#/$defs/` and resolves locally. Nothing is fetched: Go's loader fails every URL, Python's registry has no retriever, and containers run `--network none`. | `schema-remote-ref` |
| 8 | Compile with assertions on | `schema-compile-failed` (no declared case expects it) |

**Cases.** Each SN fixture is the canonical base schema with exactly one mutation, so that every earlier stage passes and only the intended stage can fail:

| ID | Base | Single declared mutation | Declared class (both languages) |
|---|---|---|---|
| SN01 | envelope | `absent`: no file | `schema-missing` |
| SN02 | envelope | `truncate-bytes` to the first 64 bytes of the canonical file | `schema-malformed-json` |
| SN03 | envelope | `replace /properties/authorization_ref/type` with `5` | `schema-meta-invalid` |
| SN04 | envelope | `add /dependentSchemas` with `{"group_id": {"required": ["business_entity_id"]}}` (meta-valid; outside the profile) | `schema-unsupported-keyword` |
| SN05 | envelope | `add /properties/authorization_ref/format` with `"ipv4"` | `schema-unsupported-format` |
| SN06 | stream-event | `replace /oneOf/0/$ref` with `"https://example.invalid/x.json"` | `schema-remote-ref` |
| SN07 | envelope | `replace /$schema` with `"http://json-schema.org/draft-07/schema#"` | `schema-wrong-draft` |
| SD01 | envelope | **normal mode:** the canonical bytes plus one trailing LF, written to a scratch schema directory. This file is still valid, meta-valid and inside the profile. | `schema-digest-mismatch`, 0 cases evaluated |
| SD02 | envelope | **normal mode:** SN03's bytes in place of the canonical file, in a scratch schema directory | `schema-digest-mismatch`, 0 cases evaluated. This proves that normal mode never reaches meta-validation with unpinned bytes. |

**Derivation check (`contractcheck sn-derive`).** For every declared SN entry, it applies the declared mutation to the parsed canonical base and requires structural equality with the parsed fixture. For SN02 it requires byte equality with the truncated prefix; for SN01, that the file is absent. A fixture with a second, undeclared difference therefore fails. This shows that each SN case isolates a single defect.

**Positive audit control.** Both canonical schemas classify as `ok` in both languages, so the audit path is not vacuously failing.

**The EI39 control** is retained, in normal mode with canonical schemas: a scratch copy of `expectations.json` with EI39 set to `valid` makes both validators exit nonzero, naming EI39.

That is 120 document and stream cases, 7 schema-negative cases, 2 digest controls and 2 positive audit controls. All eight event variants appear among the valid cases. **No production schema byte changes;** every mutation exists only as a committed SN fixture or in a scratch directory.

### 8.3 Validators (r2; independent; neither reads the other's output; r1 text in Appendix R1)

- **Supported keyword profile** (unchanged from r1). The audit is shared by specification but implemented separately in each language: `$schema`, `$id`, `$defs`, `$ref` (local `#/$defs/...` only), `title`, `description`, `type`, `properties`, `required`, `additionalProperties`, `items`, `minItems`, `maxItems`, `minLength`, `maxLength`, `pattern`, `enum`, `const`, `oneOf`, `minimum`, `format`. The only formats allowed are `date-time` and `uuid`.
- **Go (`contractcheck`, santhosh v6.0.3).** It uses `NewCompiler()` (Draft 2020-12) and `AssertFormat()`; a `UseLoader` that fails every URL (meta-schemas are embedded); `AddResource` under `file:///contracts/...` URLs; strict JSON decoding; and RE2 patterns. Subcommands:
  - `corpus`: normal mode. Digest first, then the audit stages 1–8, then the cases. Writes `results-go.json`; exits nonzero on any mismatch.
  - `ndjson`: normal mode.
  - `produce-envelopes`: the test-only Go producer, as in r1.
  - **`schema-audit --schema FILE --as envelope|event`** (new): audit mode only. Prints `{"file": …, "class": …}`.
  - **`schema-negative --schemas DIR --corpus DIR`** (new): runs, for each declared entry, SN01–SN07 through `schema-audit`, `sn-derive`, the positive audit control, and SD01/SD02 through `corpus` in a scratch directory. It compares every class with the declaration and writes `schema-negative-go.json`. A wrong class fails, even when the exit status is nonzero.
  - **`sn-derive`** (new): §8.2a.
- **Python.** The existing `streaming.schemas` validator (§8.4), in the `intelligence-check` image, with strict JSON parsing.
  - The digest pin and the audit stages are implemented in the **test tooling** (`tests/conformance_exchange.py` and helpers used by `tests/test_conformance.py`), not in the runtime module.
  - `tests/test_conformance.py`:
    - parametrized over every corpus case, as in r1;
    - **plus** SN01–SN07, asserting the declared class from `expectations.json`;
    - SD01/SD02 through `tmp_path` copies, asserting that `SchemaDigestMismatch` is raised before any case is evaluated;
    - the positive audit control on both canonical schemas.

    It is asserted not to skip.
  - `tests/conformance_exchange.py` (script) has the subcommands `corpus`, `peer` and `ndjson` (r1) and **`schema-negative`** (new), which writes `schema-negative-python.json`.
- **Shared result format** (unchanged from r1): `{"validator", "schema_sha256", "corpus_sha256", "results": [{"id", "valid", "class"}], "counts", "skipped": 0}`. The schema-negative results use `{"id", "class", "expected_class", "match"}`.

### 8.4 Python validator conformance corrections (`schemas.py`; schema bytes unchanged; Q3)

- **ECMA-262 `pattern` semantics.**
  - `validators.extend(Draft202012Validator, {"pattern": _ecma_pattern})`.
  - Each schema pattern is translated once: an unescaped `$` outside a character class becomes `\Z`.
  - Patterns containing a construct whose semantics differ between ECMA-262 and Python `re` are rejected when the validator is built: `\d`, `\w`, `\s`, `\b`, `\B`, `\p`, `(?`, backreferences and lookarounds. All current schema patterns use only literal classes, quantifiers, anchors, groups and alternation.
- **`date-time` checker.** A project-owned `FormatChecker` adds `date-time`: the RFC 3339 §5.6 grammar plus calendar validation (month length and leap years).
  - Hours 00–23, minutes 00–59, seconds 00–60, offset hours 00–23 and offset minutes 00–59.
  - Second 60 is allowed only when the UTC-normalized time is 23:59.
  - `T`/`Z` are accepted in either case; this is irrelevant in practice, because the schema pattern already requires uppercase.
  - `uuid` remains jsonschema's built-in checker.
- **Audit.** §8.3, plus an assertion that `date-time` and `uuid` are present in `checker.checkers`.
- **Effect at runtime.** Intelligence's `/internal/v1/execute` now rejects, with 422, two classes of envelope it previously accepted only through gaps:
  - an envelope whose pattern-bounded identifier ends in a newline;
  - an envelope whose `deadline` is calendar-invalid. These already received 422 through `main.py`'s parse fallback, with the identical response `{"path":"/deadline","keyword":"format"}`.
- **Leap-second note.** A valid leap-second deadline (`23:59:60Z`) remains schema-valid, but `main.py`'s `fromisoformat` rejects it, as it does today. That pre-existing runtime check is not changed and is noted in OPERATIONS.md.

### 8.5 Peer exchange (harness; scratch `$WORK/contracts-exchange`, mode 777, mounted into both containers)

1. `contractcheck produce-envelopes`, which writes `go-envelopes.ndjson`.
2. `conformance_exchange.py peer`. Python accepts and executes the Go envelopes, then emits `python-events.ndjson`.
3. `contractcheck ndjson --require-types text_delta,usage,error,done`, which checks `python-events.ndjson`.
4. The harness derives corrupt copies of both real files:
   - the last line truncated;
   - a garbage line inserted;
   - an invalid event or envelope appended;
   - a blank line.

   The peer checker must reject each one. A checker that ignores a bad line cannot pass this step.

## 9. Validation tools module (test tooling, not runtime; r2; r1 text in Appendix R1)

- **Module:** `infra/validation/tools/`, `module validationtools`, `go 1.27.1`, as in r1.
  - Direct requirement: `github.com/santhosh-tekuri/jsonschema/v6 v6.0.3`.
  - Indirect requirement: `golang.org/x/text`, at the version minimal version selection resolves.
  - Both are recorded in `go.sum`.
  - It is resolved once at S2, in a container, with a module cache under the execution's external scratch root (`$EXT/gomod`, not a Docker volume as in r1). That is the only network step (§15.6).
- **Commands:**
  - `cmd/contractcheck` (§8.3), with the new subcommands `schema-audit`, `schema-negative` and `sn-derive`.
  - `cmd/httpprobe`, as in r1: repeatable and duplicate headers, header and body values read from files, `--parallel`, `--abort-after`, `--timeout`, JSON output.
  - `cmd/logcheck`, as in r1: line classification, completion-schema validation, exactly one record per request and service, and chain checks. **r2:**
    - validates every record's `event` against the per-service catalogue in `internal/logrecord/events.json` (§5.1–§5.2 r3 additions, §5.6);
    - validates `error_type` against each event's enumerated categories;
    - adds a `startup` mode for the error-output checks (exit status and an expected event and category set);
    - adds a `scan` mode that reports canary hits by file and line number without printing the canary.

    Its unit-test negatives grow to include an uncatalogued event, a free-text `error_type`, and a JSON line whose `event` carries an exception message.
- **Image:** `infra/docker/validation-tools.Dockerfile`, with stages `build`, `check` and `runtime` (distroless static nonroot), as in r1. Images are tagged `g003-vtools-<run>`, labelled, and removed at cleanup.
- **Execution.** `--network none`, except `httpprobe`, which joins a service's network namespace. No production network is joined and no port is published.
- **Exclusions.** Nothing here is linked into core. B-MCP-SOURCE asserts that no `jsonschema` package appears in `go list -deps ./cmd/server`.

## 10. Harness changes (`infra/validation/`; every existing assertion preserved; r3; r2 table in Appendix R2, r1 table in Appendix R1)

| Check | Suite | Change |
|---|---|---|
| B-SHELL | static | Also shellchecks `infra/ci/validate.sh`, `infra/ci/lib.sh`, `infra/ci/validate-controls.sh`, `infra/ci/cleanup-owned.sh` and `infra/ci/runner-preflight.sh`. **r3:** a second pass `shellcheck -S info --include=SC2251` over `infra/ci/*.sh` fails on any `!`-negated statement, which `set -e` ignores (F27). |
| B-HADOLINT | static | Adds `validation-tools.Dockerfile` |
| B-DOCKERIGNORE | static | Adds `validation-tools.Dockerfile.dockerignore` (`.env` and `.env.*`) |
| B-DEPS | static | Adds the tools-module rule: exactly one direct requirement (`santhosh-tekuri/jsonschema/v6 v6.0.3`), and `go.sum` present. The existing rules are unchanged (goose and MCP pins, intelligence lockfile equal to the render, web C1–C12 policy). |
| **B-MCP-SOURCE (reconciled; Q16, §17)** | static | The **same security assertions** remain:<br>• the server binary links neither the MCP SDK, `core/internal/mcp` nor migration code;<br>• no binary depends on fixture or isolation packages;<br>• no transport handler or `WithIsolateDDL`;<br>• `core/internal/mcp` does not import `net/http`;<br>• the routes are exactly `GET /healthz` and `GET /readyz`.<br>**Changed:** the byte-equality test `sha(main.go) == manifest_sha(main.go)` is replaced by `sha(main.go) == MAIN_GO_G003_SHA`, **`987c933560ef77ea9a2daa4a3bec4c9747ea69d578f3c9c9d8e6179d013551ea`**: the reviewed three-hunk file of §7.1, pinned in `run.sh` at S6.<br>**Added:**<br>• `go list -f '{{join .Imports " "}}' ./cmd/server` equals the pinned rendered import set minus `fmt`;<br>• no `jsonschema` package appears in `go list -deps ./cmd/server`.<br>The rendered base stays bound by B-RENDER-PLAN (register row) and B-TOGGLE-OFF-ADDITIONS. The reason message changes from "server main.go unchanged" to "server main.go equals its reviewed G003 digest". |
| B-WEB-BOUNDARY | static | The r1 additions (`request-context.ts` imports `server-only` and is imported only by `app/api/**/route.ts` and `lib/core-client.ts`; the core fetch passes only `outboundHeaders(ctx)`), plus: the `tests/fixtures/error-probe/*` files exist only under `web/tests/` and are never imported by `app/`, `components/` or `lib/`. The routes remain exactly `/`, `/api/health` and `/api/ready`. |
| **B-CI-CONFIG** (new) | static | §11.1 assertions on `.gitlab-ci.yml` (through yq), plus the README `platform:ci` block |
| **B-CI-CONTROLS** (new) | static | `infra/ci/validate-controls.sh` groups A–E (§11.3). It includes Docker-real cleanup controls using `busybox:1.37.0` with control labels, and the nested real `contracts` run of V4. |
| B-WRAPPER-CONTROLS, B-DEPS-CONTROLS | static | Unchanged (W1–W9; C1–C12, C8b, C8c) |
| B-RENDER-PLAN, B-RENDER-REPEAT | provenance | The register count changes from 8 to **19** (3 G001 + 5 G002 + 11 G003, §14.2). Exactly 19 conflicts and proposals. |
| **contracts suite** (new; root only; no stack) | contracts | • **B-CONTRACT-TOOLS:** the `validation-tools` check stage passes; the intelligence-check image builds; the runtime intelligence image has no `/srv/conformance`.<br>• **B-CONTRACT-GO** and **B-CONTRACT-PY:** in normal mode, each validator matches every expectation, the digests match, the audit passes and nothing is skipped.<br>• **B-CONTRACT-AGREE:** as in r1.<br>• **B-CONTRACT-PEER:** §8.5.<br>• **B-CONTRACT-NEGATIVE (r2):** in both languages, with `--network none`:<br>&nbsp;&nbsp;– SN01–SN07 in audit mode, each with its declared class;<br>&nbsp;&nbsp;– the positive audit control (`ok`) for both canonical schemas;<br>&nbsp;&nbsp;– `sn-derive` passes;<br>&nbsp;&nbsp;– SD01 and SD02 in normal mode, each `schema-digest-mismatch` with 0 cases evaluated;<br>&nbsp;&nbsp;– EI39 flipped in a scratch copy: both validators exit nonzero, naming EI39.<br>A wrong class is a FAIL even when the exit status is nonzero. |
| **B-HEALTH-ROOT** (new) | root | Host probes on the actual root (ports 3100, 8180, 8100): six routes with status, exact body, no-store and a valid `X-Request-Id` |
| **B-CORE-ERROR-OUTPUT** (new) | root | Runs, like the other two error-output checks, while the root images exist, after B-CHECKS-ROOT and before the `platform:reset` block. It uses the root project's built core runtime image, run with `docker run --rm --network none --label g001.run=$RUN_ID` and environment files under `$WORK` (mode 600), with per-run canaries:<br>• **CE1:** required variables missing; canary values in `DB_PASSWORD` and `REDIS_PASSWORD`. Expect exit 1, exactly one JSON line `core.exit`/`config_invalid` whose `config_keys` are in the allow-list, and no canary.<br>• **CE2:** otherwise valid configuration with `CORE_LISTEN_ADDR=bad-<canary>.invalid:8180`. Expect exit 1 and **exactly one** output line, `core.exit`/`listen_failed`: no `core listening` record (bind-then-log, §7.1) and no canary. Today's code would emit both a `core listening` record carrying the raw address (F28) and an `Fprintln` line (P-GO).<br>• **CE3 (r3):** valid configuration with `CORE_LISTEN_ADDR=:8180` and a canary in `PRODUCT_NAME`, which is never logged. Expect exactly one `core listening` record with keys exactly `time level service event port app_env`, `port` = 8180 and `app_env` = `test`. After `docker stop` (TERM), expect exit 0 and no further line. No canary.<br>All output is classified by `logcheck startup`. Every line must be a catalogued event (§5.6); none is exempted. |
| **B-WEB-ERROR-OUTPUT** (new) | root | A disposable `copy_repo`, with the two fixture routes copied into `app/api/__probe/{wrapped,control}/route.ts`. It is built with `web.Dockerfile --target runtime` under a run-labelled tag, and run with `--network none`, `PROBE_CANARY=<canary>`, harmless identity values, and `CORE_INTERNAL_URL=http://127.0.0.1:9`. `httpprobe` joins its namespace:<br>• **WE0:** only the allow-listed startup lines (§5.4) and JSON appear;<br>• **WE1:** `mode=throw` → 500, empty body, no-store, `X-Request-Id`, one `failed`/`Error` record;<br>• **WE2:** `nonerror` → `NonError`;<br>• **WE3a (r3):** `badname-valid`, a **regex-valid** canary-bearing `name` (for example `Cx<hex>`) → `error_type` `Error` and no canary anywhere;<br>• **WE3b:** `badname`, an invalid name → `Error`;<br>• **WE4 (positive control):** `control` → 500, and the canary **is** present in Next's raw output. Otherwise the check FAILs as vacuous.<br>• **WE5:** a malformed percent-encoded query carrying a canary on `wrapped?mode=ok` → no non-JSON line beyond WE0 and no canary;<br>• **WE6:** a client abort during `mode=slow` → one `cancelled` record and no framework line.<br>The canary is absent from everything except WE4's control output. The control output is deleted after the assertion and is never retained. |
| **B-INTEL-ERROR-OUTPUT** (new) | root | The root project's built intelligence runtime image, `--network none`:<br>• **IE1:** `APP_ENV` missing, plus a variable named `AI_MODEL_CANARY<hex>=openai:x`, whose *name* reaches the `ConfigError` message. Expect exit 3, every line JSON, exactly one `intelligence.startup.failed`/`ConfigError`, no canary, and no traceback.<br>• **IE2:** valid test configuration. Start, wait for `uvicorn.startup_complete`, then stop: every line is a catalogued JSON event, `uvicorn.*` templates are mapped, and there is no text line. |
| B-CHECKS-ROOT | root | Unchanged command; it now also runs the new unit, conformance and error-output tests |
| **B-OBS-CONTEXT, B-OBS-CHAIN, B-OBS-TIMEOUTS, B-OBS-LOGS, B-OBS-PRIVACY** (new) | variants (`variants_ops`, before B-OUTAGE) | §16 AC1–AC4 on the enabled disposable copy, as in r1. B-OBS-LOGS uses the r2 event catalogue. |
| **B-OBS-OFF** (new) | variants (`variants_off`, before B-BROWSER-OFF) | As in r1 |
| B-TOGGLE-OFF-ADDITIONS | variants | `APPLICABLE` adds the r1 six (`core/internal/observability`, `core/internal/health`, `web/app/api`, `web/lib/core-client.ts`, `web/lib/request-context.ts`, `docs/foundation/OPERATIONS.md`) **and `core/cmd/server/main.go`**: 28 → **35** |
| B-SECRET-CANARY | final | Adds the G003 canaries (§16 AC4) to the scan list. With `--secrets-file`, the list lives at that external path, which is outside `$WORK` and so needs no exclusion. Canary request specifications live in `$WORK/obs-inputs/` (excluded like `secrets*`) and are deleted after use. |
| ~~B-ARTIFACT-SCAN~~ (r1 proposal, **withdrawn**) | — | Artifact scanning has a single implementation: `validate.sh`'s raw and staged scans (§11.2 F3 and F7), where retention and upload happen. This resolves the r1 raw-versus-staged mismatch. Generated secrets and canaries inside the harness's own lifetime remain covered by B-SECRET-CANARY. |
| **B-REQUIRED-CHECKS** (new) | final | Every ID in `infra/validation/required-checks.tsv` for the executed suites is present exactly once and PASS. Missing, duplicate, FAIL or UNMET → FAIL. |
| `all` | — | Becomes `static provenance contracts root migrations variants`. Required IDs: 50 existing + **19** new = **69** (B-ARTIFACT-SCAN withdrawn; B-CORE-, B-WEB- and B-INTEL-ERROR-OUTPUT added). |
| **run.sh options and hooks (r3)** | — | Defaults keep today's behaviour; `run-approved.sh` is unchanged.<br>• `--owned-registry FILE` appends registry rows (§11.4): `work_root`, `label g001.run=<RUN_ID>`, `project <name>` (from `register_project`), `rootenv <sha256>`, **`image <tag> <id>` immediately after every image the run builds** (`docker build -t`, Compose builds), and **capture rows** from `owned_capture DEST` (§11.4).<br>• `owned_capture` is called after every start or `up` of a registered project, and first in the cleanup trap. It writes `container <id>`, `network <id>`, `volume <name>` and `image <tag> <id>` rows for that project's owned containers and what they are attached to, mount and run.<br>• `--secrets-file FILE`: as in r2.<br>• **One-line `tee` change (F26):** `exec > >(trap '' TERM INT HUP; exec tee -a "$HARNESS_LOG") 2>&1`. The harness's own EXIT-trap cleanup can then still log and run under a group TERM; KILL still ends `tee`. Normal runs are unaffected.<br>• **r4: `docker_run_owned PURPOSE ARGS…` (`lib.sh`)** for every **standalone** container that is detached or may outlive a few seconds. It runs `docker run --cidfile "$REG/cid/<seq>.cid" …` with the run label, then appends `standalone <id> <purpose>` from that cidfile (§11.4).<br>&nbsp;&nbsp;– Harness sites: the boundary-control probe `g001-probe-$RUN_ID` (`-d`); the migration integration-test container; the CE3 core container (`-d`); the WE probe web server (`-d`); the IE2 intelligence container (`-d`).<br>&nbsp;&nbsp;– Short, bounded foreground tool runs (`yq`, shellcheck, hadolint, busybox `nc`/`nslookup`/`wget`, `go list`, the validators, `httpprobe`, CE1/CE2/IE1) stay plain `docker run --rm`; the daemon removes them automatically (§11.4, ephemeral class). |
| report copies | — | The end-of-run `--report-dir` copy list adds `obs-*.tsv`, `obs-timings.tsv`, `contracts-results-{go,python}.json`, `contracts-agree.tsv`, `schema-negative-{go,python}.json`, `error-output-{core,web,intelligence}.tsv` (statuses and categories only, never captured output) and `ci-controls.log`. These are the names `validate.sh` F4 and F5 expect. |
| cleanup | — | Also removes the run's `g003-*` images, by registered tag and ID. The inventory diff is unchanged. The accepted `docker compose -p … down` backstop stays, and CI cleanup does not rely on it (§11.4). Under signals the harness cleanup is best-effort (F26); `validate.sh` F2 is the guaranteed path. |

**Unchanged:** B-SCOPE, B-PROTECTED, B-SKILL-UNCHANGED, B-LITERALS (which covers the new files: no identity literal, GitLab path or runner tag in `infra/` or `docs/foundation/`), B-NAMING, the migration suite, B-PORTS-*, B-BOUNDARIES-* and their controls, B-PERSIST, B-RENAME, B-BROWSER-RENAME, B-DBRENAME, B-SOURCE-HASH, B-DOCKER-ONLY, B-WEB-DESIGN, both browser matrices (95 tests each), `run-approved.sh` and its controls.

## 11. GitLab CI (R6/R7; r2; r1 §11 in Appendix R1)

### 11.1 `.gitlab-ci.yml` (exact intended content; `<RUNNER_TAG>` comes from the owner's Q1 facts and is written at S6)

```yaml
# Baseline CI (docs/foundation/OPERATIONS.md). One gating job runs the repository's containerized
# validation entry point. There is no publish, release or deploy job, and no CI/CD secret is used.
workflow:
  rules:
    - if: $CI_PIPELINE_SOURCE == "merge_request_event"
    - if: $CI_PIPELINE_SOURCE == "push" && $CI_COMMIT_BRANCH == "staging"
    - if: $CI_PIPELINE_SOURCE == "push" && $CI_COMMIT_BRANCH == $CI_DEFAULT_BRANCH
    - when: never

stages: [validate]

validate:
  stage: validate
  tags: [<RUNNER_TAG>]
  timeout: 3h
  resource_group: platform-validation-host
  interruptible: false
  variables:
    GIT_STRATEGY: clone
    GIT_CLEAN_FLAGS: -ffdx
    RUNNER_SCRIPT_TIMEOUT: 2h45m
    RUNNER_AFTER_SCRIPT_TIMEOUT: 10m
  script:
    - bash infra/ci/validate.sh --publish-dir ci-evidence
  after_script:
    - bash infra/ci/cleanup-owned.sh --after-script
  artifacts:
    when: always
    expire_in: 90 days
    access: developer
    paths: [ci-evidence/]
```

**r2 changes from r1:**
- the two `RUNNER_*_TIMEOUT` variables;
- the `--after-script` mode of `cleanup-owned.sh`.

Why (F14):
- Their total, 2h55m, fits within `timeout: 3h`, so that on runners that support these variables, `after_script` still runs if the script times out.
- The internal deadline in `validate.sh` (§11.2) is the primary mechanism and does not depend on the runner version (F-R4).
- `after_script` is a non-gating backstop: it cannot fail a successful job, and its output is informational.
- `artifacts:when: always` does not upload on job timeout. The internal deadline therefore ends the harness early enough for the finalizer to scan and publish within the job.

**B-CI-CONFIG asserts:**
- **Structure.** The top-level keys are exactly `workflow`, `stages` and `validate`. The four workflow rules are exact. There is exactly one job.
- **No bypass.** No `allow_failure`, `when: manual` or `when: delayed`, no job-level `rules` or `only`/`except`, no `rules:changes`, `needs`, `retry`, `parallel`, `trigger`, `environment`, `release`, `pages`, `image` or `services`.
- **Script.** It is exactly the single line above, and equal to the README `platform:ci` command.
- **`after_script`.** It is exactly `bash infra/ci/cleanup-owned.sh --after-script`.
- **Artifacts.** Exactly as shown: `when: always`, one directory, `access: developer`, `expire_in: 90 days`, no `untracked`.
- **Variables.** Exactly the four shown. No secret-like names. `RUNNER_SCRIPT_TIMEOUT` plus `RUNNER_AFTER_SCRIPT_TIMEOUT` is at most `timeout`, and `timeout` is at most 3h.
- **Tags.** Exactly one tag matching `^[A-Za-z0-9._-]{1,64}$`. Its value is the owner-supplied tag; Codex checks it against Q1, and hosted evidence records `CI_RUNNER_TAGS`. It is not repeated in `docs/` or `infra/`, so a tag containing an identity literal cannot trip B-LITERALS.

**Pipeline behaviour** (as in r1, with r2 attribution in §11.2):
- Merge-request pipelines test the source content (`detached`), or the merged result when the project has merged results enabled (Premium/Ultimate, F13; owner fact F-R10).
- Pushes to `staging` and to the default branch (`main`) also run.
- Feature-branch pushes, tags, schedules, web, API and trigger pipelines run nothing. If such a pipeline ever reached the job, attribution would reject it with 16.
- A push to `staging` while an MR from `staging` is open yields two gating pipelines for the same commit. This is accepted.
- `resource_group` serializes this project's validation jobs on the shared host ports.

### 11.2 `infra/ci/validate.sh` and `infra/ci/lib.sh` (aggregate entry point; Bash ≥ 4.4; r3; r2 version in Appendix R2)

```text
validate.sh [--evidence-dir DIR] [--publish-dir ci-evidence]
```

#### 11.2.1 Execution model (corrections 1 and 2)

**Parent-held state.** `validate.sh` keeps all cross-step state in **parent-shell variables**: `EV`, `REG`, `LOC_OK`, `HEAD_SHA`, `TREE_SHA`, `STATUS_DIGEST`, `BUDGET_S`, `TESTED_CONTENT`, `HPID`, `WPID`, `DEADLINE`, `ST[...]`, `MAIN_DONE`. Only two kinds of code may set them.

1. **Parent steps,** which run directly in the parent shell and use explicit checks (`if …; then …; fi` and `|| return N`; never `!`-negated statements, F27):
   - L0 location;
   - the state import;
   - the harness launch, wait and supervision;
   - signal handling;
   - candidate selection;
   - aggregation.
2. **Isolated phases,** run as `set +e; ( set -e; FUNC ); rc=$?; set -e`. They report **only** a status in `ST[NAME]`, because subshell assignments never reach the parent (F24, P-STATE).
   - A phase that computes values writes them to a state file, `$EV/state/<phase>.env`.
   - The parent imports that file with **`state_import FILE KEY=REGEX…`**. The importer accepts only `^[A-Z][A-Z0-9_]*=…$` lines; every declared key exactly once; no unknown, duplicate or missing key; each value matching its regex; a regular, non-symlink file. It assigns with `printf -v`, never `source` or `eval`.
   - Any violation fails the import with 16.
   - The P1 keys are `HEAD_SHA` (40 hex), `TREE_SHA` (40 hex), `STATUS_DIGEST` (64 hex), `BUDGET_S` (`^[1-9][0-9]{0,5}$`) and `TESTED_CONTENT` (enum).
   - P-IMPORT: 8 of 8 unit cases, including an injection attempt with no side effect.

**Harness supervision (parent; F25, F26).** GNU `timeout` is not used.

- **Launch.**
  ```text
  setsid bash infra/validation/run.sh … >"$EV/raw/harness-console.log" 2>&1 &   HPID=$!
  ```
  `HPID` is both the PID and the process-group ID (P-SETSID).
  - The parent then waits for the group to exist (`group_ready`: up to 5 s polling `kill -0 -- -$HPID`). The model's F15 first run showed that signalling immediately after launch can miss the group.
  - It records `HPID` in `$EV/state/harness.pgid`, for diagnosis only.
- **Deadline.**
  ```text
  setsid bash -c 'sleep "$1"; kill -USR1 "$2"' watchdog "$BUDGET_S" "$$" &   WPID=$!
  ```
  `group_ready` follows. `trap 'DEADLINE=1' USR1` is set before the launch. The parent `wait`s on `HPID`; USR1 interrupts the wait.
- **Stop sequence** (`stop_harness_group`; used for the deadline, for signals, and for any harness still running at finalization):
  1. `group_ready`, or else TERM the PID itself;
  2. `kill -TERM -- -$HPID` to the **whole group** (foreground Docker CLIs, the harness Bash, its `tee` — which ignores TERM after the r3 harness change — and any backgrounded children);
  3. poll `kill -0 -- -$HPID` for a grace period of **150 s** (the harness's own cleanup, for example `compose down`, runs inside it);
  4. then `kill -KILL -- -$HPID`;
  5. reap with `wait`;
  6. record whether the group is gone.
- **Recorded harness status:**
  - **124** at the deadline;
  - **128 + n** on signal n;
  - otherwise `run.sh`'s own status.
  - If processes remain in the group after a normal exit (stragglers), they are stopped and an INFO `stragglers=1` is recorded.
- **The watchdog** is stopped (`kill -TERM -- -$WPID`, then reaped) as soon as the harness wait returns, and at finalization.

**Signals.**
- `trap 'on_signal TERM 15' TERM`, and likewise for INT (2) and HUP (1).
- `on_signal` records the interruption in the parent: `SIG_NAME`, `ST[signal]=128+n`.
  - It stops the watchdog.
  - If a harness is running, it stops the harness group, records `ST[harness]=128+n` and `HARNESS_STOP=signal`.
  - It then runs `exit 128+n`, which runs the EXIT trap.
- `finalize` first sets `trap '' TERM INT HUP USR1`, so a second signal cannot cut finalization short. A GitLab KILL still can; `after_script` is the backstop for that case.

**Provenance (r4, correction 1).** Every marker that explains *why* the run ended is recorded **in the parent shell**, before aggregation and before the final line. Nothing that records state runs inside `$( … )`.
- **Harness stopped by the finalizer.** If a harness is still running when `finalize` starts (an unexpected parent error during P2), `finalize` stops the group and records `ST[harness]=143` and `HARNESS_STOP=finalizer`.
- **`ST[internal]=19`** is recorded when the main body did not complete (`MAIN_DONE=0`), L0 succeeded, **and no signal was received**. That is an orchestration error.
  - It is not recorded for an external interruption, which `ST[signal]` records instead.
  - It is not recorded for a pre-location exit (`location=16`).
  - Any harness status is kept beside it, so precedence still reports the harness status.
- **`HARNESS_STOP`** is one of `none`, `deadline`, `signal` or `finalizer`.
- **P2's post-harness record.** After the harness wait returns, the parent writes `state/harness.result` (`status=`, `stop=`). This is a parent step: if it fails (errexit), the run is an orchestration error with `harness=<status>` **and** `internal=19` (model F15b).
- **`aggregate`** runs as a plain function in the parent. It reads `ST` and sets `FINAL`; it never mutates `ST`. r3's model called `final="$(aggregate)"`, so the `internal` marker set inside it was lost in the subshell (Codex's r3 finding).
- **The final line** is `validate: final status N; published=<…>; registry=<removed|kept:PATH|none>; harness_stop=<…>; phases: name=status …`. It includes `signal=` and `internal=` whenever they are recorded.

**Internal deadline value.** `BUDGET_S = min(CI_JOB_TIMEOUT, RUNNER_SCRIPT_TIMEOUT) − 1200 s`, written by P1 as a state value. Locally, with neither variable set, it is 14 400 s. P1 fails with 16 if it is below 3600 s.

**Run start.** `trap finalize EXIT` and the signal traps are installed immediately after argument parsing, before L0.

**Evidence (P-MODEL; r3 24 of 24, r4 29 of 29 with strengthened assertions, §15.5a and §15.5b).** These include:
- an external SIGTERM to the supervisor while a stub harness with grandchildren ran: exit 143; group gone; the harness's own cleanup ran; postcheck and cleanup ran;
- a TERM-ignoring harness: KILL after the grace; group gone; exit 143;
- the watchdog deadline: exit 124; group gone; harness cleanup ran;
- an unexpected parent failure while the harness ran: finalizer stopped the group; postcheck and cleanup ran; **r4:** the final line carries `harness=143`, `harness_stop=finalizer` **and** `internal=19` (F15);
- **r4:** an unexpected parent failure after a successful harness: exit 19, final line `harness=0` **and** `internal=19` (F15b);
- stragglers;
- three pre-location failures.

#### 11.2.2 Main body (in order)

| Step | Where | Work | Status on failure |
|---|---|---|---|
| L0 `location` | parent | **Containment before creation:**<br>• `REPO_R = realpath -e REPO`.<br>• With `--evidence-dir`: the target is `realpath -m DIR`. Reject it if it is `REPO_R` or below it (traversal and external or dangling symlinks resolve first), or if it already exists. Create it with `mkdir -m 700`.<br>• Without `--evidence-dir`: the **base** is `realpath -m "${TMPDIR:-/tmp}"`. Reject it if it lies inside the repository or is not a directory, **before** `mktemp -d "$base/g003-validate-XXXXXX"`.<br>• Re-resolve the created path with `realpath -e` and re-check containment.<br>• **Registry:** in CI, `REG=$(dirname "$CI_PROJECT_DIR")/.platform-validation-owned-$CI_JOB_ID`, resolved with `realpath -m`, checked to lie outside the repository, required not to exist, and created with `mkdir -m 700`. Locally, `REG=$EV/owned`.<br>• Create `raw/`, `state/` and `final/`.<br>• `--publish-dir` must be exactly `ci-evidence`.<br>If any check fails: `LOC_OK=0`, then `exit 16`, which leads to the **pre-location path** (§11.2.3). | 16 |
| P1 `preflight` | isolated + import | • `runner-preflight.sh` (§11.5) writes `raw/runner-preflight.tsv`; any FAIL row → 16;<br>• `docker pull` of the pinned gitleaks image (setup);<br>• the start Docker inventory with IDs (`$REG/start-inventory.ids`);<br>• attribution (§11.2.4, unchanged) → `raw/attribution.json`;<br>• precheck: `status --ignored` must be empty;<br>• write `state/p1.env`.<br>The parent then runs `state_import` on `state/p1.env`. | 16 (phase or import) |
| P2 `harness` | parent | Only if L0, P1 and the import passed. Supervised launch (§11.2.1):<br>`setsid bash infra/validation/run.sh --base HEAD --allow-file infra/ci/unchanged.allow --owned-registry "$REG/owned.tsv" --secrets-file "$REG/secrets" --report-dir "$EV/raw/report" all >"$EV/raw/harness-console.log" 2>&1 &`<br>**r4:** after the wait, the parent writes `state/harness.result`. A failure there is an orchestration error (internal=19), with the harness status kept. | the harness status (as recorded above) |

`MAIN_DONE=1` is set only after the main body's last step. An unexpected failure of a parent step (errexit) leaves it 0, and leads to `finalize`.

#### 11.2.3 Finalizer (always, once; parent orchestration with isolated bodies; order fixed)

**Pre-location path** (`LOC_OK=0`). No Docker operation and no evidence write has happened yet, because L0 is the first action.
- Only **F1** runs: the repository status check, read-only, with output to the job log only.
- F2 is recorded as `n/a (no registry)`; F3–F9 do not run.
- A directory that L0 itself created, and that it has proven to lie outside the repository, is removed. This includes a registry skeleton (job, token and repo rows only), because no Docker operation can have happened yet.
- **r4:** no `internal` marker is recorded on this path.
- **Published: none.** Exit 16.
- Model cases F20–F22 (TMPDIR inside the repository, `--evidence-dir` inside it, and a symlink into it): exit 16, nothing created inside the repository (whole-tree inventory unchanged), postcheck ran.

**Normal path** (`LOC_OK=1`). A harness still running is stopped first.

| Step | Work | Status on failure |
|---|---|---|
| F1 `postcheck` | The repository status must be empty, and `HEAD`, the tree and the status digest must equal the imported P1 values when present | 13 |
| F2 `cleanup` | `cleanup-owned.sh --registry "$REG" --mode finalize --out "$EV/raw/cleanup-inventory.tsv"` (§11.4: positive ownership; unproven resources kept and reported). **r4:** the inventory is also copied to `$REG/cleanup-inventory.tsv`. An inventory containing any row that is not `removed` makes F2 return 18, even if the tool exited 0. | 18 |
| F3 `raw_scan` | Over **all** of `$EV/raw`: literal scan against `$REG/secrets` (a missing or empty file when the harness ran is a scanner failure), gitleaks with `--exit-code 23`, and forbidden names. Any unreadable file is a scanner failure (model F11b). | 15 finding; 17 scanner failure |
| F4 `complete` | Only if P2 ran. Strict parse as in r2. **Also** writes `$EV/final/status-columns.tsv`, containing only lines matching `^[A-Z][A-Z0-9-]*\t(PASS\|FAIL\|UNMET\|INFO)$`. | 14 |
| F5 `stage` | Only if F3 returned 0. **All or nothing:** copy the allow-list into `$EV/staged.tmp`. On any copy error, delete `staged.tmp` entirely and return 19. Only when every copy succeeds, rename it to `$EV/staged`.<br>The allow-list is r2's list **without `harness-console.log`**, which is now genuinely excluded. `deps-codes.txt` is also excluded. | 19 |
| — candidate (parent) | **full** if F3 = 0 and F5 = 0;<br>**minimal** if F3 = 15 (marker `WITHHELD`) or if F5 failed (marker `STAGING-FAILED`);<br>**none** if F3 = 17.<br>After a preflight failure the raw set holds only preflight and attribution evidence. If that set is clean it is staged and published as **full**, as r2's group B expected (model F03). | — |
| F6 `summary` | Builds the candidate in a fresh `$EV/candidate`: full = staged files plus the summary; minimal = the marker plus the summary.<br>`summary.json` and `phase-status.tsv` contain only validated fields (enums, integers, hex SHAs, UTC times, allow-listed CI metadata; a malformed value fails the step). They are labelled `"status_kind": "pre-publication"`, record L0–F5, and list `pending: summary, candidate_scan, publish, report, scrub`.<br>On failure the candidate is deleted, and the outcome becomes **none**. | 19 |
| F7 `candidate_scan` | Scanners (a)–(c) over exactly `$EV/candidate` **and** `$EV/final/status-columns.tsv`.<br>• A finding in a **full** candidate demotes it to minimal: F6 is rebuilt as `WITHHELD`, then rescanned (`rescan`). If the rescan is not clean, the outcome becomes **none**.<br>• A finding in a minimal candidate, or any scanner failure, gives **none**. | 15 or 17 |
| F8 `publish` | CI only, and only if a candidate survived F7: copy to `$CI_PROJECT_DIR/ci-evidence.tmp`, then `mv` it to `ci-evidence`. On failure, remove `.tmp`; nothing is published. | 19 |
| F9 `report` | The job log gets `phase=… status=…` lines. Then, **only if F3 and F7 were clean**, the `ID STATUS` lines from the validated and scanned `$EV/final/status-columns.tsv`, which F10 has not yet removed. | 19 |
| F10 `scrub` | **r4 (correction 2; replaces r3's contradictory wording):**<br>1. **Always:** delete `$REG/secrets` and verify that it is gone (else 19). Delete `raw/`, `staged*`, `candidate/`, `final/` and `state/`.<br>2. **If F2 returned 0:** delete `$REG`. Then, in default-location mode, delete `$EV`.<br>3. **If F2 did not return 0, keep the registry:** `owned.tsv`, `start-inventory.ids`, `cid/`, `token` and `cleanup-inventory.tsv`.<br>&nbsp;&nbsp;– **CI layout:** `$REG` is a sibling of the checkout, outside `$EV`, so `$EV` is still deleted in default-location mode.<br>&nbsp;&nbsp;– **Local layout:** `$REG=$EV/owned`, so `$EV` is **kept**, reduced to `owned/`. A recursive deletion of `$EV` never runs while it holds a retained registry.<br>4. The parent then observes whether `$REG` exists and records `registry=removed` or `registry=kept:<path>`. | 19 |
| F11 `exit` | Aggregate in the parent (§11.2.5) and print the **authoritative final line**: `validate: final status N; published=<full\|minimal\|none>; registry=<removed\|kept:PATH\|none>; harness_stop=<none\|deadline\|signal\|finalizer>; phases: name=status …` (including `signal=` and `internal=` when recorded). Then exit N. | — |

**Publication outcomes** (exactly one per run; recorded in the final line):

| Outcome | When | Contents |
|---|---|---|
| **full** | F3 = 0, F5 = 0, F6 = 0, F7 = 0, F8 = 0 | the scanned staged allow-list, `summary.json` and `phase-status.tsv` |
| **minimal** | the candidate is minimal (or was demoted), F6 = 0, the minimal set's scan = 0, F8 = 0 | exactly the marker, `summary.json` and `phase-status.tsv` |
| **none** | any scanner failure (17), any F6 failure, any F8 failure, or the pre-location path | nothing: no `ci-evidence/` and no `.tmp` left behind |

There is **no partial set** (staging is all-or-nothing) and **no unscanned fallback** (every published byte passed F7).

#### 11.2.4 Checkout attribution (correction 2; documented variables only, F13)

**Unchanged from r2 (accepted by Codex).** The rules table, raw parent check, recorded metadata and the P-ATTR prototype (21 of 21 as expected) are in r2 §11.2.4, preserved in Appendix R2. It runs in P1, before the harness.

#### 11.2.5 Exit status (aggregation; structural no-masking invariant)

| Code | Meaning |
|---|---|
| 0 | every recorded step returned 0 (skipped steps are `skip` or `n/a`, and only as a consequence of another recorded nonzero status) **and** `MAIN_DONE=1` |
| harness status | P2 ran and the recorded status is nonzero (1 or 2 from `run.sh`; 124 at the deadline; 143, 130 or 129 on a signal; 143 when the finalizer stopped it after an orchestration error) |
| 128 + n (signal) | **r4:** a signal arrived while no harness was running (`ST[signal]`) |
| 13 / 14 / 15 / 16 / 17 / 18 | postcheck drift / evidence incomplete or malformed / unsafe evidence / location, preflight or state import / scanner failure / cleanup incomplete, refused or unproven |
| 19 | staging, summary, publication, report or scrub failure; an orchestration error (`internal=19`: `MAIN_DONE=0` with no signal and a valid location); or **any recorded nonzero status not otherwise selected** |

**Precedence** chooses which code is reported:
1. 16 for L0, P1 or the import;
2. otherwise the nonzero harness status;
3. otherwise the signal status (r4);
4. otherwise 17, then 15, 14, 13, 18.

Precedence picks only the exit code. **Every** recorded marker, including `internal=19` beside `harness=143`, appears in the final line (correction 1).

**Structural invariant:** after precedence, if any recorded status anywhere in the **single phase list** (which now includes `signal` and `internal`) is nonzero and the selected code is still 0, the exit is **19**. **r4:** `aggregate` runs in the parent and does not mutate `ST`; the markers it depends on are recorded beforehand. The aggregator iterates over the same `PHASES` array the finalizer records into, so a newly added step cannot be omitted. This fixes the r2 model defect (`stage=19` → exit 0): model case F11 now gives `stage=19` → **exit 19**, published=minimal, nothing partial.

**Authoritative status.** The artifact's `summary.json` and `phase-status.tsv` are **pre-publication metadata**. They record L0–F5 and list the later steps as pending. The **authoritative** result is the job's exit status (the GitLab job and pipeline status) together with the final job-log line from F11, which records every step, including F7–F10. The Result and Codex use those for H1 and H2.

### 11.3 CI controls (`infra/ci/validate-controls.sh`; scratch trees outside the checkout; no switch in `validate.sh`; r3; r2 version in Appendix R2)

**How the controls work.** As in r2:
- disposable `git init` trees containing copies of `validate.sh`, `lib.sh`, `cleanup-owned.sh`, `required-checks.tsv` and `unchanged.allow`;
- a **stub** `infra/validation/run.sh`, shaped like the real harness (`set -euo pipefail`, an EXIT-trap cleanup, and stdout and stderr through a TERM-ignoring process-substitution `tee`);
- a **stub** `runner-preflight.sh`;
- faults injected only through the stubs, a `docker` shim, file modes, GitLab-defined variables, and **signals sent from outside**.

`validate.sh` has no fault switch. Each control asserts:
- the exit status and the final line;
- the stub invocation count;
- the publication outcome and its exact contents;
- canary absence;
- that F1 and F2 ran;
- for supervision controls, that every PID recorded by the stub is gone and that `kill -0 -- -PGID` fails.

The r3 planning model (§15.5a, P-MODEL) runs the same matrix against a stand-in. The model cases are named in brackets.

**Group A — gate semantics** (r1 and r2, retained):

| Control | Scenario | Expected |
|---|---|---|
| V1 | All valid, complete PASS [F01] | 0; **full** |
| V2 **(a)** | One FAIL row; harness exit 1 [F02] | 1; full (scanned reports) |
| V3 (a) | Harness 0, but a FAIL row | 14 |
| V4 **(b)** | Real `run.sh … contracts` on a scratch copy with EI39 flipped; its `results.tsv` replayed through V2's path | 1 |
| V7 **(d)** | Harness 1, clean tree | 1; postcheck passed |
| V8 / V9 **(e)** | Harness 0; a tracked file modified, or an untracked file created [F09] | 13 |
| V10 / V11 **(f)** | A required ID missing; an UNMET row; browser JSON with `skipped: 1` | 14 |
| V12 | Harness 1 **and** drift | 1, with the drift reported |

**Group B — attribution:** unchanged from r2 (P1–P5, P3F, N1–N16; accepted).

**Group C — evidence scanning** (correction 2):

| Control | Scenario | Expected |
|---|---|---|
| V5a | Generated canary in a **retained** raw report [F06] | 15; **minimal** (`WITHHELD`); canary absent from the published set |
| V5b-planted | The canary only in `harness-console.log`, which is **excluded** from staging in r3 [F07] | 15; **minimal**; canary absent from the published set |
| V5b-clean | No canary; `harness-console.log` present in the raw set [F07b] | 0; **full**; `harness-console.log` **absent** from the published set. Proves the exclusion itself. |
| V5c | A gitleaks-detectable token generated at run time, not in the secrets file [F08] | 15; minimal |
| V6a | Scanner failure: the gitleaks shim exits 1 [F05] | 17; **none** |
| V6b | The harness ran, but the secrets file is missing | 17; none |
| V6c | Positive scanner control: clean raw and candidate sets (identical to V1) [F01] | 0; full |
| V6d | An unreadable raw file, so the literal scan returns 2 [F11b] | 17; none |

**Group D — finalization, supervision and location** (corrections 1 and 2):

| Control | Scenario | Expected |
|---|---|---|
| V14 / V15 | Malformed `results.tsv` [F04]; malformed browser JSON | 14; full; F1 and F2 ran |
| **V16** | **Staging after a partial copy:** the first allow-listed files copy, then a later one fails (it is a directory) [F11] | **19**; **minimal** (`STAGING-FAILED`); none of the partially copied files is published |
| **V16b** | Summary failure: a malformed allow-listed CI value [F12] | 19; **none** |
| V17 | Publication failure: the publish parent is not writable [F13] | 19; none; no `ci-evidence/` or `.tmp` left behind |
| **V17b** | Scrub failure: an undeletable raw subdirectory [F14] | **19**; full already published; the final line reports `scrub=19` |
| V18 | **External SIGTERM to `validate.sh`** while the stub harness and its grandchildren run [F16] | 143; the harness group is gone (every recorded PID is dead, `kill -0 -- -PGID` fails); the stub's cleanup trap ran; F1 and F2 ran; full (scanned). **r4:** the final line has `signal=143`, `harness=143` and `harness_stop=signal`, and **no** `internal=` |
| V18a | A TERM-ignoring stub harness and SIGTERM to `validate.sh` [F17] | 143; KILL after the grace; group gone; F1 and F2 ran |
| V18b | Harness 1 and a cleanup failure | 1; `cleanup=18` reported |
| V18c | The stub preflight reports a FAIL [F03] | 16; harness invocations 0; preflight evidence published as full |
| V18d | **Watchdog deadline** with the stub harness sleeping [F18] | 124; group gone; stub cleanup ran; final line `harness_stop=deadline`, no `internal=` |
| V18e | A malformed state file (unknown, duplicate, missing or bad value; a symlink) | 16; harness invocations 0 (P-IMPORT unit cases) |
| V18f | **Unexpected parent failure while the harness runs** (the parent's state write fails) [F15] | Harness group stopped by the finalizer; exit 143 by precedence; F1 and F2 ran. **r4:** the authoritative final line must contain **both** `harness=143` **and** `internal=19`, plus `harness_stop=finalizer`, and no `signal=`. An orchestration failure is never reported as only an external interruption. |
| **V18k** | **r4: unexpected parent failure after a successful harness** (the parent's `state/harness.result` write fails) [F15b] | Exit **19**; final line `harness=0` **and** `internal=19`, no `signal=`; F1–F11 ran; full (scanned) |
| V18g / V18h / V18i | **Pre-location:** `TMPDIR` inside the repository; `--evidence-dir` inside it; `--evidence-dir` through an external symlink into it [F20–F22] | 16; **nothing created** in the repository (whole-tree inventory unchanged); F1 only; published none; no `internal=` |
| V18j | Stragglers left in the harness group after a normal exit [F19] | 0; stragglers stopped; group gone |

**Group E — owned-resource cleanup** (correction 3). Docker-real controls with `busybox:1.37.0` and labels `g003.control=<token>`. The control resources are created **after** the control's start inventory. P-OWN models the decisions (§15.5a).

| Control | Scenario | Expected |
|---|---|---|
| **V19** | Owned: a container under a registered project whose `working_dir` is under `work_root`, attached to a labelled network, mounting a labelled volume, and running an image the control built and registered with an `image <tag> <id>` row; capture rows written as the harness would | All four removed; inventory `removed=4 remaining=0`; exit 0 |
| **V20c** | Unrelated **container**: registered project label, foreign `working_dir`, created after the start inventory | **Survives**; `refused`; exit 18 |
| **V20n** | Unrelated **network**: registered project label, created after the start inventory, attached to no owned container, no capture row | **Survives**; `unproven`; exit 18 |
| **V20v** | Unrelated **volume**: same conditions | **Survives**; `unproven`; exit 18 |
| **V20i** | Unrelated **image**: tagged `<registered-project>-intruder:latest`, new ID, no build or capture row | **Survives**; `unproven`; exit 18 |
| V21 | Resources of an unregistered project | Untouched; absent from the inventory |
| V22 | Removal failure (shim: `docker rm` exits 1) | 18; `remaining` reported |
| V23 | Registry tampering (symlink, token mismatch, job mismatch) | Nothing removed; 18 (`--after-script`: 0, with the refusal printed) |
| V24 | `after_script` hand-off in a separate `env -i` shell | Finds the registry by the documented rule; removes only positively owned resources; exits 0 |
| V25 | Root `.env`, matching and mismatched hash | Removed only on a match |
| V26 | A resource in the start inventory under a registered label | Untouched |
| **V27** | **r4, interrupted standalone, positive:** the stub harness starts a detached standalone container through `docker_run_owned` (cidfile and `standalone` row written), then is SIGKILLed before its own cleanup | `cleanup-owned.sh` removes it (cidfile row); inventory `removed`; exit 0 for cleanup (the run is nonzero from the harness status) |
| **V27b** | As V27, but killed after the cidfile is written and **before** the `standalone` row is appended | Removed through the cidfile in `$REG/cid/` (written by the Docker CLI into the run's mode-700 registry) |
| **V28** | **Interrupted preflight:** `validate.sh` is killed (SIGTERM) after RP6's httpd container is created | `finalize` (or, after a SIGKILL, `after_script`) removes it through its cidfile row |
| **V29** | **Unrelated label-only, negative:** a container created after the start inventory carrying `g001.run=<the run's RUN_ID>`, with no cidfile or row | **Survives**; `unproven`; exit 18 |
| **V30** | **Unrelated label-only, negative:** the same with `platform.validation.preflight=<the run's token>` | **Survives**; `unproven`; exit 18 |
| **V31** | **Ephemeral foreground `--rm` container** still running at cleanup, which finishes within the bounded wait | Gone through daemon auto-removal; not reported; exit 0 |
| **V31b** | An ephemeral run-labelled container still running after the bounded wait | **Kept** (never killed by label); `unproven`; exit 18 |
| **V32** | **Registry retention, local layout:** cleanup fails (shim) [F10] | Exit 18; final line `registry=kept:$EV/owned`; `owned.tsv`, `start-inventory.ids`, `cid/`, `token` and `cleanup-inventory.tsv` (with a non-`removed` row) survive with mode 700; **`secrets` is gone**; `$EV` reduced to `owned/` |
| **V33** | **Registry retention, CI layout:** as V32 with the checkout-sibling registry [F10c] | Exit 18; the sibling registry is kept as V32; `$EV` deleted; secrets gone |
| **V34** | Harness exit 1 **and** cleanup failure, local and CI layouts [F10h, F10hc] | Exit **1** (harness precedence); `cleanup=18` in the final line; the registry is kept as V32/V33 |
| **V35** | Successful cleanup, local and CI layouts [F01, F01c] | `registry=removed`; neither the registry nor the default-location `$EV` remains |

Every V20 control removes its own unrelated resource afterwards and verifies the removal. No case can turn a failure into exit 0. W1–W9 and C1–C12, C8b and C8c are unchanged.

**Control count:**
- group A: 10;
- group B: 22;
- group C: 8;
- group D: 18 (V14, V15, V16, V16b, V17, V17b, V18, V18a–V18k);
- group E: 22 (V19, V20c, V20n, V20v, V20i, V21–V26, V27, V27b, V28–V31, V31b, V32–V35).

**Total: 80 V-controls** (r3: 68; r2: 53), plus W1–W9 and C1–C12, C8b and C8c.

### 11.4 Owned-resource registry and `infra/ci/cleanup-owned.sh` (correction 3; r3; r2 version in Appendix R2)

**Location, permissions and modes** are unchanged from r2:
- CI registry: `$(dirname "$CI_PROJECT_DIR")/.platform-validation-owned-$CI_JOB_ID`, created by L0 (§11.2.2); local registry: `$EV/owned`;
- mode 700, never a symlink;
- `--mode finalize` (gating, 0 or 18) and `--after-script` (always 0).

**Rows** (`owned.tsv`, tab-separated; r3 additions in bold):

```text
#platform-validation-owned v1
job        <CI_JOB_ID | local-<random>>
token      <32 hex>
repo       <absolute checkout path>
work_root  <the harness's absolute $WORK>
label      g001.run=<RUN_ID>   |   platform.validation.preflight=<token>
project    <compose project name>
rootenv    <sha256>
image      <repository:tag> <image ID>   (written immediately after the run builds that tag)
standalone <full ID> <purpose>           (r4: from the --cidfile the Docker CLI wrote into $REG/cid/ at creation)
container  <full ID>                     (capture row)
network    <full ID>                     (capture row)
volume     <name>                        (capture row)
```

**Capture points (positive evidence written while ownership is provable).** `owned_capture DEST` runs:
- in `run.sh`, after every start or `up` of a registered project (the README and GETTING-STARTED start blocks, every `compose … up`, `compose_browser … run`);
- first in the harness's cleanup trap;
- in `cleanup-owned.sh` itself, before any removal.

For each container of the project that is **owned under the container rule below**, it records:
- the container's full ID;
- every network ID it is attached to (`docker inspect -f '{{range $k,$v := .NetworkSettings.Networks}}{{$v.NetworkID}} {{end}}'`);
- every named volume it mounts (`{{range .Mounts}}{{if eq .Type "volume"}}{{.Name}} {{end}}{{end}}`);
- its image tag and ID (`{{.Config.Image}}`, `{{.Image}}`).

Rows are append-only. A network, volume or image ID that is in the start inventory is never recorded.

**Standalone capture (r4, correction 2).** Containers created with plain `docker run` (the harness's probes and check containers, the G003 error-output containers, RP6's preflight httpd) carry **no** Compose labels, so the Compose `working_dir` rule cannot apply to them. Their positive binding is made **at creation**:
- `docker_run_owned` passes `--cidfile "$REG/cid/<seq>.cid"`. The Docker CLI writes the new container's full ID into that file inside the run's mode-700 registry.
- The helper then appends `standalone <id> <purpose>`. If the run is killed between those two steps, the cidfile alone still binds the ID (V27b).
- A standalone container's **image** is never removed because the container used it. The image rule (build or capture record) still decides; pre-existing images such as `busybox:1.37.0` stay.
- Standalone containers create no networks or volumes. They attach to existing networks or `none`, and `docker rm -v` removes their anonymous volumes.

**Ownership rules** (removal needs all listed conditions):

| Kind | Removed only if | Otherwise |
|---|---|---|
| **Compose container** | its ID is absent from the start inventory; **and** its `com.docker.compose.project` label is a registered project; **and** its `com.docker.compose.project.working_dir` equals `repo` or lies under `work_root` | registered project label but the conditions fail → `refused` (kept; 18); unregistered → ignored |
| **standalone container (r4)** | its ID is absent from the start inventory; **and** a `standalone <id>` row or a `$REG/cid/*.cid` file of this registry names that exact ID | a registered **`label` row** match (`g001.run=…` or `platform.validation.preflight=…`) without that binding → `unproven` (kept; 18). **Labels are only used to find candidates for the report, never as deletion authority** (V29, V30). |
| **ephemeral foreground container (r4)** | never removed by cleanup. Short `docker run --rm` tool containers are removed by the daemon when their command exits. | Cleanup waits up to 60 s for any still-running unregistered container carrying a registered label to disappear. Any that remain → `unproven` (kept; 18) (V31, V31b). |
| network | its ID is absent from the start inventory; **and** a `network` capture row names that ID, or an owned container is attached to it right now | carries a registered project label without that evidence → `unproven` (kept; 18) |
| volume | its name is absent from the start inventory; **and** a `volume` capture row names it, or an owned container mounts it right now | registered label without that evidence → `unproven` (kept; 18) |
| image | its ID is absent from the start inventory; **and** an `image <tag> <id>` row (build record or capture) matches **both** the tag and the ID the tag resolves to now | a tag with a registered project's prefix (`<project>-*`) without a matching row → `unproven` (kept; 18). **The name pattern alone never authorizes removal**; it only flags the image for the report. r2's "`<project>-<service>` absent at start" fallback is withdrawn. |
| root `.env` | its SHA-256 equals the `rootenv` row | `refused` |

- **Removal order and verification:** containers (`docker rm -f -v ID`), then networks (`docker network rm ID`), volumes (`docker volume rm NAME`) and images (`docker image rm ID`, only after checking that the tag still resolves to it). Every query is then rerun.
- **Exit 0** only if no owned resource remains, nothing was refused or unproven, and every Docker command succeeded. Otherwise **18**, with `remaining`, `refused`, `unproven` and `failed` rows.
- **Never:** any `prune`; removal by name or label alone; any resource in the start inventory; any resource without positive ownership.
- **Registry retention (r4, correction 2).** If cleanup does not verify (exit ≠ 0), the registry evidence is retained for both layouts: `owned.tsv`, `start-inventory.ids`, `cid/`, `token` and `cleanup-inventory.tsv`. The secret list is always removed (§11.2.3 F10).
  - **CI layout:** the registry is a checkout sibling. `after_script` may retry, and keeps it again if it still cannot verify.
  - **Local layout:** `$EV/owned` survives inside a kept `$EV`.
  - **The inventory is never reported empty when anything was refused, unproven, remaining or failed.** F2 returns 18 whenever a non-`removed` row exists (model F10, F10c, F10h, F10hc).
- **Consequence under signals.** If the harness is killed after creating something but before the next capture point, that resource is `unproven`. It is kept and reported (18), not guessed. The run is already failing in that case. The report lists it for owner cleanup, never deletion by inference.

### 11.5 `infra/ci/runner-preflight.sh` (r3; unchanged from r2 except the tool list)

As r2 §11.5 (RP1–RP13). **r3:** RP2 also requires `setsid` (util-linux) and `ps`. GNU `timeout` is still required by the harness's own probes; `validate.sh` no longer uses it.

**r4:** RP6's detached httpd container and any other preflight container are started through the same cidfile binding (`--cidfile "$REG/cid/<seq>.cid"`, then a `standalone <id> preflight-*` row). They are removed by the standalone rule, never by the preflight label (V28, V30).

## 12. Runner profile and hosted evidence (blocking prerequisites Q1 and Q2; R6/R7, AC7; r2; r1 §12 in Appendix R1)

### 12.1 What the harness requires (from F9)

**Profile A (required unless Codex and the owner decide otherwise).** A self-managed GitLab runner using the **shell executor**, on a **dedicated, non-production Linux host**:

- **Docker:** Docker Engine 24 or later with BuildKit, and the Compose v2 plugin 2.24 or later.
- **Host tools:** bash 4.4 or later (`inherit_errexit`), GNU coreutils/findutils/grep with `-P`, git, `timeout`, `hostname -I` or `ip`, and bash `/dev/tcp`.
- **Job user:** the job runs as the runner user in the `docker` group. That is root-equivalent on this host. It is acceptable only because the host is dedicated and disposable in purpose, and it must be acknowledged.
- **Isolation:** no other workload's Docker resources on that daemon.
- **Concurrency:** one concurrent job for this project. `resource_group` enforces it inside the project; the runner `limit`/`concurrent` value, or dedication, covers other projects.
- **Capacity:** at least 4 vCPU (8 recommended), at least 8 GiB RAM (16 recommended), and at least 40 GiB free for images and the BuildKit cache.
- **Timeouts:** the runner's `maximum_timeout` and the project's CI timeout must each be at least 3 h.
- **Egress** to the registries and module proxies used in builds: Docker Hub, `gcr.io`, `ghcr.io`, `mcr.microsoft.com`, `proxy.golang.org`/`sum.golang.org`, `registry.npmjs.org` and PyPI. The Docker Hub anonymous pull limit must be acknowledged.

**Profile A2 (acceptable only if already configured this way; no configuration change is made by G003).** A Docker executor that mounts the host Docker socket, with job-container `network_mode = "host"`, and the builds and temporary directories mounted at **identical host paths**. Without all three, the bind-mount and loopback probes cannot work.

**Not acceptable without an explicit security decision (stop condition).**
- GitLab.com hosted runners or any `docker:dind` service. These need `privileged = true`, and published ports would not be on the job's loopback.
- A socket-only Docker executor without host networking and path parity.
- Any runner that can reach production resources or credentials.

**r2 notes on §12.1.**
- **Nothing here describes the actual runner.** The profiles state requirements, not facts. No runner identity, tag, executor, capacity or permission is assumed anywhere in this Plan.
- **The registry location (§11.4) relies on the same path parity.** Profile A2 already requires the builds and temporary directories at identical host paths.
- **`runner-preflight.sh` (§11.5) turns each requirement into an executable check,** run in every pipeline and in R1.

### 12.2 Owner-supplied runner facts (Q1: **UNRESOLVED; blocking**)

The owner has confirmed only that an existing Docker-capable GitLab runner is to be used (Goal, Dependencies). Everything below is **unknown** and must be supplied by the owner, or by someone they authorize.

**Rules for supplying them:**
- Supply them as plain text in the Codex approval handoff, or in a file outside the repository whose path is given to Codex.
- Do **not** paste `config.toml` wholesale: it contains the runner token. Copy only the named fields.
- Codex checks the facts against Profile A or A2 **before approving execution**.
- R1 (§12.4) then verifies them on the runner **before S1**.
- Facts are never a substitute for R1, and R1 is never a substitute for the facts. If any fact is unavailable, execution stays blocked.

| ID | Fact the owner supplies | Where to read it | Requirement it is checked against | Status |
|---|---|---|---|---|
| F-R1 | Runner ID and description; scope (project, group or instance); self-managed or GitLab-hosted | GitLab: project → Settings → CI/CD → Runners | Self-managed (Profile A or A2). GitLab-hosted runners are a stop condition. | UNRESOLVED |
| F-R2 | The tag to use; whether it runs untagged jobs; whether it is limited to protected branches; whether it is locked to the project | same page | One tag for `.gitlab-ci.yml`. If protected-only, the evidence branches must be protected (owner action), or another runner is needed. | UNRESOLVED |
| F-R3 | **Executor** type, and for Docker executors: `privileged`, the Docker socket mount, `network_mode`, `volumes` and the builds path | runner host: the `[[runners]]` section of `config.toml` (named fields only) | Shell executor (A), or a Docker executor with socket, host network and identical paths (A2). `privileged = true` or dind is a stop condition. | UNRESOLVED |
| F-R4 | Runner version (`gitlab-runner --version`) | runner host | Records whether `RUNNER_SCRIPT_TIMEOUT` and `RUNNER_AFTER_SCRIPT_TIMEOUT` are honoured. The internal deadline does not depend on it (§11.2.1). | UNRESOLVED |
| F-R5 | Host OS and architecture; Docker Engine, Compose plugin and Bash versions; CPUs; RAM; free disk for the Docker root and the temporary directory | runner host (`uname -m`, `docker version`, `docker compose version`, `bash --version`, `nproc`, `free -g`, `df -h`) | linux/amd64; Engine ≥ 24; Compose ≥ 2.24; Bash ≥ 4.4; ≥ 4 vCPU, ≥ 8 GiB RAM, ≥ 40 GiB free | UNRESOLVED |
| F-R6 | Isolation: is the host dedicated? Other workloads or projects on the same Docker daemon; the runner's `concurrent` and `limit` | owner knowledge, `config.toml` | Dedicated, non-production, no production credentials or networks reachable; one concurrent job for this project | UNRESOLVED |
| F-R7 | Timeouts: the runner's `maximum_timeout`; the project's CI/CD general timeout | `config.toml`; project → Settings → CI/CD → General pipelines | Each ≥ 3 h (the effective `CI_JOB_TIMEOUT` is checked again by RP12) | UNRESOLVED |
| F-R8 | Network egress from the host and job; any proxy or pull-through cache; Docker Hub authentication or rate-limit status | owner and network knowledge | Registries listed in §12.1 reachable; the anonymous pull limit acknowledged | UNRESOLVED |
| F-R9 | Host port use: is anything bound to the default ports (3100, 8180, 8100, 5532, 6479, 9100, 9101) or to the variant sets (§6.3 and `run.sh` `PORTS_*`)? | runner host (`ss -ltn`) | All free | UNRESOLVED |
| F-R10 | GitLab tier of the namespace; whether merged-results pipelines are enabled; whether merge trains are enabled | group → Settings → Billing; project → Settings → Merge requests | Merged results are Premium/Ultimate (F13). Merge trains are unsupported, so attribution would return 16 (Q17). | UNRESOLVED |
| F-R11 | Artifacts: the instance maximum artifact size; the project's "Keep artifacts from most recent successful jobs" setting; who holds Developer or higher access | Admin or project settings; project → Members | `access: developer` and 90-day expiry acceptable (Q7); expected evidence size well below the maximum (measured at R1 and H0) | UNRESOLVED |
| F-R12 | Evidence retrieval mechanism (paired with A-E, §12.3) | owner decision | Owner download to a path outside the repository, or a short-lived `read_api` token supplied out of band | UNRESOLVED |
| F-R13 | Written acknowledgement that the job user's `docker` group membership is root-equivalent on that host | owner | Required for Profile A | UNRESOLVED |

### 12.3 Hosted actions requiring explicit authorization (Q2: **UNRESOLVED; blocking**)

Neither this Plan, nor its approval, nor the planning delegation authorizes any of these. Each needs the owner's or Codex's explicit, named authorization before it is taken.
- **Commits.** Creating the commits these actions push is part of each authorization. They use the repository's configured Git identity unless the authorization says otherwise.
- **Ancestors.** A push publishes every ancestor commit not yet on `origin`. If BASE includes a not-yet-pushed commit (for example the skill consolidation under route A, Q14), the first push publishes it on that branch. The authorization must cover this.

| ID | Action | Purpose | Undo |
|---|---|---|---|
| **A-R1** | From a disposable clone **outside the repository**: create branch `g003-runner-probe` at BASE plus **one** probe commit (§12.4), push it to `origin`, and let its branch pipeline run on the Q1 runner | Runner qualification before S1 | Delete the branch after the evidence is captured (part of A-R1) |
| A-H0 (optional, recommended) | Push the work-in-progress commit(s) to `g003-foundation-observability-ci` and open a **draft** MR to `staging` | Early run of the G003 job itself on the runner, after S8 | The MR stays open for H1 |
| **A-H1** | Push the final implementation commit, and the Result when available, to the same branch; its MR pipeline runs | Hosted pass on the exact submitted commit and tree | Kept for review |
| **A-H2** | Push branch `g003-gate-proof` = H1 plus one commit setting EI39 to `valid`; open a draft MR to `staging` | Hosted failure through the reviewed gate | A-H3 |
| **A-H3** | After Codex has captured H2: close its MR and delete `g003-gate-proof` | Clean-up | — |
| **A-E** | Evidence retrieval for R1, H0, H1 and H2 (logs, artifacts, pipeline and job metadata) through the F-R12 mechanism | Attributable hosted evidence | Downloads are kept outside the repository; a token is never printed or stored |

**No action pushes to `staging` or `main`, merges, publishes, deploys, or creates or changes CI/CD variables, runners or project settings.**

### 12.4 R1: runner qualification before S1 (r3; r2 version in Appendix R2)

**When.** Immediately after S0 passes, which needs Q9 and Q14 resolved, and the Plan is approved, and only with A-R1 and A-E authorized. It comes **before any repository write** (S1). It makes **no repository write**: everything happens in `$EXT/r1/probe`, a `git clone` of the local repository at BASE. Cloning a local repository reads its objects and writes nothing to it.

**Probe commit** (four files; they exist only in the disposable clone and on the deleted branch):
1. `.gitlab-ci.yml` (probe-only):

   ```yaml
   workflow:
     rules:
       - if: $CI_PIPELINE_SOURCE == "push" && $CI_COMMIT_BRANCH == "g003-runner-probe"
       - when: never
   runner-probe:
     tags: [<RUNNER_TAG>]
     timeout: 3h
     resource_group: platform-validation-host
     variables:
       GIT_STRATEGY: clone
       GIT_CLEAN_FLAGS: -ffdx
     script:
       - bash infra/ci/runner-probe-job.sh
   ```

2. `infra/ci/runner-preflight.sh`: the §11.5 script, as written for S6. Any later change is recorded in the Result.
3. `infra/ci/runner-probe-job.sh`. Under an EXIT trap, it:
   - runs the preflight, with its TSV printed to the job log (no secrets: tool versions, counts and PASS/FAIL);
   - records `docker system df` before and after, and the UTC start and end;
   - runs the **unchanged accepted baseline harness** as its own `setsid` process group, under the same parent-owned watchdog and stop sequence as `validate.sh` (§11.2.1; deadline 9900 s, then TERM to the group, a 150 s grace, then KILL). The command is `setsid bash infra/validation/run.sh --base HEAD --allow-file infra/ci/probe.allow --report-dir "$T/report" all >"$T/console.log" 2>&1 &`, with `$T` a `mktemp -d` outside the checkout;
   - **r3 caveat (F26):** the baseline harness at BASE still has the as-is `tee`. If R1 hits its deadline, the harness's own cleanup may not run. The probe job then lists, by name only, any resource whose Compose project or label matches the run's `RUN_ID` (parsed from the console's `run <id>` line), and reports it for owner cleanup. It never deletes by inference. A deadline hit already fails the decision rule;
   - prints only the `ID<TAB>STATUS` columns of `results.tsv`, the leftover-inventory **count** and the durations;
   - deletes `$T`;
   - exits with the harness status.

   It uploads no artifacts.
4. `infra/ci/probe.allow`: comments only. `--base HEAD` is the probe commit, so B-SCOPE and B-PROTECTED see no change.

**What R1 establishes (before any G003 cycle):**
- runner selection by tag;
- executor compatibility: loopback parity, bind-mount parity, port availability, non-loopback IPv4 (RP6–RP9, and the real B-PORTS-* and B-BOUNDARIES-*);
- cold image builds, including the MinIO source build and the pinned images, and registry egress;
- the browser matrices on the runner;
- the cold `all` duration against the time budget;
- disk use;
- the harness's own cleanup on that daemon.

**Decision rule** (reported to Codex before S1, and recorded in the Result):
- **Proceed to S1** only if:
  - all 50 baseline required checks PASS;
  - every RP check passes;
  - the leftover inventory is empty, apart from the renderer cache volume;
  - the cold `all` duration is **≤ 90 minutes**. The G003 harness budget is 145 minutes (2h45m − 20m), and §19 estimates that G003's contracts suite, error-output checks and observability matrix add about 25–35 minutes cold. 90 minutes keeps that estimate inside the budget with margin.
- **Otherwise stop before S1.** Return the evidence to Codex and the owner. A runner reconfiguration, a timeout change or a split (§19) is their decision. Because no implementation exists yet, **no B7 cycle is consumed**. A rerun after an owner-side change also consumes none.
- **If A-R1 is not authorized,** the B7 forecast is not credible. The Plan recommends not starting S1, and Codex decides (Q10, §19).

### 12.5 Hosted evidence plan (H0–H3; each needs its authorization from §12.3)

| Step | When | Expected evidence |
|---|---|---|
| H0 (optional, recommended) | After S8 (local CI parity passes) | The first run of the G003 job on the runner, cache warm from R1. It records the pipeline, job and runner IDs, `tested_content`, the duration, the artifact size and the published set. A failure caused by the G003 changes counts in B7; one caused by the runner environment is reported as in §12.4. |
| H1 | After S10 | A passing pipeline on the **exact submitted commit and tree**. Recorded: pipeline ID, job ID, runner, `tested_content` (`mr-source`, or `mr-merged-result` if enabled), the duration, the **authoritative** job status and final log line (§11.2.5), the pre-publication `summary.json` with 69 PASS, and the downloaded artifact's SHA-256. |
| H2 | After H1 | A **failed** pipeline: B-CONTRACT-GO, B-CONTRACT-PY and B-CONTRACT-AGREE FAIL, and `validate.sh` exits 1. The scanned artifacts show the reason. This exercises the reviewed gate, not an `exit 1` job. |
| H3 | After Codex captures H2 | The H2 MR is closed and `g003-gate-proof` is deleted; recorded |

**Evidence classification** is unchanged from r1. Local and hosted evidence are kept apart in the Result, and **local parity never substitutes for H1 or H2**. R1 is runner-qualification evidence for the unchanged baseline, not G003 acceptance evidence.

### 12.6 Order of prerequisites (r3, correction 5)

Nothing in this Plan supplies or substitutes for these. R1 and the profiles in §12.1 verify facts; they do not provide them, and they authorize nothing.

1. **Q1, owner:** the runner facts F-R1–F-R13 (§12.2). Status: **UNRESOLVED.**
2. **Q2, owner or Codex:** named authorization of A-R1, A-H0 (optional), A-H1, A-H2, A-H3 and A-E (§12.3), including commits and the publication of any unpushed ancestor. Status: **UNRESOLVED.**
3. **Codex technical review** of this Plan, using the Q1 facts to confirm Profile A or A2. Claude does not seek **execution** approval before items 1 and 2 are supplied.
4. **Q14, owner:** baseline route. Under route A, commit the completed consolidation. Status: **OPEN** (consolidation complete, uncommitted).
5. **Q9, Codex:** the fresh binding (BASE, snapshot, its SHA-256, COUNT) and explicit execution approval. Status: **PENDING** on items 3 and 4.
6. Then S0 (§15.1), then R1/S0R (§12.4), before any implementation write.

If any item is unavailable, G003 stays blocked (Goal stop condition). Codex may instead consider the split in §19.

## 13. Dependencies, images and provenance

| Item | Pin | Where | Notes |
|---|---|---|---|
| `github.com/santhosh-tekuri/jsonschema/v6` | `v6.0.3` (2026-06-28, Apache-2.0) | `infra/validation/tools/go.mod` (new) | Test tooling only. `go.sum` is recorded; `go mod verify` runs in every build. |
| `golang.org/x/text` | resolved by MVS (expected `v0.14.0`) | tools `go.mod`, indirect | Recorded in the Result |
| `golang:1.27.1-alpine`, `gcr.io/distroless/static-debian12:nonroot` | existing tags | `validation-tools.Dockerfile` | Same tag-only pins as the existing images (inherited residual). No new base image. |
| `zricethezav/gitleaks:v8.30.1` | existing | `validate.sh` artifact scan | Already used by `secret-scan` |
| `koalaman/shellcheck:v0.11.0`, `hadolint/hadolint:v2.15.1`, `mikefarah/yq:4.48.1`, `busybox:1.37.0`, Playwright `v1.63.0-noble@sha256:eff16c30…a4a27` | existing | harness | Unchanged |

**No change** to:
- web `package.json` dependencies or `package-lock.json`;
- core `go.mod`/`go.sum`;
- intelligence `pyproject.toml` dependencies or `uv.lock`;
- any runtime base image.

**No new vendor, service or infrastructure.** santhosh is a validation library used by a test tool, analogous to the Playwright image accepted in G002's Q4 (Q12).

## 14. Files (exact; `management/v0/results/G003-scope-allowlist.txt` mirrors this section; r3; r2 version in Appendix R2, r1 in Appendix R1)

### 14.1 Database, schema and contracts

- **None** of: migrations, seeds, tables, domain identities, public routes, API or MCP contracts. B-MIGRATE-SET stays `{00001_baseline.sql}`.
- The two schemas are byte-identical: both validators assert the §2 digests, and the SD controls prove that enforcement.
- Response headers are additive only:
  - `X-Request-Id` on the web and intelligence health routes;
  - `Cache-Control: no-store` on intelligence health;
  - on an unexpected web health-handler exception, the 500 with an empty body that Next already sends, plus those two headers (Q15).

### 14.2 Customized rendered files (11 new register rows → 19 total)

The r1 ten, plus `core/cmd/server/main.go`:

| Path | Register reason |
|---|---|
| `core/internal/observability/correlation.go` | diagnostic context profile, per-hop spans and completion records (G003) |
| `core/internal/observability/logging.go` | JSON record envelope (event, service, UTC time) on the redacting logger (G003) |
| `core/internal/health/checks.go` | propagates diagnostic context on the optional intelligence probe (G003) |
| **`core/cmd/server/main.go`** | **bounded JSON exit record, classified net/http error log, and bind-then-log startup record with port and app_env only (G003; three reviewed hunks; SHA-256 `987c9335…51ea`)** |
| `web/app/api/health/route.ts` | request context and completion record on liveness (G003) |
| `web/app/api/ready/route.ts` | request context, propagation and completion record on readiness (G003) |
| `web/lib/core-client.ts` | forwards diagnostic context to core readiness (G003) |
| `intelligence/src/main.py` | health completion middleware, no-store health responses and bounded startup failure (G003) |
| `intelligence/src/observability/export.py` | JSON envelope, redaction, bounded framework, library and warning records, and framework-logger takeover (G003) |
| `intelligence/src/streaming/schemas.py` | Draft 2020-12 conformance: ECMA-262 pattern anchoring, date-time assertion, keyword audit (G003) |
| `infra/docker/intelligence.Dockerfile` | conformance corpus in the check stage only (G003) |

**Already registered, content changes:** `web/package.json` (scripts only) and `docs/foundation/FOLLOW-UP.md`.

### 14.3 New files (175)

| Area | Count | Paths |
|---|---|---|
| Core | 6 | `core/internal/observability/{context_test.go,exit.go,exit_test.go,errorlog.go,errorlog_test.go}`, `core/internal/health/propagation_test.go` |
| Web | 4 | `web/lib/request-context.ts`, `web/tests/unit/request-context.test.mjs`, `web/tests/fixtures/error-probe/{wrapped-route.ts,control-route.ts}` |
| Intelligence | 6 | `intelligence/src/observability/http.py`, `intelligence/tests/{test_health_context.py,test_logging_privacy.py,test_error_output.py,test_conformance.py,conformance_exchange.py}` |
| Contracts | 128 | as in r1: `contracts/conformance/README.md`, `contracts/conformance/intelligence/expectations.json`, and 126 fixtures (EV01–EV14, EI01–EI50, XV01–XV15, XI01–XI34, NV01–NV02, NI01–NI05, SN02–SN07), with r1's anchored per-directory allow-list expressions. SD01 and SD02 have no fixture. |
| Validation tools | 20 | `infra/validation/tools/{go.mod,go.sum}`; `cmd/{contractcheck,httpprobe,logcheck}/{main.go,main_test.go}` (6); `internal/conformance/{validate.go,audit.go,derive.go,ndjson.go,produce.go,conformance_test.go}` (6); `internal/logrecord/{completion.schema.json,events.json,logrecord.go,logrecord_test.go}` (4); `infra/docker/validation-tools.Dockerfile` and `.dockerignore` (2). r1 stated "(17)" for a list of 18 (F23). |
| CI | 7 | `.gitlab-ci.yml`, `infra/ci/{validate.sh,lib.sh,validate-controls.sh,cleanup-owned.sh,runner-preflight.sh,unchanged.allow}` |
| Harness | 1 | `infra/validation/required-checks.tsv` |
| Docs | 1 | `docs/foundation/OPERATIONS.md` (r1 coverage, plus: error-output handling, attribution profiles, the finalization and exit-code model, the registry and cleanup, the runner requirements and R1. It names no runner tag, GitLab path or identity literal.) |
| Management | 2 | `management/v0/results/G003-scope-allowlist.txt` (**created at S1**, with the full final content of this section) and `management/v0/results/G003-foundation-observability-ci-result.md` (after execution) |

Count: 6 + 4 + 6 + 128 + 20 + 7 + 1 + 1 + 2 = 175.

The probe-only files of R1 (`.gitlab-ci.yml` probe variant, `infra/ci/runner-probe-job.sh`, `infra/ci/probe.allow`, and a preview of `runner-preflight.sh`) exist only in the disposable R1 clone and on the deleted probe branch. They are **not** repository paths and are not in the allow-list.

### 14.4 Modified project-owned files

- `infra/validation/run.sh`, `infra/validation/lib.sh`: r2's changes, plus (r3):
  - the one-line TERM/INT/HUP-ignoring `tee`;
  - `owned_capture` at every start or `up` of a registered project and first in the cleanup trap;
  - `image <tag> <id>` rows after every build;
  - the `MAIN_GO_G003_SHA` pin.
- `infra/validation/render-customizations.tsv` (+11 rows).
- `README.md`: health lists all six routes and the headers; new blocks `platform:contracts` and `platform:ci`; a link to OPERATIONS.md; the existing blocks are unchanged.

### 14.5 Not touched

- `core/cmd/migrate/**`, `core/internal/{config,mcp,migrate}/**`, `core/migrations/**`, `core/go.{mod,sum}`.
- `core/internal/health/{health.go,health_test.go}`, `core/internal/observability/observability_test.go`.
- `web/app/{layout.tsx,page.tsx,globals.css}`, `web/components/**`, `web/lib/{config,appearance,tokens,utils}.ts`, `web/package-lock.json`, all web configuration files, `web/tests/browser/**`, `web/tests/fixtures/token-scopes/**`.
- `intelligence/{pyproject.toml,uv.lock}`, every other `intelligence/src/**` file, the existing `intelligence/tests/*.py`.
- `contracts/README.md`, `contracts/intelligence/**`.
- `infra/compose/**`, every other Dockerfile and `.dockerignore`, `infra/{local,minio,postgres,security}/**`.
- `infra/validation/run-approved{,-controls}.sh`.
- `.env.example`, `.gitignore`, `.archetype/**`.
- `skills/**`, including the in-progress consolidation, which belongs to the owner. AGENTS.md, CLAUDE.md, `docs/ARCHITECTURE.md`, SECURITY.md, DECISIONS.md, `docs/history/**`, the other `docs/foundation/*.md`.
- ROADMAP, BACKLOG, the G003 Goal and Spec, PROGRESS (Codex-owned), and every G000–G002 artifact.
- The tool-local `.claude/settings.local.json`: never read or modified.

### 14.6 Temporary writes (none persist in the repository)

| Write | Location | Removal |
|---|---|---|
| S0 gate copy and evidence | `$EXT/g003-exec/` (outside the repository) | kept as evidence outside the repository |
| R1 probe clone | `$EXT/r1/probe` | deleted after the probe branch is deleted |
| Go module cache for the tools module (S2) | `$EXT/gomod` | deleted at the end |
| Component-check, runtime and probe images | `g003-$EXEC-{core,web,intelligence,vtools}{,-check}`, `g003-$EXEC-web-probe`, labelled `g003.exec=$EXEC` | removed at the end, by label |
| Ad hoc error-output probe inputs and outputs (S3–S5) | `$EXT/s3`, `$EXT/s4`, `$EXT/s5` (environment files mode 600) | canary-bearing inputs deleted after use; outputs kept only after a canary scan |
| Harness work, Compose projects, images, root `.env` | as today (`project-validation-*`, `g001-*-<run>`; root `.env` created and hash-removed by the root suite) | the harness's own cleanup, then `cleanup-owned.sh` verification in CI |
| CI owned registry | `$(dirname "$CI_PROJECT_DIR")/.platform-validation-owned-$CI_JOB_ID` (CI only) | F9 or `after_script` |
| Local CI-parity copy (S8) | `$EXT/s8/clone`, with a scratch commit (Q6) | deleted after the run |
| Hosted evidence downloads | `$EXT/hosted/` | kept outside the repository |

## 15. Execution sequence, gates and provenance

### 15.1 S0: read-only, fail-closed starting-state gate (adapted from approved G002 r6 S0a)

- **Reuse.** The committed wrapper `run-approved.sh`, with `--check-only`, already provides the snapshot-digest, format, count and entry checks, so G003 needs no bootstrap writes. S0 adds what the wrapper does not check:
  - the HEAD binding;
  - the NUL-safe starting state;
  - the resource and port inventory;
  - evidence containment.

  It keeps G002 r6's read-only contract: `realpath -m` containment before any `mkdir`, and Git with `--no-optional-locks -c core.fsmonitor=false`, `GIT_OPTIONAL_LOCKS=0` and redirection variables unset.
- **Running it.** The script runs from **outside the repository** after byte-exact extraction from the block below. Its SHA-256 must equal **`0409148a630f32cc79642fd8ae65b90a53001eba379775bd96a3c1f313a66723`**; a different hash means stop. BASE, SNAPSHOT, SNAPSHOT_SHA256 and COUNT are taken verbatim from Codex's approval record. Nothing is regenerated, and the G002 snapshot is never reused.

```bash
bash <ext>/g003-s0-gate.sh <repo> <BASE> <approved snapshot> <its SHA-256> <COUNT> <ext>/g003-exec/s0
```

| Step | Verifies | Exit |
|---|---|---|
| 0 | Six arguments. BASE is 40 hex, the digest is 64 hex, COUNT is 1–5. The repository resolves. The evidence path resolves **outside** the repository, including through traversal and symlinks (dangling ones included), **before any write**. | 26 |
| 1 | `HEAD` = BASE | 20 |
| 2 | The snapshot is read once; the SHA-256 of those bytes equals the pinned digest | 21 |
| 3 | Only those bytes are parsed: `<64hex>  <path>`, no escaped entries, relative, no `.` or `..`, unique, inside the G003 input set (Goal, Spec, PROGRESS.md, this Plan, ARCHITECTURE.md), exactly COUNT entries, and the Plan present | 22 |
| 4 | Every entry is a regular non-symlink file whose SHA-256 matches. This holds whether or not the entry is committed. | 23 |
| 5 | NUL-safe `status --ignored`: every differing path is `??` or ` M` **and** is a snapshot path. Nothing staged, deleted, renamed, ignored (including `.env` and `.env.local`) or extra is allowed. **The single exception** is the exact ignored path `.claude/settings.local.json`, which is counted and reported, never read (F12). | 24 |
| 6 | No root `.env`; Compose 2.24 or later; no containers, volumes or networks for the project; default ports free; the Docker inventory written to evidence | 25 |

**Exit 0 is the only way to proceed.** Any other exit stops the work before the first repository write.

Exact S0 script (SHA-256 `0409148a630f32cc79642fd8ae65b90a53001eba379775bd96a3c1f313a66723`; extract the lines between the fences, including the final newline):

```bash
#!/usr/bin/env bash
# G003 S0: read-only, fail-closed starting-state gate (G003 Plan §15.1, r1). Adapted from the
# approved G002 r6 S0a gate with the same read-only contract. Needs only bash (>= 4.4), GNU
# coreutils, git and docker. Writes nothing inside REPO, including .git: Git runs without optional
# locks, so the index is never refreshed. The evidence destination is resolved and proven outside
# REPO before anything is created.
# Usage: bash g003-s0-gate.sh REPO BASE SNAPSHOT SNAPSHOT_SHA256 COUNT EVIDENCE_DIR
#   BASE, SNAPSHOT, SNAPSHOT_SHA256 and COUNT are the values pinned by Codex's Plan approval.
# Exit: 0 pass; 20 HEAD; 21 snapshot digest; 22 snapshot format/paths/count; 23 input hash;
#       24 starting state; 25 local resources or tools; 26 usage or evidence path.
set -euo pipefail
shopt -s inherit_errexit

fail() { printf 'S0 FAIL %s: %s\n' "$1" "$2" >&2; exit "$1"; }
[[ $# -eq 6 ]] || fail 26 "usage: REPO BASE SNAPSHOT SNAPSHOT_SHA256 COUNT EVIDENCE_DIR"

# Git inspection must not write repository metadata (no index refresh, no fsmonitor) and must not
# be redirected by the caller's environment.
export GIT_OPTIONAL_LOCKS=0
unset GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE GIT_OBJECT_DIRECTORY GIT_ALTERNATE_OBJECT_DIRECTORIES \
  GIT_COMMON_DIR GIT_NAMESPACE GIT_CONFIG_PARAMETERS GIT_CONFIG_COUNT
g() { git --no-optional-locks -c core.fsmonitor=false -C "$REPO" "$@"; }

# 0. Arguments, and every location resolved before any write.
REPO="$(realpath -e -- "$1" 2>/dev/null)" || fail 26 "repository path does not resolve"
[[ -d "$REPO" ]] || fail 26 "repository path is not a directory"
BASE="$2" SNAP="$3" SNAP_SHA="$4" COUNT="$5"
[[ "$BASE" =~ ^[0-9a-f]{40}$ ]] || fail 26 "BASE must be a full 40-hex commit"
[[ "$SNAP_SHA" =~ ^[0-9a-f]{64}$ ]] || fail 26 "SNAPSHOT_SHA256 must be 64 lowercase hex"
[[ "$COUNT" =~ ^[1-5]$ ]] || fail 26 "COUNT must be 1-5"
EV="$(realpath -m -- "$6" 2>/dev/null)" || fail 26 "evidence path does not resolve"
inside_repo() { [[ "$1" == "$REPO" || "$1" == "$REPO"/* ]]; }
inside_repo "$EV" && fail 26 "evidence destination resolves inside the repository: $EV"
# Only now may anything be created; the resolved path is re-verified before every write.
mkdir -p -- "$EV"
ev_ok() {
  local now
  now="$(realpath -e -- "$EV" 2>/dev/null)" || fail 26 "evidence directory vanished"
  if [[ "$now" != "$EV" || ! -d "$now" ]] || inside_repo "$now"; then fail 26 "evidence directory changed: $now"; fi
}
ev_ok

PLAN=management/v0/plans/G003-foundation-observability-ci-plan.md
PATH_RE='^(management/v0/(goals/G003-foundation-observability-ci\.md|specs/G003-foundation-observability-ci-spec\.md|PROGRESS\.md|plans/G003-foundation-observability-ci-plan\.md)|docs/ARCHITECTURE\.md)$'
LINE_RE='^([0-9a-f]{64})  ([^\].*)$'
cd -- "$REPO"

# 1. Baseline commit.
[[ "$(g rev-parse HEAD)" == "$BASE" ]] || fail 20 "HEAD is not $BASE"

# 2. Snapshot: read once; verify the digest of exactly those bytes.
bytes="$(cat -- "$SNAP"; printf x)" || fail 21 "cannot read the snapshot"
bytes="${bytes%x}"
[[ "$(printf '%s' "$bytes" | sha256sum | cut -d' ' -f1)" == "$SNAP_SHA" ]] || fail 21 "snapshot digest mismatch"

# 3. Parse only those bytes: format, relative unique approved paths, exact count, Plan present.
[[ -n "$bytes" && "$bytes" == *$'\n' ]] || fail 22 "snapshot is empty or lacks a final newline"
declare -A want=()
n=0
while IFS= read -r line; do
  [[ "$line" =~ $LINE_RE ]] || fail 22 "malformed entry: $line"
  sha="${BASH_REMATCH[1]}" path="${BASH_REMATCH[2]}"
  [[ "$path" != /* && "/$path/" != */./* && "/$path/" != */../* ]] || fail 22 "non-relative path: $path"
  [[ "$path" =~ $PATH_RE ]] || fail 22 "path outside the approved set: $path"
  [[ -z "${want[$path]+x}" ]] || fail 22 "duplicate path: $path"
  want[$path]="$sha"; n=$((n + 1))
done <<<"${bytes%$'\n'}"
(( n == COUNT )) || fail 22 "expected $COUNT entries, found $n"
[[ -n "${want[$PLAN]+x}" ]] || fail 22 "the Plan is not an entry"

# 4. Every input is a regular file matching its approved hash (committed or not).
for path in "${!want[@]}"; do
  [[ -f "$path" && ! -L "$path" ]] || fail 23 "missing or non-regular input: $path"
  [[ "$(sha256sum -- "$path" | cut -d' ' -f1)" == "${want[$path]}" ]] || fail 23 "hash mismatch: $path"
done

# 5. Starting state (NUL-safe, no optional locks): every path that differs from HEAD is a snapshot
#    path and is untracked or unstaged-modified; nothing staged, deleted, renamed or ignored.
#    Snapshot paths that are committed unchanged at BASE are allowed (step 4 bound their bytes).
#    The single exception is the ignored, tool-local Claude Code settings file named below: it is
#    never a repository input, is not read, and is reported in the evidence.
TOOL_LOCAL=.claude/settings.local.json
ev_ok
g status --porcelain=v1 -z --untracked-files=all --ignored >"$EV/s0-status.z" || fail 24 "git status failed"
ndiff=0 ntool=0
while IFS= read -r -d '' rec; do
  xy="${rec:0:2}" path="${rec:3}"
  if [[ "$xy" == "!!" && "$path" == "$TOOL_LOCAL" ]]; then ntool=$((ntool + 1)); continue; fi
  [[ "$xy" == "??" || "$xy" == " M" ]] || fail 24 "unexpected status '$xy': $path"
  [[ -n "${want[$path]+x}" ]] || fail 24 "differing path not in the snapshot: $path"
  ndiff=$((ndiff + 1))
done <"$EV/s0-status.z"

# 6. Local resources the execution must never touch, and the required tools.
project="$(sed -n 's/^COMPOSE_PROJECT_NAME=//p' .env.example)"
[[ -n "$project" ]] || fail 25 "cannot read COMPOSE_PROJECT_NAME from .env.example"
[[ ! -e .env && ! -L .env ]] || fail 25 "a root .env exists"
cv="$(docker compose version --short 2>/dev/null)" || fail 25 "docker compose unavailable"
[[ "$cv" =~ ^v?([0-9]+)\.([0-9]+) ]] || fail 25 "unparsable compose version: $cv"
(( BASH_REMATCH[1] > 2 || (BASH_REMATCH[1] == 2 && BASH_REMATCH[2] >= 24) )) || fail 25 "Compose >= 2.24 required (found $cv)"
for kind in "ps -a -q" "volume ls -q" "network ls -q"; do
  # shellcheck disable=SC2086
  out="$(docker $kind --filter "label=com.docker.compose.project=$project")" || fail 25 "docker query failed"
  [[ -z "$out" ]] || fail 25 "existing $project resources ($kind)"
done
for port in 3100 8180 8100 5532 6479 9100 9101; do
  if (exec 3<>"/dev/tcp/127.0.0.1/$port") 2>/dev/null; then fail 25 "port $port is busy"; fi
done
ev_ok
{ docker ps -a --format 'container {{.Names}}'; docker volume ls --format 'volume {{.Name}}'
  docker network ls --format 'network {{.Name}}'; docker image ls --format 'image {{.Repository}}:{{.Tag}}'; } |
  LC_ALL=C sort >"$EV/s0-docker-inventory.txt"

ev_ok
printf 'S0 PASS: HEAD %s; snapshot %s; %d inputs verified; %d differing path(s), all snapshot paths; %d tool-local ignored file(s); no .env, no %s resources, default ports free; compose %s\n' \
  "$BASE" "$SNAP_SHA" "$n" "$ndiff" "$ntool" "$project" "$cv" | tee "$EV/s0-gate.txt"
```

### 15.1a S0 is unchanged; the baseline is PENDING (r3; r2 text in Appendix R2)

- **The script is byte-identical** (`0409148a630f32cc79642fd8ae65b90a53001eba379775bd96a3c1f313a66723`, re-extracted and verified at r3). PATH_RE, the COUNT range, the single tool-local exception (Q13) and the exit codes are unchanged. No exception has been added for the consolidation.
- **Current state (§2.1).** The consolidation is complete but uncommitted. S0 would stop at step 5 with exit 24 on the first `skills/vestive-design-skill/...` record. This is the intended outcome.
- **What unblocks S0.** Every item in §12.6, in order. Route A (commit the consolidation; bind BASE) remains recommended. Any other route needs a reviewed Plan revision of the S0 inputs, never an execution-time exception.
- **r1 demonstrations.** r1's demonstrations (§15.2) remain valid for the unchanged script.

### 15.2 Planning-time demonstration (read-only for the real repository)

- **Setup.** Run on 2026-10-05 with the exact script above.
  - shellcheck v0.11.0 (`-S warning`) is clean.
  - Every case ran in a **disposable `git clone`** of the baseline, in the job's scratch directory, with an untracked stand-in Plan.
  - Snapshots were built from the clone's files. These snapshots demonstrate the gate only; they are **not approval snapshots**.
- **Measurement.** Around every invocation: a whole-clone inventory including `.git` (path, type, size, nanosecond mtime, inode) and `.git/index` hash plus stat data.
- **Real repository.** Its inventory, including `.git`, and its index were unchanged across the whole demonstration.

| Case | Input | Want | Got | Tree / index |
|---|---|---|---|---|
| P | Plan (untracked) + Goal + Spec + PROGRESS, COUNT 4 | 0 | 0, `4 inputs verified; 1 differing path(s), all snapshot paths; 0 tool-local ignored file(s); no .env, no beginify resources, default ports free; compose 2.32.4-desktop.1` | same / same |
| P5 | As P plus ARCHITECTURE.md, COUNT 5 | 0 | 0 | same / same |
| N1 | Wrong pinned digest | 21 | 21 | same / same |
| N2 | COUNT 3 with 2 entries | 22 | 22 | same / same |
| N3 | Duplicate entry | 22 | 22 | same / same |
| N4 | Absolute path | 22 | 22 | same / same |
| N5 | `README.md` entry (outside the set) | 22 | 22 | same / same |
| N6 | Plan not an entry | 22 | 22 | same / same |
| N7 | Goal modified after snapshot | 23 | 23 | same / same |
| N8 | Extra untracked file | 24 | 24 | same / same |
| N9 | Plan staged (`A `) | 24 | 24 | same / same |
| N10 | Root `.env` (`!!`) | 24 | 24 | same / same |
| N11 | HEAD moved by an empty commit | 20 | 20 | same / same |
| N12 | Unlisted tracked file modified | 24 | 24 | same / same |
| T1 | Ignored `.claude/settings.local.json` present (ignored as in the real checkout) | 0 | 0, `1 tool-local ignored file(s)` | same / same |
| T2 | Ignored `.env.local` present | 24 | 24 (`!!` `.env.local`) | same / same |
| T3 | Untracked `.claude/other.json` | 24 | 24 (not in the snapshot) | same / same |
| A1 | Evidence path inside the repository (nonexistent) | 26 | 26; **not created** | same / same |
| A2 | Evidence through an external symlink into the repository | 26 | 26; **not created** | same / same |
| A3 | Abbreviated BASE | 26 | 26 | same / same |
| C1 | Stale-index fixture (tracked mtimes changed), positive path | 0 | 0 | same / **same** |
| C0 | Ordinary `git status` (optional locks enabled) on that fixture afterwards | index rewritten | rewritten: the detector is non-vacuous | — |

### 15.3 Sequence (r3; the r2 order, accepted by Codex, with r3 gate additions; r2 and r1 versions in the appendices)

| Step | Work | Gate (exact commands in §15.6) |
|---|---|---|
| S0 | §15.1 gate, then the wrapper `--check-only`, then `run-approved-controls.sh` (W1–W9). **Needs every prerequisite in §12.6.** | All exit 0, or stop before any write |
| **S0R** | **R1 runner qualification** (§12.4). Needs Q1, A-R1 and A-E. A disposable clone outside the repository; **no repository write.** | The §12.4 decision rule is met and reported to Codex; otherwise stop before S1 (no cycle consumed) |
| S1 | **First repository write: `management/v0/results/G003-scope-allowlist.txt`**, with the full §14 content. B-SCOPE checks that changed paths are a *subset* of the allow-list, so every later intermediate state passes it.<br>Then: the contracts corpus, `expectations.json` (including `schema_negative`, SD01 and SD02), the conformance README, the `schemas.py` corrections, `test_conformance.py` and `conformance_exchange.py`, and the `intelligence.Dockerfile` check-stage `COPY`. | The intelligence check image builds. `pytest -q` passes, with the conformance, SN-class and SD tests included and none skipped. |
| S2 | The tools module (`go get …@v6.0.3 && go mod tidy` in a container, the only network step), the commands, tests, `events.json` and the Dockerfile | The tools check stage passes; `contractcheck corpus` agrees with the expectations; `contractcheck schema-negative` reports every declared class, a passing `sn-derive` and SD01/SD02 as digest mismatches |
| S3 | Core: `correlation.go`, `logging.go`, `checks.go`, `exit.go`, `errorlog.go`, the two `main.go` hunks, and the tests | `core-check` (gofmt, vet, all tests) passes. Then the core runtime image is built and **CE1–CE3** are run ad hoc (§15.6). CE2: exactly one `core.exit`/`listen_failed` and no `core listening`. CE3: exactly one bounded `core listening`. No canary in any. `main.go` must hash to `987c9335…51ea`. |
| S4 | Web: `request-context.ts`, `core-client.ts`, both routes, the scripts, the unit test and the two error-probe fixtures | `web-check` (lint, type check, unit) passes. The production build passes. **WE0–WE6** (with WE3a and WE3b) run ad hoc on a disposable probe copy (§15.6). This is the first **runtime** evidence of F15; P-NEXT is source evidence only. A deviation stops for review (R14). |
| S5 | Intelligence: `http.py`, `export.py`, `main.py`, `test_error_output.py` and the other tests | The intelligence check passes (ruff, format, all tests). The runtime image is built and **IE1/IE2** are run ad hoc (§15.6). The uvicorn lifespan and h11 templates are confirmed against the installed 0.54.0 package. |
| **S6** | **Integration, in one step, before any static or provenance run reads the result:**<br>• (a) CI files: `.gitlab-ci.yml` (the Q1 tag), `infra/ci/validate.sh`, `lib.sh`, `validate-controls.sh`, `cleanup-owned.sh`, `runner-preflight.sh`, `unchanged.allow`;<br>• (b) documentation: the README `platform:contracts` and `platform:ci` blocks, `docs/foundation/OPERATIONS.md`, the FOLLOW-UP labels (§17);<br>• (c) `infra/validation/required-checks.tsv` (69 IDs);<br>• (d) every §10 harness change: the contracts suite, the error-output checks, the `obs_*` functions, B-HEALTH-ROOT, B-REQUIRED-CHECKS, the register (+11), `APPLICABLE` (+7), the `MAIN_GO_G003_SHA` pin and imports literal, the B-SHELL, B-HADOLINT, B-DOCKERIGNORE, B-DEPS, B-WEB-BOUNDARY and B-MCP-SOURCE changes, and the `--owned-registry` and `--secrets-file` options. | In order, all in the same cycle:<br>• **G6a** shellcheck on every script, plus `--include=SC2251` (F27) for `infra/ci/*.sh`;<br>• **G6b** `validate-controls.sh` standalone: groups A–E (68 controls), including V4's nested real `contracts` run, the external-signal supervision controls V18–V18j, and the Docker-real cleanup controls V19–V26;<br>• **G6c** the wrapper with `static provenance contracts`: **all PASS**. This is the G002 lesson: a shared harness change is followed by the static suite in the same cycle. |
| S7 | — | The wrapper with `root migrations variants`: **all PASS** (B-HEALTH-ROOT, the three error-output checks, the B-OBS-* checks, and every inherited root and variant check) |
| S8 | **Local CI parity:** `validate.sh` in a disposable scratch-commit copy (Q6). Then **H0**, if A-H0 is authorized. | `validate_rc=0`, `tested_content=local`; H0's evidence recorded (§12.5) |
| S9 | Two read-only reviewers (§20); triage; fixes | Findings resolved or recorded. The suites covering each changed file are rerun, and a harness or CI change always reruns G6a, G6b and the `static` suite in the same cycle. |
| S10 | **Final local run with no edits during the run:** the wrapper with `all` (§15.4); the cleanup inventory; the Result draft | `overall_rc=0` |
| S11 | H1 and H2 (with A-H1, A-H2 and A-E), then H3 (A-H3); the final Result; stop for Codex | A hosted pass on the exact commit; a hosted failure for the intended reason |

**r1 → r2 step map:**
- S0 → S0.
- New: S0R.
- S1–S5 → S1–S5, plus the ad hoc error-output probes; the allow-list moves to S1.
- r1 S6 (harness) + S7 (CI) + S8 (docs) → S6, with the allow-list already created at S1.
- r1 S9 → S7 + S8.
- r1 S10 → S9; r1 S11 → S10; r1 S12 → S11.

The **cycle-4 checkpoint** is recorded *after* cycle 4 closes and before cycle 5 begins, wherever that falls (§19).

### 15.4 Final local invocation (S10)

```bash
bash infra/validation/run-approved.sh \
  --snapshot <approved snapshot> --snapshot-sha256 <its SHA-256> --count <COUNT> \
  --path-re '^(management/v0/(goals/G003-foundation-observability-ci\.md|specs/G003-foundation-observability-ci-spec\.md|PROGRESS\.md|plans/G003-foundation-observability-ci-plan\.md)|docs/ARCHITECTURE\.md)$' \
  --plan management/v0/plans/G003-foundation-observability-ci-plan.md \
  -- --base <BASE> --allow-file management/v0/results/G003-scope-allowlist.txt --report-dir <ext>/g003-exec/s10/report all
echo "overall_rc=$?"
```

- **Arguments.** Every non-Plan snapshot entry becomes `--expect-hash` and an exact `--exception`, as the wrapper builds them today.
- **Outcome.** Exit 0 is the only passing result. Any other status is a failed run: it is recorded, and the handoff stops.
- **Scope.** B-SCOPE checks the exact §14 allow-list against BASE. The harness enumerates with `--exclude-standard`, so the globally ignored tool-local file does not enter B-SCOPE, the copies or the scans.
- **Inputs.** `<BASE>`, the snapshot, its SHA-256 and `<COUNT>` are Codex's Q9 binding, never self-generated.

### 15.5 r2 planning-time probes and source evidence (scratch only; no repository write)

**r3 note:** the P-FIN row below is superseded by P-MODEL (§15.5a). Codex found that r2's model omitted `stage` from aggregation (exit 0 with `stage=19`); the r3 model fixes this. The other r2 probes remain valid as recorded.

These are disposable planning probes under `/tmp/g003-plan-r2/`, the scratch root Codex will see referenced. They are **not** implementation and **not** execution evidence. Runtime confirmation of each point is an S3–S6 gate.
- **Images.** Every container used was already present locally, run with `--pull never`, `--network none` and `--rm`.
- **No application.** No application image was built or run.

| Probe | What was run | Result | Supports |
|---|---|---|---|
| **P-GO** | `main.go` (SHA-256 `de08b71b…be3a`): a stdlib-only scratch program, run with `go run` in `golang:1.27.1-alpine` (`GOTOOLCHAIN=local`), mirroring today's core behaviour and the planned adapters. Output: `out.txt` (`28f39558…1288`). | **Today's behaviour (section A):** plain-text `http: superfluous response.WriteHeader call from main.serve.func1 (main.go:63)`; `http: panic serving 127.0.0.1:…: boom CANARY-7f3a9c` with a full goroutine stack (twice: the client retried); `fmt.Fprintln` of a listen error printing `listen tcp: lookup bad-CANARY-7f3a9c.invalid on …`. That is 3 canary-bearing non-JSON lines. `panic(http.ErrAbortHandler)` printed nothing, and a query containing `;` printed nothing.<br>**Planned behaviour (section B):** with a classifying `ErrorLog` writer and a bounded exit record, 5 lines, every one valid JSON, with `canary=false` (`superfluous_write_header`, `handler_panic` ×2, `listen_failed`, `config_invalid`). | F16, F17; §5.6; §7.1 |
| P-GO-SRC | `grep -n "logf("` over `/usr/local/go/src/net/http/server.go` in the same image | Call sites for hijacked and superfluous `WriteHeader`, hijacked `Write`, invalid `Content-Length`, conflicting length headers, `panic serving`, TLS handshake and `Accept` errors; `logf` falls back to `log.Printf` when `ErrorLog` is nil. There is no semicolon-query log call. `ErrAbortHandler` suppresses the stack. | F17; the §5.6 categories |
| **P-NEXT** (source, not a runtime probe) | Next.js `v16.3.8` tag: `packages/next/src/build/templates/app-route.ts` and `packages/next/src/server/route-modules/route-module.ts` | The route error is rethrown to the outer catch, which calls `routeModule.onRequestError(…, silenceLog=false, …)`, then `sendResponse(…, new Response(null, { status: 500 }))`. `onRequestError` calls `logErrorWithOriginalStack(err, 'app-dir')` or `console.error(err)`. **Not run:** Next cannot be executed without installing dependencies or building, which this handoff forbids. The real-framework probe is WE0–WE6, first at S4 and then permanently as B-WEB-ERROR-OUTPUT. | F15; §5.3; §7.2 |
| P-UVICORN (source) | uvicorn `0.54.0` tag: `config.py` (`LOGGING_CONFIG`, `load_app`) and `server.py` | `uvicorn` has a stderr text handler, `uvicorn.error` propagates to it, and `uvicorn.access` writes to stdout with the request line. `load_app` catches only `ImportFromStringError`. The `server.py` templates are those listed in §5.6. Runtime confirmation: IE1 and IE2 (S5, B-INTEL-ERROR-OUTPUT). | F18; §7.3 |
| **P-FIN-0** | `gotcha.sh` (`238184db…6d56`) | `f \|\| rc=$?` → rc 0, and the function continued past `false`. `( set -e; f ) \|\| …` → rc 0, and it continued. `set +e; ( set -e; f ); rc=$?` → rc 1. | §11.2.1 (the phase form) |
| **P-FIN** | `model.sh` (`3b8233dd…a898`): a control-flow model of §11.2 with stub phases; 11 fault cases; `cases.tsv` (`663900c2…ee30`) | none→0 (published); harness→1; metadata→16 (harness not run; postcheck and cleanup ran; nothing published); malformed report→14; scanner error→17 (nothing published); planted canary in an excluded raw file→15 (published `WITHHELD` only); drift→13; cleanup→18; publish failure→19; unexpected error outside a phase→19 (finalizer ran); SIGTERM during the harness→143 (postcheck and cleanup ran; nothing published). In every case postcheck and cleanup ran; no failure produced 0. | §11.2.3–§11.2.5; groups C and D |
| P-TERM | `termtrap.sh` (`d4e04cbd…2b27`) under `timeout --signal=TERM --kill-after=5s 1s` (GNU coreutils 9.4, Bash 5.2.21) | `timeout` returned 124, and the child's EXIT trap ran (`tt.out`). | §11.2.1 (the internal deadline keeps the harness's own cleanup) |
| **P-ATTR-0** | A Git fixture: target T, source S and merge M (`--no-ff`, parents T and S) under `refs/merge-requests/1/merge`; a depth-1 fetch and checkout of M | `.git/shallow` = M. `git cat-file commit HEAD` lists both parents (T, then S). `git rev-parse --verify HEAD^2` fails at the shallow boundary. | §11.2.4 (the raw parent check) |
| **P-ATTR** | `classify.sh` (`1e0dfe34…2984`), the prototype of §11.2.4, over the fixture with `env -i` and GitLab-defined variables; `cases.sh` (`6cd60dab…5dd5`), results in `results.tsv` (`d5e7a3fe…498e`) | **21 of 21 as expected.** P1 and P2 `branch-push`; P3 `mr-source`; P4 `mr-merged-result`; P5 `local` (all exit 0). N1–N16 each exit 16 with their specific reason. | Correction 2; group B |
| **P-SCAN** | `zricethezav/gitleaks:v8.30.1 dir` with the repository's `gitleaks.toml`, `--redact --exit-code 23`, over scratch directories | clean → **0**; a planted generated `ghp_` token → **23** (`leaks found: 1`); a missing configuration file → **1**; a missing path → **1** | F22; §11.2.3 F3 and F7; V5c and V6a |

Two observations are recorded for Codex:
- **(i)** The Go 1.27.1 standard library no longer logs semicolon queries. r1 assumed nothing about it, and r2 asserts nothing about it.
- **(ii)** gitleaks reports both a fatal error and a default leak with exit 1. That is why `--exit-code 23` is required to tell a scanner failure from a finding.

### 15.5a r3 planning probes and models (scratch only, under `/tmp/g003-plan-r3/`; no repository write; no Docker resource created)

These are disposable, non-application planning checks. They are **not** G003 implementation and **not** execution evidence. Runtime confirmation is the S3–S6 gates and the group D and E controls.

| Probe | What was run | Result | Supports |
|---|---|---|---|
| **P-STATE** | `gotcha-state.sh` (`0e34f825…3eef`): r2's phase helper wrapping a location body | `location status=0 EV=<unset> REG=<unset>`. The body's `mktemp -d` had already created a directory, which was then removed. This reproduces Codex's finding. | F24; §11.2.1 |
| **P-SETSID** | `setsid-probe.sh` (`afd70614…cfd9`; util-linux 2.39.3) | `$!` = pid = pgid = sid of the background `setsid` child. `kill -TERM -- -$!` gave rc 143 and the group was gone. | F25 |
| **P-TEE** | `v5.sh` (`3376f082…1b70`): a stub shaped like the current `run.sh` under the supervisor's watchdog plus group TERM; outputs in `p-tee.txt` (`0f1a2d67…40fa`); variants `v2.sh`–`v4.sh` | **As-is** (`exec > >(tee …)`): harness rc **141** (SIGPIPE), cleanup **NO**, three of three runs. **`tee` ignoring TERM/INT/HUP:** rc 143, cleanup **ran**, log tail `cleanup: done`, three of three runs. Under GNU `timeout`, a process-substitution variant also skipped its EXIT trap. A TERM sent only to the leader is deferred until the foreground child ends. | F26; §10 one-line `tee` change; §11.2.1 (supervisor-owned deadline and group TERM) |
| **P-IMPORT** | `state-import-test.sh` (`89004f1b…bfa5`) against the model's `state_import`; output `state-import.txt` (`d5c22f09…d03f`) | 8 of 8 as expected. Valid input accepted and assigned. Unknown key, duplicate, bad value, missing key, `1;touch /tmp/pwn` injection (no side effect), lowercase key and symlink all rejected. | §11.2.1 state import; V18e |
| **P-MODEL** | `vmodel.sh` (`a18c0e65…863a`): a control-flow model of §11.2 with real `setsid`, watchdog, signals, `git` postcheck, literal scan, all-or-nothing staging, candidate selection, atomic publish, report, scrub and the aggregation invariant. Stubs from `stubs.sh` (`bdf2698b…f0de`). Driver `cases.sh` (`6d8f6feb…60a0`), with a fresh scratch repository per case; output `run4.txt` (`e4737b20…c3db`) | **24 of 24 as expected:** F01 all-valid 0/full · F02 harness 1/full · F03 preflight 16/full (preflight evidence) · F04 malformed 14 · F05 scanner error 17/none · F06 retained canary 15/minimal · F07 excluded-console canary 15/minimal · F07b clean exclusion 0/full, console absent · F08 gitleaks pattern 15/minimal · F09 drift 13 · F10 cleanup 18 · **F11 staging after partial copy 19/minimal, nothing partial** · F11b unreadable raw 17/none · **F12 summary 19/none** · F13 publish 19/none · **F14 scrub 19, full already published** · F15 unexpected parent failure with the harness running 143, group stopped · **F16 external SIGTERM 143, group gone, harness cleanup ran, postcheck and cleanup ran** · F17 TERM-ignoring harness 143, KILL after grace · F18 watchdog deadline 124 · F19 stragglers 0, group stopped · **F20–F22 pre-location 16, repository inventory unchanged, nothing created, published none.** Every case ran F1 and F2, except the pre-location cases (F1 only, by design). | Corrections 1 and 2; groups C and D |
| P-MODEL defects found and fixed during r3 (disclosed) | First runs of the r3 model | **(i)** `! grep …` used as a statement never failed under `set -e`, so a failing preflight passed. Fixed with explicit returns; F27 adds the SC2251 lint. **(ii)** Signalling the harness group immediately after `setsid` launch missed it (F15 first run: group alive). Fixed with `group_ready`. **(iii)** The same race on the watchdog made each case wait 60 s. Fixed the same way. **(iv)** A stub-only bare `wait` blocked on its own process-substitution `tee`; the stub now waits on named PIDs. **(v)** The scrub left `state/`, so `rmdir` failed. The no-masking invariant correctly turned this into 19 before the fix. | Shows the invariant and the controls catch defects of exactly the kind Codex found |
| **P-OWN** | `own/own-model.sh` (`bfe86575…7d04`): the §11.4 decision rules over simulated inventories (no Docker objects); output `own-model.txt` (`d526f406…6234`) | **Removed (positively owned):** the owned ops container and root container; their captured networks and volumes; the images of owned containers; the build-registered `g003-vtools` image. **Refused or unproven (kept):** V20c foreign-`working_dir` container; V20n network and V20v volume under a registered label with no capture; V20i `beginify-intruder:latest` and `g001-ops-…-extra:latest` (registered-project name, no record). **Kept (pre-existing):** network, container and image in the start inventory. **Ignored:** unregistered project resources. | Correction 3; group E |
| P-MAIN | Scratch copy of `core/cmd/server/main.go` with the three r3 hunks; `gofmt` in `golang:1.27.1-alpine` (`--pull never --network none`); `patch` replay of the embedded hunks | gofmt-clean output, SHA-256 `987c933560ef77ea9a2daa4a3bec4c9747ea69d578f3c9c9d8e6179d013551ea`. Replaying the §7.1 hunks onto the current file reproduces that hash. Three hunks. Not compiled (no dependency download in planning). | §7.1; B-MCP-SOURCE pin |
| P-BASE | Read-only: `sha256sum -c references/shadcn-kit.sha256` in the skill; status counts; web-versus-asset identity | 63 of 63 OK, exit 0; 2 modified and 65 new skill paths; no other differing path besides PROGRESS and the Plan; 0 web files identical to an asset | §2.1; correction 5 |

**Correction of the r2 model (Codex finding).** r2's `model.sh` omitted `stage` from its aggregation list, and Codex reproduced `stage=19` with **exit 0**. The r3 model aggregates over the single `PHASES` array that the finalizer records into, and applies the structural no-masking invariant (§11.2.5). F11 now exits 19. r2's model and its outputs remain under `/tmp/g003-plan-r2/` as history; they are superseded, not reused.

### 15.5b r4 planning models (scratch only, under `/tmp/g003-plan-r4/`; no repository write; no Docker resource created)

| Probe | What was run | Result | Supports |
|---|---|---|---|
| **P-MODEL r4** | `model/vmodel.sh` (`2de802f8…9946`).<br>**Model changes from r3:**<br>• parent-only provenance recording (`signal`, `internal`, `harness_stop`);<br>• `aggregate` in the parent, setting `FINAL` without mutating state;<br>• the post-harness parent record `state/harness.result`;<br>• local and CI registry layouts, with a skeleton (`owned.tsv`, `token`, `start-inventory.ids`, `cid/`);<br>• F2 copies its inventory into the registry, and a non-`removed` row means 18;<br>• F10 retention: the secret list is always scrubbed, and the registry and a local `$EV` are kept on failed cleanup.<br>Stubs `model/stubs.sh` (`345828ac…978c`; adds `block-result`). **Driver `model/cases.sh` (`a8e370c6…fe42`):**<br>• every case runs entirely under its own directory;<br>• **r3's broad `/tmp/g003-validate-*` deletion is removed**;<br>• new assertions: `has:`/`not:` on the final line, `regkept` (owned.tsv, start-inventory.ids, token, cid/ and cleanup-inventory.tsv present, with a non-`removed` row; no `secrets`; mode 700; expected layout path) and `regremoved`.<br>Output `model/run1.txt` (`4dd8451e…e059`). | **29 of 29 as expected.** The 24 r3 cases still pass, with `not:internal=` asserted wherever no orchestration error occurs. **New and strengthened cases:**<br>• **F15:** 143, with `harness=143`, `internal=19`, `harness_stop=finalizer`, no `signal=`;<br>• **F15b:** a parent error after a successful harness → **19**, with `harness=0` and `internal=19`;<br>• **F16 and F17:** `signal=143`, `harness_stop=signal`, no `internal=`;<br>• **F18:** `harness_stop=deadline`;<br>• **F10 (local) and F10c (CI):** 18, `registry=kept:…` at the expected path, every record present, secrets gone, inventory not empty;<br>• **F10h and F10hc:** harness 1 plus cleanup 18 → exit **1**, registry kept;<br>• **F01 and F01c:** `registry=removed`, no evidence root or CI registry left;<br>• F14 (scrub failure) also keeps the registry. | Corrections 1 and 2b; V18, V18d, V18f, V18k, V32–V35 |
| **Non-vacuity control** | The same r4 driver and stubs run against the **r3** model (`control-r3/`; output `run-r3-model.txt`, `124f284f…c406`) | **12 FAIL / 17 PASS.** These include F15 (lacks `internal=19` and `harness_stop=finalizer`, Codex's finding), F15b (exits **0** where 19 is required), F10 (registry not reported kept), and F16–F18 (provenance fields missing). The strengthened assertions therefore detect the defect class. | Correction 1 |
| **P-OWN r4** | `own/own-model.sh` (`bdebbb03…77f5`), adding standalone containers with cidfile rows and registered labels; output `own/own-model.txt` (`622fd020…eaa0`) | **Removed:**<br>• standalone containers bound by cidfile rows (`sa-probe` boundary probe, `sa-ce3` CE3 core, `sa-pre` preflight httpd);<br>• the CE3 image, through its **build row** only.<br>**Refused (label-only):** V27-forged (`g001.run=<run id>`) and V28-forged (`platform.validation.preflight=<token>`).<br>**Kept:** the pre-existing standalone and `busybox`.<br>**Ignored:** the unregistered standalone.<br>All r3 Compose, network, volume and image decisions are unchanged. | Correction 2a; V27–V31b |

### 15.6 Exact intermediate commands (r3; r2's commands with the r3 gate additions)

**Variables.** These are fixed at S0 and are not repository content:
- `EXT`: the execution evidence root outside the repository, created by S0 only after the containment check.
- `EXEC`: 6 lowercase hex.
- `BASE`, `SNAP`, `SNAP_SHA`, `COUNT`: Codex's Q9 binding.
- `PATH_RE`: the G003 expression in §15.4.
- `PLAN=management/v0/plans/G003-foundation-observability-ci-plan.md`.
- `ALLOW=management/v0/results/G003-scope-allowlist.txt`.

All commands run from the repository root.

```bash
# Wrapper invocation used by every intermediate and final suite run (scope bound to the exact allow-list).
wrap() { bash infra/validation/run-approved.sh --snapshot "$SNAP" --snapshot-sha256 "$SNAP_SHA" --count "$COUNT" \
  --path-re "$PATH_RE" --plan "$PLAN" -- --base "$BASE" --allow-file "$ALLOW" "$@"; }
lbl=(--label "g003.exec=$EXEC")

# S0 (read-only; any nonzero exit stops before a write)
bash "$EXT/g003-s0-gate.sh" "$PWD" "$BASE" "$SNAP" "$SNAP_SHA" "$COUNT" "$EXT/g003-exec/s0"
bash infra/validation/run-approved.sh --snapshot "$SNAP" --snapshot-sha256 "$SNAP_SHA" --count "$COUNT" \
  --path-re "$PATH_RE" --plan "$PLAN" --check-only
bash infra/validation/run-approved-controls.sh

# S0R (R1; no repository write; A-R1 and A-E required)
git clone -q --no-checkout "$PWD" "$EXT/r1/probe" && git -C "$EXT/r1/probe" checkout -q -b g003-runner-probe "$BASE"
#   write the four §12.4 probe files into $EXT/r1/probe, then:
git -C "$EXT/r1/probe" add .gitlab-ci.yml infra/ci/runner-preflight.sh infra/ci/runner-probe-job.sh infra/ci/probe.allow
git -C "$EXT/r1/probe" commit -q -m "G003 runner qualification probe (disposable; branch deleted after R1)"
git -C "$EXT/r1/probe" push git@gitlab.example:<namespace>/<project>.git g003-runner-probe
#   ... evidence retrieval per A-E ...
git -C "$EXT/r1/probe" push git@gitlab.example:<namespace>/<project>.git --delete g003-runner-probe

# S1 gate (after the S1 writes, the first of which is $ALLOW)
docker build -q "${lbl[@]}" -f infra/docker/intelligence.Dockerfile --target check --build-context contracts=contracts \
  -t "g003-$EXEC-intelligence-check" intelligence
docker run --rm --network none "${lbl[@]}" "g003-$EXEC-intelligence-check"

# S2 (the only network step: module resolution)
mkdir -p "$EXT/gomod" "$EXT/s2"
docker run --rm "${lbl[@]}" --user "$(id -u):$(id -g)" -e HOME=/tmp -e GOCACHE=/tmp/gocache -e GOPATH=/gopath \
  -e GOTOOLCHAIN=local -v "$EXT/gomod:/gopath" -v "$PWD/infra/validation/tools:/src" -w /src golang:1.27.1-alpine \
  sh -c 'go get github.com/santhosh-tekuri/jsonschema/v6@v6.0.3 && go mod tidy'
docker build -q "${lbl[@]}" -f infra/docker/validation-tools.Dockerfile --target check -t "g003-$EXEC-vtools-check" infra/validation/tools
docker run --rm --network none "${lbl[@]}" "g003-$EXEC-vtools-check"
docker build -q "${lbl[@]}" -f infra/docker/validation-tools.Dockerfile --target runtime -t "g003-$EXEC-vtools" infra/validation/tools
docker run --rm --network none "${lbl[@]}" -v "$PWD/contracts:/contracts:ro" -v "$EXT/s2:/out" "g003-$EXEC-vtools" \
  contractcheck corpus --schemas /contracts/intelligence --corpus /contracts/conformance/intelligence --out /out/results-go.json
docker run --rm --network none "${lbl[@]}" -v "$PWD/contracts:/contracts:ro" -v "$EXT/s2:/out" "g003-$EXEC-vtools" \
  contractcheck schema-negative --schemas /contracts/intelligence --corpus /contracts/conformance/intelligence --out /out/schema-negative-go.json

# S3 gate (CE1/CE2 environment files are written to $EXT/s3 with mode 600; canaries from /dev/urandom)
docker build -q "${lbl[@]}" -f infra/docker/core.Dockerfile --target check -t "g003-$EXEC-core-check" core
docker run --rm --network none "${lbl[@]}" "g003-$EXEC-core-check"
docker build -q "${lbl[@]}" -f infra/docker/core.Dockerfile --target runtime -t "g003-$EXEC-core" core
test "$(sha256sum core/cmd/server/main.go | cut -d' ' -f1)" = 987c933560ef77ea9a2daa4a3bec4c9747ea69d578f3c9c9d8e6179d013551ea
for c in ce1 ce2; do docker run --rm --network none "${lbl[@]}" --env-file "$EXT/s3/$c.env" "g003-$EXEC-core" >"$EXT/s3/$c.out" 2>&1; echo "$c rc=$?"; done
cid="$(docker run -d --network none "${lbl[@]}" --env-file "$EXT/s3/ce3.env" "g003-$EXEC-core")"
for i in $(seq 60); do docker logs "$cid" 2>&1 | grep -q '"event":"core listening"' && break; sleep 1; done
docker stop "$cid" >/dev/null; echo "ce3 rc=$(docker wait "$cid")"; docker logs "$cid" >"$EXT/s3/ce3.out" 2>&1; docker rm "$cid" >/dev/null
docker run --rm --network none "${lbl[@]}" -v "$EXT/s3:/in:ro" "g003-$EXEC-vtools" logcheck startup --service core \
  --case /in/ce1.out=1:core.exit:config_invalid --case /in/ce2.out=1:core.exit:listen_failed --exactly-one-line /in/ce2.out \
  --case "/in/ce3.out=0:core listening:port+app_env" --exactly-one-line /in/ce3.out --canaries /in/canaries.txt

# S4 gate
docker build -q "${lbl[@]}" -f infra/docker/web.Dockerfile --target check -t "g003-$EXEC-web-check" web
docker run --rm --network none "${lbl[@]}" "g003-$EXEC-web-check"
docker build -q "${lbl[@]}" -f infra/docker/web.Dockerfile --target runtime -t "g003-$EXEC-web" web
mkdir -p "$EXT/s4/copy" && git --no-optional-locks ls-files -z --cached --others --exclude-standard -- web \
  | tar --null -cf - -T - | tar -xf - -C "$EXT/s4/copy"
for r in wrapped control; do mkdir -p "$EXT/s4/copy/web/app/api/__probe/$r"
  cp "$EXT/s4/copy/web/tests/fixtures/error-probe/$r-route.ts" "$EXT/s4/copy/web/app/api/__probe/$r/route.ts"; done
docker build -q "${lbl[@]}" -f infra/docker/web.Dockerfile --target runtime -t "g003-$EXEC-web-probe" "$EXT/s4/copy/web"
#   WE0-WE6 incl. WE3a (regex-valid canary name) and WE3b: run the probe image (--network none, PROBE_CANARY from $EXT/s4/probe.env), httpprobe in its namespace with
#   $EXT/s4/we.json, docker logs -> $EXT/s4/web-probe.log, logcheck startup/scan; then docker rm -f the probe container.

# S5 gate
docker build -q "${lbl[@]}" -f infra/docker/intelligence.Dockerfile --target check --build-context contracts=contracts \
  -t "g003-$EXEC-intelligence-check" intelligence
docker run --rm --network none "${lbl[@]}" "g003-$EXEC-intelligence-check"
docker build -q "${lbl[@]}" -f infra/docker/intelligence.Dockerfile --target runtime --build-context contracts=contracts \
  -t "g003-$EXEC-intelligence" intelligence
docker run --rm --network none "${lbl[@]}" --env-file "$EXT/s5/ie1.env" "g003-$EXEC-intelligence" >"$EXT/s5/ie1.out" 2>&1; echo "ie1 rc=$?"
#   IE2: docker run -d (valid test env) -> wait for uvicorn.startup_complete in docker logs -> docker stop -> logcheck startup.

# S6 gates (G6a, G6b, G6c, in this order and in the same cycle)
docker run --rm "${lbl[@]}" -v "$PWD:/repo:ro" -w /repo koalaman/shellcheck:v0.11.0 -S warning -x -P SCRIPTDIR \
  infra/ci/validate.sh infra/ci/lib.sh infra/ci/validate-controls.sh infra/ci/cleanup-owned.sh infra/ci/runner-preflight.sh \
  infra/validation/run.sh infra/validation/lib.sh infra/validation/run-approved.sh infra/validation/run-approved-controls.sh \
  infra/local/configure-env.sh
docker run --rm "${lbl[@]}" -v "$PWD:/repo:ro" -w /repo koalaman/shellcheck:v0.11.0 -S info --include=SC2251 -x -P SCRIPTDIR \
  infra/ci/validate.sh infra/ci/lib.sh infra/ci/validate-controls.sh infra/ci/cleanup-owned.sh infra/ci/runner-preflight.sh
bash infra/ci/validate-controls.sh >"$EXT/s6/ci-controls.log" 2>&1; echo "controls_rc=$?"
wrap --report-dir "$EXT/s6/report" static provenance contracts; echo "overall_rc=$?"

# S7
wrap --report-dir "$EXT/s7/report" root migrations variants; echo "overall_rc=$?"

# S8 (local CI parity in a disposable scratch-commit copy; never pushed; deleted afterwards)
mkdir -p "$EXT/s8/clone" && git --no-optional-locks ls-files -z --cached --others --exclude-standard \
  | tar --null -cf - -T - | tar -xf - -C "$EXT/s8/clone"
git -C "$EXT/s8/clone" init -q && git -C "$EXT/s8/clone" add -A
git -C "$EXT/s8/clone" -c user.name=g003-parity -c user.email=parity@example.invalid commit -q -m "G003 local CI parity (scratch)"
(cd "$EXT/s8/clone" && env -u GITLAB_CI bash infra/ci/validate.sh --evidence-dir "$EXT/s8/evidence"); echo "validate_rc=$?"

# S10: §15.4.
```

- **S4 and S5 probe glue.** The comment lines are the steps that B-WEB-ERROR-OUTPUT and B-INTEL-ERROR-OUTPUT later perform inside `run.sh` (§10). At S4 and S5 they are executed with the same `httpprobe` and `logcheck` arguments, and the outputs are kept under `$EXT`.
- **Cleanup.** Every image and container in this block carries `g003.exec=$EXEC`. They are removed at the end with `docker ps -aq --filter label=g003.exec=$EXEC` and `docker image ls -q --filter label=g003.exec=$EXEC`, each checked against the S0 inventory.
- **Writes.** Nothing above writes inside the repository, except the S1–S6 implementation writes themselves and the module resolution of S2 (`go.mod` and `go.sum` under `infra/validation/tools/`, both allow-listed).

## 16. Test strategy and acceptance-criteria mapping (r3; r2 and r1 tables in the appendices)

| AC | Implementation | Commands and assertions (local unless marked hosted) |
|---|---|---|
| **AC1** Health semantics and bounds | §6, §7 | Unchanged from r2: unit tests; B-HEALTH-ROOT; B-WEB-HTTP; B-OBS-TIMEOUTS O1–O10 (each ≤ 5000 ms; recovery ≤ 90 s); B-OUTAGE; B-OBS-OFF; B-TOGGLE-OFF; `obs-timings.tsv`. WE1 keeps 500, empty body and no-store on an unexpected exception. **r3:** CE3 shows that a successful bind logs only the port and `app_env`, and that shutdown on TERM exits 0. |
| **AC2** Context validation, generation and isolation | §4 | Unchanged from r2 |
| **AC3** web → core → intelligence continuity | §4.3, §7 | Unchanged from r2 |
| **AC4** Structured logs and privacy | §5 | **r2 coverage, retained.** **r3:**<br>• the closed web `error_type` category set and the core startup-record keys enforced by `logcheck` and `events.json`;<br>• **CE1–CE3**, with CE2 exactly one line (`core.exit`/`listen_failed`, no `core listening`, no canary) and CE3 exactly one bounded `core listening`;<br>• **WE0–WE6 with WE3a** (a regex-valid canary-bearing `name` → `Error`, canary absent) and WE3b;<br>• IE1 and IE2;<br>• unit tests: Go `exit_test.go` (with `ListenPort`) and `errorlog_test.go`; Node category tests; Python `test_error_output.py`.<br>Source reading (P-NEXT, P-UVICORN) is **not** runtime evidence: WE, CE and IE are mandatory at the first approved runtime gate (S3–S5) and in every final run. **Artifacts:** F3 raw and F7 candidate scans; publication full, minimal or none (§11.2.3). |
| **AC5** Cross-language contracts | §8, §9 | Unchanged from r2 (accepted): B-CONTRACT-TOOLS, -GO, -PY, -AGREE, -PEER and B-CONTRACT-NEGATIVE (SN01–SN07 audit classes in both languages, `sn-derive`, positive audit controls, SD01/SD02, EI39) |
| **AC6** Runnable containerized commands | §10–§13, §15.6 | As in r2. Cleanup is shown by `inventory-leftover.txt` and `cleanup-inventory.tsv`. **r3:** the `main.go` hash check at S3; SC2251 in G6a and B-SHELL. |
| **AC7** GitLab gating with hosted evidence | §11, §12 | **Static:** B-CI-CONFIG (as in r2).<br>**Prerequisites (r3, correction 5):** Q1 facts F-R1–F-R13 and the Q2 named authorizations are **unresolved** and must be supplied before execution approval (§12.6). R1 runs the unchanged baseline cold under the supervisor's watchdog before S1.<br>**Attribution:** unchanged (accepted), with group B controls.<br>**Hosted H1:** a pass on the exact commit and tree; evidence from the **authoritative** job status and final log line, plus the scanned artifact. **Hosted H2:** a failure through the reviewed gate. |
| **AC8** Negative controls and failure propagation | §10, §11.3 | **B-CI-CONTROLS, 80 controls (r4):**<br>• group A (a)–(f): (a) V2, V3; (b) V4 and B-CONTRACT-NEGATIVE; (d) V7, V12; (e) V8, V9; (f) V10, V11 and B-REQUIRED-CHECKS;<br>• group B: attribution;<br>• group C **(c)**: V5a, V5b-planted, V5b-clean, V5c, V6a, V6b, V6c, V6d;<br>• group D: finalization, supervision and location — V14–V18k, including staging after a partial copy (V16), summary (V16b), scrub (V17b), external SIGTERM with `signal=` and no `internal=` (V18), TERM-ignoring harness (V18a), deadline (V18d), state import (V18e), **unexpected parent failure with `harness=143` and `internal=19` (V18f)**, **parent failure after a successful harness with `harness=0` and `internal=19` (V18k)**, pre-location (V18g–V18i), stragglers (V18j);<br>• group E: positive ownership — V19, V20c, V20n, V20v, V20i, V21–V26; **standalone binding** V27, V27b, V28 (interrupted positives), V29, V30 (label-only negatives), V31, V31b (ephemeral); **registry retention** V32–V35 (local and CI layouts; secrets scrubbed; never an empty claim).<br>**Retained:** W1–W9 and C1–C12, C8b and C8c.<br>**Model evidence:** P-MODEL r4 **29 of 29**, with provenance and retention assertions (§15.5b); the strengthened driver fails the r3 model on F15 and F15b (non-vacuity); P-IMPORT 8 of 8; P-OWN r4 with standalone classes. These are planning evidence only; the real controls run at S6 and S10. |
| **AC9** Scope, architecture and provenance | §14, §17 | As in r2: B-SCOPE; B-PROTECTED; B-SKILL-UNCHANGED and B-WEB-DESIGN against the BASE bound after Q14; B-RENDER-PLAN and B-RENDER-REPEAT (19); B-TOGGLE-OFF-ADDITIONS (35, including `main.go`); full regressions. **r3:** B-MCP-SOURCE pins `MAIN_GO_G003_SHA = 987c9335…51ea` (three hunks), with every security assertion retained. |
| **AC10** Documentation, Result and governance | §14, §19, §22 | As in r2. **Prerequisites Q1, Q2, Q14 and Q9 are resolved before execution** (§12.6). The Result separates local (S6–S10), runner-qualification (R1) and hosted (H0–H2) evidence; records the B7 cycles and the checkpoint after cycle 4; and records any `unproven` cleanup rows. |

## 17. Regression reconciliation and documentation labels (r3; r2 text in Appendix R2)

As in r2, with the `main.go` reconciliation updated:
- **`main.go` (Q16, Q21).** Exactly three reviewed hunks (§7.1; intended SHA-256 `987c9335…51ea`):
  - one customization-register row, so B-RENDER-PLAN and B-RENDER-REPEAT expect 19 conflicts and proposals;
  - one allow-list entry;
  - `APPLICABLE` + `core/cmd/server/main.go`, the same rendered base in the disabled variant;
  - B-MCP-SOURCE: byte equality to the rendered manifest becomes byte equality to `MAIN_GO_G003_SHA`. The routes are exactly two; the imports are the rendered set minus `fmt`; `go list -deps` excludes MCP, migration, fixture, isolation and `jsonschema` packages; there is no transport symbol; the reason code is unchanged.
- **The harness `tee` change and capture hooks** (§10) are project-owned harness edits. They are covered by B-SHELL, G6b group D and E controls, and the full static rerun in the same cycle.
- All other G001 and G002 guarantees, accepted deviations, residuals, FOLLOW-UP labels and out-of-scope items are as in r2.

## 18. Security and privacy negative controls (summary; r3)

As in r2, plus:
- **Error categories** are closed and code-controlled (web `instanceof`; core fixed enums). The core startup record carries only a numeric port and a validated `app_env`.
- **Publication** never contains partial or unscanned material (§11.2.3).
- **Cleanup** removes only positively owned resources. Unproven resources are kept and reported.
- **Containment** is checked before any evidence directory is created.
- **Signals** terminate the whole harness process group with a bounded grace. No harness process survives the run (V18, V18a, V18d, V18j).

## 19. Budget, feasibility and checkpoint (B7; r3; r2 text in Appendix R2)

**Budget (unchanged).** B7: at most **7** shared meaningful cycles; checkpoint **after cycle 4 and before cycle 5**. **Zero cycles consumed:** r1, r2 and r3 are planning, and planning models and probes are not cycles. No extension is requested.

**What changed since r2.**
- **Added implementation work:** parent-held state and the importer; the supervisor (`setsid`, watchdog, group stop); the publication-outcome logic; capture hooks and positive-ownership rules; the third `main.go` hunk and CE3; the web category table and WE3a; the one-line harness `tee` change; and 14 more controls.
- **Reduced risk:** each of these control-flow mechanisms has now been exercised by a running model with real signals (24 of 24), an importer unit test and an ownership decision model. The model run found and fixed five defects of the kind Codex flagged (§15.5a), which reduces the chance of discovering them during implementation cycles.

| Likely meaningful cycle | r2 | r3 |
|---|---|---|
| S1–S5 first unit and component failures (context, formatter, uvicorn takeover, Node, exit and error-log adapters, `instanceof` categories) | 1 | 1 |
| Cross-validator or schema-negative class mismatches | 0–1 | 0–1 |
| Error-output conformance on the real frameworks (CE1–CE3, WE0–WE6, IE1–IE2) | 0–1 | 0–1 |
| Integration observability cases | 1–2 | 1–2 |
| Harness and CI reconciliation (static, provenance, groups A–E, supervision, capture hooks) | 1 | 1 (prototyped by P-MODEL and P-OWN; the risk is in Docker-real group E and the capture hooks) |
| First hosted run of the G003 job (H0 or H1) | 0–1 | 0–1 |
| Read-only review findings | 1 | 1 |

**Forecast: 5–7 of 7, most likely 6.** The rows sum to **4–8**, as in r2; reaching 8 needs every uncertain row at its maximum at once. The forecast holds only if:
- **(a)** the Q1 facts confirm Profile A or A2;
- **(b)** A-R1 is authorized and R1 meets its decision rule before S1;
- **(c)** A-H0 is authorized, or the hosted surprise risk is accepted.

**Checkpoint after cycle 4** (BUDGET_POLICY §7). It records per-criterion status, convergence, repeated patterns and remaining work. If the remaining forecast exceeds the 3 remaining cycles, or the same material failure survives two corrective attempts, Claude stops and returns to Codex. Codex may then review a split or the budget.

**r4 note (recurrence).** Codex has now reviewed the finalization and ownership area twice. The r4 changes are localized: parent-side provenance recording, the standalone-container binding, and registry retention. The model reproduces each defect, and its strengthened assertions now pass (29 of 29). The cycle forecast is unchanged, because the added work is a bounded addition to the existing `lib.sh` and `cleanup-owned.sh` scope, not a new mechanism. If any of these mechanisms fails to converge during implementation, the cycle-4 checkpoint and Codex's split reassessment apply.

**Split option (Q10; Codex decides).** Unchanged from r2: G003a (R1–R5, R8, local R6/R7, every control) and G003b (hosted AC7). I recommend the single Goal once Q1, Q2 (A-R1) and Q14 are resolved and R1 passes. If R1 cannot run before S1, the forecast is not credible, and I recommend Codex review the split before execution.

## 20. Delegation

- **One implementation owner:** me. There are no implementation subagents: the work shares the harness, the log contract and the three runtimes.
- **Two read-only reviewers at S10**, run in parallel. Neither edits files. Each reports findings with severity and `file:line`, and the cycles they used (expected 0).
  - **SA-P (correlation and privacy):**
    - the §4 profile against W3C and the Spec;
    - isolation, cancellation and panic paths;
    - completion "exactly one" semantics;
    - redaction parity;
    - framework output;
    - canary coverage and non-vacuity.
  - **SA-C (CI, contracts and provenance):**
    - validator independence and the corpus/expectation design;
    - the Python corrections;
    - `validate.sh` exit semantics and V1–V12;
    - artifact allow-list and scan;
    - B-CI-CONFIG;
    - the register, `APPLICABLE` and the allow-list;
    - the G001/G002 regression coverage.
- **Integration and budget.** I triage and integrate every finding. Corrective work counts against the shared B7. Reviewer opinions do not replace Codex review.

**r2 additions to §20.**
- **Timing.** In the r2 sequence the two reviewers run at **S9**; r1's "S10" above refers to r1's numbering (§15.3 step map).
- **SA-P** also reviews the error-output handling:
  - the bounded events and categories (§5.6);
  - WE0–WE6, CE1–CE2 and IE1–IE2, including the WE4 positive control;
  - the uvicorn, library and warning mapping.
- **SA-C** also reviews:
  - attribution against GitLab's documented variables (§11.2.4, group B);
  - the finalizer and exit-code aggregation (§11.2.3–§11.2.5, groups C and D);
  - the registry and cleanup binding (§11.4, group E);
  - schema-negative isolation (§8.2a–§8.3);
  - the `main.go` reconciliation (§17).
- **Still no implementation subagent.** The r2 revision itself was prepared by a planning subagent (§0.2), which made no implementation change and consumed no cycle.

**r3 additions to §20.**
- **SA-P** also reviews:
  - the closed web category table and WE3a;
  - CE1–CE3 and the bind-then-log hunk;
  - that every startup record is catalogued and none is exempted.
- **SA-C** also reviews:
  - the parent-held state and importer;
  - the supervisor (`setsid`, `group_ready`, watchdog, group stop, signal paths);
  - the publication outcomes and aggregation invariant;
  - the capture hooks and positive-ownership rules;
  - the `tee` change;
  - group D and E controls against P-MODEL and P-OWN.
- **No implementation subagent** is planned. The r3 revision was prepared by a planning agent at the owner's direction (§0.5), with no implementation change and no cycle consumed.

**r4 additions to §20.** SA-C also reviews:
- the parent-side provenance markers (`signal`, `internal`, `harness_stop`) against V18, V18f and V18k;
- `docker_run_owned` and the standalone cidfile binding against V27–V31b;
- registry retention against V32–V35.

The r4 revision was prepared by the same planning agent (§0.7).

## 21. Risks and escalation (r3; r2 and r1 tables in the appendices)

| # | Risk | Handling |
|---|---|---|
| R1 | Runner profile incompatible, unavailable or privileged | **Blocker (Q1).** Facts first (§12.6), then R1 before S1. Never use dind or privileged mode; never use another CI vendor. |
| R2 | uvicorn, a library or `warnings` emits a raw line | `logcheck` fails visibly (IE2, B-OBS-LOGS). Fix inside the application's logging configuration. A `CMD` change is returned to Codex. |
| R3–R12 | As in r2 (Node type stripping, `r.Pattern`, validator disagreement, paused-target counts, cold time, artifact size, gitleaks false positive, harness weakening, budget, client settings file) | As in r2 |
| R13 | Merged-result parent structure differs | Fail closed with 16; a planning correction, as in r2 |
| R14 | Real Next 16.3.8 behaviour differs from the source reading (WE0, WE5, WE6) | Stop and return to Codex. The allow-list is never widened silently. |
| R15 | `SystemExit(3)` under uvicorn misbehaves (IE1) | Return to Codex |
| R16 | The runner ignores the `RUNNER_*` timeout variables | The supervisor's watchdog remains the gate (§11.2.1) |
| R17 | The consolidation is not committed, or makes B-WEB-DESIGN fail | Execution stays blocked (Q14). 0 identical files at r3. |
| R18 | Pushes publish unpushed ancestors | Covered by the authorization text (§12.3) |
| **R19** | The process group does not exist yet when it is signalled (launch race, F25) | `group_ready` before any group signal, with a PID fallback (§11.2.1); model F15 |
| **R20** | The harness is killed before a capture point, so a resource is `unproven` | Kept and reported (18), never deleted by inference. The run is already failing. The owner cleans up from the report (§11.4). |
| **R21** | A `!`-negated statement silently passes under `set -e` (F27) | Explicit returns; SC2251 in B-SHELL and G6a |
| **R22** | Under group TERM, the harness's own cleanup still fails in some path not covered by P-TEE | `validate.sh` F2 is the guaranteed path; group D controls V18 and V18d assert harness cleanup with the r3 `tee`. A failure is a corrective cycle, never a relaxed assertion. |
| **R23** | `net.Listen` before `Serve` changes a startup behaviour that some check relies on | CE2, CE3, B-OUTAGE, B-PORTS-*, readiness and the browser suites re-run at S7 and S10. Any regression is an escalation, not an accepted change. |
| **R24** | A standalone call site is missed and creates a long-lived container without `docker_run_owned` | It cannot be removed (label-only), so it shows as `unproven` → 18 in every run that leaves it. The static B-SHELL review lists every `docker run -d` in `infra/` and requires the helper; a missed site is a corrective cycle, never a label-based deletion. |
| **R25** | A retained local registry accumulates under `TMPDIR` after repeated failed cleanups | Each kept registry is reported in the final line with its path; the owner removes it after manual cleanup. It contains no secrets (the secret list is always scrubbed, V32). |
| **R26** | A marker recorded in a subshell is lost again (the r3 defect class) | `aggregate` and every marker assignment are parent-only. The V18f/V18k assertions require the markers in the final line; the r4 driver's run against the r3 model shows these assertions fail when the marker is lost (§15.5b). |

**Stop and escalate on:**
- any `main.go` change beyond the three reviewed hunks (`987c9335…51ea`);
- any `local.yml`, deadline, schema-byte or Dockerfile `CMD` change;
- any architecture, security-model, vendor, schema, route or API/MCP change;
- weakening any check, assertion or control expectation;
- touching developer resources;
- any unresolved Q1, Q2, Q9 or Q14;
- missing hosted authorization;
- B7 exhaustion or thrashing.

## 22. Questions for Codex and the owner, and unmet prerequisites (r3; r2 and r1 tables in the appendices)

| # | Status | Question or prerequisite | What is needed, and the recommendation |
|---|---|---|---|
| **Q1** | **UNRESOLVED — blocking** | Runner facts F-R1–F-R13 (§12.2) | **The owner supplies them before execution approval is sought** (§12.6): named fields only, never the `config.toml` token. Codex checks Profile A or A2; R1 then verifies. Nothing about the runner is assumed. |
| **Q2** | **UNRESOLVED — blocking** | Named hosted-action authorizations A-R1, A-H0 (optional), A-H1, A-H2, A-H3 and A-E (§12.3) | Explicit, per action, including commits and the publication of unpushed ancestors. Not granted by this Plan, its review or the planning delegation. |
| **Q14** | **OPEN — owner decision; blocks Q9** | Baseline route. The consolidation is **finished** (63 of 63 manifest entries verify) but **uncommitted** (2 modified and 65 new files). | **Route A (recommended):** the owner commits it, and Codex binds BASE to that commit. Routes B and C (r2) still need a reviewed Plan revision and are not planned. |
| **Q9** | **PENDING — blocking** | The fresh binding | Codex, after technical approval of this Plan and the Q14 resolution: BASE, snapshot, its SHA-256, COUNT. `/tmp/g003-r1-…` and `/tmp/g003-r2-reviewed-inputs.sha256` are continuity only, never bindings. |
| Q3–Q8, Q11–Q13 | Resolved by Codex (r1) | — | As recorded (§0.1) |
| Q10 | **Open (Codex)** | Single Goal versus split | Single Goal once Q1, A-R1 and Q14 are resolved and R1 passes (§19) |
| Q15 | Direction accepted (r2 review), corrected in r3 | Web's own empty-body 500 | r3 maps `error_type` through closed `instanceof` categories (§5.1–§5.2 r3 additions) |
| Q16 | Direction accepted (r2 review), corrected in r3 | Bounded `main.go` customization | Now three hunks (§7.1), with the reconciliation in §17 |
| Q17, Q18, Q19 | Directions accepted (r2 review) | Fail-closed profiles; R1 before S1; raw and staged scanning in `validate.sh` | Retained; R1 now uses the supervisor's watchdog (§12.4) |
| **Q20** | Direction accepted (r3 review); **extended in r4** | The project-owned harness changes: the one-line TERM/INT/HUP-ignoring `tee`, the `owned_capture` hooks, the `image <tag> <id>` build rows, and **r4's `docker_run_owned` cidfile binding for standalone containers** (§10, §11.4) | Approve the r4 extension. Without it, standalone probe and preflight containers have no positive ownership and could only be reported, never cleaned. |
| **Q21** | Direction accepted (r3 review) | `main.go` hunk 3 binds with `net.Listen` before `Serve`, and logs only the bound port and `app_env` (§7.1) | No change in r4 |

## 23. Readiness (r4; r3 text in Appendix R3)

**r4: ready for Codex plan review of the three r3 corrections. Execution is blocked by unresolved owner prerequisites.**

- **Blocking (§12.6):**
  - **Q1:** runner facts, unresolved;
  - **Q2:** named hosted-action authorizations, unresolved;
  - **Q14:** baseline route, open (consolidation complete, uncommitted);
  - **Q9:** fresh binding, pending technical approval and Q14.
- **Claude will not seek execution approval** until Q1 and Q2 are supplied.
- **Verified while writing r4:**
  - HEAD `9461c4b`; Goal, Spec and ARCHITECTURE unchanged;
  - S0 re-extracts to `0409148a…6723`;
  - Appendix R1 re-extracts to `a43aa523…dbfc`;
  - Appendix R2 plus the R1 section reproduces r2 `6e6356f3…bc82`;
  - Appendix R3 plus the R2 and R1 sections reproduces r3 `b100cb52…54b2`;
  - the §7.1 hunks replay to `987c9335…51ea`;
  - P-MODEL r4 29 of 29; the r4 driver against the r3 model fails F15 and F15b (non-vacuity); P-IMPORT 8 of 8; P-OWN r4 as expected;
  - the only repository write is this Plan.
- **Not done:** no implementation, installation, initialization, application build, commit, push, pipeline trigger, hosted action or external configuration change.
- **Next step:** Codex review of r4. No implementation may begin.

## Appendix R3 — r3 Plan body, verbatim (superseded wherever r4 differs; not an execution instruction)

This block preserves the reviewed r3 text, as review history.
- **Contents.** It holds r3 lines 1–2079: everything before r3's own "Appendix R2" section. That section and the "Appendix R1" section after it end this file, **byte-identical** to r3.
- **Verification.** The r3 file is exactly this block's lines followed by the file from the "Appendix R2" heading onwards:

```bash
P=management/v0/plans/G003-foundation-observability-ci-plan.md
{ awk '$0=="~~~~~~~~~~~~ r3-body"{f=1; next} $0=="~~~~~~~~~~~~"{f=0} f' "$P"; sed -n '/^## Appendix R2 — /,$p' "$P"; } | sha256sum
# expected: b100cb5266f70a90332afa16e3cffa70391eeed2e43fe170b9326eb7e00f54b2
```

- **Earlier revisions.** The r2 and r1 verifications in the Appendix R2 and R1 sections below still apply unchanged.
- **Precedence.** Where r3 and the body above differ, **the body above governs**. The S0 extraction source for execution is §15.1 of the body above.

~~~~~~~~~~~~ r3-body
# G003 — Health, logging, contracts and baseline CI implementation plan

Goal: [G003-foundation-observability-ci.md](../goals/G003-foundation-observability-ci.md)
Spec: [G003-foundation-observability-ci-spec.md](../specs/G003-foundation-observability-ci-spec.md)
Owner: Claude Code · Reviewer: Codex · Budget class: **B7**: at most 7 shared meaningful cycles; mandatory checkpoint after cycle 4 and before cycle 5. Planning consumed no cycle.

| Revision | Date | Change |
|---|---|---|
| r1 | 2026-10-05 | Initial Plan. Submitted for Codex review. **Not approved; no implementation has begun.** One unmet prerequisite blocks execution (the GitLab runner profile, §12 and Q1); the remaining questions are in §22. — *Appended at r2:* r1 SHA-256 `a43aa523624bab98c38e0fe6c7ff86cffef2cd70d7a3da7b6e806d5c60adbdfc` (1333 lines). Codex review: **CORRECTION LOOP**, six corrections (§0.1). Codex recorded that Q2 and Q9 were also prerequisites, so the "one unmet prerequisite" statement above was incomplete; r2 corrects it. The complete r1 text is preserved verbatim in Appendix R1. |
| r2 | 2026-10-05 | Addresses the six r1 corrections (map in §0.1). **(1)** Runner facts, hosted authorizations and the fresh binding are recorded as unresolved blocking prerequisites (Q1, Q2, Q9), with an exact owner fact list (§12.2), named authorizations (§12.3) and a runner-qualification step R1 that runs the unchanged baseline harness cold on the runner **before S1** (§12.4). The working tree is in an owner-authorized skill consolidation; baseline verification is **PENDING** and the S0 gate is unchanged, with no new exception (§2.1, §15.1a, Q14). **(2)** Checkout attribution uses only documented variables per pipeline type, with 5 positive and 16 negative metadata controls that run before, and independently of, the harness (§11.2, §11.3). **(3)** An explicit phase and finalizer model for `validate.sh`: postcheck, owned cleanup, raw and staged scans and publication run on every exit path; the harness status always survives; scanner and report errors fail closed; a job-unique registry, discoverable by `after_script`, binds owned resources to label-based removal commands (§11.1–§11.4). **(4)** Schema-negative cases run through a separate audit path with declared failure classes and single-mutation derivations; normal runs still reject any unpinned bytes (§8.1a–§8.3). **(5)** Error output is established from the actual frameworks: Next.js 16.3.8 logs a thrown route error with message and stack, so the web wrapper now answers 500 itself; `core/cmd/server/main.go` gets a bounded JSON exit record and a classifying `net/http` error log; intelligence gains bounded startup-failure, uvicorn, library and warning records; three real-artifact checks assert this (§5.3–§5.6, §7, §10). **(6)** An executable step order: the scope allow-list is the first write (S1); CI files, documentation and harness changes are integrated in one step before their first static run (S6); exact intermediate commands are given (§15.3, §15.6). B7 is unchanged. Submitted for Codex review. **Not approved; no implementation has begun. Execution is blocked by Q1, Q2, Q9 and Q14 (§22).** — *Appended at r3:* r2 SHA-256 `6e6356f3bf7e6b0155fe4a32764e1e89d5cf25afa7a96dd0642096fd0220bc82` (3279 lines). Codex review: **CORRECTION LOOP**, five corrections (§0.4). Codex accepted the attribution, schema-negative isolation and step-order work, and accepted the Q15–Q19 directions subject to the corrections. The complete r2 text is preserved in Appendix R2 (body) plus the unchanged Appendix R1 section; the Appendix R2 header gives the hash-verifiable extraction. |
| r3 | 2026-10-05 | Addresses the five r2 corrections (map in §0.4).<br>**(1) State hand-off, signals and location.** Steps that create state (location, harness launch and supervision) run in the parent. Isolated phases return values only through a validated state file. The harness runs as its own `setsid` process group, under a parent-owned watchdog deadline (GNU `timeout` is dropped). Signals stop the whole group with a bounded grace. Containment is checked before any directory is created. A real-signal model passes 24 of 24 cases (§15.5a).<br>**(2) Finalizer consistency.** Publication has exactly three outcomes (full, minimal, none), and staging is all-or-nothing. The artifact summary is labelled pre-publication; the job's exit status and final log line are authoritative. The job-log status columns come from a validated, scanned file that outlives the scrub. `harness-console.log` is now genuinely excluded. The aggregation has a structural no-masking invariant, which fixes the `stage` omission Codex found in r2's model.<br>**(3) Positive ownership.** Networks, volumes and images are removed only with positive ownership: capture rows taken from owned containers, or the run's own build records. Unrelated post-inventory resources of every kind survive (V20c, V20n, V20v, V20i).<br>**(4) Error output.** `main.go` binds before logging and logs only the bound port and `app_env` (exactly three reviewed hunks). CE2 asserts exactly one `core.exit` record; a new CE3 asserts the bounded startup record. Web `error_type` comes from code-controlled `instanceof` categories, with a syntactically valid canary-bearing `name` control (WE3a).<br>**(5) Baseline and prerequisites.** The skill consolidation is now complete but uncommitted; route A is still recommended. Q1 and Q2 remain unresolved and must be supplied before execution approval is sought.<br>B7 unchanged. Submitted for Codex review. **Not approved; no implementation has begun. Execution is blocked by Q1, Q2, Q9 and Q14 (§22).** |

## 0. Review record and planning-session disclosures

### 0.1 r1 → Codex: CORRECTION LOOP (2026-10-05)

Codex reviewed r1 (SHA-256 `a43aa523…dbfc`) at HEAD `9461c4bdbd8c1ed3e9a9986c70647cd1add57101`. Codex confirmed that the embedded S0 script extracts byte-exactly to `0409148a…6723` and passes `bash -n`. That was script-text verification, not execution approval. The review is recorded in PROGRESS.md under `### 2026-10-05 — G003 Plan r1 review: CORRECTION LOOP`. **The review text governs.** The left column below abbreviates it.

| # | Codex correction (abbreviated) | Where r2 addresses it |
|---|---|---|
| 1 | **Runner, hosted authorization and current starting state** (R6/R7, AC7/AC10). Q1, Q2 and Q9 remain prerequisites. Record the runner's identity and tag, executor, isolation and permissions, host, path and network profile, capacity and timeouts, and artifact access. Obtain explicit authorization for the named hosted success and failure actions and for evidence retrieval. Check runner compatibility and cold-build feasibility early enough for the B7 forecast, not first at S12. Reassess or isolate the execution baseline (untracked `vestive-shadcn/**`) without deleting that work or adding a broad unrelated-path exception. A fresh supervisor binding follows approval and a resolved baseline. | §2.1 (current state; baseline **PENDING**), §12.2 (owner fact list, all **UNRESOLVED**), §12.3 (named authorizations A-R1 to A-E), §12.4 (R1 runner qualification before S1, with a decision rule), §15.1a (S0 unchanged; no new exception), §15.3 (S0R), §19 (forecast conditional on R1), §22 (Q1, Q2, Q9, Q14) |
| 2 | **GitLab checkout attribution** (R6/R7, AC7). `CI_MERGE_REQUEST_SOURCE_BRANCH_SHA` is empty in detached merge-request pipelines and set only in merged-results pipelines. Use the variables actually present for each pipeline type. Distinguish detached source, merged result and branch push, without inventing a merged-result claim. Add positive controls for each metadata profile and negative controls for missing or contradictory metadata, producing the intended preflight outcome independently of harness failures. | §3 (F13), §11.2 (attribution rules, cited), §11.3 group B (P1–P5, P3F, N1–N16), §15.5 (P-ATTR prototype: 21 of 21 as expected) |
| 3 | **Aggregate finalization, scanning and owned cleanup** (R7, AC7/AC8). Specify an explicit failure and finalization model: postcheck and safe cleanup reporting run on ordinary failure paths; a nonzero harness status survives; scanner and tool errors and missing or malformed reports fail closed. Resolve the raw-versus-staged scan mismatch. The cleanup registry must be job-unique, discoverable by `after_script`, and bound to the configuration needed for removal (`docker compose -p <name> down` has neither the Compose file nor the environment). Refuse unowned resources and test it. Keep cleanup failures observable. `after_script` runs in a separate shell, cannot fail a successful job and does not run on job timeout by default. | §11.1 (job timeouts and `after_script`), §11.2 (phases, finalizer, exit codes, raw and staged scans), §11.4 (registry and `cleanup-owned.sh`), §11.3 groups C–E, §10 (proposed B-ARTIFACT-SCAN withdrawn; registry hooks), §15.5 (P-FIN, P-TERM and P-SCAN probes) |
| 4 | **Schema-negative isolation** (R5/R7, AC5/AC8). Normal runs must reject bytes that differ from the two pinned hashes. SN02–SN07 must reach the intended validator checks through a disposable negative path, so a shared digest failure cannot make them pass vacuously. Each language asserts the intended classification. Keep an ordinary-run digest-mismatch control and the EI39 control. No network access and no production schema mutation. | §8.1a, §8.2a (declared classes and single-mutation derivations), §8.3 (stage order; normal mode versus audit mode), §10 (B-CONTRACT-NEGATIVE) |
| 5 | **Actual error-output conformance** (R4, AC4). The web wrapper rethrew, and r1 assumed Next prints only startup text. Prove the handling with a real-framework error probe that uses message and stack canaries. Core's protected `cmd/server/main.go` prints `run()` errors with `fmt.Fprintln(os.Stderr, err)`; resolve it, and if `main.go` changes, reconcile the register, scope list, disabled variant and equality guard under R8. Error events need bounded categories or types; arbitrary messages under a renamed `event` key are not enough. | §3 (F15–F19), §5.3, §5.4, §5.6, §7.1–§7.3, §10 (B-CORE-, B-WEB- and B-INTEL-ERROR-OUTPUT), §15.5 (P-GO probe and P-NEXT source evidence), §17 (`main.go` reconciliation), Q15, Q16 |
| 6 | **Executable step order** (R6/R8, AC6/AC9/AC10). S6 ran static checks that read CI files created at S7, and the scope allow-list was created only at S8, although earlier wrapper runs needed it. Reorder so every required file exists when it is first checked; name the exact intermediate commands and scope binding; keep the rule that a shared harness change reruns the affected static and control checks in the same cycle. | §15.3 (new sequence, with an r1→r2 step map), §15.6 (exact commands), §14.3 (the allow-list is created at S1) |

**Planning decisions Codex recorded at r1 (not re-opened by r2):**
- **Q3:** the bounded Python format and anchor corrections enforce the unchanged canonical schemas; they authorize no new schema restriction.
- **Q4:** keep the nested 2 s deadlines, prove optional intelligence is non-gating at core, and disclose the web timeout result.
- **Q5:** scoped to the credentials each service actually holds; web gains no credential access.
- **Q6, Q11, Q12:** the fixture location, the separate pinned test-tool module and the disposable local CI-parity checkout are within scope; no new vendor ADR is needed for the validation library.
- **Q7:** developer-only artifacts with a 90-day expiry are an acceptable intended policy, subject to verifying actual instance and project access and limits (§12.2, F-R11).
- **Q8:** failure of the required job is G003's gate; owner-managed merge and protection settings are not changed.
- **Q13:** the single exact ignored client-settings exception is acceptable, without reading, deleting or broadening it.
- **Budget:** B7 retained as the Goal's ceiling, at most seven shared meaningful cycles, checkpoint after cycle 4; zero cycles consumed. The six-cycle forecast stays conditional on runner compatibility, cold-run capacity and hosted authorization.

### 0.2 r2 planning-session disclosures

- **Delegation.** This revision was prepared by a planning subagent at the product owner's direction, in parallel with a separate owner-authorized skill consolidation run by another agent. That delegation authorizes Plan revision only. It is **not** implementation authorization and **not** hosted-action authorization.
- **Repository writes.** The only repository write is this file.
  - It was revised in place in this checkout, not in a worktree.
  - To keep the r1 text byte-exact (Appendix R1), the r2 file was assembled from a hash-verified scratch copy of r1 and new sections under `/tmp/g003-plan-r2/build/`. It was installed onto this path with one scripted copy, after re-checking that the path still held the r1 bytes. This is the scripted in-place edit the owner authorized.
- **Git.** Only read-only commands were used (`status`, `log`, `rev-parse`, `ls-files`, `diff`), always with `GIT_OPTIONAL_LOCKS=0`. No index, ref, stash, commit or push operation was performed.
- **Not done.** No install, initialization, application build, `docker build`, `docker compose up/build/pull`, harness run, GitLab API call, pipeline trigger or external configuration change.
- **What was run** (scratch only, under `/tmp/g003-plan-r2/`; full list in §15.5):
  - a stdlib-only Go probe (`go run`) inside the existing local `golang:1.27.1-alpine` image (`--pull never --network none`);
  - read-only `grep` of that image's Go 1.27.1 `net/http` source;
  - Bash probes of `errexit`, traps, `timeout` and finalization;
  - Git fixture probes of shallow merge-commit parents and the attribution classifier;
  - a gitleaks v8.30.1 exit-code probe on scratch directories (`--pull never --network none`);
  - `docker image ls` and `docker ps -a` inventory listings.

  The probes created only short-lived `--rm` containers, from images already present. They created no volumes, networks or images.
- **Documentation research.** Official GitLab documentation (predefined variables, merge-request and merged-results pipelines, `CI_PIPELINE_SOURCE` values, `after_script`, runner timeouts). Next.js 16.3.8 and uvicorn 0.54.0 source files at their release tags on GitHub.
- **The other agent's skill files were not inspected as inputs.** They were listed by path, and their bytes were hashed only for the B-WEB-DESIGN identity check (§2.1). Their content was not inspected, modified or bound.
- **No credential, token or environment secret was read or printed.**

### 0.3 r1 planning-session disclosures (verbatim from r1 §0, as submitted)

No Codex review has taken place yet.

Disclosures from this planning session:

- **Writes.** The only repository write is this file.
  - Scratch writes were made only outside the repository, under the job's temporary directory: the S0 gate script, its demonstration clones and their evidence (§15.1–§15.2), and a staging copy of this Plan.
  - The background session's write guard required this Plan to be staged outside the repository first. It is placed at its canonical path only as the single repository write.
- **One Git call with optional locks enabled.** The first baseline check of the session ran an ordinary `git status --porcelain=v1 -uall`. It printed nothing, but because optional locks were enabled it may have refreshed `.git/index` stat data. It cannot have changed content. Every later Git inspection used `git --no-optional-locks` with `GIT_OPTIONAL_LOCKS=0`.
- **Tool-local ignored file.** During the session, at 17:22 local time, the Claude Code client created `.claude/settings.local.json`. It is ignored by the user's global excludes (`~/.config/git/ignore`), and was not created, read or modified by me. It is not a repository input. §15.1 treats it as the single exact ignored-path exception.
- **What I did not do.** I did not install anything, initialize any tool, build any application image, commit, push, trigger a pipeline or change any external configuration.
- **What I did run.** Docker was used only for read-only tooling on scratch files and inventories:
  - `mikefarah/yq` reading the committed manifest;
  - `koalaman/shellcheck` on the scratch S0 gate;
  - inventory and port queries.
- **Remote queries (read-only).**
  - `git ls-remote --symref origin` over SSH, with batch mode on.
  - One unauthenticated GitLab API probe. It returned 404, because the project is private.
- **No credential, token or environment secret was read or printed.**

### 0.4 r2 → Codex: CORRECTION LOOP (2026-10-05)

Codex reviewed r2 (SHA-256 `6e6356f3…bc82`) at HEAD `9461c4b…`. Codex independently verified the unchanged S0 script (`0409148a…6723`) and the verbatim r1 appendix (`a43aa523…dbfc`). The review is recorded in PROGRESS.md under `### 2026-10-05 — G003 Plan r2 review: CORRECTION LOOP`. **The review text governs.** The left column below abbreviates it.

**Resolved by Codex (preserved unchanged in r3):**
- detached-MR attribution, the 21-case prototype and the raw parent check;
- schema-negative audit path, declared classes, single-mutation derivation and positive controls, with SD01/SD02 normal-mode digest rejection;
- S1 creating the allow-list, and S6 creating the CI, docs and harness inputs before the static and provenance gates;
- the unchanged S0 rejecting unrelated uncommitted changes;
- hosted evidence remaining inside G003.

**Directions accepted subject to the corrections:** Q15 (own empty-body 500), Q16 (bounded `main.go` customization), the real production-framework error controls, Q17, Q18 and Q19.

| # | Codex correction (abbreviated) | Where r3 addresses it |
|---|---|---|
| 1 | **Phase state, signals and the read-only location gate** (R7, AC7/AC8). Subshell assignments do not propagate, so the paths from P0, the budget and precheck from P1, and the background PID from P2 never reach the parent. Define behaviour when P0 has no valid location. The default `mktemp -d` creates a directory before containment is checked. Provide disposable evidence of the actual state hand-off and background mechanism: externally signal the supervisor, prove the harness child and process group stop, and prove postcheck and cleanup still run with the intended status. | §11.2.1 (parent-held state; isolated phases plus a validated state import; parent-launched `setsid` harness; watchdog deadline; supervisor signal handling; group readiness), §11.2.2 (L0 location in the parent, containment before creation; pre-location mode), §3 (F24–F27), §15.5a (P-STATE, P-SETSID, P-TEE, P-IMPORT, P-MODEL F15–F22), §11.3 group D (V18–V18j) |
| 2 | **Finalizer, publication and control consistency** (R7, AC7/AC8). F8 gated publication on scans alone, so a partial F5 could be published. F9 deleted evidence before F10 read `results.tsv`. F6 wrote summaries before F7–F9 while §11.2.5 claimed they listed every failure. V5b's "excluded" file was on the staging list. The supplied model omitted `stage` from aggregation; Codex reproduced `stage=19`, a published report and **exit 0**. Correct the model, rerun the faults, and add staging-after-partial-copy, summary, scrub and pre-location failures. | §11.2.3 (all-or-nothing staging; three publication outcomes; status columns retained and scanned; scrub last), §11.2.5 (pre-publication summary versus the authoritative final status; structural no-masking invariant), §11.3 groups C and D (V5b clean and planted; V16, V16b, V17, V17b, V18g–V18i), §15.5a (P-MODEL: 24 of 24 cases, including F11 staging after a partial copy, F12 summary, F14 scrub, F20–F22 pre-location) |
| 3 | **Owned cleanup across resource kinds** (R7, AC7/AC8). The network, volume and image rules relied only on a project label or name plus absence from the start inventory, so an unrelated resource created later under the same label could be removed. Specify positive ownership; reconcile the `<project>-<service>` image fallback with the ban on name-pattern-only removal; split V20 into container, network, volume and image cases, each surviving. | §11.4 (positive ownership rules, capture points, image build and capture records, `unproven` status), §10 (`run.sh` capture hooks), §11.3 group E (V19, V20c, V20n, V20v, V20i, V21–V26), §15.5a (P-OWN decision model) |
| 4 | **Remaining error-output paths** (R4, AC4). CE2 cannot pass under the two-hunk rule: `main.go` logs `core listening` with the raw `cfg.ListenAddr`, carrying CE2's canary, before returning the listen failure. Add a bounded startup-field treatment, reconcile the hunk restriction, guard, register and disabled-variant mapping, and assert exactly one exit event while classifying any permitted startup event. Web copies `Error.name`, which is a writable instance property; map to code-controlled categories and add a syntactically valid canary-bearing `name` control. | §7.1 (exact three-hunk diff: bind first, then a bounded `core listening`), §5.6 (startup record fields; web categories), §5.1–§5.2 r3 additions, §7.2 (`instanceof` categories), §10 (CE1–CE3, WE3a and WE3b), §17 (reconciliation), Q16, Q21 |
| 5 | **Execution prerequisites and refreshed baseline** (R6/R7, AC7/AC9/AC10). Q1 facts and Q2 authorizations remain unresolved, and the proposed profiles and R1 supply neither. Obtain them before seeking execution approval. Update the current state: the consolidation has finished and its 63 kit entries verify, but its 2 modified documents and 65 new files remain uncommitted. Route A remains recommended; a fresh binding follows technical approval and baseline resolution. | §2.1 (refreshed state), §12.2 and §12.3 (unchanged gates, restated), §12.6 (order of prerequisites), §15.1a, §22 (Q1, Q2, Q9, Q14), §23 |

**Budget (Codex):** retain B7 as a ceiling, at most seven shared cycles, checkpoint after cycle 4 and before cycle 5; zero cycles used. Feasibility is conditional on runner qualification and the corrected control flow. Return for split or budget review if the corrected forecast or the qualified runner cannot support all acceptance evidence.

### 0.5 r3 planning-session disclosures

- **Authority.** The owner authorized sending Codex's r2 corrections to this planning agent and revising the Plan. That is not implementation, hosted-action or execution authorization.
- **Repository writes.** The only repository write is this file.
  - It was revised in place, without a worktree.
  - To keep r1 and r2 byte-exact, it was assembled under `/tmp/g003-plan-r3/` from a hash-verified copy of r2 (`6e6356f3…`) and new sections, then installed with one scripted copy after re-checking that the path still held the r2 bytes.
- **Git.** Only read-only commands with `GIT_OPTIONAL_LOCKS=0`: `status`, `rev-parse`, `ls-files`.
- **Not done.** No install, initialization, application build, `docker build`, Compose command, harness run, GitLab API call, pipeline trigger or external configuration change. No Docker container, network, volume or image was created in r3. P-OWN is a pure-Bash decision model over simulated inventories.
- **What was run** (scratch only, under `/tmp/g003-plan-r3/`; listed in §15.5a):
  - Bash and process probes: `setsid`, process-group signalling, `timeout`, `wait`, traps, process-substitution `tee`;
  - a control-flow model of `validate.sh` with stub harness, preflight, cleanup and scanner, driven through 24 fault cases with real signals;
  - a state-importer unit test;
  - the ownership decision model;
  - read-only hashing of the skill consolidation (its manifest check and the B-WEB-DESIGN identity check).
- **Leftover processes.**
  - Two driver-side `sleep 60` guard processes were orphaned by the case driver itself (not by the model) and were terminated.
  - During two aborted model runs, stuck stub processes, caused by a stub-only `wait` deadlock and a watchdog readiness race in an early model version, were terminated by PID and process group.
  - The final run left no model process. All were scratch processes under `/tmp/g003-plan-r3/`; no developer process was touched.
- **The other agent's skill files** were hashed, not inspected or modified.
- **No credential, token or environment secret was read or printed.**

## 1. Summary (r3; r1 and r2 summaries are in the appendices)

G003 completes the existing operational paths. It does not rebuild them. The points below are current.

- **Diagnostic context (R2/R3)** is unchanged (§4): request ID and W3C `traceparent` version-00 validation on the six health routes, per-hop spans, and propagation on exactly two existing calls (web `/api/ready` → core `/readyz` → intelligence `/healthz`).
- **Logs and privacy (R4).**
  - Exactly one JSON completion record per scoped request (§5.2).
  - Application `event` values are code-authored constants. Framework, library and error text is never copied into a record.
  - **r3:** every `error_type` comes from a code-controlled category (§5.1–§5.2 r3 additions, §5.6).
  - **Error output (r2 direction; r3 corrections):**
    - **Web:** the wrapper returns Next's own 500 with an empty body (plus `no-store` and `X-Request-Id`) instead of rethrowing. `error_type` is an `instanceof` category, never the writable `Error.name`.
    - **Core:** `main.go` has exactly three reviewed hunks — a bounded `core.exit` record, a classifying `ErrorLog`, and **bind-then-log**, so `core listening` carries only the bound port and `app_env` and is never emitted when binding fails.
    - **Intelligence:** bounded startup-failure, uvicorn, library and warning records.
    - **Real-artifact checks:** CE1–CE3, WE0–WE6 (with WE3a and WE3b), IE1 and IE2. These are mandatory at the first approved runtime gate; source reading is not runtime evidence.
- **Health (R1)** is unchanged: statuses, bodies and 2 s deadlines; additive headers only; each single request ≤ 5 s.
- **Contracts (R5)** are unchanged from r2, as accepted by Codex: an independent Go validator, the corrected Python validator, normal-mode digest enforcement, the audit-mode schema negatives with declared classes and derivations, and SD01/SD02.
- **CI (R6/R7).** One gating job runs `infra/ci/validate.sh`. **r3:**
  - **State and signals.** Location and harness supervision run in the parent shell. Isolated phases hand values back only through a validated state file. The harness runs as its own `setsid` process group under a parent-owned watchdog deadline. On a signal or the deadline, the supervisor terminates the whole group with a bounded grace, then finalizes.
  - **Location.** It is checked for containment **before** any directory is created; without a valid location, the run takes a read-only pre-location path and exits 16.
  - **Publication.** Exactly one outcome — full, minimal or none. Staging is all-or-nothing. Scanner errors publish nothing. The artifact summary is labelled pre-publication; the authoritative status is the job's exit code and its final log line.
  - **No masking.** A structural invariant makes any recorded failure produce a nonzero exit.
  - **Harness change.** `run.sh`'s `tee` ignores TERM, INT and HUP, so the harness's own cleanup still runs under a group TERM (probe P-TEE).
  - **Cleanup.** Positive ownership for every resource kind: containers by working directory and absence at start; networks and volumes only through capture rows from owned containers; images only through the run's build or capture records. Unproven resources are kept and reported (18).
  - **Attribution** is unchanged from r2, as accepted.
- **Runner and hosted evidence.** Q1 (runner facts) and Q2 (named hosted actions) are **unresolved owner prerequisites** and must be supplied before execution approval is sought (§12.6). R1 (runner qualification before S1) and H0–H3 are unchanged and need their authorizations.
- **Baseline.** The owner-authorized skill consolidation **has finished**: its 63 kit entries verify against its manifest. But its 2 modified documents and 65 new files are **uncommitted**, so baseline verification stays **PENDING**. S0 is unchanged; route A (commit, then bind BASE) is recommended (Q14).
- **Step order** is unchanged from r2, as accepted: the allow-list is the first write (S1), and CI, docs and harness are integrated before their first static run (S6). The S3, S4 and S6 gates add CE3, WE3a/WE3b and SC2251 shellcheck.
- **Not changed.** No database, schema, migration, public domain API, MCP or production-route change. No production execution client, telemetry vendor, metrics endpoint, deployment or UI change. `local.yml`, `core/go.mod`, the intelligence lockfile and the web dependencies are unchanged. The G021 barrier stays closed.

## 2. Verified baseline (2026-10-05, read-only)

| Item | Value |
|---|---|
| HEAD | `9461c4bdbd8c1ed3e9a9986c70647cd1add57101` ("Define G003 operational foundation and CI specification"), branch `staging` |
| Working tree | Clean at the start of the session: `git --no-optional-locks status --porcelain=v1 -z --untracked-files=all --ignored` was empty. Since 17:22 local time it shows one ignored, tool-local path, `!! .claude/settings.local.json` (§0), plus this Plan once placed. |
| Goal SHA-256 | `63f2b26676c9f1c385639c2078abc586beb48e36d6722c957aa4b03b70d73bb2` (matches handoff) |
| Spec SHA-256 | `d64a691d49219e6945b3fb115bcea59977a9b491f7e2746702e329d1c1c2e949` (matches handoff) |
| ARCHITECTURE.md SHA-256 | `6a6fd49b854ce90acf1a2ce6bd281a4deee3e68e8f4e021d8d737c3b04d90726` (matches handoff; §60 and §61 read) |
| Remote | `origin` `git@gitlab.example:<namespace>/<project>.git`. `ls-remote --symref`: `HEAD → refs/heads/main` (`99e2c79`, "Initial planning"), `refs/heads/staging` = `9461c4b` (equal to local HEAD). The default branch is `main`. |
| GitLab project | Private: the unauthenticated API returns 404. No `glab`, `gitlab-runner` or API token is available in this session, so runners, settings, pipelines and artifacts cannot be inspected here (§12). |
| Governance | SECURITY.md and DECISIONS.md are empty; there is no accepted ADR. BUDGET_POLICY, ROADMAP and PROGRESS were read. G001/G002 acceptance is recorded in PROGRESS and supersedes historical submission wording. |
| Schema digests (must remain unchanged) | `execution-envelope.schema.json` `bf3cec365e291da2cf9b201b9d0194632818721671a5844589954b978accba4f`; `stream-event.schema.json` `e1025031134253ac8d19a514f6d7a4d9563281b5006ebe9be40a84e4d63fd102` |
| Protected harness files | `run-approved.sh` `620a3e84…c53c695` and `run-approved-controls.sh` `314a16a7…7ba9` (both unchanged by G003); `run.sh` `d5e7e5be…41ac8`, `lib.sh` `9f7a3997…57c3` and `render-customizations.tsv` `9abf9e92…279f` (all three modified, §10) |
| `core/cmd/server/main.go` | `95aa5b23…7951`. It equals the rendered manifest and stays **unchanged**, so B-MCP-SOURCE's byte-equality guard needs no reconciliation. |
| Local Docker | Engine 27.5.1, Compose 2.32.4 (Docker Desktop on WSL2); 16 CPUs, 15 GiB RAM, 920 GB free. No `beginify` project containers, volumes or networks. Default ports 3100, 8180, 8100, 5532, 6479, 9100 and 9101 are free, as are the variant port sets. Unrelated developer resources exist (for example `payments-assistant-postgres-1`, `outreacher-*`, `needknow-drafter-*` and `atlas-*`/`harbor-*` images) and are never touched. |

### 2.1 r3 current working-tree state (baseline verification **PENDING**; r2's observation is preserved in Appendix R2)

Read-only, 2026-10-05 at about 23:56 UTC, with `GIT_OPTIONAL_LOCKS=0`.

| Item | r3 observation |
|---|---|
| HEAD | Unchanged: `9461c4bdbd8c1ed3e9a9986c70647cd1add57101` |
| Goal, Spec, ARCHITECTURE | Unchanged: `63f2b266…3bb2`, `d64a691d…e949`, `6a6fd49b…0726` |
| PROGRESS.md | `1d023c5ebea2160e4cc556ef11715a32482d77c88581884f73067df6432e71a3`. Modified by Codex's r2 review record; supervisor-owned. |
| Plan | r2 `6e6356f3…bc82` before this edit. The r3 hash is reported at handoff, not self-bound. |
| **Skill consolidation** | **Finished.** `sha256sum -c skills/vestive-design-skill/references/shadcn-kit.sha256` (run from the skill root): exit 0, **63 of 63 OK**. This agrees with Codex's review.<br>**Still uncommitted:**<br>• ` M skills/vestive-design-skill/README.md` and ` M skills/vestive-design-skill/SKILL.md` (2 modified documents);<br>• 63 untracked files under `assets/shadcn/` and 2 under `references/` (65 new resources).<br>`vestive-shadcn/**` is gone. |
| Everything else | The only other differing paths are `management/v0/PROGRESS.md` (` M`) and this Plan (`??`). **No production, harness, dependency or application file differs.** Status counts: 3 modified, 66 untracked, 1 ignored. |
| B-WEB-DESIGN interaction | **0** web files are byte-identical to any of the 63 consolidated asset files (re-checked at r3). Rechecked by the static suite at S6 and S10 against the BASE Codex binds. |
| Ignored | Only `!! .claude/settings.local.json` (Q13) |

**Baseline status: PENDING. Execution cannot start.**
- **S0 is unchanged** (§15.1, `0409148a…6723`). It rejects this tree at step 5 (exit 24), because the skill paths are not snapshot inputs; B-SKILL-UNCHANGED and B-PROTECTED would likewise fail against the current BASE.
- **r3 adds no exception.**
- **Route A remains recommended (Q14).** The owner commits the completed consolidation under their own authorization; Codex then binds BASE to that commit, together with the supervisor files, in the fresh binding (Q9). This review commits nothing.

**r1 row superseded (as noted in r2).** r1 §2 says `core/cmd/server/main.go` "stays **unchanged**". It is now a planned three-hunk customization (§7.1, Q16, Q21). Its current SHA-256, `95aa5b23…7951`, is the rendered base that B-RENDER-PLAN and B-TOGGLE-OFF-ADDITIONS bind.

## 3. Findings that shape the plan

| # | Area | Finding (inspected) | Consequence |
|---|---|---|---|
| F1 | Core correlation | `Middleware` accepts a single regex-valid request ID and a trace ID. It keeps no span or parent and logs the raw `path`. `r.Header.Get` reads only the first of duplicated headers. `statusRecorder` records the **last** `WriteHeader`. `ParseTraceparent` accepts any version except `ff`. | Rework inside `core/internal/observability` (§7.1). `main.go` is unchanged, because the middleware is already wired there. |
| F2 | Core → intelligence | `health.Intelligence` builds a fresh request with no diagnostic headers. Its 2 s client timeout equals core's 2 s readiness deadline. | Propagate from the request context (§7.1). No new route. |
| F3 | Web | The handlers take no `Request`. `coreReady()` sends no context and returns only a boolean. There is no completion logging. Next's standalone server prints only startup lines. | Add `web/lib/request-context.ts`; customize the two routes and `core-client.ts` (§7.2). |
| F4 | Intelligence logging | `uvicorn main:app` keeps uvicorn's default text access log, which writes the raw request line with its query string to stderr, and its default text error log, which writes tracebacks. `JsonFormatter` has no key or value redaction. `/healthz` and `/readyz` send no `Cache-Control`. | Pure-ASGI completion middleware scoped to the two health paths. Take over the `uvicorn.access` and `uvicorn.error` loggers inside `configure_logging`, so the Dockerfile `CMD` is unchanged. Add redaction and no-store (§7.3). |
| F5 | Python format assertion | `jsonschema` 4.26 registers the `date-time` checker only when `rfc3339_validator` imports. It is absent from `uv.lock`, so **`format: date-time` is silently not asserted**. `main.py` already compensates at runtime, but a conformance check would treat the keyword as success. | Register a project-owned RFC 3339 `date-time` checker, with no new dependency (§8.4). |
| F6 | Python `pattern` | `jsonschema` evaluates `pattern` with `re.search`. Python's `$` also matches before a final `\n`, so `"req-12345678\n"` passes `^[A-Za-z0-9._-]{8,64}$` in Python. It fails under ECMA-262 (JSON Schema's dialect) and under Go RE2. | An ECMA-anchored `pattern` keyword in the existing validator, with an audit of the supported syntax (§8.4, Q3). |
| F7 | Python JSON parsing | `json.loads` accepts `NaN` and `Infinity`; RFC 8259 and Go reject them. | Conformance parsing is strict. Runtime envelope parsing is unaffected, because an envelope has no free numeric fields (§8.4). |
| F8 | Go validator candidates | `google/jsonschema-go` (already indirect via the MCP SDK) ignores `format` during validation. `santhosh-tekuri/jsonschema/v6` v6.0.3 (2026-06-28, Apache-2.0) supports Draft 2020-12, `AssertFormat()`, `AddResource`/`UseLoader` and meta-validation, and implements `date-time` and `uuid`. Its `go.mod` requires `golang.org/x/text v0.14.0` (`regexp2` is test-only). | Use santhosh v6.0.3 in a separate test-tool module, so core's `go.mod` is untouched (§9). |
| F9 | Harness assumptions | The harness probes published ports on `127.0.0.1`, bind-mounts host paths (`$REPO`, `$WORK`, report directories) and refuses busy fixed port sets. | The runner must share loopback and filesystem with its Docker daemon. Docker-in-Docker and socket-only Docker executors do not meet this (§12). |
| F10 | Provenance | `contracts/README.md` and `local.yml` are rendered from templates (`.tmpl`) that differ in the intelligence-disabled render. Customizing them would break B-TOGGLE-OFF-ADDITIONS' equal-base rule. | Neither is changed. The corpus lives in `contracts/conformance/` and compose overlays are avoided (§14). |
| F11 | Nested deadlines | Web's `AbortSignal.timeout(2000)` starts before core's 2 s readiness deadline, which bounds its intelligence probe. A *hung* (paused) intelligence therefore makes core answer at about 2.0 s, and web usually aborts first, so web reports `core: unavailable`. A *stopped* intelligence fails fast and is unaffected. | Existing behaviour; the Spec preserves the deadlines. Assertions follow §6.3. Codex decision in Q4. |
| F12 | Ignored-file gate | G002's S0 gate rejects every ignored path. The Claude Code client's ignored `.claude/settings.local.json` now exists in the checkout. | One exact, reported exception in the G003 gate. Every other ignored path, `.env` and `.env.local` included, still fails (§15.1, demonstrations T1–T3). |

**r2 findings.** These are added to F1–F12 above. Where they conflict, the r2 finding governs. In particular, F1's "`main.go` is unchanged" is superseded by F16 and F17, and F3's "Next's standalone server prints only startup lines" by F15. Sources are cited, and the probes are in §15.5.

| # | Area | Finding (inspected or probed) | Consequence |
|---|---|---|---|
| F13 | GitLab attribution | GitLab's [predefined-variables reference](https://docs.gitlab.com/ci/variables/predefined_variables/) documents:<br>• `CI_MERGE_REQUEST_SOURCE_BRANCH_SHA` and `CI_MERGE_REQUEST_TARGET_BRANCH_SHA`: "The variable is empty in merge request pipelines. The SHA is present only in merged results pipelines."<br>• `CI_MERGE_REQUEST_EVENT_TYPE` is `detached`, `merged_result` or `merge_train`.<br>• `CI_COMMIT_BRANCH` is "Not available in merge request pipelines or tag pipelines".<br>• `CI_COMMIT_SHA` is "The commit revision the project is built for".<br>Further:<br>• [Merge-request pipelines](https://docs.gitlab.com/ci/pipelines/merge_request_pipelines/) "Run on the contents of the source branch only".<br>• [Merged-results pipelines](https://docs.gitlab.com/ci/pipelines/merged_results_pipelines/) (Premium/Ultimate) test "a temporary merged commit … This commit doesn't exist in either branch", and fall back to a standard merge-request pipeline on conflict.<br>• The documented `CI_PIPELINE_SOURCE` values are on the [job rules](https://docs.gitlab.com/ci/jobs/job_rules/) page. | r1's `mr-source` rule required HEAD to equal a variable that is empty in exactly that pipeline type. r2 attributes from `CI_COMMIT_SHA` plus the event type, and treats the merged-result SHAs as contradictory in a detached pipeline (§11.2). |
| F14 | `after_script` and timeouts | GitLab's [`after_script`](https://docs.gitlab.com/ci/yaml/#after_script) documentation says it:<br>• runs in a separate shell, with the working directory reset and no access to `script` changes;<br>• has its own timeout, 5 minutes by default (`RUNNER_AFTER_SCRIPT_TIMEOUT`);<br>• cannot fail a job whose `script` succeeded;<br>• "For jobs that time out: `after_script` commands do not execute by default", unless `RUNNER_SCRIPT_TIMEOUT` and `RUNNER_AFTER_SCRIPT_TIMEOUT` together fit within the job timeout ([runner configuration](https://docs.gitlab.com/ci/runners/configure_runners/)).<br>`artifacts:when: always` excludes timeouts. | Gating, scanning and cleanup verification all happen inside `validate.sh`, under an internal deadline below the job timeout. `after_script` is a non-gating backstop only (§11.1, §11.4). |
| F15 | Next.js error output | Next.js 16.3.8 source, at tag `v16.3.8`:<br>• `packages/next/src/build/templates/app-route.ts` rethrows a route-handler error. The outer catch calls `routeModule.onRequestError(req, err, …, silenceLog=false, …)` and then sends `new Response(null, { status: 500 })`.<br>• `route-modules/route-module.ts` `onRequestError` calls `routerServerContext.logErrorWithOriginalStack(err, 'app-dir')` or `console.error(err)`. Both print the error's message and stack. | r1's §5.4 assumption ("route handlers print nothing") was wrong for thrown errors. The wrapper now catches, logs the bounded record and returns that 500 with an empty body itself, plus `Cache-Control: no-store` and `X-Request-Id`. A real-framework probe (B-WEB-ERROR-OUTPUT) proves it (§5.3, §7.2, §10). |
| F16 | Core exit path | `main.go` prints `run()` errors with `fmt.Fprintln(os.Stderr, err)`. Probe P-GO shows a listen error carrying the configured host verbatim: `listen tcp: lookup bad-CANARY-….invalid …`. `config.Load` messages name variables only, never values, but they are still non-JSON free text. | A bounded JSON `core.exit` record with an `error_type` category replaces the print (§5.6, §7.1). `main.go` becomes a reviewed customization (Q16). |
| F17 | Core `net/http` error log | Go 1.27.1 `net/http/server.go` (`grep` of the local `golang:1.27.1-alpine` image) logs these through `Server.ErrorLog`, or the standard `log` package when it is nil:<br>• superfluous and hijacked `WriteHeader`/`Write` calls (with the caller's function, file and line);<br>• an invalid `Content-Length` value;<br>• conflicting length headers;<br>• `http: panic serving <addr>: <value>` with a stack;<br>• TLS handshake and `Accept` errors.<br>Probe P-GO reproduced the plain-text lines, including a canary-bearing panic value. Go 1.27.1 does **not** log semicolon queries (no such call site; the probe produced no line). `http.ErrAbortHandler` is silent. | `main.go` installs `ErrorLog: observability.ServerErrorLog(logger)`. A classifying writer turns each line into one bounded `http.server.error` record and never copies the text. The probe's prototype produced only canary-free JSON (§7.1, §15.5). |
| F18 | Intelligence startup failure | `main:app` is built lazily in the module's `__getattr__`. In uvicorn 0.54.0, `Config.load_app` catches only `ImportFromStringError`, so a `ConfigError` raised by `create_app` escapes as an uncaught traceback: non-JSON, and it carries the exception message. uvicorn's `LOGGING_CONFIG` gives the `uvicorn` logger a stderr text handler (`uvicorn.error` propagates to it) and `uvicorn.access` a stdout handler whose line includes the request line. `server.py` logs a bind `OSError` object directly (`logger.error(exc)`). | `__getattr__` catches startup failures, writes one bounded `intelligence.startup.failed` record and exits with `SystemExit(3)`. The takeover now replaces the `uvicorn` logger's handlers, and uvicorn records are mapped by message *template* to fixed events, with arguments dropped (§7.3). |
| F19 | Python fallbacks | Records from a logger that has no handlers reach `logging.lastResort`, which writes the raw message to stderr. `warnings` writes plain text unless captured. | A single JSON handler on the root logger, with third-party records reduced to `library.log`, plus `logging.captureWarnings(True)` (§7.3). The real-image and full-log checks fail on any remaining raw line (R2 in §21). |
| F20 | Schema-negative masking | r1 mounted the SN files over the pinned schema paths. In normal mode the digest check runs first, so every SN case would have failed on the digest alone. | Separate audit mode, declared classes and a derivation check, while normal mode keeps digest enforcement (§8.3). |
| F21 | Working tree | The owner-authorized skill consolidation is in progress (§2.1). | Baseline PENDING. No S0 change (§15.1a, Q14). |
| F22 | gitleaks exit codes | Probe P-SCAN (gitleaks v8.30.1 `dir`): a clean directory exits 0. A planted generated token exits with the configured `--exit-code 23`. A missing configuration or path exits **1**, the same as gitleaks' default leak code. | Scans use `--exit-code 23`. Exit 0 means clean, 23 means a finding, and anything else is a scanner failure that fails closed (§11.2). |
| F23 | r1 count error | r1 §14.3 listed 18 validation-tool files but said "(17)". | Corrected in §14.3 (r2: 20 files). |

**r3 findings.** These are added to F1–F23. Where they conflict, the r3 finding governs. The probes are in §15.5a.

| # | Area | Finding (inspected or probed) | Consequence |
|---|---|---|---|
| F24 | Phase state | Codex showed, and P-STATE reproduces, that a body run as `( set -e; f )` cannot set parent variables: r2's location phase left `EV` and `REG` unset in the parent, and its `mktemp -d` had already created a directory. | State-creating steps run in the parent. Isolated phases return values only through a validated state file (§11.2.1). |
| F25 | Process groups | P-SETSID: in a non-interactive shell, `setsid CMD &` keeps `$!` as the PID and makes it the process-group and session leader (`pid = pgid = sid`), so `kill -- -$!` reaches the whole tree. The model's F15 first run showed the group is created **asynchronously**: a signal sent immediately after launch found no group and missed the tree. | A bounded `group_ready` wait follows each launch, and every stop path uses it (§11.2.1). |
| F26 | Harness cleanup under signals | P-TEE, with a stub shaped like `run.sh` (`set -euo pipefail`, `trap cleanup EXIT`, `exec > >(tee -a LOG) 2>&1`): on a group TERM the harness died of SIGPIPE (141) **before** its cleanup, because `tee` died first. When `tee` ignores TERM, INT and HUP, the cleanup ran every time and the harness exited 143. Under GNU `timeout`, a variant with process substitution also skipped its EXIT trap. | `run.sh` changes one line: its `tee` ignores TERM, INT and HUP. GNU `timeout` is replaced by a supervisor-owned watchdog. The harness's own cleanup is best-effort; validate.sh's positive-ownership cleanup is the guaranteed path (§10, §11.2, §11.4). |
| F27 | `errexit` pitfall | The r3 model's first run passed a failing preflight: `! grep …` as a statement is exempt from `set -e` in Bash. | Every check in `validate.sh`, `lib.sh` and `cleanup-owned.sh` uses `if …; then return N; fi`. B-SHELL and G6a add `shellcheck --include=SC2251` for `infra/ci/*.sh` (§10). |
| F28 | Core startup record | `main.go` starts `ListenAndServe` in a goroutine and immediately logs `core listening` with `slog.String("addr", cfg.ListenAddr)`. A failing address is therefore logged raw before `core.exit` (Codex). | Bind first with `net.Listen`; log `core listening` only after a successful bind, with the numeric bound port and the validated `app_env` (§7.1). |
| F29 | Web error category | `Error.name` is a writable instance property, so a name that matches a syntax regex is still data (Codex). | `instanceof` mapping to a fixed category set (§5.1–§5.2 r3 additions) |
| F30 | Cleanup ownership | Network, volume and image rules relied on label or name plus absence at start (Codex). | Positive ownership through capture and build records (§11.4); P-OWN models the decisions. |
| F31 | Working tree | The consolidation is finished but uncommitted (§2.1). | Baseline PENDING; route A (Q14) |

## 4. Diagnostic context contract (R2)

### 4.1 Request ID

- **Accepted inbound value.** Exactly one `X-Request-Id` field value matching `^[A-Za-z0-9._-]{8,64}$`, evaluated with full-string semantics:
  - Go RE2 `MatchString`;
  - JavaScript without the `m` flag;
  - Python `re.fullmatch`.
- **Rejected inbound value.** A missing value, more than one field line, a comma-joined value (Node joins duplicated fields as `a, b`), an over-long or non-matching value, or a value that contains a credential the receiving service holds (§4.4).
  - A rejected value is **never reflected or logged**.
  - Instead the service generates `request_id = 32 lowercase hex` from 16 random bytes (`crypto/rand`, `node:crypto` `randomBytes`, `secrets.token_hex`), and records `request_id_source` as `generated` rather than `inbound`.
- **Response header.** All six scoped health responses carry the effective `X-Request-Id`. Core already returns it on every response.

### 4.2 `traceparent` profile (narrower than W3C; no full-compliance claim)

Based on the [W3C Trace Context Recommendation](https://www.w3.org/TR/2021/REC-trace-context-1-20211123/).

- **Accepted.** Exactly one field value, at most 512 bytes before parsing, that fully matches `^00-([0-9a-f]{32})-([0-9a-f]{16})-([0-9a-f]{2})$`, where:
  - the trace ID is not all zeros;
  - the parent ID is not all zeros;
  - the length is exactly 55, which follows from the pattern.
- **Flags.** Any two lowercase hex digits are accepted. Only the `sampled` bit (`flags & 0x01`) is interpreted. Unknown bits are ignored and **never propagated**: the outbound flags are `01` when sampled is set, otherwise `00`. A newly generated trace uses `00`. The flag confers no authority, sampling, content capture or export.
- **Versions.**
  - `ff` is invalid.
  - **Every version other than `00` is unsupported in this profile.** The service restarts with a new local trace. This is narrower than W3C's SHOULD-parse rule for higher versions, and is documented as such.
  - Uppercase hex, the wrong length, extra trailing fields on `00`, internal whitespace, a non-hex character, a duplicated field or a comma-joined value are all invalid.
- **On an invalid value.** The service generates a new nonzero 16-byte trace ID. The invalid value is never reflected, logged or forwarded, no HTTP 500 is produced, and `trace_source` is recorded as `generated`.
- **Not supported:** `tracestate` and `baggage`. They are never read, forwarded or logged.

### 4.3 Spans and parent relationship (server-span-as-caller convention)

- **Per-hop span.** Every receiving service creates a fresh nonzero 8-byte `span_id` for its server operation. It regenerates if the value is all zeros, and it never copies the incoming parent ID.
- **Parent.** `parent_span_id` is the inbound `traceparent` parent ID, present only when that `traceparent` is valid.
- **Outbound.** Each outbound call carries `traceparent: 00-<trace_id>-<caller span_id>-<flags>`. The parent ID is "the ID of the current operation", as W3C requires when mutating `traceparent`. No separate client span is recorded.
- **Resulting chain on the full stack:**
  - `web.parent_span_id` = the caller's parent, absent when the context was generated;
  - `core.parent_span_id` = `web.span_id`;
  - `intelligence.parent_span_id` = `core.span_id`.

  All three records share `trace_id` and `request_id`, and their span IDs are distinct.
- **Datastores.** Probes to PostgreSQL, Redis and MinIO carry no tracing protocol.

### 4.4 Credentials supplied as diagnostic values (Q5)

- **Core** treats an inbound request ID that contains any configured secret value (`Config.SecretValues()`, reached through the redacting logger's handler) as invalid, and generates a new one. Core never echoes, logs or forwards that ID. Its logger additionally scrubs secret values from every string (existing behaviour).
- **Intelligence** does the same for `INTELLIGENCE_SERVICE_TOKEN` and `OPENROUTER_API_KEY`, and its formatter scrubs them (§7.3).
- **Web** holds no credentials, so it cannot classify a value as one. It logs only regex-bounded IDs. Core refuses to adopt or forward such a value.

### 4.5 Header allow-list and isolation

- **Outbound headers set by application code:** exactly `X-Request-Id` and `traceparent`.
  - Transport defaults remain: Go `User-Agent` and `Accept-Encoding`, and undici's standard request headers.
  - No inbound header, cookie, `Authorization`, `baggage` or `tracestate` is ever copied.
- **Request-local state:**
  - **Core:** values stored in the per-request `context.Context`.
  - **Web:** an explicit `RequestContext` argument. There is no module-level or `AsyncLocalStorage` state.
  - **Intelligence:** local variables inside the ASGI call. There is no `ContextVar`, because the health handlers make no outbound call.
- Isolation therefore holds by construction. It is still tested under concurrency, failure and cancellation (§16, AC2).

## 5. Log contract (R4)

### 5.1 Envelope for every application-owned JSON line (all three services)

| Key | Type | Rule |
|---|---|---|
| `time` | string | RFC 3339 UTC, millisecond precision, `Z`. Go: slog `ReplaceAttr`; Node: `toISOString()`; Python: `datetime.fromtimestamp(created, UTC)` |
| `level` | string | one of `debug`, `info`, `warn`, `error`. Python `WARNING` → `warn`, `CRITICAL` → `error` |
| `service` | string | `web`, `core` or `intelligence` |
| `event` | string | Core: slog's message key is renamed to `event`. Intelligence: `msg` becomes `event`; the existing `logger` key is kept. |

Startup and error records keep their messages (for example core's `core listening` and intelligence's `configured`) under `event`.

### 5.2 Completion record (`event = "http.request.completed"`)

| Key | Type | Rule |
|---|---|---|
| `method` | string | `GET`, `HEAD`, `POST`, `PUT`, `PATCH`, `DELETE` or `OPTIONS`; anything else becomes `OTHER` |
| `route` | string | Normalized, never the raw path:<br>• core: the matched `ServeMux` pattern path (`/healthz`, `/readyz`), else `unmatched`;<br>• web: the fixed handler route (`/api/health`, `/api/ready`);<br>• intelligence: `/healthz`, `/readyz` |
| `status` | integer | The final HTTP status: the first `WriteHeader`, or an implicit 200 on the first write. `0` only when the connection aborts before headers. |
| `duration_ms` | integer ≥ 0 | Monotonic elapsed time, floored |
| `request_id`, `request_id_source` | string, enum | §4.1 |
| `trace_id`, `trace_source` | 32 hex nonzero, enum | §4.2 |
| `span_id` | 16 hex nonzero | §4.3 |
| `parent_span_id` | 16 hex nonzero | **Present only** when the inbound parent was valid |
| `trace_flags` | `00` or `01` | The outbound flags (§4.2) |
| `outcome` | enum | `completed`, `cancelled` (client gone or context cancelled) or `failed` (panic or exception) |
| `error_type` | string `^[A-Za-z][A-Za-z0-9_.]{0,63}$` | Only when `outcome = failed`. Go: `panic`; Node: the error's `name`; Python: the exception class name. Never message text or a stack. |
| `core_probe` | enum | **Web `/api/ready` only:** `ok`, `not_ready`, `timeout` or `unreachable` |

- **Level.** `failed` → `error`; status 500 or above → `warn`; otherwise `info`.
- **Strictness.** No other key is allowed, apart from intelligence's `logger`.
- **Machine-checked contract.** The record is checked by `infra/validation/tools/internal/logrecord/completion.schema.json`, through the `logcheck` command (§9).

**r3 additions to §5.1–§5.2 (corrections 4 and 5; they supersede r2's additions, preserved in Appendix R2).**
- **Constant events** (unchanged from r2). An application-owned `event` value is a constant authored in repository code: a template, never formatted with runtime values. `logcheck` validates every `event` against the per-service catalogue `infra/validation/tools/internal/logrecord/events.json`. Any uncatalogued value fails.
- **Web `error_type` is a code-controlled category, never instance data.** The wrapper maps a thrown value with `instanceof` checks against built-in constructors, in this fixed order:
  1. a `DOMException` whose `name` is exactly `AbortError` or `TimeoutError` → that name; any other `DOMException` → `DOMException`;
  2. `TypeError`, `RangeError`, `SyntaxError`, `ReferenceError`, `EvalError`, `URIError`, `AggregateError` → the matching constructor's name, **as written in the mapping table, not read from the value**; a subclass reports its built-in ancestor;
  3. any other `Error` → `Error`;
  4. a non-`Error` value → `NonError`.

  The value's `name`, `message`, `cause`, stack and own properties are never read into the record. The category set is closed, and `logcheck` enforces it through `events.json`.
- **Core and intelligence `error_type`.**
  - **Core** uses only the fixed categories in §5.6 (`panic` for handler panics).
  - **Intelligence** uses `type(exc).__name__`. That is taken from the exception's class (code), never from instance attributes, and is validated against `^[A-Za-z][A-Za-z0-9_.]{0,63}$`; `Exception` otherwise.
- **Startup, exit and framework records** use the same §5.1 envelope (§5.6).

### 5.3 Exactly one record (r2; supersedes r1 §5.3 in Appendix R1)

- **Core** is unchanged from r1.
  - The middleware logs once in a `defer` guarded by a flag.
  - A handler panic is logged with `outcome: failed`, `error_type: panic`. The middleware then re-panics with `http.ErrAbortHandler`, which `net/http` neither prints nor passes to `ErrorLog` (F17). The connection aborts as before.
  - A request whose context was cancelled is logged with `outcome: cancelled`.
- **Web** (changed in r2; F15). `withCompletionLog` logs exactly once, in `finally`.
  - **Normal path.** The handler's own response is returned, with `X-Request-Id` added.
  - **On an exception** the wrapper does **not** rethrow. It records `outcome: failed` and the bounded `error_type`, and returns `new Response(null, { status: 500, headers: { "Cache-Control": "no-store", "X-Request-Id": <effective id> } })`.
    - This is the status and empty body that Next 16.3.8 itself sends after logging, so the only observable differences are the two additive headers (Q15).
    - Because no error reaches Next, Next prints nothing.
  - **On abort.** When `request.signal.aborted` is set at completion, the wrapper records `outcome: cancelled`, with the same no-rethrow rule.
- **Intelligence** is unchanged from r1.
  - The ASGI middleware logs once in `finally`.
  - An exception becomes `failed` with its class name, then is re-raised, so uvicorn's "Exception in ASGI application" is emitted as the bounded `uvicorn.asgi_exception` record (§5.6).
  - `CancelledError` becomes `cancelled`, then is re-raised.
- **Scope** is exactly the six health routes, as in r1.
  - Core keeps logging every request it receives, including 404 and 405, with `route: unmatched`.
  - Intelligence's `/internal/*` path keeps its execution export record and no longer has an access line (F4).
- **Duplicate writes.** The recorder forwards every call unchanged, as before, but reports the first status.

### 5.4 Framework output (r2; supersedes r1 §5.4 in Appendix R1)

`logcheck` classifies **every** line of each service's raw output, in both stacks and in the error-output checks. Any line that is not valid JSON with a catalogued event, and does not match the exact allow-list below, fails.

| Service | Actual framework behaviour (source) | G003 handling | Allowed non-JSON lines |
|---|---|---|---|
| core | `net/http` writes `ErrorLog` lines in plain text; panics other than `ErrAbortHandler` print value and stack (F17, probe P-GO). `main()` prints `run()` errors raw (F16). | `ErrorLog` is the classifying adapter (`http.server.error` with a fixed `error_type`). `main()` writes one `core.exit` record (§5.6). Panics in scoped handlers are re-raised as `ErrAbortHandler`. | none |
| web | Next 16.3.8 prints its startup banner. For a thrown route error it calls `console.error(err)` or `logErrorWithOriginalStack`, which print message and stack (F15). | Route handlers never let an error escape the wrapper (§5.3). | Exactly r1's startup patterns: `^ ▲ Next\.js [0-9]+\.[0-9]+\.[0-9]+$`, `^   - Local:        http://[^ ]+$`, `^   - Network:      http://[^ ]+$`, `^ ✓ Starting\.\.\.$`, `^ ✓ Ready in [0-9.]+m?s$`, and the empty line. They are confirmed against the observed output of the real-framework probe (WE0, §10); any deviation is recorded and returned for review, and the list is never widened silently. |
| intelligence | uvicorn 0.54.0 sends text through the `uvicorn` logger's stderr handler and access lines (with the request line) to stdout. `load_app` lets non-import errors escape as tracebacks. Unhandled loggers fall back to `logging.lastResort`; `warnings` writes text (F18, F19). | `configure_logging` takes over the `uvicorn` logger, disables `uvicorn.access`, adds the root JSON handler and captures warnings. `main.__getattr__` turns a startup failure into one bounded record and `SystemExit(3)` (§7.3). | none |

Runtime crashes outside request handling, for example a Go `fatal error` or a Python interpreter abort, are not request or startup paths G003 controls. If one occurs in a tested run, `logcheck` fails it as an unclassified line; it is reported, never suppressed.

### 5.5 Redaction (existing protections preserved; equivalents added)

- **Core (existing, retained):**
  - sensitive-key regex `password|passwd|secret|token|api_?key|authorization|cookie|dsn|credential`, including nested groups;
  - configured-secret-value scrubbing in the message and in string and non-string attributes.
- **Intelligence (new, in `JsonFormatter`):**
  - the same key regex, applied recursively to `fields`;
  - configured-secret-value scrubbing of `event` and of every string;
  - exception type only (existing);
  - string values capped at 512 characters.

  `JsonFormatter()` keeps its no-argument constructor, because `test_observability.py` uses it. Secrets are supplied through `configure_logging(level, secrets=...)`, which `create_app` calls.
- **Web.** Records are built only from the fixed keys in §5.2, with values that are validated IDs, enums or integers. A key-regex guard drops any sensitive-looking key. No URL, header, error message or configuration value is ever logged.
- **Never logged anywhere:**
  - raw headers, cookies or query strings;
  - request or response bodies;
  - prompts, completions, retrieved content or envelopes;
  - authorization references;
  - tenant or user identifiers;
  - URLs, DSNs or exception text.

### 5.6 Startup, exit and framework error records (r3; r2 version in Appendix R2)

All of these use the §5.1 envelope (`time`, `level`, `service`, `event`). They never contain message text, stack frames, argument values, host names, URLs, addresses or configuration values, apart from the single numeric port and validated `app_env` named below.

| Service | `event` | Extra keys (bounded) | Emitted when |
|---|---|---|---|
| core | **`core listening`** (level `info`; the existing constant) | **r3:** exactly `port` (an integer from the **bound** listener's `*net.TCPAddr`; `0` if not TCP) and `app_env` (`development`, `test` or `production`; validated by `config.Load`). The configured listen address and host are **never** logged. | Only **after** `net.Listen` succeeds (§7.1, hunk 3). A failed bind produces no `core listening` record. |
| core | `core.exit` (level `error`) | `error_type`, one of `config_invalid`, `database_pool_invalid`, `listen_failed` (`*net.OpError` with `Op == "listen"`, now returned directly by `net.Listen`), `serve_failed`, `shutdown_timeout`, `unknown`. `config_invalid` adds `config_keys` from the allow-list derived from `config.Load` (§7.1). | `run()` returns an error; exit status 1 |
| core | `http.server.error` (level `warn`) | `error_type` ∈ {`superfluous_write_header`, `hijacked_write_header`, `hijacked_write`, `invalid_content_length`, `conflicting_length_headers`, `handler_panic`, `tls_handshake_error`, `accept_error`, `other`} | `net/http` writes to `Server.ErrorLog` |
| core | `http.request.completed` | §5.2; `error_type: panic` only | §5.3 |
| intelligence | `intelligence.startup.failed` (level `error`) | `error_type`: `type(exc).__name__` (class, regex-validated) | `create_app()` raises while uvicorn builds `main:app`; exit 3 |
| intelligence | `uvicorn.process_started`, `uvicorn.startup_waiting`, `uvicorn.startup_complete`, `uvicorn.listening`, `uvicorn.shutting_down`, `uvicorn.shutdown_waiting`, `uvicorn.shutdown_complete`, `uvicorn.process_finished`, `uvicorn.app_load_failed`, `uvicorn.invalid_request`, `uvicorn.asgi_exception`, `uvicorn.server_error`, `uvicorn.other` | `error_type` (class) when exception information is present | Mapped by **template** (`record.msg`), with arguments dropped (as in r2). `uvicorn.listening` carries **no** host or port: the template's arguments are dropped. |
| intelligence | `library.log`; `python.warning` | `logger` (regex-validated name), or `category` (warning class) | As in r2 |
| web | `http.request.completed` | §5.2; `error_type` from the closed `instanceof` category set (§5.1–§5.2 r3 additions) | §5.3 |

**Catalogued startup events per service** (permitted, classified, never exempted):
- core: `core listening` with exactly `port` and `app_env`;
- intelligence: the `uvicorn.*` lifecycle events and `configured`;
- web: none (Next's startup banner stays the exact non-JSON allow-list in §5.4).

Every other startup line fails `logcheck`. The canary checks (CE1–CE3, IE1, IE2, WE0) scan these records like any other line.

Intelligence's own logger records keep `event = record.msg`, as in r2.

## 6. Health semantics and bounds (R1)

### 6.1 Endpoints (bodies and status unchanged; additive headers only)

| Service | Route | Body (exact bytes) | Status | Headers after G003 |
|---|---|---|---|---|
| core | `GET /healthz` | `{"status":"alive"}\n` | 200 | `Content-Type: application/json`, `Cache-Control: no-store`, `X-Request-Id` (all existing) |
| core | `GET /readyz` | `{"status":…,"product_name":…,"checks":{"object_store":…,"postgres":…,"redis":…}[,"optional":{"intelligence":…}]}\n` | 200 / 503 | as above (existing) |
| web | `GET /api/health` | `{"status":"alive"}` | 200 | `Cache-Control: no-store` (existing), **`X-Request-Id` (new)** |
| web | `GET /api/ready` | `{"status":"ready"\|"unready","checks":{"core":"ok"\|"unavailable"}}` | 200 / 503 | `Cache-Control: no-store` (existing), **`X-Request-Id` (new)** |
| intelligence | `GET /healthz` | `{"status":"ok"}` | 200 | **`Cache-Control: no-store` and `X-Request-Id` (new)** |
| intelligence | `GET /readyz` | `{"status":"ready"}` (also when no model is configured) | 200 | **`Cache-Control: no-store` and `X-Request-Id` (new)** |

- Core's required checks remain `postgres`, `redis` and `object_store`. `intelligence` stays under `optional` and is absent when intelligence is disabled.
- The deadlines remain:
  - core readiness: 2 s;
  - core's intelligence client: 2 s;
  - web's fetch: 2 s.
- No error text, DSN, URL or credential appears in any body.
- Intelligence's body bytes are unchanged: FastAPI's default dict response and `JSONResponse` render identically (`separators=(",", ":")`).

### 6.2 Timing protocol

- **Clock.** Each single request is timed by `httpprobe` (§9) with Go's monotonic clock.
- **Where it runs.** Inside the target service's network namespace (`--network container:<cid>`): web on `127.0.0.1:3000`, core on `127.0.0.1:8180`, intelligence on `127.0.0.1:8000`.
- **Bound.** Every injected timeout or outage case must return in **≤ 5000 ms**, or the criterion fails.
- **Recovery.** Recovery polling is separate. It has a documented finite deadline of **90 s per case** and is never used to time a single request.

### 6.3 Outage and timeout matrix (disposable copies only)

The cases run on the enabled copy (`variants_ops`, ports 3200, 8280, 8200, 5632, 6579, 9200 and 9201) and on the disabled copy (`variants_off`).

| Case | Injection | Assertions (each single request ≤ 5 s) | Enabled | Disabled |
|---|---|---|---|---|
| O1–O3 | `compose stop` postgres / redis / minio | core `/readyz` 503 naming the dependency; web `/api/ready` 503 (`core_probe: not_ready`); both liveness probes 200; continuity records (§16 AC3); recovery to 200 | ✓ | ✓ |
| O4–O6 | `docker pause` postgres / redis / minio | Same statuses; web `core_probe` is `not_ready` or `timeout`, because core answers at about 2 s (F11); continuity in web and core; unpause, then recovery | ✓ | ✓ |
| O7 | stop intelligence | core 200 with `optional.intelligence = unavailable`; web 200; no intelligence record; recovery restores the three-hop chain | ✓ | — |
| O8 | pause intelligence | **Core `/readyz` (direct) 200 with `optional.intelligence = unavailable`.** Web `/api/ready` returns within the bound with either 200 or 503 `core_probe: timeout`. The observed value is recorded (F11, Q4). After unpause, intelligence may log 0 or 1 record for the queued probe; more than one, or a mismatched one, fails. | ✓ | — |
| O9 | stop core | web `/api/ready` 503 (`core_probe: unreachable`), `/api/health` 200, both no-store; only a web record | ✓ | ✓ |
| O10 | pause core | web `/api/ready` 503 (`core_probe: timeout`). After unpause, core logs at most one record for that request ID, with `outcome: cancelled` and matching IDs. | ✓ | ✓ |

The existing B-OUTAGE and B-PERSIST then run unchanged on the same copy, after recovery has been asserted.

## 7. Service changes (r3; r2 §7 is in Appendix R2, r1 §7 in Appendix R1)

### 7.1 Core (`core/internal/observability`, `core/internal/health`, and `core/cmd/server/main.go`, which is customized)

- **`correlation.go`, `logging.go`, `checks.go`** are unchanged from r2 (§7.1 of r2: the context profile, per-hop spans, completion records, the `ReplaceAttr` envelope and outbound propagation).
- **`exit.go` (new)** contains:
  - `LogExit(w io.Writer, err error)`, which writes exactly one `core.exit` record (§5.6) and never `err.Error()`;
  - the `config_keys` allow-list;
  - **`ListenPort(a net.Addr) int`**, which returns `a.(*net.TCPAddr).Port`, or 0.
- **`errorlog.go` (new):** `ServerErrorLog(logger *slog.Logger) *log.Logger`, the classifying writer (§5.6), as in r2.
- **`core/cmd/server/main.go` (customized; Q16, Q21): exactly three unified-diff hunks (3 lines of context), shown below.**
  - Hunk 1 drops the unused `fmt` import.
  - Hunk 2 replaces the raw exit print.
  - Hunk 3 installs `ErrorLog` and **binds before logging**.

  The intended file's SHA-256 is **`987c933560ef77ea9a2daa4a3bec4c9747ea69d578f3c9c9d8e6179d013551ea`**. It was produced from the current file (`95aa5b23…7951`) in a scratch copy and is gofmt-clean (`gofmt` in `golang:1.27.1-alpine`, `--network none`; not compiled, because the module's dependencies are not downloaded in planning). This value is `MAIN_GO_G003_SHA` in B-MCP-SOURCE (§10). Any other byte change is an escalation (§21).

```diff
@@ -8,7 +8,6 @@
 	"context"
 	"errors"
 	"flag"
-	"fmt"
 	"log/slog"
 	"net"
 	"net/http"
@@ -33,7 +32,7 @@
 		os.Exit(healthcheck(os.Args[2:]))
 	}
 	if err := run(); err != nil {
-		fmt.Fprintln(os.Stderr, err)
+		observability.LogExit(os.Stderr, err)
 		os.Exit(1)
 	}
 }
@@ -107,10 +106,15 @@
 		Addr:              cfg.ListenAddr,
 		Handler:           observability.Middleware(logger, mux),
 		ReadHeaderTimeout: 5 * time.Second,
+		ErrorLog:          observability.ServerErrorLog(logger),
+	}
+	ln, err := net.Listen("tcp", cfg.ListenAddr)
+	if err != nil {
+		return err
 	}
 	errCh := make(chan error, 1)
-	go func() { errCh <- srv.ListenAndServe() }()
-	logger.Info("core listening", slog.String("addr", cfg.ListenAddr), slog.String("app_env", cfg.AppEnv))
+	go func() { errCh <- srv.Serve(ln) }()
+	logger.Info("core listening", slog.Int("port", observability.ListenPort(ln.Addr())), slog.String("app_env", cfg.AppEnv))
 
 	select {
 	case <-ctx.Done():
```

  **Behaviour.**
  - A failing bind now returns the `*net.OpError` from `net.Listen` before any `core listening` record exists, so the run emits exactly one `core.exit`/`listen_failed`.
  - A successful bind logs `core listening` with the bound port and `app_env` only.
  - `Serve(ln)` and `Shutdown` keep the existing semantics. `http.ErrServerClosed` is still the normal shutdown result.
  - Routes, the remaining imports, timeouts, the health wiring, `healthcheck` and shutdown are unchanged.

  **Reconciliation** (unchanged in kind from r2; updated in detail):
  - one register row;
  - the allow-list entry;
  - `APPLICABLE` + `core/cmd/server/main.go` (a plain base-template file, so the disabled render has the same rendered base);
  - B-MCP-SOURCE pins `MAIN_GO_G003_SHA` and keeps every security assertion (§10, §17).
- **New tests:**
  - `context_test.go` and `propagation_test.go`: as in r2.
  - **`exit_test.go`:** as in r2, plus `ListenPort` for a `TCPAddr`, a `UnixAddr` (0) and nil (0).
  - **`errorlog_test.go`:** as in r2.

  `observability_test.go` and `health_test.go` are unchanged and must pass.
- **Real-binary checks** (B-CORE-ERROR-OUTPUT, §10):
  - **CE1:** configuration invalid.
  - **CE2:** listen failure. **Exactly one** output line, `core.exit`/`listen_failed`; **no** `core listening` line; no canary.
  - **CE3 (new):** a valid bind. Exactly one `core listening` line, with `port` and `app_env` keys only; then on TERM, a clean exit 0 with no further line; no canary.

### 7.2 Web (routes and BFF; no UI file changes)

- **`web/lib/request-context.ts` (new; `server-only`; no relative imports)** is as in r2: `deriveContext`, `outboundHeaders` and `withCompletionLog`. **r3:**
  - `error_type` comes from the closed `instanceof` category mapping (§5.1–§5.2 r3 additions), implemented as a fixed table of built-in constructors.
  - The thrown value's `name`, `message`, `cause`, stack and own properties are never read into the record.
  - The catch path still returns `new Response(null, { status: 500, headers })` with `Cache-Control: no-store` and `X-Request-Id`, and never rethrows (Q15).
  - Completion semantics are unchanged: one record, `outcome` `failed` or `cancelled`.
- **`core-client.ts` and both routes, `web/package.json` scripts:** as in r2.
- **`web/tests/unit/request-context.test.mjs`:** r2's coverage, plus:
  - an `Error` whose `name` is a **syntactically valid canary-bearing identifier** (for example `Cx<hex>`, matching `^[A-Za-z][A-Za-z0-9_.]{0,63}$`) → `Error`, with the canary absent;
  - a `TypeError` subclass with a custom `name` → `TypeError`;
  - a `DOMException` with an arbitrary name → `DOMException`;
  - an `AbortError` → `AbortError`;
  - the message, `cause`, own-property and non-`Error` canaries, as in r2.
- **`web/tests/fixtures/error-probe/{wrapped-route.ts,control-route.ts}`:** as in r2. **r3:** `wrapped` gains `mode=badname-valid` (a regex-valid canary-bearing `name`, giving WE3a) next to `badname` (an invalid name, giving WE3b).

### 7.3 Intelligence (`src/`, check stage only in the image)

- **`intelligence/src/observability/http.py` (new)** is `CompletionLogMiddleware`, as in r1.
- **`intelligence/src/observability/export.py` (customized).**
  - **`JsonFormatter`:** the §5.1 envelope, `service`, redaction (§5.5) and a 512-character value cap, as in r1. Plus the r2 event rules (§5.6):
    - intelligence loggers use the template and never `getMessage()`;
    - uvicorn records are mapped by template;
    - other libraries become `library.log`;
    - warnings become `python.warning`;
    - exceptions keep their type only.
  - **`configure_logging(level=INFO, secrets=())`** is idempotent, with the secrets updatable as in r1. **r2:**
    - it installs **one** JSON stream handler on the root logger, with `intelligence` at INFO and the root at WARNING;
    - it replaces the `uvicorn` logger's handlers with the same handler, keeping `propagate=False`;
    - it sets `uvicorn.access` to `disabled=True` with no handlers;
    - it calls `logging.captureWarnings(True)`.

    The test asserts there are no duplicate lines. `JsonFormatter()` keeps its no-argument constructor.
  - `ObservabilityHook` is unchanged, so content capture stays constructor-only and off by default.
- **`intelligence/src/main.py` (customized).** As in r1: `create_app` passes the two secrets to `configure_logging` and the middleware, and the health handlers return `JSONResponse(..., headers={"Cache-Control": "no-store"})`. **r2:**
  - `__getattr__` builds `_app` inside `try: … except Exception as exc:`.
  - On failure it calls `configure_logging()` if it has not yet run, writes one `intelligence.startup.failed` record (`error_type` only) and raises `SystemExit(3) from None`.
  - uvicorn's `Config.load_app` does not catch `SystemExit`, so the process exits 3 without a traceback. The real-image check verifies this (B-INTEL-ERROR-OUTPUT, IE1).
- **`intelligence/src/streaming/schemas.py` (customized):** §8.4, unchanged from r1.
- **New tests** (no new packages):
  - `test_health_context.py`, `test_logging_privacy.py`, `test_conformance.py` and `conformance_exchange.py`, as in r1. r2 extends the last two (§8.3).
  - **`tests/test_error_output.py` (new):**
    - the startup-failure path through `main.__getattr__` with an invalid environment that includes a canary-bearing variable *name*: `SystemExit(3)`, one record, no canary;
    - the uvicorn template mapping for each template in §5.6;
    - an unknown template becomes `uvicorn.other`;
    - a third-party logger warning becomes `library.log` with no message;
    - `warnings.warn(<canary>)` becomes `python.warning` with no canary;
    - exactly one handler, with no duplicate lines;
    - the `uvicorn.access` logger is disabled.
- **`infra/docker/intelligence.Dockerfile` (customized)** is as in r1: the `check` stage only copies `conformance/intelligence/`.

## 8. Contract conformance (R5)

### 8.1 Layout (fixtures only; `contracts/intelligence/*.schema.json` unchanged)

```text
contracts/conformance/README.md                         provenance, case IDs, how to run, stream-check rules
contracts/conformance/intelligence/expectations.json    independently declared outcome per case, schema digests, corpus rules
contracts/conformance/intelligence/cases/envelope/*.json
contracts/conformance/intelligence/cases/event/*.json
contracts/conformance/intelligence/cases/ndjson/*.ndjson
contracts/conformance/intelligence/schema-negative/*.json
```

- **Why this location.**
  - It is kept outside `contracts/intelligence/`, so it is not copied into the runtime image's `/srv/contracts/intelligence`.
  - It is absent from the intelligence-disabled render, which B-TOGGLE-OFF already requires.
  - It is fixture-only, consistent with "contract schemas' test/fixture/documentation areas" (Q11).
- **Expectation pinning.** `expectations.json` pins the two schema SHA-256 values from §2. Both validators refuse to run if the bytes they load differ.
- **Fixture data.** All data is synthetic, hand-authored for G003, and uses RFC-style UUIDs. Values never look like real credentials, so gitleaks' generic rules are not triggered; secret-scan is not loosened.

### 8.1a r2 addition: declared schema-negative entries

`expectations.json` gains a `schema_negative` array, declared independently of either validator's output. Each entry gives:
- `id`;
- `base`: `execution-envelope` or `stream-event`;
- `mutation`: one JSON-Patch-style `add` or `replace` at a JSON Pointer, or `truncate-bytes` with a byte length, or `absent`;
- `class`: the expected audit classification.

It also declares the two normal-mode digest controls, SD01 and SD02. They have no fixture file, because they are derived at test time inside a scratch directory (§8.2a).

### 8.2 Case catalogue (file name = lowercase ID plus a slug; the expected result is fixed here)

**Envelope, valid (14):**

| ID | Case |
|---|---|
| EV01 | Minimal: required fields only, empty `allowed_tools` and `context_refs`, one message |
| EV02 | Full: `group_id`, `business_entity_id`, `callback_grant_ref`, all three tool levels, context refs, system/user/assistant messages |
| EV03 | `request_id` of length 8 |
| EV04 | `request_id` of length 64 with `.`, `_` and `-` |
| EV05 | Offset `+02:00` with a 9-digit fraction |
| EV06 | `2028-02-29` (leap year) |
| EV07 | `23:59:60Z` (leap second) |
| EV08 | Uppercase UUIDs |
| EV09 | 256 tools and 256 context refs |
| EV10 | 256 messages |
| EV11 | Every string maximum: content 200000, `authorization_ref` 512, `callback_grant_ref` 512, tool name 128, ref 256, kind 64 |
| EV12 | `trace_id` all zeros. The existing wire behaviour is preserved: the envelope schema has no nonzero rule, and none is imported. |
| EV13 | Nil UUID (shape-valid) |
| EV14 | `capability: embedding` |

**Envelope, invalid (50):**

| ID | Case |
|---|---|
| EI01 | Truncated object |
| EI02 | Trailing comma |
| EI03 | `NaN` literal |
| EI04 | Top-level array |
| EI05–EI14 | Each of the 10 required fields missing (in schema `required` order) |
| EI15 | `allowed_tools` as an object |
| EI16 | `request_id` as a number |
| EI17 | Unknown field |
| EI18 | `api_key` |
| EI19 | `password` |
| EI20 | `access_token` |
| EI21 | `authorization` |
| EI22 | `capability: vision` |
| EI23 | Tool level `admin` |
| EI24 | Extra property on a tool |
| EI25 | Role `tool` |
| EI26 | Empty messages |
| EI27 | 257 messages |
| EI28 | 257 tools |
| EI29 | Content of 200001 characters |
| EI30 | `request_id` of 7 characters |
| EI31 | `request_id` of 65 characters |
| EI32 | `request_id` with a space |
| EI33 | **`request_id` with a trailing `\n`** (F6) |
| EI34 | Uppercase `trace_id` |
| EI35 | `trace_id` of 31 characters |
| EI36 | `organization_id` not a UUID |
| EI37 | `user_id` of 35 characters |
| EI38 | `group_id` containing `g` |
| EI39 | **`2027-02-30T00:00:00Z`** |
| EI40 | **Month 13** |
| EI41 | **Hour 24** |
| EI42 | **Minute 60** |
| EI43 | **Second 61** |
| EI44 | **`12:00:60Z`** |
| EI45 | **`2027-02-29`** |
| EI46 | No offset |
| EI47 | Space separator |
| EI48 | 10-digit fraction |
| EI49 | Lowercase `t`/`z` |
| EI50 | **Offset `+24:00`** |

The bold cases EI39–EI45 and EI50 satisfy the `pattern` and fail **only** on `format: date-time`. They prove the format assertion is effective in both validators.

**Event, valid (15):**

| ID | Case |
|---|---|
| XV01 | `text_delta` with empty text |
| XV02 | `tool_call` with nested arguments |
| XV03 | `tool_result`, `is_error: false`, object content |
| XV04 | `tool_result`, `is_error: true`, null content |
| XV05 | `tool_result`, string content |
| XV06 | `citation` without locator |
| XV07 | `citation`, locator of 256 characters |
| XV08 | `structured_output` without `schema_ref` |
| XV09 | `structured_output` with `schema_ref` |
| XV10 | `usage`, all values |
| XV11 | `usage` with null `input_tokens`, `output_tokens` and `estimated_cost` |
| XV12 | `usage` with `units: 0` and zero tokens |
| XV13 | `usage` with tokens `1.0` (an integer-valued number) |
| XV14 | `error`, `retryable: true`, message of 1024 characters |
| XV15 | `done` |

**Event, invalid (34):**

| ID | Case |
|---|---|
| XI01 | Malformed JSON |
| XI02 | Unknown type |
| XI03 | Missing type |
| XI04 | Extra field |
| XI05 | Numeric text |
| XI06 | `done` without `request_id` |
| XI07 | `request_id` with a space |
| XI08 | `request_id` with a trailing `\n` |
| XI09 | Bad `call_id` |
| XI10 | Array arguments |
| XI11 | `tool_result` without `is_error` |
| XI12 | Empty citation ref |
| XI13 | String data |
| XI14 | Negative input tokens |
| XI15 | Negative output tokens |
| XI16 | Negative latency |
| XI17 | Negative cost |
| XI18 | Negative units |
| XI19 | Units `1.5` |
| XI20 | `usage` without `trace_id` |
| XI21 | Unknown usage capability |
| XI22–XI25 | `usage` with `organization_id`, `user_id`, `group_id`, `business_entity_id` (one case each) |
| XI26–XI30 | `usage` with the content fields `prompt`, `completion`, `messages`, `content`, `input` (one case each) |
| XI31 | Uppercase error code |
| XI32 | Error message of 1025 characters |
| XI33 | Extra field on `done` |
| XI34 | Uppercase `configured_provider` |

**NDJSON stream files (7).** Rules:
- every record is LF-terminated;
- no blank line;
- at least one record;
- every line parses strictly and validates against `stream-event`.

| ID | Case | Expected |
|---|---|---|
| NV01 | `text_delta`, `usage`, `done` | valid |
| NV02 | `error`, `done` | valid |
| NI01 | Final line truncated, no LF | invalid |
| NI02 | Non-JSON line in the middle | invalid |
| NI03 | Schema-invalid event in the middle | invalid |
| NI04 | Blank line | invalid |
| NI05 | Empty file | invalid |

These rules are test-profile rules for checking producer output. They make no wire-contract change.

### 8.2a Schema-negative and digest controls (r2; supersedes r1's "Schema-negative (7)" block in Appendix R1)

**Two paths, never mixed (correction 4):**
- **Normal mode** (`corpus`, `ndjson`, `peer` and the pytest corpus test). The schema digests are checked **before anything else**. Any byte difference from the two pinned SHA-256 values fails with class `schema-digest-mismatch`, with zero cases evaluated. No option or argument disables this.
- **Audit mode** (`schema-audit`). It accepts only files under `contracts/conformance/intelligence/schema-negative/`, or the canonical schema paths for the positive control. It never evaluates cases and never runs the digest step. Its only output is a classification. It exits 0 only for `ok`.

**Audit stage order** (identical in both languages, implemented separately; the first failing stage sets the class):

| Stage | Check | Class on failure |
|---|---|---|
| 1 | The file exists and is a regular file | `schema-missing` |
| 2 | Strict JSON parse (RFC 8259; `NaN` and `Infinity` rejected); the top level is an object | `schema-malformed-json` |
| 3 | `$schema` is exactly `https://json-schema.org/draft/2020-12/schema` | `schema-wrong-draft` |
| 4 | Meta-validation against the embedded Draft 2020-12 meta-schema, with no network | `schema-meta-invalid` |
| 5 | Every keyword is in the audited profile (§8.3) | `schema-unsupported-keyword` |
| 6 | Every `format` is `date-time` or `uuid` | `schema-unsupported-format` |
| 7 | Every `$ref` starts with `#/$defs/` and resolves locally. Nothing is fetched: Go's loader fails every URL, Python's registry has no retriever, and containers run `--network none`. | `schema-remote-ref` |
| 8 | Compile with assertions on | `schema-compile-failed` (no declared case expects it) |

**Cases.** Each SN fixture is the canonical base schema with exactly one mutation, so that every earlier stage passes and only the intended stage can fail:

| ID | Base | Single declared mutation | Declared class (both languages) |
|---|---|---|---|
| SN01 | envelope | `absent`: no file | `schema-missing` |
| SN02 | envelope | `truncate-bytes` to the first 64 bytes of the canonical file | `schema-malformed-json` |
| SN03 | envelope | `replace /properties/authorization_ref/type` with `5` | `schema-meta-invalid` |
| SN04 | envelope | `add /dependentSchemas` with `{"group_id": {"required": ["business_entity_id"]}}` (meta-valid; outside the profile) | `schema-unsupported-keyword` |
| SN05 | envelope | `add /properties/authorization_ref/format` with `"ipv4"` | `schema-unsupported-format` |
| SN06 | stream-event | `replace /oneOf/0/$ref` with `"https://example.invalid/x.json"` | `schema-remote-ref` |
| SN07 | envelope | `replace /$schema` with `"http://json-schema.org/draft-07/schema#"` | `schema-wrong-draft` |
| SD01 | envelope | **normal mode:** the canonical bytes plus one trailing LF, written to a scratch schema directory. This file is still valid, meta-valid and inside the profile. | `schema-digest-mismatch`, 0 cases evaluated |
| SD02 | envelope | **normal mode:** SN03's bytes in place of the canonical file, in a scratch schema directory | `schema-digest-mismatch`, 0 cases evaluated. This proves that normal mode never reaches meta-validation with unpinned bytes. |

**Derivation check (`contractcheck sn-derive`).** For every declared SN entry, it applies the declared mutation to the parsed canonical base and requires structural equality with the parsed fixture. For SN02 it requires byte equality with the truncated prefix; for SN01, that the file is absent. A fixture with a second, undeclared difference therefore fails. This shows that each SN case isolates a single defect.

**Positive audit control.** Both canonical schemas classify as `ok` in both languages, so the audit path is not vacuously failing.

**The EI39 control** is retained, in normal mode with canonical schemas: a scratch copy of `expectations.json` with EI39 set to `valid` makes both validators exit nonzero, naming EI39.

That is 120 document and stream cases, 7 schema-negative cases, 2 digest controls and 2 positive audit controls. All eight event variants appear among the valid cases. **No production schema byte changes;** every mutation exists only as a committed SN fixture or in a scratch directory.

### 8.3 Validators (r2; independent; neither reads the other's output; r1 text in Appendix R1)

- **Supported keyword profile** (unchanged from r1). The audit is shared by specification but implemented separately in each language: `$schema`, `$id`, `$defs`, `$ref` (local `#/$defs/...` only), `title`, `description`, `type`, `properties`, `required`, `additionalProperties`, `items`, `minItems`, `maxItems`, `minLength`, `maxLength`, `pattern`, `enum`, `const`, `oneOf`, `minimum`, `format`. The only formats allowed are `date-time` and `uuid`.
- **Go (`contractcheck`, santhosh v6.0.3).** It uses `NewCompiler()` (Draft 2020-12) and `AssertFormat()`; a `UseLoader` that fails every URL (meta-schemas are embedded); `AddResource` under `file:///contracts/...` URLs; strict JSON decoding; and RE2 patterns. Subcommands:
  - `corpus`: normal mode. Digest first, then the audit stages 1–8, then the cases. Writes `results-go.json`; exits nonzero on any mismatch.
  - `ndjson`: normal mode.
  - `produce-envelopes`: the test-only Go producer, as in r1.
  - **`schema-audit --schema FILE --as envelope|event`** (new): audit mode only. Prints `{"file": …, "class": …}`.
  - **`schema-negative --schemas DIR --corpus DIR`** (new): runs, for each declared entry, SN01–SN07 through `schema-audit`, `sn-derive`, the positive audit control, and SD01/SD02 through `corpus` in a scratch directory. It compares every class with the declaration and writes `schema-negative-go.json`. A wrong class fails, even when the exit status is nonzero.
  - **`sn-derive`** (new): §8.2a.
- **Python.** The existing `streaming.schemas` validator (§8.4), in the `intelligence-check` image, with strict JSON parsing.
  - The digest pin and the audit stages are implemented in the **test tooling** (`tests/conformance_exchange.py` and helpers used by `tests/test_conformance.py`), not in the runtime module.
  - `tests/test_conformance.py`:
    - parametrized over every corpus case, as in r1;
    - **plus** SN01–SN07, asserting the declared class from `expectations.json`;
    - SD01/SD02 through `tmp_path` copies, asserting that `SchemaDigestMismatch` is raised before any case is evaluated;
    - the positive audit control on both canonical schemas.

    It is asserted not to skip.
  - `tests/conformance_exchange.py` (script) has the subcommands `corpus`, `peer` and `ndjson` (r1) and **`schema-negative`** (new), which writes `schema-negative-python.json`.
- **Shared result format** (unchanged from r1): `{"validator", "schema_sha256", "corpus_sha256", "results": [{"id", "valid", "class"}], "counts", "skipped": 0}`. The schema-negative results use `{"id", "class", "expected_class", "match"}`.

### 8.4 Python validator conformance corrections (`schemas.py`; schema bytes unchanged; Q3)

- **ECMA-262 `pattern` semantics.**
  - `validators.extend(Draft202012Validator, {"pattern": _ecma_pattern})`.
  - Each schema pattern is translated once: an unescaped `$` outside a character class becomes `\Z`.
  - Patterns containing a construct whose semantics differ between ECMA-262 and Python `re` are rejected when the validator is built: `\d`, `\w`, `\s`, `\b`, `\B`, `\p`, `(?`, backreferences and lookarounds. All current schema patterns use only literal classes, quantifiers, anchors, groups and alternation.
- **`date-time` checker.** A project-owned `FormatChecker` adds `date-time`: the RFC 3339 §5.6 grammar plus calendar validation (month length and leap years).
  - Hours 00–23, minutes 00–59, seconds 00–60, offset hours 00–23 and offset minutes 00–59.
  - Second 60 is allowed only when the UTC-normalized time is 23:59.
  - `T`/`Z` are accepted in either case; this is irrelevant in practice, because the schema pattern already requires uppercase.
  - `uuid` remains jsonschema's built-in checker.
- **Audit.** §8.3, plus an assertion that `date-time` and `uuid` are present in `checker.checkers`.
- **Effect at runtime.** Intelligence's `/internal/v1/execute` now rejects, with 422, two classes of envelope it previously accepted only through gaps:
  - an envelope whose pattern-bounded identifier ends in a newline;
  - an envelope whose `deadline` is calendar-invalid. These already received 422 through `main.py`'s parse fallback, with the identical response `{"path":"/deadline","keyword":"format"}`.
- **Leap-second note.** A valid leap-second deadline (`23:59:60Z`) remains schema-valid, but `main.py`'s `fromisoformat` rejects it, as it does today. That pre-existing runtime check is not changed and is noted in OPERATIONS.md.

### 8.5 Peer exchange (harness; scratch `$WORK/contracts-exchange`, mode 777, mounted into both containers)

1. `contractcheck produce-envelopes`, which writes `go-envelopes.ndjson`.
2. `conformance_exchange.py peer`. Python accepts and executes the Go envelopes, then emits `python-events.ndjson`.
3. `contractcheck ndjson --require-types text_delta,usage,error,done`, which checks `python-events.ndjson`.
4. The harness derives corrupt copies of both real files:
   - the last line truncated;
   - a garbage line inserted;
   - an invalid event or envelope appended;
   - a blank line.

   The peer checker must reject each one. A checker that ignores a bad line cannot pass this step.

## 9. Validation tools module (test tooling, not runtime; r2; r1 text in Appendix R1)

- **Module:** `infra/validation/tools/`, `module validationtools`, `go 1.27.1`, as in r1.
  - Direct requirement: `github.com/santhosh-tekuri/jsonschema/v6 v6.0.3`.
  - Indirect requirement: `golang.org/x/text`, at the version minimal version selection resolves.
  - Both are recorded in `go.sum`.
  - It is resolved once at S2, in a container, with a module cache under the execution's external scratch root (`$EXT/gomod`, not a Docker volume as in r1). That is the only network step (§15.6).
- **Commands:**
  - `cmd/contractcheck` (§8.3), with the new subcommands `schema-audit`, `schema-negative` and `sn-derive`.
  - `cmd/httpprobe`, as in r1: repeatable and duplicate headers, header and body values read from files, `--parallel`, `--abort-after`, `--timeout`, JSON output.
  - `cmd/logcheck`, as in r1: line classification, completion-schema validation, exactly one record per request and service, and chain checks. **r2:**
    - validates every record's `event` against the per-service catalogue in `internal/logrecord/events.json` (§5.1–§5.2 r3 additions, §5.6);
    - validates `error_type` against each event's enumerated categories;
    - adds a `startup` mode for the error-output checks (exit status and an expected event and category set);
    - adds a `scan` mode that reports canary hits by file and line number without printing the canary.

    Its unit-test negatives grow to include an uncatalogued event, a free-text `error_type`, and a JSON line whose `event` carries an exception message.
- **Image:** `infra/docker/validation-tools.Dockerfile`, with stages `build`, `check` and `runtime` (distroless static nonroot), as in r1. Images are tagged `g003-vtools-<run>`, labelled, and removed at cleanup.
- **Execution.** `--network none`, except `httpprobe`, which joins a service's network namespace. No production network is joined and no port is published.
- **Exclusions.** Nothing here is linked into core. B-MCP-SOURCE asserts that no `jsonschema` package appears in `go list -deps ./cmd/server`.

## 10. Harness changes (`infra/validation/`; every existing assertion preserved; r3; r2 table in Appendix R2, r1 table in Appendix R1)

| Check | Suite | Change |
|---|---|---|
| B-SHELL | static | Also shellchecks `infra/ci/validate.sh`, `infra/ci/lib.sh`, `infra/ci/validate-controls.sh`, `infra/ci/cleanup-owned.sh` and `infra/ci/runner-preflight.sh`. **r3:** a second pass `shellcheck -S info --include=SC2251` over `infra/ci/*.sh` fails on any `!`-negated statement, which `set -e` ignores (F27). |
| B-HADOLINT | static | Adds `validation-tools.Dockerfile` |
| B-DOCKERIGNORE | static | Adds `validation-tools.Dockerfile.dockerignore` (`.env` and `.env.*`) |
| B-DEPS | static | Adds the tools-module rule: exactly one direct requirement (`santhosh-tekuri/jsonschema/v6 v6.0.3`), and `go.sum` present. The existing rules are unchanged (goose and MCP pins, intelligence lockfile equal to the render, web C1–C12 policy). |
| **B-MCP-SOURCE (reconciled; Q16, §17)** | static | The **same security assertions** remain:<br>• the server binary links neither the MCP SDK, `core/internal/mcp` nor migration code;<br>• no binary depends on fixture or isolation packages;<br>• no transport handler or `WithIsolateDDL`;<br>• `core/internal/mcp` does not import `net/http`;<br>• the routes are exactly `GET /healthz` and `GET /readyz`.<br>**Changed:** the byte-equality test `sha(main.go) == manifest_sha(main.go)` is replaced by `sha(main.go) == MAIN_GO_G003_SHA`, **`987c933560ef77ea9a2daa4a3bec4c9747ea69d578f3c9c9d8e6179d013551ea`**: the reviewed three-hunk file of §7.1, pinned in `run.sh` at S6.<br>**Added:**<br>• `go list -f '{{join .Imports " "}}' ./cmd/server` equals the pinned rendered import set minus `fmt`;<br>• no `jsonschema` package appears in `go list -deps ./cmd/server`.<br>The rendered base stays bound by B-RENDER-PLAN (register row) and B-TOGGLE-OFF-ADDITIONS. The reason message changes from "server main.go unchanged" to "server main.go equals its reviewed G003 digest". |
| B-WEB-BOUNDARY | static | The r1 additions (`request-context.ts` imports `server-only` and is imported only by `app/api/**/route.ts` and `lib/core-client.ts`; the core fetch passes only `outboundHeaders(ctx)`), plus: the `tests/fixtures/error-probe/*` files exist only under `web/tests/` and are never imported by `app/`, `components/` or `lib/`. The routes remain exactly `/`, `/api/health` and `/api/ready`. |
| **B-CI-CONFIG** (new) | static | §11.1 assertions on `.gitlab-ci.yml` (through yq), plus the README `platform:ci` block |
| **B-CI-CONTROLS** (new) | static | `infra/ci/validate-controls.sh` groups A–E (§11.3). It includes Docker-real cleanup controls using `busybox:1.37.0` with control labels, and the nested real `contracts` run of V4. |
| B-WRAPPER-CONTROLS, B-DEPS-CONTROLS | static | Unchanged (W1–W9; C1–C12, C8b, C8c) |
| B-RENDER-PLAN, B-RENDER-REPEAT | provenance | The register count changes from 8 to **19** (3 G001 + 5 G002 + 11 G003, §14.2). Exactly 19 conflicts and proposals. |
| **contracts suite** (new; root only; no stack) | contracts | • **B-CONTRACT-TOOLS:** the `validation-tools` check stage passes; the intelligence-check image builds; the runtime intelligence image has no `/srv/conformance`.<br>• **B-CONTRACT-GO** and **B-CONTRACT-PY:** in normal mode, each validator matches every expectation, the digests match, the audit passes and nothing is skipped.<br>• **B-CONTRACT-AGREE:** as in r1.<br>• **B-CONTRACT-PEER:** §8.5.<br>• **B-CONTRACT-NEGATIVE (r2):** in both languages, with `--network none`:<br>&nbsp;&nbsp;– SN01–SN07 in audit mode, each with its declared class;<br>&nbsp;&nbsp;– the positive audit control (`ok`) for both canonical schemas;<br>&nbsp;&nbsp;– `sn-derive` passes;<br>&nbsp;&nbsp;– SD01 and SD02 in normal mode, each `schema-digest-mismatch` with 0 cases evaluated;<br>&nbsp;&nbsp;– EI39 flipped in a scratch copy: both validators exit nonzero, naming EI39.<br>A wrong class is a FAIL even when the exit status is nonzero. |
| **B-HEALTH-ROOT** (new) | root | Host probes on the actual root (ports 3100, 8180, 8100): six routes with status, exact body, no-store and a valid `X-Request-Id` |
| **B-CORE-ERROR-OUTPUT** (new) | root | Runs, like the other two error-output checks, while the root images exist, after B-CHECKS-ROOT and before the `platform:reset` block. It uses the root project's built core runtime image, run with `docker run --rm --network none --label g001.run=$RUN_ID` and environment files under `$WORK` (mode 600), with per-run canaries:<br>• **CE1:** required variables missing; canary values in `DB_PASSWORD` and `REDIS_PASSWORD`. Expect exit 1, exactly one JSON line `core.exit`/`config_invalid` whose `config_keys` are in the allow-list, and no canary.<br>• **CE2:** otherwise valid configuration with `CORE_LISTEN_ADDR=bad-<canary>.invalid:8180`. Expect exit 1 and **exactly one** output line, `core.exit`/`listen_failed`: no `core listening` record (bind-then-log, §7.1) and no canary. Today's code would emit both a `core listening` record carrying the raw address (F28) and an `Fprintln` line (P-GO).<br>• **CE3 (r3):** valid configuration with `CORE_LISTEN_ADDR=:8180` and a canary in `PRODUCT_NAME`, which is never logged. Expect exactly one `core listening` record with keys exactly `time level service event port app_env`, `port` = 8180 and `app_env` = `test`. After `docker stop` (TERM), expect exit 0 and no further line. No canary.<br>All output is classified by `logcheck startup`. Every line must be a catalogued event (§5.6); none is exempted. |
| **B-WEB-ERROR-OUTPUT** (new) | root | A disposable `copy_repo`, with the two fixture routes copied into `app/api/__probe/{wrapped,control}/route.ts`. It is built with `web.Dockerfile --target runtime` under a run-labelled tag, and run with `--network none`, `PROBE_CANARY=<canary>`, harmless identity values, and `CORE_INTERNAL_URL=http://127.0.0.1:9`. `httpprobe` joins its namespace:<br>• **WE0:** only the allow-listed startup lines (§5.4) and JSON appear;<br>• **WE1:** `mode=throw` → 500, empty body, no-store, `X-Request-Id`, one `failed`/`Error` record;<br>• **WE2:** `nonerror` → `NonError`;<br>• **WE3a (r3):** `badname-valid`, a **regex-valid** canary-bearing `name` (for example `Cx<hex>`) → `error_type` `Error` and no canary anywhere;<br>• **WE3b:** `badname`, an invalid name → `Error`;<br>• **WE4 (positive control):** `control` → 500, and the canary **is** present in Next's raw output. Otherwise the check FAILs as vacuous.<br>• **WE5:** a malformed percent-encoded query carrying a canary on `wrapped?mode=ok` → no non-JSON line beyond WE0 and no canary;<br>• **WE6:** a client abort during `mode=slow` → one `cancelled` record and no framework line.<br>The canary is absent from everything except WE4's control output. The control output is deleted after the assertion and is never retained. |
| **B-INTEL-ERROR-OUTPUT** (new) | root | The root project's built intelligence runtime image, `--network none`:<br>• **IE1:** `APP_ENV` missing, plus a variable named `AI_MODEL_CANARY<hex>=openai:x`, whose *name* reaches the `ConfigError` message. Expect exit 3, every line JSON, exactly one `intelligence.startup.failed`/`ConfigError`, no canary, and no traceback.<br>• **IE2:** valid test configuration. Start, wait for `uvicorn.startup_complete`, then stop: every line is a catalogued JSON event, `uvicorn.*` templates are mapped, and there is no text line. |
| B-CHECKS-ROOT | root | Unchanged command; it now also runs the new unit, conformance and error-output tests |
| **B-OBS-CONTEXT, B-OBS-CHAIN, B-OBS-TIMEOUTS, B-OBS-LOGS, B-OBS-PRIVACY** (new) | variants (`variants_ops`, before B-OUTAGE) | §16 AC1–AC4 on the enabled disposable copy, as in r1. B-OBS-LOGS uses the r2 event catalogue. |
| **B-OBS-OFF** (new) | variants (`variants_off`, before B-BROWSER-OFF) | As in r1 |
| B-TOGGLE-OFF-ADDITIONS | variants | `APPLICABLE` adds the r1 six (`core/internal/observability`, `core/internal/health`, `web/app/api`, `web/lib/core-client.ts`, `web/lib/request-context.ts`, `docs/foundation/OPERATIONS.md`) **and `core/cmd/server/main.go`**: 28 → **35** |
| B-SECRET-CANARY | final | Adds the G003 canaries (§16 AC4) to the scan list. With `--secrets-file`, the list lives at that external path, which is outside `$WORK` and so needs no exclusion. Canary request specifications live in `$WORK/obs-inputs/` (excluded like `secrets*`) and are deleted after use. |
| ~~B-ARTIFACT-SCAN~~ (r1 proposal, **withdrawn**) | — | Artifact scanning has a single implementation: `validate.sh`'s raw and staged scans (§11.2 F3 and F7), where retention and upload happen. This resolves the r1 raw-versus-staged mismatch. Generated secrets and canaries inside the harness's own lifetime remain covered by B-SECRET-CANARY. |
| **B-REQUIRED-CHECKS** (new) | final | Every ID in `infra/validation/required-checks.tsv` for the executed suites is present exactly once and PASS. Missing, duplicate, FAIL or UNMET → FAIL. |
| `all` | — | Becomes `static provenance contracts root migrations variants`. Required IDs: 50 existing + **19** new = **69** (B-ARTIFACT-SCAN withdrawn; B-CORE-, B-WEB- and B-INTEL-ERROR-OUTPUT added). |
| **run.sh options and hooks (r3)** | — | Defaults keep today's behaviour; `run-approved.sh` is unchanged.<br>• `--owned-registry FILE` appends registry rows (§11.4): `work_root`, `label g001.run=<RUN_ID>`, `project <name>` (from `register_project`), `rootenv <sha256>`, **`image <tag> <id>` immediately after every image the run builds** (`docker build -t`, Compose builds), and **capture rows** from `owned_capture DEST` (§11.4).<br>• `owned_capture` is called after every start or `up` of a registered project, and first in the cleanup trap. It writes `container <id>`, `network <id>`, `volume <name>` and `image <tag> <id>` rows for that project's owned containers and what they are attached to, mount and run.<br>• `--secrets-file FILE`: as in r2.<br>• **One-line `tee` change (F26):** `exec > >(trap '' TERM INT HUP; exec tee -a "$HARNESS_LOG") 2>&1`. The harness's own EXIT-trap cleanup can then still log and run under a group TERM; KILL still ends `tee`. Normal runs are unaffected. |
| report copies | — | The end-of-run `--report-dir` copy list adds `obs-*.tsv`, `obs-timings.tsv`, `contracts-results-{go,python}.json`, `contracts-agree.tsv`, `schema-negative-{go,python}.json`, `error-output-{core,web,intelligence}.tsv` (statuses and categories only, never captured output) and `ci-controls.log`. These are the names `validate.sh` F4 and F5 expect. |
| cleanup | — | Also removes the run's `g003-*` images, by registered tag and ID. The inventory diff is unchanged. The accepted `docker compose -p … down` backstop stays, and CI cleanup does not rely on it (§11.4). Under signals the harness cleanup is best-effort (F26); `validate.sh` F2 is the guaranteed path. |

**Unchanged:** B-SCOPE, B-PROTECTED, B-SKILL-UNCHANGED, B-LITERALS (which covers the new files: no identity literal, GitLab path or runner tag in `infra/` or `docs/foundation/`), B-NAMING, the migration suite, B-PORTS-*, B-BOUNDARIES-* and their controls, B-PERSIST, B-RENAME, B-BROWSER-RENAME, B-DBRENAME, B-SOURCE-HASH, B-DOCKER-ONLY, B-WEB-DESIGN, both browser matrices (95 tests each), `run-approved.sh` and its controls.

## 11. GitLab CI (R6/R7; r2; r1 §11 in Appendix R1)

### 11.1 `.gitlab-ci.yml` (exact intended content; `<RUNNER_TAG>` comes from the owner's Q1 facts and is written at S6)

```yaml
# Baseline CI (docs/foundation/OPERATIONS.md). One gating job runs the repository's containerized
# validation entry point. There is no publish, release or deploy job, and no CI/CD secret is used.
workflow:
  rules:
    - if: $CI_PIPELINE_SOURCE == "merge_request_event"
    - if: $CI_PIPELINE_SOURCE == "push" && $CI_COMMIT_BRANCH == "staging"
    - if: $CI_PIPELINE_SOURCE == "push" && $CI_COMMIT_BRANCH == $CI_DEFAULT_BRANCH
    - when: never

stages: [validate]

validate:
  stage: validate
  tags: [<RUNNER_TAG>]
  timeout: 3h
  resource_group: platform-validation-host
  interruptible: false
  variables:
    GIT_STRATEGY: clone
    GIT_CLEAN_FLAGS: -ffdx
    RUNNER_SCRIPT_TIMEOUT: 2h45m
    RUNNER_AFTER_SCRIPT_TIMEOUT: 10m
  script:
    - bash infra/ci/validate.sh --publish-dir ci-evidence
  after_script:
    - bash infra/ci/cleanup-owned.sh --after-script
  artifacts:
    when: always
    expire_in: 90 days
    access: developer
    paths: [ci-evidence/]
```

**r2 changes from r1:**
- the two `RUNNER_*_TIMEOUT` variables;
- the `--after-script` mode of `cleanup-owned.sh`.

Why (F14):
- Their total, 2h55m, fits within `timeout: 3h`, so that on runners that support these variables, `after_script` still runs if the script times out.
- The internal deadline in `validate.sh` (§11.2) is the primary mechanism and does not depend on the runner version (F-R4).
- `after_script` is a non-gating backstop: it cannot fail a successful job, and its output is informational.
- `artifacts:when: always` does not upload on job timeout. The internal deadline therefore ends the harness early enough for the finalizer to scan and publish within the job.

**B-CI-CONFIG asserts:**
- **Structure.** The top-level keys are exactly `workflow`, `stages` and `validate`. The four workflow rules are exact. There is exactly one job.
- **No bypass.** No `allow_failure`, `when: manual` or `when: delayed`, no job-level `rules` or `only`/`except`, no `rules:changes`, `needs`, `retry`, `parallel`, `trigger`, `environment`, `release`, `pages`, `image` or `services`.
- **Script.** It is exactly the single line above, and equal to the README `platform:ci` command.
- **`after_script`.** It is exactly `bash infra/ci/cleanup-owned.sh --after-script`.
- **Artifacts.** Exactly as shown: `when: always`, one directory, `access: developer`, `expire_in: 90 days`, no `untracked`.
- **Variables.** Exactly the four shown. No secret-like names. `RUNNER_SCRIPT_TIMEOUT` plus `RUNNER_AFTER_SCRIPT_TIMEOUT` is at most `timeout`, and `timeout` is at most 3h.
- **Tags.** Exactly one tag matching `^[A-Za-z0-9._-]{1,64}$`. Its value is the owner-supplied tag; Codex checks it against Q1, and hosted evidence records `CI_RUNNER_TAGS`. It is not repeated in `docs/` or `infra/`, so a tag containing an identity literal cannot trip B-LITERALS.

**Pipeline behaviour** (as in r1, with r2 attribution in §11.2):
- Merge-request pipelines test the source content (`detached`), or the merged result when the project has merged results enabled (Premium/Ultimate, F13; owner fact F-R10).
- Pushes to `staging` and to the default branch (`main`) also run.
- Feature-branch pushes, tags, schedules, web, API and trigger pipelines run nothing. If such a pipeline ever reached the job, attribution would reject it with 16.
- A push to `staging` while an MR from `staging` is open yields two gating pipelines for the same commit. This is accepted.
- `resource_group` serializes this project's validation jobs on the shared host ports.

### 11.2 `infra/ci/validate.sh` and `infra/ci/lib.sh` (aggregate entry point; Bash ≥ 4.4; r3; r2 version in Appendix R2)

```text
validate.sh [--evidence-dir DIR] [--publish-dir ci-evidence]
```

#### 11.2.1 Execution model (corrections 1 and 2)

**Parent-held state.** `validate.sh` keeps all cross-step state in **parent-shell variables**: `EV`, `REG`, `LOC_OK`, `HEAD_SHA`, `TREE_SHA`, `STATUS_DIGEST`, `BUDGET_S`, `TESTED_CONTENT`, `HPID`, `WPID`, `DEADLINE`, `ST[...]`, `MAIN_DONE`. Only two kinds of code may set them.

1. **Parent steps,** which run directly in the parent shell and use explicit checks (`if …; then …; fi` and `|| return N`; never `!`-negated statements, F27):
   - L0 location;
   - the state import;
   - the harness launch, wait and supervision;
   - signal handling;
   - candidate selection;
   - aggregation.
2. **Isolated phases,** run as `set +e; ( set -e; FUNC ); rc=$?; set -e`. They report **only** a status in `ST[NAME]`, because subshell assignments never reach the parent (F24, P-STATE).
   - A phase that computes values writes them to a state file, `$EV/state/<phase>.env`.
   - The parent imports that file with **`state_import FILE KEY=REGEX…`**. The importer accepts only `^[A-Z][A-Z0-9_]*=…$` lines; every declared key exactly once; no unknown, duplicate or missing key; each value matching its regex; a regular, non-symlink file. It assigns with `printf -v`, never `source` or `eval`.
   - Any violation fails the import with 16.
   - The P1 keys are `HEAD_SHA` (40 hex), `TREE_SHA` (40 hex), `STATUS_DIGEST` (64 hex), `BUDGET_S` (`^[1-9][0-9]{0,5}$`) and `TESTED_CONTENT` (enum).
   - P-IMPORT: 8 of 8 unit cases, including an injection attempt with no side effect.

**Harness supervision (parent; F25, F26).** GNU `timeout` is not used.

- **Launch.**
  ```text
  setsid bash infra/validation/run.sh … >"$EV/raw/harness-console.log" 2>&1 &   HPID=$!
  ```
  `HPID` is both the PID and the process-group ID (P-SETSID).
  - The parent then waits for the group to exist (`group_ready`: up to 5 s polling `kill -0 -- -$HPID`). The model's F15 first run showed that signalling immediately after launch can miss the group.
  - It records `HPID` in `$EV/state/harness.pgid`, for diagnosis only.
- **Deadline.**
  ```text
  setsid bash -c 'sleep "$1"; kill -USR1 "$2"' watchdog "$BUDGET_S" "$$" &   WPID=$!
  ```
  `group_ready` follows. `trap 'DEADLINE=1' USR1` is set before the launch. The parent `wait`s on `HPID`; USR1 interrupts the wait.
- **Stop sequence** (`stop_harness_group`; used for the deadline, for signals, and for any harness still running at finalization):
  1. `group_ready`, or else TERM the PID itself;
  2. `kill -TERM -- -$HPID` to the **whole group** (foreground Docker CLIs, the harness Bash, its `tee` — which ignores TERM after the r3 harness change — and any backgrounded children);
  3. poll `kill -0 -- -$HPID` for a grace period of **150 s** (the harness's own cleanup, for example `compose down`, runs inside it);
  4. then `kill -KILL -- -$HPID`;
  5. reap with `wait`;
  6. record whether the group is gone.
- **Recorded harness status:**
  - **124** at the deadline;
  - **128 + n** on signal n;
  - otherwise `run.sh`'s own status.
  - If processes remain in the group after a normal exit (stragglers), they are stopped and an INFO `stragglers=1` is recorded.
- **The watchdog** is stopped (`kill -TERM -- -$WPID`, then reaped) as soon as the harness wait returns, and at finalization.

**Signals.**
- `trap 'on_signal TERM 15' TERM`, and likewise for INT (2) and HUP (1).
- `on_signal` stops the watchdog and, if a harness is running, the harness group; records `ST[harness]=128+n`; then `exit 128+n`, which runs the EXIT trap.
- `finalize` first sets `trap '' TERM INT HUP USR1`, so a second signal cannot cut finalization short. A GitLab KILL still can; `after_script` is the backstop for that case.

**Internal deadline value.** `BUDGET_S = min(CI_JOB_TIMEOUT, RUNNER_SCRIPT_TIMEOUT) − 1200 s`, written by P1 as a state value. Locally, with neither variable set, it is 14 400 s. P1 fails with 16 if it is below 3600 s.

**Run start.** `trap finalize EXIT` and the signal traps are installed immediately after argument parsing, before L0.

**Evidence (P-MODEL).** A model with these mechanics passed 24 of 24 fault cases with real signals (§15.5a). These include:
- an external SIGTERM to the supervisor while a stub harness with grandchildren ran: exit 143; group gone; the harness's own cleanup ran; postcheck and cleanup ran;
- a TERM-ignoring harness: KILL after the grace; group gone; exit 143;
- the watchdog deadline: exit 124; group gone; harness cleanup ran;
- an unexpected parent failure while the harness ran: finalizer stopped the group; postcheck and cleanup ran;
- stragglers;
- three pre-location failures.

#### 11.2.2 Main body (in order)

| Step | Where | Work | Status on failure |
|---|---|---|---|
| L0 `location` | parent | **Containment before creation:**<br>• `REPO_R = realpath -e REPO`.<br>• With `--evidence-dir`: the target is `realpath -m DIR`. Reject it if it is `REPO_R` or below it (traversal and external or dangling symlinks resolve first), or if it already exists. Create it with `mkdir -m 700`.<br>• Without `--evidence-dir`: the **base** is `realpath -m "${TMPDIR:-/tmp}"`. Reject it if it lies inside the repository or is not a directory, **before** `mktemp -d "$base/g003-validate-XXXXXX"`.<br>• Re-resolve the created path with `realpath -e` and re-check containment.<br>• **Registry:** in CI, `REG=$(dirname "$CI_PROJECT_DIR")/.platform-validation-owned-$CI_JOB_ID`, resolved with `realpath -m`, checked to lie outside the repository, required not to exist, and created with `mkdir -m 700`. Locally, `REG=$EV/owned`.<br>• Create `raw/`, `state/` and `final/`.<br>• `--publish-dir` must be exactly `ci-evidence`.<br>If any check fails: `LOC_OK=0`, then `exit 16`, which leads to the **pre-location path** (§11.2.3). | 16 |
| P1 `preflight` | isolated + import | • `runner-preflight.sh` (§11.5) writes `raw/runner-preflight.tsv`; any FAIL row → 16;<br>• `docker pull` of the pinned gitleaks image (setup);<br>• the start Docker inventory with IDs (`$REG/start-inventory.ids`);<br>• attribution (§11.2.4, unchanged) → `raw/attribution.json`;<br>• precheck: `status --ignored` must be empty;<br>• write `state/p1.env`.<br>The parent then runs `state_import` on `state/p1.env`. | 16 (phase or import) |
| P2 `harness` | parent | Only if L0, P1 and the import passed. Supervised launch (§11.2.1):<br>`setsid bash infra/validation/run.sh --base HEAD --allow-file infra/ci/unchanged.allow --owned-registry "$REG/owned.tsv" --secrets-file "$REG/secrets" --report-dir "$EV/raw/report" all >"$EV/raw/harness-console.log" 2>&1 &` | the harness status (as recorded above) |

`MAIN_DONE=1` is set only after the main body's last step. An unexpected failure of a parent step (errexit) leaves it 0, and leads to `finalize`.

#### 11.2.3 Finalizer (always, once; parent orchestration with isolated bodies; order fixed)

**Pre-location path** (`LOC_OK=0`). No Docker operation and no evidence write has happened yet, because L0 is the first action.
- Only **F1** runs: the repository status check, read-only, with output to the job log only.
- F2 is recorded as `n/a (no registry)`; F3–F9 do not run.
- A directory that L0 itself created, and that it has proven to lie outside the repository, is removed.
- **Published: none.** Exit 16.
- Model cases F20–F22 (TMPDIR inside the repository, `--evidence-dir` inside it, and a symlink into it): exit 16, nothing created inside the repository (whole-tree inventory unchanged), postcheck ran.

**Normal path** (`LOC_OK=1`). A harness still running is stopped first.

| Step | Work | Status on failure |
|---|---|---|
| F1 `postcheck` | The repository status must be empty, and `HEAD`, the tree and the status digest must equal the imported P1 values when present | 13 |
| F2 `cleanup` | `cleanup-owned.sh --registry "$REG" --mode finalize --out "$EV/raw/cleanup-inventory.tsv"` (§11.4: positive ownership; unproven resources kept and reported) | 18 |
| F3 `raw_scan` | Over **all** of `$EV/raw`: literal scan against `$REG/secrets` (a missing or empty file when the harness ran is a scanner failure), gitleaks with `--exit-code 23`, and forbidden names. Any unreadable file is a scanner failure (model F11b). | 15 finding; 17 scanner failure |
| F4 `complete` | Only if P2 ran. Strict parse as in r2. **Also** writes `$EV/final/status-columns.tsv`, containing only lines matching `^[A-Z][A-Z0-9-]*\t(PASS\|FAIL\|UNMET\|INFO)$`. | 14 |
| F5 `stage` | Only if F3 returned 0. **All or nothing:** copy the allow-list into `$EV/staged.tmp`. On any copy error, delete `staged.tmp` entirely and return 19. Only when every copy succeeds, rename it to `$EV/staged`.<br>The allow-list is r2's list **without `harness-console.log`**, which is now genuinely excluded. `deps-codes.txt` is also excluded. | 19 |
| — candidate (parent) | **full** if F3 = 0 and F5 = 0;<br>**minimal** if F3 = 15 (marker `WITHHELD`) or if F5 failed (marker `STAGING-FAILED`);<br>**none** if F3 = 17.<br>After a preflight failure the raw set holds only preflight and attribution evidence. If that set is clean it is staged and published as **full**, as r2's group B expected (model F03). | — |
| F6 `summary` | Builds the candidate in a fresh `$EV/candidate`: full = staged files plus the summary; minimal = the marker plus the summary.<br>`summary.json` and `phase-status.tsv` contain only validated fields (enums, integers, hex SHAs, UTC times, allow-listed CI metadata; a malformed value fails the step). They are labelled `"status_kind": "pre-publication"`, record L0–F5, and list `pending: summary, candidate_scan, publish, report, scrub`.<br>On failure the candidate is deleted, and the outcome becomes **none**. | 19 |
| F7 `candidate_scan` | Scanners (a)–(c) over exactly `$EV/candidate` **and** `$EV/final/status-columns.tsv`.<br>• A finding in a **full** candidate demotes it to minimal: F6 is rebuilt as `WITHHELD`, then rescanned (`rescan`). If the rescan is not clean, the outcome becomes **none**.<br>• A finding in a minimal candidate, or any scanner failure, gives **none**. | 15 or 17 |
| F8 `publish` | CI only, and only if a candidate survived F7: copy to `$CI_PROJECT_DIR/ci-evidence.tmp`, then `mv` it to `ci-evidence`. On failure, remove `.tmp`; nothing is published. | 19 |
| F9 `report` | The job log gets `phase=… status=…` lines. Then, **only if F3 and F7 were clean**, the `ID STATUS` lines from the validated and scanned `$EV/final/status-columns.tsv`, which F10 has not yet removed. | 19 |
| F10 `scrub` | Delete `$REG/secrets`, `raw/`, `staged*`, `candidate/`, `final/` and `state/`. In CI or default-location mode, also delete `$EV`. Delete `$REG` only if F2 returned 0; otherwise keep it for `after_script` and report that. | 19 |
| F11 `exit` | Aggregate (§11.2.5) and print the **authoritative final line**: `validate: final status N; published=<full\|minimal\|none>; phases: name=status …`. Then exit N. | — |

**Publication outcomes** (exactly one per run; recorded in the final line):

| Outcome | When | Contents |
|---|---|---|
| **full** | F3 = 0, F5 = 0, F6 = 0, F7 = 0, F8 = 0 | the scanned staged allow-list, `summary.json` and `phase-status.tsv` |
| **minimal** | the candidate is minimal (or was demoted), F6 = 0, the minimal set's scan = 0, F8 = 0 | exactly the marker, `summary.json` and `phase-status.tsv` |
| **none** | any scanner failure (17), any F6 failure, any F8 failure, or the pre-location path | nothing: no `ci-evidence/` and no `.tmp` left behind |

There is **no partial set** (staging is all-or-nothing) and **no unscanned fallback** (every published byte passed F7).

#### 11.2.4 Checkout attribution (correction 2; documented variables only, F13)

**Unchanged from r2 (accepted by Codex).** The rules table, raw parent check, recorded metadata and the P-ATTR prototype (21 of 21 as expected) are in r2 §11.2.4, preserved in Appendix R2. It runs in P1, before the harness.

#### 11.2.5 Exit status (aggregation; structural no-masking invariant)

| Code | Meaning |
|---|---|
| 0 | every recorded step returned 0 (skipped steps are `skip` or `n/a`, and only as a consequence of another recorded nonzero status) **and** `MAIN_DONE=1` |
| harness status | P2 ran and the recorded status is nonzero (1 or 2 from `run.sh`; 124 at the deadline; 143, 130 or 129 on a signal) |
| 13 / 14 / 15 / 16 / 17 / 18 | postcheck drift / evidence incomplete or malformed / unsafe evidence / location, preflight or state import / scanner failure / cleanup incomplete, refused or unproven |
| 19 | staging, summary, publication, report or scrub failure; an unexpected exit outside a step (`MAIN_DONE=0`); or **any recorded nonzero status not otherwise selected** |

**Precedence** chooses which code is reported:
1. 16 for L0, P1 or the import;
2. otherwise the nonzero harness status;
3. otherwise 17, then 15, 14, 13, 18.

**Structural invariant:** after precedence, if any recorded status anywhere in the **single phase list** is nonzero and the selected code is still 0, the exit is **19**. The aggregator iterates over the same `PHASES` array the finalizer records into, so a newly added step cannot be omitted. This fixes the r2 model defect (`stage=19` → exit 0): model case F11 now gives `stage=19` → **exit 19**, published=minimal, nothing partial.

**Authoritative status.** The artifact's `summary.json` and `phase-status.tsv` are **pre-publication metadata**. They record L0–F5 and list the later steps as pending. The **authoritative** result is the job's exit status (the GitLab job and pipeline status) together with the final job-log line from F11, which records every step, including F7–F10. The Result and Codex use those for H1 and H2.

### 11.3 CI controls (`infra/ci/validate-controls.sh`; scratch trees outside the checkout; no switch in `validate.sh`; r3; r2 version in Appendix R2)

**How the controls work.** As in r2:
- disposable `git init` trees containing copies of `validate.sh`, `lib.sh`, `cleanup-owned.sh`, `required-checks.tsv` and `unchanged.allow`;
- a **stub** `infra/validation/run.sh`, shaped like the real harness (`set -euo pipefail`, an EXIT-trap cleanup, and stdout and stderr through a TERM-ignoring process-substitution `tee`);
- a **stub** `runner-preflight.sh`;
- faults injected only through the stubs, a `docker` shim, file modes, GitLab-defined variables, and **signals sent from outside**.

`validate.sh` has no fault switch. Each control asserts:
- the exit status and the final line;
- the stub invocation count;
- the publication outcome and its exact contents;
- canary absence;
- that F1 and F2 ran;
- for supervision controls, that every PID recorded by the stub is gone and that `kill -0 -- -PGID` fails.

The r3 planning model (§15.5a, P-MODEL) runs the same matrix against a stand-in. The model cases are named in brackets.

**Group A — gate semantics** (r1 and r2, retained):

| Control | Scenario | Expected |
|---|---|---|
| V1 | All valid, complete PASS [F01] | 0; **full** |
| V2 **(a)** | One FAIL row; harness exit 1 [F02] | 1; full (scanned reports) |
| V3 (a) | Harness 0, but a FAIL row | 14 |
| V4 **(b)** | Real `run.sh … contracts` on a scratch copy with EI39 flipped; its `results.tsv` replayed through V2's path | 1 |
| V7 **(d)** | Harness 1, clean tree | 1; postcheck passed |
| V8 / V9 **(e)** | Harness 0; a tracked file modified, or an untracked file created [F09] | 13 |
| V10 / V11 **(f)** | A required ID missing; an UNMET row; browser JSON with `skipped: 1` | 14 |
| V12 | Harness 1 **and** drift | 1, with the drift reported |

**Group B — attribution:** unchanged from r2 (P1–P5, P3F, N1–N16; accepted).

**Group C — evidence scanning** (correction 2):

| Control | Scenario | Expected |
|---|---|---|
| V5a | Generated canary in a **retained** raw report [F06] | 15; **minimal** (`WITHHELD`); canary absent from the published set |
| V5b-planted | The canary only in `harness-console.log`, which is **excluded** from staging in r3 [F07] | 15; **minimal**; canary absent from the published set |
| V5b-clean | No canary; `harness-console.log` present in the raw set [F07b] | 0; **full**; `harness-console.log` **absent** from the published set. Proves the exclusion itself. |
| V5c | A gitleaks-detectable token generated at run time, not in the secrets file [F08] | 15; minimal |
| V6a | Scanner failure: the gitleaks shim exits 1 [F05] | 17; **none** |
| V6b | The harness ran, but the secrets file is missing | 17; none |
| V6c | Positive scanner control: clean raw and candidate sets (identical to V1) [F01] | 0; full |
| V6d | An unreadable raw file, so the literal scan returns 2 [F11b] | 17; none |

**Group D — finalization, supervision and location** (corrections 1 and 2):

| Control | Scenario | Expected |
|---|---|---|
| V14 / V15 | Malformed `results.tsv` [F04]; malformed browser JSON | 14; full; F1 and F2 ran |
| **V16** | **Staging after a partial copy:** the first allow-listed files copy, then a later one fails (it is a directory) [F11] | **19**; **minimal** (`STAGING-FAILED`); none of the partially copied files is published |
| **V16b** | Summary failure: a malformed allow-listed CI value [F12] | 19; **none** |
| V17 | Publication failure: the publish parent is not writable [F13] | 19; none; no `ci-evidence/` or `.tmp` left behind |
| **V17b** | Scrub failure: an undeletable raw subdirectory [F14] | **19**; full already published; the final line reports `scrub=19` |
| V18 | **External SIGTERM to `validate.sh`** while the stub harness and its grandchildren run [F16] | 143; the harness group is gone (every recorded PID is dead, `kill -0 -- -PGID` fails); the stub's cleanup trap ran; F1 and F2 ran; full (scanned) |
| V18a | A TERM-ignoring stub harness and SIGTERM to `validate.sh` [F17] | 143; KILL after the grace; group gone; F1 and F2 ran |
| V18b | Harness 1 and a cleanup failure | 1; `cleanup=18` reported |
| V18c | The stub preflight reports a FAIL [F03] | 16; harness invocations 0; preflight evidence published as full |
| V18d | **Watchdog deadline** with the stub harness sleeping [F18] | 124; group gone; stub cleanup ran |
| V18e | A malformed state file (unknown, duplicate, missing or bad value; a symlink) | 16; harness invocations 0 (P-IMPORT unit cases) |
| V18f | **Unexpected parent failure while the harness runs** (the parent's state write fails) [F15] | harness group stopped by the finalizer (143 recorded); `internal=19` recorded; exit 143 by precedence; F1 and F2 ran |
| V18g / V18h / V18i | **Pre-location:** `TMPDIR` inside the repository; `--evidence-dir` inside it; `--evidence-dir` through an external symlink into it [F20–F22] | 16; **nothing created** in the repository (whole-tree inventory unchanged); F1 only; published none |
| V18j | Stragglers left in the harness group after a normal exit [F19] | 0; stragglers stopped; group gone |

**Group E — owned-resource cleanup** (correction 3). Docker-real controls with `busybox:1.37.0` and labels `g003.control=<token>`. The control resources are created **after** the control's start inventory. P-OWN models the decisions (§15.5a).

| Control | Scenario | Expected |
|---|---|---|
| **V19** | Owned: a container under a registered project whose `working_dir` is under `work_root`, attached to a labelled network, mounting a labelled volume, and running an image the control built and registered with an `image <tag> <id>` row; capture rows written as the harness would | All four removed; inventory `removed=4 remaining=0`; exit 0 |
| **V20c** | Unrelated **container**: registered project label, foreign `working_dir`, created after the start inventory | **Survives**; `refused`; exit 18 |
| **V20n** | Unrelated **network**: registered project label, created after the start inventory, attached to no owned container, no capture row | **Survives**; `unproven`; exit 18 |
| **V20v** | Unrelated **volume**: same conditions | **Survives**; `unproven`; exit 18 |
| **V20i** | Unrelated **image**: tagged `<registered-project>-intruder:latest`, new ID, no build or capture row | **Survives**; `unproven`; exit 18 |
| V21 | Resources of an unregistered project | Untouched; absent from the inventory |
| V22 | Removal failure (shim: `docker rm` exits 1) | 18; `remaining` reported |
| V23 | Registry tampering (symlink, token mismatch, job mismatch) | Nothing removed; 18 (`--after-script`: 0, with the refusal printed) |
| V24 | `after_script` hand-off in a separate `env -i` shell | Finds the registry by the documented rule; removes only positively owned resources; exits 0 |
| V25 | Root `.env`, matching and mismatched hash | Removed only on a match |
| V26 | A resource in the start inventory under a registered label | Untouched |

Every V20 control removes its own unrelated resource afterwards and verifies the removal. No case can turn a failure into exit 0. W1–W9 and C1–C12, C8b and C8c are unchanged.

**Control count:**
- group A: 10;
- group B: 22;
- group C: 8;
- group D: 17 (V14, V15, V16, V16b, V17, V17b, V18, V18a–V18j);
- group E: 11 (V19, V20c, V20n, V20v, V20i, V21–V26).

**Total: 68 V-controls** (r2: 53), plus W1–W9 and C1–C12, C8b and C8c.

### 11.4 Owned-resource registry and `infra/ci/cleanup-owned.sh` (correction 3; r3; r2 version in Appendix R2)

**Location, permissions and modes** are unchanged from r2:
- CI registry: `$(dirname "$CI_PROJECT_DIR")/.platform-validation-owned-$CI_JOB_ID`, created by L0 (§11.2.2); local registry: `$EV/owned`;
- mode 700, never a symlink;
- `--mode finalize` (gating, 0 or 18) and `--after-script` (always 0).

**Rows** (`owned.tsv`, tab-separated; r3 additions in bold):

```text
#platform-validation-owned v1
job        <CI_JOB_ID | local-<random>>
token      <32 hex>
repo       <absolute checkout path>
work_root  <the harness's absolute $WORK>
label      g001.run=<RUN_ID>   |   platform.validation.preflight=<token>
project    <compose project name>
rootenv    <sha256>
image      <repository:tag> <image ID>   (written immediately after the run builds that tag)
container  <full ID>                     (capture row)
network    <full ID>                     (capture row)
volume     <name>                        (capture row)
```

**Capture points (positive evidence written while ownership is provable).** `owned_capture DEST` runs:
- in `run.sh`, after every start or `up` of a registered project (the README and GETTING-STARTED start blocks, every `compose … up`, `compose_browser … run`);
- first in the harness's cleanup trap;
- in `cleanup-owned.sh` itself, before any removal.

For each container of the project that is **owned under the container rule below**, it records:
- the container's full ID;
- every network ID it is attached to (`docker inspect -f '{{range $k,$v := .NetworkSettings.Networks}}{{$v.NetworkID}} {{end}}'`);
- every named volume it mounts (`{{range .Mounts}}{{if eq .Type "volume"}}{{.Name}} {{end}}{{end}}`);
- its image tag and ID (`{{.Config.Image}}`, `{{.Image}}`).

Rows are append-only. A network, volume or image ID that is in the start inventory is never recorded.

**Ownership rules** (removal needs all listed conditions):

| Kind | Removed only if | Otherwise |
|---|---|---|
| container | its ID is absent from the start inventory; **and** its `com.docker.compose.project` label is a registered project (or it carries a registered `label`); **and** its `com.docker.compose.project.working_dir` equals `repo` or lies under `work_root` | carries a registered project label but fails the conditions → `refused` (kept; 18); unregistered → ignored |
| network | its ID is absent from the start inventory; **and** a `network` capture row names that ID, or an owned container is attached to it right now | carries a registered project label without that evidence → `unproven` (kept; 18) |
| volume | its name is absent from the start inventory; **and** a `volume` capture row names it, or an owned container mounts it right now | registered label without that evidence → `unproven` (kept; 18) |
| image | its ID is absent from the start inventory; **and** an `image <tag> <id>` row (build record or capture) matches **both** the tag and the ID the tag resolves to now | a tag with a registered project's prefix (`<project>-*`) without a matching row → `unproven` (kept; 18). **The name pattern alone never authorizes removal**; it only flags the image for the report. r2's "`<project>-<service>` absent at start" fallback is withdrawn. |
| root `.env` | its SHA-256 equals the `rootenv` row | `refused` |

- **Removal order and verification:** containers (`docker rm -f -v ID`), then networks (`docker network rm ID`), volumes (`docker volume rm NAME`) and images (`docker image rm ID`, only after checking that the tag still resolves to it). Every query is then rerun.
- **Exit 0** only if no owned resource remains, nothing was refused or unproven, and every Docker command succeeded. Otherwise **18**, with `remaining`, `refused`, `unproven` and `failed` rows.
- **Never:** any `prune`; removal by name or label alone; any resource in the start inventory; any resource without positive ownership.
- **Consequence under signals.** If the harness is killed after creating something but before the next capture point, that resource is `unproven`. It is kept and reported (18), not guessed. The run is already failing in that case. The report lists it for owner cleanup, never deletion by inference.

### 11.5 `infra/ci/runner-preflight.sh` (r3; unchanged from r2 except the tool list)

As r2 §11.5 (RP1–RP13). **r3:** RP2 also requires `setsid` (util-linux) and `ps`. GNU `timeout` is still required by the harness's own probes; `validate.sh` no longer uses it.

## 12. Runner profile and hosted evidence (blocking prerequisites Q1 and Q2; R6/R7, AC7; r2; r1 §12 in Appendix R1)

### 12.1 What the harness requires (from F9)

**Profile A (required unless Codex and the owner decide otherwise).** A self-managed GitLab runner using the **shell executor**, on a **dedicated, non-production Linux host**:

- **Docker:** Docker Engine 24 or later with BuildKit, and the Compose v2 plugin 2.24 or later.
- **Host tools:** bash 4.4 or later (`inherit_errexit`), GNU coreutils/findutils/grep with `-P`, git, `timeout`, `hostname -I` or `ip`, and bash `/dev/tcp`.
- **Job user:** the job runs as the runner user in the `docker` group. That is root-equivalent on this host. It is acceptable only because the host is dedicated and disposable in purpose, and it must be acknowledged.
- **Isolation:** no other workload's Docker resources on that daemon.
- **Concurrency:** one concurrent job for this project. `resource_group` enforces it inside the project; the runner `limit`/`concurrent` value, or dedication, covers other projects.
- **Capacity:** at least 4 vCPU (8 recommended), at least 8 GiB RAM (16 recommended), and at least 40 GiB free for images and the BuildKit cache.
- **Timeouts:** the runner's `maximum_timeout` and the project's CI timeout must each be at least 3 h.
- **Egress** to the registries and module proxies used in builds: Docker Hub, `gcr.io`, `ghcr.io`, `mcr.microsoft.com`, `proxy.golang.org`/`sum.golang.org`, `registry.npmjs.org` and PyPI. The Docker Hub anonymous pull limit must be acknowledged.

**Profile A2 (acceptable only if already configured this way; no configuration change is made by G003).** A Docker executor that mounts the host Docker socket, with job-container `network_mode = "host"`, and the builds and temporary directories mounted at **identical host paths**. Without all three, the bind-mount and loopback probes cannot work.

**Not acceptable without an explicit security decision (stop condition).**
- GitLab.com hosted runners or any `docker:dind` service. These need `privileged = true`, and published ports would not be on the job's loopback.
- A socket-only Docker executor without host networking and path parity.
- Any runner that can reach production resources or credentials.

**r2 notes on §12.1.**
- **Nothing here describes the actual runner.** The profiles state requirements, not facts. No runner identity, tag, executor, capacity or permission is assumed anywhere in this Plan.
- **The registry location (§11.4) relies on the same path parity.** Profile A2 already requires the builds and temporary directories at identical host paths.
- **`runner-preflight.sh` (§11.5) turns each requirement into an executable check,** run in every pipeline and in R1.

### 12.2 Owner-supplied runner facts (Q1: **UNRESOLVED; blocking**)

The owner has confirmed only that an existing Docker-capable GitLab runner is to be used (Goal, Dependencies). Everything below is **unknown** and must be supplied by the owner, or by someone they authorize.

**Rules for supplying them:**
- Supply them as plain text in the Codex approval handoff, or in a file outside the repository whose path is given to Codex.
- Do **not** paste `config.toml` wholesale: it contains the runner token. Copy only the named fields.
- Codex checks the facts against Profile A or A2 **before approving execution**.
- R1 (§12.4) then verifies them on the runner **before S1**.
- Facts are never a substitute for R1, and R1 is never a substitute for the facts. If any fact is unavailable, execution stays blocked.

| ID | Fact the owner supplies | Where to read it | Requirement it is checked against | Status |
|---|---|---|---|---|
| F-R1 | Runner ID and description; scope (project, group or instance); self-managed or GitLab-hosted | GitLab: project → Settings → CI/CD → Runners | Self-managed (Profile A or A2). GitLab-hosted runners are a stop condition. | UNRESOLVED |
| F-R2 | The tag to use; whether it runs untagged jobs; whether it is limited to protected branches; whether it is locked to the project | same page | One tag for `.gitlab-ci.yml`. If protected-only, the evidence branches must be protected (owner action), or another runner is needed. | UNRESOLVED |
| F-R3 | **Executor** type, and for Docker executors: `privileged`, the Docker socket mount, `network_mode`, `volumes` and the builds path | runner host: the `[[runners]]` section of `config.toml` (named fields only) | Shell executor (A), or a Docker executor with socket, host network and identical paths (A2). `privileged = true` or dind is a stop condition. | UNRESOLVED |
| F-R4 | Runner version (`gitlab-runner --version`) | runner host | Records whether `RUNNER_SCRIPT_TIMEOUT` and `RUNNER_AFTER_SCRIPT_TIMEOUT` are honoured. The internal deadline does not depend on it (§11.2.1). | UNRESOLVED |
| F-R5 | Host OS and architecture; Docker Engine, Compose plugin and Bash versions; CPUs; RAM; free disk for the Docker root and the temporary directory | runner host (`uname -m`, `docker version`, `docker compose version`, `bash --version`, `nproc`, `free -g`, `df -h`) | linux/amd64; Engine ≥ 24; Compose ≥ 2.24; Bash ≥ 4.4; ≥ 4 vCPU, ≥ 8 GiB RAM, ≥ 40 GiB free | UNRESOLVED |
| F-R6 | Isolation: is the host dedicated? Other workloads or projects on the same Docker daemon; the runner's `concurrent` and `limit` | owner knowledge, `config.toml` | Dedicated, non-production, no production credentials or networks reachable; one concurrent job for this project | UNRESOLVED |
| F-R7 | Timeouts: the runner's `maximum_timeout`; the project's CI/CD general timeout | `config.toml`; project → Settings → CI/CD → General pipelines | Each ≥ 3 h (the effective `CI_JOB_TIMEOUT` is checked again by RP12) | UNRESOLVED |
| F-R8 | Network egress from the host and job; any proxy or pull-through cache; Docker Hub authentication or rate-limit status | owner and network knowledge | Registries listed in §12.1 reachable; the anonymous pull limit acknowledged | UNRESOLVED |
| F-R9 | Host port use: is anything bound to the default ports (3100, 8180, 8100, 5532, 6479, 9100, 9101) or to the variant sets (§6.3 and `run.sh` `PORTS_*`)? | runner host (`ss -ltn`) | All free | UNRESOLVED |
| F-R10 | GitLab tier of the namespace; whether merged-results pipelines are enabled; whether merge trains are enabled | group → Settings → Billing; project → Settings → Merge requests | Merged results are Premium/Ultimate (F13). Merge trains are unsupported, so attribution would return 16 (Q17). | UNRESOLVED |
| F-R11 | Artifacts: the instance maximum artifact size; the project's "Keep artifacts from most recent successful jobs" setting; who holds Developer or higher access | Admin or project settings; project → Members | `access: developer` and 90-day expiry acceptable (Q7); expected evidence size well below the maximum (measured at R1 and H0) | UNRESOLVED |
| F-R12 | Evidence retrieval mechanism (paired with A-E, §12.3) | owner decision | Owner download to a path outside the repository, or a short-lived `read_api` token supplied out of band | UNRESOLVED |
| F-R13 | Written acknowledgement that the job user's `docker` group membership is root-equivalent on that host | owner | Required for Profile A | UNRESOLVED |

### 12.3 Hosted actions requiring explicit authorization (Q2: **UNRESOLVED; blocking**)

Neither this Plan, nor its approval, nor the planning delegation authorizes any of these. Each needs the owner's or Codex's explicit, named authorization before it is taken.
- **Commits.** Creating the commits these actions push is part of each authorization. They use the repository's configured Git identity unless the authorization says otherwise.
- **Ancestors.** A push publishes every ancestor commit not yet on `origin`. If BASE includes a not-yet-pushed commit (for example the skill consolidation under route A, Q14), the first push publishes it on that branch. The authorization must cover this.

| ID | Action | Purpose | Undo |
|---|---|---|---|
| **A-R1** | From a disposable clone **outside the repository**: create branch `g003-runner-probe` at BASE plus **one** probe commit (§12.4), push it to `origin`, and let its branch pipeline run on the Q1 runner | Runner qualification before S1 | Delete the branch after the evidence is captured (part of A-R1) |
| A-H0 (optional, recommended) | Push the work-in-progress commit(s) to `g003-foundation-observability-ci` and open a **draft** MR to `staging` | Early run of the G003 job itself on the runner, after S8 | The MR stays open for H1 |
| **A-H1** | Push the final implementation commit, and the Result when available, to the same branch; its MR pipeline runs | Hosted pass on the exact submitted commit and tree | Kept for review |
| **A-H2** | Push branch `g003-gate-proof` = H1 plus one commit setting EI39 to `valid`; open a draft MR to `staging` | Hosted failure through the reviewed gate | A-H3 |
| **A-H3** | After Codex has captured H2: close its MR and delete `g003-gate-proof` | Clean-up | — |
| **A-E** | Evidence retrieval for R1, H0, H1 and H2 (logs, artifacts, pipeline and job metadata) through the F-R12 mechanism | Attributable hosted evidence | Downloads are kept outside the repository; a token is never printed or stored |

**No action pushes to `staging` or `main`, merges, publishes, deploys, or creates or changes CI/CD variables, runners or project settings.**

### 12.4 R1: runner qualification before S1 (r3; r2 version in Appendix R2)

**When.** Immediately after S0 passes, which needs Q9 and Q14 resolved, and the Plan is approved, and only with A-R1 and A-E authorized. It comes **before any repository write** (S1). It makes **no repository write**: everything happens in `$EXT/r1/probe`, a `git clone` of the local repository at BASE. Cloning a local repository reads its objects and writes nothing to it.

**Probe commit** (four files; they exist only in the disposable clone and on the deleted branch):
1. `.gitlab-ci.yml` (probe-only):

   ```yaml
   workflow:
     rules:
       - if: $CI_PIPELINE_SOURCE == "push" && $CI_COMMIT_BRANCH == "g003-runner-probe"
       - when: never
   runner-probe:
     tags: [<RUNNER_TAG>]
     timeout: 3h
     resource_group: platform-validation-host
     variables:
       GIT_STRATEGY: clone
       GIT_CLEAN_FLAGS: -ffdx
     script:
       - bash infra/ci/runner-probe-job.sh
   ```

2. `infra/ci/runner-preflight.sh`: the §11.5 script, as written for S6. Any later change is recorded in the Result.
3. `infra/ci/runner-probe-job.sh`. Under an EXIT trap, it:
   - runs the preflight, with its TSV printed to the job log (no secrets: tool versions, counts and PASS/FAIL);
   - records `docker system df` before and after, and the UTC start and end;
   - runs the **unchanged accepted baseline harness** as its own `setsid` process group, under the same parent-owned watchdog and stop sequence as `validate.sh` (§11.2.1; deadline 9900 s, then TERM to the group, a 150 s grace, then KILL). The command is `setsid bash infra/validation/run.sh --base HEAD --allow-file infra/ci/probe.allow --report-dir "$T/report" all >"$T/console.log" 2>&1 &`, with `$T` a `mktemp -d` outside the checkout;
   - **r3 caveat (F26):** the baseline harness at BASE still has the as-is `tee`. If R1 hits its deadline, the harness's own cleanup may not run. The probe job then lists, by name only, any resource whose Compose project or label matches the run's `RUN_ID` (parsed from the console's `run <id>` line), and reports it for owner cleanup. It never deletes by inference. A deadline hit already fails the decision rule;
   - prints only the `ID<TAB>STATUS` columns of `results.tsv`, the leftover-inventory **count** and the durations;
   - deletes `$T`;
   - exits with the harness status.

   It uploads no artifacts.
4. `infra/ci/probe.allow`: comments only. `--base HEAD` is the probe commit, so B-SCOPE and B-PROTECTED see no change.

**What R1 establishes (before any G003 cycle):**
- runner selection by tag;
- executor compatibility: loopback parity, bind-mount parity, port availability, non-loopback IPv4 (RP6–RP9, and the real B-PORTS-* and B-BOUNDARIES-*);
- cold image builds, including the MinIO source build and the pinned images, and registry egress;
- the browser matrices on the runner;
- the cold `all` duration against the time budget;
- disk use;
- the harness's own cleanup on that daemon.

**Decision rule** (reported to Codex before S1, and recorded in the Result):
- **Proceed to S1** only if:
  - all 50 baseline required checks PASS;
  - every RP check passes;
  - the leftover inventory is empty, apart from the renderer cache volume;
  - the cold `all` duration is **≤ 90 minutes**. The G003 harness budget is 145 minutes (2h45m − 20m), and §19 estimates that G003's contracts suite, error-output checks and observability matrix add about 25–35 minutes cold. 90 minutes keeps that estimate inside the budget with margin.
- **Otherwise stop before S1.** Return the evidence to Codex and the owner. A runner reconfiguration, a timeout change or a split (§19) is their decision. Because no implementation exists yet, **no B7 cycle is consumed**. A rerun after an owner-side change also consumes none.
- **If A-R1 is not authorized,** the B7 forecast is not credible. The Plan recommends not starting S1, and Codex decides (Q10, §19).

### 12.5 Hosted evidence plan (H0–H3; each needs its authorization from §12.3)

| Step | When | Expected evidence |
|---|---|---|
| H0 (optional, recommended) | After S8 (local CI parity passes) | The first run of the G003 job on the runner, cache warm from R1. It records the pipeline, job and runner IDs, `tested_content`, the duration, the artifact size and the published set. A failure caused by the G003 changes counts in B7; one caused by the runner environment is reported as in §12.4. |
| H1 | After S10 | A passing pipeline on the **exact submitted commit and tree**. Recorded: pipeline ID, job ID, runner, `tested_content` (`mr-source`, or `mr-merged-result` if enabled), the duration, the **authoritative** job status and final log line (§11.2.5), the pre-publication `summary.json` with 69 PASS, and the downloaded artifact's SHA-256. |
| H2 | After H1 | A **failed** pipeline: B-CONTRACT-GO, B-CONTRACT-PY and B-CONTRACT-AGREE FAIL, and `validate.sh` exits 1. The scanned artifacts show the reason. This exercises the reviewed gate, not an `exit 1` job. |
| H3 | After Codex captures H2 | The H2 MR is closed and `g003-gate-proof` is deleted; recorded |

**Evidence classification** is unchanged from r1. Local and hosted evidence are kept apart in the Result, and **local parity never substitutes for H1 or H2**. R1 is runner-qualification evidence for the unchanged baseline, not G003 acceptance evidence.

### 12.6 Order of prerequisites (r3, correction 5)

Nothing in this Plan supplies or substitutes for these. R1 and the profiles in §12.1 verify facts; they do not provide them, and they authorize nothing.

1. **Q1, owner:** the runner facts F-R1–F-R13 (§12.2). Status: **UNRESOLVED.**
2. **Q2, owner or Codex:** named authorization of A-R1, A-H0 (optional), A-H1, A-H2, A-H3 and A-E (§12.3), including commits and the publication of any unpushed ancestor. Status: **UNRESOLVED.**
3. **Codex technical review** of this Plan, using the Q1 facts to confirm Profile A or A2. Claude does not seek **execution** approval before items 1 and 2 are supplied.
4. **Q14, owner:** baseline route. Under route A, commit the completed consolidation. Status: **OPEN** (consolidation complete, uncommitted).
5. **Q9, Codex:** the fresh binding (BASE, snapshot, its SHA-256, COUNT) and explicit execution approval. Status: **PENDING** on items 3 and 4.
6. Then S0 (§15.1), then R1/S0R (§12.4), before any implementation write.

If any item is unavailable, G003 stays blocked (Goal stop condition). Codex may instead consider the split in §19.

## 13. Dependencies, images and provenance

| Item | Pin | Where | Notes |
|---|---|---|---|
| `github.com/santhosh-tekuri/jsonschema/v6` | `v6.0.3` (2026-06-28, Apache-2.0) | `infra/validation/tools/go.mod` (new) | Test tooling only. `go.sum` is recorded; `go mod verify` runs in every build. |
| `golang.org/x/text` | resolved by MVS (expected `v0.14.0`) | tools `go.mod`, indirect | Recorded in the Result |
| `golang:1.27.1-alpine`, `gcr.io/distroless/static-debian12:nonroot` | existing tags | `validation-tools.Dockerfile` | Same tag-only pins as the existing images (inherited residual). No new base image. |
| `zricethezav/gitleaks:v8.30.1` | existing | `validate.sh` artifact scan | Already used by `secret-scan` |
| `koalaman/shellcheck:v0.11.0`, `hadolint/hadolint:v2.15.1`, `mikefarah/yq:4.48.1`, `busybox:1.37.0`, Playwright `v1.63.0-noble@sha256:eff16c30…a4a27` | existing | harness | Unchanged |

**No change** to:
- web `package.json` dependencies or `package-lock.json`;
- core `go.mod`/`go.sum`;
- intelligence `pyproject.toml` dependencies or `uv.lock`;
- any runtime base image.

**No new vendor, service or infrastructure.** santhosh is a validation library used by a test tool, analogous to the Playwright image accepted in G002's Q4 (Q12).

## 14. Files (exact; `management/v0/results/G003-scope-allowlist.txt` mirrors this section; r3; r2 version in Appendix R2, r1 in Appendix R1)

### 14.1 Database, schema and contracts

- **None** of: migrations, seeds, tables, domain identities, public routes, API or MCP contracts. B-MIGRATE-SET stays `{00001_baseline.sql}`.
- The two schemas are byte-identical: both validators assert the §2 digests, and the SD controls prove that enforcement.
- Response headers are additive only:
  - `X-Request-Id` on the web and intelligence health routes;
  - `Cache-Control: no-store` on intelligence health;
  - on an unexpected web health-handler exception, the 500 with an empty body that Next already sends, plus those two headers (Q15).

### 14.2 Customized rendered files (11 new register rows → 19 total)

The r1 ten, plus `core/cmd/server/main.go`:

| Path | Register reason |
|---|---|
| `core/internal/observability/correlation.go` | diagnostic context profile, per-hop spans and completion records (G003) |
| `core/internal/observability/logging.go` | JSON record envelope (event, service, UTC time) on the redacting logger (G003) |
| `core/internal/health/checks.go` | propagates diagnostic context on the optional intelligence probe (G003) |
| **`core/cmd/server/main.go`** | **bounded JSON exit record, classified net/http error log, and bind-then-log startup record with port and app_env only (G003; three reviewed hunks; SHA-256 `987c9335…51ea`)** |
| `web/app/api/health/route.ts` | request context and completion record on liveness (G003) |
| `web/app/api/ready/route.ts` | request context, propagation and completion record on readiness (G003) |
| `web/lib/core-client.ts` | forwards diagnostic context to core readiness (G003) |
| `intelligence/src/main.py` | health completion middleware, no-store health responses and bounded startup failure (G003) |
| `intelligence/src/observability/export.py` | JSON envelope, redaction, bounded framework, library and warning records, and framework-logger takeover (G003) |
| `intelligence/src/streaming/schemas.py` | Draft 2020-12 conformance: ECMA-262 pattern anchoring, date-time assertion, keyword audit (G003) |
| `infra/docker/intelligence.Dockerfile` | conformance corpus in the check stage only (G003) |

**Already registered, content changes:** `web/package.json` (scripts only) and `docs/foundation/FOLLOW-UP.md`.

### 14.3 New files (175)

| Area | Count | Paths |
|---|---|---|
| Core | 6 | `core/internal/observability/{context_test.go,exit.go,exit_test.go,errorlog.go,errorlog_test.go}`, `core/internal/health/propagation_test.go` |
| Web | 4 | `web/lib/request-context.ts`, `web/tests/unit/request-context.test.mjs`, `web/tests/fixtures/error-probe/{wrapped-route.ts,control-route.ts}` |
| Intelligence | 6 | `intelligence/src/observability/http.py`, `intelligence/tests/{test_health_context.py,test_logging_privacy.py,test_error_output.py,test_conformance.py,conformance_exchange.py}` |
| Contracts | 128 | as in r1: `contracts/conformance/README.md`, `contracts/conformance/intelligence/expectations.json`, and 126 fixtures (EV01–EV14, EI01–EI50, XV01–XV15, XI01–XI34, NV01–NV02, NI01–NI05, SN02–SN07), with r1's anchored per-directory allow-list expressions. SD01 and SD02 have no fixture. |
| Validation tools | 20 | `infra/validation/tools/{go.mod,go.sum}`; `cmd/{contractcheck,httpprobe,logcheck}/{main.go,main_test.go}` (6); `internal/conformance/{validate.go,audit.go,derive.go,ndjson.go,produce.go,conformance_test.go}` (6); `internal/logrecord/{completion.schema.json,events.json,logrecord.go,logrecord_test.go}` (4); `infra/docker/validation-tools.Dockerfile` and `.dockerignore` (2). r1 stated "(17)" for a list of 18 (F23). |
| CI | 7 | `.gitlab-ci.yml`, `infra/ci/{validate.sh,lib.sh,validate-controls.sh,cleanup-owned.sh,runner-preflight.sh,unchanged.allow}` |
| Harness | 1 | `infra/validation/required-checks.tsv` |
| Docs | 1 | `docs/foundation/OPERATIONS.md` (r1 coverage, plus: error-output handling, attribution profiles, the finalization and exit-code model, the registry and cleanup, the runner requirements and R1. It names no runner tag, GitLab path or identity literal.) |
| Management | 2 | `management/v0/results/G003-scope-allowlist.txt` (**created at S1**, with the full final content of this section) and `management/v0/results/G003-foundation-observability-ci-result.md` (after execution) |

Count: 6 + 4 + 6 + 128 + 20 + 7 + 1 + 1 + 2 = 175.

The probe-only files of R1 (`.gitlab-ci.yml` probe variant, `infra/ci/runner-probe-job.sh`, `infra/ci/probe.allow`, and a preview of `runner-preflight.sh`) exist only in the disposable R1 clone and on the deleted probe branch. They are **not** repository paths and are not in the allow-list.

### 14.4 Modified project-owned files

- `infra/validation/run.sh`, `infra/validation/lib.sh`: r2's changes, plus (r3):
  - the one-line TERM/INT/HUP-ignoring `tee`;
  - `owned_capture` at every start or `up` of a registered project and first in the cleanup trap;
  - `image <tag> <id>` rows after every build;
  - the `MAIN_GO_G003_SHA` pin.
- `infra/validation/render-customizations.tsv` (+11 rows).
- `README.md`: health lists all six routes and the headers; new blocks `platform:contracts` and `platform:ci`; a link to OPERATIONS.md; the existing blocks are unchanged.

### 14.5 Not touched

- `core/cmd/migrate/**`, `core/internal/{config,mcp,migrate}/**`, `core/migrations/**`, `core/go.{mod,sum}`.
- `core/internal/health/{health.go,health_test.go}`, `core/internal/observability/observability_test.go`.
- `web/app/{layout.tsx,page.tsx,globals.css}`, `web/components/**`, `web/lib/{config,appearance,tokens,utils}.ts`, `web/package-lock.json`, all web configuration files, `web/tests/browser/**`, `web/tests/fixtures/token-scopes/**`.
- `intelligence/{pyproject.toml,uv.lock}`, every other `intelligence/src/**` file, the existing `intelligence/tests/*.py`.
- `contracts/README.md`, `contracts/intelligence/**`.
- `infra/compose/**`, every other Dockerfile and `.dockerignore`, `infra/{local,minio,postgres,security}/**`.
- `infra/validation/run-approved{,-controls}.sh`.
- `.env.example`, `.gitignore`, `.archetype/**`.
- `skills/**`, including the in-progress consolidation, which belongs to the owner. AGENTS.md, CLAUDE.md, `docs/ARCHITECTURE.md`, SECURITY.md, DECISIONS.md, `docs/history/**`, the other `docs/foundation/*.md`.
- ROADMAP, BACKLOG, the G003 Goal and Spec, PROGRESS (Codex-owned), and every G000–G002 artifact.
- The tool-local `.claude/settings.local.json`: never read or modified.

### 14.6 Temporary writes (none persist in the repository)

| Write | Location | Removal |
|---|---|---|
| S0 gate copy and evidence | `$EXT/g003-exec/` (outside the repository) | kept as evidence outside the repository |
| R1 probe clone | `$EXT/r1/probe` | deleted after the probe branch is deleted |
| Go module cache for the tools module (S2) | `$EXT/gomod` | deleted at the end |
| Component-check, runtime and probe images | `g003-$EXEC-{core,web,intelligence,vtools}{,-check}`, `g003-$EXEC-web-probe`, labelled `g003.exec=$EXEC` | removed at the end, by label |
| Ad hoc error-output probe inputs and outputs (S3–S5) | `$EXT/s3`, `$EXT/s4`, `$EXT/s5` (environment files mode 600) | canary-bearing inputs deleted after use; outputs kept only after a canary scan |
| Harness work, Compose projects, images, root `.env` | as today (`project-validation-*`, `g001-*-<run>`; root `.env` created and hash-removed by the root suite) | the harness's own cleanup, then `cleanup-owned.sh` verification in CI |
| CI owned registry | `$(dirname "$CI_PROJECT_DIR")/.platform-validation-owned-$CI_JOB_ID` (CI only) | F9 or `after_script` |
| Local CI-parity copy (S8) | `$EXT/s8/clone`, with a scratch commit (Q6) | deleted after the run |
| Hosted evidence downloads | `$EXT/hosted/` | kept outside the repository |

## 15. Execution sequence, gates and provenance

### 15.1 S0: read-only, fail-closed starting-state gate (adapted from approved G002 r6 S0a)

- **Reuse.** The committed wrapper `run-approved.sh`, with `--check-only`, already provides the snapshot-digest, format, count and entry checks, so G003 needs no bootstrap writes. S0 adds what the wrapper does not check:
  - the HEAD binding;
  - the NUL-safe starting state;
  - the resource and port inventory;
  - evidence containment.

  It keeps G002 r6's read-only contract: `realpath -m` containment before any `mkdir`, and Git with `--no-optional-locks -c core.fsmonitor=false`, `GIT_OPTIONAL_LOCKS=0` and redirection variables unset.
- **Running it.** The script runs from **outside the repository** after byte-exact extraction from the block below. Its SHA-256 must equal **`0409148a630f32cc79642fd8ae65b90a53001eba379775bd96a3c1f313a66723`**; a different hash means stop. BASE, SNAPSHOT, SNAPSHOT_SHA256 and COUNT are taken verbatim from Codex's approval record. Nothing is regenerated, and the G002 snapshot is never reused.

```bash
bash <ext>/g003-s0-gate.sh <repo> <BASE> <approved snapshot> <its SHA-256> <COUNT> <ext>/g003-exec/s0
```

| Step | Verifies | Exit |
|---|---|---|
| 0 | Six arguments. BASE is 40 hex, the digest is 64 hex, COUNT is 1–5. The repository resolves. The evidence path resolves **outside** the repository, including through traversal and symlinks (dangling ones included), **before any write**. | 26 |
| 1 | `HEAD` = BASE | 20 |
| 2 | The snapshot is read once; the SHA-256 of those bytes equals the pinned digest | 21 |
| 3 | Only those bytes are parsed: `<64hex>  <path>`, no escaped entries, relative, no `.` or `..`, unique, inside the G003 input set (Goal, Spec, PROGRESS.md, this Plan, ARCHITECTURE.md), exactly COUNT entries, and the Plan present | 22 |
| 4 | Every entry is a regular non-symlink file whose SHA-256 matches. This holds whether or not the entry is committed. | 23 |
| 5 | NUL-safe `status --ignored`: every differing path is `??` or ` M` **and** is a snapshot path. Nothing staged, deleted, renamed, ignored (including `.env` and `.env.local`) or extra is allowed. **The single exception** is the exact ignored path `.claude/settings.local.json`, which is counted and reported, never read (F12). | 24 |
| 6 | No root `.env`; Compose 2.24 or later; no containers, volumes or networks for the project; default ports free; the Docker inventory written to evidence | 25 |

**Exit 0 is the only way to proceed.** Any other exit stops the work before the first repository write.

Exact S0 script (SHA-256 `0409148a630f32cc79642fd8ae65b90a53001eba379775bd96a3c1f313a66723`; extract the lines between the fences, including the final newline):

```bash
#!/usr/bin/env bash
# G003 S0: read-only, fail-closed starting-state gate (G003 Plan §15.1, r1). Adapted from the
# approved G002 r6 S0a gate with the same read-only contract. Needs only bash (>= 4.4), GNU
# coreutils, git and docker. Writes nothing inside REPO, including .git: Git runs without optional
# locks, so the index is never refreshed. The evidence destination is resolved and proven outside
# REPO before anything is created.
# Usage: bash g003-s0-gate.sh REPO BASE SNAPSHOT SNAPSHOT_SHA256 COUNT EVIDENCE_DIR
#   BASE, SNAPSHOT, SNAPSHOT_SHA256 and COUNT are the values pinned by Codex's Plan approval.
# Exit: 0 pass; 20 HEAD; 21 snapshot digest; 22 snapshot format/paths/count; 23 input hash;
#       24 starting state; 25 local resources or tools; 26 usage or evidence path.
set -euo pipefail
shopt -s inherit_errexit

fail() { printf 'S0 FAIL %s: %s\n' "$1" "$2" >&2; exit "$1"; }
[[ $# -eq 6 ]] || fail 26 "usage: REPO BASE SNAPSHOT SNAPSHOT_SHA256 COUNT EVIDENCE_DIR"

# Git inspection must not write repository metadata (no index refresh, no fsmonitor) and must not
# be redirected by the caller's environment.
export GIT_OPTIONAL_LOCKS=0
unset GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE GIT_OBJECT_DIRECTORY GIT_ALTERNATE_OBJECT_DIRECTORIES \
  GIT_COMMON_DIR GIT_NAMESPACE GIT_CONFIG_PARAMETERS GIT_CONFIG_COUNT
g() { git --no-optional-locks -c core.fsmonitor=false -C "$REPO" "$@"; }

# 0. Arguments, and every location resolved before any write.
REPO="$(realpath -e -- "$1" 2>/dev/null)" || fail 26 "repository path does not resolve"
[[ -d "$REPO" ]] || fail 26 "repository path is not a directory"
BASE="$2" SNAP="$3" SNAP_SHA="$4" COUNT="$5"
[[ "$BASE" =~ ^[0-9a-f]{40}$ ]] || fail 26 "BASE must be a full 40-hex commit"
[[ "$SNAP_SHA" =~ ^[0-9a-f]{64}$ ]] || fail 26 "SNAPSHOT_SHA256 must be 64 lowercase hex"
[[ "$COUNT" =~ ^[1-5]$ ]] || fail 26 "COUNT must be 1-5"
EV="$(realpath -m -- "$6" 2>/dev/null)" || fail 26 "evidence path does not resolve"
inside_repo() { [[ "$1" == "$REPO" || "$1" == "$REPO"/* ]]; }
inside_repo "$EV" && fail 26 "evidence destination resolves inside the repository: $EV"
# Only now may anything be created; the resolved path is re-verified before every write.
mkdir -p -- "$EV"
ev_ok() {
  local now
  now="$(realpath -e -- "$EV" 2>/dev/null)" || fail 26 "evidence directory vanished"
  if [[ "$now" != "$EV" || ! -d "$now" ]] || inside_repo "$now"; then fail 26 "evidence directory changed: $now"; fi
}
ev_ok

PLAN=management/v0/plans/G003-foundation-observability-ci-plan.md
PATH_RE='^(management/v0/(goals/G003-foundation-observability-ci\.md|specs/G003-foundation-observability-ci-spec\.md|PROGRESS\.md|plans/G003-foundation-observability-ci-plan\.md)|docs/ARCHITECTURE\.md)$'
LINE_RE='^([0-9a-f]{64})  ([^\].*)$'
cd -- "$REPO"

# 1. Baseline commit.
[[ "$(g rev-parse HEAD)" == "$BASE" ]] || fail 20 "HEAD is not $BASE"

# 2. Snapshot: read once; verify the digest of exactly those bytes.
bytes="$(cat -- "$SNAP"; printf x)" || fail 21 "cannot read the snapshot"
bytes="${bytes%x}"
[[ "$(printf '%s' "$bytes" | sha256sum | cut -d' ' -f1)" == "$SNAP_SHA" ]] || fail 21 "snapshot digest mismatch"

# 3. Parse only those bytes: format, relative unique approved paths, exact count, Plan present.
[[ -n "$bytes" && "$bytes" == *$'\n' ]] || fail 22 "snapshot is empty or lacks a final newline"
declare -A want=()
n=0
while IFS= read -r line; do
  [[ "$line" =~ $LINE_RE ]] || fail 22 "malformed entry: $line"
  sha="${BASH_REMATCH[1]}" path="${BASH_REMATCH[2]}"
  [[ "$path" != /* && "/$path/" != */./* && "/$path/" != */../* ]] || fail 22 "non-relative path: $path"
  [[ "$path" =~ $PATH_RE ]] || fail 22 "path outside the approved set: $path"
  [[ -z "${want[$path]+x}" ]] || fail 22 "duplicate path: $path"
  want[$path]="$sha"; n=$((n + 1))
done <<<"${bytes%$'\n'}"
(( n == COUNT )) || fail 22 "expected $COUNT entries, found $n"
[[ -n "${want[$PLAN]+x}" ]] || fail 22 "the Plan is not an entry"

# 4. Every input is a regular file matching its approved hash (committed or not).
for path in "${!want[@]}"; do
  [[ -f "$path" && ! -L "$path" ]] || fail 23 "missing or non-regular input: $path"
  [[ "$(sha256sum -- "$path" | cut -d' ' -f1)" == "${want[$path]}" ]] || fail 23 "hash mismatch: $path"
done

# 5. Starting state (NUL-safe, no optional locks): every path that differs from HEAD is a snapshot
#    path and is untracked or unstaged-modified; nothing staged, deleted, renamed or ignored.
#    Snapshot paths that are committed unchanged at BASE are allowed (step 4 bound their bytes).
#    The single exception is the ignored, tool-local Claude Code settings file named below: it is
#    never a repository input, is not read, and is reported in the evidence.
TOOL_LOCAL=.claude/settings.local.json
ev_ok
g status --porcelain=v1 -z --untracked-files=all --ignored >"$EV/s0-status.z" || fail 24 "git status failed"
ndiff=0 ntool=0
while IFS= read -r -d '' rec; do
  xy="${rec:0:2}" path="${rec:3}"
  if [[ "$xy" == "!!" && "$path" == "$TOOL_LOCAL" ]]; then ntool=$((ntool + 1)); continue; fi
  [[ "$xy" == "??" || "$xy" == " M" ]] || fail 24 "unexpected status '$xy': $path"
  [[ -n "${want[$path]+x}" ]] || fail 24 "differing path not in the snapshot: $path"
  ndiff=$((ndiff + 1))
done <"$EV/s0-status.z"

# 6. Local resources the execution must never touch, and the required tools.
project="$(sed -n 's/^COMPOSE_PROJECT_NAME=//p' .env.example)"
[[ -n "$project" ]] || fail 25 "cannot read COMPOSE_PROJECT_NAME from .env.example"
[[ ! -e .env && ! -L .env ]] || fail 25 "a root .env exists"
cv="$(docker compose version --short 2>/dev/null)" || fail 25 "docker compose unavailable"
[[ "$cv" =~ ^v?([0-9]+)\.([0-9]+) ]] || fail 25 "unparsable compose version: $cv"
(( BASH_REMATCH[1] > 2 || (BASH_REMATCH[1] == 2 && BASH_REMATCH[2] >= 24) )) || fail 25 "Compose >= 2.24 required (found $cv)"
for kind in "ps -a -q" "volume ls -q" "network ls -q"; do
  # shellcheck disable=SC2086
  out="$(docker $kind --filter "label=com.docker.compose.project=$project")" || fail 25 "docker query failed"
  [[ -z "$out" ]] || fail 25 "existing $project resources ($kind)"
done
for port in 3100 8180 8100 5532 6479 9100 9101; do
  if (exec 3<>"/dev/tcp/127.0.0.1/$port") 2>/dev/null; then fail 25 "port $port is busy"; fi
done
ev_ok
{ docker ps -a --format 'container {{.Names}}'; docker volume ls --format 'volume {{.Name}}'
  docker network ls --format 'network {{.Name}}'; docker image ls --format 'image {{.Repository}}:{{.Tag}}'; } |
  LC_ALL=C sort >"$EV/s0-docker-inventory.txt"

ev_ok
printf 'S0 PASS: HEAD %s; snapshot %s; %d inputs verified; %d differing path(s), all snapshot paths; %d tool-local ignored file(s); no .env, no %s resources, default ports free; compose %s\n' \
  "$BASE" "$SNAP_SHA" "$n" "$ndiff" "$ntool" "$project" "$cv" | tee "$EV/s0-gate.txt"
```

### 15.1a S0 is unchanged; the baseline is PENDING (r3; r2 text in Appendix R2)

- **The script is byte-identical** (`0409148a630f32cc79642fd8ae65b90a53001eba379775bd96a3c1f313a66723`, re-extracted and verified at r3). PATH_RE, the COUNT range, the single tool-local exception (Q13) and the exit codes are unchanged. No exception has been added for the consolidation.
- **Current state (§2.1).** The consolidation is complete but uncommitted. S0 would stop at step 5 with exit 24 on the first `skills/vestive-design-skill/...` record. This is the intended outcome.
- **What unblocks S0.** Every item in §12.6, in order. Route A (commit the consolidation; bind BASE) remains recommended. Any other route needs a reviewed Plan revision of the S0 inputs, never an execution-time exception.
- **r1 demonstrations.** r1's demonstrations (§15.2) remain valid for the unchanged script.

### 15.2 Planning-time demonstration (read-only for the real repository)

- **Setup.** Run on 2026-10-05 with the exact script above.
  - shellcheck v0.11.0 (`-S warning`) is clean.
  - Every case ran in a **disposable `git clone`** of the baseline, in the job's scratch directory, with an untracked stand-in Plan.
  - Snapshots were built from the clone's files. These snapshots demonstrate the gate only; they are **not approval snapshots**.
- **Measurement.** Around every invocation: a whole-clone inventory including `.git` (path, type, size, nanosecond mtime, inode) and `.git/index` hash plus stat data.
- **Real repository.** Its inventory, including `.git`, and its index were unchanged across the whole demonstration.

| Case | Input | Want | Got | Tree / index |
|---|---|---|---|---|
| P | Plan (untracked) + Goal + Spec + PROGRESS, COUNT 4 | 0 | 0, `4 inputs verified; 1 differing path(s), all snapshot paths; 0 tool-local ignored file(s); no .env, no beginify resources, default ports free; compose 2.32.4-desktop.1` | same / same |
| P5 | As P plus ARCHITECTURE.md, COUNT 5 | 0 | 0 | same / same |
| N1 | Wrong pinned digest | 21 | 21 | same / same |
| N2 | COUNT 3 with 2 entries | 22 | 22 | same / same |
| N3 | Duplicate entry | 22 | 22 | same / same |
| N4 | Absolute path | 22 | 22 | same / same |
| N5 | `README.md` entry (outside the set) | 22 | 22 | same / same |
| N6 | Plan not an entry | 22 | 22 | same / same |
| N7 | Goal modified after snapshot | 23 | 23 | same / same |
| N8 | Extra untracked file | 24 | 24 | same / same |
| N9 | Plan staged (`A `) | 24 | 24 | same / same |
| N10 | Root `.env` (`!!`) | 24 | 24 | same / same |
| N11 | HEAD moved by an empty commit | 20 | 20 | same / same |
| N12 | Unlisted tracked file modified | 24 | 24 | same / same |
| T1 | Ignored `.claude/settings.local.json` present (ignored as in the real checkout) | 0 | 0, `1 tool-local ignored file(s)` | same / same |
| T2 | Ignored `.env.local` present | 24 | 24 (`!!` `.env.local`) | same / same |
| T3 | Untracked `.claude/other.json` | 24 | 24 (not in the snapshot) | same / same |
| A1 | Evidence path inside the repository (nonexistent) | 26 | 26; **not created** | same / same |
| A2 | Evidence through an external symlink into the repository | 26 | 26; **not created** | same / same |
| A3 | Abbreviated BASE | 26 | 26 | same / same |
| C1 | Stale-index fixture (tracked mtimes changed), positive path | 0 | 0 | same / **same** |
| C0 | Ordinary `git status` (optional locks enabled) on that fixture afterwards | index rewritten | rewritten: the detector is non-vacuous | — |

### 15.3 Sequence (r3; the r2 order, accepted by Codex, with r3 gate additions; r2 and r1 versions in the appendices)

| Step | Work | Gate (exact commands in §15.6) |
|---|---|---|
| S0 | §15.1 gate, then the wrapper `--check-only`, then `run-approved-controls.sh` (W1–W9). **Needs every prerequisite in §12.6.** | All exit 0, or stop before any write |
| **S0R** | **R1 runner qualification** (§12.4). Needs Q1, A-R1 and A-E. A disposable clone outside the repository; **no repository write.** | The §12.4 decision rule is met and reported to Codex; otherwise stop before S1 (no cycle consumed) |
| S1 | **First repository write: `management/v0/results/G003-scope-allowlist.txt`**, with the full §14 content. B-SCOPE checks that changed paths are a *subset* of the allow-list, so every later intermediate state passes it.<br>Then: the contracts corpus, `expectations.json` (including `schema_negative`, SD01 and SD02), the conformance README, the `schemas.py` corrections, `test_conformance.py` and `conformance_exchange.py`, and the `intelligence.Dockerfile` check-stage `COPY`. | The intelligence check image builds. `pytest -q` passes, with the conformance, SN-class and SD tests included and none skipped. |
| S2 | The tools module (`go get …@v6.0.3 && go mod tidy` in a container, the only network step), the commands, tests, `events.json` and the Dockerfile | The tools check stage passes; `contractcheck corpus` agrees with the expectations; `contractcheck schema-negative` reports every declared class, a passing `sn-derive` and SD01/SD02 as digest mismatches |
| S3 | Core: `correlation.go`, `logging.go`, `checks.go`, `exit.go`, `errorlog.go`, the two `main.go` hunks, and the tests | `core-check` (gofmt, vet, all tests) passes. Then the core runtime image is built and **CE1–CE3** are run ad hoc (§15.6). CE2: exactly one `core.exit`/`listen_failed` and no `core listening`. CE3: exactly one bounded `core listening`. No canary in any. `main.go` must hash to `987c9335…51ea`. |
| S4 | Web: `request-context.ts`, `core-client.ts`, both routes, the scripts, the unit test and the two error-probe fixtures | `web-check` (lint, type check, unit) passes. The production build passes. **WE0–WE6** (with WE3a and WE3b) run ad hoc on a disposable probe copy (§15.6). This is the first **runtime** evidence of F15; P-NEXT is source evidence only. A deviation stops for review (R14). |
| S5 | Intelligence: `http.py`, `export.py`, `main.py`, `test_error_output.py` and the other tests | The intelligence check passes (ruff, format, all tests). The runtime image is built and **IE1/IE2** are run ad hoc (§15.6). The uvicorn lifespan and h11 templates are confirmed against the installed 0.54.0 package. |
| **S6** | **Integration, in one step, before any static or provenance run reads the result:**<br>• (a) CI files: `.gitlab-ci.yml` (the Q1 tag), `infra/ci/validate.sh`, `lib.sh`, `validate-controls.sh`, `cleanup-owned.sh`, `runner-preflight.sh`, `unchanged.allow`;<br>• (b) documentation: the README `platform:contracts` and `platform:ci` blocks, `docs/foundation/OPERATIONS.md`, the FOLLOW-UP labels (§17);<br>• (c) `infra/validation/required-checks.tsv` (69 IDs);<br>• (d) every §10 harness change: the contracts suite, the error-output checks, the `obs_*` functions, B-HEALTH-ROOT, B-REQUIRED-CHECKS, the register (+11), `APPLICABLE` (+7), the `MAIN_GO_G003_SHA` pin and imports literal, the B-SHELL, B-HADOLINT, B-DOCKERIGNORE, B-DEPS, B-WEB-BOUNDARY and B-MCP-SOURCE changes, and the `--owned-registry` and `--secrets-file` options. | In order, all in the same cycle:<br>• **G6a** shellcheck on every script, plus `--include=SC2251` (F27) for `infra/ci/*.sh`;<br>• **G6b** `validate-controls.sh` standalone: groups A–E (68 controls), including V4's nested real `contracts` run, the external-signal supervision controls V18–V18j, and the Docker-real cleanup controls V19–V26;<br>• **G6c** the wrapper with `static provenance contracts`: **all PASS**. This is the G002 lesson: a shared harness change is followed by the static suite in the same cycle. |
| S7 | — | The wrapper with `root migrations variants`: **all PASS** (B-HEALTH-ROOT, the three error-output checks, the B-OBS-* checks, and every inherited root and variant check) |
| S8 | **Local CI parity:** `validate.sh` in a disposable scratch-commit copy (Q6). Then **H0**, if A-H0 is authorized. | `validate_rc=0`, `tested_content=local`; H0's evidence recorded (§12.5) |
| S9 | Two read-only reviewers (§20); triage; fixes | Findings resolved or recorded. The suites covering each changed file are rerun, and a harness or CI change always reruns G6a, G6b and the `static` suite in the same cycle. |
| S10 | **Final local run with no edits during the run:** the wrapper with `all` (§15.4); the cleanup inventory; the Result draft | `overall_rc=0` |
| S11 | H1 and H2 (with A-H1, A-H2 and A-E), then H3 (A-H3); the final Result; stop for Codex | A hosted pass on the exact commit; a hosted failure for the intended reason |

**r1 → r2 step map:**
- S0 → S0.
- New: S0R.
- S1–S5 → S1–S5, plus the ad hoc error-output probes; the allow-list moves to S1.
- r1 S6 (harness) + S7 (CI) + S8 (docs) → S6, with the allow-list already created at S1.
- r1 S9 → S7 + S8.
- r1 S10 → S9; r1 S11 → S10; r1 S12 → S11.

The **cycle-4 checkpoint** is recorded *after* cycle 4 closes and before cycle 5 begins, wherever that falls (§19).

### 15.4 Final local invocation (S10)

```bash
bash infra/validation/run-approved.sh \
  --snapshot <approved snapshot> --snapshot-sha256 <its SHA-256> --count <COUNT> \
  --path-re '^(management/v0/(goals/G003-foundation-observability-ci\.md|specs/G003-foundation-observability-ci-spec\.md|PROGRESS\.md|plans/G003-foundation-observability-ci-plan\.md)|docs/ARCHITECTURE\.md)$' \
  --plan management/v0/plans/G003-foundation-observability-ci-plan.md \
  -- --base <BASE> --allow-file management/v0/results/G003-scope-allowlist.txt --report-dir <ext>/g003-exec/s10/report all
echo "overall_rc=$?"
```

- **Arguments.** Every non-Plan snapshot entry becomes `--expect-hash` and an exact `--exception`, as the wrapper builds them today.
- **Outcome.** Exit 0 is the only passing result. Any other status is a failed run: it is recorded, and the handoff stops.
- **Scope.** B-SCOPE checks the exact §14 allow-list against BASE. The harness enumerates with `--exclude-standard`, so the globally ignored tool-local file does not enter B-SCOPE, the copies or the scans.
- **Inputs.** `<BASE>`, the snapshot, its SHA-256 and `<COUNT>` are Codex's Q9 binding, never self-generated.

### 15.5 r2 planning-time probes and source evidence (scratch only; no repository write)

**r3 note:** the P-FIN row below is superseded by P-MODEL (§15.5a). Codex found that r2's model omitted `stage` from aggregation (exit 0 with `stage=19`); the r3 model fixes this. The other r2 probes remain valid as recorded.

These are disposable planning probes under `/tmp/g003-plan-r2/`, the scratch root Codex will see referenced. They are **not** implementation and **not** execution evidence. Runtime confirmation of each point is an S3–S6 gate.
- **Images.** Every container used was already present locally, run with `--pull never`, `--network none` and `--rm`.
- **No application.** No application image was built or run.

| Probe | What was run | Result | Supports |
|---|---|---|---|
| **P-GO** | `main.go` (SHA-256 `de08b71b…be3a`): a stdlib-only scratch program, run with `go run` in `golang:1.27.1-alpine` (`GOTOOLCHAIN=local`), mirroring today's core behaviour and the planned adapters. Output: `out.txt` (`28f39558…1288`). | **Today's behaviour (section A):** plain-text `http: superfluous response.WriteHeader call from main.serve.func1 (main.go:63)`; `http: panic serving 127.0.0.1:…: boom CANARY-7f3a9c` with a full goroutine stack (twice: the client retried); `fmt.Fprintln` of a listen error printing `listen tcp: lookup bad-CANARY-7f3a9c.invalid on …`. That is 3 canary-bearing non-JSON lines. `panic(http.ErrAbortHandler)` printed nothing, and a query containing `;` printed nothing.<br>**Planned behaviour (section B):** with a classifying `ErrorLog` writer and a bounded exit record, 5 lines, every one valid JSON, with `canary=false` (`superfluous_write_header`, `handler_panic` ×2, `listen_failed`, `config_invalid`). | F16, F17; §5.6; §7.1 |
| P-GO-SRC | `grep -n "logf("` over `/usr/local/go/src/net/http/server.go` in the same image | Call sites for hijacked and superfluous `WriteHeader`, hijacked `Write`, invalid `Content-Length`, conflicting length headers, `panic serving`, TLS handshake and `Accept` errors; `logf` falls back to `log.Printf` when `ErrorLog` is nil. There is no semicolon-query log call. `ErrAbortHandler` suppresses the stack. | F17; the §5.6 categories |
| **P-NEXT** (source, not a runtime probe) | Next.js `v16.3.8` tag: `packages/next/src/build/templates/app-route.ts` and `packages/next/src/server/route-modules/route-module.ts` | The route error is rethrown to the outer catch, which calls `routeModule.onRequestError(…, silenceLog=false, …)`, then `sendResponse(…, new Response(null, { status: 500 }))`. `onRequestError` calls `logErrorWithOriginalStack(err, 'app-dir')` or `console.error(err)`. **Not run:** Next cannot be executed without installing dependencies or building, which this handoff forbids. The real-framework probe is WE0–WE6, first at S4 and then permanently as B-WEB-ERROR-OUTPUT. | F15; §5.3; §7.2 |
| P-UVICORN (source) | uvicorn `0.54.0` tag: `config.py` (`LOGGING_CONFIG`, `load_app`) and `server.py` | `uvicorn` has a stderr text handler, `uvicorn.error` propagates to it, and `uvicorn.access` writes to stdout with the request line. `load_app` catches only `ImportFromStringError`. The `server.py` templates are those listed in §5.6. Runtime confirmation: IE1 and IE2 (S5, B-INTEL-ERROR-OUTPUT). | F18; §7.3 |
| **P-FIN-0** | `gotcha.sh` (`238184db…6d56`) | `f \|\| rc=$?` → rc 0, and the function continued past `false`. `( set -e; f ) \|\| …` → rc 0, and it continued. `set +e; ( set -e; f ); rc=$?` → rc 1. | §11.2.1 (the phase form) |
| **P-FIN** | `model.sh` (`3b8233dd…a898`): a control-flow model of §11.2 with stub phases; 11 fault cases; `cases.tsv` (`663900c2…ee30`) | none→0 (published); harness→1; metadata→16 (harness not run; postcheck and cleanup ran; nothing published); malformed report→14; scanner error→17 (nothing published); planted canary in an excluded raw file→15 (published `WITHHELD` only); drift→13; cleanup→18; publish failure→19; unexpected error outside a phase→19 (finalizer ran); SIGTERM during the harness→143 (postcheck and cleanup ran; nothing published). In every case postcheck and cleanup ran; no failure produced 0. | §11.2.3–§11.2.5; groups C and D |
| P-TERM | `termtrap.sh` (`d4e04cbd…2b27`) under `timeout --signal=TERM --kill-after=5s 1s` (GNU coreutils 9.4, Bash 5.2.21) | `timeout` returned 124, and the child's EXIT trap ran (`tt.out`). | §11.2.1 (the internal deadline keeps the harness's own cleanup) |
| **P-ATTR-0** | A Git fixture: target T, source S and merge M (`--no-ff`, parents T and S) under `refs/merge-requests/1/merge`; a depth-1 fetch and checkout of M | `.git/shallow` = M. `git cat-file commit HEAD` lists both parents (T, then S). `git rev-parse --verify HEAD^2` fails at the shallow boundary. | §11.2.4 (the raw parent check) |
| **P-ATTR** | `classify.sh` (`1e0dfe34…2984`), the prototype of §11.2.4, over the fixture with `env -i` and GitLab-defined variables; `cases.sh` (`6cd60dab…5dd5`), results in `results.tsv` (`d5e7a3fe…498e`) | **21 of 21 as expected.** P1 and P2 `branch-push`; P3 `mr-source`; P4 `mr-merged-result`; P5 `local` (all exit 0). N1–N16 each exit 16 with their specific reason. | Correction 2; group B |
| **P-SCAN** | `zricethezav/gitleaks:v8.30.1 dir` with the repository's `gitleaks.toml`, `--redact --exit-code 23`, over scratch directories | clean → **0**; a planted generated `ghp_` token → **23** (`leaks found: 1`); a missing configuration file → **1**; a missing path → **1** | F22; §11.2.3 F3 and F7; V5c and V6a |

Two observations are recorded for Codex:
- **(i)** The Go 1.27.1 standard library no longer logs semicolon queries. r1 assumed nothing about it, and r2 asserts nothing about it.
- **(ii)** gitleaks reports both a fatal error and a default leak with exit 1. That is why `--exit-code 23` is required to tell a scanner failure from a finding.

### 15.5a r3 planning probes and models (scratch only, under `/tmp/g003-plan-r3/`; no repository write; no Docker resource created)

These are disposable, non-application planning checks. They are **not** G003 implementation and **not** execution evidence. Runtime confirmation is the S3–S6 gates and the group D and E controls.

| Probe | What was run | Result | Supports |
|---|---|---|---|
| **P-STATE** | `gotcha-state.sh` (`0e34f825…3eef`): r2's phase helper wrapping a location body | `location status=0 EV=<unset> REG=<unset>`. The body's `mktemp -d` had already created a directory, which was then removed. This reproduces Codex's finding. | F24; §11.2.1 |
| **P-SETSID** | `setsid-probe.sh` (`afd70614…cfd9`; util-linux 2.39.3) | `$!` = pid = pgid = sid of the background `setsid` child. `kill -TERM -- -$!` gave rc 143 and the group was gone. | F25 |
| **P-TEE** | `v5.sh` (`3376f082…1b70`): a stub shaped like the current `run.sh` under the supervisor's watchdog plus group TERM; outputs in `p-tee.txt` (`0f1a2d67…40fa`); variants `v2.sh`–`v4.sh` | **As-is** (`exec > >(tee …)`): harness rc **141** (SIGPIPE), cleanup **NO**, three of three runs. **`tee` ignoring TERM/INT/HUP:** rc 143, cleanup **ran**, log tail `cleanup: done`, three of three runs. Under GNU `timeout`, a process-substitution variant also skipped its EXIT trap. A TERM sent only to the leader is deferred until the foreground child ends. | F26; §10 one-line `tee` change; §11.2.1 (supervisor-owned deadline and group TERM) |
| **P-IMPORT** | `state-import-test.sh` (`89004f1b…bfa5`) against the model's `state_import`; output `state-import.txt` (`d5c22f09…d03f`) | 8 of 8 as expected. Valid input accepted and assigned. Unknown key, duplicate, bad value, missing key, `1;touch /tmp/pwn` injection (no side effect), lowercase key and symlink all rejected. | §11.2.1 state import; V18e |
| **P-MODEL** | `vmodel.sh` (`a18c0e65…863a`): a control-flow model of §11.2 with real `setsid`, watchdog, signals, `git` postcheck, literal scan, all-or-nothing staging, candidate selection, atomic publish, report, scrub and the aggregation invariant. Stubs from `stubs.sh` (`bdf2698b…f0de`). Driver `cases.sh` (`6d8f6feb…60a0`), with a fresh scratch repository per case; output `run4.txt` (`e4737b20…c3db`) | **24 of 24 as expected:** F01 all-valid 0/full · F02 harness 1/full · F03 preflight 16/full (preflight evidence) · F04 malformed 14 · F05 scanner error 17/none · F06 retained canary 15/minimal · F07 excluded-console canary 15/minimal · F07b clean exclusion 0/full, console absent · F08 gitleaks pattern 15/minimal · F09 drift 13 · F10 cleanup 18 · **F11 staging after partial copy 19/minimal, nothing partial** · F11b unreadable raw 17/none · **F12 summary 19/none** · F13 publish 19/none · **F14 scrub 19, full already published** · F15 unexpected parent failure with the harness running 143, group stopped · **F16 external SIGTERM 143, group gone, harness cleanup ran, postcheck and cleanup ran** · F17 TERM-ignoring harness 143, KILL after grace · F18 watchdog deadline 124 · F19 stragglers 0, group stopped · **F20–F22 pre-location 16, repository inventory unchanged, nothing created, published none.** Every case ran F1 and F2, except the pre-location cases (F1 only, by design). | Corrections 1 and 2; groups C and D |
| P-MODEL defects found and fixed during r3 (disclosed) | First runs of the r3 model | **(i)** `! grep …` used as a statement never failed under `set -e`, so a failing preflight passed. Fixed with explicit returns; F27 adds the SC2251 lint. **(ii)** Signalling the harness group immediately after `setsid` launch missed it (F15 first run: group alive). Fixed with `group_ready`. **(iii)** The same race on the watchdog made each case wait 60 s. Fixed the same way. **(iv)** A stub-only bare `wait` blocked on its own process-substitution `tee`; the stub now waits on named PIDs. **(v)** The scrub left `state/`, so `rmdir` failed. The no-masking invariant correctly turned this into 19 before the fix. | Shows the invariant and the controls catch defects of exactly the kind Codex found |
| **P-OWN** | `own/own-model.sh` (`bfe86575…7d04`): the §11.4 decision rules over simulated inventories (no Docker objects); output `own-model.txt` (`d526f406…6234`) | **Removed (positively owned):** the owned ops container and root container; their captured networks and volumes; the images of owned containers; the build-registered `g003-vtools` image. **Refused or unproven (kept):** V20c foreign-`working_dir` container; V20n network and V20v volume under a registered label with no capture; V20i `beginify-intruder:latest` and `g001-ops-…-extra:latest` (registered-project name, no record). **Kept (pre-existing):** network, container and image in the start inventory. **Ignored:** unregistered project resources. | Correction 3; group E |
| P-MAIN | Scratch copy of `core/cmd/server/main.go` with the three r3 hunks; `gofmt` in `golang:1.27.1-alpine` (`--pull never --network none`); `patch` replay of the embedded hunks | gofmt-clean output, SHA-256 `987c933560ef77ea9a2daa4a3bec4c9747ea69d578f3c9c9d8e6179d013551ea`. Replaying the §7.1 hunks onto the current file reproduces that hash. Three hunks. Not compiled (no dependency download in planning). | §7.1; B-MCP-SOURCE pin |
| P-BASE | Read-only: `sha256sum -c references/shadcn-kit.sha256` in the skill; status counts; web-versus-asset identity | 63 of 63 OK, exit 0; 2 modified and 65 new skill paths; no other differing path besides PROGRESS and the Plan; 0 web files identical to an asset | §2.1; correction 5 |

**Correction of the r2 model (Codex finding).** r2's `model.sh` omitted `stage` from its aggregation list, and Codex reproduced `stage=19` with **exit 0**. The r3 model aggregates over the single `PHASES` array that the finalizer records into, and applies the structural no-masking invariant (§11.2.5). F11 now exits 19. r2's model and its outputs remain under `/tmp/g003-plan-r2/` as history; they are superseded, not reused.

### 15.6 Exact intermediate commands (r3; r2's commands with the r3 gate additions)

**Variables.** These are fixed at S0 and are not repository content:
- `EXT`: the execution evidence root outside the repository, created by S0 only after the containment check.
- `EXEC`: 6 lowercase hex.
- `BASE`, `SNAP`, `SNAP_SHA`, `COUNT`: Codex's Q9 binding.
- `PATH_RE`: the G003 expression in §15.4.
- `PLAN=management/v0/plans/G003-foundation-observability-ci-plan.md`.
- `ALLOW=management/v0/results/G003-scope-allowlist.txt`.

All commands run from the repository root.

```bash
# Wrapper invocation used by every intermediate and final suite run (scope bound to the exact allow-list).
wrap() { bash infra/validation/run-approved.sh --snapshot "$SNAP" --snapshot-sha256 "$SNAP_SHA" --count "$COUNT" \
  --path-re "$PATH_RE" --plan "$PLAN" -- --base "$BASE" --allow-file "$ALLOW" "$@"; }
lbl=(--label "g003.exec=$EXEC")

# S0 (read-only; any nonzero exit stops before a write)
bash "$EXT/g003-s0-gate.sh" "$PWD" "$BASE" "$SNAP" "$SNAP_SHA" "$COUNT" "$EXT/g003-exec/s0"
bash infra/validation/run-approved.sh --snapshot "$SNAP" --snapshot-sha256 "$SNAP_SHA" --count "$COUNT" \
  --path-re "$PATH_RE" --plan "$PLAN" --check-only
bash infra/validation/run-approved-controls.sh

# S0R (R1; no repository write; A-R1 and A-E required)
git clone -q --no-checkout "$PWD" "$EXT/r1/probe" && git -C "$EXT/r1/probe" checkout -q -b g003-runner-probe "$BASE"
#   write the four §12.4 probe files into $EXT/r1/probe, then:
git -C "$EXT/r1/probe" add .gitlab-ci.yml infra/ci/runner-preflight.sh infra/ci/runner-probe-job.sh infra/ci/probe.allow
git -C "$EXT/r1/probe" commit -q -m "G003 runner qualification probe (disposable; branch deleted after R1)"
git -C "$EXT/r1/probe" push git@gitlab.example:<namespace>/<project>.git g003-runner-probe
#   ... evidence retrieval per A-E ...
git -C "$EXT/r1/probe" push git@gitlab.example:<namespace>/<project>.git --delete g003-runner-probe

# S1 gate (after the S1 writes, the first of which is $ALLOW)
docker build -q "${lbl[@]}" -f infra/docker/intelligence.Dockerfile --target check --build-context contracts=contracts \
  -t "g003-$EXEC-intelligence-check" intelligence
docker run --rm --network none "${lbl[@]}" "g003-$EXEC-intelligence-check"

# S2 (the only network step: module resolution)
mkdir -p "$EXT/gomod" "$EXT/s2"
docker run --rm "${lbl[@]}" --user "$(id -u):$(id -g)" -e HOME=/tmp -e GOCACHE=/tmp/gocache -e GOPATH=/gopath \
  -e GOTOOLCHAIN=local -v "$EXT/gomod:/gopath" -v "$PWD/infra/validation/tools:/src" -w /src golang:1.27.1-alpine \
  sh -c 'go get github.com/santhosh-tekuri/jsonschema/v6@v6.0.3 && go mod tidy'
docker build -q "${lbl[@]}" -f infra/docker/validation-tools.Dockerfile --target check -t "g003-$EXEC-vtools-check" infra/validation/tools
docker run --rm --network none "${lbl[@]}" "g003-$EXEC-vtools-check"
docker build -q "${lbl[@]}" -f infra/docker/validation-tools.Dockerfile --target runtime -t "g003-$EXEC-vtools" infra/validation/tools
docker run --rm --network none "${lbl[@]}" -v "$PWD/contracts:/contracts:ro" -v "$EXT/s2:/out" "g003-$EXEC-vtools" \
  contractcheck corpus --schemas /contracts/intelligence --corpus /contracts/conformance/intelligence --out /out/results-go.json
docker run --rm --network none "${lbl[@]}" -v "$PWD/contracts:/contracts:ro" -v "$EXT/s2:/out" "g003-$EXEC-vtools" \
  contractcheck schema-negative --schemas /contracts/intelligence --corpus /contracts/conformance/intelligence --out /out/schema-negative-go.json

# S3 gate (CE1/CE2 environment files are written to $EXT/s3 with mode 600; canaries from /dev/urandom)
docker build -q "${lbl[@]}" -f infra/docker/core.Dockerfile --target check -t "g003-$EXEC-core-check" core
docker run --rm --network none "${lbl[@]}" "g003-$EXEC-core-check"
docker build -q "${lbl[@]}" -f infra/docker/core.Dockerfile --target runtime -t "g003-$EXEC-core" core
test "$(sha256sum core/cmd/server/main.go | cut -d' ' -f1)" = 987c933560ef77ea9a2daa4a3bec4c9747ea69d578f3c9c9d8e6179d013551ea
for c in ce1 ce2; do docker run --rm --network none "${lbl[@]}" --env-file "$EXT/s3/$c.env" "g003-$EXEC-core" >"$EXT/s3/$c.out" 2>&1; echo "$c rc=$?"; done
cid="$(docker run -d --network none "${lbl[@]}" --env-file "$EXT/s3/ce3.env" "g003-$EXEC-core")"
for i in $(seq 60); do docker logs "$cid" 2>&1 | grep -q '"event":"core listening"' && break; sleep 1; done
docker stop "$cid" >/dev/null; echo "ce3 rc=$(docker wait "$cid")"; docker logs "$cid" >"$EXT/s3/ce3.out" 2>&1; docker rm "$cid" >/dev/null
docker run --rm --network none "${lbl[@]}" -v "$EXT/s3:/in:ro" "g003-$EXEC-vtools" logcheck startup --service core \
  --case /in/ce1.out=1:core.exit:config_invalid --case /in/ce2.out=1:core.exit:listen_failed --exactly-one-line /in/ce2.out \
  --case "/in/ce3.out=0:core listening:port+app_env" --exactly-one-line /in/ce3.out --canaries /in/canaries.txt

# S4 gate
docker build -q "${lbl[@]}" -f infra/docker/web.Dockerfile --target check -t "g003-$EXEC-web-check" web
docker run --rm --network none "${lbl[@]}" "g003-$EXEC-web-check"
docker build -q "${lbl[@]}" -f infra/docker/web.Dockerfile --target runtime -t "g003-$EXEC-web" web
mkdir -p "$EXT/s4/copy" && git --no-optional-locks ls-files -z --cached --others --exclude-standard -- web \
  | tar --null -cf - -T - | tar -xf - -C "$EXT/s4/copy"
for r in wrapped control; do mkdir -p "$EXT/s4/copy/web/app/api/__probe/$r"
  cp "$EXT/s4/copy/web/tests/fixtures/error-probe/$r-route.ts" "$EXT/s4/copy/web/app/api/__probe/$r/route.ts"; done
docker build -q "${lbl[@]}" -f infra/docker/web.Dockerfile --target runtime -t "g003-$EXEC-web-probe" "$EXT/s4/copy/web"
#   WE0-WE6 incl. WE3a (regex-valid canary name) and WE3b: run the probe image (--network none, PROBE_CANARY from $EXT/s4/probe.env), httpprobe in its namespace with
#   $EXT/s4/we.json, docker logs -> $EXT/s4/web-probe.log, logcheck startup/scan; then docker rm -f the probe container.

# S5 gate
docker build -q "${lbl[@]}" -f infra/docker/intelligence.Dockerfile --target check --build-context contracts=contracts \
  -t "g003-$EXEC-intelligence-check" intelligence
docker run --rm --network none "${lbl[@]}" "g003-$EXEC-intelligence-check"
docker build -q "${lbl[@]}" -f infra/docker/intelligence.Dockerfile --target runtime --build-context contracts=contracts \
  -t "g003-$EXEC-intelligence" intelligence
docker run --rm --network none "${lbl[@]}" --env-file "$EXT/s5/ie1.env" "g003-$EXEC-intelligence" >"$EXT/s5/ie1.out" 2>&1; echo "ie1 rc=$?"
#   IE2: docker run -d (valid test env) -> wait for uvicorn.startup_complete in docker logs -> docker stop -> logcheck startup.

# S6 gates (G6a, G6b, G6c, in this order and in the same cycle)
docker run --rm "${lbl[@]}" -v "$PWD:/repo:ro" -w /repo koalaman/shellcheck:v0.11.0 -S warning -x -P SCRIPTDIR \
  infra/ci/validate.sh infra/ci/lib.sh infra/ci/validate-controls.sh infra/ci/cleanup-owned.sh infra/ci/runner-preflight.sh \
  infra/validation/run.sh infra/validation/lib.sh infra/validation/run-approved.sh infra/validation/run-approved-controls.sh \
  infra/local/configure-env.sh
docker run --rm "${lbl[@]}" -v "$PWD:/repo:ro" -w /repo koalaman/shellcheck:v0.11.0 -S info --include=SC2251 -x -P SCRIPTDIR \
  infra/ci/validate.sh infra/ci/lib.sh infra/ci/validate-controls.sh infra/ci/cleanup-owned.sh infra/ci/runner-preflight.sh
bash infra/ci/validate-controls.sh >"$EXT/s6/ci-controls.log" 2>&1; echo "controls_rc=$?"
wrap --report-dir "$EXT/s6/report" static provenance contracts; echo "overall_rc=$?"

# S7
wrap --report-dir "$EXT/s7/report" root migrations variants; echo "overall_rc=$?"

# S8 (local CI parity in a disposable scratch-commit copy; never pushed; deleted afterwards)
mkdir -p "$EXT/s8/clone" && git --no-optional-locks ls-files -z --cached --others --exclude-standard \
  | tar --null -cf - -T - | tar -xf - -C "$EXT/s8/clone"
git -C "$EXT/s8/clone" init -q && git -C "$EXT/s8/clone" add -A
git -C "$EXT/s8/clone" -c user.name=g003-parity -c user.email=parity@example.invalid commit -q -m "G003 local CI parity (scratch)"
(cd "$EXT/s8/clone" && env -u GITLAB_CI bash infra/ci/validate.sh --evidence-dir "$EXT/s8/evidence"); echo "validate_rc=$?"

# S10: §15.4.
```

- **S4 and S5 probe glue.** The comment lines are the steps that B-WEB-ERROR-OUTPUT and B-INTEL-ERROR-OUTPUT later perform inside `run.sh` (§10). At S4 and S5 they are executed with the same `httpprobe` and `logcheck` arguments, and the outputs are kept under `$EXT`.
- **Cleanup.** Every image and container in this block carries `g003.exec=$EXEC`. They are removed at the end with `docker ps -aq --filter label=g003.exec=$EXEC` and `docker image ls -q --filter label=g003.exec=$EXEC`, each checked against the S0 inventory.
- **Writes.** Nothing above writes inside the repository, except the S1–S6 implementation writes themselves and the module resolution of S2 (`go.mod` and `go.sum` under `infra/validation/tools/`, both allow-listed).

## 16. Test strategy and acceptance-criteria mapping (r3; r2 and r1 tables in the appendices)

| AC | Implementation | Commands and assertions (local unless marked hosted) |
|---|---|---|
| **AC1** Health semantics and bounds | §6, §7 | Unchanged from r2: unit tests; B-HEALTH-ROOT; B-WEB-HTTP; B-OBS-TIMEOUTS O1–O10 (each ≤ 5000 ms; recovery ≤ 90 s); B-OUTAGE; B-OBS-OFF; B-TOGGLE-OFF; `obs-timings.tsv`. WE1 keeps 500, empty body and no-store on an unexpected exception. **r3:** CE3 shows that a successful bind logs only the port and `app_env`, and that shutdown on TERM exits 0. |
| **AC2** Context validation, generation and isolation | §4 | Unchanged from r2 |
| **AC3** web → core → intelligence continuity | §4.3, §7 | Unchanged from r2 |
| **AC4** Structured logs and privacy | §5 | **r2 coverage, retained.** **r3:**<br>• the closed web `error_type` category set and the core startup-record keys enforced by `logcheck` and `events.json`;<br>• **CE1–CE3**, with CE2 exactly one line (`core.exit`/`listen_failed`, no `core listening`, no canary) and CE3 exactly one bounded `core listening`;<br>• **WE0–WE6 with WE3a** (a regex-valid canary-bearing `name` → `Error`, canary absent) and WE3b;<br>• IE1 and IE2;<br>• unit tests: Go `exit_test.go` (with `ListenPort`) and `errorlog_test.go`; Node category tests; Python `test_error_output.py`.<br>Source reading (P-NEXT, P-UVICORN) is **not** runtime evidence: WE, CE and IE are mandatory at the first approved runtime gate (S3–S5) and in every final run. **Artifacts:** F3 raw and F7 candidate scans; publication full, minimal or none (§11.2.3). |
| **AC5** Cross-language contracts | §8, §9 | Unchanged from r2 (accepted): B-CONTRACT-TOOLS, -GO, -PY, -AGREE, -PEER and B-CONTRACT-NEGATIVE (SN01–SN07 audit classes in both languages, `sn-derive`, positive audit controls, SD01/SD02, EI39) |
| **AC6** Runnable containerized commands | §10–§13, §15.6 | As in r2. Cleanup is shown by `inventory-leftover.txt` and `cleanup-inventory.tsv`. **r3:** the `main.go` hash check at S3; SC2251 in G6a and B-SHELL. |
| **AC7** GitLab gating with hosted evidence | §11, §12 | **Static:** B-CI-CONFIG (as in r2).<br>**Prerequisites (r3, correction 5):** Q1 facts F-R1–F-R13 and the Q2 named authorizations are **unresolved** and must be supplied before execution approval (§12.6). R1 runs the unchanged baseline cold under the supervisor's watchdog before S1.<br>**Attribution:** unchanged (accepted), with group B controls.<br>**Hosted H1:** a pass on the exact commit and tree; evidence from the **authoritative** job status and final log line, plus the scanned artifact. **Hosted H2:** a failure through the reviewed gate. |
| **AC8** Negative controls and failure propagation | §10, §11.3 | **B-CI-CONTROLS, 68 controls:**<br>• group A (a)–(f): (a) V2, V3; (b) V4 and B-CONTRACT-NEGATIVE; (d) V7, V12; (e) V8, V9; (f) V10, V11 and B-REQUIRED-CHECKS;<br>• group B: attribution;<br>• group C **(c)**: V5a, V5b-planted, V5b-clean, V5c, V6a, V6b, V6c, V6d;<br>• group D: finalization, supervision and location — V14–V18j, including staging after a partial copy (V16), summary (V16b), scrub (V17b), external SIGTERM (V18), TERM-ignoring harness (V18a), deadline (V18d), state import (V18e), unexpected parent failure (V18f), pre-location (V18g–V18i), stragglers (V18j);<br>• group E: positive ownership — V19, V20c, V20n, V20v, V20i, V21–V26.<br>**Retained:** W1–W9 and C1–C12, C8b and C8c.<br>**Model evidence:** P-MODEL 24 of 24, P-IMPORT 8 of 8, P-OWN (§15.5a). These are planning evidence only; the real controls run at S6 and S10. |
| **AC9** Scope, architecture and provenance | §14, §17 | As in r2: B-SCOPE; B-PROTECTED; B-SKILL-UNCHANGED and B-WEB-DESIGN against the BASE bound after Q14; B-RENDER-PLAN and B-RENDER-REPEAT (19); B-TOGGLE-OFF-ADDITIONS (35, including `main.go`); full regressions. **r3:** B-MCP-SOURCE pins `MAIN_GO_G003_SHA = 987c9335…51ea` (three hunks), with every security assertion retained. |
| **AC10** Documentation, Result and governance | §14, §19, §22 | As in r2. **Prerequisites Q1, Q2, Q14 and Q9 are resolved before execution** (§12.6). The Result separates local (S6–S10), runner-qualification (R1) and hosted (H0–H2) evidence; records the B7 cycles and the checkpoint after cycle 4; and records any `unproven` cleanup rows. |

## 17. Regression reconciliation and documentation labels (r3; r2 text in Appendix R2)

As in r2, with the `main.go` reconciliation updated:
- **`main.go` (Q16, Q21).** Exactly three reviewed hunks (§7.1; intended SHA-256 `987c9335…51ea`):
  - one customization-register row, so B-RENDER-PLAN and B-RENDER-REPEAT expect 19 conflicts and proposals;
  - one allow-list entry;
  - `APPLICABLE` + `core/cmd/server/main.go`, the same rendered base in the disabled variant;
  - B-MCP-SOURCE: byte equality to the rendered manifest becomes byte equality to `MAIN_GO_G003_SHA`. The routes are exactly two; the imports are the rendered set minus `fmt`; `go list -deps` excludes MCP, migration, fixture, isolation and `jsonschema` packages; there is no transport symbol; the reason code is unchanged.
- **The harness `tee` change and capture hooks** (§10) are project-owned harness edits. They are covered by B-SHELL, G6b group D and E controls, and the full static rerun in the same cycle.
- All other G001 and G002 guarantees, accepted deviations, residuals, FOLLOW-UP labels and out-of-scope items are as in r2.

## 18. Security and privacy negative controls (summary; r3)

As in r2, plus:
- **Error categories** are closed and code-controlled (web `instanceof`; core fixed enums). The core startup record carries only a numeric port and a validated `app_env`.
- **Publication** never contains partial or unscanned material (§11.2.3).
- **Cleanup** removes only positively owned resources. Unproven resources are kept and reported.
- **Containment** is checked before any evidence directory is created.
- **Signals** terminate the whole harness process group with a bounded grace. No harness process survives the run (V18, V18a, V18d, V18j).

## 19. Budget, feasibility and checkpoint (B7; r3; r2 text in Appendix R2)

**Budget (unchanged).** B7: at most **7** shared meaningful cycles; checkpoint **after cycle 4 and before cycle 5**. **Zero cycles consumed:** r1, r2 and r3 are planning, and planning models and probes are not cycles. No extension is requested.

**What changed since r2.**
- **Added implementation work:** parent-held state and the importer; the supervisor (`setsid`, watchdog, group stop); the publication-outcome logic; capture hooks and positive-ownership rules; the third `main.go` hunk and CE3; the web category table and WE3a; the one-line harness `tee` change; and 14 more controls.
- **Reduced risk:** each of these control-flow mechanisms has now been exercised by a running model with real signals (24 of 24), an importer unit test and an ownership decision model. The model run found and fixed five defects of the kind Codex flagged (§15.5a), which reduces the chance of discovering them during implementation cycles.

| Likely meaningful cycle | r2 | r3 |
|---|---|---|
| S1–S5 first unit and component failures (context, formatter, uvicorn takeover, Node, exit and error-log adapters, `instanceof` categories) | 1 | 1 |
| Cross-validator or schema-negative class mismatches | 0–1 | 0–1 |
| Error-output conformance on the real frameworks (CE1–CE3, WE0–WE6, IE1–IE2) | 0–1 | 0–1 |
| Integration observability cases | 1–2 | 1–2 |
| Harness and CI reconciliation (static, provenance, groups A–E, supervision, capture hooks) | 1 | 1 (prototyped by P-MODEL and P-OWN; the risk is in Docker-real group E and the capture hooks) |
| First hosted run of the G003 job (H0 or H1) | 0–1 | 0–1 |
| Read-only review findings | 1 | 1 |

**Forecast: 5–7 of 7, most likely 6.** The rows sum to **4–8**, as in r2; reaching 8 needs every uncertain row at its maximum at once. The forecast holds only if:
- **(a)** the Q1 facts confirm Profile A or A2;
- **(b)** A-R1 is authorized and R1 meets its decision rule before S1;
- **(c)** A-H0 is authorized, or the hosted surprise risk is accepted.

**Checkpoint after cycle 4** (BUDGET_POLICY §7). It records per-criterion status, convergence, repeated patterns and remaining work. If the remaining forecast exceeds the 3 remaining cycles, or the same material failure survives two corrective attempts, Claude stops and returns to Codex. Codex may then review a split or the budget.

**Split option (Q10; Codex decides).** Unchanged from r2: G003a (R1–R5, R8, local R6/R7, every control) and G003b (hosted AC7). I recommend the single Goal once Q1, Q2 (A-R1) and Q14 are resolved and R1 passes. If R1 cannot run before S1, the forecast is not credible, and I recommend Codex review the split before execution.

## 20. Delegation

- **One implementation owner:** me. There are no implementation subagents: the work shares the harness, the log contract and the three runtimes.
- **Two read-only reviewers at S10**, run in parallel. Neither edits files. Each reports findings with severity and `file:line`, and the cycles they used (expected 0).
  - **SA-P (correlation and privacy):**
    - the §4 profile against W3C and the Spec;
    - isolation, cancellation and panic paths;
    - completion "exactly one" semantics;
    - redaction parity;
    - framework output;
    - canary coverage and non-vacuity.
  - **SA-C (CI, contracts and provenance):**
    - validator independence and the corpus/expectation design;
    - the Python corrections;
    - `validate.sh` exit semantics and V1–V12;
    - artifact allow-list and scan;
    - B-CI-CONFIG;
    - the register, `APPLICABLE` and the allow-list;
    - the G001/G002 regression coverage.
- **Integration and budget.** I triage and integrate every finding. Corrective work counts against the shared B7. Reviewer opinions do not replace Codex review.

**r2 additions to §20.**
- **Timing.** In the r2 sequence the two reviewers run at **S9**; r1's "S10" above refers to r1's numbering (§15.3 step map).
- **SA-P** also reviews the error-output handling:
  - the bounded events and categories (§5.6);
  - WE0–WE6, CE1–CE2 and IE1–IE2, including the WE4 positive control;
  - the uvicorn, library and warning mapping.
- **SA-C** also reviews:
  - attribution against GitLab's documented variables (§11.2.4, group B);
  - the finalizer and exit-code aggregation (§11.2.3–§11.2.5, groups C and D);
  - the registry and cleanup binding (§11.4, group E);
  - schema-negative isolation (§8.2a–§8.3);
  - the `main.go` reconciliation (§17).
- **Still no implementation subagent.** The r2 revision itself was prepared by a planning subagent (§0.2), which made no implementation change and consumed no cycle.

**r3 additions to §20.**
- **SA-P** also reviews:
  - the closed web category table and WE3a;
  - CE1–CE3 and the bind-then-log hunk;
  - that every startup record is catalogued and none is exempted.
- **SA-C** also reviews:
  - the parent-held state and importer;
  - the supervisor (`setsid`, `group_ready`, watchdog, group stop, signal paths);
  - the publication outcomes and aggregation invariant;
  - the capture hooks and positive-ownership rules;
  - the `tee` change;
  - group D and E controls against P-MODEL and P-OWN.
- **No implementation subagent** is planned. The r3 revision was prepared by a planning agent at the owner's direction (§0.5), with no implementation change and no cycle consumed.

## 21. Risks and escalation (r3; r2 and r1 tables in the appendices)

| # | Risk | Handling |
|---|---|---|
| R1 | Runner profile incompatible, unavailable or privileged | **Blocker (Q1).** Facts first (§12.6), then R1 before S1. Never use dind or privileged mode; never use another CI vendor. |
| R2 | uvicorn, a library or `warnings` emits a raw line | `logcheck` fails visibly (IE2, B-OBS-LOGS). Fix inside the application's logging configuration. A `CMD` change is returned to Codex. |
| R3–R12 | As in r2 (Node type stripping, `r.Pattern`, validator disagreement, paused-target counts, cold time, artifact size, gitleaks false positive, harness weakening, budget, client settings file) | As in r2 |
| R13 | Merged-result parent structure differs | Fail closed with 16; a planning correction, as in r2 |
| R14 | Real Next 16.3.8 behaviour differs from the source reading (WE0, WE5, WE6) | Stop and return to Codex. The allow-list is never widened silently. |
| R15 | `SystemExit(3)` under uvicorn misbehaves (IE1) | Return to Codex |
| R16 | The runner ignores the `RUNNER_*` timeout variables | The supervisor's watchdog remains the gate (§11.2.1) |
| R17 | The consolidation is not committed, or makes B-WEB-DESIGN fail | Execution stays blocked (Q14). 0 identical files at r3. |
| R18 | Pushes publish unpushed ancestors | Covered by the authorization text (§12.3) |
| **R19** | The process group does not exist yet when it is signalled (launch race, F25) | `group_ready` before any group signal, with a PID fallback (§11.2.1); model F15 |
| **R20** | The harness is killed before a capture point, so a resource is `unproven` | Kept and reported (18), never deleted by inference. The run is already failing. The owner cleans up from the report (§11.4). |
| **R21** | A `!`-negated statement silently passes under `set -e` (F27) | Explicit returns; SC2251 in B-SHELL and G6a |
| **R22** | Under group TERM, the harness's own cleanup still fails in some path not covered by P-TEE | `validate.sh` F2 is the guaranteed path; group D controls V18 and V18d assert harness cleanup with the r3 `tee`. A failure is a corrective cycle, never a relaxed assertion. |
| **R23** | `net.Listen` before `Serve` changes a startup behaviour that some check relies on | CE2, CE3, B-OUTAGE, B-PORTS-*, readiness and the browser suites re-run at S7 and S10. Any regression is an escalation, not an accepted change. |

**Stop and escalate on:**
- any `main.go` change beyond the three reviewed hunks (`987c9335…51ea`);
- any `local.yml`, deadline, schema-byte or Dockerfile `CMD` change;
- any architecture, security-model, vendor, schema, route or API/MCP change;
- weakening any check, assertion or control expectation;
- touching developer resources;
- any unresolved Q1, Q2, Q9 or Q14;
- missing hosted authorization;
- B7 exhaustion or thrashing.

## 22. Questions for Codex and the owner, and unmet prerequisites (r3; r2 and r1 tables in the appendices)

| # | Status | Question or prerequisite | What is needed, and the recommendation |
|---|---|---|---|
| **Q1** | **UNRESOLVED — blocking** | Runner facts F-R1–F-R13 (§12.2) | **The owner supplies them before execution approval is sought** (§12.6): named fields only, never the `config.toml` token. Codex checks Profile A or A2; R1 then verifies. Nothing about the runner is assumed. |
| **Q2** | **UNRESOLVED — blocking** | Named hosted-action authorizations A-R1, A-H0 (optional), A-H1, A-H2, A-H3 and A-E (§12.3) | Explicit, per action, including commits and the publication of unpushed ancestors. Not granted by this Plan, its review or the planning delegation. |
| **Q14** | **OPEN — owner decision; blocks Q9** | Baseline route. The consolidation is **finished** (63 of 63 manifest entries verify) but **uncommitted** (2 modified and 65 new files). | **Route A (recommended):** the owner commits it, and Codex binds BASE to that commit. Routes B and C (r2) still need a reviewed Plan revision and are not planned. |
| **Q9** | **PENDING — blocking** | The fresh binding | Codex, after technical approval of this Plan and the Q14 resolution: BASE, snapshot, its SHA-256, COUNT. `/tmp/g003-r1-…` and `/tmp/g003-r2-reviewed-inputs.sha256` are continuity only, never bindings. |
| Q3–Q8, Q11–Q13 | Resolved by Codex (r1) | — | As recorded (§0.1) |
| Q10 | **Open (Codex)** | Single Goal versus split | Single Goal once Q1, A-R1 and Q14 are resolved and R1 passes (§19) |
| Q15 | Direction accepted (r2 review), corrected in r3 | Web's own empty-body 500 | r3 maps `error_type` through closed `instanceof` categories (§5.1–§5.2 r3 additions) |
| Q16 | Direction accepted (r2 review), corrected in r3 | Bounded `main.go` customization | Now three hunks (§7.1), with the reconciliation in §17 |
| Q17, Q18, Q19 | Directions accepted (r2 review) | Fail-closed profiles; R1 before S1; raw and staged scanning in `validate.sh` | Retained; R1 now uses the supervisor's watchdog (§12.4) |
| **Q20** | New, for review | The project-owned harness changes in r3: the one-line TERM/INT/HUP-ignoring `tee`, the `owned_capture` hooks and the `image <tag> <id>` build rows (§10, §11.4) | Approve. Needed for reliable harness cleanup under signals (P-TEE) and positive ownership. Normal runs are unaffected. |
| **Q21** | New, for review | `main.go` hunk 3 binds with `net.Listen` before `Serve`, and logs only the bound port and `app_env` (§7.1) | Approve. It removes the raw-address startup leak and makes CE2 exactly one line; regressions are covered (R23). |

## 23. Readiness (r3)

**r3: ready for Codex plan review. Execution is blocked by unresolved owner prerequisites.**

- **Blocking (§12.6):**
  - **Q1:** runner facts, unresolved;
  - **Q2:** named hosted-action authorizations, unresolved;
  - **Q14:** baseline route, open (consolidation complete, uncommitted);
  - **Q9:** fresh binding, pending technical approval and Q14.
- **Claude will not seek execution approval** until Q1 and Q2 are supplied.
- **Verified while writing r3:**
  - HEAD `9461c4b`; Goal, Spec and ARCHITECTURE unchanged;
  - S0 re-extracts to `0409148a…6723`;
  - Appendix R1 re-extracts to `a43aa523…dbfc`;
  - Appendix R2, joined with the Appendix R1 section, reproduces r2 `6e6356f3…bc82`;
  - the §7.1 hunks replay to `987c9335…51ea`;
  - P-MODEL 24 of 24, P-IMPORT 8 of 8, P-OWN as expected;
  - the only repository write is this Plan.
- **Not done:** no implementation, installation, initialization, application build, commit, push, pipeline trigger, hosted action or external configuration change.
- **Next step:** Codex review of r3. No implementation may begin.

~~~~~~~~~~~~

## Appendix R2 — r2 Plan body, verbatim (superseded wherever r3 differs; not an execution instruction)

This block preserves the reviewed r2 text, as review history; this Plan file is untracked.
- **Contents.** It holds r2 lines 1–1933: everything before r2's own "Appendix R1" section. That section is **byte-identical** to the "Appendix R1" section that ends this file.
- **Verification.** The r2 file is exactly this block's lines followed by that final section:

```bash
P=management/v0/plans/G003-foundation-observability-ci-plan.md
{ awk '$0=="~~~~~~~~~~ r2-body"{f=1; next} $0=="~~~~~~~~~~"{f=0} f' "$P"; sed -n '/^## Appendix R1 — /,$p' "$P"; } | sha256sum
# expected: 6e6356f3bf7e6b0155fe4a32764e1e89d5cf25afa7a96dd0642096fd0220bc82
```

- **Precedence.** Where r2 and the body above differ, **the body above governs**. The S0 extraction source for execution is §15.1 of the body above.

~~~~~~~~~~ r2-body
# G003 — Health, logging, contracts and baseline CI implementation plan

Goal: [G003-foundation-observability-ci.md](../goals/G003-foundation-observability-ci.md)
Spec: [G003-foundation-observability-ci-spec.md](../specs/G003-foundation-observability-ci-spec.md)
Owner: Claude Code · Reviewer: Codex · Budget class: **B7**: at most 7 shared meaningful cycles; mandatory checkpoint after cycle 4 and before cycle 5. Planning consumed no cycle.

| Revision | Date | Change |
|---|---|---|
| r1 | 2026-10-05 | Initial Plan. Submitted for Codex review. **Not approved; no implementation has begun.** One unmet prerequisite blocks execution (the GitLab runner profile, §12 and Q1); the remaining questions are in §22. — *Appended at r2:* r1 SHA-256 `a43aa523624bab98c38e0fe6c7ff86cffef2cd70d7a3da7b6e806d5c60adbdfc` (1333 lines). Codex review: **CORRECTION LOOP**, six corrections (§0.1). Codex recorded that Q2 and Q9 were also prerequisites, so the "one unmet prerequisite" statement above was incomplete; r2 corrects it. The complete r1 text is preserved verbatim in Appendix R1. |
| r2 | 2026-10-05 | Addresses the six r1 corrections (map in §0.1). **(1)** Runner facts, hosted authorizations and the fresh binding are recorded as unresolved blocking prerequisites (Q1, Q2, Q9), with an exact owner fact list (§12.2), named authorizations (§12.3) and a runner-qualification step R1 that runs the unchanged baseline harness cold on the runner **before S1** (§12.4). The working tree is in an owner-authorized skill consolidation; baseline verification is **PENDING** and the S0 gate is unchanged, with no new exception (§2.1, §15.1a, Q14). **(2)** Checkout attribution uses only documented variables per pipeline type, with 5 positive and 16 negative metadata controls that run before, and independently of, the harness (§11.2, §11.3). **(3)** An explicit phase and finalizer model for `validate.sh`: postcheck, owned cleanup, raw and staged scans and publication run on every exit path; the harness status always survives; scanner and report errors fail closed; a job-unique registry, discoverable by `after_script`, binds owned resources to label-based removal commands (§11.1–§11.4). **(4)** Schema-negative cases run through a separate audit path with declared failure classes and single-mutation derivations; normal runs still reject any unpinned bytes (§8.1a–§8.3). **(5)** Error output is established from the actual frameworks: Next.js 16.3.8 logs a thrown route error with message and stack, so the web wrapper now answers 500 itself; `core/cmd/server/main.go` gets a bounded JSON exit record and a classifying `net/http` error log; intelligence gains bounded startup-failure, uvicorn, library and warning records; three real-artifact checks assert this (§5.3–§5.6, §7, §10). **(6)** An executable step order: the scope allow-list is the first write (S1); CI files, documentation and harness changes are integrated in one step before their first static run (S6); exact intermediate commands are given (§15.3, §15.6). B7 is unchanged. Submitted for Codex review. **Not approved; no implementation has begun. Execution is blocked by Q1, Q2, Q9 and Q14 (§22).** |

## 0. Review record and planning-session disclosures

### 0.1 r1 → Codex: CORRECTION LOOP (2026-10-05)

Codex reviewed r1 (SHA-256 `a43aa523…dbfc`) at HEAD `9461c4bdbd8c1ed3e9a9986c70647cd1add57101`. Codex confirmed that the embedded S0 script extracts byte-exactly to `0409148a…6723` and passes `bash -n`. That was script-text verification, not execution approval. The review is recorded in PROGRESS.md under `### 2026-10-05 — G003 Plan r1 review: CORRECTION LOOP`. **The review text governs.** The left column below abbreviates it.

| # | Codex correction (abbreviated) | Where r2 addresses it |
|---|---|---|
| 1 | **Runner, hosted authorization and current starting state** (R6/R7, AC7/AC10). Q1, Q2 and Q9 remain prerequisites. Record the runner's identity and tag, executor, isolation and permissions, host, path and network profile, capacity and timeouts, and artifact access. Obtain explicit authorization for the named hosted success and failure actions and for evidence retrieval. Check runner compatibility and cold-build feasibility early enough for the B7 forecast, not first at S12. Reassess or isolate the execution baseline (untracked `vestive-shadcn/**`) without deleting that work or adding a broad unrelated-path exception. A fresh supervisor binding follows approval and a resolved baseline. | §2.1 (current state; baseline **PENDING**), §12.2 (owner fact list, all **UNRESOLVED**), §12.3 (named authorizations A-R1 to A-E), §12.4 (R1 runner qualification before S1, with a decision rule), §15.1a (S0 unchanged; no new exception), §15.3 (S0R), §19 (forecast conditional on R1), §22 (Q1, Q2, Q9, Q14) |
| 2 | **GitLab checkout attribution** (R6/R7, AC7). `CI_MERGE_REQUEST_SOURCE_BRANCH_SHA` is empty in detached merge-request pipelines and set only in merged-results pipelines. Use the variables actually present for each pipeline type. Distinguish detached source, merged result and branch push, without inventing a merged-result claim. Add positive controls for each metadata profile and negative controls for missing or contradictory metadata, producing the intended preflight outcome independently of harness failures. | §3 (F13), §11.2 (attribution rules, cited), §11.3 group B (P1–P5, P3F, N1–N16), §15.5 (P-ATTR prototype: 21 of 21 as expected) |
| 3 | **Aggregate finalization, scanning and owned cleanup** (R7, AC7/AC8). Specify an explicit failure and finalization model: postcheck and safe cleanup reporting run on ordinary failure paths; a nonzero harness status survives; scanner and tool errors and missing or malformed reports fail closed. Resolve the raw-versus-staged scan mismatch. The cleanup registry must be job-unique, discoverable by `after_script`, and bound to the configuration needed for removal (`docker compose -p <name> down` has neither the Compose file nor the environment). Refuse unowned resources and test it. Keep cleanup failures observable. `after_script` runs in a separate shell, cannot fail a successful job and does not run on job timeout by default. | §11.1 (job timeouts and `after_script`), §11.2 (phases, finalizer, exit codes, raw and staged scans), §11.4 (registry and `cleanup-owned.sh`), §11.3 groups C–E, §10 (proposed B-ARTIFACT-SCAN withdrawn; registry hooks), §15.5 (P-FIN, P-TERM and P-SCAN probes) |
| 4 | **Schema-negative isolation** (R5/R7, AC5/AC8). Normal runs must reject bytes that differ from the two pinned hashes. SN02–SN07 must reach the intended validator checks through a disposable negative path, so a shared digest failure cannot make them pass vacuously. Each language asserts the intended classification. Keep an ordinary-run digest-mismatch control and the EI39 control. No network access and no production schema mutation. | §8.1a, §8.2a (declared classes and single-mutation derivations), §8.3 (stage order; normal mode versus audit mode), §10 (B-CONTRACT-NEGATIVE) |
| 5 | **Actual error-output conformance** (R4, AC4). The web wrapper rethrew, and r1 assumed Next prints only startup text. Prove the handling with a real-framework error probe that uses message and stack canaries. Core's protected `cmd/server/main.go` prints `run()` errors with `fmt.Fprintln(os.Stderr, err)`; resolve it, and if `main.go` changes, reconcile the register, scope list, disabled variant and equality guard under R8. Error events need bounded categories or types; arbitrary messages under a renamed `event` key are not enough. | §3 (F15–F19), §5.3, §5.4, §5.6, §7.1–§7.3, §10 (B-CORE-, B-WEB- and B-INTEL-ERROR-OUTPUT), §15.5 (P-GO probe and P-NEXT source evidence), §17 (`main.go` reconciliation), Q15, Q16 |
| 6 | **Executable step order** (R6/R8, AC6/AC9/AC10). S6 ran static checks that read CI files created at S7, and the scope allow-list was created only at S8, although earlier wrapper runs needed it. Reorder so every required file exists when it is first checked; name the exact intermediate commands and scope binding; keep the rule that a shared harness change reruns the affected static and control checks in the same cycle. | §15.3 (new sequence, with an r1→r2 step map), §15.6 (exact commands), §14.3 (the allow-list is created at S1) |

**Planning decisions Codex recorded at r1 (not re-opened by r2):**
- **Q3:** the bounded Python format and anchor corrections enforce the unchanged canonical schemas; they authorize no new schema restriction.
- **Q4:** keep the nested 2 s deadlines, prove optional intelligence is non-gating at core, and disclose the web timeout result.
- **Q5:** scoped to the credentials each service actually holds; web gains no credential access.
- **Q6, Q11, Q12:** the fixture location, the separate pinned test-tool module and the disposable local CI-parity checkout are within scope; no new vendor ADR is needed for the validation library.
- **Q7:** developer-only artifacts with a 90-day expiry are an acceptable intended policy, subject to verifying actual instance and project access and limits (§12.2, F-R11).
- **Q8:** failure of the required job is G003's gate; owner-managed merge and protection settings are not changed.
- **Q13:** the single exact ignored client-settings exception is acceptable, without reading, deleting or broadening it.
- **Budget:** B7 retained as the Goal's ceiling, at most seven shared meaningful cycles, checkpoint after cycle 4; zero cycles consumed. The six-cycle forecast stays conditional on runner compatibility, cold-run capacity and hosted authorization.

### 0.2 r2 planning-session disclosures

- **Delegation.** This revision was prepared by a planning subagent at the product owner's direction, in parallel with a separate owner-authorized skill consolidation run by another agent. That delegation authorizes Plan revision only. It is **not** implementation authorization and **not** hosted-action authorization.
- **Repository writes.** The only repository write is this file.
  - It was revised in place in this checkout, not in a worktree.
  - To keep the r1 text byte-exact (Appendix R1), the r2 file was assembled from a hash-verified scratch copy of r1 and new sections under `/tmp/g003-plan-r2/build/`. It was installed onto this path with one scripted copy, after re-checking that the path still held the r1 bytes. This is the scripted in-place edit the owner authorized.
- **Git.** Only read-only commands were used (`status`, `log`, `rev-parse`, `ls-files`, `diff`), always with `GIT_OPTIONAL_LOCKS=0`. No index, ref, stash, commit or push operation was performed.
- **Not done.** No install, initialization, application build, `docker build`, `docker compose up/build/pull`, harness run, GitLab API call, pipeline trigger or external configuration change.
- **What was run** (scratch only, under `/tmp/g003-plan-r2/`; full list in §15.5):
  - a stdlib-only Go probe (`go run`) inside the existing local `golang:1.27.1-alpine` image (`--pull never --network none`);
  - read-only `grep` of that image's Go 1.27.1 `net/http` source;
  - Bash probes of `errexit`, traps, `timeout` and finalization;
  - Git fixture probes of shallow merge-commit parents and the attribution classifier;
  - a gitleaks v8.30.1 exit-code probe on scratch directories (`--pull never --network none`);
  - `docker image ls` and `docker ps -a` inventory listings.

  The probes created only short-lived `--rm` containers, from images already present. They created no volumes, networks or images.
- **Documentation research.** Official GitLab documentation (predefined variables, merge-request and merged-results pipelines, `CI_PIPELINE_SOURCE` values, `after_script`, runner timeouts). Next.js 16.3.8 and uvicorn 0.54.0 source files at their release tags on GitHub.
- **The other agent's skill files were not inspected as inputs.** They were listed by path, and their bytes were hashed only for the B-WEB-DESIGN identity check (§2.1). Their content was not inspected, modified or bound.
- **No credential, token or environment secret was read or printed.**

### 0.3 r1 planning-session disclosures (verbatim from r1 §0, as submitted)

No Codex review has taken place yet.

Disclosures from this planning session:

- **Writes.** The only repository write is this file.
  - Scratch writes were made only outside the repository, under the job's temporary directory: the S0 gate script, its demonstration clones and their evidence (§15.1–§15.2), and a staging copy of this Plan.
  - The background session's write guard required this Plan to be staged outside the repository first. It is placed at its canonical path only as the single repository write.
- **One Git call with optional locks enabled.** The first baseline check of the session ran an ordinary `git status --porcelain=v1 -uall`. It printed nothing, but because optional locks were enabled it may have refreshed `.git/index` stat data. It cannot have changed content. Every later Git inspection used `git --no-optional-locks` with `GIT_OPTIONAL_LOCKS=0`.
- **Tool-local ignored file.** During the session, at 17:22 local time, the Claude Code client created `.claude/settings.local.json`. It is ignored by the user's global excludes (`~/.config/git/ignore`), and was not created, read or modified by me. It is not a repository input. §15.1 treats it as the single exact ignored-path exception.
- **What I did not do.** I did not install anything, initialize any tool, build any application image, commit, push, trigger a pipeline or change any external configuration.
- **What I did run.** Docker was used only for read-only tooling on scratch files and inventories:
  - `mikefarah/yq` reading the committed manifest;
  - `koalaman/shellcheck` on the scratch S0 gate;
  - inventory and port queries.
- **Remote queries (read-only).**
  - `git ls-remote --symref origin` over SSH, with batch mode on.
  - One unauthenticated GitLab API probe. It returned 404, because the project is private.
- **No credential, token or environment secret was read or printed.**

## 1. Summary (r2)

G003 completes the existing operational paths. It does not rebuild them. The r1 summary is preserved in Appendix R1; the points below are current.

- **Diagnostic context (R2/R3)** is unchanged from r1 (§4).
  - Web, core and intelligence validate `X-Request-Id` and a W3C `traceparent` version-00 profile on the six health routes.
  - Each service creates its own nonzero span. Context is propagated on exactly two existing calls: web `/api/ready` → core `/readyz`, and core `/readyz` → intelligence `/healthz`.
  - The parent chain is client parent → web span → core span → intelligence span.
- **Logs and privacy (R4).**
  - Each service writes exactly one JSON completion record per scoped request (§5.2).
  - **New in r2:** every application-owned `event` value is a code-authored constant. Framework, library and error text is never copied into a record. Errors carry only a bounded `error_type` category or exception class (§5.6).
  - **New in r2 (correction 5): actual error output.**
    - **Web.** Next.js 16.3.8 prints a thrown route-handler error with its message and stack, then answers 500 with an empty body. The web wrapper therefore no longer rethrows: it logs the bounded record and returns that same 500 itself, plus the two additive headers (§5.3, §7.2).
    - **Core.** `core/cmd/server/main.go` gets two bounded hunks: a JSON `core.exit` record replaces `fmt.Fprintln(os.Stderr, err)`, and a classifying `net/http` `ErrorLog` is installed (§7.1, Q16).
    - **Intelligence.** It emits a bounded `intelligence.startup.failed` record instead of an uncaught traceback, and it maps uvicorn, third-party library and Python-warning output to bounded events (§7.3).
    - **Checks.** Three real-artifact checks assert the actual framework output: B-CORE-ERROR-OUTPUT, B-WEB-ERROR-OUTPUT and B-INTEL-ERROR-OUTPUT (§10).
- **Health (R1)** is unchanged from r1. Status codes, bodies and the 2 s deadlines stay the same. The additive headers are `X-Request-Id` on all six routes and `Cache-Control: no-store` on intelligence health. Each single request is bounded at 5 s or less.
- **Contracts (R5).**
  - Unchanged from r1: one synthetic corpus, an independent Go validator (santhosh v6.0.3 in a test-tool module), and the corrected Python validator.
  - **New in r2 (correction 4):** normal runs reject any schema whose bytes differ from the pinned digests, before any other check. The schema-negative cases SN01–SN07 run through a separate **audit mode**, with no digest step and no case evaluation. Each case must report its declared failure class in both languages. Each SN fixture is proven to be the canonical schema plus exactly one declared mutation (§8.1a–§8.3).
- **CI (R6/R7).** A root `.gitlab-ci.yml` defines one gating job that runs `infra/ci/validate.sh`. New in r2:
  - **Attribution (correction 2).** The tested content is classified as `mr-source`, `mr-merged-result`, `branch-push` or `local`, using only the documented variables present for each pipeline type. Missing or contradictory metadata fails the preflight with 16 before the harness runs (§11.2).
  - **Finalization (correction 3).** An EXIT-trap finalizer runs on every exit path: postcheck, owned cleanup, a raw-capture scan, a completeness check, staging, a staged scan and publication. Exit statuses are aggregated so that no failure can be masked (§11.2).
  - **Cleanup (correction 3).** A job-unique registry outside the checkout, discoverable by `after_script`, binds owned resources to label-based removal commands that need no Compose file. Unowned resources are refused. A failed removal fails the run (§11.4).
  - **Controls.** Groups A–E (§11.3) extend r1's V1–V12. W1–W9 and C1–C12 are retained.
- **Runner and hosted evidence (correction 1).**
  - **Blocked.** The runner facts (Q1) and the hosted-action authorizations (Q2) are owner prerequisites that are still unresolved. The fresh binding (Q9) is pending.
  - **R1 qualification.** After S0, and before any implementation write, a disposable probe branch runs a runner preflight and the **unchanged baseline harness, cold**, on the actual runner. Compatibility and cold duration are therefore known before the first implementation cycle (§12.4).
- **Baseline (correction 1).** The working tree contains an in-progress, owner-authorized consolidation of `skills/vestive-design-skill/**`. **Baseline verification is PENDING** until it finishes and Codex supplies a fresh binding.
  - The S0 gate is unchanged. It still rejects the extra paths, and no exception is added.
  - The recommended route is to commit the consolidation and bind BASE to that commit (Q14).
- **Step order (correction 6).**
  - The scope allow-list is the first repository write (S1).
  - CI files, documentation and harness changes are integrated together in S6, before their first static run.
  - Exact commands for every gate are in §15.6.
- **Not changed.**
  - No database, schema, migration, public domain API or MCP change.
  - No new route in the production tree. The error-probe routes exist only in disposable copies.
  - No production execution client, telemetry vendor, metrics endpoint, deployment or UI change.
  - `infra/compose/local.yml`, `core/go.mod`, the intelligence lockfile and web dependencies are unchanged.
  - The foundation barrier stays closed through G021.
- **Correction to r1 (disclosed).** `core/cmd/server/main.go` is **no longer** unchanged (r1 §1, §2 and §7.1 said it was). It is a planned, reviewed customization (§7.1, §14.2, §17, Q16).

## 2. Verified baseline (2026-10-05, read-only)

| Item | Value |
|---|---|
| HEAD | `9461c4bdbd8c1ed3e9a9986c70647cd1add57101` ("Define G003 operational foundation and CI specification"), branch `staging` |
| Working tree | Clean at the start of the session: `git --no-optional-locks status --porcelain=v1 -z --untracked-files=all --ignored` was empty. Since 17:22 local time it shows one ignored, tool-local path, `!! .claude/settings.local.json` (§0), plus this Plan once placed. |
| Goal SHA-256 | `63f2b26676c9f1c385639c2078abc586beb48e36d6722c957aa4b03b70d73bb2` (matches handoff) |
| Spec SHA-256 | `d64a691d49219e6945b3fb115bcea59977a9b491f7e2746702e329d1c1c2e949` (matches handoff) |
| ARCHITECTURE.md SHA-256 | `6a6fd49b854ce90acf1a2ce6bd281a4deee3e68e8f4e021d8d737c3b04d90726` (matches handoff; §60 and §61 read) |
| Remote | `origin` `git@gitlab.example:<namespace>/<project>.git`. `ls-remote --symref`: `HEAD → refs/heads/main` (`99e2c79`, "Initial planning"), `refs/heads/staging` = `9461c4b` (equal to local HEAD). The default branch is `main`. |
| GitLab project | Private: the unauthenticated API returns 404. No `glab`, `gitlab-runner` or API token is available in this session, so runners, settings, pipelines and artifacts cannot be inspected here (§12). |
| Governance | SECURITY.md and DECISIONS.md are empty; there is no accepted ADR. BUDGET_POLICY, ROADMAP and PROGRESS were read. G001/G002 acceptance is recorded in PROGRESS and supersedes historical submission wording. |
| Schema digests (must remain unchanged) | `execution-envelope.schema.json` `bf3cec365e291da2cf9b201b9d0194632818721671a5844589954b978accba4f`; `stream-event.schema.json` `e1025031134253ac8d19a514f6d7a4d9563281b5006ebe9be40a84e4d63fd102` |
| Protected harness files | `run-approved.sh` `620a3e84…c53c695` and `run-approved-controls.sh` `314a16a7…7ba9` (both unchanged by G003); `run.sh` `d5e7e5be…41ac8`, `lib.sh` `9f7a3997…57c3` and `render-customizations.tsv` `9abf9e92…279f` (all three modified, §10) |
| `core/cmd/server/main.go` | `95aa5b23…7951`. It equals the rendered manifest and stays **unchanged**, so B-MCP-SOURCE's byte-equality guard needs no reconciliation. |
| Local Docker | Engine 27.5.1, Compose 2.32.4 (Docker Desktop on WSL2); 16 CPUs, 15 GiB RAM, 920 GB free. No `beginify` project containers, volumes or networks. Default ports 3100, 8180, 8100, 5532, 6479, 9100 and 9101 are free, as are the variant port sets. Unrelated developer resources exist (for example `payments-assistant-postgres-1`, `outreacher-*`, `needknow-drafter-*` and `atlas-*`/`harbor-*` images) and are never touched. |

### 2.1 r2 re-verification and current working-tree state (baseline verification **PENDING**)

Read-only, 2026-10-05 at about 22:47 UTC (18:47 EDT), with `GIT_OPTIONAL_LOCKS=0`. The table above is r1's record, kept unchanged. The rows below supersede it where they differ.

| Item | r2 observation |
|---|---|
| HEAD | Unchanged: `9461c4bdbd8c1ed3e9a9986c70647cd1add57101`. `staging` was equal to `origin/staging` at r1 (`ls-remote`); r2 made no remote query. |
| Goal, Spec, ARCHITECTURE SHA-256 | Unchanged: `63f2b266…3bb2`, `d64a691d…e949`, `6a6fd49b…0726` |
| PROGRESS.md SHA-256 | `a75c573f11bdd4fc564123d06b3c2b9017ba17bf7c4722f7430692646ca18320`. It is modified (` M`) by Codex's r1 review record, a supervisor-owned change. |
| Plan | r1 `a43aa523…dbfc` before this edit. The r2 hash is reported at handoff, not self-bound. |
| AGENTS.md, CLAUDE.md, BUDGET_POLICY.md, ROADMAP.md | Read. `3bc3e81f…75e9`, `ee2b13ff…0e6c`, `9abced55…aa3b`, `bd41e556…0fcc` |
| SECURITY.md, DECISIONS.md | Both still 0 bytes. There is no accepted ADR. |
| **Other working-tree changes (owner-authorized skill consolidation, in progress)** | At observation time:<br>• ` M skills/vestive-design-skill/README.md` and ` M skills/vestive-design-skill/SKILL.md`;<br>• 63 untracked paths under `skills/vestive-design-skill/assets/shadcn/`;<br>• 2 untracked paths under `skills/vestive-design-skill/references/` (`shadcn-kit.md`, `shadcn-kit.sha256`);<br>• the untracked `vestive-shadcn/**` that Codex saw at review is **gone**: the other agent removed it after preservation, as planned.<br>The state is still changing. These paths were listed only, and hashed for one check (next row). They were not read, modified or bound. |
| B-WEB-DESIGN interaction | G002's B-WEB-DESIGN fails if a web file is byte-identical to any file under `skills/vestive-design-skill/assets/`. At observation time **0** web files matched any of the 63 asset files. This is rechecked by the static suite at S6 and S10, against whatever BASE Codex binds. |
| Ignored files | Only `!! .claude/settings.local.json`: the single exact exception Codex accepted (Q13). |

**Baseline status: PENDING. Execution cannot start.**
- **Why S0 fails today.** The approved G003 starting-state gate (§15.1, unchanged) admits only the five G003 input paths. The skill changes fail it at step 5 (exit 24): ` M skills/...` is not a snapshot path, and neither are the untracked skill paths. That is the intended behaviour. Codex confirmed at review that S0 rejects the extra tree.
- **What else would fail.** Inherited checks also depend on `skills/` matching BASE: B-SKILL-UNCHANGED (`git diff --quiet BASE -- skills/`, and every untracked skill path bound by `--expect-hash`) and B-PROTECTED.
- **What r2 does not do.** It adds **no** path exception, does not widen PATH_RE or COUNT, and does not delete or move the consolidation work.
- **Recommended route (Q14, owner or Codex decision).** Route A: when the consolidation is complete and verified, the owner commits it under their authorization. Codex then binds BASE to that commit, together with any supervisor files, through the fresh approval binding (Q9). The other two routes are recorded in §22 (Q14); each needs a reviewed Plan revision.
- **Which r1 statements this replaces.** r2 does not repeat the r1 claim that the starting state is clean. r1 §2's "Working tree: clean at the start of the session" was true when r1 was written and is kept as history.

**r1 row superseded.** r1 §2 says `core/cmd/server/main.go` "stays **unchanged**". From r2 it is a planned customization (§7.1, Q16). Its current SHA-256 (`95aa5b23…7951`) is still the rendered base that B-RENDER-PLAN and B-TOGGLE-OFF-ADDITIONS bind.

## 3. Findings that shape the plan

| # | Area | Finding (inspected) | Consequence |
|---|---|---|---|
| F1 | Core correlation | `Middleware` accepts a single regex-valid request ID and a trace ID. It keeps no span or parent and logs the raw `path`. `r.Header.Get` reads only the first of duplicated headers. `statusRecorder` records the **last** `WriteHeader`. `ParseTraceparent` accepts any version except `ff`. | Rework inside `core/internal/observability` (§7.1). `main.go` is unchanged, because the middleware is already wired there. |
| F2 | Core → intelligence | `health.Intelligence` builds a fresh request with no diagnostic headers. Its 2 s client timeout equals core's 2 s readiness deadline. | Propagate from the request context (§7.1). No new route. |
| F3 | Web | The handlers take no `Request`. `coreReady()` sends no context and returns only a boolean. There is no completion logging. Next's standalone server prints only startup lines. | Add `web/lib/request-context.ts`; customize the two routes and `core-client.ts` (§7.2). |
| F4 | Intelligence logging | `uvicorn main:app` keeps uvicorn's default text access log, which writes the raw request line with its query string to stderr, and its default text error log, which writes tracebacks. `JsonFormatter` has no key or value redaction. `/healthz` and `/readyz` send no `Cache-Control`. | Pure-ASGI completion middleware scoped to the two health paths. Take over the `uvicorn.access` and `uvicorn.error` loggers inside `configure_logging`, so the Dockerfile `CMD` is unchanged. Add redaction and no-store (§7.3). |
| F5 | Python format assertion | `jsonschema` 4.26 registers the `date-time` checker only when `rfc3339_validator` imports. It is absent from `uv.lock`, so **`format: date-time` is silently not asserted**. `main.py` already compensates at runtime, but a conformance check would treat the keyword as success. | Register a project-owned RFC 3339 `date-time` checker, with no new dependency (§8.4). |
| F6 | Python `pattern` | `jsonschema` evaluates `pattern` with `re.search`. Python's `$` also matches before a final `\n`, so `"req-12345678\n"` passes `^[A-Za-z0-9._-]{8,64}$` in Python. It fails under ECMA-262 (JSON Schema's dialect) and under Go RE2. | An ECMA-anchored `pattern` keyword in the existing validator, with an audit of the supported syntax (§8.4, Q3). |
| F7 | Python JSON parsing | `json.loads` accepts `NaN` and `Infinity`; RFC 8259 and Go reject them. | Conformance parsing is strict. Runtime envelope parsing is unaffected, because an envelope has no free numeric fields (§8.4). |
| F8 | Go validator candidates | `google/jsonschema-go` (already indirect via the MCP SDK) ignores `format` during validation. `santhosh-tekuri/jsonschema/v6` v6.0.3 (2026-06-28, Apache-2.0) supports Draft 2020-12, `AssertFormat()`, `AddResource`/`UseLoader` and meta-validation, and implements `date-time` and `uuid`. Its `go.mod` requires `golang.org/x/text v0.14.0` (`regexp2` is test-only). | Use santhosh v6.0.3 in a separate test-tool module, so core's `go.mod` is untouched (§9). |
| F9 | Harness assumptions | The harness probes published ports on `127.0.0.1`, bind-mounts host paths (`$REPO`, `$WORK`, report directories) and refuses busy fixed port sets. | The runner must share loopback and filesystem with its Docker daemon. Docker-in-Docker and socket-only Docker executors do not meet this (§12). |
| F10 | Provenance | `contracts/README.md` and `local.yml` are rendered from templates (`.tmpl`) that differ in the intelligence-disabled render. Customizing them would break B-TOGGLE-OFF-ADDITIONS' equal-base rule. | Neither is changed. The corpus lives in `contracts/conformance/` and compose overlays are avoided (§14). |
| F11 | Nested deadlines | Web's `AbortSignal.timeout(2000)` starts before core's 2 s readiness deadline, which bounds its intelligence probe. A *hung* (paused) intelligence therefore makes core answer at about 2.0 s, and web usually aborts first, so web reports `core: unavailable`. A *stopped* intelligence fails fast and is unaffected. | Existing behaviour; the Spec preserves the deadlines. Assertions follow §6.3. Codex decision in Q4. |
| F12 | Ignored-file gate | G002's S0 gate rejects every ignored path. The Claude Code client's ignored `.claude/settings.local.json` now exists in the checkout. | One exact, reported exception in the G003 gate. Every other ignored path, `.env` and `.env.local` included, still fails (§15.1, demonstrations T1–T3). |

**r2 findings.** These are added to F1–F12 above. Where they conflict, the r2 finding governs. In particular, F1's "`main.go` is unchanged" is superseded by F16 and F17, and F3's "Next's standalone server prints only startup lines" by F15. Sources are cited, and the probes are in §15.5.

| # | Area | Finding (inspected or probed) | Consequence |
|---|---|---|---|
| F13 | GitLab attribution | GitLab's [predefined-variables reference](https://docs.gitlab.com/ci/variables/predefined_variables/) documents:<br>• `CI_MERGE_REQUEST_SOURCE_BRANCH_SHA` and `CI_MERGE_REQUEST_TARGET_BRANCH_SHA`: "The variable is empty in merge request pipelines. The SHA is present only in merged results pipelines."<br>• `CI_MERGE_REQUEST_EVENT_TYPE` is `detached`, `merged_result` or `merge_train`.<br>• `CI_COMMIT_BRANCH` is "Not available in merge request pipelines or tag pipelines".<br>• `CI_COMMIT_SHA` is "The commit revision the project is built for".<br>Further:<br>• [Merge-request pipelines](https://docs.gitlab.com/ci/pipelines/merge_request_pipelines/) "Run on the contents of the source branch only".<br>• [Merged-results pipelines](https://docs.gitlab.com/ci/pipelines/merged_results_pipelines/) (Premium/Ultimate) test "a temporary merged commit … This commit doesn't exist in either branch", and fall back to a standard merge-request pipeline on conflict.<br>• The documented `CI_PIPELINE_SOURCE` values are on the [job rules](https://docs.gitlab.com/ci/jobs/job_rules/) page. | r1's `mr-source` rule required HEAD to equal a variable that is empty in exactly that pipeline type. r2 attributes from `CI_COMMIT_SHA` plus the event type, and treats the merged-result SHAs as contradictory in a detached pipeline (§11.2). |
| F14 | `after_script` and timeouts | GitLab's [`after_script`](https://docs.gitlab.com/ci/yaml/#after_script) documentation says it:<br>• runs in a separate shell, with the working directory reset and no access to `script` changes;<br>• has its own timeout, 5 minutes by default (`RUNNER_AFTER_SCRIPT_TIMEOUT`);<br>• cannot fail a job whose `script` succeeded;<br>• "For jobs that time out: `after_script` commands do not execute by default", unless `RUNNER_SCRIPT_TIMEOUT` and `RUNNER_AFTER_SCRIPT_TIMEOUT` together fit within the job timeout ([runner configuration](https://docs.gitlab.com/ci/runners/configure_runners/)).<br>`artifacts:when: always` excludes timeouts. | Gating, scanning and cleanup verification all happen inside `validate.sh`, under an internal deadline below the job timeout. `after_script` is a non-gating backstop only (§11.1, §11.4). |
| F15 | Next.js error output | Next.js 16.3.8 source, at tag `v16.3.8`:<br>• `packages/next/src/build/templates/app-route.ts` rethrows a route-handler error. The outer catch calls `routeModule.onRequestError(req, err, …, silenceLog=false, …)` and then sends `new Response(null, { status: 500 })`.<br>• `route-modules/route-module.ts` `onRequestError` calls `routerServerContext.logErrorWithOriginalStack(err, 'app-dir')` or `console.error(err)`. Both print the error's message and stack. | r1's §5.4 assumption ("route handlers print nothing") was wrong for thrown errors. The wrapper now catches, logs the bounded record and returns that 500 with an empty body itself, plus `Cache-Control: no-store` and `X-Request-Id`. A real-framework probe (B-WEB-ERROR-OUTPUT) proves it (§5.3, §7.2, §10). |
| F16 | Core exit path | `main.go` prints `run()` errors with `fmt.Fprintln(os.Stderr, err)`. Probe P-GO shows a listen error carrying the configured host verbatim: `listen tcp: lookup bad-CANARY-….invalid …`. `config.Load` messages name variables only, never values, but they are still non-JSON free text. | A bounded JSON `core.exit` record with an `error_type` category replaces the print (§5.6, §7.1). `main.go` becomes a reviewed customization (Q16). |
| F17 | Core `net/http` error log | Go 1.27.1 `net/http/server.go` (`grep` of the local `golang:1.27.1-alpine` image) logs these through `Server.ErrorLog`, or the standard `log` package when it is nil:<br>• superfluous and hijacked `WriteHeader`/`Write` calls (with the caller's function, file and line);<br>• an invalid `Content-Length` value;<br>• conflicting length headers;<br>• `http: panic serving <addr>: <value>` with a stack;<br>• TLS handshake and `Accept` errors.<br>Probe P-GO reproduced the plain-text lines, including a canary-bearing panic value. Go 1.27.1 does **not** log semicolon queries (no such call site; the probe produced no line). `http.ErrAbortHandler` is silent. | `main.go` installs `ErrorLog: observability.ServerErrorLog(logger)`. A classifying writer turns each line into one bounded `http.server.error` record and never copies the text. The probe's prototype produced only canary-free JSON (§7.1, §15.5). |
| F18 | Intelligence startup failure | `main:app` is built lazily in the module's `__getattr__`. In uvicorn 0.54.0, `Config.load_app` catches only `ImportFromStringError`, so a `ConfigError` raised by `create_app` escapes as an uncaught traceback: non-JSON, and it carries the exception message. uvicorn's `LOGGING_CONFIG` gives the `uvicorn` logger a stderr text handler (`uvicorn.error` propagates to it) and `uvicorn.access` a stdout handler whose line includes the request line. `server.py` logs a bind `OSError` object directly (`logger.error(exc)`). | `__getattr__` catches startup failures, writes one bounded `intelligence.startup.failed` record and exits with `SystemExit(3)`. The takeover now replaces the `uvicorn` logger's handlers, and uvicorn records are mapped by message *template* to fixed events, with arguments dropped (§7.3). |
| F19 | Python fallbacks | Records from a logger that has no handlers reach `logging.lastResort`, which writes the raw message to stderr. `warnings` writes plain text unless captured. | A single JSON handler on the root logger, with third-party records reduced to `library.log`, plus `logging.captureWarnings(True)` (§7.3). The real-image and full-log checks fail on any remaining raw line (R2 in §21). |
| F20 | Schema-negative masking | r1 mounted the SN files over the pinned schema paths. In normal mode the digest check runs first, so every SN case would have failed on the digest alone. | Separate audit mode, declared classes and a derivation check, while normal mode keeps digest enforcement (§8.3). |
| F21 | Working tree | The owner-authorized skill consolidation is in progress (§2.1). | Baseline PENDING. No S0 change (§15.1a, Q14). |
| F22 | gitleaks exit codes | Probe P-SCAN (gitleaks v8.30.1 `dir`): a clean directory exits 0. A planted generated token exits with the configured `--exit-code 23`. A missing configuration or path exits **1**, the same as gitleaks' default leak code. | Scans use `--exit-code 23`. Exit 0 means clean, 23 means a finding, and anything else is a scanner failure that fails closed (§11.2). |
| F23 | r1 count error | r1 §14.3 listed 18 validation-tool files but said "(17)". | Corrected in §14.3 (r2: 20 files). |

## 4. Diagnostic context contract (R2)

### 4.1 Request ID

- **Accepted inbound value.** Exactly one `X-Request-Id` field value matching `^[A-Za-z0-9._-]{8,64}$`, evaluated with full-string semantics:
  - Go RE2 `MatchString`;
  - JavaScript without the `m` flag;
  - Python `re.fullmatch`.
- **Rejected inbound value.** A missing value, more than one field line, a comma-joined value (Node joins duplicated fields as `a, b`), an over-long or non-matching value, or a value that contains a credential the receiving service holds (§4.4).
  - A rejected value is **never reflected or logged**.
  - Instead the service generates `request_id = 32 lowercase hex` from 16 random bytes (`crypto/rand`, `node:crypto` `randomBytes`, `secrets.token_hex`), and records `request_id_source` as `generated` rather than `inbound`.
- **Response header.** All six scoped health responses carry the effective `X-Request-Id`. Core already returns it on every response.

### 4.2 `traceparent` profile (narrower than W3C; no full-compliance claim)

Based on the [W3C Trace Context Recommendation](https://www.w3.org/TR/2021/REC-trace-context-1-20211123/).

- **Accepted.** Exactly one field value, at most 512 bytes before parsing, that fully matches `^00-([0-9a-f]{32})-([0-9a-f]{16})-([0-9a-f]{2})$`, where:
  - the trace ID is not all zeros;
  - the parent ID is not all zeros;
  - the length is exactly 55, which follows from the pattern.
- **Flags.** Any two lowercase hex digits are accepted. Only the `sampled` bit (`flags & 0x01`) is interpreted. Unknown bits are ignored and **never propagated**: the outbound flags are `01` when sampled is set, otherwise `00`. A newly generated trace uses `00`. The flag confers no authority, sampling, content capture or export.
- **Versions.**
  - `ff` is invalid.
  - **Every version other than `00` is unsupported in this profile.** The service restarts with a new local trace. This is narrower than W3C's SHOULD-parse rule for higher versions, and is documented as such.
  - Uppercase hex, the wrong length, extra trailing fields on `00`, internal whitespace, a non-hex character, a duplicated field or a comma-joined value are all invalid.
- **On an invalid value.** The service generates a new nonzero 16-byte trace ID. The invalid value is never reflected, logged or forwarded, no HTTP 500 is produced, and `trace_source` is recorded as `generated`.
- **Not supported:** `tracestate` and `baggage`. They are never read, forwarded or logged.

### 4.3 Spans and parent relationship (server-span-as-caller convention)

- **Per-hop span.** Every receiving service creates a fresh nonzero 8-byte `span_id` for its server operation. It regenerates if the value is all zeros, and it never copies the incoming parent ID.
- **Parent.** `parent_span_id` is the inbound `traceparent` parent ID, present only when that `traceparent` is valid.
- **Outbound.** Each outbound call carries `traceparent: 00-<trace_id>-<caller span_id>-<flags>`. The parent ID is "the ID of the current operation", as W3C requires when mutating `traceparent`. No separate client span is recorded.
- **Resulting chain on the full stack:**
  - `web.parent_span_id` = the caller's parent, absent when the context was generated;
  - `core.parent_span_id` = `web.span_id`;
  - `intelligence.parent_span_id` = `core.span_id`.

  All three records share `trace_id` and `request_id`, and their span IDs are distinct.
- **Datastores.** Probes to PostgreSQL, Redis and MinIO carry no tracing protocol.

### 4.4 Credentials supplied as diagnostic values (Q5)

- **Core** treats an inbound request ID that contains any configured secret value (`Config.SecretValues()`, reached through the redacting logger's handler) as invalid, and generates a new one. Core never echoes, logs or forwards that ID. Its logger additionally scrubs secret values from every string (existing behaviour).
- **Intelligence** does the same for `INTELLIGENCE_SERVICE_TOKEN` and `OPENROUTER_API_KEY`, and its formatter scrubs them (§7.3).
- **Web** holds no credentials, so it cannot classify a value as one. It logs only regex-bounded IDs. Core refuses to adopt or forward such a value.

### 4.5 Header allow-list and isolation

- **Outbound headers set by application code:** exactly `X-Request-Id` and `traceparent`.
  - Transport defaults remain: Go `User-Agent` and `Accept-Encoding`, and undici's standard request headers.
  - No inbound header, cookie, `Authorization`, `baggage` or `tracestate` is ever copied.
- **Request-local state:**
  - **Core:** values stored in the per-request `context.Context`.
  - **Web:** an explicit `RequestContext` argument. There is no module-level or `AsyncLocalStorage` state.
  - **Intelligence:** local variables inside the ASGI call. There is no `ContextVar`, because the health handlers make no outbound call.
- Isolation therefore holds by construction. It is still tested under concurrency, failure and cancellation (§16, AC2).

## 5. Log contract (R4)

### 5.1 Envelope for every application-owned JSON line (all three services)

| Key | Type | Rule |
|---|---|---|
| `time` | string | RFC 3339 UTC, millisecond precision, `Z`. Go: slog `ReplaceAttr`; Node: `toISOString()`; Python: `datetime.fromtimestamp(created, UTC)` |
| `level` | string | one of `debug`, `info`, `warn`, `error`. Python `WARNING` → `warn`, `CRITICAL` → `error` |
| `service` | string | `web`, `core` or `intelligence` |
| `event` | string | Core: slog's message key is renamed to `event`. Intelligence: `msg` becomes `event`; the existing `logger` key is kept. |

Startup and error records keep their messages (for example core's `core listening` and intelligence's `configured`) under `event`.

### 5.2 Completion record (`event = "http.request.completed"`)

| Key | Type | Rule |
|---|---|---|
| `method` | string | `GET`, `HEAD`, `POST`, `PUT`, `PATCH`, `DELETE` or `OPTIONS`; anything else becomes `OTHER` |
| `route` | string | Normalized, never the raw path:<br>• core: the matched `ServeMux` pattern path (`/healthz`, `/readyz`), else `unmatched`;<br>• web: the fixed handler route (`/api/health`, `/api/ready`);<br>• intelligence: `/healthz`, `/readyz` |
| `status` | integer | The final HTTP status: the first `WriteHeader`, or an implicit 200 on the first write. `0` only when the connection aborts before headers. |
| `duration_ms` | integer ≥ 0 | Monotonic elapsed time, floored |
| `request_id`, `request_id_source` | string, enum | §4.1 |
| `trace_id`, `trace_source` | 32 hex nonzero, enum | §4.2 |
| `span_id` | 16 hex nonzero | §4.3 |
| `parent_span_id` | 16 hex nonzero | **Present only** when the inbound parent was valid |
| `trace_flags` | `00` or `01` | The outbound flags (§4.2) |
| `outcome` | enum | `completed`, `cancelled` (client gone or context cancelled) or `failed` (panic or exception) |
| `error_type` | string `^[A-Za-z][A-Za-z0-9_.]{0,63}$` | Only when `outcome = failed`. Go: `panic`; Node: the error's `name`; Python: the exception class name. Never message text or a stack. |
| `core_probe` | enum | **Web `/api/ready` only:** `ok`, `not_ready`, `timeout` or `unreachable` |

- **Level.** `failed` → `error`; status 500 or above → `warn`; otherwise `info`.
- **Strictness.** No other key is allowed, apart from intelligence's `logger`.
- **Machine-checked contract.** The record is checked by `infra/validation/tools/internal/logrecord/completion.schema.json`, through the `logcheck` command (§9).

**r2 additions to §5.1–§5.2 (correction 5).**
- **Constant events.** An application-owned `event` value is a constant authored in repository code: a message template, never formatted with runtime values. The constants already in the code are kept (`core listening`, `configured`, `execution`). New events use dotted identifiers. `logcheck` validates `event` against a per-service catalogue, `infra/validation/tools/internal/logrecord/events.json` (§9). A record whose `event` is not in its service's catalogue fails, so arbitrary text cannot pass as an event name.
- **Web `error_type`.**
  - The error's `name`, when the value is an `Error` and `name` matches `^[A-Za-z][A-Za-z0-9_.]{0,63}$`;
  - otherwise `Error` for an `Error` whose `name` does not match;
  - `NonError` for a thrown non-`Error` value.
- **Startup, exit and framework records** use the same §5.1 envelope. They are specified in §5.6.

### 5.3 Exactly one record (r2; supersedes r1 §5.3 in Appendix R1)

- **Core** is unchanged from r1.
  - The middleware logs once in a `defer` guarded by a flag.
  - A handler panic is logged with `outcome: failed`, `error_type: panic`. The middleware then re-panics with `http.ErrAbortHandler`, which `net/http` neither prints nor passes to `ErrorLog` (F17). The connection aborts as before.
  - A request whose context was cancelled is logged with `outcome: cancelled`.
- **Web** (changed in r2; F15). `withCompletionLog` logs exactly once, in `finally`.
  - **Normal path.** The handler's own response is returned, with `X-Request-Id` added.
  - **On an exception** the wrapper does **not** rethrow. It records `outcome: failed` and the bounded `error_type`, and returns `new Response(null, { status: 500, headers: { "Cache-Control": "no-store", "X-Request-Id": <effective id> } })`.
    - This is the status and empty body that Next 16.3.8 itself sends after logging, so the only observable differences are the two additive headers (Q15).
    - Because no error reaches Next, Next prints nothing.
  - **On abort.** When `request.signal.aborted` is set at completion, the wrapper records `outcome: cancelled`, with the same no-rethrow rule.
- **Intelligence** is unchanged from r1.
  - The ASGI middleware logs once in `finally`.
  - An exception becomes `failed` with its class name, then is re-raised, so uvicorn's "Exception in ASGI application" is emitted as the bounded `uvicorn.asgi_exception` record (§5.6).
  - `CancelledError` becomes `cancelled`, then is re-raised.
- **Scope** is exactly the six health routes, as in r1.
  - Core keeps logging every request it receives, including 404 and 405, with `route: unmatched`.
  - Intelligence's `/internal/*` path keeps its execution export record and no longer has an access line (F4).
- **Duplicate writes.** The recorder forwards every call unchanged, as before, but reports the first status.

### 5.4 Framework output (r2; supersedes r1 §5.4 in Appendix R1)

`logcheck` classifies **every** line of each service's raw output, in both stacks and in the error-output checks. Any line that is not valid JSON with a catalogued event, and does not match the exact allow-list below, fails.

| Service | Actual framework behaviour (source) | G003 handling | Allowed non-JSON lines |
|---|---|---|---|
| core | `net/http` writes `ErrorLog` lines in plain text; panics other than `ErrAbortHandler` print value and stack (F17, probe P-GO). `main()` prints `run()` errors raw (F16). | `ErrorLog` is the classifying adapter (`http.server.error` with a fixed `error_type`). `main()` writes one `core.exit` record (§5.6). Panics in scoped handlers are re-raised as `ErrAbortHandler`. | none |
| web | Next 16.3.8 prints its startup banner. For a thrown route error it calls `console.error(err)` or `logErrorWithOriginalStack`, which print message and stack (F15). | Route handlers never let an error escape the wrapper (§5.3). | Exactly r1's startup patterns: `^ ▲ Next\.js [0-9]+\.[0-9]+\.[0-9]+$`, `^   - Local:        http://[^ ]+$`, `^   - Network:      http://[^ ]+$`, `^ ✓ Starting\.\.\.$`, `^ ✓ Ready in [0-9.]+m?s$`, and the empty line. They are confirmed against the observed output of the real-framework probe (WE0, §10); any deviation is recorded and returned for review, and the list is never widened silently. |
| intelligence | uvicorn 0.54.0 sends text through the `uvicorn` logger's stderr handler and access lines (with the request line) to stdout. `load_app` lets non-import errors escape as tracebacks. Unhandled loggers fall back to `logging.lastResort`; `warnings` writes text (F18, F19). | `configure_logging` takes over the `uvicorn` logger, disables `uvicorn.access`, adds the root JSON handler and captures warnings. `main.__getattr__` turns a startup failure into one bounded record and `SystemExit(3)` (§7.3). | none |

Runtime crashes outside request handling, for example a Go `fatal error` or a Python interpreter abort, are not request or startup paths G003 controls. If one occurs in a tested run, `logcheck` fails it as an unclassified line; it is reported, never suppressed.

### 5.5 Redaction (existing protections preserved; equivalents added)

- **Core (existing, retained):**
  - sensitive-key regex `password|passwd|secret|token|api_?key|authorization|cookie|dsn|credential`, including nested groups;
  - configured-secret-value scrubbing in the message and in string and non-string attributes.
- **Intelligence (new, in `JsonFormatter`):**
  - the same key regex, applied recursively to `fields`;
  - configured-secret-value scrubbing of `event` and of every string;
  - exception type only (existing);
  - string values capped at 512 characters.

  `JsonFormatter()` keeps its no-argument constructor, because `test_observability.py` uses it. Secrets are supplied through `configure_logging(level, secrets=...)`, which `create_app` calls.
- **Web.** Records are built only from the fixed keys in §5.2, with values that are validated IDs, enums or integers. A key-regex guard drops any sensitive-looking key. No URL, header, error message or configuration value is ever logged.
- **Never logged anywhere:**
  - raw headers, cookies or query strings;
  - request or response bodies;
  - prompts, completions, retrieved content or envelopes;
  - authorization references;
  - tenant or user identifiers;
  - URLs, DSNs or exception text.

### 5.6 Startup, exit and framework error records (r2, correction 5)

All of these use the §5.1 envelope (`time`, `level`, `service`, `event`). They never contain message text, stack frames, argument values, URLs, addresses or configuration values.

| Service | `event` | Extra keys (bounded) | Emitted when |
|---|---|---|---|
| core | `core.exit` (level `error`) | `error_type`, one of:<br>• `config_invalid`: `config.Load`'s fixed `invalid configuration: ` prefix;<br>• `database_pool_invalid`: the exact existing `cannot configure database pool` error;<br>• `listen_failed`: `*net.OpError` with `Op == "listen"`;<br>• `serve_failed`: any other `*net.OpError`;<br>• `shutdown_timeout`: `context.DeadlineExceeded`;<br>• `unknown`.<br>For `config_invalid` also `config_keys`: a sorted, de-duplicated array of variable names taken from the message, each kept only if it is in a fixed allow-list of the names `config.Load` reads. A unit test derives that list from `config.Load` itself (§7.1). | `run()` returns an error; the exit status stays 1 |
| core | `http.server.error` (level `warn`) | `error_type`, one of `superfluous_write_header`, `hijacked_write_header`, `hijacked_write`, `invalid_content_length`, `conflicting_length_headers`, `handler_panic`, `tls_handshake_error`, `accept_error`, `other`. The prefixes are those of the Go 1.27.1 call sites (F17). | `net/http` writes to `Server.ErrorLog` |
| intelligence | `intelligence.startup.failed` (level `error`) | `error_type`: the exception class name (`^[A-Za-z][A-Za-z0-9_.]{0,63}$`, else `Exception`) | `create_app()` raises while uvicorn builds `main:app`; the process then exits 3 |
| intelligence | `uvicorn.process_started`, `uvicorn.startup_waiting`, `uvicorn.startup_complete`, `uvicorn.listening`, `uvicorn.shutting_down`, `uvicorn.shutdown_waiting`, `uvicorn.shutdown_complete`, `uvicorn.process_finished`, `uvicorn.app_load_failed`, `uvicorn.invalid_request`, `uvicorn.asgi_exception`, `uvicorn.server_error`, `uvicorn.other` | `error_type` (class name) when the record has exception information, or when the logged object is an exception (bind `OSError`) | Mapped by the record's **template** (`record.msg`) against the uvicorn 0.54.0 strings, with arguments dropped. The `server.py` templates were verified from source. The lifespan and h11 templates are confirmed against the installed package at S5. Any unmatched template becomes `uvicorn.other`, so a mismatch can lose only an event name, never privacy. |
| intelligence | `library.log` | `logger` (the logger name, `^[A-Za-z0-9_.-]{1,64}$`), plus `error_type` if the record has exception information | A third-party logger at WARNING or above reaches the root handler |
| intelligence | `python.warning` | `category`: the warning class name | `warnings` output, through `logging.captureWarnings(True)` |
| web | none new | | Route errors are caught by the wrapper (§5.3). Next's own startup lines remain the allow-listed non-JSON lines (§5.4). |

Intelligence's own logger records (`intelligence`, `intelligence.*`) keep `event = record.msg`, the template, with arguments never interpolated. Their existing constants (`configured`, `execution` and the service-token warning text) are catalogued as they are. A record whose `msg` is not a string becomes `intelligence.message`.

## 6. Health semantics and bounds (R1)

### 6.1 Endpoints (bodies and status unchanged; additive headers only)

| Service | Route | Body (exact bytes) | Status | Headers after G003 |
|---|---|---|---|---|
| core | `GET /healthz` | `{"status":"alive"}\n` | 200 | `Content-Type: application/json`, `Cache-Control: no-store`, `X-Request-Id` (all existing) |
| core | `GET /readyz` | `{"status":…,"product_name":…,"checks":{"object_store":…,"postgres":…,"redis":…}[,"optional":{"intelligence":…}]}\n` | 200 / 503 | as above (existing) |
| web | `GET /api/health` | `{"status":"alive"}` | 200 | `Cache-Control: no-store` (existing), **`X-Request-Id` (new)** |
| web | `GET /api/ready` | `{"status":"ready"\|"unready","checks":{"core":"ok"\|"unavailable"}}` | 200 / 503 | `Cache-Control: no-store` (existing), **`X-Request-Id` (new)** |
| intelligence | `GET /healthz` | `{"status":"ok"}` | 200 | **`Cache-Control: no-store` and `X-Request-Id` (new)** |
| intelligence | `GET /readyz` | `{"status":"ready"}` (also when no model is configured) | 200 | **`Cache-Control: no-store` and `X-Request-Id` (new)** |

- Core's required checks remain `postgres`, `redis` and `object_store`. `intelligence` stays under `optional` and is absent when intelligence is disabled.
- The deadlines remain:
  - core readiness: 2 s;
  - core's intelligence client: 2 s;
  - web's fetch: 2 s.
- No error text, DSN, URL or credential appears in any body.
- Intelligence's body bytes are unchanged: FastAPI's default dict response and `JSONResponse` render identically (`separators=(",", ":")`).

### 6.2 Timing protocol

- **Clock.** Each single request is timed by `httpprobe` (§9) with Go's monotonic clock.
- **Where it runs.** Inside the target service's network namespace (`--network container:<cid>`): web on `127.0.0.1:3000`, core on `127.0.0.1:8180`, intelligence on `127.0.0.1:8000`.
- **Bound.** Every injected timeout or outage case must return in **≤ 5000 ms**, or the criterion fails.
- **Recovery.** Recovery polling is separate. It has a documented finite deadline of **90 s per case** and is never used to time a single request.

### 6.3 Outage and timeout matrix (disposable copies only)

The cases run on the enabled copy (`variants_ops`, ports 3200, 8280, 8200, 5632, 6579, 9200 and 9201) and on the disabled copy (`variants_off`).

| Case | Injection | Assertions (each single request ≤ 5 s) | Enabled | Disabled |
|---|---|---|---|---|
| O1–O3 | `compose stop` postgres / redis / minio | core `/readyz` 503 naming the dependency; web `/api/ready` 503 (`core_probe: not_ready`); both liveness probes 200; continuity records (§16 AC3); recovery to 200 | ✓ | ✓ |
| O4–O6 | `docker pause` postgres / redis / minio | Same statuses; web `core_probe` is `not_ready` or `timeout`, because core answers at about 2 s (F11); continuity in web and core; unpause, then recovery | ✓ | ✓ |
| O7 | stop intelligence | core 200 with `optional.intelligence = unavailable`; web 200; no intelligence record; recovery restores the three-hop chain | ✓ | — |
| O8 | pause intelligence | **Core `/readyz` (direct) 200 with `optional.intelligence = unavailable`.** Web `/api/ready` returns within the bound with either 200 or 503 `core_probe: timeout`. The observed value is recorded (F11, Q4). After unpause, intelligence may log 0 or 1 record for the queued probe; more than one, or a mismatched one, fails. | ✓ | — |
| O9 | stop core | web `/api/ready` 503 (`core_probe: unreachable`), `/api/health` 200, both no-store; only a web record | ✓ | ✓ |
| O10 | pause core | web `/api/ready` 503 (`core_probe: timeout`). After unpause, core logs at most one record for that request ID, with `outcome: cancelled` and matching IDs. | ✓ | ✓ |

The existing B-OUTAGE and B-PERSIST then run unchanged on the same copy, after recovery has been asserted.

## 7. Service changes (r2; r1 §7 is in Appendix R1)

### 7.1 Core (`core/internal/observability`, `core/internal/health`, and `core/cmd/server/main.go`, which is now customized)

- **`correlation.go` (customized)** is as in r1. It contains:
  - `TraceContext`;
  - `parseRequestID` (with `isSecret`) and `parseTraceparent`;
  - `ParseTraceparent`, kept for existing callers;
  - `RequestID`, `TraceID`, `SpanID` and `FromContext`;
  - `InjectOutbound`, which sets exactly `X-Request-Id` and `traceparent`.

  The `Middleware` signature is unchanged. It now reads `Header.Values`, stores the context, sets `X-Request-Id`, wraps the writer (first status, implicit 200, `Unwrap`), takes `route` from `r2.Pattern` (R4 fallback in §21), and logs once (§5.3).
- **`logging.go` (customized)** is as in r1: a `ReplaceAttr` gives UTC millisecond `time`, lowercase `level`, `msg` → `event` and `service=core`. Redaction is unchanged. An unexported `secretMatcher` is added.
- **`checks.go` (customized)** is as in r1: `observability.InjectOutbound(ctx, req.Header)` before `client.Do`, with the 2 s timeout unchanged.
- **`exit.go` (new).** `LogExit(w io.Writer, err error)` writes exactly one `core.exit` JSON line (§5.6) through a plain `slog` JSON handler with the same `ReplaceAttr`.
  - It does not need the configured secrets: the record contains no free text. `config_keys` values are restricted to the allow-list.
  - It never writes `err.Error()`.
- **`errorlog.go` (new).** `ServerErrorLog(logger *slog.Logger) *log.Logger` returns `log.New(classifyingWriter{logger}, "", 0)`.
  - Each `Write` maps the line's prefix to one `http.server.error` category (§5.6) and emits one record.
  - It returns `len(p), nil`, so `net/http` is never blocked.
  - The text is discarded.
- **`core/cmd/server/main.go` (customized; Q16).** Exactly two hunks; no other byte changes:
  1. `fmt.Fprintln(os.Stderr, err)` becomes `observability.LogExit(os.Stderr, err)`, and the now-unused `"fmt"` import is removed. The exit status stays 1.
  2. The `http.Server` literal gains `ErrorLog: observability.ServerErrorLog(logger),`.

  Routes, imports (apart from removing `fmt`), timeouts, the health wiring, the healthcheck subcommand and shutdown behaviour are unchanged. `main.go` is a plain base-template file in the archetype (`skills/project-archetype/templates/base/core/cmd/server/main.go`, not a `.tmpl`), so the intelligence-disabled render has the same rendered base. The customization is therefore compatible with B-TOGGLE-OFF-ADDITIONS once `core/cmd/server/main.go` is in `APPLICABLE` (§10, §17).
- **New tests:**
  - `context_test.go`: everything listed in r1 §7.1 (tables, concurrency, cancellation, panic, status, routes, keys, non-reflection, byte-identity).
  - `propagation_test.go`: as in r1.
  - **`exit_test.go` (new):**
    - each `error_type` category from a constructed error (`config.Load` with an empty environment, `*net.OpError` listen and accept, `context.DeadlineExceeded`, the exact pool error, an unknown error);
    - exactly one line, valid JSON, envelope keys present;
    - a canary in the error text, in a wrapped error and in an unknown error never appears;
    - **an allow-list sync test** that calls `config.Load` with an empty environment and asserts that every variable name in the resulting message is in the allow-list, so the list cannot drift silently from `config.Load`.
  - **`errorlog_test.go` (new):**
    - a real `httptest.Server` with `ErrorLog: ServerErrorLog(logger)` and handlers that call `WriteHeader` twice, panic with a canary value, and panic with `http.ErrAbortHandler`;
    - the captured output is JSON only, with categories `superfluous_write_header` and `handler_panic`;
    - nothing is logged for `ErrAbortHandler`;
    - the canary never appears;
    - a direct-`Write` table covers every prefix in §5.6, plus `other`.

  The existing `observability_test.go` and `health_test.go` are unchanged and must pass.

### 7.2 Web (routes and BFF; no UI file changes)

- **`web/lib/request-context.ts` (new; `import "server-only"`; no relative imports)** exports `deriveContext`, `outboundHeaders` and `withCompletionLog`.
  - **r2 change.** `withCompletionLog(route, handler)` returns `(request: Request) => Promise<Response>` and never rethrows (§5.3). The catch path returns `new Response(null, { status: 500, headers })` with `Cache-Control: no-store` and the effective `X-Request-Id`. `error_type` is bounded as in §5.1–§5.2 (r2 additions).
  - It times with `performance.now()` and writes one JSON line with `process.stdout.write`.
- **`web/lib/core-client.ts` (customized)** is as in r1: `coreReady(ctx, timeoutMs = 2000): Promise<CoreProbe>` maps to `ok`, `not_ready`, `timeout` or `unreachable`, still with `cache: "no-store"`, `AbortSignal.timeout` and `server-only`.
- **`web/app/api/health/route.ts` and `web/app/api/ready/route.ts` (customized)** are as in r1: `export const GET = withCompletionLog(...)`, with `dynamic = "force-dynamic"`. The success bodies, statuses and `Cache-Control` are byte-identical.
- **`web/package.json` (already registered)** gains the `test:unit` script and the extended `check`, as in r1. There are no dependency changes.
- **`web/tests/unit/request-context.test.mjs` (new)** keeps r1's coverage, plus:
  - a throwing handler returns status 500 with an empty body, `Cache-Control: no-store` and `X-Request-Id`;
  - it writes exactly one record, `outcome: failed`;
  - it does **not** reject;
  - non-`Error` and bad-`name` values map to `NonError` and `Error`;
  - an aborted signal gives `cancelled`;
  - error message, `cause` and own-property canaries never appear in the output.
- **`web/tests/fixtures/error-probe/wrapped-route.ts` and `control-route.ts` (new; test fixtures only).**
  - They are copied into `app/api/__probe/{wrapped,control}/route.ts` **only inside a disposable copy** built by B-WEB-ERROR-OUTPUT (§10), the same pattern as G002's token-scope fixture.
  - `wrapped` uses `withCompletionLog` and a `mode` query parameter: `ok`, `throw` (an `Error` carrying canaries from `PROBE_CANARY` in message, `cause` and an own property), `nonerror`, `badname`, or `slow` (waits 3 s).
  - `control` throws the canary **without** the wrapper. It is the positive control proving that the capture sees Next's real error output.
  - Production `app/` is unchanged: B-WEB-BOUNDARY still requires exactly `/`, `/api/health` and `/api/ready`.
  - Both files are type-checked and linted by `web-check` (`tsconfig` includes `**/*.ts`).

### 7.3 Intelligence (`src/`, check stage only in the image)

- **`intelligence/src/observability/http.py` (new)** is `CompletionLogMiddleware`, as in r1.
- **`intelligence/src/observability/export.py` (customized).**
  - **`JsonFormatter`:** the §5.1 envelope, `service`, redaction (§5.5) and a 512-character value cap, as in r1. Plus the r2 event rules (§5.6):
    - intelligence loggers use the template and never `getMessage()`;
    - uvicorn records are mapped by template;
    - other libraries become `library.log`;
    - warnings become `python.warning`;
    - exceptions keep their type only.
  - **`configure_logging(level=INFO, secrets=())`** is idempotent, with the secrets updatable as in r1. **r2:**
    - it installs **one** JSON stream handler on the root logger, with `intelligence` at INFO and the root at WARNING;
    - it replaces the `uvicorn` logger's handlers with the same handler, keeping `propagate=False`;
    - it sets `uvicorn.access` to `disabled=True` with no handlers;
    - it calls `logging.captureWarnings(True)`.

    The test asserts there are no duplicate lines. `JsonFormatter()` keeps its no-argument constructor.
  - `ObservabilityHook` is unchanged, so content capture stays constructor-only and off by default.
- **`intelligence/src/main.py` (customized).** As in r1: `create_app` passes the two secrets to `configure_logging` and the middleware, and the health handlers return `JSONResponse(..., headers={"Cache-Control": "no-store"})`. **r2:**
  - `__getattr__` builds `_app` inside `try: … except Exception as exc:`.
  - On failure it calls `configure_logging()` if it has not yet run, writes one `intelligence.startup.failed` record (`error_type` only) and raises `SystemExit(3) from None`.
  - uvicorn's `Config.load_app` does not catch `SystemExit`, so the process exits 3 without a traceback. The real-image check verifies this (B-INTEL-ERROR-OUTPUT, IE1).
- **`intelligence/src/streaming/schemas.py` (customized):** §8.4, unchanged from r1.
- **New tests** (no new packages):
  - `test_health_context.py`, `test_logging_privacy.py`, `test_conformance.py` and `conformance_exchange.py`, as in r1. r2 extends the last two (§8.3).
  - **`tests/test_error_output.py` (new):**
    - the startup-failure path through `main.__getattr__` with an invalid environment that includes a canary-bearing variable *name*: `SystemExit(3)`, one record, no canary;
    - the uvicorn template mapping for each template in §5.6;
    - an unknown template becomes `uvicorn.other`;
    - a third-party logger warning becomes `library.log` with no message;
    - `warnings.warn(<canary>)` becomes `python.warning` with no canary;
    - exactly one handler, with no duplicate lines;
    - the `uvicorn.access` logger is disabled.
- **`infra/docker/intelligence.Dockerfile` (customized)** is as in r1: the `check` stage only copies `conformance/intelligence/`.

## 8. Contract conformance (R5)

### 8.1 Layout (fixtures only; `contracts/intelligence/*.schema.json` unchanged)

```text
contracts/conformance/README.md                         provenance, case IDs, how to run, stream-check rules
contracts/conformance/intelligence/expectations.json    independently declared outcome per case, schema digests, corpus rules
contracts/conformance/intelligence/cases/envelope/*.json
contracts/conformance/intelligence/cases/event/*.json
contracts/conformance/intelligence/cases/ndjson/*.ndjson
contracts/conformance/intelligence/schema-negative/*.json
```

- **Why this location.**
  - It is kept outside `contracts/intelligence/`, so it is not copied into the runtime image's `/srv/contracts/intelligence`.
  - It is absent from the intelligence-disabled render, which B-TOGGLE-OFF already requires.
  - It is fixture-only, consistent with "contract schemas' test/fixture/documentation areas" (Q11).
- **Expectation pinning.** `expectations.json` pins the two schema SHA-256 values from §2. Both validators refuse to run if the bytes they load differ.
- **Fixture data.** All data is synthetic, hand-authored for G003, and uses RFC-style UUIDs. Values never look like real credentials, so gitleaks' generic rules are not triggered; secret-scan is not loosened.

### 8.1a r2 addition: declared schema-negative entries

`expectations.json` gains a `schema_negative` array, declared independently of either validator's output. Each entry gives:
- `id`;
- `base`: `execution-envelope` or `stream-event`;
- `mutation`: one JSON-Patch-style `add` or `replace` at a JSON Pointer, or `truncate-bytes` with a byte length, or `absent`;
- `class`: the expected audit classification.

It also declares the two normal-mode digest controls, SD01 and SD02. They have no fixture file, because they are derived at test time inside a scratch directory (§8.2a).

### 8.2 Case catalogue (file name = lowercase ID plus a slug; the expected result is fixed here)

**Envelope, valid (14):**

| ID | Case |
|---|---|
| EV01 | Minimal: required fields only, empty `allowed_tools` and `context_refs`, one message |
| EV02 | Full: `group_id`, `business_entity_id`, `callback_grant_ref`, all three tool levels, context refs, system/user/assistant messages |
| EV03 | `request_id` of length 8 |
| EV04 | `request_id` of length 64 with `.`, `_` and `-` |
| EV05 | Offset `+02:00` with a 9-digit fraction |
| EV06 | `2028-02-29` (leap year) |
| EV07 | `23:59:60Z` (leap second) |
| EV08 | Uppercase UUIDs |
| EV09 | 256 tools and 256 context refs |
| EV10 | 256 messages |
| EV11 | Every string maximum: content 200000, `authorization_ref` 512, `callback_grant_ref` 512, tool name 128, ref 256, kind 64 |
| EV12 | `trace_id` all zeros. The existing wire behaviour is preserved: the envelope schema has no nonzero rule, and none is imported. |
| EV13 | Nil UUID (shape-valid) |
| EV14 | `capability: embedding` |

**Envelope, invalid (50):**

| ID | Case |
|---|---|
| EI01 | Truncated object |
| EI02 | Trailing comma |
| EI03 | `NaN` literal |
| EI04 | Top-level array |
| EI05–EI14 | Each of the 10 required fields missing (in schema `required` order) |
| EI15 | `allowed_tools` as an object |
| EI16 | `request_id` as a number |
| EI17 | Unknown field |
| EI18 | `api_key` |
| EI19 | `password` |
| EI20 | `access_token` |
| EI21 | `authorization` |
| EI22 | `capability: vision` |
| EI23 | Tool level `admin` |
| EI24 | Extra property on a tool |
| EI25 | Role `tool` |
| EI26 | Empty messages |
| EI27 | 257 messages |
| EI28 | 257 tools |
| EI29 | Content of 200001 characters |
| EI30 | `request_id` of 7 characters |
| EI31 | `request_id` of 65 characters |
| EI32 | `request_id` with a space |
| EI33 | **`request_id` with a trailing `\n`** (F6) |
| EI34 | Uppercase `trace_id` |
| EI35 | `trace_id` of 31 characters |
| EI36 | `organization_id` not a UUID |
| EI37 | `user_id` of 35 characters |
| EI38 | `group_id` containing `g` |
| EI39 | **`2027-02-30T00:00:00Z`** |
| EI40 | **Month 13** |
| EI41 | **Hour 24** |
| EI42 | **Minute 60** |
| EI43 | **Second 61** |
| EI44 | **`12:00:60Z`** |
| EI45 | **`2027-02-29`** |
| EI46 | No offset |
| EI47 | Space separator |
| EI48 | 10-digit fraction |
| EI49 | Lowercase `t`/`z` |
| EI50 | **Offset `+24:00`** |

The bold cases EI39–EI45 and EI50 satisfy the `pattern` and fail **only** on `format: date-time`. They prove the format assertion is effective in both validators.

**Event, valid (15):**

| ID | Case |
|---|---|
| XV01 | `text_delta` with empty text |
| XV02 | `tool_call` with nested arguments |
| XV03 | `tool_result`, `is_error: false`, object content |
| XV04 | `tool_result`, `is_error: true`, null content |
| XV05 | `tool_result`, string content |
| XV06 | `citation` without locator |
| XV07 | `citation`, locator of 256 characters |
| XV08 | `structured_output` without `schema_ref` |
| XV09 | `structured_output` with `schema_ref` |
| XV10 | `usage`, all values |
| XV11 | `usage` with null `input_tokens`, `output_tokens` and `estimated_cost` |
| XV12 | `usage` with `units: 0` and zero tokens |
| XV13 | `usage` with tokens `1.0` (an integer-valued number) |
| XV14 | `error`, `retryable: true`, message of 1024 characters |
| XV15 | `done` |

**Event, invalid (34):**

| ID | Case |
|---|---|
| XI01 | Malformed JSON |
| XI02 | Unknown type |
| XI03 | Missing type |
| XI04 | Extra field |
| XI05 | Numeric text |
| XI06 | `done` without `request_id` |
| XI07 | `request_id` with a space |
| XI08 | `request_id` with a trailing `\n` |
| XI09 | Bad `call_id` |
| XI10 | Array arguments |
| XI11 | `tool_result` without `is_error` |
| XI12 | Empty citation ref |
| XI13 | String data |
| XI14 | Negative input tokens |
| XI15 | Negative output tokens |
| XI16 | Negative latency |
| XI17 | Negative cost |
| XI18 | Negative units |
| XI19 | Units `1.5` |
| XI20 | `usage` without `trace_id` |
| XI21 | Unknown usage capability |
| XI22–XI25 | `usage` with `organization_id`, `user_id`, `group_id`, `business_entity_id` (one case each) |
| XI26–XI30 | `usage` with the content fields `prompt`, `completion`, `messages`, `content`, `input` (one case each) |
| XI31 | Uppercase error code |
| XI32 | Error message of 1025 characters |
| XI33 | Extra field on `done` |
| XI34 | Uppercase `configured_provider` |

**NDJSON stream files (7).** Rules:
- every record is LF-terminated;
- no blank line;
- at least one record;
- every line parses strictly and validates against `stream-event`.

| ID | Case | Expected |
|---|---|---|
| NV01 | `text_delta`, `usage`, `done` | valid |
| NV02 | `error`, `done` | valid |
| NI01 | Final line truncated, no LF | invalid |
| NI02 | Non-JSON line in the middle | invalid |
| NI03 | Schema-invalid event in the middle | invalid |
| NI04 | Blank line | invalid |
| NI05 | Empty file | invalid |

These rules are test-profile rules for checking producer output. They make no wire-contract change.

### 8.2a Schema-negative and digest controls (r2; supersedes r1's "Schema-negative (7)" block in Appendix R1)

**Two paths, never mixed (correction 4):**
- **Normal mode** (`corpus`, `ndjson`, `peer` and the pytest corpus test). The schema digests are checked **before anything else**. Any byte difference from the two pinned SHA-256 values fails with class `schema-digest-mismatch`, with zero cases evaluated. No option or argument disables this.
- **Audit mode** (`schema-audit`). It accepts only files under `contracts/conformance/intelligence/schema-negative/`, or the canonical schema paths for the positive control. It never evaluates cases and never runs the digest step. Its only output is a classification. It exits 0 only for `ok`.

**Audit stage order** (identical in both languages, implemented separately; the first failing stage sets the class):

| Stage | Check | Class on failure |
|---|---|---|
| 1 | The file exists and is a regular file | `schema-missing` |
| 2 | Strict JSON parse (RFC 8259; `NaN` and `Infinity` rejected); the top level is an object | `schema-malformed-json` |
| 3 | `$schema` is exactly `https://json-schema.org/draft/2020-12/schema` | `schema-wrong-draft` |
| 4 | Meta-validation against the embedded Draft 2020-12 meta-schema, with no network | `schema-meta-invalid` |
| 5 | Every keyword is in the audited profile (§8.3) | `schema-unsupported-keyword` |
| 6 | Every `format` is `date-time` or `uuid` | `schema-unsupported-format` |
| 7 | Every `$ref` starts with `#/$defs/` and resolves locally. Nothing is fetched: Go's loader fails every URL, Python's registry has no retriever, and containers run `--network none`. | `schema-remote-ref` |
| 8 | Compile with assertions on | `schema-compile-failed` (no declared case expects it) |

**Cases.** Each SN fixture is the canonical base schema with exactly one mutation, so that every earlier stage passes and only the intended stage can fail:

| ID | Base | Single declared mutation | Declared class (both languages) |
|---|---|---|---|
| SN01 | envelope | `absent`: no file | `schema-missing` |
| SN02 | envelope | `truncate-bytes` to the first 64 bytes of the canonical file | `schema-malformed-json` |
| SN03 | envelope | `replace /properties/authorization_ref/type` with `5` | `schema-meta-invalid` |
| SN04 | envelope | `add /dependentSchemas` with `{"group_id": {"required": ["business_entity_id"]}}` (meta-valid; outside the profile) | `schema-unsupported-keyword` |
| SN05 | envelope | `add /properties/authorization_ref/format` with `"ipv4"` | `schema-unsupported-format` |
| SN06 | stream-event | `replace /oneOf/0/$ref` with `"https://example.invalid/x.json"` | `schema-remote-ref` |
| SN07 | envelope | `replace /$schema` with `"http://json-schema.org/draft-07/schema#"` | `schema-wrong-draft` |
| SD01 | envelope | **normal mode:** the canonical bytes plus one trailing LF, written to a scratch schema directory. This file is still valid, meta-valid and inside the profile. | `schema-digest-mismatch`, 0 cases evaluated |
| SD02 | envelope | **normal mode:** SN03's bytes in place of the canonical file, in a scratch schema directory | `schema-digest-mismatch`, 0 cases evaluated. This proves that normal mode never reaches meta-validation with unpinned bytes. |

**Derivation check (`contractcheck sn-derive`).** For every declared SN entry, it applies the declared mutation to the parsed canonical base and requires structural equality with the parsed fixture. For SN02 it requires byte equality with the truncated prefix; for SN01, that the file is absent. A fixture with a second, undeclared difference therefore fails. This shows that each SN case isolates a single defect.

**Positive audit control.** Both canonical schemas classify as `ok` in both languages, so the audit path is not vacuously failing.

**The EI39 control** is retained, in normal mode with canonical schemas: a scratch copy of `expectations.json` with EI39 set to `valid` makes both validators exit nonzero, naming EI39.

That is 120 document and stream cases, 7 schema-negative cases, 2 digest controls and 2 positive audit controls. All eight event variants appear among the valid cases. **No production schema byte changes;** every mutation exists only as a committed SN fixture or in a scratch directory.

### 8.3 Validators (r2; independent; neither reads the other's output; r1 text in Appendix R1)

- **Supported keyword profile** (unchanged from r1). The audit is shared by specification but implemented separately in each language: `$schema`, `$id`, `$defs`, `$ref` (local `#/$defs/...` only), `title`, `description`, `type`, `properties`, `required`, `additionalProperties`, `items`, `minItems`, `maxItems`, `minLength`, `maxLength`, `pattern`, `enum`, `const`, `oneOf`, `minimum`, `format`. The only formats allowed are `date-time` and `uuid`.
- **Go (`contractcheck`, santhosh v6.0.3).** It uses `NewCompiler()` (Draft 2020-12) and `AssertFormat()`; a `UseLoader` that fails every URL (meta-schemas are embedded); `AddResource` under `file:///contracts/...` URLs; strict JSON decoding; and RE2 patterns. Subcommands:
  - `corpus`: normal mode. Digest first, then the audit stages 1–8, then the cases. Writes `results-go.json`; exits nonzero on any mismatch.
  - `ndjson`: normal mode.
  - `produce-envelopes`: the test-only Go producer, as in r1.
  - **`schema-audit --schema FILE --as envelope|event`** (new): audit mode only. Prints `{"file": …, "class": …}`.
  - **`schema-negative --schemas DIR --corpus DIR`** (new): runs, for each declared entry, SN01–SN07 through `schema-audit`, `sn-derive`, the positive audit control, and SD01/SD02 through `corpus` in a scratch directory. It compares every class with the declaration and writes `schema-negative-go.json`. A wrong class fails, even when the exit status is nonzero.
  - **`sn-derive`** (new): §8.2a.
- **Python.** The existing `streaming.schemas` validator (§8.4), in the `intelligence-check` image, with strict JSON parsing.
  - The digest pin and the audit stages are implemented in the **test tooling** (`tests/conformance_exchange.py` and helpers used by `tests/test_conformance.py`), not in the runtime module.
  - `tests/test_conformance.py`:
    - parametrized over every corpus case, as in r1;
    - **plus** SN01–SN07, asserting the declared class from `expectations.json`;
    - SD01/SD02 through `tmp_path` copies, asserting that `SchemaDigestMismatch` is raised before any case is evaluated;
    - the positive audit control on both canonical schemas.

    It is asserted not to skip.
  - `tests/conformance_exchange.py` (script) has the subcommands `corpus`, `peer` and `ndjson` (r1) and **`schema-negative`** (new), which writes `schema-negative-python.json`.
- **Shared result format** (unchanged from r1): `{"validator", "schema_sha256", "corpus_sha256", "results": [{"id", "valid", "class"}], "counts", "skipped": 0}`. The schema-negative results use `{"id", "class", "expected_class", "match"}`.

### 8.4 Python validator conformance corrections (`schemas.py`; schema bytes unchanged; Q3)

- **ECMA-262 `pattern` semantics.**
  - `validators.extend(Draft202012Validator, {"pattern": _ecma_pattern})`.
  - Each schema pattern is translated once: an unescaped `$` outside a character class becomes `\Z`.
  - Patterns containing a construct whose semantics differ between ECMA-262 and Python `re` are rejected when the validator is built: `\d`, `\w`, `\s`, `\b`, `\B`, `\p`, `(?`, backreferences and lookarounds. All current schema patterns use only literal classes, quantifiers, anchors, groups and alternation.
- **`date-time` checker.** A project-owned `FormatChecker` adds `date-time`: the RFC 3339 §5.6 grammar plus calendar validation (month length and leap years).
  - Hours 00–23, minutes 00–59, seconds 00–60, offset hours 00–23 and offset minutes 00–59.
  - Second 60 is allowed only when the UTC-normalized time is 23:59.
  - `T`/`Z` are accepted in either case; this is irrelevant in practice, because the schema pattern already requires uppercase.
  - `uuid` remains jsonschema's built-in checker.
- **Audit.** §8.3, plus an assertion that `date-time` and `uuid` are present in `checker.checkers`.
- **Effect at runtime.** Intelligence's `/internal/v1/execute` now rejects, with 422, two classes of envelope it previously accepted only through gaps:
  - an envelope whose pattern-bounded identifier ends in a newline;
  - an envelope whose `deadline` is calendar-invalid. These already received 422 through `main.py`'s parse fallback, with the identical response `{"path":"/deadline","keyword":"format"}`.
- **Leap-second note.** A valid leap-second deadline (`23:59:60Z`) remains schema-valid, but `main.py`'s `fromisoformat` rejects it, as it does today. That pre-existing runtime check is not changed and is noted in OPERATIONS.md.

### 8.5 Peer exchange (harness; scratch `$WORK/contracts-exchange`, mode 777, mounted into both containers)

1. `contractcheck produce-envelopes`, which writes `go-envelopes.ndjson`.
2. `conformance_exchange.py peer`. Python accepts and executes the Go envelopes, then emits `python-events.ndjson`.
3. `contractcheck ndjson --require-types text_delta,usage,error,done`, which checks `python-events.ndjson`.
4. The harness derives corrupt copies of both real files:
   - the last line truncated;
   - a garbage line inserted;
   - an invalid event or envelope appended;
   - a blank line.

   The peer checker must reject each one. A checker that ignores a bad line cannot pass this step.

## 9. Validation tools module (test tooling, not runtime; r2; r1 text in Appendix R1)

- **Module:** `infra/validation/tools/`, `module validationtools`, `go 1.27.1`, as in r1.
  - Direct requirement: `github.com/santhosh-tekuri/jsonschema/v6 v6.0.3`.
  - Indirect requirement: `golang.org/x/text`, at the version minimal version selection resolves.
  - Both are recorded in `go.sum`.
  - It is resolved once at S2, in a container, with a module cache under the execution's external scratch root (`$EXT/gomod`, not a Docker volume as in r1). That is the only network step (§15.6).
- **Commands:**
  - `cmd/contractcheck` (§8.3), with the new subcommands `schema-audit`, `schema-negative` and `sn-derive`.
  - `cmd/httpprobe`, as in r1: repeatable and duplicate headers, header and body values read from files, `--parallel`, `--abort-after`, `--timeout`, JSON output.
  - `cmd/logcheck`, as in r1: line classification, completion-schema validation, exactly one record per request and service, and chain checks. **r2:**
    - validates every record's `event` against the per-service catalogue in `internal/logrecord/events.json` (§5.1 r2 additions, §5.6);
    - validates `error_type` against each event's enumerated categories;
    - adds a `startup` mode for the error-output checks (exit status and an expected event and category set);
    - adds a `scan` mode that reports canary hits by file and line number without printing the canary.

    Its unit-test negatives grow to include an uncatalogued event, a free-text `error_type`, and a JSON line whose `event` carries an exception message.
- **Image:** `infra/docker/validation-tools.Dockerfile`, with stages `build`, `check` and `runtime` (distroless static nonroot), as in r1. Images are tagged `g003-vtools-<run>`, labelled, and removed at cleanup.
- **Execution.** `--network none`, except `httpprobe`, which joins a service's network namespace. No production network is joined and no port is published.
- **Exclusions.** Nothing here is linked into core. B-MCP-SOURCE asserts that no `jsonschema` package appears in `go list -deps ./cmd/server`.

## 10. Harness changes (`infra/validation/`; every existing assertion preserved; r2; r1 table in Appendix R1)

| Check | Suite | Change |
|---|---|---|
| B-SHELL | static | Also shellchecks `infra/ci/validate.sh`, `infra/ci/lib.sh`, `infra/ci/validate-controls.sh`, `infra/ci/cleanup-owned.sh` and `infra/ci/runner-preflight.sh` |
| B-HADOLINT | static | Adds `validation-tools.Dockerfile` |
| B-DOCKERIGNORE | static | Adds `validation-tools.Dockerfile.dockerignore` (`.env` and `.env.*`) |
| B-DEPS | static | Adds the tools-module rule: exactly one direct requirement (`santhosh-tekuri/jsonschema/v6 v6.0.3`), and `go.sum` present. The existing rules are unchanged (goose and MCP pins, intelligence lockfile equal to the render, web C1–C12 policy). |
| **B-MCP-SOURCE (reconciled; Q16, §17)** | static | The **same security assertions** remain:<br>• the server binary links neither the MCP SDK, `core/internal/mcp` nor migration code;<br>• no binary depends on fixture or isolation packages;<br>• no transport handler or `WithIsolateDDL`;<br>• `core/internal/mcp` does not import `net/http`;<br>• the routes are exactly `GET /healthz` and `GET /readyz`.<br>**Changed:** the byte-equality test `sha(main.go) == manifest_sha(main.go)` is replaced by `sha(main.go) == MAIN_GO_G003_SHA`. That digest of the reviewed two-hunk file (§7.1) is pinned in `run.sh` at S6 and recorded in the Result.<br>**Added:**<br>• `go list -f '{{join .Imports " "}}' ./cmd/server` equals the pinned rendered import set minus `fmt`;<br>• no `jsonschema` package appears in `go list -deps ./cmd/server`.<br>The rendered base stays bound by B-RENDER-PLAN (register row) and B-TOGGLE-OFF-ADDITIONS. The reason message changes from "server main.go unchanged" to "server main.go equals its reviewed G003 digest". |
| B-WEB-BOUNDARY | static | The r1 additions (`request-context.ts` imports `server-only` and is imported only by `app/api/**/route.ts` and `lib/core-client.ts`; the core fetch passes only `outboundHeaders(ctx)`), plus: the `tests/fixtures/error-probe/*` files exist only under `web/tests/` and are never imported by `app/`, `components/` or `lib/`. The routes remain exactly `/`, `/api/health` and `/api/ready`. |
| **B-CI-CONFIG** (new) | static | §11.1 assertions on `.gitlab-ci.yml` (through yq), plus the README `platform:ci` block |
| **B-CI-CONTROLS** (new) | static | `infra/ci/validate-controls.sh` groups A–E (§11.3). It includes Docker-real cleanup controls using `busybox:1.37.0` with control labels, and the nested real `contracts` run of V4. |
| B-WRAPPER-CONTROLS, B-DEPS-CONTROLS | static | Unchanged (W1–W9; C1–C12, C8b, C8c) |
| B-RENDER-PLAN, B-RENDER-REPEAT | provenance | The register count changes from 8 to **19** (3 G001 + 5 G002 + 11 G003, §14.2). Exactly 19 conflicts and proposals. |
| **contracts suite** (new; root only; no stack) | contracts | • **B-CONTRACT-TOOLS:** the `validation-tools` check stage passes; the intelligence-check image builds; the runtime intelligence image has no `/srv/conformance`.<br>• **B-CONTRACT-GO** and **B-CONTRACT-PY:** in normal mode, each validator matches every expectation, the digests match, the audit passes and nothing is skipped.<br>• **B-CONTRACT-AGREE:** as in r1.<br>• **B-CONTRACT-PEER:** §8.5.<br>• **B-CONTRACT-NEGATIVE (r2):** in both languages, with `--network none`:<br>&nbsp;&nbsp;– SN01–SN07 in audit mode, each with its declared class;<br>&nbsp;&nbsp;– the positive audit control (`ok`) for both canonical schemas;<br>&nbsp;&nbsp;– `sn-derive` passes;<br>&nbsp;&nbsp;– SD01 and SD02 in normal mode, each `schema-digest-mismatch` with 0 cases evaluated;<br>&nbsp;&nbsp;– EI39 flipped in a scratch copy: both validators exit nonzero, naming EI39.<br>A wrong class is a FAIL even when the exit status is nonzero. |
| **B-HEALTH-ROOT** (new) | root | Host probes on the actual root (ports 3100, 8180, 8100): six routes with status, exact body, no-store and a valid `X-Request-Id` |
| **B-CORE-ERROR-OUTPUT** (new) | root | Runs, like the other two error-output checks, while the root images exist, after B-CHECKS-ROOT and before the `platform:reset` block. It uses the root project's built core runtime image, run with `docker run --rm --network none --label g001.run=$RUN_ID` and environment files under `$WORK` (mode 600), with per-run canaries:<br>• **CE1:** required variables missing; canary values in `DB_PASSWORD` and `REDIS_PASSWORD`. Expect exit 1, exactly one JSON line `core.exit`/`config_invalid` whose `config_keys` are in the allow-list, and no canary.<br>• **CE2:** otherwise valid configuration with `CORE_LISTEN_ADDR=bad-<canary>.invalid:8180`. Expect exit 1 and exactly one line `core.exit`/`listen_failed`, with no canary. P-GO shows that today's `Fprintln` leaks this host.<br>The output is classified by `logcheck startup`. |
| **B-WEB-ERROR-OUTPUT** (new) | root | A disposable `copy_repo`, with the two fixture routes copied into `app/api/__probe/{wrapped,control}/route.ts`. It is built with `web.Dockerfile --target runtime` under a run-labelled tag, and run with `--network none`, `PROBE_CANARY=<canary>`, harmless identity values, and `CORE_INTERNAL_URL=http://127.0.0.1:9`. `httpprobe` joins its namespace:<br>• **WE0:** only the allow-listed startup lines (§5.4) and JSON appear;<br>• **WE1:** `mode=throw` → 500, empty body, no-store, `X-Request-Id`, one `failed`/`Error` record;<br>• **WE2:** `nonerror` → `NonError`;<br>• **WE3:** `badname` → `Error`;<br>• **WE4 (positive control):** `control` → 500, and the canary **is** present in Next's raw output. Otherwise the check FAILs as vacuous.<br>• **WE5:** a malformed percent-encoded query carrying a canary on `wrapped?mode=ok` → no non-JSON line beyond WE0 and no canary;<br>• **WE6:** a client abort during `mode=slow` → one `cancelled` record and no framework line.<br>The canary is absent from everything except WE4's control output. The control output is deleted after the assertion and is never retained. |
| **B-INTEL-ERROR-OUTPUT** (new) | root | The root project's built intelligence runtime image, `--network none`:<br>• **IE1:** `APP_ENV` missing, plus a variable named `AI_MODEL_CANARY<hex>=openai:x`, whose *name* reaches the `ConfigError` message. Expect exit 3, every line JSON, exactly one `intelligence.startup.failed`/`ConfigError`, no canary, and no traceback.<br>• **IE2:** valid test configuration. Start, wait for `uvicorn.startup_complete`, then stop: every line is a catalogued JSON event, `uvicorn.*` templates are mapped, and there is no text line. |
| B-CHECKS-ROOT | root | Unchanged command; it now also runs the new unit, conformance and error-output tests |
| **B-OBS-CONTEXT, B-OBS-CHAIN, B-OBS-TIMEOUTS, B-OBS-LOGS, B-OBS-PRIVACY** (new) | variants (`variants_ops`, before B-OUTAGE) | §16 AC1–AC4 on the enabled disposable copy, as in r1. B-OBS-LOGS uses the r2 event catalogue. |
| **B-OBS-OFF** (new) | variants (`variants_off`, before B-BROWSER-OFF) | As in r1 |
| B-TOGGLE-OFF-ADDITIONS | variants | `APPLICABLE` adds the r1 six (`core/internal/observability`, `core/internal/health`, `web/app/api`, `web/lib/core-client.ts`, `web/lib/request-context.ts`, `docs/foundation/OPERATIONS.md`) **and `core/cmd/server/main.go`**: 28 → **35** |
| B-SECRET-CANARY | final | Adds the G003 canaries (§16 AC4) to the scan list. With `--secrets-file`, the list lives at that external path, which is outside `$WORK` and so needs no exclusion. Canary request specifications live in `$WORK/obs-inputs/` (excluded like `secrets*`) and are deleted after use. |
| ~~B-ARTIFACT-SCAN~~ (r1 proposal, **withdrawn**) | — | Artifact scanning has a single implementation: `validate.sh`'s raw and staged scans (§11.2 F3 and F7), where retention and upload happen. This resolves the r1 raw-versus-staged mismatch. Generated secrets and canaries inside the harness's own lifetime remain covered by B-SECRET-CANARY. |
| **B-REQUIRED-CHECKS** (new) | final | Every ID in `infra/validation/required-checks.tsv` for the executed suites is present exactly once and PASS. Missing, duplicate, FAIL or UNMET → FAIL. |
| `all` | — | Becomes `static provenance contracts root migrations variants`. Required IDs: 50 existing + **19** new = **69** (B-ARTIFACT-SCAN withdrawn; B-CORE-, B-WEB- and B-INTEL-ERROR-OUTPUT added). |
| **run.sh options (new, optional)** | — | Defaults keep today's behaviour; `run-approved.sh` is unchanged.<br>• `--owned-registry FILE`: the harness appends `work_root`, `label g001.run=<RUN_ID>`, `project <name>` (from `register_project`), `image <tag>` (every tag it builds with `-t`) and `rootenv <sha256>` rows (§11.4).<br>• `--secrets-file FILE`: `SECRETS_FILE` lives there instead of `$WORK/secrets`. It is appended at each generation, so it is current even if the harness is killed. |
| report copies | — | The end-of-run `--report-dir` copy list adds `obs-*.tsv`, `obs-timings.tsv`, `contracts-results-{go,python}.json`, `contracts-agree.tsv`, `schema-negative-{go,python}.json`, `error-output-{core,web,intelligence}.tsv` (statuses and categories only, never captured output) and `ci-controls.log`. These are the names `validate.sh` F4 and F5 expect. |
| cleanup | — | Also removes the run's `g003-*` images. The inventory diff is unchanged. The accepted `docker compose -p … down` backstop stays as it is, and CI cleanup does not rely on it (§11.4). |

**Unchanged:** B-SCOPE, B-PROTECTED, B-SKILL-UNCHANGED, B-LITERALS (which covers the new files: no identity literal, GitLab path or runner tag in `infra/` or `docs/foundation/`), B-NAMING, the migration suite, B-PORTS-*, B-BOUNDARIES-* and their controls, B-PERSIST, B-RENAME, B-BROWSER-RENAME, B-DBRENAME, B-SOURCE-HASH, B-DOCKER-ONLY, B-WEB-DESIGN, both browser matrices (95 tests each), `run-approved.sh` and its controls.

## 11. GitLab CI (R6/R7; r2; r1 §11 in Appendix R1)

### 11.1 `.gitlab-ci.yml` (exact intended content; `<RUNNER_TAG>` comes from the owner's Q1 facts and is written at S6)

```yaml
# Baseline CI (docs/foundation/OPERATIONS.md). One gating job runs the repository's containerized
# validation entry point. There is no publish, release or deploy job, and no CI/CD secret is used.
workflow:
  rules:
    - if: $CI_PIPELINE_SOURCE == "merge_request_event"
    - if: $CI_PIPELINE_SOURCE == "push" && $CI_COMMIT_BRANCH == "staging"
    - if: $CI_PIPELINE_SOURCE == "push" && $CI_COMMIT_BRANCH == $CI_DEFAULT_BRANCH
    - when: never

stages: [validate]

validate:
  stage: validate
  tags: [<RUNNER_TAG>]
  timeout: 3h
  resource_group: platform-validation-host
  interruptible: false
  variables:
    GIT_STRATEGY: clone
    GIT_CLEAN_FLAGS: -ffdx
    RUNNER_SCRIPT_TIMEOUT: 2h45m
    RUNNER_AFTER_SCRIPT_TIMEOUT: 10m
  script:
    - bash infra/ci/validate.sh --publish-dir ci-evidence
  after_script:
    - bash infra/ci/cleanup-owned.sh --after-script
  artifacts:
    when: always
    expire_in: 90 days
    access: developer
    paths: [ci-evidence/]
```

**r2 changes from r1:**
- the two `RUNNER_*_TIMEOUT` variables;
- the `--after-script` mode of `cleanup-owned.sh`.

Why (F14):
- Their total, 2h55m, fits within `timeout: 3h`, so that on runners that support these variables, `after_script` still runs if the script times out.
- The internal deadline in `validate.sh` (§11.2) is the primary mechanism and does not depend on the runner version (F-R4).
- `after_script` is a non-gating backstop: it cannot fail a successful job, and its output is informational.
- `artifacts:when: always` does not upload on job timeout. The internal deadline therefore ends the harness early enough for the finalizer to scan and publish within the job.

**B-CI-CONFIG asserts:**
- **Structure.** The top-level keys are exactly `workflow`, `stages` and `validate`. The four workflow rules are exact. There is exactly one job.
- **No bypass.** No `allow_failure`, `when: manual` or `when: delayed`, no job-level `rules` or `only`/`except`, no `rules:changes`, `needs`, `retry`, `parallel`, `trigger`, `environment`, `release`, `pages`, `image` or `services`.
- **Script.** It is exactly the single line above, and equal to the README `platform:ci` command.
- **`after_script`.** It is exactly `bash infra/ci/cleanup-owned.sh --after-script`.
- **Artifacts.** Exactly as shown: `when: always`, one directory, `access: developer`, `expire_in: 90 days`, no `untracked`.
- **Variables.** Exactly the four shown. No secret-like names. `RUNNER_SCRIPT_TIMEOUT` plus `RUNNER_AFTER_SCRIPT_TIMEOUT` is at most `timeout`, and `timeout` is at most 3h.
- **Tags.** Exactly one tag matching `^[A-Za-z0-9._-]{1,64}$`. Its value is the owner-supplied tag; Codex checks it against Q1, and hosted evidence records `CI_RUNNER_TAGS`. It is not repeated in `docs/` or `infra/`, so a tag containing an identity literal cannot trip B-LITERALS.

**Pipeline behaviour** (as in r1, with r2 attribution in §11.2):
- Merge-request pipelines test the source content (`detached`), or the merged result when the project has merged results enabled (Premium/Ultimate, F13; owner fact F-R10).
- Pushes to `staging` and to the default branch (`main`) also run.
- Feature-branch pushes, tags, schedules, web, API and trigger pipelines run nothing. If such a pipeline ever reached the job, attribution would reject it with 16.
- A push to `staging` while an MR from `staging` is open yields two gating pipelines for the same commit. This is accepted.
- `resource_group` serializes this project's validation jobs on the shared host ports.

### 11.2 `infra/ci/validate.sh` and `infra/ci/lib.sh` (aggregate entry point; Bash ≥ 4.4)

```text
validate.sh [--evidence-dir DIR] [--publish-dir ci-evidence]
```

#### 11.2.1 Execution model (correction 3)

- **Strict mode.** `set -euo pipefail` and `shopt -s inherit_errexit`.
- **Phases run isolated.** Every phase runs through `phase NAME FUNC`, which does `set +e; ( set -e; FUNC ); rc=$?; set -e` and records `ST[NAME]=rc`.
  - A phase is **never** called as `FUNC || …`. Bash ignores `errexit` inside a function or subshell called in an `||`/`&&`/`if` context, even after `set -e` inside it. Probe P-FIN-0 (§15.5) shows that `f || rc=$?` and `( set -e; f ) || …` both continue past a failing command (rc 0), while the isolated form returns 1.
- **The run starts at the trap.** Right after argument parsing, `trap finalize EXIT` and `trap 'on_signal TERM' TERM` are installed, likewise for `INT` and `HUP`. Every later exit, normal, error or signal, goes through `finalize` exactly once (guarded by a flag).
- **The finalizer is fail-open in control flow only.** `finalize` runs with `set +eu`, and every step inside it is itself a `phase`. A failing step is recorded and the next one still runs. No step's success can clear another's failure.
- **The harness runs in the background so signals are honoured.** The harness phase starts `timeout --signal=TERM --kill-after=120s "$BUDGET"s bash infra/validation/run.sh …` in the background and `wait`s for it. On `TERM`, `INT` or `HUP`, `on_signal` sends `TERM` to the `timeout` process group, records the signal, and `exit 143` (or 130 or 129) enters `finalize`.
  - Probe P-TERM shows that `timeout` stops a Bash child and that the child's own EXIT trap still runs. That is the harness's existing cleanup.
- **Internal deadline.**
  - `BUDGET = min(CI_JOB_TIMEOUT, RUNNER_SCRIPT_TIMEOUT) − 1200 s`. `CI_JOB_TIMEOUT` is a documented predefined variable, in seconds; `RUNNER_SCRIPT_TIMEOUT` is parsed from its Go duration.
  - Locally, with neither variable set, the budget is 4 h.
  - The preflight fails with 16 if `BUDGET < 3600 s`.
  - The 20-minute reserve covers finalization (scans, staging, cleanup and publication).

#### 11.2.2 Main body (in order; a failing preflight skips the harness)

| Phase | Work | Status on failure |
|---|---|---|
| P0 `location` | Resolve `--evidence-dir` (default `mktemp -d` under `TMPDIR`) with `realpath -m`, and reject the repository and its descendants **before creating anything**, as in the S0 gate.<br>`--publish-dir` accepts only `ci-evidence`.<br>Create the registry directory with mode 700 (§11.4): in CI, `$(dirname "$CI_PROJECT_DIR")/.platform-validation-owned-$CI_JOB_ID`; locally, `$EV/owned`. Write `job`, `token` (32 random hex) and `repo` rows. | 16 |
| P1 `preflight` | • `bash infra/ci/runner-preflight.sh --registry "$REG" --out "$EV/raw/runner-preflight.tsv"` (§11.5);<br>• `docker pull zricethezav/gitleaks:v8.30.1` (setup; scans later run `--pull never`);<br>• the Docker start inventory with IDs (`$REG/start-inventory.ids`);<br>• **attribution** (§11.2.4) to `$EV/raw/attribution.json`;<br>• **precheck:** `git --no-optional-locks status --porcelain=v1 -z --untracked-files=all --ignored` must be empty, and `HEAD`, `HEAD^{tree}` and a digest of the NUL-separated status are recorded;<br>• the budget (§11.2.1). | 16 |
| P2 `harness` | Only if P0 and P1 passed:<br>`timeout --signal=TERM --kill-after=120s "$BUDGET"s bash infra/validation/run.sh --base HEAD --allow-file infra/ci/unchanged.allow --owned-registry "$REG/owned.tsv" --secrets-file "$REG/secrets" --report-dir "$EV/raw/report" all >"$EV/raw/harness-console.log" 2>&1`<br>The status is recorded **as returned** (1 for failed checks, 2 for usage, 124 or 137 at the deadline). The console goes **only** to the raw capture, never to the job log. | the harness status |

`unchanged.allow` contains comments only, so B-SCOPE and B-PROTECTED assert that the run changed nothing relative to the checked-out commit. Goal scope is enforced separately, by the Goal's own final wrapper run.

#### 11.2.3 Finalizer (always; each step a `phase`; order fixed)

| Step | Work | Status on failure |
|---|---|---|
| F1 `postcheck` | The repository is observed first, before anything else touches the disk. The NUL-separated `status --ignored` must still be **empty**, and `HEAD`, `HEAD^{tree}` and the status digest must equal P1's whenever P1 recorded them. A leftover `.env` or report file is drift. | 13 |
| F2 `cleanup` | `bash infra/ci/cleanup-owned.sh --registry "$REG" --mode finalize --out "$EV/raw/cleanup-inventory.tsv"` (§11.4). Exit 0 only if every owned resource is gone, nothing was refused and every Docker command succeeded. | 18 |
| F3 `raw_scan` | Over **all** of `$EV/raw`: retained and excluded reports, the harness console, preflight and attribution output.<br>• (a) Literal scan: `grep -rlF -f "$REG/secrets" "$EV/raw"`. Exit 0 → finding (15); 1 → clean; ≥ 2 → scanner failure (17). If the harness ran and the secrets file is missing or empty → 17.<br>• (b) `docker run --rm --pull never --network none -v "$EV/raw:/scan:ro" -v "$PWD/infra/security/gitleaks.toml:/cfg/gitleaks.toml:ro" -v "$EV/scan:/out" zricethezav/gitleaks:v8.30.1 dir /scan --config /cfg/gitleaks.toml --redact --no-banner --exit-code 23 --report-format json --report-path /out/raw.json`. Exit 0 → clean; 23 → finding (15); anything else, or a missing or unparseable report → 17 (F22).<br>• (c) Forbidden names: `.env`, `*.env`, `secrets*`, `exec-token`, `it.env`, `static.env` → finding (15). | 15 or 17 |
| F4 `complete` | Evaluated whenever P2 ran, and reported even if P2 failed. Any parse or tool error is 14.<br>• **`results.tsv`:** every line has exactly three tab-separated fields; IDs match `^[A-Z][A-Z0-9-]*$`; status is in {PASS, FAIL, UNMET, INFO}. Every ID in `required-checks.tsv` occurs exactly once, as PASS. No FAIL or UNMET appears anywhere.<br>• **Browser JSON ×3** (parsed with yq): `expected > 0`, `skipped = 0`, `flaky = 0`, `unexpected = 0`.<br>• **Contract and schema-negative JSON:** `skipped: 0` and every `match: true`.<br>• **`inventory-leftover.txt`:** empty, apart from the documented renderer cache volume. | 14 |
| F5 `stage` | Only if F3 found nothing. Copy the allow-listed names from `$EV/raw` into `$EV/staged`, stripping `attachments` from the browser JSON with yq. The allow-list is r1's list (Appendix R1, §11.2 step 6) plus `harness-console.log`, `attribution.json`, `runner-preflight.tsv`, `cleanup-inventory.tsv` and `schema-negative-{go,python}.json`.<br>If F3 found something or failed, **withhold**: the staged set becomes only a `WITHHELD` marker and a minimal summary. | 19 |
| F6 `summary` | Write `summary.json` and `phase-status.tsv` into the set to be published. Only fixed field names, enums, integers, hex SHAs, UTC times and allow-listed CI metadata, never free text. | 19 |
| F7 `staged_scan` | Scanners (a)–(c) over **exactly** the set to be published, whether the full staged set or the withheld minimal set. Exit codes as in F3. | 15 or 17 |
| F8 `publish` | CI only. If F3 and F7 are both clean, copy the staged set to `$CI_PROJECT_DIR/ci-evidence.tmp` and then `mv` it into place, so publication is atomic. If F3 found something and F7 is clean, publish the withheld minimal set. **If any scan step returned 17, publish nothing**: no fallback summary is uploaded. A failure is 19. | 19 |
| F9 `scrub` | Delete `$REG/secrets` always. In CI, delete `$EV/raw` and `$EV/staged`. Delete the registry directory only if F2 returned 0; otherwise leave it for `after_script` and report that. | 19 |
| F10 `aggregate` | Compute the exit status (§11.2.5). The job log prints the `phase=… status=… utc=…` lines, then, **only if F3 was clean**, the `ID<TAB>STATUS` columns of `results.tsv` (never messages), then the final summary line. | — |

#### 11.2.4 Checkout attribution (correction 2; documented variables only, F13)

The classification runs in P1, **before** the harness, and depends only on the environment and the checked-out commit object. The prototype in §15.5 (P-ATTR) implements exactly these rules.

| Profile | Required (missing → 16) | Contradictory (present → 16) | Recorded `tested_content` |
|---|---|---|---|
| any CI run | `GITLAB_CI=true`; `CI_COMMIT_SHA` is 40-hex and **equals the checked-out `HEAD`**; `CI_PIPELINE_SOURCE` is `push` or `merge_request_event` | `CI_COMMIT_TAG` set; any other `CI_PIPELINE_SOURCE` value (the [documented](https://docs.gitlab.com/ci/jobs/job_rules/) `schedule`, `web`, `api`, `trigger`, `pipeline`, `parent_pipeline`, …) | — |
| branch push | `CI_PIPELINE_SOURCE=push`; `CI_COMMIT_BRANCH` set and equal to `staging` or `CI_DEFAULT_BRANCH`; `CI_COMMIT_REF_NAME == CI_COMMIT_BRANCH` | any of `CI_MERGE_REQUEST_IID`, `…_EVENT_TYPE`, `…_SOURCE_BRANCH_SHA`, `…_TARGET_BRANCH_SHA` | `branch-push` (commit, branch) |
| detached MR (source content) | `CI_PIPELINE_SOURCE=merge_request_event`; `CI_MERGE_REQUEST_IID` a positive integer; `CI_MERGE_REQUEST_SOURCE_BRANCH_NAME` and `…_TARGET_BRANCH_NAME` set; `CI_MERGE_REQUEST_EVENT_TYPE=detached` | `CI_MERGE_REQUEST_SOURCE_BRANCH_SHA` or `…_TARGET_BRANCH_SHA` non-empty (documented empty in MR pipelines); `CI_COMMIT_BRANCH` set (documented "Not available in merge request pipelines") | `mr-source` (commit = source content, IID, branch names; `CI_MERGE_REQUEST_DIFF_BASE_SHA` recorded if present, not asserted) |
| merged result | As detached, but `EVENT_TYPE=merged_result`, and `…_SOURCE_BRANCH_SHA` and `…_TARGET_BRANCH_SHA` both 40-hex (documented present only here) | `HEAD` equal to either SHA (documented: the merged commit "doesn't exist in either branch"); `HEAD`'s raw commit object not having exactly those two SHAs as parents | `mr-merged-result` (merge commit, source SHA, target SHA, IID) |
| merge train | — | `EVENT_TYPE=merge_train` → 16: unsupported profile (Q17) | — |
| local | `GITLAB_CI` unset | any of `CI_PIPELINE_SOURCE`, `CI_COMMIT_SHA`, `CI_JOB_ID`, `CI_MERGE_REQUEST_IID`, `CI_MERGE_REQUEST_EVENT_TYPE`, `CI_COMMIT_BRANCH` set | `local` |

- **Parents.** The merged-result parent check reads the raw object (`git cat-file commit HEAD`). GitLab clones are shallow (default `GIT_DEPTH` 20, [runner docs](https://docs.gitlab.com/ci/runners/configure_runners/)). Probe P-ATTR-0 shows that a depth-1 fetch of a merge commit still lists both parent SHAs in the raw object, while `git rev-parse HEAD^2` fails at the shallow boundary.
- **The parent-set rule is a repository-side consistency check, not a GitLab-documented guarantee.** If a real merged-results commit had another structure, the run fails closed with 16 rather than claiming a merged result (risk R13).
- **Recorded metadata** (allow-listed; never an environment dump): the fields above, plus `CI_PIPELINE_ID`, `CI_PIPELINE_IID`, `CI_PIPELINE_URL`, `CI_JOB_ID`, `CI_JOB_URL`, `CI_DEFAULT_BRANCH`, `CI_MERGE_REQUEST_APPROVED` (recorded only), `GITLAB_USER_LOGIN`, `CI_RUNNER_ID`, `CI_RUNNER_DESCRIPTION`, `CI_RUNNER_TAGS`, `CI_SERVER_VERSION`, `CI_JOB_TIMEOUT`, `git log -1 --format='%H %an <%ae> %cI'`, `HEAD^{tree}`, Docker and Compose versions, and UTC start and end times.
- **No overclaiming.** A merged-result run is never implied when only source content ran: the only path to `mr-merged-result` is the row above.

#### 11.2.5 Exit status (aggregation; no masking)

| Code | Meaning |
|---|---|
| 0 | P0–P2 and every finalizer step returned 0, **and** the main body completed |
| harness status | P2 ran and returned nonzero: 1 or 2 from `run.sh`, 124 or 137 at the internal deadline, 143, 130 or 129 when the job was interrupted |
| 13 | postcheck drift |
| 14 | evidence incomplete or malformed |
| 15 | unsafe evidence found (withheld) |
| 16 | preflight: location, runner preflight, attribution, precheck, budget or usage |
| 17 | scanner failure (scan incomplete; nothing published) |
| 18 | owned-resource cleanup incomplete or refused |
| 19 | internal: a staging, summary, publication or scrub failure, a missing phase status, or an unexpected exit outside a phase |

- **Precedence** chooses which code is reported; every nonzero status is still printed in `phase-status.tsv` and the job log:
  1. 16, when P2 never ran;
  2. otherwise the nonzero harness status;
  3. otherwise 17, 15, 14, 13, 18, 19.
- **No masking.** No combination of successful steps can produce 0 while any recorded status is nonzero or the main body did not complete. The model probe P-FIN (§15.5) exercises 11 fault cases.
- **Usage errors** exit 16 (r1 used 2, which collided with the harness's own usage code).

### 11.3 CI controls (`infra/ci/validate-controls.sh`; scratch trees outside the checkout; no switch in `validate.sh`)

**How the controls work.**
- Each control builds a disposable `git init` tree in `mktemp -d`, containing copies of `validate.sh`, `lib.sh`, `cleanup-owned.sh`, `required-checks.tsv` and `unchanged.allow`, a **stub** `infra/validation/run.sh`, a **stub** `infra/ci/runner-preflight.sh` (PASS by default) and one tracked input. It is committed with a fixed throwaway identity.
- The real `runner-preflight.sh` runs in every real local (S8) and hosted (R1, H0–H2) run, and is shellchecked by B-SHELL.
- The stub records each invocation in a marker file. It writes `results.tsv`, reports and console output as instructed, and can mutate the tree, sleep or exit with a given code.
- **Faults are injected only through the scratch environment:**
  - the stub harness;
  - a `docker` shim prepended to `PATH`, which passes through except for configured patterns (for example, a failing gitleaks invocation or a failing `docker rm`). It treats `docker pull` as a no-op, because the pinned images are already present when the static suite runs;
  - file modes;
  - environment variables that GitLab itself defines (for attribution);
  - signals.
- **`validate.sh` has no fault-injection switch** (R7), and B-CI-CONFIG asserts that none of these variables is set in the pipeline.
- Each control asserts its exit status, the stub invocation count, the summary fields, and the published set.

**Group A: gate semantics** (r1 V1–V4 and V7–V12, retained):

| Control | Scenario | Expected |
|---|---|---|
| V1 | All valid, complete PASS | 0; the staged set is published and scanned |
| V2 **(a)** | One service-check FAIL row (B-CHECKS-ROOT), harness exit 1 | 1; the summary names it; artifacts are staged and scanned |
| V3 (a) | Harness exit 0 but a FAIL row present | 14 |
| V4 **(b)** | **Real tooling:** a scratch copy of the implementation (`copy_repo`, `git init`, a scratch commit) with EI39 set to `valid`. The real `run.sh … contracts` exits 1 with B-CONTRACT-GO, -PY and -AGREE FAIL; that run's `results.tsv` is replayed through V2's path. | 1, for that reason |
| V7 **(d)** | Harness exit 1, clean tree | 1; postcheck passed and reported |
| V8 **(e)** | Harness exit 0, a tracked file modified | 13 |
| V9 (e) | Harness exit 0, an untracked file created | 13 |
| V10 **(f)** | Exit 0, one required ID missing | 14 |
| V11 (f) | Exit 0, an UNMET row present; or browser JSON with `skipped: 1` | 14 (two sub-cases) |
| V12 | Harness exit 1 **and** drift | 1, with the drift reported |

**Group B: attribution** (correction 2). It runs `validate.sh` with the GitLab-defined variables set as below, over scratch fixture commits: a target commit T, a source commit S, and a merge commit M with parents T and S, fetched at depth 1. The stub harness passes unless stated otherwise.

| Control | Profile | Expected |
|---|---|---|
| P1 | push to `staging`; `CI_COMMIT_SHA = HEAD = T` | 0; `tested_content = branch-push`, branch `staging` |
| P2 | push to the default branch (`main`) | 0; `branch-push`, branch `main` |
| P3 | `merge_request_event`, `detached`, `CI_COMMIT_SHA = HEAD = S`, both branch SHAs empty | 0; `mr-source`; IID recorded |
| P3F | P3, with the stub harness exiting 1 | 1; `tested_content` is still `mr-source`, so the classification does not depend on the harness result |
| P4 | `merged_result`, `HEAD = M`, source S, target T | 0; `mr-merged-result` with both SHAs |
| P5 | no GitLab variables | 0; `local` |
| N1–N16 | N1 no pipeline source · N2 `CI_COMMIT_SHA ≠ HEAD` · N3 `CI_COMMIT_SHA` missing · N4 detached with a source SHA · N5 merged result without the source SHA · N6 merged result with `HEAD` = source · N7 merged result whose parents differ from the SHAs · N8 push with an MR IID · N9 push to a feature branch · N10 MR pipeline with `CI_COMMIT_BRANCH` · N11 `merge_train` · N12 `schedule` · N13 local run carrying `CI_COMMIT_SHA` · N14 tag pipeline · N15 MR without an event type · N16 MR without an IID | each: **16**; stub harness invocations **0**; F1 and F2 recorded; only the scanned preflight evidence (`runner-preflight.tsv`, `attribution.json` when written) and the summary are published |

**Group C: evidence scanning** (correction 3; raw versus staged):

| Control | Scenario | Expected |
|---|---|---|
| V5a | A generated canary (in the stub's secrets file) in a **retained** raw report | 15; published = `WITHHELD` and a minimal summary; the canary is absent from the published directory |
| V5b | The same canary only in an **excluded** raw file (`harness-console.log`, never staged) | 15; withheld as V5a. This proves the raw scan covers files that are never staged. |
| V5c | A gitleaks-detectable token generated at run time, not in the secrets file, in a raw report | 15 (the gitleaks path, not only the literal list) |
| V6a | Scanner failure: the `docker` shim makes the gitleaks run exit 1 (gitleaks' fatal-error code, F22) | 17; **nothing** published; no unscanned summary |
| V6b | The harness ran, but the secrets file is missing | 17; nothing published |
| V6c | Positive scanner control: clean raw and staged sets | 0 (identical to V1; recorded explicitly) |

**Group D: finalization paths** (correction 3):

| Control | Scenario | Expected |
|---|---|---|
| V14 | Malformed `results.tsv` (a four-field line) | 14; F1, F2, F3 and F7 recorded |
| V15 | Malformed browser JSON | 14 |
| V16 | Staging failure: an allow-listed raw file is unreadable (mode 000) | 19; nothing partial published |
| V17 | Publication failure: the publish parent is not writable | 19; no `ci-evidence/` directory left behind |
| V18 | `SIGTERM` sent to `validate.sh` while the stub harness sleeps | 143; the stub is terminated; F1–F9 recorded; nothing unscanned published |
| V18b | Harness exit 1 and cleanup failure (shim) | 1; cleanup 18 reported; inventory not empty |
| V18c | The stub runner preflight reports a FAIL row | 16; stub harness invocations 0; F1 and F2 recorded |

**Group E: owned-resource cleanup** (correction 3; Docker-real, `busybox:1.37.0`, every resource labelled `g003.control=<token>`; the outer harness's inventory diff catches any leak):

| Control | Scenario | Expected |
|---|---|---|
| V19 | A container, network and volume carrying `com.docker.compose.project=<registered>` and `com.docker.compose.project.working_dir` under the registry's `work_root`, all created after the start inventory | `cleanup-owned.sh` removes all three; inventory `removed=3 remaining=0`; exit 0 |
| V20 | Unrelated: the same project label but a working directory outside `repo` and `work_root` | **refused** and left intact; exit 18; the control then removes it itself and verifies |
| V21 | Unrelated: a resource under an **unregistered** project name | untouched and absent from the inventory; exit 0 for the owned set |
| V22 | Removal failure: the shim makes `docker rm` exit 1 while the container still exists | 18; the inventory lists it as `remaining` and is never reported empty |
| V23 | Registry tampering: the registry directory is a symlink; the `token` row and file disagree; or the job does not match `CI_JOB_ID` | refused; nothing removed; exit 18 (`--after-script`: exit 0, with the refusal printed) |
| V24 | `after_script` hand-off: a separate `env -i bash` with only `PATH`, `CI_PROJECT_DIR` and `CI_JOB_ID` runs `cleanup-owned.sh --after-script` | It finds the registry by the documented rule (§11.4), removes the leftovers, prints the inventory and exits 0. With no registry it prints `cleanup: no registry (finalization completed)` and exits 0. |
| V25 | Root `.env` with a matching registered hash, then a mismatched one | the matching file is removed; the mismatched one is kept and reported `refused` |
| V26 | A resource present in the start inventory, carrying a registered project label | untouched (pre-existing), never removed |

No control can turn a failure into exit 0. W1–W9 (`run-approved.sh`) and C1–C12, C8b and C8c remain separate and unchanged.

### 11.4 Owned-resource registry and `infra/ci/cleanup-owned.sh` (correction 3)

**Location: job-unique, outside the checkout, and discoverable by `after_script`.**
- **CI:** `REG = "$(dirname "$CI_PROJECT_DIR")/.platform-validation-owned-$CI_JOB_ID"`. Both inputs are [documented](https://docs.gitlab.com/ci/variables/predefined_variables/) predefined variables, available to `after_script`'s separate shell. `CI_JOB_ID` is "unique across all jobs in the GitLab instance".
  - It is a sibling of the checkout, so it is not drift and is not removed by `git clean` of the checkout.
  - It is on the same host for the shell executor, and on the identically mounted builds path for Profile A2 (§12.1).
- **Local:** `$EV/owned`. There is no `after_script` locally; the finalizer is the only path.
- **Permissions:** directory mode 700, owned by the job user, never a symlink. Files: `owned.tsv` (append-only rows), `start-inventory.ids`, `secrets` (mode 600), `token`.

**Rows** (`owned.tsv`; tab-separated; written by `validate.sh` and by `run.sh --owned-registry`):

```text
#platform-validation-owned v1
job        <CI_JOB_ID | local-<random>>
token      <32 hex; equal to the token file>
repo       <absolute checkout path>
work_root  <the harness's absolute $WORK>
label      g001.run=<RUN_ID>
label      platform.validation.preflight=<token>
project    <compose project name>          (one row per register_project)
image      <repository:tag>                (one row per image built with -t)
rootenv    <sha256 of the root .env the harness created>
```

**Selection and binding** (the exact commands; no Compose file or environment is needed):
- **Containers.**
  - `docker ps -aq --no-trunc --filter "label=com.docker.compose.project=$P"` for each `project` row, and `docker ps -aq --no-trunc --filter "label=$L"` for each `label` row.
  - A container is **owned** only if its ID is absent from `start-inventory.ids` **and**, for a Compose project, `docker inspect -f '{{index .Config.Labels "com.docker.compose.project.working_dir"}}' ID` equals `repo` or lies under `work_root`. Otherwise it is **refused**.
  - Removal: `docker rm -f -v ID`.
- **Networks.** `docker network ls -q --no-trunc --filter "label=com.docker.compose.project=$P"`; owned if absent from the start inventory; `docker network rm ID`, after the containers.
- **Volumes.** `docker volume ls -q --filter "label=com.docker.compose.project=$P"`; owned if absent from the start inventory; `docker volume rm NAME`.
- **Images.**
  - Each `image` row whose `docker image inspect -f '{{.Id}}' TAG` is absent from the start inventory: `docker image rm TAG`.
  - Images named `<project>-<service>` under a registered project and absent at start, as the harness cleanup already does.
- **Root `.env`.** Removed only if its SHA-256 equals the `rootenv` row; otherwise refused and reported.
- **Verification.** Every query is rerun after removal. Every owned resource still present is a `remaining` row.
  - `--mode finalize` exits 0 only if nothing remains, nothing was refused and every Docker command succeeded; otherwise 18.
  - It writes `cleanup-inventory.tsv` (`kind`, `id`, `action` ∈ removed/refused/remaining/failed, `reason`).
- **Never:** any `prune`; removal by name pattern alone; touching any resource in the start inventory; touching resources outside the registered projects and labels.

**Modes:**
- `--registry DIR --mode finalize --out FILE`: called by `validate.sh` F2.
- `--after-script`: called by `.gitlab-ci.yml`. It derives `REG` from `CI_PROJECT_DIR` and `CI_JOB_ID`.
  - If `REG` is absent (normal: F9 deleted it after a verified cleanup), it prints `cleanup: no registry (finalization completed)`.
  - Otherwise it applies the finalize logic, removes `secrets`, prints the inventory to the job log (names only), and deletes `REG` only if nothing remains.
  - It **always exits 0** and is documented as non-gating (F14). It is a backstop for a killed `validate.sh`, not a gating assertion.

**Relation to the harness's own cleanup.** The accepted `run.sh` cleanup trap (including its `docker compose -p "$p" down -v --remove-orphans` backstop) is unchanged and still runs first, inside the harness. `cleanup-owned.sh` independently verifies the result and removes only what is provably owned and still present. A successful run therefore cannot claim an empty inventory when a removal failed (V22).

### 11.5 `infra/ci/runner-preflight.sh` (new; used by `validate.sh` P1 and by R1)

It writes one row per check to its `--out` TSV, with status PASS, FAIL or INFO. Any FAIL makes P1 exit 16. Test containers are labelled `platform.validation.preflight=<token>` and registered.

| ID | Check | Requirement (Profile A, §12.1) |
|---|---|---|
| RP1 | Bash version | ≥ 4.4 |
| RP2 | Host tools | GNU `coreutils`, `findutils`, `grep -P`, `sed`, `awk`, `tar`, `timeout`, `sha256sum`, `realpath` and `git` present |
| RP3 | Docker daemon | reachable; server ≥ 24.0; `linux/amd64` |
| RP4 | Compose | ≥ 2.24 |
| RP5 | BuildKit | `docker buildx version` succeeds |
| RP6 | **Loopback parity** | `docker run -d -p 127.0.0.1::80 busybox:1.37.0 httpd -f -p 80` is reachable from the job shell at `/dev/tcp/127.0.0.1/<port>` |
| RP7 | **Bind-mount path parity** | A marker in a job temp directory, and a tracked file under `CI_PROJECT_DIR`, are readable through `-v "$p:$p:ro"` |
| RP8 | Non-loopback IPv4 | present (the harness's loopback-only proof needs it) |
| RP9 | Ports | default and variant port sets free on `127.0.0.1` |
| RP10 | Root project | no existing resources labelled with the root Compose project |
| RP11 | Capacity | `nproc` ≥ 4; MemTotal ≥ 8 GiB; free space ≥ 40 GiB on `TMPDIR` and on the Docker root directory (INFO if the latter is not readable) |
| RP12 | Time | `CI_JOB_TIMEOUT` ≥ 10800 s in CI; harness budget ≥ 3600 s |
| RP13 | INFO only | runner ID, description and tags, GitLab version, count of running containers not owned by this job |

## 12. Runner profile and hosted evidence (blocking prerequisites Q1 and Q2; R6/R7, AC7; r2; r1 §12 in Appendix R1)

### 12.1 What the harness requires (from F9)

**Profile A (required unless Codex and the owner decide otherwise).** A self-managed GitLab runner using the **shell executor**, on a **dedicated, non-production Linux host**:

- **Docker:** Docker Engine 24 or later with BuildKit, and the Compose v2 plugin 2.24 or later.
- **Host tools:** bash 4.4 or later (`inherit_errexit`), GNU coreutils/findutils/grep with `-P`, git, `timeout`, `hostname -I` or `ip`, and bash `/dev/tcp`.
- **Job user:** the job runs as the runner user in the `docker` group. That is root-equivalent on this host. It is acceptable only because the host is dedicated and disposable in purpose, and it must be acknowledged.
- **Isolation:** no other workload's Docker resources on that daemon.
- **Concurrency:** one concurrent job for this project. `resource_group` enforces it inside the project; the runner `limit`/`concurrent` value, or dedication, covers other projects.
- **Capacity:** at least 4 vCPU (8 recommended), at least 8 GiB RAM (16 recommended), and at least 40 GiB free for images and the BuildKit cache.
- **Timeouts:** the runner's `maximum_timeout` and the project's CI timeout must each be at least 3 h.
- **Egress** to the registries and module proxies used in builds: Docker Hub, `gcr.io`, `ghcr.io`, `mcr.microsoft.com`, `proxy.golang.org`/`sum.golang.org`, `registry.npmjs.org` and PyPI. The Docker Hub anonymous pull limit must be acknowledged.

**Profile A2 (acceptable only if already configured this way; no configuration change is made by G003).** A Docker executor that mounts the host Docker socket, with job-container `network_mode = "host"`, and the builds and temporary directories mounted at **identical host paths**. Without all three, the bind-mount and loopback probes cannot work.

**Not acceptable without an explicit security decision (stop condition).**
- GitLab.com hosted runners or any `docker:dind` service. These need `privileged = true`, and published ports would not be on the job's loopback.
- A socket-only Docker executor without host networking and path parity.
- Any runner that can reach production resources or credentials.

**r2 notes on §12.1.**
- **Nothing here describes the actual runner.** The profiles state requirements, not facts. No runner identity, tag, executor, capacity or permission is assumed anywhere in this Plan.
- **The registry location (§11.4) relies on the same path parity.** Profile A2 already requires the builds and temporary directories at identical host paths.
- **`runner-preflight.sh` (§11.5) turns each requirement into an executable check,** run in every pipeline and in R1.

### 12.2 Owner-supplied runner facts (Q1: **UNRESOLVED; blocking**)

The owner has confirmed only that an existing Docker-capable GitLab runner is to be used (Goal, Dependencies). Everything below is **unknown** and must be supplied by the owner, or by someone they authorize.

**Rules for supplying them:**
- Supply them as plain text in the Codex approval handoff, or in a file outside the repository whose path is given to Codex.
- Do **not** paste `config.toml` wholesale: it contains the runner token. Copy only the named fields.
- Codex checks the facts against Profile A or A2 **before approving execution**.
- R1 (§12.4) then verifies them on the runner **before S1**.
- Facts are never a substitute for R1, and R1 is never a substitute for the facts. If any fact is unavailable, execution stays blocked.

| ID | Fact the owner supplies | Where to read it | Requirement it is checked against | Status |
|---|---|---|---|---|
| F-R1 | Runner ID and description; scope (project, group or instance); self-managed or GitLab-hosted | GitLab: project → Settings → CI/CD → Runners | Self-managed (Profile A or A2). GitLab-hosted runners are a stop condition. | UNRESOLVED |
| F-R2 | The tag to use; whether it runs untagged jobs; whether it is limited to protected branches; whether it is locked to the project | same page | One tag for `.gitlab-ci.yml`. If protected-only, the evidence branches must be protected (owner action), or another runner is needed. | UNRESOLVED |
| F-R3 | **Executor** type, and for Docker executors: `privileged`, the Docker socket mount, `network_mode`, `volumes` and the builds path | runner host: the `[[runners]]` section of `config.toml` (named fields only) | Shell executor (A), or a Docker executor with socket, host network and identical paths (A2). `privileged = true` or dind is a stop condition. | UNRESOLVED |
| F-R4 | Runner version (`gitlab-runner --version`) | runner host | Records whether `RUNNER_SCRIPT_TIMEOUT` and `RUNNER_AFTER_SCRIPT_TIMEOUT` are honoured. The internal deadline does not depend on it (§11.2.1). | UNRESOLVED |
| F-R5 | Host OS and architecture; Docker Engine, Compose plugin and Bash versions; CPUs; RAM; free disk for the Docker root and the temporary directory | runner host (`uname -m`, `docker version`, `docker compose version`, `bash --version`, `nproc`, `free -g`, `df -h`) | linux/amd64; Engine ≥ 24; Compose ≥ 2.24; Bash ≥ 4.4; ≥ 4 vCPU, ≥ 8 GiB RAM, ≥ 40 GiB free | UNRESOLVED |
| F-R6 | Isolation: is the host dedicated? Other workloads or projects on the same Docker daemon; the runner's `concurrent` and `limit` | owner knowledge, `config.toml` | Dedicated, non-production, no production credentials or networks reachable; one concurrent job for this project | UNRESOLVED |
| F-R7 | Timeouts: the runner's `maximum_timeout`; the project's CI/CD general timeout | `config.toml`; project → Settings → CI/CD → General pipelines | Each ≥ 3 h (the effective `CI_JOB_TIMEOUT` is checked again by RP12) | UNRESOLVED |
| F-R8 | Network egress from the host and job; any proxy or pull-through cache; Docker Hub authentication or rate-limit status | owner and network knowledge | Registries listed in §12.1 reachable; the anonymous pull limit acknowledged | UNRESOLVED |
| F-R9 | Host port use: is anything bound to the default ports (3100, 8180, 8100, 5532, 6479, 9100, 9101) or to the variant sets (§6.3 and `run.sh` `PORTS_*`)? | runner host (`ss -ltn`) | All free | UNRESOLVED |
| F-R10 | GitLab tier of the namespace; whether merged-results pipelines are enabled; whether merge trains are enabled | group → Settings → Billing; project → Settings → Merge requests | Merged results are Premium/Ultimate (F13). Merge trains are unsupported, so attribution would return 16 (Q17). | UNRESOLVED |
| F-R11 | Artifacts: the instance maximum artifact size; the project's "Keep artifacts from most recent successful jobs" setting; who holds Developer or higher access | Admin or project settings; project → Members | `access: developer` and 90-day expiry acceptable (Q7); expected evidence size well below the maximum (measured at R1 and H0) | UNRESOLVED |
| F-R12 | Evidence retrieval mechanism (paired with A-E, §12.3) | owner decision | Owner download to a path outside the repository, or a short-lived `read_api` token supplied out of band | UNRESOLVED |
| F-R13 | Written acknowledgement that the job user's `docker` group membership is root-equivalent on that host | owner | Required for Profile A | UNRESOLVED |

### 12.3 Hosted actions requiring explicit authorization (Q2: **UNRESOLVED; blocking**)

Neither this Plan, nor its approval, nor the planning delegation authorizes any of these. Each needs the owner's or Codex's explicit, named authorization before it is taken.
- **Commits.** Creating the commits these actions push is part of each authorization. They use the repository's configured Git identity unless the authorization says otherwise.
- **Ancestors.** A push publishes every ancestor commit not yet on `origin`. If BASE includes a not-yet-pushed commit (for example the skill consolidation under route A, Q14), the first push publishes it on that branch. The authorization must cover this.

| ID | Action | Purpose | Undo |
|---|---|---|---|
| **A-R1** | From a disposable clone **outside the repository**: create branch `g003-runner-probe` at BASE plus **one** probe commit (§12.4), push it to `origin`, and let its branch pipeline run on the Q1 runner | Runner qualification before S1 | Delete the branch after the evidence is captured (part of A-R1) |
| A-H0 (optional, recommended) | Push the work-in-progress commit(s) to `g003-foundation-observability-ci` and open a **draft** MR to `staging` | Early run of the G003 job itself on the runner, after S8 | The MR stays open for H1 |
| **A-H1** | Push the final implementation commit, and the Result when available, to the same branch; its MR pipeline runs | Hosted pass on the exact submitted commit and tree | Kept for review |
| **A-H2** | Push branch `g003-gate-proof` = H1 plus one commit setting EI39 to `valid`; open a draft MR to `staging` | Hosted failure through the reviewed gate | A-H3 |
| **A-H3** | After Codex has captured H2: close its MR and delete `g003-gate-proof` | Clean-up | — |
| **A-E** | Evidence retrieval for R1, H0, H1 and H2 (logs, artifacts, pipeline and job metadata) through the F-R12 mechanism | Attributable hosted evidence | Downloads are kept outside the repository; a token is never printed or stored |

**No action pushes to `staging` or `main`, merges, publishes, deploys, or creates or changes CI/CD variables, runners or project settings.**

### 12.4 R1: runner qualification before S1 (correction 1; early compatibility and cold-build evidence)

**When.** Immediately after S0 passes, which needs Q9 and Q14 resolved, and the Plan is approved, and only with A-R1 and A-E authorized. It comes **before any repository write** (S1). It makes **no repository write**: everything happens in `$EXT/r1/probe`, a `git clone` of the local repository at BASE. Cloning a local repository reads its objects and writes nothing to it.

**Probe commit** (four files; they exist only in the disposable clone and on the deleted branch):
1. `.gitlab-ci.yml` (probe-only):

   ```yaml
   workflow:
     rules:
       - if: $CI_PIPELINE_SOURCE == "push" && $CI_COMMIT_BRANCH == "g003-runner-probe"
       - when: never
   runner-probe:
     tags: [<RUNNER_TAG>]
     timeout: 3h
     resource_group: platform-validation-host
     variables:
       GIT_STRATEGY: clone
       GIT_CLEAN_FLAGS: -ffdx
     script:
       - bash infra/ci/runner-probe-job.sh
   ```

2. `infra/ci/runner-preflight.sh`: the §11.5 script, as written for S6. Any later change is recorded in the Result.
3. `infra/ci/runner-probe-job.sh`. Under an EXIT trap, it:
   - runs the preflight, with its TSV printed to the job log (no secrets: tool versions, counts and PASS/FAIL);
   - records `docker system df` before and after, and the UTC start and end;
   - runs the **unchanged accepted baseline harness**: `timeout --signal=TERM --kill-after=120s 9900s bash infra/validation/run.sh --base HEAD --allow-file infra/ci/probe.allow --report-dir "$T/report" all >"$T/console.log" 2>&1`, with `$T` a `mktemp -d` outside the checkout;
   - prints only the `ID<TAB>STATUS` columns of `results.tsv`, the leftover-inventory **count** and the durations;
   - deletes `$T`;
   - exits with the harness status.

   It uploads no artifacts.
4. `infra/ci/probe.allow`: comments only. `--base HEAD` is the probe commit, so B-SCOPE and B-PROTECTED see no change.

**What R1 establishes (before any G003 cycle):**
- runner selection by tag;
- executor compatibility: loopback parity, bind-mount parity, port availability, non-loopback IPv4 (RP6–RP9, and the real B-PORTS-* and B-BOUNDARIES-*);
- cold image builds, including the MinIO source build and the pinned images, and registry egress;
- the browser matrices on the runner;
- the cold `all` duration against the time budget;
- disk use;
- the harness's own cleanup on that daemon.

**Decision rule** (reported to Codex before S1, and recorded in the Result):
- **Proceed to S1** only if:
  - all 50 baseline required checks PASS;
  - every RP check passes;
  - the leftover inventory is empty, apart from the renderer cache volume;
  - the cold `all` duration is **≤ 90 minutes**. The G003 harness budget is 145 minutes (2h45m − 20m), and §19 estimates that G003's contracts suite, error-output checks and observability matrix add about 25–35 minutes cold. 90 minutes keeps that estimate inside the budget with margin.
- **Otherwise stop before S1.** Return the evidence to Codex and the owner. A runner reconfiguration, a timeout change or a split (§19) is their decision. Because no implementation exists yet, **no B7 cycle is consumed**. A rerun after an owner-side change also consumes none.
- **If A-R1 is not authorized,** the B7 forecast is not credible. The Plan recommends not starting S1, and Codex decides (Q10, §19).

### 12.5 Hosted evidence plan (H0–H3; each needs its authorization from §12.3)

| Step | When | Expected evidence |
|---|---|---|
| H0 (optional, recommended) | After S8 (local CI parity passes) | The first run of the G003 job on the runner, cache warm from R1. It records the pipeline, job and runner IDs, `tested_content`, the duration, the artifact size and the published set. A failure caused by the G003 changes counts in B7; one caused by the runner environment is reported as in §12.4. |
| H1 | After S10 | A passing pipeline on the **exact submitted commit and tree**. Recorded: pipeline ID, job ID, runner, `tested_content` (`mr-source`, or `mr-merged-result` if enabled), the duration, `summary.json` with 69 PASS, and the downloaded artifact's SHA-256. |
| H2 | After H1 | A **failed** pipeline: B-CONTRACT-GO, B-CONTRACT-PY and B-CONTRACT-AGREE FAIL, and `validate.sh` exits 1. The scanned artifacts show the reason. This exercises the reviewed gate, not an `exit 1` job. |
| H3 | After Codex captures H2 | The H2 MR is closed and `g003-gate-proof` is deleted; recorded |

**Evidence classification** is unchanged from r1. Local and hosted evidence are kept apart in the Result, and **local parity never substitutes for H1 or H2**. R1 is runner-qualification evidence for the unchanged baseline, not G003 acceptance evidence.

## 13. Dependencies, images and provenance

| Item | Pin | Where | Notes |
|---|---|---|---|
| `github.com/santhosh-tekuri/jsonschema/v6` | `v6.0.3` (2026-06-28, Apache-2.0) | `infra/validation/tools/go.mod` (new) | Test tooling only. `go.sum` is recorded; `go mod verify` runs in every build. |
| `golang.org/x/text` | resolved by MVS (expected `v0.14.0`) | tools `go.mod`, indirect | Recorded in the Result |
| `golang:1.27.1-alpine`, `gcr.io/distroless/static-debian12:nonroot` | existing tags | `validation-tools.Dockerfile` | Same tag-only pins as the existing images (inherited residual). No new base image. |
| `zricethezav/gitleaks:v8.30.1` | existing | `validate.sh` artifact scan | Already used by `secret-scan` |
| `koalaman/shellcheck:v0.11.0`, `hadolint/hadolint:v2.15.1`, `mikefarah/yq:4.48.1`, `busybox:1.37.0`, Playwright `v1.63.0-noble@sha256:eff16c30…a4a27` | existing | harness | Unchanged |

**No change** to:
- web `package.json` dependencies or `package-lock.json`;
- core `go.mod`/`go.sum`;
- intelligence `pyproject.toml` dependencies or `uv.lock`;
- any runtime base image.

**No new vendor, service or infrastructure.** santhosh is a validation library used by a test tool, analogous to the Playwright image accepted in G002's Q4 (Q12).

## 14. Files (exact; `management/v0/results/G003-scope-allowlist.txt` mirrors this section; r2; r1 §14 in Appendix R1)

### 14.1 Database, schema and contracts

- **None** of: migrations, seeds, tables, domain identities, public routes, API or MCP contracts. B-MIGRATE-SET stays `{00001_baseline.sql}`.
- The two schemas are byte-identical: both validators assert the §2 digests, and the SD controls prove that enforcement.
- Response headers are additive only:
  - `X-Request-Id` on the web and intelligence health routes;
  - `Cache-Control: no-store` on intelligence health;
  - on an unexpected web health-handler exception, the 500 with an empty body that Next already sends, plus those two headers (Q15).

### 14.2 Customized rendered files (11 new register rows → 19 total)

The r1 ten, plus `core/cmd/server/main.go`:

| Path | Register reason |
|---|---|
| `core/internal/observability/correlation.go` | diagnostic context profile, per-hop spans and completion records (G003) |
| `core/internal/observability/logging.go` | JSON record envelope (event, service, UTC time) on the redacting logger (G003) |
| `core/internal/health/checks.go` | propagates diagnostic context on the optional intelligence probe (G003) |
| **`core/cmd/server/main.go`** | **bounded JSON exit record and classified net/http error log (G003)** |
| `web/app/api/health/route.ts` | request context and completion record on liveness (G003) |
| `web/app/api/ready/route.ts` | request context, propagation and completion record on readiness (G003) |
| `web/lib/core-client.ts` | forwards diagnostic context to core readiness (G003) |
| `intelligence/src/main.py` | health completion middleware, no-store health responses and bounded startup failure (G003) |
| `intelligence/src/observability/export.py` | JSON envelope, redaction, bounded framework, library and warning records, and framework-logger takeover (G003) |
| `intelligence/src/streaming/schemas.py` | Draft 2020-12 conformance: ECMA-262 pattern anchoring, date-time assertion, keyword audit (G003) |
| `infra/docker/intelligence.Dockerfile` | conformance corpus in the check stage only (G003) |

**Already registered, content changes:** `web/package.json` (scripts only) and `docs/foundation/FOLLOW-UP.md`.

### 14.3 New files (175)

| Area | Count | Paths |
|---|---|---|
| Core | 6 | `core/internal/observability/{context_test.go,exit.go,exit_test.go,errorlog.go,errorlog_test.go}`, `core/internal/health/propagation_test.go` |
| Web | 4 | `web/lib/request-context.ts`, `web/tests/unit/request-context.test.mjs`, `web/tests/fixtures/error-probe/{wrapped-route.ts,control-route.ts}` |
| Intelligence | 6 | `intelligence/src/observability/http.py`, `intelligence/tests/{test_health_context.py,test_logging_privacy.py,test_error_output.py,test_conformance.py,conformance_exchange.py}` |
| Contracts | 128 | as in r1: `contracts/conformance/README.md`, `contracts/conformance/intelligence/expectations.json`, and 126 fixtures (EV01–EV14, EI01–EI50, XV01–XV15, XI01–XI34, NV01–NV02, NI01–NI05, SN02–SN07), with r1's anchored per-directory allow-list expressions. SD01 and SD02 have no fixture. |
| Validation tools | 20 | `infra/validation/tools/{go.mod,go.sum}`; `cmd/{contractcheck,httpprobe,logcheck}/{main.go,main_test.go}` (6); `internal/conformance/{validate.go,audit.go,derive.go,ndjson.go,produce.go,conformance_test.go}` (6); `internal/logrecord/{completion.schema.json,events.json,logrecord.go,logrecord_test.go}` (4); `infra/docker/validation-tools.Dockerfile` and `.dockerignore` (2). r1 stated "(17)" for a list of 18 (F23). |
| CI | 7 | `.gitlab-ci.yml`, `infra/ci/{validate.sh,lib.sh,validate-controls.sh,cleanup-owned.sh,runner-preflight.sh,unchanged.allow}` |
| Harness | 1 | `infra/validation/required-checks.tsv` |
| Docs | 1 | `docs/foundation/OPERATIONS.md` (r1 coverage, plus: error-output handling, attribution profiles, the finalization and exit-code model, the registry and cleanup, the runner requirements and R1. It names no runner tag, GitLab path or identity literal.) |
| Management | 2 | `management/v0/results/G003-scope-allowlist.txt` (**created at S1**, with the full final content of this section) and `management/v0/results/G003-foundation-observability-ci-result.md` (after execution) |

Count: 6 + 4 + 6 + 128 + 20 + 7 + 1 + 1 + 2 = 175.

The probe-only files of R1 (`.gitlab-ci.yml` probe variant, `infra/ci/runner-probe-job.sh`, `infra/ci/probe.allow`, and a preview of `runner-preflight.sh`) exist only in the disposable R1 clone and on the deleted probe branch. They are **not** repository paths and are not in the allow-list.

### 14.4 Modified project-owned files

- `infra/validation/run.sh`, `infra/validation/lib.sh`, `infra/validation/render-customizations.tsv` (+11 rows).
- `README.md`: health lists all six routes and the headers; new blocks `platform:contracts` and `platform:ci`; a link to OPERATIONS.md; the existing blocks are unchanged.

### 14.5 Not touched

- `core/cmd/migrate/**`, `core/internal/{config,mcp,migrate}/**`, `core/migrations/**`, `core/go.{mod,sum}`.
- `core/internal/health/{health.go,health_test.go}`, `core/internal/observability/observability_test.go`.
- `web/app/{layout.tsx,page.tsx,globals.css}`, `web/components/**`, `web/lib/{config,appearance,tokens,utils}.ts`, `web/package-lock.json`, all web configuration files, `web/tests/browser/**`, `web/tests/fixtures/token-scopes/**`.
- `intelligence/{pyproject.toml,uv.lock}`, every other `intelligence/src/**` file, the existing `intelligence/tests/*.py`.
- `contracts/README.md`, `contracts/intelligence/**`.
- `infra/compose/**`, every other Dockerfile and `.dockerignore`, `infra/{local,minio,postgres,security}/**`.
- `infra/validation/run-approved{,-controls}.sh`.
- `.env.example`, `.gitignore`, `.archetype/**`.
- `skills/**`, including the in-progress consolidation, which belongs to the owner. AGENTS.md, CLAUDE.md, `docs/ARCHITECTURE.md`, SECURITY.md, DECISIONS.md, `docs/history/**`, the other `docs/foundation/*.md`.
- ROADMAP, BACKLOG, the G003 Goal and Spec, PROGRESS (Codex-owned), and every G000–G002 artifact.
- The tool-local `.claude/settings.local.json`: never read or modified.

### 14.6 Temporary writes (none persist in the repository)

| Write | Location | Removal |
|---|---|---|
| S0 gate copy and evidence | `$EXT/g003-exec/` (outside the repository) | kept as evidence outside the repository |
| R1 probe clone | `$EXT/r1/probe` | deleted after the probe branch is deleted |
| Go module cache for the tools module (S2) | `$EXT/gomod` | deleted at the end |
| Component-check, runtime and probe images | `g003-$EXEC-{core,web,intelligence,vtools}{,-check}`, `g003-$EXEC-web-probe`, labelled `g003.exec=$EXEC` | removed at the end, by label |
| Ad hoc error-output probe inputs and outputs (S3–S5) | `$EXT/s3`, `$EXT/s4`, `$EXT/s5` (environment files mode 600) | canary-bearing inputs deleted after use; outputs kept only after a canary scan |
| Harness work, Compose projects, images, root `.env` | as today (`project-validation-*`, `g001-*-<run>`; root `.env` created and hash-removed by the root suite) | the harness's own cleanup, then `cleanup-owned.sh` verification in CI |
| CI owned registry | `$(dirname "$CI_PROJECT_DIR")/.platform-validation-owned-$CI_JOB_ID` (CI only) | F9 or `after_script` |
| Local CI-parity copy (S8) | `$EXT/s8/clone`, with a scratch commit (Q6) | deleted after the run |
| Hosted evidence downloads | `$EXT/hosted/` | kept outside the repository |

## 15. Execution sequence, gates and provenance

### 15.1 S0: read-only, fail-closed starting-state gate (adapted from approved G002 r6 S0a)

- **Reuse.** The committed wrapper `run-approved.sh`, with `--check-only`, already provides the snapshot-digest, format, count and entry checks, so G003 needs no bootstrap writes. S0 adds what the wrapper does not check:
  - the HEAD binding;
  - the NUL-safe starting state;
  - the resource and port inventory;
  - evidence containment.

  It keeps G002 r6's read-only contract: `realpath -m` containment before any `mkdir`, and Git with `--no-optional-locks -c core.fsmonitor=false`, `GIT_OPTIONAL_LOCKS=0` and redirection variables unset.
- **Running it.** The script runs from **outside the repository** after byte-exact extraction from the block below. Its SHA-256 must equal **`0409148a630f32cc79642fd8ae65b90a53001eba379775bd96a3c1f313a66723`**; a different hash means stop. BASE, SNAPSHOT, SNAPSHOT_SHA256 and COUNT are taken verbatim from Codex's approval record. Nothing is regenerated, and the G002 snapshot is never reused.

```bash
bash <ext>/g003-s0-gate.sh <repo> <BASE> <approved snapshot> <its SHA-256> <COUNT> <ext>/g003-exec/s0
```

| Step | Verifies | Exit |
|---|---|---|
| 0 | Six arguments. BASE is 40 hex, the digest is 64 hex, COUNT is 1–5. The repository resolves. The evidence path resolves **outside** the repository, including through traversal and symlinks (dangling ones included), **before any write**. | 26 |
| 1 | `HEAD` = BASE | 20 |
| 2 | The snapshot is read once; the SHA-256 of those bytes equals the pinned digest | 21 |
| 3 | Only those bytes are parsed: `<64hex>  <path>`, no escaped entries, relative, no `.` or `..`, unique, inside the G003 input set (Goal, Spec, PROGRESS.md, this Plan, ARCHITECTURE.md), exactly COUNT entries, and the Plan present | 22 |
| 4 | Every entry is a regular non-symlink file whose SHA-256 matches. This holds whether or not the entry is committed. | 23 |
| 5 | NUL-safe `status --ignored`: every differing path is `??` or ` M` **and** is a snapshot path. Nothing staged, deleted, renamed, ignored (including `.env` and `.env.local`) or extra is allowed. **The single exception** is the exact ignored path `.claude/settings.local.json`, which is counted and reported, never read (F12). | 24 |
| 6 | No root `.env`; Compose 2.24 or later; no containers, volumes or networks for the project; default ports free; the Docker inventory written to evidence | 25 |

**Exit 0 is the only way to proceed.** Any other exit stops the work before the first repository write.

Exact S0 script (SHA-256 `0409148a630f32cc79642fd8ae65b90a53001eba379775bd96a3c1f313a66723`; extract the lines between the fences, including the final newline):

```bash
#!/usr/bin/env bash
# G003 S0: read-only, fail-closed starting-state gate (G003 Plan §15.1, r1). Adapted from the
# approved G002 r6 S0a gate with the same read-only contract. Needs only bash (>= 4.4), GNU
# coreutils, git and docker. Writes nothing inside REPO, including .git: Git runs without optional
# locks, so the index is never refreshed. The evidence destination is resolved and proven outside
# REPO before anything is created.
# Usage: bash g003-s0-gate.sh REPO BASE SNAPSHOT SNAPSHOT_SHA256 COUNT EVIDENCE_DIR
#   BASE, SNAPSHOT, SNAPSHOT_SHA256 and COUNT are the values pinned by Codex's Plan approval.
# Exit: 0 pass; 20 HEAD; 21 snapshot digest; 22 snapshot format/paths/count; 23 input hash;
#       24 starting state; 25 local resources or tools; 26 usage or evidence path.
set -euo pipefail
shopt -s inherit_errexit

fail() { printf 'S0 FAIL %s: %s\n' "$1" "$2" >&2; exit "$1"; }
[[ $# -eq 6 ]] || fail 26 "usage: REPO BASE SNAPSHOT SNAPSHOT_SHA256 COUNT EVIDENCE_DIR"

# Git inspection must not write repository metadata (no index refresh, no fsmonitor) and must not
# be redirected by the caller's environment.
export GIT_OPTIONAL_LOCKS=0
unset GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE GIT_OBJECT_DIRECTORY GIT_ALTERNATE_OBJECT_DIRECTORIES \
  GIT_COMMON_DIR GIT_NAMESPACE GIT_CONFIG_PARAMETERS GIT_CONFIG_COUNT
g() { git --no-optional-locks -c core.fsmonitor=false -C "$REPO" "$@"; }

# 0. Arguments, and every location resolved before any write.
REPO="$(realpath -e -- "$1" 2>/dev/null)" || fail 26 "repository path does not resolve"
[[ -d "$REPO" ]] || fail 26 "repository path is not a directory"
BASE="$2" SNAP="$3" SNAP_SHA="$4" COUNT="$5"
[[ "$BASE" =~ ^[0-9a-f]{40}$ ]] || fail 26 "BASE must be a full 40-hex commit"
[[ "$SNAP_SHA" =~ ^[0-9a-f]{64}$ ]] || fail 26 "SNAPSHOT_SHA256 must be 64 lowercase hex"
[[ "$COUNT" =~ ^[1-5]$ ]] || fail 26 "COUNT must be 1-5"
EV="$(realpath -m -- "$6" 2>/dev/null)" || fail 26 "evidence path does not resolve"
inside_repo() { [[ "$1" == "$REPO" || "$1" == "$REPO"/* ]]; }
inside_repo "$EV" && fail 26 "evidence destination resolves inside the repository: $EV"
# Only now may anything be created; the resolved path is re-verified before every write.
mkdir -p -- "$EV"
ev_ok() {
  local now
  now="$(realpath -e -- "$EV" 2>/dev/null)" || fail 26 "evidence directory vanished"
  if [[ "$now" != "$EV" || ! -d "$now" ]] || inside_repo "$now"; then fail 26 "evidence directory changed: $now"; fi
}
ev_ok

PLAN=management/v0/plans/G003-foundation-observability-ci-plan.md
PATH_RE='^(management/v0/(goals/G003-foundation-observability-ci\.md|specs/G003-foundation-observability-ci-spec\.md|PROGRESS\.md|plans/G003-foundation-observability-ci-plan\.md)|docs/ARCHITECTURE\.md)$'
LINE_RE='^([0-9a-f]{64})  ([^\].*)$'
cd -- "$REPO"

# 1. Baseline commit.
[[ "$(g rev-parse HEAD)" == "$BASE" ]] || fail 20 "HEAD is not $BASE"

# 2. Snapshot: read once; verify the digest of exactly those bytes.
bytes="$(cat -- "$SNAP"; printf x)" || fail 21 "cannot read the snapshot"
bytes="${bytes%x}"
[[ "$(printf '%s' "$bytes" | sha256sum | cut -d' ' -f1)" == "$SNAP_SHA" ]] || fail 21 "snapshot digest mismatch"

# 3. Parse only those bytes: format, relative unique approved paths, exact count, Plan present.
[[ -n "$bytes" && "$bytes" == *$'\n' ]] || fail 22 "snapshot is empty or lacks a final newline"
declare -A want=()
n=0
while IFS= read -r line; do
  [[ "$line" =~ $LINE_RE ]] || fail 22 "malformed entry: $line"
  sha="${BASH_REMATCH[1]}" path="${BASH_REMATCH[2]}"
  [[ "$path" != /* && "/$path/" != */./* && "/$path/" != */../* ]] || fail 22 "non-relative path: $path"
  [[ "$path" =~ $PATH_RE ]] || fail 22 "path outside the approved set: $path"
  [[ -z "${want[$path]+x}" ]] || fail 22 "duplicate path: $path"
  want[$path]="$sha"; n=$((n + 1))
done <<<"${bytes%$'\n'}"
(( n == COUNT )) || fail 22 "expected $COUNT entries, found $n"
[[ -n "${want[$PLAN]+x}" ]] || fail 22 "the Plan is not an entry"

# 4. Every input is a regular file matching its approved hash (committed or not).
for path in "${!want[@]}"; do
  [[ -f "$path" && ! -L "$path" ]] || fail 23 "missing or non-regular input: $path"
  [[ "$(sha256sum -- "$path" | cut -d' ' -f1)" == "${want[$path]}" ]] || fail 23 "hash mismatch: $path"
done

# 5. Starting state (NUL-safe, no optional locks): every path that differs from HEAD is a snapshot
#    path and is untracked or unstaged-modified; nothing staged, deleted, renamed or ignored.
#    Snapshot paths that are committed unchanged at BASE are allowed (step 4 bound their bytes).
#    The single exception is the ignored, tool-local Claude Code settings file named below: it is
#    never a repository input, is not read, and is reported in the evidence.
TOOL_LOCAL=.claude/settings.local.json
ev_ok
g status --porcelain=v1 -z --untracked-files=all --ignored >"$EV/s0-status.z" || fail 24 "git status failed"
ndiff=0 ntool=0
while IFS= read -r -d '' rec; do
  xy="${rec:0:2}" path="${rec:3}"
  if [[ "$xy" == "!!" && "$path" == "$TOOL_LOCAL" ]]; then ntool=$((ntool + 1)); continue; fi
  [[ "$xy" == "??" || "$xy" == " M" ]] || fail 24 "unexpected status '$xy': $path"
  [[ -n "${want[$path]+x}" ]] || fail 24 "differing path not in the snapshot: $path"
  ndiff=$((ndiff + 1))
done <"$EV/s0-status.z"

# 6. Local resources the execution must never touch, and the required tools.
project="$(sed -n 's/^COMPOSE_PROJECT_NAME=//p' .env.example)"
[[ -n "$project" ]] || fail 25 "cannot read COMPOSE_PROJECT_NAME from .env.example"
[[ ! -e .env && ! -L .env ]] || fail 25 "a root .env exists"
cv="$(docker compose version --short 2>/dev/null)" || fail 25 "docker compose unavailable"
[[ "$cv" =~ ^v?([0-9]+)\.([0-9]+) ]] || fail 25 "unparsable compose version: $cv"
(( BASH_REMATCH[1] > 2 || (BASH_REMATCH[1] == 2 && BASH_REMATCH[2] >= 24) )) || fail 25 "Compose >= 2.24 required (found $cv)"
for kind in "ps -a -q" "volume ls -q" "network ls -q"; do
  # shellcheck disable=SC2086
  out="$(docker $kind --filter "label=com.docker.compose.project=$project")" || fail 25 "docker query failed"
  [[ -z "$out" ]] || fail 25 "existing $project resources ($kind)"
done
for port in 3100 8180 8100 5532 6479 9100 9101; do
  if (exec 3<>"/dev/tcp/127.0.0.1/$port") 2>/dev/null; then fail 25 "port $port is busy"; fi
done
ev_ok
{ docker ps -a --format 'container {{.Names}}'; docker volume ls --format 'volume {{.Name}}'
  docker network ls --format 'network {{.Name}}'; docker image ls --format 'image {{.Repository}}:{{.Tag}}'; } |
  LC_ALL=C sort >"$EV/s0-docker-inventory.txt"

ev_ok
printf 'S0 PASS: HEAD %s; snapshot %s; %d inputs verified; %d differing path(s), all snapshot paths; %d tool-local ignored file(s); no .env, no %s resources, default ports free; compose %s\n' \
  "$BASE" "$SNAP_SHA" "$n" "$ndiff" "$ntool" "$project" "$cv" | tee "$EV/s0-gate.txt"
```

### 15.1a r2: S0 is unchanged; the baseline is PENDING (correction 1)

- **The script is byte-identical to r1.** Its SHA-256 is still `0409148a630f32cc79642fd8ae65b90a53001eba379775bd96a3c1f313a66723`; r2 re-extracted the block above and verified it. The PATH_RE, COUNT range, the single tool-local ignored exception (Q13) and the exit codes are unchanged.
- **Nothing is added to tolerate the skill consolidation.** There is no path exception, no widened PATH_RE, no `skills/` allowance and no status relaxation. In the current working tree (§2.1), S0 would stop at step 5 with exit 24 on the first `skills/vestive-design-skill/...` record. That is the intended outcome, and it prevents any implementation write.
- **What unblocks S0.** Q14 resolved (recommended route A: the completed consolidation is committed by the owner), then the fresh Codex binding (Q9: BASE, snapshot file, its SHA-256, COUNT). Under route A, the only differing paths at S0 are this Plan and any supervisor-owned files Codex binds.
- **If Codex chooses another route,** it needs a reviewed Plan revision of the S0 inputs (Q14, routes B and C). It is never an execution-time exception.
- **The r1 demonstrations below (§15.2) remain valid** for the unchanged script and are not repeated.

### 15.2 Planning-time demonstration (read-only for the real repository)

- **Setup.** Run on 2026-10-05 with the exact script above.
  - shellcheck v0.11.0 (`-S warning`) is clean.
  - Every case ran in a **disposable `git clone`** of the baseline, in the job's scratch directory, with an untracked stand-in Plan.
  - Snapshots were built from the clone's files. These snapshots demonstrate the gate only; they are **not approval snapshots**.
- **Measurement.** Around every invocation: a whole-clone inventory including `.git` (path, type, size, nanosecond mtime, inode) and `.git/index` hash plus stat data.
- **Real repository.** Its inventory, including `.git`, and its index were unchanged across the whole demonstration.

| Case | Input | Want | Got | Tree / index |
|---|---|---|---|---|
| P | Plan (untracked) + Goal + Spec + PROGRESS, COUNT 4 | 0 | 0, `4 inputs verified; 1 differing path(s), all snapshot paths; 0 tool-local ignored file(s); no .env, no beginify resources, default ports free; compose 2.32.4-desktop.1` | same / same |
| P5 | As P plus ARCHITECTURE.md, COUNT 5 | 0 | 0 | same / same |
| N1 | Wrong pinned digest | 21 | 21 | same / same |
| N2 | COUNT 3 with 2 entries | 22 | 22 | same / same |
| N3 | Duplicate entry | 22 | 22 | same / same |
| N4 | Absolute path | 22 | 22 | same / same |
| N5 | `README.md` entry (outside the set) | 22 | 22 | same / same |
| N6 | Plan not an entry | 22 | 22 | same / same |
| N7 | Goal modified after snapshot | 23 | 23 | same / same |
| N8 | Extra untracked file | 24 | 24 | same / same |
| N9 | Plan staged (`A `) | 24 | 24 | same / same |
| N10 | Root `.env` (`!!`) | 24 | 24 | same / same |
| N11 | HEAD moved by an empty commit | 20 | 20 | same / same |
| N12 | Unlisted tracked file modified | 24 | 24 | same / same |
| T1 | Ignored `.claude/settings.local.json` present (ignored as in the real checkout) | 0 | 0, `1 tool-local ignored file(s)` | same / same |
| T2 | Ignored `.env.local` present | 24 | 24 (`!!` `.env.local`) | same / same |
| T3 | Untracked `.claude/other.json` | 24 | 24 (not in the snapshot) | same / same |
| A1 | Evidence path inside the repository (nonexistent) | 26 | 26; **not created** | same / same |
| A2 | Evidence through an external symlink into the repository | 26 | 26; **not created** | same / same |
| A3 | Abbreviated BASE | 26 | 26 | same / same |
| C1 | Stale-index fixture (tracked mtimes changed), positive path | 0 | 0 | same / **same** |
| C0 | Ordinary `git status` (optional locks enabled) on that fixture afterwards | index rewritten | rewritten: the detector is non-vacuous | — |

### 15.3 Sequence (r2; correction 6: every input exists before the gate that reads it; r1 sequence in Appendix R1)

| Step | Work | Gate (exact commands in §15.6) |
|---|---|---|
| S0 | §15.1 gate, then the wrapper `--check-only`, then `run-approved-controls.sh` (W1–W9). **Needs Q9 and Q14.** | All exit 0, or stop before any write |
| **S0R** | **R1 runner qualification** (§12.4). Needs Q1, A-R1 and A-E. A disposable clone outside the repository; **no repository write.** | The §12.4 decision rule is met and reported to Codex; otherwise stop before S1 (no cycle consumed) |
| S1 | **First repository write: `management/v0/results/G003-scope-allowlist.txt`**, with the full §14 content. B-SCOPE checks that changed paths are a *subset* of the allow-list, so every later intermediate state passes it.<br>Then: the contracts corpus, `expectations.json` (including `schema_negative`, SD01 and SD02), the conformance README, the `schemas.py` corrections, `test_conformance.py` and `conformance_exchange.py`, and the `intelligence.Dockerfile` check-stage `COPY`. | The intelligence check image builds. `pytest -q` passes, with the conformance, SN-class and SD tests included and none skipped. |
| S2 | The tools module (`go get …@v6.0.3 && go mod tidy` in a container, the only network step), the commands, tests, `events.json` and the Dockerfile | The tools check stage passes; `contractcheck corpus` agrees with the expectations; `contractcheck schema-negative` reports every declared class, a passing `sn-derive` and SD01/SD02 as digest mismatches |
| S3 | Core: `correlation.go`, `logging.go`, `checks.go`, `exit.go`, `errorlog.go`, the two `main.go` hunks, and the tests | `core-check` (gofmt, vet, all tests) passes. Then the core runtime image is built and **CE1/CE2** are run ad hoc (§15.6): JSON only, the expected categories, no canary. |
| S4 | Web: `request-context.ts`, `core-client.ts`, both routes, the scripts, the unit test and the two error-probe fixtures | `web-check` (lint, type check, unit) passes. The production build passes. **WE0–WE6** run ad hoc on a disposable probe copy (§15.6). This is where the real Next 16.3.8 behaviour of F15 is first observed; a deviation stops for review (R14). |
| S5 | Intelligence: `http.py`, `export.py`, `main.py`, `test_error_output.py` and the other tests | The intelligence check passes (ruff, format, all tests). The runtime image is built and **IE1/IE2** are run ad hoc (§15.6). The uvicorn lifespan and h11 templates are confirmed against the installed 0.54.0 package. |
| **S6** | **Integration, in one step, before any static or provenance run reads the result:**<br>• (a) CI files: `.gitlab-ci.yml` (the Q1 tag), `infra/ci/validate.sh`, `lib.sh`, `validate-controls.sh`, `cleanup-owned.sh`, `runner-preflight.sh`, `unchanged.allow`;<br>• (b) documentation: the README `platform:contracts` and `platform:ci` blocks, `docs/foundation/OPERATIONS.md`, the FOLLOW-UP labels (§17);<br>• (c) `infra/validation/required-checks.tsv` (69 IDs);<br>• (d) every §10 harness change: the contracts suite, the error-output checks, the `obs_*` functions, B-HEALTH-ROOT, B-REQUIRED-CHECKS, the register (+11), `APPLICABLE` (+7), the `MAIN_GO_G003_SHA` pin and imports literal, the B-SHELL, B-HADOLINT, B-DOCKERIGNORE, B-DEPS, B-WEB-BOUNDARY and B-MCP-SOURCE changes, and the `--owned-registry` and `--secrets-file` options. | In order, all in the same cycle:<br>• **G6a** shellcheck on every script;<br>• **G6b** `validate-controls.sh` standalone: groups A–E, including V4's nested real `contracts` run and the Docker-real cleanup controls;<br>• **G6c** the wrapper with `static provenance contracts`: **all PASS**. This is the G002 lesson: a shared harness change is followed by the static suite in the same cycle. |
| S7 | — | The wrapper with `root migrations variants`: **all PASS** (B-HEALTH-ROOT, the three error-output checks, the B-OBS-* checks, and every inherited root and variant check) |
| S8 | **Local CI parity:** `validate.sh` in a disposable scratch-commit copy (Q6). Then **H0**, if A-H0 is authorized. | `validate_rc=0`, `tested_content=local`; H0's evidence recorded (§12.5) |
| S9 | Two read-only reviewers (§20); triage; fixes | Findings resolved or recorded. The suites covering each changed file are rerun, and a harness or CI change always reruns G6a, G6b and the `static` suite in the same cycle. |
| S10 | **Final local run with no edits during the run:** the wrapper with `all` (§15.4); the cleanup inventory; the Result draft | `overall_rc=0` |
| S11 | H1 and H2 (with A-H1, A-H2 and A-E), then H3 (A-H3); the final Result; stop for Codex | A hosted pass on the exact commit; a hosted failure for the intended reason |

**r1 → r2 step map:**
- S0 → S0.
- New: S0R.
- S1–S5 → S1–S5, plus the ad hoc error-output probes; the allow-list moves to S1.
- r1 S6 (harness) + S7 (CI) + S8 (docs) → S6, with the allow-list already created at S1.
- r1 S9 → S7 + S8.
- r1 S10 → S9; r1 S11 → S10; r1 S12 → S11.

The **cycle-4 checkpoint** is recorded *after* cycle 4 closes and before cycle 5 begins, wherever that falls (§19).

### 15.4 Final local invocation (S10)

```bash
bash infra/validation/run-approved.sh \
  --snapshot <approved snapshot> --snapshot-sha256 <its SHA-256> --count <COUNT> \
  --path-re '^(management/v0/(goals/G003-foundation-observability-ci\.md|specs/G003-foundation-observability-ci-spec\.md|PROGRESS\.md|plans/G003-foundation-observability-ci-plan\.md)|docs/ARCHITECTURE\.md)$' \
  --plan management/v0/plans/G003-foundation-observability-ci-plan.md \
  -- --base <BASE> --allow-file management/v0/results/G003-scope-allowlist.txt --report-dir <ext>/g003-exec/s10/report all
echo "overall_rc=$?"
```

- **Arguments.** Every non-Plan snapshot entry becomes `--expect-hash` and an exact `--exception`, as the wrapper builds them today.
- **Outcome.** Exit 0 is the only passing result. Any other status is a failed run: it is recorded, and the handoff stops.
- **Scope.** B-SCOPE checks the exact §14 allow-list against BASE. The harness enumerates with `--exclude-standard`, so the globally ignored tool-local file does not enter B-SCOPE, the copies or the scans.
- **Inputs.** `<BASE>`, the snapshot, its SHA-256 and `<COUNT>` are Codex's Q9 binding, never self-generated.

### 15.5 r2 planning-time probes and source evidence (scratch only; no repository write)

These are disposable planning probes under `/tmp/g003-plan-r2/`, the scratch root Codex will see referenced. They are **not** implementation and **not** execution evidence. Runtime confirmation of each point is an S3–S6 gate.
- **Images.** Every container used was already present locally, run with `--pull never`, `--network none` and `--rm`.
- **No application.** No application image was built or run.

| Probe | What was run | Result | Supports |
|---|---|---|---|
| **P-GO** | `main.go` (SHA-256 `de08b71b…be3a`): a stdlib-only scratch program, run with `go run` in `golang:1.27.1-alpine` (`GOTOOLCHAIN=local`), mirroring today's core behaviour and the planned adapters. Output: `out.txt` (`28f39558…1288`). | **Today's behaviour (section A):** plain-text `http: superfluous response.WriteHeader call from main.serve.func1 (main.go:63)`; `http: panic serving 127.0.0.1:…: boom CANARY-7f3a9c` with a full goroutine stack (twice: the client retried); `fmt.Fprintln` of a listen error printing `listen tcp: lookup bad-CANARY-7f3a9c.invalid on …`. That is 3 canary-bearing non-JSON lines. `panic(http.ErrAbortHandler)` printed nothing, and a query containing `;` printed nothing.<br>**Planned behaviour (section B):** with a classifying `ErrorLog` writer and a bounded exit record, 5 lines, every one valid JSON, with `canary=false` (`superfluous_write_header`, `handler_panic` ×2, `listen_failed`, `config_invalid`). | F16, F17; §5.6; §7.1 |
| P-GO-SRC | `grep -n "logf("` over `/usr/local/go/src/net/http/server.go` in the same image | Call sites for hijacked and superfluous `WriteHeader`, hijacked `Write`, invalid `Content-Length`, conflicting length headers, `panic serving`, TLS handshake and `Accept` errors; `logf` falls back to `log.Printf` when `ErrorLog` is nil. There is no semicolon-query log call. `ErrAbortHandler` suppresses the stack. | F17; the §5.6 categories |
| **P-NEXT** (source, not a runtime probe) | Next.js `v16.3.8` tag: `packages/next/src/build/templates/app-route.ts` and `packages/next/src/server/route-modules/route-module.ts` | The route error is rethrown to the outer catch, which calls `routeModule.onRequestError(…, silenceLog=false, …)`, then `sendResponse(…, new Response(null, { status: 500 }))`. `onRequestError` calls `logErrorWithOriginalStack(err, 'app-dir')` or `console.error(err)`. **Not run:** Next cannot be executed without installing dependencies or building, which this handoff forbids. The real-framework probe is WE0–WE6, first at S4 and then permanently as B-WEB-ERROR-OUTPUT. | F15; §5.3; §7.2 |
| P-UVICORN (source) | uvicorn `0.54.0` tag: `config.py` (`LOGGING_CONFIG`, `load_app`) and `server.py` | `uvicorn` has a stderr text handler, `uvicorn.error` propagates to it, and `uvicorn.access` writes to stdout with the request line. `load_app` catches only `ImportFromStringError`. The `server.py` templates are those listed in §5.6. Runtime confirmation: IE1 and IE2 (S5, B-INTEL-ERROR-OUTPUT). | F18; §7.3 |
| **P-FIN-0** | `gotcha.sh` (`238184db…6d56`) | `f \|\| rc=$?` → rc 0, and the function continued past `false`. `( set -e; f ) \|\| …` → rc 0, and it continued. `set +e; ( set -e; f ); rc=$?` → rc 1. | §11.2.1 (the phase form) |
| **P-FIN** | `model.sh` (`3b8233dd…a898`): a control-flow model of §11.2 with stub phases; 11 fault cases; `cases.tsv` (`663900c2…ee30`) | none→0 (published); harness→1; metadata→16 (harness not run; postcheck and cleanup ran; nothing published); malformed report→14; scanner error→17 (nothing published); planted canary in an excluded raw file→15 (published `WITHHELD` only); drift→13; cleanup→18; publish failure→19; unexpected error outside a phase→19 (finalizer ran); SIGTERM during the harness→143 (postcheck and cleanup ran; nothing published). In every case postcheck and cleanup ran; no failure produced 0. | §11.2.3–§11.2.5; groups C and D |
| P-TERM | `termtrap.sh` (`d4e04cbd…2b27`) under `timeout --signal=TERM --kill-after=5s 1s` (GNU coreutils 9.4, Bash 5.2.21) | `timeout` returned 124, and the child's EXIT trap ran (`tt.out`). | §11.2.1 (the internal deadline keeps the harness's own cleanup) |
| **P-ATTR-0** | A Git fixture: target T, source S and merge M (`--no-ff`, parents T and S) under `refs/merge-requests/1/merge`; a depth-1 fetch and checkout of M | `.git/shallow` = M. `git cat-file commit HEAD` lists both parents (T, then S). `git rev-parse --verify HEAD^2` fails at the shallow boundary. | §11.2.4 (the raw parent check) |
| **P-ATTR** | `classify.sh` (`1e0dfe34…2984`), the prototype of §11.2.4, over the fixture with `env -i` and GitLab-defined variables; `cases.sh` (`6cd60dab…5dd5`), results in `results.tsv` (`d5e7a3fe…498e`) | **21 of 21 as expected.** P1 and P2 `branch-push`; P3 `mr-source`; P4 `mr-merged-result`; P5 `local` (all exit 0). N1–N16 each exit 16 with their specific reason. | Correction 2; group B |
| **P-SCAN** | `zricethezav/gitleaks:v8.30.1 dir` with the repository's `gitleaks.toml`, `--redact --exit-code 23`, over scratch directories | clean → **0**; a planted generated `ghp_` token → **23** (`leaks found: 1`); a missing configuration file → **1**; a missing path → **1** | F22; §11.2.3 F3 and F7; V5c and V6a |

Two observations are recorded for Codex:
- **(i)** The Go 1.27.1 standard library no longer logs semicolon queries. r1 assumed nothing about it, and r2 asserts nothing about it.
- **(ii)** gitleaks reports both a fatal error and a default leak with exit 1. That is why `--exit-code 23` is required to tell a scanner failure from a finding.

### 15.6 Exact intermediate commands (correction 6)

**Variables.** These are fixed at S0 and are not repository content:
- `EXT`: the execution evidence root outside the repository, created by S0 only after the containment check.
- `EXEC`: 6 lowercase hex.
- `BASE`, `SNAP`, `SNAP_SHA`, `COUNT`: Codex's Q9 binding.
- `PATH_RE`: the G003 expression in §15.4.
- `PLAN=management/v0/plans/G003-foundation-observability-ci-plan.md`.
- `ALLOW=management/v0/results/G003-scope-allowlist.txt`.

All commands run from the repository root.

```bash
# Wrapper invocation used by every intermediate and final suite run (scope bound to the exact allow-list).
wrap() { bash infra/validation/run-approved.sh --snapshot "$SNAP" --snapshot-sha256 "$SNAP_SHA" --count "$COUNT" \
  --path-re "$PATH_RE" --plan "$PLAN" -- --base "$BASE" --allow-file "$ALLOW" "$@"; }
lbl=(--label "g003.exec=$EXEC")

# S0 (read-only; any nonzero exit stops before a write)
bash "$EXT/g003-s0-gate.sh" "$PWD" "$BASE" "$SNAP" "$SNAP_SHA" "$COUNT" "$EXT/g003-exec/s0"
bash infra/validation/run-approved.sh --snapshot "$SNAP" --snapshot-sha256 "$SNAP_SHA" --count "$COUNT" \
  --path-re "$PATH_RE" --plan "$PLAN" --check-only
bash infra/validation/run-approved-controls.sh

# S0R (R1; no repository write; A-R1 and A-E required)
git clone -q --no-checkout "$PWD" "$EXT/r1/probe" && git -C "$EXT/r1/probe" checkout -q -b g003-runner-probe "$BASE"
#   write the four §12.4 probe files into $EXT/r1/probe, then:
git -C "$EXT/r1/probe" add .gitlab-ci.yml infra/ci/runner-preflight.sh infra/ci/runner-probe-job.sh infra/ci/probe.allow
git -C "$EXT/r1/probe" commit -q -m "G003 runner qualification probe (disposable; branch deleted after R1)"
git -C "$EXT/r1/probe" push git@gitlab.example:<namespace>/<project>.git g003-runner-probe
#   ... evidence retrieval per A-E ...
git -C "$EXT/r1/probe" push git@gitlab.example:<namespace>/<project>.git --delete g003-runner-probe

# S1 gate (after the S1 writes, the first of which is $ALLOW)
docker build -q "${lbl[@]}" -f infra/docker/intelligence.Dockerfile --target check --build-context contracts=contracts \
  -t "g003-$EXEC-intelligence-check" intelligence
docker run --rm --network none "${lbl[@]}" "g003-$EXEC-intelligence-check"

# S2 (the only network step: module resolution)
mkdir -p "$EXT/gomod" "$EXT/s2"
docker run --rm "${lbl[@]}" --user "$(id -u):$(id -g)" -e HOME=/tmp -e GOCACHE=/tmp/gocache -e GOPATH=/gopath \
  -e GOTOOLCHAIN=local -v "$EXT/gomod:/gopath" -v "$PWD/infra/validation/tools:/src" -w /src golang:1.27.1-alpine \
  sh -c 'go get github.com/santhosh-tekuri/jsonschema/v6@v6.0.3 && go mod tidy'
docker build -q "${lbl[@]}" -f infra/docker/validation-tools.Dockerfile --target check -t "g003-$EXEC-vtools-check" infra/validation/tools
docker run --rm --network none "${lbl[@]}" "g003-$EXEC-vtools-check"
docker build -q "${lbl[@]}" -f infra/docker/validation-tools.Dockerfile --target runtime -t "g003-$EXEC-vtools" infra/validation/tools
docker run --rm --network none "${lbl[@]}" -v "$PWD/contracts:/contracts:ro" -v "$EXT/s2:/out" "g003-$EXEC-vtools" \
  contractcheck corpus --schemas /contracts/intelligence --corpus /contracts/conformance/intelligence --out /out/results-go.json
docker run --rm --network none "${lbl[@]}" -v "$PWD/contracts:/contracts:ro" -v "$EXT/s2:/out" "g003-$EXEC-vtools" \
  contractcheck schema-negative --schemas /contracts/intelligence --corpus /contracts/conformance/intelligence --out /out/schema-negative-go.json

# S3 gate (CE1/CE2 environment files are written to $EXT/s3 with mode 600; canaries from /dev/urandom)
docker build -q "${lbl[@]}" -f infra/docker/core.Dockerfile --target check -t "g003-$EXEC-core-check" core
docker run --rm --network none "${lbl[@]}" "g003-$EXEC-core-check"
docker build -q "${lbl[@]}" -f infra/docker/core.Dockerfile --target runtime -t "g003-$EXEC-core" core
for c in ce1 ce2; do docker run --rm --network none "${lbl[@]}" --env-file "$EXT/s3/$c.env" "g003-$EXEC-core" >"$EXT/s3/$c.out" 2>&1; echo "$c rc=$?"; done
docker run --rm --network none "${lbl[@]}" -v "$EXT/s3:/in:ro" "g003-$EXEC-vtools" logcheck startup --service core \
  --case /in/ce1.out=1:core.exit:config_invalid --case /in/ce2.out=1:core.exit:listen_failed --canaries /in/canaries.txt

# S4 gate
docker build -q "${lbl[@]}" -f infra/docker/web.Dockerfile --target check -t "g003-$EXEC-web-check" web
docker run --rm --network none "${lbl[@]}" "g003-$EXEC-web-check"
docker build -q "${lbl[@]}" -f infra/docker/web.Dockerfile --target runtime -t "g003-$EXEC-web" web
mkdir -p "$EXT/s4/copy" && git --no-optional-locks ls-files -z --cached --others --exclude-standard -- web \
  | tar --null -cf - -T - | tar -xf - -C "$EXT/s4/copy"
for r in wrapped control; do mkdir -p "$EXT/s4/copy/web/app/api/__probe/$r"
  cp "$EXT/s4/copy/web/tests/fixtures/error-probe/$r-route.ts" "$EXT/s4/copy/web/app/api/__probe/$r/route.ts"; done
docker build -q "${lbl[@]}" -f infra/docker/web.Dockerfile --target runtime -t "g003-$EXEC-web-probe" "$EXT/s4/copy/web"
#   WE0-WE6: run the probe image (--network none, PROBE_CANARY from $EXT/s4/probe.env), httpprobe in its namespace with
#   $EXT/s4/we.json, docker logs -> $EXT/s4/web-probe.log, logcheck startup/scan; then docker rm -f the probe container.

# S5 gate
docker build -q "${lbl[@]}" -f infra/docker/intelligence.Dockerfile --target check --build-context contracts=contracts \
  -t "g003-$EXEC-intelligence-check" intelligence
docker run --rm --network none "${lbl[@]}" "g003-$EXEC-intelligence-check"
docker build -q "${lbl[@]}" -f infra/docker/intelligence.Dockerfile --target runtime --build-context contracts=contracts \
  -t "g003-$EXEC-intelligence" intelligence
docker run --rm --network none "${lbl[@]}" --env-file "$EXT/s5/ie1.env" "g003-$EXEC-intelligence" >"$EXT/s5/ie1.out" 2>&1; echo "ie1 rc=$?"
#   IE2: docker run -d (valid test env) -> wait for uvicorn.startup_complete in docker logs -> docker stop -> logcheck startup.

# S6 gates (G6a, G6b, G6c, in this order and in the same cycle)
docker run --rm "${lbl[@]}" -v "$PWD:/repo:ro" -w /repo koalaman/shellcheck:v0.11.0 -S warning -x -P SCRIPTDIR \
  infra/ci/validate.sh infra/ci/lib.sh infra/ci/validate-controls.sh infra/ci/cleanup-owned.sh infra/ci/runner-preflight.sh \
  infra/validation/run.sh infra/validation/lib.sh infra/validation/run-approved.sh infra/validation/run-approved-controls.sh \
  infra/local/configure-env.sh
bash infra/ci/validate-controls.sh >"$EXT/s6/ci-controls.log" 2>&1; echo "controls_rc=$?"
wrap --report-dir "$EXT/s6/report" static provenance contracts; echo "overall_rc=$?"

# S7
wrap --report-dir "$EXT/s7/report" root migrations variants; echo "overall_rc=$?"

# S8 (local CI parity in a disposable scratch-commit copy; never pushed; deleted afterwards)
mkdir -p "$EXT/s8/clone" && git --no-optional-locks ls-files -z --cached --others --exclude-standard \
  | tar --null -cf - -T - | tar -xf - -C "$EXT/s8/clone"
git -C "$EXT/s8/clone" init -q && git -C "$EXT/s8/clone" add -A
git -C "$EXT/s8/clone" -c user.name=g003-parity -c user.email=parity@example.invalid commit -q -m "G003 local CI parity (scratch)"
(cd "$EXT/s8/clone" && env -u GITLAB_CI bash infra/ci/validate.sh --evidence-dir "$EXT/s8/evidence"); echo "validate_rc=$?"

# S10: §15.4.
```

- **S4 and S5 probe glue.** The comment lines are the steps that B-WEB-ERROR-OUTPUT and B-INTEL-ERROR-OUTPUT later perform inside `run.sh` (§10). At S4 and S5 they are executed with the same `httpprobe` and `logcheck` arguments, and the outputs are kept under `$EXT`.
- **Cleanup.** Every image and container in this block carries `g003.exec=$EXEC`. They are removed at the end with `docker ps -aq --filter label=g003.exec=$EXEC` and `docker image ls -q --filter label=g003.exec=$EXEC`, each checked against the S0 inventory.
- **Writes.** Nothing above writes inside the repository, except the S1–S6 implementation writes themselves and the module resolution of S2 (`go.mod` and `go.sum` under `infra/validation/tools/`, both allow-listed).

## 16. Test strategy and acceptance-criteria mapping (r2; r1 table in Appendix R1)

| AC | Implementation | Commands and assertions (local unless marked hosted) |
|---|---|---|
| **AC1** Health semantics and bounds | §6, §7 | Unchanged from r1:<br>• **Unit:** Go `health_test.go` and `propagation_test.go` (5 s double → about 2 s); Python `test_health_context.py`.<br>• **Root:** B-HEALTH-ROOT and B-WEB-HTTP.<br>• **Variants:** B-OBS-TIMEOUTS O1–O10 (each single request ≤ 5000 ms by `httpprobe`'s monotonic clock; recovery ≤ 90 s), B-OUTAGE (unchanged), B-OBS-OFF and B-TOGGLE-OFF.<br>• **Evidence:** `obs-timings.tsv`.<br>**r2:** an unexpected web handler exception still yields 500 with an empty body and no-store (WE1). |
| **AC2** Context validation, generation and isolation | §4 | Unchanged from r1: table tests in all three runtimes; isolation at 64, 200 and 100 concurrent requests; cancellation and failure; the allow-list (Go `propagation_test.go`, Node `outboundHeaders`, B-WEB-BOUNDARY); B-OBS-CONTEXT on the enabled copy. |
| **AC3** web → core → intelligence continuity | §4.3, §7 | Unchanged from r1: B-OBS-CHAIN cases (i)–(vii); B-OBS-OFF for the disabled variant. No new route, demo endpoint or provider call; the error-probe routes exist only in a disposable copy, never in the stack under test. |
| **AC4** Structured logs and privacy | §5 | **r1 coverage, retained:** `logcheck` on the full raw logs of both copies; B-OBS-LOGS (exactly one record, implicit 200 and handled 503); B-OBS-PRIVACY canaries (headers, query, body, invalid IDs, `/internal/v1/execute` content, secret-valued IDs); the positive controls; B-SECRET-CANARY.<br>**r2 (correction 5), actual error output:**<br>• `logcheck` rejects any uncatalogued `event` or free-text `error_type` (§5.6), with unit negatives (§9).<br>• **B-CORE-ERROR-OUTPUT** CE1 and CE2 on the real core binary.<br>• **B-WEB-ERROR-OUTPUT** WE0–WE6 on a real Next 16.3.8 production build, with WE4 as the positive control showing the canary *would* leak without the wrapper.<br>• **B-INTEL-ERROR-OUTPUT** IE1 and IE2 on the real intelligence image.<br>• Go `exit_test.go` and `errorlog_test.go`; Node wrapper tests; Python `test_error_output.py`.<br>**Artifacts:** `validate.sh` F3 raw and F7 staged scans (correction 3), so no canary or generated secret can be retained or uploaded. |
| **AC5** Cross-language contracts | §8, §9 | **r1 coverage, retained:** B-CONTRACT-TOOLS, -GO, -PY, -AGREE and -PEER (120 cases each, 0 skipped, equal digests; Go envelopes accepted by Python; real Python NDJSON accepted by Go; four corruptions rejected).<br>**r2 (correction 4): B-CONTRACT-NEGATIVE** — SN01–SN07 in audit mode with declared classes in **both** languages, `sn-derive`, positive audit controls, SD01 and SD02 in normal mode (digest mismatch, 0 cases), and EI39. `pytest` in B-CHECKS-ROOT runs the same SN and SD tests.<br>**Boundaries:** no production client (B-MCP-SOURCE `go list -deps`, B-WEB-BOUNDARY); the schemas are byte-identical. |
| **AC6** Runnable containerized commands | §10–§13, §15.6 | README blocks executed verbatim: the existing ones by the harness; `platform:contracts` and `platform:ci` at S7, S8 and S10 (recorded) and by the CI job itself. B-DOCKER-ONLY, B-CHECKS-ROOT, B-BROWSER-ROOT and B-BROWSER-OFF (95 + 95; 0 skipped, flaky or unexpected). Images are built from locked inputs; check containers run with `--network none`. Tool and image versions and cache use (`docker system df`, cold or warm) are recorded. Cleanup is shown by `inventory-leftover.txt` and **`cleanup-inventory.tsv`** (r2). |
| **AC7** GitLab gating with hosted evidence | §11, §12 | **Static:** B-CI-CONFIG (rules, one gating job, no bypass, artifacts, access, retention, timeouts and `RUNNER_*` sum, tag shape, `after_script`).<br>**Runner (r2, correction 1):** Q1 facts F-R1–F-R13 checked by Codex. **R1** (hosted, before S1): the preflight plus the unchanged baseline harness, cold, on the actual runner, with the decision rule (§12.4). `runner-preflight.tsv` in every pipeline.<br>**Attribution (r2, correction 2):** `attribution.json` with `tested_content` from documented variables only (§11.2.4); group B controls (5 positive, P3F, 16 negative).<br>**Hosted H1:** a pass on the exact submitted commit and tree (pipeline, job, runner, `tested_content`, 69 PASS, duration, artifact digest). **Hosted H2:** a failure through the reviewed gate (B-CONTRACT-*), with scanned artifacts.<br>There is no deploy or publish job, no CI/CD secret and no `allow_failure`. |
| **AC8** Negative controls and failure propagation | §10, §11.3 | **B-CI-CONTROLS:**<br>• group A, (a)–(f): (a) V2, V3; (b) V4 and B-CONTRACT-NEGATIVE; (d) V7, V12; (e) V8, V9; (f) V10, V11 and B-REQUIRED-CHECKS;<br>• group B (attribution);<br>• group C, **(c)**: V5a, V5b and V5c (raw retained, raw excluded, gitleaks path), V6a and V6b (scanner failure, missing secrets list, so nothing published), V6c;<br>• group D (finalization): V14–V18c;<br>• group E (owned cleanup): V19–V26, Docker-real, with unrelated-resource, removal-failure, tampering and `after_script` hand-off controls.<br>**Retained:** W1–W9 (B-WRAPPER-CONTROLS) and C1–C12, C8b and C8c (B-DEPS-CONTROLS). The wrapper postcheck applies to every wrapper run. Each control's status and safe cleanup summary are recorded. |
| **AC9** Scope, architecture and provenance | §14, §17 | B-SCOPE against the exact allow-list (created at S1); B-PROTECTED; B-SKILL-UNCHANGED, against the BASE Codex binds after Q14; B-RENDER-ID; B-RENDER-PLAN and B-RENDER-REPEAT (exactly **19**); B-TOGGLE-OFF-ADDITIONS (**35** applicable, including `main.go`); B-MCP-SOURCE (reconciled, with the same security assertions); B-WEB-DESIGN (rechecked against the consolidated skill assets); and all full regressions as in r1. Diff review: no schema, production route, UI or vendor change. |
| **AC10** Documentation, Result and governance | §14, §19, §22 | OPERATIONS.md, README and FOLLOW-UP (metrics and exporters left open). The Result maps AC1–AC10 to commands and artifacts; separates inherited, new and deferred-open items; separates local (S6–S10), runner-qualification (R1) and hosted (H0–H2) evidence; lists every run, including failures; and records the B7 cycles and the checkpoint after cycle 4. B-REQUIRED-CHECKS (69) with no skips. **Prerequisites Q1, Q2, Q9 and Q14 are resolved before execution** (Spec "unresolved prerequisites are resolved before execution"). Only Codex ACCEPT completes G003. |

## 17. Regression reconciliation and documentation labels (r2; r1 text in Appendix R1)

- **G001 guarantees, preserved and re-run:** migration owner/runtime separation, owner-only metadata and the unchanged migration set; independent datastore port and IP isolation with its controls; liveness and readiness outage semantics; persistence and Redis ephemerality; naming and default ports; the dormant MCP boundary; build-context canaries.
- **`main.go` reconciliation (r2, correction 5; Q16).** r1 kept `main.go` byte-identical. r2 makes two bounded hunks (§7.1). Reconciled under R8 as follows:
  - **Customization register:** a new row (§14.2). B-RENDER-PLAN and B-RENDER-REPEAT then expect exactly 19 conflicts and proposals, the proposal for `main.go` equal to the manifest's rendered content.
  - **Scope list:** `core/cmd/server/main.go` is in the allow-list (§14.2).
  - **Disabled variant:** the file is a plain base-template file, so the disabled render has the same rendered base; adding it to `APPLICABLE` makes B-TOGGLE-OFF-ADDITIONS verify both "rendered base equals the root manifest's" and "variant copy equals the root".
  - **Obsolete equality guard:** B-MCP-SOURCE's `sha(main.go) == manifest_sha` becomes `sha(main.go) == MAIN_GO_G003_SHA`, the reviewed bytes. The guard's **security meaning is retained and strengthened**: the routes are exactly two; the imports are exactly the rendered set minus `fmt`; `go list -deps` excludes MCP, migration, fixture, isolation and `jsonschema` packages; and there is no transport symbol. Its reason code (FAIL "MCP or runtime import boundary") is unchanged.
- **G002 guarantees, preserved and re-run:** appearance and first paint; scoped tokens; the full root and disabled browser matrices; the dependency policy, with justified `devOptional`, the hash-pinned registry CSS and the literal-scan exceptions; wrapper controls; B-WEB-* rules (including B-WEB-DESIGN against the consolidated skill assets). No UI file changes.
- **Accepted deviations and residuals, unchanged:** G002's fixed angle, destructive tint and helper modules; G001's lost logs, SIGKILL residue, owner-role preflight, WSL probe limitation, data-network egress and tag-only pins.
- **Byte-equality guards** are reconciled only through the 11 exact register rows, the 7 `APPLICABLE` additions and the `main.go` digest pin. Every semantic assertion and reason code is unchanged.
- **FOLLOW-UP STAGE-OBSERVABILITY:** labels as in r1:
  - "Metrics export and trace propagation…" stays unchecked, annotated as partly done;
  - "Contract validation… in CI" → `[x]` *(G003)*;
  - "Containerized … CI … change evidence" → `[x]` *(G003: GitLab pipeline; commit, pipeline, job and actor recorded; tested content classified as merge-request source, merged result or branch push)*.
- **Out of scope here:** durable audit (G011); broader security and vulnerability controls (G012); recovery and incident monitoring (G013); Organization data and branding (G016/G017); product AI observability.
- **Architecture guidance.** §60 and §61 are referenced as guidance only. G003 adds no hostname resolver, routing grammar, UI redesign or component change.

## 18. Security and privacy negative controls (summary; r2)

- **No authority.** No diagnostic ID, flag, fixture, CI actor or attribution field confers authority or tenant scope.
- **Network boundaries unchanged.** Test tools join service namespaces only for probes, publish nothing and receive no secret. Canary and secret inputs travel by file.
- **No new secret paths.**
  - No CI/CD secrets are used, and merge-request code gets no privileged variables.
  - The harness's generated-secrets list exists only in the mode-700 registry, outside the checkout and the published set. It is deleted at F9 or by `after_script`.
  - Artifacts are published only after the raw **and** staged scans. A finding withholds everything; a scanner failure publishes nothing.
  - The job log never shows the harness console, record messages or canaries: only phase statuses and `ID STATUS` rows, and the rows only after a clean raw scan.
- **Error output is bounded** (correction 5): no message, stack, argument, address, URL or configuration value is copied into any record. The real-artifact checks prove this with canaries and include a positive control.
- **Owned cleanup refuses unowned resources** and never prunes (V20, V21, V23, V26).
- **Controls proving the checks are not vacuous:** planted canaries (raw retained, raw excluded, gitleaks pattern); WE4; leak stubs; `logcheck` negatives; W1–W9, C1–C12 and groups A–E; flipped corpus expectations; SN01–SN07 with declared classes, `sn-derive`, positive audit controls and SD01/SD02; four NDJSON corruptions.

## 19. Budget, feasibility and checkpoint (B7; r2; r1 text in Appendix R1)

**Budget (unchanged).** B7: at most **7** shared meaningful implementation/debug cycles; mandatory checkpoint **after cycle 4 and before cycle 5**. **Zero cycles have been consumed:** r1 and r2 are planning, and planning probes are not cycles (BUDGET_POLICY §4). Claude does not change the budget.

**What changed since the r1 forecast.**
- **Added work** (from the corrections): bounded error output in three runtimes, with real-artifact checks; the attribution classifier; the finalizer, scanner and registry model; and about 30 more controls.
- **Reduced risk:** R1 qualifies the runner on the unchanged baseline **before S1**. A runner incompatibility or a cold-time problem is therefore found with zero cycles spent, not at H1.
- **Prototyped already:** the r2 planning probes (§15.5) exercised the finalizer control flow, the attribution rules, the scanner exit codes and the Go error output.

| Likely meaningful cycle | r1 estimate | r2 estimate |
|---|---|---|
| S1–S5 first unit and component failures (context tables, the formatter or uvicorn takeover, Node type stripping, exit and error-log adapters) | 1 | 1 |
| Cross-validator disagreements or schema-negative class mismatches (S1/S2) | 0–1 | 0–1 |
| Error-output conformance on the real frameworks (WE, CE and IE at S3–S5) | — | 0–1 |
| Integration observability cases (timing, `logcheck` chains, Next startup lines) | 1–2 | 1–2 |
| Harness and CI reconciliation (static, provenance, controls A–E) | 1 | 1 |
| First hosted run of the G003 job (H0 or H1) | 1 (0 if H0 surfaces nothing) | 0–1 (R1 has already qualified the runner) |
| Read-only review findings | 1 | 1 |

**Run-time estimate behind R1's 90-minute rule.** These are estimates, measured for real at H0 and H1. Cold additions of G003 to an `all` run:
- the contracts suite, with two image builds and both validators: about 8–12 minutes;
- the error-output checks, with one extra web production build and three short runs: about 5–8 minutes;
- the observability matrices O1–O10 on two variants, with recovery waits: about 10–15 minutes.

That is about 25–35 minutes in total. A baseline cold run of at most 90 minutes therefore leaves G003's cold run, about 115–125 minutes, inside the 145-minute harness budget.

**Forecast: 5–7 of 7, most likely 6.** It holds only if:
- **(a)** Q1 resolves to Profile A or A2;
- **(b)** A-R1 is authorized, and R1 meets its decision rule before S1;
- **(c)** A-H0 is authorized, or the hosted surprise risk is accepted.

The margin is thin. G002 forecast 5–6 and used 7. **The rows of the table sum to 4–8.** Reaching 8 would need every uncertain row at its maximum at once; that is the case the checkpoint after cycle 4 exists to catch. If, at the checkpoint, the remaining forecast exceeds the 3 remaining cycles, Claude stops and returns to Codex rather than continuing.

**Checkpoint after cycle 4, before cycle 5** (BUDGET_POLICY §7). It records per-criterion status, suite status, failure convergence and repeated patterns, remaining work against the cycles left, and the forecast. Claude stops and returns to Codex if:
- the forecast exceeds the remaining cycles;
- the same material failure has survived two corrective attempts;
- thrashing appears (§8);
- any escalation condition occurs (§21).

**Counting** follows BUDGET_POLICY §3–§5:
- R1, R1 reruns after owner-side changes, successful gates and read-only reviews are **not** cycles.
- Doubtful cases count.
- The step order creates every file a gate reads before that gate (§15.3), so the written order creates no artificial cycle.

**Split option (for Codex; not taken unilaterally; Q10).** If Q1 cannot be resolved to an acceptable profile, or A-R1 cannot be authorized before execution, G003 as written cannot credibly complete within B7: AC7 requires hosted evidence and the Spec forbids deferring it. Codex may then split it:
- **G003a:** R1–R5 and R8, plus local R6/R7, including the CI entry point, attribution, finalization, cleanup and every control;
- **G003b:** hosted GitLab execution (AC7), once a qualified runner and the authorizations exist.

**I recommend the single Goal** once Q1, A-R1 and Q14 are resolved. The split adds review overhead but no risk reduction once R1 has passed.

## 20. Delegation

- **One implementation owner:** me. There are no implementation subagents: the work shares the harness, the log contract and the three runtimes.
- **Two read-only reviewers at S10**, run in parallel. Neither edits files. Each reports findings with severity and `file:line`, and the cycles they used (expected 0).
  - **SA-P (correlation and privacy):**
    - the §4 profile against W3C and the Spec;
    - isolation, cancellation and panic paths;
    - completion "exactly one" semantics;
    - redaction parity;
    - framework output;
    - canary coverage and non-vacuity.
  - **SA-C (CI, contracts and provenance):**
    - validator independence and the corpus/expectation design;
    - the Python corrections;
    - `validate.sh` exit semantics and V1–V12;
    - artifact allow-list and scan;
    - B-CI-CONFIG;
    - the register, `APPLICABLE` and the allow-list;
    - the G001/G002 regression coverage.
- **Integration and budget.** I triage and integrate every finding. Corrective work counts against the shared B7. Reviewer opinions do not replace Codex review.

**r2 additions to §20.**
- **Timing.** In the r2 sequence the two reviewers run at **S9**; r1's "S10" above refers to r1's numbering (§15.3 step map).
- **SA-P** also reviews the error-output handling:
  - the bounded events and categories (§5.6);
  - WE0–WE6, CE1–CE2 and IE1–IE2, including the WE4 positive control;
  - the uvicorn, library and warning mapping.
- **SA-C** also reviews:
  - attribution against GitLab's documented variables (§11.2.4, group B);
  - the finalizer and exit-code aggregation (§11.2.3–§11.2.5, groups C and D);
  - the registry and cleanup binding (§11.4, group E);
  - schema-negative isolation (§8.2a–§8.3);
  - the `main.go` reconciliation (§17).
- **Still no implementation subagent.** The r2 revision itself was prepared by a planning subagent (§0.2), which made no implementation change and consumed no cycle.

## 21. Risks and escalation (r2; r1 table in Appendix R1)

| # | Risk | Handling |
|---|---|---|
| R1 | The runner profile is incompatible, unavailable or privileged | **Blocker (Q1).** Checked by Codex from the facts, then by R1 before S1. Stop; never use dind or privileged mode silently; never substitute another CI vendor. |
| R2 | uvicorn, a library or `warnings` emits a raw line (before the takeover, or through an unmapped path) | `logcheck` fails visibly (IE2, B-OBS-LOGS). Fix it inside the application's logging configuration, never by widening the allow-list. If only an `uvicorn --log-config` or Dockerfile `CMD` change would work, that is returned to Codex. |
| R3 | Node 24 type stripping cannot load `request-context.ts` from `.mjs` | As r1: fall back to a Playwright `node`-only project in the existing browser image, with no new dependency. Recorded as a deviation. |
| R4 | Go 1.27 `ServeMux` does not expose `r.Pattern` on the wrapped request | As r1: derive `route` from the fixed set; the test asserts the result |
| R5 | santhosh and Python disagree on a corpus case **or an SN class** | Diagnose against JSON Schema and RFC semantics. **Never relax schemas, drop cases or merge classes.** A library defect with no conforming configuration is escalated (Spec risk 3). |
| R6 | The paused-target record count (O8, O10) is nondeterministic | As r1: assert 0 or 1 with matching IDs, never more |
| R7 | The cold run exceeds the budget | **Measured at R1, before S1** (§12.4 decision rule). Any timeout change is an owner or Codex decision; checks are never dropped. |
| R8 | Artifact size (the browser JSON, the harness console) | Attachments are stripped; the size is measured at R1 and H0 against F-R11 |
| R9 | Gitleaks flags a synthetic fixture | As r1: change the fixture value; never add an allow-list entry |
| R10 | Harness edits weaken G001/G002 checks | §10 lists every change; SA-C reviews; the static suite runs after each harness change, in the same cycle |
| R11 | Budget exhaustion or thrashing | Checkpoint after cycle 4; the stop rules (§19) |
| R12 | The client tool rewrites `.claude/settings.local.json` mid-run | As r1: covered only by the exact ignored-path exception (Q13). Any other path still fails S0, B-SCOPE or the postcheck. |
| **R13** | A real merged-results commit does not have exactly the source and target SHAs as parents | The run fails closed with 16, never with a false merged-result claim. The rule is a repository-side check, not a GitLab-documented guarantee (§11.2.4). If F-R10 says merged results are enabled and H1 shows another structure, that is a planning correction returned to Codex, not a silent relaxation. |
| **R14** | The real Next 16.3.8 behaviour differs from the source reading (F15). For example, it emits output for a request rejected before the wrapper (WE5), on client abort (WE6), or startup lines outside the allow-list (WE0). | Stop affected work and return it to Codex. The options (a custom server, console patching or a framework-level hook) are design changes needing review. The allow-list is never widened silently, and output is never suppressed globally. |
| **R15** | `SystemExit(3)` from `main.__getattr__` does not exit cleanly under uvicorn 0.54.0 (IE1) | Return to Codex with the observed output. An alternative, such as validating the configuration in a wrapper entry point, may change the Dockerfile `CMD` and needs review. |
| **R16** | `RUNNER_SCRIPT_TIMEOUT` and `RUNNER_AFTER_SCRIPT_TIMEOUT` are not honoured by the runner's version (F-R4) | The internal deadline (§11.2.1) remains the gating mechanism. `after_script` then simply does not run on a job timeout, which F14 already allows for. Recorded in the Result. |
| **R17** | The skill consolidation does not finish, is not committed, or contains a file byte-identical to a web source | Execution stays blocked (Q14). B-WEB-DESIGN detects byte identity (0 matches observed in §2.1). Any resulting conflict is the owner's decision, never a G003 exception. |
| **R18** | Probe-branch ancestry publishes a not-yet-pushed BASE commit (route A) | It is covered explicitly by the A-R1, A-H0 and A-H1 authorization text (§12.3) |

**Stop and escalate on:**
- an architecture or ADR conflict;
- scope expansion;
- a security-model change, including runner privileges;
- a new vendor or infrastructure;
- any schema, migration, production route, API or MCP change;
- any `main.go` change beyond the two reviewed hunks;
- any change to `local.yml`, the deadlines or the schema bytes;
- any Dockerfile `CMD` change;
- weakening a check, an assertion or a control expectation;
- touching developer resources;
- missing hosted authorization;
- an unresolved Q1, Q2, Q9 or Q14;
- B7 exhaustion.

## 22. Questions for Codex and the owner, and unmet prerequisites (r2; r1 table in Appendix R1)

| # | Status | Question or prerequisite | What is needed, and the recommendation |
|---|---|---|---|
| **Q1** | **UNRESOLVED — blocking** | The runner facts | The owner supplies F-R1–F-R13 (§12.2), as named fields only, never the `config.toml` token. Codex checks Profile A or A2 before approving execution. R1 verifies them on the runner before S1. Nothing about the runner is assumed by this Plan. |
| **Q2** | **UNRESOLVED — blocking** | Hosted-action authorization | Explicit, per-action authorization of A-R1, A-H0 (optional), A-H1, A-H2, A-H3 and A-E (§12.3), including the commits each needs and the publication of any not-yet-pushed ancestors. None is granted by this Plan, by its approval or by the planning delegation. |
| **Q9** | **PENDING — blocking** | The fresh binding | After (i) Codex approves this Plan and (ii) Q14 is resolved, Codex supplies BASE, the snapshot file (entries from Goal, Spec, PROGRESS, Plan and, optionally, ARCHITECTURE), its SHA-256 and COUNT. The r1 and r2 planning observations, `/tmp/g003-r1-reviewed-inputs.sha256` and G002's snapshot are **not** bindings, and no self-generated digest is used. |
| **Q14** | **OPEN — owner/Codex decision; blocks Q9** | The baseline route for the in-progress `skills/vestive-design-skill/**` consolidation (§2.1) | **Route A (recommended):** when it is complete and verified, the owner commits it; Codex binds BASE to that commit; S0, B-SKILL-UNCHANGED and B-PROTECTED then hold unchanged.<br>**Route B:** execute G003 in an isolated clone or worktree at the current BASE. That needs Codex to decide that the isolated checkout is the "actual root" (AC6, AC9) and to authorize creating it, and it leaves a later integration with the consolidation to be validated separately.<br>**Route C:** keep the consolidation uncommitted and bind each exact path and hash as a snapshot input. That needs a reviewed Plan revision of S0's PATH_RE and COUNT; r2 deliberately does not make it.<br>Routes B and C are not planned here. |
| Q3 | Resolved by Codex (r1) | The Python format and anchor corrections | Recorded (§0.1) |
| Q4 | Resolved by Codex (r1) | F11 nested deadlines | Recorded (§0.1) |
| Q5 | Resolved by Codex (r1) | Credentials supplied as IDs | Recorded (§0.1) |
| Q6, Q11, Q12 | Resolved by Codex (r1) | Local CI parity; corpus location; tool module | Recorded (§0.1) |
| Q7 | Resolved, subject to verification | Artifact retention and access | Verified against F-R11 and at R1 and H0 |
| Q8 | Resolved by Codex (r1) | The gate is the required job's failure | Recorded |
| Q10 | **Open (Codex)** | Single Goal versus split | Single Goal once Q1, A-R1 and Q14 are resolved (§19) |
| Q13 | Resolved by Codex (r1) | The tool-local ignored-file exception | Unchanged in the S0 gate |
| **Q15** | New, for review | The web wrapper answers an unexpected handler exception with its own 500 (empty body, plus `Cache-Control: no-store` and `X-Request-Id`) instead of rethrowing to Next (§5.3, F15) | Approve. It is the same status and body that Next sends; the change only removes Next's raw message and stack output. |
| **Q16** | New, for review | Customize `core/cmd/server/main.go` with exactly two hunks, and reconcile B-MCP-SOURCE's byte-equality guard to a reviewed digest plus the retained and added semantic assertions (§7.1, §10, §17) | Approve. It is required by correction 5 and preserves every security assertion. |
| **Q17** | New, for review | Unsupported pipeline profiles (merge trains, tags, schedules, web, API, trigger, pipeline, parent pipeline) fail closed with 16; a merged result is claimed only under the §11.2.4 rule | Approve |
| **Q18** | New, for review | R1 uses a disposable probe-only CI configuration and the unchanged baseline harness before S1 (§12.4); the job sets `RUNNER_SCRIPT_TIMEOUT` and `RUNNER_AFTER_SCRIPT_TIMEOUT` (§11.1) | Approve |
| **Q19** | New, for review | Artifact scanning moves from the proposed harness check B-ARTIFACT-SCAN to `validate.sh`'s raw and staged scans; there are 69 required IDs (§10) | Approve |

## 23. Readiness (r2)

**r2: ready for Codex plan review. Execution is blocked by unresolved prerequisites.**

- **Blocking prerequisites:**
  - **Q1:** the runner facts are unresolved;
  - **Q2:** the hosted-action authorizations are unresolved;
  - **Q14:** the baseline route is open, and baseline verification is PENDING (§2.1);
  - **Q9:** the fresh binding is pending Plan approval and Q14.
- **First execution steps once those clear:** S0 and then R1 (§15.3), before any repository write beyond the S0 evidence outside the repository.
- **Verified while writing r2:**
  - HEAD `9461c4b`; the Goal, Spec and ARCHITECTURE hashes unchanged from r1;
  - the S0 script block re-extracts to `0409148a…6723`;
  - Appendix R1 re-extracts to the r1 SHA-256 `a43aa523…dbfc`;
  - the planning probes in §15.5;
  - the only repository write is this Plan.
- **Not done:** nothing has been implemented, installed, initialized, built, committed, pushed or triggered. No hosted action and no external configuration change occurred.
- **Next step:** implementation waits for an explicit Codex approval that names BASE, the approved snapshot, its SHA-256 and COUNT, together with the owner's Q1 facts and Q2 authorizations.

~~~~~~~~~~

## Appendix R1 — complete r1 Plan, verbatim (superseded wherever r2 differs; not an execution instruction)

This block preserves the reviewed r1 text byte-for-byte, as review history, because this Plan file is untracked and has no other durable copy in the repository.
- **Precedence.** Where r1 and the body above differ, **the body above governs**.
- **Contents.** The block includes r1's own copy of the unchanged S0 script. §15.1 above is the extraction source for execution.
- **Verification.** The lines strictly between the opening `~~~~~~~~ text` line and the closing `~~~~~~~~` line have the r1 SHA-256 `a43aa523624bab98c38e0fe6c7ff86cffef2cd70d7a3da7b6e806d5c60adbdfc`:

```bash
awk '$0=="~~~~~~~~ text"{f=1; next} $0=="~~~~~~~~"{f=0} f' management/v0/plans/G003-foundation-observability-ci-plan.md | sha256sum
```

~~~~~~~~ text
# G003 — Health, logging, contracts and baseline CI implementation plan

Goal: [G003-foundation-observability-ci.md](../goals/G003-foundation-observability-ci.md)
Spec: [G003-foundation-observability-ci-spec.md](../specs/G003-foundation-observability-ci-spec.md)
Owner: Claude Code · Reviewer: Codex · Budget class: **B7**: at most 7 shared meaningful cycles; mandatory checkpoint after cycle 4 and before cycle 5. Planning consumed no cycle.

| Revision | Date | Change |
|---|---|---|
| r1 | 2026-10-05 | Initial Plan. Submitted for Codex review. **Not approved; no implementation has begun.** One unmet prerequisite blocks execution (the GitLab runner profile, §12 and Q1); the remaining questions are in §22. |

## 0. Review record and planning-session disclosures

No Codex review has taken place yet.

Disclosures from this planning session:

- **Writes.** The only repository write is this file.
  - Scratch writes were made only outside the repository, under the job's temporary directory: the S0 gate script, its demonstration clones and their evidence (§15.1–§15.2), and a staging copy of this Plan.
  - The background session's write guard required this Plan to be staged outside the repository first. It is placed at its canonical path only as the single repository write.
- **One Git call with optional locks enabled.** The first baseline check of the session ran an ordinary `git status --porcelain=v1 -uall`. It printed nothing, but because optional locks were enabled it may have refreshed `.git/index` stat data. It cannot have changed content. Every later Git inspection used `git --no-optional-locks` with `GIT_OPTIONAL_LOCKS=0`.
- **Tool-local ignored file.** During the session, at 17:22 local time, the Claude Code client created `.claude/settings.local.json`. It is ignored by the user's global excludes (`~/.config/git/ignore`), and was not created, read or modified by me. It is not a repository input. §15.1 treats it as the single exact ignored-path exception.
- **What I did not do.** I did not install anything, initialize any tool, build any application image, commit, push, trigger a pipeline or change any external configuration.
- **What I did run.** Docker was used only for read-only tooling on scratch files and inventories:
  - `mikefarah/yq` reading the committed manifest;
  - `koalaman/shellcheck` on the scratch S0 gate;
  - inventory and port queries.
- **Remote queries (read-only).**
  - `git ls-remote --symref origin` over SSH, with batch mode on.
  - One unauthenticated GitLab API probe. It returned 404, because the project is private.
- **No credential, token or environment secret was read or printed.**

## 1. Summary

G003 completes the existing operational paths. It does not rebuild them.

- **Diagnostic context (R2/R3).**
  - Web, core and intelligence validate `X-Request-Id` and a W3C `traceparent` version-00 profile on the six health routes.
  - Each service creates its own nonzero span. Context is propagated on exactly two existing calls:
    - web `/api/ready` → core `/readyz`;
    - core `/readyz` → intelligence `/healthz` (the optional probe).
  - The parent chain is inspectable: client parent → web span → core span → intelligence span.
- **Completion logs and privacy (R4).**
  - Each service writes exactly one JSON completion record per scoped request, using one shared key set (§5).
  - Core keeps its redacting logger. Intelligence gains equivalent redaction and takes over uvicorn's access and error loggers. Web writes bounded JSON records from fixed keys only.
- **Health (R1).**
  - Status codes, bodies and the 2 s deadlines are unchanged.
  - Intelligence health responses gain `Cache-Control: no-store`. All six scoped health responses carry `X-Request-Id`.
  - Timeouts and outages are measured with a monotonic client clock, with a bound of 5 s or less.
- **Contracts (R5).**
  - **Corpus:** one synthetic corpus with independently declared expectations, `contracts/conformance/intelligence/`.
  - **Go validator:** a new test tool, `santhosh-tekuri/jsonschema/v6` v6.0.3.
  - **Python validator:** the existing intelligence validator, which is corrected in two places where it silently diverges from Draft 2020-12:
    - `date-time` is never asserted, because `rfc3339-validator` is not installed;
    - Python `$` matches before a trailing newline.
  - **Cross-language peer checks:** Go-produced envelopes are checked by Python, and real fake-gateway NDJSON output from Python is checked by Go.
- **CI (R6/R7).**
  - A root `.gitlab-ci.yml` defines one gating job. It runs the existing harness through a new fail-closed entry point, `infra/ci/validate.sh`.
  - That entry point enforces evidence completeness, scans artifacts and withholds them on a hit, and checks the tree for drift after the run.
  - Controls V1–V12 cover the gate behaviour. The existing W1–W9 and C1–C12 controls are retained.
- **Not changed.** No database, schema, migration, public domain API or MCP change. No new route. No production execution client, telemetry vendor, metrics endpoint or deployment. No UI change.
  - `core/cmd/server/main.go`, `infra/compose/local.yml`, `core/go.mod`, the intelligence lockfile and web dependencies are unchanged.
  - The foundation barrier stays closed through G021.
- **Blocker.** The existing GitLab runner's execution profile cannot be read with the access available here. The harness needs a profile where the job shell and the Docker daemon share host loopback and filesystem paths, and it must not be privileged (§12). Codex or the owner must supply those facts before approval.

## 2. Verified baseline (2026-10-05, read-only)

| Item | Value |
|---|---|
| HEAD | `9461c4bdbd8c1ed3e9a9986c70647cd1add57101` ("Define G003 operational foundation and CI specification"), branch `staging` |
| Working tree | Clean at the start of the session: `git --no-optional-locks status --porcelain=v1 -z --untracked-files=all --ignored` was empty. Since 17:22 local time it shows one ignored, tool-local path, `!! .claude/settings.local.json` (§0), plus this Plan once placed. |
| Goal SHA-256 | `63f2b26676c9f1c385639c2078abc586beb48e36d6722c957aa4b03b70d73bb2` (matches handoff) |
| Spec SHA-256 | `d64a691d49219e6945b3fb115bcea59977a9b491f7e2746702e329d1c1c2e949` (matches handoff) |
| ARCHITECTURE.md SHA-256 | `6a6fd49b854ce90acf1a2ce6bd281a4deee3e68e8f4e021d8d737c3b04d90726` (matches handoff; §60 and §61 read) |
| Remote | `origin` `git@gitlab.example:<namespace>/<project>.git`. `ls-remote --symref`: `HEAD → refs/heads/main` (`99e2c79`, "Initial planning"), `refs/heads/staging` = `9461c4b` (equal to local HEAD). The default branch is `main`. |
| GitLab project | Private: the unauthenticated API returns 404. No `glab`, `gitlab-runner` or API token is available in this session, so runners, settings, pipelines and artifacts cannot be inspected here (§12). |
| Governance | SECURITY.md and DECISIONS.md are empty; there is no accepted ADR. BUDGET_POLICY, ROADMAP and PROGRESS were read. G001/G002 acceptance is recorded in PROGRESS and supersedes historical submission wording. |
| Schema digests (must remain unchanged) | `execution-envelope.schema.json` `bf3cec365e291da2cf9b201b9d0194632818721671a5844589954b978accba4f`; `stream-event.schema.json` `e1025031134253ac8d19a514f6d7a4d9563281b5006ebe9be40a84e4d63fd102` |
| Protected harness files | `run-approved.sh` `620a3e84…c53c695` and `run-approved-controls.sh` `314a16a7…7ba9` (both unchanged by G003); `run.sh` `d5e7e5be…41ac8`, `lib.sh` `9f7a3997…57c3` and `render-customizations.tsv` `9abf9e92…279f` (all three modified, §10) |
| `core/cmd/server/main.go` | `95aa5b23…7951`. It equals the rendered manifest and stays **unchanged**, so B-MCP-SOURCE's byte-equality guard needs no reconciliation. |
| Local Docker | Engine 27.5.1, Compose 2.32.4 (Docker Desktop on WSL2); 16 CPUs, 15 GiB RAM, 920 GB free. No `beginify` project containers, volumes or networks. Default ports 3100, 8180, 8100, 5532, 6479, 9100 and 9101 are free, as are the variant port sets. Unrelated developer resources exist (for example `payments-assistant-postgres-1`, `outreacher-*`, `needknow-drafter-*` and `atlas-*`/`harbor-*` images) and are never touched. |

## 3. Findings that shape the plan

| # | Area | Finding (inspected) | Consequence |
|---|---|---|---|
| F1 | Core correlation | `Middleware` accepts a single regex-valid request ID and a trace ID. It keeps no span or parent and logs the raw `path`. `r.Header.Get` reads only the first of duplicated headers. `statusRecorder` records the **last** `WriteHeader`. `ParseTraceparent` accepts any version except `ff`. | Rework inside `core/internal/observability` (§7.1). `main.go` is unchanged, because the middleware is already wired there. |
| F2 | Core → intelligence | `health.Intelligence` builds a fresh request with no diagnostic headers. Its 2 s client timeout equals core's 2 s readiness deadline. | Propagate from the request context (§7.1). No new route. |
| F3 | Web | The handlers take no `Request`. `coreReady()` sends no context and returns only a boolean. There is no completion logging. Next's standalone server prints only startup lines. | Add `web/lib/request-context.ts`; customize the two routes and `core-client.ts` (§7.2). |
| F4 | Intelligence logging | `uvicorn main:app` keeps uvicorn's default text access log, which writes the raw request line with its query string to stderr, and its default text error log, which writes tracebacks. `JsonFormatter` has no key or value redaction. `/healthz` and `/readyz` send no `Cache-Control`. | Pure-ASGI completion middleware scoped to the two health paths. Take over the `uvicorn.access` and `uvicorn.error` loggers inside `configure_logging`, so the Dockerfile `CMD` is unchanged. Add redaction and no-store (§7.3). |
| F5 | Python format assertion | `jsonschema` 4.26 registers the `date-time` checker only when `rfc3339_validator` imports. It is absent from `uv.lock`, so **`format: date-time` is silently not asserted**. `main.py` already compensates at runtime, but a conformance check would treat the keyword as success. | Register a project-owned RFC 3339 `date-time` checker, with no new dependency (§8.4). |
| F6 | Python `pattern` | `jsonschema` evaluates `pattern` with `re.search`. Python's `$` also matches before a final `\n`, so `"req-12345678\n"` passes `^[A-Za-z0-9._-]{8,64}$` in Python. It fails under ECMA-262 (JSON Schema's dialect) and under Go RE2. | An ECMA-anchored `pattern` keyword in the existing validator, with an audit of the supported syntax (§8.4, Q3). |
| F7 | Python JSON parsing | `json.loads` accepts `NaN` and `Infinity`; RFC 8259 and Go reject them. | Conformance parsing is strict. Runtime envelope parsing is unaffected, because an envelope has no free numeric fields (§8.4). |
| F8 | Go validator candidates | `google/jsonschema-go` (already indirect via the MCP SDK) ignores `format` during validation. `santhosh-tekuri/jsonschema/v6` v6.0.3 (2026-06-28, Apache-2.0) supports Draft 2020-12, `AssertFormat()`, `AddResource`/`UseLoader` and meta-validation, and implements `date-time` and `uuid`. Its `go.mod` requires `golang.org/x/text v0.14.0` (`regexp2` is test-only). | Use santhosh v6.0.3 in a separate test-tool module, so core's `go.mod` is untouched (§9). |
| F9 | Harness assumptions | The harness probes published ports on `127.0.0.1`, bind-mounts host paths (`$REPO`, `$WORK`, report directories) and refuses busy fixed port sets. | The runner must share loopback and filesystem with its Docker daemon. Docker-in-Docker and socket-only Docker executors do not meet this (§12). |
| F10 | Provenance | `contracts/README.md` and `local.yml` are rendered from templates (`.tmpl`) that differ in the intelligence-disabled render. Customizing them would break B-TOGGLE-OFF-ADDITIONS' equal-base rule. | Neither is changed. The corpus lives in `contracts/conformance/` and compose overlays are avoided (§14). |
| F11 | Nested deadlines | Web's `AbortSignal.timeout(2000)` starts before core's 2 s readiness deadline, which bounds its intelligence probe. A *hung* (paused) intelligence therefore makes core answer at about 2.0 s, and web usually aborts first, so web reports `core: unavailable`. A *stopped* intelligence fails fast and is unaffected. | Existing behaviour; the Spec preserves the deadlines. Assertions follow §6.3. Codex decision in Q4. |
| F12 | Ignored-file gate | G002's S0 gate rejects every ignored path. The Claude Code client's ignored `.claude/settings.local.json` now exists in the checkout. | One exact, reported exception in the G003 gate. Every other ignored path, `.env` and `.env.local` included, still fails (§15.1, demonstrations T1–T3). |

## 4. Diagnostic context contract (R2)

### 4.1 Request ID

- **Accepted inbound value.** Exactly one `X-Request-Id` field value matching `^[A-Za-z0-9._-]{8,64}$`, evaluated with full-string semantics:
  - Go RE2 `MatchString`;
  - JavaScript without the `m` flag;
  - Python `re.fullmatch`.
- **Rejected inbound value.** A missing value, more than one field line, a comma-joined value (Node joins duplicated fields as `a, b`), an over-long or non-matching value, or a value that contains a credential the receiving service holds (§4.4).
  - A rejected value is **never reflected or logged**.
  - Instead the service generates `request_id = 32 lowercase hex` from 16 random bytes (`crypto/rand`, `node:crypto` `randomBytes`, `secrets.token_hex`), and records `request_id_source` as `generated` rather than `inbound`.
- **Response header.** All six scoped health responses carry the effective `X-Request-Id`. Core already returns it on every response.

### 4.2 `traceparent` profile (narrower than W3C; no full-compliance claim)

Based on the [W3C Trace Context Recommendation](https://www.w3.org/TR/2021/REC-trace-context-1-20211123/).

- **Accepted.** Exactly one field value, at most 512 bytes before parsing, that fully matches `^00-([0-9a-f]{32})-([0-9a-f]{16})-([0-9a-f]{2})$`, where:
  - the trace ID is not all zeros;
  - the parent ID is not all zeros;
  - the length is exactly 55, which follows from the pattern.
- **Flags.** Any two lowercase hex digits are accepted. Only the `sampled` bit (`flags & 0x01`) is interpreted. Unknown bits are ignored and **never propagated**: the outbound flags are `01` when sampled is set, otherwise `00`. A newly generated trace uses `00`. The flag confers no authority, sampling, content capture or export.
- **Versions.**
  - `ff` is invalid.
  - **Every version other than `00` is unsupported in this profile.** The service restarts with a new local trace. This is narrower than W3C's SHOULD-parse rule for higher versions, and is documented as such.
  - Uppercase hex, the wrong length, extra trailing fields on `00`, internal whitespace, a non-hex character, a duplicated field or a comma-joined value are all invalid.
- **On an invalid value.** The service generates a new nonzero 16-byte trace ID. The invalid value is never reflected, logged or forwarded, no HTTP 500 is produced, and `trace_source` is recorded as `generated`.
- **Not supported:** `tracestate` and `baggage`. They are never read, forwarded or logged.

### 4.3 Spans and parent relationship (server-span-as-caller convention)

- **Per-hop span.** Every receiving service creates a fresh nonzero 8-byte `span_id` for its server operation. It regenerates if the value is all zeros, and it never copies the incoming parent ID.
- **Parent.** `parent_span_id` is the inbound `traceparent` parent ID, present only when that `traceparent` is valid.
- **Outbound.** Each outbound call carries `traceparent: 00-<trace_id>-<caller span_id>-<flags>`. The parent ID is "the ID of the current operation", as W3C requires when mutating `traceparent`. No separate client span is recorded.
- **Resulting chain on the full stack:**
  - `web.parent_span_id` = the caller's parent, absent when the context was generated;
  - `core.parent_span_id` = `web.span_id`;
  - `intelligence.parent_span_id` = `core.span_id`.

  All three records share `trace_id` and `request_id`, and their span IDs are distinct.
- **Datastores.** Probes to PostgreSQL, Redis and MinIO carry no tracing protocol.

### 4.4 Credentials supplied as diagnostic values (Q5)

- **Core** treats an inbound request ID that contains any configured secret value (`Config.SecretValues()`, reached through the redacting logger's handler) as invalid, and generates a new one. Core never echoes, logs or forwards that ID. Its logger additionally scrubs secret values from every string (existing behaviour).
- **Intelligence** does the same for `INTELLIGENCE_SERVICE_TOKEN` and `OPENROUTER_API_KEY`, and its formatter scrubs them (§7.3).
- **Web** holds no credentials, so it cannot classify a value as one. It logs only regex-bounded IDs. Core refuses to adopt or forward such a value.

### 4.5 Header allow-list and isolation

- **Outbound headers set by application code:** exactly `X-Request-Id` and `traceparent`.
  - Transport defaults remain: Go `User-Agent` and `Accept-Encoding`, and undici's standard request headers.
  - No inbound header, cookie, `Authorization`, `baggage` or `tracestate` is ever copied.
- **Request-local state:**
  - **Core:** values stored in the per-request `context.Context`.
  - **Web:** an explicit `RequestContext` argument. There is no module-level or `AsyncLocalStorage` state.
  - **Intelligence:** local variables inside the ASGI call. There is no `ContextVar`, because the health handlers make no outbound call.
- Isolation therefore holds by construction. It is still tested under concurrency, failure and cancellation (§16, AC2).

## 5. Log contract (R4)

### 5.1 Envelope for every application-owned JSON line (all three services)

| Key | Type | Rule |
|---|---|---|
| `time` | string | RFC 3339 UTC, millisecond precision, `Z`. Go: slog `ReplaceAttr`; Node: `toISOString()`; Python: `datetime.fromtimestamp(created, UTC)` |
| `level` | string | one of `debug`, `info`, `warn`, `error`. Python `WARNING` → `warn`, `CRITICAL` → `error` |
| `service` | string | `web`, `core` or `intelligence` |
| `event` | string | Core: slog's message key is renamed to `event`. Intelligence: `msg` becomes `event`; the existing `logger` key is kept. |

Startup and error records keep their messages (for example core's `core listening` and intelligence's `configured`) under `event`.

### 5.2 Completion record (`event = "http.request.completed"`)

| Key | Type | Rule |
|---|---|---|
| `method` | string | `GET`, `HEAD`, `POST`, `PUT`, `PATCH`, `DELETE` or `OPTIONS`; anything else becomes `OTHER` |
| `route` | string | Normalized, never the raw path:<br>• core: the matched `ServeMux` pattern path (`/healthz`, `/readyz`), else `unmatched`;<br>• web: the fixed handler route (`/api/health`, `/api/ready`);<br>• intelligence: `/healthz`, `/readyz` |
| `status` | integer | The final HTTP status: the first `WriteHeader`, or an implicit 200 on the first write. `0` only when the connection aborts before headers. |
| `duration_ms` | integer ≥ 0 | Monotonic elapsed time, floored |
| `request_id`, `request_id_source` | string, enum | §4.1 |
| `trace_id`, `trace_source` | 32 hex nonzero, enum | §4.2 |
| `span_id` | 16 hex nonzero | §4.3 |
| `parent_span_id` | 16 hex nonzero | **Present only** when the inbound parent was valid |
| `trace_flags` | `00` or `01` | The outbound flags (§4.2) |
| `outcome` | enum | `completed`, `cancelled` (client gone or context cancelled) or `failed` (panic or exception) |
| `error_type` | string `^[A-Za-z][A-Za-z0-9_.]{0,63}$` | Only when `outcome = failed`. Go: `panic`; Node: the error's `name`; Python: the exception class name. Never message text or a stack. |
| `core_probe` | enum | **Web `/api/ready` only:** `ok`, `not_ready`, `timeout` or `unreachable` |

- **Level.** `failed` → `error`; status 500 or above → `warn`; otherwise `info`.
- **Strictness.** No other key is allowed, apart from intelligence's `logger`.
- **Machine-checked contract.** The record is checked by `infra/validation/tools/internal/logrecord/completion.schema.json`, through the `logcheck` command (§9).

### 5.3 Exactly one record

- **Core.** The middleware logs once in a `defer` guarded by a flag.
  - A handler panic is logged with `outcome: failed`, `error_type: panic`. The middleware then re-panics with `http.ErrAbortHandler`, so net/http still aborts the connection, as it did before, but prints no stack.
  - A request whose context was cancelled is logged with `outcome: cancelled`.
- **Web.** `withCompletionLog` logs once in `finally`.
  - On an exception it logs `outcome: failed` and rethrows, preserving Next's 500 behaviour.
  - When `request.signal.aborted` is set, it logs `outcome: cancelled`.
- **Intelligence.** The ASGI middleware logs once in `finally`.
  - An exception becomes `failed` with its class name, then is re-raised.
  - `CancelledError` becomes `cancelled`, then is re-raised.
- **Scope.** Exactly the six health routes. Core keeps logging every request it receives, as it does today, including 404 and 405, with `route: unmatched`. Intelligence's `/internal/*` path keeps its execution export record and no longer has an access line (F4).
- **Duplicate writes.** The recorder forwards every call unchanged, as before, but reports the first status.

### 5.4 Framework output

| Service | Handling | Allowed non-JSON lines (exact patterns; any other line fails `logcheck`) |
|---|---|---|
| core | net/http has no access log. Panics are re-raised as `ErrAbortHandler`, which net/http does not print. | none |
| web | Next standalone prints only startup lines; route handlers print nothing. | `^ ▲ Next\.js [0-9]+\.[0-9]+\.[0-9]+$`, `^   - Local:        http://[^ ]+$`, `^   - Network:      http://[^ ]+$`, `^ ✓ Starting\.\.\.$`, `^ ✓ Ready in [0-9.]+m?s$`, the empty line. These are confirmed against observed output at S4; any deviation is recorded and reviewed, and the allow-list is never widened silently. |
| intelligence | `configure_logging` replaces the handlers on `uvicorn.error` with the JSON handler (`propagate=False`), and disables `uvicorn.access` (no handlers, `disabled=True`). The app is built inside `config.load()`, before uvicorn logs its first line. `Exception in ASGI application` therefore becomes a JSON record with the exception type only. | none |

### 5.5 Redaction (existing protections preserved; equivalents added)

- **Core (existing, retained):**
  - sensitive-key regex `password|passwd|secret|token|api_?key|authorization|cookie|dsn|credential`, including nested groups;
  - configured-secret-value scrubbing in the message and in string and non-string attributes.
- **Intelligence (new, in `JsonFormatter`):**
  - the same key regex, applied recursively to `fields`;
  - configured-secret-value scrubbing of `event` and of every string;
  - exception type only (existing);
  - string values capped at 512 characters.

  `JsonFormatter()` keeps its no-argument constructor, because `test_observability.py` uses it. Secrets are supplied through `configure_logging(level, secrets=...)`, which `create_app` calls.
- **Web.** Records are built only from the fixed keys in §5.2, with values that are validated IDs, enums or integers. A key-regex guard drops any sensitive-looking key. No URL, header, error message or configuration value is ever logged.
- **Never logged anywhere:**
  - raw headers, cookies or query strings;
  - request or response bodies;
  - prompts, completions, retrieved content or envelopes;
  - authorization references;
  - tenant or user identifiers;
  - URLs, DSNs or exception text.

## 6. Health semantics and bounds (R1)

### 6.1 Endpoints (bodies and status unchanged; additive headers only)

| Service | Route | Body (exact bytes) | Status | Headers after G003 |
|---|---|---|---|---|
| core | `GET /healthz` | `{"status":"alive"}\n` | 200 | `Content-Type: application/json`, `Cache-Control: no-store`, `X-Request-Id` (all existing) |
| core | `GET /readyz` | `{"status":…,"product_name":…,"checks":{"object_store":…,"postgres":…,"redis":…}[,"optional":{"intelligence":…}]}\n` | 200 / 503 | as above (existing) |
| web | `GET /api/health` | `{"status":"alive"}` | 200 | `Cache-Control: no-store` (existing), **`X-Request-Id` (new)** |
| web | `GET /api/ready` | `{"status":"ready"\|"unready","checks":{"core":"ok"\|"unavailable"}}` | 200 / 503 | `Cache-Control: no-store` (existing), **`X-Request-Id` (new)** |
| intelligence | `GET /healthz` | `{"status":"ok"}` | 200 | **`Cache-Control: no-store` and `X-Request-Id` (new)** |
| intelligence | `GET /readyz` | `{"status":"ready"}` (also when no model is configured) | 200 | **`Cache-Control: no-store` and `X-Request-Id` (new)** |

- Core's required checks remain `postgres`, `redis` and `object_store`. `intelligence` stays under `optional` and is absent when intelligence is disabled.
- The deadlines remain:
  - core readiness: 2 s;
  - core's intelligence client: 2 s;
  - web's fetch: 2 s.
- No error text, DSN, URL or credential appears in any body.
- Intelligence's body bytes are unchanged: FastAPI's default dict response and `JSONResponse` render identically (`separators=(",", ":")`).

### 6.2 Timing protocol

- **Clock.** Each single request is timed by `httpprobe` (§9) with Go's monotonic clock.
- **Where it runs.** Inside the target service's network namespace (`--network container:<cid>`): web on `127.0.0.1:3000`, core on `127.0.0.1:8180`, intelligence on `127.0.0.1:8000`.
- **Bound.** Every injected timeout or outage case must return in **≤ 5000 ms**, or the criterion fails.
- **Recovery.** Recovery polling is separate. It has a documented finite deadline of **90 s per case** and is never used to time a single request.

### 6.3 Outage and timeout matrix (disposable copies only)

The cases run on the enabled copy (`variants_ops`, ports 3200, 8280, 8200, 5632, 6579, 9200 and 9201) and on the disabled copy (`variants_off`).

| Case | Injection | Assertions (each single request ≤ 5 s) | Enabled | Disabled |
|---|---|---|---|---|
| O1–O3 | `compose stop` postgres / redis / minio | core `/readyz` 503 naming the dependency; web `/api/ready` 503 (`core_probe: not_ready`); both liveness probes 200; continuity records (§16 AC3); recovery to 200 | ✓ | ✓ |
| O4–O6 | `docker pause` postgres / redis / minio | Same statuses; web `core_probe` is `not_ready` or `timeout`, because core answers at about 2 s (F11); continuity in web and core; unpause, then recovery | ✓ | ✓ |
| O7 | stop intelligence | core 200 with `optional.intelligence = unavailable`; web 200; no intelligence record; recovery restores the three-hop chain | ✓ | — |
| O8 | pause intelligence | **Core `/readyz` (direct) 200 with `optional.intelligence = unavailable`.** Web `/api/ready` returns within the bound with either 200 or 503 `core_probe: timeout`. The observed value is recorded (F11, Q4). After unpause, intelligence may log 0 or 1 record for the queued probe; more than one, or a mismatched one, fails. | ✓ | — |
| O9 | stop core | web `/api/ready` 503 (`core_probe: unreachable`), `/api/health` 200, both no-store; only a web record | ✓ | ✓ |
| O10 | pause core | web `/api/ready` 503 (`core_probe: timeout`). After unpause, core logs at most one record for that request ID, with `outcome: cancelled` and matching IDs. | ✓ | ✓ |

The existing B-OUTAGE and B-PERSIST then run unchanged on the same copy, after recovery has been asserted.

## 7. Service changes

### 7.1 Core (`core/internal/observability`, `core/internal/health`; `main.go` unchanged)

- **`correlation.go` (customized).** It contains:
  - `type TraceContext struct{ RequestID, RequestIDSource, TraceID, TraceSource, SpanID, ParentSpanID, Flags string }`;
  - `parseRequestID(values []string, isSecret func(string) bool)` and `parseTraceparent(values []string)` (§4);
  - `ParseTraceparent(h)`, retained for existing callers and tests; it now applies the version-00 profile;
  - `RequestID`, `TraceID` and `SpanID`, plus `FromContext(ctx) (TraceContext, bool)`;
  - `InjectOutbound(ctx, http.Header)`, which sets exactly `X-Request-Id` and `traceparent`.

  The `Middleware` signature is unchanged. It now:
  - reads `Header.Values`;
  - stores the context;
  - sets `X-Request-Id`;
  - wraps the response writer, recording the first status and an implicit 200, forwarding every call unchanged and exposing `Unwrap`;
  - reads `r2.Pattern` after the mux has run, to get `route`;
  - logs once (§5.3), including on panic and cancellation.
- **`logging.go` (customized).** `NewLogger` (signature unchanged) adds `slog.HandlerOptions.ReplaceAttr`: UTC millisecond time, lowercase level, `msg` → `event`, and a `service=core` attribute. Redaction is unchanged. An unexported `secretMatcher(*slog.Logger)` exposes `containsSecret` to the middleware.
- **`checks.go` (customized).** `Intelligence` calls `observability.InjectOutbound(ctx, req.Header)` before `client.Do`. The 2 s client timeout and status handling are unchanged.
- **New tests:**
  - `core/internal/observability/context_test.go`:
    - request-ID and `traceparent` tables;
    - generation is nonzero and well-formed;
    - duplicated headers;
    - secret-valued IDs;
    - 64 concurrent requests through `httptest.Server`, with no bleed;
    - cancellation, which yields `outcome: cancelled` and leaves no state behind;
    - panic, which yields one record and `ErrAbortHandler`, with server error-log output captured and empty;
    - an implicit 200 versus explicit 503, and a duplicate `WriteHeader`, which records the first status while response behaviour is unchanged;
    - normalization of an unmatched route;
    - the completion record's keys and types;
    - an invalid raw value never appears in output;
    - handler output is byte-identical with and without the middleware, apart from `X-Request-Id`.
  - `core/internal/health/propagation_test.go`: a full `Middleware → Readiness → Intelligence` probe against an `httptest` double that records received headers.
    - The header set is exactly the allow-list, even when the inbound request carried `Cookie`, `Authorization`, `baggage`, `tracestate` and an arbitrary header.
    - `traceparent` carries the same trace, with parent = core's span.
    - A double that sleeps 5 s yields readiness at about 2 s, with `optional.intelligence` unavailable.

  The existing `observability_test.go` and `health_test.go` are unchanged and must still pass.

### 7.2 Web (routes and BFF; no UI file changes)

- **`web/lib/request-context.ts` (new, `import "server-only"`; no relative imports, so `node:test` can load it).** It exports:
  - `deriveContext(headers: Headers): RequestContext`;
  - `outboundHeaders(ctx): Record<string, string>`, which returns exactly two entries;
  - `withCompletionLog(route, handler)`, which returns `(request: Request) => Promise<Response>`. It times with `performance.now()`, sets `X-Request-Id` on the response, and writes one JSON line through `process.stdout.write`.
- **`web/lib/core-client.ts` (customized).** It becomes `coreReady(ctx, timeoutMs = 2000): Promise<CoreProbe>`, still with `cache: "no-store"` and `AbortSignal.timeout`, and with `headers: outboundHeaders(ctx)`.
  - `TimeoutError` → `timeout`; any other failure → `unreachable`; a response that is not OK → `not_ready`.
  - `import "server-only"` is kept, and `runtimeConfig()` is called exactly as now.
- **`web/app/api/health/route.ts` and `web/app/api/ready/route.ts` (customized).** `export const GET = withCompletionLog(...)`. The bodies, statuses and `Cache-Control` are byte-identical. `dynamic = "force-dynamic"` is kept.
- **`web/package.json` (already registered).** It gains `"test:unit": "node --conditions=react-server --test \"tests/unit/*.test.mjs\""`, and `"check"` becomes `npm run lint && npm run typecheck && npm run test:unit`. There are **no dependency changes**, so B-DEPS is untouched.
- **`web/tests/unit/request-context.test.mjs` (new).** `node:test` on Node 24 with type stripping, run in `web-check` with networking disabled. It covers:
  - the request-ID and `traceparent` tables, including comma-joined duplicates;
  - generated IDs and spans are nonzero;
  - `outboundHeaders` returns exactly two keys;
  - the completion record's keys and types;
  - one record when the handler throws (rethrown) and one when the signal is aborted;
  - 200 concurrent `withCompletionLog` invocations with distinct contexts produce no bleed;
  - an invalid value is never echoed.

### 7.3 Intelligence (`src/`, check stage only in the image)

- **`intelligence/src/observability/http.py` (new).** `CompletionLogMiddleware`, a pure ASGI class added with `app.add_middleware`. It is scoped to the exact paths `/healthz` and `/readyz`.
  - It reads the raw `scope["headers"]` and counts duplicates.
  - It applies §4, and rejects IDs that contain `secrets`.
  - It wraps `send` to capture the first status and to append `x-request-id`.
  - It logs once through `logging.getLogger("intelligence.http")`.
- **`intelligence/src/observability/export.py` (customized).**
  - `JsonFormatter`: the envelope (§5.1), `service`, redaction (§5.5) and a value cap.
  - `configure_logging(level=INFO, secrets=())`: idempotent; it updates the secrets on an existing handler, and takes over `uvicorn.error` and `uvicorn.access` (§5.4).
  - `ObservabilityHook` is unchanged, so content capture remains constructor-only and off by default.
- **`intelligence/src/main.py` (customized).**
  - `create_app` passes `(settings.service_token, settings.openrouter_api_key)` to `configure_logging` and the middleware.
  - The health handlers return `JSONResponse(..., headers={"Cache-Control": "no-store"})`.
  - `/internal/*` and envelope handling are unchanged.
- **`intelligence/src/streaming/schemas.py` (customized):** §8.4.
- **New tests**, all synchronous or `asyncio.run`, with `httpx.ASGITransport` (httpx is already a runtime dependency) and no new packages:
  - `tests/test_health_context.py`:
    - the context tables;
    - duplicated raw headers;
    - secret-valued IDs;
    - no-store and `X-Request-Id` on both routes;
    - bodies byte-identical to the baseline literals;
    - unconfigured models still `ready`;
    - 100 concurrent requests through `asyncio.gather`;
    - cancellation and exception paths, each producing exactly one record, re-raised;
    - non-health paths produce no completion record.
  - `tests/test_logging_privacy.py`:
    - canary secrets and sensitive keys in message, field, nested and exception paths;
    - header, query and body canaries on health requests;
    - a prompt and completion canary through `/internal/v1/execute` with `StubTransport`;
    - a `sampled=01` `traceparent` plus a forged "capture" header does not enable content capture;
    - uvicorn loggers are taken over (`uvicorn.access` disabled; `uvicorn.error` emits JSON with the type only);
    - a positive control: an unredacting formatter stub does leak, which proves the assertions are not vacuous.
  - `tests/test_conformance.py` and `tests/conformance_exchange.py`: §8.

  The existing tests are unchanged and must pass.
- **`infra/docker/intelligence.Dockerfile` (customized).** Only the `check` stage adds `COPY --from=contracts conformance/intelligence/ /srv/conformance/intelligence/`. The `runtime` and `base` stages are unchanged, so the corpus never enters the runtime image (asserted by B-CONTRACT-TOOLS: `/srv/conformance` is absent from the runtime image).

## 8. Contract conformance (R5)

### 8.1 Layout (fixtures only; `contracts/intelligence/*.schema.json` unchanged)

```text
contracts/conformance/README.md                         provenance, case IDs, how to run, stream-check rules
contracts/conformance/intelligence/expectations.json    independently declared outcome per case, schema digests, corpus rules
contracts/conformance/intelligence/cases/envelope/*.json
contracts/conformance/intelligence/cases/event/*.json
contracts/conformance/intelligence/cases/ndjson/*.ndjson
contracts/conformance/intelligence/schema-negative/*.json
```

- **Why this location.**
  - It is kept outside `contracts/intelligence/`, so it is not copied into the runtime image's `/srv/contracts/intelligence`.
  - It is absent from the intelligence-disabled render, which B-TOGGLE-OFF already requires.
  - It is fixture-only, consistent with "contract schemas' test/fixture/documentation areas" (Q11).
- **Expectation pinning.** `expectations.json` pins the two schema SHA-256 values from §2. Both validators refuse to run if the bytes they load differ.
- **Fixture data.** All data is synthetic, hand-authored for G003, and uses RFC-style UUIDs. Values never look like real credentials, so gitleaks' generic rules are not triggered; secret-scan is not loosened.

### 8.2 Case catalogue (file name = lowercase ID plus a slug; the expected result is fixed here)

**Envelope, valid (14):**

| ID | Case |
|---|---|
| EV01 | Minimal: required fields only, empty `allowed_tools` and `context_refs`, one message |
| EV02 | Full: `group_id`, `business_entity_id`, `callback_grant_ref`, all three tool levels, context refs, system/user/assistant messages |
| EV03 | `request_id` of length 8 |
| EV04 | `request_id` of length 64 with `.`, `_` and `-` |
| EV05 | Offset `+02:00` with a 9-digit fraction |
| EV06 | `2028-02-29` (leap year) |
| EV07 | `23:59:60Z` (leap second) |
| EV08 | Uppercase UUIDs |
| EV09 | 256 tools and 256 context refs |
| EV10 | 256 messages |
| EV11 | Every string maximum: content 200000, `authorization_ref` 512, `callback_grant_ref` 512, tool name 128, ref 256, kind 64 |
| EV12 | `trace_id` all zeros. The existing wire behaviour is preserved: the envelope schema has no nonzero rule, and none is imported. |
| EV13 | Nil UUID (shape-valid) |
| EV14 | `capability: embedding` |

**Envelope, invalid (50):**

| ID | Case |
|---|---|
| EI01 | Truncated object |
| EI02 | Trailing comma |
| EI03 | `NaN` literal |
| EI04 | Top-level array |
| EI05–EI14 | Each of the 10 required fields missing (in schema `required` order) |
| EI15 | `allowed_tools` as an object |
| EI16 | `request_id` as a number |
| EI17 | Unknown field |
| EI18 | `api_key` |
| EI19 | `password` |
| EI20 | `access_token` |
| EI21 | `authorization` |
| EI22 | `capability: vision` |
| EI23 | Tool level `admin` |
| EI24 | Extra property on a tool |
| EI25 | Role `tool` |
| EI26 | Empty messages |
| EI27 | 257 messages |
| EI28 | 257 tools |
| EI29 | Content of 200001 characters |
| EI30 | `request_id` of 7 characters |
| EI31 | `request_id` of 65 characters |
| EI32 | `request_id` with a space |
| EI33 | **`request_id` with a trailing `\n`** (F6) |
| EI34 | Uppercase `trace_id` |
| EI35 | `trace_id` of 31 characters |
| EI36 | `organization_id` not a UUID |
| EI37 | `user_id` of 35 characters |
| EI38 | `group_id` containing `g` |
| EI39 | **`2027-02-30T00:00:00Z`** |
| EI40 | **Month 13** |
| EI41 | **Hour 24** |
| EI42 | **Minute 60** |
| EI43 | **Second 61** |
| EI44 | **`12:00:60Z`** |
| EI45 | **`2027-02-29`** |
| EI46 | No offset |
| EI47 | Space separator |
| EI48 | 10-digit fraction |
| EI49 | Lowercase `t`/`z` |
| EI50 | **Offset `+24:00`** |

The bold cases EI39–EI45 and EI50 satisfy the `pattern` and fail **only** on `format: date-time`. They prove the format assertion is effective in both validators.

**Event, valid (15):**

| ID | Case |
|---|---|
| XV01 | `text_delta` with empty text |
| XV02 | `tool_call` with nested arguments |
| XV03 | `tool_result`, `is_error: false`, object content |
| XV04 | `tool_result`, `is_error: true`, null content |
| XV05 | `tool_result`, string content |
| XV06 | `citation` without locator |
| XV07 | `citation`, locator of 256 characters |
| XV08 | `structured_output` without `schema_ref` |
| XV09 | `structured_output` with `schema_ref` |
| XV10 | `usage`, all values |
| XV11 | `usage` with null `input_tokens`, `output_tokens` and `estimated_cost` |
| XV12 | `usage` with `units: 0` and zero tokens |
| XV13 | `usage` with tokens `1.0` (an integer-valued number) |
| XV14 | `error`, `retryable: true`, message of 1024 characters |
| XV15 | `done` |

**Event, invalid (34):**

| ID | Case |
|---|---|
| XI01 | Malformed JSON |
| XI02 | Unknown type |
| XI03 | Missing type |
| XI04 | Extra field |
| XI05 | Numeric text |
| XI06 | `done` without `request_id` |
| XI07 | `request_id` with a space |
| XI08 | `request_id` with a trailing `\n` |
| XI09 | Bad `call_id` |
| XI10 | Array arguments |
| XI11 | `tool_result` without `is_error` |
| XI12 | Empty citation ref |
| XI13 | String data |
| XI14 | Negative input tokens |
| XI15 | Negative output tokens |
| XI16 | Negative latency |
| XI17 | Negative cost |
| XI18 | Negative units |
| XI19 | Units `1.5` |
| XI20 | `usage` without `trace_id` |
| XI21 | Unknown usage capability |
| XI22–XI25 | `usage` with `organization_id`, `user_id`, `group_id`, `business_entity_id` (one case each) |
| XI26–XI30 | `usage` with the content fields `prompt`, `completion`, `messages`, `content`, `input` (one case each) |
| XI31 | Uppercase error code |
| XI32 | Error message of 1025 characters |
| XI33 | Extra field on `done` |
| XI34 | Uppercase `configured_provider` |

**NDJSON stream files (7).** Rules:
- every record is LF-terminated;
- no blank line;
- at least one record;
- every line parses strictly and validates against `stream-event`.

| ID | Case | Expected |
|---|---|---|
| NV01 | `text_delta`, `usage`, `done` | valid |
| NV02 | `error`, `done` | valid |
| NI01 | Final line truncated, no LF | invalid |
| NI02 | Non-JSON line in the middle | invalid |
| NI03 | Schema-invalid event in the middle | invalid |
| NI04 | Blank line | invalid |
| NI05 | Empty file | invalid |

These rules are test-profile rules for checking producer output. They make no wire-contract change.

**Schema-negative (7).** Each must make both validators **exit nonzero before any case is evaluated**:

| ID | Case |
|---|---|
| SN01 | Schema file missing |
| SN02 | Schema is malformed JSON |
| SN03 | Meta-invalid (`"type": 5`) |
| SN04 | Keyword outside the audited profile (`dependentSchemas`) |
| SN05 | Format outside the profile (`ipv4`) |
| SN06 | Remote `$ref` (`https://example.invalid/x.json`); must fail without network access |
| SN07 | `$schema` of draft-07 |

That is 120 document and stream cases plus 7 schema-negative cases. All eight event variants appear among the valid cases.

### 8.3 Validators (independent; neither reads the other's output)

- **Supported keyword profile.** The audit is shared by specification but implemented separately in each language: `$schema`, `$id`, `$defs`, `$ref` (local `#/$defs/...` only), `title`, `description`, `type`, `properties`, `required`, `additionalProperties`, `items`, `minItems`, `maxItems`, `minLength`, `maxLength`, `pattern`, `enum`, `const`, `oneOf`, `minimum`, `format`.
  - The only formats allowed are `date-time` and `uuid`.
  - Any other keyword or format in a loaded schema exits nonzero.
- **Go (`contractcheck`).** It uses santhosh v6.0.3:
  - `NewCompiler()` (Draft 2020-12), `AssertFormat()`;
  - `UseLoader`, whose loader fails for every URL (no remote resolution; meta-schemas are embedded);
  - `AddResource` for each schema under a `file:///contracts/...` URL;
  - compile-time meta-validation;
  - strict JSON decoding (RFC 8259; `NaN` is rejected);
  - RE2 patterns.

  Its subcommands are:
  - `corpus`: writes `results-go.json` and exits nonzero on any mismatch with the expectations or on a schema digest mismatch;
  - `ndjson`;
  - `produce-envelopes`: the test-only Go producer. It builds envelopes from typed Go structs: minimal, full, and a boundary case with a future deadline.
- **Python.** The existing `streaming.schemas` validator (§8.4), in the `intelligence-check` image, with strict JSON parsing (`parse_constant` raises). Two entry points:
  - `tests/test_conformance.py`: parametrized over every case in `expectations.json`. It runs in the normal `pytest -q` (README `platform:check`), and is asserted not to skip.
  - `tests/conformance_exchange.py`: a script, not collected by pytest. It has three subcommands:
    - `corpus`: writes `results-python.json`;
    - `peer`: validates `go-envelopes.ndjson`; posts each envelope through `create_app` with `StubTransport` (expects 200 and schema-valid events); and writes the **actual raw response bytes** from three fake-gateway scenarios (success, gateway error, capability not configured) to `python-events.ndjson`;
    - `ndjson`.
- **Shared result format.** `{"validator": "<name> <version>", "schema_sha256": {...}, "corpus_sha256": "...", "results": [{"id", "valid", "class"}], "counts": {...}, "skipped": 0}`. `corpus_sha256` is computed independently by each validator, and also by the harness over the repository files.

### 8.4 Python validator conformance corrections (`schemas.py`; schema bytes unchanged; Q3)

- **ECMA-262 `pattern` semantics.**
  - `validators.extend(Draft202012Validator, {"pattern": _ecma_pattern})`.
  - Each schema pattern is translated once: an unescaped `$` outside a character class becomes `\Z`.
  - Patterns containing a construct whose semantics differ between ECMA-262 and Python `re` are rejected when the validator is built: `\d`, `\w`, `\s`, `\b`, `\B`, `\p`, `(?`, backreferences and lookarounds. All current schema patterns use only literal classes, quantifiers, anchors, groups and alternation.
- **`date-time` checker.** A project-owned `FormatChecker` adds `date-time`: the RFC 3339 §5.6 grammar plus calendar validation (month length and leap years).
  - Hours 00–23, minutes 00–59, seconds 00–60, offset hours 00–23 and offset minutes 00–59.
  - Second 60 is allowed only when the UTC-normalized time is 23:59.
  - `T`/`Z` are accepted in either case; this is irrelevant in practice, because the schema pattern already requires uppercase.
  - `uuid` remains jsonschema's built-in checker.
- **Audit.** §8.3, plus an assertion that `date-time` and `uuid` are present in `checker.checkers`.
- **Effect at runtime.** Intelligence's `/internal/v1/execute` now rejects, with 422, two classes of envelope it previously accepted only through gaps:
  - an envelope whose pattern-bounded identifier ends in a newline;
  - an envelope whose `deadline` is calendar-invalid. These already received 422 through `main.py`'s parse fallback, with the identical response `{"path":"/deadline","keyword":"format"}`.
- **Leap-second note.** A valid leap-second deadline (`23:59:60Z`) remains schema-valid, but `main.py`'s `fromisoformat` rejects it, as it does today. That pre-existing runtime check is not changed and is noted in OPERATIONS.md.

### 8.5 Peer exchange (harness; scratch `$WORK/contracts-exchange`, mode 777, mounted into both containers)

1. `contractcheck produce-envelopes`, which writes `go-envelopes.ndjson`.
2. `conformance_exchange.py peer`. Python accepts and executes the Go envelopes, then emits `python-events.ndjson`.
3. `contractcheck ndjson --require-types text_delta,usage,error,done`, which checks `python-events.ndjson`.
4. The harness derives corrupt copies of both real files:
   - the last line truncated;
   - a garbage line inserted;
   - an invalid event or envelope appended;
   - a blank line.

   The peer checker must reject each one. A checker that ignores a bad line cannot pass this step.

## 9. Validation tools module (test tooling, not runtime)

- **Module:** `infra/validation/tools/`, `module validationtools`, `go 1.27.1`.
  - Direct requirement: `github.com/santhosh-tekuri/jsonschema/v6 v6.0.3`.
  - Indirect requirement: `golang.org/x/text`, at the version minimal version selection resolves (expected `v0.14.0`).
  - Both are recorded in `go.sum`. The module is resolved once at S2 in a container, with a scratch module cache in an execution-owned volume.
- **Commands:**
  - `cmd/contractcheck` (§8.3);
  - `cmd/httpprobe`: an HTTP client with repeatable and duplicate headers, header and body values read from files (so canaries never appear in arguments or logs), `--parallel N` with distinct generated contexts, `--abort-after`, `--timeout 10s`, and JSON output (status, monotonic `duration_ms`, `X-Request-Id`, `Cache-Control`, body SHA-256 and a body of up to 4 KiB);
  - `cmd/logcheck`: classifies lines (§5.4), validates records against `internal/logrecord/completion.schema.json` with the same santhosh library, enforces exactly one record per request and service, and verifies chain expectations from a cases file. It carries unit tests with negative controls: a planted non-JSON line, a duplicate record, a missing span, a broken parent and a mismatched trace must each fail.
- **Image:** `infra/docker/validation-tools.Dockerfile` (context `infra/validation/tools`):
  - `build`: `golang:1.27.1-alpine`, `go mod download && go mod verify`, `CGO_ENABLED=0 GOTOOLCHAIN=local -trimpath`;
  - `check`: `gofmt`, `go vet`, `go test ./...`;
  - `runtime`: `gcr.io/distroless/static-debian12:nonroot` with the three binaries.

  The `.dockerignore` excludes `.git`, `.env` and `.env.*`. The image tag is `g003-vtools-<run>`; it is labelled with the run and removed at cleanup.
- **Execution.** The tools run with `--network none`, except `httpprobe`, which joins a service's network namespace. They never join a production network, and they publish no port.
- **Exclusions.** None of this is linked into core. Core's `go.mod`, `go.sum` and binary imports are unchanged; B-MCP-SOURCE's `go list -deps` check additionally asserts that no `jsonschema` package is present.

## 10. Harness changes (`infra/validation/`; every existing assertion preserved)

| Check | Suite | Change |
|---|---|---|
| B-SHELL | static | Also shellchecks `infra/ci/validate.sh`, `infra/ci/validate-controls.sh` and `infra/ci/cleanup-owned.sh` |
| B-HADOLINT | static | Adds `validation-tools.Dockerfile` |
| B-DOCKERIGNORE | static | Adds `validation-tools.Dockerfile.dockerignore` (`.env` and `.env.*`) |
| B-DEPS | static | Adds the tools-module rule:<br>• exactly one direct requirement, `santhosh-tekuri/jsonschema/v6 v6.0.3`;<br>• `go.sum` present.<br>The existing rules are unchanged: goose and MCP pins; intelligence `uv.lock` and `pyproject.toml` equal to the render; web C1–C12 policy. |
| B-MCP-SOURCE | static | Unchanged assertions, plus "no `jsonschema` in `go list -deps ./cmd/server`" |
| B-WEB-BOUNDARY | static | Unchanged rules, plus:<br>• `lib/request-context.ts` imports `server-only` and is imported only by `app/api/**/route.ts` and `lib/core-client.ts`;<br>• `core-client.ts`'s fetch passes only `headers: outboundHeaders(ctx)` and does not reference `cookies`, `headers()` or an inbound `Request`.<br>Routes remain exactly `/`, `/api/health` and `/api/ready`. |
| **B-CI-CONFIG** (new) | static | §11.1 static assertions on `.gitlab-ci.yml` (through yq), plus the README `platform:ci` block |
| **B-CI-CONTROLS** (new) | static | `infra/ci/validate-controls.sh` V1–V12 (§11.3) |
| B-WRAPPER-CONTROLS, B-DEPS-CONTROLS | static | Unchanged (W1–W9; C1–C12, C8b, C8c) |
| B-RENDER-PLAN, B-RENDER-REPEAT | provenance | The register count changes from 8 to **18** (3 G001 + 5 G002 + 10 G003, §14.2). Exactly 18 conflicts and proposals. |
| **contracts suite** (new; root only; no stack) | contracts | • **B-CONTRACT-TOOLS:** `validation-tools` check stage passes; the intelligence-check image builds; the runtime intelligence image has no `/srv/conformance`.<br>• **B-CONTRACT-GO** and **B-CONTRACT-PY:** each validator matches every expectation, digests match, the audit passes, nothing is skipped.<br>• **B-CONTRACT-AGREE:** identical case-ID sets and verdicts between the two results and the expectations; per-category minima; all eight event types; three equal corpus digests.<br>• **B-CONTRACT-PEER:** §8.5.<br>• **B-CONTRACT-NEGATIVE:** SN01–SN07 in both validators, mounted read-only over the schema paths, each exiting nonzero for the intended reason; and one flipped expectation (EI39 valid) in a scratch copy makes both validators exit nonzero naming EI39. |
| **B-HEALTH-ROOT** (new) | root | Host probes on the actual root (ports 3100, 8180, 8100): six routes with status, exact body, no-store and a valid `X-Request-Id` |
| B-CHECKS-ROOT | root | Unchanged command; it now also exercises the new unit and conformance tests |
| **B-OBS-CONTEXT, B-OBS-CHAIN, B-OBS-TIMEOUTS, B-OBS-LOGS, B-OBS-PRIVACY** (new) | variants (`variants_ops`, before B-OUTAGE) | §16 AC1–AC4 on the enabled disposable copy |
| **B-OBS-OFF** (new) | variants (`variants_off`, before B-BROWSER-OFF) | Web → core continuity; no intelligence probe or record; O1–O6, O9 and O10; log schema and privacy on the disabled variant |
| B-TOGGLE-OFF-ADDITIONS | variants | `APPLICABLE` adds `core/internal/observability`, `core/internal/health`, `web/app/api`, `web/lib/core-client.ts`, `web/lib/request-context.ts` and `docs/foundation/OPERATIONS.md` (28 → 34) |
| B-SECRET-CANARY | final | Adds G003 canaries (§16 AC4) to the scan list. Canary request specs live in `$WORK/obs-inputs/`, which is excluded from the scan in the same way as `secrets*`, and are deleted after use. |
| **B-ARTIFACT-SCAN** (new) | final | Scans the staged report directory for generated secrets and canaries **before retention**. On a hit, every staged file is removed, a `WITHHELD` marker is left, and the check FAILs. A planted positive control is detected. |
| **B-REQUIRED-CHECKS** (new) | final | Every ID in `infra/validation/required-checks.tsv` for the executed suites is present and PASS. Missing, FAIL or UNMET → FAIL. |
| `all` | — | Becomes `static provenance contracts root migrations variants`. Required IDs: 50 existing + 17 new = **67** (§16 AC10). |
| cleanup | — | Also removes the run's `g003-vtools-*` images. The inventory diff is unchanged. |

**Unchanged:** B-SCOPE, B-PROTECTED, B-SKILL-UNCHANGED, B-LITERALS (which also covers the new files, so no identity literal or GitLab path appears in them), B-NAMING, the migration suite, B-PORTS-*, B-BOUNDARIES-* and their controls, B-PERSIST, B-RENAME, B-BROWSER-RENAME, B-DBRENAME, B-SOURCE-HASH, B-DOCKER-ONLY, both browser matrices (95 tests each), run-approved.sh and its controls.

## 11. GitLab CI (R6/R7)

### 11.1 `.gitlab-ci.yml` (exact intended content; `<RUNNER_TAG>` is set from Q1 before S7)

```yaml
# Baseline CI (docs/foundation/OPERATIONS.md). One gating job runs the repository's containerized
# validation entry point. There is no publish, release or deploy job, and no CI/CD secret is used.
workflow:
  rules:
    - if: $CI_PIPELINE_SOURCE == "merge_request_event"
    - if: $CI_PIPELINE_SOURCE == "push" && $CI_COMMIT_BRANCH == "staging"
    - if: $CI_PIPELINE_SOURCE == "push" && $CI_COMMIT_BRANCH == $CI_DEFAULT_BRANCH
    - when: never

stages: [validate]

validate:
  stage: validate
  tags: [<RUNNER_TAG>]
  timeout: 3h
  resource_group: platform-validation-host
  interruptible: false
  variables:
    GIT_STRATEGY: clone
    GIT_CLEAN_FLAGS: -ffdx
  script:
    - bash infra/ci/validate.sh --publish-dir ci-evidence
  after_script:
    - bash infra/ci/cleanup-owned.sh
  artifacts:
    when: always
    expire_in: 90 days
    access: developer
    paths: [ci-evidence/]
```

**B-CI-CONFIG asserts:**
- **Structure.** The top-level keys are exactly `workflow`, `stages` and `validate`. The four workflow rules are exact. There is exactly one job.
- **No bypass.** No `allow_failure`, `when: manual` or `when: delayed`, `rules` or `only`/`except` on the job, `rules:changes`, `needs`, `retry`, `parallel`, `trigger`, `environment`, `release`, `pages`, `image` or `services`.
- **Script.** It is exactly the single line above, and equal to the README `platform:ci` command.
- **Artifacts.** Exactly as shown: `when: always`, one directory, `access: developer`, `expire_in: 90 days`, no `untracked`.
- **Variables.** Exactly the two shown. No secret-like names.
- **Limits.** `timeout` is 3h or less. `tags` equals the approved tag.

**Pipeline behaviour.**
- **Pipeline sources.**
  - Merge-request pipelines test the MR **source** content (`detached`), or the merged result when the project enables merged results.
  - Pushes to `staging` and to the default branch (`main`) also run.
  - Feature-branch pushes, tags, schedules, web triggers and API triggers run nothing.
- **Duplicates.** A push to `staging` while an MR from `staging` is open yields two gating pipelines for the same commit. This is accepted.
- **Serialization.** `resource_group` serializes this project's validation jobs on the shared host ports.

### 11.2 `infra/ci/validate.sh` (aggregate entry point; Bash, `set -euo pipefail`, `inherit_errexit`; no masking construct)

```text
validate.sh [--evidence-dir DIR] [--publish-dir ci-evidence]
```

1. **Evidence location.**
   - Without `--evidence-dir`, the evidence directory is `mktemp -d` under `TMPDIR`.
   - The directory must resolve outside the repository: `realpath -m` with the `inside_repo` check, as in the S0 gate, done before it is created.
   - `--publish-dir` accepts only the literal `ci-evidence`, and is written last (step 8).
2. **Metadata (allow-listed fields only; never an environment dump).**
   - `git --no-optional-locks rev-parse HEAD` and `HEAD^{tree}`.
   - `git log -1 --format='%H %an <%ae> %cI'`, as attribution.
   - These CI variables when present: `CI_PIPELINE_ID`, `CI_PIPELINE_IID`, `CI_PIPELINE_URL`, `CI_JOB_ID`, `CI_JOB_URL`, `CI_PIPELINE_SOURCE`, `CI_COMMIT_SHA`, `CI_COMMIT_BRANCH`, `CI_DEFAULT_BRANCH`, `CI_MERGE_REQUEST_IID`, `CI_MERGE_REQUEST_EVENT_TYPE`, `CI_MERGE_REQUEST_SOURCE_BRANCH_SHA`, `CI_MERGE_REQUEST_TARGET_BRANCH_SHA`, `CI_MERGE_REQUEST_DIFF_BASE_SHA`, `CI_MERGE_REQUEST_APPROVED`, `GITLAB_USER_LOGIN`, `CI_RUNNER_ID`, `CI_RUNNER_DESCRIPTION`, `CI_RUNNER_TAGS`, `CI_SERVER_VERSION`.
   - Docker and Compose versions, and UTC start time.
   - `tested_content` is one of:
     - `mr-source`: `detached` and HEAD equals the source SHA;
     - `mr-merged-result`;
     - `branch-push`;
     - `local`.

     A merged-result run is never implied when only source content ran.
3. **Precheck.** `git status --porcelain=v1 -z --untracked-files=all --ignored` must be empty: the clean checkout. A digest of the tracked tree is recorded.
4. **Harness, with its status captured explicitly (`set +e`, `rc=$?`, `set -e`):**

   ```text
   infra/validation/run.sh --base HEAD --allow-file infra/ci/unchanged.allow --report-dir "$EV/raw" all
   ```

   `unchanged.allow` contains comments only, so B-SCOPE and B-PROTECTED assert that the run changed nothing relative to the checked-out commit. Goal scope is enforced separately, by the Goal's own final run.
5. **Evidence completeness:**
   - `results.tsv` holds every required ID (67) as PASS, with no FAIL or UNMET;
   - `browser-root-results.json` and `browser-off-results.json` exist with stats of more than 0 expected tests and 0 skipped, 0 flaky and 0 unexpected;
   - contract results show `skipped: 0`;
   - `inventory-leftover.txt` is empty, except the documented renderer cache volume `project-archetype-render-cache` on a first run.

   Otherwise → **14**.
6. **Staging and scan.**
   - Only allow-listed names are copied to `$EV/staged`: `results.tsv`, `images.txt`, `required-checks.tsv`, `browser-{root,off,rename}-tests.tsv`, `browser-*-results.json` (with every `attachments` entry removed, so base64 cannot hide a value), `deps-controls.tsv`, `wrapper-controls.log`, `ci-controls.log`, `contracts-results-{go,python}.json`, `contracts-agree.tsv`, `obs-*.tsv`, `obs-timings.tsv`, `plan.out`, `repeat-render.out`, `catalog-p.log`, `migrate-it.log`, `deps-server.txt`, `inventory-leftover.txt`, plus `summary.json`.
   - The staged set is scanned with the pinned gitleaks image, using the repository config, `--redact`, `--no-banner`, `--exit-code 1` and network none. A forbidden-name check also runs (`.env`, `*.env`, `secrets*`, `exec-token`).
   - On any finding, everything except a `WITHHELD` marker and a minimal `summary.json` is deleted, and the minimal summary is re-scanned → **15**.
   - The harness's own B-ARTIFACT-SCAN has already scanned for generated secrets and canaries inside its lifetime.
7. **Postcheck, always run.** The status and tracked-tree digest must be unchanged → otherwise **13**. The result is printed in every case.
8. **Publish (CI only).** The scanned staged set is copied to `$CI_PROJECT_DIR/ci-evidence/`. This happens after the postcheck, so it cannot affect drift detection.
9. **Exit.**
   - The harness status, if it is nonzero. A passing postcheck or scan can never replace a failing harness, and drift and scan results are still reported.
   - Otherwise 15, then 14, then 13, in that order.
   - Otherwise 0.
   - 2 means usage; 16 means a metadata or preflight failure.

**`cleanup-owned.sh` (`after_script`).**
- The harness appends every Compose project name and image tag it creates to a job-owned registry file, `$TMPDIR/platform-validation-owned.txt`, which validate.sh points it to.
- The cleanup script runs `docker compose -p <name> down -v --remove-orphans` and `docker image rm` **only** for entries in that file, then deletes the file.
- It never prunes and never touches any other resource. It always exits 0, and it cannot change the job status, which is already determined.

### 11.3 CI controls V1–V12 (`validate-controls.sh`; scratch Git repository outside the checkout; stub harness; no switch in the normal pipeline)

The controls create a disposable `git init` repository in `mktemp -d`, containing a copy of `validate.sh`, a stub `infra/validation/run.sh`, `required-checks.tsv`, `unchanged.allow` and one tracked input. They commit it with a fixed throwaway identity. The stub writes `results.tsv` and reports as instructed, and can mutate the tree.

| Control | Scenario | Expected |
|---|---|---|
| V1 | All valid, complete PASS | exit 0; staged set published |
| V2 **(a)** | One service-check FAIL row (B-CHECKS-ROOT), harness exit 1 | exit 1; summary names it; artifacts staged and scanned |
| V3 (a) | Harness exit 0 but a FAIL row present | 14 |
| V4 **(b)** | **Real tooling:** a scratch copy of the current implementation (`copy_repo`, then `git init`, then a scratch commit) with EI39 flipped to `valid: true`. Real `run.sh … contracts` exits 1 with B-CONTRACT-GO, B-CONTRACT-PY and B-CONTRACT-AGREE FAIL. The resulting `results.tsv` is fed through V2's path. | 1, for that reason |
| V5 **(c)** | The stub writes a throwaway credential-shaped canary (generated at run time; no literal in the repository) into a raw report | 15; staged set holds only `WITHHELD` and a summary; the canary is absent from the staged and published directories |
| V6 (c) | The harness-level `stage_and_scan` function, given a planted generated-secret canary | B-ARTIFACT-SCAN logic withholds and fails |
| V7 **(d)** | Harness exit 1, clean tree | 1; the postcheck passed and is reported |
| V8 **(e)** | Harness exit 0 but modifies a tracked file | 13 |
| V9 (e) | Harness exit 0 but creates an untracked file | 13 |
| V10 **(f)** | Exit 0, one required ID missing | 14 |
| V11 (f) | Exit 0, UNMET present, or browser JSON with `skipped: 1` | 14 (two sub-cases) |
| V12 | Harness exit 1 **and** drift | 1, with the drift reported |

No case can turn a failure into exit 0. W1–W9 (run-approved.sh) and C1–C12 remain separate.

## 12. Runner profile and hosted evidence (BLOCKER; R6/R7, AC7)

### 12.1 What the harness requires (from F9)

**Profile A (required unless Codex and the owner decide otherwise).** A self-managed GitLab runner using the **shell executor**, on a **dedicated, non-production Linux host**:

- **Docker:** Docker Engine 24 or later with BuildKit, and the Compose v2 plugin 2.24 or later.
- **Host tools:** bash 4.4 or later (`inherit_errexit`), GNU coreutils/findutils/grep with `-P`, git, `timeout`, `hostname -I` or `ip`, and bash `/dev/tcp`.
- **Job user:** the job runs as the runner user in the `docker` group. That is root-equivalent on this host. It is acceptable only because the host is dedicated and disposable in purpose, and it must be acknowledged.
- **Isolation:** no other workload's Docker resources on that daemon.
- **Concurrency:** one concurrent job for this project. `resource_group` enforces it inside the project; the runner `limit`/`concurrent` value, or dedication, covers other projects.
- **Capacity:** at least 4 vCPU (8 recommended), at least 8 GiB RAM (16 recommended), and at least 40 GiB free for images and the BuildKit cache.
- **Timeouts:** the runner's `maximum_timeout` and the project's CI timeout must each be at least 3 h.
- **Egress** to the registries and module proxies used in builds: Docker Hub, `gcr.io`, `ghcr.io`, `mcr.microsoft.com`, `proxy.golang.org`/`sum.golang.org`, `registry.npmjs.org` and PyPI. The Docker Hub anonymous pull limit must be acknowledged.

**Profile A2 (acceptable only if already configured this way; no configuration change is made by G003).** A Docker executor that mounts the host Docker socket, with job-container `network_mode = "host"`, and the builds and temporary directories mounted at **identical host paths**. Without all three, the bind-mount and loopback probes cannot work.

**Not acceptable without an explicit security decision (stop condition).**
- GitLab.com hosted runners or any `docker:dind` service. These need `privileged = true`, and published ports would not be on the job's loopback.
- A socket-only Docker executor without host networking and path parity.
- Any runner that can reach production resources or credentials.

### 12.2 Facts needed before approval (owner-readable in GitLab, Settings → CI/CD → Runners; Q1)

1. **Runner:** the ID and description; whether it is a project, group or instance runner; whether it is self-managed or hosted; its **executor**; and whether it is **privileged**.
2. **Selection:** its tags, whether it runs untagged jobs, and whether it is restricted to protected branches. If it is, the evidence branch must be protected, or a non-protected runner is needed.
3. **Host:** OS and architecture (linux/amd64 is assumed by the pinned images); Docker Engine, Compose and bash versions.
4. **Isolation and capacity:** whether the host is dedicated; other workloads; `concurrent`/`limit`; CPU, RAM and disk; `maximum_timeout`; and the project's CI/CD timeout.
5. **Network:** egress to the registries above, and any pull-through cache or proxy.
6. **Artifacts:** the instance's maximum artifact size, and the project's "keep latest artifacts" setting.
7. **Merged results:** whether merged-results pipelines are enabled (this sets `tested_content`).
8. **Access:** who can read pipelines, jobs and artifacts, and how Claude will retrieve hosted evidence (Q2).

**Read-only verification at the first hosted run.** `validate.sh` records the CI metadata. The harness PREFLIGHT records the Docker and Compose versions and the probe mode. The job log shows the runner. The **first-run duration** (cold cache) and a later warm run's duration are recorded against the 3 h timeout. If the cold run exceeds about 70% of the timeout, it is reported to Codex before relying on it.

### 12.3 Hosted evidence plan (each step needs separate, explicit authorization; Q2)

| Step | Action | Expected evidence |
|---|---|---|
| H0 (recommended, optional) | After S7 passes locally: commit the work in progress on branch `g003-foundation-observability-ci`, push it, and open a **draft** MR to `staging` | An early runner check: a passing or failing pipeline that shows profile compatibility before the final run. A failure caused by the runner environment is recorded, and its corrective work counts in B7. |
| H1 | Push the final implementation commit, plus the Result as available, to the same branch; the draft MR's pipeline runs | A passing pipeline on the **exact submitted commit and tree**. Recorded: pipeline ID, job ID, runner, `tested_content`, duration, the artifact (`summary.json` with 67 PASS) and its download digest. |
| H2 | Branch `g003-gate-proof` = H1's commit plus one commit flipping EI39 to `valid: true` in `expectations.json`; a draft MR to `staging` | A **failed** pipeline. B-CONTRACT-GO, B-CONTRACT-PY and B-CONTRACT-AGREE FAIL; `validate.sh` exits 1; artifacts uploaded scanned, showing the reason. It exercises the reviewed gate, not an `exit 1` job. |
| H3 | After Codex has captured the evidence: close H2's MR and delete `g003-gate-proof`. Keep H1's MR for review. | The deletion is recorded |

- **No step** pushes to `staging` or `main`, publishes, deploys or creates CI/CD variables.
- **Retrieval.** Hosted logs and artifacts are retrieved by the owner downloading them to a path outside the repository, or through a read-only token (`read_api`) supplied out of band, never printed or stored in the repository (Q2).
- **Evidence classification.** Local and hosted evidence are kept apart in the Result. **Local parity never substitutes for H1 or H2.**

## 13. Dependencies, images and provenance

| Item | Pin | Where | Notes |
|---|---|---|---|
| `github.com/santhosh-tekuri/jsonschema/v6` | `v6.0.3` (2026-06-28, Apache-2.0) | `infra/validation/tools/go.mod` (new) | Test tooling only. `go.sum` is recorded; `go mod verify` runs in every build. |
| `golang.org/x/text` | resolved by MVS (expected `v0.14.0`) | tools `go.mod`, indirect | Recorded in the Result |
| `golang:1.27.1-alpine`, `gcr.io/distroless/static-debian12:nonroot` | existing tags | `validation-tools.Dockerfile` | Same tag-only pins as the existing images (inherited residual). No new base image. |
| `zricethezav/gitleaks:v8.30.1` | existing | `validate.sh` artifact scan | Already used by `secret-scan` |
| `koalaman/shellcheck:v0.11.0`, `hadolint/hadolint:v2.15.1`, `mikefarah/yq:4.48.1`, `busybox:1.37.0`, Playwright `v1.63.0-noble@sha256:eff16c30…a4a27` | existing | harness | Unchanged |

**No change** to:
- web `package.json` dependencies or `package-lock.json`;
- core `go.mod`/`go.sum`;
- intelligence `pyproject.toml` dependencies or `uv.lock`;
- any runtime base image.

**No new vendor, service or infrastructure.** santhosh is a validation library used by a test tool, analogous to the Playwright image accepted in G002's Q4 (Q12).

## 14. Files (exact; `management/v0/results/G003-scope-allowlist.txt` mirrors this section)

### 14.1 Database, schema and contracts

- **None** of: migrations, seeds, tables, domain identities, public routes, API or MCP contracts. B-MIGRATE-SET stays `{00001_baseline.sql}`.
- The two schemas are byte-identical (§2 digests are asserted by both validators).
- Response headers are additive only: `X-Request-Id` on the web and intelligence health routes, and `Cache-Control: no-store` on intelligence health.

### 14.2 Customized rendered files (10 new register rows → 18 total)

| Path | Register reason |
|---|---|
| `core/internal/observability/correlation.go` | diagnostic context profile, per-hop spans and completion records (G003) |
| `core/internal/observability/logging.go` | JSON record envelope (event, service, UTC time) on the redacting logger (G003) |
| `core/internal/health/checks.go` | propagates diagnostic context on the optional intelligence probe (G003) |
| `web/app/api/health/route.ts` | request context and completion record on liveness (G003) |
| `web/app/api/ready/route.ts` | request context, propagation and completion record on readiness (G003) |
| `web/lib/core-client.ts` | forwards diagnostic context to core readiness (G003) |
| `intelligence/src/main.py` | health completion middleware and no-store health responses (G003) |
| `intelligence/src/observability/export.py` | JSON envelope, redaction and framework-logger takeover (G003) |
| `intelligence/src/streaming/schemas.py` | Draft 2020-12 conformance: ECMA-262 pattern anchoring, date-time assertion, keyword audit (G003) |
| `infra/docker/intelligence.Dockerfile` | conformance corpus in the check stage only (G003) |

**Already registered, content changes:** `web/package.json` (scripts only) and `docs/foundation/FOLLOW-UP.md`.

### 14.3 New files

- **Core (2):** `core/internal/observability/context_test.go`, `core/internal/health/propagation_test.go`.
- **Web (2):** `web/lib/request-context.ts`, `web/tests/unit/request-context.test.mjs`.
- **Intelligence (5):** `intelligence/src/observability/http.py`, `intelligence/tests/test_health_context.py`, `intelligence/tests/test_logging_privacy.py`, `intelligence/tests/test_conformance.py`, `intelligence/tests/conformance_exchange.py`.
- **Contracts (128):**
  - `contracts/conformance/README.md`;
  - `contracts/conformance/intelligence/expectations.json`;
  - 126 fixture files named `cases/envelope/<id>-<slug>.json` (EV01–EV14, EI01–EI50), `cases/event/<id>-<slug>.json` (XV01–XV15, XI01–XI34), `cases/ndjson/<id>-<slug>.ndjson` (NV01–NV02, NI01–NI05) and `schema-negative/<id>-<slug>.json` (SN02–SN07; SN01 is an absent file, so it has no fixture).

  The allow-list uses one anchored expression per directory, constrained to the declared ID ranges, for example `^contracts/conformance/intelligence/cases/envelope/(ev(0[1-9]|1[0-4])|ei(0[1-9]|[1-4][0-9]|50))-[a-z0-9-]+\.json$`. B-CONTRACT-AGREE requires the file set to equal the `expectations.json` entries exactly.
- **Validation tools (17):**
  - `infra/validation/tools/{go.mod,go.sum}`;
  - `infra/validation/tools/cmd/{contractcheck,httpprobe,logcheck}/main.go` and their `main_test.go` (6);
  - `infra/validation/tools/internal/conformance/{validate.go,audit.go,ndjson.go,produce.go,conformance_test.go}`;
  - `infra/validation/tools/internal/logrecord/{completion.schema.json,logrecord.go,logrecord_test.go}`;
  - `infra/docker/validation-tools.Dockerfile`, `infra/docker/validation-tools.Dockerfile.dockerignore`.
- **CI (5):** `.gitlab-ci.yml`, `infra/ci/validate.sh`, `infra/ci/validate-controls.sh`, `infra/ci/cleanup-owned.sh`, `infra/ci/unchanged.allow`.
- **Harness (1):** `infra/validation/required-checks.tsv`.
- **Docs (1):** `docs/foundation/OPERATIONS.md`. It covers:
  - health semantics and bounds, and F11;
  - the context profile, convention and allow-list;
  - the log contract and framework handling;
  - privacy rules;
  - the conformance corpus and the Python corrections;
  - CI: rules, runner requirements, artifacts, retention and access, gates, controls, branches;
  - inherited versus new guarantees versus remaining work (metrics and exporters explicitly open);
  - a reference to ARCHITECTURE §60 hostname guidance, as guidance only.
- **Management (2, after execution):** `management/v0/results/G003-foundation-observability-ci-result.md`, `management/v0/results/G003-scope-allowlist.txt`.

### 14.4 Modified project-owned files

- `infra/validation/run.sh`, `infra/validation/lib.sh`, `infra/validation/render-customizations.tsv` (+10 rows).
- `README.md`:
  - health lists all six routes and the headers;
  - new blocks `platform:contracts` and `platform:ci`;
  - a link to OPERATIONS.md;
  - the existing blocks are unchanged.

### 14.5 Not touched

- `core/cmd/**`, `core/internal/{config,mcp,migrate}/**`, `core/migrations/**`, `core/go.{mod,sum}`.
- `core/internal/health/{health.go,health_test.go}`, `core/internal/observability/observability_test.go`.
- `web/app/{layout.tsx,page.tsx,globals.css}`, `web/components/**`, `web/lib/{config,appearance,tokens,utils}.ts`, `web/package-lock.json`, all web configuration files, `web/tests/browser/**`, `web/tests/fixtures/**`.
- `intelligence/{pyproject.toml,uv.lock}`, every other `intelligence/src/**` file, the existing `intelligence/tests/*.py`.
- `contracts/README.md`, `contracts/intelligence/**`.
- `infra/compose/**`, every other Dockerfile and `.dockerignore`, `infra/{local,minio,postgres,security}/**`.
- `infra/validation/run-approved{,-controls}.sh`.
- `.env.example`, `.gitignore`, `.archetype/**`.
- `skills/**`, AGENTS.md, CLAUDE.md, `docs/ARCHITECTURE.md`, SECURITY.md, DECISIONS.md, `docs/history/**`, the other `docs/foundation/*.md`.
- ROADMAP, BACKLOG, the G003 Goal and Spec, PROGRESS (Codex-owned), and every G000–G002 artifact.
- The tool-local `.claude/settings.local.json`: never read or modified.

### 14.6 Temporary writes (none persist)

| Write | Location | Removal |
|---|---|---|
| S0 gate copy and evidence | outside the repository (`<ext>/g003-exec/`) | kept as evidence outside the repository |
| Go module cache for tools-module resolution (S2) | Docker volume `g003-gomod-<exec>` (label `g003.exec=<exec>`) | removed after label verification |
| Component-check images | `g003-<exec>-{core,web,intelligence,vtools}-check` | removed at the end |
| Harness work, compose projects, images, root `.env` | as today (`project-validation-*`, `g001-*-<run>`, root `.env` created and hash-removed by the root suite) | the harness's own cleanup |
| Local CI-parity clone (S9) | `<ext>/g003-ci-parity/` with a scratch commit (Q6) | deleted after the run |
| Hosted evidence downloads | `<ext>/g003-hosted/` | kept outside the repository |

## 15. Execution sequence, gates and provenance

### 15.1 S0: read-only, fail-closed starting-state gate (adapted from approved G002 r6 S0a)

- **Reuse.** The committed wrapper `run-approved.sh`, with `--check-only`, already provides the snapshot-digest, format, count and entry checks, so G003 needs no bootstrap writes. S0 adds what the wrapper does not check:
  - the HEAD binding;
  - the NUL-safe starting state;
  - the resource and port inventory;
  - evidence containment.

  It keeps G002 r6's read-only contract: `realpath -m` containment before any `mkdir`, and Git with `--no-optional-locks -c core.fsmonitor=false`, `GIT_OPTIONAL_LOCKS=0` and redirection variables unset.
- **Running it.** The script runs from **outside the repository** after byte-exact extraction from the block below. Its SHA-256 must equal **`0409148a630f32cc79642fd8ae65b90a53001eba379775bd96a3c1f313a66723`**; a different hash means stop. BASE, SNAPSHOT, SNAPSHOT_SHA256 and COUNT are taken verbatim from Codex's approval record. Nothing is regenerated, and the G002 snapshot is never reused.

```bash
bash <ext>/g003-s0-gate.sh <repo> <BASE> <approved snapshot> <its SHA-256> <COUNT> <ext>/g003-exec/s0
```

| Step | Verifies | Exit |
|---|---|---|
| 0 | Six arguments. BASE is 40 hex, the digest is 64 hex, COUNT is 1–5. The repository resolves. The evidence path resolves **outside** the repository, including through traversal and symlinks (dangling ones included), **before any write**. | 26 |
| 1 | `HEAD` = BASE | 20 |
| 2 | The snapshot is read once; the SHA-256 of those bytes equals the pinned digest | 21 |
| 3 | Only those bytes are parsed: `<64hex>  <path>`, no escaped entries, relative, no `.` or `..`, unique, inside the G003 input set (Goal, Spec, PROGRESS.md, this Plan, ARCHITECTURE.md), exactly COUNT entries, and the Plan present | 22 |
| 4 | Every entry is a regular non-symlink file whose SHA-256 matches. This holds whether or not the entry is committed. | 23 |
| 5 | NUL-safe `status --ignored`: every differing path is `??` or ` M` **and** is a snapshot path. Nothing staged, deleted, renamed, ignored (including `.env` and `.env.local`) or extra is allowed. **The single exception** is the exact ignored path `.claude/settings.local.json`, which is counted and reported, never read (F12). | 24 |
| 6 | No root `.env`; Compose 2.24 or later; no containers, volumes or networks for the project; default ports free; the Docker inventory written to evidence | 25 |

**Exit 0 is the only way to proceed.** Any other exit stops the work before the first repository write.

Exact S0 script (SHA-256 `0409148a630f32cc79642fd8ae65b90a53001eba379775bd96a3c1f313a66723`; extract the lines between the fences, including the final newline):

```bash
#!/usr/bin/env bash
# G003 S0: read-only, fail-closed starting-state gate (G003 Plan §15.1, r1). Adapted from the
# approved G002 r6 S0a gate with the same read-only contract. Needs only bash (>= 4.4), GNU
# coreutils, git and docker. Writes nothing inside REPO, including .git: Git runs without optional
# locks, so the index is never refreshed. The evidence destination is resolved and proven outside
# REPO before anything is created.
# Usage: bash g003-s0-gate.sh REPO BASE SNAPSHOT SNAPSHOT_SHA256 COUNT EVIDENCE_DIR
#   BASE, SNAPSHOT, SNAPSHOT_SHA256 and COUNT are the values pinned by Codex's Plan approval.
# Exit: 0 pass; 20 HEAD; 21 snapshot digest; 22 snapshot format/paths/count; 23 input hash;
#       24 starting state; 25 local resources or tools; 26 usage or evidence path.
set -euo pipefail
shopt -s inherit_errexit

fail() { printf 'S0 FAIL %s: %s\n' "$1" "$2" >&2; exit "$1"; }
[[ $# -eq 6 ]] || fail 26 "usage: REPO BASE SNAPSHOT SNAPSHOT_SHA256 COUNT EVIDENCE_DIR"

# Git inspection must not write repository metadata (no index refresh, no fsmonitor) and must not
# be redirected by the caller's environment.
export GIT_OPTIONAL_LOCKS=0
unset GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE GIT_OBJECT_DIRECTORY GIT_ALTERNATE_OBJECT_DIRECTORIES \
  GIT_COMMON_DIR GIT_NAMESPACE GIT_CONFIG_PARAMETERS GIT_CONFIG_COUNT
g() { git --no-optional-locks -c core.fsmonitor=false -C "$REPO" "$@"; }

# 0. Arguments, and every location resolved before any write.
REPO="$(realpath -e -- "$1" 2>/dev/null)" || fail 26 "repository path does not resolve"
[[ -d "$REPO" ]] || fail 26 "repository path is not a directory"
BASE="$2" SNAP="$3" SNAP_SHA="$4" COUNT="$5"
[[ "$BASE" =~ ^[0-9a-f]{40}$ ]] || fail 26 "BASE must be a full 40-hex commit"
[[ "$SNAP_SHA" =~ ^[0-9a-f]{64}$ ]] || fail 26 "SNAPSHOT_SHA256 must be 64 lowercase hex"
[[ "$COUNT" =~ ^[1-5]$ ]] || fail 26 "COUNT must be 1-5"
EV="$(realpath -m -- "$6" 2>/dev/null)" || fail 26 "evidence path does not resolve"
inside_repo() { [[ "$1" == "$REPO" || "$1" == "$REPO"/* ]]; }
inside_repo "$EV" && fail 26 "evidence destination resolves inside the repository: $EV"
# Only now may anything be created; the resolved path is re-verified before every write.
mkdir -p -- "$EV"
ev_ok() {
  local now
  now="$(realpath -e -- "$EV" 2>/dev/null)" || fail 26 "evidence directory vanished"
  if [[ "$now" != "$EV" || ! -d "$now" ]] || inside_repo "$now"; then fail 26 "evidence directory changed: $now"; fi
}
ev_ok

PLAN=management/v0/plans/G003-foundation-observability-ci-plan.md
PATH_RE='^(management/v0/(goals/G003-foundation-observability-ci\.md|specs/G003-foundation-observability-ci-spec\.md|PROGRESS\.md|plans/G003-foundation-observability-ci-plan\.md)|docs/ARCHITECTURE\.md)$'
LINE_RE='^([0-9a-f]{64})  ([^\].*)$'
cd -- "$REPO"

# 1. Baseline commit.
[[ "$(g rev-parse HEAD)" == "$BASE" ]] || fail 20 "HEAD is not $BASE"

# 2. Snapshot: read once; verify the digest of exactly those bytes.
bytes="$(cat -- "$SNAP"; printf x)" || fail 21 "cannot read the snapshot"
bytes="${bytes%x}"
[[ "$(printf '%s' "$bytes" | sha256sum | cut -d' ' -f1)" == "$SNAP_SHA" ]] || fail 21 "snapshot digest mismatch"

# 3. Parse only those bytes: format, relative unique approved paths, exact count, Plan present.
[[ -n "$bytes" && "$bytes" == *$'\n' ]] || fail 22 "snapshot is empty or lacks a final newline"
declare -A want=()
n=0
while IFS= read -r line; do
  [[ "$line" =~ $LINE_RE ]] || fail 22 "malformed entry: $line"
  sha="${BASH_REMATCH[1]}" path="${BASH_REMATCH[2]}"
  [[ "$path" != /* && "/$path/" != */./* && "/$path/" != */../* ]] || fail 22 "non-relative path: $path"
  [[ "$path" =~ $PATH_RE ]] || fail 22 "path outside the approved set: $path"
  [[ -z "${want[$path]+x}" ]] || fail 22 "duplicate path: $path"
  want[$path]="$sha"; n=$((n + 1))
done <<<"${bytes%$'\n'}"
(( n == COUNT )) || fail 22 "expected $COUNT entries, found $n"
[[ -n "${want[$PLAN]+x}" ]] || fail 22 "the Plan is not an entry"

# 4. Every input is a regular file matching its approved hash (committed or not).
for path in "${!want[@]}"; do
  [[ -f "$path" && ! -L "$path" ]] || fail 23 "missing or non-regular input: $path"
  [[ "$(sha256sum -- "$path" | cut -d' ' -f1)" == "${want[$path]}" ]] || fail 23 "hash mismatch: $path"
done

# 5. Starting state (NUL-safe, no optional locks): every path that differs from HEAD is a snapshot
#    path and is untracked or unstaged-modified; nothing staged, deleted, renamed or ignored.
#    Snapshot paths that are committed unchanged at BASE are allowed (step 4 bound their bytes).
#    The single exception is the ignored, tool-local Claude Code settings file named below: it is
#    never a repository input, is not read, and is reported in the evidence.
TOOL_LOCAL=.claude/settings.local.json
ev_ok
g status --porcelain=v1 -z --untracked-files=all --ignored >"$EV/s0-status.z" || fail 24 "git status failed"
ndiff=0 ntool=0
while IFS= read -r -d '' rec; do
  xy="${rec:0:2}" path="${rec:3}"
  if [[ "$xy" == "!!" && "$path" == "$TOOL_LOCAL" ]]; then ntool=$((ntool + 1)); continue; fi
  [[ "$xy" == "??" || "$xy" == " M" ]] || fail 24 "unexpected status '$xy': $path"
  [[ -n "${want[$path]+x}" ]] || fail 24 "differing path not in the snapshot: $path"
  ndiff=$((ndiff + 1))
done <"$EV/s0-status.z"

# 6. Local resources the execution must never touch, and the required tools.
project="$(sed -n 's/^COMPOSE_PROJECT_NAME=//p' .env.example)"
[[ -n "$project" ]] || fail 25 "cannot read COMPOSE_PROJECT_NAME from .env.example"
[[ ! -e .env && ! -L .env ]] || fail 25 "a root .env exists"
cv="$(docker compose version --short 2>/dev/null)" || fail 25 "docker compose unavailable"
[[ "$cv" =~ ^v?([0-9]+)\.([0-9]+) ]] || fail 25 "unparsable compose version: $cv"
(( BASH_REMATCH[1] > 2 || (BASH_REMATCH[1] == 2 && BASH_REMATCH[2] >= 24) )) || fail 25 "Compose >= 2.24 required (found $cv)"
for kind in "ps -a -q" "volume ls -q" "network ls -q"; do
  # shellcheck disable=SC2086
  out="$(docker $kind --filter "label=com.docker.compose.project=$project")" || fail 25 "docker query failed"
  [[ -z "$out" ]] || fail 25 "existing $project resources ($kind)"
done
for port in 3100 8180 8100 5532 6479 9100 9101; do
  if (exec 3<>"/dev/tcp/127.0.0.1/$port") 2>/dev/null; then fail 25 "port $port is busy"; fi
done
ev_ok
{ docker ps -a --format 'container {{.Names}}'; docker volume ls --format 'volume {{.Name}}'
  docker network ls --format 'network {{.Name}}'; docker image ls --format 'image {{.Repository}}:{{.Tag}}'; } |
  LC_ALL=C sort >"$EV/s0-docker-inventory.txt"

ev_ok
printf 'S0 PASS: HEAD %s; snapshot %s; %d inputs verified; %d differing path(s), all snapshot paths; %d tool-local ignored file(s); no .env, no %s resources, default ports free; compose %s\n' \
  "$BASE" "$SNAP_SHA" "$n" "$ndiff" "$ntool" "$project" "$cv" | tee "$EV/s0-gate.txt"
```

### 15.2 Planning-time demonstration (read-only for the real repository)

- **Setup.** Run on 2026-10-05 with the exact script above.
  - shellcheck v0.11.0 (`-S warning`) is clean.
  - Every case ran in a **disposable `git clone`** of the baseline, in the job's scratch directory, with an untracked stand-in Plan.
  - Snapshots were built from the clone's files. These snapshots demonstrate the gate only; they are **not approval snapshots**.
- **Measurement.** Around every invocation: a whole-clone inventory including `.git` (path, type, size, nanosecond mtime, inode) and `.git/index` hash plus stat data.
- **Real repository.** Its inventory, including `.git`, and its index were unchanged across the whole demonstration.

| Case | Input | Want | Got | Tree / index |
|---|---|---|---|---|
| P | Plan (untracked) + Goal + Spec + PROGRESS, COUNT 4 | 0 | 0, `4 inputs verified; 1 differing path(s), all snapshot paths; 0 tool-local ignored file(s); no .env, no beginify resources, default ports free; compose 2.32.4-desktop.1` | same / same |
| P5 | As P plus ARCHITECTURE.md, COUNT 5 | 0 | 0 | same / same |
| N1 | Wrong pinned digest | 21 | 21 | same / same |
| N2 | COUNT 3 with 2 entries | 22 | 22 | same / same |
| N3 | Duplicate entry | 22 | 22 | same / same |
| N4 | Absolute path | 22 | 22 | same / same |
| N5 | `README.md` entry (outside the set) | 22 | 22 | same / same |
| N6 | Plan not an entry | 22 | 22 | same / same |
| N7 | Goal modified after snapshot | 23 | 23 | same / same |
| N8 | Extra untracked file | 24 | 24 | same / same |
| N9 | Plan staged (`A `) | 24 | 24 | same / same |
| N10 | Root `.env` (`!!`) | 24 | 24 | same / same |
| N11 | HEAD moved by an empty commit | 20 | 20 | same / same |
| N12 | Unlisted tracked file modified | 24 | 24 | same / same |
| T1 | Ignored `.claude/settings.local.json` present (ignored as in the real checkout) | 0 | 0, `1 tool-local ignored file(s)` | same / same |
| T2 | Ignored `.env.local` present | 24 | 24 (`!!` `.env.local`) | same / same |
| T3 | Untracked `.claude/other.json` | 24 | 24 (not in the snapshot) | same / same |
| A1 | Evidence path inside the repository (nonexistent) | 26 | 26; **not created** | same / same |
| A2 | Evidence through an external symlink into the repository | 26 | 26; **not created** | same / same |
| A3 | Abbreviated BASE | 26 | 26 | same / same |
| C1 | Stale-index fixture (tracked mtimes changed), positive path | 0 | 0 | same / **same** |
| C0 | Ordinary `git status` (optional locks enabled) on that fixture afterwards | index rewritten | rewritten: the detector is non-vacuous | — |

### 15.3 Sequence

| Step | Work | Gate (all component gates use execution-owned images built directly with `docker build`; no root `.env`) |
|---|---|---|
| S0 | §15.1 gate. Then the wrapper `run-approved.sh --check-only` with the approved snapshot, digest, count, `--path-re` and `--plan`, then `run-approved-controls.sh` (W1–W9). | All exit 0, or stop before any write |
| S1 | Contracts corpus, `expectations.json`, the conformance README; `schemas.py` corrections; `test_conformance.py` and `conformance_exchange.py`; the intelligence Dockerfile check-stage `COPY` | The intelligence check image builds; `pytest -q` passes, with conformance included and none skipped |
| S2 | The tools module: `go get …@v6.0.3 && go mod tidy`, in `golang:1.27.1-alpine` with the `g003-gomod-<exec>` volume (the only network step); commands, tests and the Dockerfile | The tools check stage passes; `contractcheck corpus` agrees with the expectations |
| S3 | Core `correlation.go`, `logging.go`, `checks.go` and the two new test files | `core-check` (gofmt, vet, all tests) passes; the existing tests are unchanged and pass |
| S4 | Web `request-context.ts`, `core-client.ts`, both routes, the scripts and the unit test | `web-check` (lint, type check, unit) passes. The production build passes. Next's startup lines are confirmed against §5.4. |
| S5 | Intelligence `http.py`, `export.py`, `main.py` and the two new test files | The intelligence check image passes (ruff, format, all tests) |
| S6 | Harness (§10): the contracts suite, the `obs_*` functions, B-HEALTH-ROOT, `required-checks.tsv` and B-REQUIRED-CHECKS, B-ARTIFACT-SCAN, the register (+10), `APPLICABLE`, and the B-DEPS, B-SHELL, B-HADOLINT, B-DOCKERIGNORE, B-WEB-BOUNDARY and B-MCP-SOURCE additions | Through the wrapper: **static + provenance + contracts all PASS** (the G002 lesson: after any harness change, run the static suite) |
| S7 | CI: `.gitlab-ci.yml` (with the approved tag), `validate.sh`, `validate-controls.sh`, `cleanup-owned.sh`, `unchanged.allow`; B-CI-CONFIG and B-CI-CONTROLS | Static passes, including V1–V12. **H0** if authorized. |
| S8 | Docs: README blocks, OPERATIONS.md, FOLLOW-UP labels (§17); the G003 scope allow-list | — |
| S9 | Integrated local run through the wrapper with `root migrations variants` (and then `all` in S11), and **local CI parity**: `validate.sh` in the scratch clone (Q6) | All PASS; validate exits 0 |
| S10 | Two read-only reviewers (§20); triage; fixes | Findings resolved or recorded; the suites covering each changed file are rerun |
| S11 | **Final local run with no edits during the run:** the wrapper with `all` (§15.4); cleanup inventory; Result draft | `overall_rc=0` |
| S12 | H1 and H2 (authorized), then H3; final Result; stop for Codex | Hosted pass on the exact commit; hosted fail for the intended reason |

The **cycle-4 checkpoint** is recorded *after* cycle 4 closes and before cycle 5 begins, wherever that falls (§19).

### 15.4 Final local invocation (S11)

```bash
bash infra/validation/run-approved.sh \
  --snapshot <approved snapshot> --snapshot-sha256 <its SHA-256> --count <COUNT> \
  --path-re '^(management/v0/(goals/G003-foundation-observability-ci\.md|specs/G003-foundation-observability-ci-spec\.md|PROGRESS\.md|plans/G003-foundation-observability-ci-plan\.md)|docs/ARCHITECTURE\.md)$' \
  --plan management/v0/plans/G003-foundation-observability-ci-plan.md \
  -- --base <BASE> --allow-file management/v0/results/G003-scope-allowlist.txt --report-dir <ext>/g003-exec/s11/report all
echo "overall_rc=$?"
```

- **Arguments.** Every non-Plan snapshot entry becomes `--expect-hash` and an exact `--exception`, as the wrapper builds them today.
- **Outcome.** Exit 0 is the only passing result. Any other status is a failed run: it is recorded, and the handoff stops.
- **Scope.** B-SCOPE checks the exact §14 allow-list against BASE. The harness enumerates with `--exclude-standard`, so the globally ignored tool-local file does not enter B-SCOPE, copies or scans.

## 16. Test strategy and acceptance-criteria mapping

| AC | Implementation | Commands and assertions (local unless marked hosted) |
|---|---|---|
| **AC1** Health semantics and bounds | §6, §7 | **Unit:** Go `health_test.go` (unchanged) and `propagation_test.go` (5 s double → about 2 s); Python `test_health_context.py` (no-store, bodies, unconfigured models ready).<br>**Root:** B-HEALTH-ROOT (six routes on the actual root, with exact bodies, no-store and `X-Request-Id`); B-WEB-HTTP.<br>**Variants:** B-OBS-TIMEOUTS, O1–O10 on the enabled copy (the §6.3 matrix; each single request ≤ 5000 ms by `httpprobe`'s monotonic clock; recovery ≤ 90 s each), plus B-OUTAGE (unchanged); B-OBS-OFF (O1–O6, O9, O10, no `optional` key); B-TOGGLE-OFF.<br>**Evidence:** `obs-timings.tsv`. |
| **AC2** Context validation, generation and isolation | §4 | **Table tests in all three runtimes:**<br>• Go `context_test.go`;<br>• Node `request-context.test.mjs`;<br>• Python `test_health_context.py`.<br>**Table rows:** valid; absent; malformed; oversized (65, 513 and 8 KiB); duplicated field and comma-joined; all-zero trace and parent; uppercase; versions `01` and `ff`; extra fields; secret-valued ID.<br>**Isolation:** concurrency at 64, 200 and 100; cancellation and failure leave no state.<br>**Allow-list:** baggage, tracestate, cookie, auth and arbitrary headers are never forwarded — Go `propagation_test.go` (received set) and Node `outboundHeaders` (exactly 2 keys), plus the B-WEB-BOUNDARY static rule.<br>**Integration (B-OBS-CONTEXT, enabled copy):** every invalid row through web `/api/ready`. The response ID is generated (32 hex); the records are found by that ID; the raw invalid literal is absent from every service's raw log; there is no 500. |
| **AC3** web → core → intelligence continuity | §4.3, §7 | **B-OBS-CHAIN (enabled copy; `logcheck chain` over `docker logs` of web, core and intelligence):**<br>• (i) supplied valid context, flags `01`;<br>• (ii) no context;<br>• (iii) invalid replaced;<br>• (iv) 32 concurrent distinct contexts;<br>• (v) O1–O6 (web and core; plus intelligence while it is up);<br>• (vi) O7 (no intelligence record) and recovery;<br>• (vii) O9/O10 (web only; core ≤ 1 after unpause).<br>**Each case asserts:** exactly one record per receiving service; equal `request_id`/`trace_id`; `web.parent` equal to the supplied parent or absent; `core.parent = web.span`; `intelligence.parent = core.span`; distinct nonzero spans; propagated flags.<br>**Disabled variant (B-OBS-OFF):** web → core only.<br>**Constraints:** no new route, demo endpoint or provider call is used; `httpprobe` reaches only existing routes. |
| **AC4** Structured logs and privacy | §5 | **`logcheck validate` per service on the full raw logs of both copies:** every line is classified; completion records validate against the schema; exactly one per request; an implicit 200 (core and intelligence liveness, web health) and a handled 503 are counted correctly (B-OBS-LOGS).<br>**B-OBS-PRIVACY:** per-run random canaries.<br>• **Headers:** `Authorization: Bearer`, `Cookie`, `X-Api-Key`, `baggage`, `tracestate`, `X-Canary`, `User-Agent`, `Referer`.<br>• **Inputs:** query `token=` and `q=`; a GET body; invalid IDs that embed a canary.<br>• **Through `/internal/v1/execute` (from core's namespace, token via a file):** prompt, context-ref and authorization-ref canaries.<br>• **Secret-valued IDs sent directly** to core and to intelligence.<br>• **Scan:** raw `docker logs` of all services and every retained artifact; none may appear.<br>**Positive controls:** a planted canary in a scratch log copy is detected; the canary requests' completion records exist with route and status, so diagnostics are not suppressed; `logcheck` unit negatives (§9); the Python leak stub (§7.3).<br>**Also:** B-SECRET-CANARY (extended list) and B-ARTIFACT-SCAN. Intelligence content capture stays off (unit test plus the export record in the raw log). |
| **AC5** Cross-language contracts | §8, §9 | **Contracts suite:**<br>• B-CONTRACT-TOOLS;<br>• B-CONTRACT-GO and B-CONTRACT-PY (120 cases and 7 schema-negative cases each, 0 skipped, digests equal);<br>• B-CONTRACT-AGREE (identical verdict sets equal to the expectations; all 8 event types; categories);<br>• B-CONTRACT-PEER (Go envelopes accepted and executed by Python; real Python NDJSON accepted by Go; four corruptions of each rejected);<br>• B-CONTRACT-NEGATIVE (SN01–SN07; flipped EI39).<br>**Also:** `pytest` in B-CHECKS-ROOT runs `test_conformance.py`. No production client: B-MCP-SOURCE (`go list -deps`) and B-WEB-BOUNDARY. The schemas are byte-identical. |
| **AC6** Runnable containerized commands | §10–§13 | README blocks are executed verbatim: the existing ones by the harness; `platform:contracts` and `platform:ci` at S9/S11 (recorded) and by the CI job itself.<br>**Unchanged checks:** B-DOCKER-ONLY (shadowed runtimes; zero invocations); B-CHECKS-ROOT; B-BROWSER-ROOT and B-BROWSER-OFF (95 + 95; 0 skipped, flaky or unexpected).<br>**Images and network:** images are built from locked inputs; check containers run with `--network none`.<br>**Recorded:** tool and image versions (IMAGES, `summary.json`), and cache use (Docker `system df` before and after, plus whether each build was cold or warm). Cleanup is shown by `inventory-leftover.txt`. |
| **AC7** GitLab gating with hosted evidence | §11, §12 | **Static:** B-CI-CONFIG (rules, single gating job, no bypass, artifacts, access, retention, timeout, tag). Runner facts (§12.2) are recorded.<br>**Hosted (H1):** a pass on the exact submitted commit and tree, with pipeline, job, runner, `tested_content`, 67 PASS and duration.<br>**Hosted (H2):** a failure through the reviewed gate (B-CONTRACT-*), with sanitized artifacts. There is no deploy or publish job, no CI/CD secret and no `allow_failure`. |
| **AC8** Negative controls and failure propagation | §10, §11.3 | **B-CI-CONTROLS V1–V12, covering (a)–(f) and their combinations:**<br>• (a) V2, V3;<br>• (b) V4, plus B-CONTRACT-NEGATIVE;<br>• (c) V5, V6, plus B-ARTIFACT-SCAN;<br>• (d) V7, V12;<br>• (e) V8, V9;<br>• (f) V10, V11, plus B-REQUIRED-CHECKS.<br>**Retained:** W1–W9 (B-WRAPPER-CONTROLS) and C1–C12/C8b/C8c (B-DEPS-CONTROLS). The wrapper postcheck applies to the final run.<br>**Recorded:** each control's status and its safe cleanup summary. |
| **AC9** Scope, architecture and provenance | §14, §17 | **Scope and protection:** B-SCOPE against the exact allow-list; B-PROTECTED (skills, manifest, governance and closed artifacts); B-SKILL-UNCHANGED.<br>**Provenance:** B-RENDER-ID; B-RENDER-PLAN and B-RENDER-REPEAT (exactly 18); B-TOGGLE-OFF-ADDITIONS (34 applicable).<br>**Full regressions:** migrations (B-MIGRATE-*), roles and isolation (B-BOUNDARIES-*, B-BOUNDARY-CONTROLS), identity (B-RENAME, B-DBRENAME), ports (B-PORTS-*), outages (B-OUTAGE), persistence (B-PERSIST), MCP (B-MCP-SOURCE, B-MCP-HTTP), browsers (root, off, rename).<br>**Diff review:** no schema, route, UI or vendor change. |
| **AC10** Documentation, Result and governance | §14, §19 | **Docs:** OPERATIONS.md, README and FOLLOW-UP (metrics and exporters left open).<br>**Result:**<br>• maps AC1–AC10 to commands and artifacts;<br>• separates inherited, new and deferred-open items;<br>• separates local evidence (S9/S11) from hosted evidence (H1/H2);<br>• lists every run, including failures;<br>• records the B7 cycles and the checkpoint after cycle 4.<br>**Required checks:** B-REQUIRED-CHECKS (67) with no skips.<br>**Closure:** only Codex ACCEPT completes G003. |

## 17. Regression reconciliation and documentation labels

- **G001 guarantees, preserved and re-run:**
  - migration owner/runtime separation, owner-only metadata, the unchanged migration set;
  - independent datastore port and IP isolation with its controls;
  - liveness and readiness outage semantics;
  - persistence and Redis ephemerality;
  - naming and default ports;
  - the dormant MCP boundary;
  - build-context canaries.

  `main.go` is unchanged, so B-MCP-SOURCE's byte check stays valid without reconciliation.
- **G002 guarantees, preserved and re-run:**
  - appearance and first paint, scoped tokens, the full root and disabled browser matrices;
  - the dependency policy, with justified `devOptional`, the hash-pinned registry CSS block and the literal-scan exceptions;
  - wrapper controls and B-WEB-* rules.

  No UI file changes.
- **Accepted deviations and residuals, unchanged:**
  - G002's fixed angle, destructive tint and helper modules;
  - G001's lost logs, SIGKILL residue, owner-role preflight, WSL probe limitation, data-network egress and tag-only pins.
- **Byte-equality guards** are reconciled only through the 10 exact register rows and the 6 `APPLICABLE` additions. Their semantic assertions and reason codes are unchanged.
- **FOLLOW-UP STAGE-OBSERVABILITY:**
  - "Metrics export and trace propagation…" stays **unchecked**, annotated *"Partly done (G003): diagnostic context and per-hop spans on the existing health chain with correlated JSON logs; metrics export, collectors and external trace export remain open for a separately scoped decision."*
  - "Contract validation… in CI" → `[x]` *(G003)*.
  - "Containerized … CI … change evidence" → `[x]` *(G003: GitLab pipeline; commit, pipeline, job and actor recorded; review attribution from merge requests)*.
- **Out of scope here (owned elsewhere):**
  - durable audit (G011);
  - broader security and vulnerability controls (G012);
  - recovery and incident monitoring (G013);
  - Organization data and branding (G016/G017);
  - product AI observability.
- **Architecture guidance.** §60's hostname guidance and §61's design-library evolution are referenced as guidance for their owning Goals. G003 adds no hostname resolver, routing grammar, UI redesign or component change.

## 18. Security and privacy negative controls (summary)

- **No authority.** No diagnostic ID, flag, fixture or CI actor confers authority or tenant scope:
  - IDs are validated, bounded and never used for decisions;
  - corpus UUIDs are synthetic;
  - CI attribution is metadata only.
- **Network boundaries unchanged.**
  - Web never reaches intelligence or the datastores (B-BOUNDARIES).
  - Test tools join service namespaces only for probes, publish nothing and receive no secret.
  - Canary and secret inputs travel by file.
- **No new secret paths.** No CI/CD secrets; MR code gets no privileged variables; artifacts follow §11.2 (allow-list, gitleaks and canary scans, withholding, `access: developer`).
- **Controls proving the checks are not vacuous:** a planted canary; leak stubs; `logcheck` negatives; W1–W9, C1–C12 and V1–V12; flipped corpus expectations; SN01–SN07; four NDJSON corruptions.

## 19. Budget, feasibility and checkpoint (B7)

**Uncertainty sources:**
- three-runtime context and logging semantics;
- framework logger takeover (uvicorn) and Next's startup output;
- cross-validator agreement on edge cases (format, integer-valued numbers, patterns);
- harness integration of pause and stop timing with log-chain parsing;
- the new CI entry point and controls;
- first hosted execution on an as-yet-unverified runner.

| Likely meaningful cycle | Estimate |
|---|---|
| S1–S5 first unit and component failures (context tables, formatter or uvicorn takeover, Node type stripping) | 1 |
| Cross-validator disagreements found by the corpus (S1/S2) | 0–1 |
| Integration observability cases (timing, `logcheck` chains, Next startup lines) | 1–2 |
| Harness and CI reconciliation (static, provenance, V-controls) | 1 |
| First hosted run (runner environment, timeout, artifact limits) | 1 (0 if H0 surfaces nothing) |
| Read-only review findings | 1 |

- **Forecast: 5–7 of 7, most likely 6.**
  - This is **credible only if Q1 is resolved to Profile A (or A2) before S0, and H0 is authorized.** Without H0, a runner incompatibility found at H1 would arrive with almost no budget left.
  - The margin is thin. G002 forecast 5–6 and used 7.
  - Planning revisions consume no cycles.
- **Checkpoint after cycle 4, before cycle 5** (BUDGET_POLICY §7). It records:
  - per-criterion status;
  - suite status;
  - failure convergence and repeated patterns;
  - remaining work against the cycles left;
  - the forecast.

  Stop and return to Codex if the forecast exceeds 7, on thrashing (§8), or on any escalation condition (§21).
- **Counting** follows BUDGET_POLICY §3–§5. Doubtful cases count. The step order creates every imported module before the gate that builds it.
- **Split option (for Codex; not taken unilaterally).** If Q1 cannot be resolved to an acceptable profile before approval, G003 is blocked as written, because AC7 requires hosted evidence and the Spec forbids deferring it. Codex may then split it:
  - **G003a:** R1–R5 and R8, plus local R6/R7 including the CI entry point and controls;
  - **G003b:** hosted GitLab execution (AC7) once a runner exists.

  **I recommend the single Goal** if Q1 resolves to Profile A or A2. The split adds review overhead, but no risk reduction once a runner is available.

## 20. Delegation

- **One implementation owner:** me. There are no implementation subagents: the work shares the harness, the log contract and the three runtimes.
- **Two read-only reviewers at S10**, run in parallel. Neither edits files. Each reports findings with severity and `file:line`, and the cycles they used (expected 0).
  - **SA-P (correlation and privacy):**
    - the §4 profile against W3C and the Spec;
    - isolation, cancellation and panic paths;
    - completion "exactly one" semantics;
    - redaction parity;
    - framework output;
    - canary coverage and non-vacuity.
  - **SA-C (CI, contracts and provenance):**
    - validator independence and the corpus/expectation design;
    - the Python corrections;
    - `validate.sh` exit semantics and V1–V12;
    - artifact allow-list and scan;
    - B-CI-CONFIG;
    - the register, `APPLICABLE` and the allow-list;
    - the G001/G002 regression coverage.
- **Integration and budget.** I triage and integrate every finding. Corrective work counts against the shared B7. Reviewer opinions do not replace Codex review.

## 21. Risks and escalation

| # | Risk | Handling |
|---|---|---|
| R1 | Runner profile incompatible, unavailable or privileged | **Blocker (Q1).** Stop; never use dind or privileged mode silently; never substitute another CI vendor |
| R2 | uvicorn emits a line before the takeover, or a non-JSON warning | `logcheck` fails visibly. Fix through the logging configuration inside the app, not by widening the allow-list. If only `--log-config` would work, that changes the Dockerfile `CMD` and is returned to Codex. |
| R3 | Node 24 type stripping cannot load `request-context.ts` from `.mjs` | Fallback: a Playwright `node`-only project in the existing browser image, with no new dependency. Recorded as a deviation. |
| R4 | Go 1.27 `ServeMux` does not expose `r.Pattern` on the wrapped request | Fallback: derive `route` from the fixed set (`/healthz`, `/readyz`, else `unmatched`) in the middleware. The test asserts the result. |
| R5 | santhosh and Python disagree on a corpus case | Diagnose against JSON Schema and RFC semantics. **Never relax schemas or drop cases.** A library defect with no conforming configuration is escalated (Spec risk 3). |
| R6 | The paused-target record count (O8, O10) is nondeterministic | Assert 0 or 1 with matching IDs, never more. A duplicate or mismatch fails. |
| R7 | The cold hosted run exceeds the timeout | Measure at H0; report against the 3 h limit. Any timeout change is a Q1/Q7 decision; checks are never dropped. |
| R8 | Artifact size: the browser JSON | Attachments are stripped (§11.2). The remaining size is measured at H0. |
| R9 | Gitleaks flags a synthetic fixture | Change the fixture value. Never add an allow-list entry. |
| R10 | Harness edits weaken G001/G002 checks | §10 lists every change; SA-C reviews; the static suite runs after each harness change |
| R11 | Budget exhaustion or thrashing | Checkpoint after cycle 4; stop rules |
| R12 | The client tool rewrites `.claude/settings.local.json` mid-run | It is globally ignored and excluded from harness enumeration, and the S0 exception covers only that exact ignored path. It is never touched. Any other new path still fails S0, B-SCOPE or the wrapper postcheck. |

**Stop and escalate on:**
- an architecture or ADR conflict;
- scope expansion;
- a security-model change, including runner privileges;
- a new vendor or infrastructure;
- any schema, migration, route, API or MCP change;
- any need to change `main.go`, `local.yml`, the deadlines or the schema bytes;
- weakening a check or assertion;
- touching developer resources;
- missing hosted authorization;
- B7 exhaustion.

## 22. Questions for Codex and unmet prerequisites

| # | Question or prerequisite | Recommendation |
|---|---|---|
| **Q1 (BLOCKER)** | Supply the existing runner's facts (§12.2) and confirm it is Profile A or A2. If it is GitLab-hosted or dind-based, an explicit security decision is needed. | Owner reads them in GitLab Settings → CI/CD → Runners; Codex confirms the profile and the tag |
| **Q2** | Authorize the hosted actions H0–H3 (§12.3): branch pushes, draft MRs to `staging`, and deleting `g003-gate-proof`. Choose the evidence retrieval mechanism. | Authorize H0–H3. The owner downloads job logs and artifacts to a path outside the repository, or provides a short-lived `read_api` token out of band. |
| **Q3** | Confirm that the §8.4 Python validator corrections are conformance fixes, not a prohibited "restrictive wire change". Schema bytes are unchanged; runtime 422 now covers the two gaps. | Approve; otherwise R5 agreement cannot be met |
| **Q4** | F11: equal nested 2 s deadlines make web report core unavailable while an optional intelligence *hangs*. | Keep the deadlines, as the Spec requires; assert core-level non-gating and record the web-level outcome (§6.3 O8); track a deadline adjustment as follow-up |
| **Q5** | §4.4: web, which holds no credentials, cannot classify a credential supplied as an ID; core and intelligence refuse to adopt or forward one | Accept as the boundary definition |
| **Q6** | Local CI-parity run of `validate.sh` in a disposable clone with a scratch commit, never pushed and deleted afterwards | Approve |
| **Q7** | Artifact retention of `90 days` with `access: developer`. The Result keeps downloaded copies' SHA-256 values outside the repository. | Approve, or set another period |
| **Q8** | The "Pipelines must succeed" and protected-branch project settings are owner-managed and are not changed by G003. AC7 "gating" means the required job fails the pipeline. | Confirm; the owner may enable merge blocking separately |
| **Q9** | The fresh binding: BASE (commit any approval edits first, or keep them uncommitted), the snapshot entries (Goal, Spec, PROGRESS, Plan, and optionally ARCHITECTURE) with their digest and COUNT | Codex supplies them in the approval |
| **Q10** | Split versus single Goal (§19) | Single Goal if Q1 resolves to A or A2 |
| **Q11** | Corpus location `contracts/conformance/intelligence/`: fixtures only, outside the runtime image and the disabled variant | Approve |
| **Q12** | The santhosh v6.0.3 test-tool module and its image: no ADR, analogous to the Playwright image | Approve |
| **Q13** | F12: accept `.claude/settings.local.json` (the client's ignored local settings) as S0's single exact ignored-path exception, or have the owner remove it before S0 | Accept the exact exception; do not remove the user's client settings |

## 23. Readiness

**r1: ready for Codex plan review, with a blocking prerequisite.**

- **Prerequisite.** Execution cannot start until Q1 is resolved, Q2 is authorized and Q9 is supplied.
- **Verified at writing:**
  - HEAD `9461c4b` and its input hashes;
  - the S0 script hash `0409148a…6723` and its 22 demonstration rows, as expected, including the tool-local exception cases;
  - no repository write other than this Plan.
- **Not done:** nothing has been implemented, installed, initialized, built, committed, pushed or triggered.
- Implementation waits for an explicit Codex approval that names BASE, the approved snapshot, its SHA-256 and COUNT.
~~~~~~~~
