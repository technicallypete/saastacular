# Security controls and operating records

SaaStacular is a single-tenant and multi-tenant security foundation for building AI operating systems. Its value is the
reviewed security boundaries it ships (identity and sessions, tenant isolation through shared authorization and
PostgreSQL row-level security, least-privilege service credentials, durable audit, guarded recovery and retention) and
the repeatable checks that exercise them. This document records those engineering security controls as they behave in
this repository.

It is configuration and engineering evidence only: it is not a SOC 2, ISO 27001 or GDPR compliance claim, not a
deployed-platform attestation, and no production deployment exists. Where a control depends on a deployment that does
not exist yet, it is listed as a production prerequisite. Vendor and data-handling records are in
[VENDORS.md](foundation/VENDORS.md); audit events in [AUDIT.md](foundation/AUDIT.md); database roles in
[DATABASE-ISOLATION.md](foundation/DATABASE-ISOLATION.md); identity in [IDENTITY.md](foundation/IDENTITY.md). Before
using real personal data, work through [PRIVACY-CHECKLIST.md](PRIVACY-CHECKLIST.md); for deployment, see
[DEPLOYMENT.md](DEPLOYMENT.md).

**Not executed for this repository, and not claimed:** live identity-provider qualification of the signup flow (real
WorkOS wildcard, DNS and HTTPS behaviour on tenant hosts) has not been executed; live custom-domain qualification (real
DNS, certificates and the production edge log configuration) has not been executed; no end-to-end acceptance of the
foundation as a whole is claimed; and no hosted CI run or provider deployment has been performed for this copy. The
local, emulator and fixture evidence described in the foundation guides does not substitute for those checks.

## 1. Reporting and triage

- Report a suspected vulnerability or exposed credential privately to the maintainers, through the private
  vulnerability-reporting channel of the platform where you obtained this repository (or a confidential issue in your
  own project for your fork). Do not open a public issue, pull request or merge request that describes it.
- The maintainers decide the disposition: fix, reviewed exception (section 4.3) or accepted risk. Security-model or
  risk-acceptance changes are maintainer decisions.
- An exposed credential is rotated or withdrawn with the procedures in section 6; rotation of a real credential is an
  operator action, never part of automated validation.
- No response-time commitment is defined.

## 2. Control policy

| Control | Mechanism | Evidence |
|---|---|---|
| Submitted-source secret scanning | `bash infra/security/secret-scan.sh --env-file .env secret-scan` (README `platform:check`): exactly the submitted files listed by host Git (tracked plus untracked, non-ignored), streamed into the pinned gitleaks container without a bind mount or network; links stored, never followed, their target text scanned; untracked nested repositories skipped; gitlinks refused; a submitted file under a local-only exemption (`.env`, `.env.local`, `node_modules/`, `.next/`) fails before scanning | B-UNIT-CHECKS and B-CHECKS-ROOT (the real tree), B-SECRET-SOURCE (fixtures S0 to S7) |
| Evidence and log secret scanning | `infra/ci/validate.sh` scans every published artifact against the run's generated secrets and canaries; the harness scans its own logs | B-SECRET-CANARY, CI controls V5a to V6d |
| Static analysis | Section 2.1 | B-SHELL, B-HADOLINT, lint suites, unit checks |
| Dependency vulnerabilities | Section 4 | B-DEPSCAN |
| Change evidence | Your repository's history and review records, required CI (`.gitlab-ci.yml`: the fast job on merge requests, the integration/full gate where configured), and the durable audit trail for runtime authority changes | your review records, CI artifacts, `audit.events` |

The secret scan lists submitted files with Git. In a copy without Git metadata, initialize a fresh local repository
first (the README's "Developer checks need a Git repository" section).

### 2.1 Static checks

| Service | Retained checks | Security-relevant rules | Limits |
|---|---|---|---|
| core | gofmt, `go vet` (default analyzers), unit tests; B-MCP-SOURCE, B-IDENTITY-GUARDS, boundary unit tests (including the secret-acquisition boundary) | vet's correctness analyzers; project guards on SDK and credential placement, bypass symbols, route and import sets | no security-specific analyzer (gosec class) and no taint analysis |
| web | ESLint `next/core-web-vitals` and `next/typescript`, `tsc`; B-WEB-BOUNDARY, B-IDENTITY-GUARDS | React rules; server-only boundary; no credential in browser storage | no security lint plugin; relies on React escaping |
| intelligence | ruff `E,F,W,I,B,UP,S,SIM,RUF` (tests ignore `S101,S105,S106,S603`), pytest | flake8-bandit (`S`) rules | rule-based only |

The secure-engineering review searched for shell execution, string-built SQL, `dangerouslySetInnerHTML`, `eval` and
disabled certificate verification; only a demonstrated gap would justify a new analyzer.

## 3. Credential inventory

All configuration is read once at process start; no service reloads it. "Restart" means a new process.

| Name | Consumer and use | Source | Required when | Missing or invalid | Rotation owner | Exposure restriction |
|---|---|---|---|---|---|---|
| `DB_PASSWORD` (Compose maps `DB_RUNTIME_PASSWORD`) | core pool as `<db>_runtime` | `SecretProvider` (`EnvSecrets`) | always | startup fails, name only | platform operator | core and postgres only (B-COMPOSE); redacted from logs |
| `DB_MIGRATOR_PASSWORD` | one-shot migrate as `<db>_migrator` | migrate configuration (environment) | every migrate run | refused before connecting | platform operator | the migrate environment is exactly five names (B-MIGRATE-ENV); redacted |
| `DB_RUNTIME_PASSWORD`, `DB_RETENTION_PASSWORD`, `POSTGRES_SUPERUSER_PASSWORD` | role creation (`10-roles.sh`, `provision-roles.sql` through psql `\getenv`) | environment | first volume initialization or provisioning | initialization refuses; provisioning never changes an existing password | platform operator | postgres service only |
| `REDIS_PASSWORD` | core readiness check (AUTH and PING only); the Redis server | `SecretProvider` | always | startup fails | platform operator | redis and core |
| `OBJECT_STORE_SECRET_KEY` | core S3 client (readiness HeadBucket today); minio-init creates the scoped user | `SecretProvider` (core) | always | startup fails | platform operator | core and minio-init; passed in `mc` argv inside minio-init (F-4) |
| `MINIO_ROOT_PASSWORD` | minio, minio-init | environment | local fixture | Compose refuses | platform operator | local fixture only |
| `DB_RETENTION_PASSWORD`, `POSTGRES_SUPERUSER_PASSWORD`, `MINIO_ROOT_PASSWORD`, `OBJECT_STORE_SECRET_KEY` (recovery operator use) | `infra/recovery/recovery.sh`: the retention login for retention; the cluster superuser for capture, restore, invalidation, replay and ledger sync; the fixture MinIO administrator for the capture fence and capture; the scoped object value for restore uploads, replay deletes and purge checks ([RECOVERY.md](foundation/RECOVERY.md)) | the operator's env file, passed to one-shot containers only by `--env-file` process substitution | an operator command | refused before any change | platform operator | operator host only; never argv, output or a file outside the operator's private directories; nothing reaches core, web or intelligence |
| `INTELLIGENCE_SERVICE_TOKEN` | intelligence verifies `/internal/*` in constant time; core loads it but sends it nowhere today (F-3) | core `SecretProvider` when `INTELLIGENCE_URL` is set; intelligence environment | core: with `INTELLIGENCE_URL`; intelligence: optional | core startup fails; intelligence warns and answers 401 | platform operator | core and intelligence |
| `OPENROUTER_API_KEY` | intelligence gateway adapter | environment | optional; unset is the documented disabled state | `gateway_not_configured` per request | platform operator | intelligence only; never generated by `configure-env.sh` |
| `IDENTITY_WORKOS_API_KEY` or `IDENTITY_WORKOS_API_KEY_FILE` | core WorkOS adapter | `LoadIdentity` (value or mounted file, mutually exclusive; the identity contract, outside `SecretProvider`) | all three identity values or none; production requires them | startup fails naming variables; file failures sanitized | platform operator (WorkOS dashboard) | core only; never web or intelligence (B-COMPOSE, B-IDENTITY-GUARDS) |
| `IDENTITY_TEST_API_KEY` and per-run emulator material | identity test overlay | generated per run | tests only | not applicable | harness | test overlay only |
| `IDENTITY_STAGING_WORKOS_API_KEY` | `TestStagingContract` | protected, environment-scoped CI variable or an operator file with mode 600 | staging qualification only | refused in any other profile | project maintainers | the `identity-staging` job only |
| CI material: job-private DinD client certificates, the GitLab-issued job token | hosted jobs | GitLab runner | per job | not applicable | GitLab | job-private; remote URLs carry no embedded authentication; no registry login, deployment or project variable in `fast` or `gate` |

**Runtime authentication material** (generated per use from `crypto/rand`, 256 bits; `core/internal/identity`):

| Stored field | What it is | Stored form | Lifetime |
|---|---|---|---|
| `app.sessions.credential_hash` | session bearer (cookie value) | SHA-256 only | session (idle 30 min, absolute 8 h by default) |
| `app.sessions.csrf_token` | CSRF value returned to the browser | recoverable (32 raw bytes) | the session |
| `app.sessions.provider_logout_ref_hash` | provider-logout return reference | SHA-256 only | 10 min |
| `app.login_transactions.state_hash`, `binding_hash` | OAuth state and browser login binding | SHA-256 only | at most 10 min |
| `app.login_transactions.pkce_verifier` | PKCE verifier used at code exchange | recoverable (text) | at most 10 min; expired rows removed at a later sign-in |
| `app.sessions.provider_session_id`, `app.identity_principals.subject` | WorkOS session and user identifiers | plain (identifiers, not proofs) | session or account |
| `app.signup_sessions.credential_hash` (platform root signup) | root signup credential (`__Host-signup` cookie value; never a session) | SHA-256 only | at most 30 min |
| `app.signup_sessions.csrf_token` (platform root signup) | CSRF value of the signup credential | recoverable (32 raw bytes) | the signup credential |
| `app.signup_sessions.ref_hash`, `challenge_hash`, `nonce_hash`, `proof_hash` (platform root signup) | one-use continuation reference R, tenant challenge C (`__Host-sct` cookie value), root nonce N and proof P | SHA-256 only | one generation, 120 s; every re-mint invalidates all four |

Provider access tokens stay in memory only. A database dump or backup therefore holds recoverable PKCE verifiers and
CSRF values for unexpired rows, plus hashes, provider identifiers, names and audit rows: it is private operational
material (section 9).

**Public identifiers (not secret):** `IDENTITY_WORKOS_CLIENT_ID`, `IDENTITY_WORKOS_ISSUER`, `OBJECT_STORE_ACCESS_KEY`,
`MINIO_ROOT_USER`, `DB_USER` and the role names, the staging client ID and issuer, the emulator defaults.
**Disabled versus missing:** intelligence absent (`INTELLIGENCE_URL` unset) and the gateway absent
(`OPENROUTER_API_KEY` unset) are documented disabled states; every other row is required and fails closed.

### 3.1 Secret resolution in core

`config.Load` asks the `SecretProvider` once per name at startup, through one function (`resolveSecret`), and reports
every problem as the variable name with one fixed reason: `is not set`, `is empty` (whitespace only), `is invalid`
(NUL, CR or LF), `is denied by the secret provider`, `is ambiguous`, or `is unavailable from the secret provider`
(any other error or a missing provider). The provider's own error text and type are discarded; there is no fallback to
the environment, a default or an earlier value; a valid value is used byte-exact. Any problem returns no
configuration, so no client is built from a partial credential set (`main.go` loads configuration before any pool,
S3 client or health check). A provider implementation wraps `ErrSecretNotSet`, `ErrSecretDenied`,
`ErrSecretUnavailable` or `ErrSecretAmbiguous`; a replacement provider is substituted without touching callers
(unit tests U-SEC-1 to U-SEC-6). The WorkOS key keeps its value or mounted-file contract outside the provider.

## 4. Dependency vulnerability policy

`infra/security/dependency-scan.sh` (B-DEPSCAN, `lint-base`, so it runs in the merge-request fast job):

- **Tools:** `osv-scanner` v2.6.0 (pinned image digest) over `core/go.mod`, `web/package-lock.json` and
  `intelligence/uv.lock`; `govulncheck` (`golang.org/x/vuln` v1.8.0, checksum verified) over the core module with
  tests, at symbol level, including the Go standard library.
- **Data:** download-only. Four fixed public objects (the OSV Go, npm and PyPI exports and the Go vulnerability
  database bulk copy) are fetched by plain HTTPS GET into an owned volume, checked for HTTP status, size cap,
  Content-Length, published MD5, zip integrity, `Last-Modified` within 72 hours and a Go database `modified` within 30
  days, then scanned with network none. No package name, version, dependency graph, source or report leaves the
  machine. Go modules for govulncheck come from the Go module proxy and checksum database, as in every core build.
- **Coverage:** identities `(ecosystem, normalized name, resolved version)` are reconciled between each lockfile and
  the scanner result; project roots are excluded, repeated locations collapse, only the Go standard library at the
  `go` directive may appear on the scanner side alone, and unsupported records (links, entries without versions, Git
  or URL sources, `replace` directives) fail. Any missing or unexplained identity is an incomplete result even when
  the totals match. Not covered: OS packages and container images, the Node.js and Python runtimes, the MinIO source
  build, the validation-tools and skill graphs, browser and test images (follow-up scope, F-8).
- **Results:** a scanner error, timeout, truncated output, stale or failed advisory download, contract mismatch or
  identity mismatch fails the check; it never passes.

### 4.1 Blocking rules

- Severity class from the scanner's CVSS base score: 9.0 and above Critical, 7.0 to 8.9 High, 4.0 to 6.9 Medium,
  0.1 to 3.9 Low; otherwise the advisory's severity label; otherwise **Unclassified**.
- **Blocking:** Critical, High, any Go finding govulncheck reports as called whatever its severity, and any
  Unclassified finding (including imported-only and module-only Go findings without a severity). Go findings are
  merged across both tools by advisory ID and alias.
- **Reported, not blocking:** Low and Medium findings that are not called Go findings.
- Reachability informs disposition only; the inventory is never reduced.

### 4.2 Baseline

The initial baseline found 13 Go standard-library advisories in go1.27.1 (nine called) and one npm advisory. The Go
toolchain moved to 1.27.2, which fixes all thirteen; no Go exception exists. GO-2026-6809 (`crypto/x509`
`Certificate.Verify`, affecting Go 1.27.2) was published after that baseline; the current toolchain is Go 1.27.3, which
fixes it, and no exception was added. The npm advisory's exception has been retired (section 4.3).

### 4.3 Exceptions

`infra/security/dependency-exceptions.tsv`: one row per exact `(ecosystem, package, version, advisory or alias)` with
rationale, named owner, expiry date, review condition, compensating control and a review record. Wildcards,
ranges, group-wide (development or test) exemptions, missing fields, duplicates, expired rows (stale, compared with the
scan date in UTC) and rows that match no current finding (unused) all fail the check. A changed row is a security
disposition that needs maintainer review; risk acceptance is a maintainer decision. An npm or PyPI exception applies
only while the identity is development-only in its lockfile: if the package reaches the runtime graph, the check fails
(production inclusion revokes the exception automatically). An excepted finding stays visible in the report as
excepted, never as fixed.

No exception is active. The former row (npm `braces` 3.0.3, GHSA-vfj7-8cjw-p6xm / CVE-2026-93687, High 8.7) was
removed once the scan no longer reported a current finding for it, which the unused-row rule fails; the advisory record
and the locked `braces` 3.0.3 are unchanged, so a finding that returns needs a new exception row and a new maintainer
decision. Its recorded evidence: the package is absent from the production web runtime image (Next standalone output: no
`braces`, `micromatch` or `fast-glob` package anywhere in the image, and none in the base image's global npm); its only
path is `eslint-config-next` to `@next/eslint-plugin-next` to `fast-glob` to `micromatch`; the only loader is the
`no-html-link-for-pages` lint rule, whose only glob input is `settings.next.rootDir`, which the repository-controlled
ESLint configuration does not set. Such an exception is reviewed and revoked at the first of: a fixed release, a package
or version change, inclusion in a production artifact, or a change exposing glob pattern input. There is no automatic
renewal.

## 5. Least privilege

| Credential | Minimum use | Accepted grants and proof | Must never hold |
|---|---|---|---|
| `<db>_runtime` | application DML under FORCE row-level security; audit USAGE, filtered SELECT, INSERT of 12 columns; from version 14 column-limited INSERT only on `groups`, `role_bindings`, `business_profiles` and `group_business_profiles`, admitted by four INSERT policies (a binding only from inside the Group bootstrap trigger), with EXECUTE on `app.rls_profile_eligible` and still no DELETE on any of them; from version 19 exactly two more column privileges, `INSERT (display_name)` and `UPDATE (display_name)` on `app.mcp_connections`, under the unchanged connection policies and guard | DATABASE-ISOLATION, AUDIT, TENANCY; B-RLS-*, B-AUDIT-CATALOG, B-ONBOARDING-CATALOG, B-MCP-CATALOG | owner, migrator, SUPERUSER, BYPASSRLS, TEMP, CREATE, TRIGGER, role membership |
| `<db>_migrator` | SET ROLE `<db>_owner` to migrate | B-MIGRATE-IT, B-MIGRATE-ENV | runtime use; presence in core's environment |
| `<db>_retention` | the operator's bounded retention (migration 00009): USAGE on `audit`; SELECT of `audit.events` (`id`, `occurred_at`, `organization_id`) and `app.stored_objects` (`id`, `organization_id`, `state`, `deleted_at`); DELETE on both, limited by four retention-only policies to rows its own transaction declares eligible; EXECUTE on `app.retention_eligible` | AUDIT, RECOVERY; B-RETENTION-CATALOG, B-RETENTION-IT | INSERT, UPDATE, TRUNCATE, REFERENCES, TRIGGER, DDL, other tables, role membership, BYPASSRLS; presence in core, web or intelligence |
| `<db>_owner` and the helper roles | NOLOGIN | B-RLS-CATALOG, provisioning checks | a password |
| MinIO scoped user | bucket-only list, get, put and delete | OBJECT-STORAGE; B-PERSIST | root credentials in core |
| WorkOS API key | environment-wide WorkOS API (code exchange, session revocation) | IDENTITY | presence in web or intelligence |
| Intelligence service value | shared bearer for `/internal/*` | intelligence tests | browser or web exposure |
| Gateway value | OpenRouter calls when enabled | none (disabled) | anything but intelligence |

**Retention-login availability residual (inherent to the grants; a documented, accepted residual).** PostgreSQL grants
every LOCK TABLE mode to a DELETE holder and row-level security does not limit table locks; no separately revocable
LOCK privilege exists. A holder of the retention value can keep, for example, ACCESS EXCLUSIVE on `audit.events` or
`app.stored_objects` in an idle transaction and so block, across all Organizations, every audited write and every
object read, upload or delete until that session ends. Confidentiality and integrity stay bounded (row policies and the
append-only refusal). Statement and lock timeouts of the retention command are not a guarantee against a compromised
credential. The value is operator-only and never in a runtime service. Detection: blocking-session inspection
(`pg_locks` with `pg_stat_activity`; a lock granted to `<db>_retention` or an idle-in-transaction retention backend
outside a running retention command); recovery: terminate the backend and rotate or withdraw the value (section 6).
Details: [RECOVERY.md](foundation/RECOVERY.md).

Browsers and web bundles never acquire datastore, WorkOS or gateway values: web's container environment holds none
(B-COMPOSE), no `NEXT_PUBLIC_*` variable exists, credentials are never stored in browser storage
(B-IDENTITY-GUARDS), and the web build runs without `.env`.

**Tenancy mutations over HTTP.** Group creation and onboarding completion are the first tenancy mutations reachable
from a browser: web's BFF calls core's internal `/internal/v1/groups*` and `/internal/v1/business-profiles*` routes (not
public API) with the browser session, an exact Origin and the CSRF token. Every route calls one guarded facade
operation; bodies are strict with whole-payload budgets (8 KiB, 96 KiB for the two profile routes) checked before any
decode; errors are closed codes that never name or identify foreign objects; the creator is always the session's
verified user (no actor parameter, no provider claim). The version-14 INSERT policies and the SECURITY DEFINER
eligibility reader repeat each decision at the database, so the runtime credential with a fixed context cannot claim an
existing Group, insert a binding directly, attach an ineligible profile or delete an association
([AUTHORIZATION.md](foundation/AUTHORIZATION.md#guarded-group-onboarding-operations)). The database-isolation trust
limit is unchanged: a holder of the runtime credential can still set any context
([DATABASE-ISOLATION.md](foundation/DATABASE-ISOLATION.md)).

## 6. Rotation and withdrawal

Values are read only at startup, so a rotated value takes effect for a consumer only after its restart.

**Database login roles (`<db>_runtime`, `<db>_migrator`, `<db>_retention`):**
1. Generate the replacement in the deployment's secret store, never on a command line.
2. As the provisioning administrator, run `ALTER ROLE <role> PASSWORD …` through psql with the value read by `\getenv`,
   or psql's interactive `\password`, which sends a hashed verifier. Server statement logging can capture a plaintext
   `ALTER ROLE`; keep it off for this statement. Provisioning never changes an existing password.
3. New logins with the old value fail at once. Already-authenticated connections keep working: PostgreSQL does not
   re-authenticate open connections.
4. Update the consumer's source (`DB_PASSWORD` for core, `DB_MIGRATOR_PASSWORD` for migrate) and restart it.
5. Terminate the remaining pre-rotation backends of the role (`pg_terminate_backend`, filtered by role and
   `backend_start` before the rotation) and verify none remain.
6. Withdrawal without replacement: `ALTER ROLE … PASSWORD NULL`, then step 5; the consumer then fails closed (core
   readiness 503). Removing a source value alone revokes neither loaded values nor open connections.

**Exercised rehearsal (B-SECRET-ROTATION, `ops` suite):** on the disposable ops copy with synthetic values: the
pre-rotation connection still serves readiness after `ALTER ROLE`; the old value fails password authentication and the
new value logs in; terminating the pre-rotation backends turns core's readiness to 503; updating the copy's source and
recreating core restores 200 on new connections; `PASSWORD NULL` and termination make it 503 again. Durations and
backend counts are in each run's report.

**Other credentials:** Redis (`REDIS_PASSWORD`: change the server's password and core's source together, restart
both; open connections end with the server restart); object store (create a new scoped key, update core's source,
restart core, delete the old key); WorkOS (create a new key in the dashboard, update the value or mounted file,
restart core, revoke the old key); intelligence service value (update both services' sources, restart both, since
both read it at startup); gateway value (rotate at the provider, update intelligence, restart). The CI job token is
issued and expired by GitLab per job; the staging identity variable is rotated by a project maintainer in the
protected variable settings.

### 6.1 Human and tenant deprovisioning (existing paths only)

| Step | Existing path | Durable audit record | Notes |
|---|---|---|---|
| End a session or all sessions | user logout, logout-all (web to core `/internal/v1/auth/logout`) | `session.revoked` `logout` or `logout_all`, one per unrevoked session (expired but unrevoked included) | global rows: privileged inspection only |
| Provider sign-out | provider logout | `session.revoked` `provider_logout`; `session.provider_revocation_recorded` | revokes that WorkOS session only; remote deprovisioning is not claimed |
| Disable a user | `identity.Service.DisableUser` (trusted method; no route or CLI in the server today, F-6) | `user.disabled` and `session.revoked` `disabled`; a later sign-in: `auth.sign_in` denied `user_disabled` | `EnableUser` records `user.enabled`; old sessions never revive |
| Remove or downgrade tenant access | guarded `RemoveMember`/`ChangeMemberRole`/`DemoteOwner` through the members pages (owner rules enforced by the version-15 authority guard; see the evidence caveat in [DATABASE-ISOLATION.md](foundation/DATABASE-ISOLATION.md)); tenancy trusted primitives remain privileged-pool fixtures only | `membership.removed` or `membership.role_changed` (Organization rows, owner-readable) | current-authority checks deny the next guarded operation |
| Provider-side account removal | WorkOS dashboard (operator) | none in `audit.events` | blocks new authentication only; still run local revocation |
| Audit-only fault during revocation | the revocation-priority exception ([AUDIT.md](foundation/AUDIT.md)) | `revocation.unaudited` and one `audit.gap` diagnostic | revocation commits; check `audit.gap`; never a reason to undo logout or disable |

Privileged or bootstrap SQL is recorded with actor `unknown` and is not a substitute procedure. An authorized operator
entry point for `DisableUser` and membership removal is a production prerequisite (F-6). Service-credential rotation
is not an audit catalog event; its evidence is the operator record and CI artifacts.

### 6.2 MCP grant boundary

Implemented as described in [MCP.md](foundation/MCP.md). Core verifies Connect access tokens for one exact connection
resource and maps the subject to the existing principal; it never receives refresh values. Every connection is bound
to one user, one Organization (from the request host's current resolution, [DOMAINS.md](foundation/DOMAINS.md)) and,
after activation, one client; host, audience, client and user mismatches are refused (401) and run nothing. Each call
reauthorizes current membership and the selected Group. Logout, web revoke, sign-out everywhere, disable and expiry are
terminal locally. Compromised-client containment is the operator runbook in [OPERATIONS.md](foundation/OPERATIONS.md)
("Compromised MCP client containment"): `DisableUser` as the replacement hold, the qualified provider revocation, and a
release only under the prior-state, explicit-clearance and other-hold conditions. The test trust anchor
(`MCP_CONNECT_JWKS_FILE`) is refused outside `APP_ENV=test` (B-IDENTITY-GUARDS, B-MCP-E2E). Evidence covers only the
tested client, user and configuration: no claim about arbitrary clients, no revocation SLA and no immediate JWT
invalidation; issued access tokens stay cryptographically valid until expiry and are refused locally once their
connection is terminal.

**Credential containment limit.** With WorkOS MCP issuance, a user's client refresh credential may reach other
connections that user already authorized for the same client: local tenant-bound state is not independent credential
containment. Exact resource, user, client and host checks and live Organization/Group authority remain mandatory on
every call. Single-connection logout ends that local connection, not the provider grant; unknown or terminal records
cannot be revived by tokens. No broader provider behavior is assumed tested.

## 7. Transport and at-rest readiness

| Path | Current mechanism | Local exception | Supported production configuration | Verification | Evaluator status today |
|---|---|---|---|---|---|
| browser to web | local HTTP on loopback; TLS only in the identity test proxy | `loopback-http` cookie profile (refused in production by core and web) | external TLS termination with the `https` profile and origins | external | external-declared with evidence, else blocked-unknown |
| web to core | HTTP `CORE_INTERNAL_URL`; core serves HTTP only; the scheme is unchecked (F-2) | `http://core:8180` | an `https` URL to an external TLS terminator | client verifies; server side external | `http`: blocked-plaintext |
| core to PostgreSQL | pgx defaults: opportunistic TLS without verification, plaintext fallback, ambient `PG*` input (F-1); local server without TLS | Compose `data` network | none supported (verified TLS needs a scope decision) | none | blocked-unsupported |
| migrate to PostgreSQL | explicit plaintext (no TLS configuration, no fallbacks) | as above | none supported | none | blocked-unsupported |
| core to Redis | readiness AUTH and PING over plain TCP; no other Redis use | `data` network | none supported | none | blocked-unsupported |
| core to object store | S3 client; `https` required with `APP_ENV=production` | `http://minio:9000` | `https` endpoint, Go default verification | verified | config-verified when `https` |
| signed downloads | `OBJECT_STORE_SIGNING_ENDPOINT` rules (https in production; http data with https signing refused) | defaults to the data endpoint | `https` | the browser verifies | config-verified when `https` |
| core to DNS (custom-domain ownership proof) | system resolver configuration only, unauthenticated DNS (no DNSSEC validation); TXT via Go's native resolver; one typed CNAME query per edge to the first three `/etc/resolv.conf` nameservers (port 53, Go's loopback defaults when none; UDP with TCP on truncation, TCP only under `use-vc`; other options not interpreted), the queried owner's unique CNAME as the edge, ambiguous or malformed answers transient; every declared record framed and parsed, exact RDLENGTH enforced only for answer CNAME evidence (other RR types keep the parser's semantic-length limits and cannot supply the edge; bytes after the counted sections unobservable); a CNAME chain of at most three hops, 5 s deadline per lookup, special-use and platform names refused before lookup | `CUSTOM_DOMAIN_DNS_FIXTURE_FILE` (`APP_ENV=test` only) | as is; a DNS answer is ownership evidence, never authority | none (DNS) | outside the evaluator |
| core to the customer edge (custom-domain TLS readiness) | `edge-probe` dial of `DOMAIN_TLS_EDGE_ADDR` with SNI = hostname, default roots, chain/SAN/expiry checked | `DOMAIN_TLS_FIXTURE_FILE` (`APP_ENV=test` only) | as is | verified | outside the evaluator |
| edge to core (TLS `ask`, MCP) | HTTP on the internal network; the edge passes `Host` unchanged and no forwarding header is trusted | identity overlay nginx | an internal network reachable only from the edge and web | external | outside the evaluator |
| core to WorkOS | `https` API base (override only with `APP_ENV=test`), `https` issuer and JWKS, default roots | test overrides | defaults | verified | config-verified |
| core to intelligence | unauthenticated readiness probe over `INTELLIGENCE_URL` (unvalidated, F-3); intelligence serves HTTP | `http://intelligence:8000` | an `https` URL to an external terminator | client verifies; server side external | disabled when unset; `http`: blocked-plaintext |
| intelligence to gateway | fixed `https://openrouter.ai/api/v1`, default verification | none | as is | verified | disabled unless enabled; config-verified |
| PostgreSQL at rest | local Docker volume, no encryption setting | local fixture | provider or disk encryption | external | blocked-unknown without evidence |
| object storage at rest | local MinIO volume; an S3-compatible provider (for example R2) is a deployment choice | local fixture | provider-managed encryption | external | blocked-unknown without evidence |
| backups at rest | operator-private sets (`recovery.sh backup`: mode 0700/0600, local disk, no encryption) | local fixture | encrypted, access-controlled, immutable storage (deployment decision) | external | blocked-unknown without evidence |
| Redis at rest | no persistence | none | keep ephemeral | not applicable | not-applicable while ephemeral |
| CI daemon transport | job-private mutual TLS | none | as is | verified | outside the product result |

**Current production result: blocked.** Named prerequisites: verified TLS from core and migrate to PostgreSQL and from
core to Redis (or removing Redis), TLS on the internal hops (web to core, core to intelligence) and at the browser
edge, at-rest evidence for PostgreSQL, object storage and backup sets (the recovery tooling implements the local
procedure only) and an operator deprovisioning entry point (F-6). Local Compose is a development fixture, not evidence;
nothing here states that a local volume or link is encrypted.

**Staging profile constraints.** The included staging tooling ([STAGING.md](foundation/STAGING.md)) may use the
development loader with private local stores, real public HTTPS and explicit HTTPS cookies on web and core. It may not
use public test authentication, emulators, DNS/TLS fixtures or test signers. An identity-unconfigured staging
environment proves only infrastructure; real WorkOS configuration and privacy qualification are separate and have not
been executed for this repository. This does not clear the production result above, enable external MinIO signing
links or relax production-mode refusals. Injected container secrets can remain in Docker metadata on the retained
encrypted private disk after tmpfs delivery files are removed; disclose and protect that retention and its rotation and
removal lifecycle. These are design constraints, not observed deployment proof; the staging tooling has not been
exercised against any provider for this repository ([DEPLOYMENT.md](DEPLOYMENT.md)).

**Evaluator.** `core/cmd/readiness` reads non-secret `KEY=VALUE` lines on stdin: `APP_ENV`, `IDENTITY_COOKIE_PROFILE`,
`IDENTITY_PUBLIC_ORIGINS`, `IDENTITY_WORKOS_API_BASE_URL`, `IDENTITY_WORKOS_ISSUER`, `CORE_INTERNAL_URL`,
`OBJECT_STORE_ENDPOINT`, `OBJECT_STORE_SIGNING_ENDPOINT`, `INTELLIGENCE_URL`, `READINESS_GATEWAY_ENABLED`, any `PG*`
setting, and `READINESS_EVIDENCE_<ITEM>` references (`EDGE_TLS`, `CORE_TLS`, `INTELLIGENCE_TLS`, `POSTGRES_TLS`,
`MIGRATE_TLS`, `REDIS_TLS`, `POSTGRES_AT_REST`, `OBJECT_AT_REST`, `BACKUP_AT_REST`, and the in-platform items). Run it in
the network-free core check image:

```sh
docker compose --env-file .env -f infra/compose/local.yml --profile checks build core-check
docker compose --env-file .env -f infra/compose/local.yml --profile checks run --rm -T core-check sh -c 'go build -o /tmp/readiness ./cmd/readiness && exec /tmp/readiness' <readiness.env
```

It prints a JSON report per path and exits 0 (configuration-ready), 3 (blocked) or 2 (invalid input). Secret-named
input is refused before evaluation, as are unknown or duplicate keys. An evidence value must be a reference (a document,
ticket or record identifier); a boolean such as `true` is not one. Any `PG*` setting blocks the core PostgreSQL path
(`disable`, `allow` or `prefer` as plaintext, `require` as a verification bypass, anything else as unsupported).
`CurrentCapabilities` pins PostgreSQL, migrate and Redis as unsupported, so no evidence can lift them; changing that is
a reviewed code change. The best result is "configuration-ready", always labelled "configuration-only; external items
declared, not verified; not a deployed-platform attestation". Core still starts in production with plaintext stores;
the evaluator is the readiness gate for a later deployment review, not a startup refusal.

## 8. Findings recorded without a code change

| ID | Finding | Disposition |
|---|---|---|
| F-1 | core's pool keeps pgx's implicit TLS (unverified) with plaintext fallback and ambient `PG*` input; migrate pins plaintext | readiness block; follow-up for verified TLS under a reviewed scope |
| F-2 | web's `CORE_INTERNAL_URL` scheme is unchecked; production `http` is not refused | readiness block; follow-up |
| F-3 | `INTELLIGENCE_URL` is unvalidated; core loads the intelligence value but sends it nowhere | inventory note; the AI execution stage owns its real use |
| F-4 | minio-init passes the scoped user's secret in `mc` argv (local fixture, rendered file) | local exception; follow-up |
| F-5 | cookie-profile defaults differ outside production (core `https`, web `loopback-http`) | note; production behavior agrees |
| F-6 | `DisableUser` and membership removal have no operator entry point | production prerequisite |
| F-7 | `<db>_retention` had LOGIN and a password but no consumer or privilege before migration 00009 | closed: migration 00009 and `recovery.sh retention` are its reviewed consumer (section 5) |
| F-8 | base images are mostly tag-pinned; the graphs in section 4 marked not covered are unscanned | follow-up (image and supply-chain scope) |

## 9. Backup material and responsibilities

- **Private backup material:** a PostgreSQL dump or backup holds recoverable `pkce_verifier` and `csrf_token` values
  for unexpired rows, SHA-256 session, state, binding and logout-reference values, provider session and user
  identifiers, names and audit rows. It stays private operational material: owned storage, restricted access, never a
  CI artifact or committed fixture, encrypted at rest in production. A restore must not revive sessions or login
  transactions; logical expiry never proves physical removal.
- **Recovery tooling** ([RECOVERY.md](foundation/RECOVERY.md)): the paired backup with its writer fence and recorded
  trusted administrator window, the isolated fail-closed restore, restore-time invalidation of sessions, login
  transactions and logout references, the restore-input ledger and replay of later restrictions, bounded retention by
  the retention login (F-7 closed), backup expiry and quarantine, and the status hook. Production backup location,
  encryption, immutability, cadence, RPO/RTO and retention periods stay deployment prerequisites for the adopter.
- **MCP sessions and restore:** every MCP connection owns a session row, so the restore invalidation ends it; no MCP
  replay kind or recovery SQL change is needed. Re-enabling a user after a restore remains subject to the containment
  runbook's prior-state, explicit-clearance and other-hold checks; already-disabled or otherwise held users stay
  disabled.
- **Local private files:** Git and the Docker build contexts exclude the `.local/` directory; keep operator-private
  evidence and secrets there (directories mode 0700, secret files 0600) and select only sanitized evidence for any
  artifact upload.
- **Shipped versus adopter-owned:** the repository ships the controls above. Legal review, processor agreements,
  deployment hardening, production key management, privacy notices and subject-rights handling for an assembled
  product belong to the adopter ([PRIVACY-CHECKLIST.md](PRIVACY-CHECKLIST.md)).

## 10. Implemented tenant, domain and identity boundaries

These summaries point to the guides that hold the full mechanism and evidence.

**Verified domain routing and tenant-host authentication** ([DOMAINS.md](foundation/DOMAINS.md),
[IDENTITY.md](foundation/IDENTITY.md#tenant-hosts-and-the-callback-relay)). Routing is gated by ownership proof,
freshness and TLS readiness; sign-in on a tenant host requires local tenant membership; the shared WorkOS callback relay
is consumed only with the initiating host's browser binding, one-use state and PKCE, issues no session or cookie,
answers no-store/no-referrer and keeps entire queries out of logs. The durable resolver replaces the MCP host registry;
the TLS `ask` route answers 200 only for a fresh proof; the relay reads only the `state` origin tag and forwards the raw
query; hostnames are logged only as a 12-hex-character hash prefix, and challenge values, `state`, `code` and cookies
never. From migration 00012, a tenant-host browser session is bound to its issuing exact origin and canonical
Organization and resolves nowhere else, because a tenant controlling a custom hostname could terminate HTTPS elsewhere
and receive that host's cookies; exclusive hostname ownership is acquired atomically at valid proof, so unverified
pending claims cannot reserve another party's hostname. No new credential or secret is introduced (every domain
variable is public). Local evidence uses DNS/TLS fixtures and the WorkOS emulator; live DNS, real certificates, WorkOS
wildcard behaviour and the production edge log configuration remain deployment prerequisites and have not been
executed for this repository. Later requests keep actual current authorization; there is no instant cancellation
guarantee.

**Organization enrollment** (migration 00016; [TENANCY.md](foundation/TENANCY.md), [AUDIT.md](foundation/AUDIT.md)).
New Organizations start `invite_only`; request-access admission needs a current Organization manager's approval and
open enrollment admits ordinary membership under the current policy, which only the Organization owner changes through a
guarded, audited write (`open` is refused while membership history is unresolved). Authentication alone does not admit
a tenant member: admission runs through the claim writer's definer bodies with mandatory audit, recorded restrictions
and revoked approvals refuse re-admission, and enrollment refusals are audited `access.denied` reasons. Pending
enrollment state is separate from ordinary sessions.

**Membership, ownership and invitations** (migration 00015; [TENANCY.md](foundation/TENANCY.md#membership-ownership-and-invitations),
[AUTHORIZATION.md](foundation/AUTHORIZATION.md#ownership-continuity)). A database authority guard keeps at least one
owner for every Organization and Group (multiple owners allowed, atomic transfer); ordinary admin authority cannot
grant ownership to itself; emergency security suspension and session revocation stay immediate and keep the ownership
binding without usable access. Invitations are scoped, expire within 30 days, are revocable and are consumed once
through the claim writer; delivery e-mail or token possession alone is not authority.

**Organization branding** ([WEB-SHELL.md](foundation/WEB-SHELL.md)). Branding values are untrusted input validated by
core (closed token names, `#rrggbb` grammar, label bounds, effective-contrast pairs) and re-checked by web before they
are rendered as CSS custom properties; labels are rendered as escaped text. Logo images are decoded and re-encoded by
core within fixed bounds and served from the tenant's own origin with `nosniff`, an exact content type and an immutable,
versioned URL (cached bytes may outlive a removal for up to 24 hours). Management is owner/admin only (core-authorized,
CSRF-bound, version-checked forms); public reads carry no credential and never reveal object keys, IDs or sizes.

**Explicit Group authority** (migration 00017; [AUTHORIZATION.md](foundation/AUTHORIZATION.md)). Group content access
requires an explicit Group role; Organization staff see Group metadata, and ordinary Group management by Organization
owners and admins needs the Organization owner's default-off override. The Organization shortcuts are removed in core,
runtime SQL/RLS (`00017`), storage and the MCP/HTTP surfaces; the override is owner-only in Go and by an owner-owned SQL
guard, default false and audited (`organization.policy_changed` reasons and `access.override_used`); it grants no
content or ownership powers and no self-join; self-grants are refused at issue, claim and role-change time; an
override-only issuer loses invitation authority when the override, its role or its account goes. A page that answers
404 is not proof of backend isolation; the database and service checks are.

**Platform root signup** (migration 00018; [IDENTITY.md](foundation/IDENTITY.md#platform-root-signup)). Root signup
authenticates a global identity without granting an existing tenant session. The signup credential is a separate
30-minute bearer (`Authorization: Signup`) refused by every session consumer; creation runs only in the provision
writer's SECURITY DEFINER body (runtime and owner keep no Organization or `org_owner` INSERT path); the continuation into
the new tenant is bound by one-use 256-bit secrets stored as SHA-256 only (credential inventory above); a GET never
re-mints (recovery is a user-triggered POST); the form restoration cookie `__Host-signup-form` holds only the submitted
name and slug and is never authentication; names, slugs and continuation secrets are never logged; signup credentials
are revoked with sessions and expired by the restore invalidation. Platform credentials are never copied into tenant
cookies and membership is never inferred from a hostname or login.

**Platform addresses** (migration 00021; [DOMAINS.md](foundation/DOMAINS.md#platform-addresses-slugs-and-renames)). A
tenant's platform host is a core-generated label, resolved directly to the Organization UUID; the friendly slug is
metadata and its host is the fixed 404 with no redirect, so it discloses nothing about the label. Migration 00021's
privilege delta is exactly two column grants (auth reader `SELECT (platform_label)`, provision writer
`INSERT (platform_label)`); no role can update a label, and the writer-only provisioning policy requires one. The label
is not an audit attribute and appears in logs only inside the hashed hostname tag. Down from version 21 refuses while a
live or re-mintable signup continuation references a label.

New labels are friendly generated addresses (`<adjective>-<noun>-<four digits>`, about 25 bits from the CSPRNG over two
curated 64-word lists). They are public routing identifiers, not secrets, credentials or authorization factors.
Security rests on the unchanged exact-host classification, authentication, membership and RLS checks, not on label
unguessability; a guessed or enumerated label discloses nothing beyond the 404/sign-in entry a tenant host already
presents. Earlier 26-character base32 labels (128 CSPRNG bits, issued before friendly labels) keep resolving
unchanged. Labels that are reserved under the active configuration, equal to
the Organization's slug or already taken are redrawn in core, and exhaustion is handled without retained state. Live
identity-provider qualification of the signup flow on such hosts has not been executed for this repository.

**Labels** (migration 00020; [TENANCY.md](foundation/TENANCY.md#labels),
[AUTHORIZATION.md](foundation/AUTHORIZATION.md#guarded-label-operations)). Labels are private administrative metadata,
never authority: no capability, policy, RLS predicate, MCP eligibility or web gate reads a label, and assigning, removing
or seeing one grants or revokes nothing. Organization-owned labels on Groups are visible only through current
Organization owner/admin authority (`label.organization.manage`); Group roles never read them and they cannot opt into
member visibility. Group-owned labels are Group content: only actual roles in the exact owning Group read them (managers
all, members only `target_members` definitions) and only its owners/admins write them; Organization authority, Group
selection, shared profiles and the Group-management override confer nothing. Names, target ids and payloads never enter
logs or audit rows. Production storage admits no Group-owned assignment; a future target type must meet the
registration contract in [TENANCY.md](foundation/TENANCY.md#labels) under its own reviewed change. Implementation:
migration `00020_labels.sql` (two forced tables, seven policies, five strict audit actions), the three capabilities in
the closed evaluator, the closed target registry, the internal label routes and the labels pages.

---

Copyright 2026 Vitruvian Technology, Corp.
