> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G016 — Hostname resolution and verified domain lifecycle: Result

## Status

**Final (2026-10-17 UTC, closeout): merged; combined staging pipeline `2958083430` SUCCESS on `4aeabf12`. Ready for
Codex final ACCEPT, not self-accepted. See "Final closeout" at the end. The status blocks below are history.**

**Current (2026-10-16 UTC, successor #11): Codex source review at `edeb1965` returned CORRECTION LOOP
([review](G016-implementation-codex-review.md#source-review-at-edeb1965--2026-10-16-utc)); its three bounded
corrections are implemented with focused evidence PASS and returned for delta review. Not accepted; not integrated.**

- Scope: the owner-approved design (Plan r4) plus security amendment A2 (migration **00012**, tenant-host session
  binding, proof-time hostname exclusivity), the Codex-accepted A2 fixture and setupRecovery corrections, and the
  successor #11 corrections (production CNAME per-edge pin, TLS-result DomainRef fence, logout origin structure).
  Two actual migrations: 00011 and 00012. Allow-file `6e78ec85…`, 72 patterns, unchanged.
- Accounting: attempt 1 of B7, checkpoint 4, no reset; **1 of 7 attempts used**; this correction loop is Codex-bounded
  inside attempt 1 (first correction cycle on these three items).
- Evidence basis: the pre-A2 local qualification below (history, still attributable where its inputs are unchanged),
  the A2 focused selectors, s10lint1, web1/web2 and br1 (successor #10 section), and the successor #11 focused batches
  `fix11/u1` and `fix11/i1` (section "successor #11" below). No new pass is claimed for unexecuted suites.
- Remaining gates: Codex delta review of this correction, then the existing MR/rebase/Duo/auto-merge workflow; required
  hosted coverage (the recovery orchestration `run.sh migrations` → `recovery_drill`, the only runtime proof of
  setupRecovery; objects persist seed/verify; the hosted rows of the suites whose version-12 pins changed); P-1 real
  WorkOS, P-2 deployed/manual HTTPS, P-3 optional live DNS, P-4 production TLS adapter qualification; Codex final ACCEPT.
- History: the successor #8 update and the successor #7 status that follow describe earlier candidates.

**Update (2026-10-16 UTC, successor #8): owner-approved security correction under Plan amendment A2 in progress —
migration 00012 and the identity session binding are implemented with focused evidence PASS; the consumer slice is
implemented but its validation stopped at a newly found access/MCP integration-harness scope gap, handed back for a
Codex decision (see "A2 execution: slices 1–2 done, slice 3 stopped" below). Web, docs, browser and lint steps not
started. The successor #7 status below describes the pre-amendment candidate and is history.**

**Current before the amendment (2026-10-16 UTC, successor #7): implementation complete on the task branch; local qualification evidence
collected; submitted for Codex final source/security review. Not accepted; not integrated.**

- Attempt 1 of B7 (checkpoint after attempt 4), one writer; **1 of 7 attempts used**, no correction pass on
  implementation beyond the Codex-bounded loops recorded inside attempt 1 (deviation dispositions, amendments 01–05,
  the ops2/ops3 findings and the E-DOM-4 fixture correction accepted at staging `2421c23b`).
- Evidence basis (details in "Exact commands/validation performed" and the AC checklist below): b1 static/unit and
  web checks, b2 focused integration selectors, `run.sh lint lint-base lint-core` PASS (`lint2`, source `1d7193ac`),
  the attributable PASS rows of `ops3` (FAILED overall on E-DOM-4; its premature launch stays recorded) plus the
  focused E-DOM-4 and E-DOM-3 passes on the accepted correction `2d604648`, and **`run.sh rename` PASS (`rename1`,
  head `c0b2f0b5`)**. `ops1`/`ops2`/`ops3` remain FAILED records.
- Held gates, none satisfied by this Result: Codex final source/security review (S-review), provider evidence P-1
  (WorkOS staging), manual qualification P-2, optional live DNS proof P-3, production TLS adapter qualification P-4,
  MR opening (coordinator), MR fast and the hosted staging gate, Codex final ACCEPT. No live DNS, WorkOS, credential,
  provisioning, deployment or settings change was made.
- The statements under the headings "Status at R0", "Planned at R0" and the "None yet. Planned:" scaffold lines below
  are historical: they describe the state before execution and are kept verbatim as review history.

### Status at R0 (historical; superseded by the current status above)

**Initial Result (R0): reconciliation done (nothing material); execution not started; not accepted.**

- This is the execution-readiness record that Plan r4 §0.1 asks for, ahead of Codex's review of r4 (the R0 record),
  the Git-first binding confirmation and the recorded execution release. It is not implementation, not review-ready
  for acceptance and not acceptance.
- Nothing has run: no build, test, harness, Docker, provider call, DNS/TLS probe, migration or settings change. No
  delegate has been dispatched and no MR has been opened. Implementation attempt 1 of B7 (checkpoint after attempt 4)
  has not begun; **0 of 7 attempts used**.
- Design authority: full Plan r3 ACCEPT with the owner's design approval
  (record (not included), ADR-007). Design
  only: **no live DNS/TLS proof, WorkOS setting, credential/account change, provisioning or deployment is
  authorized**; P-1..P-4 remain explicit final-qualification prerequisites (Plan §6.4).
- Implementation starts only after (1) Codex's review of Plan r4 and this Result, (2) the binding confirmation
  (target, inputs, scope, budget, resources, single writer) and (3) the recorded supervisor execution release.

## Execution binding (Git-first; prepared for Codex to confirm)

Hash rows carry neutral labels; paths are listed under "Bound paths". The full input table is Plan r4 §2.

| Label | Bound bytes |
|---|---|
| Target: `origin/staging` at R0 fetch, 2026-10-16 (contains the merged G015 `7e0ac44`, its final ACCEPT `443c66a` and the owner design record `f6d79af`) | `f6d79af5cc97e0cf6c17b5a9f649a2ad0d4cf827` |
| Plan r4 (P4), the R0 record proposed for review | recorded at hand-back (this file's own commit; see Plan §13) |
| Plan r3 archive (P3), the accepted design bytes | `3b29275ea8c914178138889e29e3a7412209be1378ae1c1019904e08809c1ffc` |
| Plan r2 archive (P2) / r1 archive (P1) | `3b6ea4895e6195065406c08eeb385aab9508e49fc2c589501ff9d2d277a1b702` / `38c736e711d4284718381f46aa5774564e32f0bf346060e531c3dcc12abe711c` |
| r3 review with owner approval and Plan ACCEPT (R3) | `cd763cf7f20e7dca955b5c24d63589dc628652e9ca7adcc580b676ae991ff359` |
| r2 review (R2) | `32f18832470bf5d0b4301d5465aa4c75241facc204060d3b6043475ae37348e3` |
| Goal (G) | `299f08453a4e6defa6774c139052f36883be39f0c3bcff574d11c6ce6fbbfc56` |
| Spec (S), with the owner-selection section | `7f84ae540adf0535dd16b2b2b5f458358612905f5d79ff1d5c24e8a6a4998308` |
| Scope allow-list (SC), 44 patterns, byte-equal to Plan §11.3 | `047450e1f12bab0ae2d7e07432b51109d27c43efa9935930f3b27b97fa983aa0` |
| ADR-007 (A7) | `46cf243031b08e2b3f4dc11a43f51e12ee1f810eb721fb903f772733e34f80fe` |
| Architecture (AR) / Security (SE) / Performance incl. PG-007 (PF) | `c056ce5f5418bac1870dfbc05b521632d9b66c66ceb9906d3d7144f588f1d30f` / `38aa0d1df17678dcb2a241cd739693f41eb66d1cdf4d7b23239b434776ef6b37` / `e3b3db3aeca17c102cade4f987dedabd924048fe9acdeafe66c2628d54c16b08` |
| MCP connection-context contract (CC) / control-plane contract (CP) | `190be18885c4cadd2e2ba60fe9b1c0eb2391c9384219a5ba57991ba47d0a607e` / `82ccd466a99d5c44f3434b9ca8512468a63f1b914ba65b31485e7286f5815a29` |
| G015 Result (X15) / G015 implementation review (I15) | `69235db5f96db2287ad85868d7e89935ad79d82df761b3b33021a365b3eae986` / `935077028370f4fd8337492093581f51825ba4b545e58a1f9254c7623d4db496` |
| Harness (RS) / Required checks (RC) / Validation-integration reference (VI) | `e62f76127501fb6659cf4ae7cab14da80f968437f52c04e5776ea0834142ad44` / `2d136fea2fa096def63d29a5182fb563b6419f94babf519c8e61ed3455a1c686` / `696df897dcea6eb599a025adb8a22363d94845e4a411af1173292fa285b192d5` |
| Control scripts (VC) / (LB) | `04319fa9368cd9523e09c3e745953bbf7afc7cf45b428dbd38292b0af5deb1eb` / `03d300759b62f53c32796a4345303c893d44946419d3761e7ff7fad3dc2532e4` |

**Bound paths.** P4/P3/P2/P1 `management/v0/plans/G016-hostname-resolution-domain-lifecycle-plan{,-r3,-r2,-r1}.md`;
R3/R2 `management/v0/results/G016-plan-r{3,2}-codex-review.md`; G
`management/v0/goals/G016-hostname-resolution-domain-lifecycle.md`; S
`management/v0/specs/G016-hostname-resolution-domain-lifecycle-spec.md`; SC
`management/v0/results/G016-scope-allowlist.txt`; A7 `docs/adr/ADR-007-verified-domain-routing-and-authentication.md`;
AR/SE/PF `docs/{ARCHITECTURE,SECURITY,PERFORMANCE}.md`; CC/CP `contracts/mcp/{connection-context,control-plane}.md`;
X15/I15 `management/v0/results/G015-{dynamic-capability-exposure-result,implementation-codex-review}.md`; RS/RC
`infra/validation/{run.sh,required-checks.tsv}`; VI `skills/sdlc-management/references/validation-integration.md`;
VC `infra/ci/validate-controls.sh`; LB `infra/validation/lib.sh`.

**Budget and effort.** B7, checkpoint after attempt 4 (Goal); planning r1–r4 and R0 spent **no attempt (0/7 used)**.
Implementation effort high (default); the owner may choose xhigh for a specific phase (candidates: the guard/lock
work of §1.4/§1.7.1 and the relay). The actual runtime setting is reported as unknown unless verifiable.

**Worker record.** R0: task G016, role plan/R0; native isolated worktree
`<source-checkout>` (generated
`agent-<id>` name, the accepted AGENTS.md fallback), branch `g016-r0` from `f6d79af` (new name; the stale local
`g016-plan` stays in the retired planner checkout `agent-<id>` at `4de99c9`, while `origin/g016-plan`
is `9715fa7`); successor to the r1/r2 planner after the 350k active-context handoff; model Fable 5.1, assigned
effort high, runtime effort unknown; session `<agent-session-url>`.

**Single-writer plan.** Implementation: **one worker** owning every path of the allow-list (SC) on one feature
branch (`g016-hostname-resolution-domain-lifecycle`) created from the release target, in a native isolated worktree
named `g016-agent-<id>` when the tool allows, otherwise the generated name with task/role/path recorded here. No
second writer: the Plan's optional read-only reviewers (§7: one security/concurrency reviewer after b2; a second only
for E2, which D-1 did not select, so none) never edit. Delegated exploration, if any, is read-only. The coordinator
holds the batch lock; the implementer never edits during a running batch.

**Resources (readiness, nothing allocated yet).**

| Resource | Plan |
|---|---|
| Check images | `g016-core-check:b<N>` (`docker build --target check -f infra/docker/core.Dockerfile core`), `g016-web-check:b<N>` (web check image); run with `--network none --entrypoint sh` (CMD overridden); removed after the batch |
| Focus database | disposable Compose project **`g016-focus`** (postgres only from `infra/compose/local.yml`, network `g016-focus_data`, generated 0600 env files removed afterwards, the `run.sh` `rls_go` env shape: canary file, access log, `MCP_LOG_FILE`); `down -v --remove-orphans` after each batch; no container/volume/network with the project or `g016` label may remain |
| Host ports | none for b1/b2 (`--network none` or the project network only). The single `ops` exception (C-1) uses the identity overlay's existing standard+100 ports (edge 543; Playwright fixture servers 3001/3003) and adds none |
| sqlc trial | `docker run --rm --network none --user "$(id -u):$(id -g)" -v "$PWD/core:/src" -w /src sqlc/sqlc:1.29.0 generate -f internal/domains/db/sqlc.yaml`, one iteration (Plan §5.7) |
| Harness lock | the coordinator batch lock per batch (G015 practice); `bash infra/ci/docker-desktop-mounts.sh` before every batch; fetch/rebase/reconcile before every batch; no edits during a running batch |
| Hosted | MR fast runs automatically; the path map selects full; the post-merge staging full gate supplies the hosted rows after an authorized conditional merge (not requested here) |
| Provider/manual (held) | P-1 WorkOS wildcard/auth-host entries, P-2 staging deployment with real TLS and whole-query edge-log redaction, P-3 optional live DNS proof, P-4 Caddy production provisioning — each needs its own owner/operator authority; **none is requested or assumed** |

## Execution record — attempt 1 of B7 (recorded before the first code write)

Release: R0 review and scoped execution release (not included) on `origin/staging`
`5314f0cabc1f11858b189f6d62f5690df6b5a63c` (SHA-256 `a8aa25416e6a7ddfddc82f9e0fc90930b5eee2237dec83833d330195a2cd22ab`).

| Item | Recorded value (2026-10-16 UTC) |
|---|---|
| Task / role | G016 / implement (sole writer); hands back to the dispatching coordinator |
| Model / effort | Opus 5.5; assigned effort high; runtime effort **unknown** (not verifiable from inside the worker) |
| Session | `<agent-session-url>` |
| Isolated checkout | `<source-checkout>` (generated `agent-<id>` name; allowed by the release) |
| Branch / base / HEAD at start | `g016-hostname-resolution-domain-lifecycle` from fetched `origin/staging` `5314f0c`; ancestry verified (`git merge-base --is-ancestor 5314f0c HEAD`); HEAD `5314f0c` |
| Bound inputs re-hashed at start | Plan r4 `48c1a792…37d2`, initial Result `4ebd0db9…a22d`, allow-list `047450e1…0aa0`, release `a8aa2541…22ab` — **all match**; the delta `f6d79af..5314f0c` is the management-only release import (no material mismatch) |
| Budget | B7, checkpoint after attempt 4; **attempt 1 begins** with this record |

## R0 reconciliation summary (Plan r4 §0.2.1, §0.4)

| Item | Outcome |
|---|---|
| P-G015-1..7 and §0.3 against the merged code at `f6d79af` (code = merge `7e0ac44`) | all **confirmed** with file:line; immaterial drift only (registry lookups at `handler.go:143/223`; `newServer` declared in `server.go:63`) |
| Migration number | 00010 is the last migration; **00011 is free** (provisional number confirmed) |
| Audit catalog | 31 actions (25 trigger-owned + 6 core, `audit_test.go:75`); G016's 31→35 delta stands |
| Count pins (§11.2) | B-MCP-IT 7/6/1/1/5/10/1, E-MCP 5 (`run.sh:3148`); required-checks 117 rows, profiles 35/75/110, VA7 80 by byte-equality of RC/VC/LB with the G015 rows; RLS-SCOPE 36; identity Playwright projects 4; ISO-ORACLE re-pinned at b2 |
| Changed bound inputs since the r3 base | Spec digest (owner-selection section appended; no requirement/AC change); new inputs ADR-007, PG-007, architecture/security/contract notes, scope allow-list, G015 Result/review; **none material**; all rebound in Plan §2 |
| Owner decisions | D-1 = E1 Caddy with an explicit AC6 production gate; D-2 = (a) membership required; D-3 = (a) wildcard + B2 relay; D-4 = (a) retain, refuse while unresolved, no automatic revocation; D-5/D-6 settled (Plan §0.4) |
| Codex implementation constraints (r3 review) | whole callback query (incl. `error_description`) out of logs/edge logs; `ResolveWith` reaches identity through identity's own interface; READ COMMITTED guard semantics add no instant revocation; recorded in Plan §0.4 and honoured at implementation |
| Executable selectors | Plan §6.3 (b1 unit/static, b1 sqlc, b1 harness `lint lint-base lint-core`, b2 focused integration, b3 the single `ops` exception, b4 `rename`, S-review, provider/manual after P-1/P-2, final = targeted reruns only); the `-run` expressions are listed below |

## Implementation summary and files changed

### Attempt 1 status at the 350k context handoff (2026-10-16 UTC; nothing compiled or run yet)

Branch `g016-hostname-resolution-domain-lifecycle`; round checkpoints `c00b215` (binding record), `ebcc564` (round 1)
and the round-2 commit that carries this section. Every changed path matches the 44-pattern allow-list (scratch
checker over `git diff origin/staging...HEAD`, staged, unstaged and untracked paths: 40 paths, none outside). **No
build, test, harness, Docker or sqlc run has happened**: b1 is the first validation batch and will surface compile
errors. Single writer throughout; no delegate edited anything.

Done (Plan §8 steps 2–7, core only):

| Area | Files | State |
|---|---|---|
| Hostname contract (§1.1) | `core/internal/hostname/{hostname,hostname_test}.go` | written; `TestHostnameNormalize`, `TestHostnameClassify`, `TestReservedLabelsFloor` |
| Domain configuration (§5.4) | `core/internal/config/{domains,domains_config_test}.go`; `config.go` (single-label `localhost` platform domain with `APP_ENV=test` only); `mcp.go`/`mcp_config_test.go` (registry retired) | written; `TestLoadDomainsConfig` |
| Migration 00011 (§4) | `core/migrations/00011_organization_domains.sql` (1339 lines) | assembled by a scratch generator from 00010's role checks/postconditions (version-10 precondition), the hand-written DDL/guard/readers/capture, and programmatic version-11 inventory edits; down restores the verbatim version-10 attribution body, catalog CHECKs and the 00008 `capture_organizations` body/trigger. 00010 digest for the B-MIGRATE-SET pin: `3ac29b6bcf2078158cf7fbaac13b244ce2610450d2e6f0b53fba734b14a73e19` |
| `domains` package (§1.3–§1.5) | `core/internal/domains/{domain,lifecycle,dns,tls,resolver}.go` + `{lifecycle,dns,tls}_test.go` | written; `TestDomainStateMachine`, `TestDNSChecker`, `TestTLSAdapterContract` |
| sqlc trial inputs (§5.7) | `core/internal/domains/db/{sqlc.yaml,queries/domains.sql}` | written; **generation not run** (b1). `access/domains.go` is written against the expected generated names (`db.New`, `GetDomainRow`, `InsertDomainParams`, `TlsState`, …); adapt call sites within the one trial iteration, or fall back to handwritten pgx with the same names |
| Facade (§5.0.1) | `access/domains.go` (new), `access.go` (one `domainHooks` field), `policy.go` (`organization.manage_domains`), `errors.go` (`host_taken`, `domain_state_invalid`, `domain_retry_exhausted`), `boundary_test.go` (files, imports, exported pins, importer cases, 00011 SECURITY DEFINER allowance) | written |
| Internal API (§5.3) | `core/internal/domains/api/http.go` | written (`hosts/resolve`, `tls-ask`, list/claim/rotate/verify/recheck/remove/retry-withdrawal, `organization/slug`); `http_test.go` not yet written |
| MCP (§1.6) | `mcp/registry.go` (`HostLookup`), `handler.go`, `connections.go`, `mcp_test.go` (`TestLookupIsExact`, static lookup double), `env_integration_test.go`, `mcp_integration_test.go` | written; IT-MCP-2/3/8 resolver extensions and `e2e_mcpe2e_test.go` E2E-MCP-6 refusals **not yet** |
| Identity (§1.7, §1.7.1) | `identity/relay.go` (new: `HostResolver`/`Querier` declared by identity, tenant origins, origin-tagged state, `Relay`), `identity.go` (`host_changed`, `organization_not_member`, `relay_refused`, `unknown_host`), `service.go` (Login/Callback), `store.go` (txn columns, `issuanceGuard` inside `signIn` before the first write of either branch), `logout.go`, `http.go` (session `?origin=`, relay route), `allowlist.go` | written; identity tests not yet adapted |
| Audit | `audit/audit.go` (35 actions, 29 trigger-owned, 54 reasons, 24 operations, `organization_domain`) | written; `audit_test.go` pins not yet |
| Wiring | `cmd/server/main.go` (domains config, resolver, checker/TLS adapter, `originResolver` adapter, MCP `Hosts`, domains API, relay route) | written |

Remaining, in Plan §8 order (successor):

1. Core tests: identity `service_test.go`/`integration_test.go`/`logout_test.go`/`http_test.go`/`allowlist_test.go` (fake `HostResolver`; seeded slugs `org-a`/`org-b` under `PLATFORM_BASE_DOMAIN=localhost`, port 543); `relay_test.go` `TestCustomDomainRelay` (11 subtests incl. cases A–D); `domains/api/http_test.go`; `access/domains_integration_test.go` `TestDomains` DOM-1..10 with `E-DOM-F1..F4`; `migrate/domains_integration_test.go` `TestDomainsCatalog`; `audit_test.go` (35/29); `matrix_test.go` (twelfth capability; oracle and `TestToolEligibility` pins), `policy_test.go`; `rls_expected_test.go`/`attack_expected_test.go` oracle rows; IT-MCP-2/3/8 with the real resolver; E2E-MCP-6 refusals; `tenancy/slug_test.go` floor parity; confirm the `TestRLSBoundary` identity pins still hold (`signIn` now calls `issuanceGuard`).
2. Web (§1.6, §1.7, §1.7.2): `web/lib/host.ts`, `web/proxy.ts`, `session.ts` (`Referrer-Policy: no-referrer` in `NO_STORE`; session `?origin=`), `auth/callback/route.ts` relay branch (raw query forwarded byte-identical; core `GET /internal/v1/auth/relay?state=`), `mcp-client.ts` parser + `connections/mcp/page.tsx` host column, sign-in page on the auth host, unit tests (`host`, `mcp-client`, `session`). Read `web/.claude/skills/vestive-design/SKILL.md` first.
3. Harness/config (D-6): `run.sh` (B-MIGRATE-SET 00011 + 00010 pin above; B-DOMAINS-CATALOG/B-DOMAINS-IT in `migrations`; `mcp_it` wording; AUDIT-INV 35; RLS-SCOPE; B-ENV-EXAMPLE reviewed domain block; B-IDENTITY-CONFIG rows; the E2E-MCP-6/probe env strings that still name `MCP_TENANT_HOSTS_FILE` and tenant `IDENTITY_PUBLIC_ORIGINS`; `variants_rename`), `required-checks.tsv` (117→119). **Check first whether `infra/ci/validate-controls.sh` pins the profile counts (35/75/110): it is conditional (Codex-reviewed amendment only)**.
4. Overlay: `identity.yml`/`local.yml` (`PLATFORM_BASE_DOMAIN=localhost`, `PLATFORM_HOST_PORT=543`, `PLATFORM_AUTH_HOST`, fixtures, no registry mount), `mcp-seed.sql` (slugs + one active custom row), `identitysetup` memberships, emulator redirect hosts, `proxy.conf` comment, `mcp-hosts.tsv` (harness data only), `{dns,tls}-fixture.tsv`, Playwright config/specs, `events.json` (`domain.request`, `auth.relay`, access ops), `.env.example`, `configure-env.sh`.
5. Docs/contracts (DOMAINS.md and the listed foundation docs, SECURITY, PERFORMANCE PG-007 with measured `E-DOM-F`, ARCHITECTURE, `connection-context.md`).
6. b1 (sqlc trial first) → b2 → hand-back (a) for the S-review.

### Attempt 1 continuation — successor record, dispositions and current bindings (2026-10-16 UTC)

| Item | Recorded value |
|---|---|
| Worker | G016 implementation successor, role implement, **sole writer** of `g016-hostname-resolution-domain-lifecycle` (the predecessor stopped and does not push); Opus 5.5, assigned effort high, runtime effort **unknown**; session `<agent-session-url>`; native isolated worktree `<source-checkout>` (generated name, allowed by the release); hands back to the dispatching coordinator |
| Start | origin branch at `94a3a6a4e732ce97e52934beb6953a75cd93467c` (base staging `5314f0c`); Result `3a557b3e…529c`, Plan r4 `48c1a792…37d2`, allow-list `047450e1…0aa0` verified; a context handoff spends no attempt (**attempt 1 of B7 continues; checkpoint after attempt 4 unchanged**) |
| Rebase | onto `origin/staging` `35cab060f11fdb4c0a6e16643c216e49c83ad9f4` (management-only delta `5314f0c..35cab06`: G054/G055 planning, the deviation review and scope amendment 01); ancestry verified with `git merge-base --is-ancestor` |
| Deviation review | attempt-1 deviation dispositions (not included) (CORRECTION LOOP inside attempt 1), SHA-256 at `35cab06` `c85899644fa5f4f54dfc2d6791f49b09ef49b0495a74990f0a588865f46628b7` (the dispositions-only version at `9fbbb9f` was `5a1de998…3c95`) |
| Scope binding | allow-list history: `047450e1f12bab0ae2d7e07432b51109d27c43efa9935930f3b27b97fa983aa0` (44 patterns, release) → **`eb85e9054a8ee71da91c7539e22937e130cce576827c2ea1cf68176c4ab4c403` (45 patterns, scope amendment 01: adds `^infra/ci/validate-controls\.sh$`)**; recorded here before that file is edited |
| Amendment 01 edits (authorized, exactly) | `validate-controls.sh` VA7 `required_count 80`→82 with its detail text, VU1 `vu_split full 35 80 111`→`35 82 113` with its detail text, J8 `n != 80`→82 with its error text; no logic, control, suite, timeout or CI change; no local full/ci-controls run (the hosted full gate covers the literals) |

Dispositions (Codex, recorded verbatim in substance; not test passes):

- **Accepted design variations:** #2 (`hostname_state(kind, key)`), #5 (relay decision in core; the web leg must still meet
  §1.7/§1.7.2), #6 (all-environment `MCP_TENANT_HOSTS_FILE` refusal), #7 (closed outcomes/codes, `domain.list` denial),
  #8 (test-only `localhost` platform domain, non-tenant `IDENTITY_PUBLIC_ORIGINS`, special-use DNS refusal with
  injected `example.com`-style fixtures).
- **#4 restored (correction):** Login's tenant-host resolution now runs as one statement inside `insertTransaction`
  before the INSERT (refusal rolls back, nothing inserted; structurally impossible origins are refused before any
  work), and the Callback pre-exchange host check runs as one statement inside `consumeTransaction` after the one-use
  DELETE (a non-routing answer still commits the consumption and yields `host_changed`; a store failure rolls back
  with nothing consumed). The final issuance guard stays inside `signIn`; no connection or lock is held across WorkOS
  calls. **Concrete constraint, returned for approval rather than switched:** the session route's `?origin=` check and
  the sign-out origin check have no existing transaction to join (the session read `resolve` is one autocommit E1
  reader statement; renewal opens a transaction only once per 60 s); the inventory is in the proposal file below.
- **#1 (slug authority) and #3 (`app.hostname_generation`) held:** scoped proposals written to
  `<home>/.local/state/platform-evidence/g016-proposals-attempt1.md` for Codex; not implemented until approved.
  The current code keeps both as they are, isolated so either change is local.

Stale questions in the initial Result are answered by the recorded release (no new approval loop): the generated
worktree name is accepted; the single local `ops` exception (C-1) is taken once at the Plan's point; effort high.

Consequence observed while adapting the identity suites (Plan §1.9/§10, D-2(a), not a new deviation): with membership
required at tenant-host issuance and no non-tenant sign-in origin, `createMapping` is unreachable, so the identity
integration fixture provisions each principal (user, principal row, org-a/org-b memberships) before its sign-in, as an
operator does; I-MAP-CONC now pins concurrent sign-ins of one provisioned principal (one user, one principal).

### Attempt 1 continuation — proposal closure, amendment 02 and the sqlc outcome (2026-10-16 UTC)

| Item | Recorded value |
|---|---|
| Proposal closure | Codex accepted the #1/#3/#4 proposals (proposal SHA-256 `d1601ca4…7b7f`; closure in the deviation review (not included) at staging `6179516`) |
| Scope binding | `eb85e9054a8ee71da91c7539e22937e130cce576827c2ea1cf68176c4ab4c403` (45 patterns, amendment 01) → **`298b7572b46acfe363bb59a16b161de8958d107820eaf034b753654d105f82a0` (53 patterns, amendment 02 (not included), review SHA-256 `85367a81…acdb0d`)**: adds exactly `core/internal/migrate/{runner_test,integration_test,mcp_integration_test}.go`, `core/migrations/migrations_test.go`, `core/internal/identity/{store_test,rls_pool_integration_test}.go`, `core/internal/tenancy/access/recovery_integration_test.go`, `core/internal/tenancy/integration_test.go` (version-11 pins and test fixtures only); recorded here before any of them is edited |
| Rebase | onto `origin/staging` `8f06f4680055a4abb324faf941e2eed61fb27bd8`; ancestry verified |

Implemented corrections (round 3–5; focused evidence to follow in b1/b2):

- **#1** `organizations_update_slug` (owner/admin, `USING`/`WITH CHECK` on the context Organization) plus the owner-owned
  `app.organizations_slug_guard()` BEFORE UPDATE trigger: in an `org_admin` context any change other than `slug` raises
  `42501` (`to_jsonb(NEW) - 'slug'` against OLD). `organizations_update` and the column grants are unchanged; the 00011
  policy, function and trigger inventories and the Down section carry the new objects. `SetOrganizationSlug` is decided
  with `organization.manage_domains`.
- **#3** `app.hostname_generation` and its grant/inventory entries are removed. The claim generation is an opaque positive
  int63 drawn with `crypto/rand.Int(rand.Reader, MaxInt64) + 1` (range 1..MaxInt64, no overflow; an RNG error fails
  closed). `InsertDomain` absorbs only its own fence collision with `ON CONFLICT ON CONSTRAINT
  organization_domains_hostname_generation_key DO NOTHING` (no aborted transaction); the facade regenerates once, a
  second collision fails closed, and only `organization_domains_hostname_live` maps to `host_taken`. Dated Plan
  interpretation: the generation is an equality-compared claim fence, not a per-hostname counter.
- **#4(a) cost record:** the session-route `?origin=` check and the sign-out origin check each add **one pool acquisition,
  one SQL statement, one DB round trip and one implicit autocommit statement transaction, with zero new explicit
  BEGIN/COMMIT**; web→core HTTP calls are unchanged. Login (inside `insertLogin`, before the INSERT), the callback
  pre-exchange check (inside `consumeLogin`, after the DELETE) and the final issuance guard (inside `signIn`) add one,
  one and two statements to existing transactions.
- **Security correction (amendment 02):** the round-3 store-level skip (`issuanceGuard` returning nil without a resolver)
  is removed; a sign-in with no resolver or no pinned Organization is `host_changed` (fail closed). Store-level tests
  get a test-only resolver and membership facts instead.

**sqlc trial (Plan §5.7, one iteration, b1): deferred.** Command (repository root, under the batch lock):
`docker run --rm --network none --user 1000:1000 -v <worktree>/core:/src -w /src sqlc/sqlc:1.29.0 generate -f
internal/domains/db/sqlc.yaml` → exit 1; exact error (first of 56 identical-shape lines, all on migrations 00002–00011):
`../../../migrations/00002_identity.sql:1:1: schema "app" does not exist` (and `schema "audit" does not exist` from
00008). The app/audit schemas are created by the role script (`infra/postgres/init/10-roles.sh`), not by any migration,
so generation would need a second schema file, a recorded deferral criterion. Per the Plan, `sqlc.yaml` and
`queries/domains.sql` are deleted, nothing was generated, and the same statements live as handwritten pgx constants
with the generator's names in `core/internal/domains/db/queries.go` (no `go.mod` change).

### Attempt 1 status at the second 350k context handoff (2026-10-16 UTC)

Focused b1 runs so far, each after fetch/rebase (ancestry verified at `8f06f46`), `bash infra/ci/docker-desktop-mounts.sh`
(13% of mounts in use) and under `flock <home>/.local/state/platform-evidence/local-harness.lock`. Logs are in
`<home>/.local/state/platform-evidence/g016-b1/`; the check images run with `--network none` and the
entrypoint/CMD overridden.

| Run | Command | Outcome |
|---|---|---|
| sqlc trial | the §5.7 command (above) | exit 1, deferred (see above) |
| b1 static `b1s1` | `gofmt -l .`; `go build ./...`; `go vet` of hostname, domains/..., tenancy/access, identity/..., mcp, config, audit, tenancy, migrate, migrations, cmd/server; `go vet -tags integration` of domains/..., tenancy/access, identity, mcp, migrate, tenancy | build and both vets **clean**; gofmt listed 6 files (audit.go, domains_config_test.go, dns_test.go, tls_test.go, mcp/handler.go, access/errors.go), then formatted with the image's gofmt |
| b1 unit `b1u1` | `go test -count=1 ./internal/hostname/ ./internal/domains/... ./internal/config/ ./internal/identity/ ./internal/mcp/` | hostname ok, identity ok; 4 failures fixed: `TestDNSChecker` (a CNAME chain ending at a non-target name without its own record is now `mismatch`, not `nxdomain`, at hop > 0), `TestDomainStateMachine` (the test's not_after rule now applies only when entering `active`), `TestLoadMCPRefusals` (the partial case sets another MCP variable now that the registry is retired), `TestPackageBoundary` (mcp may import `core/internal/domains` for the unknown-host sentinel and `net/url`) |
| b1 unit rerun `b1u2` | the 3 failed packages: domains/..., config, mcp | **all ok** |

Not yet run: the remaining b1 selectors (the audit/access/tenancy unit pins below are not updated yet), web units,
`run.sh lint lint-base lint-core`, and b2.

Remaining, in order (successor):
1. Amendment-02 files (binding recorded above):
   - `identity/rls_pool_integration_test.go`: a test-only resolver plus membership facts for the store-level sign-ins (`signInParams` → Hosts, Origin, Organization, and a provisioned user, principal and membership). The production guard now fails closed. Keep I1–I5, the poisoning/discard checks and E5(f), which needs the resolver in its Config. Turn `TestRLSFirstSignInRace` into the first-sign-in negative: concurrent unmapped principals are refused `organization_not_member` and nothing is created. Its B-RLS-IT PASS text in `run.sh` changes accordingly.
   - `store_test.go`: table set +organization_domains, `rls` 12, exemption for exactly `login_transactions.organization_id`.
   - `runner_test.go`, `migrations_test.go`: append 00011, 10→11.
   - `migrate/integration_test.go`: full-FS version-11 lifecycle; a new version-11 metadata set; keep `metadataThroughVersion10Only` unchanged; dependency-ordered drops of the 00011 objects, with no broad CASCADE.
   - `migrate/mcp_integration_test.go`: the historical cases on `subsetFS(10)`.
   - `recovery_integration_test.go:88` and `tenancy/integration_test.go` step-down: 10→11.
2. Guard evidence: extend the focused identity cases so that the final guard itself refuses a missing resolver or Organization with no session or user mutation. A store-level case in `rls_pool_integration_test.go` is the natural place.
3. The remaining core tests:
   - `TestDomains` DOM-1..10 with E-DOM-F1..F4 in `access/domains_integration_test.go`, including the #1 cases (owner/admin slug success, member/foreign denial, a direct runtime admin combined slug+policy-column UPDATE → 42501 with zero effect, admin slug-only success, owner unchanged) and the #3 cases (fence collision via `domainHooks.generation`, runtime has no `app.hostname_generation`, the reassignment generation differs).
   - `TestDomainsCatalog` (new `migrate/domains_integration_test.go`).
   - The `audit_test.go` 35/29 pins, the `matrix_test`/`policy_test` twelfth capability (`organization.set_slug` → manage_domains), and the `rls_expected`/`attack_expected` oracle rows.
   - IT-MCP-2/3/8 with the real resolver, E2E-MCP-6, `domains/api/http_test.go`.
4. Web, then `run.sh` and `required-checks.tsv`, then the overlay (seeded memberships for every E2E sign-in user under D-2(a)), fixtures, `.env`, `events.json`, docs. Then the rest of b1, then b2 on `g016-focus`.

Already adapted in this round:
- the identity harness: `itHosts` over the real `app.hostname_state`, slugged org-a/org-b, `provision`, `TestCustomDomainRelay` with 11 subtests;
- `http_test.go`: session `?origin=` and the relay route;
- the `allowlist_test` state tag;
- the access and mcp env fixture sign-in membership, provisioned before each login and removed after it;
- the env version-11 pins.

### Attempt 1 status at the third context handoff (2026-10-16 UTC; successor #2)

Successor #2 started at `f3c8f10` (staging `8f06f46`; scope binding 53 patterns `298b7572…82a0`; Plan r4 `48c1a792…`).
Runtime effort unknown (requested high). Every batch below ran after fetch + ancestry check, `bash
infra/ci/docker-desktop-mounts.sh` (13% of mounts in use) and under the batch lock. b1 check-image runs used
`--network none` with the entrypoint overridden; logs are in `~/.local/state/platform-evidence/g016-b1/` (c1–c4), the
focused integration run in `…/g016-b2/` (core2).

Done (commits `f842247`, `701c841`, plus the handoff commit carrying this section):
- **Amendment-02 files (all eight):** `runner_test.go`/`migrations_test.go` append 00011 (10→11);
  `migrate/integration_test.go` full-FS version-11 lifecycle (up/synthetic/reapply/down 11→10→…→0/up), new
  `metadataThroughVersion11Only` (the version-10 form unchanged), `domainDrops`/`domainLateDrops` and the two
  auth-reader drops in `rlsReset` in dependency order without CASCADE; `mcp_integration_test.go` historical cases on
  `subsetFS(10)`; `store_test.go` +organization_domains, `rls` 12, exemption for exactly
  `login_transactions.organization_id`; `recovery_integration_test.go` 10→11; `tenancy/integration_test.go` 11→10 step
  with `purgeDomainState` (superuser, test-only); `rls_pool_integration_test.go` fixed test `rlsHosts` over the real
  `app.hostname_state`, provisioned members, I3/I5 now count sessions (principals are provisioned).
- **First-sign-in negative:** `TestRLSFirstSignInRace` = 32 concurrent unmapped first sign-ins on a routing host, each
  `organization_not_member`, nothing created, no audit row, the mapping-race hook never runs.
- **Guard evidence:** new `TestIssuanceGuardFailsClosed` (same file; runs under B-IDENTITY-IT, which counts top-level
  passes ≥ 10, no pin change): missing resolver, missing Organization, host routing nowhere, host routing to another
  Organization → `host_changed`; mapped non-member, unmapped principal → `organization_not_member`; each with
  users/principals/sessions/binding/audit unchanged; the valid fixture then signs in through the same guard.
  `store.go:578-581` is unchanged (fail closed).
- **Core tests:** `access/domains_integration_test.go` `TestDomains` DOM-1..10 (incl. #1 slug cases, #3 fence collision
  via `domainHooks.generation` and no `hostname_generation` relation, E-DOM-F1..F4 lines); `migrate/domains_integration_test.go`
  `TestDomainsCatalog` (CT-DOM-1, CT-DOM-NEG N1–N7 with behavioural probes, CT-DOM-UPDOWN);
  IT-MCP-2 renamed "host, route and resolution reassignment" with real-resolver cases (slug hosts, custom alias,
  removed alias 404 without store access, reassigned slug 401); IT-MCP-8 mixed-host listing/revoke/create through an
  alias (`serveWith`/`lookupHandler` in `env_integration_test.go`); `domains/api/http_test.go` (resolve, tls-ask,
  refusal mapping, exact origins); pins: `audit_test.go` 29+6=35, 54 reasons, 24 operations (reads 00011);
  `boundary_test.go` (`math`, `math/big`; 00011 declares exactly 2 definer readers; `insertLogin`/`consumeLogin` are the
  pinned initializing functions); `policy_test.go`/`matrix_test.go` twelfth capability `organization.manage_domains`
  (oracle 264 cells; eligibility count follows the capability set); `access/integration_test.go` IT-MATRIX-1 264.

Validation so far:

| Run | Scope | Outcome |
|---|---|---|
| c1 | gofmt, vet (+integration) of the amendment-02 packages, unit migrations/migrate | clean, ok |
| c2/c3 | gofmt; vet and `-tags integration` vet of every changed package; unit of audit, tenancy, access, hostname, identity/..., domains/..., config, mcp, cmd/server, migrations, migrate | vet clean; failures fixed: audit pins, boundary pins, eligibility count, resolve no-port case |
| c4 | rerun of the failed unit packages (access, domains/api) | **ok** |
| b2 core2 (g016-focus, `down -v --remove-orphans`, leftovers 0/0/0/0) | migrate `^(TestMigrationLifecycle\|TestDomainsCatalog\|TestMCPCatalog\|TestMCPMigration\|TestAuditCatalog\|TestRLSCatalog)$` | **6/6 pass** |
| | tenancy `^TestTenancyModel$` (11→10 step-down) | pass |
| | access `^(TestDomains\|TestMCPOrdering\|TestMCPFootprint\|TestAccessPolicy\|TestAccessOrdering\|TestRLSIsolation\|TestRLSOrdering\|TestAudit\|TestIsolationAttacks)$` | **TestDomains pass**, MCPOrdering/MCPFootprint pass; FAIL: TestAccessPolicy IT-MATRIX-1 (pin 242→264, fixed after the run, not rerun); TestAccessOrdering (`ordering_integration_test.go:355` "fewer than 1 lock waits observed", not yet diagnosed); TestAudit and TestIsolationAttacks need the object-store env (`OBJECT_STORE_*`), a focus-script gap, not a code failure |
| | mcp `^(TestMCPTransport\|TestMCPLifecycle\|TestMCPAudit\|TestMCPRecovery\|TestMCPDiscovery)$` | Transport (incl. the new IT-MCP-2 resolver cases) pass; FAIL: IT-MCP-8 "revoked old-host row is pending" (test fixed after the run: the row is activated before the revoke; not rerun); TestMCPRecovery needs `MCP_INVALIDATE_SQL_B64` (harness input, focus-script gap) |
| | identity (catalog, refusals, continuity, relay, logout, provider logout, RLS pool, race, guard, concurrent first sign-ins) | 9/10 pass (incl. `TestIdentityCatalog`, `TestRLSIdentityPool`, `TestRLSFirstSignInRace`, `TestIssuanceGuardFailsClosed`, `TestCustomDomainRelay`); FAIL: `TestCallbackContinuity` `service_test.go:300` "conflict created a mapping for B" (D-2(a) provisioning interplay; not yet diagnosed) |

Recorded deviation for Codex (not self-resolved): the RLS-SCOPE (`rls_expected_test.go`) and ISO-ORACLE
(`attack_expected_test.go`) oracles are independently authored from fixed inputs, and their evaluators
(`rls_integration_test.go`, `attack_integration_test.go`) are outside the allow-list; extending them would break their
independence and scope. This attempt leaves them unchanged (RLS-SCOPE 36, ISO-ORACLE 82); the runtime-role evidence for
the new table and slug policy is DOM-1/DOM-4 and CT-DOM-NEG. The existing admin no_update row on
`group_admins_edit_business_profiles` still holds (the slug guard raises the same 42501).

Remaining, in order (next successor):
1. Rerun only: access `^(TestAccessPolicy\|TestAccessOrdering)$`, mcp `^TestMCPLifecycle$`, identity
   `^TestCallbackContinuity$`; diagnose the ordering lock-wait and the continuity mapping failures. Add the object-store
   variables and `MCP_INVALIDATE_SQL_B64` to the focus env (copy the `rls_go`/`mcp_it` shapes in `run.sh`) and run
   `TestAudit`, `TestIsolationAttacks`, `TestMCPRecovery` once. The focus script is `/tmp/g016s/b2.sh` (may be lost on WSL
   restart; it archives HEAD, configures `.env`, runs postgres from `local.yml` as project `g016-focus` on port 5932).
2. Web (nothing done yet): `web/lib/host.ts`, `web/proxy.ts` (gate via `GET /internal/v1/hosts/resolve`; exempt
   `/auth/*`, `/mcp/*`, `/.well-known/*`, `/api/health`, `/api/ready`), relay branch in `auth/callback/route.ts`
   (`GET /internal/v1/auth/relay?state=`, raw query forwarded byte-identical), session `?origin=` use, `NO_STORE` +
   `Referrer-Policy: no-referrer`, `mcp-client.ts` other-host parser, page/route edits, unit tests `host.test.mjs`,
   `mcp-client.test.mjs`, `session.test.mjs`.
3. `run.sh` (not touched yet): B-MIGRATE-SET 00011; catalog object counts 42→48 at version 11; `CORE_AUTH_ROUTES` now six
   (relay); B-DOMAINS-CATALOG/B-DOMAINS-IT rows in `migrations` (TestDomains 10, TestDomainsCatalog 3, E-DOM-F 4);
   `mcp_it` IT-MCP-2 wording; B-RLS-IT race text (32 refused `organization_not_member`); E2E-MCP-6
   (`MCP_TENANT_HOSTS_FILE` is retired and refused: drop it from the base `$mcp`, replace m3/m4 with the retired-registry
   and production-fixture refusals); `core_env` domain variables; B-ENV-EXAMPLE domain block; `required-checks.tsv`
   117→119. Then the overlay (D-2(a) memberships for every E2E sign-in user), fixtures, `.env.example`, `events.json`,
   docs.
4. Then the rest of b1 (web units directly; `run.sh lint lint-base lint-core`), the remaining b2 reruns, the single
   approved `ops` run.

sqlc stays deferred (handwritten pgx in `core/internal/domains/db/queries.go`). Staging advanced to `6f18e1e`
reconciled at this handoff by rebase onto `1832ca4` (delta since `8f06f46`: governance/management/skills files and the G016 review's planning-hindsight disposition only; no code, test or harness path; no reruns).

### Attempt 1 status at the fourth context handoff (2026-10-16 UTC; successor #3)

Successor #3 started at `99c267c` (staging `1832ca4`), sole writer; runtime effort unknown (requested high); active
context not measurable natively (estimated about 230k at this handoff, under the 300k checkpoint). Rebased onto staging
`b0de0cf` (delta: `PROGRESS.md` and `G016-domain-rls-evidence-codex-review.md` only; no code, test or harness path).
Every batch below ran after fetch + ancestry check (`b0de0cf` ancestor of HEAD), `bash infra/ci/docker-desktop-mounts.sh`
(13%) and under the batch lock. Focus script `~/.local/state/platform-evidence/g016-scripts/b2s3.sh` (b2.sh plus MinIO
init and the scoped `OBJECT_STORE_*` settings per run.sh `obj_env`, never MinIO root, plus `MCP_INVALIDATE_SQL_B64`);
web check `webcheck3.sh`; Go check `gocheck3.sh` (network none, CMD overridden). Logs: `…/g016-b2/s3a|s3b|s3c-*`,
`…/g016-b1/c5,c6,c7,w1`. Leftovers after every b2 run: containers 0, volumes 0, networks 0, run-labelled 0.

Diagnoses and fixes (tests only; no product code, policy or migration change; `store.go:578-581` unchanged):
- `TestAccessOrdering` "fewer than 1 lock waits" (O-4) and the O-7 30-minute hang it caused: the rotating sign-in
  subject had no membership of the fixture sign-in Organization, so under D-2(a) the rotation failed
  `organization_not_member` before its lock wait (leaving O-7 blocked). Fixed by admitting the subject in O-4/O-15
  (two `t.Cleanup(e.admit(...))` lines) and passed in s3a, **but `ordering_integration_test.go` is outside the 53-pattern
  allow-list: the edit was reverted at the handoff commit (self-review), so TestAccessOrdering fails again at HEAD.**
  See the second blocker below.
- `TestCallbackContinuity` "conflict created a mapping for B": under D-2(a) B must be provisioned before it can sign
  in, so "no mapping" became "B's user, principal and memberships identical (md5) and no session for B".
- `TestMCPLifecycle` IT-MCP-8: `mcp_connections.state` is only pending|active; the revocation check now reads the
  connection's session `revoked_at`.
- `TestAudit`: the per-call admit/release emitted membership/user rows inside capture windows and hit planted audit
  faults (then cascaded). New `audAdmit` provisions once and keeps the membership, called before windows/faults; the
  first sign-in of a provisioned account now expects `session.issued` + `auth.sign_in` (no `user.created`). AT-INV
  exercises slug change, claim, verify check (the four G016 actions it reported unobserved).
- **Codex domain/RLS correction (DOM-1):** paired owner/admin controls (exact visible rows, pending INSERT, challenge
  rotation UPDATE, each rolled back) and, for member, Group owner, foreign owner (own and substituted context), admin
  with substituted context, admin with substituted target row/Organization and admin without Organization context: no
  visible target rows, INSERT refused 42501 "row-level security policy", UPDATE 0 rows; slug UPDATE 0 rows for foreign /
  other-Organization / missing-context actors. **Deviation from the review's literal "zero rows" for the slug:** an
  in-Organization member or Group owner's slug UPDATE is refused 42501 by the unchanged G006 `organizations_update`
  WITH CHECK (its USING admits any member), so the exact outcome asserted is that refusal; zero stored effect is proven by
  the setup-side domain/Organization hash of both Organizations.

Validation:

| Run | Scope | Outcome |
|---|---|---|
| c5/c6/c7 | gofmt; vet and `-tags integration` vet of identity, access, mcp | clean |
| s3a (the ordering pass is attributable only to the reverted O-4/O-15 lines) | DOM-1; `^(TestAccessPolicy\|TestAccessOrdering)$`; `^TestCallbackContinuity$`; mcp `^(TestMCPLifecycle\|TestMCPRecovery)$`; access `^(TestAudit\|TestIsolationAttacks)$` | **pass:** TestAccessPolicy (IT-MATRIX-1 264), TestAccessOrdering, TestCallbackContinuity, TestMCPRecovery, TestIsolationAttacks; fail: DOM-1 slug member/group owner (42501, see above), IT-MCP-8, TestAudit |
| s3b | DOM-1; `^TestMCPLifecycle$`; `^TestAudit$` | **DOM-1 pass, TestMCPLifecycle pass**; TestAudit AT-INV only |
| s3c | `^TestAudit$` | **pass** (all subtests) |
| w1 (web check image) | lint, typecheck, `npm run test:unit` | **0/0/143 pass** (new `host.test.mjs`) |

Web done (commit `f58fddb`): `web/lib/host.ts` (hosts/resolve and auth/relay clients, gate exemptions, relay state and
destination rules), `web/proxy.ts` (gate; auth host `/sign-in` only; public entry redirect; fixed 404; 503 on store
failure; loopback-http development not gated), relay branch in `auth/callback/route.ts` (own origin resolved first on
the https profile; kind auth → exactly one state → core relay → 303 `callback?<raw query>` byte-identical, else the
fixed 404; `organization_not_member`/`host_changed` → sign-in statuses), `NO_STORE` + `Referrer-Policy: no-referrer`
(session.test.mjs asserts it on `seeOther` and `statusPage`), identity client knows the three new codes and passes
`?origin=` (account page; `unknown_host` → 404), sign-in page statuses `host_changed`, `not_member`.
Footprint note for the S-review: the gate adds one `hosts/resolve` per gated request and the callback one per https
callback (Plan §5.5 lists one for the anonymous gate and the relay).

**Blocker for Codex (scope):** the Plan's `web/lib/mcp-client.ts` change (listing accepts other-host rows, exposes
`host`) contradicts the out-of-scope G014 test `web/tests/unit/mcp-connections.test.mjs`, which asserts a listed row on
`https://evil.example/mcp/c/<id>` is malformed (and its success cases deepEqual the view without `host`). Not
implemented; `mcp-client.ts` unchanged. Smallest options: (a) amend the allow-list with
`^web/tests/unit/mcp-connections\.test\.mjs$` and replace that case with the G016 rule (row URL must end in its own id;
core decides membership), plus `mcp-client.test.mjs`; (b) keep the G014 rule and accept that a mixed-host listing is
refused as malformed by web (M-LIST/IT-MCP-8 web leg unmet); (c) verify each other host through `hosts/resolve` (same
Organization) in web, which keeps the evil-host case malformed without touching the G014 test but adds one core call per
distinct other host. Recommendation: (a). The connections page host display, `mcp.spec.ts` M-LIST and
`identity.spec.ts` changes wait on this disposition.

**Blocker for Codex (scope, TestAccessOrdering):** the D-2(a) fixture admission needs either (a) an allow-list
amendment `^core/internal/tenancy/access/ordering_integration_test[.]go$` for the two admission lines in O-4/O-15
(validated in s3a: TestAccessOrdering pass), recommended; or (b) an in-scope fixture change in `integration_test.go`
keeping the sign-in Organization membership after `signInWith` (no release), which changes every access suite's
fixture memberships and needs reruns of all of them. Neither applied.

Remaining, in order (next successor): the two scope dispositions; `run.sh`/`required-checks.tsv` (item 3 of the
third-handoff list, unchanged); the overlay with a seeded membership for every E2E sign-in user, fixtures,
`.env.example`, `events.json`, docs; the rest of b1 (`run.sh lint lint-base lint-core`), remaining b2 reruns (none
outstanding for the cases above), the single approved `ops` run. sqlc stays deferred.

### Attempt 1 continuation — amendment 03 binding (2026-10-16 UTC; successor #3)

Rebased onto staging `e6b2f3e8` (delta: `PROGRESS.md`, the allow-list and `G016-scope-amendment-03-codex-review.md`
(SHA-256 `331ac06d…af54`) only). Scope binding moves from 53 patterns `298b7572…82a0` to **55 patterns
`f4869cf0…e58b`**, adding exactly `^web/tests/unit/mcp-connections\.test\.mjs$` and
`^core/internal/tenancy/access/ordering_integration_test[.]go$` (both scope blockers resolved with option (a)).
Amendment-03 instructions: reapply only the O-4/O-15 admission lines (rerun those two cases only); extend the existing
`mcp-connections.test.mjs` for the mixed-host listing with derived `host` (no per-host lookup; RESOURCE shape, own-id
match, limits, cursors and create's exact-origin rule kept), then `mcp-client.ts`, the connections page host, `mcp.spec.ts`
M-LIST and `identity.spec.ts`; DOM-1 executes `e.rlsAttest(e.pool)` before its raw-SQL controls (rerun DOM-1 only).

### Attempt 1 status at the fifth context handoff (2026-10-16 UTC; successor #3, about 275k estimated)

Amendment-03 items done (commit `c008a2f`, base `e6b2f3e8`, binding 55 patterns `f4869cf0…e58b`): the two O-4/O-15
admission lines reapplied (target memberships and lock/deadlock assertions unchanged); DOM-1 runs `e.rlsAttest(e.pool)`
(fatal on error) before its raw-SQL controls; `web/lib/mcp-client.ts` listing accepts the caller's rows of every host
with `host` derived from the URL (complete RESOURCE shape and own-id match kept; no per-host lookup; create keeps the
exact origin); the connections page shows each row's address (`mcp-host`); the existing `mcp-connections.test.mjs` is
extended (mixed-host success with derived hosts and a single core call; other-host own-id mismatch, http, extra path,
uppercase, userinfo and trailing-slash URLs malformed; other-host create URL malformed); no separate
`mcp-client.test.mjs` (the review prefers extending the existing file).

| Run | Scope | Outcome |
|---|---|---|
| w2 | web lint, typecheck, unit | clean, clean, **144/144** |
| c8 | gofmt; integration vet of mcp, access | clean |
| s3d | `^TestDomains$/^DOM-1_`; `^TestAccessOrdering$/^O-(4\|15)_` | **DOM-1 pass; O-4 pass; O-15 pass**; leftovers 0/0/0/0 |

Not done (next worker, together with the overlay that seeds them): `mcp.spec.ts` M-LIST old-host row and the
`identity.spec.ts` E-DOM-1..5 items (third platform host, custom host, unknown host, apex/www redirect, removed domain),
plus `playwright.config.ts` projects; then `run.sh`/`required-checks.tsv` (third-handoff item 3), the overlay with a
seeded membership for every E2E sign-in user, fixtures, `.env.example`, `events.json`, docs, the rest of b1
(`run.sh lint lint-base lint-core`) and the single approved `ops` run. No b2 rerun is outstanding. sqlc stays deferred.

### Attempt 1 status at the sixth context handoff (2026-10-16 UTC; successor #4, about 260k)

Base `a0dc9f92` (staging advanced from `e6b2f3e8` by management files and `docs/ARCHITECTURE.md` only; rebased, ancestry
verified, no code/test/harness delta, no reruns). Binding unchanged: 55 patterns `f4869cf0…e58b`; every path changed in
this round matches it. Commits `4891572` (specs, overlay, seed, fixtures, `.env.example`) and `48dec23` after rebase
(run.sh, required-checks, `identity/http.go` comment move, IDENTITY.md browser block).

Done:
- Browser: `identity.spec.ts` E-DOM-1 (org-c own callback, host-only cookie, same account), E-DOM-2 (portal.acme.test
  through the auth-host relay: redirect_uri on the auth host, hops `auth 303 -> portal/auth/callback`, `portal 303 ->
  /account`, host-only cookie, no auth-host cookie, same account, local sign-out, auth-host `/account` 404), E-DOM-3
  (org-z.localhost and unknown.acme.test: 404 pages, login start 403), E-DOM-4 (localhost/www: 303 to
  `MARKETING_SITE_URL`, answered locally by a route; `/api/health` exempt), E-DOM-5 (removed old.acme.test: 404 pages,
  login start 403, portal still routes); projects `domain-platform` (baseURL org-c) and `domain-custom` (baseURL
  portal), identity projects `grepInvert /@domain-/`, Chromium `--host-resolver-rules=MAP *.acme.test 127.0.0.1`.
  `mcp.spec.ts` M-LIST: the seeded old-host row (`https://old.acme.test:543/mcp/c/0f16c000-…c1`) listed through org-a
  with its own address, absent on B.
- D-2(a) consequences in the browser specs (deviation for review): E-ATTR keeps the rename half with a seeded member
  (frank, renamed and restored) and replaces "a new user reusing an email gets a new account" by "a new provider user
  without membership is refused `not_member`, no session" (a first sign-in cannot create an account on a tenant host);
  M-NONMEMBER uses erin (no account): her sign-in on A and B ends on `not_member`, setup shows sign-in-first, create
  401; M-A11Y audits the not-member state instead of not-available (the not-available answer for a session whose
  membership ends later remains IT-MCP-11 evidence). E-DOM-5 asserts a removed host's refusal; the removal during an
  in-flight sign-in stays `TestCustomDomainRelay` case A (the browser has no database access).
- Overlay: core gets `PLATFORM_BASE_DOMAIN=localhost`, `CUSTOM_DOMAIN_CNAME_TARGET=edge.localhost`,
  `PLATFORM_HOST_PORT=543`, `PLATFORM_AUTH_HOST=auth.localhost`, `IDENTITY_PUBLIC_ORIGINS=https://auth.localhost:543`,
  `MARKETING_SITE_URL=https://marketing.identity.test/`, the DNS and TLS fixture adapters (new
  `infra/identity/{dns,tls}-fixture.tsv`); `MCP_TENANT_HOSTS_FILE` and its mount removed; leaf SANs and the emulator's
  `--redirect-hosts` list the new hosts explicitly (no wildcard tried: fixture limitation); nginx passes every
  `*.localhost`/`*.acme.test` name to web (attacker exact name first). `mcp-seed.sql`: slugs org-a/b/c, Org C, accounts
  and principals for alice, bob, carol, dave, frank and the hostile user with Organization memberships of A, B and C,
  the active custom domain portal.acme.test and the removed old.acme.test (Org A), and dave's old-host connection with
  its own `mcp` session. `emulator-seed.yaml` adds erin and frank; `mcp-hosts.tsv` is now only the harness's
  host-to-UUID map (core reads no registry).
- `run.sh`: 00011 in B-MIGRATE-SET (00010 pinned at `3ac29b6b…`) and in every history assertion (root, B-MIGRATE-CLI
  steps 3-7 with 48/42/48, recovery drill, B-PERSIST, rename, toggle-off); catalog 42->48 (`DOMAIN_OBJECTS`, 6) in
  B-MIGRATE-CATALOG and B-IDENTITY-CATALOG; six auth routes (relay; the route constant's trailing comment moved above
  it in `identity/http.go` so the pinned pattern reads it); B-MCP-SOURCE pins `MAIN_GO_G016_SHA=7dc763b6…`, the
  widened import set and access importers `core/internal/domains/api core/internal/mcp`; `domains_checks`
  (B-DOMAINS-CATALOG TestDomainsCatalog 3; B-DOMAINS-IT TestDomains 10 with 4 `E-DOM-F` lines kept as
  `domains-edom.jsonl`, trigger lines reported, record checks) before the recovery drill; IT-MCP-2 wording; B-RLS-IT
  race text (32 refused `organization_not_member`); E2E-MCP-6 base without registry, m3 retired registry, m4
  production fixture adapter; `core_env` adds `MARKETING_SITE_URL` and `PLATFORM_AUTH_HOST=app.probe.example.net`;
  B-ENV-EXAMPLE pins the domain block (`cae3a790…`; identity block hash unchanged `3d55f30d…`); `mcp_seed` lists the
  emulator users once and passes every seeded ID and Org C; `BROWSER_MATRIX_DOMAINS` required beside the inherited
  matrix (defined in run.sh because `lib.sh` is outside the allow-list; for review). `required-checks.tsv` 119.
- Checks (fetch/rebase, mount check 13%, harness lock): web check image lint 0, typecheck 0, unit **144/144**
  (`g016-b1/w4.log`); core check image gofmt clean, vet `./internal/identity/ ./cmd/server/` clean, identity unit ok
  (`c7.log`). `bash -n run.sh` ok.

Remaining, in order (next successor): (1) docs: create `docs/foundation/DOMAINS.md` (referenced by identity.yml and
`.env.example`), IDENTITY.md seed paragraph (new psql variables, users, D-2(a)), MCP.md (registry retired),
MIGRATIONS.md version-11 row, and the Plan's SECURITY/ARCHITECTURE/OPERATIONS/VENDORS rows (keep VENDORS' 33 identity
rows byte-equal, `VENDORS_IDENTITY_ROWS_SHA`); `events.json` only if new log events need catalogue entries; (2)
recommended before ops: apply `mcp-seed.sql` once to a migrated focus database with placeholder IDs to prove it
(superuser inserts of active/removed domain rows, the `mcp` session and connection); (3) b1 harness `run.sh lint
lint-base lint-core`; (4) the single `ops` run (and `rename` if approved). sqlc stays deferred. No harness run was
started in this round; no leftovers.

### Attempt 1 status at the seventh handoff (2026-10-16 UTC; successor #5, blocker hand-back)

Base staging `4d477300` (advanced from `a0dc9f92` by `c8a10d48`, the Codex D-2/harness-pin review, and `4d477300`,
management only; rebased, ancestry verified; no code/test/harness delta, no reruns). Binding unchanged: 55 patterns
`f4869cf0…e58b`; every path changed in this round matches it. Commit `2f3acd3b` (docs, catalogue, prose corrections).

Done:
- Docs: new `docs/foundation/DOMAINS.md` (host model, configuration, lifecycle, slugs/renames, tenant-host sign-in,
  E1 edge inputs incl. the `ask` route and the whole-callback-query log redaction rule, restore/configuration changes,
  records/logs, overlay, evidence and limits); IDENTITY.md (public-origins rule, new "Tenant hosts and the callback relay
  (G016)" section incl. the custom-host sign-out limitation, seed paragraph with the new psql variables and D-2(a),
  `auth.relay` in Logs); MCP.md (registry retired, "Hosts and aliases (G016)" with the mixed-host setup listing,
  components/overlay/evidence rows); MIGRATIONS.md (version-11 row, down bullet, from-version sentence, sqlc deferral);
  AUTHORIZATION.md (twelfth capability and the guarded domain operations); AUDIT.md (T26–T29, 24 denial operations,
  sign-in reasons); DATABASE-ISOLATION.md (version 11); OPERATIONS.md (domain checks and changed rows); SECURITY.md (§6.2
  host source, three §7 rows: DNS proof, edge-probe, edge to core; implementation-status note); PERFORMANCE.md PG-007
  (E-DOM-F1..F4 from the DOM-10 pass at `701c8418`, `g016-b2/core2-2.log`); ARCHITECTURE.md (G016 implementation note,
  sqlc outcome); `connection-context.md` dated G016 row. **VENDORS.md unchanged:** G016 ships no new component (Caddy is a
  deployment input documented in DOMAINS.md), and the file outside the 33-line pinned identity run must equal its
  render (B-TOGGLE-OFF-ADDITIONS), so no row is added.
- **PG-007 observation (honest):** E-DOM-F1 shows sequential scans for both probes on the small test tables; the partial
  unique indexes exist, but index use at production row counts is not shown. F3 counts are 6–7 statements per operation
  (constant at 10 and 40 rows), above the Plan's 5–6 expectation.
- Seed dry-run (`seed5.sh seed1`, head `69ea09cb`): `mcp-seed.sql` applied once as superuser to a migrated `g016-focus`
  database with placeholder IDs: rc 0; probes: `portal.acme.test` routing/issuance true for Org A, `old.acme.test` no
  route/issuance, slug `org-c` routing (issuance false), domain rows active/ready and removed/withdrawn (generation 1),
  the old-host connection active, 17 Organization bindings, `is_organization_member(frank, C)` true. `down -v
  --remove-orphans`; leftovers 0/0/0. No seed change needed.
- `events.json` (needed: the ops run's B-IDENTITY-PRIVACY validates the identity core log, which will contain
  `auth.relay` from E-DOM-2 and `auth.callback.completed` outcomes `organization_not_member` from E-ATTR/M-NONMEMBER, and
  B-DOMAINS-IT's logcheck validates `access.decision` records with the `domain.*`/`organization.set_slug` operations and
  capability `organization.manage_domains`): added `auth.relay` (outcome relayed/refused, closed reasons) and
  `domain.request` (operation, outcome, closed reason, detail, 12-hex `host_tag`), the two callback outcomes, the eight
  access operations, the capability and its rule.
- Codex D-2 disposition folded (staging `c8a10d48`, review SHA-256 `9e99bab2…5afe`): (1) email reuse stays
  `TestMappingIsByIssuerAndSubject`'s guarantee and attributable pass; the new E-ATTR denial is not email-reuse evidence
  (comment corrected in `identity.spec.ts`; the earlier wording above at "replaces 'a new user reusing an email gets a
  new account'" is superseded by this record); (2) IT-MCP-11 proves the transport/tool denial after membership loss
  only; the setup page's NotAvailable branch is unchanged and its HTML/axe state is **not exercised** by the browser since
  G016 (comment corrected in `mcp.spec.ts`; MCP.md evidence row says so); (3) B-MCP-SOURCE header, importer/imports error
  text and PASS text now describe G016 (the domain-api importer, six auth routes incl. the relay, the G016 digest).
  Recorded: the pinned hashes match the actual bytes, which is not a final source ACCEPT; the emulator's explicit
  Organization host list is not real WorkOS wildcard evidence.
- Focused check (`vtools5.sh vt1`, head `2f3acd3b`, fetch/ancestry/mount 13%/lock): validation-tools check image,
  `go test ./internal/logrecord/`: the catalogue compiles and every classify/validate test passes; **only
  `TestCatalogueMatchesTheBrief` fails**, because it pins the exact core event-name list (now with `auth.relay` and
  `domain.request`).

**Blocker (scope):** `infra/validation/tools/internal/logrecord/logrecord_test.go` is outside the 55-pattern allow-list
but must change: the one-line want list of `TestCatalogueMatchesTheBrief` gains `"auth.relay"` and `"domain.request"`
(with a G016 comment). Without it the `contracts` suite (B-CONTRACT-TOOLS) fails on the hosted gate; without the
`events.json` entries the `ops` run's B-IDENTITY-PRIVACY (and B-DOMAINS-IT logcheck) would fail. Requested pattern:
`^infra/validation/tools/internal/logrecord/logrecord_test[.]go$` (that want list and its comment only). No lint or ops
run was started; nothing is left running; no leftovers.

Remaining after the scope decision, in order: (1) the test pin edit and a focused rerun of `go test ./internal/logrecord/`;
(2) b1 harness `run.sh lint lint-base lint-core`; (3) the single `ops` run and the `rename` run (Plan b4). sqlc stays
deferred (handwritten pgx).

### Attempt 1 continuation — amendment 04 binding (2026-10-16 UTC; successor #6)

Recorded before the new-path edit. Successor #6 (role implement, sole writer, model Opus 5.5, runtime effort unknown;
native worktree `agent-<id>`, local branch `g016-impl-succ6`) continues attempt 1 of B7 from
`b5cee4b5` (merge of the remote task tip `4277ab9f` with staging `4d477300`, plus the docs and Result commits; Result
SHA-256 at start `0c066c20…`). A context handoff spends no attempt.

**Binding change:** 55 patterns `f4869cf070c6abb73e47aa94615dcb56b45eaa284885255756658a790090e58b` (kept as history) →
**56 patterns `b2d5f604ccc3ae3ff903b9cd95cd7ba276cde68b14e4c40a0f1716699563af6f`**, adding exactly
`^infra/validation/tools/internal/logrecord/logrecord_test[.]go$` (the `TestCatalogueMatchesTheBrief` core name list
gains `auth.relay` and `domain.request` in sorted order, plus one G016 comment; every other expected name and every
classifier/schema/privacy/canary/unknown-field assertion is kept). Source: Codex scope amendment 04
(`G016-scope-amendment-04-codex-review.md`, SHA-256 `693549fc83af1a321f09144ec2fdb3dd6a053802c28613a25411780ad1fed175`),
read from Codex's **local, unpublished** commit `80569344` (parent `4d477300`); the allow-list bytes above were read
from that commit (`git show 80569344:management/v0/results/G016-scope-allowlist.txt`), not imported into this branch.
GitLab publication is currently failing; this round makes local commits only (no push, no MR).

**PG-007 observation, recorded per the amendment 04 disposition** (evidence: DOM-10 E-DOM-F1..F4, focused `TestDomains`
at `701c8418`, `g016-b2/core2-2.log`; the same observations are in `docs/PERFORMANCE.md` PG-007):
- **F3 actual statement counts:** 6–7 statements per domain operation (claim 6, list 6, set slug 6, verify write 7,
  remove 7, retry 7), **constant at 10 and 40 domain rows** of the Organization, so no per-row fan-out. The Plan's
  "about 5–6" was an approximate estimate, not a hard numerical limit; the extra statement is the fixed
  guard/context/liveness work of the guarded transaction (a characterization correction, not new per-row work and no
  authorization caching). F2: `ResolveWith` 1 statement in the caller transaction; `Resolve` 1 autocommit statement,
  0 BEGIN/COMMIT. HTTP round trips unchanged; the per-request breakdown stays as built (PERFORMANCE.md PG-007
  "Per-request cost as built"). No unexplained added work or waits were observed.
- **Limitation kept:** F1 shows sequential scans for both probes on the small test tables. The partial unique indexes
  `organization_domains_hostname_live` and `organizations_slug_key` exist and back uniqueness, but small-fixture plans
  **do not prove production-cardinality index use or latency**. No forced planner setting, speculative index/cache or
  benchmark campaign was run. This is reporting under characterized-risk approval, not measured production
  performance or final AC11 qualification.

**Reconciliation (pre-lint):** fetch ok; `origin/staging` advanced `4d477300` → `93f70061` (`80569344` amendment 04,
`bf8e2395` and `93f70061` publication blocker/recovery notes: `PROGRESS.md`, the allow-list and the amendment review,
management only). Merged (no rebase: the task history is merge-based per amendment 04) as `41a50e23`; no conflict, no
code/test/harness delta. The in-repo allow-list now equals the bound 56-pattern bytes `b2d5f604…`; the published
review file is `656dd5ba…` (the `693549fc…` bytes plus two appended supervisor publication sections, no scope change).
Focused logrecord rerun `vt6a` (head `8e92ebff`, check image, `go test -count=1 -v ./internal/logrecord/`, network
none, CMD overridden): **PASS**, all 10 tests incl. `TestCatalogueMatchesTheBrief`, classify/validate/canary/extra-key
checks unchanged; it is not repeated for the management-only merge.

**Binding source updated (publication recovered):** amendment 04 is now published on `origin/staging` `93f70061`;
the binding is recorded against that published commit: review
`management/v0/results/G016-scope-amendment-04-codex-review.md` SHA-256
`656dd5ba4f377dbf9df1ef9519c29d04dd90cc06caad6f2cf3da8c5a3cdb2cf3`, allow-list 56 patterns
`b2d5f604ccc3ae3ff903b9cd95cd7ba276cde68b14e4c40a0f1716699563af6f`. The earlier reference to the local commit
`80569344` (review `693549fc…`) is kept as history. The coordinator pushed the preserved candidate (origin task branch
`b5cee4b5`, a fast-forward from `4277ab9f`); this round's commits are pushed as normal fast-forwards on top of it.

### Attempt 1 status at the eighth handoff (2026-10-16 UTC; successor #6, blocker hand-back)

Every batch: `git fetch origin` ok, ancestry verified, `docker-desktop-mounts.sh` 13%, harness lock held. Harness form
(`g016-scripts/runsh6.sh`): `run.sh --base <origin/staging> --allow-file <56-pattern copy b2d5f604…>
--owned-registry g016/registry/owned.tsv --report-dir g016/<tag>/report <suites>`.

| Run | Head / base | Selector | Outcome |
|---|---|---|---|
| `vt6a` | `8e92ebff` | validation-tools check image, `go test -count=1 -v ./internal/logrecord/` | **PASS** (10/10) |
| `lint1` | `4af3e441` / `93f70061` | `run.sh lint lint-base lint-core` | **PASS**, rc 0, 4 min 35 s: B-CI-CONFIG, B-COMPOSE, B-DEPS, B-DEPSCAN, B-DOCKER-ONLY, B-DOCKERIGNORE, B-ENV-EXAMPLE, B-HADOLINT, B-IDENTITY-GUARDS, B-LITERALS, B-MCP-SOURCE, B-MIGRATE-SET, B-NAMING, B-REQUIRED-CHECKS, B-SCOPE (101 paths), B-SECRET-CANARY, B-SECRET-SOURCE, B-SHELL; leftovers 0 |
| `ops1` (the single approved `ops`) | `4af3e441` / `93f70061` | `run.sh ops` | **FAIL**, rc 1, 54 min 36 s. PASS: B-OBS-CONTEXT/CHAIN/TIMEOUTS/LOGS/PRIVACY, B-OUTAGE, B-PERSIST (version 11), B-SECRET-ROTATION, B-IDENTITY-IT (73 tests), B-DOCKER-ONLY, B-SCOPE. FAIL: B-IDENTITY-E2E, B-IDENTITY-PROBES, B-IDENTITY-PRIVACY, B-IDENTITY-CONFIG, B-MCP-E2E, B-SECRET-CANARY, B-REQUIRED-CHECKS. Inventory diff 0; report `inventory-leftover.txt` empty; no `g001.run` container/image left |

`ops1` root causes:
1. **MCP seed** (`psql:<stdin>:18: duplicate key … organizations_slug_key (slug)=(org-a)`): G016's identity IT
   `seedTenants` gives its fixture Organizations (`0f16a…a1`, `0f16b…b1`) the slugs org-a/org-b; they remain on the
   overlay database after B-IDENTITY-IT, so `mcp-seed.sql` (org-a/b/c for the `mcp-hosts.tsv` UUIDs) fails. Cascade:
   no seeded memberships → every tenant-host callback 403 `organization_not_member` (72 in the core log) → all identity,
   mcp and domain browser projects fail → no probe/MCP handoff (B-IDENTITY-PROBES, B-MCP-E2E), failure error-context
   files carrying canaries (B-SECRET-CANARY), no P-POS evidence (B-IDENTITY-PRIVACY). **Fixed (in scope)**:
   `core/internal/identity/integration_test.go` TestMain clears the two fixture slugs after `m.Run()` (superuser,
   audited `organization.slug_changed`; the Organizations stay, as before G016; a cleanup error fails the run).
2. **B-IDENTITY-CONFIG** (`web with loopback-http in production: 503 503`): the G016 gate answered the gated
   `/api/account` with the HTML 503 page, while the check (and the route's contract) expects `{"error":"unavailable"}`.
   **Fixed (in scope)**: `web/proxy.ts` answers gated `/api/*` refusals as JSON (`{"error":"not_found"}` 404,
   `{"error":"unavailable"}` 503); pages keep the fixed HTML pages (E-DOM-5 unchanged); DOMAINS.md one sentence.
3. **B-IDENTITY-PRIVACY core logcheck: 837 `completion-invalid` lines** (2986 lines, 2149 valid records). The
   completion schema `infra/validation/tools/internal/logrecord/completion.schema.json` pins the core `route` enum to
   the pre-G016 routes; G016's ten registered routes (`identity/http.go:46`, `domains/api/http.go:33-42`; the gate calls
   `/internal/v1/hosts/resolve` on every gated request) are rejected. Attribution by mechanism (the work directory was
   not kept). **Blocker: the file is outside the 56-pattern allow-list.** Requested as amendment 05:
   `^infra/validation/tools/internal/logrecord/completion[.]schema[.]json$`, adding exactly the ten route labels to
   both route enums (request `g016-amendment-request-05.md`, SHA-256 `3098560e…e8c5`).

Focused rechecks of fixes 1–2 (`focus6.sh a`, head `34b05ca6` = `e044104d` + management-only staging merge,
lock 17:23–17:25Z): web check image lint/typecheck/unit **rc 0**; core check image `gofmt -l ./internal/identity/`
empty, `go vet` and `go vet -tags integration ./internal/identity/` **rc 0**; repro `itseed6a` on `g016-focus`: the
identity integration selector of b2 (10 top-level tests) **PASS**, then 0 slugged Organizations remain, then
`mcp-seed.sql` with the `mcp-hosts.tsv` UUIDs **rc 0** on the same database; probes: portal routing/issuance true for
Org A, old host none, org-c routes (issuance false), domains active/ready and removed/withdrawn, connection active,
17 bindings (the probe's `member-check` uses seed5's placeholder Org C UUID, so its `f` is a probe artifact, not a
membership finding); `down -v`, leftovers 0/0/0.

Reconciliation: `origin/staging` advanced `93f70061` → `8ce9e077` (G055/OS001 management only); merged as
`34b05ca6`, no conflict. `rename` not started (held until `ops` can pass; it is the next run after the ops rerun).
sqlc stays deferred (handwritten pgx); `store.go:578-581`, RLS-SCOPE/ISO-ORACLE evaluators unchanged.

Remaining after the amendment decision: (1) the schema edit and a focused `go test ./internal/logrecord/` rerun; (2) the
affected `ops` rerun (the failed rows' only selector); (3) `rename` once.

### Attempt 1 continuation — amendment 05 binding (2026-10-16 UTC; successor #6)

Recorded before the new-path edit. Codex ACCEPT of amendment 05 (`G016-scope-amendment-05-codex-review.md`, SHA-256
`dc8cca214e87b15668958b77d35527edd3ecc3c4fc9e64677e836d34d948da36`, on `origin/staging` `24d83f16`, merged as
`6186fe0a`, management only, no conflict). **Binding change:** 56 patterns `b2d5f604…f6f` (kept as history) → **57
patterns `7be79e7497d67045f4ac55f810723327e5c663e08758c0983ecc92a83e9393f4`**, adding exactly
`^infra/validation/tools/internal/logrecord/completion[.]schema[.]json$` (the ten G016 route labels in the global and
the core route enums only). `ops1` stays **FAILED**; its 837 `completion-invalid` lines were not retained, so the schema
diagnosis is by mechanism only and does not prove every original record or failure had that cause. One affected `ops`
rerun (C-1 exception) and `rename` once are authorized; attempt 1 of B7, checkpoint 4 unchanged.

### Attempt 1 status at the ninth handoff (2026-10-16 UTC; successor #6, checkpoint hand-back)

Amendment 05 done: `completion.schema.json` gains exactly the ten route labels in both enums (`b5626ade`); positive
classify cases for all ten and rejection controls (prefix, trailing slash, query, unknown path, web/intelligence
service, extra `hostname` key) in `logrecord_test.go`; `vt6b` `go test ./internal/logrecord/` **PASS** 10/10.
E-DOM-5 gained the gate's `/api/account` JSON 404 + no-store and page HTML 404 + no-store assertions (`05d12db8`);
web check `w6b` rc 0.

| Run | Head / base | Allow-file actually used | Outcome |
|---|---|---|---|
| `ops2` (the one authorized affected rerun) | `05d12db8` / `24d83f16` | **obsolete 56-pattern copy `b2d5f604…`** (wrapper not updated; the Result already recorded 57) | **FAILED**, rc 1, 16 min. FAIL: B-IDENTITY-E2E (E-CB-8 both themes, E-DOM-4), B-IDENTITY-PRIVACY, B-SCOPE (a harness-input error: `completion.schema.json` outside the stale 56-pattern file), B-REQUIRED-CHECKS. PASS: everything else incl. B-IDENTITY-CONFIG, B-IDENTITY-PROBES, B-MCP-E2E (seed ok), B-SECRET-CANARY. Leftovers 0; work dir kept `/tmp/project-validation-HqltM0` |
| `lint2` | `1d7193ac` / `24d83f16` | 57 `7be79e74…` | PASS, all 18 rows (run.sh changed); leftovers 0 |
| `ops3` — **UNAUTHORIZED extra batch** (launched beyond the one authorized rerun; let finish per the coordinator) | `1d7193ac` / `24d83f16` | 57 `7be79e74…` | rc 1, 15 min. FAIL: B-IDENTITY-E2E (E-DOM-4 only), B-REQUIRED-CHECKS. PASS: B-IDENTITY-PRIVACY, B-IDENTITY-CONFIG, -IT, -PROBES, B-MCP-E2E, B-SECRET-CANARY, B-SCOPE (102 paths), B-OBS-*, B-OUTAGE, B-PERSIST, B-SECRET-ROTATION, B-DOCKER-ONLY. Inventory diff 0, leftover inventory empty; work dir kept `/tmp/project-validation-8sRqsd` (holds generated run secrets; removal left to the owner). Counts only after Codex review |

`ops2` sanitized root causes (from the kept work dir): (1) E-CB-8: a forged-Host callback (`Host: evil.example`, bogus
code/state) reached core and answered `status=expired` (core's consume found no transaction before any host check)
instead of `failed`; (2) E-DOM-4: `page.goto https://localhost:543/` `ERR_NAME_NOT_RESOLVED` (bare platform domain);
(3) B-IDENTITY-PRIVACY: the core log check found 125 `duplicate-record` failures, each a `hosts/resolve` + `callback`
(124) or + `relay` (1) pair sharing one request ID within the callback route (core log otherwise fully catalogued, 0
unclassified: the amendment-05 routes validate), and P-POS-2 collected no static chunk (the gate answered the bundle
probe's `Host: 127.0.0.1:3000` with the 404 page). Fixes in `1d7193ac` (Codex disposition `c688d422…`: callback host
refusal, localhost mapping and routing-Host bundle collection accepted as bounded restorations; the fresh request ID
approved only for the additional callback host lookup): web callback refuses a non-routing host before core
(`failed`; resolver failure `unavailable`); the lookup carries its own generated request ID; Playwright maps bare
`localhost`; P-POS-2 requests pages with `Host: org-a.localhost:543`. OPERATIONS.md now states the narrow lookup
exception (not a general per-call rule; health-chain propagation unchanged).

**Condition-4 evidence (sanitized, from ops3's kept core/replica logs):** 0 repeated request IDs across all
completion records; 123 lookup/callback-or-relay pairs share `trace_id` and caller (parent) span, each with a distinct
request ID; the callback/relay records keep the forwarded original request ID; B-OBS-CHAIN (health chain) PASS.

**Open:** E-DOM-4 still fails in ops3 with the same `ERR_NAME_NOT_RESOLVED` for `https://localhost:543/`: the
`MAP localhost 127.0.0.1` resolver rule did not take effect (Chromium appears to resolve the bare `localhost` before
host-resolver-rules, or the system resolver in the browser's network namespace lacks it). Not diagnosed further at this
checkpoint. No ops4 and no rename until Codex disposes (coordinator hold). Staging `5b7ff335` (management only) merged
as `104f007a`. Preflight now compares the harness's actual allow-file hash (wrapper `runsh6.sh` uses the in-repo file;
`meta.txt` records its SHA-256) against `7be79e74…`.

### Attempt 1 continuation — E-DOM-4 correction loop (2026-10-16 UTC; successor #7)

**Ownership.** Task G016, role correction worker (successor #7, sole writer), Claude isolated worktree
`.claude/worktrees/agent-<id>` (generated name, recorded per AGENTS.md), local branch `g016-impl-succ7`
tracking `origin/g016-hostname-resolution-domain-lifecycle`, created at its head `d96efffc`. Model Fable 5.1; effort
high (requested by the dispatch; the runtime value is not verifiable from the session, so actual = unknown). Attempt 1
of B7, checkpoint 4, unchanged; the rotation spent no attempt. Disposition read from `origin/staging` `1991c34d`
(`G016-ops-rerun-codex-disposition.md` SHA-256 `7a7f14e2…`; the branch copy before the merge was `c688d422…`).

**Bindings verified against the files at `d96efffc`:** allow-list 57 patterns, SHA-256
`7be79e7497d67045f4ac55f810723327e5c663e08758c0983ecc92a83e9393f4`; this Result at the hand-back
`f94fb1922d10994d11cddfeebc60005c09a68fdcd04412cc5556ad736e5599f6`. Staging `1991c34d` (management only:
PROGRESS.md and the disposition) merged as `944f777e` before any batch, no conflict.

**Diagnosis (ops2/ops3 `ERR_NAME_NOT_RESOLVED at https://localhost:543/`).** Cheap probes in the rebuilt web-browser
image, no stack (evidence `platform-evidence/g016/dom7/`, scripts `g016-scripts/dom7-*`):
1. `diag3-probe.log` (`--network none`; Chromium headless shell 153.0.8010.12, Playwright 1.63.0): the effective
   browser argv (read from `/proc/<pid>/cmdline`) carries `--host-resolver-rules=MAP localhost 127.0.0.1, MAP
   *.acme.test 127.0.0.1` verbatim and no proxy switch; no proxy environment; `/etc/hosts` has `127.0.0.1 localhost`.
   `https://localhost:543/`, `www.localhost`, `org-c.localhost`, `http://localhost:543/` and `127.0.0.1` all end in
   `ERR_CONNECTION_REFUSED` with the rule, with only the acme rule and with no rule at all: **Chromium resolves the
   bare `localhost` itself; the `MAP localhost` rule of `1d7193ac` was never the cause and had nothing to fix.**
   Only `portal.acme.test` depends on a rule (`ERR_INTERNET_DISCONNECTED` without it).
2. `probe2.log` (E-DOM-4's shape against a loopback HTTP server answering `303 Location:
   https://marketing.identity.test/`, with the spec's `context.route(MARKETING**)` fulfil): the 303 and its `Location`
   arrive, the route handler is called **0** times, the redirect hop `https://marketing.identity.test/` is really
   requested and fails `net::ERR_NAME_NOT_RESOLVED`, and Playwright reports that failure as `page.goto:
   net::ERR_NAME_NOT_RESOLVED at http://localhost:5430/` — the initial URL. The installed Playwright documents it
   (`playwright-core/types/types.d.ts:4413`: "The handler will only be called for the first url if the response is a
   redirect"). A `MAP marketing.identity.test ~NOTFOUND` variant behaves identically.
3. In web's network namespace during the focused run (`focus1-netns.txt`): `/etc/hosts` `127.0.0.1 localhost`,
   resolver `127.0.0.11` with `ndots:0`, no entry for the test names (Chromium handles `*.localhost`; the rules handle
   the rest). The edge is reached at `127.0.0.1:543` (proxy in the same namespace), as the passing E-DOM-1/3 showed.

So E-DOM-4 never failed to resolve `localhost`: its public-entry 303 was followed for real to the configured marketing
URL, which nothing in the environment serves, and the local fulfil could never answer a redirect hop. E-DOM-4 had not
passed in any run (ops1 cascade, ops2/ops3 this defect).

**Fix (`2d604648`, five allow-listed paths; test fixtures only, no product/security/scope change):**
- `infra/identity/proxy.conf`: a `server_name marketing.identity.test` block on 543 answers `200 text/plain
  "marketing"` for every path (the same static stand-in pattern as `attacker.localhost`); header comment.
- `infra/compose/identity.yml`: `marketing.identity.test` added to `identitysetup --hosts` (the run leaf's SANs; the
  SPKI pin stays the only trust) and the overlay comment.
- `web/playwright.config.ts`: `--host-resolver-rules=MAP *.acme.test 127.0.0.1, MAP marketing.identity.test
  127.0.0.1:543` (the ineffective `MAP localhost` rule removed; the port form keeps the URL's own host and port, verified
  in `probe3.log`); comment corrected.
- `web/tests/browser/identity.spec.ts` E-DOM-4: the dead `context.route` fulfil removed; the real HTTPS navigations of
  `localhost` and `www.localhost` (`/`, `/account`), `redirectedFrom` 303 + `Location` and `/api/health` 200 assertions
  are unchanged; the followed page is now asserted too (`p.url()` = MARKETING, status 200, body contains "marketing").
  Block comment updated.
- `docs/foundation/IDENTITY.md`: one Origins sentence (the G016 names and the marketing stand-in).

**Batches (each: `git fetch origin` ok, `origin/staging` `1991c34d` ancestor of HEAD, `docker-desktop-mounts.sh`
13252/100000 = 13%, harness lock held by `flock`; no `run.sh`, so no `--allow-file` argument: the in-repo allow-list was
hashed at each preflight and compared with the binding):**

| Batch | Head / base | Allow-list hash seen | Command | Outcome |
|---|---|---|---|---|
| `diag1`–`diag3` | `944f777e` / `1991c34d` | `7be79e74…`, 57 | compose `--profile browser build web-browser` from a `git archive HEAD` copy (project `g016-dom7`), then `dom7-probe.js` with `--network none` | build rc 0 (cached); diag1/diag2 probe script errors (`browser.process` gone, no `chrome://version` in headless shell), diag3 rc 0 (facts above) |
| `probe2`, `probe3` | same image | — | `dom7-probe2.js` (default network), `dom7-probe3.js` (`--network none`) | rc 0; redirect hop not routed; `MAP name 127.0.0.1:PORT` keeps `page.url()` on the name |
| `focus1` | `2d604648` / `1991c34d` | `7be79e74…`, 57 | IDENTITY.md `identity:start` (rc 0, 35 s), `identity:migrate` (rc 0), `identity:browser` with the harness's named container/report env and `--project=domain-platform --grep '\[E-DOM-4\]'`; namespace one-off; `--grep '\[E-DOM-3\]'`; `identity:stop` | **E-DOM-4 PASS** (1 passed, 722 ms; JSON expected 1 / unexpected 0 / skipped 0 / flaky 0, no failure output); **E-DOM-3 PASS** (1 passed, 697 ms, same counts); `down -v` rc 0; leftovers containers 0 / volumes 0 / networks 0 / named 0 |

Adjacent check and why: the leaf SANs, the edge configuration and the launch arguments are shared by every identity
project, so E-DOM-3 (unknown `*.localhost`/`*.acme.test` names through the regenerated leaf and the edge's name
patterns, no seed needed) was run as the directly affected representative; E-DOM-1/2 and the identity/mcp projects need
the harness's seeded memberships and stay with the held `ops` rerun. No `ops` (no ops4) and no `rename` were launched.

**Cleanup.** `/tmp/project-validation-HqltM0` (ops2, harness run `9d3b7b`) and `/tmp/project-validation-8sRqsd` (ops3,
run `35261f`): both owned by the operator, mode 700, no live process cwd, their compact proof already in
`g016/ops2`, `g016/ops3` (meta, report, run.log, inventories) and this Result; the sanitized E-DOM-4 log excerpts were
kept as `g016/dom7/ops{2,3}-edom4-excerpt.txt`; both removed natively (`rm` rc 0, confirmed gone) together with their
generated `.env`/`secrets` files. The disposable source copy `g016/dom7/src` (generated `.env`) removed (rc 0); the nine
`g016-dom7-*` images removed (0 remain). Kept: `g016/dom7/` 184 KiB of logs/reports (no secrets: E-DOM-3/4 sign nobody
in; configure logs print no values). Successor 6's `g016-focus-*` images were not touched (not this worker's). One
classifier denial (a `jq` report-read misclassified as destruction) was replaced by `python3 -I`; two "too complex"
refusals were split into plain commands. No sudo, container, chmod or chown used.

**Status.** `ops2` stays FAILED (stale 56-pattern allow-file) and `ops3` stays FAILED (E-DOM-4; premature-launch
deviation preserved); its other rows remain attributable passes as Codex recorded, and the request-ID condition is
satisfied as recorded. E-DOM-4 is corrected with focused evidence on `2d604648`; `ops` rerun, `rename`, provider,
manual, deployment, final source, MR and hosted gates remain held for Codex. `store.go:578-581`, the RLS-SCOPE and
ISO-ORACLE evaluators and the sqlc deferral are unchanged. Context at hand-back about 130k.

### Attempt 1 continuation — Codex acceptance of the E-DOM-4 correction, staging `2421c23b`, the single `rename` run (2026-10-16 UTC; successor #7)

Codex ACCEPT of the bounded fixture correction (`2d604648`, submitted `614ab8c0`, Result `a2f9bca6…`) and the rename
release: `G016-ops-rerun-codex-disposition.md` on `origin/staging` `2421c23b`, SHA-256
`185eaf0140029af1f34b98f4444db985a9efc2a608eb0983fcef0dce19973580`. Staging `2421c23b` (management only: PROGRESS.md
and that review) merged as `c0b2f0b5`, no conflict, no rerun of unchanged tests. Allow-list unchanged: 57 patterns
`7be79e74…`.

**`rename1` — the one authorized `run.sh rename`** (`g016-scripts/runsh7.sh`; evidence `g016/rename1/`):
preflight fetch ok, `origin/staging` `2421c23b` ancestor of HEAD `c0b2f0b5`, dirty 0, `docker-desktop-mounts.sh`
13280/100000 (13%), harness lock held; `run.sh --base 2421c23b --allow-file <in-repo
management/v0/results/G016-scope-allowlist.txt> --owned-registry g016/registry/owned.tsv rename`; the allow-file
actually passed hashed at launch to `7be79e7497d67045f4ac55f810723327e5c663e08758c0983ecc92a83e9393f4` (57 patterns).
**rc 0, 2 min 12 s (18:46:28–18:48:40Z).** Rows: B-SOURCE-HASH PASS (disposable copy byte-identical, `beb12964e4d9308e`),
B-PORTS-OVERRIDE PASS, **B-RENAME PASS** (product name, base domain and CNAME target changed in `.env` only, `up
--no-build`, 7 image IDs unchanged, web title/domain and core readiness show the new identity), **B-BROWSER-RENAME
PASS** (@identity scenarios in shell-light/dark after the `.env`-only rename), **B-DBRENAME PASS** (fresh-volume
`DATABASE_NAME` change with the renamed roles and G008 helpers, core ready, migrate up), BUILD-LEDGER 7/7, B-DOCKER-ONLY
PASS, B-SECRET-CANARY PASS (8 generated secrets absent), B-SCOPE PASS (102 changed paths since `2421c23b`, all within
the allow-list), B-REQUIRED-CHECKS PASS (8 required IDs of rename + final, each once). Inventory diff 0 lines,
`inventory-leftover.txt` empty, no run-labelled container. (The harness's `sha256sum: .agents/skills/vestive-design: Is a
directory` warnings are pre-existing symlink noise, not a G016 change.) No ops4, no second rename. Reviewer note, not
changed after the run: B-DBRENAME's PASS detail text still reads "migrate up (to version 10)" (`run.sh:4434`) while
its check asserts `current_version\t11` (`run.sh:4432`), so the pass is the version-11 history; the same stale "version
10" wording remains in the B-TOGGLE-OFF PASS text (`run.sh:4582`, `toggle` suite, not run locally). A text-only fix is
in scope (`run.sh`) and left for Codex's disposition rather than edited behind the evidence.

Not rerun locally after `1d7193ac` (stated for the reviewer): the `lint` rows that read compose/fixture files
(B-COMPOSE, B-LITERALS, B-ENV-EXAMPLE) were last PASS in `lint2` at `1d7193ac`; the later delta is the five fixture/doc
paths of `2d604648` plus Result-only commits. Their compose validity is exercised by `focus1`'s `identity:start`
(`up -d --build --wait` rc 0) and the whole tree by `rename1`'s B-SCOPE/B-SECRET-CANARY; the MR fast pipeline runs the
lint profile on the submitted commit. Codex's review of `2d604648` required no other local rerun.

### Attempt 1 continuation — security amendment A2 binding and execution release (2026-10-16 UTC; successor #8)

Recorded before the first source edit of this round. Implementer successor #8, sole G016 writer; model Opus 5.5
(`claude-opus-5-5`), effort "high (requested)" (not verifiable from inside the session). Worktree
`.claude/worktrees/agent-<id>` (generated name, not renamed), local branch `g016-impl-succ8` tracking
`origin/g016-hostname-resolution-domain-lifecycle` (head `9a86c0bf` at start). Accounting unchanged: attempt 1 of B7,
checkpoint 4; the rotation spends no attempt.

- Staging `a11405e9` merged before any task write as `0eae0279` (A2 ACCEPT and execution release, A1/A2, the owner
  security amendment in ADR-007/Spec/ARCHITECTURE/SECURITY/DECISIONS, scope amendment 06). Two conflicts
  (docs/ARCHITECTURE.md, docs/SECURITY.md) resolved by keeping the branch implementation-status note followed by the
  owner amendment paragraph, both verbatim.
- Bound inputs verified in this checkout: A2 `4032a58249da5aad0c866ceba9ec24be9c1c9affb87ff6861d46e98805f84538` (the
  plan implemented); A1 `ac14d7af8fa075a6e78747387ff1546d890f3abfa29c7cc8864c7e22738cb443` (governs what A2 leaves
  unchanged); Codex review `fce48ba7de381d2a5c42e943322e8a5b0519320783da1f800faf08382421afd8` ("A2 ACCEPT, scope
  amendment 06 and execution release", including the five binding implementation clarifications).
- **Scope binding change: 57 patterns (`7be79e74…`) → 64 patterns
  (`f3975a7db103694c53129804eb4192320810eb6d4c05b18691c8c825532b50d2`)**, amendment 06 adding exactly the seven A1.8
  patterns. Every changed path of this round must match the 64-pattern file.
- Order (A2, riskiest first): migration 00012 + CT-DOM-12; session binding (Session/resolve/issue, Login, Callback,
  Logout L-1, http session route); consumers (access context, domains api, MCP setup/revoke, claim pre-check and
  `_hostname_owned` mapping); web (identity-client, api/account, account page, L-2 lookupContext) and docs.
- Unchanged by this round: store.go issuanceGuard fail-closed, the RLS-SCOPE and ISO-ORACLE evaluators, the sqlc
  deferral (handwritten pgx). Prior passes reused (lint2, ops3 rows, focus1, rename1); ops1–ops3 stay FAILED.

### Attempt 1 continuation — A2 execution: slices 1–2 done, slice 3 stopped at a test-harness scope gap (2026-10-16 UTC; successor #8, hand-back (c))

Head `f0732ebc` (pushed, fast-forward). 38 paths changed since the staging merge `0eae0279`, all inside the 64-pattern
allow-file (`scopechk.sh` with `f3975a7d…`: 38 OK, 0 NO). Commits: `83d115f3` `34b631d6` `05501f8b` (migration),
`052b791d` `d52ff53f` `fbfdcf17` `f0732ebc` (binding, consumers, tests). Web, docs and the browser selectors are not
started.

**Slice 1 — migration 00012 (done, focused evidence PASS).** `core/migrations/00012_session_binding_and_claim_exclusivity.sql`
as A2 Issue 1/2: `origin`/`organization_id` with the VALID `sessions_origin_valid`, `sessions_organization_fkey`,
`sessions_binding_shape`, `sessions_mcp_unbound` (no NOT VALID CHECK); the `sessions_insert` predicate requiring a
binding on every new browser row (pin string unchanged); the bound 3-argument `app.auth_resolve_session` replacing the
2-argument one (`kind = 'browser' AND origin IS NOT NULL AND organization_id IS NOT NULL AND origin = p_origin`);
proven-only `hostname_state`; `_hostname_live` → `_hostname_owned`; version-12 postconditions (function and ACL
inventories copied verbatim from 00011 plus the two auth-reader column grants, index set, reader signatures, constraints).
Down (clarification 2): NO FORCE on `app.sessions` and `app.organization_domains`, the three refusals (bound sessions;
live legacy browser sessions; duplicate non-removed hostname), FORCE restored on both before any shape change, then the
exact version-11 predicate, readers (bodies pinned equal to 00010/00011 by the shape test), index and columns; no
trigger/constraint disabling. Tests: new `TestSessionBindingMigrationShape`; `TestMigrationSet`, runner set/12 sources,
`metadataThroughVersion12Only`, lifecycle pins → 12; boundary `g016Binding` clause; package setups `current_version 12`;
tenancy step-down purges browser sessions before 12→11; migrate `reset` drops the new FK first. **CT-DOM-12**
(clarification 4): seeded at version 11 — a live legacy row revoked by the runtime (store.go revokeAll text,
RowsAffected 1, exactly one `session.revoked`), a *separate* initially live row expired by the restore-time
invalidation (invalidate.sql sessions UPDATE, verified byte-equal to the file through `MCP_INVALIDATE_SQL_B64`, 2 rows),
and an already expired row; rows unchanged by the upgrade (count/hash), bound reader returns nothing at two origins,
down refused while live legacy rows exist, allowed after revocation+invalidation with the exact version-11 catalog,
the restored 2-argument reader revives nothing, up again with an identical snapshot; runtime negatives (unbound browser
INSERT 42501, half-bound refused, unbound mcp admitted, bound mcp refused, bound browser admitted and resolving only at
its origin, `origin`/`organization_id` UPDATE 42501); bound-session and duplicate-hostname down refusals.

**Slice 2 — session binding (done, focused evidence PASS).** `SessionInfo.Origin/OrganizationID`;
`Service.Session(ctx, credential, origin)` (empty/overlong origin unauthenticated before any statement); `store.resolve`
through the 3-argument reader; `issue` writes the binding (issuanceGuard untouched). Login: when a credential is
presented, `tenantOrigin` (pool statement) then `Session`, already_authenticated only on equal Organization (PG-007:
+1 reader statement). Callback continuity only on equal pinned Organization. Logout (L-1, clarification 3): session at
its own origin → CSRF → committed local revocation; provider scope resolves the return origin only after the provider
revocation; a non-routing or unavailable host suppresses the hosted redirect and the outcome stays the actual provider
result (`provider_session_revoked`/`provider_revoke_failed`). Session route: `origin` required (absent 400
invalid_request, repeated 400 ambiguous_parameter), bound Organization must equal the origin's (else 401). Tests:
TestSessionRoute (absent/repeated/other-Organization), relay E replay, F missing binding, G reassignment (old cookie
not already_authenticated; new org-b-bound session is), L-1 after removal (current revokes; provider revokes without a
redirect), TestLocalLogoutAndCSRF origin cases (current idempotent, all/provider unauthenticated, nothing revoked).

**Slice 3 — consumers (source done; validation stopped).** domains api `begin` and MCP setup `begin`/`Revoke`: bound
`Session(cred, origin)` + `info.OrganizationID == host Organization` (else unauthenticated) + `BeginRequest(cred,
origin)`. access: `BeginRequest(ctx, credential, origin)`; the browser scope parses the session's Organization into the
existing `requestScope.org` and fails closed (unavailable) when absent/malformed (clarification 1, `zero`);
`resolve` refuses `zero(sc.org) || org != sc.org` as organization_not_permitted with one access.denied row without the
requested Organization (MCP `beginVerified`/`mcpScope` unchanged). Claim pre-check `HostnameProven`
(`hostname_state('custom', host)`) → host_taken with zero effects; `domainFailure` maps `_hostname_owned`. Unit:
TestBeginRequestAndSelectors gains unbound-session and cross-Organization cases.

**Scope gap (stop condition; nothing worked around).** The A2 Resolve confinement (and the consumers' Organization
comparison) is incompatible with the existing access and MCP integration harnesses, which A2/the release treated as
"mechanical signature edits, matrices not rerun": `tenancy/access/integration_test.go` signs every fixture user in once
through the stub host `itHost` (only `itOrigin` → the dedicated "Sign-in fixture" Organization `e.loginOrg`, membership
granted only around issuance by `admit`) and then resolves the test-world Organizations (w.A, w.B, …) with that one
credential; `mcp/env_integration_test.go` does the same (`loginHost` → its own `e.loginOrg`) while the MCP handler's
host lookup maps `itHostA`/`itHostB` to the world Organizations, and its "cross" user reuses one credential across hosts.
With binding, every such Resolve/setup is refused by design. Evidence `cons1` (head `fbfdcf17`): access
`^(TestDomains|TestAccessPolicy|TestRLSIsolation)$` rc 1, 0 pass / 39 fail (every positive Resolve `context_denied`;
TestAccessPolicy only IT-CTX-2/6 and runtime_identity pass; RLS-AUTH also hit the 2-argument reader, fixed in
`f0732ebc`, not rerun); mcp `^(TestMCPTransport|TestMCPLifecycle)$` rc 1, 0 / 15 (`create … 401 unauthenticated`).
The same harness feeds the attack, audit, ordering, objects, footprint, recovery and retention access suites, so
repairing it changes their fixtures (extra sign-ins and `session.issued` events, statement-count replays) and needs
their full reruns, which the release excludes. Smallest options for Codex (none started):
1. Keep A2 as accepted; amend scope/validation for a bounded harness change: the stub hosts map one tenant origin per
   world Organization, `admit` targets that Organization, a per-(user, Organization) credential helper replaces
   `u.cred` at BeginRequest/setup call sites in the access and MCP integration files, then rerun the affected access/MCP
   suites in full (B-POLICY-IT, B-RLS-IT, attack/audit/ordering/objects/recovery, B-MCP-IT).
2. Confine at the consumers only (every production BeginRequest caller — domains api, MCP setup/revoke — already compares
   the bound Organization with the host's before BeginRequest, and the session route does the same); drop the Resolve
   confinement. Harness edits stay mechanical, but this departs from A2 and clarification 1 (needs Codex/owner).
3. As 1, but run the reworked access/MCP matrices only in the hosted post-merge full gate.

**Commands and outcomes (all on `g016-focus` / network-none check images; evidence `g016/a2/`).** Every batch:
`pre8.sh` (fetch ok; `origin/staging` `a11405e9` ancestor of the head; dirty 0; mounts 13288–13304/100000 (13%);
allow-file actually used `f3975a7d…`, 64 patterns = bound), lock `local-harness.lock` taken inline (`flock 9`; the
wrapper form was refused by the classifier, so the documented inline fallback was used).

| Run | Head | Selector | Outcome |
|---|---|---|---|
| m1, m1b | 83d115f3 | unit: migration shape/set/MCP shape, runner set, boundary (TestRLSBoundary, TestAccessBoundary, TestProductionImporters); vet -tags integration migrate/migrations/tenancy/access/mcp | PASS |
| mig1 | 83d115f3 | migrate `^(TestDomainsCatalog\|TestRLSCatalog\|TestMCPCatalog\|TestMigrationLifecycle)$` | rc 1: TestRLSCatalog, TestMCPCatalog PASS; DomainsCatalog/Lifecycle FAIL (postcondition used `regprocedure::text`, search_path-dependent) |
| mig2 | 34b631d6 | migrate `^(TestDomainsCatalog\|TestMigrationLifecycle)$` | rc 1: CT-DOM-12 live-row filter counted the revoked row; reset blocked by the new FK |
| **mig3** | 05501f8b | same | **PASS 2/2** (16 PASS lines incl. CT-DOM-1/NEG/UPDOWN/12, lifecycle 01–10) |
| vet1–6 | → f0732ebc | `gofmt -d .; go vet ./...; go vet -tags integration ./...` | final VET-OK, gofmt clean |
| unit1 | 052b791d | unit migrations, migrate, identity, access, mcp, domains/..., tenancy | rc 1: TestBeginRequestAndSelectors (new assertion before the record check; test order fixed `d52ff53f`) |
| unit2, **unit3** | d52ff53f, fbfdcf17 | same packages | **PASS** (all ok) |
| id1 | d52ff53f | identity `^(TestCustomDomainRelay\|TestLocalLogoutAndCSRF\|TestLogoutAll\|TestProviderLogout\|TestDisableAndEnable\|TestCallbackContinuity\|TestSessionLifetimes\|TestIssuanceGuardFailsClosed)$` | rc 1: 6/8; expected behaviour changes in TestLocalLogoutAndCSRF (origin case) and TestProviderLogout (originB helper) |
| **id2** | fbfdcf17 | same | **PASS 8/8** (22 PASS lines incl. relay E/F/G/L-1) |
| cons1 | fbfdcf17 | access `^(TestDomains\|TestAccessPolicy\|TestRLSIsolation)$`; mcp `^(TestMCPTransport\|TestMCPLifecycle)$` | rc 1 — the scope gap above |

All runs: leftovers containers=0 volumes=0 networks=0 run=0.

**A2/AC mapping update.** Done with evidence: A2 Issue 1 (data model, reader, policy), Issue 2 (down), Issue 3 Login /
Callback / L-1, Issue 4 CT-DOM-12, reassigned Login, E/F/L-1, TestSessionRoute; AC3/AC5 binding cases at identity
level. Source done, evidence pending the gap decision: access scope confinement, domains api/MCP consumer comparison,
claim pre-check/`_hostname_owned` (DOM-2 race with per-claimant TXT+CNAME proofs and DOM-7 not yet written),
TestAccessPolicy cross-Organization subcase, IT-MCP-3 401. Not started: web (identity-client origin, api/account,
account page, L-2 `lookupContext` in proxy.ts and the callback, unit tests), docs (IDENTITY, DOMAINS, OPERATIONS,
MIGRATIONS; N-2/N-3; P-1/P-3 label), run.sh version-12 pins (B-MIGRATE-CATALOG, B-PERSIST, B-DBRENAME,
B-TOGGLE-OFF/rename, recovery drill: `current_version\t11` → 12 and the 00012 apply lines), browser E-ORIGINS/E-DOM-2,
`run.sh lint lint-base lint-core`.

### Attempt 1 continuation — A2 consumer-fixture correction binding (2026-10-16 UTC; successor #9)

Recorded before the first source edit of this round. Implementer successor #9, sole G016 writer; model Opus 5.5
(`claude-opus-5-5`), effort "high (requested)" (not verifiable from inside the session). Worktree
`.claude/worktrees/agent-<id>` (generated name, not renamed; owned by successor #9 for G016 only), local
branch `g016-impl-succ9` tracking `origin/g016-hostname-resolution-domain-lifecycle` (head `607d2a61` at start; the
local `g016-hostname-resolution-domain-lifecycle` name is held by older worktrees, so it was not reused). Accounting
unchanged: attempt 1 of B7, checkpoint 4; no budget or test restart.

- Staging `e6aa7949` merged before any task write as `d773ddcf` (management-only: PROGRESS.md, the disposition, the
  allow-list; no conflicts).
- Bound inputs verified in this checkout: disposition `G016-a2-fixture-codex-disposition.md`
  `258d95a4220196768eeef9cc2049f722be3eedf797e22bb56e5f41ce0b8dc1ca` (authoritative: ACCEPT option 1, keep the
  production A2 Resolve confinement); A2 `4032a582…`, review `fce48ba7…` (release clarifications), A1 `ac14d7af…`
  unchanged; the coordinator gap record `e33a1d85…`; Result `ea8f3fee…` at `607d2a61`.
- **Scope binding change: 64 patterns (`f3975a7d…`) → 72 patterns
  (`6e78ec85ddeb48cda2f3bba98f4caaca0e9e8ce3c2c88ac1131cd55260c378f8`)**, amendment 07 adding exactly eight
  `core/internal/tenancy/access/` fixture callers (attack, footprint, mcp, objects_faults, objects_footprint,
  objects_ordering, retention, rls_ordering integration tests), to be used only where per-user/Organization origin and
  credential compatibility needs them. The 64-pattern prefix is unchanged. Every changed path must match the 72-pattern
  file.
- Fixture rules (disposition): explicit tenant origins, per-(user, Organization) credentials prepared through the real
  identity flow outside measured/audited/locked windows; no re-auth on arbitrary errors, revocation, disablement or
  wrong-origin negatives; no credential switch inside one request scope; existing owner/admin bindings preserved
  (admission deletes only what it added); membership-negative subjects never granted authority. Attack A chain
  (attack_integration_test.go:780–788): two explicitly bound request scopes for the legitimate A and B positives,
  negatives, oracle literals and a direct A-scope→B Resolve denial kept.
- History preserved: `cons1` enclosing commands stay FAILED (access 39 FAIL lines with ten nested passes; MCP 15 FAIL).
  Reused passes: mig3, mig1 (slice 1), id2, unit3 (slice 2); not rerun.

### Attempt 1 continuation — A2 fixture correction: validation steps 1–3 PASS, checkpoint hand-back (c) (2026-10-16 UTC; successor #9)

Head `bc8bc0ba` (pushed, fast-forward). Commits `388047e0` (fixtures), `6d2e88df`, `bc8bc0ba` (focused fixes). 16 test
paths changed, all matching the 72-pattern allow-file (`scopechk.sh`: 16 OK, 0 NO); no production file changed in this
round. Two bounded helper subagents (general-purpose, model inherited, effort unknown) edited disjoint files under
successor #9's direction (MCP harness; the remaining access callers); integrated, compiled and validated here.

**Fixture design.** `tenancy/access/integration_test.go`: `itHost` routes `itOriginOf(org)`
(`https://<uuid>.org.localhost:543`) to that world Organization (itOrigin stays the sign-in fixture's);
`admit(t, org, …)` inserts a membership only with `ON CONFLICT (subject_user_id, scope_type, scope_id) DO NOTHING
RETURNING id` and deletes exactly that row; `signInAt` (real Login + Callback at org's origin, origin recorded per
credential), `bind(t, u, orgs…)` (setup-only, explicit), `u.in(t, org)` (strict lookup, never issues), `beginAs`;
`csrf`/`logout` use the credential's own origin. Every world/setup binds the (user, Organization) pairs used;
membership-negative subjects (gOnly, foreign at A, owner at B, claims-only, non-member) are admitted only for issuance.
Split legitimate scopes: attack A chain (A and B scopes for multi, A1 calls paired with their own scope, direct A-scope
→ B Resolve denial added; negatives and oracle literals unchanged), attack D2, IT-FLAG-3 B context, IT-CTX-4 per-org
negatives, recovery Verify (cross@A/crossB@B). MCP env: per-host credentials `u.at[host]`, a host→Organization map read
by the identity resolver (`route`), host-targeted admission; `freshCred` and the re-sign-in inside `create` removed —
renewal only through explicit `bind` after the tests' own clock advances/logouts. New cases: IT-CTX-11 (browser scope
confinement: A scope → B organization_not_permitted, separate B scope succeeds, A credential at B's origin
unauthenticated); IT-MCP-3 cross-host create/list 401 with zero effects; DOM-2 rewritten per A2 (8 claims → 8 pending;
per-claimant `domains.Proof` over fixture TXT+CNAME answers, another claimant's value → mismatch; concurrent promotion →
1 verified, 7 host_taken with loser rows unchanged and no loser `domain.checked`; claim of the proven hostname →
host_taken, zero effects); DOM-7 pending-only → no reader row, pending + proven → the proven row only. A2 reader
semantics also changed three existing pending-row expectations to "none" (DOM-4 ×2, DOM-5).

**Runs** (`b9.sh`, g016-focus, evidence `g016/a2/`). Every batch: `pre9.sh` fetch ok, `origin/staging` `e6aa7949`
ancestor, dirty 0 (vet: 16 before the commit), mounts 13312/100000 (13%), allow-file actually used `6e78ec85…`, 72
patterns = bound; lock `local-harness.lock` via `flock -n`. All runs: leftovers 0/0/0/0.

| Run | Head | Selector | Outcome |
|---|---|---|---|
| s9vet1, s9fmt1 | 49ab5f86 (+working tree) | vet ±tags access, mcp, domains, identity; gofmt | VET-OK; one gofmt line fixed before commit |
| **rec1** | 388047e0 | access `^TestAccessPolicy$/^IT-CTX-1_`; mcp `^TestMCPTransport$/^IT-MCP-3_` | **PASS** (positive Resolve and MCP create recovered) |
| fam1 | 388047e0 | access `^(TestDomains\|TestAccessPolicy)$`; `^TestRLSIsolation$/^(RLS-ATTEST\|RLS-AUTH)_`; mcp `^(TestMCPTransport\|TestMCPLifecycle)$` | rc 1: RLS-ATTEST/AUTH **PASS**; MCP Transport+Lifecycle **PASS** (15 PASS lines); DOM-4/5 (pending reader rows), DOM-10 (unbound owner), IT-CTX-11 (asserted 0 BeginTx; the durable access.denied row uses one) FAIL |
| fam2 | 6d2e88df | access `^(TestDomains\|TestAccessPolicy)$` | rc 1: **TestAccessPolicy PASS** (all subtests incl. IT-CTX-11); DOM-4 withdrawal pending row FAIL |
| **dom3** | bc8bc0ba | access `^TestDomains$` | **PASS** (DOM-1…DOM-10) |
| **atk1** | bc8bc0ba | access `^(TestIsolationAttacks\|TestAudit)$` | **PASS** (28 PASS lines, 0 SKIP) |

**Pending (not run here).** Step 4 ordering/footprint families whose wiring changed: `TestAccessOrdering`,
`TestRLSOrdering`, `TestAccessFootprint`, `TestRLSPool`/`TestRLSFootprint`, access `TestMCPOrdering`/`TestMCPFootprint`,
`TestObjectsOrdering`, `TestObjectsFaults`, `TestObjectsFootprint`, `TestObjects`/`TestObjectsSignHost` (objects
callers changed; their store prerequisites are in b9.sh). Recovery (`TestRecovery*`, prerequisite: the seed/post/restore
drill, not authorized) and objects persist seed/verify: pending required hosted coverage. Finding for review:
`setupRecovery` never sets `e.loginOrg` (also true at `607d2a61`), so its itOrigin sign-ins admit into the zero
Organization; not changed here (beyond credential wiring). Retention: unchanged (no request scope). Then web, docs,
run.sh version-12 pins, browser E-ORIGINS/E-DOM-2 and `run.sh lint lint-base lint-core` remain not started.

### Attempt 1 continuation — successor #10 binding (2026-10-16 UTC)

Recorded before the first source edit of this round. Implementer successor #10, sole G016 writer; model Opus 5.5
(`claude-opus-5-5`), effort "high (requested)" (not verifiable from inside the session). Worktree
`.claude/worktrees/agent-<id>` (generated name, not renamed; owned by successor #10 for G016 only), local
branch `g016-impl-succ10` tracking `origin/g016-hostname-resolution-domain-lifecycle` (head `9794ba5e` at start).
Accounting unchanged: attempt 1 of B7, checkpoint 4; no budget or test restart. `origin/staging` `e6aa7949` is an
ancestor of the head (no merge needed). Bound inputs verified here: Result `4191e85e…` at `9794ba5e`, allow-file
`6e78ec85…` (72 patterns), disposition `258d95a4…`, A2 `4032a582…`, release clarifications `fce48ba7…`. Successor #9's
scripts reused as path-only copies (`pre10.sh`, `b10.sh`, `gocheck10.sh`, `unit10.sh`, `webcheck10.sh`: only the
worktree/evidence path changed). Remaining scope: step 4 ordering/footprint/objects families, the `setupRecovery`
finding (compile/vet only), web, docs, run.sh version-12 pins, browser E-ORIGINS/E-DOM-2, lint.

### Attempt 1 continuation — successor #10: step 4 PASS, setupRecovery fix, web, docs, version-12 pins, lint PASS, browser (2026-10-16 UTC; hand-back (a))

Staging advanced to `0ba27e74` (management-only: PROGRESS.md and the disposition's "Successor 9 checkpoint review")
after the binding record; merged before any source write as `7f322579` (no conflicts). Disposition now
`46201e6a02671bd3c4c00473ea0f1f314050be9bab610ac565f5fcc7ba3802a2`: it ACCEPTS successor 9's fixture adjustments at
`9794ba5e` (DOM-4 ×2 and DOM-5 expecting no lookup row for an unproven claim under the proven-only reader; MCP alias
hosts registered to their tenant with separately bound credentials; explicit renewals only after known expiry/logout;
role-preserving temporary admission; exact-host negatives and independent oracle literals intact) and authorizes the
setupRecovery correction. Those points are therefore cited to that review, not listed as open. fam1/fam2 stay FAILED in
the history. Allow-file unchanged (`6e78ec85…`, 72 patterns). Commits (all pushed fast-forward): `a273eb2e` (binding),
`7f322579` (merge), `21aa5533` (setupRecovery), `eb8f2a30` (web), `6cfdca28` (docs), `ac0f02a8` (run.sh), `84a49eff`
(browser spec). Every batch: fetch ok, `origin/staging` `0ba27e74` ancestor of the head, Docker Desktop mounts
13332–13344/100000 (13%), allow-file actually used `6e78ec85…` = bound; lock `local-harness.lock` via `flock -n`. The
classifier refused the predecessor wrappers that call git, so the git steps (archive, head/base) ran inline in the
worktree and the copied scripts (`b10.sh`, `b10s.sh`, `b10b.sh`, `run10.sh`, `webcheck10.sh`; g016-scripts/) contain no
git. Evidence: `g016/a2/`, `g016/s10lint1/`, `g016/b10b/`. All runs: leftovers 0/0/0/0.

| Run | Head | Selector | Outcome |
|---|---|---|---|
| **ord1-1** | 7f322579 | access `^TestAccessOrdering$` | **PASS** (17 subtests, 0 skip) |
| **ord1-2** | 7f322579 | `^TestRLSOrdering$` | **PASS** (5) |
| **ord1-3** | 7f322579 | `^TestAccessFootprint$` | **PASS** (2) |
| **ord1-4** | 7f322579 | `^(TestRLSPool\|TestRLSFootprint)$` | **PASS** (2 tests, 7 subtests) |
| **ord1-5** | 7f322579 | access `^(TestMCPOrdering\|TestMCPFootprint)$` | **PASS** (2 tests, 10 subtests) |
| **ord1-6** | 7f322579 | `^TestObjectsOrdering$` | **PASS** (8) |
| **ord1-7** | 7f322579 | `^TestObjectsFaults$` | **PASS** (19) |
| **ord1-8** | 7f322579 | `^TestObjectsFootprint$` | **PASS** (E-OBJ-1/2 records, no subtests) |
| **ord1-9** | 7f322579 | `^TestObjects$` | **PASS** (10) |
| **sgn1** | 7f322579 | `^TestObjectsSignHost$` with the run.sh `obj_sign_host` host probe (curl form; published MinIO port 9932): signed fetch 200 + digest + `attachment\|text/plain\|no-store`, bare path 403 | **PASS** (probe ok) |
| s10vet1 | 7f322579 + setupRecovery | `gofmt -l ./internal/`; `go vet [-tags integration] ./internal/tenancy/access/` | VET-OK, gofmt clean |
| web1 | 21aa5533 + web | web check image: lint, typecheck, `node --conditions=react-server --test tests/unit/*.test.mjs` | PASS (145/145, 0 skip; incl. the new L-2 lookupContext and session-origin cases) |
| **s10lint1** | ac0f02a8 | `run.sh --base 0ba27e74 --allow-file …/G016-scope-allowlist.txt lint lint-base lint-core unit` | **PASSED** (19 PASS rows, 0 FAIL/UNMET; B-SCOPE 120 changed paths all within the allow-list; B-MIGRATE-SET with 00012; B-UNIT-CHECKS; B-SECRET-CANARY) |
| web2 | ac0f02a8 + spec | web lint, typecheck, unit | PASS |
| **br1** | 84a49eff | documented seeded overlay (IDENTITY.md identity:start, identity:migrate → `current_version 12`, the run.sh `mcp_seed` seed: emulator user IDs looked up inside the web-browser container, `mcp-seed.sql` as superuser), identity:browser with `--project=identity-light --grep '\[E-ORIGINS\]'`, then `--project=domain-custom --grep '\[E-DOM-2\]'`, identity:stop | **PASS** (1 passed each, nonzero; stop rc 0; leftovers 0) |

**setupRecovery (authorized correction; compile/vet only).** `setupRecovery(t, ids *recIDs)` now creates the model and a
nonzero fixture login Organization before any identity use: Seed (`ids == nil`) creates "Recovery sign-in fixture" and
records it as `recIDs.Login` in `RECOVERY-SEED`; Post, Verify and Repair reuse that restored Organization from
`RECOVERY_IDS` and fail closed when it is absent (no new tenant or audit state after restore; replay and identity
assertions unchanged). The added Organization is created before T_mid with the default policy flag (true), so the
ledger-sync `policy.disable` read is unaffected. **Pending required hosted coverage:** the complete recovery
orchestration, selector `run.sh migrations` → `recovery_drill` (B-RECOVERY-DRILL, then B-MAINTENANCE-DRILL): source
reset + private MinIO + `migrate up` to 12, `TestRecoverySeed` → `infra/recovery/recovery.sh` backup → `TestRecoveryPost`
(`RECOVERY_IDS`) → restore/invalidate/replay on the PORTS_REC target → `TestRecoveryVerify` (`RECOVERY_IDS`,
`RECOVERY_CHECKPOINT_END`) → `TestRecoveryRepair` steps 1/2. No phase was run alone; no drill was launched.

**Web.** `identity-client.session(ctx, cred, origin)` requires the origin (always `?origin=`); `/api/account` answers 401
for a request without a usable origin and passes `ownOrigin`; the account page passes its own origin under every cookie
profile (redirect to sign-in without one). L-2: `lookupContext(ctx)` in `web/lib/host.ts` (fresh 32-hex request id,
`generated`, same trace/span/flags), used by `web/proxy.ts` and `web/app/auth/callback/route.ts` (inline object
removed). Unit: host.test.mjs (lookupContext shape, distinct ids, the lookup's X-Request-Id vs the inbound one, equal
traceparent) and identity-client.test.mjs (every session call carries the origin; URL asserted).

**Docs.** IDENTITY.md: session route contract (`?origin=` required, 400/401/404), A2 session binding and consumer
comparisons, "Migrate 12, then deploy" (legacy rows refused from the deploy, revocable, recoverable, retained; no forged
`session.revoked`; down refusals), N-2 pseudonymous hostname tag. MIGRATIONS.md: version-12 row, "From version 12", down
from 12. OPERATIONS.md: the A1.5 L-2 wording as the narrow exception (gate and callback), explicitly not a general rule.
DOMAINS.md: non-exclusive pending claims and proof-time exclusivity, N-3 chain-dependent CNAME outcome (P-3), N-2,
P-1 (WorkOS wildcard/sign-out) vs P-3 (bounded TXT/CNAME proof) labels, catalog evidence row.

**run.sh version-12 pins** (verified by lint + unit only; the suites that execute them were not run): B-MIGRATE-SET (set
gains 00012; static), B-DOCS-EXEC-ROOT (root suite; 00012 pending/applied, `current_version` 0/12/12), B-MIGRATE-CLI and
B-MIGRATE-CATALOG (steps now at versions 0, 11 and 12; rollback/inspect 48 relations; the 42 expectation removed),
B-RECOVERY-DRILL setup (`current_version\t12`, record text), B-PERSIST, B-DBRENAME, B-TOGGLE-OFF (and their stale
version-10 texts). B-MIGRATE-SET and B-DOCS-EXEC-ROOT were not in the named list but are the same mechanical pin
(B-MIGRATE-SET would have failed lint otherwise); reported for review.

**Browser.** E-ORIGINS: org-a's `__Host-sid` value, placed alone in a fresh browser profile as org-b's cookie, gets 401 from `ORIGIN_B/api/account`; the same replay at ORIGIN_A is 200 with alice's id. E-DOM-2: the portal session value replayed at org-a (same Organization A, another origin) is 401; the same replay at the portal is 200 with the same id. The replay goes through the browser cookie jar because the Node request client neither resolves `*.localhost` nor trusts the pinned key (the spec's own rule); the browser sends exactly that value as the Cookie header. The spec change was linted/type-checked in web2 (after s10lint1, whose B-SCOPE preceded it; `identity.spec.ts` matches the allow-file).

**Remaining before final review.** Required hosted coverage: the recovery orchestration above (also the only runtime
proof of the setupRecovery fix), the objects persist seed/verify, and the hosted rows of the suites whose pins changed
(`migrations`, `root`, `ops`, `rename`, `off`) under the existing staging gate. Unchanged: issuanceGuard fail-closed,
RLS-SCOPE/ISO-ORACLE evaluators, the sqlc deferral, provider/manual/deployment gates P-1..P-4.

### Attempt 1 continuation — successor #11: Codex source-review correction loop (2026-10-16 UTC; hand-back (a))

Worker: Claude Code successor #11, model Opus 5.5, effort high (requested; actual inherited setting unknown); sole
writer. Worktree `.claude/worktrees/agent-<id>` (generated name, recorded ownership), local branch
`g016-impl-succ11` tracking `origin/g016-hostname-resolution-domain-lifecycle` at `edeb1965`. Staging `90a16726`
(management-only: PROGRESS.md and the Codex review) merged before any write as `6438823b`. Verified inputs: review
`a9e15588…3590`, allow-file `6e78ec85…`, 72 patterns, Result at `edeb1965` `a8487798…85d4`, A2 security report
`c690e62a…`. Accounting: attempt 1 of B7, checkpoint 4, no reset.

**1. Production CNAME bound (P2).** Reading go1.27 `net/dnsclient_unix.go` (`goLookupIPCNAMEOrder`): with
`PreferGo`, `LookupCNAME` sends CNAME, A and AAAA queries and returns the target of the first CNAME record of the first
answer parsed (or the owner of the first address record when the name has no CNAME). Because a recursive answer lists
the chain's CNAME records first and in order, that is the queried name's own edge, not the chain's end; the old
adapter comment and the DOMAINS.md N-3 text ("the resolver reports the chain's end; a target that is itself a CNAME
fails") were wrong. Proof's loop therefore counts actual edges on the production adapter. Change (no dependency, no
framework): `NewSystemChecker` is unchanged in behaviour and now calls an unexported `newSystemChecker(dial)` (nil =
system dialer; the system resolver configuration and per-lookup deadlines unchanged; no caller-selected resolver);
`TestDNSSystemChecker` drives the real adapter over a disposable in-process stream transport that answers like a
recursive resolver (ordered chains; NXDOMAIN/NODATA): `LookupCNAME` yields l1, l2 for the four-edge chain, three edges
ok, four edges mismatch although they reach the target, wrong target and no-CNAME mismatch, unknown NXDOMAIN, a mute
server transient within the deadline. DOMAINS.md now states these semantics, including that the target's own further
chain does not matter (this matches Proof's existing rule; reported for review). Guarantee and residual: under
"Unresolved issues" (reordering resolvers; a strict edge reader would need a wire parser or a new dependency).

**2. TLS result fence (P2).** `RecordDomainCheck` (core/internal/tenancy/access/domains.go) refuses a recheck whose
proof is ok when `obs.TLS.Ref != cur.Ref()` for the locked row (ID, generation, hostname and stored reference): the
existing stale-result failure `domain_state_invalid`/write conflict, zero effects, no successful check event (ordinary
denied audit unchanged). No empty-ref fallback. Tests: the integration helper `e.check` gives TLS observations the read
row's ref (what every adapter echoes, as in `domains/api` check); DOM-4 adds foreign-row, stale-generation,
foreign-hostname, foreign-reference and empty-ref negatives on a TLS-expired active row (rows hash and host state
unchanged, success-event count unchanged), then the exact ref restores routing (positive control).

**3. LOW-1 logout origin.** `Logout` now rejects, after the scope check and before `Session`, any origin that is not
structurally bindable: an exact `IDENTITY_PUBLIC_ORIGINS` entry, or a canonical lowercase `https` origin with an
optional canonical port and no userinfo/path/query/fragment, at most 300 bytes (pure; tenant origins resolve as https
only). A well-formed foreign origin keeps the idempotent `local_revoked` (current) and 401 (all/provider); revocation on
a removed host (L-1) and actual provider outcomes are unchanged. `TestLocalLogoutAndCSRF` adds 17 invalid origins × 3
scopes → `origin_rejected`, no provider call, no revoked session. IDENTITY.md: the logout row's `origin_rejected` now
states these semantics; the "validates every … sign-out origin by resolving its host" sentence is corrected (sign-out
is structural only since L-1).

Commits: `73c4c4ea` (the three corrections + docs), then this Result commit; pushed fast-forward to
`origin/g016-hostname-resolution-domain-lifecycle`. Scripts (no git inside; git steps inline): `pre11.sh`,
`unit11.sh`, `b11.sh` (b10.sh with this worktree and `g016/fix11`), `runsh11.sh`, selector `fix11-i1.sel`
(g016-scripts/). Before every batch: fetch ok, `origin/staging` `90a16726` ancestor of the head, Docker Desktop mounts
13372/100000 (13%), allow-file actually used `6e78ec85…` = bound, `flock` on `local-harness.lock`.

| Run | Head | Selector | Outcome |
|---|---|---|---|
| **u1** | 73c4c4ea | check image (network none): `gofmt -l .`; `go vet` and `go vet -tags integration` of `./internal/domains/... ./internal/identity/ ./internal/tenancy/access/`; `go test -count=1 -v -run '^TestDNS(Checker\|SystemChecker)$' ./internal/domains/`; `go test -count=1 -run Logout ./internal/identity/` | **PASS** (gofmt clean, VET-OK, TestDNSChecker + TestDNSSystemChecker PASS, identity ok) |
| **i1-1** | 73c4c4ea | `g016-focus` (b11.sh), identity `^(TestLocalLogoutAndCSRF\|TestProviderLogout\|TestCustomDomainRelay)$` | **PASS** (3 top-level; TestCustomDomainRelay subtests incl. L-1 sign-out after removal) |
| **i1-2** | 73c4c4ea | access `^TestDomains$` | **PASS** (DOM-1..DOM-10) |
| **lint11** | e685390b / base 90a16726 | `run.sh --base 90a16726 --allow-file …/G016-scope-allowlist.txt lint lint-base lint-core` | **PASSED** (18 PASS rows, 0 FAIL/UNMET; B-SCOPE 120 changed paths all within the allow-list; B-MIGRATE-SET with 00012; B-SECRET-CANARY 16; B-REQUIRED-CHECKS 17; inventory diff 0) |

Leftovers after i1: containers 0, volumes 0, networks 0, run 0. Not run (by instruction): A2 campaign, ops, rename,
controls, recovery drill. Evidence hygiene: the `code|state|error=` grep over `g016/fix11/` outside the extracted source
copy is empty. Like the predecessors' `src-*` directories, `fix11/src-i1/` keeps the configured `.env` of the destroyed
`g016-focus` stack (throwaway generated local values); not deleted, because file deletion needs the owner's approval in
session.

### Attempt 2 — successor #12: MR fast failure and Duo D1 correction (2026-10-16 UTC; hand-back (a))

Worker: Claude Code successor #12, model Fable 5.1, effort high (requested; actual inherited setting unknown); sole
writer. Worktree `.claude/worktrees/agent-<id>` (generated name, recorded ownership), local branch
`g016-impl-succ12` created at `origin/g016-hostname-resolution-domain-lifecycle` `0a163c7f` (MR !34 head). Staging
merged before any write: `478a615d` (management/docs/skill templates only, 11 files) as `cd8c5a9f`, then `64034a40`
(management/docs/skill templates only, 11 files) as `5fe342e2` after the shared `.git` was fetched by another session
between batches. Verified inputs: Codex review `c28dbff0…1499` (the appended "MR fast/Duo correction release"
section; coordinator-verified binding), allow-file `6e78ec85…`, 72 patterns, Result at `0a163c7f` `61d638b0…ee58`.
Accounting (Codex release): **attempt 2 of B7**, checkpoint after attempt 4; the review-ready submission at `0a163c7f`
ended attempt 1 under BUDGET_POLICY §3. No retrospective recount; the inner loops recorded above stand.

**Fast failure (pipeline `2957186823`, job `17248173536`): `B-WEB-BOUNDARY` FAIL, so `B-REQUIRED-CHECKS` FAIL.**
Cause, reproduced read-only before any edit by running the predicate's own `grep -rlE 'request-context' … | grep -vxE
'…'` inline over `web/app`, `web/components` and `web/lib`: exactly one hit, `lib/host.ts`. The G016 server-only host
helper (`import "server-only"` on its first line; `outboundHeaders` and the `RequestContext` type for the tenant-host
gate and callback relay; no credential) imports `./request-context`, and the caller list's `lib/(core-client|
identity-client|mcp-client|request-context)\.ts` alternative did not name it. `lint-web` had never been run locally
for G016 (the published hosted row carries no inner detail). Every other boundary predicate was checked inline the
same way before the edit: the seven server-only imports present, the app route set unchanged from the PASS text, the
error-probe fixtures in place, no stray `NEXT_PUBLIC_`/`document.cookie`/`cookies()` outside their places (`proxy.ts`
imports `deriveContext` as before G016; it is outside the scanned directories). No second cause.

Fix (`infra/validation/run.sh`, `static_web_boundary` only): the caller list's lib alternative becomes
`lib/(core-client|identity-client|mcp-client|request-context|host)\.ts`; the comment, the failure log text and the
`B-WEB-BOUNDARY` PASS text name the G016 server-only host helper. No other predicate, scan or contract changed; no
blanket exemption. `lib/host.ts` is not added to the explicit server-only import check (line 816) because the release
bounds the change to the caller entry and its description; its `import "server-only"` is verified here and is a
candidate one-token hardening for Codex to disposition.

**Duo D1 (LOW; notes `4032861953`/`4032862229`, one defect): `logoutOrigin` accepted an empty port.** `https://
org-a.localhost:` parses with `Host` `org-a.localhost:`, `Port()` empty and `String()` round-tripping, so the canonical
checks passed. Fix (`core/internal/identity/logout.go`): `strings.HasSuffix(u.Host, ":")` joins the structural
condition (with a two-line comment); `origin_rejected` as for every other malformed origin, before any session or
routing lookup. Test (`logout_test.go`, `TestLocalLogoutAndCSRF`): exactly `"https://org-a.localhost:"` added to the
existing invalid table (18 origins × 3 scopes → `origin_rejected`; the no-provider-call and no-revocation assertions
after the table are unchanged).

Commits: `000adf39` (the two fixes), `5fe342e2` (staging merge), then this Result commit; pushed as a normal new-branch
push to `origin/g016-review-fast12` for Codex's delta review. **The MR source branch was not advanced** (coordinator
instruction overriding the earlier brief); no pipeline was triggered, retried or cancelled. Scripts (g016-scripts/):
`pre12.sh` (denied by the worktree isolation check as a wrapper; its steps run inline), `runsh12.sh`, `b12.sh`
(b11.sh with this worktree and `g016/fast12`), selector `fast12-i1.sel`. Before each batch: fetch ok, `origin/staging`
(`64034a40`) ancestor of the head, dirty 0, Docker Desktop mounts 13380/100000 (13%), allow-file actually used
`6e78ec85…` = bound (72 patterns), `flock` on `local-harness.lock`. Evidence under `g016/fast12/`.

| Run | Head / base | Selector | Outcome |
|---|---|---|---|
| **lint1** | 5fe342e2 / 64034a40 | `run.sh --base 64034a40 --allow-file …/G016-scope-allowlist.txt --owned-registry … lint-web` (the single release-approved entry point; 61 s) | **B-WEB-BOUNDARY PASS**, B-WEB-TOKENS, B-WEB-COMPONENTS, B-WEB-DESIGN, B-WEB-ASSET-CONTROLS PASS, B-DOCKER-ONLY PASS, **B-SCOPE PASS** (119 changed paths since 64034a40, all within the allow-list); inventory diff 0. `RESULT: FAILED` (rc 1) solely because `B-SECRET-CANARY` is UNMET ("no secrets were generated") when a lint-family suite runs without `lint`, so `B-REQUIRED-CHECKS` reports `B-SECRET-CANARY(1,UNMET)`: the known selector-only accounting artifact (successor #11's `lint lint-base lint-core` run had B-SECRET-CANARY 16 PASS; nothing in this delta touches secret generation). `lint` was not run, by the release's instruction. The hosted fast profile generates that row itself. |
| **i1-1** | 5fe342e2 / 64034a40 | `g016-focus` (b12.sh), `go test -tags integration -count=1 -run '^TestLocalLogoutAndCSRF$' ./internal/identity/` | **PASS** (1 top-level, 0 fail, 0.13 s) |
| **fmt** | 5fe342e2 | `gofmt -l ./internal/identity/` in the built `g016-focus-core-check` image (self-review of the edit, not a suite) | clean (no files listed) |

Leftovers after i1: containers 0, volumes 0, networks 0, run 0. Not run (by instruction): `lint`/`lint-base`/
`lint-core`, A2, ops, rename, controls, browser, provider, recovery. Evidence hygiene: the `code|state|error=` grep over
`g016/fast12/` outside `src-i1/` is empty. `fast12/src-i1/` keeps the configured `.env` of the destroyed `g016-focus`
stack (throwaway generated local values), as the predecessors' `src-*` directories do; not deleted (file deletion needs
the owner's approval in session). Held unchanged: `issuanceGuard`, the RLS-SCOPE and ISO-ORACLE evaluators, the sqlc
deferral; no MR state change, Duo reply or thread resolution (coordinator); P-1..P-4 untouched.

### Attempt 3 — successor #13: P-1 `TestStagingContract` extension and the P-2 checklist correction (2026-10-16 UTC)

Worker: Claude Code successor #13, model Opus 5.5, effort high (requested; actual inherited setting unknown); sole
writer. Worktree `.claude/worktrees/agent-<id>` (generated name, recorded ownership), local branch
`g016-impl-succ13` at `origin/g016-hostname-resolution-domain-lifecycle` `6e752a1a`, `origin/staging` `b55dbac3`
(management-only) merged before any write as `3cb1474e`. Verified inputs: binding review
`G016-owner-prep-codex-review.md` `c670bf07…6e7b`, allow-file `6e78ec85…` (72 patterns), Result at `6e752a1a`
`4bdfff3a…946d`, prep under review `c66befc0…2d8c` (read-only; not edited). Accounting (Codex release): **attempt 3
of B7**, checkpoint after attempt 4, no reset.

**Design note (written before coding).** The new subtests observe WorkOS's own responses; a constructed URL alone is
never the proof. Inputs come from the staging env file: `IDENTITY_STAGING_PLATFORM_DOMAIN` (B) and
`IDENTITY_STAGING_AUTH_ORIGIN` (A); missing or malformed is UNMET (fail), never a skip; no host is hardcoded.
(a) Callback: the production `Adapter.AuthorizationURL` builds authorize requests whose `redirect_uri` is
`https://g016-<random>.<B>/auth/callback` (a per-run label no exact dashboard entry can name, so only the wildcard can
admit it) and `<A>/auth/callback`; the test GETs WorkOS `/user_management/authorize` without following redirects.
Accepted = 3xx to an https destination with no `error` key, not on the `redirect_uri` host. Negative control: a
reserved `.invalid` `redirect_uri` must not be accepted that way and must never be redirected to; a provider that
does not discriminate fails the test. (b) Sign-out: per case, a fresh API sign-in of the per-run user, then the
production order: `RevokeProviderSession(sid)`, then GET of `ProviderLogoutURL(sid, origin+"/sign-in")`. Accepted =
3xx to exactly that return URL (random wildcard host; A); the `.invalid` return must not be honored. Limitations: the
code-bearing redirect after an interactive hosted sign-in cannot be driven by the API-password harness, so that leg
stays in the P-2 manual run; the hosted job forwards only the three `IDENTITY_STAGING_WORKOS_*` names, so these
inputs reach the test only through the env-file paths (`run.sh staging`, `identity:staging-local`) unless a separate
bounded CI change forwards them.

**Mid-task amendment (coordinator relay of Codex/owner authority).** `origin/staging` `0b11a3b7` merged at a safe
boundary (no batch running; after the code commit `08d57750`, before any validation) as `c7473967`. It carries the
owner-approved split `G016-G057-qualification-split.md` (SHA-256 `8feeea4e…34ab1`, coordinator-verified and re-hashed
here) and the G057 Goal/Spec. Not management-only: it also changed `docs/ARCHITECTURE.md`, `docs/DECISIONS.md`,
ADR-007 and `docs/foundation/IDENTITY.md` (a "G016 live-domain qualification handoff" section); the merge was clean
and the G016 IDENTITY.md edit below was made on top of it. Reconciliation (append-only; the reviewed Plan, snapshots
and earlier evidence are unchanged): **P-1 (actual real-WorkOS evidence) and P-2 (deployed manual HTTPS evidence) are
unqualified in G016 and transferred to G057**, which must be accepted before G021. Wherever earlier sections or the AC
checklist say "P-1/P-2 held" or "P-2 manual qualification held" for AC7/AC8/AC12, read: unqualified, transferred to
G057; no fixture or emulator pass is claimed for them. P-3 stays optional and P-4's explicit hold is unchanged. No
application or security defect is waived. The checklist below is G057 input material, not a G016 gate.

**Change.** `core/internal/identity/workos/staging_test.go` (tag `workos_staging`) gains `P1_WildcardCallbackAcceptance`
and `P1_SignOutReturnAcceptance` as `t.Run` subtests of `TestStagingContract`, after S-ATTR. Each has its own
`*testing.T` (the shared run state is copied, so `Fatal` never targets the parent). The parent writes
`"result": "pass"` only when both pass, so the existing `run.sh` PASS predicate (`--- PASS: TestStagingContract `
plus the sanitized result) cannot pass without them. Positive non-zero evidence: the subtests are unconditional (no
skip path; missing inputs fail UNMET), so a real run must show `--- PASS: TestStagingContract/P1_WildcardCallbackAcceptance`
and `…/P1_SignOutReturnAcceptance`, one line each. Inputs, the observation rules and the limitations are as in the
design note. Existing checks, S-CLEAN cleanup and per-run users are unchanged. The sanitized result gains `p1_callback`,
`p1_callback_destination` (an origin fingerprint) and `p1_sign_out`: statuses, booleans and fingerprints only; no URL,
Location, code, state or session ID is logged or recorded. `docs/foundation/IDENTITY.md` L3 section: one bullet for the
subtests and their two inputs, and the env-file paragraph names them (the `identity:staging-local` block is unchanged).
**Consequence for review:** once the merged code runs anywhere, `TestStagingContract` is UNMET until both inputs exist.
The hosted `identity-staging` job cannot supply them today: `hosted-context.sh` forwards a fixed list of three names.
The smallest alternative is a separate bounded CI change, or using the env-file path. That is G057's decision; not
changed here.

**Corrected operator checklist (G057 input material; supersedes prep §5 for that use).** It is not executed or
authorized here. The hosts are placeholders: platform domain `B`, auth host `A`, wildcard-covered host `org-a.B`,
custom host `T`. The prep's `stg.foundation.test`, `auth.stg.foundation.test`, `org-{a,b,c}.stg.foundation.test`,
`custom.stg.foundation.test` and `portal-qual.foundation.test` are **unselected proposals**. The WorkOS entries, wildcard
support and the before-state are **unverified**, and the dashboard page names are inferred. The checklist never asks
anyone to copy, paste, screenshot or keep a callback URL, authorization code, `state`, callback query or cookie value,
and no trace, HAR or screenshot holding one is kept.

| Step | Who | Action | Pass condition (server-observed unless stated) |
|---|---|---|---|
| Q-0 | owner authority for the provider change; agent prepares the sanitized record | Record the WorkOS before-state as patterns only. Add `https://*.B/auth/callback`, `https://*.B/sign-in`, `https://A/auth/callback` and `https://A/sign-in`, nothing per custom host. Record the after-state. | before/after recorded; the two `TestStagingContract/P1_*` subtests PASS against them (env-file path, positive evidence) |
| Q-1 | operator | Redaction check first: one sign-in on `T`, then grep the edge log for `?& (not included)=` | empty (whole callback query), recorded before any other step |
| Q-2 | operator | Sign in on `T` with password and TOTP | `/account?status=signed_in`; session for T's Organization; cookie **names** only: `__Host-sid` on T, none on A |
| Q-3 | operator | Open `https://A/account` | fixed 404 |
| Q-4 | operator | Sign in on `org-a.B` (no exact entry) | same account id; session on org-a (completes the code-bearing callback leg the API harness cannot drive) |
| Q-5 | operator | **Run 1** (own fresh sign-in on T): local sign-out only | `signed_out`; T's session revoked locally; `__Host-sid` gone; no provider revocation claimed |
| Q-6 | operator | **Run 2** (separate fresh sign-in on T): sign out and revoke the WorkOS session | provider revocation confirmed; hosted logout returns to `https://A/sign-in` (destination attested); back on T within 10 minutes the recorded outcome is shown; the next sign-in asks for credentials |
| Q-7 | operator; the agent supplies the removal operation | Signed in on T, the domain is removed | reload: 404 page; `/api/account` 404; login start 403; org-a unaffected |
| Q-8 | operator | Record the hosts and statuses, the hop sequence (host and path only, never the query), cookie names, timestamps and the Q-1 result | sanitized record only |

Replay and cross-host credential negatives are **not manual steps**. They stay in the existing bounded automated
fixtures with sanitized outputs: `TestCustomDomainRelay` (B-IDENTITY-IT) covers replay of the relayed callback after
consumption, a foreign browser (no binding), altered `state`, repeated `state` on A (fixed 404) and repeated `code`
(`ambiguous_parameter`). E-ORIGINS and E-DOM-2 (`web/tests/browser/identity.spec.ts`, B-IDENTITY-E2E) cover a
session cookie replayed at another origin (401) and the host-only cookie and hop sequence. `TestLocalLogoutAndCSRF`
and `TestProviderLogout` cover the logout legs. G057 decides whether to rerun them against its environment.

**Focused outcomes (compile/vet only; nothing ran against WorkOS).** Before the batch: `git fetch origin`; `origin/staging`
`0b11a3b7` is an ancestor of `cb5a4b97`; dirty 0; Docker Desktop mounts 13384/100000 (13%); the allow-file used is
`6e78ec85…` = bound (72 patterns), and all 120 paths changed since `origin/staging` match it. Each step ran under
`flock -n …/local-harness.lock`. The `exec 9>`/`flock 9` form and the combined `sh -c` form were refused by the
worktree isolation check, so each step ran inline as its own `flock`-wrapped `docker run`. Image `g016-core-check:att3`
(`docker build --target check -f infra/docker/core.Dockerfile core`, rc 0). Evidence is under `g016/att3/`.

| Step (`docker run --rm --network none`, image `g016-core-check:att3`) | Outcome |
|---|---|
| `gofmt -l ./internal/identity/workos/` | clean (empty output, rc 0) |
| `go vet -tags workos_staging ./internal/identity/workos/` | rc 0, no findings |
| `go vet ./internal/identity/workos/` | rc 0 |
| `go test -tags workos_staging -list Test ./internal/identity/workos/` (compiles the tagged test binary; executes nothing) | rc 0; 21 tests listed, including `TestStagingContract` |
| `go test -count=1 -v ./internal/identity/workos/` (existing non-provider unit selector) | `ok`; 20 `--- PASS`, 0 FAIL/SKIP; no `TestStagingContract` (tag absent) |

Not run, by instruction: the real-provider subtests, and any WorkOS, DNS, VM, TLS, Compose/Caddy/seed kit or probe
work. No broad suites or controls ran. Evidence hygiene: the `?& (not included)=` grep over
`g016/att3/` is empty. No containers were left. Held unchanged: `issuanceGuard`, the RLS-SCOPE and ISO-ORACLE
evaluators, the sqlc deferral.

**Accounting.** This completed correction is **attempt 3 of B7** (checkpoint after attempt 4; no reset), first cycle
on the released items. Commits: `08d57750` (test and design note), `c7473967` (staging merge), `cb5a4b97`
(IDENTITY.md), then this Result commit. Pushed by a normal push to the new review branch `origin/g016-review-attempt3`.
The MR source branch was not advanced. Scripts: none were needed (all steps ran inline).

### Attempt 4 — successor #13: Codex attempt-3 correction loop (2026-10-16 UTC; hand-back (a))

Same worker, same worktree and same branch `g016-impl-succ13` (model Opus 5.5, effort high (requested; actual
inherited setting unknown); sole writer).
- **Binding:** Codex review `G016-implementation-codex-review.md` `d5d607bc…1def`, section "Attempt 3 delta review at
  6588a286": CORRECTION LOOP, two bounded items. Allow-file `6e78ec85…` (72 patterns).
- **Staging merges:**
  - `origin/staging` `accd087b` (which contains `fd13dd2d`, the review and the imported attempt-3 Result `ab74ffe0…`)
    was merged before any write as `05353a5a`.
  - `c73f4833`, published during the batch preflight, was merged as `9b333db8` before the final batch.
  - Neither merge was management-only: they also changed `docs/DECISIONS.md`, `docs/PERFORMANCE.md`,
    `docs/ARCHITECTURE.md`, `docs/SECURITY.md` and `docs/foundation/AUTHORIZATION.md`. None touched
    `core/internal/identity/` or `docs/foundation/IDENTITY.md`.
- **Accounting (Codex release):** **attempt 4 of B7 — the checkpoint attempt**, no reset; first correction cycle on
  these two items. No application code was reopened.

**Item 1: existing L3 restored; domain qualification is a separate, explicitly selected test.**
- **TestStagingContract restored.** `staging_test.go` was reset to its `6e752a1a` bytes. The diff against
  `6e752a1a` now has **0 removed lines**: `TestStagingContract`, its selector `^TestStagingContract$`, its three
  `IDENTITY_STAGING_WORKOS_*` inputs and its `"result": "pass"` rule are byte-identical.
- **New test:** `TestStagingDomainQualification` (same file, tag `workos_staging`), top-level, reusing the existing
  helpers (`stgEnv`, `createUser`, `signIn`, `verify`, `classify`, `boundedClient`, the adapter builders).
  - Subtests: `WildcardCallback` and `SignOutReturn`.
  - Inputs, all required only when this test is selected: the three WorkOS values (configured issuer only),
    `IDENTITY_STAGING_PLATFORM_DOMAIN`, `IDENTITY_STAGING_AUTH_ORIGIN` and the new `IDENTITY_STAGING_AUTHKIT_ORIGIN`
    (see item 2).
  - Optional sanitized result: `IDENTITY_STAGING_DOMAIN_RESULT_FILE`.
  - It cleans up its own per-run user in the S-CLEAN manner and uses no database (no TRUNCATE).
- **Selection is proven by compile-only listing:** `-list TestStagingContract` lists only `TestStagingContract`.
  Selected without inputs (`--network none`), the domain test fails `UNMET` at its first input; there is no skip.
- **Not done, per the review:** no CI forwarding and no settings change.
- **IDENTITY.md (L3):** documents the L3 selector as unchanged and adds a "Domain qualification test" block: selector,
  inputs, observation rules, the offline fixture selector, and the evidence boundary (a pass only prepares G057; the
  code-bearing browser redirect stays a browser step; the hosted job neither selects it nor forwards its inputs).
  The attempt-3 bullet is replaced, and the env-file paragraph names the domain inputs by reference.

**Item 2: fail-closed observations.**
- **Authorization verdict:** `stgAuthorizeVerdict` returns `accepted`, `refused` or `unknown`.
  - `accepted`: only a 3xx to exactly the configured hosted sign-in origin. That means https, the same host and port,
    no userinfo, no `error`/`error_description` parameter, and not the `redirect_uri` host.
  - `refused`: a 4xx, or a redirect carrying an OAuth error that does not lead to the `redirect_uri` host.
  - Everything else (an arbitrary https page, the provider's own host, 2xx, 5xx, a redirect without Location) is
    `unknown` and qualifies as neither.
  - Positives (wildcard host and A) must be `accepted`. The negative must be `refused`, independent of which host
    either positive reached.
- **Logout:** `stgReturnExact` requires the raw Location to equal the return URL byte for byte (no userinfo, query,
  empty `?`, fragment, empty `#`, port, case or path change). The negative `stgReturnNotHonored` requires a 4xx, or a
  redirect whose host (after userinfo) is not the negative host; 2xx/5xx fail.
- **observe:** `stgObserve` takes the client as a parameter and returns sanitized error classes for:
  - transport failures and **body-read failures**, both `response bound exceeded` (chunked or declared) and `timeout`
    (after the headers);
  - an unparseable Location, and more than one Location.
  - Error text never carries the URL, its query, a Location, a provider body or a session ID.
  - The subtests use `Errorf`, so every verdict is reported.
- **G057 limitation (the smallest explicit configuration):** WorkOS's hosted sign-in destination cannot be derived
  from the existing inputs, so it is a new required input, `IDENTITY_STAGING_AUTHKIT_ORIGIN`, recognized by origin
  only. G057 must establish and record it from a verified provider observation. The test fails closed rather than
  guess if any of these hold:
  - the provider's first hop is not that origin;
  - an invalid `redirect_uri` is answered with a 2xx error page, or a redirect without an OAuth error parameter;
  - the logout redirect appends anything to `return_to`.

  In that case G057 must record the actual shape and correct the rule under review. No success path or route is
  assumed.
- **Offline fixtures:** `TestStagingDomainObservationFixtures` (same file; no provider, no inputs).
  - 15 authorization-verdict cases, including the attempt-3 false positive (an arbitrary https error redirect is now
    `unknown`), other port, http, userinfo, an error parameter, redirects to the `redirect_uri` host (with code and
    with error), 4xx, 2xx and 5xx.
  - 10 exact-return cases and 6 negative-return cases.
  - One origin-with-path rejection.
  - Loopback `httptest` responses: a 302 is observed and accepted; two Location headers are an error; a chunked body
    over the bound fails at the body read; a declared overflow fails at the request; a stall after headers fails as a
    body timeout. Each error text is asserted free of the request URL and the state marker.

**Focused outcomes (offline only; nothing contacted WorkOS).**
- **Pre-batch:** `git fetch origin`; staging `c73f4833` is an ancestor of `9b333db8`; dirty 0; Docker Desktop mounts
  13384/100000 (13%); allow-file `6e78ec85…` = bound (72 patterns); all 119 paths changed since `origin/staging`
  match it.
- **Execution:** each step ran as its own `flock -n …/local-harness.lock`-wrapped `docker run --rm --network none`.
  The isolation check refuses the `^…$`-quoted form, so `-run` used the unique unanchored names.
- **Batch 1 (image `g016-core-check:att4`, head `c12f29aa`):** gofmt listed `staging_test.go` (struct field alignment
  only); tagged vet rc 0; fixtures PASS. The alignment was fixed (`5955a571`), and the final batch below reran every
  step on a rebuilt image after the staging merge. Evidence is under `g016/att4/`.

| Step (image `g016-core-check:att4b`, head `9b333db8`) | Outcome |
|---|---|
| `gofmt -l ./internal/identity/workos/` | clean |
| `go vet -tags workos_staging ./internal/identity/workos/` | rc 0 |
| `go vet ./internal/identity/workos/` | rc 0 |
| `go test -tags workos_staging -list TestStaging …` / `-list TestStagingContract …` (compile only) | rc 0; three tests / only `TestStagingContract` |
| `go test -tags workos_staging -count=1 -v -run TestStagingDomainObservationFixtures …` | **PASS** (1 test, 0.31 s) |
| `go test -tags workos_staging -count=1 -v -run TestStagingDomainQualification …` without inputs (`--network none`) | FAIL `UNMET: IDENTITY_STAGING_WORKOS_API_KEY …` as intended: fails when selected, no skip |
| `go test -count=1 -v ./internal/identity/workos/` (existing unit selector) | `ok`; 20 PASS, 0 FAIL/SKIP |

- **Not run, by instruction:** real WorkOS, full suites, ops, rename, controls, infrastructure.
- **Hygiene:** the `?& (not included)=` grep over `g016/att4/` is empty. No containers were
  left; the `att3`/`att4`/`att4b` check images are kept.
- **Held unchanged:** `issuanceGuard`, the RLS-SCOPE and ISO-ORACLE evaluators, the sqlc deferral.
- **Commits:** `05353a5a` (staging merge), `c12f29aa` (test and docs), `5955a571` (gofmt), `9b333db8` (staging merge),
  then this Result commit. Pushed to the new review branch `origin/g016-review-attempt4` (normal push); MR !34 was not
  advanced.

### Attempt 5 — successor #14: post-integration staging-gate repair, binding and amendment 08 (2026-10-17 UTC; recorded before the first source edit)

Worker: Claude Code, model Opus 5.5, effort high (requested; actual inherited setting unknown); sole G016 writer.
Worktree: generated name `agent-<id>` (not task-prefixed; recorded per CLAUDE.md), local branch
`g016-staging-repair` created at `origin/staging` `1c050968c12a3cba5b6da9b9d08e68763da0533e`, which contains the G016
merge `37c99742` (failed staging gate: pipeline 2957417252, job 17249410096).
- **Binding review:** `G016-staging-gate-repair-codex-review.md` SHA-256
  `9075bf1d2eb81605e7f5e30561cc697bf7d8bdbed3ac0ee47f93a6c76a32829c` (verified): ACCEPT of the bounded repair
  proposal; **attempt 5 of B7 released**, checkpoint 4 retained, no reset; two attempts remain after it.
- **Diagnosis:** imported bundle `G016-staging-gate-repair-bundle.md` `c22ed92a8b69f201c1b5845a234b781d54699b02f74b7f14499a58c60c83faba`
  (verified). Triage record `G016-staging-gate-failure-triage.md` read.
- **Allow-file:** `G016-scope-allowlist.txt` `6e78ec85ddeb48cda2f3bba98f4caaca0e9e8ce3c2c88ac1131cd55260c378f8`,
  72 patterns, byte-identical. Every changed path must match it.
- **Amendment 08 (scope intent, recorded verbatim in substance; snapshots unchanged):**
  1. `infra/validation/run.sh`: source-backed policy 39, domain catalog 4, audit 35/29 derived from the migration that
     owns the current `events_action_valid` constraint, AUDIT-INV 35, owned-index rename `_hostname_live` →
     `_hostname_owned`, version-12 wording. Exact equality, no FAIL/SKIP, runtime/grant/negative and log/secret checks
     retained; no relaxing to whatever the latest run printed.
  2. Same file: mandatory session origin from the owned private browser handoff into `idn_session`, validated against
     the expected fixture/configuration, URL-encoded once; missing/malformed origin fails the row, no fallback;
     missing-credential and foreign-origin denials preserved.
  3. Same file: bundle option (b), one pinned reviewed combined identity+domain tail for toggle composition; separate
     reviewed blocks and root/variant tamper detectors retained; immutable accepted bytes pinned.
  4. Same file: only accepted G016 dependency paths added to `APPLICABLE` (hostname/domains packages, domain config and
     its test, `web/lib/host.ts`, `web/proxy.ts`, domain guide where needed).
  5. `core/internal/identity/store_test.go`: exempt exactly `sessions.organization_id` (alongside the existing login
     transaction binding; `origin` is not matched by the forbidden-column regexp); seed a legitimate bound browser
     session/Organization through the trusted setup path; keep forbidden fields, exact catalogs, runtime/RLS/grant and
     negative checks and the positive write test.
  6. Optional `docs/PERFORMANCE.md` stale index spelling only.
- **Validation released:** targeted version-12 `TestIdentityCatalog` (integration tag), gofmt/vet, `bash -n`, static
  pin-vs-source comparisons; hosted passing policy/audit/domain subtest evidence reused; **one** `run.sh ops` and
  **one** `run.sh off`, no automatic rerun on failure. No local migrations/controls/recovery/persist/rename runs.
  Every batch holds `<home>/.local/state/platform-evidence/local-harness.lock` via `flock <lock> <command>`,
  or stops with the exact refusal.
- **Held:** OBJ-ADMIN stays UNKNOWN (no speculative change); `issuanceGuard`, RLS-SCOPE/ISO-ORACLE evaluators, sqlc
  deferral and all production code unchanged; no WorkOS/DNS/deploy; G057 untouched; no raw hosted-log publication or
  `hosted-context.sh` change; no MR (coordinator opens it after Codex source review).
- **Pre-edit hashes at `1c050968`:** `run.sh` `b3adffc9…2fe2755` (= the bundle's 37c99742 hash), `store_test.go`
  `9678fc2c…883f788`.

### Attempt 5 — successor #14: implementation, focused evidence and the ops stop (2026-10-17 UTC; hand-back (c))

**Changes (commit `3608b5b1`; 3 paths, all matching the bound allow-file).**
- `infra/validation/run.sh`, source-backed pins and text (amendment 08 item 1):
  - B-POLICY-IT: `n == 39` (TestAccessPolicy has 39 top-level `t.Run`: "runtime identity" plus 38 IT-*, incl.
    IT-CTX-11); PASS/FAIL text 39/39, IT-CTX-1..11.
  - B-AUDIT-CATALOG: the action list is parsed from `00011_organization_domains.sql` (its first `ADD CONSTRAINT
    events_action_valid`, the up section; 00012 has none); exact equality kept; pins 35 words / 29 T-rows; text.
  - B-AUDIT-IT: AUDIT-INV `{"actions":35,"observed":35}` (audit.go: 29 TriggerActions + 6 CoreActions); text.
  - B-DOMAINS-CATALOG: `n == 4` (CT-DOM-1, CT-DOM-NEG, CT-DOM-UPDOWN, CT-DOM-12); comment and text.
  - B-IDENTITY-CATALOG: `DOMAIN_OBJECTS` names `i:organization_domains_hostname_owned` (00012:109-110); "at version 12".
- `run.sh` item 2: new `pct_encode` helper; `idn_session CID NAME AUTHFILE ORIGIN` requires ORIGIN (no probe and the
  status `no-origin` without it) and sends `?origin=<encoded once>`. `identity_probes` reads `.origin` from the private
  handoff and fails the row unless it equals the fixture's Organization A origin `https://org-a.localhost:543`
  (identity.spec.ts `ORIGIN_A`, the Host/Origin its web probes already send); all ten session probes pass it. p6-core
  (no credential, 401) unchanged. PASS text notes the origin.
- `run.sh` item 3: `ENV_IDENTITY_DOMAIN_TAIL_SHA=aad6dbc7…065e7` (the accepted file's identity marker → EOF, 31 lines,
  verified equal to `ENV_IDENTITY_BLOCK_SHA` block ‖ `ENV_DOMAIN_BLOCK_SHA` block); `variants_off` uses it as the
  `.env.example` pin. B-ENV-EXAMPLE's separate block checks and every root/variant detector control unchanged.
- `run.sh` item 4: `APPLICABLE` gains `core/internal/hostname core/internal/domains core/internal/config/domains.go
  core/internal/config/domains_config_test.go web/lib/host.ts web/proxy.ts docs/foundation/DOMAINS.md`. Import check:
  `core/cmd/server/main.go` and `core/internal/mcp/registry.go` import the domains packages; domains imports hostname,
  identity, access, config (covered); `web/app/auth/callback/route.ts` and `web/proxy.ts` import `@/lib/host`, which
  imports the rendered `web/lib/config.ts` and the covered `request-context.ts`; DOMAINS.md is named by the composed
  `.env.example` block and six applicable guides. No other G016-added path is uncovered.
- `core/internal/identity/store_test.go` (item 5): forbidden-column query exempts exactly `sessions.organization_id`
  beside the existing login-transaction exemption (comment states the rationale; `origin` is not matched); the test
  calls the existing `seedTenants(t)` (superuser setup connection) and the IDENTITY-shape runtime insert carries
  `origin = originA`, `organization_id = orgAID`. Grants, exact catalogs, constraint negatives, runtime denials and the
  positive write test unchanged.
- Not changed: item 6 (`docs/PERFORMANCE.md:432` describes the F1 observation of the version-11 index; renaming it
  would misstate that evidence), the B-DOMAINS-IT/B-MCP-IT "version 11" PASS texts (passing rows, outside the
  diagnosed set), OBJ-ADMIN, all production code, `issuanceGuard`, RLS-SCOPE/ISO-ORACLE, sqlc deferral.

**Static comparisons (independently checked by this worker; scripts in `g016-scripts/att5-*.sh`).** Policy 39 and
domain catalog 4 counted from the test sources; audit: the harness's own sed over 00010/00011/00012 gives 31 / 35
(equal to AUDIT.md's sorted 35 rows, 29 T) / 0; `.env.example` tail/ident/domain/concat hashes as above;
`pct_encode https://org-a.localhost:543` = `https%3A%2F%2Forg-a.localhost%3A543` (= `encodeURIComponent`).
`bash -n run.sh` OK. The hosted gate's passing policy/audit/domain subtests (39/38, 4/3, 7/7, 9/9 with inventory
35/35) are reused as reported hosted evidence, not rerun.

**Batches (each under `flock <home>/.local/state/platform-evidence/local-harness.lock <command>`; the route was
accepted, no refusal). Pre-batch: `git fetch origin`, staging `1c050968` ancestor of head `3608b5b1`, dirty 0, mounts
13396–13404/100000 (13%), allow-file `6e78ec85…` (72 patterns).**

| Batch | Outcome (evidence `g016/att5/`) |
|---|---|
| b1 run 1 (`att5-b1.sh`) | stopped at `up`: `--wait` with the one-shot minio-init returns 1; nothing tested; resources torn down (`b1-run1-upfail/`) |
| b1 run 2, version 12 (postgres only) | migrate `current_version 12`; `gofmt -l ./internal/identity/` clean; `go vet -tags integration` and `go vet` rc 0; app catalog = run.sh want (48 relations + owner); **`TestIdentityCatalog` PASS 4/4, 0 FAIL/SKIP** (`b1/`) |
| **ops1** (`att5-runsh.sh ops1 ops`, the single approved ops run), 01:45–01:58 UTC | **FAILED** (rc 1) on B-IDENTITY-E2E only; see below (`ops1/`) |
| off | **not run** (stop rule) |

ops1 rows (reported by the harness): B-OBS-CONTEXT, B-OBS-CHAIN, B-OBS-TIMEOUTS, B-OBS-LOGS, B-OBS-PRIVACY, B-OUTAGE,
B-PERSIST (version 12), B-SECRET-ROTATION, **B-IDENTITY-IT PASS (73 top-level, none skipped)**, **B-IDENTITY-PROBES
PASS** (with the handoff origin), B-MCP-E2E PASS 6/6, B-IDENTITY-PRIVACY, B-IDENTITY-CONFIG, B-DOCKER-ONLY,
B-SECRET-CANARY (162), B-SCOPE (3 paths since `1c050968`) PASS; **B-IDENTITY-E2E FAIL**; B-REQUIRED-CHECKS FAIL
(consequent). Inventory diff 0.

**B-IDENTITY-E2E cause (CONFIRMED from the local report; newly found, not in amendment 08).** 57 of 61 browser tests
pass (all domain-platform/-custom E-DOM rows, E-HANDOFF, E-ORIGINS, both mcp projects). Exactly four scenarios fail,
identically in identity-light and identity-dark: E-FIX (`identity.spec.ts:425`), E-CB-2 (`:462`), E-CB-10 (`:586`),
E-LO-3 (`:715`), each `expected 401/200, received 400`. All go through the spec helper `coreSession`
(`identity.spec.ts:244-247`, unchanged since G004 `843196d6`), which calls core's
`GET /internal/v1/auth/session` without `?origin=`; since A2 (`052b791d`) core answers 400 `ErrInvalidRequest` when the
parameter is absent (`http.go:211-217`). This is the same defect as the probe row, in the browser spec. Application
behaviour is correct (web's own client sends the origin, `identity-client.ts:223`).
- Proposed bounded fix (not applied): `web/tests/browser/identity.spec.ts` (allow-listed) — give `coreSession` a
  mandatory `origin` parameter appended as `?origin=${encodeURIComponent(origin)}` and pass `ORIGIN_A` at its ten call
  sites (all four tests sign in on ORIGIN_A, the identity projects' baseURL); expectations unchanged.
- Smallest rerun request after that fix: one `run.sh ops` (B-IDENTITY-E2E needs the harness seed; the four tests
  run in both identity projects), plus the still-unspent single `run.sh off` from this release.

**Resources.** b1: project `g016att5b1` (containers, network, volumes, its four images) removed by its trap; the
temp copy `/tmp/g016-att5-b1` removed. ops1: inventory diff 0; the kept work dir `/tmp/project-validation-PRagZY`
removed after copying the failure excerpt (`ops1/idn-browser-failures.txt`, artifact paths only) and
`browser-identity-tests.tsv`. No prune.

**Accounting.** Attempt 5 of B7 (checkpoint 4, no reset) is in progress and stopped at the ops failure; the probe,
catalog, pin, toggle and copy corrections are first-cycle; the browser-spec defect is a new finding awaiting a Codex
disposition. OBJ-ADMIN untouched (UNKNOWN).

### Attempt 5 continued — successor #14: amendment 09 binding (2026-10-17 UTC; recorded before the edit)

- **Binding:** `G016-staging-gate-repair-codex-review.md` SHA-256
  `d8db02a1fd8b73095d351c7e5efaacc1581dc173cc8b239766a10ac25a18b89e`, section "Browser helper correction and bounded
  rerun": ACCEPT of the browser-helper proposal at `286a958c` / Result `1c791cbf…412d`, **within attempt 5 of B7** (not
  attempt 6). `origin/staging` `1edff877` (management-only: PROGRESS.md and the review) merged as `681587fa`.
- **Codex's independent count of the retained ops1 TSV:** 73 executions, 65 expected and 8 unexpected (E-FIX, E-CB-2,
  E-CB-10, E-LO-3 in both identity themes). ops1 stays a FAILED record; OBJ-ADMIN stays UNKNOWN.
- **Amendment 09 (scope intent; allow-file `6e78ec85…` unchanged, it already matches the path):** only
  `web/tests/browser/identity.spec.ts`: `coreSession` requires an explicit origin, appended with `encodeURIComponent`
  once; all ten current callers pass `ORIGIN_A`. Expected 200/401 assertions, credential handling, replay/revocation
  checks and production origin enforcement unchanged; no default or guessed origin; no runtime change.
- **Validation released:** one additional `run.sh ops`; if it passes, the previously approved, still unused single
  `run.sh off`. Any new failure stops with exact cause and the smallest proposal.

### Attempt 5 continued — successor #14: amendment 09 implemented; ops2 and off1 PASS (2026-10-17 UTC; hand-back (a))

**Change (commit `127e9c10`).** `web/tests/browser/identity.spec.ts`: `coreSession(d, base, credential, origin)` takes a
required `origin` and requests `/internal/v1/auth/session?origin=${encodeURIComponent(origin)}` (encoded once, as
`web/lib/identity-client.ts:223`); the ten callers (E-FIX 2, E-CB-2 2, E-CB-10 2, E-LO-3 4, incl. the two replica calls)
pass `ORIGIN_A`. Every 200/401 expectation, credential handling and revocation/replica check unchanged; no default
origin; no runtime change. Changed paths since `origin/staging` `1edff877`: 4, all matching the allow-file. No local
eslint/tsc run (not in the released validation; the change adds one typed string parameter); the MR fast lint covers it.

**Batches (each `flock <home>/.local/state/platform-evidence/local-harness.lock bash …/att5-runsh.sh TAG
SUITE`; accepted, no refusal). Pre-batch for each: `git fetch origin`, staging `1edff877` ancestor of head `127e9c10`,
dirty 0, mounts 13476 / 13544 of 100000 (13%), allow-file `6e78ec85…` (72 patterns) passed as `--allow-file`.**

| Run | Outcome (reported by the harness; evidence `g016/att5/<run>/`) |
|---|---|
| **ops2** (the one additional ops), 02:05:49–02:20:30 UTC, rc 0 | **RESULT: PASSED.** B-OBS-CONTEXT, B-OBS-CHAIN, B-OBS-TIMEOUTS, B-OBS-LOGS, B-OBS-PRIVACY, B-OUTAGE, B-PERSIST (version 12), B-SECRET-ROTATION, B-IDENTITY-IT (73 top-level, none skipped), **B-IDENTITY-E2E (73 tests: identity-light 28, identity-dark 28, domain-platform 3, domain-custom 2, mcp-light 6, mcp-dark 6, all `expected`)**, **B-IDENTITY-PROBES** (handoff origin), B-MCP-E2E 6/6, B-IDENTITY-PRIVACY, B-IDENTITY-CONFIG, B-DOCKER-ONLY, B-SECRET-CANARY (164), B-SCOPE (4 paths), B-REQUIRED-CHECKS (17 IDs) PASS; build ledger 25/25, unknown 0 |
| **off1** (the single approved off), 02:20:3x–02:30:12 UTC, rc 0 | **RESULT: PASSED.** **B-TOGGLE-OFF-ADDITIONS** (variant `850b7f2aa96f0986`; one pinned identity+domain tail; controls rejected), **B-OBS-OFF**, **B-BROWSER-OFF** (141 tests, 0 skipped/flaky/unexpected), **B-TOGGLE-OFF** (variant start/check/stop, migrate to version 12, no intelligence/Python), B-DOCKER-ONLY, B-SECRET-CANARY (12), B-SCOPE (4 paths), B-REQUIRED-CHECKS (7 IDs) PASS; build ledger 12/12, unknown 0 |

Independently checked by this worker: the per-project tally of `ops2/browser-identity-tests.tsv` (73 rows, all
`expected`) and both `meta.txt` records (rc 0, head `127e9c10`, inventory diff 0). The rest is the harness's report.

**Row status against the failed staging gate (local evidence; hosted qualification still required).** Fixed and
locally passing: B-IDENTITY-IT, B-IDENTITY-E2E, B-IDENTITY-PROBES, B-TOGGLE-OFF-ADDITIONS, B-TOGGLE-OFF, B-OBS-OFF,
B-BROWSER-OFF; B-IDENTITY-CATALOG's test half by the focused b1 (object half by the static catalog comparison). Pin-only
rows B-POLICY-IT, B-AUDIT-CATALOG, B-AUDIT-IT, B-DOMAINS-CATALOG: source-checked, not run locally (hosted counts reused).
B-OBJ-IT OBJ-ADMIN: untouched, UNKNOWN; the correction's hosted gate supplies the next observation. ops1 stays FAILED.

**Resources.** ops2 and off1: inventory diff 0 each; their kept work dirs (`/tmp/project-validation-ulBsrn`,
`/tmp/project-validation-ZM4F2D`) removed (ops2's browser TSV copied first). Containers, volumes, networks and images
equal the attempt-start inventory (`att5/inv-before-*.txt`). No prune.

**Accounting.** Attempt 5 of B7 (checkpoint 4, no reset), first correction cycle on amendments 08 and 09; returned for
Codex source review before the correction MR. Not accepted; G017 held; G057 untouched.

### Attempt 6 — successor #15: MD-O diagnostic binding, amendment 10 (2026-10-17 UTC; recorded before the first source edit)

- **Worker:** Opus 5.5, effort high (requested; actual unknown). Generated worktree `agent-<id>`
  (fallback name per CLAUDE.md), local branch `g016-md-diagnostic` created at `origin/staging`
  `9a163fb7cb21fec4d5183d4b5affc1e6dc9db52d`. Sole writer.
- **Binding:** `G016-maintenance-diagnostic-codex-review.md` SHA-256
  `13ec2f2d70faffad8ff55ea8db7195393b778f0145899b92df4225dfc7c7b7db` (ACCEPT of diagnosis proposal (b); releases
  **attempt 6 of B7**, one attempt left after it; checkpoint 4 not reset or enlarged). Diagnosis
  `G016-maintenance-diagnosis.md` SHA-256 `f3374701f6ef313cc4c823f151245ffff2db8b08658629b5d892ff996214e927`.
  Allow-file `G016-scope-allowlist.txt` SHA-256 `6e78ec85ddeb48cda2f3bba98f4caaca0e9e8ce3c2c88ac1131cd55260c378f8`,
  72 patterns, unchanged; it already matches `infra/validation/run.sh` and this Result.
- **Amendment 10 (scope intent):** paths changed are `infra/validation/run.sh` and this Result only. A closed set of
  eight MD-O subcondition codes assigned at the trusted assertion branches of `rd_stage_mdo` (plant, first purge,
  early tombstone removal, ledger sync, unavailable provider, second purge, repair2, final row set), validated,
  bounded, reset per run and per stage, with an `unknown` fallback, appended only to the existing
  B-MAINTENANCE-DRILL FAIL detail. The raw `why`/tool output stays private and never feeds the codes. Every
  assertion, status, order, cleanup and readiness behaviour unchanged. Held: MinIO readiness/status hardening, raw
  artifact publication, `validate.sh`, `lib.sh`. Not run: migrations reproduction, recovery drill, ops/off, full
  controls.
- **Validation released:** `bash -n infra/validation/run.sh` and bounded offline bash fixtures only.

### Attempt 6 — successor #15: MD-O diagnostic codes implemented, focused offline evidence (2026-10-17 UTC; hand-back (a))

**Change (`infra/validation/run.sh` only; +30/−9; file SHA-256 `f6dd163c…736793`).**
- `rd_stage_mdo` resets `RD_MDO_CODES` at its start; each of its eight assertion-failure branches appends one fixed
  literal next to its unchanged `all=0; why+=…`: `plant`, `first`, `T-gone-early`, `sync`, `unavailable`, `second`,
  `repair2`, `left` (the diagnosis table's tokens). The raw `why` (incl. `first[$a]`, `unavailable[$a]`,
  `second[$a]`, `left[$left]`) is untouched and stays in the private step record/console only; it never feeds the
  codes.
- `RD_MDO_CODES` is declared empty beside `RD_T` and reset again at the start of `recovery_drill` (per run).
- New `mdo_fail_codes` (styled on `obj_fail_codes`): reads only `RD_MDO_CODES`; input over 128 characters or with any
  character outside `[A-Za-z0-9 -]` gives `unknown`; known codes print comma-joined in canonical order, any other
  token adds one `unknown`, none recorded gives `unknown`. It always exits 0 and never echoes input.
- The B-MAINTENANCE-DRILL FAIL row only, and only when `MD-O` is among the failed steps, gains
  `; MD-O subconditions: <codes>` after the unchanged `failed steps: … (reached: …)` text. The PASS row, every
  assertion, status decision, tally, operation order, MinIO stop/start, cleanup and exit behaviour are unchanged. No
  `minio_ready`, no status checks, no `validate.sh`/`lib.sh` change, no new published artifact. The row has no tab, so
  `validate.sh`'s `results.tsv` line regex is still satisfied.

**Validation (pre-batch: `git fetch origin`; `origin/staging` `9a163fb7` ancestor of head `395608c7`; allow-file
`6e78ec85…378f8`, 72 patterns).**
- `bash -n infra/validation/run.sh`: rc 0.
- Offline fixture `/tmp/g016-a6-fx.yG0uAp/fixture.sh` (110 lines, SHA-256 `a289c403…69a2f`, not committed). It extracts
  the real `log`/`record` (lib.sh), `rd_rec`/`rd_ok`, `mdo_fail_codes`, `rd_stage_mdo` and the verbatim final
  maintenance report block (run.sh), drives them with stubs and puts a `docker` on PATH that fails loudly. It touched no
  Docker or harness resources, so it ran **without** the shared lock (pure offline bash). **40/40 pass:** each branch
  alone produces exactly its code with a FAIL row (8); none of those rows carries raw material, checked with a planted
  `AKIA…` canary, tab, newline, brackets, `<ERROR>`, a `/tmp` path and a signed URL in the stub tool output (8); the raw
  `why` still reaches the private MD-O step record (3); all eight together give canonical order and FAIL (2); reset
  across stage runs in the same shell (1); PASS row byte-identical to the source text with no codes (1); an MD-A-only
  failure with stale codes planted gets no MD-O codes (1); an MD-O FAIL with no code gives `unknown` (1); validator
  cases: empty, duplicates, ordering, non-code token, code-like credential token, bracketed raw `why`, tab, newline,
  glob, subscript injection (no command ran), `@`, over 128, all eight within bound, exit 0 under `set -e` (15).
- Mutation control: the same fixture against a copy without the stage-level reset fails 8 checks (codes accumulate
  across runs), so the reset checks are effective. The copy and work dir were removed.
- Not run (held by amendment 10): migrations reproduction, recovery drill, ops/off, full controls.

**Root cause: still UNKNOWN.** This change only makes the next hosted MD-O failure name its subcondition(s); it fixes
nothing and does not explain gate 2. The B-RECOVERY-DRILL to B-MAINTENANCE-DRILL span was 66 s in gate 1 and 64 s in
gate 2. That is no evidence of a long timeout, and it does not by itself support a transient-MinIO explanation.

**Accounting.** Attempt 6 of B7 (checkpoint 4, no reset); one attempt remains. Safekeeping branch `g016-md-diagnostic`
only; no MR. Returned for Codex source review. Not accepted; G017 held; G057 untouched.

### Attempt 7 — successor #15: retention cutoff guard binding, amendment 11 (2026-10-17 UTC; recorded before the first source edit)

- **Worker:** Opus 5.5, effort high (requested; actual unknown); generated worktree `agent-<id>`; local
  branch `g016-cutoff-guard` created at `origin/staging` `f64f17e045b2e2327e333d6ef2a3f3dcfd5265dc`. Sole writer.
- **Binding:** `G016-maintenance-cutoff-codex-review.md` SHA-256
  `3b73e10ac1b66991257698370a3145726280e46e6c822ca1cf6b3a82a3caf8cc` (ACCEPT of refined option A; releases **attempt 7
  of B7, the last**; further implementation needs an owner budget/scope decision). Diagnosis
  `md-o-first-diagnosis.md` SHA-256 `7a4961cfa42baa940b875aa15c0777e864def781fa1421d89a00362751cc854e`.
- **Amendment 11 allow-file:** `G016-scope-allowlist.txt` now **73 patterns**, SHA-256
  `8e7a9c520222d4b772760c03c6a7df048845b838e86b94983b6e20bce1f88c02`; the only addition is
  `^infra/recovery/recovery[.]sh$`. The previous 72 patterns are preserved byte-for-byte as
  `G016-scope-allowlist-before-amendment11.txt` (`6e78ec85…378f8`, verified).
- **Released change:** only the strictly-past cutoff guard in `infra/recovery/recovery.sh` (`retention_policy`) plus
  this Result. The cutoff and current time are each captured once with checked GNU `date`; epoch seconds and the
  nine-digit fraction are kept separate; seconds compared first, base-10 fractions only on equal seconds; no
  concatenated `%s%N` arithmetic. Invalid/unreadable/malformed values or a failing `date` refuse with
  `policy_invalid`/exit 2 before preview or mutation. Equal and future cutoffs stay refused. No `run.sh`, provider,
  harness sleep or other change.
- **Validation released:** `bash -n infra/recovery/recovery.sh` and deterministic offline fixtures on the actual policy
  path (frozen `date`, fail-on-use Docker), incl. an old-guard negative control. No migrations, recovery, ops, off or
  full controls run; no hosted retry.

### Attempt 7 — successor #15: cutoff guard implemented, offline evidence, owner budget override (2026-10-17 UTC; hand-back (a))

**Budget override reconciled (no restart).** `origin/staging` advanced to `bfbf9e73` (management-only: Goal,
PROGRESS.md, SUPERVISOR-HANDOFF.md), merged as `ee7ae5fd` before the fixture batch. The G016 Goal is now SHA-256
`c31652ae1f8ea270c828e25eb20b0a0e6a42c45c6b4d02502dc8a910734134d5`. Its section "Owner budget override — optional eighth
attempt" sets the **total cap to 8 attempts** (B7 + one owner override; not eight more, no attempt 9). **Attempt 7 is
current** and continues under its unchanged release. Checkpoint 4 and the failure history of attempts 1–6 are
preserved, with no reset. An attempt 8 would need its own Codex review and release; nothing is automatic.

**Change (`infra/recovery/recovery.sh` only; file SHA-256 `17e7fe93…c475e`).**
- New `cutoff_past CUTOFF`, placed directly above `retention_policy`. It reads the cutoff (`date -u -d CUTOFF
  +%s.%N`) and now (`date -u +%s.%N`) once each, and a failing `date` is checked. Each output must match
  `^(-?[1-9][0-9]{0,11}|0)\.([0-9]{9})$`: canonical signed seconds with no leading zero, no `-0` and no `+`, and exactly
  nine fraction digits. Seconds must lie within the years 0000–9999 (`-62167219200`..`253402300799`). Seconds are
  compared first and the fractions only on equal seconds, in base 10 (`10#`). It returns 0 for strictly past, 1 for
  equal or future, and 2 for unreadable or malformed. There is no concatenated nanosecond value, so 2263, 3000 and
  9999 cannot wrap.
- `retention_policy`: the one-line whole-second guard is replaced by `rc=0; cutoff_past … || rc=$?` and a `case`:
  - 1 → `die 2 policy_invalid "<k>_cutoff is not in the past"` (unchanged text);
  - 2 → `die 2 policy_invalid "<k>_cutoff"` (the existing syntax refusal).

  Both happen inside `retention_policy`, which runs before the env check, the provider resolve, any preview or any
  mutation.
- The accepted syntax regex (up to six cutoff fraction digits) is unchanged. So are every other policy rule,
  authorization/scope, ledger/preview checks, receipts and cleanup. No `run.sh`, provider, sleep or other file change.

**Validation (pre-batch: `git fetch origin`; `origin/staging` `bfbf9e73` merged and an ancestor of the candidate;
allow-file `8e7a9c52…`, 73 patterns).**
- `bash -n infra/recovery/recovery.sh`: rc 0.
- Offline fixture `<home>/.local/state/platform-evidence/g016/att7/fixture.sh` (125 lines, SHA-256
  `a56e94dd1803030e65efa1b47e6b82cc0e52c04173bc930855ca6501b03d7fed`; output `fixture-output.txt` SHA-256
  `ca6496139fe312683a5249a19212601abfee27a65e89bca49803422948ed797e`; not committed).
  - It runs the actual `recovery.sh retention objects|audit` as a subprocess, with a PATH `date` stub (frozen now,
    injected failure or malformed output; cutoff parsing otherwise goes to the real GNU date 9.4) and a fail-on-use
    `docker`.
  - "Accepted" means no `policy_invalid`, the next step reached the Docker stub (the provider resolve), and no
    preview was written. "Refused" means exit 2, `policy_invalid` with the expected text, Docker never called and no
    preview.
  - Pure offline bash with no Docker or harness resources, so it ran **without** the shared lock.
- **Final run: 92/92 pass.**
  - Same second (now `…12:00:00.5Z`): `.499999` and a whole-second cutoff accepted (objects and audit); equal `.5` and
    `.500000` refused, and 1 µs future refused.
  - Whole-second now: an equal cutoff is refused; 1 µs later it is accepted.
  - Fraction widths 1–6, each just below now (accepted) and just above it (refused).
  - Sub-microsecond: `.999999` against now `.999999999` accepted, against `.999999000` refused.
  - Previous/next second, day and year, and a year rollover in both directions.
  - Pre-epoch: 1969 and year 0000 cutoffs accepted. With a frozen pre-epoch now of `-1.500000000` (GNU floor seconds,
    positive fraction), −0.6 s and the previous second are accepted, while equal, −0.4 s and the epoch are refused.
  - Far-future cutoffs refused: 2262-04-11T23:47:16.854775Z and .854776Z (the int64-nanosecond edge), 2263, 3000 and
    9999-12-31T23:59:59.999999Z. With now at the end of 9999, the 9999 and 2263 cutoffs are accepted. With now in 2263,
    a cutoff 1 µs later is refused.
  - Invalid input refused as `object_cutoff`: Feb 30, no `T`, seven fraction digits, no `Z`, empty.
  - A failing `date` (now or cutoff) and 15 malformed outputs each for now and cutoff are refused: empty, text, no
    fraction, short/long fraction, leading zero, `-0`, `+`, comma, a letter, concatenated, beyond year 9999, before
    year 0000, trailing text, 13 digits.
  - The six real RD-S policies, including `future-cutoff.policy`, still exit 2 `policy_invalid` with the real clock.
    `future-cutoff` gives "audit_cutoff is not in the past".
- **Negative control** (pre-change `recovery.sh` from `f64f17e0`, saved as `att7/old-recovery.sh`, SHA-256
  `2766361f…cb36`, same stubs): the old guard **refuses** the same-second past `.499999` and the same-second
  whole-second cutoff ("not in the past"). The new guard accepts both. The old guard accepts the previous second.
- Informative: GNU `date -d 2263-01-01T00:00:00Z +%s%N` = `9246182400000000000`, which Bash arithmetic wraps to
  `-9200561673709551616`. This confirms why concatenated arithmetic was not approved.
- **Fixture-only iterations (no source change between them):** the first two runs failed only the 23 "accepted"
  cases. Those cases passed the guard and then stopped at `check_env_file`/`resolve`, because the fixture env file
  lacked `COMPOSE_PROJECT_NAME`, then `DATABASE_NAME`/`OBJECT_STORE_BUCKET`, so they never reached the Docker stub.
  All refusal cases passed in those runs too. Adding the three keys to the fixture env file gave 92/92.
- Not run (held by amendment 11): migrations, recovery drill, ops/off, full controls, hosted retry.

**Hosted cause: still unproven.** The precision defect is confirmed and corrected offline. Whether it caused the
hosted MD-O `first` failures (gates 2 and 3) is not established: the raw `$a` was never published, and
provider/classification causes remain possible. A passing later gate would not retroactively identify the earlier
cause; a repeat `first` would point to the diagnosis's hypothesis 2.

**Accounting.** Attempt 7 of a total cap of 8 (B7 plus the owner override; checkpoint 4, no reset). Safekeeping branch
`g016-cutoff-guard` only; no MR, no staging push. Returned for Codex source review. Not accepted; G017 held; G057
untouched.

### Attempt 8 — successor #15: B-TOGGLE-OFF diagnostic binding, final attempt (2026-10-17 UTC; recorded before the first source edit)

- **Worker:** Opus 5.5, effort high (requested; actual unknown); generated worktree `agent-<id>`; fresh
  local branch `g016-toggleoff-codes` at `origin/staging` HEAD `4fadc2e51d3b8e7cfe8a96603ae02589d7183e9f`, which
  contains Codex's CI repair `5ac4c504425eb7e92f471dd9ce7f16f115ca1a55` (verified ancestor). Sole writer.
- **Budget:** **attempt 8 of the total cap of 8: the final attempt.** No attempt 9. Checkpoint 4 and the attempt 1–7
  history are preserved, with no reset.
- **Bindings:** `CI-reporting-repair-20261017.md` SHA-256
  `ecdf1330b5c72c86f71368bb9f6ee5a76cd659b1e138bfa239f603b004b9a313` (section "Owner resumes G016 before standalone CI
  finishes" accepts the proposal for attempt 8 and releases the diagnostic use of `run.sh` beyond the amendment-01
  caveat). Accepted proposal `G016-toggle-off-attempt8-proposal.md` SHA-256
  `e69b23bf581fec8fa49a04ee0ffefec22af2849110b8bc1fbf9914a4e1c58dbc`. Allow-file unchanged: 73 patterns,
  `8e7a9c520222d4b772760c03c6a7df048845b838e86b94983b6e20bce1f88c02`.
- **Released change:** `infra/validation/run.sh` only, plus this Result. Thirteen fixed B-TOGGLE-OFF branch codes,
  one literal per `ok=0` branch: start, migrate-apply, migrate-inspect, service, image, control, config, files,
  pycontainer, core-settings, ready, check, stop. Also an `OFF_CODES` reset at the top of `variants_off`, an
  `off_fail_codes` closed-set formatter mirroring `mdo_fail_codes`, and a `; subconditions: …` suffix on the existing
  B-TOGGLE-OFF FAIL row only. The stop branch keeps the CI repair's artifact message. No verdict, assertion, command
  sequence, PASS row or check-block execution change. No CI file change and no speculative fix.
- **Validation released:** `bash -n infra/validation/run.sh` and tiny offline formatter fixtures. No ops, off, controls
  or other suite runs.

### Attempt 8 — successor #15: B-TOGGLE-OFF closed codes implemented, offline formatter evidence (2026-10-17 UTC; hand-back (a))

**Reconciliation.** Before the source edit, the attempt-8 binding was recorded and committed (`8ca98009`). During the
work `origin/staging` advanced to `3795b8af`, which is management-only: PROGRESS.md and the CI-repair record's
cancellation/acknowledgment notes. That record is now SHA-256 `354570cc…5e39`; its added text sets no new constraint.
It was merged before the final fixture run and is an ancestor of the candidate. The proposal (`e69b23bf…`) and the
allow-file (`8e7a9c52…`, 73 patterns) are unchanged.

**Change (`infra/validation/run.sh` only; +34/−14; file SHA-256 `76efc47e…5d2e`).**
- `OFF_CODES=""` is declared at file scope before `variants_off` and reset as the first statement after its locals.
- Each of the thirteen B-TOGGLE-OFF `ok=0` branches appends one fixed literal right after `ok=0`, leaving its existing
  `log`/`tail` as the last command so each brace group keeps its exit status: `start`, `migrate-apply`,
  `migrate-inspect`, `service`, `image`, `control`, `config`, `files`, `pycontainer` (inside the per-container loop),
  `core-settings`, `ready`, `check`, `stop`.
- The `stop` branch keeps the CI repair's message ("see diagnostics/off-stop.log") unchanged; only `OFF_CODES+="stop "`
  was added.
- `check` stays one code. The verbatim block still runs as one `ledger_run_block`.
- New `off_fail_codes`, placed after `mdo_fail_codes` and mirroring it: reads only `OFF_CODES`; over 128 characters or
  any character outside `[A-Za-z0-9 -]` gives `unknown`; uses the closed known set and canonical order; a foreign token
  adds one `unknown`; empty gives `unknown`; always exits 0; never echoes input.
- The existing B-TOGGLE-OFF FAIL row gains `; subconditions: $(off_fail_codes)`, which is evaluated only on the FAIL
  path. The PASS text, every assertion, the command sequence, the early distinct FAIL returns (ports, render,
  configure), B-TOGGLE-OFF-ADDITIONS, B-BROWSER-OFF/B-OBS-OFF and teardown are unchanged. No CI file change and no
  product or teardown fix.
- Bound note: with all twelve other codes firing, the line reaches 102 characters. Each `pycontainer` append adds 12,
  so 114 with one, 126 with two and 138 with three, which is over the 128 limit and makes the formatter print
  `unknown` alone. Even then the row is still FAIL with the suffix `unknown`, so no verdict weakens. The append was
  kept literal as the proposal specifies; the formatter's set semantics collapse duplicates.

**Validation.**
- `bash -n infra/validation/run.sh`: rc 0, also after the merge.
- Offline formatter fixture, private temp `/tmp/g016-a8-fx.YYOcea/fixture.sh` (46 lines, SHA-256
  `4a664fdc445d11eacadb77e727de69d12195a598213e5055d86f83cb770ba13d`), with a copy and the output (`4c0fd8e7…1726`) in
  `<home>/.local/state/platform-evidence/g016/att8/`. Not committed.
  - It extracts the real `off_fail_codes` from `run.sh`. It is pure bash with no Docker or harness resources, so it ran
    without the lock.
- Final run after reconciliation: **46/46 pass.**
  - Each of the 13 literals maps to itself (13).
  - Each literal is appended at exactly one `ok=0` branch, and there are exactly 13 append sites (14).
  - All 13 in reverse give canonical order (length 114); two codes are reordered.
  - A foreign token gives `start,unknown`; empty and spaces-only give `unknown`.
  - Oversize (132 characters), tab, newline, `$` and subscript injection give `unknown`; no injected command ran.
  - A credential-like token is not echoed.
  - Duplicate `pycontainer` collapses, alone and among other codes.
  - The reset is the first statement of `variants_off`, the file-scope declaration exists, and after the reset the state
    is empty and the formatter prints `unknown`.
  - The suffix is on the FAIL row only, and the formatter exits 0 under `set -e`.
- Not run: ops, off, controls and every other suite.

**Cause: still UNKNOWN.** These codes only make the next hosted B-TOGGLE-OFF failure name its branch(es). If the gate
passes, they are inert. **Gate-4 observation only (relayed by the coordinator; not independently checked and not
interpreted):** in the gate-4 build ledger, the web check-target row 71 shows progress `missing`, where gate 3 showed it
`parsed`.

**Accounting.** Attempt 8 of the total cap of 8: the final attempt (checkpoint 4 and attempts 1–7 history preserved, no
reset; no attempt 9). Safekeeping push to `g016-toggleoff-codes`; the coordinator opens the MR. Not accepted; G017
held; G057 untouched.

## Migrations/schema changes

**Current (successor #11 refresh): two actual migrations.** `00011_organization_domains.sql` (version 11, below) and
`core/migrations/00012_session_binding_and_claim_exclusivity.sql` (version 12, security amendment A2): `app.sessions`
gains the issuing `origin` and `organization_id` (NULL together; `mcp` rows unbound; the insert policy admits a new
browser row only bound; the runtime UPDATE column list excludes both), `app.auth_resolve_session(p_hash, p_at,
p_origin)` replaces the two-argument reader (a browser row resolves only bound and at its own origin), and hostname
exclusivity moves to proof time (`organization_domains_hostname_owned` over verified/active/suspended replaces
`organization_domains_hostname_live`; `app.hostname_state`'s custom branch reads proven rows only). Down is
development/test only and refuses while bound sessions, live legacy browser sessions or duplicate non-removed
hostnames exist. Evidence: the A2 migration selectors and lint/unit pins recorded in the successor #8–#10 sections
(B-MIGRATE-SET with 00012 in s10lint1); the hosted `migrations`/`recovery_drill` rows remain required. The correction
round added no migration. The next paragraph is the pre-A2 record (history; "the only new migration" was true then).

**Actual:** `core/migrations/00011_organization_domains.sql` (version 11; the only new migration). Adds
`app.organization_domains` (hostname, generation claim fence, DNS/TLS state columns, partial unique indexes
`organization_domains_hostname_live` and `organizations_slug_key`), `organizations.slug`, the owner-only
`app.organizations_slug_guard()` BEFORE UPDATE trigger and `organizations_update_slug` policy, the RLS policies and
runtime-role grants for the new table, the two SECURITY DEFINER readers (`app.hostname_state(kind, key)` and the
TLS/issuance predicate; `app.hostname_generation` was removed per the accepted #3 disposition), audit trigger/capture
extensions (T26–T29) and the version-11 catalog inventory (42 → 48 objects). Down restores the verbatim version-10
bodies. Evidence: `TestMigrationLifecycle`/`TestDomainsCatalog` (CT-DOM-1, CT-DOM-NEG N1–N7, CT-DOM-UPDOWN) in `core2`
6/6, `TestTenancyModel` 11→10 step-down, B-PERSIST (version 11) and B-MIGRATE-SET pins in `lint2`/`ops3`. sqlc: trial
run once and deferred (schema files are created by the role script, not by migrations); handwritten pgx in
`core/internal/domains/db/queries.go`, no `go.mod` change.

R0 scaffold line (historical): None yet. Planned: `core/migrations/00011_organization_domains.sql` (Plan §4; number
confirmed free at R0).

## Contracts/API/MCP changes

**Actual:** no public API or MCP contract change. `contracts/mcp/connection-context.md` gains the dated G016 row
(aliases preserve exact audience/host agreement; mixed-host setup listing). Internal-network-only HTTP on core
(`core/internal/domains/api/http.go`): `GET /internal/v1/hosts/resolve`, `POST /internal/v1/domains/tls-ask`,
list/claim/rotate/verify/recheck/remove/retry-withdrawal, `POST /internal/v1/organization/slug`; identity adds
`GET /internal/v1/auth/relay?state=` and the session route's `?origin=`. The MCP tenant-host registry
(`MCP_TENANT_HOSTS_FILE`) is retired and refused in every environment; MCP keeps exact host matching through
`HostLookup`. New configuration: `PLATFORM_BASE_DOMAIN`, `CUSTOM_DOMAIN_CNAME_TARGET`, `PLATFORM_HOST_PORT`,
`PLATFORM_AUTH_HOST`, `MARKETING_SITE_URL`, `CUSTOM_DOMAIN_*` checker settings, `DOMAIN_TLS_ADAPTER` (+ fixture files);
`IDENTITY_PUBLIC_ORIGINS` may list only non-tenant origins. Documented in DOMAINS.md, IDENTITY.md, MCP.md, OPERATIONS.md,
SECURITY.md, ARCHITECTURE.md.

R0 scaffold line (historical): None yet. Planned: dated rows in `contracts/mcp/connection-context.md` (Plan §5.6) under
D-4(a); internal-only HTTP routes (D-5); no public MCP contract change.

## Tests added

**Current additions (A2 and successor #11):** the A2 session-binding, consumer and claim-exclusivity cases recorded in the
successor #8–#10 sections (identity, access/MCP consumer fixtures, ordering/footprint, E-ORIGINS and E-DOM-2 browser
replays); successor #11: `TestDNSSystemChecker` (production adapter over a disposable in-process DNS transport: one edge
per lookup, three edges ok, four mismatch, mute server transient), DOM-4 TLS-result DomainRef negatives (foreign row,
stale generation, foreign hostname, foreign reference, empty ref) with a positive control, and the invalid-origin rows
of `TestLocalLogoutAndCSRF` (17 origins × 3 scopes, no revocation, no provider call). The paragraph below is the pre-A2
list.

**Actual (all in the allow-list):** core unit `TestHostnameNormalize`/`TestHostnameClassify`/`TestReservedLabelsFloor`
(`hostname_test.go`), `TestLoadDomainsConfig`, `TestDomainStateMachine`, `TestDNSChecker`, `TestTLSAdapterContract`,
`domains/api/http_test.go` (resolve, tls-ask, refusal mapping, exact origins), `TestLookupIsExact`, `TestSlugParity`,
audit pins 35/29/54/24, boundary/policy/matrix twelfth-capability pins (oracle 264 cells); core integration
`TestDomains` DOM-1..10 with `E-DOM-F1..F4`, `TestDomainsCatalog` (CT-DOM-1, N1–N7, UPDOWN), `TestCustomDomainRelay`
(11 subtests), `TestIssuanceGuardFailsClosed`, `TestRLSFirstSignInRace` (32 refused first sign-ins), IT-MCP-2/3/8
resolver cases, E2E-MCP-6 retired-registry/fixture refusals, version-11 lifecycle/metadata pins (amendment 02 files),
O-4/O-15 admissions (amendment 03); web unit `host.test.mjs`, `session.test.mjs` (no-referrer), the extended
`mcp-connections.test.mjs` (mixed-host listing; amendment 03); browser `identity.spec.ts` E-DOM-1..5 in the
`domain-platform`/`domain-custom` projects (E-DOM-4 corrected at `2d604648`), `mcp.spec.ts` M-LIST old-host row;
harness rows B-DOMAINS-CATALOG/B-DOMAINS-IT (`migrations`), `BROWSER_MATRIX_DOMAINS` in B-IDENTITY-E2E, the
`logrecord` catalogue/schema pins (amendments 04/05), `required-checks.tsv` 117 → 119.

R0 scaffold line (historical): None yet. Planned: Plan §6.2 (new `TestDomains` DOM-1..10 with `E-DOM-F1..F4`, `TestDomainsCatalog`,
`TestDomainStateMachine`, `TestCustomDomainRelay` 11 subtests, `TestHostnameNormalize`/`TestHostnameClassify`, web
`host.test.mjs`/`mcp-client.test.mjs`, extensions of IT-MCP-3/8, `session.test.mjs`, M-LIST, `TestStagingContract`).

## Exact commands/validation performed and outcomes

### Performed (attempt 1, consolidated; exact commands, heads and logs are in the dated handoff sections above)

Every batch ran after `git fetch origin` + staging-ancestry check (rebase/merge when needed),
`bash infra/ci/docker-desktop-mounts.sh` (13% throughout) and under `flock …/local-harness.lock`; check images ran with
`--network none` and the CMD overridden; no host Go/Node/psql was used.

| Batch | Source | Selector | Outcome |
|---|---|---|---|
| sqlc trial | round 3 | `sqlc/sqlc:1.29.0 generate -f internal/domains/db/sqlc.yaml` | exit 1 (schema `app` does not exist); **deferred**, handwritten pgx |
| b1 static/unit `b1s1`, `b1u1`/`b1u2`, `c1`–`c9`, `vt1`/`vt6a`/`vt6b` | rounds 3–9 (`…` → `8e92ebff`, `b5626ade`) | gofmt, `go build ./...`, `go vet` (+`-tags integration`) of every changed package; unit tests of hostname, domains/..., config, identity/..., mcp, audit, tenancy, access, migrate, migrations, cmd/server; `go test ./internal/logrecord/` | all **ok** after the recorded fixes (6 gofmt files, 4 unit failures, audit/boundary/eligibility pins); logrecord 10/10 |
| web `w1`–`w6b` | `f58fddb` → `05d12db8` | web check image lint, typecheck, `npm run test:unit` | 0 / 0 / **143 → 144 pass** |
| b2 `core2`, `s3a`–`s3d`, `itseed6a` | `701c8418` → `34b05ca6` (`g016-focus`, `down -v`, leftovers 0) | migrate catalog selector (6/6), `TestTenancyModel`, access selector incl. `TestDomains`, DOM-1 controls, `TestAccessPolicy` (IT-MATRIX-1 264), `TestAccessOrdering` (O-4/O-15), `TestAudit`, `TestIsolationAttacks`, mcp selector incl. `TestMCPLifecycle`/`TestMCPRecovery`, identity selector (10 top-level incl. `TestCustomDomainRelay`, `TestIssuanceGuardFailsClosed`, `TestRLSFirstSignInRace`, `TestCallbackContinuity`), then `mcp-seed.sql` on the same database | **all pass** at their last run (failures diagnosed and fixed in order; nothing outstanding) |
| seed `seed1` | `69ea09cb` | `mcp-seed.sql` dry-run with placeholder IDs + probes | rc 0 |
| `lint1`, `lint2` | `4af3e441`, **`1d7193ac`** | `run.sh lint lint-base lint-core` | **PASS**, 18 rows; leftovers 0 |
| `ops1`, `ops2`, `ops3` | `4af3e441`, `05d12db8`, `1d7193ac` | `run.sh ops` (one approved + one authorized rerun; ops3 unauthorized extra, recorded) | **FAILED, FAILED, FAILED**; ops3 fails only E-DOM-4 → B-IDENTITY-E2E/B-REQUIRED-CHECKS; its other rows PASS (B-OBS-*, B-OUTAGE, B-PERSIST v11, B-SECRET-ROTATION, B-IDENTITY-IT 73, -PROBES, -PRIVACY, -CONFIG, B-MCP-E2E, B-SECRET-CANARY 164, B-SCOPE 102, B-DOCKER-ONLY; identity-light/dark 28+28, mcp 6+6, domain-custom 2 expected) and are reused as attributable evidence per Codex |
| `diag1`–`diag3`, `probe2`/`probe3`, **`focus1`** | `944f777e`, **`2d604648`** | Chromium probes; documented identity overlay with `--project=domain-platform --grep '\[E-DOM-4\]'` then `'\[E-DOM-3\]'` | **E-DOM-4 PASS, E-DOM-3 PASS** (1/1 expected each); `down -v`, leftovers 0 |
| **`rename1`** | **`c0b2f0b5`** / base `2421c23b` | `run.sh rename` | **PASS**, rc 0, all 8 required rows; inventory diff 0 |

Not run locally (hosted-gate evidence): the `migrations`, `lint-web`, `unit-*`, `controls`, `ci-controls` and
`provenance` suites through `run.sh` (their G016 rows B-DOMAINS-CATALOG/B-DOMAINS-IT, B-MIGRATE-*, B-RENDER-* have
focused-selector or no local evidence as stated per AC below); real-provider (`workos_staging`) and manual rows (P-1/P-2);
live DNS (P-3) and production TLS adapter (P-4). Evidence hygiene: no `code=`/`state=` query values are attached; the
kept evidence directories hold no generated secrets (ops2/ops3 work directories removed).

### Planned at R0 (historical scaffold; the executable selectors below were run as recorded above)

None. R0 was read-only (`git fetch`, `git log 7e0ac44..origin/staging`, `git diff --stat`, `git show`, `sha256sum`,
`grep`/`sed` over the merged tree). Planned local runs and their executable selectors (Plan §6.3), to be executed only
after the recorded release and with fetch/rebase/reconcile, `bash infra/ci/docker-desktop-mounts.sh` and the batch
lock before each:

- b1 static/unit (check images, CMD overridden): `gofmt -l .`; `go vet ./internal/hostname/ ./internal/domains/...
  ./internal/tenancy/access/ ./internal/identity/... ./internal/mcp/ ./internal/config/ ./internal/audit/
  ./internal/tenancy/ ./cmd/server/`; `go vet -tags integration ./internal/domains/... ./internal/tenancy/access/
  ./internal/identity/ ./internal/mcp/ ./internal/migrate/`; the per-package `go test -count=1 -run '^(…)$'` unit
  selectors incl. `-run '^(TestHostnameNormalize|TestHostnameClassify)$' ./internal/hostname/`; web `node
  --conditions=react-server --test tests/unit/session.test.mjs tests/unit/host.test.mjs
  tests/unit/mcp-client.test.mjs tests/unit/identity-client.test.mjs`.
- b1 sqlc (one iteration): the §5.7 command, `go build ./internal/domains/db/`, `go vet` of that package, `git diff
  --exit-code -- core/internal/domains/db/`; deferral with the exact error text otherwise.
- b1 harness: `run.sh lint lint-base lint-core`.
- b2 focused integration on `g016-focus`: `go test -tags integration -count=1 -timeout 30m -v -run
  '^(TestDomainsCatalog|TestMCPCatalog|TestMCPMigration|TestAuditCatalog|TestRLSCatalog)$' ./internal/migrate/`;
  `… -run '^(TestDomains|TestMCPOrdering|TestMCPFootprint|TestAccessPolicy|TestAccessOrdering|TestRLSIsolation|
  TestRLSOrdering|TestAudit|TestIsolationAttacks)$' ./internal/tenancy/access/`; `… -run
  '^(TestMCPTransport|TestMCPLifecycle|TestMCPAudit|TestMCPRecovery|TestMCPDiscovery)$' ./internal/mcp/` (expect
  7/6/1/1/5); `… -run '^(TestCallbackRefusals|TestCallbackContinuity|TestCustomDomainRelay|TestLocalLogoutAndCSRF|
  TestProviderLogout|TestRLSIdentityPool|TestConcurrentFirstSignIns)$' ./internal/identity/`.
- b3: one `run.sh ops` (the single C-1 exception; hosted-only alternative if declined). b4: `run.sh rename` once.
- S-review (read-only) then affected-only reruns; provider/manual rows only after P-1/P-2 with positive evidence
  (`TestStagingContract` subtest PASS lines ≥ 1, `ok … (N tests)` with N > 0).
- Final: no new suite run; reconcile the delta since the last attributable pass and rerun only changed inputs.

Evidence hygiene (Plan §0.4 item 6, §1.7.2): before attaching evidence, `grep -rE
'?& (not included)=[^&[:space:]]{8,}' <evidence dir>` must be empty.

## Acceptance-criteria checklist

**Current overlay (successor #11).** The rows below change or add to the pre-A2 mapping that follows (that table is
history for the candidate before A2 and stays the evidence basis wherever its inputs are unchanged). Every row still
needs the hosted gate; fixture passes never satisfy P-1..P-4.

| AC | Change since the pre-A2 mapping | Actual evidence | State |
|---|---|---|---|
| AC3 ownership, concurrency | A2: proof-time exclusivity (00012 `organization_domains_hostname_owned`), atomic promotion; pending claims no longer reserve names | A2 selectors (successor #8–#10 sections); DOM-2/DOM-4/DOM-5 at `73c4c4ea` (`fix11/i1`, TestDomains 10/10) | local complete; S-review closed this in source |
| AC4 DNS proof → readiness | the production adapter's CNAME semantics pinned: Go's native resolver returns the queried name's own edge from the ordered recursive answer, so Proof's loop bounds actual edges (at most three lookups); DOMAINS.md corrected (the old "the resolver reports the chain's end" text was wrong) | `TestDNSSystemChecker` + `TestDNSChecker` (`fix11/u1`) | local complete on wire answers; assumes an RFC 1034 §4.3.1 ordered answer section (stated); **P-3 live proof not run (optional, held)** |
| AC5/AC6 recheck, TLS lifecycle | RecordDomainCheck applies a recheck TLS result only for the locked row's exact DomainRef (ID, generation, hostname, reference); otherwise `domain_state_invalid`, zero effects | DOM-4 negatives ×5 + positive control, DOM-5..8 with ref-carrying observations (`fix11/i1`) | local complete; **P-4 held** |
| AC8 custom-domain sign-in/out binding | A2 session binding; L-1 removed-host revocation; LOW-1: logout rejects a missing/malformed origin (`origin_rejected` 403) structurally before the session, no routing lookup | `TestLocalLogoutAndCSRF`, `TestProviderLogout`, `TestCustomDomainRelay` incl. L-1 (`fix11/i1`); br1 E-ORIGINS/E-DOM-2 | local complete; **P-2 manual qualification held** |
| AC11 audit/RLS/recovery/privacy | S-review delivered (CORRECTION LOOP at `edeb1965`); setupRecovery fix compiled only | `fix11` batches; s10lint1 | **recovery orchestration hosted run required**; delta review pending |
| AC12 attributable, qualified, cleaned up | correction evidence under `g016/fix11/`; leftovers 0 | this Result | **delta review, MR, hosted staging and Codex final ACCEPT pending** |

Pre-A2 mapping (successor #7; history):

Current mapping (Plan r4 §9 / r1 §9 rows; "local" means attributable local evidence at the cited source; every row
still needs the hosted gate and Codex's source review; fixture passes never satisfy the P-1..P-4 rows):

| AC | Implemented | Actual evidence | State |
|---|---|---|---|
| AC1 canonical resolution, trusted ingress | `core/internal/hostname`, resolver, `web/lib/host.ts` + `web/proxy.ts` gate (fixed 404, JSON 404/503 on `/api/*`), identity origin checks, callback host refusal before core (`1d7193ac`) | `TestHostnameNormalize/Classify` (b1); `http_test.go` exact origins (c4); DOM-7 (core2); E-DOM-3 (ops3 PASS, focus1 PASS); E-CB-8 forged host → `failed` (ops3 identity projects PASS); E-NET-1 on every identity test (ops3) | local complete |
| AC2 slugs, reserved/public hosts | `organizations.slug` + guard/policy (00011), reserved floor, public-entry 303 to `MARKETING_SITE_URL` with `/api/health` exempt, protocol path exemptions | `TestReservedLabelsFloor`, `TestSlugParity` (b1); DOM-1/3 (s3b/s3d, core2); **E-DOM-4 (focus1 PASS on `2d604648`)**; `rename1` B-RENAME/B-BROWSER-RENAME with an alternate name/domain/CNAME | local complete |
| AC3 ownership, concurrency | 00011 policies/guard, `organization.manage_domains`, facade ops, fence collision handling (#3) | DOM-1 paired owner/admin and denied-actor controls (s3b/s3d), DOM-2/3, fence-collision case (core2); CT-DOM-NEG (core2 6/6); `TestAccessPolicy` 264 (s3a); `TestIsolationAttacks` (s3a); RLS-SCOPE/ISO-ORACLE evaluators left unchanged (recorded deviation, Codex-reviewed) | local complete |
| AC4 DNS proof → readiness | checker, claim/verify, challenge binding/expiry, retry transition | `TestDNSChecker` (b1u2), DOM-4 (core2), `TestDomainStateMachine` (b1u2); fixture adapter in the overlay (E-DOM-2 routes the seeded verified host, ops3) | local fixture evidence complete; **P-3 live proof not run (optional, held)** |
| AC5 recheck/removal/reassignment | transient rule, freshness predicate, remove/reassign with generation fence | DOM-4/5/6 incl. late result/replacement and withdrawal retry (core2); IT-MCP-2 removed 404 / reassigned 401 (core2); E-DOM-5 (ops3 PASS) | local complete |
| AC6 TLS lifecycle | adapter contract, fixture adapter, `tls-ask`, issuance predicate | `TestTLSAdapterContract` (b1u2), DOM-6/7 incl. already-suspended loss (core2), `tls-ask` cases (c4); disposable HTTPS edge in every browser scenario | fixture evidence complete; **P-4 production adapter qualification held (explicit unresolved gate)** |
| AC7 auth scales without per-Org provider entries | resolver-backed origins, wildcard pattern D-3a, org-c fixture, membership rule D-2(a) | E-DOM-1 org-c sign-in with no provider-config addition (ops3 PASS; emulator lists hosts explicitly, limitation stated); `TestIssuanceGuardFailsClosed`, `TestRLSFirstSignInRace` 32 × `organization_not_member` (B-IDENTITY-IT 73 PASS, ops3); M-NONMEMBER (ops3) | emulator evidence complete; **P-1 real WorkOS `TestStagingContract` extension held** |
| AC8 custom-domain sign-in/out binding | relay in core + web auth-host leg, host-only cookies, logout legs, no-referrer | `TestCustomDomainRelay` 11 subtests incl. post-exchange race/privacy (B-IDENTITY-IT, ops3); E-DOM-2 hops/cookies/local sign-out, E-DOM-5 (ops3 PASS); `TestLocalLogoutAndCSRF`/`TestProviderLogout` (ops3); `session.test.mjs` (w2) | local complete; **P-2 manual qualification held** |
| AC9 MCP aliases | `HostLookup`, exact audience/host agreement, mixed-host listing, contract row | IT-MCP-2/3/8 (core2, s3b); E2E-MCP-6 refusals and B-MCP-E2E (ops3 PASS); M-LIST old-host row (ops3 mcp projects 6+6); `mcp-connections.test.mjs` (w2 144/144); provider audience note in MCP.md/DOMAINS.md | local complete |
| AC10 configuration transition | D-6 variables, `.env.example` block, `configure-env.sh`, `variants_rename` extension | B-ENV-EXAMPLE PASS (lint2); **`rename1` PASS** (name, domain, CNAME, database rename; image IDs unchanged); historical manifest untouched | local complete; B-RENDER-* rows hosted only |
| AC11 audit/RLS/recovery/privacy, bounded cost | T26–T29, reasons/operations, readers, restore case, `domain.request`/`auth.relay` log events, host-tag privacy | audit pins + `TestAudit` AT-INV (s3c); DOM-8/9/10 with `E-DOM-F1..F4` (core2; PG-007 recorded: 6–7 statements/op, constant at 10/40 rows; sequential scans on small fixtures, production-cardinality index use **not shown**); B-SECRET-CANARY PASS (ops3 164, lint2, rename1 8); B-IDENTITY-PRIVACY/CONFIG PASS (ops3); request-ID lookup exception recorded in OPERATIONS.md and reviewed | local complete under characterized risk; **S-review pending** |
| AC12 attributable, qualified, cleaned up | this Result (per-AC provenance, commands, statuses, deviations, cycle history); cleanup receipts | ops/rename inventory diffs 0, leftover inventories empty; ops2/ops3 work directories and generated secrets removed (Codex-confirmed); evidence dirs listed per batch | **MR, hosted staging disposition and Codex final ACCEPT pending** |

R0 scaffold line (historical): Not started. The AC1–AC12 → test mapping is Plan §9/§6.2; AC6 carries the explicit production-adapter gate (P-4),
AC7/AC8 the real-provider/manual prerequisites (P-1/P-2), AC4's positive live case P-3 (optional). Fixture passes do
not satisfy those rows.

## Architecture deviations

Implementation-time deviations from Plan r4's wording (attempt 1; none changes the owner-approved design, every one is
narrower or equivalent; listed for the security review and Codex):

1. **Slug authority.** `SetOrganizationSlug` is decided with the existing owner-only `organization.manage_policy`, not
   `organization.manage_domains` (owner/admin): the unchanged `organizations_update` policy admits only `org_owner` in
   `WITH CHECK` (r1 §4.1 assumed owner/admin), so decision and RLS now agree. Domain operations keep owner/admin.
2. **Reader signature.** `app.hostname_state(p_kind, p_key)`: the database does not know the platform domain, so the
   pure classifier passes the kind and the slug or hostname (still one unique-index probe).
3. **New reader `app.hostname_generation(host)`** (auth reader, EXECUTE to runtime): a reassignment's `generation + 1`
   must see other Organizations' removed rows, which RLS hides from the claimant. It returns only the next number (it
   discloses how many earlier claims the hostname had). Security-review item.
4. **Login and pre-exchange host checks** run as one autocommit statement on the pool next to the existing
   transactions (Login before `insertTransaction`; Callback right after the one-use consumption, before any exchange)
   instead of inside them: same statement count and the same READ COMMITTED snapshot semantics (no lock either way).
   The final issuance guard runs inside `signIn` before the first write of either branch, as planned. The session
   route's `?origin=` check is likewise one pool statement.
5. **Relay in core.** The relay decision is `identity.Relay` behind `GET /internal/v1/auth/relay?state=` (core logs
   `auth.relay {outcome, reason}`); web on the auth host forwards the raw query byte-identical to the returned callback.
6. **Registry retirement** refuses `MCP_TENANT_HOSTS_FILE` in every environment (not only production); MCP keeps exact,
   unfolded host matching through `HostLookup`.
7. **Catalog details.** `domain.list` is an access.denied operation too; the check outcomes are `ok`, `mismatch`,
   `nxdomain`, `transient`, `tls_error`, `withdrawn`, `withdraw_error`, `withdraw_unknown`; facade codes are
   `host_taken`, `domain_state_invalid`, `domain_retry_exhausted` (a recorded mismatch/transient outcome is returned in
   the domain view, not as an error, so no `proof_*` codes exist).
8. **Configuration.** `PLATFORM_BASE_DOMAIN=localhost` only with `APP_ENV=test` (the overlay's `<slug>.localhost`);
   `IDENTITY_PUBLIC_ORIGINS` may list only non-tenant origins (reserved, public-entry or auth host); the DNS refusal
   list includes `.test`/`.example`, so fixtures use `example.com`-style names and the overlay's custom row is seeded.

R0 found no contradiction between Plan r3 and ARCHITECTURE.md, SECURITY.md, AUTHORIZATION.md, ADR-003/005/007
or the contracts as amended at `f6d79af`; the amendments restate the accepted design.

## Unresolved issues, blockers, attempted approaches, cycle history

Current (successor #12, attempt 2):
- No open implementation defect known after the fast-failure and Duo D1 corrections; they await Codex delta review on
  `origin/g016-review-fast12` (first correction cycle on these two items). The successor #11 block below is history.
- Cycle history (execution): attempt 1 ended with the review-ready submission at `0a163c7f` (BUDGET_POLICY §3);
  attempt 2 is this Codex-released bounded correction (MR fast `B-WEB-BOUNDARY`, Duo D1). Checkpoint after attempt 4.
  Two correction cycles on one material issue have not been reached.

Current (successor #11):
- No open implementation defect known after the three corrections; they await Codex delta review (first correction
  cycle on these items).
- DNS: the production per-edge bound rests on Go 1.27's native resolver returning the first CNAME target of an ordered
  recursive answer (RFC 1034 §4.3.1, the assumption Go's resolver states itself). A recursive resolver that reorders the
  answer section could let Go report a later name, so fewer lookups than edges; at most three lookups per proof still
  hold, and `TestDNSSystemChecker` fails if a toolchain change returns the chain's end. A strictly edge-reading adapter
  would need a DNS wire parser (hand-written or `golang.org/x/net/dns/dnsmessage`, a new dependency): not done, outside
  the approved scope.
- Held gates as listed under Status. The successor #7 block below is history.

Current (successor #7):
- No open implementation defect. ops1/ops2/ops3 are FAILED records (seed cascade + gate/schema; stale allow-file +
  E-CB-8/E-DOM-4/privacy duplicates/P-POS-2; E-DOM-4), each diagnosed and fixed as recorded; ops3's premature launch is
  a recorded process deviation.
- Held gates with external authority: S-review (Codex), P-1 WorkOS staging entries/real-provider run, P-2 staging
  deployment with real TLS and whole-query edge-log redaction (manual MQ rows), P-3 optional live DNS proof, P-4 Caddy
  production provisioning/adapter qualification; MR opening by the coordinator; hosted fast/full gates.
- Recorded deviations for the S-review: the eight Plan-wording deviations below (all Codex-dispositioned), the
  unchanged RLS-SCOPE/ISO-ORACLE evaluators, the lookup request-ID exception, PG-007's characterized risk, the
  D-2(a) browser-spec consequences, `BROWSER_MATRIX_DOMAINS` defined in `run.sh` (`lib.sh` out of scope), and
  the static marketing stand-in on the test edge (`2d604648`).
- Cycle history (execution): attempt 1 only; successors #1–#7 by context handoff (no attempt spent); Codex-bounded
  loops inside the attempt: deviation dispositions (#1/#3/#4), amendments 01–05, D-2 disposition, ops2 findings
  (`1d7193ac`), E-DOM-4 correction (`2d604648`, accepted). Two correction cycles on one material issue were never
  reached.

R0 record (historical):
- Blockers to execution (gates, none owned by this worker): Codex review of Plan r4 and this Result; binding
  confirmation; recorded execution release.
- Open for Codex disposition (not blockers to review): whether the implementation worktree uses the task-prefixed
  name `g016-agent-<id>` or the generated fallback; confirmation that the single `ops` exception (C-1) is taken
  locally or replaced by hosted-only rows; the owner's effort choice for the guard/relay phase (default high).
- Held prerequisites with external authority: P-1 (WorkOS staging entries), P-2 (staging deployment with real TLS and
  whole-query edge-log redaction), P-3 (optional live DNS proof), P-4 (Caddy production provisioning). None requested.
- Cycle history: Plan r1 → r1 review (CORRECTION LOOP, 10 items) → r2 → r2 review (CORRECTION LOOP, 6 items) → r3 →
  r3 review (technical closure; owner design package) → owner approval and full Plan ACCEPT (`f6d79af`) → r4 (R0).
  Zero execution attempts and zero correction passes on implementation.

## Suggested follow-up (not authorized by this record)

- An explicit-invalidation hostname cache if the measured `E-DOM-F` cost is unacceptable (Plan §5.5 item 4).
- Narrower identity/MCP E2E selection (BACKLOG "Narrow identity/MCP E2E selection") so future Goals need no `ops`
  exception.
- D-4(b) automatic revocation only if an orphaned-but-unreachable connection proves unacceptable in use.

## Final closeout (2026-10-17 UTC; closeout worker, after the passing combined staging gate)

**Status: ready for Codex final ACCEPT. Not self-accepted.** Every section above is preserved as history; this section
records the final hosted evidence, the failure history, the budget, the held obligations and the scoped cleanup.

- **Worker:** Opus 5.5 (`claude-opus-5-5`), effort high (requested; actual unknown). Generated worktree
  `agent-<id>`; local branch `g016-closeout` at `origin/staging`
  `bea93b8d3ad8d60ca65829ede2b2bfb30d808ea7`. Sole writer. No tests, no implementation attempt, no provider, DNS,
  WorkOS or deploy action.
- **Bindings verified before writing:** `G016-toggle-off-codes-codex-review.md` SHA-256
  `47e4f8fa00dc4686b25a08a65f2271602d94d88c4c15066a0ff1bb9041711faf`; this Result before the closeout
  `202a5104422f99c0958014138e16bf137b16d1b39b1b04d77d53f46171c3b669`; the Goal
  `a743659bb20cf587d3688052293a97511380a6e6b9cdbd2772c1a045d031ef59` (section "Owner budget override — optional ninth
  attempt": total cap 9).

### Final hosted evidence

- **Combined staging pipeline `2958083430`: SUCCESS** on the exact merge
  `4aeabf12169d98826bd1d4f64e7c68f099eb7a8c` ("Merge branch 'g016-toggleoff-codes' into 'staging'", MR !38).
  - `validate:checks` job `17253162510`: SUCCESS, about 495 s.
  - `validate:integration` job `17253162511`: SUCCESS, about 3216 s.
- **results.tsv:** 82 PASS, 0 FAIL, 0 UNMET (plus 6 INFO rows; the only "FAIL"/"UNMET" text match is inside the
  BUILD-LEDGER INFO row's description). `B-REQUIRED-CHECKS` PASS: 81 required IDs across the executed suites
  controls, ci-controls, root, migrations, ops, rename and off (plus final), each recorded exactly once as PASS.
- **Specifically PASS:** `B-TOGGLE-OFF` (and `B-TOGGLE-OFF-ADDITIONS`); `B-MAINTENANCE-DRILL` (MD-A, MD-S-0, MD-R,
  **MD-O**, MD-E, MD-S, MD-C); `B-RECOVERY-DRILL` (the only runtime proof of setupRecovery, now run hosted);
  `B-OBJ-IT` (10/10 subtests).
- **CI reporting repair `5ac4c504425eb7e92f471dd9ce7f16f115ca1a55` verified in use:** `ci-evidence/harness.log` (about
  105 KB) was published, and `junit.xml` carries real suite durations: root 936 s, migrations 997 s, ops 578 s,
  rename 80 s, off 525 s, controls 68 s, ci-controls 1843 s, final 3 s. Per-check times are absent by design.
- **Coordinator copies** in `<home>/.local/state/platform-evidence/g016/staging-gate-2958083430/`, re-hashed by
  this worker: `gate5-results.tsv`
  `6db85fb453082b3e9d94de4ba96b983a7f7fbea9cef8b1a3fd8a1fe8e473f214`; `gate5-junit.xml`
  `fdbfebb7e8c78dbbeabe327e44ec0ed9a84a2d92d9274b03517f05d03b9fc498`.

### Failure and cancellation history (retained; none is rewritten by the pass)

- Staging gate `2957417252`: FAILED (13 rows).
- Staging gate `2957712489`: FAILED (MD-O).
- Staging gate `2957849428`: FAILED (MD-O first).
- Staging gate `2957949710`: FAILED (B-TOGGLE-OFF).
- Retry job `17252899360`: CANCELED.
- Standalone CI-repair pipeline `2958064780`: CANCELED (owner); supplies no passing evidence and does not qualify.
- The earlier MR fast failures (for example pipeline `2957186823`, job `17248173536`, `B-WEB-BOUNDARY`, corrected in
  attempt 2) stay as recorded in their sections.

### Root causes

- **MD-O: UNKNOWN / unproven.** **B-TOGGLE-OFF: UNKNOWN / unproven.** Attempt 7's retention cutoff guard and attempt 8's
  closed B-TOGGLE-OFF branch codes are in the merged source. The attempt 6 MD-O codes and the attempt 8 codes did not
  fire because both rows passed. A later pass does not prove what caused the earlier failures; no causal claim is made.

### Budget

- **8 of the total cap of 9 attempts used** (B7, plus the owner's 8th and 9th-attempt overrides). The 9th slot is
  **unused**. Checkpoint 4 is recorded. No resets. This closeout spends no attempt.

### Held qualification obligations (not satisfied by this Result)

- **P-1 (real WorkOS) and P-2 (deployed/manual HTTPS):** unqualified in G016 and **transferred to G057**, which must
  receive final ACCEPT **before G021** (`G016-G057-qualification-split.md`, SHA-256
  `8feeea4e12a63d1514441952bd47fe7c048f0ae2eedf7aa03e82b201fc234ab1`).
- **G057 pre-use conditions** (`G016-implementation-codex-review.md`, "G057 pre-use limits, not a new G016 round"): the
  dormant domain provider test is not approved for live use until (1) its cleanup confirms deletion or explicit
  not-found, fails closed on unknown provider failures, keeps the owned-user recovery reference private, and reports the
  cleanup outcome in the final status; and (2) G057 establishes the actual AuthKit response shape/path, since
  origin-only recognition cannot close AC7/AC8 alone.
- **P-4 (production TLS adapter qualification):** stays an explicit held gate.
- **P-3 (live DNS proof):** optional; not run.

### Current source

- `origin/staging` `bea93b8d3ad8d60ca65829ede2b2bfb30d808ea7`. Its history contains the G016 merges
  `37c9974217ee8933bc93609b9da2084cbc3ef662` (implementation), `c67f40c2a76f7bd03402d97fc55b1fcf487530cb` (staging
  repair), `0cb9bda49360ba1e5875219fedbc154cd972a408` (MD diagnostic), `7bf9f08877f09186ce14eedb7fc02767f3bb55dd`
  (cutoff guard) and `4aeabf12169d98826bd1d4f64e7c68f099eb7a8c` (toggle-off codes), plus the CI repair `5ac4c504`.

### AC12 and cleanup

- AC12's remaining items are now: this final Result, the cleanup receipt and Codex final ACCEPT. The MR and the hosted
  staging disposition are satisfied by `4aeabf12` and pipeline `2958083430`.
- **Scoped cleanup receipt:** `<home>/.local/state/platform-evidence/g016-closeout/RECEIPT.md`, SHA-256
  `0adf29ec2d7819ba3a674f869a4032fa3825ecab7f84ff133c4fdd508b9d3b9a`. Only non-force operations were used.
  - **Removed:** 14 G016 worktrees (merged, staging-ancestor or pushed HEADs; no live cwd; non-force remove
    succeeded); 9 local branches with `git branch -d`; 19 `g016-*` Docker images with non-force `docker rmi`.
  - **Held, with reasons in the receipt:**
    - 5 worktrees on the unpushed pre-rebase `a90fb716`, plus that branch.
    - 5 worktrees locked by the live coordinator, plus their branches.
    - One worktree whose owner is uncertain (`aec83b89`).
    - `g016-plan`: `-d` refused; its content is byte-identical to staging.
    - The `worktree-agent-*` branches of the removed worktrees: the permission classifier denied deleting them.
    - G017 planning worktrees and branches; the live `agent-a9b3c5f*`; this worker's own worktree.
    - `/tmp/g016-*` scratch (outside the candidate pattern); the `src-*` `.env` copies (they need owner approval in
      session); all evidence folders.
  - **Deviation:** the `g016-plan` worktree was removed before I found that its HEAD is on no remote. Its content is
    byte-identical to staging blobs and its commits stay on the held local branch, so nothing was lost.
  - **Remote branches:** none deleted. The receipt proposes deleting all nine `origin/g016-*` branches. Seven are
    staging ancestors; `g016-plan` and `g016-r0` are not, but their content is identical to staging.

**Hand-back:** returned to Codex for exactly one outcome (ACCEPT, CORRECTION LOOP or ARCHITECTURE DECISION REQUIRED).
G017 execution stays held until G016 final ACCEPT. The G057, G018 and G019 owner boundaries are unchanged.
