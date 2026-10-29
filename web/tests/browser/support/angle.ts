import { expect, type Locator, type Page } from "@playwright/test";

import { computedColor, contrast, cssToRgba, distance, hex, over } from "./color";
import { ringSamples, settled } from "./contrast";

// Resolves a computed `polygon(...)` against a box: each coordinate is N%, Npx or calc(100% - Npx).
export function polygonPoints(poly: string, width: number, height: number): [number, number][] {
  const inner = poly.replace(/^polygon\(/, "").replace(/\)$/, "");
  const coord = (expr: string, size: number): number => {
    const e = expr.trim();
    const calc = e.match(/^calc\(\s*100%\s*-\s*([\d.]+)px\s*\)$/);
    if (calc) return size - Number(calc[1]);
    if (e.endsWith("%")) return (Number(e.slice(0, -1)) / 100) * size;
    if (e.endsWith("px")) return Number(e.slice(0, -2));
    if (e === "0") return 0;
    throw new Error(`unsupported polygon coordinate: ${e}`);
  };
  return inner.split(/,(?![^(]*\))/).map((pair) => {
    const parts = pair.trim().match(/^(calc\([^)]*\)|\S+)\s+(calc\([^)]*\)|\S+)$/);
    if (!parts) throw new Error(`unsupported polygon point: ${pair}`);
    return [coord(parts[1], width), coord(parts[2], height)];
  });
}

// Every corner of the label's box lies inside the angled polygon, and the focus outline is whole: present
// on all four sides and at least 3:1 against the adjacent surface on each of them (G048, R6).
export async function expectAngleIntact(page: Page, button: Locator, name: string) {
  const geo = await button.evaluate((el) => {
    // ::before (absolute, inset 0) fills the padding box, inside the border.
    const b = el.getBoundingClientRect();
    const r = { left: b.left + el.clientLeft, top: b.top + el.clientTop, width: el.clientWidth, height: el.clientHeight };
    const range = document.createRange();
    range.selectNodeContents(el);
    const t = range.getBoundingClientRect();
    const poly = getComputedStyle(el, "::before").clipPath;
    return { r: { x: r.left, y: r.top, w: r.width, h: r.height }, t: { x: t.left, y: t.top, w: t.width, h: t.height }, poly };
  });
  const pts = polygonPoints(geo.poly, geo.r.w, geo.r.h).map(([x, y]) => [geo.r.x + x, geo.r.y + y]);
  expect(pts.length, `${name}: polygon`).toBe(6);
  const insidePoly = (x: number, y: number) => {
    let inside = false;
    for (let i = 0, j = pts.length - 1; i < pts.length; j = i++) {
      const [xi, yi] = pts[i], [xj, yj] = pts[j];
      if (yi > y !== yj > y && x < ((xj - xi) * (y - yi)) / (yj - yi) + xi) inside = !inside;
    }
    return inside;
  };
  for (const [x, y] of [[geo.t.x, geo.t.y], [geo.t.x + geo.t.w, geo.t.y], [geo.t.x, geo.t.y + geo.t.h], [geo.t.x + geo.t.w, geo.t.y + geo.t.h]]) {
    expect(insidePoly(x, y), `${name}: label corner (${x.toFixed(1)}, ${y.toFixed(1)}) clipped by the angle`).toBe(true);
  }
  await button.focus();
  await page.keyboard.press("Shift+Tab");
  await page.keyboard.press("Tab");
  await settled(button);
  // The ring is sampled up to 8px outside each edge; focusing only scrolls the button itself into view, which can
  // leave a button at the viewport edge with a side outside it. Centre it so all four sides are on screen.
  await button.evaluate((el) => el.scrollIntoView({ block: "center", inline: "center", behavior: "instant" }));
  await page.evaluate(() => new Promise<void>((r) => requestAnimationFrame(() => requestAnimationFrame(() => r()))));
  const ring = over(await cssToRgba(page, await computedColor(button, "outline-color")), [255, 255, 255]);
  const samples = await ringSamples(page, button);
  expect(samples.map((s) => s.side).sort(), `${name}: outline sampled on all sides`).toEqual(["bottom", "left", "right", "top"]);
  for (const s of samples) {
    expect(distance(s.ring, ring), `${name}: focus outline present on the ${s.side}`).toBeLessThanOrEqual(40);
    expect(contrast(s.ring, s.adjacent), `${name}: focus outline ${hex(s.ring)} against ${hex(s.adjacent)} on the ${s.side}`).toBeGreaterThanOrEqual(3);
  }
}
