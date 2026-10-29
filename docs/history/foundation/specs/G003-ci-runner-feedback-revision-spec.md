> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G003 — CI runner and feedback revision work specification

Date: 2026-10-06
Status: Approved revision scope, 2026-10-06; accepted ADR-002 and Plan r2; fresh binding and named external-action gates remain
Supervisor: Codex
Parent: [G003 Goal](../goals/G003-foundation-observability-ci.md) and [original Work Spec](G003-foundation-observability-ci-spec.md)
Decision prerequisite: [ADR-002 draft](../../../adr/ADR-002-gitlab-hosted-ci-runner.md)

## 1. Objective and rationale

Complete G003 with a measured GitLab.com hosted execution profile and proportionate feedback gates. Preserve every existing runtime, contract, security and regression assertion while avoiding the entire integration/control matrix on every ordinary change. No deployment is introduced.

The original single full job and unresolved existing-runner premise are inadequate for the owner's development cost/feedback requirements. API inspection now succeeds as <owner-account> for project <gitlab-project-id> (`<namespace>/<project>`): hosted runners are enabled, and hosted small and medium AMD64 entries are online. Listing availability is not qualification or a guarantee of usable quota. The final correction wrapper run locally took about 48 minutes; the earlier two-hour-per-job/four-hour estimate was a forecast, not measured hosted performance.

This linked revision preserves original Goal/Spec/Plan and Result history. Once approved, it supersedes only R6's universal full-job trigger interpretation, AC7's corresponding pipeline coverage interpretation, the existing-runner prerequisite and the privilege exclusion for the narrowly accepted ADR-002 profile. It adds the AC7a–AC7g checks below. R1–R5, R7–R8, AC1–AC6 and AC8–AC10 remain binding. No existing full-suite assertion is removed or relaxed.

## 2. Boundaries

In scope: the GitLab job profile; fast/affected/full test selection; minimum suite/report handling changes required by that selection; exact trigger/coverage evidence; containerized job setup; controlled build caching where necessary; runner qualification; attributable hosted success/failure evidence; operating documentation.

Out of scope: application behavior, UI/G048, business/auth/tenant schema, production deployment, registry publishing, cloud provisioning, paid dedicated runner, new CI vendor, runner fleet/autoscaler, generalized scheduling/classification framework, replacement harness, unrelated dependency upgrades and new product/domain contracts. Canonical UUID identity and the G021 barrier remain unchanged.

Reuse `infra/validation/run.sh`, `run-approved.sh`, existing check images and `infra/ci` failure/sanitization/cleanup controls. A profile is a bounded selection of existing checks, not a new validation engine. Do not claim that the existing `static` suite is fast without measuring it: it currently invokes expensive controls.

## 3. Prerequisites and change boundary

1. Finish the in-flight G003 correction validation; retain its local evidence without invalidating or rewriting its five-input snapshot.
2. Codex reviews the returned correction diff/evidence before releasing revised implementation. G003 remains unaccepted pending this revision and hosted evidence.
3. Obtain owner acceptance of ADR-002's privileged, job-isolated DinD profile. Current `glab` authentication authorizes read-only inspection; this specification alone grants no pushes, pipelines, schedule creation, project-setting changes or spend.
4. Claude writes a short linked G003 Plan revision with exact files, reviewed dependency/image pins, suite-to-check inventory, trigger matrix, network/path integration, minimum negative controls, capacity forecast and hosted action boundaries. Preserve Plan r4 and correction evidence; do not reproduce accepted control design.
5. Codex reviews that revision and supplies a fresh exact baseline/input binding. No main-checkout draft is inserted into the currently bound execution worktree during its run.
6. Verify namespace allowance, project permissions, actual runner/job profile, artifact limits and access. Tokens remain local and never enter repository files, logs or evidence.

## 3.1. Planning and execution validation boundary

The next Claude handoff is **planning only**: write a short linked G003 Plan revision against this Spec. Do not implement, install dependencies, build application images or run the full suite during planning. Reuse completed correction-pass evidence and accepted execution controls by exact reference; do not recreate their proof. These instructions do not interrupt the already authorized correction run.

Resolve the medium-runner profile, fast/affected/full check mapping, trigger rules and Docker-in-Docker compatibility using read-only inspection and lightweight syntax/rule checks. If a concrete uncertainty needs a disposable hosted probe, propose its bounded purpose, command, expected evidence and external actions before running it; no pipeline is authorized merely by this planning request.

For the eventual approved revision implementation, specify focused checks for changed CI/profile behavior and relevant regressions. Reserve one complete hosted run on the final revised tree for G003 acceptance, plus the required hosted failure proof and profile/qualification evidence. Do not automatically run full local validation followed by an identical full CI-parity run followed by full hosted validation. A further full run requires a concrete failure, subsequent change affecting coverage or unresolved compatibility/integrity concern; record that reason. Reuse an existing complete pass only for the exact source/validation inputs it actually tested. Earlier local correction evidence supports unchanged behavior but cannot replace final revised-tree hosted acceptance.

This boundary changes the timing and duplication of validation, not its assertions, fail-closed guarantees, baseline binding or required final evidence. A Plan that cannot retain those guarantees with focused checks must identify the specific gap for Codex review instead of expanding process work autonomously.

## 4. Required profiles and triggering

Every runnable pipeline identifies its selected profile, selection reason, tested commit/tree, comparison base and required check set. Profile success means every check required for that profile passed; it does not mean unselected full checks passed.

| Profile / event | Minimum checks and gate |
|---|---|
| Fast: ordinary branch pushes and merge requests | Existing unit tests, lint/type/static boundary checks, deterministic independent contract checks and source-secret checks; no application stack/browser/outage or validation-framework fault matrix by default. All selected checks gate the pipeline. Suppress duplicate branch/MR pipelines when the branch already has an MR, preserving the MR gate. |
| Affected integration: runtime or UI changes | Fast plus service builds and real-stack health/correlation/privacy checks where affected; core/intelligence interfaces test both sides; web changes run the relevant browser/appearance/scoped-token scenarios. Shared boundaries run all affected sides. The Plan enumerates the existing checks; do not substitute mocks for required real-runtime assertions. |
| Full: foundation changes before merge | Complete existing harness, all declared operational/contracts/provenance/migration/isolation/rename/persistence/browser checks, wrapper/dependency/CI negative controls and enabled/disabled variants. Applies to Compose/Dockerfiles/build configuration, dependency/lock files, migrations, CI/harness/control/profile rules and shared execution/security configuration. Tests/fixtures/oracles affecting these guarantees also select full. |
| Full: release candidate | Required passing complete suite on the exact release commit/tree before a future release/deployment. A recorded full pass for that exact content and unchanged validation inputs may satisfy the gate; no automatic second identical run. This adds no deployment job. |
| Full: weekly diagnostics | Default-branch full run once per week; failure stays failed and artifacts are retained safely. Weekly success is not a substitute for a changed release or foundation merge. Creating an external GitLab schedule needs separate authorization. |
| Explicit full invocation | Documented local command and reviewed manual GitLab invocation; always selects all required checks, regardless of changed paths. Required manual gates block until they pass; no optional/allow-failure manual substitute. |

Foundation changes must pass full before merge even when the MR target is staging. Pushes to staging/default still run fast and whatever affected/full checks the actual changed content requires; not every push inherently requires full. For ordinary application changes, fast plus affected integration is the merge gate; weekly and exact-release full gates preserve complete regression coverage.

Path classification must be deterministic and conservative, with a reviewed exact mapping including tests, fixture inputs, build configs and policy files. Use GitLab rules or a small extension to existing tooling. Missing comparison metadata, unknown paths, deletions/renames or unsupported comparison contexts must select full or fail closed, never silently select fast. Compare rename source and destination; treat a newly added/unclassified path conservatively. Reviewed docs-only changes may select fast; executable documentation/commands, security policy and governing test contracts are not automatically docs-only.

GitLab detached MR, merged-result and push attribution remain distinct. Any newly supported schedule/manual/tag context requires explicit minimal attribution rules and controls; unsupported contexts fail. No release tags are created by this revision.

## 5. Runner profile and security

Propose `saas-linux-medium-amd64` for qualification and all profiles, including fast jobs: publicly documented 4 vCPUs, 16 GB RAM, 50 GB storage and compute-minute factor 2. Small (2 vCPUs, 8 GB RAM, 30 GB, factor 1) is an optional later cost optimization, not a G003 prerequisite. No downgrade/upgrade is inferred from timing alone; report measured resource and cost implications. Do not run a second full suite solely to benchmark small.

Use only a fresh job-isolated Docker-in-Docker daemon on the existing GitLab.com hosted offering. Pin client/daemon/setup images and versions in the Plan; verify actual executor and isolation rather than treating runner description or public documentation as proof. All application/test containers remain unprivileged under the existing application model. DinD privilege is confined to the accepted test profile; no production/developer daemon or host Docker socket is mounted. The daemon API is job-private and its authenticated transport or access isolation is explicitly reviewed; no publicly reachable unauthenticated daemon.

Qualification must resolve two concrete compatibility risks: Compose ports bound to 127.0.0.1 live in the daemon namespace rather than automatically on the job shell's localhost; bind-mount paths are interpreted by the daemon and must expose exactly the intended disposable checkout. Keep loopback-only application bindings and the real-root command guarantees. A bounded CI-only namespace/path adapter may be proposed; widening bind addresses, shared sockets or weakening port-isolation tests is not authorized. If compatibility requires substantial execution redesign, return to Codex before implementation.

Measure cold setup/build and warm execution separately, job timeout and cancellation behavior, memory/disk headroom, registry access, selected browser image support, Compose >=2.24, resources/ports, cleanup, artifacts and quota. Controlled setup may download pinned dependencies; test execution retains offline/fake-provider behavior. Cache absence must still work; keys identify relevant locked inputs and no cached report replaces test execution. Cross-trust cache reuse cannot alter dependencies or mask lockfile drift.

Application integration uses disposable PostgreSQL, Redis and the existing MinIO S3-compatible service. No Cloudflare R2 account, bucket, token or external data-plane call is required; provider-specific R2 adoption is outside G003. Unit/static jobs do not start the datastore/application stack unless an explicitly declared test actually requires it.

Keep finite job/process timeouts and the existing failure, drift, secret scan, complete-report and owned-cleanup guarantees. A cancelled/timed-out job must not be reported as passing. Do not expose deploy/provider/production secrets to test jobs. Artifact access remains Developer-or-higher with 90-day retention unless a narrower reviewed limit is necessary.

## 6. Feedback and compute evidence

Planning targets: fast checks <=5 minutes warm (<=10 minutes cold); routine affected integration <=20 minutes warm (<=30 minutes cold). These are proposed feedback objectives, not permission to skip coverage and not timing assertions tied to shared-runner noise. Record job execution separately from queue time, image/dependency setup, builds and suite execution. Full duration is measured and bounded by reviewed timeouts, not assumed to take two hours.

For G003 acceptance, provide at least one measured hosted run for each implemented profile. If a target is exceeded, disclose the slow checks and seek Codex disposition of a measured target or bounded optimization before acceptance; do not keep retrying to manufacture a fast sample. One representative cold qualification plus final hosted profile evidence is enough unless a failure/change warrants more. No benchmark campaign or repeated complete local parity run is required on unchanged source.

Report compute minutes using job duration times the applicable runner factor, with paid-minute estimates separated from remaining included allowance. Medium's factor 2 can cost the same as small if it finishes in half the time; fixed waits may prevent that speedup. Inspect actual namespace quota; no minute purchase, quota increase or dedicated server is authorized here.

## 7. Acceptance checks and minimum verification

These seven subcriteria extend original AC7; original AC1–AC10 still apply with the limited trigger supersession in section 1.

| ID | Required outcome | Verification / evidence |
|---|---|---|
| AC7a | Every current required check belongs to full; fast/affected check selections have explicit coverage and semantics. | Inventory actual declared checks/scenarios and profile membership; compare full to the inherited required list including aggregate checks; no missing/duplicate-disguised check. Run selected profiles; failure and missing report in each gates it. A fast report cannot satisfy full completeness. |
| AC7b | Trigger selection cannot omit a required gate. | Disposable rule/selection controls cover docs-only, web-only, core-only, intelligence/contracts, Compose, each lock/build/CI/profile/control change, fixture changes, unknown/new path, deletion/rename, missing comparison base, new branch, MR source/merged result, weekly/manual full and release intent. Assert exact required profiles/reasons and duplicate suppression. Preserve existing attribution controls, adding only changed contexts. Foundation changes require full before merge. |
| AC7c | Hosted runner passes the reviewed isolated medium profile with working Compose/network/path/browser access. | Real hosted qualification records runner/job IDs, tag/executor, tool/image pins, cold build, filesystem/namespace behavior, loopback/isolation checks, resource peaks and timeout/cancellation/cleanup. Fast jobs use the same qualified medium profile; a later change to small requires capacity/compatibility evidence. Any architectural exception needs accepted ADR/Plan. |
| AC7d | Final hosted passing evidence binds complete and selected checks to the submitted content. | H1 full on exact final code/validation commit/tree; hosted fast and affected-profile evidence for that content. Retain schema/corpus/check-set counts, root/off/browser results, all required controls, authoritative job/pipeline status, sanitized artifact hashes and cleanup. Required reports absent/skipped/malformed fail. Current local evidence is retained but does not replace hosted proof. |
| AC7e | Hosted gate failure and finalization remain fail-closed. | Reuse H2's disposable invalid-contract fault through the real selected gate, with intended check failure/nonzero job and safe artifacts; inspect wiring for every profile. Reuse correction-pass controls for failure, drift, missing evidence, secrets and nested cancellation rather than inventing another suite. Clean fault branch/MR after evidence as authorized; no fault setting remains enabled. |
| AC7f | Feedback/cost claims reflect measured hosted execution and documented gate reuse. | Cold/warm phase durations, runner factors, resource headroom and quota report; document exact profile commands and when full is mandatory. Target overruns require explicit recorded Codex disposition. Test/inspect exact-content full-evidence reuse conditions; a different source tree/validation input cannot inherit a release pass. |
| AC7g | Scope, provenance, governance and historical controls remain intact. | Exact allowed diff/baseline/new snapshot checks; original protected inputs and historical run evidence stay unchanged until revised binding is approved. Preserve application/UI/schema/framework pins and original AC1–AC10 evidence. Result appends supersession, revised hosted actions, budget and unresolved issues; no false acceptance, deploy, spend or early G048/product work. |

## 8. Artifacts, provenance and completion

Claude's deliverables: linked short Plan revision, exact changed-path/check/trigger mapping, minimal updates to CI/existing validation/docs, qualification and profile-control evidence, final hosted H1/H2 evidence and append-only Result revision. No application changes unless a separately reviewed concrete runner-compatibility requirement establishes necessity; this scope does not grant them in advance.

Record base/commit/tree, approved input hashes, Plan revision, runner profile and versions, commands, UTC start/end, pipeline/jobs/status, counts/skips/selection reasons, artifact digests and owned cleanup. Keep credentials, raw unsafe captures and Compose-resolved secrets out of artifacts. Evidence-only fault commits and schedules are separate external actions; prior hosted authorization must be reconciled explicitly to the new profile and new tested ancestors rather than treated as blanket permission.

Completion requires original AC1–AC10 and AC7a–AC7g, accepted ADR-002, approved Plan/binding, measured runner capacity and hosted passing/failing evidence, resolved feedback-target dispositions and Codex implementation ACCEPT. Drafting this revision does not approve current G003 implementation or consume an execution cycle.

## 9. Budget, deviations and stop conditions

Proposed separate revision budget: B3, maximum three shared meaningful implementation/debug cycles, checkpoint after cycle 2. Requires Codex Plan acceptance before activation; original B7 (6/7) and correction pass 1 (reported 0/3) remain separately recorded and are not reset. One primary implementer plus a bounded read-only reviewer is sufficient; implementation delegation must have exact non-overlapping ownership. Forecast beyond B3 or need for substantial new runner orchestration returns for split/budget review.

Permitted deviations: measured feedback target adjustments only through explicit Codex disposition; a later small fast-job profile only with reviewed capacity/compatibility evidence. No silently weaker tests, missing hosted evidence, optional full merge gates, runtime port widening, extra vendors/privileges or runner purchases.

Stop affected work for unresolved ADR/Plan/binding; insufficient quota/capacity/access; unsafe daemon/path isolation; substantial harness redesign; new schema/API/vendor/deploy requirements; architecture contradiction; or exhausted budget. Report one concrete blocker rather than repeatedly rerunning local checks to resolve an external prerequisite.

## 10. Revision history

- 2026-10-06 — Owner requests a bounded G003 revision after excessive validation turnaround, fast/full gates and GitLab hosted-runner qualification. Small was the initial cost-saving recommendation; owner asks why not medium, so medium is the proposed first profile. The owner subsequently notes fast-job cost should be modest; medium is the default for fast, integration and full jobs, avoiding a second qualification prerequisite. No implementation or hosted action occurred during drafting.

Primary references: [hosted Linux runners](https://docs.gitlab.com/ci/runners/hosted_runners/linux/), [compute-minute factors](https://docs.gitlab.com/ci/pipelines/compute_minutes/), [job rules](https://docs.gitlab.com/ci/jobs/job_rules/), [Docker-in-Docker](https://docs.gitlab.com/ci/docker/docker_in_docker/).


## 11. Codex correction review and revision planning findings — 2026-10-06

**Formal G003 outcome: CORRECTION LOOP — remaining CI revision/hosted evidence; corrections 1–4 are resolved locally.** No new defect requiring a second pass over these four corrections was identified in the reviewed code/evidence. This is not implementation ACCEPT, release of ADR-002 or approval to execute this draft. The next step is the short planning-only revision in §3.1; no additional full local run is requested by this review.

### Reviewed submission and evidence

Reviewed execution checkout HEAD `cbffd6da5d9936c90fdfd9e51d84d36c0f17ad81`, 192 NUL-enumerated changed paths, and Result SHA-256 `b91aa8764a8641852ed3ec277215def3c86e0bd71ebb8aef91552102e1091000`. Comparison with the supervisor's correction-pass manifest finds only the four authorized implementation/doc files and appended Result changed; implementation file hashes match Result §14.2. All five protected input hashes still match the original snapshot, and the original Result sections 1–13 match their submitted digest (before the appended section separator).

Independently inspected retained evidence at `~/.local/state/platform-evidence/g003-69407d/g003-exec/`: `c1-all` has 69 PASS and 3 INFO with no duplicate check IDs, FAIL or UNMET; root/off/rename browser reports have 103/103/2 expected, zero skipped/flaky/unexpected; each language report contains exactly the 120 unique declared `(id, valid, class)` tuples; leftover inventory is empty. `c1-final` has 27 PASS and 3 INFO. Wrapper logs preserve harness exit 0 and successful postchecks; local CI parity ends with status 0 and all phases 0. These are local results, not hosted evidence. No Docker/build/full-suite execution was performed by Codex during review.

- Outage correction preserves required-service 503, optional intelligence semantics, O7's exact `200/ok|503/timeout` pairing, context/log/body assertions and existing request/recovery bounds; stopped-service probe results now use the already reviewed DNS/deadline interpretation.
- Normal contract-report completion now checks typed/unique entries, consistent counts, skips/mismatches and the exact expected tuple set. V13a–V13f cover empty/truncated/duplicate/altered reports, valid reports and isolated malformed-count/type/skip/mismatch failures through the real completion function. The empty `match`-key false-green path is closed for normal reports.
- Nested-control cancellation uses owned launch/process identity, bounded termination and reaping before resource sweep. V18l–V18n prove TERM/HUP/deadline behavior with a TERM-resistant descendant, delayed resource-creation checks and surviving unrelated decoys. The reviewed 89-control log reports exactly 89 passing defined controls.
- N1–N16 now require exact expected rejection reasons in job output and attribution evidence, alongside existing exit/no-harness/publication assertions. The valid attribution cases remain covered.

Original B7 remains 6/7. Correction pass 1 reports 0/3 meaningful failure/fix cycles; all correction validation runs passed, so that accounting is consistent with BUDGET_POLICY. No reset or new implementation authorization follows from this review.

### Specific constraints the revised Plan must address

| Finding | Required guarantee / smallest planning correction | Minimum revision validation |
|---|---|---|
| F1 — Current `validate.sh` launches `run.sh all` and `do_complete` always requires the full required-check inventory plus root/off/rename browser and both contract reports. | Make execution and completeness profile-aware using the existing entry point. Each profile has a fixed reviewed required-check/report set; full retains the complete inherited set. The selector/report cannot redefine required checks from arbitrary caller data or manufacture PASS rows for excluded checks. Include profile/reason in safe artifacts. | Focused controls: each declared profile passes with exactly its required reports; a selected check/report omitted or failed rejects it; an explicitly unselected full-only report may be absent in fast; a fast result or tampered profile cannot satisfy full. Retain existing full-report V13 checks. |
| F2 — `B-CI-CONFIG` currently insists on one job and forbids job image/services/rules/needs; `validate.sh` supports only staging/default push and MR attribution. | Reconcile this exact guard to the reviewed YAML/profile invariants and hosted DinD setup, preserving gating, safe artifact allow-list and attribution. Add only the ordinary feature-push, schedule and manual/release-intent contexts actually used; unsupported/contradictory metadata still rejects. Do not turn the guard into unrestricted YAML acceptance. | YAML/profile rule cases from AC7b; new supported-source positive cases and intended-reason negative cases; retain applicable N1–N16 and passing MR-source/merged-result cases. No new execution framework or generalized GitLab emulator. |
| F3 — RP6/RP7/RP8/RP9 assume loopback/IP/path parity with the job shell; hosted DinD changes those namespaces. | Identify the exact CI-only namespace/path arrangement before claiming qualification. Run the actual harness from the namespace where its loopback/default-port and disposable bind-path assertions remain true; preserve tested checkout attribution and local commands. Do not substitute service-internal health calls for host-port/isolation proof or expose ports on all interfaces. | One bounded approved hosted qualification demonstrates checkout/temp-file mount identity, loopback reachability and non-loopback denial, default/variant ports and Docker/browser access from the actual validation execution context. Failure returns a concrete compatibility blocker; no repeated local full runs. |
| F4 — RP11 requires >=4 CPUs, >=8 GiB memory and >=40 GiB free in both temporary and daemon storage; RP12/global budget require a three-hour CI job and >=one-hour harness allowance. | Public medium capacity (50 GB nominal disk) does not prove 40 GiB free after setup. Measure usable daemon/job storage and peaks. Separate finite profile-specific timeouts/capacity checks where the smaller selected workload justifies it, preserving full capacity and cleanup grace unless measured evidence supports an explicitly reviewed adjustment. Fast jobs must not inherit a mandatory three-hour profile. | Focused preflight controls for profile limits, insufficient resources and malformed/insufficient time budgets; hosted qualification records actual free/peak capacity and durations. No blind lowering of full limits, automated size upgrade or paid purchase. |
| F5 — Shared suite names hide cost; static includes full CI fault controls. | Enumerate check membership and measured expensive components; fast is not simply the existing `static` suite. Keep validation-framework fault controls in full and run affected controls when their implementation changes. Reuse existing unit/contracts/boundary checks with explicit dependency coverage. | Timing breakdown for implemented profiles and exact membership comparison; changed controls rerun once with intended expected outcomes. The 48-minute local full result replaces the earlier two-hour forecast as the observed reference, not a hosted SLA. |
| F6 — Result §14.3/§14.6 still records absent API access; that evidence predates the owner's successful glab login. Hosted compatibility, quota and action approval remain unproven. | Preserve the historical investigation; append current access facts and remaining runner/security/quota prerequisites in the next Result revision. Resolve existing A-R1/H1/H2/H3/evidence permissions against the new profile and committed ancestors. ADR-002 and a fresh exact binding remain prerequisites. | Read-only user/project/runner/quota/permission facts plus attributable hosted qualification, final exact-tree full pass and intended failing gate. No local parity substitution, no permanent project-setting/schedule write without named authorization. |

The revised Plan should map F1–F6 to concrete changes and AC7a–AC7g in a short table. These findings specify the integration gaps; they do not require rebuilding unchanged protections or performing new pre-approval full-suite demonstrations. G003 implementation ACCEPT still waits for the revised CI phase and hosted evidence; G048 execution and the G021 product barrier remain gated.

## 12. Approved P0 evidence disposition — 2026-10-06

Codex accepts the P0 follow-up review and bounded continuation (not included). It approves minimum pre-mount credential sanitization/fail-closed absence checks, accepts the verified shorter-script-timeout cleanup behavior, and authorizes one final capacity-guarded H1 on medium with a provisional hosted integration/full 35 GiB start floor and 4 GiB runtime reserve. Local full/integration 40 GiB and CPU/RAM requirements remain. Disk measurement/guard failure must fail the overall job independently of harness success. Permanent qualification uses measured incremental workload consumption, explicit 1.5× margin plus reserve and Codex disposition; no automatic upgrade or presumed capacity pass.

This amendment and its immutable review are part of a fresh nine-input approval binding; historical five-input protection and Plan r2 bytes remain unchanged. Minimal changes stay in r2's reviewed paths and B3, with focused credential/guard controls and final H1 evidence. No further pre-implementation full qualification, new framework, secret exposure, purchase, settings/schedule action or waived G003 acceptance requirement is authorized.
