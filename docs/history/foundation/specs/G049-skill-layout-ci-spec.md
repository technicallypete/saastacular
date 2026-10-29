> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G049 — Skill layout and CI classification work specification

Goal: [G049-skill-layout-ci.md](../goals/G049-skill-layout-ci.md)
State: Draft r2; post-G004 implementation requires separate Claude Plan ACCEPT
Authority: owner direction, 2026-10-08; canonical architecture, governance and accepted ADRs remain binding

## Goal and Rationale

Make development skills discoverable by the relevant agents without treating a session-handoff documentation update as an application/foundation change. Keep the generated reusable foundation deliverable under `skills/` with its intended saas-foundation identity. Current CI classifies every `skills/` path as foundation, and the harness hard-codes the existing archetype/Vestive paths and verifies unchanged skill content. Relocation must reconcile these guarantees explicitly rather than disable them.

## Scope and Prerequisites

Implement only after G004 ACCEPT and integration into staging, on a freshly reviewed baseline. Read its final Result/review and actual harness delta alongside G000/G001/G002/G003/G048 provenance and relevant validation before planning. Reuse accepted input binding, approved wrapper, CI controls, disposable-copy/resource protection and cleanup. No change to G004's execution checkout or frozen inputs is permitted.

### Required final layout

| Current path | Required destination | Purpose |
|---|---|---|
| `skills/claude-session-handoff/` | `.agents/skills/claude-session-handoff/` | Flat repository Codex skill; retain its three files and authoritative runbook routing. |
| `skills/vestive-design-skill/` | `web/.claude/skills/vestive-design/` | Flat Claude skill scoped to web; current `name: vestive-design` retained. |
| No current link | `web/.agents/skills/vestive-design` → `../../.claude/skills/vestive-design` | Relative symlink to the same Vestive source for Codex in web. |
| `skills/project-archetype/` | `skills/saas-foundation/` | Generated deliverable; frontmatter `name: saas-foundation`, current license retained. |

No category directory nesting and no second copied Vestive tree. The repository handoff move does not change the installed personal copy at `~/.agents/skills/claude-session-handoff/`. It is a Codex-to-Claude supervisor skill; a root Claude discovery alias is not required. Vestive discovery is scoped to work in web; a root session need not list it before entering/working in that subtree. Guidance for mocks/assets elsewhere may be read explicitly, but no additional installs are in scope.

## Requirements and Constraints

**R1 — Discovery and tracking.** Both agents must resolve their intended skill entrypoints on the supported Linux/WSL environment. Commit the relative symlink as Git mode `120000` and regular skill files as their original modes. Check tracked/ignored status explicitly; do not rely on untracked local `.claude` directories. A fresh clone must work without a link-creation bootstrap command. Document native Windows Git symlink requirements as a portability limitation, not a reason to duplicate content or change the supported runtime contract.

**R2 — Preserve source and history.** Capture source/destination file inventory, relative paths, Git modes and SHA-256 values. Relocations preserve bytes except for exact approved archetype name/title or live path references identified by the Plan. Preserved Vestive kit/assets/licenses remain byte-identical and its existing source checksum verification still passes. Retain old naming in archived Goal/Spec/Plan/Result/review records and immutable snapshots; add current mapping documentation rather than rewriting history.

**R3 — Naming compatibility.** Update the generated skill's canonical name and current user-facing references to saas-foundation. Update executable path consumers and relevant live docs. Do not globally replace `project-archetype`: existing manifest schema identifiers, managed-block markers, renderer output identity, hashes and accepted artifacts retain their compatibility semantics. No template/domain redesign is authorized. Confirm the renamed generator produces the same intended output and preserves existing no-overwrite/reapply behavior in disposable copies.

**R4 — Minimum provenance reconciliation.** Adapt existing renderer, source-asset checks, unchanged-skill/protected-path guards and approved-input checks to all new locations and the symlink. For this relocation, old-path removal is allowed only through the exact reviewed mapping plus source/destination evidence; it is not a general missing-file exception. Subsequent unchanged-content checks must cover the new canonical sources. Fail on unapproved source drift, source loss, duplicate mutable trees, unexpected/unbound files or a changed/escaping link. Preserve harness failure status, pre/post drift checks and secret/resource protections. Reuse accepted mechanisms; do not introduce a generalized migration registry or separate validation framework.

**R5 — Narrow CI policy.** Automatic push/MR pipelines are omitted when changes consist only of existing documentation exemptions and these three exact instruction/metadata files:

- `.agents/skills/claude-session-handoff/SKILL.md`
- `.agents/skills/claude-session-handoff/agents/openai.yaml`
- `.agents/skills/claude-session-handoff/references/local-session-handoff.md`

This skill has no executable helpers. New paths, including a script added beneath this directory, are not automatically exempt from fast checks. Mark these exact files as documentation for profile selection. All other paths under root `.claude/` and `.agents/` select the fast profile: with a valid changed-path comparison, root-agent-only changes execute fast and no full/integration suites. These root directory classifications do not extend to nested web paths. Any project-subtree `<subdirectory>/.claude/skills/**` requires full foundation validation, not only the current Vestive skill. This includes `web/.claude/skills/**`, `core/.claude/skills/**` and a future `docs/.claude/skills/**`; a Markdown documentation exemption must not hide a subtree skill. The exact Vestive Codex discovery symlink and all root `skills/**` also require full, including `skills/saas-foundation/**`. Root `.claude/` and `.agents/` retain their explicit fast classification; these agent configuration directories are not project subtrees for this nested-skill rule. Root `skills/` is the generated/deliverable skill area; this rule is not limited to one named skill. Unclassified web agent configuration or skill paths also require full validation. Keep existing non-skill infra/build/dependency/migration/security rules, mixed-change escalation, unknown-path and missing-base fail-closed behavior, and explicit web/API/scheduled full requests unchanged. Root `.agents/` and `.claude/` are exempt from automatic full selection by location, not from fast checks. No blanket pipeline bypass for those directories, YAML, Markdown or all skills. Explicit full requests and unavailable/invalid changed-path comparisons retain existing fail-closed behavior; do not infer root-only changes without a valid diff. With a valid diff, deletion of a root-agent path and renames wholly within the two root-agent directories also select fast. Evaluate both source and destination for cross-boundary renames; a move into runtime/foundation paths retains the higher required profile. Keep fail-closed deletion/rename behavior outside these explicit root-agent exceptions. The exact handoff-file pipeline bypass requires independent workflow controls for any deletion/rename it admits.

**R6 — Operational boundaries.** No production code/style/font/dependency/schema/identity/permission/public API/MCP behavior changes. No Skill source semantic updates to the latest domain architecture, vestive palette changes, obsolete guidance cleanup or expansion of adopted components. Marketing-copy and dark-only rules stay excluded; accepted Radix/appearance/security contracts override historical source guidance. Update only necessary live references and discovery/CI documentation. No runner migration, GitLab settings change, pipeline cancellation or branch deletion.

## Expected Artifacts

The relocated skills/link, canonical naming metadata, minimum existing path/guard/workflow controls, necessary live documentation, Claude's concise Plan and Result, and attributable external evidence. The Plan enumerates concrete changed paths and the exact allowed byte/name-reference differences. No application artifacts or new dependencies are expected.

## Acceptance Criteria and Verification

| ID | Required outcome | Verification / evidence |
|---|---|---|
| AC1 | Execution starts only from G004's accepted integrated baseline with approved G049 inputs. | Record G004 review/integration references, base commit/tree, exact approved inputs and successful existing pre/post binding checks; no writes to active predecessor checkout. |
| AC2 | All four required layout entries exist, old three source locations are removed, and fresh-checkout discovery works. | Inspect tracked paths/modes and `readlink`; in a disposable clone verify Codex root handoff discovery and both agents' web Vestive discovery using supported installed interfaces. Record versions/entrypoints; no skill invocation or separate implementation task required. |
| AC3 | Every relocated file/mode is accounted for, preserved sources match, and historical records remain intact. | Compare mapped inventories/hashes/modes with only Plan-listed name/reference differences; run existing Vestive source checksum check from its new directory; diff archives/snapshots/accepted evidence against base. |
| AC4 | saas-foundation is canonical without changing generator output or reapply safety. | Validate skill frontmatter; run existing renderer/repeat/conflict/provenance checks from the new path in owned disposable copies; compare output to the accepted manifest/template contract. No schema/marker compatibility rename. |
| AC5 | Updated provenance guards detect source and link tampering. | Covering disposable negatives: alter one preserved skill byte, remove a required file, add an unbound source file, and change the link target outside its approved source. Each must fail the relevant guard; intended relocation passes. Do not remove existing controls. |
| AC6 | Handoff-only automatic pushes/MRs create no application CI, and explicit full requests retain full selection. | Existing workflow/profile controls cover each of the three exempt files, combined with management/docs, plus explicit web/API/schedule cases. Check workflow admission and profile selection separately; automatic cases omit pipelines, explicit cases retain full. |
| AC7 | All root skills, project-subtree Claude skills, the Vestive discovery link and unknown web agent paths trigger full; root agent paths select fast; mixed changes retain their affected profile. | Controls for root `.claude/settings.json`, root `.claude/skills/example/SKILL.md`, root `.agents/skills/example/SKILL.md`, root unknown agent configuration and a new handoff helper script: each admits automatic fast checks and selects fast with no full/integration execution. Vestive CSS/kit file, archetype generator/template, a newly added `skills/example/SKILL.md`, new `web/.claude/skills/example/SKILL.md`, `core/.claude/skills/example/SKILL.md` and `docs/.claude/skills/example/SKILL.md`, its discovery link and unknown web agent config select full. Each nested-skill case must admit automatic push/MR pipelines as well as select full, including the docs case. Root-agent changes mixed with runtime select integration, and mixed with lockfile/infra select full. Test push/MR contexts, root-agent deletion and within-root rename (fast), and root-to-foundation rename (full). Retain missing-base and all other deletion/rename fail-closed controls. |
| AC8 | Final content passes required inherited foundation validation without application behavior changes. | Run focused path/skill/CI/provenance controls first, then one required successful final local full run and one authorized hosted fast/full run against the final candidate. Reuse predecessor evidence only for unchanged contracts and exact inputs; passing old application tests cannot substitute for the changed harness/CI validation. Record failures/interruptions and rerun only for relevant changes/failures. |
| AC9 | Handoff documentation and Result provide reproducible provenance and completion accounting. | Live docs state canonical layout, relative link/discovery behavior, Windows limitation, root fast-only policy, exact handoff pipeline exemption and historical mapping. Result records commits/tree, source/destination hashes and permitted deltas, checks/commands/controls, final hosted IDs/results, no application delta, owned-resource cleanup and B3 cycle/checkpoint history. Codex reviews actual diff before ACCEPT. |

## Permitted Deviations and Gates

Permitted adaptations are the relative discovery link, exact archetype skill name/title, reviewed live path references and minimum existing guard/CI controls. File transfer tooling, internal helper naming and test organization belong to the Plan. Layout, source-content or CI-policy substitutions require Codex review; no broader fallback is preapproved.

Before implementation, the Plan must resolve installed-agent discovery support, tracked link behavior, concrete hard-coded/protected path consumers, exact relocation delta and how existing controls verify the new workflow exemption. If native agent discovery cannot be demonstrated without unreviewed configuration, stop for a layout decision. No unresolved major architecture/vendor choice is expected.

G004 acceptance/integration, Claude Plan ACCEPT, fresh binding and hosted-action authority are prerequisites to their dependent actions. Completion requires AC1–AC9 and Codex implementation ACCEPT. B3/checkpoint 2, other Goal dependencies and the G021 product-feature barrier remain in force; this Spec starts no G049 or G005 implementation.

## Revision Record

- 2026-10-08 — Draft r1 defines the owner's proposed skill layout and narrow CI policy for execution after G004. Historical source naming and active execution inputs are preserved; no implementation performed.
- 2026-10-08 — r2: owner excludes root `.claude/` and `.agents/` from path-triggered full validation. R5/AC7 now require fast for other root agent paths and preserve the exact three-file automatic handoff bypass. This supersedes r1's root-unknown-agent/full and new-handoff-script/full requirements only. Web foundation paths, mixed-change escalation, explicit full, invalid-diff fail-closed behavior, nine criteria and B3 are unchanged.

- 2026-10-08 — Owner clarifies the full rule covers all root `skills/**`, not only the renamed saas-foundation skill. R5/AC7 include a new unrelated root skill control; root agent-directory and web discovery policies are unchanged. No additional skill move or implementation scope.

- 2026-10-08 — Owner generalizes the subtree Claude skill rule beyond Vestive. R5/AC7 cover any project-subtree `.claude/skills/**`, including beneath docs where the ordinary Markdown exemption must not suppress CI. Root agent directories keep fast; all root skills keep full. This classification authorizes no additional skill installation or relocation.
