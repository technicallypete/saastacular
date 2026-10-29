> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G062 — Root-host tenant provisioning signup: Implementation Plan (r5, PROVISIONAL)

Goal: [G062](../goals/G062-root-host-tenant-signup.md) · Spec: [G062](../specs/G062-root-host-tenant-signup-spec.md).
Status: **provisional Plan r5 (two bounded corrections after the Codex r4 CORRECTION LOOP recorded in
`management/v0/results/G062-plan-codex-review.md` at staging `71ce8d28`, plus current-source reconciliation); no Plan
ACCEPT and no execution release.** Planning only: no implementation, tests, builds, probes, preview, provider/DNS/TLS
or infrastructure action occurred. Inputs: `management/v0/plans/G062-planning-inputs.json`. Proposed scope allow-list:
`management/v0/results/G062-scope-allowlist.txt` (40 ERE patterns; pattern set unchanged since r4; SHA-256 in the
inputs JSON and the handback). r0–r4 bytes are preserved in commits `3ec3732a`/`4c0c165d`/`e0ae0b1c`/`4a73fff0`/
`f6ac134d` on `g062-plan`; §13 maps every correction and the decision to its change.

**Owner decision record (D-4 CLOSED — source: owner via Codex, 2026-10-20).** Owner flow, verbatim: *"authenticate or
create the identity at root signup, enter an Organization name plus a platform-wide unique tenant slug, create the
Organization with the authenticated user as owner, then automatically redirect into that tenant."* Owner: *"That should be
the answer to D4. Do you agree?"* Codex agrees. D-4's UX choice is **automatic continuation**: no extra manual Continue
button and no planned second sign-in action. This is the owner's UX choice only; it is not Plan ACCEPT or execution release.

## 0. Baseline, dependencies and gates

- Baseline `origin/staging 47df73f5a81df653a90c9a636bfec4e7b538a759` (migration 00017 is the head; accepted
  G016/G017/G018/G019/G058/G060 code integrated). Staging advanced to `d0226561` (management-only, verified by Codex)
  and then to the G055 merge `5b813e32` and the current head `71ce8d28` (r5 check): over the G062 input set (`core/`,
  `web/`, `infra/`, `.env.example`, `docs/foundation/`, architecture/security/decisions/ADRs/performance, Goal/Spec) the
  only delta is still the skills-path link in `docs/SECURITY.md:446` and five lines of `docs/foundation/OPERATIONS.md`
  (not a G062 input); bindings stay `47df73f5`, no rebase. The management index is re-bound at `71ce8d28`
  (`PROGRESS.md` `55a38d7a…`, `ROADMAP.md` `c2d801a6…`; inputs JSON r5). R0′ rebinding will hash SECURITY.md at the
  then-current staging.
- **Technical dependencies (all in the baseline):** G016 host model/resolver/session binding (00011/00012,
  `identity.tenantOrigin`, `issuanceGuard`); G019 owner lock order, authority guard and the claim-writer DEFINER/role
  precedent (00015, `10-roles.sh`, `provision-roles.sql`); G058 enrollment-intent login, sealed address evidence,
  limiter and cookie contract (00016, `identity/enroll.go`, `enrollment/api`); G060 `group_management_override DEFAULT
  false` (00017); G013 restore invalidation (`infra/recovery/sql/invalidate.sql`, `recovery.sh invalidate-auth`).
- **Scheduling order only:** G055 preceded G062 in delivery and is now **merged and finally ACCEPTed** (r5
  reconciliation): the accepted bundle is `skills/restruct-sdlc/` with the aliases `.agents/skills/restruct-sdlc` and
  `.claude/skills/restruct-sdlc`; the standalone `skills/sdlc-management` paths are retired, and governance links now
  point at `skills/restruct-sdlc/sdlc-management/references/...`. G055 touched skills/management only, not
  `core/`/`web/`/`infra/postgres|recovery`, and no longer blocks G062 on sequence. G062 ACCEPT precedes G020 execution.
  G057 consumes G062 (live root-callback proof); it is not a prerequisite. The separate certifi maintenance and the
  `0.2.0-rc.1` cut are **not** G062 dependencies (Codex r4 review).
- Gates before execution: Codex Plan ACCEPT with explicit security/architecture review of §3/§4/§10 (D-4 is now closed
  by the owner; the technical findings of the r2 review remain Codex's to disposition); current-input rebinding (R0′);
  B5/checkpoint 3 (zero attempts spent); resource readiness; recorded supervisor release. G021 barrier unchanged.

## 1. Approach

1. **Root and `www` landing, `/signup` (web, thin).** Core's `hosts/resolve` keeps kind `public_entry` for `B` and
   `www.B` (Spec R1 preserves configured root/`www` aliases) and replaces `location` with `signup_origin` = the exact
   configured platform root origin (`https://B[:PLATFORM_HOST_PORT]`). Web (`proxy.ts`) passes `/` on both; on `B`
   also `/signup`, `/signup/*` and `/auth/*`; every other path on either is the fixed 404 (`/api/health`, `/api/ready`
   stay exempt). The landing's only authentication/acquisition action is **Sign Up**, an absolute link to
   `${signup_origin}/signup` (so `www.B` canonicalizes to `B`). Under `loopback-http` the gate is skipped and `/` keeps
   the generic page (`[S-LANDMARKS]` unaffected).
2. **Platform signup intent (core identity; mirrors G058's enrollment intent).** `POST /internal/v1/auth/login` gains
   `intent: "signup"`, accepted **only** when `origin` equals the exact root origin (never `www.B`, tenant, auth or
   custom: `origin_rejected`); the login row records `signup_intent = true`, no Organization, return path
   `/signup/details`; redirect URI is the root's own `/auth/callback` (no relay). After `consumeLogin` + `Exchange`,
   the signup branch (T1) creates the mapping when unknown (ADR-007 G062 amendment), never merges by email, and inserts
   one `app.signup_sessions` row (30 min; host-only `__Host-signup`; `Authorization: Signup <cred>`). It is not a
   session: no `app.sessions` row, no TenantContext; no existing consumer accepts the scheme or reads the table.
3. **Constrained provisioning (T2).** `POST /internal/v1/signup/organizations` runs one IDENTITY-shape transaction
   calling `app.signup_provision_organization(...)`, a SECURITY DEFINER body owned by the new NOLOGIN helper
   `<db>_provision_writer` (provisioned by `10-roles.sh`/`provision-roles.sql` exactly like `<db>_claim_writer`; §3.3):
   users `FOR SHARE` (active) → own signup row `FOR UPDATE` → INSERT Organization (name, slug; unique index = one winner)
   → INSERT `org_owner` binding → record intent/result/handoff. Capture triggers write `organization.created` and
   `membership.granted` in the same statement under a pw-scoped audit policy. Runtime/owner keep no Organization or
   `org_owner` INSERT path.
4. **Automatic continuation into the new tenant (T3; owner D-4 decision) with Spec R4/AC7 browser continuity
   (r4-1: tenant challenge → root proof with the original signup credential → tenant completion; all automatic 303s,
   no Continue button, no provider-SID assumption).** The design proves, before the tenant host starts the G016 login,
   that the browser at the tenant host is the browser that holds the **original host-only root signup credential**.
   Four one-use secrets per *continuation generation* live on the signup row as hashes (§3.2); every hop is a 303 with
   `Cache-Control: no-store` and `Referrer-Policy: no-referrer`, and no secret is logged:
   1. **Create (root, POST, Origin + CSRF + `__Host-signup`).** T2 commits and mints the **initial reference R**
      (256-bit; `ref_hash`, 120 s). The root answers 303 `https://<slug>.B/signup/continue?ref=R`. `__Host-signup` is
      **kept** (r4-2; §4.4).
   2. **Challenge (tenant, GET `/signup/continue?ref=R`).** The tenant host (runtime: resolves its own Organization
      with `hostname_state`) calls `signup_challenge_issue` (§3.4): consumes R one-use (`ref_consumed_at`), stores
      `challenge_hash = H(C)` and `nonce_hash = H(N)` (120 s), returns the secrets to web, which sets the **host-only tenant
      challenge cookie `__Host-sct` = C** (`Path=/; HttpOnly; SameSite=Lax; Secure; Max-Age=120`) in the same response and
      303s to `https://B/signup/prove?n=N`. C never leaves this browser; N travels once in the Location.
   3. **Proof (root, GET `/signup/prove?n=N`).** The root host reads the browser's **original `__Host-signup`** and calls
      `POST /internal/v1/signup/prove {origin: root, nonce: N}` with `Authorization: Signup <cred>` (runtime, IDENTITY
      shape of the credential's user, own-row policy): the row selected by `credential_hash` must be the row whose
      `nonce_hash = H(N)` with `nonce_consumed_at IS NULL`, unexpired, unrevoked, same generation; core marks the nonce
      consumed and mints the **proof P** (`proof_hash = H(P)`, 120 s), and the root 303s to
      `https://<slug>.B/signup/complete?p=P`. A missing/foreign/other-row/revoked/expired credential ⇒
      `signup_context_mismatch`: the UPDATE matches **no row, nothing is consumed**, and the generation stays exactly
      as it was (usable only by the browser that holds the row's original credential, until expiry or re-mint).
   4. **Completion (tenant, GET `/signup/complete?p=P`).** The tenant host reads **`__Host-sct` = C** from the cookie and
      calls `signup_challenge_complete(H(C), H(P), origin, org, now)` (§3.4): both hashes must match the same row and
      generation, `proof_consumed_at IS NULL`, unexpired, origin/Organization match, user active ⇒ `completed_at` set and
      the pinned login starts (runtime: `is_organization_member`, pinned `login_transactions` row with `expected_user_id`,
      fresh `__Host-login`; `__Host-sct` cleared) ⇒ 303 WorkOS ⇒ silent code return while the provider session is live
      (emulator-observed; G057-qualified for real WorkOS) ⇒ pinned callback (`vi.Subject` ↦ `expected_user_id`,
      `issuanceGuard` host/membership/status) ⇒ session ⇒ `/?status=tenant_created` (closed status via a new callback
      outcome, §4.2/r4-3; `return_path` stays `/`).
   **What is bound, and by what (Spec R4/AC7, no residual relied on):**
   - **browser (root ↔ tenant):** completion needs C (host-only cookie set only in the browser that performed step 2 on
     the tenant host) **and** P (issued only to a browser presenting the original `__Host-signup` at step 3 **and** 303'd
     back into that same browser). A browser that captured R performs step 2 (R is spent, that browser gets its own C
     and N) but fails step 3 (no `__Host-signup` ⇒ `signup_context_mismatch`, **no row change, nothing consumed**).
     **Bounded availability consequence (accepted at Plan level by the Codex r4 review; no claim of zero impact):** the
     *temporary generation* is consumed by the captured R, so the original browser's own `continue` 303 finds R spent
     and lands on the root recovery page; the **durable** root signup credential and create intent are intact, the
     owned tenant is unchanged, no tenant session exists anywhere, and the original browser's authenticated "Try again"
     re-mints generation g+1 (all four secrets of g dead) and completes with no duplicate tenant/owner/audit. A browser
     that captured N and tricks the original root browser into step 3 (link/img) makes the *original* browser mint P and
     carry it to step 4, where it lacks that attacker's C ⇒ **no match, nothing consumed**, and the attacker never
     receives P; P stays bound to a C the attacker holds only in its own browser and expires with the generation (≤ 120
     s) or at re-mint — what genuinely remains is that an attacker who has already forced the victim's step 3 would
     additionally have to read P from the victim's address bar within that window (`no-store`/`no-referrer`; 256-bit).
     A browser that captured P lacks C ⇒ **no match, nothing consumed**. Each secret is one-use per generation; a
     re-mint invalidates all four. The tenant `__Host-login` binding then covers the callback as today.
   - **root signup context:** step 3 accepts only the exact original credential of the row (`credential_hash` ↔ row ↔
     `nonce_hash`); another user's credential, another signup row of the same user, a revoked/expired credential or no
     credential ⇒ `signup_context_mismatch`.
   - **user:** after `Exchange`, before any write, `vi.Subject` must map to `expected_user_id`, else
     `signup_user_mismatch` (consumed; no session; no mapping change; `auth.sign_in` denied). A legitimate
     same-user/new-provider-session callback (the user re-authenticated at WorkOS meanwhile) is **accepted**: continuity
     rests on C/P/`__Host-signup`/`__Host-login` and the user pin, never on the provider `sid`.
   - **Organization and exact host:** steps 2 and 4 require the runtime-resolved request Organization = the row's
     `organization_id` and `target_origin` = the exact request origin; the pinned row is rechecked by `issuanceGuard`
     (`host_changed` on a slug rename/host reassignment; `organization_not_member`; `user_disabled`); wrong port/scheme
     never match.
   - **purpose, one use, expiry:** only `signup_handoff` rows carry `expected_user_id` and only step 4 creates them;
     R/N/P/C are one-use and ≤ 120 s; the login row keeps its 10-minute one-use TTL; step 3 and step 4 are also
     rate-limited (§4.5).
   **Provider session not live** (signed out at WorkOS or provider session expired): steps 2–4 are unaffected (they
   never touch the provider); at the final WorkOS hop the hosted prompt appears instead of a silent return; the user
   authenticates **as the expected user** and the pinned callback completes to `/?status=tenant_created`. Another user
   ⇒ `signup_user_mismatch` → tenant `/sign-in?status=failed`; cancel ⇒ `login_canceled` → `/sign-in?status=canceled`.
   The tenant **already exists and is owned** in every non-happy case (never rolled back by a browser outcome); no
   Continue button is added anywhere on the happy path.
   **Recovery (r4-2; three distinct cases, same Organization, no duplicate owner/audit rows):** (a) **lost create
   response** — the browser never left `/signup/details`; an identical re-POST answers `already_created` and mints
   generation g+1 (R′; old R/N/P/C dead); (b) **acknowledged create, failed continuation** — the browser was redirected
   and a hop refused (`continuation_invalid`); the tenant sends it to `https://B/signup/details?status=continue_failed`
   (root; `__Host-signup` still present within its 30-minute life), which shows the closed text and one **"Try again"**
   POST (identical payload ⇒ `already_created`, generation g+1) — a failure-path action, not a happy-path button; (c)
   **credential gone** (expired/abandoned/revoked) — `/signup/details` 303s to `/signup` and the owner signs in at the
   tenant `/sign-in` (ordinary G016 login, membership true) — the manual fallback, not part of the designed flow. No
   cookie crosses hosts; no bearer; the destination is derived from the canonical created tenant and
   `PLATFORM_BASE_DOMAIN`, never from input.
5. **Platform-host routing is a database mapping only.** Platform tenants resolve through the **preconfigured wildcard
   DNS/ingress/TLS for `*.<PLATFORM_BASE_DOMAIN>`** (DOMAINS.md "Edge deployment inputs": the shared wildcard certificate
   and ingress outside the tenant lifecycle; `hostname_state('platform', slug)` is one unique-index probe). Organization
   creation adds **only the slug-to-UUID mapping row**; **no per-tenant CNAME/DNS record is created and no DNS-propagation
   wait exists in onboarding**. Custom-domain claim/verification/TLS readiness remains the separate existing G016 flow and
   is unaffected. The G062 local overlay proves this with the existing `*.localhost:543` proxy and DNS/TLS fixtures; real
   wildcard ingress/TLS behaviour is G057-qualified (§5.6).

Riskiest slice first: role bootstrap + 00018 + identity signup branch + provisioning body + pinned-callback binding and
automatic continuation, with catalog, direct-SQL refusal, atomicity, race, lost-response, binding and not-live-provider
proofs (Batches 1–2) before UI (Batch 3).

## 2. Files and modules (touch-list; new unless marked *modify*)

| Area | File | Change |
|---|---|---|
| roles *modify* | `infra/postgres/init/10-roles.sh`, `infra/postgres/provision-roles.sql` | fourth helper `<db>_provision_writer` (NOLOGIN NOSUPERUSER NOBYPASSRLS NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION; owner its only member `INHERIT FALSE, SET TRUE`); `mk_provision_writer`/`g_provision_writer` \gset branches; unsafe-attribute and membership allow-lists extended; header "before G062 adds only the fourth helper" |
| migration | `core/migrations/00018_platform_signup.sql` | precondition requires the four helpers (00015:46-50 pattern); tables/columns, bodies, reader, grants/policies, audit catalog delta, postconditions (§3) |
| migration *modify* | `core/migrations/migrations_test.go` | `TestMigrationSet` 00001–00018 |
| core | `core/internal/signup/{signup.go,store.go,slug.go,names_copy.go,errors.go,log.go}` | credential resolution, provisioning, handoff mint/re-mint/consume, slug suggestion/validation, pinned copy of `tenancy.ValidateName`, closed codes, `signup.request` log event |
| core | `core/internal/signup/api/{http.go,limiter_copy.go}` | `/internal/v1/signup/*` (§4.2); byte-pinned copy of the `enrollment/api` limiter (parity test) |
| core *modify* | `core/internal/identity/{service.go,store.go,enroll.go,identity.go,http.go,allowlist.go}` + new `core/internal/identity/limiter_copy.go` | `LoginInput.SignupIntent`; `platformRootOrigin`; `loginTxn.{SignupIntent,SignupHandoff,ExpectedUserID}`; `Service.LoginPinned` (automatic continuation: runtime `tenantOrigin` host resolution, DEFINER reference consumption in its own short transaction, then runtime `is_organization_member` and the pinned-row INSERT under the existing `login_transactions_insert` policy); callback signup branch (T1) and the user/Organization pin checks before `signIn`; `SignupCredential` via `app.auth_resolve_signup`; expired sweep; revocation on logout/logout-all/disable; `ReturnPath` + `/signup/details`; **root-origin callback limiter** (§4.5); HTTP shapes |
| core *modify* | `core/internal/hostname/hostname.go`, `core/internal/config/domains.go`, `domains_config_test.go` | `PlatformRootOrigin(cfg)`; `IsPlatformRoot(h)`; **remove `MarketingSiteURL`** (field, `domainVariables`, validation, tests) |
| core *modify* | `core/internal/domains/api/http.go`, `http_test.go` | `public_entry` answer `{kind, signup_origin}` (no `location`) |
| core *modify* | `core/internal/audit/audit.go`, `audit_test.go` | catalog delta (§3.6); `TestCatalogParity` → `mig18` |
| core *modify* | `core/cmd/server/main.go` | mount `signup/api`; wire the signup service and the root-origin callback limiter |
| core tests | `core/internal/migrate/signup_integration_test.go`; *modify* `rls_integration_test.go` (`rlsLabel` gains `pw`; inventory strings), `membership_integration_test.go` (role list), `audit_integration_test.go`, `integration_test.go` | `TestSignupCatalog`; kept owner/runtime INSERT 42501 expectations; audit parity |
| core tests | `core/internal/signup/*_test.go`, `signup_integration_test.go`; *modify* identity tests (§5); `core/internal/tenancy/access/recovery_integration_test.go` (one `TestRecoveryPost` subcase) | |
| core tests *modify* | `core/internal/tenancy/boundary_test.go`, `access/boundary_test.go`, `parity_test.go` | expected zero delta (`signup` does not import `tenancy`, D-9) |
| recovery *modify* | `infra/recovery/sql/invalidate.sql` | fold: `UPDATE app.signup_sessions SET expires_at = least(expires_at, now()), continuation_expires_at = least(continuation_expires_at, now()) WHERE revoked_at IS NULL` + count `signup_sessions_expired` + postcondition; **no new SQL file, `recovery.sh` unchanged** (`cmd_invalidate` invokes only `invalidate.sql`, recovery.sh:713-720) |
| harness *modify* | `infra/validation/run.sh` (`WEB_ROUTES`; B-MIGRATE-SET/CLI/CATALOG at 18; B-AUDIT-CATALOG path → 00018; PROV-1 admin-residual and PROV-6 delta expected strings + B-RLS-MIGRATION/B-DBRENAME PASS texts and role lists gain `provision_writer`; `core_env` drops `MARKETING_SITE_URL` (run.sh:1880); new `B-SIGNUP-CATALOG`/`B-SIGNUP-IT`; browser matrices; `identity:browser` projects), `lib.sh`, `required-checks.tsv`, `tools/internal/logrecord/{events.json,completion.schema.json,logrecord_test.go}`, `infra/ci/profile-paths.tsv` (`web/app/signup/`, `core/internal/signup/`, `infra/postgres/` already foundation) | |
| overlay *modify* | `infra/compose/identity.yml` (`--redirect-hosts` + `localhost`; drop `MARKETING_SITE_URL`; drop `marketing.identity.test` from `identity-setup --hosts`), `infra/identity/proxy.conf` (drop the `marketing.identity.test` server block), `infra/identity/emulator-seed.yaml` (comment only) | |
| config *modify* | `.env.example` (drop lines 80–86's marketing text/variable; re-pin `ENV_DOMAIN_BLOCK_SHA`, `ENV_IDENTITY_DOMAIN_TAIL_SHA` in `run.sh:636/641`) | |
| web *modify* | `web/proxy.ts`, `web/lib/host.ts`, `web/lib/session.ts`, `web/lib/identity-client.ts`, `web/app/page.tsx`, `web/app/auth/callback/route.ts`, `web/app/sign-in/page.tsx`, `web/playwright.config.ts` (drop the marketing host-resolver rule; add `signup-light/dark`) | per §4.4 |
| web | `web/lib/signup-client.ts`; `web/app/signup/{page.tsx,details/page.tsx,start/route.ts,create/route.ts,abandon/route.ts,continue/route.ts,prove/route.ts,complete/route.ts}`; `web/components/signup/{landing,signup-form,details-form}.tsx` | reuse `AppShell`, `Section/Kicker/Heading/Text`, `FormField`, `PendingSubmit`, `StatusMessage`, `TextLink`; `continue`/`complete` are tenant-host routes, `prove` a root-host route (§4.4) |
| web tests | `web/tests/unit/{signup-client,signup-forms}.test.mjs`; *modify* `host.test.mjs`, `session.test.mjs`; `web/tests/browser/signup.spec.ts`; *modify* `identity.spec.ts` (E-DOM-4 rewrite; `MARKETING` constant removed) | |
| docs *modify* | `docs/ARCHITECTURE.md` (§51 row), `docs/SECURITY.md`, `docs/PERFORMANCE.md` (PG-009), `docs/foundation/{IDENTITY,DOMAINS,AUTHORIZATION,DATABASE-ISOLATION,AUDIT,TENANCY,WEB-SHELL,MIGRATIONS,RECOVERY}.md` | supersession notes (§9) |

## 3. Schema and migration — `core/migrations/00018_platform_signup.sql` (flagged for explicit security review)

3.1 **Preconditions.** Version-17 catalog as pinned (cw 7 bodies; `events_action_valid`; `organizations` triggers) and
the four helpers present with safe attributes and owner-only membership (`to_regrole(db||'_provision_writer')`,
refusing with "run infra/postgres/provision-roles.sql"). `login_transactions` gains `signup_intent boolean NOT NULL
DEFAULT false`, `signup_handoff boolean NOT NULL DEFAULT false` and `expected_user_id uuid NULL → users`; CHECKs: `signup_intent` ⇒ `organization_id IS NULL AND NOT enrollment_intent`;
`signup_handoff` ⇒ `organization_id IS NOT NULL AND expected_user_id IS NOT NULL AND NOT signup_intent AND NOT
enrollment_intent`; `NOT signup_handoff` ⇒ `expected_user_id IS NULL`. No provider-session column (r2-1).
**Return-path closed sets (r4-3), both updated together:** `ALTER TABLE app.login_transactions DROP CONSTRAINT
login_transactions_return_path_check, ADD CONSTRAINT login_transactions_return_path_check CHECK (return_path IN ('/',
'/account', '/join/status', '/signup/details'))` (00016:172-173 pattern); Go `identity.returnPaths` = `{"/", "/account",
"/signup/details"}` (`allowlist.go:28`; `/join/status` is set by code, not requested). The tenant completion keeps
`return_path = '/'` and carries the status through a new closed callback **outcome** `tenant_created` (§4.2), so
`/?status=tenant_created` is never a stored return path and `ReturnPath()` keeps reducing every non-member of the set to
`/`. Down restores the 00016 set and refuses while any row has `return_path = '/signup/details'`; the postcondition
pins the four-member CHECK text. Proofs: `TestSignupCatalog` (CHECK text; INSERT of `/signup/details` admitted; INSERT
of `/signup`, `/signup/details/`, `//evil`, `https://evil`, `/%2Fevil`, `/signup/../account` ⇒ 23514), extended
`TestReturnPathAllowList` (positives `/`, `/account`, `/signup/details`; negatives external URL, scheme-relative `//`,
backslash `/\evil`, encoded `%2F%2F`, traversal `/signup/details/../account`, trailing slash, query/fragment, mixed case
⇒ `/`), `TestLoginStart` (signup intent stores `/signup/details`), `TestCallbackPinned` (pinned success answers
`tenant_created` with `return_to = "/"`; web maps it to `/?status=tenant_created` from its closed map).

3.2 **Table `app.signup_sessions`** (FORCE RLS): `id uuid PK`, `user_id → users`, `origin text`, `credential_hash
bytea UNIQUE`, `csrf_token bytea`, `provider_session_id text NOT NULL`, `created_at`, `expires_at`, `revoked_at`,
`revoke_reason`, `intent_hash bytea NULL`, `organization_id uuid NULL → organizations`, `provisioned_at NULL`,
`target_origin text NULL`, and the **continuation state** (r4-1; one generation at a time): `continuation_generation
integer NOT NULL DEFAULT 0`, `continuation_expires_at NULL`, `ref_hash bytea NULL UNIQUE`, `ref_consumed_at NULL`,
`challenge_hash bytea NULL UNIQUE`, `nonce_hash bytea NULL UNIQUE`, `nonce_consumed_at NULL`, `proof_hash bytea NULL
UNIQUE`, `proof_consumed_at NULL`, `completed_at NULL`. CHECKs: provisioned ⇔ `organization_id`, `intent_hash`,
`provisioned_at`, `target_origin` set; continuation columns NULL unless provisioned; `challenge_hash`/`nonce_hash` set ⇔
`ref_consumed_at` set; `proof_hash` set ⇒ `nonce_consumed_at` set; `completed_at` set ⇒ `proof_consumed_at` set. No
issuer/subject/auth_time (nothing exchangeable for a session); every secret is stored as SHA-256 only.

3.3 **Helper role `<db>_provision_writer`** (`pw`) — created by **`10-roles.sh` (fresh volume) and
`provision-roles.sql` (existing volumes; adds only the fourth helper and its membership on a pre-G062 volume)**, never
by a migration (the migrator/owner has no CREATEROLE). Grants (in 00018, via the 00006 `SET ROLE` chain): USAGE
`app`/`audit`; EXECUTE on the four context parsers only; `users` SELECT(id, status) + lock-only UPDATE(status);
`signup_sessions` SELECT + UPDATE of exactly **14** columns (intent_hash, organization_id, provisioned_at, target_origin,
continuation_generation, continuation_expires_at, ref_hash, ref_consumed_at, challenge_hash, nonce_hash,
**nonce_consumed_at, proof_hash** (r5-1: needed only so the re-mint can reset both to NULL atomically),
proof_consumed_at, completed_at); non-null proof issuance stays confined to the authenticated root path by the pw UPDATE
policy's WITH CHECK (`proof_hash IS NULL OR completed_at IS NOT NULL`, §3.5) and the pinned body text (only the re-mint
branch touches the two columns, and only to NULL); `organizations` INSERT(id, name, slug) + SELECT(id, slug);
`role_bindings` INSERT of the eight 00015 columns; `audit.events` INSERT of the nine claim-writer columns. **Explicitly
not granted (r2-3/r2-4; pinned as negatives in `TestSignupCatalog`): EXECUTE on `app.hostname_state` and
`app.is_organization_member` (both stay runtime-only per 00011:774-775), any privilege on `login_transactions`, DELETE
anywhere, audit SELECT; no pool, no credential.** The host-routing and membership checks of the handoff path run in the
runtime's own LOGIN-shape transaction (§3.4, §4.3), where the runtime already holds those EXECUTEs.

3.4 **DEFINER bodies (00018 allowance = 3 pw + 1 ar; owner/volatility/search_path/ACL/body hash pinned by `TestSignupCatalog`):**
- `app.signup_provision_organization(p_hash bytea, p_name text, p_slug text, p_intent bytea, p_ref bytea, p_target text, p_now timestamptz) RETURNS TABLE(outcome text, organization_id uuid, slug text, generation integer)` — VOLATILE, pw; refuses unless
  IDENTITY shape (42501 CONSTRAINT `signup_context`). Order: (1) `users … FOR SHARE` → not `active` ⇒ `user_disabled`;
  (2) own row `FOR UPDATE` (`credential_hash = p_hash AND user_id = ctx_user() AND revoked_at IS NULL AND expires_at >
  p_now`) ⇒ none: `signup_invalid`; (3) if provisioned: `intent_hash = p_intent` ⇒ **re-mint generation g+1**:
  `ref_hash = p_ref, continuation_expires_at = p_now + 120 s, ref_consumed_at/challenge_hash/nonce_hash/
  nonce_consumed_at/proof_hash/proof_consumed_at/completed_at = NULL, continuation_generation + 1` (every secret of the
  previous generation stops matching; correction r0-4/r4-2) and return `already_created`; else `intent_used`; (4)
  grammar/reserved are the 00011 CHECKs; (5) `INSERT organizations (id = gen_random_uuid(), name, slug)` (no RETURNING;
  defaults supply `invite_only`, `created_after_ledger`, override false); `unique_violation` in a sub-block ⇒
  `slug_taken` (zero writes retained); (6) `INSERT role_bindings (… subject = ctx_user(), org_owner, 'membership')`;
  (7) UPDATE own row with intent/result/target/generation 1/`ref_hash`; return `created`. Lock order users → signup row
  → organizations → role_bindings (G019 F-01: users first).
- `app.signup_challenge_issue(p_ref bytea, p_origin text, p_org uuid, p_challenge bytea, p_nonce bytea, p_now timestamptz) RETURNS TABLE(ok boolean)` — VOLATILE, pw, callable in LOGIN shape (`ctx_login_binding()` NOT NULL,
  `ctx_user()`/`ctx_org()` NULL, else 42501 CONSTRAINT `signup_continuation_context`). One statement: `UPDATE
  signup_sessions SET ref_consumed_at = p_now, challenge_hash = p_challenge, nonce_hash = p_nonce WHERE ref_hash = p_ref
  AND ref_consumed_at IS NULL AND continuation_expires_at > p_now AND target_origin = p_origin AND organization_id =
  p_org AND revoked_at IS NULL AND EXISTS (SELECT 1 FROM app.users u WHERE u.id = user_id AND u.status = 'active')`;
  zero rows ⇒ `continuation_invalid` (R is spent by any first match; `p_org` is the runtime-resolved request
  Organization; the body calls no reader). Secrets C and N are generated by core (`crypto/rand`, 256 bits) and reach SQL
  only as hashes.
- `app.signup_challenge_complete(p_challenge bytea, p_proof bytea, p_origin text, p_org uuid, p_now timestamptz) RETURNS TABLE(user_id uuid, organization_id uuid)` — VOLATILE, pw, LOGIN shape (same guard). One statement:
  `UPDATE signup_sessions SET proof_consumed_at = p_now, completed_at = p_now WHERE challenge_hash = p_challenge AND
  proof_hash = p_proof AND proof_consumed_at IS NULL AND nonce_consumed_at IS NOT NULL AND continuation_expires_at >
  p_now AND target_origin = p_origin AND organization_id = p_org AND revoked_at IS NULL AND EXISTS (… active user)
  RETURNING user_id, organization_id`; zero rows ⇒ `continuation_invalid`. The UPDATE matches only when **both**
  hashes belong to the same row and generation, so a wrong-C or wrong-P attempt changes nothing and spends nothing;
  guessing is denied by the 256-bit secrets and the per-tenant-origin limit on step 4 (§4.5).
  The caller (`identity.Service.LoginPinned`, runtime) then, in a second short LOGIN transaction, runs
  `app.is_organization_member(user, org)` (runtime EXECUTE, 00011) and inserts the **pinned** `login_transactions` row
  (`origin = p_origin`, `organization_id = p_org`, `signup_handoff = true`, `expected_user_id = user_id`, `return_path
  = '/'`, state/binding/verifier/client/TTL as the ordinary row) under the unchanged `login_transactions_insert`
  policy (`binding_hash = ctx_login_binding()`, 00006:470) through the existing `insertLogin` path. A failure there
  leaves the generation completed and no login row (`continuation_invalid` → recovery (b), §1.4). Pinning by the runtime
  is a *restriction* (which user may complete), never authority (G008 trust limit unchanged).
- **Step 3 is runtime SQL, not a DEFINER body** (the credential holder's own row, IDENTITY shape): `UPDATE
  app.signup_sessions SET nonce_consumed_at = $now, proof_hash = $proof WHERE credential_hash = $cred AND user_id =
  app.ctx_user() AND nonce_hash = $nonce AND nonce_consumed_at IS NULL AND continuation_expires_at > $now AND
  revoked_at IS NULL` under `signup_sessions_prove` (§3.5); zero rows ⇒ `signup_context_mismatch` with **no row change
  and nothing consumed** (r5-2): a foreign, revoked or expired credential, or a nonce of another row, matches nothing,
  so no stranger's generation is burned; a nonce presented with a different credential of the **same** row is impossible
  by construction (one credential per row). There is no second "consume anyway" statement.
- **Atomicity.** Each hop is one row-level UPDATE whose WHERE encodes the complete precondition; the row is the single
  serialization point (unique hashes; READ COMMITTED re-evaluation under concurrent attempts yields exactly one winner
  per secret per generation). No hop holds a lock across a redirect or a provider call. Re-mint (generation g+1) takes
  the row `FOR UPDATE` inside the provisioning body, so a concurrent hop of generation g either committed before (and is
  then superseded) or matches nothing.
- `app.auth_resolve_signup(bytea, timestamptz)` — STABLE, `<db>_auth_reader`: unrevoked, unexpired row of an active
  user; returns id, user_id, origin, csrf_token, expires_at, organization_id, slug, target_origin, provisioned_at,
  continuation_generation, completed_at (the status/recovery projection; no hash is ever returned).

3.5 **Grants and policies.** `rt`: `signup_sessions_insert` (T1; `user_id = ctx_user() AND ctx_org() IS NULL`),
`signup_sessions_revoke` (UPDATE revoked_at/reason, own rows), **`signup_sessions_prove`** (UPDATE of exactly
`nonce_consumed_at`, `proof_hash`; USING/WITH CHECK `user_id = ctx_user() AND ctx_org() IS NULL`; column-limited UPDATE
grant), `signup_sessions_expire` (EMPTY shape DELETE of expired rows in the login-start sweep). The pinned login row is a runtime INSERT admitted by the **existing**
`login_transactions_insert` policy (`binding_hash = ctx_login_binding()`); the new CHECKs of §3.1 bound its shape; no
new `login_transactions` policy or helper grant exists (r2-4). `ar`: `signup_sessions_auth_reader`. `pw`:
`signup_sessions_provision_select/_update` (the provisioning re-mint, challenge-issue and challenge-complete UPDATEs;
r5-1: USING `true` on provisioned rows of an active user, **WITH CHECK `(proof_hash IS NULL OR completed_at IS NOT NULL)`**
so pw may reset `nonce_consumed_at`/`proof_hash` to NULL at re-mint and keep an existing non-null `proof_hash` only while
completing, but can never issue a proof itself; the runtime's `signup_sessions_prove` remains the sole non-null `proof_hash`
writer), `organizations_provision_insert` (WITH CHECK
`pg_trigger_depth() = 0 AND ctx_user() IS NOT NULL AND ctx_org() IS NULL AND slug IS NOT NULL`),
`role_bindings_insert_owner` (`subject_user_id = ctx_user() AND scope_type = 'organization' AND group_id IS NULL AND
role = 'org_owner' AND source = 'membership'`), `users_provision_select/_lock`, and **`events_insert_provision`**
(`pg_trigger_depth() >= 1 AND action IN ('organization.created','membership.granted') AND actor_user_id = ctx_user()`;
mirrors `events_insert_claim`; required because `events_insert` demands `ctx_org()`). Runtime/owner keep **no** INSERT
path on `organizations`/`org_owner` bindings (TestRLSCatalog 42501 unchanged). `pw` has no policy on any other table.
Catalog negatives (`TestSignupCatalog`): `has_function_privilege(pw, 'app.hostname_state(text,text)', 'EXECUTE')` and
`… is_organization_member …` are false; `has_table_privilege(pw, 'app.login_transactions', 'INSERT|SELECT|UPDATE|DELETE')`
false; a pw-executed direct `INSERT INTO app.login_transactions` ⇒ 42501; a runtime INSERT of a `signup_handoff` row
without a matching `ctx_login_binding()` ⇒ 42501; a `signup_handoff` row with `expected_user_id IS NULL` ⇒ 23514;
(r5-1) the pw column-privilege set on `signup_sessions` equals exactly the 14 columns above (`information_schema.
column_privileges`), a pw-executed `UPDATE … SET proof_hash = <non-null> WHERE completed_at IS NULL` ⇒ 42501 (policy),
a pw-executed `UPDATE … SET nonce_consumed_at = NULL, proof_hash = NULL` on a provisioned row succeeds (the re-mint
shape), and the runtime cannot UPDATE any continuation column except `nonce_consumed_at`/`proof_hash` under
`signup_sessions_prove` (42501 for `ref_hash`, `challenge_hash`, `completed_at`, …).

3.6 **Audit catalog delta** (re-creates the CHECKs as 00016/00017 did): no new action; `auth.sign_in` success
`signup_authenticated` (T1), `tenant_created` (the pinned callback's success reason, alongside the new callback
outcome), denied `signup_origin_rejected`, `signup_user_mismatch`; `access.denied` operations `signup.start`,
`signup.status`, `signup.slug`, `signup.create`, `signup.challenge`, `signup.prove`, `signup.complete`, `signup.abandon`
(65 → 73) with reasons `signup_invalid`, `intent_used`, `slug_taken`, `continuation_invalid`, `signup_context_mismatch`,
`host_changed` (reused). AUDIT.md rows, `TestCatalogParity` (`mig18`) and B-AUDIT-CATALOG's path move to 00018.

3.7 **Postconditions / down.** pw owns exactly 3 bodies; ar +1; cw 7; policy-reader inventory unchanged;
`app_objects_beyond_metadata` 71 → **76** (re-pinned at R0′); the four-member return-path CHECK text. Down refuses
while any `signup_sessions` row, a `signup_intent`/`signup_handoff` login transaction or a `return_path =
'/signup/details'` row exists; otherwise drops policies/bodies/grants/table/columns, restores the 00016 return-path
CHECK and the 00017 catalog CHECKs; it does not drop the role (roles are provisioning-owned, like `claim_writer`).

## 4. Contracts and interfaces

4.1 **Config.** Root origin = `https://` + `PLATFORM_BASE_DOMAIN` (+ `:PLATFORM_HOST_PORT` outside production).
`MARKETING_SITE_URL` is **removed** (D-6 closed): its only consumers are the root redirect (`config/domains.go:21,64,88,92-98`,
`domains/api/http.go:198-199`, their tests, `identity.yml:96`, `run.sh:1880` probe env, `.env.example:80-86`,
DOMAINS.md rows, the `marketing.identity.test` stand-in in `proxy.conf`/`identity.yml:37`/`playwright.config.ts:97`,
`identity.spec.ts` E-DOM-4). No other consumer exists (grep over `core web infra docs .env.example`). An operator
environment that still sets it is ignored (core reads only `domainVariables`).

4.2 **Core HTTP (`/internal/v1`, web BFF only; not public API; no MCP change).**
- *modified* `POST /auth/login` `{origin, intent?: "enroll"|"signup", …}`: `signup` ⇒ `origin_rejected` unless the
  exact root origin; `invalid_request` when combined with enrollment fields.
- *modified* `POST /auth/callback`: signup transaction ⇒ `202 {"signup": {"credential": {value, max_age}, "next":
  "/signup/details"}}` (the stored `return_path` is `/signup/details`, r4-3); pinned (`signup_handoff`) transaction ⇒
  the ordinary answer shape with outcome **`tenant_created`** (session set; `return_to = "/"`; web maps the closed
  outcome to `/?status=tenant_created`), the new denial `signup_user_mismatch` 403 (consumed; nothing issued);
  cancel/expired/replayed/forged ⇒ existing codes, nothing created.
- new `core/internal/signup/api` (`Authorization: Signup <cred>`; JSON ≤ 8 KiB; unknown fields refused; `csrf_token`
  in body; exact `origin` = root origin; `no-store`, `nosniff`, no CORS): `GET /signup/status?origin=` →
  `{status: pending|created|completed, csrf_token, suggested_slug?, organization?: {slug, target_origin}}`; `POST /signup/slug`
  `{origin, csrf_token, name}` → `{suggested, available}` (advisory); `POST /signup/organizations` `{origin, csrf_token,
  name, slug}` → `201 {outcome: created|already_created, organization: {id, slug}, continue: "<target_origin>/signup/continue?ref=…"}`
  (`already_created` with the **same** `{name, slug}` re-mints generation g+1; §3.4), `409 conflict {field:
  slug_taken|intent_used}`, `400 invalid_request {field}`, `403 csrf_failed|user_disabled`, `401 unauthenticated`,
  `429 rate_limited`, `503 signup_unavailable` (unknown commit; nothing claimed; identical retry ⇒ `already_created`);
  **`POST /signup/challenge`** `{origin, ref}` (tenant origin; no credential; `X-Login-Binding` as `/auth/login`) →
  `200 {challenge: {value, max_age: 120}, prove: "<root_origin>/signup/prove?n=…"}`, else `404 continuation_invalid`;
  **`POST /signup/prove`** `{origin: root, nonce}` with `Authorization: Signup <cred>` (no CSRF body token: the request
  is a GET-initiated same-origin BFF call whose only effect is spending this credential's own nonce; the credential
  itself is the authorization) → `200 {complete: "<target_origin>/signup/complete?p=…"}`, else `403
  signup_context_mismatch` / `401 unauthenticated`; **`POST /signup/complete`** `{origin, challenge, proof}` (tenant
  origin; `X-Login-Binding`) → the ordinary `Login` answer `{authorization_url, binding}` (pinned row, `return_to = "/"`),
  else `404 continuation_invalid`; `POST /signup/abandon` → `{}` (revokes the credential and the current generation).
- Credential boundary: `Signup` is refused by `identity.credentialOf`, every session consumer, `enrollment/api` and MCP;
  `Session` is refused by `signup/api`.

4.3 **Go (`core/internal/signup`).** `Resolve`, `SuggestSlug`, `Provision(info, name, slug)` (name via the pinned
tenancy rule; slug via `hostname` grammar + `IsReservedLabel` + configured labels; `intent_hash = SHA-256(name‖0x00‖slug)`;
256-bit initial reference; `target_origin` from configuration only), `Challenge(ref, origin, binding)` (runtime
`tenantOrigin` resolution → `signup_challenge_issue` → returns C and the root `prove` URL carrying N), `Prove(info,
nonce)` (own-row runtime UPDATE under `signup_sessions_prove` → returns the tenant `complete` URL carrying P),
`Complete(challenge, proof, origin, binding)` → `identity.Service.LoginPinned` (short LOGIN transaction calling
`signup_challenge_complete` → second LOGIN transaction: runtime `is_organization_member`, pinned-row INSERT via the
existing `insertLogin` path with `return_path = '/'`; returns the authorization URL), `Abandon`. Identity:
`LoginInput.SignupIntent`, `loginTxn.{SignupHandoff,ExpectedUserID}`, `returnPaths` + `/signup/details`, the
user/Organization pin checks (§1.4) in `Callback` before `signIn` and the `tenant_created` outcome,
`SignupCredential`, `revokeSignupSessions{Bound,All}`, `rootCallbackLimiter`. No tenancy import; no actor/role/tenant parameter.

4.4 **Web (BFF, https profile).** `/` on `public_entry`: landing; Sign Up → `${signup_origin}/signup`
(`data-testid="root-sign-up"`); no sign-in, picker, directory or lookup. `/signup` (root only): `POST /signup/start`
(`readForm` Origin check) → `login({intent:"signup"})` → 303 WorkOS, `__Host-login` set. `/auth/callback` on the root:
`202 signup` ⇒ `__Host-signup` set, 303 `/signup/details`; failures 303 `/signup?status=<closed>`; on `www.B` 404.
`/signup/details`: name + editable suggested slug (`FormField`), `POST /signup/create` → 201 ⇒ **`__Host-signup`
kept** (r4-2), 303 `continue`; 409 ⇒ field error; when the row is already provisioned the page shows the closed
"Your organization is ready" text with the **"Try again"** POST (identical payload) only when it arrives with
`?status=continue_failed` — otherwise it 303s straight into a fresh continuation (re-mint) with no button.
`/signup/continue?ref=` (**tenant** host): `POST /internal/v1/signup/challenge` ⇒ set `__Host-sct` (`sct` under
`loopback-http`; `Max-Age=120`), 303 `${root}/signup/prove?n=`; failure ⇒ 303 `${root}/signup/details?status=continue_failed`.
`/signup/prove?n=` (**root** host): requires `__Host-signup` → `POST /internal/v1/signup/prove` ⇒ 303
`${target}/signup/complete?p=`; failure ⇒ 303 `/signup/details?status=continue_failed` (or `/signup` when the credential
is gone). `/signup/complete?p=` (**tenant** host): requires `__Host-sct` → `POST /internal/v1/signup/complete` ⇒ clear
`__Host-sct`, set `__Host-login`, 303 WorkOS; failure ⇒ clear `__Host-sct`, 303 `${root}/signup/details?status=continue_failed`.
The root origin for these redirects is configuration (`signup_origin` from `hosts/resolve`), never a query parameter.
Tenant `/?status=tenant_created`: existing shell + closed "Your organization is ready" message; no onboarding redirect.
`/signup/*` on tenant hosts is 404 except `/signup/continue` and `/signup/complete`; `/signup/prove` is 404 on tenant
hosts. **Cookie lifetimes (r4-2):** `__Host-signup` 1800 s = row `expires_at`; cleared only on abandon, on any 401 from
the status route, or by expiry (the root cannot observe the tenant session, so completion never clears it; after
`completed_at` the credential can only re-mint its own continuation, read its own status or abandon — bounded 30 min).
`__Host-sct` 120 s, cleared at completion or failure.
`__Host-login` as today. All host-only, `Path=/`, `HttpOnly`, `SameSite=Lax`, `Secure`; nothing set on the auth host.
`WEB_ROUTES` += 8.

4.5 **Rate limits (correction 5; all buckets per process, in memory; honest, not comprehensive).** No limiter exists on
`/internal/v1/auth/login` or `/auth/callback` today (only `mcp/handler.go` and `workos/jwks.go` have one); the only
existing callback guard is structural: a callback performs at most one indexed DELETE and calls the provider only after
a matching one-use transaction, so provider work ≤ started transactions. G062 adds, with the byte-pinned
`enrollment/api` limiter type: (i) **root login start** — one global bucket keyed by the constant `signup/start`
(`rate 2/s, burst 20`); (ii) **root callback** — in the identity HTTP adapter, applied only when `origin` equals the
root origin, keyed per login binding hash (`signup/callback/<sha256(X-Login-Binding)>`, `rate 1/s, burst 5`, max
10 000 keys, reclaim of fully refilled buckets only) **and** one global root-callback bucket (`rate 5/s, burst 50`): a
churned binding header cannot exceed the global bucket, and a pinned binding cannot exceed its own; (iii) **provisioning
and own-status** — per-credential bucket (`credentialRate 5, burst 10`) plus a per-user bucket on `create` (key = resolved
`user_id`, `rate 1/10 s, burst 5`) so a fresh credential cannot replenish it; (iv) **continuation hops** — `challenge` and
`complete` per tenant-origin bucket (`rate 5/s, burst 20`; 256-bit secrets make guessing infeasible within it), `prove`
per credential (the credential bucket of (iii)); every secret is one-use. Exhaustion answers `429 rate_limited` (JSON) / 303 `?status=rate_limited`
(forms). Multi-process deployments multiply capacity by process count; stated in IDENTITY.md.

## 5. Validation and test mapping

5.1 **Existing coverage inspected (reuse unless stated)** — as r0: identity (`TestLoginStart`,
`TestCallbackParametersAndTransactions`, `TestCallbackContinuity`, `TestCallbackRefusals`, `TestMappingIsByIssuerAndSubject`,
`TestNamespaceBinding`, `TestParallelIdenticalCallbacksConsumeOnce`, `TestSignInIsAtomic`, `TestDisableVersusSignInRace`,
`TestIssuanceGuardFailsClosed`, `TestRLSFirstSignInRace`, `TestCustomDomainRelay`, `TestResponsesAreNoStoreWithoutCORS`,
`TestCredentialAndBindingHeaders`, `TestLoginEnrollmentIntent`, `TestCallbackEnrollment`, `TestEnrollmentCredential`,
`TestEnrollmentPendingCredentialRevocation`); enrollment API (`TestEnrollmentCredentialBinding`, `TestEnrollmentRateLimited`,
`TestEnrollmentLimiterChurn`, `TestEnrollmentRoutesPinned`, `TestEnrollmentCopyIsExact`); SQL (`TestRLSCatalog`,
`TestRLSMigration` incl. PROV-1..6, `TestMembershipCatalog`, `TestEnrollmentCatalog`, `TestAuditCatalog`, `TestCatalogParity`,
`TestReservedLabelsFloor`, `TestSlugParity`, `TestMigrationParity`, `TestDomains`, `TestMigrationLifecycle`,
`TestRecoverySeed/Post/Verify/Repair`, `TestAuthority`); harness B-RLS-MIGRATION (PROV-1..6), B-DBRENAME, B-RECOVERY-DRILL
(`rd_tool invalidate-auth`, run.sh:3849), B-RLS-IT access pass with the actual `invalidate.sql` (`MCP_INVALIDATE_SQL_B64`,
run.sh:3461); web `host.test.mjs`, `session.test.mjs`; browser `[E-DOM-1..5]`, `[E-BRD-2]`, `[E-ATTR]`, `[E-ENR-1..5]`,
`[S-LANDMARKS]`, `[S-A11Y]`/`[S-CONTRAST]`.

5.2 **Scenario groups → tests (R/E/N; distinct boundary).**

| ID | Scenario | Test | Boundary |
|---|---|---|---|
| SUP-1 | signup intent only on the exact root origin; `www.B`, tenant, auth, custom, port/scheme variants refused | E `TestLoginStart` | intent × origin kind |
| SUP-2 | T1: new principal → user+mapping+signup row, no session; existing principal → same UUID; cancel/expired/replay/foreign binding/forged state → nothing; no email linking | N `TestCallbackSignup`; R `TestMappingIsByIssuerAndSubject`, `TestCallbackParametersAndTransactions` | non-Organization intent T1 |
| SUP-3 | signup credential is not a session (`/auth/session` 401; `Session` consumers, `enrollment/api`, MCP refuse `Signup`; `signup/api` refuses `Session`); existing user's bindings unchanged | N `TestSignupCredentialBoundary`; E `TestEnrollment` credential-as-session case | second pending-credential scheme |
| SUP-4 | catalog: pw inventory/ACL/body hashes; rt/ow direct `INSERT organizations`/`org_owner` → 42501; DEFINER refuses TENANT shape/foreign user/missing context; `signup_sessions` own-row RLS; audit policy admits only the two captures at depth ≥ 1; `login_transactions` pinned CHECKs; **pw negatives: no EXECUTE on `hostname_state`/`is_organization_member` (still runtime-only), no `login_transactions` privilege (42501), consume body refuses outside LOGIN shape and ignores `p_org` mismatches (zero rows)** | N `TestSignupCatalog`; R `TestRLSCatalog`, `TestAuditCatalog`, `TestDomainsCatalog` (hostname_state ACL unchanged) | new helper/bodies and their non-grants |
| SUP-5 | atomicity: one org + one `org_owner` + T1/T7 audit in one tx; injected failure after org INSERT / after binding / audit refusal ⇒ zero rows; disabled creator ⇒ `user_disabled` | N `TestSignupProvisioning` (`faultDB` pattern, `tenancy/integration_test.go:2003`) | runtime-pool creation path |
| SUP-6 | slug grammar/reserved (`www`, `auth`, CNAME label, configured labels, nested, `B`); concurrent distinct intents one slug ⇒ one winner, loser `slug_taken` zero writes; `hostname_state` resolves the new slug to the new UUID only; custom hosts untouched | N `TestSignupSlug`, `TestSignupSlugRace`; R `TestReservedLabelsFloor`, `TestSlugParity`, `TestDomains` | creation-time uniqueness race |
| SUP-7 | idempotency and recovery (corrections r0-4, r4-2, r5-1), three separate cases: (a) **lost create response** — fault injected after COMMIT before the answer ⇒ `signup_unavailable`, identical retry ⇒ `already_created` + generation 2 reference, exactly one org/owner/`organization.created`/`membership.granted`; (b) **acknowledged create, failed continuation** — 201 received, R consumed, a later hop refused (`continuation_invalid`), identical re-POST with the still-valid credential ⇒ `already_created`, generation 3, old R/N/P/C all refused, still one org/owner/audit pair; **re-mint after a completed step 3** (nonce consumed, `proof_hash` non-null) resets `nonce_consumed_at` and `proof_hash` to NULL atomically under the pw grants/policy — asserted by column reads after the call, and the same re-mint executed as the runtime ⇒ 42501; (c) **credential gone** — expired/abandoned/revoked credential ⇒ `401`, no re-mint, tenant ordinary sign-in still succeeds; changed payload ⇒ `intent_used`; **concurrent identical retries** (two goroutines on the locked row) ⇒ one org, exactly one live generation; disable race (`DisableUser FOR UPDATE` vs `FOR SHARE`) ⇒ one durable outcome; audit row count asserted equal across all retries | N subcases of `TestSignupProvisioning`; R `TestDisableVersusSignInRace` shape | intent-bound idempotency + recoverable continuation under the real cookie lifecycle |
| SUP-8 | continuity matrix (r4-1, r5-2 semantics: every refusal below is "no matching row, nothing consumed" unless a consumption is named) over the four hops, with two simulated browsers (cookie jars) and two root credentials: **captured R** spent by browser 2 ⇒ R consumed (the only consumption), browser 2 gets C₂/N₂ but `prove` without `__Host-signup` ⇒ `signup_context_mismatch`, no row change, nothing completes; browser 1's `continue` finds R spent ⇒ recovery page; browser 1's authenticated "Try again" ⇒ generation g+1, browser 1 completes, one org/owner/audit pair, browser 2's C₂/N₂ dead (bounded availability consequence, accepted); **captured N** presented by browser 1 after browser 2 spent R ⇒ N consumed and P issued into browser 1 (a legitimate match on browser 1's own credential), `complete` in browser 1 lacks C₂ ⇒ `continuation_invalid`, **no row change**, P remains unconsumed until expiry/re-mint, browser 2 never receives P and cannot complete; **captured P** presented without C or with another generation's C ⇒ `continuation_invalid`, no row change; **wrong root context** — another user's credential / another signup row of the same user / revoked / expired credential at `prove` ⇒ `signup_context_mismatch`, no row change, the row's nonce stays unconsumed and the original browser still completes; **wrong tenant browser** (C from a different generation or host) ⇒ refused, no row change; each of R/N/P/C replayed after its legitimate use ⇒ zero rows; expired generation (clock advanced 121 s at each hop) ⇒ refused; wrong target origin/port/other tenant host (runtime-resolved `p_org` mismatch) ⇒ refused at `challenge` and `complete`; host no longer routing ⇒ refused; failure between `complete` and the pinned INSERT ⇒ generation completed, no login row; callback from a browser without the tenant `__Host-login` binding ⇒ `consumeNone` (existing); **account substitution** at the callback ⇒ `signup_user_mismatch`, consumed, no session, no mapping change; **same user, new provider session** in the original browser ⇒ accepted (no SID comparison); **changed slug/host mapping between creation and callback** ⇒ `host_changed`; removed owner binding ⇒ `organization_not_member`; disabled ⇒ `user_disabled`; success ⇒ session bound to the tenant origin and the created Organization only, callback outcome `tenant_created`, `return_to = "/"`; **provider session not live** (fake provider requires interaction, then returns the expected user ⇒ accepted; another user ⇒ `signup_user_mismatch`; cancel ⇒ `login_canceled`, tenant row and owner binding intact, no session) | N `TestSignupContinuation` (signup integration; replaces `TestSignupHandoff`), N `TestCallbackPinned` (identity service_test), E `TestIssuanceGuardFailsClosed`, `TestCallbackContinuity`, `TestCustomDomainRelay` subtests | cross-host automatic continuation with browser (C+P+original credential), user, Organization, exact-host, purpose, one-use and expiry binding |
| SUP-9 | HTTP bounds, CSRF, no-store/nosniff/no-CORS, closed errors, routes pinned, copies exact; limiter exhaustion and key churn for **every** bucket of §4.5: root login start (global), root callback (per-binding and global; churned `X-Login-Binding` headers), create (per-credential and per-user with a re-minted credential), status, prove (per credential), challenge and complete (per tenant origin); `prove` with a `Session` scheme or no credential ⇒ 401; `challenge`/`complete` with any credential header ⇒ ignored (no authority) | N `TestSignupHTTP`, `TestSignupRateLimited`, `TestSignupLimiterChurn`, `TestSignupRoutesPinned`, `TestSignupCopyIsExact`; N `TestRootCallbackRateLimited` (identity http_test) | new surfaces incl. callback and the three continuation hops |
| SUP-10 | logout/logout-all/DisableUser revoke signup credentials; expired rows swept at login start | E `TestEnrollmentPendingCredentialRevocation`, `TestDisableAndEnable` | second pending-credential table |
| SUP-11 | **restore invalidation (correction 3):** the real `invalidate.sql` (via `MCP_INVALIDATE_SQL_B64`) expires live signup rows and handoffs and its postcondition holds; the real `recovery.sh invalidate-auth` command prints `signup_sessions_expired` in `invalidate.tsv` and a restored signup credential/handoff is refused afterwards | E `TestRecoveryPost` (one subcase seeding a signup row before the checkpoint); E B-RECOVERY-DRILL (`rd_tool invalidate-auth`, receipt key asserted in run.sh) | restore cannot revive signup state |
| SUP-12 | **role bootstrap (correction 2):** fresh volume via `10-roles.sh` creates `<db>_provision_writer` with the exact attributes/membership (PROV-1 reference catalog); `provision-roles.sql` as superuser and as a CREATEROLE owner equals it (PROV-1), repeats unchanged (PROV-4), refuses a `provision_writer` with LOGIN/INHERIT (PROV-5 extended); a pre-G062 volume (the pinned v5 script → provisioned) gains exactly the four helpers (PROV-6 delta string); 00018 refuses when the role is missing (`TestSignupCatalog` negative on a pre-provision database); `B-DBRENAME` role list | E `TestRLSMigration` PROV-1..6 (`run.sh` expected strings updated), E B-DBRENAME, N `TestSignupCatalog` precondition case | least-privilege role in both bootstrap paths |
| SUP-13 | audit parity (71 operations; reasons), AUDIT.md rows, logrecord `signup.request` closed | E `TestCatalogParity`, B-AUDIT-CATALOG, `logrecord_test.go` | catalog delta |
| SUP-14 | web unit: `public_entry {signup_origin}` (no `location`; `javascript:`/`http:` origins malformed); pass-through sets for `B` vs `www.B`; cookie kind; client shapes; form errors | E `host.test.mjs`, `session.test.mjs`; N `signup-client.test.mjs`, `signup-forms.test.mjs` | new BFF surface |
| SUP-15 | browser `signup.spec.ts` (`signup-light/dark`, base `https://localhost:543`): `[E-SUP-1]` root and `www.localhost:543` show only Sign Up linking to `https://localhost:543/signup`, no 303, `/account` 404, `www` `/signup` 404, `/api/health` 200; `[E-SUP-2]` fresh emulator user → details → create → `<slug>.localhost:543` → `tenant_created`; `__Host-sid` only on the tenant host; no request outside allowed origins; `[E-SUP-3]` seeded org-a member creates a second tenant, org-a roles unchanged; `[E-SUP-4]` slug taken/reserved/invalid field errors, keyboard, axe light/dark; `[E-SUP-5]` (two browser contexts, real navigated origins) the three automatic hops observed as 303s with `no-store`/`no-referrer` and the cookies `__Host-sct` (tenant, 120 s) and `__Host-signup` (root, retained) exactly where expected; captured `continue?ref=` URL opened in context 2 → context 2 ends on root `/signup/details?status=continue_failed` without `__Host-signup` → `/signup`, no session anywhere, context 1's continuation refused until "Try again" re-mints and completes; captured `prove?n=` URL opened in context 1 after context 2 spent R → context 1 lands on tenant `complete` → `continue_failed`, context 2 cannot complete; captured `complete?p=` in context 2 → refused; reference/nonce/proof replay → refused; `/signup/details` without cookie → `/signup`; provider cancel at the final hop → tenant `/sign-in?status=canceled`, tenant persists; second context signing in as **another** seeded user at the final hop after a legitimate continuation (jar copy of `__Host-login` only) → `signup_user_mismatch`; original browser re-authenticated at the emulator (new provider session) before the final hop → accepted; acknowledged create then forced hop failure → `/signup/details?status=continue_failed` → "Try again" → `tenant_created` with one org; **provider session not live**: emulator session ended before the continuation → the hosted prompt appears on the way to the tenant host (no Continue button anywhere), authenticate as the same user → `tenant_created`; cancel at the prompt → tenant `/sign-in?status=canceled` while `<slug>.localhost:543` already routes and a later ordinary sign-in there succeeds; identical create re-posted after a dropped response → continues; **wildcard routing**: the new `<slug>.localhost:543` is reachable immediately through the existing `*.localhost` proxy with no DNS/CNAME step (fixture-proven only); `[E-SUP-6]` after completion `/groups`, `/account` usable without Group/profile/billing; G018 `groups/new` loads | N | real-origin HTTPS proof of AC1/AC7/AC8 |
| SUP-16 | `[E-DOM-4]` rewritten: platform domain and `www` serve the landing, protocol paths exempt, no marketing hop; `[E-DOM-2]`/`[E-DOM-3]` unchanged | E `identity.spec.ts` | supersession |
| SUP-17 | rename variant: landing/signup origin follow configuration | E B-BROWSER-RENAME + N one `@identity` assertion if the shell projects reach the https root; otherwise covered by SUP-15 + `TestSignupSlug` config cases (recorded) | §50 renameability |

5.3 **Batches** (working directory `core/`; fetch/rebase/reconcile + Docker Desktop mount check before each; private
per-run DB/browser stacks under the two-run pool; one image build at a time):
- **b0 static:** core-check image `gofmt`/`go vet` + `go test ./internal/signup/... ./internal/identity/ ./internal/hostname/ ./internal/config/ ./internal/audit/ ./internal/domains/...`;
  `npm run test:unit`; `static_web_boundary`; B-ENV-EXAMPLE (re-pinned), B-LITERALS, B-IDENTITY-GUARDS, B-SHELL (10-roles.sh).
- **b1 roles/migration/catalog (private DB):** `rls_go ./internal/migrate/ '^TestSignupCatalog$'`, `'^TestRLSCatalog$'`,
  `'^TestAuditCatalog$'` + AUDIT.md diff, `'^TestEnrollmentCatalog$'`, `'^TestMembershipCatalog$'`, `'^TestMigrationLifecycle$'`;
  `'^TestRLSMigration$'` with PROV-1..6 (`rls_provision`), B-DBRENAME; B-MIGRATE-SET/CLI/CATALOG at 18.
- **b2 provisioning/identity/recovery (private DB):** `go test -tags integration -run '^(TestSignupProvisioning|TestSignupSlugRace|TestSignupContinuation)$' ./internal/signup/`;
  `-run '^(TestCallbackSignup|TestCallbackPinned|TestLoginStart|TestRootCallbackRateLimited|TestEnrollmentPendingCredentialRevocation|TestDisableVersusSignInRace|TestRLSIdentityPool|TestRLSFirstSignInRace)$' ./internal/identity/`;
  `-run '^(TestEnrollment|TestEnrollmentAdmission)$' ./internal/tenancy/access/`; `-run '^TestRecoveryPost$' ./internal/tenancy/access/`
  (`MCP_INVALIDATE_SQL_B64` = the modified `invalidate.sql`); B-RECOVERY-DRILL (`rd_tool invalidate-auth`).
- **b3 browser (private identity overlay):** the `identity:browser` line with `--project=signup-light --project=signup-dark --project=domain-platform`;
  then unchanged `shell-light/dark` landmarks/a11y on the loopback stack.
- **Final local coverage before MR:** affected `migrations`/`ops`/`root` rows (B-SIGNUP-*, B-RLS-CATALOG/MIGRATION,
  B-AUDIT-*, B-MIGRATE-*, B-RECOVERY-DRILL, B-DBRENAME, B-IDENTITY-IT/E2E/PRIVACY/PROBES/CONFIG, B-ENROLLMENT-*,
  B-CONFIGURE) + `lint`/`unit`. Changed paths are foundation ⇒ MR fast profile runs and the post-merge coordinator-owned
  exact-merged-SHA **full** local qualification is required.

5.4 **Change-impact inventory** — r0 list plus: `run.sh:2978` PROV-1 admin-residual string, `:3001` PROV-6 delta string,
`:3024` B-RLS-MIGRATION PASS text, `:4732/:4737` B-DBRENAME role list/text, `:1880` `core_env`; `rls_integration_test.go:113-119`
`rlsLabel` (`pw`), `:105` role list; `membership_integration_test.go:89`; `provision-roles.sql` role allow-lists
(`:72-94`, `:114-116`, `:143-148`); `invalidate.sql` postcondition; `RECOVERY.md:22,139-145`; `DATABASE-ISOLATION.md:59-77,261-290`;
`.env.example` block SHAs; `playwright.config.ts:97`; `identity.yml:37,59,96`; `proxy.conf:77-86`.

5.5 **Manual feedback and owner input.** **M1 (required, before final coverage):** owner walk-through on the local
overlay — `https://localhost:543` and `https://www.localhost:543` → Sign Up → emulator → details → new tenant host —
plus the existing-identity second tenant, the slug-taken error and the not-live-provider prompt (sign out at the
emulator first); light/dark/keyboard. Readiness: b0–b2 green, b3 focused pass. **Live:** registering
`https://B/auth/callback` and `https://B/sign-in` in the WorkOS application is a G057 operator action (apex has no
wildcard match); G062 claims no live PASS and adds no staging test. G004 L4 is not invoked.

5.6 **Fixture-proven here vs G057-qualified (do not conflate).** *Fixture-proven in G062 (local overlay: emulator,
`*.localhost:543` nginx proxy, DNS/TLS fixture adapters):* every SQL/catalog/atomicity/race/idempotency/binding proof;
the automatic continuation's control flow, pins and refusals; slug-to-UUID routing on the fixture wildcard; the
silent-return behaviour of the **emulator** while its session is live; the not-live prompt path with the emulator.
*G057-qualified (required before G021; no live guarantee is claimed by G062 fixtures):* real WorkOS redirect behaviour
for the exact apex callback `https://B/auth/callback` and the `*.B` wildcard tenant callbacks; real provider-session
liveness and silent code return (whether the continuation completes without a prompt in production); real wildcard
DNS/ingress/TLS resolution of a newly created `<slug>.B` with no per-tenant record; production edge callback-query
redaction. G057's existing "G062 platform-root signup qualification input" clause owns these, reusing G062 local
evidence and adding only the affected real-provider/edge proof.

## 6. Delegation

One implementation writer (Opus 5.5 high) owns every allow-listed file and integration. One bounded read-only security
review (Fable 5.1 high) after Batch 2, before Batch 3: §3 bodies/policies/role, the signup credential boundary, the
pinned-callback binding and automatic continuation (incl. the not-live-provider path), lost-response/re-mint contract,
limiter keys, recovery fold, log/capture privacy. Findings return to the writer in the same attempt. No other subagents; shares B5.

## 7. Sequence and slicing (one attempt)

1. R0′ rebinding (D-4 closed: automatic continuation). 2. **Slice 1 (risk first):** `10-roles.sh`/`provision-roles.sql`
+ harness PROV/DBRENAME strings; 00018; `signup` store/provisioning; identity T1, pinned callback, handoff consumer
(`LoginPinned`), revocation/sweep; `invalidate.sql` fold + `TestRecoveryPost` subcase; `TestSignupCatalog`,
`TestSignupProvisioning` (incl. SUP-7 (a)/(b)/(c) and concurrent retries), `TestSignupSlug`/`TestSignupSlugRace` (format,
reserved, platform-wide uniqueness), `TestSignupContinuation` (captured R/N/P, wrong root context, wrong tenant browser,
replay/expiry/changed host/user substitution/not-live provider), `TestCallbackSignup`, `TestCallbackPinned`
(`tenant_created` outcome), `TestReturnPathAllowList` extension, `TestRootCallbackRateLimited`; b0–b2. 3. Security review;
fixes. 4. **Slice 2:** `signup/api` + limiter copy + HTTP tests; `hosts/resolve`/proxy/host.ts/config removal of
`MARKETING_SITE_URL` + `.env.example` re-pin; landing/signup pages; overlay/playwright changes; E-DOM-4 rewrite +
`signup.spec.ts`; b0, b3. 5. M1; fixes; affected reruns. 6. Docs; PG-009; Result; final local coverage; fetch/rebase; push;
MR (fast + Duo + Codex source review) → conditional merge → merged-SHA full local → cleanup receipt → final ACCEPT.

## 8. Acceptance-criteria mapping

Each row marks **F** = fixture-proven in G062 (local overlay/emulator/fixtures) and **Q** = G057-qualified (real
provider/edge; required before G021; not claimed here).

| AC | Evidence |
|---|---|
| AC1 | F: SUP-15 `[E-SUP-1]` (root and `www` landing, canonical Sign Up, no marketing hop), SUP-16, SUP-14, SUP-17; seeded `org-a.localhost:543` in `[E-SUP-3]`/`[E-BRD-2]`. Q: real apex/`www` edge serving of the landing and the apex callback registration |
| AC2 | F: SUP-2; SUP-15 `[E-SUP-2]`/`[E-SUP-3]` (fresh emulator user via `emulatorAdmin().create`; seeded user); optional-MFA/assurance code untouched (`TestCallbackRefusals`). Q: real WorkOS identity creation/sign-in on the apex callback |
| AC3 | F: SUP-3; `[E-SUP-3]` prior roles unchanged; B-ENROLLMENT-IT reuse. Q: none (pure application boundary) |
| AC4 | F: SUP-4, SUP-5, SUP-12 (role in both bootstrap paths; fresh/upgrade PROV proofs), SUP-13; B-MIGRATE-IT up/down; SUP-11 restore disposition. Q: none |
| AC5 | F: SUP-6 (format, reserved floor, platform-wide uniqueness race), `[E-SUP-4]`, slug-to-UUID routing on the fixture wildcard with no DNS/CNAME step; custom hosts untouched (`TestDomains`). Q: real wildcard DNS/ingress/TLS resolution of a new `<slug>.B` |
| AC6 | F: SUP-7 (same-intent duplicates, changed payload, lost response, concurrent retries, disable race), `[E-SUP-5]` cancel and re-post. Q: none |
| AC7 | F: SUP-8 (`TestSignupContinuation`, `TestCallbackPinned`: browser continuity through C + P + the original `__Host-signup` + tenant `__Host-login`; user, Organization, exact host, purpose, one use, expiry; captured R/N/P, wrong root context, wrong tenant browser, replay, expiry, changed host, account substitution, not-live-provider prompt/cancel), SUP-7 (b) recovery, `[E-SUP-2]`/`[E-SUP-5]` on real local origins with two browser contexts, `TestReturnPathAllowList`/`TestSignupCatalog` return-path sets, `TestCustomDomainRelay` reuse. Q: real WorkOS silent code return while the provider session is live (whether the final hop completes without a prompt), real wildcard tenant callbacks, production edge callback-query redaction. No residual is relied on for compliance |
| AC8 | F: `[E-SUP-6]`, `TestAuthority`/`TestSignupCatalog` (override default false), M1 recorded honestly. Q: none |
| AC9 | F: SUP-9 (every bucket incl. callback and the three hops), SUP-10, SUP-11, B-IDENTITY-PRIVACY/PROBES (root callback and the `continue`/`prove`/`complete` URLs in the redaction probe list: `ref`, `n`, `p` never logged), SUP-5 rollback. Q: production edge log redaction |
| AC10 | Result mapping with F/Q labels per claim; MR fast + Duo + Codex; merged-SHA full local; cleanup receipt; §5.6 G057 reconciliation inputs |

## 9. Adopter documentation (supersession notes)

As r0, plus: DATABASE-ISOLATION.md roles table (fourth helper) and "Provisioning" (four helpers; pre-G062 volumes);
MIGRATIONS.md 00018 row and `provision-roles.sql` precondition; RECOVERY.md `invalidate-auth` counts
(`signup_sessions_expired`) and the inventory row for `signup_sessions`; DOMAINS.md host-model row (`B`/`www.B` →
landing; `MARKETING_SITE_URL` removed) and test overlay; IDENTITY.md "Platform root signup" (cookie/credential rows,
pinned callback, limiter statement); SECURITY.md credential-inventory rows (`signup_sessions.credential_hash`,
`ref_hash`/`challenge_hash`/`nonce_hash`/`proof_hash`, each SHA-256 only) and the G062 note → implemented status.
Historical G016/G017 evidence is not rewritten.

## 10. Decisions, risks and escalations

- **D-1 Helper role** `<db>_provision_writer` (new NOLOGIN helper; provisioning-owned, migration-granted; two bodies
  (`signup_provision_organization`, `signup_challenge_issue`, `signup_challenge_complete`); no reader EXECUTE beyond the parsers; no
  `login_transactions` privilege). *Security review.*
- **D-2 `events_insert_provision`** (pw-scoped audit INSERT policy). *Security review.*
- **D-3** Separate `signup_sessions` table (Spec R2). Unchanged.
- **D-4 Handoff mechanism — CLOSED by the owner (automatic continuation; source: owner via Codex, 2026-10-20; record
  at the top of this Plan).** Option B (r1) and Option C (manual Continue / ordinary second sign-in, r2) are removed as
  selectable designs. The mechanism is §1.4/§3.4 (r4-1): initial reference → tenant challenge (`__Host-sct`) → root
  proof with the **original `__Host-signup`** → tenant completion (C + P) → pinned G016 login → pinned callback. All
  hops are automatic 303s; no Continue button; no provider-SID assumption; the final WorkOS hop completes silently while
  the provider session is live (G057-qualified) and otherwise shows the provider's own prompt. **No residual is relied
  on:** a captured initial reference, nonce or proof, a wrong root signup context and a wrong tenant browser cannot
  spend or complete the intended browser's continuation (SUP-8 matrix, `[E-SUP-5]`). What genuinely remains is only
  what the existing G016 login already accepts: a party holding the user's WorkOS credentials can sign in on the tenant
  host through the ordinary login (not through this continuation), and a same-browser re-authentication with a new
  provider session is accepted by design. Spec R4/AC7 are met without relaxation; Codex owns the technical review.
- **D-5** Root callback on the root origin itself (no relay); exact apex provider entries are a G057 operator input;
  local `--redirect-hosts` + `localhost`.
- **D-6 CLOSED (mechanical):** `MARKETING_SITE_URL` has no consumer beyond the root redirect (§4.1 inventory) ⇒ removed
  from config/validation/example/overlay/probe env/tests; `location` contract dropped; `.env.example` domain block
  re-pinned. No owner decision remains.
- **D-7** Audit catalog delta (§3.6). *Review.*
- **D-8** `signup` label not reserved (path, not label). Unchanged.
- **D-9** `core/internal/signup` with pinned copies; no tenancy import. Unchanged.
- **D-10** PG-009 at execution: unique-index INSERT on `organizations(slug)`; one signup row per root authentication
  (swept with login transactions); handoff probe by unique hash; pinned login row adds two nullable columns. Hypothesis:
  negligible; no measurement claimed.
- **D-11 CLOSED:** restore invalidation folded into the invoked `invalidate.sql` (expire signup rows and handoffs;
  count; postcondition); `recovery.sh` untouched; proven by the real command (SUP-11). Retention unaffected.
- **D-12** Rate limits as §4.5, including the root callback (per-binding + global) and the three continuation hops —
  honest per-process statement.
- **D-14 (r4-3) Completion status mechanism:** a closed callback **outcome** (`tenant_created`) rather than a stored
  return path with a query, so both return-path closed sets stay finite (`/signup/details` added to both; nothing
  else). *Review.*
- **D-15 (r4-1) `prove` without a CSRF body token:** the step is GET-initiated by a 303 and its only effect is to spend
  the presenting credential's **own** current nonce; a forced navigation by an attacker (link/img) carrying a nonce of
  the victim's own row can at most consume that generation's nonce (recoverable by "Try again") and sends P into the
  victim's browser, never to the attacker; a nonce of any other row matches nothing and consumes nothing (r5-2).
  Recorded as a bounded, reviewed exception to the "every mutation carries `csrf_token`" pattern. *Review.*
- **D-13** Feasibility: one B5 attempt; forecast 2–3 attempts (r1 adds the role bootstrap, pinned callback and recovery
  fold — each a template reuse). B7 is not requested; no split recommended.
- **www resolved (correction 6):** `www.B` serves the landing with Sign Up canonicalized to `${signup_origin}/signup`;
  signup intent, `/signup/*` and `/auth/callback` exist only on `B`. No ambiguity remains.
- **Stops honored:** no privileged-pool exposure; no runtime Organization INSERT; no new credential/pool/issuer/vendor;
  no custom-host/MCP change; no G016/G057 deferral; anything beyond §3 returns to the owner.

## 11. Budget, pins and scope

**B5, checkpoint after attempt 3; 0 attempts spent.** Pins at R0′: `WEB_ROUTES` +8; `access.denied`
operations 65 → 73; 00018 DEFINER allowance 4 (3 pw, 1 ar); cw 7; `app_objects_beyond_metadata` 71 → 76 (verify);
`login_transactions_return_path_check` four members; `identity.returnPaths` three members;
migrations 18; `productionFiles` 19; PROV strings with four helpers; browser matrix `signup-light/dark|E-SUP-1 … E-SUP-6
S-ERRORS`, `domain-platform|E-DOM-1 E-DOM-3 E-DOM-4 S-ERRORS`. Scope: the 40-pattern allow-list (r0 38 + 2 patterns:
`infra/postgres/init/10-roles.sh`, `infra/postgres/provision-roles.sql`; widened existing patterns admit
`core/internal/migrate/membership_integration_test.go` (role list), `core/internal/identity/limiter_copy(_test).go`,
`core/internal/config/domains_config_test.go`, `core/internal/domains/api/http_test.go`; `infra/recovery/sql/invalidate-signup.sql`
replaced in place by the invoked `invalidate.sql`). Model/effort:
implementation Opus 5.5 high; planning/review/corrections Fable 5.1 high (runtime effort unverifiable from inside the agent).

## 12. Revision history

- r0, 2026-10-20 (commit `3ec3732a`): provisional Plan; three bounded Explore inventories; no probe/build/test/Docker/
  DB/browser/provider/infrastructure action.
- r1, 2026-10-20 (commit `4c0c165d`): Codex CORRECTION LOOP (7 items) applied; same planner, same branch; read-only
  inspection of `infra/postgres/init/10-roles.sh`, `infra/postgres/provision-roles.sql`, `infra/recovery/recovery.sh:713-720`,
  `infra/recovery/sql/invalidate.sql`, `run.sh` PROV/DBRENAME/core_env blocks and every `MARKETING_SITE_URL` consumer;
  no execution action. Not self-accepted.
- r2, 2026-10-20: second Codex CORRECTION LOOP (4 items) applied; same planner, same branch, base kept at `47df73f5`
  (staging `5b813e32` re-verified: no in-scope input change beyond a SECURITY.md skills link). Read-only inspection of
  `00006:469-473` (`login_transactions` policies) and `00011:774-775` (`hostname_state`/`is_organization_member` grants).
  No execution action. Not self-accepted.
- r3, 2026-10-20 (on r2 commit `e0ae0b1c`): owner D-4 decision (automatic continuation; via Codex) folded in; Option C
  removed; not-live-provider and recovery paths specified; wildcard-routing statement (§1.5); fixture-proven vs
  G057-qualified labels (§5.6, §8); selectors/sequence updated. r2 technical resolutions kept; r2 review still pending.
  No inspection beyond the Plan's own inputs; no execution action. Not self-accepted.
- r4, 2026-10-20 (on r3 commit `4a73fff0`): Codex r3 CORRECTION LOOP (3 items; `G062-plan-codex-review.md` at
  staging `0c8f3f83`) applied: challenge/proof/completion browser continuity (no residual relied on), cookie-lifecycle-
  consistent recovery with three distinct cases, both return-path closed sets reconciled with a closed completion
  outcome. Read-only inspection of `core/migrations/00002_identity.sql:84`, `00016:172-173,1548-1549` and
  `core/internal/identity/allowlist.go:27-34`. Base kept at `47df73f5`. No execution action. Not self-accepted.
- r5, 2026-10-20 (on r4 commit `f6ac134d`): Codex r4 CORRECTION LOOP (2 items; review at staging `71ce8d28`)
  applied: pw column-level UPDATE on `nonce_consumed_at`/`proof_hash` for the re-mint reset with a WITH CHECK that
  confines non-null proof issuance to the runtime root path, plus matching catalog/privilege and repeat-create proofs;
  rejection prose and SUP-8/D-15 aligned to the SQL (no match ⇒ nothing consumed), the captured-R bounded availability
  consequence stated as accepted. Reconciliation: staging `71ce8d28` in-scope delta unchanged; PROGRESS/ROADMAP
  re-hashed; G055 merged and finally ACCEPTed (bundle `skills/restruct-sdlc`, aliases), standalone paths retired;
  certifi maintenance and `0.2.0-rc.1` recorded as non-dependencies. No protocol change. No execution action. Not self-accepted.

## 13. r1 correction record — finding → change

| # | Finding | Change |
|---|---|---|
| 1 | handoff not browser/user/org/host-bound through callback/issuance | §1.4, §3.1 pinned login columns, §3.4 `signup_handoff_consume` inserts the pinned row, §4.2 new denials, SUP-8/`TestCallbackPinned`/`[E-SUP-5]`; D-4 kept open with proof for both options |
| 2 | provisioning role absent from bootstrap | §2 roles row, §3.3 (10-roles.sh + provision-roles.sql, migration precondition), SUP-12, PROV/DBRENAME strings, docs; allow-list 38 → 40 |
| 3 | restore invalidation not wired | §2 recovery row: fold into `invalidate.sql`; `recovery.sh` unchanged; SUP-11 real command + `TestRecoveryPost`; D-11 closed |
| 4 | lost first response unrecoverable | §3.4 step 3 re-mint with `handoff_generation`; §4.2 `already_created` returns a fresh reference; SUP-7 commit-success/response-loss and concurrent retries |
| 5 | callback limiter missing | §4.5 (no existing callback limiter identified; structural guard stated; root-callback per-binding + global buckets); `TestRootCallbackRateLimited`; SUP-9 |
| 6 | `www` ambiguity | §1.1, §4.1 `signup_origin`, §4.4, `[E-SUP-1]`; §10 resolved |
| 7 | D-6 owner-level | §4.1 consumer inventory; removal everywhere; D-6 closed |

### r2 correction record — finding → change

| # | Finding | Change |
|---|---|---|
| r2-1 | `ProviderSessionID` not established as browser-continuity proof | SID comparison and `expected_provider_session_id` withdrawn everywhere (§1.4, §3.1, §3.6, §4.2, SUP-8, `[E-SUP-5]`); binding restated honestly: browser binding starts at consumption (tenant `__Host-login`), user pin, Organization/host pin, purpose/one-use/expiry; same-user/different-browser recorded as the disclosed residual, same-user/new-provider-session accepted; no provider contract is relied on |
| r2-2 | Option B could not locate/authorize the signup row | Option B withdrawn; Option C (manual, no reference, ordinary G016 login) specified with its R4/AC7 reading flagged for Codex; D-4 remains the owner's decision |
| r2-3 | pw cannot call `app.hostname_state` | consume body no longer calls any reader; the runtime resolves the host (`tenantOrigin`) and passes `p_org`; membership via runtime `is_organization_member`; §3.3 non-grants pinned as catalog negatives (SUP-4) |
| r2-4 | pinned-row INSERT ACL contradictory | runtime inserts the pinned row under the unchanged `login_transactions_insert` policy (00006:470) in the same LOGIN transaction path as `insertLogin`; pw has no `login_transactions` privilege; CHECKs bound the shape; negatives in SUP-4 |

### r3 owner-decision record — item → change

| # | Item | Change |
|---|---|---|
| r3-1 | D-4 closed: automatic continuation (owner via Codex, 2026-10-20) | decision recorded verbatim at the top; §1.4 restated as the automatic flow with the honest binding, the not-live-provider path (prompt, mismatch, cancel; tenant persists; closed status texts) and lost-response/failed-continuation recovery; Option C removed from §1/§5/§8/§10/§11 |
| r3-2 | platform tenants via preconfigured wildcard DNS/ingress/TLS | §1.5: creation adds only the slug-to-UUID row; no per-tenant CNAME, no propagation wait; custom-domain flow unaffected |
| r3-3 | G057 qualifies real provider/edge behaviour | §5.6 fixture-proven vs G057-qualified inventory; §8 F/Q labels per AC; AC10 carries them into the Result |
| r3-4 | mapping/sequence/selectors | §7 step 2 names the slug (format/reserved/uniqueness), handoff (replay/expiry/changed host/user substitution/not-live) and callback selectors; SUP-8/SUP-15 extended; allow-list pattern set unchanged (40) |

### r4 correction record — finding → change

| # | Finding | Change |
|---|---|---|
| r4-1 | disclosed same-user/different-browser residual does not meet R4/AC7; no relaxation | §1.4 four-hop design: initial reference → tenant challenge cookie `__Host-sct` → root proof requiring the exact original `__Host-signup` → tenant completion (C + P) → pinned login; §3.2 generation state (R/N/P/C hashes, one use each), §3.4 `signup_challenge_issue`/`signup_challenge_complete` (pw) and the runtime own-row `prove` UPDATE, §3.5 `signup_sessions_prove` policy and pw column split, §4.2–§4.5 routes/cookies/limits, SUP-8 `TestSignupContinuation` matrix (captured R/N/P, wrong root context, wrong tenant browser, replay/expiry/host/user), `[E-SUP-5]` two-context browser proof; residual language removed from §1.4/§8/§10 and reduced in D-4 to what the ordinary G016 login already accepts; D-15 records the `prove` CSRF exception |
| r4-2 | recovery contradicted the cookie lifecycle | §4.4 keeps `__Host-signup` for its bounded 30 min (cleared on abandon/401/expiry only); §1.4 recovery (a) lost response, (b) acknowledged create + failed continuation (root `?status=continue_failed` + "Try again" re-mint, failure path only), (c) credential gone; SUP-7 tests (a)/(b)/(c) separately with audit-row equality; `[E-SUP-5]` exercises (b) |
| r4-3 | return-path closed sets unreconciled | §3.1: DB CHECK gains `/signup/details` (up/down/postcondition), Go `returnPaths` gains `/signup/details`; completion status via the closed callback outcome `tenant_created` with `return_path = '/'` (D-14); `TestSignupCatalog`/`TestReturnPathAllowList`/`TestLoginStart`/`TestCallbackPinned` positives and malicious-return negatives (external, `//`, backslash, encoded, traversal, trailing slash, query/fragment, case) |

### r5 correction record — finding → change

| # | Finding | Change |
|---|---|---|
| r5-1 | pw lacked UPDATE on the two fields the re-mint clears | §3.3: exact 14-column pw UPDATE grant now includes `nonce_consumed_at`, `proof_hash`; §3.5: `signup_sessions_provision_update` WITH CHECK `(proof_hash IS NULL OR completed_at IS NOT NULL)` keeps non-null proof issuance on the runtime root path; catalog negatives (column-privilege set, pw non-null proof ⇒ 42501, runtime limited to the two prove columns); SUP-7 (b) re-mint-after-prove reset proof. No broader grant |
| r5-2 | prose said foreign proof / wrong challenge consumes nonce/proof | §1.4 step 3 and browser bullet, §3.4 step-3 paragraph, SUP-8 and D-15 restated: a non-matching `prove`/`complete` changes no row and consumes nothing; the only consumptions are the legitimate R/N/P/C uses; captured-R transient generation loss recorded as the accepted bounded availability consequence with durable credential/intent intact and authenticated re-mint recovery; SQL predicates unchanged |
| r5-rec | current-source reconciliation | §0: staging `71ce8d28` in-scope delta unchanged; PROGRESS/ROADMAP re-hashed (inputs JSON r5); G055 merged + FINAL ACCEPT, bundle/aliases named, standalone paths retired; certifi/`0.2.0-rc.1` not dependencies; D-4 closed and B5 kept |
