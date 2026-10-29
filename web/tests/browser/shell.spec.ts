import type { Page } from "@playwright/test";

import { box, dialogSettled, expectFocusRing, indicatorFor, localTextContrast } from "./support/contrast";
import { openReady } from "./support/fonts";
import { expect, test } from "./support/page-errors";
import { resetFocus, waitForHydration } from "./support/modes";
import { expectReflow, reflowLayout } from "./support/spacing";

const PRODUCT = process.env.EXPECTED_PRODUCT_NAME ?? "";
const DOMAIN = process.env.EXPECTED_PLATFORM_BASE_DOMAIN ?? "";
// The fixture build with a 64-character configured identity (playwright.config.ts, webServer 4); typed here
// independently of the configuration.
const LONG_IDENTITY_URL = "http://127.0.0.1:3004/";
const LONG_PRODUCT_NAME = "Configured Identity Unbrokenproductnamewithoutanyspacesforwrap Q";

async function open(page: Page) {
  await page.goto("/");
  await waitForHydration(page);
}

test("[S-IDENTITY] title and visible identity come from configuration @identity", async ({ page }) => {
  expect(PRODUCT, "EXPECTED_PRODUCT_NAME").not.toBe("");
  expect(DOMAIN, "EXPECTED_PLATFORM_BASE_DOMAIN").not.toBe("");
  await open(page);
  await expect(page).toHaveTitle(PRODUCT);
  await expect(page.getByRole("banner").getByText(PRODUCT, { exact: true })).toBeVisible();
  await expect(page.getByTestId("product-name")).toHaveText(PRODUCT);
  await expect(page.getByTestId("platform-domain")).toHaveText(DOMAIN);
  await expect(page.getByRole("contentinfo")).toContainText(DOMAIN);
});

test("[S-LANDMARKS] landmarks, headings and no account or tenant surface beyond the one Account link", async ({ page }) => {
  await open(page);
  await expect(page.getByRole("banner")).toHaveCount(1);
  await expect(page.getByRole("navigation", { name: "Primary" })).toHaveCount(1);
  // G018: the Workspace areas navigation appears only for a signed-in member on a tenant host, never on this frame.
  await expect(page.getByRole("navigation", { name: "Workspace areas" })).toHaveCount(0);
  await expect(page.getByRole("main")).toHaveCount(1);
  await expect(page.getByRole("contentinfo")).toHaveCount(1);
  // Amendment 03 (kit Select since G069): the appearance control is one compact combobox named "Appearance" (no radio
  // group).
  await expect(page.getByRole("combobox", { name: "Appearance", exact: true })).toHaveCount(1);
  await expect(page.getByRole("radio")).toHaveCount(0);
  const levels = await page.locator("h1, h2, h3, h4, h5, h6").evaluateAll((hs) => hs.map((h) => Number(h.tagName[1])));
  expect(levels.filter((l) => l === 1)).toHaveLength(1);
  expect(levels[0]).toBe(1);
  levels.forEach((level, i) => {
    if (i > 0) expect(level - levels[i - 1], `heading order ${levels.join(",")}`).toBeLessThanOrEqual(1);
  });
  // G004 (RC-1): the header's single Account link is the only account surface on this page; nothing signs in, signs
  // up, switches or names an organization or tenant here.
  const forbidden = /log ?in|sign ?in|sign ?up|account|organi[sz]ation|tenant|switch/i;
  for (const role of ["link", "button", "radio", "textbox", "combobox"] as const) {
    await expect(page.getByRole(role, { name: forbidden })).toHaveCount(role === "link" ? 1 : 0);
  }
  await expect(page.getByRole("banner").getByRole("link", { name: "Account", exact: true })).toHaveAttribute("href", "/account");
});

test("[S-NAV] each navigation link reaches an existing section by keyboard", async ({ page }) => {
  await open(page);
  const links = page.getByRole("navigation", { name: "Primary" }).getByRole("link");
  const count = await links.count();
  expect(count).toBe(3);
  for (let i = 0; i < count; i++) {
    const link = links.nth(i);
    const href = await link.getAttribute("href");
    expect(href).toMatch(/^#[a-z-]+$/);
    await link.focus();
    await page.keyboard.press("Enter");
    await expect(page).toHaveURL(new RegExp(`${href}$`));
    await expect(page.locator(href!)).toBeInViewport();
    // The target's heading is in the viewport and not covered (the header is static, X16).
    const heading = page.locator(`${href} h1, ${href} h2`).first();
    await expect(heading).toBeInViewport();
    const covered = await heading.evaluate((h) => {
      const r = h.getBoundingClientRect();
      const top = document.elementFromPoint(r.left + Math.min(r.width / 2, 40), r.top + r.height / 2);
      return top && (top === h || h.contains(top)) ? null : top ? `${top.tagName.toLowerCase()} ${top.className}`.slice(0, 120) : "nothing";
    });
    expect(covered, `${href}: heading covered`).toBeNull();
  }
});

async function expectSkipLink(page: Page) {
  await resetFocus(page);
  await page.keyboard.press("Tab");
  const skip = page.getByRole("link", { name: "Skip to main content" });
  await expect(skip).toBeFocused();
  const b = await box(skip);
  const vp = page.viewportSize()!;
  expect(b.width).toBeGreaterThan(0);
  expect(b.height).toBeGreaterThanOrEqual(24);
  expect(b.x).toBeGreaterThanOrEqual(0);
  expect(b.y).toBeGreaterThanOrEqual(0);
  expect(b.x + b.width).toBeLessThanOrEqual(vp.width + 0.5);
  await page.keyboard.press("Enter");
  await expect(page.locator("main#main")).toBeFocused();
  // The sequential focus starting point moved to main: Shift+Tab reaches the control just before it (since G004, the
  // header's Account link after Keyboard help).
  await page.keyboard.press("Shift+Tab");
  await expect(page.getByRole("link", { name: "Account", exact: true })).toBeFocused();
}

test("[S-SKIP] the skip link is first, visible on focus and moves focus to main", async ({ page }) => {
  await open(page);
  await expectSkipLink(page);
});

// G004 (RC-1) appends the header's Account link after Keyboard help.
const ORDER = ["Skip to main content", "Overview", "Getting around", "About", "appearance", "Keyboard help", "Account"];

async function focusedName(page: Page): Promise<string> {
  return page.evaluate(() => {
    const el = document.activeElement as HTMLElement | null;
    if (!el) return "";
    if (el.getAttribute("role") === "combobox" && el.getAttribute("aria-label") === "Appearance") return "appearance";
    return (el.textContent ?? "").trim();
  });
}

async function expectTabStops(page: Page) {
  await resetFocus(page);
  const seen: string[] = [];
  for (let i = 0; i < ORDER.length; i++) {
    await page.keyboard.press("Tab");
    seen.push(await focusedName(page));
    const indicator = indicatorFor(page);
    await expect(indicator).toHaveCount(1);
    const b = await box(indicator);
    const vp = page.viewportSize()!;
    expect(b.height, `${seen[i]} target height`).toBeGreaterThanOrEqual(24);
    expect(b.width, `${seen[i]} target width`).toBeGreaterThanOrEqual(24);
    expect(b.y + b.height, `${seen[i]} within the viewport`).toBeLessThanOrEqual(vp.height + 0.5);
    // Unobscured: the indicator's centre is not covered by anything else.
    const covering = await indicator.evaluate((el) => {
      const r = el.getBoundingClientRect();
      const top = document.elementFromPoint(r.left + r.width / 2, r.top + r.height / 2);
      if (top && (top === el || el.contains(top))) return null;
      return top ? `${top.tagName.toLowerCase()}[data-slot=${top.getAttribute("data-slot")}] ${top.className}`.slice(0, 160) : "nothing";
    });
    expect(covering, `${seen[i]} is obscured`).toBeNull();
    await expectFocusRing(page, seen[i], indicator);
  }
  expect(seen).toEqual(ORDER);
  const back: string[] = [];
  for (let i = 0; i < ORDER.length - 1; i++) {
    await page.keyboard.press("Shift+Tab");
    back.push(await focusedName(page));
  }
  expect(back).toEqual([...ORDER].reverse().slice(1));
}

test("[S-TAB] tab order, visible unobscured focus and target size", async ({ page }) => {
  await open(page);
  await expectTabStops(page);
});

// The Getting around grid's visible columns (G069 Plan S14: the kit RuledGrid with the page's
// repeat(auto-fit, minmax(min(100%, 16rem), 1fr)), 1px gaps, inside the kit Container's clamp(20px, 4vw, 40px) gutters
// and the grid's 1px border). Typed from that arithmetic, never read from the page: the grid holds as many 16rem tracks
// as its content width W fits, floor((W + 1) / (16rem + 1)), at least one, and its three cells fill at most three.
// 320: W 278, 1 column. 768: W 704.6, 2. 1280: W 1198, 4 tracks of which 3 hold cells (the 4th collapses). 640: W 586.8,
// 2. 1280 with 200% text: 16rem = 512px, 2. Each column is therefore at least 16rem wide or the whole grid width.
const VIEWPORTS = [
  { name: "320x640", width: 320, height: 640, textScale: 1, columns: 1 },
  { name: "768x1024", width: 768, height: 1024, textScale: 1, columns: 2 },
  { name: "1280x800", width: 1280, height: 800, textScale: 1, columns: 3 },
  { name: "640x400 (1280x800 at 200% zoom)", width: 640, height: 400, textScale: 1, columns: 2 },
  { name: "1280x800 with 200% text", width: 1280, height: 800, textScale: 2, columns: 2 },
];

// The dialog opens, fits the viewport, its close control is reachable, and Escape closes it.
async function expectDialogFits(page: Page, width: number, height: number) {
  await page.getByRole("button", { name: "Keyboard help" }).click();
  const dialog = page.getByRole("dialog");
  await expect(dialog).toBeVisible();
  // Measured at rest, after the open fade and zoom (G067 motion delta).
  await dialogSettled(page);
  const d = await box(dialog);
  expect(d.x).toBeGreaterThanOrEqual(0);
  expect(d.x + d.width).toBeLessThanOrEqual(width + 0.5);
  expect(d.y, "the dialog top is in the viewport").toBeGreaterThanOrEqual(-0.5);
  expect(d.y + d.height, "the dialog bottom is in the viewport").toBeLessThanOrEqual(height + 0.5);
  await expect(dialog.getByRole("button", { name: "Close" })).toBeInViewport();
  await page.keyboard.press("Escape");
  // The overlay unmounts on close, after its exit fade (G067 motion delta), before tabbing.
  await expect(page.locator("[data-slot=dialog-overlay]")).toHaveCount(0);
}

for (const vp of VIEWPORTS) {
  test(`[S-RESPONSIVE] ${vp.name}: reflow without horizontal scrolling or overlap`, async ({ page }) => {
    await page.setViewportSize({ width: vp.width, height: vp.height });
    await open(page);
    if (vp.textScale !== 1) await page.addStyleTag({ content: "html { font-size: 200% !important; }" });
    const layout = await reflowLayout(page);
    expectReflow(layout);
    expect(layout.columns, "grid columns").toBe(vp.columns);
    await expectDialogFits(page, vp.width, vp.height);
    if (vp.width === 320) {
      await expectSkipLink(page);
      await expectTabStops(page);
    }
  });
}

for (const vp of [
  { width: 320, height: 640 },
  { width: 1280, height: 800 },
]) {
  test(`[S-LONG-IDENTITY] ${vp.width}x${vp.height}: a 64-character configured identity wraps, reflows and stays operable`, async ({ page }, testInfo) => {
    expect(LONG_PRODUCT_NAME).toHaveLength(64);
    await page.setViewportSize(vp);
    // Wrapping and local contrast are measured with the bundled fonts loaded (B0).
    await openReady(page, LONG_IDENTITY_URL);
    await expect(page).toHaveTitle(LONG_PRODUCT_NAME);
    await expect(page.getByTestId("product-name")).toHaveText(LONG_PRODUCT_NAME);
    await expect(page.getByRole("banner").getByText(LONG_PRODUCT_NAME, { exact: true })).toBeVisible();
    const layout = await reflowLayout(page);
    expectReflow(layout);
    expect(layout.h1Lines, "the long identity wraps onto several lines").toBeGreaterThan(1);
    // Large text over the imagery behind it: the local worst case is at least 3:1.
    const h1 = await localTextContrast(page, "long-identity H1", page.getByTestId("product-name"), 3);
    expect(h1.ratio, `long-identity H1 local contrast: ${h1.fg} over ${h1.worst}`).toBeGreaterThanOrEqual(3);
    // The help dialog stays operable by keyboard and fits.
    const trigger = page.getByRole("button", { name: "Keyboard help" });
    await trigger.focus();
    await page.keyboard.press("Enter");
    await expect(page.getByRole("dialog")).toBeVisible();
    await page.keyboard.press("Escape");
    await expect(page.getByRole("dialog")).toBeHidden();
    await expect(trigger).toBeFocused();
    await expectDialogFits(page, vp.width, vp.height);
    await testInfo.attach("long-identity.json", { body: JSON.stringify({ viewport: vp, layout, h1 }, null, 2), contentType: "application/json" });
  });
}
