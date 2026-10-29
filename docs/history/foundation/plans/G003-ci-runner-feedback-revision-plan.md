> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G003 — CI runner and feedback revision: implementation plan

Goal: [G003](../goals/G003-foundation-observability-ci.md) · Revision Spec: [G003 CI runner and feedback revision](../specs/G003-ci-runner-feedback-revision-spec.md) (§3.1, §11: F1–F6, AC7a–AC7g) · Original Spec: [G003 Work Spec](../specs/G003-foundation-observability-ci-spec.md) · Historical Plan: [r4](G003-foundation-observability-ci-plan.md) (unchanged; referenced, never edited) · Decision prerequisite: [ADR-002](../../../adr/ADR-002-gitlab-hosted-ci-runner.md) (**PROPOSED**)
Owner: Claude Code · Reviewer: Codex · Proposed budget: **B3, not active**

**Status: PLAN ONLY — not executable.** Execution needs, in order: (1) owner acceptance of ADR-002; (2) Codex Plan ACCEPT; (3) Codex's fresh execution binding (supplied separately after review and reconciliation); (4) named authorization of each hosted action in §6.3; (5) a compute-quota decision (§10.2). This Plan changes no Goal, Spec, Plan r4, Result, PROGRESS, snapshot or source file.

## Revision history

- **r1 — 2026-10-06.** First linked revision plan for revision Spec §3.1/§11. Planning used read-only file inspection, read-only `glab api` GETs (§0.3), public registry metadata GETs (§0.4) and lightweight reasoning about syntax and rules. No implementation, install, Docker command, build, full suite, CI parity, hosted probe, pipeline, push or settings write. Zero cycles consumed. Archived byte-identical as G003-ci-runner-feedback-revision-plan-r1.md (not included).
- **r1 → Codex: CORRECTION LOOP (2026-10-06).** Reviewed r1 SHA-256 `37b6a2d17ba1d88d090c9c309b942675da403634c6e5e77419a711cb7fe1bcd3`. Three findings:
  1. **Selection coverage (Spec §4, AC7b).** Mapping docs and governance to fast could omit security or governing test-contract changes, and merged-result classification ignored target-side changes.
  2. **H-INT construction (AC7d).** This revision changes only foundation and docs paths, so a "foundation subset + runtime remainder" pair could produce an empty diff and select fast.
  3. **P0 proportionality.** The pre-implementation control/root/browser/contracts slice duplicated H1.
- **r2 — 2026-10-06 (this file).** Owner asked for minimal churn, so only the affected sections changed.
  1. Finding 1:
     - Architecture, security, decisions/ADRs, `AGENTS.md`, `CLAUDE.md`, governing Goal/Spec/Plan/Result/allow-list, roadmap and validation-policy files now select **full**. Only `PROGRESS.md`, `BACKLOG.md` and `docs/history/` stay fast.
     - Merged-result pipelines now select **full**.
     - Exact scenarios were added to the selector-control table.
     - Sections changed: §2.3, §2.4, §2.5, §6.1, §10.2, §12 Q6.
  2. Finding 2:
     - H-INT is now a disposable base equal to H1 plus one named runtime-file perturbation that is never executed. Its head restores H1 byte-exactly.
     - Required evidence: tree equality, a nonempty runtime-only diff, and selected = integration.
     - Sections changed: §6.3, §7 AC7d, §12 Q11.
  3. Finding 3:
     - P0-lite is the qualification probe. It resolves the executor and privilege model, the TLS daemon, shared checkout/temp paths, loopback/non-loopback access, tool versions, capacity and timeouts, and credential isolation.
     - Group E and the full root/browser/contracts application qualification move to H1.
     - A probe failure remains a stop condition.
     - Sections changed: §6.2 L2, §6.3, §7 F3/F4, §8 R1, §10.2, §11, §12 Q10.
  4. Unchanged: everything else, including the B3 forecast, medium default, full assertions, measured-target dispositions, and the final hosted pass/failure/profile evidence. ADR-002 remains PROPOSED. Planning only, with no execution or tests.

## 0. Pinned baselines and verified facts

### 0.1 Planning baseline (main checkout, committed `staging`)

| Item | Value |
|---|---|
| Commit / tree | `329ba4514e926305045cfef383693a433e49b7ae` / `54314a61f6175f3a5f90fc9e719e1cf269160f0d` (`origin/staging` is the same commit) |
| Revision Spec | `150c767bca84c4d95eaf82c26f50ab86ebf7f39dbdf91bd10205ce1b3db33a97` |
| ADR-002 (PROPOSED) | `1f51b505184a7636586f096e1899f2ebfd33b9e407f5f98ff9f234d4fe7df035` |
| Goal / original Spec | `032cc23bbd1aaeba6f294ad4a6a312bc06383c2c10c2591e3bc1793b002501f4` / `b85a88f6e23dc1aea85b08c7d88622fc75d7e681c60f461798efdd5360d2de6b` |
| Plan r4 (historical) | `10a827884bb811f808fc82bc230df9d12b95e237b84d93a36f3067348d00ccb4` |
| PROGRESS / ARCHITECTURE / DECISIONS | `303e84bd…dbed87` / `a9fd1bde…22d0` / `acbb610d…0149` |
| AGENTS / CLAUDE / BUDGET_POLICY | `d75cf2a4…16e4` / `8a37ae48…1267` / `b183179b…f437` |
| `docs/SECURITY.md` | empty (0 bytes, SHA-256 `e3b0c442…b855`) in both trees. Security requirements come from ARCHITECTURE.md, the original Spec's constraints and revision Spec §5. |

### 0.2 Execution source (inspected read-only; never written)

| Item | Value |
|---|---|
| Checkout | `.claude/worktrees/g047-exec`, HEAD `cbffd6da5d9936c90fdfd9e51d84d36c0f17ad81` (not on `origin`) |
| NUL-safe status digest | `e0911d2f6a937efafb81fac61d52e728d9f373a5a2d9e9eb760d05703f758c6b`, 192 entries. Command: `GIT_OPTIONAL_LOCKS=0 git -c core.fsmonitor=false status --porcelain=v1 -z --untracked-files=all \| sha256sum`. This equals the reviewed-submission digest in Result §14.1. |
| Result | `b91aa8764a8641852ed3ec277215def3c86e0bd71ebb8aef91552102e1091000` (§14 = correction pass 1) |
| Retained evidence reused | `~/.local/state/platform-evidence/g003-69407d/g003-exec/`: `c1-all` (run `3c8cd5`: 69 PASS, per-check timestamps used in §2.1), `c1-s8/validate.out` (local parity, exit 0), `runs.tsv`, `cycle-log.md`; `../batch-19f945a/g003-runner-facts.md` (its historical "no API access" finding is superseded by §0.3) |

Pre-change SHA-256 of every file this Plan changes. The final binding re-pins them.

| File | SHA-256 |
|---|---|
| `.gitlab-ci.yml` | `a7fa0214f7b80e2e25297be1c6d4557888a155bc54464053a258c794c630aa2c` |
| `infra/ci/validate.sh` | `75c988c082d3391067b126f0cc9ba5d993b7192a982774aa54ceacec203ee223` |
| `infra/ci/lib.sh` | `52b6861712712fe1da5c9affe80b8668be540fc7305f8a0d1e263661c6df5dc8` |
| `infra/ci/runner-preflight.sh` | `eeaa2d17278404a612703b8cdeaa28e85f410ac7c62ccbc036e6918671b4b19d` |
| `infra/ci/validate-controls.sh` | `c1b88a850b451b8c0cfdcad8ef632b7fbd079fa526d7ac333d6162c715a09083` |
| `infra/validation/run.sh` | `7f42828a85f0b80256a2b029c5445d647f284d081733bdfecc6fee4bdebb43e8` |
| `infra/validation/lib.sh` | `abe240dd5a0cc651e41d967a3666c7f1ae537f16d68cb1f7d8487ea02942bc22` |
| `infra/validation/required-checks.tsv` | `60df2fec1c18c505f1f4406a51b5da1dc70656af46e1bf13a21acbab5945b36a` |
| `README.md` | `11c445b6fe7774761c7ca08343adfbb3ef4500112d34763af707d43b2daa5ca6` |
| `docs/foundation/OPERATIONS.md` | `01e121cf53095179a1513832557ed97cf937b7f5647a2058e3e31cd2a1eef21e` |
| `management/v0/results/G003-scope-allowlist.txt` (append only) | `4d802263f127cc00959942e73628f1ad33df58dd0c8ffe8746cf45ac5aa78fe3` |
| `management/v0/results/G003-foundation-observability-ci-result.md` (append §15 only) | `b91aa8764a8641852ed3ec277215def3c86e0bd71ebb8aef91552102e1091000` |

These are relied on and stay unchanged: `infra/ci/cleanup-owned.sh` `2d5f3d64…5304`, `infra/ci/unchanged.allow` `87a28951…938b`, `infra/validation/run-approved.sh` `620a3e84…c695`.

### 0.3 Read-only GitLab facts (`glab api` GET only, as `<owner-account>`; no token printed or stored)

| Area | Verified value (2026-10-06) |
|---|---|
| Identity | user `<owner-account>` (id <gitlab-user-id>), active, 2FA on. Group access level 50 (Owner) through `<namespace>`. |
| Project | <gitlab-project-id> `<namespace>/<project>`; private; default branch `main`; created 2026-10-04. The namespace plan is **free** (top-level `<namespace>`, 14 projects). |
| Branches | `main` `99e2c794` is protected (push and merge: Maintainers; no force push). `staging` `329ba451` is **unprotected**. |
| Activity | 0 pipelines, 0 schedules, 0 open MRs. |
| CI settings | shared runners enabled; `build_timeout` 3600 s; `ci_default_git_depth` 20; `build_git_strategy` fetch; `auto_cancel_pending_pipelines` enabled; `public_jobs` true; `keep_latest_artifact` true; `restrict_user_defined_variables` false; `ci_pipeline_variables_minimum_override_role` developer; `protect_merge_request_pipelines` true; `ci_allow_fork_pipelines_to_run_in_parent_project` true; `ci_job_token_scope_enabled` false. |
| Merge settings | `merge_method` merge. **`only_allow_merge_if_pipeline_succeeds` false:** pipelines do not yet block merges. The merged-results and merge-train fields are absent (Free tier). |
| Medium runner | id 55314121, "k8s.saas-linux-medium-amd64.runners-manager.gitlab.com/default"; instance type; online; `tag_list` [`saas-linux-medium-amd64`]; `run_untagged` false; `maximum_timeout` 10800; `not_protected`; Runner `19.5.0~pre` (`9e463190`); linux/amd64. |
| Small runner | id 54907241, tags [`saas-linux-small-amd64`, `gitlab--duo`]. **`run_untagged` true**, so every job must carry the medium tag. |
| Not inspected | CI/CD variables (the endpoint returns values); compute usage and quota (GraphQL POST or UI only); maximum artifact size (admin only). |

**Unproven until P0 (§6.3):**
- the actual executor (the "k8s." prefix names the manager, not the executor);
- whether job containers are privileged;
- the DinD filesystem and network-namespace model, and whether `/certs/client` is shared;
- whether each job gets a fresh VM;
- usable disk after setup;
- Docker Hub pull limits on the runner's egress;
- whether the job token persists in `.git/config`;
- the effective `CI_JOB_TIMEOUT` for `timeout: 3h` against the project's 3600 s;
- `after_script` behavior on timeout;
- the remaining quota.

### 0.4 Image pins (public registry metadata GET, 2026-10-06; re-verified by pull-by-digest in P0)

| Role | Pin |
|---|---|
| Job image (outer shell) and context CLI stage | `docker:29.8.2-cli@sha256:b1805116a6a86cc591b5d5f60a910a0715cdcc9d18d866ad68b1457ead25c35c` |
| DinD service | `docker:29.8.2-dind@sha256:7dcdfc4a20246236f558175182ccace1eb15a41bd3eb119dd2284f393498b7c1` |
| Context base (bash, git, GNU tools, procps; no package-manager step) | `buildpack-deps:trixie-scm@sha256:ec9476a31fe1d3d74057675bc0bc1672b21ce727315f42ca241643a1c6050f25` |
| Named fallback (only if P0 shows an Engine 29 incompatibility) | `docker:28.5.2-cli@sha256:625d9431a9f54c5a2bc90f24f0e1c3d55b1349fd857dd85035f98c2c9acbdd4d`, `docker:28.5.2-dind@sha256:2a232a42256f70d78e3cc5d2b5d6b3276710a0de0596c145f627ecfae90282ac` |

All existing harness image pins are unchanged (Plan r4 §13).

## 1. Approach (summary)

1. **Profiles are fixed selections of existing checks:**
   - **fast** = suites `lint provenance contracts unit`;
   - **integration** = fast plus `root ops rename off`;
   - **full** = `run.sh … all`, exactly as today.

   Two suites are split, with the same check functions and IDs: `static` → `lint` + `controls`, and `variants` → `ops` + `rename` + `off`. The aliases `static` and `variants` keep their old meaning. One stack-free unit suite is added (§2.1).
2. **Pipeline topology.** There are two gating jobs on `saas-linux-medium-amd64`:
   - `fast` runs in every supported pipeline;
   - `gate` (`needs: [fast]`) runs in merge-request pipelines, `staging`/default pushes, schedules and web pipelines. A small in-repo selector deterministically classifies the content. The gate then executes integration or full, or records that no higher profile is required.

   GitLab rules key only on the pipeline source, never on paths (§2.3–§2.5).
3. **Completeness is profile-aware in the existing entry point.** `validate.sh --profile` fixes the suites and the required ID and report sets from reviewed repository data, not caller data. It rejects rows that fall outside the profile. Full keeps today's gate byte for byte in coverage (§2.6).
4. **Hosted DinD arrangement (F3).** The harness runs, unchanged, inside a pinned, unprivileged "context" container that shares the job-private DinD daemon's network namespace. The checkout's parent build directory is mounted at the identical path. The loopback, default/variant-port and bind-path assertions therefore stay literally true (§3).
5. **Measured and proportionate evidence.** One bounded pre-implementation probe (P0) runs first. Local validation is focused. One final hosted full run (H1) follows, plus one integration run on identical content (H-INT) and one failing real-gate run (H2). Further full runs require a recorded reason (§6).

## 2. Profiles, inventory and selection (F1, F2, F5; AC7a, AC7b)

### 2.1 Suite-to-check inventory and profile membership

Local warm timings are from the retained `c1-all` harness log (run `3c8cd5`, 47m46s). Hosted runs start cold; their timings are forecasts until measured.

| Suite (new name ← old) | Required IDs | Local warm | fast | integration | full |
|---|---|---|---|---|---|
| `lint` ← `static` | B-SHELL, B-HADOLINT, B-COMPOSE, B-ENV-EXAMPLE, B-LITERALS, B-NAMING, B-DEPS, B-WEB-BOUNDARY, B-WEB-TOKENS, B-WEB-COMPONENTS, B-WEB-DESIGN, B-MCP-SOURCE, B-MIGRATE-SET, B-DOCKERIGNORE, B-CI-CONFIG (15) | 1m00s | ✓ | ✓ | ✓ |
| `controls` ← `static` | B-DEPS-CONTROLS (C1–C12, C8b, C8c), B-WRAPPER-CONTROLS (W1–W9), B-CI-CONTROLS (89+ controls) (3) | **16m11s** (B-CI-CONTROLS alone 15m22s) | – | – | ✓ |
| `provenance` | B-RENDER-ID, B-RENDER-PLAN, B-RENDER-REPEAT, B-PROTECTED, B-SKILL-UNCHANGED (5) | 9s | ✓ | ✓ | ✓ |
| `contracts` | B-CONTRACT-TOOLS, -GO, -PY, -AGREE, -PEER, -NEGATIVE (6) | 54s | ✓ | ✓ | ✓ |
| `unit` (new; stack-free) | **B-UNIT-CHECKS** (1) | ≈33s (as B-CHECKS-ROOT) | ✓ | ✓ | – |
| `root` | B-CONFIGURE, B-DEFAULTS, B-PORTS-ROOT, B-BOUNDARIES-ROOT, B-BOUNDARY-CONTROLS, B-MCP-HTTP, B-WEB-HTTP, B-HEALTH-ROOT, B-BROWSER-ROOT, B-CHECKS-ROOT, B-CORE/WEB/INTEL-ERROR-OUTPUT, B-DOCS-EXEC-GS, B-BUILD-CONTEXT, B-DOCS-EXEC-ROOT (16) | 12m34s (browser 7m45s) | – | ✓ | ✓ |
| `migrations` | B-MIGRATE-CLI, -CATALOG, -ENV, -IT (4) | 51s | – | – | ✓ |
| `ops` ← `variants` | B-OBS-CONTEXT, -CHAIN, -TIMEOUTS, -LOGS, -PRIVACY, B-OUTAGE, B-PERSIST (7) | 7m12s | – | ✓ | ✓ |
| `rename` ← `variants` | B-SOURCE-HASH, B-PORTS-OVERRIDE, B-RENAME, B-BROWSER-RENAME, B-DBRENAME (5) | 1m18s | – | ✓ | ✓ |
| `off` ← `variants` | B-TOGGLE-OFF-ADDITIONS, B-OBS-OFF, B-BROWSER-OFF, B-TOGGLE-OFF (4) | 7m34s | – | ✓ | ✓ |
| `final` | B-DOCKER-ONLY, B-SECRET-CANARY, B-SCOPE, B-REQUIRED-CHECKS (4) | <1s | ✓ | ✓ | ✓ |
| **Required IDs / local warm** | | | **31 / ≈3 min** | **63 / ≈31 min** | **69 / ≈48 min** |

**Full stays the same.** `all` = `lint controls provenance contracts root migrations ops rename off`. It is the same set of check functions as today's `static provenance contracts root migrations variants`. Order changes only between suites (the three control matrices now run after the other static checks), and those checks are independent.

**B-UNIT-CHECKS** runs the README `platform:check` block extracted with `docs_block`, replacing exactly one `--env-file .env` per line with a generated `$WORK/unit.env`, in the way `static_compose` already does. It covers core-check, web-check, intelligence-check and secret-scan, with no stack and no root `.env`. Fast's existing unit tests and source-secret check therefore have no new assertion. In full, the same commands remain covered by B-CHECKS-ROOT, so full keeps exactly its 69 IDs (Q5). `required-checks.tsv` gains one row (`B-UNIT-CHECKS<TAB>unit`); the other 69 IDs keep their IDs and get their refined suite labels.

**Integration is deliberately conservative.** It runs all sides (root, ops, rename, off) for any runtime or UI change. This satisfies "core/intelligence interfaces test both sides" and "shared boundaries run all affected sides" without a new per-side engine. It will exceed the ≤20/≤30 min target; disposition is requested after measurement (Q3).

### 2.2 Required reports per profile (`do_complete`)

| Report | fast | integration | full |
|---|---|---|---|
| `results.tsv`: every profile ID exactly once and PASS; no FAIL or UNMET; **no row for a `required-checks.tsv` ID outside the profile** (new) | ✓ | ✓ | ✓ |
| `contracts-results-{go,python}.json`: the exact 120-case `(id, valid, class)` set (correction-pass F4 logic, unchanged) | ✓ | ✓ | ✓ |
| `schema-negative-{go,python}.json` | ✓ | ✓ | ✓ |
| `browser-{root,off,rename}-results.json`: expected > 0; 0 skipped, flaky or unexpected | – (absence allowed) | ✓ | ✓ |
| `inventory-leftover.txt` empty | ✓ | ✓ | ✓ |

The required ID set is `{id : suite(id) ∈ PROFILE_SUITES ∪ {final}}`, computed from the checked-out `required-checks.tsv` and the profile constants in `infra/ci/lib.sh`. `summary.json` records `profile`, `pipeline_profile`, `executed_profile`, `selection_reason`, `comparison_base`, `required_count` and `required_set_sha256`. A fast or integration summary can never satisfy full: full requires all 69 IDs and all three browser reports.

### 2.3 Pipeline topology and attribution contexts

| Context | Pipeline | `fast` job | `gate` job | Comparison base | Gate executes | `tested_content` |
|---|---|---|---|---|---|---|
| push, feature branch, no open MR | yes | fast | – | none (context) | – | `branch-push` |
| push, feature branch with an open MR | **suppressed** (workflow rule); the MR pipeline gates | – | – | – | – | – |
| push to `staging` or default branch | yes | fast | yes | `CI_COMMIT_BEFORE_SHA` | by paths; a zero, absent or non-ancestor base → full | `branch-push` |
| `merge_request_event` detached | yes | fast | yes | `CI_MERGE_REQUEST_DIFF_BASE_SHA`..HEAD | by paths | `mr-source` |
| `merge_request_event` merged_result | yes | fast | yes | – (the tested merged content includes target-side changes) | **full** (r2; reason `merged-result`) | `mr-merged-result` |
| `schedule` on the default branch (weekly) | yes | fast | yes | – | **full** | `branch-schedule` (new) |
| `web` (Run pipeline; explicit full / release intent) | yes | fast | yes | – | **full** | `branch-web` (new) |
| tag, `api`, `trigger`, `pipeline`, `parent_pipeline`, merge train, other | not created, or rejected 16 | | | | | |

Workflow rules, in order:
1. MR event;
2. `push && $CI_COMMIT_BRANCH && $CI_OPEN_MERGE_REQUESTS && $CI_COMMIT_BRANCH != "staging" && $CI_COMMIT_BRANCH != $CI_DEFAULT_BRANCH` → never;
3. `push && $CI_COMMIT_BRANCH`;
4. `schedule && $CI_COMMIT_BRANCH == $CI_DEFAULT_BRANCH`;
5. `web && $CI_COMMIT_BRANCH`;
6. never.

The gate's rules are source-only (MR; push to `staging`/default; schedule; web). There are no `changes`, `exists`, `when` or `allow_failure` keys anywhere.

**Attribution changes in `validate.sh classify`** (minimal; everything else unchanged):
- **push:** any branch is accepted. The "push to a branch outside the workflow rules" refusal is removed, because that context is now supported.
- **schedule:** must be on the default branch, with no MR variables and `CI_COMMIT_REF_NAME` equal to the branch. Otherwise it is rejected with "schedule outside the default branch".
- **web:** must have a branch and no MR variables.
- **Unchanged:** tags, merge trains, `api` and other sources still fail with 16. Detached, merged-result and push attribution stay distinct (Q13).

### 2.4 Path classes (`infra/ci/profile-paths.tsv`, new reviewed data)

Format: one `class<TAB>kind<TAB>path` row per line. `kind` is `file` (exact match) or `dir` (a prefix ending in `/`). An exact file wins, otherwise the longest directory prefix. Malformed rows, duplicates, absolute paths or `..` fail closed (16).

| Class → profile | Entries |
|---|---|
| `docs` → fast (r2: status, proposal and history prose only) | files `management/v0/PROGRESS.md`, `management/v0/BACKLOG.md`; dir `docs/history/` |
| `runtime` → integration | dirs `core/`, `web/`, `intelligence/` (unit tests colocated with the code included) |
| `foundation` → full | **Governance, architecture, security and governing test contracts (r2):** dirs `docs/` (covers `ARCHITECTURE.md`, `SECURITY.md`, `DECISIONS.md`, `adr/`, `agents/`, `foundation/`, any new file) and `management/` (covers `goals/`, `specs/`, `plans/`, `results/` including scope allow-lists, `ROADMAP.md`, `BUDGET_POLICY.md`, any new file); files `AGENTS.md`, `CLAUDE.md`. The exact `docs` entries above are the only exceptions. **Executable docs:** file `README.md`. **CI, compose and build:** files `.gitlab-ci.yml`, `.env.example`, `.gitignore`; dirs `.archetype/`, `skills/`, `infra/` (compose, Dockerfiles, validation, CI, profile and control rules, gitleaks policy, init scripts). **Contracts and oracles:** dir `contracts/` (schemas and conformance oracle). **Dependencies and locks:** `core/go.mod`, `core/go.sum`, `web/package.json`, `web/package-lock.json`, `intelligence/pyproject.toml`, `intelligence/uv.lock`. **Migrations:** dirs `core/migrations/`, `core/internal/migrate/`, `core/cmd/migrate/`. **Build and test configuration:** `web/next.config.ts`, `web/tsconfig.json`, `web/eslint.config.mjs`, `web/postcss.config.mjs`, `web/components.json`, `web/playwright.config.ts`, `web/next-env.d.ts`. **Harness-bound fixtures and oracles:** dirs `web/tests/browser/`, `web/tests/fixtures/`; file `intelligence/tests/conformance_exchange.py`. |
| *no match* → full | Unknown or new top-level paths; reason `unknown-path` |

### 2.5 Selection algorithm (`infra/ci/select-profile.sh`, new; bash and git only, no Docker)

1. **Inputs.** The selector takes `--job fast|gate|local`, `--requested fast|integration|full` (local only), the attribution already validated by `classify`, the map, and the out-file path.
2. **Context first.**
   - `local` → the requested profile (reason `local-explicit`).
   - Feature push → `fast` (`feature-push`).
   - `schedule` → `full` (`schedule`).
   - `web` → `full` (`web-explicit`; this is also the release-intent path).
   - Merged-result MR → `full` (`merged-result`; r2). The tested merged content includes target-side changes, so it is never classified from the source diff alone.
   - Detached MR or `staging`/default push → classify (steps 3–5).
3. **Base validation.** The base must be 40 hex digits and not all zeros. It must pass `git cat-file -e <base>^{commit}` and `git merge-base --is-ancestor <base> <target>`, and it must differ from the target. Otherwise the result is **full**, with reason `missing-base` or `degenerate-base`. Jobs set `GIT_DEPTH: "0"` because the project's default depth is 20.
4. **Diff.** `git diff -z --name-status -M --no-ext-diff --no-textconv <base> <target>`:
   - A/M → class of the path;
   - **D, R, C, T, U, X, B → full** (`deletion-or-rename`; rename source and destination are both recorded; Q2);
   - a path with control characters, `..` or a leading `/` → full (`unsafe-path`).
5. **Result.** The required profile is the maximum over all rows. The reason comes from the first path, in byte order, that reaches the maximum. An empty diff with a valid base gives fast (`no-change`).
6. **Per-job outcome.**
   - The fast job always executes fast.
   - The gate executes integration or full when required. It executes **none** (`not-required`) only when the pipeline profile is fast; `validate.sh` re-asserts that invariant.
   - A gate in a feature-push or local context is refused (16).
7. **Output.** The selector writes KEY=VALUE pairs, imported only through `state_import` with fixed regular expressions, and `selection.tsv` (status, class, escaped path) into the scanned evidence.

### 2.6 How the full profile stays identical in coverage

- The 69 required IDs and their check bodies are unchanged. The exceptions are deliberate subject updates:
  - B-CI-CONFIG asserts the new reviewed YAML (§4);
  - B-SHELL and B-HADOLINT add the new script and Dockerfile to their target lists;
  - `record()` also appends a time row (§4).
- The full completeness gate is unchanged: 69 IDs exactly once and PASS, browser root/off/rename statistics, both exact 120-case contract sets, both schema-negative reports and an empty leftover inventory. The out-of-profile-row rule is vacuous for full.
- W1–W9 (B-WRAPPER-CONTROLS), C1–C12/C8b/C8c (B-DEPS-CONTROLS) and B-CI-CONTROLS stay in full, through the `controls` suite.
- All 89 CI controls keep their assertions. The two exceptions, whose old refusal scenarios became supported contexts, are redefined as intended negatives with exact reasons (§6.1). V1–V18n run with the default full profile.
- `run.sh all`, the `--owned-registry` and `--secrets-file` paths, postcheck, scans, cleanup and publication are unchanged. Without `--profile`, `validate.sh` defaults to full.

## 3. Hosted execution profile (F3, F4; AC7c)

### 3.1 Arrangement (Model A: DinD service with a context container)

- **Job.** Tag `saas-linux-medium-amd64`; `image: docker:29.8.2-cli@…`; one service `docker:29.8.2-dind@…`, alias `docker`.
- **Variables.**
  - `DOCKER_HOST=tcp://docker:2376`, `DOCKER_TLS_CERTDIR=/certs`, `DOCKER_TLS_VERIFY=1`, `DOCKER_CERT_PATH=/certs/client`;
  - `GIT_STRATEGY=clone`, `GIT_DEPTH="0"`, `GIT_CLEAN_FLAGS=-ffdx`;
  - `FF_GIT_URLS_WITHOUT_TOKENS="true"`, so no job token is written into the checkout's Git configuration (verified in P0).
- **Adapter `infra/ci/hosted-context.sh` (new; POSIX sh for the outer busybox shell; CI-only).** It does nothing outside GitLab CI with the TLS variables (16). Steps:
  1. Wait, with a bound, for the DinD daemon.
  2. Set `NS=$(dirname "$CI_PROJECT_DIR")` and create `T=$NS/.platform-tmp-$CI_JOB_ID`. T is mode 700, must not already exist, and lies outside the checkout. The registry sibling is unchanged.
  3. Build `platform-ci-context:$CI_JOB_ID` from `infra/docker/ci-context.Dockerfile` on stdin, with no build context.
  4. Run:

     ```sh
     docker run --rm --init --name platform-ci-$CI_JOB_ID --network host \
       --security-opt no-new-privileges \
       -v "$NS:$NS" -v /certs/client:/certs/client:ro -w "$CI_PROJECT_DIR" \
       -e TMPDIR="$T" -e DOCKER_HOST=tcp://localhost:2376 -e DOCKER_TLS_VERIFY=1 \
       -e DOCKER_CERT_PATH=/certs/client <allow-listed -e NAME…> \
       <image> bash infra/ci/validate.sh "$@"
     ```

     - The allow-list is the existing `ATTR_VARS` plus `CI_PROJECT_DIR`, `CI_JOB_STARTED_AT`, `CI_COMMIT_BEFORE_SHA`, `CI_MERGE_REQUEST_DIFF_BASE_SHA` and `RUNNER_SCRIPT_TIMEOUT`.
     - Variables are passed as `-e NAME` only, so no value appears in `argv`.
     - `CI_JOB_TOKEN`, `CI_REGISTRY_*`, `CI_DEPENDENCY_PROXY_*`, `CI_DEPLOY_*`, ID tokens and project or group variables are never passed.
     - The container gets no `--privileged`, no host PID and no socket mount.
  5. Propagate the exact exit status.
  6. Forward TERM and INT to the named container with a bounded stop.
  7. Run a 30 s sampler (`docker system df` and `docker stats --no-stream`; no container creation). It prints peak lines to the job log as INFO, non-gating.

  `after_script` runs `sh infra/ci/hosted-context.sh --after-script`. That stops any still-running context, then runs the unchanged `cleanup-owned.sh --after-script` inside a fresh context. It always exits 0.
- **Why it satisfies F3.**
  - Inside the context, `127.0.0.1` is the DinD daemon's loopback, where Compose publishes. `hostname -I` gives the daemon's non-loopback address, so RP6, RP8, RP9, B-PORTS-ROOT/OVERRIDE and the boundary checks run unmodified and still prove loopback-only binding.
  - The job's `/builds` volume is shared with the DinD service (GitLab DinD documentation; verified in P0). The checkout, registry and `$TMPDIR` therefore resolve to identical files for the daemon (RP7, every `-v $REPO…`, `$WORK` mounts).
  - No bind address is widened, no host or shared socket is mounted, and no port-isolation test changes.
- **Context image (`infra/docker/ci-context.Dockerfile`, new).**
  - It is two stages: `FROM docker:29.8.2-cli@… AS dockercli` and `FROM buildpack-deps:trixie-scm@…`.
  - It copies `/usr/local/bin/docker` and `/usr/local/libexec/docker/cli-plugins/` (static compose and buildx) from the CLI stage. There is no apt/apk step, so hadolint passes with no ignore.
  - It provides bash 5, GNU coreutils, findutils, grep (`-P`), sed and tar, `util-linux setsid`, `procps ps`, git and `hostname -I`. awk is `mawk`; the harness's awk use is POSIX.
  - The base also carries unused tools (mercurial, python3). B-DOCKER-ONLY's shadowing still applies.
  - It runs as root, the same uid as the runner's checkout, so `cleanup-owned`'s owner check holds. It is unprivileged.
- **Contingent models, each needing Codex review before implementation.**
  - **Model B:** if P0 shows `/builds` is not visible to the daemon, the fix is a bounded copy-in/copy-out of the checkout into the daemon filesystem at the identical path, with a tree-hash identity check. That is a substantial adapter change, so stop.
  - **TLS fallback:** if `/certs/client` is not shared, plain `tcp://docker:2375`, job-private, with isolation evidence.
  - **Pull limits:** if P0 shows Docker Hub rate limiting, DinD `--registry-mirror=https://mirror.gcr.io`.
  - **Engine 28:** if Engine 29 breaks controls, the §0.4 pins.

### 3.2 Timeouts and budgets (F4)

The budget formula becomes `min(CI_JOB_TIMEOUT, RUNNER_SCRIPT_TIMEOUT) − elapsed − reserve`, then capped per profile.
- `elapsed` is measured from a strictly validated `CI_JOB_STARTED_AT`. This also fixes the disclosed N6 issue, where the deadline was counted from the harness launch.
- In CI, a missing or malformed `CI_JOB_STARTED_AT` → 16. Local runs keep the 14400 s budget.

| Job / profile | Job `timeout` | Script / after-script | Reserve | Harness cap | Minimum budget | RP12 floor (`CI_JOB_TIMEOUT`) |
|---|---|---|---|---|---|---|
| `fast` | 30m | 25m / 4m | 300 s | 1200 s | 600 s | ≥ 1800 |
| `gate` → integration | 3h | 2h45m / 10m (unchanged) | 1200 s | 5400 s | 1800 s | ≥ 10800 |
| `gate` → full | 3h | 2h45m / 10m | 1200 s | none (the existing ≈ 8700 s minus elapsed) | 3600 s (unchanged) | ≥ 10800 |
| `gate` → none | 3h | — | — | no harness | — | preflight with fast thresholds |

- The runner maximum is 10800 s (verified). Job-level 3h overrides the project's 3600 s, to be confirmed in P0.
- Fast values are provisional and are calibrated from measured H1 fast timings only through a reviewed change. They are never lowered for full.
- Both jobs keep `interruptible: false`. Auto-cancel could otherwise drop a protected-branch push pipeline whose diff a later push would not re-classify.
- `resource_group` is removed. Each job gets its own VM and daemon, which P0 verifies through boot identity and uptime.

### 3.3 Capacity (RP11; profile-specific, measured)

| Profile | vCPU | RAM | Free in `$TMPDIR` and in daemon storage |
|---|---|---|---|
| fast | ≥ 2 | ≥ 4 GiB | ≥ 15 GiB (provisional; calibrated to measured peak × 1.5) |
| integration | ≥ 4 | ≥ 8 GiB | ≥ 40 GiB (same workload class as full) |
| full | ≥ 4 | ≥ 8 GiB | ≥ 40 GiB (unchanged) |

- **Measuring daemon storage under DinD.** `DockerRootDir` is not readable in the context. RP11 therefore measures daemon storage with an owned `docker run --rm --network none busybox df -Pk /` container (cidfile-bound, like RP6/RP7) instead of reporting it as INFO.
- **Testable thresholds.** Threshold decisions move into pure `lib.sh` functions (`ci_capacity_ok`, `ci_time_ok`), so controls can test them without a fault switch.
- **If medium has less than 40 GiB free.** The nominal 50 GB disk does not prove ≥ 40 GiB free. If P0 shows less, stop and return the measured free and peak for an explicitly reviewed adjustment, which may require Codex to authorize an extended probe. No blind lowering, size upgrade or purchase follows.

## 4. Files (exact) and changes

| Path | Change |
|---|---|
| `.gitlab-ci.yml` | Workflow rules as in §2.3. `stages: [fast, gate]`. Jobs `fast` and `gate` with the keys in §3.1/§3.2; `gate` adds source-only `rules` and `needs: [fast]`. Scripts: `sh infra/ci/hosted-context.sh -- --profile fast\|gate --publish-dir ci-evidence`. Artifacts unchanged (`when: always`, 90 days, developer access, `ci-evidence/`). No `allow_failure`, `when`, `only`/`except`, retry, parallel, trigger, environment, release, pages, cache, extends, include, dependencies or `resource_group`. |
| `infra/ci/validate.sh` | `--profile fast\|gate\|integration\|full`. CI accepts fast or gate; local accepts fast, integration or full; the default is full. The selector runs inside the isolated preflight before `runner-preflight.sh --profile`. New `state_import` keys. The harness gets the profile's suites (`all` for full). `do_complete` uses the profile ID and report sets plus the out-of-profile-row rule. Gate `none` path records `harness=not-required` and is valid only when the gate's pipeline profile is fast. `classify` adds the schedule and web contexts. Staged evidence adds `selection.tsv` and `timings.tsv`. New summary fields. Tested-content values `branch-schedule` and `branch-web`. |
| `infra/ci/lib.sh` | Profile constants (suites, reserves, caps, minimums, floors, capacity); `ci_budget_seconds` with elapsed time; pure `ci_capacity_ok` and `ci_time_ok`. |
| `infra/ci/runner-preflight.sh` | `--profile`. RP11 and RP12 through the lib functions. Daemon-storage measurement in a nested container. RP3 detail adds the storage driver and cgroup version. RP13 detail adds uptime and a boot-id digest. Still RP1–RP13, each exactly once. |
| `infra/ci/select-profile.sh` (new) | §2.5. |
| `infra/ci/profile-paths.tsv` (new) | §2.4. |
| `infra/ci/hosted-context.sh` (new) | §3.1 adapter, including `--after-script`. |
| `infra/docker/ci-context.Dockerfile` (new) | §3.1 context image. |
| `infra/ci/validate-controls.sh` | The stub harness emits rows and reports only for its suite arguments (`all` expands as in `run.sh`). The template copies the new files. New and redefined controls as in §6.1. |
| `infra/validation/run.sh` | Suites `lint controls provenance contracts unit root migrations ops rename off`; aliases `static`, `variants` and `all`; `suite_unit` (B-UNIT-CHECKS); per-suite `timings.tsv`; usage text. B-CI-CONFIG is rewritten for the reviewed YAML (§6.1 rule list), with profile/suite consistency checks (`all` ≡ the full profile constant; every TSV suite label known; fast ⊆ integration; IDs per profile reported) and README block equality. B-SHELL adds `select-profile.sh` and `hosted-context.sh`; B-HADOLINT adds `ci-context`. |
| `infra/validation/lib.sh` | `record()` also appends `check<TAB>ID<TAB>epoch` to `$WORK/timings.tsv` (statuses and IDs only). |
| `infra/validation/required-checks.tsv` | Suite labels refined (§2.1); `B-UNIT-CHECKS<TAB>unit` added; header updated. |
| `README.md` | `platform:ci` becomes `bash infra/ci/validate.sh --profile full` (explicit full). New blocks `platform:ci-fast` and `platform:ci-integration`. CI paragraph. |
| `docs/foundation/OPERATIONS.md` | CI section: profiles, triggers, path classes, the hosted medium DinD profile (replacing "DinD/hosted not supported"), timeouts and capacity, authoritative status, the release-reuse rule, cost notes, local commands and owner settings. |
| `management/v0/results/G003-scope-allowlist.txt` | Append the four new paths and this Plan's path. |
| `management/v0/results/G003-foundation-observability-ci-result.md` | Append §15 after evidence. §1–§14 are untouched. |

**Explicitly unchanged:**
- application code, schemas, migrations, Compose files and service Dockerfiles;
- dependencies and lockfiles, existing image pins, UI and Skills;
- `cleanup-owned.sh` and `run-approved.sh`;
- the Goal, both Specs, Plan r4, PROGRESS and ADRs.

## 5. Schema, contracts and interfaces

- **Database:** none.
- **Public API and MCP:** none.
- **CI interfaces changed:**
  - `validate.sh --profile`;
  - new tested-content values and summary fields;
  - job names `fast` and `gate`;
  - README markers.
- **Identity:** UUID and domain boundaries are untouched. Evidence carries only infrastructure identifiers: SHAs, pipeline and job IDs, and a boot-id digest.
- **Datastores:** PostgreSQL, Redis and MinIO remain disposable Compose dependencies inside the job daemon. No R2 or other external data plane is used.

## 6. Tests and validation

### 6.1 Controls and static rules (new or changed; all run in B-CI-CONTROLS in full)

| Group | Control | Expected outcome |
|---|---|---|
| A — profile completeness (F1) | **VP1**: fast with exactly the fast rows and reports | 0; summary `executed_profile=fast`; published set = the fast set |
| | **VP2**: a selected ID missing (B-UNIT-CHECKS) | 14 |
| | **VP3**: a selected ID FAIL with harness 0 | 14 |
| | **VP4**: a selected report missing (Go contracts) | 14 |
| | **VP5**: a manufactured out-of-profile row (B-BROWSER-ROOT PASS in fast) | 14 |
| | **VP6**: integration missing `browser-off` | 14 |
| | **VP7**: full given only fast rows (a fast result offered as full) | 14 |
| | **VP8**: unknown `--profile`, or `--profile gate` locally | 16 |
| | **VP9**: gate, docs-only MR | 0, `harness=not-required`, `pipeline_profile=fast`; a gate on a feature push is refused with 16 |
| | V1–V13f (full) | unchanged |
| B — attribution (F2) | P1–P5, P3F | unchanged |
| | **P6**: feature push → `branch-push` | fast |
| | **P7**: schedule on `main` → `branch-schedule` | full |
| | **P8**: web → `branch-web` | full |
| | **N9** (redefined; its old scenario is now supported): push with `CI_COMMIT_REF_NAME ≠ CI_COMMIT_BRANCH` | 16, exact reason |
| | **N12** (redefined): schedule on a non-default branch | 16, "schedule outside the default branch" |
| | **N17**: `api` source | 16, "unsupported pipeline source" (source recorded) |
| | **N18**: web carrying an MR IID | 16 |
| | N1–N8, N10, N11, N13–N16 | unchanged, exact reasons |
| D — budgets and capacity (F4) | **V36** `ci_capacity_ok` table: fast passes at 2 CPU / 4 GiB / 15 GiB and fails just below; integration and full fail at 39 GiB, 3 CPU or 7 GiB | as stated |
| | **V37** `ci_time_ok` and budget | fast 1800 ok; fast with job timeout 900 → fail; gate 10800 ok, 3600 → fail; malformed `CI_JOB_TIMEOUT`, `RUNNER_SCRIPT_TIMEOUT` or `CI_JOB_STARTED_AT` → 16; elapsed time leaving less than the profile minimum → 16; the profile cap applies |
| | V14–V18n | unchanged |
| F — selector (AC7b; unit-level fixture repos, no Docker) | docs-only (`management/v0/PROGRESS.md`); web-only; core-only; intelligence-only; contracts; Compose; each lockfile; build config (Next config, Dockerfile); CI/profile/control (`.gitlab-ci.yml`, `profile-paths.tsv`, `validate-controls.sh`); fixtures and oracles (`web/tests/browser`, `web/tests/fixtures`, `contracts/conformance`); README, `docs/foundation`, `docs/SECURITY.md`; unknown new path; deletion; rename; mixed docs+web (maximum); missing, zero, absent, non-ancestor and degenerate base; new feature branch; protected push with zero before-SHA; MR source; schedule; web/release intent; local explicit; gate on docs-only (`none`); malformed map (duplicate, bad class, absolute path); unsafe path name; empty diff (`no-change`) | Exact `pipeline_profile`, `executed_profile` and `reason` for each case; malformed map → 16 |
| F — governance and security (r2; one fixture commit per path, in a detached-MR context) | Each alone: `docs/ARCHITECTURE.md`, `docs/SECURITY.md`, `docs/DECISIONS.md`, `docs/adr/ADR-002-gitlab-hosted-ci-runner.md`, `docs/agents/CLAUDE-SESSION-HANDOFF.md`, `AGENTS.md`, `CLAUDE.md`, `management/v0/goals/G003-foundation-observability-ci.md`, `management/v0/specs/G003-ci-runner-feedback-revision-spec.md`, `management/v0/plans/G003-ci-runner-feedback-revision-plan.md`, `management/v0/results/G003-scope-allowlist.txt`, `management/v0/ROADMAP.md`, `management/BUDGET_POLICY.md`, a new unlisted `docs/x.md`, a new unlisted `management/x.md` | Each: `pipeline_profile=full`, reason `foundation`; the gate executes full |
| | Each alone: `management/v0/PROGRESS.md`, `management/v0/BACKLOG.md`, `docs/history/initial-product-architecture-discussion.md` | Each: fast, reason `docs-only`; the gate executes `none` |
| | `PROGRESS.md` + `management/v0/specs/…-spec.md` together; `BACKLOG.md` + `web/app/page.tsx` together | full, `foundation` (maximum); integration, `runtime` (maximum) |
| F — merged result (r2) | `merged_result` pipelines, each case separately: source change docs-only (`PROGRESS.md`); source unchanged with a target-side runtime change; source and target both docs-only | Every case: `pipeline_profile=full`, reason `merged-result`; the gate executes full; no diff is consulted |
| G — adapter (docker shim on PATH) | **G1**: outside CI or without TLS variables | 16, no docker call |
| | **G2**: `docker run` exit propagation | 0→0, 14→14 |
| | **G3**: argument audit (canary `CI_JOB_TOKEN` and `CI_REGISTRY_PASSWORD` set in the environment) | Only allow-listed `-e NAME`; no `NAME=VALUE` for CI variables; no `--privileged`, host PID or socket mount; mounts exactly `$NS` and `/certs/client:ro` |
| | **G4**: `--after-script` | Always 0; stops the named context |
| | **G5**: pre-existing or symlinked `$T`, or `$T` outside `$NS` | 16 |
| Static B-CI-CONFIG (lint, so in fast as well) | Exact top-level keys, workflow rules, stages, per-job key sets, forbidden keys, gate rules and needs, pinned image and service, tag, timeouts with script plus after-script ≤ job timeout, variables (no secret-like names), `interruptible: false`, exact scripts and after-script, exact artifacts, README blocks, profile/suite consistency, CLI-stage pin equal to the job image | PASS; any deviation → FAIL |

Control count: 89 existing controls (with N9 and N12 redefined), plus 9 profile, 5 attribution (P6–P8, N17, N18), 2 budget, about 40 selector (r2: plus the governance, status and merged-result scenarios) and 5 adapter controls. `--list` and B-CI-CONTROLS's exact-set rule cover them.

### 6.2 Local focused validation (implementation phase; using the existing wrapper and binding)

| Step | Command | Expected |
|---|---|---|
| L1 | `run-approved.sh … -- --base <BASE> --allow-file management/v0/results/G003-scope-allowlist.txt --report-dir $EXT/rev/l1/report lint provenance contracts unit` | wrapper 0; 31 required IDs PASS, INFO only otherwise; B-CI-CONFIG, B-SHELL, B-HADOLINT and B-UNIT-CHECKS PASS; B-SCOPE within the allow-list |
| L2 | `bash infra/ci/validate-controls.sh --group A,B,C,D,F,G --log $EXT/rev/l2/controls.log` | Every selected control PASS. Group E (Docker-real ownership; `cleanup-owned.sh` unchanged) runs in hosted full (H1). |
| L3 | Scratch-commit copy (as in c1-s8): `env -u GITLAB_CI bash infra/ci/validate.sh --profile fast --evidence-dir $EXT/rev/l3/ev` | Exit 0; final line status 0; summary `profile=fast`, 31 IDs; `tested_content=local`; repository status digest unchanged |

There are no local integration, full or full-parity runs. The changed coverage wiring is mechanical, and B-CI-CONFIG checks it statically. Its first complete execution is hosted H1. If implementation touches any check body other than those in §2.6, that suite runs locally once, with the reason recorded.

### 6.3 Hosted actions (each needs named authorization; evidence is read with `glab` GET only and stored outside both checkouts)

| ID | Exact action | Purpose and expected evidence | Limits / cleanup | Compute (≈, factor 2) |
|---|---|---|---|---|
| **HA-P0 = P0-lite** (r2; before implementation) | **Setup:** from a disposable clone outside both checkouts, make one commit on parent `cbffd6d` = the bound G003 tree plus probe files (probe-only `.gitlab-ci.yml`, `infra/ci/probe-dind.sh`, candidate `ci-context.Dockerfile`). Push it as `g003-probe-dind`. This **publishes `cbffd6d` and the uncommitted G003 tree** on a disposable branch. **Job `probe-dind`** (medium; §0.4 pins; TLS DinD; `timeout: 3h`; the script stops itself at 25 min; no artifacts; job-log lines only) records the following. (1) **Executor and privilege:** runner, tags and version; VM uptime, boot-id digest, `/.dockerenv` and k8s-variable presence; the job container's and a nested container's CapEff; `docker info` (server, driver, cgroup, CPU, memory). (2) **TLS daemon:** 2376 without a client certificate refused; 2375 closed. (3) **Shared checkout and temp paths:** a marker in `$NS` visible to the daemon at the identical path; job-local `/tmp` not visible. (4) **Loopback and non-loopback:** an owned busybox `127.0.0.1` publication is unreachable from the job shell, reachable from the context, and refused on the context's non-loopback address. (5) **Tool versions:** the context build, plus bash, git, GNU tools, docker, compose and buildx. (6) **Capacity and timeouts:** real `runner-preflight.sh` RP1–RP13 in the context, including RP11 daemon and `$NS` free space and RP12 `CI_JOB_TIMEOUT`. (7) **Credential isolation:** credential-bearing remote URLs counted (0 expected); the context's environment names checked against the allow-list, with no values printed (no `CI_JOB_TOKEN`). (8) INFO only: Docker Hub rate-limit headers from one anonymous HEAD. **Job `probe-timeout`** (medium, `timeout: 3m`, `sleep 600`) records the timeout failure reason and the `after_script` marker. **No harness suite, control group, contracts, root stack or browser run:** H1's full run supplies complete application qualification and the inherited group-E evidence. | F3, F4 and the environment part of AC7c. Any check not as expected is a concrete stop (§8 R1). | Delete the branch after evidence capture. | ≈ 35–40 |
| HA-P0c (optional) | Cancel one 5-minute probe job through `glab api` POST | Hosted cancel semantics (canceled ≠ success; `after_script`) | — | ≈ 10 |
| **HA-H1** | Commit the final code/validation tree on the bound base; the Result stays local and is excluded. Push `g003-foundation-observability-ci` with `-o ci.skip`. Open a **draft** MR to `staging`. | MR pipeline: `fast` (fast evidence on the exact tree) plus `gate` (foundation → **full**, `mr-source`). This is the first hosted full run, serving both as qualification (cold build, peaks, durations, cleanup) and as H1: exact commit/tree, 69 IDs, browsers, contracts, 89+ controls including group E on DinD, authoritative job statuses, final lines and sanitized artifact digests. | The MR stays open for review; nothing is merged. | ≈ 230 |
| **HA-HINT** (r2) | From the disposable clone: **B** = the H1 commit plus one commit appending exactly one line, `// g003 integration-evidence base perturbation; never executed`, to `web/app/page.tsx` (runtime class; status M). **F′** = B plus one commit restoring `web/app/page.tsx` byte-exactly, so **F′'s tree equals H1's tree**. Push `g003-evidence/int-base` (B) and `g003-evidence/int-head` (F′), both with `-o ci.skip`, so the perturbed B is never executed (it is only an MR target). Open a draft MR int-head → int-base (detached; `CI_MERGE_REQUEST_DIFF_BASE_SHA` = B). | Required evidence: (1) `git rev-parse F′^{tree}` = H1's tree SHA. (2) `git diff --name-status B F′` = exactly `M web/app/page.tsx`: nonempty and runtime-only. (3) The gate's `selection.tsv` and summary show `comparison_base` = B, `pipeline_profile` = `executed_profile` = **integration**, reason `runtime`. Classification, profiles and gates are unchanged; this only chooses a disposable MR target, so ordinary foundation MRs to `staging` still select full. Also yields integration timing for the target disposition. | Close the MR; delete both branches. | ≈ 160 |
| **HA-H2** | Push `g003-gate-proof` = the H1 commit plus one commit setting EI39 `valid` (a feature push, no MR) | The real fast job fails: B-CONTRACT-GO/PY/AGREE FAIL, `validate.sh` exits 1 and safe artifacts are published. Hosted feature-push and new-branch context evidence. | Delete the branch. | ≈ 20–25 |
| HA-E | `glab api` GET of pipelines, jobs, traces and artifact downloads | Attributable evidence; no token printed | read-only | 0 |

**No action** merges, pushes to `staging` or `main`, creates schedules or tags, or changes variables, runners or settings. Owner settings in Q12 are separate decisions. Duplicate suppression is asserted statically (B-CI-CONFIG). It is observed on hosted runs only if a natural second push to an open-MR branch occurs; no extra push is made for it.

### 6.4 Re-run policy

H1 and H-INT share one tree. Any code or validation change after H1 requires a new H1, and a new H-INT if integration is affected, with the reason recorded. An environment-only failure is reported first; a fix that changes code consumes a cycle. Local evidence (c1) supports unchanged behavior but never replaces H1. Weekly and release full runs are operational (schedule creation needs separate authorization). The documented reuse rule: a full pass may be reused only when the GitLab job status is success, and summary `profile=executed_profile=full`, the tree SHA, the required-set digest and `images.txt` all equal the candidate's.

## 7. Coverage mapping (F1–F6, AC7a–AC7g)

| Item | Concrete change | Test / control (expected) | Evidence |
|---|---|---|---|
| **F1** fixed full run and completeness | `--profile`; per-profile suites and ID/report sets from `required-checks.tsv` and lib constants; out-of-profile-row rule; summary fields (§2.1–2.2) | VP1–VP9 pass as stated; V1–V13f unchanged | L2, L3; H1, H-INT and H2 summaries and `selection.tsv` |
| **F2** one-job guard and attribution | Two-job YAML; B-CI-CONFIG rewritten as an exact reviewed-structure guard; feature-push, schedule and web contexts (§2.3) | B-CI-CONFIG PASS on the exact YAML and FAIL on deviation; P6–P8 accepted; N9, N12, N17, N18 with exact reasons; the rest of N1–N16 unchanged | L1, L2; H1 `mr-source`; H2 `branch-push` |
| **F3** DinD loopback and paths | Context container on the DinD network namespace; identical `$NS` mount; `$TMPDIR` under `$NS`; TLS; adapter (§3.1) | G1–G5; real RP6–RP9 in the context | P0-lite (path, loopback, TLS and privilege model; preflight); H1 B-PORTS-*, boundaries, browsers, group E |
| **F4** capacity and timeouts | Profile timeouts, reserves, caps and floors; elapsed-time budget; capacity table; daemon-storage measurement (§3.2–3.3) | V36, V37 | P0-lite free space and timeout behavior; H1/H-INT measured durations and peaks (sampler) |
| **F5** expensive static controls | `static` split into `lint` and `controls`; fast excludes `controls` and `migrations`; inventory with timings (§2.1) | B-CI-CONFIG profile consistency; L2 reruns the changed control groups once | `timings.tsv` from H1, H-INT and fast; membership comparison |
| **F6** access and prerequisites | glab GET facts (§0.3); reconciled hosted actions (§6.3) replace A-R1/H0–H3/A-E | — | Result §15 appends the access facts. History §1–§14 is unchanged. |
| **AC7a** | §2.1 inventory: fast 31 ⊂ integration 63; full 69 = the inherited set; `unit` documented | VP1–VP7; B-CI-CONFIG consistency | `required_set_sha256` and counts in each summary |
| **AC7b** | Selector, map and workflow rules (§2.3–2.5) | Group F (every listed case), P6–P8, N-set; B-CI-CONFIG workflow-rule exactness (duplicate suppression) | H1 (foundation → full), H-INT (runtime → integration), H2 (feature push → fast) |
| **AC7c** | Medium profile, pins and adapter (§3) | G1–G5; RP rows | P0 qualification; H1 cold run (runner/job IDs, tag, versions, peaks, cleanup); probe-timeout |
| **AC7d** | Hosted final evidence plan (§6.3) | `do_complete` per profile | H1 full plus fast on the exact tree. H-INT (r2): tree equality with H1, a nonempty runtime-only diff (`M web/app/page.tsx`), and selected = executed = integration. Artifact digests, statuses, final lines. |
| **AC7e** | Fault through the real fast gate; wiring is static and controlled | VP2–VP7; B-CI-CONFIG (no allow-failure; exact scripts); G2 exit propagation | H2 failing job and its artifacts; branch deletion recorded |
| **AC7f** | Timings, sampler, OPERATIONS profile commands, reuse rule, cost (§3, §6.4, §10.2) | VP1 and VP7 (a fast summary cannot pass as full) | Measured cold phase durations, factors and quota note; target dispositions (Q3) |
| **AC7g** | Allow-list append; binding wrapper; Result §15 append only; no G048, product or deploy work | B-SCOPE, B-PROTECTED, wrapper pre- and postchecks | Binding gate output; protected-input hashes unchanged |

## 8. Sequence and gates

- **R0 — prerequisites.** ADR-002 accepted; Plan ACCEPT; Codex binding (base, an input snapshot including the revision Spec, the accepted ADR-002, this Plan and Plan r4, allow-list additions); HA authorizations; quota decision.
- **R1 — P0, before any implementation write.**
  - **Proceed** on Model A plus TLS, with ≥ 40 GiB free and every P0-lite check as expected. r2: no harness, control or browser slice belongs to this gate; H1 provides that qualification.
  - **Otherwise stop and return evidence:** Model B, TLS fallback, mirror, Engine 28, capacity. Any of these may need Plan r2.
- **R2 — binding precheck.** The existing pattern; no write before it passes.
- **R3 — implementation, in order.** `run.sh`/`lib.sh`/TSV → `infra/ci/lib.sh` → selector and map → `validate.sh` → preflight → adapter and Dockerfile → YAML and B-CI-CONFIG → controls → README/OPERATIONS → allow-list.
- **R4 — local validation.** L1–L3.
- **R5 — review.** Self-review plus one read-only CI/security reviewer.
- **R6 — H1.** Evidence.
- **R7 — H-INT.**
- **R8 — H2.**
- **R9 — cleanup.** Delete disposable branches and close evidence MRs.
- **R10 — handback.** Append Result §15; return to Codex.

The checkpoint comes after cycle 2. Execution never writes to the main checkout.

## 9. Delegation

One primary implementer, the parent, owns every file, integration and the shared-cycle accounting. One bounded read-only reviewer for CI/security checks the adapter, selector, completeness and YAML after L1–L3. That review is 0 cycles. There is no implementation delegation; the change set is small and tightly coupled.

## 10. Budget and compute forecast

### 10.1 Cycles

**Proposed separate B3** (maximum 3 shared meaningful cycles; checkpoint after cycle 2). It is not activated here. The original **B7 (6/7)** and **correction pass 1 (0/3)** remain separate and unchanged.

- **Forecast:** 1–2 cycles. The likely candidates are one local cycle (selector, completeness or B-CI-CONFIG expressions) and up to one hosted cycle at H1.
- **What costs no cycle:** P0, evidence retrieval, reviews and successful runs.
- **Risk:** skipping P0, or hosted surprises (Engine 29 store, pull limits, capacity, browser, timing), could exhaust B3. At exhaustion, stop and return for a split or budget review; no self-extension.
- **Scope:** about 16 paths; no split recommended.

### 10.2 Compute (medium factor 2; hosted runs are cold every time, with no persistent layer cache)

| Run | Forecast job time | Compute minutes |
|---|---|---|
| P0-lite (r2: the qualification probe) | ≈ 15 + 3 min | ≈ 35–40 |
| H1 (fast 12–17 min + full 70–115 min cold) | — | ≈ 230 |
| H-INT (fast + integration 45–75 min cold) | — | ≈ 160 |
| H2 (fast fails at contracts) | ≈ 10–12 min | ≈ 20–25 |
| **G003 evidence total** | | **≈ 445–455 (+ ≈ 230 if H1 must rerun)** |

- **Quota.** The documented Free quota is 400 compute minutes per month for the whole `<namespace>` namespace (14 projects; actual remaining quota unverified). Execution therefore still needs an owner purchase decision, since ≈ 445–455 exceeds 400; at $10 per 1000 minutes, one pack covers it.
- **Ongoing cost.** Each pipeline runs fast at ≈ 25–35 compute minutes, including status and docs pushes to `staging`. r2: a `staging` push or MR that changes governing documents (Goal, Spec, Plan, Result, architecture, security, decisions) now also runs full. A foundation merge to `staging` re-runs full on the merged content. A weekly full costs ≈ 1000 compute minutes per month.
- **Fast target.** Fast will likely exceed the 5/10-minute target cold. Disposition comes after measurement (Q3).

## 11. Risks and escalation

| Risk | Handling |
|---|---|
| ADR-002 not accepted, or its wording | No execution. Hosted runners run every job container privileged through runner configuration; this Plan relies on privilege only through the DinD service, and P0 records CapEff (Q7). |
| Executor or filesystem model differs (Model B) | Stop at R1 with evidence; Plan r2 needed. |
| Medium disk < 40 GiB free; full run cold > 2h25m | Stop with measurements for reviewed disposition; no blind threshold change. |
| Docker Hub pull limits; Engine 29 containerd store | P0-lite reads the pull-limit headers. Engine 29 store compatibility first shows in H1 (full harness, group E). A mirror or the Engine 28 fallback only after review (Q8, Q9). |
| Quota or cost | Owner decision before R1 (§10.2). |
| Merges not yet gated (`only_allow_merge_if_pipeline_succeeds` false; `staging` unprotected); Developer-role pipeline-variable overrides could spoof predefined variables | Owner settings decisions (Q12). The selector cross-checks base ancestry and degeneracy and fails safe to full. |
| Security model | Nested app and test containers stay unprivileged. The daemon is job-private behind mutual TLS. The job token is kept out of the context and out of `.git/config`. No deploy or provider secret is used, and artifacts stay at Developer access. |
| Gates | G048 execution and the G021 product barrier stay closed. No application, UI, schema, deploy, purchase or schedule work is part of this Plan. |

## 12. Questions for Codex

1. **Feature-branch pushes:** fast by context, with no path classification, since the merge gate is the MR pipeline. Path classification there would select full on every new branch because there is no base.
2. **Deletions and renames:** always full (the literal Spec reading), or classify both paths?
3. **Integration scope and targets:** one conservative all-sides profile (63 IDs; ≈ 31 min local warm; 45–75 min hosted cold). After H-INT, either accept the measured targets or request bounded per-side or parallel splits as separate work. Fast's cold overrun likewise.
4. **Vacuous gate:** is `not-required` (gate passes with no harness for fast-only content) acceptable, rather than GitLab `changes` rules, which would create two sources of truth?
5. **B-UNIT-CHECKS:** fast-only, so full keeps exactly 69 IDs, rather than adding `unit` to `all`.
6. **The §2.4 path map (r2):** confirm that only `management/v0/PROGRESS.md`, `management/v0/BACKLOG.md` and `docs/history/` stay fast, with all other `docs/` and `management/` content, `AGENTS.md` and `CLAUDE.md` selecting full. Also confirm the runtime/foundation boundaries.
7. **ADR-002:** should the text acknowledge that runner-wide privilege also applies to job containers? Are the criteria for a non-TLS job-private fallback acceptable?
8. **Pins:** Engine 29.8.2 primary versus 28.5.2 (graph driver; unmaintained since 2025-11); `buildpack-deps:trixie-scm` as the package-manager-free context base.
9. **Docker Hub limits:** conditionally pre-approve a DinD `--registry-mirror` (`mirror.gcr.io`) if P0 shows limiting, or require Plan r2?
10. **Quota:** purchase before R1? The total is ≈ 445–455 compute minutes, plus ≈ 230 as contingency, against a shared 400 per month.
11. **AC7d construction (r2):** is H-INT acceptable as built? Its disposable base B adds one `web/app/page.tsx` line and is never executed; its head F′ restores H1 byte-exactly. Is H1 on the code/validation tree with the Result kept local acceptable?
12. **Owner settings, separately authorized:** "Pipelines must succeed", pipeline-variable role, `staging` protection, the weekly schedule, and `keep_latest_artifact` retention beyond 90 days. Which, if any, must precede G003 acceptance?
13. **Manual source:** only `web`. `api` (glab-created) pipelines fail closed. Confirm.
14. **Final base:** commit on `cbffd6d` (MR diff against `staging`'s merge base) or integrate with current `staging` `329ba45` first. This is a binding decision.
