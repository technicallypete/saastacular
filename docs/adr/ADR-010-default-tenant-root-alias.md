# ADR-010 — Optional existing-tenant alias at the platform root

> **Editorial note (public export).** This historical decision record is retained for its reusable architecture and security reasoning. Links to private planning, review and evidence records that are not part of this repository are marked "(not included)"; references to "Architecture §N" or ARCHITECTURE.md point to the source project's architecture document, which is not included. Goal IDs, dates, approvals and superseded history are kept as written; product-specific passages are redacted where marked "[Editorial redaction: …]". Current behaviour is documented in the foundation guides under [`docs/foundation/`](../foundation/).

Date: 2026-10-26 UTC
Status: ACCEPTED owner architecture direction; delivery not yet released
Decision owner: Product owner; recorded by Codex
Delivery: G071 (not included) /
Spec (not included)

## Decision

An optional server-side `DEFAULT_TENANT_SLUG` selects an already-existing
Organization at startup. Resolve its canonical UUID and serve its tenant entry
and ordinary tenant experience at the exact configured root origin. No browser
redirect to a tenant subdomain, tenant creation/seeding or new database mode.
Unknown slug logs a warning and leaves normal multi-tenant behavior active;
unset/empty also preserves normal behavior. A lookup outage is not absence.

This is an entry/host alias, not a guarantee that only one Organization exists.
The trusted deployment operator can select an existing Organization regardless
of its ownership, but this grants users no authority. Preserve tenant isolation,
membership/enrollment, host-bound sessions and separate root SignupIntent.
Under the later route refinement, the active valid alias returns 404 for root
`/signup` and its subpaths; `/join` remains policy-controlled. Backend Organization
provisioning is unchanged. Inactive/unset/unknown restores normal root signup. Other tenant/custom/auth/www/unknown
hosts retain their current handling. Pin the UUID until configuration restart;
configuration edits are not a hot-reload mechanism.

In multi-tenant operation only, server-side
`PLATFORM_ROOT_PAGE=signup_redirect|landing` selects exact-root GET/HEAD entry,
defaulting to a temporary no-store redirect to `/signup`. `landing` retains the
existing foundation page; marketing can be a separately hosted site. A valid
default-tenant alias wins and ignores this selector. Unknown alias falls back to
the selected multi-tenant presentation. No Organization-count detection, login
side effect, arbitrary redirect target or marketing deployment is included.

## Relationship to prior decisions

This is a bounded opt-in exception to G062's default root-as-platform-entry
behavior and ADR-007's absence of default-tenant routing. Their normal behavior
and security/privacy contracts otherwise remain. G059 stays deferred: its broader
single-tenant/default-Group proposal is not revived or reused. The owner requested
a new Spec to prevent that confusion, initially after G070; the later scheduling
correction below moves it before G070.
No global provisioning restriction, tenant adoption machinery, Group auto-provisioning,
apex-domain product or arbitrary-host fallback is authorized.

## Delivery constraints

Core remains authoritative for tenant selection. Plan the minimal private
slug-to-UUID read without broad runtime grants or reinterpreting platform labels.
Keep the normal signup implementation intact; the existing web gate excludes its
root paths only while a valid alias is active. Preserve `/join` and tenant auth.
Source review and focused config/routing/identity/signup checks cover the option;
no mandatory full-suite campaign or real provider deployment is added. G070/G069
and G057 frozen work stay protected. Code requires Plan ACCEPT and normal release;
publication is specification/scheduling only, not final Goal acceptance.


## Later owner scheduling correction — 2026-10-26 UTC

“We should queue this for before 070 btw” supersedes the initial ordering above.
Current priority is G069 → G071 → G070; G071 does not depend on G070 hosting or
final acceptance. Preserve in-flight G070 source checkpoints/frozen proofs, then
hold its new phases/integration behind G071 and reconcile actual config effects.
No source/Plan/budget reset or live-action release. The route inventory is root
`/signup` for Organization creation, `/join` for membership admission, and
`/groups/new` for authorized Group onboarding; existing meanings/intents remain.


## Later owner route simplification — 2026-10-26 UTC

A valid active default tenant makes root `/signup` and its web subpaths return
404 through the host/path gate. `/join` is expressly retained for ordinary
Organization enrollment; Group onboarding follows existing admission/authority.
No automatic redirects between these distinct flows or backend provisioning
switch. Unset/unknown restores normal signup. This supersedes only the earlier
active-alias direct-signup requirement; other routes/security contracts remain.


## Later owner multi-tenant root-page option — 2026-10-26 UTC

The owner prefers signup_redirect as the default multi-tenant app-root experience,
with the existing landing available via PLATFORM_ROOT_PAGE=landing. The platform
can host application entry while a marketing site is served independently elsewhere.
This is Spec r3's presentation option, ignored for a valid active default tenant;
it does not change admission, signup404 or backend provisioning rules. G071's
priority before G070 and its focused validation/release gates stay unchanged.
