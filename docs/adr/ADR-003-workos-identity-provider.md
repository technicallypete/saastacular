# ADR-003 — WorkOS authentication behind the core IdentityProvider boundary

> **Editorial note (public export).** This historical decision record is retained for its reusable architecture and security reasoning. Links to private planning, review and evidence records that are not part of this repository are marked "(not included)"; references to "Architecture §N" or ARCHITECTURE.md point to the source project's architecture document, which is not included. Goal IDs, dates, approvals and superseded history are kept as written; product-specific passages are redacted where marked "[Editorial redaction: …]". Current behaviour is documented in the foundation guides under [`docs/foundation/`](../foundation/).

Date: 2026-10-06 (2026-10-07 UTC)
Status: ACCEPTED — explicit product-owner decision
Decision owner: Product owner
Recorded by: Codex
Implementation scope: G004; Claude Plan approval and fresh execution binding required

## Context

G004 introduces global identity and application sessions, not a custom identity server. After comparing Keycloak, WorkOS and other managed offerings, the owner directs: "Update to work os and continue claude to planning". This selects the managed provider and releases revised planning; it does not approve G004 implementation or provision an account.

## Decision

1. Select **WorkOS AuthKit** for hosted user authentication, MFA and provider SSO. Keep provider-specific SDK/API types inside a Go core-owned `IdentityProvider` adapter. Next.js remains a thin BFF; do not use a frontend auth installer to transfer identity/session authority out of core.
2. Preserve immutable canonical application user UUIDs, exact trusted provider-principal mapping and PostgreSQL-authoritative opaque application sessions, expiry, revocation and local account disable. WorkOS user/session identifiers are integration references, never canonical application IDs. WorkOS Organizations, roles, permissions and entitlements confer no Beginify authority. G005+ owns the application tenancy model; the G021 barrier stays closed.
3. Use WorkOS's documented hosted authorization-code exchange, with mandatory browser-bound one-use state and PKCE S256. Verify its actual signed token, configured issuer, client binding, subject and time contract. Do not assume AuthKit returns generic OIDC ID tokens, discovery, nonce, `aud` or `acr`; the Plan maps equivalent protections to documented fields and escalates any missing required guarantee. Preserve signature validation and token/user consistency where tokens supply identity or assurance.
4. Preserve a stable provider identity namespace across the approved callbacks. A future custom auth domain, WorkOS environment/client change or provider migration must not silently change existing mappings or trigger email auto-linking. Avoid provider bearer/refresh-token retention; a necessary exception requires explicit security review. Provider logout metadata/return handling and observable guarantees are pinned in the Plan; local committed revocation is independent of upstream completion.
5. Plan containerized deterministic tests using the **official WorkOS emulator**, with pinned provenance and feature limits, alongside a bounded **real WorkOS staging** qualification. The emulator is not real authentication or sufficient MFA/SSO acceptance evidence. Never enable it against production traffic. Retain genuine hosted sign-in/callback, MFA, provider SSO reuse and reauthentication evidence, plus protected real-service integration checks on the final implementation. Define safe manual qualification where WorkOS advises against automated hosted-UI login; do not bypass bot protection. Reuse accepted CI entry points and controls rather than introduce a new orchestration framework.
6. Select hosted WorkOS rather than operating a production Keycloak service. Ordinary auth is the initial integration; paid enterprise SSO connections, Directory Sync, custom auth domains, vendor Organization administration and enterprise policy UX are not G004 scope. No purchase, production account/data upload, API-key creation, callback registration, CI-secret/project-setting change, evidence push or hosted pipeline is authorized by this decision. Name necessary staging/account/callback/secret actions and their authorization prerequisites in the Plan.

## Consequences and remaining implementation prerequisites

- WorkOS hosts authentication attributes and factors; Beginify retains its own application identity and authorization records. Production terms, data-processing/residency, subprocessors and support requirements remain rollout prerequisites, not assertions already established by selecting the vendor. No production data is permitted in test environments.
- Claude revises G004's Plan at xhigh effort, preserves r1/review history, resolves the five Codex review findings, pins compatible libraries/test artifacts and describes account access, callback reachability and secret isolation. G048 acceptance/integration remains the implementation prerequisite; planning can proceed in its isolated clone.
- Real-service secrets must never be available to untrusted MR code or included in source, logs, screenshots or artifacts. The Plan defines which tests are offline and which require approved protected credentials, with fail-closed unmet evidence rather than successful skips. Account availability and these concrete permissions remain unresolved.
- Required assurances cannot be invented from a requested challenge or an unverified claim. If WorkOS cannot expose sufficient trusted evidence, return the exact capability gap to Codex before implementation; no weaker session/authentication fallback is approved.
- B7 remains a maximum of seven shared meaningful cycles with checkpoint after four. Vendor selection does not validate r1's five-cycle forecast or release execution.

## Owner-approved assurance clarification — 2026-10-07

The owner approves Codex's recommendation following the G004 Plan r2 review (not included): retain WorkOS, enable provider MFA for the qualification environment, demonstrate real MFA success/cancellation and SSO reuse, and leave per-session MFA assurance **unknown** unless trusted session-specific proof exists. This accepts Plan r2's M2 direction and rejects M1's `mfa_by_provider_policy` inference. It supersedes the original G004 R7/AC7 requirement to obtain a verified-MFA session during capability qualification; the preserved r2 Spec records that historical requirement.

- Ordinary sign-in may issue an otherwise valid application session with unknown MFA assurance. Lack of an MFA claim is not itself a reason to deny ordinary sign-in; all signed identity/client/time, PKCE, callback, session and revocation checks remain mandatory.
- Environment-required MFA, an authentication method or a periodic qualification result cannot mark a particular session as verified MFA or grant authorization. Do not persist or expose a policy-derived MFA assurance label. Provider authentication method may be used only within its separately reviewed protocol/validation purpose.
- Any request or test profile that specifically requires verified MFA must fail closed without trusted proof particular to that session. Accepting a future proof source/mapping requires explicit security/Plan review; do not invent a claim or silently trust a browser parameter. G004 authorizes no product operation requiring MFA and no new factor-handling/step-up UI.
- Required real provider MFA challenge success/cancellation, provider SSO reuse and signed `auth_time`/forced reauthentication evidence remain G004 acceptance obligations. Emulator/fixtures cannot substitute for those real flows. Cancellation must issue no application session from that transaction.
- Provider policy drift must not produce a false per-session assurance claim. Qualification records the tested policy/configuration separately from application session assurance; it is not continuous attestation.

This changes assurance semantics only. UUID/core/session authority, the G021 barrier, B7, G048 acceptance/integration, separate Plan approval and external-action gates remain. It releases revised planning, not implementation, account/secret changes or hosted execution.

## Owner-approved optional MFA and onboarding policy — 2026-10-07

Owner direction: MFA should remain optional so that it adds no onboarding friction [Editorial redaction: the original quotation named the source product's user groups and is paraphrased]. This supersedes the conversational recommendation to require MFA for every human user as a production policy. It does not supersede the preceding session-assurance clarification or G004's required-MFA qualification setup.

- MFA enrollment is voluntary for every ordinary user type [Editorial redaction: source-product user-type names removed]. A user who has not opted in must be able to complete otherwise valid ordinary authentication and onboarding without installing an authenticator or enrolling a factor. No default enrollment mandate for any user type or privileged role is introduced.
- For users who opt into MFA, retain provider-owned enrollment and challenge handling. Optional enrollment is not permission to skip a required challenge on an enrolled account, issue a session after cancellation or weaken identity/session validation. Verified-MFA requirements remain fail-closed without reviewed session-specific proof; changing or introducing a sensitive-operation exception requires separate review.
- Configure eventual production authentication for optional enrollment; do not copy the qualification environment's Required setting as a global onboarding policy. G004's staging remains Required to prove MFA success/cancellation and real SSO/freshness behavior. Ordinary sessions retain the accepted unknown-assurance semantics. No active Goal/Spec/Plan/snapshot bytes, acceptance criteria or budgets are changed by this clarification.
- Evaluate provider-supported passkeys and a usable account-recovery path for future low-friction authentication. Their precise behavior and any interaction with MFA must be qualified before adoption. This does not authorize a passkey dependency, custom factor UI, SMS fallback, account/factor reset, Organization policy schema or provider configuration changes.
- Carry this policy into future invitation/onboarding and rollout Specs. Objective check: an otherwise eligible user without an enrolled factor can authenticate, accept an invitation and finish ordinary onboarding without an MFA enrollment screen; an opted-in user's canceled challenge cannot issue a session. Those future checks do not introduce invitations/onboarding implementation into G004 or waive G021.

## Alternatives and history

- Keycloak: self-hostable and suitable for disposable provider tests, but adds production operation responsibility. Claude's r1 recommendation is superseded by the owner's WorkOS choice; no Keycloak installation is authorized for G004.
- Clerk/Auth0/Okta: discussed alternatives, not additional adapters or adopted vendors.
- Original G004 Goal/Spec r1 and Plan r1 are retained as historical provider-neutral/OIDC assumptions. The revised Goal/Spec and this ADR govern renewed planning; Codex must approve the revised Plan separately.

## References

- [AuthKit code exchange](https://workos.com/docs/reference/authkit/authentication)
- [Authorization URL and PKCE](https://workos.com/docs/reference/authkit/authentication/get-authorization-url)
- [Signed session tokens](https://workos.com/docs/reference/authkit/session-tokens)
- [MFA](https://workos.com/docs/authkit/mfa) and [reauthentication](https://workos.com/docs/authkit/reauthentication)
- [Provider logout](https://workos.com/docs/reference/authkit/logout)
- [Testing layers and emulator limitations](https://workos.com/docs/authkit/testing)
- [Pricing and free staging](https://workos.com/pricing)

These references inform the decision; pinned-version behavior and final application evidence remain G004 obligations.


## Owner-approved G016 domain/authentication refinement — 2026-10-16

[ADR-007](ADR-007-verified-domain-routing-and-authentication.md) records the owner's
explicit approval of tenant-host membership, shared WorkOS callback relay and
exact-host historical MCP retention, with unchanged canonical identity/authority.
It preserves this ADR's historical evidence. Design approval alone grants no live
provider/settings/deployment action or G016 execution release.
