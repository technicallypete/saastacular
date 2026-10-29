> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G063 — Account MCP connections and friendly names — Implementation Plan (r1)

Goal: [G063](../goals/G063-account-mcp-connections.md) · Spec: [G063 Spec r1 + 2026-10-21 priority note](../specs/G063-account-mcp-connections-spec.md) ·
[ADR-005](../../../adr/ADR-005-tenant-bound-mcp-connections.md) · [ADR-007](../../../adr/ADR-007-verified-domain-routing-and-authentication.md) ·
[connection context](../../../../contracts/mcp/connection-context.md) · [control plane](../../../../contracts/mcp/control-plane.md) · [MCP.md](../../../foundation/MCP.md) ·
Codex r0 review (not included) (sha256 `87d874cc209f9b7860e46a2b0bd7e4749dd5a8f70186586511777e4946571897`).
Status: **r1 submitted for Codex Plan review (one correction pass over r0; the five r0 corrections and the D-1…D-6 rulings
are applied below). PLANNING ONLY: no source, build, test, Docker, Compose, npm or Go command was run; implementation
is NOT started and NOT released. Zero of B3 execution attempts spent (checkpoint after attempt 2).**
Base: `g063-plan-r1` branched from `g063-plan` `bea8b161ab0fbef717909a026487f64e278dbbe6` (r0) and merged with
`origin/staging` `fb9a8ba3e36ff5048b7fe3b88c1004842138e803` (management-only deltas since r0's `7d62e8ff`: `git diff --stat
7d62e8ff fb9a8ba3 -- core web infra contracts docs` is empty; the delta is PROGRESS, the G062 Result/review records and the
G063 r0 review record). Bound inputs with blob SHAs at `fb9a8ba3`: input inventory (not included).
Planner: Fable 5.1 (requested high; runtime effort unknown from inside the agent), isolated worktree
`.claude/worktrees/agent-<id>`, branch `g063-plan-r1`. File:line references are at `fb9a8ba3` (identical
blobs to `7d62e8ff` for every source, test, harness, contract and doc file).

## 0. Order, gates and tier

- **Delivery order (owner priority, 2026-10-21):** G060 → G055 → G062 → **G063** → G020 → G057 → G021 → OS001 → OS002.
  G063 final ACCEPT gates G020 integration/final acceptance; isolated overlap needs the normal supervisor execution release;
  G020's accepted technical Plan is not rewritten for order alone (see D-3 on migration numbering).
- **Tier:** architecture/security/data (persisted metadata, SQL/RLS/grant delta, a mutation at the session/tenant boundary).
  Explicit Codex Plan ACCEPT required; the D-items in §10 are now **recorded rulings**, not open requests.
- **Gates before any implementation:** Codex Plan ACCEPT of r1; imported scope allow-list (§9); current input bindings
  reconciled at the exact staging SHA of release; G062 integration state per the order above; budget B3 unchanged; isolated
  task branch/checkout (`g063-impl` from the release SHA); resource readiness (one build slot, test-run pool of two);
  **recorded supervisor execution release**; the D-3 slot re-check (§10). Plan existence, merge or CI alone is not release.
- **Budget:** B3, checkpoint after attempt 2, zero spent. Forecast: 1–2 attempts.
- **Context policy:** 150k checkpoint / 200k safe handoff for the writer.

## 1. Approach (one sentence each)

1. Add one nullable `display_name` column to `app.mcp_connections` by forward additive migration `00019`, with a defensive
   CHECK, minimum runtime column grants and a trigger-owned audit action for name changes; policies and the guard unchanged.
2. Core owns the Unicode validation/normalization, strict wire decoding of the name field, the create-with-name path and
   one new guarded facade operation `RenameMCPConnection` (own live pending/active row, current Organization membership,
   liveness locked against revocation in the order user → browser session → target 'mcp' session → connection row).
3. The internal setup contract gains `display_name` on create (optional) and list rows (always present, `null` for legacy)
   and one new route `POST /internal/v1/mcp/connections/rename`; one new closed code `invalid_name` (D-2).
4. Web gets a shared server-rendered `ConnectionManager` component embedded in `/account` (through a shared
   `AccountSections` composition) and `/connections/mcp`, thin form routes (`create`, `revoke`, new `rename`) with a closed
   `return_to` selector, page-specific status/cursor params, and strict raw-bytes reading of the name field.
5. Validation reuses the existing B-MCP-*/B-IDENTITY-E2E/B-UNIT-*/B-AUDIT-*/B-BROWSER-ROOT controls and extends them; one
   new focused web form test file, one new core name-validation test file and one test-only fixture page (existing
   fixture-build pattern) cover genuinely new boundaries.

## 2. Files and modules

| Area | Path | Change |
|---|---|---|
| Migration | `core/migrations/00019_mcp_connection_names.sql` | **new** (§3) |
| Migration tests | `core/internal/migrate/mcp_integration_test.go` | extend: `TestMCPNamesCatalog`, `TestMCPNamesMigration` (§6) |
| Audit catalog | `core/internal/audit/audit.go` (`TriggerActions`, `audit.go:85`) + `audit_test.go:102` (39/7/46 → 40/7/47) | add `MCPConnectionRenamed Action = "mcp.connection_renamed"` |
| Identity | `core/internal/identity/mcp.go` | `NormalizeMCPDisplayName`, `ErrInvalidMCPName`, `MCPConnectionInfo.DisplayName *string` |
| Identity store | `core/internal/identity/mcp_store.go` | insert/list SQL gain `display_name`; new `ReadMCPConnectionSessionTx`, `RenameMCPConnectionTx` (§4.2) |
| Identity tests | `core/internal/identity/mcp_name_test.go` (**new**), `mcp_integration_test.go` (extend: `TestMCPIdentityRename`) | §6 |
| Facade | `core/internal/tenancy/access/mcp.go` | `RenameMCPConnection`, `opMCPRename = "mcp.connection.rename"`; `CreateMCPConnection` gains `displayName *string` |
| Facade tests | `core/internal/tenancy/access/mcp_test.go`, `mcp_integration_test.go` | extend (scope binding; O-MCP-9 both orderings + no inverse-order deadlock) |
| **Access boundary (r1 C1)** | `core/internal/tenancy/access/boundary_test.go` | `pinnedExportedMethods` (`:78`) += `Service.RenameMCPConnection`; identity exported functions taking a user (`:879`) += `Service.ReadMCPConnectionSessionTx`, `Service.RenameMCPConnectionTx`; `identityMCPProtectedFuncs` (`:494`) unchanged by construction (§4.2: the new store helpers touch `app.mcp_connections` only; the target session lock reuses `liveSession` through `LockLiveSession`) — asserted, not assumed: if the writer's SQL names `app.sessions`/`app.users`, the inventory gains the function as `callee` and the Result records it |
| **Audit integration (r1 C1)** | `core/internal/tenancy/access/audit_integration_test.go` | call site `:1250` gets the new `nil` name argument; AT-INV (`:1209`) adds one real `RenameMCPConnection` so `mcp.connection_renamed` is observed (its completeness loop over `audit.TriggerActions` would otherwise fail); AT-ATOM (`:529`) adds the rename fault case (§6) |
| Setup routes | `core/internal/mcp/connections.go`, `routes.go` | `Rename` handler, strict name-field decoding (§4.3), `createBody.DisplayName json.RawMessage`, `connectionView` + `display_name`, `RouteConnectionsRename`, `setupFailure` maps `access.CodeNotFound` → `not_found` |
| Server | `core/cmd/server/main.go` | mount `RouteConnectionsRename` next to `main.go:271` |
| Core tests | `core/internal/mcp/mcp_test.go` (`TestSetupNameDecoding` **new function, existing file**), `mcp_integration_test.go` (IT-MCP-3 extend, IT-MCP-15 new case under `TestMCPLifecycle`, `TestMCPRecovery` extend) | §6 |
| Log catalog | `infra/validation/tools/internal/logrecord/events.json` | `mcp.setup.operation` += `rename`; `mcp.setup.reason` += `invalid_name`; `access.decision` operation enums (lines 433, 488) += `mcp.connection.rename` |
| Harness pins | `infra/validation/run.sh` | mechanical pins only: B-MCP-SOURCE (four setup routes; `main.go` reviewed digest), B-MCP-CATALOG `-run` regex and PASS text, B-MIGRATE-SET/CLI/CATALOG version texts (00019), B-AUDIT-CATALOG/B-AUDIT-IT action counts (46 → 47) and PASS texts, **B-WEB-BOUNDARY** (`run.sh:986` request-context allow regex += `app/connections/mcp/rename/route.ts`, `components/account/(connection-manager|account-sections).tsx`; `WEB_ROUTES` inventory and the `web_routes_ok` message += the rename route; PASS text) |
| Browser fixture build | `infra/docker/web-browser.Dockerfile:11-12` | one more `cp tests/fixtures/account-mcp/page.tsx app/account-fixtures/page.tsx` in the `fixture` stage (test-only; production image unchanged) |
| Web client | `web/lib/mcp-client.ts` | `displayName` in `ConnectionView`, validators, `renameConnection`, bound recompute (§5.3), `invalid_name`, `normalizeDisplayNameInput`, `strictFormField` (§5.4) |
| Web components | `web/components/account/connection-manager.tsx` (**new**, §5.1), `web/components/account/account-sections.tsx` (**new**, §5.2: the signed-in `/account` body extracted so the fixture renders the real composition) | |
| Web pages | `web/app/account/page.tsx`, `web/app/connections/mcp/page.tsx` | embed (§5.2) |
| Web routes | `web/app/connections/mcp/{create,revoke}/route.ts`, `web/app/connections/mcp/rename/route.ts` (**new**) | `return_to` selector; strict name field (§5.4) |
| Web fixture | `web/tests/fixtures/account-mcp/page.tsx` (**new**, test-only) | the rendered unavailable-state test page (§6, D-5) |
| Web tests | `web/tests/unit/mcp-connections.test.mjs` (extend), `web/tests/unit/mcp-forms.test.mjs` (**new**), `web/tests/browser/mcp.spec.ts` (extend), `web/tests/browser/account.spec.ts` (extend: S-ACCOUNT-MCP-UNAVAILABLE) | §6 |
| Contracts/docs | `contracts/mcp/control-plane.md`, `docs/foundation/{MCP,AUDIT,DATABASE-ISOLATION,MIGRATIONS,WEB-SHELL}.md`, `docs/SECURITY.md` §5 row | §7 |
| Result | `management/v0/results/G063-account-mcp-connections-result.md` | **new** at execution |

`CreateMCPConnection` call sites (all four change with the signature): definition `core/internal/tenancy/access/mcp.go:568`;
`core/internal/tenancy/access/mcp_integration_test.go:41`; `core/internal/tenancy/access/audit_integration_test.go:1250`;
`core/internal/mcp/connections.go:385`.

Not touched: historical migrations `00001`–`00018`, `app.mcp_connections_guard()` body, the three RLS policies,
`sessions_insert`, provisioning SQL, `web/components/ui/**`, `web/lib/session.ts` (`readForm`/`field` keep their lossy
decoding for every legacy field; the name is re-read strictly by the route, §5.4), `infra/identity/**`, provider/OAuth
configuration, MCP tools/catalog, the transport/binding reads (`sqlMCPBinding`, `readMCPBinding`), `docs/adr/**`,
Codex-owned Goal/Spec/roadmap/progress, G020 artifacts.

## 3. Migration `00019_mcp_connection_names.sql` (additive, forward, one goose transaction)

- **Preconditions** (00018 style): run as the `app` owner; the five roles exist with the 00010 attribute checks; exact
  version-18 inventories as pinned by the 00018 postconditions (forced-table list, relation count, policy/ACL/function/
  trigger inventories, `audit.events_attribute` and `audit.capture_mcp_connections` body digests); no column
  `app.mcp_connections.display_name`; no `mcp.connection_renamed` in `events_action_valid`.
- **Column:** `ALTER TABLE app.mcp_connections ADD COLUMN display_name text NULL CONSTRAINT mcp_connections_display_name_valid
  CHECK (display_name IS NULL OR (char_length(display_name) BETWEEN 1 AND 80 AND display_name !~ E'[\\x00-\\x1F\\x7F-\\x9F\\u2028\\u2029]'
  AND display_name = btrim(display_name)))`. Nullable, no default, no table rewrite, **no backfill** (every existing row
  stays `NULL` = "Unnamed connection"). `char_length` counts code points. The CHECK is defense in depth for the stored shape
  (length, no C0/DEL/C1 control, no LS/PS, no ASCII edge space); core's Unicode trimming/validation (§4.1) is authoritative.
- **Grants:** `GRANT INSERT (display_name), UPDATE (display_name) ON app.mcp_connections TO <db>_runtime`. Nothing else:
  no DELETE, no new role, no BYPASSRLS, no definer function. The runtime INSERT column set becomes nine, UPDATE six.
- **RLS:** unchanged. `mcp_connections_select` (own rows, no role) keeps terminal-row ownership/revoke behavior;
  `mcp_connections_update` (own rows **with a current Organization role**) already requires membership for any UPDATE, so
  the rename cannot widen beyond create/list membership and cannot reach a foreign user/tenant row at the database either (D-4).
- **Guard trigger:** `app.mcp_connections_guard()` unchanged (body digest unchanged). It never inspects `display_name`,
  so a name-only UPDATE passes; state/client/binding/selection rules stay exactly as pinned by CT-MCP-1.
- **Audit (D-1, approved):** one trigger-owned action `mcp.connection_renamed` (T30): `CREATE OR REPLACE FUNCTION
  audit.capture_mcp_connections()` adds the branch `TG_NAME = 'audit_capture_name'` inserting `(action, outcome='success',
  organization_id, resource_type='mcp_connection', resource_id=NEW.id, subject_user_id=NEW.user_id)`; new trigger
  `audit_capture_name AFTER UPDATE OF display_name ON app.mcp_connections FOR EACH ROW WHEN (OLD.display_name IS DISTINCT
  FROM NEW.display_name)` (strict: an audit fault refuses the rename); `audit.events_attribute()` replaced with the action
  in its trigger list and the `mcp.*` shape rule (`group_id IS NULL`, `reason IS NULL`; the current body is 00018's
  `core/migrations/00018_platform_signup.sql:829`); `events_action_valid` (00018:1049) widened by that one value. **No name,
  prior value, new value or reason enters the row** (a clear and a set are both "renamed"). No new reason, operation or
  resource type; the audit reason/operation counts (79/73) are unchanged.
- **Down** (development/test only; refused in production by the runner): refuses while any `display_name IS NOT NULL` or
  any `mcp.connection_renamed` event exists (inspected with FORCE lifted inside the transaction, as 00018 does); then
  drops the trigger, restores the version-18 bodies of both audit functions, narrows `events_action_valid`, revokes the
  two column grants and drops the column. Postconditions pin the exact version-19 inventories.
- **Locks/runtime:** brief ACCESS EXCLUSIVE on `app.mcp_connections` (nullable column, no rewrite) and `audit.events`
  (one validating scan of the swapped CHECK). Migrate, then deploy: the version-18 code never reads the column; the
  version-19 code requires version 19 (startup version check as for 00011 in MCP.md). Rollback story: a down is possible
  only on an unnamed dataset; otherwise names are retained — they are user-supplied presentation metadata on the tenant-
  bound row, readable only by their owner under `mcp_connections_select`, never listed once the row is terminal, and
  untouched by retention (no retention grant on this table). No PII claim beyond what the user typed; §7 documents it.

## 4. Core: validation, create, rename

### 4.1 `identity.NormalizeMCPDisplayName(raw string) (name *string, err error)` — authoritative (D-6 order)

Precondition: `raw` is a Go string produced by the strict wire decoding of §4.3 (the wire layer, not this function, is the
proof that malformed UTF-8 and unpaired surrogates were refused rather than replaced). Steps, in this order:

1. `utf8.ValidString(raw)` false → `ErrInvalidMCPName` (defense in depth behind §4.3; a replaced U+FFFD cannot reach here
   from the wire because §4.3 refuses the raw bytes first).
2. **Reject before trimming:** any rune with `unicode.IsControl` (Cc: C0 incl. `\t \n \v \f \r`, DEL, C1 incl. U+0085) or
   `unicode.Is(unicode.Zl, r) || unicode.Is(unicode.Zp, r)` (U+2028/U+2029) **anywhere** → `ErrInvalidMCPName` (never
   sanitized, never trimmed away).
3. **Trim** `strings.TrimFunc(raw, unicode.IsSpace)` (Unicode White_Space). Because step 2 already refused the Cc/Zl/Zp
   members of White_Space, the **effective trim set** is exactly `SP U+00A0 U+1680 U+2000–U+200A U+202F U+205F U+3000`;
   U+FEFF is not White_Space, is not trimmed and counts as a character. Pinned by the vector table (§6).
4. Empty after trimming → `nil` (clear / unnamed). Otherwise `utf8.RuneCountInString` must be 1..80 → `&name`;
   81+ → `ErrInvalidMCPName`. Markup is kept verbatim (never interpreted anywhere; §5.1 escapes on render).
5. Duplicates are allowed; nothing is derived from client ID, URL, user or provider.

Web (`normalizeDisplayNameInput`, §5.3) mirrors the same order with explicit sets (`/[\p{Cc}\p{Zl}\p{Zp}]/u` first, then
the explicit effective trim set as a regex, then `Array.from(s).length` 1..80); JS `trim()` is not used (it differs on
U+FEFF). Core remains authoritative; the web pre-check only avoids a core call.

### 4.2 Store and facade

- **Create:** `InsertMCPConnectionTx(..., displayName *string)`; `sqlMCPInsertConnection` adds `display_name`. Facade
  `CreateMCPConnection(ctx, tc, id, host, resource, lifetimes, displayName *string)`: unchanged transaction (browser
  session lock, facts with lock, membership check, 'mcp' session + row). Setup `Create` decodes an optional
  `display_name` (§4.3; absent → `nil`; invalid → `invalid_name` **before** `begin`, so nothing is created and no
  session/CSRF detail is disclosed to a malformed name; order otherwise unchanged).
- **List:** `sqlMCPListFrom` adds `c.display_name`; `MCPConnectionInfo.DisplayName *string`; `connectionView` emits
  `"display_name": string|null` always. Cursor, page size, ordering and filters are byte-for-byte unchanged.
- **Rename — lock order (r1 C2, replaces r0's joined `FOR NO KEY UPDATE OF c FOR SHARE OF s` statement and its §10
  fallback).** Facade `RenameMCPConnection(ctx, tc, connectionID string, name *string) error`, Organization-only `tc`,
  one TENANT-ORG write transaction (`inContextTx(write)`), in this statement order:
  1. `s.id.LockLiveSession(ctx, tx, sc.session, sc.subjectText)` — the **live user** (`sqlLockLiveUser`, users FOR SHARE)
     then the **browser session** (`sqlLockLiveSession`, sessions FOR SHARE), exactly as `CreateMCPConnection`
     (`access/mcp.go:574`) and as identity documents the order (`identity/store.go:589-594`: user first, then session,
     the order of DisableUser's users FOR UPDATE → sessions UPDATE).
  2. Store `ReadMCPConnectionSessionTx(ctx, tx, userID, org, id) (sessionID string, err)`: a **plain** read
     `SELECT session_id::text FROM app.mcp_connections WHERE id = $1 AND user_id = $2 AND organization_id = $3` (own-row
     by predicate and by `mcp_connections_select`); no row → `ErrUnauthenticated` → facade `fail(CodeNotFound,
     reasonTargetNotFound)` (`access/errors.go:22`). `session_id` is immutable (guard), so this unlocked read cannot be
     stale for the next step.
  3. `s.id.LockLiveSession(ctx, tx, sessionID, sc.subjectText)` — the **target 'mcp' session** locked FOR SHARE (the users
     FOR SHARE is re-taken on the same row, a no-op). This is the locking read of the target session: it rereads
     `revoked_at IS NULL`, both expiries at the service clock and the owner's `status = 'active'` under the lock. A dead
     target → facade `fail(CodeNotFound, reasonTargetNotFound)` (not `CodeUnauthenticated`: the caller's own session is
     live; the target is not one of the caller's live connections, byte-identical to unknown/foreign).
  4. `tenancy.ReadFacts(..., Lock: true)` + `contextCheck` (Organization FOR SHARE, the subject's Organization binding
     FOR SHARE; a non-member → `CodeContextDenied` → route `not_available` 403, D-4).
  5. Store `RenameMCPConnectionTx(ctx, tx, userID, org, id string, name *string, now)`: `SELECT state, pending_expires_at
     FROM app.mcp_connections WHERE id = $1 AND user_id = $2 AND organization_id = $3 FOR NO KEY UPDATE` (the **connection
     row**, last, as `sqlMCPRowLock` is the last lock of M2/M3 at `mcp_store.go:35`); no row, or `state = 'pending' AND
     pending_expires_at <= now` → `ErrUnauthenticated` → `CodeNotFound`; then `UPDATE app.mcp_connections SET display_name
     = $2 WHERE id = $1` (the only write; the guard trigger passes a name-only UPDATE; `audit_capture_name` fires only on a
     distinct value). Same-value rename: the UPDATE runs, the `WHEN` clause skips the audit row, outcome `renamed`.
  - **Why this order is race-safe:** `revokeSession` (`identity/store.go:1047`) takes the 'mcp' session row exclusively
    (`UPDATE app.sessions SET revoked_at …`) and never touches the connection row; the audit capture triggers on
    `app.sessions` (00008:631-635) insert audit rows only. A rename holding the target session FOR SHARE (step 3) blocks
    that UPDATE until the rename commits; a revoke committed first is seen by the locking read in step 3 (locking reads
    see the latest committed version). `DisableUser` (users FOR UPDATE → sessions UPDATE), activation M2 (user → the
    connection's session → facts → row) and selection M3 (same prefix, row last) all take the session before the row and
    the user before the session, so no transaction ever waits on the row while another waits on its session in the
    opposite direction; every lock before the row is FOR SHARE, so two renames, or a rename and an activation/selection,
    serialize only on the row and cannot deadlock. The connection row is never locked before the session.
  - **Fail closed, no fallback:** `LockLiveSession` on the target session is the same pinned statement the runtime already
    executes in the TENANT-ORG shape on the subject's own session rows (`rls_integration_test.go:1335`;
    `migrate/policy_integration_test.go:306`; `sessions_update` admits `user_id = ctx user`, and the 'mcp' session belongs
    to the subject). If, during execution, that lock is refused (42501, policy invisibility) or any other serialization
    mechanism cannot be acquired, the rename fails with `CodeUnavailable` → `mcp_unavailable` and **the writer stops the
    rename slice and returns the finding to Codex for review**; it is a stop, not an implementation deviation, and no
    connection-only lock, re-read or weaker ordering is substituted (r0 §10 fallback withdrawn).
  - The facade touches no column other than `display_name`: UUID, URL/resource, host, user/Organization/client binding,
    selection, context version, activation, session expiry/renewal are unchanged by construction (§6 asserts
    before/after). It logs `access.decision` op `mcp.connection.rename`.

### 4.3 Route `POST /internal/v1/mcp/connections/rename` and strict name-field decoding (r1 C4)

Body `{"origin","connection_id","csrf_token","display_name"}`; `display_name` required on rename, optional on create;
`""` (or whitespace-only) clears; unknown fields refused; 16 KiB body bound as today (`decodeSetup`, `connections.go:119`).
Go's `encoding/json` replaces invalid UTF-8 and invalid surrogate escapes with U+FFFD instead of failing, so the name
field is decoded as `json.RawMessage` and validated on its raw token before any string exists:

- **Presence/duplicates/null (both endpoints):** `nameField(body []byte)` is a depth-1 `json.Decoder.Token()` scan of the
  already-bounded body counting the `display_name` key: absent → create: no name / rename: `invalid_request`; **more than
  one `display_name` key → `invalid_request`** (both endpoints); `null` → `invalid_request` (a name is a JSON string; clear
  is `""`); any non-string token → `invalid_request`. **Every other field keeps `decodeSetup`'s existing behavior
  exactly** (last-wins duplicates, replacement decoding of `origin`/`csrf_token`/`connection_id`, which then fail their
  own exact checks as today); the scan reads only the one key.
- **Encoding (both endpoints):** the raw string token must be `utf8.Valid` **and** every `\u` escape in the surrogate
  range D800–DFFF must be a high surrogate immediately followed by a `\u` low surrogate (case-insensitive hex); a
  malformed byte sequence or a lone/unpaired/reversed surrogate escape → **`invalid_name` 400**, refused, never replaced.
  Only then `json.Unmarshal(raw, &s)` and §4.1 (`ErrInvalidMCPName` → `invalid_name`).
- **Order:** media type → body (`decodeSetup`) → `nameField` → encoding + §4.1 → single `Authorization` → origin registry
  → canonical UUID (rename) → session bound to origin/Organization → CSRF → facade. The name is settled **before
  `begin`** on both endpoints: a malformed name learns nothing about session, CSRF or target.
- **Answers:** success `200 {"outcome":"renamed"}` (create: unchanged `201` shape). Errors: `invalid_request`,
  `ambiguous_parameter` 400; `invalid_name` 400 (D-2); `unsupported_media_type` 415; `origin_rejected`, `csrf_failed`,
  `not_available` (not a current member, D-4) 403; `unauthenticated` 401 (the **caller's** browser session: absent, dead,
  wrong origin/Organization, disabled owner, recovery-invalidated); `not_found` 404 for unknown, foreign-user,
  foreign-tenant and terminal **targets** alike (byte-identical body); `mcp_unavailable` 503. `mcp.setup` event
  `operation=rename` with the closed outcome/reason only. No GET; existing `create`/`revoke` contracts, their membership
  rules (revoke still needs no membership) and cache policy are untouched.

## 5. Web: shared component, pages, routes

### 5.1 `web/components/account/connection-manager.tsx` (server component; `import "server-only"` via `mcp-client`)

Props: `{ ctx, credential, origin, cursor?: string, notice?: string, page: "account" | "connections", ids: { list, create,
listTitle, createTitle }, params: { cursor, status }, csrfToken?: string }`. Renders the list `Section` (heading, intro
text, `StatusMessage`, rows, page nav) and the create `Section` (name `Input` + submit). Each row: `data-testid="mcp-name"`
heading text = `displayName ?? "Unnamed connection"` (escaped React text, `[overflow-wrap:anywhere]`), then the existing
`dl` (URL, Address, Status, **Bound client** `clientId ?? "Not yet bound"` — kept separate from the name, Created,
Activated), a rename form (`display_name` text `Input` with visible label "Name", `aria-describedby` hint "Up to 80
characters; leave empty to remove the name", current value prefilled, **no `maxlength`** (UTF-16 units are not the limit),
hidden `csrf_token`, `connection_id`, `return_to`; submit "Save name") and the revoke form (unchanged + `return_to`).
Pagination links are built from `page`: `/connections/mcp?cursor=…` or `/account?mcp_cursor=…#mcp-connections`; "Newest
connections" links to the bare page path. States inside the component (both pages): `ok`, `empty`, `invalid_page`
(closed message + newest link), `not_available` (closed message), `unavailable` (closed message, `data-testid=
"mcp-unavailable"`, for `mcp_unavailable`, `timeout`, `unreachable` and `malformed` — the unconfigured-MCP 404 reduces to
`malformed` today and is shown as unavailable; no create/rename/revoke form is rendered in this state). No `cookies()`,
no client component, no new primitive: existing `Section`, `RuledGrid`, `Heading`, `Text`, `Input`, `Button`,
`PendingSubmit`, `StatusMessage` and semantic tokens; Light/Dark/System untouched. The component makes exactly one core
call (the list) and catches nothing: `listConnections` already reduces every failure to a closed `McpResult`.

### 5.2 Pages and the shared `AccountSections` composition (r1 C5/D-5)

- `web/components/account/account-sections.tsx` (**new**, server component, no `cookies()`/`headers()`): the signed-in body
  of `/account` extracted verbatim from `web/app/account/page.tsx:97-180` — the Session `dl`, retirement notice, management
  links, the `ConnectionManager` (page `"account"`, ids `mcp-connections`/`mcp-create`) and the Sign-out `Section` with its
  three forms — taking `{ value: SessionView, notice, soleOwned, manages, mcp: { ctx, credential, origin, cursor, status } }`.
  `/account` keeps all its data fetching, redirects, `notFound`, branding, retirement and `NAV` logic and renders
  `AccountSections` when `result.ok`; the not-ok branch is unchanged. NAV gains "AI connections" / "New connection".
  Account `status` notices stay the existing closed table; `mcp_status` has its own closed table (`created`, `revoked`,
  `renamed`, `invalid_name`); an unknown or duplicated value shows nothing. A failed list never breaks Session/Sign-out:
  the component renders its bounded `unavailable`/`not_available` state in its own section (proven rendered, §6
  S-ACCOUNT-MCP-UNAVAILABLE). The page's existing CSRF token is passed through, so the component's one list call is the
  only extra core call.
- `/connections/mcp` (compatibility): keeps its whole-page `SignInFirst`, `NotAvailable` and `InvalidPage` branches, NAV,
  ids `connections`/`create`, params `cursor`/`status` and the `STATUS` table (+ `renamed`, `invalid_name`); the body is
  the component with `page: "connections"`. Existing M-* test ids are preserved.
- `web/tests/fixtures/account-mcp/page.tsx` (**new**, test-only; copied to `app/account-fixtures/page.tsx` only by the
  browser image's `fixture` stage, exactly like `token-scopes` at `infra/docker/web-browser.Dockerfile:11-12`): renders
  `AppShell` + `AccountSections` with a scripted `SessionView` (fixed display name, times, a dummy CSRF token) and real
  `ConnectionManager` props (`credential: "fixture"`, the fixture's own origin). The fixture build's
  `CORE_INTERNAL_URL=http://127.0.0.1:9` (`web/playwright.config.ts:49`) makes the real `listConnections` call fail
  `unreachable`, so the page exercises the production fallback path, not a stub. Production returns 404 for the route
  (asserted by the new browser case; S-NOFIXTURE's pattern).

### 5.3 `web/lib/mcp-client.ts`

- `ConnectionView.displayName: string | null`; row validation: key present; `null`, or a well-formed string that passes
  `isDisplayName` (1..80 code points via `Array.from(s).length`, no `/[\p{Cc}\p{Zl}\p{Zp}]/u`, equal to its trimmed form
  under the §4.1 effective trim set); anything else → `malformed`.
- `createConnection({ origin, csrfToken, displayName? })` sends `display_name` only when given; `renameConnection({ origin,
  connectionId, csrfToken, displayName })` → `McpResult<"renamed">` (exactly `{"outcome":"renamed"}`); `MCP_ERRORS` +=
  `invalid_name`; `normalizeDisplayNameInput(raw): { ok: true; value: string } | { ok: false }` (§4.1 order and sets; agrees
  with core on the shared vector table in §6; core remains authoritative).
- `strictFormField(request: Request, name: string, maxBytes = 16_384)` (server-only helper, r1 C4): reads the raw body
  bytes of a cloned request within the same 16 KiB budget as `readForm`, splits `application/x-www-form-urlencoded` pairs
  on `&`, percent-decodes key and value **as bytes** (`+` → 0x20), selects the pairs whose key equals `name`, and returns
  `{ kind: "absent" }`, `{ kind: "ambiguous" }` (more than one), `{ kind: "malformed" }` (`TextDecoder("utf-8", { fatal:
  true })` throws: invalid UTF-8, overlong forms and UTF-8-encoded surrogates are all refused, never replaced) or
  `{ kind: "value", value }`. A lone surrogate cannot arise from a fatal UTF-8 decode; `isWellFormed()` stays in
  `normalizeDisplayNameInput` as belt-and-braces.
- **Bound recompute (R5):** per row add `"display_name":` (15 bytes) + separator (1) + value: worst case 80 code points
  each `<`, `>` or `&` escaped by Go to `<` (6 bytes) = 480, plus two quotes = 482 (a `null` is 4; non-BMP is 4 bytes
  per code point = 320; control/invalid never leave core). `LIST_ITEM_MAX_BYTES = 12_288 + 310 + 171 + 498 = 13_267`;
  `MCP_LIST_MAX_RESPONSE_BYTES = 222 + 1 + 99 + 100 × 13_267`. The exact figure is verified by the extended "largest answer
  / one byte more" unit test against Go's encoder behavior (sorted keys: `display_name` sits between `created_at` and `state`).

### 5.4 Form routes (`create`, `revoke`, new `rename`) (r1 C3/C4: no raw byte pre-check; normalize first, then bound)

Existing order kept (method → `readForm` encoding/exact Origin/16 KiB bound → fields → cookie → core). The route calls
`request.clone()` before `readForm`, so `readForm` and every legacy field (`csrf_token`, `connection_id`, `return_to`)
keep their current lossy decoding and `field()` ambiguity rules unchanged, while the **name field alone** is read from the
clone's raw bytes by `strictFormField` (§5.3): `malformed` → redirect with `invalid_name` (no core call);
`ambiguous` → 400; `absent` → create: no name / rename: 400. The accepted string then goes through
`normalizeDisplayNameInput` (reject Cc/Zl/Zp → trim the effective set → bound 1..80 **code points of the normalized
value**; a value that normalizes to empty is a clear on rename and "no name" on create); `{ ok: false }` → redirect
`invalid_name` without calling core; core still decides. **The r0 `Buffer.byteLength ≤ 320` pre-normalization check is
removed**: a valid 80-code-point name surrounded by any amount of trim whitespace within the 16 KiB form budget is
accepted (vector in §6). New hidden `return_to` with the closed set `account` | `connections`; absent → `connections`
(existing behavior); any other value or a duplicate → 400. Redirect map: `connections` → `/connections/mcp?status=<code>`;
`account` → `/account?mcp_status=<code>#mcp-connections`. Codes: `created`, `revoked`, `renamed`, `invalid_name`. The raw
name is never placed in a query string, redirect, log or status page. `rename`: `connection_id` canonical UUID;
`not_found` → the existing 404 status page wording ("not one of your active connections here"); `unauthenticated` → the
existing 401 "Not signed in" page; `invalid_name` from core → redirect `invalid_name`. GET/PUT/PATCH/DELETE/OPTIONS → 405
as today.

## 6. Tests and validation (reuse / extend / new, with the gap each new item covers)

**Shared vector table** (core `TestNormalizeMCPDisplayName` and web `normalizeDisplayNameInput` run the same list):
80 BMP ok; 80 non-BMP (80× U+1F600) ok; **padded non-BMP (r1 C3): `U+3000×8 + U+1F600×80 + U+00A0 U+2000 SP` → ok, stored
value = the 80 emoji**; 81 refused; blank/whitespace-only → clear; outer NBSP/U+3000/U+2000–U+200A/U+202F/U+205F trimmed;
inner spaces kept; U+FEFF kept and counted; `\n`/`\t`/`\v`/`\f`/`\r`/NUL/U+0085/U+2028/U+2029/DEL/U+009F refused **even when
only leading/trailing** (D-6: rejected before trimming, never trimmed away); `<script>` kept verbatim.

| Selector / file | Kind | Covers |
|---|---|---|
| `core/internal/identity/mcp_name_test.go` `TestNormalizeMCPDisplayName` | **new** (no existing name/Unicode boundary) | the shared vector table (a Go-string test of the normalizer only; the wire proof is below) |
| `core/internal/mcp/mcp_test.go` `TestSetupNameDecoding` (**new function in the existing file**, the `testHandler`/`serve` harness of `TestTransportRefusalsBeforeTheStore`, `mcp_test.go:232`) | extend | **endpoint-level wire negatives (r1 C4)** on both `POST /internal/v1/mcp/connections` and `/rename` with raw body bytes: `"display_name":"\xff"` and `"\xc0\xaf"` (invalid UTF-8) → 400 `invalid_name`; `"\ud800"`, `"\udc00"`, `"\udc00\ud800"` (lone/reversed surrogate escapes) → 400 `invalid_name`; a valid pair `"😀"` → accepted (reaches `begin`, which the stub refuses with `unauthenticated`); `"display_name":null` → 400 `invalid_request`; two `display_name` keys → 400 `invalid_request`; non-string → `invalid_request`; rename without `display_name` → `invalid_request`; create without it → proceeds; **two `origin` keys still last-wins (legacy behavior pinned)**; every refusal answered before any session/CSRF check (the stub verifier records no call); malformed UUID → `invalid_request` |
| `core/internal/mcp/mcp_integration_test.go` IT-MCP-3 (extend) and new `IT-MCP-15 rename` under `TestMCPLifecycle` | extend | create with/without name (legacy request shape still valid; `null` in list); rename/clear pending and active; the padded non-BMP vector over real HTTP (stored as 80 code points, listed back byte-identically); same-value idempotent (no audit row); dave (member of both Organizations) renaming through the other Organization's host and alice renaming bob's row → `not_found` byte-identical; **target-layer terminal classes with a live caller session** (r1 C5): own-revoked (setup revoke and logout tool), expired-pending, idle-/absolute-expired 'mcp' session (clock advance) → `not_found`, no change; the wire negatives of `TestSetupNameDecoding` repeated over the real server with no row change; CSRF/origin/method/duplicate-header refusals; **before/after fact comparison** of id, resource, host, user/Organization/client, selected Group, context version, activated_at, session expiries/renewed_at; audit row shape of `mcp.connection_renamed`; `mcp.setup` `rename` events catalogued; old-host historical row (mcp-seed) renamable through the current host (alias regression) |
| `core/internal/mcp/mcp_integration_test.go` `TestMCPRecovery` (`:1425`) | extend | **browser-auth layer (r1 C5):** after the bound G013 invalidation SQL, a rename of the pending and the active connection with the pre-invalidation browser credential → 401 `unauthenticated` (the caller's session is invalidated before any target lookup; no change); after the explicit re-sign-in the rows are absent/terminal → `not_found` |
| `core/internal/identity/mcp_integration_test.go` `TestMCPIdentityRename` (new function, existing file; `tenantTx` harness of `:63`) | extend | **store/liveness layer (r1 C5):** in a TENANT-ORG test transaction, `LockLiveSession(target)` + `RenameMCPConnectionTx`: live pending/active → renamed; owner set `disabled` by the superuser → `ErrUnauthenticated`, name unchanged (the only layer that can reach the disabled-owner class, since the browser session dies with the owner); 'mcp' session expired by the G013 SQL → `ErrUnauthenticated`; revoked → `ErrUnauthenticated`; foreign Organization shape → no row; the facade-level disabled-owner case is `unauthenticated` at `begin` (IT-MCP-15 asserts that layer) |
| `core/internal/tenancy/access/mcp_integration_test.go` `TestMCPOrdering` O-MCP-9 (`pauseTracer.arm`, `waitWaiters`, `isDeadlock` of `ordering_integration_test.go:121,156`; `integration_test.go:975`) | extend | **both orderings and no inverse-order deadlock (r1 C2):** (a) *rename-first*: pause the rename after its connection-row `FOR NO KEY UPDATE` statement (the target session FOR SHARE is held); start `RevokeOwnMCPConnection` → `waitWaiters(1)` (the revoke's `UPDATE app.sessions` waits); release → rename commits `renamed`, revoke then commits; final: name set, session revoked, no 40P01; (b) *revoke-first*: revoke committed, then rename → `CodeNotFound`, name unchanged; (c) *revoke held open* (a superuser transaction holding `UPDATE app.sessions SET revoked_at …` uncommitted, the pattern of `retirement_integration_test.go:337`): start rename → it waits at the target session FOR SHARE (`waitWaiters(1)`), not at the row, and no deadlock is reported; release → rename observes the revocation → `CodeNotFound`; (d) a rename never revives or extends a session (expiries/renewed_at equal before/after in every case); statement order recorded (`ordering(...)` line) and asserted: user, browser session, target session, facts, row |
| `core/internal/tenancy/access/mcp_test.go` `TestMCPScopeBinding` | extend | rename refused with a Group-selected `tc` and with `s.mcp == nil` (scope mismatch) |
| `core/internal/tenancy/access/boundary_test.go` `TestAccessBoundary`, `TestRLSBoundary` (r1 C1) | extend (pins) | exported facade method inventory includes `Service.RenameMCPConnection`; identity user-taking functions include the two new store helpers; `identityMCPProtectedFuncs` unchanged (asserted: the new helpers name no protected table) or extended as `callee` with the reason recorded; no SQL literal in access files; `liveSession` still reached only through the two liveness methods |
| `core/internal/tenancy/access/audit_integration_test.go` `TestAudit` AT-INV, AT-ATOM (r1 C1/C5) | extend | AT-INV observes `mcp.connection_renamed` from a real `RenameMCPConnection` (the completeness loop over the 47 actions passes); **rename audit-fault rollback**: under `w.fault(t, "false")` (`:191`) `RenameMCPConnection` fails, `display_name` stays NULL, no `mcp.connection_renamed` row, the connection row's `xmin` and every security fact (state, client, selection, context version, session `revoked_at`/expiries/`renewed_at`) are unchanged; after `remove()` the rename succeeds and the event's `xmin` equals the connection row's (one transaction, the AT-ATOM M1 pattern at `:531-545`) |
| `core/internal/audit/audit_test.go` (`:102`) | extend (pin) | 40 trigger-owned + 7 core = 47 |
| `core/internal/migrate/mcp_integration_test.go` `TestMCPNamesCatalog` (CT-MCPN-1, CT-MCPN-NEG) / `TestMCPNamesMigration` (MG-MCPN-1..4) | extend (existing B-MCP-CATALOG file; selector regex widened) | exact version-19 delta over the version-18 catalog (column, CHECK, two column grants, trigger, two function bodies, widened action CHECK; policies/guard digests unchanged); detectors: missing CHECK, UPDATE grant on another column, NO FORCE, dropped trigger, a definer capture; MG: up over a populated version-18 world with unnamed rows (all `NULL`, every other column identical, `mcp.connection_established`/`context_changed` events untouched), down refused with a named row or a rename event, down/up on an unnamed world restores the exact 18/19 catalogs, rolled-back up restores the owner and version 18; runtime cannot `UPDATE` any other column via the name path (negative DML) |
| `web/tests/unit/mcp-connections.test.mjs` | extend | `display_name` row validation (legacy `null`, hostile markup accepted as text, 81 code points / control chars → `malformed`), recomputed largest-answer bound (one byte more refused), create sends the name only when given, rename request/answer, `invalid_name` passthrough, `normalizeDisplayNameInput` and `strictFormField` on the shared vector table **plus the padded non-BMP vector and raw-byte vectors** (`%FF`, `%C0%AF`, `%ED%A0%80` (encoded surrogate) → `malformed`; two keys → `ambiguous`; `+` → space) |
| `web/tests/unit/mcp-forms.test.mjs` (**new**; the `signup-forms.test.mjs` route-import harness, `:13-25`) | **new** (no existing unit test exercises the MCP form routes) | **endpoint-level form negatives (r1 C4):** raw `Uint8Array` bodies to `create` and `rename`: `display_name=%FF%FE` and `display_name=%ED%A0%80` → 303 to the page's `invalid_name` status with **no fetch to core**; duplicate `display_name` → 400; rename without it → 400; create without it → outbound JSON has no `display_name`; padded non-BMP → outbound `display_name` is the 80 emoji; `return_to` closed set and omission default, duplicate/omitted legacy fields unchanged (400 / last value as today), method restrictions, no name in any redirect/status page, Origin refusal through `readForm`, `not_found`/`unauthenticated` page mapping |
| `web/tests/browser/account.spec.ts` **S-ACCOUNT-MCP-UNAVAILABLE** (new case; `shell-light`/`shell-dark`; B-BROWSER-ROOT) | extend | **rendered `/account` + shared-component failure (D-5):** `http://127.0.0.1:3001/account-fixtures` (the fixture build with an unreachable core, the existing absolute-URL pattern of S-FAILCLOSED/S-LONG-IDENTITY) renders the Session `dl` (scripted values), the three Sign-out forms with their hidden `csrf_token`, the NAV with "AI connections", and the connections section in its closed `unavailable` state (`mcp-unavailable` visible; no create/rename/revoke form; no `connection_id`, credential or core URL in the HTML); axe clean at 1280 and 320 in both appearances; keyboard reaches Sign out from the header; production `/account-fixtures` → 404 (`request.get` on `baseURL`) |
| `web/tests/browser/mcp.spec.ts` (projects `mcp-light`, `mcp-dark`; B-IDENTITY-E2E) | extend | M-SETUP: create with a name and a legacy unnamed create ("Unnamed connection"); **M-RENAME** (new case): rename, clear, duplicate names, the padded 80-non-BMP vector accepted and shown as 80 characters, hostile markup rendered as text, control/81-code-point input → closed `invalid_name` notice, nothing echoed, a revoked target → the 404 page wording; **M-ACCOUNT** (new case): the same list/create/rename/revoke flows from `/account`, Session and Sign-out sections intact, `mcp_status` disjoint from `status`, pagination links carry `mcp_cursor` and return to `/account`, `return_to` round trip, an invalid `mcp_cursor` renders the component's `invalid_page` state while Session/Sign-out stay intact (the second rendered isolation proof, on the real stack), keyboard reaches each form from the header, focus/status announcements; M-A11Y: add the `/account` embedding at 1280 and 320 in both appearances (axe, reflow); M-LIST unchanged (name column added) |
| Existing unchanged controls (reuse, cited, not rerun pre-merge unless affected) | reuse | `B-MCP-SOURCE`, `B-WEB-BOUNDARY` (pins updated), `B-MCP-HTTP` (unconfigured stack: the new route must also 404), `B-IDENTITY-GUARDS`, `B-MCP-E2E` `TestMCPE2E` (only if the M-HANDOFF payload changes — it does not), B-RLS-*, B-AUDIT-CATALOG (`TestAuditCatalog`; AUDIT.md count pin), B-MIGRATE-SET/CLI/CATALOG (version-19 texts), S-NOFIXTURE (unchanged; the new fixture route has its own 404 assertion) |

**Local batches (focused selectors; verify what each actually runs before dispatch; fetch/rebase + Docker Desktop mount
check before each batch):**

1. *Static/unit, stack-free, no harness lock:* `bash infra/validation/run.sh lint lint-base lint-web lint-core unit-web
   unit-core` (B-MIGRATE-SET, B-SECRET-CANARY, B-WEB-BOUNDARY with its updated pins, B-MCP-SOURCE, B-UNIT-WEB = README
   `web-check` lines (lint/type/unit incl. both `.test.mjs` files), B-UNIT-CORE = README `core-check` (gofmt, vet, unit
   tests: **this is where `TestAccessBoundary`, `TestProductionImporters`, `TestRLSBoundary`, `TestSetupNameDecoding`,
   `TestNormalizeMCPDisplayName` and the `audit_test.go` count run** — they carry no integration tag). Suites are the
   smallest existing selectors containing the affected IDs; no umbrella profile.
2. *Private database runs (pool of two, plain native Docker, read-only source, disposable DB):* `go test -count=1 -run
   '^(TestMCPCatalog|TestMCPMigration|TestMCPNamesCatalog|TestMCPNamesMigration|TestAuditCatalog)$' ./internal/migrate/`;
   `-run '^(TestMCPTransport|TestMCPLifecycle|TestMCPAudit|TestMCPRecovery)$' ./internal/mcp/`; `-run '^TestMCPIdentity'
   ./internal/identity/`; `-run '^(TestMCPOrdering|TestMCPFootprint)$' ./internal/tenancy/access/` (B-MCP-IT's facade
   invocation, `run.sh:3584`); **`-run '^TestAudit$' ./internal/tenancy/access/` with B-AUDIT-IT's stored-object
   environment (`rls_go audit-it … "$OBJ_ENV"`, `run.sh:3384`: the private MinIO credential; one pool slot; the whole
   `TestAudit` function is the smallest selector containing AT-INV/AT-ATOM)**; `-run '^TestSignup(Catalog|Migration)$'` only
   if their head-version pins are touched. The whole `migrations` suite is not run pre-merge (technical reason: it is an
   umbrella of ~30 unaffected catalogs; it runs in the post-merge full qualification).
3. *Browser:* (a) the `IDENTITY.md identity:browser` command restricted to `--project mcp-light --project mcp-dark --grep
   "M-"` (private HTTPS identity environment, one pool slot); (b) the README `platform:browser` command restricted to
   `--project shell-light --project shell-dark --grep "S-ACCOUNT"` (the root stack + the fixture servers; one pool slot;
   the browser image rebuild for the new fixture route takes the single build slot). The full B-IDENTITY-E2E (14 projects)
   and B-BROWSER-ROOT run post-merge.
4. *Hosted:* MR fast pipeline automatic; changed paths (`web/tests/browser/`, `web/tests/fixtures/`, `core/migrations/`,
   `web/app/account/`, `web/components/account/`, `web/lib/`, `contracts/`, `docs/`, `infra/docker/`) map to `foundation`
   → the optional MR full gate stays manual; the coordinator-owned local full qualification runs on the exact merged
   staging SHA (frozen checkout).

**Manual/operator feedback:** none required by Goal/Spec (no new owner gate); writer self-review of component reuse and
authority separation plus Codex source review are the required human-equivalent reviews. The r0 D-5 evidence limit is
withdrawn: the unavailable state is rendered and asserted in the browser (S-ACCOUNT-MCP-UNAVAILABLE) on the existing
fixture-build harness, and `invalid_page` isolation is rendered on the real identity stack (M-ACCOUNT).

## 7. Contracts and documentation deltas

- `contracts/mcp/control-plane.md` "Internal setup contract": list row `display_name` (string or null; legacy rows null),
  create `display_name?` with its strict decoding rules (string only; `null`/duplicate/non-string refused; invalid UTF-8 or
  unpaired surrogate escapes refused as `invalid_name`), the `rename` route row, `invalid_name` error row, `mcp.setup`
  `operation` `rename`, the D-6 validation order; everything under "Connections and URLs", "Access tokens", "Transport",
  "Tools" and "Discovery" unchanged (names never enter MCP payloads or binding reads).
- `docs/foundation/MCP.md`: setup flow (My Account + compatibility page), names as presentation metadata, the rename lock
  order (§4.2), Components/Evidence rows; `AUDIT.md`: T30 row, counts 46 → 47 / 39 → 40 trigger-owned, capture trigger
  count 20 → 21; `DATABASE-ISOLATION.md`: "Version 19" paragraph (column, grants, unchanged policies); `MIGRATIONS.md`:
  row 19; `WEB-SHELL.md`: `/account` sections, the `rename` route in the route table, the second fixture route
  `/account-fixtures` in the fixture-build paragraph (`:330`) and S-ACCOUNT-MCP-UNAVAILABLE in the shell project row
  (`:342`); `docs/SECURITY.md` §5 runtime row (two more column privileges). `connection-context.md` unchanged.

## 8. Sequence (one writer)

S0 reconcile bound inputs at the release SHA (exact-delta rule for irrelevant staging changes) and re-check the 00019 slot
(D-3) → S1 migration + migrate tests (highest-risk slice first; prove CT-MCPN/MG-MCPN) → S2 identity/facade/route + strict
decoding + core tests + boundary/audit pins + `main.go` + events.json; **prove O-MCP-9 (both orderings) and the target-
session lock acquisition early** (the fail-closed stop of §4.2 is checked here, before any web work) → S3 web client,
components, pages, routes, fixture page + unit tests → S4 docs/contracts/harness pins (incl. B-WEB-BOUNDARY, Dockerfile)
→ S5 batches 1–3 (finish, fix, rerun failed + affected, then advance) → S6 self-review (component reuse, authority
separation, no name leakage, no replaced-character acceptance) → S7 fetch/rebase/reconcile, safekeeping push, Result →
Codex source review → push/MR (auto-merge after gates) → merged-SHA local full qualification → acceptance cleanup receipt.

## 9. Delegation, ownership, resources

One implementation writer (Opus 5.5 high; isolated worktree `g063-impl`). Optional, at the writer's request only: one
read-only Fable 5.1 high security/SQL review of §3–§4 after S2 (no edits, no second writer). Resources: one build slot
(the browser fixture image rebuild and the core/web check images are sequential), at most two private test runs, unique
run labels, cleanup receipt; the frozen qualification checkout is never rebased.
Scope allow-list: §"Scope allow-list" of the input inventory (not included).

## 10. Recorded decisions (binding, per the Codex r0 review), risks, escalations

| Decision | Ruling (Codex, 2026-10-21) | Where applied |
|---|---|---|
| **D-1** rename audit action | **Approved:** trigger-owned `mcp.connection_renamed`, no name payload; log-only evidence rejected | §3 audit; §6 AT-INV, AT-ATOM rename fault, IT-MCP-15 row shape, `TestAuditCatalog`/`audit_test.go` counts |
| **D-2** closed `invalid_name` | **Approved** for create and rename (bounded, non-echoing, accessible closed error) | §4.3, §5.4, events.json, control-plane.md |
| **D-3** migration slot | **Approved `00019`, subject to the slot still being free at execution release.** At `fb9a8ba3` the latest migration is `00018_platform_signup.sql`. The writer re-checks at S0 and the supervisor records it in the release; if taken, G063 stops and returns for a new ruling. **G020 reconciliation input:** G020 must rebind its migration slot and head-version pins after G063 claims 00019; this Plan does not edit any G020 artifact | §0 gates, §8 S0 |
| **D-4** rename membership | **Approved:** rename requires current Organization membership through a current host of that Organization (`contextCheck` on locked facts, matching create/list; `mcp_connections_update` already requires a current role); the exact-own revoke (`RevokeOwnMCPConnection`, `identity/mcp.go:203`) stays membership-independent and unchanged | §3 RLS, §4.2 step 4, §4.3 `not_available`, IT-MCP-15 negatives |
| **D-5** unavailable-state evidence | **Required:** a bounded rendered `/account`/shared-component failure test in the existing harness; no separate stack | §5.2 `AccountSections` + test-only fixture page (existing fixture-build pattern), §6 S-ACCOUNT-MCP-UNAVAILABLE and M-ACCOUNT `invalid_page` isolation |
| **D-6** Unicode whitespace | **Approved:** Unicode White_Space semantics with matched web behavior; invalid UTF-8, Cc (incl. newline and U+0085), Zl/Zp (U+2028/U+2029) are refusals **before** trimming; accepted surrounding White_Space (the effective set `SP U+00A0 U+1680 U+2000–U+200A U+202F U+205F U+3000`) is trimmed; U+FEFF kept | §4.1 order and effective set, §5.3 web mirror, §6 vector table (leading/trailing control vectors are refusals) |

- **Mechanical pins** (run.sh B-MCP-SOURCE route count and `main.go` digest, B-MCP-CATALOG regex/text, B-MIGRATE-* version
  texts, B-AUDIT-CATALOG/B-AUDIT-IT counts, B-WEB-BOUNDARY regex/inventory/text, events.json enums, `audit_test.go` and
  `boundary_test.go` inventories, the Dockerfile `cp` line) are in-scope mechanical corrections under standing authority;
  they are listed, not silently absorbed.
- **Risks:** Go JSON escaping assumptions in the bound recompute (verified by test); the target-session FOR SHARE is the
  already-granted, already-pinned `sqlLockLiveSession` statement — if its acquisition is refused at execution, §4.2's
  fail-closed stop applies (**no fallback; returned for review**); head-version catalog pins in later migrate tests
  (00018) may need digest refresh if they migrate "up all" — allow-listed; `request.clone()` doubles the bounded form body
  in memory for the two MCP routes only (≤ 32 KiB); the second fixture route adds one `cp` to the browser image's fixture
  stage (no production route).
- **Stops:** any need to touch the guard body, a policy, `sessions_insert`, provisioning roles, provider/OAuth, MCP tool
  payloads, the transport binding read, `web/lib/session.ts`'s shared decoding, or a weaker lock order is a scope/security
  stop → Codex, not an autonomous correction.

## 11. Acceptance-criteria mapping

| AC | Evidence (where) |
|---|---|
| AC1 shared manager on `/account` and `/connections/mcp`; account functions intact; MCP unavailable isolated | source review of one component and one `AccountSections` composition in both pages (§5.1–5.2); M-ACCOUNT, M-SETUP/M-LIST (browser, real stack, incl. `invalid_page` isolation); **S-ACCOUNT-MCP-UNAVAILABLE rendered unavailable state with Session/Sign-out intact (D-5)**; unit `unavailable` reduction (§6) |
| AC2 optional create names, rename/clear, duplicates, legacy fallback, separate client display | IT-MCP-3/15, `TestNormalizeMCPDisplayName`, mcp-connections unit (legacy `null`), M-SETUP/M-RENAME (browser: "Unnamed connection" + "Bound client") |
| AC3 trim, 80 code points, non-BMP, control/newline refusal, escaped display, bounded responses | shared vector table in core and web unit tests incl. the padded non-BMP vector; **wire-level refusal of malformed UTF-8 and lone surrogates at both endpoints (`TestSetupNameDecoding`, IT-MCP-15, mcp-forms)**; IT-MCP-15 invalid writes leave no change; largest-answer bound test; M-RENAME hostile markup rendered as text |
| AC4 session/CSRF/host-bound, own-user/current-Organization, non-disclosing negatives | IT-MCP-15 negatives (dave in both tenants, alice vs bob, stale/wrong session, CSRF/Origin, duplicate/malformed/null fields), `TestSetupNameDecoding`, mcp-forms unit, D-4 membership (`not_available`) |
| AC5 metadata-only rename; no revival/renewal; isolation | IT-MCP-15 before/after fact comparison and target-layer terminal classes; `TestMCPRecovery` and IT-MCP-15 browser-auth-layer classes (disabled owner, recovery-invalidated → `unauthenticated`); `TestMCPIdentityRename` store-layer classes; **O-MCP-9 both orderings + no inverse-order deadlock**; **AT-ATOM rename fault rollback**; guard/policies unchanged (CT-MCPN-1) |
| AC6 alias management, stored Address, 100-row cursor semantics, safe page-specific returns/status | IT-MCP-15 old-host row via current host; existing cursor tests (`TestCursor`, IT list cases) rerun; M-ACCOUNT pagination/`mcp_cursor`/`return_to`; mcp-forms `return_to` closed set |
| AC7 additive migration, least privilege/RLS, bounded wire/log/audit | CT-MCPN-1/NEG, MG-MCPN-1..4, `TestAuditCatalog`, `audit_test.go` 47, B-MCP-SOURCE/B-MCP-HTTP/B-WEB-BOUNDARY pins, `TestAccessBoundary`/`TestRLSBoundary` inventories, events.json catalogued `mcp.setup`/`access.decision` records, log inspection for absence of names |
| AC8 focused checks, accessible responsive appearances, reviews, merged-SHA qualification, cleanup | batches 1–3 evidence in the Result; M-A11Y and S-ACCOUNT-MCP-UNAVAILABLE (1280/320, light/dark, keyboard); hosted fast + Codex/Duo dispositions; coordinator full local qualification receipt on the merged SHA; cleanup receipt; final Codex ACCEPT |

## 12. Readiness gates remaining (restated)

Codex Plan ACCEPT of r1 → accepted Plan/inventory imported to the integration branch → G062 integration state per the
owner order and current staging reconciliation → D-3 slot re-check at release → supervisor execution release (attempt 1)
→ implementation → Codex source ACCEPT → MR/auto-merge after gates → merged-SHA local full qualification → cleanup →
Codex final ACCEPT (gates G020 integration/final acceptance and G021). **Nothing in this document starts, builds, tests
or releases anything.**

## Revision history

- r0, 2026-10-21 — initial Plan on `origin/staging` `7d62e8ff` (fast-forwarded from `3c2539ab`; management-only delta
  reflected in §0: new order G062 → G063 → G020 → G057 → G021, and the Spec's 2026-10-21 priority note). Fable 5.1 planner.
  Retrievable at `g063-plan` `bea8b161ab0fbef717909a026487f64e278dbbe6`; Plan SHA-256
  `b62c9935828f7d37b2a3c2fc814f559827b070500f35b34e95f37302927320ed`, inventory SHA-256
  `a20a73e0d1853e07904e9bc44e6ce7e85bc83adca7517944b30a8d905c012761` (as recorded by the Codex r0 review).
- r1, 2026-10-21 — one correction pass per the Codex r0 review (not included)
  (CORRECTION REQUIRED; sha256 `87d874cc…1897`), on base `origin/staging` `fb9a8ba3` (management-only delta since
  `7d62e8ff`; every source blob unchanged). Fable 5.1 planner, worktree `agent-<id>`, branch `g063-plan-r1`.
  Per correction → sections changed:
  1. *Scope and control coverage* → §2 (two new rows: `boundary_test.go`, `audit_integration_test.go`; the four
     `CreateMCPConnection` call sites; B-WEB-BOUNDARY/Dockerfile pin rows), §6 (boundary, `TestAudit`, `audit_test.go`
     rows), batches 1–2 (unit-core contains the boundary tests; `TestAudit` with B-AUDIT-IT's environment), inventory
     allow-list and bound inputs.
  2. *Lock order; fallback removed* → §1.2, §4.2 (new step list user → browser session → target 'mcp' session → facts →
     connection row; `ReadMCPConnectionSessionTx` + `LockLiveSession` replace the joined locking statement; race-safety
     argument; fail-closed stop returned for review), §6 O-MCP-9 (both orderings, held revoke, no 40P01, statement order),
     §8 (early proof), §10 (fallback withdrawn; stop list extended).
  3. *Padded Unicode* → §5.4 (320-byte pre-check removed; normalize first, then bound the normalized value), §6 vector
     table (padded 80-non-BMP vector in core, web unit, IT-MCP-15 and M-RENAME).
  4. *Wire-level strictness* → §4.3 (raw-token validation, presence/duplicate/null rules, legacy fields unchanged, order
     before `begin`), §5.3 `strictFormField`, §5.4 (`request.clone()`, name read from raw bytes), §6 `TestSetupNameDecoding`,
     IT-MCP-15 wire negatives, `mcp-forms.test.mjs`, §7 contract text.
  5. *Failure outcomes* → §6 AT-ATOM rename fault rollback; §5.2 `AccountSections` + test-only fixture page and §6
     S-ACCOUNT-MCP-UNAVAILABLE (rendered, existing fixture-build harness) plus M-ACCOUNT `invalid_page` isolation; terminal
     classes split by reachable layer (browser-auth: `TestMCPRecovery`/IT-MCP-15 `unauthenticated`; target lookup:
     IT-MCP-15 `not_found`; store: `TestMCPIdentityRename`), §4.3 answer table, §11 AC5.
  - D-1…D-6 recorded as binding rulings in §10 (table), applied in §3, §4.1, §4.2, §4.3, §5.3, §5.4, §6; "ruling
    requested" wording removed; D-3 re-check added to §0/§8/§12 with the G020 reconciliation note (no G020 edits).
  - Unchanged: B3 (zero attempts spent), planning-only status, one write owner (Opus 5.5 high), readiness gates, the
    order G062 → G063 → G020 → G057 → G021 → OS001, §3 migration design, §5.1 component, §7 doc set, §9.
