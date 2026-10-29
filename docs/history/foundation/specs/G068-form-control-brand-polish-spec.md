> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G068 Work Spec — Form control brand treatment and motion polish

Revision: r1, 2026-10-26 UTC
Goal: [G068](../goals/G068-form-control-brand-polish.md).
Authority: owner-approved isolated follow-up to the G067 preview findings.

## Requirements

**R1 — Audit actual form controls against the adopted source.** Read repository
`.agents/skills/vestive-brand/SKILL.md`, ADOPTION.md and relevant kit components.
Read architecture §37A/§61, SECURITY, DECISIONS, ADR-001's Base UI amendment and
WEB-SHELL; retain G067's accepted appearance, token and accessibility adaptations.
The source kit stays pinned and read-only. Claude's discovery alias is
`.claude/skills/vestive-brand -> ../../.agents/skills/vestive-brand`; the home
skill is not a live task input.

Initial concrete radio consumers found in the observed baseline:

- `web/components/labels/label-form.tsx` — Group label visibility.
- `web/components/enrollment/policy-form.tsx` — Organization enrollment policy.
- `web/components/membership/group-authority-form.tsx` — Group authority.
- `web/app/groups/[id]/onboarding/page.tsx` — Group onboarding.

These native inputs bypass the kit RadioGroup component. This is an inspection
lead, not a direction to replace primitive behavior. Record actual visual gaps,
source reference and smallest correction in the Result. Audit the existing
Input/Textarea/Select wrappers and affected consumers for remaining motion/style
overrides; avoid general UI work.

**R2 — Correct radio appearance and label layout.** Apply the source's Rhea radio
shape, sizing and checked/unchecked treatment using existing semantic tokens and
retained accessible boundary/focus adaptations. Align controls with the first
label line and use consistent label line height, spacing and clickable
association. Wrapped labels remain legible/aligned at narrow widths and enlarged
text. Preserve native grouping, default state, disabled rules, submitted names/
values and keyboard behavior. A small reusable presentation wrapper is allowed;
different primitive behavior requires scope review before that change.

**R3 — Fix confirmed form-motion gaps.** Retain accepted dialog/backdrop behavior
and field-focus transitions. For affected app fields that miss the intended
subtle transition, identify the cause and correct the shared style or consumer
override. Name each changed interaction in the Result. No indiscriminate
animations, animated layout, new overlay system, broad transition-all changes or
delayed field usability. Honor reduced motion and retain accessible settled
focus/invalid states in both appearances and supported token scopes.

**R4 — Boundaries.** Scope is app form presentation, relevant existing tests/
fixtures and WEB-SHELL adaptation notes. No Core, schema, authority/session,
route/API/MCP, form parsing, dependency, theme initialization, palette/token
contract, pinned kit, harness/CI or release changes. Preserve visibility and
authorization decisions, CSRF, native POST semantics, validation messages and
existing scenario guarantees.

## Acceptance and validation

| ID | Required outcome | Evidence |
|---|---|---|
| AC1 | All four radio consumers use the intended treatment and correct label line height/alignment. | Compact source comparison and representative light/dark app captures, including a wrapped label at narrow width. |
| AC2 | Inventoried form-motion gaps are corrected. | Named controls/interactions and focused browser assertions or visual observations; reduced-motion evidence for changed transitions. |
| AC3 | Selection/default/disabled state, submitted values, keyboard use, visible focus and contrast remain correct. | Reuse affected labels/enrollment/membership/onboarding cases and accessibility checks; extend only for a demonstrated gap. |
| AC4 | Reviewed code integrates with affected coverage and clean resources. | Exact source/merge references, hosted fast/Duo disposition, selected post-merge outcomes and compact cleanup receipt. |

Claude selects executable checks from actual changed controls: affected web
lint/type/unit checks and existing browser cases for these forms and shared
focus/motion. Cover light/dark, wrapped labels/reflow and reduced motion where
affected; reuse fixtures instead of duplicating suites. Record selectors,
rationale, exact commits and outcomes in the Result. A shared component path
alone does not justify the whole browser matrix or CI-controls. No full local
qualification is selected; retain current hosted fast policy.

Work on a separate G068 branch while G067 row 6 uses frozen merge `bb605d5b`.
Schedule Docker work through existing reservations. After G067 final ACCEPT,
follow normal staging integration and affected merged-SHA qualification, reusing
unchanged attributable passes. Do not rebuild M1 automatically. Preserve failed
runs; clean finished workers/resources at handbacks and closeout.
