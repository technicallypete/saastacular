import { expect, type Locator, type Page } from "@playwright/test";

import { APPEARANCE_STORAGE_KEY, type AppearanceMode } from "@/lib/appearance";

import { contrast, hex, hexToRgb, over, type RGBA } from "./color";
import type { PairResult } from "./contrast";
import type { RGB } from "./png";

// The 15 saved-branding names (core branding.Tokens), listed here independently of the application modules.
export const PALETTE_NAMES = [
  "background",
  "foreground",
  "primary",
  "primary-foreground",
  "accent",
  "muted-foreground",
  "border",
  "popover",
  "popover-foreground",
  "muted",
  "secondary",
  "secondary-foreground",
  "destructive",
  "input",
  "ring",
] as const;
export type PaletteName = (typeof PALETTE_NAMES)[number];

// The kit variable that carries each saved name at runtime (G069 Plan §2.5 fan-out): saved accent is the emphasis
// ink --vx-gold (never the kit's neutral --accent fill); every other name keeps its own variable.
export const CSS_VARIABLE: Record<PaletteName, string> = Object.fromEntries(PALETTE_NAMES.map((n) => [n, n === "accent" ? "--vx-gold" : `--${n}`])) as Record<PaletteName, string>;

// The kit's theme-fixed dialog scrim (dialog.tsx bg-black/30, the same in both appearances; G069 S16: no overlay token).
export const SCRIM = "rgb(0 0 0 / 0.3)";

// Independent expected platform values of the 15 names in both appearances: exactly the kit's styles/theme.css values
// (G069 owner kit-fidelity direction, 2026-10-26: "exact kit, overrides included"; the app's --vx-blue, muted-foreground,
// destructive and input corrections are withdrawn), typed here and never read from the stylesheets. The kit's opaque
// oklch values are converted to sRGB hex (each channel clamped, as the browser paints an out-of-gamut colour; the same
// conversion as core/internal/branding/colors.go): muted-foreground oklch(0.547 0.021 43.1) = #7c6d67, destructive
// oklch(0.577 0.245 27.325) = #e7000b Light and oklch(0.704 0.191 22.216) = #ff6467 Dark, input oklch(0.922 0.005 34.3)
// = #e8e4e3 Light; primary is the kit's --vx-blue #0075ff in both appearances. The kit's two translucent Dark values,
// border oklch(1 0 0 / 10%) and input oklch(1 0 0 / 15%), are kept as the kit writes them, so a consumer composites them
// over its own ground (alphaOver, paintedOver, roleColor) and never reads them as opaque. A transcription error in
// either the stylesheets or this table fails S-TOKENS, S-STATES and S-PALETTE.
// The retired overlay token has no entry: the dialog scrim is the kit's theme-fixed SCRIM (S16).
export const SOURCES: Record<"light" | "dark", Record<PaletteName, string>> = {
  light: {
    background: "#ffffff",
    foreground: "#0c0a09",
    primary: "#0075ff",
    "primary-foreground": "#ffffff",
    accent: "#8a5a00",
    "muted-foreground": "#7c6d67",
    border: "#e8e4e3",
    popover: "#ffffff",
    "popover-foreground": "#0c0a09",
    muted: "#f3f1f1",
    secondary: "#f3f1f1",
    "secondary-foreground": "#1d1816",
    destructive: "#e7000b",
    input: "#e8e4e3",
    ring: "#8a5a00",
  },
  dark: {
    background: "#0c0a09",
    foreground: "#fbfaf9",
    primary: "#0075ff",
    "primary-foreground": "#ffffff",
    accent: "#f8ce8a",
    "muted-foreground": "#aba09c",
    border: "oklch(1 0 0 / 10%)",
    popover: "#1d1816",
    "popover-foreground": "#fbfaf9",
    muted: "#2b2422",
    secondary: "#2b2422",
    "secondary-foreground": "#fbfaf9",
    destructive: "#ff6467",
    input: "oklch(1 0 0 / 15%)",
    ring: "#f8ce8a",
  },
};

// The owner's contrast waiver (G069 Spec §8; owner instruction 2026-10-26, "exact kit, overrides included"; record
// management/v0/results/G069-exact-kit-owner-decision-20261026.md): a pair the kit's own treatment paints below its former
// threshold is OWNER-WAIVED. It is never asserted against that threshold and never dropped: it is asserted at the kit's
// exact ratio within WAIVED_TOLERANCE, so a drift in either direction fails. Kit platform ratios (Core TestDefaultsPass,
// the coordinator's kit-fidelity record): white on #0075ff 4.21 in both appearances, on its Light 80 % hover 3.18;
// CtaButton white 4.21 / 3.77 (x1.10) / 4.02 (x0.90); Light destructive text on its 10 % / 20 % tints 3.99 / 3.32, Dark
// on its 30 % tint 4.33; the field boundary (N3: input over the page) 1.26 Light / 1.48 Dark.
export const OWNER_WAIVED = 'OWNER-WAIVED (G069 Spec §8, owner instruction 2026-10-26: "exact kit, overrides included")';
// Ratios are compared at two decimals; the measured pixel and the typed-table model differ by at most a unit or two per
// 8-bit channel, which moves these ratios by a few hundredths.
export const WAIVED_TOLERANCE = 0.1;

export type KitPairResult = PairResult & { kit: number; kitWaiver?: string };

// One contrast pair measured against the kit's own rendering: `a`/`b` as painted, `kitA`/`kitB` the colours the kit's
// treatment paints from the independent palette tables (or the page's own tokens where a helper reads them). When the kit
// ratio meets `min`, the measured pair must meet `min` (the former threshold is kept). When it does not, the pair is
// OWNER-WAIVED: the measured ratio must equal the kit ratio within WAIVED_TOLERANCE, and the result records the waiver.
export function expectAtKit(name: string, a: RGB, b: RGB, kitA: RGB, kitB: RGB, min: number): KitPairResult {
  const ratio = contrast(a, b);
  const kit = contrast(kitA, kitB);
  const base = { name, ratio: Number(ratio.toFixed(2)), fg: hex(a), bg: hex(b), min, kit: Number(kit.toFixed(2)) };
  if (kit >= min) {
    expect(ratio, `${name}: ${base.fg} against ${base.bg} (the kit meets ${min}:1 here: ${base.kit})`).toBeGreaterThanOrEqual(min);
    return base;
  }
  const waived = `${OWNER_WAIVED}: the kit paints ${base.kit}:1, below ${min}:1`;
  expect(Math.abs(ratio - kit), `${name}: ${base.fg} against ${base.bg} = ${base.ratio}:1, ${waived}; asserted at the kit ratio ±${WAIVED_TOLERANCE}`).toBeLessThanOrEqual(WAIVED_TOLERANCE);
  return { ...base, kitWaiver: waived };
}

// The kit's foreground at 80 % (Text lead, text-foreground/80) as its opaque equivalent over a ground.
export function foreground80(foreground: string, ground: string): string {
  return hex(over([...hexToRgb(foreground), 204] as RGBA, hexToRgb(ground)));
}

// Default palette values used as expected values in appearance and first-paint assertions. The H1 (kit Heading) is
// painted in foreground, so first-paint oracles tell it apart from the other text inks (foreground at 80 %, muted
// foreground) and from the other appearance instead.
const pick = (m: "light" | "dark") => {
  const s = SOURCES[m];
  return { background: s.background, foreground: s.foreground, primary: s.primary, popover: s.popover, ring: s.ring, accent: s.accent, ink: s.foreground, text2: foreground80(s.foreground, s.background), text3: s["muted-foreground"] };
};
export const PALETTE = { light: pick("light"), dark: pick("dark") } as const;

export type Resolved = "light" | "dark";

export function resolve(mode: AppearanceMode, os: Resolved): Resolved {
  return mode === "system" ? os : mode;
}

// Stores a mode for the origin before the page is loaded (same-origin health endpoint first).
export async function storeMode(page: Page, value: string | null): Promise<void> {
  await page.goto("/api/health");
  await page.evaluate(
    ([key, v]) => {
      if (v === null) window.localStorage.removeItem(key);
      else window.localStorage.setItem(key, v);
    },
    [APPEARANCE_STORAGE_KEY, value] as const,
  );
}

// Makes every localStorage access throw, before any page script (including the inline theme script) runs.
export async function blockStorage(page: Page): Promise<void> {
  await page.addInitScript(() => {
    Object.defineProperty(window, "localStorage", {
      configurable: true,
      get() {
        throw new DOMException("storage blocked", "SecurityError");
      },
    });
  });
}

// Hydration is complete once the help dialog opens from a click (an interactive React handler).
export async function waitForHydration(page: Page): Promise<number> {
  const trigger = page.getByRole("button", { name: "Keyboard help" }).first();
  const dialog = page.getByRole("dialog");
  // A click before React attaches handlers does nothing, so retry until the dialog opens.
  await expect(async () => {
    if (!(await dialog.isVisible())) await trigger.click();
    await expect(dialog).toBeVisible({ timeout: 500 });
  }).toPass({ timeout: 60_000 });
  const t = await page.evaluate(() => performance.now());
  await page.keyboard.press("Escape");
  await page.getByRole("dialog").waitFor({ state: "hidden" });
  // The dialog returns focus to the trigger after the popup unmounts; let it land before any test moves
  // focus itself.
  await expect(trigger).toBeFocused();
  return t;
}

export async function storedValue(page: Page): Promise<string | null> {
  return page.evaluate((key) => window.localStorage.getItem(key), APPEARANCE_STORAGE_KEY);
}

// The open popup of a kit Select (G069): the kit's select-content, portalled into the document body with its items as
// role=option inside the kit's padded select-group. At most one is open at a time, but a popup that has just closed
// stays in the document while its close motion runs (and Base UI marks it data-closed), so the locator names the open
// one by its data-open attribute: a page-wide select-content locator resolves to two elements when a second dropdown
// opens right after the first (strict mode violation, hosted B-IDENTITY-E2E). Every app dropdown is this control: the
// appearance control and, since G069 r3 (owner finding), every form dropdown through components/controls/form-select
// (the trigger is the combobox the site's label names).
export const SELECT_POPUP = "[data-slot=select-content][data-open]";

// Chooses an option of a kit Select as a pointer user does: open the list from its trigger, click the labelled option
// and wait for the popup to leave the document (the kit's close motion ends with its unmount). The replacement for
// selectOption({ label }) on the native selects the form dropdowns used to be.
export async function chooseOption(trigger: Locator, label: string): Promise<void> {
  const popup = trigger.page().locator(SELECT_POPUP);
  await trigger.click();
  await expect(popup).toBeVisible();
  await popup.getByRole("option", { name: label, exact: true }).click();
  await expect(popup).toHaveCount(0);
  // The closing popup's motion (100ms) ends with its unmount; let it finish before the next dropdown opens.
  await trigger.page().evaluate(() => Promise.allSettled(document.getAnimations().map((a) => a.finished)));
}

// The appearance control (G069, owner preview finding): the kit Select. Its trigger is the combobox named
// "Appearance" inside the [data-appearance-control] wrapper (the hidden value input is its sibling there); its popup is
// the kit's select-content in the document body, with one option per mode, named by the labels below.
export const APPEARANCE_CONTROL = "[data-appearance-control]";
export const APPEARANCE_POPUP = SELECT_POPUP;
export type AppearanceLabel = "Light" | "Dark" | "System";
export const APPEARANCE_LABELS: Record<AppearanceMode, AppearanceLabel> = { light: "Light", dark: "Dark", system: "System" };

export function appearanceControl(page: Page): Locator {
  return page.getByRole("combobox", { name: "Appearance", exact: true });
}

// The selected mode: an icon-only combobox has no text value, so the trigger mirrors the Select's value.
export async function appearanceValue(page: Page): Promise<string | null> {
  return appearanceControl(page).getAttribute("data-appearance-value");
}

// Chooses a mode as a pointer user does (chooseOption on the appearance trigger).
export async function chooseAppearance(page: Page, label: AppearanceLabel): Promise<void> {
  await chooseOption(appearanceControl(page), label);
}

// Parks the pointer in the viewport's bottom-right corner. (0, 0) is where the focused skip link appears,
// so a pointer left there would hover it.
export async function parkPointer(page: Page): Promise<void> {
  const vp = page.viewportSize() ?? { width: 1280, height: 800 };
  await page.mouse.move(vp.width - 1, vp.height - 1);
}

// Moves focus (and the sequential focus navigation starting point) to the start of the document,
// so the next Tab reaches the first focusable element. Pending tasks run first: a dialog that has just
// closed restores focus to its trigger asynchronously (the dialog's focus manager), which would otherwise
// land after the reset.
export async function resetFocus(page: Page): Promise<void> {
  await page.evaluate(() => new Promise<void>((r) => setTimeout(() => requestAnimationFrame(() => requestAnimationFrame(() => r())), 0)));
  await page.evaluate(() => {
    const body = document.body;
    body.setAttribute("tabindex", "-1");
    body.focus();
    body.removeAttribute("tabindex");
  });
}
