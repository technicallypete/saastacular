# MCP control plane — wire contract (G014)

Status: implemented by G014 (architecture decision ADR-005, record not part of this distribution; behavioral contract
[connection-context.md](connection-context.md); operator guide [docs/foundation/MCP.md](../../docs/foundation/MCP.md)).
Core owns every decision. Web, clients and models supply no authority. This page pins the endpoint, discovery,
transport bounds, the five tools with their payloads and closed errors, and the internal setup contract that web uses.

## Connections and URLs

- A connection URL is `https://{host}/mcp/c/{uuid}`. `{host}` is an exact entry of the trusted host registry (lowercase
  `host[:port]`, bound one-to-one to a canonical Organization UUID). `{uuid}` is a core-minted, canonical lowercase UUID.
  The URL is a selector shown to its owner, not a credential.
- A connection is bound to one user and that host's Organization when it is created, and to one client at activation.
  Host, Organization, resource and user never change. A registry change never retargets an existing connection.
- States: `pending` (created, not yet used; expires after the pending lifetime), `active`, and terminal (revoked,
  expired or absent, all answering alike). Terminal is final; a new connection gets a new URL.

## Discovery

- `GET` or `HEAD` `https://{host}/.well-known/oauth-protected-resource/mcp/c/{uuid}` on a registered host with a
  syntactically valid UUID and no query answers `200` with exactly:

  ```json
  {"resource": "https://{host}/mcp/c/{uuid}", "authorization_servers": ["<MCP_CONNECT_ISSUER>"], "bearer_methods_supported": ["header"]}
  ```

  It is built from the URL alone, with no store lookup, so it is identical for pending, active, unknown and terminal
  UUIDs and while the store is unavailable. `scopes_supported` is omitted; scopes never confer authority. Other methods
  answer `405` (`Allow: GET, HEAD`); anything else `404`.
- A protected call without a token answers `401` with
  `WWW-Authenticate: Bearer resource_metadata="https://{host}/.well-known/oauth-protected-resource/mcp/c/{uuid}"`.

## Access tokens

The authorization server is WorkOS AuthKit Connect. Core is a resource server only: it mints no token and never
receives a refresh value. A token is accepted for one call only if all of these hold:
- signed by a key of the issuer's key set with an algorithm of `IDENTITY_JWT_ALGORITHMS`;
- `iss` equals `MCP_CONNECT_ISSUER` exactly; `aud` is exactly one value equal to the connection's resource URL;
- `exp`, `iat` and `nbf` hold with 30 s leeway; `sub` is a WorkOS user ID (`user_…`); `client_id` is present; an `act`
  (impersonation) claim is refused;
- `sub` maps through the configured G004 issuer to an existing principal (never provisioned here);
- the connection exists for that user in the host's Organization, its session is live, and an active connection's
  bound client equals `client_id`.

Any failure answers `401` (`error="invalid_token"` when the token itself is invalid) and runs nothing. The first
accepted call on a pending connection activates it and binds the client (one winner). A verified, mapped subject
presenting a token at a connection, host or client its own connection is not bound to is recorded as `access.denied`
(`resource_mismatch`).

## Transport

Streamable HTTP, stateless, JSON responses: `POST https://{host}/mcp/c/{uuid}` only (`405`, `Allow: POST`, otherwise).
No query string, no `Mcp-Session-Id`, no server-sent events.

| Rule | Answer |
|---|---|
| Unregistered host, malformed path | `404` |
| Repeated or unlisted `Origin` (none is allowed unless `MCP_ALLOWED_ORIGINS` lists it) | `403 origin_rejected` |
| Body over 64 KiB | `413 request_too_large` |
| Body not complete within the request deadline (read before any in-flight slot is claimed) | `400 invalid_request` |
| More than 64 requests in flight per instance; more than 20/s (burst 40) per connection after token verification | `429 rate_limited` |
| A JSON-RPC batch | `400 batch_unsupported` |
| A malformed or repeated `Authorization` header | `400 invalid_request` |
| Store or key set unavailable | `503 unavailable` |
| Request deadline | 10 s |

Claimed protocol revisions: `2025-11-25` and `2025-06-18`; another `MCP-Protocol-Version` header is refused by the
SDK. Accepted methods: `initialize`, `notifications/initialized`, `notifications/cancelled`, `ping`, `tools/list`,
`tools/call`; any other method is JSON-RPC "method not found" (the name is not echoed). `tools/list` is authorized on
current facts; a refusal is JSON-RPC `-32040` (`context_denied`) or `-32041` (`unavailable`).

## Tools

Every tool takes a closed JSON object (`additionalProperties: false`) and returns `structuredContent` with the same JSON
as text. Every call reauthorizes the connection, the user's current Organization membership and, when one is selected,
the Group. An unknown tool is JSON-RPC invalid params, and so is a tool that is not eligible for this request
([Discovery](#discovery-stale-context-and-notifications-g015)).

| Tool | Input | Output |
|---|---|---|
| `account.get_current_user` | `{}` | `{"user_id": uuid, "display_name": string or null}` |
| `account.get_current_context` | `{}` | `{"organization_id": uuid, "organization_name": string, "group": null or {"group_id", "slug", "name"}, "context_version": integer ≥ 0}` |
| `context.list_groups` | `{"cursor"?: string ≤ 128, "limit"?: 1..100 (default 50)}` | `{"groups": [{"group_id", "slug", "name", "selected": boolean}], "next_cursor": string or null}` |
| `context.switch_group` | `{"group_id": uuid or null}` (null clears) | the context, as `account.get_current_context` |
| `logout` | `{}` | `{"outcome": "terminated"}`; the connection is terminal |

- `context.list_groups`: Organization owners and admins see every Group of the Organization, other members only their
  own, ordered by Group UUID. The cursor is opaque and bound to the Organization.
- `context.switch_group`: selecting the current Group or clearing an empty selection changes nothing. A select needs the
  target selectable and any current selection still selectable; a clear needs only membership. A definitive refusal
  (`context_denied`, `group_not_selectable`, `invalid_input`) or a known abort leaves the previous context unchanged.
  An `unavailable` answer after a possible commit (a lost COMMIT reply, or a failed context read after the selection
  was stored) does not establish that nothing changed: before retrying, the client rereads the context with
  `account.get_current_context` when it can. The selection is stored per connection; other connections and the
  browser are unaffected.
- `logout` needs no membership: exact ownership is enough, so a former member can end their own connection. It ends
  this connection only, never the provider grant or other connections.

Tool errors are results with `isError: true` and `structuredContent` `{"error": code}` (plus `"field"` for
`invalid_input`); no identifier, name, role or driver text:

| Code | Meaning |
|---|---|
| `context_denied` | the connection or membership no longer authorizes the call (including a selected Group that lost validity) |
| `group_not_selectable` | the target Group is unknown, foreign or not selectable (identical for all three) |
| `invalid_input` | `field`: `invalid_arguments`, `invalid_cursor`, `invalid_page_size` or `invalid_group_id` |
| `unavailable` | the store or a dependency failed. Nothing changed if the call failed before its write; after a possible commit (a lost COMMIT reply, or a failed read after the write) the outcome is unknown, so reread the context before retrying |

## Discovery, stale context and notifications (G015)

These rules are additive. Production serves exactly the five tools above. The rules also govern any later tool that
is mapped to a capability, which is listed only when the request's context allows it
([AUTHORIZATION.md](../../docs/foundation/AUTHORIZATION.md#mcp-discovery-g015)). Listing is a hint, not a statement
that a target exists. The Organization owner or admin path is role-level and can list a Group or BusinessProfile tool
in an otherwise empty Organization. An unknown or foreign target still gets its non-disclosing refusal when called.

| Rule | Behavior |
|---|---|
| F-1 list | `tools/list` returns exactly the tools registered for this request, sorted by name, with no `nextCursor`. No pagination contract is added. |
| F-2 refusal | When the context no longer holds, `tools/list` answers `-32040`. A store failure answers `-32041`. Transport and liveness refusals are the 401 and 503 rules above. A list is never partial or cached. |
| F-3 local change | A successful `context.switch_group` commits before its response, so the next request on that connection lists the new context. Other connections are unaffected. A definitive refusal or a known abort changes neither the context nor the list. After `unavailable`, the outcome is unknown: reread or list again before retrying. |
| F-4 external change | A committed membership, role, link, policy, session, user or lifecycle change governs the next `tools/list` and `tools/call` on every connection. The credential stays the same and `context_version` does not change. A terminal connection stays 401. |
| F-5 no fallback | An invalid selected Group makes the read tools and `tools/list` answer `context_denied`. An explicit clear and `logout` stay available; the server never clears a selection on its own. |
| C-1 hidden = unknown | A `tools/call` naming a tool that is not eligible on this request gets the unknown-tool answer: JSON-RPC `-32602`, message `unknown tool "<requested name>"`, and nothing runs. This holds whether or not the tool was listed earlier. For one requested name the response is byte-identical whether the name exists or not. |
| C-6 refusal uniformity | A `tools/call` naming anything other than the five tools runs one eligibility check, whether the name exists or not. When the context no longer holds, both answer as C-1, with one `access.denied` (`mcp.request`) record each; a session that ends during the check answers as C-1. A store failure answers `-32041` for both. The five tools keep their own answers. |
| C-2 no list required | Eligibility is decided on the call's own request. An earlier list or `context_version` never authorizes, and listing first is not required. |
| C-3 visible still checked | A listed tool still runs its own check. A Group returned by `context.list_groups` and made unselectable afterwards refuses `context.switch_group`. |
| C-4 ordering | A call runs on the facts of its own snapshot or guarded write. It can also be unknown (not eligible at registration), refused by the tool's own check, or `unavailable` (F-3). It never runs on authority older than its own snapshot. |
| C-5 coherence | One `tools/list` result reflects one selection and one set of facts. The server never stores a catalog. |
| N-1 advertisement | `capabilities.tools.listChanged` is `false`. |
| N-2 delivery | No `notifications/tools/list_changed` is delivered: the stateless JSON transport has no stream. A client learns about a change by listing again. |
| N-3 selection revision | `context_version` counts committed selection changes only. It does not change for external changes, so an unchanged value does not mean the catalog is unchanged. List again regardless. |
| N-4 reconnect | A new client session on the same URL shares the connection's stored selection and lists fresh. No catalog is replayed. |
| N-5 truth | The server never claims that a catalog changed or stayed the same. A no-op select or clear leaves `context_version` unchanged. A select between Groups with equal catalogs bumps it without changing the list. |

## Internal setup contract (core, internal network only)

Web's setup surface (the shared connection manager on My Account, `/account`, and on the compatibility page
`/connections/mcp`; G063) calls these with `Authorization: Session <browser credential>`. The `origin` is
the page's own origin; core resolves the Organization only from the registry and checks the session, the CSRF token and
the membership itself. Answers are `Cache-Control: no-store` JSON; every error is exactly `{"error": code}`.

| Route | Request | Success |
|---|---|---|
| `GET /internal/v1/mcp/connections?origin=<origin>[&cursor=<next_cursor>]` | — | `200 {"connections": [{"connection_id", "url", "state": "pending" or "active", "client_id": string or null, "created_at", "activated_at": RFC 3339 UTC or null, "display_name": string or null}], "csrf_token": string, "next_cursor": string or null}`: one page of the caller's own live connections in that Organization, newest first (`created_at`, then `connection_id`, both descending; at most 100 per page), the session's CSRF token, so the page makes one core call, and the next page's cursor (`null` on the last page) |
| `POST /internal/v1/mcp/connections` | `{"origin", "csrf_token"[, "display_name"]}` | `201 {"connection": {"connection_id", "url", "state": "pending"}}` |
| `POST /internal/v1/mcp/connections/revoke` | `{"origin", "connection_id", "csrf_token"}` | `200 {"outcome": "revoked"}` |
| `POST /internal/v1/mcp/connections/rename` (G063) | `{"origin", "connection_id", "csrf_token", "display_name"}` | `200 {"outcome": "renamed"}`: sets, changes or (with `""` or whitespace only) clears the friendly name of one of the caller's own live pending or active connections, through a current host of its Organization (current membership required); nothing else about the connection changes |

| Error | Status |
|---|---|
| `invalid_request`, `ambiguous_parameter` | 400 |
| `invalid_name` (G063: a `display_name` outside the name rule or not validly encoded; create and rename; bounded and never echoed) | 400 |
| `unsupported_media_type` | 415 |
| `origin_rejected` (not a registry origin), `csrf_failed`, `not_available` (not a member; create, list and rename) | 403 |
| `unauthenticated` | 401 |
| `not_found` (not the caller's live connection, including one already terminal, unknown, another user's or another Organization's, byte-identically; revoke and rename) | 404 |
| `mcp_unavailable` | 503 |

`next_cursor` is an opaque, untrusted position: it only chooses where the newest-first listing resumes. A malformed
cursor, one outside the supported time range, or one issued for another user, Organization or setup origin answers
`invalid_request`, after the session, origin and membership checks. Every page rechecks the session and membership and
lists only the caller's live connections, whatever position is presented. Rows created after the first page appear on
a fresh first page; rows that end between pages are omitted, never duplicated. Revoking works from any page.

When MCP is not configured none of these routes, the endpoint or the metadata exists (`404`).

**Friendly names (G063).** `display_name` is presentation metadata chosen by the connection's owner: it never becomes
identity, client binding, tenant selection, authority or a claim about the client (the bound client stays `client_id`),
and it never enters the MCP endpoint, metadata, tool payloads, discovery or the transport binding reads. Listed rows
always carry the key (`null` for an unnamed or legacy connection); creating without it keeps the unnamed create.
Duplicates are allowed. The field is decoded strictly from the raw request body before any session or CSRF check: on
either route a second `display_name` key, `null` or a non-string value is `invalid_request` (a clear is `""`); a value
that is not valid UTF-8, or holds an unpaired or reversed surrogate escape, is `invalid_name`, refused and never
replaced; every other field keeps its existing decoding. Core then applies the rule in this order: any control
character (Unicode Cc, including tab, newline and U+0085), line separator or paragraph separator anywhere is refused
before trimming; surrounding Unicode White_Space is trimmed (effectively SP, U+00A0, U+1680, U+2000–U+200A, U+202F,
U+205F and U+3000; U+FEFF is kept and counted); an empty result clears the name; otherwise 1–80 code points are stored.
The 100-row listing's worst-case escaped response bound grows by 498 bytes per row. A rename locks the live user, the
caller's browser session, the target connection's session and the Organization facts before the connection row, so it
serializes with revocation and never revives, renews or extends anything ([MCP.md](../../docs/foundation/MCP.md)).

## Operational events

Core writes three catalogued events (`infra/validation/tools/internal/logrecord/events.json`) with closed fields and
request/trace correlation only; never a connection UUID, URL, token, subject, client ID, host or path:
- `mcp.request` (warn/error): a protected request refused before the tools run, with `outcome` `refused` or `failed`
  and a closed `reason` (`method_not_allowed`, `invalid_request`, `origin_rejected`, `rate_limited`,
  `request_too_large`, `batch_unsupported`, `token_missing`, `token_invalid`, `subject_unmapped`, `connection_unknown`
  (unknown, terminal or an expired pending setup, alike), `binding_mismatch`, `context_denied`, `unavailable`);
- `mcp.connection`: `activated` and `renewed` (info), or `refused` (`context_denied`, `connection_unknown`) and `failed`
  (`unavailable`);
- `mcp.setup`: one per setup request, `operation` `list`, `create`, `revoke` or (G063) `rename`, `outcome` `ok`, `refused` (the closed
  setup code) or `failed` (`mcp_unavailable`).
Guarded operations also write `access.decision` (operations `access.begin_verified_request` and `mcp.*`).
