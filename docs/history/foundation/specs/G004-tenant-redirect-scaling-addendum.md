> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G004 / G016 — Tenant authentication redirect scaling clarification

Date: 2026-10-07. Owner requested this specification record after configuring WorkOS staging.

## Authority and scope

This addendum clarifies the boundary between [G004 identity](G004-global-identity-sessions-spec.md), [G005 tenancy](G005-organization-group-memberships-spec.md) and G016 domain routing (not included). It records requirements to inherit when G016 is specified; it is not a complete G016 Work Spec or implementation authorization. Architecture §§53.2 and 60 and accepted ADR-003 govern.

G004 Spec r3 and Plan r4 remain byte-identical. Their AC1–AC12, two explicit HTTPS test origins, exact callback/sign-out allowlists, host-only cookies, browser-bound state, PKCE and execution gates remain unchanged. Do not implement wildcard origins, a shared callback, cross-host session transfer or tenant provisioning in G004. Include this clarification in the pending mechanical planning reconciliation and fresh input binding without reopening unchanged design.

## Requirements inherited by tenant-domain integration

1. **Application versus tenant.** Use one WorkOS web application per platform environment for the shared web surface. Creating an application Organization must not require creating another WorkOS application or client ID. Separate application surfaces or intentionally isolated identity environments require their own reviewed rationale.
2. **Owned tenancy and identity.** G005 owns UUID-identified Organizations, memberships and shared RoleBinding authority. Provider Organization IDs, if later needed for enterprise SSO, are external integration references. A hostname, provider organization claim, slug or successful login cannot grant membership or tenant permission. No vendor Organization provisioning or new schema is authorized by this addendum.
3. **Bounded redirect configuration.** Platform-managed tenant growth must not require one provider redirect/sign-out entry per Organization. The eventual G016 Spec must cover platform subdomains and verified custom domains and state any remaining provider configuration requirements. Preserve configuration-based platform naming; example domains are not source constants.
4. **Platform subdomains.** A provider-supported wildcard callback/sign-out pattern is a candidate for verified platform-controlled subdomains, with a concrete default URI. A provider wildcard is not an application origin allowlist: application code must resolve each actual normalized hostname to a verified Organization and validate the transaction's exact origin and destination. Review wildcard limits, reserved hosts, domain ownership and production TLS before adoption.
5. **Custom domains.** A shared platform callback plus a secure return to a verified custom domain is the recommended candidate for avoiding per-customer provider entries. Its implementation and security contract remain unresolved until G016 Plan review; this record does not select a cross-host session mechanism or authorize broader cookies. Any proposal must preserve host-only Secure/HttpOnly cookies, browser-bound login transactions, PKCE, core session authority and protection against replay, login CSRF and open redirects. Provider tokens and reusable application session credentials must not appear in redirect URLs.

## Verification to require in the G016 Spec

The future Goal must turn these requirements into numbered acceptance criteria with:

- Two platform Organizations and a newly added third Organization authenticating without adding provider applications/client IDs or individual redirect/sign-out entries. Record configuration before/after and concrete URI patterns, excluding secrets.
- A verified custom-domain flow under the selected mechanism, including its final host-only session cookie and same global user UUID across authorized hosts. Record whether custom-domain addition requires any provider configuration; do not claim configuration-independent behavior without testing it.
- Denials for unknown/unverified/reserved hosts, forged forwarding headers, foreign Organization access, altered return destinations, replayed handoffs and a mismatched initiating browser. Prove no unauthorized session or authority is created, rather than checking response codes alone.
- Sign-out behavior on platform and custom hosts, preserving truthful local/provider revocation outcomes and preventing arbitrary external return destinations.

Use disposable HTTPS/DNS fixtures and accepted validation controls. These tests belong to G016's later reviewed scope and do not add G004 acceptance work. Enterprise connection administration, paid features, production deployment and all product-specific features remain outside this clarification; preserve the G021 barrier.

## Decisions still required before that implementation

G016 must select and review wildcard versus shared-callback use, browser binding across origins, any single-use handoff, domain verification/removal lifecycle and sign-out returns. Changing G004's established authentication/session contract requires explicit security review and revised execution inputs. No shared callback or wildcard has been approved or configured by this documentation change.

## Provider references

Checked 2026-10-07: WorkOS [Applications](https://workos.com/docs/authkit/applications) distinguishes application surfaces from tenants; [authorization redirect rules](https://workos.com/docs/reference/authkit/authentication/get-authorization-url) and [sign-out URI rules](https://workos.com/docs/authkit/sessions) document constrained wildcards and require a concrete default. Reverify the deployed provider contract when planning G016; these documents do not establish application authorization or demonstrate a cross-host handoff.
