> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G063 — Account MCP connections and friendly names — Result (B3 attempt 1, in progress)

Goal: [G063](../goals/G063-account-mcp-connections.md) · Plan: [r1](../plans/G063-account-mcp-connections-plan.md)
(sha256 `0be2cff9…5ec20`) · Inventory sha256 `2a56b412…69b6e` · Release: Codex confirmation (not included).

**Status: S1–S4 complete; corrections 1 (00019 postcondition) and 2 (E-LOGIN count) pushed; S5 batches 1–3 recorded
(batch 3a failed and was rerun as 3c; batch 2 recorded after the Codex source-review CORRECTION LOOP, see
"S5 ledger and S7 status" at the end); resubmitted for pre-integration review — not accepted.** Codex owns
acceptance. Attempt 1 of B3 (no budget reset); one write owner (writer 2 from the S2 checkpoint, writer 3 from correction 1; see below).

## Execution record

- Writer: Opus 5.5 (high requested; runtime effort not observable from inside the agent), isolated worktree
  `.claude/worktrees/agent-<id>`, branch `g063-impl` created at release SHA
  `50c5c421efab9f0764772da864faa17d13d425b5` (HEAD and branch asserted before any edit).
- S0: `git diff --stat fb9a8ba3 50c5c421 -- core web infra contracts docs` empty. 74 of 77 inventory blobs match at
  `50c5c421`; the three differing rows (Goal, Spec, PROGRESS) changed only in the release commit `50c5c421` (status
  lines and the owner execution-independence amendment; "no product requirements or technical scope changed").
  Migration slot: latest `00018_platform_signup.sql`; `00019` free when the migration was written.
- Context checkpoint at ~150k after S2; handed back at the S2 boundary for a fresh S3 worker (policy: safe boundary).

## Slices delivered

**S1 — migration (commit `741b1358`).** `core/migrations/00019_mcp_connection_names.sql`: preconditions (owner, five
roles, version 18, exact runtime column privileges / three policies / three triggers / FORCE on
`app.mcp_connections` and `audit.events`, `md5(prosrc)` of the 00018 attribution, 00010 capture and guard bodies, no
00019 object); nullable `display_name` with CHECK `mcp_connections_display_name_valid` (1..80 code points, `= btrim`,
`(display_name COLLATE "C") !~ '[\u0001-\u001f\u007f-\u009f\u2028\u2029]'` — the 00004 control-class precedent);
`events_action_valid` +`mcp.connection_renamed` (46→47); attribution body = 00018 + the action in the trigger-owned
list and the `mcp.*` shape rule (generated from the 00018 text; only those two edits); capture body = 00010 + an
`audit_capture_name` branch; strict trigger `audit_capture_name AFTER UPDATE OF display_name … WHEN (OLD IS DISTINCT
FROM NEW)`; `GRANT INSERT (display_name), UPDATE (display_name)` to runtime; postconditions. Down refuses named rows or
rename events (FORCE lifted inside the transaction) and restores the exact version-18 bodies (digests verified equal to
00018/00010) and action CHECK. Guard, policies, `sessions_insert` untouched.
Tests: `TestMCPNamesCatalog` (CT-MCPN-1 exact keyed version-18→19 catalog delta, CHECK/trigger text, column ACL,
bodies read from the embedded file, 47 actions, unchanged policies/guard/FORCE; CT-MCPN-CHECK accept/refuse vectors and
strict capture incl. same-value/clear; CT-MCPN-DML runtime name-only/own-row/role, other columns and DELETE 42501;
CT-MCPN-NEG six rolled-back detectors) and `TestMCPNamesMigration` (MG-MCPN-1 populated 18→19 rows identical, MG-2
down refused with a name / with history, MG-3 down/up/reapply exact catalogs, MG-4 failed up leaves 18 and the owner).
`audit.go` +`MCPConnectionRenamed` (40 trigger + 7 core = 47); `audit_test.go` reads actions/trigger-owned from 00019
and pins the 00018→00019 delta; `migrations_test.go` set pin.

**S2 — core (commits `24ed376f`, `5af87491`, `d7fc4e29`, `b7df50e0`).** identity: `NormalizeMCPDisplayName` (D-6 order:
UTF-8 → Cc/Zl/Zp refused anywhere → `TrimFunc(unicode.IsSpace)` → empty = clear → ≤80 code points),
`ErrInvalidMCPName`, `MCPConnectionInfo.DisplayName`, insert/list carry `display_name`, new `ReadMCPConnectionSessionTx`
and `RenameMCPConnectionTx` (row `FOR NO KEY UPDATE` last, expired pending refused, name-only UPDATE). access:
`RenameMCPConnection` (user → browser session → target 'mcp' session via `LockLiveSession` → locked facts +
`contextCheck` → row; dead/unknown/foreign target `not_found`/`target_not_found`; non-member `context_denied`;
Organization-only); `CreateMCPConnection(..., displayName *string)`. mcp: `POST /internal/v1/mcp/connections/rename`,
optional `display_name` on create, raw-token decoding before `begin` (duplicate incl. case-folded alias / null /
non-string → `invalid_request`; invalid UTF-8 or unpaired surrogate escape → `invalid_name`), list rows always carry
`display_name`, `setupFailure` maps `not_found`; `main.go` mounts the route. events.json catalogs `rename`,
`invalid_name`, `mcp.connection.rename`.
Tests: `TestNormalizeMCPDisplayName` (shared vector table), `TestSetupNameDecoding`, `TestMCPScopeBinding` rename cases,
list wire bound; integration: `TestMCPIdentityRename`, O-MCP-9 (rename-first / revoke-first / held revocation, no
40P01, statement order `user,session,target,user,session,facts,row,write`), AT-INV rename producer, AT-ATOM rename
fault rollback, IT-MCP-3 named create, IT-MCP-15 (full matrix), `TestMCPRecovery` rename cases.

## Mechanical pins changed (in allow-listed files)

`audit_test.go` 39/7/46 → 40/7/47; `migrations_test.go` set; `boundary_test.go` `pinnedExportedMethods`
(+`Service.RenameMCPConnection`), user-taking identity functions (+the two Tx helpers), **`inContextTx` call-site count
18 → 19** (not named in Plan §2; same file, a count pin caused by the planned facade operation — Plan wording
reconciliation); `identityMCPProtectedFuncs` unchanged (asserted: the new helpers name only `app.mcp_connections`);
`mcp_test.go` package import allow-list +`unicode/utf8` and list wire bound 1 277 222 → 1 327 022 (+100 × 498; the Plan
§5.3 recompute); call sites (`mcp_integration_test.go`, `audit_integration_test.go`, identity `mcp_integration_test.go`).

## Validation run (stack-free, plain Docker, `golang:1.27.3-alpine`, `--network none`, read-only source and module
cache volume `g058-a1-gomod`, `GOFLAGS=-mod=readonly GOPROXY=off`, disposable `GOCACHE`)

- `gofmt -l .` clean; `go vet ./...` and `go vet -tags integration ./...` OK (every integration test compiles).
- `go test -count=1 ./...` (core unit): every package ok **except `core/internal/migrate`**: `TestNormalSetIsTheReviewedSet`
  and `TestProviderAcceptsProtectedStore` fail on their 18-migration pins in `runner_test.go` (outside the allow-list;
  see below). `TestSetupNameDecoding` confirmed executed (`-v`: PASS).
- Not run by writer 1 (need DB/stack/browser/build; coordinator-owned): every integration selector. (Later run by the
  coordinator: S5 batch 2 at `cec82c82`; see "Batch 2" at the end.)

## Allow-list gap (stop item for Codex disposition)

Adding `00019` and the rename route requires the same mechanical head-version / route-catalog parity edits in files
the inventory allow-list does not admit (no edit made):

1. `core/internal/migrate/runner_test.go` (unit): normal set list += `00019_mcp_connection_names.sql`; `ListSources`
   count 18 → 19 and `srcs[18].Version != 19`.
2. `core/internal/migrate/integration_test.go` (migrations suite): full-set pins (`current_version\t18` → 19 at :984/
   :1053/:1064/:1075/:1195, applied lists += 00019, reapply/down of 00019 to 18), and a `metadataThroughVersion19`
   form (00019 adds no relation or function).
3. `core/internal/tenancy/integration_test.go` (:144): a version-19 step-down branch (purge names
   `UPDATE app.mcp_connections SET display_name = NULL` and the audit history, then down to 18).
4. `core/internal/tenancy/access/integration_test.go` (:236) and `recovery_integration_test.go` (:92):
   `current_version\t18` → 19 (harness of TestMCPOrdering/TestAudit).
5. `core/internal/mcp/env_integration_test.go` (:253): `current_version\t18` → 19 (harness of every MCP IT; the rename
   route is mounted by the test-local `serveRename`, so no other edit is needed there).
6. `infra/validation/tools/internal/logrecord/completion.schema.json` (:39, :106): core route enums +=
   `/internal/v1/mcp/connections/rename` (core logs the matched pattern; the validated core log would otherwise fail).

`infra/validation/run.sh` (allow-listed, "pins named in Plan §2") also carries the same version-18 text in B-PERSIST,
B-TOGGLE-OFF, root-migrate, recovery-drill and rename-suite checks (:1867-1904, :4247, :4802-4808, :4855, :4968-5015) —
same class as the named B-MIGRATE-* pins; to be updated in S4 with the named pins unless Codex rules otherwise.

## Remaining (S3–S7) — for the next worker (writer 1 handoff; historical, current state at the end)

S3 web (`mcp-client.ts`, `connection-manager.tsx`, `account-sections.tsx`, both pages, three form routes, fixture page,
`mcp-connections.test.mjs`, new `mcp-forms.test.mjs`, browser specs); S4 contracts/docs, run.sh named pins (+ the
version-text class above), `web-browser.Dockerfile` fixture `cp`; S5 batches via the coordinator; S6 self-review; S7
fetch/rebase/reconcile and final Result (AC mapping, cycle counts).

## Deviations

None from the approved design. Plan-wording reconciliations: `inContextTx` count pin (above); CHECK regex written with
the 00004 `COLLATE "C"` control-class idiom (same refused set as the Plan text); `RenameMCPConnectionTx` reads the
service clock instead of a `now` parameter (as `ActivateMCPConnectionTx`).

## Writer 2 (S3–S4 and allow-list amendment 02)

**Execution record.** Opus 5.5 (high requested; runtime effort not observable). Isolated worktree
`.claude/worktrees/agent-<id>`, local branch `g063-impl-w2` created from `origin/g063-impl` with HEAD and
branch asserted equal to `3929788bfe0bb6e796f396d7bf41a96d19e79e2c` / `g063-impl-w2` before any edit; pushes are
fast-forward `g063-impl-w2:g063-impl`. The first writer's worktree was not touched.

**Amendment 02 reconciliation.** `origin/staging` `23cf1a6a90ceaecfcc4c31bd955c216af35e2ac9` merged (merge commit
`d70f22f4`; a merge, not a rebase, because the task branch is fast-forward only). Delta: management only (PROGRESS, the
amended inventory, the Codex amendment review); before the merge `git diff --stat HEAD origin/staging -- core web infra
contracts docs` listed only this branch's own G063 changes. Amended inventory sha256
`9152b75f271186f644f751159fd0d299d8fbf0a281776e0c69dff1432345304a`; review record
`G063-allowlist-amendment-codex-review-20261021.md` sha256
`792a0c0192197d43bcdb85265af992f0bcc92be331ff229fcc0bb27760ec41fb` (both recomputed after the merge).
Parity edits (commit `05b8a97a`), only within the amendment's limits:
`runner_test.go` (normal set + `00019`, `ListSources` 19, `srcs[18].Version == 19`); migrate `integration_test.go`
(`metadataThroughVersion19` = the version-18 form, since 00019 adds no relation or app function; up/synthetic-down/
reapply/up-again at 19, and the down sequence gains its 19→18 step with the version-18 metadata and Group-authority
checks before the existing 18→17 step); tenancy `integration_test.go` (a version-19 step-down branch: names cleared
through the attested superuser, audit history purged, then down to 18); access `integration_test.go`,
`recovery_integration_test.go` and MCP `env_integration_test.go` (`current_version\t19`); `completion.schema.json` (both
core route enums += `/internal/v1/mcp/connections/rename`). No test removed or skipped.

**S3 (commit `45b9a71d`).** `web/lib/mcp-client.ts`: `displayName` (key required; `null` or an already-normalized name,
else `malformed`), the recomputed bound (`LIST_ITEM_MAX_BYTES` 13 267, `MCP_LIST_MAX_RESPONSE_BYTES` 1 327 022 = core's
pin), create sends `display_name` only when given, `renameConnection` (exactly `{"outcome":"renamed"}`), `invalid_name`,
`normalizeDisplayNameInput` (well-formed → no Cc/Zl/Zp anywhere → trim of the explicit effective set by code point
(no regex escapes, no `trim()`) → ≤ 80 code points), `strictFormField` (raw bytes of `request.clone()`, byte-level
percent-decoding, `TextDecoder` fatal with `ignoreBOM: true` so U+FEFF is kept as in core), `returnTarget` and
`statusLocation`. `components/account/connection-manager.tsx` (shared, server-rendered; one list call, or the
compatibility page's own `preloaded` answer; closed in-section states `mcp-unavailable`, `mcp-not-available`,
`mcp-invalid-page`; name heading as escaped text with "Unnamed connection", bound client separate, rename form with
visible "Name" label, hint, no `maxlength`; revoke and create forms carry `return_to`); `components/account/
account-sections.tsx` (the signed-in `/account` body: Session, the manager, Sign-out). `/account` renders it with
`mcp_cursor`/`mcp_status` and a lookup request ID for the list; `/connections/mcp` keeps its whole-page branches and
renders the manager. Form routes `create` (optional name), new `rename`, `revoke` (+`return_to`). Test-only fixture
`web/tests/fixtures/account-mcp/page.tsx`. Unit: `mcp-connections.test.mjs` (fixtures carry `display_name`; widest
answer with an 80-`<` name; shared vector table; listed-name validation; create/rename requests; `strictFormField`
raw-byte table; `return_to`), new `mcp-forms.test.mjs` (raw-body route negatives with no core call, padded vector,
return selector, legacy fields, methods, Origin/media type/no session, answer mapping, no name in Location/body/log).
Browser: `account.spec.ts` S-ACCOUNT-MCP-UNAVAILABLE; `mcp.spec.ts` M-SETUP (named + unnamed), M-RENAME, M-ACCOUNT
(creates disposable pending rows to cross one 100-row page for the `mcp_cursor` links and revokes every row it
created), M-A11Y (+ `/account` at 1280/320; Tab bound 40 → 120 because each row now has three tab stops).

**S4 (commit `7c247e83`).** `contracts/mcp/control-plane.md`, `docs/foundation/{MCP,AUDIT,DATABASE-ISOLATION,MIGRATIONS,
WEB-SHELL}.md`, `docs/SECURITY.md` §5 runtime row; `infra/docker/web-browser.Dockerfile` (fixture stage only: one more
`mkdir` target and `cp tests/fixtures/account-mcp/page.tsx app/account-fixtures/page.tsx`); `core/internal/mcp/
routes.go` (the `// G063` trailing comment moved above the block: B-MCP-SOURCE extracts route patterns only from lines
ending in the quoted value, so with the comment the rename route was silently unpinned).

**`run.sh` pins (all recomputed from the real files).** Named (Plan §2/§10): B-WEB-BOUNDARY (`WEB_ROUTES` +
`connections/mcp/rename/route.ts`, request-context allow regex `(create|revoke|rename)` +
`components/account/(connection-manager|account-sections)\.tsx`, `web_routes_ok` message, PASS text); B-MCP-SOURCE
(`MAIN_GO_G016_SHA` → `73d8baa283019a3183975faf40a8f1877f09591afec4e91bc9cbcbe64258da8e`, `CORE_MCP_ROUTES` four routes
C-sorted, the `mux.HandleFunc` set + `RouteConnectionsRename`, texts); B-MCP-CATALOG (`-run` + `TestMCPNamesCatalog|
TestMCPNamesMigration`, counts 4/4 and 4/4, texts); B-MCP-IT (TestMCPLifecycle 6 → 7 for IT-MCP-15, TestMCPOrdering
10 → 11 for O-MCP-9, texts); B-MIGRATE-SET (set + 00019, text); B-MIGRATE-CLI/CATALOG (fresh status 19 pending; apply,
reapply, re-apply at 19; rollback/inspect at 18 with 79 relations, so the unique catalog rows are P, 0 and 79 — the 71
row retires; production-refusal version 19; texts); B-AUDIT-CATALOG (action source `00019_mcp_connection_names.sql`,
47 actions / 40 trigger-owned, text); B-AUDIT-IT (`AUDIT-INV` `{"actions":47,"observed":47}`, version 19, texts).
Under the coordinator-relayed existing authority for the reported version-18 texts: root-migrate (19 pending, applied,
listed; versions `0 19 19`), B-RECOVERY-DRILL (source up to 19, text), B-PERSIST (history + 00019, version 19, text),
B-DBRENAME (version 19, text), B-TOGGLE-OFF (version 19, inspect + 00019, text).

**Stack-free checks run by writer 2** (plain Docker, `golang:1.27.3-alpine`, `--network none`, read-only source,
`g058-a1-gomod` read-only, `GOFLAGS=-mod=readonly GOPROXY=off`, disposable `GOCACHE`):
`gofmt -l` empty, `go vet ./...` and `go vet -tags integration ./...` OK; `go test -count=1 ./...` (core): **every
package ok, including `core/internal/migrate`** (`TestNormalSetIsTheReviewedSet`, `TestProviderAcceptsProtectedStore`
PASS with `-v`); `infra/validation/tools` `go test ./internal/logrecord/` ok; `bash -n infra/validation/run.sh` OK; a
disposable read-only simulation of the B-WEB-BOUNDARY request-context and route-inventory checks, the component
server-config scan and the main.go digest: all match. **Non-evidentiary sanity only** (not B-UNIT-WEB): the two MCP
unit files under `node:24-slim --network none` on a disposable copy with a stub `server-only` package: 26/26 pass
(it caught and fixed a duplicate `body` declaration and a test-helper default before commit). No lint/typecheck was run
(no prepared web check image; `web-check` needs a Compose build). Not run by writer 2: every DB, Compose, browser and
harness batch (later run by the coordinator as S5 batches 1–3; see the end).

**Plan-wording reconciliations (no design change).** The audit row is T40 (the Plan's "T30" number is taken by
`organization.branding_changed`). `/account` NAV gains only "AI connections" (a "New connection" anchor would dangle in
the manager's failure states). WEB-SHELL.md had no MCP route table, so the rename route is documented in a new G063
section with a three-route table. The manager accepts the compatibility page's `preloaded` list so that page still makes
one core call; its forms use the list's CSRF token (the same session token as the Sign-out forms). The manager's closed
states use their own test ids so the page's single `status-message` stays unique.

**Open items for Codex/coordinator (not edited).**
1. **Stop item — `web/tests/browser/identity.spec.ts` (outside the allow-list):** its `csrfOf` (line 235–237) asserts
   exactly 3 `csrf_token` inputs on `/account`. With MCP configured in the identity environment, the embedded manager
   adds its create form (and two per connection), so the identity E-* tests that call `csrfOf` on `/account` (lines 578,
   595, 620, 645, 649, 678, 1012) would fail. Smallest parity fix: scope the locator to the Sign-out forms
   (`'#sign-out input[name="csrf_token"]'`), keeping `toHaveLength(3)`. Needs an allow-list amendment before batch 3a.
2. Required browser-ID matrices live in `infra/validation/lib.sh` (not allow-listed): `BROWSER_MATRIX_IDENTITY_E2E`
   (`mcp-light mcp-dark|M-NOAUTH … M-HANDOFF`) and `BROWSER_MATRIX_FULL` (shell projects) do not list M-RENAME,
   M-ACCOUNT or S-ACCOUNT-MCP-UNAVAILABLE. The new cases run and must pass (`unexpected` 0), but are not required IDs
   unless Codex admits that edit.
3. Descriptive "at version 18" PASS texts not named by the Plan or the relayed authority remain: B-RLS-IT (`run.sh`
   :3150), B-ONBOARDING-IT (:3510), B-MEMBERSHIP-IT (:3534), B-ENROLLMENT-IT (:3558), B-SIGNUP-IT (:3580),
   B-IDENTITY-CATALOG (:5149). Text only (no assertion depends on them); left for disposition.

**Amendment 03 reconciliation (browser parity; items 1–3 above resolved).** `origin/staging`
`2254d7b0db8107ce09adb547573963f4a430d03c` merged (management-only delta: PROGRESS, the inventory, the review record).
Amended inventory sha256 `29c9b872f45e4b1f7643a6eba5f8a532681a0b7e41478dd58d3b53dca9bca383`; review record
`G063-browser-parity-amendment-codex-review-20261021.md` sha256
`8123c9d137ad051a0ed6ef101053daf6cbd52b454cf1089d9f1b76846c0ec093` (both recomputed after the merge). Edits:
`web/tests/browser/identity.spec.ts` `csrfOf` locator scoped to `#sign-out input[name="csrf_token"]` (count 3, same
token, secret shape and observation unchanged); `infra/validation/lib.sh` `BROWSER_MATRIX_FULL` shell entry +
`S-ACCOUNT-MCP-UNAVAILABLE` and `BROWSER_MATRIX_IDENTITY_E2E` mcp entry + `M-RENAME M-ACCOUNT` (nothing else; note
`BROWSER_MATRIX_FULL` is also the matrix of the B-TOGGLE-OFF browser run, which runs the same shell projects);
`run.sh` the six descriptive "at version 18" PASS texts above → 19 (word diff: exactly six `18`→`19` tokens, one per
line, all inside those PASS strings). Checks: `bash -n` on `run.sh` and `lib.sh` OK; a read-only simulation that parses
both edited matrices as `browser_matrix` splits them and looks each required ID up as a `("[ID]` test title in the
specs the projects collect: every new ID found; the three IDs it reports missing (S-RESPONSIVE, S-BLUEPRINT,
S-LONG-IDENTITY) are pre-existing template-literal titles, outside the simulation's pattern. Still open: the amendment-02
pattern `^core/internal/tenancy/access/(integration|recovery_integration_test)[.]go$` admits `integration.go`, not the
edited `integration_test.go` (unchanged by amendment 03).

**Guard fix (`9349814f`).** B-IDENTITY-GUARDS (batch 1 at `047c1955`) reported COOKIES: the comment at
`connection-manager.tsx:16` contained the literal `cookies()`; reworded to "No cookie access". The `identity_guards`
function extracted verbatim from `run.sh` now prints no code on this tree, and the guard literal scans over
`web/{app,lib,components}` find nothing outside the admitted files.

**Parity amendment 04 reconciliation.** `origin/staging` `609048db5f23cf88dd7a17b18f3650f6a5ce1c90` merged
(management-only delta: PROGRESS, the inventory, the review record). Amended inventory sha256
`cfe23a02030d808baea6c3d4a1687b5fd9f00aa2815de6df73024cb80eaf39c6` (line 136 now
`(integration_test|recovery_integration_test)`, resolving the pattern item); review record
`G063-parity-amendment-codex-review-20261021.md` sha256
`15b2bb43c420d7df603aade3287d47781dc2593c08ba860128dcd0a0df199a20` (both recomputed after the merge). Codex rulings:
the pattern is corrected; B-TOGGLE-OFF's shared `BROWSER_MATRIX_FULL` requirement stands with no extra edit; the
B-IDENTITY-E2E PASS description (`run.sh` :5406) now names M-RENAME and M-ACCOUNT in both mcp projects (word diff:
`M-HANDOFF)` → `M-HANDOFF; G063 M-RENAME and M-ACCOUNT)` plus `(mcp-light and mcp-dark)`; no selector or required-ID
logic changed; `bash -n` OK).

## Writer 3 — Correction 1 (00019 postcondition boolean text; B3 attempt 1, no budget reset)

**Execution record.** Opus 5.5 (high requested; runtime effort not observable). Isolated worktree
`.claude/worktrees/agent-<id>`, local branch `g063-impl-w3` created from `origin/g063-impl` with HEAD
asserted equal to `d0f2872d0c7a564bece360cbf455266229ccc852` before any edit; push fast-forward `g063-impl-w3:g063-impl`.
Earlier writers' worktrees were not touched.

**Defect (first private DB batch at `d0f2872d`).** Every up to version 19 failed its own postcondition: `ERROR: 00019
postcondition: the name column, its CHECK, the policies, the triggers or FORCE differ (SQLSTATE P0001)`. So the migrate
lifecycle failed from step 05, and selectors 2–5 (mcp, identity, access facade, access audit) never reached their
tests. The version-18 precondition block and the first postcondition (runtime column privileges) did pass: the error
names the second check. The tenancy log's one failure (`IT-BP-REF-2`, SQLSTATE 53300 "remaining connection slots are
reserved") is connection exhaustion in that run, not 00019; it is not addressed here. (`IT-BP-REF-2` passed in the
later selector-6 run `db5`; see "Batch 2" at the end.)

**Root cause.** `00019_mcp_connection_names.sql` :436-437 built
`format('%s:%s:%s', format_type(atttypid, atttypmod), attnotnull, atthasdef)`. `format('%s', boolean)` uses the boolean
output function (`t`/`f`), so the value is `text:f:f`; it was compared with `'text:false:false'`. (Only an explicit
`::text` cast gives `true`/`false`.) The coordinator's throwaway pgvector:0.8.7-pg17 check returned `text:f:f`.
`TestMCPNamesCatalog` CT-MCPN-1 had the same mistake: `namesCatalog` builds the `col:` value with
`format('%s|%s|%s', …, a.attnotnull, a.atthasdef)` (`text|f|f`), but the test expected `text|false|false`.

**Changes (2 lines, no semantic change):**
- `core/migrations/00019_mcp_connection_names.sql` :437 `IS DISTINCT FROM 'text:false:false'` → `'text:f:f'`.
- `core/internal/migrate/mcp_integration_test.go` :798 `got != "text|false|false"` → `"text|f|f"` (the assertion was
  itself the formatting bug: nullable text with no default is still required).

**Every other postcondition and catalog-text pin reviewed (no other change needed):**

| # | Suspect | Disposition and evidence |
|---|---|---|
| 1 | Up postcondition: runtime column privileges (string_agg of `information_schema.column_privileges`) | No change. It passed in the batch (it runs before the failing check); `sql_identifier` sorts in C order. |
| 2 | Up postcondition: CHECK exists, `contype = 'c'`, `convalidated` | No change. An inline column CHECK added by `ADD COLUMN` is validated. |
| 3 | Up postcondition: policy list | No change. Same text as the precondition, which passed. |
| 4 | Up postcondition: trigger list order `audit_capture_name,audit_capture_selection,audit_capture_state,mcp_connections_guard` | No change. `tgname` is type `name` (C collation): `…_n` < `…_se` < `…_st` < `m…`. |
| 5 | Up postcondition: FORCE on both tables | No change. Same expression as the precondition, which passed. |
| 6 | Up postcondition: guard `md5(prosrc)` `3f34bb…` | No change. Same digest as the passing precondition; local extraction of the 00010 guard body also gives `3f34bbcc…`. |
| 7 | Up postcondition: `proconfig IS DISTINCT FROM ARRAY['search_path=pg_catalog, pg_temp']` | No change. Same formulation as 00018's postcondition, which passes against real PostgreSQL. The existing audit/MCP catalog pins (`search_path=pg_catalog, pg_temp`) show the same stored text. |
| 8 | Up postcondition: `has_function_privilege('public', oid, 'EXECUTE')` | No change. Same call as 00018's passing postcondition. 00008:652 and 00010:632 `REVOKE ALL … FROM PUBLIC` on both functions, and `CREATE OR REPLACE` keeps the ACL. |
| 9 | Up postcondition: `position('''mcp.connection_renamed''' IN prosrc)` and `IN pg_get_constraintdef(...)` | No change. Both bodies contain the quoted literal, and the deparsed CHECK prints `'mcp.connection_renamed'::text`, which contains it (the precondition's `> 0` test on the same expression ran correctly). |
| 10 | Precondition digests `d5b633…` (00018 attribution) and `fa7e3b…` (00010 capture) | Proven by the batch (the precondition passed). The same local extraction method reproduces both from 00018/00010, which validates the method for #11. |
| 11 | Down: the restored capture and attribution bodies must be byte-identical (reapply precondition, MG-MCPN-3 catalog equality) | Verified: the 00019 down bodies hash to `fa7e3bbb…` and `d5b63324…`, equal to the precondition digests. Diff of the 00019 up bodies against 00018/00010: only the two attribution edits and the capture `audit_capture_name` branch. |
| 12 | Down: restored `events_action_valid` must deparse exactly as version 18 | Verified: same `CHECK (action IN (…))` form and the same 46 values in the same order as 00018 :195-210 (deparse order follows the list). |
| 13 | Down: dropped column's ACL in the `acl:` key (the query has no `attisdropped` filter) | No change. The down revokes both column grants before `DROP COLUMN`, so `attacl` is empty or NULL, and `cardinality(...) > 0` excludes it. The dropped column is also renamed `........pg.dropped.N........`. |
| 14 | Down: refusal message `% named connections / % rename audit rows` (bigint) | No change. Prints `1`/`2`, matching MG-MCPN-2's expected strings. The FORCE lift is the 00010 down precedent. |
| 15 | `mcpnWantCheck` (pg_get_constraintdef with BETWEEN, `COLLATE "C"`, `::text`, parentheses) | No change. The structure is identical to the passing `business_profiles_legal_name_valid` pin (business_profile_integration_test.go:310), with the same nullable form, BETWEEN expansion, `(x COLLATE "C") !~ '…'::text` and `))))` closing. Go raw string, so the `\u` escapes stay literal. The file holds literal `\u0001… ` escapes, not raw characters (checked with byte grep). |
| 16 | `mcpnWantTrigger` (`WHEN ((old.x IS DISTINCT FROM new.x))`, qualified `ON app.…` and `audit.…()`) | No change. Same shape as the passing pins at mcp_integration_test.go:82-84 (`audit_capture_state` and `audit_capture_selection`). `catalogConn` uses `search_path = pg_catalog`, so the function is qualified. |
| 17 | Column ACL `string_agg(... ORDER BY privilege_type)` = `rt:INSERT,rt:UPDATE`; `$1::regrole` with a Go string | No change. A column has no default ACL entries, so only the runtime appears. `$1::regrole` with a string parameter is used by the passing onboarding/membership tests. |
| 18 | `strings.Count(def, "::text") == 47` | No change. `action` is `text` (00008:198), and each `IN` element deparses as `'…'::text` (see 00018's passing return-path pin). |
| 19 | Body digests from `migBody` (`AS $$` … `$$`) vs `md5(prosrc)` | No change. Same extraction as the passing `mcpBody`/`onbBody` helpers. No `$$` inside the bodies. |
| 20 | `namesCatalog` on `cat` (owner, `search_path = pg_catalog`) vs a superuser `tx` (default path) in CT-MCPN-NEG | No change. Neither path contains `app` or `audit`, and no migration or init script creates objects in `public` (no extension or type there), so regclass, regprocedure, triggerdef and policy text match. |
| 21 | Other G063 DB tests' boolean text (`core/internal/mcp` rename shape `…|true|true|user|true|false`) | No change. Every boolean there is an explicit `::text` cast (`true`/`false`). The other `concat_ws` snapshots compare before and after with the same query. |
| 22 | Guard behaviour for name-only writes (CT-MCPN-DML / CHECK) | No change. The 00010 guard inspects only the binding, state, client and selection columns. |

No design problem found. At correction 1, every check after line 437 in the up block, plus the down block and its
dependent tests, was not yet executed against PostgreSQL. They have since run and passed in S5 batch 2 at `cec82c82`
(migrate up/down/re-up 19/18/19, `TestMCPNamesCatalog` and `TestMCPNamesMigration`; see "Batch 2" at the end).

**Stack-free checks (writer 3).** Plain Docker `golang:1.27.3-alpine`, `--network none`, `core` mounted read-only,
`g058-a1-gomod` read-only, `GOFLAGS=-mod=readonly GOPROXY=off`, `GOCACHE` on a disposable tmpfs, container `--rm`.
`gofmt -l .` printed nothing. `go vet ./...` OK. `go vet -tags integration ./...` OK. `go test -count=1 ./...`: every
package ok, including `core/internal/migrate` and `core/migrations`. Static: the only remaining `format()` calls in
00019 are the two `%I` grant statements. Body and action-list extraction used a read-only script over the migration
files. No DB, Compose, browser or build was run.

## Writer 4 — Correction 2 (E-LOGIN account-page CSRF count; no budget reset)

**Execution record.** Opus 5.5 (high requested; runtime effort not observable). Isolated worktree
`.claude/worktrees/agent-<id>`, local branch `g063-impl-w4` from `origin/g063-impl` with HEAD asserted equal
to `de7dc6be29f432cdcf333c814df648149b8201dc` before any edit; push fast-forward `g063-impl-w4:g063-impl`.

**Defect.** Harness `ops` B-IDENTITY-E2E at `de7dc6be` (batch 3a, run `e5db42`; console sha256
`e8c09eaf10e50212689a078199a7ea9841ab4779f0725f665ede59b7c07a3272`): 153 passed, including
mcp-light/mcp-dark; exactly one failure in identity-light and identity-dark, `[E-LOGIN] [A-1]` at
`web/tests/browser/identity.spec.ts:359`: `expect(saved.replaced).toBe(3)` received 4.

**Cause (verified from source).** `sanitizeHtml` counts every `<input … name="csrf_token" … value="…">`. My Account now
renders the three Sign-out forms plus `ConnectionManager` (`connection-manager.tsx`): one create form when the list
answer is ok, and two forms (rename, revoke) per connection row. So the count is 3 + 1 + 2n when the list succeeds and 3
otherwise; 4 means only "list ok, zero rows". That holds for Alice on org-a: the harness applies `mcp-seed.sql` before the
identity projects (every sign-in user a member), and no identity test creates a connection (mcp.spec.ts uses Dave). No
leak or per-row form is involved. `csrfOf` was already scoped to `#sign-out`.

**Change (one assertion, plus a comment).** `identity.spec.ts` E-LOGIN: `toBe(3)` → `toBe(4)`, commented as the three
Sign-out forms plus the G063 create form. It is still an exact count; no assertion was removed or weakened.

**Other parity checks: none needed.** `account.spec.ts` S-ACCOUNT-MCP-UNAVAILABLE scopes its counts to `#sign-out` and
asserts zero MCP forms in the unavailable state. The `mcp.spec.ts`, `onboarding.spec.ts` and `enrollment.spec.ts`
`csrfOf` helpers do not use exact counts. `infra/validation/run.sh` and `lib.sh` hold no identity.spec.ts digest or count
pin. Stack-free only: diff review. No Compose, browser, build or harness suite was run.

**Coordinator batches (recorded by writer 5, docs only; evidence under
`<home>/.local/state/platform-evidence/g063-a1/`).** The fix is `68b663f4`: `toBe(3)` → `toBe(4)` with the comment.
The account page's CSRF form count is 3 Sign-out forms + 1 create form + 2n row forms when the list loads.

| Batch | Source | Scope | Outcome | Evidence |
|---|---|---|---|---|
| 3a | `de7dc6be` | `ops` suite, run `e5db42` | **FAILED.** One assertion, `identity.spec.ts:359` E-LOGIN `toBe(3)` received 4 (the G063 create-form token), in identity-light and identity-dark; 153 passed. B-REQUIRED-CHECKS FAIL on it. Not waived. | `batch3/` (console sha256 `e8c09eaf…a3272`, above) |
| 3b | `de7dc6be` | restricted `platform:browser`, `--project=shell-light --project=shell-dark --grep "S-ACCOUNT"` | PASS 10/10 (S-ACCOUNT-ANON, -UNAVAILABLE, -LINK, -A11Y, -MCP-UNAVAILABLE, both themes). The later identity.spec.ts change does not affect these shell tests. | `batch3b/` |
| 3c | `68b663f4` | whole `ops` suite, `infra/validation/run.sh --base 609048db`, run `11c21a`, 2026-10-21 14:28:16–14:46:32 local | **RESULT: PASSED**, no FAIL. B-IDENTITY-IT (95 tests), B-IDENTITY-E2E (155 tests, 0 skipped/flaky/unexpected; M-RENAME and M-ACCOUNT pass in mcp-light and mcp-dark; every other identity project green), B-IDENTITY-PROBES, B-MCP-E2E (6/6), B-IDENTITY-PRIVACY, B-IDENTITY-CONFIG, B-OBS-CONTEXT/CHAIN/TIMEOUTS/LOGS/PRIVACY, B-OUTAGE, B-PERSIST (migration history version 19), B-SECRET-ROTATION, B-DOCKER-ONLY, B-SECRET-CANARY, B-SCOPE (52 paths in the allow-list) and B-REQUIRED-CHECKS (17 IDs) all PASS; build ledger 25/25, `ac5_status MET`. | `batch3c/console.txt`, sha256 `f9bd8217af86aaa98e45bc09938b22d0a5fe22fed351203860cea2bdb1e2e9f0` |

Deviations:
- (a) Batch 3c ran the whole `ops` suite, not a restricted M- selector. A restricted M- grep cannot be reproduced
  standalone, because identity_e2e seeds the emulator users and `mcp-seed.sql` in the same suite.
- (b) MinIO used the local images that were already present. Nothing was pulled or substituted. In the 3c ledger the
  `minio.Dockerfile` server and init rows have 0 steps executed and are fully cached, with pinned golang, distroless and
  alpine bases.
- (c) Logcheck: the 3c console has no skipped or SKIP line for logcheck. Logcheck ran inside B-OBS-CHAIN (62 cases) and
  B-IDENTITY-PRIVACY (`logcheck validate`), and both passed. The console shows no separate logcheck step being skipped.
- (d) Batch 3b ran at `de7dc6be` and was not repeated at `68b663f4`.
- The consoles do not print the source commit. The `de7dc6be`/`68b663f4` attribution comes from the coordinator's run
  record.

## Batch 2 — private PostgreSQL selectors (S5), source `cec82c82`

Recorded by writer 6 (docs only) after the Codex source review
(CORRECTION LOOP (not included)). The coordinator ran it. Source: a copy of the frozen
`cec82c82` worktree without `.git`; `diff -r` against `git archive cec82c82` is empty. `core/` is unchanged from
`68b663f4` (the rebase onto `8dd30f8c` changed only staging's docs/management files and `appearance.spec.ts`). Setup:
private Compose project `platform-g063-db4`, 0 published ports (asserted from the effective config), private MinIO
with the scoped service credential, a `core-check` image built from that source (single build slot). Migrate CLI
up/down/re-up exit 0 at versions 19/18/19 (`db4/0-migrate-{up,down,reup}.log`). Every selector ran as
`go test -tags integration -count=1 -v -run <regex> <pkg>`. Evidence root
`<home>/.local/state/platform-evidence/g063-a1/`; script, driver log and status table are under
`<home>/.claude/jobs/580cdb2f/tmp/g063-db/` (`run-g063-db4.sh` `a6f230a6…`, `driver4.log` `db9a8f5c…`,
`status4.tsv` `26166a82…`). The counts below come from the `--- PASS/FAIL/SKIP` lines in each log.

| # | Package, `-run` | Environment | Exit | Top / sub PASS (FAIL, SKIP) | Log, sha256 |
|---|---|---|---|---|---|
| 1 | `./internal/migrate/` `^(TestMCPCatalog\|TestMCPMigration\|TestMCPNamesCatalog\|TestMCPNamesMigration\|TestAuditCatalog\|TestMigrationLifecycle)$` (the Plan's five + `TestMigrationLifecycle`) | object env | 0 | 6 / 33 (0, 0) | `db4/1-migrate.log` `3887cf2e3cede48197f44265942cb128695f29f6db5dfd930d59712f3970caa6` |
| 2 | `./internal/mcp/` `^(TestMCPTransport\|TestMCPLifecycle\|TestMCPAudit\|TestMCPRecovery)$` | object env + `MCP_LOG_FILE`, `MCP_INVALIDATE_SQL_B64` | 0 | 4 / 14 (0, 0) | `db4/2-mcp.log` `353dbaa7dce01e2f273c22e4df747b8a4c86d76f2eb4b8988136782566e7bbbb` |
| 3 | `./internal/identity/` `^TestMCPIdentity` | plain IT env | 0 | 5 / 0 (0, 0) | `db4/3-identity.log` `2a14239a1d1236a2b3646ed851bb6784a87928944a39ce967349107ba1435cc4` |
| 4 | `./internal/tenancy/access/` `^(TestMCPOrdering\|TestMCPFootprint)$` | object env | 0 | 2 / 11 (0, 0) | `db4/4-mcp-facade.log` `901e7a31f3619e23684a5e73cdfac2bb10c3c489f375d39c5c0b899e8554f26c` |
| 5 | `./internal/tenancy/access/` `^TestAudit$` | B-AUDIT-IT stored-object env (private MinIO, scoped credential) | 0 | 1 / 9 (0, 0) | `db4/5-audit.log` `ff185b0a9d8220901a16a17b5e853fefce072050c6b11e344149b7ee12280852` |

Named sub-results, as printed in the logs:
- Selector 1: `CT-MCPN-1`, `-CHECK`, `-DML`, `-NEG` and `MG-MCPN-1`..`4` all PASS.
- Selector 2: `IT-MCP-3_setup_and_activation` and `IT-MCP-15_rename` PASS; `TestMCPRecovery` PASS. Selector 3:
  `TestMCPIdentityRename` PASS.
- Selector 4: `O-MCP-9_a_rename_and_a_revocation_serialize_on_the_target_session_in_both_orders,_without_a_deadlock`
  PASS, with `ORDERING {"case":"O-MCP-9","order":"user,session,target,user,session,facts,row,write"}`. This is one
  subtest (`access/mcp_integration_test.go:574`). Its source covers (a) rename first (the revocation waits; both commit;
  no 40P01; liveness unchanged; the statement order is asserted), (b) revocation first (`not_found`, nothing written) and
  (c) a held, uncommitted session revocation (the rename waits at the target session, not the row; no deadlock;
  `not_found` after the commit).
- Selector 5: `AT-ATOM_same-transaction_writes,_strict_faults_and_revocation_priority` PASS. Its source includes the
  rename audit-fault rollback: the refused rename leaves the facts and `xmin` unchanged with no event, and the retry's
  event `xmin` equals the row's. `AT-INV_every_catalog_action_observed_from_real_calls` PASS, with
  `AUDIT-INV {"actions":47,"observed":47}`. The script does not run the harness's logcheck/canary post-processing.

**Selector 6, extra to the Plan (`TestTenancyModel`, `TestBusinessProfileModel`, for version-19 harness parity).**
- **db4: FAILED, preserved as a failed attempt.** Both exited 1 (`status4.tsv`). This is a script run-order artifact,
  not a product defect. The script ran them on the database that selector 5's `TestAudit` had populated:
  - `TestTenancyModel` started at version 19 and failed at `migrate down to 6: … refusing to roll back 00007:
    app.stored_objects holds rows`.
  - The script's migrate-up before `TestBusinessProfileModel` then failed with `legacy owner preflight` (at 00015).
    That test started at version 14 and failed with the same 00007 refusal.
  - The harness's `migrations` suite runs both before `audit_it` (`run.sh` :2568–2571 vs :2593).
  - Logs and sha256: `db4/6-TestTenancyModel.log` `b8048cfe…a86dc`; `db4/6-TestBusinessProfileModel.log`
    `162adf87…3873e`.
- **db5: rerun on a fresh database. Both PASS (exit 0).** Same source copy, project `platform-g063-db5`, script
  `run-g063-db5.sh` `6691a0f3…`, `driver5.log` `26268e95…`, `status5.tsv` `32839194…`. The migrate CLI again ran
  19/18/19.
  - `TestTenancyModel` started at version 19 and passed 1/37 (0, 0): `db5/6-TestTenancyModel.log`
    `0e71ddd9bac9f6f9e25b3cb0c81887c0c7ebc8f1f0e36f663d661e262d5aa8cd`.
  - `TestBusinessProfileModel` passed 1/35 (0, 0), including `IT-BP-REF-2`: `db5/6-TestBusinessProfileModel.log`
    `e6fe579c77bafb091e3f6fe19a2cdc1ebc79a29db16aaa425cc1ac3f7af6a691`.
  - Caveat: in db5 the script's migrate-up before `TestBusinessProfileModel` again failed with `legacy owner preflight`
    (the rows `TestTenancyModel` leaves). So that test started at version 14, not 19. The script does not record that
    pre-step's exit. Only `TestTenancyModel` exercised the version-19 starting point.

**Cleanup (coordinator-verified).** No `platform-g063-db*` container or volume remains. The source copy's `.env` and the
IT env files are removed (re-checked: absent). The script removed the `core-check` images (`down.log` shows the db5
removal). No product code or test changed in this round.

## S5 ledger and S7 status

- **S5 batch 1 (static/unit, harness `lint`/`unit-*`).** The first run (`batch1/`) FAILED on B-IDENTITY-GUARDS; it was
  fixed (guard fix above). Reruns `batch1b/` (all six suites), `batch1c/` (lint, lint-base) and `batch1d/` (lint,
  lint-base, unit-core) all report `RESULT: PASSED` (consoles `9675c541…`, `25adfa52…`, `059e1634…`). The consoles do
  not print the source commit.
- **S5 batch 2 (private DB).** Selectors 1–5 PASSED at `cec82c82` (above). Extra selector 6: FAILED in db4 (run order),
  PASSED in db5.
- **S5 batch 3.**
  - 3a (`ops` at `de7dc6be`): FAILED, kept on record above.
  - 3c (`ops` rerun at `68b663f4`, after the `identity.spec.ts` fix): PASSED.
  - 3b (S-ACCOUNT at `de7dc6be`): 10/10.
- **Not run pre-MR.** The whole `migrations` umbrella suite, full B-IDENTITY-E2E and B-BROWSER-ROOT. They remain in the
  merged-SHA full qualification (Plan §6), which does not stand in for the batches above.
- **S6.** No separate self-review section is written in this Result. The writer reviews are in the sections above.
- **S7.** Each round fetched, merged or rebased onto staging, reconciled and pushed a safekeeping checkpoint; this
  round rebased onto current `origin/staging`. The final AC mapping and cycle counts are not yet written. (Completed in
  the section "Final acceptance-criteria mapping and cycle accounting" below.) Correction
  cycles so far: 2 (correction 1: the 00019 postcondition; correction 2: E-LOGIN) plus this docs-only ledger round. All
  within B3 attempt 1.
- **Correction 3 (Duo finding on MR !77; tests only).** The escaped surrogate-pair vectors had become the literal U+1F600
  character (the escape text was rewritten on write), so the accepting path of `surrogatesPaired` was never exercised.
  Corrected in `G063: restore escaped surrogate-pair test vectors (Duo finding, tests only)`: `mcp_test.go`
  TestSetupNameDecoding ("escaped surrogate pair" and "upper-case pair" wire cases; lower- and upper-case pair entries
  in the helper table, negatives unchanged) and `mcp_integration_test.go` IT-MCP-15 (the request body; the stored value
  is still asserted as the decoded character). Offline `gofmt -l`, `go vet` (plain and `-tags integration`) and
  `go test -run '^(TestSetupNameDecoding)$' ./internal/mcp/` pass; a scratch mutant refusing valid low surrogates fails
  the four wire cases and both helper entries. Not committed to this ledger: the private-DB rerun, which the coordinator owns.
- No Codex outcome is claimed. Pre-integration and final ACCEPT remain Codex's.

## Post-merge qualification repair (branch `g063-repair` from `origin/staging` ef2c0181; diagnosis from retained evidence only)

The post-fix full qualification of merged staging `36d1189b` (coordinator run `qual-36d1189b`, retained candidate
evidence) failed two checks; run 1 (`a490bf37`) had neither. Scope here is exactly those two. The repair worker built
and ran nothing; the coordinator's focused runs are recorded under "Focused evidence" below.

- **B-CI-CONTROLS, control VS3** (`FAIL VS3 an available observation is not ok;a size is not a Docker CLI size;`,
  `ci-controls.log:185`; 523/524 pass; VS3 took 8 s per `ci-controls-timings.tsv`). Both problems come from one
  cause: `expect_storage ... ok` (`validate-controls.sh:3620-3621`) flags a daemon row whose status is not `ok` and
  whose value is the literal `unavailable` (which also fails the size grammar). `storage_observe`
  (`validate.sh:1038-1048`) writes exactly that when one `docker system df` query exits non-zero, returns no
  `Images`/`Build Cache` row, or a malformed figure; the published `storage.tsv` of the run itself is clean (all 14
  rows `ok`, sizes `56.41GB`, `21.41GB`, ... accepted by both grammars), so the size regex is not the issue. The
  control's own `$C/ev/storage.tsv` was scrubbed with its case, so the exact status string is not retained; the 8 s
  duration excludes the 60 s bound (exit 124). VS3 ran at about 18:14:50 local (cumulative timings from the
  background start 17:27:33), while the foreground harness was between B-SECRET-ROTATION (18:14:06) and
  B-IDENTITY-IT (18:16:15), which builds and removes images on the same Docker Desktop daemon (27.5.1, containerd
  store). Docker's release notes record `docker system df` failing while images are pruned concurrently under the
  containerd image store (moby/moby#52672). **Hypothesis, not proven from the retained bytes:** a single transient
  `docker system df` failure under concurrent foreground image removal; the producer reads it as
  `unavailable:docker-system-df-exit-N` with no retry, and VS1-VS3 require `ok` from the real shared daemon.
  Repair: `validate.sh storage_observe` retries a failed query twice, 2 s apart (not after a bounded-out/killed
  query; the last exit is the reported status; `unavailable`, never zero, is unchanged); `validate-controls.sh`
  `expect_storage` now names the offending rows (`phase:metric=value[status]`) so a recurrence is attributable;
  VS4 additionally asserts a persistent failure is queried exactly 3 times per phase (bounded); new control VS5
  (an injected synthetic transient failure, shim `SHIM_FAIL_DF=transient`: every odd-numbered query fails; it does
  not reproduce a concurrent daemon mutation) asserts both phases read `ok` after one retry each, 4 queries in all.
  Group E is VS1-VS5, and `docs/foundation/OPERATIONS.md:415` now lists VS1-VS5 (Codex amendment 05).
- **B-OBJ-IT, subtest OBJ-ADMIN** (`objects_integration_test.go:1542`:
  `Put "http://minio:9000/saas-foundation-objects/objects/anonymous": EOF`, 0.01 s; the other 9 subtests, footprint,
  attack and oracle lines pass). The anonymous probes used `http.DefaultClient`, whose `DefaultTransport` pool is
  shared with OBJ-SIGN's `&http.Client{Timeout: 10s}` (no Transport; `:1370`), which reads its fetched bodies to EOF
  and so leaves keep-alive connections to `minio:9000` idle in that pool seconds before OBJ-ADMIN. A plain `EOF`
  from `Client.Do` is net/http's `transportReadFromServerError` unwrapped: the server closed the connection before
  any response byte; net/http replays that only for replayable requests (GET/HEAD/OPTIONS/TRACE, or an
  Idempotency-Key), so the two anonymous GETs self-heal and only the PUT (a body) can surface it, which is exactly
  the one failure observed; the probes also iterated a map (random order). The exact MinIO-side trigger for closing
  the pooled connection (idle timeout, or a close after an earlier denied request) is **not proven** from the
  retained logs (no MinIO server log is retained). Repair (tests only): the admin-API and anonymous probes use a
  dedicated client with `DisableKeepAlives: true` (every probe on its own connection; a fresh connection that ends
  without a response remains `t.Fatal`, never a denial), in a fixed order (list, get, put) with bodies drained; and
  the test now also asserts the denied PUT wrote nothing: `HeadObject objects/anonymous` with the scoped credential
  (s3:GetObject on the bucket's keys) must be 404. Every previous assertion (403 for every scoped bucket operation,
  the admin API and the three anonymous requests) is kept; no retry, no EOF-as-success.
- Changed files: `infra/ci/validate.sh`, `infra/ci/validate-controls.sh`,
  `core/internal/tenancy/access/objects_integration_test.go`, this Result; correction round 1 adds
  `docs/foundation/OPERATIONS.md` (catalog VS1-VS5) and the inventory's Codex amendment 05, which admits exactly
  these paths for this repair (review record `management/v0/results/G063-repair1-codex-review-20261021.md`).
  Amended inventory SHA-256: `1a1e447ab57870e87968735a210aac861b35aaf0d02841d9fe33bc0e7a1bb49f`.
  Correction round 1 changed only comments, the VS5 PASS-description string, this Result, the inventory and the
  OPERATIONS.md catalog line; no logic or assertion changed.

### Focused evidence (coordinator runs)

Focused source: exact SHA `1ece9fcaddb8d784e76437dccc3b34d5fade4875`, frozen checkout
`.claude/worktrees/g063-repair-chk` (clean). Evidence directory:
`<home>/.local/state/platform-evidence/g063-repair-a1/`.

- **VS3, VS4, VS5:** `bash infra/ci/validate-controls.sh --only VS3,VS4,VS5` printed
  `ci-controls: 3 passed, 0 failed (of 3 selected; 525 defined)`, exit 0. `vs345.console.txt` SHA-256
  `e81ea13129aec89fa6a34113b9e5e5621bce96b02e562b3b9a878fa7df91dd12` (693 bytes); `vs345.log` SHA-256
  `5d1d3c4f0004426c4c2664f8a2251c92fbbb1fded7fe021202647c9e4d9fe3e8` (618 bytes).
- **Exact-source `TestObjects`:** `go test -tags integration -count=1 -v -run '^TestObjects$'
  ./internal/tenancy/access/` in the harness's core-check image, against a private PostgreSQL (migrations applied up
  to version 19) and a private MinIO with the scoped credential; Compose project `g063r-20261021230522-84617`, no
  published host ports, `--pull never`. Exit 0, `--- PASS: TestObjects (12.06s)`, 10/10 subtests PASS including
  OBJ-ADMIN (0.01s). `obj-it/obj-it.log` SHA-256 `03813df8106f5d67f72027a7b9353340203ee07e997420c151c148052f0c099a`
  (222961 bytes); `obj-it.console.txt` SHA-256 `0c5574f9cebc90ff45c830e077ed94d85212f4f7a09e23c5c196c561db56428c`.
  Cleanup (`obj-it/cleanup.log` SHA-256 `7b275ce2b9ca9af3cd70126b29647cfe3394b83c8abaccd93305de59e885cfc7`): 0
  containers, 0 volumes, 0 networks, 0 images and 0 `g001.run` containers left. The driver was a coordinator-written
  disposable script, not repository code: it sourced `infra/validation/lib.sh` and the `rls_go`, `obj_minio`,
  `obj_env` and `new_canary` definitions of `infra/validation/run.sh` (SHA-256
  `6f88838084e91e6fe6773820e1095219885956d68431dd12d4dadb07ebac06ae`). It is not the full B-OBJ-IT check:
  B-OBJ-IT's `TestObjectsFootprint`, `TestIsolationAttacks` and the OBJ-SIGN-HOST host probe were not run in this
  focused run.
- **Failed driver attempt `obj-it-driver-defect-run0`** (Compose project `g063r-20261021230359-82908`): a failed
  driver attempt, not a test failure. The coordinator driver exited 127 before any `go test` ran
  (`SECRETS_FILE: unbound variable`, `new_canary: command not found`), so there is no `TestObjects` result from it.
  Its cleanup shows rc=127 and 0 containers, volumes, networks, images and `g001.run` containers left
  (`obj-it-driver-defect-run0/cleanup.log` SHA-256
  `f1ddc67517bde51e7b5be0f6683191cd0dfd68e1d8b0e825e0c91ec275dbe025`). The driver was corrected (defined
  `SECRETS_FILE` and `new_canary`, added log redaction) and rerun at the same SHA; this is part of the same focused
  attempt, not a code failure.
- **Root causes remain unproven hypotheses.** VS3: a transient `docker system df` failure read as unavailable.
  OBJ-ADMIN: a stale pooled keep-alive connection closed by the server, surfacing as EOF on the non-replayable PUT.
  Neither original failure was reproduced on `36d1189b` or `ef2c0181` in this session. The passes show that the
  repaired code behaves as designed in these runs, not that the original flakes cannot recur.
- Not claimed: no B-CI-CONTROLS suite run, no G063 full-suite result and no full qualification. The failed full run
  of `36d1189b` above stands as failed. No Codex outcome is claimed.

### Merged-SHA focused verification (MR !78, staging `feb1f991`)

- **Integration:** MR !78 (`g063-repair-mr`, source `2e77dce2fc0862a09be8424dca3372c9bcacae09`, rebased onto staging
  `8def52c4`; two commits, content identical to the Codex-accepted `9e661fd6`) merged into staging as
  `feb1f9910e2eb76845ba6cd5e568c0e307c88eae` (parents `8def52c4` and `2e77dce2`; tree
  `72936a35046fe2d69bc624fbb1f3422efbb3c1d5`, equal to the source tree; source branch deleted). Hosted MR pipeline
  `2974206829` finished `success` (the fast job; one manual job was not run). GitLab Duo approved with no findings;
  open threads 0. Pre-integration review: `management/v0/results/G063-repair1-preintegration-codex-review-20261021.md`.
- **Exact merged-SHA checks** (frozen checkout `.claude/worktrees/g063-staging-chk`, HEAD `feb1f991`, clean before and
  after; evidence `<home>/.local/state/platform-evidence/g063-repair-a1/staging-feb1f991/`):
  - `bash infra/ci/validate-controls.sh --only VS3,VS4,VS5`: `ci-controls: 3 passed, 0 failed (of 3 selected; 525
    defined)`, exit 0. `vs345.console.txt` SHA-256 `78cedc04758c6e43635034a13c5a4a9c9caac8b88403ea08879cb8f00bd4640c`.
  - Exact-source `TestObjects` (`go test -tags integration -count=1 -v -run '^TestObjects$'
    ./internal/tenancy/access/`, same private PostgreSQL + MinIO driver as above, no published host ports,
    `--pull never`, Compose project `g063r-20261021233421-18299`): driver exit 0, `--- PASS: TestObjects (11.26s)`,
    10/10 subtests PASS including OBJ-ADMIN. `obj-it.console.txt` SHA-256
    `68f1c5cc3713e6687d5bc409526d3b51b682df41d3ec288b3e972c265d8919c2`; `obj-it/obj-it.log` SHA-256
    `33598271433bda92d1947f16ebfd85db1b082fb61f56a470f6289e204926092d`.
  - Cleanup (`obj-it/cleanup.log` SHA-256 `7bf00863b1c3cafff84ae436078e82bec5f0a19d6f8da717bf319a11f984340d`): rc 0;
    0 containers, 0 volumes, 0 networks, 0 images and 0 `g001.run` containers left.
- **Scope of this verification:** focused only, per the owner disposition (no dedicated G063 full suite). Not run:
  `TestObjectsFootprint`, `TestIsolationAttacks`, the OBJ-SIGN-HOST host probe, the B-CI-CONTROLS suite and any full
  qualification. The failed full run at `36d1189b` stands as failed. The two root causes remain unproven
  hypotheses; these passes show the merged repaired code behaves as designed, not that the original flakes cannot
  recur. No final G063 acceptance is claimed here; that is Codex's decision.

### Acceptance cleanup disposition (coordinator record; Codex final ACCEPT still pending)

- **First cleanup list (answered "Approve all as listed" through the session's native prompt; Codex reports it selected
  that answer under its relay of the standing owner authorization; the coordinator could not itself attribute the
  responder; completed 2026-10-21):** 15 finished G063 worktrees removed (the frozen
  qualification and verification checkouts `g063-full`, `g063-full-fix`, `g063-qual` to `g063-qual7`,
  `g063-repair-chk`, `g063-repair-r1-chk`, `g063-repair-mr`, `g063-staging-chk`, and the two repair-worker agent
  worktrees); local branches `g063-repair`, `g063-repair-c1` and `g063-repair-mr` deleted; remote branches
  `g063-plan`, `g063-plan-r1`, `g063-repair` and `g063-repair-r1` deleted; the seven `platform-g063-3b-*` images
  removed. No G063 container, volume or network remains.
- **Unique-history bundle (verified before any deletion):**
  `<home>/.local/state/platform-evidence/g063-closeout/g063-unique-history.bundle`, 14 heads, 18506605
  bytes, SHA-256 `3b8103de3270a1a299e826c097e9f0f53fba46bf34be539520bdbc80be3ed4dc`; receipt
  `g063-closeout/cleanup-receipt.txt`.
- **Held outside the approved scope (preserved, not deleted):** ten finished G063 planner/writer agent worktrees
  (`agent-<id>`, `agent-<id>`, `agent-<id>`, `agent-<id>`,
  `agent-<id>`, `agent-<id>`, `agent-<id>`, `agent-<id>`,
  `agent-<id>`, `agent-<id>`; all clean and unlocked) and the ten local branches they have
  checked out (`g063-impl`, `g063-impl-w2` to `g063-impl-w8`, `g063-plan`, `g063-plan-r1`). The coordinator's first
  list missed these worktrees. The native prompt that asked about them was answered "Hold"; this is recorded as a
  Codex disposition under the existing owner-authorized scope (the items were outside the exact first list), not
  as a separate owner reply. Codex directed that they be preserved and the scope boundary recorded. Their history
  is included in the bundle above. The `worktree-agent-*` branches are held until the owner-approved RC7 cleanup
  point.
- **Preserved by design:** `g063-result-final` and MR !79 until merged; evidence directories `g063-a1` (including the
  failed `36d1189b` full run) and `g063-repair-a1`.
- **Codex final G063 ACCEPT is still pending.** This cleanup record claims no acceptance.

### Cleanup receipt, second list (supersedes the "held" bullet above; Codex final ACCEPT still pending)

- **Authority:** Codex asked (handoff 5c632937) that the ten held items be removed under the mandatory task-acceptance
  cleanup policy. The coordinator did not treat that relay as approval; it put the question to the session's native
  prompt again, which was answered "Delete the 10 worktrees + 10 branches" (the responder is not attributable by the
  coordinator). Executed 2026-10-21.
- **Checks before removal:** the unique-history bundle verified again; each of the ten branch tips is contained in
  the bundle (`g063-impl` `3929788b`, `g063-impl-w2` `d0f2872d`, `-w3` `de7dc6be`, `-w4` `68b663f4`, `-w5` `cec82c82`,
  `-w6` `952e9fe3`, `-w7` `47a5d70f`, `-w8` `60d3794d`, `g063-plan` `bea8b161`, `g063-plan-r1` `f90e0d88`); every
  worktree was clean and unlocked and none was live.
- **Removed:** the ten worktrees `agent-<id>`, `agent-<id>`, `agent-<id>`,
  `agent-<id>`, `agent-<id>`, `agent-<id>`, `agent-<id>`,
  `agent-<id>`, `agent-<id>`, `agent-<id>`; and the ten local branches
  `g063-impl`, `g063-impl-w2` to `g063-impl-w8`, `g063-plan`, `g063-plan-r1`. No failures.
- **Remaining by design:** the `g063-result-final` worktree, local branch and remote branch (MR !79, until merged);
  the 39 `worktree-agent-*` local refs (held until the owner-approved RC7 cleanup point, not touched); the evidence
  directories `g063-a1` and `g063-repair-a1`; the bundle `g063-closeout/g063-unique-history.bundle` (SHA-256
  `3b8103de3270a1a299e826c097e9f0f53fba46bf34be539520bdbc80be3ed4dc`) with its receipts
  `cleanup-receipt.txt` and `cleanup-receipt-2.txt`. After this cleanup the only G063 worktree is
  `g063-result-final` and the only local `*g063*` branch is `g063-result-final`.
- **Codex final G063 ACCEPT is still pending.** This receipt claims no acceptance.

## Final acceptance-criteria mapping and cycle accounting (S7)

Docs only (closeout writer, Opus 5.5, runtime effort unknown). **No test, build or harness was rerun for this
section.** It maps Spec AC1–AC8 (Plan r1 §11) to evidence already recorded above or in the cited records. Claims no
acceptance: final G063 ACCEPT is Codex's decision. Classes: **(a)** executed, attributable PASS with exact source;
**(b)** only a FAILED run or a partial/focused run (what ran and what did not is stated); **(c)** not evidenced.

**Source lineage (verified at this round with `git diff --stat`, read-only).** The Result's pre-rebase SHAs and the
merged chain differ only as follows. `cec82c82` (batch 2) and the rebased `47a5d70f` have identical `core web infra
contracts docs`. `68b663f4` (batch 3c) → `47a5d70f` adds only staging's `appearance.spec.ts`, `WEB-SHELL.md` and
`GITLAB-BRANCH-WORKFLOW.md` changes. `47a5d70f` → `60d3794d` is correction 3 (`mcp_test.go`,
`mcp_integration_test.go`, this Result). The MR !77 merge `a490bf37` (second parent `60d3794d`) equals `60d3794d` on
those paths. `36d1189b` equals `a490bf37` on `core web contracts`; it adds the staging retention fix in `infra/ci`.
`feb1f991` adds the repair (`objects_integration_test.go`, `validate.sh`, `validate-controls.sh`, OPERATIONS.md) and
staging's `docs/PERFORMANCE.md`.

**Failed full run at `36d1189b` (preserved as FAILED).** Console `g063-a1/qual-36d1189b.console.txt` (sha256
`9ed9b4ccb0aa4122061bea11d12e9d0f413d54b23dc2c1456c64bd33b124fd22`, `tested_content=local head=36d1189b…`, final status
1). Its per-check lines, read from that retained console at this round (not previously quoted in this Result):
119 PASS, 3 FAIL (B-OBJ-IT, B-CI-CONTROLS, B-REQUIRED-CHECKS), 7 INFO. PASS lines include B-MCP-IT, B-MCP-CATALOG,
B-MCP-SOURCE, B-MCP-HTTP, B-MCP-E2E, B-IDENTITY-IT, B-IDENTITY-E2E, B-IDENTITY-GUARDS, B-IDENTITY-PRIVACY,
B-BROWSER-ROOT, B-MIGRATE-SET/CLI/CATALOG/IT, B-AUDIT-CATALOG/IT, B-RLS-IT, B-WEB-BOUNDARY, B-TOGGLE-OFF, B-PERSIST
and B-OBS-PRIVACY. A check PASS inside a failed run is not a qualification pass. Per the owner disposition
(focused closeout (not included)) there is **no dedicated G063 full
suite**; the two failures were repaired by MR !78 and verified only by focused runs at `feb1f991`.

| AC | Evidence (slice / test ID / check ID → reference) | Class |
|---|---|---|
| AC1 shared manager, account intact, unavailable isolated | S3 `connection-manager.tsx`, `account-sections.tsx`, both pages. M-ACCOUNT, M-SETUP in B-IDENTITY-E2E batch 3c (`68b663f4`, run `11c21a`, 155 tests, `batch3c/console.txt` `f9bd8217…e9f0`). S-ACCOUNT-MCP-UNAVAILABLE in batch 3b (`de7dc6be`, 10/10, `batch3b/`). B-UNIT-WEB `unavailable` reduction in batch 1b. Full run: B-IDENTITY-E2E and B-BROWSER-ROOT PASS inside the failed `36d1189b` run. | (a) focused; caveat: consoles do not print the commit (coordinator run record) and 3b was not repeated at `68b663f4` |
| AC2 optional names, rename/clear, legacy fallback, client display | S1/S2 store and route, S3 client. IT-MCP-3, IT-MCP-15 (batch 2, `cec82c82`, `db4/2-mcp.log` `353dbaa7…bbbb`). `TestNormalizeMCPDisplayName` and mcp-connections unit (batch 1b; commit not printed). M-SETUP, M-RENAME (3c). | (a) at `cec82c82`/`68b663f4`. After correction 3 changed the IT-MCP-15 request body, the private-DB rerun is not recorded here; B-MCP-IT PASS only inside the failed `36d1189b` run |
| AC3 trim, 80 code points, non-BMP, refusal, escaped display, bounds | Shared vector table (core and web unit, batch 1b); `TestSetupNameDecoding`, IT-MCP-15 (batch 2), mcp-forms (batch 1b); M-RENAME hostile markup (3c); list bound pins (§ Mechanical pins). Correction 3 (Duo, MR !77): before it the accepting surrogate-pair path was never exercised. After it: writer's offline `go test -run '^(TestSetupNameDecoding)$'` pass plus a mutant check (source SHA not recorded); B-MCP-IT PASS inside the failed `36d1189b` run. | (b) the corrected accepting path has only a focused offline run (no SHA) and a check pass inside a failed run; refusal paths are (a) at `cec82c82` |
| AC4 session/CSRF/host-bound, own-user/current-Org, non-disclosing | IT-MCP-15 negatives (dave in both tenants, alice vs bob, stale/wrong session, CSRF/Origin, duplicate/malformed/null) batch 2 `cec82c82`; `TestSetupNameDecoding` refusals and mcp-forms unit (batch 1b); `TestMCPScopeBinding` (unit-core, batch 1b). | (a) at `cec82c82`; unit attribution caveat as AC1 |
| AC5 metadata-only rename, no revival, isolation | Batch 2 `cec82c82`: IT-MCP-15 fact comparison, `TestMCPRecovery`, `TestMCPIdentityRename` (`db4/3-identity.log` `2a14239a…35cc4`), O-MCP-9 three orders with asserted statement order (`db4/4-mcp-facade.log` `901e7a31…f26c`), AT-ATOM rename fault rollback (`db4/5-audit.log` `ff185b0a…0852`), CT-MCPN-1 guard/policies unchanged (`db4/1-migrate.log` `3887cf2e…caa6`). | (a) |
| AC6 alias, stored Address, 100-row cursor, page-specific returns | IT-MCP-15 old-host row via current host (batch 2); M-ACCOUNT pagination, `mcp_cursor`, `return_to`, `invalid_page` isolation (3c); mcp-forms `return_to` closed set (batch 1b). `TestCursor` rerun: in unit-core batch 1b by Plan §6; per-test line not quoted in this Result. | (a) for IT/browser; `TestCursor` per-test result not recorded here |
| AC7 additive migration, least privilege/RLS, bounded wire/log/audit | S1 `00019` + correction 1. CT-MCPN-1/-CHECK/-DML/-NEG, MG-MCPN-1..4, `TestAuditCatalog`, `TestMigrationLifecycle` (batch 2 selector 1); AT-INV 47/47 (selector 5); migrate CLI 19/18/19 (db4, db5). `TestAccessBoundary`/`TestRLSBoundary`, `audit_test.go` 47, B-MCP-SOURCE, B-WEB-BOUNDARY, B-MIGRATE-SET (batch 1b–1d). Log/name absence: B-IDENTITY-PRIVACY, B-OBS-PRIVACY PASS in 3c (batch 2 ran no logcheck/canary). Whole `migrations` umbrella: not run pre-MR; B-MIGRATE-IT PASS only inside the failed `36d1189b` run. | (a) focused; umbrella (b) |
| AC8 focused checks, a11y, reviews, merged-SHA full qualification, cleanup | Focused batches 1–3 above; M-A11Y (3c) and S-ACCOUNT-MCP-UNAVAILABLE 1280/320 light/dark (3b). Reviews: Codex source CORRECTION LOOP then pre-integration ACCEPT at `952e9fe3` (review (not included)); correction 3 (Duo finding) landed after that ACCEPT, and its Codex review is not recorded here. MR !77 hosted pipeline id: not recorded here. Merged-SHA full qualification: run 1 FAILED, run 2 cancelled, run 3 FAILED (see cycles). Repair MR !78: Codex review + correction round, pre-integration ACCEPT, pipeline `2974206829` success (fast job; one manual job not run), Duo approved, focused VS3/VS4/VS5 3/3 and `TestObjects` 10/10 at `feb1f991`. Cleanup receipts 1 and 2 above. | (b) the merged-SHA full qualification FAILED and, by owner disposition, is not repeated; final ACCEPT pending |

No AC is class (c) in the sense of having no executed evidence. Unproven: the VS3 and OBJ-ADMIN root causes remain
hypotheses. Not run on any G063 SHA after the repair: `TestObjectsFootprint`, `TestIsolationAttacks`, the
OBJ-SIGN-HOST probe, the B-CI-CONTROLS suite and any full profile.

**Cycle accounting (from this Result's records).**
- **Goal-level attempts:** 1 of B3 (no reset, no extension; the attempt-2 checkpoint was not reached). Every round
  below is an inner loop of attempt 1.
- **Writers:** 1 (S1–S2), 2 (S3–S4, amendments 02/03/04, guard fix), 3 (correction 1), 4 (correction 2), 5 and 6
  (docs-only ledgers), the correction-3 writer (not named in this Result), the repair worker (two agent worktrees;
  repair plus its Codex correction round 1) and closeout docs writers on `g063-result-final` (MR !79).
- **Implementation correction passes:** 3. Correction 1 (00019 postcondition boolean text), correction 2 (E-LOGIN
  CSRF count) and correction 3 (Duo surrogate-pair vectors, tests only). Plus one Codex source-review CORRECTION LOOP
  (validation/ledger only; no product change). Post-merge: one repair (MR !78) with one Codex correction round
  (comments, PASS text, docs, inventory only).
- **Allow-list amendments:** 02, 03, 04 (pre-MR) and 05 (repair). No amendment 01 is recorded here.
- **S5 batches:** batch 1 FAILED (B-IDENTITY-GUARDS), reruns 1b/1c/1d PASSED. Private DB: first batch at
  `d0f2872d` FAILED (00019 postcondition → correction 1); db4 selectors 1–5 PASS at `cec82c82`, extra selector 6
  FAILED in db4 (run order), PASS in db5. The evidence root also holds `db`, `db2`, `db3` and `db6` directories that
  this Result does not describe (not recorded here). Batch 3a FAILED, 3b PASS, 3c PASS.
- **Post-merge full qualification:** run 1 at `a490bf37` FAILED (console `postmerge-a490bf37-run1-FAILED.console.txt`
  sha256 `0e7f2a34b71f83486c69c73ee92fa73301cde5693d6aec24887c94306acfa092`: B-CI-CONTROLS and B-REQUIRED-CHECKS FAIL,
  B-OBJ-IT PASS, final status 1; nothing published, so the failing control ID is not retained or recorded here).
  Staging then took the owner-authorized retention fix `36d1189b`. Run 2 at `a490bf37` was cancelled by the owner
  (console `postmerge-a490bf37-r2.console.txt` sha256
  `a05504741c67e6f768103607932737311f67c7efd030fa3155ad5ef4c45d4b0a`, status 130, signal). Run 3 at `36d1189b`
  FAILED (VS3, OBJ-ADMIN; above).
- **Repeated material issue beyond two cycles (CLAUDE.md):** none derivable. Corrections 1–3 and the two repaired
  failures are distinct defects. B-CI-CONTROLS failed in runs 1 and 3, but run 1's control is not retained, so
  whether that is the same issue cannot be derived (the Result above states run 1 had neither run-3 failure).

**Bindings.**
- Repair staging merge `feb1f9910e2eb76845ba6cd5e568c0e307c88eae` (MR !78; tree `72936a35…c1d5`).
- Feature merge `a490bf37bb7742411f7da6e8d9ed6f8a9db09f5a` (MR !77; source tip `60d3794d`).
- Failed full-run SHA `36d1189bf252ef97504ef2440bd2ec0a8a424bef`; evidence
  `<home>/.local/state/platform-evidence/g063-a1/qual-36d1189b/` and its console (sha256 above).
- Pre-merge evidence root `<home>/.local/state/platform-evidence/g063-a1/` (batch hashes as quoted above).
- Focused repair evidence `<home>/.local/state/platform-evidence/g063-repair-a1/` (`1ece9fca`:
  `vs345.console.txt` `e81ea131…dd12`, `obj-it/obj-it.log` `03813df8…099a`) and `…/staging-feb1f991/`
  (`vs345.console.txt` `78cedc04…640c`, `obj-it/obj-it.log` `33598271…092d`, `obj-it/cleanup.log` `7bf00863…340d`).
- Unique-history bundle `<home>/.local/state/platform-evidence/g063-closeout/g063-unique-history.bundle`,
  sha256 `3b8103de3270a1a299e826c097e9f0f53fba46bf34be539520bdbc80be3ed4dc`.
