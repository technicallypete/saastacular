> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G003 — Health, logging, contracts, and baseline CI work specification

Goal: [G003-foundation-observability-ci.md](../goals/G003-foundation-observability-ci.md)
Epic: EPIC-001 — Repository and local platform
State: Authorized under the owner-requested sequential execution addendum; gates apply

## Owner-authorized sequential execution override — 2026-10-05

For the owner-requested combined G047 → G003 batch, follow the supervisor execution authorization (not included). Fully validated G047 with a submitted Result is the phase-transition prerequisite; Codex reviews both implementations afterward. The contract supplies a deterministic committed-baseline transition and exact supervisor-created inputs. G003 local implementation may proceed while runner access is investigated, but hosted qualification/security checks and all required hosted evidence remain mandatory. This overrides earlier execution-order and planning-only wording; no requirement, acceptance criterion or protection is removed.

## Goal, Rationale and Authority

Complete the repository's remaining baseline operational guarantees: correlated privacy-safe diagnostics on existing service calls, cross-language contract conformance, and an attributable GitLab pipeline that runs the containerized checks and fails when required checks fail. Preserve accepted behavior rather than rebuilding the health or web foundation.

Follow AGENTS.md, CLAUDE.md, ARCHITECTURE.md, accepted ADRs, ROADMAP.md and BUDGET_POLICY.md. This Spec defines WHAT must be true; Claude's reviewed Plan selects modules, validators, dependencies, job arrangement and commands. Architecture and security conflicts require escalation. No product-specific feature work is allowed until G000–G021 are accepted and G021 releases the barrier.

## Inspected Baseline and Prerequisites

**Added execution prerequisite (2026-10-05):** [G047 successor Vestive shell refresh](G047-vestive-shell-refresh-spec.md) must be accepted before G003 implementation. G003 Plan corrections may continue, but final approval must reconcile its protected baseline, browser/provenance assumptions and feasibility to the accepted refresh and bind fresh exact inputs. This does not add UI implementation to G003 or weaken any G003 criterion, hosted evidence obligation or existing prerequisite. The baseline below describes the original inspection.

Baseline inspected on 2026-10-05: `7193ed84237be9362b7b596bdd135dd592447b07` (`G002: accessible web shell (accepted)`). Both ordinary and ignored-inclusive status inventories were empty before drafting. G001 and G002 acceptance is committed in PROGRESS.md; their historical submission wording and old snapshots remain historical.

| Area | Accepted/current state | G003 increment |
|---|---|---|
| Health | Core `/healthz`, `/readyz`; web `/api/health`, `/api/ready`; intelligence `/healthz`, `/readyz`. Core checks PostgreSQL, Redis and object storage; intelligence is optional. Web readiness has a 2s core-fetch bound. | Preserve semantics; prove bounded responses and propagation on actual HTTP calls. |
| Correlation | Core accepts bounded request IDs and extracts/generates trace IDs. Web core-readiness fetch and core intelligence-health probe do not forward context. | Consistent diagnostic context with inspectable parent relationships across the existing health chain. |
| Logs/privacy | Core uses a redacting JSON logger. Intelligence has JSON execution export with content off by default. No corresponding web completion logger or correlated intelligence health access log. | Service request-completion records, error-path privacy and canary evidence. |
| Contracts | Two JSON Schema Draft 2020-12 intelligence contracts; Python loads/validates them and tests all eight event types. Go has no independent conformance corpus. | Shared valid/invalid corpus checked independently by Go and Python; fake producer-output validation. |
| Checks | Docker build/check targets, project harness static/provenance/root/migrations/variants, G002 root/disabled browser matrices and dependency/wrapper controls. | Reuse checks in a fail-closed CI entry point; add only missing operational/contract assertions. |
| Hosting | `origin` is `git@gitlab.example:<namespace>/<project>.git`; no CI definition exists. The owner confirms use of an existing Docker-capable GitLab runner. | GitLab CI, with runner-profile verification and attributable hosted evidence. |

Roadmap prerequisite G001 is satisfied. Accepted G002 supplies the current consumer baseline and mandatory regression coverage. Before execution, the approved Plan must name the actual starting commit (including any committed planning artifacts), runner profile/capacity, exact approved-input binding and exact scope allow-list. Uncommitted supervisor/Plan inputs require exact hashes and inspected ownership; no directory-wide management exception.

## Scope

In scope: existing health routes and their diagnostic headers; bounded HTTP correlation/logging in web/core/intelligence; independent conformance tests for existing contracts; containerized CI commands, GitLab workflow and negative controls; project-owned harness/provenance reconciliation and focused operational documentation.

Out of scope: new application routes, domain/authentication/MCP contracts, production AI execution clients, tool/callback execution, tenant authorization/schema, audit or usage ledgers, metrics exporters/endpoints, telemetry collectors/vendors, external trace export, dashboards/alerting, deployment/publishing, runner fleet provisioning, vendor replacement, unrelated refactors/upgrades, G002 visual/accessibility follow-up implementation, and any product feature. No new migrations, seeds or durable product tables.

## Requirements

### R1 — Preserve Health Semantics and Bound Failure

Retain the existing endpoint paths and JSON meanings. Core liveness reports `alive`; core readiness reports `ready`/`unready`, configured `product_name`, required dependency states under `checks` and optional intelligence under `optional`. Web liveness reports `alive`; web readiness reports `ready`/`unready` with core `ok`/`unavailable`. Intelligence keeps its existing health `ok` and readiness `ready` meanings; an unconfigured model/provider is not a readiness failure.

Core's required checks remain PostgreSQL, Redis and `object_store`; intelligence failure never gates core readiness. With intelligence disabled, no optional intelligence entry/probe is present. Web readiness depends on core; web liveness does not. Healthy readiness is HTTP 200; required-dependency or core unavailability is HTTP 503. Core/web retain `Cache-Control: no-store`; intelligence health/readiness must also be no-store. Responses carry no errors, DSNs, internal URLs or credentials.

Preserve the existing 2s probe/fetch deadlines and cancellation. An unresponsive upstream must not turn a health response into an unbounded wait. In a settled running service, injected timeout/outage cases must return within 5s measured by a monotonic client clock, or fail the criterion. Startup/recovery polling has a separately documented finite deadline and cannot be used to conceal a single slow request. Logging must not change response status/body or duplicate-write behavior.

### R2 — Diagnostic Context Contract

Use the existing `X-Request-Id` and `traceparent` conventions. Request IDs are diagnostic strings matching `^[A-Za-z0-9._-]{8,64}$`; accept a single valid inbound value, otherwise generate a new bounded opaque ID. Return the effective request ID on each scoped health response. Trace/span IDs are operational infrastructure identifiers, not domain UUIDs, authorization references or persisted identities.

The minimum supported trace profile is W3C `traceparent` version `00`: lowercase hex version/trace ID/parent ID/flags, nonzero 32-hex trace ID and nonzero 16-hex parent ID, with version-specific length/flag handling. Missing, malformed, oversized, duplicated or unsupported values must produce a new valid local context without reflecting the invalid value, raising HTTP 500 or logging it raw. The Plan must specify accepted flags and unsupported-version behavior using the [W3C Trace Context recommendation](https://www.w3.org/TR/2021/REC-trace-context-1-20211123/); do not claim complete W3C compliance for a narrower profile.

Continue a valid trace ID; create nonzero per-hop span identifiers and document their parent relationship. Do not equate copying the same incoming header at every hop with participating in a trace. Request-local state must be isolated under concurrent requests and reset after failures/cancellation. Do not forward arbitrary inbound headers, cookies, authorization, `baggage` or vendor `tracestate`; this bounded foundation transports diagnostic context only.

Neither diagnostic ID nor sampling flag confers authority, selects a tenant, permits content capture or enables external export. IDs and trace fields are untrusted inputs; logging limits/redaction still apply. Authenticated intelligence execution retains its existing envelope attribution and authentication semantics; the health correlation mechanism must not overwrite/widen the envelope or require a production Go execution path.

### R3 — Actual Service Continuity Without New Product Paths

Demonstrate on the built stack: web `GET /api/ready` → core `GET /readyz` → the existing optional intelligence `GET /healthz` probe. The same effective request ID and trace ID must be visible in sanitized completion records across all participating hops. Record distinct valid span IDs and an inspectable parent chain consistent with the Plan's documented server/client-span convention. Datastore probes need no invented HTTP tracing protocol.

Cover a supplied valid context, no supplied context, invalid inputs replaced, simultaneous distinct requests, required datastore outage, optional intelligence outage/recovery, core timeout/unavailability and the intelligence-disabled variant. Failures must retain the caller's effective context in the services that actually received the request; do not claim a log from an unreachable service. Health body/status remains R1. Existing network boundaries and internal DNS remain unchanged; web never calls intelligence/datastores directly.

No new demo route, execution endpoint, public streaming edge, synthetic authenticator or model invocation may be used to manufacture continuity evidence. Existing authenticated intelligence tests may use fake gateways in a dedicated test process; no such fixture or test configuration may enter the production image/runtime authority path.

### R4 — Structured Logs and Privacy

Application-owned startup/error records and request-completion records are one JSON object per line to the existing service output. For each request to a scoped health route, require exactly one completion record from that service, including on handled failures. A request-completion record identifies service, timestamp, severity, event, method, normalized route, final HTTP status, nonnegative duration in milliseconds, effective request ID, trace ID and active span ID; parent attribution is present when a valid parent exists. The Plan defines actual key names/types and distinguishes completion from optional startup/error records. Test actual implicit-200 and error-status accounting.

Scope HTTP completion logging to the existing operational routes and any existing intelligence request/error logging being changed; G003 need not introduce a framework-wide logging abstraction for the UI. Framework access/error output for tested requests must either obey the privacy/structure contract or be suppressed in favor of the correlated records; no second raw access line or traceback may leak inputs. Build/test CLI output need not be JSON.

Never record raw headers, cookies, query strings, request/response bodies, prompts/completions, retrieved content, complete envelopes, authorization references, tenant/user identifiers or credential-bearing URLs/DSNs. Errors use bounded error categories or exception types, not arbitrary exception text/stack payloads. Preserve core's sensitive-key, configured-secret-value and nested-field redaction; add equivalent protections wherever G003 introduces logging. Non-sensitive diagnostic metadata must remain observable: redacting/disabling all output is not conformance.

Intelligence's existing execution-export content capture remains off in production; inbound flags/environment cannot enable it. No external sink is introduced. Canary tests cover secret values and sensitive keys in message/field/nested/error paths and request header/query/body inputs, plus fake prompt/completion content. Verify their absence from captured raw logs and retained artifacts, not only formatted summaries. Positive controls prove the scanner detects a planted canary and that expected non-sensitive records are present. Known credentials supplied as diagnostic values must not escape the redaction boundary.

### R5 — Independent Contract Conformance

Keep `contracts/` as schemas/protocols only. The existing execution-envelope and stream-event schemas remain the canonical wire contracts; preserve their IDs, version, field/type/additional-property rules, UUID domain fields, deadline representation and all eight event variants. No duplicate handwritten schema becomes a new authority. No protocol broadening, restrictive wire change or new product DTO is authorized by this Spec.

Provide one attributable synthetic fixture corpus and machine-readable expected valid/invalid outcomes. An independent Go validator and the existing Python validator must agree with those expectations using the same canonical schema bytes, with Draft 2020-12 references and applicable format checks effective. No implementation may pass by reading the other validator's answers or silently treating unsupported schema keywords as success. Meta-validate schemas; missing, malformed or unsupported schema inputs fail with a nonzero check status. Use only local schema resolution during tests; no runtime remote schema download.

Required valid coverage: minimal and fully populated envelopes; representative boundary values for optional fields/collections; each of `text_delta`, `tool_call`, `tool_result`, `citation`, `structured_output`, `usage`, `error`, `done`; nullable usage values and optional usage units. Required invalid coverage: malformed JSON, missing required fields, wrong types, unknown fields/event kinds/capabilities, invalid UUID/request/trace identifiers, invalid calendar/deadline formats, prohibited fields and negative usage counts. Include each forbidden Organization/user/Group/BusinessEntity field and content field in usage events. Preserve existing wire behavior rather than importing traceparent's nonzero rule into the separate envelope schema without review.

Use test-only Go-produced envelope bytes and actual Python-emitted event bytes from the existing fake-gateway tests to prove producer output conforms and is independently accepted by the peer-language checker. Include corrupt/truncated NDJSON negative inputs so the check cannot pass after ignoring an invalid line. This authorizes tests/validation adapters, not a production core execution client, public streaming controller, callback system or usage persistence. Shape/format conformance is not authorization, tenancy isolation or proof that the entire future AI protocol is implemented.

### R6 — Containerized Commands and GitLab CI

Provide exact root commands for contract checks, operational/correlation checks and the CI validation entry point. Commands must work as written against the actual submitted repository using Docker/Compose and baseline host Git/shell utilities. Project Go/Python/Node/browser/datastore runtimes stay in containers; tests execute with installed dependencies without registry/model downloads. Registry access is allowed during controlled image/dependency setup. Pin added direct dependencies/tools; preserve current framework/runtime versions unless a concrete requirement and reviewed Plan authorize a bounded addition. Record lockfile/module attribution.

CI targets the existing GitLab host. Define merge-request validation and pushes to `staging`/the configured default branch, with no publishing/deployment job or external production account. Document branch/default-branch rules and the runner requirements. Job layout/parallelism belongs to the Plan; every required job/check must be gating, with finite timeouts and no allow-failure, success-forcing, broad skip or path rule that can omit a required check. Record checkout commit and whether merge-request source or merged-result content was tested; do not imply a merged-result run when only source content ran.

The pipeline must cover containerized builds and core/web/intelligence existing checks, static/provenance checks, independent contracts, new health/correlation/logging/privacy cases, migration/isolation/default-port/rename/persistence regressions and the full G002 browser matrices for enabled and disabled variants. Reuse the existing project harness, checks profile and scenario matrix. Preserve every accepted assertion; add G003 cases and explicitly review exact-path reconciliation rather than bypassing old guards. The current 50-PASS harness and 95-test browser matrices are baselines, not upper limits or a reason to freeze required new assertions.

Disabled-stack runtime tests must still omit intelligence/Python service images/configuration and AI credentials. A dedicated contract-validation tool using Python is permitted outside that stack and must be distinguished from a supposedly intelligence-disabled runtime. No live model/provider calls or gateway key is required anywhere.

Runner access must be limited to disposable job-owned resources, support Compose >=2.24 and the approved browser image, and have measured capacity/timeouts for a first build and integration run. No production Docker socket, privileged mode, shared developer resource access or runner security-model change is implicitly authorized. The Plan must resolve the execution profile and permissions before implementation; inability to run required checks is a blocker, not a green optional stage.

### R7 — Failure Gates and Evidence That Can Be Audited

Required local checks and the aggregate CI entry point exit nonzero when any required check fails or is unmet. Preserve the failing check's status through log collection and cleanup; a successful cleanup/postcheck cannot turn failure into success. Snapshot prechecks prevent normal work when reviewed inputs drift; postchecks always run and fail a previously successful run on drift. Reuse the existing wrapper/controls with a newly approved G003 snapshot and preserve their failure semantics. Old G002 snapshot bytes remain historical and must not be regenerated to conceal a change.

Committed disposable controls must prove, independently: (a) one service check failure gates the aggregate; (b) one invalid contract fixture gates it; (c) a planted secret is detected and gates artifact retention; (d) harness/check failure survives successful cleanup/postcheck; (e) successful checks followed by input drift fail; and (f) missing/skipped required report/scenario evidence fails. Use synthetic copies/stub commands for forced faults; production source and real credentials are never altered to plant failures.

Provide hosted GitLab evidence of the final passing required pipeline on the exact submitted commit/tree and of a disposable failing gate case. The failing hosted case must exercise the reviewed gate implementation rather than an unrelated job containing `exit 1`. No failure-injection switch remains enabled in the normal pipeline. Local controls cover the full negative matrix; hosted evidence must corroborate gate/job wiring and status propagation. Local YAML parsing or a local reproduction alone cannot satisfy hosted pipeline acceptance.

Artifacts identify commit/tree, run/pipeline/job identifiers, tool/image/schema versions, commands, per-check results, exit status, UTC times, test counts and skips, plus cleanup inventory. Record source actor/review attribution using existing Git/GitLab metadata; do not build an audit ledger. Collect safe status summaries even on failure, but scan raw captures before upload. Do not upload .env, credentials, Compose-resolved secret config, raw sensitive bodies or unsafe logs. A sanitization/scan failure prevents unsafe upload and fails the gate; public synthetic source fixtures containing no credentials may be retained. Evidence retention/access and artifact allow-list must be stated in the Plan. Avoid broad upload globs and privileged CI secrets for merge-request code.

### R8 — Consumer Provenance and Inherited Guarantees

Preserve both Skills, original `.archetype` manifest, canonical governance/architecture and G000–G002 historical artifacts. Register exact new consumer customizations and demonstrate repeat plan/render preserving all existing files with only the documented conflicts/proposals. Extend the intelligence-disabled consumer with the applicable operational changes; preserve its own manifest. Reconcile obsolete byte-equality source guards only for reviewed changed paths while preserving their semantic security assertions, negative controls and reason-code meanings.

Do not weaken G001 migration owner/runtime-role separation, owner-only metadata, unchanged normal migration set, independent datastore-port/IP isolation controls, liveness/readiness outage semantics, persistence/Redis ephemerality, runtime naming/default ports or dormant MCP isolation. Preserve G002 dependency-policy controls including justified `devOptional`, hash-pinned registry CSS and exact literal-scan exceptions, wrapper controls and first-paint evidence requirements. Changes to shared validation need affected static/control checks plus relevant integration/browser regressions in the same implementation cycle.

G002's fixed angle, destructive tint, helper-module additions and other accepted visual/testing deviations are protected accepted behavior, not cleanup targets. Its text-spacing automation, locally dependent condensed fonts, Chromium-only matrix and exposed-side focus contrast limitation remain disclosed limits. G001's lost historical logs/cached-build caveat, temporary-file residue after SIGKILL, production owner-role preflight concern, WSL probe limitation, inherited data-network egress and tag-only image pins remain historical/residual constraints. New G003 CI evidence must be retained and identify cache use; it does not retroactively recover lost logs or require unrelated provisioning/image hardening.

Metrics export listed in the generic STAGE-OBSERVABILITY follow-up is not part of the roadmap's bounded G003 health/logging/contracts/CI scope. Completion documentation may mark only implemented obligations and must leave exporter/collector adoption explicitly open for a separately scoped decision. Request log status/duration fields are not a claim of metrics-export completion. Durable audit is G011; broader secrets/vulnerability/engineering controls G012; recovery/incident monitoring G013; Organization data/branding G016/G017; product AI observability/usage remains in its later owning Goals. MinIO replacement remains a product-owner/vendor decision; G003 uses only the accepted local interface/stack.

## Architectural and Security Constraints

- Core owns facts, policy, tenancy, MCP, audit and public streaming authority. Web remains UI/thin server-only BFF; intelligence remains internal inference with no datastore credentials. Contracts contain no shared business logic.
- Durable product/domain entities use UUID canonical IDs; names/slugs are display/routing identifiers. Trace/span/request IDs and framework/migration metadata are infrastructure, not exceptions that authorize early domain schema.
- Diagnostic headers, test scopes/envelopes and CI actors are not authenticated tenant context. No fixture authenticator, fake membership or client-selected authority may enter production.
- Preserve Compose networks, runtime credential scoping and default loopback ports 3100, 8180, 8100, 5532, 6479, 9100, 9101. Test runners add no production network/port exposure. Existing --env-file configuration conventions stay in place.
- No external content/trace export or provider call, no new major vendor/service without an accepted ADR, and no SOC 2 compliance or full accessibility/tracing certification claim.
- Preserve developer data/resources; tests touch only verified owned disposable projects and current implementation copies. Existing conflicts are reported, not resolved by deleting other resources.

## Expected Artifacts and Allowed Areas

- Bounded operational additions/tests under `core/`, `web/lib` and existing web health routes, `intelligence/`, and contract schemas' test/fixture/documentation areas under `contracts/`. Schema bytes/versions are preserved unless separately reviewed.
- Root `.gitlab-ci.yml`, repository-owned CI/validation tooling under `infra/`, and narrowly justified Docker/check configuration/dependency-lock changes named in the Plan.
- Focused root README/foundation operational documentation and exact consumer-customization registration.
- Claude-owned `plans/G003-foundation-observability-ci-plan.md` and, after execution, `results/G003-foundation-observability-ci-result.md` plus a goal-specific exact scope allow-list. These are future deliverables; this Spec creates neither.
- Supervisor-owned Goal/Spec/PROGRESS changes at review handoffs only. These area names are not blanket write permissions; the Plan must enumerate exact intended paths, dependencies and temporary writes.

## Acceptance Criteria and Verification Strategy

| ID | Required outcome | Verification and evidence |
|---|---|---|
| AC1 | R1 health semantics and response bounds hold on the actual stack. | Capture status/JSON/no-store for every listed endpoint; independently stop each required dependency, optional intelligence and core on disposable copies, verify 200/503 distinctions and <=5s timeout cases, then recovery. Disabled intelligence is absent. Preserve inherited liveness/readiness tests. |
| AC2 | R2 context validation, generation and request isolation are enforced. | Table-driven tests in every participating runtime for valid, absent, malformed, oversized, duplicated, zero/invalid trace IDs and unsupported versions; effective response request IDs are bounded, generated IDs/spans nonzero; concurrency/failure/cancellation tests detect context bleed. Verify no baggage/arbitrary-header/authority propagation. |
| AC3 | Existing web → core → intelligence HTTP calls preserve context with parent attribution. | Built-stack request/header and completion-log captures for supplied/generated contexts, invalid replacement, concurrent requests, datastore/core/optional-intelligence outages and recovery; assert ID continuity and distinct spans/parents per documented convention. Disabled variant proves web → core continuity without intelligence. No new route or provider call. |
| AC4 | R4 structured logging and privacy hold without suppressing required diagnostics. | Parse actual service logs as JSON; exactly one completion per received scoped request with schema/type/status/duration/ID checks; exercise implicit success and handled failure. Raw canary scans cover nested keys/messages/errors and request/content inputs, with positive scanner and non-sensitive-log controls. Verify no duplicate unsafe access output or content export. |
| AC5 | R5 contracts agree across independent Go/Python validation. | Both validators check the same schema digest/corpus against independently declared expectations; valid/invalid categories and all eight event types are reported with no skips. Validate test-only Go producer and real fake-gateway Python emitter bytes across languages; missing/broken schema, invalid/truncated JSON/NDJSON and forbidden usage dimensions fail. No production execution client/schema broadening. |
| AC6 | Documented checks are runnable in containers from the actual repository. | Execute root commands verbatim; shadow host runtimes and record zero project-runtime invocations; build runtime/check/browser images from locked inputs and execute without test-time downloads/live providers. Preserve existing core/web/intelligence checks and full enabled/disabled browser matrices. Record tool versions, cache use and owned-resource cleanup. |
| AC7 | Required GitLab pipelines gate the reviewed tree and produce attributable hosted evidence. | Inspect workflow/default-branch/MR rules, runner privilege/resource scope, finite timeouts and installed tools. Hosted final passing pipeline identifies exact commit/tree and all required jobs/checks; a hosted disposable failing gate case proves wiring. No required allow-failure/skip, deployment/publish, production credentials or unsafe artifact upload. |
| AC8 | R7 negative controls and snapshot/failure propagation cannot produce a false green. | Disposable controls (a)–(f) produce nonzero aggregate results for the intended reason, do not bypass required checks and keep inputs isolated. Retain W1–W9 and dependency control expectations; failing status survives successful cleanup, post-run drift fails success, missing evidence fails, unsafe artifacts are withheld. Capture statuses and safe cleanup summaries. |
| AC9 | Scope, architecture and consumer provenance preserve G001/G002 guarantees. | Review complete tracked/untracked diff against exact paths/base; protected hashes, both Skills/manifest unchanged, exact customization conflicts/proposals and disabled parity pass. Full project regressions cover migration/roles/isolation/identity/default ports/outages/persistence/MCP/browsers; no new domain schema/routes/UI redesign/vendor or external authority. |
| AC10 | Documentation, Result and governance gates support bounded completion. | Result maps AC1–AC10 to commands/artifacts, distinguishes inherited/new/deferred and local/hosted evidence, records deviations, all failures, shared B7 cycles and checkpoint after cycle 4. Required evidence has no skips/unmet gates; unresolved prerequisites are resolved before execution. Codex reviews and returns ACCEPT before PROGRESS marks G003 complete or G004 proceeds. |

## Provenance and Evidence Requirements

Before any implementation write, record the reviewed baseline, approved Plan/Goal/Spec/input digests, NUL-safe tracked/untracked/ignored status, exact path set and resource/port inventory. Evidence destinations are outside the repository (except expressly allowed sanitized deliverables); containment and inspection must not mutate repository/Git state. Use read-only Git inspection with optional locks disabled for a read-only gate. Bind approvals to the bytes actually parsed; do not authorize an unchecked manifest to generate its own trusted hash.

Final validation covers the actual root and byte-identical current implementation copies, not pristine templates. Run the final integrated project suites after all shared harness/contract/logging changes, plus the G003 cases and required hosted pipeline. Keep before/after reviewed-input checks and cleanup inventory. Preserve exact commands, aggregate and per-check statuses/counts, scenario annotations and failure diagnostics outside the repository without secret values. Record unavailable/failed/aborted runs; later successes do not erase them. Hosted pipeline evidence must match the final submitted implementation; any differing local supervisor-only management update is identified explicitly.

G002's historical approval snapshot is not a baseline for new execution. At G003 Plan review Codex supplies the new exact binding. A later supervisor acceptance edit may change PROGRESS after validation; record its timing rather than pretending the prior run covered it.

## Permitted Deviations

Claude may choose job arrangement, library/validator selection, exact module names, test-helper decomposition, diagnostic field naming and server/client-span convention in the approved Plan when all measurable requirements remain satisfied. HTTP body meanings/routes and the existing AI wire schemas are fixed; additive diagnostic response headers and intelligence no-store headers described here are the only pre-authorized observable health-contract additions.

No acceptance criterion, privacy guarantee, failure control, full regression suite, hosted evidence or baseline provenance obligation may be weakened or deferred to fit B7. New dependencies/files must be attributable and reviewed; deviations discovered during implementation are recorded and returned to Codex. Major architecture/security/vendor changes require escalation, not retrospective acceptance.

## Risks and Questions Before Implementation

1. **GitLab runner profile and capacity:** the owner confirms an existing Docker-capable GitLab runner must be used; no runner was inspected or provisioned during this task. The Plan must identify its isolated execution profile, Docker/Compose/browser support, resource/port capacity, timeout budget and artifact access/retention. A need for privileged mode, production socket access or a new runner security model requires explicit review; no such decision is made here.
2. **Hosted evidence authorization:** planning does not permit pushes, pipeline triggers or CI/settings changes. Arrange separate authorization for the reviewed hosted success/failure runs before execution; if unavailable, report a prerequisite blocker and do not claim local evidence completes AC7.
3. **Validator semantics:** verify the selected Go validator handles the canonical Draft 2020-12/reference/format cases and agrees with Python without schema relaxation. Existing dependencies may be reused if they meet the corpus; otherwise enumerate/pin a bounded validator dependency for Plan review. Inability to meet the corpus requires a planning correction, not omitted cases.
4. **Framework logging and context lifetime:** the Plan must establish how existing Next.js/FastAPI/Go routes emit one private completion record and preserve status/cancellation/parent context without modifying the accepted UI or introducing unsafe generic middleware. Resolve uncertainties through read-only inspection and disposable planning probes, not application implementation before approval.
5. **First-build and artifact safety:** MinIO source compilation, pinned image availability and full browser regressions may be expensive on CI. Measure capacity/timeout assumptions; retain failure output safely without broad uploads or credentials. Unavailable images/tools/runner leave criteria unmet.

No unresolved product/domain identity, tenant model, telemetry vendor or UI design decision is necessary for this bounded Goal. The questions above are execution/Plan prerequisites; they may not become implicit permissions or future-Goal deferrals of G003 evidence.

## Completion and Gate Conditions

This Spec is ready for Claude planning, subject to resolving the listed prerequisites in the Plan. No implementation is authorized. Codex must approve the Plan, exact execution inputs/scope, B7 feasibility and runner/hosted-evidence arrangements before work begins. At handoff all AC1–AC10 must be supported, the final reviewed tree identified, data/resources preserved and failures/limits accounted for. Only Codex ACCEPT completes G003 and releases its dependency for G004 planning/execution under that Goal's separate gates. The G021 product-feature barrier remains closed.

## Revision Record

- 2026-10-05 — Initial specification from committed accepted G001/G002 state; ten acceptance criteria with per-criterion verification. No application/CI/test implementation, dependency installation, build, push or hosted action performed.

- 2026-10-05 — Product-owner insertion of G047 before G003 execution: add acceptance and refreshed Plan/baseline binding prerequisites. G003 scope, AC1–AC10, B7 and hosted evidence obligations remain unchanged; no implementation authorization.


## Proposed CI profile revision — 2026-10-06

See [G003 CI runner and feedback revision](G003-ci-runner-feedback-revision-spec.md) for the owner-requested bounded change to R6/AC7 triggering and existing-runner qualification. This is a linked draft; original text and historical bindings stay intact. The revision proposes medium-first hosted qualification and fast/affected/full gates, adds AC7a–AC7g, and preserves all full-suite assertions and hosted evidence. The narrowly changed privilege profile requires acceptance of proposed ADR-002, followed by Claude Plan/Codex review and a fresh binding before revised implementation.
