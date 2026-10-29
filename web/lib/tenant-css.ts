import type { SemanticColorToken, TokenOverrides, TokenValues } from "./tokens";

// G069 (Plan r1 S7–S9, §2.5): an Organization's saved branding rendered onto the adopted kit's variables. Pure and
// framework-free (unit-tested in tests/unit/branding-client.test.mjs); components/branding/tenant-style.tsx and
// components/branding/brand-preview.tsx render what it returns.

/** The 15 saved colour names (core branding.Tokens). */
export type BrandingToken = SemanticColorToken;

/**
 * The saved-branding fan-out: each saved name sets exactly these kit variables, in this order. `primary` also drives
 * `--vx-blue` (the kit's primary and blue utilities); `accent` is the gold emphasis ink and never the kit's neutral
 * `--accent`, which follows the saved `muted` fill; `popover` also paints the kit's card. Every other kit variable
 * (radius, charts, the remaining brand colours, the ribbon) keeps the kit's defaults and is not overridable.
 */
export const TENANT_FAN_OUT: Readonly<Record<BrandingToken, readonly string[]>> = {
  background: ["--background", "--sidebar"],
  foreground: ["--foreground", "--accent-foreground", "--sidebar-foreground"],
  primary: ["--primary", "--vx-blue", "--sidebar-primary"],
  "primary-foreground": ["--primary-foreground", "--sidebar-primary-foreground"],
  accent: ["--vx-gold"],
  muted: ["--muted", "--accent", "--sidebar-accent"],
  "muted-foreground": ["--muted-foreground"],
  secondary: ["--secondary"],
  "secondary-foreground": ["--secondary-foreground"],
  popover: ["--popover", "--card"],
  "popover-foreground": ["--popover-foreground", "--card-foreground"],
  border: ["--border", "--sidebar-border"],
  input: ["--input"],
  destructive: ["--destructive"],
  ring: ["--ring", "--sidebar-ring"],
};

const NAMES = Object.keys(TENANT_FAN_OUT) as BrandingToken[];
const COLOUR = /^#[0-9a-f]{6}$/;

// One appearance's declarations in fan-out order, or undefined when the values hold anything but saved names with
// lowercase #rrggbb colours. The branding reader already guarantees both; the emitter checks again, because what it
// returns is written into a stylesheet.
function declarations(values: TokenValues | undefined): [string, string][] | undefined {
  if (values === undefined) return [];
  if (typeof values !== "object" || values === null) return undefined;
  for (const [name, colour] of Object.entries(values)) {
    if (!Object.hasOwn(TENANT_FAN_OUT, name) || typeof colour !== "string" || !COLOUR.test(colour)) return undefined;
  }
  const out: [string, string][] = [];
  for (const name of NAMES) {
    const colour = values[name];
    if (colour !== undefined) for (const variable of TENANT_FAN_OUT[name]) out.push([variable, colour]);
  }
  return out;
}

/**
 * The tenant stylesheet text for the document root: `html:root:not(.dark){…}` from the Light values and
 * `html.dark:root{…}` from the Dark values (specificity 0,2,1: above the kit's `:root`/`.dark`), each
 * omitted when it has no values. Empty when there is no branding or anything in it fails the re-check (the platform
 * defaults then apply in both appearances).
 */
export function tenantCss(tokens: TokenOverrides | undefined): string {
  const light = declarations(tokens?.light);
  const dark = declarations(tokens?.dark);
  if (light === undefined || dark === undefined) return "";
  const block = (selector: string, list: [string, string][]) =>
    list.length === 0 ? "" : `${selector}{${list.map(([variable, colour]) => `${variable}:${colour}`).join(";")}}`;
  return block("html:root:not(.dark)", light) + block("html.dark:root", dark);
}

/**
 * One appearance's values as the inline custom properties of an isolated preview element (the same fan-out); empty
 * when anything fails the re-check.
 */
export function previewVariables(values: TokenValues | undefined): Record<string, string> {
  return Object.fromEntries(declarations(values) ?? []);
}
