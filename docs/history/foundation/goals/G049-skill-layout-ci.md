> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G049 — Agent skill discovery, foundation naming and CI classification

Epic: EPIC-001 — Repository and local platform follow-up
Status: Draft r2; Claude Plan and implementation not authorized
Implementation owner: Claude Code
Supervisor and acceptance reviewer: Codex
Budget: B3 — maximum three shared meaningful cycles; checkpoint after cycle 2

## Objective

Separate agent-development guidance from the generated SaaS foundation deliverable. Put the handoff skill in Codex's repository discovery directory, scope Vestive guidance to web with Claude/Codex discovery, rename the generated archetype to saas-foundation, and prevent handoff-only edits from consuming application CI while retaining foundation gates.

## In Scope

- The three exact relocations and one relative discovery symlink in the [Work Spec](../specs/G049-skill-layout-ci-spec.md).
- Canonical archetype skill name/frontmatter, live path references and minimum existing renderer/harness/protected-path reconciliation.
- Narrow workflow/profile-map changes and covering controls: root `.claude/` and `.agents/` select fast, the three known handoff files bypass automatic CI, and all root `skills/**`, any project-subtree `.claude/skills/**` and the Vestive discovery link retain full. Existing CI job/runner architecture remains.
- Relocation provenance, discovery checks, inherited relevant final validation and a separate Result.

## Out of Scope

- Application behavior, component/style/font changes, schema, authorization, provider or dependency changes.
- Updating the archetype's domain model/templates to later architecture decisions; rewriting preserved Vestive kit content.
- Revising historical Plans, accepted reviews/Results, snapshots, render manifests or G004 execution inputs.
- General skill categorization, agent configuration bypasses, runner/harness redesign, CI benchmark campaigns or unrelated cleanup.
- Personal skill reinstalls, external repository renames, OS001 export execution or product work.

## Dependencies

- G004 implementation ACCEPT and integration into staging before implementation; accepted G000–G003/G047/G048 history retained.
- Reassess the actual integrated baseline, including G004's final harness changes. Do not infer a clean baseline from an older planning checkout.
- Canonical governance, architecture/security/accepted ADRs, roadmap/progress, BUDGET_POLICY (not included), current skill files, CI map/workflow and harness.
- Separate concise Claude Plan, Codex Plan ACCEPT and fresh baseline/input binding using accepted procedures.

## Completion Definition

All nine Work Spec criteria pass. Claude's Result contains relocation/name diffs, discovery evidence, CI decisions/controls, exact final validation provenance, cleanup and B3 accounting. Codex reviews the actual diff and returns ACCEPT before marking G049 complete. Include the accepted layout in G021's foundation review; no product barrier is released.

## Execution Policy

Ordinary bounded repository/tooling execution with explicit Plan review, because this changes CI and provenance path handling. Drafting permits management artifacts and read-only inspection only. After approval, Claude may execute the reviewed relocations, minimum reference/guard changes and tests within scope. Reuse existing wrappers, harness and CI controls; do not create a new execution framework. Obtain hosted-action authorization through the existing procedure before triggering pipelines.

G049 does not automatically block independent approved G005 implementation. Coordinate with any active checkout that binds these paths: do not move its files or merge incompatible baseline changes into it. Later Goal release/reconciliation uses the accepted layout.

## Iteration Budget

**B3: at most three shared meaningful cycles; checkpoint after cycle 2.** Forecast two to three, conditional on an accepted integrated G004 baseline. Successful inspection/validation commands are not additional cycles. If the path/provenance changes require broader harness redesign, stop for scope review instead of extending the budget unilaterally.

## Context / Scope Budget

Three skill moves, one link, a bounded canonical-name delta and minimum existing path/CI controls. Cite unchanged execution guarantees rather than reproducing them. No separate comprehensive agent orchestration framework or new product/data contract.

## Suggested Subagents / Delegation

No implementation subagents expected. Claude may propose one read-only diff/provenance reviewer in its Plan; shared budget and final integration responsibility remain with Claude. This does not request Codex delegation.

## Stop / Escalation Conditions

- G004 is not accepted/integrated, baseline/input drift, overlapping active checkout or resource conflict.
- The proposed layout is not discoverable with the installed agents or requires more links/configuration than specified; return the compatibility finding to Codex before substituting a layout.
- Source bytes/permissions cannot be preserved except for reviewed naming/path references, or historical evidence must be rewritten.
- Need to weaken unchanged-skill/protected-path checks, broaden documentation bypasses, add vendors/dependencies or change application behavior.
- B3 exhaustion, repeated nonconvergence or required evidence unavailable. Report the blocker; no skipped acceptance or silent fallback.

## Revision Record

- 2026-10-08 — r1: owner requests a post-G004 specification for flat agent skill discovery, web-scoped Vestive, generated saas-foundation naming and targeted skill CI policy. This draft authorizes no implementation and leaves active G004 inputs unchanged.
- 2026-10-08 — r2: owner directs that root `.claude/` and `.agents/` must not trigger full merely by their location. Root-only agent changes select fast; the known three-file handoff exemption remains. Supersedes r1's root-unknown-agent full classification; budget, prerequisites and no-implementation gate unchanged.

- 2026-10-08 — Owner clarifies all root `skills/**` retain full validation, not only saas-foundation. No other relocation, budget or execution-gate change.

- 2026-10-08 — Subtree `.claude/skills/**` full classification is general, not Vestive-specific; root agent directories remain fast. No additional move or installation is in scope.
