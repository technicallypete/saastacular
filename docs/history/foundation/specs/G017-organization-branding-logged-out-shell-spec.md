> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G017 — Organization branding and logged-out shell work specification

Goal: [G017](../goals/G017-organization-branding-logged-out-shell.md)

Epic: EPIC-006 — Organization domains and branding

State: Definition reviewed and ready; provisional planning at G016 implementation MR; execution gated

## Purpose, authority and baseline

Provide generic white-label identity and a minimal tenant entry frame using the
accepted design/token foundation. Branding is presentation and intentionally
public information, not evidence of identity, membership or connection authority.
G016 supplies trusted hostname resolution; G017 consumes it without reopening
DNS/TLS, browser authentication or exact MCP audience decisions.

Owner direction, 2026-10-15: “Spec 17 and continue auto for epic 5 and 6.” This
authorizes this definition and the normal bounded lifecycle through the recorded
EPIC-006 endpoint. It adds no G018+ authority or G021 product-feature release.
Provisional planning begins at the actual G016 implementation MR under AGENTS;
implementation waits for final G016 ACCEPT and the normal review/input gates.

Definition baseline: `staging` at `75b73a51939b3ed41ed490e8ac8aefed58e02635`.
Existing `web/lib/tokens.ts` exposes 16 named per-appearance color overrides and
derives internal ink roles; it explicitly delegates untrusted-value validation
to future trusted branding resolution. TokenScope is visual only. G048 delivered
Radix source components and bundled fonts; the preserved Vestive skill's older
Base UI prose is overridden by ADR-001 and accepted G048. G016 is a definition, not
delivered routing. Reconcile actual accepted G014–G016 before implementation.

Canonical inputs: AGENTS/CLAUDE, Architecture §§37A/50/51/61, SECURITY, DECISIONS,
ADR-001/003/004/005/006, ROADMAP/PROGRESS, G016 and accepted predecessor artifacts,
WEB-SHELL, AUTHORIZATION, DATABASE-ISOLATION, OBJECT-STORAGE, AUDIT and PERFORMANCE.
For UI planning/implementation/review explicitly load repository
`web/.claude/skills/vestive-design/SKILL.md`, its README and relevant React-kit
guide. Canonical product/security requirements override marketing copy, Dark-only
guidance, obsolete Base UI requirements and unreviewed source defaults.

## Scope and non-goals

Implement core-owned Organization branding metadata, guarded configuration/reset,
safe logo reference/publication and a finite public projection; minimal authorized
management controls; generic hostname-resolved logged-out shell; authenticated
Organization indicators; token application and tenant-safe response/assets.
Schema changes/migrations are limited to these needs; choose the actual next
migration in the Plan after accepted predecessors, not during definition.

Exclude a landing-page CMS, arbitrary content blocks, HTML/CSS/JS editor, custom
fonts/themes, remote image fetching, agency acquisition/recruitment, services or
workshop pages, marketing testimonials, product taglines, Group/profile onboarding,
new navigation/connection switching, identity/domain lifecycle changes, vendor
selection, production deployment, generic media-processing service and CI redesign.
The generic entry may use existing sign-in/account actions only as authorized;
it must not imply that any visitor can enroll in the Organization.

## Requirements, architecture and security invariants

### R1 — Durable configuration and privileged mutation

Keep one Organization-owned branding configuration keyed by canonical UUID, with
a finite schema: bounded display label, optional validated logo reference and
per-appearance color overrides from the existing semantic vocabulary. No Group,
BusinessProfile or domain alias becomes a separate branding owner or permission
tier. Missing optional branding inherits configured platform defaults.

Core owns validation, current authorization, version/conflict handling and audit;
web remains a thin BFF. Reuse existing role facts, guards and FORCE RLS. Pin the
finite read/update/reset/logo-management operation/capability matrix in the Plan;
the current catalog has no branding-specific mutation mapping. Do not silently
borrow `organization.manage_members` or `organization.manage_policy`, introduce
roles, infer authority from a hostname or give Group ownership tenant-wide powers.
Any new bounded capability/public contract needs explicit supervisor review;
owner-level role-policy choices escalate if not already decided canonically.

Validate all fields before commit. Unknown fields/token names, oversized text or
payload, foreign object references and invalid override values fail with zero
branding mutation/publication. Update/reset is atomic and has explicit stale-edit
behavior; a concurrent stale request cannot silently erase a later accepted
configuration. Denied/failed audit handling follows accepted ordinary mutation
semantics. Reset removes the override without deleting unrelated tenant files.

### R2 — Safe semantic token resolution and user appearance

Accept a closed set of per-Light/per-Dark color values only: existing supported
token names, bounded literal color grammar and lengths pinned in the Plan. Reject
`url()`, `var()`, CSS declarations/selectors, scripts, arbitrary custom properties
and unreviewed non-color values. Do not pass untrusted strings directly into
`tokenOverrideStyle()`. New font, geometry, motion and component settings are out
of scope; keep existing source/semantic/internal-role separation.

Persist only accepted values. For each allowed override, validate the effective
palette after inheritance/derived roles, including the actual shell/control uses.
Reject combinations that cannot meet text contrast (4.5:1 normal, 3:1 large) and
3:1 essential controls/focus against their rendered adjacent surfaces, or retain
the documented safe default for a missing value; never silently publish an unsafe
theme. Pin finite affected pairs/states, including hover and portals, rather than
constructing an exhaustive color-permutation test generator.

Organization branding never sets the user's appearance preference. Retain Light,
Dark and default System; stored preference, explicit choice and live OS tracking
behave as accepted. Apply resolved tenant tokens in server-rendered scope before
themed content paints; preserve the existing pre-paint appearance script and
hydration guarantees. Do not show a known wrong tenant's branding while loading.
Dialogs/overlays remain in their TokenScope, independent of sibling scopes.

#### Owner M1 amendment — compact appearance selector, 2026-10-17

The owner reviewed screenshots and said: “Looks good just change light dark
system control switch to a simple small. Dropdown menu select. The activator
item shows the current selected symbol.” Replace the shared three-radio segmented
control with one compact dropdown/select. Its closed trigger displays the selected
preference's symbol: sun for Light, moon for Dark, monitor for System (not the
currently resolved OS theme). The opened options retain the explicit Light/Dark/
System labels and selected state. Give the trigger an accessible Appearance name
and current preference; decorative icons must not replace accessible labels.

Reuse native select semantics and existing icons/components/dependencies where
practical. No new package, bespoke menu framework, appearance state store or user
setting. Preserve keyboard operation, visible focus, usable touch target, narrow
reflow, stored preference, default System/live OS tracking, pre-paint selected
symbol and theme/hydration behavior. Use current tenant tokens in both appearances.
This changes the shared control wherever the shell uses it, including unbranded
pages. Do not change identity, permissions, branding APIs or unrelated geometry.

Other displayed G017 visuals received positive feedback; this correction and its
focused screenshots/owner feedback remain before closing M1. Update the existing
radio-specific appearance/shell/design/accessibility/spacing expectations to
assert equivalent dropdown guarantees. Run affected selectors only, extending
existing meaningful cases; no blanket browser/ops/full local or CI-controls run.
Current MR checks remain required after the scoped checkpoint push. Record a
short implementation/scope note within the existing Plan/Result; no full replan.

### R3 — Logo bytes, explicit publication and cleanup

Use existing ObjectStore/guarded storage boundaries and canonical same-Organization
metadata. G009 already supports Organization objects with no Group; never invent
a default Group or derive logo ownership from an arbitrary Group. Any new logo
association schema, guarded publication operation or serving route is a bounded
Plan-reviewed delta, not an anonymous exception to generic storage policy.
Pin a closed non-executable image format list, byte and decoded-dimension
limits, actual content validation and a safe response content type/disposition.
No arbitrary SVG/HTML/active payload, external URL import, tenant-chosen storage
key, bucket policy change, public generic object reader or unbounded image decode.
Do not turn filename/content-type claims into proof of image safety or ownership.

Publishing a logo makes only that explicitly selected, validated logo available
through the narrow public branding asset projection for its current Organization.
Ordinary private objects remain private; generic signed-download, deletion and
metadata authorization stay unchanged. Pin how a public request proves the
trusted host/current logo association without accepting a caller-supplied foreign
object identifier. URL knowledge or object prefixes are not authority. No provider
credential or reusable application session appears in asset URLs or client state.

Replacement/removal and interrupted uploads use existing staged lifecycle and
cleanup. A failed upload never becomes the published logo; successful replacement
does not delete a file referenced for another purpose. Specify bounded orphan
handling and truthful deletion-pending behavior; reuse accepted recovery/retention
operations instead of adding a new cleanup service. Broken/missing logo displays
the configured text fallback with useful accessible text, not private object errors.

### R4 — Trusted-host public shell, indicators and cache boundaries

Resolve the active Organization only through accepted G016 core routing. Unknown,
unverified, suspended/removed and spoofed hosts retain G016 refusals; do not use a
default Organization or branding query selector. Apex/`www` marketing redirects
and exact callback/logout/MCP protocol routes remain unchanged.

Expose a closed public branding projection limited to the approved display label,
safe token values and selected logo presentation reference needed by the shell.
Exclude memberships, people/profile/contact details, private policy, provider IDs,
secrets, domain verification data, storage metadata/keys and other tenant records.
Publishing Organization branding does not make Organization APIs anonymous or
disclose which users are members. Escape labels as text; no tenant HTML rendering.

The logged-out shell contains tenant identity, neutral sign-in entry, existing
appearance/accessibility controls and configured platform fallback/attribution.
No product-specific marketing, pricing, signup/recruitment funnel or false signed-in
state. Authenticated indicators distinguish global user sign-in from membership
in the current routed Organization; obtain current membership/context through core.
No global Organization listing, group picker, membership grant or MCP retarget.
Revocation must not leave private identity/context in an anonymous cached response.

Pin HTML/BFF/asset cache behavior so host, canonical tenant, configuration revision
and authentication state cannot bleed across A/B or public/private responses.
Domain removal and branding update/reset must invalidate or version the actual
public projection. No shared cache of personalized identity/permissions. Public
branding is intentionally public: do not promise retroactive erasure of screenshots
or previously fetched logo bytes; bound serving/cache behavior explicitly.

### R5 — Accepted UI composition, bounded cost and evidence

Use existing adapted Vestive/Radix components, bundled Oswald/Barlow/Roboto fonts,
semantic roles and appropriate existing blueprint styling. No wholesale kit import,
new global theme engine or copied Vestive logos/marketing identity. Additional
small source components are permitted only if the Plan identifies need, provenance,
exact dependencies and adaptations. Brand values remain data/configuration, never
hardcoded special-casing of Beginify or a dogfood tenant.

Preserve keyboard/focus/dialog behavior, reduced motion, text spacing, 320px reflow
and accessible image alternatives. Required manual feedback uses the actual tenant
shell in both appearances and representative narrow/wide viewports before final
qualification; focused checks prepare the environment first. Do not rerun an
unchanged full UI matrix for every tenant/theme/content edit.

Assess new branding lookup count/index paths, payload/logo size, token validation
cost and cache invalidation in PERFORMANCE.md when material. Avoid per-component
duplicate queries, domain-wide scans, unbounded asset work or network fetches in
token rendering. Bounded fixture observations suffice unless a concrete query/
contention concern requires escalation. No speculative cache service or benchmark
campaign. Update finite audit/storage/public-surface documentation as necessary.

## Acceptance criteria and verification

| ID | Required outcome | Objective verification |
|---|---|---|
| AC1 | Organization branding has one canonical owner, finite validated state and a reviewed privileged mutation contract. | Runtime-role/RLS and guarded service cases cover approved read/update/reset, ordinary/foreign denial and no Group/profile-derived privilege. Inspect schema/grants and exact operation mapping; invalid/oversized/unknown input has zero persisted effects. |
| AC2 | Update/reset and logo association are atomic, auditable and safe under concurrent stale edits. | Deterministic stale-version/concurrent update/reset cases assert the Plan's conflict outcome and final committed state; invalid/foreign logo and required append failure cannot publish partial configuration. Reset preserves unrelated files; scoped audit contains no secret/private payload. |
| AC3 | Only valid bounded colors enter semantic override resolution and the effective palettes retain contrast. | Focused value/schema negatives refuse CSS/URL/script/custom-property injection; permitted Light-only/Dark-only/paired overrides inherit as specified. Check the finite rendered shell/control/focus pairs in both appearances; invalid combinations cannot become effective persisted branding. |
| AC4 | Tenant branding preserves appearance, first paint, scoped portals and accessibility. | Existing targeted appearance/pre-paint/hydration/portal tests plus one branded scope verify default System, explicit Light/Dark, OS changes and sibling isolation. Focused keyboard/reflow/text-spacing/reduced-motion checks and required manual feedback cover the changed shell. |
| AC5 | Logo ingestion/publication exposes only validated same-Organization bytes and has bounded failure handling. | Existing storage fixtures prove allowed image, oversized/invalid/active content, foreign/private selector and interrupted upload outcomes; actual public asset path returns only the current selected logo with safe headers. Replace/remove/broken-logo cases preserve private objects, fallback and truthful cleanup state. |
| AC6 | Public shell projection is intentionally minimal and host-bound. | Actual HTTPS requests/browser navigation to A/B, unbranded tenant and G016 refused hosts assert exact public fields, escaped label and neutral sign-in shell. Canary private records, secrets, storage keys and foreign branding never enter body, metadata, asset response or captured logs. |
| AC7 | Authenticated context indicators reflect current authorized context without granting authority. | Anonymous, signed-in/member, signed-in/nonmember and revoked-member cases show the declared state using core facts, with no global membership directory or tenant switch side effect. Existing authentication and exact MCP connection regressions remain unchanged. |
| AC8 | Cache behavior cannot mix tenant branding or personalized/private context and respects updates/removal. | Interleave A/B anonymous and authenticated requests through the actual configured cache path; assert declared headers/keys/version behavior. Update/reset, logo replacement and domain suspension/removal show new/refused state within the pinned serving contract, with no cross-host or personalized response reuse. |
| AC9 | Identity/configuration and accepted UI foundations remain reusable with a bounded query/asset footprint. | Existing relevant rename/config checks and source review show configured platform identity, no dogfood special case, no remote font/asset fetch or obsolete Base UI imports. Record actual new query/asset/invalidation cost and dispositions for material risks; no broader infrastructure added. |
| AC10 | Required evidence, integration and cleanup are attributable and complete. | Result maps AC1–AC9 to reviewed/tested source, relevant inputs, commands/outcomes, manual feedback and deviations. Record actual path-selected local/MR/staging disposition and scoped storage/Docker/worktree cleanup; pending/failed evidence stays explicit, followed by separate Codex final ACCEPT. |

## Verification strategy, provenance and expected artifacts

Five scenario groups suffice: (1) finite configuration/privileged mutations and
stale-edit ordering; (2) safe semantic overrides/appearance/accessibility;
(3) logo publication/lifecycle; (4) trusted public/authenticated shell and cache
isolation; (5) configuration/query/integration/cleanup. Reuse two Organizations,
one missing-brand fallback, existing identity/storage fixtures and selected browser
cases. Share setup; do not take a Cartesian product of roles, colors, themes,
hosts and lifecycle states or build a second policy oracle/harness.

Unit tests cover pure schema/color/state functions; real runtime PostgreSQL checks
cover actual guards/RLS/atomicity; actual storage and HTTP/browser cases cover
publication and cache composition. Validate at real hostname URLs for origin/
cookie behavior. Public rendering never bypasses G016; simulated routing alone
is not proof of the accepted host boundary. Existing shared controls are reused.

Claude's concise Plan identifies exact schema/files/contracts, finite update and
public-field matrices, storage/public-serving delta, cache/contrast rules and AC
mapping. State focused local checks and the actual CI profile selected by changed
paths; shared UI/security/browser paths may require broader existing checks, but
do not invent a full-suite-per-case mandate. Complete relevant manual feedback
before required final qualification. Reconcile predecessor changes and tested
content under accepted Git-first procedures; no separate evidence framework.

Expected delivery: reviewed Plan, bounded core branding model/guards/migration,
narrow public projection/asset path, minimal management and tenant shell/indicator
UI, focused tests, relevant existing documentation, Result and cleanup receipt.
Keep compact commands/results/provenance in Git and evidence in existing CI
artifacts; local raw evidence is temporary. Retain failures/unknown outcomes and
required proof holds honestly. New runtime authority cannot be justified solely
by screenshots or passing appearance checks.

## Decisions, deviations and completion gates

Resolve during planning before Plan ACCEPT: exact privileged branding capability/
role mapping using existing roles; finite display/logo fields and allowed color
subset/grammar; concurrent edit semantics; public logo delivery/cache invalidation
and format/size limits; default/fallback and authenticated nonmember presentation.
No mapping is implied by navigation or `organization.read`. If these require a new
security model, public tool, vendor or destructive schema decision, escalate to
the owner rather than silently extending scope.

No blanket deviation waiver. Internal file/layout choices and token subsets may
differ if the finite accepted contract and all outcomes hold. New capability/
public projection changes need explicit review; weakened authorization, unsafe
assets, cross-tenant caches, inaccessible colors, forced Dark or changed first-
paint/portal guarantees are not permitted implementation deviations.

Execution requires final G016 and other dependency ACCEPTs, accepted/current Plan,
decisions/input/budget/resource gates and supervisor release. Completion requires
AC1–AC10, reviewed diff/Result, required staging disposition and scoped cleanup,
then Codex final ACCEPT. This is the recorded EPIC-006 endpoint, not authority to
start G018 or product implementation before G021.

## Query tooling note — 2026-10-16

Follow Architecture §27.1 (not included):
Goose handles migrations; incremental sqlc adoption for ordinary queries is a
separate bounded follow-up, not a required deliverable or new acceptance criterion.
A new ordinary branding-query package could be a later incremental sqlc candidate,
but this Goal does not authorize or depend on the separate adoption trial. Use
accepted repository tooling at planning time; generation cannot bypass privileged
mutation guards, the deliberately public projection, logo ownership or cache
invalidation. Retain the existing feature checks without adding a tooling suite.

## Owner continuation endpoint update — 2026-10-16

Owner **“I think we can authorize autopilot up to 19 spec”** extends continuation
after G017 under the current exact mandate (not included):
G018 may plan at this task’s first implementation MR and follow normal reviewed
delivery; Codex may define G019 after G018 Plan interfaces settle. Stop before
G019 planning/implementation and all G020+ work. This supersedes older G017-only
continuation stops, not this task’s functional scope, budget, review/dependency/
release requirements, live authority or G021.


## Dependency clarification — G016/G057 qualification split (2026-10-16)

Under the owner-approved [split](../results/G016-G057-qualification-split.md),
G016 final ACCEPT covers the reviewed implementation/local security boundary;
P-1/P-2 actual-provider/deployed HTTPS proof moves to G057 before G021. This Goal
consumes accepted predecessor routing/session contracts and does not wait for
G057. Its own manual visual/UX feedback, security, Plan, dependency, input,
budget and supervisor-release gates remain. No live custom-host readiness claim,
new infrastructure authority or waiver of implementation defects follows.


## Owner decision D-1 — Organization branding roles (2026-10-16)

Owner: **“017 owners and admins.”** Resolve D-1 to
`organization.manage_branding` for current `org_owner` and `org_admin` in the
trusted active Organization. The mapping covers branding update/reset, logo
upload/replacement/removal and the scoped branding-upload cleanup route under
the existing guarded object-deletion policy. Group owner/admin/member roles or
Organization membership alone do not grant this capability; an actor holding
one of those roles may act only if it also has a current qualifying Organization
role. Preserve live membership, origin/Organization session binding, tenant/RLS,
CSRF, stale-edit and audit checks; branding never grants identity or membership.

Claude appends the finite role mapping and affected positive/negative assertions
to the accepted r2 Plan/history, preserving its exact imported snapshot. No full
replan is needed. This closes only D-1: G016 final ACCEPT/current-target
reconciliation, input/budget/resource readiness and recorded execution release
remain required. B5/checkpoint3 and the G057 qualification split are unchanged.

## Mechanical correction autonomy — owner direction, 2026-10-17

Owner “Make that change” authorizes the current mechanical-correction policy (not included):
a released implementer handles in-scope selectors/fixtures, approved-value catalog/
count/version pins and harness/import bookkeeping directly, preserving assertions
and recording focused evidence/Plan wording reconciliation in the Result. No
separate approval or full replan for each mechanical inconsistency. Exact scope,
security/contracts, unapproved schema/constraints, budget/native stops, required
validation and final Codex review remain. Apply at the current safe boundary; no
worker restart, new test campaign, frozen-input rewrite or budget reset. For held
Goals this changes guidance only, not execution/live/publication authority.

## Staging repair scope05 — 2026-10-17 UTC

Codex review (not included) permits the exact
boundary.spec.ts appearance-selector correction (scope49→50), existing harness
branding inventories/copy lists and actual320px layout diagnosis/fix. Attempt2/B5,
checkpoint3; no weakened assertions, full local campaign or change to final gates.
The separately identified DNS hop-bound issue is proposal-only until reviewed.

## Owner includes G016 CNAME repair alongside G017 fixes — 2026-10-17

Owner: “The 016 cname issue should be fixed alongside 17 fixes.” Include restoration
of the existing immediate-CNAME-edge/three-hop proof contract in the current G017
staging repair effort. This supersedes any deferral of that correction to a later
roadmap Goal. It does not authorize weakening the TXT challenge/hop bound or
replacing a real fix with serial tests/retries. G016's historical acceptance stays
recorded; the newly verified defect and its correction remain explicit.

Use the current G017 correction attempt2/B5/checkpoint3, with a bounded DNS worker
in parallel to the mechanical/layout repair worker and G018 implementation.
Keep exact non-overlapping file/branch ownership and attributable DNS/web commits;
no new Goal, fresh budget or reopening the old G016 implementation loop. Coordinate
review/integration and qualify the combined repaired source; avoid separate full
local gate campaigns for each fix. G017 final acceptance remains held until the
DNS correction and all required G017 qualification pass.

Expedite the already-running concise DNS proposal and supervisor review of exact
paths, dependency pin, system-resolver policy and deterministic failure cases.
Then extend the current G017 repair allow-list and release that bounded production
correction without another routine owner prompt. Until those details are reviewed,
the current50-pattern scope remains; do not silently edit production DNS/module
files. Existing security/contract/risky-migration/native gates remain, and material
resolver-policy changes escalate. G018 is not assigned predecessor repairs.
Autopilot continues through G019/G058 and stops before G020 planning.

## DNS companion scope06 execution release — 2026-10-17 UTC

Codex design/scope release (not included)
accepts the exact authored r0+r1 proposal and six new exact patterns (50→56) for
the DNS helper/tests and core module manifests. Restore the approved immediate
edge/three-hop guarantee with typed CNAME queries, bounded resolver behavior and
reviewed x/net pin; preserve TXT/authority/public contracts. Shared G017 attempt2,
parallel bounded writer, no new budget/task. Required qualification still gates
acceptance; G018 continues. Earlier proposal-only production hold is now released.

## Prospective platform-root amendment — G062, 2026-10-20

R4's requirement to preserve apex/`www` marketing redirects is superseded for the
configured platform signup host by [G062](G062-root-host-tenant-signup-spec.md).
The root offers Sign Up only and `/signup` provisions a new Organization. Preserve
the tenant-specific branded shell, strict unknown-host refusal and reviewed
protocol routes. G017's accepted history and other requirements remain intact.
