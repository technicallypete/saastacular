import { expect, test, type Locator, type Page } from "@playwright/test";

import { computedColor, contrast, cssToRgba, distance, hex, hexToRgb, over, scaled } from "./color";
import { decodePng, pixelAt, type RGB } from "./png";

type Box = { x: number; y: number; width: number; height: number };

export async function box(locator: Locator): Promise<Box> {
  await locator.scrollIntoViewIfNeeded();
  const b = await locator.boundingBox();
  if (!b) throw new Error("element has no box");
  return b;
}

// Rendered fill inside a control: 3px below the top edge, inside the 1px border and the padding box, above
// the label; at 60% of the width, clear of the rounded corners (G067: 18px radius on 24-36px controls) and
// inside the angled silhouette of the cta variant (whose top edge spans 15%-100% of the width).
export async function fillOf(page: Page, control: Locator): Promise<RGB> {
  const b = await box(control);
  return pixelAt(page, b.x + b.width * 0.6, b.y + 3);
}

// Rendered surface just outside a text block (inside its container's padding).
export async function surfaceBeside(page: Page, el: Locator): Promise<RGB> {
  const b = await box(el);
  return pixelAt(page, b.x - 6, b.y + Math.min(b.height / 2, 8));
}

// The brightness() factors of the computed `filter` on an element and on each of its ancestors, innermost first (a
// filter applies to its element's whole painted subtree). Only brightness() is modelled; any other filter function
// fails, because the painted text colour could not be derived from it.
export async function brightnessFactors(el: Locator): Promise<number[]> {
  const filters = await el.evaluate((e) => {
    const out: string[] = [];
    for (let n: Element | null = e; n; n = n.parentElement) {
      const f = getComputedStyle(n).filter.trim();
      if (f && f !== "none") out.push(f);
    }
    return out;
  });
  return filters.flatMap((f) =>
    f.split(/\s+(?=[a-z-]+\()/).map((fn) => {
      const m = fn.match(/^brightness\(([\d.]+)(%?)\)$/);
      if (!m) throw new Error(`brightnessFactors: unmodelled filter "${f}"`);
      return Number(m[1]) / (m[2] ? 100 : 1);
    }),
  );
}

// Text colour as painted. CSS `filter` never changes the computed `color`, while the backing is sampled from
// rendered pixels and so already carries any filter: the computed colour is therefore scaled by every brightness()
// factor on the element and its ancestors (`scaled`, core's `scale` model), then composited over the rendered backing
// if translucent. Without a filter (no factor) this is the computed colour, as before.
export async function textColor(page: Page, el: Locator, backing: RGB): Promise<RGB> {
  const [r, g, b, a] = await cssToRgba(page, await computedColor(el));
  const [sr, sg, sb] = (await brightnessFactors(el)).reduce<RGB>((c, f) => scaled(c, f), [r, g, b]);
  return over([sr, sg, sb, a], backing);
}

// `waived` is set only on an OWNER-WAIVED pair (below): its WCAG threshold and the kit's own ratio, which `min` holds.
export type PairResult = { name: string; ratio: number; fg: string; bg: string; min: number; waived?: { threshold: number; kit: number } };

export async function expectText(page: Page, name: string, el: Locator, backing: RGB, min = 4.5): Promise<PairResult> {
  const fg = await textColor(page, el, backing);
  const ratio = contrast(fg, backing);
  const result = { name, ratio: Number(ratio.toFixed(2)), fg: hex(fg), bg: hex(backing), min };
  expect(ratio, `${name}: ${result.fg} on ${result.bg}`).toBeGreaterThanOrEqual(min);
  return result;
}

export function expectNonText(name: string, a: RGB, b: RGB, min = 3): PairResult {
  const ratio = contrast(a, b);
  const result = { name, ratio: Number(ratio.toFixed(2)), fg: hex(a), bg: hex(b), min };
  expect(ratio, `${name}: ${result.fg} against ${result.bg}`).toBeGreaterThanOrEqual(min);
  return result;
}

// --- G069 owner contrast waiver ---------------------------------------------------------------------------------------
// The owner directed the exact kit, its colour overrides included, and waived the contrast shortfalls of the kit's own
// treatment and defaults (management/v0/results/G069-exact-kit-owner-decision-20261026.md and
// G069-owner-kit-fidelity-delta-20261026.md). Such a pair is recorded as OWNER-WAIVED with its measured ratio, never as
// a threshold pass: it must keep at least the kit's own ratio (KIT_FLOOR_TOLERANCE absorbs the 8-bit rounding of a
// painted or composited colour; Core holds tenant palettes to the same kit floor, core/internal/branding/colors.go) and
// stay below its threshold, otherwise it is no longer a waived shortfall and the ordinary threshold assertion applies.
// Every other pair keeps its threshold.
export const KIT_FLOOR_TOLERANCE = 0.03;

export type KitWaiver = { name: string; kit: number; threshold: number };

// The kit's waived pairs these specs measure. Ratios from the kit's sRGB values (core/internal/branding/colors.go): the
// primary label pairs as Core pins them (branding_test.go TestDefaultsPass, T8); the field fills computed from the same
// values, the kit's bg-input/50 composited in sRGB over the kit background (Light (243.7, 241.5, 241.0) on #ffffff;
// Dark #ffffff at 7.5 % over #0c0a09, (30.2, 28.4, 27.5)); the invalid field's value from the kit destructive #e7000b
// (its oklch clipped to sRGB, as painted) on that Light fill. The Dark invalid value (#ff6467 on the Dark fill, about
// 5.88:1) keeps 4.5:1.
export const KIT_WAIVERS: Readonly<Record<"primaryLabel" | "primaryLabelLightHover" | "fieldFillLight" | "fieldFillDark" | "invalidFieldValueLight", KitWaiver>> = {
  // White on --vx-blue #0075ff (the default Button and Badge), at rest and on focus, in both appearances.
  primaryLabel: { name: "T8 white label on the kit primary #0075ff", kit: 4.2097, threshold: 4.5 },
  // The default Button's hover:bg-primary/80 over the Light page (the Dark hover over the Dark page keeps 4.5:1).
  primaryLabelLightHover: { name: "T8 white label on the kit primary's 80 % hover over the Light page", kit: 3.1768, threshold: 4.5 },
  // A kit field with no boundary (border-transparent, the Select trigger): its bg-input/50 fill against the page.
  fieldFillLight: { name: "kit field fill (bg-input/50, no boundary) against the Light page", kit: 1.116, threshold: 3 },
  fieldFillDark: { name: "kit field fill (bg-input/50, no boundary) against the Dark page", kit: 1.164, threshold: 3 },
  // A kit Field marked invalid (data-invalid: text-destructive) passes that colour to its Input's value text (the
  // preflight's color: inherit), on the kit's bg-input/50 fill over the Light page.
  invalidFieldValueLight: { name: "kit invalid Field's input text (text-destructive) on its bg-input/50 fill over the Light page", kit: 4.274, threshold: 4.5 },
};

function waivedPair(w: KitWaiver, what: string, ratio: number, fg: RGB, bg: RGB): PairResult {
  const floor = w.kit - KIT_FLOOR_TOLERANCE;
  const result = { name: `OWNER-WAIVED ${what} (${w.name})`, ratio: Number(ratio.toFixed(2)), fg: hex(fg), bg: hex(bg), min: Number(floor.toFixed(2)), waived: { threshold: w.threshold, kit: w.kit } };
  expect(ratio, `${result.name}: ${result.fg} on ${result.bg} keeps at least the kit's own ${w.kit}:1`).toBeGreaterThanOrEqual(floor);
  expect(ratio, `${result.name}: ${result.fg} on ${result.bg} is below ${w.threshold}:1, so it is the owner-waived shortfall (assert the threshold if it is met)`).toBeLessThan(w.threshold);
  return result;
}

// An OWNER-WAIVED text pair: the element's painted text colour (textColor) on its rendered backing.
export async function expectWaivedText(page: Page, what: string, el: Locator, backing: RGB, w: KitWaiver): Promise<PairResult> {
  const fg = await textColor(page, el, backing);
  return waivedPair(w, what, contrast(fg, backing), fg, backing);
}

// An OWNER-WAIVED non-text pair of two painted colours.
export function expectWaivedNonText(what: string, a: RGB, b: RGB, w: KitWaiver): PairResult {
  return waivedPair(w, what, contrast(a, b), a, b);
}

// axe's color-contrast rule on the app's pages meets the waived T8 label pair wherever the kit primary renders without a
// tenant palette (every default Button and Badge, the shell's Keyboard help and Account included), its Light hover
// when the pointer rests on such a control, and an invalid kit Field's input text on the Light kit palette (the signup
// field errors). Each is matched by axe's own reported colours (within 2 per channel of the kit's sRGB values) and must
// keep at least the kit's ratio (axe reports it truncated to two decimals). Only these, and the tenant invalid-field
// pair below, are waived: another rule, another colour pair or a lower ratio stays a finding.
const AXE_WAIVED: readonly (KitWaiver & { fg: RGB; bg: RGB })[] = [
  { ...KIT_WAIVERS.primaryLabel, fg: [255, 255, 255], bg: [0, 117, 255] },
  { ...KIT_WAIVERS.primaryLabelLightHover, fg: [255, 255, 255], bg: [51, 145, 255] },
  { ...KIT_WAIVERS.invalidFieldValueLight, fg: [231, 0, 11], bg: [244, 242, 241] },
];

// The kit's invalid-field treatment on a tenant palette (owner decision, 2026-10-26): an invalid kit Field's input value
// is painted text-destructive on the tenant's bg-input/50 fill, a pair Core's contrast validator does not hold (about
// 3.15:1 Light and 3.37:1 Dark on the seeded org A palette). Waived only for an axe color-contrast node on an input or
// textarea marked aria-invalid="true", and only from 3:1 up to 4.5:1; any lower ratio or any other element stays a finding.
export const TENANT_INVALID_FIELD_WAIVER: KitWaiver = { name: "tenant palette: kit invalid Field's input text (text-destructive) on its bg-input/50 fill", kit: 3, threshold: 4.5 };

type AxeCheck = { id: string; data?: unknown };
type AxeNode = { target: unknown[]; html?: string; any: AxeCheck[] };
type AxeViolation = { id: string; nodes: AxeNode[] };
export type AxeWaiver = { pair: string; target: string; fg: string; bg: string; ratio: number; kit: number; threshold: number };

function invalidFieldNode(n: AxeNode): boolean {
  const open = /^<(input|textarea)\b[^>]*>/i.exec(n.html ?? "");
  return open !== null && /\saria-invalid="true"/.test(open[0]);
}

// The OWNER-WAIVED kit pair one axe color-contrast node reports, if it is one (by colours and ratio, or for a tenant
// palette's invalid field by its element and ratio).
function axeWaiver(v: AxeViolation, n: AxeNode, target: string): AxeWaiver | undefined {
  if (v.id !== "color-contrast") return undefined;
  const d = n.any.find((c) => c.id === "color-contrast")?.data as { fgColor?: unknown; bgColor?: unknown; contrastRatio?: unknown } | undefined;
  if (d === undefined || typeof d.fgColor !== "string" || typeof d.bgColor !== "string" || typeof d.contrastRatio !== "number") return undefined;
  const fgColor: string = d.fgColor;
  const bgColor: string = d.bgColor;
  const ratio: number = d.contrastRatio;
  const fg = hexToRgb(fgColor);
  const bg = hexToRgb(bgColor);
  const pair = AXE_WAIVED.find((p) => distance(fg, p.fg) <= 2 && distance(bg, p.bg) <= 2);
  if (pair !== undefined && ratio >= Math.floor(pair.kit * 100) / 100 - KIT_FLOOR_TOLERANCE && ratio < pair.threshold) {
    return { pair: `OWNER-WAIVED ${pair.name}`, target, fg: fgColor, bg: bgColor, ratio, kit: pair.kit, threshold: pair.threshold };
  }
  const t = TENANT_INVALID_FIELD_WAIVER;
  if (invalidFieldNode(n) && ratio >= t.kit && ratio < t.threshold) {
    return { pair: `OWNER-WAIVED ${t.name}`, target, fg: fgColor, bg: bgColor, ratio, kit: t.kit, threshold: t.threshold };
  }
  return undefined;
}

// Splits axe violations into findings (in the specs' "<rule>: <targets>" form) and the OWNER-WAIVED kit label pairs.
export function axeFindings(violations: readonly AxeViolation[]): { findings: string[]; waived: AxeWaiver[] } {
  const findings: string[] = [];
  const waived: AxeWaiver[] = [];
  for (const v of violations) {
    const rest: string[] = [];
    for (const n of v.nodes) {
      const target = n.target.join(" ");
      const w = axeWaiver(v, n, target);
      if (w === undefined) rest.push(target);
      else waived.push(w);
    }
    if (rest.length > 0) findings.push(`${v.id}: ${rest.join(", ")}`);
  }
  return { findings, waived };
}

// The specs' axe gate: no finding except the OWNER-WAIVED kit label pairs, which are recorded as a test annotation
// (type OWNER-WAIVED, with each target, its colours and axe's ratio) and returned for the spec's own evidence.
export function expectAxeClean(violations: readonly AxeViolation[], message: string): AxeWaiver[] {
  const { findings, waived } = axeFindings(violations);
  expect(findings, message).toEqual([]);
  if (waived.length > 0) {
    test.info().annotations.push({
      type: "OWNER-WAIVED",
      description: `${message}: ${waived.map((w) => `${w.target} ${w.fg} on ${w.bg} ${w.ratio}:1 (${w.pair}, kit ${w.kit}:1, threshold ${w.threshold}:1)`).join("; ")}`,
    });
  }
  return waived;
}

export type LocalContrast = { name: string; ratio: number; fg: string; worst: string; pixels: number; min: number };

// Local worst-case text contrast over whatever is painted behind the text (imagery included; G048):
// the text is hidden (colour and fill transparent), each of its text rectangles is captured, and the
// minimum over every captured pixel of contrast(text colour composited over the pixel, pixel) is taken.
// Requires deviceScaleFactor 1. The element's inline style is restored afterwards.
export async function localTextContrast(page: Page, name: string, el: Locator, min = 4.5): Promise<LocalContrast> {
  await el.scrollIntoViewIfNeeded();
  const color = await cssToRgba(page, await computedColor(el));
  const vp = page.viewportSize() ?? { width: 1280, height: 800 };
  const rects = await el.evaluate((e) => {
    const range = document.createRange();
    range.selectNodeContents(e);
    return Array.from(range.getClientRects())
      .map((r) => ({ x: r.left, y: r.top, w: r.width, h: r.height }))
      .filter((r) => r.w >= 1 && r.h >= 1);
  });
  expect(rects.length, `${name}: text rectangles`).toBeGreaterThan(0);
  const saved = await el.evaluate((e) => {
    const h = e as HTMLElement;
    const before = h.style.cssText;
    h.style.setProperty("color", "transparent", "important");
    h.style.setProperty("-webkit-text-fill-color", "transparent", "important");
    return before;
  });
  let worst = Infinity;
  let worstPx: RGB = [0, 0, 0];
  let pixels = 0;
  try {
    for (const r of rects) {
      const x = Math.max(0, Math.floor(r.x));
      const y = Math.max(0, Math.floor(r.y));
      const width = Math.min(vp.width - x, Math.ceil(r.x + r.w) - x);
      const height = Math.min(vp.height - y, Math.ceil(r.y + r.h) - y);
      if (width < 1 || height < 1) continue;
      const img = decodePng(await page.screenshot({ clip: { x, y, width, height }, animations: "disabled" }));
      for (const px of img.pixels) {
        const ratio = contrast(over(color, px), px);
        pixels++;
        if (ratio < worst) {
          worst = ratio;
          worstPx = px;
        }
      }
    }
  } finally {
    await el.evaluate((e, s) => {
      (e as HTMLElement).style.cssText = s;
    }, saved);
  }
  expect(pixels, `${name}: captured pixels`).toBeGreaterThan(0);
  const result = { name, ratio: Number(worst.toFixed(2)), fg: hex(over(color, worstPx)), worst: hex(worstPx), pixels, min };
  return result;
}

// The painted centre pixel of the dialog Close cross: its two 2px bars cross at the control's centre, so
// that pixel is solid icon ink (G048; the icon is measured as painted, not from an svg style).
export async function crossCentre(page: Page, close: Locator): Promise<RGB> {
  const b = await box(close);
  return pixelAt(page, b.x + b.width / 2, b.y + b.height / 2);
}

// The element that carries the visible focus indicator for the focused element.
// The focused control that paints the focus indicator: a link, button or label, or the kit Select's combobox trigger
// (the appearance control, G069).
export function indicatorFor(page: Page): Locator {
  return page.locator(":focus").locator("xpath=ancestor-or-self::*[self::label or self::a or self::button or @role='combobox'][1]");
}

export type RingSample = { side: string; ring: RGB; adjacent: RGB };

// Samples the 2px outline (offset 2px) on each side that lies in the viewport.
export async function ringSamples(page: Page, el: Locator): Promise<RingSample[]> {
  const b = await box(el);
  const vp = page.viewportSize() ?? { width: 1280, height: 800 };
  const cx = b.x + b.width / 2;
  const cy = b.y + b.height / 2;
  const sides = [
    { side: "top", ring: [cx, b.y - 3], adjacent: [cx, b.y - 8] },
    { side: "bottom", ring: [cx, b.y + b.height + 2.5], adjacent: [cx, b.y + b.height + 7] },
    { side: "left", ring: [b.x - 3, cy], adjacent: [b.x - 8, cy] },
    { side: "right", ring: [b.x + b.width + 2.5, cy], adjacent: [b.x + b.width + 7, cy] },
  ];
  const out: RingSample[] = [];
  for (const s of sides) {
    const inside = [s.ring, s.adjacent].every(([x, y]) => x >= 0 && y >= 0 && x < vp.width && y < vp.height);
    if (!inside) continue;
    out.push({ side: s.side, ring: await pixelAt(page, s.ring[0], s.ring[1]), adjacent: await pixelAt(page, s.adjacent[0], s.adjacent[1]) });
  }
  return out;
}

// Waits for the element's running transitions (buttons use transition-all, so the outline colour,
// width and offset animate in after focus) so computed style and pixels describe the resting state.
export async function settled(el: Locator): Promise<void> {
  await el.evaluate((e) => Promise.allSettled(e.getAnimations().map((a) => a.finished)));
}

// Waits for the open dialog's motion (G067 motion delta: the Backdrop fades and the Popup fades and zooms from 95%,
// 100ms; G069: under reduced motion as well, since the kit's own reduced-motion rule covers only its brand motion and
// the app adds none) and its descendants' transitions, so geometry, computed style and pixels describe the open dialog
// at rest.
export async function dialogSettled(page: Page): Promise<void> {
  await page.evaluate(() =>
    Promise.allSettled(
      [...document.querySelectorAll("[data-slot=dialog-overlay], [data-slot=dialog-content]")].flatMap((e) => e.getAnimations({ subtree: true })).map((a) => a.finished),
    ),
  );
}

type FocusStyle = { outlineStyle: string; outlineWidth: string; outlineOffset: string; outlineColor: string; borderWidth: string; borderColor: string; boxShadow: string };

async function focusStyle(el: Locator): Promise<FocusStyle> {
  return el.evaluate((e) => {
    const c = getComputedStyle(e);
    return { outlineStyle: c.outlineStyle, outlineWidth: c.outlineWidth, outlineOffset: c.outlineOffset, outlineColor: c.outlineColor, borderWidth: c.borderTopWidth, borderColor: c.borderTopColor, boxShadow: c.boxShadow };
  });
}

// The outer, unblurred, unshifted ring of a computed box-shadow list (Tailwind's ring-N: `<colour> 0px 0px 0px Npx`), or
// null when there is none with a spread.
export function haloOf(boxShadow: string): { color: string; spread: number } | null {
  if (!boxShadow || boxShadow === "none") return null;
  for (const part of boxShadow.split(/,(?![^(]*\))/)) {
    const m = part.trim().match(/^(.+?\))\s+(-?[\d.]+)px\s+(-?[\d.]+)px\s+([\d.]+)px\s+([\d.]+)px(\s+inset)?$/);
    if (!m || m[6]) continue;
    if (parseFloat(m[2]) === 0 && parseFloat(m[3]) === 0 && parseFloat(m[4]) === 0 && parseFloat(m[5]) > 0) return { color: m[1], spread: parseFloat(m[5]) };
  }
  return null;
}

export type CompositeSample = { side: string; border: RGB; halo: RGB; page: RGB };

// Samples the kit's whole focus state on each side in the viewport (G069): the 1px border just inside the edge, the
// ring halo half its spread outside the edge, and the page 5px beyond the halo.
export async function compositeSamples(page: Page, el: Locator, spread: number): Promise<CompositeSample[]> {
  const b = await box(el);
  const vp = page.viewportSize() ?? { width: 1280, height: 800 };
  const cx = b.x + b.width / 2;
  const cy = b.y + b.height / 2;
  const h = spread / 2;
  const far = spread + 5;
  const sides = [
    { side: "top", border: [cx, b.y + 0.5], halo: [cx, b.y - h], page: [cx, b.y - far] },
    { side: "bottom", border: [cx, b.y + b.height - 0.5], halo: [cx, b.y + b.height + h], page: [cx, b.y + b.height + far] },
    { side: "left", border: [b.x + 0.5, cy], halo: [b.x - h, cy], page: [b.x - far, cy] },
    { side: "right", border: [b.x + b.width - 0.5, cy], halo: [b.x + b.width + h, cy], page: [b.x + b.width + far, cy] },
  ];
  const out: CompositeSample[] = [];
  for (const s of sides) {
    if (![s.border, s.halo, s.page].every(([x, y]) => x >= 0 && y >= 0 && x < vp.width && y < vp.height)) continue;
    out.push({ side: s.side, border: await pixelAt(page, s.border[0], s.border[1]), halo: await pixelAt(page, s.halo[0], s.halo[1]), page: await pixelAt(page, s.page[0], s.page[1]) });
  }
  return out;
}

export type FocusRingOptions = {
  // The control's own rendered fill (fields: the kit's bg-input/50 composite): the focus border must also reach 3:1
  // against it (G069 Plan A6).
  fill?: RGB;
};

// Focus indicator. Three shapes are measured as painted:
// - an outline (offset >= 0): present in its colour and >= 3:1 against the adjacent surface on at least one side;
// - an inset outline (negative offset, e.g. the CtaButton's S15 outline inside its angled clip): present in its colour
//   and >= 3:1 against the control's fill next to it, on at least one side;
// - the kit's whole focus state (no outline: focus-visible:border-ring plus a 3px ring at 30 %, G069 S22): the 1px
//   border is present in its colour with the halo outside it, and the border reaches 3:1 against the page beyond the
//   halo and, with `fill`, against the control's fill, on at least one side.
export async function expectFocusRing(page: Page, name: string, el: Locator, options: FocusRingOptions = {}): Promise<PairResult> {
  await settled(el);
  const s = await focusStyle(el);
  if (s.outlineStyle !== "none" && parseFloat(s.outlineWidth) >= 2) {
    const ringToken = over(await cssToRgba(page, s.outlineColor), [255, 255, 255]);
    const offset = parseFloat(s.outlineOffset);
    const samples = offset < 0 ? await insetOutlineSamples(page, el, -offset, parseFloat(s.outlineWidth)) : await ringSamples(page, el);
    const present = samples.filter((p) => distance(p.ring, ringToken) <= 40);
    expect(present.length, `${name}: focus outline pixels in the ring colour ${hex(ringToken)}`).toBeGreaterThan(0);
    const best = present.reduce((a, p) => (contrast(p.ring, p.adjacent) > contrast(a.ring, a.adjacent) ? p : a));
    return expectNonText(`${name} focus outline${offset < 0 ? " (inset)" : ""} (${best.side})`, best.ring, best.adjacent, 3);
  }
  const halo = haloOf(s.boxShadow);
  expect(halo, `${name}: focus ring halo in box-shadow ${s.boxShadow}`).not.toBeNull();
  const haloColor = await cssToRgba(page, halo!.color);
  expect(haloColor[3], `${name}: visible focus halo ${halo!.color}`).toBeGreaterThan(0);
  expect(parseFloat(s.borderWidth), `${name}: focus border width`).toBeGreaterThanOrEqual(1);
  const border = await cssToRgba(page, s.borderColor);
  const samples = await compositeSamples(page, el, halo!.spread);
  const present = samples.filter((p) => distance(p.border, over(border, p.page)) <= 40 && distance(p.halo, over(haloColor, p.page)) <= 12);
  expect(present.length, `${name}: focus border ${s.borderColor} with its halo ${halo!.color} (${JSON.stringify(samples.map((p) => ({ side: p.side, border: hex(p.border), halo: hex(p.halo), page: hex(p.page) })))})`).toBeGreaterThan(0);
  const worst = (p: CompositeSample) => Math.min(contrast(p.border, p.page), options.fill ? contrast(p.border, options.fill) : Infinity);
  const best = present.reduce((a, p) => (worst(p) > worst(a) ? p : a));
  const onPage = expectNonText(`${name} focus border against the page (${best.side})`, best.border, best.page, 3);
  if (!options.fill) return onPage;
  const onFill = expectNonText(`${name} focus border against the control's fill (${best.side})`, best.border, options.fill, 3);
  return onFill.ratio < onPage.ratio ? onFill : onPage;
}

// Samples an inset outline (offset -o, width w) at the middle of each side: the band's centre, o - w/2 inside the edge,
// and the control's own fill 3px further in.
async function insetOutlineSamples(page: Page, el: Locator, o: number, w: number): Promise<RingSample[]> {
  const b = await box(el);
  const d = o - w / 2;
  const a = o + 3;
  const cx = b.x + b.width / 2;
  const cy = b.y + b.height / 2;
  const sides = [
    { side: "top", ring: [cx, b.y + d], adjacent: [cx, b.y + a] },
    { side: "bottom", ring: [cx, b.y + b.height - d], adjacent: [cx, b.y + b.height - a] },
    { side: "left", ring: [b.x + d, cy], adjacent: [b.x + a, cy] },
    { side: "right", ring: [b.x + b.width - d, cy], adjacent: [b.x + b.width - a, cy] },
  ];
  const out: RingSample[] = [];
  for (const s of sides) out.push({ side: s.side, ring: await pixelAt(page, s.ring[0], s.ring[1]), adjacent: await pixelAt(page, s.adjacent[0], s.adjacent[1]) });
  return out;
}
