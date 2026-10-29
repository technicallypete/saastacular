> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G012 — Secrets and secure engineering controls: Implementation Plan

Revision: r1, 2026-10-15 UTC. Status: **provisional, review-ready for Codex Plan re-review**. This is not approval,
not an execution binding and not implementation. The G011 predecessor gate is satisfied (final ACCEPT); execution
waits for explicit Codex Plan ACCEPT and the Git-first execution binding with input and resource release (§0). Tier:
architecture/security (Goal Execution Policy).

r1 is the bounded revision for the r0 review (not included) (CORRECTION LOOP, SHA-256
`bf2fb1f4a10ced0ebcc2713ac621d5842eedff6d6bd81cafd07d83fa95ec7620`). r0 is archived byte-identically as
G012-secrets-secure-engineering-plan-r0.md (not included) (SHA-256
`21d73dff2e62a1e5315d22f94a09670eb0342743bb1205e962e01a2a3808c82f`). Only the affected sections change; the ten ACs,
B5 with a checkpoint after attempt 3, the scope and every other reviewed disposition are unchanged. No attempt is spent.

**r1 change map:**

| r0 review item | Change | Where |
|---|---|---|
| 1 Pin the tool contract | Exact releases and digests from official metadata (sources and access times recorded): osv-scanner v2.6.0 image index digest, govulncheck `golang.org/x/vuln` v1.8.0, `golang:1.27.1-alpine` index digest; CLI flags, JSON schema and exit codes of those versions. D0 now only collects the findings baseline and confirms behavior | §5.1, §5.3, §5.5 |
| 2 Link accounting | The stream carries regular files and links separately counted; links are stored, never dereferenced; their target text is scanned. Clean fixture includes a Vestive-style directory link and an outside-target link; canary, exemption and incomplete-stream cases still fail | §4.2, §4.3 |
| 3 Identity coverage | Coverage is reconciled on normalized (ecosystem, name, resolved version); project roots (the virtual uv root `intelligence`, the npm root entry, the Go main module) are excluded; repeated locations collapse; unsupported records are listed. Negative: omission plus substitution or duplication with an unchanged count fails | §5.3, §5.6 |
| 4 Sensitive-state inventory | Recoverable values (`pkce_verifier`, `csrf_token`) distinguished from hash-only values; G013 private-backup handoff made exact | §2.1, §16 |
| E1 (not approved online) | Switched to download-only public advisory databases with no dependency data sent: four HTTPS objects into an owned volume, scanners run with network none; size, time, integrity and freshness bounds; owned cleanup. No owner decision request is needed (§5.2) | §1, §5.2, §14, §15 |
| E2 | Recorded: containerized test tooling within the Spec; no runtime ADR | §14 |
| Reviewer option | Fixed offline synthetic fixtures for both scanners replace the online known-input fixtures | §5.6 |
| G011 final ACCEPT (coordinator relay) | Rebased onto `bf462f6`; A-G011-1..4 bound to the accepted final G011 (final review `481afdb9…`, Result `e62069da…`) | header, §0 |

- **Goal / Spec / definition review (bound by SHA-256; unchanged from `795b95a` to `ee3fcac`):**
  - [Goal](../goals/G012-secrets-secure-engineering.md)
    `fb6becf9ebfaeeffb3d2b5833f0894a8dbb9c649d5be103d924710be45ce4174`;
  - [Spec](../specs/G012-secrets-secure-engineering-spec.md) (r0 plus the privacy inheritance amendment)
    `a48d865b3bc462b86785406c6991f4e35b323a73d21af3d8079cf49ec6e9c493`;
  - definition review (not included)
    `e7a13253ea081e23b32c4670f59e9e0bcc27f5412e9a46c1cf221abb5c7e41d1`.
- **Accepted final G011 bound (§0; r1 update after G011 FINAL ACCEPT):** merge
  `1af469634225e6e6527aca485b40aa55953227a8`, merged tree `c312478966715a729d1e5431dd567f8c275b3cb6`;
  [implementation review with final acceptance](../results/G011-implementation-codex-review.md)
  `481afdb96a5ff91152f238233967b9026650d04311667aa2996cfb6b0bd73219`;
  [final Result](../results/G011-durable-audit-events-result.md)
  `e62069daa0107d82c08bf0f15f63bba47a5e0d8a39c7f3382689e1e0e3b36572`;
  closeout receipt (not included)
  `95e0cd46ef819e2a45c10435f376e837d052e2d828ee2578cc0b40f41940ce02`; accepted Plan r2 `72042f5c…`;
  [AUDIT.md](../../../foundation/AUDIT.md)
  `33a9bac93b867342e81a949681e742b90406d06cdf5471f9e495e4c889c44760`; `core/migrations/00008_audit_events.sql`
  `b1dedf1f…`; ADR-006 `dc7eda77…`; ADR-003 `537522dc…`. r0 bound the pre-acceptance review `3cc2bf84…` and Result
  `506e8217…`.
- **Planner:** task G012, role plan; native isolated worktree `.claude/worktrees/agent-<id>` (standing
  owner fallback for native names, AGENTS.md "Task-prefixed Claude worktrees"); branch `g012-plan`. Effort xhigh as
  instructed; the runtime setting cannot be verified from inside the worker (unknown). Implementation default: high.
- **Baseline (Git-first):** r1 is rebased onto `origin/staging` `bf462f6fd331bfc4215fe6820c481bf1b1abbf62` (G011
  final acceptance). Its delta from r0's base `795b95a` is management only: `ee3fcac` (the r0 review,
  SUPERVISOR-HANDOFF, PROGRESS) and `bf462f6` (the G011 final review, final Result and closeout receipt,
  SUPERVISOR-HANDOFF, PROGRESS). r0's base delta from the G011 merge `1af4696` was management only too. No probe,
  install, build, scan, test, Docker run or harness command was
  run. r0 used two read-only exploration delegates (harness/CI wiring; secret/transport/dependency inventory). For r1,
  item 1 was answered only from read-only official metadata as the review permits (§5.1): release APIs and docs, the Go
  module proxy and checksum database, registry manifests without pulls, and object metadata (HEAD) of the advisory
  downloads. Nothing was pulled, installed or executed.
- **Read:** AGENTS.md, CLAUDE.md, ARCHITECTURE §3, §21/21.1, §27, §40, §47, §56/56.1, DECISIONS, ADR-002/003/004/006,
  ROADMAP, PROGRESS, SUPERVISOR-HANDOFF, BUDGET_POLICY, PERFORMANCE (PG-001..004), the G011 Plan r2, its three Plan
  reviews, binding confirmation, Result and implementation review, foundation AUDIT/IDENTITY/DATABASE-ISOLATION/
  OBJECT-STORAGE/OPERATIONS/VENDORS/FOLLOW-UP, and the code/config named below. `docs/SECURITY.md` is empty; no
  `docs/RISKS.md` exists. Core has no service-specific CLAUDE.md or skill; no UI work, so no Vestive skill applies.

## 0. Gates and G011 reconciliation (bound to the accepted final G011 in r1)

| Gate | Status at r1 |
|---|---|
| G011 final ACCEPT (staging pipeline `2948834614`, cleanup, final Result) | **satisfied**: Codex final ACCEPT 2026-10-15 (review `481afdb9…`); staging fast and gate SUCCESS on merge `1af4696`; final Result `e62069da…` and closeout receipt `95e0cd46…` imported |
| Codex Plan ACCEPT | pending (this r1; r0 received CORRECTION LOOP) |
| Git-first execution binding: allow-list (§11.7) and initial Result imported by Codex | pending |
| Owner decisions | none required by this Plan; the Codex review items in §14/§15 may escalate one |

**G011 assumptions (A-G011-n), now bound to the accepted final G011** (r0 labelled them provisional). The final
acceptance changed no application, migration or validation file after the merge: the merged tree equals the reviewed
source, and every later commit is management only. Each assumption therefore holds as written; binding re-checks only
the then-current target.

- A-G011-1: G012 adds **no migration, audit event, role, grant or schema**, and edits no G011 file except shared
  harness/count pins and OPERATIONS/FOLLOW-UP prose, which G011 also edited. The accepted totals are those §11.3
  starts from: 107 required rows; full pipeline fast 33 plus gate 72, 101 distinct IDs (final review). A later change
  to these files is reconciled by rebase, and §11.3 is recomputed from the then-current totals.
- A-G011-2: the deprovisioning runbook (§3.4) uses the delivered catalog exactly: T8 `membership.role_changed`, T9
  `membership.removed`, T15 `user.disabled`, T16 `user.enabled`, T18 `session.revoked`
  (`logout`, `logout_all`, `provider_logout`, `disabled`, `superseded`), T19, G1 `auth.sign_in` denied `user_disabled`,
  G6 `revocation.unaudited` and the `audit.gap` diagnostic; global rows are privileged-inspection only; Organization
  rows are owner-readable through `ReadAuditEvents`. The final review confirms the 29/29 action inventory and
  B-AUDIT-CATALOG/MIGRATION/IT on staging. Any later catalog change escalates rather than being absorbed.
- A-G011-3: the M5 revocation-priority exception (ADR-006 owner clarification) is unchanged; G012 adds no audit bypass
  and no new event. Service-credential rotation is not a catalog event (no producer exists); its evidence is the Result,
  CI artifacts and the operator record, and G013 receives that boundary.
- A-G011-4: the G011 security-review residuals (F1 fixed, F2 accepted deviation) do not touch configuration, secrets,
  scanners or transports. The final review hands G013 audit immutability and restricted maintenance, unknown-outcome
  and sensitive-backup responsibilities; G012's §2.1 classification and §16 handoff feed the last of these.

## 1. Implementation approach

One finite inventory, the existing resolver boundary hardened in place, two bounded scanner integrations at existing
entry points, one disposable credential rehearsal and one configuration-readiness evaluator. No new runtime service,
endpoint, schema, role or framework.

| Req | Approach (detail in §2–§7) |
|---|---|
| R1 | Inventory every credential by name in `docs/SECURITY.md`. Harden `config.Load`'s secret acquisition: closed reason codes, provider text discarded, empty/whitespace/control-character values refused, no partial result. A test-only replacement provider proves substitution through the same loader. A boundary test pins acquisition to `config.Load` at startup |
| R2 | Privilege matrix and operator procedures in `docs/SECURITY.md`. One exercised rotation and withdrawal of the core runtime PostgreSQL credential on the disposable `ops` stack (new check `B-SECRET-ROTATION`), observing pre-existing connections, their termination and core's restart recovery. Human deprovisioning runbook mapped to the delivered G011 events |
| R3 | Pin the README source scan to the submitted set (`git ls-files --cached --others --exclude-standard`, streamed, no bind mount) and fail closed when a submitted path matches a local-only scanner exemption. New fixture check `B-SECRET-SOURCE`. Static-rule mapping; no new analyzer |
| R4 | `osv-scanner` v2.6.0 over the three lock inputs and `govulncheck` v1.8.0 over the core module (reachability and standard library), containerized, credential-free, streamed inputs, bounded time, **offline against downloaded public advisory databases** (no dependency data leaves the machine). A small policy evaluator applies the Spec's blocking/exception rules and reconciles coverage by package identity. New check `B-DEPSCAN`. The current findings baseline is collected in the first authorized batch (D0) |
| R5 | Transport and at-rest matrix in `docs/SECURITY.md`. A pure evaluator in `core/internal/config` plus a stdin-driven command `core/cmd/readiness` refuse a production-ready result for plaintext, verification bypass, unsupported transports, local exceptions and missing at-rest evidence; it only ever reports "configuration-ready (configuration-only)" |
| R6 | Extend `docs/foundation/VENDORS.md` inside its reviewed customization run with the missing integrations and a responsibilities/privacy section (§7). Populate `docs/SECURITY.md` |
| R7 | Three new required IDs, minimal count/path pins, existing suites and partitions; one broad final local run; staging full gate |

**Rejected alternatives (recorded):**
- *Ecosystem-native scanners* (`npm audit`, `pip-audit`): `pip-audit` needs a Python install of its own unpinned
  transitive dependencies, and it reports no severity; `npm audit` adds a third result format. One pinned static
  scanner (`osv-scanner`) covers all three lock formats with severity, and `govulncheck` adds Go reachability.
- *Online OSV queries* (r0 default): they send the dependency graph to a newly selected service; not approved (E1).
  The download-only databases give the same advisory data for the three ecosystems, bounded by export freshness.
- *Image/OS scanners* (Trivy, Grype): container scanning is follow-up scope by the Spec.
- *A production file-backed `SecretProvider`*: no deployment target exists; a test-only provider demonstrates the
  boundary. The WorkOS `_FILE` path remains the documented mounted-secret route (Q1).
- *Changing core's PostgreSQL transport* (pin plaintext like migrate, or add TLS settings): new transport support needs
  a scope decision; recorded as finding F-1 and a readiness block.
- *Customizing `infra/compose/local.yml` or `docs/foundation/GETTING-STARTED.md`*: both are toggle-dependent renders;
  the harness's `TOGGLE_RENDERED` mechanism supports only `.env.example` and VENDORS.md. A new overlay file and the
  non-rendered README carry the change instead.
- *A new harness suite*: it would touch `SUITES_KNOWN`, `CI_SUITES_*`, validate.sh, pipeline-union.sh and many control
  pins. Existing `lint`, `lint-base` and `ops` suites hold the new checks.
- *Appending a G012 section at the end of VENDORS.md*: it would need new `variants_off` composition logic; extending
  the existing contiguous reviewed run changes two constants only (§7.1).

**Spec "Before Plan approval" items:**

| # | Spec item | Resolved in |
|---|---|---|
| 1 | Actual accepted G011 interfaces, residuals and audit coupling | §0 (A-G011-1..4, bound to the accepted final G011 in r1); §3.4; no migration or event expansion |
| 2 | Credential source/reload semantics; the single rotation fixture incl. authenticated connections | §2.1, §3.2, §3.3 |
| 3 | Scanner choices, lock/toolchain versions, sources/network, severity/reachability; baseline unknown until D0 | §5.1–§5.5 (exact pins in §5.1; D0 collects findings only) |
| 4 | Checker inputs, measured versus declared evidence; unsupported encryption stays blocked | §6 |
| 5 | Changed files, AC mapping, actual CI selection, bounded timing | §8, §10, §11 |

## 2. R1 — Current-secret inventory and safe resolution

### 2.1 Inventory (names only; recorded in `docs/SECURITY.md`)

All configuration is read once at process start; no service reloads it. "Restart" below means a new process.

| Name | Consumer and use | Source | Required when | Missing / empty behavior | Rotation owner | Exposure restriction (existing proof) |
|---|---|---|---|---|---|---|
| `DB_PASSWORD` (Compose maps `DB_RUNTIME_PASSWORD`) | core pool as `<db>_runtime` | `SecretProvider` (`EnvSecrets`) | always | startup fails, names only | platform operator | core and postgres only (B-COMPOSE); redacted (`Config.SecretValues`) |
| `DB_MIGRATOR_PASSWORD` | one-shot migrate as `<db>_migrator` | migrate `LoadConfig` getenv | every migrate run | refused before connecting | platform operator | migrate env is exactly five names (B-COMPOSE, B-MIGRATE-ENV); redacted by migrate's logger |
| `DB_RUNTIME_PASSWORD`, `DB_RETENTION_PASSWORD`, `POSTGRES_SUPERUSER_PASSWORD` | role creation (`10-roles.sh`, `provision-roles.sql` via psql `\getenv`) | environment | first volume init or provisioning | init refuses; provisioning never changes an existing role's password | platform operator | postgres service only |
| `REDIS_PASSWORD` | core readiness AUTH+PING only; redis server | `SecretProvider` | always | startup fails | platform operator | redis, core |
| `OBJECT_STORE_SECRET_KEY` | core S3 client (readiness HeadBucket today; the G009 adapter is not wired in the server); minio-init creates the scoped user | `SecretProvider` (core) | always | startup fails | platform operator | core, minio-init; argv exposure inside minio-init (F-4) |
| `MINIO_ROOT_PASSWORD` | minio, minio-init | environment | local fixture | Compose refuses | platform operator | local fixture only |
| `INTELLIGENCE_SERVICE_TOKEN` | intelligence verifies `/internal/*` (constant-time); core loads it but sends it nowhere today (F-3) | core `SecretProvider` when `INTELLIGENCE_URL` is set; intelligence env | core: with `INTELLIGENCE_URL`; intelligence: optional | core startup fails; intelligence warns and answers 401 | platform operator | core, intelligence |
| `OPENROUTER_API_KEY` | intelligence gateway adapter | environment | optional; unset is the documented disabled state | `gateway_not_configured` per request | platform operator | intelligence only; never generated by `configure-env.sh` |
| `IDENTITY_WORKOS_API_KEY` or `IDENTITY_WORKOS_API_KEY_FILE` | core WorkOS adapter | `LoadIdentity` (value or file, mutually exclusive; G004 contract, outside `SecretProvider`) | all three identity values, or none; production requires them | startup fails naming variables; file failures sanitized | platform operator (WorkOS dashboard) | core only; web/intelligence never (B-COMPOSE, B-IDENTITY-GUARDS) |
| `IDENTITY_TEST_API_KEY` and per-run emulator material (`identity-run` volume) | identity test overlay | generated per run | tests only | — | harness | test overlay only |
| `IDENTITY_STAGING_WORKOS_API_KEY` | `TestStagingContract` | protected, environment-scoped CI variable or an operator 0600 file | staging qualification only | refused in any other profile | project maintainers | `identity-staging` job only |
| CI material: job-private DinD client certificates (`DOCKER_CERT_PATH`), GitLab-issued job token | hosted jobs | GitLab runner | per job | — | GitLab | job-private; remote URLs stripped of embedded authentication; no registry login, deploy or CI/CD secret in `fast`/`gate` |

**Runtime authentication material** (generated per use with 256 bits from `crypto/rand` in
`core/internal/identity/credential.go`; not configuration). Checked against `core/migrations/00002_identity.sql` and
`core/internal/identity/store.go` (r1 correction):

| Stored field | What it is | Stored form | Lifetime |
|---|---|---|---|
| `app.sessions.credential_hash` | session bearer (cookie value) | SHA-256 only | session (idle 30 min, absolute 8 h defaults) |
| `app.sessions.csrf_token` | CSRF value returned to the browser | **recoverable** (32 raw bytes) | the session |
| `app.sessions.provider_logout_ref_hash` | provider-logout return reference | SHA-256 only | 10 min |
| `app.login_transactions.state_hash`, `binding_hash` | OAuth state; browser login binding | SHA-256 only | at most 10 min |
| `app.login_transactions.pkce_verifier` | PKCE verifier used at code exchange | **recoverable** (text) | at most 10 min; expired rows removed at a later sign-in |
| `app.sessions.provider_session_id`, `app.identity_principals.subject` | WorkOS session and user identifiers | plain (identifiers, not proofs) | session / account |

Provider access tokens stay in memory only. A database dump or backup therefore contains recoverable PKCE verifiers
and CSRF values for unexpired rows, plus hashes, provider identifiers, names and audit rows: it is private operational
material (G013 handoff, §16).

**Public identifiers (not secret):** `IDENTITY_WORKOS_CLIENT_ID`, `IDENTITY_WORKOS_ISSUER`, `OBJECT_STORE_ACCESS_KEY`,
`MINIO_ROOT_USER`, `DB_USER` and role names, the staging client ID and issuer, the emulator defaults. **Disabled
versus missing:** intelligence absent (`INTELLIGENCE_URL` unset) and gateway absent (`OPENROUTER_API_KEY` unset) are
documented disabled states; every other row is required and fails closed.

### 2.2 Resolver hardening (`core/internal/config`)

Current gap (Spec R1): `Load`'s `secret` closure appends `err.Error()` verbatim, so provider text enters the returned
error; `EnvSecrets` accepts whitespace-only values; a provider returning `("", nil)` is accepted as a credential.
`observability.LogExit` already prints only allow-listed names, so no current log leaks provider text; the fix closes
the boundary itself.

- **New `secrets.go`:** exported sentinels for provider implementations to wrap: `ErrSecretNotSet`,
  `ErrSecretDenied`, `ErrSecretUnavailable`, `ErrSecretAmbiguous`. One unexported `resolveSecret(p SecretProvider,
  name string) (string, string)` asks the provider exactly once and returns either the exact value or one closed
  problem, never both:

  | Provider outcome | Problem text (name plus fixed reason) |
  |---|---|
  | error wrapping `ErrSecretNotSet`, or `("", nil)` | `<NAME> is not set` (keeps the rendered test's wording) |
  | value empty after trimming whitespace | `<NAME> is empty` |
  | value containing NUL, CR or LF | `<NAME> is invalid` |
  | error wrapping `ErrSecretDenied` | `<NAME> is denied by the secret provider` |
  | error wrapping `ErrSecretAmbiguous` | `<NAME> is ambiguous` |
  | any other error or type, or a nil provider | `<NAME> is unavailable from the secret provider` |

  No fallback to the environment, a default or a stale value; the provider's error text and type are discarded.
  A valid value is returned byte-exact (no trimming).
- **`config.go` (customized render, one hunk):** `EnvSecrets` wraps `ErrSecretNotSet`; the `secret` closure calls
  `resolveSecret`. `Load` still returns `Config{}` with any problem, so a partial credential set cannot be returned, and
  `main.go`'s reviewed order (both loaders before any pool, S3 client or health check; digest-pinned `MAIN_GO_G004_SHA`)
  means invalid configuration constructs no client. `main.go`, `LoadIdentity`, migrate and intelligence/web code are
  unchanged.
- **Not added:** file-source support for the four `Load` secrets (no deployment needs it; Q1), any secret RPC, cloud
  adapter or request-time lookup.
- **Compatibility:** no configuration name is added, so `observability.configKeys` and its parity test
  (`TestConfigKeysAllowListMatchesConfigLoad`, which records the names `Load` reads and requires every name in `Load`'s
  messages to be allow-listed) stay unchanged; the new texts carry only the name. The rendered `config_test.go` is not
  edited (its "is not set" expectation holds); new tests go in a new file.

## 3. R2 — Least privilege, rotation and deprovisioning

### 3.1 Privilege matrix (recorded in `docs/SECURITY.md`; reuses accepted evidence)

| Credential | Minimum use | Accepted grants / proof | Must never hold |
|---|---|---|---|
| `<db>_runtime` | app DML under FORCE RLS; audit USAGE, SELECT (filtered), INSERT of 12 columns | DATABASE-ISOLATION, AUDIT; B-RLS-*, B-AUDIT-CATALOG CT-AUD-3/5 | owner, migrator, SUPERUSER, BYPASSRLS, TEMP/CREATE/TRIGGER, role membership |
| `<db>_migrator` | SET ROLE `<db>_owner` to migrate | B-MIGRATE-IT, B-MIGRATE-ENV | runtime use; presence in core's env |
| `<db>_retention` | LOGIN, no current consumer, no audit privilege until G013 | AUDIT G013 contract | anything before G013's reviewed migration (F-7) |
| `<db>_owner`, helpers | NOLOGIN | B-RLS-CATALOG, PROV checks | a password |
| MinIO scoped user | bucket-only list/get/put/delete | OBJECT-STORAGE; B-PERSIST (scoped write) | root credentials in core |
| WorkOS API key | environment-wide WorkOS API (code exchange, session revocation) | IDENTITY | presence in web/intelligence |
| Intelligence token | shared bearer for `/internal/*` | intelligence `test_auth.py` | browser/web exposure |
| Gateway key | OpenRouter calls when enabled | none (disabled) | anything but intelligence |

Browsers and web bundles never acquire datastore, WorkOS or gateway values: web's container environment holds none
(B-COMPOSE), no `NEXT_PUBLIC_*` variable exists, credentials are never stored in browser storage (B-IDENTITY-GUARDS), and
the web build runs without `.env`. No grant changes are planned; a grant defect found in review would be a separate
reviewed change.

### 3.2 Rotation and withdrawal procedures

Documented for every credential class in `docs/SECURITY.md`; **exercised for one**, the core runtime PostgreSQL
credential, because it has the open-connection semantics the Spec names.

1. Generate the replacement in the deployment's secret store; never on a command line.
2. The provisioning administrator runs `ALTER ROLE <db>_runtime PASSWORD …` through psql with the value read by
   `\getenv` (or psql's interactive `\password`, which sends a hashed verifier). Caveat recorded: server statement
   logging can capture a plaintext `ALTER ROLE`. Provisioning itself never changes an existing password (G008 design).
3. New logins with the old value fail at once. **Already-authenticated sessions keep working**: PostgreSQL does not
   re-authenticate open connections.
4. Update core's source (`DB_PASSWORD`) and restart core: values are read only at startup.
5. Terminate remaining pre-rotation `<db>_runtime` backends (`pg_terminate_backend`, filtered by role and
   `backend_start` before the rotation) and verify none remain.
6. Withdrawal without replacement: `ALTER ROLE … PASSWORD NULL` plus step 5; the consumer then fails closed
   (readiness 503). Removing a source value alone revokes neither loaded values nor open connections.

### 3.3 `B-SECRET-ROTATION` (new required ID, `ops` suite, disposable copy)

Runs last in `variants_ops` (after `persist` and the identity probes, before `save_logs` and `down -v`) on the existing
disposable `g001-ops-<run>` stack. Synthetic credentials only: the new value is generated with `random_hex` and appended
to `SECRETS_FILE`, so B-SECRET-CANARY proves it never reaches logs or evidence. Administrative SQL goes to the postgres
container's local socket as the disposable superuser on stdin; password logins use TCP (`-h 127.0.0.1`, so the image's
`scram-sha-256` rule applies) with `PGPASSWORD` passed by name through `compose exec -e`, never in argv. No bind mount.

| Step | Assertion |
|---|---|
| R0 | core `/readyz` 200; at least one `<db>_runtime` backend exists; record their PIDs and `backend_start` |
| R1 rotate | `ALTER ROLE` to the new value; time T recorded |
| R2 old connection | `/readyz` still 200 and pre-T backends still present. A new connection could only use the old value and would fail, so a 200 here was served by a pre-rotation connection |
| R3 new logins | old value: `password authentication failed`; new value: `SELECT 1` succeeds |
| R4 terminate | terminate pre-T runtime backends (count ≥ 1); none remain; core, still holding the old value, reaches `/readyz` 503 within 30 s |
| R5 restart | `env_set` the copy's `DB_RUNTIME_PASSWORD`; `compose up -d --no-build --no-deps --force-recreate core`; `/readyz` 200 within 90 s; its backends started after T |
| R6 withdraw | `ALTER ROLE … PASSWORD NULL`; new-value login fails; terminate; `/readyz` 503 within 30 s |
| Record | durations of R4/R5/R6 and backend counts; any failure → FAIL (stack torn down by the existing `down -v`) |

### 3.4 Human/tenant deprovisioning runbook (existing paths only)

| Step | Existing path | Durable G011 record | Notes |
|---|---|---|---|
| End a session / all sessions | user logout, logout-all (web → core `/internal/v1/auth/logout`) | `session.revoked` `logout` / `logout_all`, one per unrevoked session (expired-but-unrevoked included) | global rows: privileged inspection only |
| Provider sign-out | provider logout | `session.revoked` `provider_logout`; `session.provider_revocation_recorded` | revokes that WorkOS session only; remote deprovisioning not claimed |
| Disable a user | `identity.Service.DisableUser` (trusted method; **no route or CLI in the server today**, F-6) | `user.disabled` (system) and `session.revoked` `disabled`; later sign-in: `auth.sign_in` denied `user_disabled` | `EnableUser` → `user.enabled` (strict); old sessions never revive |
| Remove or downgrade tenant access | tenancy trusted primitives (privileged pools only since version 6; no product surface until G018/G019) | `membership.removed` / `membership.role_changed` (Organization rows, owner-readable) | current-authority checks (G007 liveness, G008 RLS) deny the next guarded operation |
| Provider-side account removal | WorkOS dashboard (operator) | none in `audit.events` | blocks new authentication only; still run local revocation |
| Audit-only fault during revocation | M5 (owner decision, ADR-006) | `revocation.unaudited` plus one `audit.gap` diagnostic | revocation commits; check `audit.gap` `independent: failed`; never a reason to undo logout/disable |

Privileged or bootstrap SQL is recorded with actor `unknown` and is not a substitute procedure. An authorized operator
entry point for `DisableUser` and membership removal is a named **production prerequisite** (F-6), not added here.

## 4. R3 — Submitted-source secret scan and static checks

### 4.1 Demonstrated gaps (read-only evidence)

- G-1: `infra/security/gitleaks.toml:8-16` exempts every path matching `(^|/)\.env$`, `(^|/)\.env\.local$`,
  `node_modules/`, `.next/` and `.archetype/outputs/*/proposed/`, tracked or not. A force-added `.env` would be unscanned.
- G-2: the README `secret-scan` line runs the Compose service `infra/compose/local.yml:213-219`, which bind-mounts the
  whole working directory (`../..:/repo:ro`). The scanned set is the ambient tree, not a pinned submitted set; from the
  main checkout it also contains nested worktrees (their exclusion is only a machine-local `info/exclude` entry).
- No tracked file currently matches an exemption (checked with `git ls-files`); the one tracked symlink is
  `.agents/skills/vestive-design`.

### 4.2 Design

- **`infra/security/secret-scan.sh`** (new, host bash/git/GNU tar only):
  `bash infra/security/secret-scan.sh --env-file FILE [--root DIR] secret-scan`.
  1. Lists the submitted set: `git -C ROOT ls-files -z --cached --others --exclude-standard`, de-duplicated, and
     classifies each entry with `test -L` **before** any other test (so a link to a directory is a link, not a
     directory): **link** (kept as a link; its target is never read or followed); **regular file**; **directory**
     (an untracked nested repository, which Git lists as one entry: skipped and counted, never descended); anything
     else (gitlink, device, missing path) → exit 2. A non-Git root → exit 2. The current tree has one tracked link,
     `.agents/skills/vestive-design` → `../../web/.claude/skills/vestive-design` (mode 120000), whose target files are
     themselves tracked and therefore scanned as regular files.
  2. **Exemption guard:** reads the path patterns of the `[[allowlists]]` block described as "local-only files that are
     never committed" from the candidate's `gitleaks.toml` (fails if the block or its patterns are missing) and exits 3
     if any submitted path (regular file or link) matches one. The exemption therefore stays effective only for
     untracked, ignored files and can never hide submitted content. `gitleaks.toml` is unchanged.
  3. Streams one tar on stdin (GNU tar `--null --no-recursion -T -`, which stores links as links and never
     dereferences): `scan/src/<regular files and links>`, `scan/link-targets.txt` (one tab-separated path and target
     line per link, so link target text is itself scanned), and `meta/count-regular`, `meta/count-links`, `meta/gitleaks.toml`. The
     consumer is `docker compose --env-file FILE -f infra/compose/local.yml -f infra/compose/secret-scan.yml --profile checks run --rm -T secret-scan`.
- **`infra/compose/secret-scan.yml`** (new overlay, not rendered; Compose ≥ 2.24 already required): for the existing
  `secret-scan` service (same pinned image) `volumes: !reset []` and an `sh -ec` entrypoint that unpacks into `/tmp/w`
  and checks the accounting: `find /tmp/w/scan/src -type f` equals `meta/count-regular` and is positive; `-type l`
  equals `meta/count-links`; no entry of another type. Links stay unresolved inside the container (a target outside the
  submitted set does not exist there, and gitleaks runs without `--follow-symlinks`). It then runs
  `gitleaks dir /tmp/w/scan --config /tmp/w/meta/gitleaks.toml --redact -v --no-banner --exit-code 23 --report-format json --report-path /tmp/r.json`.
  Exit mapping: 0 only with gitleaks 0 **and** a report equal to `[]`; 1 for gitleaks 23 with a parseable list; 2 for
  any unpack, accounting, report or scanner failure. `network_mode: none` is inherited.
- **README `platform:check`** line 4 becomes `bash infra/security/secret-scan.sh --env-file .env secret-scan`. It keeps
  exactly one `--env-file .env`, ends in ` secret-scan` (B-UNIT-*, B-CHECKS-ROOT parsing) and contains no "docker" text
  or shell metacharacter, so the build ledger ignores it (lib.sh `_lg_plan_line`). The README scan already runs in both
  partitions (`unit` in fast; `root` in gate); G012 adds no further execution. GETTING-STARTED's rendered
  `archetype:check` keeps the archetype's working-tree scan (unchanged render; its exemption stays needed there because
  the root suite writes a generated `.env` into the repository). A tracked exempt-path file still fails CI through the
  README scan in both partitions. The README's "all in containers" sentence gains that the file list comes from host Git.
- Exit codes: 0 clean, 1 finding, 2 scanner/stream/incomplete error, 3 submitted path matches an exemption, 64 usage.
  Any non-zero fails the README line. Scanner unavailable (image pull failure) is non-zero.

### 4.3 `B-SECRET-SOURCE` (new required ID, `lint` suite)

Stack-free, real pinned gitleaks image, network none for the scanner, disposable Git repositories under `$WORK` (host
git with explicit `-c user.*`). Canaries are generated credential-shaped values (a `generic-api-key` assignment with 32
random alphanumerics), appended to `SECRETS_FILE` and never printed (`--redact`).

| Case | Fixture | Expected |
|---|---|---|
| S0 | overlay config (`compose config --format json`, static env) | `secret-scan`: no volumes, network none, no ports, no environment, image equal to `local.yml`'s |
| S1 | clean repository-like tree containing a tracked relative directory link shaped like the Vestive one (`.agents/skills/x` → `../../web/.claude/skills/x`, target files tracked) and a tracked link whose target lies outside the repository (a sibling fixture directory holding a canary) | exit 0; counts report regular files and two links; the outside canary never appears (the link is not followed) |
| S2 | canary in a tracked file | exit 1, finding names that file |
| S3 | canary in an untracked, non-ignored file | exit 1 (submitted additions included) |
| S4 | canary in a force-added tracked `.env`, and in `sub/.env.local` | exit 3 for each, before scanning |
| S5 | canary in an ignored untracked `.env` with a clean tracked tree | exit 0, `.env` absent from the stream listing |
| S6 | untracked nested Git repository containing a canary | exit 0, reported as a skipped directory |
| S7 | incomplete stream: `meta/count-regular` altered; `meta/count-links` altered; an invalid config | exit 2 each, never 0 |

The real repository's README scan (unit and root suites) is the second clean case: it includes the Vestive link.

The publication-time evidence scanner (`infra/ci/validate.sh` `scan_set`) and its controls V5a–V6d are unchanged and
reused; no scanner-of-scanners framework is added.

### 4.4 Static checks mapping (recorded in `docs/SECURITY.md`; no new analyzer)

| Service | Retained checks | Security-relevant rules | Limits |
|---|---|---|---|
| core | gofmt, `go vet` (default analyzers), unit tests; B-MCP-SOURCE, B-IDENTITY-GUARDS, boundary unit tests | vet's correctness analyzers; project guards on SDK/token placement, bypass symbols, route/import sets | no security-specific analyzer (gosec class); no taint analysis |
| web | ESLint `next/core-web-vitals` + `next/typescript`, `tsc`; B-WEB-BOUNDARY, B-IDENTITY-GUARDS | React rules; server-only boundary; no browser credential storage | no security lint plugin; relies on React escaping |
| intelligence | ruff `E,F,W,I,B,UP,S,SIM,RUF` (tests ignore `S101,S105,S106,S603`), pytest | flake8-bandit (`S`) rules | rule-based only |

The S-review (§12) greps for named risky constructs (shell execution, string-built SQL, `dangerouslySetInnerHTML`,
`eval`, disabled certificate verification). Only a demonstrated current-code gap would justify an analyzer, and that
would be a scope item for Codex.

## 5. R4 — Dependency vulnerability policy and checks

### 5.1 Tools, sources and exact pins (r1)

| Item | Pin |
|---|---|
| Lock scanner | `osv-scanner` v2.6.0 (Google, Apache-2.0), released 2026-09-14T02:55:16Z |
| its image (index) | `ghcr.io/google/osv-scanner:v2.6.0@sha256:afd838850ac1a0fcc15ff4a041dc9ba11123c3f0d2666217a5f0fcf9222b55fa` |
| its linux/amd64 manifest | `sha256:b8cc6dc8805bb77fca4ada69c07cfef5a1e07784c22830e0bd49a440aa27d9fe`; built on Go 1.27.1 alpine, entrypoint `/osv-scanner`, has `/bin/sh` |
| Go reachability and standard library | `govulncheck` from module `golang.org/x/vuln` v1.8.0 (BSD-3-Clause), released 2026-09-08T20:52:42Z, origin commit `709015412431dd2b5b28a53c06c70bc02d49074c`, module `go 1.26.0` |
| its checksum-database lines | `golang.org/x/vuln v1.8.0 h1:clG4qBU6zH5VKjti8n5j8BBuYzoSha392xXMkXS351U=` and `golang.org/x/vuln v1.8.0/go.mod h1:Fzm4XK3Hbl1ZvZ7JpNTEWb7CJWOZ7m2LX0GLu4Fsrwo=` |
| Go toolchain image (index) | `golang:1.27.1-alpine@sha256:8a5910f31396cd4d89662f56c68b3ae31d374308270a1c3bd96672ee5ed43414` (the shipped core toolchain; `GOTOOLCHAIN=local`) |
| its linux/amd64 manifest | `sha256:cd9a32216aee5667f957a62d13a10032a63fd58e14b3f3d9cc8c2122f501e95e`, created 2026-09-17T21:42:23Z |
| Advisory data (download only) | `https://osv-vulnerabilities.storage.googleapis.com/{Go,npm,PyPI}/all.zip` (OSV's per-ecosystem export) and `https://vuln.go.dev/vulndb.zip` (Go vulnerability database bulk copy) |
| Policy evaluator | `vulnpolicy`, a new stdlib-only command in the validation-tools module (no new module requirement; B-DEPS keeps its one-requirement rule); compiled with `go run` in the pinned golang image, network none; unit tests in the existing check stage (B-CONTRACT-TOOLS) |

Newer releases existed at lookup (osv-scanner v2.7.0 on 2026-10-12, `x/vuln` v1.9.0 on 2026-10-11) and were not chosen
because they were under 14 days old. The pins live as constants in `infra/security/dependency-scan.sh` (single source;
the harness reports them). A version, digest, flag or schema change is a reviewed Plan amendment, never a D0 choice.
Advisory data changes independently of source, so results are not described as byte-reproducible.

**Metadata sources (read-only, 2026-10-15 UTC; no pull, install or execution):**

| Time | Source | Used for |
|---|---|---|
| 01:05:09Z | GitHub releases endpoint `https://api.github.com/repos/google/osv-scanner/releases` | osv-scanner release list and dates |
| 01:05:25Z, ~01:07Z | `https://raw.githubusercontent.com/google/osv-scanner/v2.6.0/docs/` files `offline-mode.md`, `usage.md`, `scan-source.md`, `output.md`, `supported_languages_and_lockfiles.md` | offline database layout and flags, `--all-packages`, `--no-call-analysis`, lockfile support (`go.mod`, `package-lock.json`, `uv.lock`), JSON layout |
| ~01:07Z | same tag: `cmd/osv-scanner/internal/cmd/run.go`, `pkg/models/results.go` | exit codes; JSON field names |
| 01:05:31Z | HEAD of the three OSV export objects; `ecosystems.txt` | sizes, Last-Modified, published MD5; ecosystem names `Go`, `npm`, `PyPI` |
| 01:06:04Z, 01:06:13Z | Go module proxy `https://proxy.golang.org/golang.org/x/vuln/@v/` (`list`, `.info`, `.mod`) and checksum database `https://sum.golang.org/lookup/golang.org/x/vuln@v1.8.0` | `x/vuln` versions, release time, origin commit, module Go version, checksum lines |
| 01:06:26Z | HEAD of `https://vuln.go.dev/vulndb.zip`; `https://vuln.go.dev/index/db.json`; `https://go.dev/doc/security/vuln/database` | size, Last-Modified, published MD5; database modified time; documented bulk download and `-db file://` use |
| 01:06:40Z, 01:06:59Z | registry manifests via `docker buildx imagetools inspect` (manifest and config metadata only) | image index/platform digests, base, entrypoint, environment |
| 01:08:02Z | `https://go.googlesource.com/vuln/+/refs/tags/v1.8.0/` files `cmd/govulncheck/doc.go`, `internal/govulncheck/govulncheck.go` | exit codes, `-db`, `-test`, JSON protocol and message schema |

**CLI and output contract (pinned versions):**

- osv-scanner, in its pinned image, `--network none`, with `OSV_SCANNER_LOCAL_DB_CACHE_DIRECTORY=/db` (documented
  layout `/db/osv-scanner/<ecosystem>/all.zip`):
  `osv-scanner scan source --offline --no-call-analysis=go --format json --all-packages -L /in/core/go.mod -L /in/web/package-lock.json -L /in/intelligence/uv.lock`.
  `--offline` documents that nothing is downloaded or sent; Go call analysis is disabled because govulncheck owns
  reachability. JSON (`pkg/models/results.go`): `results[].source{path,type}`, `results[].packages[].package{name,
  version, ecosystem}`, `dependency_groups`, `vulnerabilities[]` (full OSV records), `groups[]{ids, aliases,
  max_severity, experimental_analysis}`. Exit codes (`run.go`): 0 no vulnerabilities, 1 vulnerabilities found,
  127 general error, 128 no package sources, 129 API failure, 130 invalid configuration.
- govulncheck v1.8.0 in the pinned golang image: phase A (network) `go install golang.org/x/vuln/cmd/govulncheck@v1.8.0`
  and `go mod download` plus `go mod verify` for the streamed core module; phase B (`--network none`, `GOPROXY=off`,
  `GOFLAGS=-mod=readonly`, `GOTOOLCHAIN=local`) `govulncheck -db file:///db/govulndb -format json -test ./...`.
  Streaming JSON protocol `v1.0.0` (`internal/govulncheck`): one `config` message (`protocol_version`, `scanner_name`,
  `scanner_version`, `db`, `db_last_modified`, `go_version`, `scan_level`, `scan_mode`), `progress`, `SBOM` (`go_version`,
  `modules[]{path, version}`, `roots`), `osv` entries and `finding` messages (`osv`, `fixed_version`, `trace[]{module,
  version, package, function, receiver, position}`). A finding is **called** when its first trace frame names a
  function. Exit: 0 with `-format json` whatever the findings; non-zero only on error.

### 5.2 Invocation, isolation and the download-only advisory route (`infra/security/dependency-scan.sh`, new)

- **Inputs** are exactly the authoritative tracked files: `core/go.mod`, `core/go.sum` and the tracked `core/` sources
  (for govulncheck), `web/package-lock.json`, `intelligence/uv.lock` (plus `package.json`/`pyproject.toml` for the
  direct-dependency report), streamed into containers on stdin (both images have a shell and tar). No bind mount.
- **Advisory fetch (the only advisory network step):** a fetch container from the pinned golang image (busybox `wget`,
  CA certificates) downloads exactly the four public objects of §5.1 by plain HTTPS GET (no query string, cookie,
  authentication or project data) into a run-labelled named volume `g012-advisory-<run>` (no host path), recorded with
  `registry_row volume` at creation. Checks, any failure being a failed check: HTTP 200; size within the cap and equal
  to `Content-Length`; MD5 equal to the response's published `x-goog-hash` value when present (both hosts return it
  today; an absence is recorded, and the per-entry CRC checks still apply); `unzip -t` passes with at least one entry;
  `Last-Modified` no older than 72 h at scan time; `vulndb.zip` unpacks to
  `index/db.json` whose `modified` is no older than 30 days. Caps: Go 64 MiB, npm 512 MiB, PyPI 128 MiB, `vulndb.zip`
  32 MiB (observed 12.3, 218.4, 36.0 and 3.4 MB at lookup, about 270 MB in all). Fetch bound 300 s.
- **What leaves the machine:** four fixed public file requests; the servers learn the client address and which
  ecosystems' exports were fetched. No package name, version, graph, source, `.env`, evidence or report is sent. The
  OSV query API is not used.
- **Go modules:** phase A downloads core's modules and the pinned `x/vuln` module from the Go module proxy and
  verifies them with the checksum database, as every core image build already does (`core.Dockerfile`
  `go mod download`); this adds no new recipient. The module cache lives in a second owned volume
  `g012-gocache-<run>`. Go telemetry is turned off (`go telemetry off`), and phase B has no network.
- **Scans run with `--network none`:** osv-scanner (advisory volume read-only at `/db`), govulncheck phase B and the
  evaluator. Containers get no `-e` secret or `--env-file`, no Docker socket and no Compose network;
  `--cap-drop ALL`, `--security-opt no-new-privileges`, non-root where the image allows, the run label and a cidfile.
- **Bounds:** fetch 300 s, govulncheck phase A 300 s, osv-scanner 180 s, govulncheck phase B 240 s, evaluator 60 s;
  on expiry the container is killed and the check fails, never passes.
- **Cleanup:** both volumes and all containers are removed at the end of the check, and by the existing interrupted
  cleanup from their registry rows. Host files are limited to JSON outputs under `$WORK` (temporary).
- **Freshness and completeness:** each `all.zip` is OSV's complete export for that ecosystem and `vulndb.zip` the
  complete Go database, so coverage matches the online data for Go, npm and PyPI up to the export age that the 72-hour
  and 30-day checks bound. No transitive resolution is needed for lockfiles (`--offline` also disables it). The route
  meets the Goal, so **no owner data-sharing decision is requested**. Download cost is recorded per run (§14 E12).

### 5.3 Result semantics and identity coverage (r1)

| Source | Clean | Findings | Failure / incomplete (never passes) |
|---|---|---|---|
| osv-scanner | exit 0 | exit 1 with a parseable report | exit 127, 128, 129, 130 or any other code; unparseable or truncated JSON; an identity mismatch (below) |
| govulncheck (JSON) | exit 0, no `finding` | exit 0 with `finding` messages | non-zero exit; no `config` message, `protocol_version` other than `v1.0.0`, `scanner_version` other than v1.8.0, `scan_level` other than `symbol`, `db` other than the local file database; truncated stream |
| advisory fetch | all §5.2 checks pass | — | any §5.2 check fails |
| vulnpolicy | exit 0 | exit 1 (blocking finding, stale/unused/invalid exception) | exit 2 (any input invalid, missing, stale or incomplete) |

**Identity reconciliation (replaces r0's count equality).** The evaluator derives the expected identity set from the
same streamed lockfiles and compares it, as sets, with the identities the scanner reports (`--all-packages`):

- Identity is `(ecosystem, normalized name, resolved version)`: Go module path with a leading `v` stripped from the
  version; npm name from the entry's `name` field or its last `node_modules/` segment (scopes kept); PyPI names
  normalized per PEP 503 (lowercase, runs of `-`, `_`, `.` folded to `-`).
- **Project roots are excluded** and never count as coverage: the npm lockfile's root entry `""`, the uv project
  `intelligence` (`source = { virtual = "." }`, a virtual root, not an external dependency) and the Go main module.
  If the scanner reports a root, it is matched against this list and logged, not counted.
- **Repeated lockfile locations** (one npm name@version under several `node_modules/` paths) collapse to one identity;
  the location count is reported.
- **Allowed scanner-only identity:** Go `stdlib` at the `go` directive version (1.27.1). Nothing else may appear only
  on the scanner side.
- **Unsupported records** (npm `link: true` workspace entries, entries without a version, Git or URL sources, Go
  `replace` directives) are listed by identity and fail the check until classified by a reviewed Plan amendment; none
  are expected in the current lockfiles, and D0 confirms that.
- **Result:** any expected identity missing from the scanner side, or any unexplained scanner-side identity, is an
  incomplete result (exit 2) even when the totals are equal. govulncheck's `SBOM.modules` must be a subset of the
  `go.mod` requirements plus the main module and `stdlib`.
- **Reported per graph:** identities, locations, direct versus transitive (npm root dependencies, `pyproject.toml`
  dependencies, `go.mod` lines without `// indirect`), development/test membership (npm `dev` entries and the uv `dev`
  group; osv-scanner `dependency_groups`), toolchain coverage (Go standard library at go1.27.1) and unsupported
  categories.

**Unsupported, reported explicitly:** OS packages and container images, the Node.js and Python runtimes, the MinIO
source build's module graph, the validation-tools and skills/template graphs, browser/test images. No full
supply-chain or image audit is claimed.

### 5.4 Blocking policy

- Severity class from osv-scanner's group `max_severity` (CVSS base score: ≥ 9.0 Critical, 7.0–8.9 High,
  4.0–6.9 Medium, 0.1–3.9 Low); fallback to the advisory's `database_specific.severity` label; otherwise
  **Unclassified**.
- **Blocking:** Critical or High; any Go finding govulncheck reports at symbol level ("called"), regardless of severity;
  any Unclassified finding. Go findings are merged by OSV ID and aliases across both tools.
- **Reported, non-blocking:** Low and Medium findings that are not reachable Go findings; each is listed in the Result
  with an owner and action or an explicit reviewed follow-up.
- Reachability informs disposition only; the inventory is never reduced.
- **Exceptions** (`infra/security/dependency-exceptions.tsv`, created with a header only): one row per exact
  `(ecosystem, package, version, advisory ID or alias)` with rationale, named owner, expiry date, review condition,
  compensating control and a committed Codex review record path. Rejected: wildcards or ranges, missing fields, an
  expired row (stale; compared with the scan date in UTC), a row matching no current finding (unused) and any
  test-dependency exemption. No blanket baseline, suppression, automatic upgrade or weakened threshold. A change to this
  file is a security disposition that needs supervisor review; risk-acceptance changes escalate to the owner.

### 5.5 D0 — dependency baseline (first authorized implementation batch)

After release and the pre-batch reconciliation (§11.1), before other implementation: run the pinned tools (§5.1) with
the exact §5.2 route and isolation as standalone commands against the current tree. D0 **does not select tools**; it
collects the findings baseline and confirms the pinned contract on real runs. Record digests actually pulled, advisory
object sizes, MD5 and Last-Modified, `db_last_modified`, the identity reconciliation per graph (including the root
exclusions and any unsupported record) and every finding with class and reachability. A behavior that differs from
the §5.1/§5.3 contract (flag, exit code, schema, offline behavior) stops affected work and returns to Codex with the
observation; it is never resolved by changing the pin or relaxing a failure. No repository change is made by D0.

- No blocking finding: proceed.
- Blocking findings: correct within reviewed scope or obtain an explicit disposition before ACCEPT. A bounded
  correction is the smallest version change to the first fixed release within the same major line, regenerated in a
  container without scripts (`go get` + `go mod tidy`; `npm install <pkg>@<ver> --package-lock-only --ignore-scripts`;
  `uv lock --upgrade-package`), with the owning service's tests. Lockfiles are conditional allow-list entries
  (§11.7): the D0 triage and the remediation are handed to Codex for an allow-list amendment, while independent G012
  work continues. B-DEPS pins (exact go.mod lines, the web dependency lists, `uv.lock` equal to its render) then need
  matching reviewed edits. A major-version change, an API change or more than a few packages stops for review.

### 5.6 `B-DEPSCAN` (new required ID, `lint-base` suite)

`lint-base` holds the repository-wide guards (G051 input boundaries). B-DEPSCAN reads core sources and all three lock
inputs, so it belongs there rather than in foundation-only `lint`; it therefore also runs in any future affected
profile, where a core-only change can alter Go reachability. One check with two parts, both required:

1. **Scanner semantics on fixed offline fixtures** (`infra/security/testdata/`, no network, no live advisory drift):
   - govulncheck against a synthetic local database directory (`index/db.json`, `index/modules.json`,
     `index/vulns.json`, `ID/<id>.json` in the documented format) and a tiny fixture module whose `replace`d local
     dependency carries one synthetic advisory: one package calls the vulnerable symbol (called → blocking), one only
     imports it (imported → Unclassified → blocking unless excepted); a fixture module that fails to compile →
     failure.
   - osv-scanner `--offline` against synthetic per-ecosystem databases zipped at run time from committed OSV JSON
     records (by a test-support `fixture-db` subcommand of `vulnpolicy`, stdlib `archive/zip`) and tiny fixture
     lockfiles naming synthetic packages: one npm package with a CVSS vector scoring High, one PyPI package scoring
     Low, one PyPI package without severity, one clean package; a malformed lockfile → failure.
   - the evaluator over those real outputs with three exception files (none, a valid exact row, an expired row):
     blocking → fail, valid → excepted pass, expired → fail. The fixture outputs also reconcile by identity.
2. **Current-tree scan** of the three graphs against the downloaded public databases, through the evaluator: identity
   reconciliation must hold for Go, npm and PyPI, and any blocking or incomplete result fails the check.

Branch-level policy behavior is unit-tested in `vulnpolicy` (synthetic JSON): every class boundary, merge by alias,
reachable Go without severity, unused/wildcard/expired/mismatched exceptions, truncated or malformed inputs, exit-code
consistency, project-root exclusion, repeated-location collapse, the `stdlib` allowance, unsupported records and
**equal-count negatives** (one identity omitted while another is duplicated; one identity substituted by a different
version), which must fail although the totals match. Those tests run in the validation-tools check stage
(B-CONTRACT-TOOLS).

## 6. R5 — Encryption configuration and honest readiness

### 6.1 Matrix (recorded in `docs/SECURITY.md`; "evaluator" = §6.2 status with current capabilities)

| Path | Current actual mechanism | Local exception | Supported production configuration | Verification | Evaluator |
|---|---|---|---|---|---|
| browser → web | local HTTP on loopback; TLS only in the identity test proxy | `loopback-http` cookie profile (refused in production by core and web) | external TLS termination plus `https` profile and origins | external | external-declared with evidence, else blocked-unknown |
| web → core | HTTP `CORE_INTERNAL_URL`; core serves HTTP only; session headers cross it; scheme unchecked (F-2) | `http://core:8180` | `https` URL to an external TLS terminator (Node default verification) | client verifies; server side external | `http` → blocked-plaintext |
| core → PostgreSQL | `pgxpool.ParseConfig("")` defaults: opportunistic TLS without verification, plaintext fallback, ambient `PG*` input (F-1); local server has TLS off → plaintext | Compose `data` network | **none supported** (verified TLS needs a scope decision) | none | blocked-unsupported |
| migrate → PostgreSQL | explicit plaintext (`TLSConfig = nil`, no fallbacks) | as above | none supported | none | blocked-unsupported |
| core → Redis | readiness AUTH+PING over plain TCP; no other Redis use | `data` network | none supported | none | blocked-unsupported |
| core → object store | S3 client; `https` required with `APP_ENV=production` (config and objectstore) | `http://minio:9000` | `https` endpoint, Go default verification | verified | config-verified when `https` |
| signed-download endpoint | `OBJECT_STORE_SIGNING_ENDPOINT` rules (https in production; http data with https signing refused) | defaults to the data endpoint | `https` | browser verifies | config-verified when `https` |
| core → WorkOS | `https` API base (override only with `APP_ENV=test`); `https` issuer and JWKS; default transport roots | test overrides | defaults | verified | config-verified |
| core → intelligence | unauthenticated readiness probe over `INTELLIGENCE_URL` (unvalidated, F-3); intelligence serves HTTP | `http://intelligence:8000` | `https` URL to an external terminator | client verifies; server side external | disabled when unset; `http` → blocked-plaintext |
| intelligence → gateway | fixed `https://openrouter.ai/api/v1`, httpx default verification | none | as is | verified | disabled unless enabled; config-verified |
| PostgreSQL at rest | local Docker volume, no encryption setting | local fixture | provider/disk encryption | external | blocked-unknown without evidence |
| object storage at rest | local MinIO volume; R2 planned | local fixture | provider-managed encryption | external | blocked-unknown without evidence |
| backups at rest | none exist (G013) | — | G013 | external | blocked-unknown |
| Redis at rest | none (no persistence) | — | keep ephemeral | n/a | not applicable while ephemeral |
| CI daemon transport | job-private mutual TLS (ADR-002) | — | as is | verified | out of the product result; recorded |

Current production result: **blocked**. Named prerequisites: verified TLS for core and migrate to PostgreSQL and for
core to Redis (or removing Redis), TLS on internal hops, at-rest evidence for PostgreSQL and object storage, backups
(G013) and an operator deprovisioning entry point (F-6). Local Compose is a development fixture, not evidence.

### 6.2 Evaluator (`core/internal/config/readiness.go`) and command (`core/cmd/readiness`)

- `AssessReadiness(input map[string]string, caps Capabilities) (Readiness, error)`: pure, no I/O. Input is non-secret
  `KEY=VALUE` configuration (the runtime names: `APP_ENV`, `IDENTITY_COOKIE_PROFILE`, `IDENTITY_PUBLIC_ORIGINS`,
  `IDENTITY_WORKOS_API_BASE_URL`, `IDENTITY_WORKOS_ISSUER`, `CORE_INTERNAL_URL`, `OBJECT_STORE_ENDPOINT`,
  `OBJECT_STORE_SIGNING_ENDPOINT`, `INTELLIGENCE_URL`, `PG*` if present) plus finite enablement and evidence
  references `READINESS_EVIDENCE_<ITEM>` (edge TLS, internal TLS terminators, PostgreSQL/object/backup at rest).
- Object-store and origin rules reuse the package's existing unexported validators, so the accepted refusals remain
  the single source.
- Per-path status: `config-verified`, `external-declared`, `disabled`, `not-applicable`, `blocked-plaintext`,
  `blocked-verification-bypass`, `blocked-unsupported`, `blocked-unknown`, `blocked-local-exception`.
  Result: `configuration-ready` only when every enabled path is config-verified or external-declared and every
  at-rest item has a reference; otherwise `blocked`. Every report carries the label "configuration-only; external
  items declared, not verified; not a deployed-platform attestation". The evaluator never says "encrypted".
- `Capabilities` gives each path a mode (`in-platform`, `external`, `unsupported`). `CurrentCapabilities` pins today's
  code facts (PostgreSQL, migrate and Redis unsupported); a unit test pins its values, so changing one needs a
  reviewed code change. An evidence reference cannot lift an unsupported path.
- Refusals: any `PG*` setting (`PGSSLMODE` `disable`/`allow`/`prefer`/`require` → plaintext or verification bypass;
  any other value → unsupported: core has no reviewed TLS configuration, and ambient `PG*` input to
  `pgxpool.ParseConfig("")` ahead of the host override is not a supported path, F-1); boolean-like
  evidence (`true`, `yes`, `1`, …) is not a reference → blocked; `APP_ENV` other than `production` → blocked with the
  development allowances listed as local exceptions; a duplicate key → ambiguous input.
- **Secret refusal:** input naming any §2.1 secret, or a key containing `PASSWORD`, `SECRET`, `TOKEN` or `API_KEY`,
  is refused before evaluation.
- `core/cmd/readiness/main.go`: reads the input from stdin, prints the JSON report, exits 0 (configuration-ready),
  3 (blocked) or 2 (invalid input). Documented use runs it inside the network-free core check image. It is in no
  runtime image and adds no route.
- **No startup refusal is added.** Core still starts in production with plaintext stores (B-IDENTITY-CONFIG relies on
  that, and a refusal would be a behavior change outside R5); the evaluator is the readiness gate for a later
  deployment review.

## 7. R6 — Vendor and security records

### 7.1 `docs/foundation/VENDORS.md`

VENDORS.md is a toggle-rendered customization: the harness accepts exactly its render with one contiguous reviewed run
(`VENDORS_IDENTITY_PREFIX`, `VENDORS_IDENTITY_ROWS=5`, `VENDORS_IDENTITY_ROWS_SHA`) in place of the rendered identity
row (`run.sh` 416-420, `variants_off`). G012 extends **that run** directly after the five WorkOS rows, keeping every
rendered row byte-identical:

1. New rows in the existing six columns for the missing current integrations: core libraries (aws-sdk-go-v2 family,
   pgx, goose, MCP Go SDK, `x/time`), web direct and development dependencies, Python direct and development
   dependencies, images (python, uv, alpine, distroless, Playwright, docker CLI/DinD, buildpack-deps, yq, busybox,
   shellcheck, hadolint), services (GitLab.com and its hosted runners, container and package registries, the Go module
   proxy and checksum database) and the G012 additions (osv-scanner, govulncheck, the OSV per-ecosystem exports, the
   Go vulnerability database bulk copy).
   Version sources cite the lockfile or pin rather than copying transitive versions.
2. A section "Integration data handling, credentials and review status (G012)" with one row per integration that
   receives data or holds a credential: version source, purpose and environment, personal data and categories sent,
   the owner and privilege of any credential involved; processor and licence review status, processing locations and
   subprocessors, retention/deletion and data-use/training terms, transfer/agreement review, deprovisioning
   responsibility and production prerequisites. Unknown or unreviewed items are written as production prerequisites;
   nothing is invented. Identity, AI-gateway and advisory metadata flows are recorded as they behave today (gateway
   disabled, no live AI traffic).
   Intelligence-only rows say "intelligence-enabled renders only".
3. A heading and the table header that continue the rendered inventory, so the rendered rows that follow
   (OpenRouter, FastAPI, Go, Node, gitleaks) stay unchanged.

The harness change is two constants (`VENDORS_IDENTITY_ROWS` and its SHA) with updated comments and PASS text; the
composition logic and its detector controls are unchanged. The render register reason gains "G012".

### 7.2 `docs/SECURITY.md` (currently empty; Spec R6 assigns it)

Contents: scope and non-claims (no SOC 2 or GDPR compliance claim; configuration-only evidence); reporting and triage
route (private report to the repository owner, confidential issue, owner/Codex disposition; no invented service
levels); control policy (submitted-source and evidence secret scanning, static rules and limits, dependency policy and
the exception process, change evidence through Git, MR, Codex/Duo dispositions, required CI and G011 audit); the R1
inventory; the R2 matrix, rotation/withdrawal procedures and the deprovisioning runbook; the R5 matrix, evaluator usage
and named production blockers; the findings F-1..F-8 with their dispositions; G013 handoff (responsibilities for
encryption, vendors and stored credentials, plus the audit maintenance boundary); shipped controls versus
adopter-owned legal, deployment and application privacy work; G056 named as later owner of assembled-product rights
work, not a predecessor.

`docs/SECURITY.md` is under B-PROTECTED (docs outside `docs/foundation`). CI compares with the candidate itself; local
runs with `--base <target>` pass `--expect-hash docs/SECURITY.md=<sha256>`, recorded in the Result.

Other docs: OPERATIONS.md (new checks, counts, the ops rotation step), FOLLOW-UP.md (STAGE-OPS status), README (scan
line and a sentence), DATABASE-ISOLATION.md (one pointer: provisioning never rotates; see SECURITY.md).

### 7.3 Findings recorded without a code change (dispositions for Codex review)

| ID | Finding | Disposition in G012 |
|---|---|---|
| F-1 | core's pool keeps pgx's implicit TLS (unverified) with plaintext fallback to the default host and ambient `PG*` input; migrate pins plaintext explicitly | readiness block; follow-up with verified TLS under a reviewed scope |
| F-2 | web's `CORE_INTERNAL_URL` scheme is unchecked; production `http` is not refused | readiness block; follow-up |
| F-3 | `INTELLIGENCE_URL` unvalidated; core loads the intelligence token but sends it nowhere | inventory note; STAGE-AI envelopes own the real use |
| F-4 | minio-init passes the scoped user's secret in `mc` argv (local fixture, rendered file) | local exception; follow-up |
| F-5 | cookie-profile defaults differ outside production (core `https`, web `loopback-http`) | note; production behavior agrees |
| F-6 | `DisableUser` and membership removal have no operator entry point | production prerequisite (G018/G019/G021) |
| F-7 | `<db>_retention` has LOGIN and a password but no consumer or privilege until G013 | note for G013 |
| F-8 | base images mostly tag-pinned; unscanned graphs listed in §5.3 | follow-up (image/supply-chain scope) |

## 8. Files and modules

| Path | Change |
|---|---|
| `core/internal/config/secrets.go`, `secrets_test.go` | **new**: sentinels, `resolveSecret`; tests (§10) incl. the boundary test |
| `core/internal/config/config.go` | one hunk: `EnvSecrets` wraps `ErrSecretNotSet`; `secret` uses `resolveSecret` |
| `core/internal/config/readiness.go`, `readiness_test.go` | **new**: evaluator, `CurrentCapabilities`, fixtures |
| `core/cmd/readiness/main.go`, `main_test.go` | **new**: stdin command |
| `infra/security/secret-scan.sh`, `infra/compose/secret-scan.yml` | **new** (§4.2) |
| `infra/security/dependency-scan.sh`, `infra/security/dependency-exceptions.tsv`, `infra/security/testdata/…` | **new** (§5): pins, fetch, scans, reconciliation hand-off; exceptions file with a header only; fixed offline fixtures (synthetic OSV JSON records, a synthetic Go database, tiny lockfiles and a fixture module) |
| `infra/validation/tools/cmd/vulnpolicy/*.go` (+ `testdata/`) | **new** evaluator, identity reconciliation, a test-support `fixture-db` subcommand and unit tests (stdlib only; tested by the existing check stage) |
| `infra/validation/run.sh` | `static_secret_source` (B-SECRET-SOURCE) in `suite_lint`; `static_depscan` (B-DEPSCAN) in `suite_lint_base`; `credential_rotation` (B-SECRET-ROTATION) in `variants_ops`; B-SHELL script list; `APPLICABLE` += the new core files (compile parity of the intelligence-disabled variant; the root-only `infra/security` tooling is not added); VENDORS constants and PASS text; the IMAGES INFO list reports the scanner pins |
| `infra/validation/required-checks.tsv` | +3 IDs; header totals |
| `infra/validation/render-customizations.tsv` | reason text for config.go, VENDORS.md, FOLLOW-UP.md (row count stays 26) |
| `infra/ci/validate-controls.sh` | count pins only (§11.3) |
| `README.md` | scan line; one sentence |
| `docs/SECURITY.md` | populated (§7.2) |
| `docs/foundation/{VENDORS,OPERATIONS,FOLLOW-UP,DATABASE-ISOLATION}.md` | §7 |

**Unchanged:** `core/cmd/server/main.go` (digest pin), `LoadIdentity`, migrate, intelligence and web code,
`infra/compose/local.yml`, GETTING-STARTED, `.env.example`, `infra/security/gitleaks.toml`, `.gitlab-ci.yml`,
`infra/ci/{lib.sh,profile-paths.tsv,select-profile.sh,validate.sh,pipeline-union.sh}`, migrations, contracts, AUDIT.md,
lockfiles and manifests (unless a D0 remediation is approved).

## 9. Schema, contracts and interfaces

- **Database/schema/migrations: none.** No role, grant, policy or event change. The rotation check alters only the
  disposable copy's role password and is torn down with the copy.
- **Public API/MCP/BFF contracts: none.** New internal Go API: config sentinels and `AssessReadiness`; a test-only
  provider; the `core/cmd/readiness` tool (no runtime image, no route).
- **Developer-facing change:** the README check's scan line and two new scripts. Live services are unchanged except
  clearer, closed secret-error texts at startup.

## 10. Tests and acceptance-criteria mapping

**Unit (core, stack-free, in the core check image):**
- U-SEC-1: `EnvSecrets` success (existing `TestLoadValid`) and missing → `<NAME> is not set`.
- U-SEC-2: a map-backed replacement provider supplies all four `Load` secrets while the environment holds none →
  success; values come from the provider and appear in `SecretValues`; `EnvSecrets` over the same environment fails.
- U-SEC-3: table over every §2.2 outcome, with provider error text carrying a generated canary → the exact closed text,
  no canary or other secret value in the error, `Config{}` returned.
- U-SEC-4: `INTELLIGENCE_URL` unset → the token is never requested (spy); set with an unavailable provider → failure.
- U-SEC-5: exactly one provider call per required name.
- U-SEC-6 boundary: no non-test file outside `core/internal/config` and `core/cmd/server/main.go` references
  `SecretProvider`, `EnvSecrets` or calls `.Secret(`; inside config only `resolveSecret` calls it.
- U-RDY-1..n: current production-shaped fixture → blocked with the exact prerequisite set; development fixture →
  blocked plus local exceptions; negatives for `http` internal URLs and object endpoints in production, `PG*` values,
  `loopback-http`, the test API-base override, boolean evidence, missing at-rest evidence and secret-named input;
  synthetic positive with injected capabilities and every reference → `configuration-ready` with the label;
  `CurrentCapabilities` pinned; command exit codes 0/3/2.

**Unit (validation tools):** `vulnpolicy` policy branches (§5.6). **Harness:** B-SECRET-SOURCE (§4.3), B-DEPSCAN
(§5.6), B-SECRET-ROTATION (§3.3). Existing WorkOS file-failure tests, redaction and exit tests are reused unchanged.

| AC | Evidence |
|---|---|
| AC1 | §2.1 inventory checked against config, Compose, call sites and the identity schema (recoverable versus hash-only fields; S-review); U-SEC-2 (replacement) and U-SEC-1 (EnvSecrets) through the same `Load`; U-SEC-5/6 (startup-only acquisition); disabled paths labelled; no new runtime service |
| AC2 | U-SEC-3/4, `Config{}` plus `main.go`'s pinned order; existing `identity_config_test.go` file-failure cases; B-SECRET-CANARY (positive control and output scan) |
| AC3 | B-SECRET-ROTATION R0–R6 with durations and backend counts; scoped teardown in the cleanup receipt; synthetic values only |
| AC4 | §3.1 matrix; unchanged accepted proofs run in the final batch (B-IDENTITY-IT, B-AUDIT-IT AT-ID/AT-ATOM, B-POLICY-IT, B-RLS-IT, B-OBJ-IT incl. G010's attack suite); B-COMPOSE, B-MIGRATE-ENV, B-IDENTITY-GUARDS; §3.4 runbook mapped to G011 events and the M5 exception |
| AC5 | README scan through B-UNIT-CHECKS/B-CHECKS-ROOT (the real tree, with its tracked link); B-SECRET-SOURCE S1–S7 (link accounting, outside target not followed, canary, exemption and incomplete-stream failures); lint and gitleaks configuration not weakened |
| AC6 | B-DEPSCAN current-tree reports with identity reconciliation of all three graphs (roots excluded, development/test membership, toolchain coverage) and pinned tool and advisory-object provenance; fixed offline fixture semantics; `vulnpolicy` units incl. equal-count omission/substitution negatives; D0 triage and every real finding corrected or disposed |
| AC7 | §6.1 matrix; U-RDY negatives and labelled synthetic positive; no claim about local volumes or links |
| AC8 | VENDORS/SECURITY content compared with manifests, images, CI and §2.1/§6.1; F-1..F-8; the private-backup material handed to G013; G013/G056 and adopter responsibilities explicit; no schema, provider, product surface, MFA, audit bypass or compliance claim |
| AC9 | §11 wiring, actual full + ci-controls selection, final local run, MR fast, staging full gate; failure paths and output privacy preserved; provenance per batch |
| AC10 | S-review, Result with AC map, B5 ledger, cleanup receipt, staging disposition; PROGRESS/G013 handoff by Codex; G021 closed |

## 11. Validation

### 11.1 Before every batch

`git fetch origin`; apply the exact-delta rule (record SHAs and reason) or rebase and reconcile. A management-only
delta is still an input to the repository-wide secret scans (G011 lesson): run the README secret-scan line on the
rebased tree before carrying evidence forward. Then `bash infra/ci/docker-desktop-mounts.sh` (stop on failure), take the
harness flock and record target/candidate SHAs. The Result and allow-list stay at base bytes while validating
(B-PROTECTED); `docs/SECURITY.md` uses `--expect-hash`. Standard form: `bash infra/validation/run.sh --base <target>
--allow-file management/v0/results/G012-scope-allowlist.txt [--expect-hash docs/SECURITY.md=<sha>] --report-dir
<EVID>/<batch> <suites>`.

### 11.2 Batches

| Batch | Content | Purpose | Wall (forecast) |
|---|---|---|---|
| D0 baseline | pinned tools (no selection): advisory fetch, osv-scanner (three graphs) and govulncheck (core) standalone with the §5.2 route; contract confirmation | first authorized batch; findings and identity baseline, triage (§5.5) | 10–20 min (about 270 MB of advisory downloads plus two image pulls) |
| b1 focus (stack-free) | core check image (gofmt, vet, tests incl. U-SEC/U-RDY), validation-tools check stage, then `run.sh … lint lint-base unit contracts provenance` (B-SECRET-SOURCE, B-DEPSCAN, B-SHELL, B-COMPOSE, B-DEPS, B-SECRET-CANARY with lint, README check line, B-CONTRACT-TOOLS, B-RENDER-PLAN) | inner loop | 15–20 min per run, 2–4 runs |
| b2 focus (stack) | `run.sh … ops off` (B-SECRET-ROTATION and the inherited ops checks; VENDORS composition and `APPLICABLE` in B-TOGGLE-OFF-ADDITIONS) | inner loop | 25–35 min per run, 1–2 runs |
| S-review | independent read-only security review (§12) | fix; rerun affected b1/b2 only | — |
| final | reconcile, then `all unit` (ci-controls auto-selected by infra/README changes) | the **only** broad run; every required ID PASS once (B-REQUIRED-CHECKS); a failure reruns only the failed suites and checks its fix affects | 60–70 min |

**No manual or operator-feedback stage** is required: no UI, no live provider, and the rehearsal is automated on a
disposable fixture. Total local machine time without failures: about 2–3 h. Unchanged passing batches are not repeated
for a SHA-only or management-only delta, except the scan-input rule above.

### 11.3 Harness wiring and count pins

New IDs: `B-SECRET-SOURCE` (lint), `B-DEPSCAN` (lint-base), `B-SECRET-ROTATION` (ops). From the current 107 rows:

| Pin | Now | After |
|---|---|---|
| `required-checks.tsv` rows / lint / lint-base / ops | 107 / 8 / 3 / 12 | 110 / 9 / 4 / 13 |
| profiles fast / integration / full | 33 / 71 / 100 | 35 / 74 / 103 |
| VP1 fast `required_count` (validate-controls 948, 952) | 33 | 35 |
| VA7 and the remainder-row check (1112, 1115, 1948) | 72 | 73 |
| VU1 full split (1206) | 33/72/101 | 35/73/104 |
| VU9 integration split (1207) | 33/42/71 | 35/43/74 |
| VA1 (1070, 1079) and J7 (1937) affected web | 19 | 20 (lint-base is in every affected profile) |
| OPERATIONS.md 90/91/92/93/99–100/120/121 | 33; 19/15/19; 71; 100; lint "eight", lint-base three; 33/72/101; 33/42/71 | 35; 20/16/20; 74; 103; "nine" and B-DEPSCAN added; 35/73/104; 35/43/74 |
| render register count (run.sh 1153) | 26 | 26 |

The implementer inventories every pin at binding with `grep -nE '\b(15|19|33|42|71|72|100|101|107)\b'` over
`infra/ci/validate-controls.sh`, `infra/validation/required-checks.tsv` and OPERATIONS.md, and records each change in the
Result. CI meta-tests run because `infra/` and README changed (current selection), not as an extra campaign.

### 11.4 CI selection (actual paths)

`infra/` and `core/internal/config/` select **full**; `infra/*` and `README.md` also require **ci-controls**;
`core/cmd/readiness/` is runtime (integration) and is subsumed. No `profile-paths.tsv` change. Documentation-only
commits keep the pipeline bypass. No affected-only hosted run is promised. Partitions: B-SECRET-SOURCE and B-DEPSCAN
run only in the fast job (`lint`, `lint-base`); B-SECRET-ROTATION only in the gate (`ops`); the README scan keeps its
existing fast (`unit`) and gate (`root`) executions. No service scan is added to both partitions.

### 11.5 Hosted

1. After Codex's pre-integration review: reconcile, push, `glab` MR to `staging` with source-branch deletion; auto-merge
   armed once the Codex/Duo holds clear.
2. MR fast runs automatically (it includes B-DEPSCAN, which downloads the public advisory exports; its scans run without
   network); the optional manual gate is not requested.
3. The post-merge staging pipeline (full gate) is **required**: success or owner-canceled satisfies it (canceled
   recorded as canceled); any failed job fails; pending does not qualify.

### 11.6 Provenance and evidence

Git-first. Per batch: candidate/target commits, commands, outcomes (failed, interrupted and skipped explicit), image IDs
(core check, validation tools, gitleaks, osv-scanner, golang), scanner versions, each advisory object's size, MD5 and
Last-Modified, `db_last_modified`, the identity reconciliation summary, Go version, CI links. Durable proof lives in
the Result and CI artifacts; local raw output is temporary and purged at closeout. Never retained: secrets, canary
values, credentials. The Result is a scanned file:
it quotes no flagged phrase and is checked with the README scan line before commit.

### 11.7 Proposed B-SCOPE allow-list (`management/v0/results/G012-scope-allowlist.txt`, imported by Codex at binding)

```text
# Create
^core/internal/config/(secrets|readiness)(_test)?\.go$
^core/cmd/readiness/main(_test)?\.go$
^infra/security/(secret-scan|dependency-scan)\.sh$
^infra/security/dependency-exceptions\.tsv$
^infra/security/testdata/(osv|govulncheck|exceptions)/[A-Za-z0-9._/-]+$
^infra/compose/secret-scan\.yml$
^infra/validation/tools/cmd/vulnpolicy/[a-z]+(_test)?\.go$
^infra/validation/tools/cmd/vulnpolicy/testdata/[A-Za-z0-9._/-]+$
^management/v0/results/G012-(scope-allowlist\.txt|secrets-secure-engineering-result\.md)$
# Modify: production
^core/internal/config/config\.go$
# Modify: harness, build and pins
^infra/validation/(run\.sh|required-checks\.tsv|render-customizations\.tsv)$
^infra/ci/validate-controls\.sh$
# Modify: docs
^README\.md$
^docs/SECURITY\.md$
^docs/foundation/(VENDORS|OPERATIONS|FOLLOW-UP|DATABASE-ISOLATION)\.md$
```

Conditional entries (only with a Codex-reviewed amendment after D0 triage or a review finding): `core/go.mod`,
`core/go.sum`, `web/package.json`, `web/package-lock.json`, `intelligence/pyproject.toml`, `intelligence/uv.lock` (with
the matching B-DEPS edits in `infra/validation/{run.sh,lib.sh}` and register rows), and
`infra/ci/{lib.sh,profile-paths.tsv}` only if the fast-budget fallback (§14 R-a) is chosen.

## 12. Delegation and sequence

- **Implementer:** one owner of all files (task G012, role impl; `g012-agent-<id>-impl` or the native-name fallback,
  recorded). Default effort high; the operator may choose xhigh. No file split: config, scripts, harness wiring and docs
  overlap.
- **Delegate 1 (required), read-only security review** after b2: closed secret errors and no fallback; boundary test;
  submitted-set listing, link accounting (no dereference) and the exemption guard; the download-only advisory route
  (only the four fixed objects fetched; scans without network; freshness, integrity and cleanup); identity
  reconciliation and root exclusions; policy and exception strictness;
  rotation fixture correctness and synthetic-only values; readiness labels and refusals; the risky-construct grep
  (§4.4); privacy of logs and evidence. Findings are fixed by the implementer.
- **Delegate 2 (optional), read-only scanner/configuration analysis** in D0 only if the baseline has many findings
  (triage and remediation options). Both share B5; decisions stay with Codex/owner.
- The coordinator owns the shared harness lock and stack scheduling.

**Sequence:** (1) binding: reconciliation against the then-current target (G011 final ACCEPT is recorded; A-G011-1..4
hold), allow-list and initial Result import;
(2) D0 baseline and triage (remediation proposal to Codex if needed, independent work continues); (3) R1 code and U-SEC;
(4) R5 evaluator, command and U-RDY; (5) R3 script, overlay, README line, B-SECRET-SOURCE; (6) R4 script, `vulnpolicy`,
exceptions file, fixtures, B-DEPSCAN; (7) B-SECRET-ROTATION; (8) wiring, pins, `APPLICABLE`, register reasons, VENDORS
constants; (9) docs; (10) b1/b2 loops → S-review → fixes and affected reruns → reconcile → final → Result;
(11) safekeeping push after each round (no MR) → Codex review → MR.

## 13. Budget, forecast and scope

**B5**, shared attempts, checkpoint after attempt 3; planning spends none. **Forecast: 2 likely, 3 plausible.**

| Attempt | Expected content |
|---|---|
| 1 | D0 through the final run, review-ready |
| 2 | one bounded Codex correction (likely for a security-tier Goal with new tooling) |
| 3 | only if D0 findings need a separately reviewed remediation, or hosted advisory/network drift appears late |
| 4–5 | not expected; the checkpoint after 3 reassesses convergence and scope |

**Context and scope:** one cycle (estimated diff about 2.5–3.5k lines including tests, fixtures and docs). Neither cloud
deployment nor whole-stack transport change is needed: transports are recorded and blocked, not repaired. **No split is
recommended.** If the checkpoint shows risk, the split would be G012a (R1, R2, R5, R6: resolver, rehearsal, readiness,
records) and G012b (R3, R4: scanners and their wiring), with no weakened checks.

## 14. Escalation items for explicit review; risks

| # | Item | Class |
|---|---|---|
| E1 | **r1 disposition applied:** online OSV queries are not used. Download-only public advisory data: four fixed HTTPS objects (three OSV per-ecosystem exports, the Go database bulk copy), no dependency data sent, scans with network none, bounded size, time, integrity and freshness, owned cleanup (§5.2). No account, cost or production relationship. No owner data-sharing decision is requested | advisory data source (download only) |
| E2 | **r1 disposition recorded:** `osv-scanner` and `govulncheck` are containerized test tooling within the Spec (no runtime ADR), pinned exactly (§5.1). B-DEPSCAN is the first source check that downloads data (`lint-base`, so it runs in MR fast) | tooling; CI behavior |
| E3 | Blocking policy (Critical/High, reachable Go, Unclassified) and an exception mechanism (exact, expiring, reviewed; unused or expired fails; file created empty) | security policy / suppression control |
| E4 | Source-scan change: submitted-set listing, exemption guard, README line and overlay; gitleaks configuration unchanged | security control (strengthening) |
| E5 | Rehearsal runs `ALTER ROLE` and `pg_terminate_backend` as the disposable superuser in the `ops` copy; synthetic values only | disposable privileged test action |
| E6 | +3 required IDs and count pins; VENDORS run constants; `APPLICABLE`; register reasons; B-SHELL list | validation-scope change |
| E7 | Readiness evaluator and command: current production result is **blocked** with named prerequisites; never an encryption claim | readiness truth |
| E8 | Findings F-1..F-8 recorded without code change, notably F-1 (implicit PostgreSQL TLS/fallback) and F-2 (unchecked internal URL) (Q5) | security observations |
| E9 | D0 remediation may need lockfile and B-DEPS changes (conditional allow-list entries) | dependency change (conditional) |
| E10 | `docs/SECURITY.md`, a canonical source, populated by Claude as Spec R6 directs (Q7) | governance document |
| E11 | **None:** schema/migration, role/grant, public contract, access to or rotation of real credentials; destructive change, finding suppression, production deployment | explicit none |
| E12 | Performance: no query, index, helper or lock change (PERFORMANCE register unaffected). CI cost: about +3–6 min on fast (about 270 MB of advisory downloads, two image pulls, module download and analysis), +1–2 min on ops; temporary storage under 1 GiB in two owned volumes; an unavailable or stale advisory export fails the check by design | performance policy |

**Risks:**
- R-a: fast-job budget (20 min harness cap). Mitigation: D0/b1 measure B-DEPSCAN; if fast would exceed the budget,
  propose moving it to the gate with a reviewed amendment, never skipping it.
- R-b: availability and freshness of the advisory exports in hosted CI. Mitigation: bounded fetch, integrity and age
  checks; failure is reported, not retried silently, and attributed in the Result; the 72-hour and 30-day bounds are
  recorded with each run.
- R-c: the pinned versions behave differently from the recorded contract. Mitigation: D0 and the fixtures confirm it
  before wiring; a mismatch returns to Codex (§5.5), and no failure is relaxed.
- R-d: the VENDORS run constant or `APPLICABLE` edit breaks B-TOGGLE-OFF-ADDITIONS. Mitigation: b2 runs `off` early.
- R-e: generated text (Plan, Result, docs) triggers gitleaks `generic-api-key`. Mitigation: names in backticks, no
  keyword-separator-token prose; README scan before each commit.

## 15. Open questions for Codex

- Q1: keep the WorkOS key on its G004 value/`_FILE` contract outside `SecretProvider` (the mounted-file route is the
  substitution path), rather than routing it through the provider. Plan default: keep.
- Q2: **resolved by the r0 review:** online OSV queries not approved; r1 uses the download-only route (§5.2).
- Q3: B-DEPSCAN in `lint-base` (fast job, so MR time sees new advisories) or gate-only. Plan default: `lint-base`.
- Q4: **resolved by the r0 review:** exact versions, digests and the CLI/output contract are pinned in §5.1; D0 only
  collects findings and confirms behavior.
- Q5: F-1..F-8 recorded without code change. Plan default: yes; any bounded fix Codex directs becomes a scope item.
- Q6: unused exceptions fail. Plan default: yes.
- Q7: Claude populates `docs/SECURITY.md` per Spec R6, reviewed by Codex. Plan default: yes.

## 16. Closeout and handoffs

- Bounded cleanup: one scoped preflight, one removal pass, compact receipt; run-labelled containers, the ops copy and
  temporary fixtures recorded at creation; uncertain items held; native refusals reported once.
- **G013** receives: the credential inventory and rotation procedures, the encryption/at-rest matrix and its blockers,
  the vendor responsibilities, the backup-at-rest prerequisite and the audit maintenance boundary (no G012 change to
  it). **Private backup material (r1):** a PostgreSQL dump or backup contains recoverable `pkce_verifier` and
  `csrf_token` values for unexpired rows, SHA-256 session, state, binding and logout-reference values, provider
  session and user identifiers, names and audit rows. It must stay private operational material (owned storage,
  access-restricted, no CI artifact or committed fixture, encrypted at rest in production), and a restore must not
  revive sessions or login transactions, which G013 R2 (private handling) and R4 (restored-auth invalidation)
  already require; G012's §2.1 classification is its input.
- G012 does not close EPIC-004 (G013 remains). G021 stays closed. No G013 specification is implied.

## Revision history

- r0, 2026-10-15 UTC: initial provisional Plan on staging `795b95a` (G011 merged at `1af4696`, final ACCEPT pending).
  Read-only inspection plus two read-only exploration delegates; no probe, install, build, scan, test, Docker or harness
  command was run.
- r1, 2026-10-15 UTC: bounded revision for the r0 review (not included) CORRECTION LOOP,
  rebased onto staging `bf462f6` (management-only deltas `ee3fcac` and `bf462f6`); A-G011-1..4 bound to the accepted
  final G011. Four corrections and the E1/E2 dispositions (change map at top): exact tool pins and contract from
  read-only official metadata (§5.1); regular-file and link accounting with a
  link-bearing clean fixture (§4.2, §4.3); identity-based dependency coverage with root exclusion and equal-count
  negatives (§5.3, §5.6); accurate recoverable versus hash-only authentication fields and the G013 private-backup
  handoff (§2.1, §16); download-only advisory databases instead of online queries (§5.2). r0 archived byte-identically
  (`21d73dff…`). Metadata lookups only (release endpoints and docs, module proxy and checksum database, registry
  manifests, HEAD of the advisory objects); nothing was pulled, installed, built, scanned or run. Ten ACs and B5 with a
  checkpoint after attempt 3 are unchanged.
