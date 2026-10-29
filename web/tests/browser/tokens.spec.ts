import AxeBuilder from "@axe-core/playwright";
import type { Locator, Page, TestInfo } from "@playwright/test";

import { FIXTURE_PREVIEWS, FIXTURE_TENANTS } from "../fixtures/token-scopes/fixtures";
import { polygonPoints } from "./support/angle";
import { FONT_PATHS } from "./support/assets";
import { computedColor, contrast, cssToRgba, distance, hex, hexToRgb, over, rgb, scaled, type RGBA } from "./support/color";
import { box, compositeSamples, crossCentre, dialogSettled, expectFocusRing, expectNonText, expectText, fillOf, haloOf, textColor, type PairResult } from "./support/contrast";
import { evaluateTimeline, FIRST_PAINT_CASES, installProbe, type PaintPalette, type ProbeTarget } from "./support/first-paint";
import { expectFontBytes, expectFontMetadata, familyMap, openReady, renderedFonts, synthesized } from "./support/fonts";
import {
  blockStorage,
  CSS_VARIABLE,
  expectAtKit,
  foreground80,
  OWNER_WAIVED,
  PALETTE_NAMES,
  resetFocus,
  SCRIM,
  SOURCES,
  storeMode,
  WAIVED_TOLERANCE,
  waitForHydration,
  type KitPairResult,
  type PaletteName,
} from "./support/modes";
import { expect, test } from "./support/page-errors";
import { decodePng, pixelAt, type Image, type RGB } from "./support/png";
import { BUTTON_ICON, BUTTON_SIZES, SOURCE_ANGLE } from "./support/source-reference";
import { expectNegativeCases, expectNoFindings, inspectSpacing, SPACING_CSS, type SpacingItem, type SpacingState } from "./support/spacing";

// G069 (Plan §2.4, S5–S8, S22, S24): the fixture page renders one root tenant palette through the application's own
// tenant style at the document root (?tenant=<label>; branded by default), so every body portal inherits it, plus two
// isolated brand previews (A: Light only, B: Dark) that render no portaled component. Surfaces are located as
// [data-fixture-surface="root|A|B"] with their content in the surface's own section. The contract the page implements
// is recorded with the G069 W4 handback (fixture contract).
type Mode = "light" | "dark";
const modeOf = (info: TestInfo) => (info.project.use.colorScheme === "dark" ? "dark" : "light") as Mode;
// The kit Button variants the application uses (S11; link is unused). The CtaButton renders only under the platform
// palette (S24) and is probed on its own.
const VARIANTS = ["default", "outline", "secondary", "ghost", "destructive"] as const;
type Variant = (typeof VARIANTS)[number];
// Variants with a fill of their own at rest (the kit's alpha model: primary, secondary, the destructive tint).
const FILLED = ["default", "secondary", "destructive"] as const;
const ROOT = "root";
const PLATFORM = "platform";
const DEFAULT_TENANT = "branded";
const TENANTS = FIXTURE_TENANTS.map((t) => t.label);
const PREVIEWS = FIXTURE_PREVIEWS.map((p) => p.label);

type Values = Partial<Record<PaletteName, string>>;
// A surface: the root of one tenant's page, or a preview (on the default tenant's page).
type Surface = { tenant: string; label: string };

function previewOf(label: string): { appearance: Mode; tokens: Values } {
  const p = FIXTURE_PREVIEWS.find((x) => x.label === label);
  if (!p) throw new Error(`unknown fixture preview ${label}`);
  if (p.appearance !== "light" && p.appearance !== "dark") throw new Error(`preview ${label}: appearance ${String(p.appearance)}`);
  return { appearance: p.appearance, tokens: p.tokens as Values };
}

function tenantOwn(tenant: string, mode: Mode, name: PaletteName): string | undefined {
  const t = FIXTURE_TENANTS.find((x) => x.label === tenant);
  if (!t) throw new Error(`unknown fixture tenant ${tenant}`);
  return (t.tokens?.[mode] as Values | undefined)?.[name];
}

// The palette appearance of a surface: the document's for the root, the preview's own otherwise (one per preview, S8).
function paletteMode(s: Surface, mode: Mode): Mode {
  return s.label === ROOT ? mode : previewOf(s.label).appearance;
}

// Whether the kit's dark: variants apply on a surface: the surface's own appearance (Codex P0.2 decision 4). The app's
// dark variant (web/app/globals.css) excludes a Light preview, so a Light preview on a Dark page takes the Light
// utilities, and a Dark preview on a Light page still takes the Dark ones.
function darkVariants(s: Surface, mode: Mode): boolean {
  return paletteMode(s, mode) === "dark";
}

// The expected source: the surface's own value for its appearance (the root tenant's, or the preview's), else the
// independent platform value (SOURCES, never read from the page).
function expectedSource(s: Surface, mode: Mode, name: PaletteName): string {
  if (s.label === ROOT) return tenantOwn(s.tenant, mode, name) ?? SOURCES[mode][name];
  const p = previewOf(s.label);
  return p.tokens[name] ?? SOURCES[p.appearance][name];
}

async function expected(page: Page, s: Surface, mode: Mode, name: PaletteName): Promise<RGB> {
  return rgb(await cssToRgba(page, expectedSource(s, mode, name)));
}

const fixtureUrl = (tenant: string, extra = "") => `/token-fixtures?tenant=${encodeURIComponent(tenant)}${extra}`;

// Fixture page of one root tenant with the bundled fonts loaded (the bounded phase-B readiness gate, §7.4 B0). The fixture
// renders every bundled face (Geist Mono included, unlike the shell pages), so it is gated on all five binaries.
async function load(page: Page, tenant = DEFAULT_TENANT) {
  await page.emulateMedia({ reducedMotion: "reduce" });
  await openReady(page, fixtureUrl(tenant), FONT_PATHS);
  await page.mouse.move(0, 0);
}

const surfaceEl = (page: Page, label: string) => page.locator(`[data-fixture-surface="${label}"]`);
const content = (page: Page, label: string) => page.locator(`[data-fixture-surface="${label}"] > section`);

function near(actual: RGB, want: RGB, what: string, tolerance = 2) {
  expect(distance(actual, want), `${what}: ${hex(actual)} vs expected ${hex(want)}`).toBeLessThanOrEqual(tolerance);
}

function distance4(a: RGBA, b: RGBA): number {
  return Math.max(...a.map((v, i) => Math.abs(v - b[i])));
}

// Computed style of an element (or a pseudo-element) as a plain record.
async function styleOf(el: Locator, props: string[], pseudo?: string): Promise<Record<string, string>> {
  return el.evaluate((e, [ps, p]) => Object.fromEntries(ps.map((x) => [x, getComputedStyle(e, p).getPropertyValue(x)])), [props, pseudo ?? null] as const);
}

// The amount of a computed `brightness(...)` filter (a number or a percentage); NaN for anything else.
function brightness(filter: string): number {
  const m = filter.trim().match(/^brightness\(([\d.]+)(%?)\)$/);
  return m ? Number(m[1]) / (m[2] ? 100 : 1) : NaN;
}

const frames = (page: Page) => page.evaluate(() => new Promise<void>((r) => requestAnimationFrame(() => requestAnimationFrame(() => r()))));

// Waits for every finite running animation and transition in the document, then two frames. G069 owner kit-fidelity
// direction (2026-10-26): the app's global reduced-motion rule is withdrawn and the kit's components carry no reduced-motion
// variant, so under the emulated reduced motion a palette change still transitions the kit Buttons' colours (transition-all,
// 150ms) and the dialog's (duration-100 with the initial transition-property all) before anything is sampled. Infinite
// animations (the kit stops its brand ribbon under reduced motion itself) are not waited for.
async function settledPage(page: Page): Promise<void> {
  await page.evaluate(() => Promise.allSettled(document.getAnimations().filter((a) => a.effect?.getComputedTiming().endTime !== Infinity).map((a) => a.finished)));
  await frames(page);
}

// Waits for every running animation and transition of a control, its pseudo-elements included, then for
// two frames, so computed styles describe the settled state.
async function settledDeep(page: Page, el: Locator): Promise<void> {
  await el.evaluate((e) => Promise.allSettled(e.getAnimations({ subtree: true }).map((a) => a.finished)));
  await frames(page);
}

async function colorOf(page: Page, value: string): Promise<RGB> {
  return rgb(await cssToRgba(page, value));
}

// Switches the emulated OS appearance (reduced motion kept). With no stored preference next-themes follows it (System)
// and rewrites the <html> class; the switch has landed once the class is the new appearance, then the kit's colour
// transitions run out (settledPage) and two frames let it reach style and paint.
async function setScheme(page: Page, scheme: Mode): Promise<void> {
  await page.emulateMedia({ colorScheme: scheme, reducedMotion: "reduce" });
  await expect.poll(() => page.evaluate((s) => document.documentElement.classList.contains(s), scheme), { message: `<html> class follows OS ${scheme}` }).toBe(true);
  await frames(page);
  await settledPage(page);
}

// Edits one variable directly on an element, then lets the change reach style and paint (two frames, then the kit's
// colour transitions it starts) before anything is measured: a computed-style read in the same turn can still return
// the old value.
async function setSource(page: Page, el: Locator, name: string, value: string): Promise<void> {
  await el.evaluate((e, [n, v]) => (e as HTMLElement).style.setProperty(n, v), [name, value] as const);
  await frames(page);
  await settledPage(page);
  // The declaration is read back from the element, so a consumer that keeps an old value is told apart from an edit
  // that did not land or did not survive the two frames.
  expect(await el.evaluate((e, n) => (e as HTMLElement).style.getPropertyValue(n), name), `${name}: the inline declaration after the edit`).toBe(value);
}

// The emphasis variable as one element resolves it, for the message of an ink assertion.
async function resolvedGold(el: Locator): Promise<string> {
  return el.evaluate((e) => getComputedStyle(e).getPropertyValue("--vx-gold").trim());
}

// The surface's rendered ground just inside its top-left corner (the root surface is transparent over the body's
// background; a preview paints its own). A surface can be taller than the viewport, so its top edge is scrolled into
// view first (scrollIntoViewIfNeeded may leave it above the viewport).
async function surfaceOf(page: Page, label: string): Promise<RGB> {
  const el = surfaceEl(page, label);
  await el.evaluate((e) => e.scrollIntoView({ block: "start" }));
  const b = await el.boundingBox();
  if (!b) throw new Error(`${label}: surface has no box`);
  return pixelAt(page, b.x + 3, b.y + 3);
}

// The resolved colour of a CSS colour expression (typed tokens in, the browser's own colour-mix out), read
// through a detached probe's computed style; an invalid expression leaves it empty and fails cssToRgba.
async function resolveColor(page: Page, expr: string): Promise<RGBA> {
  const computed = await page.evaluate((e) => {
    const probe = document.createElement("div");
    probe.style.backgroundColor = e;
    document.body.appendChild(probe);
    const c = getComputedStyle(probe).backgroundColor;
    probe.remove();
    return c;
  }, expr);
  return cssToRgba(page, computed);
}

// A token at an opacity modifier (`bg-<token>/<percent>`: color-mix in oklab with transparent, as Tailwind
// compiles it) composited over the rendered ground.
async function alphaOver(page: Page, token: string, percent: number, ground: RGB): Promise<RGB> {
  return over(await resolveColor(page, `color-mix(in oklab, ${token} ${percent}%, transparent)`), ground);
}

// A computed colour as painted over a ground (identity for an opaque colour).
async function paintedOver(page: Page, value: string, ground: RGB): Promise<RGB> {
  return over(await cssToRgba(page, value), ground);
}

// Where a portaled node landed: inside a preview or a fixture surface (never expected), in the body, and the parent of
// Base UI's portal host (the body for a body portal).
async function placement(el: Locator) {
  return el.evaluate((n) => ({
    preview: !!n.closest("[data-brand-preview]"),
    surface: n.closest("[data-fixture-surface]")?.getAttribute("data-fixture-surface") ?? null,
    inBody: document.body.contains(n),
    host: (() => {
      const host = n.closest("[data-base-ui-portal]");
      return host === null ? null : host.parentElement === document.body ? "body" : "other";
    })(),
  }));
}

// The precondition of the "no preview value" negatives: every preview's popover and popover-foreground are
// distinguishable (> 8 per channel) from the root tenant's in the given appearance (typed tables only).
async function expectPreviewsDistinct(page: Page, tenant: string, mode: Mode) {
  const root = { tenant, label: ROOT };
  for (const p of PREVIEWS) {
    for (const name of ["popover", "popover-foreground"] as const) {
      const a = await expected(page, root, mode, name);
      const b = await expected(page, { tenant, label: p }, mode, name);
      expect(distance(a, b), `precondition: preview ${p} ${name} ${hex(b)} is distinguishable from the ${tenant} root's ${hex(a)} (${mode})`).toBeGreaterThan(8);
    }
  }
}

// The negative: a sampled portal colour is none of the previews' values for that slot.
async function expectNoPreviewValue(page: Page, tenant: string, mode: Mode, name: "popover" | "popover-foreground", got: RGB, what: string) {
  for (const p of PREVIEWS) {
    const v = await expected(page, { tenant, label: p }, mode, name);
    expect(distance(got, v), `${what}: ${hex(got)} is not preview ${p}'s ${name} ${hex(v)}`).toBeGreaterThan(2);
  }
}

test("[S-TOKENS] every root tenant palette and both isolated previews, on identical kit components", async ({ page }, testInfo) => {
  const mode = modeOf(testInfo);
  const primaries: Record<string, string> = {};
  for (const tenant of TENANTS) {
    await load(page, tenant);
    await expect(page.locator("[data-fixture-surface]")).toHaveCount(1 + PREVIEWS.length);
    await expect(page.locator("[data-brand-preview]")).toHaveCount(PREVIEWS.length);
    for (const label of tenant === DEFAULT_TENANT ? [ROOT, ...PREVIEWS] : [ROOT]) {
      const s = { tenant, label };
      const key = label === ROOT ? tenant : `preview ${label}`;
      const el = surfaceEl(page, label);
      const own = content(page, label);
      // The 15 names through their runtime variables (accent: --vx-gold, Plan §2.5), alpha included: the kit's Dark border
      // and input are translucent (G069 owner kit-fidelity direction: no opaque app input correction).
      for (const name of PALETTE_NAMES) {
        const value = await el.evaluate((e, n) => getComputedStyle(e).getPropertyValue(n).trim(), CSS_VARIABLE[name]);
        const want = expectedSource(s, mode, name);
        expect(distance4(await cssToRgba(page, value), await cssToRgba(page, want)), `${key} ${CSS_VARIABLE[name]}: ${value} vs ${want}`).toBeLessThanOrEqual(1);
      }
      const ground = await surfaceOf(page, label);
      near(ground, await expected(page, s, mode, "background"), `${key} rendered surface`);
      near(await colorOf(page, await computedColor(own.getByTestId("body-text"))), await expected(page, s, mode, "foreground"), `${key} text`);
      near(await colorOf(page, await computedColor(own.getByTestId("muted-text"))), await expected(page, s, mode, "muted-foreground"), `${key} muted text`);
      // The kit typography's inks: Heading foreground, Text lead foreground at 80 %, body and meta muted-foreground, the
      // Kicker and TextLink gold (saved accent).
      for (const testid of ["sample-h1", "sample-cta"]) near(await colorOf(page, await computedColor(own.getByTestId(testid))), await expected(page, s, mode, "foreground"), `${key} ${testid} (foreground)`);
      const fg = expectedSource(s, mode, "foreground");
      const bg = expectedSource(s, mode, "background");
      near(await paintedOver(page, await computedColor(own.getByTestId("sample-lead")), ground), hexToRgb(foreground80(fg, bg)), `${key} sample-lead (foreground at 80 %)`);
      for (const [testid, token] of [["sample-kicker", "accent"], ["sample-body", "muted-foreground"], ["sample-meta", "muted-foreground"], ["sample-text-link", "accent"]] as const) {
        near(await colorOf(page, await computedColor(own.getByTestId(testid))), await expected(page, s, mode, token), `${key} ${testid} (${token})`);
      }
      const action = own.getByTestId("button-default");
      near(await fillOf(page, action), await expected(page, s, mode, "primary"), `${key} action fill`);
      near(await colorOf(page, await computedColor(action)), await expected(page, s, mode, "primary-foreground"), `${key} action label`);
      await action.focus();
      await page.keyboard.press("Shift+Tab");
      await page.keyboard.press("Tab");
      await settledDeep(page, action);
      // The kit's focus state: the border takes the ring colour (focus-visible:border-ring).
      near(await colorOf(page, await computedColor(action, "border-top-color")), await expected(page, s, mode, "ring"), `${key} focus border colour`);
      await resetFocus(page);
      primaries[key] = hex(await fillOf(page, action));
    }
  }
  expect(primaries.branded, "the full branding palette differs from the platform default").not.toBe(primaries.platform);
  expect(primaries["dark-sparse-destructive"], "a sparse destructive override keeps the platform primary").toBe(primaries.platform);
  // Partial overrides: grey-primary is Light-only and keeps the platform Dark palette.
  if (mode === "light") expect(primaries["grey-primary"], "grey-primary differs from the default in Light").not.toBe(primaries.platform);
  else expect(primaries["grey-primary"], "grey-primary equals the default in Dark").toBe(primaries.platform);
  for (const p of PREVIEWS) expect(primaries[`preview ${p}`], `preview ${p} keeps its own primary beside the root tenant`).not.toBe(primaries.branded);
  expect(primaries["preview A"], "the two previews differ").not.toBe(primaries["preview B"]);
});

// --- S-PORTAL (G069 Plan §6.2 row 1): body portals with the root tenant's values -----------------------------

// The open dialog's surface, title ink, close cross, scrim and focus border equal the root tenant's values in mode m,
// and none of the previews' values.
type PortalValues = { mode: Mode; surface: string; title: string; cross: string; overlay: string; focusBorder: string };

async function expectPortalValues(page: Page, tenant: string, m: Mode): Promise<PortalValues> {
  const root = { tenant, label: ROOT };
  const popup = page.locator("[data-slot=dialog-content]");
  const pb = await box(popup);
  // The popup's 24px corners are rounded: its surface is probed at the bottom centre, inside the padding.
  const surface = await pixelAt(page, pb.x + pb.width / 2, pb.y + pb.height - 4);
  near(surface, await expected(page, root, m, "popover"), `${tenant} ${m} popup surface`);
  await expectNoPreviewValue(page, tenant, m, "popover", surface, `${tenant} ${m} popup surface`);
  const title = await colorOf(page, await computedColor(popup.locator("[data-slot=dialog-title]")));
  near(title, await expected(page, root, m, "popover-foreground"), `${tenant} ${m} popup title`);
  await expectNoPreviewValue(page, tenant, m, "popover-foreground", title, `${tenant} ${m} popup title`);
  // The kit DialogContent's own corner Close (showCloseButton; G069 owner kit-fidelity direction, 2026-10-26, replacing
  // the app's secondary-foreground substitute): a ghost Button on bg-secondary that sets no text colour, so its cross
  // inherits the popup's popover-foreground at rest, and the ghost variant's foreground on its muted hover fill.
  const close = popup.getByRole("button", { name: "Close", exact: true });
  const cross = await crossCentre(page, close);
  near(cross, await expected(page, root, m, "popover-foreground"), `${tenant} ${m} close cross (painted centre, the inherited popover-foreground)`, 8);
  await close.hover();
  await page.waitForTimeout(50);
  await settledDeep(page, close);
  near(await crossCentre(page, close), await expected(page, root, m, "foreground"), `${tenant} ${m} close cross on hover (painted centre)`, 8);
  await page.mouse.move(0, 0);
  await settledDeep(page, close);
  // The kit's theme-fixed scrim (bg-black/30), alpha included, in both appearances.
  const overlay = await page.locator("[data-slot=dialog-overlay]").evaluate((o) => getComputedStyle(o).backgroundColor);
  expect(distance4(await cssToRgba(page, overlay), await cssToRgba(page, SCRIM)), `${tenant} ${m} overlay ${overlay}`).toBeLessThanOrEqual(1);
  await close.focus();
  await page.keyboard.press("Shift+Tab");
  await page.keyboard.press("Tab");
  const focused = popup.locator(":focus");
  await settledDeep(page, focused);
  const focusBorder = await colorOf(page, await computedColor(focused, "border-top-color"));
  near(focusBorder, await expected(page, root, m, "ring"), `${tenant} ${m} focus border in the popup`);
  expect(haloOf(await computedColor(focused, "box-shadow")), `${tenant} ${m} focus halo in the popup`).not.toBeNull();
  return { mode: m, surface: hex(surface), title: hex(title), cross: hex(cross), overlay, focusBorder: hex(focusBorder) };
}

// The other body portals of the root surface: the kit Select and Popover popups (Base UI portals) and the kit Toaster's
// toast (rendered where the Toaster is, outside every preview).
const POPUPS = [
  { name: "select", trigger: "select-trigger", popup: "[data-slot=select-content]", basePortal: true },
  { name: "popover", trigger: "popover-trigger", popup: "[data-slot=popover-content]", basePortal: true },
  { name: "toast", trigger: "toast-trigger", popup: "[data-sonner-toast]", basePortal: false },
] as const;

test("[S-PORTAL] dialog, select, popover and toast land in the body with the root tenant's values, including across a live appearance change", async ({ page }, testInfo) => {
  const mode = modeOf(testInfo);
  const other: Mode = mode === "light" ? "dark" : "light";
  const evidence: Record<string, unknown> = {};
  for (const tenant of TENANTS) {
    for (const m of [mode, other]) await expectPreviewsDistinct(page, tenant, m);
    await load(page, tenant);
    const root = { tenant, label: ROOT };
    const trigger = content(page, ROOT).getByRole("button", { name: "Keyboard help" });
    await trigger.click();
    const popup = page.locator("[data-slot=dialog-content]");
    await expect(popup).toBeVisible();
    await dialogSettled(page);
    for (const node of [popup, page.locator("[data-slot=dialog-overlay]")]) {
      expect(await placement(node), `${tenant}: dialog inserted under the body's portal host, outside every surface`).toEqual({ preview: false, surface: null, inBody: true, host: "body" });
    }
    await popup.evaluate((p) => p.setAttribute("data-g069-probe", "dialog"));
    const live: PortalValues[] = [];
    for (const m of [mode, other, mode]) {
      await setScheme(page, m);
      live.push(await expectPortalValues(page, tenant, m));
    }
    await expect(page.locator("[data-slot=dialog-content][data-g069-probe]"), `${tenant}: the same dialog node throughout (never re-portalled)`).toHaveCount(1);
    await page.keyboard.press("Escape");
    await expect(popup).toBeHidden();
    // Escape returns focus to the trigger.
    await expect(trigger).toBeFocused();
    const others: Record<string, unknown> = {};
    for (const p of POPUPS) {
      await resetFocus(page);
      await content(page, ROOT).getByTestId(p.trigger).click();
      const el = page.locator(p.popup).first();
      await expect(el).toBeVisible();
      await settledDeep(page, el);
      const where = await placement(el);
      expect([where.preview, where.surface, where.inBody], `${tenant} ${p.name}: in the body, outside every preview and surface`).toEqual([false, null, true]);
      if (p.basePortal) expect(where.host, `${tenant} ${p.name}: under the body's portal host`).toBe("body");
      await el.evaluate((n) => n.setAttribute("data-g069-probe", "popup"));
      const samples: { mode: Mode; background: string; color: string }[] = [];
      for (const m of [mode, other, mode]) {
        await setScheme(page, m);
        await settledDeep(page, el);
        const st = await styleOf(el, ["background-color", "color"]);
        const background = await colorOf(page, st["background-color"]);
        const color = await colorOf(page, st.color);
        near(background, await expected(page, root, m, "popover"), `${tenant} ${p.name} ${m} surface`, 1);
        near(color, await expected(page, root, m, "popover-foreground"), `${tenant} ${p.name} ${m} text`, 1);
        await expectNoPreviewValue(page, tenant, m, "popover", background, `${tenant} ${p.name} ${m} surface`);
        await expectNoPreviewValue(page, tenant, m, "popover-foreground", color, `${tenant} ${p.name} ${m} text`);
        samples.push({ mode: m, background: hex(background), color: hex(color) });
      }
      await expect(page.locator(`${p.popup}[data-g069-probe]`), `${tenant} ${p.name}: the same node throughout`).toHaveCount(1);
      others[p.name] = samples;
      if (p.basePortal) {
        await page.keyboard.press("Escape");
        await expect(el).toBeHidden();
      }
    }
    evidence[tenant] = { dialog: live, ...others };
  }
  await testInfo.attach("portal.json", { body: JSON.stringify(evidence, null, 2), contentType: "application/json" });
});

// --- S-PORTAL-INIT (§5.4.1): insertion history, first frame and live mode -------------------------------

type PortalEntry = { seq: number; kind: "added" | "removed"; target: string; parentAtDelivery: string; slot: string | null; control: string | null };
type PortalSample = { when: "delivery" | "frame"; content: string; overlay: string; title: string };

// Installed before navigation: records, from MutationRecords only, every insertion or removal of a portal
// node with the record's own target (the parent at that mutation), plus first-frame and per-frame samples.
// Live references stay in the page; only a serialized, identity-based summary leaves it.
// Base UI first inserts its own empty portal host ([data-base-ui-portal]) into the container, then inserts
// the dialog nodes into that host. A record whose target is such a host is attributed to the parent the host was
// inserted into (from the host's own insertion record), so the oracle still sees exactly where the nodes landed;
// a host inserted into document.body classifies as "body". G069: the body is the expected target; an isolated preview
// element is where a misdirected portal would leak the preview's values.
async function installPortalRecorder(page: Page) {
  await page.addInitScript(() => {
    const PORTAL = '[data-slot="dialog-overlay"], [data-slot="dialog-content"]';
    const HOST = "[data-base-ui-portal]";
    type Raw = { seq: number; kind: "added" | "removed"; target: Node; parent: Node | null; node: Element; control: string | null };
    const hostParent = new WeakMap<Node, Node>();
    // The effective parent: a Base UI portal host stands for the parent it was inserted into.
    const via = (t: Node | null): Node | null => (t !== null && hostParent.has(t) ? (hostParent.get(t) ?? null) : t);
    const entries: Raw[] = [];
    const samples: { when: "delivery" | "frame"; content: string; overlay: string; title: string }[] = [];
    let seq = 0;
    let sampling = true;
    const controlOf = (n: Element) => n.closest("[data-g048-control]")?.getAttribute("data-g048-control") ?? null;
    const sample = (when: "delivery" | "frame") => {
      const c = document.querySelector('[data-slot="dialog-content"]:not([data-g048-control])');
      const o = document.querySelector('[data-slot="dialog-overlay"]:not([data-g048-control])');
      const t = c?.querySelector('[data-slot="dialog-title"]');
      if (!c || !o || !t) return;
      samples.push({ when, content: getComputedStyle(c).backgroundColor, overlay: getComputedStyle(o).backgroundColor, title: getComputedStyle(t).color });
    };
    const observer = new MutationObserver((records) => {
      let realAdded = false;
      for (const r of records) {
        for (const n of r.addedNodes) {
          if (!(n instanceof Element)) continue;
          if (n.matches(HOST)) {
            // The host is inserted empty; its dialog nodes arrive in later records targeting the host.
            hostParent.set(n, r.target);
            continue;
          }
          const found = [...(n.matches(PORTAL) ? [n] : []), ...n.querySelectorAll(PORTAL)];
          for (const node of found) {
            const control = controlOf(node);
            entries.push({ seq: seq++, kind: "added", target: via(r.target) ?? r.target, parent: via(node.parentNode), node, control });
            if (!control) realAdded = true;
          }
        }
        for (const n of r.removedNodes) {
          if (n instanceof Element && n.matches(PORTAL)) entries.push({ seq: seq++, kind: "removed", target: via(r.target) ?? r.target, parent: via(n.parentNode), node: n, control: n.getAttribute("data-g048-control") });
        }
      }
      // Before the rendering step that first paints the inserted nodes.
      if (realAdded && sampling) sample("delivery");
    });
    observer.observe(document, { childList: true, subtree: true });
    const frame = () => {
      if (!sampling) return;
      sample("frame");
      requestAnimationFrame(frame);
    };
    requestAnimationFrame(frame);
    (window as unknown as { __g048Portal: unknown }).__g048Portal = {
      stop: () => {
        sampling = false;
      },
      samples: () => samples,
      count: () => entries.length,
      summary: (from = 0) => {
        const classOf = (t: Node | null) =>
          t === null ? "detached" : t === document.body ? "body" : t instanceof Element && t.matches("[data-brand-preview]") ? "preview" : "other";
        return entries.slice(from).map((e) => ({
          seq: e.seq,
          kind: e.kind,
          target: classOf(e.target),
          parentAtDelivery: classOf(e.parent),
          slot: e.node.getAttribute("data-slot"),
          control: e.control,
        }));
      },
    };
  });
}

type Recorder = { stop: () => void; samples: () => PortalSample[]; count: () => number; summary: (from?: number) => PortalEntry[] };
type RecorderWindow = { __g048Portal: Recorder };

// Accessors for the in-page recorder; each returns only serialized data.
const portalLog = (page: Page) => ({
  stop: () => page.evaluate(() => (window as unknown as RecorderWindow).__g048Portal.stop()),
  count: () => page.evaluate(() => (window as unknown as RecorderWindow).__g048Portal.count()),
  samples: () => page.evaluate(() => (window as unknown as RecorderWindow).__g048Portal.samples()),
  summary: (from = 0) => page.evaluate((f) => (window as unknown as RecorderWindow).__g048Portal.summary(f), from),
});

// One same-task control mutation (PI-C1..PI-C4) in a single evaluate; returns the control's own entries. Each control
// passes a node through preview A's element (where a misdirected portal would leak), and the recorder must report
// exactly its expected entries.
async function runControl(page: Page, id: string): Promise<{ entries: PortalEntry[]; deliveryParent?: string }> {
  const log = portalLog(page);
  const from = await log.count();
  const deliveryParent = await page.evaluate(async (c) => {
    const preview = document.querySelector('[data-brand-preview][data-fixture-surface="A"]')!;
    const make = (slot: string) => {
      const n = document.createElement("div");
      n.setAttribute("data-slot", slot);
      n.setAttribute("data-g048-control", c);
      return n;
    };
    let probe: Element | null = null;
    if (c === "PI-C1") {
      preview.appendChild(make("dialog-overlay"));
    } else if (c === "PI-C2") {
      probe = make("dialog-overlay");
      preview.appendChild(probe);
      document.body.appendChild(probe);
    } else if (c === "PI-C3") {
      const wrapper = make("control-wrapper");
      wrapper.appendChild(make("dialog-content"));
      preview.appendChild(wrapper);
      document.body.appendChild(wrapper);
    } else if (c === "PI-C4") {
      const wrapper = make("control-wrapper");
      const node = make("dialog-content");
      wrapper.appendChild(node);
      preview.appendChild(wrapper);
      document.body.appendChild(node);
    }
    // Let the observer deliver (a microtask checkpoint, then a task).
    await new Promise((r) => setTimeout(r, 0));
    if (!probe) return undefined;
    return probe.parentNode === document.body ? "body" : probe.parentNode === preview ? "preview" : "other";
  }, id);
  const entries = (await log.summary(from)).filter((e) => e.control === id);
  return { entries, deliveryParent };
}

const CONTROL_EXPECTED: Record<string, Pick<PortalEntry, "kind" | "target" | "slot">[]> = {
  "PI-C1": [{ kind: "added", target: "preview", slot: "dialog-overlay" }],
  "PI-C2": [
    { kind: "added", target: "preview", slot: "dialog-overlay" },
    { kind: "removed", target: "preview", slot: "dialog-overlay" },
    { kind: "added", target: "body", slot: "dialog-overlay" },
  ],
  "PI-C3": [
    { kind: "added", target: "preview", slot: "dialog-content" },
    { kind: "added", target: "body", slot: "dialog-content" },
  ],
  "PI-C4": [
    { kind: "removed", target: "other", slot: "dialog-content" },
    { kind: "added", target: "body", slot: "dialog-content" },
  ],
};

for (const tenant of [DEFAULT_TENANT, PLATFORM]) {
  test(`[S-PORTAL-INIT] ?open=root (${tenant}): a dialog open at mount is inserted only into the body and first paints with the root tenant's values`, async ({ page }, testInfo) => {
    const mode = modeOf(testInfo);
    const other: Mode = mode === "light" ? "dark" : "light";
    const root = { tenant, label: ROOT };
    const evidence: Record<string, unknown> = { tenant, mode };
    // Preconditions (independent tables): the previews' values are distinguishable from the root's; a tenant's values are
    // distinguishable from the platform defaults (so the samples prove the tenant palette from first insertion).
    for (const m of [mode, other]) await expectPreviewsDistinct(page, tenant, m);
    if (tenant !== PLATFORM) {
      const platform = { tenant: PLATFORM, label: ROOT };
      const gaps = await Promise.all((["popover", "popover-foreground"] as const).map(async (n) => distance(await expected(page, root, mode, n), await expected(page, platform, mode, n))));
      expect(Math.max(...gaps), "the tenant's popup values are distinguishable from the platform defaults").toBeGreaterThan(2);
    }
    await page.emulateMedia({ reducedMotion: "reduce" });
    await installPortalRecorder(page);
    // 1. Navigation readiness (the dialog is already open and modal: no waitForHydration).
    const response = await page.goto(fixtureUrl(tenant, "&open=root"), { waitUntil: "load" });
    const popup = page.locator("[data-slot=dialog-content]");
    await expect(popup).toBeVisible();
    expect(await placement(popup), "the open dialog is under the body's portal host").toEqual({ preview: false, surface: null, inBody: true, host: "body" });
    // 2. SSR renders no portal.
    const html = await response!.text();
    expect(html.includes('data-slot="dialog-content"') || html.includes('data-slot="dialog-overlay"'), "no dialog in the server HTML").toBe(false);
    // 4 (sampling ends before any appearance change). 3. Body insertion.
    const log = portalLog(page);
    await log.stop();
    const initial = await log.summary();
    const real = initial.filter((e) => !e.control);
    evidence.entries = initial;
    expect(
      real.map(({ kind, target, slot }) => ({ kind, target, slot })).sort((a, b) => String(a.slot).localeCompare(String(b.slot))),
      "exactly one overlay and one content insertion, both into the body (through Base UI's portal host)",
    ).toEqual([
      { kind: "added", target: "body", slot: "dialog-content" },
      { kind: "added", target: "body", slot: "dialog-overlay" },
    ]);
    // 4. First frame: every delivery and frame sample has the root tenant's values and none of the previews'.
    const samples = await log.samples();
    evidence.samples = samples;
    expect(samples.filter((s) => s.when === "delivery").length, "a delivery-time sample").toBeGreaterThan(0);
    const popover = await expected(page, root, mode, "popover");
    const ink = await expected(page, root, mode, "popover-foreground");
    const scrim = await cssToRgba(page, SCRIM);
    for (const [i, s] of samples.entries()) {
      const c = await colorOf(page, s.content);
      const t = await colorOf(page, s.title);
      near(c, popover, `sample ${i} (${s.when}) content`);
      near(t, ink, `sample ${i} (${s.when}) title ink`);
      await expectNoPreviewValue(page, tenant, mode, "popover", c, `sample ${i} (${s.when}) content`);
      await expectNoPreviewValue(page, tenant, mode, "popover-foreground", t, `sample ${i} (${s.when}) title ink`);
      expect(distance4(await cssToRgba(page, s.overlay), scrim), `sample ${i} (${s.when}) scrim ${s.overlay}`).toBeLessThanOrEqual(1);
    }
    // 5. Live mode on the initially open dialog; it is never re-portalled.
    const before = await log.count();
    const live: PortalValues[] = [];
    for (const m of [mode, other, mode]) {
      await setScheme(page, m);
      live.push(await expectPortalValues(page, tenant, m));
    }
    evidence.live = live;
    const later = await log.summary(before);
    expect(later.filter((e) => e.kind === "added"), "no re-portalling across appearance changes").toEqual([]);
    // 6. Modal and close.
    await expect.poll(() => popup.evaluate((p) => p.contains(document.activeElement)), { message: "focus is inside the dialog" }).toBe(true);
    for (const key of ["Tab", "Tab", "Shift+Tab"]) {
      await page.keyboard.press(key);
      await frames(page);
      expect(await popup.evaluate((p) => p.contains(document.activeElement)), `focus stays inside after ${key}`).toBe(true);
    }
    const beforeClose = await log.count();
    await page.keyboard.press("Escape");
    await expect(popup).toBeHidden();
    await expect(page.locator("[data-slot=dialog-overlay]")).toHaveCount(0);
    const removed = (await log.summary(beforeClose)).filter((e) => !e.control);
    evidence.removed = removed;
    expect(removed.map(({ kind, target, slot }) => ({ kind, target, slot })).sort((a, b) => String(a.slot).localeCompare(String(b.slot)))).toEqual([
      { kind: "removed", target: "body", slot: "dialog-content" },
      { kind: "removed", target: "body", slot: "dialog-overlay" },
    ]);
    // 7. Controls: each must be detected with exactly its expected entries.
    const controls: Record<string, unknown> = {};
    try {
      for (const id of ["PI-C1", "PI-C2", "PI-C3", "PI-C4"]) {
        const outcome = await runControl(page, id);
        controls[id] = outcome;
        expect(outcome.entries.map(({ kind, target, slot }) => ({ kind, target, slot })), `${id} detected exactly`).toEqual(CONTROL_EXPECTED[id]);
        if (id === "PI-C2") {
          expect(outcome.deliveryParent, "PI-C2: the node's parent at delivery is the body").toBe("body");
          expect(outcome.entries[0].parentAtDelivery, "PI-C2: a current-parent recorder would have missed the preview insertion").toBe("body");
        }
      }
    } finally {
      await page.evaluate(() => document.querySelectorAll("[data-g048-control]").forEach((n) => n.remove()));
      evidence.controls = controls;
      await testInfo.attach(`portal-init-${tenant}.json`, { body: JSON.stringify(evidence, null, 2), contentType: "application/json" });
    }
  });
}

// --- S-RUNTIME --------------------------------------------------------------------------------------------

async function panelImage(page: Page, label: string): Promise<Image> {
  return decodePng(await content(page, label).getByTestId("blueprint-panel").screenshot({ animations: "disabled" }));
}

// The blueprint panel's strokes lie on the line from the surface's ground toward the expected composite
// (muted-foreground at 0.5 over the ground): residual <= 4, nothing beyond the composite, and strokes present.
type StrokeFit = { strokes: number; maxToward: number; worstResidual: number; composite: string; ground: string };
function fitStrokes(img: Image, ground: RGB, mutedForeground: RGB): StrokeFit {
  const composite = over([...mutedForeground, 128] as RGBA, ground);
  const d = [0, 1, 2].map((i) => composite[i] - ground[i]);
  const len = d.reduce((s, v) => s + v * v, 0);
  let strokes = 0;
  let maxToward = 0;
  let worstResidual = 0;
  for (const p of img.pixels) {
    const t = len === 0 ? 0 : [0, 1, 2].reduce((s, i) => s + (p[i] - ground[i]) * d[i], 0) / len;
    if (t < 0.2) continue;
    strokes++;
    maxToward = Math.max(maxToward, t);
    const q = [0, 1, 2].map((i) => ground[i] + t * d[i]) as RGB;
    worstResidual = Math.max(worstResidual, Math.max(...[0, 1, 2].map((i) => Math.abs(p[i] - q[i]))));
  }
  return { strokes, maxToward: Number(maxToward.toFixed(3)), worstResidual: Number(worstResidual.toFixed(2)), composite: hex(composite), ground: hex(ground) };
}

function expectFit(fit: StrokeFit, what: string) {
  expect(fit.strokes, `${what}: stroke pixels`).toBeGreaterThanOrEqual(50);
  expect(fit.maxToward, `${what}: strokes reach the structure composite`).toBeGreaterThanOrEqual(0.5);
  expect(fit.maxToward, `${what}: nothing painted beyond the composite`).toBeLessThanOrEqual(1.1);
  expect(fit.worstResidual, `${what}: every stroke pixel is the surface's muted-foreground over its ground`).toBeLessThanOrEqual(4);
}

// G069 (Plan §6.2): a preview's values are inline variables on its element (S8), so a runtime edit there applies at once,
// to that preview only, whatever the document's appearance (a preview has one appearance).
test("[S-RUNTIME] a changed preview value applies immediately, to that preview only, without a rebuild", async ({ page }, testInfo) => {
  await load(page);
  const a = surfaceEl(page, "A");
  const others = [ROOT, "B"];
  const fills = async (scheme: Mode) => {
    await setScheme(page, scheme);
    const out: Record<string, string> = {};
    for (const label of others) out[label] = hex(await fillOf(page, content(page, label).getByTestId("button-default")));
    return out;
  };
  const before = { light: await fills("light"), dark: await fills("dark") };
  await setScheme(page, "light");
  await setSource(page, a, "--primary", "#123456");
  near(await fillOf(page, content(page, "A").getByTestId("button-default")), [0x12, 0x34, 0x56], "A primary after the runtime change");
  await setScheme(page, "dark");
  near(await fillOf(page, content(page, "A").getByTestId("button-default")), [0x12, 0x34, 0x56], "A keeps its changed primary under a Dark document (one appearance per preview)");
  const after = { light: await fills("light"), dark: await fills("dark") };
  expect(after, "the root tenant and preview B are unchanged in both appearances").toEqual(before);
  expect(new Set(Object.values(before.light)).size + new Set(Object.values(before.dark)).size, "the compared fills are real, distinct samples").toBeGreaterThan(2);

  // The emphasis ink (saved accent, --vx-gold): an edit on A changes A's Kicker and TextLink only.
  await setScheme(page, "light");
  const inkOf = async (label: string, testid: string) => colorOf(page, await computedColor(content(page, label).getByTestId(testid)));
  const inkBefore = Object.fromEntries(await Promise.all(others.map(async (l) => [l, hex(await inkOf(l, "sample-kicker"))] as const)));
  await setSource(page, a, "--vx-gold", "#2a1a3a");
  // Measured (G069 row 3): the Kicker's computed colour read the old value after the two frames while the variable already
  // resolved to the new one, and read the new one a moment later. Each ink is therefore read until it settles, within a
  // second, and then compared exactly (G069: no app reduced-motion rule; the kit's own transitions run in full).
  const inkSettles = async (label: string, testid: string, want: RGB, what: string) => {
    await expect.poll(async () => distance(await inkOf(label, testid), want), { message: what, timeout: 1000 }).toBeLessThanOrEqual(2);
    near(await inkOf(label, testid), want, what);
  };
  await inkSettles("A", "sample-kicker", [0x2a, 0x1a, 0x3a], `A Kicker after the --vx-gold change (the preview resolves --vx-gold as ${await resolvedGold(a)})`);
  await inkSettles("A", "sample-text-link", [0x2a, 0x1a, 0x3a], "A TextLink after the --vx-gold change");
  for (const l of others) expect(hex(await inkOf(l, "sample-kicker")), `${l} Kicker unchanged`).toBe(inkBefore[l]);

  // The blueprint layer follows a runtime muted-foreground change in A only (§8.5 item 7).
  const panelsBefore = Object.fromEntries(await Promise.all(others.map(async (l) => [l, await content(page, l).getByTestId("blueprint-panel").screenshot({ animations: "disabled" })] as const)));
  await setSource(page, a, "--muted-foreground", "#1b4d7a");
  const fit = fitStrokes(await panelImage(page, "A"), await expected(page, { tenant: DEFAULT_TENANT, label: "A" }, "light", "background"), [0x1b, 0x4d, 0x7a]);
  expectFit(fit, "A blueprint after the runtime change");
  for (const l of others) {
    const now = await content(page, l).getByTestId("blueprint-panel").screenshot({ animations: "disabled" });
    expect(now.equals(panelsBefore[l]), `${l} blueprint panel unchanged`).toBe(true);
  }
  await testInfo.attach("runtime-blueprint.json", { body: JSON.stringify({ A: fit }, null, 2), contentType: "application/json" });
});

// The axe color-contrast findings the owner waived (G069 Spec §8; owner instruction 2026-10-26, "exact kit, overrides
// included"): the kit's white label on the platform primary #0075ff (4.21 in both appearances) on every default-variant
// control of the root surface, and the kit's Light platform destructive text on its 10 % tint (3.99). Each applies only
// where the root renders that platform value, and each finding must report the kit's ratio (±WAIVED_TOLERANCE); any
// other finding, of any rule, still fails.
const AXE_WAIVED_PRIMARY = ["button-default", "button-size-sm", "button-size-default", "button-size-lg", "form-submit", "dialog-action", "cta-link"];
// The waived kit ratio for a flagged root-surface element (by its test id and whether it sits in a preview), if any.
function axeWaivedRatio(tenant: string, mode: Mode, testid: string | null, inPreview: boolean): number | undefined {
  const root = { tenant, label: ROOT };
  if (!testid || inPreview) return undefined;
  if (AXE_WAIVED_PRIMARY.includes(testid) && expectedSource(root, mode, "primary") === SOURCES[mode].primary) return 4.21;
  if (testid === "button-destructive" && mode === "light" && expectedSource(root, mode, "destructive") === SOURCES.light.destructive) return 3.99;
  return undefined;
}

test("[S-A11Y] fixtures: no automated WCAG 2.2 AA violations under any root tenant, including the open dialog", async ({ page }, testInfo) => {
  const mode = modeOf(testInfo);
  const tags = ["wcag2a", "wcag2aa", "wcag21a", "wcag21aa", "wcag22aa"];
  const waived: { tenant: string; label: string; node: string; ratio: number; kit: number; waiver: string }[] = [];
  const check = async (tenant: string, label: string) => {
    const r = await new AxeBuilder({ page }).withTags(tags).analyze();
    const failing: string[] = [];
    for (const v of r.violations) {
      const left: string[] = [];
      for (const n of v.nodes) {
        const node = n.target.join(" ");
        if (v.id === "color-contrast") {
          // The flagged element itself (axe's selector for a node outside any shadow root), for its test id and surface.
          const where = await page.evaluate((sel) => {
            const el = document.querySelector(sel);
            return { testid: el?.getAttribute("data-testid") ?? null, inPreview: !!el?.closest("[data-brand-preview]") };
          }, String(n.target[0]));
          const kit = axeWaivedRatio(tenant, mode, where.testid, where.inPreview);
          const ratio = (n.any[0]?.data as { contrastRatio?: unknown } | undefined)?.contrastRatio;
          if (kit !== undefined && typeof ratio === "number" && Math.abs(ratio - kit) <= WAIVED_TOLERANCE) {
            waived.push({ tenant, label, node, ratio, kit, waiver: OWNER_WAIVED });
            continue;
          }
        }
        left.push(node);
      }
      if (left.length) failing.push(`${v.id}: ${left.join(", ")}`);
    }
    expect(failing, label).toEqual([]);
  };
  for (const tenant of TENANTS) {
    await load(page, tenant);
    await check(tenant, `${tenant} fixture page`);
    await content(page, ROOT).getByRole("button", { name: "Keyboard help" }).click();
    await expect(page.getByRole("dialog")).toBeVisible();
    await dialogSettled(page);
    await check(tenant, `${tenant} dialog open`);
    await page.keyboard.press("Escape");
    await expect(page.getByRole("dialog")).toBeHidden();
  }
  await testInfo.attach("a11y-owner-waived.json", { body: JSON.stringify(waived, null, 2), contentType: "application/json" });
});

// A pair exempt from contrast minimums, recorded without a threshold: a disabled control's composited label and fill
// (inactive user interface component), or a text button's hairline (its label identifies the control).
function exempt(name: string, a: RGB, b: RGB, reason = "inactive component"): PairResult & { exempt: string } {
  return { name, ratio: Number(contrast(a, b).toFixed(2)), fg: hex(a), bg: hex(b), min: 0, exempt: reason };
}

// Every surface the contrast and invalid-field checks visit: each tenant's root, and the previews on the default page.
function* surfacesOf(tenant: string): Generator<Surface> {
  yield { tenant, label: ROOT };
  if (tenant === DEFAULT_TENANT) for (const label of PREVIEWS) yield { tenant, label };
}

// The S-CONTRAST rows are gathered (G069 P0.2 correction round 3): a row whose own assertion fails is recorded and the
// walk goes on, so one run reports every failing pair, and the test then fails once listing them all. Only assertion
// failures are gathered (a Playwright expect message carries its `expect(` matcher hint); any other error, such as a
// missing element or a closed page, still stops the test at once. Every row keeps its threshold, or, where the kit's
// own treatment paints below it, is OWNER-WAIVED and asserted at the kit's exact ratio (support/modes.ts expectAtKit).
const ANSI = new RegExp(`${String.fromCharCode(27)}\\[[0-9;]*m`, "g");
function rowGatherer(results: PairResult[], failures: string[]) {
  return async (row: () => PairResult | Promise<PairResult>): Promise<void> => {
    try {
      results.push(await row());
    } catch (e) {
      const message = (e instanceof Error ? e.message : String(e)).replace(ANSI, "");
      if (!message.includes("expect(")) throw e;
      failures.push(message.split("\n").map((l) => l.trim()).filter(Boolean).join(" | "));
    }
  };
}

// The S-CONTRAST rows the owner waived (G069 Spec §8; owner instruction 2026-10-26, "exact kit, overrides included"), per
// project appearance: exactly these rows paint below their former threshold with the kit's own treatment on the fixture
// palettes, and each is asserted at its kit ratio (expectAtKit); every other row keeps its threshold, and a row that
// falls below it without being listed here fails. Desk-computed from the typed tables (sRGB composite at 8 bits). The
// kit's platform ratios: white on #0075ff 4.21, on its Light 80 % hover 3.18 (page and popover); Light destructive text on
// its 10 % / 20 % tints 3.99 / 3.32; Dark destructive on its 30 % tint 4.33; the kit's destructive focus border
// (focus-visible:border-destructive/40; the app's solid N5 border is withdrawn) against the page about 2.1 Light / 2.0
// Dark, and 2.0 to 2.4 on every fixture palette; the dialog field boundary (N3: input over the popover) 1.26 Light / 1.59
// Dark; the kit Close's inherited popover-foreground on the branded mid-tone secondary 2.30 Light / 1.81 Dark. Preview B
// keeps the platform Dark destructive on its own ground (hover 3.88). Kept at their thresholds, for example: the Dark
// 80 % hover over the page 5.89 and over the popover 5.73, Dark destructive at rest 5.26, preview A's hover 4.63, the
// grey primary's hover 4.61.
const CONTRAST_WAIVED: Record<Mode, string[]> = (() => {
  const rows = (key: string, list: string[]) => list.map((r) => `${key} ${r}`);
  const focusBorder = "destructive focus border against the page";
  const platformLight = ["default at rest", "default hover", "default pressed", "default focus", "destructive at rest", "destructive hover", "destructive pressed", "destructive focus", focusBorder, "dialog action at rest", "dialog action hover", "dialog field boundary"];
  const platformDark = ["default at rest", "default focus", "destructive hover", "destructive pressed", focusBorder, "dialog action at rest", "dialog field boundary"];
  const shared = [...rows("branded", [focusBorder, "close cross"]), ...rows("preview A", [focusBorder]), ...rows("preview B", ["destructive hover", "destructive pressed", focusBorder])];
  return {
    light: [
      ...rows("platform", platformLight),
      ...shared,
      ...rows("grey-primary", ["destructive at rest", "destructive hover", "destructive pressed", "destructive focus", focusBorder, "dialog field boundary"]),
      ...rows("dark-sparse-destructive", platformLight),
    ],
    dark: [...rows("platform", platformDark), ...shared, ...rows("grey-primary", platformDark), ...rows("dark-sparse-destructive", ["default at rest", "default focus", focusBorder, "dialog action at rest", "dialog field boundary"])],
  };
})();

// The kit destructive variant's keyboard focus border (focus-visible:border-destructive/40 over the page, bg-clip-padding;
// the app's solid N5 border is withdrawn) against the page beyond its halo, as painted on the side whose page sample is
// the surface, compared with the kit's destructive at 40 % over the surface (typed tables).
async function destructiveFocusBorder(page: Page, name: string, button: Locator, kitBorder: RGB, surface: RGB): Promise<KitPairResult> {
  await settledDeep(page, button);
  const st = await styleOf(button, ["box-shadow", "border-top-color"]);
  const halo = haloOf(st["box-shadow"]);
  expect(halo, `${name}: focus ring (${st["box-shadow"]})`).not.toBeNull();
  const border = await cssToRgba(page, st["border-top-color"]);
  const present = (await compositeSamples(page, button, halo!.spread)).filter((p) => distance(p.border, over(border, p.page)) <= 40);
  expect(present.length, `${name}: the focus border ${st["border-top-color"]} present on a side`).toBeGreaterThan(0);
  const best = present.reduce((a, p) => (distance(p.page, surface) < distance(a.page, surface) ? p : a));
  return expectAtKit(name, best.border, best.page, kitBorder, surface, 3);
}

test("[S-CONTRAST] fixtures: every root tenant and preview, every variant, at rest, hover, press and focus; the dialog's text, field and action", async ({ page }, testInfo) => {
  const mode = modeOf(testInfo);
  const results: PairResult[] = [];
  const failures: string[] = [];
  const check = rowGatherer(results, failures);
  for (const tenant of TENANTS) {
    await load(page, tenant);
    for (const s of surfacesOf(tenant)) {
      const key = s.label === ROOT ? tenant : `preview ${s.label}`;
      const own = content(page, s.label);
      const surface = await surfaceOf(page, s.label);
      const tok = (name: PaletteName) => expected(page, s, mode, name);
      await check(async () => expectText(page, `${key} body`, own.getByTestId("body-text"), surface));
      await check(async () => expectText(page, `${key} muted`, own.getByTestId("muted-text"), surface));
      const link = own.getByTestId("link");
      await check(async () => expectText(page, `${key} link`, link, surface));
      await link.hover();
      await check(async () => expectText(page, `${key} link hover`, link, surface));
      await page.mouse.move(0, 0);
      for (const variant of VARIANTS) {
        const button = own.getByTestId(`button-${variant}`);
        // The kit's look of the variant on this surface from the typed tables (S-STATES): the kit pair of each label row.
        const look = await lookOf(page, s, mode, variant, surface);
        // The backing of a label is the control's rendered fill (the surface itself for the transparent variants); the label
        // as painted on it is held to 4.5:1 where the kit's own pair meets it, else to the kit's ratio (OWNER-WAIVED).
        const label = (name: string, kitLabel: RGB, kitFill: RGB) =>
          check(async () => {
            const fill = await fillOf(page, button);
            return expectAtKit(name, await textColor(page, button, fill), fill, kitLabel, kitFill, 4.5);
          });
        await label(`${key} ${variant} at rest`, look.label, look.rest);
        // No 3:1 boundary-against-the-surface requirement for text buttons: not for the filled ones (G017 Plan r2
        // withdrew N1: their labels identify them) and, since G069 P0.2, not for the outline's 1px hairline either (the
        // unaltered kit's border-border paints 1.26:1 on the platform page; WCAG 1.4.11 requires no boundary contrast
        // when the text label identifies the control, and Core r3 has no border pair). The hairline is recorded as
        // evidence; every hover, pressed and focus state stays checked.
        if (variant === "outline") {
          const hairline = await paintedOver(page, await computedColor(button, "border-top-color"), surface);
          results.push(exempt(`${key} ${variant} boundary (exempt: text label identifies the control)`, hairline, surface, "text label identifies the control"));
        }
        // Hover and pressed: the kit's alpha states (hover:bg-primary/80, the secondary 5 % mix, the destructive tint, the
        // outline/ghost muted fills), measured as painted against the rendered fill.
        await button.hover();
        await page.waitForTimeout(50);
        await settledDeep(page, button);
        await label(`${key} ${variant} hover`, look.hoverLabel, look.hover);
        if ((FILLED as readonly string[]).includes(variant)) {
          // Pressed: mouse down on the hovered control, sample, release (the hover fill stays: the kit has no active fill).
          await page.mouse.down();
          await page.waitForTimeout(50);
          await settledDeep(page, button);
          await label(`${key} ${variant} pressed`, look.label, look.hover);
          await page.mouse.up();
        }
        await page.mouse.move(0, 0);
        await button.focus();
        await page.keyboard.press("Shift+Tab");
        await page.keyboard.press("Tab");
        // The kit's 150ms transition back from the hover fill (no app reduced-motion rule since G069) runs out first.
        await settledDeep(page, button);
        await label(`${key} ${variant} focus`, look.label, look.rest);
        if (variant === "destructive") {
          const kitBorder = await alphaOver(page, expectedSource(s, mode, "destructive"), 40, surface);
          await check(async () => destructiveFocusBorder(page, `${key} destructive focus border against the page`, button, kitBorder, surface));
        } else await check(async () => expectFocusRing(page, `${key} ${variant}`, button));
        await resetFocus(page);
        // Disabled (opacity 0.5): the rendered fill, and the label as painted at half opacity over the surface.
        const disabled = own.getByTestId(`button-${variant}-disabled`);
        const dfill = await fillOf(page, disabled);
        const dlabel = await cssToRgba(page, await computedColor(disabled));
        results.push(exempt(`${key} ${variant} disabled (exempt: inactive component)`, over([dlabel[0], dlabel[1], dlabel[2], Math.round(dlabel[3] * 0.5)], surface), dfill));
      }
      if (s.label !== ROOT) continue;
      // The kit Table's header and cell text on the surface.
      const table = own.getByTestId("fixture-table");
      await check(async () => expectText(page, `${key} table header`, table.locator("th").first(), surface));
      await check(async () => expectText(page, `${key} table cell`, table.locator("td").first(), surface));
      await own.getByRole("button", { name: "Keyboard help" }).click();
      const popup = page.locator("[data-slot=dialog-content]");
      await expect(popup).toBeVisible();
      await dialogSettled(page);
      const pb = await box(popup);
      const popupBg = await pixelAt(page, pb.x + pb.width / 2, pb.y + pb.height - 4);
      await check(async () => expectText(page, `${key} dialog title`, popup.locator("[data-slot=dialog-title]"), popupBg));
      await check(async () => expectText(page, `${key} dialog description`, popup.locator("[data-slot=dialog-description]"), popupBg));
      // The kit DialogContent's own corner Close (G069 owner kit-fidelity direction): its cross is drawn in currentColor, the
      // inherited popover-foreground, on bg-secondary at rest, and in foreground on the ghost hover fill (Light muted, Dark
      // muted at 50 % over the popover). The ink is the computed colour (S-PORTAL proves the painted cross matches it), the
      // fill as rendered; the kit pair from the typed tables.
      const close = popup.getByRole("button", { name: "Close", exact: true });
      const popover = await tok("popover");
      const cross = (name: string, kitInk: RGB, kitFill: RGB) =>
        check(async () => {
          const fill = await fillOf(page, close);
          return expectAtKit(name, await textColor(page, close, fill), fill, kitInk, kitFill, 3);
        });
      await cross(`${key} close cross`, await tok("popover-foreground"), await tok("secondary"));
      await close.hover();
      await page.waitForTimeout(50);
      await settledDeep(page, close);
      await cross(`${key} close cross on hover`, await tok("foreground"), mode === "dark" ? await alphaOver(page, expectedSource(s, mode, "muted"), 50, popover) : await tok("muted"));
      await page.mouse.move(0, 0);
      await settledDeep(page, close);
      // T8 over the popover: the default action at rest and at its 80 % hover over the popup surface.
      const action = popup.getByTestId("dialog-action");
      const actionLabel = (name: string, kitFill: RGB) =>
        check(async () => {
          const fill = await fillOf(page, action);
          return expectAtKit(name, await textColor(page, action, fill), fill, await tok("primary-foreground"), kitFill, 4.5);
        });
      await actionLabel(`${key} dialog action at rest`, await tok("primary"));
      await action.hover();
      await settledDeep(page, action);
      await actionLabel(`${key} dialog action hover`, await alphaOver(page, expectedSource(s, mode, "primary"), 80, popover));
      await page.mouse.move(0, 0);
      await settledDeep(page, action);
      // T12 over the popover: the field's text on its 50 % input fill. N3: the kit field draws no boundary of its own
      // (border-transparent; the app's solid input boundary is withdrawn), so the boundary pair is Core's model, the
      // page's --input at the field composited over the rendered popup surface, against that surface.
      const field = popup.getByTestId("dialog-input");
      const fieldFill = await fillOf(page, field);
      await check(async () => expectText(page, `${key} dialog field text`, field, fieldFill));
      // The dialog opens with the field focused (focus-visible:border-ring); the kit's resting border is read unfocused.
      await field.evaluate((e) => (e as HTMLElement).blur());
      await settledDeep(page, field);
      const fieldStyle = await styleOf(field, ["border-top-color", "--input"]);
      expect((await cssToRgba(page, fieldStyle["border-top-color"]))[3], `${key} dialog field: the kit's border-transparent (${fieldStyle["border-top-color"]})`).toBe(0);
      const kitInput = await paintedOver(page, expectedSource(s, mode, "input"), popover);
      await check(async () => expectAtKit(`${key} dialog field boundary`, await paintedOver(page, fieldStyle["--input"].trim(), popupBg), popupBg, kitInput, popover, 3));
      await page.keyboard.press("Escape");
      await expect(popup).toBeHidden();
    }
  }
  const waivedRows = results.filter((r) => (r as KitPairResult).kitWaiver).map((r) => r.name);
  await testInfo.attach("contrast.json", { body: JSON.stringify(results, null, 2), contentType: "application/json" });
  await testInfo.attach("contrast-failures.json", { body: JSON.stringify(failures, null, 2), contentType: "application/json" });
  expect(failures, `${failures.length} contrast row(s) below their minimum or off their kit ratio`).toEqual([]);
  // The OWNER-WAIVED inventory: exactly the listed rows sit below their former threshold, each at its kit ratio.
  expect([...waivedRows].sort(), `the ${OWNER_WAIVED} rows (${mode})`).toEqual([...CONTRAST_WAIVED[mode]].sort());
});

// The S-FIELD-INVALID rows the owner waived (G069 Spec §8; owner instruction 2026-10-26, "exact kit, overrides
// included"), per project appearance. With the app's dark:aria-invalid:border-destructive withdrawn, a Dark surface draws
// the kit's half-alpha invalid border (dark:aria-invalid:border-destructive/50) over its fill; Light surfaces keep the
// kit's solid aria-invalid:border-destructive and meet 3:1 (platform Light 4.77 / 3.31 / 4.27 against ground / halo /
// fill). Desk-computed from the typed tables (sRGB composite at 8 bits): platform Dark (also grey-primary's Dark) 2.80 /
// 1.43 / 2.41; preview B (Dark platform destructive on its own ground) 2.81 / 1.39 / 2.32; branded Dark 5.21 (kept) /
// 2.15 / 1.89; dark-sparse-destructive Dark 3.62 (kept) / 1.50 / 3.11 (kept).
const FIELD_INVALID_WAIVED: Record<Mode, string[]> = (() => {
  const rows = (key: string, against: string[]) => against.map((a) => `${key} invalid field boundary against ${a}`);
  const all = ["the ground", "its halo", "its fill"];
  const b = rows("preview B", all);
  return {
    light: b,
    dark: [...rows("platform", all), ...rows("branded", ["its halo", "its fill"]), ...b, ...rows("grey-primary", all), ...rows("dark-sparse-destructive", ["its halo"])],
  };
})();

// G067 correction (Codex implementation review 2026-10-25, item 1; Plan §5.2 / Spec R4), retyped to the bare kit field by
// the G069 owner kit-fidelity direction (2026-10-26): a blurred field with aria-invalid draws the kit's invalid border, the
// solid destructive token in Light and destructive at 50 % in Dark (the dark variant of the surface's own appearance),
// over its fill, which is input at 50 % over the surface; the kit draws its invalid ring (3px, destructive at 20 %, Dark
// 40 %) outside the border at rest. The border is measured against the ground beyond the halo, against the halo and
// against its fill, under every root palette and preview in the project's appearance: 3:1 where the kit's own rendering
// meets it, else at the kit's ratio (OWNER-WAIVED, FIELD_INVALID_WAIVED).
test("[S-FIELD-INVALID] blurred invalid field: the kit's invalid boundary at ≥3:1, or at its waived kit ratio, under valid overrides", async ({ page }, testInfo) => {
  const mode = modeOf(testInfo);
  const results: KitPairResult[] = [];
  for (const tenant of TENANTS) {
    await load(page, tenant);
    for (const s of surfacesOf(tenant)) {
      const key = s.label === ROOT ? tenant : `preview ${s.label}`;
      const field = content(page, s.label).getByTestId("input-invalid");
      await expect(field).toHaveAttribute("aria-invalid", "true");
      await resetFocus(page);
      await page.mouse.move(0, 0);
      const surface = await surfaceOf(page, s.label);
      await field.scrollIntoViewIfNeeded();
      await settledDeep(page, field);
      expect(await field.evaluate((el) => document.activeElement !== el), `${key} invalid field is not focused`).toBe(true);
      const dark = darkVariants(s, mode);
      const fill = await fillOf(page, field);
      const kitFill = await alphaOver(page, expectedSource(s, mode, "input"), 50, surface);
      near(fill, kitFill, `${key} invalid field fill (input at 50 % over the surface)`);
      const border = over(await cssToRgba(page, await computedColor(field, "border-top-color")), fill);
      const kitBorder = await alphaOver(page, expectedSource(s, mode, "destructive"), dark ? 50 : 100, kitFill);
      near(border, kitBorder, `${key} invalid field border: the kit's ${dark ? "destructive at 50 %" : "solid destructive"} over the fill`);
      const b = await box(field);
      const halo = await pixelAt(page, b.x - 2, b.y + b.height / 2);
      const kitHalo = await alphaOver(page, expectedSource(s, mode, "destructive"), dark ? 40 : 20, surface);
      near(halo, kitHalo, `${key} the kit's invalid halo beside the field`);
      near(await pixelAt(page, b.x - 8, b.y + b.height / 2), surface, `${key} ground beyond the halo`);
      results.push(expectAtKit(`${key} invalid field boundary against the ground`, border, surface, kitBorder, surface, 3));
      results.push(expectAtKit(`${key} invalid field boundary against its halo`, border, halo, kitBorder, kitHalo, 3));
      results.push(expectAtKit(`${key} invalid field boundary against its fill`, border, fill, kitBorder, kitFill, 3));
    }
  }
  await testInfo.attach("field-invalid.json", { body: JSON.stringify(results, null, 2), contentType: "application/json" });
  expect(results.filter((r) => r.kitWaiver).map((r) => r.name).sort(), `the ${OWNER_WAIVED} rows (${mode})`).toEqual([...FIELD_INVALID_WAIVED[mode]].sort());
});

// G067 motion delta, retyped to the bare kit fields (G069; owner kit-fidelity direction, 2026-10-26): Input, Textarea
// and NativeSelect transition colour and box-shadow over the kit's 200ms (transition-[color,box-shadow] duration-200) in
// both motion preferences: the kit's components carry no reduced-motion variant (its own reduce rule, in the kit's brand stylesheet,
// stops only the brand ribbon motion) and the app's global 0.01ms rule is withdrawn, so reduced motion leaves the kit's
// motion as it is. At rest a valid field draws the kit's transparent border on its input fill and no focus indicator;
// after keyboard focus settles, the kit's focus state is drawn: the border in the ring colour plus a 3px ring at 30 % of
// it, never an outline (its contrast stays with S-FIELD-STATES and S-CONTRAST).
test("[S-FIELD-MOTION] fields: the focus state transitions over the kit's 200ms in both motion preferences, the settled state is the kit's ring", async ({ page }, testInfo) => {
  const mode = modeOf(testInfo);
  await load(page);
  const root = { tenant: DEFAULT_TENANT, label: ROOT };
  const own = content(page, ROOT);
  const fields = { input: own.getByTestId("input-rest"), textarea: own.getByTestId("field-textarea"), select: own.getByTestId("field-select") };
  const seconds = (v: string) => (v.trim().endsWith("ms") ? parseFloat(v) / 1000 : parseFloat(v));
  const ring = await expected(page, root, mode, "ring");
  const surface = await surfaceOf(page, ROOT);
  for (const motion of ["no-preference", "reduce"] as const) {
    await page.emulateMedia({ reducedMotion: motion });
    await frames(page);
    for (const [name, field] of Object.entries(fields)) {
      const what = `${name} (${motion})`;
      await resetFocus(page);
      await field.scrollIntoViewIfNeeded();
      await settledDeep(page, field);
      const rest = await styleOf(field, ["transition-property", "transition-duration", "outline-style", "box-shadow", "border-top-color"]);
      expect(rest["transition-property"].split(",").map((p) => p.trim()), `${what}: transitioned properties`).toEqual(expect.arrayContaining(["color", "box-shadow"]));
      const durations = rest["transition-duration"].split(",").map(seconds);
      expect(durations.every((d) => Math.abs(d - 0.2) < 1e-6), `${what}: the kit's own 200ms transition, never suppressed (${rest["transition-duration"]})`).toBe(true);
      expect(rest["outline-style"], `${what}: no outline`).toBe("none");
      expect(haloOf(rest["box-shadow"]), `${what}: no focus ring at rest (${rest["box-shadow"]})`).toBeNull();
      expect((await cssToRgba(page, rest["border-top-color"]))[3], `${what}: the kit's border-transparent at rest (${rest["border-top-color"]})`).toBe(0);
      await field.focus();
      await page.keyboard.press("Shift+Tab");
      await page.keyboard.press("Tab");
      await expect(field).toBeFocused();
      await settledDeep(page, field);
      const focus = await styleOf(field, ["outline-style", "box-shadow", "border-top-color"]);
      expect(focus["outline-style"], `${what}: settled focus draws no outline`).toBe("none");
      near(await colorOf(page, focus["border-top-color"]), ring, `${what}: settled focus border colour`, 1);
      const halo = haloOf(focus["box-shadow"]);
      expect(halo, `${what}: settled focus ring (${focus["box-shadow"]})`).not.toBeNull();
      expect(halo!.spread, `${what}: ring width`).toBe(3);
      const haloColor = await cssToRgba(page, halo!.color);
      // The ring is the ring token at 30 %, compared as painted over the surface: a translucent colour read back
      // through the canvas is premultiplied at 8 bits, so its unpremultiplied channels can be off by up to 2 at this
      // alpha, while the opaque composite is exact to within 1.
      near(await paintedOver(page, halo!.color, surface), await alphaOver(page, expectedSource(root, mode, "ring"), 30, surface), `${what}: ring colour (the ring token at 30 %, composited)`, 1);
      expect(Math.abs(haloColor[3] - 0.3 * 255), `${what}: ring alpha ${haloColor[3]}`).toBeLessThanOrEqual(2);
    }
  }
  await resetFocus(page);
});

// The S-FIELD-STATES rows the owner waived (G069 Spec §8; owner instruction 2026-10-26, "exact kit, overrides included"):
// every root palette that keeps the kit's platform input (platform, grey-primary and dark-sparse-destructive in both
// appearances; branded saves its own) paints the field boundary model (N3: input over the surface) at 1.26 Light / 1.48
// Dark and the bare kit radio's resting input/90 fill at 1.23 / 1.40 against the surface and about 1.09 / 1.3 against the
// muted panel. Desk-computed from the typed tables (sRGB composite at 8 bits). Every other row keeps its threshold.
const FIELD_STATES_WAIVED: string[] = ["platform", "grey-primary", "dark-sparse-destructive"].flatMap((t) => [
  ...["input-rest", "field-textarea", "field-select"].map((f) => `${t} ${f} resting boundary (N3)`),
  `${t} resting radio fill on the surface`,
  `${t} resting radio fill on muted`,
]);

// G069 S22 retyped by the owner's kit-fidelity direction (2026-10-26: the bare kit fields and radio, no app border-input
// or dark:aria-invalid:border-destructive), measured under every root palette in the project's appearance: the kit's
// transparent field border on the 50 % input fill, the boundary model (N3: the input token over the surface, Core's
// pair), the text on the fill (T12, >= 4.5:1), keyboard focus (the kit's ring state, >= 3:1 against the page and the
// field's fill), disabled at 50 %, and the radio's transparent border, resting 90 % fill, checked primary disc with its
// primary-foreground dot, undimmed disabled item, and its visible edge on the surface and on a muted panel. A pair the
// kit paints below its former threshold is asserted at the kit's ratio (OWNER-WAIVED, FIELD_STATES_WAIVED).
test("[S-FIELD-STATES] fields and radios: rest boundary, fill and text, focus, disabled and checked under every root palette", async ({ page }, testInfo) => {
  const mode = modeOf(testInfo);
  const results: (PairResult & { exempt?: string; kitWaiver?: string })[] = [];
  for (const tenant of TENANTS) {
    await load(page, tenant);
    const root = { tenant, label: ROOT };
    const own = content(page, ROOT);
    const surface = await surfaceOf(page, ROOT);
    // The input token as the kit paints it over the surface (the kit's Dark input is translucent), measured from the page's
    // own --input at the form and from the typed tables.
    const inputOver = async (ground: RGB) => paintedOver(page, (await styleOf(own.getByTestId("fixture-form"), ["--input"]))["--input"].trim(), ground);
    const kitInputOver = async (ground: RGB) => paintedOver(page, expectedSource(root, mode, "input"), ground);
    const background = await expected(page, root, mode, "background");
    for (const testid of ["input-rest", "field-textarea", "field-select"]) {
      const what = `${tenant} ${testid}`;
      const field = own.getByTestId(testid);
      await resetFocus(page);
      await page.mouse.move(0, 0);
      await field.scrollIntoViewIfNeeded();
      await settledDeep(page, field);
      expect((await cssToRgba(page, await computedColor(field, "border-top-color")))[3], `${what}: the kit's border-transparent`).toBe(0);
      const fill = await fillOf(page, field);
      near(fill, await alphaOver(page, expectedSource(root, mode, "input"), 50, surface), `${what}: fill (input at 50 % over the surface)`);
      results.push(expectAtKit(`${what} resting boundary (N3)`, await inputOver(surface), surface, await kitInputOver(background), background, 3));
      results.push(await expectText(page, `${what} text on its fill (T12)`, field, fill));
      await field.focus();
      await page.keyboard.press("Shift+Tab");
      await page.keyboard.press("Tab");
      await expect(field).toBeFocused();
      results.push(await expectFocusRing(page, what, field, { fill }));
    }
    await resetFocus(page);
    const disabled = own.getByTestId("input-disabled");
    await expect(disabled).toBeDisabled();
    expect((await styleOf(disabled, ["opacity"])).opacity, `${tenant} disabled field: 50 %`).toBe("0.5");
    results.push(exempt(`${tenant} disabled field (exempt: inactive component)`, await paintedOver(page, await computedColor(disabled), surface), await fillOf(page, disabled)));
    // The kit radio (Base UI): a 16px circle with role radio.
    const radio = (testid: string) => own.getByTestId(testid);
    for (const testid of ["radio-one", "radio-two", "radio-three"]) {
      const r = radio(testid);
      await expect(r).toHaveAttribute("role", "radio");
      const b = await box(r);
      expect([Math.round(b.width), Math.round(b.height)], `${tenant} ${testid}: 16px`).toEqual([16, 16]);
      expect((await cssToRgba(page, await computedColor(r, "border-top-color")))[3], `${tenant} ${testid}: the kit's border-transparent`).toBe(0);
    }
    await expect(radio("radio-one")).toHaveAttribute("aria-checked", "true");
    await expect(radio("radio-two")).toHaveAttribute("aria-checked", "false");
    // Resting: the 90 % input fill at the centre, the radio's only visible edge, against the surface beside it.
    const restBox = await box(radio("radio-two"));
    const restFill = await pixelAt(page, restBox.x + restBox.width / 2, restBox.y + restBox.height / 2);
    near(restFill, await alphaOver(page, expectedSource(root, mode, "input"), 90, surface), `${tenant} resting radio fill (input at 90 %)`);
    const kitRest = await alphaOver(page, expectedSource(root, mode, "input"), 90, background);
    results.push(expectAtKit(`${tenant} resting radio fill on the surface`, restFill, await pixelAt(page, restBox.x - 6, restBox.y + restBox.height / 2), kitRest, background, 3));
    // Checked: the primary disc and the primary-foreground dot at the centre. The disc is sampled on the centre row, 5px
    // right of the centre: the pixel just inside the 1px border carries that border's anti-aliasing (about 9 % of the
    // input colour, measured in P0.2), and the dot (8px, 10px in Dark) anti-aliases only inward, so this pixel is the
    // disc alone in both appearances.
    const checkedBox = await box(radio("radio-one"));
    const cx = checkedBox.x + checkedBox.width / 2;
    const cy = checkedBox.y + checkedBox.height / 2;
    const disc = await pixelAt(page, cx + 5, cy);
    const dot = await pixelAt(page, cx, cy);
    near(disc, await expected(page, root, mode, "primary"), `${tenant} checked disc`);
    near(dot, await expected(page, root, mode, "primary-foreground"), `${tenant} checked dot`);
    results.push(expectNonText(`${tenant} checked dot against its disc`, dot, disc));
    results.push(expectAtKit(`${tenant} checked radio disc on the surface`, disc, await pixelAt(page, checkedBox.x - 6, cy), await expected(page, root, mode, "primary"), background, 3));
    // Disabled item: the kit leaves it undimmed (disabled:opacity-50 matches :disabled, which the Base UI <span> root never
    // does; data-disabled carries the state; the app's data-disabled:opacity-50 is withdrawn).
    expect((await styleOf(radio("radio-three"), ["opacity"])).opacity, `${tenant} disabled radio: the kit's undimmed item`).toBe("1");
    // On a muted panel: the resting fill (input at 90 % over muted) against the panel.
    const muted = await box(radio("radio-muted"));
    const mutedRgb = await expected(page, root, mode, "muted");
    results.push(
      expectAtKit(
        `${tenant} resting radio fill on muted`,
        await pixelAt(page, muted.x + muted.width / 2, muted.y + muted.height / 2),
        await pixelAt(page, muted.x - 6, muted.y + muted.height / 2),
        await alphaOver(page, expectedSource(root, mode, "input"), 90, mutedRgb),
        mutedRgb,
        3,
      ),
    );
    // Keyboard focus: Tab into the group lands on the checked item.
    await radio("radio-one").focus();
    await page.keyboard.press("Shift+Tab");
    await page.keyboard.press("Tab");
    await expect(radio("radio-one")).toBeFocused();
    results.push(await expectFocusRing(page, `${tenant} radio`, radio("radio-one")));
    await resetFocus(page);
  }
  await testInfo.attach("field-states.json", { body: JSON.stringify(results, null, 2), contentType: "application/json" });
  expect(results.filter((r) => r.kitWaiver).map((r) => r.name).sort(), `the ${OWNER_WAIVED} rows (${mode})`).toEqual([...FIELD_STATES_WAIVED].sort());
});

// G069 S12/S16 and Plan A7: the kit controls inside a native <form method="post">: typed values, the native select, the
// kit RadioGroup's documented form bridge (one submitted value for the group name, required, a disabled item that cannot
// be chosen, arrow keys moving the selection and skipping it, role-based .check()), a disabled field that is not
// submitted, the kit Button submit, and the form reflowing at 320px.
test("[S-FIELD-FORM] kit form controls submit through a native form: values, radio group, disabled and required, keyboard, 320px", async ({ page }) => {
  await load(page);
  const form = page.getByTestId("fixture-form");
  const formData = () => form.evaluate((f) => [...new FormData(f as HTMLFormElement).entries()].map(([k, v]) => [k, String(v)]));
  const choice = async () => (await formData()).filter(([k]) => k === "choice").map(([, v]) => v);
  expect(await choice(), "the group's default value, submitted once").toEqual(["one"]);
  expect(
    await form.evaluate((f) => [...(f as HTMLFormElement).elements].some((e) => (e as HTMLInputElement).name === "choice" && (e as HTMLInputElement).required)),
    "the group's form value is required",
  ).toBe(true);
  const group = page.getByTestId("radio-group");
  await group.getByRole("radio", { name: "Two" }).check();
  expect(await choice()).toEqual(["two"]);
  await expect(group.getByRole("radio", { name: "Two" })).toBeFocused();
  // Arrow keys move the selection and skip the disabled item (one, two, three disabled).
  await page.keyboard.press("ArrowDown");
  expect(await choice(), "ArrowDown from two skips the disabled three").toEqual(["one"]);
  await page.keyboard.press("ArrowUp");
  expect(await choice(), "ArrowUp from one skips the disabled three").toEqual(["two"]);
  const three = group.getByRole("radio", { name: "Three" });
  await three.click({ force: true });
  expect(await choice(), "the disabled item cannot be chosen").toEqual(["two"]);
  await expect(three).toHaveAttribute("aria-checked", "false");
  await page.getByTestId("input-rest").fill("Fixture title");
  await page.getByTestId("field-textarea").fill("Fixture notes");
  await page.getByTestId("field-select").selectOption("two");
  // 320px: the form reflows without horizontal scrolling, every control inside the viewport.
  const size = page.viewportSize() ?? { width: 1280, height: 800 };
  await page.setViewportSize({ width: 320, height: 640 });
  await frames(page);
  const reflow = await page.evaluate(() => ({
    scroll: document.documentElement.scrollWidth,
    client: document.documentElement.clientWidth,
    overflow: [...document.querySelectorAll('[data-testid="fixture-form"] :is(input:not([type="hidden"]), textarea, select, button, [role="radio"])')]
      .map((e) => ({ e: e.getAttribute("data-testid") ?? e.tagName, r: e.getBoundingClientRect() }))
      .filter((x) => x.r.width > 0 && (x.r.left < 0 || x.r.right > 320))
      .map((x) => `${x.e} ${x.r.left.toFixed(1)}..${x.r.right.toFixed(1)}`),
  }));
  expect(reflow.scroll, "no horizontal scrolling at 320px").toBeLessThanOrEqual(reflow.client);
  expect(reflow.overflow, "form controls inside the 320px viewport").toEqual([]);
  await page.setViewportSize(size);
  const posted: string[] = [];
  await page.route("**/token-fixtures/submit", async (route) => {
    posted.push(route.request().postData() ?? "");
    // A self-contained answer page (an inline icon, so no further request is made).
    await route.fulfill({ status: 200, contentType: "text/html", body: '<!doctype html><html lang="en"><head><title>Submitted</title><link rel="icon" href="data:,"></head><body><main><h1>Submitted</h1></main></body></html>' });
  });
  await page.getByTestId("form-submit").click();
  await expect.poll(() => posted.length, { message: "the native form posted once" }).toBe(1);
  const sent = [...new URLSearchParams(posted[0]).entries()].sort(([a], [b]) => a.localeCompare(b));
  expect(sent, "exactly the enabled named controls, with their values").toEqual([
    ["choice", "two"],
    ["choice-select", "two"],
    ["notes", "Fixture notes"],
    ["title", "Fixture title"],
  ]);
});

// A control without an angled clip (every kit Button variant and size): neither the element nor a ::before layer is
// clipped, the label lies inside the element's box, and the kit focus state is whole: its border present on all four
// sides with the halo outside it, at least 3:1 against the page on each. The kit destructive variant's focus border
// (focus-visible:border-destructive/40; the app's solid N5 border is withdrawn by the owner's kit-fidelity direction,
// 2026-10-26) paints about 2.0 to 2.4 against the page on every fixture palette: OWNER-WAIVED, each side asserted at the
// ratio of the kit's computed border colour composited over that side's page sample.
async function expectUnclipped(page: Page, button: Locator, name: string, waived = false) {
  const geo = await button.evaluate((el) => {
    const b = el.getBoundingClientRect();
    const range = document.createRange();
    range.selectNodeContents(el);
    const t = range.getBoundingClientRect();
    return { clip: getComputedStyle(el).clipPath, before: getComputedStyle(el, "::before").clipPath, b: { l: b.left, t: b.top, r: b.right, btm: b.bottom }, t: { l: t.left, t: t.top, r: t.right, btm: t.bottom } };
  });
  expect([geo.clip, geo.before], `${name}: neither the element nor a ::before layer is clipped`).toEqual(["none", "none"]);
  expect(
    geo.t.l >= geo.b.l - 0.5 && geo.t.t >= geo.b.t - 0.5 && geo.t.r <= geo.b.r + 0.5 && geo.t.btm <= geo.b.btm + 0.5,
    `${name}: label ${JSON.stringify(geo.t)} inside the box ${JSON.stringify(geo.b)}`,
  ).toBe(true);
  await button.focus();
  await page.keyboard.press("Shift+Tab");
  await page.keyboard.press("Tab");
  await settledDeep(page, button);
  await button.evaluate((el) => el.scrollIntoView({ block: "center", inline: "center", behavior: "instant" }));
  await frames(page);
  const st = await styleOf(button, ["box-shadow", "border-top-color"]);
  const halo = haloOf(st["box-shadow"]);
  expect(halo, `${name}: focus ring (${st["box-shadow"]})`).not.toBeNull();
  const border = await cssToRgba(page, st["border-top-color"]);
  const samples = await compositeSamples(page, button, halo!.spread);
  expect(samples.map((s) => s.side).sort(), `${name}: focus state sampled on all sides`).toEqual(["bottom", "left", "right", "top"]);
  for (const s of samples) {
    expect(distance(s.border, over(border, s.page)), `${name}: focus border present on the ${s.side}`).toBeLessThanOrEqual(40);
    if (!waived) {
      expect(contrast(s.border, s.page), `${name}: focus border ${hex(s.border)} against the page ${hex(s.page)} on the ${s.side}`).toBeGreaterThanOrEqual(3);
      continue;
    }
    const kit = contrast(over(border, s.page), s.page);
    expect(kit, `${name}: the kit's focus border paints below 3:1 on the ${s.side} (${kit.toFixed(2)}), the ${OWNER_WAIVED} pair`).toBeLessThan(3);
    expect(Math.abs(contrast(s.border, s.page) - kit), `${name}: focus border ${hex(s.border)} against the page ${hex(s.page)} on the ${s.side}, ${OWNER_WAIVED}, at the kit ratio ${kit.toFixed(2)}`).toBeLessThanOrEqual(WAIVED_TOLERANCE);
  }
}

test("[S-ANGLE] fixtures: no kit Button clips its label or focus state; the platform CtaButton's angle keeps its label and inset outline", async ({ page }) => {
  for (const tenant of TENANTS) {
    await load(page, tenant);
    const own = content(page, ROOT);
    for (const variant of VARIANTS) {
      await expectUnclipped(page, own.getByTestId(`button-${variant}`), `${tenant} ${variant}`, variant === "destructive");
      await resetFocus(page);
    }
    for (const size of ["sm", "default", "lg", "icon"]) {
      await expectUnclipped(page, own.getByTestId(`button-size-${size}`), `${tenant} size ${size}`);
      await resetFocus(page);
    }
  }
  // The CtaButton (S24: platform palette only) is clipped to --vx-angle on the element itself (kit vx-angle).
  await load(page, PLATFORM);
  const cta = content(page, ROOT).getByTestId("cta-link");
  const geo = await cta.evaluate((el) => {
    const b = el.getBoundingClientRect();
    const range = document.createRange();
    range.selectNodeContents(el);
    const t = range.getBoundingClientRect();
    return { poly: getComputedStyle(el).clipPath, b: { x: b.left, y: b.top, w: b.width, h: b.height }, t: { x: t.left, y: t.top, w: t.width, h: t.height } };
  });
  expect(polygonPoints(geo.poly, 160, 32), "CtaButton clip-path is the source angle").toEqual(polygonPoints(SOURCE_ANGLE, 160, 32));
  const pts = polygonPoints(geo.poly, geo.b.w, geo.b.h).map(([x, y]) => [geo.b.x + x, geo.b.y + y]);
  const insidePoly = (x: number, y: number) => {
    let inside = false;
    for (let i = 0, j = pts.length - 1; i < pts.length; j = i++) {
      const [xi, yi] = pts[i], [xj, yj] = pts[j];
      if (yi > y !== yj > y && x < ((xj - xi) * (y - yi)) / (yj - yi) + xi) inside = !inside;
    }
    return inside;
  };
  for (const [x, y] of [[geo.t.x, geo.t.y], [geo.t.x + geo.t.w, geo.t.y], [geo.t.x, geo.t.y + geo.t.h], [geo.t.x + geo.t.w, geo.t.y + geo.t.h]]) {
    expect(insidePoly(x, y), `CtaButton: label corner (${x.toFixed(1)}, ${y.toFixed(1)}) clipped by the angle`).toBe(true);
  }
  await cta.focus();
  await page.keyboard.press("Shift+Tab");
  await page.keyboard.press("Tab");
  await settledDeep(page, cta);
  await expectFocusRing(page, "CtaButton", cta);
});

test("[S-SCREENSHOTS] fixtures: every root tenant, and the dialog open under the default tenant (visual evidence)", async ({ page }, testInfo) => {
  for (const tenant of TENANTS) {
    await load(page, tenant);
    await page.evaluate(() => (document.activeElement as HTMLElement | null)?.blur());
    await testInfo.attach(`fixtures-${tenant}.png`, { body: await page.screenshot({ fullPage: true, animations: "disabled" }), contentType: "image/png" });
  }
  await load(page);
  await content(page, ROOT).getByRole("button", { name: "Keyboard help" }).click();
  await expect(page.getByRole("dialog")).toBeVisible();
  await expect(page.locator("[data-slot=dialog-content]")).toHaveCSS("opacity", "1");
  await testInfo.attach("fixtures-dialog.png", { body: await page.screenshot({ animations: "disabled" }), contentType: "image/png" });
  await page.keyboard.press("Escape");
  await expect(page.getByRole("dialog")).toBeHidden();
});

// --- S-STATES (G069 Plan §6.2 row 3): the kit Button's alpha model ------------------------------------------

// The expected look of each variant on a surface (kit button.tsx, unaltered): the rendered fill at rest and on hover,
// the label and the 1px border (null: transparent). Pressed keeps the hover fill (no active fill) and moves 1px down;
// no state paints a whole-element filter.
type Look = { rest: RGB; label: RGB; border: RGBA | null; hover: RGB; hoverLabel: RGB; transparentAtRest: boolean };

async function lookOf(page: Page, s: Surface, mode: Mode, variant: Variant, surface: RGB): Promise<Look> {
  const src = (name: PaletteName) => expectedSource(s, mode, name);
  const c = async (name: PaletteName) => colorOf(page, src(name));
  const dark = darkVariants(s, mode);
  switch (variant) {
    case "default":
      // bg-primary text-primary-foreground hover:bg-primary/80
      return { rest: await c("primary"), label: await c("primary-foreground"), border: null, hover: await alphaOver(page, src("primary"), 80, surface), hoverLabel: await c("primary-foreground"), transparentAtRest: false };
    case "secondary":
      // bg-secondary text-secondary-foreground hover:bg-[color-mix(in_oklch,var(--secondary),var(--foreground)_5%)]
      return {
        rest: await c("secondary"),
        label: await c("secondary-foreground"),
        border: null,
        hover: rgb(await resolveColor(page, `color-mix(in oklch, ${src("secondary")}, ${src("foreground")} 5%)`)),
        hoverLabel: await c("secondary-foreground"),
        transparentAtRest: false,
      };
    case "destructive":
      // bg-destructive/10 hover:bg-destructive/20 text-destructive; Dark: bg-destructive/20, hover /30
      return { rest: await alphaOver(page, src("destructive"), dark ? 20 : 10, surface), label: await c("destructive"), border: null, hover: await alphaOver(page, src("destructive"), dark ? 30 : 20, surface), hoverLabel: await c("destructive"), transparentAtRest: false };
    case "outline":
      // border-border bg-background hover:bg-muted hover:text-foreground; Dark: bg-transparent, hover:bg-input/30
      return {
        rest: dark ? surface : await c("background"),
        label: await c("foreground"),
        border: await cssToRgba(page, src("border")),
        hover: dark ? await alphaOver(page, src("input"), 30, surface) : await c("muted"),
        hoverLabel: await c("foreground"),
        transparentAtRest: dark,
      };
    case "ghost":
      // hover:bg-muted hover:text-foreground; Dark: hover:bg-muted/50
      return { rest: surface, label: await c("foreground"), border: null, hover: dark ? await alphaOver(page, src("muted"), 50, surface) : await c("muted"), hoverLabel: await c("foreground"), transparentAtRest: true };
  }
}

// The pressed offset (active:not-aria-[haspopup]:translate-y-px) as [x, y] pixels.
function translateOf(value: string): number[] {
  return value.trim() === "none" ? [0, 0] : value.trim().split(/\s+/).map(parseFloat);
}

test("[S-STATES] fixtures: every variant's rest, hover, press, focus and disabled states and sizes under every root palette", async ({ page }, testInfo) => {
  const mode = modeOf(testInfo);
  const tok = async (s: Surface, name: PaletteName) => expected(page, s, mode, name);
  const col = async (value: string) => colorOf(page, value);
  for (const tenant of TENANTS) {
    await load(page, tenant);
    // Every surface's variants: the root and, on the default page, both previews, so the two projects measure the four
    // page/preview appearance combinations (Light preview A and Dark preview B on a Light and on a Dark page).
    for (const s of surfacesOf(tenant)) {
      const own = content(page, s.label);
      const surface = await surfaceOf(page, s.label);
      for (const variant of VARIANTS) {
        const what = s.label === ROOT ? `${tenant} ${variant}` : `${tenant} preview ${s.label} ${variant}`;
        const button = own.getByTestId(`button-${variant}`);
        await button.scrollIntoViewIfNeeded();
        await page.mouse.move(0, 0);
        await settledDeep(page, button);
        const rest = await styleOf(button, ["color", "background-color", "cursor", "border-radius", "clip-path", "filter", "border-top-width", "border-top-style", "border-top-color"]);
        // The kit Button sets no cursor and Tailwind v4's preflight none, so the browser's own button cursor shows (the app's
        // pointer-cursor base rule is withdrawn by the owner's kit-fidelity direction, 2026-10-26).
        expect(["default", "auto"], `${what}: the browser's button cursor, no app pointer rule (${rest.cursor})`).toContain(rest.cursor);
        // The kit's rounded-2xl (calc(var(--radius) * 1.8) = 18px).
        expect(rest["border-radius"], `${what}: radius`).toBe("18px");
        expect(rest["clip-path"], `${what}: the element itself is never clipped`).toBe("none");
        expect(rest.filter, `${what}: no filter at rest`).toBe("none");
        const look = await lookOf(page, s, mode, variant, surface);
        // A preview whose appearance differs from the page's (S8; the original case is the Light destructive
        // in Light preview A on a Dark page): the page appearance's tint must be distinguishable from the preview's
        // own, so the fill checks below prove which variant applied, by computed style and as rendered.
        if (variant === "destructive" && paletteMode(s, mode) !== mode) {
          const other = await alphaOver(page, expectedSource(s, mode, "destructive"), darkVariants(s, mode) ? 10 : 20, surface);
          expect(distance(other, look.rest), `${what}: the ${mode} page's resting tint ${hex(other)} differs from the preview's own ${hex(look.rest)}`).toBeGreaterThan(2);
        }
        expect((await styleOf(button, ["clip-path"], "::before"))["clip-path"], `${what}: no angled ::before layer`).toBe("none");
        // A 1px border on every side (border border-transparent bg-clip-padding); outline's is the kit border token.
        expect([rest["border-top-width"], rest["border-top-style"]], `${what}: 1px border`).toEqual(["1px", "solid"]);
        if (look.border) expect(distance4(await cssToRgba(page, rest["border-top-color"]), look.border), `${what}: border colour ${rest["border-top-color"]}`).toBeLessThanOrEqual(1);
        else expect((await cssToRgba(page, rest["border-top-color"]))[3], `${what}: transparent border`).toBe(0);
        if (look.transparentAtRest) expect((await cssToRgba(page, rest["background-color"]))[3], `${what}: transparent at rest`).toBe(0);
        else near(await paintedOver(page, rest["background-color"], surface), look.rest, `${what}: resting fill (computed, composited)`, 1);
        near(await fillOf(page, button), look.rest, `${what}: resting fill (rendered)`);
        near(await col(rest.color), look.label, `${what}: label`);
        await button.hover();
        await settledDeep(page, button);
        const hover = await styleOf(button, ["color", "background-color", "filter"]);
        expect(hover.filter, `${what}: hover paints no whole-element filter (kit alpha model)`).toBe("none");
        near(await paintedOver(page, hover["background-color"], surface), look.hover, `${what}: hover fill (computed, composited)`, 1);
        near(await fillOf(page, button), look.hover, `${what}: hover fill (rendered)`, 3);
        near(await col(hover.color), look.hoverLabel, `${what}: hover label`);
        if ((FILLED as readonly string[]).includes(variant)) {
          await page.mouse.down();
          await settledDeep(page, button);
          const pressed = await styleOf(button, ["translate", "filter", "color"]);
          expect(translateOf(pressed.translate), `${what}: pressed offset`).toEqual([0, 1]);
          expect(pressed.filter, `${what}: pressed paints no whole-element filter`).toBe("none");
          // Pressed while hovered: the hover fill stays (the kit has no active fill).
          near(await fillOf(page, button), look.hover, `${what}: pressed fill (rendered)`, 3);
          near(await col(pressed.color), look.label, `${what}: pressed label`);
          await page.mouse.up();
        }
        await page.mouse.move(0, 0);
        // Keyboard focus: the kit's focus state — no outline; the border takes the ring colour and a 3px ring at 30 % of it
        // is drawn outside (destructive: the kit's own border at destructive/40 over the surface, bg-clip-padding, and the
        // ring at destructive/20, Dark /40; the app's solid N5 border is withdrawn by the owner's kit-fidelity direction).
        await button.focus();
        await page.keyboard.press("Shift+Tab");
        await page.keyboard.press("Tab");
        await expect(button).toBeFocused();
        await settledDeep(page, button);
        const focus = await styleOf(button, ["outline-style", "border-top-color", "box-shadow"]);
        expect(focus["outline-style"], `${what}: focus draws no outline`).toBe("none");
        const destructive = variant === "destructive";
        const dark = darkVariants(s, mode);
        const borderWant = destructive ? await alphaOver(page, expectedSource(s, mode, "destructive"), 40, surface) : await tok(s, "ring");
        near(await paintedOver(page, focus["border-top-color"], surface), borderWant, `${what}: focus border`, 1);
        const halo = haloOf(focus["box-shadow"]);
        expect(halo, `${what}: focus ring (${focus["box-shadow"]})`).not.toBeNull();
        expect(halo!.spread, `${what}: focus ring width`).toBe(3);
        const haloWant = destructive ? await alphaOver(page, expectedSource(s, mode, "destructive"), dark ? 40 : 20, surface) : await alphaOver(page, expectedSource(s, mode, "ring"), 30, surface);
        near(await paintedOver(page, halo!.color, surface), haloWant, `${what}: focus ring colour (composited)`, 1);
        await resetFocus(page);
        // Disabled: inactive, half opacity, no pointer events, and focus does not move to it.
        const disabled = own.getByTestId(`button-${variant}-disabled`);
        expect(await disabled.evaluate((e) => e.matches(":disabled")), `${what} disabled: :disabled`).toBe(true);
        const d = await styleOf(disabled, ["opacity", "pointer-events"]);
        expect([d.opacity, d["pointer-events"]], `${what} disabled: opacity and pointer events`).toEqual(["0.5", "none"]);
        const activeBefore = await page.evaluate(() => document.activeElement?.outerHTML.slice(0, 120) ?? null);
        await disabled.evaluate((e) => (e as HTMLButtonElement).focus());
        expect(await page.evaluate(() => document.activeElement?.outerHTML.slice(0, 120) ?? null), `${what} disabled: focus does not move`).toBe(activeBefore);
      }
    }
    const s = { tenant, label: ROOT };
    const own = content(page, ROOT);
    // Sizes and the icon size: computed metrics (the merge of size and variant is verified). The kit Button has no block
    // variant (G069 S12), so the retired block row is gone.
    for (const [size, want] of Object.entries(BUTTON_SIZES)) {
      const sized = own.getByTestId(`button-size-${size}`);
      const st = await styleOf(sized, ["padding-top", "padding-right", "padding-bottom", "padding-left", "font-size", "line-height", "white-space"]);
      const got = [st["padding-top"], st["padding-right"], st["padding-bottom"], st["padding-left"], st["font-size"], st["line-height"]].map(parseFloat);
      const exp = [want.py, want.px, want.py, want.px, want.fontSize, want.lineHeight];
      got.forEach((v, i) => expect(Math.abs(v - exp[i]), `${tenant} size ${size} metric ${i}: ${v} vs ${exp[i]}`).toBeLessThanOrEqual(1));
      expect(st["white-space"], `${tenant} size ${size}: no wrapping`).toBe("nowrap");
      const height = (await sized.boundingBox())!.height;
      expect(Math.abs(height - want.height), `${tenant} size ${size} height: ${height} vs ${want.height}`).toBeLessThanOrEqual(0.5);
    }
    const icon = await own.getByTestId("button-size-icon").boundingBox();
    expect([icon!.width, icon!.height], `${tenant} icon size`).toEqual([BUTTON_ICON.width, BUTTON_ICON.height]);
    // ButtonLink renders the anchor itself as the button, with the variant's classes and data-slot="button".
    const link = own.getByTestId("button-as-link");
    expect(await link.evaluate((el) => [el.tagName, el.getAttribute("data-slot")]), `${tenant} ButtonLink`).toEqual(["A", "button"]);
    // Its variant (secondary) and the control geometry come from buttonVariants: the secondary fill and the 18px radius.
    near(await fillOf(page, link), await tok(s, "secondary"), `${tenant} ButtonLink: secondary fill`);
    expect((await styleOf(link, ["border-radius"]))["border-radius"], `${tenant} ButtonLink: radius`).toBe("18px");
    // S24: no CtaButton under a tenant palette.
    if (tenant !== PLATFORM) await expect(page.getByTestId("cta-link"), `${tenant}: no CtaButton under a tenant palette`).toHaveCount(0);
  }
});

// G069 S24/S15: the kit CtaButton (bg-blue text-white, hover:brightness-110 active:brightness-90) renders only under the
// platform palette, through ButtonLink, with the S15 inset outline. Its fixed white label is measured as painted on the
// kit's own --vx-blue #0075ff (the app's darker --vx-blue correction is withdrawn by the owner's kit-fidelity direction,
// 2026-10-26) at rest, hover (x1.10) and active (x0.90): 4.21 / 3.77 / 4.02 in both appearances (Core TestDefaultsPass),
// below 4.5:1, so each is OWNER-WAIVED and asserted at that kit ratio. Its focus outline inside the clip keeps 3:1.
test("[S-STATES] CtaButton on the platform palette: fixed white label on --vx-blue at rest, hover and active; inset focus outline", async ({ page }, testInfo) => {
  const mode = modeOf(testInfo);
  await load(page, PLATFORM);
  const cta = content(page, ROOT).getByTestId("cta-link");
  await expect(cta).toHaveCount(1);
  expect(await cta.evaluate((el) => [el.tagName, el.getAttribute("data-slot")]), "CtaButton through ButtonLink").toEqual(["A", "button"]);
  // The platform --vx-blue is the platform primary (theme.css --primary: var(--vx-blue), the kit's #0075ff).
  const blue = hexToRgb(SOURCES[mode].primary);
  const white: RGB = [255, 255, 255];
  // The label as painted (the brightness filter scales the whole element, label included) on the rendered fill, against
  // the kit's pair at the same factor (Core's scale model).
  const label = async (name: string, fill: RGB, f: number) => expectAtKit(name, await textColor(page, cta, fill), fill, scaled(white, f), scaled(blue, f), 4.5);
  const results: PairResult[] = [];
  await cta.scrollIntoViewIfNeeded();
  await page.mouse.move(0, 0);
  await settledDeep(page, cta);
  const rest = await styleOf(cta, ["background-color", "color", "filter", "outline-style"]);
  near(await colorOf(page, rest["background-color"]), blue, "CtaButton fill (computed): --vx-blue", 1);
  expect(rgb(await cssToRgba(page, rest.color)), "CtaButton label: fixed white").toEqual([255, 255, 255]);
  expect(rest.filter, "CtaButton at rest: no filter").toBe("none");
  const restFill = await fillOf(page, cta);
  near(restFill, blue, "CtaButton fill (rendered)");
  results.push(await label("CtaButton at rest", restFill, 1));
  await cta.hover();
  await settledDeep(page, cta);
  expect(brightness((await styleOf(cta, ["filter"])).filter), "CtaButton hover filter").toBeCloseTo(1.1, 6);
  const hoverFill = await fillOf(page, cta);
  near(hoverFill, scaled(blue, 1.1), "CtaButton hover fill (rendered, x1.10)", 3);
  results.push(await label("CtaButton hover", hoverFill, 1.1));
  await page.mouse.down();
  await settledDeep(page, cta);
  expect(brightness((await styleOf(cta, ["filter"])).filter), "CtaButton active filter").toBeCloseTo(0.9, 6);
  const activeFill = await fillOf(page, cta);
  near(activeFill, scaled(blue, 0.9), "CtaButton active fill (rendered, x0.90)", 3);
  results.push(await label("CtaButton active", activeFill, 0.9));
  await page.mouse.up();
  await page.mouse.move(0, 0);
  // S15: the kit's 2px focus outline moved inside the angled clip (offset -3px) and drawn in the fixed label colour.
  await cta.focus();
  await page.keyboard.press("Shift+Tab");
  await page.keyboard.press("Tab");
  await expect(cta).toBeFocused();
  await settledDeep(page, cta);
  const focus = await styleOf(cta, ["outline-style", "outline-width", "outline-offset", "outline-color"]);
  expect([focus["outline-style"], focus["outline-width"], focus["outline-offset"]], "CtaButton focus outline").toEqual(["solid", "2px", "-3px"]);
  expect(rgb(await cssToRgba(page, focus["outline-color"])), "CtaButton focus outline colour: white").toEqual([255, 255, 255]);
  results.push(await expectFocusRing(page, "CtaButton", cta));
  await testInfo.attach("cta.json", { body: JSON.stringify(results, null, 2), contentType: "application/json" });
  // The three label rows are the kit's waived pairs (white on #0075ff is below 4.5:1 at every factor).
  expect(results.filter((r) => (r as KitPairResult).kitWaiver).map((r) => r.name), `the ${OWNER_WAIVED} CtaButton rows`).toEqual(["CtaButton at rest", "CtaButton hover", "CtaButton active"]);
});

// --- S-FIRSTPAINT on the tenant fixture (G069 S5/S7) ---------------------------------------------------------

// The default tenant's first paint: next-themes writes the resolved class before the body is parsed and the server-
// rendered tenant style is in the HTML, so the first content frame already has the root tenant's ground and heading
// ink for the stored preference.
const FIXTURE_TARGET: ProbeTarget = { content: `[data-fixture-surface="${ROOT}"]`, heading: `[data-fixture-surface="${ROOT}"] [data-testid="sample-h1"]` };
function tenantPalette(tenant: string, m: Mode): PaintPalette {
  const s = { tenant, label: ROOT };
  const background = expectedSource(s, m, "background");
  const ink = expectedSource(s, m, "foreground");
  return { background, ink, text2: foreground80(ink, background), text3: expectedSource(s, m, "muted-foreground"), primary: expectedSource(s, m, "primary") };
}

for (const c of FIRST_PAINT_CASES) {
  test(`[S-FIRSTPAINT] fixture ${DEFAULT_TENANT}: stored ${c.stored}, OS ${c.os}: the root tenant's palette from the first content frame`, async ({ page }, testInfo) => {
    // The case names its OS appearance and emulates it, so it runs in both projects (the harness allows no skips).
    const palettes = { light: tenantPalette(DEFAULT_TENANT, "light"), dark: tenantPalette(DEFAULT_TENANT, "dark") };
    await page.emulateMedia({ colorScheme: c.os });
    if (c.stored === "blocked") await blockStorage(page);
    else await storeMode(page, c.stored);
    await installProbe(page, FIXTURE_TARGET);
    await page.goto(fixtureUrl(DEFAULT_TENANT));
    await page.waitForLoadState("load");
    const ground = await surfaceOf(page, ROOT);
    const hydratedAt = await waitForHydration(page);
    const evidence = await evaluateTimeline(page, c.stored, c.os, hydratedAt, palettes);
    near(ground, hexToRgb(palettes[evidence.want].background), "painted ground");
    await testInfo.attach("first-paint.json", { body: JSON.stringify({ case: c, hydratedAt, ground: hex(ground), ...evidence }, null, 2), contentType: "application/json" });
  });
}

// --- S-TEXT-SPACING ----------------------------------------------------------------------------------------

// The expected elements of one surface's own section (exact texts from the fixture page contract): the controls, the
// size row, the typography sampler and the font probes; the dialog trigger on the root only.
const PROBE_IDS = ["oswald-400", "oswald-500", "oswald-600", "oswald-700", "barlow-400", "barlow-600", "geist-400", "geist-500", "geist-700", "geist-mono-400"];
function surfaceItems(label: string): SpacingItem[] {
  const s = `[data-fixture-surface="${label}"] > section`;
  const items: SpacingItem[] = [
    { name: `${label} body`, selector: `${s} [data-testid="body-text"]`, texts: ["Body text in this scope."] },
    { name: `${label} muted`, selector: `${s} [data-testid="muted-text"]`, texts: ["Muted text in this scope."] },
    { name: `${label} link`, selector: `${s} [data-testid="link"]`, texts: ["A link in this scope"], kind: "control" },
  ];
  for (const variant of VARIANTS) {
    items.push({ name: `${label} ${variant}`, selector: `${s} [data-testid="button-${variant}"]`, texts: [`${variant} action`], kind: "control" });
    items.push({ name: `${label} ${variant} disabled`, selector: `${s} [data-testid="button-${variant}-disabled"]`, texts: [`${variant} action`], kind: "control", disabled: true });
  }
  if (label === ROOT) items.push({ name: `${label} Keyboard help`, selector: `${s} [data-testid="dialog-trigger"]`, texts: ["Keyboard help"], kind: "control" });
  for (const size of ["sm", "default", "lg"]) items.push({ name: `${label} size ${size}`, selector: `${s} [data-testid="button-size-${size}"]`, texts: [`${size} size`], kind: "control" });
  items.push({ name: `${label} size icon`, selector: `${s} [data-testid="button-size-icon"]`, texts: [""], kind: "icon" });
  items.push({ name: `${label} ButtonLink`, selector: `${s} [data-testid="button-as-link"]`, texts: ["Link as button"], kind: "control" });
  for (const [testid, text] of [
    ["sample-kicker", "Kicker label"],
    ["sample-h1", "Heading one"],
    ["sample-h2", "Heading two"],
    ["sample-h3", "Heading three"],
    ["sample-cta", "Call to action heading"],
    ["sample-lead", "Lead text in this scope."],
    ["sample-body", "Body text in this scope."],
    ["sample-meta", "Meta text in this scope."],
  ]) {
    items.push({ name: `${label} ${testid}`, selector: `${s} [data-testid="${testid}"]`, texts: [text] });
  }
  items.push({ name: `${label} sample text link`, selector: `${s} [data-testid="sample-text-link"]`, texts: ["A text link"], kind: "control" });
  items.push({ name: `${label} sample bullets`, selector: `${s} [data-testid="sample-bullets"] [data-slot="bullet-item"] > span:last-child`, texts: ["First bullet item", "Second bullet item"] });
  items.push({ name: `${label} font probes`, selector: `${s} [data-font-probe]`, texts: PROBE_IDS.map((id) => `Font probe ${id}`) });
  return items;
}

// Per surface: 3 text items, 5 variants enabled and disabled, 5 size-row controls, 11 sampler texts (8 type samples, the
// text link, 2 bullets) and 10 font probes; the root adds the dialog trigger.
const ITEMS_PER_SURFACE = 3 + 10 + 5 + 11 + 10;

// The fixture dialog's expected elements (exact texts from the fixture page contract).
const POPUP = '[data-slot="dialog-content"]';
const FIXTURE_MODAL_ITEMS: SpacingItem[] = [
  { name: "dialog title", selector: `${POPUP} [data-slot="dialog-title"]`, texts: ["Fixture dialog"] },
  { name: "dialog description", selector: `${POPUP} [data-slot="dialog-description"]`, texts: ["A kit dialog portalled into the document body."] },
  { name: "dialog action", selector: `${POPUP} [data-testid="dialog-action"]`, texts: ["Confirm"], kind: "control" },
  { name: "dialog done", selector: `${POPUP} [data-slot="dialog-footer"] [data-slot="dialog-close"]`, texts: ["Done"], kind: "control" },
  { name: "Close", selector: `${POPUP} > [data-slot="dialog-close"]`, texts: ["Close"], kind: "icon" },
];

test("[S-TEXT-SPACING] fixtures: touched control text keeps content and function under increased text spacing", async ({ page }, testInfo) => {
  await load(page);
  await page.addStyleTag({ content: SPACING_CSS });
  const states: SpacingState[] = [...(await expectNegativeCases(page))];
  const labels = [ROOT, ...PREVIEWS];
  const items = labels.flatMap((label) => surfaceItems(label));
  expect(items.reduce((n, i) => n + i.texts.length, 0), `${ITEMS_PER_SURFACE} expected elements per surface, plus the root's trigger`).toBe(ITEMS_PER_SURFACE * labels.length + 1);
  const closed = await inspectSpacing(page, "fixtures closed", items);
  states.push(closed);
  await testInfo.attach("spacing-fixtures.png", { body: await page.screenshot({ fullPage: true, animations: "disabled" }), contentType: "image/png" });
  await content(page, ROOT).getByRole("button", { name: "Keyboard help" }).click();
  await expect(page.locator(POPUP)).toBeVisible();
  // The negative cover is placed from the popup's geometry, so the open zoom has settled first (G067 motion delta).
  await dialogSettled(page);
  // Only the modal is active: every surface's own content is inert or hidden while it is open.
  for (const label of labels) {
    await expect.poll(() => content(page, label).evaluate((el) => !!el.closest("[inert], [aria-hidden='true']")), { message: `${label} content is inert while the dialog is open` }).toBe(true);
  }
  states.push(...(await expectNegativeCases(page, POPUP)));
  const modal = await inspectSpacing(page, "fixtures modal", FIXTURE_MODAL_ITEMS, POPUP);
  states.push(modal);
  await testInfo.attach("spacing-fixtures-dialog.png", { body: await page.screenshot({ animations: "disabled" }), contentType: "image/png" });
  await page.keyboard.press("Escape");
  await expect(page.locator(POPUP)).toBeHidden();
  await testInfo.attach("spacing-report.json", { body: JSON.stringify(states, null, 2), contentType: "application/json" });
  expectNoFindings(modal);
  expectNoFindings(closed);
});

// --- S-FONTS on the fixture probes (§7.4) -------------------------------------------------------------------

const PROBE_FACES = { oswald: "Oswald", barlow: "Barlow Condensed", geist: "Geist", "geist-mono": "Geist Mono" } as const;
const PROBE_FONTS: Record<string, { family: (typeof PROBE_FACES)[keyof typeof PROBE_FACES]; weight: string }> = Object.fromEntries(
  PROBE_IDS.map((id) => {
    const cut = id.lastIndexOf("-");
    const face = id.slice(0, cut) as keyof typeof PROBE_FACES;
    return [id, { family: PROBE_FACES[face], weight: id.slice(cut + 1) }];
  }),
);

test("[S-FONTS] fixtures: every probe renders in its bundled face and weight, without synthesis, on every surface", async ({ page }, testInfo) => {
  await page.emulateMedia({ reducedMotion: "reduce" });
  const ready = await openReady(page, fixtureUrl(DEFAULT_TENANT), FONT_PATHS);
  await page.mouse.move(0, 0);
  expectFontBytes(ready.records);
  const families = familyMap(expectFontMetadata(ready.bodies));
  const results: { surface: string; probe: string; verdict: string; weight: string; synthesized?: boolean }[] = [];
  for (const label of [ROOT, ...PREVIEWS]) {
    const own = content(page, label);
    const targets = Object.fromEntries(PROBE_IDS.map((id) => [id, own.locator(`[data-font-probe="${id}"]`)]));
    const fams = Object.fromEntries(PROBE_IDS.map((id) => [id, families[PROBE_FONTS[id].family]]));
    const rendered = await renderedFonts(page, targets, fams);
    for (const id of PROBE_IDS) {
      const el = targets[id];
      const weight = await el.evaluate((e) => getComputedStyle(e).fontWeight);
      // The synthesis twin check once per probe (on the root; the faces are not surface-dependent).
      const synth = label === ROOT ? await synthesized(el) : undefined;
      results.push({ surface: label, probe: id, verdict: rendered.find((r) => r.name === id)!.verdict, weight, synthesized: synth });
    }
  }
  await testInfo.attach("fonts-fixtures.json", { body: JSON.stringify({ records: ready.records, families, results }, null, 2), contentType: "application/json" });
  for (const r of results) {
    expect(r.verdict, `${r.surface} ${r.probe}: rendered in its bundled face`).toBe("custom");
    expect(r.weight, `${r.surface} ${r.probe}: weight`).toBe(PROBE_FONTS[r.probe].weight);
    if (r.synthesized !== undefined) expect(r.synthesized, `${r.probe}: no synthetic bold`).toBe(false);
  }
});

// --- S-BLUEPRINT on the fixture surfaces (§8.5 item 7) -------------------------------------------------------

test("[S-BLUEPRINT] fixtures: each surface's blueprint panel strokes are its muted-foreground over its ground", async ({ page }, testInfo) => {
  const mode = modeOf(testInfo);
  const fits: Record<string, StrokeFit> = {};
  for (const tenant of TENANTS) {
    await load(page, tenant);
    for (const label of tenant === DEFAULT_TENANT ? [ROOT, ...PREVIEWS] : [ROOT]) {
      const s = { tenant, label };
      const key = label === ROOT ? tenant : `preview ${label}`;
      const panel = content(page, label).getByTestId("blueprint-panel");
      const b = await box(panel);
      expect([Math.round(b.width), Math.round(b.height)], `${key} panel size`).toEqual([320, 160]);
      expect(await panel.locator("[data-blueprint]").getAttribute("aria-hidden")).toBe("true");
      const m = paletteMode(s, mode);
      fits[key] = fitStrokes(await panelImage(page, label), await expected(page, s, m, "background"), await expected(page, s, m, "muted-foreground"));
      await testInfo.attach(`blueprint-panel-${key.replace(/\s+/g, "-")}.png`, { body: await panel.screenshot({ animations: "disabled" }), contentType: "image/png" });
    }
  }
  await testInfo.attach("blueprint-surfaces.json", { body: JSON.stringify(fits, null, 2), contentType: "application/json" });
  for (const [key, fit] of Object.entries(fits)) expectFit(fit, `${key} blueprint panel`);
});
