# Architecture Decisions

> **Editorial note (public export).** This historical decision record is retained for its reusable architecture and security reasoning. Links to private planning, review and evidence records that are not part of this repository are marked "(not included)"; references to "Architecture §N" or ARCHITECTURE.md point to the source project's architecture document, which is not included. Goal IDs, dates, approvals and superseded history are kept as written; product-specific passages are redacted where marked "[Editorial redaction: …]". Current behaviour is documented in the foundation guides under [`docs/foundation/`](foundation/).

Accepted decisions refine the source project's architecture document (not included) for their stated scope/effective transition. They do not grant implementation approval or overwrite historical Goal evidence.

| Decision | Status / date | Scope |
|---|---|---|
| [ADR-001 — Vestive component foundation and successor](adr/ADR-001-radix-vestive-components.md) | Original and G067 decisions retained; owner G069 amendment ACCEPTED 2026-10-26 | Complete unchanged kit source; adapt application and standard tenant themes. Base UI remains primary, with cmdk-required transitive Radix allowed. Target implementation requires reviewed G069 Plan/release; current accepted app is unchanged until cutover. |
| [ADR-002 — GitLab hosted CI runner](adr/ADR-002-gitlab-hosted-ci-runner.md) | ACCEPTED by product owner, 2026-10-06 | Medium hosted runner with privileged outer job/DinD in an ephemeral isolated job VM; nested validation/application/test containers unprivileged, job-private TLS, separate execution/action gates. |
| [ADR-003 — WorkOS identity provider](adr/ADR-003-workos-identity-provider.md) | ACCEPTED by product owner, 2026-10-06 (2026-10-07 UTC); assurance and optional-enrollment clarifications approved 2026-10-07 | WorkOS AuthKit behind core-owned IdentityProvider; canonical UUIDs and PostgreSQL application sessions retained. MFA enrollment is optional for clients, consultants and agency users; required-MFA staging qualification is distinct from production policy and per-session assurance, which stays unknown without reviewed trusted proof; requests requiring verified MFA fail closed. Revised G004 planning only; implementation, staging account/secret actions and production rollout require their separate gates. |
| [ADR-004 — Group-owned access to one Business Profile](adr/ADR-004-group-business-profile-cardinality.md) | ACCEPTED by product owner, 2026-10-08 | Organization → Group tenancy; one required profile per completed Group, shared across same-Organization Groups. Creator is Group owner; independent BusinessProfile roles are removed and profile access derives from current Group/Organization roles. Global users may manage multiple tenant-owned profiles; UX visibility follows active tenant and authorization. G005/G006 Plan reconciliation, G007/G008 isolation and G018 completion; active G004 execution unchanged. |

| [ADR-005 — Tenant-bound external MCP connections](adr/ADR-005-tenant-bound-mcp-connections.md) | ACCEPTED by product owner, 2026-10-11 | Each tenant-specific external MCP connection is immutably bound to one canonical Organization UUID; another tenant requires another authorized connection. Retain optional connection-local Group switching and per-execution authorization. G014/G015 implement; G016 adds verified hostname aliases. Historical accepted artifacts and active G007/G051 bindings remain unchanged. |
| [ADR-006 — Dedicated PostgreSQL audit schema](adr/ADR-006-audit-schema-boundary.md) | ACCEPTED by product owner, 2026-10-14 UTC | G011 introduces `audit` in the existing database, with separate ownership/grants, controlled append-only writes, authorized tenant reads/RLS and restricted maintenance. No runtime update/delete/truncate/DDL or premature schema implementation; active G008/G009 contracts unchanged. |
| [ADR-007 — Verified domains and tenant-host authentication](adr/ADR-007-verified-domain-routing-and-authentication.md) | ACCEPTED by product owner, 2026-10-16 UTC | G016 selects Caddy target, membership-bound tenant-host sign-in, shared WorkOS callback relay, exact-host MCP retention and bounded lookup-risk measurements. Design only; live provider/settings/provisioning/deployment actions and execution release remain separate. |

| [ADR-008 — Explicit Group authority](adr/ADR-008-explicit-group-authority.md) | ACCEPTED design by product owner, 2026-10-19; implementation pending G060 | Read-only Organization oversight, explicit Group roles for content, default-off owner-controlled ordinary management override; G060 before G055 completion and SaaStacular. |
| [ADR-009 — Reusable on-demand GCP staging](adr/ADR-009-on-demand-gcp-staging.md) | ACCEPTED direction 2026-10-26; code/live actions await release | Generic Terraform + GCP VM/Compose staging, Cloudflare DNS, Caddy/Let's Encrypt, explicit stop/reset/retire lifecycle; G057 Part B consumes this environment. Managed PostgreSQL/R2 are a future production target only. |
| [ADR-010 — Optional default-tenant root alias](adr/ADR-010-default-tenant-root-alias.md) | ACCEPTED owner direction 2026-10-26; G071 definition, not implementation | Optional existing Organization slug selects the exact-root tenant experience after restart; unknown warns/falls back. No seeding, backend provisioning restriction or tenancy-model change. Active alias gives root `/signup` 404; `/join` stays, unset/unknown restores signup. G071 BEFORE G070; G059 remains deferred. |

ADR-005 owner amendment, 2026-10-15: retain WorkOS login/MCP issuance, accepting
broader same-user/client credential reach while preserving local tenant/Group
authority. Independent refresh-family-per-connection isolation is superseded;
supported compromised-client grant/family revocation plus local invalidation
must be qualified before G014 implementation. See the dated ADR/Spec amendment.

ADR-007 owner security amendment, 2026-10-16: tenant-host browser sessions bind
to issuing exact origin and canonical Organization, preserving global identity;
unverified pending claims do not globally reserve hostnames, and exclusivity is
acquired atomically at valid proof. G016 requires bounded Plan/migration review
and focused evidence before execution release; live-action gates remain.


Owner qualification sequencing amendment, 2026-10-16: G016 → G057 (not included)
transfers P-1/P-2 actual WorkOS/deployed HTTPS qualification to a required Goal
before G021. ADR-007 behavior is unchanged; G016 retains implementation/local
security, planned test extension and required MR/staging evidence. G017–G020
may proceed after their normal dependency ACCEPTs without live qualification.
No live pass, infrastructure authority or production-readiness claim is inferred.


G017 owner role decision, 2026-10-16: **“017 owners and admins.”** The planned
branding capability maps to current Organization owner/admin roles under the
existing tenant/security checks; Group authority alone does not qualify.
Spec and decision (not included).
No separate ADR, new role or G017 execution release is implied.


Owner ownership-cardinality decision, 2026-10-16: after proposing exactly one
owner, the owner accepted **one or more owners, with at least one required** and
instructed “write this in for 18 and 19 goals and specs.” Creator is the initial
owner with admin powers; last-owner ordinary retirement needs replacement or
atomic transfer. Emergency suspension is immediate, with separate verified
recovery and no automatic promotion. Architecture §53.7 and G018/G019 record this
superseding policy. G019 definition is provisional; planning/execution still held.

G018 owner decision D-1, 2026-10-17: the owner answered **“Yes”** to the
recommendation that only **Organization owners and admins** may create Groups.
Current active-Organization membership and the guarded `group.create` capability
are required; a Group role alone is insufficient. The verified creator becomes
the new Group's initial owner, without an Organization grant. See the
Spec decision (not included).
This resolves the role choice only; Plan corrections and execution gates remain.

**Superseding G018 D-1 clarification, 2026-10-17:** after that initial answer,
the owner confirmed generic Group self-onboarding [Editorial redaction: the source product's meaning of a
Group was described here]. Permit `group.create` for
current Organization owners, admins **and ordinary members**, creating a new
Group with the verified creator as its initial owner. No Organization-admin role
or access to other Groups follows. Public signup/enrollment separately establishes
ordinary membership under reviewed policy; G018 never infers membership from
login/hostname. The former owner/admin-only choice above is historical and
superseded. See the current Spec amendment (not included).
Revise the pending Plan; no implementation, new public enrollment surface or
G019 continuation is released.

**Generic public enrollment belongs in the foundation — 2026-10-17:** the owner
asked why public enrollment was deferred, then approved “Yes make these updates”
to a bounded new Goal after G019 and before G021. G058 Goal (not included)
and Spec (not included)
own invite-only/request-access/open signup, closed defaults, policy-authorized
ordinary membership and return to G018's own-Group/Business Profile flow. Reuse
G019 grants/invitations and existing WorkOS/auth/session protections. The narrow
applicant boundary is recorded in Architecture §53.11 and ADR-007; authentication
alone remains insufficient for tenant access. G021 requires G058 final ACCEPT
and SaaStacular includes that accepted generic behavior. [Editorial redaction: later product-specific Goals are named here; they retain only
product/commercial composition of these primitives.] Definition/docs only: no
G058 planning/execution or live action; current G018 work, its owner check-in and
the G019 stop remain. No existing automatic-delivery range is enlarged.

**G019 Spec finalization after owner check-in — 2026-10-17:** after provisional
G018 Plan r5 ACCEPT, the owner directed “Do that” to reconcile/finalize the existing
G019 Spec. This lifts the former stop before reassessment for definition only;
G019 planning/implementation remain held. Preserve ordinary member administration,
scope-owner-only succession, at-least-one-owner under runtime/concurrent/cascade
paths and immediate emergency disable. Ordinary Organization departure removes
dependent Group grants after required succession; readmission cannot revive them.

The owner explicitly selected **“Yes—support first-time invitees; complete their
sign-in and admission in G058.”** G019 owns invitation intent and the verified-subject
atomic membership consumer; G058 owns provider recipient proof/public admission and
session issuance. No prior account is required to issue an invitation. Delivery
email alone is not proof or an identity-merging key; the current adapter has no
trusted-email projection, so the G058 Plan must pin actual verified provider evidence.
No new vendor, email service, applicant-session bypass or public Organization signup.


**G019 bounded invitation claim helper — Codex security design disposition, 2026-10-17:**
Review r2 (not included) accepts D-14a's
separate claim_writer NOLOGIN helper as a bounded G019 design, with no credential,
pool, runtime membership or DML expansion of existing readers. Exactly two
kind-split consumption functions and one authorized-target lock helper are proposed;
precise ACL/audit corrections and whole-Plan approval remain pending. This is a
forward design disposition, not an implemented role or owner policy change.
Organization consumption stays owner-executable only for synthetic G019 contract
evidence; production proof/execution authority is an explicit G058 readiness gate.
No application owner login or future consumer grant is approved here.


**G058 enrollment package selected by owner — 2026-10-18:** the owner answered
“I'll take your recommendation” to D-4r1 durable restrictions, D-15(i) existing
trusted-core/runtime first-time invitation execution, and D-16(c) fail-closed
open-mode readiness for unresolved legacy departure history. The core adapter
supplies verified provider address evidence; SQL does not independently prove
provider provenance. No new credential/pool. SDK qualification remains held.
Restrictions survive audit retention/replay and lift only under fresh authorized
reinstatement. A trusted explicit readiness fact gates open mode until legacy
history is reviewed/backfilled or verified absent; fresh installations need no
fabricated backfill, and pre-ledger restore re-closes unresolved readiness.
Request-access under fresh admin approval remains available. Client self-signup
remains policy admission → ordinary member → G018 own Group/profile, without
Organization elevation or other-Group access. See the selected
G058 Spec amendment (not included).
Bounded Plan reconciliation/review and execution gates still apply.

## Owner decision — platform-root signup instead of marketing redirect, 2026-10-20

The configured platform root serves a generic **Sign Up-only** foundation landing;
`/signup` uses WorkOS to authenticate/register a global identity, then securely
provisions a new Organization and its initial owner. Tenant sign-in stays on tenant
hostnames, with no central login or tenant picker. See
G062 (not included) and the
architecture amendment (not included).
Supersede only G016/G017 root marketing-redirect behavior, preserving accepted
security/protocol contracts and history. Definition authorized; planning, delivery
scheduling, provider/DNS changes and implementation need their normal releases.


## Owner decision — optional existing-tenant root alias, 2026-10-26

Owner narrows the proposed single-tenant experience to DEFAULT_TENANT_SLUG:
create the Organization normally, configure its slug and restart. An existing
slug activates root tenant entry; an unknown slug warns and preserves normal
multi-tenancy. Organization creation and APIs remain untouched, with existing
permissions. No creation/seeding or single-tenant database mode. The owner
requests a fresh Spec to avoid confusion with deferred G059, and schedules
G071 (not included) immediately after
G070. [ADR-010](adr/ADR-010-default-tenant-root-alias.md) records the bounded root
exception and unchanged authority/identity/other-host contracts. Specification/
scheduling is authorized; planning/code still follow their recorded gates.


### Later G071 order correction — 2026-10-26

Owner: “We should queue this for before 070 btw.” This supersedes the initial
after-G070 instruction: next priority G069 → G071 → G070 before G021. Preserve
existing G070 work through safe checkpoints and its immutable proofs, hold new
phases/integration, and reconcile the optional configuration after G071. Planning
is the next G071 phase; code still needs Plan/release/source gates. Root `/signup`
creates Organizations; `/join` admits members by policy; `/groups/new` is Group
onboarding after admission. No implicit route-meaning change or signup disable
switch is included. Existing G059 deferral and G057/provider/M1 holds remain.


### Later G071 route refinement — 2026-10-26

Owner explicitly keeps `/join` and chooses 404 for root `/signup` when the default
tenant is set and resolves. Apply that host/path allow-list behavior to its web
subpaths and GET/POST entry points; do not repurpose signup as Group creation.
Backend Organization provisioning/permissions remain unchanged. Unset/unknown
restores the normal root signup flow. This supersedes the earlier requirement
for direct signup access with an active alias, preserving the earlier history.


### G071 multi-tenant root-page option — 2026-10-26

Owner adds a selectable signup redirect versus the existing simple landing, only
in multi-tenant operation, preferring signup redirect as the default. Select
`PLATFORM_ROOT_PAGE=signup_redirect|landing`; unset defaults to signup_redirect.
A valid DEFAULT_TENANT_SLUG takes precedence and ignores this page selector;
unknown slug uses the selected multi-tenant entry. Redirect exact-root GET/HEAD
`/` temporarily to `/signup`, never tenant roots, other paths or user-selected
URLs. No automatic authentication/provisioning, new domain or marketing deployment.
G071 remains before G070; its alias signup404/Join/API boundaries remain intact.

## Owner decision — G069 exact kit visuals and contrast waiver, 2026-10-26

Direct owner instruction: “Omg waive the contrast issues... I want it EXACTLY
like the brand kit - no compromises”. This supersedes G069's earlier requirement
to correct kit contrast through app-owned platform tokens or consumer styling.
Remove those visual corrections, retain the exact pinned kit files and treatment,
and use kit compound Select for every dropdown. Record affected measured contrast
outcomes as owner-waived without claiming threshold PASS. Preserve functional
accessibility, form/security/tenant guarantees and authorized tenant customization.
The Spec amendment (not included)
and decision record (not included)
define the scope. The supervisor's preceding hold is lifted; no new Goal or
implementation attempt is created. The resulting executable delta still needs
review and the owner-selected full hosted pre-merge qualification.


**G069 unaltered-kit cutover — implementation record, 2026-10-26:** under the
accepted ADR-001 G069 amendment and Plan r1, the complete Vestive Brand `kit/src`
tree is the application's UI source at `web/src`, byte-verified against the
skill's `LOCAL.sha256`; the application adapts around it with app-owned standard
theme variables, a server-rendered tenant `<style>` with an explicit saved-branding
fan-out, next-themes appearance, and consumer `className` compositions. Radix is
present only as cmdk's reviewed transitive closure. Core palette delta r3 aligns
the platform defaults and contrast pairs with the rendered kit states without a
persisted-contract change. This records the implemented mechanism, not a new
decision; acceptance is recorded only by Codex in the G069 Result and `PROGRESS.md`.
See [WEB-SHELL](foundation/WEB-SHELL.md#unaltered-kit-adoption).
