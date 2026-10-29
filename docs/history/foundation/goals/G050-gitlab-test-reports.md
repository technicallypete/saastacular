> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G050 — GitLab test-report visibility

Epic: EPIC-001 — Repository and local platform follow-up
Status: Defined for Claude planning; implementation not authorized
Implementation owner: Claude Code
Supervisor and acceptance reviewer: Codex
Budget: B3 — maximum three shared meaningful cycles; checkpoint after cycle 2

## Objective

Expose existing CI check outcomes in GitLab's Tests tab through JUnit XML, without running additional application tests or changing gating behavior.

## In Scope

- A small reporting adapter for existing validated check outcomes and CI phases.
- JUnit report publication in existing fast, gate and protected identity-staging jobs when applicable.
- Focused conversion/publication/privacy controls, necessary existing guard reconciliation, minimal operations documentation and a Result.

## Out of Scope

- New application tests or test executions solely to populate the Tests tab; exporting every nested browser/unit case.
- Test selection, workflow/job triggers, runners/images/dependencies, schedules, identity qualification or security-model changes.
- Harness redesign, generalized reporting/orchestration, schema/product implementation or G049 skill relocations.

## Dependencies

- Accepted G003 reporting, profile-selection, privacy and resource-protection controls; ADR-002 remains binding.
- G004 accepted and integrated into staging before execution, so the real baseline includes identity-staging and reconciled CI controls.
- Read canonical governance/architecture/security/decisions, roadmap/progress, BUDGET_POLICY and actual integrated CI/report publication paths.
- Separate concise Claude Plan ACCEPT and exact baseline/input binding through existing procedures.
- Coordinate with G049 if it changes shared CI files. Reconcile its integrated changes before final validation when applicable; neither Goal creates a blanket blocker for independent G005 implementation.

## Completion Definition

AC1–AC7 in the [Work Spec](../specs/G050-gitlab-test-reports-spec.md) pass, the Result records exact evidence and budget accounting, and Codex accepts the actual diff. Reporting is supplemental: existing TSV evidence, exit statuses, required checks and G021's product-feature barrier remain authoritative.

## Execution Policy

Ordinary bounded tooling execution with a short Claude Plan and explicit Codex approval. Planning permits read-only inspection and goal-specific management artifacts only. After approval, implement the reviewed adapter/configuration and focused controls using existing execution and publication safeguards. No new framework or additional provider actions. Hosted runs require existing action authorization; do not launch duplicate pipelines for demonstration.

## Iteration Budget

B3: maximum three shared meaningful cycles, checkpoint after cycle 2; forecast one to two. Diagnose focused reporting/control failures before required broad validation. Stop on exhaustion; no unilateral extension or weakening checks.

## Context / Scope Budget

Existing harness-level check IDs and bounded phase outcomes only. Keep the Plan short: concrete files, status mapping, privacy/publication integration, criterion mapping and validation. Reuse accepted controls and cite their guarantees rather than recreate them.

## Suggested Subagents / Delegation

No implementation subagent expected. One read-only reporting/privacy reviewer may be proposed; shared budget and final integration remain Claude's responsibility.

## Stop / Escalation Conditions

- Missing integrated predecessor, input drift or overlapping uncoordinated CI edits.
- Need to publish unsanitized details, weaken artifact scans/access, change exit statuses/triggers or add dependencies/vendors.
- Required generation cannot run with existing container tools, or cancellation/partial-output handling requires a harness redesign.
- Reporting requires live provider qualification, application changes or new authority.
- B3 exhaustion, unavailable required evidence or repeated material failure. Return findings to Codex rather than skip criteria.

## Revision Record

- 2026-10-08 — Owner requests a separate bounded reporting Goal after discussing GitLab's empty Tests tab. Planning definition only; no implementation or hosted action authorized.
