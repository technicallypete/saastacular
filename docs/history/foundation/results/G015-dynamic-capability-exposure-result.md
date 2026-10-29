> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G015 — Dynamic capability exposure and stale-context handling: Result

## Status

**Integrated, with closeout complete for final review; final ACCEPT belongs to Codex.** MR !33 merged as `7e0ac44`
(exact head `b80c283`, code identical to the accepted attempt-2 `2ef5fcd`). Staging pipeline 2954782898 succeeded on its
first run. See "Integration and closeout" at the end. The statuses below are kept as history.

**Earlier (2026-10-16 UTC): attempt 1 of B7 implemented. b1 and b2 focused validation PASS. Ready for the required
S-review (read-only security and concurrency). Not final, not accepted.** Details are in
[Attempt 1: implementation and focused validation](#attempt-1-implementation-and-focused-validation). The paragraph
below is the initial status at R0, kept as history.

**Initial Result (R0): reconciliation done (nothing material); execution not started; not accepted.**

- This is the execution-readiness record that Plan r4 §0.1 asks for, ahead of Codex's review of r4, the Git-first
  binding import and the recorded execution release. It is not implementation, not review-ready and not acceptance.
- Nothing has run: no build, test, harness, Docker, provider call, probe or migration. No delegate has been
  dispatched and no MR has been opened. Implementation attempt 1 of B7 (checkpoint after attempt 4) has not begun.
- Implementation starts only after (1) G014's recorded final ACCEPT/closeout, (2) Codex's review of Plan r4 (the R0
  record), (3) the binding import (allow-list, this Result, target, budget, resources) and (4) the recorded release.

## Execution binding (Git-first; prepared for Codex to import)

Hash rows carry neutral labels; the paths are listed under "Bound paths". The full input table is Plan r4 §2; the
rows below are the ones the binding gates on.

| Label | Bound bytes |
|---|---|
| Target: `origin/staging` at R0 fetch, 2026-10-16 (contains the merged G014 `a8d146c` and its qualified gate) | `f45ad716569cd791756a9786e976b3d1f1a82b4a` |
| Plan r4 (P4), the R0 record proposed for review | `5b6290662d14f594b52c4720df8dfe8f63e9ae5db1d3bdc508041e7b11033cc6` |
| Plan r3 archive (P3), provisional design ACCEPT | `8e1718b5a04e67ea8bdfa6a0aba40610745821e5d478e0e58fc89d831f4ce9ba` |
| Plan r3 review (R3) | `456361c86444721a42fbaaf5fc46160942749b4d8c9ae972c685be67fdc5e13d` |
| Goal (G) | `ca32a84d3a0144feb04446293fab8376720a97e7b56c2f87d007f9e4b7a7dbc0` |
| Spec (S) | `1f7149ff2205549c2d7a826f2079efeee8d6e9bbf1425bc31828d305457016b4` |
| G014 Result (X14) | `a40111023744d13b738b4bc0230e31f053fc7012406dcecc9e6adffe4f47c225` |
| G014 implementation review incl. the hosted-retry closeout record (I14) | `97ce874d3d344d5d0f15a592db1a805bbe1d30a4e77a39e42dc385379173ca46` |
| G014 F3 correction review (F3R) | `9b0b374b6456911d5542b37e95618b0e98df0fbff281ee95da9c2c80c342ebb5` |
| MCP control-plane contract (CP) | `5e77d3bbc9c946b853a35edb6b31ff7ff48cd632146d7305b070eae0d60238da` |
| Validation-integration reference (VI), with the focused-selection policy | `696df897dcea6eb599a025adb8a22363d94845e4a411af1173292fa285b192d5` |
| Harness (RS) | `2718f944dd2325b7cff7deb7c673123b7c076c87a4434eba587ef303452e9d7f` |
| Required checks (RC) | `2d136fea2fa096def63d29a5182fb563b6419f94babf519c8e61ed3455a1c686` |

**Bound paths.** P4/P3 `management/v0/plans/G015-dynamic-capability-exposure-plan{,-r3}.md`; R3
`management/v0/results/G015-plan-r3-codex-review.md`; G `management/v0/goals/G015-dynamic-capability-exposure.md`;
S `management/v0/specs/G015-dynamic-capability-exposure-spec.md`; X14
`management/v0/results/G014-tenant-bound-mcp-connections-result.md`; I14
`management/v0/results/G014-implementation-codex-review.md`; F3R
`management/v0/results/G014-f3-correction-codex-review.md`; CP `contracts/mcp/control-plane.md`; VI
`skills/sdlc-management/references/validation-integration.md`; RS `infra/validation/run.sh`; RC
`infra/validation/required-checks.tsv`. The B-SCOPE allow-list is Plan r4 §11.3, to be imported byte-exact as
`management/v0/results/G015-scope-allowlist.txt`.

**Budget and effort.** B7, checkpoint after attempt 4 (Goal); planning and R0 spent no attempt. Implementation
effort high (default); the owner may choose xhigh for the ordering work (O-MCP-7/8); the actual runtime setting is
reported as unknown unless verifiable.

**Worker record.** R0: task G015, role plan/R0; native isolated worktree `.claude/worktrees/agent-<id>`
(generated `agent-<id>` name, the accepted AGENTS.md fallback), branch `g015-r0` from `f45ad71`; successor to the r3
planner after the 350k active-context handoff (staging `cee09c9`); session
`<agent-session-url>`. Implementation: one worker, `g015-agent-<id>-impl` when
the native tool supports it, otherwise the generated name with task/role/path recorded here.

## R0 reconciliation summary (Plan r4 §0.2 R0 record, §0.3)

| Item | Outcome |
|---|---|
| P-G014-1..13 against the merged code at `f45ad71` | all **confirmed** with file:line; immaterial drift recorded (third inert seam `selectDecided`; `mcp.setup` at `connections.go:96`; `mcpLocks` is a helper; `TestCursor` added) |
| P-G014-5 (F3) | implemented as the Codex-accepted design: one Organization-only guarded write locking the previous and target Group facts; the 00010 guard requires `ctx_group` NULL and visible same-Organization previous and target Groups; a raced loss of the previous Group is `context_denied`; G015 adds no lock, SQL or guard change |
| P-G014-9 | 00010 is the last migration; **00011 is the next free number, unused** |
| P-G014-11 | `TestMCPOrdering` 8/8 at the base (G014's O-MCP-5/6); G015's cases renumbered **O-MCP-7/8**, pin 8→10; B-MCP-CATALOG 3/5 (CT-MCP-GUARD); B-MCP-E2E 6; required checks 117/33/14; profiles 35/75/110, VA7 80 |
| Setup-listing changes since the r3 baseline (events, `csrf_token`, pagination, 1,277,222-byte BFF bound) | no G015 assumption depends on the setup listing; outside the allow-list |
| `unavailable` wording | F-3/C-4/N-3/N-5 and CAT-2 reconciled with the corrected contract: a definitive refusal or known abort means nothing changed; `unavailable` after a possible commit means reread before retrying; CAT-2's "unchanged" is scoped to the definitive cases |
| Executable selectors | named per local run in Plan r4 §6.3; the `unit`/`migrations` suite dispositions and the `ops` exception are requested before launch (Q9); `TestMCPCatalog` naming note (Q8) |

## Implementation summary and files changed

None yet. Planned scope: Plan r4 §3 and §11.3.

## Migrations/schema changes

None planned (Plan r4 §4); none made.

## Contracts/API/MCP changes

None made. Planned: additive F-1..F-5, C-1..C-6, N-1..N-5 appended to `contracts/mcp/control-plane.md` (E-1), with
the r4 `unavailable` wording aligned to the contract's corrected `switch_group` and error-table text.

## Tests added

None yet. Planned (Plan r4 §6): `TestToolEligibility` (unit), `TestMCPCatalog` in package `mcp` CAT-1..5 (or
`TestMCPDiscovery`, Q8), `catalog_fixture_test.go`, `TestMCPOrdering` O-MCP-7/8, `TestMCPFootprint` E-MCP-4, unit
and boundary pins.

## Exact commands/validation performed and outcomes

None. R0 was read-only (`git fetch`, `git diff --stat c8f8f9af origin/staging` over the §0.2 paths, `git log`,
`sha256sum`, file reads). Planned local runs and their executable selectors: Plan r4 §6.3 (b1 unit `-run`
expressions and `run.sh lint lint-base lint-core`; b2 focused `go test -tags integration -run …` on a disposable
focus database; final `run.sh lint lint-base lint-core unit migrations`; `ops` once as the requested exception, or
hosted-only E2E evidence if declined). Before every batch: fetch/rebase/reconcile, `bash
infra/ci/docker-desktop-mounts.sh`, the coordinator batch lock, no edits during a running batch.

## Acceptance-criteria checklist

| AC | Status | Evidence (planned, Plan r4 §9) |
|---|---|---|
| AC1–AC9 | not started | as mapped in Plan r4 §6.2/§9 |
| AC10 | in preparation | this Result; the allow-list; the B7 ledger; the cleanup receipt |

## Architecture deviations

None. R0 found no contradiction with ARCHITECTURE.md, SECURITY.md, AUTHORIZATION.md (incl. the F3 lock-order
paragraph), ADR-005 or the accepted MCP contract.

## Unresolved issues, blockers, attempted approaches, cycle history

- Blockers to execution (gates, none owned by this worker): G014 final ACCEPT/closeout record; Codex review of Plan
  r4; binding import; recorded release.
- Open for Codex disposition (not blockers to review): Q8 test naming; Q9 the `unit`/`migrations` suite dispositions
  and the `ops` exception.
- Planning cycles: r1 (CORRECTION LOOP), r2 (CORRECTION LOOP), r3 (ACCEPT, design only), r4 (R0 record, this
  hand-back). Execution attempts used: 0 of 7.

## Suggested follow-up (not authorized by this record)

- A narrow harness entrypoint for the identity/MCP E2E orchestration (BACKLOG "Narrow identity/MCP E2E selection")
  would remove the recurring `ops` exception; separate CI work, not G015.

## Execution binding and dispositions (coordinator-dispatched)

Recorded at dispatch, 2026-10-16 UTC, before any code write, under the
r4 review and execution release (not included) (SHA-256
`9863c00659e106b917ba1fca66decf34252a34dfbe69ae5e997c0b254f11f334`), which supersedes the pending/stale binding
labels above.

**Target.** origin/staging `0900ee5e6fa40e4d28d5d1b884762b7cef2ff06f` (OBJ-LOG MR32 merged). Overrides copied
verbatim from the review:

| Path | SHA-256 at execution target |
|---|---|
| `management/v0/PROGRESS.md` | `3192ecc3c7af58fc15dfa7f01c8f7d683d8a10b7fe221134d6b893700fe87df0` |
| `management/v0/results/G014-tenant-bound-mcp-connections-result.md` | `43161a6a31f906d7acd6f68d6c77977ea815305b27218f6c00216b9571accb5c` |
| `management/v0/results/G014-implementation-codex-review.md` | `563404ff0905841eb356b411c1e10e5af3e407199a4013e93d5b930149105f95` |
| `infra/validation/run.sh` | `79af82fb86a3f290db88f484383a22a268f9742f6cae3e3db6f631f9f94f1f52` |
| `infra/validation/lib.sh` | `03d300759b62f53c32796a4345303c893d44946419d3761e7ff7fad3dc2532e4` |
| `infra/ci/validate-controls.sh` | `04319fa9368cd9523e09c3e745953bbf7afc7cf45b428dbd38292b0af5deb1eb` |

Imported bytes: Plan r4 `5b6290662d14f594b52c4720df8dfe8f63e9ae5db1d3bdc508041e7b11033cc6`; r3 archive
`8e1718b5a04e67ea8bdfa6a0aba40610745821e5d478e0e58fc89d831f4ce9ba`; initial Result (before this section)
`ee0a4d23b671a278d4b12a87e2181fc8ca22a912e3a95c9f2518394a67a821ad`; scope allow-list
`eeb84eb34f0d983be41e788f4c3f307ca99b18435aaa389dc3cf09ef40bc1986`.

**Checkout verification (read-only, before writes).** Branch created from origin/staging
`dba74d3dc4c5f80ec44956939ec07f32e3e642bd`; `17432d8` (release) and `0900ee5` (target) are ancestors. The
`0900ee5..dba74d3` delta is management only: PROGRESS, the r3 archive, Plan r4, the initial Result, the r4 review and
the allow-list. Of the 47 Plan r4 §2 rows, 43 match the checkout byte-exactly (D3–D6 included). X14, I14 and RS match
the override table above; `lib.sh` and `validate-controls.sh` match their override rows. PROGRESS reads
`edb076b133963054361192aab436e20984500dd7f89f7e9f1efecf568f6a91bd`, not `3192ecc3…`, because Codex's release
(`17432d8`) and progress-index (`dba74d3`) commits updated it after the target. That is a Codex-owned management
record, not an application input, so it is reported and not treated as a blocker. No application, harness or
contract mismatch was found.

**Q8.** `TestMCPDiscovery` holds the five CAT scenarios (CAT-1..5) in package `core/internal/mcp`. The migrate
package's `TestMCPCatalog` (B-MCP-CATALOG) is unchanged. Only the approved B-MCP-IT regex, count and text change.

**Q9.** Local final `unit`, `migrations` and the `ops` exception are declined. Local evidence is b1 (focused core
gofmt, scoped vet and the two `-run` unit sets in the check image, with the default CMD explicitly overridden), the
`lint lint-base lint-core` static pairing, and b2 (focused MCP and policy integration selectors on a disposable focus
database). B-MCP-E2E, B-IDENTITY-E2E and the full driver are hosted-only, supplied by the required staging gate.

**Worker.** Task G015, role implement, attempt 1 of B7 (checkpoint after 4). Native isolated worktree
`<source-checkout>` (generated
`agent-<id>` name, the accepted fallback). Branch `g015-dynamic-capability-exposure` from `dba74d3`. Sole writer:
no other worktree holds this branch. Model Claude Opus 5.5 (`claude-opus-5-5`), assigned effort high, runtime
effort unknown. Session `<agent-session-url>`. No core-service skill exists
under `core/`, so repository governance and the foundation skill's MCP invariant (11) were applied.

## Attempt 1: implementation and focused validation

Attempt 1 of B7 (checkpoint after 4). The implementation round ran under one isolated worker. There was no
delegate, and the worker itself made no correction loop. Source head under test: `8dcfa38b66f836c184b6352e93f6cc7397ab1ef2`
on the base origin/staging `dba74d3dc4c5f80ec44956939ec07f32e3e642bd`, which was verified as an ancestor before every
batch. No staging advance occurred during the round.

### Implementation summary

| Area | Change |
|---|---|
| `core/internal/tenancy/access/mcp.go` | New exported names: the opaque, comparable `ToolRequirement` (zero value invalid) and its constructors `RequireBinding`, `RequireMembership`, `RequireContext` and `RequireCapability(name)`. `RequireCapability` accepts the eleven canonical values only; anything else is `invalid_input`. New method `AuthorizeCatalog(ctx, reqs) ([]string, error)`: one read-only snapshot through the new `snapshotFor`, which is G014's `snapshot` with the fact options of the requirements. It returns eligible names only. Its refusals, record and log are `AuthorizeConnection`'s (`opMCPAuthorize`; `access.denied`/`mcp.request` on a context refusal). New inert seam `mcpHooks.catalogComputed`, nil in production. |
| `core/internal/tenancy/access/policy.go` | New `eligible(req, sel, facts)`: `contextCheck` first. Control requirements always pass. Organization capabilities go through `decide` on the active Organization. Group and profile capabilities pass through the Organization owner/admin path from the binding alone (`all`/`Y`; `privileged_settings.manage` is owner-only) or through the Group-derived path, `decide` over the candidate Groups and their current associations. New `factQueryFor(reqs, sel)`: G014's `Groups: []`/`[selected]`, or, when a Group or profile capability is present, `Groups: nil`/`[selected]` with `AllProfiles: true`. `decide`, `contextCheck`, `linkedPath` and the matrix are unchanged. |
| `core/internal/mcp/catalog.go` (new) | The closed catalog spec (`catalogEntry{name, req, input, output, handler}`). `catalogSpec` validation refuses empty or repeated names, a zero requirement, a missing input schema or one whose type is not object, and a missing handler. It validates input-schema shape only; output schemas and schema closedness are not validated (qualified in attempt 2). The production spec is the five control tools with their requirements. Control entries are registered on every request (`registerControl`); capability entries only when AuthorizeCatalog names them (`registerEligible`). The test-fixture seam `extraEntries` is set only from `_test.go`. Attempt 2 replaced it with a per-Handler override (D2). |
| `core/internal/mcp/server.go`, `tools.go` | The middleware calls `AuthorizeCatalog` for `tools/list`, which is refused as a whole as before, and registers the eligible entries. A `tools/call` naming anything but a control tool runs the same snapshot. On a context refusal nothing is registered, so the SDK's own unknown-tool path answers for both kinds of name. A store failure answers `-32041`. `tools.register` is replaced by the catalog's registration. `ServerOptions` are unchanged (`ListChanged: false`, versions). |
| `infra/validation/run.sh` (B-MCP-IT) | Regex adds `TestMCPDiscovery`, count 5. `TestMCPOrdering` 8 to 10. The E-MCP grep includes `-4` and its count goes 4 to 5. PASS and FAIL texts updated. No new check ID, row or profile change. |
| Docs and contract | `contracts/mcp/control-plane.md`: the unknown-tool line plus a new section, "Discovery, stale context and notifications (G015)", with F-1..F-5, C-1..C-6 and N-1..N-5. `docs/foundation/MCP.md`: catalog component, evidence rows, a "Discovery and stale context" section that includes the disclosure that a cached catalog is a hint. `docs/foundation/AUTHORIZATION.md`: an "MCP discovery (G015)" section covering the names-only surface, the eligibility table and fact options, and the evidence. |

No migration, schema, index, RLS, grant, audit action, reason or operation, log event or field, cache, vendor,
production tool or capability. `events.json` is unchanged. Of the 47 Plan r4 §2 bound inputs, only the allow-listed
paths changed.

### Tests added or extended

| Test (package, check) | Kind | Cases |
|---|---|---|
| `TestToolEligibility` (access, B-UNIT-CORE) | new | 11 capabilities × 14 subjects × 3 contexts × 2 flag values × 2 link states = 1,848 cases. Each compares `eligible` on the facts that `factQueryFor` loads, and on the complete facts, with "some oracle target allowed by `decide`". It adds 19 literal pins (Organization admin without bindings, unlinked profile, linked+flag, link removal, owner-only paths), the control and zero requirements, the closed capability names and the `factQueryFor` shapes. |
| `TestAccessBoundary` (access) | extend | five new exported declarations and `Service.AuthorizeCatalog` pinned; the exact-equality pin also proves `AuthorizeTool` is absent |
| `TestMCPScopeBinding` (access) | extend | `AuthorizeCatalog` refuses a missing, browser, ended or foreign scope before the database and returns no names |
| `TestToolCatalog` (mcp) | extend | the production spec is the five tools with their requirements and no capability entry; a new handler serves it; the fixture spec is accepted; six invalid specs refused |
| `TestErrorMapping` (mcp) | extend | the SDK unknown-tool envelope (`-32602`, `unknown tool "<name>"`) pinned for a fixture name and a never-existing name |
| `catalog_fixture_test.go` (mcp) | new support | `test.org_policy`, `test.group_update`, `test.profile_update` and `test.audit_read`: canonical capabilities, closed schemas, invocation counters, no effect and no check |
| `TestMCPDiscovery` CAT-1..5 (mcp, B-MCP-IT) | new | as in Plan r4 §6.2; Q8 name |
| `TestMCPOrdering` O-MCP-7, O-MCP-8 (access, B-MCP-IT) | extend, 8 to 10 | O-MCP-7: a snapshot during a held switch, and one spanning its commit, return the old pair, and the next one the new pair. O-MCP-8: a downgrade, or a loss of the selected Group, committed after `catalogComputed` governs the next catalog. The production tools refuse on their own checks, and the clear still succeeds. |
| `TestMCPFootprint` E-MCP-4 (access, B-MCP-IT) | extend | eligibility statements per catalog shape, constant over 30 extra Groups and 30 extra linked profiles |

### Exact commands and outcomes

Every batch was preceded by `git fetch origin`, an ancestry check (origin/staging `dba74d3` is an ancestor of HEAD) and
`bash infra/ci/docker-desktop-mounts.sh` (12984 then 12988 of 100000, 12%). Each batch ran under `flock
<home>/.local/state/platform-evidence/local-harness.lock`. Evidence is kept under
`<home>/.local/state/platform-evidence/g015/`.

| Batch | Command (exact) | Outcome |
|---|---|---|
| b1 core | `docker build --target check -t g015-core-check:b1 -f infra/docker/core.Dockerfile core`, then runs with the image's CMD overridden (`--network none --entrypoint sh`): `gofmt -l` on the 13 changed Go files; `go vet ./internal/mcp/ ./internal/tenancy/access/`; `go vet -tags integration ./internal/mcp/ ./internal/tenancy/access/`; `go vet -tags mcpe2e ./internal/mcp/`; `go test -count=1 -v -run '^(TestToolCatalog\|TestErrorMapping\|TestPackageBoundary\|TestCursor)$' ./internal/mcp/`; `go test -count=1 -v -run '^(TestToolEligibility\|TestAccessBoundary\|TestProductionImporters\|TestMCPScopeBinding)$' ./internal/tenancy/access/` | **PASS.** Build rc 0. gofmt empty, all three vets empty. mcp 4/4 PASS; access 4/4 PASS. Image `sha256:cd236c72…6788` (`b1-r1/`) |
| b1 static | `infra/validation/run.sh --base dba74d3dc4c5f80ec44956939ec07f32e3e642bd --allow-file management/v0/results/G015-scope-allowlist.txt --owned-registry …/g015/registry/owned.tsv --report-dir …/g015/static-r2 lint lint-base lint-core` | **PASS**, RESULT PASSED, 17 required IDs each PASS once. Includes B-SHELL (edited `run.sh`), B-SECRET-SOURCE and B-SECRET-CANARY, B-LITERALS, B-NAMING, B-IDENTITY-GUARDS, B-DEPSCAN, B-MCP-SOURCE, B-DOCKER-ONLY, and B-SCOPE (18 changed paths, all allow-listed). `inventory-leftover.txt` is empty. A first invocation (`static-r1`) stopped at argument validation (`--owned-registry: not a registry file`, exit 2, before any work) because the registry file and its `cid` directory did not exist yet; it was created and rerun once |
| b2 focused integration | `b2-focus-it.sh` (G014 attempt 2 recipe): Compose project `g015-focus`, postgres only from `infra/compose/local.yml`, generated 0600 env files (removed afterwards), the run.sh `rls_go` env shape (canary file, access log, `MCP_LOG_FILE`, `MCP_INVALIDATE_SQL_B64`), the b1 image on `g015-focus_data`. In harness order: `go test -tags integration -count=1 -timeout 30m -v -run '^(TestAccessPolicy\|TestAccessOrdering)$' ./internal/tenancy/access/`; `… -run '^(TestMCPTransport\|TestMCPLifecycle\|TestMCPAudit\|TestMCPRecovery\|TestMCPDiscovery)$' ./internal/mcp/`; `… -run '^(TestMCPOrdering\|TestMCPFootprint)$' ./internal/tenancy/access/`; then `down -v --remove-orphans` | **PASS.** Transport 7/7, lifecycle 6/6, audit 1/1, recovery 1/1, **discovery 5/5**, **ordering 10/10**, footprint 1/1, E-MCP lines 5/5 with 0 trigger lines, policy 38/38, policy ordering 17/17. No SKIP or FAIL. Canary hits in the core records and the PostgreSQL log: 0. Teardown rc 0; no container, volume or network with the project or `g015` label remains (`b2-r1/meta.txt`) |

E-MCP-4 (statements per snapshot, after 30 extra Groups and profiles, identical before):

| Subject | context check | production catalog | capability catalog |
|---|---|---|---|
| owner, Organization-only | 6 | 6 | 10 |
| member, Organization-only | 6 | 6 | 10 |
| member, Group selected | 8 | 8 | 10 |

The core records captured in b2 hold only existing events and fields: `access.decision` (with `mcp.authorize` among
the existing operations), `mcp.request`, `mcp.connection`, `mcp.setup`, `auth.*` and `http.request.completed`.

**Driver pins: static inspection, hosted-only execution.** The changed B-MCP-IT block was checked against the b2
logs. The `-run` regex selects the five mcp tests. `it_count … TestMCPDiscovery` counts the five `--- PASS:
TestMCPDiscovery/` lines (observed 5). `TestMCPOrdering` counts 10 (observed 10). `grep -E '^E-MCP(-3|-4|-TRIGGER)?
'` keeps the two `E-MCP` lines, `E-MCP-3` and `E-MCP-4`, and `grep -cE '^E-MCP(-3|-4)? '` counts 5 (observed: 3
`E-MCP ` + 1 `E-MCP-3 ` + 1 `E-MCP-4 `). The PASS and FAIL texts name the new counts. The harness's `rls_logcheck`
(logcheck validate and canary scan), the full migrations driver, B-MCP-E2E and B-IDENTITY-E2E did not run locally
(Q9). They are hosted-only evidence for the required staging gate.

Not run locally (Q9, declined): the `unit`, `migrations`, `ops` and `all` suites and the control suites. Reused,
not rerun, because their inputs are unchanged: B-RLS-ORDERING, B-OBJ-IT, B-AUDIT-IT and B-MCP-CATALOG (`./internal/migrate/`,
unchanged per Q8).

### Acceptance-criteria status (attempt 1, before S-review)

| AC | Status | Evidence |
|---|---|---|
| AC1 finite catalogs | met locally | `TestToolCatalog` and `TestPackageBoundary` (b1); B-MCP-SOURCE (static); CAT-1 production list (b2); IT-MCP-1 reused and run in b2 |
| AC2 eligibility per the matrix | met locally | `TestToolEligibility` (b1); CAT-1 (b2) |
| AC3 select and clear | met locally | IT-MCP-7/11/12/13 through the new middleware (b2); CAT-2 |
| AC4 external changes | met locally | CAT-2 (b2); OR-4 and IT-REVOKE reused (unchanged inputs); IT-MCP-9 (b2) |
| AC5 hidden = unknown, stale calls | met locally | CAT-3 (b2), `TestErrorMapping` (b1) |
| AC6 ordering | met locally | `TestMCPOrdering` 10/10 incl. O-MCP-7/8 (b2); `TestAccessOrdering` 17/17 (b2) |
| AC7 isolation | met locally | CAT-4, IT-MCP-8/9/13 (b2); E2E-MCP-2/3 hosted-only |
| AC8 notifications | met locally | CAT-5, IT-MCP-1 (b2) |
| AC9 security, privacy, footprint | met locally, S-review pending | E-MCP-1..4 (b2), boundary pins (b1), canary 0, existing events only; S-review required |
| AC10 evidence | in progress | this Result; the hosted MR fast and staging gates and the cleanup receipt come later |

### Deviations from the Plan wording (no design change)

1. **Production spec validation runs at package initialization, not in `New(cfg)`.** `handler.go` is outside the
   allow-list. `productionCatalog = mustCatalog(productionEntries())` refuses an invalid production spec before any
   server is built: the process stops instead of serving. The production spec uses only the three infallible
   constructors, and `TestToolCatalog` asserts its exact contents.
2. **Statement count depends on Group bindings.** `ReadFacts` skips the Group-row and association statements when
   the subject has no bound Group. A capability snapshot for such a subject (for example an Organization admin
   without bindings) adds 2 statements, not 4. The bound "at most four more, constant in the Group and profile count"
   holds, so Plan r4 §1's "role-independent" wording is slightly stronger than the behavior. E-MCP-4 records the
   actual counts.
3. **A liveness failure inside the C-6 snapshot answers the unknown-tool envelope.** This covers a session that ends
   between the binding check and the snapshot. It answers alike for both kinds of name. The transport 401 still
   precedes it in practice, and `tools/list` keeps G014's `-32040`.
4. **Store-down under C-6 writes a failed decision record.** Both names answer `-32041` and write no audit row. Each
   also writes the truthful `failed` `access.decision` record, as any facade outage does. Plan §6.2's "record nothing"
   is read as "no audit record".
5. **Renames.** `registerByRequirement` became `registerControl` plus `registerEligible`, and `TestMCPCatalog`
   (package mcp) became `TestMCPDiscovery` under Q8.

### Unresolved, blockers, attempted approaches, cycles

No blockers. The S-review (read-only security and concurrency, Plan r4 §7) is required before final readiness.
Attempts used: 1 of 7. Correction cycles: 0. Allow-list: all 18 changed paths match. Retained owned resources: none.
The b1 image was removed after b2, and the evidence files under `platform-evidence/g015/` stay until closeout.

## S-review and final (attempt 1, correction cycle 1)

### S-review verdict and fixes

S-review 01 was an independent, read-only security and concurrency review (Opus 5.5, high) of candidate 709f551
(code 8dcfa38) on staging dba74d3. Record: `platform-evidence/g015/s-review-01.md`. Verdict: **PASS WITH FINDINGS,
nothing blocking.** All 18 paths are inside the allow-list. The five deviations were judged safe and honestly
described. The fixes are in commit 1db9693 and change no production code.

| Finding | Fix |
|---|---|
| F1 (low, test gap): no test proved that the five production-tool calls skip the C-6 snapshot (`server.go` `callName`) | `core/internal/mcp/mcp_integration_test.go`, CAT-3. A local `quiet` helper reads `e.mcpDenials(t,u)` and `len(e.authorizeRecords(t))` before and after each call and fails if either changes. It wraps the logout after Organization membership loss, and the explicit clear and the logout after selected-Group loss. The optional `callName` unit test was not added, because the integration assertion covers the real request path. |
| F2 (nit) | `docs/foundation/AUTHORIZATION.md`, Group-capability row: "with every profile and its associations" now reads "with every profile, and the associations of those Groups". |
| F3 (nit) | `contracts/mcp/control-plane.md` C-6: adds "a session that ends during the check answers as C-1" (deviation 3). |

There were no other changes.

### Affected-only reruns

Before each batch: fetch; check the ancestry; run `bash infra/ci/docker-desktop-mounts.sh` (12992 and then 12996 of
100000 mounts, 12%); hold `flock <home>/.local/state/platform-evidence/local-harness.lock`.

Staging moved between the two batches, from dba74d3 to 73a4274 ("Accept hosted object diagnostics and authorize
bounded closeout"; management only: `management/v0/PROGRESS.md` and one G014 review record). The branch was rebased
and pushed with a lease on 4661202. The rebase rewrote the commit SHAs as follows:

- a8f04c1 became 59f03f8;
- 8dcfa38 became a8c15d4;
- 709f551 became 5427786;
- 4661202 became 1db9693.

The `core` tree is identical before and after the rebase (`3ff182c3…`), so the b3 evidence applies to the rebased
head.

| Batch | Command (exact) | Head / base | Outcome |
|---|---|---|---|
| b3 focused | `bash …/g015/b3-discovery.sh b3-r1 g015-core-check:b3`. This is derived from `b2-focus-it.sh` and takes the lock itself. It runs `docker build --target check -t g015-core-check:b3 -f infra/docker/core.Dockerfile core`, then the following with the CMD overridden (`--network none --entrypoint sh`): `gofmt -l internal/mcp/mcp_integration_test.go` and `go vet -tags integration ./internal/mcp/`. On disposable Compose project `g015-focus` (postgres only, same env shape as b2) it runs `go test -tags integration -count=1 -timeout 30m -v -run '^TestMCPDiscovery$' ./internal/mcp/`, then `down -v --remove-orphans` | 4661202 / dba74d3 (core tree = 1db9693) | **PASS.** Build rc 0, image `sha256:05c1c725…a968`. gofmt empty; vet empty. **Discovery 5/5** (CAT-1..5, including CAT-3 with the F1 assertions), with no SKIP or FAIL. Canary hits: 0. Teardown rc 0. Leftovers: none (`b3-r1/meta.txt`). The image was removed afterwards. |
| lint (docs and test changed) | `flock …/local-harness.lock infra/validation/run.sh --base 73a427429ae4f194ae723b14b51b0e3acf41bfeb --allow-file management/v0/results/G015-scope-allowlist.txt --owned-registry …/g015/registry/owned.tsv --report-dir …/g015/static-r3 lint lint-base lint-core` | 1db9693 / 73a4274 | **PASS**, RESULT PASSED (run 04483b). 17 required IDs each PASS once. B-SCOPE: 18 changed paths since 73a4274, all allow-listed. B-SECRET-CANARY PASS. `inventory-leftover.txt` is empty. No run-labelled container, volume, network or work directory remains. |

### Reuse of the earlier passes

The b1 and b2 passes of attempt 1 (`b1-r1/`, `static-r2/` and `b2-r1/`, on code 8dcfa38) are reused with attribution
and were not rerun. Between 8dcfa38 and 1db9693 the only non-management changes are the three files above.

- The only Go change is in the `integration`-tagged `mcp_integration_test.go`, inside `TestMCPDiscovery`/CAT-3.
- These b1 checks keep their 8dcfa38 results:
  - the non-integration and `mcpe2e` vets;
  - the 4+4 unit selectors;
  - gofmt of the other 12 files.
- The integration vet and gofmt of the changed file were rerun in b3.
- These b2 tests keep their results because their files and the production code are unchanged:
  - TestMCPTransport 7/7, Lifecycle 6/6, Audit 1/1 and Recovery 1/1;
  - TestAccessPolicy 38/38 and TestAccessOrdering 17/17;
  - TestMCPOrdering 10/10, TestMCPFootprint 1/1 and E-MCP 5/5.
- The earlier reuse of B-RLS-ORDERING, B-OBJ-IT, B-AUDIT-IT and B-MCP-CATALOG still applies.

### Hosted-only evidence

These did not run locally (Q9) and remain hosted-only evidence for the required staging gate:

- `rls_logcheck`;
- the full migrations driver;
- B-MCP-E2E (E2E-MCP-2/3) and B-IDENTITY-E2E;
- the executed B-MCP-IT driver pins, which were inspected statically against the b2 logs.

The `unit`, `migrations`, `ops`, `all` and control suites stay declined locally under Q9. AC9's S-review condition is
now met. AC10 still awaits the hosted MR fast and staging gates and the cleanup receipt.

### Handoff

- **G015 implementer** (Opus 5.5): wrote 59f03f8 (binding), a8c15d4 (code) and 5427786 (attempt 1 Result). It stopped
  at the 350k context threshold.
- **S-review 01 reviewer** (read-only): wrote `s-review-01.md`.
- **G015 finishing worker** (Opus 5.5; requested effort high, runtime effort unknown; native worktree
  `agent-<id>`, the accepted fallback name): the only writer after the handoff. It wrote 1db9693 (F1–F3),
  `b3-discovery.sh`, the b3 and static-r3 evidence, the rebase onto 73a4274, and this section.

No MR was opened.

### Budget and cycles

- Attempts: 1 of 7 (the S-review fixes are a correction inside attempt 1).
- Correction cycles: 1.
- Blockers: none.
- Retained owned resources: none.
- Evidence under `platform-evidence/g015/` stays until closeout.

## Attempt 2 (Duo correction)

### Binding

This is a Duo CORRECTION LOOP, attempt 2 of B7, under
[the Duo disposition](G015-implementation-codex-review.md) ("Duo disposition and bounded correction release").
Inputs are bound at staging `b219ef4f99c773d5919f6a6fb776e8e05169f855`:

| Input | SHA-256 |
|---|---|
| `management/v0/results/G015-implementation-codex-review.md` | `4d4f1e59a779ac5dcdc13f00b640b08f1d17959a6e741f8f8ec466400a7aca34` |
| `management/v0/specs/G015-dynamic-capability-exposure-spec.md` (with the supervisor's target-existence clarification) | `8c22fbe6c430ed5f900da419165652270010465fa63a22649e1be831b2be0855` |
| `management/v0/results/G015-scope-allowlist.txt` (adds `core/internal/mcp/handler.go` only; previously `eeb84eb3…`) | `f01feafc05b8e24c23156eef7557afc43ba94172849819919d4212d763ea39d4` |
| This Result before this section (the bytes staging holds) | `70e3ac47e3573e521e68bfe483c6d6de29cea29ed20e15ae7a720b70cb4d69f4` |

Plan r4 and the earlier binding sections are kept unchanged.

The branch (MR !33 head 154ba82) was rebased onto b219ef4 without conflict. The delta was management only: PROGRESS,
the review, the allow-list and the Spec. The code commit became 0d5cd80 and the F1–F3 commit became 1df7a76. The
branch had no Result commit to replay, because the d36f2a7 rebuild had dropped them. It was pushed with a lease on
154ba82.

### Changes

| Item | Change |
|---|---|
| D1 wording | The `eligible` doc comment (`policy.go`), the `AuthorizeCatalog` comment (`mcp.go`), AUTHORIZATION.md (a paragraph after the table) and control-plane.md (the Discovery preamble) now describe the Organization owner/admin path as a role-level hint. It asserts no existing Group or BusinessProfile target, so it holds in an otherwise empty Organization, and privileged settings stay owner-only. Group-derived paths need their current candidate, link and policy facts. Eligibility never proves a target exists, grants an effect or replaces the exact-target check. Unknown and foreign targets keep their non-disclosing refusal at call time. The decide calls, queries, fact options and matrix are unchanged; there is no profile-length gate and no new oracle. |
| D1 tests | `TestToolEligibility` has 13 literal empty-Organization cases built from the existing oracle subject facts, with Group bindings, Groups, associations and profiles cleared. Each runs through `eligibilityLoaded`/`factQueryFor`. **Owner:** `group.update`, `profile.update` and `privileged_settings.manage` are eligible. **Admin:** `group.update` and `profile.read` are eligible; `privileged_settings.manage` is not. **Non-admin member:** `group.read`, `profile.read` and `select_for_onboarding` are not eligible. **Context denial:** a selected Group that does not exist (owner, admin) and no membership (`no-bindings`) are not eligible. |
| D2 | The package-level `extraEntries` and its mutex are removed. `Handler` gains a private `fixture atomic.Pointer[catalog]` (`handler.go`), nil in production. `(*Handler).catalog()` does one atomic load and falls back to `productionCatalog`, so a production request takes no lock and shares no fixture state across handlers. `catalog_fixture_test.go`: `fixtureCatalog` builds a new immutable catalog; `useFixtures(t, h)` stores it on `h` only, and its cleanup clears `h` only; `clearFixtures(h)` is per handler. No exported field, setter, flag or environment setting was added, and `Handler` is never copied (it already holds a mutex; vet copylocks is clean). The four `clearFixtures()` call sites in `TestMCPDiscovery` now name their handler (`e.h`, or `down` in the store-down case). |
| D2 test | New `TestCatalogFixtureIsolation`. Four readers load handlers A and B while A's override is stored and cleared 2000 times. B always serves the production catalog. A serves the production catalog or the whole fixture catalog, never a mix. After both handlers get overrides, resetting A leaves B's override in place. |
| Schema nit | `errCatalog`/`catalogSpec` comments, and the attempt 1 Result row for `catalog.go`, now say input-schema shape only: output schemas and closedness are not validated. The `catalogEntry` comment no longer claims closed schemas. |

**Deviation (test pin, in scope).** U-MCP-BOUNDARY (`TestPackageBoundary`, `mcp_test.go`) lists the package's permitted
production imports. `sync/atomic`, which the approved D2 design needs, was not on that list, and b4-r1 caught it.
The pin now allows `sync/atomic` and nothing else (55a6f31, reformatted in 904de79). `mcp_test.go` is in the
allow-list.

**Race detector: not supported locally.** The core check image builds with `CGO_ENABLED=0` and has no `gcc`.
`go test -race` stops with "-race requires cgo", and `CGO_ENABLED=1` stops with "gcc not found" (`b4-r1/test-race.log`,
`b4b-r2/race-support.log`). Supporting it would need a new image or tool install, so it was not done. The isolation
test therefore ran without `-race`. Its synchronization is `atomic.Pointer`, so there are no unsynchronized writes.

### Focused validation

Before each batch: fetch; staging stayed at b219ef4 and is an ancestor of HEAD; `bash infra/ci/docker-desktop-mounts.sh`
reported 13000–13004 of 100000 mounts (13%); the batch ran under
`flock <home>/.local/state/platform-evidence/local-harness.lock`. The check image's default CMD was overridden
throughout (`--entrypoint sh`, `--network none` except on the focus database).

| Batch | Command (exact) | Head / base | Outcome |
|---|---|---|---|
| b4-r1 | `bash …/g015/b4-attempt2.sh b4-r1 g015-core-check:b4`. It builds `--target check`, then runs: `gofmt -l` on the 7 changed Go files; `go vet` (default, `-tags integration` and `-tags mcpe2e`) on `./internal/mcp/` and `./internal/tenancy/access/`; `go test -count=1 -v -run '^(TestToolCatalog\|TestPackageBoundary\|TestCatalogFixtureIsolation)$' ./internal/mcp/`; `go test -count=1 -v -run '^TestToolEligibility$' ./internal/tenancy/access/`; `go test -race … '^TestCatalogFixtureIsolation$'`. On project `g015-focus` it runs `go test -tags integration -count=1 -timeout 30m -v -run '^TestMCPDiscovery$' ./internal/mcp/`, then `down -v --remove-orphans` | a1c1dc0 / b219ef4 | Build, gofmt and all three vets pass. **TestToolEligibility PASS** (with the empty-Organization cases). **TestToolCatalog and TestCatalogFixtureIsolation PASS.** **TestPackageBoundary FAIL**: `handler.go imports sync/atomic` (fixed as above). Race not supported. **Discovery 5/5 PASS** (CAT-1..5, with the per-handler fixture toggles and the store-down handler), no SKIP or FAIL. Canary hits: 0. Teardown rc 0. Leftovers: none |
| b4b-r1 (rerun) | `bash …/g015/b4b-rerun.sh b4b-r1 g015-core-check:b4`: rebuild; `gofmt -l internal/mcp/mcp_test.go`; `go vet ./internal/mcp/ && go vet -tags integration ./internal/mcp/`; the mcp unit selector above | 55a6f31 / b219ef4 | Unit tests 3/3 PASS; vet clean. **gofmt FAIL** (value alignment of the new map line); fixed in 904de79 |
| b4b-r2 (rerun) | the same script, `b4b-r2` | 904de79 / b219ef4 | **PASS.** gofmt empty; vet clean; TestToolCatalog, TestPackageBoundary and TestCatalogFixtureIsolation 3/3 PASS, no SKIP or FAIL. Race support: `CGO_ENABLED=0`, no gcc |
| static-r4 | `flock …/local-harness.lock infra/validation/run.sh --base b219ef4f99c773d5919f6a6fb776e8e05169f855 --allow-file management/v0/results/G015-scope-allowlist.txt --owned-registry …/g015/registry/owned.tsv --report-dir …/g015/static-r4 lint lint-base lint-core` | 904de79 / b219ef4 | **PASS**, RESULT PASSED (run 0d7364). 17 required IDs each PASS once. B-MCP-SOURCE, B-SECRET-SOURCE and B-SECRET-CANARY PASS. B-SCOPE: 18 changed paths since b219ef4, all inside the new allow-list (`handler.go` included). `inventory-leftover.txt` is empty |

The Discovery and eligibility results from b4-r1 still apply at 904de79. Between a1c1dc0 and 904de79 the only change is
the U-MCP-BOUNDARY import map in `mcp_test.go`, and b4b-r2 reran that file's checks. The b4 image was removed. No
container, volume, network or image with a `g015` or run label remains.

**Reused, not rerun.** These keep their earlier attributed results, because the attempt 2 production delta is
comments plus the `catalog()` load, which returns `productionCatalog` whenever the pointer is nil:

- attempt 1 b1, b2 and b3: TestMCPTransport, Lifecycle, Audit and Recovery; TestAccessPolicy and TestAccessOrdering;
  TestMCPOrdering, TestMCPFootprint and E-MCP; the other unit selectors;
- the earlier reuse of B-RLS-ORDERING, B-OBJ-IT, B-AUDIT-IT and B-MCP-CATALOG.

The access package's production changes are comments only. Declined locally under Q9 and the disposition: `unit`,
`migrations`, `ops`, `all`, full controls and benchmarks. Hosted-only, as before: `rls_logcheck`, the full migrations
driver, B-MCP-E2E, B-IDENTITY-E2E and the executed B-MCP-IT pins (the MR fast run and Duo apply after the push; the
staging gate applies after an authorized merge).

### Handoff, budget and cycles

- **Writer:** the G015 finishing worker (Opus 5.5; requested effort high, runtime effort unknown; native worktree
  `agent-<id>`, the accepted fallback name), the only writer.
- **What it wrote:** the rebase onto b219ef4, a1c1dc0 (D1, D2 and the wording), 55a6f31 and 904de79 (boundary pin),
  `b4-attempt2.sh`, `b4b-rerun.sh`, the b4/b4b/static-r4 evidence, and this section.
- **Not done here:** no MR was opened, and the merge stays held.
- **Budget:** attempt 2 of 7.
- **Correction cycles:** the attempt 1 S-review cycle (1), and Duo D1/D2 cycle 1. No issue has repeated.
- **Blockers:** none.
- **Retained owned resources:** none.

## Integration and closeout (2026-10-16; coordinator)

### Integration
- **Draft and ready:** the early draft MR !33 was opened at `c7ce0bb` under Codex's disposition; that opening was
  G016's planning trigger. It was marked ready after the attempt-2 pre-integration ACCEPT.
- **Duo:** both threads (D1 eligibility wording, D2 test seam) were answered with fix evidence and resolved. Duo's
  current-head review found nothing new.
- **Rebuilt head:** `b80c28307c097c72041012a60385e82720734c08`. The five code commits were cherry-picked onto staging
  `a34241a`, and the Result-only commit was dropped because staging already held its bytes. `git diff 2ef5fcd b80c283`
  over core, contracts, docs, infra and web is empty.
- **Fast:** pipeline 2954753697 passed on `b80c283`.
- **Merge:** auto-merge was armed with the exact SHA and source deletion. It merged as
  `7e0ac44c6f85f9de73a8e3c53d209296f9f90a6f`. The merged code equals `b80c283`, and the source branch is deleted.
  The only staging commit between the candidate base and the merge was management-only (`36c7f4e`, the G016 review).

### Staging gate (pipeline 2954782898)
- **Result:** SUCCESS on the first run. Fast job 17229626722 and gate job 17229626723.
- **Rows:** 80 PASS / 0 FAIL / 6 INFO, 79 required IDs, B-CI-CONTROLS 441.
- **Checks recorded as hosted-only in attempt 1, now passing:** B-MCP-IT (including TestMCPDiscovery and the ordering
  and E-MCP pins), B-MCP-E2E and B-IDENTITY-E2E.
- **Leftover inventory:** the baseline `project-archetype-render-cache` only.

### Budget and wording
- **Budget:** 2 of B7 used: attempt 1, then the attempt-2 Duo correction.
- **Race detector:** it stayed unavailable locally (cgo disabled). That's an accepted bounded limitation, not a pass.
- **Wording:** `catalogSpec` validates the input-schema shape only, as attempt 2 recorded.

### Closeout
- **Receipt:** `~/.local/state/platform-evidence/g015-closeout/RECEIPT.md`, SHA-256
  `edb4d484cabd159dd1dff4436b0db6885aae2c65625f9843a8dd7c1b1486d021`.
- **Removed:**
  - the bulk run evidence (about 1.7 MB);
  - the finishing-worker and R0 worktrees;
  - three merged or equal local branches.
  - No G015 Docker resource remains.
- **Held:**
  - The implementer-1 worktree. Non-force removal refused, but it has 0 untracked files and all 1,286 files are
    byte-identical to `709f551`, which is patch-equivalent to the merged code, not an ancestor. It needs an
    authorized force-remove.
  - The retired G015 planner's worktree, held by the agent-tool lock.
  - The remotes `origin/g015-plan` and `origin/g015-r0` (history).
  - The G016 planning resources (owner hold).
