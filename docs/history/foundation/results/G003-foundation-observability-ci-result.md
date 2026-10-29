> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G003 — Health, logging, contracts and baseline CI: implementation result

Goal: [G003-foundation-observability-ci.md](../goals/G003-foundation-observability-ci.md) · Spec: [G003-foundation-observability-ci-spec.md](../specs/G003-foundation-observability-ci-spec.md) · Plan: [r4](../plans/G003-foundation-observability-ci-plan.md) (SHA-256 `10a827884bb811f808fc82bc230df9d12b95e237b84d93a36f3067348d00ccb4`)
Owner: Claude Code · Reviewer: Codex · Budget: **B7**

## 1. Status

**Submitted for Codex review, together with the G047 Result, under the owner-authorized combined execution (not included). Not complete and not accepted.** Only Codex ACCEPT completes G003; this Result does not mark itself accepted and does not edit PROGRESS, the Goal, the Spec or the Plan.

Recommended review outcome to consider: **ARCHITECTURE DECISION REQUIRED** (two items), then a bounded correction or acceptance decision:

1. **Plan premise F11 is false on the execution host (AC1/AC3 assertions O1–O3, O7, O9).** Plan F11 states "A *stopped* intelligence fails fast and is unaffected"; §6.3 builds the stop-case expectations on it. Measured here, Docker's embedded DNS answers a lookup for a stopped container's name only after **≈ 4 s** (3.99 s, scratch busybox pair). Every `compose stop` case therefore behaves like a hung dependency and runs into the nested 2 s deadlines that Codex's Q4 kept. Observed in every run once the outage records were read correctly (S7-r1, S7-r2, S8, S10): web `/api/ready` 503 with `core_probe: timeout` for O1–O10 (enabled) and O1–O6/O9/O10 (disabled); **O7 web 503 instead of 200** (core direct 200 with `optional.intelligence = unavailable`, so non-gating at core holds). The strict Plan assertions were **kept** (not weakened), so B-OBS-TIMEOUTS and B-OBS-OFF FAIL on exactly those assertions and every final run is non-zero. Codex/owner decide: accept host-dependent `not_ready|timeout` / `unreachable|timeout` labels and disclose O7 (F11 race), change the deadline design (for example an optional-probe deadline shorter than core readiness), or change the injection.
2. **Hosted evidence (AC7) is blocked.** No GitLab API token or runner-host access exists in this session; the runner facts F-R1–F-R13 (including the tag F-R2) and owner acknowledgement F-R13 are unresolved. R1, H0–H3 were not run; `.gitlab-ci.yml` carries the placeholder tag `unresolved-owner-runner-tag`. Local parity is **not** reported as hosted success.

Everything else has attributable local evidence (§7–§8): the final complete wrapper-driven run on the delivered tree (S10, run `6421a3`, suites `all`) has **wrapper 1, `tee` 0, postcheck passed, overall 1: 66 PASS, 3 INFO, 3 FAIL, 0 UNMET**. The three FAILs are B-OBS-TIMEOUTS and B-OBS-OFF on exactly the F11-premise assertions, and B-REQUIRED-CHECKS naming only those two. Both browser matrices pass 103/103 (root and intelligence-disabled) and the identity rename 2/2, with 0 skipped, flaky or unexpected. All 80 CI controls pass. Local CI parity through `validate.sh` (S8) ran end to end with the owned registry and failed only on the same assertions.

- **B7: 6 of 7 cycles consumed**, checkpoint after cycle 4 recorded (§10). Cycle 7 was not used: the remaining failures need a Codex decision, not an implementation change.
- The foundation barrier stays closed through G021.

## 2. Execution binding and starting state

| Item | Value |
|---|---|
| Binding | `/tmp/g047-g003-19f945a-approved-binding.json`, SHA-256 `ffb8b7c6…ead9`; retained byte-identical copies at `~/.local/state/platform-evidence/batch-19f945a/` (`RETAINED.sha256`) |
| G003 BASE | Transition checkpoint `cbffd6da5d9936c90fdfd9e51d84d36c0f17ad81` (tree `59868332…8f90`, single parent = initial BASE `19f945af…`), per the addendum's local checkpoint transition; mechanical pin/count reconciliation recorded outside the repository (`batch-19f945a/transition/g003-mechanical-reconciliation.md`, SHA-256 `30237d16…`) |
| G003 snapshot | `/tmp/g003-19f945a-approved-inputs.sha256`, SHA-256 `3fb544d344ea341bc3fff2b2a5000718efb15d3d216c233c01419fc9bcf103ec`, 5 entries (Goal `50143f34…`, Spec `751cfd22…`, Plan `10a82788…`, PROGRESS `fc0c711f…`, ARCHITECTURE `6a6fd49b…`), unchanged throughout |
| Execution checkout / branch | `.claude/worktrees/g047-exec` / `worktree-g047-exec` (same checkout as G047; no other worktree created) |
| Evidence root | `EXEC=69407d`; `EXT=~/.local/state/platform-evidence/g003-69407d` (outside both checkouts and `/tmp`; mode 700) |

**S0 (Plan §15.1):** the extracted gate script hashed to `0409148a…6723` as pinned; result `S0 PASS: HEAD cbffd6d…; snapshot 3fb544d3…; 5 inputs verified; 0 differing path(s); 0 tool-local ignored file(s); no .env, no beginify resources, default ports free; compose 2.32.4-desktop.1` (`$EXT/g003-exec/s0/`). Wrapper `--check-only` and `run-approved-controls.sh` (W1–W9) passed. Unrelated developer containers/networks/volumes (`payments-assistant-*`, `needknow-*`, `outreacher-*`) were present throughout and never touched.

**S0R (R1 runner qualification): not run — blocked** (`$EXT/g003-exec/s0r/r1-status.md`). A-R1 and A-E are authorized, but the Q1 facts and an evidence-retrieval mechanism are unavailable with existing access (SSH git transport works; there is no API token; no runner host). An untagged probe could run on GitLab.com shared runners (§12.1 stop condition) with no way to read its result, so no branch was pushed. Per the addendum, local implementation proceeded.

**S1:** the allow-list was the first repository write (SHA-256 `4d802263f127cc00959942e73628f1ad33df58dd0c8ffe8746cf45ac5aa78fe3`, mirrors Plan §14).

## 3. Implementation summary and files changed

191 paths differ from BASE `cbffd6d` (17 modified tracked files, 174 new), all within the §14 allow-list (B-SCOPE). Plan §14 lists 192: the 192nd is this Result. Delegated components had exact, non-overlapping ownership (§9 item 9); the parent integrated and validated everything.

| Area | Paths | Change |
|---|---|---|
| Core | `core/internal/observability/{correlation.go,logging.go}` (customized); new `exit.go`, `errorlog.go` and tests `context_test.go`, `exit_test.go`, `errorlog_test.go`; `core/internal/health/checks.go` (customized) and new `propagation_test.go`; `core/cmd/server/main.go` (three reviewed hunks, SHA-256 `987c9335…51ea`) | §4 context profile, per-request completion record, bounded exit/server-error records, bind-then-log, outbound header allow-list on the intelligence probe |
| Web | `web/lib/request-context.ts` (new, server-only), `web/lib/core-client.ts`, `web/app/api/{health,ready}/route.ts`, `web/package.json` (scripts only: `test:unit`, `check`), `web/tests/unit/request-context.test.mjs`, `web/tests/fixtures/error-probe/{wrapped,control}-route.ts` | Context derivation, two-header outbound, one completion record with closed `instanceof` categories; fixtures are test-only (B-WEB-BOUNDARY) |
| Intelligence | `intelligence/src/observability/{http.py (new),export.py}`, `src/main.py`, `src/streaming/schemas.py`; tests `test_health_context.py`, `test_logging_privacy.py`, `test_error_output.py`, `test_conformance.py`, `conformance_exchange.py`; `infra/docker/intelligence.Dockerfile` (check stage `COPY` only) | ASGI completion middleware, JSON formatter with redaction and uvicorn takeover, startup-failure record, validator conformance corrections |
| Contracts | `contracts/conformance/` (128 files: README, `expectations.json`, 120 cases, SN02–SN07) | Declared corpus (31 valid, 89 invalid; all eight event types), schema-negative and digest controls; corpus SHA-256 `a01f1a85…a87c` |
| Validation tools | `infra/validation/tools/**` (18), `infra/docker/validation-tools.Dockerfile` (+ `.dockerignore`) | `contractcheck`, `httpprobe`, `logcheck`; single direct dependency `santhosh-tekuri/jsonschema/v6 v6.0.3`; distroless nonroot runtime |
| Harness | `infra/validation/{run.sh,lib.sh,render-customizations.tsv}`, new `required-checks.tsv` (69 IDs) | §10: contracts suite, error-output checks, B-HEALTH-ROOT, B-OBS-*, B-REQUIRED-CHECKS, B-CI-CONFIG/CONTROLS, register 19, `APPLICABLE` 35, B-MCP-SOURCE reconciliation, `--owned-registry`/`--secrets-file`, capture hooks, `docker_run_owned`, the `tee` line |
| CI | `.gitlab-ci.yml`, `infra/ci/{validate.sh,lib.sh,validate-controls.sh,cleanup-owned.sh,runner-preflight.sh,unchanged.allow}` | One gating job; supervised, fail-closed entry point with finalizer, scans, publication outcomes, owned-resource cleanup; 80 controls (groups A–E) |
| Docs | `README.md` (`platform:contracts`, `platform:ci`, health line, link), `docs/foundation/OPERATIONS.md` (new), `docs/foundation/FOLLOW-UP.md` (labels) | Operations guide; FOLLOW-UP: metrics `[ ]` (partly done), contract validation `[x]`, CI `[ ]` (partly done; hosted evidence awaits the runner) |
| Management | `management/v0/results/G003-scope-allowlist.txt`, this Result | — |

Key final hashes: `.gitlab-ci.yml` `a7fa0214…0aa2c`; `infra/ci/validate.sh` `0d27bc58…1e37`; `infra/ci/validate-controls.sh` `d3ada4dc…2249`; `infra/ci/cleanup-owned.sh` `2d5f3d64…5304`; `infra/ci/lib.sh` `52b68617…5dc8`; `infra/ci/runner-preflight.sh` `eeaa2d17…b19d`; `infra/validation/run.sh` `17620169…6ea6`; `infra/validation/lib.sh` `abe240dd…bc22`; `required-checks.tsv` `60df2fec…b36a`; `render-customizations.tsv` `8b7e1903…020d`.

## 4. Migrations and schema

None. No database schema, migration, role or RLS change. `contracts/intelligence/*.schema.json` are byte-unchanged (digests `bf3cec36…a4f`, `e1025031…63fd102`, pinned and checked first by both validators).

## 5. Contracts, API and MCP

- No new route, provider call, public API or MCP change. B-MCP-SOURCE keeps every security assertion; its byte-equality test is reconciled to `MAIN_GO_G003_SHA = 987c9335…51ea` (the reviewed three-hunk `main.go`) plus the pinned import set and "no jsonschema in the server's deps", as Plan §10/§17 specify.
- Observable health-contract additions are only those the Spec pre-authorizes: `X-Request-Id` on web and intelligence health routes, `Cache-Control: no-store` on intelligence health routes. Bodies and statuses are byte-identical (B-HEALTH-ROOT exact-body checks).
- The AI wire schemas are unchanged; the Python validator corrections (§8.4) make `streaming.schemas` conform to the unchanged schemas (ECMA-262 pattern anchoring, RFC 3339 `date-time`, strict JSON) without broadening production execution.

## 6. Tests added and modified

| Runtime | New tests | Run by |
|---|---|---|
| Go (core) | `context_test.go` 13, `propagation_test.go` 4, `exit_test.go` 5, `errorlog_test.go` 3 test functions (tables: valid/absent/malformed/oversized 65, 513, 8 KiB/duplicated/comma-joined/zero/uppercase/versions 01 and ff/extra fields/secret-valued; 64-way concurrency; cancellation; panic; implicit 200 vs 503; header allow-list on a real recorder; 5 s double → readiness at about 2 s) | `core-check` (B-CHECKS-ROOT) |
| Node (web) | `request-context.test.mjs` (`node:test`; 81 tests reported by sa-web's gate run: tables, exactly two outbound keys, 200-way concurrency, throw/abort records, closed categories) | `web-check` (B-CHECKS-ROOT) |
| Python (intelligence) | `test_health_context.py` 15, `test_logging_privacy.py` 16 (incl. the unredacting-stub positive control and, from cycle 6, no raw-text fallback), `test_error_output.py` 14, `test_conformance.py` 20 (+ `conformance_exchange.py`) | `intelligence-check` (B-CHECKS-ROOT, B-CONTRACT-TOOLS); 421 passed, 0 skipped |
| Tools | `contractcheck` 6, `httpprobe` 6, `logcheck` 6, `conformance` 15, `logrecord` 10 test functions (incl. logcheck unit negatives) | `validation-tools` check stage (B-CONTRACT-TOOLS) |
| Harness | B-HEALTH-ROOT; B-CORE/WEB/INTEL-ERROR-OUTPUT (CE1–CE3, WE0–WE6 with WE3a/WE3b and the WE4 positive control, IE1–IE2); B-OBS-CONTEXT/CHAIN/TIMEOUTS/LOGS/PRIVACY/OFF; contracts suite (6 checks); B-CI-CONFIG; B-CI-CONTROLS (80 controls); B-REQUIRED-CHECKS | `run.sh` suites |

Existing tests are unchanged and pass (inherited `observability_test.go`, `health_test.go`, intelligence and web suites, both 103-test browser matrices and the rename run).

## 7. Validation commands and outcomes

All wrapper runs use `$EXT/bin/g003-vw.sh <step> <suites…>` (SHA-256 `c181488a…8a25`): `bash infra/validation/run-approved.sh --snapshot /tmp/g003-19f945a-approved-inputs.sha256 --snapshot-sha256 3fb544d3… --count 5 --path-re '<Plan §15.4 expression>' --plan management/v0/plans/G003-foundation-observability-ci-plan.md -- --base cbffd6d… --allow-file management/v0/results/G003-scope-allowlist.txt --report-dir $EXT/g003-exec/<step>/report <suites>`, under `bash --noprofile --norc` with immediate `PIPESTATUS` capture and the G047 precedence (wrapper status, then tee 70, then record 71). Every run is listed, including failures; later successes do not erase them.

| Step | Command / suites | Run | Outcome |
|---|---|---|---|
| S0 | §15.1 gate (`0409148a…6723`), wrapper `--check-only`, `run-approved-controls.sh` | — | PASS (S0 PASS; precheck passed; W1–W9 PASS) |
| S0R | R1 runner qualification | — | **Not run: blocked** (§2) |
| S1–S5 | Delegated component gates (Plan §15.6): intelligence check (S1/S5), tools check + corpus (S2), core-check + CE1–CE3 (S3), web-check + build + WE0–WE6 (S4), IE1/IE2 (S5) | — | All final gates PASS (sa-intel 420 passed, 0 skipped; tools 45 test functions; core-check; web-check incl. 81 unit tests). One material failure (cycle 1) |
| S6a (sa-ci) | shellcheck both passes; runner-preflight; controls runs 1 and 2 | — | Controls run 1 78/79 (V18a), run 2 78/79 (V18d); cycles 2–3 |
| G6a | shellcheck `-S warning` (10 scripts) + `-S info --include=SC2251` (infra/ci) | — | PASS (and again as G6a-r2, G6a-r3) |
| G6b | `validate-controls.sh` standalone, all 80 | — | **exit 1: 79/80** (V4: nested run aborted under `set -e` on the comments-only allow-list); V18d fixed (cycle 4 closed) |
| G6b-V4 | `--only V4 --keep-scratch` (diagnostic) | — | FAIL, diagnosed (nested GO/PY/AGREE FAIL as intended, then the final-check abort) |
| G6c | wrapper `static provenance contracts` | `ac8b47` | **wrapper 1, overall 1: 30 PASS, 3 FAIL** (B-CI-CONTROLS V4, B-RENDER-PLAN count 8, B-REQUIRED-CHECKS) |
| S7 (diag.) | wrapper `root migrations variants` on the unchanged tree | `ef66c4` | **overall 1: 35 PASS, 5 FAIL** (B-WEB-/B-INTEL-ERROR-OUTPUT, B-OBS-TIMEOUTS, B-OBS-OFF, B-REQUIRED-CHECKS) → cycle 5 |
| G6b-r2 | controls, all 80 | — | exit 0, 80/80 |
| G6c-r2 | wrapper `static provenance contracts` | `f5f65d` | **wrapper 0, overall 0: 33 PASS, 0 FAIL** |
| S7-r1 | wrapper `root migrations variants` | `4caaa7` | overall 1: 37 PASS, 3 FAIL (F11 premise only) — cycle 5 closed |
| S9 | SA-P, SA-C read-only reviews | — | Findings triaged → cycle 6 (§12) |
| cycle-6 component checks | shellcheck; intelligence check image; `--only V27,V27b` | — | ruff E501 on one comment (trivial; re-wrapped) → 421 passed; V27/V27b PASS |
| G6b-r3 | controls, all 80 | — | exit 0, 80/80 |
| G6c-r3 | wrapper `static provenance contracts` | `a35424` | **wrapper 0, overall 0: 33 PASS, 0 FAIL** |
| S7-r2 | wrapper `root migrations variants` | `b46059` | overall 1: 37 PASS, 3 FAIL (F11 premise only) — cycle 6 closed |
| S8 | local CI parity: scratch-commit copy (HEAD `8cf72160…`), `env -u GITLAB_CI bash infra/ci/validate.sh --evidence-dir …` | — | `validate_rc=1`: final line `validate: final status 1; published=none; registry=removed; harness_stop=none; phases: location=0 preflight=0 state=0 harness=1 postcheck=0 cleanup=0 raw_scan=0 complete=14 stage=0 summary=0 candidate_scan=0 report=0 scrub=0`; 66 PASS, 3 FAIL (F11 premise); `tested_content=local`; repository status digest identical before and after; scratch copy deleted |
| **S10** | **wrapper `all`, no edits during the run** | **`6421a3`** | **wrapper 1, tee 0, postcheck passed, overall 1: 66 PASS, 3 INFO, 3 FAIL** (B-OBS-TIMEOUTS, B-OBS-OFF, B-REQUIRED-CHECKS); `results.tsv` SHA-256 `a57887e2…d711` |
| S11 | static + provenance on the final tree including this Result | `s11` (recorded after this file) | Run on the final tree including this Result; its outcome is recorded in `$EXT/g003-exec/runs.tsv` and the Codex handoff (it cannot be written into the file it validates) |
| H0–H3 | hosted pipelines | — | **Not run: blocked** (§2) |

**S10 outage evidence** (`report/obs-outages.tsv`, `obs-off-outages.tsv`, `obs-timings.tsv`): every request ≤ 2004 ms (bound 5000 ms); every recovery ≤ 2.7 s (bound 90 s); web `/api/ready` 503 with `core_probe: timeout` in all enabled O1–O10 and disabled O1–O6/O9/O10 cases; core records `cancelled` (status 503, or 200 for O7/O8); O9 no core record; O10 at most one core record, `cancelled`. Strict expectations failing: O1–O3 `not_ready`, O9 `unreachable`, O7 web 200/`ok`.

## 8. Acceptance criteria

| AC | Status | Evidence (S10 run `6421a3` unless noted) |
|---|---|---|
| AC1 Health semantics and bounds | **Partly met; F11 premise escalated** | B-HEALTH-ROOT (six routes: status, exact body, no-store, valid `X-Request-Id`), B-WEB-HTTP, B-OUTAGE, B-TOGGLE-OFF, unit tests PASS; every outage request ≤ 5000 ms and recovery ≤ 90 s (`obs-timings.tsv`); core 503 names each dependency; liveness 200; O7/O8 core 200 with `optional.intelligence = unavailable`. **FAIL:** B-OBS-TIMEOUTS / B-OBS-OFF strict labels O1–O3, O9 and O7 web 200 (§1 item 1). CE3 binds and logs only `port=8180`, `app_env=test`. |
| AC2 Context validation, generation, isolation | Met (local) | Go/Node/Python table tests incl. secret-valued IDs and 64/200/100-way concurrency (B-CHECKS-ROOT); B-OBS-CONTEXT PASS (7 request-ID and 9 `traceparent` rows through web, each with records and no invalid literal in any raw log); header allow-list: `propagation_test.go`, `outboundHeaders` unit test, B-WEB-BOUNDARY. |
| AC3 web → core → intelligence continuity | Met for continuity; outage labels as AC1 | B-OBS-CHAIN PASS: 62 cases (supplied context, none, invalid replaced, 32 concurrent, O1–O10, O7 recovery): counts, equal request/trace IDs, parent = caller span, distinct spans. Disabled variant: all 9 web → core chain cases PASS inside B-OBS-OFF (which fails only on the AC1 labels). |
| AC4 Structured logs and privacy | Met (local) | B-OBS-LOGS (every line of all three raw logs catalogued; one record per request; a 200 and a handled 503), B-OBS-PRIVACY (canary headers, query, GET body, invalid IDs, execute-endpoint prompt/context-ref/authorization-ref, secret-valued IDs to core and intelligence; planted-canary control; records exist), B-CORE/WEB/INTEL-ERROR-OUTPUT (CE1–CE3, WE0–WE6 with WE3a/WE3b and WE4 positive control, IE1–IE2), B-SECRET-CANARY, disabled-variant privacy with control; unit privacy tests incl. the leak-stub control. |
| AC5 Cross-language contracts | Met (local) | B-CONTRACT-TOOLS/GO/PY/AGREE/PEER/NEGATIVE PASS: 120 cases each, 0 skipped, identical verdicts equal to expectations, all eight event types, three equal corpus digests; Go-produced envelopes executed by Python; real Python NDJSON accepted by Go; four corruptions rejected; SN01–SN07, SA01/SA02, sn-derive, SD01/SD02 (0 evaluated), EI39 flip rejected by both. |
| AC6 Containerized, runnable documented checks | Met (local) | README blocks run verbatim (B-CONFIGURE, B-CHECKS-ROOT, browser, migrate); B-DOCKER-ONLY (shadowed host runtimes, zero invocations); images from locked inputs; 103/103 ×2 + 2/2 browsers; cleanup inventory empty (`inventory-leftover.txt` empty). |
| AC7 Hosted pipelines | **Not met — blocked** | `.gitlab-ci.yml` structure B-CI-CONFIG PASS (placeholder tag); local parity S8 only. No R1, H0, H1, H2 or H3 evidence; required hosted pass and hosted failing case missing. |
| AC8 Negative controls, no false green | Met (local) | B-CI-CONTROLS: the exact 80-control set PASS (groups A–E, incl. V4's nested real contracts run, supervision V18–V18k, Docker-real cleanup V19–V35, V27 two-call binding); B-CONTRACT-NEGATIVE; B-WRAPPER-CONTROLS W1–W9; B-DEPS-CONTROLS C1–C12/C8b/C8c; B-REQUIRED-CHECKS (it is FAIL in S10 precisely because two required checks failed); wrapper postcheck passed on every run. Deferred SA-C minors disclosed (§12). |
| AC9 Scope, architecture, provenance | Met (local) | B-SCOPE (191 paths within the allow-list), B-PROTECTED, B-SKILL-UNCHANGED, B-RENDER-ID, B-RENDER-PLAN/REPEAT (exactly 19), B-TOGGLE-OFF-ADDITIONS (35), migrations, boundaries and controls, ports, B-RENAME/B-BROWSER-RENAME/B-DBRENAME, B-PERSIST, B-MCP-SOURCE/B-MCP-HTTP; no schema, route, UI or vendor change. |
| AC10 Docs, Result, governance | Met for documentation; completion gated | OPERATIONS.md, README blocks, FOLLOW-UP labels; this Result maps AC1–AC10, separates inherited/new/open and local/hosted evidence, lists every run and failure, the B7 cycles and the cycle-4 checkpoint. Required evidence is not complete (AC1 labels, AC7), so G003 is not complete. |

## 9. Deviations and disclosures

1. **Plan premise F11 (escalated, §1 item 1).** Not a code deviation: the strict §6.3 assertions are implemented and fail on this host. Measurement: busybox `nslookup` of a stopped container's alias from a peer on a labelled scratch network took 3.99 s (127.0.0.11, `ndots:0`); the scratch containers and network were removed. Observed outage table: `obs-outages.tsv` / `obs-off-outages.tsv` in each S7/S10 report.
2. **Next 16.3.8 startup banner (R14).** Real output differs from Plan §5.4's anticipated lines (no leading spaces; `- Local:`/`- Network:` spacing; no `Starting...`; an extra `✓ Running next.config took …`). `logrecord.WebStartupAllowList` holds exactly the five observed patterns, each at most once per process start (a new `▲ Next.js` line starts a new process); no empty-line exemption.
3. **`app/api/__probe` is private in Next.** B-WEB-ERROR-OUTPUT copies the fixtures to `app/api/%5F%5Fprobe/{wrapped,control}/route.ts` in the disposable copy; the URL stays `/api/__probe/...`.
4. **IE2 exit 0.** The real image runs uvicorn as PID 1 (exec-form CMD, no init); its re-raised SIGTERM has default disposition and is ignored by the kernel, so the container exits 0. The Plan names no IE2 code; the exact observed code is asserted. (sa-intel's lifecycle test, run not as PID 1, accepts 0 or SIGTERM.)
5. **`intelligence.message` catalogue event** for application-owned intelligence log calls (fixed constant for non-string messages; logger `^intelligence(\.[A-Za-z0-9_-]+)*$`; no extra keys).
6. **Usage token counts appear as `[REDACTED]`** in the intelligence `execution` line: the inherited sensitive-key regex matches `token`. No redaction rule was weakened; `events.json` accepts the redacted value.
7. **Placeholder runner tag** `unresolved-owner-runner-tag` in `.gitlab-ci.yml` (matches no runner); otherwise byte-identical to Plan §11.1. B-CI-CONFIG's PASS text names the tag in use; runner qualification is outside that static check.
8. **FOLLOW-UP labels:** CI stays `[ ]` (partly done) rather than the Plan's `[x]`, so the document does not claim hosted completion.
9. **Implementation subagents.** Plan §20 named one implementation owner; the owner instruction and the accepted addendum (§"Delegation and final handoff") allow non-overlapping delegated implementation. Five implementation subagents had exact ownership (sa-intel: intelligence + contracts; sa-tools: validation tools; sa-core: core; sa-web: web; sa-ci: CI) under a shared brief; the parent owned the shared harness, docs and allow-list, and integrated and validated everything. Codex should confirm this fits the addendum's "non-overlapping contract/test implementation" wording. Two read-only reviewers (SA-P, SA-C) ran at S9 as planned.
10. **sa-ci implementation choices (§11.2/§11.4 detail):** start inventory taken before runner-preflight (V28); P1 also pulls the pinned yq image used by F4/F5 (SA-C N4); TMPDIR inside the repository rejected in both modes; gitleaks exits 17 on unreadable or non-regular entries; image removal by owned tag re-resolved to the registered ID (`docker image rm TAG`, never `-f`) rather than `rm ID`; unparseable browser JSON left out of staging (F4 still 14); extra job-log line `validate: tested_content=… head=…`; 3 s settle before stragglers. The Compose working_dir rule accepts the repository, a path under it (Compose records `<checkout>/infra/compose`) or a path under the run's `work_root` (SA-C N3).
11. **Stricter completion rules in `logcheck`:** `trace_source` must agree with the presence of `parent_span_id`; status 0 only with `cancelled`/`failed`. Real logs pass (B-OBS-LOGS).
12. **B-OBS-PRIVACY strengthened:** generated secrets are 32 hex (a valid request-ID format), so the check also requires the reply ID to differ from the secret and scans all raw logs for the secret values.
13. **Host `python3`** was used as a text-editing helper by the parent (scripted, all-or-nothing file edits) and once by sa-intel; no project validation ran on the host (B-DOCKER-ONLY PASS).
14. **Intelligence exception status (SA-P F-4, fixed):** an exception before the response starts is recorded with status 500, the status FastAPI's outermost `ServerErrorMiddleware` sends; that framework 500 carries neither `X-Request-Id` nor `no-store` (it is produced outside the middleware). Health handlers do not raise in practice.
15. **Web method scope (SA-P F-5, disclosed):** web logs the requests Next routes to the exported `GET` handler (GET, and HEAD through it); other methods are answered by Next before the handler and are not logged. OPERATIONS.md states this; Codex may want it as a scoping decision.
16. **Intelligence disconnects (SA-P F-7):** uvicorn does not cancel the ASGI task when a client disconnects, so intelligence records `completed`; core and web record `cancelled`. Documented in OPERATIONS.md.
17. **No implicit 200 on the running stack (SA-P F-9):** every real health handler sets its status explicitly; implicit-200 accounting is proven by core's `TestMiddlewareStatusAndResponseBehaviour` against a real server. B-OBS-LOGS wording says so.
18. **Web→core allow-list on the wire (SA-P F-10):** proven at object level (unit test: exactly two keys) and by the B-WEB-BOUNDARY static rule, not by capturing what Next's fetch sends to core.
19. **`core/cmd/migrate` stderr (SA-P F-12)** now uses the shared envelope (its pre-existing free-text message appears under `event`); outside `logcheck`'s scope.
20. **Plan text counts:** appendix (r1/r2) text says 18 register rows / 67 required IDs; the live r3/r4 text says 19 / 69 (B-ARTIFACT-SCAN withdrawn; error-output checks added). The live values are implemented and checked.

## 10. Execution budget, cycles, checkpoint and unresolved issues

**B7: maximum 7 shared meaningful cycles; consumed 6** (sa-intel 1, sa-ci 2, parent 3; sa-core, sa-web, sa-tools 0; reviewers 0). Counted per BUDGET_POLICY §3/§5: a material failure followed by diagnosis, corrective implementation and retest, by any agent. Trivial authoring slips (formatter, unused variable, one over-long comment line) are listed in the cycle log but not counted.

| Cycle | Agent | Failure → correction → retest |
|---|---|---|
| 1 | sa-intel | Real-uvicorn lifecycle test expected exit 0 after SIGTERM; uvicorn 0.54.0 re-raises → test accepts 0 or SIGTERM; check image 420 passed |
| 2 | sa-ci | F4 rejected valid contract JSON: in yq 4.48.1 the right side of `and` loses context → one vector expression per file; V1 passed |
| 3 | sa-ci | V18a: stub's TERM-ignoring sleep ended before the 150 s grace → 600 s (plus V4 scratch-path ordering); 78/79 |
| 4 | parent | V18d control race (watchdog looked up once, before it exists) → bounded retry; G6b 79/80 with V18d PASS |
| — | — | **Checkpoint after cycle 4** (2026-10-06): per-criterion status, convergence (no repeated material failure), forecast 1–3 more within the 3 remaining → continue |
| 5 | parent | G6b/G6c/diagnostic S7 round: comments-only allow-list grep under `set -e` (V4/B-SCOPE); register count 8→19; WE5 input unreadable and not malformed; IE2 exit code; outage records read before log capture; O10 variable overwrite → G6b 80/80, G6c overall 0, S7 only F11 items |
| 6 | parent | S9 review: cidfile counter in `$(...)` (blocker for every `--owned-registry` run); secret list beside an external registry; name-pattern image records; chain-iv vacuity; per-service log table; CE3 values; context-row completeness; WE5 value; B-OBS-OFF privacy (Plan §10); exact control set; late-signal status; intelligence exception status; logging fallback → G6b 80/80 (V27 two-call), G6c overall 0, S7 only F11 items, S8 B1 verified under the registry |

**Unresolved:** (1) the F11 premise (Codex decision; cycle 7 deliberately not spent and no assertion weakened); (2) hosted R1/H0–H3 (blocked prerequisites); (3) deferred SA-C minors m2, m3, m4, m6, m7, m8 (§12). Attempted approaches for (1): measured and attributed the cause; kept the Plan's strict assertions; recorded per-case statuses for review. For (2): read-only access investigation (`batch-19f945a/g003-runner-facts.md`); no push, no runner change.

## 11. Provenance, integrity and cleanup

- **Inputs:** the G003 snapshot (5 entries) verified by the wrapper precheck and postcheck on every run (all "postcheck passed"); the binding and both snapshots retained byte-identically at `~/.local/state/platform-evidence/batch-19f945a/` (`RETAINED.sha256`); `/tmp` copies still verified at the end.
- **Scope:** B-SCOPE PASS on every run (191 paths, then 192 with this Result in S11). Nothing outside the allow-list; `.env` and ignored files never read; no push, merge, MR or pipeline; no project setting, runner or credential change.
- **Isolation:** no other worktree was created. The separately requested G048 planning clone lives outside both checkouts (`~/.local/state/platform-evidence/g048-planning/platform`) and is not part of G003.
- **Docker hygiene:** every run's `inventory-leftover.txt` is empty; harness images are run-labelled and removed by the harness; execution-owned images (`g003.exec=69407d`) are removed at the end and checked against the S0 inventory (11 execution-built images `g003-69407d-*`, none present in the S0 inventory, removed by tag without `-f`; no execution-labelled container, network or volume remained; scratch V4 copy, S8 scratch clone and the S2 module cache deleted); unrelated developer resources untouched.
- **Evidence:** `$EXT/g003-exec/` — `cycle-log.md`, `runs.tsv`, each step's `harness.out` and `report/`, `g6a*/`, `g6b*/`, `s8-r1/validate.out`, `s9/` triage, `s0/`, `s0r/`; scratch trees with synthetic canaries deleted. No secret value is retained (B-SECRET-CANARY, report copies hold statuses and categories only).
- **Commit:** the delivered tree (this Result included) is left **uncommitted** in the execution worktree on `worktree-g047-exec`, on top of the transition checkpoint `cbffd6d`. The addendum authorized only G047's checkpoint commit; a G003 commit is made only with the owner's go-ahead (local, no push), and its hash is then reported. Nothing was pushed, merged or integrated into main.

## 12. Review (S9) and remaining findings

Two read-only reviewers ran at S9 (0 cycles each; triage files `$EXT/g003-exec/s9/sa-p-triage.md`, `sa-c-triage.md`).

- **SA-P (correlation and privacy):** no runtime privacy leak; §4, §5.3 and §5.5 hold in all three runtimes. Fixed in cycle 6: F-1 (MAJOR, chain-iv evidence), F-2, F-3, F-4, F-6, F-8, F-9 (wording), F-11. Disclosed: F-5, F-7, F-10, F-12 (§9 items 14–19).
- **SA-C (CI, contracts, provenance):** fixed in cycle 6: **B1 (BLOCKER)** — `docker_run_owned` incremented its cidfile counter inside `$(...)`, so every validate.sh/CI run would have failed CE3, WE, IE2 and migrate-it (now a unique per-call name; V27 calls it twice through `$(...)`); **M1** — the de-duplicated secret list (`secrets.u`) was written beside an external `--secrets-file` and would survive in a kept registry (now under `$WORK`); **M2** — Compose images were registered by a `<project>-` name pattern, the rule Codex withdrew in r3 (now exact build services read from the Compose files, only after an explicit build); m1, m5, N1, N2. **Deferred and disclosed for Codex** (not fixed, to avoid unreviewed control redesign within the last cycle): m2 F4 contract-result check counts no `match` keys (defence in depth; B-CONTRACT-GO/PY/AGREE PASS is required through results.tsv; a stronger (id, valid, class) comparison was prototyped and matches real results but needs stub changes); m3 N1–N16 do not pin the rejection reason; m4 nested `timeout`/`setsid` processes in validate-controls.sh escape the harness process-group stop on cancel/deadline; m6 (suspicion) P1-pulled scanner images could be captured through an interrupted harness's secret-scan container; m7 superseded IDs of rebuilt registered tags are neither removed nor reported; m8 many controls assert exit status and one phase rather than every generic §11.3 assertion; N5 no attribution end time; N6 deadline counted from harness launch; N7 V4's Python leg fails on its declared-class precondition; N8 third-party images re-pulled each CI run; N9 no static R24 check; N10 RP13 counts all running containers.

## 13. Questions for Codex

1. **F11 premise (decision required):** for stop-case outages on hosts whose Docker DNS does not fail fast, choose: (a) accept `not_ready|timeout` (O1–O3) and `unreachable|timeout` (O9) and treat O7 like O8 (web 200 or 503 `timeout`, recorded) — the Plan's own F11 rule applied to the observed condition; (b) change the deadline design so optional intelligence cannot consume core's whole readiness budget (for example a shorter optional-probe deadline); or (c) change the injection. Any choice is a Plan/assertion change for Codex, not something this Result applies.
2. **Hosted prerequisites (AC7):** Q1 runner facts F-R1–F-R13 and the tag, F-R12 evidence-retrieval mechanism (for example a scoped read token) and F-R13 owner acknowledgement. With them, R1 then H0–H3 can run on exact evidence branches as authorized.
3. **Delegation fit (§9 item 9)** and the **web method scope** (§9 item 15).
4. **Deferred SA-C minors** (§12): fold into a correction loop or accept as disclosed.
5. **Commit:** the G003 tree is uncommitted in the execution worktree (§11); whether to make a local commit is the owner's decision. Nothing was pushed or merged.

## 14. Correction pass 1 (Codex CORRECTION LOOP, 2026-10-06)

Sections 1–13 above are the reviewed submission (Result SHA-256 `b654111115cfde5cfebe88430ecf067927f037602dcd138a17f6a896229457d2`) and remain unchanged as history; this section appends the correction pass. Where it differs, this section is current.

### 14.1 Binding and starting state

| Item | Value |
|---|---|
| Binding | `~/.local/state/platform-evidence/g003-69407d/correction-1/approved-binding.json`, SHA-256 `6998f4bdf97ab2628fd067941e274a7dfcc42ccb35fee5321d3ba9aa74236892` (supervisor-created) |
| Review | `codex-review.md`, SHA-256 `aa14dc1042fc06ded2d8a95f47dc42acd33b8011d48584d5c2f125fc88eb744a` (five bounded corrections; supersedes only the r4 §6.3 stop-case expectations and the identified CI/control defects) |
| Reviewed submission | `reviewed-submission.json`, SHA-256 `fee4f44fce493f7ab4411556aa7aee68bc49e863a9915dd740fbcb13b05a9876` (192 files with hash and mode; status digest `e0911d2f…` of `git status --porcelain=v1 -z --untracked-files=all`) |
| Unchanged | HEAD/BASE `cbffd6d`; Plan `10a82788…cb4`; protected snapshot `3fb544d3…103ec` (5 inputs); Goal/Spec bytes; no Plan revision |

**Fail-closed precheck before any correction write** (`$EXT/bin/g003-c1-precheck.sh`, SHA-256 `5a595caa22af348f949b12d72ee23e833fbf01113a636fef1287918bf19c194d`; evidence `c1/precheck-0.txt`, repeated by the correction subagent in `c1/precheck-sa-c1.txt`): binding, review and manifest digests; HEAD; NUL-safe status digest and exact 192-path set; every manifest file hash and mode; the five protected inputs, the Plan and the initial Result hash; no `.env`, no root-project resources, default ports free — **PASS**. The original S0 clean-tree assertion was not rerun (the submitted tree is intentionally dirty).

### 14.2 Delegation and ownership

One implementation-correction subagent (sa-c1), in-process, configured effort **xhigh** (inherited from the session and confirmed from its transcript metadata), working only in `.claude/worktrees/g047-exec` from the reviewed submission, with exclusive ownership of `infra/validation/run.sh`, `infra/ci/validate.sh`, `infra/ci/validate-controls.sh` and `docs/foundation/OPERATIONS.md` (`infra/ci/lib.sh` only if demonstrated necessary — not needed, unchanged), and exclusive G003 Docker validation while active. The parent verified the hand-back, owned integration, final validation, cycle accounting, this Result and the read-only hosted investigation. The G048 planning work ran in its separate clone with no Docker activity.

Parent verification of the hand-back: exactly these four files differ from the reviewed manifest (modes unchanged): `run.sh` `7f42828a…43e8`, `validate.sh` `75c988c0…e223`, `validate-controls.sh` `c1b88a85…9083`, `OPERATIONS.md` `01e121cf…ef21e`; HEAD, the 192-path set and the protected inputs unchanged; no ignored file; no execution-labelled, run-labelled or root-project container left.

### 14.3 Corrections

1. **Stopped-service DNS timing (AC1/AC3).** The accepted F11 nested-deadline interpretation now applies to stop cases (`run.sh`: `obs_web_probe`, `obs_outage`, `obs_ops`, `obs_off`, new `obs_observed`). O1–O3: direct core 503 naming the dependency, web 503, liveness 200, recovery; web `core_probe` `not_ready` or `timeout`. O9: web 503, web liveness 200, no core record; `unreachable` or `timeout`. O7: direct core 200 with optional intelligence unavailable, no intelligence record while stopped, recovered three-hop continuity; web exactly `200/ok` or `503/timeout` (status and probe checked as a pair). Every ≤ 5000 ms request bound, the ≤ 90 s recovery bound, O8, O10, the production 2 s deadlines, optional-core semantics, the injection and all status/body/header/correlation assertions are unchanged; the observed values are written into the B-OBS-TIMEOUTS/B-OBS-OFF result lines and the outage tables. OPERATIONS.md's nested-deadline disclosure covers stopped services. The earlier strict-assertion failures (S7-r1, S7-r2, S8, S10) remain recorded history.
2. **Contract-report completeness (AC5/AC8; SA-C m2).** `validate.sh` F4 (`do_complete`) now validates each normal Go/Python report against the checked-out `contracts/conformance/intelligence/expectations.json`: typed fields and token IDs/classes, unique IDs, counts (`total`, `valid`, `invalid`, `by_class`) consistent with the cases, integer `skipped` 0, no mismatch, and the exact sorted `(id, valid, class)` set equal to the declared corpus; each class has its own message and exit 14. Schema-negative reports keep their `match` semantics. Controls use a declared synthetic 3-case corpus in the scratch trees; the stub derives consistent reports from it.
3. **Cancellation of nested controls (AC7/AC8; SA-C m4).** `validate-controls.sh` registers every nested launch of this control run before it starts, runs controls in the background so TERM/INT/HUP reach the trap, and on signal, launch deadline or a launch outliving its control stops exactly its own tree (TERM, 30 s grace, KILL, reap) before sweeping resources and removing scratch; exit 128+n on signal n. Group ownership is accepted only while the recorded leader is alive with its recorded start time (pid-reuse safe). No global, name- or pattern-based killing.
4. **Attribution negative-control reasons (AC8; SA-C m3).** N1–N16 assert each case's exact rejection reason in both the job-log line and `attribution.json`, with their existing exit-16, no-harness-invocation and publication checks (N1 and N12 share validate.sh's fixed "unsupported pipeline source" text and are distinguished by the recorded source). `validate.sh`'s attribution logic is unchanged; the five normal cases and P3F pass.
5. **Hosted completion (AC7/AC10): blocked, not attempted.** Read-only re-investigation with existing access (`c1/hosted-investigation.md`): SSH git transport authenticates; there is no API token or other evidence-retrieval mechanism (no `~/.netrc`, `~/.git-credentials`, glab or python-gitlab configuration, credential helper, `GITLAB_*`/`CI_*` variables or runner binary), and the runner facts are still unresolved (main-checkout PROGRESS.md). The original runner/security gates cannot pass, so the named hosted authorizations were not used: no R1/H0–H3, no push, no MR, no settings or runner change; the placeholder tag is unchanged; missing access was not retried as debug cycles. **Exact unmet prerequisites:** F-R1–F-R13 including the tag F-R2, an evidence-retrieval mechanism (F-R12; for example an owner-provided scoped read token), the owner acknowledgement F-R13; then R1 and H0–H3 on the exact corrected submission.

**Controls:** the set grows from 80 to **89** (`validate-controls.sh --list`): new **V13a** (empty Go case set → 14), **V13b** (Python set truncated → 14), **V13c** (Go duplicate case → 14), **V13d** (Python verdict altered → 14), **V13e** (valid full reports → 0), **V13f** (string verdicts, wrong `counts.total`, `skipped` 1, a mismatch entry → each 14), **V18l** (TERM to the nested runner's group → 143), **V18m** (HUP to its PID → 129), **V18n** (nested launch deadline → 1). V18l–V18n each prove all 14 processes of the nested tree (incl. a TERM-resistant harness and a container-creating holder) are gone, 0 containers remain and stay at 0 after 6 s, the nested scratch is removed and does not reappear, no PASS line and no publication, and three decoys survive (a process whose command line names the nested harness path, a `sleep 300`, an unrelated container). The original 80-control evidence (G6b-r3, S10) stays historical. B-CI-CONTROLS checks the exact `--list` set; `required-checks.tsv` is unchanged (69 IDs).

### 14.4 Validation runs (every run, failures included)

| Step | Command / suites | Run | Outcome |
|---|---|---|---|
| Precheck | `g003-c1-precheck.sh --resources` (parent; repeated by sa-c1 before its first write) | — | PASS |
| G6a | shellcheck both passes (10 scripts; SC2251 over `infra/ci`) | — | rc 0/0 (twice: `c1/g6a-1`, `c1/g6a-2` after the last edit) |
| Focused controls | `validate-controls.sh --only` V13a–f, V18l–n, P1–P5, P3F, N1–N16, V18–V18k, V19–V35 | — | 65/65 PASS (`c1/focused-1`) |
| G6b | `validate-controls.sh` complete, standalone | — | **89/89 PASS**, exit 0, PASS set = `--list` (`c1/g6b-1`) |
| F4 on real reports | `do_complete` on run 0feaa8's real Go/Python reports | — | 0; 14 with one case dropped (`c1/f4-real`) |
| c1-spc-1 | wrapper `static provenance contracts` | `0feaa8` | wrapper 0, overall 0: 33 PASS, 3 INFO, 0 FAIL |
| c1-var-1 | wrapper `variants` | `60bca9` | wrapper 0, overall 0: 20 PASS, 3 INFO, 0 FAIL |
| **c1-all** | **wrapper `all` on the final corrected tree (parent)** | **`3c8cd5`** | **wrapper 0, tee 0, postcheck passed, overall 0: 69 PASS, 3 INFO, 0 FAIL, 0 UNMET**; B-REQUIRED-CHECKS 68 required IDs PASS; B-CI-CONTROLS 89; B-SCOPE 192; browsers 103/103, 103/103, 2/2 (0 skipped/flaky/unexpected); `inventory-leftover.txt` empty; `results.tsv` SHA-256 `04f2d2e9…954a` |
| **c1-s8** | **local CI parity: scratch-commit copy (HEAD `3c388772…202d`), `env -u GITLAB_CI bash infra/ci/validate.sh --evidence-dir …`** | — | **`validate_rc=0`**; final line `validate: final status 0; published=none; registry=removed; harness_stop=none; phases: location=0 preflight=0 state=0 harness=0 postcheck=0 cleanup=0 raw_scan=0 complete=0 stage=0 summary=0 candidate_scan=0 report=0 scrub=0`; 69 PASS; `tested_content=local`; repository status digest identical before and after; scratch copy deleted |
| c1-final | wrapper `static provenance` on the final tree including this appended Result | `c1-final` (recorded after this file) | Outcome recorded in `$EXT/g003-exec/runs.tsv` and the handoff |
| Hosted | R1, H0–H3 | — | **Not run: blocked** (correction 5) |

**Observed outage values (run `3c8cd5`):** enabled O1, O2, O3, O7, O9 = 503/`timeout`; disabled off-O1, off-O2, off-O3, off-O9 = 503/`timeout` — all within the review's set; every single request ≤ 5000 ms and recovery ≤ 90 s (sa-c1's run 60bca9: maximum request 2012 ms, maximum recovery 5650 ms).

### 14.5 Budget

**Correction pass 1: 0 of 3 shared meaningful cycles; checkpoint after correction cycle 2 not reached (no correction cycle was consumed).** No validation run of the correction subagent failed; authoring slips fixed before the first run (an unbound `$!` under `set -u`, the pid-reuse hazard found in self-review, a trailing space) are not cycles. The original implementation budget is unchanged at **B7: 6 of 7**; it was not reset and no correction used the remaining original cycle. Read-only review, hosted-access investigation and successful checks are not cycles.

### 14.6 Status after correction pass 1

- **Local evidence (corrections 1–4): complete.** The final wrapper-driven full run (`3c8cd5`) and complete local CI parity (`validate.sh`, exit 0) pass on the final corrected tree; AC1–AC6 and AC8–AC9 have attributable local evidence, AC10 documentation is complete.
- **AC7 (hosted) is not met and G003 remains blocked and unaccepted** until the hosted prerequisites above are provided and R1/H0–H3 run on the exact corrected submission. Local parity is not reported as hosted success.
- Remaining follow-up observations from the review (other notes/control-coverage refinements, superseded-build-image reporting) are not part of this pass; SA-C m6 is closed by Codex's disposition.
- Nothing committed or pushed; the corrected tree is uncommitted in the execution worktree. The G021 foundation barrier remains closed.

## 15. CI runner feedback revision (revision Spec §11–§12, Plan r2, ADR-002; 2026-10-06)

Sections 1–14 remain unchanged as history; this section appends the CI runner feedback revision. Where it differs, this section is current. Evidence root: `$EXT/g003-exec/rev/` (`$EXT` = `~/.local/state/platform-evidence/g003-69407d`), outside both checkouts.

### 15.1 Authority, binding and inputs

| Item | Value |
|---|---|
| Authority | Owner-authorized combined batch (G047 then G003); owner answer "Yes, proceed fully" for the hosted DinD actions; Codex handoffs: revision Plan r2 ACCEPT, ADR-002 accepted (c43796e), P0 review ACCEPT continuation (R3 onward with the token, disk-guard, 35 GiB and capacity-formula dispositions) |
| Binding | `g003-ci-revision-approval-c43796e/p0-disposition/approved-inputs.sha256` (SHA-256 `acae154f…e5ab`, 9 entries), used through the existing wrapper `infra/validation/run-approved.sh` by `$EXT/bin/g003-cont-vw.sh` (SHA-256 `0b5e4983…`) |
| Inputs (unchanged) | ADR-002 `12329779…de0f`; revision Plan r2 `83dcbd96…2dd3`; revision Spec (with §12) `0ac5bcc1…450c`; P0 review `a467684a…d253`; Goal, Spec, Plan r4 and §1–§14 of this Result unchanged |
| Base / tree | HEAD `cbffd6d`, uncommitted execution tree of 200 paths (199 plus this Result); scope allow-list `d102ec39…0de9` (appended, prefixes intact); every changed path matches it (B-SCOPE 200) |
| Budget | Separate **B3** (max 3 meaningful cycles, checkpoint after 2); original B7 6/7 and correction pass 1 0/3 unchanged |

### 15.2 Hosted qualification before implementation (P0-lite and P0 follow-up; recorded earlier, summarized)

- **P0-lite** (pipeline 2919134219, disposable branch, deleted): Model A, TLS (2376 mutual TLS, 2375 closed), shared build/temp paths at identical paths, loopback semantics, tools and timeouts as expected. **Two stop findings:** (1) RP11 capacity: 37.9 GiB free on the 45.0 GiB medium VM (one filesystem shared by daemon storage and the build directory) against the 40 GiB floor; (2) `.git/config` held a userinfo URL despite `FF_GIT_URLS_WITHOUT_TOKENS=true`. A plain job-level timeout skips `after_script`.
- **P0 follow-up** (pipeline 2919393052, deleted): the literal job token is the `origin` remote URL's password (one file under `$NS`, no other copy); synthetic username-only/password controls proved the classifier; `RUNNER_SCRIPT_TIMEOUT` below the job timeout yields an authoritative failed job **and** a running `after_script`.
- **Codex dispositions (P0 review):** §1 credential removal before any mount with a fail-closed proof; §2 timeout design accepted; §3 provisional 35 GiB hosted start floor for integration/full only inside the qualified adapter context, a bounded disk guard and the durable floor `max(35 GiB, ceil(1.5 × incremental_peak + 4 GiB))`.

### 15.3 Implementation (R3) and files

New: `infra/ci/select-profile.sh` (`be30a7e7`), `infra/ci/profile-paths.tsv` (`03bed7ec`), `infra/ci/hosted-context.sh` (`7bb49340`), `infra/docker/ci-context.Dockerfile` (`180973f9`). Changed: `.gitlab-ci.yml` (`22ee5f1b`), `infra/ci/validate.sh` (`12ecc414`), `infra/ci/lib.sh` (`6d550a4a`), `infra/ci/runner-preflight.sh` (`ad58dc3f`), `infra/ci/validate-controls.sh` (`fedd2885`), `infra/validation/run.sh` (`8c049f66`), `infra/validation/lib.sh` (`8663e1d8`), `infra/validation/required-checks.tsv` (`23f7385b`), `README.md` (`c19aa9a8`), `docs/foundation/OPERATIONS.md` (`abf3d260`). Unchanged as planned: application code, schemas, migrations, Compose, service Dockerfiles, dependencies, `cleanup-owned.sh`, `run-approved.sh`.

1. **Profiles and suites.** `run.sh` suites `lint controls provenance contracts unit root migrations ops rename off` (aliases `static`, `variants`, `all`; deduplicated; per-suite and per-check `timings.tsv`); new **B-UNIT-CHECKS** (README check block without a stack, generated environment file deleted after use); `required-checks.tsv` relabelled plus B-UNIT-CHECKS (70 rows): **fast 31, integration 63, full 69** IDs (full unchanged).
2. **Selection.** `select-profile.sh` with the reviewed map (docs → fast, runtime → integration, foundation → full; exact file, then longest prefix; maximum class). Fail-safe to full on unknown/unsafe paths, deletion/rename, a missing/zero/absent/non-ancestor/degenerate base and (after review) a failed diff or sort (`diff-failed`). Merged-result, schedule and web pipelines are full without a diff; feature pushes are fast with no gate; an empty gate (`harness=not-required`) only when the pipeline profile is fast.
3. **validate.sh / lib.sh / preflight.** `--profile` (CI: `fast`|`gate`, default full; local: `fast`|`integration`|`full`); new tested contents `branch-schedule`, `branch-web`; feature pushes attributed (`branch-push`); profile-specific completeness (exactly the profile's IDs, out-of-profile rows rejected, profile reports); budgets `min(job, script) − elapsed − reserve` with caps and minimums; capacity per profile with the 35 GiB hosted floor active only in the verified adapter context; selection and timings published; summary fields `profile`, `pipeline_profile`, `executed_profile`, `selection_reason`, `comparison_base`, `required_count`, `required_set_sha256`.
4. **Hosted adapter.** Credential removal from remote URLs (fail-closed parser), include classification, a regular-file scan of `$NS` with error detection, an unprivileged context container (exact mounts, runner `<project>.tmp` masked, allow-listed environment, `no-new-privileges`), the disk guard (time-limited samples, stall and death detection, failure kept in memory and on disk, separate-filesystem sampler container) and the capacity report; `after_script` backstop after the same proof.
5. **YAML.** Jobs `fast` and `gate` on `saas-linux-medium-amd64` with the pinned CLI image and TLS DinD service; six workflow rules (open-MR feature-push suppression); gate `needs: [{job: fast, artifacts: false}]` (cycle 3); B-CI-CONFIG asserts the definition exactly, README blocks and profile/suite consistency.
6. **Controls: 89 → 195** (`--list`): VP1–VP9, P6–P8, N9/N12 redefined, N17, N18, V36, V37, 68 selector controls (group F, incl. the governance, status and merged-result scenarios and F-diff-fail), 22 adapter controls (group G: G1–G5, G6a–G6h credentials, G7a–G7i disk guard). As root (the hosted context) validate.sh launches drop `CAP_DAC_OVERRIDE`/`CAP_DAC_READ_SEARCH` (setpriv) so the file-mode controls keep their meaning; group G containers run with the same capabilities dropped.

### 15.4 Deviations and disclosures

1. **Root in the hosted context (found during R3).** The context runs as root with DAC override; V6d, V17 and V17b depend on permission denial. Fixed in the controls (setpriv capability drop), not by changing the qualified context privileges.
2. **Adapter test seam without a fault switch.** Busybox ash runs `df` as a built-in applet, so a PATH fixture cannot replace it; group G uses a size-limited tmpfs as the build directory (real `df` figures), a docker shim on PATH and kill/stop of the identified guard process.
3. **Separate-filesystem sampling** uses one long-lived labelled sampler container started before the workload, so no transient guard container can appear in the harness's leftover inventory (a planned one-shot sampler would have risked a spurious 14).
4. **Summary `profile`** records the requested job kind (`gate` in the gate job); the documented reuse rule therefore keys on `pipeline_profile` and `executed_profile` both `full` (Plan r2 §6.4 wording "profile = executed_profile = full").
5. **B-CI-CONFIG** drops the old "runner tag not repeated in docs/infra" rule (the hosted tag is public and documented) and derives the timeout bound from the definition's own values.
6. **Read-only review findings** (one confirmed fail-open, the rest plausible) were fixed or documented in cycle 1 (§15.5); finding 4 (a credential URL in a section header or an include target is refused rather than cleaned) is kept fail-closed, since P0 found no such case.
7. **Trust boundary (documented):** GitLab passes `CI_JOB_TOKEN` to the DinD service's environment; repository code is trusted with the token as in any job script. Selection trusts predefined variables; a pipeline-variable override (developer role today) can narrow a push pipeline's profile — tied to the owner setting.
8. **GitLab's own runner** prints a truncated job-token prefix in its artifact-upload log line (standard runner output, not ours; the token expires with the job).

### 15.5 Validation (every run, failures included)

| Step | What | Outcome |
|---|---|---|
| Dev | B-CI-CONFIG isolated + 9 YAML mutations; controls F 65, G 20, A–D 83, E 22 | all pass; every mutation detected |
| Review | one read-only CI/security reviewer (subagent, no edits) | 1 confirmed fail-open (selector diff/sort status), busybox `grep -r` unreadable directory → exit 1 and FIFO hang (both verified), plus plausible items |
| Cycle 1 | review fixes + 5 new controls; dev F+G | 90/90 PASS |
| L1 | wrapper `lint provenance contracts unit`, run `2525e8` | **FAIL**: B-SECRET-CANARY (generated secrets in `$WORK/unit.env`, not on the canary's name list) → B-REQUIRED-CHECKS |
| Cycle 2 | B-UNIT-CHECKS deletes its environment file after use | — |
| L1 | run `db88f0` (`l1-r2`) | wrapper 0, **31 PASS**, 6 INFO, 0 FAIL |
| L2 | `validate-controls.sh --group A,B,C,D,F,G` | **173/173 PASS**, repository status unchanged |
| L3 | scratch-commit copy, `env -u GITLAB_CI validate.sh --profile fast` | **exit 0**, `tested_content=local`, profile fast/fast, 31 PASS, repository status unchanged, copy deleted |
| H1 | MR !1 pipeline 2919757084 on `5dd8d204` | fast **success** (295 s, 31 PASS, published full, guard ok); gate **failed in 40 s, exit 16**: the gate's `needs` downloaded `ci-evidence/` into its checkout |
| H-INT (1) | MR !2 pipeline 2919774128 | fast success with `pipeline=integration reason=runtime base=B`; gate failed as H1 |
| H2 (1) | push pipeline 2919774328 on `8e0deced` | fast only, B-CONTRACT-GO/PY/AGREE FAIL, exit 1 (as intended); superseded |
| Cycle 3 | gate `needs: {job: fast, artifacts: false}`; B-CI-CONFIG positive/negative | PASS / FAIL as intended |
| L1 | run `57b3e5` (`l1-r3`) | wrapper 0, 31 PASS, 0 FAIL |
| **H1′** | MR !1 pipeline 2919793632 on `110e00c7` (fast-forward; tree `b0c468c6`); no branch pipeline was created (open-MR duplicate suppression observed) | fast **success** (335 s; 31 PASS; `pipeline=full executed=fast`; published full); gate: **full executed** (`foundation`, base `19f945a`, 69 required IDs), **67 PASS, B-TOGGLE-OFF FAIL → B-REQUIRED-CHECKS FAIL**, harness 1, complete 14, published full, job **failed** (2582 s) |
| **H-INT** | MR !3 int-head `f17e4756` (tree `b0c468c6` = H1′) → int-base `b38b1f4e`; pipeline 2919806781 | fast success (213 s); gate: **`pipeline=integration executed=integration`, reason `runtime`, base B, `selection.tsv` = `M runtime web/app/page.tsx`, 63 required IDs**; **61 PASS, B-TOGGLE-OFF FAIL** → job failed (2037 s) |
| **H2** | push `g003-gate-proof` `6feb55b8` (H1′ + EI39 valid); pipeline 2919806822 | **as intended:** fast only (`branch-push`, `feature-push`), B-CONTRACT-GO/PY/AGREE FAIL, `validate.sh` final status 1, published full (20 files), job **failed** exit 1 (295 s) |

**The single hosted failure, B-TOGGLE-OFF (H1′ full and H-INT integration).** The published reason is the check's generic "intelligence-disabled variant"; the specific assertion is in the harness console, which validate.sh never publishes (by design). Every other check of the off variant passed in both runs (B-TOGGLE-OFF-ADDITIONS, B-OBS-OFF, B-BROWSER-OFF 103 tests), so the disabled stack rendered, started, served and was probed. `variants_off` is unchanged since the last fully passing local run (`3c8cd5`) and by R3. **Most likely cause (inferred, not proven):** the control assertion `infra/validation/run.sh:1945` requires `python:3.13.16-slim` in the daemon's *image store* (`docker image inspect … | grep PYTHON_VERSION`). On the fresh hosted daemon (Engine 29, classic overlay2 store) the intelligence build pulls that base through BuildKit's build cache, which does not create an image-store entry; locally the image is already present (pulled 5 days ago), so the control passes. This is an inherited G001/G002 check exposed by the first run on a fresh daemon, not an R3 regression. **Proposed minimal fix (not implemented; needs a cycle):** make the control self-sufficient, for example `docker image inspect python:3.13.16-slim >/dev/null 2>&1 || docker pull python:3.13.16-slim` immediately before it (a pinned public base image; image rows are outside the leftover rule), or assert on the root intelligence image built earlier, which inherits `PYTHON_VERSION`. Optionally confirm first with a one-line diagnostic. A new H1 (and H-INT) is then needed on the exact corrected tree.

**Hosted measurements (cold medium VM, runner 55314121, `saas-linux-medium-amd64`, Docker 29.8.2 DinD over TLS).**

| Run | Harness | Suite timings (s) | Disk (shared filesystem, 45.0 GiB) |
|---|---|---|---|
| fast (H1′) | ≈ 4.8 min | lint 77–112, provenance 9–13, contracts 37–57, unit 33–77 (fast jobs 213–335 s in total) | baseline 37.28 GiB; incremental peak **2.67 GB (2.49 GiB)**; min 34.79 GiB |
| integration (H-INT) | 33 min | lint 89, provenance 13, contracts 57, unit 77, root 956, ops 207, rename 60, off 507 | peak **12.09 GB (11.26 GiB)**; min 26.02 GiB |
| full (H1′) | 42 min | lint 77, controls 735, provenance 11, contracts 37, root 919, migrations 36, ops 197, rename 54, off 450 | peak **12.32 GB (11.47 GiB)**; min **25.81 GiB**; 86 samples |

Durable start floor (P0 review §3): `max(35, ceil(1.5 × 11.47 + 4))` = `max(35, 22)` = **35 GiB** for full and integration; every hosted baseline (37.28 GiB) meets it; the guard never fired. Fast's 15 GiB provisional floor is far above its measured peak (1.5 × 2.49 + 4 ≈ 7.7 GiB). The run as root inside the context also qualified the controls: **B-CI-CONTROLS 195/195** (groups A–G, including Docker-real group E on DinD and the adapter group G), browsers 103/103 (root, off), identity 2/2 (rename), B-SECRET-CANARY over 73 secrets, `inventory-leftover.txt` only the permitted render-cache volume. Credential proof on every hosted job: `remotes=1 sanitized=1 includes=1 include_inside=1 … password_urls=0 token_files=0 scan=clean`; after_script re-proof `sanitized=0` (already clean).

Evidence digests (SHA-256, `$EXT/g003-exec/rev/`): `h1b/gate-artifacts.zip` `c474db78…0782` (results.tsv `230ee260…6b`, images.txt `9f310a72…ab`, required_set_sha256 `b6b662ab…268f`), `h1b/gate-trace.log` `f5c844e9…7e65`, `h1b/fast-artifacts.zip` `e45ea68a…e1b3`; `hint2/gate-artifacts.zip` `36d2782e…ea77`, `hint2/gate-trace.log` `6610a955…091a`; `h2b/artifacts.zip` `bec9d1c2…3c36`, `h2b/trace.log` `6e6c26ba…7820`; job list `hosted-jobs.tsv`.

### 15.6 Acceptance-criteria status (revision Spec)

| AC | Status | Evidence |
|---|---|---|
| AC7a profiles | **met** | fast 31 ⊂ integration 63, full 69 = inherited set; VP1–VP7; B-CI-CONFIG consistency; hosted summaries carry `required_count` and `required_set_sha256` |
| AC7b selection | **met (hosted selection evidence complete)** | group F 68 controls, P6–P8, N-set; H1′ foundation → full; H-INT runtime → integration (base B, one runtime row); H2 feature push → fast only; duplicate suppression observed (no branch pipeline for H1′) |
| AC7c hosted profile and adapter | **met for the profile; capacity disposition pending Codex** | P0; H1′ cold run: runner, tag, versions, credential proof, guard, peaks, cleanup; script-timeout after_script (P0 follow-up) |
| AC7d hosted final evidence | **NOT met** | fast passes on the exact tree; the full gate (H1′) and the integration gate (H-INT) each fail one inherited check (B-TOGGLE-OFF); H-INT construction verified (tree equality, `M web/app/page.tsx`, selected = executed = integration) |
| AC7e real failing fast gate | **met** | H2: EI39 → GO/PY/AGREE FAIL, exit 1, safe artifacts, branch deleted |
| AC7f timings, sampler, docs, reuse, cost | **met; target dispositions for Codex** | measured durations and peaks above; OPERATIONS profiles, commands, reuse rule, cost, owner settings |
| AC7g scope and binding | **met** | allow-list append only; wrapper pre/postchecks (`l1-r3` wrapper 0); protected inputs unchanged; this section appended only; no G048, OS001, product or deploy work |

### 15.7 Settings prerequisites and cleanup

- **Remote:** kept by design — branch `g003-foundation-observability-ci` (`110e00c7`) and **draft MR !1** to `staging` (open; nothing merged). Closed: MR !2 and !3. Deleted after evidence capture: `g003-evidence/int-base`, `g003-evidence/int-head` (both attempts), `g003-gate-proof` (both attempts); earlier probe branches already deleted. No push to `staging`/`main`, no merge, no tag, schedule, variable, runner or setting change.
- **Local:** no labelled containers, volumes or networks; scratch directories removed; the disposable clones `$EXT/work/h1-clone` and `p0-clone` (outside both checkouts) are kept for a corrected re-run; nothing pulled was left (an accidental unpinned `yq:4` pull and a `buildpack-deps` digest pull during development were removed immediately).
- **Owner settings, unchanged and still separate decisions:** "Pipelines must succeed" (off), pipeline-variable override role (developer; selection depends on it, §15.4 item 7), `staging` protection, a weekly default-branch schedule, artifact retention beyond 90 days.

### 15.8 Budget

**B3: 3 of 3 meaningful cycles used; checkpoint after cycle 2 performed** (BUDGET_POLICY §7: converging, distinct failures, Goal bounded). Cycle 1 = review-driven fixes; cycle 2 = the L1 canary; cycle 3 = the hosted artifact propagation. The implementation with passing development runs, the review, P0, successful runs and evidence retrieval cost no cycle (Plan r2 §10.1). Original **B7 6/7** and **correction pass 1 0/3** unchanged. Compute: revision hosted jobs 6379 runner-seconds ≈ **213 compute minutes** (medium factor 2; `hosted-jobs.tsv`), plus P0/P0-follow-up ≈ 10; within the purchased allowance.

### 15.9 Status

- **Stopped at a budget condition (B3 exhausted) — returned to Codex.** The CI revision is implemented and locally validated (L1 31/31, L2 173/173, L3 exit 0), reviewed, and qualified on the hosted medium runner: adapter, credential removal, disk guard, selection, the fast profile and 195 controls all pass hosted. The hosted full and integration gates each fail exactly one inherited check, B-TOGGLE-OFF, most likely because its control requires a base image in the image store of a fresh daemon. Fixing it needs a fourth cycle, which B3 does not allow; no self-extension.
- **Decisions for Codex:** (1) a bounded correction (one cycle) for the B-TOGGLE-OFF control as proposed in §15.5, then a new H1 and H-INT on the exact tree; (2) the capacity disposition (measured full peak 11.47 GiB → durable floor 35 GiB; baseline 37.28 GiB); (3) the summary `profile` field semantics for the reuse rule (§15.4 item 4); (4) the owner settings above.
- **AC7d is not met; G003 is not complete and not accepted.** The tree is uncommitted in the execution worktree (the hosted evidence commits exist only on the disposable branch). OS001 remains paused by the owner. The G021 foundation barrier remains closed.

## 16. CI revision correction: B-TOGGLE-OFF (budget extension; 2026-10-06/07)

Sections 1–15 remain unchanged as history (Result SHA-256 before this append `f98dcf00a02a7cafcafc74a71b848588770983bf6e3f214e7a43cdff9afc4d28`); this section appends the bounded correction. Where it differs, this section is current. Evidence: `$EXT/g003-exec/rev/c2/` and the per-run directories named below.

### 16.1 Authorization and starting state

| Item | Value |
|---|---|
| Authorization 1 | Codex, at the owner's explicit request: revision budget B3 (3/3) → **B5** (maximum 5, 3 used), convergence checkpoint after cycle 4; bounded correction scope (prove the failing B-TOGGLE-OFF assertion, smallest self-sufficient control correction, fail closed, no removal or skip of the detection control); allowed files `infra/validation/run.sh`, relevant control tests only if necessary, OPERATIONS only for a resulting inaccuracy; Result append. Record: main-checkout Goal supplement `954d0c1a…f0f` (read-only, not copied) |
| Authorization 2 | Codex-approved custom B5 override, at the owner's explicit request: **maximum 6** shared meaningful cycles (was 5), no reset; cycle-4 checkpoint and all scope/stop rules retained. Record: main-checkout Goal `5bf38bc7…6c16` ("Owner-approved six-cycle maximum") |
| Starting state | `g003-c2-precheck.sh` (SHA-256 `4eb1ccee…`), before any corrective write and again immediately before it (`c2/precheck-0.txt`, `precheck-1.txt`): binding `acae154f…` (9 inputs) and the original 5 protected inputs verified against the checkout; HEAD `cbffd6d`; no ignored files; the 585 Git-visible files equal the submitted H1′ commit tree `b0c468c6` byte- and mode-exactly; Result `f98dcf00…`; 200 changed paths; no `.env`, labelled or root-project resources; default ports free — **PASS** |

### 16.2 Diagnosis (proof, not a cycle)

A bounded cold-daemon diagnostic on the disposable branch `g003-evidence/diag-toggle-off` (commit `473072ef` = H1′ plus instrumentation only — each B-TOGGLE-OFF `ok=0` additionally logs its own `run.sh` line, no logic change — a probe-only CI job through a copy of the adapter differing in one line (the context runs the diagnostic instead of `validate.sh`), no artifacts; pipeline 2919924280, job 16984604253, 1807 s; branch deleted):

- Daemon 29.8.2, `overlay2` classic store; `python:3.13.16-slim` **absent** from the image store at start and **still absent after `root` built the intelligence image** (BuildKit keeps base images in its build cache).
- Instrumented `root off`: every check PASS except B-TOGGLE-OFF (→ B-REQUIRED-CHECKS); **exactly one failed assertion: `run.sh:1945`**, message "control: python base image not detectable by PYTHON_VERSION".
- The unchanged line-1945 command: exit 1 cold; after an explicit `docker pull` exit 0 — the detector still detects its known positive control.
- Evidence: `c2/diag-trace.log` `0860c032…b36c`, `c2/diag-lines.txt` `1d3a3b8b…9888`.

The §15.5 hypothesis is therefore proven for the cold daemon; the earlier hosted B-TOGGLE-OFF failures had no other failing assertion.

### 16.3 Correction (cycle 4)

`infra/validation/run.sh` only: a helper `control_image IMAGE` (image already in the store → `present`; otherwise a pull of the pinned image → `pulled`; otherwise `unavailable`, nonzero) runs before the **unchanged** PYTHON_VERSION detector assertion; an unavailable image still fails the check (fail closed). The detector state appears in the B-TOGGLE-OFF result line (`… the PYTHON_VERSION detector control holds (python:3.13.16-slim present|pulled)`; on failure `intelligence-disabled variant (detector control image …)`). Every disabled-intelligence assertion (absence of the service, Python images and containers, AI/intelligence configuration and files; core readiness; migrations; start/check/stop; browsers) is unchanged. OPERATIONS does not describe this control and is unchanged. A pulled pinned base image is left in place, like the harness's other pinned tool images.

### 16.4 Validation (every run)

| Step | What | Outcome |
|---|---|---|
| Unit | the real `control_image` (extracted from `run.sh`) on the local daemon | `python:3.13.16-slim` → present (0); `busybox:1.36.1` (absent) → pulled (0, then removed: created by the test); `python:0.0.0-g003-absent` → unavailable (1, fail closed); detector: PYTHON_VERSION found on `python:3.13.16-slim`, not on busybox (`c2/control-image-unit.txt`) |
| Local (image available) | wrapper `lint provenance contracts unit off`, run `7912fc` (`c2-local`) | wrapper 0, postcheck passed, **35 PASS**, 0 FAIL; **B-TOGGLE-OFF PASS "(python:3.13.16-slim present)"**; B-REQUIRED-CHECKS 34 IDs |
| **H1″** | `5ec78dec` (fast-forward on the H1 branch, tree `4dcdc044`; only `infra/validation/run.sh` changed); MR !1 pipeline 2920018355 (no branch pipeline: open-MR suppression) | **SUCCESS.** fast 31/31 PASS (301 s; `pipeline=full executed=fast`); gate **full: 69 PASS, 0 FAIL** (2478 s; `foundation`, base `19f945a`; required_count 69, `required_set_sha256` `b6b662ab…268f`); **B-TOGGLE-OFF PASS "(python:3.13.16-slim pulled)"** — the absent-image path on a cold daemon; `validate.sh` final status 0, published full |
| **H-INT** | MR !4 int-head `1b5ac82d` (tree `4dcdc044` = H1″) → int-base `f52f306f` (one never-executed line); pipeline 2920020748 | **SUCCESS.** fast 31/31; gate **`pipeline=integration executed=integration`, reason `runtime`, base B, `selection.tsv` = `M runtime web/app/page.tsx`, 63 PASS, 0 FAIL** (1783 s); B-TOGGLE-OFF "(pulled)"; final status 0; MR closed, both branches deleted |
| **H2** | push `g003-gate-proof` `8d5de574` (H1″ + EI39 valid); pipeline 2920020849 | **as intended:** fast only (`branch-push`, `feature-push`, 31 IDs), B-CONTRACT-GO/PY/AGREE FAIL, `validate.sh` final status 1, published full, job failed exit 1 (336 s); artifacts `b74c2aa5…aa34`; branch deleted. Rerun because the tree changed (reuse rule: tree SHA) |

**Hosted evidence on the exact corrected tree (`4dcdc044`).** Every hosted job proved the credential boundary (`remotes=1 sanitized=1 … password_urls=0 token_files=0 scan=clean`; after_script re-proof `sanitized=0`) and ended `guard=ok`. Full gate: B-CI-CONTROLS **195/195** as root in the context, browsers 103/103 (root and off), B-SECRET-CANARY over 73 generated secrets, `inventory-leftover.txt` only the permitted render-cache volume; suite timings (s) lint 76, controls 724, provenance 9, contracts 38, root 878, migrations 37, ops 184, rename 51, off 421 (harness ≈ 40 min); disk baseline 37.28 GiB, minimum 25.84 GiB, **incremental peak 12.28 GB (11.44 GiB)**, durable floor `max(35, ceil(1.5 × 11.44 + 4))` = **35 GiB**, met. Integration: timings lint 76, provenance 10, contracts 45, unit 63, root 851, ops 191, rename 54, off 435 (≈ 29 min); peak 12.07 GB (11.24 GiB), floor 35 GiB, met. Fast (H1″): 301 s job, peak ≈ 2.5 GiB.

**Reuse rule.** The H1″ full pass is reusable for this tree: job status success, `pipeline_profile` and `executed_profile` both `full`, tree `4dcdc044`, `required_set_sha256` `b6b662ab…268f`, `images.txt` `9f310a72…abc56`. H2 was rerun on H1″ (`8d5de574`, tree `8d7954b8`) because the final tree changed; its intended failure proof is unweakened.

**Evidence digests** (SHA-256; `$EXT/g003-exec/rev/`): `h1c/gate-artifacts.zip` `49303415…45f1` (results.tsv `1ac24c13…5ade`), `h1c/gate-trace.log` `7adedc1e…5bcd`, `h1c/fast-artifacts.zip` `da3905de…48b8` (results.tsv `1133d42f…3b63`); `hint3/gate-artifacts.zip` `7a731aec…27c5` (results.tsv `af532859…4217`, `required_set_sha256` `47ae36df…89eb`), `hint3/gate-trace.log` `a8d186e4…965b`; `h2c/artifacts.zip` `b74c2aa5…aa34`, `h2c/trace.log` `16f750d8…735d`; job list `c2/hosted-jobs-c2.tsv`.

### 16.5 Budget

**Aggregate revision count: 4 of 6 shared meaningful cycles used** (custom B5 override, maximum 6; authorizations in §16.1). Cycle 1 = review-driven fixes; cycle 2 = the L1 canary; cycle 3 = the hosted artifact propagation; **cycle 4 = this correction** (corrective implementation after the proven hosted failure, then local and hosted retest — passed). The cold-daemon diagnostic, the precheck, successful runs and evidence retrieval cost no cycle. **Convergence checkpoint after cycle 4 (BUDGET_POLICY §7):** converged — the single proven failing assertion is corrected, every hosted gate is green on the exact tree, no repeated failure; cycle 5 not needed and not used. The original **B7 6/7** and **correction pass 1 0/3** are unchanged. Compute in this pass: 6942 runner-seconds ≈ **231 compute minutes** (diagnostic 1807 s; H1″ 2779 s; H-INT 2020 s; H2 336 s); revision total ≈ 213 + 231 + 10 (P0) ≈ 454.

### 16.6 Acceptance-criteria status and remaining decisions

| AC | Status after this correction | Evidence |
|---|---|---|
| AC7a profiles | met | fast 31, integration 63, full 69 in hosted summaries with their `required_set_sha256` |
| AC7b selection | met | H1″ foundation → full; H-INT runtime → integration (base B, one runtime row); H2 feature push → fast only; open-MR suppression observed |
| AC7c hosted profile and adapter | met for the profile; **capacity disposition is Codex's** | three hosted full/integration runs on the corrected and previous trees: peaks 11.24–11.47 GiB → durable floor 35 GiB, baselines 37.28 GiB |
| AC7d hosted final evidence | **met on the exact corrected tree** | H1″ fast and full PASS; H-INT integration PASS with tree equality and the runtime-only diff |
| AC7e failing fast gate | met | H2 on H1″ |
| AC7f timings, sampler, docs, reuse, cost | met; target dispositions are Codex's | timings and peaks above; reuse rule satisfied for H1″ |
| AC7g scope and binding | met | precheck before writes; only `infra/validation/run.sh` corrected; Result appended only; no G048, OS001, product, deploy, settings or main/staging work |

**Remaining decisions (not part of this correction):** Codex's G003 acceptance; the hosted capacity and timing dispositions (provisional 35 GiB floor; measured full peak ≈ 11.4 GiB); the summary `profile` field and reuse-rule wording (§15.4 item 4); the owner project settings (§15.7).

### 16.7 Cleanup and status

- **Remote:** branch `g003-foundation-observability-ci` at `5ec78dec` with **draft MR !1** to `staging` kept open by design (nothing merged). MR !4 closed; `g003-evidence/int-base`, `int-head`, `g003-gate-proof` and `g003-evidence/diag-toggle-off` deleted after evidence capture. No push to `staging`/`main`, merge, tag, schedule, variable, runner or setting change.
- **Local:** no labelled containers, volumes or networks; a test-created `busybox:1.36.1` was removed; the disposable clone `$EXT/work/h1-clone` (outside both checkouts) is kept.
- **Status:** the G003 CI revision's hosted evidence is complete and passing on the exact corrected tree. **G003 is not accepted** — returned to Codex for review. The tree remains uncommitted in the execution worktree (the hosted commits exist only on the evidence branch). OS001 remains paused by the owner; the G021 foundation barrier remains closed.
