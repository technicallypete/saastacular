import { createHash } from "node:crypto";

import type { Locator, Page, Route, TestInfo } from "@playwright/test";

import { polygonPoints } from "./support/angle";
import { BLUEPRINT_PATH, BLUEPRINT_SHA256, SHELL_FONT_PATHS } from "./support/assets";
import { contrast, cssToRgba, distance, hex, hexToRgb, over, rgb, type RGBA } from "./support/color";
import { box, compositeSamples, crossCentre, dialogSettled, fillOf, haloOf, settled } from "./support/contrast";
import {
  BOUND,
  BUNDLED_FACES,
  BUNDLED_FAMILIES,
  expectFaces,
  expectFontBytes,
  expectFontMetadata,
  expectHeld,
  faceStatuses,
  familyMap,
  fontsReady,
  holdFonts,
  openReady,
  recordFontResponses,
  renderedFonts,
  synthesisProbe,
  synthesized,
  wantedStatus,
  type RenderedFont,
} from "./support/fonts";
import { expect, test } from "./support/page-errors";
import { APPEARANCE_CONTROL, APPEARANCE_POPUP, appearanceControl, CSS_VARIABLE, PALETTE_NAMES, parkPointer, resetFocus, SCRIM, SOURCES, waitForHydration, type PaletteName } from "./support/modes";
import { decodePng, pixelAt, type Image, type RGB } from "./support/png";
import { RESPONSIVE, SKIP_LINK, SOURCE_ANGLE, SOURCE_REFERENCE, STACKS, type Expectation, type Role } from "./support/source-reference";
import { reflowLayout } from "./support/spacing";

const DISPLAY = /^"?Oswald"?, "?Arial Narrow"?/;
const LABEL = /^"?Barlow Condensed"?, "?Arial Narrow"?/;
// The application UI face (G067): the bundled Geist first in the kit's sans stack (never Geist Mono).
const SANS = /^"?Geist"?, ui-sans-serif/;

const TRIGGER = 'header button[aria-haspopup="dialog"]';
// The appearance control's wrapper (G069, owner preview finding): a bare inline-flex span around the kit Select's
// trigger, with no radius or border of its own, so geometry and paint are measured on the trigger (appearanceControl).
const APPEARANCE = APPEARANCE_CONTROL;
// The sign-in page's primary action: the kit Button at size lg (G069 S24; the page is served on tenant hosts, so it is
// never the kit's fixed-white angled call to action).
const SIGN_IN_SUBMIT = "#sign-in form button[type=submit]";
const KICKER = "#overview [data-slot=kicker]";
const LEAD = "#overview h1 + [data-slot=text]";
const CELL_BODY = "[data-slot=ruled-grid] h3 + [data-slot=text]";
// The grid's index numbers (G069 S12: the kit's meta text).
const INDEX = "[data-slot=ruled-grid] [data-slot=text]:first-child";
const LAYER = "#overview [data-blueprint]";
const CLOSE = '[data-slot=dialog-content] > [data-slot="dialog-close"]';

type Mode = "light" | "dark";
const modeOf = (info: TestInfo) => (info.project.use.colorScheme === "dark" ? "dark" : "light") as Mode;

// The independent expected colour of a role as a CSS colour: a saved palette name's platform value (SOURCES), or that
// value at a kit opacity modifier (Tailwind's `<utility>-<name>/<percent>`: color-mix in oklab with transparent).
function roleColor(mode: Mode, role: Role): string {
  const [name, percent] = role.split("/");
  const base = SOURCES[mode][name as PaletteName];
  return percent === undefined ? base : `color-mix(in oklab, ${base} ${percent}%, transparent)`;
}

// A colour compared with a role, alpha included: both are composited over the page ground of the appearance (identity
// for an opaque colour). A translucent colour read back through the canvas is premultiplied at 8 bits, so translucent
// roles allow 2 per channel, opaque ones 1.
async function roleDelta(page: Page, mode: Mode, value: string, role: Role): Promise<{ measured: RGB; expected: RGB; delta: number; tolerance: number }> {
  const ground = hexToRgb(SOURCES[mode].background);
  const want = await cssToRgba(page, roleColor(mode, role));
  const measured = over(await cssToRgba(page, value), ground);
  const expected = over(want, ground);
  return { measured, expected, delta: distance(measured, expected), tolerance: want[3] < 255 ? 2 : 1 };
}

async function cssOf(page: Page, selector: string, props: string[]): Promise<Record<string, string>> {
  return page
    .locator(selector)
    .first()
    .evaluate((el, ps) => Object.fromEntries(ps.map((p) => [p, getComputedStyle(el).getPropertyValue(p)])), props);
}

async function sameColour(page: Page, value: string, want: string, what: string, tolerance = 1) {
  const got = rgb(await cssToRgba(page, value));
  const expected = rgb(await cssToRgba(page, want));
  expect(distance(got, expected), `${what}: ${value} vs ${want}`).toBeLessThanOrEqual(tolerance);
}

function nearRgb(actual: RGB, want: RGB, what: string, tolerance = 2) {
  expect(distance(actual, want), `${what}: ${hex(actual)} vs ${hex(want)}`).toBeLessThanOrEqual(tolerance);
}

// Resizes the viewport mid-test and waits until the page has laid out at the new size. Besides the window size, it
// waits (bounded) until viewport units resolve against the new viewport: a test-owned, hidden 100vw x 100vh probe,
// removed afterwards, reaches exactly the new size. That is an independent signal, never the measured styles, so a
// wrong style still fails its assertion. Viewport-unit sizes (the gutter's and headings' clamp(..vw..)) were once
// observed at the previous width after the window already reported the new one (G051 O2). Returns a bounded record
// of what was observed before and after the wait, for the caller to attach.
type ResizeObservation = { width: number; height: number; waitedMs: number; settled: boolean; before: Record<string, unknown>; after: Record<string, unknown> };
async function resize(page: Page, width: number, height: number): Promise<ResizeObservation> {
  await page.setViewportSize({ width, height });
  await page.waitForFunction(([w, h]) => window.innerWidth === w && window.innerHeight === h, [width, height] as const);
  await page.evaluate(() => new Promise<void>((r) => requestAnimationFrame(() => requestAnimationFrame(() => r()))));
  const observe = () =>
    page.evaluate(([w, h]) => {
      const probe = document.getElementById("g051-viewport-probe");
      const gutter = document.querySelector("#overview [data-slot=container]");
      const h1 = document.querySelector("#overview h1");
      return {
        innerWidth: window.innerWidth,
        innerHeight: window.innerHeight,
        clientWidth: document.documentElement.clientWidth,
        probe: probe ? `${probe.getBoundingClientRect().width}x${probe.getBoundingClientRect().height}` : "absent",
        mediaMatchesViewport: matchMedia(`(width: ${w}px) and (height: ${h}px)`).matches,
        gutterPaddingLeft: gutter ? getComputedStyle(gutter).paddingLeft : "absent",
        h1FontSize: h1 ? getComputedStyle(h1).fontSize : "absent",
      };
    }, [width, height] as const);
  await page.evaluate(() => {
    const probe = document.createElement("div");
    probe.id = "g051-viewport-probe";
    probe.setAttribute("aria-hidden", "true");
    probe.style.cssText = "position:fixed;left:0;top:0;width:100vw;height:100vh;visibility:hidden;pointer-events:none;contain:strict";
    document.body.append(probe);
  });
  const before = await observe();
  const started = Date.now();
  let settled = true;
  try {
    await page.waitForFunction(
      ([w, h]) => {
        const r = document.getElementById("g051-viewport-probe")?.getBoundingClientRect();
        return !!r && r.width === w && r.height === h;
      },
      [width, height] as const,
      { timeout: 5000 },
    );
    await page.evaluate(() => new Promise<void>((r) => requestAnimationFrame(() => requestAnimationFrame(() => r()))));
  } catch {
    settled = false;
  }
  const after = await observe();
  await page.evaluate(() => document.getElementById("g051-viewport-probe")?.remove());
  const observation = { width, height, waitedMs: Date.now() - started, settled, before, after };
  if (!settled) throw new Error(`viewport units did not resolve against ${width}x${height} within 5 s: ${JSON.stringify(observation)}`);
  return observation;
}

// The layers of a computed box-shadow (Chromium serializes each as "<colour> <x> <y> <blur> <spread>[ inset]").
type ShadowLayer = { color: string; geometry: string };
function shadowLayers(value: string): ShadowLayer[] {
  if (value.trim() === "none") return [];
  return value.split(/,(?![^(]*\))/).map((layer) => {
    const m = layer.trim().match(/^(.*\S)\s+(-?[\d.]+px\s+-?[\d.]+px\s+-?[\d.]+px\s+-?[\d.]+px)(\s+inset)?$/);
    return m ? { color: m[1], geometry: m[2].replace(/\s+/g, " ") } : { color: "", geometry: layer.trim() };
  });
}

async function openDialog(page: Page): Promise<Locator> {
  await page.evaluate(() => window.scrollTo(0, 0));
  await page.locator(TRIGGER).click();
  const dialog = page.getByRole("dialog");
  await expect(dialog).toBeVisible();
  // Read at rest, after the open fade and zoom (G067 motion delta).
  await dialogSettled(page);
  return dialog;
}

test("[S-DESIGN] the supplied visual language is observable in computed styles", async ({ page }, testInfo) => {
  const mode = modeOf(testInfo);
  await page.setViewportSize({ width: 1280, height: 800 });
  await page.goto("/");
  await waitForHydration(page);
  await page.mouse.move(0, 0);

  // Rhea geometry (kit, `--radius` 0.625rem): controls and fields rounded-2xl = 18px; the appearance control is the
  // kit SelectTrigger as shipped (18px, a 1px border kept transparent over its bg-input/50 fill); the dialog
  // min(--radius-4xl, 24px) = 24px (below).
  const trigger = page.locator(TRIGGER);
  await expect(trigger).toHaveCount(1);
  const account = page.locator('[data-slot=site-header] a[href="/account"]');
  await expect(account).toHaveCount(1);
  const appearanceTrigger = appearanceControl(page);
  await expect(appearanceTrigger).toHaveCount(1);
  for (const [what, control] of [["help trigger", trigger], ["Account link", account], ["appearance control", appearanceTrigger]] as const) {
    expect(await control.evaluate((el) => getComputedStyle(el).borderRadius), `${what} radius`).toBe("18px");
  }
  // G069 S7: the palette lives at the document root (the kit's :root/.dark, a tenant's stylesheet).
  const tokenOf = (name: string) => page.evaluate((n) => getComputedStyle(document.documentElement).getPropertyValue(n).trim(), name);
  // The kit SelectTrigger at rest (border-transparent border bg-input/50): a 1px solid border with no colour of its
  // own, over the input role at 50 % (color-mix in oklab with transparent, as Tailwind compiles the modifier).
  const frame = await appearanceTrigger.evaluate((el) => {
    const s = getComputedStyle(el);
    return { width: s.borderLeftWidth, style: s.borderLeftStyle, color: s.borderLeftColor, fill: s.backgroundColor };
  });
  expect([frame.width, frame.style], "appearance control frame").toEqual(["1px", "solid"]);
  expect((await cssToRgba(page, frame.color))[3], `appearance control frame colour ${frame.color}: the kit's transparent border`).toBe(0);
  const fillWant = await cssToRgba(page, `color-mix(in oklab, ${await tokenOf("--input")} 50%, transparent)`);
  const fillGot = await cssToRgba(page, frame.fill);
  expect(Math.max(...fillGot.map((v, i) => Math.abs(v - fillWant[i]))), `appearance control fill ${frame.fill}: the kit's bg-input/50`).toBeLessThanOrEqual(2);
  // Visible structure: 1px hairline rules in the border token on header, sections, footer and the ruled grid.
  const ruleColor = await tokenOf("--border");
  for (const [sel, side] of [["[data-slot=site-header]", "bottom"], ["#overview", "bottom"], ["#getting-around", "bottom"], ["footer", "top"], ["[data-slot=ruled-grid] > div", "top"]] as const) {
    const s = await cssOf(page, sel, [`border-${side}-width`, `border-${side}-color`, `border-${side}-style`]);
    expect(s[`border-${side}-width`], `${sel} rule width`).toBe("1px");
    expect(s[`border-${side}-style`]).toBe("solid");
    await sameColour(page, s[`border-${side}-color`], ruleColor, `${sel} rule colour`);
  }
  // Typography roles (G069: the kit's typography, Plan S12): every heading is the kit's Heading in the heading face,
  // uppercase by CSS; navigation and buttons stay Geist in sentence case; the label face for the kicker.
  for (const sel of ["h1", "#getting-around h2", "#getting-around h3"]) {
    const s = await cssOf(page, sel, ["font-family", "text-transform"]);
    expect(s["font-family"], `${sel} font (kit Heading)`).toMatch(DISPLAY);
    expect(s["text-transform"], `${sel} case`).toBe("uppercase");
  }
  for (const sel of ["nav a", TRIGGER]) {
    const s = await cssOf(page, sel, ["font-family", "text-transform"]);
    expect(s["font-family"], `${sel} font`).toMatch(SANS);
    expect(s["text-transform"], `${sel} case`).toBe("none");
  }
  const kicker = await cssOf(page, KICKER, ["font-family", "letter-spacing", "text-transform", "color"]);
  expect(kicker["font-family"], "brand kicker font").toMatch(LABEL);
  expect(kicker["letter-spacing"]).toBe("2.5px");
  expect(kicker["text-transform"]).toBe("uppercase");
  // The kit's gold emphasis ink (text-gold): saved accent drives --vx-gold, never the kit's neutral --accent fill (§2.5).
  await sameColour(page, kicker.color, await tokenOf("--vx-gold"), "kicker colour");
  // The angled silhouette is the kit CtaButton's alone (S24, S-ANGLE); the kit's token is the source polygon and an
  // everyday button draws no clipped layer and is never clipped itself.
  const angle = await tokenOf("--vx-angle");
  expect(polygonPoints(angle, 160, 32), "--vx-angle is the source polygon").toEqual(polygonPoints(SOURCE_ANGLE, 160, 32));
  expect(await trigger.evaluate((el) => getComputedStyle(el, "::before").clipPath), "the everyday trigger has no angled layer").toBe("none");
  expect(await trigger.evaluate((el) => getComputedStyle(el).clipPath), "the control itself is never clipped").toBe("none");
  // Bullets are 14x2 gold rules (the kit's bg-gold fill).
  const bullet = await cssOf(page, "#about [data-slot=bullet-item] > span[aria-hidden]", ["width", "height", "background-color"]);
  expect(bullet.width).toBe("14px");
  expect(bullet.height).toBe("2px");
  await sameColour(page, bullet["background-color"], await tokenOf("--vx-gold"), "bullet rule colour");
  // The band surface: the Getting around section sits on the kit's muted tone (bg-muted/40), unlike the page ground.
  const band = await page.locator("#getting-around").evaluate((el) => {
    const probe = document.createElement("div");
    probe.style.backgroundColor = "color-mix(in oklab, var(--muted) 40%, transparent)";
    el.appendChild(probe);
    const want = getComputedStyle(probe).backgroundColor;
    probe.remove();
    return { got: getComputedStyle(el).backgroundColor, want, ground: getComputedStyle(document.body).backgroundColor };
  });
  expect(band.got, "Getting around is a muted band").toBe(band.want);
  expect(band.got, "the band differs from the page ground").not.toBe(band.ground);
  // Container and flush-left hierarchy.
  expect((await cssOf(page, "[data-slot=site-header] > [data-slot=container]", ["max-width"]))["max-width"]).toBe("1360px");
  for (const sel of ["h1", "#getting-around h2", KICKER]) expect((await cssOf(page, sel, ["text-align"]))["text-align"]).toMatch(/start|left/);

  // The dialog is the kit's Rhea panel: 24px radius, no border, a 1px foreground ring (ring-1, colour in S-PALETTE)
  // over the shadow-xl layers, no blur.
  const dialog = await openDialog(page);
  const popup = await cssOf(page, "[data-slot=dialog-content]", ["box-shadow", "border-radius", ...["top", "right", "bottom", "left"].map((side) => `border-${side}-width`)]);
  const geometries = shadowLayers(popup["box-shadow"]).map((l) => l.geometry);
  expect(geometries, "the dialog's 1px ring layer (ring-1)").toContain("0px 0px 0px 1px");
  expect(geometries, "the dialog's shadow-xl layers").toEqual(expect.arrayContaining(["0px 20px 25px -5px", "0px 8px 10px -6px"]));
  expect(popup["border-radius"]).toBe("24px");
  for (const side of ["top", "right", "bottom", "left"]) expect(popup[`border-${side}-width`], `dialog ${side}: no border (the ring is its edge)`).toBe("0px");
  // Close (G069: the kit DialogContent's own Close, web/src/components/ui/dialog.tsx showCloseButton): the kit's
  // icon-sm ghost Button (18px radius) on bg-secondary with no text colour of its own, so at rest it inherits the
  // popup's text-popover-foreground; the ghost variant's foreground on its muted hover fill (Dark: muted at 50 % over
  // the popover); the painted centre of its cross in both states.
  const close = dialog.getByRole("button", { name: "Close" });
  expect(await close.evaluate((el) => getComputedStyle(el).borderRadius), "Close radius").toBe("18px");
  const closeRest = await close.evaluate((el) => ({ color: getComputedStyle(el).color, background: getComputedStyle(el).backgroundColor }));
  await sameColour(page, closeRest.color, SOURCES[mode]["popover-foreground"], "Close colour at rest (inherited popover-foreground)");
  await sameColour(page, closeRest.background, SOURCES[mode].secondary, "Close fill at rest");
  nearRgb(await crossCentre(page, close), hexToRgb(SOURCES[mode]["popover-foreground"]), "Close cross at rest (painted centre)", 8);
  await close.hover();
  await settled(close);
  await sameColour(page, await close.evaluate((el) => getComputedStyle(el).color), SOURCES[mode].foreground, "Close colour on hover");
  const cb = await box(close);
  const hoverFill = mode === "dark" ? over(await cssToRgba(page, `color-mix(in oklab, ${SOURCES.dark.muted} 50%, transparent)`), hexToRgb(SOURCES.dark.popover)) : hexToRgb(SOURCES.light.muted);
  nearRgb(await pixelAt(page, cb.x + cb.width / 2, cb.y + 3), hoverFill, "Close fill on hover (ghost hover:bg-muted, Dark muted/50 over the popover)");
  nearRgb(await crossCentre(page, close), hexToRgb(SOURCES[mode].foreground), "Close cross on hover (painted centre)", 8);
  await page.mouse.move(0, 0);
  // G069 S16: the kit's theme-fixed scrim (bg-black/30) with its backdrop blur (backdrop-blur-sm, 8px).
  const overlay = await cssOf(page, "[data-slot=dialog-overlay]", ["backdrop-filter", "background-color"]);
  expect(overlay["backdrop-filter"]).toBe("blur(8px)");
  const want = await cssToRgba(page, SCRIM);
  const got = await cssToRgba(page, overlay["background-color"]);
  expect(Math.max(...got.map((v, i) => Math.abs(v - want[i]))), `overlay ${overlay["background-color"]} vs ${SCRIM}`).toBeLessThanOrEqual(1);
  await page.keyboard.press("Escape");
});

// Every palette variable the document root resolves parses to its independent expected value (alpha included), in both
// appearances (G069 S4/S10: the kit's :root/.dark exactly, no app corrections; the 15 saved names through the variables
// they set, §2.5; no per-appearance source variables remain). The other appearance is reached by emulating the OS
// scheme (no stored choice: the theme provider follows it); the project's own scheme is restored.
async function expectSources(page: Page, mode: Mode) {
  const scheme = async (m: Mode) => {
    await page.emulateMedia({ colorScheme: m });
    await expect.poll(() => page.evaluate((c) => document.documentElement.classList.contains(c), m), { message: `<html> class follows OS ${m}` }).toBe(true);
    await page.evaluate(() => new Promise<void>((r) => requestAnimationFrame(() => requestAnimationFrame(() => r()))));
  };
  for (const m of [mode, mode === "light" ? "dark" : "light"] as const) {
    await scheme(m);
    for (const name of PALETTE_NAMES) {
      const variable = CSS_VARIABLE[name];
      const value = await page.evaluate((v) => getComputedStyle(document.documentElement).getPropertyValue(v).trim(), variable);
      const got = await cssToRgba(page, value);
      const expected = await cssToRgba(page, SOURCES[m][name]);
      expect(Math.max(...got.map((v, i) => Math.abs(v - expected[i]))), `${m} ${variable}: ${value} vs ${SOURCES[m][name]}`).toBeLessThanOrEqual(1);
    }
  }
  await scheme(mode);
}

test("[S-PALETTE] the reviewed successor palette sources and their mapped consumers", async ({ page }, testInfo) => {
  const mode = modeOf(testInfo);
  await page.setViewportSize({ width: 1280, height: 800 });
  await page.emulateMedia({ reducedMotion: "reduce" });
  await page.goto("/");
  await waitForHydration(page);
  await page.mouse.move(0, 0);
  await expectSources(page, mode);
  // A role as painted over the page ground (translucent roles composited, as the browser paints them).
  const src = async (role: Role): Promise<RGB> => over(await cssToRgba(page, roleColor(mode, role)), hexToRgb(SOURCES[mode].background));
  const color = async (sel: string, prop = "color"): Promise<RGB> =>
    over(await cssToRgba(page, await page.locator(sel).first().evaluate((el, p) => getComputedStyle(el).getPropertyValue(p), prop)), hexToRgb(SOURCES[mode].background));
  // Rendered grounds and rules.
  const header = await box(page.locator("[data-slot=site-header]"));
  nearRgb(await pixelAt(page, header.x + 4, header.y + 4), await src("background"), "header ground");
  nearRgb(await pixelAt(page, header.x + 10, header.y + header.height - 1), await src("border"), "header bottom rule");
  const band = await box(page.locator("#getting-around"));
  nearRgb(await pixelAt(page, band.x + 4, band.y + 4), await src("muted/40"), "Getting around band (the kit's muted tone, muted at 40 % over the page)");
  // Text roles (G069 Plan §6.2: the kit's typography): H1 foreground, kicker gold (saved accent), lead foreground/80,
  // cell body muted-foreground, nav foreground/80, footer meta muted-foreground; each over the page ground.
  nearRgb(await color("#overview h1"), await src("foreground"), "H1");
  nearRgb(await color(KICKER), await src("accent"), "kicker");
  nearRgb(await color(LEAD), await src("foreground/80"), "lead");
  nearRgb(await color(CELL_BODY), await src("muted-foreground"), "cell body");
  nearRgb(await color("[data-slot=site-header] nav a"), await src("foreground/80"), "nav link");
  nearRgb(await color("footer [data-slot=text]"), await src("muted-foreground"), "footer meta");
  // Primary action and appearance control.
  const trigger = page.locator(TRIGGER);
  nearRgb(await fillOf(page, trigger), await src("primary"), "trigger fill");
  nearRgb(await color(TRIGGER), await src("primary-foreground"), "trigger label");
  // Amendment 03 (the kit Select since G069): the compact dropdown trigger is the kit SelectTrigger as shipped, its
  // bg-input/50 fill over the header ground inside a 1px transparent border (the background reaches the border box, so
  // the edge shows the same fill and no separate rule), its selected symbol (System: Monitor) drawn in the foreground
  // role beside the kit's chevron (muted-foreground, the trigger's two visible symbols). G067: the trigger is rounded to
  // 18px, so the fill is read inside the straight top edge (clear of the corner radius) and on that straight top edge at
  // mid-width, in the device pixel the border is painted on (box decorations are pixel-snapped to the nearest edge; the
  // left-edge midpoint of the 18px pill is the curve's apex and anti-aliased).
  const appearance = appearanceControl(page);
  const frame = await box(appearance);
  const inputFill = over(await cssToRgba(page, `color-mix(in oklab, ${SOURCES[mode].input} 50%, transparent)`), hexToRgb(SOURCES[mode].background));
  nearRgb(await pixelAt(page, frame.x + frame.width / 2, frame.y + 4), inputFill, "appearance trigger fill (input at 50 % over the header ground)");
  nearRgb(await pixelAt(page, frame.x + frame.width / 2, Math.round(frame.y) + 0.5), inputFill, "appearance frame (transparent border over the same fill)");
  nearRgb(await color(`${APPEARANCE} [data-appearance-icon=system]`), await src("foreground"), "appearance symbol");
  await expect(appearance.locator("svg:visible")).toHaveCount(2);
  // Dialog: panel, title, description, ring and the scrim over the page ground (top-left corner).
  await openDialog(page);
  const popup = page.locator("[data-slot=dialog-content]");
  await expect(popup).toHaveCSS("opacity", "1");
  const pb = await box(popup);
  // Bottom padding at mid-width: clear of the 24px corner radius.
  const panel = await src("popover");
  nearRgb(await pixelAt(page, pb.x + pb.width / 2, pb.y + pb.height - 4), panel, "dialog panel");
  // The panel edge is the kit's ring-1 in foreground at 5 % (Light) / 10 % (Dark), composited over the panel (the
  // shadow-xl layers beneath it make a painted sample inexact, so the computed ring layer is compared).
  const ring = shadowLayers(await popup.evaluate((el) => getComputedStyle(el).boxShadow)).find((l) => l.geometry === "0px 0px 0px 1px");
  expect(ring, "the dialog's ring layer").toBeTruthy();
  const ringWant = await cssToRgba(page, `color-mix(in oklab, ${roleColor(mode, "foreground")} ${mode === "dark" ? 10 : 5}%, transparent)`);
  nearRgb(over(await cssToRgba(page, ring!.color), panel), over(ringWant, panel), "dialog panel ring", 1);
  // The title inherits the popup's text-popover-foreground.
  nearRgb(await color("[data-slot=dialog-title]"), await src("popover-foreground"), "dialog title");
  nearRgb(await color("[data-slot=dialog-description]"), await src("muted-foreground"), "dialog description");
  // G069 S16: the kit's theme-fixed scrim (bg-black/30) over the page ground; its backdrop blur leaves the uniform
  // ground in the corner unchanged.
  const scrim: RGBA = await cssToRgba(page, SCRIM);
  nearRgb(await pixelAt(page, 2, 2), over(scrim, await src("background")), "scrim over the page ground");
  await page.keyboard.press("Escape");
});

// --- S-FIDELITY (§4.7) ---------------------------------------------------------------------------------------

type FidelityRow = { name: string; cite: string; property: string; expected: string | number; measured: string; delta?: number; verdict: "pass" | "fail" };

const normalizeStack = (value: string) => value.split(",").map((f) => f.trim().replace(/^["']|["']$/g, ""));

// Measures one expectation row on the first matching element. Lengths ±1 CSS px; categories exact.
async function measure(page: Page, mode: Mode, e: Expectation, rows: FidelityRow[]) {
  const loc = page.locator(e.selector);
  expect(await loc.count(), `${e.name}: ${e.selector} is rendered`).toBeGreaterThan(0);
  const props = [...Object.keys(e.px ?? {}), ...Object.keys(e.lengthExact ?? {}), ...Object.keys(e.exact ?? {}), "font-family", "color", "background-color", "border-bottom-color", "max-width"];
  const got = await loc.first().evaluate(
    (el, [ps, ch]) => {
      const s = getComputedStyle(el);
      const out: Record<string, string> = {};
      for (const p of ps) out[p] = s.getPropertyValue(p);
      // Offsets of an absolutely positioned control from its containing block's padding box.
      const parent = (el as HTMLElement).offsetParent as HTMLElement | null;
      if (parent) {
        const r = el.getBoundingClientRect();
        const pr = parent.getBoundingClientRect();
        out["@top"] = `${r.top - (pr.top + parent.clientTop)}px`;
        out["@right"] = `${pr.left + parent.clientLeft + parent.clientWidth - r.right}px`;
      }
      // N ch of this element's own font: a probe with the same font and max-width Nch, in the same parent.
      if (ch) {
        const probe = document.createElement("div");
        probe.style.position = "absolute";
        probe.style.visibility = "hidden";
        probe.style.fontFamily = s.fontFamily;
        probe.style.fontSize = s.fontSize;
        probe.style.fontWeight = s.fontWeight;
        probe.style.fontStyle = s.fontStyle;
        probe.style.fontStretch = s.fontStretch;
        probe.style.maxWidth = `${ch}ch`;
        el.parentElement!.appendChild(probe);
        out["@ch"] = getComputedStyle(probe).maxWidth;
        probe.remove();
      }
      return out;
    },
    [props, e.maxWidthCh ?? 0] as const,
  );
  const row = (property: string, expected: string | number, measured: string, ok: boolean, delta?: number) =>
    rows.push({ name: e.name, cite: e.cite, property, expected, measured, delta, verdict: ok ? "pass" : "fail" });
  for (const [prop, want] of Object.entries(e.px ?? {})) {
    const raw = prop === "top" || prop === "right" ? got[`@${prop}`] : got[prop];
    const delta = Math.abs(parseFloat(raw) - want);
    row(prop, want, raw, delta <= 1, Number(delta.toFixed(3)));
  }
  for (const [prop, want] of Object.entries(e.lengthExact ?? {})) {
    const delta = Math.abs(parseFloat(got[prop]) - want);
    row(prop, want, got[prop], delta <= 0.05, Number(delta.toFixed(3)));
  }
  for (const [prop, want] of Object.entries(e.exact ?? {})) row(prop, want, got[prop], got[prop] === want);
  if (e.maxWidthCh) {
    const delta = Math.abs(parseFloat(got["max-width"]) - parseFloat(got["@ch"]));
    row("max-width", `${e.maxWidthCh}ch = ${got["@ch"]}`, got["max-width"], delta <= 0.5, Number(delta.toFixed(3)));
  }
  if (e.stack) {
    const stack = normalizeStack(got["font-family"]);
    row("font-family", STACKS[e.stack].join(", "), got["font-family"], JSON.stringify(stack) === JSON.stringify(STACKS[e.stack]));
  }
  // Colours with their alpha: both composited over the page ground (roleDelta), so a translucent kit colour such as
  // text-foreground/80 is compared as painted and an opaque one exactly as before.
  for (const [prop, role] of [["color", e.color], ["background-color", e.background], ["border-bottom-color", e.border]] as const) {
    if (!role) continue;
    const c = await roleDelta(page, mode, got[prop], role);
    row(prop, `${role} ${roleColor(mode, role)} = ${hex(c.expected)}`, `${got[prop]} = ${hex(c.measured)}`, c.delta <= c.tolerance, c.delta);
  }
}

const SLOTS_CLOSED = ["kicker", "heading", "text", "text-link", "container", "section", "ruled-grid", "bullet-list", "bullet-item", "site-header", "button"];
const SLOTS_OPEN = ["dialog-overlay", "dialog-content", "dialog-header", "dialog-title", "dialog-description", "dialog-close"];

async function slotsPresent(page: Page, slots: string[]): Promise<string[]> {
  return page.evaluate((ss) => ss.filter((s) => document.querySelector(`[data-slot="${s}"]`) === null), slots);
}

test("[S-FIDELITY] source reference geometry, type, colour roles and rendered structure", async ({ page }, testInfo) => {
  const mode = modeOf(testInfo);
  const rows: FidelityRow[] = [];
  await page.setViewportSize({ width: 1280, height: 800 });
  await page.emulateMedia({ reducedMotion: "reduce" });
  // Measured only after the bounded phase-B readiness gate (§7.4 B0): never on fallback metrics.
  await openReady(page);
  await page.mouse.move(0, 0);
  // The hydration probe closes the help dialog with Escape, which returns keyboard-visible focus to the trigger; the kit
  // Button draws its focus state as a ring (box-shadow) and a ring-coloured border, so the rest rows below are measured
  // with focus moved off the trigger.
  await resetFocus(page);
  expect(await slotsPresent(page, SLOTS_CLOSED), "module slots rendered on the production page").toEqual([]);

  const closedRows = SOURCE_REFERENCE.filter((e) => !e.selector.includes("dialog"));
  const openRows = SOURCE_REFERENCE.filter((e) => e.selector.includes("dialog"));
  for (const e of closedRows) await measure(page, mode, e, rows);

  // Help trigger (the kit's everyday Button, size sm): its own primary fill and no angled ::before layer (the angle is
  // the kit CtaButton's alone, S24); hover is the kit's bg-primary/80 composite over the header ground (G069 S11: core
  // T8 checks primary-foreground on primary at 80 % over background and popover; no brightness filter); a pressed
  // dialog trigger does not shift.
  const trigger = page.locator(TRIGGER);
  const rest = await trigger.evaluate((el) => ({ background: getComputedStyle(el).backgroundColor, before: getComputedStyle(el, "::before").clipPath }));
  rows.push({
    name: "help trigger ::before",
    cite: "kit web/src/components/ui/button.tsx:buttonVariants (no ::before layer; only the kit CtaButton clips to --vx-angle)",
    property: "clip-path",
    expected: "none",
    measured: rest.before,
    verdict: rest.before === "none" ? "pass" : "fail",
  });
  const fill = rgb(await cssToRgba(page, rest.background));
  rows.push({
    name: "help trigger",
    cite: "kit web/src/components/ui/button.tsx:buttonVariants default bg-primary",
    property: "background-color",
    expected: `primary ${SOURCES[mode].primary}`,
    measured: hex(fill),
    verdict: distance(fill, hexToRgb(SOURCES[mode].primary)) <= 1 ? "pass" : "fail",
  });
  // The expected hover: the source primary at 80 % (color-mix in oklab with transparent, as Tailwind compiles the
  // modifier) composited over the header's ground, the page background.
  const hoverMix = `color-mix(in oklab, ${SOURCES[mode].primary} 80%, transparent)`;
  const hoverWant = over(await cssToRgba(page, hoverMix), hexToRgb(SOURCES[mode].background));
  await trigger.hover();
  // Read once the transition has settled. The kit Button's transition-all (150ms) runs under reduced motion too (the
  // kit has no reduced-motion variant for it), and its time advances only when a frame renders, so on a CPU-starved
  // runner a fixed wait can still read its start value.
  await settled(trigger);
  const hover = await trigger.evaluate((el) => ({ filter: getComputedStyle(el).filter, background: getComputedStyle(el).backgroundColor }));
  rows.push({
    name: "help trigger (hover)",
    cite: "kit web/src/components/ui/button.tsx:buttonVariants default hover:bg-primary/80 (no filter)",
    property: "filter",
    expected: "none",
    measured: hover.filter,
    verdict: hover.filter === "none" ? "pass" : "fail",
  });
  const hoverComputed = await roleDelta(page, mode, hover.background, "primary");
  const hoverComputedDelta = distance(hoverComputed.measured, hoverWant);
  rows.push({
    name: "help trigger (hover)",
    cite: "kit web/src/components/ui/button.tsx:buttonVariants default hover:bg-primary/80",
    property: "background-color",
    expected: `${hoverMix} over background = ${hex(hoverWant)}`,
    measured: `${hover.background} = ${hex(hoverComputed.measured)}`,
    delta: hoverComputedDelta,
    verdict: hoverComputedDelta <= 2 ? "pass" : "fail",
  });
  const hoverGot = await fillOf(page, trigger);
  // Channels compared on the rendered 8-bit fill (±2).
  const hoverDelta = distance(hoverGot, hoverWant);
  rows.push({
    name: "help trigger (hover)",
    cite: "kit web/src/components/ui/button.tsx:buttonVariants default hover:bg-primary/80",
    property: "rendered fill",
    expected: `primary ${SOURCES[mode].primary} at 80 % over background ${SOURCES[mode].background} = ${hex(hoverWant)}`,
    measured: hex(hoverGot),
    delta: hoverDelta,
    verdict: hoverDelta <= 2 ? "pass" : "fail",
  });
  // Active without activating: press, read, move away, release elsewhere (no click is delivered).
  await page.mouse.down();
  await settled(trigger);
  const activeTranslate = await trigger.evaluate((el) => getComputedStyle(el).translate);
  await page.mouse.move(0, 0);
  await page.mouse.up();
  rows.push({
    name: "help trigger (active)",
    cite: "kit components/ui/button.tsx:buttonVariants active:not-aria-[haspopup]:translate-y-px (a dialog trigger has aria-haspopup)",
    property: "translate",
    expected: "none",
    measured: activeTranslate,
    verdict: activeTranslate === "none" ? "pass" : "fail",
  });
  await expect(page.getByRole("dialog"), "pressing and releasing elsewhere does not open the dialog").toHaveCount(0);

  // The skip link while focused (the pointer parked away from where it appears, so it is not hovered).
  await parkPointer(page);
  await resetFocus(page);
  await page.keyboard.press("Tab");
  await expect(page.getByRole("link", { name: "Skip to main content" })).toBeFocused();
  await measure(page, mode, SKIP_LINK, rows);
  await resetFocus(page);

  // G069: the kit's DialogHeader has no right padding (the legacy module's pr-9, X13), so that the corner Close never
  // covers the title or description is measured geometrically: no text rectangle of either intersects the Close.
  const clearOfClose = async (label: string) => {
    const hits = await page.evaluate((closeSelector) => {
      const close = document.querySelector(closeSelector)!.getBoundingClientRect();
      const out: string[] = [];
      for (const sel of ["[data-slot=dialog-title]", "[data-slot=dialog-description]"]) {
        const range = document.createRange();
        range.selectNodeContents(document.querySelector(sel)!);
        for (const t of range.getClientRects()) {
          if (t.width > 0 && t.left < close.right - 0.5 && close.left < t.right - 0.5 && t.top < close.bottom - 0.5 && close.top < t.bottom - 0.5) out.push(sel);
        }
      }
      return out;
    }, CLOSE);
    rows.push({
      name: `${label} dialog title and description clear of Close (X13)`,
      cite: "kit web/src/components/ui/dialog.tsx:DialogHeader gap-1.5 flex flex-col; DialogContent showCloseButton Close absolute top-4 right-4",
      property: "text intersecting the Close",
      expected: "none",
      measured: hits.length === 0 ? "none" : hits.join(", "),
      verdict: hits.length === 0 ? "pass" : "fail",
    });
  };

  // The open dialog.
  await openDialog(page);
  expect(await slotsPresent(page, SLOTS_OPEN), "dialog slots rendered while open").toEqual([]);
  for (const e of openRows) await measure(page, mode, e, rows);
  await clearOfClose("1280");
  await page.keyboard.press("Escape");
  await expect(page.locator("[data-slot=dialog-overlay]")).toHaveCount(0);

  // Responsive expectations (§4.7).
  const responsive = async (name: string, selector: string, prop: string, want: number | string) => {
    const s = await cssOf(page, selector, [prop]);
    if (typeof want === "number") {
      const delta = Math.abs(parseFloat(s[prop]) - want);
      rows.push({ name, cite: "§4.7 responsive", property: prop, expected: want, measured: s[prop], delta: Number(delta.toFixed(3)), verdict: delta <= 1 ? "pass" : "fail" });
    } else rows.push({ name, cite: "§4.7 responsive", property: prop, expected: want, measured: s[prop], verdict: s[prop] === want ? "pass" : "fail" });
  };
  const n = RESPONSIVE.narrow;
  const observations: ResizeObservation[] = [];
  observations.push(await resize(page, n.width, n.height));
  await responsive("320 gutter", "#overview [data-slot=container]", "padding-left", n.gutter);
  await responsive("320 H1", "#overview h1", "font-size", n.h1);
  await responsive("320 H1 wrapping", "#overview h1", "overflow-wrap", "anywhere");
  await responsive("320 H2", "#getting-around h2", "font-size", n.h2);
  await openDialog(page);
  await responsive("320 dialog", "[data-slot=dialog-content]", "width", n.dialogWidth);
  const content = await page.locator("[data-slot=dialog-content]").evaluate((el) => {
    const s = getComputedStyle(el);
    return el.clientWidth - parseFloat(s.paddingLeft) - parseFloat(s.paddingRight);
  });
  rows.push({ name: "320 dialog content box", cite: "§4.7 responsive", property: "content width", expected: n.dialogContent, measured: `${content}px`, delta: Math.abs(content - n.dialogContent), verdict: Math.abs(content - n.dialogContent) <= 1 ? "pass" : "fail" });
  await clearOfClose("320");
  await page.keyboard.press("Escape");
  await expect(page.locator("[data-slot=dialog-overlay]")).toHaveCount(0);
  const z = RESPONSIVE.zoomed;
  observations.push(await resize(page, z.width, z.height));
  await responsive("640x400 gutter", "#overview [data-slot=container]", "padding-left", z.gutter);
  await openDialog(page);
  // G069: the kit's DialogContent sets no maximum height and no overflow (the legacy module's max-h and overflow-y-auto,
  // X12, are gone); the help dialog lies wholly inside the 640x400 viewport (top and bottom included).
  await responsive("640x400 dialog max-height (kit: none)", "[data-slot=dialog-content]", "max-height", "none");
  await responsive("640x400 dialog overflow (kit: visible)", "[data-slot=dialog-content]", "overflow-y", "visible");
  const fits = await page.locator("[data-slot=dialog-content]").evaluate((el) => {
    const r = el.getBoundingClientRect();
    return { top: r.top, bottom: r.bottom, inside: r.top >= -0.5 && r.bottom <= window.innerHeight + 0.5 };
  });
  rows.push({
    name: "640x400 dialog inside the viewport (X12)",
    cite: "§4.7 responsive",
    property: "vertical extent",
    expected: `0..${z.height}`,
    measured: `${fits.top.toFixed(1)}..${fits.bottom.toFixed(1)}`,
    verdict: fits.inside ? "pass" : "fail",
  });
  await page.keyboard.press("Escape");

  await testInfo.attach("source-fidelity.json", { body: JSON.stringify(rows, null, 2), contentType: "application/json" });
  await testInfo.attach("responsive-observation.json", { body: JSON.stringify(observations, null, 2), contentType: "application/json" });
  const failed = rows.filter((r) => r.verdict === "fail").map((r) => `${r.name} ${r.property}: expected ${r.expected}, measured ${r.measured} (${r.cite})`);
  expect(failed, "source reference mismatches").toEqual([]);
});

// --- S-FONTS (§7.4) -----------------------------------------------------------------------------------------

type Family = (typeof BUNDLED_FAMILIES)[number];
type FontNode = { name: string; selector: string; family: Family | "fallback"; weight?: string; state?: "skip" | "dialog" | "appearance" };

// Production nodes (§7.4 item 4). G069 (the kit's typography, Plan S12): every heading is the kit's Heading in Oswald
// (H1 700, H2 600, H3 500) as are the header identity (600), the kit TextLink of the skip link (400) and the kit dialog
// title (an h2, the base layer's heading face, 500); Barlow Condensed on the kicker (600) and the meta lines (footer and
// the grid's index numbers, 400); Geist on navigation, buttons, body text and the dialog description. No shell text is
// mono (SHELL_FONT_PATHS: Geist Mono is covered on the fixture page).
const FONT_NODES: FontNode[] = [
  { name: "H1", selector: "#overview h1", family: "Oswald", weight: "700" },
  { name: "header identity", selector: "[data-slot=site-header] [data-slot=container] > p", family: "Oswald", weight: "600" },
  { name: "nav link", selector: "[data-slot=site-header] nav a", family: "Geist", weight: "500" },
  { name: "help trigger", selector: TRIGGER, family: "Geist", weight: "500" },
  { name: "H2", selector: "#getting-around h2", family: "Oswald", weight: "600" },
  { name: "H3", selector: "[data-slot=ruled-grid] h3", family: "Oswald", weight: "500" },
  { name: "kicker", selector: KICKER, family: "Barlow Condensed", weight: "600" },
  // Amendment 03 (the kit Select since G069): the option labels are the items of its popup, in the document body while
  // it is open; the trigger itself shows a symbol, not text.
  { name: "appearance options", selector: `${APPEARANCE_POPUP} [data-slot=select-item] > :first-child`, family: "Geist", weight: "400", state: "appearance" },
  { name: "About bullet", selector: "#about [data-slot=bullet-item] > span:last-child", family: "Geist", weight: "400" },
  { name: "footer meta", selector: "footer [data-slot=text]", family: "Barlow Condensed", weight: "400" },
  { name: "lead", selector: LEAD, family: "Geist", weight: "400" },
  { name: "cell body", selector: CELL_BODY, family: "Geist", weight: "400" },
  { name: "index (meta)", selector: INDEX, family: "Barlow Condensed", weight: "400" },
  { name: "skip link (focused)", selector: 'a[href="#main"]', family: "Oswald", weight: "400", state: "skip" },
  { name: "dialog title", selector: "[data-slot=dialog-title]", family: "Oswald", weight: "500", state: "dialog" },
  { name: "dialog description", selector: "[data-slot=dialog-description]", family: "Geist", weight: "400", state: "dialog" },
];

type NodeResult = { name: string; verdict: RenderedFont["verdict"]; fonts: RenderedFont["fonts"]; weight: string; synthesized: boolean };

async function checkNodes(page: Page, nodes: FontNode[], families: Record<Family, string[]>, out: NodeResult[]) {
  const targets = Object.fromEntries(nodes.map((n) => [n.name, page.locator(n.selector).first()]));
  const fams = Object.fromEntries(nodes.filter((n) => n.family !== "fallback").map((n) => [n.name, families[n.family as Family]]));
  const rendered = await renderedFonts(page, targets, fams);
  for (const n of nodes) {
    const r = rendered.find((x) => x.name === n.name)!;
    const el = targets[n.name];
    const weight = await el.evaluate((e) => getComputedStyle(e).fontWeight);
    const synth = n.family === "fallback" ? false : await synthesized(el);
    out.push({ name: n.name, verdict: r.verdict, fonts: r.fonts, weight, synthesized: synth });
  }
}

test("[S-FONTS] phase B: the five bundled fonts are delivered, declared and rendered without synthesis", async ({ page }, testInfo) => {
  await page.setViewportSize({ width: 1280, height: 800 });
  await page.emulateMedia({ reducedMotion: "reduce" });
  const ready = await openReady(page);
  await page.mouse.move(0, 0);
  // 1. Bytes and 2. binary metadata, from the same responses.
  expectFontBytes(ready.records);
  const meta = expectFontMetadata(ready.bodies);
  const families = familyMap(meta);
  // 3. Loaded faces (B0 already required them): every face the shell requests is loaded; Geist Mono, which no shell
  // text uses, stays unloaded (G069, SHELL_FONT_PATHS).
  const faces = await faceStatuses(page);
  expect(faces.length, "bundled faces").toBe(BUNDLED_FACES);
  expect(faces.map((f) => `${f.family}: ${f.status}`), "loaded faces").toEqual(faces.map((f) => `${f.family}: ${wantedStatus(f, "loaded")}`));
  expect(faces.filter((f) => f.status === "loaded").length, "the four shell faces loaded").toBe(SHELL_FONT_PATHS.length);
  // 4. Rendered fonts and 5. weights and synthesis, per state.
  const results: NodeResult[] = [];
  await checkNodes(page, FONT_NODES.filter((n) => !n.state), families, results);
  // The focused skip link appears at (0, 0): park the pointer so the twin screenshots are not split by hover.
  await parkPointer(page);
  await resetFocus(page);
  await page.keyboard.press("Tab");
  await expect(page.getByRole("link", { name: "Skip to main content" })).toBeFocused();
  await checkNodes(page, FONT_NODES.filter((n) => n.state === "skip"), families, results);
  await resetFocus(page);
  await openDialog(page);
  await checkNodes(page, FONT_NODES.filter((n) => n.state === "dialog"), families, results);
  await page.keyboard.press("Escape");
  await expect(page.locator("[data-slot=dialog-content]")).toBeHidden();
  // The appearance options exist only while the kit Select's popup is open. The popup's open motion (fade and
  // zoom-in-95, 100ms) runs under reduced motion too (the kit has no reduced-motion variant for it), so the options are
  // read once it has settled.
  await appearanceControl(page).click();
  const appearancePopup = page.locator(APPEARANCE_POPUP);
  await expect(appearancePopup).toBeVisible();
  await appearancePopup.evaluate((e) => Promise.allSettled(e.getAnimations({ subtree: true }).map((a) => a.finished)));
  await checkNodes(page, FONT_NODES.filter((n) => n.state === "appearance"), families, results);
  await page.keyboard.press("Escape");
  await expect(appearancePopup).toBeHidden();
  // Exactly the shell's pinned binaries were requested on this page, and nothing else under /fonts/.
  const rec = await recordFontResponses(page);
  expect(rec.responses.map((r) => new URL(r.url()).pathname).sort(), "font requests on /").toEqual([...SHELL_FONT_PATHS].sort());
  await testInfo.attach("fonts.json", {
    body: JSON.stringify({ records: ready.records, metadata: meta, families, faces, nodes: results }, null, 2),
    contentType: "application/json",
  });
  for (const n of FONT_NODES) {
    const r = results.find((x) => x.name === n.name)!;
    expect(r.verdict, `${n.name}: rendered font ${JSON.stringify(r.fonts)}`).toBe(n.family === "fallback" ? "fallback" : "custom");
    if (n.weight) expect(r.weight, `${n.name}: computed weight`).toBe(n.weight);
    expect(r.synthesized, `${n.name}: synthetic bold or oblique`).toBe(false);
  }
});

test("[S-FONTS] control 7: a face missing the requested weight is detected as synthesized", async ({ page }, testInfo) => {
  await page.setViewportSize({ width: 1280, height: 800 });
  await openReady(page);
  const probe = await synthesisProbe(page);
  let detected: boolean;
  try {
    detected = await synthesized(probe);
  } finally {
    await page.evaluate(() => document.getElementById("g048-synthesis-probe")?.remove());
  }
  await testInfo.attach("fonts-control7.json", { body: JSON.stringify({ probe: "Synthesis Probe 400 at 700", detected }, null, 2), contentType: "application/json" });
  expect(detected, "synthetic bold detected by the twin check").toBe(true);
});

const CONTROL6_STEPS = ["held", "hydrated", "faces loading", "fallback detected", "readable", "keyboard", "reflow 1280", "reflow 320", "still held"];

// Focus moves by Tab from the document start to the help trigger within 20 presses.
async function tabToTrigger(page: Page) {
  await resetFocus(page);
  for (let i = 0; i < 20; i++) {
    await page.keyboard.press("Tab");
    if (await page.locator(TRIGGER).evaluate((el) => el === document.activeElement)) return;
  }
  throw new Error("Tab did not reach the help trigger within 20 presses");
}

async function insideAfterFrames(page: Page, dialog: Locator): Promise<boolean> {
  await page.evaluate(() => new Promise<void>((r) => requestAnimationFrame(() => requestAnimationFrame(() => r()))));
  return dialog.evaluate((d) => d.contains(document.activeElement));
}

test("[S-FONTS] control 6: held fonts are detected as fallback while the page stays readable, operable and reflowing", async ({ page }, testInfo) => {
  const mode = modeOf(testInfo);
  const steps: string[] = [];
  const evidence: Record<string, unknown> = {};
  // Each completed step with its duration (ms since the previous step), for the evidence (§7.4).
  const timings: { step: string; ms: number }[] = [];
  let lastMark = Date.now();
  const mark = (step: string) => {
    const now = Date.now();
    steps.push(step);
    timings.push({ step, ms: now - lastMark });
    lastMark = now;
  };
  await page.setViewportSize({ width: 1280, height: 800 });
  const hold = await holdFonts(page);
  try {
    // Phase A only: no screenshot and no wait for a font.
    await page.goto("/", { waitUntil: "domcontentloaded" });
    const held = await expectHeld(hold);
    evidence.held = held;
    mark("held");
    await waitForHydration(page);
    mark("hydrated");
    evidence.facesBefore = await expectFaces(page, "loading");
    mark("faces loading");
    const targets = { H1: page.locator("#overview h1"), kicker: page.locator(KICKER), lead: page.locator(LEAD) };
    const verdicts = await renderedFonts(page, targets);
    evidence.verdicts = verdicts;
    expect(verdicts.map((v) => ({ [v.name]: v.verdict })), "the held fonts are detected").toEqual([{ H1: "fallback" }, { kicker: "fallback" }, { lead: "fallback" }]);
    mark("fallback detected");
    for (const [name, loc, role] of [["H1", targets.H1, "foreground"], ["kicker", targets.kicker, "accent"], ["lead", targets.lead, "foreground/80"]] as const) {
      const r = await loc.evaluate((el) => {
        const range = document.createRange();
        range.selectNodeContents(el);
        const b = range.getBoundingClientRect();
        return { width: b.width, height: b.height, visibility: getComputedStyle(el).visibility, color: getComputedStyle(el).color };
      });
      expect(r.width * r.height, `${name}: rendered text box`).toBeGreaterThan(0);
      expect(r.visibility, `${name}: visible during the swap period`).toBe("visible");
      const c = await roleDelta(page, mode, r.color, role);
      expect(c.delta, `${name}: colour role ${role}: ${r.color} = ${hex(c.measured)} vs ${hex(c.expected)}`).toBeLessThanOrEqual(c.tolerance);
    }
    mark("readable");
    await tabToTrigger(page);
    await page.keyboard.press("Enter");
    const dialog = page.getByRole("dialog");
    await expect(dialog).toBeVisible();
    await expect.poll(() => dialog.evaluate((d) => d.contains(document.activeElement)), { message: "focus moved into the dialog" }).toBe(true);
    for (const key of ["Tab", "Tab", "Shift+Tab", "Shift+Tab"]) {
      await page.keyboard.press(key);
      expect(await insideAfterFrames(page, dialog), `focus stays inside after ${key}`).toBe(true);
    }
    await page.keyboard.press("Escape");
    await expect(dialog).toBeHidden();
    await expect(page.locator(TRIGGER)).toBeFocused();
    mark("keyboard");
    const wide = await reflowLayout(page);
    const wideHeader = await reflowLayout(page, "[data-slot=site-header]");
    expect(wide.scrollWidth, "1280: no horizontal overflow").toBeLessThanOrEqual(wide.clientWidth);
    expect(wideHeader.overlaps, "1280: header controls do not overlap").toEqual([]);
    mark("reflow 1280");
    await resize(page, 320, 640);
    const narrow = await reflowLayout(page);
    const narrowHeader = await reflowLayout(page, "[data-slot=site-header]");
    expect(narrow.scrollWidth, "320: no horizontal overflow").toBeLessThanOrEqual(narrow.clientWidth);
    expect(narrowHeader.overlaps, "320: header controls do not overlap").toEqual([]);
    await page.locator(TRIGGER).click();
    await expect(dialog).toBeVisible();
    await dialogSettled(page);
    const inside = await page.locator("[data-slot=dialog-content]").evaluate((d) => {
      const r = d.getBoundingClientRect();
      return r.left >= -0.5 && r.top >= -0.5 && r.right <= window.innerWidth + 0.5 && r.bottom <= window.innerHeight + 0.5;
    });
    expect(inside, "320: the dialog lies inside the viewport").toBe(true);
    await expect(dialog.getByRole("button", { name: "Close" })).toBeInViewport();
    await page.keyboard.press("Escape");
    await expect(dialog).toBeHidden();
    evidence.reflow = { wide, wideHeader, narrow, narrowHeader };
    mark("reflow 320");
    const facesAfter = await faceStatuses(page);
    evidence.facesAfter = facesAfter;
    expect(facesAfter.length, "bundled faces").toBe(BUNDLED_FACES);
    expect(facesAfter.map((f) => `${f.family}: ${f.status}`), "faces still loading at the end").toEqual(facesAfter.map((f) => `${f.family}: ${wantedStatus(f, "loading")}`));
    expect([...new Set(hold.held())].sort(), "held set unchanged").toEqual(held);
    expect(hold.completed(), "no font response completed while held").toEqual([]);
    mark("still held");
  } finally {
    await hold.release();
    await testInfo.attach("fonts-control6.json", { body: JSON.stringify({ steps, timings, evidence, releaseErrors: hold.releaseErrors() }, null, 2), contentType: "application/json" });
  }
  expect(hold.releaseErrors(), "release errors").toEqual([]);
  expect(steps, "every control step completed").toEqual(CONTROL6_STEPS);
});

type DialogState = { open: boolean; closeFocused: boolean; inside: boolean; headerOverlaps: string[]; horizontalOverflow: boolean };

async function dialogState(page: Page): Promise<DialogState> {
  const s = await page.evaluate(() => {
    const d = document.querySelector<HTMLElement>("[data-slot=dialog-content]");
    // G069: the kit DialogContent's own corner Close (showCloseButton; an sr-only name, no aria-label attribute).
    const close = d?.querySelector(':scope > [data-slot="dialog-close"]') ?? null;
    const r = d?.getBoundingClientRect();
    return {
      open: !!d,
      closeFocused: !!close && document.activeElement === close,
      inside: !!r && r.left >= -0.5 && r.top >= -0.5 && r.right <= window.innerWidth + 0.5 && r.bottom <= window.innerHeight + 0.5,
    };
  });
  const header = await reflowLayout(page, "[data-slot=site-header]");
  return { ...s, headerOverlaps: header.overlaps, horizontalOverflow: header.scrollWidth > header.clientWidth };
}

const STABLE: DialogState = { open: true, closeFocused: true, inside: true, headerOverlaps: [], horizontalOverflow: false };

const SWAP_STEPS = [
  "held",
  "hydrated",
  "faces loading",
  "dialog open, Close focused",
  "fallback before release",
  "stable before release",
  "released",
  "fonts ready",
  "custom after release",
  "stable after release",
];

for (const vp of [
  { width: 1280, height: 800 },
  { width: 320, height: 640 },
]) {
  test(`[S-FONTS] swap stability ${vp.width}x${vp.height}: focus, dialog and layout survive the fonts arriving`, async ({ page }, testInfo) => {
    await page.setViewportSize(vp);
    const rec = await recordFontResponses(page);
    const hold = await holdFonts(page);
    const evidence: Record<string, unknown> = {};
    const steps: string[] = [];
    const timings: { step: string; ms: number }[] = [];
    const started = Date.now();
    let lastMark = started;
    const mark = (step: string) => {
      const now = Date.now();
      steps.push(step);
      timings.push({ step, ms: now - lastMark });
      lastMark = now;
    };
    try {
      await page.goto("/", { waitUntil: "domcontentloaded" });
      evidence.held = await expectHeld(hold);
      mark("held");
      await waitForHydration(page);
      mark("hydrated");
      evidence.facesBefore = await expectFaces(page, "loading");
      mark("faces loading");
      const dialog = await openDialog(page);
      const close = dialog.getByRole("button", { name: "Close" });
      await close.focus();
      mark("dialog open, Close focused");
      const title = { "dialog title": page.locator("[data-slot=dialog-title]") };
      const before = await renderedFonts(page, title);
      evidence.before = before;
      expect(before[0].verdict, "the dialog title is in a fallback before release (a swap really happens)").toBe("fallback");
      mark("fallback before release");
      const stateBefore = await dialogState(page);
      evidence.stateBefore = stateBefore;
      expect(stateBefore, "state before release").toEqual(STABLE);
      mark("stable before release");
      await hold.release();
      evidence.release = { atMs: Date.now() - started, errors: hold.releaseErrors() };
      mark("released");
      const ready = await fontsReady(page, rec);
      expectFontBytes(ready.records);
      const families = familyMap(expectFontMetadata(ready.bodies));
      evidence.records = ready.records;
      evidence.facesAfter = await faceStatuses(page);
      mark("fonts ready");
      // G069: the kit dialog title is an h2 in the kit's heading face.
      const after = await renderedFonts(page, title, { "dialog title": families.Oswald });
      evidence.after = after;
      expect(after[0].verdict, `the dialog title is Oswald after release: ${JSON.stringify(after[0].fonts)}`).toBe("custom");
      mark("custom after release");
      const stateAfter = await dialogState(page);
      evidence.stateAfter = stateAfter;
      expect(stateAfter, "state after the swap").toEqual(STABLE);
      mark("stable after release");
      await page.keyboard.press("Escape");
    } finally {
      await hold.release();
      await testInfo.attach(`fonts-swap-${vp.width}.json`, { body: JSON.stringify({ steps, timings, evidence, releaseErrors: hold.releaseErrors() }, null, 2), contentType: "application/json" });
    }
    expect(hold.releaseErrors(), "release errors").toEqual([]);
    expect(steps, "every swap step completed").toEqual(SWAP_STEPS);
  });
}

// --- S-BLUEPRINT (§8.5) --------------------------------------------------------------------------------------

type Clip = { x: number; y: number; width: number; height: number };

// The layer box intersected with the viewport, in whole pixels inside both.
async function layerClip(page: Page): Promise<Clip> {
  await page.locator(LAYER).scrollIntoViewIfNeeded();
  const r = await page.locator(LAYER).evaluate((el) => {
    const b = el.getBoundingClientRect();
    return { left: b.left, top: b.top, right: b.right, bottom: b.bottom, vw: window.innerWidth, vh: window.innerHeight };
  });
  const x = Math.ceil(Math.max(0, r.left));
  const y = Math.ceil(Math.max(0, r.top));
  const right = Math.floor(Math.min(r.vw, r.right));
  const bottom = Math.floor(Math.min(r.vh, r.bottom));
  expect(right - x, "layer clip width").toBeGreaterThan(0);
  expect(bottom - y, "layer clip height").toBeGreaterThan(0);
  return { x, y, width: right - x, height: bottom - y };
}

async function overviewBoxes(page: Page) {
  return page.evaluate(() => {
    const overview = document.querySelector<HTMLElement>("#overview")!;
    const container = overview.querySelector("[data-slot=container]")!;
    const children = [...container.children].filter((c) => !c.hasAttribute("data-blueprint")).map((c) => {
      const r = c.getBoundingClientRect();
      return [r.x, r.y + window.scrollY, r.width, r.height].map((v) => Math.round(v * 100) / 100);
    });
    return { section: overview.getBoundingClientRect().height, children };
  });
}

function differing(a: Image, b: Image, threshold = 8): number {
  let n = 0;
  for (let i = 0; i < a.pixels.length; i++) if (distance(a.pixels[i], b.pixels[i]) > threshold) n++;
  return n;
}

// Progress of a pixel from the ground toward the expected stroke composite (projection, 0 = ground, 1 = stroke).
function toward(p: RGB, ground: RGB, stroke: RGB): number {
  const d = [0, 1, 2].map((i) => stroke[i] - ground[i]);
  const len = d.reduce((s, v) => s + v * v, 0);
  return len === 0 ? 0 : [0, 1, 2].reduce((s, i) => s + (p[i] - ground[i]) * d[i], 0) / len;
}

const GROUPS = ["construction", "dimensions", "structure"] as const;

for (const vp of [
  { name: "1280x800", width: 1280, height: 800 },
  { name: "320x640", width: 320, height: 640 },
]) {
  test(`[S-BLUEPRINT] ${vp.name}: decorative, same-origin, painted in the scope's role, absent without the asset or under forced colours`, async ({ page }, testInfo) => {
    const mode = modeOf(testInfo);
    const ground = hexToRgb(SOURCES[mode].background);
    const stroke = over([...hexToRgb(SOURCES[mode]["muted-foreground"]), Math.round(0.5 * 255)] as RGBA, ground);
    const evidence: Record<string, unknown> = { viewport: vp, mode, ground: hex(ground), stroke: hex(stroke) };
    await page.setViewportSize({ width: vp.width, height: vp.height });
    await page.emulateMedia({ reducedMotion: "reduce" });
    // 2. The response: same origin, 200, image/svg+xml and the pinned bytes.
    const svgResponse = page.waitForResponse((r) => new URL(r.url()).pathname === BLUEPRINT_PATH, { timeout: BOUND.response });
    await openReady(page);
    const response = await svgResponse;
    const body = await response.body();
    evidence.response = {
      status: response.status(),
      contentType: await response.headerValue("content-type"),
      bytes: body.length,
      sha256: createHash("sha256").update(body).digest("hex"),
    };
    expect(new URL(response.url()).origin, "same origin").toBe(new URL(page.url()).origin);
    expect(response.status()).toBe(200);
    expect((await response.headerValue("content-type")) ?? "", "content type").toMatch(/^image\/svg\+xml/);
    expect(createHash("sha256").update(body).digest("hex"), "SVG SHA-256").toBe(BLUEPRINT_SHA256);
    await page.mouse.move(0, 0);

    // 1. Decorative semantics.
    const layer = page.locator(LAYER);
    await expect(layer).toHaveCount(1);
    await layer.scrollIntoViewIfNeeded();
    const facts = await layer.evaluate((el) => {
      const s = getComputedStyle(el);
      const r = el.getBoundingClientRect();
      const hit = document.elementFromPoint(r.left + r.width / 2, Math.min(r.top + r.height / 2, window.innerHeight - 1));
      return {
        ariaHidden: el.getAttribute("aria-hidden"),
        tabIndex: (el as HTMLElement).tabIndex,
        tabindexAttribute: el.hasAttribute("tabindex"),
        role: el.getAttribute("role"),
        text: el.textContent,
        children: el.childElementCount,
        pointerEvents: s.pointerEvents,
        animations: el.getAnimations().length,
        centreHitIsLayer: hit === el,
      };
    });
    evidence.semantics = facts;
    expect(facts).toEqual({ ariaHidden: "true", tabIndex: -1, tabindexAttribute: false, role: null, text: "", children: 0, pointerEvents: "none", animations: 0, centreHitIsLayer: false });
    const shown = await overviewBoxes(page);
    const hide = await page.addStyleTag({ content: "[data-blueprint] { display: none !important; }" });
    const without = await overviewBoxes(page);
    await hide.evaluate((el) => (el as Element).remove());
    expect(without, "the layer changes no content box or the section height").toEqual(shown);

    // 3. Imagery on versus off, over the layer clip with no text in it.
    const clip = await layerClip(page);
    const textInClip = await page.evaluate((c) => {
      const hits: string[] = [];
      const walker = document.createTreeWalker(document.querySelector("#overview")!, NodeFilter.SHOW_TEXT);
      for (let n = walker.nextNode(); n; n = walker.nextNode()) {
        if (!n.textContent?.trim()) continue;
        const range = document.createRange();
        range.selectNodeContents(n);
        for (const r of range.getClientRects()) {
          if (r.left < c.x + c.width && r.right > c.x && r.top < c.y + c.height && r.bottom > c.y) hits.push(n.textContent.trim().slice(0, 30));
        }
      }
      return hits;
    }, clip);
    expect(textInClip, "no text intersects the imagery clip with the default identity").toEqual([]);
    const visible = decodePng(await page.screenshot({ clip, animations: "disabled" }));
    const off = await page.addStyleTag({ content: "[data-blueprint] { visibility: hidden !important; }" });
    const hidden = decodePng(await page.screenshot({ clip, animations: "disabled" }));
    await off.evaluate((el) => (el as Element).remove());
    const area = clip.width * clip.height;
    const changed = differing(visible, hidden);
    const offGround = hidden.pixels.filter((p) => distance(p, ground) > 2).length;
    const strokePixels = visible.pixels.filter((p) => toward(p, ground, stroke) >= 0.5).length;
    evidence.imagery = { clip, area, changed, offGround, strokePixels };
    expect(offGround, "with the layer hidden the clip is the section ground (the comparator measures paint)").toBe(0);
    expect(changed, "pixels changed by the imagery").toBeGreaterThanOrEqual(Math.max(400, 0.005 * area));
    expect(strokePixels, "pixels at least half way to the structure-stroke composite").toBeGreaterThanOrEqual(200);
    await testInfo.attach(`blueprint-clip-${vp.width}.png`, { body: await page.screenshot({ clip, animations: "disabled" }), contentType: "image/png" });

    // 4. Motif evidence: the SVG as a document, each group's removal visible.
    if (vp.width === 1280) {
      const doc = await page.context().newPage();
      try {
        await doc.setViewportSize({ width: 960, height: 640 });
        await doc.goto(BLUEPRINT_PATH);
        const whole = await doc.screenshot({ animations: "disabled" });
        await testInfo.attach("blueprint-motifs.png", { body: whole, contentType: "image/png" });
        const motif: Record<string, number> = {};
        for (const id of GROUPS) {
          await doc.evaluate((g) => {
            (document.getElementById(g) as unknown as SVGElement).style.display = "none";
          }, id);
          const less = await doc.screenshot({ animations: "disabled" });
          await doc.evaluate((g) => {
            (document.getElementById(g) as unknown as SVGElement).style.display = "";
          }, id);
          motif[id] = differing(decodePng(whole), decodePng(less));
          await testInfo.attach(`blueprint-without-${id}.png`, { body: less, contentType: "image/png" });
        }
        evidence.motif = motif;
        for (const id of GROUPS) expect(motif[id], `removing ${id} changes the drawing`).toBeGreaterThanOrEqual(300);
      } finally {
        await doc.close();
      }
    }

    // 5. Missing asset: a separate page whose /blueprint/ requests are held paints only the ground.
    const held: Route[] = [];
    const missing = await page.context().newPage();
    try {
      await missing.setViewportSize({ width: vp.width, height: vp.height });
      await missing.emulateMedia({ reducedMotion: "reduce" });
      await missing.route("**/blueprint/**", (route) => {
        held.push(route);
      });
      // Held images can delay the load event, so wait for the document only.
      await missing.goto("/", { waitUntil: "domcontentloaded" });
      await waitForHydration(missing);
      await expectFaces(missing, "loaded");
      await expect.poll(() => held.length, { message: "the blueprint request is held", timeout: BOUND.holdCapture }).toBeGreaterThan(0);
      expect(await overviewBoxes(missing), "content boxes without the asset").toEqual(shown);
      const missingClip = await layerClip(missing);
      const bare = decodePng(await missing.screenshot({ clip: missingClip, animations: "disabled" }));
      const bareOff = bare.pixels.filter((p) => distance(p, ground) > 2).length;
      evidence.missing = { held: held.length, clip: missingClip, offGround: bareOff };
      expect(bareOff, "without the asset the layer paints nothing").toBe(0);
      await missing.locator(TRIGGER).click();
      await expect(missing.getByRole("dialog")).toBeVisible();
      await missing.keyboard.press("Escape");
      await expect(missing.getByRole("dialog")).toBeHidden();
    } finally {
      for (const route of held.splice(0)) await route.continue().catch(() => undefined);
      await missing.close();
    }

    // 6. Forced colours: hidden layer, visible text, and the kit's own keyboard focus state. The kit Button draws its
    // focus as a ring-coloured border with a ring-3 halo and sets outline-none (outline-style none), and the app adds no
    // forced-colours outline of its own (G069: the kit's focus as shipped), so the keyboard still reaches the trigger,
    // which takes the kit's :focus-visible state with no outline.
    await page.emulateMedia({ forcedColors: "active" });
    // Let the emulated media reach style before computed values are read.
    await page.evaluate(() => new Promise<void>((r) => requestAnimationFrame(() => requestAnimationFrame(() => r()))));
    try {
      expect(await layer.evaluate((el) => getComputedStyle(el).display), "forced colours hide the layer").toBe("none");
      for (const sel of [KICKER, "#overview h1", LEAD]) await expect(page.locator(sel)).toBeVisible();
      await tabToTrigger(page);
      const focus = await page.locator(TRIGGER).evaluate((el) => ({ focusVisible: el.matches(":focus-visible"), style: getComputedStyle(el).outlineStyle }));
      evidence.forcedColours = focus;
      expect(focus.focusVisible, "keyboard focus on the trigger takes the kit's :focus-visible state under forced colours").toBe(true);
      expect(focus.style, "the kit Button's outline-none under forced colours (no app outline rule)").toBe("none");
    } finally {
      await page.emulateMedia({ forcedColors: "none" });
      await resetFocus(page);
    }

    // 8. Overview clip evidence.
    await page.evaluate(() => window.scrollTo(0, 0));
    await page.mouse.move(0, 0);
    await testInfo.attach(`blueprint-overview-${vp.width}.png`, { body: await page.locator("#overview").screenshot({ animations: "disabled" }), contentType: "image/png" });
    await testInfo.attach(`blueprint-${vp.width}.json`, { body: JSON.stringify(evidence, null, 2), contentType: "application/json" });
  });
}

// G069 S24 (retyping G067 D-3): the kit's angled call to action (CtaButton, a fixed white label on --vx-blue) renders
// only where the platform palette is guaranteed, the root-host landing (its probe runs on the fixture page, tokens.spec
// S-ANGLE and the CtaButton S-STATES). The sign-in page is served on tenant hosts, so its primary action is the kit
// Button at size lg: no angled element on the page, the action neither clipped nor carrying a clipped ::before layer,
// its label inside its box, and its keyboard focus the kit's whole state (the ring-coloured border with its halo outside
// it) present on all four sides at >= 3:1 against the page. The kit's angle token stays the source polygon.
test("[S-ANGLE] sign-in: the primary action is the kit Button, unclipped with a whole focus state, and no angled call to action at 1280 and 320", async ({ page }) => {
  for (const [w, h] of [[1280, 800], [320, 640]]) {
    await page.setViewportSize({ width: w, height: h });
    await page.goto("/sign-in");
    await waitForHydration(page);
    const submit = page.locator(SIGN_IN_SUBMIT);
    await expect(submit).toHaveCount(1);
    await expect(submit).toHaveAccessibleName("Sign in");
    await expect(submit, `sign-in at ${w}: the kit Button`).toHaveAttribute("data-slot", "button");
    await expect(page.locator('.vx-angle, [data-slot="cta-button"]'), `sign-in at ${w}: no angled call to action`).toHaveCount(0);
    const angle = await page.evaluate(() => getComputedStyle(document.documentElement).getPropertyValue("--vx-angle").trim());
    expect(polygonPoints(angle, 160, 32), `sign-in at ${w}: --vx-angle is the source polygon`).toEqual(polygonPoints(SOURCE_ANGLE, 160, 32));
    const geo = await submit.evaluate((el) => {
      const b = el.getBoundingClientRect();
      const range = document.createRange();
      range.selectNodeContents(el);
      const t = range.getBoundingClientRect();
      return { own: getComputedStyle(el).clipPath, before: getComputedStyle(el, "::before").clipPath, height: b.height, b: { l: b.left, t: b.top, r: b.right, btm: b.bottom }, t: { l: t.left, t: t.top, r: t.right, btm: t.bottom } };
    });
    expect([geo.own, geo.before], `sign-in at ${w}: neither the action nor a ::before layer is clipped`).toEqual(["none", "none"]);
    expect(Math.abs(geo.height - 36), `sign-in at ${w}: the kit Button size lg (h-9, 36px): ${geo.height}`).toBeLessThanOrEqual(0.5);
    expect(
      geo.t.l >= geo.b.l - 0.5 && geo.t.t >= geo.b.t - 0.5 && geo.t.r <= geo.b.r + 0.5 && geo.t.btm <= geo.b.btm + 0.5,
      `sign-in at ${w}: label ${JSON.stringify(geo.t)} inside the box ${JSON.stringify(geo.b)}`,
    ).toBe(true);
    await submit.focus();
    await page.keyboard.press("Shift+Tab");
    await page.keyboard.press("Tab");
    await expect(submit).toBeFocused();
    await settled(submit);
    // The halo is sampled up to its spread + 5px outside each edge: centre the action so all four sides are on screen.
    await submit.evaluate((el) => el.scrollIntoView({ block: "center", inline: "center", behavior: "instant" }));
    await page.evaluate(() => new Promise<void>((r) => requestAnimationFrame(() => requestAnimationFrame(() => r()))));
    const st = await submit.evaluate((el) => ({ shadow: getComputedStyle(el).boxShadow, border: getComputedStyle(el).borderTopColor, outline: getComputedStyle(el).outlineStyle }));
    expect(st.outline, `sign-in at ${w}: the kit focus state draws no outline`).toBe("none");
    const halo = haloOf(st.shadow);
    expect(halo, `sign-in at ${w}: focus ring (${st.shadow})`).not.toBeNull();
    const border = await cssToRgba(page, st.border);
    const samples = await compositeSamples(page, submit, halo!.spread);
    expect(samples.map((s) => s.side).sort(), `sign-in at ${w}: focus state sampled on all sides`).toEqual(["bottom", "left", "right", "top"]);
    for (const s of samples) {
      expect(distance(s.border, over(border, s.page)), `sign-in at ${w}: focus border present on the ${s.side}`).toBeLessThanOrEqual(40);
      expect(contrast(s.border, s.page), `sign-in at ${w}: focus border ${hex(s.border)} against the page ${hex(s.page)} on the ${s.side}`).toBeGreaterThanOrEqual(3);
    }
    await resetFocus(page);
  }
});

// A form page of the shell projects (G069 Plan §6.2/§9.3: S-SCREENSHOTS extended with one form page): the test-only My
// Account fixture of the fixture build (account.spec.ts S-ACCOUNT-MCP-UNAVAILABLE), the real account sections with their
// native sign-out forms on the kit Button.
const FORM_PAGE_URL = "http://127.0.0.1:3001/account-fixtures";

test("[S-SCREENSHOTS] shell, overview imagery, open dialog, sign-in and a form page at desktop and 320px (visual evidence)", async ({ page }, testInfo) => {
  for (const [w, h] of [[1280, 800], [320, 640]]) {
    await page.setViewportSize({ width: w, height: h });
    // Captured with the bundled fonts loaded (B0) and the pointer away from every control.
    await openReady(page);
    await page.mouse.move(0, 0);
    await page.evaluate(() => (document.activeElement as HTMLElement | null)?.blur());
    await testInfo.attach(`shell-${w}.png`, { body: await page.screenshot({ fullPage: true, animations: "disabled" }), contentType: "image/png" });
    await testInfo.attach(`shell-${w}-overview.png`, { body: await page.locator("#overview").screenshot({ animations: "disabled" }), contentType: "image/png" });
    await openDialog(page);
    await expect(page.locator("[data-slot=dialog-content]")).toHaveCSS("opacity", "1");
    await testInfo.attach(`shell-${w}-dialog.png`, { body: await page.screenshot({ animations: "disabled" }), contentType: "image/png" });
    await page.keyboard.press("Escape");
    // /sign-in (Plan §7.1 row 3): the brand H1 and the kit Button primary action (S24). That page uses a subset of the
    // bundled faces, so it waits for the faces it uses (document.fonts.ready) rather than the B0 gate of /.
    await page.goto("/sign-in");
    await waitForHydration(page);
    expect(await page.evaluate(() => document.fonts.ready.then(() => document.fonts.status)), "sign-in fonts ready").toBe("loaded");
    await expect(page.locator(SIGN_IN_SUBMIT)).toBeVisible();
    await page.mouse.move(0, 0);
    await page.evaluate(() => (document.activeElement as HTMLElement | null)?.blur());
    await testInfo.attach(`sign-in-${w}.png`, { body: await page.screenshot({ fullPage: true, animations: "disabled" }), contentType: "image/png" });
    // The form page, likewise waiting for the faces it uses.
    const form = await page.goto(FORM_PAGE_URL);
    expect(form?.status(), "form page").toBe(200);
    await expect(page.getByRole("heading", { level: 1, name: "Your account" })).toBeVisible();
    await expect(page.locator("#sign-out form")).toHaveCount(3);
    expect(await page.evaluate(() => document.fonts.ready.then(() => document.fonts.status)), "form page fonts ready").toBe("loaded");
    await page.mouse.move(0, 0);
    await page.evaluate(() => (document.activeElement as HTMLElement | null)?.blur());
    await testInfo.attach(`form-${w}.png`, { body: await page.screenshot({ fullPage: true, animations: "disabled" }), contentType: "image/png" });
  }
  expect(testInfo.attachments.length).toBe(10);
});
