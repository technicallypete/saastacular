import type { BrandingToken } from "@/lib/tenant-css";
import type { TokenOverrides } from "@/lib/tokens";

// Test styling data only: never persisted, never an Organization, never selectable at runtime. G069 (Plan r1 §4 P0.2):
// the fixture page renders one tenant palette at the document root (through the app's own TenantStyle, ?tenant=) and
// two isolated previews (BrandPreview, one appearance each). Every name a palette omits keeps the platform default.

/** One appearance's saved values: a partial record of the 15 saved names. */
export type Values = Partial<Record<BrandingToken, string>>;

// Preview A (Light only): surface, text, action, muted, accent, destructive and focus values; its popover pair differs from
// every root tenant's by more than 8 per channel, so a preview value that leaked into a portal would be told apart.
// Its destructive is the fixtureBranded Light value (G069 P0.2 run 4): the then platform value (the app correction
// #bb0009, withdrawn by the owner's kit-fidelity direction of 2026-10-26; the platform destructive is now the kit's own
// oklch(0.577 0.245 27.325), about #e7000b) passed the r3 T10 pairs only on a white ground and painted 4.19 / 3.80 on this
// tinted background / popover; #9a2c25 gives 4.96 / 4.53 on the 20 % tint (5.89 / 5.33 on the 10 % tint), above the
// full 4.5:1 threshold, so the value stays a tenant colour that never relies on the kit-baseline floor.
// Computed: sRGB composite at 8 bits.
export const fixtureA: Values = {
  background: "#eef5fa",
  foreground: "#10202e",
  primary: "#0b5c8f",
  "primary-foreground": "#ffffff",
  "muted-foreground": "#3d566b",
  accent: "#1f5f3a",
  destructive: "#9a2c25",
  ring: "#0b5c8f",
  popover: "#dcebf5",
  "popover-foreground": "#0b3a5c",
};

// Preview B (Dark only), with the same popover separation.
export const fixtureB: Values = {
  background: "#1c120b",
  foreground: "#f3e6d8",
  primary: "#f59e5b",
  "primary-foreground": "#1c120b",
  "muted-foreground": "#c9ad95",
  accent: "#f2b3c8",
  ring: "#f2b3c8",
  popover: "#2e1d10",
  "popover-foreground": "#f3e6d8",
};

// G017: a complete Organization-branding palette — every one of the 15 overridable names (lib/branding.ts
// BRANDING_TOKENS) in both appearances, as an accepted branding row would supply them. G069 (S23, D-M3a (i), Codex
// disposition 2026-10-26): retyped to the r3 pairs (core/internal/branding/colors.go: the kit Button's 80 % hover, the
// destructive text on its tints) for the rendered kit states. Light primary was #0f6b5c (white label 6.41 at rest but
// 4.25 / 4.13 on the 80 % hover over background / popover); #0a5a4d gives 8.13 / 5.09 / 4.93. Light destructive was
// #a8322a (4.42 as text on its 20 % tint over background); #9a2c25 gives 5.89 / 4.99 on the 10 % / 20 % tints over
// background, 6.46 / 5.43 over popover, 6.92 on background. Dark destructive was #ff7a6b (4.36 / 4.03 on the 30 % tint
// over background / popover); #ff9c8f gives 6.29 / 4.90 on the 20 % / 30 % tints over background, 5.81 / 4.55 over
// popover, 9.19 on background. Computed: sRGB composite at 8 bits. Every other value is unchanged.
export const fixtureBranded: TokenOverrides = {
  light: {
    background: "#f1f5f4",
    foreground: "#13201e",
    primary: "#0a5a4d",
    "primary-foreground": "#ffffff",
    accent: "#7a4a00",
    "muted-foreground": "#4a5d59",
    border: "#c9d6d3",
    popover: "#ffffff",
    "popover-foreground": "#13201e",
    muted: "#e6eeec",
    secondary: "#3b5a7a",
    "secondary-foreground": "#ffffff",
    destructive: "#9a2c25",
    input: "#4a5d59",
    ring: "#0a5a4d",
  },
  dark: {
    background: "#0b1514",
    foreground: "#e2ecea",
    primary: "#3fb8a3",
    "primary-foreground": "#06201b",
    accent: "#f0c674",
    "muted-foreground": "#93a8a4",
    border: "#24332f",
    popover: "#101d1b",
    "popover-foreground": "#e2ecea",
    muted: "#0f1b19",
    secondary: "#8fb3d9",
    "secondary-foreground": "#0b1514",
    destructive: "#ff9c8f",
    input: "#93a8a4",
    ring: "#f0c674",
  },
};

// G067 (Plan delta r2 §2.2, §3.2): a rendered tenant-override edge, Light only (Dark keeps the platform palette).
// G069 (S23, D-M3a (i), Codex disposition 2026-10-26): the G067 edge #696969 passed T8 at rest (white 5.49) only under
// the retired whole-element brightness hover and paints 3.59 on the kit Button's 80 % hover over the white page, so
// the r3 pairs refuse it. The retyped edge #525252 keeps the grey primary near the r3 limit: white 7.81 at rest and
// 4.61 on the 80 % hover (computed: sRGB composite at 8 bits).
export const fixtureGreyPrimary: TokenOverrides = {
  light: {
    primary: "#525252",
    "primary-foreground": "#ffffff",
  },
};

// G067 correction (Codex implementation review 2026-10-25, item 1): a sparse Dark-only destructive override (every
// other token keeps the platform palette). G069 (S23, D-M3a (i), Codex disposition 2026-10-26): the G067 value #c07070
// paints 3.67 / 3.18 as text on its 20 % / 30 % tints over the platform Dark background / popover, which the r3 pairs
// refuse; #e8a6a6 gives 6.88 / 5.80 on the 20 % tint and 5.31 / 4.51 on the 30 % tint over background / popover and 9.82
// on background (computed: sRGB composite at 8 bits). Since the owner's kit-fidelity direction (2026-10-26) the field
// renders the kit's own Dark invalid border (aria-invalid:border-destructive/50, no app solid boundary): S-FIELD-INVALID
// measures it as painted, about 3.6 against the page and 3.1 against the field fill, 1.5 against the kit's invalid halo
// (OWNER-WAIVED).
export const fixtureDarkSparseDestructive: TokenOverrides = {
  dark: {
    destructive: "#e8a6a6",
  },
};

/** A root palette of the fixture page, selected by ?tenant=<label> (default branded). */
export type FixtureTenant = { label: string; tokens?: { light?: Values; dark?: Values } };

// Order: platform (no tenant values: the kit's own defaults, no app correction since G069's owner kit-fidelity
// direction), branded, grey-primary, dark-sparse-destructive.
export const FIXTURE_TENANTS: readonly FixtureTenant[] = [
  { label: "platform" },
  { label: "branded", tokens: fixtureBranded },
  { label: "grey-primary", tokens: fixtureGreyPrimary },
  { label: "dark-sparse-destructive", tokens: fixtureDarkSparseDestructive },
];

/** An isolated preview of the fixture page: one appearance and that appearance's values. */
export type FixturePreview = { label: string; appearance: "light" | "dark"; tokens: Values };

export const FIXTURE_PREVIEWS: readonly FixturePreview[] = [
  { label: "A", appearance: "light", tokens: fixtureA },
  { label: "B", appearance: "dark", tokens: fixtureB },
];
