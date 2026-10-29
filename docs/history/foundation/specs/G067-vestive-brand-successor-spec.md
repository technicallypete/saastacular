> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G067 Work Spec — Vestive Brand successor for generic SaaStacular

Date: 2026-10-25 UTC
Revision: r2 — owner-authorized Base UI overhaul; exact initial r1 (not included) preserved.
Goal: [G067](../goals/G067-vestive-brand-successor.md).
Authority: owner requests a new task before G057, immediate work alongside the
existing staging preview, repository-local copy of the home Vestive Brand skill,
reapplication to generic SaaStacular, and removal of the old repository skill
when complete. Owner states the new input was created from Claude Design.
The owner now expressly chooses migration to its Base UI preset and a new UI
overhaul, including complete Radix removal. This supersedes r1's Radix-retaining
default and resolves the previous optional library question.

## Required outcome and inputs

The existing application adopts the new Vestive Brand v0.2.1 source components
and Base UI/Rhea preset as the successor generic UI foundation, with complete
Radix retirement. A visual imitation layered over the old Radix components
does not satisfy this revision. Keep configured SaaStacular identity and existing
functional/tenant boundaries. The old reference folder is removed from the
current repository tree once the successor is integrated and verified.

Initial inspection baseline: staging `ebb2742ad631cc270d8fb106d99716c1301c9c8b`.
R2 amendment observes staging `b6bfffb020188c2f0c0e81a16d50509e689b8d87`;
its delta is the G067 definition/sequencing publication, with no application
change. The supplied source and digests below remain the same.
Supplied directory: `<home>/.agents/skills/vestive-brand`, inspected as
319 regular files, 2,694,542 bytes, no symlinks. `SKILL.md` SHA-256:
`188ab915be8d8a98a3563477306c54cfb8cd162eaf34a805d7b4699443904662`.
Tree SHA-256:
`67f92e83e2cc61d24691baf2eebe5893261eebdc7056fc863d69897aa602ac88`.
Algorithm: sort regular files by relative POSIX path, concatenate UTF-8 lines
`<file SHA-256>  <relative path>\n`, then SHA-256 that byte sequence. Recheck the
input before copying; report changed bytes rather than silently rebinding them.
This records source identity, not license clearance or validation.

Read `AGENTS.md`, `CLAUDE.md`, architecture §37A/§61, SECURITY, DECISIONS,
ADR-001, `docs/foundation/WEB-SHELL.md`, G017 branding contracts and applicable
G048/G049 provenance/integration records. Read the supplied skill's `SKILL.md`,
`kit/README.md`, `design-system/README.md`, tokens, and the source of consumed
components. Existing feature acceptance and historical source hashes remain.

## Requirements

### R1 — Repository-local successor and completed cutover

Copy the complete supplied skill into `.agents/skills/vestive-brand/`, retaining
the useful bundled code/assets and directory structure. The source of truth is
the tracked repository copy, not a link into the owner's home directory or a
remote Claude artifact. Document exact source version/bytes and owner-reported
Claude Design origin in the existing Plan/Result; treat source code as input,
not automatically installed runtime dependencies.

Add the relative Claude alias
`.claude/skills/vestive-brand -> ../../.agents/skills/vestive-brand`.
At completed cutover remove `.agents/skills/vestive-design/` and its old Claude
alias. The owner authorizes deletion of these two repository entries as part of
this task after their history is preserved and consumers have migrated.
The home-directory skills and unrelated skill bundles are outside this cutover.

Update current discovery/instructions, web lint exclusions, relevant source
classification/validation fixtures, provenance checks and runtime documentation
to the new paths. Preserve the same safety guarantees; historical old-path
references in closed evidence remain historical. Resolve deleted-source citations
by explicit immutable Git commit/path references where active docs still need
them. Do not bulk-rewrite old Results or duplicate the retired skill as a second
active folder. Reusable Restruct guidance that does not name this skill needs no
unrelated publication/update.

### R2 — Adopt the successor Base UI/Rhea foundation and overhaul current UI

Adopt the new skill's actual application components and visual design as the
source baseline, using preset `b1Ha5oWoa` (Base UI, Rhea, Taupe, Geist, Lucide).
Migrate the repository-owned shared components and existing screens to it:
Taupe neutrals, semantic blue/gold roles,
successor spacing/radius/control treatment, Geist application typography and
Geist Mono where appropriate. Source app UI is plain Rhea with brand colors;
Oswald/Barlow/angled CTA/ribbon treatments are for appropriate brand moments,
rather than every app heading or everyday button. The Plan inventories current
generic consumers and maps old versus new role/shape/typography with a small
number of representative visual comparisons. Retire old treatments that
conflict with the successor; adding tokens nobody consumes is insufficient.

Apply through shared components and necessary existing consumers, including
the root/signup/sign-in and account/Organization/Group/label surfaces affected
by those components. Keep configured product identity, functional copy, routes,
form/error semantics, conditional feature behavior and current authorization.
Vestive logos, social icons, sales CTA, plan prices or marketing defaults do not
become SaaStacular identity. This task adds no product feature or new page flow.

**Migrate fully from Radix to Base UI.** The owner's explicit decision amends
ADR-001 for G067; retaining Radix under a new visual skin is no longer in scope.
Inventory and migrate all production, fixture and test consumers, including
Dialog/Slot/asChild composition, refs/events, portal behavior, primitive state
selectors and configuration. Internal APIs, component structure and obsolete
application CSS may be replaced to follow the new source rather than preserve
legacy styling. Update shadcn configuration to the actual compatible Base UI/
Rhea setup. Preserve needed functional scenarios when replacing selectors.

After cutover there must be no `@radix-ui/*` or `radix-ui` application/test
imports, obsolete Radix wrappers, active Radix preset/primitive configuration,
or Radix packages in the resolved application dependency tree, direct or
transitive. Check the actual lockfile/tree, not only top-level dependencies.
Choose Base UI-compatible replacements for any consumed dependency that would
otherwise retain Radix. Record legitimate historical/provenance mentions
separately; old decision/acceptance records must not be rewritten to remove
history. Remove redundant old packages once all consumers have migrated.

The complete source catalogue belongs in the copied skill; adopt the runtime
modules and utilities needed by the application's current surfaces, with exact
dependency versions and upstream notices in the Plan. Add no unused dependency
catalogue merely because the bundle's generic install command lists it. Adopt
the source stylesheet composition as appropriate, adapting its integration to
the retained appearance, scope, identity and accessibility contracts below.
Do not replace those guarantees or add remote font/asset requests by blindly
running the source's starter setup. New local font/assets need source hashes,
actual consumed purposes and the appropriate upstream notices.

### R3 — Carry forward useful guidance and adaptation fixes

Provide a concise repo-specific adoption reference under the new local skill,
linked from its entrypoint, retaining the meaningful old integration knowledge
while preserving the supplied upstream sources and documenting local additions.
Source blue/green button pairs are about 4.21:1/4.05:1 with white labels; medium
14px text does not qualify as large text and does not resolve the deficit.
Measure and adapt actual adopted states to existing application requirements.
Likewise verify/adapt clipped outside focus on angled CTAs, resting field
identification, minimum-width grids and reduced-motion coverage. Cite actual
new-source behavior; do not blindly import old component-specific bug claims.

Keep reusable accessibility cautions distinct from repository-specific rules.
Carry forward the latter explicitly: role-based semantic mapping, nested/partial
scopes, scoped portals, appearance/first-paint and local
asset rules. Old palette/Roboto/square-control/MDI instructions are superseded.
Use the new source's native component-slot semantics; where they differ from
existing override input meanings (for example neutral `accent` background versus
old gold accent ink), provide explicit translation/aliases at the integration
boundary. Legacy implementation naming must not force the old visual design
back into the new kit. Preserve the existing public override input contract.
Record important reusable gaps for the owner, including contradictory source
contrast guidance. Do not treat Claude Design origin as replacing required
third-party code/font/license notices; reuse established notices where valid,
and add the necessary lineage for newly consumed files.

### R4 — Preserve existing application and security guarantees

Preserve Light, Dark and System with System default, existing preference
storage/fallback, System-only OS following, pre-paint initialization and
hydration. Preserve the 16-token Organization override contract and partial,
nested and sibling scope behavior. Both overlay and content must enter their
intended scope from first insertion. Preserve keyboard interaction, visible
unclipped focus, dialog modality/focus return/inertness, accessible names and
errors, rendered contrast, text spacing, narrow-screen reflow and reduced
motion. Brand source values never waive these existing requirements.

No schema, tenancy/RLS, capability, identity/session, HTTP/MCP contract or
permission changes are required. Web remains a thin BFF with existing authority
boundaries. Maintain generic configuration/rename and intelligence-disabled
behavior. Historical G002/G047/G048/G049 evidence stays unchanged; record new
source/adaptation mapping rather than rewriting past hashes or acceptance.

### R5 — Deliver alongside the existing owner preview

The owner is testing an earlier staging preview in parallel. Preserve its frozen
checkout, containers/volumes and held shared resources. G067 planning and
authorized isolated implementation do not depend on that walkthrough finishing.
Use existing private resource slots and build reservation for later proofs;
respect a real resource conflict without stopping safe source work. Show the
successor via isolated representative screenshots/evidence as needed, rather
than silently replacing the preview the owner is using.

## Acceptance and required validation

| ID | Required outcome | Attributable evidence |
|---|---|---|
| AC1 | Complete repository-local successor, correct Claude alias, retired old directory/alias and working current references. | Source digest/copy comparison, tracked path/alias inventory and affected discovery/provenance/classification checks; historical references remain resolvable. |
| AC2 | Current generic UI adopts actual Base UI/Rhea source components, palette, typography and geometry; Radix is completely retired from active application code/configuration and the resolved dependency tree. | Complete source/consumer migration mapping, import/config and resolved direct/transitive dependency inspection, plus representative Light/Dark/narrow screenshots; verify local fonts actually render. |
| AC3 | Theme/branding scope, first paint, hydration and generic identity remain correct. | Existing targeted appearance/token/overlay, invalid-storage/System, partial/nested-scope and renamed-identity coverage affected by the change. |
| AC4 | Base UI replacement preserves adopted interactions and accessibility, including scoped portal insertion, modality/focus/ref/event behavior and SSR/hydration. | Relevant existing keyboard/dialog/forms/accessibility tests against the actual adopted Base UI version, rendered contrast/focus/reflow and reduced-motion checks in both appearances; migrate primitive-specific selectors without deleting their scenarios. |
| AC5 | Old integration cautions and new source lineage are documented without stale design mandates. | Local adoption note, consumed source/font/component notices and reviewed architecture/governance updates; no public remote asset requests. |
| AC6 | Normal integration/affected qualification and cleanup complete. | Reviewed current source, required hosted fast/Duo disposition, exact merged-SHA affected local coverage and Result/resource receipt; Codex final ACCEPT. |

Claude's Plan chooses concrete existing executable selectors and named owning
controls from actual changed behavior and consumers. Include necessary web
lint/type/unit and browser/appearance/token/overlay/design/accessibility cases,
plus affected skill-layout/source/provenance/classification cases. Include
renamed/disabled variants where the shared UI change affects them. Reuse
attributable passes and rerun failed/affected cases after fixes.

Do not select the whole `ci-controls` suite merely for skill path, source hashes,
component registration, fixture or count-pin updates. No blanket full local
run is required for G067. Any broader selection needs a concrete supervisor
coverage reason under current policy; hosted RC qualification remains separate.
Use the existing harness and available private slots; no additional test or
cleanup framework, runtime catalog installation or mandatory exhaustive kit
component campaign. Retain actual failures and unexecuted coverage honestly.

## Expected artifacts and gates

Goal/Spec here; Claude's proportional
`management/v0/plans/G067-vestive-brand-successor-plan.md` and
`management/v0/results/G067-vestive-brand-successor-result.md`; the local skill,
consumed UI/font changes and relevant governance/source references. Use the
existing task inventory and compact receipt for resources, source preservation
and cleanup. Current owner task authority releases planning now; Codex Plan
ACCEPT precedes implementation. Final G067 ACCEPT precedes G057 implementation;
G057 must reconcile integrated G067 and the owner's preview disposition before
its independent execution/readiness gates. G021 remains a later barrier.

## Initial definition review (r1) — 2026-10-25

Codex ACCEPTs this Goal/Spec after bounded read-only independent review and
supervisor inspection: no material corrections required. The source comparison
identifies the reusable accessibility gaps and repository-specific carryover in
R3/R4. The Plan must name the actual shared component/font inventory, source
mapping and focused executable selectors already required above. This releases
planning under the owner's instruction; it is not implementation acceptance or
a waiver of Plan review. No application source, assets or resources changed in
this definition batch.

## Revision and owner decision — r2, 2026-10-25

The r1 review above accepted the initial Radix-preserving definition only; exact
bytes are retained in the linked r1 file and publication `b6bfffb0`. The owner
then explicitly directs migration to the new kit as supplied, a new overhaul
now, and complete Radix removal. R2 makes Base UI/Rhea source adoption and zero
resolved Radix packages mandatory, permits the necessary component/API/CSS/
dependency migration, and records corresponding architecture authority. The
useful old integration/accessibility safeguards remain; obsolete visual and
primitive constraints do not. The existing B5/checkpoint3 ledger and preview
reservation continue; planning reconciles this amendment without another owner
library approval. Implementation still follows Codex's current-Plan ACCEPT.

R2 definition review: Codex ACCEPT after independent review and correction of
three remaining live architecture target references (preferred-stack diagram,
assistant-ui integration and foundation stack). Current target/ADR and
Goal/Spec now agree on actual Base UI adoption and complete Radix retirement.
No application code, assets, dependencies or test resources changed here.
