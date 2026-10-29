> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G067 — Adopt Vestive Brand and refresh the generic SaaStacular UI

Date: 2026-10-25 UTC
Owner: Codex Goal/Spec/review; Claude Code Plan/implementation/validation/Result.
Status: owner-expanded Base UI/design overhaul authorized; planning released immediately; implementation follows Plan ACCEPT and ordinary readiness. Runs before G057.

## Objective and owner authority

Replace the repository's old Vestive Design input with the owner-supplied
Vestive Brand v0.2.1 skill and apply its successor design to the existing generic
SaaStacular application. The owner directs this task to start now while they
test the existing staging preview, and to complete before G057. This authorizes
the Goal/Spec and normal reviewed delivery lifecycle, rather than definition only.

Owner source clarification: the new skill was created from Claude Design.
Identify important old guidance worth carrying forward and report it; do not
carry over superseded old palette, typography or geometry as new requirements.

## Scope

- Copy the supplied home `~/.agents/skills/vestive-brand` into a real, tracked
  repository `.agents/skills/vestive-brand` directory; give Claude a relative
  `.claude/skills/vestive-brand` alias to that repository copy.
- Reapply the new visual input to existing generic SaaStacular components and
  screens using the new Base UI/Rhea components, tokens, typography and geometry;
  migrate all affected consumers and remove Radix completely from the resolved
  application dependency tree.
- Preserve and document applicable old accessibility/integration knowledge.
- Update live agent instructions, source/provenance references and affected
  checks; remove the old repository skill and Claude alias at the completed
  cutover, preserving historical source through Git and existing evidence.

## Dependencies and parallel work

G020/G062/G063 and the existing G048/G049 UI/skill baseline are accepted.
G067 does not wait for completion of the owner's current preview. Planning and
later authorized implementation use isolated checkouts; preserve the preview's
frozen source, stack, volumes and resource reservation while it is being tested.
Use only existing available private test slots and sequential build reservation.
Do not stop, rebuild, rebase or reclaim that preview without owner direction.

G067 final ACCEPT precedes G057 implementation/integration and G021 foundation
acceptance. G057's accepted provisional r4 Plan remains a preserved design
baseline; reconcile its real inputs with integrated G067 and preview findings
before a separate G057 execution release. No G057 or external-action hold is
waived. Existing RC numbers and later roadmap ordering are unchanged.

## Execution policy

Claude prepares a proportional Plan now from the [Work Spec](../specs/G067-vestive-brand-successor-spec.md).
Return the exact Plan commit/hash, bounded consumer/file inventory, source
mapping, dependency/font deltas, executable focused checks and realistic time
estimate separating implementation, validation and queue/operator waits.
Codex reviews the Plan; approved implementation then continues under standing
staging-MR review/auto-merge and affected merged-SHA qualification policy.
No additional routine owner proceed prompt is needed for this authorized task.

The owner has now explicitly chosen the Base UI migration and a new UI overhaul,
using the new Vestive Brand source as the implementation/design baseline, with
useful old integration/accessibility carryover. ADR-001's dated owner amendment
records this successor target. The earlier Radix-retaining default and optional
decision are superseded. Completion requires no Radix application/test imports,
obsolete wrappers/configuration or direct/transitive resolved packages. Pin and
review all needed new source/dependency changes in Claude's Plan; migrate actual
consumers before removing the old primitive packages.

## Budget and boundaries

**B5: at most five Goal-level execution attempts; checkpoint after attempt 3**,
under `management/BUDGET_POLICY.md`. Planning/review currently spends zero
implementation attempts. Bound the work to skill intake/cutover and the shared
visual layer plus necessary existing consumers. Reuse the existing harness,
scoped theming, browser fixtures and resource inventory; do not invent another
design system, cleanup system or test campaign.

Independent intake/docs and web work may run in parallel with named ownership
and nonoverlapping writable paths after Plan approval. Each writer explicitly
reads the applicable repository skill and governance. No active preview resource
or immutable evidence checkout becomes a writer's development checkout.

Stop affected work and report an actual source-input mismatch, unresolved rights
for shipped assets, required architecture/token/public-contract change outside
scope, unavailable safe resource reservation, material strategy failure or budget
exhaustion. An unused source module's dependency list is not a required install.

## Owner scope amendment — Base UI overhaul, 2026-10-25

Owner: “we should migrate and use the new Vestive brand as is plus any additional
useful carry over from the old skill. This can be a new overhaul” and “this
should mean to competely remove radix correct?” Codex records this as explicit
architecture and scope authority for the complete Radix-to-Base-UI migration
within G067, before more SaaStacular work. The initial Radix-preserving Spec is
preserved as r1 (not included). Current
[Spec r2](../specs/G067-vestive-brand-successor-spec.md) supersedes that choice;
the budget, current attempt, preview protection, review and affected-validation
gates continue. This authorizes the migration without another owner library
decision; Codex Plan review still precedes implementation.

## Completion

Meet the Spec acceptance criteria with reviewed source, applicable current
hosted fast/Duo disposition, integrated affected local evidence and a concise
Result/cleanup receipt. Codex gives final ACCEPT and publishes closeout directly
to staging. Do not open a separate Result-only acceptance MR. Retain the owner's
preview and unrelated resources; clean finished task-owned resources at handbacks
and closeout. G021's foundation barrier remains closed.
