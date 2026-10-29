> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G062 — Root-host tenant provisioning signup

Epic: EPIC-008 — Foundation completion, before G021.
Status: Provisional planning RELEASED by owner, 2026-10-20; implementation held.
Spec: [G062](../specs/G062-root-host-tenant-signup-spec.md).

## Objective

Give a reusable platform a minimal application-root landing page with **Sign Up**
and `/signup`: authenticate through WorkOS, create or verify the canonical global
identity, then create a new Organization with that verified creator as its initial
`org_owner`. Complete minimal tenant setup and enter its trusted tenant hostname.
Existing global users may create a new tenant; provider account creation is not
required on every signup. Existing-tenant joining remains G058.

For the owner's local example, `https://localhost:543` shows the landing page and
links to `https://localhost:543/signup`; after authentication and tenant creation,
the browser completes onboarding on the new tenant host. The seeded
`https://org-a.localhost:543` remains an independent existing tenant. Hosts, product
names, HTTPS origin and port come from configuration, not Beginify constants.

## In Scope

- Platform application-root landing and signup entry, with Sign Up as its only
  authentication/acquisition action; no general sign-in or tenant selector.
- Distinct core-owned platform signup authentication intent and bounded browser
  state, with verified provider proof, current account-status checks and no normal
  Organization session or existing-tenant admission before successful provisioning.
- Reviewed constrained production provisioning boundary, forward database changes
  where necessary, atomic new Organization/initial-owner/routing/audit outcome,
  reserved-address collision checks and bounded idempotent retries/concurrency.
- Minimal tenant name and platform-subdomain address setup; a fresh trusted
  target-host handoff and normal origin/Organization-bound tenant session.
- Necessary core, SQL/RLS/grants, HTTP/BFF/UI, configuration, audit, recovery,
  bounded abuse controls and focused automated/browser evidence and documentation.
- Narrow prospective replacement of G016 R2's root/apex/`www` public marketing
  redirect requirement. Keep its accepted security and protocol-routing contracts.

## Out of Scope

Billing, pricing, payment or commercial onboarding; mandatory Group or Business
Profile creation; Organization management toggles; automatic invitation emails;
existing-tenant grants; public tenant directories or account/tenant discovery;
new identity vendors, issuers or abuse services; marketing-site implementation;
custom-domain auto-verification or customer apex/ALIAS support; production rollout,
live WorkOS/DNS/TLS mutations or infrastructure provisioning. G059 stays deferred.

## Dependencies

Accepted G016/G018/G019/G058/G060 contracts and transitive identity, tenancy,
authorization, audit, security and recovery controls. G055 is earlier in the
owner-directed delivery sequence, **G055 → G062 → G020 → G057 → G021**, but is
not a technical prerequisite for G062 planning. The owner explicitly released
provisional planning now, without waiting for G055's MR or final ACCEPT. G062
ACCEPT precedes G020 execution; the order is a schedule gate, not a newly assessed
technical dependency. G057 remains mandatory before G021; G016 remains accepted.

## Completion Definition

Spec AC1–AC10 have attributable source and focused automated evidence, including
real-browser local HTTPS host transitions; required source/Duo/feature-MR fast
review, exact merged-SHA coordinator-owned full local qualification and scoped
cleanup are complete, followed by final supervisor ACCEPT. Actual provider behavior
is separately qualified through the reconciled G057 live evidence before G021;
local/emulator evidence never becomes a live PASS. G021 independently reviews the
integrated foundation, including this accepted generic capability; the product
feature barrier remains closed until G021 ACCEPT.

## Execution Policy

Architecture/security/data tier: explicit Claude Plan and security/architecture
review are required. Owner-authorized provisional planning is released; Claude may
inspect and prepare the G062 Plan and planning inputs under its existing scope,
budget and stops. This does not release implementation, test campaigns, provider
probes, deployment or infrastructure actions. Preserve active task bindings and
current owner stops. Execution requires Codex Plan ACCEPT, exact scope and inputs,
satisfied technical dependencies, budgets, resource readiness and a recorded
supervisor release. Major product/security/architecture changes remain owner-owned.

The later Plan must assess feasibility of the coupled platform authentication,
least-privilege provisioning and tenant-host transition before execution. The
existing privileged `CreateOrganizationWithOwner` has no production caller/route;
exposing it directly or granting runtime blanket Organization INSERT is forbidden.
Do not reuse G058's Organization-bound applicant credential as platform authority
without a reviewed design that preserves its original boundary.

## Iteration Budget

Provisional **B5: five Goal-level attempts, checkpoint after attempt3** under
BUDGET_POLICY (not included); zero implementation attempts spent.
Reassess at Plan review and propose a bounded split or budget amendment if the
security-critical transition cannot fit. No weaker acceptance or automatic reset.

## Context / Scope Budget

One platform signup intent, one atomic tenant bootstrap and one target-host
completion path; reuse existing WorkOS, G016 host/session, G019 ownership, G058
proof and limiter controls. **150k context checkpoint / 200k safe transfer** for
this new Goal; transfer preserves the same task, attempt and bound inputs.
One implementation writer by default and a finite two-tenant negative matrix;
do not expand into general onboarding, provisioning infrastructure or CI redesign.

## Suggested Subagents / Delegation

Claude owns the concise Plan and eventual implementation branch. Prefer one
bounded read-only security review of platform identity versus tenant authority,
runtime SQL provisioning and cross-host handoff before broader UI composition.
Use existing coordinator/review workflow and task resource ownership; no new
persistent monitor or duplicate coordinator. Delegation shares this Goal's budget.

## Stop / Escalation Conditions

Stop affected work for ambiguous platform-versus-tenant credentials; unverified
provider proof; stale disabled-account eligibility; arbitrary actor/role/tenant
assignment; non-atomic bootstrap or unknown outcomes treated as success; slug or
reserved-host takeover; open redirect or cross-host cookie copying; widened runtime
DML/privileged credentials; custom-host security changes; architecture/ADR conflict;
unreviewed migration/API expansion; new vendor; native refusal; exhausted budget
or inability to collect required evidence. G016/G057 accepted scope and qualification
must not be deferred wholesale to make signup appear complete. Escalate material
product/architecture choices; leave routine layout/form sequencing to the Plan.
