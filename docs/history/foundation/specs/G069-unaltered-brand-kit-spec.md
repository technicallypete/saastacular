> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G069 Work Spec — Complete, unaltered Vestive Brand source adoption

Revision: r0, 2026-10-26 UTC. Owner: Codex. Status: ready for Claude planning.
Current controlling revision: the [owner amendment below](#8-owner-amendment--exact-kit-visuals-and-contrast-waiver-2026-10-26-utc)
supersedes conflicting visual/contrast requirements in r0 and its accepted Plan.
[Goal](../goals/G069-unaltered-brand-kit.md) / [ADR-001 amendment](../../../adr/ADR-001-radix-vestive-components.md#g069-owner-amendment--unaltered-kit-adoption-2026-10-26).

## 1. Owner outcome and boundaries

Adapt SaaStacular to the shipped kit, rather than adapting the kit to the existing
application. Import 100% of its source, preserving component names and behavior,
utilities, hooks and CSS, even where no current page uses a component. The owner
accepts the required cmdk dependency closure. This supersedes G067's selective,
adapted adoption and the blanket ban on transitive Radix only as specified below.
G067/G068 acceptance and failure history remain historical, not rewritten.

This authorizes generic UI integration only. Keep current routes, functionality,
native form semantics, authorization, CSRF, tenant/Group boundaries, audit and
server-side validation. No new command palette, charts or marketing pages are
required merely because their source is now available. Protect M1 and G057's
frozen checks; no live environment action is included.

## 2. Immutable source and destination

Source is the repository `.agents/skills/vestive-brand/kit/src` at staging
`57d40b8eb8130eb09619d13f498d62393a877d3a`, Vestive Brand v0.2.1:

| Binding | Value |
|---|---|
| Regular files / bytes | 69 / 241557 |
| Git source subtree | `b1be1af48bf92ae84a6ce077b0042592ca0b2700` |
| Source-manifest SHA-256 | `597f2aac251987be26050108dc6affe3a39a5594e7bfc7e2744de2fb5829a6ee` |

Digest algorithm: sort regular files by relative POSIX path; concatenate UTF-8
lines `<sha256(file bytes)>  <relative path>\n`; hash those manifest bytes.
Verify individual files against the existing upstream manifest as well. This
69-file digest is not the whole skill's 319-file digest. No home-directory update,
remote artifact fetch, generator output or second design-system source is a live
replacement for the pinned repository source.

Copy every source file into one application source root (prefer `web/src`), keeping
all relative paths, bytes, names and exports unchanged. Include `globals.css`, all
`styles/`, `lib/utils.ts`, `hooks/use-mobile.ts`, UI and Vestive components. Configure
application aliases to resolve the source's canonical imports; avoid relocating
unrelated application code just to accomplish this. Compile/typecheck every
copied TS/TSX file, including unused files, with real dependencies. No source
patches, stubs, ts-ignore, disabled type checking or catalogue exclusions.

Production consumers must actually use this source and its CSS. A pristine unused
copy next to the old component implementations does not satisfy this Spec. Remove
superseded implementations/utilities/styles after consumer migration. Keep
application-specific composition outside the copied tree, with honest names and
purpose; no shadow components or global CSS recreation of the old kit. Do not
rename NativeSelect to disguise it as the kit's compound Select, for example.

## 3. Dependencies and application integration

Inventory every source import and required peer/CSS dependency. Pin compatible
versions and reconcile the application lockfile and any governed archetype
manifest/template dependency representation under existing provenance rules.
Required candidates include cmdk, clsx, tailwind-merge, next-themes, sonner,
react-day-picker, embla-carousel-react, input-otp, react-resizable-panels, recharts,
and the shadcn CSS import, in addition to existing React/Base UI/Tailwind packages.
The Plan determines exact versions and actual dependency closure; no floating
latest, unused speculative extras or blanket security waiver.

Base UI remains the application's primitive foundation. The owner's explicit
cmdk exception permits the Radix packages required transitively by that dependency;
no new direct application Radix imports or parallel Radix component migration.
Record why each resolved Radix package exists and remove obsolete direct Radix
packages. Do not run shadcn init/add to rewrite the pinned source. An installed
package needed for the shipped CSS is different from generator execution.

Migrate actual consumers to the kit's exports and prop contracts, including
buttons/links, dialogs, inputs, textarea, radio groups, selects and typography.
Use NativeSelect where native submission is intended, or correctly bridge the
compound control's documented form API. Preserve names, values, defaults,
disabled behavior, errors, keyboard use and submission semantics. Use kit CtaButton
where appropriate rather than retaining an invented shared button variant.
No product identity or Vestive marketing defaults may leak into rendered app copy.
The exact copied source can receive a narrowly scoped literal-guard allowance;
application-owned files and runtime identity remain governed. Copy referenced
assets only when required and permitted by the product/provenance rules; no
automatic shipment of Vestive logos as SaaStacular's identity or remote font loader.

## 4. Standard themes and tenant branding

Consume the shipped stylesheet chain unchanged. Adapt app initialization and
theme selection to its class/data-theme and semantic CSS-variable conventions,
using one authoritative appearance state (including next-themes where required
by the source). Preserve Light, Dark and System, default System, stable preference,
server/client hydration and correct first paint. Do not operate two competing
theme managers or retain legacy selectors merely to satisfy obsolete tests.

Keep tenant customization as application-owned variable values/configuration
outside the copied source. Preserve authorized branding updates, independent
Organizations, tenant switching, reset/default behavior and isolated previews.
Map existing saved branding values explicitly into the new roles; the old gold
accent and the kit's neutral accent surface are not silently interchangeable.
Bounded Core branding defaults/validation alignment is in scope; any persisted
contract migration or incompatible interpretation requires a specific Plan
proposal and reviewed amendment before execution. Do not assume data is disposable.

Portals must receive the correct active tenant/appearance variables from their
first visible insertion, with no other tenant's values leaking into overlays,
toasts or previews. Some shipped components create body portals internally and
do not expose the old container hook. Therefore the old mandatory scope-container
DOM placement is superseded by the isolation/first-paint guarantee, not preserved
as an impossible implementation constraint. Plan a compatible app theme host
(e.g. active tenant variables at the document root, separately isolated preview
surfaces), including simultaneous preview and tenant-switch handling. No internal
kit portal patch or delayed post-render repair.

Retain real accessibility: keyboard, labels, focus visibility/return, modality,
inertness, contrast, spacing/reflow and reduced motion. Evaluate whole rendered
states (including solid focus borders), not an isolated translucent ring. Theme
variable overrides needed for accessibility may be app-owned and documented;
no hidden source mutation or wide CSS override to restore the old visuals. If
source fidelity and an actual requirement cannot both hold, report the exact
case for a bounded decision before widespread migration. Neither copying source
nor this Spec waives accessibility/security requirements.

## 5. Planning and execution sequence

Claude prepares one Plan, with parallel research on dependency/build compatibility,
themes/portals/branding, and consumers/validation. Bind exact input hashes and
writable paths. Identify any concrete upstream incompatibilities together; distinguish
unused cosmetic defects from build/security/used-component blockers. Do not fix
upstream source silently. Give coding and validation estimates separately.

After Plan ACCEPT and execution release, first prove the full catalogue resolves
and compiles, plus a representative themed tenant overlay/form with the proposed
app integration. Then run dependency-ready consumer migrations in parallel under
named nonconflicting ownership. Shared package/lockfile, theme/alias and final
integration edits each have one owner. Preserve the existing private test/build
reservations and continuous cleanup; planning alone launches no installs/builds.

## 6. Validation and acceptance

Use the existing checks and focused browser suites. The Plan must supply executable
selectors and coverage rationale for the actual final changes and affected security
boundaries. Update stale source-text/token/geometry oracles only where the approved
contract changes; retain functional/security negative tests. Catalogue registration,
allowlists or count pins alone do not require full ci-controls.

| AC | Required outcome and evidence |
|---|---|
| AC1 — Source fidelity | All 69 files present at the declared app source root, exact per-file and aggregate comparison to the bound kit; upstream skill unchanged; no competing old runtime implementations. |
| AC2 — Complete build/dependencies | All copied TS/TSX and CSS imports resolve under real pinned dependencies; application typecheck and production build pass; dependency/security/provenance checks pass and cmdk's transitive Radix exception is explicit. |
| AC3 — Working consumers | Actual used routes/forms consume canonical kit exports; affected unit/browser checks cover submit/default/error/disabled semantics, dialog/select/radio interaction and existing authorization/CSRF/tenant boundaries. No new feature required for unused components. |
| AC4 — Appearance and branding | Light/Dark/System and initial paint/hydration, tenant overrides/reset/switching, overlay styling from first insertion and isolated previews pass representative focused checks; saved-branding mapping and Core alignment are explicit. |
| AC5 — Rendered design/accessibility | Used input, textarea, radio/label, button and overlay examples follow kit treatment; keyboard, focus, contrast, reflow and reduced-motion checks pass in light/dark and representative tenant overrides. Explain any app-owned variable correction. |
| AC6 — Integration and closeout | Current hosted fast and applicable source/Duo disposition, authorized staging merge, attributable affected exact-merged-SHA checks, Result AC mapping and all-category cleanup receipt; preserved failures and honest known limitations. |

Full-source compilation does not require a new behavioral campaign for every
unused catalogue component. Identify unused upstream defects honestly in the
Result and bounded follow-ups; unused cosmetic defects alone do not block adoption.
A build/security issue or defect in a used component needs an explicit disposition.

Default qualification is focused pre-merge evidence, hosted fast and review, then
affected local suites on the exact merged SHA. No blanket local full suite or full
CI-controls run. Broader staging/full RC qualification remains available under
standing policy with a concrete reason or owner request; do not claim omitted
checks passed. Reuse unchanged attributable evidence after corrections. No new
harness, driver framework or mandatory screenshot-review gate is introduced.

## 7. History and delivery

- 2026-10-26 r0: owner authorizes this source-first migration Spec and parallel
  planning, including cmdk's required transitive dependencies. Implementation is
  not yet released. G069 and remaining G057 work are independent lanes; both
  acceptance outcomes precede G021. RC15 promotion and M1 remain unchanged.

## 8. Owner amendment — exact kit visuals and contrast waiver, 2026-10-26 UTC

The owner directly instructs: “Omg waive the contrast issues... I want it EXACTLY
like the brand kit - no compromises”. This is explicit authority to prioritize
the pinned kit's exact visual treatment over the inherited contrast minimums.
The supervisor's preceding contrast hold is lifted; no further confirmation or
new Goal/replan is required for this bounded amendment. See the
decision and supersession record (not included).

Use the shipped components and their default styles, colors, spacing, borders,
focus treatment and overlays. Remove app-owned platform token corrections for
`input`, `muted-foreground`, `destructive` and `vx-blue`, and consumer styling or
replacement close controls introduced to change kit appearance for contrast.
Do not substitute another visual correction. All dropdowns, including Appearance,
use the shipped compound `Select`; this supersedes r0's NativeSelect option and
Plan S5/S12. Use the matching shipped Label/Field/FieldSet/Alert/Card parts for
existing consumers where their semantics match. All 69 kit source files remain
byte-identical to the bound source.

For AC5 and related Plan checks, contrast shortfalls attributable to the shipped
kit treatment/defaults are owner-waived and do not block G069. Record affected
ratios, states and outcomes as waived, retaining earlier failures and evidence;
never claim that a waived ratio meets its former threshold. Update obsolete
visual assertions to the exact kit contract without suppressing unrelated
functional, keyboard, labeling, focus-operation, form, security or isolation checks.
This supersedes conflicting contrast correction/stop requirements in §4, AC5,
the Goal, accepted Plan and earlier G069 reviews only for this visual decision.

Preserve the existing authorized tenant-branding feature and its variable mapping.
The platform fallback/default must reflect the exact shipped kit values; no
palette correction is allowed merely to make an old validator pass. Reconcile
necessary Core default/test expectations explicitly under the existing bounded
alignment scope. This waiver does not authorize global removal of tenant-provided
palette validation, persisted-data changes or unrelated visual redesign.

Supervisor clarification within this alignment scope: for each individual
contrast pair/state and appearance, the save validator requires at least the
lesser of its existing threshold and that same pair's ratio on pinned kit
defaults. This admits unchanged/default palettes while retaining the original
threshold wherever the kit meets it and refusing tenant palettes below the
applicable kit floor. Keep all input/security validation and no data mutation.
Record baseline shortfalls as owner-waived. See the
Core validator disposition (not included).

The owner's full hosted `validate:integration` pre-merge run, plus normal hosted
fast and applicable review, qualifies the final executable candidate. It replaces
the full post-merge run when tested and merged executable trees match. Retain
already-required focused post-merge checks; no local full run is requested.
Changed source still needs delta review and attributable validation. Existing
attempts, worker ownership and failure history continue without reset.

Later sequence clarification, verified from the coordinator's committed
owner instruction record (not included):
after the existing writers finish, rebuild the retained M1 preview from their
combined candidate for the owner's manual check before starting the subsequent
focused test phase. Resume that test phase after the owner's check permits it;
do not infer approval from elapsed time. This is a testing-sequence hold, not a
renewed contrast hold or waiver question. The full hosted pre-merge gate follows
the resulting corrections and reviews under the qualification policy above.
