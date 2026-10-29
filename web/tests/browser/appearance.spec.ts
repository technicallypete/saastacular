import type { Page, Route } from "@playwright/test";

import { cssToRgba, distance, hex, rgb } from "./support/color";
import { box, fillOf } from "./support/contrast";
import { evaluateTimeline, expectInkDiscriminates, FIRST_PAINT_CASES, inkOf, installProbe, type ProbeTarget, type Stored } from "./support/first-paint";
import {
  APPEARANCE_CONTROL,
  APPEARANCE_LABELS,
  APPEARANCE_POPUP,
  appearanceControl,
  blockStorage,
  chooseAppearance,
  PALETTE,
  resolve,
  SOURCES,
  storedValue,
  storeMode,
  waitForHydration,
  type Resolved,
} from "./support/modes";
import { expect, test } from "./support/page-errors";
import { pixelAt } from "./support/png";

// G069 (S5/S6): next-themes is the one appearance manager. <html> carries the resolved appearance as its class (light or
// dark); the stored preference (light, dark or system, key unchanged) is the control's value and, after hydration, its
// symbol. The palette is the body's own background (kit base: bg-background).
async function rootState(page: Page) {
  return page.evaluate((controlSelector) => {
    const c = document.documentElement.classList;
    return {
      resolved: c.contains("dark") ? "dark" : c.contains("light") ? "light" : null,
      background: getComputedStyle(document.body).backgroundColor,
      scheme: getComputedStyle(document.documentElement).colorScheme,
      // The dropdown's selected value (mirrored on the kit Select's trigger) and the one symbol its closed trigger shows.
      checked: document.querySelector(`${controlSelector} [role=combobox]`)?.getAttribute("data-appearance-value") ?? null,
      icon: [...document.querySelectorAll("[data-appearance-icon]")].filter((el) => el.getClientRects().length > 0).map((el) => el.getAttribute("data-appearance-icon")).join(","),
    };
  }, APPEARANCE_CONTROL);
}

const control = appearanceControl;

// The application UI face (design.spec SANS): the bundled Geist first in the kit's sans stack.
const SANS = /^"?Geist"?, ui-sans-serif/;

// The highlighted option of the open list: the kit item marks it (data-highlighted) and the list moves focus to it.
async function highlightedIndex(page: Page): Promise<number> {
  return page.locator(APPEARANCE_POPUP).getByRole("option").evaluateAll((els) => els.findIndex((el) => el.hasAttribute("data-highlighted") || el === document.activeElement));
}

async function expectNearColor(page: Page, got: string, want: string, what: string) {
  const a = rgb(await cssToRgba(page, got));
  const b = rgb(await cssToRgba(page, want));
  expect(distance(a, b), `${what}: ${hex(a)} vs ${want}`).toBeLessThanOrEqual(1);
}

// After hydration: <html> carries the resolved appearance, the palette is that appearance's, and the closed trigger shows
// the stored preference's symbol (System: Monitor, whatever the OS resolves).
async function expectResolved(page: Page, want: Resolved, stored: string, what: string) {
  const s = await rootState(page);
  expect(s.resolved, `${what}: resolved <html> class`).toBe(want);
  expect(s.scheme, `${what}: color-scheme`).toBe(want);
  const bg = rgb(await cssToRgba(page, s.background));
  const target = rgb(await cssToRgba(page, PALETTE[want].background));
  expect(distance(bg, target), `${what}: palette ${hex(bg)} vs ${PALETTE[want].background}`).toBeLessThanOrEqual(1);
  expect(s.icon, `${what}: trigger symbol`).toBe(stored);
  return s;
}

// Emulates an OS appearance change that System must follow. next-themes applies it from the MediaQueryList change
// event, which the browser dispatches in its next rendering update, after the emulation call has returned, so the
// resolved class is read until it is the OS appearance (bounded by the expect timeout, as tokens.spec.ts's setScheme
// does); expectResolved then checks the settled state. An explicit choice ignores the event (S-APPEARANCE-OVERRIDE
// emulates directly and expects no change).
async function followOs(page: Page, os: Resolved) {
  await page.emulateMedia({ colorScheme: os });
  await expect.poll(async () => (await rootState(page)).resolved, { message: `OS ${os}: the resolved <html> class follows` }).toBe(os);
}

// Keyboard choice on the closed kit Select: Tab-focusable; an arrow key opens its list with an option highlighted (the
// stored mode), the arrow keys move the highlight one option per press, Enter chooses and closes, and focus returns to
// the trigger.
async function choose(page: Page, from: Stored, to: Stored) {
  const order: Stored[] = ["light", "dark", "system"];
  const trigger = control(page);
  const popup = page.locator(APPEARANCE_POPUP);
  await expect(trigger).toHaveAttribute("data-appearance-value", from);
  await trigger.focus();
  await expect(trigger).toBeFocused();
  await page.keyboard.press("ArrowDown");
  await expect(popup).toBeVisible();
  // The list settles on the stored mode: an earlier transient highlight (the first option, or focus on the list) is not
  // the position the arrow presses count from, so wait for the stored mode itself rather than for any highlight.
  let at = order.indexOf(from);
  await expect.poll(() => highlightedIndex(page), { message: `the open list highlights the stored mode (${from})` }).toBe(at);
  // Press only once the highlighted option also holds focus: before that the list's keyboard handling is not attached to it and a press is lost.
  await expect.poll(() => page.locator(APPEARANCE_POPUP).getByRole("option").evaluateAll((els) => els.some((el) => el.hasAttribute("data-highlighted") && el === document.activeElement)), { message: "the highlighted option holds focus" }).toBe(true);
  const want = order.indexOf(to);
  // At most two presses separate three options.
  for (let i = 0; i < 2 && at !== want; i++) {
    const next = want > at ? at + 1 : at - 1;
    await page.keyboard.press(want > at ? "ArrowDown" : "ArrowUp");
    await expect.poll(() => highlightedIndex(page), { message: `the highlight moves to ${order[next]}` }).toBe(next);
    at = next;
  }
  await page.keyboard.press("Enter");
  await expect(popup).toBeHidden();
  await expect(trigger).toHaveAttribute("data-appearance-value", to);
  // The icon-only trigger exposes the chosen preference as its text value (visually hidden label), not only an attribute.
  // Rendered text: Base UI's Select icon holds a never-painted "▼" text node inside its aria-hidden svg.
  await expect(trigger, "the combobox's accessible value").toHaveText(APPEARANCE_LABELS[to], { useInnerText: true });
  await expect(trigger.locator("[data-appearance-label]"), "the value label is exposed to assistive technology").not.toHaveAttribute("aria-hidden", "true");
  await expect(trigger).toBeFocused();
}

for (const os of ["light", "dark"] as const) {
  test(`[S-APPEARANCE-DEFAULT] no stored choice means System, following OS ${os}`, async ({ page }) => {
    await page.emulateMedia({ colorScheme: os });
    await storeMode(page, null);
    await page.goto("/");
    await waitForHydration(page);
    const s = await expectResolved(page, os, "system", "fresh session");
    expect(s.checked).toBe("system");
    expect(await storedValue(page)).toBeNull();
  });
}

test("[S-APPEARANCE-LIVE] System follows live OS changes without a reload", async ({ page }) => {
  await page.emulateMedia({ colorScheme: "dark" });
  await storeMode(page, null);
  await page.goto("/");
  await waitForHydration(page);
  for (const os of ["dark", "light", "dark", "light"] as const) {
    await followOs(page, os);
    await expectResolved(page, os, "system", `OS ${os}`);
  }
});

for (const [choice, os] of [["light", "dark"], ["dark", "light"]] as const) {
  test(`[S-APPEARANCE-OVERRIDE] explicit ${choice} overrides OS ${os} and survives OS changes`, async ({ page }) => {
    await page.emulateMedia({ colorScheme: os });
    await storeMode(page, null);
    await page.goto("/");
    await waitForHydration(page);
    await choose(page, "system", choice);
    await expectResolved(page, choice, choice, "after keyboard choice");
    expect(await storedValue(page)).toBe(choice);
    for (const next of [choice, os] as const) {
      await page.emulateMedia({ colorScheme: next });
      await expectResolved(page, choice, choice, `explicit ${choice} with OS ${next}`);
    }
  });
}

test("[S-APPEARANCE-RETURN] returning to System applies the current OS preference immediately", async ({ page }) => {
  await page.emulateMedia({ colorScheme: "light" });
  await storeMode(page, "dark");
  await page.goto("/");
  await waitForHydration(page);
  await expectResolved(page, "dark", "dark", "stored dark");
  await choose(page, "dark", "system");
  await expectResolved(page, "light", "system", "back to System");
  expect(await storedValue(page)).toBe("system");
  await followOs(page, "dark");
  await expectResolved(page, "dark", "system", "System then OS dark");
});

for (const mode of ["light", "dark", "system"] as const) {
  test(`[S-APPEARANCE-PERSIST] ${mode} persists across reloads as the mode, not the palette`, async ({ page }) => {
    await page.emulateMedia({ colorScheme: "dark" });
    await storeMode(page, null);
    await page.goto("/");
    await waitForHydration(page);
    // Choose a different mode first, so choosing the target is a real change (System included); the
    // options are chosen by their visible labels in the kit's popup, as a pointer or touch user picks them.
    await chooseAppearance(page, APPEARANCE_LABELS[mode === "light" ? "dark" : "light"]);
    await chooseAppearance(page, APPEARANCE_LABELS[mode]);
    expect(await storedValue(page)).toBe(mode);
    await page.reload();
    await waitForHydration(page);
    const s = await expectResolved(page, resolve(mode, "dark"), mode, "after reload");
    expect(s.checked).toBe(mode);
    expect(await storedValue(page)).toBe(mode);
  });
}

for (const value of ["purple", "", '{"mode":"dark"}', "DARK"]) {
  test(`[S-APPEARANCE-INVALID] stored ${JSON.stringify(value)} resolves to System`, async ({ page }) => {
    await page.emulateMedia({ colorScheme: "light" });
    await storeMode(page, value);
    await page.goto("/");
    await waitForHydration(page);
    const s = await expectResolved(page, "light", "system", `stored ${JSON.stringify(value)}`);
    expect(s.checked).toBe("system");
  });
}

for (const os of ["light", "dark"] as const) {
  test(`[S-APPEARANCE-NOSTORAGE] unavailable storage: System by default, in-session choice works (OS ${os})`, async ({ page }) => {
    await page.emulateMedia({ colorScheme: os });
    await blockStorage(page);
    await page.goto("/");
    await waitForHydration(page);
    await expectResolved(page, os, "system", "storage blocked");
    const other = os === "light" ? "dark" : "light";
    await chooseAppearance(page, APPEARANCE_LABELS[other]);
    const s = await expectResolved(page, other, other, "in-session choice");
    expect(s.checked).toBe(other);
    await page.reload();
    await waitForHydration(page);
    await expectResolved(page, os, "system", "after reload without storage");
  });
}

// Amendment 03 (owner M1), kit Select since the owner's G069 preview finding: the compact dropdown is one combobox named
// "Appearance" whose value is the selected preference, with labelled options in its popup; Tab reaches it, the arrow
// keys and Enter choose, and on a narrow touch viewport a tap anywhere on the symbol trigger opens the list and a
// tapped option applies.
test("[S-APPEARANCE-CONTROL] compact dropdown: name, value, labelled options, keyboard and touch", async ({ page, browser }, testInfo) => {
  await page.emulateMedia({ colorScheme: "dark" });
  await storeMode(page, null);
  await page.goto("/");
  await waitForHydration(page);
  const select = control(page);
  const popup = page.locator(APPEARANCE_POPUP);
  await expect(select).toHaveCount(1);
  await expect(select).toHaveAccessibleName("Appearance");
  await expect(select).toHaveAttribute("data-appearance-value", "system");
  await expect(page.locator(APPEARANCE_CONTROL).getByRole("radio")).toHaveCount(0);
  await expect(popup).toHaveCount(0);
  await select.click();
  await expect(popup).toBeVisible();
  const options = await popup.getByRole("option").evaluateAll((os) => os.map((o) => [(o.textContent ?? "").trim(), o.getAttribute("aria-selected") === "true"]));
  expect(options).toEqual([["Light", false], ["Dark", false], ["System", true]]);
  await page.keyboard.press("Escape");
  await expect(popup).toBeHidden();
  // Keyboard: Tab from the last primary navigation link lands on the trigger; the arrow keys and Enter choose.
  await page.getByRole("navigation", { name: "Primary" }).getByRole("link").last().focus();
  await page.keyboard.press("Tab");
  await expect(select).toBeFocused();
  await choose(page, "system", "dark");
  await expectResolved(page, "dark", "dark", "keyboard ArrowUp from System");
  expect(await storedValue(page)).toBe("dark");

  const context = await browser.newContext({ baseURL: testInfo.project.use.baseURL, hasTouch: true, viewport: { width: 320, height: 640 }, colorScheme: "light" });
  try {
    const touch = await context.newPage();
    await storeMode(touch, null);
    await touch.goto("/");
    await waitForHydration(touch);
    const trigger = touch.locator(APPEARANCE_CONTROL);
    const t = await box(trigger);
    const s = await box(control(touch));
    expect(t.width, "trigger target width").toBeGreaterThanOrEqual(24);
    expect(t.height, "trigger target height").toBeGreaterThanOrEqual(24);
    for (const k of ["x", "y", "width", "height"] as const) expect(Math.abs(s[k] - t[k]), `the combobox fills the trigger (${k})`).toBeLessThanOrEqual(2.5);
    expect(t.x + t.width, "trigger inside the 320px viewport").toBeLessThanOrEqual(320);
    await trigger.tap();
    const touchPopup = touch.locator(APPEARANCE_POPUP);
    await expect(touchPopup).toBeVisible();
    await touchPopup.getByRole("option", { name: "Dark", exact: true }).tap();
    await expect(touchPopup).toBeHidden();
    await expectResolved(touch, "dark", "dark", "touch choice");
    expect(await storedValue(touch)).toBe("dark");
  } finally {
    await context.close();
  }
});

// Owner preview finding (G069 r2): the open list is the kit's own Select popup, not the browser's option list. Its
// items sit inside the kit's padded group, inset from the popup's edges, set in the application face on the palette's
// popover ground and ink in both appearances; the keyboard opens, moves, chooses and closes it, with focus returned to
// the trigger; a keyboard choice persists like a pointer choice.
for (const os of ["light", "dark"] as const) {
  test(`[S-APPEARANCE-MENU] the kit Select popup: inset items, face, palette, keyboard and persistence (OS ${os})`, async ({ page }) => {
    await page.emulateMedia({ colorScheme: os });
    await storeMode(page, null);
    await page.goto("/");
    await waitForHydration(page);
    const trigger = control(page);
    const popup = page.locator(APPEARANCE_POPUP);
    await expect(popup).toHaveCount(0);
    await trigger.click();
    await expect(popup).toBeVisible();
    await expect(popup).toHaveCount(1);
    // The kit popup's open motion (fade and zoom-in-95, 100ms) scales its box; measure once it has settled.
    await popup.evaluate((e) => Promise.allSettled(e.getAnimations({ subtree: true }).map((a) => a.finished)));
    const items = popup.locator("[data-slot=select-item]");
    await expect(items).toHaveText(["Light", "Dark", "System"]);
    await expect(popup.getByRole("option")).toHaveCount(3);
    await expect(popup.getByRole("option", { name: "System", exact: true })).toHaveAttribute("aria-selected", "true");
    // The items are inset from every popup edge by the kit group's padding (p-1: 4px), so none touches the edge.
    const pb = await box(popup);
    const first = await box(items.first());
    const last = await box(items.last());
    expect(first.y - pb.y, "first item inset from the top edge").toBeGreaterThanOrEqual(4);
    expect(first.x - pb.x, "first item inset from the left edge").toBeGreaterThanOrEqual(4);
    expect(pb.x + pb.width - (last.x + last.width), "last item inset from the right edge").toBeGreaterThanOrEqual(4);
    expect(pb.y + pb.height - (last.y + last.height), "last item inset from the bottom edge").toBeGreaterThanOrEqual(4);
    // The application face and the popover palette: the popup's ground and ink, and an item at rest (the first, never
    // the highlighted selected one) in that ink.
    const surface = await popup.evaluate((el) => ({ background: getComputedStyle(el).backgroundColor, color: getComputedStyle(el).color }));
    const item = await items.first().evaluate((el) => ({ color: getComputedStyle(el).color, font: getComputedStyle(el).fontFamily }));
    expect(item.font, "item face").toMatch(SANS);
    await expectNearColor(page, surface.background, PALETTE[os].popover, `${os}: popup ground (popover)`);
    await expectNearColor(page, surface.color, SOURCES[os]["popover-foreground"], `${os}: popup ink (popover-foreground)`);
    await expectNearColor(page, item.color, SOURCES[os]["popover-foreground"], `${os}: item ink (popover-foreground)`);
    // Escape closes the list and returns focus to the trigger; Enter and Space reopen it.
    await page.keyboard.press("Escape");
    await expect(popup).toBeHidden();
    await expect(trigger).toBeFocused();
    for (const key of ["Enter", "Space"]) {
      await page.keyboard.press(key);
      await expect(popup, `${key} opens the list`).toBeVisible();
      await page.keyboard.press("Escape");
      await expect(popup).toBeHidden();
      await expect(trigger).toBeFocused();
    }
    // An arrow key opens the list, the arrow keys move, Enter chooses: the choice applies, persists and reloads as the
    // mode (System included on the way back).
    await choose(page, "system", "dark");
    await expectResolved(page, "dark", "dark", "keyboard choice");
    expect(await storedValue(page)).toBe("dark");
    await page.reload();
    await waitForHydration(page);
    const s = await expectResolved(page, "dark", "dark", "after reload");
    expect(s.checked).toBe("dark");
    await choose(page, "dark", "system");
    await expectResolved(page, os, "system", "keyboard return to System");
    expect(await storedValue(page)).toBe("system");
  });
}

// ---------------------------------------------------------------------------------------------
// First paint (Plan §11.4; probe and timeline in support/first-paint.ts). On "/" the theme-bearing content is the
// header and the heading is the page's H1, painted in the platform palette.
const SHELL_TARGET: ProbeTarget = { content: "header", heading: "h1" };
const SHELL_PALETTES = { light: PALETTE.light, dark: PALETTE.dark };

for (const c of FIRST_PAINT_CASES) {
  test(`[S-FIRSTPAINT] stored ${c.stored}, OS ${c.os}: correct palette from the first content frame`, async ({ page }, testInfo) => {
    await page.emulateMedia({ colorScheme: c.os });
    if (c.stored === "blocked") await blockStorage(page);
    else await storeMode(page, c.stored);
    await installProbe(page, SHELL_TARGET);
    await page.goto("/");
    await page.waitForLoadState("load");
    const header = await box(page.locator("header"));
    const ground = await pixelAt(page, header.x + 2, header.y + 2);
    const hydratedAt = await waitForHydration(page);
    const evidence = await evaluateTimeline(page, c.stored, c.os, hydratedAt, SHELL_PALETTES);
    const wantBg = rgb(await cssToRgba(page, PALETTE[evidence.want].background));
    expect(distance(ground, wantBg), `painted ground ${hex(ground)}`).toBeLessThanOrEqual(2);
    await testInfo.attach("first-paint.json", { body: JSON.stringify({ case: c, hydratedAt, ground: hex(ground), ...evidence }, null, 2), contentType: "application/json" });
  });
}

for (const [stored, os] of [["light", "dark"], ["dark", "light"]] as const) {
  test(`[S-FIRSTPAINT-DELAYED] framework JS held: stored ${stored} paints correctly against OS ${os}`, async ({ page }, testInfo) => {
    await page.emulateMedia({ colorScheme: os });
    await storeMode(page, stored);
    await installProbe(page, SHELL_TARGET);
    const held: Route[] = [];
    const heldUrls: string[] = [];
    let released = false;
    await page.route(/\/_next\/static\/.+\.js(\?.*)?$/, (route) => {
      if (released) return route.continue();
      held.push(route);
      heldUrls.push(route.request().url());
    });
    await page.goto("/", { waitUntil: "commit" });
    await expect(page.locator("header")).toBeVisible();
    await expect(page.locator("h1")).toBeVisible();
    await page.waitForTimeout(2000);
    // Hydration is impossible: framework chunks were requested and none has been released.
    expect(held.length, "framework chunks requested and held").toBeGreaterThan(0);
    expect(heldUrls.every((u) => u.endsWith(".js") || u.includes(".js?")), "only scripts are held").toBe(true);
    const want = rgb(await cssToRgba(page, PALETTE[stored].background));
    const osPalette = rgb(await cssToRgba(page, PALETTE[os].background));
    const wantInk = rgb(await cssToRgba(page, PALETTE[stored].ink));
    const osInk = rgb(await cssToRgba(page, PALETTE[os].ink));
    for (const m of [stored, os]) await expectInkDiscriminates(page, m, SHELL_PALETTES);
    const samples: { ground: string; action: string; h1Ink: string }[] = [];
    for (let i = 0; i < 3; i++) {
      const header = await box(page.locator("header"));
      const ground = await pixelAt(page, header.x + 2, header.y + 2);
      const action = await fillOf(page, page.getByRole("button", { name: "Keyboard help" }));
      const h1Ink = await inkOf(page, page.locator("h1"), ground);
      expect(distance(ground, want), `held: painted ground ${hex(ground)} is the stored ${stored} palette`).toBeLessThanOrEqual(2);
      expect(distance(ground, osPalette), "held: not the OS palette").toBeGreaterThan(20);
      expect(distance(action, rgb(await cssToRgba(page, PALETTE[stored].primary))), "held: primary fill").toBeLessThanOrEqual(2);
      expect(distance(h1Ink, wantInk), `held: painted H1 ink ${hex(h1Ink)} is the stored ${stored} ink`).toBeLessThanOrEqual(8);
      expect(distance(h1Ink, osInk), "held: H1 ink is not the OS palette's").toBeGreaterThan(20);
      // G069 S6 (Codex D-A1 option a): until hydration <html> carries only the resolved class, so the closed trigger shows
      // the resolved appearance's symbol (Sun or Moon); the stored preference's symbol follows after hydration.
      expect((await rootState(page)).icon, "held: the trigger shows the resolved appearance's symbol").toBe(resolve(stored, os));
      samples.push({ ground: hex(ground), action: hex(action), h1Ink: hex(h1Ink) });
      await page.waitForTimeout(200);
    }
    const releasedAt = await page.evaluate(() => performance.now());
    const fcpWhileHeld = await page.evaluate(() => window.__probe.paints.find((p) => p.name === "first-contentful-paint")?.startTime);
    expect(fcpWhileHeld, "FCP happened while framework JS was held").toBeLessThan(releasedAt);
    released = true;
    for (const r of held) await r.continue();
    const hydratedAt = await waitForHydration(page);
    expect(hydratedAt).toBeGreaterThan(releasedAt);
    const evidence = await evaluateTimeline(page, stored, os, hydratedAt, SHELL_PALETTES);
    const s = await expectResolved(page, stored, stored, "after hydration");
    expect(s.checked).toBe(stored);
    const expectedInk = { stored: PALETTE[stored].ink, os: PALETTE[os].ink };
    const expectedForeground = { stored: PALETTE[stored].foreground, os: PALETTE[os].foreground };
    await testInfo.attach("first-paint-delayed.json", { body: JSON.stringify({ stored, os, heldUrls, releasedAt, fcpWhileHeld, samples, hydratedAt, ...evidence, expectedInk, expectedForeground }, null, 2), contentType: "application/json" });
  });
}

for (const stored of ["light", "dark", "system"] as const) {
  for (const os of ["light", "dark"] as const) {
    test(`[S-HTML-ATTRS] stored ${stored}, OS ${os}: <html> carries exactly lang and the resolved appearance class`, async ({ page }) => {
      await page.emulateMedia({ colorScheme: os });
      await storeMode(page, stored);
      await page.goto("/");
      await waitForHydration(page);
      const attrs = await page.evaluate(() => ({
        html: [...document.documentElement.attributes].map((a) => [a.name, a.value]).sort(),
        suppressed: document.querySelectorAll("[suppresshydrationwarning]").length,
      }));
      // G069 S5: next-themes writes the class only (enableColorScheme off: no style attribute); React serializes no
      // hydration-warning attribute.
      expect(attrs.html).toEqual([["class", resolve(stored, os)], ["lang", "en"]]);
      expect(attrs.suppressed).toBe(0);
    });
  }
}

for (const stored of ["light", "dark", "system"] as const) {
  for (const os of ["light", "dark"] as const) {
    test(`[S-HYDRATION-DEV] stored ${stored}, OS ${os}: development hydration reports no mismatch @dev-hydration`, async ({ page }) => {
      const warnings: string[] = [];
      page.on("console", (m) => {
        if (m.type() === "warning" && /hydrat|did not match|server rendered/i.test(m.text())) warnings.push(m.text());
      });
      await page.emulateMedia({ colorScheme: os });
      await storeMode(page, stored);
      await page.goto("/");
      await waitForHydration(page);
      await expectResolved(page, resolve(stored, os), stored, "dev server");
      expect(warnings).toEqual([]);
    });
  }
}
