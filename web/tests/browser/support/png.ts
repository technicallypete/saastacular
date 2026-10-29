import { inflateSync } from "node:zlib";

import type { Page } from "@playwright/test";

export type RGB = [number, number, number];

// Decodes the single pixel of a 1x1 PNG (colour type 2 or 6, 8-bit). With one row and one
// pixel, every PNG filter predicts zero, so the bytes after the filter byte are the pixel itself.
export function decodeSinglePixel(png: Buffer): RGB {
  let offset = 8;
  let colorType = -1;
  const data: Buffer[] = [];
  while (offset < png.length) {
    const length = png.readUInt32BE(offset);
    const type = png.toString("ascii", offset + 4, offset + 8);
    const body = png.subarray(offset + 8, offset + 8 + length);
    if (type === "IHDR") {
      if (body.readUInt32BE(0) !== 1 || body.readUInt32BE(4) !== 1 || body[8] !== 8) throw new Error("expected a 1x1 8-bit PNG");
      colorType = body[9];
    } else if (type === "IDAT") data.push(body);
    offset += 12 + length;
  }
  if (colorType !== 2 && colorType !== 6) throw new Error(`unsupported PNG colour type ${colorType}`);
  const raw = inflateSync(Buffer.concat(data));
  return [raw[1], raw[2], raw[3]];
}

export type Image = { width: number; height: number; pixels: RGB[] };

// Decodes an 8-bit, non-interlaced RGB (colour type 2) or RGBA (6) PNG of any size, undoing the
// per-row filters 0-4, for region analysis (G048). Alpha is dropped: screenshots are opaque.
export function decodePng(png: Buffer): Image {
  let offset = 8;
  let width = 0;
  let height = 0;
  let colorType = -1;
  const data: Buffer[] = [];
  while (offset < png.length) {
    const length = png.readUInt32BE(offset);
    const type = png.toString("ascii", offset + 4, offset + 8);
    const body = png.subarray(offset + 8, offset + 8 + length);
    if (type === "IHDR") {
      width = body.readUInt32BE(0);
      height = body.readUInt32BE(4);
      if (body[8] !== 8 || body[12] !== 0) throw new Error("expected an 8-bit, non-interlaced PNG");
      colorType = body[9];
    } else if (type === "IDAT") data.push(body);
    offset += 12 + length;
  }
  if (colorType !== 2 && colorType !== 6) throw new Error(`unsupported PNG colour type ${colorType}`);
  const bpp = colorType === 6 ? 4 : 3;
  const stride = width * bpp;
  const raw = inflateSync(Buffer.concat(data));
  const out = Buffer.alloc(stride * height);
  for (let y = 0; y < height; y++) {
    const filter = raw[y * (stride + 1)];
    const line = raw.subarray(y * (stride + 1) + 1, (y + 1) * (stride + 1));
    for (let x = 0; x < stride; x++) {
      const a = x >= bpp ? out[y * stride + x - bpp] : 0;
      const b = y > 0 ? out[(y - 1) * stride + x] : 0;
      const c = x >= bpp && y > 0 ? out[(y - 1) * stride + x - bpp] : 0;
      let v = line[x];
      if (filter === 1) v += a;
      else if (filter === 2) v += b;
      else if (filter === 3) v += Math.floor((a + b) / 2);
      else if (filter === 4) {
        const p = a + b - c;
        const pa = Math.abs(p - a), pb = Math.abs(p - b), pc = Math.abs(p - c);
        v += pa <= pb && pa <= pc ? a : pb <= pc ? b : c;
      } else if (filter !== 0) throw new Error(`unsupported PNG filter ${filter}`);
      out[y * stride + x] = v & 0xff;
    }
  }
  const pixels: RGB[] = [];
  for (let i = 0; i < width * height; i++) pixels.push([out[i * bpp], out[i * bpp + 1], out[i * bpp + 2]]);
  return { width, height, pixels };
}

// The rendered colour at a CSS-pixel point: everything painted there, composited (translucent
// layers, pseudo-element fills, filters, backdrops). Requires deviceScaleFactor 1.
export async function pixelAt(page: Page, x: number, y: number): Promise<RGB> {
  const png = await page.screenshot({ clip: { x: Math.floor(x), y: Math.floor(y), width: 1, height: 1 }, animations: "disabled" });
  return decodeSinglePixel(png);
}
