# ADR-001 — Use Radix for the Vestive component foundation

> **Editorial note (public export).** This historical decision record is retained for its reusable architecture and security reasoning. Links to private planning, review and evidence records that are not part of this repository are marked "(not included)"; references to "Architecture §N" or ARCHITECTURE.md point to the source project's architecture document, which is not included. Goal IDs, dates, approvals and superseded history are kept as written; product-specific passages are redacted where marked "[Editorial redaction: …]". Current behaviour is documented in the foundation guides under [`docs/foundation/`](../foundation/). The brand agent skill this record mentions is not included in this repository; the UI kit source it selected ships unaltered in `web/src` ([WEB-SHELL.md](../foundation/WEB-SHELL.md)).

Date: 2026-10-06
Status: ACCEPTED — historical G048/G067 decisions; unaltered-kit target approved by owner amendment, 2026-10-26, through G069
Decision owner: Product owner
Recorded by: Codex
Implementation scope: G048; separate Claude Plan approval required

Current target: the owner selects the complete unaltered Vestive Brand kit
through G069, including cmdk's necessary transitive Radix dependency exception.
See the 2026-10-26 amendment below. Earlier G048/G067 decisions and their accepted
implementations remain historical; application behavior changes only through the
reviewed G069 cutover.

## Context

The accepted G002/G047 shell uses Base UI. The supplied Vestive React/shadcn component kit uses Radix. The owner requests actual source components, bundled fonts and blueprint backgrounds, then explicitly chooses Radix instead of porting the source back to Base UI. Preserving the source primitive composition reduces adaptation work; this is a project-specific compatibility decision, not a claim that one primitive library is universally superior.

The prior Base UI requirement is in the source architecture document's Frontend UI and Design-System Architecture (not included). That requirement must change explicitly, while preserving historical acceptance and the currently executing G003 baseline.

## Decision

1. Use **Tailwind CSS + repository-owned shadcn/Vestive components with Radix primitives** as the target frontend component foundation. React/Next.js remain unchanged. Future interactive generic primitives use Radix under their owning reviewed Goals; this does not install the whole catalogue.
2. G048 migrates the existing shell button/dialog and all their production/test consumers. Its permitted direct dependency delta is exact-pinned `@radix-ui/react-dialog` and `@radix-ui/react-slot` additions and `@base-ui/react` removal. Claude pins versions, compatibility, transitive effects and notices in its Plan; this ADR selects no floating/latest version.
3. Adopt the selected Vestive source components with Radix composition, subject to resolved provenance and necessary accessibility/token/content adaptations. Do not import the kit's default body portal, global theme/reset, marketing copy or external font loader as overriding application behavior.
4. Preserve Light/Dark/System with default System, pre-paint/hydration guarantees, semantic token contracts, scope-local override isolation and scope-container placement of both dialog overlay/content. Preserve keyboard/modality/focus/inertness, contrast, text spacing, reflow and reduced motion. No schema, authority, public HTTP/MCP or product feature changes.
5. The migration takes effect in implementation only through approved G048 after G003 acceptance/integration and fresh input binding. G003 continues with its existing Base UI implementation and approved inputs. Do not rewrite G002/G047 acceptance artifacts or claim those goals used Radix.
6. Completion leaves no Base UI imports or packages in the resolved application dependency tree. Any need to retain it, add other primitive suites/packages, break token contracts or alter appearance architecture returns to Codex before execution.

## Consequences and Validation

- Existing internal component call sites and primitive-specific test selectors may need migration. The Plan maps every affected consumer and old behavioral assertion to its Radix counterpart; no behavior scenario is dropped because library internals differ.
- Radix dialog portal container, focus lifecycle, modal semantics and SSR/hydration must be proven against actual adopted package versions. Official [Dialog](https://www.radix-ui.com/primitives/docs/components/dialog) and [Slot](https://www.radix-ui.com/primitives/docs/utilities/slot) references describe the relevant APIs; they are not substitutes for application evidence.
- Update dependency/import/geometry guards only for the exact reviewed migration. Retain deterministic multi-violation reporting, independent package/lockfile validation, negative controls, wrapper drift/status checks and full root/disabled/rename/CI regressions.
- Source adoption does not fix the kit's clipped focus or low-contrast labels automatically. Required accessibility corrections and original SVG/font provenance remain explicit G048 obligations.
- No vendor account, infrastructure service, production deployment or hosted configuration is introduced by selecting Radix. The G021 foundation barrier remains closed.

## Alternatives

- Retain Base UI and port selected source components: original G048 r1; superseded by the owner's explicit choice.
- Install the whole Radix/Vestive catalogue or run both primitive libraries indefinitely: outside bounded G048 scope.

## Related Artifacts

- G048 Goal (not included)
- G048 Work Spec (not included)
- Original Goal r1 (not included)
- Original Work Spec r1 (not included)

This accepted architecture decision authorizes the planning direction, not application implementation, dependency installation, commit, push or hosted actions.

## Owner amendment — Base UI/Rhea successor, 2026-10-25

The owner directs: “we should migrate and use the new Vestive brand as is plus
any additional useful carry over from the old skill. This can be a new
overhaul”, followed by explicit clarification that this means complete Radix
removal. The supplied Vestive Brand v0.2.1 kit was generated with shadcn/create
preset `b1Ha5oWoa` (Base UI, Rhea, Taupe, Geist, Lucide).

1. G067 (not included) adopts this
   source as the successor generic UI foundation. Future generic interactive
   primitives use Base UI after cutover. This supersedes the Radix target and
   Base UI removal requirements of the original decision for this successor.
2. Migrate every affected production/test/fixture consumer and active primitive
   configuration, then remove Radix imports and direct/transitive packages from
   the resolved application dependency tree. Do not retain a parallel Radix
   foundation or a Radix implementation restyled to imitate the new kit.
   Claude's reviewed Plan pins source mappings, dependency versions and actual
   lockfile/consumer changes; no floating install is selected by this ADR.
3. Adopt the new component/CSS/token design as supplied, making the adaptations
   needed for generic product identity, existing public branding inputs,
   Light/Dark/System with System default, scoped portals from first insertion,
   pre-paint/hydration, keyboard/focus/modal behavior, rendered contrast, reflow,
   reduced motion and locally served reviewed fonts/assets. Existing legacy
   visual shape/component APIs may change; functional/security boundaries stay.
4. Copy the new skill into the repository, switch active discovery/governance
   to it and remove the old repository skill/alias after the successful cutover.
   Carry forward useful old integration/accessibility and provenance knowledge.
   Preserve original sources and G002/G047/G048 acceptance through Git/history;
   the accepted G048 migration remains historically correct.

This is the owner's library/architecture approval for G067, without another
routine owner decision. Codex Plan review precedes implementation; source review,
applicable hosted checks, affected qualification and final acceptance still
apply. G067 proceeds alongside the owner's frozen preview and precedes G057;
the G021 foundation barrier remains.
Current Spec (not included).

## G069 owner amendment — unaltered kit adoption, 2026-10-26

The owner now selects the complete unaltered Vestive Brand v0.2.1 `kit/src`
for SaaStacular and explicitly accepts cmdk's necessary transitive dependencies.
G069 Spec (not included) governs
this target; implementation awaits its reviewed Plan and execution release.
G067/G068 are accepted history and remain the installed behavior until cutover.

1. Copy all 69 pinned source files byte-for-byte, preserving names, exports,
   relative paths, utilities, hooks and stylesheet chain. Adapt app consumers,
   aliases and configuration, not kit source. Importing unused components and
   their real build dependencies is intentional; no pages/features are implied.
2. Base UI remains primary. cmdk's required transitive Radix packages are permitted;
   the earlier complete-resolved-Radix-removal requirement is superseded for that
   closure only. No direct application Radix imports or unrelated primitive change.
3. Use the shipped standard theme selectors and CSS-variable conventions. App-owned
   tenant configuration and appearance initialization may replace data-appearance,
   the old cn utility, token plumbing and scope-container mechanism. Keep System
   default, Light/Dark choices, correct initial paint/hydration and tenant isolation.
   Body-level portals are permitted when correctly themed from first insertion;
   independent branding previews must remain isolated. Actual guarantees replace
   the former requirement for a specific portal DOM ancestor.
4. Keep keyboard/focus/modality, accessible rendered contrast, reduced motion,
   form semantics and server-side authority. App-owned supported theme variables
   can address contrast without editing source. Explicitly review saved-branding
   mapping and bounded Core alignment; no silent reinterpretation or data loss.
5. This supersedes conflicting selective-import, source-adaptation, dependency and
   visual-mechanism instructions in this ADR, Architecture §§37A/38/61, WEB-SHELL
   and local skill ADOPTION for G069's target only. Generic identity, security,
   accessibility and provenance requirements remain. Preserve upstream bytes and
   historical decisions; the implementer updates installed-behavior docs at cutover.

### Later G069 owner decision — exact visuals and contrast waiver, 2026-10-26

The owner explicitly requires the exact shipped kit visuals and waives inherited
contrast shortfalls attributable to that kit treatment/defaults. Remove app-owned
platform color corrections and consumer restyling introduced for contrast; keep
kit files unchanged and use the shipped compound Select for all dropdowns.
The controlling Spec amendment (not included)
supersedes conflicting rendered-contrast/correction requirements in this ADR and
earlier architecture guidance for G069. Record waived ratios honestly; functional
keyboard/label/form, appearance/tenant isolation and server-side authority checks
remain required. Existing tenant-branding functionality remains; the platform
default uses exact kit values. This does not authorize arbitrary tenant-palette
validation removal or persisted-data changes. Prior decisions/evidence remain
historical, and the amended implementation still requires source/CI review.
