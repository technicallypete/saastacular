> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G062 — Root-host tenant provisioning signup work specification

Goal: [G062](../goals/G062-root-host-tenant-signup.md).
State: Provisional planning RELEASED by owner, 2026-10-20; implementation held.

## Purpose, authority and continuity

The owner requires a generic application-root landing with **Sign Up**, leading
to `/signup`, WorkOS authentication and creation of a new tenant with its creator
as initial Organization owner. This is reusable foundation behavior; Beginify
agencies and SaaStacular tenants are examples, not distinct provisioning models.
It creates an Organization; [G058](G058-generic-organization-enrollment-spec.md)
continues to admit users to an **existing** Organization under its current policy.

Read AGENTS/CLAUDE, Architecture §§3/50/51/53/60/63, SECURITY, DECISIONS,
ADR-003/004/007/008, accepted G016/G018/G019/G058/G060 Plans/Results and the
current roadmap/progress before later planning. Preserve global canonical UUIDs,
RoleBindings as sole role authority, core-owned facts/auth/session/audit, RLS and
thin Next.js BFF. Concrete modules, schema, route contracts beyond `/signup`,
lifetimes, limits and form order belong in Claude's reviewed Plan.

This prospective owner change supersedes **only** G016 R2/Architecture §51's
requirement to redirect application root/apex/`www` public entries to marketing.
It does not revoke accepted G016, change its ordinary tenant-sign-in membership
checks, weaken domain/protocol/callback isolation or transfer its implementation
to this Goal. G058's exclusion of public Organization creation remains true for
G058; G062 owns this separate addition. Preserve all historical artifacts and
append/link the new decision in live architecture/security guidance with adoption.
G057 retains required real-provider/HTTPS qualification, reconciled against the
accepted authentication delta before G021. No live action is authorized here.

## R1 — Explicit platform root and separate tenant entry

Serve the deployment-configured platform application root as a minimal generic
landing page with one **Sign Up** action to same-origin `/signup`. No Sign In,
tenant picker, membership directory or tenant/account lookup is offered at the
root. Tenant sign-in remains available exclusively at the intended tenant
hostname through current G016/G017 behavior; protocol callbacks/logout returns
are not general sign-in pages. The root selects no default or seeded Organization.

The required local example is exactly `https://localhost:543` →
`https://localhost:543/signup` → WorkOS → newly provisioned tenant hostname and
onboarding completion. Keep HTTPS and configured `:543` throughout local origin
construction; a name resolving to loopback is not permission to weaken cookie,
callback or TLS checks. `https://org-a.localhost:543` remains the seeded existing
tenant, with independent current membership and sessions. Do not redirect the
application root to the marketing site or show that tenant's landing there.

The independently deployed marketing site stays separate. Illustratively,
`https://foundation.test` links signup to `https://app.foundation.test/signup`, while the
application root serves its own landing. Product/domain strings are configuration,
never code constants. Explicitly configured root/`www` public aliases preserve
this platform-entry contract, with exact callback origins separately validated.
Registered callback, sign-out, MCP metadata/transport and other protocol routes
keep their accepted route-specific behavior; no catch-all landing intercepts them.

## R2 — Platform authentication grants only provisioning intent

Use WorkOS through the existing core-owned identity adapter. A new principal may
receive its canonical global identity/mapping; an existing verified principal
retains the same UUID and can create a new Organization. WorkOS may authenticate
that existing user inside this provisioning intent without a central Sign In UI.
Never promise a new provider account on every signup, merge by email or derive authority from WorkOS
Organizations/roles, typed email, a tenant name, hostname or client claims.

Define a distinct platform signup intent with short-lived, browser-bound, one-use
authentication state and PKCE S256. Core validates the actual provider response and the selected protocol’s
issuer/client/subject/time and trusted identity guarantees through the accepted
ADR-003/G058 adapter. Do not invent an ID token/signature requirement for an
authorization-code exchange that provides identity through the trusted SDK/API;
qualify any necessary projection rather than inventing provider evidence.
Missing or mismatched proof fails closed. Preserve optional MFA and unknown
per-session assurance; do not introduce mandatory MFA enrollment for signup.

Platform signup state/credentials permit only the caller's bounded tenant-create,
own outcome/resume and completion operations. They are **not** an Organization
member session, TenantContext or MCP credential. Ordinary tenant HTTP/MCP/data
boundaries reject them, including where this identity already belongs to a tenant.
They confer no admission, reinstatement, role elevation or Group/profile access
in an existing Organization. A normal tenant session is not interchangeable with
the explicit platform signup intent. G058's existing applicant credential is
Organization-bound; reuse of helpers is allowed, reuse as platform authority needs
explicit reviewed design. No broadening of ordinary tenant login is permitted.

Bind state and browser intent to the exact configured platform origin and finite
signup purpose. Reject expired/replayed/missing/foreign-browser state, changed
host/scheme/port, wrong purpose and forged forwarding headers. Recheck current
canonical account eligibility at provisioning and target-host issuance, including
disable/retirement races. Provider/network work holds no database locks. Provider
cancellation or failed verification creates no tenant or usable tenant session.

## R3 — Minimal tenant name and address; constrained bootstrap

Collect only the tenant display name and its platform-subdomain address/slug
needed for setup. Here “address” means the tenant URL; no postal/billing identity
or commercial profile is required. Validate bounded input and the actual G016
normalization, suffix and reserved-host rules. The Plan may recommend a suggested
editable slug and a compact form sequence; these are routine UX choices, not new
product commitments. Availability must be revalidated at commit, not trusted from
a preview. No caller can submit an arbitrary Organization UUID, owner user, role,
custom hostname, routing proof or ready/policy flag as authority.

At the authoritative transaction, create exactly one new canonical Organization,
exactly one initial `org_owner` RoleBinding for the verified enabled creator, its
safe platform-subdomain routing identity and required audit/one-use outcome state
atomically. A successful tenant cannot be ownerless or route to another tenant.
Database uniqueness/reserved checks and current owner eligibility survive
concurrency; retain G019 lock order/at-least-one-owner and G060 default-false
management override. New Organization enrollment policy remains the accepted
closed default; any fresh-install readiness fact follows G058 trusted semantics.
Creation grants no existing-tenant privileges or Group memberships.

`core/internal/tenancy/membership.go` currently exposes a privileged
`CreateOrganizationWithOwner` service method with no production caller/route;
the runtime role has no Organization INSERT. The Plan must define a constrained
production provisioning boundary, its SQL/runtime trust limits, exact grants,
guards, actor derivation and atomic mandatory audit. Reuse its invariants rather
than exposing the privileged pool or trusting a client-supplied owner. **No blanket
runtime Organization INSERT, owner/migrator/superuser credential in application
services, BYPASSRLS, generic membership writer or unauthenticated bootstrap path.**
A bounded function/guard or equivalent requires security review, including real
runtime direct-SQL refusal cases and honest limits of trusted-core provider proof.
Any need for a new credential/pool or materially different security model returns
to the owner; it is not an implied implementation choice.

Platform-subdomain routing uses the configured suffix and accepted platform-owned
mechanism. Do not auto-verify an arbitrary custom domain, bypass DNS/TLS readiness,
reserve another party's host through pending claims or introduce customer apex
support. Later custom domains use the existing verified lifecycle independently.

## R4 — Idempotency, failure and fresh target-host completion

Bind one finite create intent/idempotency outcome to the verified canonical user,
browser/signup purpose and submitted tenant details. Duplicate submits, concurrent
callbacks/create attempts, refresh, network loss and commit uncertainty cannot
create multiple tenants for the same intent, duplicate initial owners/audit events
or consume an unrelated person's result. A different explicit create intent may
create another tenant; do not impose one tenant per global identity. Competing
distinct intents for the same slug have one winner and a safe actionable conflict.
Changed details on an already-used intent cannot silently create a second tenant.

Document finite invalid-input/collision/authentication/cancellation/expiry/rate/
unavailable/unknown-outcome states and bounded retry/resume. A failed precommit
bootstrap leaves no partial tenant, owner or routing claim; expire/clean transient
signup state through existing retention controls. Provider-created identity may
remain after local failure; do not delete an external account to simulate rollback.
A committed tenant with a failed browser return remains owned and recoverable,
not falsely rolled back or automatically deleted. Resolve unknown commit outcome
through a guarded own-result read before retry; never label uncertainty success.

After committed creation, derive the target from the canonical created tenant and
current trusted configured hostname. Establish fresh target-host browser binding
and one-use handoff under an explicitly reviewed mechanism, then issue a normal
G016 origin/Organization-bound session only after current host, owner membership
and account checks. Creation success alone is not proof of session issuance.
No copying root cookies to the target, shared-domain session cookie, reusable
bearer in redirects, caller-provided return URL or unvalidated custom destination.
Preserve accepted callback relay isolation: relay hosts gain no application session.
Scope every completion reference to exact user, target origin, new Organization,
purpose and expiry; reject foreign-browser/replay/changed-host use.

Completion enters the new tenant UI and reports minimal tenant setup complete.
Do not force Group/Business Profile creation, billing or Organization-management
settings to finish tenant signup, including through an inherited automatic setup
redirect. The owner can subsequently choose the existing
independent G018 Group/profile flow, with its own permissions and completion rules.
New Organization ownership alone grants no G060 Group-content access.

## R5 — Privacy, abuse, UI and operating continuity

Keep decisions in core; BFF/browser presentation never supplies trusted identity,
tenant, role or routing state. Strictly bound bodies, fields, state lifetimes,
retained results and retry/status traversal. Use existing rate-limiter machinery
correctly at auth start/callback/provisioning/own-status boundaries: pin trusted
keys and finite capacity, prevent key churn from replenishing exhausted buckets,
and limit authenticated tenant-create attempts rather than allowing a new client
idempotency key to bypass throttling. Do not trust arbitrary forwarding/IP headers
or add a new vendor/captcha/abuse platform. State limits and multi-process limits
honestly; no claim of comprehensive abuse prevention.

Prevent identity/membership/tenant enumeration beyond necessary public address
availability; private failures never expose another tenant's UUID, owners or
membership. Use closed safe messages, CSRF/browser intent protection, no-store and
no-referrer on private/auth transitions. Never log raw code/state/callback queries,
signup/session/handoff credentials or provider payloads. Attribute successful
creation/initial ownership and relevant security failures with minimal trusted
actor/Organization/request identifiers under current audit rules. Mandatory audit
and creation commit together; preserve unknown commit and audit-failure semantics.

Use existing shell, locally bundled assets, semantic tokens, Radix components and
repository Vestive guidance (not included),
subject to accepted architecture precedence. Retain Light/Dark/System, keyboard,
focus, reflow and field-error association. No marketing asset/copy import, new
design system or extra onboarding screens. Keep source and observed evidence
separate from proposed behavior in adopter documentation. Include necessary
forward migration/grants/RLS/catalog, restore/invalidation and transient-state
retention changes; do not edit historical migrations or fabricate readiness.

## Acceptance criteria and required validation

| ID | Measurable outcome | Required focused proof |
|---|---|---|
| AC1 | Platform landing and `/signup` work without selecting a tenant | Browser at `https://localhost:543` sees only Sign Up as its authentication/acquisition action, navigates to exact HTTPS `/signup`, and does not redirect to marketing; seeded `org-a.localhost:543` remains a separate tenant. Configuration rename/re-domain and protocol-route regressions remain valid. |
| AC2 | New and existing verified identities can explicitly create a tenant | Provider-adapter/emulator positives for new mapping and existing canonical UUID, actual verified proof and optional-MFA semantics; missing/forged/mismatched proof and cancellation cannot provision or issue a tenant session. No email auto-link. |
| AC3 | Platform signup has no existing-tenant authority | Real core/HTTP/MCP negatives reject signup state as a normal session; G058 policy and ordinary tenant nonmember login still refuse unauthorized access. Existing user's prior tenant/Group roles are unchanged. |
| AC4 | Bootstrap is atomic and least privileged | Exactly one Organization plus creator `org_owner`, routing identity and mandatory audit under runtime production path; injected write/audit failure leaves none. Real runtime SQL cannot directly insert Organizations/owner bindings or substitute actor/role/tenant. Catalog/grant proof and forward upgrade/restore disposition. |
| AC5 | Tenant address is canonical and collision safe | Invalid/reserved/root/`www`/infra/suffix-confusion addresses refuse; two concurrent distinct intents for one slug yield one winner; custom hosts never auto-verify. Created platform host resolves only the new UUID under current G016 rules. |
| AC6 | Current eligibility and retries remain truthful | Deterministic disable/retirement race, duplicate/concurrent same-intent request, changed-payload retry and lost-response/unknown-commit recovery; one durable outcome/owner/audit, safe conflict and no unintended second tenant or false rollback. |
| AC7 | Completion establishes a fresh tenant-host session | Actual local HTTPS browser flow root → provider/emulator → committed tenant → trusted new host → completion; wrong browser/origin/port/purpose, expired/replayed handoff and changed host refuse. Cookies stay host-only; no root/relay tenant session, open redirect or cross-tenant replay. |
| AC8 | Minimal tenant completion preserves downstream boundaries | Tenant UI is usable without billing, Group/profile setup or management override; optional later G018 flow uses accepted current permissions, and G060 Group-content isolation/default-off policy remain. Focused accessible light/dark/error/resume browser checks and changed-flow feedback recorded honestly. |
| AC9 | Bounded public behavior and privacy survive failure | Input/state/status/rate bounds, exhausted-bucket/key-churn/provision retry cases, safe errors/CSRF, redacted capture/log checks, transient expiry/cleanup and failed provisioning recovery. No secrets or foreign identity/member data in results. |
| AC10 | Integrated evidence and qualification obligations are attributable | Result maps source/inputs/commands/outcomes and ACs; source/Duo/MR fast, coordinator-owned exact merged-SHA full local evidence and scoped cleanup close. G057 reconciliation identifies actual signup/provider/HTTPS proof still required or observed before G021, reusing unaffected evidence without live claims from fixtures. |

The later Plan maps each AC to reused, extended or necessary new executable
selectors: existing identity/callback, G016 routing/HTTPS, tenancy ownership,
runtime-RLS, audit, G058 limiter/proof, G018/G060 and browser harnesses. Add only
distinct missing boundaries, not another general auth/provisioning test framework
or role×host×theme campaign. Prove the risky signup/provisioning/handoff slice
before broader UI composition. Bind local browser evidence to real navigated
origins; a Host-header-only probe is insufficient.

Retain focused local proof + source/Duo/feature-MR fast → merge → existing
coordinator-owned full local qualification of the frozen exact merged staging
SHA. Preserve failed/unknown/unperformed outcomes and only rerun affected coverage
after corrections. Required changed-flow feedback follows current governance;
no waiver or manual PASS is inferred. G057's actual provider callbacks/sign-out,
new platform-host scaling and deployed custom-host HTTPS obligations remain
separate: its reviewed Plan reconciles the relevant accepted signup delta and
adds only the affected genuine browser/provider proof, reusing unchanged evidence.
No duplicate live campaign or automatic staging cloud pipeline is introduced.

## Expected artifacts and gates

Later Claude Plan: concise feasibility assessment of this provisional B5 scope;
exact routes/files/contracts; platform credential trust and lifecycle; production
provisioning/grants; transaction/lock/idempotency order; callback/target-host
mechanism; migration/restore/audit; finite errors/rates and AC selectors. Obtain
explicit security/architecture review before execution; unresolved material
choices return to the owner rather than silent fallback. Routine layout and
name/slug form sequence remain Plan choices.

Expected execution artifacts are approved source/migrations/tests, updated live
architecture/security/decision and adopter guidance, goal-specific Plan/history
and Result/evidence/cleanup receipt. Do not rewrite closed Goal evidence. The
owner-directed sequence is G055 → G062 → G020 → G057 → G021. The owner expressly
released provisional G062 planning now, without waiting for G055's implementation
MR or final ACCEPT; G055 is a sequence predecessor, not a technical prerequisite
for planning. G062 ACCEPT precedes G020 execution. Implementation remains held
pending Plan ACCEPT and separate supervisor release. This planning authorization
releases no external actions. Context
150k/200k, provisional B5/checkpoint3 and the G021 product barrier remain binding.


## Prospective platform-address amendment — owner, 2026-10-23

G057's opaque-address amendment (not included)
prospectively replaces friendly-slug platform routing with a generated stable
opaque label directly mapped to Organization UUID. Friendly slugs remain metadata;
no routable slug alias/redirect or production compatibility period is required.
The change is implemented and qualified under G057's later reviewed Plan/release,
including affected signup/URL/seed consumers. This Goal's historical acceptance
and evidence remain unchanged; current code is not claimed to have changed.
