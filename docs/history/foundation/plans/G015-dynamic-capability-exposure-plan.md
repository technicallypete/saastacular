> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G015 Implementation Plan — Dynamic capability exposure and stale-context handling

Revision r4, 2026-10-16 UTC. Author: Claude (Fable 5.1, planning effort high as instructed; the runtime setting is not
verifiable from inside the worker). Status: **R0 reconciliation of the provisionally accepted r3 design
(r3 review (not included): ACCEPT, design only, execution held) against the merged G014
code on `origin/staging`, proposed for Codex review.** Not an execution binding and not implementation. Tier:
architecture/security/public-contract (Goal Execution Policy).

**r4 revision note (R0; narrow).** r3 is archived byte-exact as r3 (not included)
(digest `8e1718b5…4ce9ba`, §2 label A3). Base moves to `origin/staging` `f45ad716569cd791756a9786e976b3d1f1a82b4a`,
which contains the merged G014 (merge `a8d146c`, source head `8c56943`) and the accepted hosted retry
(gate job 17225023104 SUCCESS, pipeline 2953945251; [closeout record](../results/G014-implementation-codex-review.md)).
**R0 finding: nothing material.** The design of r3 (§1, §5.1, §5.3, §5.4, §5.5, §6.1, §6.2 new cases, §7, §8, §10
decisions) is unchanged; only the sections below change. Everything else is r3.

| R0 item | r4 change |
|---|---|
| 1 P-G014-1..13 re-read against the merged code | §0.2 R0 record: every row **confirmed** with file:line at `f45ad71`; immaterial drift recorded (a third inert seam `selectDecided`; `mcp.setup` at `connections.go:96`; `mcpLocks` is a helper function, not an `itEnv` method; `TestCursor` added to the unit set). P-G014-5 is the implemented F3 shape (one Organization-only guarded write locking the previous and target Group facts; the 00010 guard requires `ctx_group` NULL and visible same-Organization previous and target Groups) — exactly the expected shape r3 recorded, so F-3/C-3 need no design change. P-G014-9: 00010 is the last migration; **00011 is the next free number, unused**. P-G014-11: pins re-read; `TestMCPOrdering` is 8/8 at the base with G014's own **O-MCP-5 and O-MCP-6**, so G015's two ordering cases are renumbered **O-MCP-7 and O-MCP-8** and the pin becomes 8→10 (§6.2, §9, §10 E-4, §11.2) |
| 2 Changes since the r3 baseline | the `mcp.request`/`mcp.connection`/`mcp.setup` events, `csrf_token` and the paginated internal setup listing (G014 attempt 2) touch only the setup surface (`connections.go`, `identity/mcp_store.go`, `ListMCPConnections(tc, after)`); **no G015 assumption depends on the setup listing** (§0.3). F-1's "no pagination contract is added" concerns `tools/list` and still holds |
| 3 `unavailable` wording | F-3, C-4, N-3, N-5 and CAT-2 reconciled with the corrected contract (`contracts/mcp/control-plane.md`, switch_group and the `unavailable` row): a definitive refusal or a known abort means nothing changed; `unavailable` after a possible commit means the outcome is unknown and the client rereads before retrying; CAT-2's "unchanged" assertion is scoped to the definitive cases |
| 4 Executable validation selectors | §6.3 names the actual `go test -run` expression, suite or `validate-controls.sh --only` for every planned local run (owner policy `d926012`, focused selection (not included)); the `ops` suite for the MCP browser/E2E evidence is proposed as an explicit exception for Codex's disposition |
| 5 Process | the r3 planner handed off at the 350k active-context threshold (staging `cee09c9`); this revision was written by a bounded successor worker (§2 planner record). The r3 review's "unchanged design may be reconciled without rewriting the approved bytes" is honoured by scope: no design section is rewritten |

**r3 revision note (retained).** r2 is archived byte-identically by Codex as
r2 (not included) (digest `d1046350…e9b9`, §2 label A2). Base moves to staging
`cd54d443d6c2db149128b3c31c0cfc7cbe98ac00`; the delta from `51e4df0` is management only (PROGRESS, the r2 archive and
the r2 review). The G014 candidate is unchanged at `c8f8f9af` (draft, not accepted; F3 open). Only the affected parts
change; everything else is r2.

| Review item | r3 change |
|---|---|
| 1 Independent Organization authority | §1 "Facts per requirement" and "eligibility" follow AUTHORIZATION.md's matrix literally: an Organization owner/admin's `all` path reaches any Group of the Organization without Group bindings (restricted to the selected Group only in selected context, as the context gate says), and their profile paths are independent of the selected Group and include unlinked profiles. Only Group-derived paths use the subject's bindings, links, selection and the flag. Target sets and fact bounds are corrected (§1, §5.5); `decide` and policy are untouched. CAT-1 and `TestToolEligibility` gain the Organization admin without Group bindings and the unlinked profile under valid selected context (§6.1, §6.2) |
| 2 Hidden/unknown equivalence under context refusal | §1 and §5.3 C-1/C-6: every `tools/call` whose name is not a pre-registered production entry takes the same path — one `AuthorizeCatalog` snapshot, whether the name is a capability entry or absent. On an authorization refusal (`contextCheck`: membership or selection) both answer the SDK unknown-tool envelope for the requested name and record the same truthful catalog denial (`access.denied`/`mcp.request` audit row, `access.decision` deny log with `opMCPAuthorize`); a store/verifier failure answers `-32041 unavailable` for both; transport authentication (401) precedes everything; recovery calls and whole-list refusal are unchanged. C-1's "no facade call, no audit event" claim is corrected. CAT-3 is extended with the invalid-context cases; §5.5 adds the honest inventory row for non-registered names |
| G014 F3 design accepted (coordinator notice, staging `2ffff8b`; FYI, not an r2 item) | P-G014-5 and F-3 record the accepted expected shape (one TENANT-ORG `switch_group` transaction locking old and target Group bindings; guard requires `ctx_group` NULL and visible same-Organization old/target Groups; raced old-Group loss → `context_denied`; `TestMCPOrdering` grows). Still provisional until G014 final ACCEPT; R0 re-reads the actual code and counts. The base of this revision stays `cd54d44`; the F3 records are cited by staging commit and path, not hash-bound here |
| Closed by the r2 review | Q2–Q7 and D-2..D-4 are resolved without an owner decision and recorded in §10/§12: names-only discovery within the existing import boundary; the inherited whole-list refusal; the SDK template subject to item 2; production +0 for the five controls, subject to implementation evidence, with no owner performance acceptance inferred; one `TestMCPCatalog` in the existing B-MCP-IT driver; MCP.md disclosure only; exact counts reconciled at R0 (provisional counts never drop inherited tests). D-1 remains resolved |

**r2 revision note (retained).** r1 is archived byte-identically by Codex as
r1 (not included) (digest `ac71980f…b843`, §2 label A1). Base moves to staging
`51e4df03d28c292ac29cf67cac19d040d80fc567`; the delta from `37aa23a` is management only (PROGRESS, the r1 archive, the
r1 review and the G014 request-03 review). The G014 candidate advanced to `c8f8f9af2cc322047ddd21f01e9e60f0cc1a5ae2`
(still draft MR !31, not accepted; F3 in a Codex correction loop).

| Review item | r2 change |
|---|---|
| 1 Recovery calls reachable | §1/§5.1/§5.3: per-call authorization is **requirement-aware**. Binding/Membership/Context tools are registered without a pre-check and keep their own G014 facade checks (`logout` binding-only; `switch_group` clear after selected-Group loss), so `snapshot()`'s unconditional `contextCheck` is never put in front of them. Only capability-mapped entries (none in production) pass an eligibility snapshot. Whole-list refusal is unchanged and hides private metadata. §6.2 routes IT-MCP-11/12/13 (which drive the real HTTP path, hence the new middleware) as reuse evidence for the exceptions |
| 2 Actual facts, canonical names | §1 "Facts per requirement" and §5.5: the candidate's Organization-only snapshot passes a non-nil empty Group set and no profiles; eligibility for Group/profile capabilities therefore pins the exact extra `FactQuery` options (`Groups: nil`, `AllProfiles: true`) and statement counts. Capability names are the eleven canonical values (`business_profile.profile.update`, not `profile.update`); fixtures keep Organization-admin, Group-derived, link and policy cases; R-b is withdrawn |
| 3 No check-then-effect | `AuthorizeTool` is **removed**. `AuthorizeCatalog` returns names only (a discovery hint; no context, facts or token of authority) and is confined to discovery; fixtures have no effect and perform no check of their own. AC5 exact-target evidence uses the existing guarded operation (`switch_group`'s guarded write with the target lock) — §5.3 C-3, §6.2 CAT-3. AUTHORIZATION.md's "no exported evaluator/check authority" is preserved and pinned in §5.1 |
| 4 Query/PG-006 claim | §5.5 inventories real statements: production `tools/list`, `tools/call`, `initialize`, `ping` are **+0**; the capability path (test-only until a future Goal adds such a tool) is one snapshot of the G014 shape plus up to four statements, set-bounded by the subject's Group bindings and the Organization's profiles (the existing manageable-profile footprint). The r1 "PG-006 owner disposition covers this" claim is withdrawn; there is no production residual cost to dispose of, and a future capability-mapped production tool must inventory its own |
| echo vs byte-equality | §5.3 C-1: the SDK envelope echoes the requested name; byte-identity is asserted for the **same requested name** in two states (not in the spec at all; in the spec but ineligible); other names share classification/template with caller-input variation only |
| selection-only `context_version` | §5.4 N-3/N-5: `context_version` is a selection revision, not a catalog revision; external changes do not bump it; clients re-list regardless; a bump does not imply an effective catalog change |
| SDK sorted order | §5.2 F-1: the SDK lists tools in sorted name order; no custom listing machinery |
| strict log fields | §5.5: no count/version fields; the list path reuses `opMCPAuthorize` and the existing `access.decision` fields; no `events.json` change; E-3 withdrawn |
| D-1 resolved | removed from §10/§12; §5.4 records the accepted rule (truthful `listChanged:false`, fresh request-bound discovery, no SSE, no owner decision) |
| G014 drift | §0.2 rows cite `c8f8f9a`: `TestMCPOrdering` has six subtests (O-MCP-2b); `mcp.request`/`mcp.connection`/`mcp.setup` log events exist in `core/internal/mcp`; the internal list response carries `csrf_token` (accepted bounded internal contract; not relied on); P-G014-5 is marked **changing** (F3) |

- Inputs (r4): Goal and Spec on `origin/staging` `f45ad716569cd791756a9786e976b3d1f1a82b4a` (§2); predecessor G014
  **merged** to staging (merge `a8d146c`, source head `8c56943`; implementation ACCEPT with the hosted gate retry
  SUCCESS; closeout in progress per the [G014 review record](../results/G014-implementation-codex-review.md)). Every
  G014 interface used here was re-read at R0 against that merged code (§0.2 R0 record) and is **confirmed**; reuse of
  G014 evidence (IT-MCP-n, O-MCP-n, E2E-MCP-n) rests on that acceptance, not on this Plan. (r3 inputs, retained for
  history: staging `cd54d44`; candidate `c8f8f9af` on base `5bc1bee`, draft MR !31, F3 then open.)
- Tier and budget: architecture/security/public-contract; **B7, checkpoint after attempt 4** (§11).
- Trigger: the standing automatic-planning rule at first verified predecessor implementation MR
  (AGENTS.md "Automatic planning at implementation MR creation"); PROGRESS 2026-10-16 dispositions.

## 0. Gates and predecessor reconciliation

### 0.1 Gates

| Gate | Status at r4 (R0) | Owner |
|---|---|---|
| G014 final ACCEPT with its required staging disposition and scoped cleanup | **implementation ACCEPT and merged** (`a8d146c`); hosted gate retry 17225023104 SUCCESS qualifies the staging gate; **closeout in progress** (Codex record). G015 execution needs the recorded final ACCEPT/closeout | Codex |
| R0: reconcile P-G014-1..13 (§0.2), including the accepted F3 resolution and request-03, against the *accepted* G014 content; revise this Plan if material | **done, nothing material** (§0.2 R0 record, §0.3); this r4 is the bounded record, not a redesign | Claude; reviewed by Codex |
| Codex Plan review of r3 | **ACCEPT, provisional design only; execution held** (r3 review (not included)) | Codex |
| Codex review of this r4 (R0 record, wording reconciliation, executable selectors incl. the `ops` exception request, final allow-list) | **this hand-back** | Codex |
| Codex decisions D-1..D-4 and Q2–Q7 | **closed** (D-1 in the r1 review; D-2..D-4 and Q2–Q7 in the r2 review); no owner decision | — |
| Git-first binding: allow-list (§11.3 final), bound inputs (§2, rebound to `f45ad71`), initial Result (drafted: `management/v0/results/G015-dynamic-capability-exposure-result.md`) | after this review | Codex imports |
| Budget (B7, checkpoint after attempt 4) and resource readiness; recorded execution release | after all above | Codex |
| G016+ and product tools | not authorized; G021 closed | — |

### 0.2 Provisional G014 interfaces (P-G014-n)

Read-only inspection of `origin/g014-tenant-bound-mcp-connections` at `c8f8f9af` (plain `git show`; r1 rows were read
at `d4d2f1cd` and re-checked at `c8f8f9af` where the review named drift), plus the G014 Plan r8 (§3.2, §3.5, §4.2–§4.6,
§10, §11.5, §16) and the request-03 review (not included).

| # | Interface relied on | Where (candidate) | G015 use |
|---|---|---|---|
| P-G014-1 | One throwaway SDK server per verified HTTP request: `(*Handler).newServer(actx, conn)` adds a receiving middleware (method allow-list; `tools/list` → `acc.AuthorizeConnection`), then `tools{...}.register(s)` adds all five tools via `Server.AddTool`. No process-global catalog, no session store, no cache. | `core/internal/mcp/server.go`, `tools.go` | the registration/middleware hook (§1) |
| P-G014-2 | Catalog constants `ToolNames` (five names), `ToolSchemas(name)`, closed schemas, `ContextView{..., ContextVersion int64}` returned by the read/select tools | `core/internal/mcp/tools.go` | production catalog unchanged; `ContextVersion` is a selection revision (§5.4) |
| P-G014-3 | Transport: `NewStreamableHTTPHandler(getServer, {Stateless: true, JSONResponse: true, MaxRequestBodyBytes: 64 KiB})`; GET/DELETE 405; `ServerOptions.Capabilities.Tools.ListChanged = false`; `ProtocolVersions = {"2025-11-25", "2025-06-18"}` (2026-07-28 not claimed) | `core/internal/mcp/handler.go`, `server.go` | fixes the supported notification rule (§5.4; accepted in the r1 review) |
| P-G014-4 | Per-call authorization in the access facade: `snapshot(ctx, op, fn)` = one REPEATABLE READ READ ONLY TENANT-ORG transaction: `CheckLiveSession` → `readSelection` (selected Group, `context_version`) → `tenancy.ReadFacts(FactQuery{Groups: []} or {Groups: [selected]})` → **unconditional** `contextCheck` → `fn`; `AuthorizeConnection(ctx) error`; `GetConnectionContext`, `ListSelectableGroups`, `SelectConnectionGroup(ctx, groupID)` (`""` clears; clear checks membership only; `logout` never enters `snapshot`); `BeginVerifiedRequest(ctx, conn)`; test hooks `mcpHooks{selectLocked, activateLocked}` | `core/internal/tenancy/access/mcp.go:183-222`, `:324-400` | the capability eligibility snapshot is a **variant** of this shape, never in front of the recovery tools (§1) |
| P-G014-5 | **Provisional; accepted F3 design, code not yet implemented at `c8f8f9af`.** Candidate today: decision snapshot, then one READ COMMITTED guarded write (users → sessions → organizations → Organization binding → target Group binding/row → connection row `FOR NO KEY UPDATE`), optimistic guard on `selected_group_id == expected`, `context_version + 1`; `ErrMCPSelectionChanged` → `CodeConflict` → `unavailable`. **Expected shape (Codex-accepted F3 correction design, staging `2ffff8b`: `management/v0/results/G014-f3-correction-codex-review.md`, `management/v0/plans/G014-f3-correction-proposal.md`):** `switch_group` writes in **one TENANT-ORG transaction** locking users → session → organization → Organization binding → the subject's own **old and target** Group bindings (sorted) → Group rows → connection row; the 00010 guard no longer requires selection = `ctx_group` but requires `ctx_group` NULL and visible same-Organization old and target Groups; a raced old-Group loss yields `context_denied` with selection/version unchanged; `TestMCPOrdering` grows. | `access/mcp.go` `selectGroup`; `identity/mcp_store.go` `sqlMCPSelect`; `core/migrations/00010_mcp_connections.sql` guard | reused as the local invalidation point and the exact-target guard (§5.2 F-3, §5.3 C-3). G015 adds no lock, SQL or guard change of its own; the TENANT-ORG write shape does not affect the eligibility snapshot (read-only, no locks). R0 re-reads the actual merged code, guard SQL, lock order and counts |
| P-G014-6 | Policy primitives (staging, unchanged by G014): unexported `Capability` with canonical values `organization.read`, `organization.manage_members`, `organization.manage_policy`, `group.read`, `group.update`, `group.manage_members`, `business_profile.profile.read`, `business_profile.profile.update`, `business_profile.select_for_onboarding`, `business_profile.privileged_settings.manage`, `audit.read`; `decide(cap, ctx, facts, target) (bool, Reason)`, `contextCheck`, `linkedPath`, `manageableProfiles` (unexported, tests only); `tenancy.Facts{Organization, Policy, OrganizationBindings, GroupBindings, Groups, Associations, Profiles}`; `FactQuery{Groups nil = every bound Group; non-nil empty = no Group facts; AllProfiles; TargetGroup; Lock}`; AUTHORIZATION.md: no exported allow result, capability check or evaluator; `TestAccessBoundary` pins the exported set | `core/internal/tenancy/access/policy.go:11-21`, `core/internal/tenancy/facts.go:42-76`, `boundary_test.go` | eligibility derives from `decide` over facts read with the pinned options (§1); exported surface is names-only discovery (§5.1, D-2) |
| P-G014-7 | Error envelopes: tool errors `context_denied`/`group_not_selectable`/`invalid_input`/`unavailable`; JSON-RPC `-32040`/`-32041` for a refused `tools/list`; `methodNotFound()` does not echo; unknown tool = SDK default `jsonrpc.CodeInvalidParams` (-32602), message `unknown tool "<name>"` (echoes the caller's own string); `contracts/mcp/control-plane.md`: "An unknown tool is JSON-RPC invalid params" | `core/internal/mcp/errors.go`, SDK `server.go` `callTool` | the hidden-tool envelope is this unknown-tool envelope by construction (§5.3 C-1) |
| P-G014-8 | Connection record: `identity.MCPConnection` accessors (`ID`, `SessionID`, `UserID`, `OrganizationID`, `Host`, `Resource`, `Active`, `ClientID`, `SelectedGroupID`, `ContextVersion`, `RenewalDue`, `Valid`); `ReadMCPConnection` is the liveness read; `RevokeMCPConnection` (logout, binding only) | `core/internal/identity/mcp.go` | no change |
| P-G014-9 | Migration `00010_mcp_connections.sql`: `app.mcp_connections` with `selected_group_id`, `context_version bigint`, guard trigger (selection/version change only when active and version+1 and equal to `ctx_group`), RLS select/insert/update policies, audit widening to 31 actions, no explicit index beyond pkey/unique. F3 may add SQL/lock changes here. | `core/migrations/00010_mcp_connections.sql` | G015 adds **no** migration; next free number after acceptance is **00011**, unused (§4) |
| P-G014-10 | Fixtures: `setupMCP(t) *mcpEnv` (A: a1,a2; B: b1; `ownerA`, `adminB`, `memberA1`, `cross`, `nonmember`, `dave`; two hosts; real SDK client via `dial`; `state()` reads selection/version); access `itEnv.mcpConn/verified/activeConn/selection/mcpLocks` on the G007 `setupIT` world (`owner`, `admin`, `member`, `gOwner`, `gAdmin`, `gMember`, `multi`, profiles p1 (g1+g2), p2 (g3), p3 unlinked) | `core/internal/mcp/env_integration_test.go`, `access/mcp_integration_test.go` | every G015 integration case extends these (§6) |
| P-G014-11 | Tests and pins at `c8f8f9af`: `TestMCPTransport` (IT-MCP-1..7, 7/7), `TestMCPLifecycle` (IT-MCP-8,9,11,12,13,14, 6/6), `TestMCPAudit`, `TestMCPRecovery`, `TestMCPOrdering` (activation race, O-MCP-1, -2, **-2b**, -3, -4: **6/6**; F3 evidence will extend it), `TestMCPFootprint` (E-MCP-1..3), `TestMCPE2E` (6/6), unit `TestToolCatalog`, `TestErrorMapping`, `TestPackageBoundary` (one `NewStreamableHTTPHandler` call site; SSE symbols forbidden), `mcp.spec.ts` M-*; run.sh pins exact `-run` regexes and counts; required-checks 117 rows / 33 migrations / 14 ops; profiles 35/75/110; VA7 80 | `infra/validation/run.sh`, `required-checks.tsv`, `infra/ci/validate-controls.sh` | counts/regexes change (§6, §11.2); no new check ID; exact numbers re-pinned at R0 |
| P-G014-12 | Audit catalog after G014: 25 trigger + 6 core = 31 actions; `access.denied` operation `mcp.request`; access log operations `mcp.authorize`, `mcp.get_context`, `mcp.list_groups`, `mcp.switch_group` (`events.json:178`); **mcp-package log events `mcp.request`, `mcp.connection`, `mcp.setup`** (`handler.go:405,427`, `connections.go:92`; `events.json:408-466`) | `core/internal/audit/audit.go`, `access/mcp.go`, `core/internal/mcp/{handler,connections}.go`, `events.json` | no new audit action/reason/operation and **no new log operation or field** (§5.5) |
| P-G014-13 | Documents: `docs/foundation/MCP.md`, `contracts/mcp/control-plane.md` (internal list response now documents `csrf_token`, accepted bounded internal contract; not used by G015), AUTHORIZATION/DATABASE-ISOLATION/SECURITY edits; `server.go` comment "G015 owns catalog change notifications" | as named | G015 appends the discovery/stale/notification sections (§3) |

**R0 procedure (r3, retained):** diff `origin/staging` (containing the merged G014) against `c8f8f9af` for
the paths in this table; re-verify each row with file:line, with particular attention to P-G014-5/-9/-11 (the accepted
F3 design, its SQL and the extended ordering fixture) and request-03; record the result here and in the revision
history. Material changes (registration/middleware shape, transport mode, error envelope, `context_version`
semantics, fixture names or check pins) return to Codex with a bounded revision; immaterial drift is recorded only.
Reuse of G014 evidence (IT-MCP-n) requires its final ACCEPT, not this Plan.

**R0 record (r4, 2026-10-16 UTC; `origin/staging` `f45ad716569cd791756a9786e976b3d1f1a82b4a`).** Read-only: `git diff
--stat c8f8f9af origin/staging` over the §0.2 paths (12 files, 714+/63−: the F3 correction `e3f7805`, its review
fixes `82e435a`, the Result commit `f0cf814` and the pagination commit `8c56943`; no change to `facts.go`,
`policy.go`, `audit.go`, `events.json`, `go.mod`, `tools.go`, `server.go`, `handler.go` or `errors.go` after
`c8f8f9af`), then each row re-read at the merged file:line. Every row is **confirmed**; no material change; the
immaterial drift is recorded in the last column. "Material" means the r3 criteria above.

| Row | Finding at `f45ad71` | Where (merged) | Material |
|---|---|---|---|
| P-G014-1 | one throwaway server per verified call; receiving middleware with the method allow-list and `tools/list` → `acc.AuthorizeConnection`; `tools{...}.register(s)` adds the five tools with `Server.AddTool`; no catalog store or cache | `core/internal/mcp/server.go:48-52` (`allowedMethods`), `:56-74` (`newServer`, middleware `:58-70`, `AuthorizeConnection` `:64`, `t.register(s)` `:72`); `tools.go:75-80` | no |
| P-G014-2 | `ToolNames` (five), `ToolSchemas(name)`, `ContextView{…, ContextVersion int64}`, `context_version` in the context output | `tools.go:29`, `:63`, `:150-156`, `:146` | no |
| P-G014-3 | `NewStreamableHTTPHandler` with `Stateless: true`, `JSONResponse: true`, `MaxRequestBodyBytes`; GET/DELETE 405; `ListChanged: false`; `ProtocolVersions = {2025-11-25, 2025-06-18}` | `handler.go:95-104`, `:146`, `:227`; `server.go:41`, `:26` | no |
| P-G014-4 | `snapshot` = one REPEATABLE READ READ ONLY Organization-only transaction: `CheckLiveSession` → `readSelection` → `ReadFacts` → unconditional `contextCheck` → `fn`; `AuthorizeConnection`, `GetConnectionContext`, `ListSelectableGroups`, `SelectConnectionGroup`, `BeginVerifiedRequest`; inert seams | `access/mcp.go:179-218` (`ReadFacts` `:198`, `contextCheck` `:203`), `:220`, `:226`, `:250`, `:303`, `:81`; `mcpHooks{selectDecided, selectLocked, activateLocked}` `:33-37` | no (drift: a third inert seam, `selectDecided`, runs after the switch's decision snapshot and before its guarded write; nil in production; usable by O-MCP-7/8) |
| P-G014-5 | **implemented F3 = the expected shape.** Decision snapshot (Organization-only, no locks; no-op detection; refusals recorded) `:335-380`; `selectDecided` `:381-383`; the guarded write is **one Organization-only transaction for select and clear alike**: `LockLiveSession` (user, session) `:388` → `ReadFacts{Groups: [target, current] distinct, Lock: true}` (Organization, Organization binding, the subject's bindings in both Groups, those Group rows) `:390-399` → decision on the locked facts: membership `:406`, the current selection still valid (Spec R5; `context_denied`, recorded) `:411-414`, the target selectable (`group_not_selectable`) `:416-422` → `selectLocked` `:425-427` → `SelectMCPGroupTx(expected = snapshot selection)` locks the connection row and writes nothing on any change (`ErrMCPSelectionChanged` → `CodeConflict`/`write_conflict` → `unavailable`, `:151-157`). The 00010 guard requires `g IS NULL` (`ctx_group` NULL), `context_version + 1`, active, a target visible in the same Organization and a previous selection (if any) likewise; a clear needs nothing more | `access/mcp.go:324-437`; `identity/mcp_store.go:38-39` (`sqlMCPSelect`), `:216-243` (`SelectMCPGroupTx`); `core/migrations/00010_mcp_connections.sql:357-370`; `docs/foundation/AUTHORIZATION.md:170-187` | no — G015 adds no lock, SQL or guard change; F-3/C-3 already cite this shape; the eligibility snapshot stays read-only |
| P-G014-6 | eleven canonical `Capability` values; `decide`/`contextCheck`/`linkedPath`/`manageableProfiles` unexported; `Facts`/`FactQuery{Groups, TargetGroup, AllProfiles, Lock}` as recorded; files unchanged since the r3 base | `access/policy.go:7-21`; `tenancy/facts.go:42-65`; `boundary_test.go:56-61` | no |
| P-G014-7 | tool codes `context_denied`/`group_not_selectable`/`invalid_input`/`unavailable`; `-32040`/`-32041`; `methodNotFound()` does not echo; unknown tool = SDK default (contract: "An unknown tool is JSON-RPC invalid params") | `errors.go:12-15`, `:28-29`, `:55`; `contracts/mcp/control-plane.md:85-86` | no (the `unavailable` row text changed; reconciled in §5, item 3) |
| P-G014-8 | `MCPConnection` accessors `ID`…`Valid`; `ReadMCPConnection`; `RevokeMCPConnection`; new setup-listing types `MCPListPage`, `MCPListPosition`, `MCPConnectionPage` | `identity/mcp.go:48-83`, `:159`, `:187`; `:97-112` | no (the new types serve the setup listing only, §0.3) |
| P-G014-9 | `00010_mcp_connections.sql` with `selected_group_id`, `context_version`, the guard (selection branch as in P-G014-5), audit widening to 31 actions, the two capture triggers; **the last migration file is 00010; the next free number is 00011, unused by G015** | `00010:316`, `:334-375`, `:404`, `:529-548`, `:562`; `core/migrations/` (10 SQL files) | no |
| P-G014-10 | `setupMCP(t) *mcpEnv`: `ownerA`, `adminB`, `memberA1`, `cross`, `nonmember`, `dave`; Groups a1, a2, b1; `grantOrg`, `grantGroup`, `dial`; access `itEnv.mcpConn/verified/activeConn/selection` | `mcp/env_integration_test.go:74`, `:136`, `:183-197`, `:264`, `:271`, `:494`; `access/mcp_integration_test.go:26`, `:48`, `:61`, `:72` | no (drift: `mcpLocks(t, e)` is a package helper, `:406-408`, not an `itEnv` method) |
| P-G014-11 | B-MCP-IT: `-run '^(TestMCPTransport\|TestMCPLifecycle\|TestMCPAudit\|TestMCPRecovery)$'` in `./internal/mcp/` with 7/6/1/1, and `-run '^(TestMCPOrdering\|TestMCPFootprint)$'` in `./internal/tenancy/access/` with **8**/1 and four `E-MCP` lines; **G014's own `O-MCP-5` (a removal from the selected Group after the decision refuses a switch away; a clear succeeds) and `O-MCP-6` (such a removal waits for a held switch)**; B-MCP-CATALOG: `-run '^(TestMCPCatalog\|TestMCPMigration)$'` in `./internal/migrate/` with **3**/5 (CT-MCP-1, CT-MCP-NEG, **CT-MCP-GUARD**); B-MCP-E2E `^TestMCPE2E$` 6; unit `TestToolCatalog`, `TestErrorMapping`, `TestPackageBoundary` and (new) `TestCursor`; required-checks 117 rows / 33 migrations / 14 ops; profiles fast 35 / integration 75 / full 110 (`required-checks.tsv` header; VU1 full-pipeline union 111 distinct); VA7 80 | `run.sh:3137-3140`, `:3148-3163`, `:4746-4764`; `mcp_test.go:335`, `:388`, `:525`, `:552`; `access/mcp_integration_test.go:87-428`; `required-checks.tsv:2-7`; `validate-controls.sh:1206-1207` (VU1/VU9), `:1107-1115` (VA7) | no — the re-pins anticipated by r3: `TestMCPOrdering` 8 → **10** after G015; G015's cases are **O-MCP-7/O-MCP-8** (previously named O-MCP-5/6 in r3 §6.2/§9/§10/§11.2; names only, content unchanged); the migrate package's `TestMCPCatalog` (B-MCP-CATALOG) is a different package and log from G015's `TestMCPCatalog` in package `mcp` (B-MCP-IT) — see §12 |
| P-G014-12 | 25 trigger + 6 core = 31 actions; `access.denied` operation `mcp.request`; access log operations `mcp.connection.create/list/activate`, `mcp.authorize`, `mcp.get_context`, `mcp.list_groups`, `mcp.switch_group`; mcp-package events `mcp.request`, `mcp.connection`, `mcp.setup` | `audit/audit.go:22`, `:48-49`, `:60`, `:69`; `access/mcp.go:43-51`; `mcp/handler.go:405`, `:427`; `mcp/connections.go:96`; `infra/validation/tools/internal/logrecord/events.json:408`, `:436`, `:466` | no (drift: `mcp.setup` moved from `:92` to `:96`) |
| P-G014-13 | `docs/foundation/MCP.md`; `control-plane.md` with the corrected `switch_group`/`unavailable` wording and the paginated internal listing; AUTHORIZATION.md's MCP table and F3 lock-order paragraph; `server.go` "G015 owns catalog change notifications" | `MCP.md:8-82`; `control-plane.md:88-94`, `:106`, `:116`, `:129-133`; `AUTHORIZATION.md:170-187`; `server.go:35` | no (G015 appends; §3 unchanged) |

### 0.3 Changes since the r3 baseline (item 2)

| Change (G014 after `c8f8f9af`) | Where | Effect on G015 |
|---|---|---|
| Operational events `mcp.request`, `mcp.connection`, `mcp.setup` | `handler.go:405,427`; `connections.go:96`; `events.json:408-466` | none: §5.5 adds no event, operation or field; the C-6 record uses the existing `access.decision`/`access.denied` records with `opMCPAuthorize`/`mcp.request` |
| `csrf_token` on the internal list response | `connections.go` (`writeJSON … "csrf_token"`); `control-plane.md:116` | none: internal setup contract; not read by G015 |
| Paginated internal setup listing (attempt 2): newest-first keyset on `(created_at, id)`, opaque untrusted cursor (`s1`, ≤128 bytes, user/Organization/origin context tag), `next_cursor`, `MCPListPage = 100` plus one look-ahead row, the BFF list-success bound 1,277,222 bytes (`web/lib/mcp-client.ts`, pinned by `mcp_test.go:488-492`) | `identity/mcp_store.go:40-49`, `:246-289`; `identity/mcp.go:97-112`; `access/mcp.go:529-559` (`ListMCPConnections(tc, after)`); `connections.go` (cursor codec, `List`); `access/mcp_test.go:36-39` (stub signature) | **none**: G015 reads no setup listing. §3's "unchanged" list already names `core/internal/identity/*`, web and `infra/ci/*`; `connections.go` is outside the §11.3 allow-list; the `stubMCP` signature in `access/mcp_test.go` is untouched by G015's pin additions. F-1 ("no pagination contract is added") is about `tools/list` (SDK `DefaultPageSize` 1000, `nextCursor` absent) and still holds; `ListSelectableGroups` paging is G014's and unchanged |
| `TestMCPTransport` IT-MCP-3 extended for pagination; unit `TestCursor` added | `mcp_integration_test.go`; `mcp_test.go:388` | none: counts stay 7 and the unit set grows by a G014 test G015 does not touch |

## 1. Approach

G014 already computes everything per request with nothing cached: one SDK server per verified call, one read-only
snapshot per facade operation, `tools/list` allowed or refused as a whole, and each tool's handler performing its own
exact check (binding-only `logout`, membership-only clear, membership plus valid selection for the read tools, the
guarded write for select). G015 keeps every one of those checks where it is and adds **requirement-aware registration**:
a tool is registered on the request's throwaway server only when its declared requirement can hold, and for
capability-mapped tools that decision is a discovery-only eligibility snapshot.

```text
POST /mcp/c/{uuid}  ──(G014 steps 1–15: path/host, bearer, binding, BeginVerifiedRequest, activation/renewal)──▶
  newServer(actx, conn)
    register every Binding/Membership/Context entry  (the five production tools; no statement; their handlers
                                                      keep their own G014 checks: logout, clear, read tools, select)
    receiving middleware:
      initialize / notifications/initialized / notifications/cancelled / ping → unchanged (no facts read)
      tools/list → acc.AuthorizeCatalog(actx, capabilityEntries)     ← one snapshot; whole-list refusal unchanged
                   register the eligible capability entries; next → SDK listTools (sorted names)
      tools/call → name is a pre-registered production entry: next → its handler (own G014 check; no snapshot)
                   any other name (capability entry or absent — same path):
                     acc.AuthorizeCatalog(actx, capabilityEntries)                    ← one snapshot, same for both
                       context refusal → SDK unknown-tool envelope for the requested name + truthful denial record
                       store/verifier failure → -32041 unavailable
                       allowed → register the eligible entries; next → SDK callTool
                                 (unregistered name = unknown tool; registered entry runs the fixture)
```

- **Catalog spec (closed, compile-time).** Each entry carries one `access.ToolRequirement`:
  - `RequireBinding()` — valid exact binding only: `logout` (G014 termination exception; no membership);
  - `RequireMembership()` — current Organization membership: `context.switch_group` (explicit clear stays reachable
    after selected-Group loss; select re-checks current and target Group inside the guarded write — P-G014-5);
  - `RequireContext()` — membership and a valid current selection: `account.get_current_user`,
    `account.get_current_context`, `context.list_groups`;
  - `RequireCapability(name)` — one of the eleven canonical capability values (P-G014-6), decided with the existing
    `decide` semantics in the current context. **Production uses none**; the test-only fixtures do (§6.1). An unknown
    name fails closed at construction (`New` refuses; the fixture test asserts the refusal).
- **Requirement-aware per-call authorization (review item 1).** Binding/Membership/Context entries are registered for
  every request; no check runs before their handlers, so the G014 recovery and termination exceptions are untouched
  and keep their exact non-disclosing refusals. Capability entries are registered only when `AuthorizeCatalog` says
  eligible. The middleware is therefore on the path of every existing recovery test (§6.2).
- **Facts per requirement (review item 2).** Eligibility is pure over `Facts`; what is read is pinned per requirement:
  - Binding: G014's binding read only (no `ReadFacts`).
  - Membership, Context: G014's existing `ReadFacts{Groups: [] | [selected]}` (2 statements Organization-only; 3–4 with
    a selection: organization+policy row, Organization binding, [Group bindings of the selected Group, Group row]).
  - Capability, Organization kind (`organization.*`, `audit.read`): the Context read suffices; `decide` with the
    Organization target.
  - Capability, Group and profile kinds: when any such entry is in the catalog, the eligibility snapshot reads
    `ReadFacts{Groups: nil, AllProfiles: true}` in Organization-only context (+4 statements: all Group bindings of the
    subject, those Group rows, their associations, every profile of the Organization) or
    `{Groups: [selected], AllProfiles: true}` in selected context (+2: associations, profiles). The sets are bounded by
    the subject's Group bindings and the Organization's profiles — the existing manageable-profile footprint
    (AUTHORIZATION.md pin (d)). The reads are fixed per catalog shape, not per role, so the statement count is
    role-independent.
  **Eligibility (review item 1; AUTHORIZATION.md "Capabilities and the matrix", unchanged `decide`):**
  - *Organization owner/admin paths.* `group.*` with the `all` cell: eligible from the Organization binding alone —
    any Group of the Organization, **without** Group bindings; in selected context the context gate restricts the
    target to the selected Group, which is still an eligible target. `business_profile.*` with the `Y` cell:
    eligible from the Organization binding alone, **independent of the selected Group and including unlinked
    profiles**. No Group or profile row is inspected for these paths; the exact target is decided at call time.
  - *Group-derived paths* (`own`, `linked`, `linked+flag`): eligible when some candidate Group of the subject holds the
    required Group role — the candidates are the selected Group in selected context, otherwise every Group of the
    subject (`Groups: nil`) — and, for `linked`, that Group's current association is a profile; `linked+flag` also
    needs the Organization flag. `eligible()` reuses `decide` over the loaded candidates and profiles.
  - *Both paths combine* as independent paths (pin (a)): a `group_admin` who is also `org_admin` is eligible through
    the Organization path regardless of bindings.
  Eligibility means "some in-context target is allowed"; it is a discovery hint only.
- **Whole-list refusal is kept.** When `contextCheck` fails (membership lost, selected Group invalid), `tools/list`
  stays `-32040 context_denied` with no catalog (G014 matrix, IT-MCP-11); the recovery tools remain callable directly
  under their own handler rules. No silent clear or fallback.
- **Hidden equals unknown by construction, including under context refusal (review item 2).** Every `tools/call`
  whose name is not a pre-registered production entry takes one path: the same `AuthorizeCatalog` snapshot whether the
  name is a capability entry or absent. If the context is valid, the eligible entries are registered and the SDK's
  own unknown-tool path answers for any unregistered name before any handler. If `contextCheck` refuses (membership
  or selection), the middleware answers the SDK unknown-tool envelope for the requested name itself and the snapshot
  records the same truthful catalog denial for both names; a real store/verifier failure answers `-32041
  unavailable` for both. The five production entries never enter this path, so the clear/logout exceptions and the
  read tools' own `context_denied` are untouched.
- **Freshness is request-bound.** Every request is a new snapshot; a successful select/clear commits
  `context_version+1` before its response, so the next request on that connection sees the new row and facts.
  Externally committed membership/role/link/policy/session/user changes are visible to the next snapshot without a
  version change. Nothing persists between requests except G014's non-authoritative `selected_group_id` and
  `context_version`.
- **Notifications (accepted rule).** The inherited transport (P-G014-3) has no delivery path for server-initiated
  notifications; `tools.listChanged` stays `false`, truthfully, and discovery is fresh on every request (§5.4).

**Rejected alternatives (recorded):** a stored catalog digest or `catalog_version` column (migration 00011, RLS/guard
changes, serves nothing without a delivery path); a Redis or in-process authority cache (forbidden); a pre-check
snapshot in front of every `tools/call` (r1 — blocked the clear/logout exceptions and doubled production reads); an
exported `AuthorizeTool` check (r1 — check-then-effect authority, removed); one request-scoped transaction held across
SDK dispatch (holds a snapshot across the 10 s deadline, refactors every G014 handler); a custom `tools/list` handler
or per-session filter (the SDK has none: tools are server-global, `listTools` paginates `s.tools` in sorted order);
SSE responses for in-request notifications (resolved in the r1 review: not needed); a new `hidden_tool` envelope
(would distinguish hidden from unknown).

## 2. Source binding and extra inputs

Committed inputs bind Git-first by commit and path on `origin/staging` `f45ad716569cd791756a9786e976b3d1f1a82b4a`
(r4; the r3 base was `cd54d443d6c2db149128b3c31c0cfc7cbe98ac00`). Hash rows carry neutral labels (scanner rule); the
paths are listed below the table. Rows marked r4 are new or changed since r3 (the `cd54d44..f45ad71` delta: the merged
G014 and its closeout records, the r3 archive/review, the 2026-10-16 owner governance notes in AGENTS/CLAUDE, the
focused-selection policy `d926012` in VI, the G014 rows in RC/PP and the G014 edits to AR/SE/FA/FD); every other row
is byte-identical to r3. The merged G014 binds by this commit and the §0.2 paths (the candidate binding `c8f8f9af`
is retired); the new code rows M10/AM/IS/FG/PO pin the files P-G014-5/-6 rely on.

| Label | SHA-256 at the base |
|---|---|
| G | `ca32a84d3a0144feb04446293fab8376720a97e7b56c2f87d007f9e4b7a7dbc0` |
| S | `1f7149ff2205549c2d7a826f2079efeee8d6e9bbf1425bc31828d305457016b4` |
| A3 (r4) | `8e1718b5a04e67ea8bdfa6a0aba40610745821e5d478e0e58fc89d831f4ce9ba` |
| R17 (r4) | `456361c86444721a42fbaaf5fc46160942749b4d8c9ae972c685be67fdc5e13d` |
| PG (r4) | `66a6de306f811699e63e0cda6732c8dca0cb1ba286d6704dbc1b00d2755fd5ae` |
| X14 (r4) | `a40111023744d13b738b4bc0230e31f053fc7012406dcecc9e6adffe4f47c225` |
| I14 (r4) | `97ce874d3d344d5d0f15a592db1a805bbe1d30a4e77a39e42dc385379173ca46` |
| F3R (r4) | `9b0b374b6456911d5542b37e95618b0e98df0fbff281ee95da9c2c80c342ebb5` |
| F3P (r4) | `6ec8784d3964f34efaa29607f0b20707a671901fa84782131d1a814ae28451cc` |
| PGR (r4) | `c66fdb0b96b45d60f6c8df02d48a5283a7e09217521a1c56b2dcd31d29519f15` |
| CP (r4) | `5e77d3bbc9c946b853a35edb6b31ff7ff48cd632146d7305b070eae0d60238da` |
| MC (r4) | `9da2726a915e08852858f8e9ffdc90815a269a171af3a28a294bd15d645c5919` |
| M10 (r4) | `3ac29b6bcf2078158cf7fbaac13b244ce2610450d2e6f0b53fba734b14a73e19` |
| AM (r4) | `a40b1b4bc8d468b5b573f6df3e71079ecffc98fb9e50317bef0c69304b796cfd` |
| IS (r4) | `a78b7c1030ee85201ada6d8bb2dc13ca8480a94ecfdad86f6e2d2e5186623082` |
| FG (r4) | `06895d17b57c71a31c2251df6d8cab996972e238c9d1725339b6db9c21f80557` |
| PO (r4) | `8f563f557e40fa74eb35911dd48077acdc779558227ce266f038a625ac8d99a3` |
| AR (r4) | `8cc90850b00285d9540cac8b3347eb22d05e6413eff839b4195efc9a4df004b1` |
| SE (r4) | `370591c046e68183a2c049c579664352d7c6c4619594b596c418571a8f33ea32` |
| FA (r4) | `1d6159c4f8c4ac549efcaafb09089dd487ef612eb97118e6ce498e74de171b43` |
| FD (r4) | `408ae7d2509ec88758de03328873b1e64dc55f9f05e0624f5ae51110af558dc1` |
| AG (r4) | `5d5d818519bf16c0ecd2e6eef67426cc0e0fc25b36ccc7ec6ec1d93c9c90d468` |
| CL (r4) | `629bb40acbcc7f04d9f6d4790dd272aa47dcf2f4be6e6794822e223e332629ec` |
| VI (r4) | `696df897dcea6eb599a025adb8a22363d94845e4a411af1173292fa285b192d5` |
| RC (r4) | `2d136fea2fa096def63d29a5182fb563b6419f94babf519c8e61ed3455a1c686` |
| RS (r4) | `2718f944dd2325b7cff7deb7c673123b7c076c87a4434eba587ef303452e9d7f` |
| PP (r4) | `d78868bd9b396a0e9640d9418802f3ec82323ccff94f1554302206f9e751ef5b` |
| MS (r4) | `4b36b43a3937e3a08096ed3482dbe79928c82f4ae4a56e809c12545ce9047154` |
| A2 (r3) | `d1046350e5f57927930c1cf1be833e288aff3107d7f972695e05ec6065ede9b9` |
| R16 (r3) | `2cb32f90ab6412726470789bb38449390b9df6ccee76fa8e7c5a35b36c7a19a2` |
| A1 (r2) | `ac71980f134a54b52b6a232bf29bbc16f4bd3cd103c13572257f70f83b40b843` |
| R15 (r2) | `58d16beadd39ddf2f9dcd67b491de2d031a84e96a69885106c68c9fe811d29ce` |
| Q3 (r2) | `861cabc35a7f9008b8f892852bfffcbe790c0baea0cf0a7bac55769bd1331a2b` |
| G14 | `e35a9e5dbf503c22b81a07653c17841d354818159fab6a05947268c4888fc255` |
| S14 | `c5bc60c7c4bfc3ae92a0e63711e41a57eea6d375b6b78139bf2599eb68021a44` |
| P14 | `e53728103993c6c545abbf7c778564d4194bff2b7c4813002551be5afc0412dc` |
| R14 | `a2d0dd4435d584a484c50965e30224a0356202bb84b11914f776e0277f66b9db` |
| L14 | `fbdbe7b26715afa900c80a270a1a9ae823c19af4821d7f4c3271fd385931d022` |
| A14a | `72cf8ac75e414a41b462dbc81e1d7f5173afe0697f8e39d705f87f762fddcbc4` |
| A14b | `a0619f4a534933d8ef1e4d5312fcdf12695e1ae8a2df68c48d446f85030b71a1` |
| D5 | `db31ffd657d168796773bd21c8257085e24daca0d4947200facdf9ffac573b79` |
| D3 | `537522dca62e57cd291ffdad581ea000a1b04c45b378d5f505af061fdb952a62` |
| D4 | `51a487844d74e913d195c963095c2a381deeb293d1829085f2f026f060955a68` |
| D6 | `dc7eda774dc06c8ca1eafc52e723772d4e4c8f32d3141187759b9c47cadabc2a` |
| DD | `d9a90800ff43b16d53a3bef8e2db9027ae14ee9a42a494354d849d00d1235f5e` |
| CC | `23edbe6404694fa15866929775d224f9456a43f6d6f6df0e070faa0228ba2140` |
| PF | `53d6b7f208f3530d797179dee92a613c9d6e243b079a8b224a95c60b0f95e180` |
| RM | `b8ee52d9201eeb8abcf897646ed69939b60e60c2643eb685cf14f74cabc53995` |
| BP | `0115fb585cbde0ffaa067507c477ec7d5dcd28c6fd37c2d47c700c9dbd657cd0` |
| GW | `9b2909825db7316156f209c8f1c7bd2a103a4093c34b007d373ce26b86fa3655` |
| TP | `ac60d5907ccfc1609c38f05c6adc62426032515fb738718fd43c0285691dfa6a` |
| GM | `9575950e0dcefa95dd0793785c5c24fd2802c1a29f1b1cd0d7e380ef822aed14` |

**Paths.** G `management/v0/goals/G015-dynamic-capability-exposure.md`; S
`management/v0/specs/G015-dynamic-capability-exposure-spec.md`; A1/A2/A3
`management/v0/plans/G015-dynamic-capability-exposure-plan-r{1,2,3}.md` (archives, byte-exact); R15/R16/R17
`management/v0/results/G015-plan-r{1,2,3}-codex-review.md`; Q3
`management/v0/results/G014-request-03-codex-review.md`; PG `management/v0/PROGRESS.md`; G14/S14/P14 the G014 Goal,
Spec and Plan (r8 file) under `management/v0/{goals,specs,plans}/G014-tenant-bound-mcp-connections*.md`; R14
`management/v0/results/G014-plan-r8-codex-review.md`; L14 `management/v0/results/G014-scope-allowlist.txt`; A14a/A14b
`management/v0/results/G014-scope-amendment-0{1,2}-codex-review.md`; X14
`management/v0/results/G014-tenant-bound-mcp-connections-result.md`; I14
`management/v0/results/G014-implementation-codex-review.md`; F3R
`management/v0/results/G014-f3-correction-codex-review.md`; F3P `management/v0/plans/G014-f3-correction-proposal.md`;
PGR `management/v0/results/G014-pagination-proposal-codex-review.md`; CP `contracts/mcp/control-plane.md`; MC
`docs/foundation/MCP.md`; M10 `core/migrations/00010_mcp_connections.sql`; AM `core/internal/tenancy/access/mcp.go`;
IS `core/internal/identity/mcp_store.go`; FG `core/internal/tenancy/facts.go`; PO
`core/internal/tenancy/access/policy.go`; D5/D3/D4/D6 `docs/adr/ADR-00{5,3,4,6}-*.md`; DD
`docs/DECISIONS.md`; CC `contracts/mcp/connection-context.md`; AR `docs/ARCHITECTURE.md`; SE `docs/SECURITY.md`; PF
`docs/PERFORMANCE.md`; FA `docs/foundation/AUTHORIZATION.md`; FD `docs/foundation/DATABASE-ISOLATION.md`; RM
`management/v0/ROADMAP.md`; BP `management/BUDGET_POLICY.md`; AG `AGENTS.md`; CL `CLAUDE.md`; GW
`docs/agents/GITLAB-BRANCH-WORKFLOW.md`; VI `skills/sdlc-management/references/validation-integration.md`; TP
`skills/sdlc-management/templates/management/release/plans/PLAN.md.tmpl`; RC `infra/validation/required-checks.tsv`;
RS `infra/validation/run.sh`; PP `infra/ci/profile-paths.tsv`; MS `core/internal/mcp/server.go` (the merged G014
server); GM `core/go.mod`.

**Non-Git inputs.** Go SDK `github.com/modelcontextprotocol/go-sdk v1.8.0`, bound by the `core/go.sum` module lines
(GM); behaviour pinned from its `mcp/server.go`, `mcp/streamable.go` and `mcp/shared.go` at tag v1.8.0 (read 2026-10-16,
§5.4). MCP tools specification 2025-06-18 and 2025-11-25 (`modelcontextprotocol.io/specification/<rev>/server/tools`),
read 2026-10-16. No provider, build, test, install, Docker or harness input was used; no probe was run.

**Planner record.** Task G015, role plan; native isolated worktree
`.claude/worktrees/agent-<id>` (generated `agent-<id>` name, the accepted AGENTS.md fallback); branch
`g015-plan`, rebased onto `origin/staging` `51e4df0` for r2 and `cd54d44` for r3; session
`<agent-session-url>`. Method: read-only file/Git inspection; r1 used two
read-only Explore delegates (candidate inventory; existing-suite inventory); r2 fetched nothing new beyond the
staging delta and the `c8f8f9af` drift the review named; r3 re-read only AUTHORIZATION.md's matrix section and the
staging delta. **r4 (R0):** the r3 planner handed off at the 350k active-context threshold (staging `cee09c9`,
AGENTS.md owner default 2026-10-16); R0 was executed by a bounded successor worker (task G015, role plan/R0; Fable
5.1, effort high as instructed, runtime setting not verifiable) in native isolated worktree
`.claude/worktrees/agent-<id>` (generated name, the accepted fallback), branch `g015-r0` from
`origin/staging` `f45ad71`, same session. Method: read-only `git diff`/`git log` over the §0.2 paths and targeted
file reads; no build, test, install, Docker, probe, provider call or settings change; no delegate.

## 3. Files

| Path | Change |
|---|---|
| `core/internal/mcp/catalog.go` | **new**: the closed catalog spec (`catalogEntry{name, requirement, handler}`), production entries for the five controls, `catalogSpec` validation (unknown requirement refused), requirement-aware registration helper, the unexported test-fixture seam (`extraEntries`, nil in production) |
| `core/internal/mcp/server.go`, `tools.go` | middleware: `tools/list` calls `AuthorizeCatalog` for the capability entries (or `AuthorizeConnection` when there are none — identical cost); `tools/call` for any name that is not a pre-registered production entry calls `AuthorizeCatalog` and maps a context refusal to the SDK unknown-tool envelope for the requested name (§5.3 C-6); `tools.register` becomes `registerByRequirement`; `ServerOptions` unchanged (`ListChanged: false`, versions) |
| `core/internal/mcp/errors.go` | no code change expected; the unit test pins the unknown/hidden envelope against the SDK default |
| `core/internal/tenancy/access/mcp.go` | **exported** `ToolRequirement` (opaque), constructors `RequireBinding`, `RequireMembership`, `RequireContext`, `RequireCapability(string) (ToolRequirement, error)`, `AuthorizeCatalog(ctx, reqs map[string]ToolRequirement) ([]string, error)` (eligible names only, logged as `opMCPAuthorize`); test hook `mcpHooks.catalogComputed` |
| `core/internal/tenancy/access/policy.go` | `eligible(req, sel, facts) bool` built on `contextCheck`/`decide`/`linkedPath`; `factQueryFor(reqs, sel) tenancy.FactQuery` (the pinned options of §1) |
| `core/internal/tenancy/access/boundary_test.go`, `mcp_test.go`, `matrix_test.go` (or `policy_test.go`), `mcp_integration_test.go` | pins and tests (§6) |
| `core/internal/mcp/mcp_test.go`, `mcp_integration_test.go`, `env_integration_test.go` | tests and fixture additions (§6); `catalog_fixture_test.go` **new** (test-only capability-mapped tools, `package mcp`, `_test.go`) |
| `infra/validation/run.sh` | B-MCP-IT regex/count pins (`TestMCPCatalog`; `TestMCPOrdering` 6→8 provisional); unit count text if pinned; no new check ID |
| `docs/foundation/MCP.md`, `docs/foundation/AUTHORIZATION.md`, `contracts/mcp/control-plane.md` | discovery/eligibility ladder, hidden-equals-unknown, freshness and notification rule (§5), the names-only exported surface; G015 evidence rows; the concise disclosure that cached catalogs are hints |

**Unchanged:** every migration (§4), `core/internal/audit/*`, `core/internal/identity/*`, `core/cmd/server/main.go`,
`core/go.mod`/`go.sum`, web (`app/`, `lib/`, browser specs, Playwright config), `infra/compose/*`, `infra/identity/*`,
`infra/ci/*`, `required-checks.tsv`, `events.json` and `logrecord_test.go` (no new log operation or field), `skills/*`
(guidance edits are conditional, §11.3), `docs/SECURITY.md` (the disclosure lives in MCP.md; a SECURITY.md line only
if Codex asks, Q6), ARCHITECTURE, ADRs, the connection contract, PERFORMANCE.md and every Codex-owned file.

## 4. Schema and migrations

**None.** The catalog is derived per request from the existing `selected_group_id`/`context_version` (P-G014-9) and
current facts; no column, index, audit action/reason/operation, role or grant changes. The next free number after the
accepted G014 is **00011** (provisional: G014 owns 00010, and F3 may change 00010's SQL; re-checked at R0). It is
reserved only if R0 or Codex review requires a stored marker, which would then be its own flagged escalation (E-6).

## 5. Contracts and interfaces

### 5.1 Go surfaces (internal; `TestAccessBoundary`/`TestPackageBoundary` pins updated)

- access (exported, new): `ToolRequirement`, `RequireBinding/RequireMembership/RequireContext/RequireCapability`,
  `AuthorizeCatalog`. `AuthorizeCatalog` needs a verified MCP scope (`BeginVerifiedRequest`), runs one read-only
  snapshot of the P-G014-4 shape with the §1 fact options, writes nothing, returns **eligible names only** (no context,
  facts, decision object or anything a later operation could consume), and records `access.denied`/`mcp.request`
  exactly as `AuthorizeConnection` does when `contextCheck` fails. Its three outcomes are distinguishable to the
  caller (mcp package) only as: eligible names; `CodeContextDenied` (authorization refusal, recorded); or
  `CodeUnauthenticated`/`CodeUnavailable` (liveness or store/verifier failure, not an authorization decision). The
  mcp middleware maps them per §5.3 C-6 so that the wire never reveals whether a name was a hidden entry. It is
  discovery: no facade operation accepts its
  result, and no exported method returns an allow result for a target. The evaluator (`decide`), `Capability`,
  `Facts` handling and `manageableProfiles` stay unexported (AUTHORIZATION.md "no exported evaluator/check authority"
  preserved; D-2 records the names-only addition). Capability names are accepted only from the eleven canonical values.
- mcp (internal): `catalogEntry`, `catalogSpec`, `registerByRequirement`; the five production entries; the `_test.go`
  fixture seam. `New(cfg)` validates the production spec at start (an unknown requirement is a startup refusal, like
  G014's `config_invalid`).

### 5.2 Discovery and freshness (public behaviour, additive to `control-plane.md`)

| Rule | Pinned behaviour |
|---|---|
| F-1 list | `tools/list` returns exactly the registered entries of this request, in the SDK's deterministic sorted name order, `nextCursor` absent (five production tools; SDK `DefaultPageSize` 1000; no pagination contract is added) |
| F-2 refusal | invalid context (no membership, invalid selection, terminal/unknown connection, verifier/store failure) → G014's existing refusals (`-32040`, `-32041`, 401, 503); never a partial or cached catalog |
| F-3 local invalidation | a successful `switch_group` commits before its response; the next request on **that** connection lists the new context. Other connections of the same user are unaffected (connection-local row). **(r4, corrected contract)** A *definitive* refusal (`context_denied` — including a raced loss of the previous Group under the implemented F3 write —, `group_not_selectable`, `invalid_input`) or a known abort (the write-nothing conflict decided inside the guarded write, or any failure before the write) leaves selection, version and the next list unchanged. An `unavailable` answer does **not** by itself establish that nothing changed: after a possible commit (a lost COMMIT reply, or a failed context read after the selection was stored) the outcome is unknown; the client rereads with `get_current_context` or lists again before retrying, and the next `tools/list` reflects whichever state committed (the selection write is atomic: there is no partial state). The guarded write's lock order and previous/target authority recheck are G014's (P-G014-5, confirmed at R0: one Organization-only transaction locking the subject's previous and target Group facts, connection row last) |
| F-4 external changes | a committed membership/role/link/policy, session/user or lifecycle change governs the next `tools/list` and `tools/call` on every connection, with the same provider credential, no reconnect and **no version change**. Grants appear; removals disappear; a re-grant after revocation/logout/disable/expiry stays 401 (terminal rows never revive; P-G014-8) |
| F-5 no fallback | an invalid selected Group yields `context_denied` for `RequireContext` tools and `tools/list`; `switch_group` (clear/select) and `logout` remain reachable through their own handler rules, never an automatic clear |

### 5.3 Independent call authorization

| Rule | Pinned behaviour |
|---|---|
| C-1 hidden = unknown | a `tools/call` naming a capability entry that is not eligible on this request (never visible, or visible on an earlier list) is answered with the SDK unknown-tool envelope: JSON-RPC `-32602`, message `unknown tool "<requested name>"`, no handler invocation and no fixture effect. The envelope is the same classification and template as for a never-existing name, differing only in the echoed caller input; for the **same requested name** the response is byte-identical whether the name is absent from the spec or present but ineligible, in every context state (C-6). Truthful record: in a valid context the shared eligibility snapshot logs one `access.decision` allow (`opMCPAuthorize`, existing fields) and no audit row, for hidden and unknown names alike; the earlier "no facade call, no audit event" wording is withdrawn |
| C-6 refusal uniformity | every `tools/call` whose name is not a pre-registered production entry runs the same `AuthorizeCatalog` snapshot (capability entry or absent name alike). Outcomes, identical for both kinds of name: (i) `contextCheck` refusal (membership lost, selected Group invalid) → the SDK unknown-tool envelope for the requested name, plus one `access.denied`/`mcp.request` audit row and one `access.decision` deny log, exactly as a refused `tools/list` records; (ii) liveness failure → 401 as G014's transport rule (precedes the snapshot in practice); (iii) store/verifier failure → `-32041 unavailable`, not an authorization decision; (iv) allowed → registration then SDK dispatch (C-1). Pre-registered production entries never take this path, so `logout`, clear and the read tools keep their own G014 answers (`context_denied` for a refused read tool is G014-accepted behaviour, not a hidden-tool signal: those names are always listed when the list is served at all) |
| C-2 no list required | eligibility is decided on the call's own request; an earlier list or version never authorizes, and listing first is not required |
| C-3 visible still guarded | a registered production tool runs its unchanged G014 handler check: `logout` binding, clear membership, read tools membership+selection, select the guarded write that locks and rechecks the current and target Group (P-G014-5). Exact-target refusal "despite earlier listing" is the existing guarded operation: a Group returned by `list_groups` and then made unselectable refuses `switch_group` with `group_not_selectable`/`context_denied`. Fixture tools perform **no effect and no check of their own**; they only count invocations |
| C-4 ordering | per-request allowed outcomes for `tools/call`: (a) the tool runs under the facts of its own snapshot/guarded write; (b) unknown-tool (capability entry ineligible at registration); (c) the tool's own refusal (ineligible at execution, including a change that committed between a capability eligibility snapshot and the handler); **(d) (r4)** `unavailable` from a store or dependency failure — nothing changed when the call failed before its write (every read tool, the eligibility snapshot, and a select that aborted or wrote nothing), **unknown** when a select's commit outcome was lost (F-3): the next request observes the committed state, never a mixed one. Never: a handler run on authority older than its own snapshot, or a write that did not observe a conflicting committed change (G007 guarded-write linearization and G014 M3 locks, unchanged). In-flight authorized reads may finish on their snapshot; no cancellation or undo is promised |
| C-5 coherence | a `tools/list` result is one (selection, `context_version`, facts) snapshot; it cannot mix one Group's catalog with another's authority or publish a catalog as a selection (the server never writes a catalog) |

### 5.4 Notifications and re-list (Spec R5) — accepted rule (r1 review)

Facts (SDK v1.8.0, P-G014-3): tools are server-global and `AddTool` only notifies when `ListChanged` is advertised;
with `Stateless: true` every request is a temporary session closed at response, GET is 405, server→client requests are
rejected, and a notification not tied to an in-flight request is routed to the standalone or listen stream, which do
not exist; with `JSONResponse: true` even an in-request notification is routed out of band and dropped. The 2026-07-28
`subscriptions/listen` path is not claimed by G014.

| Rule | Pinned behaviour |
|---|---|
| N-1 advertisement | `capabilities.tools.listChanged = false` (unchanged, truthful; asserted) |
| N-2 delivery | **no `notifications/tools/list_changed` is delivered** on the inherited transport; no coalescing, timer or idle detection exists. Observation is request-bound: a local delta is known at the `switch_group` response; an external delta is observed by the next protected request. Codex confirmed this meets R5's conditional supported-delivery requirement; no SSE, background push or owner decision |
| N-3 selection revision | `context_version` in `get_current_context`/`switch_group` results is the **selection** revision only: it changes on a committed select/clear and never on external role/link/policy/session changes. A changed value is a hint that the selection moved; an unchanged value does not mean the catalog is current. Clients must re-list (`tools/list` is always fresh, F-1..F-4) regardless of the value. **(r4)** After an `unavailable` answer to `switch_group` the value may or may not have advanced (F-3): the client learns the current value by rereading, never by assuming the pre-call value |
| N-4 transport loss | reconnect = a new SDK client session on the same URL; it shares the connection's stored state (G014 IT-MCP-8) and lists fresh; no catalog is replayed |
| N-5 truth | the server never claims a catalog changed: no notification exists, and a `context_version` bump is not an effective-catalog signal (a select between two Groups with identical catalogs bumps it; a no-op select/clear does not). An idempotent re-list is identical. **(r4)** Symmetrically, the server never claims a catalog is *unchanged* through an `unavailable` answer: only a definitive refusal or a known abort carries that meaning (F-3) |

### 5.5 Audit, privacy, logging and query footprint (Spec R6; review items 2 and 4)

- Audit: no new action, reason or operation; catalog refusals (list, and non-registered-name calls under C-6) reuse
  `access.denied`/`mcp.request` through the existing snapshot context failure; in a valid context a hidden or
  unknown name records no audit row and one allow decision log — identical for both kinds of name.
- Logging: no new log event, operation or field. `AuthorizeCatalog` logs the existing `access.decision` record with
  `opMCPAuthorize` and the existing fields; nothing about tool names, counts, roles, Groups or profiles is added.
  `events.json` is unchanged.
- **Query inventory (real, per request):**

| Request | G014 today | G015 | Delta |
|---|---|---|---|
| `initialize`, `ping` | binding read only | same | **+0** |
| `tools/list` (production, no capability entries) | one snapshot: liveness, selection, `ReadFacts{Groups: []}` (2 statements) or `{Groups: [selected]}` (3–4) | the same snapshot via `AuthorizeConnection` | **+0** |
| `tools/call` (one of the five production tools) | the handler's own transaction | same; registration issues no statement | **+0** |
| `tools/call` naming anything else (unknown name; in tests a capability entry) — production too | SDK unknown-tool, no statement | one snapshot of the G014 shape (`Groups: []`/`[selected]`, 2–4 statements; plus the capability reads below only when capability entries exist — never in production), then the C-6 answer | **+1 snapshot on the error path only**; bounded by G014's per-connection rate limit (20/s, burst 40) and deadline; the five tools' calls are unaffected |
| `tools/list` with capability entries (test fixtures only) | — | one snapshot of the G014 shape with the §1 options: Organization kind +0; Group/profile kinds `Groups: nil, AllProfiles: true` (+4 Organization-only: all Group bindings, Group rows, associations, all profiles) or `Groups: [selected], AllProfiles: true` (+2); sets = the subject's bound Groups and the Organization's profiles | **+1 snapshot, ≤ G014 + 4 statements** |
| `tools/call` naming a capability entry (test fixtures only) | — | the same snapshot, then the fixture (no effect) | **+1 snapshot** |

  No lock is taken (plain reads in a REPEATABLE READ snapshot); no index, cache or SLO. The sets are the ones the
  accepted G007 Organization-only profile path already reads (PG-001 disposition). Production therefore has **no
  statement, round-trip, set or lock delta for the five controls, `tools/list`, `initialize` or `ping`**; the only
  production addition is the C-6 snapshot on calls to names that are not tools (an error path, rate-limited). The r1
  claim that PG-006's owner disposition covered an extra production snapshot is withdrawn; no owner performance
  acceptance is inferred (r2 review, D-4). E-MCP-1..3 are reused unchanged for production; E-MCP-4 asserts the
  capability-path counts above constant in Group count on the fixture plus 30 Groups and records the C-6 error-path
  count. A future Goal that adds a capability-mapped production tool must inventory that tool's cost under
  PERFORMANCE.md then.

## 6. Validation batches and test mapping

Existing suites were inspected (staging and the candidate): the G007 oracle/matrix and `TestAccessPolicy`
(B-POLICY-IT), `TestAccessOrdering` O-1..O-15 (B-POLICY-ORDERING), `TestRLSOrdering` OR-1..5 (B-RLS-ORDERING),
`TestIsolationAttacks` (G010, in B-OBJ-IT), `TestAudit` (B-AUDIT-IT), G014's `TestMCPTransport/Lifecycle/Audit/
Recovery`, `TestMCPOrdering/Footprint`, `TestMCPE2E`, unit `mcp_test.go`, `mcp.spec.ts`, and the superseded archetype
`skills/saas-foundation/validation/isolation/mcp` (pattern precedent only; protected, not reused). Finding: **no
existing test covers filtered listing, hidden-versus-unknown, catalog freshness across authority changes, catalog
isolation across connections or the notification rule**; everything else G015 touches is covered and reused.

### 6.1 Test-only fixtures (Spec R1; AC1/AC2)

`core/internal/mcp/catalog_fixture_test.go` (`package mcp`, `_test.go`, never compiled into the binary) registers four
capability-mapped tools through the unexported `extraEntries` seam, each with an invocation counter and **no effect
and no check** (review item 3): `test.org_policy` (`organization.manage_policy`), `test.group_update`
(`group.update`), `test.profile_update` (`business_profile.profile.update`), `test.audit_read` (`audit.read`). They
use only canonical capabilities and the existing two-Organization world (P-G014-10), extended with the existing
helpers by an **Organization admin of A with no Group bindings** (`grantOrg`), a Group admin of a1, a profile linked
to a1 (and shared with a2) and an **unlinked profile of A**, so that the Organization owner/admin `all`/`Y` paths
(any Group, any profile including unlinked, independent of selection), the Group-derived paths (`own`, `linked`,
`linked+flag`), link removal and the policy flag are all exercised. Their
schemas are closed. `New(cfg)` never sets the seam; `TestToolCatalog` and B-MCP-SOURCE prove the production catalog
is exactly the five names.

### 6.2 AC and scenario mapping (reuse / extend / new, with path and check ID)

| Scenario group | Test (path · check ID) | Kind | Covers / why new |
|---|---|---|---|
| (1) finite catalogs | `core/internal/mcp/mcp_integration_test.go` IT-MCP-1 · B-MCP-IT | reuse | exact five names, `ListChanged:false`, retired/unknown names error, GET 405 (AC1, AC8) |
| (1) | `core/internal/mcp/mcp_test.go` `TestToolCatalog`, `TestPackageBoundary`; `infra/validation/run.sh` B-MCP-SOURCE · B-UNIT-CORE/B-CHECKS-ROOT, lint-core | extend | add: production spec = five entries with their requirements; `extraEntries` nil from `New`; unknown requirement refused; importer pins unchanged (AC1) |
| (1) | `core/internal/tenancy/access/matrix_test.go` `TestToolEligibility` (new function on the existing `oracle*` tables and world) · B-UNIT-CORE | **new** (pure mapping) | uncovered: `eligible()` and `factQueryFor()` outcomes per canonical capability × role × Org-only/selected × flag on/off × link present/removed, asserted equal to "some oracle cell allows for that role and context" over the matrix's own target set (the oracle world's Groups g1–g4 and profiles P1–P3 including the unlinked P3, so the `all`/`Y` paths of an owner/admin without bindings are covered); nothing else evaluates visibility (AC2) |
| (1) | `core/internal/mcp/mcp_integration_test.go` `TestMCPCatalog` CAT-1 (new function, `setupMCP` fixture) · B-MCP-IT | **new** | uncovered: real SDK HTTP list differs by role (ownerA vs memberA1 vs Group admin), Org-only vs a1 selected, Group-admin downgrade, profile-link removal, policy off/on; **(r3) an Organization admin of A with no Group bindings lists `test.group_update`, `test.profile_update` and `test.org_policy`-class entries per the matrix in Organization-only context and, with a1 validly selected, still lists the Group entry (target restricted to a1) and the profile entry (the unlinked profile remains reachable), whereas the Group admin of a1, holding no Organization-admin path, loses the profile entry when a1's link is removed and never sees the unlinked profile**; production list unaffected by fixtures; sorted order and no cursor (AC1, AC2) |
| (2) select/clear | IT-MCP-7, IT-MCP-11, IT-MCP-12(a)(b) · B-MCP-IT | reuse | authorized select/clear, denied/foreign/unknown selection preserve state, **clear after selected-Group loss and whole-list `context_denied` through the new middleware**, downgrade shrinks `list_groups` (AC3, AC4; review item 1) |
| (2) | IT-MCP-13 · B-MCP-IT | reuse | **`logout` after Organization membership loss through the new middleware** while every other protected call is denied (AC3, AC9; review item 1) |
| (2) | `TestMCPCatalog` CAT-2 · B-MCP-IT | **new** | uncovered: `tools/list` before/after select and clear; **(r4)** a *definitively* refused select (`context_denied`, `group_not_selectable`, `invalid_input`) leaves version and list unchanged — the "unchanged" assertion is scoped to these cases; the write-nothing conflict is G014's O-MCP-2b (facade run), and no CAT case asserts "unchanged" after an `unavailable` answer (F-3: the lost-commit outcome is unknown by contract; the only assertion is that the next list reflects the committed state); grant, downgrade, unlink, policy flip, session revoke, disable, idle/absolute expiry each govern the next list/call with no new credential and no version change; store outage → 503 with no catalog; re-grant after terminal → 401 (AC3, AC4) |
| (2) | `core/internal/tenancy/access/rls_ordering_integration_test.go` OR-4 · B-RLS-ORDERING; `access/integration_test.go` IT-REVOKE-1..8 · B-POLICY-IT | reuse | the underlying stale-context denials per committed change (AC4, AC9) |
| (3) hidden/stale calls | `TestMCPCatalog` CAT-3 · B-MCP-IT | **new** | uncovered: the same requested fixture name called with the seam empty (unknown) and with the seam set but the caller ineligible (hidden) returns byte-identical envelopes, counter 0, no audit row and one allow decision log each; a different never-existing name shares classification/template; a tool listed earlier and hidden after a downgrade likewise; **(r3, C-6) after Organization membership loss and, separately, after selected-Group loss: the hidden name and an unknown name both answer the unknown-tool envelope byte-identically, each records exactly one `access.denied`/`mcp.request` audit row and one deny decision log with identical fields, while `logout` (and clear, in the selection case) still succeed and `tools/list` is `-32040`; with the store unavailable both answer `-32041` and record nothing**; exact-target refusal despite earlier listing uses `list_groups` then `switch_group` to a Group made unselectable in between (`group_not_selectable`/`context_denied`, selection/version unchanged); the paired authorized select and fixture call succeed once (AC5) |
| (3) | `core/internal/mcp/mcp_test.go` `TestErrorMapping` · B-UNIT-CORE | extend | pin the unknown-tool envelope template against the SDK default (AC5) |
| (3) | IT-MCP-11, IT-MCP-12(c)(d), IT-MCP-13 · B-MCP-IT | reuse | visible production tools keep G014 guards and the logout exception (AC5, AC9) |
| (4) ordering | `core/internal/tenancy/access/mcp_integration_test.go` `TestMCPOrdering` (activation race, O-MCP-1, -2, -2b, -3, -4 and G014's accepted F3 evidence **O-MCP-5** (a removal from the selected Group after the decision refuses a switch away; a clear succeeds) and **O-MCP-6** (such a removal waits for a held switch); 8/8 at `f45ad71`) · B-MCP-IT (facade run) | reuse | held switch vs read, serialized switches, decision/write selection change writes nothing, membership removal vs held switch, logout before switch, lost previous Group vs switch/clear (AC6) |
| (4) | `TestMCPOrdering` + **O-MCP-7, O-MCP-8** (r4: renumbered from r3's "O-MCP-5/6", which G014 now uses) · B-MCP-IT (count **8→10**, pinned at R0) | extend | uncovered: O-MCP-7 a catalog computed while a switch is held is coherent (old pair or new pair, never mixed; uses `selectLocked`); O-MCP-8 a downgrade committed between `catalogComputed` and a registered fixture's invocation leaves the fixture unregistered on the *next* request and, for a production tool, yields the tool's own refusal, never a stale run (AC6) |
| (4) | `TestAccessOrdering` O-1..O-15 · B-POLICY-ORDERING; `TestObjectsOrdering` OOR-6 · B-OBJ-ORDERING | reuse (unchanged) | G007/G008 read-snapshot and write-linearization proofs; no database interleaving is added because no new adapter gap exists (AC6) |
| (4) isolation | IT-MCP-8, IT-MCP-9, IT-MCP-13 · B-MCP-IT; E2E-MCP-2/3 · B-MCP-E2E | reuse | two A connections + B for `cross`, same-URL sharing, renewal keeps URL, binding regressions (AC7) |
| (4) | `TestMCPCatalog` CAT-4 · B-MCP-IT | **new** | uncovered: catalog isolation — select a1 on C1 with a fixture visible there; C2 and B lists unchanged; a fresh SDK client session (reconnect) and a renewal-due call list fresh without copying another connection's context (AC7) |
| (5) notifications | IT-MCP-1 (`ListChanged:false`, GET 405) · B-MCP-IT | reuse | truthful advertisement, no stream (AC8) |
| (5) | `TestMCPCatalog` CAT-5 · B-MCP-IT | **new** | uncovered: local delta (selection revision changes, re-list differs), external delta observed on the next list with the revision unchanged, no-op select leaves revision/list identical, a select between equal-catalog Groups bumps the revision with an identical list, reconnect re-list succeeds; asserts no notification frame is ever received by the SDK client's `ToolListChangedHandler` (AC8) |
| footprint | `TestMCPFootprint` E-MCP-1..3 + E-MCP-4 · B-MCP-IT | extend | E-MCP-1..3 unchanged for production (+0); E-MCP-4: capability-path statement counts of §5.5 constant across 30 extra Groups and extra profiles (AC9) |
| boundaries | `access/boundary_test.go` `TestAccessBoundary`, `TestProductionImporters`; `access/mcp_test.go` `TestMCPScopeBinding` · B-UNIT-CORE | extend | new exported names pinned (and `AuthorizeTool` absent); `AuthorizeCatalog` refuses an unverified scope and returns names only (AC9) |
| inherited | `TestIsolationAttacks` · B-OBJ-IT; `TestAudit` AT-INV · B-AUDIT-IT; `TestMCPAudit`, `TestMCPRecovery` · B-MCP-IT; `mcp.spec.ts` M-* · B-IDENTITY-E2E; B-MCP-HTTP, B-IDENTITY-GUARDS, privacy/canary scans | reuse | no new audit action; recovery/logout exceptions; setup page unchanged; sanitized logs (AC9) |
| evidence | Result map, path-selected local/MR/staging gates, B7 ledger, cleanup receipt | — | AC10 |

New-test justification summary: three new test functions (`TestToolEligibility`, `TestMCPCatalog` in package `mcp`,
`catalog_fixture_test.go` support) and three extensions (`TestMCPOrdering` +2, `TestMCPFootprint` +1, unit pins). Each
new case names behaviour no existing test exercises; all run on existing fixtures and existing check IDs. No
Goal-named check, no cloned suite, no browser or E2E addition (the five-tool production catalog shows no delta at the
edge), no repeated WorkOS qualification.

### 6.3 Batches

Before every batch: fetch `origin/staging`, assess the delta, rebase/reconcile the task worktree, run
`bash infra/ci/docker-desktop-mounts.sh`, take the coordinator batch lock, B-PROTECTED expectations, the standard
`run.sh` form with the G015 allow-list. Finish, fix, rerun, advance; no edits during a running batch.

**Executable selectors (r4; owner policy `d926012`, [focused selection before the first
run](../../../skills/sdlc-management/references/validation-integration.md#focused-selection-before-the-first-run)).**
Every planned local run names the command that actually executes; the harness has no per-check selector (its
granularity is the suite, `run.sh:41`), so Go-level `-run` expressions carry the focused work and suites are named
only where a required-check row or a changed harness pin can be proved no other way. Umbrella portions are stated,
not hidden. The r3 b1 `provenance` suite is dropped (no `go.mod`/vendor change); `lint-web`, `contracts`, `root`,
`rename`, `off`, `controls` and `ci-controls` are not run locally (no web, contract-schema, root, rename, off or
`infra/ci/*` change; the conditional `infra/ci/*` rows of §11.3 would add `bash infra/ci/validate-controls.sh --only
<affected IDs>` under their amendment). `all` is not proposed.

| Batch | Executable selectors | Reason / evidence |
|---|---|---|
| b1 static and unit | dev image `docker build --target check -f infra/docker/core.Dockerfile core`; in it, network none: `gofmt -l .`; `go vet ./...`; `go vet -tags integration ./internal/mcp/ ./internal/tenancy/access/`; `go vet -tags mcpe2e ./internal/mcp/`; `go test -count=1 -run '^(TestToolCatalog\|TestErrorMapping\|TestPackageBoundary\|TestCursor)$' ./internal/mcp/`; `go test -count=1 -run '^(TestToolEligibility\|TestAccessBoundary\|TestProductionImporters\|TestMCPScopeBinding)$' ./internal/tenancy/access/` | the changed packages' own unit pins (catalog spec, envelope template, import/one-call-site boundary, exported-surface pins, scope binding, the new oracle-table eligibility test); `TestCursor` because `mcp_test.go` is edited |
| b1 harness | `run.sh lint lint-base lint-core` (standard form: `--base`, `--allow-file` the G015 allow-list, `--owned-registry`, `--report-dir`) | `lint`: B-SHELL over the edited `run.sh`, B-SECRET-SOURCE, and B-SECRET-CANARY needs `lint` beside `lint-base`; `lint-base`: B-LITERALS/B-NAMING/B-DEPSCAN/B-IDENTITY-GUARDS over the new and changed Go files; `lint-core`: B-MCP-SOURCE, the import boundary and one-call-site pin that `catalog.go` and the access exports touch. Each is a static suite over the changed sources; no umbrella runtime work |
| b2 integration (focused) | disposable focus database as G014 attempt 2 (Compose project `g015-focus`, postgres only from `infra/compose/local.yml`, generated 0600 env; the dev core-check image on its data network with the run.sh IT env shape, `MCP_INVALIDATE_SQL_B64` set; `down -v` and leftover check after): `go test -tags integration -count=1 -timeout 30m -v -run '^(TestMCPTransport\|TestMCPLifecycle\|TestMCPAudit\|TestMCPRecovery\|TestMCPCatalog)$' ./internal/mcp/` (expect 7/6/1/1/**5**); `go test -tags integration -count=1 -timeout 30m -v -run '^(TestMCPOrdering\|TestMCPFootprint)$' ./internal/tenancy/access/` (expect **10**/1, four `E-MCP` lines incl. E-MCP-4); `go test -tags integration -count=1 -timeout 30m -v -run '^(TestAccessPolicy\|TestAccessOrdering)$' ./internal/tenancy/access/` | the B-MCP-IT content with the new counts; the policy/ordering regression over the edited `policy.go`/`mcp.go` (B-POLICY-IT/B-POLICY-ORDERING equivalents). **Reused, not rerun** (inputs unchanged by G015: no migration, RLS, objects or audit change): B-RLS-ORDERING `TestRLSOrdering`, B-OBJ-IT `TestIsolationAttacks`, B-AUDIT-IT `TestAudit`, B-MCP-CATALOG (`./internal/migrate/`), with the staging gate 17225023104 at merge `a8d146c` as the attributable pass; they run again inside the final `migrations` suite anyway |
| S-review | read-only security/concurrency review (delegate, §7); no run | findings resolved before final |
| final (local) | reconcile (fetch/rebase, `docker-desktop-mounts.sh`, batch lock), then `run.sh lint lint-base lint-core unit migrations` (standard form, G015 allow-list) | `unit` = B-UNIT-CHECKS, the only harness row recording core gofmt/vet/unit (it also runs the unchanged web and intelligence checks — umbrella portion acknowledged); `migrations` = the only harness selector that executes the **changed B-MCP-IT pin** (regex/count/PASS text in `run.sh`) and records B-MCP-IT, B-MCP-CATALOG, B-POLICY-IT/ORDERING, B-RLS-ORDERING, B-OBJ-IT and B-AUDIT-IT as required-check rows (its other ~25 rows are unaffected — umbrella portion acknowledged; **Codex disposition requested** for this suite as the smallest existing harness selector). B-SCOPE and B-REQUIRED-CHECKS come with the run. Expected: every executed row PASS; cleanup receipt rc 0 |
| **exception request: `ops`** | one `run.sh ops` after b2 and S-review, before the MR (G014 precedent: the single approved exception, PGR record) | **Reason:** B-MCP-E2E (`go test -tags mcpe2e -run '^TestMCPE2E$' ./internal/mcp/`, 6/6: exactly the five tools, all five called over the real TLS edge through the **changed** `server.go`/`tools.go` registration) and B-IDENTITY-E2E (`web/tests/browser/mcp.spec.ts` M-*, setup page unchanged) are reachable only through the `ops` orchestration (`suite_ops` → `variants_ops`/`variants_identity`: TLS edge, seed, private browser handoff, MCP driver, cleanup; `run.sh:4704-4706`); no narrower existing entrypoint exists (BACKLOG "Narrow identity/MCP E2E selection", open) and the Plan does not add one (Keep CI changes separate). Unaffected ops rows (B-IDENTITY-IT/PROBES/PRIVACY/CONFIG, B-OBS-*, B-OUTAGE, B-PERSIST, B-SECRET-ROTATION, B-DOCKER-ONLY) run alongside — acknowledged. **Alternative if declined:** no local `ops`; B-MCP-E2E/B-IDENTITY-E2E evidence comes from the hosted staging full gate after the authorized conditional merge (path map: `core/internal/{mcp,tenancy}/` and `infra/validation/` select full), recorded in the Result as hosted-only |

Manual testing and operator feedback: none required (no UI, no third-party client claim); state so in the Result. MR
fast runs automatically; the post-merge staging full gate supplies hosted evidence after an authorized conditional
merge; success/canceled satisfies, failure fails, pending does not qualify. Targeted reruns after a fix follow the
attributable-coverage rule (not included):
the smallest failed `-run` expression plus the checks the fix affects, never a restart for a new SHA alone.

## 7. Delegation

- **Implementer:** one, effort high (default). Worker name `g015-agent-<id>-impl` when the native tool supports it;
  otherwise the generated `agent-<id>` with task/role/path recorded in the Result. Owns every path in §11.3.
- **Security/concurrency reviewer: required**, read-only, after b2 (tier is architecture/security). Scope: the
  eligibility rule and fact options versus the oracle, hidden-equals-unknown construction, recovery/termination paths
  reachable through the middleware, no check-then-effect path (names-only `AuthorizeCatalog`), no cached authority,
  log privacy, exported-surface minimality, C-4 ordering.
- **R0 (done, r4):** executed by the successor planning worker without a delegate; nothing material (§0.2).
- No delegate edits; Claude integrates and validates. One shared B7 ledger.
- Planner record: §2 (task G015, role plan, worktree `agent-<id>`, branch `g015-plan`).

## 8. Sequence

1. G014 final ACCEPT (with the accepted F3 resolution) → R0 (§0.2) → Plan revision if material → Codex review →
   binding and release.
2. access: `ToolRequirement`/`eligible`/`factQueryFor`/`AuthorizeCatalog`, hook; unit tests and pins.
3. mcp: `catalog.go`, requirement-aware registration and middleware, fixture seam, unit pins, `catalog_fixture_test.go`.
4. Integration: `TestMCPCatalog` CAT-1..5, `TestMCPOrdering` +2, `TestMCPFootprint` +1; `run.sh` pins.
5. Docs: MCP.md, AUTHORIZATION.md, control-plane.md.
6. b1 → b2 → S-review → reconcile → final; Result; safekeeping pushes; Codex review; MR to staging.

## 9. Acceptance-criteria mapping

| AC | Changes | Evidence |
|---|---|---|
| AC1 | §1 catalog spec; `New` validation; fixture seam nil in production | IT-MCP-1 (reuse); `TestToolCatalog`/`TestPackageBoundary`/B-MCP-SOURCE (extend); CAT-1 production list (new); IT-MCP-11 no catalog on invalid context (reuse) |
| AC2 | `eligible()`/`factQueryFor()` on `decide`/`linkedPath`; `RequireCapability` closed to the eleven canonical values | `TestToolEligibility` (new, oracle tables); CAT-1 (new) |
| AC3 | F-3/F-5; requirement-aware registration; no change to G014 select/clear/logout checks | IT-MCP-7/11/12/13 (reuse, through the middleware); CAT-2 (new) |
| AC4 | F-4; request-bound snapshot; selection-only revision | CAT-2 (new); OR-4, IT-REVOKE-* (reuse); IT-MCP-9 terminal refusals (reuse) |
| AC5 | C-1..C-3; existing guarded operation as the exact-target guard | CAT-3 (new); `TestErrorMapping` (extend); IT-MCP-7/12/13 (reuse) |
| AC6 | C-4/C-5; `catalogComputed` hook | O-MCP-1..6 and -2b (reuse, incl. G014's F3 evidence); O-MCP-7/8 (extend; r4 renumbering); O-1..O-15, OOR-6 (reuse, unchanged) |
| AC7 | connection-local rows unchanged | IT-MCP-8/9/13, E2E-MCP-2/3 (reuse); CAT-4 (new) |
| AC8 | N-1..N-5 | IT-MCP-1 (reuse); CAT-5 (new) |
| AC9 | §5.5; boundary pins; no migration/cache; production +0 | E-MCP-1..3 (reuse) + E-MCP-4 (extend); boundary/scope-binding pins (extend); `TestIsolationAttacks`, `TestAudit`, `TestMCPAudit/Recovery`, scans (reuse); diff review; S-review |
| AC10 | Result, allow-list, ledger | Result with content/input provenance, command outcomes, reused vs new evidence; MR fast + staging full gate; cleanup receipt; Codex ACCEPT separate |

## 10. Risks and escalations

**Decisions — all closed (r1 and r2 reviews; no owner decision):**

| # | Item | Resolution |
|---|---|---|
| D-1 | Notification transport | truthful `listChanged:false` on the inherited stateless JSON transport, fresh lists/calls, no SSE/background machinery (§5.4) |
| D-2 | Exported access surface (`ToolRequirement`, four constructors, names-only `AuthorizeCatalog`) | approved as names-only discovery within the existing import boundary; no later operation consumes it as authority; no exported evaluator (§5.1) |
| D-3 | Hidden/unknown envelope keeps the SDK default (`-32602`, echoes the requested name) | acceptable with the C-6 refusal uniformity (§5.3) |
| D-4 | Performance record | production +0 for the five controls is credible subject to implementation evidence; no owner performance acceptance inferred or transferred; a future production capability tool assesses its own cost (§5.5) |

**Escalation items (identified for review; none decided here):**

| # | Item | Class |
|---|---|---|
| E-1 | Public contract additions: F-1..F-5, C-1..C-5, N-1..N-5 appended to `control-plane.md` | public MCP contract (additive) |
| E-2 | Authorization boundary: access exports (D-2); `TestAccessBoundary` pin change; names-only discovery result | authorization boundary |
| E-4 | Validation: B-MCP-IT regex/count pins; `TestMCPOrdering` **8→10** (pinned at R0; G015's cases are O-MCP-7/8); `TestMCPCatalog` (package `mcp`) 5; no new ID or profile change; the §6.3 selectors, the `migrations`/`unit` suite dispositions and the `ops` exception request | validation mapping |
| E-5 | G014 drift: **closed at R0** — every P-G014 row confirmed at `f45ad71`, no material change (§0.2 R0 record, §0.3) | predecessor reconciliation |
| E-6 | **None proposed:** migration, index, audit action/reason/operation, log event/operation/field, Redis/in-process cache, new tool/capability/role, pagination contract, protocol upgrade (2026-07-28/`subscriptions/listen`), vendor, service, CI redesign, web change, G016 host work | explicit none |

(E-3 of r1 — a new log operation — is withdrawn.)

**Risks:** R-a SDK behaviour differs from the read source (verified in b1 by the unit tests; no planning probe);
R-c `catalogComputed` hook placement inside the eligibility snapshot (test-only; nil in production); R-d count/regex
pins in `run.sh` (allow-listed); R-e scanner-sensitive text in docs (neutral wording); R-f G014 acceptance slips or
F3 reshapes P-G014-5/-9 (readiness reassessment at R0, not execution); R-g the fixture world additions (Group admin,
linked/unlinked profiles) must use existing helpers only (`grantGroup`, profile/link helpers of `setupIT`/`setupMCP`);
if a helper is missing, add it to the fixture file, never to production. (R-b of r1 is withdrawn: the required facts
are read with the pinned options.)

**Stop conditions (Goal):** any owner stop; weakened identity/audience/user/client/Organization/liveness checks; a
recovery/clear/logout path made unreachable; stale authority used for an effect; shared selection/catalog authority;
a new public tool/capability/role, schema/index or infrastructure need; unexpected fan-out/scans/contention; native
permission refusal; context overrun or exhausted budget.

## 11. Budget, forecast and scope

### 11.1 Budget

**B7, checkpoint after attempt 4** (Goal). Planning spends no attempt (r1–r3 included). Forecast: **1 attempt
likely, 2 plausible** (≈1.5–2.5k lines including tests; the largest uncertainty is the accepted F3 shape at R0).
Effort: implementation high; the owner may select xhigh for the ordering work (O-MCP-5/6). Verify actual settings or
report unknown; no budget reset.

### 11.2 Count pins (pinned at R0 against the merged G014 at `f45ad71`)

| Pin | After G014 (merged, `f45ad71`) | After G015 |
|---|---|---|
| required-checks rows / migrations / ops | 117 / 33 / 14 | 117 / 33 / 14 (unchanged) |
| B-MCP-IT `TestMCPTransport` / `TestMCPLifecycle` / `TestMCPAudit` / `TestMCPRecovery` / `TestMCPCatalog` (package `mcp`) / `TestMCPOrdering` / `TestMCPFootprint`; `E-MCP` lines | 7 / 6 / 1 / 1 / — / **8** / 1; 4 | 7 / 6 / 1 / 1 / **5** / **10** / 1; 4 or 5 (the harness counts `^E-MCP(-3)? ` lines, `run.sh:3158-3159`; an `E-MCP-4 ` line does not match that pattern, so the implementer either extends the grep and pins 5 — `run.sh` is allow-listed — or records E-MCP-4 as a `TestMCPFootprint` assertion and leaves 4; stated in the Result) |
| B-MCP-CATALOG `TestMCPCatalog` (package `migrate`) / `TestMCPMigration` | 3 / 5 | unchanged |
| B-MCP-E2E `TestMCPE2E` | 6 | unchanged (no E2E addition) |
| profiles fast / integration / full; VA7; VU1 full-pipeline union | 35 / 75 / 110; 80; 111 | unchanged |

### 11.3 B-SCOPE allow-list (final at R0; for Codex to import byte-exact as `management/v0/results/G015-scope-allowlist.txt` at binding)

The r3 list is confirmed complete against §3 and the R0 record: no R0 hit adds a path (`connections.go`,
`identity/*`, `events.json`, `required-checks.tsv` and `infra/ci/*` stay outside). The initial Result file already
exists (drafted at R0); its row stays so that execution-time edits remain in scope.

```text
# Create
^core/internal/mcp/catalog\.go$
^core/internal/mcp/catalog_fixture_test\.go$
^management/v0/results/G015-(scope-allowlist\.txt|dynamic-capability-exposure-result\.md)$
# Modify: production
^core/internal/mcp/(server|tools|errors)\.go$
^core/internal/tenancy/access/(mcp|policy)\.go$
# Modify: tests and fixtures
^core/internal/mcp/(mcp_test|mcp_integration_test|env_integration_test)\.go$
^core/internal/tenancy/access/(boundary_test|mcp_test|matrix_test|policy_test|mcp_integration_test)\.go$
# Modify: harness
^infra/validation/run\.sh$
# Modify: docs and contract
^docs/foundation/(MCP|AUTHORIZATION)\.md$
^contracts/mcp/control-plane\.md$
```

**Conditional** (only with a Codex-reviewed amendment): `core/internal/tenancy/access/(context|errors|log)\.go` if a
new code or log wording is needed; `infra/validation/required-checks.tsv` and
`infra/ci/validate-controls.sh` (only if a pin text must change); `skills/saas-foundation/references/mcp-control-plane.md`
(guidance note); any R0 reconciliation hit.

## 12. Open questions

**None open.** Resolved by the r2 review, recorded here: Q2 names-only discovery API within the existing import
boundary; Q3 the inherited whole-list refusal is retained and direct recovery paths remain usable; Q4 the SDK error
template with the C-6 fix, and production +0 subject to implementation evidence; Q5 one `TestMCPCatalog` inside the
existing B-MCP-IT driver, extending current fixtures and exact regex/count pins, no new check ID or suite; Q6 the
MCP.md disclosure suffices (no SECURITY.md expansion; `docs/SECURITY.md` leaves the allow-list); Q7 exact test counts
are reconciled at R0 against the accepted G014/F3 — provisional counts are not authority to drop inherited tests or
freeze an obsolete count. Q1 of r1 was closed by the D-1 resolution.

**R0 items for Codex disposition (r4; none is a design question):**

- Q8 `TestMCPCatalog` naming: G014 already has `TestMCPCatalog` in `core/internal/migrate` (B-MCP-CATALOG, 3
  subtests). G015's new integration function in package `core/internal/mcp` (B-MCP-IT, CAT-1..5) keeps the r3 name
  by default — different package, different harness log and counter — but the two PASS texts would both read
  "TestMCPCatalog n/n". If Codex prefers unambiguous evidence, the implementer names it `TestMCPDiscovery` (regex
  `^(TestMCPTransport|TestMCPLifecycle|TestMCPAudit|TestMCPRecovery|TestMCPDiscovery)$`; CAT IDs unchanged). Names
  only; no other change.
- Q9 the `unit` and `migrations` suites at the final local run and the `ops` exception (§6.3): disposition before
  launch, as the focused-selection policy requires. The alternative for each is stated in §6.3.

## 13. Revision history

- r4, 2026-10-16 UTC: **R0 reconciliation** after the G014 merge (`a8d146c`) and qualified staging gate; base
  `f45ad71`; r3 archived byte-exact by the R0 worker (A3). Nothing material: P-G014-1..13 confirmed with file:line
  (§0.2 R0 record), setup-listing changes assessed as no-impact (§0.3), the `unavailable` wording of F-3/C-4/N-3/N-5
  and CAT-2 reconciled with the corrected contract, G015's ordering cases renumbered O-MCP-7/8 with `TestMCPOrdering`
  8→10, executable selectors named for every local run with the `migrations`/`unit` dispositions and the `ops`
  exception request (§6.3, Q9), the `TestMCPCatalog` naming note (Q8), inputs rebound (§2), allow-list final (§11.3),
  initial Result drafted. Successor worker after the 350k handoff. Read-only; no build, test, install, Docker, probe
  or provider call.
- r3, 2026-10-16 UTC: correction for the r2 review (not included), two items: (1)
  eligibility follows the canonical matrix — Organization owner/admin `all`/`Y` paths reach any Group and any profile
  (including unlinked) without Group bindings and independent of selection, Group-derived paths keep bindings, links,
  selection and the flag; fact bounds and CAT-1/`TestToolEligibility` corrected, `decide` untouched; (2) C-6 refusal
  uniformity — non-registered names share one `AuthorizeCatalog` path, a context refusal answers the unknown-tool
  envelope with a truthful denial record for hidden and unknown names alike, store failure answers `unavailable`;
  C-1 wording corrected; CAT-3 extended; §5.5 inventories the error-path snapshot. D-1..D-4 and Q2–Q7 recorded as
  closed. Base `cd54d44`; r2 archived byte-exact by Codex. Read-only; no build, test, install, Docker, probe or
  provider call.
- r2, 2026-10-16 UTC: correction for the r1 review (not included): requirement-aware
  per-call authorization preserving the clear/logout exceptions; actual fact options and canonical capability names;
  `AuthorizeTool` removed (names-only discovery); real query inventory with production +0 and the PG-006 claim
  withdrawn; echo/byte-identity, selection-only revision, sorted order and log-field corrections; D-1 recorded as
  resolved; P-G014 rows re-checked at `c8f8f9af` (O-MCP-2b, mcp log events, `csrf_token`, F3 pending). Base
  `51e4df0`; r1 archived byte-exact by Codex. Read-only; no build, test, install, Docker, probe or provider call.
- r1, 2026-10-16 UTC: initial provisional Plan at the G014 draft MR trigger (candidate `d4d2f1cd`, staging `37aa23a`).
  Read-only inspection; no build, test, install, Docker, probe or provider call. Reviewed: CORRECTION LOOP.
