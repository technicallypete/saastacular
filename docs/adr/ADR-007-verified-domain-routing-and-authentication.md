# ADR-007 — Verified domains and tenant-host authentication

> **Editorial note (public export).** This historical decision record is retained for its reusable architecture and security reasoning. Links to private planning, review and evidence records that are not part of this repository are marked "(not included)"; references to "Architecture §N" or ARCHITECTURE.md point to the source project's architecture document, which is not included. Goal IDs, dates, approvals and superseded history are kept as written; product-specific passages are redacted where marked "[Editorial redaction: …]". Current behaviour is documented in the foundation guides under [`docs/foundation/`](../foundation/).

Date: 2026-10-16 UTC
Status: ACCEPTED — explicit product-owner design approval
Decision owner: Product owner
Recorded by: Codex
Implementation scope: G016 under its accepted Plan, bindings and execution release

## Context and authority

After G016 Plan r3 technical review, the owner answered **“Approve those details”**
to the recommended design package: Caddy target, tenant-host membership, shared
WorkOS callback relay, retained historical MCP connections, and DB-backed hostname
lookups with bounded measurements. The preceding prompt explicitly distinguished
design approval from live DNS/WorkOS/deployment authority. This records that choice;
no provider or production action occurred. Historical G004/G014/G015 evidence stays.

## Decision

1. Select **Caddy on-demand automatic TLS** as the v0 deployment target, with core-
   owned issuance permission and bounded edge TLS readiness probing (G016 E1).
   This is the selected self-hosted integration, rather than a Cloudflare API
   adapter or per-tenant Terraform. Shared infrastructure may use separately
   authorized Terraform; assign each resource to one manager. Domain proof,
   freshness, TLS expiry, canonical ownership and removal govern serving readiness.
   No multi-provider controller or new production deployment is approved here.
2. Require current local Organization membership for tenant-host sign-in. Global
   provider authentication does not grant membership or map identities by hostname.
   An unmapped principal cannot be a tenant member: refuse before creating a local
   user/mapping/session. Existing global canonical UUIDs, identity namespace,
   optional-MFA/unknown-assurance policy and core-owned sessions remain.
3. Keep **WorkOS** as authentication provider. Platform hosts use reviewed provider
   wildcard configuration; custom hosts use the G016 B2 shared callback relay.
   Relay only the bounded provider query to a resolution-derived initiating custom
   host; consume the original browser-bound one-use transaction and PKCE there.
   No relay session/cookie, shared-domain cookie, new issuer or reusable bearer in
   redirects. Recheck original Organization and membership inside final issuance,
   after provider exchange. WorkOS/network calls hold no DB locks. Preserve declared
   READ COMMITTED request semantics: changes after the guard apply through later
   requests/operations, not instantaneous cancellation. Entire callback queries,
   including provider error descriptions, stay out of logs; no-store/no-referrer.
4. Retain historical MCP connection records on domain removal/slug change. Requests
   refuse while the exact stored host no longer routes to the bound Organization;
   never rewrite audiences/resources or transfer tenant/Group state. Other aliases
   need separate authorized connections. Historical rows remain listable/revocable
   through a current authorized host; no automatic revocation is added. Existing
   terminal/expiry/revocation rules remain, including no token-based revival.
5. Accept DB-backed indexed hostname lookup and the bounded PG-007 measurement/
   escalation approach. This accepts proceeding with the stated query/index/lock
   risk, not measured production performance. No authority cache or benchmark
   campaign; material unexpected cost returns for review.

## Consequences and remaining gates

This refines ADR-003/005 and Architecture's earlier provider-selection deferral for
G016 only. Internal-only management and the reviewed provenance transition remain.
One ordinary-query sqlc trial/fallback is within the reviewed G016 scope; Goose
remains migration authority, generated queries use caller-owned guarded pgx work.

Live DNS/TLS proof, WorkOS settings, credentials/account changes, provisioning and
deployment remain separately gated. Fixtures do not prove provider behavior;
AC7/AC8 retain real-provider/manual evidence, and any permitted AC6 unresolved
production-adapter gate is explicit. Plan acceptance is not Goal completion.
R0, exact inputs/scope, initial Result, B7/checkpoint4, resource readiness and a
recorded supervisor execution release precede implementation. G021 stays closed.

## References

- G016 Plan r3 review and owner record (not included)
- G016 accepted Plan (not included)
- [Caddy on-demand TLS permission](https://caddyserver.com/docs/caddyfile/options#on_demand_tls)
- [WorkOS redirect rules](https://workos.com/docs/reference/authkit/authentication/get-authorization-url)

## Owner-approved security amendment — 2026-10-16 UTC

After the G016 H-1/M-1 findings, the owner answered **“I approve”** to the two
recommended corrections in the security disposition (not included).
This supersedes any assumption that host-only browser cookies alone contain
credential exposure from tenant-controlled custom DNS.

Tenant-host browser application sessions bind to their issuing exact origin and
canonical Organization. Core enforces that scope throughout browser credential
resolution and request/context consumption; missing origin/binding, another
Organization or reassigned host cannot expand authority. User UUID/provider
identity remains global, with separate local sessions for other tenant hosts or
aliases through existing WorkOS authentication. Preserve current membership,
browser binding, PKCE, logout scopes and separate MCP connection/resource rules.
This contains cross-tenant replay; a tenant controlling its hostname can still
observe credentials sent to its own server.

Unverified pending domain claims do not reserve a hostname globally. Exclusive
ownership is acquired atomically on successful valid DNS proof, with established
verified/active ownership protected and one winner for concurrent promotions.
Pending rows remain non-routing; preserve challenge ownership, generation/version
fencing, audit and proof freshness. No cross-tenant write privilege, new quota,
reconciler or cache is implied.

The bounded Claude Plan amendment must pin migration/legacy-session handling,
every affected consumer, promotion conflicts and focused negative evidence before
supervisor execution release. This approval is from the owner and grants the
design change, not final implementation acceptance or live deployment authority.


## Owner-approved qualification sequencing amendment — 2026-10-16 UTC

The owner approved the G016 → G057 split (not included).
G016 delivers reviewed implementation, local security/HTTPS/emulator evidence and
the planned provider-test extension, with normal integration/acceptance gates.
P-1/P-2 real WorkOS callback/sign-out scaling and deployed custom-host HTTPS
qualification transfer to G057 and remain required before G021. G017–G020 may
consume accepted implementation contracts without G057, while live guarantees
remain unqualified. This supersedes earlier in-G016 timing only; all security
semantics and the requirement for actual provider/manual proof remain. P-3 is
optional and P-4's explicit production-adapter hold is unchanged. No provider,
DNS/VM/deployment or browser action is authorized by this amendment.

## Owner-approved prospective enrollment amendment — 2026-10-17 UTC

The owner approved moving generic public enrollment into the foundation as
G058 (not included),
after G018/G019 and before G021. This is a narrow prospective exception to
decision 2's refusal before local identity persistence, only for an explicit
verified enrollment intent: minimal global identity mapping/applicant state may
be created before membership. Ordinary tenant sign-in still refuses nonmembers.
Until G058 is reviewed, released and implemented, existing behavior remains.

Pending enrollment credentials grant only own applicant operations, never normal
tenant sessions, TenantContext, HTTP/MCP or tenant data. Preserve canonical UUID
and namespace uniqueness, account-disable, optional-MFA/unknown-assurance rules,
browser-bound one-use state, PKCE and exact-origin/Organization binding. Provider
identity, email/domain or hostname alone cannot create membership. Invite-only,
request-access and open signup apply the current reviewed policy/G019 authority;
open signup grants ordinary membership and never silently reinstates restrictions.

Membership, relevant invitation/approval consumption and mandatory audit commit
atomically before fresh G016 session issuance. Recheck current policy, subject
eligibility, grant authority and host binding; handle races/retries without
claiming an uncommitted or unknown outcome succeeded. Reuse WorkOS/core-owned
adapters, not a new issuer/vendor. G057 must reconcile relevant accepted G058 auth
changes before G021 live qualification closes, reusing unaffected evidence.

This records the owner-approved design boundary, not G058 Plan acceptance,
implementation release or live provider/DNS/TLS/deployment authority. Historical
G016 evidence and its ordinary-sign-in membership guarantees remain unchanged.

### First-time invitation scope clarification — 2026-10-17 UTC

At G019 Spec finalization, the owner explicitly required first-time invitees, with
sign-in/admission completed in G058. G019 records intended recipient/scope/role and
defines atomic verified-subject consumption. G058 obtains the required trusted
recipient proof through the reviewed core identity adapter before invoking it.
Existing VerifiedIdentity has no verified-email projection; any required extension
must pin actual provider evidence/normalization and fail closed on absent proof.
A delivery address, token possession or client-submitted claim does not establish
membership, and email matching never merges canonical identities. This does not
release implementation, weaken ordinary sign-in or authorize live provider actions.


## Owner-selected G058 proof and legacy boundary — 2026-10-18 UTC

The owner selected the reviewed existing trusted-core/runtime execution path
for first-time Organization invitation admission: fresh provider recipient proof
remains core-owned, principal-bound and SDK-qualified; SQL enforces the bounded
claim predicates without independently proving provider provenance. No second
credential/pool or issuer. Durable restriction records preserve departure/denial/
revocation across audit retention and replay. Open enrollment additionally
requires trusted readiness when legacy departure history is unresolved; verified
fresh installs need no fabricated backfill and pre-ledger recovery re-closes
unresolved readiness. Fresh authorized request approval remains available.

This prospective amendment implements the owner's selected G058 Spec package;
it does not weaken ordinary sign-in membership checks or accept pending applicant
credentials as ordinary sessions. Review the bounded selected Plan delta before
execution; SDK/dependency/native/live-qualification gates remain.

## Owner-approved prospective platform signup amendment — 2026-10-20

The owner replaces configured platform-root marketing redirects with a generic
Sign Up-only landing and `/signup` under
G062 (not included). Platform
signup is a separate trusted context for authenticating a global identity and
creating its new Organization/initial owner; it is not a tenant session, existing
Organization enrollment, default-tenant mapping or central sign-in/directory.
Preserve exact-origin/Organization cookies, one-use browser/intent-bound callbacks,
current account status and verified host routing. The transition after committed
creation requires a reviewed short-lived one-use target-host binding and fresh
tenant session; do not share cookies or generalize cross-host redirects. Root,
`www`, CNAME/protocol/infrastructure labels remain reserved. Existing protocol
endpoints keep their handling. G057 must qualify the selected actual-provider
path before G021; this amendment grants definition authority only.


## Owner-approved opaque platform addresses — 2026-10-23

The owner approves G057's bounded platform-address amendment (not included).
A generated stable opaque DNS label replaces the friendly-slug public address;
the indexed resolver maps it directly to the canonical Organization UUID. Friendly
slugs remain metadata, with no public slug alias/redirect. No production-host
compatibility campaign is needed; normal development-data migration is retained.
Existing session/MCP exact-host binding, provider wildcard scaling, custom-domain
verification and real P-1/P-2 qualification remain. This changes future design
scope, not historical acceptance, execution release or external-action authority.
