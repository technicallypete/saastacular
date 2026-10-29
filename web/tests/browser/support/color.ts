import type { Locator, Page } from "@playwright/test";

import type { RGB } from "./png";

export type RGBA = [number, number, number, number];

// Converts any CSS colour string to 8-bit sRGB through a canvas readback in the page, so the
// browser's own colour parsing (hex, rgb(), oklab(), color-mix results) is used.
export async function cssToRgba(page: Page, color: string): Promise<RGBA> {
  if (!color.trim()) throw new Error("cssToRgba: empty colour");
  return page.evaluate((c) => {
    const canvas = document.createElement("canvas");
    canvas.width = canvas.height = 1;
    const ctx = canvas.getContext("2d", { willReadFrequently: true });
    if (!ctx) throw new Error("no 2d context");
    ctx.clearRect(0, 0, 1, 1);
    // An unparsable colour leaves fillStyle unchanged; fail instead of measuring the sentinel.
    ctx.fillStyle = "#010203";
    ctx.fillStyle = c;
    if (ctx.fillStyle === "#010203" && c.trim().toLowerCase() !== "#010203") throw new Error(`cssToRgba: not a colour: ${c}`);
    ctx.fillRect(0, 0, 1, 1);
    const d = ctx.getImageData(0, 0, 1, 1).data;
    return [d[0], d[1], d[2], d[3]] as [number, number, number, number];
  }, color);
}

export async function computedColor(locator: Locator, property = "color"): Promise<string> {
  return locator.evaluate((el, p) => getComputedStyle(el).getPropertyValue(p), property);
}

// Composites a possibly translucent foreground over an opaque backing colour.
export function over(fg: RGBA, bg: RGB): RGB {
  const a = fg[3] / 255;
  return [0, 1, 2].map((i) => Math.round(fg[i] * a + bg[i] * (1 - a))) as RGB;
}

// A whole-element CSS brightness(f) as core models it (core/internal/branding `scale`: each sRGB channel
// multiplied by f, clamped at 255), rounded to the 8-bit value the page paints.
export function scaled(c: RGB, f: number): RGB {
  return c.map((v) => Math.min(255, Math.round(v * f))) as RGB;
}

function channel(v: number): number {
  const c = v / 255;
  return c <= 0.04045 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4;
}

export function luminance([r, g, b]: RGB): number {
  return 0.2126 * channel(r) + 0.7152 * channel(g) + 0.0722 * channel(b);
}

// WCAG 2.x contrast ratio.
export function contrast(a: RGB, b: RGB): number {
  const [hi, lo] = [luminance(a), luminance(b)].sort((x, y) => y - x);
  return (hi + 0.05) / (lo + 0.05);
}

export function rgb(c: RGBA): RGB {
  return [c[0], c[1], c[2]];
}

export function distance(a: RGB, b: RGB): number {
  return Math.max(Math.abs(a[0] - b[0]), Math.abs(a[1] - b[1]), Math.abs(a[2] - b[2]));
}

export function hex(rgb: RGB): string {
  return `#${rgb.map((v) => v.toString(16).padStart(2, "0")).join("")}`;
}

export function hexToRgb(h: string): RGB {
  const s = h.replace("#", "");
  return [0, 2, 4].map((i) => parseInt(s.slice(i, i + 2), 16)) as RGB;
}
