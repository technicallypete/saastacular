import { readFileSync, writeFileSync } from "node:fs";
import { dirname } from "node:path";

import AxeBuilder from "@axe-core/playwright";
import { type Browser, type BrowserContext, type Page, type TestInfo } from "@playwright/test";

import { expectAxeClean } from "./support/contrast";
import { expect, test } from "./support/page-errors";
import { observe } from "./support/sanitize";

// MCP connection setup scenarios (G014 Plan r8 §10.5; docs/foundation/MCP.md) in the disposable HTTPS identity
// environment (infra/compose/identity.yml), projects mcp-light and mcp-dark. The harness seeds the two registry
// Organizations before the run (infra/identity/mcp-seed.sql): dave is a member of both; erin has no account (G016). Pages that
// answer an error status are reached only through same-origin fetches from extra pages, so the default page's
// S-ERRORS check covers the setup page itself. Every page request must stay on the app origins and the emulator.

const ORIGIN_A = "https://org-a.localhost:543";
const ORIGIN_B = "https://org-b.localhost:543";
const EMULATOR = "http://emulator:4100";
const FORM = "application/x-www-form-urlencoded";
const SECRET = /^[A-Za-z0-9_-]{43}$/;
// G016: the old-host row infra/identity/mcp-seed.sql creates for dave in Organization A (removed host old.acme.test).
const OLD_HOST_CONNECTION = "https://old.acme.test:543/mcp/c/0f16c000-0000-4000-8000-0000000000c1";
const CONNECTION_URL = (origin: string) => new RegExp(`^${origin.replace(/[.]/g, "\\.")}/mcp/c/[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`);

type Credentials = { email: string; password: string };
type SeedUser = Credentials & { name: string };
const SEED = JSON.parse(readFileSync("/identity/browser/users.json", "utf8")) as { users: SeedUser[] };
function seeded(name: string): SeedUser {
  const found = SEED.users.find((u) => u.name === name);
  if (!found) throw new Error(`seed user ${name} missing`);
  return found;
}
const ERIN = seeded("erin");
const DAVE = seeded("dave");

// --- page traffic and extra contexts ------------------------------------------------------------------------------------
type Tracking = { origins: Set<string>; contexts: BrowserContext[] };
let tracking: Tracking;

function track(context: BrowserContext) {
  context.on("request", (request) => {
    const url = request.url();
    if (/^https?:/.test(url)) tracking.origins.add(new URL(url).origin);
  });
}

test.beforeEach(async ({ context }) => {
  tracking = { origins: new Set(), contexts: [] };
  track(context);
});

test.afterEach(async () => {
  for (const c of tracking.contexts) await c.close();
  const outside = [...tracking.origins].filter((o) => ![ORIGIN_A, ORIGIN_B, EMULATOR].includes(o));
  expect(outside, "page requests outside the app origins under test and the emulator").toEqual([]);
});

async function open(browser: Browser, testInfo: TestInfo): Promise<{ context: BrowserContext; page: Page }> {
  const context = await browser.newContext({ baseURL: ORIGIN_A, colorScheme: testInfo.project.use.colorScheme, viewport: testInfo.project.use.viewport });
  track(context);
  tracking.contexts.push(context);
  return { context, page: await context.newPage() };
}

// --- sign-in through the emulator's hosted page ------------------------------------------------------------------------
async function signIn(page: Page, who: Credentials, origin = ORIGIN_A) {
  await page.goto(`${origin}/sign-in?return_to=/account`);
  await page.getByRole("button", { name: "Sign in" }).click();
  await page.waitForURL((u) => u.origin === EMULATOR);
  await page.locator("#email").fill(who.email);
  await page.getByRole("button", { name: "Continue" }).click();
  const password = page.locator('input[name="password"]');
  await password.fill(who.password);
  await password.press("Enter");
  await page.waitForURL((u) => u.origin === origin && u.pathname === "/account");
}

// A sign-in that the membership rule refuses (G016 D-2(a)): back on ORIGIN's sign-in page with the not-member notice.
async function signInRefused(page: Page, who: Credentials, origin = ORIGIN_A) {
  await page.goto(`${origin}/sign-in?return_to=/account`);
  await page.getByRole("button", { name: "Sign in" }).click();
  await page.waitForURL((u) => u.origin === EMULATOR);
  await page.locator("#email").fill(who.email);
  await page.getByRole("button", { name: "Continue" }).click();
  const password = page.locator('input[name="password"]');
  await password.fill(who.password);
  await password.press("Enter");
  await page.waitForURL((u) => u.origin === origin && u.pathname === "/sign-in" && u.searchParams.get("status") === "not_member");
}

async function sessionValue(context: BrowserContext, origin = ORIGIN_A): Promise<string> {
  const cookie = (await context.cookies(origin)).find((c) => c.name === "__Host-sid");
  observe(cookie?.value);
  expect(cookie?.value).toMatch(SECRET);
  return cookie!.value;
}

// The CSRF token rendered into every form of the current page (all forms carry the same value).
async function csrfOf(page: Page): Promise<string> {
  const values = await page.locator('input[name="csrf_token"]').evaluateAll((els) => els.map((e) => (e as HTMLInputElement).value));
  expect(values.length).toBeGreaterThan(0);
  expect(new Set(values).size).toBe(1);
  expect(values[0]).toMatch(SECRET);
  observe(values[0]);
  return values[0];
}

// A same-origin form POST from an extra page of the context (the browser adds Origin and the cookies itself).
type Posted = { status: number; redirected: boolean; text: string };
async function browserPost(context: BrowserContext, path: string, body: string, origin = ORIGIN_A): Promise<Posted> {
  const p = await context.newPage();
  await p.goto(`${origin}/api/health`);
  const r = await p.evaluate(
    async ({ path, body, type }) => {
      const res = await fetch(path, { method: "POST", body, headers: { "content-type": type }, redirect: "manual", cache: "no-store" });
      const redirected = res.type === "opaqueredirect";
      return { status: res.status, redirected, text: redirected ? "" : await res.text() };
    },
    { path, body, type: FORM },
  );
  await p.close();
  return r;
}

// --- the setup page ----------------------------------------------------------------------------------------------------
async function setupPage(page: Page, origin = ORIGIN_A) {
  await page.goto(`${origin}/connections/mcp`);
  await expect(page.getByRole("heading", { level: 1 })).toBeVisible();
}

async function urls(page: Page): Promise<string[]> {
  return page.getByTestId("mcp-url").allTextContents();
}

// Creates one connection through the page's form and returns its URL (the one new entry of the list).
async function create(page: Page, origin = ORIGIN_A): Promise<string> {
  await setupPage(page, origin);
  const before = new Set(await urls(page));
  await page.getByRole("button", { name: "Create connection" }).click();
  await page.waitForURL((u) => u.pathname === "/connections/mcp" && u.searchParams.get("status") === "created");
  await expect(page.getByTestId("status-message")).toContainText("A new connection was created");
  const added = (await urls(page)).filter((u) => !before.has(u));
  expect(added).toHaveLength(1);
  expect(added[0]).toMatch(CONNECTION_URL(origin));
  return added[0];
}

function entry(page: Page, url: string) {
  return page.getByTestId("mcp-connection").filter({ has: page.getByTestId("mcp-url").getByText(url, { exact: true }) });
}

// --- G063: friendly names and the My Account embedding -----------------------------------------------------------------
// Code points are built with String.fromCodePoint so no escape sequence in this file can be rewritten by tooling.
const cp = (...points: number[]) => String.fromCodePoint(...points);
const EMOJI80 = cp(0x1f600).repeat(80);
// The shared padded vector (Plan r1 §6): Unicode White_Space around 80 non-BMP code points; stored as the 80 emoji.
const PADDED = cp(0x3000).repeat(8) + EMOJI80 + cp(0xa0, 0x2000, 0x20);

// Saves a name through the row's rename form and waits for the page's closed status.
async function rename(page: Page, url: string, name: string, status = "renamed") {
  const item = entry(page, url);
  await item.getByLabel("Name", { exact: true }).fill(name);
  await item.getByRole("button", { name: "Save name" }).click();
  await page.waitForURL((u) => u.searchParams.get("status") === status || u.searchParams.get("mcp_status") === status);
}

async function nameOf(page: Page, url: string): Promise<string | null> {
  return entry(page, url).getByTestId("mcp-name").textContent();
}

// Several same-origin form POSTs from one extra page of the context, in order; a redirect counts as 303.
async function bulkPost(context: BrowserContext, origin: string, path: string, bodies: string[]): Promise<number[]> {
  const p = await context.newPage();
  await p.goto(`${origin}/api/health`);
  const statuses = await p.evaluate(
    async ({ path, bodies, type }) => {
      const out: number[] = [];
      for (const body of bodies) {
        const res = await fetch(path, { method: "POST", body, headers: { "content-type": type }, redirect: "manual", cache: "no-store" });
        out.push(res.type === "opaqueredirect" ? 303 : res.status);
      }
      return out;
    },
    { path, bodies, type: FORM },
  );
  await p.close();
  return statuses;
}

// =======================================================================================================================

test("[M-NOAUTH] anonymous visitors get the sign-in-first page; the setup routes refuse without a session", async ({ page, context }) => {
  await setupPage(page);
  await expect(page.getByTestId("mcp-sign-in-first")).toBeVisible();
  await expect(page.getByRole("heading", { level: 1 })).toHaveText("Sign in first");
  await expect(page.getByRole("button", { name: "Create connection" })).toHaveCount(0);
  // G018: the Workspace areas navigation is for a signed-in member only.
  await expect(page.getByRole("navigation", { name: "Workspace areas" })).toHaveCount(0);
  const created = await browserPost(context, "/connections/mcp/create", "csrf_token=x");
  expect(created.status).toBe(401);
  expect(created.text).toContain("Not signed in");
  const revoked = await browserPost(context, "/connections/mcp/revoke", "csrf_token=x&connection_id=7f000000-0000-4000-8000-000000000001");
  expect(revoked.status).toBe(401);
});

// G016 D-2(a): a tenant-host session is issued only to a current member, so a non-member never reaches the setup page
// signed in. erin has no account and no membership: her sign-in on either host ends on the not-member notice with no
// session, and the setup page and routes answer as for an anonymous visitor. The page's not-available branch (a session
// whose membership ends later) is unchanged and no longer exercised in the browser: IT-MCP-11 proves the transport and
// tool denial after membership loss only, not that page's HTML or axe state.
test("[M-NONMEMBER] a non-member's sign-in is refused on both hosts; no session reaches the setup routes", async ({ page, context }) => {
  for (const origin of [ORIGIN_A, ORIGIN_B]) {
    await signInRefused(page, ERIN, origin);
    expect((await context.cookies(origin)).find((c) => c.name === "__Host-sid")).toBeUndefined();
    await setupPage(page, origin);
    await expect(page.getByTestId("mcp-sign-in-first")).toBeVisible();
    await expect(page.getByRole("button", { name: "Create connection" })).toHaveCount(0);
    const created = await browserPost(context, "/connections/mcp/create", "csrf_token=x", origin);
    expect(created.status).toBe(401);
  }
});

test("[M-SETUP] a member creates a connection on org-a and sees its URL and pending status", async ({ page }) => {
  await signIn(page, DAVE);
  await setupPage(page);
  await expect(page.getByRole("heading", { level: 1 })).toHaveText("Your AI client connections");
  // G018: a member's tenant page carries the Workspace areas navigation (Groups, Business Profiles; G019 adds Members);
  // none is the current area on this page.
  const areas = page.getByRole("navigation", { name: "Workspace areas" }).getByRole("link");
  await expect(areas).toHaveText(["Groups", "Business Profiles", "Members"]);
  expect(await areas.evaluateAll((as) => as.map((a) => [a.getAttribute("href"), a.getAttribute("aria-current")]))).toEqual([
    ["/groups", null],
    ["/business-profiles", null],
    ["/organization/members", null],
  ]);
  const url = await create(page);
  const item = entry(page, url);
  await expect(item).toHaveCount(1);
  await expect(item).toHaveAttribute("data-state", "pending");
  await expect(item.getByTestId("mcp-state")).toHaveText("Waiting for your AI client");
  await expect(item.getByTestId("mcp-client")).toHaveText("Not yet bound");
  // The connection URL is shown to its owner; it is a selector, not a credential, and nothing else is rendered.
  expect(await page.content()).not.toMatch(/eyJ[A-Za-z0-9_-]{10,}/);
  // G063: an unnamed (legacy-shape) create shows the neutral fallback; a named create shows its name, with the bound
  // client still shown separately.
  await expect(item.getByTestId("mcp-name")).toHaveText("Unnamed connection");
  await setupPage(page);
  const before = new Set(await urls(page));
  await page.getByLabel("Name (optional)").fill("  Work laptop  ");
  await page.getByRole("button", { name: "Create connection" }).click();
  await page.waitForURL((u) => u.pathname === "/connections/mcp" && u.searchParams.get("status") === "created");
  const named = (await urls(page)).filter((u) => !before.has(u));
  expect(named).toHaveLength(1);
  expect(await nameOf(page, named[0])).toBe("Work laptop");
  await expect(entry(page, named[0]).getByTestId("mcp-client")).toHaveText("Not yet bound");
  expect(page.url()).not.toContain("laptop");
});

test("[M-RENAME] rename, clear, duplicates, padded non-BMP, markup as text; invalid names refused without echo; revoked target not found", async ({ page, context }) => {
  await signIn(page, DAVE);
  const first = await create(page);
  const second = await create(page);
  await rename(page, first, "Desk client");
  await expect(page.getByTestId("status-message")).toHaveText("The connection's name was saved.");
  expect(await nameOf(page, first)).toBe("Desk client");
  await expect(entry(page, first).getByLabel("Name", { exact: true })).toHaveValue("Desk client");
  // Duplicate names are allowed.
  await rename(page, second, "Desk client");
  expect(await nameOf(page, second)).toBe("Desk client");
  expect(await nameOf(page, first)).toBe("Desk client");
  // The padded 80-non-BMP vector is accepted and shown as exactly 80 characters.
  await rename(page, first, PADDED);
  const stored = await nameOf(page, first);
  expect(stored).toBe(EMOJI80);
  expect(Array.from(stored ?? "").length).toBe(80);
  // Markup is a name like any other: rendered as text, never interpreted.
  const hostile = "<img src=x onerror=alert(1)><b>bold</b>";
  await rename(page, second, hostile);
  expect(await nameOf(page, second)).toBe(hostile);
  await expect(entry(page, second).locator("img, b, script")).toHaveCount(0);
  // Clearing returns to the neutral fallback.
  await rename(page, second, "   ");
  await expect(entry(page, second).getByTestId("mcp-name")).toHaveText("Unnamed connection");
  // 81 code points: the closed invalid_name notice; nothing changes and the attempted value is echoed nowhere.
  const tooLong = "Q".repeat(81);
  await rename(page, first, tooLong, "invalid_name");
  await expect(page.getByTestId("status-message")).toContainText("That name was not saved");
  expect(page.url()).not.toContain("QQQ");
  expect(await page.content()).not.toContain("QQQ");
  expect(await nameOf(page, first)).toBe(EMOJI80);
  // A control character (here a newline, which a text input cannot hold) is refused by the route with no core call.
  const csrf = await csrfOf(page);
  const id = first.slice(-36);
  const refused = await browserPost(context, "/connections/mcp/rename", `csrf_token=${csrf}&connection_id=${id}&display_name=Lap%0Atop`);
  expect(refused.redirected).toBe(true);
  await setupPage(page);
  expect(await nameOf(page, first)).toBe(EMOJI80);
  // Malformed bytes and a repeated name field.
  expect((await browserPost(context, "/connections/mcp/rename", `csrf_token=${csrf}&connection_id=${id}&display_name=%FF%FE`)).redirected).toBe(true);
  expect((await browserPost(context, "/connections/mcp/rename", `csrf_token=${csrf}&connection_id=${id}&display_name=a&display_name=b`)).status).toBe(400);
  expect(await nameOf(page, first)).toBe(EMOJI80);
  // A revoked target (and an unknown one) is not found, with the existing wording; nothing changes.
  await entry(page, second).getByRole("button", { name: "Revoke this connection" }).click();
  await page.waitForURL((u) => u.searchParams.get("status") === "revoked");
  for (const target of [second.slice(-36), "7f000000-0000-4000-8000-000000000001"]) {
    const gone = await browserPost(context, "/connections/mcp/rename", `csrf_token=${csrf}&connection_id=${target}&display_name=Revived`);
    expect(gone.status).toBe(404);
    expect(gone.text).toContain("not one of your active connections here");
  }
  await setupPage(page);
  expect(await urls(page)).not.toContain(second);
  // A forged CSRF token is refused.
  expect((await browserPost(context, "/connections/mcp/rename", `csrf_token=wrong&connection_id=${id}&display_name=x`)).status).toBe(403);
});

test("[M-LIST] the list shows only own connections on this host; revoking one removes it", async ({ page, context }) => {
  await signIn(page, DAVE);
  const first = await create(page);
  const second = await create(page);
  const listed = await urls(page);
  expect(listed).toContain(first);
  expect(listed).toContain(second);
  // G016 (Plan r4 §1.6, D-4(a)): the seeded connection dave created on Organization A's since-removed custom host is
  // listed through the current host with its own address; every other row is on this host.
  const old = listed.filter((u) => !CONNECTION_URL(ORIGIN_A).test(u));
  expect(old).toEqual([OLD_HOST_CONNECTION]);
  await expect(entry(page, OLD_HOST_CONNECTION).getByTestId("mcp-host")).toHaveText(new URL(OLD_HOST_CONNECTION).host);
  await expect(entry(page, first).getByTestId("mcp-host")).toHaveText(new URL(ORIGIN_A).host);
  // Newest first: the later connection is listed before the earlier one; with fewer than 100 there is no older page.
  expect(listed.indexOf(second)).toBeLessThan(listed.indexOf(first));
  await expect(page.getByTestId("mcp-page-older")).toHaveCount(0);
  // A page link that is not a valid cursor shows the closed invalid-page notice, which leads back to the first page.
  await page.goto(`${ORIGIN_A}/connections/mcp?cursor=!!`);
  await expect(page.getByRole("heading", { level: 1 })).toHaveText("Page link no longer valid");
  await page.getByTestId("mcp-page-newest").click();
  await page.waitForURL((u) => u.pathname === "/connections/mcp" && u.search === "");
  expect(await urls(page)).toContain(second);
  // The other host lists only its own Organization's connections.
  await signIn(page, DAVE, ORIGIN_B);
  await setupPage(page, ORIGIN_B);
  const onB = await urls(page);
  expect(onB.filter((u) => u.startsWith(ORIGIN_A) || u === OLD_HOST_CONNECTION)).toEqual([]);
  // Revoke the first on A.
  await setupPage(page);
  await entry(page, first).getByRole("button", { name: "Revoke this connection" }).click();
  await page.waitForURL((u) => u.searchParams.get("status") === "revoked");
  await expect(page.getByTestId("status-message")).toContainText("The connection was revoked");
  expect(await urls(page)).not.toContain(first);
  expect(await urls(page)).toContain(second);
  // A connection that is not the caller's (or no longer exists) is not found; nothing changes.
  const csrf = await csrfOf(page);
  const missing = await browserPost(context, "/connections/mcp/revoke", `csrf_token=${csrf}&connection_id=7f000000-0000-4000-8000-000000000001`);
  expect(missing.status).toBe(404);
  const stale = await browserPost(context, "/connections/mcp/revoke", `csrf_token=${csrf}&connection_id=${first.slice(-36)}`);
  expect(stale.status).toBe(404);
  const forged = await browserPost(context, "/connections/mcp/create", "csrf_token=wrong");
  expect(forged.status).toBe(403);
  expect(await urls(page)).toContain(second);
});

test("[M-A11Y] sign-in-first, not-member, empty and list states: keyboard and axe at 1280 and 320", async ({ page, browser }, testInfo) => {
  const tags = ["wcag2a", "wcag2aa", "wcag21a", "wcag21aa", "wcag22aa"];
  const audit = async (p: Page, label: string) => {
    const results = await new AxeBuilder({ page: p }).withTags(tags).analyze();
    // G069: only the kit's white label on its #0075ff primary (where no tenant palette applies) is an OWNER-WAIVED
    // color-contrast node, recorded with its ratio; every other violation fails (support/contrast.ts expectAxeClean).
    expectAxeClean(results.violations, `axe (${label})`);
  };
  const widths = async (p: Page, label: string) => {
    for (const width of [1280, 320]) {
      await p.setViewportSize({ width, height: width === 320 ? 640 : 800 });
      await audit(p, `${width} ${label}`);
      expect(await p.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth), `no horizontal scroll at ${width} (${label})`).toBe(true);
    }
    await p.setViewportSize({ width: 1280, height: 800 });
  };
  await setupPage(page);
  await widths(page, "sign-in-first");
  const erin = await open(browser, testInfo);
  await signInRefused(erin.page, ERIN);
  await widths(erin.page, "not-member");
  // dave on B: the list (or the empty state) with a long connection URL that must wrap at 320.
  await signIn(page, DAVE, ORIGIN_B);
  await setupPage(page, ORIGIN_B);
  await widths(page, "B before create");
  await create(page, ORIGIN_B);
  await widths(page, "list");
  // G063: the same shared manager embedded in My Account (names, rename and revoke forms), at both widths.
  await page.goto(`${ORIGIN_B}/account`);
  await expect(page.locator("#mcp-connections").getByTestId("mcp-connection").first()).toBeVisible();
  await widths(page, "account with connections");
  await setupPage(page, ORIGIN_B);
  // Keyboard: the create button is reachable and submits. (G063: each row now also has its name field and Save name
  // button, so the bound on Tab presses is raised.)
  const button = page.getByRole("button", { name: "Create connection" });
  for (let i = 0; i < 120 && !(await button.evaluate((b) => b === document.activeElement)); i++) await page.keyboard.press("Tab");
  await expect(button).toBeFocused();
  const before = (await urls(page)).length;
  await page.keyboard.press("Enter");
  await page.waitForURL((u) => u.searchParams.get("status") === "created");
  await expect(page.getByTestId("mcp-url")).toHaveCount(before + 1);
});

// G063 (Plan r1 §5.2, §6): the same shared manager on My Account (dave on Organization B): create, rename and revoke
// return to /account with mcp_status (disjoint from the page's status notices), Session and Sign-out stay intact,
// pagination links carry mcp_cursor and return to /account, and an invalid mcp_cursor renders the manager's own
// invalid-page state without breaking the rest of the page. For the page links, enough pending connections are created
// to pass one 100-row page; every connection this test creates is revoked again before it ends.
test("[M-ACCOUNT] My Account embeds the manager: create, rename, revoke, pages and isolation of its notices", async ({ page, context }) => {
  await signIn(page, DAVE, ORIGIN_B);
  const account = page.locator("#mcp-connections");
  const accountIntact = async () => {
    await expect(page.getByRole("heading", { level: 1, name: "Your account" })).toBeVisible();
    await expect(page.getByTestId("display-name")).toBeVisible();
    await expect(page.locator("#sign-out form")).toHaveCount(3);
  };
  await accountIntact();
  await expect(page.getByRole("navigation", { name: "Primary" }).getByRole("link", { name: "AI connections" })).toHaveAttribute("href", "#mcp-connections");
  const before = new Set(await account.getByTestId("mcp-url").allTextContents());
  // Create with a name from My Account.
  await page.locator("#mcp-create").getByLabel("Name (optional)").fill("Account client");
  await page.locator("#mcp-create").getByRole("button", { name: "Create connection" }).click();
  await page.waitForURL((u) => u.pathname === "/account" && u.searchParams.get("mcp_status") === "created" && u.hash === "#mcp-connections");
  await expect(account.getByTestId("status-message")).toContainText("A new connection was created");
  await accountIntact();
  const created = (await account.getByTestId("mcp-url").allTextContents()).filter((u) => !before.has(u));
  expect(created).toHaveLength(1);
  const mine = created[0];
  expect(await nameOf(page, mine)).toBe("Account client");
  // Rename from My Account.
  await rename(page, mine, "Account client two");
  expect(new URL(page.url()).pathname).toBe("/account");
  expect(await nameOf(page, mine)).toBe("Account client two");
  await accountIntact();
  // The two notice tables are disjoint: an MCP code in status and an account code in mcp_status show nothing.
  await page.goto(`${ORIGIN_B}/account?status=renamed&mcp_status=signed_in`);
  await expect(page.getByTestId("status-message")).toHaveCount(0);
  await page.goto(`${ORIGIN_B}/account?status=signed_in&mcp_status=renamed`);
  await expect(page.locator("#session").getByTestId("status-message")).toHaveText("You are signed in.");
  await expect(account.getByTestId("status-message")).toHaveText("The connection's name was saved.");
  await page.goto(`${ORIGIN_B}/account?mcp_status=renamed&mcp_status=created`);
  await expect(page.getByTestId("status-message")).toHaveCount(0);
  // An invalid mcp_cursor: the manager's invalid-page state, which leads back to /account; the rest stays intact.
  for (const bad of ["!!", "abc"]) {
    await page.goto(`${ORIGIN_B}/account?mcp_cursor=${bad}`);
    await expect(account.getByTestId("mcp-invalid-page")).toBeVisible();
    await expect(page.locator('form[action^="/connections/mcp/"]')).toHaveCount(0);
    await accountIntact();
  }
  await account.getByTestId("mcp-page-newest").click();
  await page.waitForURL((u) => u.pathname === "/account" && u.search === "" && u.hash === "#mcp-connections");
  // Page links: pass one 100-row page with disposable pending connections, then follow Older and Newest.
  const csrf = await csrfOf(page);
  const live = (await account.getByTestId("mcp-url").allTextContents()).length;
  const extra = Math.max(0, 101 - live);
  const listedBefore = new Set(await account.getByTestId("mcp-url").allTextContents());
  const disposable: string[] = [];
  try {
    expect(await bulkPost(context, ORIGIN_B, "/connections/mcp/create", Array.from({ length: extra }, () => `csrf_token=${csrf}&return_to=account`))).toEqual(
      Array.from({ length: extra }, () => 303),
    );
    await page.goto(`${ORIGIN_B}/account`);
    disposable.push(...(await account.getByTestId("mcp-url").allTextContents()).filter((u) => !listedBefore.has(u)));
    expect(disposable).toHaveLength(Math.min(extra, 100));
    await expect(account.getByTestId("mcp-connection")).toHaveCount(100);
    const older = account.getByTestId("mcp-page-older");
    await expect(older).toHaveAttribute("href", /^\/account\?mcp_cursor=[A-Za-z0-9_-]{1,128}#mcp-connections$/);
    await older.click();
    await page.waitForURL((u) => u.pathname === "/account" && u.searchParams.has("mcp_cursor") && u.hash === "#mcp-connections");
    await accountIntact();
    await expect(account.getByTestId("mcp-connection").first()).toBeVisible();
    // The created-first rows of this test may sit on the older page.
    disposable.push(...(await account.getByTestId("mcp-url").allTextContents()).filter((u) => !listedBefore.has(u) && !disposable.includes(u)));
    const newest = account.getByTestId("mcp-page-newest");
    await expect(newest).toHaveAttribute("href", "/account#mcp-connections");
    await newest.click();
    await page.waitForURL((u) => u.pathname === "/account" && u.search === "");
  } finally {
    const ids = disposable.map((u) => u.slice(-36));
    if (ids.length > 0) {
      expect(await bulkPost(context, ORIGIN_B, "/connections/mcp/revoke", ids.map((id) => `csrf_token=${csrf}&connection_id=${id}&return_to=account`))).toEqual(
        ids.map(() => 303),
      );
    }
  }
  expect(disposable).toHaveLength(extra);
  // Revoke this test's named connection from My Account; the list no longer shows it.
  await page.goto(`${ORIGIN_B}/account`);
  await entry(page, mine).getByRole("button", { name: "Revoke this connection" }).click();
  await page.waitForURL((u) => u.pathname === "/account" && u.searchParams.get("mcp_status") === "revoked");
  await expect(account.getByTestId("status-message")).toContainText("The connection was revoked");
  expect(await account.getByTestId("mcp-url").allTextContents()).not.toContain(mine);
  await accountIntact();
  // Keyboard: from the header, Tab reaches the first row's name field on My Account.
  await page.goto(`${ORIGIN_B}/account`);
  const field = account.getByTestId("mcp-connection").first().getByLabel("Name", { exact: true });
  for (let i = 0; i < 60 && !(await field.evaluate((f) => f === document.activeElement)); i++) await page.keyboard.press("Tab");
  await expect(field).toBeFocused();
});

// Hands B-MCP-E2E dave's browser sessions and CSRF tokens on both registry hosts, a spare session on A (for the
// current-session sign-out check) and the URLs of the connections created here. The file sits beside the report; the
// harness moves it into its 0700 inputs, registers every value as a secret and deletes it after use.
test("[M-HANDOFF] browser sessions and created connection URLs for the MCP end-to-end driver", async ({ page, context, browser }, testInfo) => {
  await signIn(page, DAVE);
  const a1 = await create(page);
  const a2 = await create(page);
  const liveA = await sessionValue(context);
  const csrfA = await csrfOf(page);
  await signIn(page, DAVE, ORIGIN_B);
  const b1 = await create(page, ORIGIN_B);
  const liveB = await sessionValue(context, ORIGIN_B);
  const csrfB = await csrfOf(page);
  const spare = await open(browser, testInfo);
  await signIn(spare.page, DAVE);
  const spareA = await sessionValue(spare.context);
  await setupPage(spare.page);
  const spareCsrfA = await csrfOf(spare.page);
  const report = dirname(process.env.PLAYWRIGHT_JSON_OUTPUT_FILE ?? "/tmp/playwright/results.json");
  const file = `${report}/handoff-${testInfo.project.name}.json`;
  const handoff = { origin_a: ORIGIN_A, origin_b: ORIGIN_B, live_a: liveA, csrf_a: csrfA, live_b: liveB, csrf_b: csrfB, spare_a: spareA, spare_csrf_a: spareCsrfA, a1, a2, b1 };
  writeFileSync(file, JSON.stringify(handoff) + "\n", { mode: 0o644 });
});
