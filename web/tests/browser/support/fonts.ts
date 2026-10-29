import { createHash } from "node:crypto";

import { expect, type Locator, type Page, type Response, type Route } from "@playwright/test";

import { FONT_FACTS, FONT_PATHS, FONT_SHA256, SHELL_FONT_PATHS, type FontPath } from "./assets";
import { waitForHydration } from "./modes";

// The actual-font oracle (G048 Plan §7.4). Phase A reads what is rendered now and never waits for a font;
// it is the only phase used while fonts are held. Phase B is the positive oracle behind a bounded readiness
// gate (B0). Every wait has a named bound that fails on expiry; no timeout can pass a test.
export const BOUND = {
  holdCapture: 10_000,
  faceStatus: 10_000,
  cdp: 5_000,
  response: 15_000,
  fontsReady: 15_000,
} as const;

export const BUNDLED_FAMILIES = ["Oswald", "Barlow Condensed", "Geist", "Geist Mono"] as const;
// The bundled FontFace entries (Oswald, Barlow Condensed 400 and 600, Geist, Geist Mono), one per pinned binary.
export const BUNDLED_FACES = 5;

// The family each pinned binary serves (the font rules of web/app/globals.css).
const FAMILY_OF: Record<FontPath, (typeof BUNDLED_FAMILIES)[number]> = {
  "/fonts/oswald/oswald-wght.ttf": "Oswald",
  "/fonts/barlow-condensed/barlow-condensed-400.ttf": "Barlow Condensed",
  "/fonts/barlow-condensed/barlow-condensed-600.ttf": "Barlow Condensed",
  "/fonts/geist/geist-wght.ttf": "Geist",
  "/fonts/geist-mono/geist-mono-wght.ttf": "Geist Mono",
};

export async function bounded<T>(name: string, ms: number, p: Promise<T>): Promise<T> {
  let timer: ReturnType<typeof setTimeout> | undefined;
  const expiry = new Promise<never>((_, reject) => {
    timer = setTimeout(() => reject(new Error(`${name}: no result within ${ms} ms`)), ms);
  });
  try {
    return await Promise.race([p, expiry]);
  } finally {
    if (timer) clearTimeout(timer);
  }
}

const pathOf = (url: string) => new URL(url).pathname;
const isFontPath = (p: string): p is FontPath => (FONT_PATHS as string[]).includes(p);

// --- Phase A: rendered fonts and face status (no waiting for fonts) ------------------------------------

export type PlatformFont = { familyName: string; isCustomFont: boolean; glyphCount: number };
export type FontVerdict = "custom" | "fallback" | "mixed" | "none";
export type RenderedFont = { name: string; verdict: FontVerdict; fonts: PlatformFont[] };

// custom: exactly one platform font, a custom (web) font, with glyphs, whose family is one of `families`
// (the binary's own names) when given; fallback: no custom font; mixed: anything else.
export function verdictOf(fonts: PlatformFont[], families?: string[]): FontVerdict {
  // Nothing rendered is never a detected fallback.
  if (!fonts.some((f) => f.glyphCount > 0)) return "none";
  if (!fonts.some((f) => f.isCustomFont)) return "fallback";
  const [only] = fonts;
  if (fonts.length === 1 && only.isCustomFont && only.glyphCount > 0 && (!families || families.includes(only.familyName))) return "custom";
  return "mixed";
}

type CdpSend = <T>(method: string, params?: Record<string, unknown>) => Promise<T>;

// CSS.getPlatformFontsForNode for each target (one CDP call per bounded step). Targets are tagged with a
// temporary attribute so the CDP DOM can find them; the tags are removed afterwards.
export async function renderedFonts(page: Page, targets: Record<string, Locator>, families?: Record<string, string[]>): Promise<RenderedFont[]> {
  const cdp = await page.context().newCDPSession(page);
  const send: CdpSend = (method, params) => bounded(`CDP ${method}`, BOUND.cdp, cdp.send(method as never, params as never) as Promise<never>);
  const names = Object.keys(targets);
  try {
    for (const name of names) await targets[name].evaluate((el, n) => el.setAttribute("data-g048-font-target", n), name);
    await send("DOM.enable");
    await send("CSS.enable");
    const doc = await send<{ root: { nodeId: number } }>("DOM.getDocument", { depth: 0 });
    const out: RenderedFont[] = [];
    for (const name of names) {
      const { nodeId } = await send<{ nodeId: number }>("DOM.querySelector", { nodeId: doc.root.nodeId, selector: `[data-g048-font-target="${name}"]` });
      expect(nodeId, `${name}: CDP node`).toBeGreaterThan(0);
      const { fonts } = await send<{ fonts: PlatformFont[] }>("CSS.getPlatformFontsForNode", { nodeId });
      const plain = fonts.map((f) => ({ familyName: f.familyName, isCustomFont: f.isCustomFont, glyphCount: f.glyphCount }));
      out.push({ name, verdict: verdictOf(plain, families?.[name]), fonts: plain });
    }
    return out;
  } finally {
    for (const name of names) await targets[name].evaluate((el) => el.removeAttribute("data-g048-font-target")).catch(() => undefined);
    await cdp.detach().catch(() => undefined);
  }
}

export type FaceStatus = { family: string; weight: string; stretch: string; status: string };

// The bundled FontFace entries and their status, read immediately.
export async function faceStatuses(page: Page): Promise<FaceStatus[]> {
  return page.evaluate((families) => {
    return [...document.fonts]
      .filter((f) => (families as readonly string[]).includes(f.family.replace(/["']/g, "")))
      .map((f) => ({ family: f.family.replace(/["']/g, ""), weight: f.weight, stretch: f.stretch, status: f.status }));
  }, BUNDLED_FAMILIES);
}

// The status a bundled face must have when the page's binaries (`paths`) have `status`: a face of a family the page
// requests has that status; a face of a family it never requests (G069: Geist Mono on the shell pages) stays
// "unloaded", so nothing outside the page's set loads behind the gate's back.
export function wantedStatus(face: FaceStatus, status: string, paths: readonly FontPath[] = SHELL_FONT_PATHS): string {
  return paths.some((p) => FAMILY_OF[p] === face.family) ? status : "unloaded";
}

// Polls until every bundled face has its wanted status for the page's binaries, within the face-status bound.
export async function expectFaces(page: Page, status: "loading" | "loaded", paths: readonly FontPath[] = SHELL_FONT_PATHS, ms: number = BOUND.faceStatus): Promise<FaceStatus[]> {
  const end = Date.now() + ms;
  let last: FaceStatus[] = [];
  while (Date.now() < end) {
    last = await faceStatuses(page);
    if (last.length === BUNDLED_FACES && last.every((f) => f.status === wantedStatus(f, status, paths))) return last;
    await page.waitForTimeout(100);
  }
  throw new Error(`face status: not all ${BUNDLED_FACES} bundled faces as wanted (${status} for ${JSON.stringify(paths)}, otherwise unloaded) within ${ms} ms: ${JSON.stringify(last)}`);
}

// --- Holding font requests (controls 6 and 8) --------------------------------------------------------------

export type Hold = {
  held: () => string[];
  completed: () => string[];
  release: () => Promise<void>;
  releaseErrors: () => string[];
};

// Holds every same-origin request whose path starts with /fonts/: neither fulfilled nor aborted (an aborted
// font load would log a console error). release() lets new requests through, continues every held one and
// unroutes; it is idempotent and records its own errors.
export async function holdFonts(page: Page): Promise<Hold> {
  const held: { path: string; route: Route }[] = [];
  const completed: string[] = [];
  const errors: string[] = [];
  let released = false;
  // Same origin as the document that requests the font (the page URL is still about:blank before commit).
  const originOf = (route: Route) => {
    try {
      return new URL(route.request().frame().url()).origin;
    } catch {
      return "";
    }
  };
  const handler = async (route: Route) => {
    const url = new URL(route.request().url());
    if (released || url.origin !== originOf(route)) {
      await route.continue();
      return;
    }
    held.push({ path: url.pathname, route });
  };
  await page.route("**/fonts/**", handler);
  page.on("requestfinished", (req) => {
    const p = pathOf(req.url());
    if (p.startsWith("/fonts/")) completed.push(p);
  });
  return {
    held: () => held.map((h) => h.path),
    completed: () => [...completed],
    releaseErrors: () => [...errors],
    release: async () => {
      if (released) return;
      released = true;
      for (const h of held.splice(0)) {
        try {
          await h.route.continue();
        } catch (e) {
          errors.push(`${h.path}: ${(e as Error).message}`);
        }
      }
      try {
        await page.unroute("**/fonts/**", handler);
      } catch (e) {
        errors.push(`unroute: ${(e as Error).message}`);
      }
    },
  };
}

// Waits (bounded) until exactly the page's pinned font paths are held.
export async function expectHeld(hold: Hold, paths: readonly FontPath[] = SHELL_FONT_PATHS): Promise<string[]> {
  const want = JSON.stringify([...paths].sort());
  const end = Date.now() + BOUND.holdCapture;
  while (Date.now() < end) {
    const h = [...new Set(hold.held())].sort();
    if (h.every(isFontPath) && JSON.stringify(h) === want) return h;
    await new Promise((r) => setTimeout(r, 100));
  }
  throw new Error(`held-request capture: expected the ${paths.length} pinned font paths ${want} within ${BOUND.holdCapture} ms, held ${JSON.stringify(hold.held())}`);
}

// --- Phase B: readiness gate, bytes and metadata -------------------------------------------------------------

export type FontRecorder = { responses: Response[] };

const recorders = new WeakMap<Page, FontRecorder>();

// Records every response whose path starts with /fonts/ (attach before navigation; one recorder per page,
// emptied by openReady before each navigation). Routing the font paths disables the page's HTTP cache (a
// Playwright property of routing), so every navigation requests the pinned files again and B0 sees exactly
// one response for each. The route only falls through; a later holdFonts route takes precedence over it.
export async function recordFontResponses(page: Page): Promise<FontRecorder> {
  const existing = recorders.get(page);
  if (existing) return existing;
  const rec: FontRecorder = { responses: [] };
  recorders.set(page, rec);
  await page.route("**/fonts/**", (route) => route.fallback());
  page.on("response", (r) => {
    if (pathOf(r.url()).startsWith("/fonts/")) rec.responses.push(r);
  });
  return rec;
}

export type FontResponseRecord = { path: string; status: number; contentType: string; sameOrigin: boolean; sha256: string; bytes: number; ms: number };

// B0: exactly the page's pinned responses (`expected`: the shell pages' four binaries by default, every binary on the
// fixture page), each finished within its bound; document.fonts.ready within its bound (raced in the page; "timeout"
// fails); every face of those binaries loaded and every other bundled face unloaded. Returns the response records and
// the bodies.
export async function fontsReady(page: Page, rec: FontRecorder, expected: readonly FontPath[] = SHELL_FONT_PATHS): Promise<{ records: FontResponseRecord[]; bodies: Map<FontPath, Buffer> }> {
  const pageOrigin = new URL(page.url()).origin;
  const start = Date.now();
  const end = start + BOUND.response;
  while (rec.responses.length < expected.length && Date.now() < end) await new Promise((r) => setTimeout(r, 100));
  const paths = rec.responses.map((r) => pathOf(r.url())).sort();
  expect(paths, `B0: exactly the ${expected.length} pinned font responses of the page`).toEqual([...expected].sort());
  const records: FontResponseRecord[] = [];
  const bodies = new Map<FontPath, Buffer>();
  await Promise.all(
    rec.responses.map(async (r) => {
      const p = pathOf(r.url()) as FontPath;
      const t0 = Date.now();
      await bounded(`B0: font response ${p}`, BOUND.response, r.finished());
      const body = await bounded(`B0: font body ${p}`, BOUND.response, r.body());
      bodies.set(p, body);
      records.push({
        path: p,
        status: r.status(),
        contentType: (await r.headerValue("content-type")) ?? "",
        sameOrigin: new URL(r.url()).origin === pageOrigin,
        sha256: createHash("sha256").update(body).digest("hex"),
        bytes: body.length,
        ms: Date.now() - t0,
      });
    }),
  );
  const ready = await page.evaluate(
    (ms) => Promise.race([document.fonts.ready.then(() => "ready"), new Promise<string>((r) => setTimeout(() => r("timeout"), ms))]),
    BOUND.fontsReady,
  );
  expect(ready, "B0: document.fonts.ready").toBe("ready");
  await expectFaces(page, "loaded", expected);
  return { records: records.sort((a, b) => a.path.localeCompare(b.path)), bodies };
}

export type Ready = { hydratedAt: number; records: FontResponseRecord[]; bodies: Map<FontPath, Buffer> };

// B0 in full: goto with the default load event, waitForHydration, then the bounded readiness gate above for the page's
// binaries (the shell pages' by default; the fixture page passes FONT_PATHS).
export async function openReady(page: Page, url = "/", expected: readonly FontPath[] = SHELL_FONT_PATHS): Promise<Ready> {
  const rec = await recordFontResponses(page);
  rec.responses.length = 0;
  await page.goto(url);
  const hydratedAt = await waitForHydration(page);
  return { hydratedAt, ...(await fontsReady(page, rec, expected)) };
}

// Bytes check (item 1): same origin, 200, a font content type and the independent pin.
export function expectFontBytes(records: FontResponseRecord[]) {
  for (const r of records) {
    expect(r.sameOrigin, `${r.path}: same origin`).toBe(true);
    expect(r.status, `${r.path}: status`).toBe(200);
    expect(r.contentType, `${r.path}: content type`).toMatch(/^(font\/ttf|font\/sfnt|application\/x-font-ttf|application\/font-sfnt)\b/);
    expect(r.sha256, `${r.path}: SHA-256`).toBe(FONT_SHA256[r.path as FontPath]);
  }
}

export type FontMetadata = { names: Record<number, string>; axes: Record<string, [number, number]>; weightClass: number };

// Parses the name (IDs 1, 2, 4, 6, 16), fvar (axes and ranges) and OS/2 usWeightClass tables of a TrueType binary.
export function parseFont(buf: Buffer): FontMetadata {
  const numTables = buf.readUInt16BE(4);
  const tables = new Map<string, { offset: number; length: number }>();
  for (let i = 0; i < numTables; i++) {
    const rec = 12 + i * 16;
    tables.set(buf.toString("latin1", rec, rec + 4), { offset: buf.readUInt32BE(rec + 8), length: buf.readUInt32BE(rec + 12) });
  }
  const names: Record<number, string> = {};
  const name = tables.get("name");
  if (name) {
    const count = buf.readUInt16BE(name.offset + 2);
    const strings = name.offset + buf.readUInt16BE(name.offset + 4);
    for (let i = 0; i < count; i++) {
      const r = name.offset + 6 + i * 12;
      const platform = buf.readUInt16BE(r);
      const encoding = buf.readUInt16BE(r + 2);
      const id = buf.readUInt16BE(r + 6);
      const length = buf.readUInt16BE(r + 8);
      const offset = strings + buf.readUInt16BE(r + 10);
      if (![1, 2, 4, 6, 16].includes(id) || names[id]) continue;
      if (platform === 3 && (encoding === 1 || encoding === 10)) {
        let s = "";
        for (let j = 0; j < length; j += 2) s += String.fromCharCode(buf.readUInt16BE(offset + j));
        names[id] = s;
      } else if (platform === 1 && encoding === 0) names[id] = buf.toString("latin1", offset, offset + length);
    }
  }
  const axes: Record<string, [number, number]> = {};
  const fvar = tables.get("fvar");
  if (fvar) {
    const axesOffset = fvar.offset + buf.readUInt16BE(fvar.offset + 4);
    const axisCount = buf.readUInt16BE(fvar.offset + 8);
    const axisSize = buf.readUInt16BE(fvar.offset + 10);
    for (let i = 0; i < axisCount; i++) {
      const a = axesOffset + i * axisSize;
      axes[buf.toString("latin1", a, a + 4)] = [buf.readInt32BE(a + 4) / 65536, buf.readInt32BE(a + 12) / 65536];
    }
  }
  const os2 = tables.get("OS/2");
  const weightClass = os2 ? buf.readUInt16BE(os2.offset + 4) : 0;
  return { names, axes, weightClass };
}

// The binary's own family names (typographic ID 16, else legacy ID 1, plus ID 1 itself).
export function familyNames(meta: FontMetadata): string[] {
  return [...new Set([meta.names[16], meta.names[1]].filter((n): n is string => Boolean(n)))];
}

// Metadata check (item 2): family, axis ranges and static weight classes as declared by each binary the readiness gate
// returned (B0 already required exactly the page's set).
export function expectFontMetadata(bodies: Map<FontPath, Buffer>): Record<string, FontMetadata> {
  const out: Record<string, FontMetadata> = {};
  const gated = FONT_PATHS.filter((p) => bodies.has(p));
  expect(gated.length, "font bodies from the readiness gate").toBeGreaterThan(0);
  for (const p of gated) {
    const body = bodies.get(p);
    expect(body, `${p}: body`).toBeTruthy();
    const meta = parseFont(body as Buffer);
    out[p] = meta;
    const facts = FONT_FACTS[p];
    expect(familyNames(meta), `${p}: family name`).toContain(facts.family);
    for (const [tag, [lo, hi]] of Object.entries(facts.axes ?? {})) {
      expect(meta.axes[tag], `${p}: fvar ${tag}`).toEqual([lo, hi]);
    }
    if (facts.weightClass) expect(meta.weightClass, `${p}: OS/2 usWeightClass`).toBe(facts.weightClass);
  }
  return out;
}

// Families per bundled CSS family, from the binaries the page loaded (for phase-A/B verdicts; a family whose binary the
// page never requested maps to no names, so a "custom" verdict against it cannot pass).
export function familyMap(meta: Record<string, FontMetadata>): Record<(typeof BUNDLED_FAMILIES)[number], string[]> {
  const of = (...paths: FontPath[]) => [...new Set(paths.filter((p) => meta[p] !== undefined).flatMap((p) => familyNames(meta[p])))];
  return {
    Oswald: of("/fonts/oswald/oswald-wght.ttf"),
    "Barlow Condensed": of("/fonts/barlow-condensed/barlow-condensed-400.ttf", "/fonts/barlow-condensed/barlow-condensed-600.ttf"),
    Geist: of("/fonts/geist/geist-wght.ttf"),
    "Geist Mono": of("/fonts/geist-mono/geist-mono-wght.ttf"),
  };
}

// Synthesis twin check (item 5): two element screenshots in the same state, the second with
// font-synthesis: none; byte-identical means no synthetic bold or oblique was in use.
export async function synthesized(el: Locator): Promise<boolean> {
  await el.scrollIntoViewIfNeeded();
  const a = await el.screenshot({ animations: "disabled" });
  const saved = await el.evaluate((e) => {
    const h = e as HTMLElement;
    const s = h.style.cssText;
    h.style.setProperty("font-synthesis", "none", "important");
    return s;
  });
  try {
    const b = await el.screenshot({ animations: "disabled" });
    return !a.equals(b);
  } finally {
    await el.evaluate((e, s) => {
      (e as HTMLElement).style.cssText = s;
    }, saved);
  }
}

// Negative control 7: a test-only face using only the 400 binary, loaded through FontFace.load() within the
// response bound; an element in it at weight 700 must be reported as synthesized.
export async function synthesisProbe(page: Page): Promise<Locator> {
  await page.evaluate(async (ms) => {
    const face = new FontFace("Synthesis Probe", 'url("/fonts/barlow-condensed/barlow-condensed-400.ttf")', { weight: "400" });
    const loaded = await Promise.race([face.load(), new Promise<null>((r) => setTimeout(() => r(null), ms))]);
    if (!loaded) throw new Error(`synthesis probe face: not loaded within ${ms} ms`);
    document.fonts.add(face);
    const p = document.createElement("p");
    p.id = "g048-synthesis-probe";
    p.textContent = "Synthesis probe text";
    p.style.cssText = 'font-family:"Synthesis Probe";font-weight:700;font-size:24px;position:relative;z-index:2147483647;background:#fff;color:#000;margin:8px;display:inline-block';
    document.body.prepend(p);
  }, BOUND.response);
  return page.locator("#g048-synthesis-probe");
}
