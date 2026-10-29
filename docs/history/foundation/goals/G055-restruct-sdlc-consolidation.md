> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G055 — Restruct SDLC skill-suite packaging and distribution

Revision: r3, owner-directed minimal bundle, 2026-10-20.
Status: **FINAL ACCEPT — 2026-10-20**, with the owner's qualification exception recorded in the final review. The full-suite failure remains failed in the evidence; the separate certifi maintenance does not reopen G055.
Epic: EPIC-001 — Repository and agent-workflow follow-up.
Owner: Codex Goal/Spec/review; Claude Code Plan/implementation/validation.

## Objective

Distribute the three working skills together as **Restruct SDLC**, with a thin
`restruct-sdlc` parent routing to their preserved child `SKILL.md` files. Prioritize
minimal disruption and reuse of proven behavior over the proposed v2 consolidation.
This remains G055; no replacement task or acceptance of the unmerged candidate.

## In Scope

- Canonical `skills/restruct-sdlc/` with parent `SKILL.md` and complete named children
  `sdlc-management/`, `repository-supervisor/`, `claude-session-handoff/`.
- Preserve child instructions, helpers, references, templates, names, behavior,
  notices and accepted workflow changes. Include the owner-directed rule that a
  supervisor relies on the Claude session-handoff watcher and does not poll the
  coordinator unless specifically asked; the watcher already polls. Only necessary
  relocation/path/packaging adaptations plus this clarification of existing communication behavior;
  classify any other correction before implementation.
- Ship the full `sdlc-management` adoption/bootstrap capability: root `AGENTS.md`/
  `CLAUDE.md` templates and existing root guidance, an optional project-scoped
  `.codex/config.toml` model-default template, plus the complete release
  management tree, including `ROADMAP.md`,
  `PROGRESS.md`, `BACKLOG.md`, `goals/`, `specs/`, `plans/`, `results/` and dated
  `archive/` templates. An adopter can initialize this whole layout by explicitly
  choosing adoption; package discovery alone creates no files or sessions.
- Root `.agents/skills/restruct-sdlc` and `.claude/skills/restruct-sdlc` relative
  links to the canonical package. Parent explicitly loads the needed child;
  nested automatic discovery and separate child installation are not required.
- Retire old standalone source/discovery paths after active-reference/runtime
  reconciliation; one maintained source per child. Preserve unrelated skills/history.
- Export the same real-file bundle to owner-moved `../../Restruct/sdlc-skill`, with
  ordinary README/license/inventory material. Installed skill name remains `restruct-sdlc`.
- Focused existing checks for packaging, routing, discovery, installation and
  relocated helpers; reuse attributable unchanged validation evidence.

## Out of Scope

Merging child instruction bodies; new transport, code-binding, installer, registry,
update or home-runtime frameworks; changed CLI/security semantics; application,
foundation, schema, API, CI or personal settings changes; new sessions; private-state
migration/reset; unrelated skills; a mandatory 2.0.0 release; public tags/releases;
renaming the GitHub remote repository.

## Dependencies

G049/G051/G052/G053 and G060 are accepted. Order remains
**G060 → G055 → G020 → G057 → G021 → OS001 → OS002**. G054 is deferred; G021
is not a G055 prerequisite. G020 provisional planning remains accepted but its
implementation waits G055 final ACCEPT and readiness release. OS001 stays paused
and later bundles this accepted suite as optional tooling.

## Completion Definition

[Spec AC1–AC10](../specs/G055-restruct-sdlc-consolidation-spec.md#acceptance-criteria)
are evidenced; Codex accepts actual source/export diff; required integration,
qualification and scoped cleanup complete. Both authorized repository pushes are
verified. Existing Result retains original failures, partial acceptances and
superseded-design work; a merge alone does not complete G055.

## Execution Policy

Security-sensitive tooling relocation with bounded Plan review. Owner authorizes
respec/replan now. One existing coordinator produces a concise linked Plan revision
(r4 recommended), refreshed inventory/allow-list and criterion mapping. Preserve
prior Plan/review/Result bytes/history. Earlier r3 implementation releases cover
the superseded consolidation only; this revised implementation needs reviewed
Plan/scope/bindings and recorded release under the existing delivery mandate.
The Plan must classify this clarification of the existing supervisor communication rules
in the bundled coordination/supervisor instructions and adoption templates. It
does not authorize changing the watcher implementation or adding a new monitor.

Use current committed standalone skills on staging as the behavioral baseline.
Inspect unmerged v2 work only for reusable accepted policy/path changes; do not
ship new mechanisms simply because they were implemented. Preserve existing
candidate commits and Result history; never force-rewrite staging. Reconcile any
live helper/observer executing from a retired path at a supported safe point.

The owner confirms moving the outside checkout to `../../Restruct/sdlc-skill` and
explicitly permits completely replacing its contents and discarding pending changes
there if needed for this bundle. This supersedes its dirty-overlay, README-wording
and stage-only export holds. Inventory first; preserve `.git`, committed history,
remote configuration and unrelated/private state. After Plan/review gates, normally
commit/push reviewed bundle changes. No history reset/force push, remote rename or
public tag/release is authorized. This permission does not cover other checkouts.

## Iteration Budget

Keep **B5, checkpoint after attempt 3** and the shared attempt/correction ledger.
Latest Result reports attempt 1 of B5; Claude verifies actual usage in the revised
Plan. Planning/respec spends no implementation attempt and resets none. Escalate
actual exhaustion before execution rather than silently extending the budget.

## Context / Scope Budget

One bundle, one parent entrypoint ≤200 physical lines, three preserved children;
no new child line cap or instruction rewrite. Apply current 150k checkpoint/200k
handoff and supported safe compaction. Prefer task/run-scoped ignored `.local/tmp/`
for accessible scratch over generic `/tmp`; preserve bound home-local evidence/
secret roots and supported worker fallbacks. Disposable HOME fixtures may retain
controlled temp roots. No scratch/private state enters distribution.

## Suggested Subagents / Delegation

One Claude Plan/implementation/export owner with current model-aware defaults.
Bounded read-only review or existing disposable helpers only when needed; no extra
coordinator/monitor or duplicated implementation. Codex owns final disposition.

## Stop / Escalation Conditions

- Behavioral/security/CLI changes beyond necessary relocation, weakened integrity,
  target-isolation, role or native permission boundaries.
- Unsupported parent discovery/child execution requiring substantive redesign;
  live runtime still uses a path scheduled for removal.
- Unsafe collisions, unrelated/private work at risk, unapproved releases, budget
  exhaustion or missing required evidence. Record complete failure inventories.

## Revision history and supersession

Owner decision/replan handoff (not included).
Exact pre-bundle Goal (not included)
preserves all earlier amendments. Prior Plans/reviews/Result remain history; their
consolidation-only requirements and holds are superseded by this revision.
Current repository governance and accepted workflow policies remain authoritative.

<a id="permission-layers-and-portable-defaults--2026-10-18-utc"></a>
Historical permission amendment: preserved original (not included).

Owner clarification, 2026-10-20: G055 must also preserve the reusable live-guidance
update rule: whenever a supervisor updates Restruct SDLC workflow rules, update
the adopting repository's corresponding live governance files in the same batch
and commit/push to its configured integration branch under its publication policy.
This applies to all adopting repositories, preserving local settings and task
bindings. Include this alongside the no-poll rule in Plan reconciliation; no
watcher implementation change or execution release is implied.

Owner direction, 2026-10-20: apply future Restruct SDLC workflow changes to
the corresponding live guidance in every adopting repository in the same update
batch, publishing to its configured integration branch under local policy. G055
must package and preserve this rule. Reconcile it with the no-poll update in the
next Plan submission; implementation remains held pending Plan review/release.

### Owner clarification: immediate asynchronous coordinator handoffs — 2026-10-20

The supervisor sends needed coordinator decisions, corrections and requests
immediately using nonblocking delivery, then continues its work without waiting
for a reply. The existing watcher delivers acknowledgment and responses
asynchronously. No-polling restricts status checks, not outbound communication.
G055 must preserve this clarification in its bundled guidance and templates.
