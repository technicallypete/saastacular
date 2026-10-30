# Hostnames and custom domains

Every tenant request names its Organization through its `Host`. Core decides which canonical Organization a host routes
to right now; names are routing aliases, UUIDs stay identity, and every request still authorizes on current facts.
This runbook covers the host model, the
custom-domain lifecycle, renames and removals, the edge inputs an operator must provide, the evidence and the limits.

## Host model

Core normalizes one `Host` value (ASCII lowercase, one trailing dot stripped, labels 1–63, at most 253 characters; IP
literals, userinfo, paths, non-ASCII and `xn--` labels refused, no IDNA in v0). A port is accepted only when it equals
`PLATFORM_HOST_PORT` (empty in production). Forwarding headers are never read; the edge terminates TLS and passes
`Host` as the client sent it. With `B = PLATFORM_BASE_DOMAIN` and `A = PLATFORM_AUTH_HOST`:

| Host | Meaning | Answer |
|---|---|---|
| `B`, `www.B` | public entry | web serves the platform landing on `/` (on `B` per `PLATFORM_ROOT_PAGE`, and `B` is a platform host while the default-tenant alias is active, below); `B` alone also serves the root signup surface and its own `/auth/callback`; everything else is the fixed 404 (`/api/health`, `/api/ready` exempt) |
| `A`, `<reserved>.B`, the CNAME target | reserved | 404 (A serves only the callback relay, its fixed `/sign-in` page, the product mark and favicon files, and a GET or HEAD of exactly the four font files that page loads: `/fonts/geist/geist-wght.ttf`, `/fonts/oswald/oswald-wght.ttf` and `/fonts/barlow-condensed/barlow-condensed-400.ttf` and `-600.ttf`; `/fonts/` is no prefix exemption there) |
| `<label>.B` | platform host of the Organization whose generated `platform_label` is that label (the friendly slug never routes) | routes while the label is set |
| anything else | custom host | routes only while its live domain row is routing-ready (below) |

Apart from the optional exact-root alias (below), there is no default tenant: an unknown, removed, suspended or not-yet-ready host is the fixed 404 page (core
`unknown_host`), and a store failure is 503, never a clean unknown host (on a gated JSON route, `/api/*`, the same
statuses come as `{"error":"not_found"}` and `{"error":"unavailable"}`). The reserved floor is the Group reserved set
plus `www`, `auth`, `custom`, `app`, `mail`, `smtp`, `ns`, `ns1`, `ns2`, `cdn`, `static`, `status` and
`PLATFORM_RESERVED_LABELS`; a custom claim equal to `B` or `A`, or under `.B`, is `host_reserved`.

**Public entry (replaces an earlier 303 to `MARKETING_SITE_URL`).** `hosts/resolve` answers `B` and `www.B` with
exactly `{"kind": "public_entry", "signup_origin": "https://B[:PLATFORM_HOST_PORT]"}` (the configured platform root
origin; the former `location` member is gone, and web treats an answer carrying it as malformed). Web serves the public
entry instead of redirecting it. The root (the request's own origin equals `signup_origin`) serves `/`, `/signup`,
`/signup/details` and the root routes `/signup/start`, `/signup/create`, `/signup/abandon` and `/signup/prove`; its
`/auth/callback` is the root signup callback (on `www.B` that route is the fixed 404). `www.B` serves `/` only, and its
single Sign Up action links to `${signup_origin}/signup`, so signup always runs on `B`. Both also serve the shell's
static `/fonts/` and `/blueprint/` files, and a GET or HEAD of exactly the product mark `/brand/mark.png` and the
favicon `/icon.png` (as does `A`). The landing has no sign-in, picker, directory or lookup. On a
platform tenant host `/signup/*` is the fixed 404 except the two continuation hops `/signup/continue` and
`/signup/complete`; on a custom host every `/signup/*` path is the fixed 404. The flow is in
[IDENTITY.md](IDENTITY.md#platform-root-signup). Under `loopback-http` the gate is skipped and `/` keeps the
generic page. The root's `/` follows `PLATFORM_ROOT_PAGE` (a 307 to `/signup` by default, or this landing),
and an active default-tenant alias makes the root its Organization's platform host instead of a public entry
([Default-tenant root alias and root presentation](#default-tenant-root-alias-and-root-presentation)).

Resolution is one SECURITY DEFINER reader, `app.hostname_state(kind, key)`, one unique-index probe per request with no
cache: a commit is the invalidation. Web's host gate (`web/proxy.ts`) asks core on every page and form request; the
identity login, callback and session routes, the MCP handler and the TLS `ask` route resolve the same way.

## Configuration (core)

None of these values is a secret ([.env.example](../../.env.example), "Hostnames and custom domains").

| Variable | Meaning (default) |
|---|---|
| `PLATFORM_BASE_DOMAIN`, `CUSTOM_DOMAIN_CNAME_TARGET` | the platform domain and the edge name customers point their CNAME at |
| ~~`MARKETING_SITE_URL`~~ | removed (the public entry is served, not redirected); core no longer reads it, and an environment that still sets it is ignored |
| `PLATFORM_AUTH_HOST` | the shared auth host (`auth.B`); its origin must be the only entry of `IDENTITY_PUBLIC_ORIGINS` |
| `PLATFORM_HOST_PORT` | the only port accepted in a `Host` (test overlay `543`; refused in production) |
| `PLATFORM_RESERVED_LABELS` | extra reserved first labels |
| `CUSTOM_DOMAIN_CHALLENGE_LABEL` | TXT record label (`_platform-verify`) |
| `CUSTOM_DOMAIN_CHALLENGE_TTL` | how long a pending claim may verify (168 h; 1 h–720 h) |
| `CUSTOM_DOMAIN_PROOF_FRESHNESS` | how long a successful check keeps the proof fresh (720 h; 24 h–2160 h) |
| `CUSTOM_DOMAIN_DNS_TIMEOUT`, `DOMAIN_TLS_TIMEOUT` | per-check deadlines (5 s; 1–15 s) |
| `DOMAIN_TLS_ADAPTER`, `DOMAIN_TLS_EDGE_ADDR` | `edge-probe` dialing `<CNAME target>:443` (the `fixture` adapter is `APP_ENV=test` only) |
| `DOMAIN_WITHDRAWAL_RETRY_MAX` | explicit withdrawal retries per removed row (5; 1–20) |
| `CUSTOM_DOMAIN_DNS_FIXTURE_FILE`, `DOMAIN_TLS_FIXTURE_FILE` | **test only**; refused outside `APP_ENV=test` |
| `DEFAULT_TENANT_SLUG` | optional: the slug of an existing Organization whose tenant pages the exact root `B` serves (unset: multi-tenant root); never a configuration error (below) |
| `PLATFORM_ROOT_PAGE` | optional: the multi-tenant root presentation, `signup_redirect` or `landing` (`signup_redirect`, also when unset or empty); never a configuration error (below) |

An invalid combination stops core with one `core.exit` `config_invalid` record naming variables only. The retired MCP
host registry (`MCP_TENANT_HOSTS_FILE`) is refused in every environment ([MCP.md](MCP.md)).

### Default-tenant root alias and root presentation

Two optional, non-secret core settings decide what the exact platform root `B` (its configured origin,
`https://B[:PLATFORM_HOST_PORT]`) serves. Core reads both once at startup; neither is ever a `config_invalid` problem.
Web has no setting of its own: it applies core's `hosts/resolve` answer on every request.

**Default-tenant alias (`DEFAULT_TENANT_SLUG`).** Set it to the slug of an Organization that already exists (created
through the normal flow) and restart core. The alias creates, seeds, renames or converts nothing and changes neither the
multi-tenant data model, other hosts nor the Organization provisioning APIs. At startup core trims environment
whitespace and judges the value exactly as given against the Organization slug grammar and the reserved first labels of
the active configuration (no case folding: `Alpha-Agency` never selects `alpha-agency`); a well-formed value is looked
up with one statement, the auth-reader body `app.organization_id_by_slug` (migration 00022,
[DATABASE-ISOLATION.md](DATABASE-ISOLATION.md)), and the Organization's UUID is pinned in the resolver for the life of
the process. Display names and platform labels are never selection keys.

| Value at startup | Effect | Startup record |
|---|---|---|
| unset or empty | multi-tenant root (`PLATFORM_ROOT_PAGE` below); no statement | none |
| the slug of an existing Organization | alias active | `info` `default_tenant` `outcome=active`, `slug`, `organization_id` |
| a well-formed slug no Organization holds | alias inactive; multi-tenant root | `warn` `default_tenant` `outcome=inactive`, `reason=unknown_slug`, `slug` |
| outside the slug grammar, or a reserved label | alias inactive; multi-tenant root; no statement | `warn` `default_tenant` `outcome=inactive`, `reason=invalid_slug` (the value is never logged) |
| the lookup fails (database, permission or transport) | core stops, exit 1 | `core.exit` `default_tenant_unavailable` |

A lookup failure is never treated as an unknown slug: serving the multi-tenant root while a tenant is configured would
silently expose the wrong behaviour. The orchestrator's restart policy retries it, as for the other startup exits. A
version-22 core with the setting against a version-21 database exits the same way (migrate, then deploy).

While the alias is active, `hosts/resolve` answers the exact root, without a database statement, with
`{"kind": "platform", "organization_id": "<uuid>", "root_alias": true}` (`root_alias` is absent from every other
answer), and every consumer treats the root as that Organization's platform host at the root's own address: the tenant
entry and branding, `/sign-in`, `/auth/*` (the root's own `/auth/callback`, a host-only session cookie), `/account`,
`/join` under that Organization's enrollment policy, `/groups/new` and MCP connections bound to the root's exact host.
Every root `/signup` and `/signup/*` path, GET and POST, the continuation hops `/signup/continue` and
`/signup/complete` included, is the fixed 404 (JSON on `/api/*`); nothing redirects to `/join` or Group creation. Root
signup is therefore unreachable while every provisioning API and permission stays unchanged; a root signup in flight at
the restart cannot finish its `/signup/prove` hop and expires without becoming a tenant session. To create another
Organization, unset the alias and restart. The root callback limiters (1/s burst 5 per binding, 5/s burst 50
overall) also meter ordinary sign-ins on the aliased root.

The alias covers the exact root only. `www.B`, platform label hosts, custom hosts, `A`, reserved and unknown hosts are
unchanged; `www.B` keeps its landing, whose Sign Up then reaches the 404 root `/signup`, so an aliased deployment should
not publish `www.B` or should redirect it at the edge (an edge input outside this repository).

A change takes a restart; an env-file edit alone changes nothing, and there is no watcher or per-request slug search.
Restarting with another slug pins that Organization: root sessions and login transactions of the previous one stay
bound to it and are refused, as on any reassigned host (members sign in again), and MCP connections bound to the root
for the previous Organization are refused. Removing the setting, leaving it empty or restarting with an unknown slug
restores the multi-tenant root; a former tenant cookie on the root is never read as a signup credential. If the
selected Organization disappears while the process runs, the pinned UUID stays and every context, membership and
session check fails closed as for any other host.

**Root presentation (`PLATFORM_ROOT_PAGE`).** Selects the multi-tenant root, that is whenever no valid alias is active
(unset, empty, unknown or malformed slug): `signup_redirect` (the default, also when unset or empty) or `landing` (the
platform landing with its Sign Up action). Values are exact; anything else logs `warn` `platform_root_page`
`outcome=invalid_value`, `fallback=signup_redirect` (never the value) and uses the default. Core adds the effective
mode as `root_page` to every `public_entry` answer. In `signup_redirect` mode web answers a GET or HEAD of exact `/` on
the root origin with **307**, `Location: <root origin>/signup` (the origin core named in its own answer, never the
request's Host or URL; fixed; the request's query is dropped, never forwarded) and `Cache-Control: no-store`, setting no cookie, starting no provider authentication and creating nothing;
`/signup` itself, POST `/`, every other path, `www.B` (its `/` keeps the landing) and every other host are not
redirected. A valid alias always takes precedence: the aliased root never answers `public_entry`, so the setting has no
effect there and no redirect loop is possible. The mode is never inferred from the number of Organizations.

**Scope and transport.** Both settings act through web's host gate, which runs on the `https` cookie profile only:
under `loopback-http` (the quick start) the gate is skipped, `/` keeps the generic page, and `infra/compose/local.yml`
passes neither variable to core. The test overlay passes `IDENTITY_TEST_DEFAULT_TENANT_SLUG` (default empty) and
`IDENTITY_TEST_PLATFORM_ROOT_PAGE` (default `landing`, so the shared identity browser scenarios keep the landing) to
`core` and `core-replica`. The included staging deployment overlay ([STAGING.md](STAGING.md)) carries both as optional non-secret lines: the
environment's `/run/<env>/app.env` written by `staging.conf` gains `DEFAULT_TENANT_SLUG=<slug or empty>` and
`PLATFORM_ROOT_PAGE=<signup_redirect|landing or empty>`, and `infra/compose/staging.yml` passes
`DEFAULT_TENANT_SLUG: ${DEFAULT_TENANT_SLUG:-}` and `PLATFORM_ROOT_PAGE: ${PLATFORM_ROOT_PAGE:-}` to core, with the
semantics of this section. Shared staging may leave the slug unset and select `landing` explicitly.

## Custom-domain lifecycle

Rows live in `app.organization_domains` (migration 00011, [MIGRATIONS.md](MIGRATIONS.md)): one row per claim
generation, immutable Organization, hostname and generation, at most one live (non-removed) row per hostname. Every
operation is an owner/admin operation (`organization.manage_domains`) through the access facade's guarded transaction;
DNS and TLS work runs between two facade calls, outside every transaction. The routes are internal to core
(`/internal/v1/domains*`, `/internal/v1/organization/slug`); they need the browser session and its CSRF token and resolve
the Organization from the setup origin. No domain management page is shipped.

1. **Claim** (`POST /internal/v1/domains {hostname}`): a `pending` row with a 52-character random challenge value (a
   public DNS value, not a credential, never logged). From version 12 pending claims are not exclusive: several
   Organizations may hold a pending claim of one hostname, each with its own challenge. A claim of a hostname that is
   already proven (`verified`, `active` or `suspended`) anywhere is `host_taken` with no effect, and at promotion the
   first valid proof to commit wins; the other claimants' promotions are `host_taken` and their rows stay pending.
2. **Customer DNS**: a TXT record `<CUSTOM_DOMAIN_CHALLENGE_LABEL>.<hostname>` with the challenge value, and a CNAME
   chain (at most three hops) from the hostname ending exactly at `CUSTOM_DOMAIN_CNAME_TARGET`. Both are required; a
   shared CNAME target alone never proves ownership. The customer keeps the TXT record: it is the ongoing proof.
   The production checker reads **one actual edge per lookup**, at most three lookups per proof: each lookup is one
   typed `CNAME IN` query for the exact absolute name, and its result is the unique CNAME record owned by that name
   (owner compared case-insensitively), whatever the order of the answer section; no CNAME (NODATA) is the chain's
   end, NXDOMAIN a missing name, and several owner CNAMEs or a malformed answer are transient (never ok). The check
   passes when one of the first three edges points exactly at `CUSTOM_DOMAIN_CNAME_TARGET`; whatever the target
   itself points at (for example a provider edge name) does not matter, and a fourth edge fails even when it would
   reach the target. The query goes to the trusted nameservers of the system configuration only: the first three
   `nameserver` lines of `/etc/resolv.conf` (port 53; Go's loopback defaults when there are none), over UDP with the
   same query retried over TCP on truncation, or TCP only under `options use-vc`; the answer's ID, header and
   question echo are validated before truncation is acted on, a TCP frame holds 12–4096 bytes and every section is
   parsed before an edge is accepted. Every declared record in every section must be framed inside the message and
   is parsed, but exact content-length validation applies only to answer CNAME evidence: the encoded CNAME target
   must consume exactly its declared RDLENGTH (a compression pointer counts as its own two bytes). Other record types
   keep the upstream parser's semantic-length limitations; this is not a strict validator for every RR type, and
   such records cannot supply the owner-CNAME edge.
   Each server gets an equal share of the remaining lookup time (the caller's
   deadline, or 5 s when it has none), and cancellation closes the connection at once. Other resolv.conf settings
   (`search`, `domain`, `ndots`, `rotate`, `attempts`, `timeout`, `edns0`, `trust-ad`, other options) are deliberately
   not interpreted; this is not native resolver parity, and DNS stays unauthenticated (no DNSSEC). Bytes after the
   counted sections are not observable through the parser. The TXT challenge is looked up through Go's native
   resolver. `TestDNSSystemChecker` and the `TestDNSWire` tests pin these semantics on wire answers (three edges pass,
   four fail, any record order); a bounded live TXT/CNAME proof at a domain the operator controls is deployment
   prerequisite **P-3** and has not been executed for this export.
3. **Verify** (`/verify`): before the challenge expires; ok → `verified` with `verified_until = now + freshness`. An
   expired pending challenge is renewed with **rotate** (`/rotate`, pending only).
4. **TLS readiness** (`/recheck`): the edge-probe dials `DOMAIN_TLS_EDGE_ADDR` with SNI = hostname and checks chain,
   SAN and expiry. Ready → `active`.
5. **Recheck** periodically (operator- or owner-invoked; no background reconciler): ok renews freshness; a confirmed
   proof loss suspends and sets `verified_until = now()`, so TLS issuance stops at that commit; a TLS loss suspends
   while issuance continues on a fresh proof; a transient failure changes neither state nor deadline.
6. **Remove** (`/remove`): non-routing and non-issuing at commit. The adapter's withdrawal runs after the commit on the
   row's own id and generation; under E1 it is `withdrawn` at once.

Routing requires `active`, a fresh proof and TLS `ready` and unexpired. TLS **issuance** (the `ask` answer) requires
`verified`, `active` or `suspended` with a fresh proof, so a certificate can be obtained before the host routes and
renewed through a TLS-only suspension. Results of DNS/TLS work apply only to the row id and generation they were
started for, under the version guard; a late result for another generation is dropped (`stale_result`).

**Withdrawal retry** (`/retry-withdrawal`) exists for adapters whose removal can fail; under E1 it always answers
`domain_state_invalid` because withdrawal completed at the removal commit. After `DOMAIN_WITHDRAWAL_RETRY_MAX`
non-withdrawn outcomes (`domain_retry_exhausted`) nothing in the platform retries: the remaining provider cleanup is a
manual operator step against the recorded `tls_reference` of that row (for a future provider adapter: delete exactly
that resource id in the provider console), recorded in the change log. A replacement claim never waits for it.

**Reassignment** of a hostname to another Organization is a new claim with a new generation after the old row is
removed; it needs its own fresh proof before issuance or routing.

## Platform addresses, slugs and renames

**Opaque platform address (migration 00021; replaces the earlier slug-host routing).** A platform host is
`<label>.B`, where the label is `organizations.platform_label`: optional (NULL = no platform host), unique, `C`-collated,
checked against the slug grammar and the same static reserved floor, and never updated at runtime (no role holds
`UPDATE (platform_label)`). Root signup generates it in core from 6 CSPRNG bytes as a lowercase adjective-noun-four-digit label such as
`gusty-flowerpot-2643` (two curated 64-word lists, 40,960,000 combinations; a public routing identifier, not a credential;
the earlier 26-character base32 labels keep resolving unchanged), atomically with the Organization; a label reserved under the active configuration (`PLATFORM_RESERVED_LABELS`, or
the first label of `CUSTOM_DOMAIN_CNAME_TARGET`/`PLATFORM_AUTH_HOST` under `B`) or equal to the submitted slug (web refuses a tenant host equal to the Organization's own slug) is redrawn in
core before any statement, because the SQL CHECK holds the static floor only. A collision with another label is retried with a fresh label within
one bound of four draws per create, shared by the reserved, slug-equal and taken redraws; exhaustion is `503 signup_unavailable` with nothing retained. `app.hostname_state`
resolves the label directly to the Organization UUID with one unique-index probe, as before. The friendly slug is
metadata: `<slug>.B` routes nowhere (the fixed 404, no redirect, nothing disclosed about the label), and there is no
alias, directory or lookup from slug to label. Rows that existed before version 21 received a random 40-character hex
label (160 bits) in the migration.

`organizations.slug` stays optional, unique and checked against the same grammar and reserved floor; slug availability
is answered by its own reader `app.signup_slug_taken`, never by the routing reader. An Organization owner or admin sets
the slug through `POST /internal/v1/organization/slug`; the slug guard lets an admin change the slug and nothing else.
Since version 21 a slug change moves no host: the label, its routing and the sessions bound to it are unchanged (DOM-11).
A custom-domain removal (and, before version 21, a slug rename):

- stops the old host at commit: pages 404, a login there is refused, and a browser session cookie (host-only
  `__Host-sid`) is useless on the new host, so members sign in again there;
- keeps historical MCP connections: a connection is admitted only while its exact stored host still routes to its
  bound Organization, so it is 404 on a removed or renamed host and 401 on a host reassigned to another Organization;
  it is never retargeted or automatically revoked. Owners list and revoke it through any current host of the
  Organization ([MCP.md](MCP.md) "Hosts and aliases");
- needs no provider change for platform hosts (provider wildcard entries, below).

**Creation-time slugs.** Root signup creates an Organization together with its slug, chosen by the creator
(core's advisory suggestion is editable) and checked against the same grammar, reserved floor and configured labels;
the unique slug index makes the first creation win and the other attempt `slug_taken` with nothing retained. Since version 21
the new Organization's generated `<label>.B` (not `<slug>.B`) routes from that commit: platform hosts are one database
mapping behind the shared wildcard DNS, ingress and certificate (edge inputs below), so creation adds no per-tenant DNS
or CNAME record and waits for no propagation. The continuation targets the stored label origin; an identical repeat
returns the stored origin, never one recomputed from the slug.

## Browser sign-in on tenant hosts

A tenant-host sign-in completes only for a current member of the resolved Organization (D-2(a)): a principal without an
account cannot create one on a tenant host. Platform hosts use the unchanged per-host sign-in flow with provider wildcard
redirect and sign-out entries (`https://*.B/auth/callback`, `https://*.B/sign-in`). Custom hosts use the shared
callback relay on `A`: the provider returns to `https://A/auth/callback`, web on `A` reads only the `state` origin tag,
resolves it and answers 303 to `https://<custom host>/auth/callback` with the raw query unchanged, or the fixed 404.
`A` sets and reads no cookie and stores nothing. Details, refusals and limits are in [IDENTITY.md](IDENTITY.md).
The platform root signup adds the root's own signup sign-in on `B` (`https://B/auth/callback`, no relay), which creates no tenant session;
the only way from it into a tenant is the continuation to the new tenant's own host and its ordinary pinned login.

## Edge deployment inputs (E1, Caddy on-demand TLS)

The production edge is a deployment input (P-2/P-4); no deployment is claimed. The operator configures:

- **On-demand TLS permission**: Caddy's `on_demand_tls { ask http://<core internal address>/internal/v1/domains/tls-ask }`
  (in the Compose stack `http://core:8180`); Caddy adds `?domain=<hostname>`. Core answers 200 only for a custom hostname whose live row has a fresh proof and 403 otherwise,
  including from the removal commit on. Caddy keeps an already issued certificate until its expiry; routing already
  refuses.
- **Platform certificate**: the shared wildcard for `*.B` (and `B`), outside the tenant lifecycle.
- **Host passthrough**: the original `Host` to web; core reachable only from the edge/web network; no forwarding header
  is trusted.
- **Callback query redaction**: the access log must omit `/auth/callback` requests or redact the **entire** query
  (`code`, `state`, `error`, `error_description` and anything else), for example a Caddy `log` `filter` that replaces
  `request>uri` for that path. The test overlay keeps `access_log off` (`infra/identity/proxy.conf`). The operator
  confirms the log configuration before manual qualification. The same applies to `/signup/continue`,
  `/signup/prove` and `/signup/complete` on every host, whose queries carry one-use continuation secrets (`ref`, `n`,
  `p`).
- **Provider entries** (one WorkOS application per environment): the wildcard platform entries above and the exact
  `https://A/auth/callback` and `https://A/sign-in`; platform root signup adds the exact apex entries `https://B/auth/callback` and
  `https://B/sign-in` (a wildcard does not match the apex; registering them is an operator action). Custom hosts need
  no per-domain provider entry.

The included reusable on-demand staging tooling supplies these inputs with its own edge (`infra/edge/Caddyfile`, image
`infra/docker/edge.Dockerfile`): `B` and `*.B` by DNS-01 through the pinned Cloudflare module, on-demand TLS only in
the catch-all site behind the `ask` endpoint above, the original `Host` passed with `X-Forwarded-*` removed, only the
two MCP paths routed to core, and the whole query redacted in the access log. It is staging tooling, not a production
edge or deployment claim, and it has not been exercised against any provider for this export; see [STAGING.md](STAGING.md).

## Restore and configuration changes

Domain rows, platform labels and slugs are durable data and restore with the database ([RECOVERY.md](RECOVERY.md)). After a restore or a
`PLATFORM_BASE_DOMAIN` change, platform hosts follow the configuration, custom rows keep their stored proof and state,
and nothing stale can route a host to another Organization (DOM-9). Recheck restored custom rows before relying on
them: DNS and certificates may have changed since the backup. Migration 00011's down refuses while any domain row,
slug, tenant-host login transaction or domain audit event exists. Migration 00021's down (development and test only)
refuses while a live or re-mintable signup continuation references a platform label ([MIGRATIONS.md](MIGRATIONS.md));
after it the slug routes again.

## Records and logs

Audit (trigger-owned, [AUDIT.md](AUDIT.md)): `domain.claimed`, `domain.state_changed` (reason = new state),
`domain.checked` (reason = outcome, including the withdrawal outcomes) and `organization.slug_changed`; denied
operations write `access.denied` with the `domain.*` and `organization.set_slug` operations. Logs: `domain.request`
(operation, closed outcome and reason) and `auth.relay` (outcome, closed reason). Hostnames appear in logs only as the
first 12 hex characters of their SHA-256; challenge values, `state`, `code` and cookies never. That tag is
pseudonymous, not anonymous: it is an unkeyed hash prefix, so anyone can compute it for a guessed hostname and link log
lines to it. Treat logs carrying it as containing the hostname, not as privacy-safe. The default-tenant settings add the startup records
`default_tenant` and `platform_root_page` and the `core.exit` category `default_tenant_unavailable`
([above](#default-tenant-root-alias-and-root-presentation)); neither records a malformed configured value.

## Test overlay

`infra/compose/identity.yml` runs core and web with `PLATFORM_BASE_DOMAIN=localhost` (web derives the platform root
of a signup tenant continuation hop from it), core also with `PLATFORM_HOST_PORT=543`,
`PLATFORM_AUTH_HOST=auth.localhost`, `CUSTOM_DOMAIN_CNAME_TARGET=edge.localhost` and the DNS/TLS fixture adapters
(`infra/identity/{dns,tls}-fixture.tsv`). `infra/identity/mcp-seed.sql` seeds the platform labels `org-a`, `org-b`,
`org-c` (fixed synthetic test labels; production labels are always generated) with the friendly slugs `alpha-agency`,
`bravo-agency`, `charlie-agency` (so `alpha-agency.localhost:543` is the non-routing slug host), the
routing custom host `portal.acme.test` and the removed `old.acme.test` (Organization A) with one historical MCP
connection; `infra/identity/mcp-hosts.tsv` is only the harness's host-to-UUID map. The emulator's redirect list names
the seeded test hosts and `localhost` (the root's own signup callback) and `*.localhost` (a tenant created
during the run, unknown in advance); core alone builds redirect URIs from the resolved host, so the list is only the
emulator's open-redirect guard and is not evidence of real WorkOS wildcard behaviour. The marketing stand-in host is
gone: `https://localhost:543` and `https://www.localhost:543` are the public entry, and a newly created
`<label>.localhost:543` reaches web through the existing `*.localhost` proxy pattern with no DNS step (fixture-proven only).

`.env.example` still says "Tenant hosts are `<slug>.PLATFORM_BASE_DOMAIN`"; that comment predates the opaque platform addresses and is kept
unchanged because its reviewed block is hash-bound (B-ENV-EXAMPLE). This runbook is authoritative: tenant platform hosts
are `<label>.PLATFORM_BASE_DOMAIN`. Likewise its sentence that the platform domain serves the landing predates the root host settings:
the root's `/` follows the two root host lines at the end of that block (section above).

## Evidence and limits

| Check | What it proves |
|---|---|
| B-DOMAINS-CATALOG (`TestDomainsCatalog`) | the version-11 and version-12 catalogs, readers, policies, guard and down refusals, and the legacy-session transition to version 12 (CT-DOM-12) |
| B-DOMAINS-IT (`TestDomains` DOM-1..12) | lifecycle, guard edges, races, generation fencing, withdrawal retry, audit rows, restore/configuration mismatch and the `E-DOM-F` footprint; DOM-11 (a slug change keeps the label, its routing and sessions) and DOM-12 (the label routes to the exact UUID; slug, unknown, reserved and suffix-confusion hosts route nowhere) |
| B-PLATFORM-ADDRESS-CATALOG (`TestPlatformLabelsCatalog`, `TestPlatformLabelsMigration`) | the version-21 label column, CHECKs, index, the two column grants, the writer-only policy, the three definer bodies under their real roles, and the 00021 down refusal for live or re-mintable signup state |
| B-UNIT-CORE (hostname, domains, relay) | normalization, classification, transition table, DNS proof, TLS adapters, relay refusals |
| B-IDENTITY-IT (`TestCustomDomainRelay` and the callback suites) | relay consumption, `host_changed`/`organization_not_member`, removal during the provider exchange |
| B-IDENTITY-E2E (`domain-platform`, `domain-custom`) | E-DOM-1..6 through the TLS edge: platform and custom sign-in, relay hops, unknown and removed hosts, the public entry (E-DOM-4: root and `www` landing, one Sign Up to the root `/signup`, no redirect hop); E-DOM-6: the friendly slug's host is the fixed 404 with no Location |
| B-IDENTITY-E2E (`signup-light`, `signup-dark`) | E-SUP-1..6: root and `www` landing, fresh and existing identity into a new tenant host, field errors, the automatic continuation with its captured-secret, recovery and no-re-mint-on-GET cases, first use without onboarding |

Limits: no live DNS, real certificate, WorkOS wildcard or production edge evidence exists locally; those are the
deployment prerequisites P-1..P-4 (P-1 the WorkOS wildcard redirect and sign-out entries; P-3 one bounded TXT/CNAME proof at a
domain the operator controls, which also settles the CNAME-target chain outcome above). Live custom-domain and
identity-provider qualification has not been executed for this export. No IDNA, no automatic reconciler, no authority cache, no automatic MCP revocation on alias
change. A session issued in the instant between the issuance guard and its commit is refused at its next request (the
authority is request-bound, not instant).

**Organization branding.** Branding follows the Organization, not the hostname: every routing platform or custom host of an
Organization shows the same public branding, and an unrouted, removed or unknown host keeps the fixed 404 (the public
projection and logo route resolve the host first). A logo URL carries only the content version; it is valid at a host
only while that host routes to the logo's Organization.
