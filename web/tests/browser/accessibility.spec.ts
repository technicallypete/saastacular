import AxeBuilder from "@axe-core/playwright";
import type { Locator, Page } from "@playwright/test";

import {
  box,
  crossCentre,
  dialogSettled,
  expectAxeClean,
  expectFocusRing,
  expectNonText,
  expectText,
  expectWaivedNonText,
  expectWaivedText,
  fillOf,
  KIT_WAIVERS,
  localTextContrast,
  settled,
  surfaceBeside,
  type LocalContrast,
  type PairResult,
} from "./support/contrast";
import { computedColor, cssToRgba, distance, hex, over, rgb } from "./support/color";
import { openReady } from "./support/fonts";
import { expect, test } from "./support/page-errors";
import { APPEARANCE_CONTROL, APPEARANCE_POPUP, parkPointer, resetFocus, waitForHydration } from "./support/modes";
import { pixelAt } from "./support/png";
import { expectNegativeCases, expectNoFindings, inspectSpacing, MODAL_ITEMS, POPUP, SPACING_CSS, type SpacingItem, type SpacingState } from "./support/spacing";

const TAGS = ["wcag2a", "wcag2aa", "wcag21a", "wcag21aa", "wcag22aa"];

// G069: the kit's white label on its #0075ff primary (Keyboard help, Account) is an OWNER-WAIVED axe color-contrast
// node, recorded with its ratio (support/contrast.ts expectAxeClean); every other violation fails.
async function audit(page: Page, label: string) {
  const results = await new AxeBuilder({ page }).withTags(TAGS).analyze();
  const waived = expectAxeClean(results.violations, `axe violations (${label})`);
  return { label, passes: results.passes.length, incomplete: results.incomplete.map((i) => i.id), waived };
}

test("[S-A11Y] shell: no automated WCAG 2.2 AA violations, closed and with the dialog open", async ({ page }, testInfo) => {
  const records = [];
  for (const width of [1280, 320]) {
    await page.setViewportSize({ width, height: width === 320 ? 640 : 800 });
    await page.goto("/");
    await waitForHydration(page);
    // At rest: a pointer left over Keyboard help would audit its 80 % hover fill (#3391ff, 3.16:1), not the waived rest pair.
    await parkPointer(page);
    await page.evaluate(() => Promise.allSettled(document.getAnimations().map((a) => a.finished)));
    records.push(await audit(page, `${width} closed`));
    await page.getByRole("button", { name: "Keyboard help" }).click();
    await expect(page.getByRole("dialog")).toBeVisible();
    // The dialog fades and zooms in (100ms, G067 motion delta); settle everything still running so the resting state
    // is audited.
    await page.evaluate(() => Promise.allSettled(document.getAnimations().map((a) => a.finished)));
    records.push(await audit(page, `${width} dialog open`));
    await page.keyboard.press("Escape");
  }
  await testInfo.attach("axe-summary.json", { body: JSON.stringify(records, null, 2), contentType: "application/json" });
});

// Selectors by slot (G048 §11.3); the configured identity and nav by role.
const KICKER = "#overview [data-slot=kicker]";
const LEAD = "#overview h1 + [data-slot=text]";
const INDEX = "[data-slot=ruled-grid] [data-slot=text]:first-child";
const CELL_BODY = "[data-slot=ruled-grid] h3 + [data-slot=text]";
const TRIGGER = 'header button[aria-haspopup="dialog"]';

// The kit field fill (bg-input/50: the input token at 50 % mixed in oklab with transparent) as the browser computes it
// beside EL, so the page's own palette applies.
async function kitInputFill(el: Locator): Promise<string> {
  return el.evaluate((e) => {
    const probe = document.createElement("span");
    e.parentElement!.append(probe);
    probe.style.backgroundColor = "color-mix(in oklab, var(--input) 50%, transparent)";
    const value = getComputedStyle(probe).backgroundColor;
    probe.remove();
    return value;
  });
}

// G069 (owner direction, exact kit with its overrides): the pairs the exact kit renders below their threshold are
// recorded OWNER-WAIVED at the kit's own ratio (support/contrast.ts KIT_WAIVERS), never as threshold passes; every other
// pair keeps its threshold. The app no longer shortens transitions under reduced motion (the kit's own rule covers only
// its brand motion), so each hover and focus state is read after its transitions settle.
test("[S-CONTRAST] shell: rendered text and non-text contrast at rest, hover and focus", async ({ page }, testInfo) => {
  await page.emulateMedia({ reducedMotion: "reduce" });
  await openReady(page);
  await page.mouse.move(0, 0);
  const dark = await page.evaluate(() => document.documentElement.classList.contains("dark"));
  const r: PairResult[] = [];

  // Text on its actual surfaces.
  const lead = page.locator(LEAD);
  r.push(await expectText(page, "lead (text-2) on page", lead, await surfaceBeside(page, lead)));
  const kicker = page.locator(KICKER);
  r.push(await expectText(page, "kicker (accent) on page", kicker, await surfaceBeside(page, kicker)));
  const h1 = page.locator("h1");
  r.push(await expectText(page, "heading (ink) on page", h1, await surfaceBeside(page, h1), 3));
  for (const [name, sel] of [
    ["H3 (accent)", "#getting-around h3"],
    ["index (muted-foreground)", INDEX],
    ["cell body (text-3)", CELL_BODY],
    ["About paragraph (text-3)", "#about p"],
    ["About item (text-2)", "#about li span:last-child"],
  ]) {
    const el = page.locator(sel).first();
    const b = await box(el);
    r.push(await expectText(page, `${name} on its surface`, el, await pixelAt(page, b.x - 4, b.y + 2)));
  }
  const footer = page.getByRole("contentinfo").locator("[data-slot=text]");
  const fb = await box(page.getByRole("contentinfo").locator("[data-slot=container]"));
  r.push(await expectText(page, "footer (muted-foreground) on page", footer, await pixelAt(page, fb.x + 4, fb.y + 2)));

  // Navigation links at rest and on hover.
  const navLinks = await page.getByRole("navigation", { name: "Primary" }).getByRole("link").all();
  expect(navLinks.length, "primary navigation links").toBe(3);
  for (const link of navLinks) {
    const bg = await surfaceBeside(page, link);
    r.push(await expectText(page, `nav "${await link.textContent()}" at rest`, link, bg));
    await link.hover();
    await settled(link);
    r.push(await expectText(page, `nav "${await link.textContent()}" on hover`, link, bg));
    await page.mouse.move(0, 0);
    await settled(link);
  }

  // Primary action (the default variant, a solid primary fill since G067): label on its rendered fill at rest,
  // hover and focus, and its focus ring. No fill-against-the-header boundary: the accepted Core palette delta r2
  // (clarification 3) withdrew fill-boundary comparisons for filled text buttons, identified by their labels (G017 N1).
  // G069 exact kit: the white label on the kit's #0075ff primary measures about 4.21:1 at rest and on focus in both
  // appearances and about 3.18:1 on the Light 80 % hover: OWNER-WAIVED, held at the kit's own ratio. The Dark hover
  // (the 80 % primary over the Dark header, about 5.91:1) keeps the 4.5:1 threshold.
  const trigger = page.getByRole("button", { name: "Keyboard help" });
  await expect(page.locator(TRIGGER)).toHaveCount(1);
  const tb = await box(trigger);
  const headerBg = await pixelAt(page, tb.x - 10, tb.y + 4);
  let fill = await fillOf(page, trigger);
  r.push(await expectWaivedText(page, "primary label at rest", trigger, fill, KIT_WAIVERS.primaryLabel));
  await trigger.hover();
  await settled(trigger);
  fill = await fillOf(page, trigger);
  r.push(dark ? await expectText(page, "primary label on hover", trigger, fill) : await expectWaivedText(page, "primary label on hover", trigger, fill, KIT_WAIVERS.primaryLabelLightHover));
  await page.mouse.move(0, 0);
  await trigger.focus();
  await page.keyboard.press("Shift+Tab");
  await page.keyboard.press("Tab");
  await settled(trigger);
  fill = await fillOf(page, trigger);
  r.push(await expectWaivedText(page, "primary label on focus", trigger, fill, KIT_WAIVERS.primaryLabel));
  r.push(await expectFocusRing(page, "primary button", trigger));

  // Appearance dropdown (amendment 03; the kit Select since the owner's G069 preview finding): the closed trigger shows
  // only the selected preference's symbol (System: Monitor), drawn in currentColor; the symbol against the header
  // surface and its trigger ground at rest and on hover (non-text 3:1); the combobox's name and value; its focus ring
  // by keyboard; then its labelled options, opened by keyboard, as text on the kit popup's ground.
  const appearance = page.locator(APPEARANCE_CONTROL);
  const symbol = appearance.locator("[data-appearance-icon=system]");
  await expect(symbol).toBeVisible();
  await expect(appearance.locator("[data-appearance-icon=light]")).toBeHidden();
  await expect(appearance.locator("[data-appearance-icon=dark]")).toBeHidden();
  const select = page.getByRole("combobox", { name: "Appearance", exact: true });
  await expect(select).toHaveAttribute("data-appearance-value", "system");
  // The combobox's value for assistive technology: the visually hidden label of the chosen preference (icon-only trigger).
  await expect(select, "the combobox's accessible value").toHaveText("System", { useInnerText: true });
  const ink = rgb(await cssToRgba(page, await computedColor(symbol)));
  const ab = await box(select);
  // G067: the trigger is a rounded field (rounded-2xl, a 1px border), so its ground is probed at the horizontal centre
  // 3px inside the top edge: inside the rounded shape and the border, above the symbols.
  const groundAt = () => pixelAt(page, ab.x + ab.width / 2, ab.y + 3);
  const restGround = await groundAt();
  r.push(expectNonText("appearance symbol against the header surface", ink, headerBg));
  r.push(expectNonText("appearance symbol on its trigger at rest", ink, restGround));
  // G069 exact kit (supersedes the app's solid input boundary, Plan §5.1/S22): the kit trigger draws no boundary (its
  // 1px border is transparent, border-transparent) and the resting field is identified by its own fill, the kit's
  // bg-input/50, as painted over the header. That fill against the header is the kit's own about 1.12:1 Light and
  // 1.16:1 Dark: OWNER-WAIVED below the 3:1 a boundary needed, held at the kit's ratio.
  const closed = await select.evaluate((e) => ({ border: getComputedStyle(e).borderTopColor, width: getComputedStyle(e).borderTopWidth, fill: getComputedStyle(e).backgroundColor }));
  expect((await cssToRgba(page, closed.border))[3], `appearance trigger: the kit's transparent border (${closed.border})`).toBe(0);
  expect(closed.width, "appearance trigger: the kit's 1px border").toBe("1px");
  const kitFill = await cssToRgba(page, await kitInputFill(select));
  const ownFill = await cssToRgba(page, closed.fill);
  expect(distance(rgb(ownFill), rgb(kitFill)) <= 1 && Math.abs(ownFill[3] - kitFill[3]) <= 1, `appearance trigger: the kit's bg-input/50 fill (${closed.fill})`).toBe(true);
  const paintedFill = over(kitFill, headerBg);
  expect(distance(restGround, paintedFill), `appearance trigger ground ${hex(restGround)}: the kit fill over the header, ${hex(paintedFill)}`).toBeLessThanOrEqual(2);
  r.push(expectWaivedNonText("appearance trigger fill against the header surface", restGround, headerBg, dark ? KIT_WAIVERS.fieldFillDark : KIT_WAIVERS.fieldFillLight));
  await select.hover();
  await settled(select);
  // The kit trigger has no hover fill: the ground under the pointer is the resting fill.
  const triggerHoverGround = await groundAt();
  expect(distance(triggerHoverGround, restGround), `appearance trigger on hover: no hover fill (${hex(triggerHoverGround)} vs ${hex(restGround)})`).toBeLessThanOrEqual(2);
  r.push(expectNonText("appearance symbol on its trigger on hover", ink, triggerHoverGround));
  await page.mouse.move(0, 0);
  await trigger.focus();
  await page.keyboard.press("Shift+Tab");
  await expect(select).toBeFocused();
  r.push(await expectFocusRing(page, "appearance trigger", select));
  await page.keyboard.press("Enter");
  const appearancePopup = page.locator(APPEARANCE_POPUP);
  await expect(appearancePopup).toBeVisible();
  await settled(appearancePopup);
  const options = appearancePopup.getByRole("option");
  expect(await options.allTextContents()).toEqual(["Light", "Dark", "System"]);
  // The popup's 16px corners are rounded: its ground is probed at the bottom centre, inside the group's padding.
  const apb = await box(appearancePopup);
  const appearancePopupBg = await pixelAt(page, apb.x + apb.width / 2, apb.y + apb.height - 2);
  for (const [i, option] of (await options.all()).entries()) r.push(await expectText(page, `appearance option ${i + 1}`, option, appearancePopupBg));
  await page.keyboard.press("Escape");
  await expect(appearancePopup).toBeHidden();
  await expect(select).toBeFocused();

  // Skip link: its text on its own ground while focused, and its focus outline (pointer parked away from it).
  await parkPointer(page);
  await resetFocus(page);
  await page.keyboard.press("Tab");
  const skip = page.getByRole("link", { name: "Skip to main content" });
  await expect(skip).toBeFocused();
  await settled(skip);
  const sb = await box(skip);
  r.push(await expectText(page, "skip link on its ground (focused)", skip, await pixelAt(page, sb.x + 3, sb.y + 3)));
  r.push(await expectFocusRing(page, "skip link", skip));
  await resetFocus(page);

  // Dialog: title, description, tips and the painted close cross on the popup (G069: the kit DialogContent's own
  // corner Close, a 28px ghost Button on bg-secondary named by its sr-only "Close").
  await page.evaluate(() => window.scrollTo(0, 0));
  await trigger.click();
  const dialog = page.getByRole("dialog");
  await expect(dialog).toBeVisible();
  await dialogSettled(page);
  const db = await box(dialog);
  // The popup's 24px corners are rounded: its surface is probed at the bottom centre, inside the bottom padding.
  const popupBg = await pixelAt(page, db.x + db.width / 2, db.y + db.height - 4);
  r.push(await expectText(page, "dialog title (ink)", dialog.locator("[data-slot=dialog-title]"), popupBg));
  r.push(await expectText(page, "dialog description", dialog.locator("[data-slot=dialog-description]"), popupBg));
  const tips = await dialog.locator("[data-slot=bullet-item] > span:last-child").all();
  expect(tips.length, "dialog tips").toBe(4);
  for (const [i, tip] of tips.entries()) r.push(await expectText(page, `dialog tip ${i + 1}`, tip, popupBg));
  const close = dialog.getByRole("button", { name: "Close" });
  r.push(expectNonText("close cross (painted centre) against the popup", await crossCentre(page, close), popupBg));
  await close.hover();
  await settled(close);
  const cb = await box(close);
  // The 28px Close (icon-sm, rounded) shows its hover fill at the horizontal centre 3px inside the top edge,
  // above the 16px cross; its corners are outside the rounded fill.
  const hoverGround = await pixelAt(page, cb.x + cb.width / 2, cb.y + 3);
  r.push(expectNonText("close cross (painted centre) on hover", await crossCentre(page, close), hoverGround));
  await page.mouse.move(0, 0);
  // Recorded without a threshold: the panel against the scrimmed page, and the panel's 1px ring (drawn just outside
  // its box since G067, ring-1).
  r.push(expectNonText("panel against the scrimmed page (recorded, no threshold)", popupBg, await pixelAt(page, 2, 2), 0));
  r.push(expectNonText("dialog ring against the panel (recorded, no threshold)", await pixelAt(page, db.x - 0.5, db.y + db.height / 2), popupBg, 0));

  await testInfo.attach("contrast.json", { body: JSON.stringify(r, null, 2), contentType: "application/json" });
  await resetFocus(page);
});

const LOCAL_VIEWPORTS = [
  { name: "1280x800", width: 1280, height: 800, textScale: 1 },
  { name: "320x640", width: 320, height: 640, textScale: 1 },
  { name: "1280x800 with 200% text", width: 1280, height: 800, textScale: 2 },
];

// Test-only negative control (§11.4): a 3px bar in the foreground colour behind the kicker, inside the
// overview's isolated container at z-index -1. Removed by the returned function.
async function kickerBar(page: Page): Promise<() => Promise<void>> {
  await page.evaluate((selector) => {
    const kicker = document.querySelector<HTMLElement>(selector)!;
    const container = kicker.parentElement!;
    const k = kicker.getBoundingClientRect();
    const c = container.getBoundingClientRect();
    const bar = document.createElement("div");
    bar.id = "g048-kicker-bar";
    bar.setAttribute("aria-hidden", "true");
    bar.style.cssText = `position:absolute;left:${k.left - c.left}px;top:${k.top - c.top + k.height / 2 - 1.5}px;width:${k.width}px;height:3px;background:var(--foreground);z-index:-1;pointer-events:none`;
    container.appendChild(bar);
  }, KICKER);
  return async () => {
    await page.evaluate(() => document.getElementById("g048-kicker-bar")?.remove());
  };
}

test("[S-CONTRAST] shell: local worst-case text contrast over the overview imagery, with a negative control", async ({ page }, testInfo) => {
  await page.emulateMedia({ reducedMotion: "reduce" });
  const results: LocalContrast[] = [];
  const targets: [string, string, number][] = [
    ["kicker", KICKER, 4.5],
    ["H1", "#overview h1", 3],
    ["lead", LEAD, 4.5],
  ];
  let control: LocalContrast | undefined;
  for (const vp of LOCAL_VIEWPORTS) {
    await page.setViewportSize({ width: vp.width, height: vp.height });
    await openReady(page);
    if (vp.textScale !== 1) await page.addStyleTag({ content: "html { font-size: 200% !important; }" });
    await page.mouse.move(0, 0);
    for (const [name, selector, min] of targets) {
      const local = await localTextContrast(page, `${vp.name} ${name}`, page.locator(selector), min);
      results.push(local);
      expect(local.ratio, `${vp.name} ${name}: ${local.fg} over ${local.worst}`).toBeGreaterThanOrEqual(min);
    }
    if (vp.name === "1280x800") {
      const remove = await kickerBar(page);
      try {
        control = await localTextContrast(page, "control: kicker over a 3px foreground bar", page.locator(KICKER), 4.5);
      } finally {
        await remove();
      }
      expect(control.ratio, `the comparator detects the injected bar (${control.fg} over ${control.worst})`).toBeLessThan(4.5);
    }
  }
  expect(control, "the negative control ran").toBeTruthy();
  await testInfo.attach("contrast-local.json", { body: JSON.stringify({ results, control }, null, 2), contentType: "application/json" });
});

const PRODUCT = process.env.EXPECTED_PRODUCT_NAME ?? "";
const DOMAIN = process.env.EXPECTED_PLATFORM_BASE_DOMAIN ?? "";

// The shell's expected elements with the dialog closed (exact texts from the page source; the
// configured identity from the same variables S-IDENTITY uses).
function shellItems(): SpacingItem[] {
  return [
    // The SiteHeader element is itself the <header>.
    { name: "product name", selector: "[data-slot=site-header] [data-slot=container] > p", texts: [PRODUCT] },
    { name: "nav links", selector: 'nav[aria-label="Primary"] a', texts: ["Overview", "Getting around", "About"], kind: "control" },
    // Amendment 03 (kit Select since G069): the symbol trigger draws no text, only the chosen preference's visually
    // hidden label (its value for assistive technology); the option labels live in its popup, in the document body only
    // while it is open.
    { name: "appearance control", selector: APPEARANCE_CONTROL, texts: ["System"], kind: "icon" },
    // G067: the default variant is a rounded solid fill, no angled layer (only cta is angled): its label stays in its box.
    { name: "Keyboard help", selector: TRIGGER, texts: ["Keyboard help"], kind: "control" },
    { name: "kicker", selector: KICKER, texts: ["Platform foundation"] },
    { name: "H1", selector: "h1", texts: [PRODUCT] },
    { name: "lead", selector: LEAD, texts: [`The generic application shell for ${PRODUCT}, served for ${DOMAIN}. Features appear here as they are built.`] },
    { name: "H2", selector: "h2", texts: ["Getting around", "About"] },
    { name: "indices", selector: INDEX, texts: ["01", "02", "03"] },
    { name: "H3", selector: "#getting-around h3", texts: ["Skip to content", "Keyboard help", "Appearance"] },
    {
      name: "cell bodies",
      selector: CELL_BODY,
      texts: [
        "The first Tab on the page reveals a link straight to the main content.",
        "The help dialog explains moving through the page and closing dialogs.",
        "Choose Light, Dark or System. System follows your device setting.",
      ],
    },
    { name: "About paragraph", selector: "#about p", texts: ["This shell is the shared frame for the platform's pages. It holds no product features yet."] },
    { name: "About items", selector: "#about li span:last-child", texts: ["Configured product identity", "Keyboard-operable structure", "Light, dark and system appearance"] },
    { name: "footer", selector: "footer > div", texts: [DOMAIN] },
  ];
}

test("[S-TEXT-SPACING] shell: touched text keeps content and function under increased text spacing", async ({ page }, testInfo) => {
  expect(PRODUCT, "EXPECTED_PRODUCT_NAME").not.toBe("");
  expect(DOMAIN, "EXPECTED_PLATFORM_BASE_DOMAIN").not.toBe("");
  const states: SpacingState[] = [];
  const real: SpacingState[] = [];
  for (const [w, h] of [[1280, 800], [320, 640]]) {
    await page.setViewportSize({ width: w, height: h });
    await page.emulateMedia({ reducedMotion: "reduce" });
    // The bundled fonts are loaded first, through the bounded phase-B readiness gate (B0).
    await openReady(page);
    await page.addStyleTag({ content: SPACING_CSS });
    states.push(...(await expectNegativeCases(page)));
    // Closed page.
    const closed = await inspectSpacing(page, `${w}x${h} closed`, shellItems());
    states.push(closed);
    real.push(closed);
    // The skip link while focused (keyboard focus; re-asserted as :focus when measured).
    await page.evaluate(() => window.scrollTo(0, 0));
    await resetFocus(page);
    await page.keyboard.press("Tab");
    await expect(page.getByRole("link", { name: "Skip to main content" })).toBeFocused();
    const skip = await inspectSpacing(page, `${w}x${h} skip link focused`, [{ name: "skip link", selector: 'a[href="#main"]', texts: ["Skip to main content"], kind: "control" }], undefined, 'a[href="#main"]');
    states.push(skip);
    real.push(skip);
    await resetFocus(page);
    await page.evaluate(() => window.scrollTo(0, 0));
    await testInfo.attach(`spacing-shell-${w}.png`, { body: await page.screenshot({ fullPage: true, animations: "disabled" }), contentType: "image/png" });
    // The open modal: only its own elements are active.
    await page.evaluate(() => window.scrollTo(0, 0));
    await page.getByRole("button", { name: "Keyboard help" }).click();
    await expect(page.locator(POPUP)).toBeVisible();
    await page.evaluate(() => Promise.allSettled(document.getAnimations().map((a) => a.finished)));
    for (const region of ["header", "main#main", "footer"]) {
      await expect
        .poll(() => page.locator(region).evaluate((el) => !!el.closest("[inert], [aria-hidden='true']")), { message: `${region} is inert while the dialog is open` })
        .toBe(true);
    }
    states.push(...(await expectNegativeCases(page, POPUP)));
    const modal = await inspectSpacing(page, `${w}x${h} modal`, MODAL_ITEMS, POPUP);
    states.push(modal);
    real.push(modal);
    await testInfo.attach(`spacing-shell-${w}-dialog.png`, { body: await page.screenshot({ animations: "disabled" }), contentType: "image/png" });
    await page.keyboard.press("Escape");
    await expect(page.locator("[data-slot=dialog-overlay]")).toHaveCount(0);
  }
  await testInfo.attach("spacing-report.json", { body: JSON.stringify(states, null, 2), contentType: "application/json" });
  for (const state of real) expectNoFindings(state);
});
