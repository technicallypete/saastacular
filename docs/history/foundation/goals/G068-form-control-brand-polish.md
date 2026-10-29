> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G068 — Form control brand treatment and motion polish

Date: 2026-10-26 UTC
Owner: Codex Goal/Spec/review; Claude Code implementation/validation/Result.
Status: direct presentation execution RELEASED in isolation; integration waits for G067 final ACCEPT.

## Objective and authority

Correct the owner's G067 preview findings: radio controls lack the adopted
Vestive Brand treatment, radio/label line heights and alignment are wrong, and
some form controls still lack expected subtle transitions. The owner calls
these nonblocking for G067 and directs "Do it" after the recommendation to
begin an isolated follow-up while G067 qualification completes. This authorizes
this named Goal/Spec and its delivery lifecycle only.

## Scope and inputs

Follow the [Work Spec](../specs/G068-form-control-brand-polish-spec.md).
Initial observed baseline: staging `6e7265b2` (G067 implementation merge
`bb605d5b` plus a management-only backlog entry). Reconcile later staging changes
in the development branch without rebasing G067's frozen qualification checkout.
Use the repository `.agents/skills/vestive-brand` and its ADOPTION rules.

Inventory the four native radio consumers (labels, enrollment, Group authority
and Group onboarding), their labels, and existing Input/Textarea/Select consumers
relevant to the reported motion gap. Correct confirmed presentation drift through
small shared components/classes and necessary call-site changes. Report
compliant controls without changing them; avoid a whole-application redesign.

## Execution policy and parallelism

**Direct execution**, under AGENTS.md's risk-based policy: no separate Plan file
or Plan-approval handoff. Claude starts with a compact inventory in the Result
identifying affected paths, source references and focused checks, then implements
within this presentation scope. Codex reviews actual source/evidence before
integration acceptance.

The existing coordinator owns G067 closeout and G068; it may dispatch one isolated
G068 writer. Start service work in `web/` after reading governance and the root
repository skill; record worker, branch and base. Start source work now without
waiting for G067 row 6. Protect its frozen SHA/run/evidence and the M1 preview,
volumes and reservations. Use the existing private-test pool and sequential build
reservation. No new coordinator, stack framework or test driver is in scope.

## Budget and stops

**B3: maximum three Goal execution attempts; checkpoint after attempt 2**, under
BUDGET_POLICY (not included). Use the current model/effort policy.
Stop the affected change for scope review if it requires dependency/lockfile,
token-contract, auth/authority/data/API, primitive-behavior, framework or pinned
source-kit changes, or an unsafe resource reservation. Preserve native form
submission, keyboard use, accessible contrast and visible focus.

## Completion and sequencing

Implementation may overlap G067 qualification; G067 final ACCEPT still precedes
G068 staging integration/final ACCEPT. G068 precedes G057; existing preview
reconciliation and G057 readiness/live-action holds remain. The G021 foundation
barrier and RC numbering/authority remain unchanged.

Complete the Spec under normal source review, hosted fast/Duo disposition,
standing staging auto-merge and attributable affected post-merge checks. Reuse
unchanged passes with exact tree/input attribution. No full local suite or full
CI-controls run is selected. Return a concise Result and all-category cleanup
receipt for supervisor publication directly to staging; no Result-only MR.
