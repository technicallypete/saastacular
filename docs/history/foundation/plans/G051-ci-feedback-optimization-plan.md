> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G051 — CI feedback and build optimization: Implementation Plan (r3)

Goal: [G051-ci-feedback-optimization.md](../goals/G051-ci-feedback-optimization.md) · Spec: [G051-ci-feedback-optimization-spec.md](../specs/G051-ci-feedback-optimization-spec.md) · Inputs: G051-planning-inputs.json (not included)
Author: Claude Code (planning only) · Date: 2026-10-09 UTC

**Status: Plan r3, submitted for Codex Plan review after the r2 CORRECTION LOOP, in which only C4 remained. Implementation is not authorized.** No build, harness, control, Docker or pipeline was run.

r3 amends only:

- the build-ledger identity and accounting (§3.8);
- the VC controls (§6.1);
- the AC5 mapping (§7);
- the risk text (§12);
- the inventory.

C1–C3 and Q1–Q7, resolved by the r2 review, are not reopened; every other section is r2's text.

**History**

| Artifact | Path | SHA-256 |
|---|---|---|
| Plan r1 (commit `19dc50e`) | G051-ci-feedback-optimization-plan-r1.md (not included) | `2f387259c876b986b54bc8cea7cd4fd8e70a7c2d8c59d231d2ebbfa6ff926dc9` |
| r1 inventory | G051-planning-inputs-r1.json (not included) | `cfaba839f547cde8acc33afc056e1680866234e9cbf83e43cb035015b158ab41` |
| Codex review of r1 (`origin/staging` `26b14e8`) | `management/v0/plans/G051-ci-feedback-optimization-plan-codex-review-r1.md` | `2e67522c28914578491c559732dfa075546706062b01a91e28d213b169341ec8` |
| Plan r2 (commit `d08bc60`, archived byte-identically) | G051-ci-feedback-optimization-plan-r2.md (not included) | `68e020d9cd781dfb79b80c732c312c6399cdb72aa24c1af0899b6acd253a4677` |
| r2 inventory | G051-planning-inputs-r2.json (not included) | `6267ba8c176531503151ea52264d4fc3a1eda6f71ee099b5fb0f7662d0d8ef06` |
| Codex review of r2 (`origin/staging` `26b14e8`) | `management/v0/results/G051-plan-r2-codex-review.md` | `4dad1be27ee48cf12d4260e906cdc37555fb584f0fdc528849e27e8d2b1c2791` |

- **Planning baseline:** unchanged from the r2 review, `570b40caf47ff9257c5b5aa0588f60fecd58d0a6`, tree `a09895950c8077f7f50b116d05edeb76c81c657e`, with G005 integrated (§10). `origin/staging` has since advanced to `26b14e8`, but only with management documents (G006 planning, the two G051 reviews, PROGRESS); no runtime, CI or harness file changed. It is not an execution snapshot. G005's formal Codex ACCEPT remains an execution prerequisite; PROGRESS at `26b14e8` still lists it as pending.
- **Tier and budget:** architecture/security/data execution tier. Budget **B5**, checkpoint after cycle 3 (unchanged; §8).

## Revision record (r3)

| Item | r3 change | Where |
|---|---|---|
| C4.1 Build identity | The ledger key now also contains: the resolved base and tool image identities (FROM bases, named stages from external images, the Dockerfile `syntax` frontend), the target platform, the effective Dockerfile and ignore configuration (one documented precedence; both candidates recorded), and the context entry inventory (type, mode, owner, symlink target, content digest) for the main and named contexts. All of it comes from BuildKit progress, `docker image inspect`, `docker version`, `--iidfile` and `compose config`. No dockerignore evaluation or orchestration library is added. | §3.8.1 |
| C4.2 Evidence completeness | A confirmed zero-step cache hit requires fully parsed progress. Absent, unparseable or unexpected progress, an unknown key field, or an unaccounted image change is `unknown`: it stays in the denominator and makes AC5 **UNMET**. `ledger_check` is deterministic and is re-run on published artifacts. | §3.8.2–§3.8.3 |
| C4.3 Gating | `BUILD-LEDGER INFO` stays non-gating for merges. Any unexcepted duplicate or near-duplicate executing build, or any incomplete evidence, in a required AC5 run **blocks G051 acceptance** and stops before readiness is claimed. This answers r2 §12 R1. | §3.8.4, §7, §12 |
| C4.4 Controls | VC3 is tightened; VC5–VC8 are added: changed base identity, platform, effective context or ignore configuration, and unknown or unparseable progress or missing accounting. | §6.1 |
| Scope note | The hosted adapter's one pre-harness context-image build is recorded as a `bootstrap` row (§3.8.3); §12 Q1 asks whether that is acceptable. | §3.8.3, §12 |

## Revision record (r2)

| Item | r2 change | Where |
|---|---|---|
| Baseline | Refreshed from `8f5f453` to `570b40c` (G005 integrated). Counts reconciled: 83 rows, fast 33, integration 71, full 81, staging 5, one `F-tenancy` control. Post-G005 hosted fast job measured at 368.0 s; the gate was still running when read. | §2, §3.1, §10 |
| **C1** Dependency selection | Supported Go, web and Python import forms are listed; unsupported, dynamic or ambiguous forms give full. Old (BASE) and new (HEAD) edges are both scanned. Test- and fixture-only obligations are preserved. A per-kind file allow-list stops scripts, build or config inputs under an area root from narrowing. `registered-absent`/`area-removed` rules are defined, and unresolved inputs give full. | §3.3 |
| **C2** Partitions and union | Each job's partition and the pipeline union have deterministic expected sets, bound to the candidate and selection. Coverage is marked complete or partial, and the gate reports `gate:remainder`. A small evidence check, `pipeline-union.sh`, is the release-reuse rule; the r1 identical-hash rule is withdrawn. | §3.7, §6.4 |
| **C3** Background work | Explicit supervision: private outputs per child, wait-status capture and a `BACKGROUND` row, join before `final`, and an EXIT trap that stops and reaps children before ownership cleanup. Prebuild is limited to affected runs and can never mask an authoritative build. | §3.6 |
| **C4** Build reuse | An input-bound ledger records the Dockerfile, target, context and argument digests plus executed/cached step counts. The accounting rule for prebuilds is defined, existing validity controls are referenced, and focused build controls are added. | §3.8 |
| Q1–Q7 | Codex decisions applied. L6 (Playwright workers) is now planned, with stability checks. Q5/Q7 conventions are kept subject to C1. | §3.2–§3.5 |
| Sequence | S4 stability and S7 final qualification are merged: focused concurrency checks first, then a single final full local qualification after the rebase. | §8 |
| Projection | Recomputed: L4 removed from full pipelines, L6 added: ≈40–41 min. | §3.5, §5 |

## 0. Governing rules

The dated owner rules at the baseline apply:

- AGENTS.md, *Temporary solo-operator validation and integration policy*, including the owner status rule;
- GITLAB-BRANCH-WORKFLOW.md: *Temporary solo-operator gate placement*, *Rebase before final local qualification* and *Staging pipeline acceptance status*;
- OPERATIONS.md, *Solo-operator MR validation order*.

**Superseded wording.** The Goal's *"Hosted CI supplies final equivalent automated coverage"* and the Spec amendment's *"Final ACCEPT and dependent execution await required staging success"* are read through those rules. Neither the Goal/Spec bytes nor any historical approval is edited. Concretely:

- **Order.** Implementation and focused checks come first. Then fetch and rebase onto `origin/staging` and reconcile. The required final **local** full coverage runs on the rebased candidate before the push and MR.
- **MR pipeline.** `fast` runs automatically. The `gate` stays manual and optional, and this Plan never starts it.
- **Staging push.** It runs the automatic gate. For the CI status gate, `success` or `canceled` qualifies unless a job failed.
- **Canceled runs** never prove the timing or coverage of AC6–AC8.

**Codex decisions on r1 (applied):**

- **Q1:** narrowed affected static checks are allowed under the corrected C1 mapping.
- **Q2:** one browser execution is allowed if every assertion and scenario is kept.
- **Q3:** pipeline-level completeness is allowed subject to C2.
- **Q4:** worker count and the group-E contingency are included, with stability and isolation checks; resource limits are unchanged.
- **Q5/Q7:** the conventions stand, subject to C1. No product feature and no actual area registration is authorized.
- **Q6:** the design for hosted runs H2–H4 stands; no push, MR or merge is authorized by the review.
- **Split:** B5 is unchanged and no G051a/G051b is created.

G021 stays closed.

## 1. Approach

The existing machinery is optimized while keeping the job structure (`fast` → `gate`) and the semantics of `.gitlab-ci.yml`. No cache, registry, vendor or orchestration framework is added. There are five parts:

1. **An affected profile.** An explicitly registered feature area is a fail-closed path → obligation mapping, whose dependency closure over-approximates (C1). In MR pipelines the automatic `fast` job executes the narrow obligation set of a pure-area change. Security and global-UI paths are widened to full (R3).
2. **Suite-level obligations.** The `lint` suite is split with no ID change, and five new IDs are added. Completeness is checked against the executed suite list. A **pipeline partition** (fast and gate) is verified against its union (C2).
3. **Full-pipeline time.**
   - The gate does not repeat the fast job's suites (L1).
   - B-BROWSER-ROOT executes once (L2).
   - The control matrices run concurrently with the stack suites, under explicit supervision (L5, C3).
   - The harness's own browser runs use Playwright `workers: 2` (L6).
   - Cache-only prebuild is used only in affected runs (L4).
4. **Evidence.** An input-bound build ledger (C4), per-control timing, operations documentation, two representative hosted runs and the ordinary staging full run.
5. **Safety of new execution.** Background children are supervised and reaped on every exit path (C3). Prebuild can never mask the authoritative build.

## 2. Measured bottlenecks (AC1; R1)

### 2.1 Hosted baseline: G050 MR pipeline 2930980752

Exact source `b0ac030`; selection `full`/`foundation`. Evidence: `<home>/.local/state/platform-evidence/g050-exec-0d52421/hosted/`.

| Job | Duration | Before harness | Harness | After harness | Queue |
|---|---:|---:|---:|---:|---:|
| fast `17063444297` | 360.2 s | 33 s: runner 14, clone 1, hosted-context 17, preflight 4 | 316 s | 11 s | 0.41 s |
| gate `17063444298` | 3807.3 s | 34 s | 3725 s | 48 s: finalize 30 (cleanup 17), upload 18 | 0.22 s |

The pipeline sum is **4167 s (69.5 min), or 139 compute-minutes**.

**Suites by job:**

- **fast:** lint 148, provenance 15, contracts 59, unit 91.
- **gate:** lint 160, **controls 1005** (B-CI-CONTROLS 972 for 281 controls; B-DEPS-CONTROLS 32), provenance 16, contracts 52, **root 1287**, migrations 49, ops 427, rename 63, **off 658**, final 2.

**Inside the suites:**

- **root:**
  - stack start → B-DEFAULTS: 230 s, a cold build with MinIO compiled from source;
  - **B-BROWSER-ROOT: 843 s**: the README verbatim run took about 481 s including the cold web-browser build, and the report run about 362 s (Playwright duration 354.9 s; tokens-light/dark 217 s);
  - B-CHECKS-ROOT 50 s, B-DOCS-EXEC-GS 63 s, error outputs 40 s.
- **off:** B-OBS-OFF 199 s; B-BROWSER-OFF 454 s (Playwright 370 s).
- **ops:** observability 185 s, identity (IT 60 s, E2E 84 s, probes 36 s), outage 18 s, persistence 22 s.
- **fast lint:** B-SHELL 29 s, B-WEB-ASSET-CONTROLS 33 s, B-MCP-SOURCE 39 s (cold core-check build), B-CI-CONFIG 40 s (about 110 `yq` containers).

### 2.2 Post-G005 and corroborating evidence

- **Post-G005 staging push pipeline 2931683534** (`570b40c`; read-only GET at 2026-10-09T18:50Z):
  - **fast `17068999910` succeeded in 368.0 s** (queue 0.27 s);
  - gate `17068999911` was **running** (2275 s elapsed). It is not yet evidence; its terminal result refreshes the baseline at S0.
- **G049 MR pipeline 2929968148:** fast 364.6 s, gate 3632.8 s. Variance is about ±5%.
- **G005 local runs** (read only; warm daemon, a slower host): the controls suite took 1688 s (r3), 1743 s (r4) and 1930 s (r5, the "about 32 min"; B-CI-CONTROLS 1843 s, 282 controls). B-BROWSER-ROOT took 644 s in r5.
- **G050 local:** 18–19 focused controls about 3.5 min; lint plus provenance about 4 min.

### 2.3 Findings

| # | Finding | Cost per pipeline |
|---|---|---:|
| F1 | The gate repeats the fast job's lint, provenance and contracts on the same commit. | 228 s |
| F2 | B-BROWSER-ROOT runs the full 141-test suite twice. | ≈355 s |
| F3 | B-CI-CONTROLS runs serially and is latency-bound (≈3.5 s per control), while nothing else runs. | 972 s |
| F4 | Cold builds sit on the critical path: the stack (230 s), web-browser (≈126 s), and the fast job's check and contract images. | ≈400 s |
| F5 | Every service path selects integration: root/ops/rename/off ≈41 min, plus ≈5 min of repeated fast suites. Feature browser tests select full. The R3 security/global paths are only integration today. | — |
| F6 | Three full Playwright suites per full pipeline run with `workers: 1`. | ≈1080 s |
| F7 | Fixed per-job overhead: about 44 s for fast, 82 s for gate. | small |

**Same-job reuse alone does not reach the targets.**

## 3. Design

### 3.1 Suites, IDs and execution obligations (R2; the Goal's mapping)

`required-checks.tsv` keeps the `ID<TAB>suite` format. Suites are relabelled with no ID change, and five rows are added.

| IDs | Suite | fast | affected | integration | full | Input boundary |
|---|---|:-:|:-:|:-:|:-:|---|
| B-LITERALS, B-NAMING, B-IDENTITY-GUARDS (3) | `lint-base` | ✓ | always | ✓ | ✓ | Repository-wide |
| B-WEB-BOUNDARY, -TOKENS, -COMPONENTS, -DESIGN, -ASSET-CONTROLS (5) | `lint-web` | ✓ | web areas | ✓ | ✓ | The whole web tree |
| B-MCP-SOURCE (1) | `lint-core` | ✓ | core areas | ✓ | ✓ | The core module |
| B-SHELL, -HADOLINT, -COMPOSE, -ENV-EXAMPLE, -DEPS, -MIGRATE-SET, -DOCKERIGNORE, -CI-CONFIG (8) | `lint` | ✓ | — | ✓ | ✓ | Foundation-class inputs only, so a change already selects full |
| B-RENDER-ID/PLAN/REPEAT, B-PROTECTED, B-SKILL-UNCHANGED (5) | `provenance` | ✓ | always | ✓ | ✓ | Repository-wide |
| B-CONTRACT-* (6) | `contracts` | ✓ | intelligence areas | ✓ | ✓ | contracts, intelligence, tools |
| B-UNIT-CHECKS | `unit` | ✓ | — | ✓ | — | Unchanged; full keeps B-CHECKS-ROOT |
| B-DEPS-CONTROLS, B-WRAPPER-CONTROLS, B-CI-CONTROLS | `controls` | — | — | — | ✓ | Unchanged scenarios; supervised background (L5) |
| Root 17 (incl. B-TENANCY-HTTP) | `root` | — | — | ✓ | ✓ | B-BROWSER-ROOT single run (L2) |
| Migrations 8 (incl. B-TENANCY-CATALOG/MIGRATION/IT) | `migrations` | — | — | — | ✓ | Unchanged |
| Ops 12, rename 5, off 4 | | — | — | ✓ | ✓ | Unchanged |
| B-IDENTITY-STAGING | `staging` | identity-staging job only | | | | Unchanged |
| B-DOCKER-ONLY, B-SECRET-CANARY, B-SCOPE, B-REQUIRED-CHECKS (4) | `final` | every job | | | | Unchanged |
| **new** B-UNIT-WEB / -CORE / -INTEL | `unit-web/-core/-intel` | — | per area kind | — | — | The README `platform:check` line of that service plus secret-scan (an exact subset of B-UNIT-CHECKS) |
| **new** B-FEATURE-WEB | `feature-web` | — | touched web areas and stack-spec consumers | all registered present | all | §3.4 |
| **new** B-FEATURE-CORE | `feature-core` | — | touched core areas with integration tests, plus importers | all | all | §3.4 |

**Counts** at `570b40c` (actual) and after G051:

| | At `570b40c` | After G051 |
|---|---:|---:|
| Registry rows | 83 | 88 |
| fast | 33 | 33 |
| integration | 71 | 71 |
| full | 81 | 81 |
| staging | 5 | 5 |

With no registered area, the feature suites add nothing. **Affected examples:** web only 19 IDs, core only 15, intelligence only 19.

**Partitions (C2):**

| Pipeline | fast job | gate job | Union (distinct) | Overlap |
|---|---:|---:|---|---|
| full | 33 | 53 (remainder) | **82** = the 81 full IDs + B-UNIT-CHECKS | exactly the 4 `final` IDs |
| integration | 33 | 42 (remainder) | 71 | the 4 `final` IDs |

**Test-path obligations:**

- **Full:** `web/tests/browser/*.spec.ts` and `support/`; `web/tests/fixtures/{error-probe,token-scopes}`; `web/tests/unit/` (widened); foundation-package `*_test.go`; `intelligence/tests/conformance_exchange.py`; `infra/validation/tools/**`.
- **Integration:** the other `intelligence/tests/*.py`.
- **Area:** `web/tests/{browser,fixtures}/features/<n>/` and an area package's own tests. Unregistered `features/` directories stay foundation, so they select full.

### 3.2 Path map

**Widened to full (R3; Q7 accepted):**

| Class | Paths |
|---|---|
| Identity and authorization | `core/internal/identity/`, `core/cmd/server/`, `web/app/{auth,sign-in,account}/`, `web/app/api/account/`, `web/components/account/`, `intelligence/src/security/` |
| First paint and tokens | `web/app/globals.css`, `web/app/layout.tsx`, `web/public/`, `web/components/layout/` |
| UI primitives | `web/components/ui/` |
| Shared helpers | `web/lib/`, `web/tests/unit/` |
| Contracts | `core/internal/mcp/`, `intelligence/src/streaming/` |

`core/internal/tenancy/` is already foundation (G005). **Unchanged:** every existing row, and integration for `core/internal/health/`, `web/app/page.tsx`, `web/app/api/{health,ready}/`, `intelligence/src/{main.py,models,observability,providers}/` and the other runtime prefixes. `F-review-web` changes from integration to full.

**Area grammar** (Q5 accepted subject to C1). The classes are `area-web`, `area-core` and `area-intel`; the area name is the last path segment, and the row kind is `dir` only.

| Class | Allowed roots |
|---|---|
| `area-web` | `web/app/(features)/<n>/`, `web/tests/browser/features/<n>/`, `web/tests/fixtures/features/<n>/` |
| `area-core` | `core/internal/<n>/` |
| `area-intel` | `intelligence/src/<n>/`, `intelligence/tests/<n>/` |

Names match `^[a-z][a-z0-9-]{0,31}$` for web and `^[a-z][a-z0-9]{0,31}$` for core and intelligence. A duplicate, an overlap with another row, a bad root or a bad name exits 16. No `web/features/` root is allowed until a reviewed guard extension. An area exists only through explicit rows (a foundation change); **G051 registers none.**

### 3.3 Selection: over-approximating dependency closure (C1)

The selector stays bash and git only: no Docker, no network, no graph framework. The outputs gain `AREAS`, `PIPELINE_SUITES` and `EXECUTED_SUITES`, and `selection.tsv` gains `area` and `via` columns. Ranks are fast 1 < affected 2 < integration 3 < full 4. Any doubt resolves to the higher rank.

1. **Unchanged rules.** Documentation and agent handling, workflow admission, the context rules (schedule, web, api and merged-result → full; feature push → fast), base checks → full, unsafe or unknown path → full.

2. **Area file allow-list** (C1: registration never narrows scripts, build or config inputs). A path inside an area root is *area-ordinary* only for:

   | Kind | Area-ordinary files |
   |---|---|
   | web | `*.ts`, `*.tsx`, `*.css`, `*.json` (not `package*.json`, `tsconfig*.json`, `*.config.*` or `components.json`), `*.svg/png/webp/jpg` data, `*.md`, `*.txt` |
   | core | `*.go`, anything under `testdata/`, `*.md` |
   | intelligence | `*.py` (not `conftest.py` under `src/`), `*.json/yaml/txt` under `testdata/` or `fixtures/`, `*.md` |

   - **Anything else selects full**, with reason `area-unclassified`. This covers `*.sh`, `Dockerfile*`, `go.mod/go.sum`, `*.sql`, `pyproject.toml`, `requirements*`, `.env*`, `*.config.*`, `Makefile`, and symlinks or submodules (git modes 120000 and 160000).
   - **Web `route.ts|route.js`** (BFF handlers) selects integration, with reason `area-route-handler`.

3. **Change records:**
   - **A or M** on an area-ordinary path selects that area.
   - **D, R or C:** the old and the new path are each classified. When both are area-ordinary, the union of their areas is selected; otherwise the existing `deletion-or-rename` rule applies → full.

4. **Dependency closure over BASE and HEAD.** For every touched area, references are searched with `git grep` in **both** the comparison-base tree (old edges) and HEAD (new edges). The search covers every tracked file outside that area, **including tests, fixtures, specs and testdata**. A referencing file's own classification joins the selection:

   | Referencing file | Effect |
   |---|---|
   | another area | union |
   | runtime | integration |
   | foundation or `core/cmd/*` | full |
   | Markdown documentation (the docs class) | contributes no obligation, consistent with today's documentation rule. Executable README and foundation-doc blocks are already exercised only by root/full checks. |

   The supported, conservative forms are:

   | Kind | Detected as a reference to area `<n>` |
   |---|---|
   | Go | Any occurrence of the string `core/internal/<n>` followed by `"` or `/`. This covers subpackages `core/internal/<n>/sub/...` and every import form (aliased, dot, blank), plus any non-Go text naming the path (embed/testdata/fixture paths). |
   | Web, alias | Any occurrence of `(features)/<n>` or `features/<n>` in any file, which covers the `@/app/(features)/<n>/…` alias (tsconfig `paths` has only `@/*`; tsconfig is foundation). |
   | Web, relative | Every quoted relative specifier (`'./…'`, `'../…'`) in `import`, `export … from`, `import()`, `require()`, CSS `@import`/`url()` and `new URL()` is resolved against its file's directory. A specifier landing inside an area root is a reference. |
   | Python | `import <n>`, `import <n>.<sub>`, `from <n>[.<sub>] import`, `… as …` in `intelligence/**`. Relative imports (`from .… import …`) are resolved against the file's package path; landing in an area is a reference. |

   **Ambiguous forms give full when any web or intelligence area is touched** (reason `area-ambiguous`; there are no such forms today):

   - in web files outside an area: a non-literal or template (`${`) specifier in `import()`/`require()`, a relative specifier escaping `web/`, or an unresolvable `(features)` reference;
   - in Python outside an area: `importlib.import_module(`, `__import__(`, `exec(` or `sys.path` manipulation, or an unresolvable relative import.

   **Iteration** runs to a fixpoint over newly joined areas. A failed `git grep` or an unreadable BASE gives full (`closure-failed`).

5. **Test- and fixture-only obligations are preserved.** Specs, fixtures and testdata outside an area that reference it carry their own class (rule 4). A foundation test referencing an area therefore selects full, and a spec in area B referencing area A's fixture selects B too.

6. **Cross-service consumers.** A touched core or intelligence area adds every registered web area containing `*.stack.spec.ts` at BASE or HEAD.

7. **Absent areas:**
   - **`registered-absent`:** the area has no tracked files at **both** BASE and HEAD. It adds no feature obligation and is reported as INFO. This is safe because no code or consumer can exist.
   - **`area-removed`:** the area had files at BASE and none at HEAD.
     - Its deletion records select its kind's `lint-<kind>` and `unit-<kind>`, plus every old-edge consumer from rule 4 at its class.
     - If any BASE consumer cannot be classified, the selection is full.
     - A deletion never drops an obligation silently: the consumer obligations replace the area's own feature suite, and the reason `area-removed` is recorded.

8. **Executed suites:**

   | Pipeline profile | Executed suites |
   |---|---|
   | full | full suites, plus feature suites for every registered area present |
   | integration | integration suites, plus the same feature suites |
   | affected | `lint-base provenance final`, plus per kind: web `lint-web unit-web feature-web`; core `lint-core unit-core` (+ `feature-core` with integration tests); intelligence `contracts unit-intel`. Any fast/agent-class path adds every fast suite. |
   | fast | the fast suites |

   **Per job:** the fast job runs `affected` or `fast`. The gate runs `none` for fast or affected, and for integration or full **in CI** the *remainder* (§3.7). Local runs are complete.

9. **Local selection** (refused in CI with exit 16):
   - `validate.sh --profile affected --base REV`: REV must be an ancestor, otherwise full.
   - `--compare OLD`: a tree comparison for post-rebase reconciliation, selecting exactly the suites to rerun. Unchanged suites' evidence is reusable only with exact tree and input provenance. An unknown OLD or a failed diff gives full.

### 3.4 Focused execution

- **`lint-*`:** regrouped, with unchanged assertions.
- **`unit-*`:** the README check line with exactly one `--env-file .env` replaced, plus secret-scan.
- **`feature-web`** (B-FEATURE-WEB, single execution per area set):
  - **Build:** web-browser `--target browser`, which includes the fixture stage's production `npm run build`.
  - **Run:** Playwright with `FEATURE_E2E=1` (feature projects only, only the 3001 fixture server, the `IDENTITY_E2E` pattern) in a run-labelled `docker run --rm --network none` container. Areas with `*.stack.spec.ts` instead run on a disposable full stack, `PORTS_FEAT` 3700/8780/8700/6132/7079/9700/9701.
  - **Assertions:**
    - each selected area with page routes has at least one spec, otherwise UNMET;
    - stats show expected > 0 with 0 skipped, flaky or unexpected;
    - S-ERRORS is present on every test.
  - **Inherited projects** gain `testIgnore: /features\//`, so the current 141 tests are unchanged.
- **`feature-core`** (B-FEATURE-CORE):
  - PostgreSQL, plus Redis or MinIO only when the selected packages import their clients;
  - `migrate up`;
  - `go test -tags integration -count=1 -v` over the selected packages in the core-check image, with the 0600 env-file pattern;
  - asserts `ok` per package, at least one test passed, and no FAIL or SKIP.
- **Prebuild (L4, affected runs only):** the fast job starts cache-only builds, for the exact Dockerfile, target, context and arguments of the images its selected suites will build. They run under background supervision (§3.6) and are accounted in the ledger (§3.8). Each consumer **waits for its prebuild to end** (whatever its status) before running its authoritative build.

### 3.5 Full-pipeline levers and projection

| Lever | Change | Saving (estimate) | Status |
|---|---|---:|---|
| L1 | The CI gate executes the remainder of the pipeline profile; the fast job's suites are verified by the C2 union check, not by `needs` alone. | 228 s | Planned (Q3) |
| L2 | B-BROWSER-ROOT executes the README `platform:browser` line **once**, with only the reviewed report injection (the B-IDENTITY-E2E pattern). Every assertion is kept: matrix, S-ERRORS, init hash, saved response bodies, Chromium version, and exit 0 of the documented command. | ≈355 s | Planned (Q2) |
| L5 | The `controls` suite runs as a supervised background child when it is selected with any later suite (§3.6). **Isolation:** scratch roots, run-token labels, no host ports, no checkout writes (V4 copies only tracked/non-ignored files), ownership cleanup. **Group-E contingency:** run group E after the stack suites. | 800–950 s | Planned |
| L6 | The harness injects `PLAYWRIGHT_WORKERS=2` for its own runs: root, off, identity E2E and feature runs. The config default stays 1, so the README is unchanged. CPU, memory and timeouts are unchanged. | ≈340 s | Planned (Q4); stability checks in §6.2 |
| L4 | Cache-only prebuild in affected runs only, not in full pipelines. | 60–90 s (UI) | Planned |
| L3 | B-CI-CONFIG in two `yq` passes. | ≈35 s | Reserve; checkpoint decision |

**Projection** of the full pipeline after G005 (estimates; the gate baseline is refreshed from 2931683534 at S0):

| Component | Seconds |
|---|---:|
| fast job (measured post-G005) | ≈368 |
| gate foreground: root (1287 − 355 L2 − ≈155 L6 + ≈10 tenancy) | ≈787 |
| migrations (49 + ≈60 tenancy) | ≈109 |
| ops (427 − ≈29 L6) | ≈398 |
| rename | 63 |
| off (658 − ≈160 L6) | ≈498 |
| final | 2 |
| contention while controls overlap | 100–150 |
| gate overhead | 82 |
| **Total** | **≈2408–2458 (40.1–41.0 min)** |

Without L6 the total is ≈45.9–46.7 min, so **L6 is required**.

### 3.6 Background supervision (C3; L5, L4)

New bounded helpers go in `infra/validation/lib.sh`: `bg_start NAME -- CMD…`, `bg_join NAME`, `bg_stop_all`. This is not a framework: one record per child, existing `record` and `registry_row` only.

- **Private outputs.** A child runs as a subshell in the harness process group, so validate.sh's group TERM and KILL reach it. It sets its own `RESULTS`, `TIMINGS` (`record` writes `${TIMINGS:-$WORK/timings.tsv}`; one-line helper change) and log under `$WORK/bg/<name>/`. Every long-running command it launches is started as `cmd & echo $! >>supervised` and then waited for.
  - **No child writes the shared results, timings or ownership files.** The controls suite writes no ownership rows (verified: `static_deps_controls`, `static_ci_controls` and the wrapper controls call no `registry_row`, `docker_run_owned` or `record_built_image`). Should one ever be needed, the child writes a private fragment that the parent appends with `registry_row` before cleanup.
- **Join (normal path, before `final_checks`).**
  1. `wait PID` captures the exit status.
  2. The parent validates the child's private rows: a well-formed format, only that suite's IDs, each once.
  3. The parent appends them to `RESULTS` and the timings, appends the child log to the harness log, and records `BACKGROUND INFO "controls joined (exit 0, N s)"`.
  4. **On a nonzero status, malformed rows or foreign IDs:** the rows are still merged as evidence, and the parent records **`BACKGROUND FAIL "<name> exited <rc>"`**. That row makes run.sh exit 1 and completeness fail (14), so PASS rows emitted before a nonzero exit can never pass. This follows the existing non-required `PREFLIGHT` row pattern.
- **Every exit path.** run.sh's trap becomes `trap 'bg_stop_all; cleanup' EXIT`, so children are stopped **before** ownership cleanup removes `$WORK` or resources. `bg_stop_all`, for each child not yet joined:
  1. TERM each recorded PID and the subshell;
  2. a bounded wait (`BG_GRACE_S=60`, under validate.sh's 150 s grace; the controls runner's own grace is 30 s);
  3. KILL any survivor, `wait` to reap it, and verify that no recorded PID is alive;
  4. record `BACKGROUND FAIL "<name> stopped before join"` (and `… survived KILL` if needed). **Unjoined rows are never merged.**

  Cancellation (validate.sh group TERM), timeout (validate.sh deadline → group TERM), foreground failure (main-shell exit) and normal completion all pass through this trap. validate.sh finalization and publication start only after run.sh has exited, and cleanup's `inventory-leftover.txt` still fails completeness for any leaked resource.
- **Prebuild children.** A prebuild failure records `PREBUILD INFO "<image> exit <rc>; authoritative build follows"`. The consumer always runs its own authoritative build, whose failure records the owning check's FAIL. An unfinished prebuild is stopped by `bg_stop_all`. A prebuild never records PASS and never removes the need for the consumer build.

### 3.7 Pipeline partitions and union (C2)

**Deterministic sets.** Each job computes the same selection from the same candidate, map and CI variables.

- **`selection_digest`**: SHA-256 over the candidate commit and tree, the profile-paths.tsv, select-profile.sh and required-checks.tsv blob IDs, the comparison base, `AREAS` and `PIPELINE_SUITES`.
- **E_pipeline**: the required IDs of `PIPELINE_SUITES`, plus `unit` when the fast job runs it, plus `final`.
- **E_fast**: the required IDs of the fast job's suites plus `final`.
- **E_gate**: the required IDs of the remainder (`PIPELINE_SUITES` − fast-job suites) plus `final`.
- **Invariants:** E_fast ∪ E_gate = E_pipeline, and E_fast ∩ E_gate = exactly the 4 `final` IDs. For full this gives 33 + 53 rows → 82 distinct obligations; for integration 33 + 42 → 71. Emitted-row totals are never presented as distinct coverage.

**Per-job evidence** in summary.json, format `/1` plus fields:

- `pipeline_profile`, `pipeline_suites`, `pipeline_set_sha256` (E_pipeline);
- `executed_profile`, `executed_suites`, `required_set_sha256` (that job's own set);
- `coverage` = `complete` or `partial` (partial when E_job ≠ E_pipeline), `partition` = `fast` or `remainder`;
- `selection_digest`, candidate commit and tree.

**In-job completeness** (validate.sh) is unchanged in strength: exactly E_job, each once as PASS, rows outside it rejected. It also re-derives E_pipeline and the partition from `selection.env` and fails (16) on an inconsistent partition.

**Truthful reports.** JUnit uses `fast:fast` for the fast partition and **`gate:remainder`** for the gate's partial coverage; junit.sh gains the tokens `affected` and `remainder`. The Tests tab therefore shows two partitions, never a gate labelled full. OPERATIONS states that full coverage is their verified union.

**Union check (evidence and reuse).** `infra/ci/pipeline-union.sh` is a small deterministic script, run outside CI jobs. It reads both jobs' published `summary.json`, `results.tsv` and `phase-status.tsv` (JSON through the pinned `yq` container, Docker-first) and **fails** when any of these holds:

- a partition is missing or not successful (job status or any phase ≠ 0);
- the candidate commit or tree differs;
- `selection_digest`, `pipeline_suites` or `pipeline_set_sha256` differs;
- a job's executed suites differ from its expected partition;
- the union ≠ E_pipeline (a gap);
- any overlap other than the 4 `final` IDs;
- a required ID is not PASS.

**Release-reuse rule** (replaces the r1 rule): a release may reuse a full pipeline pass only when `pipeline-union.sh` passes on both jobs' artifacts, both GitLab jobs succeeded, and each job's `images.txt` equals the candidate's (pins are a function of the tree). Each job keeps its own attribution, build ledger (§3.8) and cleanup inventory; image inventories are **not** compared across jobs.

### 3.8 Build ledger and cache accounting (C4; R4/AC5) — r3

**Cache decision (unchanged):**

- no GitLab `cache:`, registry, persistent daemon or credential;
- reuse comes only from the job's own BuildKit cache (prebuild included);
- hosted jobs are cold by construction;
- local runs are warm and are **not** AC5 evidence.

The ledger exists to *prove* AC5 for the required cold runs. It is built from existing Docker and BuildKit metadata and accepted inventories, with no dockerignore evaluator and no build-orchestration layer.

#### 3.8.1 Build key and provenance contract (C4.1)

`builds.tsv` holds one row per build invocation in scope (§3.8.3), with validated fields only. The **key** is the tuple marked K. Every K field must be known, or the row is `unknown` (§3.8.2).

| Field | K | Source (existing metadata only) |
|---|:-:|---|
| `job`, `suite`, `seq`, `kind` (`authoritative`/`prebuild`/`bootstrap`/`named:<ID>`) | | The harness |
| `dockerfile`, `dockerfile_sha256` | K | The effective Dockerfile path. For Compose this is `context` + `dockerfile` from `docker compose … config --format json` of the same files, env file and profiles; for explicit builds it is the `-f` argument. |
| `target` | K | `--target`, or Compose `build.target`, or `final` |
| `frontend` | K | The `# syntax=` directive and its resolved `docker-image://…@sha256:` from progress; `builtin` when there is none |
| `bases` | K | A sorted list of `stage-ref → name@sha256:<digest>` for every **external** `FROM` and `COPY --from=<external image>` reachable from the target. It is taken from the resolved `FROM <name>@sha256:<64hex>` vertices in BuildKit plain progress. For a base resolved from the local image store, the identity is `docker image inspect --format '{{.Id}}'` taken immediately before the invocation. Each name must also appear in the job's `images.txt` (the accepted pin inventory); otherwise the field is `unknown`. |
| `platform` | K | Explicit `--platform` / Compose `platforms`, otherwise the daemon default from `docker version --format '{{.Server.Os}}/{{.Server.Arch}}'`. Any `FROM --platform=` value in the Dockerfile is recorded too. For image outputs, `docker image inspect --format '{{.Os}}/{{.Architecture}}{{with .Variant}}/{{.}}{{end}}'` must agree, otherwise `unknown`. |
| `ignore_file`, `ignore_sha256` | K | The **effective** ignore file, by the documented BuildKit precedence. `<Dockerfile path>.dockerignore` (for example `infra/docker/web.Dockerfile.dockerignore`) is used if it exists; otherwise `<context>/.dockerignore` if it exists; otherwise `none`. When both exist, both are recorded and the Dockerfile-specific one is chosen. A Dockerfile on stdin has no context and `none`. No pattern is evaluated. |
| `context_inventory` | K | SHA-256 over the sorted `find <context> -printf '%y %m %U:%G %P %l'` lines, each regular file's line extended with its SHA-256. This covers entry type (file, dir, symlink, other), permission mode, numeric owner, relative path, symlink target text and content. Modification times are excluded. It covers the **whole** context; the ignore file identity (above) fixes which of these entries BuildKit can see. Non-regular, non-directory, non-symlink entries make the field `unknown`. |
| `named_contexts` | K | For each `--build-context`/`additional_contexts` entry: its name plus the same inventory digest (for example intelligence's `contracts`) |
| `args_digest` | K | Sorted `--build-arg` and Compose `build.args` |
| `daemon` | | `docker version` server version and BuildKit driver: job-level provenance, constant within a job |
| `progress` | | `parsed` or `unknown` (§3.8.2) |
| `steps_executed`, `steps_cached` | | From parsed progress only |
| `result`, `image_id` | | Exit status; `--iidfile` for explicit builds, `docker image inspect` of `<project>-<service>` for Compose, `cacheonly` for prebuilds |

**Soundness and limits:**

- **Equal keys imply equal BuildKit inputs.** Equal Dockerfile, target, frontend, resolved bases, platform, ignore file, complete context inventories and arguments determine identical inputs.
- **Unequal keys can still be BuildKit-equivalent**, when they differ only in ignored files (for example the B-BUILD-CONTEXT canary). §3.8.3's near-duplicate rule keeps that from hiding a duplicate execution.
- **No new pull.** Base identity is taken as resolved by the build itself; tool images used only at runtime (gitleaks, yq and the like) are not builds and remain in `images.txt` and job attribution.

#### 3.8.2 Progress and confirmed cache hits (C4.2)

**Capture.**

- **Explicit harness builds** (`docker build`/`buildx build`) go through the ledger wrapper, which replaces `-q` with `--progress=plain --iidfile <file>` and captures the progress.
- **Compose** and **documentation-block** builds run with `BUILDKIT_PROGRESS=plain` exported by the harness, which leaves the README command text unchanged. Their logs are already captured per call.
- **S1 verification** (Docker 29.8.2 in CI; the local engine also recorded) covers three things: Compose honours the variable; resolved `FROM …@sha256:` lines appear; per-service vertex prefixes identify each service. A failure of any of them is a stop condition for the affected build path, never a silent exclusion.

**Parse.** An invocation is `progress=parsed` only if all of the following hold:

1. The expected plain-format header is present (the `#N [internal] load build definition` vertex). No TTY-style (`=> [`) or quiet output appears.
2. Every step vertex `#N [<svc> <stage> i/n] (FROM|RUN|COPY|ADD|…)` has exactly one terminal line (`#N CACHED`, `#N DONE <t>s` or `#N ERROR`), and for every stage printed every index 1..n appears.
3. Every external base of the target has a resolved identity (§3.8.1).
4. The export/naming, or cache-only completion, line is present, and it agrees with the exit status.

The status is then one of three:

| Status | Condition |
|---|---|
| **Confirmed cache hit** | `parsed`, and every RUN/COPY/ADD vertex is `CACHED` (`steps_executed = 0`) |
| **Executing** | `parsed`, and at least one such vertex is `DONE` |
| **`unknown`** | Anything else: absent or empty progress, an unexpected or TTY format, a missing or duplicate terminal line, an unknown status token, a missing stage index, a nonzero exit without `ERROR`, or any unknown K field |

An `unknown` row is never counted as a cache hit, never dropped and never excluded.

#### 3.8.3 Accounting and the AC5 status

**Scope (denominator).** Every build performed in the job:

- every explicit harness build (7 call sites at `570b40c`), and every cache-only prebuild;
- every Compose call that can build:
  - `build` and `--build`, with the targeted build services from `compose config`;
  - `up` or `run` without `--no-build`, for each targeted build service whose `<project>-<service>` image was absent before the call;
- every documentation-block build (`run_docs_block`): each `<project>-<service>` image that appeared or changed ID during the block, plus each service whose build section was printed, and the block's targeted build services whenever its text contains `--build` or ` build `;
- **one `bootstrap` row** for the hosted adapter's pre-harness context image (`infra/docker/ci-context.Dockerfile`, stdin, no context, both bases pinned by digest in the Dockerfile).
  - Its build is not run by the harness and uses `-q`, so step counts are not observable. validate.sh records the Dockerfile SHA-256, the pinned base digests, the platform and the image ID (`docker image inspect platform-ci-context:$CI_JOB_ID`).
  - Its once-per-job property is structural: a single call site in hosted-context.sh, pinned by the existing G controls.
  - It counts in the denominator as `bootstrap`, not as a cache-hit claim. Its acceptability is §12 Q1; if refused, the alternative is a progress capture in hosted-context.sh with its G-control update.

**Excluded by definition, not for being unknown.** Builds inside the controls' scratch trees — V4's nested harness, the VB and VC fixtures — are the controls' own scenarios. They deliberately construct duplicates and changes, and each control asserts its own outcome.

**`ledger_check`** is one deterministic function in `infra/validation/lib.sh`. run.sh runs it after all builds, before `final_checks`, writing `BUILD-LEDGER INFO` and a staged `ledger-summary.tsv`. The same function re-runs on published artifacts through `pipeline-union.sh --ledger <job-dir>` (one mode of the C2 checker; no new tool). It computes:

| Measure | Definition |
|---|---|
| `expected` | The denominator above |
| `rows` | The rows present; each expected invocation must have exactly one |
| `unknown` | Rows with `progress=unknown`, any unknown K field, or no matching row |
| `duplicates` | Keys with more than one executing row, other than one authoritative plus `named:<exception>` rows from the reviewed list |
| `near_duplicates` | Executing rows sharing every K field except `context_inventory`/`named_contexts`. They are listed with their differing paths (a cheap diff of the two inventories) and must each carry a named exception |
| `ac5_status` | **MET** only when `unknown = 0`, `duplicates = 0`, every near-duplicate carries a named exception, and `rows == expected`; otherwise **UNMET** |

**Named exceptions** (reviewed list):

- `prebuild-fallback`: a prebuild that failed, then its authoritative build;
- B-BUILD-CONTEXT (canary context);
- B-WEB-ERROR-OUTPUT (probe-route build);
- the web-browser fixture stage;
- B-TOGGLE-OFF (variant render).

Anything else is unexcepted.

#### 3.8.4 Gating and acceptance (C4.3)

- **Merge gating is unchanged.** `BUILD-LEDGER INFO` is informational and adds no permanent efficiency merge gate. A broken ledger still cannot pass silently: `ledger_check` failing to run or to write the summary records `BUILD-LEDGER FAIL` (an existing non-required-row pattern), which fails the job.
- **G051 acceptance is blocked by ledger evidence.** The required AC5 evidence runs are the hosted cold jobs: H2 fast, H3 fast, H4 fast and H4 gate. For each, `ledger_check` (re-run by `pipeline-union.sh --ledger` on the published artifacts) must report **`ac5_status = MET`**.
  - An unexcepted duplicate or near-duplicate executing build, or any incomplete evidence (`unknown > 0`, a missing row or a missing summary), **blocks G051 acceptance**. The Result must not claim readiness or AC5.
  - The affected work stops and is corrected within B5 (§11). For H4, which is post-merge, that means a correction MR and a new qualifying staging run.
- **Local runs** (the probe clone and the final qualification) also run `ledger_check`. They must be parseable, and an `unknown` there is a defect to fix before S7. They are warm and are not AC5 evidence.

#### 3.8.5 Validity of reuse

**BuildKit's content-addressed cache keys decide reuse.** The ledger only proves what happened. Accepted controls already prove the inputs it relies on:

| Control | Proves |
|---|---|
| B-DEPS | pinned base and tool tags and locks; `images.txt` is cross-checked in §3.8.1 |
| B-DOCKERIGNORE | every Dockerfile-specific ignore file excludes `.env` and `.env.*` |
| B-BUILD-CONTEXT | no environment file in core, migrate or web-browser stages |
| B-HADOLINT | Dockerfiles |
| B-SOURCE-HASH | disposable copies are byte-identical to the implementation |
| B-RENAME | no rebuild for `.env`-only changes |

Source and lock changes are covered by those controls together with VC2. The new behavior — the ledger key, the progress classification, prebuild-then-consumer and the accounting — is covered by the focused VC1–VC8 controls (§6.1).

### 3.9 Finalizer and reporting

- **Profile tokens:** `affected` is added to validate.sh, junit.sh (+ `remainder`), runner-preflight and the summary validation. Affected capacity and budget equal integration's; within the fast job the cap is 1200 s and the minimum 600 s.
- **State import** adds `AREAS`, `PIPELINE_SUITES`, `EXECUTED_SUITES`, `COVERAGE` and `SELECTION_DIGEST`, each with fixed regular expressions.
- **Completeness** uses the executed suites. Contracts reports are required if and only if `contracts` executed. Browser reports are required per executed `root`/`off`/`rename`/`feature-web`.
- **Staged allow-list** adds `builds.tsv`, `ci-controls-timings.tsv` and `browser-feature-results.json` (attachments removed).
- **Unchanged:** aggregation precedence, postcheck, scans before publication, ownership cleanup and identity-staging forwarding.

## 4. Files, modules and interfaces

| File | Change |
|---|---|
| `infra/ci/select-profile.sh` | Areas, allow-list, BASE/HEAD closure, ambiguity, absent/removed rules, suites, `selection_digest`, local `--base`/`--compare` |
| `infra/ci/profile-paths.tsv` | Area grammar comment; R3 widening rows |
| `infra/ci/lib.sh` | `affected`; split suite lists; capacity and budget; partition helpers |
| `infra/ci/validate.sh` | Local options; state; L1 partition; completeness; summary fields; staged allow-list |
| `infra/ci/junit.sh`, `infra/ci/runner-preflight.sh` | The `affected`/`remainder` tokens |
| `infra/ci/pipeline-union.sh` (**new**, small) | The C2 union evidence check (§3.7) |
| `infra/ci/validate-controls.sh` | §6.1 controls; per-control timing; the stub accepts the new suites |
| `infra/validation/run.sh` | Suite split and aliases; `unit-*`; `feature-*`; `--areas`; L2; L5 via the bg helpers; EXIT trap; L6 injection; B-CI-CONFIG expectations; ledger |
| `infra/validation/lib.sh` | `bg_*` helpers; the `record` `TIMINGS` variable; feature run helper; `PORTS_FEAT`; ledger helpers |
| `infra/validation/required-checks.tsv` | Relabels and 5 rows |
| `web/playwright.config.ts` | `FEATURE_E2E` mode; inherited `testIgnore`; `PLAYWRIGHT_WORKERS` (default 1) |
| `docs/foundation/OPERATIONS.md`, `README.md` | `platform:ci-affected` |
| `.gitlab-ci.yml` | No rule, job, variable or timeout change; at most a comment |

- **Database, schema, API, MCP, application, vendor:** none.
- **Interfaces:** the validate.sh local CLI; the selector outputs; run.sh suites and `--areas`; the map grammar; the summary fields; the JUnit tokens; the Playwright environment variables; `pipeline-union.sh A_DIR B_DIR`. The CI job CLI, workflow, artifacts, retention and access are unchanged.
- **Proposed scope:** the files above plus `management/v0/results/G051-.*`.

## 5. Cold timing targets (R5)

**Definitions** (also in OPERATIONS):

- **Required jobs:** jobs that run automatically in the context. A manual MR gate is not one.
- **Wall time:** max(`finished_at`) − min(`started_at`).
- **Queue:** the sum of `queued_duration`, reported and excluded.
- **Compute:** the sum of `duration` × 2.
- **Cold:** a fresh hosted VM and DinD with no project cache.
- **Qualifying runs:** successful, non-retried runs only. Canceled, failed, timed-out or incomplete runs are history.
- **Measurement:** read-only `glab api` GETs.

| Class | Run | Projection (estimate) | Target |
|---|---|---|---|
| UI (AC6) | H2: MR fast job, affected web area | ≈3.4 min / 6.8 compute (with L4) | 5 / 10 |
| Backend (AC7) | H3: MR fast job, affected core area, PostgreSQL round trip | ≈3.1 min / 6.2 compute | 15 / 30 |
| Full (AC8) | H4: staging push, fast + gate | ≈40–41 min (§3.5) | 45 / 90 |

## 6. Validation and evidence plan

**Evidence location:** `<home>/.local/state/platform-evidence/g051-exec-<base7>/` (runs.tsv, per-run directories, `bin/`). Every local harness or controls run takes `local-harness.lock` and uses the Codex `run-approved.sh` binding. No repository edit is made during a run.

### 6.1 Controls

**F, selection** (table-driven `F_ROWS` plus fixture maps with area rows; each asserts `PIPELINE`/`EXECUTED`/reason, **`EXECUTED_SUITES`, `AREAS` and the exact consumers**, or the full fallback).

r1 set (kept):

| Category | Controls |
|---|---|
| Area cases | page+spec, test-only, fixture-only, core IT, core unit-only, intelligence, cross-stack, two web areas |
| Mixed | area + docs, area + agent, area + runtime, area + lock |
| Unregistered | core package, feature test |
| Map refusals | bad root, bad name, overlap |
| Widening | ten R3 widening cases |
| Contexts | push-staging, fast-job, gate, web/api/schedule/merged (cannot narrow) |
| Base and local | base-missing, local base/no-base/compare/compare-unknown, CI compare refused |
| Gate | full and integration remainder |

`F-review-web` → full.

**C1 additions:**

| Control | Fixture | Expected |
|---|---|---|
| `F-dep-go-subpkg` | foundation file imports `core/internal/a/sub` | full |
| `F-dep-go-subpkg-area` | area b imports `a/sub` | {a, b} |
| `F-dep-go-test-only` | a foundation `_test.go` references the area | full |
| `F-dep-go-testdata-path` | a foundation file names `core/internal/a/testdata/x` | full |
| `F-dep-web-alias` | `@/app/(features)/a/x` imported from `web/lib/` | full |
| `F-dep-web-relative` | `web/app/page.tsx` imports `./(features)/a/w` | integration |
| `F-dep-web-relative-area` | area b imports `../a/w` | {a, b} |
| `F-dep-web-css` | `url(../(features)/a/x.svg)` from runtime CSS | integration |
| `F-dep-web-dynamic` | non-literal `import(x)` outside the area | full |
| `F-dep-web-escape` | a specifier escaping `web/` | full |
| `F-dep-py-submodule` | `from a.sub import x` in `main.py` | integration |
| `F-dep-py-relative` | a relative import resolving into the area | its class |
| `F-dep-py-relative-unresolved` | an unresolvable relative import | full |
| `F-dep-py-dynamic` | `importlib.import_module(name)` outside | full |
| `F-dep-removed-edge` | an area file deleted while area b imported it at BASE | {a, b} |
| `F-dep-removed-foundation-consumer` | same, with a foundation importer at BASE | full |
| `F-dep-renamed-area` | rename a→c with importer b at BASE and HEAD | {a, b, c} |
| `F-fixture-other-area` | area a's fixture referenced by area b's spec | {a, b} |
| `F-fixture-inherited` | a fixture referenced by an inherited spec | full |
| `F-feature-test-only` | spec-only change | {a}, `feature-web` |
| `F-area-unclassified-script` | `core/internal/a/run.sh` | full |
| `F-area-unclassified-gomod` | a `go.mod` inside an area | full |
| `F-area-unclassified-config` | `web/app/(features)/a/next.config.ts` | full |
| `F-area-unclassified-pyproject` | a `pyproject.toml` inside an area | full |
| `F-area-symlink` | a symlink inside an area | full |
| `F-area-route-handler` | `route.ts` inside an area | integration |
| `F-closure-base-unreadable` | BASE unreadable | full |
| `F-registered-absent` | absent at BASE and HEAD | INFO, no obligation |
| `F-area-removed-clean` | all files deleted, no consumers | `lint-<kind>` + `unit-<kind>` |
| `F-area-removed-consumer` | all files deleted, foundation consumer at BASE | full |

**A, in-job completeness** (stub harness, real validate.sh):

| Control | Fault | Expected |
|---|---|---|
| VA1 | — (affected web) | 0; `fast:affected` |
| VA2 | B-FEATURE-WEB missing | 14 |
| VA3 | duplicate row | 14 |
| VA4 | FAIL row under harness exit 0 | 14 |
| VA4b | UNMET row | 14 |
| VA5 | a row outside the executed suites | 14 |
| VA6 | malformed `AREAS`/`EXECUTED_SUITES`/`COVERAGE` state | 16 |
| VA7 | — (gate remainder, `coverage=partial`, `gate:remainder`) | 0 |
| VA8 | a fast-suite row in the gate | 14 |
| VA9 | — (contracts reports conditional) | V13* pass |
| VA10 | missing `browser-feature-results.json` | 14 |
| VA11 | inconsistent partition state (remainder ≠ pipeline − fast) | 16 |

**A, union check** (C2; `pipeline-union.sh` on fixture artifacts):

| Control | Fault | Expected |
|---|---|---|
| VU1 | — (correct full split: 33 + 53 → 82, overlap = 4 `final`) | 0 |
| VU2 | missing fast partition | fail |
| VU3 | wrong fast partition (affected or a different suite set) | fail |
| VU4 | union gap (a remainder ID absent) | fail |
| VU5 | unexpected non-final overlap | fail |
| VU6 | candidate tree mismatch | fail |
| VU7 | selection or `pipeline_set_sha256` mismatch | fail |
| VU8 | a job phase ≠ 0 | fail |
| VU9 | — (correct integration split, 33 + 42 → 71) | 0 |

**D(J), reports:**

- **J7:** `fast:affected` counts equal the TSV.
- **J8:** `gate:remainder` (partial) counts equal the TSV; the suite name never states `full`.

**D, background supervision** (C3; drivers source the real `lib.sh`, and VB1/VB3b use the real run.sh in a scratch copy with a stub `validate-controls.sh`):

| Control | Scenario | Assertions |
|---|---|---|
| VB1 | The stub prints every PASS line, then KILLs the controls subshell | `BACKGROUND FAIL`; run.sh 1; validate.sh nonzero; summary harness=1; no successful publication; no live PID |
| VB2 | Foreground `exit 2` while a child holds a run-labelled busybox container | exit 2; child reaped before cleanup; container removed; inventory-leftover empty |
| VB3 | Group TERM while the child ignores TERM for 5 s | 143; reaped; resources removed |
| VB3b | validate.sh deadline (V18 budget pattern) with a holding stub | 124; no survivor; no successful-looking publication |
| VB4 | A failing fake prebuild, then a failing fake authoritative build | nonzero; `PREBUILD INFO` plus the owning FAIL; no live child |
| VB5 | Normal join, with a variant whose private rows hold a foreign ID | merged rows and `BACKGROUND INFO`; the foreign-ID variant gives `BACKGROUND FAIL` |

**E, build ledger** (C4, r3). These controls use Docker-real busybox builds with control labels and the real `ledger_check`; the parser-only cases use fixture progress files. Every control asserts four things:

- the exact rows and keys;
- `expected == rows` (complete invocation accounting);
- the resulting `ac5_status`;
- cleanup of the labelled images.

| Control | Scenario | Expected |
|---|---|---|
| VC1 | Identical input built twice (per-run nonce file) | same key; the first `parsed` and executing, the second `parsed` with executed 0 (a confirmed hit); MET |
| VC2 | A changed context file | a new key; COPY executed; MET |
| VC3 | Two executing invocations of one key, with and without `named:prebuild-fallback` | with the exception MET; without it `duplicates = 1` and **UNMET** |
| VC4 | Cache-only prebuild, then consumer | the consumer executed 0; no image or tag from the prebuild; one executing row; MET |
| VC5 | **Base identity changed:** the same Dockerfile text and arguments, with the control-labelled local base tag re-pointed between builds (`docker tag` of two busybox versions) | `bases` differ, so the keys differ; both rows accounted; MET. If the identity cannot be resolved, the row is `unknown` and **UNMET** (the refusal path is asserted, not skipped) |
| VC6 | **Platform changed:** a COPY-only Dockerfile (no RUN, so no emulation) built with the default platform and with `--platform linux/arm64` | `platform` and `bases` differ, so the keys differ; both accounted; MET. If the arm64 base is unavailable: `unknown`, UNMET |
| VC7 | **Effective context or ignore configuration changed:** (a) a Dockerfile-specific `.dockerignore` added; (b) both ignore files present; (c) only a file mode changed; (d) a symlink retargeted; (e) a regular file replaced by a symlink with the same content; (f) only an ignored file differs, second build a confirmed hit; (g) a near-duplicate executing pair with and without a named exception | (a, c, d, e) distinct keys; (b) the Dockerfile-specific file chosen and both recorded; (f) distinct keys, the second executed 0, no duplicate; (g) without the exception **UNMET**, with it MET |
| VC8 | **Unknown, unparseable or unaccounted:** parser fixtures for empty progress, TTY `=>` format, a missing terminal line, an unknown status token, a duplicate terminal line, a missing stage index i/n, a nonzero exit without `ERROR`, and a missing resolved base; one real Compose build with `BUILDKIT_PROGRESS=quiet`; one fixture where a new `<project>-<service>` image has no row | each case `progress=unknown` or a missing row, **kept in the denominator**, and **UNMET**; never counted as a cache hit |

**Existing controls** run unchanged (282 at `570b40c`), and `count == --list` must hold.

### 6.2 Local evidence (focused first; one final qualification)

1. **Focused development checks:**
   - `validate-controls.sh --group F`, `--only <new A/J/VU/VB/VC IDs>`, `--group A,D` after validate.sh changes;
   - `run-approved.sh … -- --base <BASE> --allow-file <scope> lint lint-base lint-web lint-core provenance`;
   - targeted retests per the control-group map.
2. **Probe clone** (outside the repository; a registration commit plus probe commits). `validate.sh --profile affected --base <registration>` runs for:
   - UI, core (PostgreSQL round trip) and intelligence probes;
   - failing spec → FAIL; page without spec → UNMET; IT failure → FAIL;
   - a feature page with a TypeScript error, where prebuild and authoritative build both fail → FAIL with `PREBUILD INFO`, no live child (VB4 real-Docker counterpart).

   For each, record the executed IDs, `cleanup-inventory.tsv` and `builds.tsv`, and confirm that identical-input invocations show `steps_executed = 0`.
3. **Focused concurrency and stability** (before rebase; not a full qualification):
   - one `run.sh controls root` pairing run under the wrapper, the hosted overlap pair, with L5 and L6;
   - two further browser-suite executions with `PLAYWRIGHT_WORKERS=2` against one started root stack (README start, the injected browser line ×2, stop). Three consecutive passes in total are required to keep L6. A flake disables L6 and is reported at the checkpoint.
4. **Final local qualification, once,** after fetching and rebasing onto `origin/staging`: `run-approved.sh … --report-dir <EV>/final all unit`, BASE-aware, all 81 full IDs plus B-UNIT-CHECKS, with L5/L6 on and suite timings and the ledger recorded. Then the probe-clone affected runs on the rebased candidate.
   - Step 3's evidence is reused only where its exact inputs are unchanged.
   - If the target advances again, `--compare` selects only the suites to rerun.
5. **Admission and lint.** The admission oracle (Ruby `File.fnmatch`, pinned container, evidence only, the G049 method) for the area and widened classes against the unchanged workflow. `glab ci lint` (read-only).

### 6.3 Hosted runs (named actions to be requested; not authorized by the review)

Before any action, a read-only quota check is made: `ci_quota_exceeded` is a resource stop.

| Run | Action | Evidence |
|---|---|---|
| H1 | Push the G051 branch and open the MR to `staging` | Automatic `fast` (selection full → 33) |
| H2, H3 | `git push -o ci.skip` of `g051-qual-base` (the rebased candidate plus a **registration-only** commit for `qualprobe` web and core areas), `g051-qual-ui` (+ page and spec) and `g051-qual-core` (+ a package with a unit test and a PostgreSQL-round-trip integration test). Two draft "DO NOT MERGE" MRs target `g051-qual-base`; each automatic fast job runs `affected`. | AC6/AC7. Afterwards: close the MRs, delete the 3 branches after verifying their tips, record the SHAs, and show that `origin/staging` has no `qualprobe` path. |
| H4 | After Codex review and the owner-authorized conditional merge: the automatic staging push (fast + gate) | AC8: timings, `pipeline-union.sh` on both jobs' artifacts, both ledgers (canceled does not qualify) |

**Not requested:** a manual MR gate, schedules, settings, protected-branch or identity actions, or extra full runs.

### 6.4 Documentation (AC10)

OPERATIONS.md gets:

- the profile and trigger matrix; area registration; the §3.3 dependency forms and fallbacks;
- the §3.1 obligation table;
- partitions, the union check and the release-reuse rule (§3.7);
- the definitions of §5;
- full, release and weekly use;
- the cache and ledger accounting (§3.8);
- background supervision semantics (§3.6);
- focused debugging: `--base`/`--compare` and the control-group map;
- limitations.

README gains `platform:ci-affected`.

## 7. Acceptance-criteria mapping

| AC | Evidence |
|---|---|
| AC1 | Pinned baseline, tree and inputs (this Plan, the refreshed inventory, Codex's binding); the G049/G050/G005 records; §2; only the §4 sites. |
| AC2 | Pre/post registry diff and §3.1. The final local full run and H4 cover 82 distinct obligations (81 full + B-UNIT-CHECKS): locally in one job, hosted as 33 + 53 rows with 4 documented `final` overlaps, verified by `pipeline-union.sh`. VA2–VA11, VU2–VU8 and V10/V11/V13* show missing, duplicate, corrupt, FAIL/UNMET, gap and overlap all failing. |
| AC3 | F area and C1 dependency controls (exact consumers and suites); probe-clone IDs, `cleanup-inventory.tsv` and `builds.tsv` (built and started services); H2/H3. |
| AC4 | F widening, deletion/rename, mixed, unknown, ambiguous, base and context controls; the admission oracle; `glab ci lint`; the B-CI-CONFIG glob limit. |
| AC5 | Per-job `builds.tsv` with the complete §3.8.1 key: Dockerfile, target, frontend, resolved bases, platform, effective ignore file, context and named-context inventories, arguments. Step counts are confirmed from parsed progress. `ledger_check` must report **MET** (`unknown = 0`, `rows == expected`, no unexcepted duplicate or near-duplicate) for each required cold run: H2 fast, H3 fast, H4 fast and H4 gate, re-verified with `pipeline-union.sh --ledger`. UNMET in any of them blocks G051 acceptance (§3.8.4). Also: VC1–VC8; L1/L2 duplicate removal; cleanup inventories; variant scenarios intact. |
| AC6 | H2: changed paths, IDs, timestamps, durations, queue, factor 2, no project cache. |
| AC7 | H3: B-FEATURE-CORE PostgreSQL round trip, completeness, closure, timings. |
| AC8 | H4: union check, clean status, wall and compute time, the inherited scenarios present. |
| AC9 | VA/VU/J/VB controls; the full B-CI-CONTROLS run; JUnit per job equals its executed partition and marks partial coverage; identity-staging and C-SEC controls unchanged. |
| AC10 | OPERATIONS and README; the Result with commits, inputs, commands, jobs, artifacts, failures, exclusions, cleanup and B5 accounting. |

## 8. Sequence and budget

**No manual or operator feedback stage is required** by the Spec.

| Step | Content |
|---|---|
| S0 | Preconditions (§10); Codex binding; evidence directory; quota; lock. Refresh the baseline from pipeline 2931683534's terminal result. |
| S1 | Timing and ledger instrumentation (including verifying `BUILDKIT_PROGRESS` under Compose); suite split; registry; tokens and state; `bg_*` helpers and the EXIT trap; B-CI-CONFIG expectations. Focused lint/provenance and the A/J/VB/VC controls. |
| S2 | Selector and map (C1); F controls. |
| S3 | `unit-*`, `feature-*`, Playwright modes, prebuild; L1/L2/L5/L6 wiring; `pipeline-union.sh` and the VU controls; probe-clone runs (§6.2 step 2); focused concurrency and stability (§6.2 step 3). |
| S4 | OPERATIONS and README; self-review; the read-only reviewer. |
| S5 | **Checkpoint** (after cycle 3, or here if fewer cycles were used): ACs, measured versus projected times, convergence, the L6 outcome, the L3 decision. If a target is infeasible: **stop** with measurements and the smallest adjustment. |
| S6 | Fetch and rebase onto `origin/staging`, reconcile, and run the **single** final local qualification (§6.2 step 4); oracle and lint. |
| S7 | H1, H2/H3 with cleanup; Result; return to Codex. |
| S8 | After conditional integration: H4; append AC8 and the final status. |

**B5 feasibility after the corrections.** Planning consumes no cycle.

- The corrections add mostly deterministic code and controls: the closure rules, the union check, the supervision helpers and the ledger. They do not add new failure-prone hosted iterations.
- Meaningful cycles are forecast at **4**: S2 closure defects, S3 feature or concurrency issues, an S6 reconciliation or timing correction, and a reserve for H4.
- **B5 remains feasible but has little slack.**
- The main risks are L5/L6 stability and the H4 timing margin (≈4 min). At the checkpoint, non-convergence or a projection above 45 min means stop rather than extend. A split is decided by Codex.

## 9. Delegation

- **Main agent:** implements, integrates and validates everything.
- **Optional implementation subagent:** owns only `infra/ci/validate-controls.sh` (the F C1, VA, VU, J, VB and VC controls) after S1 freezes the interfaces. The main agent reviews and runs the controls.
- **One read-only coverage and security reviewer at S4:** C1 closure and allow-list, C2 partitions and union, C3 supervision, C4 ledger, the widening and L2 mapping. It reports only.
- All cycles count against B5. Acceptance and integration stay with Codex.

## 10. G005 reconciliation precondition

G005 is integrated at `570b40c` (MR merge of `g005-organization-group-memberships`; its Result `f862f18` states "hosted evidence pending"). Reconciled facts:

- `required-checks.tsv` has 83 rows (`18242e72…`): B-TENANCY-HTTP in `root`; B-TENANCY-CATALOG/MIGRATION/IT in `migrations`. The profiles are fast 33, integration 71, full 81, staging 5.
- `profile-paths.tsv` (`6c2c2e0a…`) has `foundation dir core/internal/tenancy/`.
- `validate-controls.sh` (`0fb81a84…`) has the `F-tenancy` row (full/full/foundation); with G005, 282 controls.
- `run.sh` (`5ba83ee4…`) has the tenancy wiring: migrations at version 3, `tenancy_http`, `tenancy_migrations`, `tenancy_model`, the B-MCP-SOURCE tenancy assertion and APPLICABLE.

`core/internal/tenancy/` stays foundation and can never be an area.

**Execution prerequisites:**

- G005's formal Codex ACCEPT;
- staging pipeline 2931683534 reaching a qualifying terminal status (currently running);
- a fresh Codex binding at the then-current `origin/staging`.

Any later G005 correction is reconciled the same way. G051 never touches G005's checkout, branch or resources.

## 11. Stop and escalation conditions

Stop and return to Codex on any of:

- input or predecessor drift, checkout overlap, or unavailable authorization or quota;
- an unaccountable required check, or an unreviewed exclusion;
- an unresolvable dependency (use full and report);
- any change to the runner, privilege, vendor, credential, registry, security model, authority or application;
- a workflow or settings change becoming necessary;
- a target infeasible after the checkpoint (measurements plus the smallest adjustment; no target or check change);
- an L5 or L6 stability defect that the contingency or disabling does not resolve within targets;
- a background process that cannot be proven reaped;
- unsafe cache or resource behavior;
- budget exhaustion or thrashing.

## 12. Remaining risks and questions

- **Q1 (genuine question, r3).** §3.8.3 accounts for the hosted adapter's single pre-harness context-image build as a `bootstrap` row: identity from Docker metadata, once per job by structure, step counts not observable because the build uses `-q`. It stays in the denominator and is not a cache-hit claim.
  - Is that acceptable for AC5?
  - The alternative is to capture BuildKit progress in `infra/ci/hosted-context.sh`. That would change the security adapter's build invocation and the G controls that pin it, so the change is not planned unless Codex requires it.
- **R1 (resolved by the r2 review).** `BUILD-LEDGER INFO` stays non-gating for merges. An unexcepted duplicate or near-duplicate executing build, or incomplete ledger evidence, in any required AC5 run **blocks G051 acceptance** (§3.8.4). No permanent efficiency merge gate is added.
- **R6 (r3).** The ledger depends on BuildKit and Compose plain-progress formats: the header, `FROM …@sha256:` resolution lines, and per-service vertex prefixes.
  - These are verified in S1 on the pinned CI engine (Docker 29.8.2) and recorded for the local engine.
  - A format change or an unverifiable path yields `unknown`, which means AC5 **UNMET**: a stop and correction within B5, never an exclusion.
  - Post-merge, an UNMET H4 ledger would require a correction MR and a new staging run before acceptance.
- **R2.** The full target margin is about 4 min, and it relies on L6 and on the refreshed post-G005 gate baseline (pipeline 2931683534, running). If the refreshed baseline or the S3 stability results move the projection above 45 min, S5 stops with measurements.
- **R3.** Narrower affected selection raises the stakes of the existing owner-setting risk: pipeline-variable override of the comparison base (OPERATIONS, *Owner settings*). Base checks are unchanged; the risk is documented.
- **R4.** No area exists today, so current changes keep integration or widen to full. The benefit appears once areas are registered (product work after G021). H2/H3 demonstrate the mechanism.
- **R5.** Hosted variance is about ±5%. A single cold run per class is an acceptance demonstration, not a percentile claim.
