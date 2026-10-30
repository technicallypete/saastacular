# Operations: health, diagnostics, logs, contracts and CI

This guide describes the operational foundation on top of the inherited health and readiness paths. It covers health semantics and bounds, diagnostic context (request IDs and trace context), structured logs and privacy, cross-service contract conformance, and the GitLab CI gate. It adds no route, contract, schema or authority. Metrics, exporters, collectors, dashboards and alerting remain open (see [FOLLOW-UP.md](FOLLOW-UP.md)).

## Health semantics and bounds

| Service | Route | Body | Status | Headers |
|---|---|---|---|---|
| core | `GET /healthz` | `{"status":"alive"}` | 200 | `Content-Type: application/json`, `Cache-Control: no-store`, `X-Request-Id` |
| core | `GET /readyz` | `{"status":…,"product_name":…,"checks":{"object_store":…,"postgres":…,"redis":…}[,"optional":{"intelligence":…}]}` | 200 / 503 | as above |
| web | `GET /api/health` | `{"status":"alive"}` | 200 | `Cache-Control: no-store`, `X-Request-Id` |
| web | `GET /api/ready` | `{"status":"ready"\|"unready","checks":{"core":"ok"\|"unavailable"}}` | 200 / 503 | `Cache-Control: no-store`, `X-Request-Id` |
| intelligence | `GET /healthz` | `{"status":"ok"}` | 200 | `Cache-Control: no-store`, `X-Request-Id` |
| intelligence | `GET /readyz` | `{"status":"ready"}` (also when no model is configured) | 200 | `Cache-Control: no-store`, `X-Request-Id` |

- Core's required checks are PostgreSQL, Redis and object storage. Intelligence is reported under `optional` and never gates readiness; it is absent when intelligence is disabled.
- Deadlines: core readiness 2 s; core's intelligence probe 2 s; web's call to core 2 s. Every single request in the outage matrix answers within 5 s.
- **Nested deadlines.** Web's 2 s timeout starts before core's 2 s readiness deadline. A *paused* (hung) dependency therefore makes core answer at about 2 s while web usually times out first (`core_probe: timeout`) and reports `core: unavailable`. A *stopped* dependency can behave the same way: where Docker's embedded DNS answers a lookup of a stopped container's name only after the deadline (about 4 s on a measured validation host), that lookup takes the whole 2 s. So, with PostgreSQL, Redis or object storage stopped, core answers 503 naming the dependency while web answers 503 with `core_probe` `not_ready` or `timeout`; with core stopped, web answers 503 with `unreachable` or `timeout`; with the optional intelligence service stopped, core answers 200 with `optional.intelligence: unavailable` (it never gates core's readiness) while web answers either 200 (`ok`) or 503 (`timeout`), because the optional probe can use up web's budget for its call to core. The validation records the observed outcome of every case (`obs-outages.tsv`, `obs-off-outages.tsv`). This is inherited behaviour, kept on purpose; the 2 s deadlines are unchanged.
- No error text, DSN, URL or credential appears in any body. Health bodies are byte-identical to the previous foundation; only the headers are additive.

The harness verifies these on the actual root stack (B-HEALTH-ROOT) and, on a disposable copy, under stopped and paused PostgreSQL, Redis, object storage, intelligence and core (B-OBS-TIMEOUTS), followed by the inherited outage and persistence checks.

## Diagnostic context

**Request ID.** Exactly one `X-Request-Id` field value matching `^[A-Za-z0-9._-]{8,64}$` (full-string match) is adopted. A missing, duplicated, comma-joined, over-long or non-matching value, or one that contains a credential the receiving service holds, is never reflected or logged: the service generates 32 lowercase hex characters instead and records `request_id_source: generated`.

**Trace context.** A narrow profile of the [W3C Trace Context Recommendation](https://www.w3.org/TR/2021/REC-trace-context-1-20211123/), not a full-compliance claim:
- accepted: exactly one `traceparent` value of at most 512 bytes matching `^00-<32 hex>-<16 hex>-<2 hex>$`, with nonzero trace and parent IDs;
- every other version, uppercase hex, a wrong length, extra fields, whitespace, duplicates or comma-joined values are invalid: the service starts a new trace (`trace_source: generated`) without an error and never reflects the invalid value;
- only the sampled bit (`01`) is interpreted; outbound flags are `01` or `00`; the flag confers no authority, sampling or content capture;
- `tracestate` and `baggage` are never read, forwarded or logged.

**Spans.** Each receiving service creates its own nonzero span ID for its server operation. `parent_span_id` is the inbound parent ID, present only when the inbound `traceparent` was valid. Outbound calls carry `traceparent: 00-<trace>-<own span>-<flags>` (server-span-as-caller convention). On the full stack: `core.parent_span_id = web.span_id` and `intelligence.parent_span_id = core.span_id`; all three share `trace_id` and `request_id`.

**Propagation allow-list.** Application code sets exactly two outbound headers, `X-Request-Id` and `traceparent`, on the calls web `/api/ready` → core `/readyz`, core readiness → intelligence `/healthz` and web's identity client → core `/internal/v1/auth/*`. The identity client also sends exactly `Authorization: Session <credential>` (the browser's own session), `X-Login-Binding` (login and callback only) and `Content-Type: application/json`. The identity web routes emit no completion record; core records their closed `auth.*` events with the forwarded `request_id` and `trace_id` (`IDENTITY.md`). One narrow
hostname-routing diagnostic exception: web's host gate (`web/proxy.ts`, every gated page and form request) and its `/auth/callback`
route each make a host lookup (`GET /internal/v1/hosts/resolve`) before the request's own core call; only that lookup
carries a freshly generated `X-Request-Id` (`lookupContext` in `web/lib/host.ts`), with the same `trace_id` and caller
span, while the page's, form's, callback's or relay's own core call keeps the inbound request ID and trace context, so
core's one-completion-record-per-request-ID expectation holds even when the browser or edge supplies an `X-Request-Id`.
Health-chain propagation is unchanged, and this narrow exception is not a general per-call request-ID rule. No inbound header, cookie, `Authorization`, `baggage` or `tracestate` is copied. Request-local state is a `context.Context` value in core, an explicit argument in web and local variables in intelligence; there is no shared mutable request state.

## Structured logs

Every application-owned line is one JSON object with `time` (RFC 3339 UTC, milliseconds), `level` (`debug`, `info`, `warn`, `error`), `service` and `event`. `event` is a constant authored in code; runtime values are never formatted into it.

**Completion record** (`event: http.request.completed`), exactly one per scoped health request that reaches the service's handler (web: the exported `GET` handler, which Next also uses for `HEAD`; other methods are answered by Next before the handler and are not logged): `method`, normalized `route`, final `status`, `duration_ms`, `request_id`, `request_id_source`, `trace_id`, `trace_source`, `span_id`, `parent_span_id` (only when valid inbound), `trace_flags`, `outcome` (`completed`, `cancelled`, `failed`), `error_type` (only on failure), and on web `/api/ready` `core_probe` (`ok`, `not_ready`, `timeout`, `unreachable`). Core logs every request it receives, including 404 and 405 (`route: unmatched`); its matched routes also include exactly the five identity patterns `/internal/v1/auth/session`, `/callback`, `/login`, `/logout` and `/provider-logout/return` (never a raw path or query). Intelligence records a client that disconnects as `completed` with the application's status, because uvicorn does not cancel the request task on disconnect; core and web record `cancelled`.

**Error output.** Framework, library and error text is never copied into a record:
- **core:** a failing start writes exactly one `core.exit` record with a fixed `error_type` (`config_invalid` with the allow-listed variable names, `database_pool_invalid`, `listen_failed`, `serve_failed`, `shutdown_timeout`, `unknown`); `net/http` server errors become `http.server.error` records with fixed categories; the server binds before logging `core listening`, which carries only the bound port and `app_env`;
- **web:** the route wrapper never lets an error reach Next.js; it records `outcome: failed` with a closed `instanceof` category and answers 500 with an empty body, `Cache-Control: no-store` and `X-Request-Id`. Next.js prints only its startup banner;
- **intelligence:** uvicorn records are mapped by message template to fixed `uvicorn.*` events with arguments dropped; other libraries become `library.log`, warnings `python.warning`; a startup failure writes one `intelligence.startup.failed` record and exits 3 without a traceback.

`infra/validation/tools` (`logcheck`) classifies every output line of each service against a per-service event catalogue and the completion-record schema; any unclassified line fails.

## Privacy

Never logged anywhere: raw headers, cookies or query strings; request or response bodies; prompts, completions, retrieved content or envelopes; authorization references; tenant or user identifiers; URLs, DSNs or exception text. Sensitive-looking keys (`password`, `secret`, `token`, `api_key`, `authorization`, `cookie`, `dsn`, `credential`) are redacted, and each service's configured secret values are scrubbed from every string. Intelligence content capture stays off unless a hook is constructed with it explicitly. The harness proves this with per-run canaries in headers, queries, bodies, identifiers and execution envelopes, plus positive controls that show the scans detect a planted canary (B-OBS-PRIVACY, B-SECRET-CANARY).

## Contract conformance

`contracts/conformance/` holds a synthetic, hand-authored corpus for the two intelligence schemas, outside `contracts/intelligence/` so it never enters the runtime image. `expectations.json` declares every case's outcome independently of either validator and pins both schema digests.

- **Go:** `contractcheck` (test tooling in `infra/validation/tools`, santhosh-tekuri/jsonschema v6.0.3, Draft 2020-12, format assertion, strict JSON, no remote loading).
- **Python:** the existing `streaming.schemas` validator, corrected to enforce what the schemas already say: ECMA-262 `pattern` anchoring (a trailing newline no longer passes `$`), a project-owned RFC 3339 `date-time` checker with calendar rules, and a keyword-profile audit. Intelligence's `/internal/v1/execute` consequently rejects identifiers ending in a newline. A leap-second deadline (`23:59:60Z`) remains schema-valid but is still refused by the existing runtime deadline parsing.
- **Normal mode** checks the schema digests first; any byte difference fails with zero cases evaluated. **Audit mode** classifies deliberately broken schemas (`schema-negative/`) by a fixed stage order. Peer exchange checks Go-produced envelopes in Python and real Python NDJSON in Go, and rejects corrupted copies.

Run them with the README `platform:contracts` block; `pytest` in `intelligence-check` runs the Python side.

## CI

GitLab CI is the retained, supported CI provider for this template. It needs your own GitLab project and runner; the local commands in the README need no GitLab account, token, remote or hosted runner. Hosted pipelines have not been run for this exported copy. For a GitHub Actions migration, see the [parity mapping](../CI-GITHUB-ACTIONS-PARITY.md); for the optional staging deployment tooling, see [STAGING.md](STAGING.md).

`.gitlab-ci.yml` defines two gating jobs on GitLab's hosted Linux runner (`saas-linux-medium-amd64`): `validate:checks` (stage `checks`) in every supported pipeline, and `validate:integration` (stage `integration`, after `validate:checks`, without downloading its artifacts, so the gate's checkout stays clean) optional/manual (rule-level `allow_failure: true`) in ordinary merge-request pipelines, and automatic and required in a release-candidate promotion merge request (`release/X.Y.Z-rc.N` to the default branch: the complete full profile) and in schedules, web pipelines and API branch pipelines. **No push starts a pipeline**: feature branches, `staging` (not even `validate:checks`), the default branch and every tag. The workflow runs merge-request pipelines (the release-candidate promotion merge request without the documentation filter), default-branch schedules, the protected identity-qualification schedule, web pipelines and branch pipelines started through the pipelines API (for example `glab ci run -b <branch>`), and nothing else. A feature branch is checked by its merge request's pipeline or by an explicit web or API pipeline; a full local run (`bash infra/ci/validate.sh --profile full`) is the broader check for an integrated branch. There is no publish or release job; the only deploy jobs are the manual staging jobs (below), which no validation context creates. Neither validation job uses a CI/CD secret; both are `interruptible: false`.

**Identity staging.** A third job, `identity-staging` (stage integration, no `needs`, 45 minutes), runs only for a web or scheduled pipeline of the protected `identity-qualification` branch (the workflow's identity-qualification schedule rule admits only a *protected* qualification schedule). It declares `environment: identity-staging`, so it alone receives the protected, environment-scoped project variables `IDENTITY_STAGING_WORKOS_API_KEY` (masked and hidden), `IDENTITY_STAGING_WORKOS_CLIENT_ID` and `IDENTITY_STAGING_WORKOS_ISSUER`. Live identity-provider qualification through this job has not been executed for this export.

- **Forwarding.** `hosted-context.sh` forwards them to `validate.sh --profile staging` as one mode-600 file mounted read-only, and only when everything holds (the file reaches the job-private daemon through a job-scoped volume filled over a helper container's stdin, never as a bind-mount source, because a `-v` path would name the daemon's filesystem; the volume is removed when the job ends): the approved project ID and path (pinned by SHA-256 in `infra/ci/hosted-context.sh`; an adopter sets these pins for their own GitLab project), branch `identity-qualification` with `CI_COMMIT_REF_PROTECTED=true`, source `web` or `schedule` without a merge request, job and environment `identity-staging`.
- **Refusals.** A staging profile anywhere else is refused (16). A staging variable visible to any other profile fails the job closed before a context starts.
- **Missing inputs.** A missing input is UNMET, never skipped.
- **Secret handling.** Only the API key joins the credential patterns and the secret registry; the client ID and issuer are public metadata.
- **Disk floor.** The staging profile's 15 GiB floor is provisional and should be re-derived from a first run's measured peak.
- **Schedules.** In the same pipelines `validate:checks` and `validate:integration` run the full profile: a protected qualification schedule is attributed `branch-schedule`, like a default-branch schedule. Any other non-default schedule is refused.
- **Controls.** C-SEC-1..10, N12q, F-schedule-qualification and F-schedule-other cover these rules.

**Manual staging deployment.** Deploy jobs exist behind the typed pipeline inputs `action` (`validate` by default, `start-deploy`, `stop`), `environment` (`none` by default, or an entry of `infra/staging/ci-environments.txt`), `candidate` and `duration`. Defaults leave merge-request, release-candidate promotion, schedule and ordinary web pipelines exactly as above. Gating: a non-default action is admitted only for a web or API pipeline of the protected `staging` branch (any other branch or source: workflow `when: never`); `environment=none`, or `stop` with a candidate or duration, yields zero jobs, so GitLab refuses to create the pipeline; the validation jobs and `identity-staging` do not run in deployment mode. `start-deploy` runs `staging:prepare` and `staging:build` (no ID token, environment or protected variable; the build holds the repository-wide `staging-image-build` resource group), `staging:check` (read-only identity) and `staging:deploy`; `stop` runs `staging:stop` alone in its own pipeline. Deploy and stop declare the selected protected GitLab environment, hold its `staging-env-<env>` resource group and federate to Google Cloud with an environment- and action-specific ID-token audience; no service-account key or deployment secret is a CI/CD variable. The launcher `infra/ci/staging-action.sh` re-checks the approved context (project, protected `staging`, web/API source, no merge request, job and environment names) and refuses with 16 before any token is touched. These jobs are ordinary jobs (not manual, never allowed to fail). Live use needs the adopter's own GitLab settings and cloud federation; this tooling has not been exercised against any provider for this export. See [STAGING.md](STAGING.md#manual-ci-startdeploy-and-stop).

**Profiles.** Both jobs run the unchanged entry point `infra/ci/validate.sh`, which runs a list of `infra/validation/run.sh` suites. Profiles are suite lists in `infra/ci/lib.sh` (`CI_SUITES_*`, canonical order `CI_SUITE_ORDER`); `infra/validation/required-checks.tsv` maps each check ID to one suite:

| Profile | Suites | Required IDs | Covers |
|---|---|---|---|
| fast | lint, lint-base, lint-web, lint-core, provenance, contracts, unit | 32 | static rules, provenance, both contract validators, the README check block without a stack |
| affected | lint-base and provenance, plus per selected area kind: web `lint-web unit-web feature-web`; core `lint-core unit-core` (+ `feature-core` when the area has integration tests); intelligence `contracts unit-intel`. All fast suites too when a fast/agent-class path is also changed | per selection (web only 17, core only 13, intelligence only 17) | the changed registered feature areas, their dependents and consumers; fast job only |
| integration | fast plus root, ops, rename, off (+ the feature suites of all registered areas) | 72 base IDs, plus selected feature IDs | the root stack and browsers, observability and outages, the identity environment (emulator, including B-MCP-E2E), the rename and toggle-off copies |
| full | lint, lint-base, lint-web, lint-core, controls, ci-controls, provenance, contracts, root, migrations, ops, rename, off (`run.sh all`; + the feature suites of all registered areas) | 124 base IDs, plus selected feature IDs | everything, including the control matrices and migrations (among them the four B-RLS checks, the five B-OBJ checks, B-SECRET-SOURCE, B-DEPSCAN and B-SECRET-ROTATION, the two B-RETENTION checks and the two drills, B-MCP-CATALOG, B-MCP-IT and B-MCP-E2E, and B-IT-ISOLATION; base count from `infra/validation/required-checks.tsv`, which B-CI-CONFIG reports) |
| staging | staging | 5 | the real WorkOS staging contract (B-IDENTITY-STAGING), only in the `identity-staging` job |

For the purpose of each suite and how a live check ID maps to its implementation and scenario documentation, see the [test and validation index](TEST-INDEX.md).

**Lint split.**
- The four `final` checks (B-DOCKER-ONLY, B-SECRET-CANARY, B-SCOPE, B-REQUIRED-CHECKS) run in every job.
- The lint group is split by input boundary:
  - `lint`: the nine checks whose inputs are foundation-class only (B-SHELL, B-HADOLINT, B-COMPOSE, B-ENV-EXAMPLE, B-DEPS, B-MIGRATE-SET, B-DOCKERIGNORE, B-CI-CONFIG and B-SECRET-SOURCE);
  - `lint-base`: the repository-wide guards (B-LITERALS, B-NAMING, B-IDENTITY-GUARDS and B-DEPSCAN, which reads core sources and all three lockfiles);
  - `lint-web`: the five web guards;
  - `lint-core`: B-MCP-SOURCE.
- B-UNIT-WEB, B-UNIT-CORE and B-UNIT-INTEL each run their service's README `platform:check` line plus `secret-scan`. This is an exact subset of B-UNIT-CHECKS, which is unchanged.
- B-FEATURE-WEB and B-FEATURE-CORE run a registered area's own browser specs and its real-PostgreSQL integration tests.

**Security checks ([SECURITY.md](../SECURITY.md)).**
- B-SECRET-SOURCE (`lint`): the submitted-source scan (`infra/security/secret-scan.sh` and `infra/compose/secret-scan.yml`) on disposable repositories: overlay configuration, a clean tree with a directory link and an outside-target link, tracked and untracked canaries, force-added exempt paths, an ignored `.env`, a nested repository and a gitlink, incomplete streams and an invalid configuration. The README `platform:check` scan line runs the same script on the real tree.
- B-DEPSCAN (`lint-base`, so it runs in the merge-request fast job and every affected profile): `infra/security/dependency-scan.sh`, fixed offline fixtures plus the current-tree scan of the three dependency graphs. It is the one source check that downloads data: four fixed public advisory objects (about 270 MB) and Go modules for govulncheck; every scan runs with network none. An unavailable or stale advisory export fails it by design; its pins are reported as `DEPSCAN-PINS`.
- B-SECRET-ROTATION (`ops`): last on the disposable ops copy, one rotation and withdrawal of the copy's runtime PostgreSQL role with synthetic values (open-connection behavior, termination, core restart, fail-closed withdrawal).

**Recovery and retention checks ([RECOVERY.md](RECOVERY.md)).** All four run last in `migrations`:
- B-RETENTION-CATALOG: TestRetentionCatalog and TestRetentionMigration, the exact version-9 delta, the role boundary, eight rolled-back detectors and the 00009 up/down/reapply cases.
- B-RETENTION-IT: TestRetention with the actual retention login (absent and malformed settings, scope, holds, revalidation, refusals for every other role) and the bounded E-RET-1..3 plan observations, kept as `retention-eret.jsonl`.
- B-RECOVERY-DRILL: on the migrations project as the source and a second recovery project on the same copy (host ports `PORTS_REC`: 3700 8780 8700 6132 7079 9700 9701), the seeded two-Organization fixture, the capture fence and its interruptions, backups B1-B3, the incident exercise, restore refusals, the staged restore, invalidation, the ledger gate, replay with an injected audit fault, the runtime verification and the reopen sign-off; sanitized measurements in `recovery-drill.tsv`.
- B-MAINTENANCE-DRILL: on the reopened target, audit retention, pending/deleting repair, tombstone purge, backup expiry and the status checks; then the private drill material is removed and asserted gone. UNMET (never PASS) when the recovery drill did not reach its reopen step.

**MCP checks ([MCP.md](MCP.md)).**
- B-MCP-CATALOG (`migrations`, after the retention checks): TestMCPCatalog and TestMCPMigration, the exact version-10
  delta, ten rolled-back detectors and the 00010 up/down/reapply cases.
- B-MCP-IT (`migrations`, before the recovery drill): the MCP transport, lifecycle, audit and recovery tests over the
  production constructor with the official SDK client, then the facade ordering and footprint (`mcp-emcp.jsonl`); the
  core records pass logcheck and the canary scans. TestMCPRecovery applies the actual restore invalidation and expires
  every session; the invocation starts on its own fresh database copy, so its position does not matter.
- B-MCP-E2E (`ops`, in the identity project after B-IDENTITY-PROBES): the Go driver (tag `mcpe2e`) through the TLS edge
  with the mcp-* browser projects' handoff and the run's Connect test signer, then the real core image's MCP
  configuration refusals. Before the browser run the harness seeds the test Organizations and accounts
  (`infra/identity/mcp-seed.sql`, with the emulator user IDs looked up through the emulator's API; see the domain checks below).
- Also covering MCP: B-MCP-SOURCE (the server links core/internal/mcp, the access facade, the tenancy model and the pinned SDK
  closure; one handler call site), B-MCP-HTTP (404 everywhere on the unconfigured default stack), B-WEB-BOUNDARY,
  B-IDENTITY-GUARDS (the C10 test-anchor guard, N11), B-IDENTITY-E2E (the mcp-light and mcp-dark projects),
  B-SKILL-UNCHANGED and the version-10 migration pins.

**Domain checks ([DOMAINS.md](DOMAINS.md)).**
- B-DOMAINS-CATALOG (`migrations`, before the recovery drill): TestDomainsCatalog, the exact version-11 delta, the
  readers, policies and guards, and the 00011 down refusals.
- B-DOMAINS-IT (`migrations`, after B-DOMAINS-CATALOG): TestDomains DOM-1..10 on the attested runtime login; the four
  `E-DOM-F` footprint lines are kept as `domains-edom.jsonl`; the access records pass logcheck and the canary scans.
- Also covering domains: B-MCP-E2E's seed (`infra/identity/mcp-seed.sql` creates Organizations A, B and C with the slugs
  `org-a`/`org-b`/`org-c`, accounts and memberships for every browser sign-in user, the routing custom host
  `portal.acme.test`, the removed `old.acme.test` and dave's historical old-host connection; the harness passes every
  seeded emulator user ID and the three Organization UUIDs) and its configuration refusals (the retired registry and the
  production fixture TLS adapter); B-IDENTITY-E2E's `domain-platform` and `domain-custom` projects
  (`BROWSER_MATRIX_DOMAINS`); B-MCP-SOURCE's pins of the server source, imports and access importers; the version-11
  migration pins and catalog counts (48 relations).
- Operator steps for domains (claim, verify, recheck, removal, the manual provider cleanup after an exhausted
  withdrawal retry, restore rechecks and the edge inputs) are in [DOMAINS.md](DOMAINS.md). None of them is automated.

**Jobs and partitions.** `infra/ci/select-profile.sh` decides one pipeline profile and suite list from the pipeline context and the changed paths. Each job executes its partition. Internal profile/job-kind tokens remain `fast` and `gate` for CLI, summary, JUnit and pipeline-union compatibility; the external job names below do not change selection or required coverage:

- **`validate:checks` job.** It runs the affected profile when the pipeline profile is affected. Otherwise it runs the fast profile: complete for a fast pipeline, or the fast partition of an integration or full pipeline.
- **`validate:integration` job.** It executes nothing (`harness=not-required`) for fast or affected pipelines. Otherwise it runs the **remainder** of the pipeline profile, meaning its suites minus the fast job's. Its JUnit suite is then `gate:remainder`, and its summary says `coverage: partial`.
- **Contexts.**
  - Manually started merged-result MR gates, automatic schedules, web pipelines and API branch pipelines select full without consulting a diff.
  - The release-candidate promotion merge request (tested content `mr-rc-promotion`, reason `rc-promotion`) selects full without consulting a diff, documentation-only candidates included, with every present registered area and `ci-controls` regardless of `CI_CONTROLS_FORCE`; both jobs are automatic and required.
  - Detached MR pipelines classify the paths changed since `CI_MERGE_REQUEST_DIFF_BASE_SHA`. The automatic MR job is `validate:checks`, which also runs affected coverage for registered areas; the MR gate stays optional and manual.
  - No push starts a pipeline: feature branches, `staging`, the default branch and every tag. A feature branch is checked by its merge request's automatic `validate:checks` job or an explicit web or API pipeline (full). The selector keeps its `branch-push` semantics (paths changed since `CI_COMMIT_BEFORE_SHA`) for explicit attribution only.

**Pipeline coverage.** A full or integration pipeline's coverage is the **union of both partitions**, never a single job:

| Pipeline | validate:checks | validate:integration | Distinct obligations | Overlap |
|---|---:|---:|---|---|
| full | 35 | 82 | 113: the 112 full IDs plus B-UNIT-CHECKS | exactly the four `final` checks |
| integration | 35 | 44 | 75 | the four `final` checks |

Each job's own completeness is exact: its executed suites plus `final`, each ID once as PASS, and any row outside them fails.

**Summary fields:**
- `pipeline_profile`, `pipeline_suites`, `pipeline_set_sha256`;
- `executed_profile`, `executed_suites`;
- `coverage`: `complete`, `partial` or `none`;
- `partition`: `fast`, `remainder`, `complete` or `none`;
- `areas`;
- `selection_digest`, over the candidate commit and tree, the map, selector and registry blob IDs, the base, the areas and the pipeline suites;
- `comparison_base`;
- `required_count` and `required_set_sha256`, over the job's own set;
- `ci.project_id` (the forwarded `CI_PROJECT_ID`) beside the pipeline, job and commit identity;
- `release`: `null`, or for `mr-rc-promotion` the release-candidate identity `{tag, tag_object, peeled, staging_sha, merge_request_iid, source_branch, target_branch}`. `staging_sha` is that job's own observation of `refs/remotes/origin/staging`.

**Union check.** `bash infra/ci/pipeline-union.sh FAST_DIR GATE_DIR` checks two published `ci-evidence/` directories. It verifies:
- both partitions are present and successful;
- they share the candidate, the selection digest and the pipeline set;
- they share the tested content, CI project, pipeline and commit, and for a release-candidate promotion (always full) the release identity; each job's observed staging tip is validated per job only, since staging may advance between the jobs;
- each job's suites equal its expected partition;
- the union equals the pipeline set;
- nothing overlaps except the four `final` checks.

`--single DIR` checks a complete single-job pipeline.

The reviewed map `infra/ci/profile-paths.tsv` classifies each path (an exact file first, then the longest directory prefix):

- **docs** selects fast. Root Markdown files, Markdown under `docs/` and under the project-management directory, a few exact non-executable tracking and planning-provenance files, and three exact legacy handoff-skill paths (inert) are docs; automatic docs-only push/MR pipelines are omitted entirely. Root `.gitignore` changes are fast metadata through the same class (see [Git ignore metadata validation](#git-ignore-metadata-validation)).
- **agent** selects fast: other paths under the root agent configuration directories (Claude and Codex), except a foundation-classified design-skill directory and its alias (not present in this template), including additions, deletions and renames within them. Automatic pipelines are admitted for them (fast checks run); a move out of them has its destination classified on its own.
- **Nested Claude skills:** any project-subtree Claude skills directory (for example under `web/`, `core/` or `docs/`, but not the root agent directories) is foundation and always admits automatic pipelines, ahead of the Markdown documentation rule. Web-local agent configuration directories as a whole (unknown or future web agent configuration) are foundation. Legacy `distributions/skills/**` paths retain foundation classification if reintroduced; this template has no distributions directory.
- **area-web, area-core, area-intel** select affected, for an explicitly registered feature area (below).
- **runtime** selects integration: the remaining `core/`, `web/` and `intelligence/` paths. Examples: `core/internal/health/`, `web/app/page.tsx`, `web/app/api/{health,ready}/`, `intelligence/src/{main.py,models,observability,providers}/`.
- **foundation** selects full. It covers:
  - non-Markdown controls under `docs/` and the project-management directory;
  - CI, Compose, build and `infra/`;
  - contracts and the conformance oracle;
  - dependency locks;
  - migrations, and the migration CLI's own core dependencies (`core/internal/observability/`, `core/internal/config/`);
  - the tenancy model;
  - build and test configuration;
  - shared browser tests, support helpers and fixtures;
  - the security and global-UI boundaries:

    | Boundary | Paths |
    |---|---|
    | Identity, session, authorization | `core/internal/identity/`, `core/cmd/server/`, `web/app/{auth,sign-in,account}/`, `web/app/api/account/`, `web/components/account/`, `intelligence/src/security/` |
    | Tokens, themes, fonts, first paint | `web/app/globals.css`, `web/app/layout.tsx`, `web/public/`, `web/components/layout/` |
    | Global UI primitives | `web/components/ui/` |
    | Shared helpers | `web/lib/`, `web/tests/unit/` |
    | Contracts | `core/internal/mcp/`, `intelligence/src/streaming/` |

**Feature areas.** An area exists only through explicit map rows. Adding those rows is a foundation change, so the registering change runs full and is reviewed; later changes inside the area run affected.

- **Rows and roots.** Rows are `area-web dir`, `area-core dir` or `area-intel dir`, with these roots only:

  | Class | Roots |
  |---|---|
  | `area-web` | `web/app/(features)/<name>/`, `web/tests/browser/features/<name>/`, `web/tests/fixtures/features/<name>/` |
  | `area-core` | `core/internal/<name>/` |
  | `area-intel` | `intelligence/src/<name>/`, `intelligence/tests/<name>/` |

- **Names.** The name is the last segment: web `^[a-z][a-z0-9-]{0,31}$`; core and intelligence `^[a-z][a-z0-9]{0,31}$`. A bad root, a bad name, a duplicate or an overlap is refused (16).
- **What is registered.** No area is registered in this template. `web/features/` is not an allowed root until the web guards are extended to cover it. A new core package without a row stays runtime (integration).
- **File kinds.** Inside an area only ordinary file kinds stay narrow:
  - web: `*.ts`, `*.tsx`, `*.css`, data `*.json` other than package, tsconfig and config files, images, `*.md`, `*.txt`;
  - core: `*.go`, `testdata/**`, `*.md`;
  - intelligence: `*.py` except `conftest.py` under `src/`, data under `testdata/` or `fixtures/`, `*.md`.

  Anything else selects full: scripts, Dockerfiles, `go.mod`, `*.sql`, `pyproject.toml`, `.env*`, `*.config.*`, symlinks, submodules. A web `route.ts` or `route.js` handler selects integration.
- **Browser specs.** A web page area needs at least one feature spec under `web/tests/browser/features/<name>/`; otherwise B-FEATURE-WEB is UNMET. Feature specs run in the `feature-light`/`feature-dark` Playwright projects (`FEATURE_E2E=1`) against the fixture build, without a stack. `*.stack.spec.ts` specs run against a disposable full stack.

**Dependency closure.** For each touched area, the selector runs `git grep` over **both** the comparison base and HEAD. It searches every tracked file outside the area, including tests, fixtures and testdata:

| Kind | References detected |
|---|---|
| Go | `core/internal/<name>` followed by `"` or `/`: subpackages and every import form |
| Web | `(features)/<name>` or `features/<name>` strings, plus every quoted relative specifier resolved against its file: import, export-from, `import()`, `require()`, CSS `@import`/`url()`, `new URL()` |
| Python | `import <name>`, `from <name>[.sub] import`, plus relative imports resolved against the package path |

**How a reference widens the selection:**
- **Reference classes.** A referencing area joins the selection; a runtime reference raises it to integration; a foundation or `core/cmd/*` reference raises it to full. Documentation contributes nothing.
- **Ambiguous forms** outside an area select full when a web or intelligence area is touched: a non-literal or template `import()`/`require()`, a specifier escaping `web/`, `importlib.import_module`, `__import__`, `exec`, a `sys.path` change, or an unresolvable relative import.
- **Stack consumers.** A changed core or intelligence area also adds every web area with stack specs.
- **Deleted and empty areas.** A deleted area keeps its kind's static and unit checks plus every base-tree consumer. An area registered but empty in both trees adds nothing (`registered-absent`).
- **Failure.** Any failure of the closure selects full (`closure-failed`).

**Precedence.** The highest class wins: fast < affected < integration < full. Any of the following also selects full:

- an unknown path, a deletion or rename outside registered areas, or an unsafe path name;
- a comparison base that is missing, zero, absent, not an ancestor of HEAD, or equal to HEAD;
- a diff that cannot be computed (`diff-failed`).

The selection is published with the evidence: `selection.tsv` (status, class, escaped path, area, via) and the summary fields above.

**Hosted adapter.** Each job runs `sh infra/ci/hosted-context.sh` in the pinned `docker:cli` image beside a job-private TLS Docker-in-Docker service (alias `docker`).

1. **Credentials.** No nested container mounts the build directory until the adapter has made it safe. GitLab writes the job token into the job-created `.git/config` even with `FF_GIT_URLS_WITHOUT_TOKENS`. The adapter removes the password from every remote URL, keeping scheme, user, host and path, and examines included configuration. It then scans every regular file of the build-directory parent (FIFOs, sockets and devices are never opened) to prove that no job credential remains, using a mode-600 pattern file in job-local `/tmp`; no value is printed or passed as an argument, and any read error counts as unproven. The job is refused (16) by a residual credential, a password-bearing included URL, a nested or unsafe include, a quoted or otherwise unsafe configuration, an unreadable file or directory, or a missing job token.
2. **Context.** The adapter builds `infra/docker/ci-context.Dockerfile` and runs the entry point in an unprivileged context container. The container has `--network host` on the daemon's network namespace, the build-directory parent at the identical path with the runner's `<project>.tmp` (file variables, Git template) masked by an empty tmpfs, the client certificates read-only, `no-new-privileges` and an allow-listed environment.
3. **Disk guard.** The guard records available space after setup, before the workload. It then samples at most every 30 s, covering daemon storage too when that is a separate filesystem (through one sampler container started before the workload). Each measurement is time-limited. It fails the job below 4 GiB, on a failed or hung measurement, or when the guard dies or stops sampling: the context is stopped, and the job exits 20, or with the entry point's own nonzero status. It never deletes or prunes anything. For each filesystem it reports the incremental peak (`baseline − minimum`, in bytes) and the durable start floor `max(35 GiB, ceil(1.5 × peak + 4 GiB))`.
4. **after_script.** It stops a remaining context. After the same credential proof (which does not read the job's own context temporary directory, since the context never held a credential), it runs `cleanup-owned.sh --after-script` in a fresh context.

**Recommended local developer workstation.** Recommend **32 GB of system RAM** for local development and validation. **64 GB** provides additional headroom for developers who keep other business applications open alongside the development stack. This is a workstation recommendation, separate from the runner preflight minima below.

**Capacity and time.** `runner-preflight.sh` enforces capacity for the executed profile:

- **fast:** 2 CPU, 4 GiB memory and 15 GiB free;
- **integration and full:** 4 CPU, 8 GiB memory, and 40 GiB free on the temporary directory and on daemon storage. Only inside the qualified hosted adapter context is this a provisional 35 GiB; a caller flag never lowers it.

Job timeouts are 40 min for validate:checks (script 35 min, after_script 4 min) and 3 h for validate:integration (2 h 45 min, 10 min). The harness budget is `min(CI_JOB_TIMEOUT, RUNNER_SCRIPT_TIMEOUT)`, minus the time elapsed since `CI_JOB_STARTED_AT`, minus a reserve (5 min for fast, otherwise 20 min). The budget is capped at 25 min for fast/affected and 90 min for integration, and must be at least 10, 30 and 60 min respectively.

**validate.sh** (Bash 4.4 or later):
1. **Location** checks the evidence and registry paths for containment before creating anything, and checks the profile: in CI only `--profile fast`, `--profile gate` or `--profile staging`, where no flag means full; locally `fast`, `integration` or `full`, default full. Otherwise it exits 16. `--profile staging` first requires a job timeout of at least 2700 s, then the hosted adapter context and the forwarded identity staging file; each refusal is 16 before preflight.
2. **Preflight** records the Docker start inventory, attributes the checkout, selects the profile, and runs `infra/ci/runner-preflight.sh` for it (tools, Docker and Compose versions, BuildKit, loopback and bind-mount parity, ports, capacity, timeouts). It requires a clean tree. **Attribution** uses only documented GitLab variables and yields one of:
   - `branch-push`;
   - `branch-schedule`, for the default branch and the protected `identity-qualification` branch;
   - `branch-web`;
   - `branch-api`: a branch pipeline started through the pipelines API, attributed like `branch-web` (branch and ref must match, no merge-request metadata) and recorded with source `api`;
   - `mr-source`: detached merge-request pipelines test the source content;
   - `mr-merged-result`: only when GitLab supplies both branch SHAs and the merge commit has exactly those parents;
   - `mr-rc-promotion`: a merge request whose source branch is `release/...` is verified or refused, never treated as `mr-source`. It needs a detached event; the merged-result SHAs, `CI_COMMIT_BRANCH` and `CI_COMMIT_TAG` empty (GitLab leaves them empty there; a value is contamination); the default branch as target; source `release/<canonical RC>` equal to `CI_COMMIT_REF_NAME`; `refs/tags/<RC>` annotated and peeled to `CI_COMMIT_SHA` = HEAD; and `refs/remotes/origin/staging` containing it. The refs come only from the runner's credentialed fetch (`GIT_FETCH_EXTRA_FLAGS` in both jobs' promotion rule); a missing ref is refused, with no fallback fetch. The source head is proven offline from the merge request API by `release-verify.sh qualified`;
   - `local`.

   Merge trains, tags and other sources (for example trigger pipelines) fail with 16.
3. **Harness:** `infra/validation/run.sh` runs the profile's suites as its own process group under a watchdog deadline (the budget above). A signal or the deadline stops the whole group with a bounded grace. While it runs, the job log shows `N/TOTAL ICON STATUS CHECK-ID` (for example `27/67 ✅ PASS B-SHELL` or `13/93 ❌ FAIL B-DEPSCAN`; no `validate:` prefix on these lines; ✅ for PASS, ❌ for FAIL and UNMET) the first time each required check of the executed suites records PASS, FAIL or UNMET (TOTAL = those required checks, each counts once, so N never exceeds TOTAL; other rows are not shown; informational only, the final status stays authoritative).
4. **Finalizer** (always, exactly once) runs, in order:
   - postcheck (no drift);
   - owned-resource cleanup;
   - raw evidence scan (literal secrets, gitleaks, forbidden names);
   - completeness check;
   - all-or-nothing staging of an allow-listed evidence set;
   - a pre-publication summary;
   - the supplemental JUnit report (below);
   - a scan of exactly what will be published;
   - publication, report and scrub.

   Completeness means exactly the required checks of the executed suites plus `final`, each once as PASS, with those suites' reports: contract reports only when `contracts` ran, and browser reports for each executed `root`, `off`, `rename` and `feature-web`. A row from a check outside the executed suites fails, so a narrower result cannot pass as a wider one. validate.sh also re-derives the job's partition from the selection, and an inconsistent one is refused with 16.

**Publication:** exactly one of the following:
- *full*: the scanned evidence set;
- *minimal*: a marker and the summary, when a finding withholds the evidence;
- *none*: after any scanner failure.

**Exit status**, in order of precedence:
1. 16: location, interrupted-run recovery (local runs), preflight or attribution;
2. otherwise a nonzero harness status;
3. otherwise a signal;
4. otherwise 17 scanner failure, 15 unsafe evidence, 14 incomplete evidence, 13 drift, or 18 cleanup incomplete;
5. any other recorded failure: 19.

The job's exit status and the final log line are authoritative; the artifact summary is labelled pre-publication. **Reuse:** a release may reuse a full pass only when all of these hold:
- **a full pipeline, not a single job:** `bash infra/ci/pipeline-union.sh FAST_DIR GATE_DIR` passes on both jobs' published artifacts;
- both GitLab jobs succeeded;
- each job's tree SHA and `images.txt` equal the candidate's (the pins are a function of the tree).

A release-candidate promotion to the default branch reuses exactly one proof: `bash infra/ci/release-verify.sh qualified ...` on the promotion merge request's latest, freshly successful pipeline. It binds the MR (same project as source and target, `release/<RC>` to the default branch, open, squash off, no queued auto-merge, head = RC peel), the pipeline and its two required jobs, both artifacts' project/pipeline/job/MR/RC/tree attribution, and coverage recomputed from the candidate's own selector, map, suite lists and registry. Then `promotable`, `recheck` (fresh success, unchanged qualified jobs) and `promoted` verify the actual merge's parents, tree and presence on the freshly listed remote default branch; `final pre|post` handles a separate final tag. No pipeline and no further full run follows the promotion.

Image inventories and build ledgers differ by partition by design. They are kept per job, never compared across jobs.

**Owned-resource cleanup** (`infra/ci/cleanup-owned.sh`) removes only resources with positive ownership evidence. That evidence is recorded in a job-unique, mode-700 registry beside the checkout, and covers:
- Compose containers under the checkout or the harness work directory that were absent at start;
- standalone containers bound by a creation-time cidfile;
- networks and volumes captured from owned containers;
- images with an exact build or capture record.

Labels and names alone never authorize removal; unproven resources are kept and reported (exit 18). `after_script` runs the same cleanup as a non-gating backstop.

**Local report retention.** Run local validation with `--evidence-dir <private-directory>`. After finalization, the
wrapper retains its scanned artifact set under `<evidence-dir>/candidate/`: `results.tsv`, `ci-controls.log` and other
selected suite logs, `harness.log`, bounded failure `diagnostics/`, `summary.json`, `phase-status.tsv` and supplemental
`junit.xml` when available. Successful runs retain the same scanned set. Root-level `storage.tsv`/`recovery.tsv`
compatibility reports remain. Raw logs, generated secrets, state and temporary harness work are still scrubbed. If
scans withhold evidence, only the safe minimal candidate remains; a scanner failure retains no unscanned candidate.
Check the report path/files before a retry, and preserve failed outcomes. JUnit includes check and wrapper-phase
failures; its failure count is not the number of independent underlying failed tests. The process exit and final
status line remain authoritative.

**Local-run storage discipline.** The same owned-resource rule applies to local
runs. A completed or interrupted local validation must run the ownership-registry
cleanup, and the launcher must record both its exit and a zero-leftover postcheck.
Owned images and owned volumes created for the run must be removed; do not rely on
Docker label or name matching alone, and do not use broad `docker system prune`,
`docker image prune`, or `docker volume prune` because those can remove unrelated
user resources. If cleanup cannot run or reports owned leftovers, stop and resolve
them before starting the next batch. Repeated local batches must not accumulate
images, networks, volumes, or work directories.

**Interrupted local runs.** A catchable interruption (TERM, INT, HUP or the deadline) stops and reaps the harness process group before the ownership cleanup. The run keeps the interruption's status; a clean cleanup never turns it into a pass. A crash or an uncatchable KILL can skip finalization, so `validate.sh` tracks every local run:
- **State root.** Recovery state lives on persistent private storage, the *state root* (`G003_STATE_ROOT`; by default a `platform-evidence` directory under `${XDG_STATE_HOME:-$HOME/.local/state}`, see `infra/ci/validate.sh`). It must be an absolute path outside the repository, not a symbolic link, mode 700 and owned by the operator, and not on tmpfs or ramfs (checked before anything is created and again after). `validate.sh` creates it mode 700 when absent and exports it as `G003_STATE_ROOT`, so the harness reads the same marker. A failed check is location 16. The default evidence directory (and with it the run's ownership registry, `<evidence>/owned`) is created under the state root, never under `/tmp`; a local `--evidence-dir` on tmpfs or ramfs is refused (16). CI is unchanged.
- **Marker.** A local run writes a marker file (`<job>.tsv` in the state root's pending directory) before it touches Docker. The marker names the run's registry, evidence directory, PID and PID start time, the boot identity (`boot_id`, `/proc/sys/kernel/random/boot_id` as 32 hex digits) and the checkout (`repo`), and its launch stage (`prelaunch`, `launching`, `launched`). Once launched, it also records the harness process group and the watchdog, each with its start time (or `none` when it had already ended). It holds no resource rows: the registry stays the only ownership record. The finalizer removes the marker when cleanup verified and the registry was removed. When the registry is kept, the marker stays with a `finalized` row. Markers of earlier versions under `$TMPDIR` are still read (same-boot rules), never written.
- **Watchdog.** The deadline watchdog closes every inherited descriptor above stdio, so it never holds a launcher's lock. Its only child is the `sleep` timer for the run's budget. Once a second it checks that its parent is still the same process (PID and start time). If not, it stops the timer and ends, so a KILLed `validate.sh` leaves it behind for about a second at most. At the deadline it signals only that verified parent, never a reused PID.
- **Recovery gate.** Before preflight, every local run resolves each earlier marker with the existing `cleanup-owned.sh --mode finalize` on that run's registry. A verified result is a *delayed recovery* for the original run. It writes `recovery.tsv` (original job, times, recovering job, inventory rows and verdict) into the old registry and the new run's evidence, removes the old run's generated-secret list and its marker, and lets the new run proceed. A marker whose `boot_id` differs from this boot's belongs to a run whose processes ended with that boot: the PID, survivor and `launching` checks are skipped and the unchanged Docker, registry and `cleanup-owned.sh` path decides (verified: delayed recovery, marker removed; anything else blocks). Such a run is never treated as clean without that verification.
- **Blocking.** The launch is blocked (16, phase `recovery`) and the marker kept when:
  - the marker is unreadable or malformed, or its `boot_id` or `repo` row is missing or malformed, or this boot's identity cannot be read;
  - the marker names another checkout (`recover it, or wait for it, from that checkout`);
  - that run is still alive (same PID and start time);
  - it ended while launching its harness (stage `launching`), so its processes were not recorded;
  - its recorded harness or watchdog identities are malformed;
  - its harness group or watchdog is still alive (same leader and start time), after a bounded 5 s wait for the watchdog to end and a KILLed group to be reaped;
  - a recorded group still has live members but its leader has ended, so they cannot be attributed (a group number can be reused);
  - Docker is unavailable;
  - the registry is missing, invalid or another run's;
  - cleanup is refused or does not verify every row.

  Nothing is guessed from names or labels, and recovery never signals a process. A live PID that is outside the recorded group, with another start time, is not the run's and is ignored.
- **Operator resolution.** For a surviving harness or watchdog, check the named process group (`ps -o pid,pgid,lstart,args -g <pgid>`) against the marker. Then wait for it, or stop exactly that group, and rerun; the next run recovers the registry. Do not stop a group you cannot attribute. For resources, read the job log line and the registry's `recovery-inventory.tsv`. Then either finish the cleanup with `bash infra/ci/cleanup-owned.sh --registry <registry> --mode finalize --out <file>` from the run's checkout, or, for evidence that is gone, decide on the reported resources by hand. Only then delete the marker.
- **Scope.** CI jobs write no marker: each job has a fresh daemon and the `after_script` backstop. Direct `infra/validation/run.sh` and `run-approved.sh` runs have no ownership registry. They keep the harness's own inventory cleanup (`inventory-leftover.txt`), and their launcher's postcheck stays responsible for them. Without a registry the harness also starts no batch runner (every Go invocation takes the per-invocation container path) and makes no resource-accounting claim (it reports accounting as unavailable). A direct top-level `infra/ci/validate-controls.sh` run registers itself like a local `validate.sh` run (see **Resource accounting and recovery for CI controls** below), so a crash of it is recovered by the next local launch.
- **Reboot.** Recovery after a reboot is designed (the `boot_id` rule) but not yet proven by a real reboot; the controls prove it by rewriting a marker's `boot_id`.
- **Launching.** Do not start a batch under `nohup`: a shell that starts with SIGHUP ignored cannot trap it, so a HUP cancellation would not reach the run.

**Storage report.** `validate.sh` observes Docker storage read-only before the harness and again after cleanup, including failed and interrupted runs, into `storage.tsv`.
- **Rows.** Each row carries the observation time, phase (`pre` or `post`), job, candidate, scope, metric, value, unit and status.
- **Metrics.** `docker system df` gives the daemon's image count, size and reclaimable size, and the build-cache record count, size and reclaimable size of the daemon's built-in builder. Other buildx builders are not observed, and none is bootstrapped. `post` adds the run's cleanup status and its owned resources remaining.
- **Values.** Sizes are the Docker CLI's own figures (decimal units). A failed or timed-out command (bounded at 60 s), a missing row or a malformed figure reads `unavailable` with an `unavailable:<reason>` status, never zero.
- **Limits.** These are daemon-wide and builder-wide observations. Shared layers and concurrent users mean they are not this project's usage, and a pre/post difference is not bytes freed. Retained build cache and shared base images are not run-owned leftovers. Leftover owned images would show in the cleanup inventory, and the Docker Desktop mount-table leak (below) is a separate condition. There is no size threshold and no automatic eviction.
- **Publication.** The report is not a phase and cannot change the run's status. It is scanned with the raw evidence and the candidate like any other allow-listed file. CI publishes `ci-evidence/storage.tsv`; a local run with `--evidence-dir` keeps the scanned copy (with `recovery.tsv`) in that directory after the scrub. After clean scans the job log prints one validated summary line per phase and says where the report is.

**Controls.** `infra/ci/validate-controls.sh` runs in scratch trees; `validate.sh` has no fault-injection switch. It covers:
- gate semantics and profile completeness;
- attribution;
- evidence scanning;
- finalization, supervision, budgets and capacity;
- Docker-real ownership, the storage report and local interrupted-run recovery, including process survivors and descriptors after a KILL (VS1-VS5, VR1-VR8, and VR9 for crash recovery of a reuse runner and of a self-registered controls run);
- profile selection on fixture repositories;
- the hosted adapter, in the pinned CLI image with a docker shim and a tmpfs build directory.

As root, the launches drop `CAP_DAC_OVERRIDE` and `CAP_DAC_READ_SEARCH`, so file-mode faults behave as they do for an unprivileged user.

**Artifacts:** `ci-evidence/`, uploaded `when: always`, developer access, 90-day expiry. `validate:checks`, `validate:integration` and `identity-staging` also declare `ci-evidence/junit.xml` as `artifacts:reports:junit`.

**Failure logs.** A fully scanned publication includes `harness.log`. Before cleanup, failed runs export text command logs, command lists and recovery logs to `diagnostics/`, including `off-check.log` and `off-stop.log` for toggle-off failures. Known generated secret literals are redacted before bounding each diagnostic to its first and last 512 KiB; logs up to 1 MiB remain complete. Download the failed job's `ci-evidence/` archive and read the corresponding log for the assertion, compiler or command error behind a failed row. Existing raw and candidate scans both still gate publication: a secret finding, scanner error or publication failure may withhold these logs, which the phase statuses distinguish. No environment files or database dumps join diagnostics. Cancellation or hard runner termination may prevent finalization.

**Test reports.** GitLab's pipeline Tests tab and merge-request test summary read `ci-evidence/junit.xml`:
- **Scope.** It lists harness-level check IDs (for example `B-SHELL`), not the unit, contract or browser assertions inside a check; use the job log and the published evidence for those.
- **Format.** There is one suite named `JOB:PROFILE`, for example `fast:fast` or `gate:remainder` (internal tokens). The profile is `none` when nothing executed: a gate in a fast pipeline, or a run stopped before the harness. Each check `results.tsv` reports is a case of class `validate.JOB.check`, in `results.tsv` order. Each phase that had already failed when the report was built (for example the harness, completeness, staging or a scan) is a case named `phase:NAME` (class `validate.JOB.phase`). A failing phase therefore stays visible even when no check failed.
- **Later phases.** Failures after the report is built (publication, the job-log report or the scrub) appear only in the job status.
- **Status semantics:**
  - **PASS** passes.
  - **FAIL** fails with the message "check failed".
  - **UNMET** fails with "required prerequisite unmet".
  - **INFO** rows (PREFLIGHT, PROBE, IMAGES, B-SCOPE-EXCEPTION) are omitted, never shown as skipped.
  - A check that was not executed is absent; nothing is inferred from `required-checks.tsv`. A row that the harness does report outside the executed profile is shown as reported, and completeness also fails for it (`phase:complete`).
- **Content.** `validate.sh` builds it (`infra/ci/junit.sh`) only from fields it has already validated:
  - the check IDs and statuses of `results.tsv`;
  - the numeric phase statuses;
  - the job and profile tokens;
  - the run's measured validation elapsed seconds, when executed-suite timing rows exist.

  **Timing.** GitLab's test-report totals sum case times, and suite and phase scopes overlap (background suites run concurrently; the harness phase contains them), so suite and phase durations are never cases or case times. One explicitly named case, `timing:validation-elapsed` (class `validate.JOB.timing`, always passing), carries the only `time`: the job's measured validation elapsed seconds, one wall-clock interval of `validate.sh`, never a sum. It is present exactly when the harness reported timing rows for executed suites. No real check or phase case carries a time, and check and phase failures and their counts are unchanged. The fixed-shape self-check admits a time only on that case and at most once. The per-suite breakdown stays in the diagnostics: `timings.tsv` (start/end per suite) in the published evidence. Individual check rows do not have measured start boundaries; check completion timestamps are never treated as runtimes. The job's actual duration remains GitLab's native job metric. A WITHHELD report includes no timing case.

  It never reads free-text detail, paths or URLs. It joins the candidate before the publication scans, and a WITHHELD publication carries phase cases only. Reports therefore contain no sensitive or raw detail. Artifact access and retention are unchanged; the report's UI/API visibility follows GitLab's test-report rules, not the archive download restriction.
- **Supplementary, never authoritative.** The job status and the final job-log line decide.
  - **Report defects** (an invalid sanitized input, a repeated non-INFO check ID, or a written file that fails the fixed-shape and yq self-check) publish no `junit.xml`. On an otherwise passing job the job then fails with 19; on a failing job the original status stands.
  - **Raw `results.tsv` defects** keep the completeness status (14). A malformed line or unknown status is dropped, and the report shows `phase:complete` with only the well-formed rows. A repeated check ID leaves no report.
- **Limits.** A location failure, or a preflight failure before the scanner and yq images are available, publishes no report; the 16 and its log line remain authoritative. A later preflight failure publishes a phase-only report. A job GitLab cancels or times out may upload no artifacts, so its pipeline may show no test report; the internal deadline (124) normally lets the finalizer publish first.

**Local commands.** The README `platform:ci` (full), `platform:ci-fast`, `platform:ci-integration` and `platform:ci-affected` blocks run the same entry point. Locally the profile is explicit, `--profile gate` is refused, and nothing is published. The full, fast, integration and affected profiles are developer checks that need a Git checkout (an exported copy without Git history can be given a fresh local repository first). Two options apply only locally, with `--profile affected`; in CI they are refused with 16:

- **`--base REV`** classifies `REV..HEAD`. REV must be an ancestor of HEAD; otherwise the run is full.
- **`--compare OLD`** classifies the tree difference between a previously validated candidate and HEAD. It is meant for post-rebase reconciliation:
  - it names exactly the suites to rerun;
  - the other suites' earlier evidence may be reused only with exact tree and input provenance;
  - an unknown OLD or a failed diff gives full.

**Timing definitions.** These definitions are used for the cold pipeline targets (5, 15 and 45 minutes):

| Term | Definition |
|---|---|
| Required jobs | The jobs that run automatically in a pipeline context. An MR's optional manual gate is not one. |
| Wall time | max(`finished_at`) − min(`started_at`) over the required jobs |
| Queue | The sum of `queued_duration`, reported separately and excluded |
| Compute | The sum of `duration` × 2 (the medium runner's factor) |
| Cold | A fresh hosted job VM and DinD with no project cache. Hosted jobs are cold by construction, since no GitLab `cache:` is configured. |
| Warm | Local daemons with BuildKit cache; never acceptance evidence |

Only successful, non-retried runs qualify. Canceled, failed, timed-out or incomplete runs are kept as history only. Measurements use read-only `glab api` GETs. These targets have not been measured for this exported copy.

**Build ledger.** Each job writes `builds.tsv`, one row per build invocation, and `ledger-summary.tsv`. They prove at-most-once building per job; they never gate merges.

- **Key.** Each row's key is built from existing Docker metadata only:
  - the Dockerfile path and SHA-256, the target, and the `syntax` frontend identity;
  - the resolved base images, cross-checked against `images.txt`, and the platform;
  - the effective ignore file. `<Dockerfile>.dockerignore` takes precedence over `<context>/.dockerignore`; both are recorded when both exist;
  - a NUL-safe inventory of the whole build context (entry type, mode, owner, path, symlink target, file SHA-256), plus named contexts and arguments.
- **Step counts.** They come only from parsed BuildKit plain progress. The harness exports `BUILDKIT_PROGRESS=plain`, and explicit builds use `--progress=plain --iidfile`.
  - A confirmed cache hit means every step is `CACHED`.
  - Absent, unparseable or unexpected progress is `unknown`, and so is an unknown key field or an image change with no row.
  - **Compose Bake progress (C3).** The hosted toolchain (the CLI plugins of the docker CLI image that `infra/docker/ci-context.Dockerfile` pins: Compose v5.5.1, buildx v0.37.2) builds Compose services through Bake. Its plain progress has no `#0 building with` header, and a single-service build has no service prefix. The ledger supports exactly this captured format:
    - Each invocation's `[internal] load local bake definitions` vertex opens its session. The driver comes only from `docker buildx inspect`, read before and after the call (it resolves `--builder`, `BUILDX_BUILDER` or the selected builder as Bake does). Missing or changed builder evidence leaves the row `unknown`.
    - An unprefixed session is attributed only to the one service of the call or block whose image its one completed export names on its `naming to` line. Otherwise it is an `unknown` row and its predicted builds stay missing.
    - A multi-service session is prefixed. A service whose build definition Bake printed once for several services may use that vertex only when its Compose context and Dockerfile equal theirs and exactly one such vertex exists. Its own completed export is still required.
    - A vertex counts once per session however often BuildKit re-prints its name (C4). In a multi-service session a shared vertex may be re-printed under another service's prefix; only that prefix may differ (same stage, index and operation), and each block keeps its own service row (C5). Any other name change is `unknown`.
    - The frontend vertices carry only the service as their label (`[S] resolve image config for docker-image://…` and `[S] docker-image://…`, the form an observed daemon printed). Such a vertex may be re-printed the same way only when its text is identical and both services are build targets of the call (from its Compose plan) that print their own export in that session (C6). The reserved `internal` and `auth` tokens, other services, a missing or longer label and other operations stay `unknown`. This normalizes an observed format; it is not shown to explain every hosted `unknown` row.
    - VC9 builds with that pinned toolchain, replays the captured streams, and keeps every derived defect `unknown` and UNMET.
- **Rules.** Each key may execute at most once per job. A cache-only prebuild counts as that build. The named exceptions are:
  - `prebuild-fallback`;
  - B-BUILD-CONTEXT;
  - B-WEB-ERROR-OUTPUT;
  - the web-browser fixture stage;
  - B-TOGGLE-OFF.
- **`ac5_status`.** It is MET only when nothing is `unknown`, every expected invocation has exactly one row, and there is no unexcepted duplicate or near-duplicate executing build.
- **Unknown-vertex context (C6).** When a progress reason names a vertex (`#N`), `ledger-summary.tsv` keeps that vertex's print blocks as `unknown_context` lines, so a failed run shows the actual form:
  - vertex names, terminal lines and BuildKit status lines are kept verbatim; log lines become `<log>`, layer progress `<layer progress>` and anything else `<other>`; repeats are counted;
  - bounded to the first and last 10 entries of at most 6 vertices per session, 200 characters per line and 120 lines per job;
  - raw progress stays private. The published file passes the existing evidence gates (generated-value literal scan, gitleaks, scrub); V5d plants a canary in such a line and expects the evidence withheld;
  - `ac5_status` and `pipeline-union.sh --ledger` ignore these lines.
- **Context image.** The hosted adapter builds its context image once per job, before the harness runs, so it gets a `bootstrap` row:
  - it is counted as one executing build, with step counts `not_observed`;
  - validate.sh compares the image ID before and after cleanup (`ledger-bootstrap.tsv`).

  Equal IDs support only a **structural inference** that the after-script reuses the image; this is never step or cache evidence. An absent or changed ID is UNMET. `bash infra/ci/pipeline-union.sh --ledger DIR` re-checks a published job.
- **Scope.** Builds inside the controls' scratch trees are the controls' own scenarios and are excluded by definition.

**Background suites.** When `controls` is selected together with a later suite, it runs as a supervised background child of the harness, concurrently with the stack suites:

- **Isolation.** The child has private results, timings and log. It uses scratch roots outside the repository and `$WORK`, run-token labels, no host ports, and never writes the checkout.
- **Join.** The parent joins it before the `final` checks. It captures the exit status and merges only well-formed rows that belong to the controls suite, recording `BACKGROUND INFO`. A nonzero exit, or foreign or malformed rows, records `BACKGROUND FAIL`, which fails the run.
- **Every exit path** (failure, timeout, cancellation) runs `bg_stop_all` before ownership cleanup. It sends TERM to the child and its verified descendants, waits a bounded 60-second grace, sends KILL, then reaps them. Rows from a child that never joined are never merged.
- **Prebuilds.** Cache-only prebuilds in affected runs use the same supervision. A failed prebuild records `PREBUILD INFO`, and the authoritative build always follows. The web-check build (unit, unit-web) waits for a running web-browser prebuild first, since BuildKit merges identical in-flight vertices of concurrent builds under the other build's stage label and index (C4).
- **Browser workers.** The harness's own browser runs use `PLAYWRIGHT_WORKERS=2`, except the identity E2E run, which
  keeps 1 worker because its light and dark projects share the seeded users and sessions. The README browser command
  keeps Playwright's default of 1 worker.

**Focused debugging.** Fix a failure with targeted retests before any broad run, then run the required final profile once on the rebased candidate. For CI-machinery changes, focused control groups are:

| Change | Control groups |
|---|---|
| `select-profile.sh` or the map | F (and A VP*) |
| `validate.sh` | A, B, C, D, J |
| `cleanup-owned.sh` | E |
| `hosted-context.sh` | G |
| `junit.sh` | D (J*) |
| `infra/ci/lib.sh` | all |
| ledger, background or `pipeline-union.sh` | VC*, VB*, VU*, VA* |

A change to the CI machinery itself always ends with a full run, which includes every control. Ordinary application changes run their affected product checks and only the directly affected control cases (`bash infra/ci/validate-controls.sh --only <IDs>`).

**Limitations.**
- The intelligence-disabled variant carries no feature routes.
- Wiring a new core package into the server still changes `core/cmd/server/main.go`, which is pinned by B-MCP-SOURCE and selects full.
- No area exists until one is registered, so current changes keep integration or widen to full.

**Cost.** Pushes start no pipeline at all. Automatic MR pipelines (except the release-candidate promotion merge request, which is always full) are omitted entirely when all changed paths are documentation-class (root Markdown files, Markdown under `docs/` or the project-management directory, except nested Claude skills, and the exact docs-class files above): no fast job, gate job, runner or Docker bootstrap. Any other path under the root agent directories admits the pipeline. Mixed commits classify those Markdown paths as docs, so the other changed files choose integration/full. Non-Markdown documentation controls (for example scripts), infrastructure, dependencies and test contracts retain full validation. Runtime-only changes select integration; changes confined to registered feature areas select the affected profile. Manual web/API requests, default-branch schedules and manually started merged-result MR gates retain full validation. GitLab's changed-path evaluation limits may conservatively create an MR pipeline; the selector still fails closed for missing/invalid bases. These rules affect CI execution only, not review or acceptance practice. Compute minutes come from the namespace's shared monthly quota.

A documentation-only MR can still be blocked by GitLab's project-wide "Pipelines must succeed" setting when no pipeline exists. The bypass does not change that setting or allow skipped code pipelines to merge. Request explicit validation when required for an MR.

**GitLab project settings** (outside this repository; adopter decisions):
- "Pipelines must succeed" (whether pipelines block merges);
- the role allowed to override pipeline variables: selection trusts GitLab's predefined variables, so anyone who may override them (for example `CI_COMMIT_BEFORE_SHA` through a push option) can narrow a push pipeline's profile; the base checks only reject missing, non-ancestor and degenerate values;
- `staging` protection;
- a weekly default-branch schedule;
- artifact retention beyond 90 days.

### Per-test Go evidence

**Fresh database per invocation.** In `migrations`, every independent Go integration invocation starts on a fresh copy
of the run-private, write-closed template (`db_fresh`; [MIGRATIONS.md](MIGRATIONS.md#validation),
[DATABASE-ISOLATION.md](DATABASE-ISOLATION.md#test-databases)); persist pairs get one before the seed, and the
recovery drill keeps its own chain (its `down -v` removes template and copy). In `ops`, B-IDENTITY-IT runs inside the
identity isolation window ([IDENTITY.md](IDENTITY.md)). A refusal (stale template, shared-catalog delta, a client
session, which is never terminated) means the invocation does not run and its check fails with the reason.

**Reporter.** `go_it` (`run.sh`; `rls_go` remains a compatibility wrapper) runs, in one container, `go test -json -list`
for the expected top-level tests per `(package, test)`, then `go test -json -run`, and keeps the container's exit status.
`gotestreport` (validation tools image) rebuilds the exact `-v` text, so every existing count and PASS-text check reads
the same log as before, and writes per-test rows. An invocation passes only with exit 0, a converter PASS and the
existing count checks. The converter fails closed on: a selected package without exactly one terminal event; build or
setup failure; an expected test that never ran; a started test or subtest without a terminal; duplicate or
contradictory terminals; any skipped `Test*`; an undeclared test, or one in another package; malformed JSON; a package
`pass` with a non-zero exit.

**Retained reports** (in `REPORT_DIR`, exported after each invocation, not only at suite end):
- `go-tests-expected.tsv`: `suite<TAB>invocation`, the declared inventory written before the suite's first invocation;
- `go-tests.tsv`: one `invocation` row (status, exit code, expected/passed/failed/unfinished counts, provenance: run,
  check, suite, tested tree hash, HEAD with a head/dirty flag, base as baseline only, image ID, package, tag, selector,
  timeout, command, and the lifecycle row the copy came from), one `package` row per selected package and one row per
  test or subtest with its log line range;
- `go-tests-junit.xml`: one suite per invocation, supplementary (the authoritative `junit.xml` is unchanged);
- `go-tests/<NAME>.log`: the rebuilt `-v` log with every registered generated secret replaced by `[REDACTED]` and
  bounded to 1 MiB (first and final 512 KiB), the same `redact_log` as `diagnostics/`; the raw JSON stream never leaves
  the work directory;
- `db-lifecycle.tsv`: every pristine check, template, fresh copy, refusal, drain wait and residue row with purpose,
  timestamps and durations, exported after each row.

The raw and candidate scans gate these files like every other evidence file; B-SECRET-CANARY is unchanged.

**Interruption.** On TERM (the wrapper's group stop sends it first) the harness's EXIT trap converts every started
invocation that has no row yet as `interrupted` (observed terminal rows kept, started tests `unfinished`) before
diagnostics and cleanup; the run stays failed. KILL cannot run the trap; the invocations already exported stay, and the
missing rows are findings.

**Completeness (`validate.sh complete`).** When `migrations` or `ops` executed, the four files must be regular,
non-empty and well-formed; every declared invocation of an executed suite needs exactly one `invocation` row with
status `pass` or `fail` and an exit code; interrupted, unfinished, duplicate, missing or undeclared invocations, count
mismatches and missing `go-tests/` logs are named findings (status 14). `results.tsv` and the required-check rules still
decide PASS. Controls GT1 (missing `go-tests.tsv` → 14) and GT2 (a failing and an interrupted invocation → `complete` 14
naming the interrupted one, the FAIL row reported and a planted generated secret withheld by the raw scan, exit 15)
are part of B-CI-CONTROLS.

### Batch Docker reuse and resource accounting

**Scope.** `migrations` reuses one long-lived runner container for its eligible Go invocations (batch B-MIG). PostgreSQL
and MinIO were already shared per suite. Every other suite and scenario keeps its own resources: PROV instances, the
recovery drill, root, ops/identity, rename, off and the CI controls. Every Docker allocation of a run with a registry is
counted by exact ID against a finite ceiling.

**Runner.** The B-MIG batch opens after `db_template_create`. The first eligible invocation starts the runner lazily: a
container named for the run, batch and generation, from the suite's core-check image (its ID must equal the template
record's `core_check_image`), `--init`, `--entrypoint sleep … 2147483647`, `--pull never`, the project's data network,
the run label and a batch label, no environment, no mount and no published port. It goes through `docker_run_owned`,
so its cidfile and `standalone` row exist from creation. Its baseline is recorded at start: started time, the host PID
of its one `sleep` child, `go env GOCACHE GOTELEMETRYDIR`. An anonymous volume or a different image refuses it.

**Exec routing (`go_it`).** An invocation runs in the runner only when the batch is open, its network is the runner's,
its image ID is the runner's, and every Docker argument is `-e K=V` or `--env-file F`. Otherwise it takes the
per-invocation container path. `GO_IT_RUNNER=off` forces the container path. The mount-bearing persist and model
invocations and obj-sign-host keep the container path by rule. The reporter path is unchanged
(`go_it_begin` → `go test -list` and `-run` → `gotestreport`): rows, log and export per invocation. Each invocation still
gets its own `db_fresh` first. The provenance (`go-tests.tsv`) adds:
- `runner=<12 hex>` on the exec path;
- `runner=container` with `runner_reason=off|no-batch|network|image|docker-args|no-registry|no-durable-recovery` on the
  container path;
- `runner=none` with `runner_finding=<reason>` for an invocation that was refused and did not run (`batch held: …`);
- `runner_finding=` for a runner fault or residue, and `res_finding=` for an accounting finding at that boundary.

**Per-exec checks.** Before each exec the runner must be running with the same ID, start time and `sleep` PID. Its
process set must be exactly `/sbin/docker-init -- sleep …` plus that `sleep` (`docker top`). `docker diff` may change
only the Go build cache, the telemetry directory and `/tmp`, and `/tmp` must be empty. After the exec, identity, process
set and diff are checked again before anything is reset; only then is `/tmp` emptied and verified. A changed identity
is `runner-fault`; a lingering process or a file outside the allowed roots is `runner-residue`. Either way the
invocation FAILs whatever its exit code, and the runner is removed at once with verified teardown. The next eligible
invocation starts a new generation. A second fault in the batch holds it ("repeated runner fault"). Controls T-9
(residue) and T-10 (runner removed during an exec) are part of B-IT-ISOLATION.

**Teardown and holds.** Verified teardown is `docker rm -f <exact ID>`, then `docker inspect` must report "No such
object", then the batch label query must succeed empty now and again 6 s later. A late hit is removed and still
holds. The runner is stopped this way at every named batch end (before the recovery drill, in the MinIO-failure branch,
before the final `compose down -v`) and on EXIT (`bg_stop_all`, `it_runner_stop_all`, then `go_tests_finalize`). An
unverified teardown writes the batch's `held` state and a `RESOURCES FAIL` row. The remaining B-MIG invocations FAIL
`batch held: …`. The recovery drill is held too, because it resets the same project and network. Unaffected suites
continue.

**Durable-recovery precondition.** The runner starts only with a valid run registry and a durable recovery
registration. In CI that is the job registry of this `CI_JOB_ID`. Locally it is the pending marker in the state root,
naming this registry, with stage `launching` or `launched`. Without it, every invocation takes the container path with
`runner_reason=no-registry|no-durable-recovery`. There is no deferred cleanup: a crash leaves the runner as an ordinary
registry row for the next launch's recovery gate (above).

**Ceilings, latch and reservations.** The containers/networks/volumes ceilings (source-derived, enforced, never
report-only) are:

| Scope | Ceiling |
|---|---|
| B-MIG outside PROV / PROV window | 5/1/2 / 9/1/7 |
| S-REC (recovery drill) | 10/2/10 |
| S-ROOT, S-OPS, S-IDN, S-REN, S-OFF | 11/3/3, 8/3/3, 15/3/4, 8/3/3, 8/2/3 (S-OFF: 6 stack incl. the exited minio-init + retained browser one-off + capture yq) |
| instrumentation (the witness) | 1/0/0 |
| run-wide (and scopes without their own row) | 63/5/12 |

Allocation helpers reserve capacity under `flock` before they create anything:
- `docker_run_owned`/`docker_create_owned`;
- Compose `up|run|create|start|restart` through `compose_call` (latch-checked, then recounted);
- documented blocks, the recovery tool;
- the runner and the witness.

A crossing or a failed Docker query refuses (exit 125, nothing created) and latches the scope. The latch is a file under
the registry directory, `hold.d/<scope>` or `hold.d/run`, so it also holds in subshells. Teardown is never gated. At a
scope's end the foreground container count must be a verified zero: then the latch is lifted (`held.d/`). For a nested
scope (PROV, S-REC, S-IDN) that zero is counted beyond the containers live when it began, which belong to the enclosing
scope (the migrations stack); an unreachable daemon or a missing baseline leaves the count unverified. Otherwise
`hold.d/run` holds every later allocation for the rest of the run. Each refusal is appended to `held.log`. A latch is a
`RESOURCES FAIL <scope>: <reason>` row: non-required, but any FAIL row fails the run. `RES_CEILING_LOWER=<scope>:<n>` may
lower a ceiling for a proof; an override that would raise one is refused. A background `ci-controls` child's networks and
volumes (label `sdlc.ci.run`, mirrored as bare `network`/`volume` registry rows) are owned, cleaned and sampled run-wide,
but, like its containers, they are never a foreground object of any scope or ceiling (control VB8).

**Witness and events.** With a registry, `res_witness_start` runs before the suite loop. It starts one `busybox` witness
container W (`--network none`; refused if the image declares volumes) and one `docker events` stream for container and
volume `create|destroy|mount|top` events, with no label filter, read by a supervised parser.
- **Containers.** A container is attributed from its own `create` event: a registered project, a working directory under
  the checkout or work directory, and absent from the start inventory. It gets a registry capture row at once.
- **Volumes.** An anonymous volume is attributed only from its own `mount` event naming an attributed container.
- **Markers.** After each witnessed Compose call or block, and at scope boundaries, `docker top W` serves as a marker:
  the caller waits for the parser's acknowledgment.
- **Volume removal.** A bound anonymous volume whose container is gone must be verified removed (`docker volume rm` when
  needed). A witnessed re-`create` under the same name (Compose `down -v` then `up`) voids the earlier `destroy`: the
  volume's own labels decide again at the window (a Compose project label is `named`, project teardown).
- **Failures.** A marker timeout, a stream end, a lost W or an image volume never seen mounted (`unbound-volume`)
  latches. None is ever reported "unknown".

Files: `$WORK/res-witness/` (journal `events.tsv`, `win-<n>.tsv`, `attr.tsv`, `bound.tsv`, `gone.tsv`).

**Reading `resources.tsv`** (in `REPORT_DIR`; staged by `validate.sh`; `complete` 14 when the harness ran with a registry
and the file is missing):
- `total` rows, one per scope and class: `creations` (unique IDs from the registry; measured, exact),
  `observed_max` (sampled at helper returns, Go invocation boundaries and scope boundaries; a lower bound of the true
  peak, never called a peak), `ceiling` (source-derived bound), `final_leftover`, `unaccounted` (labelled objects without
  a registry row, run row only) and `status`. `FAIL latched;`, `observed_max>N;`, `final_leftover;` and `unaccounted;`
  name the cause. Images are `reported`, never failed.
- `created`, `gone`, `pulled` and `finding` detail rows give the full ID, scope and purpose or mount relation.
- Scope boundaries are registry line offsets in `$WORK/res-scopes.tsv`; samples are in `$WORK/res-obs.tsv`.

`validate.sh complete` prints the leftover count, the `resources.tsv` findings and the `RESOURCES FAIL` rows as
diagnostics before any rule can return. They never change a status.

**Diagnostics.** `diagnostics/it-runner-<batch>.log` (exported like every `*.log`) has one line per runner start,
fault, teardown and late recreation. `hold.d/`, `held.d/`, `held.log` and `rsv-reconciled.tsv` stay in the registry
directory with the registry: removed after a verified cleanup, kept for recovery otherwise.

**Resource accounting and recovery for CI controls.** `validate-controls.sh` gates each control's persisting creations
before they happen: `create`, `run -d`, network/volume create, build, commit. Each control may hold 8 containers, 2
networks, 2 volumes and 2 images (label `sdlc.ci.control=<run>-<control>`); the V18l-n nested holder may hold 39
containers (`sdlc.ci.outer`). The check is serialized under the control's lock. At the ceiling the creation is refused
(`resource ceiling N/MAX for <control>`), the control FAILs and its latch is set. After each control, every labelled
object without an inventory row is an `unrecorded` FAIL. The next control starts only after a verified clean sweep; an
unverified one reports each remaining control `held: cleanup of <control> unverified`. Every inventory row is mirrored
into the run registry: run.sh passes `CI_CONTROLS_OWNED_REG` when it has one, so a crash leaves registry-owned objects
for the recovery gate. A container creation (the control's `D()` and the case shim alike) hands the Docker CLI a
cidfile inside that registry's `cid/` directory, the binding `cleanup-owned.sh` resolves by exact ID, so a creator
killed during an attached `docker run` leaves a recoverable object, not an unbound one; a caller's own `--cidfile`
receives the ID after the call. When the registry cannot take a cidfile the creation is refused. A case's image builds
(`build`, `buildx build`, `image build`; the build ledger) pass the same gate as `D()`'s: the run latch, the control
latch and the image ceiling refuse them, and the built ID (the caller's `--iidfile`, else an injected one) is recorded
and mirrored. The image ceiling counts every live labelled image ID, dangling ones included: 2 per control, except the
build-ledger controls VC1-VC9, whose existing scenarios build up to about 16 images in one control and get a fixed cap of
24 (an operational cap for that finite fixture family, not a measured peak); `D()` and the shim use the same value. A mirrored control container (`standalone <id> ci-<control>`) is a background object of the run: it is
owned, cleaned and recovered, counted in the run-wide sample, and not counted by any foreground scope, ceiling or zero
check (ci-controls run in the background before the foreground scopes). A container that a supervised background child
of the harness creates itself (`standalone <id> bg-<child>-<purpose>`; the controls suite's `yq` runs carry the run label)
is classified the same way. Its registry row is appended only after the call returns, and an attached `--rm` run
returns after the container is gone, so the creation also carries a background-child label, atomic with the create:
a run-labelled container with that label is background from its first listing, and a `bg-*`/`ci-*` row keeps it so
even over the run label. A foreground scope's verified-zero end is therefore not tripped by a concurrent background
creation, while a run-labelled container without the label or a row, a main-shell creation and a container whose only
row is a dead owner's `reconciled` leftover (the row overrides the label) still latch.
- **Self-registration.** A top-level direct run registers itself like a local `validate.sh` run, using `validate.sh`'s
  own registry and marker functions. The run has no `CI_CONTROLS_OWNED_REG`, is not nested and is not
  `CI_CONTROLS_HOLD_SIM`. It re-executes once under `setsid` when it is not its own process group leader. Its registry is
  a controls registry under the state root, plus a start inventory and a `launched` marker.
- **Exit and finalize.** A registration failure exits 16 before any control runs. On exit the registry is finalized
  through `cleanup-owned.sh`. Only a verified finalize removes it and the marker; otherwise both stay for the next local
  launch's recovery.
- **List.** `--list` registers nothing. It prints 530 controls, including VR9 (marker/recovery entry path, cases (a)-(g)).

**Unregistered control objects.** Control objects that are not in a mirrored inventory are outside the recovery
guarantee. The startup guard holds on them (16) and never removes them. Resolve each one once:
1. List them: `docker ps -a --filter label=sdlc.ci.owner` (likewise for `sdlc.ci.outer`, `sdlc.ci.control`, networks and
   volumes).
2. Confirm that the owner named in `sdlc.ci.owner` (boot, PID, start time, checkout) is not a live run.
3. Remove exactly those IDs (`docker rm -f -v <id>`; `docker network rm` / `docker volume rm <id>`), then relaunch.

Never remove by label or name pattern alone, and never touch an object whose owner is live or another checkout's.

**Measured and unmeasured.** The resource figures in `resources.tsv` are measurements of the run that wrote them.
Nothing above claims a time saving. The runtime added by observation points, witness markers, ceiling queries and
per-helper cidfiles has **not been measured**. Not yet observed on real Docker:
- the `docker events` attribute templates;
- daemon ordering of create/mount events before the `top` marker;
- Compose `run --rm` volume removal;
- `docker top -eo`/`diff` formats;
- `flock` timing under load;
- a real reboot transition.

### Runner requirements

**Hosted jobs** need GitLab's hosted Linux runner with the medium tag: a fresh VM per job with a privileged Docker-in-Docker service. On it the context container itself stays unprivileged. A measured hosted job had about 37.9 GiB free on the 45 GiB filesystem shared by the build directory and daemon storage.

**Local runs** need:
- Docker Engine 24 or later with BuildKit, and Compose 2.24 or later;
- Bash 4.4 or later;
- GNU coreutils, findutils, and grep with `-P`;
- git, `setsid`, `ps`, `timeout` and `flock` on `PATH` (`flock` is required for resource locking);
- capacity as above.

For a local run, the user's `docker` group membership is root-equivalent on that host.

**Docker Desktop bind-mount leak (WSL).** Docker Desktop on WSL never releases a bind mount's host path. Each distinct path a container bind-mounts from the WSL distro stays mounted (four times) in Docker Desktop's VM until WSL shuts down. Removing the container, deleting the directory or restarting Docker Desktop does not release it. The VM's mount table is capped by `fs.mount-max` (100,000). Once it is full, every container start fails with `failed to mount .../rootfs/overlayfs/...: no space left on device`, although the disk and inodes have room. Docker Desktop itself then hangs on "starting".

- **Avoidance in this repository.** The evidence scan streams the evidence and configuration to gitleaks as a tar archive on stdin and reads the report back the same way, so it mounts nothing (an earlier per-directory bind mount, multiplied by the per-case scratch trees of `validate-controls.sh`, exhausted the table). New harness code should do the same, or reuse a fixed path, instead of bind-mounting a new per-run directory.
- **Check.** `bash infra/ci/docker-desktop-mounts.sh [PERCENT]` reports the mount count. It fails at or above PERCENT of `fs.mount-max` (default 80) and skips outside WSL. `validate-controls.sh` runs it before it starts. Run it yourself before a long local validation batch.
- **Recovery.** Quit Docker Desktop, run `wsl --shutdown` from Windows, then start Docker Desktop again. This stops every WSL distro and container and clears the distro's `/tmp`. Named volumes are kept. Do not do this while a run is writing evidence under `/tmp`. Compose stacks without a restart policy need `docker compose up -d` afterwards.
- Hosted runners are unaffected: each job gets a fresh VM and Docker-in-Docker daemon.

**Trust boundary.** GitLab passes job variables, including `CI_JOB_TOKEN`, to service containers, so the job's own DinD service holds the token in its environment; the context reaches that daemon by design. Repository code is therefore trusted with the job token, as it is in any job script; the adapter's guarantee is that the token is not stored in the mounted build directory or passed into the context's environment, and evidence never captures daemon or process state.

## Agent skills and development tooling

No agent skill, AI account or agent session is required to configure, build, test, migrate or run this template.

- **Restruct SDLC (default development workflow).** The GitHub submodule `.agents/skills/restruct-sdlc/` is discovered by Claude through `.claude/skills/restruct-sdlc` → `../../.agents/skills/restruct-sdlc`. It keeps its MIT license. Initialize it for agent work with `git submodule update --init -- .agents/skills/restruct-sdlc`; application testing and CI do not initialize or validate it. See [DEVELOPMENT-WORKFLOW.md](../DEVELOPMENT-WORKFLOW.md) for updates and removal.
- **Retired generator.** The foundation generator Skill this repository was originally rendered from is retired and is not part of this distribution; the committed repository is the maintained template. See [BOOTSTRAP.md](BOOTSTRAP.md).
- **Component and theming guidance.** The UI kit source, tokens, themes and browser checks are described in [WEB-SHELL.md](WEB-SHELL.md#component-and-theming-guidance).
- **Fresh clone.** Application commands work without Restruct initialization. For agent work, initialize the pinned submodule explicitly (or clone with `--recurse-submodules`). Linux/WSL support the Claude symlink; native Windows requires symlink support.
- **Windows.** Skill aliases are Git symlinks. A native Windows checkout without symlink support (`core.symlinks`, Developer Mode) gets a small text file instead; Linux and WSL are the supported runtime.
- **Compatibility identifiers.** Some identifiers keep the retired generator's former name on purpose: the `archetype:start`, `archetype:check` and `archetype:stop` documentation-block markers that the harness executes, and the validation volume names.
- **CI.** The submodule and its Claude alias are outside application validation and provenance checks; CI disables automatic submodule initialization. Parent metadata changes retain their ordinary path classification.

## Inherited, new and open

- **Inherited and preserved:** health and readiness semantics and deadlines, redacting core logger, dependency outage behaviour, the containerized checks and the browser suites.
- **Operational foundation:** diagnostic context and per-hop spans, one completion record per health request in all three services, bounded error and framework output, Python validator conformance corrections, the conformance corpus and Go validator, and the CI entry point with its controls, validation profiles and path selection, and the hosted adapter with its credential proof and disk guard.
- **Open:** metrics and exporters, trace export and collectors, dashboards and alerting, the permanent hosted capacity floor (provisional 35 GiB until measured), and the GitLab project settings above.

## Git ignore metadata validation

Root `.gitignore` changes are classified as fast metadata, rather than
foundation/full. `.gitignore` alone still admits an automatic pipeline and runs
fast checks; it is not a workflow documentation bypass. The existing docs path
class supplies its fast profile. Mixed runtime/foundation changes independently
select integration/full; deleted/renamed files retain classification of all
affected paths, and unknown destinations remain fail-closed.

B-ENV-EXAMPLE in the lint suite checks effective Git exclusions for `.env` and
`.env.local`, including later negations, and rejects a missing/symlinked root
ignore file. Adding a machine-local log exclusion does not justify a full stack
run. Changes to the CI map/harness themselves remain foundation/full.

## Targeted CI meta-tests

The expensive `B-CI-CONTROLS` check belongs to the `ci-controls` suite. The smaller
`controls` suite retains dependency and approved-wrapper controls. Full application
coverage remains; selector/completeness logic includes `ci-controls` when changed
CI/harness/control dependencies or unknown paths require it. Shared policy is in
`ci_controls_required` in `infra/ci/lib.sh`; force a diagnostic run with
`CI_CONTROLS_FORCE=1`. Scheduled diagnostics retain CI controls. Missing comparison
information fails toward running them. `run.sh all` uses its base and the same rule;
explicit `ci-controls` remains available. Omitted controls have no fabricated PASS.

## Browser report ownership

Local browser image `pwuser` need not share the operator's UID. The harness's root,
intelligence-off/shared, rename and identity browser runs write reports to
`/tmp/report` inside a uniquely named, run-labelled one-off. After execution, the
harness captures ownership, copies reports out by immutable container ID (without
`docker cp -a`), then removes the owned container and its anonymous volumes.
Reports land with the invoking operator's ownership. A copy/removal failure fails
the check; a missing/foreign/start-inventory container is not copied or deleted.
Existing project cleanup covers interrupted runs. Feature browser runs use
copy-out or stdout archives.

The default image user, browser cases, worker counts, identity bearer scrubbing,
private handoff handling and documented test commands remain. Only harness report
collection differs: no report bind mount, world-writable report directory or
privileged post-run chown is needed. Do not blindly override the browser UID: its
development server also needs image-owned `/app` write access.

## Compromised MCP client containment

Normal MCP logout is local to one connection. Containing a leaked or compromised client credential of a user U is a
separate, explicitly authorized operator procedure. It is a runbook, not a route, tool or registry, and it reuses the
trusted functions `identity.Service.DisableUser` and `EnableUser`, which still need the production operator entry point
([SECURITY.md](../SECURITY.md) F-6).

0. **Record prior state.** Read U's account state (`app.users.status`, `disabled_at`) and check, by operator
   coordination, whether another disable or hold for U is unresolved. Record both in the operator record.
   - U active: step 1 always establishes this procedure's own hold, even when another hold is open; that other hold
     then forbids release in step 4.
   - U already disabled: skip step 1. The existing disable already revokes sessions and blocks sign-in, setup and
     activation. This procedure holds nothing of its own and never releases that state.
1. **Hold and local invalidation** (U active): `DisableUser(U)`. Every browser session and every pending and active MCP
   connection of U becomes terminal (issued access tokens then fail locally); sign-in, setup and activation are blocked.
2. **Provider revocation:** delete U's authorized Connect application for the client, then read back U's authorized
   applications (WorkOS records no event, so the operator record carries the operation, scope, time and outcome, never
   identifier values).
3. **Uncertainty:** a failed, timed-out or unconfirmed provider step leaves containment unresolved. U stays disabled and
   the step is retried or escalated; U is never re-enabled on uncertainty.
4. **Release:** `EnableUser(U)` only when this procedure established the hold, provider revocation is confirmed, the
   operator explicitly cleared this procedure's hold in the operator record, and no other disable or hold for U is
   unresolved. Otherwise U stays as found. Old connections stay terminal; U creates new ones (new URLs) through the
   setup page.

**After a restore.** The restore invalidation already ends every MCP connection (each owns a session row), and a
database restore never undoes a provider revocation. "The operator re-enables after restore" is not unconditional: the
prior-state, explicit-clearance and other-hold checks above still govern, and an already-disabled or otherwise held user
stays disabled ([RECOVERY.md](RECOVERY.md)).

Side effects and limits: every session and connection of U ends across clients and Organizations; the provider step
ends every token family of U with that client. Access tokens already issued remain cryptographically valid until their
expiry but are refused locally because their connections are terminal. This is evidence for the tested client, user and
configuration only; there is no claim about arbitrary clients, no revocation SLA and no immediate JWT invalidation.
Live provider revocation has not been exercised for this export.
