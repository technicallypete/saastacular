> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G016 Implementation Plan — Hostname resolution and verified domain lifecycle

Revision r4 (R0; narrow), 2026-10-16 UTC. Author: Claude (Fable 5.1, planning effort high as instructed by the
coordinator; the runtime setting is not verifiable from inside the worker). Status: **R0 reconciliation of the
accepted r3 design (r3 review and owner approval (not included): full Plan r3 ACCEPT,
design approved by the owner, execution binding/release pending) against the merged and accepted G015 code on
`origin/staging`, proposed for Codex review together with the initial Result.** Not an execution binding and not
implementation. Tier: architecture/security/data (Goal Execution Policy).

**r4 revision note (R0; narrow).** r3 is archived byte-exact as r3 (not included)
(SHA-256 `3b29275ea8c914178138889e29e3a7412209be1378ae1c1019904e08809c1ffc`, the accepted bytes). Base moves to
`origin/staging` `f6d79af5cc97e0cf6c17b5a9f649a2ad0d4cf827`, which contains the merged G015 (merge `7e0ac44`, source
head `b80c283`, final ACCEPT and EPIC-005 closeout at `443c66a`) and the owner-approved design record (`f6d79af`:
ADR-007, the Spec's owner-selection section, the r3 review's acceptance section, the architecture/security/PG-007/
contract notes and the imported scope allow-list). **R0 finding: nothing material.** No design section of r3 is
rewritten; only §0 (gates, R0 record, owner-decision record), §2 (rebind), §11.2 (pins) and §13 change.

| R0 item | r4 change |
|---|---|
| 1 P-G015-1..7 and §0.3 against the merged code at `f6d79af` | §0.2.1: every row **confirmed** with file:line; immaterial drift only (the two `Registry.Lookup` sites moved from `handler.go:138/218` to `143/223`; `newServer` is declared in `server.go:63` and called at `handler.go:336`). 00010 is the last migration; **00011 is free**. Audit catalog 31 (25 trigger-owned + 6 core, `audit_test.go:75`). No Organization slug column exists. `TestMCPCatalog` migrates to its own version 10 |
| 2 Changes since the r3 base `36c7f4e` | code: only the G015 merge (18 files, already assumed as the predecessor). Management/docs: ADR-007, Spec owner section (new digest), ARCHITECTURE/SECURITY/PERFORMANCE (PG-007)/DECISIONS/ADR-003/ADR-005/`connection-context.md` notes, compacted PROGRESS, the G015 Result/review/receipt, the r1/r2 archives and the scope allow-list. None contradicts r3; the notes restate r3's own design |
| 3 Owner decisions | §0.4 records D-1..D-4 as decided by ADR-007 and the two Codex implementation constraints from the r3 review (entire callback query out of logs; `ResolveWith` reaches identity through identity's own interface, no identity→domains import). §1.9 stays as the historical brief |
| 4 Binding | §2 rebinds every input at `f6d79af` with its SHA-256; the allow-list imported as `management/v0/results/G016-scope-allowlist.txt` is byte-equal to r3 §11.3 (44 patterns) |
| 5 Count pins | §11.2 re-pinned: harness-derived values confirmed from `run.sh`/`required-checks.tsv`/`validate-controls.sh`/`lib.sh` content or byte-equality with the G015 R0 rows |
| 6 Process | r3 and r4 by the successor worker (§2 planner record); the initial Result is drafted at `management/v0/results/G016-hostname-resolution-domain-lifecycle-result.md` |

**r3 revision note (retained).** r2 is archived byte-exact as r2 (not included)
(SHA-256 `3b6ea4895e6195065406c08eeb385aab9508e49fc2c589501ff9d2d277a1b702`, the reviewed `4de99c9` bytes). Base
moves to `origin/staging` `36c7f4eb5ed6d8b2aa6fef438e2705300ee5bffa` (delta from `b8ceeab`: management only — the r2
review and the G015 Duo acceptance). Only the sections named below change; everything else is r2. Facts cited for
items 3–6 come from read-only inspection of `access.go`, `boundary_test.go`, `identity/{service,store,http}.go`,
`web/lib/session.ts`, `web/app/auth/callback/route.ts`, `infra/identity/proxy.conf` and `mcp_integration_test.go`
at the base; no code, build, test, Docker, provider or settings action was taken.

| r2 review item | r3 change (section) |
|---|---|
| 1 Already-suspended proof loss | §1.3 issuance wording; §1.4 table gains `suspended → recheck confirmed proof loss → suspended` with `verified_until = now()` (issuance stops immediately, not at the previous deadline); §4.3 guard edge widened to that source state; DOM-6 case added (§6.2) |
| 2 Terminal withdrawal retry | §1.5 E2 "retried by an explicit recheck" replaced by the explicit, bounded, operator-invoked `RetryWithdrawal` on the removed row's own generation/provider resource (§1.4 removed-row rows, §5.3 route, §5.4 bound, §4.3 columns); never restores routing; no reconciler or retry framework; DOM-4 cases (§6.2) |
| 3 Facade wiring scope | §5.0.1 pins the exact `access.go` change (one `domainHooks *domainHooks` field beside `mcpHooks`; constructors untouched; hooks set only in-package by tests; nil in production), the `TestProductionImporters` case edits, and adds `access.go` to the allow-list (§11.3); the rest of the `access` package stays out of scope |
| 4 Actual lookup transaction inventory | §5.5 withdraws the "one query" framing: the resolver has no transaction of its own; `ResolveWith(ctx, q, host)` is one statement on the caller's `pgx.Tx` (identity/session paths, the E2 `principalSubject` pattern) or one autocommit statement on the pool (MCP, `hosts/resolve`, `tls-ask`: one connection acquisition, no BEGIN/COMMIT/`set_config`); the per-operation table is re-counted on that basis; footprint lines `E-DOM-F1..F4` defined on the existing `e.hook.counts()`/EXPLAIN/pg_locks mechanisms (named apart from r1's browser cases E-DOM-1..5) |
| 5 Final session-issuance guard | §1.7.1: inside the existing `signIn` transaction, after the exchange and before any write, re-resolve T → the transaction's Organization (`host_changed`) and check current membership (`organization_not_member`); snapshot reads, no new lock, no network inside any transaction; declared ordering against concurrent removal/revocation; the pre-exchange check stays as the cheap early refusal; `TestCustomDomainRelay` extended, no new suite (§6.2) |
| 6 Relay privacy and refusals | §1.7.2: `Referrer-Policy: no-referrer` added to the web form layer's `NO_STORE` header set (relay 303, fixed 404, callback 303, ambiguous 400 all inherit it); core JSON stays no-store; overlay edge keeps `access_log off`; production edge query redaction for `code`/`state` recorded as a D-1/P-2 deployment input; the relay forwards the raw query byte-identical so G004's `access_denied`/provider-error/ambiguous handling on T is unchanged; evidence hygiene rule |
| Settled by the r2 review | C1 (one justified `ops` exception; no duplicate final suites), C2 (two domain check IDs), C3 (one sqlc trial with fallback), D5, D6 — recorded in §0.1 and §12; C4 waits for D-4; C5 goes to PG-007 wording |
| Owner design package | §0.1/§1.9: the single owner question (Caddy target, membership, shared relay, retained MCP rows refused while unresolved, DB lookups with bounded measurements) is **pending**; D-1..D-4 stay open with no answer assumed; no live DNS, WorkOS, provisioning or deployment authority exists |

**r2 revision note (retained).** r1 is archived byte-exact as r1 (not included)
(SHA-256 `38c736e711d4284718381f46aa5774564e32f0bf346060e531c3dcc12abe711c`, the reviewed `c1532d0` bytes). Base moved
to `origin/staging` `b8ceeab5895a2eef551b99fbddfbcc67072c262d` (delta from `d36f2a7`: management only — the r1 review,
the G015 Duo disposition and its Spec clarification).

| r1 review item | r2 change |
|---|---|
| 1 Consistent routing and issuance readiness | §1.3 pins one routing predicate (current proof **and** valid stored TLS readiness/expiry **and** not removed) and one issuance predicate (fresh proof, not removed) evaluated by the same reader; no per-request network; §1.4 `active` requires a non-expired certificate; DOM-6 extended with TLS expiry (no new suite) |
| 2 Executable lifecycle/guard rules | §1.4 and §4.3 now state the exact allowed column sets per state (pending failures stay `pending`; same-state refresh and transient recording allowed with `version + 1`; removed rows accept only the terminal cleanup outcome columns); the challenge rule is explicit (pending-only expiry and rotation; the TXT value is the ongoing proof for rechecks; freshness renewal independent of the challenge expiry) |
| 3 Generation-bound external completion | §1.5: every adapter call carries `DomainRef{ID, Generation, Hostname, Reference}`; results are applied only to that row id/generation under the `version` guard; `Remove` uses the row's own provider reference (E2) or the issuance predicate (E1), never name-only; timeout/unknown pinned; E1 withdrawal vs E2 deletion semantics stated. No controller framework |
| 4 Complete B2 contract | §1.7 rewritten as **B2 = shared callback relay**: both legs, durable `login_transactions` columns (§4.6; RLS unchanged), original canonical Organization recorded at login, code-exchange/one-use ordering unchanged from G004 and executed on the initiating host with its `__Host-login` binding (final browser proof), reassignment/membership re-check at consumption, session issuance, identity conflict/assurance, audit/recovery and reference privacy; B1 kept as the explicit alternative |
| 5 Historical alias setup listing | §1.6 and §3/§11.3 add the minimum web parser/internal-contract change so a mixed list (old-host rows) lists and revokes through a current host; new URLs keep the exact-origin check; MCP execution keeps stored host/audience checks; IT-MCP-3/8 and a new web unit test pin it |
| 6 Acyclic construction boundary | §5.0 dependency/construction path: `hostname` ← `domains` ← `tenancy/access` ← `domains/api`; `TestProductionImporters` pin widened; `HostLookup.Lookup(ctx, host) (id, error)` distinguishes `ErrUnknownHost` from store failure (503, never a clean unknown host) |
| 7 Executable sqlc trial | §5.7: `-f internal/domains/db/sqlc.yaml`, `--user "$(id -u):$(id -g)"`, one bounded compatibility/generation/compile trial, concrete handwritten fallback, no day-long campaign |
| 8 Attributable PG-007 cost | §5.5 inventories round trips, context/liveness/facts/lock statements per operation and the resolver call separately; the former "≤ 6" claim is withdrawn; query/index/lock concerns listed for owner attention |
| 9 Focused execution and real provider selectors | §6.3: no duplicated final suites; changed web unit files run directly with `node --conditions=react-server --test <files>`; vet scoped to changed packages; the single `ops` exception records its HTTPS fixture dependency and smallest alternative; `TestStagingContract` selector corrected to `./internal/identity/workos/` with positive non-zero evidence |
| 10 Truthful qualification boundaries | §6.4 separates the AC6 production-adapter gate from AC7/AC8 provider/manual requirements and lists remaining prerequisites with the authority each needs |
| D-5/D-6 settled by the review | §0.1/§1.9: M2 and the reviewed domain block are recorded as defaults; the owner brief covers D-1..D-4 only |
| G015 Duo correction (staging `b219ef4`, branch `904de79`) | §0.2: P-G015-4 (owner/admin discovery is a role-level hint without target existence; `organization.manage_domains` follows the same rule) and P-G015-7 (the seam is a per-Handler `atomic.Pointer[catalog]`, no package-global state) reconciled |

- Inputs: Goal `management/v0/goals/G016-hostname-resolution-domain-lifecycle.md` and Spec
  `management/v0/specs/G016-hostname-resolution-domain-lifecycle-spec.md` (unchanged between `b8ceeab` and `36c7f4e`;
  digests §2). Predecessor baseline: G015 draft MR !33, pre-integration ACCEPT at
  `c7ce0bb3eed25c32199ef531fb4efaa364212464`; Duo correction accepted by Codex at staging `36c7f4e` (branch tip
  `904de79ad1f11fbf58d80c577a6b83041783498a`; integration and the staging gate still ahead, re-pinned at R0).
- Tier and budget: architecture/security/data; **B7, checkpoint after attempt 4**. Planning spends no attempt.
- Planning rules honoured: inspection and documentation only; no install, generation, build, test, Docker, provider
  call, DNS/TLS probe or settings change.

## 0. Gates, predecessor reconciliation and provisional interfaces

### 0.1 Gates

| Gate | Status at r4 | Owner |
|---|---|---|
| G015 final ACCEPT (MR !33 merged `7e0ac44`, pipeline 2954782898 SUCCESS, final ACCEPT and EPIC-005 closeout `443c66a`) | **closed** | — |
| G002 final ACCEPT | accepted and integrated | — |
| R0: re-read P-G015-1..7 and §0.3 against the accepted staging content; re-pin migration number, counts, fixture shapes | **done in this revision (§0.2.1, §11.2): nothing material** | Claude; reviewed by Codex |
| Codex technical closure of r3 (six items) and **full Plan r3 ACCEPT** | **closed** (`f6d79af`, record (not included)) | — |
| **Owner design package** | **approved** by the owner ("Approve those details"), recorded in [ADR-007](../../../adr/ADR-007-verified-domain-routing-and-authentication.md) (ACCEPTED) and the Spec's owner-selection section; **design only** — no live DNS, WorkOS settings, credentials, provisioning or deployment authority (§0.4) | — |
| Owner decisions D-1..D-4 | **decided** by ADR-007 (§0.4); §1.9 is retained as the historical brief | — |
| D-5 (M2 internal routes) and D-6 (reviewed domain-block transition) | settled (r1 review; reconfirmed r2/r3) | — |
| C-1, C-2, C-3 | settled by the r2 review (§12) | — |
| C-4 (exact-audience retention and mixed-host wording) | **resolved by D-4 = (a)** (ADR-007 item 4): retain, refuse while unresolved, listable/revocable through a current host, no retargeting, no automatic revocation; the `connection-context.md` dated rows and the MCP.md note follow that disposition | — |
| C-5 (PG-007 wording) | PG-007 entry recorded by Codex at `f6d79af`; the Result adds the measured `E-DOM-F1..F4` values at delivery, no performance guarantee, no authority cache | Claude at delivery |
| Codex implementation constraints from the r3 review (§0.4 items 6 and 7) | recorded; honoured at implementation without a further Plan round | Claude |
| Prerequisites P-1..P-4 (§6.4) with their authority | open; nothing performed; none authorized | owner/operator |
| Codex review of this r4 (R0 record) and of the initial Result | **this hand-back** | Codex |
| Git-first binding: allow-list (imported, byte-equal to §11.3), bound inputs (§2 at `f6d79af`), initial Result | prepared; Codex imports/confirms | Codex |
| Budget B7/checkpoint 4 (0 used), resource readiness, single writer, recorded execution release | after the above; implementation does not start before the recorded release | Codex (supervisor) |
| G017 planning trigger: this Goal's first implementation MR; G018+ not authorized; G021 closed | — | — |
| G017 planning trigger: this Goal's first implementation MR; G018+ not authorized; G021 closed | — | — |

### 0.2 Provisional G015 interfaces (P-G015-n)

Read-only inspection of `c7ce0bb` (17 paths, +1687/−46) and the Duo correction at `904de79` (`catalog.go`,
`handler.go`, `policy.go`, `mcp.go`, `matrix_test.go`, fixtures and tests; no migration, no identity/web/compose
change).

| # | Interface relied on | Where | G016 use / R0 check |
|---|---|---|---|
| P-G015-1 | `(*Handler).newServer(actx, conn)` registration/middleware; host lookup precedes it at `handler.go` (`/mcp/c/` and metadata paths) and in `connections.go` (`LookupOrigin` in `begin` and `Create`) | `core/internal/mcp/{server,handler,connections}.go` | G016 replaces the registry lookups with `HostLookup` (§1.6, §5.0); `newServer`, `catalog.go`, `tools.go` untouched |
| P-G015-2 | `setupMCP(t) *mcpEnv` with the G015 helpers; `e.handler(t, map[host]OrgID, verifier)` builds a registry | `core/internal/mcp/env_integration_test.go` | G016 extends the same fixture (seeded slugs/domains; a custom alias host; a removed/reassigned host); rebased at R0 |
| P-G015-3 | `run.sh` `mcp_it` pins: Transport 7, Lifecycle 6, Audit 1, Recovery 1, **Discovery 5**, Ordering **10**, Footprint 1; E-MCP lines **5** | `infra/validation/run.sh` | G016 edits the same driver (IT-MCP-2/3/8 wording); merge risk at R0 |
| P-G015-4 | Canonical capability set, `capabilityKind`, `eligible`, `factQueryFor`, `ToolRequirement`; **Duo clarification (Spec, `b219ef4`)**: Organization owner/admin discovery paths are role-level hints after the context gate and assert no target existence (literal empty-Organization case in `TestToolEligibility`) | `access/policy.go`, `mcp.go`, `matrix_test.go`, `boundary_test.go` | `organization.manage_domains` (owner/admin, Organization scope) is a role-level capability with no target; it joins the same tables, pins and the empty-Organization case. Discovery never grants an effect; every domain operation re-decides |
| P-G015-5 | No migration, audit action or log operation; catalog 31 actions; `00010` last | `core/migrations/`, `audit.go` | **00011 provisional**; 31→35 is G016's delta |
| P-G015-6 | `contracts/mcp/control-plane.md`, MCP.md, AUTHORIZATION.md G015 wording (the output-schema claim qualified by the Duo correction) | those paths | G016 appends host rows to `connection-context.md` only (§5.6) |
| P-G015-7 | **Test seam is per-Handler**: `Handler.fixture atomic.Pointer[catalog]` (nil in production; no package-global state, no mutex; `Handler` values are never copied) | `core/internal/mcp/{handler,catalog}.go` | G016 adds no MCP fixture seam; its `domainHooks` is one field of the access `Service` beside `mcpHooks`, with the exact wiring pinned in §5.0.1; the MCP handler is not copied by the `HostLookup` wiring |

#### 0.2.1 R0 record (r4): the accepted G015 code at `f6d79af` (code identical to merge `7e0ac44`)

Read-only: `git fetch`, `git log`/`git diff --stat 7e0ac44 origin/staging` (management/docs only after the merge),
`sha256sum`, `grep`/`sed` over the merged tree. No build, test, install, Docker, provider call or probe.

| Finding | file:line at `f6d79af` | Material? |
|---|---|---|
| P-G015-1 host lookups precede `newServer`; registry sites are the ones G016 replaces | `core/internal/mcp/handler.go:143` (`serveMetadata`) and `:223` (`serveConnection`) `h.cfg.Registry.Lookup(r.Host)`; `connections.go:181` and `:407` `LookupOrigin`; `newServer` declared `server.go:63`, called `handler.go:336` | no (line drift from r2's 138/218; same code) |
| P-G015-2 fixture helpers | `env_integration_test.go:136` `func setupMCP(t)`; `:223` `func (e *mcpEnv) handler(t, hosts map[string]tenancy.OrganizationID, v Verifier) *Handler` | no |
| P-G015-3 driver pins | `infra/validation/run.sh:3148` `mcp_it()`: Transport 7, Lifecycle 6, Audit 1, Recovery 1, Discovery 5; Ordering **10**, Footprint 1; `E-MCP(-3|-4)?` lines **5**, `E-MCP-TRIGGER` 0 | no |
| P-G015-4 capability set and eligibility; Duo empty-Organization pins | `access/policy.go:12-13` capability constants, `:89` `capabilityKind`, `:309` `eligible`, `:370` `factQueryFor`; `access/mcp.go:247` `ToolRequirement`; `matrix_test.go:1104` `TestToolEligibility` with the literal empty-Organization pins at `:1197-1228` | no — `organization.manage_domains` joins these exactly as §0.2 says |
| P-G015-5 no G015 migration or audit action | `core/migrations/` ends at `00010_mcp_connections.sql` (**00011 free**); `audit_test.go:75` pins 25 trigger + 6 core = **31** actions; G016's delta 31→35 stands | no |
| P-G015-6 contract/doc wording | `contracts/mcp/control-plane.md` (1 G015 reference), `docs/foundation/MCP.md` (4), `AUTHORIZATION.md` (1); `connection-context.md` gained only the ADR-007 alias-lifecycle note (`190be188…`) | no |
| P-G015-7 per-Handler seam | `core/internal/mcp/handler.go:81` `fixture atomic.Pointer[catalog]`; no package-global state | no |
| §0.3 G014 interfaces | `config/mcp.go:19,55` `MCP_TENANT_HOSTS_FILE`; `identity/mcp.go:90` `Resource` (exact stored resource); `web/lib/session.ts:97` `ownOrigin`; `web/lib/mcp-client.ts:128,190` `url !== \`${origin}/mcp/c/${id}\``; `00006_row_level_security.sql:469-472` `login_transactions` policies on `binding_hash = app.ctx_login_binding()`; `tenancy/slug.go:24` `ReservedSlugs()` | no |
| §12 assumptions | `TestMCPCatalog` (`migrate/mcp_integration_test.go:229`) pins the version-10 delta over version 9 (its own version); no `organizations.slug` column in any migration (Group slugs only); G015 added no migration/audit action; the Duo-corrected seam is merged as read | no |
| Harness/control scripts | `required-checks.tsv` `2d136fea…` and `validate-controls.sh` `04319fa9…`/`lib.sh` `03d30075…` are byte-equal to the rows G015's R0/dispatch pinned (117/33/14 rows; profiles 35/75/110, VA7 80); `run.sh` `e62f7612…` differs from G015's dispatch row only by the merged G015 B-MCP-IT text (+23 lines) | no |
| Spec digest changed (`0ec7a8e2…` → `7f84ae54…`) | appended "Owner-approved design selection — 2026-10-16" only (12 lines; no requirement or AC change) | no (rebind only) |

### 0.3 Accepted G014 interfaces relied on (unchanged from r1; confirmed at R0, §0.2.1)

`mcp.Registry`/`config.MCP{Hosts}`/`MCP_TENANT_HOSTS_FILE`; `identity.MCPConnection{Host, Resource}` and the exact
stored host/resource agreement; `resource_mismatch`; `identity.Origins`/`CallbackURL`/`SignInURL`/`ReturnPath`;
`loginTxn` and `app.login_transactions` (admitted under RLS **only by `binding_hash = app.ctx_login_binding()`**,
`00006:469-472`); web `ownOrigin` (Host only) and the `__Host-sid/login/logout` cookies; `web/lib/mcp-client.ts`
(`connection()` requires `url === origin + "/mcp/c/" + id`); the auth-reader pattern (`USING (true)` policies and
SECURITY DEFINER readers owned by `<db>_auth_reader`); the 00010 guard/RLS/audit template; `tenancy.ReservedSlugs()`;
the identity overlay.

### 0.4 Owner decisions and Codex implementation constraints recorded at R0 (r4)

Source: [ADR-007](../../../adr/ADR-007-verified-domain-routing-and-authentication.md) (ACCEPTED, owner
"Approve those details"), the Spec's owner-selection section and the r3 review's acceptance section at `f6d79af`.
Design only; **no live DNS/TLS proof, WorkOS setting, credential/account change, provisioning or deployment is
authorized**, and P-1..P-4 stay explicit final-qualification prerequisites.

| ID | Decision | Effect on this Plan |
|---|---|---|
| D-1 | **E1 Caddy on-demand automatic TLS** with core-owned `tls-ask` and the `edge-probe` readiness check; production qualification is an **explicit AC6 gate** (P-4). No Cloudflare adapter, account or ADR for E2; no per-tenant Terraform or multi-provider controller | §1.5 E1 branch, §1.3 issuance predicate, §6.4 AC6 row; the E2-only parts of `RetryWithdrawal` are inert under E1 (always `domain_state_invalid` after the removal commit, §1.5); the fixture adapter still scripts the E2-shaped cases for the contract |
| D-2 | **(a) membership required** at tenant-host sign-in; an unmapped principal is refused before any user/mapping/session is created | §1.7 membership rule, §1.7.1 step 2, `organization_not_member`; overlay seeds memberships (`identitysetup`) |
| D-3 | **(a) provider wildcard for platform hosts + B2 shared callback relay** for custom hosts; no relay session/cookie, shared-domain cookie, new issuer or reusable bearer in redirects | §1.7 legs 1–3, §1.7.1, §1.7.2; B1 remains the recorded fallback only if the security reviewer finds a defect |
| D-4 | **(a) retain** historical MCP records; refuse while the exact stored host does not currently route to the bound Organization; listable/revocable through a current authorized host; **no automatic revocation**, no audience/resource rewrite | §1.6; C-4 resolved; `connection-context.md` dated rows and the MCP.md note follow this |
| D-5 / D-6 | settled defaults confirmed (internal-only management; reviewed provenance) | unchanged |
| 6 (Codex) | the production/overlay edge and every log must **omit or redact the entire callback query** (including `error_description`), not merely `code`/`state`; no relay cookie/session | tightens §1.7.2's redaction wording: the overlay keeps `access_log off`; DOMAINS.md records the whole-query rule for the Caddy `log` filter; the evidence-hygiene grep covers the whole query (`?& (not included)=`) |
| 7 (Codex) | `ResolveWith` uses the caller's query interface; **identity receives the resolver through an interface it declares itself** (no identity→domains import); standalone pool/autocommit lookups follow the restricted definer-reader pattern; the inventory keeps SQL, transactions, pool acquisitions and HTTP calls distinct | §5.0 (`identity.Config.Resolve` interface value) and §5.0.1 boundary cases already say this; §5.5 unchanged |
| 8 (Codex) | the final `signIn` guard's READ COMMITTED semantics add **no instant revocation or stronger serialization**; later changes apply through subsequent requests/operations; existing user-disablement locks remain; WorkOS calls hold no DB locks | §1.7.1 as written; the Result must state the request-bound authority truthfully |

## 1. Approach

G016 adds one durable fact — which canonical Organization a verified hostname routes to — one pure normalization
contract, one bounded lifecycle for customer hostnames, and composes the existing identity and MCP paths on top of
it without changing their authority model. Names stay routing aliases; UUIDs stay identity; every request still
authorizes on current facts.

### 1.1 Hostname normalization and classification contract (R1, R2; AC1, AC2) — unchanged from r1

Pure package `core/internal/hostname`: single `Host` header value; ASCII lowercase; one trailing dot stripped;
non-ASCII and `xn--` rejected (no IDNA in v0); labels 1–63, total ≤ 253; IP literals, userinfo, path, scheme,
whitespace rejected; the normalized key is portless; a port only when equal to `PLATFORM_HOST_PORT` (empty in
production; `543` in the overlay). Classification: `B`/`www.B` → public entry (303 to `MARKETING_SITE_URL`); `T`, the
shared auth host and `<reserved>.B` → reserved (404); `<slug>.B` → platform slug lookup; anything else → custom lookup;
no default tenant. Reserved floor = Group reserved set ∪ {`www`, `auth`, `custom`, `app`, `mail`, `smtp`, `ns`,
`ns1`, `ns2`, `cdn`, `static`, `status`} plus configured labels. A custom claim equal to `B`/`T` or under `.B` is
`host_reserved`.

### 1.2 Trusted edge boundary (R1; AC1) — unchanged from r1

Web and core read only `Host`; forwarding headers are never read; the edge terminates TLS and passes `Host` as sent;
core is reachable only from the edge/web network.

### 1.3 Resolution path, routing and issuance predicates (R1, R4, R7; AC1, AC6, AC11)

One SECURITY DEFINER reader `app.hostname_state(host text) RETURNS TABLE (organization_id uuid, kind text, routing
boolean, issuance boolean)` owned by `<db>_auth_reader` (E1/E2 pattern), one unique-index probe, no per-request
network work, no cache (invalidation = commit):

- platform: `organizations.slug = <label>` → `routing = true, issuance = false` (platform TLS is the shared wildcard
  certificate, outside the tenant lifecycle);
- custom: the single live row (`state <> 'removed'`) for `hostname` with
  **routing** ⇔ `state = 'active' ∧ verified_until > now() ∧ tls_state = 'ready' ∧ tls_not_after > now()` and
  **issuance** ⇔ `state ∈ {'verified','active','suspended'} ∧ verified_until > now()` (fresh ownership proof is
  required for issuance; `suspended` by a TLS-only loss keeps issuance so renewal can recover **until** a recheck
  confirms DNS proof loss; every confirmed proof loss — from `verified`, `active` **or an already `suspended`** row —
  sets `verified_until = now()` in the same update, so issuance stops at that commit and never waits for the previous
  freshness deadline, §1.4; an unconfirmed (transient) failure changes neither state nor deadline);
- no row, removed or non-matching → no tenant (`unknown_host`), both false.

The resolver exposes `Resolve(ctx, host) (Resolution, error)` (routing callers: identity login/callback/session, the
MCP handler, web's host gate) and `IssuanceAllowed(ctx, host) (bool, error)` (the E1 `tls-ask` route). A store failure
is an error (503/`unavailable`), never a clean unknown host.

### 1.4 Domain records and lifecycle (R3, R4; AC3–AC6)

`app.organization_domains` (§4): canonical UUID, immutable `organization_id`/`hostname`, finite `state`
{`pending`, `verified`, `active`, `suspended`, `removed`}, `generation` (claim generation of this hostname, §1.5),
challenge, freshness and TLS columns, `version`. Writes run in the facade's guarded TENANT-ORG transaction with lock
order user → session → Organization (FOR SHARE via `ReadFacts`) → domain row (`FOR NO KEY UPDATE`); uniqueness is
the partial unique index; every update carries `version + 1` (optimistic guard; a raced change is
`domain_state_invalid`, zero effects).

**Challenge rule.** A claim creates a 32-byte random challenge value (base32, 52 characters; a public DNS record
value, not a credential; never logged). `challenge_expires_at` (default 7 d) bounds **initial verification only**:
a `pending` claim whose challenge has expired cannot verify; `RotateChallenge` (pending only) issues a new value and
expiry. Once verified, the stored value is the **ongoing proof**: every recheck compares the live TXT with the same
stored value, and the customer keeps the record. Freshness (`verified_until`, default 30 d) renews on each successful
recheck independently of the challenge expiry, which is irrelevant after `pending`. The value is never rotated after
verification (a rotation would need a new claim generation).

**Transitions and the exact allowed updates (mirrored by the guard, §4.3).**

| From → event | To | Columns that may change (plus `version + 1`) |
|---|---|---|
| `pending` → verify ok | `verified` | `state`, `verified_at`, `verified_until`, `last_check_*`, `check_failures = 0` |
| `pending` → verify mismatch/nxdomain/transient | `pending` | `last_check_*`, `check_failures` (no suspension before first verification) |
| `pending` → rotate | `pending` | `challenge_value`, `challenge_expires_at` |
| `verified` → TLS status ready | `active` | `state`, `tls_state = 'ready'`, `tls_checked_at`, `tls_not_after`, `tls_reference` |
| `verified`/`active`/`suspended` → recheck ok | `verified` or `active` (active iff TLS ready and not expired) | `state`, `verified_until` (forward only), `last_check_*`, `check_failures = 0`, `tls_*` |
| `verified`/`active` → recheck confirmed proof loss | `suspended` | `state`, `verified_until = now()` (issuance stops), `last_check_*`, `check_failures` |
| `suspended` → recheck confirmed proof loss (**r3, review item 1**: a row suspended for TLS loses its DNS proof) | `suspended` | `verified_until = now()` (issuance stops at this commit, not at the previous deadline), `last_check_*`, `check_failures`; `state` unchanged (same-state refresh) |
| `active` → TLS error/expired/withdrawn | `suspended` | `state`, `tls_*`, `last_check_*` (proof untouched, issuance continues while fresh) |
| any non-terminal → transient failure | same state | `last_check_at`, `last_check_outcome = 'transient'`, `check_failures` |
| any non-terminal → remove | `removed` | `state`, `removed_at`, `tls_state = 'withdrawing'` |
| `removed` → cleanup outcome recorded (post-commit `Remove` or an explicit `RetryWithdrawal`, §1.5) | `removed` | **only** `tls_state` (`withdrawn`/`error`/`unknown`), `tls_checked_at`, `last_check_at`, `last_check_outcome`, `check_failures` (incremented on every non-`withdrawn` outcome; the retry bound, §5.4) |

Everything else raises `organization_domains_guard`. Removal is non-routing at commit; the terminal cleanup outcome
lives on the removed row's TLS columns and in the trigger-owned `domain.checked` event. Reassignment = a new row with
`generation + 1` for the same hostname (allowed by the partial unique index once the old row is `removed`); the old
row keeps its generation forever. The **ordinary recheck excludes removed rows** (`domain_state_invalid`); the only
operation that touches a removed row after its removal commit is the explicit withdrawal retry of §1.5.

**Guard/state case for the already-suspended loss (review item 1).** `suspended` has two causes that the row
distinguishes by its columns, not by a sub-state: TLS-only (`verified_until > now()`, `tls_state ∉ {'ready'}`) and
proof loss (`verified_until ≤ now()`). A recheck on a `suspended` row therefore ends in exactly one of: ok (→
`verified`/`active`, `verified_until` moves forward), confirmed proof loss (stays `suspended`, `verified_until = now()`
whether or not it was already in the past — the guard allows a decrease only to `now()`), transient (deadline and
state untouched). `hostname_state` needs no new column: issuance is `verified_until > now()`.

### 1.5 DNS checker and generation-bound TLS adapter (R3, R4; AC4, AC6)

- `domains.Checker` — unchanged from r1 (system resolver, 5 s deadline, ≤ 3 CNAME hops, ≤ 16 TXT answers, refusal
  list before any lookup, fixture only with `APP_ENV=test`).
- `domains.TLSAdapter` operates on `DomainRef{ID uuid, Generation int64, Hostname, Reference string}` and returns
  `Result{Ref DomainRef, State, NotAfter, Reference, Outcome}`:
  `Request(ctx, ref)`, `Status(ctx, ref)`, `Remove(ctx, ref)`, each with `DOMAIN_TLS_TIMEOUT` (default 5 s) and an
  explicit `unknown` outcome on timeout/ambiguity. **Results are applied only to the row with `ref.ID` and
  `ref.Generation` under the `version` guard**; a result for another id/generation (a late completion after
  removal and re-claim) is dropped and logged `stale_result`. No name-only operation exists.
  - E1 (Caddy on-demand): `Request` is a no-op (`requested`); `Status` = the `edge-probe` (dial `DOMAIN_TLS_EDGE_ADDR`
    with SNI = hostname; verify chain, SAN, expiry) → `ready`/`pending`/`error`/`expired`/`unknown`; `Remove` =
    `withdrawn` immediately (withdrawal is the issuance predicate answering 403 from the removal commit on; Caddy keeps
    an already issued certificate until its expiry, and routing already refuses). A replacement claim's issuance waits
    for its own fresh proof (`pending` ⇒ issuance false).
  - E2 (Cloudflare for SaaS): `Request` creates the custom hostname and stores its resource id in `tls_reference`;
    `Status` reads that id; `Remove` deletes **that id** only (a replacement claim creates a new resource with a new
    id); a deletion that returns not-found is `withdrawn`; a `Remove` timeout or ambiguous reply is `unknown`.
  - Fixture adapter (test only): scripted outcomes per `DomainRef`, including a late result for a removed generation
    and a scripted `unknown` → `withdrawn` sequence for the retry.

**Explicit withdrawal retry (r3, review item 2).** Because the ordinary recheck excludes removed rows, an `unknown`
or `error` cleanup outcome is retried only by `RetryWithdrawal(ctx, org, id)` — an explicit, operator- or owner-invoked
guarded operation (`organization.manage_domains`; route §5.3), never a background reconciler, timer or generic retry
framework:

- precondition (read inside the guarded write transaction, row `FOR NO KEY UPDATE`): `state = 'removed' ∧ tls_state ∈
  {'withdrawing','unknown','error'} ∧ check_failures < DOMAIN_WITHDRAWAL_RETRY_MAX`; otherwise `domain_state_invalid`
  (already `withdrawn`, not removed) or `domain_retry_exhausted` (bound reached), zero effects;
- the adapter call uses the removed row's **own** `DomainRef{ID, Generation, Hostname, Reference}` — the stored
  `tls_reference` of that generation under E2, the no-op `withdrawn` under E1 (so under E1 the retry is always
  `domain_state_invalid`: withdrawal completed at the removal commit) — and runs **after** the precondition transaction
  commits, outside every lock, like the post-commit `Remove`;
- the result is applied by a second guarded write under the `version` guard to the same id/generation and may change
  **only** the removed-row cleanup columns of §1.4 (`tls_state`, `tls_checked_at`, `last_check_at`,
  `last_check_outcome`, `check_failures`); `state`, `removed_at`, `generation` and `hostname` are immutable; routing
  and issuance stay false for that row whatever the outcome; a replacement claim of the hostname (new generation) is
  unaffected by, and never waits for, the old generation's cleanup;
- after `DOMAIN_WITHDRAWAL_RETRY_MAX` non-`withdrawn` outcomes the row stays `removed` with its last outcome, the
  `domain.checked` event records it, and the remaining provider cleanup is a documented manual step in DOMAINS.md (E2
  console deletion of the recorded resource id); nothing in the platform retries further.

### 1.6 MCP aliases and historical setup listing (R6; AC9)

Production host lookup is the durable resolver through `mcp.HostLookup`; `MCP_TENANT_HOSTS_FILE` set in production is
`config_invalid`. Exact agreement rule unchanged: a connection stores the exact `host[:port]`/`resource` it was created
on; a request is admitted only when the request host resolves to the bound Organization **and** equals the stored
host. Retention while resolving; refusal (`resource_mismatch`) at any other alias; removal/slug change → 404 before the
store; reassignment → 401; no rewrite, no automatic revocation (D-4 alternatives in §1.9).

**Historical aliases in setup listing (review item 5).** The core list is already Organization-scoped (RLS:
`user_id = ctx_user AND organization_id = ctx_org`), so rows created on an old host are returned; web's
`connection()` parser rejects them (`url !== origin + "/mcp/c/" + id`), making a mixed page unavailable. Minimum
change: the parser accepts any `RESOURCE`-shaped URL whose path id equals `connection_id` and exposes `host` for
display; the page shows the host next to each URL; revoke stays by `connection_id` through the current host (core
`Revoke` resolves the current origin → Organization and revokes the caller's own row in that Organization whatever its
stored host — already its behaviour). **Create** keeps the exact-origin check (`url === origin + …`) and MCP execution
keeps stored host/audience agreement. No old connection is retargeted. The setup-listing contract (mixed hosts in one
page, revoke by id through any current host of the Organization) is recorded in `docs/foundation/MCP.md`, where the
internal setup contract lives; `control-plane.md` is unchanged.

### 1.7 Browser authentication composition (R5; AC7, AC8)

Unchanged invariants: one WorkOS application per environment; the `(issuer, client_id)` namespace binding; host-only
`__Host-` cookies; browser-bound one-use login transactions (RLS by binding hash); PKCE S256; identity conflict,
`already_signed_in`, assurance `unknown` and disabled-user rules of G004. Core validates every login/callback/session
origin by **resolving its host** (§1.3); `IDENTITY_PUBLIC_ORIGINS` stays the exact allow-list of non-tenant origins
(the shared auth host and nothing else). **Membership rule (D-2, recommended):** a tenant-host sign-in completes only
for a current member of the resolved Organization recorded on the transaction at login; otherwise no session is
issued and `auth.sign_in` records `denied · organization_not_member`.

**Platform subdomains (D-3a):** provider wildcard redirect/sign-out entries (`https://*.B/auth/callback`,
`https://*.B/sign-in`, concrete defaults) and the unchanged G004 per-host flow; application-side exact validation by
resolution.

**Custom domains — B2 = shared callback relay (D-3b, recommended).** The shared host `A = PLATFORM_AUTH_HOST`
(reserved label, in `IDENTITY_PUBLIC_ORIGINS`) is a **relay with no cookie, no session, no transaction read**:

| Leg | Where | What happens | Durable/security facts |
|---|---|---|---|
| 1 login | custom host `T` | `POST /auth/login` on T; web sets `__Host-login` (binding `b`) on T; core `Login{Origin: https://T, Binding: b}` resolves T → Organization `O` (kind custom); creates the login transaction with `origin = https://T`, `binding_hash = H(b)`, `organization_id = O`, `callback_origin = https://A`, PKCE verifier, `state = <22-char random>.<origin tag>` where the origin tag is base64url(`https://T`) — opaque to the provider, carries **no secret**; the stored `state_hash` covers the whole state; `redirect_uri = https://A/auth/callback` | row admitted by the binding hash as today (no RLS change); the Organization is pinned at login |
| 2 relay | shared host `A` | provider returns `GET https://A/auth/callback?<raw query>` (`code&state`, or `error[&error_description]&state` on a provider refusal); web(A) reads **only** the `state` value's origin tag (exactly one `state`, ≤ 512 characters, decodable tag; the whole query ≤ the existing 8 KiB bound), asks core `GET /internal/v1/hosts/resolve?origin=` (no credential), and **only if** the origin is a routing-ready custom host whose configured callback origin is A answers `303 https://T/auth/callback` **with the raw query appended byte-identical** (no re-serialization: parameter multiplicity, `error`, `error_description` and `code` reach T exactly as the provider sent them); anything else is the fixed 404 page; nothing is stored, logged (beyond `auth.relay` outcome and a closed reason) or set; response headers per §1.7.2 | the authorization code is already in the provider's redirect URL by protocol; it is one-use, bound to the PKCE verifier held server-side and useless without T's binding cookie; no provider bearer, refresh token or application credential appears; the redirect destination is derived from resolution, never from a free-form parameter (open-redirect safe) |
| 3 consume | custom host `T` | `GET /auth/callback?<raw query>` on T: web reads `__Host-login` (b) and calls core `Callback{Origin: https://T, Binding: b, Query}` — **G004's unchanged one-use consumption** (`consumeTransaction(state_hash, H(b), client, origin)` under the binding-hash policy) with two added pre-exchange checks inside that same transaction: the transaction's `callback_origin` equals the configured A, and T currently resolves to the transaction's `organization_id` (removed/reassigned domain → `denied · host_changed`, transaction consumed, no session, **no exchange**); then G004's unchanged provider `error` handling, code exchange (PKCE), identity mapping, namespace/conflict/disabled rules; then the **final issuance guard of §1.7.1** (T still resolves to O; current membership of O under D-2) inside the issuance transaction; session issuance on T (`__Host-sid`); audit `auth.sign_in` | final browser proof = the binding cookie on the initiating host, exactly as G004; replay of the relayed URL fails (transaction consumed); a foreign browser fails the binding; an altered `state` fails the hash; an altered return cannot exist (the return path is stored); a removal or reassignment **during** the exchange is caught by §1.7.1 |

Sign-out on custom hosts: local and "everywhere" unchanged; provider sign-out ends on A's registered `/sign-in`
page (a fixed "signed out of WorkOS; return to your workspace" page with no tenant data); the truthful recorded outcome
is shown on T when the user returns within 10 min (`__Host-logout`), recorded in IDENTITY.md as a limitation.
Logout-wildcard behaviour is not assumed proven by redirect documentation (P-1 verifies both).

**B1 (explicit alternative):** no relay; each custom domain needs an operator-added exact provider callback and
sign-out entry; AC7/AC8 record "provider configuration required per custom domain". Everything else identical,
including §1.7.1 and §1.7.2 (they do not depend on the relay).

#### 1.7.1 Final session-issuance guard and its ordering (r3, review item 5)

The existing `Callback` runs three phases on separate connections: the consume transaction (LOGIN shape) commits,
then the provider exchange runs over the network with **no database connection or lock held**, then `signIn` opens the
issuance transaction (IDENTITY shape; `inTxAudited`) that maps the principal, handles continuity and calls `issue`
(user `FOR SHARE`, session INSERT). A domain removal/reassignment or a membership revocation that commits during the
exchange is invisible to the pre-exchange checks of leg 3. G016 therefore adds the **final guard inside the issuance
transaction**, placed after the principal mapping decision and **before the first write of either branch** (before the
supersession `UPDATE` of the continuity branch, before `createMapping`, before `issue`):

1. `ResolveWith(ctx, tx, T)` (§5.5; one statement through the same `tx`): the resolution must be `routing = true` with
   `organization_id = txn.organization_id`; otherwise `ErrHostChanged` → rollback-only (no row written, no session),
   `auth.sign_in denied · host_changed` through the existing post-consume `signInFailure` path.
2. D-2(a): `SELECT app.is_organization_member($user, $org)` (the auth-reader definer function of §5.5, one statement,
   current `role_bindings` row existence); otherwise `ErrNotMember` → rollback-only, `denied · organization_not_member`.
   An **unmapped principal** on a tenant host cannot be a member, so under D-2(a) it is refused before `createMapping`:
   no user, mapping or session is created (consequence recorded in the D-2 row of §1.9 for the owner). Under D-2(b)
   step 2 is absent and only step 1 runs.
3. Platform-slug hosts run step 1 (slug still maps to O) and, under D-2(a), step 2; the shared auth host and the
   `IDENTITY_PUBLIC_ORIGINS` entries never issue a session (unchanged).

**Declared ordering and locks.** Both guard reads are plain snapshot reads of SECURITY DEFINER readers at READ
COMMITTED: the issuance transaction takes **no lock** on `organization_domains`, `organizations` or `role_bindings`
(adding one would create new lock edges between the IDENTITY writer and the TENANT-ORG domain/binding writers; the
existing lock order user → session → Organization → domain row is unchanged and the issuance path stays outside it).
Consequently: a removal/revocation that commits **before** a guard statement starts is seen and refused; one that
commits **after** the guard but before the session INSERT commits leaves a session on a host that no longer routes or
for a user who is no longer a member — that session is refused at its **next request** by the per-request resolution
(routing false → `unknown_host`) and by the existing per-request authorization (RLS/`role_bindings`), exactly as a
post-sign-in revocation is handled today; G016 adds no session-revocation write for it (consistent with D-4(a)). The
window is one request. No network call occurs inside any transaction: the exchange result (`vi`) is already in hand
when the issuance transaction begins; the guard adds two statements and no round trip to the provider.

**Evidence.** `TestCustomDomainRelay` (§6.2) gains the race cases by making the **fixture provider's `Exchange`**
perform the removal (case A) or the membership revocation (case B) before returning, so the change commits between
the exchange and `signIn` with no new test seam: expected `host_changed` / `organization_not_member`, no session row,
one `auth.sign_in denied` event, the login transaction consumed. Case C pins the bound: sign in, remove the domain,
and the next `Session(?origin=T)` resolution is `unknown_host`. The continuity branch (live session for the same user
on a reassigned host) is case D: the old session stays untouched, no supersession.

#### 1.7.2 Relay and callback privacy, refusals and evidence hygiene (r3, review item 6)

- **Headers.** The web form layer already answers every `seeOther` (303) and `statusPage` (400/404/405) with
  `Cache-Control: no-store` and `X-Content-Type-Options: nosniff` (`web/lib/session.ts` `NO_STORE`). G016 adds
  `Referrer-Policy: no-referrer` to that same constant, so the relay 303 on A, the fixed 404 on A, the callback 303 on
  T and the ambiguous 400 on T all carry no-store **and** no-referrer with one change; `web/tests/unit/session.test.mjs`
  asserts both headers on `seeOther` and `statusPage`. Why: the only documents rendered at a `code`/`state`-bearing URL
  are the fixed 404 and the ambiguous 400, whose same-origin links would otherwise send the full URL as `Referer`; on
  the 303s the header also governs the redirected navigation. Core's identity JSON is already no-store with no CORS
  headers; `hosts/resolve` and the other `domains/api` routes answer through the same no-store writer. No new cookie
  is set on A; the relay never sets or reads a cookie.
- **Edge query redaction.** The identity overlay edge (`infra/identity/proxy.conf`) has `access_log off` on both
  servers; it stays off, with a one-line comment recording the reason (callback URLs carry `code`/`state`). The
  production edge is a **deployment input** under D-1/P-2: its access log must either omit `/auth/callback` requests or
  redact the `code` and `state` query values (Caddy: a `log` `filter` encoder replacing those `request>uri` query
  fields; recorded in DOMAINS.md with the `ask` endpoint input). No deployment authority is claimed; the operator
  confirms the edge log configuration as part of P-2 before manual qualification.
- **Logs.** web's request context logs no URL (verified: `request-context.ts` records request id/trace only); core
  logs `auth.callback.completed {outcome}` and `auth.relay {outcome, reason}` with closed values; hostnames appear only
  as the 12-hex-character hash prefix of §5.5; `state`, its origin tag, `code`, challenge values and cookies are never
  logged. The E2E log assertions (`events.json`) pin `auth.relay`'s closed field set.
- **Provider error and refusal behaviour (G004 preserved).** The relay keys on `state` only and forwards the raw query
  byte-identical, so a provider refusal (`error=access_denied`, no `code`) or any other provider error reaches T where
  the **unchanged** post-consume G004 handling records `denied · login_canceled` or `failed · provider_error`, and
  repeated `code`/`state`/`error`/`error_description` keys still yield `ambiguous_parameter` (400) on T. On A itself:
  no `state`, a repeated `state`, an oversized query, an undecodable tag, or a tag that is not a routing-ready custom
  host with callback origin A → the fixed 404 (nothing consumed, nothing reflected, `auth.relay outcome=refused
  reason=<closed>`); the relay never answers 400, since it consumes nothing.
- **Evidence hygiene.** Recorded evidence (Result excerpts, harness logs, Playwright artefacts attached to the Result)
  must contain no raw `code=`/`state=` values or cookie values: before attaching, the implementer runs
  `grep -rE '?& (not included)=[^&[:space:]]{8,}' <evidence dir>` and records the empty result; the fixture provider's
  emulator codes are fixture values but are treated the same way.

### 1.8 Options compared — unchanged from r1 except as noted

E1/E2/E3 (edge/TLS), B1/B2/B3 (callback; B2 now the relay of §1.7), M1/M2 (management, **M2 settled**). Trade-off
tables as r1 §1.8 with the B2 row replaced by §1.7's relay (security delta vs G004: one extra same-TLS browser
redirect carrying the provider's own code/state; two added consumption checks; no cookie or session on A).

### 1.9 Owner decision brief (D-1..D-4; before Plan ACCEPT; nothing selected here)

| ID | Choice | Executable alternatives | Recommendation and consequence |
|---|---|---|---|
| **D-1 edge/TLS mechanism and live authority** | (a) **E1** Caddy on-demand TLS + core `tls-ask` (self-hosted ACME; `edge-probe` status); (b) **E2** Cloudflare for SaaS custom hostnames (vendor API adapter, account/credential, sub-processor review, ADR); (c) **E3** interfaces + `edge-probe` + fixture only, production adapter evidence as an explicit held AC6 gate. Independently: authorize (y/n) one bounded live DNS/TLS proof at an owner-controlled domain (P-3/P-4) | **E1 + E3's explicit gate**: no vendor, adapter reusable; production provisioning stays a separate deployment authority. E2 if a managed edge is preferred (adds ADR, credential, review). Without live authority, AC6 closes on fixtures plus the recorded gate only |
| **D-2 membership-bound session issuance** | (a) **membership required** at tenant-host sign-in (no session otherwise; `organization_not_member`); (b) global session on any tenant host, tenant denial later | **(a)**: matches Spec R5/AC8 "no unauthorized session"; changes G004's issuance contract (security review); overlay users need seeded memberships; **consequence (r3, §1.7.1)**: a principal with no user yet cannot complete a first sign-in on a tenant host (nothing is created; the user signs in at the platform host first and is then granted membership). (b) keeps G004 bytes but leaves a session on a host whose tenant denies the user. Either way the final guard of §1.7.1 re-resolves the host at issuance |

Historical note (r3): the owner design package of §0.1 was the owner-facing form of D-1..D-4 and no answer was
assumed in r3. **r4: the owner approved the recommended package; the decisions are recorded in §0.4 (ADR-007).**
| **D-3 callback scaling** | (a) provider wildcard pattern for platform subdomains (required either way); (b) custom domains: **B1** per-domain operator entries vs **B2** shared callback relay (§1.7) | **(a) + B2**: no per-customer provider entry; one extra TLS redirect of the provider's own code/state; G004 consumption unchanged. B1 if the owner rejects any shared host: scales by operator action per domain, recorded truthfully |
| **D-4 connection retention on alias change** | (a) **retain while resolving, refuse elsewhere, no automatic revocation** (owner revokes via setup page through a current host); (b) automatic revocation of connections whose host was removed/renamed (extra guarded write path, lock order, audit) | **(a)**: no new write path; old rows stay listable/revocable (§1.6). (b) adds a mutation in the removal/slug transaction and a `session.revoked` reason; choose if an orphaned-but-unreachable connection is unacceptable |

Owner vendor/security choices precede Plan ACCEPT; no live probe is authorized by this brief.

## 2. Source binding and extra inputs

All rows are at the R0 target `origin/staging` `f6d79af5cc97e0cf6c17b5a9f649a2ad0d4cf827` unless stated (rebind,
r4). Hash rows carry neutral labels (the Result repeats the binding rows).

| Input | Git commit/path | Digest (SHA-256) |
|---|---|---|
| Goal (G) | `management/v0/goals/G016-hostname-resolution-domain-lifecycle.md` (unchanged since `b8ceeab`) | `299f08453a4e6defa6774c139052f36883be39f0c3bcff574d11c6ce6fbbfc56` |
| Work Spec (S) | `management/v0/specs/G016-hostname-resolution-domain-lifecycle-spec.md` (owner-selection section appended at `f6d79af`; previously `0ec7a8e2…`) | `7f84ae540adf0535dd16b2b2b5f458358612905f5d79ff1d5c24e8a6a4998308` |
| Plan r3 archive (P3), the accepted design bytes | `management/v0/plans/G016-hostname-resolution-domain-lifecycle-plan-r3.md` | `3b29275ea8c914178138889e29e3a7412209be1378ae1c1019904e08809c1ffc` |
| r2 archive / r1 archive | `…-plan-r2.md` / `…-plan-r1.md` | `3b6ea4895e6195065406c08eeb385aab9508e49fc2c589501ff9d2d277a1b702` / `38c736e711d4284718381f46aa5774564e32f0bf346060e531c3dcc12abe711c` |
| r1 review | `b8ceeab` `management/v0/results/G016-plan-r1-codex-review.md` | Git-bound |
| r2 review (R2) | `management/v0/results/G016-plan-r2-codex-review.md` | `32f18832470bf5d0b4301d5465aa4c75241facc204060d3b6043475ae37348e3` |
| r3 review with the owner approval and Plan ACCEPT (R3) | `management/v0/results/G016-plan-r3-codex-review.md` | `cd763cf7f20e7dca955b5c24d63589dc628652e9ca7adcc580b676ae991ff359` |
| Scope allow-list (SC), imported by Codex, byte-equal to §11.3 | `management/v0/results/G016-scope-allowlist.txt` (44 patterns) | `047450e1f12bab0ae2d7e07432b51109d27c43efa9935930f3b27b97fa983aa0` |
| ADR-007 (A7) | `docs/adr/ADR-007-verified-domain-routing-and-authentication.md` | `46cf243031b08e2b3f4dc11a43f51e12ee1f810eb721fb903f772733e34f80fe` |
| Architecture / Security / Performance (PG-007) notes (AR, SE, PF) | `docs/ARCHITECTURE.md` / `docs/SECURITY.md` / `docs/PERFORMANCE.md` | `c056ce5f5418bac1870dfbc05b521632d9b66c66ceb9906d3d7144f588f1d30f` / `38aa0d1df17678dcb2a241cd739693f41eb66d1cdf4d7b23239b434776ef6b37` / `e3b3db3aeca17c102cade4f987dedabd924048fe9acdeafe66c2628d54c16b08` |
| MCP contracts (CC, CP) | `contracts/mcp/connection-context.md` / `contracts/mcp/control-plane.md` | `190be18885c4cadd2e2ba60fe9b1c0eb2391c9384219a5ba57991ba47d0a607e` / `82ccd466a99d5c44f3434b9ca8512468a63f1b914ba65b31485e7286f5815a29` |
| Predecessor (X15, I15) | `management/v0/results/G015-dynamic-capability-exposure-result.md` / `management/v0/results/G015-implementation-codex-review.md`; code = merge `7e0ac44` (source `b80c283`), final ACCEPT `443c66a` | `69235db5f96db2287ad85868d7e89935ad79d82df761b3b33021a365b3eae986` / `935077028370f4fd8337492093581f51825ba4b545e58a1f9254c7623d4db496` |
| Harness (RS, RC, VI) | `infra/validation/run.sh` / `infra/validation/required-checks.tsv` / `skills/sdlc-management/references/validation-integration.md` | `e62f76127501fb6659cf4ae7cab14da80f968437f52c04e5776ea0834142ad44` / `2d136fea2fa096def63d29a5182fb563b6419f94babf519c8e61ed3455a1c686` / `696df897dcea6eb599a025adb8a22363d94845e4a411af1173292fa285b192d5` |
| Control scripts (VC, LB) | `infra/ci/validate-controls.sh` / `infra/validation/lib.sh` (byte-equal to the G015 dispatch rows) | `04319fa9368cd9523e09c3e745953bbf7afc7cf45b428dbd38292b0af5deb1eb` / `03d300759b62f53c32796a4345303c893d44946419d3761e7ff7fad3dc2532e4` |
| Governance | AGENTS.md, CLAUDE.md, DECISIONS.md, ADR-003/004/005/006, ROADMAP.md, PROGRESS.md (compacted at `443c66a`), G004 redirect-scaling addendum | Git-bound at `f6d79af` |
| Planner record | task G016, role plan/R0; r1/r2 from worktree `agent-<id>` (branch `g016-plan`); r3 from the successor worktree `agent-<id>` (context-threshold handover; `g016-plan` tip `9715fa7` on `36c7f4e`); r4 from the same successor worktree on branch **`g016-r0`** created from `f6d79af` (a new branch name, since the stale local `g016-plan` is held by the retired planner checkout) | — |

## 3. Files

**Create**

| Path | Content |
|---|---|
| `core/internal/hostname/{hostname.go,hostname_test.go}` | §1.1 |
| `core/internal/domains/{domain.go,lifecycle.go,dns.go,tls.go,resolver.go,errors.go}` | model, transition table, checker, generation-bound adapter (`edge-probe`, fixture), resolver (`Resolve`, `IssuanceAllowed`), closed errors. **Imports nothing from `access`** |
| `core/internal/domains/db/{sqlc.yaml,queries/domains.sql,querier.go,models.go,domains.sql.go}` | sqlc trial (§5.7) |
| `core/internal/domains/api/{http.go,http_test.go}` | internal routes (§5.3) incl. `tls-ask` and `hosts/resolve`; imports `access`, `identity`, `domains` |
| `core/internal/domains/{lifecycle_test.go,dns_test.go,tls_test.go}` | unit coverage |
| `core/internal/tenancy/access/domains.go`, `domains_integration_test.go` | guarded operations; `TestDomains` DOM-1..10 |
| `core/internal/migrate/domains_integration_test.go` | `TestDomainsCatalog` |
| `core/migrations/00011_organization_domains.sql` | §4 |
| `web/lib/host.ts`, `web/proxy.ts`, `web/tests/unit/host.test.mjs`, `web/tests/unit/mcp-client.test.mjs` | host gate; relay route logic lives in `web/app/auth/callback/route.ts` (modify); the parser test is new (no existing `mcp-client` unit test) |
| `infra/identity/{dns,tls}-fixture.tsv` | overlay fixtures |
| `docs/foundation/DOMAINS.md` | runbook (§3 r1) incl. the historical-alias listing/revoke steps and the Caddy `ask` deployment input |
| `management/v0/results/G016-scope-allowlist.txt`, `…-result.md` | binding artefacts |

**Modify** — as r1 §3, plus: `web/lib/mcp-client.ts` (parser accepts other-host URLs, exposes `host`),
`web/app/connections/mcp/page.tsx` (shows host), `web/tests/browser/mcp.spec.ts` (M-LIST mixed-host row),
`core/internal/identity/{service,store}.go` (`organization_id`, `callback_origin`, state origin tag, the two added
consumption checks, **the §1.7.1 final guard inside `signIn`**), `core/internal/identity/identity.go` (the closed
error set gains `host_changed` and `organization_not_member`), `web/app/auth/callback/route.ts` (relay branch on the auth host; raw query
forwarded byte-identical), `web/lib/session.ts` (**`Referrer-Policy: no-referrer` in `NO_STORE`**, §1.7.2),
`web/tests/unit/session.test.mjs` (header assertions), `infra/identity/proxy.conf` (comment only: `access_log off`
retained for the callback URLs), **`core/internal/tenancy/access/access.go` (exactly one field, §5.0.1)**,
`core/internal/tenancy/access/boundary_test.go` (`TestProductionImporters`: facade importers = `mcp`, `domains/api`;
`domains` never imports `access`/`identity`), `docs/foundation/MCP.md` (setup listing contract: mixed hosts).

## 4. Schema and migrations

**`core/migrations/00011_organization_domains.sql` (provisional number).** Additive; `down` refuses while any domain
row, non-NULL slug, handoff-shaped login transaction or domain audit event exists.

1. `app.organizations.slug` — as r1 §4.1 (nullable, grammar/reserved CHECKs, partial unique index, runtime
   `UPDATE (slug)`, `organizations_auth_reader` SELECT policy for `<db>_auth_reader`).
2. `app.organization_domains` — r1 §4.2 columns plus `generation bigint NOT NULL DEFAULT 1 CHECK (generation >= 1)`;
   `tls_expires_at` renamed `tls_not_after`; `tls_state` CHECK in {`none`, `requested`, `pending`, `ready`, `error`,
   `expired`, `withdrawing`, `withdrawn`, `unknown`}; shape CHECKs: `(state = 'removed') = (removed_at IS NOT NULL)`;
   `state IN ('verified','active','suspended') → verified_at IS NOT NULL AND verified_until IS NOT NULL`;
   `state = 'active' → tls_state = 'ready' AND tls_not_after IS NOT NULL`; `tls_state = 'ready' → tls_not_after IS
   NOT NULL`; `UNIQUE (hostname, generation)`; partial unique `organization_domains_hostname_live (hostname) WHERE
   state <> 'removed'`; index `(organization_id, created_at)`.
3. Guard `app.organization_domains_guard()` implements the §1.4 table literally: immutable
   `id, organization_id, hostname, generation, created_by, created_at`; `version = OLD.version + 1` on every update;
   the challenge columns change only when `OLD.state = NEW.state = 'pending'`; state edges exactly as listed;
   `verified_until` may decrease only to `now()` on a confirmed-loss edge, whose source state is `verified`, `active`
   **or `suspended`** (r3, item 1: `suspended → suspended` with `NEW.verified_until = now()` is a listed edge even
   when `OLD.verified_until` is already past); `removed` rows accept changes to `tls_state, tls_checked_at,
   last_check_at, last_check_outcome, check_failures, version` only (`check_failures` may only increase there; r3,
   item 2); anything else raises.
4. Reader `app.hostname_state(text)` (§1.3) and policy `organization_domains_auth_reader` (SELECT `USING (true)` for
   `<db>_auth_reader`); runtime policies `organization_id = ctx_org AND rls_org_role() IN ('owner','admin')` for
   SELECT/INSERT/UPDATE (INSERT also `state = 'pending' AND version = 0 AND generation >= 1`); runtime grants SELECT,
   INSERT (id, organization_id, hostname, generation, challenge_value, challenge_expires_at, created_by, created_at),
   UPDATE (state, challenge_value, challenge_expires_at, verified_at, verified_until, last_check_at,
   last_check_outcome, check_failures, tls_state, tls_checked_at, tls_not_after, tls_reference, removed_at, version);
   never DELETE.
5. Audit: T26 `domain.claimed` (INSERT), T27 `domain.state_changed` (UPDATE OF state; reason = new state), T28
   `domain.checked` (UPDATE OF last_check_at; success/failed; reason = outcome), T29 `organization.slug_changed`;
   resource type `organization_domain`; `access.denied` operations `domain.claim|rotate|verify|recheck|remove|
   retry_withdrawal`, `organization.set_slug`; closed facade errors add `domain_retry_exhausted` (r3, item 2);
   `auth.sign_in` reasons `organization_not_member`, `host_changed` (both `denied`, recorded by the post-consume
   `signInFailure` path, §1.7.1); catalog 31 → **35** (four actions; operations and reasons are not catalog actions).
6. **`app.login_transactions`** (B2/D-2 durable change; RLS policies unchanged): `ADD COLUMN organization_id uuid
   NULL` (the Organization pinned at a tenant-host login; NULL only for the non-tenant auth-host origin, which never
   signs in), `ADD COLUMN callback_origin text NULL CHECK (char_length BETWEEN 1 AND 2048)`; runtime INSERT widened to
   both columns; `state_hash` stays the primary key (the origin-tagged state is hashed whole). No new table, no
   definer writer, no cross-binding read.
7. Postconditions pin the version-11 inventories (relations, policies, functions, grants); locks and rollout as r1
   §4.7 plus a brief ACCESS EXCLUSIVE on `app.login_transactions` (two nullable columns, no rewrite).

## 5. Contracts and interfaces

### 5.0 Dependency and construction path (review item 6)

```text
hostname  ←  domains (model, lifecycle, dns, tls, resolver, db)  ←  tenancy/access (domains.go: guarded ops; imports domains for types/adapters)
                                                                  ←  mcp (HostLookup consumer; builds its own facade as today)
                                                                  ←  domains/api (HTTP; imports access + identity + domains; builds its own facade like mcp)
cmd/server/main.go: resolver := domains.NewResolver(pool); checker := domains.NewChecker(cfg); tls := domains.NewTLSAdapter(cfg)
                    identity service gets resolver (origin validation); mcp.New(... Hosts: resolver ...); domainsapi.New(domainsapi.Config{Pool, Identity, Resolver, Checker, TLS, ...})
```

No package imports a package to its right. `TestProductionImporters` pins `mcp` and `domains/api` as the only facade
importers; `TestPackageBoundary` (mcp) is unchanged. `mcp.HostLookup interface { Lookup(ctx context.Context, host
string) (organizationID string, err error) }` with `domains.ErrUnknownHost`; the MCP handler maps a non-unknown error
to 503 (`mcp.request` outcome `unavailable`) and never to 404. `identity.Config.Resolve` is the same interface value.

#### 5.0.1 Exact `access.go` wiring and boundary pin (r3, review item 3)

Inspected at the base: `access.Service` (`core/internal/tenancy/access/access.go`) holds `hooks *objectHooks` and
`mcpHooks *mcpHooks` as plain fields; `NewService`, `NewServiceWithObjects` and `newService` never set them (nil in
every production construction); tests assign `e.svc.mcpHooks = &mcpHooks{…}` in-package and reset to nil
(`mcp_integration_test.go`). G016 mirrors that exactly and nothing more:

| Change | Exact scope |
|---|---|
| `access.go` | **one field** added to `Service` after `mcpHooks`: `// G016: inert domain-lifecycle ordering hooks (nil in every production construction); domains.go.` + `domainHooks *domainHooks`. No constructor, parameter, import or other line changes; the zero value is the production value |
| `access/domains.go` (new) | `type domainHooks struct { claimLocked, checkLocked, removeLocked func() }` and the guarded operations (`ClaimDomain`, `RotateChallenge`, `BeginDomainCheck`, `RecordDomainCheck`, `RemoveDomain`, `RetryWithdrawal`, `ListDomains`, `SetSlug`) over `s.db`, `s.id` (`CheckLiveSession`) and the existing `ReadFacts`/decision/log helpers. The Service holds **no** resolver, checker or TLS adapter: network steps run in `domains/api` between two facade calls (`BeginDomainCheck` read transaction → DNS/TLS → `RecordDomainCheck` write transaction), so no network call can occur under a facade lock |
| `access/domains_integration_test.go` (new) | the only writer of `e.svc.domainHooks` (set/reset per case, like `mcpHooks`) |
| `boundary_test.go` | `TestProductionImporters`: the `access` case becomes `dir != "internal/mcp" && dir != "internal/domains/api"`; two cases added: `strings.HasPrefix(p, "core/internal/domains")` imported from `internal/identity` or `internal/tenancy` is an error (identity receives `Resolve` as an interface value it declares itself), and `p == "core/internal/domains/api"` imported from anywhere but `cmd/server` is an error |
| Out of scope | every other file of the `access` package except the already-listed `policy.go`/`errors.go`/`log.go` additions (capability, closed errors, logged operations) and the listed tests; no change to `objectHooks`, `mcpHooks`, `newService` or the `sessions`/`mcpIdentity` interfaces |

`mcp.Handler` keeps its per-Handler `fixture atomic.Pointer[catalog]` seam (P-G015-7) and gains only the `HostLookup`
configuration value; it is not copied.

### 5.1 Go surfaces — as r1, with `RotateChallenge(ctx, org, id)` and `RetryWithdrawal(ctx, org, id)` (§1.5) added,
`DomainRef`/`Result` per §1.5, `domains.ResolveWith(ctx, q Querier, host)` beside `Resolver.Resolve` (§5.5), the two
identity consumption checks (`callback_origin`, host-still-resolves) in `Service.Callback`, and the final issuance
guard inside `Store.signIn` (§1.7.1) with `ErrHostChanged`/`ErrNotMember`.

### 5.2 State machine — §1.4 table is normative (including the r3 `suspended → suspended` and removed-row rows).

### 5.3 HTTP (internal network only) — as r1 §5.3 plus `POST /internal/v1/domains/rotate {id}` and `POST
/internal/v1/domains/retry-withdrawal {id}` (§1.5; owner/admin of the Organization through the setup page or the
operator runbook; closed outcomes `withdrawn|error|unknown`, errors `domain_state_invalid|domain_retry_exhausted`);
`GET /internal/v1/hosts/resolve?origin=` returns `{kind, organization_id?, callback_origin?}` (the relay uses
`callback_origin`); the session route accepts `?origin=` and returns the resolution with the session so an
authenticated page needs one core round trip. All answers are `Cache-Control: no-store` JSON without CORS headers
(the identity `http.go` pattern).

### 5.4 Configuration — as r1 §5.4; `CUSTOM_DOMAIN_CHALLENGE_TTL` governs pending verification only;
`DOMAIN_WITHDRAWAL_RETRY_MAX` (default 5, range 1..20) bounds the explicit withdrawal retries per removed row (§1.5).

### 5.5 Audit, logs, privacy and PG-007 cost inventory (review item 8)

Log events: `domain.request` (closed outcome/reason/operation; hostnames appear only as a 12-hex-character hash
prefix; challenge values and certificates never), `auth.relay` (outcome only), adapter `stale_result`.

**Resolver execution model (r3, review item 4; replaces r2's "one query" framing).** The resolver owns **no
transaction**. Inspected precedent: the E2 reader `principalSubject(ctx, q, …)` takes a `QueryRow` interface and is
called both on the pool (`lookupUser`, one autocommit statement) and on a `pgx.Tx` inside `signIn`; the definer
readers "ignore ambient settings" (00006), so they need no `set_config`. G016 reuses exactly that shape:

- `domains.ResolveWith(ctx, q Querier, host) (Resolution, error)` — **one statement**, `SELECT organization_id, kind,
  routing, issuance FROM app.hostname_state($1)`, on whatever `q` the caller passes; `Querier` is `QueryRow(ctx, sql,
  args...) pgx.Row`, satisfied by `*pgxpool.Pool` and `pgx.Tx`. Kind is decided by the pure classifier before the call,
  so the function runs **one** index probe (custom: the partial unique index `organization_domains_hostname_live`;
  platform: the `organizations(slug)` partial unique index).
- `Resolver.Resolve(ctx, host)` = `ResolveWith(ctx, pool, host)`: one pool connection acquisition and **one autocommit
  round trip** — no BEGIN, no `set_config`, no COMMIT. It does **not** execute inside any session transaction, and the
  Plan no longer claims it does.
- Inside-transaction use is by call site, with `tx` passed explicitly: identity `insertTransaction` (Login, LOGIN
  shape), `consumeTransaction` (Callback pre-exchange check), `signIn` (final guard, §1.7.1), `resolve` (Session
  `?origin=`). Each adds **one statement to the existing transaction and zero transactions**.
- Pool use (autocommit): the MCP `serveConnection`/`serveMetadata` host decision (it precedes method, query, body and
  bearer handling and the metadata route has no transaction, so it cannot ride the later `BeginVerifiedRequest`
  transaction without reordering the 404 after the body read — rejected), `hosts/resolve` (web host gate and relay)
  and `tls-ask` (`IssuanceAllowed`). Each is **one autocommit statement on its own pooled connection**.
- `app.is_organization_member(user uuid, org uuid)`: a STABLE SECURITY DEFINER auth-reader function over
  `role_bindings` (same ownership/grant pattern as `hostname_state`), one statement, used only inside `signIn`.

**Per-operation inventory (statements counted inside core; round trips are web→core HTTP calls; "+N in tx" means N
statements added to an existing transaction, "1 ac" one autocommit statement on a pooled connection; measured values
replace these expectations in the Result):**

| Operation | web→core round trips | Core transactions and statements (expected, to be measured) | Locks |
|---|---|---|---|
| anonymous page/host gate | 1 (`hosts/resolve`) | **1 ac** (`hostname_state`); no transaction, no `set_config` | none |
| authenticated page | 1 (session route with `?origin=`) | existing `resolve` tx (BEGIN + initializer + E1 reader + … + COMMIT) **+1 in tx** | none |
| login (tenant host) | unchanged (1) | existing `insertTransaction` tx **+1 in tx** (before the INSERT; unknown/non-routing host → `origin_rejected`, nothing inserted) | none |
| callback (tenant host) | unchanged (1) | consume tx **+1 in tx** (pre-exchange host check); network exchange with no connection held; `signIn` tx **+2 in tx** (`hostname_state`, `is_organization_member`) before the first write; total **+3 statements, 0 new transactions, 0 new round trips** | none new (snapshot reads) |
| relay (auth host) | 1 (`hosts/resolve`) | **1 ac** | none |
| MCP request / metadata request | 0 (edge→core) | **1 ac** before the existing admission (replaces the in-memory registry lookup): one pool acquisition + one statement per request, **not** inside the admission transaction; recorded for owner attention | none |
| `tls-ask` (E1) | 0 (edge→core) | **1 ac** | none |
| claim | 1 | 1 write tx: BEGIN, `set_config`, liveness (`CheckLiveSession` 1), facts (`ReadFacts{Groups: []}`: organization FOR SHARE 1 + bindings 1), INSERT (unique-index conflict = `host_taken`), COMMIT → **≈ 5 statements + BEGIN/COMMIT**; denial adds 1 separate audit tx | organization FOR SHARE; new row |
| verify / recheck | 1 | read tx (ctx 1 + liveness 1 + facts 2 + row read 1 = 5) → network (DNS 2 lookups + chain ≤ 3; TLS 1 dial; no connection held) → write tx (ctx 1 + liveness 1 + facts 2 + row `FOR NO KEY UPDATE` 1 + UPDATE 1 = 6) → **≈ 11 statements across two transactions**; trigger-owned audit adds no statement | organization FOR SHARE twice (briefly); domain row |
| remove | 1 | write tx as the verify write (6) → post-commit adapter `Remove` → one cleanup write tx (ctx 1 + liveness 1 + row 1 + UPDATE 1 = 4) → **≈ 10** | as above |
| retry withdrawal (§1.5) | 1 | precondition write tx (ctx 1 + liveness 1 + facts 2 + row `FOR NO KEY UPDATE` 1 + `check_failures` UPDATE 1 = 6) → adapter `Remove` outside locks → cleanup write tx (4) → **≈ 10**; refused preconditions end at the first tx (≈ 5, zero effects) | domain row (removed) |
| set slug | 1 | write tx: ctx 1 + liveness 1 + facts 2 + `UPDATE organizations … WHERE id = $1` (row `FOR NO KEY UPDATE`, unique index) 1 = **≈ 5** | **organization row FOR NO KEY UPDATE conflicts with concurrent FOR SHARE fact reads of the same Organization** (brief; recorded concern) |
| list | 1 | read tx: ctx 1 + liveness 1 + facts 2 + index-ordered SELECT 1 = 5 | none |

**Bounded observations (existing mechanisms only; no new campaign).** The access integration environment already
counts statements through its hook (`e.hook.counts()`, used by `TestMCPFootprint`), prints `EXPLAIN (COSTS OFF)` as
E-MCP-1/2 and lists held locks as E-MCP-3. DOM-10 records, with the same helpers, the footprint lines `E-DOM-F1..F4`
(named apart from r1's browser cases E-DOM-1..5): **F1** EXPLAIN of the custom-host and slug probes (expected one
index scan each); **F2** statement count of `ResolveWith` on a tx (asserted 1) and of `Resolve` on the pool (asserted
1 statement, 0 BEGIN/COMMIT); **F3** per-operation statement counts for claim, verify write, remove, retry, set slug
and list, printed and asserted **constant** between 10 and 40 domain rows of the Organization (the E-MCP-TRIGGER
pattern; the fan-out guard); **F4** the relation locks held by the verify write and the issuance transaction (the
latter expected to hold none on domain/binding relations, §1.7.1). The Result records the measured values and the
PG-007 entry quotes them with no performance guarantee and no authority cache (C-5).

Query/index/lock concerns for owner attention before execution: (1) one autocommit statement per MCP request,
metadata request, anonymous page and relay, and +1..+3 statements inside the existing identity transactions
(F1/F2 evidence); (2) the slug update's row lock versus fact reads; (3) partial unique index maintenance on claims
(small); (4) no cache — if measured cost is unacceptable, an explicit-invalidation cache is a reviewed follow-up, not
added here. No benchmark campaign; bounded observations only.

### 5.6 Public contract wording — as r1 (dated rows in `connection-context.md`; setup-listing note in MCP.md).

### 5.7 sqlc trial — one bounded executable trial (review item 7)

Scope, config and conventions as r1 §5.7 (one package `core/internal/domains/db`; authoritative Goose schema
`schema: ["../../../migrations"]` relative to the config; pgx/v5; `emit_methods_with_db_argument`; UUIDs as `string`
through the repository's `::text::uuid` casts; nulls as pointers; committed deterministic output; generated methods
run only on the caller's guarded `pgx.Tx`/runtime role; no `go.mod` change).

- Command (b1, from the repository root): `docker run --rm --network none --user "$(id -u):$(id -g)" -v
  "$PWD/core:/src" -w /src sqlc/sqlc:1.29.0 generate -f internal/domains/db/sqlc.yaml` (host-owned output, removable
  by the operator); reproducibility: a second run then `git diff --exit-code -- core/internal/domains/db/`.
- **One trial, bounded to one b1 iteration:** generate → `go build ./internal/domains/db/` → `go vet` of that package
  → the package's own compile-time use in `access/domains.go`. Success = adoption. Any of: parser error on the
  migration set, a construct that would need a second schema file, a mapping that needs a new module, or generated
  code that cannot satisfy the facade's transaction contract → **deferral recorded in the Result with the exact error
  text**, `sqlc.yaml`/generated files deleted, the same query file kept as handwritten pgx constants in
  `domains/store.go`. A bad command or wrong path is fixed within the iteration and is not evidence of
  incompatibility. No day-long generator work.

## 6. Validation batches and test mapping

### 6.1 Existing suites inspected — as r1 §6.1 (suite-only harness selection; the check image CMD is always overridden).

### 6.2 Scenario mapping — as r1 §6.2, with these changes

| Group | Test (path · check ID) | Kind | Change in r2 |
|---|---|---|---|
| 3 | DOM-6 freshness expiry (`TestDomains` · B-DOMAINS-IT) | **new** (as r1) | also covers stored TLS expiry (`tls_not_after` passed → routing false, issuance true while proof fresh), confirmed proof loss → issuance false immediately, TLS-only suspension keeps issuance, **and (r3, item 1) a TLS-suspended row whose recheck confirms proof loss: stays `suspended`, `verified_until = now()`, issuance false at that commit although its previous deadline was days away** |
| 3 | DOM-4 races | **new** | adds the late adapter result for a removed generation (dropped, `stale_result`), the replacement-claim issuance (403 until its own proof), the removed-row cleanup update (only the allowed columns) **and (r3, item 2) the explicit retry: `unknown` → `RetryWithdrawal` → `withdrawn` on the old generation's own resource while the replacement claim is untouched; retry on a `withdrawn` or live row → `domain_state_invalid`; `DOMAIN_WITHDRAWAL_RETRY_MAX` reached → `domain_retry_exhausted`; ordinary recheck of a removed row → `domain_state_invalid`; routing/issuance false throughout** |
| 3 | `TestDomainStateMachine`, `TestDomainsCatalog` negatives | **new** | mirror the §1.4 table (pending failure stays pending; rotation pending-only; forward-only `verified_until` except the confirmed-loss decrease to `now()` from `verified`/`active`/`suspended`; removed-row column set incl. increasing `check_failures`) |
| 4 | `TestCustomDomainRelay` (`identity/relay_test.go` · B-IDENTITY-IT) replaces r1's `TestCustomDomainHandoff` | **new** (B2 only; the guard cases also run under B1) | relay to a routing-ready custom host only (404 for unknown/reserved/platform/removed), consumption on T with the binding, replay after consumption, foreign browser, altered state, `callback_origin` mismatch, host reassigned/removed before the exchange (`host_changed`, no exchange, no session), membership denial (`organization_not_member`); **r3 (items 5, 6)**: removal committed inside the fixture provider's `Exchange` → `host_changed` at issuance, no session row, one denied event (case A); membership revoked inside `Exchange` → `organization_not_member` (B); sign-in then removal → next `Session(?origin=T)` is `unknown_host` (C); continuity branch on a reassigned host leaves the live session untouched (D); provider `error=access_denied` relayed byte-identical → `login_canceled` on T; repeated `state` on A → fixed 404, nothing consumed; repeated `code` relayed → `ambiguous_parameter` on T; relay/callback responses carry `Cache-Control: no-store` and `Referrer-Policy: no-referrer` (asserted in the web unit test row below; the core JSON rows assert no-store) |
| 4 | `web/tests/unit/session.test.mjs` (B-UNIT-WEB content; run directly) | extend | `seeOther` and `statusPage` responses carry `Cache-Control: no-store`, `X-Content-Type-Options: nosniff` and `Referrer-Policy: no-referrer` (r3, item 6) |
| 5 | IT-MCP-3 setup/activation and IT-MCP-8 listing (`mcp_integration_test.go` · B-MCP-IT) | extend | mixed-host list through a current host includes the old-host row; revoke of that row through the current host; create still exact-origin |
| 5 | `web/tests/unit/mcp-client.test.mjs` (B-UNIT-WEB content; run directly) | **new** | parser accepts other-host `RESOURCE` URLs with matching id, rejects mismatched id/malformed; create keeps exact origin (no existing parser test) |
| 5 | `mcp.spec.ts` M-LIST (B-IDENTITY-E2E) | extend | one seeded old-host row shown with its host |
| 4 | `TestStagingContract` (`core/internal/identity/workos/staging_test.go`, tag `workos_staging` · B-IDENTITY-STAGING) | extend | new subtest: wildcard-covered third host authorization URL/callback acceptance and the sign-out return rule as configured (P-1); **positive evidence required**: `--- PASS: TestStagingContract/<subtest>` lines ≥ 1, `ok … (N tests)` with N > 0; a "no tests to run" result fails |

Everything else in r1 §6.2 stands (reuse/extend/new rows and the justification summary).

### 6.3 Batches (executable selectors; review item 9)

Before every batch: fetch `origin/staging`, assess the delta, rebase/reconcile, `bash infra/ci/docker-desktop-mounts.sh`,
coordinator batch lock, standard `run.sh` form with the G016 allow-list. Finish, fix, rerun, advance; no edits during
a running batch; **no duplicated final suites** — final qualification assesses the delta since the last attributable
pass and reruns only affected failures.

| Batch | Executable selectors | Reason / evidence |
|---|---|---|
| b1 static and unit | check image `docker build --target check -t g016-core-check:b1 -f infra/docker/core.Dockerfile core`; `docker run --rm --network none --entrypoint sh g016-core-check:b1 -c '…'` with: `gofmt -l .`; `go vet ./internal/hostname/ ./internal/domains/... ./internal/tenancy/access/ ./internal/identity/... ./internal/mcp/ ./internal/config/ ./internal/audit/ ./internal/tenancy/ ./cmd/server/`; `go vet -tags integration ./internal/domains/... ./internal/tenancy/access/ ./internal/identity/ ./internal/mcp/ ./internal/migrate/`; the per-package `go test -count=1 -run '^(…)$'` selectors of r1 b1 plus `-run '^(TestHostnameNormalize\|TestHostnameClassify)$' ./internal/hostname/`; web (web-check image, `docker run --rm --network none --entrypoint sh <web-check> -c 'node --conditions=react-server --test tests/unit/session.test.mjs tests/unit/host.test.mjs tests/unit/mcp-client.test.mjs tests/unit/identity-client.test.mjs'`) | changed packages only; the changed web unit files run directly (no `unit-web` umbrella) |
| b1 sqlc (if adopted) | §5.7 command, build, vet, diff | one bounded trial |
| b1 harness | `run.sh lint lint-base lint-core` | the static suites over changed sources (B-ENV-EXAMPLE reviewed domain block, B-MIGRATE-SET, B-SHELL, B-SECRET-SOURCE/CANARY, B-LITERALS/B-NAMING/B-DEPSCAN/B-IDENTITY-GUARDS, B-MCP-SOURCE); B-SCOPE comes with the run |
| b2 focused integration | focus database `g016-focus` (as r1): `go test -tags integration -count=1 -timeout 30m -v -run '^(TestDomainsCatalog\|TestMCPCatalog\|TestMCPMigration\|TestAuditCatalog\|TestRLSCatalog)$' ./internal/migrate/`; `… -run '^(TestDomains\|TestMCPOrdering\|TestMCPFootprint\|TestAccessPolicy\|TestAccessOrdering\|TestRLSIsolation\|TestRLSOrdering\|TestAudit\|TestIsolationAttacks)$' ./internal/tenancy/access/`; `… -run '^(TestMCPTransport\|TestMCPLifecycle\|TestMCPAudit\|TestMCPRecovery\|TestMCPDiscovery)$' ./internal/mcp/` (expect 7/6/1/1/5 with the `mcp_it` env); `… -run '^(TestCallbackRefusals\|TestCallbackContinuity\|TestCustomDomainRelay\|TestLocalLogoutAndCSRF\|TestProviderLogout\|TestRLSIdentityPool\|TestConcurrentFirstSignIns)$' ./internal/identity/` | the new catalog/runtime-role/race evidence and the affected accepted regressions; `E-DOM`/`E-MCP` lines kept. Unaffected `migrations` rows (tenancy model, retention, drills, objects) are **not** rerun locally; their attributable passes are the staging gate at the merged predecessor and are reconciled by input, not borrowed |
| b3 `ops` (**single exception**) | one `run.sh ops` after b2 and the S-review | **Actual fixture dependency**: B-IDENTITY-E2E/PROBES/CONFIG and B-MCP-E2E need the identity overlay (`infra/compose/identity.yml`: per-run CA/leaf, nginx TLS edge on 543 with the seeded hosts, emulator with `--redirect-hosts`, `identitysetup` users/memberships, web netns host mappings); no narrower entry point exists (BACKLOG "Narrow identity/MCP E2E selection", open) and this Plan adds none. **Smallest alternative if declined**: no local `ops`; those rows come from the hosted staging full gate after the authorized conditional merge, recorded hosted-only; the browser evidence for manual feedback then comes from the hosted job's artefacts. No second final `ops` run either way |
| b4 rename | `run.sh rename` once | the only selector for the five B-*RENAME*/B-SOURCE-HASH/B-PORTS-OVERRIDE rows; all affected by the configuration change |
| S-review | read-only (§7) | findings fixed → rerun only the affected `-run` expressions |
| provider / manual | after P-1: `docker run --rm --network "<project>_data" --env-file "$IDENTITY_STAGING_ENV_FILE" <project>-core-check go test -tags workos_staging -run '^TestStagingContract$' -count=1 -v -timeout 600s ./internal/identity/workos/` (B-IDENTITY-STAGING on the `identity-qualification` branch; positive non-zero evidence as §6.2); after P-2: manual browser qualification (IDENTITY.md L4 pattern) of custom-host sign-in/sign-out against WorkOS staging | real provider/manual evidence for AC7/AC8 |
| final (local) | **no new suite run**: reconcile the delta since the last attributable pass of each batch; rerun only the `-run` expressions and suites whose inputs changed since; collect the Result | attributable coverage = earlier passes + targeted reruns, with provenance recorded |

MR fast runs automatically; the path map selects **full**; the post-merge staging full gate supplies the hosted
rows after an authorized conditional merge.

### 6.4 Truthful qualification boundaries and prerequisites (review item 10)

| Requirement | What fixtures prove | What they cannot prove | Remaining prerequisite and authority |
|---|---|---|---|
| AC6 production adapter | adapter contract, `edge-probe` against a local TLS listener, generation binding, removal/late-result handling | a live certificate provider or deployed ingress | **P-4**: D-1 mechanism provisioned by the owner's separate deployment authority; otherwise AC6 closes with an **explicit held gate** stated in the Result. This gate does **not** waive AC7/AC8 |
| AC7 provider scaling | three platform hosts in the emulator only where `--redirect-hosts` admits them (fixture limitation) | WorkOS wildcard redirect **and** sign-out behaviour; "no per-Organization entry" | **P-1**: owner registers the wildcard entries (and the auth-host entries under B2) in the WorkOS staging dashboard (a provider settings change needing explicit owner authority), then B-IDENTITY-STAGING with positive evidence and sanitized before/after configuration |
| AC8 custom-domain browser flow | the relay and consumption against the emulator over the overlay TLS edge at `portal.acme.test`; the final issuance guard and the response headers (§1.7.1/§1.7.2) | real provider acceptance of the shared callback, real TLS at a customer host, the production edge's log configuration | **P-2**: a staging deployment reachable at the hosts with real TLS (deployment authority) whose edge log omits or redacts `code`/`state` on `/auth/callback` (§1.7.2, confirmed by the operator and recorded), then manual qualification (L4) and operator feedback before final CI |
| AC4 real DNS proof | all negative cases and the positive case with the fixture resolver | a live public DNS record | **P-3** (optional): one bounded TXT/CNAME proof at an owner-controlled domain under explicit authority |

No live action, provider settings change or deployment is performed or authorized by this Plan.

## 7. Delegation — as r1 §7 (one implementer at high; required read-only security reviewer after b2, scope now
including the relay's two legs and the generation-bound adapter; optional second reviewer only for E2).

## 8. Sequence — as r1 §8, with step 7 reading "resolver-backed origins, membership rule (D-2), platform wildcard
pattern, **B2 relay** (if selected) with `TestCustomDomainRelay`, **the final issuance guard in `signIn` and the
`NO_STORE` header change (§1.7.1/§1.7.2)**; web host gate, `proxy.ts`, relay branch in `auth/callback/route.ts`;
mixed-host parser and page", step 5 (lifecycle) including `RetryWithdrawal` and the `suspended → suspended` edge, and
step 9 ending with targeted reruns only.

## 9. Acceptance-criteria mapping — as r1 §9, with AC5 adding DOM-4's late-result/replacement **and withdrawal-retry**
cases, AC6 citing the issuance predicate (DOM-6/7, **including the already-suspended loss case**) and §6.4's gate,
AC8 citing `TestCustomDomainRelay` (**with the post-exchange race and privacy cases**) and `session.test.mjs`, AC9
adding the mixed-host listing (IT-MCP-3/8, `mcp-client.test.mjs`, M-LIST) and AC11 citing the §5.5 inventory and
measured `E-DOM-F` lines.

## 10. Risks and escalations — as r1 §10, plus

- The relay places the provider's authorization code in one additional same-TLS browser redirect; it is one-use and
  PKCE-bound and the destination is resolution-derived. If the security reviewer or owner rejects any code transit
  through a shared host, B1 is the fallback (operator entries per custom domain).
- One autocommit statement per MCP/metadata/anonymous/relay request and +1..+3 statements inside the identity
  transactions from the durable lookup (no cache in v0) — owner attention (§5.5).
- Generation-bound adapter semantics depend on D-1: E1 has no per-claim provider resource (withdrawal = predicate),
  E2 has one (deletion by id); the explicit retry (§1.5) is E2-only in effect and bounded per row.
- The issuance guard's one-request window (§1.7.1) is accepted by design: closing it would need a lock on domain or
  binding rows from the IDENTITY writer (new lock edges) or a session-revocation write path (D-4(b) territory); either
  is an escalation, not a silent addition.
- Under D-2(a) a first-time principal cannot complete sign-in on a tenant host (§1.9); if the owner needs tenant-host
  onboarding, D-2 must be revisited before Plan ACCEPT.
- The production edge log configuration (§1.7.2) is outside this repository's control until D-1 is decided; the
  Result records it as a P-2 prerequisite, not as satisfied.

## 11. Budget, forecast and scope

### 11.1 Budget — as r1 (B7/checkpoint 4; forecast 2 likely, 3 plausible).

### 11.2 Count pins (re-pinned at R0, r4) — as r1 §11.2, plus `TestDomains` 10 subtests with 4 `E-DOM-F`
lines (DOM-4 and DOM-6 gain cases, not subtests), `TestCustomDomainRelay` **11** subtests (the r2 seven plus cases
A–D of §1.7.1; the provider-error, repeated-key and header cases join existing subtests), `TestStagingContract` +1
subtest, `session.test.mjs` +1 test, and `TestProductionImporters` pin widened to two importers with two added cases.

R0 confirmation of the "after G015" column at `f6d79af`: required-checks rows **117** (file byte-equal to the G015
R0 row that recorded 117/33/14); B-MCP-IT 7/6/1/1/5/**10**/1 and E-MCP lines 5 (`run.sh:3148` `mcp_it`); AUDIT-INV
31 and trigger-owned 25 (`audit_test.go:75`); RLS-SCOPE 36 (`run.sh:2849`); profiles 35/75/110 and VA7 80
(`validate-controls.sh`/`lib.sh` byte-equal to the G015 dispatch rows); B-IDENTITY-E2E identity projects 4
(`playwright.config.ts:95-98`: identity/mcp × light/dark; one configured `baseURL` host `org-a.localhost:543`, the
second host navigated in the spec); ISO-ORACLE 82 is read from the test log by `run.sh`, not from source, and is
re-pinned by the implementer at b2. The "after G016" column is unchanged.

### 11.3 B-SCOPE allow-list (imported by Codex at `f6d79af` as `management/v0/results/G016-scope-allowlist.txt`, SHA-256 `047450e1…aa0`, byte-equal to this block; unchanged in r4)

```text
# Create
^core/internal/hostname/[a-z_]+\.go$
^core/internal/domains/([a-z_]+\.go|api/[a-z_]+\.go|db/(sqlc\.yaml|[a-z_.]+\.go|queries/[a-z_]+\.sql))$
^core/internal/tenancy/access/domains(_integration_test)?\.go$
^core/internal/migrate/domains_integration_test\.go$
^core/migrations/00011_organization_domains\.sql$
^core/internal/config/domains(_config_test)?\.go$
^core/internal/identity/relay(_test)?\.go$
^web/lib/host\.ts$
^web/proxy\.ts$
^web/tests/unit/(host|mcp-client)\.test\.mjs$
^infra/identity/(dns|tls)-fixture\.tsv$
^docs/foundation/DOMAINS\.md$
^management/v0/results/G016-(scope-allowlist\.txt|hostname-resolution-domain-lifecycle-result\.md)$
# Modify: production
^core/cmd/server/main\.go$
^core/internal/config/(config|identity|mcp)\.go$
^core/internal/mcp/(registry|handler|connections)\.go$
^core/internal/identity/(allowlist|service|store|http|logout|identity)\.go$
^core/internal/tenancy/access/(policy|errors|log)\.go$
^core/internal/tenancy/access/access\.go$
^core/internal/tenancy/slug\.go$
^core/internal/audit/audit\.go$
^web/app/(auth/(callback|login|logout|logout-all|logout-return|provider-logout)/route\.ts|sign-in/page\.tsx|account/page\.tsx|connections/mcp/page\.tsx)$
^web/lib/(session|identity-client|mcp-client|config)\.ts$
# Modify: tests and fixtures
^core/internal/config/(identity_config_test|mcp_config_test)\.go$
^core/internal/mcp/(mcp_test|mcp_integration_test|env_integration_test|e2e_mcpe2e_test)\.go$
^core/internal/identity/(allowlist_test|service_test|integration_test|logout_test|http_test)\.go$
^core/internal/identity/workos/staging_test\.go$
^core/internal/tenancy/access/(boundary_test|matrix_test|policy_test|rls_expected_test|attack_expected_test|audit_integration_test|integration_test)\.go$
^core/internal/tenancy/slug_test\.go$
^core/internal/audit/audit_test\.go$
^web/tests/unit/session\.test\.mjs$
^web/tests/browser/(identity|mcp)\.spec\.ts$
^web/playwright\.config\.ts$
^infra/compose/(identity|local)\.yml$
^infra/identity/(proxy\.conf|mcp-seed\.sql|mcp-hosts\.tsv|emulator-seed\.yaml)$
^infra/validation/tools/cmd/identitysetup/[a-z_]+\.go$
^infra/validation/tools/internal/logrecord/events\.json$
^infra/local/configure-env\.sh$
^\.env\.example$
# Modify: harness
^infra/validation/run\.sh$
^infra/validation/required-checks\.tsv$
# Modify: docs and contracts
^docs/foundation/(IDENTITY|MCP|AUTHORIZATION|AUDIT|MIGRATIONS|DATABASE-ISOLATION|OPERATIONS|VENDORS)\.md$
^docs/(SECURITY|PERFORMANCE|ARCHITECTURE)\.md$
^contracts/mcp/connection-context\.md$
```

Conditional (Codex-reviewed amendment only): `infra/validation/render-customizations.tsv` and `core/go.{mod,sum}`
(not expected); `infra/ci/validate-controls.sh` (pin text only); `skills/saas-foundation/**` guidance; R0 hits.

## 12. Open questions for Codex (C-n)

- C-1 **settled (r2 review)**: at most one local `ops` exception for the selected browser/TLS composition on the
  documented identity-overlay dependency; no duplicate final `ops`/umbrella run; the hosted-only alternative stays
  available; changed web tests run directly (§6.3).
- C-2 **settled**: B-DOMAINS-CATALOG and B-DOMAINS-IT map distinct catalog/runtime evidence; no further control campaign.
- C-3 **settled**: the one-package sqlc trial with the concrete fallback (§5.7); no further owner prompt.
- C-4 `connection-context.md` dated rows and the MCP.md setup-listing note (mixed hosts): **waits for the D-4
  disposition**; the wording authorizes no retargeting of existing connections.
- C-5 PG-007 entry wording: record the §5.5 execution model, inventory, index/lock concerns and the measured
  `E-DOM-F` observations at implementation; no performance guarantee and no implicit authority-cache addition.
- Assumptions to confirm at R0: `TestMCPCatalog` migrates to its own version; no Organization slug on accepted staging;
  G015's final content adds no migration/audit action; the Duo-corrected `Handler.fixture` seam is merged as read.

## 13. Revision history

| Revision | Date | Change |
|---|---|---|
| r1 | 2026-10-16 UTC | Provisional Plan at G015 MR !33 (`c7ce0bb`), base `d36f2a7`; archived byte-exact as `-plan-r1.md` |
| r2 | 2026-10-16 UTC | Correction loop (r1 review (not included)): items 1–10 and the G015 Duo reconciliation per the retained r2 note; D-5/D-6 recorded as settled; D-1..D-4 decision brief; base `b8ceeab`; archived byte-exact as `-plan-r2.md` |
| r3 | 2026-10-16 UTC | Correction loop (r2 review (not included)): the six remaining items per the r3 note (§1.3/§1.4 already-suspended loss; §1.5 explicit withdrawal retry; §5.0.1 `access.go` wiring; §5.5 real lookup execution model; §1.7.1 final issuance guard; §1.7.2 relay privacy/refusals); C1–C3, D5, D6 recorded as settled; owner design package recorded as pending with no answer assumed; base `36c7f4e`; **accepted in full with the owner's design approval at `f6d79af`**; archived byte-exact as `-plan-r3.md` |
| r4 | 2026-10-16 UTC | R0 (narrow; r3 review/acceptance (not included)): §0.1 gates, §0.2.1 R0 record against the merged G015 code (nothing material; 00011 free), §0.4 owner decisions (ADR-007: E1 Caddy, D-2(a), D-3(a)+B2, D-4(a)) and Codex implementation constraints, §2 rebind at `f6d79af`, §11.2 pins confirmed; initial Result drafted; base `f6d79af` |
