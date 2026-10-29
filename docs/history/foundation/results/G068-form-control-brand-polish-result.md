> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G068 — Form control brand treatment and motion polish: Result

Current status, supervisor import 2026-10-26 UTC: **FINAL ACCEPT** at staging
`8f4e930abf8776c3251bbc11fb9d26f4e8b43473` (MR !103). Sections below retain
historical writer submissions and proposed/pending checks; §7 is the current
qualification and closeout disposition. [Final review](G068-final-acceptance-codex-review-20261026.md).

Status: **writer W8 source submitted, not validated.** Direct presentation tier (Goal; no Plan). No Docker, build, lint,
type check or browser run was executed by the writer (stack-free); every outcome below is pending the coordinator's
batches. Not Codex-accepted.

Writer: W8, Opus 5.5 (effort inherited, not verified), isolated agent worktree
`.claude/worktrees/agent-<id>`, branch `g068-w8` from staging `af5dec51325c5619e149e059ac740e9f8aa1c44d`
(verified equal to `origin/staging` after fetch). Kit input: repository `.agents/skills/vestive-brand` (read-only;
`.claude/skills/vestive-brand -> ../../.agents/skills/vestive-brand` verified in this checkout).

## 1. Inventory (before implementation)

Source references (kit, read-only): `kit/src/components/ui/radio-group.tsx` (RadioGroupItem: `size-4 rounded-2xl`,
`bg-input/90 border-transparent`, checked `bg-primary` + `primary-foreground` dot `size-2`, Dark `size-2.5`, no
transition), `kit/src/components/ui/label.tsx` (`text-sm font-medium`, `gap-2`), `kit/src/components/ui/field.tsx`
(FieldSet radio group `gap-3`, FieldLegend `mb-3 text-base font-medium`, FieldLabel `leading-snug`, horizontal Field
radio `mt-px` on wrapped content), `kit/src/components/ui/{input,textarea,native-select}.tsx` (`transition-[color,box-shadow]
duration-200`), RadioGroup preview (`flex items-center gap-2` + Label). Retained G067 adaptations: ADOPTION §6 (field
fill `muted` + 3:1 `border-input` boundary instead of the kit's translucent fill; solid 2px `ring` outline offset 2),
G067 Result §12 (fields' outline motion, reduced-motion rule), WEB-SHELL focus-adaptation paragraph.

| # | Control (path) | Actual gap | Smallest correction |
|---|---|---|---|
| F1 | Group label visibility radios (`web/components/labels/label-form.tsx` `VisibilityChoice`, also used by `label-list.tsx` row forms) | Platform-drawn native radio (13px, UA margins, browser accent), no kit treatment; option text inherits 16px body; radio top-aligned with `items-start` and UA margin, not centred on the first line; legend has no gap to the first option (a rendered legend is outside the grid, so `gap-3.5` never applies to it) | Shared `RadioOption`; legend `mb-3`, fieldset `gap-3` |
| F2 | Organization enrollment policy radios (`web/components/enrollment/policy-form.tsx`) | Same as F1; the disabled `open` option has no treatment | Same; disabled = circle 50 %, text `muted-foreground` |
| F3 | Group authority radios (`web/components/membership/group-authority-form.tsx`) | Same as F1; long options wrap beside an unaligned radio | Same |
| F4 | Onboarding profile picker (`web/app/groups/[id]/onboarding/page.tsx`) | Native radio `size-5 accent-primary mt-1` (20px, platform-drawn); `gap-4` differs from the other three | `RadioOption`; `gap-3` (bordered fieldset/legend unchanged) |
| F5 | Branding text fields (`web/components/branding/branding-form.tsx` `FIELD`: workspace name and 30 colour inputs) | Native inputs copying Input's geometry without its motion: focus outline snaps (no transition, no resting outline) | Add Input's transition and zero-width transparent resting outline to `FIELD` |
| F6 | Logo file input (`web/components/branding/logo-forms.tsx`) | Focus outline snaps (no transition) | Same motion classes on the input |

Compliant, unchanged: `components/ui/input.tsx`, `textarea.tsx`, `select.tsx` (G067 motion delta present) and all their
consumers (`form-field.tsx`, `accept-form.tsx`, `invitation-form.tsx`, `join-form.tsx`, `label-form.tsx` NameInput,
`group-labels-section.tsx`, `ownership-forms.tsx`, `member-table.tsx`, `connection-manager.tsx`: no className motion
override; the only consumer className is FormField's `min-h-44`); no `transition-none` anywhere in `web/{app,components,lib}`;
Button (`transition-all`), links (`transition-colors`), dialog (G067) unchanged.

Reported, not changed: the header appearance control (`components/layout/appearance-control.tsx`, native select inside a
label with hover fill and `has-[:focus-visible]` outline, no transition). It is a shell control, not an app form field;
G067 §12 accepted it unchanged; S-CONTRAST/S-DESIGN sample its hover and focus pixels without a settle wait, so a
transition would need those tests changed and the shell matrix rerun. Candidate follow-up only if the owner names it.

Constraints found: B-WEB-COMPONENTS fixes `components/ui` to eleven files, the `data-slot` set and the Base UI subpaths
(button, dialog, input), so the kit RadioGroup (Base UI radio) cannot be adopted without a harness change and a
primitive-behaviour change: out of scope. The wrapper therefore lives at `components/account/radio-option.tsx` (beside
the shared `PendingSubmit`/`StatusMessage`) and emits no `data-slot`.

## 2. Implementation (per AC)

**AC1 — radio treatment and label layout.** New `web/components/account/radio-option.tsx` (`RadioOption`): a `<label>`
wrapping a native `<input type="radio">` and the option text. Circle: `size-4 rounded-full appearance-none border
border-input bg-muted`; checked `bg-primary-foreground` with `inset-ring-3 inset-ring-primary` inside the same `border-input` boundary (correction C2)
(8px dot; Dark `inset-ring-2`, 10px, as the kit); disabled `opacity-50 cursor-not-allowed`;
`forced-colors:appearance-auto`. Text: `text-sm font-medium leading-snug` (kit Label size/weight, FieldLabel leading),
`[overflow-wrap:anywhere]`, `gap-3`. The circle sits in a `h-[1lh]` flex box, so it is centred on the first line box at
any text size or line spacing and wrapped lines continue beside the text column. All props (name, value,
defaultChecked, disabled, required, data-testid) pass to the native input. The four consumers use it; legends `mb-3`,
fieldsets `gap-3` (onboarding's bordered legend unchanged). Native fieldset/legend grouping, defaults, disabled rule,
names/values, CSRF hidden fields, native POST and arrow keys are untouched.

**AC2 — motion.** Changed interactions: (a) keyboard focus on each radio option: outline colour/offset/width ease over
200ms (resting outline zero-width transparent; settled outline unchanged: 2px solid `ring`, offset 2); (b) keyboard
focus on the branding workspace-name and colour fields; (c) keyboard focus on the logo file input. All use the existing
global reduced-motion rule (0.01ms). The checked/unchecked change is not animated (the kit radio has none). No
`transition-all`, layout animation or overlay change.

**AC3 — preserved behaviour.** No change to form markup semantics beyond the wrapper elements; existing scenario
assertions (check(), toBeChecked, keyboard arrow changes, disabled `open`, forged-POST refusals, axe 1280/320) are kept
and now run over the new markup. Contrast: resting boundary `input` on white/muted 3.71/3.30 (Light), 4.23/3.27 (Dark);
checked dot on disc 5.65; disabled text `muted-foreground` 5.28-7.75. Original finding (W8, before correction C2): the
checked disc's edge was the kit's `primary` fill, 2.70:1 against the Dark `muted` band of the onboarding picker (3.50 on
the Dark ground, 5.02-5.65 in Light). Codex's early conformance disposition required the retained 3:1 control-boundary
guarantee; correction C2 keeps the `input` boundary in the checked state (no token, palette or kit change) and the
helper now measures both states' boundary against the real surface.

Docs: `docs/foundation/WEB-SHELL.md` (one G068 paragraph after the focus-adaptation paragraph).

Tests (extended, no new suite): new helper module `web/tests/browser/support/form-controls.ts`
(`expectRadioOptions`, `expectFocusMotion`), called from one existing scenario per affected form:

| Spec / scenario | Call | Covers |
|---|---|---|
| `membership.spec.ts` [E-MEM-8] | `expectRadioOptions(..., "Group administration", true)` after the owner audit | treatment, alignment, label association, boundary/dot contrast, keyboard focus outline + ring contrast, focus motion with reduced motion, fieldset capture, and at 320px a wrapped option still aligned plus a 320px capture |
| `enrollment.spec.ts` [E-ENR-5] | `expectRadioOptions(..., "Who can join")` after the owner audit | as above at 1280, including the disabled `open` option |
| `labels.spec.ts` [E-LAB-4] | `expectRadioOptions(..., "Label visibility")` on the row visibility form | as above at 1280 (table cell) |
| `onboarding.spec.ts` [E-ONB-2] | `expectRadioOptions(..., "My Business Profiles")` after `check()` | as above at 1280 (muted band) |
| `identity.spec.ts` [E-BRD-5] (@domain-platform) | `expectFocusMotion` on `#display_label`, `#light-primary`, `#logo-file` | F5/F6 transition properties, 0.2s, reduced ≤ 0.01s, 0px resting outline |

Reduced motion uses `page.emulateMedia({ reducedMotion: "reduce" })` then `null`, as S-FIELD-MOTION does. Light/Dark come
from the existing `*-light`/`*-dark` projects; expected colours are read from the page's tokens at the control.

## 3. Proposed focused checks (coordinator)

1. Static: harness `lint lint-web` (B-WEB-TOKENS, B-WEB-COMPONENTS, B-WEB-DESIGN, eslint, `tsc --noEmit`), allow-file
   covering the 14 paths in §4. `unit-web` is not affected (no `lib/` or route change; unit tests cover routes only);
   run it only if the coordinator's selection requires it.
2. Identity stack (`IDENTITY_E2E=1`, one private run):
   `npx playwright test tests/browser/membership.spec.ts --grep "E-MEM-8" --project=membership-light --project=membership-dark`;
   `tests/browser/enrollment.spec.ts --grep "E-ENR-5" --project=enrollment-light --project=enrollment-dark`;
   `tests/browser/labels.spec.ts --grep "E-LAB-4" --project=labels-light --project=labels-dark`;
   `tests/browser/onboarding.spec.ts --grep "E-ONB-2|E-ONB-5" --project=onboarding-light --project=onboarding-dark`
   (E-ONB-5 for the keyboard/axe pass over the changed picker);
   `tests/browser/identity.spec.ts --grep "E-BRD-5" --project=domain-platform`.
   Rationale: every changed component is exercised; E-MEM-8/E-ENR-5 also carry the existing keyboard and axe
   1280/320 checks on the changed forms; the labels create form is the same `VisibilityChoice`. Broader: the rest of the
   four feature specs touch the same radios only through `check()`; run whole files only if the coordinator wants
   submission-path coverage beyond E-LAB-4/E-ONB-2/E-ENR-5/E-MEM-8.
3. Not selected: shell/tokens projects (no shared component in `components/ui`, `globals.css` or fixture changed),
   full matrix, `ci-controls`.

## 4. Files changed (writer commit)

`docs/foundation/WEB-SHELL.md`, `management/v0/results/G068-form-control-brand-polish-result.md` (new),
`web/app/groups/[id]/onboarding/page.tsx`, `web/components/account/radio-option.tsx` (new),
`web/components/branding/branding-form.tsx`, `web/components/branding/logo-forms.tsx`,
`web/components/enrollment/policy-form.tsx`, `web/components/labels/label-form.tsx`,
`web/components/membership/group-authority-form.tsx`, `web/tests/browser/support/form-controls.ts` (new),
`web/tests/browser/{enrollment,identity,labels,membership,onboarding}.spec.ts`.

No migration, schema, contract/API/MCP, route, form-parsing, dependency, lockfile, token/palette, theme-initialization,
pinned-kit, harness or CI change. Architecture deviations: none.

## 5. Unverified without a stack

Rendered geometry and pixels (alignment, dot size, legend spacing), Tailwind 4.3.3 generation of `inset-ring-3`,
`has-disabled:`, `h-[1lh]`, `forced-colors:appearance-auto` and the `dark:checked:` override order, eslint/tsc on the
new TypeScript, and the new assertions themselves. Attempt count: Goal attempt 1 (B3), writer round 1.

## 6. Coordinator outcomes

(to be appended)

## Correction record (coordinator, 2026-10-25; append-only, original outcomes kept)

- **C1 (test helper race).** Focused browser run 1 on `187c6d18` (evidence `.local/evidence/g068-a1/browser-1`):
  3 expected, 8 unexpected, 0 skipped/flaky. All 8 ([E-ONB-2], [E-MEM-8], [E-ENR-5], [E-LAB-4], light and dark) failed
  `expectFocusMotion` at `support/form-controls.ts:39`: outline width at rest expected `0px`, received `1px`. The
  helper blurred the control and read the width two frames later, during the 200ms transition back to no outline;
  [E-ONB-5] and [E-BRD-5] passed. Fix `8fa5395a`: poll the computed outline width until it settles at `0px`; the
  transitioned properties and durations are still asserted first. Attempt 2 unit-web PASS on `8fa5395a`.
- **C2 (checked boundary).** Codex's early conformance disposition (2026-10-25): the 2.70:1 checked edge on the Dark
  `muted` band had to be resolved under the 3:1 control-boundary guarantee, with the checked and resting boundaries
  measured against actual surfaces in the same helper; the earlier statement that this needed a palette change was too
  strong. Fix: the checked radio keeps the 1px `border-input` boundary (the kit's `primary` edge removed); dot, ring,
  focus and native semantics unchanged. Run outcomes are appended below.

## Run outcomes (coordinator, 2026-10-26; private local checks on frozen detached checkouts, evidence `.local/evidence/g068-a1/`)

Selectors (no full suite, no `ci-controls`, no M1 rebuild; base `af5dec51`, allow-list of the 15 paths plus this Result):
`lint lint-web` (static B-WEB-* checks and B-SCOPE), `unit-web` (eslint `--max-warnings=0`, `tsc --noEmit`, unit tests, secret
scan), and the focused browser run: the documented `identity:browser` line on the disposable HTTPS identity stack with the
nine projects `membership-*`, `enrollment-*`, `labels-*`, `onboarding-*` and `domain-platform` and `--grep "E-MEM-8|E-ENR-5|
E-LAB-4|E-ONB-2|E-ONB-5|E-BRD-5"` (generated from the G067 row 6 driver; `B-IDENTITY-IT` and `mcp_seed` run in its prefix).

| Run | Commit | Outcome |
|---|---|---|
| lint-web `static-1`, unit-web `unit-1` | `187c6d18` | both PASS |
| focused browser `browser-1` | `187c6d18` | **FAIL**: 3 expected, 8 unexpected, 0 skipped/flaky (C1: `expectFocusMotion` outline sampled mid-transition) |
| unit-web `unit-2`, lint-web `static-2`, focused browser `browser-2` | `8fa5395a` (C1) | all PASS (browser run: no unexpected, skipped or flaky rows) |
| unit-web `unit-3`, lint-web `static-3`, focused browser `browser-3` | `d7724621` (C1 + C2) | all PASS: 11 test rows, all expected, 0 skipped/flaky; `B-IDENTITY-IT` 95/95; `B-SCOPE` 0 changed paths since `d7724621`; cleanup receipt CLEAN (71 owned resources gone, 0 leftover/unknown); host counts 26 containers/14 volumes/9 networks/60 images before and after |

`d7724621` is the candidate. The oracle on it measures, for every enabled option in the four radio fieldsets, the boundary
(checked and resting, `input` token) against the pixel of the actual surface beside the circle at 3:1, the dot against
its disc at 3:1, the checked/resting fill and dot geometry, 16px circle centring on the first text line, label
association, keyboard focus outline and ring contrast, the 200ms focus motion with reduced motion, and at 320px a wrapped
option still aligned. Captures (10 PNG attachments, light and dark, including the 320px wrapped Group administration
labels) are exported to `.local/evidence/g068-a1/browser-3/shots/`; the onboarding picker (Dark) and the 320px
membership capture were viewed: the checked circle is bounded by the input boundary and the wrapped option's circle
stays on the first line.

Not run, by plan: the full browser/unit suites, `ci-controls`, hosted fast/Duo (after the MR opens), the post-merge
affected qualification (on the exact merged SHA), and an M1 rebuild. The header appearance control still has no
focus transition (unchanged since G067; reported by the writer). This Result commit adds only this section to
the tested `d7724621`.

## 7. Supervisor import, final acceptance and cleanup — 2026-10-26 UTC

**FINAL ACCEPT.** Codex reused the accepted source review and an independent
read-only final evidence review; no further implementation or test run was needed.
The reviewed executable source `d7724621`, rebased MR head `72b2732f` and merged
`8f4e930a` have identical executable inputs (`web/` tree `164e7b24`). The merge
has parents staging `31faead9` and source `72b2732f`; full merge tree is
`c021934b34c03490e0b4d63c3d72f64bac8cfd8d`.

Hosted fast pipeline `2987584368` succeeded; optional integration remained
manual/unexecuted. Duo approved with no unresolved resolvable discussion notes
(the latest independently checked API has five discussions; the earlier receipt
recorded three). The unchanged static-3/unit-3 passes on `d7724621` remain
attributable. `unit-web` did run; the earlier "full browser/unit suites" exclusion
means broader browser and other service/full-project selections, not that web
unit coverage was omitted.

The exact merged-SHA focused run (`browser-pm`) passed the same nine projects and
six selected scenario IDs: `E-MEM-8|E-ENR-5|E-LAB-4|E-ONB-2|E-ONB-5|E-BRD-5`.
It produced **11/11 expected browser cases**, zero skipped/flaky/unexpected/retries,
and **95/95 supporting identity integration tests**. The driver exited 0 after
284 seconds; selected scope, secret-canary, Docker-only and required-check rows
passed. JSON SHA-256:
`ba26e032c788cf997185a4e64f5b187fd6a79175ad1479d90d726f9e50f85e46`.
This is affected qualification, not a full-local or full-CI-controls PASS. The
original eight helper-race failures remain failed and preserved in the Result
and retained failure JSON/log; C1 and C2 outcomes remain attributable above.

| Criterion | Accepted evidence |
|---|---|
| AC1 radio/label treatment | Four consumers use the reviewed native presentation wrapper; light/dark and narrow wrapped-label captures reviewed, with first-line alignment assertions. |
| AC2 form motion | Radio focus, branding name/colour fields and logo input use the existing focus transition; selected motion and reduced-motion assertions pass. Header appearance control remains unchanged and excluded. |
| AC3 behavior/accessibility | Existing grouping/default/disabled/submitted values/CSRF/native POST retained; selected keyboard, focus, contrast, reflow and accessibility scenarios pass. Checked boundary uses the existing input token. |
| AC4 integration/cleanup | Source/merge equality, hosted fast/Duo and exact merged-SHA affected qualification verified; task resource/ref/worktree/scratch cleanup verified below. |

The coordinator's closeout receipt (not included) is imported
byte-for-byte (SHA-256 `d96bdb3d126321e3a29d84cf60f8bf8737b44eecb5d1ac96bdab227c55d38e4e`).
Codex independently compared all 71 final-run resource identities with the current
Docker inventory: **53 containers, 12 images, three volumes and three networks
absent**, zero matches/unknowns. G068 worktrees and local/remote task branches are
absent, including the exact native auto-branch; no G068 temporary handoff files
or `/tmp/g067g068f.*` scratch remain. Coordinator and M1 resources are retained
under their existing named ownership; unrelated held refs are outside this task.
The run's 26/14/9/60 container/volume/network/image-entry baseline was unchanged;
a later supervisor audit counted 58 unique image IDs (distinct from tagged image
entries), without deleting shared resources.

The receipt's proposed evidence hold is now resolved: 247 disposable/generated
or duplicate files removed from 279 files / 5,101,479 bytes. Retained proof and
the retention receipt total **33 files / 683,801 bytes**, including original
failure details, attributable final static/unit outcomes, focused run summaries,
merged-SHA JSON/cleanup/test tables and ten merged-SHA light/dark/narrow captures.
`.local/evidence/g068-a1/retention-20261026.tsv` inventories removed and retained
paths with hashes, owner and release condition (Codex/G068 coordinator; reassess
minimal proof at G021 closeout or earlier explicit disposition). Its SHA-256 is
`eebe9bcb746233ad72a320aa8bfe49f3c232743e1dcecf8a1ac3991b2c6ea382`.

Progress compaction assessed: superseded G068 status and the stale operating
snapshot are preserved in the dated archive (not included),
with current acceptance and gates retained in PROGRESS. No active input checkout
was rebound. RC14 is an independent immutable candidate at the same implementation
merge; its full hosted qualification continues, without a promotion instruction.
Preview reconciliation and G057's separate implementation/live-action readiness
holds remain; this acceptance does not release those actions.
