> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G069 — Adopt the complete, unaltered Vestive Brand kit

Date: 2026-10-26 UTC
Status: final ACCEPT — 2026-10-27. See the final acceptance record below.
Owner: Codex Goal/Spec/review; existing Claude coordinator Plan/implementation/validation/Result.

## Objective

Make SaaStacular use the supplied Vestive Brand kit directly: all 69 files in
`kit/src` copied byte-for-byte into application source, with the application
adapted to the kit's component APIs, utilities, hooks, styles and theme conventions.
The owner explicitly requests the complete catalogue, including unused components,
and accepts `cmdk` and its necessary transitive Radix dependencies.

## In Scope

The [Work Spec](../specs/G069-unaltered-brand-kit-spec.md) defines the immutable
source, application consumer migration, dependencies, standard theme integration,
tenant branding, affected validation and source fidelity requirements. Bounded
Core branding default/validation alignment is included if the new semantic roles
require it. Update applicable adoption/architecture documentation and static
oracles to the approved contract, preserving security and accessibility guarantees.

## Out of Scope

New product features or pages merely to exercise unused components; a kit fork;
remote assets, redesign of the test harness, new cleanup frameworks, deployment,
DNS/provider actions, M1 refresh/teardown, or changes to G057 frozen qualification.

## Dependencies

Initial repository baseline: staging `57d40b8eb8130eb09619d13f498d62393a877d3a`.
G067 and G068 are accepted. Consume G057's already merged Part A (`c840d714`),
reconciling actual later source deltas at a safe boundary. G069 is an independent
UI task: G057's remaining local/Part B qualification is not its planning or
integration prerequisite. Neither task can claim the other's unfinished coverage.
Both final G057 and G069 ACCEPT are required before G021 foundation acceptance;
no product-specific work before G021. RC15 promotion remains independent.

## Completion Definition

Every Spec acceptance criterion has attributable evidence, the implementation
has passed applicable source/hosted review and affected merged-SHA checks, and
Codex accepts the Result and all-category cleanup. No broad full-suite claim is
required or inferred. Publish closeout directly under supervisor authority.

## Execution Policy

Architecture-affecting migration; Claude Plan and Codex review are required.
The owner's instruction "please spec this and we can start planning this" releases
planning now, not unreviewed implementation. Use the current repository model/
effort policy. Plan ACCEPT and recorded execution release start implementation
without another routine owner permission request when all actual gates are clear.

## Iteration Budget

B5: at most five Goal execution attempts, checkpoint after attempt 3, under
BUDGET_POLICY (not included). Planning spends zero attempts. Parallel
workers share this allowance; individual tests are not independent Goal attempts.

## Context / Scope Budget

One source adoption and consumer/theme migration, not a UI product redesign.
The Plan must inventory APIs, dependencies, CSS/portal integration and affected
contracts before dispatch. Give an evidence-based wall-clock estimate, identifying
parallel coding, sequential builds and validation time separately. Escalate genuine
source incompatibilities together before broad consumer edits, not repeated
speculative review expansions.

## Suggested Subagents / Delegation

Begin independent read-only planning lanes in parallel: (1) all-source dependency,
alias and build compatibility; (2) appearance, tenant branding, portals and Core
alignment; (3) consumer migration and affected validation. One planner integrates
one Plan. Reuse surviving assigned workers; do not create another coordinator.

The implementation Plan must likewise dispatch dependency-ready disjoint work
in parallel, assigning one owner per file/lockfile/theme contract and one integration
owner. Establish shared APIs/aliases/theme boundaries before dependent consumer
lanes. Use native isolated worktrees, current two-run private test pool and sequential
image builds; protect M1. Reclaim finished worker trees/scratch after preserving
unique source, with resource accounting at each handback and batch end.

## Stop / Escalation Conditions

Stop the affected lane for an unavoidable upstream-source edit, dependency/security
conflict, inaccessible rendered design that cannot be corrected through supported
theme configuration, breaking persisted-branding semantics, unsafe shared resources,
or scope/budget exhaustion. Report the exact conflict and smallest options; do not
silently patch kit files, mask errors, weaken authorization, or treat unused cosmetic
catalogue defects as automatic whole-task blockers. Preserve unaffected progress.

## History

- 2026-10-26 — Owner selects full unaltered source adoption, accepts the cmdk
  dependency consequence, authorizes this Spec and parallel Claude planning.

## Final acceptance — 2026-10-27

Codex final ACCEPT recorded in the [Result](../results/G069-unaltered-brand-kit-result.md).
MR !108 is merged at `09440305`; owner-authorized qualification and cleanup are
recorded there. RC16/RC17 did not qualify and remain unpromoted. All G069 task
resources are cleaned; the verified bundle is intentionally retained as the
named archive of unique historical branch commits. The G069 predecessor hold on
G071 is lifted; G071's own hosted qualification and review gates remain.

## Later owner amendment — exact kit visuals, 2026-10-26 UTC

The owner explicitly waives contrast shortfalls caused by the exact shipped kit
visuals and requires removing app-owned visual corrections. The
[Spec amendment](../specs/G069-unaltered-brand-kit-spec.md#8-owner-amendment--exact-kit-visuals-and-contrast-waiver-2026-10-26-utc)
supersedes this Goal's conflicting accessibility/contrast stop wording for that
scope. The preceding supervisor hold is lifted. Preserve all kit bytes, existing
application functionality, tenant branding and security boundaries; record
waived measurements honestly. The full hosted pre-merge suite replaces a full
post-merge repeat on matching tested/merged executable trees. This continues the
current implementation attempt and requires review of the resulting source delta;
it is not final Goal ACCEPT.
