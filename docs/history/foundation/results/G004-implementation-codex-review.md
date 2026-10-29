> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G004 implementation review — ACCEPT

Date: 2026-10-08 (2026-10-09 UTC)
Reviewer: Codex
Outcome: **ACCEPT — G004 implementation, AC1–AC12.**

This closes G004's implementation review, superseding the outstanding
implementation-acceptance status in earlier checkpoint/manual qualification
records. Those records, failures and approved corrections remain historical.
Acceptance does not authorize integration, production rollout, G005 execution
or product features before the G021 foundation barrier.

## Exact reviewed delivery and inputs

- Execution checkout: `.claude/worktrees/g004-exec`, branch
  `g004-global-identity-sessions`.
- BASE: `201f5d1114bd43d978dca6afb26bf410e5f79da7`.
- Final delivery: `7dc4c5cc28fb05d79daa52be3903599c09be34b8`;
  tree `6f24080bab1ce65e1312578d91f54886c3997a04`.
- Hosted-qualified commit: `4a46da96136b54fd816cfd5dca2cfebd92ae1662`;
  tree `6669f6033d0ba52ac5d3ea9fc84a6bdde1e362ae`.
- Frozen manual candidate: `0f5a5dca6dff43a470de7b57c247cdeb80277868`.
- Final Result SHA-256:
  `45210c3a438af0b8992c254095037da93a8abf04a8c0c333ee4d180cc40da4a9`.
- Plan r5 SHA-256:
  `fcbc863d372eb6e93a811311c0e85bd47e1df898086265a9eb1d3d15d2963aa6`.
- Approved 42-input snapshot SHA-256:
  `31360f9c87ac60bfe68a3767c9d1777e27e75906cdcc9bb6b5988c2fbb499a07`.
- C7 addendum SHA-256:
  `841faa15e835d5aec2195d3978e4651f64ed4bbf2f87d1510f03eb51cd0e9f21`.
- Final 91-entry scope SHA-256:
  `0001c719ee49c72711c440111be02c0a6af30003adb342a3f76d4111239c846e`.

Codex independently verified the 42 unique input paths and every input hash,
the Plan/addendum/scope hashes, and all 88 changed paths against the anchored
scope expressions. No change falls outside that scope. The manual candidate,
hosted commit and final delivery differ only in the Result. Execution and main
checkouts were clean when final review began. Review changed management records
only; it did not modify Claude's checkout or application files.

## Review and acceptance evidence

Read the canonical governance, architecture and accepted provider decision,
G004 Goal/Spec/Plan, approved correction records, C7 addendum and final Result.
Inspected the actual implementation diff, including the provider token/JWKS
boundary, core identity/session store and lifecycle, HTTP contracts, migration,
web cookies/origin/CSRF and logout-return handling, tests and CI secret forwarding.
Passing tests alone were not treated as acceptance.

| Criterion | Accepted behavior and evidence |
|---|---|
| AC1 | Core-owned WorkOS adapter, exact issuer/client binding, asymmetric signature and time validation, trusted bounded JWKS retrieval and fail-closed configuration. Real staging contract passed; deterministic trusted-key rotation and attacker-key controls inspected. No live provider key rotation is claimed. |
| AC2 | Canonical UUID users, exact provider-principal uniqueness and stable mapping without email linking or vendor Organization authority. PostgreSQL concurrent provisioning/namespace tests and both-origin manual mapping evidence. |
| AC3 | One-use browser-bound transactions, S256 PKCE and validated code exchange; malformed, replayed, canceled and conflicting callbacks cannot create or switch unauthorized sessions. Real wrong-verifier qualification and deterministic callback/PKCE evidence. |
| AC4 | PostgreSQL-authoritative opaque sessions with finite idle/absolute expiry; no provider bearer retention or Redis authority. Fixation, replicas, restart, revocation/renewal races and store-outage checks passed. |
| AC5 | Host-only Secure/HttpOnly/Lax cookies, exact origin and CSRF checks, bounded form inputs, no-store responses and production HTTP-relaxation refusal. Actual HTTPS browser and negative controls passed. |
| AC6 | Local revocation commits before upstream action; logout-all and disable preserve the reviewed concurrency semantics. C7 one-use logout notification is distinct from identity authority. Real same-origin provider return, cookie removal and prompted subsequent authentication passed on both origins. |
| AC7 | Real MFA success/cancellation, SSO reuse and signed-time forced freshness demonstrated. Session assurance remains unknown; explicit verified-MFA requirements deny. Qualification policy is not production's optional-enrollment policy. |
| AC8 | Thin account/authentication BFF, core-owned principal authority, closed errors and no tenant/MCP exposure. Identity browser, keyboard/accessibility/network and inherited appearance/scope/fidelity checks passed. |
| AC9 | Only the five approved global tables and their indexes; UUID user identity, exact uniqueness/FKs, NOLOGIN-owner migrations and bounded runtime grants. Baseline migration unchanged; actual PostgreSQL catalog/CRUD/DDL negatives and disposable migration lifecycle passed. |
| AC10 | Closed correlated diagnostics and privacy scanning of actual logs, delivered pages and evidence, with planted leak controls. Test helpers/emulator endpoints remain isolated or refused in production. |
| AC11 | Local final wrapper and hosted fast/full/real-provider staging qualification pass on the same application/validation bytes. Existing suites, variants, provenance and fail-closed secret/context controls retained. |
| AC12 | Exact hashes, scope, final evidence mapping, historical failures, cycle-4 checkpoint, final Result and resource-cleanup evidence accepted. Seven of seven B7 cycles used; no budget reset. Workstation-only cleanup remains the explicit owner handoff below. |

Local final S8d run `8a9014` passed **77/77** required checks, with wrapper
harness exit 0 and passing postcheck. Evidence remains outside the repository at
`<home>/.local/state/platform-evidence/g004-exec-64f1c470/`.
Identity integration: 57 tests; identity browser: 56; inherited browser root and
disabled: 141 each; CI controls: 224. Codex inspected the final manual observer
PASS segment while retaining its earlier FAIL segment as history.

Codex independently checked GitLab pipeline metadata and downloaded the final
job artifacts into memory to inspect their `results.tsv` records:

| Pipeline/job | Outcome | Duration |
|---|---|---|
| 2928428140 (not included), protected `identity-qualification`, source `web`, qualified commit above | success | 3,752 seconds pipeline duration |
| fast 17045592437 | 33 PASS; 0 FAIL/UNMET | 380 seconds |
| gate 17045592438 | 77 PASS; 0 FAIL/UNMET | 3,372 seconds |
| identity-staging 17045592439 | 5 PASS; 0 FAIL/UNMET | 142 seconds |

Hosted adapter summaries report final status 0 and registry removal. No new
pipeline or broad test run was launched during review.

## Dispositions and remaining operator tasks

1. **Integration coverage:** accept the completed hosted fast + full jobs in
   place of a duplicate standalone integration-profile run. Strictly, full's
   check-ID set does not include `B-UNIT-CHECKS`; fast passes it, and full passes
   `B-CHECKS-ROOT` using the documented check commands. Codex compared the actual
   passing artifacts: their union contains every required integration ID.
   This precise disposition supersedes the Result's literal `full ⊇ integration`
   shorthand; no required test or profile-selection rule is relaxed.
2. **Earlier operator evidence:** accept the explicit C7-approved reuse for
   unchanged login/callback/cancellation/disable behavior and the linked actual
   SSO reuse. Original MQ-3 is ordinary sign-in only; original MQ-5/MQ-6 failures
   remain recorded. C7's new freshness and both-origin return behavior has fresh
   focused manual evidence, followed by final local/hosted qualification.
3. **Inherited G048 intermittency:** retain S8c's disabled dark fidelity failure
   and unchanged S8d pass. No test was weakened or unrelated application change
   made. Track stabilization as a follow-up; a passing unchanged final run and
   hosted regression are present, so no further qualification is required here.
4. **Cleanup:** owned Linux test projects, scratch databases/checkouts, generated
   secret copies and per-run WorkOS users were removed. Codex independently found
   no G004/G001-named containers, volumes or networks. Owner credential files,
   approved persistent test users and protected private evidence were deliberately
   retained. No claim is made that workstation cleanup has already happened:
   close Firefox; remove `C:\Temp\g004-mq-ff` and existing Chrome attempt profiles
   `C:\Temp\g004-mq-profile`, `-profile2`, `-profile3`; remove the exact-leaf ESET
   exception identified in the Result. The run CA was trusted only in the test
   Firefox profile. This bounded operator cleanup is a recorded handoff, not an
   application defect requiring another implementation cycle.

   **Owner disposition after acceptance, 2026-10-08 (2026-10-09 UTC):**
   "Ignore the ws cleanup on my end." The workstation-only browser-profile and
   ESET exception cleanup follow-up above is waived at the owner's direction,
   not represented as performed. It is no longer an outstanding G004 task or
   integration prerequisite. This does not change the completed local/hosted
   resource cleanup, credential handling or future test-isolation requirements.

## Next gates

G004 is accepted, **not yet integrated into staging**. Integration follows the
owner-authorized GitLab MR route with default source-branch deletion; retain
`main`, `staging` and `identity-qualification`. This review performs no commit,
push, MR creation or merge.

G049 is ready for its post-integration planning/binding steps; no implementation
is released here. G005's bounded provisional Plan amendment must reconcile to
the actual accepted/integrated G004 baseline before execution approval. Its
future schema/foreign keys require checking the existing disposable staging
test setup; this is a downstream reconciliation obligation, not a G004 defect
or authorization for G005 schema work. Preserve G006+ ownership and the closed
G021 product-feature barrier.
