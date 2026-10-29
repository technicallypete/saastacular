# Identity and sessions

Global identity and application sessions (migration 00002). WorkOS AuthKit authenticates people; **core** owns everything the
application decides about them: the mapping from a provider identity to a stable application user, sessions, CSRF,
sign-out and every authorization decision. **web** is a thin browser layer (BFF): it holds the session cookie, renders
the sign-in and account pages and forwards to core. Organizations, groups and roles are later foundation stages; nothing
here grants tenant scope.

## What is decided where

| Concern | Owner | Notes |
|---|---|---|
| Who the person is | WorkOS AuthKit (hosted sign-in, MFA, SSO) | Reached only through `core/internal/identity/workos`, the one package that imports the WorkOS SDK and the token libraries. |
| Which application user that is | core | `(issuer, subject)` → one `users.id` (UUID), created once, never by email. Display name is refreshed from the provider; email is never stored. |
| Whether a request is signed in | core | Opaque 256-bit session credential; only its SHA-256 is stored. Idle 30 min and absolute 8 h by default. |
| CSRF | core | A per-session synchronizer token, rendered only into the user's own account forms. |
| Cookies, pages, forms | web | `__Host-sid` (session) and `__Host-login` (login binding): `HttpOnly; Secure; Path=/; SameSite=Lax`, no `Domain`. |

## Multi-factor assurance

Every session's multi-factor assurance is **`unknown`**. The account page says "Multi-factor status: not verified by
this application", and `/api/account` returns `"mfa_assurance": "unknown"`. Nothing upgrades it: not the WorkOS
environment's MFA setting, the token's authentication method or any claim, a browser parameter or header, or
configuration (there is no MFA variable). The database enforces it (`CHECK (mfa_assurance = 'unknown')`), so a future
proof source needs a reviewed migration. `identity.RequireVerifiedMFA` is the fail-closed guard for anything that
requires verified MFA: with no reviewed proof source it always refuses (`mfa_unverified`, 403). No identity route uses it.

Whether WorkOS enforces MFA for a tested environment is qualification evidence recorded with its configuration and time
(L3 S-MFA, MQ-1/MQ-2), never session data.

## Configuration (core; web reads only `IDENTITY_COOKIE_PROFILE`)

| Variable | Rule |
|---|---|
| `IDENTITY_PROVIDER` | `workos` (or unset). |
| `IDENTITY_WORKOS_CLIENT_ID`, `IDENTITY_WORKOS_ISSUER`, and one of `IDENTITY_WORKOS_API_KEY` / `IDENTITY_WORKOS_API_KEY_FILE` | All three set, or none. None: identity is **unconfigured**. Sign-in reports that it is unavailable and sessions still resolve, but production refuses to start. Some but not all: `config_invalid` at startup, naming variables only. The API key is a secret and stays in core. |
| `IDENTITY_WORKOS_ISSUER` | The **exact** `iss` of a real access token from the environment. Copy it from a verified token (as the L3 contract test does); never guess it. For WorkOS staging it has the shape `https://api.workos.com/user_management/<client_…>`, where the client segment is the **environment's** client rather than the configured application client. Trailing-slash variants are refused. |
| `IDENTITY_WORKOS_API_BASE_URL` | Default `https://api.workos.com`. Any other value only with `APP_ENV=test` (the emulator). |
| `IDENTITY_PUBLIC_ORIGINS` | Exact `https` origins. The callback is `<origin>/auth/callback`; the sign-out return is exactly `<origin>/sign-in`, without a query. Both must be registered in WorkOS for each origin. Tenant hosts are not listed here: when the shared auth host is configured (`PLATFORM_AUTH_HOST`), its origin must be the only entry ([DOMAINS.md](DOMAINS.md)). |
| `IDENTITY_SESSION_IDLE`, `IDENTITY_SESSION_ABSOLUTE` | Defaults 30m and 8h. Production requires each to lie in [5m, 24h]. |
| `IDENTITY_JWT_ALGORITHMS` | Default `RS256`; asymmetric algorithms only. |
| `IDENTITY_MAX_AUTH_AGE` | Optional freshness: sends `max_age` and refuses a token whose `auth_time` is older (`stale_authentication`). |
| `IDENTITY_COOKIE_PROFILE` | `https` (`__Host-` cookies) or `loopback-http` (plain-HTTP loopback development). Production refuses `loopback-http` in both core and web. When unset, web uses `https` in production and `loopback-http` elsewhere; core uses `https`. **Configured HTTPS use outside production needs `IDENTITY_COOKIE_PROFILE=https` set explicitly for web**, as `infra/compose/identity.yml` does. An http flow needs an explicit non-production `loopback-http` in core too, and core then accepts http origins only for `localhost` and `127.0.0.1`. |

**Namespace continuity.** The first sign-in binds the deployment to one `(issuer, client_id)` pair in
`identity_provider_binding`. A different configured pair (an issuer change, or a client that belongs to another WorkOS
environment) makes login and callback answer `identity_namespace_mismatch` (503). No mapping is created, no session
issued and no transaction consumed. Existing sessions and user IDs are untouched. There is never email linking.
Re-binding needs a separately reviewed migration; none is provided.

## Browser contract (web origin)

| Route | Outcomes |
|---|---|
| `GET /sign-in`, `GET /account` | Server-rendered, `no-store`. Anonymous `/account` redirects with 307 (a page redirect; GET is preserved) to the fixed `/sign-in?return_to=/account`. Every form, callback and sign-out redirect is 303. Query values are never reflected: `status` selects fixed text from a closed map. |
| `POST /auth/login` | 303 to the hosted WorkOS page (with the login-binding cookie); 303 `/account?status=already_signed_in`; unconfigured, unavailable or namespace mismatch → 303 `/sign-in?status=unavailable`; 403 foreign origin; 405; 415 for anything but a form post. |
| `GET /auth/callback` | 303 `<return path>?status=signed_in` or `session_rotated` (session cookie set); 303 `/account?status=already_signed_in` or `identity_conflict`; 303 `/sign-in?status=canceled\|expired\|stale\|unavailable\|failed`; 400 for an ambiguous query. |
| `POST /auth/logout`, `/auth/logout-all`, `/auth/provider-logout` | See sign-out below; 400, 401, 403, 405, 415 and 503 as minimal `no-store` pages. |
| `GET /auth/logout-return` | The provider sign-out's notification step (below): a minimal `no-store` page with the recorded outcome, clearing the `__Host-logout` cookie in the same response; 503 when the outcome cannot be read; 405 for other methods. It never changes a session. |
| `GET /api/account` | 200 `{id, display_name, mfa_assurance: "unknown", auth_time, session_expires_at}`; 401 `{"error":"unauthenticated"}` whether or not a provider is configured; 503 `{"error":"unavailable"}` when sessions cannot be resolved. Never carries the CSRF token. |

Form routes check, in order: method (405), `application/x-www-form-urlencoded` (415), an `Origin` exactly equal to the
request's own origin (403), then core checks the session (401 where one is needed) and the CSRF token (403). The
origin comes from the `Host` header and the cookie profile's scheme. Forwarding headers are never read.

## Core contract (`/internal/v1/auth/*`, internal network only)

| Route | Input | Success | Errors |
|---|---|---|---|
| `POST /login` | `{origin, return_to?}`, optional session credential and login binding | `{authorization_url, binding: {value, max_age}}` | `identity_unconfigured` 503, `origin_rejected` 403, `already_authenticated` 409, `identity_namespace_mismatch`/`identity_unavailable` 503 |
| `POST /callback` | `{origin, raw_query}`, login binding, optional credential | `{outcome: signed_in\|session_rotated, session: {value, max_age}, return_to}` or `{outcome: already_signed_in, return_to}` | `ambiguous_parameter` 400, `transaction_invalid`/`transaction_expired`/`login_canceled`/`code_rejected` 401, `identity_conflict` 409, `user_disabled`/`impersonation_refused`/`stale_authentication` 403, `provider_error` 502, `provider_timeout` 504, `identity_unconfigured`/`identity_namespace_mismatch`/`identity_unavailable` 503 |
| `GET /session?origin=` | credential and exactly one `origin` (required) | `{user_id, display_name, mfa_assurance, auth_time, expires_at, csrf_token, host}` | `invalid_request` 400 (no `origin`), `ambiguous_parameter` 400 (repeated), `unauthenticated` 401 (a dead credential, or a session not bound to that origin and the Organization it routes to now), `unknown_host` 404, `identity_unavailable` 503 |
| `POST /logout` | `{scope: current\|all\|provider, csrf_token, origin}`, credential | `{outcome, provider_logout_url?, return_to?, logout_ref?}`: for scope `provider`, `return_to` is the fixed `/sign-in` and the one-use `logout_ref` travels only in its own field | `invalid_request` 400, `origin_rejected` 403 (a missing or malformed `origin`, every scope: decided on its structure alone, before the session and without a routing lookup), `unauthenticated` 401 (all, provider), `csrf_failed` 403, `identity_unavailable` 503 |
| `POST /provider-logout/return` | `{logout_ref}` | `{outcome}`, the outcome recorded by the server (generic for unknown, expired or reused references) | — |

Mutations accept exactly one `application/json` body of at most 16 KiB (415, then 400). The credential comes only from
`Authorization: Session <credential>`, and the login binding only from `X-Login-Binding`. Caller-supplied user, role,
MFA and forwarding headers are ignored. Responses are `no-store` with no CORS headers.

## Sign-in and continuity

PKCE (S256) is mandatory. Every start creates a one-use transaction (10 minutes) bound to this browser's login binding.
A callback consumes only the transaction whose state, binding, client and origin all match. An unknown, replayed or
foreign-browser callback consumes nothing. An expired one is deleted.

| Browser at the callback | Callback for | Effect |
|---|---|---|
| Anonymous or dead credential | anyone | New session (mapping found or created). |
| Signed in as A with provider session S | A, the same S | Nothing changes: `already_signed_in`. |
| Signed in as A with S | A, a new provider session S′ | A's old session is revoked (`superseded`); a new session on S′ with a new CSRF token: `session_rotated`. Forms rendered in other tabs answer 403 until reloaded. No provider session is revoked. |
| Signed in as A | B | A is unchanged; nothing is created for B; B's provider session is revoked best effort: `identity_conflict`. |

## Sign-out

| Action | Live session with a valid CSRF token | No live session |
|---|---|---|
| Sign out (this browser) | Revocation committed, then `/sign-in?status=signed_out`. A commit failure shows a 503 page; success is never claimed. | Idempotent: cookie cleared, `signed_out`. |
| Sign out everywhere | Every session of the user is revoked; `signed_out_all`. | 401 |
| Sign out and revoke the WorkOS session | The local revocation is committed first. Then WorkOS revokes exactly this sign-in's session (5 s). Core records the outcome under a one-use reference (10 minutes), which web keeps in the `__Host-logout` cookie (`HttpOnly; Secure; Path=/; SameSite=Lax`, no `Domain`, 10 minutes), never in a URL. Only if the revocation is confirmed does the browser continue to WorkOS's hosted logout, which returns to this origin's registered `/sign-in`; that page, seeing the cookie, hands over with a fixed `307` GET to `/auth/logout-return`. Otherwise the browser goes to `/auth/logout-return` directly. | 401 |

`/auth/logout-return` consumes the reference once through core and shows the server-recorded outcome, clearing the
cookie in the same response (the hosted logout does not carry a query back):
- **revoked:** "Signed out of this application. The WorkOS session for this sign-in was revoked."
- **not revoked:** "Signed out of this application. The WorkOS session for this sign-in could not be revoked or
  confirmed."

An unknown, expired, forged or reused reference shows only "Signed out of this application."; a missing, malformed or
repeated cookie shows "There is no sign-out outcome to show here." The cookie is only an opaque pointer to core's
record, never identity, session authority or a provider result. A confirmed revocation proves only that
WorkOS revoked **that one session**. It does not prove provider-wide sign-out or the end of other applications'
sessions. Remote deprovisioning is not claimed; the trusted path is core's `DisableUser`, which revokes every session in
one transaction. Re-enabling never revives old sessions.

Membership administration (migration 00015) adds a second trusted, route-less disable: `FinalizeRetirement` (`core/internal/identity/retire.go`). Its only
caller is ordinary retirement (`access.RetireUser`), after every Organization departure. It runs in one READ COMMITTED
transaction in the subject's IDENTITY context, with this lock order:

1. Lock `users(subject) FOR UPDATE`.
2. Run access's read-only check once, on that transaction. It rereads the sole-owned scopes and every own-Organization
   page through the 00015 definer readers.
3. If the check returns nothing, run `DisableUser`'s disable and session-revocation statements with the same
   revocation priority. If anything remains, return access's refusal unchanged and write nothing (`conflict`,
   `retirement_blocked`).

The check sees only a guarded view of the transaction. It allows `Query` and `QueryRow` while the check runs and
refuses every other method, and every method after the check returns. The view filters methods, not SQL: `Query` and
`QueryRow` do not themselves enforce read-only statements or an unchanged database context. Those guarantees come from
the actual pinned callback, access's trusted final check, not from a SQL sandbox. An unknown COMMIT is `identity_unavailable` and never a claimed disable.
`DisableUser` itself is unchanged: an immediate emergency disable with no membership check.

## Tenant hosts and the callback relay

Core validates every login, callback and session origin by **resolving its host** ([DOMAINS.md](DOMAINS.md)); an
unknown or non-routing host is `origin_rejected` and nothing is stored. A sign-out origin is checked only for its
structure before the session (an exact `IDENTITY_PUBLIC_ORIGINS` entry, or a lowercase `https` origin with an optional
port and no userinfo, path, query or fragment, at most 300 bytes; otherwise `origin_rejected` 403, nothing revoked), so
sign-out still revokes on a host that stopped routing. The login transaction records the resolved
Organization (and, on a custom host, the shared callback origin).

- **Membership rule (D-2(a)).** A tenant-host sign-in completes only for a current member of that Organization. A
  principal without an account, or an account without a membership, gets no session and no new user or mapping
  (`auth.sign_in` denied `organization_not_member`; `/sign-in?status=not_member`). Accounts for new people are created
  by trusted setup paths, not by a first sign-in on a tenant host.
- **Final issuance guard.** Inside the issuance transaction, before its first write, core checks again that the host
  still routes to the transaction's Organization (`host_changed`, `/sign-in?status=host_changed`) and that the user is
  still a member. The checks are snapshot reads with no new lock; a removal or revocation committed after them is
  refused at the session's next request. Authority stays request-bound, not instant.
- **Session binding (version 12).** A browser session is bound at issuance to the exact origin
  and the Organization of its sign-in (`sessions.origin`, `sessions.organization_id`; set once, never updated). Core
  resolves a credential only together with the request's origin, and only at that origin; the same `__Host-sid` value
  presented on another tenant host is unauthenticated, exactly like a dead credential. Login answers
  `already_authenticated` only when the presented session belongs to the host's Organization; callback continuity
  likewise requires the same Organization; sign-out `current` revokes the session at its own origin (a foreign or dead
  credential stays an idempotent `local_revoked` that revokes nothing), while `all` and provider sign-out also need the
  session at its own origin (otherwise `unauthenticated`, nothing revoked) and then end every session of the user as
  before. Every consumer compares too: `/internal/v1/auth/session` and `/api/account` (401), the domain routes, MCP
  connection setup and revocation (401) and the tenant request scope, which refuses any Organization other than the
  session's (`organization_not_permitted`). A reassigned host therefore does not carry the old Organization's sessions.
- **Migrate 12, then deploy.** Sessions issued before version 12 carry no binding. Apply migration 12 before (or with)
  the release that reads it: from that deploy on such legacy browser sessions are refused by the bound reader, so their
  users sign in again. They stay revocable (sign-out everywhere, `DisableUser`) and recoverable as before (the restore
  invalidation counts and expires them), are retained until retention, and no `session.revoked` event is forged for
  them ([MIGRATIONS.md](MIGRATIONS.md)). Down from 12 is refused while any bound session, any live legacy session or a
  duplicate non-removed hostname exists.
- **Platform hosts** (`<label>.<platform domain>`, the Organization's generated platform label since version 21; the friendly
  slug's host routes nowhere, [DOMAINS.md](DOMAINS.md)) use the unchanged per-host flow; the provider needs wildcard
  redirect and sign-out entries (`https://*.<platform domain>/auth/callback`, `/sign-in`).
- **Custom hosts** use the shared auth host `A` as a relay: the provider's redirect URI is `https://A/auth/callback`;
  web on `A` reads only the `state` value's origin tag, resolves it, and answers 303 to the custom host's
  `/auth/callback` with the raw query byte-identical, or the fixed 404 (`auth.relay` refused with a closed reason). `A`
  sets and reads no cookie, consumes nothing and stores nothing; the custom host consumes the transaction with its own
  `__Host-login` binding exactly as above. Relay and callback responses carry `Cache-Control: no-store` and
  `Referrer-Policy: no-referrer`.
- **Sign-out on custom hosts** works as above, except that WorkOS's hosted logout returns to `A`'s registered
  `/sign-in` (a fixed page with no tenant data). **Limitation:** the recorded outcome is shown on the custom host only
  when the user returns there within 10 minutes (the `__Host-logout` cookie is host-only).

## Platform root signup

Migration `00018_platform_signup.sql`. A person authenticates at the configured
platform root, names a new Organization and its platform-wide unique slug, becomes its owner, and is taken automatically
into the new tenant host. Host routing is in [DOMAINS.md](DOMAINS.md); roles, policies and readers in
[DATABASE-ISOLATION.md](DATABASE-ISOLATION.md); audit rows in [AUDIT.md](AUDIT.md). The root `/` redirects to
`/signup` by default (307; `PLATFORM_ROOT_PAGE=landing` keeps the landing page), and while a default-tenant alias is
active every root `/signup` path is the fixed 404
([DOMAINS.md](DOMAINS.md#default-tenant-root-alias-and-root-presentation)).

- **Root origin.** `https://<PLATFORM_BASE_DOMAIN>` (plus `:PLATFORM_HOST_PORT` outside production), from configuration
  only; core's `hosts/resolve` reports it as `signup_origin`. It is not an `IDENTITY_PUBLIC_ORIGINS` entry. The root's
  callback is its own `/auth/callback` (no relay); the provider needs the exact apex entries (DOMAINS.md "Edge
  deployment inputs"). `www.<domain>` only links to the root; a signup login there is refused.
- **Sign-in at the root.** `POST /signup/start` (form route, exact Origin) asks core for
  `POST /internal/v1/auth/login {origin, intent: "signup"}`, accepted only when `origin` is exactly the root origin (any
  other origin, `www` included, is `origin_rejected`; combined with enrollment fields `invalid_request`), and sets
  `__Host-login` as usual. The login row records `signup_intent`, no Organization and return path `/signup/details`. At
  the callback (T1) core maps `(issuer, subject)` to the user exactly as an ordinary sign-in (created when unknown, never
  by email), creates no session and no tenant context, and inserts one `app.signup_sessions` row: the **signup
  credential** (256-bit, SHA-256 stored, 30 minutes absolute). The answer is `202 {"signup": {"credential": {value,
  max_age}, "next": "/signup/details"}}`; web sets `__Host-signup` and 303s to `/signup/details`; refusals 303 to
  `/signup?status=<closed>`. Audit: `auth.sign_in` success `signup_authenticated` (denials `signup_origin_rejected` and
  the ordinary ones).
- **The signup credential is not a session.** It travels only as `Authorization: Signup <credential>` to
  `/internal/v1/signup/*` at the exact root origin. `/internal/v1/auth/session`, every session consumer, the enrollment
  routes and MCP refuse it; the signup routes refuse a `Session` credential. It carries its own CSRF token (returned by
  the status read, required on the slug, create and abandon calls). Logout and provider logout revoke the signup
  credentials minted from the revoked session's provider session (as does supersession); sign-out everywhere,
  `DisableUser` and `FinalizeRetirement` revoke all of the user's; expired rows are removed at the next login start.
- **Core routes** (`/internal/v1/signup/*`, web BFF only, not public API; JSON ≤ 8 KiB with exactly the route's members;
  `no-store`, `nosniff`, no CORS): `GET status?origin=` → `{status: pending|created|completed, csrf_token,
  organization?: {slug, target_origin}}`; `POST slug {origin, csrf_token, name}` → `{suggested, available}` (advisory);
  `POST organizations {origin, csrf_token, name, slug}` → `201 {outcome: created|already_created, organization: {id,
  slug}, continue}`, `409 conflict {field: slug_taken|intent_used}`, `400 invalid_request {field}`, `403
  csrf_failed|user_disabled`, `401`, `429 rate_limited`, `503 signup_unavailable` (unknown commit: an identical retry
  answers `already_created`); `POST prove {origin, nonce}` → `{complete}` or `403 signup_context_mismatch`; `POST
  challenge {origin, ref}` and `POST complete {origin, challenge, proof}` (tenant origin; no credential, an
  Authorization header is ignored; `X-Login-Binding` as for login) → `{challenge: {value, max_age: 120}, prove}` and
  the ordinary login answer, else `404 continuation_invalid`; `POST abandon {origin, csrf_token}` → `{}` (revokes the
  credential and its current generation).
- **Creation.** `/signup/details` asks for the name (1–200 code points, the tenancy rule), then shows core's suggested
  slug for editing; `POST /signup/create` provisions in one IDENTITY transaction through the provision writer's
  `app.signup_provision_organization`: the Organization, the creator's `org_owner` binding, `organization.created` and
  `membership.granted`, all or nothing. The intent digest `SHA-256(name ‖ 0x00 ‖ slug)` makes an identical repeat
  `already_created` (same Organization, owner and audit pair) and a changed one `intent_used`.
- **Opaque tenant address (migration 00021).** The tenant host is `<label>.<domain>`: core generates the label (an
  adjective-noun-four-digit label such as `gusty-flowerpot-2643` from 6 CSPRNG bytes; earlier 26-character base32 labels keep
  resolving) with each create attempt and passes it to the 8-argument provisioning body, which
  stores it with the Organization (the provision writer's policy requires it) and records the label origin as the
  signup's `target_origin`. A label collision (`label_taken`, nothing retained), a label reserved under the active
  configuration, or a label equal to the submitted slug (web refuses that host) is redrawn within one shared bound of four draws (`503 signup_unavailable` when exhausted); `already_created`
  returns the stored `target_origin`. The slug stays the creator's friendly name: it is not part of the address, its host
  routes nowhere, and the DTOs keep `{slug, target_origin}` (web validates `target_origin` structurally, an `https` origin
  of exactly one label under the root host that is not the typed slug, instead of recomputing it from the slug). The
  label is not an audit attribute.
- **Automatic continuation (four hops, every one a 303 with `Cache-Control: no-store` and `Referrer-Policy:
  no-referrer`; no Continue button).** Each generation holds four one-use 256-bit secrets, stored as SHA-256 only, valid
  120 s:
  1. *create* (root, POST, Origin + CSRF + `__Host-signup`): 201 mints the reference R; 303
     `https://<label>.<domain>/signup/continue?ref=R`. `__Host-signup` is kept.
  2. *challenge* (tenant, GET): core consumes R for that exact origin and its runtime-resolved Organization and returns
     the challenge C and nonce N; web sets `__Host-sct` = C (host-only on the tenant, `Max-Age=120`) and 303s to
     `<root>/signup/prove?n=N`.
  3. *prove* (root, GET): the browser's original `__Host-signup` spends its own row's N and receives the proof P; 303
     `https://<label>.<domain>/signup/complete?p=P`. A missing, foreign, revoked or expired credential matches no row
     and consumes nothing (`signup_context_mismatch`); no CSRF body token here (the only effect is spending
     the presenting credential's own nonce).
  4. *complete* (tenant, GET): `__Host-sct` (C) and P must belong to the same row and generation; core then starts the
     **pinned** tenant-host login (`signup_handoff`, `expected_user_id`, return path `/`), web clears `__Host-sct`, sets
     `__Host-login` and 303s to WorkOS.

  The pinned callback requires the provider subject to map to `expected_user_id` (else `signup_user_mismatch`, 403:
  consumed, nothing issued, `/sign-in?status=failed`), then the ordinary issuance guard (`host_changed`,
  `organization_not_member`, `user_disabled`); success issues the tenant session and answers the closed outcome
  `tenant_created` (`return_to "/"`), which web maps to `/?status=tenant_created` ("Your organization is ready."). Audit:
  `auth.sign_in` success `tenant_created`. A same-user callback on a new provider session is accepted (continuity rests
  on C, P, `__Host-signup`, `__Host-login` and the user pin, never on the provider `sid`). When the provider session is
  not live, WorkOS shows its own prompt at the final hop; another user is `signup_user_mismatch` and a cancel
  `/sign-in?status=canceled`. The tenant already exists and is owned in every non-happy case. A failed hop answers 303
  to `<root>/signup/details?status=continue_failed` (or `/signup?status=expired` when the root credential is gone).
  A captured R, N or P cannot complete another browser's continuation; a captured R can spend that generation, which
  the original browser recovers with Try again (a bounded, accepted availability consequence).
- **Recovery is read-only on GET (no automatic re-mint on a provisioned-state GET).** Every GET of `/signup/details` for an already provisioned signup renders the recovery view
  without any core write (its only signup call is the status read): "Your organization is ready", a manual tenant
  sign-in link, abandon and, only when the browser's restoration record names the created slug, a user-triggered
  **Try again** POST
  (`/signup/create` with the original credential, its CSRF token and the exact recorded name and slug; core's intent
  digest decides). `already_created` re-mints generation g+1 (every secret of g stops matching) and continues; still one
  Organization, owner and audit pair. A missing, stale or mismatched record offers no Try again and makes no core call;
  a dead credential sends the browser to `/signup` (through `GET /signup/abandon`, which only clears the two root
  cookies after core confirms the credential is dead and never revokes), and the owner signs in at the tenant's own
  `/sign-in`. Only the first successful create continues automatically.
- **Cookies** (all host-only, `Path=/`, `HttpOnly`, `SameSite=Lax`, `Secure`, no `Domain`; under `loopback-http` the
  names lose the `__Host-` prefix): `__Host-signup` (root; the signup credential; `Max-Age` 1800 = the row's lifetime;
  cleared on abandon, on any unauthenticated answer and at a new `/signup/start`, otherwise it expires; completion does
  not clear it), `__Host-sct` (tenant; the challenge C; 120 s; cleared at completion or failure) and
  `__Host-signup-form` (root). The last is a **form restoration record, never authentication or
  authority**: value `v1.<base64url UTF-8 name>.<slug>` (name 1–800 bytes, slug empty or one host label), `Max-Age`
  1800, exactly the submitted name and slug and nothing else (no credential, CSRF token, secret or identity data). It is
  written at the suggestion step and on a create success, field refusal, `slug_taken`, rate limit or unknown outcome,
  never rewritten by Try again or on `intent_used`, and cleared on abandon, on any unauthenticated answer and at a new
  `/signup/start`. The create route still requires the credential and its CSRF token; a tampered record is refused as
  `intent_used` and creates nothing. Nothing is set on the auth host.
- **Rate limits** (in-memory buckets per core process; a multi-process deployment multiplies capacity by the process
  count; honest, not comprehensive): root login start, one global bucket (2/s, burst 20); root callback, per login
  binding digest (1/s, burst 5, at most 10 000 keys) and one global bucket (5/s, burst 50); per signup credential on
  every root route (5/s, burst 10); creation per user (one per 10 s, burst 5, so a fresh credential cannot replenish
  it); challenge and complete per tenant origin, each its own bucket (5/s, burst 20). Exhaustion is `429 rate_limited`
  (JSON) or `?status=rate_limited` (forms).
- **Privacy.** Names and slugs travel only in POST bodies and the restoration cookie, never in a URL or log. The
  secrets R, N and P appear only in their own hop's Location and request URL; C only in its cookie; none is logged. Core
  writes one `signup.request` record per signup request (operation, closed outcome and reason, request/trace
  correlation) and never a credential, secret, name, slug, origin or ID; like the identity routes, the web signup
  routes emit no `http.request.completed` record.
- **Evidence boundary.** The local overlay (emulator, `*.localhost:543` proxy, DNS/TLS fixtures) proves the flow's
  control logic, pins and refusals (B-SIGNUP-CATALOG, B-SIGNUP-IT, B-IDENTITY-IT, B-IDENTITY-E2E `signup-*`). Real
  WorkOS apex and wildcard callbacks, provider-session liveness on the final hop, real wildcard DNS/ingress/TLS for a
  new `<label>.<domain>` and production edge query redaction need live qualification, which is not claimed here: live
  identity-provider qualification of the signup flow has not been executed for this export.

## Session liveness in tenant transactions

The trusted context ([AUTHORIZATION.md](AUTHORIZATION.md)) authenticates a request once through `Service.Session`, then
rechecks the session inside every guarded tenant transaction with two trusted `Service` methods. No route or client
reaches them: the session and user IDs come from that earlier resolution.

- `LockLiveSession(ctx, tx, sessionID, userID)` runs in the caller's READ COMMITTED read-write transaction. It reads the
  user row with `status = 'active'` `FOR SHARE`, then the session row (the same user, not revoked, both expiries after
  the service clock's now) `FOR SHARE`, and holds both locks until that transaction ends. This is `DisableUser`'s
  user-then-sessions order. Sign-out, sign-out everywhere, provider sign-out, supersession and `DisableUser` therefore
  either wait for an in-flight guarded write or are observed by it, which then denies.
- `CheckLiveSession(...)` runs the same two reads without locks, for a REPEATABLE READ READ ONLY snapshot.

Both return `ErrUnauthenticated` when either row is absent (no live session, or a disabled user) and `ErrUnavailable` on
any other failure. They renew nothing, write nothing and log nothing; no existing method, statement or table changed. A
session revoked or expired after the request was authenticated is refused by the next guarded operation. One rare lock
cycle is known: a sign-in that supersedes the session in use while the provider display name changed can deadlock with a
guarded write. PostgreSQL aborts one side; an aborted sign-in issues no session and an aborted guarded write changes
nothing (AUTHORIZATION.md).

## Row-level security

From version 6 the identity store reads and writes `users`, `sessions`, `identity_principals` and `login_transactions`
under forced row-level security ([DATABASE-ISOLATION.md](DATABASE-ISOLATION.md)):
- credential resolution, provider-principal lookup and logout-reference lookup call three SELECT-only readers;
- every write runs in a transaction that first initializes its context (IDENTITY for the established subject, LOGIN for
  the browser's binding, EMPTY only before the expired-login cleanup, which now uses the database clock);
- a first sign-in's user UUID is generated in core (version 4), never supplied by a caller;
- a connection carrying prior context is discarded before any protected statement.

The five routes and their responses are unchanged.

## MCP connection sessions

From version 10 a session row has a `kind`: `browser` (every browser sign-in session) or `mcp` ([MCP.md](MCP.md)). Each MCP
connection owns exactly one `mcp` session that carries its liveness (idle and absolute expiry), revocation and the recovery
restore invalidation ([RECOVERY.md](RECOVERY.md)). It has no issued bearer: its credential hash is the hash of a discarded random value, and E1
resolves browser sessions only, so an MCP session can never authenticate a browser request or the `/internal/v1/auth`
routes.
- **Subject mapping (E2).** A verified Connect `sub` is mapped through the configured identity issuer's principal, the same
  namespace as browser sign-in; an unmapped subject is refused (401), never provisioned.
- **Lifecycle.** Setup needs a live browser session and a current membership; activation needs an active user. Renewal
  is conditional (at most once an hour, capped at the absolute expiry). The logout tool and web revoke end one
  connection (`auth.session.revoked` with scope `mcp`); sign-out everywhere and `DisableUser` end every MCP connection of
  the user with the browser sessions; signing out one browser session leaves connections.
- **Test overlay.** The overlay configures core and the replica for MCP with the run's Connect test signer
  (identitysetup's `mcp/connect-jwks.json`, accepted only with `APP_ENV=test`) and no host registry: the
  test hosts resolve through the seeded platform labels and domains ([DOMAINS.md](DOMAINS.md) "Test overlay"). The proxy passes
  only `/mcp/c/` and its protected-resource metadata on the tenant test origins to core. `infra/identity/mcp-seed.sql`
  creates Organizations A, B and C (platform labels `org-a`, `org-b`, `org-c`; friendly slugs `alpha-agency`,
  `bravo-agency`, `charlie-agency`), accounts with their identity principals and
  Organization memberships of A, B and C for alice, bob, carol, dave, frank and the hostile user (D-2(a): a first
  tenant-host sign-in cannot create an account), dave's Group memberships, the routing custom host `portal.acme.test`
  and the removed `old.acme.test` of A, and dave's historical connection on the old host. For Group onboarding it also
  makes frank an `org_admin` of A (alice stays a plain member), bob the `group_owner` of research, dave a `group_member`
  of Group planning of C with no Organization C binding (the Group-role-only subject, refused at C sign-in), and the
  Business Profiles Research Labs (A, attached to research), Design Studio (A, attached to design) and Operations
  Partner (B, unattached). erin (seeded in the emulator
  only) has no account and is refused `not_member`. Every browser project of `identity:browser` needs that seed first:
  the harness looks up the emulator user IDs and applies the file as the database superuser after `identity:migrate`
  and the isolation window (below), as the first writer on the fresh copy;
  for a manual run, apply it with `psql -v issuer=... -v alice=... -v bob=... -v carol=... -v dave=... -v frank=...
  -v hostile=... -v org_a=... -v org_b=... -v org_c=...` (the Organization UUIDs from `infra/identity/mcp-hosts.tsv`).

## Logs

Core records `auth.login.started`, `auth.callback.completed`, `auth.session.revoked`, `auth.provider.revocation`,
`auth.relay` and `identity.namespace_mismatch`, with a closed `outcome`/`scope` and the request's `request_id`/`trace_id`. No
identifier, session or provider value, name, email, code, state, verifier, binding, cookie, CSRF token, token, claim or
WorkOS error body is ever logged. A hostname appears only as its 12-hex tag ([DOMAINS.md](DOMAINS.md) "Records and
logs"): that tag is pseudonymous, not anonymous (an unkeyed hash prefix that anyone can compute for a guessed hostname),
so logs carrying it are not privacy-safe by themselves. An invalid identity configuration stops core with one `core.exit` `config_invalid`
record that names only the allow-listed `IDENTITY_*` variables involved, never a value or file name. Platform root signup adds
`signup.request` for the root signup routes under the same rules ([Platform root signup](#platform-root-signup)).
The default-tenant settings add two core startup events, `default_tenant` (the alias outcome) and `platform_root_page` (an invalid
presentation value), and the `core.exit` category `default_tenant_unavailable`; neither event ever records a malformed
configured value ([DOMAINS.md](DOMAINS.md#default-tenant-root-alias-and-root-presentation)).

The web identity routes (`/sign-in`, `/account`, `/auth/*`, `/api/account`) emit no `http.request.completed` record.
They forward `X-Request-Id` and `traceparent` to core, and core's `auth.*` events carry the correlation. Core's own completion record for each `/internal/v1/auth/*` request names only the matched route pattern (one of the five), as for its health routes (`OPERATIONS.md`). These are
operational logs, not the durable audit trail. From version 8 ([AUDIT.md](AUDIT.md)) the durable record is
`audit.events`: sign-in success (`auth.sign_in` with `user.created`, `session.issued` and a `superseded`
`session.revoked`) commits in the sign-in transaction, so an audit fault fails the sign-in; post-consume callback
failures record `auth.sign_in` denied/failed through a bounded recorder; logout (current, all, provider) and
`DisableUser` record `session.revoked`/`user.disabled` in their transaction but, on an audit-only fault, still commit,
then record `revocation.unaudited` and write one `audit.gap` (revocation priority); `EnableUser` and provider bookkeeping
stay strict.

## Evidence layers

| Layer | What it proves | Where |
|---|---|---|
| L1 fixtures | Token validation, PKCE, binding, continuity, namespace, sign-out logic, the assurance guard, configuration, log privacy | `core/internal/identity/...` unit tests (`core-check`) |
| L2 integration | The store against real PostgreSQL: concurrency, atomicity, lifetimes, replicas, outages, disable, the catalog | `-tags integration` (B-IDENTITY-IT, B-IDENTITY-CATALOG) |
| L2 emulator | HTTPS cookies, CSRF, origins, caching, callbacks, sessions and sign-out end to end; network boundaries; accessibility | `identity:*` below, `web/tests/browser/identity.spec.ts`, and `onboarding.spec.ts` (B-IDENTITY-E2E, -PROBES, -PRIVACY, -CONFIG) |
| L3 staging API | The real token, issuer, client, subject, `sid`, `auth_time` and JWKS contract; MFA enforcement of the tested configuration; email reassignment; `RevokeSession` | `core/internal/identity/workos/staging_test.go` (`workos_staging`; B-IDENTITY-STAGING) |
| L4 manual | Hosted sign-in, MFA success and cancel, SSO reuse, forced reauthentication, the sign-out sequence, two-origin identity | `identity:manual` below (MQ-0..7) |

### The emulator (L2)

The overlay `infra/compose/identity.yml` runs the pinned `ghcr.io/workos/emulate:0.14.0` with a per-run API key,
signing key and seed (`identitysetup`; `infra/identity/emulator-seed.yaml`), the TLS test proxy (pinned
`nginx:1.30.5-alpine3.24`) in web's network namespace, and a core replica.
- **Origins.** The proxy serves `https://org-a.localhost:543`, `https://org-b.localhost:543` and
  `https://attacker.localhost:543` (a static cross-site attack page); it also serves every `*.localhost` and
  `*.acme.test` name to web. There is no static marketing stand-in: the public entry
  `https://localhost:543` (also the platform root and its own signup callback) and `https://www.localhost:543` are
  served by web, and a tenant created during a run (`<label>.localhost:543`) matches the existing `*.localhost` name.
- **Browser projects.** `identity:browser` runs `identity-light`/`identity-dark` (the L2 identity scenarios),
  `domain-platform`/`domain-custom` (E-DOM-1..5; E-DOM-6, the slug host's fixed 404), `mcp-light`/`mcp-dark` (MCP setup) and
  `onboarding-light`/`onboarding-dark` (E-ONB-1..7: Group creation with the suggested and edited address,
  create-or-select Business Profile completion, resume of an incomplete Group that no URL status marks complete, both
  entry points of the shared profile with read-only members, the D-1 member creator and cross-Group isolation, the
  Group-role-only subject refused at sign-in, keyboard/axe/reflow, a concurrent duplicate address, unchanged MCP
  setup-row metadata, and the member-page S-LANDMARKS/S-NAV) and `membership-light`/`membership-dark`
  (E-MEM-1..7: invitation by address with the one-time link and revocation, Group invitation of a co-member and its
  acceptance, promotion/demotion/transfer, the truthful last-owner refusal, owner-only controls and the stale-action
  status page, keyboard/axe/reflow, unchanged MCP connections, and the members-page S-LANDMARKS/S-NAV) and
  `enrollment-light`/`enrollment-dark` (E-ENR-1..5: open signup on B through the ordinary fresh sign-in to an own
  Group, the Organization invitation entry `/join/invitations/<id>#token` on invite-only A, a denial and a
  closed-after-start request on C, a request revoked, re-requested, approved and admitted on C with wrong-authority
  refusals, keyboard/axe/reflow) and `signup-light`/`signup-dark` (E-SUP-1..6, base `https://localhost:543`: root
  and `www` landing, a fresh emulator user and the seeded alice each creating a new tenant through the automatic
  continuation, field errors and abandon, the captured-secret, recovery and no-re-mint-on-GET cases, the provider's
  hosted prompt on the way into the tenant (the emulator's interactive mode prompts on every authorize; a real
  provider's silent return while its session is live needs live qualification, not executed for this export), and first use of the new tenant without onboarding;
  the creators come from the emulator API and the seed, no seed user is added) and `labels-light`/`labels-dark`
  (E-LAB-1..6: Organization labels on `/organization/labels`, assignment on a Group page with no label markup for
  Group roles, Group-owned labels on `/groups/<id>/labels` with the member and staff views, immediate narrowing and
  revocation, forged POSTs, keyboard/axe/reflow; [WEB-SHELL.md](WEB-SHELL.md#labels-pages)). All of them need the seed above (it adds owners for every Organization and Group and a
  second admin of Organization A, makes B open and C request access, closes A's membership history, makes carol an
  owner and frank an admin of C and leaves dave and bob without a C binding; the emulator adds the account-less grace).
- **Browser trust.** The browser tests' Chromium trusts exactly the run's leaf key (`--ignore-certificate-errors-spki-list`),
  never certificate errors in general.

Observed for emulator 0.14.0:

| Feature | Behaviour | Consequence |
|---|---|---|
| Interactive password login, code exchange, JWKS | Supported; `iss = <issuer>/user_management/<client>`, RS256, `sid`, `auth_time` | Core is configured with that issuer |
| PKCE | A wrong verifier is rejected; codes are single-use | E-PKCE runs at L2 |
| Same `sid` on a later sign-in | Every code grant creates a new provider session | The same-`sid` case is L1 and MQ-3 |
| Cancellation | No cancel control | E-CB-7 delivers the provider's `error=access_denied` callback |
| Revocation | Revoke 200 (repeat 404); a revoked session leaves the user's session list | Revocations are counted from that list |
| Hosted logout | Accepts a revoked `sid` and returns to an allowed `return_to` | E-LO-8; whether the hosted session really ends is MQ-6 |
| Error hooks | `POST /_emulate/hooks` fails a route a given number of times | E-LO-6 |
| MFA, SSO reuse, `max_age` | Not modelled | MQ and L3 only |

Run it from a configured checkout (`.env` from `README.md`), in this order:

<!-- identity:start -->
```sh
docker compose --env-file .env -f infra/compose/local.yml -f infra/compose/identity.yml up -d --build --wait
```

<!-- identity:migrate -->
```sh
docker compose --env-file .env -f infra/compose/local.yml -f infra/compose/migrate.yml run --rm --build migrate up
```

<!-- identity:browser -->
```sh
docker compose --env-file .env -f infra/compose/local.yml -f infra/compose/browser.yml -f infra/compose/identity.yml --profile browser run --rm --build --no-deps -e IDENTITY_E2E=1 -v identity-run:/identity:ro web-browser npx playwright test --project=identity-light --project=identity-dark --project=domain-platform --project=domain-custom --project=mcp-light --project=mcp-dark --project=onboarding-light --project=onboarding-dark --project=membership-light --project=membership-dark --project=enrollment-light --project=enrollment-dark --project=signup-light --project=signup-dark --project=labels-light --project=labels-dark
```

<!-- identity:stop -->
```sh
docker compose --env-file .env -f infra/compose/local.yml -f infra/compose/identity.yml down -v
```

**Harness order in `ops`.** After `identity:start` and `identity:migrate` report healthy, B-IDENTITY-IT runs in an
isolation window: (1) `core` and `core-replica`, the only database clients, are stopped and zero client sessions on the
database are asserted; (2) the overlay's own read-only `migrate status` feeds its source-bound pristine check (the same
goose-history, catalog, empty-baseline, ACL and session checks as the `migrations` suite, without any `migrations`-suite
row), and the database is copied into the run-private template; (3) the identity integration tests run on the
canonical database, every package the selector names expected per test (`identity`, `workos`, and `identitytest` with no
test files); (4) the database is recreated from the template; (5) `core` and `core-replica` start again; then
`mcp-seed.sql` is applied and the browser projects, probes and B-MCP-E2E follow. The seed and every browser spec
therefore see only the seed's content, not residue from the identity tests. A refusal in the window (a lingering session,
a failed stop or restart, a dirty baseline) records B-IDENTITY-IT FAIL with the cause and stops the overlay; the tests
are never run on the shared database instead.

### WorkOS staging (L3)

`TestStagingContract` (build tag `workos_staging`) runs against the real WorkOS staging API.
- **Inputs.** It reads `IDENTITY_STAGING_WORKOS_API_KEY`, `IDENTITY_STAGING_WORKOS_CLIENT_ID`,
  `IDENTITY_STAGING_WORKOS_ISSUER` and the database settings. In CI these come from the protected `identity-staging`
  job; on an operator workstation, from a 0600 env file outside the repository.
- **Per-run users.** It creates its own users (`example.org` addresses, API-enrolled TOTP), asserts the contract, then
  deletes them.
- **Output.** It writes only a sanitized result: issuer shape, client fingerprints and outcomes.
- **Selector.** `-run '^TestStagingContract$'` (the hosted job and the block below).

Recorded observations of this contract test (it needs real provider credentials and has not been executed for this
export):
- **Issuer and token.** The issuer shape above, distinct from the configured client. RS256 with `kid`. `sub` equals the
  user ID; `sid` and a recent `auth_time` are present.
- **MFA.** A password alone returns the MFA error class and no tokens; TOTP returns tokens. The tested configuration has
  MFA required.
- **Attributes.** An email change plus rename keeps the subject and account. A new user reusing the former email gets a
  new subject and account.
- **Revocation.** A repeat revoke is accepted (idempotent).
- **Tampering.** Tampered tokens, the wrong client, a shifted clock and the trailing-slash issuer are all refused.

**Warning.** Before it runs, the test empties every identity table of the database it migrates (the Compose
database that `.env` names), together with the membership bindings that reference those users (`TRUNCATE … CASCADE`,
since `app.role_bindings` references `app.users`). Organizations and Groups remain, without those bindings. Run it only
against a disposable qualification database, never against one whose identities or memberships you need; see
`TENANCY.md`.

<!-- identity:staging-local -->
```sh
docker compose --env-file .env -f infra/compose/local.yml --profile checks build core-check
docker compose --env-file .env -f infra/compose/local.yml up -d --wait postgres
docker compose --env-file .env -f infra/compose/local.yml -f infra/compose/migrate.yml run --rm --build migrate up
docker run --rm --network "$(sed -n 's/^COMPOSE_PROJECT_NAME=//p' .env)_data" --env-file "$IDENTITY_STAGING_ENV_FILE" "$(sed -n 's/^COMPOSE_PROJECT_NAME=//p' .env)-core-check" go test -tags workos_staging -run '^TestStagingContract$' -count=1 -v ./internal/identity/workos/
```

`IDENTITY_STAGING_ENV_FILE` names the operator's 0600 file. It holds the three `IDENTITY_STAGING_WORKOS_*` values plus
`DATABASE_NAME`, `DB_HOST=postgres`, `DB_PORT=5432`, `DB_MIGRATOR_PASSWORD`, `DB_RUNTIME_PASSWORD` and `APP_ENV=test`.
For the domain qualification test below it also holds that test's three domain inputs. It never enters the
repository, an image or a log.

**Domain qualification test (preparation only; live evidence is not claimed).** `TestStagingDomainQualification` (same
tag, same package) is a separate test that runs only when selected with `-run '^TestStagingDomainQualification$'`.
- **Inputs.** When selected it requires the three `IDENTITY_STAGING_WORKOS_*` values plus:
  - `IDENTITY_STAGING_PLATFORM_DOMAIN`, the platform domain;
  - `IDENTITY_STAGING_AUTH_ORIGIN`, the shared auth origin;
  - `IDENTITY_STAGING_AUTHKIT_ORIGIN`, the verified origin of WorkOS's hosted sign-in, which the provider redirects an
    accepted authorization request to (it cannot be derived from the other inputs).

  A missing or malformed input fails as UNMET; there is no skip. It needs no database.
- **What it observes.** WorkOS's own responses, never a constructed URL alone.
  - Authorization: a callback on a per-run random host under the platform domain (only a wildcard entry can admit
    it) and on the auth origin must be redirected to exactly the hosted sign-in origin.
  - Hosted logout: run after the production API revocation, on a fresh session per case, it must return to exactly
    that host's `/sign-in`.
  - Negative control: a reserved `.invalid` host must be refused at authorization and not honored at logout.
  - Any other response qualifies as neither, so the test fails.
- **Offline fixtures.** `-run '^TestStagingDomainObservationFixtures$'` checks these rules against table cases and
  loopback responses. It needs no provider or inputs.
- **Evidence boundary.** A pass only prepares live-domain qualification evidence; that qualification has not been
  executed for this export. The code-bearing redirect after an interactive hosted
  sign-in stays a browser step. The hosted `identity-staging` job neither selects this test nor forwards its inputs;
  it runs only through the env-file path, with the same `docker run` and the domain selector.

### Manual qualification (L4, MQ)

**Local HTTPS troubleshooting.** Endpoint-security software on a development workstation may need host-specific
configuration for each local preview host (for example `org-a.localhost`, `org-b.localhost`); a working host does not
establish that another is covered. Check name resolution, the run's certificate names and trust, and the exact browser
error; a host-side 200 alone does not establish browser-side success. No security-software setting or TLS validation
is changed by the repository.

Real hosted sign-in on the operator's workstation, against the exact commit under review. The overlay runs with core
pointed at WorkOS staging and the proxy published on `127.0.0.1:543`.
- **Staging configuration.** A 0600 file outside the repository supplies `IDENTITY_TEST_CLIENT_ID`,
  `IDENTITY_TEST_ISSUER`, `IDENTITY_TEST_API_BASE_URL=https://api.workos.com`, `IDENTITY_TEST_API_KEY`, an empty
  `IDENTITY_TEST_API_KEY_FILE=`, `IDENTITY_TEST_MAX_AUTH_AGE` and `IDENTITY_PROXY_PUBLISH=127.0.0.1:543`.
- **Browser trust.** The run's CA (`ca.pem` from the run volume) is trusted only in a dedicated browser profile.
- **Observer.** `TestManualQualification` (build tag `workos_manual`) prints each step, watches the database and logs
  for the expected transition, and writes sanitized results.
- **Operator.** Drives the hosted pages only and attests UI-only facts. Radar and bot protection are never bypassed.

<!-- identity:manual -->
```sh
docker compose --env-file .env --env-file "$IDENTITY_MANUAL_ENV_FILE" -f infra/compose/local.yml -f infra/compose/identity.yml up -d --build --wait
docker compose --env-file .env -f infra/compose/local.yml -f infra/compose/migrate.yml run --rm --build migrate up
docker compose --env-file .env -f infra/compose/local.yml -f infra/compose/identity.yml cp identity-setup:/identity/ca.pem "$IDENTITY_MANUAL_CA_FILE"
```

| Step | Operator | Server-observed pass condition |
|---|---|---|
| MQ-0 | Preflight | Tree, configuration and CA hashes; binding row absent or matching |
| MQ-1 | Sign in on `org-a` with password and a TOTP challenge | Session; `mfa_assurance = unknown`; recent signed `auth_time` |
| MQ-2 | Start as user B; cancel at TOTP | No B session |
| MQ-3 | Sign out locally, sign in again (no prompt) | Single sign-on reuse only when not prompted, the same user, a new session after the local revocation and the latest prior signed `auth_time` kept; a prompted sign-in is an ordinary sign-in and fails the step |
| MQ-4 | Sign in on `org-b` | Same user |
| MQ-5 | Sign out; wait until the latest signed authentication of any session is older than `IDENTITY_MAX_AUTH_AGE` plus 30 s (re-checked before "ready"; a sign-in during the wait extends it, within a bounded wait), then sign in | Prompt attested; `auth_time` advanced past the baseline and fresh; baseline, bound, ready time and new `auth_time` recorded |
| MQ-6, MQ-6B | On `org-a` (MQ-6) and `org-b` (MQ-6B): sign in on that origin, then sign out and revoke the WorkOS session | That session revoked locally and confirmed at the provider; the return reaches that origin's `/auth/logout-return` (reference consumed, `__Host-logout` gone; destination attested); the next sign-in asks for credentials with an advanced signed `auth_time` |
| MQ-7 | Disable via the observer, sign in, re-enable, sign in | `user_disabled`, then a new session; old sessions stay revoked |

The two provider capability checks use the same setup.
- **PKCE refusal.** The observer exchanges a real code with a *different* PKCE verifier; WorkOS must refuse it.
- **Sign-out and freshness.** Checks the sign-out sequence and freshness: the logout URL after a revocation returns, and the next
  sign-in asks for credentials.

A failed capability stops the work it supports; there is no fallback. The manual qualification needs real provider
configuration and an operator, and has not been executed for this export.

**Browser evidence collection.** The harness runs the identity browser command as a run-owned named one-off and copies
its internal reports out to operator ownership before container removal. The identity browser command, single-worker
behavior, required scenarios and observed-secret registration/removal are unchanged. See
[operations](OPERATIONS.md#browser-report-ownership).

### Live-domain qualification

The domain-aware code and the provider-test extension above are present. Actual WorkOS wildcard callback/sign-out
acceptance and custom-host HTTPS browser behavior require live qualification, which has not been executed for this
export; local emulator passes do not satisfy it, and earlier L3/L4 contract observations do not automatically prove the
domain contract. All qualification hosts are configurable inputs; no permanent staging environment or product-specific
domain is required. Capture only sanitized outcomes; never require manual copying of raw callback parameters or
session cookies.
