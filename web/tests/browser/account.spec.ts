import AxeBuilder from "@axe-core/playwright";
import type { Page } from "@playwright/test";

import { expectAxeClean } from "./support/contrast";
import { resetFocus } from "./support/modes";
import { expect, test } from "./support/page-errors";

// The account and sign-in surface on the ordinary stack (G004 Plan r5 §5, §7.1, §8.2), where no identity provider is
// configured: anonymous visitors are sent to sign in, the account API answers 401 whether or not a provider is
// configured, and starting a sign-in reports that sign-in is unavailable. The header's one Account link sits after
// Keyboard help. Runs in shell-light and shell-dark (and on the variants that run the shell projects).

const TAGS = ["wcag2a", "wcag2aa", "wcag21a", "wcag21aa", "wcag22aa"];

// G069: the kit's white label on its #0075ff primary is an OWNER-WAIVED axe color-contrast node, recorded with its ratio
// (support/contrast.ts expectAxeClean); every other violation fails.
async function audit(page: Page, label: string) {
  const results = await new AxeBuilder({ page }).withTags(TAGS).analyze();
  expectAxeClean(results.violations, `axe violations (${label})`);
}

async function focusedName(page: Page): Promise<string> {
  return page.evaluate(() => {
    const el = document.activeElement as HTMLElement | null;
    if (!el) return "";
    if (el.getAttribute("role") === "combobox" && el.getAttribute("aria-label") === "Appearance") return "appearance";
    return `${el.tagName.toLowerCase()}:${(el.textContent ?? "").trim()}`;
  });
}

test("[S-ACCOUNT-ANON] anonymous: /account goes to sign-in; the account API is 401 and never cached", async ({ page }) => {
  const response = await page.goto("/account");
  expect(new URL(page.url()).pathname).toBe("/sign-in");
  expect(new URL(page.url()).searchParams.get("return_to")).toBe("/account");
  expect(response?.headers()["cache-control"]).toMatch(/\bno-store\b/);
  await expect(page.getByRole("heading", { level: 1, name: "Sign in" })).toBeVisible();
  await expect(page.locator('input[type="hidden"][name="return_to"]')).toHaveValue("/account");
  await expect(page.getByTestId("status-message")).toHaveCount(0);
  const api = await page.request.get("/api/account");
  expect(api.status()).toBe(401);
  expect(api.headers()["cache-control"]).toBe("no-store");
  expect(await api.json()).toEqual({ error: "unauthenticated" });
  // Query values are never reflected into the page: an unknown status and a hostile return path render nothing and
  // are dropped. (Next's own router payload carries the current URL as data; it is not page content.)
  await page.goto("/sign-in?status=%3Cb%3Ehello%3C%2Fb%3E&return_to=https%3A%2F%2Fevil.example%2F");
  await expect(page.getByTestId("status-message")).toHaveCount(0);
  await expect(page.locator('input[name="return_to"]')).toHaveCount(0);
  await expect(page.locator("main b")).toHaveCount(0);
  const text = await page.locator("body").innerText();
  expect(text).not.toContain("evil.example");
  expect(text).not.toContain("hello");
  const attributes = await page.locator("main [href], main [action], main [value]").evaluateAll((els) =>
    els.map((e) => `${e.getAttribute("href") ?? ""}${e.getAttribute("action") ?? ""}${e.getAttribute("value") ?? ""}`),
  );
  expect(attributes.join(" ")).not.toContain("evil.example");
});

test("[S-ACCOUNT-UNAVAILABLE] without a configured provider, starting a sign-in reports that sign-in is unavailable", async ({ page, context }) => {
  await page.goto("/sign-in?return_to=/account");
  await page.getByRole("button", { name: "Sign in" }).click();
  await page.waitForURL((u) => u.searchParams.get("status") === "unavailable");
  expect(new URL(page.url()).pathname).toBe("/sign-in");
  await expect(page.getByTestId("status-message")).toHaveText("Sign-in is unavailable right now. Please try again later.");
  expect(await context.cookies(page.url()), "no session or login-binding cookie").toEqual([]);
});

test("[S-ACCOUNT-LINK] the header's Account link follows Keyboard help and leads to sign-in", async ({ page }) => {
  await page.goto("/");
  const link = page.getByRole("banner").getByRole("link", { name: "Account", exact: true });
  await expect(link).toHaveCount(1);
  await expect(link).toHaveAttribute("href", "/account");
  await expect(link).toHaveAttribute("data-slot", "button");
  const order = await page.locator("[data-slot=site-header] [data-slot=container] > :last-child").evaluate((region) =>
    [...region.children].map((c) => (c.matches("[data-appearance-control]") ? "appearance" : (c.textContent ?? "").trim())),
  );
  expect(order).toEqual(["appearance", "Keyboard help", "Account"]);
  await link.click();
  await page.waitForURL((u) => u.pathname === "/sign-in");
  expect(new URL(page.url()).searchParams.get("return_to")).toBe("/account");
});

test("[S-ACCOUNT-A11Y] sign-in states: no axe findings at 1280 and 320; the keyboard reaches the sign-in button after the header", async ({ page }) => {
  for (const width of [1280, 320]) {
    await page.setViewportSize({ width, height: width === 320 ? 640 : 800 });
    for (const path of ["/sign-in", "/sign-in?status=unavailable", "/sign-in?status=signed_out"]) {
      await page.goto(path);
      await audit(page, `${width} ${path}`);
      expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth), `${width} ${path}: no horizontal scroll`).toBe(true);
    }
  }
  await page.setViewportSize({ width: 1280, height: 800 });
  await page.goto("/sign-in");
  await resetFocus(page);
  const seen: string[] = [];
  for (let i = 0; i < 6; i++) {
    await page.keyboard.press("Tab");
    seen.push(await focusedName(page));
  }
  expect(seen).toEqual(["a:Skip to main content", "a:Sign in", "appearance", "button:Keyboard help", "a:Account", "button:Sign in"]);
  await page.keyboard.press("Enter");
  await page.waitForURL((u) => u.searchParams.get("status") === "unavailable");
});

// G063 (Plan r1 §5.2, §6; D-5): the rendered My Account composition with the MCP connection manager unavailable. The
// test-only fixture route (the fixture build on 127.0.0.1:3001, whose CORE_INTERNAL_URL is unreachable; typed here
// independently of the configuration) renders the real AccountSections and ConnectionManager with a scripted session:
// the manager's real list call fails, so its section shows the closed unavailable state while Session and Sign-out stay
// complete and usable. The production image has no such route.
const ACCOUNT_FIXTURE_URL = "http://127.0.0.1:3001/account-fixtures";

test("[S-ACCOUNT-MCP-UNAVAILABLE] My Account with MCP unavailable: Session and Sign-out intact, no MCP mutation; axe at 1280 and 320", async ({ page, request }) => {
  for (const width of [1280, 320]) {
    await page.setViewportSize({ width, height: width === 320 ? 640 : 800 });
    const response = await page.goto(ACCOUNT_FIXTURE_URL);
    expect(response?.status()).toBe(200);
    await expect(page.getByRole("heading", { level: 1, name: "Your account" })).toBeVisible();
    await expect(page.getByTestId("display-name")).toHaveText("Fixture Account Holder");
    await expect(page.getByTestId("mfa-status")).toHaveText("Not verified by this application");
    // The three sign-out forms, each with its hidden CSRF token.
    await expect(page.locator("#sign-out form")).toHaveCount(3);
    await expect(page.locator('#sign-out form input[type="hidden"][name="csrf_token"]')).toHaveCount(3);
    for (const name of ["Sign out", "Sign out everywhere", "Sign out and revoke WorkOS session"]) {
      await expect(page.getByRole("button", { name, exact: true })).toBeVisible();
    }
    await expect(page.getByRole("navigation", { name: "Primary" }).getByRole("link", { name: "AI connections" })).toHaveAttribute("href", "#mcp-connections");
    // The connections section: its closed unavailable state and nothing that could mutate a connection.
    await expect(page.locator("#mcp-connections").getByTestId("mcp-unavailable")).toBeVisible();
    await expect(page.locator('form[action^="/connections/mcp/"]')).toHaveCount(0);
    await expect(page.locator('input[name="connection_id"], input[name="display_name"], input[name="return_to"]')).toHaveCount(0);
    await expect(page.locator("#mcp-create")).toHaveCount(0);
    await expect(page.getByTestId("mcp-connection")).toHaveCount(0);
    // No account notice was asked for; the manager's closed state uses its own test id, not the page status message.
    await expect(page.getByTestId("status-message")).toHaveCount(0);
    const html = await page.content();
    expect(html).not.toMatch(/127\.0\.0\.1:9\b|\/internal\/v1\/|Authorization|Session fixture/);
    await audit(page, `${width} account fixture (MCP unavailable)`);
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth), `${width}: no horizontal scroll`).toBe(true);
  }
  // Keyboard: from the top of the page, Tab reaches the Sign out button through the header and the sections.
  await page.setViewportSize({ width: 1280, height: 800 });
  await page.goto(ACCOUNT_FIXTURE_URL);
  await resetFocus(page);
  const signOut = page.getByRole("button", { name: "Sign out", exact: true });
  for (let i = 0; i < 40 && !(await signOut.evaluate((b) => b === document.activeElement)); i++) await page.keyboard.press("Tab");
  await expect(signOut).toBeFocused();
  // The fixture route exists only in the fixture build, never in the production image.
  expect((await request.get("/account-fixtures")).status()).toBe(404);
});
