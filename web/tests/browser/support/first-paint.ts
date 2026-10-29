import { expect, test, type Locator, type Page } from "@playwright/test";

import { cssToRgba, distance, rgb } from "./color";
import { resolve, type Resolved } from "./modes";
import { pixelAt, type RGB } from "./png";

// First paint (Plan §11.4; G069 S5/S6: one appearance manager, next-themes, writing the resolved class on <html>). The
// probe is installed at document creation, before any page script. Computed-style samples record their own observation
// time, kept apart from the rAF timestamp argument and from observer delivery; paint correlation uses the entries' own
// startTime. Shared by appearance.spec.ts (the real service at "/") and tokens.spec.ts (the tenant fixture page).

declare global {
  interface Window {
    __probe: {
      frames: { observed: number; rafTs: number; attr: string | null; present: boolean; bg?: string; scheme?: string; h1?: string }[];
      paints: { name: string; startTime: number }[];
      lcp: { startTime: number; renderTime: number; tag?: string; inScope: boolean }[];
      insert: { observed: number; attr: string | null } | null;
      stop: boolean;
    };
  }
}

// What the probe samples: `content` is the theme-bearing content that must be laid out for a sample to count (the first
// contentful element in document order, so it is on screen in the first contentful paint); `heading` is the H1 whose
// ink is sampled from the frame it is laid out in. The ground is the body's own background (kit base: bg-background).
export type ProbeTarget = { content: string; heading: string };

// The expected first-paint values of one appearance: the ground, the heading ink, the two other text inks it must be
// distinguishable from (foreground at 80 % and muted-foreground) and the primary fill.
export type PaintPalette = { background: string; ink: string; text2: string; text3: string; primary: string };

export type Stored = "light" | "dark" | "system";

export async function installProbe(page: Page, target: ProbeTarget) {
  await page.addInitScript((t) => {
    const probe: Window["__probe"] = { frames: [], paints: [], lcp: [], insert: null, stop: false };
    window.__probe = probe;
    // The resolved appearance <html> carries: next-themes writes exactly one of the classes light/dark.
    const resolved = () => {
      const c = document.documentElement.classList;
      return c.contains("dark") ? "dark" : c.contains("light") ? "light" : null;
    };
    new PerformanceObserver((list) => {
      for (const e of list.getEntries()) probe.paints.push({ name: e.name, startTime: e.startTime });
    }).observe({ type: "paint", buffered: true });
    new PerformanceObserver((list) => {
      for (const entry of list.getEntries()) {
        const e = entry as PerformanceEntry & { renderTime: number; element: Element | null };
        // Theme-bearing: rendered content inside the body (G069: the palette is the document's, carried by <html>).
        probe.lcp.push({ startTime: e.startTime, renderTime: e.renderTime, tag: e.element?.tagName, inScope: !!e.element?.closest("body") });
      }
    }).observe({ type: "largest-contentful-paint", buffered: true });
    new MutationObserver(() => {
      if (!probe.insert && document.querySelector(t.heading)) probe.insert = { observed: performance.now(), attr: resolved() };
    }).observe(document, { childList: true, subtree: true });
    // A content sample is the body with the theme-bearing content laid out. The heading follows it in the streamed HTML,
    // and Chromium's incremental parser may paint the content in a frame before the heading has been parsed (no theme
    // changes between the two), so the heading's ink is sampled from the frame it is laid out in rather than being a
    // condition of the sample itself.
    const sample = (rafTs: number) => {
      const content = document.querySelector(t.content);
      const h1 = document.querySelector(t.heading);
      const present = !!(document.body && content && content.getBoundingClientRect().height > 0);
      const rec: Window["__probe"]["frames"][number] = { observed: 0, rafTs, attr: resolved(), present };
      if (present) {
        rec.bg = getComputedStyle(document.body).backgroundColor;
        rec.scheme = getComputedStyle(document.documentElement).colorScheme;
        if (h1 && h1.getBoundingClientRect().height > 0) rec.h1 = getComputedStyle(h1).color;
      }
      rec.observed = performance.now();
      probe.frames.push(rec);
      if (!probe.stop && probe.frames.length < 5000) requestAnimationFrame(sample);
    };
    requestAnimationFrame(sample);
  }, target);
}

// The painted ink of a text element: of points sampled along its text's vertical middle, the one farthest from the
// ground behind it (a glyph stem's interior).
export async function inkOf(page: Page, el: Locator, ground: RGB): Promise<RGB> {
  const t = await el.evaluate((node) => {
    const range = document.createRange();
    range.selectNodeContents(node);
    const r = range.getClientRects()[0];
    return { x: r.left, y: r.top + r.height / 2, w: r.width };
  });
  let best: RGB = ground;
  for (let i = 1; i < 48; i++) {
    const p = await pixelAt(page, t.x + (t.w * i) / 48, t.y);
    if (distance(p, ground) > distance(best, ground)) best = p;
  }
  return best;
}

// Precondition from the typed tables only: the expected heading ink of appearance m is distinct (> 16 per channel) from
// that appearance's other text inks and from the other appearance's ink, so the frame-level ink comparisons (tolerance
// 1, painted 8) can only pass for the right ink in the right palette.
export async function expectInkDiscriminates(page: Page, m: Resolved, palettes: Record<Resolved, PaintPalette>) {
  const other: Resolved = m === "light" ? "dark" : "light";
  const ink = rgb(await cssToRgba(page, palettes[m].ink));
  for (const [what, value] of [["foreground 80 %", palettes[m].text2], ["muted-foreground", palettes[m].text3], [`${other} ink`, palettes[other].ink]] as const) {
    expect(distance(ink, rgb(await cssToRgba(page, value))), `precondition: ${m} ink ${palettes[m].ink} is distinct from ${what} ${value}`).toBeGreaterThan(16);
  }
}

export async function evaluateTimeline(page: Page, stored: Stored | "blocked", os: Resolved, hydratedAt: number, palettes: Record<Resolved, PaintPalette>) {
  await page.evaluate(() => (window.__probe.stop = true));
  const probe = await page.evaluate(() => window.__probe);
  // The raw timeline first, so a rejected run still carries what the probe observed.
  await test.info().attach("first-paint-timeline.json", {
    body: JSON.stringify({ hydratedAt, paints: probe.paints, lcp: probe.lcp, insert: probe.insert, frameCount: probe.frames.length, frames: probe.frames.slice(0, 1000) }, null, 2),
    contentType: "application/json",
  });
  // Missing or unreadable storage means System; <html> carries the resolved appearance (G069 S5).
  const want = resolve(stored === "blocked" ? "system" : stored, os);
  const fcp = probe.paints.find((p) => p.name === "first-contentful-paint")?.startTime;
  expect(fcp, "first-contentful-paint entry").toBeDefined();
  // A content sample at or before the first contentful paint: without one the probe cannot vouch for the first painted
  // content, and the run is rejected, never passed.
  const content = probe.frames.filter((f) => f.present);
  const valid = content.filter((f) => f.observed <= fcp!);
  expect(valid.length, "a visible theme-bearing content sample observed at or before FCP (run rejected otherwise)").toBeGreaterThan(0);
  expect(probe.lcp.length, "largest-contentful-paint entry").toBeGreaterThan(0);
  expect(probe.lcp[probe.lcp.length - 1].inScope, "LCP element is theme-bearing content").toBe(true);
  expect(probe.insert?.attr, "resolved <html> class when the heading was inserted").toBe(want);
  const wantBg = rgb(await cssToRgba(page, palettes[want].background));
  const wantInk = rgb(await cssToRgba(page, palettes[want].ink));
  await expectInkDiscriminates(page, want, palettes);
  const window_ = content.filter((f) => f.observed <= hydratedAt);
  expect(window_.length).toBeGreaterThan(0);
  // The heading is laid out before hydration completes (hydration needs the full DOM), so its ink is always checked
  // from its first laid-out frame on; a content frame without it is one the parser had not yet reached it in.
  const headings = window_.filter((f) => f.h1 !== undefined);
  expect(headings.length, "the laid-out heading sampled in a content frame before hydration completed").toBeGreaterThan(0);
  for (const f of window_) {
    expect(f.attr, `frame at ${f.observed.toFixed(1)}ms: resolved <html> class`).toBe(want);
    expect(f.scheme, `frame at ${f.observed.toFixed(1)}ms: color-scheme`).toBe(want);
    expect(distance(rgb(await cssToRgba(page, f.bg!)), wantBg), `frame at ${f.observed.toFixed(1)}ms: palette ${f.bg}`).toBeLessThanOrEqual(1);
    if (f.h1 !== undefined) expect(distance(rgb(await cssToRgba(page, f.h1)), wantInk), `frame at ${f.observed.toFixed(1)}ms: heading ink ${f.h1}`).toBeLessThanOrEqual(1);
  }
  return {
    want,
    expectedInk: palettes[want].ink,
    fcp,
    firstContentFrame: content[0],
    firstH1Frame: headings[0],
    validSamples: valid.length,
    framesUntilHydration: window_.length,
    h1FramesUntilHydration: headings.length,
    lcp: probe.lcp.at(-1),
    insert: probe.insert,
  };
}

// Every stored preference against both OS appearances, plus unavailable storage (System) against both.
export const FIRST_PAINT_CASES: { stored: Stored | "blocked"; os: Resolved }[] = [
  ...(["light", "dark", "system"] as const).flatMap((stored) => (["light", "dark"] as const).map((os) => ({ stored, os }))),
  { stored: "blocked", os: "light" },
  { stored: "blocked", os: "dark" },
];
