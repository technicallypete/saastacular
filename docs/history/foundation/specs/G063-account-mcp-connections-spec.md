> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G063 Work Spec — Account MCP connections and friendly names

Revision: r1, 2026-10-21 UTC; ready for planning.
Goal: [G063](../goals/G063-account-mcp-connections.md).
Authority: owner requested reuse of the existing MCP setup UI under My Account
and friendly connection names, before SaaStacular completion. Planning authorized;
implementation requires the Goal's reviewed Plan/readiness/release gates.

## Requirements

**R1 — Reuse the management surface.** Embed the existing connection-management
experience in tenant-hosted `/account` using a shared server-rendered component.
Retain `/connections/mcp` as a working compatibility entrypoint using that same
component. Both support list, create, rename/clear and revoke; preserve the current
create/revoke routes and their request compatibility. My Account retains its
existing session/account/sign-out behavior. Unconfigured/unavailable MCP renders
a bounded unavailable state without breaking the rest of My Account.

Preserve own-live-connection newest-first 100-row keyset pagination, opaque cursor
validation and its user/Organization/origin binding. Component pagination links,
form returns, section/heading IDs and status messages must work for each embedding
page. Keep MCP cursor/status query handling disjoint from account/auth notices;
closed known same-origin return destinations only. Omitted return hints retain
the existing compatibility-page behavior. No arbitrary URL or query reflection.

**R2 — Optional names.** Creation accepts an optional user-supplied friendly name;
omitting it retains the existing unnamed-create contract. A caller may rename or
clear their own live pending or active connection through the current authorized
Organization host. Show `Unnamed connection` for null/legacy names. Do not derive
or backfill a name from a client ID, URL, user or provider. Duplicate names are
allowed. A friendly name is presentation metadata and must never become identity,
client binding, tenant selection, authority or an assertion of verified client type.

Authoritative core validation trims surrounding Unicode whitespace; ordinary blank
input stores null/clears the name. A stored name has 1–80 Unicode code points after
trimming. Reject invalid Unicode and control/newline input anywhere, including
before trimming (Unicode control characters and line/paragraph separators); do not
silently sanitize those characters. Browser/BFF convenience validation must agree
with core; JavaScript UTF-16 units are not the code-point limit. Render names as
escaped text, with safe wrapping and accessible labels/errors. Never interpret
HTML/Markdown or put the raw name in a query string or redirect.

**R3 — Preserve connection facts.** Keep the actual bound client separately visible,
including the current unbound state. Preserve URL, stored Address, status and
existing useful timestamps. Listing remains all of the caller's live rows in the
current Organization across stored aliases. List/rename/revoke may use a current
authorized host of that same Organization; a connection's stored host/resource and
transport's exact host/audience requirement remain unchanged. Removed or reassigned
aliases do not retarget a row. A label edit changes neither connection/session UUID,
URL/resource, user/Organization/client binding, selected Group/context version,
activation state/timestamp nor expiry/renewal/lifetime. It creates no extra session
or connection, renews no credential and cannot restore missing, expired, revoked,
disabled or recovery-invalidated authority.

**R4 — Existing authorization boundary.** Core owns validation and the name mutation,
using current exact-origin browser session, CSRF, current user/membership, explicit
own-row and resolved Organization checks under the existing transaction/access
facade and RLS defense in depth. A browser-supplied ID or friendly name grants
nothing. Serialize the rename's relied-on liveness/authority with concurrent
revocation/expiry using accepted transaction rules; stale UI must not authorize it.
Unknown, foreign-user/tenant and terminal targets fail under a bounded non-disclosing
contract, without changes to either connection or its security state.

Preserve create/list membership checks and the existing exact-own revoke contract,
including its intentional ability to revoke without current membership. Do not
widen any other caller's scope or make existing revoke depend on a new membership
gate. No mutation through GET; preserve form encoding, duplicate-field handling,
method restrictions, exact-host/Origin and session/CSRF refusals and cache policy.

**R5 — Additive persistence and bounded contracts.** Use a forward additive migration
for nullable metadata with compatible legacy rows, bounded storage constraints and
no fabricated backfill; never edit historical migrations. The Plan must identify
minimum column privileges, guard/RLS implications, migration preconditions and
rollback/retention implications. Preserve existing role separation; no blanket
runtime DML, bypass-RLS privilege or new general-purpose privileged surface.

The concise Plan pins the create optional-field and list response changes, rename
route/request/response/error schema, route/boundary allow-lists, form field ambiguity
and omission behavior, safe return selectors, Unicode validation and encoded
request/response byte bounds. Recompute the existing worst-case escaped 100-row
list response bound for names instead of dropping the bounded-read contract.
Existing clients omitting a create name remain valid. Name metadata need not enter
MCP discovery/tool payloads or transport binding reads unless narrowly justified.

Update only the affected closed diagnostic/audit contracts under existing policy.
Record the name mutation's required attributable audit disposition atomically with
the write using the existing audit mechanism; pin any narrow action/catalog/trigger
changes during Plan review. Raw names, prior/new values and request bodies never
enter diagnostic logs or audit payloads: use allowed IDs and closed event metadata.
No new logging/audit framework or broad catalog redesign.

**R6 — Existing application design.** Use repository Vestive guidance and current
application components/semantic tokens, preserving Light/Dark/System, System
default, keyboard/focus, accessible names/status, contrast, narrow-screen reflow,
theme/token-scope and reduced-motion guarantees. Follow current accepted shell
components and ADRs where preserved design source guidance differs. No new font,
remote asset, primitive library or frontend framework is needed.

## Architecture constraints and security invariants

Read AGENTS.md, CLAUDE.md, architecture/security/decisions, accepted ADR-005 and
relevant ADR-007 host rules, current roadmap/progress and this Goal. Core owns
facts, authorization and SQL; web remains a thin server-rendered UI/BFF and never
exposes internal core URLs or credentials. Existing global identity, immutable
Organization binding, optional connection-local Group selection, live execution
authorization and terminal lifecycle remain canonical. Naming changes no WorkOS
grant or compromised-client containment claim. G021's foundation barrier remains.

Current implementation/contract anchors (inspect and reconcile at Plan binding):

- `web/app/account/page.tsx`, `web/app/connections/mcp/{page.tsx,create/route.ts,revoke/route.ts}`,
  `web/lib/mcp-client.ts` and existing account/UI components.
- `core/internal/mcp/{connections.go,routes.go}`, `core/internal/identity/mcp{,_store}.go`,
  `core/internal/tenancy/access/mcp.go` and current migration/SQL security catalog.
- [Control-plane setup/wire contract](../../../../contracts/mcp/control-plane.md),
  [connection context](../../../../contracts/mcp/connection-context.md),
  [MCP guide](../../../foundation/MCP.md), database-isolation/audit contracts,
  [web shell](../../../foundation/WEB-SHELL.md) and
  repository design skill (not included).

## Expected artifacts

Claude-owned concise Plan with AC mapping, exact bounded scope/input inventory,
migration/grant/RLS and contract disposition, affected validation selectors and
resource ownership. After execution release: shared UI/forms and narrow core/store
changes, forward migration, affected contract/guide updates, meaningful regression
coverage and Result with attempt ledger, exact source/evidence, deviations, cleanup
and remaining gates. Keep approval/revision history and active snapshots intact.

## Acceptance criteria and verification

| ID | Required outcome | Verification |
|---|---|---|
| AC1 | Shared management works in My Account and the compatibility page without breaking account/session functions; MCP unavailable is isolated. | Source review plus meaningful browser list/create/rename/revoke flows through both entrypoints and disabled/unavailable checks. |
| AC2 | Optional create names, rename/clear, duplicates, neutral legacy fallback and separate actual client display obey R2. | Core/BFF/store tests and browser persistence/display checks, including legacy create requests and rows. |
| AC3 | Trim, 80-code-point bound, non-BMP Unicode, invalid/control/newline refusal and escaped display are consistent and bounded. | Boundary cases at 80/81, blank/clear, Unicode, hostile markup and multibyte/escaped maximum response tests; invalid writes leave no change. |
| AC4 | Mutations remain session/CSRF/host-bound, own-user/current-Organization only; negative targets disclose no foreign metadata. | Positive self-user tests and distinct other-user/tenant negatives, including a user belonging to both tenants, wrong/stale sessions, CSRF/Origin and duplicate/malformed fields. |
| AC5 | Rename changes metadata alone and cannot revive/renew a terminal or expired connection; independent connections stay isolated. | Real store/integration assertions for pending/active, expiry/revocation/disable/recovery and concurrent terminal transition; compare binding, Group/context, activation and session/lifetime facts before/after. |
| AC6 | Current-host alias management, stored Address, existing 100-row cursor semantics and safe page-specific returns/status remain intact. | Same-Organization alias and removed-host-row regression, wrong-tenant refusal, cursor-boundary tests and browser account/compatibility navigation/query-isolation checks. |
| AC7 | Additive migration, least privileges/RLS and bounded wire/log/audit changes conform to R5. | Forward migration over existing unnamed data, runtime-role/catalog and cross-scope negative evidence, focused guard tests, bounded encoding/parser tests and sanitized audit/log inspection. |
| AC8 | Required focused checks, accessible responsive appearances, reviews, merged-SHA full qualification and cleanup complete. | Attributable Result, existing browser/accessibility harness evidence, hosted fast and review dispositions, coordinator-owned local full receipt on exact merged SHA and cleanup receipt; final supervisor ACCEPT. |

## Required tests and validation

Reuse current `web/tests/unit/mcp-connections.test.mjs`, browser MCP/account
scenarios, core MCP/identity/access integration and migration/RLS harnesses. Extend
them for the new behavior and actual gaps; do not construct a new general suite.
The Plan selects affected web lint/type/unit, core unit/source/route-boundary,
MCP/store/access, migration/catalog/RLS and local HTTPS browser selectors. Cover
both appearances and a narrow viewport with keyboard, focus and useful status/error
behavior through the actual forms. Source review must verify component reuse and
authority separation; browser success alone is insufficient.

Finish safe batches, correct and rerun failures plus affected checks, retain failed
evidence and reuse unchanged passing coverage. Follow normal focused pre-merge +
hosted fast + Codex/Duo reviews → authorized staging merge → coordinator-owned
exact merged-SHA local full qualification. Keep that qualification checkout frozen,
respect shared-resource locks and clean owned resources using existing inventory
and bounded cleanup. Full qualification is not a new pre-merge gate or owner-run
command. No new owner manual gate is added; G062's separate M1 remains unchanged.

## Non-goals

No new provider/OAuth configuration or probes, MCP tool/catalog surface, connection
authentication model, tenant switching, Group provisioning rule, global manager,
terminal restoration, name uniqueness or inferred client identity. No account
redesign, clipboard sophistication requirement, framework, audit/cleanup machinery,
deployment or product feature work. G063 acceptance is required before G021 and
SaaStacular/OS001 completion; it does not itself accept or release those tasks.

## Clarification history

- 2026-10-21 — Initial bounded specification. Owner authorizes this foundation capability
  and planning now; supervisor schedules it after G057/before G021 while preserving
  G062 → G020 → G057. No implementation release or broader continuation mandate.

- 2026-10-21 — Later owner priority: plan G063 next and deliver it immediately
  after G062, before G020. Supersedes the initial after-G057 scheduling above;
  requirements, bounds and Plan/readiness/release gates remain unchanged.

- 2026-10-21 — Owner execution-independence clarification: G063 implementation
  may proceed while G062's M1, cleanup and final Goal ACCEPT are pending. G062's
  code integration and full local qualification are complete; its remaining
  closeout is not a G063 implementation dependency. Keep the G062 M1 preview
  stack untouched. G063 final ACCEPT remains a gate for G020 and G021. No product
  requirements or technical scope changed.
