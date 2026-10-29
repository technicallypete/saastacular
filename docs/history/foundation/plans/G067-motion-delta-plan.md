> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G067 Plan delta: subtle field and overlay motion (owner preview feedback, 2026-10-25)

Goal: G067 (Vestive Brand successor: Base UI/Rhea overhaul). Base Plan: `management/v0/plans/G067-vestive-brand-successor-plan.md`
(accepted r1). Integration base: `g067-int` `be6467efd59b478100e5b9fedbbc804628aaa0d0`. Writer: W5 (isolated, stack-free; the
coordinator runs every build and test). Status: submitted for supervisor review; no outcome is claimed here.

## 1. Owner direction

The owner previewed G067: everything looks good except that some form fields and overlays lack the subtle shadcn/Base UI
style animations, to be fixed inside G067 (relayed by the supervisor, 2026-10-25). This **supersedes Plan §5.3 D-6 ("no dialog
open/close motion", S-MOTION unchanged) for subtle interaction and open/close motion only**. Every other accepted decision
stands: D-4 no backdrop blur, D-5 blueprint backdrop, the solid 2px `ring` focus outline (offset 2px, at least 3:1), the
solid invalid border, every contrast oracle, and the global reduced-motion rule (`web/app/globals.css`; Plan §5.4 row "§9.11 reduced motion").

## 2. Inventory

| Behaviour | Before | Change |
|---|---|---|
| Dialog Backdrop open/close | none (D-6) | the kit's `data-open:animate-in data-closed:animate-out data-closed:fade-out-0 data-open:fade-in-0 duration-100`; `bg-overlay` kept, no blur |
| Dialog Popup open/close | none (D-6) | the kit's `data-open:animate-in data-closed:animate-out data-closed:fade-out-0 data-open:fade-in-0 data-closed:zoom-out-95 data-open:zoom-in-95` with the kit's `duration-100`; centring/translate, scroll region, ring and shadow kept |
| Field focus indicator (Input, Textarea, Select) | outline not in the transition list: appears/disappears instantly | resting `outline-0 outline-transparent`; transition list adds `outline-color`, `outline-offset`, `outline-width`; same `duration-200`; focused state unchanged (`focus-visible:outline-2 outline-offset-2 outline-ring outline-solid`) |

Explicitly **not duplicated**: the fields' existing `transition-[color,box-shadow] duration-200` (kept, extended only by the
outline properties). The resting outline is zero-width, so it is never a focus indicator at rest and draws nothing in forced
colours. Not changed (reported only): `button.tsx` (`transition-all` already animates the outline), TextLink, nav links and
the site header (`transition-colors` in Tailwind 4 already includes `outline-color`), the table scroll region and the
appearance control (no transition; not fields or overlays).

## 3. Tests

- `overlay.spec.ts` **[S-MOTION] (replaced assertions):** without reduced motion the popup and backdrop have an animation
  (name not `none`, duration > 0 and at most 0.15s); with `reducedMotion: "reduce"` every popup/backdrop animation and
  transition duration and the button-transition control are at most 0.01s; the button-transition control (> 0.05s without
  reduced motion) is kept. The close path (Escape, `toBeHidden`) waits for the exit animation before Base UI unmounts.
- `tokens.spec.ts` **[S-FIELD-MOTION] (new, one scope):** Input, Textarea and Select in the default scope: transition-property
  includes `outline-color`, `outline-offset`, `outline-width` (plus `color`, `box-shadow`); duration 0.2s, at most 0.01s under
  reduced motion; no outline at rest (width 0px); after keyboard focus and settling, solid 2px offset 2px in the ring colour.
  Its contrast stays with S-CONTRAST/S-STATES. Fixture: a Textarea and a Select added to the default scope only
  (`web/tests/fixtures/token-scopes/page.tsx`). Like S-FIELD-INVALID it is not added to the harness matrix (harness untouched);
  `browser_matrix` already requires every collected test to pass.
- **Mid-animation reads:** new helper `dialogSettled(page)` (`support/contrast.ts`) awaits the backdrop's and popup's
  animations. Added where geometry, computed style, axe or a placed negative cover follows opening: `overlay` S-OVERLAY (Close
  box), `shell` S-RESPONSIVE/S-LONG-IDENTITY (`expectDialogFits`), `accessibility` S-CONTRAST (dialog rows), `design`
  `openDialog` (S-DESIGN, S-PALETTE, S-FIDELITY, S-FONTS swap, S-SCREENSHOTS) and the S-FONTS control-6 320px check,
  `tokens` S-PORTAL, S-A11Y, S-CONTRAST, S-TEXT-SPACING. Already settled: `accessibility` S-A11Y and S-TEXT-SPACING
  (document animations), `support/spacing.ts` `inspectSpacing`, `tokens` S-SCREENSHOTS (opacity 1 + `animations: "disabled"`).
- Source-reference and S-FIDELITY rows compare computed geometry/type/colour, not class strings: no edit.

## 4. Reduced motion

The global rule forces `animation-duration` and `transition-duration` to 0.01ms on every element; it covers the
`tw-animate-css` classes (the `animation` shorthand is overridden by `!important`) and the fields' transition. Asserted by
S-MOTION and S-FIELD-MOTION.

## 5. Files

`web/components/ui/{dialog,input,textarea,select}.tsx`; `web/tests/browser/{overlay,shell,accessibility,design,tokens}.spec.ts`;
`web/tests/browser/support/contrast.ts`; `web/tests/fixtures/token-scopes/page.tsx`; `docs/foundation/WEB-SHELL.md`;
this delta; the G067 Result §12. `web/THIRD_PARTY_NOTICES.md` lists no dialog difference: unchanged.

## 6. Focused selectors (coordinator)

- Static: `lint lint-web`, `unit-web`.
- `overlay.spec.ts` whole file, projects `shell-light`, `shell-dark`.
- `shell.spec.ts --grep "S-RESPONSIVE|S-LONG-IDENTITY"`, `accessibility.spec.ts --grep "S-A11Y|S-CONTRAST|S-TEXT-SPACING"`,
  `design.spec.ts --grep "S-DESIGN|S-PALETTE|S-FIDELITY|S-FONTS|S-SCREENSHOTS"`, projects `shell-light`, `shell-dark`.
- `tokens.spec.ts --grep "S-FIELD-MOTION|S-FIELD-INVALID|S-PORTAL|S-A11Y|S-CONTRAST|S-TEXT-SPACING|S-SCREENSHOTS|S-STATES|S-TOKENS"`,
  projects `tokens-light`, `tokens-dark` (S-PORTAL also selects S-PORTAL-INIT; S-STATES/S-TOKENS cover the fixture change).

## 7. Scope

This fits accepted G067 scope: the change re-adopts the kit's own Dialog motion classes and adds the outline to the fields'
existing transition. No new dependency (`tw-animate-css` 1.4.0 is already pinned and imported), no token, contract, API or
Core change, no harness or CI change, and no change to the pinned skill bytes (`.agents/skills/vestive-brand/**`,
`.claude/skills/**`). Focus, invalid-border and contrast requirements are unchanged in their settled state.
