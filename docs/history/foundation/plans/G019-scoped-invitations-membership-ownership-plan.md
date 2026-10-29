> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G019 Implementation Plan — Scoped invitations, membership and ownership administration

Revision **r1** (provisional; bounded correction of r0 under the Codex review
`management/v0/results/G019-plan-codex-review-r0.md`, SHA-256
`0f2f57881e031fe73e550432c0089c0e57ff134d72208bea80962fad9f28302d`, verified on `origin/staging` `9cf9beed`).
2026-10-17 UTC. Author: Claude (model **Fable 5.1**; planning effort "high (requested)" — the runtime setting is not
verifiable from inside the worker, actual value **unknown**). Worktree: generated name `agent-<id>`
(recorded per AGENTS "Task-prefixed Claude worktrees"; not renamed; task G019, role plan r1; the r0 worktree
`agent-<id>` holds the local `g019-plan` checkout and was not touched). Branch: local `g019-plan-r1`
from `origin/g019-plan` `dbbfea2b83edbb7fe8f2f17fd4c9aa73de9d1c72` (r0) merged with `origin/staging`
`9cf9beed16c96a94d6254ef2108a91e35fd1194d`, pushed to `g019-plan`. r0 is archived byte-exact as
`G019-scoped-invitations-membership-ownership-plan-r0.md` (SHA-256
`5560ca8e00182d741254b2065d3fe357ea8650aaff3a698c3a3991acde8f8c55`) with `G019-scope-allowlist-r0.txt`
(`4c888ba45b30086cc9596273cc1367907721e84c5d2b9bb4e926157bb7d5b7f2`, 39 patterns). Trigger unchanged: G018 draft
MR **!41** (Goal §"Owner lifts planning hold at G018 MR trigger"). Status: **not an execution binding; no
implementation, build, test, probe, container, Docker or MR.** Read-only source inspection only (two bounded
read-only Explore delegations: the G018 branch state at `46f1dc5e` and the RLS/lock/caller facts; Claude verified the
load-bearing findings). Tier: architecture/security/data (Goal). Budget **B7, checkpoint after attempt 4**, unspent;
planning spends no attempt (review §"next handoff"). Forecast: 3–4 attempts (§11.1). Implementation effort per the
owner's model direction: **Opus, high**. §13 maps every C1–C6 and Q1–Q8 item to its change; unchanged text is r0's.

**Every G018 dependency below is provisional.** G018 is neither accepted nor merged. Observed head of
`origin/g018-group-business-profile-onboarding` at r1: `46f1dc5e0fafbb248acb791fcf913dc46d2a50de` (16 commits after
`f32d049b`): slices 1–2 (migration `00014`, `tenancy/onboarding.go`, `access/onboarding.go`, `onboarding/api`, harness
rows, logrecord enums) and **slice 3 (web client, pages, routes, components, `ui/{input,textarea}`, web pins) exist**
(`763f8508`, attribution at `9c928f61`); **slice 4 (seed, `onboarding.spec.ts`, Playwright projects, shell/mcp pins)
exists only as a WIP commit** (`46f1dc5e`, no checkpoint or batch recorded yet); **slice 5 (docs) does not exist**
(`docs/` untouched). Codex reconciles each P-row (§0.2) at the G019 release; G018's final ACCEPT gates G019's merge
and final ACCEPT, not its planning. Nothing here plans G020 or G058; G058 is referenced only as the owner of
first-time-invitee sign-in, provider recipient proof, public admission and session issuance (Goal/Spec R4).

**Required reading, read at `origin/staging` `988ed020` (SHA-256):** `AGENTS.md`
`a99c0bcb6a81778315b2af95a672f6075880612b0f1ebddfad7a53f01932a61f`; `CLAUDE.md`
`dc50f98c27cdb57c3a0df0f58107494dfc72c092294cac7f3e82c1e125b66461`; `docs/ARCHITECTURE.md`
`d25bc7028d585aefd50663af1487aa8b77b835b18cdf03a2b7aff5dc3079f0f4` (§53.2, §53.7 incl. "G019 membership and
invitation boundary — 2026-10-17", §53.11); `docs/SECURITY.md`
`143de1d2f5278817a363dfeaf995ab807e2d1899f1623aa2fcd7becddb94cf00`; `docs/DECISIONS.md`
`6c713d58682caceb4bdfcca45a63bcef89f976355c92bbdf46459372199324d4`; ADR-004/005/006/007;
`management/v0/ROADMAP.md` `2569c5d544bee1fea3cc00d9e70dd0f8670741df945f96f39c50fbe6a836551a`;
`management/v0/PROGRESS.md` `24c5d07dd4d8b6f16d2fdd11bfbdc25cd1d88fb101c2da11c6b04a0a9c33511f`; Goal
`management/v0/goals/G019-scoped-invitations-membership-ownership.md`
`ed9d1208946314812b7bcc8a2cec345d9b453031a46da439e1175a1dfddd4e4b`; Spec
`management/v0/specs/G019-scoped-invitations-membership-ownership-spec.md`
`2e61cc41e7d96937f7ce155330ba7e655268ef25757cc66f02c806ccb9517a7a`; definition review
`management/v0/results/G019-definition-codex-review.md`
`21d85977d9fb24e53bddb8514cdbe0f310297726ca44a124fd2d7c19a5b07c8a`; r0 history
(`…-definition-r0.md` `4618001e0fc4f298f2fe5b3bcf3fa534e2703e51e595f1ba1c05a365a88a4b55`, `…-spec-r0.md`
`41e5ffb633a246a183c02509ad2e4ebd756fc732b8ab482d893e622132ff76d3`); G058 Goal
`17b0ac539f60bfcea5574cf602be753abd8e502903b438c3fcb1808b39ce6fab` (boundary only); G018 Plan r5
`management/v0/plans/G018-group-business-profile-onboarding-plan.md`
`4fe4642ff0d5457db156ff3d5d31e723891dad1b4dd303a86ef0f70fb6dd75fa` (patterns reused: allow-list form, slices,
focused selectors, harness rows, count pins, audit operations, §5.6 settled interfaces); G018 scope allow-list
`5a94f3ac6cc585bfd437e3b5cf7b577bec147ca0c2d3d816d9e8e8d267a84a7e`; G018 Plan r5 review
`5184febe95b79c38fd1c217858c1bc940c8c390ed5060cd5aebf1ea28fc7ad11`; the G018 Result on the branch (status and
successor #6 checkpoint, for what actually exists at `f32d049b`). G016/G017 Results consulted only for the `run.sh`
version-pin and `required-checks.tsv` arithmetic patterns (no new reading beyond G018's P-4/P-8 citations).

## 0. Gates, provisional G018 interfaces and R0

### 0.1 Gates

| Gate | Status | Owner |
|---|---|---|
| G018 final ACCEPT, hosted qualification and integration into `staging` (MR !41, observed head `f32d049b`) | **open** — gates G019 **merge and final ACCEPT**; planning proceeds provisionally (Goal §"Owner lifts planning hold") | Codex / owner |
| G018 slices 3–5 | at `46f1dc5e`: slice 3 **done**, slice 4 **WIP** (seed/browser/pins, unvalidated), slice 5 **absent** — G019's web slice reuses them (P-9..P-12, P-15); §0.3 applies if G019 execution is released before they are integrated | G018 implementer |
| R0 reconciliation (§0.3) after G018 integration; re-pin §11.2; rebind §2 | after integration | Claude; Codex reviews |
| Codex Plan ACCEPT; allow-list import (`management/v0/results/G019-scope-allowlist.txt`); initial Result; Git-first binding; B7 ledger; resource readiness; recorded execution release; later phase authority for implementation (Goal §"Execution Policy") | after the above | Codex |
| §12 (r1): Q1–Q8 carry the review's dispositions; only the R0 readiness inventory (Q1) and the two new review items D-13/D-14 remain open | before execution of the affected slice | Codex |
| G021 barrier closed; **no G020 or G058 planning/authoring** | — | — |

### 0.2 Provisional G018 interfaces relied on (P-1..P-16; file:line at `f32d049b` for P-1..P-8/P-13/P-14 (unchanged at `46f1dc5e`), at `46f1dc5e` for P-9..P-12/P-15/P-16)

| ID | Interface | Where |
|---|---|---|
| P-1 | session-scope behaviour (`Authorization: Session` header, exact `https://` origin ≤ 300 chars, resolver lookup, session Organization = resolved Organization, CSRF for mutations, `BeginRequest` then Organization-only `Resolve`), byte-pinned copy of `branding/api` (`TestOnboardingCopyIsExact`) | `core/internal/onboarding/api/http.go:216-277`, `:492` (`mutation`), `:120-137` (codes) |
| P-2 | fifteen capabilities (`group.create`, `group.complete_onboarding` last); `organization.manage_members` and `group.manage_members` defined **without any operation**; evaluator treats owner and admin alike (`orgAdmin`); gate order capability → Organization gate (exactly one `source='membership'` binding) → target → context → matrix | `core/internal/tenancy/access/policy.go:12-34,243-267`; `matrix_test.go` (`len(oracleCapabilities) != 15`, 330 primary cells, `:882`); `integration_test.go:1473` |
| P-3 | audit catalog: 30 trigger + 6 core actions, 58 reasons, **36** operations (`TestCatalogParity` reads `00014` for `events_operation_valid`); membership actions T7/T8/T9 are trigger-owned with `role`/`previous_role` and reason NULL; `events_attribute` scope check (`rls_org_role()` non-NULL) applies to **depth-1 direct appends only** (`00013:176-196`) | `core/internal/audit/audit.go`; `audit_test.go:69-97`; `core/migrations/00013_organization_branding.sql:138-300` |
| P-4 | migration set `00001`–`00014`; `current_version\t14` pins | `core/migrations/migrations_test.go:50`; `core/internal/migrate/{runner_test.go,integration_test.go:760,816,827,838,930,onboarding_integration_test.go:93}`; `core/internal/mcp/env_integration_test.go:253`; `infra/validation/run.sh` (suite_root, `mig_step`, B-MIGRATE-CLI, recovery_drill, persist, variants, B-IDENTITY-CATALOG) |
| P-5 | `00014` goose style: precondition `DO`, StatementBegin/End per function, **postconditions pinning the full ACL string of the four tables, six policies, two functions and the `groups` trigger set** (`:235-255`); SECURITY INVOKER trigger `app.groups_bootstrap_owner()` admitted by `role_bindings_insert_bootstrap` (`pg_trigger_depth() > 0`, self, `group_owner`, the inserted Group); boolean DEFINER reader `app.rls_profile_eligible` owned by `<db>_policy_reader` | `core/migrations/00014_group_onboarding.sql:76-255` |
| P-6 | `role_bindings` runtime surface after 00014: column SELECT (7 non-key columns), INSERT (7 columns, admitted only by the bootstrap policy), UPDATE(role) with policy `role_bindings_lock` `WITH CHECK (false)`, **no DELETE**; `role_bindings_scope_consistent` admits only `source='membership'`; `role_bindings_keys_immutable` freezes every column but `role`; all FKs `ON DELETE RESTRICT` (no cascade can remove a binding); unique `role_bindings_subject_scope_key` | `00003:64-139`; `00006:399,422-431`; `00014:138-203` |
| P-7 | importer boundary: only `access` imports `tenancy`; importers of `access` are exactly `internal/mcp`, `domains/api`, `branding/api`, `onboarding/api`; `onboarding/api` imported only by `cmd/server`; `inContextTx` call sites **12**; exported-method pin; SECURITY DEFINER allowance per named migration (00014 = 1) | `core/internal/tenancy/access/boundary_test.go:44-66,197,239-242,478-519`; `core/internal/tenancy/boundary_test.go` (15 production files, +6 package functions, `TestSQLBoundary` 54 statements); `infra/validation/run.sh:1165` (B-MCP-SOURCE importer strings, `MAIN_GO_G016_SHA`, `CORE_ONBOARDING_ROUTES`) |
| P-8 | harness: `rls_go` runner (`run.sh:2652`), `onboarding_checks()` template (`:3273`: catalog + IT rows with `it_count` pins 5/14, `rls_logcheck`), `required-checks.tsv` **123** rows (header arithmetic `… + 2 G017 + 2 G018 = 123`), `WEB_ROUTES` (`:618`, basenames under `web/app/`), B-TENANCY-HTTP probe arrays (`:2165-2207`; `TENANCY_PROBE_CORE` 18, `TENANCY_PROBE_WEB` 15 incl. `/memberships`, `/api/memberships`, `/api/role-bindings`), `UI_FILES` (`lib.sh:568`, seven modules), `BROWSER_MATRIX_IDENTITY_E2E` (`lib.sh:628-641`), `profile-paths.tsv` | as cited |
| P-9 | G018 web BFF **as built** (slice 3): `force-dynamic` pages, `readCookie`/`ownOrigin`, hidden `csrf_token`, `readForm(request)` (`web/lib/session.ts:177`; 16 KiB), `seeOther`/`statusPage`/`methodNotAllowed`, closed `?status=`/`?error=<field>` maps that never echo values, strict-shape clients with `successBytes` (`web/lib/onboarding-client.ts`, 537 lines); `tenantRequest()`/`session()` page pattern; `ShellNavItem.href` stays `` `#${string}` `` (`app-shell.tsx:16`); **`WorkspaceArea = "groups" \| "business-profiles"`** with `WORKSPACE_AREAS` (`:20-21`), `AppShellProps.area?` (singular) and a `nav[aria-label="Workspace areas"]` rendered on every tenant page for a member | `web/lib/{session,onboarding-client,request-context}.ts`, `web/components/layout/app-shell.tsx:16-21` |
| P-10 | G018 `components/ui/{input,textarea}.tsx` **exist** (kit-adapted; header `// Adapted from the design kit components/ui/<f> (sha256 <kit sha>); // changes per docs/foundation/WEB-SHELL.md (G018): …; // Derived from shadcn/ui (MIT); licence text in web/THIRD_PARTY_NOTICES.md`); `UI_FILES` **9** entries and `UI_SLOTS` ending `input textarea` (`lib.sh:568,570`); `components/onboarding/{form-field,onboarding-frame,profile-fields,state-badge}.tsx`; `web/app/groups/{page,new/page,[id]/page,[id]/onboarding/page}.tsx`, `groups/{create,[id]/onboarding/complete}/route.ts`, `business-profiles/*`; `web/THIRD_PARTY_NOTICES.md` §shadcn/ui names input/textarea (G018 needed scope amendment 01 for that file) | as cited at `46f1dc5e` |
| P-11 | G018 seed/browser **WIP** (`46f1dc5e`: `mcp-seed.sql` 26 lines, `onboarding.spec.ts`, projects `onboarding-light/dark` `playwright.config.ts:109-110`, `mcp.spec`/`shell.spec` nav pins, `BROWSER_MATRIX_IDENTITY_E2E` third entry, `profile-paths.tsv:102-103` `web/app/groups/`, `web/app/business-profiles/`). Staging seed ownership: carol `org_owner` A, frank `org_admin` A, bob `group_owner` research; **Organizations B/C and Groups design/operations/planning have no owner**; dave is Group-role-only in C | `infra/identity/mcp-seed.sql:49-65` (staging); G018 Result successor #9 record |
| P-12 | G018 docs slice **absent** at `46f1dc5e` — G019 edits the same files after it; if G018's docs are not integrated when G019's slice 5 runs, G019 documents only its own rows (§0.3) | G018 Plan r5 §8 slice 5 |
| P-15 | G018 slice-3 corrections G019 inherits: every new `lib/*.ts` or route importing `request-context` is added to the exact (no-glob) `static_web_boundary` allow-list (`run.sh:860-861`; missing it failed MR !41 pipeline 2958938657, B-WEB-BOUNDARY); pages/components never import `request-context`; components never import `@/lib/config` (pages pass runtime config); B-WEB-TOKENS rejects kit classes even in comments (`bg-surface` → `bg-background`, `border-input`, shell focus outline); `react/display-name` on closures; `no-html-link-for-pages` misfire needs one justified disable; `not_available`/`forbidden` reads render "Not available", not 404; conflict/unknown pages never claim success | G018 Result successors #7/#8 |
| P-16 | `pageInput[T ~[16]byte](after, limit, parse)` (`access/onboarding.go:254`; limit 1..`tenancy.MaxOnboardingPage` = 100, UUID cursor) and the `limit+1` readers `tenancy.GroupsPage`/`ProfilesPage` (`tenancy/onboarding.go:103,145`, trimmed at `access/onboarding.go:375-381`) — G019's page readers copy this shape | as cited |
| P-13 | Group row lock order: ascending Group id with per-row strength in `ReadFacts` (`sqlFactGroupsLock` below target FOR SHARE, target FOR NO KEY UPDATE, above FOR SHARE); Organization FOR SHARE or FOR NO KEY UPDATE (`ExclusiveOrganization`); subject's own bindings FOR SHARE through `role_bindings_lock`; **`ReadFacts` reads only the caller's own bindings** (no other subject's binding, no owner count) | `core/internal/tenancy/facts.go:78-160` |
| P-14 | G018 §5.6 settled interfaces 1–8 (creation, cardinality "one or more", runtime write surface, `group.manage_members` not owner authority, derived completion, 15 capabilities, lock order, emergency suspension unchanged) | G018 Plan r5 §5.6 |

Existing non-G018 interfaces reused unchanged: identity `DisableUser`/`EnableUser` (trusted, no route; IDENTITY
shape; revocation-priority audit exception, ADR-006), `LockLiveSession`/`CheckLiveSession`, `newSecret`/`hashSecret`/
`equalSecret` (`identity/credential.go`, unexported — copied, not exported, §1.4), `app.is_organization_member(uuid,
uuid)` (`<db>_auth_reader`-owned STABLE DEFINER, EXECUTE to runtime only, 00011:767-776), the `<db>_auth_reader`
helper policies `users_auth_reader`/`identity_principals_auth_reader` (00006:481-482) and `role_bindings_auth_reader`
(00011:825) as the precedent for an auth-trust DEFINER boundary, `tenancy.Service` model primitives
`CreateOrganization`/`CreateGroup`/`Grant*`/`Change*`/`Remove*` (`service.go:74,185,438-466`; **no production caller;
`tenancy.NewService` itself has none; runtime has no INSERT on `organizations`/`groups`/`role_bindings` since
00006:397-400, so they already work only on privileged test pools — §1.4 "Trusted bootstrap"), the G013 recovery
ledger kinds `membership.remove`/`membership.downgrade` and `replay-restrict.sql` (superuser; `recovery.sh:894`),
P-16 paging, the MCP per-connection limiter pattern (`golang.org/x/time/rate`, `mcp/handler.go:206-219`), and the
documented trust basis of the context settings: `sqlInitContext` uses `set_config(…, true)` with shape-only parsers and
**no database check that a runtime connection cannot set `app.user_id` to any UUID** (DATABASE-ISOLATION.md:187-191;
`tenancy/store.go:136-143`) — hence no G019 control may treat a transaction setting as proof (C1).

### 0.3 R0 after G018 integration

Re-read every P-row against integrated `staging`; re-pin §11.2 (migrations, capabilities, operations, audit
counts, `required-checks.tsv`, route enums, `WEB_ROUTES`, `UI_FILES`, importer strings, `inContextTx` sites, seed
bindings, Playwright projects, `BROWSER_MATRIX_IDENTITY_E2E`); rebind §2; record drift in the Result. **Material
drift returns as a Plan revision:** a change to the `role_bindings` policies or trigger of 00014, to the
`events_attribute` body, to the session-scope helper, the shell navigation model, the seed ownership shape, or the
harness row pattern. R0 also inventories (C6): the exact `static_web_boundary` allow-list entries, `WEB_ROUTES`
(31 at `46f1dc5e`), `UI_FILES`/`UI_SLOTS` (9), the final G018 seed bindings and Playwright projects, the
`THIRD_PARTY_NOTICES.md` §shadcn/ui sentence, `TENANCY_PROBE_WEB` length, the `WorkspaceArea` union and nav pins, and
the actual persisted databases the operator names for the ownerless-record readiness check (Q1: a scoped inventory,
no unrelated-host scan, no automatic repair). If G018's web slices are not integrated when G019 execution is
released, slices 1–2 (core) may proceed under the release; slices 3–4 (web) wait for the integrated G018
`components/ui`, client and shell decisions (they are not re-implemented in G019; Q6: one Goal, bounded slices).

## 1. Approach

### 1.1 Trust boundaries first (what changes, what a malicious actor can now control)

| Boundary | Change | Containment |
|---|---|---|
| **Runtime write surface on `role_bindings`** | 00015 admits, for the runtime and FORCE-bound owner roles: UPDATE(role) through a real policy (`role_bindings_update`, replacing the lock-only `WITH CHECK (false)`) and a new column-less **DELETE** grant with policy `role_bindings_delete`. The consumption-written INSERT/upgrade shapes are admitted **only to `<db>_auth_reader`** inside the claim function (`role_bindings_insert_invited`, `role_bindings_auth_upgrade`, §4.4) — the runtime role gains no INSERT shape (C1) | every admitted runtime statement is bound to the active Organization (`ctx_org()`), a current membership of the actor, and the scope the actor administers (Organization: `MGR`; Group: `MGR` or `group_owner`/`group_admin` of that Group; self rows for leaving). **Owner rows and owner roles are not administrable by policy alone**: the SECURITY INVOKER trigger `app.role_bindings_authority_guard()` (BEFORE UPDATE OF role / BEFORE DELETE) refuses any statement that reads or writes an owner role unless the actor currently holds `org_owner` (Organization scope) or `group_owner` **of that Group** (Group scope) per the DEFINER readers — an Organization owner/admin has **no** override of Group ownership (Spec R1; Architecture §53.7; Q2 settled). The same trigger enforces **full promotion eligibility** (same-scope admin → owner only, enabled recipient under a lock consistent with `DisableUser`, current Organization membership for Group promotion; C2), last-owner protection (`app.rls_owner_count`, §4.3) after a separate scope-row lock command, the dependent-grant rule through a DEFINER count that sees sibling Groups regardless of the selected Group, and the Organization-only context requirement for Organization departure (C3, §4.5). A direct runtime statement with a fixed context meets the same guards (direct-vs-facade lock inversion is detected, not prevented — §1.5); the superuser repair path (RECOVERY) bypasses RLS, and the guard body returns early for roles other than runtime/owner/auth-reader, exactly as `groups_bootstrap_owner` does (P-5) |
| **Owner cardinality** | first database enforcement of "one or more owners" on every supported mutation: role change, removal, transfer, departure, ordinary retirement | count of authoritative bindings by a DEFINER count reader (`app.rls_owner_count(scope_type, scope_id)` owned by `<db>_policy_reader`, reading through a helper policy limited to `ctx_org()` rows — not the actor's `role_bindings_select` projection, Spec R2; it answers only for a scope the caller can currently see, C1), evaluated **inside the VOLATILE plpgsql trigger as a separate command after `FOR NO KEY UPDATE` on the scope row**, so two concurrent demotions/removals of the two last owners serialize and the second's count command takes a fresh READ COMMITTED snapshot after the lock wait (§4.5 pins language/volatility). The Go guard holds the same row at the same strength first (`ExclusiveOrganization` for Organization-scope owner writes/departures; target Group `FOR NO KEY UPDATE`) so the ordinary path never upgrades a lock inside the trigger. No unique-owner index, owner column, mirrored owner store or admin binding (P-14 item 2). Creation stays atomic (G018 trigger for Groups; the trusted `CreateOrganizationWithOwner` for tenants, §1.4) |
| **New table `app.invitations`** | the single invitation lifecycle: immutable intent (`organization_id`, `scope_type/scope_id/group_id`, `role` ∈ admin/member only, `issuer_user_id`, intended recipient, `expires_at`), one-use consumption, revocation, 32-byte SHA-256 token digest | runtime: column-limited INSERT (issue), UPDATE of `revoked_at` only (revoke), SELECT **without `token_digest`**, no DELETE; policies bind issue and revoke to the actor's administrative scope and reading to that scope or the intended/consumed recipient. **No transaction-scoped invitation context exists** (r0's `app.ctx_invitation()` is withdrawn, C1). `invitations_keys_immutable` admits exactly `revoked_at` NULL→value for runtime/owner and the consumption transition for `<db>_auth_reader` only; `audit.capture_invitations` records `invitation.issued/revoked/consumed` in the same transaction (strict, ADR-006). The raw token exists only in the issue response and the acceptance request body; never in logs, audit payloads, URLs served by core, redirects, or cacheable responses (`no-store`) |
| **Invitation consumption writes membership** | one `<db>_auth_reader`-owned SECURITY DEFINER function `app.invitation_claim(p_id, p_digest, p_address)` (§4.4; the 00011 `is_organization_member` ownership precedent, now with a bounded write body) performs, for `ctx_user()` only, the whole transition: lock the invitation row, compare the digest, check expiry/revocation/consumption/intent/address, lock and check the enabled recipient, lock and check the issuer's **current** authority and (Group) the recipient's current Organization binding, then UPDATE the invitation (consumed triple, `intended_user_id` claim for an address-addressed Organization invitation) and INSERT the binding or upgrade a lower own role — one function call, atomic by construction | **why ordinary fixed-context runtime SQL cannot claim another invitation (C1):** the runtime role holds no UPDATE privilege on `consumed_at`/`consumed_by_user_id`/`intended_user_id`, cannot SELECT `token_digest`, and holds no INSERT on `role_bindings`; the only path to those writes is the claim function, which requires the raw token's digest (unreadable by runtime, known only to the delivery recipient) for the row it names, binds the subject to `app.ctx_user()`, and copies the intent from the immutable row — so a context subject U can consume only an invitation whose token U presents, only as U, only for the row's recorded scope/role (never an owner role, CHECK), only while unexpired/unrevoked/unconsumed, only while the issuer currently holds issuing authority, and (Group) only while U is a current Organization member. A used token cannot restore a removed membership (one-use, immutable). What the database **cannot** verify is that an address-addressed Organization invitation's recipient owns that mailbox: that is the trusted consumer's job (verified provider evidence, §1.4 "Address and recipient evidence"), and the Go consumer is the only production caller of the claim for Organization invitations (pinned by `TestSQLBoundary`); a request-path context cannot reach it because no route serves Organization consumption (§5.3) |
| **Who may consume** | Group invitations: the intended **current Organization member** in their own ordinary tenant session (browser acceptance, §1.6) → `AcceptGroupInvitation` → the claim. Organization invitations: **no route and no session in G019**; only the bounded Go consumer contract `access.ConsumeOrganizationInvitation(ctx, VerifiedRecipient, id, token)` (§5.5) in IDENTITY shape (`InitContext(user, nil, nil)`; no tenant context, no nonmember session) whose `VerifiedRecipient` is built only from `ProviderAddressEvidence` (unexported fields; `access.AttestAddress` is the sole constructor, pinned to `core/internal/identity/**` and `_test.go` call sites by `TestAccessBoundary` — a boolean in an adapter request is not evidence, C4) | no client can construct the evidence or choose the consuming subject; the consumer passes the canonical `UserID` mapped from the provider `(issuer, subject)` and the provider's **current** verified address for that exact subject; the claim rechecks everything else at the authoritative statement. Spec R4: this is not a nonmember session, applicant session or callback bypass |
| **First HTTP exposure of membership administration** | new adapter `core/internal/membership/api` (**15 method/path pairs on 14 paths** under `/internal/v1/members*`, `/internal/v1/ownership/transfer`, `/internal/v1/invitations*`, `/internal/v1/account/retirement-eligibility`; §5.3, C6) | P-1 scope copied byte-pinned (`TestMembershipCopyIsExact`); every tenant route calls a guarded facade operation and the one IDENTITY-shape route reuses the existing session trust; closed codes; no foreign names/UUIDs in errors; **every list is keyset-paged** (≤ 100 rows per page, `next` cursor; C5); whole-payload budgets 8 KiB; issuance bound (≤ 100 pending unexpired invitations per scope, serialized on the scope row, `invitation_limit`); the intended address is validated and normalized but never echoed to anyone but the issuing scope's administrators; B-TENANCY-HTTP keeps every existing absence probe (`/memberships`, `/role-bindings`, `/internal/v1/memberships`, `/internal/v1/role-bindings` stay unserved) and adds positive closed-route pins for the new roots |
| **Ordinary retirement vs. security disable** | `access.RetirementEligibility` (self, IDENTITY shape, all Organizations) and trusted `access.RetireUser` (no route) | eligibility is a DEFINER reader returning only the subject's own sole-owned scope identifiers; retirement = per-Organization guarded departure (ascending Organization id, each atomic) then `identity.DisableUser`; a tenant administrator has no path to it (no route, no capability; `DisableUser` unchanged). `DisableUser` on a last owner stays immediate: the binding is retained (no binding write), sessions revoked, authority denied at guard steps 1–2 (P-14 item 8) |
| **Browser** | member/invitation/owner pages and forms | server-rendered, escaped; `readForm` exact-Origin + CSRF (16 KiB default); no client authority; every action re-decides on fresh facts; stale list choices (recipient removed, role changed, last owner) answer closed status pages with the current state; the raw token is shown once in the **direct 200 `no-store` response of the invite POST** (no redirect can carry it; the BFF never writes it to a `Location`, URL, title or log) and is posted back only in the acceptance form body (§1.6, C4) |

### 1.2 Riskiest slice first

Slice 1 (§8) proves, against real PostgreSQL as the runtime role and as the FORCE-bound owner, the owner-cardinality
and authority guards (promotion, transfer, demotion, removal, departure, the last-owner races), the invitation
lifecycle with its consumption trigger and the Organization-consumer contract with synthetic recipient evidence,
plus the legacy preflight and the injected mid-transaction failures, before any HTTP or UI exists. Everything after
it is thin composition (lists, adapter, pages).

### 1.3 Operation and capability matrix (R1 pin; the finite set)

Capabilities 15 → **19**: `organization.manage_owners` (Organization kind), `group.manage_owners` (Group kind),
`organization.leave` (Organization kind), `group.leave` (Group kind). `organization.manage_members` and
`group.manage_members` (existing, P-2) gain their first operations.

| Capability | org_owner | org_admin | org_member | group_owner | group_admin | group_member |
|---|---|---|---|---|---|---|
| `organization.manage_members` (existing) | Y | Y | — | — | — | — |
| `organization.manage_owners` **new** | Y | — | — | — | — | — |
| `organization.leave` **new** | Y (last-owner guard) | Y | Y | — | — | — |
| `group.manage_members` (existing) | all | all | — | own | own | — |
| `group.manage_owners` **new** | **—** | **—** | — | own | — | — |
| `group.leave` **new** | — | — | — | own (last-owner guard) | own | own |

`group.manage_owners` is deliberately **not** held by Organization owners/admins (Spec R1 "no implicit
Organization-owner override of Group ownership"; settled — Q2 disposition, no owner re-question; the matrix oracle
gains an explicit Group-owner-only row). Organization-role subjects without a Group binding cannot `group.leave`
(nothing to leave). Context gate as today: a Group target must be the selected Group when one is selected;
**Organization departure and Organization-scope removal additionally require an Organization-only context**
(`ctx_group() IS NULL`; C3).

| Guarded operation (all new; `tc *TenantContext`) | Capability | Target | Tx | Writes |
|---|---|---|---|---|
| `ListMembers(tc, scope, after, limit, roleFilter)` | `organization.manage_members` / `group.manage_members` (`roleFilter` ∈ any/`owner_candidates` = enabled same-scope admins, so promotion/transfer reaches eligible admins on every page — C5) | Organization or exact Group | RR RO | — |
| `ListInvitations(tc, scope, after, limit)` | same | same | RR RO | — |
| `ListMyInvitations(tc, after, limit)` | none (own pending Group invitations; keyset) | own rows | RR RO | — |
| `ListInvitationCandidates(tc, group, after, limit)` | `group.manage_members` (Group kind; the least-disclosing current-scope source for the Group recipient picker: current Organization members not yet in the Group, ids + display names only, via `app.rls_group_invitation_candidates`, §4.3; D-13) | exact Group | RR RO | — |
| `IssueInvitation(tc, scope, role, recipient)` | same (role ∈ admin/member of the scope; Group recipient must be a current Organization member) | same | RC RW | one `invitations` INSERT (keyed id; no RETURNING of the digest; the raw token is returned by Go) |
| `RevokeInvitation(tc, id)` | same (the invitation's scope) | the invitation | RC RW | one UPDATE (`revoked_at`) |
| `GetInvitation(tc, id)` | same, **or** the intended recipient (own pending invitation; no capability) | the invitation | RR RO | — |
| `AcceptGroupInvitation(tc, id, token)` | none (recipient = `tc` subject; current Organization member) | the invitation | RC RW | one `SELECT app.invitation_claim($1, $2, NULL)` (the function updates the invitation and writes/upgrades the binding) |
| `ChangeMemberRole(tc, scope, subject, role)` | `*.manage_members` (admin ↔ member only; subject ≠ actor; subject must not be an owner) | the subject's binding | RC RW | one UPDATE(role) |
| `RemoveMember(tc, scope, subject)` | `*.manage_members` (subject not an owner; Organization scope requires an Organization-only context and removes dependent Group bindings atomically after locking them, refusing if any Group would lose its last owner) | the binding(s) | RC RW | n DELETEs (Group bindings ascending `group_id`, then the Organization binding) |
| `PromoteToOwner(tc, scope, subject)` | `*.manage_owners` (subject is an enabled current same-scope admin; Group: also a current Organization member — all re-enforced by the guard at the statement, C2) | the subject's binding | RC RW | one UPDATE(role → owner) |
| `DemoteOwner(tc, scope, subject, role)` | `*.manage_owners` (subject is an owner; `role` ∈ admin/member; last-owner guard; self allowed only when another owner remains) | the subject's binding | RC RW | one UPDATE(role) |
| `TransferOwnership(tc, scope, recipient, formerRole)` | `*.manage_owners` (actor is an owner; recipient an enabled current same-scope admin; `formerRole` ∈ admin/member/none) | recipient's and actor's bindings | RC RW | UPDATE recipient → owner, **then** UPDATE/DELETE actor (the guard sees the recipient already an owner) |
| `LeaveScope(tc, scope)` | `organization.leave` / `group.leave` (self; last-owner guard; Organization departure requires an Organization-only context and removes own dependent Group bindings first, refusing before partial writes) | own binding(s) | RC RW | n DELETEs |
| `ConsumeOrganizationInvitation(ctx, VerifiedRecipient, id, token)` (no `tc`; IDENTITY shape; §5.5) | none — the bounded consumer (R1 exception) | the invitation | RC RW | one `SELECT app.invitation_claim($1, $2, $address)` (claims `intended_user_id`, consumes, writes the binding) |
| `RetirementEligibility(ctx, credential, origin)` (self, IDENTITY shape; served by `GET /internal/v1/account/retirement-eligibility`, C6) | none (existing session trust) | own scopes | RR RO | — |
| `RetireUser(ctx, user)` (trusted; no route) | none (operator path) | all own scopes | per-Organization RC RW, then `DisableUser` | departures, then disable |
| `tenancy.CreateOrganizationWithOwner(ctx, name, owner)` (trusted model primitive on a privileged pool; the **only** supported creation shape; no route or production caller in G019) | none | — | RC RW | `organizations` INSERT + `role_bindings` INSERT (org_owner) in one transaction |

Nothing else changes in the matrix. Invitations never carry an owner role (CHECK). No direct grant without an
invitation (Spec R4: one lifecycle); no BusinessProfile role; no MCP tool; `business_profile.privileged_settings.manage`
and `audit.read` unchanged.

### 1.4 Contracts (R1, R2, R3, R4)

**Guard shape.** Every operation uses the existing `guard`/`guarded` sequence (P-7): `bind` → `pre` → `inContextTx`
(`InitContext` first statement; `LockLiveSession` for writes; `ReadFacts` with `Lock`; `contextCheck`; `decide`; `act`).
Two additions to the facts (`tenancy/facts.go`, `FactQuery`): `Subject *UserID` (the target subject's bindings in the
active Organization, read **and locked `FOR NO KEY UPDATE`** after the scope locks — a new fact statement over
`role_bindings` admitted by the new `role_bindings_update`/`role_bindings_delete` USING branches; invisible rows mean
`not_found`) and `OwnerCount bool` (one `SELECT app.rls_owner_count($scope_type, $scope_id)` after the scope lock, for
the truthful UI feedback and the facade's own precheck; the trigger recheck is the guarantee). "Fresh trusted context"
is the existing per-request `Resolve` snapshot plus `LockLiveSession` in the write transaction; **no new MFA or
authentication-age check is added** (none exists: `mfa_assurance` is always `unknown`, `MaxAuthAge` applies at sign-in
only — recorded, not designed around).

**Promotion / demotion / transfer (Organization and Group).** Lock in the §1.5 order: `LockLiveSession`; Organization
`FOR NO KEY UPDATE` (`ExclusiveOrganization`) for Organization scope or `FOR SHARE` for Group scope; the actor's own
bindings `FOR SHARE` (`ReadFacts`, P-13); the target Group `FOR NO KEY UPDATE` for Group scope; then the new `Subject`
fact: the subject's (and, for transfer, the actor's) scope binding(s) `FOR NO KEY UPDATE` in one statement ordered by
`id` (visible to the actor through `role_bindings_select`'s manager branches and lockable through
`role_bindings_update` USING — plain `org_member`/`group_member` actors see no such row and get `not_found` after the
capability denial). Decide `*.manage_owners`. Recheck on the locked facts for truthful feedback: exact scope, current
role as required (`admin` for a promotion/transfer recipient; `owner` for demotion), subject enabled, Group recipients
also hold a current Organization binding. Act: `tenancy.ChangeRoleTx(ctx, tx, binding, role)` — the existing
`sqlChangeRole` statement shape keyed by `id`, now admitted by policy and **guarded by the trigger, which re-enforces
every eligibility condition at the statement (C2, §4.5): same-scope admin → owner only, `app.rls_user_enabled_lock()`
(the subject's `users` row `FOR SHARE`, conflicting with `DisableUser`'s `FOR UPDATE` at `identity/store.go:968`, then a
fresh-snapshot status read), and `app.is_organization_member()` for Group promotion**. Transfer executes two
statements in the order recipient-up then actor-down (or actor DELETE for `formerRole = none`); the guard's last-owner
count on the second statement already sees the recipient. T8 `membership.role_changed` (with `previous_role`) and T9
fire with the request actor. Failure semantics: trigger refusals map through `checkCodes`/`writeFailure`:
`last_owner` → `conflict` (field `last_owner`), `owner_authority` → `forbidden`, `dependent_bindings` → `conflict`
(field `dependent_bindings`), `recipient_ineligible` → `conflict` (field `recipient_ineligible`), `context_required`
→ `invalid_input` (field `context`); `membership_not_found` → `not_found`; a deadlock abort (40P01, only possible
against a direct non-facade statement, §1.5) → `conflict` (field `retry`), zero writes. **Audit fault** (strict
trigger capture) rolls back the whole transaction: ownership unchanged (AC2).

**Ordinary role change / removal.** `*.manage_members`; subject ≠ actor (`self_target` → `invalid_input`); subject's
current role must be admin/member (an owner subject → `forbidden`, `owner_subject`, even when the requested final
role is admin/member or absent — Spec R1); requested role ∈ admin/member. Organization removal (**Organization-only
context required**: `tc` with no selected Group, else `invalid_input` `context` — a selected-Group context hides
sibling Group bindings from `role_bindings_select`, C3): after the Organization `FOR NO KEY UPDATE` lock, lock the
subject's Group bindings in that Organization ascending `group_id` (all of them are visible to `MGR` with `G IS NULL`),
delete them (each DELETE passes the guard: a Group last-owner refusal aborts the transaction before any Organization
write), then delete the Organization binding — the guard's dependent-grant step recounts through the DEFINER reader
`app.rls_dependent_group_bindings(subject)` (not the invoker's filtered view) and refuses `context_required` when
`ctx_group()` is set, so even a direct selected-context DELETE cannot leave a dormant Group grant. A Group owner subject
makes Organization removal **refuse** (`forbidden`, `owner_subject`): explicit Group succession/removal by that Group's
owner must occur first (Spec R3 "cannot silently strip a Group owner").

**Leave.** Self only; Organization departure = the same Organization-only-context dependent-removal sequence on own
rows; a last owner of the Organization or of any own Group → `conflict` (`last_owner`, with the blocking scope kind —
never another tenant's details). Readmission later is a fresh binding (new UUID) with no Group rows (the
dependent-grant guard guarantees none survive on any path, AC3).

**Invitations.** `IssueInvitation`: `pre` validates `role` (admin/member of the scope), the recipient form (Group:
`intended_user_id`; Organization: `intended_address`, normalized §"Address normalization"), `expires_in` (default 7
days, maximum 30 days, minimum 1 hour); facts: **the scope row at exclusive strength** — Organization scope
`ExclusiveOrganization` (`FOR NO KEY UPDATE`), Group scope the target Group `FOR NO KEY UPDATE` — which is the existing
scope serialization (`policyWriteQuery`/target-Group pattern), so count-then-insert is serialized per scope (C5);
then the pending **unexpired** count (`sqlInvitationsPendingCount`: `consumed_at IS NULL AND revoked_at IS NULL AND
expires_at > now()`, served by `invitations_pending_idx` on `(organization_id, scope_type, scope_id, expires_at)`
WHERE pending — expired history never counts against the bound and is never scanned as if it did); ≥ 100 →
`conflict`, field `invitation_limit`; the BEFORE INSERT trigger `app.invitations_bound()` repeats lock-then-count for
direct statements (§4.5); for Group scope the recipient's current Organization binding must exist
(`app.is_organization_member(recipient, org)`; EXECUTE pinned for the runtime role at precondition — §4.1); token: 32
bytes `crypto/rand`, base64url (43 chars), digest `sha256("invitation/1\x00" || token)` (tagged like `domain/1`),
stored only as the digest; one keyed INSERT (id from `newKey`), no RETURNING; returns `Invitation` + the raw token
once. `RevokeInvitation`: one UPDATE `SET revoked_at = now() WHERE id = $1 AND revoked_at IS NULL AND consumed_at IS
NULL` → zero rows = `conflict` (`invitation_not_pending`) or `not_found`. `ListInvitations`/`ListMyInvitations`:
keyset pages (P-16 form, cursor = invitation id, ≤ 100 + `next`), status derived (`pending`, `expired`, `revoked`,
`consumed`) from the timestamps — no status column; rows carry `intended_address` only to the issuing scope's
administrators (the policy's admin branch), never the digest. `GetInvitation` for the recipient answers scope kind,
Organization/Group name, role, expiry, status — nothing about the issuer beyond display name.

**Consumption (one contract, two entry points, one database transition — C1/C4).** Both entry points end in
`SELECT * FROM app.invitation_claim($id, $digest, $address)` (§4.4), executed as the context subject `U =
app.ctx_user()`. The function (VOLATILE plpgsql, SECURITY DEFINER, owned by `<db>_auth_reader`) runs these commands,
each with a fresh READ COMMITTED snapshot: (1) `SELECT … FROM app.invitations WHERE id = p_id FOR NO KEY UPDATE`
(serializes duplicate claims); (2) refuse `not_found` unless `token_digest = p_digest AND consumed_at IS NULL AND
revoked_at IS NULL AND expires_at > now()` and the recipient intent matches — Group: `intended_user_id = U`;
Organization: `intended_user_id = U` or (`intended_user_id IS NULL AND intended_address = p_address AND p_address IS
NOT NULL`); (3) `SELECT status FROM app.users WHERE id = U FOR SHARE` → `recipient_ineligible` unless `active`;
(4) lock, in one statement ordered by `id` `FOR SHARE`, the issuer's current binding(s) for the scope and the
recipient's existing Organization/scope bindings, then with the post-lock snapshot: refuse `issuer_authority_lost`
unless the issuer **currently** holds `MGR` of the Organization or (`Group`) `group_owner`/`group_admin` of that Group
(no recipient-side revocation UPDATE: refusal only, the issuer's administrators revoke or reissue); Group: refuse
`recipient_ineligible` unless the recipient holds a current Organization binding; (5) existing scope binding with an
equal/higher role → consume and answer `already_member` (no binding write); lower → `UPDATE … SET role` (`upgraded`);
none → `INSERT` the binding copied from the row's immutable intent (`joined`); (6) `UPDATE app.invitations SET
consumed_at = now(), consumed_by_user_id = U, intended_user_id = COALESCE(intended_user_id, U), consumed_binding_id =
<id>`. **Replay (C4):** when step (2) finds `consumed_by_user_id = U AND token_digest = p_digest` (same token, same
recipient intent), the function answers `already_member` **only if a current qualifying binding actually exists**
(equal/higher role, `source='membership'`); a consumed row whose binding was later removed answers `consumed` (closed:
"already used; membership not current" — it never fabricates or restores membership). Any other mismatch (unknown id,
wrong token, token of another invitation, another consumer, revoked, expired, intent mismatch, address mismatch) is one
closed `not_found`. The facade reads nothing before the call; the function's outcome row carries the intent
(`organization_id`, scope, role, `binding_id`) the adapter needs. **Group entry point** `AcceptGroupInvitation(tc, id,
token)`: the ordinary guarded write (`LockLiveSession`, `ReadFacts` own bindings) then the call with `p_address =
NULL`. **Organization entry point** `ConsumeOrganizationInvitation(ctx, VerifiedRecipient, id, token)` (§5.5): no
`TenantContext` (the recipient is not a member) and no session; IDENTITY shape `InitContext(user, nil, nil)` on the
runtime pool (`ctx_org()` NULL, so no tenant policy admits anything to this context — the claim function is the only
reachable write) then the call with the evidence address. Denials on this path are recorded as `access.denied` with
**NULL Organization** (the depth-1 audit scope check refuses a non-member's Organization, P-3; Q7 accepted) plus the
`access.decision` log event; the function's own `invitation.consumed`/`membership.granted` captures are trigger-owned
(depth ≥ 2) and attributed to the row's Organization. Session issuance is not part of either entry point (G058).

**Address and recipient evidence (C4; Spec R4).** `intended_address` is a delivery hint and the exact match key
for Organization invitations, never an identity. Normalization (`tenancy.NormalizeAddress`, the single
implementation, unit-pinned): trim leading/trailing ASCII whitespace; require exactly one `@` with non-empty local
part and domain; refuse any whitespace or control character, a local part > 64 bytes, or > 254 bytes in total;
**preserve the local part byte-exact** (RFC 5321 §2.4: local-part case is the receiving host's to interpret; no
provider equivalence is assumed, so no case folding, no NFC and no dot/plus rewriting of the local part); the domain
is Unicode-NFC-normalized then lowercased per Unicode default case mapping, with IDN labels kept as given (no
punycode conversion — recorded; ASCII domains are therefore plain lowercase). The CHECK
`invitations_address_normalized` pins the syntax, the byte bound and `split_part(a,'@',2) = lower(split_part(a,'@',2))`
only (no whole-address `lower()`); matching is byte equality of normalized values, with **no email-based account
merge** (the canonical user is always the `(issuer, subject)` principal). The consumer's `VerifiedRecipient` is built
only from `access.ProviderAddressEvidence` (unexported fields: `issuer`, `subject`, `address`, `verifiedAt`,
`source`), constructed by `access.AttestAddress(...)` whose production call sites are pinned to
`core/internal/identity/**` (none exist in G019; G058's reviewed adapter adds the first after obtaining **current**
provider evidence for the exact session subject — for WorkOS the candidate fields are the user object's `email` and
`email_verified`, documented but not qualified here; Q3) — a `verified` boolean carried by an adapter request is not
evidence and cannot reach the constructor. G019 pins the contract and tests it synthetically; no provider probe.

**Retirement.** `RetirementEligibility(ctx, session)`: IDENTITY shape (`InitContext(user, nil, nil)`), one call of
`app.rls_sole_owner_scopes()` (DEFINER, §4.3): returns the subject's sole-owned scopes as `(organization_id,
scope_type, scope_id)` identifiers plus a count — the subject sees only their own blocking scopes, never another
tenant's members; the web Account page shows the count and the per-Organization names only for the current tenant
(§1.6). `RetireUser(ctx, user)` (trusted; no route; operator/G058-era caller): refuses with the same list while
non-empty (`retirement_blocked`); else, for each Organization of the subject in ascending id, one guarded departure
transaction (the `LeaveScope` sequence under an operator-established Organization-only context for that user —
`InitContext(user, org, nil)`, not a tenant administrator's context), then `identity.DisableUser`. **The initial
global eligibility is a read/preflight, not a promise (C3):** each per-Organization departure is an authoritative
operation that re-locks and rechecks (`last_owner`, `dependent_bindings`, audit faults) and may refuse later because
of concurrent change; the call then stops, reports the Organizations already departed and the remaining list
(`RetirementBlockers` + `Remaining`), leaves the user enabled, and is resumable (re-running continues from the
remaining list; nothing is undone). No physical deletion, no global delete/disable endpoint, no tenant-administrator
path (Spec R3).

**Emergency suspension.** Unchanged: `DisableUser` is immediate for a last owner; bindings retained; the owner count
counts the disabled owner (assigned cardinality ≠ available operators, Spec R3). Recovery when all owners are
unavailable: the documented restricted procedure (RECOVERY.md addition) — the superuser repair path inserts or
re-enables an explicit owner binding with a ledger entry of the existing kinds (`membership.downgrade`/`remove` for
undo, and a recorded manual grant), under the existing incident runbook; **no new script, no automatic promotion, no
credential revival**; exercised synthetically in `TestRecoveryPost` (§6.2 MEM-15b).

**Trusted bootstrap and the legacy entry points (C2, Q4).** Caller inventory at `9cf9beed`/`46f1dc5e`: the
`tenancy.Service` primitives `CreateOrganization`, `CreateGroup`, `GrantOrganizationRole`, `ChangeOrganizationRole`,
`RemoveOrganizationMembership`, `GrantGroupRole`, `ChangeGroupRole`, `RemoveGroupMembership` (`service.go:74-466`) and
`tenancy.NewService` have **no non-test caller** in `core/`, `infra/` or `cmd/`; the only production Group creation is
G018's guarded `access.CreateGroup` (trigger-owned owner); Organizations are created in production by **nothing**
(the seed `mcp-seed.sql` runs as superuser in test environments only; `infra/recovery/sql/replay-restrict.sql:6-14`
mutates bindings as superuser under the recovery runbook). Therefore G019: (a) adds
`tenancy.CreateOrganizationWithOwner(ctx, name, owner UserID) (Organization, Binding, error)` — one transaction,
`sqlInsertOrganization` then `sqlInsertBinding` (`org_owner`, `membership`) — as the **only supported creation shape**
(privileged pool; the trigger returns early for non-runtime roles, P-5; atomicity is the transaction's, MEM-1);
(b) moves the ownerless `CreateOrganization` and the unguarded `Grant*/Change*/Remove*` primitives to
`tenancy/fixture.go`, keeps their names (no test-rename campaign), documents them as **fixture-only, not application
operations**, and pins that status with `TestTenancyFixturePrimitives` (boundary test: the inventory of call sites is
exactly the `_test.go` files recorded in the Result; any production caller fails the build's test) — not a warning;
(c) leaves the seed and the recovery replay as the explicitly isolated superuser paths (the seed gains owners for
browser scenarios, §3; RECOVERY.md documents that an operator-authored restriction that removes a last owner must be
followed by the explicit owner re-grant step, since the guard does not run for the superuser replay). No historical
migration rewrite, no production ownership repair, no new supported ownerless path (Spec R2).

**Retry / double submit / unknown commit (pinned, no intent key):**

| Operation | Duplicate request | Lost response after COMMIT | Changed payload |
|---|---|---|---|
| Issue invitation | a second issue is a distinct invitation (intent is immutable; duplicates to the same recipient are allowed up to the scope bound; the UI shows pending ones) | the list shows the committed invitation; the one-time token of a lost response is unrecoverable by design (no redirect, read or reissue can carry it) → the page says so and offers **Revoke** and a fresh invite | distinct invitation |
| Accept / consume | same token + same recipient: `already_member` only with a current qualifying binding, else `consumed` (no binding, nothing restored) | same read | a token for another invitation is `not_found` |
| Promote / demote / transfer | the state check on locked facts answers `conflict` (`recipient_ineligible` / `not_owner`) with the current roles, never "changed by this request" | the BFF re-reads the member list | `conflict` |
| Remove / leave | `not_found` (binding gone) → the BFF shows the current list | same | — |

### 1.5 Concurrency, lock order and the serialization argument (R2, AC3)

**Actual facade order (C3; from `facade.go:44-53`, `facts.go:93-182`, `store.go:81-109` at `46f1dc5e`), extended by
G019's new facts in the same sequence:** (1) `users` then `sessions` rows `FOR SHARE` (`LockLiveSession`,
`identity/store.go:531-532`); (2) the Organization row — `FOR NO KEY UPDATE` (`ExclusiveOrganization`) for
Organization-scope owner writes, Organization-scope removal/departure and Organization-scope issue; `FOR SHARE`
otherwise (`facts.go:93`); (3) the actor's own Organization binding `FOR SHARE` (`:103`); (4) the actor's own Group
bindings ascending `group_id` `FOR SHARE` (`:125`); (5) Group rows ascending id with per-row strength (`:172-182`:
below-target `FOR SHARE`, target `FOR NO KEY UPDATE`, above-target `FOR SHARE`; Group-scope writes and Group-scope
issue use the target; Organization departure/removal locks the subject's dependent Groups ascending `FOR NO KEY
UPDATE` as "targets"); (6) **new** `Subject` fact: the subject's (and transfer actor's) binding rows in one statement
`ORDER BY id FOR NO KEY UPDATE`; (7) the `invitations` row (`FOR NO KEY UPDATE` inside the claim, or the revoking
UPDATE); (8) the write statement, whose BEFORE trigger re-locks the scope row **at a strength the transaction already
holds** (no-op) — so the facade never upgrades a lock inside a trigger. Steps (3)–(4) before (5) is the existing G018
order and is kept (r0's "scope before every binding" claim was wrong for Groups).

**Direct runtime statements (C3).** A raw `UPDATE/DELETE app.role_bindings` locks the target binding row first and
only then takes the scope lock in its BEFORE trigger — the inverse of (2)/(5) → (6). That inversion can form a
wait-for cycle against a facade transaction (facade holds the scope row and waits for the binding; the direct
statement holds the binding and waits for the scope row). PostgreSQL detects it and aborts one transaction with
`40P01`; both outcomes are whole-transaction rollbacks with zero partial writes, mapped to `conflict` (`retry`) in the
facade. The Plan **does not claim cycles are impossible**: facade-vs-facade is cycle-free by the shared order, and
facade-vs-direct is detected, not prevented (O-MEM-3 pins both the detection and the rollback). The claim function
takes (1)-equivalent self `users` `FOR SHARE`, the invitation row, then binding rows ascending `id` `FOR SHARE`, and
no scope row, so it sits at the end of the order (it can wait for a facade, never the reverse except on the binding
rows it holds last). Revocations: a revocation that commits before the Organization lock is observed and denies; one
waiting behind it is not reported as having won (existing orderings A/B/C, reused).

**Why the post-wait recount is authoritative.** `app.role_bindings_authority_guard()` and `app.invitations_bound()`
are plpgsql **VOLATILE** (pinned in CT-MEM-1 with owner, `search_path` and `SECURITY INVOKER`); in READ COMMITTED each
SQL command inside a VOLATILE plpgsql function takes its own snapshot, so the guard issues the scope-row lock as one
command (`SELECT 1 FROM app.organizations WHERE id = … FOR NO KEY UPDATE` / `… app.groups …`) and the count/EXISTS as a
**separate following command**, which therefore sees every change committed by the lock holder it waited for. The
STABLE DEFINER readers it calls (`rls_owner_count`, `rls_dependent_group_bindings`) run inside that second command
and inherit its fresh snapshot; they never lock. Competing last-owner mutations serialize on the scope row and the
second recounts after the first commits (AC3: two owners each leaving → exactly one succeeds).

**Per-operation statement inventory (C5; counts are statements issued by core, excluding `InitContext`, the two
`LockLiveSession` locks and the audit trigger captures that fire inside the write statements):**

| Operation | Facts/locks | Writes | Trigger work inside the write |
|---|---|---|---|
| ChangeMemberRole / PromoteToOwner / DemoteOwner | Organization 1, own bindings 2, Groups 1–3 (Group scope), subject lock 1, owner count 1 | 1 UPDATE | scope lock 1 + eligibility reads ≤ 3 (user lock, Organization-member check, owner count) |
| TransferOwnership | as above + actor row in the subject lock statement | 2 (UPDATE, then UPDATE/DELETE) | 2 × (scope lock + count) |
| RemoveMember / LeaveScope (Organization) | Organization 1, own bindings 2, dependent Groups 1, subject bindings 1, owner count 1 | n Group DELETEs + 1 Organization DELETE | per DELETE: scope lock 1 + counts ≤ 2 |
| RemoveMember / LeaveScope (Group) | Organization 1, own bindings 2, Groups 1–3, subject lock 1 | 1 DELETE | scope lock 1 + count 1 |
| IssueInvitation | Organization 1 (exclusive for Organization scope), own bindings 2, Groups 1–3, pending count 1, (Group) member check 1 | 1 INSERT | `invitations_bound` scope lock 1 + count 1; issue audit capture |
| RevokeInvitation | Organization 1, own bindings 2 | 1 UPDATE | immutability check; audit capture |
| AcceptGroupInvitation / ConsumeOrganizationInvitation | (Group) Organization 1, own bindings 2 | 1 `invitation_claim` call | ≤ 7 commands inside the function (§1.4) |
| ListMembers / ListInvitations / ListMyInvitations / ListInvitationCandidates / GetInvitation / RetirementEligibility | Organization 1, own bindings 2 (tenant lists) or none (IDENTITY) | — | 1 page read (`limit+1`) |

No advisory locks; indexes only on `invitations` (§4.2). One E-PG observation per operation (`TestAccessFootprint`),
not a benchmark.

### 1.6 Web and BFF (R5, AC6)

Pattern P-9, server-rendered forms, no client authority, **provisional on G018's web slice (P-9..P-12)**:

| Route | Purpose |
|---|---|
| `GET /organization/members?after=&invitations_after=&role=` | Organization members (keyset page: display name, role, Self marker; `role=owner_candidates` filter so the promote/transfer picker pages through eligible admins — C5) with per-row controls the server decided (`can_change_role`, `can_remove`, `can_promote`, `can_demote`, `can_transfer` — display hints from the same snapshot; every action re-decides); pending invitations (keyset) with **Revoke**; **Invite** form (address, role admin/member, expiry); **Leave** with the truthful last-owner notice |
| `POST /organization/members/action` | one BFF route with a closed `action` field (`invite`, `revoke`, `change_role`, `remove`, `promote`, `demote`, `transfer`, `leave`) → the matching core route; every action except `invite` answers 303 to `/organization/members?status=<closed code>`; **`invite` answers a direct `200` `no-store` result page** (the invitation link `/invitations/<id>`, the token **once**, copy/paste instructions, the lost-response note; a redirect cannot carry the one-time token — C4); `conflict` → status page naming the field (`last_owner`, `recipient_ineligible`, `invitation_limit`, `context`) and the current state; transport failure → unknown-state page with a link to the list |
| `GET /groups/<uuid>/members?after=&candidates_after=` / `POST /groups/<uuid>/members/action` | the Group-scope twin; the recipient picker is fed by `GET /internal/v1/invitations/candidates` (current Organization members not in the Group, ids + display names, keyset ≤ 100 — the only candidate source a Group manager is authorized to read, C5/D-13), server-validated; `promote/demote/transfer` only for the Group's current owner |
| `GET /invitations?after=` | the signed-in member's own pending Group invitations (scope, role, expiry; keyset) → each links to `/invitations/<uuid>` |
| `GET /invitations/<uuid>` / `POST /invitations/<uuid>/accept` | acceptance page: shows the invitation the server reads for this recipient, a token field (prefilled by a ≤ 20-line inline script from `location.hash` when the link carried `#<token>`; otherwise pasted — the fragment never reaches a server), hidden `csrf_token`; POST sends the token in the form body → `POST /internal/v1/invitations/{id}/accept`; 303 to `/groups/<uuid>?status=joined|upgraded|already_member|consumed`; Organization invitations have **no web acceptance in G019** (the page says "not available yet"; G058 owns it) |
| `GET /account` (existing) | adds the retirement eligibility notice from `GET /internal/v1/account/retirement-eligibility` (count of sole-owned scopes and a sentence to transfer ownership in each workspace first; **no names, no scope ids rendered, no retire action** — the minimal private self-read, C6) |

Seven web routes (`WEB_ROUTES` +7: `organization/members/page.tsx organization/members/action/route.ts
groups/[id]/members/page.tsx groups/[id]/members/action/route.ts invitations/page.tsx invitations/[id]/page.tsx
invitations/[id]/accept/route.ts`; `WEB_ROUTES` 31 → 38). Shell: a third `WorkspaceArea` **`"members"`** (href
`/organization/members`) in G018's `WORKSPACE_AREAS` (P-9; the nav pins in `onboarding.spec`/`mcp.spec`/`shell.spec`
that count areas move 2 → 3) and a **Members** link on the Group page (`groups/[id]/page.tsx`, one line). Components:
adapt kit `select` and `table` (`components/ui`, `data-slot`, the P-10 three-line header with the kit sha, `UI_FILES`
9 → 11, `UI_SLOTS` +2, and one sentence added to `web/THIRD_PARTY_NOTICES.md` §shadcn/ui naming the two files "(G019)"
— existing licence text unchanged, no new package) and
`components/membership/{member-table,invitation-form,invitation-list,ownership-forms,accept-form}.tsx` (never
importing `@/lib/config` or `request-context`, P-15). `membership-client.ts` and the three new route handlers are added
to the exact `static_web_boundary` allow-list (`run.sh:860-861`) because they import `request-context` (P-15).
Keyboard/focus/error association, Light/Dark/System and 320 px reflow as G017/G018 (M-A11Y pattern). The token link is
`/invitations/<id>#<token>`: the fragment never reaches a server log; the BFF never puts the token in a URL,
`Location` header, page title or completion log (the `invite` result page is the only carrier, `no-store`); the
acceptance POST body is read by `readForm` (16 KiB) and forwarded in JSON. Privacy: lists show display names and roles
of the current scope only; the candidate list shows ids and display names of current Organization members to a
Group manager (D-13); no address of a pending invitation outside the issuing scope's administrators; no cross-tenant
directory; the retirement notice names no tenant.

### 1.7 Configuration

None expected: no new environment variable, label, vendor, database role or pool. Invitation lifetime bounds are
constants (`tenancy.InvitationDefaultLifetime = 168h`, `InvitationMaxLifetime = 720h`, `InvitationMinLifetime = 1h`,
`MaxPendingInvitationsPerScope = 100`, `MaxMembersPage = 100`; Q8 accepted, enforced per §1.4/§4.5 and paged per §5.3).

## 2. Source binding (provisional; rebind at R0)

| Input | Value |
|---|---|
| Target | `origin/staging` `9cf9beed16c96a94d6254ef2108a91e35fd1194d` (r1; management-only ahead of `63d12bf6` main) |
| Provisional predecessor | `origin/g018-group-business-profile-onboarding` `46f1dc5e0fafbb248acb791fcf913dc46d2a50de` observed at r1 (draft MR !41; slice 4 WIP) |
| Goal / Spec / review | SHA-256 in the header |
| Migration name | `core/migrations/00015_membership_invitations.sql` — **provisional** (the Goal reserves no number; R0 confirms 00015 is the next after the integrated G018 and that no parallel migration landed) |
| Execution | after Codex Plan ACCEPT, allow-list import, initial Result, Git-first binding, resource readiness, phase authority and recorded release; implementation model/effort Opus high (owner direction) |

## 3. Files (touch-list; actual inventory, G018 C5 pattern)

**Create (core):** `core/migrations/00015_membership_invitations.sql`; `core/internal/tenancy/membership.go` (Tx
writers/readers: `ChangeRoleTx`, `RemoveBindingTx`, `SubjectBindingsLock`, `OwnerCount`, `MembersPage`,
`InvitationCandidatesPage`, `CreateOrganizationWithOwner`), `core/internal/tenancy/fixture.go` (the relocated
fixture-only primitives, §1.4), `core/internal/tenancy/invitations.go` (`Invitation` model, `NormalizeAddress`,
`InsertInvitationTx`, `RevokeInvitationTx`, `ClaimInvitationTx` (the `invitation_claim` call), `InvitationsPage`,
`MyInvitationsPage`, `PendingInvitationCount`, lifetime constants) + `_test.go` (unit: normalization table,
validation, `TestServiceRejectsBeforeSQL` extension, `TestTenancyFixturePrimitives`);
`core/internal/tenancy/access/membership.go` (member list/candidates/change/remove/leave), `access/ownership.go`
(promote/demote/transfer), `access/invitations.go` (issue/revoke/list/mine/get/accept, `ProviderAddressEvidence`,
`AttestAddress`, `VerifiedRecipient`, `ConsumeOrganizationInvitation`), `access/retirement.go`
(`RetirementEligibility`, `RetireUser`) + unit `membership_test.go` (page bounds, token/digest, evidence constructor
unreachable from JSON, `AttestAddress` call-site pin) + `membership_integration_test.go` (`TestMembership`,
MEM-1..15, O-MEM in `ordering_integration_test.go`); `core/internal/migrate/membership_integration_test.go`
(`TestMembershipCatalog`, CT-MEM-1, preflight and up/down); `core/internal/membership/api/http.go` + `http_test.go`.

**Create (web, provisional on G018 slices 3–4):** `web/lib/membership-client.ts`; `web/app/organization/members/
{page.tsx,action/route.ts}`; `web/app/groups/[id]/members/{page.tsx,action/route.ts}`; `web/app/invitations/
{page.tsx,[id]/page.tsx,[id]/accept/route.ts}`; `web/components/ui/{select,table}.tsx`; `web/components/membership/
*.tsx`; `web/tests/unit/membership-{client,forms}.test.mjs`; `web/tests/browser/membership.spec.ts`.

**Create (management):** `management/v0/results/G019-scoped-invitations-membership-ownership-result.md`,
`management/v0/results/G019-scope-allowlist.txt` (Codex import of §11.3).

**Modify (core production):** `core/cmd/server/main.go` (register `membershipapi`; new digest pin);
`core/internal/tenancy/{facts.go (Subject/OwnerCount facts, InitContext invitation argument), store.go (statements),
service.go (CreateOrganizationWithOwner; doc note on CreateOrganization), tenancy.go (Invitation types/roles helpers),
errors.go (codes/constraint maps), log.go (ops)}`; `core/internal/tenancy/access/{policy.go (4 capabilities, owner
rules), errors.go (fields), log.go (14 ops), facade.go (if a shared helper moves), context.go (consumer context
builder), tx.go (only if the IDENTITY/consumer shape needs a variant)}`; `core/internal/audit/audit.go` (actions
+3, resource type +1, operations +14, `TestCatalogParity` source → 00015).

**Modify (web production, provisional):** `web/components/layout/app-shell.tsx` (`"members"` area in
`WorkspaceArea`/`WORKSPACE_AREAS`), `web/app/groups/[id]/page.tsx` (Members link; G018 file), `web/app/account/page.tsx`
(eligibility notice via the new client method), `web/playwright.config.ts` (`membership-light/dark`),
`web/THIRD_PARTY_NOTICES.md` (one sentence for `select`/`table`; P-10 lesson — listed in scope now, not by amendment).

**Modify (tests and pins, core):** `core/internal/tenancy/{boundary,parity,integration,tenancy}_test.go`;
`core/internal/tenancy/access/{boundary_test (importers +1, methods, inContextTx sites, definer allowance 00015 = 6, `AttestAddress` call sites),
matrix_test (19 capabilities; cells recomputed), policy_test, rls_expected_test (oracle r4 rows), attack_expected_test,
integration_test (IT-MATRIX-1 count), ordering_integration_test (O-MEM-1..3), rls_integration_test, attack_integration_test,
audit_integration_test (AT catalog), footprint_integration_test (E-PG rows), recovery_integration_test (MEM-15b)}.go`;
`core/internal/audit/audit_test.go`; `core/internal/migrate/{runner,integration,rls_integration,audit_integration,
tenancy_integration (the `invitation` table-name regex at `:411`), onboarding_integration}_test.go`;
`core/migrations/migrations_test.go`; `core/internal/mcp/env_integration_test.go` (`current_version\t15`).

**Modify (tests, web):** `web/tests/browser/{shell,onboarding,mcp}.spec.ts` (S-LANDMARKS/S-NAV wording; the
workspace-area count 2 → 3; Group Members link), `web/tests/browser/support/{assets,modes}.ts` if a pin moves.

**Modify (fixtures/seeds):** `infra/identity/mcp-seed.sql` (owners for Org B/C and Groups design/operations/planning;
a second Organization admin and a Group admin for transfer/promotion scenarios; M1 users; **reconciled with G018's
slice-4 seed at R0** — at `46f1dc5e` its 26-line change is WIP and unvalidated).

**Modify (harness, bounded):** `infra/validation/run.sh` (version-15 pins; `membership_checks()` two rows after
`onboarding_checks`; fallback FAIL list; `MAIN_GO_*`/importer strings; `CORE_MEMBERSHIP_ROUTES` pin (15 pairs);
B-TENANCY-HTTP positive pins for the new roots; `WEB_ROUTES` +7; `static_web_boundary` allow-list +4 exact entries;
`UI_FILES`/`UI_SLOTS` +2 via `lib.sh`; `BROWSER_MATRIX_IDENTITY_E2E` +1 line; `TestAccessOrdering`/`TestAccessPolicy`
`it_count` pins), `infra/validation/lib.sh`, `infra/validation/required-checks.tsv` (+2 rows),
`infra/ci/profile-paths.tsv` (+2 foundation rows: `web/app/organization/members/`, `web/app/invitations/`;
`web/app/groups/` exists at `:102`), `infra/ci/validate-controls.sh` (+1 `F-widen-*` row),
`infra/validation/tools/internal/logrecord/{events.json,completion.schema.json,logrecord_test.go}` (+14 paths, +1
event `membership.request`, +16 operations, +4 capabilities, rules).

**Modify (docs):** `docs/foundation/{TENANCY,AUTHORIZATION,DATABASE-ISOLATION,AUDIT,MIGRATIONS,WEB-SHELL,FOLLOW-UP,
RECOVERY,IDENTITY}.md`, `docs/{SECURITY,PERFORMANCE,ARCHITECTURE}.md` (implementation records only; no new
decision text — architecture changes are escalated, not documented into existence).

**Not touched:** `infra/recovery/**` (no new ledger kind: `membership.downgrade`/`remove` cover undo; the recovery
procedure is documentation), `core/internal/identity/**` (`DisableUser` reused as is; if `RetireUser` needs an export
beyond the existing `Service.DisableUser`, that is a §12 Q5 scope question, not a silent addition), `core/internal/mcp/**`
production, `web/lib/session.ts` (G018's `maxBytes` option suffices; forms ≤ 16 KiB), any historical migration.

## 4. Schema and migration `00015_membership_invitations.sql` (P-5 style; flagged items for explicit review: §4.3 DEFINER readers, §4.4 policies/grants, §4.5 trigger guards, §4.6 `events_attribute` replacement, §4.7 preflight)

### 4.1 Precondition `DO`

Owner role; version-14 shape (the 00014 postconditions' objects exist exactly; none of the 00015 objects exist;
runtime holds no DELETE on `role_bindings`; `role_bindings_lock` exists with `WITH CHECK (false)`; runtime SELECT on
`role_bindings` is the 00006:401 seven-column list); `has_function_privilege` for `app.is_organization_member` is
true for runtime (00011:775-776) and **false for the owner role → 00015 grants EXECUTE to the owner role** (the
FORCE-bound owner reaches the guard too); the `audit.capture_*` execution mode is inspected (`prosecdef` of the
capture functions and the `audit.events` ACL): if the captures run as invoker, 00015 grants `<db>_auth_reader` INSERT
on `audit.events` for the depth ≥ 2 captures fired inside `invitation_claim` (both branches pinned in CT-MEM-1);
**legacy owner preflight (§4.7)** — refuse unless every `organizations` row has ≥ 1 `org_owner` binding and every
`groups` row ≥ 1 `group_owner` binding.

### 4.2 Table `app.invitations`

```
id uuid PK (core-generated), organization_id uuid NOT NULL FK organizations ON DELETE RESTRICT,
scope_type text NOT NULL, scope_id uuid NOT NULL, group_id uuid NULL,
  FK (organization_id, group_id) → groups(organization_id, id) ON DELETE RESTRICT,
role text NOT NULL, issuer_user_id uuid NOT NULL FK users RESTRICT,
intended_user_id uuid NULL FK users RESTRICT, intended_address text NULL,
token_digest bytea NOT NULL UNIQUE CHECK (octet_length(token_digest) = 32),
created_at timestamptz NOT NULL DEFAULT now(), expires_at timestamptz NOT NULL,
revoked_at timestamptz NULL, consumed_at timestamptz NULL, consumed_by_user_id uuid NULL FK users RESTRICT,
consumed_binding_id uuid NULL
CHECK invitations_scope_consistent: (scope_type='organization' AND scope_id=organization_id AND group_id IS NULL
  AND role IN ('org_admin','org_member') AND intended_address IS NOT NULL)
  OR (scope_type='group' AND group_id IS NOT NULL AND scope_id=group_id AND role IN ('group_admin','group_member')
  AND intended_user_id IS NOT NULL AND intended_address IS NULL)
CHECK invitations_address_normalized: intended_address IS NULL OR (
  octet_length(intended_address) BETWEEN 3 AND 254 AND octet_length(split_part(intended_address,'@',1)) <= 64
  AND intended_address ~ '^[^@[:space:][:cntrl:]]+@[^@[:space:][:cntrl:]]+$'
  AND split_part(intended_address,'@',2) = lower(split_part(intended_address,'@',2)))   -- local part byte-exact (C4)
CHECK invitations_lifetime: expires_at > created_at AND expires_at <= created_at + interval '30 days'
CHECK invitations_consumption_shape: (consumed_at IS NULL) = (consumed_by_user_id IS NULL)
  AND (consumed_at IS NULL OR revoked_at IS NULL) AND (consumed_at IS NOT NULL OR consumed_binding_id IS NULL)
CHECK invitations_consumed_is_intended: consumed_by_user_id IS NULL OR intended_user_id = consumed_by_user_id
INDEX invitations_scope_idx (organization_id, scope_type, scope_id, id)
INDEX invitations_pending_idx (organization_id, scope_type, scope_id, expires_at)
  WHERE consumed_at IS NULL AND revoked_at IS NULL                      -- serves the pending-unexpired count (C5)
INDEX invitations_recipient_idx (intended_user_id, id) WHERE consumed_at IS NULL AND revoked_at IS NULL
```
Triggers: `invitations_keys_immutable` (BEFORE UPDATE; owner-owned invoker, 00003 pattern; **no early return for any
role**): every column immutable except — for `current_user` runtime/owner — `revoked_at` NULL→value on an unconsumed
row; and — for `current_user = <db>_auth_reader` only — the consumption transition (`consumed_at`,
`consumed_by_user_id`, `consumed_binding_id`) NULL→value together on an unrevoked row, with `intended_user_id`
NULL→value admitted only in that same row update and only for `scope_type='organization'`. A revoked row cannot be
consumed and vice versa; `expires_at` never changes (changing intent = revoke + reissue, Spec R4).
`invitations_bound` (BEFORE INSERT; owner-owned invoker, VOLATILE; runtime/owner only): locks the scope row `FOR NO
KEY UPDATE` as one command, then counts pending unexpired invitations of the scope as a second command; ≥ 100 →
`integrity_constraint_violation` `invitation_limit` (C5). `audit.capture_invitations` (AFTER INSERT / AFTER UPDATE;
`audit.capture_*` pattern, strict): `invitation.issued`, `invitation.revoked`, `invitation.consumed` with
`resource_type='invitation'`, `subject_user_id` = intended/consumed user (NULL for an address-addressed pending
invitation), `role`, `group_id`; **never the digest or the address** (the shape CASE in `events_attribute`, §4.6,
refuses `reason` for these actions as for membership actions). `FORCE ROW LEVEL SECURITY`; owner bypass as the other
`app` tables.

**No new context variable** (r0's `app.invitation_id`/`app.ctx_invitation()` are withdrawn: a transaction setting is
not recipient proof, C1). Context parsers unchanged.

### 4.3 SECURITY DEFINER functions (allowance for 00015 = **6**: four `<db>_policy_reader`-owned STABLE readers, two
`<db>_auth_reader`-owned VOLATILE functions — the user lock and the claim; every one `SET search_path = pg_catalog, pg_temp`, PUBLIC revoked, EXECUTE to
runtime and owner only; owner/volatility/body pinned in CT-MEM-1)

Helper policies for pr (the existing `role_bindings_policy_reader` is **own-row only**, 00006:476-477, so it cannot
serve counts over other subjects): `role_bindings_scope_reader` FOR SELECT TO pr USING `organization_id =
app.ctx_org()`; `role_bindings_retirement_reader` FOR SELECT TO pr USING `subject_user_id = app.ctx_user() OR role IN
('org_owner','group_owner')` (D-2); `users_policy_reader` FOR SELECT TO pr USING `(true)` with `GRANT SELECT (id,
display_name, status) ON app.users TO pr` (candidates and enabled checks; D-13). Each reader **answers only for a
scope the caller can currently see** (C1): Organization scope needs `app.rls_org_role() IS NOT NULL`; Group scope
needs the Group visible under `groups_select` (selected-Group rule included); otherwise it returns 0/NULL/no rows —
a direct runtime call cannot learn an inaccessible sibling's counts or bypass the selected context.

- `app.rls_owner_count(p_scope_type text, p_scope_id uuid) RETURNS integer` STABLE: owner-role bindings of the scope
  in `ctx_org()`. A **count only**; serialization is the calling trigger's separate lock command (§1.5, §4.5; D-1).
- `app.rls_dependent_group_bindings(p_subject uuid) RETURNS integer` STABLE: the subject's Group-scope bindings in
  `ctx_org()` regardless of `ctx_group()`; callable usefully only by `MGR` or for `p_subject = ctx_user()` (else 0).
- `app.rls_sole_owner_scopes() RETURNS TABLE (organization_id uuid, scope_type text, scope_id uuid)` STABLE: for
  `ctx_user()`, every scope where the subject holds an owner role and that scope's owner count is 1, across all
  Organizations (IDENTITY shape; D-2). Identifiers only; nothing when `ctx_user()` is NULL.
- `app.rls_group_invitation_candidates(p_group uuid, p_after uuid, p_limit int) RETURNS TABLE (user_id uuid,
  display_name text)` STABLE: current Organization members (`scope_type='organization'`, `source='membership'`,
  `users.status='active'`) of `ctx_org()` without a binding in `p_group`, keyset by `user_id`, `p_limit` ≤ 101; only
  when the caller is `MGR` or `app.rls_group_role(p_group) IN ('group_owner','group_admin')` and (`ctx_group()` IS NULL
  or = `p_group`); else no rows (C5; D-13).
- `app.rls_user_enabled_lock(p_user uuid) RETURNS boolean` **VOLATILE**, owned by `<db>_auth_reader`: `SELECT status
  FROM app.users WHERE id = p_user FOR SHARE` then returns `status = 'active'` — the lock conflicts with
  `DisableUser`'s `FOR UPDATE` (`identity/store.go:968`), so a promotion that waits behind a disable re-reads the
  disabled status (C2). Needs `GRANT UPDATE (status) ON app.users TO ar` and `users_auth_lock` FOR UPDATE TO ar USING
  `(true)` WITH CHECK `(false)` (lock, never write; pinned).
- `app.invitation_claim(p_id uuid, p_digest bytea, p_address text) RETURNS TABLE (outcome text, organization_id uuid,
  scope_type text, scope_id uuid, group_id uuid, role text, binding_id uuid)` **VOLATILE**, owned by
  `<db>_auth_reader`: the §1.4 consumption sequence; refuses with `insufficient_privilege` when `ctx_user()` is NULL
  or `current_user` is not runtime/owner; the single write body among the ar-owned functions (CT-MEM-1 inventories
  every ar-owned function with its volatility so a second write body cannot appear unnoticed).

### 4.4 Grants and policies (`TO <db>_runtime, <db>_owner`; U/O/G/R/MGR as 00006; `GMGR(g)` = `app.rls_group_role(g)
IN ('group_owner','group_admin')`; `OWN_ROLE(x)` = `x IN ('org_owner','group_owner')`)

- **`role_bindings` (runtime/owner)**: `GRANT DELETE ON app.role_bindings TO <runtime>` (new); runtime SELECT
  gains **`id`** (7 → 8 columns: the `Subject` lock, `ChangeRoleTx`/`RemoveBindingTx` keyed by `id` and the claim's
  `consumed_binding_id` need it; `id` is a random key under the unchanged row filter — the one grant widening, C1);
  INSERT grant unchanged (runtime gets **no** new INSERT shape). `DROP POLICY role_bindings_lock`;
  **`role_bindings_update`** FOR UPDATE TO runtime, owner USING `organization_id = O AND source='membership' AND
  (subject_user_id = U OR (scope_type='organization' AND MGR) OR (scope_type='group' AND (G IS NULL OR group_id = G)
  AND (MGR OR GMGR(group_id))))` WITH CHECK `organization_id = O AND source='membership' AND ((scope_type='organization'
  AND MGR) OR (scope_type='group' AND (G IS NULL OR group_id = G) AND (MGR OR GMGR(group_id))))` — owner-role
  transitions, eligibility and last-owner protection are the guard's (§4.5), because a policy cannot see `OLD`. The
  own-row USING branch preserves `ReadFacts`' `FOR SHARE` on own bindings (P-13) and admits self-leave.
  **`role_bindings_delete`** FOR DELETE TO runtime, owner USING (the same expression as the update USING).
  `role_bindings_insert_bootstrap` and `role_bindings_select` are **untouched** (no shared reader body changes;
  `role_bindings_select` hides the issuer from a plain recipient — which is why the issuer check lives in the
  ar-owned claim, not in a runtime read). `role_bindings_policy_reader`/`_auth_reader` untouched.
- **`role_bindings` (`<db>_auth_reader`, the claim's write boundary)**: `GRANT INSERT (id, organization_id,
  subject_user_id, scope_type, scope_id, group_id, role, source), UPDATE (role) ON app.role_bindings TO ar`;
  **`role_bindings_insert_invited`** FOR INSERT TO ar WITH CHECK `subject_user_id = app.ctx_user() AND
  source='membership' AND NOT OWN_ROLE(role)`; **`role_bindings_auth_upgrade`** FOR UPDATE TO ar USING `(true)` (lock
  any row: the issuer's and the recipient's) WITH CHECK `subject_user_id = app.ctx_user() AND NOT OWN_ROLE(role)` (write
  only the subject's own non-owner row). ar has no DELETE. The organization/scope values come from the immutable
  invitation row inside the function, so no `organization_id = O` term is needed (the consumer runs with `ctx_org()`
  NULL). Helper policies for pr: §4.3.
- **`invitations`**: runtime `SELECT (id, organization_id, scope_type, scope_id, group_id, role, issuer_user_id,
  intended_user_id, intended_address, created_at, expires_at, revoked_at, consumed_at, consumed_by_user_id,
  consumed_binding_id)` — **15 columns, `token_digest` excluded** (C1; CT-MEM-1 greps that no Go statement names it
  outside the INSERT and the claim call); `INSERT (id, organization_id, scope_type, scope_id, group_id, role,
  issuer_user_id, intended_user_id, intended_address, token_digest, expires_at)`; `UPDATE (revoked_at)` — one column;
  **no DELETE**. Policies (runtime, owner): `invitations_select` USING `organization_id = O AND
  ((scope_type='organization' AND MGR) OR (scope_type='group' AND (G IS NULL OR group_id = G) AND (MGR OR
  GMGR(group_id))) OR intended_user_id = U OR consumed_by_user_id = U)`; `invitations_insert` WITH CHECK
  `organization_id = O AND issuer_user_id = U AND ((scope_type='organization' AND G IS NULL AND MGR) OR
  (scope_type='group' AND (G IS NULL OR group_id = G) AND (MGR OR GMGR(group_id)) AND
  app.is_organization_member(intended_user_id, O)))`; `invitations_update` USING `organization_id = O AND
  ((scope_type='organization' AND MGR) OR (scope_type='group' AND (G IS NULL OR group_id = G) AND (MGR OR
  GMGR(group_id))))` WITH CHECK (the same) — revocation only; the transition legality is the immutability trigger's.
  For ar: `GRANT SELECT, UPDATE (consumed_at, consumed_by_user_id, intended_user_id, consumed_binding_id) ON
  app.invitations TO ar`; `invitations_auth_reader` FOR SELECT TO ar USING `(true)`; `invitations_auth_claim` FOR
  UPDATE TO ar USING `(true)` WITH CHECK `consumed_by_user_id = app.ctx_user()`. No `invitations` policy for pr.
- **No consume trigger.** The binding write is the claim function's own body (§1.4, §4.3): one call, one
  transaction, outcomes returned as a row (never raised for the closed `not_found`/`consumed`/`already_member` cases,
  raised as `integrity_constraint_violation` `issuer_authority_lost` / `recipient_ineligible` for the conflict cases
  so the facade maps them with `checkCodes`). The function never downgrades, never writes an owner role (CHECK +
  policy), and sets `consumed_binding_id` to the inserted or upgraded binding's `id`.

### 4.5 Authority and cardinality guard `app.role_bindings_authority_guard()` (SECURITY INVOKER, owner-owned,
`search_path = pg_catalog`; BEFORE UPDATE OF role / BEFORE DELETE ON app.role_bindings FOR EACH ROW)

plpgsql **VOLATILE** (each command below is a separate SQL command with its own READ COMMITTED snapshot, §1.5).
Returns early unless `current_user` is runtime, owner or `<db>_auth_reader` (superuser repair, the seed and the
recovery replay are unchanged and documented, §1.4). Steps: (1) **owner authority** — if `OWN_ROLE(OLD.role)` or
(TG_OP='UPDATE' and `OWN_ROLE(NEW.role)`): Organization scope requires `app.rls_org_role() = 'org_owner'`; Group scope
requires `app.rls_group_role(OLD.group_id) = 'group_owner'` (an Organization owner/admin without that Group role is
refused — no override, Q2); the ar upgrade never reaches here (invitation roles are never owner roles); a
self-demotion/self-leave by an owner passes this step and meets step 4. Refusal: `insufficient_privilege` CONSTRAINT
`owner_authority` (→ `forbidden`). (2) **promotion eligibility (C2)** — if TG_OP='UPDATE' and NOT `OWN_ROLE(OLD.role)`
and `OWN_ROLE(NEW.role)`: require `OLD.role` = the admin role of the same scope (`org_admin` → `org_owner`,
`group_admin` → `group_owner`; a member or any other role is refused), `app.rls_user_enabled_lock(OLD.subject_user_id)`
(locks the recipient's `users` row `FOR SHARE`, then reads `active` with the post-lock snapshot) and, for Group scope,
`app.is_organization_member(OLD.subject_user_id, OLD.organization_id)`; refusal `integrity_constraint_violation`
`recipient_ineligible` (→ `conflict`). Demotion (`OWN_ROLE(OLD)` and NOT `OWN_ROLE(NEW)`) requires `NEW.role` ∈ the
scope's admin/member roles (already a CHECK). (3) **departure context and dependent grants (C3)** — TG_OP='DELETE'
and Organization scope: `app.ctx_group() IS NOT NULL` → `insufficient_privilege` `context_required` (→
`invalid_input`); then `app.rls_dependent_group_bindings(OLD.subject_user_id) > 0` → `integrity_constraint_violation`
`dependent_bindings` (→ `conflict`) — the DEFINER count sees every sibling Group regardless of the invoker's selected
context, so no dormant Group grant survives any Organization departure path. (4) **last owner** — if `OWN_ROLE(OLD.role)`
and (TG_OP='DELETE' or NOT `OWN_ROLE(NEW.role)`): command A locks the scope row (`SELECT 1 FROM app.organizations WHERE
id = OLD.scope_id FOR NO KEY UPDATE` or `… app.groups …`; admitted by the existing `organizations_update`/`groups_update`
USING for any member of the Organization, 00006:434,439 — pinned in CT-MEM-1 for every actor role that can reach
this step); command B evaluates `app.rls_owner_count(OLD.scope_type, OLD.scope_id) - 1 < 1` →
`integrity_constraint_violation` `last_owner` (→ `conflict`). Command B's snapshot is taken after command A's wait,
so it sees every committed owner change and this transaction's own prior statements (transfer: recipient already an
owner). INSERTs are not guarded here (bootstrap policy; ar insert policy). The G018 `groups_bootstrap_owner` trigger
inserts only and never reaches this guard; the ar upgrade (member → admin) reaches it and passes (no owner role).

### 4.6 Audit catalog

`events_action_valid` +3 (`invitation.issued`, `invitation.revoked`, `invitation.consumed`), `events_resource_type_valid`
+1 (`invitation`), `events_operation_valid` +14 (§5.2); `audit.events_attribute()` **replaced** (00013 precedent) only
to add the `invitation.*` shape CASE (resource `invitation`, `role` valid for the scope, reason NULL, no
`previous_role`) and to list the three actions as trigger-owned; the depth-1 scope rule and every other branch are
byte-identical (hash-diffed in CT-MEM-1 against the 00013 body with the new CASE removed). Reasons unchanged (58).
Runtime grants on `audit.events` unchanged.

### 4.7 Legacy preflight, down and risk

Preflight refuses with counts only (`<n> organizations / <m> groups without an owner binding`), never ids or names;
an operator resolves each through the documented restricted procedure (§1.4 "Emergency suspension") **before**
re-running the migration — no automatic promotion or deletion (Spec R2). Known at planning: the identity seed
(`infra/identity/mcp-seed.sql`) creates ownerless Groups and Organizations B/C **after** migration on an empty database
(preflight passes; the fixtures are amended in slice 4 so that browser scenarios have owners); integration fixtures
using `CreateOrganization` run above version 15 and are unaffected (the invariant guards mutations, not existence);
**the persisted databases the operator names are inventoried at R0** (Q1: scoped readiness check, no unrelated-host
scan; an actual ownerless record is escalated for an attributable owner/operator resolution — the Plan asserts no
environment is clean). **Down (C3):** a precondition `DO` **refuses the down migration transactionally** when
`app.invitations` holds any row or `audit.events` holds any `invitation.*` action (`<n> invitations / <m> audit rows`
in the message, no ids), because the old shape cannot represent that history and history is never erased to regain
an old constraint; when empty: drop the runtime/ar policies and grants on `role_bindings` (restore
`role_bindings_lock` and the seven-column SELECT), drop the pr/ar helper policies and grants, the six DEFINER
functions, the guard and bound trigger/functions, the immutability trigger, the `invitations` table; restore the 00013
`events_attribute` body and the 00014 constraint lists. Risk: `ADD CONSTRAINT`/`DROP POLICY` take short ACCESS
EXCLUSIVE locks on `role_bindings` (small table; within the migration transaction's normal bounds); no rewrite; no
data mutation.

## 5. Contracts and interfaces

### 5.1 Go surfaces (`core/internal/tenancy/access`)

```go
type Scope struct { Kind tenancy.ScopeType; Group *tenancy.GroupID }          // Organization scope: Group == nil
type Member struct { UserID tenancy.UserID; DisplayName *string; Role string; Self bool; Disabled bool
    CanChangeRole, CanRemove, CanPromote, CanDemote, CanTransfer bool }        // hints from the same snapshot
type MemberPage struct { Members []Member; Next *tenancy.UserID }
type Invitation struct { ID tenancy.InvitationID; Scope Scope; Role string; IssuerDisplayName *string
    IntendedUserID *tenancy.UserID; IntendedAddress *string /* admin views only */; ExpiresAt time.Time
    Status string /* pending|expired|revoked|consumed */; ConsumedBy *tenancy.UserID }
type InvitationPage struct { Invitations []Invitation; Next *tenancy.InvitationID }
type IssuedInvitation struct { Invitation Invitation; Token string /* once */ }
type Recipient struct { UserID *tenancy.UserID; Address *string }             // exactly one set
type AcceptOutcome struct { Outcome string /* joined|upgraded|already_member|consumed */; Binding *tenancy.Binding /* nil for consumed */ }
type Candidate struct { UserID tenancy.UserID; DisplayName *string }
type CandidatePage struct { Candidates []Candidate; Next *tenancy.UserID }
type RetirementBlockers struct { Count int; Scopes []ScopeRef /* ids only; ≤ 100 */; Truncated bool }
type RetirementProgress struct { Departed []tenancy.OrganizationID; Remaining []tenancy.OrganizationID; Blockers RetirementBlockers }
func (s *Service) ListMembers(ctx, tc, sc Scope, after string, limit int, filter string /* ""|owner_candidates */) (MemberPage, error)
func (s *Service) ListInvitations(ctx, tc, sc Scope, after string, limit int) (InvitationPage, error)
func (s *Service) ListMyInvitations(ctx, tc, after string, limit int) (InvitationPage, error)
func (s *Service) ListInvitationCandidates(ctx, tc, group tenancy.GroupID, after string, limit int) (CandidatePage, error)
func (s *Service) IssueInvitation(ctx, tc, sc Scope, role string, r Recipient, lifetime time.Duration) (IssuedInvitation, error)
func (s *Service) RevokeInvitation(ctx, tc, id tenancy.InvitationID) error
func (s *Service) GetInvitation(ctx, tc, id tenancy.InvitationID) (Invitation, error)
func (s *Service) AcceptGroupInvitation(ctx, tc, id tenancy.InvitationID, token string) (AcceptOutcome, error)
func (s *Service) ChangeMemberRole(ctx, tc, sc Scope, subject tenancy.UserID, role string) (Member, error)
func (s *Service) RemoveMember(ctx, tc, sc Scope, subject tenancy.UserID) error
func (s *Service) PromoteToOwner(ctx, tc, sc Scope, subject tenancy.UserID) (Member, error)
func (s *Service) DemoteOwner(ctx, tc, sc Scope, subject tenancy.UserID, role string) (Member, error)
func (s *Service) TransferOwnership(ctx, tc, sc Scope, recipient tenancy.UserID, formerRole string) error
func (s *Service) LeaveScope(ctx, tc, sc Scope) error
func (s *Service) RetirementEligibility(ctx, credential, origin string) (RetirementBlockers, error) // IDENTITY shape; served by HTTP (C6)
func (s *Service) RetireUser(ctx, user tenancy.UserID) (RetirementProgress, error)                   // trusted; no route; resumable (C3)
// §5.5 — the bounded consumer for G058 and tests (C4)
type ProviderAddressEvidence struct { /* unexported: issuer, subject, address, verifiedAt, source */ }
func AttestAddress(issuer, subject, address string, verifiedAt time.Time, source string) (ProviderAddressEvidence, error) // call sites pinned to core/internal/identity/** and _test.go
type VerifiedRecipient struct { /* unexported: userID, evidence */ }
func NewVerifiedRecipient(userID tenancy.UserID, ev ProviderAddressEvidence) (VerifiedRecipient, error)                    // refuses a zero evidence value
func (s *Service) ConsumeOrganizationInvitation(ctx, r VerifiedRecipient, id tenancy.InvitationID, token string) (AcceptOutcome, error)
```
`tenancy`: `CreateOrganizationWithOwner`, `ChangeRoleTx`, `RemoveBindingTx`, `SubjectBindingsLock`, `OwnerCount`,
`MembersPage`, `InvitationCandidatesPage`, `InsertInvitationTx`, `RevokeInvitationTx`, `ClaimInvitationTx`,
`InvitationsPage`, `MyInvitationsPage`, `PendingInvitationCount`, `NormalizeAddress`, `NewInvitationToken`/
`InvitationDigest` (copied 32-byte/base64url/SHA-256 helpers), `InvitationID`/`ParseInvitationID`. **`InitContext` is
unchanged** (no invitation argument; the consumer uses the existing IDENTITY shape). Existing exported signatures
unchanged; the fixture primitives move file, not signature.

### 5.2 Capabilities (15 → 19), operations (36 → 52) and audit actions

Capabilities: `organization.manage_owners`, `group.manage_owners`, `organization.leave`, `group.leave`. Operations
(`access.denied`, `access.decision`, `events_operation_valid`; **16**): `membership.list`, `membership.change_role`,
`membership.remove`, `membership.leave`, `ownership.promote`, `ownership.demote`, `ownership.transfer`,
`invitation.list`, `invitation.mine`, `invitation.candidates`, `invitation.issue`, `invitation.revoke`,
`invitation.get`, `invitation.accept`, `invitation.consume`, `retirement.eligibility` (`RetireUser` logs
`retirement.retire` as a tenancy-level op without an `access.decision`, like the model primitives). Trigger actions
30 → 33, resource types 9 → 10, reasons 58.

### 5.3 HTTP — `core/internal/membership/api` (all authenticated; P-1 scope except the IDENTITY-shape row; JSON;
`no-store`; `maxBody = 8 << 10`)

**Fifteen method/path pairs on fourteen paths (C6)** — `completion.schema.json` route enums take the fourteen path
templates (two occurrences each); `CORE_MEMBERSHIP_ROUTES` pins the fifteen `METHOD path` pairs in C-sorted order
against `http.go` (the `CORE_ONBOARDING_ROUTES` mechanism, `run.sh:184,1203`); `WEB_ROUTES` takes the seven web
files. The literal segments `mine` and `candidates` are matched before the `{id}` template (pinned: a request for
`/internal/v1/invitations/mine` never parses `mine` as an id).

| Method and path | Body / query | Answer |
|---|---|---|
| `GET /internal/v1/members?origin=&scope=organization|group&group_id=&after=&limit=&role=` | — (`role=owner_candidates` optional) | `{"members":[{user_id, display_name|null, role, self, disabled, can_*}], "next"}` (≤ 100; keyset by `user_id`) |
| `POST /internal/v1/members/{user_id}/role` | `{origin, csrf_token, scope, group_id?, role}` | `200 {"member":{…}}`; `403 forbidden` (`owner_subject`); `409 conflict` (`recipient_ineligible`) |
| `POST /internal/v1/members/{user_id}/remove` | `{origin, csrf_token, scope, group_id?}` | `200 {}`; `403` (`owner_subject`); `409 conflict {"field":"last_owner"|"dependent_bindings"}` |
| `POST /internal/v1/members/{user_id}/promote` | `{origin, csrf_token, scope, group_id?}` | `200 {"member":{…}}`; `409` (`recipient_ineligible`) |
| `POST /internal/v1/members/{user_id}/demote` | `{origin, csrf_token, scope, group_id?, role}` | `200`; `409 {"field":"last_owner"}` |
| `POST /internal/v1/ownership/transfer` | `{origin, csrf_token, scope, group_id?, recipient_user_id, former_role}` | `200 {}`; `409` (`recipient_ineligible`) |
| `POST /internal/v1/members/leave` | `{origin, csrf_token, scope, group_id?}` | `200 {}`; `409 {"field":"last_owner"}` |
| `GET /internal/v1/invitations?origin=&scope=&group_id=&after=&limit=` | — | `{"invitations":[…], "next"}` (admin view; ≤ 100; keyset by id) |
| `POST /internal/v1/invitations` | `{origin, csrf_token, scope, group_id?, role, intended_user_id | intended_address, expires_in_hours?}` | `201 {"invitation":{…}, "token"}` (`no-store`; the only place the token appears); `409 {"field":"invitation_limit"}` |
| `GET /internal/v1/invitations/mine?origin=&after=&limit=` | — | the subject's own pending Group invitations `{"invitations":[…], "next"}` (≤ 100; **keyset**, C5) |
| `GET /internal/v1/invitations/candidates?origin=&group_id=&after=&limit=` | — | `{"candidates":[{user_id, display_name|null}], "next"}` (≤ 100; Group managers and `MGR` only; C5/D-13) |
| `GET /internal/v1/invitations/{id}` | — | `{"invitation":{…}}` (recipient view omits the address) |
| `POST /internal/v1/invitations/{id}/revoke` | `{origin, csrf_token}` | `200 {}`; `409` (`invitation_not_pending`) |
| `POST /internal/v1/invitations/{id}/accept` | `{origin, csrf_token, token}` | `200 {"outcome": "joined"|"upgraded"|"already_member"|"consumed", "group_id"}`; `404` for unknown/expired/revoked/foreign/wrong-token/wrong-recipient (**one** closed answer, no distinction that would confirm an invitation's existence to a non-recipient); `409` (`issuer_authority_lost` / `recipient_ineligible`) |
| `GET /internal/v1/account/retirement-eligibility?origin=` | — (session credential header; IDENTITY shape, no tenant context — the `/api/account` trust) | `{"count", "scopes":[{organization_id, scope_type, scope_id}], "truncated"}` (own sole-owned scope ids only, ≤ 100; no names, no tenant detail; **no retire action**) |

Closed codes as `onboarding/api` plus `membership_unavailable` 503. Errors carry no names, addresses or tokens. One
`membership.request` log event per request (operation, outcome, reason, host tag; never a token or address). **No
Organization-invitation consumption route** (G058). **No MCP change.** Internal web-BFF contracts, not public API;
B-TENANCY-HTTP positive pins: the roots `/internal/v1/members`, `/internal/v1/invitations`,
`/internal/v1/ownership/transfer` and `/internal/v1/account/retirement-eligibility` answer `400
ambiguous_parameter`/`401`/`403`/`405` closed JSON without a session; every unlisted method → 405; sub-paths → 404;
the existing 18/15 absence probes are kept (none is served; `TENANCY_PROBE_WEB` length re-pinned at R0).

### 5.4 Web — §1.6. `membership-client.ts` exports `listMembers`, `listInvitations`, `myInvitations`,
`invitationCandidates`, `issueInvitation`, `revokeInvitation`, `getInvitation`, `acceptInvitation`, `changeRole`,
`removeMember`, `promote`, `demote`, `transfer`, `leave`, `retirementEligibility` (**15**, one per pair); strict
shapes with per-endpoint `successBytes` (member row ≤ 36 + 1,200 + 80 → ≤ 1,400 bytes; 100 rows < 150 KiB;
invitation row ≤ 36×3 + 1,530 + 200 → ≤ 2 KiB; 100 rows < 210 KiB; candidate row ≤ 1,300 bytes; eligibility ≤ 100 ×
120 + 64 bytes); anything unexpected → `malformed`.

### 5.5 Downstream consumer contract (AC8; for G058 — defined and tested here, consumed there)

`ConsumeOrganizationInvitation` is the **only** way an Organization invitation becomes membership. Preconditions the
caller must establish and G019 rechecks: a canonical `UserID` mapped from the provider `(issuer, subject)` through
the existing principal path (never from the address); `ProviderAddressEvidence` attested by the identity adapter
from **current** provider evidence for that exact subject (`AttestAddress`, call sites pinned; a request-supplied
boolean cannot produce it — C4); exact `(id, token)`. Guarantees: atomic consumption + binding + audit (or nothing);
`joined`/`upgraded`/`already_member` truthfully and `consumed` for a used token whose membership is no longer
current (nothing restored); unknown/expired/revoked/consumed-by-another/foreign/mismatched-address → `not_found`;
issuer authority lost → `conflict` with the invitation left pending for its administrators to revoke/reissue (no
recipient-side revocation); disabled recipient → `conflict` (`recipient_ineligible`); no session issued; the
invitation's `organization_id` is the only tenant the binding can belong to. Tests cover a synthetic first-time user
(created by the test fixture through the existing identity mapping, then consumed) and an existing user from another
Organization (MEM-10). The public proof acquisition, callback intent, applicant state and session issuance remain
G058's and are **explicitly pending** in the Result.

## 6. Validation and test mapping

### 6.1 Existing coverage inspected (reuse unless stated; actual names at `f32d049b`)

`core/internal/tenancy/integration_test.go` `TestTenancyModel` (IT-OWN-1..6 creator atomicity; membership
grant/change/remove primitives; the `invitation` source oracle at `:674,919` asserting it confers nothing — kept);
`access/integration_test.go` `TestAccessPolicy` (two-Organization `itWorld`; IT-CTX-6 Group-role-only subject;
IT-MATRIX-1 cell count); `access/matrix_test.go` `TestOracleMatrix` (capability count/cells); `access/policy_test.go`;
`access/rls_expected_test.go` + `rls_integration_test.go` `TestRLSIsolation` (oracle r3 rows incl. RLS-SCOPE-INSERT,
RLS-IMMUT mutation negatives — the `role_bindings` UPDATE/DELETE denials that 00015 turns into policy-gated writes);
`access/ordering_integration_test.go` `TestAccessOrdering` (two-connection step harness, orderings A/B/C, O-13 trusted
unlink, O-ONB-D/E); `access/attack_expected_test.go` + `attack_integration_test.go` `TestIsolationAttacks` (C2
retained context after membership removal — the readmission/dormant-grant shape); `access/audit_integration_test.go`
`TestAudit` (AT-FMT/AT-ATOM/AT-INV: trigger attribution, same-statement capture, catalog from real calls);
`access/footprint_integration_test.go` `TestAccessFootprint` (E-PG rows); `access/recovery_integration_test.go`
`TestRecoveryPost` (six restriction kinds through the model primitives and `DisableUser`; replay);
`core/internal/migrate/{integration_test.go (TestMigrationLifecycle, catalog inventories), rls_integration_test.go
(TestRLSCatalog CT-RLS-2/3/5: policy inventory/ACLs/runtime denials), audit_integration_test.go (TestAuditCatalog),
tenancy_integration_test.go (:411 table-name regex), onboarding_integration_test.go (TestOnboardingCatalog, CT-ONB-1)}`;
`core/internal/audit/audit_test.go` `TestCatalogParity`; identity `TestDisableUser*` (immediate revocation; reused
as is); `onboarding/api/http_test.go` `TestOnboardingHTTP`/`TestOnboardingBodyBudgets`/`TestOnboardingCopyIsExact` (the
adapter test model); web `mcp-connections.test.mjs`, `session.test.mjs`; browser `mcp.spec.ts` M-A11Y,
`shell.spec.ts` S-LANDMARKS/S-NAV. **Not reused as oracles:** `TestRLSPool`, `TestRLSFootprint`, RLS-ATTEST/RLS-AUTH
(no changed input: no shared reader body changes — slice 1's grep confirms and the Result records it).

### 6.2 Scenario groups → tests (reuse / extend / new; the distinct failure each new case catches)

| Group | Case | Kind | Distinct failure caught |
|---|---|---|---|
| (1) bootstrap / invariant | IT-OWN-1..6, G018 ONB-1..3 (creator-owner trigger) | **reuse** | creation atomicity (unchanged) |
| | **MEM-1** `CreateOrganizationWithOwner`: tenant + `org_owner` in one transaction; hookDB fault between the two statements → neither row; `CreateOrganization` still ownerless (documented; the test pins that it is **not** called by any production path — importer/call grep) | **new** (`TestTenancyModel` extension, privileged pool) | ownerless intermediate commit (Spec R2) |
| | **MEM-2** two owners valid (promote a second; both administer; demote one → still one owner); one role per subject/scope (`membership_exists` on a second binding; consume of an already-member invitation does not add a row) | **new** (`TestMembership`) | cardinality model |
| | **CT-MEM-1** `TestMembershipCatalog` (`migrate`): preflight refuses on an ownerless Organization and on an ownerless Group (synthetic rows inserted at version 14 as superuser; counts only in the message), passes once resolved; table/columns/CHECKs/indexes/triggers/policies/ACL strings exact (incl. runtime `invitations` SELECT **without** `token_digest`, UPDATE `revoked_at` only, no DELETE; runtime `role_bindings` SELECT with `id`, no new INSERT shape; the ar/pr grants and lock-only policies of D-14); DEFINER functions' owner/search_path/volatility and the allowance of 6, with the inventory of every ar-owned function (exactly one write body); trigger functions VOLATILE with the lock command separate from the count command; `events_attribute` body = 00013 body + the invitation CASE (hash-diff); no context parser added; `has_function_privilege` for `is_organization_member` (runtime and owner); the `audit.capture_*` execution-mode branch recorded; no Go statement names `token_digest` outside the INSERT and the claim call (`TestSQLBoundary` inventory grep); down refused transactionally with rows/audit history present, then down/up on an empty history (policy/grant/body restoration) | **new** | migration shape and least privilege |
| | **MEM-4** fixed-actor runtime SQL negatives (FORCE-bound owner too): UPDATE a last owner to admin → `last_owner`; DELETE a last owner → `last_owner`; UPDATE an owner row as `org_admin` → `owner_authority`; UPDATE a `group_owner` row as `org_owner` without that Group role → `owner_authority`; **UPDATE a member (not admin) row to owner as the owner → `recipient_ineligible`; promote a disabled admin → `recipient_ineligible`; promote a Group admin without an Organization binding → `recipient_ineligible`** (C2); DELETE an Organization binding whose subject holds Group bindings → `dependent_bindings`; **DELETE own Organization binding from a selected-Group context whose sibling Group binding is hidden → `context_required`** (C3); runtime INSERT of a binding (any shape, depth 0) → 42501; **the C1 direct-statement set: as a fixed nonmember/member context, `UPDATE app.invitations SET consumed_at/consumed_by_user_id/intended_user_id` → 42501 (no column privilege); `SELECT token_digest` → 42501; `invitation_claim` with a guessed/wrong digest for a known pending Organization invitation → `not_found`, zero writes; with the right digest but for another user's Group invitation → `not_found`; `set_config('app.invitation_id', …)` has no effect (no such parser); as the issuer with the token, claiming an invitation that would grant an owner role is impossible (CHECK at issue)**; INSERT/UPDATE `invitations` with an owner role → CHECK; UPDATE `expires_at`/`token_digest`/`role` → immutability; revoke a consumed row / claim a revoked row → immutability/`not_found`; **direct INSERT beyond 100 pending → `invitation_limit`; INSERT with a whole-address-lowercased spelling of a mixed-case local part is accepted as a distinct address (no CHECK folds it)**; a `rls_owner_count` call for an invisible sibling Group → 0 (C1); the foreign-Organization context sees nothing; **and** the same binding statements as superuser succeed (the documented restricted repair path), with the guard's early return pinned | **new** (`TestMembership` runtime SQL + `rls_expected` r4 rows) | bypass of the invariant or the recipient boundary by a direct statement |
| (2) permission / succession / concurrency | IT-CTX-6, `TestOracleMatrix`, `TestAccessPolicy` matrix/selected-context cases | **reuse/extend** (+4 capabilities, 19; cells recomputed; Group-owner-only `group.manage_owners` row) | matrix semantics |
| | **MEM-5** legitimate succession: Organization owner promotes an admin; Group owner promotes a Group admin who is an Organization member; owner demotes self with another owner present; transfer (recipient up, actor → admin / member / none) in one transaction; T8/T9 with actor; `OwnerCount` and `can_*` hints match before/after | **new** | guarded positive paths |
| | **MEM-6** denials, zero writes, `access.denied` where recordable: admin promotes self; admin changes/removes an owner (final role admin/member/absent); admin issues an owner-role invitation (`invalid_input`); Organization owner/admin promotes/demotes/transfers a `group_owner` (no override); Group admin promotes; recipient disabled / foreign Organization / Group-role-only (no Organization binding) / stale (removed between list and action) / not an admin; selected-Group context targeting another Group; revoked session; a subject relabelled "member" by `ChangeMemberRole` when it is an owner → `forbidden` | **new** | finite authority (Spec R1) and stale-choice safety |
| | **MEM-7** audit fault (`hookDB` refusing the trigger's `audit.events` insert) during transfer → full rollback, original owner intact; fault after the recipient's promotion statement → rollback | **new** (`TestAudit` AT-ATOM pattern) | partial succession |
| | **O-MEM-1** two last owners (A, B) each leaving/demoting concurrently at the lock seams: both orders → exactly one succeeds, the other `last_owner`; one owner remains | **new** (`TestAccessOrdering`) | count-then-write race (Spec R2) |
| | **O-MEM-2** transfer racing the recipient's removal/disable (and the actor's session revocation): revocation committing before the Organization lock → transfer denied; waiting behind it → not reported as winning | **new** | transfer-vs-revocation ordering (AC3) |
| | **O-MEM-3** (a) facade promotion racing a facade ownership change on another Group of the same actor (shared order, no 40P01); (b) **a direct runtime `UPDATE role_bindings` racing a facade demotion on the same scope at the binding/scope seam: either serializes or one side aborts with 40P01 — both ends assert zero partial writes and the facade maps the abort to `conflict` (`retry`)** (C3); (c) promotion racing `DisableUser` of the recipient: the disable commits first → `recipient_ineligible`; the promotion commits first → the owner is then disabled with ownership retained (C2) | **new** (O-ONB-D pattern) | lock-order cycle handling; stale eligibility |
| (3) invitation / recipient / atomicity | `TestServiceRejectsBeforeSQL`, `TestTenancyModel` invitation-source oracle | **reuse/extend** | validation before SQL |
| | **MEM-8** issue (Organization address; Group known member) with default/explicit lifetime; `invitation_limit` at 100 **pending unexpired** (an expired 101st does not count; **two concurrent issues at 99 pending → exactly one `invitation_limit`**, C5); min/max lifetime; address normalization table (unit: trim, exactly one `@`, local part byte-exact incl. case and Unicode, domain NFC + lowercase, 64/254-byte bounds, whitespace/control refusals; C4); revoke; list/mine/candidates paging and derived status with **one 101-row boundary fixture** (page 1 = 100 + `next`, page 2 = 1, no `next`; C5); recipient view omits the address; no digest in any answer, log or audit row (canary grep of the transaction log and `audit.events`) | **new** | lifecycle bounds, paging and token secrecy |
| | **MEM-9** Group acceptance by the intended member: binding created (T7 with actor = recipient); `upgraded` for a lower existing role; `already_member` for equal/higher (no row, invitation consumed); `consumed_binding_id` set; second accept with the same token → `already_member` only while the binding exists; **after the member is removed, the same token → `consumed`, no binding, nothing restored** (C4) | **new** | atomic consumption and truthful replay |
| | **MEM-10** Organization consumer with synthetic `VerifiedRecipient` (evidence built by the test through `AttestAddress`): a first-time user (fixture maps a new principal, then consumes) and an existing user of Organization B joining A; both `joined`; the claim sets `intended_user_id`; IDENTITY shape, `ctx_org()` NULL, no session; audit `invitation.consumed` + `membership.granted` in one transaction; a denial on this path records `access.denied` with NULL Organization (Q7) | **new** | the G058 contract (AC5) |
| | **MEM-11** negatives → `not_found`/`conflict`, zero rows: expired; revoked; already consumed by another user; wrong token; token of another invitation; Group token presented to the Organization consumer and vice versa; address mismatch (a local-part case variant is a **different** address; a domain case variant matches after normalization); zero-value evidence / JSON-constructed `VerifiedRecipient` → refused (unit); disabled recipient → `recipient_ineligible`; Group invitation to a user whose Organization binding was removed after issue → `recipient_ineligible`; **issuer demoted/removed after issue → `issuer_authority_lost`, invitation still pending (no recipient-side revocation), the administrators' revoke then works**; tampered role impossible (intent copied from the row); `AttestAddress` call-site pin (unit) | **new** | foreign/escalating/replay acceptance (AC5) |
| | **MEM-12** concurrent duplicate acceptance (two connections, same token): one `joined`, one `already_member`, one binding; acceptance racing the recipient's Organization removal (Group invitation): ordering decides, never a dormant Group binding (the DB `invitation_recipient` check) | **new** (`TestAccessOrdering`) | duplicate effect / dormant grant |
| (4) departure / suspension / recovery | `TestRecoveryPost` six kinds, identity `DisableUser` tests | **reuse** | revocation primitives |
| | **MEM-13** Organization departure: owner of Group g1 and member of g2 leaves → `last_owner` for g1 (zero writes); after transferring g1 → one transaction removes g2 binding, g1 binding (now admin) and the Organization binding; readmission by a fresh invitation → new Organization binding, **no** Group rows; attack C2 shape extended (retained selected-Group context after departure sees nothing) | **new** | partial removal / dormant grants (AC3) |
| | **MEM-14** retirement: a user owning scopes in two Organizations (sole owner in one) → `RetirementEligibility` lists exactly that scope id (two-Organization fixture) and `GET /internal/v1/account/retirement-eligibility` answers it with the session credential only (no tenant header; 401 without; ids only); after succession → `RetireUser` departs both Organizations (ascending) and disables; **a blocker injected between the preflight and the second Organization's departure → the call stops with `Departed=[A]`, `Remaining=[B]`, user still enabled; re-run completes** (C3); a tenant administrator has no retire route/capability (HTTP 404 on any attempt; facade method absent from `TenantContext` paths); IDENTITY-shape reader returns nothing without `ctx_user()` | **new** | cross-tenant retirement privacy and honest partial progress (AC4) |
| | **MEM-15** (a) `DisableUser` on a last owner: immediate (sessions revoked, guard step denies), binding retained, `OwnerCount` still 1, audit `user.disabled` with the revocation exception path untouched; (b) restricted recovery exercised synthetically: superuser inserts a replacement owner binding with a ledger entry; replay of `membership.downgrade`/`remove` restrictions after a restore keeps owner cardinality ≥ 1 and does not revive a removed binding (`TestRecoveryPost` extension, **no drill rerun**) | **extend** | suspension delayed by cardinality / restore reviving access (AC4, AC7) |
| (5) changed UX / delivery | ONB-H pattern | **reuse** | adapter composition |
| | **MEM-H1..H15** `TestMembershipHTTP` (one per pair, 15): no session 401, foreign/`http` origin 403, parameter rules 400, non-JSON/trailing/unknown fields 400, other methods 405, sub-paths 404, `mine`/`candidates` literal-before-template routing, budgets (8 KiB + 1 → 400), `TestMembershipCopyIsExact`, refusal mapping (`last_owner`/`recipient_ineligible`/`invitation_limit`/`issuer_authority_lost`/`context` fields), the token only in the 201 issue answer, `no-store` on every answer, the eligibility route with the session credential and no tenant parameters; the authenticated cases are covered at the facade and in the browser slice (G018's disclosed pattern) | **new** | closed surface |
| | unit `membership-client.test.mjs` (15 methods; shapes, `successBytes`, `malformed`), `membership-forms.test.mjs` (closed `action` set; status pages never claim a change on conflict; **`invite` answers 200 `no-store` and never a 303; the token never appears in a `Location`, URL or completion record**; address field trimmed only — normalization is server-side; the Account notice renders count only) | **new** | BFF honesty |
| | **E-MEM-1** seeded Organization owner invites by address → sees the link + token once → revokes a pending invitation; **E-MEM-2** seeded Group owner invites a seeded co-member to Research → that member accepts via `/invitations/<id>#<token>` → appears in the Group members list; second visit → "already a member"; **E-MEM-3** owner promotes an admin, demotes, transfers; **E-MEM-4** last-owner leave refused with the truthful notice; after promotion, leave succeeds; **E-MEM-5** admin sees no owner controls; an Organization owner sees no Group-owner controls on a Group they do not own; stale action (member removed in another tab) → status page, nothing changed; **E-MEM-6** keyboard-only run of E-MEM-2 + axe at 1280/320 (M-A11Y), Light/Dark projects; **E-MEM-7** no MCP connection context change across navigation (E-ONB-7 pattern) | **new** `membership.spec.ts`; S-LANDMARKS/S-NAV **extend** | end-to-end UX/accessibility/security |
| | `TestRLSIsolation` oracle r4: `role_bindings_update`/`_delete` admitted/denied rows per context; runtime has **no** admitted INSERT shape on `role_bindings` and no consumption column on `invitations`; `invitations_*` rows (admin, recipient, nonmember IDENTITY context sees nothing, foreign, no-context); RLS-IMMUT rows that become admitted writes are **moved**, not deleted (the oracle must still deny the unchanged shapes) | **extend** `rls_expected_test.go` | database-level least privilege independent of the facade |
| | `TestIsolationAttacks` r3: ATK rows for direct binding writes, invitation tampering, and **consumer forgery without the token** (a fixed nonmember/member context naming a known pending Organization invitation cannot read its digest, cannot update its consumption columns, and the claim without the digest writes nothing — MEM-4's C1 set as attack rows) | **extend** | attack-suite coverage |
| | AT-INV (dynamic) and `TestCatalogParity` (33/6/58/52) | **extend** pins | audit invariants |
| | `TestAccessFootprint` E-PG rows for promote/transfer/remove-with-dependents/accept/consume/list | **extend** | fan-out/lock regression |

### 6.3 Batches — focused executable selectors; no default full local profile

Selectors are entry points; zero selected tests is a batch failure. Working directory `core/`, module `core`, unless
stated. Before every batch: fetch/rebase/reconcile per the solo-operator policy; on Docker Desktop/WSL run
`bash infra/ci/docker-desktop-mounts.sh` first; no edits during a running batch.

- **b1** (unit, no database): `go test ./internal/tenancy/... ./internal/audit/... ./internal/membership/...
  ./internal/onboarding/... -run '^(TestServiceRejectsBeforeSQL|TestPackageBoundary|TestSQLBoundary|TestAccessBoundary|
  TestProductionImporters|TestRLSBoundary|TestOracleTables|TestOracleMatrix|TestOracleManageable|TestOracleGates|
  TestPolicyGates|TestPolicyScope|TestCatalogParity|TestMigrationParity|TestNormalizeAddress|TestInvitationToken|
  TestVerifiedRecipientConstruction|TestMembershipPageBounds|TestMembershipHTTP|TestMembershipRefusalMapping|
  TestMembershipBodyBudgets|TestMembershipCopyIsExact|TestOnboardingCopyIsExact)$'`; `go test ./migrations/ -run
  '^TestMigrationSet$'`; `go test ./internal/migrate/ -run '^TestNormalSetIsTheReviewedSet$'`; logrecord pin from
  `infra/validation/tools/` (module `validationtools`): `go test ./internal/logrecord/ -run
  '^(TestCatalogueMatchesTheBrief|TestClassifyAccepts|TestClassifyRejects|TestValidate|TestCanaries)$'`.
- **b2** (runtime PostgreSQL, `rls_go`, harness rows **B-MEMBERSHIP-CATALOG** and **B-MEMBERSHIP-IT** in
  `membership_checks()` on the `onboarding_checks` template with `it_count` pins): `./internal/migrate/
  '^TestMembershipCatalog$'`; `./internal/tenancy/access/ '^TestMembership$'`. Affected existing rows with pinned
  subcases, each with its changed input:
  - B-RLS-CATALOG `./internal/migrate/ -run '^TestRLSCatalog$/^(CT-RLS-2|CT-RLS-3|CT-RLS-5)'` (policy inventory,
    ACLs/functions, runtime denials change); CT-RLS-1/4/6 and MG-RLS-* unchanged.
  - B-RLS-IT `./internal/tenancy/access/ -run '^TestRLSIsolation$/^(RLS-SCOPE|RLS-CTX|RLS-ORG|RLS-IMMUT|RLS-PROJ)'`.
    **Wider covering batch (bounded, supervisor disposition):** the whole `'^TestRLSIsolation$'` once only if slice 1
    changes `rls_org_role()`/`rls_group_role()` bodies or `role_bindings_select` (not planned).
  - B-RLS-ORDERING `'^TestRLSOrdering$'` only if a lock statement text changes (the new Subject lock statement is
    new, not changed; the Result records the decision).
  - access rows: `-run '^TestAccessOrdering$/^(O-MEM|O-ONB|O-13|O-14|E-PG-3)'`; `-run '^TestAccessFootprint$/^(E-PG-1|
    E-PG-2|E-PG-MEM)'`; `-run '^TestIsolationAttacks$/^(ISO-ATTEST|B|C|D|ISO-CITATIONS)'`; `-run '^TestAccessPolicy$/^(
    IT-CTX-6|IT-MATRIX-1|IT-FACADE-)'`; `-run '^TestAudit$/^(AT-FMT|AT-ATOM|AT-INV)'`; `-run '^TestRecoveryPost$'` (MEM-15b
    extension; the drill B-RECOVERY-DRILL is **not** rerun locally — hosted `migrations` suite after conditional merge).
  - version/count pins (**subcase-selected, C6**): the first slice-1 run lists, with `-v`, the subcases of
    `TestTenancyModel`, `TestBusinessProfileModel`, `TestMigrationLifecycle`, `TestAuditCatalog`, `TestOnboardingCatalog`,
    `TestTenancyCatalog` and `TestRecoverySeed` whose inputs changed (the `current_version` pin, catalog counts, the
    `:411` table-name regex, the seed ownership rows); only those subcases are selected thereafter by name
    (recorded in the Result) — a version or count pin alone does not justify the whole model/migration/recovery
    suite. One whole entry point is technically necessary and stated once: `./internal/mcp/ '^TestMCPTransport$'`,
    whose shared fixture asserts the version at setup and has no sub-selectable subcase. Unchanged-input rows
    (B-OBJ-*, B-AUDIT-*, B-RETENTION-*, B-MCP-CATALOG, B-DOMAINS-*, B-BRANDING-*, B-ONBOARDING-* unless 00014 objects'
    postconditions are restated) are not rerun. No full local gate; no CI meta-test campaign.
- **b3** (web unit): `npm run test:unit` (whole suite, seconds).
- **b4** (browser, identity mode): `IDENTITY_E2E=1 npx playwright test --project=membership-light
  --project=membership-dark --project=shell-light --project=shell-dark --grep 'E-MEM|S-LANDMARKS|S-NAV|S-A11Y'`;
  `BROWSER_MATRIX_IDENTITY_E2E` extended; B-TENANCY-HTTP and B-WEB-COMPONENTS via their harness rows.
- **Required final local coverage before the MR:** the union of affected checks only — B-UNIT-CORE, B-UNIT-WEB,
  B-MEMBERSHIP-CATALOG/IT, B-RLS-CATALOG, B-RLS-IT, the access rows above, B-MCP-IT's fixture entry point,
  B-TENANCY-HTTP, B-WEB-COMPONENTS, the browser rows for the membership/shell projects, B-SCOPE/B-REQUIRED-CHECKS —
  each with changed-input provenance in the Result. **No default `full` profile.** Hosted evidence: MR fast pipeline
  automatically; the path-based staging gate (foundation rows; `migrations` suite for `core/migrations/`) after an
  authorized conditional merge. CI meta-tests (`validate-controls.sh`) only for the one `F-widen-*` row added with the
  profile-paths change (a mapping edit; no control-logic change).

### 6.4 Manual review (M1), readiness and the draft MR

Same preview environment as G017/G018 §6.4 (L2 emulator overlay, `IDENTITY_PROXY_PUBLISH=127.0.0.1:543`, seeded
users, run CA in a dedicated browser profile; no WorkOS, no hosts-file edit). Readiness: candidate SHA, seed digest,
CA path, `browser/users.json` pointer; the operator confirms readiness before any timer. Feedback items: invite by
address and see the one-time token; invite a co-member to a Group and accept from the second seeded user (link with
fragment, and paste fallback); revoke; promote/demote/transfer on Organization and Group; last-owner leave refusal
wording; stale-action status page; admin view without owner controls; Account retirement notice; Light/Dark/System
at 1280 and 320; keyboard-only acceptance; error association on an invalid address. Findings → fixes → rerun affected
E-MEM/S-*/unit selectors → repeat M1 for changed surfaces only; cleanup `identity:stop` (`down -v`) with the
owned-resource receipt. **Draft MR opens when slice 2 compiles and b1/b2 pass**; M1 completes before final local
qualification, Codex pre-integration acceptance and any merge. No CI-produced artifact is a prerequisite for M1.

## 7. Delegation

One implementer owns core and web (overlapping edits; one write owner; the G018 successor-checkpoint pattern with
precise remaining lists is the context-management tool). Two optional bounded read-only reviewers: **R-A**
(security/data: the guard trigger, policies, DEFINER readers, consumer context, token handling, lock order) after
slice 1; **R-B** (UX/accessibility/privacy of slices 3–4) after slice 4. No background ownership-repair agent, no
duplicate session (Goal). Delegated instructions carry the task-prefixed worktree rule (generated `agent-<id>`
accepted; record task/role/path) and the read-only boundary; Claude validates and integrates all work and keeps the
one B7 ledger. Model/effort per owner direction: implementation on Opus high; reviewers at the coordinator's stated
effort; actual values reported as unknown where unverifiable.

## 8. Sequence

0. R0 after G018 integration (§0.3); rebind §2; Codex Plan ACCEPT (incl. §10/§12 dispositions); allow-list import;
   initial Result; execution release.
1. **Slice 1 (riskiest; core, no HTTP):** 00015 + CT-MEM-1 (preflight, shape, down refusal/up); capabilities 19,
   matrix, `rls_expected` r4; `tenancy/membership.go` + `invitations.go` + `fixture.go` + `facts.go` (Subject/
   OwnerCount); guarded ownership/membership/invitation operations, the claim function and the consumer contract
   (`AttestAddress`/`VerifiedRecipient`); `CreateOrganizationWithOwner` + `TestTenancyFixturePrimitives`;
   MEM-1..15, O-MEM-1..3, MEM-15b; version-15 pins; audit catalog; the §6.3 subcase selection from the first run.
   b1, b2. R-A.
2. **Slice 2:** lists (`ListMembers` + filter, `ListInvitations`, `ListMyInvitations`, `ListInvitationCandidates`),
   `RetirementEligibility`/`RetireUser`; `membership/api` (15 pairs) + `http_test`; server registration; logrecord
   enums; importer/route/B-TENANCY-HTTP pins; `membership_checks()` rows; footprint rows. b1, b2. **Open the draft
   MR.**
3. **Slice 3 (after G018's web slice is integrated, §0.3):** `membership-client.ts`, `ui/{select,table}`,
   components, pages/routes, Account notice, shell/Group links; unit tests; `WEB_ROUTES`, `UI_FILES`, `profile-paths`,
   B-WEB-COMPONENTS. b1, b3.
4. **Slice 4:** seed owners/admins (reconciled with G018's seed); `membership.spec.ts`, projects, matrix pins;
   S-LANDMARKS/S-NAV. b4. R-B.
5. **Slice 5:** docs; M1; reconcile; required final local coverage (affected only); Result (incl. the explicit
   "G058 public proof/admission pending" statement, AC5/AC8); safekeeping push; MR to ready after Codex
   pre-integration acceptance; checkpoint review after attempt 4 if reached. Merge and final ACCEPT wait for G018's
   final ACCEPT.

## 9. Acceptance-criteria mapping

| AC | Evidence |
|---|---|
| AC1 | G018 creator path retained (ONB-1..3 rerun unchanged); MEM-1 (atomic bootstrap + fault); MEM-2 (two owners; one role per subject); CT-MEM-1 preflight (explicit legacy refusal, counts only); MEM-4 runtime/FORCE-owner negatives; FK/cascade inventory (all RESTRICT; recorded in TENANCY) and the dependent-grant guard; MEM-7 rollback |
| AC2 | MEM-5 (promotion/transfer positives), MEM-6 (admin self-promotion, owner removal by admin, inherited Organization-owner override of Group ownership, disabled/foreign/stale recipients), MEM-7 (audit fault preserves ownership), matrix row pins (19), AUTHORIZATION rows |
| AC3 | MEM-6 single last-owner refusal; O-MEM-1 competing departures; O-MEM-2 transfer vs. revocation; MEM-13 Organization departure with dependent Groups (refusal before partial writes; atomic removal after succession; readmission revives nothing); attack C2 extension |
| AC4 | MEM-14 (eligibility across Organizations, private; tenant admin cannot retire; no route), MEM-15a (last-owner disable immediate, ownership retained), MEM-15b (restricted recovery and restriction replay synthetic), RECOVERY/SECURITY documentation |
| AC5 | MEM-8..12 (intended canonical/verified subject, scope/role/issuer checks, expiry/revoke/replay/foreign/escalation negatives, acceptance/removal races, atomic consumption/audit, truthful lost-response status); MEM-10 first-time and existing-user consumer with synthetic evidence; §5.5 contract; Result states G058 public proof/admission **pending** |
| AC6 | E-MEM-1..7 (existing-member Group acceptance, role/owner controls, eligible-admin paging, last-owner/stale feedback, privacy, CSRF, keyboard, appearance, reflow), unit BFF tests, M1 record; no global directory (member picker is scope-bounded, ≤ 100) |
| AC7 | CT-MEM-1 (migration/grants/RLS/helper trust/lock-relevant policies, FK inventory), `TestRLSIsolation` r4, `TestIsolationAttacks` r3, `TestCatalogParity`, MEM-15b restore/replay, E-PG rows → PG-009; no mirrored roles (no `source='invitation'`, D-3) |
| AC8 | Result with source/inputs/commands/outcomes/deviations (Git-first), M1 findings, MR/staging disposition, cleanup receipt; Codex final ACCEPT after G018's; the public first-time journey explicitly not claimed |

## 10. Decisions, risks and escalations (flagged for explicit review)

| ID | Item | Recommendation / status |
|---|---|---|
| **D-1** | Last-owner enforcement by a **BEFORE trigger with a scope-row lock + DEFINER count reader** (`rls_owner_count`, pr-owned, helper policy on owner rows of `ctx_org()`), not a unique index, deferred constraint trigger, owner column or Go-only check. G018 D-9 rejected a count reader because it was subject-filtered and could not serialize; here the reader is not subject-filtered and serialization comes from the lock the trigger takes before counting. A deferred constraint trigger was rejected: it would evaluate at COMMIT after the Go guard's own facts, making `conflict` reporting late and the FORCE-owner path harder to test; a `COUNT`-maintaining owner column was rejected as a mirrored store | security/data review item |
| **D-2** | Cross-Organization helper policy `role_bindings_retirement_reader` (pr; owner rows, no `ctx_org()` term) behind `rls_sole_owner_scopes()`, which returns only `ctx_user()`'s own sole-owned scope identifiers in IDENTITY shape. Alternative rejected: per-Organization loops from Go in tenant contexts (needs an Organization context the subject may no longer be able to resolve, and exposes nothing more). The reader returns no names, no co-members, nothing when `ctx_user()` is NULL | tenant-visibility review item (Spec R2 "do not broaden tenant visibility") |
| **D-3** | **No `source='invitation'`**: all bindings keep `source='membership'`; invitation provenance lives in `invitations.consumed_binding_id` and the audit trail. Widening `role_bindings_scope_consistent` and every `source='membership'` authority predicate (gate, `role_bindings_select`, `events_attribute`) would be a broad authority change for a label; TENANCY's "later source" note is updated to say so | acceptable in principle (proposed) |
| **D-4** (r1) | Consumption and the binding write happen inside **one `<db>_auth_reader`-owned SECURITY DEFINER function** `app.invitation_claim` (§4.3/§4.4) instead of r0's AFTER UPDATE trigger + depth-admitted runtime INSERT. Why: C1 showed that any runtime-admitted consumption shape keyed on context settings is forgeable by a fixed context; the function keys the transition on token possession (the digest is unreadable by runtime) and binds the subject to `ctx_user()`, and ar's write privileges exist only inside that body (CT-MEM-1 inventories ar's functions). The G018 reservation against DEFINER *mutation* functions is answered by scope: one function, one bounded body, policies `TO ar` that admit only the subject's own non-owner row, no DELETE | security review item |
| **D-5** (r1) | **Withdrawn**: no `app.ctx_invitation()`; a nonmember consumer runs in IDENTITY shape with `ctx_org()` NULL and reaches only the claim function. Alternatives rejected: a new database role/pool for the consumer (a new credential and infrastructure change for the same boundary), making pending Organization invitations visible to any context subject (enumeration), a nonmember session (Spec R4 forbids) | security review item |
| **D-6** | `access.denied` for a nonmember consumer denial carries **NULL Organization** (the depth-1 audit scope rule refuses a non-member's Organization). The tenant's `audit.read` therefore does not list failed consumption attempts; the log event and the invitation's unchanged state remain. **Accepted by the review (Q7)**: observable sanitized failure, no audit-attribution widening | accepted |
| **D-7** | Organization invitations are **address-addressed only** (no `intended_user_id` at issue); Group invitations are **user-addressed only** (candidate picker). Rationale: no global directory exists for an issuer to name a foreign user; a Group recipient must already be a member. A known user invited to another Organization is consumed through the same address-matched contract by G058 | acceptable in principle (proposed) |
| **D-8** | Organization owners/admins hold **no** `group.manage_owners` (no override). Consequence: an Organization whose Group has an unavailable sole owner needs the restricted recovery procedure, not a tenant-admin action. **Settled by Spec R1/§53.7 (Q2); not re-asked** | settled |
| **D-9** | Ordinary removal of a subject who is a Group owner **refuses** (`owner_subject`) rather than cascading; explicit Group succession/removal first | per Spec R3; acceptable in principle |
| **D-10** | `RetireUser` is per-Organization transactional (not one global transaction) because `InitContext` admits one context per transaction; eligibility is global and atomic refusal is guaranteed before any write; partial progress is recorded and resumable | acceptable in principle (proposed) |
| **D-11** (r1) | Token delivery link carries the token in the URL **fragment** (`#`) with a paste fallback; the only server-visible token transports are the issue response (the `invite` action's direct 200 `no-store` page — no redirect, since a redirect cannot carry it) and the acceptance form body / JSON body; a lost issue response is unrecoverable by design (revoke and reissue) | acceptable in principle (proposed) |
| **D-12** | `events_attribute` body replacement (00013 precedent) limited to the invitation shape CASE; hash-diffed in CT-MEM-1 | review item (trigger function body change) |
| **D-13** (r1) | Group recipient picker source: `app.rls_group_invitation_candidates` discloses **ids and display names of current enabled Organization members not in the Group** to that Group's owner/admin (and to `MGR`). Rationale: `role_bindings_select` hides the Organization member list from a Group-only manager, yet Spec R4 makes Group invitations user-addressed, so some authorized candidate source must exist; this is the least-disclosing one (no roles, no other Groups, no addresses, keyset-bounded, current scope only, no cross-tenant directory). Alternative rejected: address-typed Group invitations (users hold no address; would need a new identity column) | tenant-visibility review item |
| **D-14** (r1) | Narrow privilege additions for `<db>_auth_reader` (lock-only `FOR UPDATE … WITH CHECK (false)` policies on `users` and `role_bindings`, INSERT/UPDATE(role) on `role_bindings` admitted only for the subject's own non-owner row, UPDATE of the four consumption columns on `invitations`) and for `<db>_policy_reader` (`role_bindings_scope_reader`, `users_policy_reader`); runtime SELECT on `role_bindings` gains `id`; EXECUTE on `is_organization_member` extended to the owner role. Every grant/policy string is pinned in CT-MEM-1's ACL postconditions | security/data review item |

**Escalation risks (stop, do not design around):** material G018 drift at R0 (§0.3); any ownerless legacy record in
a persisted environment (preflight refusal → return the counts, never promote/delete); a need to widen
`role_bindings_select`, `rls_org_role()`/`rls_group_role()` bodies or the depth-1 audit scope rule; any owner-role
grant path outside `PromoteToOwner`/`TransferOwnership`/bootstrap; a provider that cannot supply a verified address
claim (G058's problem, but if G019's synthetic contract proves unimplementable it returns here); a 42P17 policy cycle
or a facade-vs-facade deadlock in b2 (facade-vs-direct 40P01 is expected and handled, §1.5); any need for a nonmember
session, applicant state, email vendor, public endpoint, new database role or pool; a need to touch
`infra/recovery/**` or `core/internal/identity/**` beyond reuse (Q5: none identified by the corrected design; an
exact demonstrated need returns as a handoff); a second write body under `<db>_auth_reader`; budget exhaustion or
context overrun (handoff at ~250k with the precise remaining list).

## 11. Budget, forecast, pins and scope

### 11.1 Budget — B7, checkpoint after attempt 4 (Goal); recommendation: keep B7; forecast 3–4 attempts; unspent.

**Context/split (Q6 settled).** One Goal with bounded slices, B7 and the checkpoint after attempt 4; the G018
successor-checkpoint pattern (a fresh worker per ~250k context, precise remaining lists in the Result) is the
context-management tool. The after-slice-2 seam noted in r0 is retained only as the natural handoff boundary, not as
a split proposal; no new Goal or duplicate planner. No scope is enlarged.

### 11.2 Count pins (provisional at `46f1dc5e`; re-pin at R0 against the integrated G018)

Migrations 14 → **15** (all P-4 sites); capabilities 15 → **19**; matrix primary cells 330 → computed and recorded
(+4 capabilities × the oracle's formula); `access.denied` operations 36 → **52**; trigger actions 30 → **33**, core
actions 6, reasons **58**, resource types 9 → **10**; policies: `role_bindings` −1 (`_lock`) +2 runtime/owner
(`_update`, `_delete`) +2 ar (`_insert_invited`, `_auth_upgrade`) +2 pr helper (`_scope_reader`,
`_retirement_reader`); `users` +2 (`users_policy_reader` pr, `users_auth_lock` ar); `invitations` +3 runtime/owner
+2 ar; SECURITY DEFINER functions +6 (00015 allowance = **6**: four pr STABLE readers, `rls_user_enabled_lock` ar
VOLATILE, `invitation_claim` ar VOLATILE); SECURITY INVOKER trigger functions +3 (`invitations_keys_immutable`,
`invitations_bound`, `role_bindings_authority_guard`; all VOLATILE plpgsql) and the replaced `events_attribute`
body; context parsers **unchanged**; runtime grants: `role_bindings` +DELETE, SELECT columns 7 → **8** (`id`), INSERT
unchanged; `invitations` SELECT **15** (no `token_digest`) / INSERT 11 / UPDATE **1** (`revoked_at`), no DELETE; ar
grants: `invitations` SELECT all / UPDATE 4, `role_bindings` INSERT 8 / UPDATE 1 (`role`), `users` UPDATE 1 (`status`,
lock-only); pr grants: `users` SELECT 3; EXECUTE on `is_organization_member` +owner; `audit.capture_*` triggers +1
table (ar INSERT on `audit.events` only if the precondition finds invoker-mode captures); importers of `access` +1
(`membership/api`); `inContextTx` call sites 12 → recorded; exported `access` methods +**20** (15 operations +
`ListMyInvitations`, `ListInvitationCandidates`, `AttestAddress`, `NewVerifiedRecipient`,
`ConsumeOrganizationInvitation` counted once each as listed in §5.1); tenancy production files 15 → **18**
(`membership.go`, `invitations.go`, `fixture.go`); `required-checks.tsv` 123 → **125** (B-MEMBERSHIP-CATALOG,
B-MEMBERSHIP-IT); `completion.schema.json` route enums **+14 paths** (two occurrences); `events.json` +1 event
(`membership.request`), +16 operations, +4 capabilities; HTTP **15 method/path pairs on 14 paths**;
`CORE_MEMBERSHIP_ROUTES` 15 pairs; `WEB_ROUTES` 31 → **38**; `static_web_boundary` allow-list +4 exact entries
(`lib/membership-client.ts`, `app/organization/members/action/route.ts`, `app/groups/[id]/members/action/route.ts`,
`app/invitations/[id]/accept/route.ts`); `UI_FILES` 9 → **11** and `UI_SLOTS` +2 (`select`, `table`);
`THIRD_PARTY_NOTICES.md` +1 sentence; `WorkspaceArea` 2 → **3** (`members`) and the nav-count pins in
`onboarding.spec`/`mcp.spec`/`shell.spec`; Playwright projects +2 (`membership-light/dark`);
`BROWSER_MATRIX_IDENTITY_E2E` +1 line (`E-MEM-1 … E-MEM-7 S-ERRORS`); `profile-paths.tsv` +2 foundation rows
(`web/app/organization/members/`, `web/app/invitations/`); `validate-controls.sh` +1 `F-widen-*` row; seed: +2
`org_owner` (B, C), +3 `group_owner` (design, operations, planning), +1 `org_admin`, +1 `group_admin` (reconciled
with G018's slice-4 seed at R0); `TestAccessOrdering`/`TestAccessPolicy` `it_count` pins recomputed (mechanical
correction under the owner's 2026-10-17 autonomy rule if still stale at R0); `SEMANTIC_COLOR_TOKENS`, fonts,
`web/public`, `appearance-init`, `MAX_FORM_BYTES` unchanged.

### 11.3 Scope allow-list — `management/v0/plans/G019-scope-allowlist.txt` (r1; literal `[.]` dots; Codex imports it as `management/v0/results/G019-scope-allowlist.txt`): **40 executable patterns**, 52 lines with comments; SHA-256 in the hand-back and §13. r1 changes versus r0 (39): `web/THIRD_PARTY_NOTICES[.]md` added (P-10 lesson); `web/tests/browser/(shell|onboarding)` widened to `(shell|onboarding|mcp)` (nav-count pins); `core/internal/tenancy/(membership|invitations)` widened to `(membership|invitations|fixture)` (the relocated fixture primitives); the harness comment rows updated (fourteen paths, sixteen operations, `static_web_boundary` entries, `THIRD_PARTY_NOTICES`). r0 archived as `G019-scope-allowlist-r0.txt`.

Explicitly **not** listed and not needed: `infra/recovery/**` (RECOVERY.md documents the replay caveat; no tooling
change), `core/internal/identity/**` (Q5: the corrected design needs no identity change — the evidence constructor
lives in `access` with a call-site pin; `DisableUser` reused as is), `core/internal/mcp/**` production,
`web/lib/session.ts`, any historical migration, `web/app/groups/create/**` and the other G018 web files except the
two one-line link edits (`groups/[id]/page.tsx`, `app-shell.tsx`), `.gitlab-ci.yml`, any new database role or pool
configuration. Patterns for G018-created files (`web/app/groups/\[id\]/page[.]tsx`, `web/tests/browser/onboarding[.]
spec[.]ts`, `core/internal/migrate/onboarding_integration_test[.]go`) are listed only because G019 edits them after
integration; if G018 lands with different names, R0 reconciles the pattern (mechanical).

## 12. Question dispositions (r1; the review's answers applied — none is re-asked)

1. **Persisted environments (Q1).** R0 performs a scoped readiness inventory of the persisted databases the
   operator names (no unrelated-host scan, no automatic repair, no claim that every environment is clean); an
   actual ownerless Organization/Group found there is escalated for an attributable owner/operator resolution
   through the restricted recovery procedure before 00015 runs on it (§4.7). The seed amendment covers fresh
   environments only.
2. **Organization-owner override of Group ownership (Q2).** Settled by Spec R1/§53.7: no override (D-8). Not re-asked.
3. **Address and proof (Q3).** C1/C4 applied: local part byte-exact, domain NFC + lowercase, no mailbox-equivalence
   assumption, no email-based merge; the WorkOS `email`/`email_verified` fields are documented candidates, G058's
   actual proof flow remains pending and unqualified here; no probe.
4. **Ownerless primitive (Q4).** C2 applied: no supported production escape; `CreateOrganizationWithOwner` is the
   only supported shape, the ownerless/unguarded primitives are fixture-only in `fixture.go` with a boundary-test pin,
   callers inventoried (none in production), seed and recovery replay explicitly isolated as superuser paths.
5. **Identity-file expansion (Q5).** No speculative scope: the corrected design needs no `core/internal/identity/**`
   change (`DisableUser` reused; the evidence constructor lives in `access`; the user-row lock is an ar-owned DEFINER
   in 00015). An exact demonstrated need found during implementation returns as a handoff, not a silent edit.
6. **Split/sequencing (Q6).** One Goal, bounded slices, B7/checkpoint 4, existing overlap/reassessment gates; no new
   Goal or duplicate planner (§11.1).
7. **Nonmember denial audit (Q7).** NULL-Organization `access.denied` accepted within the current audit boundary with
   the observable sanitized failure (D-6); no widening.
8. **Finite defaults (Q8).** 100 pending per scope and 1h–30d (default 7d) accepted as constants, now enforced
   (serialized scope lock + `invitations_bound`, expiry-aligned count) and fully traversable (keyset on every list).

Open for Codex's review (new in r1): **D-13** (candidate-list disclosure to Group managers) and **D-14** (the exact
narrow privilege additions for `<db>_auth_reader`/`<db>_policy_reader`); both are design choices the review asked
r1 to pin, flagged for explicit review rather than owner questions.

## 13. Revision history

| Rev | Date (UTC) | Inputs | Summary |
|---|---|---|---|
| r0 | 2026-10-17 | `origin/staging` `988ed020`; G018 draft MR !41 at `f32d049b`; branch `g019-plan`; worktree `agent-<id>`; model Fable 5.1, effort "high (requested)", actual unknown | initial provisional Plan; scope allow-list r0 (39 patterns). Archived byte-exact as `G019-scoped-invitations-membership-ownership-plan-r0.md` (SHA-256 `5560ca8e…8f55`) and `G019-scope-allowlist-r0.txt` (`4c888ba4…5b7f2`) |
| r1 | 2026-10-17 | r0 `dbbfea2b` + `origin/staging` `9cf9beed` (Codex review r0, SHA-256 `0f2f5788…302d`); G018 observed at `46f1dc5e`; branch `g019-plan` (local `g019-plan-r1`); worktree `agent-<id>`; model Fable 5.1, effort "high (requested)", actual unknown | bounded correction; no implementation, build, test, probe, container or MR; B7 unspent. Item map below |

**C/Q → change map (r1).**

| Item | Change |
|---|---|
| C1 | `ctx_invitation` withdrawn; consumption moved into the ar-owned `app.invitation_claim` keyed on token possession and `ctx_user()`; runtime loses `token_digest` SELECT and every consumption column, gains no INSERT shape; the "why a fixed context cannot claim another invitation or manufacture an upgrade" argument (§1.1); issuer check moved into the claim (the recipient cannot read the issuer's binding), refusal only, no recipient-side revocation; `id` added to the runtime SELECT (the one grant widening); readers answer only for visible scopes; MEM-4 C1 direct-statement set and attack rows (§1.1, §1.4, §4.2–4.5, §6.2, D-4, D-5, D-14) |
| C2 | guard step (2): same-scope admin → owner only, `rls_user_enabled_lock` (FOR SHARE vs `DisableUser` FOR UPDATE), `is_organization_member` for Group promotion; caller inventory (none in production); `CreateOrganizationWithOwner` the only supported shape; fixture-only primitives in `fixture.go` with `TestTenancyFixturePrimitives`; seed/replay isolated and documented; emergency suspension unchanged; MEM-4/O-MEM-3(c) (§1.1, §1.4, §4.5, §6.2) |
| C3 | real lock order from `facts.go` (actor bindings before Groups), facade-vs-direct inversion acknowledged with 40P01 rollback semantics (O-MEM-3(b)), VOLATILE plpgsql with separate lock/count commands, STABLE readers never lock; Organization departure requires `ctx_group() IS NULL` at facade and guard, dependent grants counted by a DEFINER over all Groups; retirement eligibility as preflight with honest resumable progress (`RetirementProgress`, MEM-14); down migration refuses non-empty invitation/audit history transactionally (§1.4, §1.5, §4.5, §4.7, §5.1) |
| C4 | local part byte-exact, domain NFC + lowercase, syntax/bounds CHECK without whole-address `lower()`; `ProviderAddressEvidence`/`AttestAddress` with call-site pin, no request boolean; replay guarded by same token + recipient, `already_member` only with a current binding, else `consumed`; `invite` answers a direct 200 `no-store` page (no redirect can carry the token), copy/paste fallback, lost-response revoke/reissue (§1.1, §1.4, §1.6, §4.2, §5.5, §6.2, D-11) |
| C5 | issue serialized on the scope row (`ExclusiveOrganization`/target Group) plus `invitations_bound`; pending count aligned with expiry and the index; per-operation statement inventory; keyset on `mine`, members (`owner_candidates` filter), invitations and candidates; `rls_group_invitation_candidates` as the least-disclosing Group picker source (D-13); one 101-row boundary fixture (§1.4, §1.5, §1.6, §4.2, §4.3, §5.3, §6.2) |
| C6 | 15 pairs / 14 paths reconciled with the route enums, `CORE_MEMBERSHIP_ROUTES`, 15 client methods; `GET /internal/v1/account/retirement-eligibility` + `retirementEligibility()` for the Account notice (count only, no retire action); G018 state at `46f1dc5e` reconciled (P-9..P-12 rewritten, P-15 lessons, P-16 paging); `static_web_boundary`, `UI_FILES`/`UI_SLOTS`, `WorkspaceArea`, nav pins, `THIRD_PARTY_NOTICES.md` inventoried; subcase selection from the first run with the one stated whole entry point; scope 39 → 40 (§0, §1.6, §3, §5.3, §5.4, §6.3, §11.2, §11.3) |
| Q1–Q8 | §12 dispositions: Q1 scoped R0 inventory; Q2 settled (D-8); Q3 per C1/C4; Q4 per C2; Q5 no identity scope; Q6 one Goal (§11.1); Q7 D-6 accepted; Q8 defaults accepted and enforced |
