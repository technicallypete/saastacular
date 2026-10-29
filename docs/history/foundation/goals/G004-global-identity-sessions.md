> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G004 — Global identity and sessions

Epic: EPIC-002 — Identity and tenant foundation
Status: r3 — owner-approved MFA assurance clarification; revised Claude planning authorized; implementation not authorized
Implementation owner: Claude Code
Supervisor and acceptance reviewer: Codex
Budget class: B7 — maximum seven shared meaningful cycles; checkpoint after cycle 4

## Objective

Integrate a mature external identity provider through a Beginify-owned `IdentityProvider` boundary, map authenticated principals to stable global user UUIDs, and establish server-authoritative application sessions with expiry, logout, revocation and account-disable behavior. Provide the minimum browser sign-in/current-account/sign-out surface and prove provider-owned MFA and SSO integration capability.

This is reusable authentication foundation, not a custom identity provider or tenant/product implementation. Identity is global; authorization remains scoped and belongs to its later Goals. The product-feature barrier remains closed through G021.

## Context and Rationale

Original specification drafting observed `19f945af7b7bfd1b27cb395fd56b0c2cec3b163b`; the r1 Goal (not included) preserves that historical state. At this r2 planning handoff, main HEAD is `cc25a16ae4a30cf2c1d0e68252609eace128938e`, G003/G047 are accepted and integrated, and G048 is implementing in its separately bound checkout. G048 acceptance/integration and fresh G004 input binding still gate implementation; do not modify its active inputs.

This r3 handoff observes main HEAD `126165e0682e77baa5e8810a7df4d198703cc5b6` with supervisor-owned planning/review changes uncommitted; it is not an execution binding. The preserved r2 Goal (not included) retains the previous handoff. On 2026-10-07 the owner approves the MFA clarification in ADR-003: demonstrate WorkOS MFA capability, retain unknown per-session MFA assurance without reviewed trusted proof, and fail closed when verified MFA is specifically required. No policy-derived assurance label is approved.

The repository still configures `IDENTITY_PROVIDER=oidc` without an implemented provider, user store or browser session. The owner now selects hosted **WorkOS AuthKit**, recorded in accepted [ADR-003](../../../adr/ADR-003-workos-identity-provider.md). Existing configuration is a historical protocol placeholder, not proof of compatibility with AuthKit. Claude's revised Plan defines the actual WorkOS adapter/configuration contract and reconciles the accepted G003 operations/CI and post-G048 browser foundation.

G004 owns the first global identity schema. G005 owns Organization/Group membership and RoleBinding; G007 owns trusted tenant context/policy; G008 owns tenant RLS; G014/G015 own the authenticated MCP control plane. Do not pull those implementations forward to demonstrate authentication.

## In Scope

- One WorkOS AuthKit integration, replaceable behind a core-owned interface, using its documented hosted code exchange and reviewed signed identity/assurance contract.
- Minimal global user and provider-principal mapping, concurrent provisioning integrity, and server-side session lifecycle/revocation state.
- Provider-redirect sign-in, validated callback, current-account display, current-session logout and self-service logout of all local sessions.
- A trusted core account-disable/re-enable seam, exercised through internal tests; no new public administration interface.
- Provider-owned MFA challenge success/cancellation, SSO reuse and signed authentication-time/freshness evidence; unknown session MFA assurance without reviewed trusted proof and denial of requests requiring verified MFA. No application-owned password/factor implementation or product step-up operation.
- Focused identity migrations, runtime grants, configuration, authenticated-boundary tests, Docker-first commands and inherited GitLab CI integration.
- Exact consumer-customization/provenance reconciliation, foundation identity documentation, and a separate G004 Result.

## Out of Scope

- Organization, Group, BusinessEntity, membership, RoleBinding, invitation, onboarding, tenant context/switching or tenant-policy schema and UX.
- A custom authentication provider, passwords, reset/magic-link email services, local TOTP/WebAuthn enrollment, social-provider catalog, account linking/merging or production SAML implementation.
- Organization-specific SSO configuration, verified custom domains, cross-domain session transfer or wildcard-domain cookies.
- Public MCP transport/tools, machine/API credentials, public AI execution, intelligence authentication replacement, product features or billing.
- A user directory, public administration, account erasure/retention program, durable audit ledger, broad security-tool expansion, production deployment or unrelated visual/refactor work.
- Installing a major identity vendor/service without an accepted provider ADR and reviewed Plan.

## Dependencies

- **G003/G047 implementation ACCEPT and integration** are satisfied. **G048 implementation ACCEPT and integration** remain required before G004 implementation. Owner-authorized revised planning may proceed in an isolated clone while G048 executes; no active G048 input/resource change is allowed.
- Reassess the accepted post-G003 tree before Claude planning: actual migrations, logger/redaction behavior, browser matrix, boundary/provenance guards, CI commands and accepted deviations. Obtain a fresh G004 baseline/input binding at Plan approval.
- Provider selection is resolved by accepted **ADR-003: managed WorkOS AuthKit**. Its exact client/issuer/subject, token/assurance/logout, test access, HTTPS and CI-secret contracts remain Plan prerequisites. No Keycloak service or additional identity vendor is authorized.
- An available approved provider test environment and test HTTPS/callback configuration, with synthetic accounts and a repeatable containerized test path. Missing access/capability is a blocker, not permission to replace real integration evidence with a fake provider.
- Read canonical governance, architecture/security/decisions, roadmap/progress, BUDGET_POLICY, predecessor Specs/Results/reviews and the [G004 Work Spec](../specs/G004-global-identity-sessions-spec.md).

## Completion Definition

1. Work Spec AC1–AC12 pass; no mandatory provider, security, migration, browser or hosted check is skipped, unmet or deferred.
2. The submitted Result identifies the exact implementation/tree, safe evidence, all failures/deviations and shared B7 cycle accounting, with the cycle-4 checkpoint.
3. Codex reviews the diff, provider integration, authenticated boundary, migrations, CI and inherited guarantees, returns ACCEPT, and updates PROGRESS.
4. Authentication completion grants no tenant membership/role and does not mount MCP. G005 remains a separate Goal with its own Spec/Plan and the G021 barrier remains closed.

## Execution Policy

This r3 handoff authorizes revised Claude planning at **xhigh** in the existing isolated G004 planning clone, with public-documentation/read-only prerequisite research and goal-specific Plan/history writes only. Incorporate the owner-approved MFA semantics and four bounded corrections in the r2 review (not included); preserve Plan r2 and reissue r3. Unavailable WorkOS account/configuration or unresolved post-G048 facts must be named as execution blockers; they do not prevent drafting the concrete revised Plan. It does not authorize implementation, dependency installation, provider provisioning, new hosted actions, application builds or migrations. Codex must approve the revised Plan and fresh inputs before execution.

After explicit G004 Plan ACCEPT, Claude implements/tests only enumerated paths, schemas, routes and dependencies. New authentication endpoints and identity migrations are anticipated scope; the Plan must define and obtain review of their exact contracts and migration risk before execution. New public/MCP contracts outside that inventory stop for escalation.

Reuse the accepted Docker/CI harness, approved-input wrapper, isolation and cleanup. Preserve existing `.env`, data and unrelated provider/Docker resources. Rollback/outage/replay trials use verified owned disposable projects. A live provider account change, callback registration, commit/push or hosted test action needs its own explicit authorization; the G047/G003 execution authorization does not carry forward.

## Iteration Budget

**B7: at most seven meaningful implementation/debug cycles shared across all agents, checkpoint after cycle 4 and before cycle 5.** Follow BUDGET_POLICY. No cycles are consumed by this specification; no predecessor cycles transfer and no automatic correction allowance is granted.

B7 is the smallest class suited to identity migrations, a real-provider/browser/core security boundary and revocation/CI evidence. Provider selection and environment availability must be settled before its execution forecast. If the Plan cannot satisfy the bounded criteria within B7, propose a split or revised budget before implementation; do not omit negative cases or provider evidence.

## Context / Scope Budget

One provider, one global identity boundary, one minimal account surface and one session system. Keep provider-specific details in its adapter and application session authority in core. Reference accepted execution controls rather than building another framework. Future tenant membership, enterprise connection management and MCP transport remain explicit follow-up owners.

## Suggested Subagents / Delegation

Prefer one implementation owner with bounded read-only reviewers for provider-protocol/session security and migration/CI evidence. Claude may delegate non-overlapping provider-adapter or test work when the Plan names exact file ownership, integration responsibility and shared-cycle accounting. Provider/vendor and architecture decisions remain with the owner; reviewer opinions do not replace Codex acceptance. Use the owner's xhigh preference for planning.

## Stop / Escalation Conditions

- For implementation: missing G048 acceptance/integration, actual WorkOS staging/test HTTPS access, exact reviewed inputs or authorization for external actions. Missing account access during planning is reported explicitly, never treated as satisfied.
- Architecture/security conflict, unreviewed provider/hosting change, new major dependency/vendor, irreversible identity mapping or destructive migration.
- Need to build credential/factor handling, infer membership from provider claims/hostnames, expose admin/MCP/product authority or implement future tenant schema.
- Need to auto-link by email, weaken callback/session/CSRF/revocation validation, ship a development authenticator, expose secrets/PII or substitute mock/local evidence for required real-provider/hosted evidence.
- Need to mark a session as verified MFA without reviewed session-specific proof, infer MFA assurance from provider configuration/authentication method, or grant a verified-MFA request while assurance is unknown.
- Conflicting developer resources, scope growth, thrashing, B7 exhaustion or inability to satisfy all criteria within available context.

## Revision Record

- 2026-10-05 — Initial G004 drafting while the owner-authorized G047 → G003 batch executes separately. Provider remains unselected; no Plan, implementation, provider action, commit or push is performed by this drafting task.
- 2026-10-06 (2026-10-07 UTC) — r2, owner selects WorkOS and requests continued Claude planning. Accepted ADR-003 supersedes provider-neutral/vendor-unselected assumptions and Claude r1's Keycloak recommendation. G003/G047 acceptance is reconciled; G048 implementation stays isolated and is still an execution prerequisite. Original r1 bytes are preserved at the linked historical path. AC1–AC12, B7/checkpoint four and the G021 barrier remain; no implementation/provider/account action is authorized.
- 2026-10-07 — r3, owner approves MFA capability qualification with unknown per-session assurance unless reviewed trusted proof exists; requests requiring verified MFA fail closed. Preserve r2 bytes, retain all genuine MFA/SSO/freshness evidence and the four bounded Plan corrections. B7/checkpoint four, dependencies, scope and implementation/external-action gates remain unchanged.
